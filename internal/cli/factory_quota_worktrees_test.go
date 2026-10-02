// factory_quota_worktrees_test.go — SPEC-QUOTA-RECORD-WORKTREES-001 M1 AC tests
// (card t1442): the quota gate sees a reading that exists only in a linked
// worktree's record directory. AC-QWR-007 (a disabled gate reads and enumerates
// nothing), -008 (the lane gate, the MCP twin, the status block, and the
// acquire warning see a worktree-only reading), -009 (the same reading renders
// identically wherever it was found), -010b (a root without worktree metadata
// reads its own directory alone), and -013b (the production seam applies the
// configured bound).
//
// Every fixture is built under t.TempDir(). A cli fixture that needs a
// registered linked worktree makes it with a real `git worktree add` inside the
// initGitRepo-style temporary repository, because the cli path runs git through
// factoryAssertParentCheckout. ./internal/cli runs only through the anchored
// -run selectors naming one of these tests; the lane environment this suite may
// run in is cleared per test by the predecessor's fixtures (sdClearLaneEnv,
// qasLaneEnv), so it cannot leak into a verdict.
package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/statusline"
)

// The signatures of the two seams: the aggregator seam keeps its single-directory
// shape (AC-QWR-003 signature_unchanged, cli half) and the enumerator variable
// has the shape of statusline.QuotaStateDirs. A change to either fails to
// compile here.
var (
	_ func(string, time.Time, time.Duration) statusline.QuotaAggregate = factoryQuotaAggregate
	_ func(string, int) []string                                       = factoryQuotaStateDirs
)

// qwrStatuslinePkg is the import path whose QuotaStateDirs the static check
// pins to one reference.
const qwrStatuslinePkg = "github.com/modu-ai/moai-adk/internal/statusline"

// qwrAddWorktree makes a real linked worktree of the repository at root, named
// name (the leaf directory, which git also uses as the metadata entry name),
// under a fresh temporary directory, and returns its path.
func qwrAddWorktree(t *testing.T, root, name string) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), name)
	runGitIn(t, root, "worktree", "add", "-q", "-b", "qwr-"+name, wt)
	return wt
}

// qwrWriteGate writes root's workflow.yaml with the gate enabled or not, plus
// any extra quota_gate lines (already indented to the block).
func qwrWriteGate(t *testing.T, root string, enabled bool, extra string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n    quota_gate:\n        enabled: " + strconv.FormatBool(enabled) + "\n" + extra
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// qwrWorktreeOnly is the AC-QWR-008 fixture: qasFixture with no record in the
// primary directory, plus one linked worktree whose record directory holds the
// one fresh reading (five). It returns the root, the worktree directory, and the
// queue store.
func qwrWorktreeOnly(t *testing.T, o qasFixtureOpts, five *statusline.QuotaWindowRecord) (string, string, *kanban.BacklogStore) {
	t.Helper()
	o.noRecord = true
	root, store := qasFixture(t, o)
	wt := qwrAddWorktree(t, root, "lane-wt")
	if five != nil {
		qasWriteRecord(t, wt, "sess-wt", fcNow, five, nil)
	}
	return root, wt, store
}

// qwrCalls counts the calls the two seams receive.
type qwrCalls struct{ aggregate, enumerate int }

// qwrCountCalls wraps the aggregator seam and the enumerator variable with
// counters that delegate to the values in force (the production ones unless a
// test replaced them), and restores both at cleanup.
func qwrCountCalls(t *testing.T) *qwrCalls {
	t.Helper()
	c := &qwrCalls{}
	prevAgg, prevDirs := factoryQuotaAggregate, factoryQuotaStateDirs
	factoryQuotaAggregate = func(stateDir string, now time.Time, maxAge time.Duration) statusline.QuotaAggregate {
		c.aggregate++
		return prevAgg(stateDir, now, maxAge)
	}
	factoryQuotaStateDirs = func(root string, maxDirs int) []string {
		c.enumerate++
		return prevDirs(root, maxDirs)
	}
	t.Cleanup(func() { factoryQuotaAggregate, factoryQuotaStateDirs = prevAgg, prevDirs })
	return c
}

// qwrPkgAliases maps each local import name of a parsed file to its import path.
func qwrPkgAliases(f *ast.File) map[string]string {
	pkgs := map[string]string{}
	for _, imp := range f.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		name := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		pkgs[name] = path
	}
	return pkgs
}

