package session

// Checkpoint is the interface for per-phase checkpoint data.
type Checkpoint interface {
	PhaseName() Phase
}

// PlanCheckpoint holds state after the plan phase completes.
type PlanCheckpoint struct {
	SPECID       string `json:"spec_id"`
	Status       string `json:"status"`        // "approved", "draft", "rejected"
	ResearchPath string `json:"research_path"`
}

// PhaseName returns the phase for this checkpoint.
func (c *PlanCheckpoint) PhaseName() Phase {
	return PhasePlan
}

// RunCheckpoint holds state after the run phase.
type RunCheckpoint struct {
	SPECID        string `json:"spec_id"`
	Status        string `json:"status"`  // "pass", "fail", "partial"
	TestsTotal    int    `json:"tests_total"`
	TestsPassed   int    `json:"tests_passed"`
	FilesModified int    `json:"files_modified"`
}

// PhaseName returns the phase for this checkpoint.
func (c *RunCheckpoint) PhaseName() Phase {
	return PhaseRun
}

// SyncCheckpoint holds state after the sync phase.
type SyncCheckpoint struct {
	SPECID     string `json:"spec_id"`
	PRNumber   int    `json:"pr_number,omitempty"`
	PRURL      string `json:"pr_url,omitempty"`
	DocsSynced bool   `json:"docs_synced"`
}

// PhaseName returns the phase for this checkpoint.
func (c *SyncCheckpoint) PhaseName() Phase {
	return PhaseSync
}
