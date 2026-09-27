package cli

import (
	"bytes"
	"strings"
	"testing"
)

// seedLiveFactoryClaims writes claims the stubbed probe reports alive.
func seedLiveFactoryClaims(t *testing.T, root string, labels ...string) {
	t.Helper()
	reg := map[string]factoryLaneEntry{}
	for i, label := range labels {
		reg[label] = factoryLaneEntry{PID: 21000 + i}
	}
	if err := saveFactoryRegistry(factoryRegistryPath(root), reg); err != nil {
		t.Fatalf("seed registry: %v", err)
	}
	probe := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid >= 21000 && pid < 21000+len(labels) }
	t.Cleanup(func() { factoryProcessAlive = probe })
}

// TestResolveFactoryWorkerNameRefusesLegacyRequest: a legacy label typed on
// the input path is refused with an error naming the canonical lane-<n>
// (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009; the dedicated rejection wording is
// M2).
func TestResolveFactoryWorkerNameRefusesLegacyRequest(t *testing.T) {
	for legacy, canonical := range map[string]string{"worker-3": "lane-3", "agent-2": "lane-2"} {
		var notes bytes.Buffer
		got, err := resolveFactoryLaneName(t.TempDir(), legacy, false, &notes)
		if err == nil {
			t.Fatalf("legacy %s launched as %q; want an error naming %s", legacy, got, canonical)
		}
		if !strings.Contains(err.Error(), canonical) {
			t.Errorf("error %q lacks the canonical form %s", err, canonical)
		}
	}
}

// TestResolveFactoryWorkerNameAutoIgnoresLegacyRows: an auto claim proceeds
// past a live legacy row (the legacy import shape carries no run id, so it
// belongs to no run — P3) and the row is never rewritten.
func TestResolveFactoryWorkerNameAutoIgnoresLegacyRows(t *testing.T) {
	root := t.TempDir()
	seedLiveFactoryClaims(t, root, "lane-1", "worker-2")

	var notes bytes.Buffer
	got, err := resolveFactoryLaneName(root, "", true, &notes)
	if err != nil || got != "lane-2" {
		t.Fatalf("auto claim = (%q, %v), want lane-2 (legacy worker-2 holds no number)", got, err)
	}
	reg := loadFactoryRegistry(factoryRegistryPath(root))
	if _, ok := reg["worker-2"]; !ok {
		t.Errorf("legacy row worker-2 must survive untouched, registry = %v", reg)
	}
	if _, ok := reg["lane-2"]; !ok {
		t.Errorf("new claim lane-2 must be recorded, registry = %v", reg)
	}
}

// TestParseLauncherEntryMarksAutoAssignedNumbers: only the `-f lane` role
// token desugars into an auto-assigned number; the legacy role tokens are
// refused before the merge (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-003).
func TestParseLauncherEntryMarksAutoAssignedNumbers(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for args, want := range map[string]bool{"-f lane": true, "-f lane-2": false} {
		p, err := parseLauncherEntry(strings.Fields(args))
		if err != nil {
			t.Fatalf("parseLauncherEntry(%s): %v", args, err)
		}
		if p.FactoryAutoNumber != want {
			t.Errorf("parseLauncherEntry(%s).FactoryAutoNumber = %v, want %v", args, p.FactoryAutoNumber, want)
		}
	}
	for _, legacy := range []string{"-f worker", "-f agent"} {
		if _, err := parseLauncherEntry(strings.Fields(legacy)); err == nil {
			t.Errorf("parseLauncherEntry(%s) = nil error, want the legacy-token refusal", legacy)
		}
	}
}
