package hook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Acceptance tests for the sync-phase quality gate's failure-state record
// (SPEC-SYNC-GATE-FAILSTATE-001, acceptance.md). Every test drives the LOCAL
// hook copy (.claude/hooks/moai/sync-phase-quality-gate.sh) inside a
// t.TempDir() git fixture, with a stub toolchain first on PATH, HOME and TMPDIR
// redirected into the temp tree, and an explicit environment (no inherited
// MOAI_SYNC_GATE_BLOCKING / MOAI_AUTONOMY_TIER unless a row sets one).
//
// Row classes follow acceptance.md: a release-blocking row carries a named RED
// reason; a regression-guard row must already pass on the pre-fix hook. Every
// assertion message is tagged with its row id and class so a -v run can be
// compared row by row against the named RED reasons.
//
// The process-group harness for interrupted runs (AC-006c, AC-015) lives in
// sync_gate_failstate_unix_test.go.

const (
	sgfBaseSubject = "feat(fx): base commit"
	sgfSyncSubject = "docs(fx): sync-phase artifacts"
	sgfRecordName  = "sync-quality-gate.last"
	sgfPayloadName = "sync-quality-gate.payload"
	sgfBlockingEnv = "MOAI_SYNC_GATE_BLOCKING"
	sgfStaleAge    = 120 * time.Second
	sgfRunGuard    = 90 * time.Second
	rbClass        = "release-blocking"
	rgClass        = "regression-guard"
)

// sgfOpts selects the fixture shape.
type sgfOpts struct {
	python      bool   // pyproject.toml + .py change, ruff stub (otherwise go.mod + .go change, go stub)
	initialOnly bool   // HEAD is the repository's first commit (no HEAD~1)
	subject     string // HEAD subject; empty means sgfSyncSubject
	vetExit     int    // go stub exit for `go vet`
	buildExit   int    // go stub exit for `go build`
	ruffExit    int    // ruff stub exit
}

// sgfStubSpec describes the stub toolchain binary. The stub always appends one
// line to the counter file first, so readCounter reports invocations.
type sgfStubSpec struct {
	exits       map[string]int // first argument -> exit code
	defaultExit int
	writeMarker bool // create the marker file before sleeping or exiting
	sleep       bool // sleep in a self-bounded loop (at most 30 x 1s)
	// probePayload (AC-013 S2 only): on every invocation, append "present" or
	// "absent" for the payload file to the fixture probe file. Unset, the stub
	// script bytes are unchanged for every other row.
	probePayload bool
}

type sgfFixture struct {
	t        *testing.T
	repo     string
	stubBin  string
	stubName string
	home     string
	gtmp     string
	counter  string
	marker   string
	probe    string
	script   string
}

func sgfRepoRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(mustGetwd(t), "..", "..")
}

func newSGFFixture(t *testing.T, o sgfOpts) *sgfFixture {
	t.Helper()
	requireBash(t)
	requireGit(t)
	tmp := t.TempDir()
	f := &sgfFixture{
		t:       t,
		repo:    filepath.Join(tmp, "repo"),
		stubBin: filepath.Join(tmp, "bin"),
		home:    filepath.Join(tmp, "home"),
		gtmp:    filepath.Join(tmp, "tmp"),
		counter: filepath.Join(tmp, "stub.count"),
		marker:  filepath.Join(tmp, "stub.marker"),
		probe:   filepath.Join(tmp, "stub.probe"),
		script:  filepath.Join(sgfRepoRoot(t), ".claude", "hooks", "moai", "sync-phase-quality-gate.sh"),
	}
	for _, d := range []string{f.repo, f.stubBin, f.home, f.gtmp} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("fixture mkdir %s: %v", d, err)
		}
	}
	if _, err := os.Stat(f.script); err != nil {
		t.Fatalf("local hook copy not found: %s (%v)", f.script, err)
	}
	subject := o.subject
	if subject == "" {
		subject = sgfSyncSubject
	}
	if o.python {
		f.stubName = "ruff"
		mustRunGit(t, f.repo, "init")
		mustRunGit(t, f.repo, "config", "user.email", "sync-gate-test@example.com")
		mustRunGit(t, f.repo, "config", "user.name", "Sync Gate Test")
		mustRunGit(t, f.repo, "config", "core.hooksPath", "/dev/null")
		f.writeRepoFile("pyproject.toml", "[project]\nname = \"fx\"\nversion = \"0.1.0\"\n")
		f.writeRepoFile("app.py", "print('base')\n")
		mustRunGit(t, f.repo, "add", "pyproject.toml", "app.py")
		if o.initialOnly {
			mustRunGit(t, f.repo, "commit", "-m", subject)
		} else {
			mustRunGit(t, f.repo, "commit", "-m", sgfBaseSubject)
			f.writeRepoFile("app.py", "print('sync change')\n")
			mustRunGit(t, f.repo, "add", "app.py")
			mustRunGit(t, f.repo, "commit", "-m", subject)
		}
		f.setStub(sgfStubSpec{defaultExit: o.ruffExit})
		return f
	}
	f.stubName = "go"
	if o.initialOnly {
		initGoFixtureRepo(t, f.repo, subject)
	} else {
		initGoFixtureRepo(t, f.repo, sgfBaseSubject)
		f.writeRepoFile("main.go", "package main\n\nfunc main() { _ = 1 }\n")
		mustRunGit(t, f.repo, "add", "main.go")
		mustRunGit(t, f.repo, "commit", "-m", subject)
	}
	f.setStub(sgfGoStub(o.vetExit, o.buildExit))
	return f
}

func sgfGoStub(vetExit, buildExit int) sgfStubSpec {
	return sgfStubSpec{exits: map[string]int{"vet": vetExit, "build": buildExit}}
}

func (f *sgfFixture) writeRepoFile(rel, body string) {
	f.t.Helper()
	if err := os.WriteFile(filepath.Join(f.repo, rel), []byte(body), 0o644); err != nil {
		f.t.Fatalf("write %s: %v", rel, err)
	}
}

