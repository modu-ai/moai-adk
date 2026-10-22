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
	// The oversize judgment reads state.profile_max_bytes (no compiled
	// fallback), so the fixture configures a small HOME-tier cap and pins the
	// PROJECT tier to an empty dir.
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	writeHomeStateYAML(t, home, "state:\n    profile_max_bytes: 4096\n")
	writeHomeFixtureFile(t, oversize, 4096+1, time.Now())
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

// writeHomeStateYAML seeds the HOME-tier state.yaml
// (~/.moai/config/sections/state.yaml, the paths.UserConfigSectionsDir()
// tier hygiene.LoadProfileCapMaxBytes reads first).
func writeHomeStateYAML(t *testing.T, home, body string) {
	t.Helper()
	sections := filepath.Join(home, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "state.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("seed state.yaml: %v", err)
	}
}

// TestHomeDiskReportOversizeUsesConfiguredProfileCap pins the doctor's
// oversize-profile judgment to state.profile_max_bytes — the same loader the
// SessionStart advisory reads (HOME tier, then PROJECT tier, no compiled
// fallback) — instead of a compiled constant (card t1665). The PROJECT tier is
// pinned to an empty temp dir so the ambient checkout never leaks in.
func TestHomeDiskReportOversizeUsesConfiguredProfileCap(t *testing.T) {
	const capBytes int64 = 4096

	t.Run("configured cap flags exactly the profiles above it", func(t *testing.T) {
		home := hermeticHomeEnv(t)
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
		writeHomeStateYAML(t, home, "state:\n    profile_max_bytes: 4096\n")
		profiles := filepath.Join(home, ".moai", "claude-profiles")
		writeHomeFixtureFile(t, filepath.Join(profiles, "above", "projects", "s.jsonl"), capBytes+1, time.Now())
		writeHomeFixtureFile(t, filepath.Join(profiles, "at-cap", "projects", "s.jsonl"), capBytes, time.Now())
		writeHomeFixtureFile(t, filepath.Join(profiles, "below", "projects", "s.jsonl"), 100, time.Now())

		report := gatherHomeDiskReport(filepath.Join(home, ".moai"))
		if len(report.OversizeProfiles) != 1 || report.OversizeProfiles[0] != "above" {
			t.Fatalf("oversize profiles = %v, want exactly [above] under a %d-byte cap", report.OversizeProfiles, capBytes)
		}
		detail := renderHomeDiskDetail(report)
		want := "profiles above " + formatDiskBytes(capBytes) + ": above"
		if !strings.Contains(detail, want) {
			t.Fatalf("detail missing %q:\n%s", want, detail)
		}
	})

	t.Run("key absent in every tier flags nothing", func(t *testing.T) {
		home := hermeticHomeEnv(t)
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
		// state.yaml present, key absent — the sharper absent case.
		writeHomeStateYAML(t, home, "state:\n    home_retention_days: 30\n")
		// 6 GiB logical (sparse): above any plausible compiled cap, so a
		// compiled fallback would flag it here.
		writeHomeFixtureFile(t,
			filepath.Join(home, ".moai", "claude-profiles", "huge", "projects", "s.jsonl"),
			6*1024*1024*1024, time.Now())

		report := gatherHomeDiskReport(filepath.Join(home, ".moai"))
		if len(report.OversizeProfiles) != 0 {
			t.Fatalf("oversize profiles = %v, want none when state.profile_max_bytes is absent", report.OversizeProfiles)
		}
		if detail := renderHomeDiskDetail(report); strings.Contains(detail, "profiles above") {
			t.Fatalf("detail must carry no \"profiles above\" line when the cap is absent:\n%s", detail)
		}
	})

	t.Run("loader error skips the check and names the error", func(t *testing.T) {
		home := hermeticHomeEnv(t)
		t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
		writeHomeStateYAML(t, home, "state:\n    profile_max_bytes: -1\n")
		writeHomeFixtureFile(t,
			filepath.Join(home, ".moai", "claude-profiles", "huge", "projects", "s.jsonl"),
			6*1024*1024*1024, time.Now())

		report := gatherHomeDiskReport(filepath.Join(home, ".moai"))
		if len(report.OversizeProfiles) != 0 {
			t.Fatalf("oversize profiles = %v, want none when the cap cannot be loaded", report.OversizeProfiles)
		}
		detail := renderHomeDiskDetail(report)
		if strings.Contains(detail, "profiles above") {
			t.Fatalf("detail must carry no \"profiles above\" line on a loader error:\n%s", detail)
		}
		if !strings.Contains(detail, "state.profile_max_bytes must be strictly positive") {
			t.Fatalf("detail must name the loader error:\n%s", detail)
		}
	})
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

// SPEC-DOCTOR-PLUGIN-DIGEST-PREFILTER-001 AC-DPP-001 — negative control. A
// profile alone in its (Size, Files) group can never be a member of a
// byte-identical cluster, so nothing here is a hashing candidate.
func TestPluginDigestCandidatesAllDistinctYieldsNone(t *testing.T) {
	stats := map[string]map[string]profileCategoryStat{
		"alpha":   {"plugins": {Size: 10, Files: 1}},
		"beta":    {"plugins": {Size: 20, Files: 1}},
		"gamma":   {"plugins": {Size: 10, Files: 2}},
		"delta":   {"plugins": {Size: 0, Files: 0}},
		"epsilon": {"projects": {Size: 10, Files: 1}},
	}
	if got := pluginDigestCandidates(stats); len(got) != 0 {
		t.Fatalf("no profile shares a (Size, Files) pair, so none may be a candidate: %+v", got)
	}
}

// SPEC-DOCTOR-PLUGIN-DIGEST-PREFILTER-001 AC-DPP-002 — positive control. Two
// profiles sharing one pair both survive the prefilter, so the hashing and
// cluster-confirmation path still runs for them.
func TestPluginDigestCandidatesSharedPairSurvives(t *testing.T) {
	stats := map[string]map[string]profileCategoryStat{
		"alpha": {"plugins": {Size: 4, Files: 1}},
		"beta":  {"plugins": {Size: 4, Files: 1}},
	}
	got := pluginDigestCandidates(stats)
	want := []string{"alpha", "beta"}
	if len(got) != len(want) {
		t.Fatalf("both sharing profiles must be candidates, got %+v want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidates mismatch at %d: got %+v want %+v", i, got, want)
		}
	}
}

// SPEC-DOCTOR-PLUGIN-DIGEST-PREFILTER-001 AC-DPP-003 — mixed input plus
// determinism: exactly the sharing pair, sorted, and stable across repeated
// calls on the same input (map iteration order must not reach the result).
func TestPluginDigestCandidatesMixedIsSortedAndDeterministic(t *testing.T) {
	stats := map[string]map[string]profileCategoryStat{
		"zulu":    {"plugins": {Size: 7, Files: 3}},
		"mike":    {"plugins": {Size: 7, Files: 3}},
		"alpha":   {"plugins": {Size: 9, Files: 3}},
		"bravo":   {"plugins": {Size: 7, Files: 4}},
		"charlie": {"plugins": {Size: 0, Files: 0}},
	}
	want := []string{"mike", "zulu"}
	for call := 0; call < 3; call++ {
		got := pluginDigestCandidates(stats)
		if len(got) != len(want) {
			t.Fatalf("call %d: got %+v want %+v", call, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("call %d: candidates mismatch at %d: got %+v want %+v", call, i, got, want)
			}
		}
	}
}
