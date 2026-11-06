package cli

// launcher_worktree_stage_cases_test.go — the SPEC-LAUNCHER-ARGV-001 AC-LA-014
// golden table (card t1748, defects D8/D9/D10/D11): what the non-leader launch
// path observably does for each worktree spelling, per reader stage.
//
// The rows were measured on the unrefactored tree (HEAD 6e729ba21) and pinned
// here before any resolver code existed; the refactor is judged by these cells
// staying GREEN. A row that differs between stages is the point of the table:
// the readers must keep their per-stage values, never share one.

import (
	"path/filepath"
	"reflect"
	"testing"
)

// stageRow is one spelling with its pinned per-stage observations.
type stageRow struct {
	name string
	// args are the entry arguments; the token "ELSEWHERE" stands for an
	// absolute path outside every accepted worktree prefix.
	args []string
	// bare run (no tree directories): the launch result.
	launch []string // argv handed to the launch seam
	exec   []string // exec-seam argv tail
	chain  string   // chain-ledger tree ("" = no node)
	mat    []string // --branch materializations
	// precheck is the tree the cc writer precheck read ("" = none read).
	precheck string
	// seed is the tree the hook seeder resolved ("" = none).
	seed string
	// stages names the claim a row pins, for the failure message.
	stages string
}

var stageRows = []stageRow{
	{
		name: "control short -w", args: []string{"-w", "plain"},
		launch: []string{"--worktree", "plain"}, exec: []string{"--worktree", "plain"},
		chain: ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=B=C=plain",
	},
	{
		// §D.4 row 14 (D8): the first value wins in every non-guard reader even
		// though the backend takes the last option.
		name: "duplicate options first value", args: []string{"-w", "plain", "--worktree", "ELSEWHERE"},
		launch: []string{"--worktree", "plain", "--worktree", "ELSEWHERE"},
		exec:   []string{"--worktree", "plain", "--worktree", "ELSEWHERE"},
		chain:  ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=B=C=plain (first), launch allowed",
	},
	{
		// D9: stage A reads no value (the next token is --branch), stage B and
		// C read plain after the --branch consumption.
		name: "bare -w before --branch", args: []string{"-w", "--branch", "main", "plain"},
		launch: []string{"--worktree", "plain"}, exec: []string{"--worktree", "plain"},
		mat:   []string{"plain@main"},
		chain: ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=bare B=plain C=plain",
	},
	{
		// D9: the empty equals form reads "" at stages A and B, while the
		// normalized stream (stage C) carries the next token as the value.
		name: "empty equals form", args: []string{"--worktree=", "plain"},
		launch: []string{"--worktree", "plain"}, exec: []string{"--worktree", "plain"},
		chain: ".claude/worktrees/plain", precheck: "", seed: "",
		stages: "A=B=empty C=plain",
	},
	{
		// D10: stage C reads the consumption loop's passThrough, where the
		// first separator is already gone.
		name: "tail --worktree after separator", args: []string{"--", "--worktree", "plain", "PROMPT"},
		launch: []string{"--", "--worktree", "plain", "PROMPT"},
		exec:   []string{"--worktree", "plain", "PROMPT"},
		chain:  ".claude/worktrees/plain", precheck: "", seed: "",
		stages: "A=B=none C=plain",
	},
	{
		// D10: -b is consumed by the launcher, which brings plain next to the
		// bare --worktree in passThrough.
		name: "consumed option between flag and value", args: []string{"-w", "-b", "plain"},
		launch: []string{"--worktree", "-b", "plain"}, exec: []string{"--worktree", "plain"},
		chain: ".claude/worktrees/plain", precheck: "", seed: "",
		stages: "A=B=bare C=plain",
	},
	{
		// D11: the chain extractor knows only the 2-token and equals long forms.
		name: "tail short -w after separator", args: []string{"--", "-w", "plain", "PROMPT"},
		launch: []string{"--", "-w", "plain", "PROMPT"}, exec: []string{"-w", "plain", "PROMPT"},
		chain: "", precheck: "", seed: "",
		stages: "A=B=C=none",
	},
	{
		name: "tail short equals after separator", args: []string{"--", "-w=plain", "PROMPT"},
		launch: []string{"--", "-w=plain", "PROMPT"}, exec: []string{"-w=plain", "PROMPT"},
		chain: "", precheck: "", seed: "",
		stages: "A=B=C=none",
	},
	{
		name: "long equals form", args: []string{"--worktree=plain"},
		launch: []string{"--worktree", "plain"}, exec: []string{"--worktree", "plain"},
		chain: ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=B=C=plain",
	},
	{
		name: "short equals form", args: []string{"-w=plain"},
		launch: []string{"--worktree", "plain"}, exec: []string{"--worktree", "plain"},
		chain: ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=B=C=plain",
	},
	{
		// The attached short spelling is invisible to every non-guard reader.
		name: "attached short form", args: []string{"-wplain"},
		launch: []string{"-wplain"}, exec: []string{"-wplain"},
		chain: "", precheck: "", seed: "",
		stages: "A=B=C=none",
	},
	{
		name: "tail short cluster", args: []string{"--", "-pwplain", "PROMPT"},
		launch: []string{"--", "-pwplain", "PROMPT"}, exec: []string{"-pwplain", "PROMPT"},
		chain: "", precheck: "", seed: "",
		stages: "A=B=C=none",
	},
	{
		name: "--branch after the value", args: []string{"-w", "plain", "--branch", "main"},
		launch: []string{"--worktree", "plain"}, exec: []string{"--worktree", "plain"},
		mat:   []string{"plain@main"},
		chain: ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=B=C=plain",
	},
	{
		// §D.4 row 12: a second separator in the tail belongs to the backend.
		name: "second separator in the tail", args: []string{"-w", "plain", "--", "P1", "--", "--worktree", "other"},
		launch: []string{"--worktree", "plain", "--", "P1", "--", "--worktree", "other"},
		exec:   []string{"--worktree", "plain", "P1", "--", "--worktree", "other"},
		chain:  ".claude/worktrees/plain", precheck: "plain", seed: "plain",
		stages: "A=B=C=plain (the tail --worktree is behind the backend terminator)",
	},
}

