package lock

import (
	"context"
	"sync"

	"neptune/internal/domain"
)

// IsPROpenFunc returns true if the given PR number is still open.
type IsPROpenFunc func(ctx context.Context, prNumber string) (bool, error)

// Interface is the lock file interface: changed stacks, lock state, and lock/unlock operations.
type Interface struct {
	Config           *domain.NeptuneConfig
	Storage          *GCSStorage
	TerraformStacks  *domain.TerraformStacks
	LockStackDetails []domain.LockStackDetail
	IsPROpen         IsPROpenFunc
}

// NewInterface builds the lock interface: runs terramate, loads lock details from GCS.
func NewInterface(ctx context.Context, cfg *domain.NeptuneConfig, isPROpen IsPROpenFunc) (*Interface, error) {
	stacks, err := ChangedStacks(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if len(stacks.Stacks) == 0 {
		return &Interface{
			Config:           cfg,
			TerraformStacks:  stacks,
			LockStackDetails: nil,
			IsPROpen:         isPROpen,
		}, nil
	}
	storage, err := NewGCSStorage(ctx, cfg.Repository.ObjectStorage, cfg.Repository.GitHub.Repository)
	if err != nil {
		return nil, err
	}
	details, err := getLockDetails(ctx, storage, stacks.Stacks)
	if err != nil {
		_ = storage.Close()
		return nil, err
	}
	return &Interface{
		Config:           cfg,
		Storage:          storage,
		TerraformStacks:  stacks,
		LockStackDetails: details,
		IsPROpen:         isPROpen,
	}, nil
}

func getLockDetails(ctx context.Context, s *GCSStorage, stacks []string) ([]domain.LockStackDetail, error) {
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
			_ = iface.Storage.DeleteLockFile(ctx, d.Path)
			continue
		}
		if lockedBy != currentPR {
			out.Locked = true
			out.StackPath = append(out.StackPath, d.Path)
			out.PRs = append(out.PRs, lockedBy)
		}
	}
	return out, nil
}

// DependsOnCompleted returns true if all depends_on phases for the given phase are completed for every stack.
func (iface *Interface) DependsOnCompleted(phase string) bool {
	wf, ok := iface.Config.Workflows.Workflows[iface.Config.Repository.AllowedWorkflow]
	if !ok {
		return false
	}
	ph, ok := wf.Phases[phase]
	if !ok {
		return true
	}
	if len(ph.DependsOn) == 0 {
		return true
	}
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
	var wg sync.WaitGroup
	errCh := make(chan error, len(stacks))
	for _, s := range stacks {
		wg.Add(1)
		go func(stack string) {
			defer wg.Done()
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
	for _, stack := range stacks {
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

// Close releases resources (e.g. GCS client).
func (iface *Interface) Close() error {
	if iface.Storage != nil {
		return iface.Storage.Close()
	}
	return nil
}
