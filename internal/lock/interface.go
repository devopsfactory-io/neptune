package lock

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"neptune/internal/domain"
	"neptune/internal/log"
	"neptune/internal/stacks"
)

// IsPROpenFunc returns true if the given PR number is still open.
type IsPROpenFunc func(ctx context.Context, prNumber string) (bool, error)

// Interface is the lock file interface: changed stacks, lock state, and lock/unlock operations.
type Interface struct {
	Config           *domain.NeptuneConfig
	Storage          ObjectStorage
	TerraformStacks  *domain.TerraformStacks
	LockStackDetails []domain.LockStackDetail
	IsPROpen         IsPROpenFunc
}

// NewInterface builds the lock interface: gets changed stacks from the configured provider (terramate or local), loads lock details from object storage (GCS or S3).
func NewInterface(ctx context.Context, cfg *domain.NeptuneConfig, isPROpen IsPROpenFunc) (*Interface, error) {
	provider := stacks.NewProvider(cfg)
	terraformStacks, err := provider.ChangedStacks(ctx, cfg)
	if err != nil {
		return nil, err
	}
	stacksMgmt := cfg.Repository.StacksManagement
	if stacksMgmt == "" {
		stacksMgmt = "terramate"
	}
	log.For("lock").Info("Getting changed Terraform stacks", "stacks_management", stacksMgmt)
	if len(terraformStacks.Stacks) == 0 {
		return &Interface{
			Config:           cfg,
			TerraformStacks:  terraformStacks,
			LockStackDetails: nil,
			IsPROpen:         isPROpen,
		}, nil
	}
	storage, err := NewObjectStorage(ctx, cfg.Repository.ObjectStorage, cfg.Repository.GitHub.Repository)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(cfg.Repository.ObjectStorage, "gs://") {
		log.For("lock").Info("Initializing GCS storage client")
	} else {
		log.For("lock").Info("Initializing S3 storage client")
	}
	log.For("lock").Info("Getting lock details for stacks")
	details, err := getLockDetails(ctx, storage, terraformStacks.Stacks)
	if err != nil {
		if closeErr := storage.Close(); closeErr != nil {
			return nil, fmt.Errorf("get lock details: %w; close: %v", err, closeErr)
		}
		return nil, err
	}
	return &Interface{
		Config:           cfg,
		Storage:          storage,
		TerraformStacks:  terraformStacks,
		LockStackDetails: details,
		IsPROpen:         isPROpen,
	}, nil
}

func getLockDetails(ctx context.Context, s ObjectStorage, stacks []string) ([]domain.LockStackDetail, error) {
	type result struct {
		i   int
		lf  *domain.LockFile
		err error
	}
	ch := make(chan result, len(stacks))
	for i, stack := range stacks {
		go func(i int, stack string) {
			lf, err := s.GetLockFile(ctx, stack)
			ch <- result{i: i, lf: lf, err: err}
		}(i, stack)
	}
	results := make([]result, len(stacks))
	for range stacks {
		r := <-ch
		results[r.i] = r
	}
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
	}
	out := make([]domain.LockStackDetail, len(stacks))
	for i, stack := range stacks {
		out[i] = domain.LockStackDetail{Path: stack, LockFile: results[i].lf}
	}
	return out, nil
}

// StacksLocked returns which stacks are locked by other PRs. If a lock's PR is closed, the lock is deleted.
func (iface *Interface) StacksLocked(ctx context.Context) (*domain.LockedStacks, error) {
	log.For("lock").Info("Checking locked status of stacks...")
	out := &domain.LockedStacks{Locked: false}
	currentPR := iface.Config.Repository.GitHub.PullRequestNumber
	for _, d := range iface.LockStackDetails {
		if d.LockFile == nil {
			continue
		}
		lockedBy := d.LockFile.LockedByPRID
		open, err := iface.IsPROpen(ctx, lockedBy)
		if err != nil {
			return nil, err
		}
		if !open {
			log.For("lock").Info("PR " + lockedBy + " is not open, unlocking stack " + d.Path)
			if err := iface.Storage.DeleteLockFile(ctx, d.Path); err != nil {
				log.For("lock").Error("delete lock file", "path", d.Path, "err", err)
			} else {
				log.For("lock").Info("Deleting lock file for stack " + d.Path)
			}
			continue
		}
		if lockedBy != currentPR {
			out.Locked = true
			out.StackPath = append(out.StackPath, d.Path)
			out.PRs = append(out.PRs, lockedBy)
		}
	}
	log.For("lock").Info("Stacks lock status: LockedStacks(locked=" + fmt.Sprint(out.Locked) + ", stack_path=" + fmt.Sprint(out.StackPath) + ", prs=" + fmt.Sprint(out.PRs) + ")")
	return out, nil
}

