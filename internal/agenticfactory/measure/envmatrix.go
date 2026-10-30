// Package measure holds the two record shapes the Agentic Factory uses to say what was measured and what was
// not (Issue #1948): the local environment matrix (REQ-AF-024, design §5.1) and the efficiency comparison
// record (REQ-AF-025, design §4.7). Both exist to keep an unobserved value from reading as an observed one:
// a step that did not run is "not-run", never success; an estimate is never filed as a measurement.
//
// This package only models and validates the records. It runs nothing, calls no host and writes no file.
package measure

import (
	"errors"
	"fmt"
	"strings"
)

// Step is one column of the environment matrix that is executed. The installable column is an observation,
// not a step; the six steps below are the ones a PASS needs.
type Step string

const (
	StepVersion    Step = "version"
	StepStart      Step = "start"
	StepReceive    Step = "receive"
	StepTest       Step = "test"
	StepHoldResume Step = "hold_resume"
	StepCollect    Step = "collect"
)

// Steps lists the six steps in matrix order.
var Steps = []Step{StepVersion, StepStart, StepReceive, StepTest, StepHoldResume, StepCollect}

// ErrStepEvidence marks a step cell that claims to have run but lacks its command, exit code or output.
var ErrStepEvidence = errors.New("measure: a step that ran must carry its command, exit code and output")

// StepRun is one step cell. A cell with Ran false is "not-run" whatever else it holds.
type StepRun struct {
	Ran      bool
	Command  string
	ExitCode *int // nil means the exit code was not observed
	Output   string
}

// StepStatus is what a cell says about its step.
type StepStatus string

const (
	StatusNotRun StepStatus = "not-run"
	StatusOK     StepStatus = "ok"
	StatusFailed StepStatus = "failed"
)

// Status classifies the cell. Only a cell that ran, exited 0, is ok; a cell that did not run is not-run and
// is never success.
func (s StepRun) Status() StepStatus {
	switch {
	case !s.Ran:
		return StatusNotRun
	case s.ExitCode != nil && *s.ExitCode == 0:
		return StatusOK
	}
	return StatusFailed
}

// Installable is the installation observation.
type Installable string

const (
	InstallYes    Installable = "yes"
	InstallNo     Installable = "no"
	InstallDenied Installable = "denied"
)

// Row is one environment of the matrix.
type Row struct {
	Environment string
	Installable Installable
	// Configured is false for an environment that needs an explicit setting that is absent (Codex without
	// its opt-in). Such a row stays BLOCKED(not-configured); the matrix is never filled by switching it on.
	Configured bool
	Cells      map[Step]StepRun
}

// Validate refuses a row whose executed cells lack evidence. A cell that did not run needs none.
func (r Row) Validate() error {
	if strings.TrimSpace(r.Environment) == "" {
		return errors.New("measure: a matrix row needs an environment name")
	}
	for _, st := range Steps {
		c, ok := r.Cells[st]
		if !ok || !c.Ran {
			continue
		}
		var missing []string
		if strings.TrimSpace(c.Command) == "" {
			missing = append(missing, "command")
		}
		if c.ExitCode == nil {
			missing = append(missing, "exit code")
		}
		if strings.TrimSpace(c.Output) == "" {
			missing = append(missing, "output")
		}
		if len(missing) > 0 {
			return fmt.Errorf("%w: %s of %q lacks %s", ErrStepEvidence, st, r.Environment, strings.Join(missing, ", "))
		}
	}
	return nil
}

// Verdict is PASS only when the environment is installable, configured, and all six steps ran and succeeded.
// Anything else is BLOCKED with the first reason found. A row that does not validate is BLOCKED(invalid-evidence).
func (r Row) Verdict() string {
	switch {
	case r.Validate() != nil:
		return "BLOCKED(invalid-evidence)"
	case !r.Configured:
		return "BLOCKED(not-configured)"
	case r.Installable != InstallYes:
		return "BLOCKED(not-installable)"
	}
	for _, st := range Steps {
		switch r.Cells[st].Status() {
		case StatusNotRun:
			return "BLOCKED(not-run:" + string(st) + ")"
		case StatusFailed:
			return "BLOCKED(failed:" + string(st) + ")"
		}
	}
	return "PASS"
}

// Matrix is the set of environment rows.
type Matrix struct{ Rows []Row }

// Verdict is PASS only when at least one row is PASS (that environment's six steps really ran and succeeded).
// A matrix whose every row is BLOCKED fails; an empty matrix fails.
func (m Matrix) Verdict() string {
	for _, r := range m.Rows {
		if r.Verdict() == "PASS" {
			return "PASS"
		}
	}
	return "FAIL"
}
