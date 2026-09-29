package discovery

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The classifier matrix (AC-004/005/011/017, design.md §D): every test runs
// on injected fakes with zero platform dependence, under the REQ-012 sandbox
// (project directory outside the repository's worktree set). The seams are
// restored on cleanup so no test leaks a fake into another.

// fakeProcess is one staged candidate: what the injected readers answer for a
// pid.
type fakeProcess struct {
	fingerprint  string
	state        homestate.ProcessIdentityState
	argv         []string
	argvOK       bool
	env          map[string]string
	envOK        bool
	cwd          string
	cwdOK        bool
}

// stageFakes swaps every probe seam for fakes answering the staged processes.
func stageFakes(t *testing.T, procs map[int]fakeProcess) {
	t.Helper()
	origProbe, origArgv, origEnv, origCwd := probeIdentity, processArgv, processEnv, processCwd
	t.Cleanup(func() {
		probeIdentity, processArgv, processEnv, processCwd = origProbe, origArgv, origEnv, origCwd
	})
	probeIdentity = func(pid int) (string, homestate.ProcessIdentityState) {
		p, ok := procs[pid]
		if !ok {
			return "", homestate.ProcessIdentityDead
		}
		return p.fingerprint, p.state
	}
	processArgv = func(_ context.Context, pid int) ([]string, bool) {
		p, ok := procs[pid]
		return p.argv, ok && p.argvOK
	}
	processEnv = func(_ context.Context, pid int) (map[string]string, bool) {
		p, ok := procs[pid]
		return p.env, ok && p.envOK
	}
	processCwd = func(_ context.Context, pid int) (string, bool) {
		p, ok := procs[pid]
		return p.cwd, ok && p.cwdOK
	}
}

// stageCandidates injects the enumerated candidate list.
func stageCandidates(t *testing.T, pids ...int) {
	t.Helper()
	orig := candidatePIDs
	t.Cleanup(func() { candidatePIDs = orig })
	candidatePIDs = func(context.Context, string) ([]int, error) { return pids, nil }
}

// liveLeader returns a fully-verified-shaped candidate: live, named leader,
// carrying runID, sitting in dir.
func liveLeader(pid int, runID, dir string) fakeProcess {
	return fakeProcess{
		fingerprint: "1700000000." + strconv.Itoa(pid),
		state:       homestate.ProcessIdentityLive,
		argv:        []string{"claude", "--permission-mode", "default", "--name", "leader"},
		argvOK:      true,
		env:         map[string]string{config.EnvMoaiKanbanID: runID},
		envOK:       true,
		cwd:         dir,
		cwdOK:       true,
	}
}

// discoverySandbox is the REQ-012 project directory: outside this repository.
func discoverySandbox(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "discovery-project")
}

// The happy shape of AC-001's Given at the discovery level: one live candidate
// the probe can verify verifies, carrying its own run identity.
func TestDiscoverLeaderVerifiesLiveLeader(t *testing.T) {
	dir := discoverySandbox(t)
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: liveLeader(4242, "runlead1", dir)})

	verified, err := DiscoverLeader(context.Background(), dir, "leader")
	if err != nil {
		t.Fatalf("DiscoverLeader: %v", err)
	}
	if len(verified) != 1 {
		t.Fatalf("verified = %d, want 1 (%+v)", len(verified), verified)
	}
	v := verified[0]
	if v.RunID != "runlead1" || v.PID != 4242 || v.Name != "leader" {
		t.Errorf("verified = %+v, want run runlead1 pid 4242 name leader", v)
	}
	if v.ProcessStart == "" {
		t.Errorf("verified leader carries an empty process-start fingerprint")
	}
	if v.Basis == "" {
		t.Errorf("verified leader carries an empty basis (AC-008 rides this into the resume event)")
	}
}

// AC-005 — record presence is never liveness: a candidate the enumeration
// surfaced (a socket file, a run row) whose pid probes dead is declined.
// Mutation probe: deleting the liveness check (accepting on record presence
// alone) must turn this test red.
func TestDiscoverLeaderDeclinesDeadPidDespiteRecordPresence(t *testing.T) {
	dir := discoverySandbox(t)
	dead := liveLeader(4242, "runlead1", dir)
	dead.state = homestate.ProcessIdentityDead
	dead.fingerprint = ""
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: dead})

	verified, err := DiscoverLeader(context.Background(), dir, "leader")
	if err != nil {
		t.Fatalf("DiscoverLeader: %v", err)
	}
	if len(verified) != 0 {
		t.Errorf("dead-pid candidate verified = %+v, want none (record presence is not liveness, AC-005)", verified)
	}
}