// substituteElsewhere replaces the ELSEWHERE placeholder token.
func substituteElsewhere(args []string, elsewhere string) []string {
	out := make([]string, len(args))
	for i, a := range args {
		if a == "ELSEWHERE" {
			a = elsewhere
		}
		out[i] = a
	}
	return out
}

// assertStageRow runs one row on one arm in the three tree shapes and compares
// every observation against the pinned values.
func assertStageRow(t *testing.T, arm string, row stageRow) {
	t.Helper()
	elsewhere := filepath.Join(t.TempDir(), "elsewhere")
	args := substituteElsewhere(row.args, elsewhere)
	want := func(s []string) []string { return substituteElsewhere(s, elsewhere) }

	// Run 1, bare root: allowed launch, argv at the launch and exec seams, the
	// --branch materializations, the chain node and the child env.
	bare := launchWithStageSeams(t, arm, args, "bare")
	if bare.Err != "" {
		t.Fatalf("[%s] launch refused: %q", row.stages, bare.Err)
	}
	if !reflect.DeepEqual(bare.LaunchArgs, want(row.launch)) {
		t.Errorf("[%s] launch argv = %q, want %q", row.stages, bare.LaunchArgs, want(row.launch))
	}
	if !reflect.DeepEqual(bare.ExecTail, want(row.exec)) {
		t.Errorf("[%s] exec argv tail = %q, want %q", row.stages, bare.ExecTail, want(row.exec))
	}
	if !reflect.DeepEqual(bare.Materialize, row.mat) {
		t.Errorf("[%s] --branch materializations = %q, want %q", row.stages, bare.Materialize, row.mat)
	}
	var wantChain []string
	if row.chain != "" {
		wantChain = []string{row.chain}
	}
	if len(bare.ChainTrees) != len(wantChain) || (len(wantChain) == 1 && bare.ChainTrees[0] != wantChain[0]) {
		t.Errorf("[%s] chain nodes = %q, want %q", row.stages, bare.ChainTrees, wantChain)
	}
	if bare.EnvChainNode != (row.chain != "") {
		t.Errorf("[%s] MOAI_CHAIN_NODE_ID in child env = %v, want %v", row.stages, bare.EnvChainNode, row.chain != "")
	}

	// Run 2, trees present in a non-git root: the cc writer precheck refuses on
	// the tree it reads; the glm flow has no precheck and always launches.
	nongit := launchWithStageSeams(t, arm, args, "dirs-nongit")
	switch {
	case arm == "cc" && row.precheck != "":
		if nongit.Err != "writer-anchored" || nongit.PrecheckTree != ".claude/worktrees/"+row.precheck {
			t.Errorf("[%s] precheck = err %q tree %q, want writer-anchored on %q", row.stages, nongit.Err, nongit.PrecheckTree, ".claude/worktrees/"+row.precheck)
		}
	default:
		if nongit.Err != "" {
			t.Errorf("[%s] %s with trees present: unexpected refusal %q (tree %q)", row.stages, arm, nongit.Err, nongit.PrecheckTree)
		}
	}

	// Run 3, trees present in a git root: the precheck passes and the seeder
	// resolves the tree the stage-B value names.
	git := launchWithStageSeams(t, arm, args, "dirs-git")
	if git.Err != "" {
		t.Fatalf("[%s] %s in a git root: refused %q", row.stages, arm, git.Err)
	}
	wantSeed := ""
	if row.seed != "" {
		wantSeed = ".claude/worktrees/" + row.seed
	}
	if git.SeedTree != wantSeed {
		t.Errorf("[%s] seeded tree = %q, want %q", row.stages, git.SeedTree, wantSeed)
	}
}

