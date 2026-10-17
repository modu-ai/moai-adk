package cli

// launch_view_test.go — SPEC-LAUNCHER-ARGV-001 M1: the canonical resolver.
//
//   - TestLaunchViewBackendSelection  the backend-model worktree selection (REQ-LA-003)
//   - TestLaunchViewConsumption       the extracted consumption loop (REQ-LA-001)
//   - TestLaunchViewStageReads        the three consumer-scoped stage reads (SPEC §2, D8-D11)
//   - TestLauncherAnchorFinalArgv/view_exec_byte_equality   AC-LA-005 (grown by M2)

import (
	"path/filepath"
	"reflect"
	"testing"
)

// pinSnapshotModel pins the compile-time claude option model for the duration
// of a test: the derived model depends on the claude binary on PATH, which a
// unit test must not.
func pinSnapshotModel(t *testing.T) {
	t.Helper()
	prev := activeClaudeOptionModel
	activeClaudeOptionModel = claudeOptionModelSnapshot
	t.Cleanup(func() { activeClaudeOptionModel = prev })
}

func TestLaunchViewBackendSelection(t *testing.T) {
	pinSnapshotModel(t)
	cases := []struct {
		name string
		argv []string
		want worktreeSelection
	}{
		{"no option", []string{"--model", "opus", "prompt"}, worktreeSelection{}},
		{"two-token long", []string{"--worktree", "WT-card"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "WT-card", HasValue: true}},
		{"two-token short", []string{"-w", "WT-card"}, worktreeSelection{Present: true, Spelling: "-w", Value: "WT-card", HasValue: true}},
		{"long equals", []string{"--worktree=WT-card"}, worktreeSelection{Present: true, Spelling: "--worktree=", Value: "WT-card", HasValue: true}},
		{"short equals", []string{"-w=WT-card"}, worktreeSelection{Present: true, Spelling: "-w=", Value: "WT-card", HasValue: true}},
		{"attached", []string{"-wWT-card"}, worktreeSelection{Present: true, Spelling: "-w<value>", Value: "WT-card", HasValue: true}},
		{"cluster boolean then w", []string{"-pwWT-card", "prompt"}, worktreeSelection{Present: true, Spelling: "-<cluster>w<value>", Value: "WT-card", HasValue: true}},
		{"cluster takes the next token", []string{"-pw", "WT-card"}, worktreeSelection{Present: true, Spelling: "-<cluster>w<value>", Value: "WT-card", HasValue: true, Index: 0}},
		{"cluster with an unknown letter ends the walk", []string{"-xwWT-card"}, worktreeSelection{}},
		{"cluster whose value option precedes w", []string{"-nwWT-card"}, worktreeSelection{}},
		{"bare option", []string{"--worktree"}, worktreeSelection{Present: true, Spelling: "--worktree"}},
		{"bare before another option", []string{"--worktree", "--model", "opus"}, worktreeSelection{Present: true, Spelling: "--worktree"}},
		{"empty equals form names no tree", []string{"--worktree=", "plain"}, worktreeSelection{Present: true, Spelling: "--worktree="}},
		{"the last option wins", []string{"--worktree", "plain", "--worktree", "WT-card"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "WT-card", HasValue: true, Index: 2}},
		{"a later bare option wins over an earlier value", []string{"--worktree", "WT-card", "--worktree"}, worktreeSelection{Present: true, Spelling: "--worktree", Index: 2}},
		{"separator ends option parsing", []string{"--worktree", "plain", "--", "--worktree", "WT-card"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "plain", HasValue: true}},
		{"separator is never a value", []string{"--worktree", "--", "WT-card"}, worktreeSelection{Present: true, Spelling: "--worktree"}},
		{"a required-value option takes a flag-shaped token as its value", []string{"--name", "--worktree", "WT-card"}, worktreeSelection{}},
		{"a required-value option takes the next token", []string{"--model", "-wWT-card"}, worktreeSelection{}},
		{"an optional-value option leaves a flag-shaped token alone", []string{"--debug", "-wWT-card"}, worktreeSelection{Present: true, Spelling: "-w<value>", Value: "WT-card", HasValue: true, Index: 1}},
		{"a prompt that merely contains the words is positional", []string{"--worktree", "plain", "fix -w WT-card please"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "plain", HasValue: true}},
		{"a lone dash is positional", []string{"-", "--worktree", "x"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "x", HasValue: true, Index: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := selectBackendWorktree(tc.argv); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("selectBackendWorktree(%q) = %+v, want %+v", tc.argv, got, tc.want)
			}
		})
	}
}