// REQ-002's "together with": a pid that answers alive but whose process-start
// fingerprint cannot be read is indeterminate, and indeterminate never
// verifies — the resume stamp must be probe-measured, so an unreadable
// fingerprint is as useless as a dead pid. (Discovery's direction of the
// fingerprint predicate has no recorded side to mismatch: the measured value
// IS what the resume stamps, which is what makes it retire-grade.)
func TestDiscoverLeaderDeclinesIndeterminateIdentity(t *testing.T) {
	dir := discoverySandbox(t)
	fuzzy := liveLeader(4242, "runlead1", dir)
	fuzzy.fingerprint = ""
	fuzzy.state = homestate.ProcessIdentityIndeterminate
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: fuzzy})

	verified, _ := DiscoverLeader(context.Background(), dir, "leader")
	if len(verified) != 0 {
		t.Errorf("indeterminate-identity candidate verified = %+v, want none (REQ-002: live AND fingerprint)", verified)
	}
}

// AC-011 at the classifier level — a live candidate that does not carry the
// targeted label is not the leader the operator named.
func TestDiscoverLeaderDeclinesLabelMismatch(t *testing.T) {
	dir := discoverySandbox(t)
	stageCandidates(t, 4242)
	procs := map[int]fakeProcess{4242: liveLeader(4242, "runlead1", dir)}
	stageFakes(t, procs)

	// Target leader-2, candidate names leader.
	verified, _ := DiscoverLeader(context.Background(), dir, "leader-2")
	if len(verified) != 0 {
		t.Errorf("label-mismatch candidate verified = %+v, want none (AC-011)", verified)
	}

	// A lane-shaped name never passes as a leader, even when targeted.
	lane := liveLeader(4242, "runlead1", dir)
	lane.argv = []string{"claude", "--name", "lane-2"}
	stageFakes(t, map[int]fakeProcess{4242: lane})
	verified, _ = DiscoverLeader(context.Background(), dir, "lane-2")
	if len(verified) != 0 {
		t.Errorf("lane-shaped candidate verified = %+v, want none (label evidence is leader-specific)", verified)
	}
}

// AC-004 at the classifier level — discovery returns EVERY verified leader;
// selection among them is the caller's prohibition (REQ-005), and returning
// one of two here would be that selection.
func TestDiscoverLeaderReturnsBothOnMultiLeaders(t *testing.T) {
	dir := discoverySandbox(t)
	stageCandidates(t, 4242, 4243)
	stageFakes(t, map[int]fakeProcess{
		4242: liveLeader(4242, "runaaaa1", dir),
		4243: liveLeader(4243, "runbbbb2", dir),
	})

	verified, err := DiscoverLeader(context.Background(), dir, "leader")
	if err != nil {
		t.Fatalf("DiscoverLeader: %v", err)
	}
	if len(verified) != 2 {
		t.Fatalf("verified = %d, want both candidates (AC-004: fail closed names them, never selects)", len(verified))
	}
	ids := map[string]bool{verified[0].RunID: true, verified[1].RunID: true}
	if !ids["runaaaa1"] || !ids["runbbbb2"] {
		t.Errorf("verified runs = %v, want both runaaaa1 and runbbbb2", ids)
	}
}

// AC-017 leg (a) — a live, correctly-named candidate whose cwd canonicalizes
// to a DIFFERENT project is not this project's leader.
func TestDiscoverLeaderDeclinesCrossProjectCandidate(t *testing.T) {
	dir := discoverySandbox(t)
	other := liveLeader(4242, "runlead1", filepath.Join(t.TempDir(), "another-project"))
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: other})

	verified, _ := DiscoverLeader(context.Background(), dir, "leader")
	if len(verified) != 0 {
		t.Errorf("cross-project candidate verified = %+v, want none (AC-017 leg a)", verified)
	}
}

// AC-017 leg (b) — the run identity cannot be read, so the candidate is
// declined: the run id is never guessed, defaulted, or minted.
func TestDiscoverLeaderDeclinesUnreadableEnv(t *testing.T) {
	dir := discoverySandbox(t)
	blind := liveLeader(4242, "runlead1", dir)
	blind.envOK = false
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: blind})

	verified, _ := DiscoverLeader(context.Background(), dir, "leader")
	if len(verified) != 0 {
		t.Errorf("unreadable-env candidate verified = %+v, want none (AC-017 leg b)", verified)
	}
}

