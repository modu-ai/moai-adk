package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorylane"
)

// AC-FLA-007 at the CLI edge: with no classification input — the M2 default
// while t1332 has not landed — the pickup decision exits 0 under the
// fallback classification and carries the single-dispatch behavior marker.
func TestFactoryPickupPlanFallsBackWithoutMetadata(t *testing.T) {
	t.Chdir(t.TempDir())
	laneEnv(t)
	out, _, err := runFactory(t, "pickup", "plan", "--card", "t1")
	if err != nil {
		t.Fatalf("pickup plan: %v", err)
	}
	if !strings.Contains(out, "fallback") {
		t.Fatalf("output = %q, want the fallback classification named", out)
	}
	if !strings.Contains(out, "single-dispatch") {
		t.Errorf("output = %q, want the single-dispatch behavior marker", out)
	}
}

// AC-FLA-006 at the CLI edge: a sequential classification evaluates under
// the classified rules (here: free group, so allowed and exclusive).
func TestFactoryPickupPlanClassifiedSequentialFreeGroup(t *testing.T) {
	t.Chdir(t.TempDir())
	laneEnv(t)
	out, _, err := runFactory(t, "pickup", "plan", "--card", "t1", "--axis", "sequential", "--group", "alpha")
	if err != nil {
		t.Fatalf("pickup plan --axis sequential: %v", err)
	}
	if !strings.Contains(out, "sequential group alpha") {
		t.Fatalf("output = %q, want the sequential classification with its group", out)
	}
	if !strings.Contains(out, "allowed") {
		t.Errorf("output = %q, want the allowed decision", out)
	}
}

// AC-FLA-008 at the CLI edge: an unknown axis value is tolerated — exit 0,
// fallback classification, tolerated-unknown condition logged.
func TestFactoryPickupPlanUnknownAxisTolerated(t *testing.T) {
	t.Chdir(t.TempDir())
	laneEnv(t)
	out, _, err := runFactory(t, "pickup", "plan", "--card", "t1", "--axis", "banana")
	if err != nil {
		t.Fatalf("unknown axis must not error (exit 0): %v", err)
	}
	if !strings.Contains(out, "fallback") || !strings.Contains(out, "tolerated") {
		t.Fatalf("output = %q, want the fallback classification and the tolerated-unknown log", out)
	}
}

// The --json surface carries the decision machine-readably.
func TestFactoryPickupPlanJSONCarriesDecision(t *testing.T) {
	t.Chdir(t.TempDir())
	laneEnv(t)
	out, _, err := runFactory(t, "pickup", "plan", "--card", "t1", "--json")
	if err != nil {
		t.Fatalf("pickup plan --json: %v", err)
	}
	var got factorylane.PickupDecision
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("parse pickup json %q: %v", out, err)
	}
	if !got.Fallback || !got.Allowed || got.MultiPick {
		t.Fatalf("json decision = %+v, want the allowed fallback single-dispatch", got)
	}
	if got.Card != "t1" || got.Lane != "lane-1" {
		t.Errorf("json decision = %+v, want card t1 for lane-1", got)
	}
}

// The verb is lane-scoped: no lane identity, no decision.
func TestFactoryPickupPlanRequiresLaneLabel(t *testing.T) {
	t.Chdir(t.TempDir())
	_, errOut, err := runFactory(t, "pickup", "plan", "--card", "t1")
	if err == nil {
		t.Fatal("pickup plan without a lane identity must be refused")
	}
	if !strings.Contains(errOut, config.EnvMoaiFactoryWorker) {
		t.Fatalf("refusal %q must name the lane-identity variable %s", errOut, config.EnvMoaiFactoryWorker)
	}
}
