package hook

// session_start_drift_fill_burst_test.go — AC-DCF-006 clause (a): a
// SIMULTANEOUS BURST of separate PROCESSES produces exactly one spawn.
//
// ===========================================================================
// Why this file exists at all, and why goroutines would not do
// ===========================================================================
//
// The property under test is cross-process create-versus-create. An in-process
// test cannot witness it: a sync.Mutex around a read-check-then-write makes a
// goroutine burst green while two real processes still race, and — the
// iteration-3 plan-audit's finding, preserved verbatim in plan.md §G —
//
//	"an error value is not evidence of a syscall — fs.ErrExist is a sentinel
//	 any code can wrap."
//
// So the burst is N separate processes: this test binary re-executed, the
// in-tree shape at internal/cli/gate_lock_cli_test.go.
//
// ===========================================================================
// The release mechanism, and why a bare Start() loop is not one
// ===========================================================================
//
// Starting N children in a loop does NOT produce contention. Process start-up
// — fork/exec plus Go runtime init — is tens of milliseconds and staggered,
// while the claim window is sub-millisecond, so child 1 typically finishes
// claiming before child 2 is scheduled. Under a NON-ATOMIC implementation with
// no overlap the census is still exactly one winner: green against a broken
// implementation. That is the empty-sweep principle applied to concurrency — a
// green whose race never happened is uninterpreted output.
//
// Two mechanisms are used together, because each closes a different half:
//
//  1. A SHARED START INSTANT. Every child announces readiness, then spins on a
//     gate file; the parent writes the gate only once all N are ready, and the
//     gate CARRIES A TARGET WALL-CLOCK NANOTIME. Children busy-wait to that
//     instant, so the release is aligned to microseconds rather than to the
//     filesystem's notification latency.
//
//  2. A WIDENED CRITICAL SECTION. The winner holds the claim lock for
//     driftFillBurstHoldMS milliseconds, installed in the CHILD from an env var
//     through the in-critical-section probe seam. This is what makes the
//     overlap a guarantee rather than a hope: every loser's claim attempt
//     necessarily falls inside the winner's section. It is also what makes the
//     test HOSTILE to the mutant — for an implementation with no containment,
//     the hold sits between the read and the write, opening the hazard window
//     as wide as the hold.
//
// ===========================================================================
// The overlap is ASSERTED, never assumed
// ===========================================================================
//
// Each child reports the wall-clock interval of its own claim attempt. The test
// then computes how many of those intervals cover a common instant, and FAILS
// when fewer than two do — without at least two simultaneous attempts, the
// census below says nothing about atomicity. The measured simultaneity and the
// full per-child timings are logged either way, so a reader sees whether the
// burst was 8-wide or barely 2-wide.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// helperDriftFillBurstDirEnv carries the shared project directory; its
	// presence is what turns the helper Test... into a child.
	helperDriftFillBurstDirEnv = "MOAI_TEST_DRIFT_FILL_BURST_DIR"
	// helperDriftFillBurstOutEnv is where this child writes its report.
	helperDriftFillBurstOutEnv = "MOAI_TEST_DRIFT_FILL_BURST_OUT"
	// helperDriftFillBurstGateEnv is the start gate: the file the parent writes
	// once every child is ready, carrying the target start nanotime.
	helperDriftFillBurstGateEnv = "MOAI_TEST_DRIFT_FILL_BURST_GATE"
	// helperDriftFillBurstHoldEnv is the winner's in-critical-section hold, in
	// milliseconds — the overlap guarantee.
	helperDriftFillBurstHoldEnv = "MOAI_TEST_DRIFT_FILL_BURST_HOLD_MS"

	driftFillBurstChildren = 8
	driftFillBurstHoldMS   = 400
	// driftFillBurstReadyBound caps the child's own wait for the gate, so the
	// child exits on its own whatever the parent does. -test.timeout caps it
	// again from outside. There is no trailing kill anywhere in this file.
	driftFillBurstReadyBound = 30 * time.Second
)

