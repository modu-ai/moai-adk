package hook

// session_start_drift_fill_test.go — the out-of-band drift-cache fill, hook side.
//
// Covers AC-DCF-001, -002, -003, -007, -008, -009, -010, -011(b) and -012.
// The cross-process burst criterion (AC-DCF-006) lives in its own file, because
// it needs a re-executed helper child and a release mechanism.
//
// ===========================================================================
// OBSERVATION METHOD for "no process was started"
// ===========================================================================
//
// AC-DCF-003's basename case and AC-DCF-012 both assert a NEGATIVE — that no
// process was started. That claim is stated here rather than left implicit,
// because "I did not see a process" is not an observation: process absence is
// unobservable from inside the test and racy from outside it.
//
// The negative is established by a COUNTER AT THE ONLY START POINT, plus a
// structural guard that there is only one such point:
//
//  1. driftFillStartFn is the single place a fill child is ever started.
//     Production assigns (*exec.Cmd).Start to it; a test replaces it with a
//     counting function. "No process was started" is observed as that counter
//     reading zero.
//  2. TestDriftFill_SingleExecChokePoint makes (1) sound rather than a
//     transcription: the package's non-test sources construct the fill child
//     exactly once, so a start that bypassed the counter would have to
//     introduce a second exec.Command site and fail that guard.
//  3. TestDriftFill_RealExecutableIsBlockedUnderGoTest closes the remaining
//     gap — that the counter itself is what makes the tests safe. It runs the
//     PRODUCTION guard against the PRODUCTION start (os.Executable() resolves
//     to this test binary), and observes the counter at zero: the guard, not
//     the stub, is what stops the exec.
//
// The spawn PROBE (driftFillSpawnProbe) is a different instrument and is not
// interchangeable with the start counter. It sits after the suppression gates
// and before the self-invocation guard, so probe=1 / start=0 reads "gated
// before the spawn" while probe=0 reads "never reached the spawn at all" — the
// same distinction internal/statusline's forge spawn gate exists to make.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// driftFillCounters bundles the two instruments described above.
type driftFillCounters struct {
	spawnAttempts atomic.Int32
	startCalls    atomic.Int32
}

// installDriftFillSeams wires both instruments and a deterministic HEAD.
//
// The start seam is replaced by a COUNTING STUB that starts nothing, so no test
// in this file can create a real process however its executable seam is set.
// startCalls therefore measures "reached the start point", not "a process ran".
func installDriftFillSeams(t *testing.T, head string) *driftFillCounters {
	t.Helper()
	c := &driftFillCounters{}

	origProbe, origStart, origHead, origNow := driftFillSpawnProbe, driftFillStartFn, driftFillHeadFn, driftFillNowFn
	origExec := driftFillExecutableFn
	origProbe2 := driftFillPostReadProbe
	t.Cleanup(func() {
		driftFillSpawnProbe, driftFillStartFn, driftFillHeadFn, driftFillNowFn = origProbe, origStart, origHead, origNow
		driftFillExecutableFn = origExec
		driftFillPostReadProbe = origProbe2
	})

	driftFillSpawnProbe = func(string) { c.spawnAttempts.Add(1) }
	driftFillStartFn = func(*exec.Cmd) error {
		c.startCalls.Add(1)
		return nil
	}
	driftFillHeadFn = func(string) (string, error) { return head, nil }
	driftFillPostReadProbe = nil
	return c
}

// driftFillProject returns a project dir with .moai/state present.
func driftFillProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mkStateDir(t, dir)
	return dir
}

// seedDriftCache writes a drift cache entry for head with the given count.
func seedDriftCache(t *testing.T, projectDir, head string, count int) {
	t.Helper()
	payload := map[string]any{
		"head_sha":    head,
		"computed_at": time.Now().UTC().Format(time.RFC3339),
		"count":       count,
		"records":     []any{},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".moai", "state", "drift-cache.json"), raw, 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}
}

// installCachedCountSeam points the deferred advisory's cache resolve at a
// fixed answer.
func installCachedCountSeam(t *testing.T, count int, ok bool) {
	t.Helper()
	orig := driftCachedCountFn
	t.Cleanup(func() { driftCachedCountFn = orig })
	driftCachedCountFn = func(string) (int, bool) { return count, ok }
}

