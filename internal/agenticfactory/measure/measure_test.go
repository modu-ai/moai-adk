package measure

// Defect answer: these tests catch a matrix that accepts a step cell without its evidence, one that lets a
// not-run step read as success or a fully blocked matrix read as a pass, an unconfigured Codex row that is
// filled in anyway, an efficiency record with a missing field, an estimate filed as a measurement, a gain
// reported next to a regression, and a checkpoint cost that is a guess. Positive controls sit next to each
// refusal so a validator that refuses everything cannot pass.

import (
	"errors"
	"strings"
	"testing"
)

func ptr(n int) *int { return &n }

func ran(cmd string, code int, out string) StepRun {
	return StepRun{Ran: true, Command: cmd, ExitCode: ptr(code), Output: out}
}

func allOK() map[Step]StepRun {
	m := map[Step]StepRun{}
	for _, s := range Steps {
		m[s] = ran("run "+string(s), 0, "ok "+string(s))
	}
	return m
}

func goodRow() Row {
	return Row{Environment: "local-a", Installable: InstallYes, Configured: true, Cells: allOK()}
}

func TestEnvMatrix_StepEvidenceRequired_NotRunNeverSuccess(t *testing.T) {
	t.Run("a_executed_step_without_command_exit_code_or_output_is_refused", func(t *testing.T) {
		if err := goodRow().Validate(); err != nil {
			t.Fatalf("control: a complete row: %v", err)
		}
		for name, cell := range map[string]StepRun{
			"no command":   {Ran: true, ExitCode: ptr(0), Output: "x"},
			"no exit code": {Ran: true, Command: "c", Output: "x"},
			"no output":    {Ran: true, Command: "c", ExitCode: ptr(0)},
		} {
			row := goodRow()
			row.Cells[StepTest] = cell
			err := row.Validate()
			if !errors.Is(err, ErrStepEvidence) {
				t.Errorf("%s: %v, want ErrStepEvidence", name, err)
			}
			if v := row.Verdict(); v == "PASS" {
				t.Errorf("%s: the row passed", name)
			}
		}
	})

	t.Run("b_a_step_that_did_not_run_is_not_run_and_never_success", func(t *testing.T) {
		row := goodRow()
		row.Cells[StepHoldResume] = StepRun{} // not run
		if got := row.Cells[StepHoldResume].Status(); got != StatusNotRun {
			t.Fatalf("status = %s, want not-run", got)
		}
		// even a cell that carries a successful-looking exit code is not-run when it did not run.
		row.Cells[StepCollect] = StepRun{Ran: false, Command: "c", ExitCode: ptr(0), Output: "ok"}
		if got := row.Cells[StepCollect].Status(); got != StatusNotRun {
			t.Errorf("an unrun cell with an exit code 0 = %s, want not-run", got)
		}
		if v := row.Verdict(); v != "BLOCKED(not-run:hold_resume)" {
			t.Errorf("verdict = %q, want BLOCKED(not-run:hold_resume)", v)
		}
		row = goodRow()
		row.Cells[StepStart] = ran("start", 3, "failed")
		if v := row.Verdict(); v != "BLOCKED(failed:start)" {
			t.Errorf("a failed step: verdict = %q", v)
		}
		if v := goodRow().Verdict(); v != "PASS" {
			t.Errorf("control: all six steps ran and succeeded = %q, want PASS", v)
		}
	})

	t.Run("c_a_matrix_whose_every_row_is_blocked_fails", func(t *testing.T) {
		blocked := Row{Environment: "codex", Installable: InstallYes, Configured: false}
		failed := goodRow()
		failed.Cells[StepReceive] = StepRun{}
		if v := (Matrix{Rows: []Row{blocked, failed}}).Verdict(); v != "FAIL" {
			t.Errorf("all rows blocked: %q, want FAIL", v)
		}
		if v := (Matrix{}).Verdict(); v != "FAIL" {
			t.Errorf("an empty matrix: %q, want FAIL", v)
		}
		if v := (Matrix{Rows: []Row{blocked, goodRow()}}).Verdict(); v != "PASS" {
			t.Errorf("control: one environment ran all six steps: %q, want PASS", v)
		}
	})

	t.Run("d_codex_without_its_setting_stays_blocked_not_configured", func(t *testing.T) {
		row := Row{Environment: "codex", Installable: InstallYes, Configured: false, Cells: allOK()}
		if v := row.Verdict(); v != "BLOCKED(not-configured)" {
			t.Errorf("verdict = %q, want BLOCKED(not-configured) even though every cell is filled", v)
		}
	})
}

func goodRecord() Record {
	return Record{
		Axis: AxisTime, BaselineSource: "preserved log scratchpad/b2", Comparability: "task differs; limits stated",
		ValueKind: KindMeasured, Value: "588.7", Unit: "s", Command: "go test -race -count=1 -run ^TestFactory ./internal/cli/",
		Tree:    "4cdaba5d7e6821aa9751d100cf75a328b4726704",
		Quality: QualityGuard{Metric: "tests", Result: "pass"}, ThresholdState: ThresholdProposed,
	}
}

