package domain

// WorkflowStatus is the status of a workflow phase for a stack.
type WorkflowStatus string

const (
	// WorkflowStatusInProgress indicates a workflow phase is currently running.
	WorkflowStatusInProgress WorkflowStatus = "in_progress"
	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
)

// LockWorkflowPhase is the stored phase status in a lock file.
type LockWorkflowPhase struct {
	Status WorkflowStatus `json:"status"`
}

// LockFile is the JSON structure stored in object storage per stack.
type LockFile struct {
	LockedByPRID   string                       `json:"locked_by_pr_id"`
	WorkflowPhases map[string]LockWorkflowPhase `json:"workflow_phases"`
}

// LockedStacks is the result of checking if any stack is locked by another PR.
type LockedStacks struct {
	Locked    bool
	StackPath []string
	PRs       []string
}

// LockStackDetail is one stack's path and optional lock file.
type LockStackDetail struct {
	Path     string
	LockFile *LockFile
}

// LockStacksDetails holds details for all changed stacks.
type LockStacksDetails struct {
	Details []LockStackDetail
}

// TerraformStacks is the list of changed stack paths from the stacks provider (terramate or local).
type TerraformStacks struct {
	Stacks []string
}
