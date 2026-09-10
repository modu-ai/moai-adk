package cli

// SPEC-V3R6-MOAI-CLEAN-HOME-001 REQ-MCH-001/002 — characterization of the
// advisory Home Disk Usage doctor check. All fixtures are hermetic per
// REQ-MCH-008: HOME pinned to a temp dir, ambient MOAI_HOME scrubbed,
// non-parallel.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

// hermeticHomeEnv pins the home resolution to an isolated temp HOME with any
// ambient MOAI_HOME override scrubbed (REQ-MCH-008; paths.go treats "" as
// unset). Non-parallel by contract: it mutates process env.
func hermeticHomeEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", "")
	return home
}

// agedTime returns a timestamp comfortably older than the default home
// retention window so fixture entries land past every cutoff used here.
func agedTime(t *testing.T) time.Time {
	t.Helper()
	return time.Now().AddDate(0, 0, -(config.DefaultHomeCleanRetentionDays + 10))
}

// writeHomeFixtureFile creates a file at path with the given logical size
// (truncated — sparse where the filesystem supports it, so multi-hundred-MB
// threshold fixtures stay cheap) and pins its mtime when aged is true.
func writeHomeFixtureFile(t *testing.T, path string, size int64, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if err := f.Truncate(size); err != nil {
		t.Fatalf("truncate %s to %d: %v", path, size, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
	if !mtime.IsZero() {
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatalf("chtimes %s: %v", path, err)
		}
	}
}

// TestCheckHomeDisk_NoHomeReportsOK covers the deterministic absent-home path
// (also the path golden snapshots exercise): OK, never a failure state.
func TestCheckHomeDisk_NoHomeReportsOK(t *testing.T) {
	hermeticHomeEnv(t)

	check := checkHomeDisk(false)
	if check.Name != "Home Disk Usage" {
		t.Errorf("Name = %q, want %q", check.Name, "Home Disk Usage")
	}
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok for absent ~/.moai", check.Status)
	}
	if !strings.Contains(check.Message, "no ~/.moai") {
		t.Errorf("Message should mention the absent home, got %q", check.Message)
	}
}

// TestCheckHomeDisk_BreakdownProfilesAndClaudeReportOnlyLine verifies the
// REQ-MCH-001 report surface: top-level breakdown, per-profile detail,
// releases count, and the ~/.claude summary line marked report-only.
func TestCheckHomeDisk_BreakdownProfilesAndClaudeReportOnlyLine(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")

	// Two profiles; plugins/ byte-equal so the cluster detector also fires.
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "alpha", "plugins", "p1.bin"), 4096, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "beta", "plugins", "p1.bin"), 4096, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "alpha", "projects", "session.jsonl"), 2048, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "alpha", "debug", "trace.log"), 100, agedTime(t))
	writeHomeFixtureFile(t, filepath.Join(root, "releases", "moai-v9.9.9-darwin-arm64"), 5000, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "logs", "run.log"), 30, agedTime(t))
	writeHomeFixtureFile(t, filepath.Join(home, ".claude", "settings.json"), 10, time.Time{})

	check := checkHomeDisk(true)

	if check.Name != "Home Disk Usage" {
		t.Errorf("Name = %q, want %q", check.Name, "Home Disk Usage")
	}
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok (cleanable %d bytes is below the warn threshold)", check.Status, 130)
	}
	for _, want := range []string{"~/.moai", "claude-profiles", "releases", "~/.claude", "(report-only)"} {
		if !strings.Contains(check.Message, want) {
			t.Errorf("Message should contain %q, got %q", want, check.Message)
		}
	}
	for _, want := range []string{"alpha", "beta", "plugins", "releases"} {
		if !strings.Contains(check.Detail, want) {
			t.Errorf("verbose Detail should mention %q, got %q", want, check.Detail)
		}
	}
}

// TestCheckHomeDisk_DuplicateClusterReported covers REQ-MCH-001/006 through
// the full check: two profiles carrying byte-equal plugins/ dirs surface as a
// report-only duplicate cluster naming both profiles.
func TestCheckHomeDisk_DuplicateClusterReported(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "alpha", "plugins", "a.bin"), 8192, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "alpha", "plugins", "b.bin"), 4096, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "beta", "plugins", "a.bin"), 8192, time.Time{})
	writeHomeFixtureFile(t, filepath.Join(root, "claude-profiles", "beta", "plugins", "b.bin"), 4096, time.Time{})

	check := checkHomeDisk(true)
	if !strings.Contains(check.Detail, "alpha") || !strings.Contains(check.Detail, "beta") {
		t.Errorf("Detail should name both duplicated profiles, got %q", check.Detail)
	}
	if !strings.Contains(check.Detail, "plugins") {
		t.Errorf("Detail should name the duplicated category, got %q", check.Detail)
	}
	if !strings.Contains(check.Detail, "report-only") {
		t.Errorf("cluster line must be marked report-only, got %q", check.Detail)
	}
}

