// Package cli — retention spawn gate wiring and child verb tests
// (SPEC-HARNESS-DETACHED-PRUNE-001). REQ-DP-007: no test spawns a real detached
// child — the child-entry tests drive the verb's RunE in-process, and the
// wiring tests inject recording fakes through the retentionSpawnImpl seam.
package cli

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness"
)

// stubRetentionSpawnNoop replaces retentionSpawnImpl with a no-op fake for the
// duration of the test and restores it via t.Cleanup (REQ-DP-007: no test
// spawns a real detached child — every handler-driving CLI test installs this
// stub so the gate's stale-stamp arm can never launch a real process). The
// calling test must not run parallel (package-var override discipline).
func stubRetentionSpawnNoop(t *testing.T) {
	t.Helper()
	orig := retentionSpawnImpl
	t.Cleanup(func() { retentionSpawnImpl = orig })
	retentionSpawnImpl = func(_ string, _ []string) error { return nil }
}

// retentionProbeEvent builds an Event with the given subject and timestamp,
// sharing one context hash so same-key fixtures aggregate into one pattern.
func retentionProbeEvent(subject string, at time.Time) harness.Event {
	return harness.Event{
		Timestamp:     at.UTC(),
		EventType:     harness.EventTypeAgentInvocation,
		Subject:       subject,
		ContextHash:   "ch-t1497",
		TierIncrement: 0,
		SchemaVersion: harness.LogSchemaVersion,
	}
}

// seedRetentionProbeLog writes the given events as JSONL at logPath (fresh file).
func seedRetentionProbeLog(t *testing.T, logPath string, events []harness.Event) {
	t.Helper()
	var b strings.Builder
	enc := json.NewEncoder(&b)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			t.Fatalf("seed event encode: %v", err)
		}
	}
	if err := os.WriteFile(logPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("seed usage log: %v", err)
	}
}

// writeSpawnStamp writes a prune stamp with the given wall-clock time next to
// the log path (".prune-state", the retention package's state-file suffix).
func writeSpawnStamp(t *testing.T, logPath string, writtenAt time.Time) {
	t.Helper()
	if err := os.WriteFile(logPath+".prune-state", []byte(writtenAt.UTC().Format(time.RFC3339Nano)), 0o644); err != nil {
		t.Fatalf("write stamp file: %v", err)
	}
}

// runRetentionPruneVerb drives the registered hidden verb's RunE in-process
// (REQ-DP-007: never a detached child) against the given paths.
func runRetentionPruneVerb(t *testing.T, logPath, archiveDir string) {
	t.Helper()
	var verb *cobra.Command
	for _, c := range hookCmd.Commands() {
		if c.Name() == "retention-prune" {
			verb = c
			break
		}
	}
	if verb == nil {
		t.Fatal("retention-prune verb is not registered on hookCmd")
	}
	for flag, val := range map[string]string{"log": logPath, "archive": archiveDir} {
		if err := verb.Flags().Set(flag, val); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
	}
	if err := verb.RunE(verb, nil); err != nil {
		t.Fatalf("retention-prune RunE returned error: %v", err)
	}
}

// readArchiveEventCount decompresses every .jsonl.gz member under archiveDir
// and returns the total event count found across the members.
func readArchiveEventCount(t *testing.T, archiveDir string) int {
	t.Helper()
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("read archive dir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".jsonl.gz") {
			continue
		}
		f, err := os.Open(filepath.Join(archiveDir, e.Name()))
		if err != nil {
			t.Fatalf("open archive member: %v", err)
		}
		gz, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			t.Fatalf("gzip reader: %v", err)
		}
		dec := json.NewDecoder(gz)
		for {
			var evt harness.Event
			if err := dec.Decode(&evt); err != nil {
				if err == io.EOF {
					break
				}
				_ = gz.Close()
				_ = f.Close()
				t.Fatalf("decode archive member %s: %v", e.Name(), err)
			}
			count++
		}
		_ = gz.Close()
		_ = f.Close()
	}
	return count
}

