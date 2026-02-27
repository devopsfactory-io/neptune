package domain

// RunOutput is the result of running one step.
// When the step runs per-stack (terramate: true), Stack is set to the stack path; otherwise it is empty.
type RunOutput struct {
	Command string
	Stack   string // stack path when run per-stack (e.g. "stack-a"); empty when terramate: false
	Output  string
	Error   string
	Status  int
}

// StepsOutput aggregates run outputs for a phase.
type StepsOutput struct {
	Phase         string
	OverallStatus int
	Outputs       []RunOutput
}
