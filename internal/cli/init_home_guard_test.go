package cli

// Home-safety helper for init execution tests (SPEC-INIT-QUIET-WIZARD-001 M1,
// spec.md §4.2, design.md §5, AC-IQW-005 / AC-IQW-016).
//
// runInit reaches three home-rooted sinks: the user-scope settings splice
// (userHomeDirFn), the profile ledger (profile.BaseDirOverride), and the
// private MoAI home (MOAI_HOME). It also reaches the shell rc writer, which
// resolves HOME directly and cannot be redirected without overriding the HOME
// environment variable, which is forbidden here. prepareSafeInitHome redirects
// the three seams under an owned temporary directory outside the real home,
// replaces the shell-config seam with a counting spy, refuses to
// continue when any redirected home still resolves inside the real home, and
// compares an 8-item real-home fingerprint before and after the test.
//
// The real-home resolver and the containment judgment (testguard.CheckSeamHomes)
// are shared with every other home guard in the repository; they live in
// internal/testguard and are not duplicated here.
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
	"github.com/modu-ai/moai-adk/internal/testguard"
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

// safeInitTempBase validates an existing temp base before any fixture is
// created. Prefer the standard temp directory when it is strictly outside the
// real home; otherwise validate RUNNER_TEMP under the same strict judgment.
func safeInitTempBase(realHome, runnerTemp, tempDir string) (string, error) {
	validate := func(base string) (string, error) {
		if err := testguard.CheckSeamHomes(realHome, []string{base}); err != nil {
			return "", err
		}
		resolved, err := filepath.EvalSymlinks(base)
		if err != nil {
			return "", fmt.Errorf("resolve fixture temp base %q: %w", base, err)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return "", fmt.Errorf("stat fixture temp base %q: %w", base, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("fixture temp base %q is not a directory", base)
		}
		if err := testguard.CheckSeamHomes(realHome, []string{resolved}); err != nil {
			return "", err
		}
		return resolved, nil
	}
	base, tempErr := validate(tempDir)
	if tempErr == nil {
		return base, nil
	}
	base, runnerErr := validate(runnerTemp)
	if runnerErr != nil {
		return "", fmt.Errorf("standard temp base: %v; RUNNER_TEMP base: %w", tempErr, runnerErr)
	}
	return base, nil
}

// safeInitTempDir owns only its fresh child, never the runner's temp base.
// It does not change HOME or the process-wide temporary-directory variables.
// @MX:ANCHOR: Allocates strict init fixture homes and projects for five caller functions.
// @MX:REASON: Every caller depends on the base and child checks to keep fixture writes outside the real home without HOME or TEMP overrides.
func safeInitTempDir(t *testing.T) string {
	t.Helper()
	realHome, err := testguard.RealHome()
	if err != nil {
		t.Fatalf("home guard: resolve real home: %v", err)
	}
	base, err := safeInitTempBase(realHome, os.Getenv("RUNNER_TEMP"), os.TempDir())
	if err != nil {
		t.Fatalf("home guard: select fixture temp base: %v", err)
	}
	dir, err := os.MkdirTemp(base, "moai-init-fixture-")
	if err != nil {
		t.Fatalf("home guard: create owned fixture under %q: %v", base, err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("home guard: remove owned fixture %q: %v", dir, err)
		}
	})
	if err := testguard.CheckSeamHomes(realHome, []string{dir}); err != nil {
		t.Fatalf("%v; refusing to use fixture directory", err)
	}
	t.Logf("home guard: owned fixture %q under validated temp base %q", dir, base)
	return dir
}

func TestSafeInitTempBase_PreservesStrictHomeBoundary(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "real-home")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{home, outside} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, runner, temp string
		wantOK             bool
	}{
		{"runner outside home", outside, home, true},
		{"safe standard temp ignores unsafe runner", home, outside, true},
		{"standard temp and runner both in home", home, home, false},
		{"standard temp in home", "", home, false},
		{"ancestor runner with unsafe standard temp", root, home, false},
		{"relative runner with unsafe standard temp", "relative", home, false},
		{"missing runner with unsafe standard temp", filepath.Join(outside, "missing"), home, false},
		{"file runner with unsafe standard temp", file, home, false},
		{"missing standard temp", "", filepath.Join(outside, "missing"), false},
		{"empty bases", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := safeInitTempBase(home, tc.runner, tc.temp)
			if (err == nil) != tc.wantOK {
				t.Fatalf("safeInitTempBase = (%q, %v), want success %v", got, err, tc.wantOK)
			}
			if tc.wantOK {
				want, err := filepath.EvalSymlinks(outside)
				if err != nil || got != want {
					t.Fatalf("safeInitTempBase = %q, want validated outside directory %q (resolve error: %v)", got, want, err)
				}
			}
		})
	}
	if _, err := os.Stat(filepath.Join(outside, "missing")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("base validation created the missing directory: %v", err)
	}
}

// prepareSafeInitHome makes a runInit execution test safe for the real home
// (spec.md §4.2 items 1-6) and returns the fake home plus the shell-config
// seam spy. Call it before runInit; never combine it with t.Parallel.
//
// @MX:WARN: [AUTO] Swaps package variables and environment (userHomeDirFn, profile.BaseDirOverride, MOAI_HOME, the shell-config seam); callers must not run in parallel.
// @MX:REASON: Without the spy in the shell-config seam, runInit reaches the real shell configurator, which resolves HOME directly and writes the developer's real rc files.
func prepareSafeInitHome(t *testing.T) (fakeHome string, spy *shellSeamSpy) {
	t.Helper()

	realHome, err := testguard.RealHome()
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

	// Item 2: redirect the three home seams under an owned temp directory
	// strictly outside the real home. MOAI_HOME goes through t.Setenv, which
	// also forbids t.Parallel for the calling test.
	fakeHome = safeInitTempDir(t)

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
	if err := testguard.CheckSeamHomes(realHome, []string{seamHome, profile.GetBaseDir(), os.Getenv("MOAI_HOME")}); err != nil {
		t.Fatalf("%v; refusing to run init", err)
	}

	return fakeHome, spy
}
