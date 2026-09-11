package cli

// Home-safety helper for init execution tests (SPEC-INIT-QUIET-WIZARD-001 M1,
// spec.md §4.2, design.md §5, AC-IQW-005 / AC-IQW-016).
//
// runInit reaches three home-rooted sinks: the user-scope settings splice
// (userHomeDirFn), the profile ledger (profile.BaseDirOverride), and the
// private MoAI home (MOAI_HOME). It also reaches the shell rc writer, which
// resolves HOME directly and cannot be redirected without overriding the HOME
// environment variable, which is forbidden here. prepareSafeInitHome redirects the three seams under
// t.TempDir(), replaces the shell-config seam with a counting spy, refuses to
// continue when any redirected home still resolves inside the real home, and
// compares an 8-item real-home fingerprint before and after the test.
//
// Tests using this helper must not call t.Parallel: the helper calls t.Setenv,
// so Go testing panics if they do.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/shell"
)

// initialShellConfigSeam is the shell-config seam value captured when the test
// binary initializes, before any test can swap it. prepareSafeInitHome compares
// against it to detect a previous test that did not restore the seam.
var initialShellConfigSeam = project.ConfigureShellEnvFn

// guardedShellRCFiles are the six shell rc files in the real-home fingerprint.
var guardedShellRCFiles = []string{".zshenv", ".zshrc", ".zprofile", ".profile", ".bashrc", ".bash_profile"}

// shellSeamSpy counts calls to the shell-config seam. It never writes.
type shellSeamSpy struct {
	calls atomic.Int32
}

// Calls reports how many times the seam was invoked.
func (s *shellSeamSpy) Calls() int {
	return int(s.calls.Load())
}

// checkSeamHomes is the pure guard decision: it returns an error when realHome
// is unknown, or when any seam path is empty, relative, equal to realHome, or
// located under realHome. Paths on another volume are treated as outside.
func checkSeamHomes(realHome string, seamPaths []string) error {
	if strings.TrimSpace(realHome) == "" {
		return errors.New("home guard: real home is empty; refusing to judge seam paths")
	}
	home := filepath.Clean(realHome)
	for _, p := range seamPaths {
		if strings.TrimSpace(p) == "" {
			return errors.New("home guard: a seam path is empty and would fall back to the real home")
		}
		clean := filepath.Clean(p)
		if !filepath.IsAbs(clean) {
			return fmt.Errorf("home guard: seam path %q is relative and cannot be judged", p)
		}
		rel, err := filepath.Rel(home, clean)
		if err != nil {
			// No relative path exists (e.g. a different Windows volume).
			continue
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		return fmt.Errorf("home guard: seam path %q resolves inside the real home %q", p, realHome)
	}
	return nil
}

// captureRealHomeFingerprint reads the 8 guarded real-home items, one line
// each, in a fixed order. A missing item is recorded as "absent" so that a
// newly created file shows up as a change. Read-only.
func captureRealHomeFingerprint(realHome string) ([]string, error) {
	lines := make([]string, 0, 2+len(guardedShellRCFiles))

	settingsPath := filepath.Join(realHome, ".claude", "settings.json")
	switch sum, err := homeGuardFileSHA256(settingsPath); {
	case err == nil:
		lines = append(lines, "settings.json sha256="+sum)
	case errors.Is(err, fs.ErrNotExist):
		lines = append(lines, "settings.json absent")
	default:
		return nil, fmt.Errorf("fingerprint settings.json: %w", err)
	}

	switch _, err := os.Stat(filepath.Join(realHome, ".claude", "hooks", "moai")); {
	case err == nil:
		lines = append(lines, "hooks/moai present")
	case errors.Is(err, fs.ErrNotExist):
		lines = append(lines, "hooks/moai absent")
	default:
		return nil, fmt.Errorf("fingerprint hooks/moai: %w", err)
	}

	for _, name := range guardedShellRCFiles {
		path := filepath.Join(realHome, name)
		info, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			lines = append(lines, name+" absent")
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("fingerprint %s: %w", name, err)
		}
		sum, err := homeGuardFileSHA256(path)
		if err != nil {
			return nil, fmt.Errorf("fingerprint %s: %w", name, err)
		}
		lines = append(lines, fmt.Sprintf("%s mtime=%d sha256=%s", name, info.ModTime().UnixNano(), sum))
	}
	return lines, nil
}

