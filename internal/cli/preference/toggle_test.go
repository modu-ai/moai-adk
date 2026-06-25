package preference

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// sessionDisabledSentinelName matches design.md §A.5: the sentinel file under
// the project's .moai/state/ whose presence signals "personalization disabled
// for this session". Re-declared here as a cross-check on the production
// constant; if they drift this test fails.
const sessionDisabledSentinelNameTest = "session-preference-disabled"

// TestIsPersonalizationDisabled_SentinelAbsent verifies the default state:
// personalization is ACTIVE when the sentinel does not exist (REQ-ADM-013,
// NFR-ADM-005).
func TestIsPersonalizationDisabled_SentinelAbsent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	if IsPersonalizationDisabled(tmp) {
		t.Error("IsPersonalizationDisabled = true, want false (sentinel absent → personalization active)")
	}
}

// TestDisablePersonalization_CreatesSentinel verifies DisablePersonalization
// writes the sentinel file so IsPersonalizationDisabled subsequently returns
// true.
func TestDisablePersonalization_CreatesSentinel(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	if err := DisablePersonalization(tmp, now); err != nil {
		t.Fatalf("DisablePersonalization: %v", err)
	}
	if !IsPersonalizationDisabled(tmp) {
		t.Error("DisablePersonalization did not flip IsPersonalizationDisabled to true")
	}
	// The sentinel file exists at the design.md §A.5 path.
	stateDir := filepath.Join(tmp, ".moai", "state")
	_, err := os.Stat(filepath.Join(stateDir, sessionDisabledSentinelNameTest))
	if err != nil {
		t.Errorf("sentinel file missing after DisablePersonalization: %v", err)
	}
}

// TestEnablePersonalization_RemovesSentinel verifies EnablePersonalization
// removes the sentinel (re-activating personalization) and is idempotent on a
// clean state.
func TestEnablePersonalization_RemovesSentinel(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	// Disable then Enable — back to active.
	if err := DisablePersonalization(tmp, now); err != nil {
		t.Fatalf("DisablePersonalization: %v", err)
	}
	if err := EnablePersonalization(tmp); err != nil {
		t.Fatalf("EnablePersonalization: %v", err)
	}
	if IsPersonalizationDisabled(tmp) {
		t.Error("EnablePersonalization did not flip state back to active")
	}

	// Idempotent: Enable on an already-active state is a no-op (not an error).
	if err := EnablePersonalization(tmp); err != nil {
		t.Errorf("EnablePersonalization (idempotent re-call) returned error: %v", err)
	}
}

// TestToggleIsIdempotentRoundTrip verifies Disable→Enable→Disable→Enable leaves
// the state active (the CLI toggle command's "idempotent: toggling twice
// returns to original state" contract, constraint #6 of the spawn prompt).
func TestToggleIsIdempotentRoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	// Initial: active.
	if IsPersonalizationDisabled(tmp) {
		t.Fatal("initial state not active")
	}
	// Disable → disabled.
	_ = DisablePersonalization(tmp, now)
	if !IsPersonalizationDisabled(tmp) {
		t.Fatal("after Disable: not disabled")
	}
	// Enable → active.
	_ = EnablePersonalization(tmp)
	if IsPersonalizationDisabled(tmp) {
		t.Fatal("after Enable: not active (round-trip broken)")
	}
}

// TestCleanupStaleSentinel_RemovesSentinel verifies the SessionStart helper
// removes a stale sentinel so a NEW session reactivates personalization
// (REQ-ADM-013, NFR-ADM-005: "신규 세션에서 자동 재활성화").
func TestCleanupStaleSentinel_RemovesSentinel(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	// Session A disables.
	if err := DisablePersonalization(tmp, now); err != nil {
		t.Fatalf("DisablePersonalization: %v", err)
	}
	if !IsPersonalizationDisabled(tmp) {
		t.Fatal("session A: not disabled")
	}
	// New session B starts: SessionStart hook calls CleanupStaleSentinel.
	if err := CleanupStaleSentinel(tmp); err != nil {
		t.Fatalf("CleanupStaleSentinel: %v", err)
	}
	// Session B: personalization reactivated.
	if IsPersonalizationDisabled(tmp) {
		t.Error("CleanupStaleSentinel did not reactivate personalization for new session")
	}
}

// TestCleanupStaleSentinel_IdempotentOnCleanState verifies CleanupStaleSentinel
// does NOT error when the sentinel is already absent (the common SessionStart
// path — most sessions do not have a stale sentinel to clean).
func TestCleanupStaleSentinel_IdempotentOnCleanState(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	if err := CleanupStaleSentinel(tmp); err != nil {
		t.Errorf("CleanupStaleSentinel on clean state returned error: %v", err)
	}
}

