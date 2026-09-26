package hook

import (
	"context"
	"encoding/json"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// hookSet is the four handlers the detector is wired into.
type hookSet struct{ pre, post, failure, stop Handler }

func escalationHookSet(t *testing.T, w *escalationtest.Worktree) hookSet {
	t.Helper()
	p := escalationProvider(t, w)
	return hookSet{
		pre:     NewPreToolHandler(p, DefaultSecurityPolicy()),
		post:    WithEscalationConfig(NewPostToolHandlerWithMxValidatorAndTimeout(nil, nil, "", 0), p),
		failure: WithEscalationConfig(NewPostToolUseFailureHandler(), p),
		stop:    WithEscalationConfig(NewStopHandler(), p),
	}
}

// bashInput builds a Bash hook input.
func escBashInput(w *escalationtest.Worktree, cmd string) *HookInput {
	return escalationInput(w, "Bash", map[string]string{"command": cmd})
}

// AC-AE-013 (REQ-AE-011): push, tag, release, and a denylisted command each
// write irreversible-action before execution; the denylisted command's deny
// output is byte-identical with the detector off; a push of develop is
// authorized by push-develop and trips only without it.
func TestIrreversibleActionTrips(t *testing.T) {
	run := func(t *testing.T, edit func(string) string, mode string, cmds []string) (*escalationtest.Worktree, []string) {
		t.Helper()
		w := escalationHookFixture(t, "")
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Edit: edit})
		pre, _ := escalationHandlersMode(t, w, mode)
		if _, err := pre.Handle(context.Background(), escalationInput(w, "Write",
			map[string]string{"file_path": w.Path("internal/fixture/a.go"), "content": "x\n"})); err != nil {
			t.Fatal(err)
		}
		var outs []string
		for _, c := range cmds {
			out, err := pre.Handle(context.Background(), escBashInput(w, c))
			if err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(out)
			outs = append(outs, string(b))
		}
		return w, outs
	}
	irreversible := []string{"git push origin main", "git push origin v1.0.0", "git tag v1.0.0", "gh release create v1.0.0", "rm -rf /"}

	w, contractOuts := run(t, nil, "", append(irreversible, "git push origin develop"))
	recs := 0
	for _, n := range escalationRecordNames(t, w) {
		if strings.HasPrefix(n, escalation.ClassIrreversibleAction+"-") {
			recs++
		}
	}
	if recs != 5 {
		t.Errorf("irreversible-action records = %d (%v), want 5", recs, escalationRecordNames(t, w))
	}
	_, guidedOuts := run(t, nil, "guided", append(irreversible, "git push origin develop"))
	for i := range contractOuts {
		if contractOuts[i] != guidedOuts[i] {
			t.Errorf("output for %q differs from the detector-off output:\n got %s\nwant %s",
				append(irreversible, "git push origin develop")[i], contractOuts[i], guidedOuts[i])
		}
	}
	if !strings.Contains(contractOuts[4], `"deny"`) {
		t.Errorf("the denylisted command was not denied: %s", contractOuts[4])
	}

	w2, _ := run(t, func(d string) string { return strings.Replace(d, "  - push-develop\n", "", 1) }, "",
		[]string{"git push origin develop"})
	recs = 0
	for _, n := range escalationRecordNames(t, w2) {
		if strings.HasPrefix(n, escalation.ClassIrreversibleAction+"-") {
			recs++
		}
	}
	if recs != 1 {
		t.Errorf("develop push without push-develop: %d records, want 1", recs)
	}
}

// snapshotRel lists the files under root relative to it, skipping .git.
func snapshotRel(t *testing.T, root string) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, p)
			out[filepath.ToSlash(rel)] = true
		}
		return nil
	})
	return out
}