// setStub (re)writes the stub toolchain binary.
func (f *sgfFixture) setStub(s sgfStubSpec) {
	f.t.Helper()
	var b strings.Builder
	b.WriteString("#!/bin/bash\n")
	fmt.Fprintf(&b, "echo 1 >> '%s'\n", f.counter)
	if s.probePayload {
		fmt.Fprintf(&b, "if [ -e '%s' ]; then echo present >> '%s'; else echo absent >> '%s'; fi\n", f.payloadPath(), f.probe, f.probe)
	}
	if s.writeMarker {
		fmt.Fprintf(&b, ": > '%s'\n", f.marker)
	}
	if s.sleep {
		b.WriteString("i=0\nwhile [ \"$i\" -lt 30 ]; do sleep 1; i=$((i+1)); done\n")
	}
	keys := make([]string, 0, len(s.exits))
	for k := range s.exits {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString("case \"$1\" in\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "  %s) exit %d ;;\n", k, s.exits[k])
	}
	b.WriteString("esac\n")
	fmt.Fprintf(&b, "exit %d\n", s.defaultExit)
	if err := os.WriteFile(filepath.Join(f.stubBin, f.stubName), []byte(b.String()), 0o755); err != nil {
		f.t.Fatalf("write stub: %v", err)
	}
}

// commitSync makes a new sync-phase commit with a code change and returns HEAD.
func (f *sgfFixture) commitSync(n int) string {
	f.t.Helper()
	f.writeRepoFile("main.go", fmt.Sprintf("package main\n\nfunc main() { _ = %d }\n", n+100))
	mustRunGit(f.t, f.repo, "add", "main.go")
	mustRunGit(f.t, f.repo, "commit", "-m", sgfSyncSubject)
	return f.rev("HEAD")
}

func (f *sgfFixture) rev(ref string) string {
	f.t.Helper()
	out, err := exec.Command("git", "-C", f.repo, "rev-parse", ref).Output()
	if err != nil {
		f.t.Fatalf("git rev-parse %s: %v", ref, err)
	}
	return strings.TrimSpace(string(out))
}

func (f *sgfFixture) count() int { return readCounter(f.t, f.counter) }

func (f *sgfFixture) stateDir() string { return filepath.Join(f.repo, ".moai", "state") }

func (f *sgfFixture) recordPath() string { return filepath.Join(f.stateDir(), sgfRecordName) }

func (f *sgfFixture) payloadPath() string { return filepath.Join(f.stateDir(), sgfPayloadName) }

func (f *sgfFixture) writeRecord(content string) {
	f.t.Helper()
	if err := os.MkdirAll(f.stateDir(), 0o755); err != nil {
		f.t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(f.recordPath(), []byte(content), 0o644); err != nil {
		f.t.Fatalf("write record: %v", err)
	}
}

// readRecord returns the record content with trailing newlines removed.
func (f *sgfFixture) readRecord(tag string) (string, bool) {
	f.t.Helper()
	b, err := os.ReadFile(f.recordPath())
	if err != nil {
		f.t.Errorf("%s: read record: %v", tag, err)
		return "", false
	}
	return strings.TrimRight(string(b), "\n"), true
}

// ageRecord sets the record file's mtime to now minus age (age 0 = now).
func (f *sgfFixture) ageRecord(tag string, age time.Duration) {
	f.t.Helper()
	ts := time.Now().Add(-age)
	if err := os.Chtimes(f.recordPath(), ts, ts); err != nil {
		f.t.Errorf("%s: age record by %s: %v", tag, age, err)
	}
}

func (f *sgfFixture) env(extra ...string) []string {
	e := []string{
		"PATH=" + f.stubBin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + f.home,
		"CLAUDE_PROJECT_DIR=" + f.repo,
		"TMPDIR=" + f.gtmp,
	}
	return append(e, extra...)
}

// invoke runs the hook to completion. A nil stdin means /dev/null.
func (f *sgfFixture) invoke(stdin *string, extra ...string) (string, int) {
	f.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), sgfRunGuard)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", f.script)
	cmd.Dir = f.repo
	cmd.Env = f.env(extra...)
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = 5 * time.Second
	err := cmd.Run()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			f.t.Fatalf("run hook: %v; stderr=%s", err, stderr.String())
		}
		code = ee.ExitCode()
	}
	if ctx.Err() != nil {
		f.t.Errorf("hook exceeded the %s run guard; stderr=%s", sgfRunGuard, stderr.String())
	}
	return stdout.String(), code
}

func (f *sgfFixture) run(stdin string, extra ...string) (string, int) {
	f.t.Helper()
	return f.invoke(&stdin, extra...)
}

func tierEnv(tier string) string { return config.EnvAutonomyTier + "=" + tier }

func sgfHasBlock(out string) bool {
	return strings.Contains(out, `"hookSpecificOutput"`) && strings.Contains(out, `"decision":"block"`)
}

func sgfHasDecision(out string) bool { return strings.Contains(out, `"decision"`) }

// sgfExpectRecord checks a record by field rather than by exact bytes: it must
// name this HEAD and this outcome. The third field is the work-tree content
// identifier (card t601), whose value the hook computes, so a test that spelled
// it out would be asserting its own arithmetic rather than the hook's.
func sgfExpectRecord(t *testing.T, tag, rec, head, outcome string) {
	t.Helper()
	fields := strings.Fields(rec)
	if len(fields) < 2 || fields[0] != head || fields[1] != outcome {
		t.Errorf("%s: record = %q; want a record naming HEAD %s with outcome %q", tag, rec, head, outcome)
	}
}

func sgfHasSystemMessage(out string) bool { return strings.Contains(out, `"systemMessage"`) }

