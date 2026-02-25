package domain

// RunOutput is the result of running one step.
type RunOutput struct {
	Command string
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