// homeGuardFileSHA256 returns the hex sha256 of the file at path.
func homeGuardFileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// prepareSafeInitHome makes a runInit execution test safe for the real home
// (spec.md §4.2 items 1-6) and returns the fake home plus the shell-config
// seam spy. Call it before runInit; never combine it with t.Parallel.
//
// @MX:WARN: [AUTO] Swaps package variables and environment (userHomeDirFn, profile.BaseDirOverride, MOAI_HOME, the shell-config seam); callers must not run in parallel.
// @MX:REASON: Without the spy in the shell-config seam, runInit reaches the real shell configurator, which resolves HOME directly and writes the developer's real rc files.
func prepareSafeInitHome(t *testing.T) (fakeHome string, spy *shellSeamSpy) {
	t.Helper()

	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home guard: resolve real home: %v", err)
	}

	// Item 5: fingerprint first, so the comparison registered here runs last
	// (t.Cleanup is LIFO), after every seam below has been restored.
	before, err := captureRealHomeFingerprint(realHome)
	if err != nil {
		t.Fatalf("home guard: fingerprint before: %v", err)
	}
	t.Cleanup(func() {
		after, err := captureRealHomeFingerprint(realHome)
		if err != nil {
			t.Errorf("home guard: fingerprint after: %v", err)
			return
		}
		if !slices.Equal(before, after) {
			t.Errorf("home guard: real home fingerprint changed during the test\nbefore:\n  %s\nafter:\n  %s",
				strings.Join(before, "\n  "), strings.Join(after, "\n  "))
		}
	})

	// Item 2: redirect the three home seams under t.TempDir(). MOAI_HOME goes
	// through t.Setenv, which also forbids t.Parallel for the calling test.
	fakeHome = t.TempDir()

	origHomeFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return fakeHome, nil }
	t.Cleanup(func() { userHomeDirFn = origHomeFn })

	origBase := profile.BaseDirOverride
	profile.BaseDirOverride = filepath.Join(fakeHome, ".moai", "claude-profiles")
	t.Cleanup(func() { profile.BaseDirOverride = origBase })

	t.Setenv("MOAI_HOME", filepath.Join(fakeHome, ".moai"))
	// Neutralize operator environment the init path may inherit (mirrors the
	// plan-phase reproduction probe that ran with a clean real-home result).
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	// Item 3: the seam must still hold its original value on entry; a
	// mismatch means an earlier test in this process did not restore it.
	if got, want := reflect.ValueOf(project.ConfigureShellEnvFn).Pointer(), reflect.ValueOf(initialShellConfigSeam).Pointer(); got != want {
		t.Fatalf("home guard: shell-config seam is %#x on entry, want the original %#x (a previous test did not restore it)", got, want)
	}
	spy = &shellSeamSpy{}
	origSeam := project.ConfigureShellEnvFn
	project.ConfigureShellEnvFn = func(_ *slog.Logger) (*shell.ConfigResult, error) {
		spy.calls.Add(1)
		// Report "already configured" and never call the real writer.
		return &shell.ConfigResult{Skipped: true}, nil
	}
	t.Cleanup(func() { project.ConfigureShellEnvFn = origSeam })

	// Item 4: refuse to run init unless every redirected home is outside the
	// real home. A post-run fingerprint alone would be too late.
	seamHome, err := userHomeDirFn()
	if err != nil {
		t.Fatalf("home guard: resolve seam home: %v", err)
	}
	if err := checkSeamHomes(realHome, []string{seamHome, profile.GetBaseDir(), os.Getenv("MOAI_HOME")}); err != nil {
		t.Fatalf("%v; refusing to run init", err)
	}

	return fakeHome, spy
}

// TestHomeGuard_RejectsPathInsideRealHome is the negative case for the guard
// decision: a seam path equal to or under the real home is refused, and so is
// an unjudgeable input.
func TestHomeGuard_RejectsPathInsideRealHome(t *testing.T) {
	realHome := filepath.Join(t.TempDir(), "home")
	outside := t.TempDir()

	cases := []struct {
		name      string
		realHome  string
		seamPaths []string
	}{
		{name: "equal_to_real_home", realHome: realHome, seamPaths: []string{realHome}},
		{name: "equal_with_trailing_separator", realHome: realHome, seamPaths: []string{realHome + string(filepath.Separator)}},
		{name: "direct_child", realHome: realHome, seamPaths: []string{filepath.Join(realHome, ".moai")}},
		{name: "deep_child", realHome: realHome, seamPaths: []string{filepath.Join(realHome, ".moai", "claude-profiles")}},
		{name: "child_named_with_leading_dots", realHome: realHome, seamPaths: []string{filepath.Join(realHome, "..moai")}},
		{name: "one_inside_among_outside", realHome: realHome, seamPaths: []string{outside, filepath.Join(realHome, ".claude"), outside}},
		{name: "empty_seam_path", realHome: realHome, seamPaths: []string{""}},
		{name: "relative_seam_path", realHome: realHome, seamPaths: []string{filepath.Join(".moai", "home")}},
		{name: "empty_real_home", realHome: "", seamPaths: []string{outside}},
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		cases = append(cases, struct {
			name      string
			realHome  string
			seamPaths []string
		}{name: "actual_real_home_child", realHome: home, seamPaths: []string{filepath.Join(home, ".moai")}})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkSeamHomes(tc.realHome, tc.seamPaths); err == nil {
				t.Errorf("checkSeamHomes(%q, %q) = nil, want an error", tc.realHome, tc.seamPaths)
			}
		})
	}
}

// TestHomeGuard_AcceptsTempDir is the positive case: t.TempDir() paths are
// accepted against the actual real home, and a sibling that only shares a
// name prefix with the real home is not mistaken for a child.
func TestHomeGuard_AcceptsTempDir(t *testing.T) {
	t.Run("temp_dirs_against_actual_real_home", func(t *testing.T) {
		realHome, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("resolve real home: %v", err)
		}
		fakeHome := t.TempDir()
		seamPaths := []string{
			fakeHome,
			filepath.Join(fakeHome, ".moai", "claude-profiles"),
			filepath.Join(fakeHome, ".moai"),
		}
		if err := checkSeamHomes(realHome, seamPaths); err != nil {
			t.Errorf("checkSeamHomes(real home, temp dirs) = %v, want nil", err)
		}
	})

	t.Run("sibling_sharing_name_prefix", func(t *testing.T) {
		base := t.TempDir()
		realHome := filepath.Join(base, "home")
		sibling := filepath.Join(base, "home2", ".moai")
		if err := checkSeamHomes(realHome, []string{sibling}); err != nil {
			t.Errorf("checkSeamHomes(%q, [%q]) = %v, want nil", realHome, sibling, err)
		}
	})
}
