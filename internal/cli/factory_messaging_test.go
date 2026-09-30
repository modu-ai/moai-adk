package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/sessionmsg"
)

// laneEnv pins the lane identity the messaging/fallback verbs read.
func laneEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
}

// seedRegistryPeer registers one messaging peer in the project-local broker
// registry the probe reads.
func seedRegistryPeer(t *testing.T, root, name string) {
	t.Helper()
	reg := sessionmsg.NewStore(filepath.Join(root, sessionmsg.DefaultStateRoot), nil)
	if _, err := reg.Register(sessionmsg.KindClaude, name, "test peer"); err != nil {
		t.Fatalf("register peer: %v", err)
	}
}

// AC-FLA-001: with nothing registered the probe reports channel-unavailable.
func TestFactoryMessagingProbeReportsChannelUnavailableWithoutRegistry(t *testing.T) {
	t.Chdir(t.TempDir())
	out, _, err := runFactory(t, "messaging", "probe")
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	if !strings.Contains(out, "channel-unavailable") {
		t.Fatalf("probe output = %q, want it to report channel-unavailable", out)
	}
}

// AC-FLA-001: a registered lead peer flips the verdict to available.
func TestFactoryMessagingProbeReportsAvailableWithRegisteredLead(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	seedRegistryPeer(t, root, "team-lead")
	out, _, err := runFactory(t, "messaging", "probe", "--leader", "team-lead")
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	if !strings.Contains(out, "available") || strings.Contains(out, "channel-unavailable") {
		t.Fatalf("probe output = %q, want available verdict", out)
	}
}

// The --json surface carries the verdict machine-readably.
func TestFactoryMessagingProbeJSONCarriesVerdict(t *testing.T) {
	t.Chdir(t.TempDir())
	out, _, err := runFactory(t, "messaging", "probe", "--json")
	if err != nil {
		t.Fatalf("probe --json: %v", err)
	}
	var got factorylane.Availability
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("parse probe json %q: %v", out, err)
	}
	if got.Verdict != factorylane.VerdictUnavailable {
		t.Fatalf("json verdict = %q, want %q", got.Verdict, factorylane.VerdictUnavailable)
	}
}

// AC-FLA-002: the directed request lands as an observation, and the ack marks
// it — the lane's own request lifecycle closes mechanically.
func TestFactoryMessagingRequestAndAckLifecycle(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	laneEnv(t)
	if _, _, err := runFactory(t, "messaging", "request"); err != nil {
		t.Fatalf("request: %v", err)
	}
	store := factorylane.NewStore(root, nil)
	obs, err := store.Observations("lane-1")
	if err != nil {
		t.Fatalf("observations: %v", err)
	}
	if len(obs) != 1 || obs[0].AckedAt != nil {
		t.Fatalf("observations = %+v, want one unacked request", obs)
	}
	if _, _, err := runFactory(t, "messaging", "ack"); err != nil {
		t.Fatalf("ack: %v", err)
	}
	obs, err = store.Observations("lane-1")
	if err != nil {
		t.Fatalf("observations after ack: %v", err)
	}
	if len(obs) != 1 || obs[0].AckedAt == nil {
		t.Fatalf("observations after ack = %+v, want the request acked", obs)
	}
}

// The directed-request verbs are lane-scoped: no lane label, no request.
func TestFactoryMessagingRequestRequiresLaneLabel(t *testing.T) {
	t.Chdir(t.TempDir())
	_, errOut, err := runFactory(t, "messaging", "request")
	if err == nil {
		t.Fatal("request without a lane label must be refused")
	}
	if !strings.Contains(errOut, config.EnvMoaiFactoryWorker) {
		t.Fatalf("refusal %q must name the lane-identity variable %s", errOut, config.EnvMoaiFactoryWorker)
	}
}

// AC-FLA-003: declare → restore → declare produces exactly the three events
// of two counted switches, and a declare while active is refused.
func TestFactoryFallbackDeclareRestoreLifecycle(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	laneEnv(t)
	store := factorylane.NewStore(root, nil)

	if _, _, err := runFactory(t, "fallback", "declare", "--trigger", "channel-unavailable", "--card", "t1"); err != nil {
		t.Fatalf("declare: %v", err)
	}
	if _, _, err := runFactory(t, "fallback", "declare", "--trigger", string(factorylane.TriggerNoResponse), "--card", "t1"); err == nil {
		t.Fatal("second declare while active must be refused")
	}
	events, err := store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("transitions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count after refused declare = %d, want exactly 1", len(events))
	}
	if _, _, err := runFactory(t, "fallback", "restore"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if _, _, err := runFactory(t, "fallback", "declare", "--trigger", string(factorylane.TriggerNoResponse), "--card", "t2"); err != nil {
		t.Fatalf("declare 2: %v", err)
	}
	events, err = store.Transitions("lane-1")
	if err != nil {
		t.Fatalf("transitions: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3 (two switches + one restore)", len(events))
	}
}

// The fallback query surface answers count-by-lane (AC-FLA-003's check).
func TestFactoryFallbackQueryPrintsCountByLane(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	laneEnv(t)
	if _, _, err := runFactory(t, "fallback", "declare", "--trigger", "channel-unavailable", "--card", "t1"); err != nil {
		t.Fatalf("declare: %v", err)
	}
	out, _, err := runFactory(t, "fallback")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !strings.Contains(out, "lane-1") {
		t.Fatalf("query output = %q, want it to name the lane", out)
	}
	if !strings.Contains(out, "channel-unavailable") {
		t.Fatalf("query output = %q, want it to carry the trigger kind", out)
	}
}

// An unknown trigger kind is refused at the CLI edge before any write.
func TestFactoryFallbackDeclareRefusesUnknownTrigger(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	laneEnv(t)
	if _, _, err := runFactory(t, "fallback", "declare", "--trigger", "banana"); err == nil {
		t.Fatal("unknown trigger must be refused")
	}
	if _, err := os.Stat(filepath.Join(root, factorylane.DefaultStateRoot, "transitions", "lane-1")); !os.IsNotExist(err) {
		t.Fatalf("invalid trigger wrote state: %v", err)
	}
}