// TestHookRetentionPruneVerb verifies the hidden child verb is registered,
// hidden from help, and performs a prune against a temp log (REQ-DP-003).
func TestHookRetentionPruneVerb(t *testing.T) {
	var verb *cobra.Command
	for _, c := range hookCmd.Commands() {
		if c.Name() == "retention-prune" {
			verb = c
			break
		}
	}
	if verb == nil {
		t.Fatal("retention-prune verb is not registered")
	}
	if !verb.Hidden {
		t.Error("retention-prune must register Hidden so it never appears in help output")
	}

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Now().UTC()
	seedRetentionProbeLog(t, logPath, []harness.Event{
		retentionProbeEvent("verb-stale", now.AddDate(0, 0, -40)),
		retentionProbeEvent("verb-fresh", now),
	})

	runRetentionPruneVerb(t, logPath, archiveDir)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read pruned log: %v", err)
	}
	if strings.Contains(string(data), "verb-stale") {
		t.Error("the retention-expired event survived the verb run — the verb performed no prune")
	}
	if !strings.Contains(string(data), "verb-fresh") {
		t.Error("the fresh event was lost — the verb pruned beyond the retention window")
	}
}

// TestDetachedChildPrunes verifies the child entry's observable prune effects
// (AC-DP-003): the kept-line count shrinks to the fresh event and an archive
// member (<YYYY-MM>.jsonl.gz) carries the expired events.
func TestDetachedChildPrunes(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Now().UTC()
	seedRetentionProbeLog(t, logPath, []harness.Event{
		retentionProbeEvent("child-stale-1", now.AddDate(0, 0, -40)),
		retentionProbeEvent("child-stale-2", now.AddDate(0, 0, -35)),
		retentionProbeEvent("child-fresh", now),
	})

	runRetentionPruneVerb(t, logPath, archiveDir)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read kept log: %v", err)
	}
	kept := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			kept++
		}
	}
	if kept != 1 {
		t.Errorf("kept-line count: got=%d, want=1 (only the fresh event survives)", kept)
	}
	if got := readArchiveEventCount(t, archiveDir); got != 2 {
		t.Errorf("archived expired events: got=%d, want=2", got)
	}
}

// TestDetachedChildDoubleSpawnCollapses verifies the once-per-interval collapse
// (AC-DP-004): driving the child entry twice against one stamp file leaves the
// second invocation exit without a second rewrite — the fresh stamp the first
// child wrote suppresses the second (pruneLocked's stamp re-check contract).
func TestDetachedChildDoubleSpawnCollapses(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Now().UTC()
	seedRetentionProbeLog(t, logPath, []harness.Event{
		retentionProbeEvent("collapse-stale", now.AddDate(0, 0, -40)),
		retentionProbeEvent("collapse-fresh", now),
	})

	runRetentionPruneVerb(t, logPath, archiveDir)

	snapshot, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log after first child: %v", err)
	}

	runRetentionPruneVerb(t, logPath, archiveDir)

	after, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log after second child: %v", err)
	}
	if string(after) != string(snapshot) {
		t.Errorf("second child rewrote the log — the fresh stamp did not collapse the double spawn")
	}
}

