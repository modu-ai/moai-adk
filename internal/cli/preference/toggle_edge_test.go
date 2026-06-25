package preference

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDisablePersonalization_EmptyProjectRoot verifies the validation gate.
func TestDisablePersonalization_EmptyProjectRoot(t *testing.T) {
	t.Parallel()
	err := DisablePersonalization("", time.Now())
	if err == nil {
		t.Error("DisablePersonalization(empty) returned nil; want validation error")
	}
}

// TestEnablePersonalization_EmptyProjectRoot verifies the validation gate.
func TestEnablePersonalization_EmptyProjectRoot(t *testing.T) {
	t.Parallel()
	err := EnablePersonalization("")
	if err == nil {
		t.Error("EnablePersonalization(empty) returned nil; want validation error")
	}
}

// TestDisablePersonalization_ProjectRootFlagPath verifies runToggle honors the
// explicit --project-root flag (not just the env-derived root).
func TestRunToggle_ProjectRootFlagPath(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	var stdout, stderr strings.Builder
	flags := &toggleFlags{projectRoot: tmp, json: false}
	if err := runToggle(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runToggle: %v\nstderr: %s", err, stderr.String())
	}
	if !IsPersonalizationDisabled(tmp) {
		t.Errorf("runToggle --project-root=%s did not disable personalization", tmp)
	}
	// Clear env to prove the flag path does not depend on $CLAUDE_PROJECT_DIR.
	// (We cannot t.Setenv to empty here because parallel; but the explicit
	// --project-root wins regardless.)
}

// TestRunToggle_TextOutputContainsState verifies the default (non-JSON) output
// line names the resulting state so a human reader sees the toggle result.
func TestRunToggle_TextOutputContainsState(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	var stdout, stderr strings.Builder
	flags := &toggleFlags{projectRoot: tmp, json: false}
	if err := runToggle(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runToggle: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "personalization") || !strings.Contains(out, "disabled") {
		t.Errorf("runToggle text output = %q, must contain 'personalization' + 'disabled'", out)
	}
}

// TestCleanupStaleSentinel_EmptyProjectRootIsNoop verifies the SessionStart
// helper tolerates an empty root (no-op, no error) so the hook does not crash
// on a misconfigured environment.
func TestCleanupStaleSentinel_EmptyProjectRootIsNoop(t *testing.T) {
	t.Parallel()
	if err := CleanupStaleSentinel(""); err != nil {
		t.Errorf("CleanupStaleSentinel(empty) returned error: %v", err)
	}
}

// TestDisablePersonalization_OverwritesExistingSentinel verifies that disabling
// when the sentinel already exists overwrites it with the fresh timestamp
// (idempotent, no error) rather than failing.
func TestDisablePersonalization_OverwritesExistingSentinel(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	t1 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	if err := DisablePersonalization(tmp, t1); err != nil {
		t.Fatalf("Disable(1): %v", err)
	}
	if err := DisablePersonalization(tmp, t2); err != nil {
		t.Fatalf("Disable(2): %v", err)
	}
	// Sentinel content reflects the LATER timestamp.
	data, err := os.ReadFile(sentinelPath(tmp))
	if err != nil {
		t.Fatalf("read sentinel: %v", err)
	}
	if !strings.Contains(string(data), t2.Format(time.RFC3339)) {
		t.Errorf("sentinel after overwrite = %q, want timestamp %s", string(data), t2.Format(time.RFC3339))
	}
}

// TestSentinelPersistsInStateDirNotConfigDir is the filesystem-level cross-check
// for NFR-ADM-005: the sentinel MUST live under .moai/state/, and the .moai/
// config dir MUST be untouched by the toggle. (TestSentinelDoesNotLeakIntoPermanentConfig
// already covers the config-side grep; this test asserts the state-side path.)
func TestSentinelPersistsInStateDirNotConfigDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	if err := DisablePersonalization(tmp, time.Now()); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	stateDir := filepath.Join(tmp, ".moai", "state")
	if _, err := os.Stat(filepath.Join(stateDir, sessionDisabledSentinelName)); err != nil {
		t.Errorf("sentinel not in expected state dir %s: %v", stateDir, err)
	}
}
