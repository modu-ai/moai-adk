package factorylane

// merge_readiness_pr_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-B (card t1453),
// AC-GFD-005: the condition triple, evaluated against the github-flow
// integration target (`main`) on real repositories, answers the four fixture
// shapes the PR edge relies on before it opens anything. The edge itself
// (refuse, open nothing) is the same-named test in internal/cli; this half
// pins the triple's verdicts it consumes. It is a guard row — green on arrival,
// because EvaluateMergeTriple takes its target as a parameter and the M2-B
// change is in the caller.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeReadinessBeforePR(t *testing.T) {
	type fixture struct {
		name       string
		syncStatus string // "" = no §E.4 section
		mainMoves  string // "" | "conflict" | "unrelated"
		failing    string // "" = the triple passes
	}
	visited := 0
	for _, tc := range []fixture{
		{"conflicting_card", "complete", "conflict", CheckConflictFree},
		{"conflict_free_card", "complete", "", ""},
		{"no_sync_audit_pass_record", "", "", CheckSyncAudit},
		{"broken_tree_identity", "complete", "unrelated", CheckTreeIdentity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			visited++
			dir := t.TempDir()
			git := func(args ...string) {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
			}
			write := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q", "-b", "main")
			git("config", "user.email", "test@example.com")
			git("config", "user.name", "Test")
			write("base.txt", "base\n")
			git("add", ".")
			git("commit", "-q", "-m", "base")
			git("checkout", "-q", "-b", "WT-card")
			write("base.txt", "base\ncard edit\n")
			spec := filepath.Join(dir, ".moai", "specs", "SPEC-X-001")
			if err := os.MkdirAll(spec, 0o755); err != nil {
				t.Fatal(err)
			}
			if tc.syncStatus != "" {
				write(filepath.Join(".moai", "specs", "SPEC-X-001", "progress.md"),
					"## §E.4 Sync-phase Audit-Ready Signal\n\nsync_status: "+tc.syncStatus+"\n")
			}
			git("add", ".")
			git("commit", "-q", "-m", "card")
			git("checkout", "-q", "main")
			switch tc.mainMoves {
			case "conflict":
				write("base.txt", "base\nmain edit\n")
				git("commit", "-q", "-am", "main edits the same line")
			case "unrelated":
				write("other.txt", "other\n")
				git("add", ".")
				git("commit", "-q", "-m", "main adds a file")
			}
			// The lane evaluates from its card worktree: the card branch is checked
			// out, and its progress.md is the sync phase record the triple reads.
			git("checkout", "-q", "WT-card")

			run, err := EvaluateMergeTriple(MergeTripleInput{
				Lane: "lane-1", Card: "t1", SpecDir: spec, Branch: "WT-card", Develop: "main", RepoDir: dir,
			}, ExecGitRunner{Dir: dir})
			if err != nil {
				t.Fatalf("EvaluateMergeTriple: %v", err)
			}
			if tc.failing == "" {
				if !run.AllPassed {
					t.Fatalf("a clean card failed %s: %+v", run.FailedCondition, run.Checks)
				}
				return
			}
			if run.AllPassed || run.FailedCondition != tc.failing {
				t.Fatalf("all_passed=%v failed_condition=%q, want %q (checks %+v)", run.AllPassed, run.FailedCondition, tc.failing, run.Checks)
			}
			var detail string
			for _, c := range run.Checks {
				if c.Name == tc.failing {
					detail = c.Detail
				}
			}
			if strings.TrimSpace(detail) == "" {
				t.Errorf("the failing condition carries no reason text")
			}
		})
	}
	if visited != 4 {
		t.Fatalf("visited %d fixtures, want 4", visited)
	}
}