func TestLaunchViewConsumption(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want launchConsumption
	}{
		{"nothing", nil, launchConsumption{}},
		{"bypass and continue are eaten", []string{"-b", "-c", "--worktree", "plain"},
			launchConsumption{PassThrough: []string{"--worktree", "plain"}, PermissionMode: "bypassPermissions", PermissionModeSet: true, Continue: true, ChainTarget: "plain"}},
		{"model forms, last wins", []string{"--model", "a", "-m", "b", "--model=c"},
			launchConsumption{Model: "c", ModelSet: true}},
		{"a valueless --permission-mode is dropped", []string{"--permission-mode"}, launchConsumption{}},
		{"permission mode last wins over -b", []string{"-b", "--permission-mode=plan"},
			launchConsumption{PermissionMode: "plan", PermissionModeSet: true}},
		{"the first separator is dropped and the tail is verbatim", []string{"-b", "--", "-m", "x", "--", "y"},
			launchConsumption{PassThrough: []string{"-m", "x", "--", "y"}, PermissionMode: "bypassPermissions", PermissionModeSet: true}},
		{"a consumed option brings the value next to the flag", []string{"--worktree", "-b", "--", "card"},
			launchConsumption{PassThrough: []string{"--worktree", "card"}, PermissionMode: "bypassPermissions", PermissionModeSet: true, ChainTarget: "card"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := consumeLaunchArgs(tc.args); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("consumeLaunchArgs(%q) = %+v, want %+v", tc.args, got, tc.want)
			}
		})
	}
}