// sgfNamesExhaustedRetry reports whether a notice says this HEAD's gate run has
// not completed and names deleting the state file as the way to force a re-gate.
func sgfNamesExhaustedRetry(out string) bool {
	low := strings.ToLower(out)
	return strings.Contains(low, "not completed") && strings.Contains(low, sgfRecordName) && strings.Contains(low, "delet")
}

// sgfExpectNotice asserts the non-blocking notice outcome of one call.
func sgfExpectNotice(t *testing.T, tag string, delta int, out string) {
	t.Helper()
	if delta != 0 {
		t.Errorf("%s: stub invoked %d time(s) on this call; want 0 (checks must not run)", tag, delta)
	}
	if !sgfHasSystemMessage(out) {
		t.Errorf("%s: stdout lacks \"systemMessage\"; stdout=%q", tag, out)
	}
	if sgfHasDecision(out) {
		t.Errorf("%s: stdout carries \"decision\"; want none; stdout=%q", tag, out)
	}
}

// sgfExpectRegate asserts the checks ran on this call and the call blocked.
func sgfExpectRegate(t *testing.T, tag string, delta int, out string) {
	t.Helper()
	if delta < 1 {
		t.Errorf("%s: stub invoked %d time(s) on this call; want >= 1 (checks must run)", tag, delta)
	}
	if !sgfHasBlock(out) {
		t.Errorf("%s: stdout is not a block; stdout=%q", tag, out)
	}
}

// TestSyncGateFailState_AC001_FailureRedeliveredOnSameHead — AC-001 (release-blocking).
func TestSyncGateFailState_AC001_FailureRedeliveredOnSameHead(t *testing.T) {
	tag := "AC-001 [" + rbClass + "]"
	f := newSGFFixture(t, sgfOpts{vetExit: 1})
	out1, code1 := f.run("{}")
	n1 := f.count()
	out2, code2 := f.run("{}")
	n2 := f.count()
	t.Logf("%s call1 stdout (%d bytes)=%q stub=%d; call2 stdout (%d bytes)=%q stub=%d", tag, len(out1), out1, n1, len(out2), out2, n2)
	if !sgfHasBlock(out1) {
		t.Errorf("%s: call 1 is not a block; stdout=%q", tag, out1)
	}
	if out2 != out1 {
		t.Errorf("%s: call 2 stdout is not byte-identical to call 1 (call1 %d bytes, call2 %d bytes); call2=%q", tag, len(out1), len(out2), out2)
	}
	if n2 != n1 {
		t.Errorf("%s: stub count after call 2 = %d, after call 1 = %d; want equal (no re-run)", tag, n2, n1)
	}
	if code1 != 0 || code2 != 0 {
		t.Errorf("%s: exit codes call1=%d call2=%d; want 0 and 0", tag, code1, code2)
	}
}

// TestSyncGateFailState_AC002_NewFailingHeadBlocksWithNewResult — AC-002 (regression-guard).
func TestSyncGateFailState_AC002_NewFailingHeadBlocksWithNewResult(t *testing.T) {
	tag := "AC-002 [" + rgClass + "]"
	f := newSGFFixture(t, sgfOpts{vetExit: 1})
	_, _ = f.run("{}")
	newHead := f.commitSync(1)
	before := f.count()
	out, code := f.run("{}")
	sgfExpectRegate(t, tag, f.count()-before, out)
	if code != 0 {
		t.Errorf("%s: exit %d; want 0", tag, code)
	}
	if rec, ok := f.readRecord(tag); ok {
		fields := strings.Fields(rec)
		if len(fields) == 0 || fields[0] != newHead {
			t.Errorf("%s: record %q does not name the new HEAD %s", tag, rec, newHead)
		}
	}
}

// TestSyncGateFailState_AC003_PassThenSameHeadStaysSilent — AC-003 (regression-guard).
func TestSyncGateFailState_AC003_PassThenSameHeadStaysSilent(t *testing.T) {
	tag := "AC-003 [" + rgClass + "]"
	f := newSGFFixture(t, sgfOpts{})
	out1, _ := f.run("{}")
	n1 := f.count()
	out2, _ := f.run("{}")
	n2 := f.count()
	if out1 != "" || out2 != "" {
		t.Errorf("%s: stdouts must be empty; call1=%q call2=%q", tag, out1, out2)
	}
	if n1 < 1 {
		t.Errorf("%s: call 1 did not run the checks (stub=%d)", tag, n1)
	}
	if n2 != n1 {
		t.Errorf("%s: stub count after call 2 = %d, after call 1 = %d; want equal", tag, n2, n1)
	}
}

// TestSyncGateFailState_AC004_StopHookActiveDefersRedelivery — AC-004 (release-blocking, a/b/c).
func TestSyncGateFailState_AC004_StopHookActiveDefersRedelivery(t *testing.T) {
	forms := []struct{ id, stdin string }{
		{"F1", `{"stop_hook_active": true}`},
		{"F2", `{"stop_hook_active":true}`},
		{"F3", `{"stop_hook_active"   :   true}`},
		{"F4", "{\"stop_hook_active\":\ttrue}"},
	}
	for _, fm := range forms {
		t.Run(fm.id, func(t *testing.T) {
			f := newSGFFixture(t, sgfOpts{vetExit: 1})
			out1, _ := f.run("{}")
			n1 := f.count()
			outA, _ := f.run(fm.stdin)
			tagA := "AC-004 " + fm.id + " (a) [" + rbClass + "]"
			if sgfHasDecision(outA) {
				t.Errorf("%s: flagged call carries \"decision\"; stdout=%q", tagA, outA)
			}
			if n := f.count(); n != n1 {
				t.Errorf("%s: stub count changed on flagged call: %d -> %d", tagA, n1, n)
			}
			outB, _ := f.run("{}")
			tagB := "AC-004 " + fm.id + " (b) [" + rbClass + "]"
			if outB != out1 {
				t.Errorf("%s: unflagged call after the flag is not byte-identical to call 1 (call1 %d bytes, now %d bytes); stdout=%q", tagB, len(out1), len(outB), outB)
			}
		})
	}
	t.Run("c", func(t *testing.T) {
		tag := "AC-004 (c) [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		out1, _ := f.run("{}")
		outC, _ := f.run(`{"stop_hook_active": false, "last_assistant_message": "note: \"stop_hook_active\": true"}`)
		if outC != out1 {
			t.Errorf("%s: stdout is not byte-identical to call 1 (call1 %d bytes, now %d bytes); stdout=%q", tag, len(out1), len(outC), outC)
		}
	})
}