// qwrEnumeratorUses scans one Go source for the selector
// statusline.QuotaStateDirs (import aliases resolved): varInit counts the uses
// inside the initializer of the package variable factoryQuotaStateDirs, other
// counts every remaining use (function bodies and any other declaration), and
// directReads lists the os.ReadDir, os.Open, and os.ReadFile selectors in the
// body of factoryQuotaEvaluate. src nil reads filename.
func qwrEnumeratorUses(t *testing.T, filename string, src any) (varInit, other int, directReads []string) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), filename, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	pkgs := qwrPkgAliases(f)
	selects := func(n ast.Node, importPath string, names ...string) (string, bool) {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return "", false
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || pkgs[id.Name] != importPath {
			return "", false
		}
		for _, name := range names {
			if sel.Sel.Name == name {
				return name, true
			}
		}
		return "", false
	}
	count := func(root ast.Node) int {
		n := 0
		ast.Inspect(root, func(node ast.Node) bool {
			if _, ok := selects(node, qwrStatuslinePkg, "QuotaStateDirs"); ok {
				n++
			}
			return true
		})
		return n
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if ok && len(vs.Names) == 1 && vs.Names[0].Name == "factoryQuotaStateDirs" {
					for _, v := range vs.Values {
						varInit += count(v)
					}
					continue
				}
				other += count(spec)
			}
		case *ast.FuncDecl:
			other += count(d)
			if d.Name.Name == "factoryQuotaEvaluate" && d.Body != nil {
				ast.Inspect(d.Body, func(node ast.Node) bool {
					if name, ok := selects(node, "os", "ReadDir", "Open", "ReadFile"); ok {
						directReads = append(directReads, "os."+name)
					}
					return true
				})
			}
		}
	}
	return varInit, other, directReads
}