// TestLaunchViewStageReads pins the three consumer-scoped reads against the
// measured per-stage table (stageRows): StageB is the value the precheck and
// the seeder read, ChainTarget the value the chain populator reads, and the two
// stages the table marks as differing do differ in the view.
func TestLaunchViewStageReads(t *testing.T) {
	for _, row := range stageRows {
		t.Run(row.name, func(t *testing.T) {
			args := substituteElsewhere(row.args, filepath.Join(t.TempDir(), "elsewhere"))
			v := resolveLaunchView(args)
			if v.StageB.Value != row.precheck || v.StageB.Value != row.seed {
				t.Errorf("[%s] StageB = %+v, want value %q", row.stages, v.StageB, row.precheck)
			}
			wantChain := ""
			if row.chain != "" {
				wantChain = filepath.Base(row.chain)
			}
			if v.Consumed.ChainTarget != wantChain {
				t.Errorf("[%s] ChainTarget = %q, want %q", row.stages, v.Consumed.ChainTarget, wantChain)
			}
		})
	}

	stageA := []struct {
		name string
		args []string
		want worktreeRead
	}{
		{"bare -w before --branch reads no value at stage A", []string{"-w", "--branch", "main", "plain"}, worktreeRead{}},
		{"the empty equals form is a present empty value", []string{"--worktree=", "plain"}, worktreeRead{Has: true}},
		{"nothing before the separator", []string{"--", "--worktree", "plain"}, worktreeRead{}},
	}
	for _, tc := range stageA {
		t.Run("stage A: "+tc.name, func(t *testing.T) {
			if got := resolveLaunchView(tc.args).StageA; got != tc.want {
				t.Errorf("StageA = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// viewEqualityRow is one §D.4 spelling for the view-exec equality test.
type viewEqualityRow struct {
	name string
	args []string
	// selection is the expected backend selection over the final argv. Index is
	// not asserted here (it is a position, covered by TestLaunchViewBackendSelection).
	spelling string
	value    string
	hasValue bool
}

// viewEqualityRows are the acceptance.md §D.4 spellings (rows 1-14). The
// spelling column names the form in the FINAL argv: the launcher rewrites the
// pre-separator -w / --worktree / --worktree= forms to the two-token form, so
// rows 1-3 and 5-6 end up as `--worktree <value>`; the attached and cluster
// forms reach the backend untouched.
func viewEqualityRows(elsewhere string) []viewEqualityRow {
	return []viewEqualityRow{
		{"1 -w WT-card", []string{"-w", "WT-card"}, "--worktree", "WT-card", true},
		{"2 --worktree WT-card", []string{"--worktree", "WT-card"}, "--worktree", "WT-card", true},
		{"3 --worktree=WT-card", []string{"--worktree=WT-card"}, "--worktree", "WT-card", true},
		{"4 -wWT-card", []string{"-wWT-card"}, "-w<value>", "WT-card", true},
		{"5 -w -b -- WT-card (consumed option shadowing)", []string{"-w", "-b", "--", "WT-card"}, "--worktree", "WT-card", true},
		{"6 -w plain-tree --worktree --branch main WT-card (last wins)", []string{"-w", "plain-tree", "--worktree", "--branch", "main", "WT-card"}, "--worktree", "WT-card", true},
		{"7 -- -pwWT-card PROMPT (bundled cluster)", []string{"--", "-pwWT-card", "PROMPT"}, "-<cluster>w<value>", "WT-card", true},
		{"8 bare -w", []string{"-w"}, "--worktree", "", false},
		{"9 -- --worktree WT-card PROMPT", []string{"--", "--worktree", "WT-card", "PROMPT"}, "--worktree", "WT-card", true},
		{"10 --worktree=WT-card -- TAIL", []string{"--worktree=WT-card", "--", "TAIL"}, "--worktree", "WT-card", true},
		{"11 -w plain -- prompt that contains -w WT-card", []string{"-w", "plain", "--", "fix -w WT-card please"}, "--worktree", "plain", true},
		{"12 -w plain -- P1 -- --worktree WT-card (second separator)", []string{"-w", "plain", "--", "P1", "--", "--worktree", "WT-card"}, "--worktree", "plain", true},
		{"13 -- --worktree plain PROMPT", []string{"--", "--worktree", "plain", "PROMPT"}, "--worktree", "plain", true},
		{"14 -w plain --worktree <elsewhere> (duplicate options)", []string{"-w", "plain", "--worktree", elsewhere}, "--worktree", elsewhere, true},
	}
}

// TestLauncherAnchorFinalArgv holds the guard-input cells of SPEC-LAUNCHER-ARGV-001.
// M1 lands view_exec_byte_equality (AC-LA-005); M2 adds the three iter8 families.
func TestLauncherAnchorFinalArgv(t *testing.T) {
	pinSnapshotModel(t)
	// AC-LA-005: for every §D.4 spelling the view's FinalArgv equals, token by
	// token, the argv the exec boundary receives (minus the flags the launcher
	// injects from profile state), and the view carries the expected backend
	// selection. Equality alone would pass a resolver that misparses every
	// value identically on both sides, so the selection is asserted too.
	t.Run("view_exec_byte_equality", func(t *testing.T) {
		elsewhere := filepath.Join(t.TempDir(), "elsewhere")
		for _, arm := range []string{"cc", "glm"} {
			for _, row := range viewEqualityRows(elsewhere) {
				t.Run(arm+"/"+row.name, func(t *testing.T) {
					view := resolveLaunchView(row.args)
					obs := launchWithStageSeams(t, arm, row.args, "bare")
					if obs.Err != "" {
						t.Fatalf("launch refused: %q", obs.Err)
					}
					if !reflect.DeepEqual(view.FinalArgv, obs.ExecTail) {
						t.Errorf("view FinalArgv = %q, exec boundary received %q", view.FinalArgv, obs.ExecTail)
					}
					sel := view.Selection
					if !sel.Present || sel.Spelling != row.spelling || sel.Value != row.value || sel.HasValue != row.hasValue {
						t.Errorf("selection = %+v, want spelling %q value %q hasValue %v", sel, row.spelling, row.value, row.hasValue)
					}
				})
			}
		}
	})
}