func handleOnce(t *testing.T, projectDir, sessionID string) map[string]any {
	t.Helper()
	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:     sessionID,
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Data, &payload); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	return payload
}

// ---------------------------------------------------------------------------
// AC-DCF-001 / AC-DCF-002 — the miss starts exactly one fill, the hit starts
// none and still renders the advisory. Adopted together: an always-spawning
// mutant satisfies the first, a never-spawning one satisfies the second.
// ---------------------------------------------------------------------------

func TestSessionStart_CacheMissStartsExactlyOneFill(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-miss")
	installCachedCountSeam(t, 0, false)
	completed := registerDeferredScanSeam(t)

	payload := handleOnce(t, dir, "sess-drift-fill-miss")
	waitDeferred(t, completed, 5*time.Second)

	if got := c.spawnAttempts.Load(); got != 1 {
		t.Fatalf("spawn attempts on a cache miss = %d, want exactly 1", got)
	}
	if _, ok := payload["status_drift_warning"]; ok {
		t.Fatalf("a cache miss rendered status_drift_warning; the advisory must be omitted for this session")
	}
}

func TestSessionStart_CacheHitStartsNoFillAndRendersAdvisory(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-hit")
	installCachedCountSeam(t, driftWarningThreshold, true)
	completed := registerDeferredScanSeam(t)

	payload := handleOnce(t, dir, "sess-drift-fill-hit")
	waitDeferred(t, completed, 5*time.Second)

	if got := c.spawnAttempts.Load(); got != 0 {
		t.Fatalf("spawn attempts on a cache hit = %d, want 0", got)
	}
	if _, ok := payload["status_drift_warning"]; !ok {
		t.Fatalf("a cache hit did not render status_drift_warning; payload=%v", payload)
	}
}

// ---------------------------------------------------------------------------
// AC-DCF-003 — the detach contract, and the basename guard.
// ---------------------------------------------------------------------------

func TestDriftFill_DetachContractOnTheConstructedCommand(t *testing.T) {
	dir := driftFillProject(t)
	cmd := buildDriftFillCommand("/usr/local/bin/moai", dir)

	if cmd.Stdin != nil || cmd.Stdout != nil || cmd.Stderr != nil {
		t.Fatalf("the fill child inherits a parent stream: stdin=%v stdout=%v stderr=%v",
			cmd.Stdin, cmd.Stdout, cmd.Stderr)
	}
	if cmd.Dir != dir {
		t.Fatalf("cmd.Dir = %q, want the project dir %q", cmd.Dir, dir)
	}
	joined := strings.Join(cmd.Args, " ")
	for _, want := range []string{"spec", "drift", "--fill-cache", "--fill-timeout"} {
		if !strings.Contains(joined, want) {
			t.Errorf("fill command %q is missing %q", joined, want)
		}
	}
	if !strings.Contains(joined, config.DefaultDriftCacheFillTimeout.String()) {
		t.Errorf("fill command %q does not carry the child deadline %v", joined, config.DefaultDriftCacheFillTimeout)
	}
}