// AC-QWR-007 — with the gate disabled nothing is read and nothing is
// enumerated.
func TestQWR_AC007_GateDisabledReadsNothing(t *testing.T) {
	// fixture: a git root with one real linked worktree holding a fresh high
	// record, the gate written enabled or not, and a Claude lane.
	fixture := func(t *testing.T, enabled bool) string {
		t.Helper()
		sdClearLaneEnv(t)
		root, _ := sdMoaiFixture(t)
		qwrWriteGate(t, root, enabled, "")
		wt := qwrAddWorktree(t, root, "lane-wt")
		qasWriteRecord(t, wt, "sess-wt", fcNow, qasWin(92, qasReset5), qasWin(96, qasReset7))
		qasLaneEnv(t, "lane-1", kanban.BackendClaude)
		t.Chdir(root)
		return root
	}

	t.Run("disabled_zero_calls", func(t *testing.T) {
		root := fixture(t, false)
		qasLaneSeam(t)
		calls := qwrCountCalls(t)
		if ev := factoryQuotaEvaluate(root); ev.Enabled || len(ev.Windows) != 0 {
			t.Errorf("evaluation under a disabled gate = %+v, want disabled with no windows", ev)
		}
		if held, line := (&factoryQuotaLatch{}).evaluate(root); held || line != "" {
			t.Errorf("lane gate under a disabled gate held = %v with %q", held, line)
		}
		if block := factoryQuotaStatusBlock(root); block != nil {
			t.Errorf("status block under a disabled gate = %+v, want nil", block)
		}
		text, js := qasStatus(t)
		if _, present := qasStatusQuota(t, js); present || strings.Contains(text, "quota") {
			t.Errorf("status output carries quota text under a disabled gate:\n%s\n%s", text, js)
		}
		if calls.aggregate != 0 || calls.enumerate != 0 {
			t.Errorf("seam calls under a disabled gate: aggregator %d, enumerator %d, want 0 and 0", calls.aggregate, calls.enumerate)
		}
	})

	t.Run("enabled_one_call_each", func(t *testing.T) {
		// The positive control: the same fixture with the gate enabled calls each
		// seam exactly once per evaluation, on every consumer.
		root := fixture(t, true)
		qasLaneSeam(t)
		calls := qwrCountCalls(t)
		for _, c := range []struct {
			name string
			run  func()
		}{
			{"evaluation", func() { factoryQuotaEvaluate(root) }},
			{"lane gate", func() { (&factoryQuotaLatch{}).evaluate(root) }},
			{"status block", func() { factoryQuotaStatusBlock(root) }},
		} {
			*calls = qwrCalls{}
			c.run()
			if calls.aggregate != 1 || calls.enumerate != 1 {
				t.Errorf("%s: seam calls = aggregator %d, enumerator %d, want 1 and 1", c.name, calls.aggregate, calls.enumerate)
			}
		}
	})

	t.Run("enumerator_referenced_only_by_the_variable", func(t *testing.T) {
		// Positive control: the scan flags an aliased reference in a function
		// body and the three direct reads in factoryQuotaEvaluate.
		const control = `package control
import (
	sl "github.com/modu-ai/moai-adk/internal/statusline"
	osx "os"
)
var factoryQuotaStateDirs = sl.QuotaStateDirs
func elsewhere() { _ = sl.QuotaStateDirs }
func factoryQuotaEvaluate(root string) {
	_, _ = osx.ReadDir(root)
	_, _ = osx.Open(root)
	_, _ = osx.ReadFile(root)
}
`
		if v, o, reads := qwrEnumeratorUses(t, "control.go", control); v != 1 || o != 1 || len(reads) != 3 {
			t.Fatalf("positive control: variable-initializer uses %d, other uses %d, direct reads %v; want 1, 1, and three reads", v, o, reads)
		}
		files, err := filepath.Glob("*.go")
		if err != nil {
			t.Fatal(err)
		}
		var varInit, other int
		var reads []string
		scanned := 0
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			raw, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(raw), "QuotaStateDirs") && !strings.Contains(string(raw), "factoryQuotaEvaluate") {
				continue
			}
			scanned++
			v, o, r := qwrEnumeratorUses(t, file, string(raw))
			varInit, other, reads = varInit+v, other+o, append(reads, r...)
		}
		if scanned == 0 {
			t.Fatal("the scan matched no non-test file of internal/cli")
		}
		if varInit != 1 || other != 0 {
			t.Errorf("statusline.QuotaStateDirs is referenced %d time(s) as the variable's initializer and %d time(s) elsewhere, want exactly 1 and 0", varInit, other)
		}
		if len(reads) != 0 {
			t.Errorf("the body of factoryQuotaEvaluate reads the filesystem directly: %v", reads)
		}
	})

	t.Run("missing_dir_fixture_untouched", func(t *testing.T) {
		// A registered worktree whose directory is gone, the gate disabled, and
		// seams that fail the test if called: the evaluation returns untouched.
		sdClearLaneEnv(t)
		root, _ := sdMoaiFixture(t)
		qwrWriteGate(t, root, false, "")
		wt := qwrAddWorktree(t, root, "gone-wt")
		if err := os.RemoveAll(wt); err != nil {
			t.Fatal(err)
		}
		prevAgg, prevDirs := factoryQuotaAggregate, factoryQuotaStateDirs
		factoryQuotaAggregate = func(string, time.Time, time.Duration) statusline.QuotaAggregate {
			t.Fatal("the aggregator seam was called while the gate is disabled")
			return statusline.QuotaAggregate{}
		}
		factoryQuotaStateDirs = func(string, int) []string {
			t.Fatal("the enumerator was called while the gate is disabled")
			return nil
		}
		t.Cleanup(func() { factoryQuotaAggregate, factoryQuotaStateDirs = prevAgg, prevDirs })
		if ev := factoryQuotaEvaluate(root); ev.Enabled || len(ev.Windows) != 0 {
			t.Errorf("evaluation = %+v, want disabled with no windows", ev)
		}
	})
}

