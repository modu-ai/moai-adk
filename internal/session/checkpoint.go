package session

import "fmt"

// Checkpoint is the interface for per-phase checkpoint data.
type Checkpoint interface {
	PhaseName() Phase
	Validate() error
}

// PlanCheckpoint holds state after the plan phase completes.
type PlanCheckpoint struct {
	SPECID       string `json:"spec_id" validate:"required"`
	Status       string `json:"status" validate:"required,oneof=approved draft rejected"` // SPEC-V3R2-RT-004 REQ-015
	ResearchPath string `json:"research_path"`
}

// PhaseName returns the phase for this checkpoint.
func (c *PlanCheckpoint) PhaseName() Phase {
	return PhasePlan
}

// Validate checks PlanCheckpoint fields against validator/v10 rules.
func (c *PlanCheckpoint) Validate() error {
	// TODO: SPEC-V3R2-SCH-001 validator/v10 통합 후 구현
	// 현재는 manual validation
	if c.SPECID == "" {
		return fmt.Errorf("SPECID: required field is empty")
	}
	validStatuses := map[string]bool{
		"approved": true,
		"draft":    true,
		"rejected": true,
	}
	if !validStatuses[c.Status] {
		return fmt.Errorf("status: must be oneof=[approved draft rejected], got %q", c.Status)
	}
	return nil
}

// RunCheckpoint holds state after the run phase.
type RunCheckpoint struct {
	SPECID        string `json:"spec_id" validate:"required"`
	Status        string `json:"status" validate:"required,oneof=pass fail partial"` // SPEC-V3R2-RT-004 REQ-015
	Harness       string `json:"harness" validate:"required,oneof=minimal standard thorough"` // SPEC-V3R2-RT-004 AC-15
	TestsTotal    int    `json:"tests_total"`
	TestsPassed   int    `json:"tests_passed"`
	FilesModified int    `json:"files_modified"`
}

// PhaseName returns the phase for this checkpoint.
func (c *RunCheckpoint) PhaseName() Phase {
	return PhaseRun
}

// Validate checks RunCheckpoint fields against validator/v10 rules.
func (c *RunCheckpoint) Validate() error {
	// TODO: SPEC-V3R2-SCH-001 validator/v10 통합 후 구현
	if c.SPECID == "" {
		return fmt.Errorf("SPECID: required field is empty")
	}
	validStatuses := map[string]bool{
		"pass":    true,
		"fail":    true,
		"partial": true,
	}
	if !validStatuses[c.Status] {
		return fmt.Errorf("status: must be oneof=[pass fail partial], got %q", c.Status)
	}
	if c.Harness == "" {
		return fmt.Errorf("harness: required field is empty")
	}
	validHarnesses := map[string]bool{
		"minimal":  true,
		"standard": true,
		"thorough": true,
	}
	if !validHarnesses[c.Harness] {
		return fmt.Errorf("harness: must be oneof=[minimal standard thorough], got %q", c.Harness)
	}
	return nil
}

// SyncCheckpoint holds state after the sync phase.
type SyncCheckpoint struct {
	SPECID     string `json:"spec_id" validate:"required"`
	PRNumber   int    `json:"pr_number,omitempty"`
	PRURL      string `json:"pr_url,omitempty"`
	DocsSynced bool   `json:"docs_synced"`
}

// PhaseName returns the phase for this checkpoint.
func (c *SyncCheckpoint) PhaseName() Phase {
	return PhaseSync
}

// Validate checks SyncCheckpoint fields against validator/v10 rules.
func (c *SyncCheckpoint) Validate() error {
	if c.SPECID == "" {
		return fmt.Errorf("SPECID: required field is empty")
	}
	return nil
}