// TestSentinelDoesNotLeakIntoPermanentConfig verifies NFR-ADM-005: the toggle
// state MUST NOT leak into permanent config files. The sentinel lives ONLY
// under .moai/state/, NEVER under .moai/config/.
func TestSentinelDoesNotLeakIntoPermanentConfig(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	_ = DisablePersonalization(tmp, now)

	// .moai/config/ MUST NOT contain the sentinel name anywhere.
	configDir := filepath.Join(tmp, ".moai", "config")
	leaked := false
	_ = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // config dir may not exist — that's fine
		}
		if strings.Contains(path, sessionDisabledSentinelNameTest) {
			leaked = true
		}
		return nil
	})
	if leaked {
		t.Error("sentinel name leaked into .moai/config/ — NFR-ADM-005 violation (toggle must be session-only, NOT permanent config)")
	}
}

// ---- CLI subcommand (newToggleCmd) ----

// TestNewToggleCmd_FlagsRegistered verifies the CLI surface.
func TestNewToggleCmd_FlagsRegistered(t *testing.T) {
	t.Parallel()
	cmd := newToggleCmd()
	for _, name := range []string{"memory-dir", "project-root", "json"} {
		if f := cmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag --%s is not registered on toggle", name)
		}
	}
}

// TestRunToggle_DisableThenEnable verifies the command body round-trip with
// JSON output. The toggle subcommand flips state on each invocation.
func TestRunToggle_DisableThenEnable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	// First invocation: active → disabled.
	var stdout1, stderr1 strings.Builder
	flags1 := &toggleFlags{projectRoot: tmp, json: true}
	if err := runToggle(&stdout1, &stderr1, flags1); err != nil {
		t.Fatalf("runToggle (1st): %v\nstderr: %s", err, stderr1.String())
	}
	var res1 struct {
		Disabled bool `json:"disabled"`
	}
	if err := json.Unmarshal([]byte(stdout1.String()), &res1); err != nil {
		t.Fatalf("parse toggle JSON %q: %v", stdout1.String(), err)
	}
	if !res1.Disabled {
		t.Errorf("1st toggle: disabled=false, want true (active→disabled)")
	}
	if !IsPersonalizationDisabled(tmp) {
		t.Error("1st toggle did not actually disable personalization (sentinel missing)")
	}

	// Second invocation: disabled → enabled.
	var stdout2, stderr2 strings.Builder
	flags2 := &toggleFlags{projectRoot: tmp, json: true}
	if err := runToggle(&stdout2, &stderr2, flags2); err != nil {
		t.Fatalf("runToggle (2nd): %v\nstderr: %s", err, stderr2.String())
	}
	var res2 struct {
		Disabled bool `json:"disabled"`
	}
	if err := json.Unmarshal([]byte(stdout2.String()), &res2); err != nil {
		t.Fatalf("parse toggle JSON %q: %v", stdout2.String(), err)
	}
	if res2.Disabled {
		t.Errorf("2nd toggle: disabled=true, want false (disabled→enabled)")
	}
	if IsPersonalizationDisabled(tmp) {
		t.Error("2nd toggle did not actually enable personalization (sentinel still present)")
	}
}

// TestRunToggle_EmptyProjectRootResolvedFromEnv verifies the empty
// --project-root path: resolution falls back to $CLAUDE_PROJECT_DIR (mirrors
// the cmd.go resolveProjectRoot priority).
func TestRunToggle_EmptyProjectRootResolvedFromEnv(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	var stdout, stderr strings.Builder
	flags := &toggleFlags{projectRoot: "", json: false}
	if err := runToggle(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runToggle empty project-root: %v", err)
	}
	if !IsPersonalizationDisabled(tmp) {
		t.Errorf("runToggle with empty --project-root did not resolve to $CLAUDE_PROJECT_DIR and disable")
	}
}

// TestSentinelPath is a minimal sanity check that the production sentinel-name
// constant matches the test's expected name (drift-detection cross-check).
func TestSentinelPath(t *testing.T) {
	t.Parallel()
	got := sentinelPath("/tmp/proj")
	want := filepath.Join("/tmp/proj", ".moai", "state", sessionDisabledSentinelNameTest)
	if got != want {
		t.Errorf("sentinelPath = %q, want %q", got, want)
	}
}

// errIsNotExist is a tiny helper to check os.ErrNotExist without importing
// errors in every test (kept for the one place it's used below).
func errIsNotExist(err error) bool { return errors.Is(err, os.ErrNotExist) }

// TestEnablePersonalization_NoSentinelIsNotError guards the idempotency path:
// Enable when no sentinel exists returns nil (not os.ErrNotExist).
func TestEnablePersonalization_NoSentinelIsNotError(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	err := EnablePersonalization(tmp)
	if err != nil && errIsNotExist(err) {
		t.Errorf("EnablePersonalization on clean state surfaced os.ErrNotExist: %v", err)
	}
	if err != nil {
		t.Errorf("EnablePersonalization on clean state returned unexpected error: %v", err)
	}
}