// TestSyncGateFailState_AC005_UnknownAndLegacyRecordsRegate — AC-005 L1, U1-U5.
func TestSyncGateFailState_AC005_UnknownAndLegacyRecordsRegate(t *testing.T) {
	rows := []struct {
		id       string
		runClass string
		prep     func(f *sgfFixture, head string)
		unread   bool
	}{
		{"L1", rbClass, func(f *sgfFixture, head string) { f.writeRecord(head + "\n") }, false},
		{"U1", rgClass, func(f *sgfFixture, head string) { f.writeRecord(head + " bogus\n") }, false},
		{"U2", rgClass, func(f *sgfFixture, head string) { f.writeRecord(head + " fail extra\n") }, false},
		{"U3", rgClass, func(f *sgfFixture, _ string) { f.writeRecord("") }, false},
		{"U4", rgClass, func(f *sgfFixture, head string) {
			f.writeRecord(head + " fail\n")
			if err := os.Chmod(f.recordPath(), 0o200); err != nil {
				f.t.Fatalf("chmod 0200: %v", err)
			}
		}, true},
		{"U5", rgClass, func(f *sgfFixture, head string) { f.writeRecord(head + " fail\n") }, false},
	}
	for _, r := range rows {
		t.Run(r.id, func(t *testing.T) {
			if r.unread && (runtime.GOOS == "windows" || os.Geteuid() == 0) {
				t.Skip("U4 needs a non-root Unix user (mode 0200 must deny read)")
			}
			f := newSGFFixture(t, sgfOpts{vetExit: 1})
			head := f.rev("HEAD")
			r.prep(f, head)
			before := f.count()
			out, code := f.run("{}")
			tag := "AC-005 " + r.id + " run-and-block [" + r.runClass + "]"
			sgfExpectRegate(t, tag, f.count()-before, out)
			if code != 0 {
				t.Errorf("%s: exit %d; want 0", tag, code)
			}
			if r.unread {
				if err := os.Chmod(f.recordPath(), 0o600); err != nil {
					t.Errorf("AC-005 U4 read step: chmod 0600: %v", err)
				}
			}
			recTag := "AC-005 " + r.id + " record-format [" + rbClass + "]"
			if rec, ok := f.readRecord(recTag); ok {
				sgfExpectRecord(t, recTag, rec, head, "fail")
			}
		})
	}
}

// TestSyncGateFailState_AC005_TornWriteNeverSilentPass — AC-005 torn-write rows TA1-TA5, TB1-TB2.
func TestSyncGateFailState_AC005_TornWriteNeverSilentPass(t *testing.T) {
	type tornRow struct {
		id     string
		class  string
		stateB bool // torn state (b): setup blocks, auxiliary files removed
		prep   func(f *sgfFixture, head string)
		notice bool
	}
	running := func(age time.Duration) func(f *sgfFixture, head string) {
		return func(f *sgfFixture, head string) {
			f.writeRecord(head + " running\n")
			f.ageRecord("torn prep", age)
		}
	}
	rows := []tornRow{
		{"TA1", rbClass, false, running(0), true},
		{"TA2", rgClass, false, running(sgfStaleAge), false},
		{"TA3", rgClass, false, func(f *sgfFixture, _ string) {
			if err := os.Remove(f.recordPath()); err != nil {
				f.t.Fatalf("remove record: %v", err)
			}
		}, false},
		{"TA4", rbClass, false, func(f *sgfFixture, head string) { f.writeRecord(head + "\n") }, false},
		{"TA5-bare", rgClass, false, func(f *sgfFixture, _ string) { f.writeRecord(f.rev("HEAD~1") + "\n") }, false},
		{"TA5-pass", rgClass, false, func(f *sgfFixture, _ string) { f.writeRecord(f.rev("HEAD~1") + " pass\n") }, false},
		{"TB1", rbClass, true, running(0), true},
		{"TB2", rgClass, true, running(sgfStaleAge), false},
	}
	for _, r := range rows {
		t.Run(r.id, func(t *testing.T) {
			f := newSGFFixture(t, sgfOpts{vetExit: 1})
			head := f.rev("HEAD")
			if r.stateB {
				_, _ = f.run("{}")
				entries, err := os.ReadDir(f.stateDir())
				if err != nil {
					t.Fatalf("read state dir: %v", err)
				}
				for _, e := range entries {
					if e.Name() != sgfRecordName {
						if err := os.RemoveAll(filepath.Join(f.stateDir(), e.Name())); err != nil {
							t.Fatalf("remove %s: %v", e.Name(), err)
						}
					}
				}
			} else {
				_, _ = f.run("{}", tierEnv(config.AutonomyTierFullyAutonomous))
			}
			r.prep(f, head)
			before := f.count()
			out, code := f.run("{}")
			tag := "AC-005 " + r.id + " [" + r.class + "]"
			if code != 0 {
				t.Errorf("%s: exit %d; want 0", tag, code)
			}
			if r.notice {
				sgfExpectNotice(t, tag, f.count()-before, out)
			} else {
				sgfExpectRegate(t, tag, f.count()-before, out)
			}
		})
	}
}

