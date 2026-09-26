package cli

import (
	"bytes"
	"strings"
	"testing"
)

// seedLiveFactoryClaims writes claims the stubbed probe reports alive.
func seedLiveFactoryClaims(t *testing.T, root string, labels ...string) {
	t.Helper()
	reg := map[string]factoryWorkerEntry{}
	for i, label := range labels {
		reg[label] = factoryWorkerEntry{PID: 21000 + i}
	}
	if err := saveFactoryRegistry(factoryRegistryPath(root), reg); err != nil {
		t.Fatalf("seed registry: %v", err)
	}
	probe := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid >= 21000 && pid < 21000+len(labels) }
	t.Cleanup(func() { factoryProcessAlive = probe })
}

// TestResolveFactoryWorkerNameExplicitLegacyCollisionIsAnError: `-f
// worker-3` (or `--name lane-3`) against a live legacy row fails with a
// message naming the row and saying what to do.
func TestResolveFactoryWorkerNameExplicitLegacyCollisionIsAnError(t *testing.T) {
	root := t.TempDir()
	seedLiveFactoryClaims(t, root, "worker-3")

	var notes bytes.Buffer
	got, err := resolveFactoryWorkerName(root, "agent-3", false, &notes)
	if err == nil {
		t.Fatalf("explicit agent-3 over live worker-3 launched as %q; want an error naming worker-3", got)
	}
	t.Logf("operator-facing error: %v", err)
	for _, want := range []string{"agent-3", "legacy label worker-3", "-f agent"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q lacks %q", err, want)
		}
	}
}

// TestResolveFactoryWorkerNameAutoNamesSkippedLegacyRows: `-f worker` whose
// number lands past live legacy rows says so on stderr, by label.
func TestResolveFactoryWorkerNameAutoNamesSkippedLegacyRows(t *testing.T) {
	root := t.TempDir()
	seedLiveFactoryClaims(t, root, "agent-1", "lane-2")

	var notes bytes.Buffer
	got, err := resolveFactoryWorkerName(root, "agent-3", true, &notes)
	if err != nil || got != "agent-3" {
		t.Fatalf("auto agent-3 = (%q, %v), want agent-3", got, err)
	}
	out := notes.String()
	t.Logf("operator-facing stderr: %s", out)
	for _, want := range []string{"lane-2", "agent-3", "share"} {
		if !strings.Contains(out, want) {
			t.Errorf("stderr %q lacks %q (the skipped legacy row must be named)", out, want)
		}
	}
}

// TestParseLauncherEntryMarksAutoAssignedNumbers: only the `-f worker` role
// token desugars into an auto-assigned number.
func TestParseLauncherEntryMarksAutoAssignedNumbers(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	for args, want := range map[string]bool{"-f agent": true, "-f agent-2": false} {
		p, err := parseLauncherEntry(strings.Fields(args))
		if err != nil {
			t.Fatalf("parseLauncherEntry(%s): %v", args, err)
		}
		if p.FactoryAutoNumber != want {
			t.Errorf("parseLauncherEntry(%s).FactoryAutoNumber = %v, want %v", args, p.FactoryAutoNumber, want)
		}
	}
}