// driftFillBurstReport is what each child writes back.
type driftFillBurstReport struct {
	PID       int   `json:"pid"`
	Won       bool  `json:"won"`
	Spawns    int32 `json:"spawns"`
	AttemptT0 int64 `json:"attempt_t0_ns"`
	AttemptT1 int64 `json:"attempt_t1_ns"`
}

// TestDriftFillBurstHelper is not a test. Re-executed as a child by
// TestDriftFill_SimultaneousBurstProducesOneSpawn, it waits on the shared start
// gate, makes exactly ONE claim attempt, reports the attempt's interval and
// outcome, and exits.
//
// The child BOUNDS ITSELF: one attempt, then exit; the gate wait is capped by
// driftFillBurstReadyBound; -test.timeout caps the whole process from outside.
func TestDriftFillBurstHelper(t *testing.T) {
	dir := os.Getenv(helperDriftFillBurstDirEnv)
	out := os.Getenv(helperDriftFillBurstOutEnv)
	gate := os.Getenv(helperDriftFillBurstGateEnv)
	if dir == "" || out == "" || gate == "" {
		t.Skip("helper process only")
	}

	// Count spawn attempts without ever starting a process: os.Executable() in
	// a test binary is not a moai binary, so the basename guard blocks the exec
	// regardless — the probe is the instrument, exactly as in the sibling file.
	var spawns int32
	driftFillSpawnProbe = func(string) { spawns++ }
	driftFillHeadFn = func(string) (string, error) { return "burst-head", nil }

	// The winner holds the critical section open. Installed HERE, in the child,
	// so the widened window is a property of the test run rather than of the
	// implementation.
	if raw := os.Getenv(helperDriftFillBurstHoldEnv); raw != "" {
		ms, err := strconv.Atoi(raw)
		if err != nil {
			t.Fatalf("%s=%q: %v", helperDriftFillBurstHoldEnv, raw, err)
		}
		driftFillPostReadProbe = func(string) { time.Sleep(time.Duration(ms) * time.Millisecond) }
	}

	// Announce readiness, then wait for the gate.
	if err := os.WriteFile(out+".ready", []byte("1"), 0o644); err != nil {
		t.Fatalf("write ready marker: %v", err)
	}
	deadline := time.Now().Add(driftFillBurstReadyBound)
	var targetNS int64
	for {
		raw, err := os.ReadFile(gate)
		if err == nil {
			targetNS, err = strconv.ParseInt(string(raw), 10, 64)
			if err == nil {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("start gate never opened within %v", driftFillBurstReadyBound)
		}
		time.Sleep(200 * time.Microsecond)
	}
	// Busy-wait to the shared instant: the last few hundred microseconds must
	// not go through the scheduler, or the release is only as tight as a sleep.
	for time.Now().UnixNano() < targetNS {
	}

	t0 := time.Now()
	maybeFillDriftCache(dir)
	t1 := time.Now()

	// "Won" is read from the record on disk rather than from a return value:
	// maybeFillDriftCache reports nothing, and the record is the artefact the
	// claim actually produces.
	report := driftFillBurstReport{
		PID:       os.Getpid(),
		Won:       spawns > 0,
		Spawns:    spawns,
		AttemptT0: t0.UnixNano(),
		AttemptT1: t1.UnixNano(),
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if err := os.WriteFile(out, raw, 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
}

// TestDriftFill_SimultaneousBurstProducesOneSpawn is AC-DCF-006 clause (a).
func TestDriftFill_SimultaneousBurstProducesOneSpawn(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	reportDir := t.TempDir()
	gate := filepath.Join(reportDir, "gate")

	type child struct {
		cmd  *exec.Cmd
		out  string
		logs *bytes.Buffer
	}
	children := make([]*child, driftFillBurstChildren)

	for i := range children {
		out := filepath.Join(reportDir, fmt.Sprintf("child-%d.json", i))
		cmd := exec.Command(os.Args[0],
			"-test.run=^TestDriftFillBurstHelper$",
			"-test.timeout=60s",
		)
		cmd.Env = append(os.Environ(),
			helperDriftFillBurstDirEnv+"="+dir,
			helperDriftFillBurstOutEnv+"="+out,
			helperDriftFillBurstGateEnv+"="+gate,
			helperDriftFillBurstHoldEnv+"="+strconv.Itoa(driftFillBurstHoldMS),
		)
		logs := &bytes.Buffer{}
		cmd.Stdout, cmd.Stderr = logs, logs
		children[i] = &child{cmd: cmd, out: out, logs: logs}
	}

	// Every child bounds itself; Wait() below is the cleanup, and the helper's
	// own gate bound plus -test.timeout are what guarantee it terminates even
	// if this process dies first.
	for _, c := range children {
		if err := c.cmd.Start(); err != nil {
			t.Fatalf("start child: %v", err)
		}
	}
	t.Cleanup(func() {
		for _, c := range children {
			_ = c.cmd.Wait()
		}
	})

	// Wait until every child is spinning on the gate. Only then is the release
	// simultaneous rather than staggered by process start-up.
	deadline := time.Now().Add(45 * time.Second)
	for {
		ready := 0
		for _, c := range children {
			if _, err := os.Stat(c.out + ".ready"); err == nil {
				ready++
			}
		}
		if ready == len(children) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("only %d of %d children reached the start gate", ready, len(children))
		}
		time.Sleep(2 * time.Millisecond)
	}

	target := time.Now().Add(25 * time.Millisecond).UnixNano()
	if err := os.WriteFile(gate, []byte(strconv.FormatInt(target, 10)), 0o644); err != nil {
		t.Fatalf("open the start gate: %v", err)
	}

	var wg sync.WaitGroup
	for _, c := range children {
		wg.Add(1)
		go func(c *child) {
			defer wg.Done()
			_ = c.cmd.Wait()
		}(c)
	}
	waited := make(chan struct{})
	go func() { wg.Wait(); close(waited) }()
	select {
	case <-waited:
	case <-time.After(90 * time.Second):
		t.Fatal("children did not finish within 90s")
	}

	// Census.
	reports := make([]driftFillBurstReport, 0, len(children))
	for i, c := range children {
		raw, err := os.ReadFile(c.out)
		if err != nil {
			t.Fatalf("child %d wrote no report: %v\n%s", i, err, c.logs)
		}
		var r driftFillBurstReport
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatalf("child %d report unparseable: %v", i, err)
		}
		reports = append(reports, r)
	}
	if len(reports) != len(children) {
		t.Fatalf("%d of %d children reported in", len(reports), len(children))
	}

	// ---- the overlap assertion, before the census is read as evidence ----
	simultaneity, base := maxSimultaneousAttempts(reports)
	for _, r := range reports {
		t.Logf("child pid=%d won=%v spawns=%d attempt=[%.3fms, %.3fms]",
			r.PID, r.Won, r.Spawns,
			float64(r.AttemptT0-base)/1e6, float64(r.AttemptT1-base)/1e6)
	}
	t.Logf("maximum simultaneous claim attempts: %d of %d", simultaneity, len(reports))
	if simultaneity < 2 {
		t.Fatalf("no two claim attempts overlapped (max simultaneity %d) — the burst never raced, so the census below would be uninterpreted output",
			simultaneity)
	}

	// ---- the census itself ----
	winners, spawns := 0, int32(0)
	for _, r := range reports {
		if r.Won {
			winners++
		}
		spawns += r.Spawns
	}
	if winners != 1 {
		t.Fatalf("%d of %d contending PROCESSES won the claim, want exactly 1", winners, len(reports))
	}
	if spawns != 1 {
		t.Fatalf("total spawn attempts across the burst = %d, want exactly 1", spawns)
	}

	// Exactly one record, naming the burst HEAD, written whole.
	rec, ok := readDriftFillRecord(driftFillRecordPath(dir))
	if !ok {
		t.Fatal("the winner left no readable suppression record")
	}
	if rec.HeadSHA != "burst-head" {
		t.Fatalf("record head_sha = %q, want %q", rec.HeadSHA, "burst-head")
	}
	// The companion lock is released, not leaked.
	if _, err := os.Stat(driftFillLockPath(dir)); err == nil {
		t.Fatal("the claim lock was left behind — a later fill would have to wait out its staleness bound")
	}
}

// maxSimultaneousAttempts returns the largest number of claim-attempt intervals
// covering a common instant, and the earliest attempt start (as a log base).
//
// Interval overlap counted by a sweep over endpoints: a start raises the
// running count, an end lowers it, and the maximum running count is the
// simultaneity. Ends are processed before starts at an equal timestamp, so two
// intervals that merely touch are not counted as overlapping.
func maxSimultaneousAttempts(reports []driftFillBurstReport) (int, int64) {
	type event struct {
		at    int64
		delta int
	}
	events := make([]event, 0, 2*len(reports))
	base := int64(0)
	for i, r := range reports {
		if i == 0 || r.AttemptT0 < base {
			base = r.AttemptT0
		}
		events = append(events, event{at: r.AttemptT0, delta: 1}, event{at: r.AttemptT1, delta: -1})
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].at != events[j].at {
			return events[i].at < events[j].at
		}
		return events[i].delta < events[j].delta
	})
	cur, best := 0, 0
	for _, e := range events {
		cur += e.delta
		if cur > best {
			best = cur
		}
	}
	return best, base
}