// TestSyncGateFailState_AC006_RunningRecordStaleWindow — AC-006 rows a0, a50, b61, b70, b120.
func TestSyncGateFailState_AC006_RunningRecordStaleWindow(t *testing.T) {
	rows := []struct {
		id     string
		class  string
		age    time.Duration
		notice bool
	}{
		{"a0", rbClass, 0, true},
		{"a50", rbClass, 50 * time.Second, true},
		{"b61", rgClass, 61 * time.Second, false},
		{"b70", rgClass, 70 * time.Second, false},
		{"b120", rgClass, 120 * time.Second, false},
	}
	for _, r := range rows {
		t.Run(r.id, func(t *testing.T) {
			f := newSGFFixture(t, sgfOpts{vetExit: 1})
			f.writeRecord(f.rev("HEAD") + " running\n")
			tag := "AC-006 " + r.id + " [" + r.class + "]"
			f.ageRecord(tag, r.age)
			before := f.count()
			out, code := f.run("{}")
			if code != 0 {
				t.Errorf("%s: exit %d; want 0", tag, code)
			}
			if r.notice {
				sgfExpectNotice(t, tag, f.count()-before, out)
			} else {
				sgfExpectRegate(t, tag, f.count()-before, out)
			}
		})
	}
}

// TestSyncGateFailState_AC007_NoPathLooserThanToday — AC-007 R1-R17 (regression-guard).
func TestSyncGateFailState_AC007_NoPathLooserThanToday(t *testing.T) {
	strp := func(s string) *string { return &s }
	type row struct {
		id      string
		opts    sgfOpts
		env     []string
		stdin   *string // nil = /dev/null
		history func(f *sgfFixture)
		prep    func(f *sgfFixture)
	}
	obj := strp("{}")
	rows := []row{
		{id: "R1", opts: sgfOpts{vetExit: 1}, stdin: obj},
		{id: "R2", opts: sgfOpts{buildExit: 1}, stdin: obj},
		{id: "R3", opts: sgfOpts{vetExit: 1}, env: []string{sgfBlockingEnv + "=1"}, stdin: obj},
		{id: "R4", opts: sgfOpts{vetExit: 1}, env: []string{tierEnv(config.AutonomyTierSemiAuto)}, stdin: obj},
		{id: "R5", opts: sgfOpts{buildExit: 1}, env: []string{tierEnv(config.AutonomyTierAutomatic)}, stdin: obj},
		{id: "R6", opts: sgfOpts{vetExit: 1}, env: []string{tierEnv("bogus")}, stdin: obj},
		{id: "R7", opts: sgfOpts{}, stdin: obj, history: func(f *sgfFixture) {
			_, _ = f.run("{}")
			f.setStub(sgfGoStub(1, 0))
			f.commitSync(7)
		}},
		{id: "R8", opts: sgfOpts{vetExit: 1}, stdin: obj, history: func(f *sgfFixture) {
			out, _ := f.run("{}")
			if !sgfHasBlock(out) {
				f.t.Errorf("AC-007 R8 history: first failing HEAD did not block; stdout=%q", out)
			}
			f.commitSync(8)
		}},
		{id: "R9", opts: sgfOpts{vetExit: 1}, stdin: obj, prep: func(f *sgfFixture) {
			if err := os.MkdirAll(filepath.Join(f.repo, ".moai"), 0o755); err != nil {
				f.t.Fatalf("mkdir .moai: %v", err)
			}
		}},
		{id: "R10a", opts: sgfOpts{vetExit: 1}, stdin: strp(`{"stop_hook_active": true}`)},
		{id: "R10b", opts: sgfOpts{vetExit: 1}, stdin: strp(`{"stop_hook_active":true}`)},
		{id: "R11", opts: sgfOpts{vetExit: 1}, stdin: nil},
		{id: "R12", opts: sgfOpts{vetExit: 1, initialOnly: true}, stdin: obj},
		{id: "R13", opts: sgfOpts{python: true, ruffExit: 1}, stdin: obj},
		{id: "R14", opts: sgfOpts{vetExit: 1, subject: "chore: sync docs"}, stdin: obj},
		{id: "R15", opts: sgfOpts{vetExit: 1, buildExit: 1}, env: []string{tierEnv(config.AutonomyTierAutomatic)}, stdin: obj},
		{id: "R16", opts: sgfOpts{vetExit: 1}, env: []string{sgfBlockingEnv + "="}, stdin: obj},
		{id: "R17", opts: sgfOpts{vetExit: 1}, stdin: obj, history: func(f *sgfFixture) {
			f.writeRecord(f.rev("HEAD") + "\n")
			f.commitSync(17)
		}},
	}
	for _, r := range rows {
		t.Run(r.id, func(t *testing.T) {
			tag := "AC-007 " + r.id + " [" + rgClass + "]"
			f := newSGFFixture(t, r.opts)
			if r.history != nil {
				r.history(f)
			}
			if r.prep != nil {
				r.prep(f)
			}
			if r.id == "R12" {
				if _, err := exec.Command("git", "-C", f.repo, "rev-parse", "--verify", "-q", "HEAD~1").Output(); err == nil {
					t.Fatalf("%s: fixture has HEAD~1; want an initial commit", tag)
				}
			}
			before := f.count()
			out, code := f.invoke(r.stdin, r.env...)
			if code != 0 {
				t.Errorf("%s: exit %d; want 0", tag, code)
			}
			if !sgfHasBlock(out) {
				t.Errorf("%s: final invocation is not a block; stdout=%q", tag, out)
			}
			if r.history != nil {
				if delta := f.count() - before; delta < 1 {
					t.Errorf("%s: stub count did not increase on the final invocation (delta %d)", tag, delta)
				}
			}
		})
	}
}