// TestDriftFill_SpawnPathReleasesAndNeverWaits asserts the remaining half of the
// detach contract on the SOURCE of the spawn path rather than on a live exec:
// the parent releases the process handle and never waits on it. A live-exec
// assertion cannot distinguish "never waited" from "waited and returned fast".
func TestDriftFill_SpawnPathReleasesAndNeverWaits(t *testing.T) {
	raw, err := os.ReadFile("session_start_drift_fill.go")
	if err != nil {
		t.Fatalf("read session_start_drift_fill.go: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, "cmd.Process.Release()") {
		t.Error("the spawn path does not call Process.Release() — the child handle is never released")
	}
	if strings.Contains(body, ".Wait()") {
		t.Error("the spawn path calls Wait() — the fill is awaited, which reintroduces the latency this card removes")
	}
}

// TestDriftFill_NonMoaiExecutableStartsNoProcess is the basename case, and the
// place the R2 observation method is exercised: the spawn seam records the
// ATTEMPT (so the gate is demonstrably reached) while the start counter — the
// single point at which a process is actually started — reads zero.
//
// A mutant that never spawns at all also starts no process; the seam-counter
// assertion is what excludes it.
func TestDriftFill_NonMoaiExecutableStartsNoProcess(t *testing.T) {
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-basename")
	driftFillExecutableFn = func() (string, error) {
		return filepath.Join(t.TempDir(), "definitely-not-moai"), nil
	}

	maybeFillDriftCache(dir)

	if got := c.spawnAttempts.Load(); got != 1 {
		t.Fatalf("spawn attempts = %d, want 1 — the gate before the basename guard was not reached", got)
	}
	if got := c.startCalls.Load(); got != 0 {
		t.Fatalf("process starts = %d, want 0 — a non-moai executable was started", got)
	}
}

// TestDriftFill_RealExecutableIsBlockedUnderGoTest runs the PRODUCTION
// self-invocation guard against the PRODUCTION start path: os.Executable() is
// left alone (it resolves to this test binary) and driftFillStartFn wraps the
// real (*exec.Cmd).Start rather than replacing it.
//
// It is what makes the counting stub used everywhere else a safe instrument
// rather than the only thing preventing a fork bomb: if the basename guard ever
// regressed, this test would re-execute the hook test binary and the counter
// would read non-zero — observable, and failed here.
func TestDriftFill_RealExecutableIsBlockedUnderGoTest(t *testing.T) {
	dir := driftFillProject(t)

	origProbe, origStart, origHead := driftFillSpawnProbe, driftFillStartFn, driftFillHeadFn
	t.Cleanup(func() { driftFillSpawnProbe, driftFillStartFn, driftFillHeadFn = origProbe, origStart, origHead })

	var probes, starts int32
	driftFillHeadFn = func(string) (string, error) { return "head-realexec", nil }
	driftFillSpawnProbe = func(string) { probes++ }
	driftFillStartFn = func(cmd *exec.Cmd) error {
		starts++
		return cmd.Start() // the real thing
	}

	maybeFillDriftCache(dir)

	if probes != 1 {
		t.Fatalf("spawn attempts = %d, want 1 — the gate before the guard was not reached", probes)
	}
	if starts != 0 {
		t.Fatalf("the production guard let %d real start(s) through with os.Executable()=%q", starts, os.Args[0])
	}
}

func TestDriftFill_MoaiExecutableIsRecognised(t *testing.T) {
	cases := map[string]bool{
		"/usr/local/bin/moai":      true,
		"/usr/local/bin/moai.exe":  true,
		"/tmp/go-build/hook.test":  false,
		"/usr/local/bin/moai-tool": false,
		"":                         false,
	}
	for path, want := range cases {
		if got := isMoaiExecutable(path); got != want {
			t.Errorf("isMoaiExecutable(%q) = %v, want %v", path, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// AC-DCF-007 — a child that never runs still suppresses the next session.
// ---------------------------------------------------------------------------

func TestDriftFill_BrokenChildStillSuppressesTheNextSession(t *testing.T) {
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-broken")
	// The "child" succeeds at Start() and then does nothing at all: the broken
	// binary, the OOM-kill-at-startup, the mislinked executable. It writes no
	// record, because the PARENT owns that write.
	driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }

	maybeFillDriftCache(dir)
	maybeFillDriftCache(dir)

	if got := c.spawnAttempts.Load(); got != 1 {
		t.Fatalf("spawn attempts across two invocations = %d, want 1 — a broken child produced a respawn loop", got)
	}

	rec, ok := readDriftFillRecord(driftFillRecordPath(dir))
	if !ok {
		t.Fatal("no suppression record on disk — the record was not written by the parent")
	}
	if rec.HeadSHA != "head-broken" {
		t.Fatalf("record head_sha = %q, want %q", rec.HeadSHA, "head-broken")
	}
}

// ---------------------------------------------------------------------------
// AC-DCF-008 — HEAD change re-enables; a future stamp does not suppress; an
// unreadable record is expired; the reclaim is contained; TTL >= deadline+slack.
// ---------------------------------------------------------------------------

func TestDriftFill_HeadChangeReEnablesTheFill(t *testing.T) {
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-A")
	driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }

	maybeFillDriftCache(dir)
	if got := c.spawnAttempts.Load(); got != 1 {
		t.Fatalf("first invocation spawn attempts = %d, want 1", got)
	}
	// HEAD advances well inside the TTL.
	driftFillHeadFn = func(string) (string, error) { return "head-B", nil }
	maybeFillDriftCache(dir)

	if got := c.spawnAttempts.Load(); got != 2 {
		t.Fatalf("spawn attempts after a HEAD change = %d, want 2 — the record is keyed on time alone", got)
	}
}

func TestDriftFill_FutureStampedRecordIsExpired(t *testing.T) {
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-future")
	driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }

	// A clock stepped backwards, or a clone carrying a foreign stamp: without
	// the future test this record reads as "younger than the TTL" forever.
	if !writeDriftFillRecord(driftFillRecordPath(dir), driftFillRecord{
		HeadSHA:   "head-future",
		StartedAt: time.Now().Add(time.Hour),
	}) {
		t.Fatal("seed record")
	}

	maybeFillDriftCache(dir)
	if got := c.spawnAttempts.Load(); got != 1 {
		t.Fatalf("spawn attempts with a future-stamped record = %d, want 1 (treated as expired)", got)
	}
}

func TestDriftFill_UnreadableRecordIsExpired(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"truncated":   `{"head_sha":"head-un`,
		"unparseable": "not json at all",
		"no head":     `{"started_at":"2026-01-01T00:00:00Z"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			dir := driftFillProject(t)
			c := installDriftFillSeams(t, "head-un")
			driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }

			if err := os.WriteFile(driftFillRecordPath(dir), []byte(body), 0o644); err != nil {
				t.Fatalf("seed record: %v", err)
			}

			maybeFillDriftCache(dir)
			if got := c.spawnAttempts.Load(); got != 1 {
				t.Fatalf("spawn attempts with an unreadable record = %d, want 1 (the single stated disposition: expired)", got)
			}
			if _, ok := readDriftFillRecord(driftFillRecordPath(dir)); !ok {
				t.Fatal("the handler did not leave a well-formed record behind")
			}
		})
	}
}

// TestDriftFill_ReclaimIsJudgedAtRemovalTime is AC-DCF-008(b3). A record
// replacement is injected into the window between the handler's read and the
// removal that follows it — the exact window the reclaim TOCTOU lived in — and
// the handler must NOT destroy the replacement.
//
// With the judgement carried over from the earlier read, the handler removes a
// record another handler had just written and spawns a second child. With the
// judgement made at removal time, it re-reads, sees a fresh record, and
// declines.
func TestDriftFill_ReclaimIsJudgedAtRemovalTime(t *testing.T) {
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-toctou")
	driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }

	recordPath := driftFillRecordPath(dir)
	// An EXPIRED record, so the handler's first read judges it reclaimable.
	if !writeDriftFillRecord(recordPath, driftFillRecord{
		HeadSHA:   "head-toctou",
		StartedAt: time.Now().Add(-2 * config.DefaultDriftCacheFillTTL),
	}) {
		t.Fatal("seed expired record")
	}

	injected := driftFillRecord{HeadSHA: "head-toctou", StartedAt: time.Now()}
	fired := false
	driftFillPostReadProbe = func(path string) {
		if fired {
			return
		}
		fired = true
		if !writeDriftFillRecord(path, injected) {
			t.Error("inject replacement record")
		}
	}

	maybeFillDriftCache(dir)

	if !fired {
		t.Fatal("the injection point between the read and the removal was never reached")
	}
	if got := c.spawnAttempts.Load(); got != 0 {
		t.Fatalf("spawn attempts = %d, want 0 — the handler acted on the stale read and destroyed a fresh claim", got)
	}
	got, ok := readDriftFillRecord(recordPath)
	if !ok {
		t.Fatal("the injected replacement was destroyed")
	}
	if !got.StartedAt.Equal(injected.StartedAt.Truncate(0)) && got.StartedAt.Sub(injected.StartedAt).Abs() > time.Millisecond {
		t.Fatalf("record started_at = %v, want the injected replacement %v", got.StartedAt, injected.StartedAt)
	}
}

// TestDriftFill_RecordSuppressionRules pins the three tests as a unit.
func TestDriftFill_RecordSuppressionRules(t *testing.T) {
	now := time.Now()
	ttl := config.DefaultDriftCacheFillTTL
	cases := []struct {
		name string
		rec  driftFillRecord
		want bool
	}{
		{"fresh, same head", driftFillRecord{HeadSHA: "h", StartedAt: now.Add(-time.Second)}, true},
		{"different head", driftFillRecord{HeadSHA: "other", StartedAt: now.Add(-time.Second)}, false},
		{"older than ttl", driftFillRecord{HeadSHA: "h", StartedAt: now.Add(-ttl - time.Second)}, false},
		{"stamped in the future", driftFillRecord{HeadSHA: "h", StartedAt: now.Add(time.Hour)}, false},
		{"empty head", driftFillRecord{HeadSHA: "", StartedAt: now}, false},
	}
	for _, tc := range cases {
		if got := driftFillRecordSuppresses(tc.rec, "h", now, ttl); got != tc.want {
			t.Errorf("%s: suppresses = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestDriftFill_StaleLockIsReclaimable covers the lock's own staleness rule —
// a handler that died inside the critical section must not wedge the fill
// forever — and pins the RESIDUAL that reclaim does not close (see the
// acquireDriftFillClaimLock doc comment).
func TestDriftFill_StaleLockIsReclaimable(t *testing.T) {
	dir := driftFillProject(t)
	installDriftFillSeams(t, "head-lock")
	lockPath := driftFillLockPath(dir)

	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatalf("seed lock: %v", err)
	}
	if acquireDriftFillClaimLock(lockPath) {
		t.Fatal("acquired a lock another handler holds")
	}

	// Age it past the staleness bound.
	old := time.Now().Add(-2 * config.DriftCacheFillLockStaleness)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("age lock: %v", err)
	}
	if !acquireDriftFillClaimLock(lockPath) {
		t.Fatal("a lock older than its staleness bound was not reclaimable — a died-inside handler wedges every later fill")
	}
	_ = os.Remove(lockPath)
}

// ---------------------------------------------------------------------------
// AC-DCF-009 — the miss path spends no join budget on drift.
// ---------------------------------------------------------------------------

// TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift observes the property
// AC-DCF-009 is about: on a cache MISS the deferred advisory step performs no
// in-band drift computation, Handle does not wait for the drift work, and the
// fill that replaces the computation is dispatched out of band.
//
// MEASUREMENT SUBJECT (a run-phase observation-method decision, recorded here
// beside the test). AC-DCF-009's green path states the ratio as "Handle returns
// with elapsed < deferredScanJoinBound / 5". Handle's total wall clock cannot
// carry that ratio for a reason unrelated to this card: its SYNCHRONOUS work —
// config load, session registry, migration, settings — measures ~170 ms on this
// machine, already above the 50 ms figure, and no change to the drift path can
// move it. Asserting the ratio against that total would be red at arrival and
// red forever, which is the "impossible" direction acceptance.md §2 names as
// disqualifying.
//
// The earlier form of this test applied the ratio to two derived durations
// instead (the deferred step alone, and the miss-minus-hit delta). Both were
// wall clock, and a third assertion compared Handle's ENTIRE cost against a
// drift-specific bound at a ~1.4x margin, so the test measured the machine: a
// sibling lane observed 6/6 FAIL at load 29.6-36.2 while idle CI stayed green
// (card t1166). Its "instrument that makes the timings interpretable" — a slow
// driftCountFn that must never be consulted — could not fire either: nothing in
// this package reads driftCountFn any more (it is assigned in session_start.go
// and read by no non-test code), so the injected cost was never paid and the
// assertion was vacuous against ANY implementation. That is the same trap
// TestSessionStart_DeferredScanDoesNotBlockReturn documents in its own comment.
//
// The ratio is therefore replaced by ORDERING observations, which are what the
// property actually says and carry no clock:
//
//	(a) the deferred step on a miss omits the advisory and dispatches the
//	    out-of-band fill (spawn seam = 1, start seam = 0);
//	(b) Handle returns while the deferred drift work is still inside a gate the
//	    TEST holds — so Handle demonstrably did not wait for it. The gate is
//	    installed at driftCachedCountFn, the seam the deferred step actually
//	    consults, and the test asserts the gate was entered, so a seam nobody
//	    calls cannot satisfy (b) the way the old instrument did;
//	(c) the in-band compute seam is never consulted. Dead on this path today, so
//	    it is a REGRESSION GUARD, not a verdict: it is red under a mutant that
//	    re-introduces the in-band computation in computeDeferredAdvisory.
//
// Absolute input lag is deliberately NOT asserted here. That axis is owned by
// TestSessionStart_HandleInputLagBudget, whose 1.5 s bound is the repository's
// convention for a load-sensitive wall-clock assertion: a gross-regression
// guard derived from a measured maximum (n=300) times 2.5, never a latency SLO
// (.moai/reports/t662/verdict.md).
func TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")

	// (c) The in-band compute seam, counted rather than slowed. A sleep here
	// buys nothing: the seam is unreachable from the deferred path, so the cost
	// was never paid — see the header comment.
	var inBandCalls atomic.Int32
	origDrift := driftCountFn
	t.Cleanup(func() { driftCountFn = origDrift })
	driftCountFn = func(context.Context, string) (int, error) {
		inBandCalls.Add(1)
		return driftWarningThreshold, nil
	}

	// (a) The deferred step itself, observed directly.
	t.Run("the deferred step omits the advisory and dispatches the fill", func(t *testing.T) {
		dir := driftFillProject(t)
		c := installDriftFillSeams(t, "head-budget-step")
		installCachedCountSeam(t, 0, false)

		h := &sessionStartHandler{}
		advisory := h.computeDeferredAdvisory(dir, true)

		if _, ok := advisory["status_drift_warning"]; ok {
			t.Fatalf("the deferred step rendered status_drift_warning on a miss; advisory=%v", advisory)
		}
		if got := c.spawnAttempts.Load(); got != 1 {
			t.Fatalf("spawn attempts on a cache miss = %d, want exactly 1 — the compute was not replaced by an out-of-band fill", got)
		}
		if got := c.startCalls.Load(); got != 0 {
			t.Fatalf("process starts = %d, want 0 — the counting stub is the only start point in this binary", got)
		}
	})

	// (b) Handle's relationship to the deferred drift work, observed as an
	// ORDER rather than a duration.
	t.Run("Handle does not wait for the deferred drift work", func(t *testing.T) {
		dir := driftFillProject(t)
		c := installDriftFillSeams(t, "head-budget-handle")

		// The gate: the deferred step's cache resolve blocks here until this
		// test releases it, and answers a MISS when it does. The backstop is NOT
		// the discriminator — it exists so a join-forever mutant fails loudly
		// instead of hanging until the package timeout.
		const gateBackstop = 5 * time.Second
		release := make(chan struct{})
		var gateEntries atomic.Int32
		origCached := driftCachedCountFn
		t.Cleanup(func() { driftCachedCountFn = origCached })
		driftCachedCountFn = func(string) (int, bool) {
			gateEntries.Add(1)
			select {
			case <-release:
			case <-time.After(gateBackstop):
			}
			return 0, false
		}

		completed := registerDeferredScanSeam(t)
		payload := handleOnce(t, dir, "sess-drift-budget-miss")

		// THE ORDERING OBSERVATION: the deferred goroutine cannot have finished,
		// because the gate is still held by this test. If Handle had joined the
		// drift work, it could only have returned after the gate released.
		select {
		case <-completed:
			t.Fatal("Handle returned only after the deferred advisory scan completed — either it waited for the drift work instead of bounding the join, or the deferred step no longer consults the gated seam (the gate-entry assertion below separates the two)")
		default:
		}
		if _, ok := payload["status_drift_warning"]; ok {
			t.Fatalf("a cache miss rendered status_drift_warning; payload=%v", payload)
		}

		close(release)
		waitDeferred(t, completed, 5*time.Second)

		// The positive control over the gate: an instrument nothing consults
		// proves nothing (that is the defect this test was rewritten to remove).
		if got := gateEntries.Load(); got != 1 {
			t.Fatalf("the gated cache resolve was entered %d time(s), want 1 — the seam the ordering assertion rests on is not on the deferred path", got)
		}
		if got := c.spawnAttempts.Load(); got != 1 {
			t.Fatalf("spawn attempts on a cache miss = %d, want exactly 1 — the out-of-band fill was not dispatched", got)
		}
		if got := c.startCalls.Load(); got != 0 {
			t.Fatalf("process starts = %d, want 0", got)
		}
	})

	if got := inBandCalls.Load(); got != 0 {
		t.Fatalf("the in-band drift computation was consulted %d time(s) on the miss path", got)
	}
}

// ---------------------------------------------------------------------------
// AC-DCF-010 — every fill failure path is fail-open, and the success path
// still spawns.
// ---------------------------------------------------------------------------

func TestDriftFill_EveryFailurePathIsFailOpen(t *testing.T) {
	cases := []struct {
		name       string
		arrange    func(t *testing.T, dir string)
		wantSpawns int32
	}{
		{
			name: "positive control: nothing injected",
			arrange: func(*testing.T, string) {
				driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }
			},
			wantSpawns: 1,
		},
		{
			name: "executable unresolvable",
			arrange: func(*testing.T, string) {
				driftFillExecutableFn = func() (string, error) { return "", os.ErrNotExist }
			},
			wantSpawns: 1, // the gate is reached; the spawn is not
		},
		{
			name: "HEAD unresolvable",
			arrange: func(*testing.T, string) {
				driftFillHeadFn = func(string) (string, error) { return "", os.ErrNotExist }
			},
			wantSpawns: 0,
		},
		{
			name: "state dir unwritable",
			arrange: func(t *testing.T, dir string) {
				// A FILE where the state directory must be: MkdirAll fails, so
				// neither the lock nor the record can be created.
				stateDir := filepath.Join(dir, ".moai", "state")
				if err := os.RemoveAll(stateDir); err != nil {
					t.Fatalf("clear state dir: %v", err)
				}
				if err := os.WriteFile(stateDir, []byte("not a dir"), 0o644); err != nil {
					t.Fatalf("block state dir: %v", err)
				}
				driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }
			},
			wantSpawns: 0,
		},
		{
			name: "lock already held",
			arrange: func(t *testing.T, dir string) {
				if err := os.WriteFile(driftFillLockPath(dir), nil, 0o644); err != nil {
					t.Fatalf("hold lock: %v", err)
				}
				driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }
			},
			wantSpawns: 0,
		},
		{
			name: "Start() errors",
			arrange: func(*testing.T, string) {
				driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }
				driftFillStartFn = func(*exec.Cmd) error { return os.ErrPermission }
			},
			wantSpawns: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ANTHROPIC_BASE_URL", "")
			dir := driftFillProject(t)
			c := installDriftFillSeams(t, "head-failopen")
			installCachedCountSeam(t, 0, false)
			tc.arrange(t, dir)

			completed := registerDeferredScanSeam(t)
			payload := handleOnce(t, dir, "sess-failopen-"+strings.ReplaceAll(tc.name, " ", "-"))
			waitDeferred(t, completed, 5*time.Second)

			// Fail-open: Handle returned normally with its other keys intact.
			if len(payload) == 0 {
				t.Fatal("Handle returned an empty payload")
			}
			if got := c.spawnAttempts.Load(); got != tc.wantSpawns {
				t.Fatalf("spawn attempts = %d, want %d", got, tc.wantSpawns)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-DCF-011(b) — the opt-out, read through the resolved config value.
// ---------------------------------------------------------------------------

func TestSessionStart_OptOutStartsNoFill(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-optout")
	installCachedCountSeam(t, 0, false)

	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "workflow:\n    drift_cache_fill:\n        enabled: false\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	completed := registerDeferredScanSeam(t)
	handleOnce(t, dir, "sess-drift-optout")
	waitDeferred(t, completed, 5*time.Second)

	if got := c.spawnAttempts.Load(); got != 0 {
		t.Fatalf("spawn attempts with the opt-out set = %d, want 0", got)
	}
	if _, err := os.Stat(driftFillRecordPath(dir)); err == nil {
		t.Fatal("the disabled fill still wrote a suppression record")
	}
}

// ---------------------------------------------------------------------------
// AC-DCF-012 — the hook test binary starts zero fill children.
// ---------------------------------------------------------------------------

// TestSessionStart_SyncDeferredScansStartNoFill is REQ-DCF-015: while the
// async seam is off — the test binary's default, and the state a handler
// carrying WithSynchronousDeferredScans is in — the handler starts no fill.
func TestSessionStart_SyncDeferredScansStartNoFill(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-sync")
	installCachedCountSeam(t, 0, false)

	// No registerDeferredScanSeam: deferredScansAsync stays false (TestMain's
	// setting for the whole binary), so the inline path runs.
	handleOnce(t, dir, "sess-drift-sync")

	if got := c.spawnAttempts.Load(); got != 0 {
		t.Fatalf("spawn attempts on the inline path = %d, want 0", got)
	}
}

// TestDriftFill_SingleExecChokePoint is the structural guard that makes the
// start counter a sound witness: the package's non-test sources construct the
// fill child exactly once. A second spawn path added later fails this test
// rather than slipping past a counter that never sees it.
func TestDriftFill_SingleExecChokePoint(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	sites := 0
	var where []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.Contains(line, `"--fill-cache"`) {
				sites++
				where = append(where, name+": "+strings.TrimSpace(line))
			}
		}
	}
	if sites != 1 {
		t.Fatalf("the fill child is constructed at %d site(s), want exactly 1:\n%s", sites, strings.Join(where, "\n"))
	}
}

// ---------------------------------------------------------------------------
// Production wiring — the seams above must point at the real implementations.
// ---------------------------------------------------------------------------

// runGitForDriftFill initialises a one-commit repository in dir and returns its
// HEAD SHA.
func runGitForDriftFill(t *testing.T, dir string) (string, error) {
	t.Helper()
	env := append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	)
	run := func(args ...string) ([]byte, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		return cmd.CombinedOutput()
	}
	if out, err := run("init", "--quiet", "--initial-branch=main"); err != nil {
		return "", fmt.Errorf("git init: %v (%s)", err, out)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("fixture\n"), 0o644); err != nil {
		return "", err
	}
	if out, err := run("add", "-A"); err != nil {
		return "", fmt.Errorf("git add: %v (%s)", err, out)
	}
	if out, err := run("commit", "--quiet", "-m", "fixture"); err != nil {
		return "", fmt.Errorf("git commit: %v (%s)", err, out)
	}
	out, err := run("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %v (%s)", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// TestDriftCachedCountFnIsWiredToTheRealCache exercises the production cache
// read end to end on a real git repository, so the seam used by the tests above
// is not the only thing ever proven to work.
func TestDriftCachedCountFnIsWiredToTheRealCache(t *testing.T) {
	dir := t.TempDir()
	mkStateDir(t, dir)

	head, err := runGitForDriftFill(t, dir)
	if err != nil {
		t.Skipf("git unavailable: %v", err)
	}

	if _, ok := spec.CachedDriftCount(dir); ok {
		t.Fatal("precondition: the cache reports a hit before anything was written")
	}
	seedDriftCache(t, dir, head, 7)
	count, ok := spec.CachedDriftCount(dir)
	if !ok {
		t.Fatal("the seeded cache did not read back as a hit")
	}
	if count != 7 {
		t.Fatalf("cached count = %d, want 7", count)
	}

	// A HEAD-mismatched entry is a miss, which is what makes the fill fire.
	seedDriftCache(t, dir, "0000000000000000000000000000000000000000", 7)
	if _, ok := spec.CachedDriftCount(dir); ok {
		t.Fatal("an entry written against a different HEAD read back as a hit")
	}
}