// DependsOnCompleted returns true if all depends_on phases for the given phase are completed for every stack.
func (iface *Interface) DependsOnCompleted(phase string) bool {
	log.For("lock").Debug("Checking depends_on for phase " + phase)
	wf, ok := iface.Config.Workflows.Workflows[iface.Config.Repository.AllowedWorkflow]
	if !ok {
		return false
	}
	ph, ok := wf.Phases[phase]
	if !ok {
		return true
	}
	if len(ph.DependsOn) == 0 {
		log.For("lock").Debug("Depends on: None")
		return true
	}
	log.For("lock").Debug("Depends on: " + strings.Join(ph.DependsOn, ", "))
	for _, dep := range ph.DependsOn {
		for _, d := range iface.LockStackDetails {
			if d.LockFile == nil {
				return false
			}
			p, ok := d.LockFile.WorkflowPhases[dep]
			if !ok {
				return false
			}
			if p.Status != domain.WorkflowStatusCompleted {
				return false
			}
		}
	}
	return true
}

func (iface *Interface) defineLockFile(ctx context.Context, phase, stackPath string, status domain.WorkflowStatus) (*domain.LockFile, error) {
	existing, err := iface.Storage.GetLockFile(ctx, stackPath)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return &domain.LockFile{
			LockedByPRID:   iface.Config.Repository.GitHub.PullRequestNumber,
			WorkflowPhases: map[string]domain.LockWorkflowPhase{phase: {Status: status}},
		}, nil
	}
	if existing.WorkflowPhases == nil {
		existing.WorkflowPhases = make(map[string]domain.LockWorkflowPhase)
	}
	existing.WorkflowPhases[phase] = domain.LockWorkflowPhase{Status: status}
	return existing, nil
}

// LockStacks writes lock files for all stacks with the given phase and status.
func (iface *Interface) LockStacks(ctx context.Context, phase string, stacks []string, status domain.WorkflowStatus) error {
	log.For("lock").Info("Locking stacks for phase " + phase + " and status " + string(status))
	var wg sync.WaitGroup
	errCh := make(chan error, len(stacks))
	for _, s := range stacks {
		wg.Add(1)
		go func(stack string) {
			defer wg.Done()
			log.For("lock").Debug("Defining lock file for stack " + stack + " and phase " + phase)
			lf, err := iface.defineLockFile(ctx, phase, stack, status)
			if err != nil {
				errCh <- err
				return
			}
			if err := iface.Storage.CreateOrUpdateLockFile(ctx, stack, lf); err != nil {
				errCh <- err
			}
		}(s)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateStacks updates the lock status for the given phase for all stacks.
func (iface *Interface) UpdateStacks(ctx context.Context, phase string, stacks []string, status domain.WorkflowStatus) error {
	log.For("lock").Info("Updating lock status for stacks " + fmt.Sprint(stacks) + " and phase " + phase)
	for _, stack := range stacks {
		log.For("lock").Debug("Defining lock file for stack " + stack + " and phase " + phase)
		lf, err := iface.defineLockFile(ctx, phase, stack, status)
		if err != nil {
			return err
		}
		if err := iface.Storage.CreateOrUpdateLockFile(ctx, stack, lf); err != nil {
			return err
		}
	}
	return nil
}

// UnlockAllStacks deletes lock files for all stacks in the interface.
func (iface *Interface) UnlockAllStacks(ctx context.Context) error {
	for _, d := range iface.LockStackDetails {
		if err := iface.Storage.DeleteLockFile(ctx, d.Path); err != nil {
			return err
		}
	}
	return nil
}

// Close releases resources (e.g. storage client).
func (iface *Interface) Close() error {
	if iface.Storage != nil {
		return iface.Storage.Close()
	}
	return nil
}