// TestLauncherWorktreeStageCharacterization pins the per-stage reads of the
// non-leader launch path for both arms (SPEC §2 literal preservation, AC-LA-014).
func TestLauncherWorktreeStageCharacterization(t *testing.T) {
	for _, arm := range []string{"cc", "glm"} {
		for _, row := range stageRows {
			t.Run(arm+"/"+row.name, func(t *testing.T) {
				assertStageRow(t, arm, row)
			})
		}
	}
}

// TestLauncherWorktreeDuplicateOptionFirstValue is the AC-LA-014 regression cell
// for plan-audit D8 (§D.4 row 14): the non-leader `-w plain --worktree <path
// outside every worktree prefix>` stays ALLOWED, and the five non-guard readers
// keep reading the first value `plain` — never the last option the backend takes.
func TestLauncherWorktreeDuplicateOptionFirstValue(t *testing.T) {
	elsewhere := filepath.Join(t.TempDir(), "elsewhere")
	for _, arm := range []string{"cc", "glm"} {
		t.Run(arm, func(t *testing.T) {
			args := []string{"-w", "plain", "--worktree", elsewhere}
			bare := launchWithStageSeams(t, arm, args, "bare")
			if bare.Err != "" {
				t.Fatalf("a non-leader duplicate-option launch must stay allowed, got %q", bare.Err)
			}
			if len(bare.ChainTrees) != 1 || bare.ChainTrees[0] != ".claude/worktrees/plain" {
				t.Errorf("chain-node reader tree = %q, want the first value plain", bare.ChainTrees)
			}
			if want := []string{"--worktree", "plain", "--worktree", elsewhere}; !reflect.DeepEqual(bare.ExecTail, want) {
				t.Errorf("exec argv tail = %q, want %q (both options reach the backend)", bare.ExecTail, want)
			}
			git := launchWithStageSeams(t, arm, args, "dirs-git")
			if git.Err != "" || git.SeedTree != ".claude/worktrees/plain" {
				t.Errorf("hook seeder: err %q tree %q, want allowed on the first value plain", git.Err, git.SeedTree)
			}
			if arm == "cc" {
				nongit := launchWithStageSeams(t, arm, args, "dirs-nongit")
				if nongit.PrecheckTree != ".claude/worktrees/plain" {
					t.Errorf("writer precheck read %q, want the first value plain", nongit.PrecheckTree)
				}
			}
		})
	}
}