// AC-QWR-008 — the surfaces see a worktree-only reading, with unchanged output
// form.
func TestQWR_AC008_SurfacesSeeWorktreeReading(t *testing.T) {
	wantHold := "quota hold: five_hour used=92.0% resets_at=" + time.Unix(qasReset5, 0).UTC().Format(time.RFC3339)

	t.Run("next_holds_on_worktree_only_92", func(t *testing.T) {
		root, _, store := qwrWorktreeOnly(t, qasFixtureOpts{}, qasWin(92, qasReset5))
		recordBefore, queueBefore := qasRecordDump(t, root), sdQueueBytes(t, store)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err == nil {
			t.Fatalf("want hold line, got a leased card (stdout %q, stderr %q)", out, stderr)
		}
		if line := qasAssertHeld(t, out, stderr, err); line != wantHold {
			t.Errorf("hold line = %q, want %q", line, wantHold)
		}
		if got := qasRecordDump(t, root); got != recordBefore {
			t.Errorf("factory record changed under a hold:\nbefore:\n%s\nafter:\n%s", recordBefore, got)
		}
		if got := sdQueueBytes(t, store); got != queueBefore {
			t.Errorf("queue changed under a hold")
		}
	})

	t.Run("mcp_next_holds_on_worktree_only_92", func(t *testing.T) {
		root, _, _ := qwrWorktreeOnly(t, qasFixtureOpts{}, qasWin(92, qasReset5))
		text, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": root, "run": fcRun})
		if err != nil {
			t.Errorf("a hold is an error result: %v", err)
		}
		if strings.Contains(text, "no card is available") {
			t.Errorf("a hold returned the empty-queue text: %q", text)
		}
		if text != wantHold {
			t.Errorf("held result text = %q, want exactly the hold line %q", text, wantHold)
		}
	})

	t.Run("assigned_card_still_leased", func(t *testing.T) {
		root, _, _ := qwrWorktreeOnly(t, qasFixtureOpts{assigned: true}, qasWin(92, qasReset5))
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		sdAssertLeasedOutput(t, out, "t4", homestate.CardRun, "t4")
		if cd := fcCard(t, root, "t4"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t4 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
	})

	t.Run("leases_when_no_record_anywhere", func(t *testing.T) {
		root, _, _ := qwrWorktreeOnly(t, qasFixtureOpts{}, nil)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if cd := fcCard(t, root, "t2"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t2 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
	})

	t.Run("status_block_pressure", func(t *testing.T) {
		qwrWorktreeOnly(t, qasFixtureOpts{}, qasWin(92, qasReset5))
		qasLaneSeam(t)
		text, js := qasStatus(t)
		quota, present := qasStatusQuota(t, js)
		if !present {
			t.Fatalf("status --json carries no quota block for a worktree-only 92%% reading:\n%s", js)
		}
		if quota["pressure"] != true {
			t.Errorf("status quota pressure = %v, want true", quota["pressure"])
		}
		if !strings.Contains(text, "quota five_hour: state=fresh used=92.0%") {
			t.Errorf("status text carries no five_hour fresh 92.0%% line:\n%s", text)
		}
	})

	t.Run("acquire_warning_for_claude_caller", func(t *testing.T) {
		control := qasAcquire(t, qasAcquireRoot(t, kanban.BackendClaude, nil, nil, qasFixtureOpts{gateOff: true}))
		root := qasAcquireRoot(t, kanban.BackendClaude, nil, nil, qasFixtureOpts{})
		wt := qwrAddWorktree(t, root, "lane-wt")
		qasWriteRecord(t, wt, "sess-wt", fcNow, qasWin(92, qasReset5), nil)
		got := qasAcquire(t, root)
		var lines []string
		for _, l := range strings.Split(strings.TrimRight(got.stderr, "\n"), "\n") {
			if strings.Contains(l, "quota") {
				lines = append(lines, l)
			}
		}
		if len(lines) != 1 || !qasAcquireWarningRE.MatchString(lines[0]) {
			t.Fatalf("stderr quota lines = %q, want exactly one matching %s", lines, qasAcquireWarningRE)
		}
		if got.err != nil {
			t.Errorf("acquire failed: %v (a warning must never refuse the window)", got.err)
		}
		if got.stdout != control.stdout {
			t.Errorf("stdout differs from the gate-disabled run:\n got: %q\nwant: %q", got.stdout, control.stdout)
		}
		if got.lock != control.lock {
			t.Errorf("lock record differs from the gate-disabled run:\n got: %s\nwant: %s", got.lock, control.lock)
		}
	})

	t.Run("wait_recheck_sees_worktree_update", func(t *testing.T) {
		sdClearLaneEnv(t)
		root, store := sdMoaiFixture(t)
		fcQueue(t, store, kanban.BacklogStateQueued)
		sdRegisterLane(t, root, "lane-1")
		qasEnableGate(t, root)
		wt := qwrAddWorktree(t, root, "lane-wt")
		qasWriteRecord(t, wt, "sess-wt", fcNow, qasWin(92, qasReset5), nil)
		qasLaneEnv(t, "lane-1", kanban.BackendClaude)
		t.Chdir(root)
		// After the first wait the worktree's session renders again at 80%, below
		// hold minus margin (90 - 5), captured at the advanced clock.
		sleeps := qasFakeClock(t, nil, func(n int) {
			if n == 1 {
				qasWriteRecord(t, wt, "sess-wt", fcNow.Add(5*time.Second), qasWin(80, qasReset5), nil)
			}
		})
		out, stderr, err := qasRunNext(t, "--wait", "--wait-bound", "1h", "--run", fcRun)
		if err != nil {
			t.Fatalf("next --wait: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if *sleeps != 1 {
			t.Errorf("wait slept %d times, want 1 (held at 92 on the first pass, released at 80 on the second)", *sleeps)
		}
		if cd := fcCard(t, root, "t1"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t1 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
	})
}

// AC-QWR-009 — the same reading renders identically wherever it was found: the
// primary-directory fixture is the positive control (it must render quota
// output at all) and the worktree-directory fixture must render the same bytes.
func TestQWR_AC009_WorktreeSourceRendersIdentically(t *testing.T) {
	// identical compares the two renderings and fails when the primary one is
	// empty of quota output (a vacuous comparison).
	identical := func(t *testing.T, what, primary, worktree string) {
		t.Helper()
		if !strings.Contains(primary, "quota") {
			t.Fatalf("positive control: the primary-directory fixture rendered no quota output for %s: %q", what, primary)
		}
		if primary != worktree {
			t.Errorf("%s differs between the primary-directory and worktree-directory fixtures:\nprimary:  %q\nworktree: %q", what, primary, worktree)
		}
	}
	// place writes the one fresh record at the root's own directory or at a
	// linked worktree's; both fixtures carry byte-identical record files.
	place := func(t *testing.T, root string, inWorktree bool) {
		t.Helper()
		dir := root
		if inWorktree {
			dir = qwrAddWorktree(t, root, "lane-wt")
		}
		qasWriteRecord(t, dir, "sess-same", fcNow, qasWin(92, qasReset5), qasWin(96, qasReset7))
	}
	laneFixture := func(t *testing.T, inWorktree bool) {
		t.Helper()
		root, _ := qasFixture(t, qasFixtureOpts{noRecord: true})
		place(t, root, inWorktree)
		qasLaneSeam(t)
	}

	t.Run("status_text", func(t *testing.T) {
		var out [2]string
		for i, inWorktree := range []bool{false, true} {
			laneFixture(t, inWorktree)
			out[i], _ = qasStatus(t)
		}
		identical(t, "status text", out[0], out[1])
	})
	t.Run("status_json", func(t *testing.T) {
		var out [2]string
		for i, inWorktree := range []bool{false, true} {
			laneFixture(t, inWorktree)
			_, out[i] = qasStatus(t)
		}
		identical(t, "status --json", out[0], out[1])
	})
	t.Run("hold_line", func(t *testing.T) {
		var out [2]string
		for i, inWorktree := range []bool{false, true} {
			laneFixture(t, inWorktree)
			stdout, stderr, err := qasRunNext(t, "--run", fcRun)
			out[i] = strings.Join([]string{stdout, stderr, qwrErrString(err)}, "|")
		}
		identical(t, "next output (stdout|stderr|error)", out[0], out[1])
	})
	t.Run("acquire_warning", func(t *testing.T) {
		var out [2]string
		for i, inWorktree := range []bool{false, true} {
			root := qasAcquireRoot(t, kanban.BackendClaude, nil, nil, qasFixtureOpts{})
			place(t, root, inWorktree)
			got := qasAcquire(t, root)
			out[i] = strings.Join([]string{got.stdout, got.stderr, got.lock}, "|")
		}
		identical(t, "acquire output (stdout|stderr|lock)", out[0], out[1])
	})
}

// qwrErrString renders an error for a comparison, "<nil>" for none.
func qwrErrString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

// AC-QWR-010 (cli complement) — a root whose .git is a directory with no
// worktrees inside reads only its own record directory and leases as before.
// The fixtures are real git repositories; the no_git_entry and git_is_a_file
// shapes are exercised at the statusline layer only.
func TestQWR_AC010b_CliRootWithoutGitMetadata(t *testing.T) {
	t.Run("unregistered_directory_is_not_read", func(t *testing.T) {
		root, _ := qasFixture(t, qasFixtureOpts{noRecord: true})
		if _, err := os.Stat(filepath.Join(root, ".git", "worktrees")); err == nil {
			t.Fatal("the fixture's .git carries a worktrees directory; the shape under test is a repository without one")
		}
		// A fresh high record in a directory no git metadata registers.
		qasWriteRecord(t, filepath.Join(root, ".moai", "worktrees", "stray"), "sess-stray", fcNow, qasWin(92, qasReset5), nil)
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("next: %v (stderr %q)", err, stderr)
		}
		qasAssertNotHeld(t, out, stderr)
		if cd := fcCard(t, root, "t2"); cd.State != homestate.CardLeased || cd.LeaseHolder != "lane-1" {
			t.Fatalf("t2 = %s holder=%q, want leased/lane-1", cd.State, cd.LeaseHolder)
		}
	})
	t.Run("roots_own_record_is_read", func(t *testing.T) {
		// The control: the same fixture with the record in the root's own
		// directory holds, so the lease above is attributable to the record's
		// location and not to a fixture that can never hold.
		qasFixture(t, qasFixtureOpts{five: qasWin(92, qasReset5)})
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		qasAssertHeld(t, out, stderr, err)
	})
}

// AC-QWR-013b — the production seam applies the configured directory bound:
// with max_scan_dirs 3 only the first three of five registered worktrees are
// read, with the key absent all five, and the seam and the enumerator are each
// called once per evaluation.
func TestQWR_AC013b_ProductionSeamAppliesConfiguredBound(t *testing.T) {
	sdClearLaneEnv(t)
	root, _ := sdMoaiFixture(t)
	// Five worktrees wt-0..wt-4 (the entry names sort in index order), each with
	// one fresh record whose percentage and capture time grow with the index, so
	// the freshest record among the examined entries is the highest-index one.
	for i := range 5 {
		name := "wt-" + strconv.Itoa(i)
		wt := qwrAddWorktree(t, root, name)
		qasWriteRecord(t, wt, "sess-"+name, fcNow.Add(-25*time.Minute+time.Duration(i)*time.Minute), qasWin(float64(60+5*i), qasReset5), nil)
	}
	calls := qwrCountCalls(t)
	fiveHour := func(t *testing.T) statusline.QuotaReading {
		t.Helper()
		*calls = qwrCalls{}
		ev := factoryQuotaEvaluate(root)
		if !ev.Enabled || len(ev.Windows) == 0 {
			t.Fatalf("the evaluation is not enabled: %+v", ev)
		}
		if calls.aggregate != 1 || calls.enumerate != 1 {
			t.Errorf("seam calls per evaluation = aggregator %d, enumerator %d, want 1 and 1", calls.aggregate, calls.enumerate)
		}
		return ev.Windows[0].Reading
	}
	expect := func(t *testing.T, got statusline.QuotaReading, want float64) {
		t.Helper()
		if got.State != statusline.QuotaFresh || got.UsedPercentage != want {
			t.Errorf("five_hour reading = %s %v, want fresh %v", got.State, got.UsedPercentage, want)
		}
	}

	t.Run("bound_3_reads_three", func(t *testing.T) {
		qwrWriteGate(t, root, true, "        max_scan_dirs: 3\n")
		expect(t, fiveHour(t), 70) // wt-2; wt-3 and wt-4 are beyond the bound
	})
	t.Run("key_absent_reads_all_five", func(t *testing.T) {
		qwrWriteGate(t, root, true, "")
		expect(t, fiveHour(t), 80) // wt-4
	})
}
