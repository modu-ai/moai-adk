package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Card t1444 (card t1444 · AMBIGUOUS_FACTORY multi-run join): two owner-live
// runs keep the fail-closed refusal, and the operator's confirmed repair
// direction adds the missing selection paths on that branch — leader-name
// join (--leader), the codex --factory-run acceptance its own usage line
// already advertises, a refusal message that carries the resolution commands,
// and leader-name distinguishability so the name selection always resolves.

// seedLiveRun stamps one owner-live run row whose owner is THIS test process:
// it is alive for the test's duration, so reconcile classifies the row
// owner-live — the incident shape.
func seedLiveRun(t *testing.T, root, runID string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory db: %v", err)
	}
	defer func() { _ = db.Close() }()
	fp := homestate.CurrentProcessFingerprint()
	if fp == "" {
		t.Fatal("this test process has no live fingerprint to seed with")
	}
	err = db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "claude", ManifestJSON: "{}",
		LeadPID: os.Getpid(), LeadProcessStart: fp,
		LaneCapacity: homestate.LaneCapacityDerived,
	})
	if err != nil {
		t.Fatalf("seed live run %s: %v", runID, err)
	}
}

// holdLeaderName writes a live leader-name claim into the leader registry
// through the production writer (the registry is a SQLite workers table
// behind a legacy .json name — a hand-written file does not parse): the pid
// is this test process's, so the claim reads as held by a live session for
// the test's duration.
func holdLeaderName(t *testing.T, root, name string) {
	t.Helper()
	path := leaderRegistryPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveFactoryRegistry(path, map[string]factoryLaneEntry{
		name: {PID: os.Getpid(), RegisteredAt: "2026-10-02T00:00:00Z"},
	}); err != nil {
		t.Fatalf("hold leader name %s: %v", name, err)
	}
}

// ① operator test: with two owner-live runs, a --leader join selects THAT
// leader's run instead of refusing. The discovery seam is the verified-leader
// stub; the gate must ASK discovery on this branch (it never did before) and
// join the verified leader's own run identity.
func TestAmbiguousLeaderJoinSelectsNamedLeaderRun(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")
	asked := stageDiscoveredLeaders(t, []discovery.VerifiedLeader{verifiedTestLeader("ambigruna")})

	restore, err := enterFactoryLaneRun(root, "", "leader", nil)
	if err != nil {
		t.Fatalf("leader-named join among two live runs = %v, want the named leader's run", err)
	}
	defer restore()
	if len(*asked) != 1 || (*asked)[0] != "leader" {
		t.Errorf("discovery asked = %v, want exactly [leader]", *asked)
	}
	if runID := os.Getenv(config.EnvFactoryRunID); runID != "ambigruna" {
		t.Errorf("%s = %q, want the named leader's run ambigruna", config.EnvFactoryRunID, runID)
	}
}

// ① fail-closed boundary: --leader names a target nobody verified — the
// original AMBIGUOUS refusal stands (now with ③'s guidance).
func TestAmbiguousLeaderJoinUnknownLeaderStillRefuses(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")
	stageDiscoveredLeaders(t, nil)

	_, err := enterFactoryLaneRun(root, "", "leader-nowhere", nil)
	if err == nil {
		t.Fatal("leader-named join with zero verified leaders succeeded, want the ambiguity refusal")
	}
	if !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Errorf("err = %v, want the AMBIGUOUS_FACTORY refusal to stand", err)
	}
}

// ① fail-closed boundary: two verified leaders answer one name — the
// AMBIGUOUS_FACTORY_LEADER refusal (selection among live leaders stays
// forbidden, REQ-005).
func TestAmbiguousLeaderJoinTwoLeadersFailsClosed(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")
	stageDiscoveredLeaders(t, []discovery.VerifiedLeader{
		verifiedTestLeader("ambigruna"), verifiedTestLeader("ambigrunb"),
	})

	_, err := enterFactoryLaneRun(root, "", "leader", nil)
	if err == nil || !strings.Contains(err.Error(), ambiguousFactoryLeaderSentinel) {
		t.Fatalf("two verified leaders err = %v, want %s", err, ambiguousFactoryLeaderSentinel)
	}
}