// TestSyncGateFailState_AC008_RedeliveryFollowsModeResolution — AC-008 A1-A9.
func TestSyncGateFailState_AC008_RedeliveryFollowsModeResolution(t *testing.T) {
	full := tierEnv(config.AutonomyTierFullyAutonomous)
	auto := tierEnv(config.AutonomyTierAutomatic)
	off := sgfBlockingEnv + "=0"
	rows := []struct {
		id              string
		class           string
		env1, env2      []string
		vet, build      int
		redeliver       bool // true: call 2 byte-identical to call 1's block
		call1NoDecision bool
	}{
		{"A1", rgClass, []string{full}, []string{full}, 0, 1, false, true},
		{"A2", rgClass, []string{auto}, []string{auto}, 1, 0, false, true},
		{"A3", rgClass, []string{off}, []string{off}, 1, 0, false, true},
		{"A4", rbClass, []string{auto}, []string{auto}, 0, 1, true, false},
		{"A5", rgClass, nil, []string{full}, 1, 0, false, false},
		{"A6", rgClass, []string{full}, nil, 1, 0, false, false},
		{"A7", rgClass, nil, []string{off}, 1, 0, false, false},
		{"A8", rgClass, nil, []string{auto}, 1, 0, false, false},
		{"A9", rbClass, []string{auto}, []string{auto}, 1, 1, true, false},
	}
	for _, r := range rows {
		t.Run(r.id, func(t *testing.T) {
			tag := "AC-008 " + r.id + " [" + r.class + "]"
			f := newSGFFixture(t, sgfOpts{vetExit: r.vet, buildExit: r.build})
			out1, _ := f.run("{}", r.env1...)
			n1 := f.count()
			out2, _ := f.run("{}", r.env2...)
			n2 := f.count()
			t.Logf("%s call1=%q call2=%q stub %d -> %d", tag, out1, out2, n1, n2)
			if r.call1NoDecision && sgfHasDecision(out1) {
				t.Errorf("%s: call 1 carries \"decision\"; stdout=%q", tag, out1)
			}
			if r.redeliver {
				if !sgfHasBlock(out1) {
					t.Errorf("%s: call 1 is not a block; stdout=%q", tag, out1)
				}
				if out2 != out1 {
					t.Errorf("%s: call 2 stdout is not byte-identical to call 1's block (call1 %d bytes, call2 %d bytes); call2=%q", tag, len(out1), len(out2), out2)
				}
				return
			}
			if out2 != "" {
				t.Errorf("%s: call 2 stdout = %q; want empty", tag, out2)
			}
			if n2 != n1 {
				t.Errorf("%s: call 2 re-ran the checks (stub %d -> %d)", tag, n1, n2)
			}
		})
	}
}

