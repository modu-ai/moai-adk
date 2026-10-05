package hygiene

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestProbePidAliveProduction — the real (non-seam) pid probe: a live
// process reads unmeasured (no recorded fingerprint to compare), a reaped
// process reads negative, and a zero pid reads unmeasured.
func TestProbePidAliveProduction(t *testing.T) {
	if s := probePidAlive(0); s != SignalUnmeasured {
		t.Fatalf("zero pid = %s, want unmeasured", s)
	}
	if s := probePidAlive(os.Getpid()); s != SignalUnmeasured {
		t.Fatalf("live pid = %s, want unmeasured (no recorded fingerprint)", s)
	}
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn probe child: %v", err)
	}
	deadPID := cmd.Process.Pid
	_, _ = cmd.Process.Wait()
	if s := probePidAlive(deadPID); s != SignalNegative {
		t.Fatalf("reaped pid = %s, want negative", s)
	}
}

// TestReadRegistryEntriesProduction — the plain registry decode: an array
// reads back, and an unparseable body errors (the caller then treats the
// registry as absent — fail-closed).
func TestReadRegistryEntriesProduction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "active-sessions.json")
	entries := []RegistryEntry{{SessionID: keyDead, PID: 4242}}
	blob, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, blob, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readRegistryEntries(path)
	if err != nil || len(got) != 1 || got[0].SessionID != keyDead {
		t.Fatalf("read = %v, %v; want one entry", got, err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if _, err := readRegistryEntries(path); err == nil {
		t.Fatalf("unparseable registry read without error")
	}
}

// TestScanTranscriptsProduction — the real (non-seam) transcript scan over
// fixture profile roots: fresh transcript affirmative, stale negative,
// absent under resolvable roots unmeasured, and an unresolvable root
// unmeasured.
func TestScanTranscriptsProduction(t *testing.T) {
	root := t.TempDir()
	projects := filepath.Join(root, "projects")
	projDir := filepath.Join(projects, "some-project")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	l := &Liveness{
		TranscriptWindow: testActivityWindow,
		now:              func() time.Time { return fixtureNow },
	}

	// Absent under a resolvable root: unmeasured, never negative.
	l.TranscriptRoots = []string{projects}
	if s := l.scanTranscripts(keyDead); s != SignalUnmeasured {
		t.Fatalf("absent transcript = %s, want unmeasured", s)
	}

	// Fresh transcript: affirmative (two-level shape: projects/<dir>/<key>).
	freshPath := filepath.Join(projDir, keyDead+".jsonl")
	if err := os.WriteFile(freshPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write fresh: %v", err)
	}
	if s := l.scanTranscripts(keyDead); s != SignalAffirmative {
		t.Fatalf("fresh transcript = %s, want affirmative", s)
	}

	// Stale transcript: negative.
	stale := fixtureNow.Add(-testActivityWindow).Add(-time.Hour)
	if err := os.Chtimes(freshPath, stale, stale); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	if s := l.scanTranscripts(keyDead); s != SignalNegative {
		t.Fatalf("stale transcript = %s, want negative", s)
	}

	// A top-level transcript (root/<key>.jsonl) is found too.
	l2 := &Liveness{TranscriptWindow: testActivityWindow, TranscriptRoots: []string{projects},
		now: func() time.Time { return fixtureNow }}
	if err := os.WriteFile(filepath.Join(projects, keyGoal+".jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write top-level: %v", err)
	}
	if s := l2.scanTranscripts(keyGoal); s != SignalAffirmative {
		t.Fatalf("top-level fresh transcript = %s, want affirmative", s)
	}

	// Unresolvable root: unmeasured.
	l3 := &Liveness{TranscriptWindow: testActivityWindow,
		TranscriptRoots: []string{filepath.Join(root, "does-not-exist")},
		now:             func() time.Time { return fixtureNow }}
	if s := l3.scanTranscripts(keyDead); s != SignalUnmeasured {
		t.Fatalf("unresolvable root = %s, want unmeasured", s)
	}

	// No roots at all: unmeasured.
	l4 := &Liveness{TranscriptWindow: testActivityWindow, TranscriptRoots: nil,
		now: func() time.Time { return fixtureNow }}
	if s := l4.scanTranscripts(keyDead); s != SignalUnmeasured {
		t.Fatalf("no roots = %s, want unmeasured", s)
	}
}

// TestSettingsProductionSurfaces — DefaultSettings mirrors the compiled
// config constants, ConfigError renders, and the home/config-root
// resolvers return without error on a developer machine.
func TestSettingsProductionSurfaces(t *testing.T) {
	s := DefaultSettings()
	if err := s.Validate(); err != nil {
		t.Fatalf("compiled defaults do not validate: %v", err)
	}
	if s.AuditLogMaxBytes <= 0 || s.MinAgeDays <= 0 {
		t.Fatalf("compiled defaults carry non-positive thresholds: %+v", s)
	}
	ce := &ConfigError{Key: "k", Value: "v", Rule: "r"}
	if ce.Error() == "" || !strings.Contains(ce.Error(), "config-invalid") {
		t.Fatalf("ConfigError rendering incomplete: %q", ce.Error())
	}
	_ = homeDir()
	_ = DefaultTranscriptRoots()
	if got := OutcomeRotated.String(); got != "rotated" {
		t.Fatalf("outcome render = %q", got)
	}
}

// TestRegisterTestRootExported — the exported wiring-test vouching entry
// registers the same registry the guard consults.
func TestRegisterTestRootExported(t *testing.T) {
	root := t.TempDir()
	RegisterTestRoot(root)
	if err := validateRoot(root); err != nil {
		t.Fatalf("exported registration not honored: %v", err)
	}
}

// TestLoadSettingsFromFixture — the yaml override surface end-to-end
// inside the owning package: defaults when absent, pointer overrides
// applied (including an explicit zero, which validation then refuses).
func TestLoadSettingsFromFixture(t *testing.T) {
	root := t.TempDir()
	got := LoadSettingsFrom(root)
	if err := got.Validate(); err != nil {
		t.Fatalf("absent config did not yield compiled defaults: %v", err)
	}
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yamlBody := "workflow:\n  hygiene:\n    mode: apply\n    min_age_days: 2\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got = LoadSettingsFrom(root)
	if got.Mode != "apply" || got.MinAgeDays != 2 {
		t.Fatalf("overrides not applied: %+v", got)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte("workflow:\n  hygiene:\n    min_age_days: 0\n"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	got = LoadSettingsFrom(root)
	if err := got.Validate(); err == nil {
		t.Fatalf("explicit zero did not reach validation")
	}
}
