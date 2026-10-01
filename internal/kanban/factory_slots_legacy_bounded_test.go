package kanban

// factory_slots_legacy_bounded_test.go — SPEC-CODEX-LANE-SLOTS-001 AC-004
// (REQ-009): the legacy-run refusal fires on the BOUNDED claim path exactly
// as on the unbounded path — a live legacy record stamped with a run id
// refuses the whole claim naming the legacy value, the run id, and the
// retire step, and a capacity-open record never grows into the legacy
// vocabulary.

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestClaimFactoryLaneWithinRefusesLiveLegacy(t *testing.T) {
	root := t.TempDir()
	if err := seedWorkerRowRun(t, root, "worker-1", 999999, "legacyrun1"); err != nil {
		t.Fatal(err)
	}
	alive := func(int) bool { return true }
	_, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "legacyrun1", 1, alive)
	var legacy *FactoryLegacyRunError
	if !errors.As(err, &legacy) {
		t.Fatalf("bounded claim on live legacy row: err = %v, want *FactoryLegacyRunError", err)
	}
	if legacy.Label != "worker-1" {
		t.Errorf("legacy error label = %q, want worker-1", legacy.Label)
	}
	if legacy.RunID != "legacyrun1" {
		t.Errorf("legacy error run id = %q, want legacyrun1", legacy.RunID)
	}
	msg := legacy.Error()
	for _, want := range []string{"worker-1", "legacyrun1", "moai factory runs --retire legacyrun1"} {
		if !strings.Contains(msg, want) {
			t.Errorf("legacy error message %q missing %q", msg, want)
		}
	}
	// Nothing was written: the registry still holds exactly the legacy row,
	// and no capacity-open growth slipped past the refusal.
	reg := LoadFactoryRegistry(FactoryRegistryPath(root))
	if len(reg) != 1 {
		t.Fatalf("registry holds %d rows after refused bounded claim, want 1 (legacy row untouched)", len(reg))
	}
	if _, ok := reg["worker-1"]; !ok {
		t.Errorf("legacy row worker-1 vanished after the bounded refusal")
	}
}