// TestSyncGateFailState_AC013_RetryByDeletionNoStaleAuxState — AC-013 D1, S1 (regression-guard).
func TestSyncGateFailState_AC013_RetryByDeletionNoStaleAuxState(t *testing.T) {
	t.Run("D1", func(t *testing.T) {
		tag := "AC-013 D1 [" + rgClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		_, _ = f.run("{}")
		if err := os.Remove(f.recordPath()); err != nil {
			t.Fatalf("%s: remove record: %v", tag, err)
		}
		before := f.count()
		out, _ := f.run("{}")
		sgfExpectRegate(t, tag, f.count()-before, out)
	})
	t.Run("S1", func(t *testing.T) {
		tag := "AC-013 S1 [" + rgClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		out1, _ := f.run("{}")
		if !sgfHasBlock(out1) {
			t.Errorf("%s: call 1 is not a block; stdout=%q", tag, out1)
		}
		if err := os.Remove(f.recordPath()); err != nil {
			t.Fatalf("%s: remove record: %v", tag, err)
		}
		before2 := f.count()
		out2, _ := f.run("{}", tierEnv(config.AutonomyTierFullyAutonomous))
		if f.count() <= before2 {
			t.Errorf("%s: call 2 did not run the checks", tag)
		}
		if sgfHasDecision(out2) {
			t.Errorf("%s: call 2 carries \"decision\"; stdout=%q", tag, out2)
		}
		before3 := f.count()
		out3, _ := f.run("{}")
		if out3 != "" {
			t.Errorf("%s: call 3 stdout = %q; want empty (call-1 block must not be re-delivered)", tag, out3)
		}
		if n := f.count(); n != before3 {
			t.Errorf("%s: call 3 ran the checks (stub %d -> %d)", tag, before3, n)
		}
	})
	// S2: while an invocation runs the checks, the stored payload is already gone.
	t.Run("S2", func(t *testing.T) {
		tag := "AC-013 S2 [" + rgClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		f.setStub(sgfStubSpec{exits: map[string]int{"vet": 1, "build": 0}, probePayload: true})
		out1, _ := f.run("{}")
		if !sgfHasBlock(out1) {
			t.Errorf("%s: call 1 is not a block; stdout=%q", tag, out1)
		}
		if err := os.Remove(f.recordPath()); err != nil {
			t.Fatalf("%s: remove record: %v", tag, err)
		}
		if _, err := os.Stat(f.payloadPath()); err != nil {
			t.Fatalf("%s setup: payload file does not exist before call 2 (%v); the row cannot be interpreted", tag, err)
		}
		if err := os.WriteFile(f.probe, nil, 0o644); err != nil {
			t.Fatalf("%s: empty probe file: %v", tag, err)
		}
		before2 := f.count()
		out2, _ := f.run("{}")
		delta2 := f.count() - before2
		raw, err := os.ReadFile(f.probe)
		if err != nil {
			t.Fatalf("%s: read probe file: %v", tag, err)
		}
		obs := strings.Fields(string(raw))
		t.Logf("%s call 2 stub delta=%d observations=%v stdout=%q", tag, delta2, obs, out2)
		if len(obs) == 0 {
			t.Errorf("%s: the probe holds no observation from call 2 (stub delta %d); the checks never ran — a gap, not a pass", tag, delta2)
		}
		present := 0
		for _, o := range obs {
			switch o {
			case "absent":
			case "present":
				present++
			default:
				t.Errorf("%s: unexpected probe observation %q", tag, o)
			}
		}
		if present != 0 {
			t.Errorf("%s: %d of %d stub invocation(s) during call 2 saw the payload file present; want every observation absent", tag, present, len(obs))
		}
	})
	// S3: after a partial write (payload move refused, fail record written), the
	// stale call-1 payload is not re-delivered; call 3 re-runs the checks.
	t.Run("S3", func(t *testing.T) {
		tag := "AC-013 S3 [" + rgClass + "]"
		if runtime.GOOS == "windows" {
			t.Skip("S3 installs a POSIX mv shim first on PATH")
		}
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		head := f.rev("HEAD")
		out1, _ := f.run("{}")
		if !sgfHasBlock(out1) {
			t.Errorf("%s: call 1 is not a block; stdout=%q", tag, out1)
		}
		if err := os.Remove(f.recordPath()); err != nil {
			t.Fatalf("%s: remove record: %v", tag, err)
		}
		realMv, err := exec.LookPath("mv")
		if err != nil {
			t.Fatalf("%s: resolve system mv before installing the shim: %v", tag, err)
		}
		if realMv, err = filepath.Abs(realMv); err != nil {
			t.Fatalf("%s: absolute system mv path: %v", tag, err)
		}
		shimLog := filepath.Join(filepath.Dir(f.counter), "mv-shim.log")
		shim := filepath.Join(f.stubBin, "mv")
		shimBody := fmt.Sprintf("#!/bin/bash\nlast=\nfor a in \"$@\"; do last=$a; done\nif [ \"$last\" = '%s' ]; then echo \"refused $last\" >> '%s'; exit 1; fi\nexec '%s' \"$@\"\n",
			f.payloadPath(), shimLog, realMv)
		if err := os.WriteFile(shim, []byte(shimBody), 0o755); err != nil {
			t.Fatalf("%s: install mv shim: %v", tag, err)
		}
		before2 := f.count()
		out2, _ := f.run("{}")
		delta2 := f.count() - before2
		if err := os.Remove(shim); err != nil {
			t.Fatalf("%s: remove mv shim: %v", tag, err)
		}
		logRaw, _ := os.ReadFile(shimLog)
		refused := strings.Count(string(logRaw), "refused "+f.payloadPath())
		t.Logf("%s call 2 stub delta=%d refused-moves=%d stdout=%q", tag, delta2, refused, out2)
		// Reachability, asserted separately: a failure here is a gap, never S3 passing.
		if refused < 1 {
			t.Errorf("%s setup: the shim logged no refused move to %s during call 2; the run cannot be interpreted", tag, f.payloadPath())
		}
		if delta2 < 1 {
			t.Errorf("%s setup: call 2 stub count did not increase (delta %d)", tag, delta2)
		}
		if rec, ok := f.readRecord(tag); ok {
			sgfExpectRecord(t, tag+" setup", rec, head, "fail")
		}
		if t.Failed() {
			return
		}
		if _, err := os.Stat(f.payloadPath()); err == nil {
			t.Errorf("%s: a payload file exists after call 2; want none", tag)
		} else if !os.IsNotExist(err) {
			t.Errorf("%s: stat payload after call 2: %v", tag, err)
		}
		before3 := f.count()
		out3, _ := f.run("{}")
		delta3 := f.count() - before3
		t.Logf("%s call 3 stub delta=%d stdout=%q", tag, delta3, out3)
		if delta3 < 1 {
			t.Errorf("%s: call 3 stub count unchanged (delta %d); want increased — the checks must re-run", tag, delta3)
		}
		if !sgfHasBlock(out3) {
			t.Errorf("%s: call 3 stdout is not a block; stdout=%q", tag, out3)
		}
	})
}