// An env value that cannot name a broker address declines the same way.
func TestDiscoverLeaderDeclinesUnparseableRunID(t *testing.T) {
	dir := discoverySandbox(t)
	garbage := liveLeader(4242, "../not-a-run-id", dir)
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: garbage})

	verified, _ := DiscoverLeader(context.Background(), dir, "leader")
	if len(verified) != 0 {
		t.Errorf("unparseable-run-id candidate verified = %+v, want none (the id is a broker address)", verified)
	}
}

// Where the platform cannot read the cwd, the candidate's own run id
// appearing in THIS project's run records is the membership evidence; with
// neither evidence, the candidate declines.
func TestDiscoverLeaderMembershipFallbackOnUnreadableCwd(t *testing.T) {
	dir := discoverySandbox(t)
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("MOAI_HOME", filepath.Join(root, ".moai"))

	blind := liveLeader(4242, "runlead1", "")
	blind.cwdOK = false
	stageCandidates(t, 4242)
	stageFakes(t, map[int]fakeProcess{4242: blind})

	// No run record: the fallback has nothing to stand on.
	verified, _ := DiscoverLeader(context.Background(), dir, "leader")
	if len(verified) != 0 {
		t.Errorf("unprovable candidate verified = %+v, want none (cwd unreadable AND no local run record)", verified)
	}

	// A local run record for the same id makes the candidate a member.
	db, err := homestate.OpenFactory(dir)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: "runlead1", Backend: "glm", LeadPID: 4242, LeadProcessStart: "1700000000.4242"}); err != nil {
		t.Fatalf("record fixture run: %v", err)
	}
	verified, err = DiscoverLeader(context.Background(), dir, "leader")
	if err != nil {
		t.Fatalf("DiscoverLeader: %v", err)
	}
	if len(verified) != 1 {
		t.Fatalf("verified = %d, want 1 via the run-record membership evidence", len(verified))
	}
	if verified[0].RunID != "runlead1" {
		t.Errorf("verified run id = %q, want runlead1", verified[0].RunID)
	}
}

// The same pid surfacing from two enumeration sources is ONE candidate: a
// duplicate would masquerade as the multi-leader ambiguity (AC-004's
// fail-closed path must never fire for one real leader).
func TestDiscoverLeaderDeduplicatesCandidates(t *testing.T) {
	dir := discoverySandbox(t)
	stageCandidates(t, 4242, 4242)
	stageFakes(t, map[int]fakeProcess{4242: liveLeader(4242, "runlead1", dir)})

	verified, _ := DiscoverLeader(context.Background(), dir, "leader")
	if len(verified) != 1 {
		t.Errorf("deduplicated candidate verified = %d, want 1", len(verified))
	}
}

// The enumeration itself composes its three sources and dedupes: socket-dir
// pid names, run-row lead pids, and broker peer pids.
func TestDefaultCandidatePIDsComposesSources(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("MOAI_HOME", filepath.Join(root, ".moai"))

	// Socket source: fixture directory with pid-named sockets (and noise).
	sockDir := t.TempDir()
	origDir := socketScanDir
	socketScanDir = sockDir
	t.Cleanup(func() { socketScanDir = origDir })
	for _, name := range []string{"777001.sock", "777002.sock", "not-a-pid.sock", "sub.sock"} {
		if err := os.WriteFile(filepath.Join(sockDir, name), []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(sockDir, "777003.dir"), 0o700); err != nil {
		t.Fatal(err)
	}

	// Run-row source: a recorded run's lead pid.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RecordRun(ctx, homestate.FactoryRun{RunID: "runsrc01", Backend: "glm", LeadPID: 777004, LeadProcessStart: "1700000000.000001"}); err != nil {
		t.Fatalf("record fixture run: %v", err)
	}

	pids, err := defaultCandidatePIDs(ctx, root)
	if err != nil {
		t.Fatalf("defaultCandidatePIDs: %v", err)
	}
	got := map[int]bool{}
	for _, pid := range pids {
		got[pid] = true
	}
	for _, want := range []int{777001, 777002, 777004} {
		if !got[want] {
			t.Errorf("candidate pids missing %d (got %v)", want, pids)
		}
	}
	if got[777003] {
		t.Errorf("directory entry 777003 leaked into candidates (got %v)", pids)
	}
	for _, pid := range pids {
		if count := freq(pids, pid); count > 1 {
			t.Errorf("pid %d appears %d times; enumeration must dedupe", pid, count)
		}
	}
}

func freq(pids []int, pid int) int {
	n := 0
	for _, p := range pids {
		if p == pid {
			n++
		}
	}
	return n
}