// AC-AE-004 (REQ-AE-003, REQ-AE-021): with every class tripped, no hook
// output differs from the guided-mode output of the same input, no call
// waits, the detector writes only records, the card log, and the card state,
// and no detector code references the user question channel.
func TestDetectorNeverAltersToolCall(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	type result struct {
		outs  []string
		files map[string]bool
		home  map[string]bool
		w     *escalationtest.Worktree
	}
	runOnce := func(t *testing.T, mode string) result {
		home := t.TempDir()
		t.Setenv(config.EnvHome, home)
		w := escalationtest.NewGitWorktree(t, "t9001")
		if mode == "contract" {
			w.WriteContractMode("")
		} else {
			w.WriteMode("guided", false)
		}
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Edit: func(d string) string {
			return strings.Replace(d, "\nescalate_on:", "\nbudget:\n  turns: 1\n  operations: 3\n  audit_retries: 0\n\nescalate_on:", 1)
		}})
		w.Write("internal/foo/foo.go", "package foo\n\nfunc Existing() {}\n")
		w.Write("internal/fixture/y.go", "package fixture\n")
		w.Commit("base")
		w.Git("checkout", "-q", "-b", "WT-x")
		t.Setenv(config.EnvClaudeProjectDir, w.Root)
		hs := escalationHookSet(t, w)

		var outs []string
		call := func(h Handler, in *HookInput) {
			start := time.Now()
			out, err := h.Handle(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			if d := time.Since(start); d > 5*time.Second {
				t.Errorf("hook call took %s", d)
			}
			b, _ := json.Marshal(out)
			outs = append(outs, string(b))
		}
		write := func(rel string) *HookInput {
			return escalationInput(w, "Write", map[string]string{"file_path": w.Path(rel), "content": "x\n"})
		}
		failing := func(cmd string) *HookInput {
			in := escBashInput(w, cmd)
			in.Error = "Exit code 1\n--- FAIL: TestX"
			return in
		}

		call(hs.pre, write("internal/fixture/a.go"))          // arms
		call(hs.pre, write("internal/bar/x.go"))              // 3 ownership-move
		call(hs.pre, escBashInput(w, "git push origin main")) // 6 irreversible-action
		call(hs.pre, escBashInput(w, "rm -rf /"))             // 6 + existing deny
		for i := 0; i < 3; i++ {
			call(hs.failure, failing("go test ./...")) // 2 invariant, 8 same-diagnostic
		}
		call(hs.post, escBashInput(w, "ls")) // 7 operations 4 > 3
		call(hs.stop, &HookInput{SessionID: "s", CWD: w.Path("internal")})
		call(hs.stop, &HookInput{SessionID: "s", CWD: w.Path("internal")}) // 7 turns 2 > 1
		w.Write(".moai/state/audit-multi/s.json", `{"per_backend_verdicts":[],"overall_verdict":"pass","disagreement_flag":true}`)
		w.Write(".moai/reports/t9001/plan-audit-iter1.md", "Verdict: FAIL\n")
		call(hs.post, escBashInput(w, "git commit -m one")) // 5 contradictory, 9 audit cap
		w.Write("internal/foo/foo.go", "package foo\n\nfunc Existing() {}\n\nfunc Added() {}\n")
		w.Commit("api")
		settings := config.ResolveAutonomy(mustLoad(t, w).Workflow)
		escalation.Checkpoint(settings, w.Root) // 4 new-architecture-or-api
		w.Replace(".moai/specs/SPEC-A-001/acceptance.md", "then it passes.", "then it passed.")
		call(hs.post, escBashInput(w, "git commit -m two")) // 1 acceptance-change, 10 disarm

		return result{outs: outs, files: snapshotRel(t, w.Root), home: snapshotRel(t, home), w: w}
	}

	contract := runOnce(t, "contract")
	guided := runOnce(t, "guided")

	for i := range contract.outs {
		if contract.outs[i] != guided.outs[i] {
			t.Errorf("hook call %d output differs:\n contract %s\n   guided %s", i, contract.outs[i], guided.outs[i])
		}
	}

	names := escalationRecordNames(t, contract.w)
	for _, class := range []string{
		escalation.ClassAcceptanceChange, escalation.ClassInvariantViolation, escalation.ClassOwnershipMove,
		escalation.ClassNewArchitectureOrAPI, escalation.ClassContradictoryEvidence, escalation.ClassIrreversibleAction,
		escalation.ClassBudgetExceeded, escalation.ClassSameDiagnosticRepeat, escalation.ClassAuditFailAtRetryCap,
		escalation.ClassDetectionDisarmed,
	} {
		found := false
		for _, n := range names {
			found = found || strings.HasPrefix(n, class+"-")
		}
		if !found {
			t.Errorf("class %s did not trip; records: %v", class, names)
		}
	}

	for f := range contract.files {
		if !guided.files[f] && !strings.HasPrefix(f, ".moai/reports/t9001/escalation/") {
			t.Errorf("detector wrote an unexpected worktree file: %s", f)
		}
	}
	for f := range contract.home {
		cardFile := strings.HasSuffix(f, ".log.jsonl") || strings.HasSuffix(f, ".json") || strings.HasSuffix(f, ".lock")
		if !guided.home[f] && (!strings.Contains(f, "/contract/escalation/") || !cardFile) {
			t.Errorf("detector wrote an unexpected MOAI_HOME file: %s", f)
		}
	}

	// No detector code path references the user question channel.
	fset := token.NewFileSet()
	pkgDir := filepath.Join("..", "escalation")
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(pkgDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parser.ParseFile(fset, e.Name(), src, parser.ImportsOnly); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "AskUserQuestion") || strings.Contains(string(src), "mcp__askuser") {
			t.Errorf("%s references the user question channel", e.Name())
		}
	}
}

func mustLoad(t *testing.T, w *escalationtest.Worktree) *config.Config {
	t.Helper()
	cfg, err := config.NewLoader().Load(w.Path(".moai"))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}