// TestSyncGateFailState_AC014_StaleWindowEqualsRegisteredTimeout — AC-014 (release-blocking).
//
// Contract for the named stale-window variable: exactly one top-level shell
// assignment of an integer literal whose name contains STALE.
func TestSyncGateFailState_AC014_StaleWindowEqualsRegisteredTimeout(t *testing.T) {
	tag := "AC-014 [" + rbClass + "]"
	root := sgfRepoRoot(t)
	hook, err := os.ReadFile(filepath.Join(root, "internal", "template", "templates", ".claude", "hooks", "moai", "sync-phase-quality-gate.sh"))
	if err != nil {
		t.Fatalf("%s: read template hook: %v", tag, err)
	}
	assign := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=([0-9]+)[ \t]*(#.*)?$`)
	var names []string
	var window int
	for _, ln := range strings.Split(string(hook), "\n") {
		m := assign.FindStringSubmatch(ln)
		if m == nil || !strings.Contains(strings.ToUpper(m[1]), "STALE") {
			continue
		}
		names = append(names, m[1])
		window, _ = strconv.Atoi(m[2])
	}
	t.Logf("%s stale-window assignments found: %v", tag, names)

	tmpl, err := os.ReadFile(filepath.Join(root, "internal", "template", "templates", ".claude", "settings.json.tmpl"))
	if err != nil {
		t.Fatalf("%s: read settings template: %v", tag, err)
	}
	timeoutRe := regexp.MustCompile(`"timeout":\s*([0-9]+)`)
	lines := strings.Split(string(tmpl), "\n")
	var timeouts []int
	for i, ln := range lines {
		if !strings.Contains(ln, "sync-phase-quality-gate.sh") {
			continue
		}
		found := false
		for j := i; j < len(lines) && j <= i+8; j++ {
			if m := timeoutRe.FindStringSubmatch(lines[j]); m != nil {
				n, _ := strconv.Atoi(m[1])
				timeouts = append(timeouts, n)
				found = true
				break
			}
			if j > i && strings.Contains(lines[j], "}") {
				break
			}
		}
		for j := i - 1; !found && j >= 0 && j >= i-8; j-- {
			if m := timeoutRe.FindStringSubmatch(lines[j]); m != nil {
				n, _ := strconv.Atoi(m[1])
				timeouts = append(timeouts, n)
				break
			}
			if strings.Contains(lines[j], "{") {
				break
			}
		}
	}
	t.Logf("%s settings timeouts for sync-phase-quality-gate.sh entries: %v", tag, timeouts)

	if len(timeouts) != 1 {
		t.Errorf("%s: want exactly one settings entry timeout for sync-phase-quality-gate.sh; got %v", tag, timeouts)
	}
	switch {
	case len(names) == 0:
		t.Errorf("%s: the template hook declares no named stale-window variable (no NAME=<int> assignment whose name contains STALE)", tag)
	case len(names) > 1:
		t.Errorf("%s: want exactly one stale-window assignment; got %v", tag, names)
	case len(timeouts) == 1 && window != timeouts[0]:
		t.Errorf("%s: stale window %s=%d differs from the settings timeout %d", tag, names[0], window, timeouts[0])
	}
}

// TestSyncGateFailState_T601_WorkTreeContentKeysTheRecord — card t601 (H06 residual).
//
// The gate's checks read the WORK TREE (`go vet ./...`, `go build ./...`), while the
// outcome record named only HEAD. A repair landing in the work tree under an
// unchanged HEAD therefore re-delivered the stored block forever and never re-ran
// the checks. Row R1 is the release-blocking row for that; R3 is its symmetric case
// on a stored pass. R2 is the regression guard that keeps the memo working: with the
// work tree untouched, the stored block must still be re-delivered without re-running
// anything, so "always re-gate" cannot pass this suite.
func TestSyncGateFailState_T601_WorkTreeContentKeysTheRecord(t *testing.T) {
	// repairWorkTree makes the checks pass and changes tracked work-tree content
	// under the same HEAD. The stub lives outside the repository, so only the
	// writeRepoFile call alters the work tree the gate identifies.
	repairWorkTree := func(f *sgfFixture) {
		f.setStub(sgfGoStub(0, 0))
		f.writeRepoFile("main.go", "package main\n\nfunc main() { _ = 42 }\n")
	}
	breakWorkTree := func(f *sgfFixture) {
		f.setStub(sgfGoStub(1, 0))
		f.writeRepoFile("main.go", "package main\n\nfunc main() { _ = 43 }\n")
	}

	t.Run("R1-repaired-tree-regates", func(t *testing.T) {
		tag := "t601 R1 [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		head := f.rev("HEAD")
		out1, _ := f.run("{}")
		if !sgfHasBlock(out1) {
			t.Fatalf("%s setup: call 1 did not block; stdout=%q", tag, out1)
		}
		repairWorkTree(f)
		if f.rev("HEAD") != head {
			t.Fatalf("%s setup: HEAD moved; the row measures an unchanged HEAD", tag)
		}
		before := f.count()
		out2, code := f.run("{}")
		delta := f.count() - before
		t.Logf("%s call 2 stub delta=%d stdout=%q", tag, delta, out2)
		if delta < 1 {
			t.Errorf("%s: stub invoked %d time(s) after the work tree was repaired; want >= 1 — the checks must re-run", tag, delta)
		}
		if sgfHasDecision(out2) {
			t.Errorf("%s: stdout still carries a decision after the repair; stdout=%q", tag, out2)
		}
		if code != 0 {
			t.Errorf("%s: exit %d; want 0", tag, code)
		}
	})

	t.Run("R2-untouched-tree-redelivers", func(t *testing.T) {
		tag := "t601 R2 [" + rgClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		out1, _ := f.run("{}")
		if !sgfHasBlock(out1) {
			t.Fatalf("%s setup: call 1 did not block; stdout=%q", tag, out1)
		}
		before := f.count()
		out2, code := f.run("{}")
		delta := f.count() - before
		t.Logf("%s call 2 stub delta=%d stdout=%q", tag, delta, out2)
		if delta != 0 {
			t.Errorf("%s: stub invoked %d time(s) with the work tree untouched; want 0 — the record must be reused", tag, delta)
		}
		if out2 != out1 {
			t.Errorf("%s: stdout is not byte-identical to the stored block;\n call1=%q\n call2=%q", tag, out1, out2)
		}
		if code != 0 {
			t.Errorf("%s: exit %d; want 0", tag, code)
		}
	})

	// R4 guards the parse this card introduced. Reading the record field by field
	// instead of matching its exact bytes would, on its own, accept a record whose
	// first line looks right and whose remaining lines are anything at all — looser
	// than the exact match it replaced. A multi-line record must re-gate.
	t.Run("R4-multiline-record-regates", func(t *testing.T) {
		tag := "t601 R4 [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 1})
		head := f.rev("HEAD")
		f.writeRecord(head + " pass\nanything at all\n")
		before := f.count()
		out, code := f.run("{}")
		sgfExpectRegate(t, tag, f.count()-before, out)
		if code != 0 {
			t.Errorf("%s: exit %d; want 0", tag, code)
		}
	})

	t.Run("R3-broken-after-pass-regates", func(t *testing.T) {
		tag := "t601 R3 [" + rbClass + "]"
		f := newSGFFixture(t, sgfOpts{vetExit: 0, buildExit: 0})
		out1, _ := f.run("{}")
		if sgfHasDecision(out1) {
			t.Fatalf("%s setup: call 1 was not a silent pass; stdout=%q", tag, out1)
		}
		breakWorkTree(f)
		before := f.count()
		out2, code := f.run("{}")
		delta := f.count() - before
		t.Logf("%s call 2 stub delta=%d stdout=%q", tag, delta, out2)
		if delta < 1 {
			t.Errorf("%s: stub invoked %d time(s) after the work tree broke under a stored pass; want >= 1", tag, delta)
		}
		if !sgfHasBlock(out2) {
			t.Errorf("%s: stdout is not a block after the work tree broke; stdout=%q", tag, out2)
		}
		if code != 0 {
			t.Errorf("%s: exit %d; want 0", tag, code)
		}
	})
}
