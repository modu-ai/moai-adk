package hook

// stale_run_gate_test.go — SPEC-STALE-RUN-LABEL-001: the run-state-gated
// prescription. A legacy launch label alone no longer decides the hook's
// answer: the named factory run's measured state does (REQ-SRL-001..003).
// Active → the stale-run prescription once per session identity (REQ-SRL-002);
// measured not-active → the one-time unbind notice (REQ-SRL-005/006);
// measurement failure → the degraded answer, never a prescription
// (REQ-SRL-003); and the inbound Claim path stays env-label-blind
// (REQ-SRL-007, the worker-70 separation).
//
// Every test runs on a t.TempDir() factory DB and seeds the MOAI_ variables it
// reads (m3ScrubEnv + explicit Setenv), so a lane-stamped environment cannot
// decide an assertion.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// srlGateEnv seeds the factory launch environment the prescription gate reads.
// label must be legacy vocabulary ("worker-<n>", "agent-<n>") for the gate
// branch to fire.
func srlGateEnv(t *testing.T, run, label string) {
	t.Helper()
	m3ScrubEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, run)
	t.Setenv(config.EnvMoaiFactoryWorkers, "4")
	t.Setenv(config.EnvMoaiFactoryWorker, label)
	t.Setenv(config.EnvMoaiKanbanBackend, "claude")
}