func TestEfficiencyRecord_RefusesMissingAndMisfiledFields(t *testing.T) {
	if err := goodRecord().Validate(); err != nil {
		t.Fatalf("control: a complete measured record: %v", err)
	}
	t.Run("a_a_missing_field_is_refused_and_named", func(t *testing.T) {
		for name, mod := range map[string]func(*Record){
			"axis":            func(r *Record) { r.Axis = "" },
			"baseline_source": func(r *Record) { r.BaselineSource = " " },
			"comparability":   func(r *Record) { r.Comparability = "" },
			"quality_guard":   func(r *Record) { r.Quality = QualityGuard{} },
			"threshold_state": func(r *Record) { r.ThresholdState = "" },
			"command":         func(r *Record) { r.Command = "" },
			"tree":            func(r *Record) { r.Tree = "main" }, // a branch name is not a tree SHA
			"unit":            func(r *Record) { r.Unit = "" },
			"value_kind":      func(r *Record) { r.ValueKind = "" },
		} {
			r := goodRecord()
			mod(&r)
			err := r.Validate()
			if !errors.Is(err, ErrFieldMissing) || !strings.Contains(err.Error(), name) {
				t.Errorf("%s: %v, want ErrFieldMissing naming it", name, err)
			}
		}
	})
	t.Run("b_an_estimate_in_the_measured_slot_is_refused", func(t *testing.T) {
		r := goodRecord()
		r.EstimatedFrom = "3356 spawns times an assumed size"
		if err := r.Validate(); !errors.Is(err, ErrEstimateAsMeasure) {
			t.Errorf("measured value with an estimate basis: %v", err)
		}
		u := goodRecord()
		u.ValueKind = KindUnknown
		if err := u.Validate(); !errors.Is(err, ErrEstimateAsMeasure) {
			t.Errorf("UNKNOWN carrying a value: %v", err)
		}
		u.Value, u.Unit, u.Command, u.Tree = "", "", "", ""
		if err := u.Validate(); err != nil {
			t.Errorf("control: UNKNOWN without a value: %v", err)
		}
		e := goodRecord()
		e.ValueKind, e.Command, e.Tree = KindEstimated, "", ""
		if err := e.Validate(); !errors.Is(err, ErrFieldMissing) {
			t.Errorf("an estimate with no basis: %v, want ErrFieldMissing", err)
		}
		e.EstimatedFrom = "operator figure"
		if err := e.Validate(); err != nil {
			t.Errorf("control: an estimate with its basis: %v", err)
		}
	})
}

func TestEfficiencyRecord_NoGainReportedOverARegression(t *testing.T) {
	r := goodRecord()
	if err := r.ReportGain(true); err != nil {
		t.Fatalf("control: a gain with no regression: %v", err)
	}
	r.Quality = QualityGuard{Metric: "correctness", Result: "2 tests fail", Regressed: true}
	if err := r.ReportGain(true); !errors.Is(err, ErrGainHidesRegress) {
		t.Errorf("a gain next to a regression: %v, want ErrGainHidesRegress", err)
	}
	if err := r.ReportGain(false); err != nil {
		t.Errorf("no gain claimed: %v", err)
	}
	e := goodRecord()
	e.ValueKind, e.Command, e.Tree, e.EstimatedFrom = KindEstimated, "", "", "assumed"
	if err := e.ReportGain(true); !errors.Is(err, ErrEstimateAsMeasure) {
		t.Errorf("a gain backed by an estimate: %v", err)
	}
}

func TestEfficiencyRecord_CheckpointCostIsMeasuredOrAGap(t *testing.T) {
	m := goodRecord()
	if err := CheckpointCost(&m, ""); err != nil {
		t.Errorf("a measured cost: %v", err)
	}
	if err := CheckpointCost(nil, "no checkpoint was run in this tree"); err != nil {
		t.Errorf("a recorded Gap: %v", err)
	}
	if err := CheckpointCost(nil, ""); !errors.Is(err, ErrCheckpointCost) {
		t.Errorf("an empty cell: %v", err)
	}
	est := goodRecord()
	est.ValueKind, est.Command, est.Tree, est.EstimatedFrom = KindEstimated, "", "", "assumed"
	if err := CheckpointCost(&est, "a gap note"); !errors.Is(err, ErrCheckpointCost) {
		t.Errorf("an estimate, even with a gap note beside it: %v, want refusal", err)
	}
	unk := goodRecord()
	unk.ValueKind, unk.Value, unk.Unit, unk.Command, unk.Tree = KindUnknown, "", "", "", ""
	if err := CheckpointCost(&unk, ""); !errors.Is(err, ErrCheckpointCost) {
		t.Errorf("UNKNOWN with no Gap: %v", err)
	}
}