func TestPluginHashDoesNotEquateSameSizeDifferentContent(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai", "claude-profiles")
	a := filepath.Join(root, "alpha", "plugins", "same.bin")
	b := filepath.Join(root, "beta", "plugins", "same.bin")
	if err := os.MkdirAll(filepath.Dir(a), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(b), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(a, []byte("AAAA"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("BBBB"), 0o600); err != nil {
		t.Fatal(err)
	}
	stats := map[string]map[string]profileCategoryStat{
		"alpha": {"plugins": {Size: 4, Files: 1}},
		"beta":  {"plugins": {Size: 4, Files: 1}},
	}
	if got := findPluginHashClusters(root, stats); len(got) != 0 {
		t.Fatalf("same-size different-content trees must not cluster: %+v", got)
	}
	if err := os.WriteFile(b, []byte("AAAA"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := findPluginHashClusters(root, stats)
	if len(got) != 1 || len(got[0].Profiles) != 2 || !strings.HasPrefix(got[0].Category, "plugins sha256=") {
		t.Fatalf("byte-identical plugin trees must form one SHA-256 cluster: %+v", got)
	}
}

func TestHomeDiskReportDetectsInactiveOversizeAndInsecureProfiles(t *testing.T) {
	home := hermeticHomeEnv(t)
	profiles := filepath.Join(home, ".moai", "claude-profiles")
	stale := filepath.Join(profiles, "stale-empty")
	fresh := filepath.Join(profiles, "fresh-empty")
	oversize := filepath.Join(profiles, "oversize", "projects", "session.jsonl")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(fresh, 0o700); err != nil {
		t.Fatal(err)
	}
	writeHomeFixtureFile(t, oversize, config.DefaultProfileMaxBytes+1, time.Now())
	old := time.Now().AddDate(0, 0, -(config.DefaultProfileUnusedDays + 1))
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}

	report := gatherHomeDiskReport(filepath.Join(home, ".moai"))
	if !containsDiskProfile(report.InactiveProfiles, "stale-empty") || containsDiskProfile(report.InactiveProfiles, "fresh-empty") {
		t.Fatalf("inactive profiles = %v, want stale-empty only from empty-profile pair", report.InactiveProfiles)
	}
	if !containsDiskProfile(report.OversizeProfiles, "oversize") {
		t.Fatalf("oversize profiles = %v, want oversize", report.OversizeProfiles)
	}
	if report.InsecureHomeDirs == 0 {
		t.Fatal("insecure profile directory count = 0, want stale mode 0755 detected")
	}
}

func containsDiskProfile(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestCheckHomeDisk_WarnsWhenCleanableExceedsDefaultThreshold covers
// REQ-MCH-002 above-threshold: an aged debug entry whose size pushes the
// cleanable estimate past DefaultHomeDiskWarnBytes yields WARN. The fixture
// file is truncated (sparse), so the >threshold bytes cost no real disk.
func TestCheckHomeDisk_WarnsWhenCleanableExceedsDefaultThreshold(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	writeHomeFixtureFile(t,
		filepath.Join(root, "claude-profiles", "alpha", "debug", "huge.log"),
		int64(config.DefaultHomeDiskWarnBytes)+1,
		agedTime(t))

	check := checkHomeDisk(false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("Status = %q, want warn when cleanable exceeds the default threshold", check.Status)
	}
	if !strings.Contains(check.Message, "moai clean --home") {
		t.Errorf("WARN message should point at 'moai clean --home', got %q", check.Message)
	}
}

// TestCheckHomeDisk_OKWhenCleanableBelowDefaultThreshold: a small aged
// fixture stays OK — the threshold decision is >=-style only above the bar.
func TestCheckHomeDisk_OKWhenCleanableBelowDefaultThreshold(t *testing.T) {
	home := hermeticHomeEnv(t)
	root := filepath.Join(home, ".moai")
	writeHomeFixtureFile(t, filepath.Join(root, "logs", "run.log"), 42, agedTime(t))

	check := checkHomeDisk(false)
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok below threshold", check.Status)
	}
}

// TestCheckHomeDisk_RegisteredInMoaiADKGroup verifies the registration point:
// the check appears in the MoAI-ADK group of runGroupedChecksObserved (not
// only in the legacy flattener).
func TestCheckHomeDisk_RegisteredInMoaiADKGroup(t *testing.T) {
	hermeticHomeEnv(t)

	groups := runGroupedChecks(false, "")
	found := false
	for _, g := range groups {
		if g.title != "MoAI-ADK" {
			continue
		}
		for _, c := range g.checks {
			if c.Name == "Home Disk Usage" {
				found = true
			}
		}
	}
	if !found {
		t.Error("Home Disk Usage check is not registered in the MoAI-ADK group of runGroupedChecks")
	}
}