// TestDriftFill_ExistingRecordSuppresses is AC-DCF-006 clause (b).
//
// It is SEQUENTIAL by construction and therefore says NOTHING about the
// cross-process property: a mutex-guarded os.Stat returning an error that wraps
// fs.ErrExist passes it while two real processes still race. Clause (a) above
// is the only witness for atomicity; this clause only pins that a pre-existing
// fresh record suppresses and is left untouched.
func TestDriftFill_ExistingRecordSuppresses(t *testing.T) {
	dir := driftFillProject(t)
	c := installDriftFillSeams(t, "head-suppress")
	driftFillExecutableFn = func() (string, error) { return "/usr/local/bin/moai", nil }

	recordPath := driftFillRecordPath(dir)
	if !writeDriftFillRecord(recordPath, driftFillRecord{HeadSHA: "head-suppress", StartedAt: time.Now()}) {
		t.Fatal("seed record")
	}
	before, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("read seeded record: %v", err)
	}

	maybeFillDriftCache(dir)

	if got := c.spawnAttempts.Load(); got != 0 {
		t.Fatalf("spawn attempts with a fresh record present = %d, want 0", got)
	}
	after, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("read record after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("the suppressed handler rewrote the record\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestDriftFill_SingleClaimChokePointTargetsTheLock is AC-DCF-006 clause (c):
// the package's non-test sources contain exactly ONE atomicfile.Claim call site
// on the fill path, and its argument is the record's COMPANION lock — never the
// record itself.
//
// This is the clause that kills the three mutants that drove this design —
// the plain write, the in-process mutex, and the os.Stat-then-fs.ErrExist
// counter-mutant — INDEPENDENTLY OF TIMING: none of them has an
// atomicfile.Claim call site at the lock path.
func TestDriftFill_SingleClaimChokePointTargetsTheLock(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	var sites []string
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
			if strings.Contains(line, "atomicfile.Claim(") {
				sites = append(sites, name+": "+strings.TrimSpace(line))
			}
		}
	}
	// Claim is called twice on the same path — the attempt and the single
	// post-reclaim re-attempt — so the assertion is on the ARGUMENT, not on a
	// raw count: every site must target the lock, and none the record.
	if len(sites) == 0 {
		t.Fatal("no atomicfile.Claim call site on the fill path — the claim is not an exclusive create")
	}
	for _, s := range sites {
		if !strings.Contains(s, "lockPath") {
			t.Fatalf("an atomicfile.Claim site does not target the companion lock: %s", s)
		}
		if strings.Contains(s, "recordPath") {
			t.Fatalf("atomicfile.Claim targets the RECORD, which reopens the reclaim TOCTOU and the empty-record window: %s", s)
		}
	}

	// And the lock path really is the record's companion.
	dir := t.TempDir()
	if got, want := driftFillLockPath(dir), driftFillRecordPath(dir)+".lock"; got != want {
		t.Fatalf("driftFillLockPath = %q, want %q", got, want)
	}
}