// ③ operator test: the unnamed ambiguous refusal carries the resolution —
// the surviving run ids AND the commands that resolve them.
func TestAmbiguousRefusalCarriesResolutionGuidance(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")

	_, err := enterFactoryLaneRun(root, "", "", nil)
	if err == nil {
		t.Fatal("unnamed join among two live runs succeeded, want the ambiguity refusal")
	}
	msg := err.Error()
	for _, want := range []string{
		"AMBIGUOUS_FACTORY", "ambigruna", "ambigrunb",
		"--factory-run", "--leader", "--retire",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal message %q misses %q", msg, want)
		}
	}
}

// ② the codex head classifier stops refusing --factory-run on the lane
// entry: the token travels to the parser, which owns its validation.
func TestCodexClassifyAllowsFactoryRunOnLaneEntry(t *testing.T) {
	entry, diag := codexFactoryEntryClassify([]string{"-l", "--factory-run", "ambigruna"})
	if diag != "" {
		t.Fatalf("classify diag = %q, want empty (parser owns --factory-run)", diag)
	}
	if entry != codexFactoryEntryLane {
		t.Errorf("classify entry = %v, want codexFactoryEntryLane", entry)
	}
}

// ② the refusal MOVED, not vanished: --factory-run without a lane entry is
// still refused, now by the parser with its precise message.
func TestCodexFactoryRunWithoutLaneStillRefused(t *testing.T) {
	entry, diag := codexFactoryEntryClassify([]string{"--factory-run", "ambigruna"})
	if diag != "" {
		t.Fatalf("classify diag = %q, want empty (parser owns --factory-run)", diag)
	}
	_ = entry
	_, _, err := parseCodexFactoryEntry([]string{"--factory-run", "ambigruna"})
	if err == nil || !strings.Contains(err.Error(), "--factory-run requires -l/--lane") {
		t.Fatalf("parse err = %v, want the requires--l refusal", err)
	}
}

// ② pins the shared-gate semantic the codex plumbing relies on: an explicit
// run id joins the NAMED run among two live ones (works today on cc/glm;
// guards the contract the codex twin inherits).
func TestExplicitFactoryRunJoinsNamedRunAmongTwoLive(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")

	restore, err := enterFactoryLaneRun(root, "ambigrunb", "", nil)
	if err != nil {
		t.Fatalf("explicit join among two live runs = %v, want ambigrunb", err)
	}
	defer restore()
	if runID := os.Getenv(config.EnvFactoryRunID); runID != "ambigrunb" {
		t.Errorf("%s = %q, want the explicitly named ambigrunb", config.EnvFactoryRunID, runID)
	}
}

// ② the codex relaunch loop's join carries the parsed entry's run id: the
// named run joins among two live ones instead of the ambiguity refusal.
// (Lands with the fix — the seam enterCodexRelaunchJoin is the plumbing this
// card unblocked.)
func TestCodexRelaunchJoinCarriesFactoryRunID(t *testing.T) {
	root := discoveryTestRoot(t)
	clearFactoryTestEnv(t)
	seedLiveRun(t, root, "ambigruna")
	seedLiveRun(t, root, "ambigrunb")

	_, entry, perr := parseCodexFactoryEntry([]string{"-l", "--factory-run", "ambigruna"})
	if perr != nil {
		t.Fatalf("parse -l --factory-run: %v", perr)
	}
	restore, err := enterCodexRelaunchJoin(root, entry, nil)
	if err != nil {
		t.Fatalf("codex relaunch join with --factory-run among two live runs = %v, want ambigruna", err)
	}
	defer restore()
	if runID := os.Getenv(config.EnvFactoryRunID); runID != "ambigruna" {
		t.Errorf("%s = %q, want the explicitly named ambigruna", config.EnvFactoryRunID, runID)
	}
}

// ④ operator test: an operator-supplied leader name that a LIVE leader
// already holds is bumped to a distinguishable name, so ①'s name selection
// always has an unambiguous target. The default-name path already bumps;
// the operator-supplied path is the gap.
func TestOperatorLeaderNameCollisionBumps(t *testing.T) {
	root := discoveryTestRoot(t)
	holdLeaderName(t, root, "leader")

	args := []string{"-f", "--name", "leader"}
	var notes bytes.Buffer
	finalArgs, name := appendLeaderName(args, root, &notes)
	if name == "leader" {
		t.Fatalf("operator-supplied colliding leader name kept as %q, want a bumped name", name)
	}
	if !strings.Contains(strings.Join(finalArgs, " "), name) {
		t.Errorf("argv %v does not carry the resolved name %q", finalArgs, name)
	}
	if !strings.Contains(notes.String(), "leader-1") {
		t.Errorf("bump note = %q, want it to name the bumped leader-1", notes.String())
	}
}