// TestStopClassificationFiltersExpiredEvents verifies the classifier applies
// the retention window at the aggregation input (REQ-DP-009, AC-DP-008): with
// 4 retention-expired events + 1 recent sharing one pattern key, the
// recent-only pattern classifies `observation`, never `rule`, and the expired
// four contribute nothing to the aggregated count.
func TestStopClassificationFiltersExpiredEvents(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("mkdir harness dir: %v", err)
	}
	now := time.Now().UTC()
	events := make([]harness.Event, 0, 5)
	for i := range 4 {
		events = append(events, retentionProbeEvent("filter-probe", now.AddDate(0, 0, -60-(3-i))))
	}
	events = append(events, retentionProbeEvent("filter-probe", now))
	seedRetentionProbeLog(t, logPath, events)

	patternCount, promoCount, err := classifyHarnessPatterns(dir)
	if err != nil {
		t.Fatalf("classifyHarnessPatterns returned error: %v", err)
	}
	if patternCount != 1 || promoCount != 1 {
		t.Fatalf("pattern/promo counts: got=(%d,%d), want=(1,1)", patternCount, promoCount)
	}

	promoPath := filepath.Join(dir, ".moai", "harness", "learning-history", "tier-promotions.jsonl")
	data, err := os.ReadFile(promoPath)
	if err != nil {
		t.Fatalf("read tier-promotions.jsonl: %v", err)
	}
	var promo harness.Promotion
	if err := json.Unmarshal(data, &promo); err != nil {
		t.Fatalf("decode promotion: %v", err)
	}
	if promo.ToTier == "rule" || promo.ToTier == "heuristic" || promo.ToTier == "auto_update" {
		t.Errorf("promotion tier: got=%q — the expired four were aggregated into the classification (REQ-DP-009 violation)", promo.ToTier)
	}
	if promo.ToTier != "observation" {
		t.Errorf("promotion tier: got=%q, want=%q", promo.ToTier, "observation")
	}
	if promo.ObservationCount != 1 {
		t.Errorf("observation count: got=%d, want=1 — the expired four contributed to the count", promo.ObservationCount)
	}
}

// TestHarnessObserveGateWiring verifies the wrapper is on the observe path
// (AC-DP-006): with a stale stamp, exactly one spawn is attempted through the
// retentionSpawnImpl seam; with a fresh stamp, none. The fake is injected at
// the wrapper layer and restored via t.Cleanup; this test does not run parallel
// (package-var override discipline, REQ-DP-007).
func TestHarnessObserveGateWiring(t *testing.T) {
	t.Run("stale_stamp_spawns_once_through_the_seam", func(t *testing.T) {
		dir := t.TempDir()
		writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
		t.Chdir(dir)
		logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			t.Fatalf("mkdir harness dir: %v", err)
		}
		writeSpawnStamp(t, logPath, time.Now().Add(-2*time.Hour))
		// Resolve the root by cwd, never by an ambient CLAUDE_PROJECT_DIR the
		// lane environment may carry (the fixture stamp lives under dir).
		t.Setenv(config.EnvClaudeProjectDir, "")

		calls := 0
		orig := retentionSpawnImpl
		t.Cleanup(func() { retentionSpawnImpl = orig })
		retentionSpawnImpl = func(_ string, _ []string) error {
			calls++
			return nil
		}

		cmd := &cobra.Command{}
		withStdin(t, `{"toolName":"Bash","toolInput":{"command":"echo hi"}}`, func() {
			if err := runHarnessObserve(cmd, nil); err != nil {
				t.Fatalf("runHarnessObserve returned error: %v", err)
			}
		})
		if calls != 1 {
			t.Fatalf("spawn attempts through the seam: got=%d, want=1", calls)
		}
	})

	t.Run("fresh_stamp_spawns_nothing", func(t *testing.T) {
		dir := t.TempDir()
		writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
		t.Chdir(dir)
		logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			t.Fatalf("mkdir harness dir: %v", err)
		}
		writeSpawnStamp(t, logPath, time.Now().Add(-5*time.Minute))
		t.Setenv(config.EnvClaudeProjectDir, "")

		calls := 0
		orig := retentionSpawnImpl
		t.Cleanup(func() { retentionSpawnImpl = orig })
		retentionSpawnImpl = func(_ string, _ []string) error {
			calls++
			return nil
		}

		cmd := &cobra.Command{}
		withStdin(t, `{"toolName":"Bash","toolInput":{"command":"echo hi"}}`, func() {
			if err := runHarnessObserve(cmd, nil); err != nil {
				t.Fatalf("runHarnessObserve returned error: %v", err)
			}
		})
		if calls != 0 {
			t.Errorf("fresh stamp spawned: calls=%d, want=0", calls)
		}
	})
}