// recordFactoryRunWithStatus seeds a runs row with an explicit status, so a
// test can pin the measured state the gate must read (the shared accessor's
// contract: homestate.FactoryDBPath, runs.status).
func recordFactoryRunWithStatus(t *testing.T, root, run, status string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close factory state: %v", err)
		}
	}()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, LeadSessionID: "lead", Backend: "test", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE runs SET status=? WHERE run_id=?`, status, run); err != nil {
		t.Fatal(err)
	}
}

func TestStaleRunNoticeSilentWhenRunRetired(t *testing.T) { // AC-SRL-001
	root := t.TempDir()
	run := "srl-retired"
	recordFactoryRunWithStatus(t, root, run, "retired")
	srlGateEnv(t, run, "worker-69")

	in := &HookInput{SessionID: "srl-001-session", ProjectDir: root}
	if notice := registerFactoryHookPeer(context.Background(), in, factoryPeerBindUserPrompt); strings.Contains(notice, "runs --retire") {
		t.Fatalf("peer path prescribed a retire for a measured-retired run: %q", notice)
	}
	if notice := staleRunNoticeFor(root, "srl-001-session", "en"); strings.Contains(notice, "runs --retire") {
		t.Fatalf("staleRunNoticeFor prescribed a retire for a measured-retired run: %q", notice)
	}
}

func TestStaleRunNoticeFiresWhenRunActive(t *testing.T) { // AC-SRL-002 positive control
	root := t.TempDir()
	run := "srl-active"
	recordActiveFactoryRun(t, root, run)
	srlGateEnv(t, run, "worker-69")

	notice := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "srl-002-session", ProjectDir: root}, factoryPeerBindUserPrompt)
	for _, want := range []string{"stale run:", "worker-69", "runs --retire " + run} {
		if !strings.Contains(notice, want) {
			t.Fatalf("active-run prescription %q missing %q", notice, want)
		}
	}
}

func TestStaleRunNoticeOncePerSession(t *testing.T) { // AC-SRL-003
	root := t.TempDir()
	run := "srl-once"
	recordActiveFactoryRun(t, root, run)
	srlGateEnv(t, run, "worker-69")
	in := &HookInput{SessionID: "srl-003-session", ProjectDir: root}

	first := registerFactoryHookPeer(context.Background(), in, factoryPeerBindUserPrompt)
	if !strings.Contains(first, "runs --retire") {
		t.Fatalf("first turn lost the prescription: %q", first)
	}
	second := registerFactoryHookPeer(context.Background(), in, factoryPeerBindUserPrompt)
	if second != "" {
		t.Fatalf("second turn repeated the prescription: %q", second)
	}
	// The dedup carrier is shared across ALL prescription surfaces of the
	// session identity (plan M1.2): the SessionStart bootstrap must stay
	// silent too, so startup + first prompt cannot both emit.
	if notice := factoryBootstrapNotice(root, "srl-003-session", "en"); notice != "" {
		t.Fatalf("bootstrap surface repeated the prescription: %q", notice)
	}
}

func TestUnbindNoticeThenSilence(t *testing.T) { // AC-SRL-005 (a)
	root := t.TempDir()
	run := "srl-unbind"
	recordFactoryRunWithStatus(t, root, run, "retired")
	srlGateEnv(t, run, "worker-69")
	in := &HookInput{SessionID: "srl-004-session", ProjectDir: root}

	first := registerFactoryHookPeer(context.Background(), in, factoryPeerBindUserPrompt)
	if first == "" {
		t.Fatal("first turn silent, want the one-time unbind notice")
	}
	if strings.Contains(first, "runs --retire") {
		t.Fatalf("unbind notice carries the retire prescription: %q", first)
	}
	if !strings.Contains(first, "worker-69") {
		t.Fatalf("unbind notice does not name the orphan label: %q", first)
	}
	if !strings.Contains(first, "retired") {
		t.Fatalf("unbind notice does not name the measured run state: %q", first)
	}
	if again := registerFactoryHookPeer(context.Background(), in, factoryPeerBindUserPrompt); again != "" {
		t.Fatalf("second turn repeated the unbind notice: %q", again)
	}
	if again := registerFactorySessionStartPeer(context.Background(), in); again != "" {
		t.Fatalf("SessionStart surface repeated the unbind notice: %q", again)
	}
}

func TestUnbindNoticeRebindLinePresence(t *testing.T) { // AC-SRL-005 (b)
	root := t.TempDir()
	dead := "srl-dead"
	recordFactoryRunWithStatus(t, root, dead, "retired")
	recordActiveFactoryRun(t, root, "srl-live")
	srlGateEnv(t, dead, "worker-69")

	notice := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "srl-005a-session", ProjectDir: root}, factoryPeerBindUserPrompt)
	if strings.Contains(notice, "runs --retire") {
		t.Fatalf("unbind path prescribed a retire: %q", notice)
	}
	if !strings.Contains(notice, "moai cc -f lane-") {
		t.Fatalf("unbind notice omits the re-bind entry although an active run exists: %q", notice)
	}

	alone := t.TempDir()
	recordFactoryRunWithStatus(t, alone, dead, "retired")
	srlGateEnv(t, dead, "worker-69")
	solo := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "srl-005b-session", ProjectDir: alone}, factoryPeerBindUserPrompt)
	if strings.Contains(solo, "moai cc -f lane-") {
		t.Fatalf("unbind notice names a re-bind with no active run in the root: %q", solo)
	}
}

func TestPrescriptionGateUnavailableFailsOpen(t *testing.T) { // AC-SRL-008
	root := t.TempDir()
	run := "srl-unavailable"
	dir := filepath.Join(root, ".moai", "factory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "factory.db"), []byte("definitely not a sqlite database"), 0o600); err != nil {
		t.Fatal(err)
	}
	srlGateEnv(t, run, "worker-69")

	notice := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "srl-006-session", ProjectDir: root}, factoryPeerBindUserPrompt)
	if strings.Contains(notice, "runs --retire") {
		t.Fatalf("unmeasurable run state still prescribed a retire: %q", notice)
	}
	if !strings.HasPrefix(notice, "factory messaging degraded:") {
		t.Fatalf("unmeasurable run state must degrade to the degraded answer, got %q", notice)
	}
}

func TestClearSourceDeadRunEnvYieldsUnbound(t *testing.T) { // AC-SRL-004
	root := t.TempDir()
	run := "srl-clear"
	recordFactoryRunWithStatus(t, root, run, "retired")
	srlGateEnv(t, run, "worker-69")

	in := &HookInput{SessionID: "srl-007-session", Source: "clear", ProjectDir: root}
	notice := registerFactorySessionStartPeer(context.Background(), in)
	if strings.Contains(notice, "runs --retire") {
		t.Fatalf("the /clear boundary prescribed a retire for a dead-run label: %q", notice)
	}
	if strings.Contains(notice, "factory messaging bound") {
		t.Fatalf("the /clear boundary bound a peer under a dead-run label: %q", notice)
	}
	// No effective dead-run binding is carried into the fresh session: the
	// run's broker was never opened on this path, so no broker can exist.
	if _, err := os.Stat(filepath.Join(root, ".moai", "factory", "messages", run, "broker.db")); !os.IsNotExist(err) {
		t.Errorf("a broker was created for a dead-run session at the /clear boundary")
	}
}

func TestInboundClaimIndependentOfEnvLabel(t *testing.T) { // AC-SRL-006
	root := t.TempDir()
	run := "srl-claim"
	recordFactoryRunWithStatus(t, root, run, "retired")
	owner, start := factoryHookOwnerIdentity(t)
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", s)
	from := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "claude", Role: "leader", Slot: "leader", SessionUUID: "srl-sender", Generation: 1, PID: owner, ProcessStart: start}
	p := from
	p.Role, p.Slot, p.SessionUUID = "lane", "lane-1", "srl-receiver"
	if from, err = s.RegisterPeer(context.Background(), from); err != nil {
		t.Fatal(err)
	}
	if p, err = s.RegisterPeer(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Send(context.Background(), factorymsg.SendRequest{From: from, To: p, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "srl-claim-1", TaskRef: "t1373", CorrelationID: "c-1", TTL: time.Minute, Payload: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	// A legacy label for the measured-retired run rides the environment; the
	// Claim path must not read it (worker-70 separation, REQ-SRL-007).
	srlGateEnv(t, run, "worker-70")
	msg, _, state := factoryHookBatch(context.Background(), &HookInput{SessionID: p.SessionUUID, ProjectDir: root}, EventUserPromptSubmit)
	if !strings.Contains(msg, "Factory inbox run="+run) {
		t.Fatalf("claim delivery altered under a stale env label: msg=%q state=%s", msg, state)
	}
}
