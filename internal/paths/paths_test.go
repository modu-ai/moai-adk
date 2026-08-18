package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// This file never calls t.Parallel(): every test mutates process-wide
// environment (HOME, MOAI_HOME) via t.Setenv, which forbids parallel tests
// (CLAUDE.local.md §6/§13 discipline, REQ-MHP-011).

// Signature pins: the exported surface is exactly the sketched accessor set
// (spec.md §3). Adding, removing, or reshaping an accessor fails compilation
// here (plan.md §4 Resolved Decisions).
var (
	_ func() (string, error) = Home
	_ func() (string, error) = MoaiHome
	_ func() (string, error) = StateDir
	_ func() (string, error) = CacheDir
	_ func() (string, error) = ReleasesDir
	_ func() (string, error) = WorktreesDir
	_ func() (string, error) = ProfilesDir
	_ func() (string, error) = GlmEnvFile
	_ func() (string, error) = UserSettingsFile
	_ func() (string, error) = UserConfigSectionsDir
)

func TestHome_HOMEEnvWins(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	got, err := Home()
	if err != nil {
		t.Fatalf("Home() error = %v, want nil", err)
	}
	if got != tmp {
		t.Errorf("Home() = %q, want %q (HOME must win)", got, tmp)
	}
}

func TestHome_EmptyHOMEFallsBackToUserHomeDir(t *testing.T) {
	// An empty HOME is handed to os.UserHomeDir, whose result (or error)
	// must pass through unchanged. Characterized rather than hard-coded so
	// the assertion holds on every platform (unix: error; windows:
	// USERPROFILE-based value).
	t.Setenv("HOME", "")

	got, gotErr := Home()
	want, wantErr := os.UserHomeDir()

	if got != want {
		t.Errorf("Home() = %q, want %q (os.UserHomeDir passthrough)", got, want)
	}
	if (gotErr == nil) != (wantErr == nil) {
		t.Errorf("Home() error = %v, want os.UserHomeDir()'s error state (%v)", gotErr, wantErr)
	}
}

func TestMoaiHome_OverrideAbsolute(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Platform-native absolute fixture: a unix "/tmp/..." literal is not
	// filepath.IsAbs on Windows, so the override would be disregarded there
	// and the assertion would compare against the fallback root instead.
	override := filepath.Join(os.TempDir(), "moai-home-override")
	t.Setenv("MOAI_HOME", override)

	got, err := MoaiHome()
	if err != nil {
		t.Fatalf("MoaiHome() error = %v, want nil", err)
	}
	if got != override {
		t.Errorf("MoaiHome() = %q, want %q (absolute MOAI_HOME honoured verbatim)", got, override)
	}
}

func TestMoaiHome_WindowsVolumeAbsolute(t *testing.T) {
	// Volume-letter absoluteness (D:\...) is only observable where
	// filepath.IsAbs carries Windows semantics; elsewhere the stdlib
	// predicate correctly reports it non-absolute, so the case is skipped
	// rather than asserted against the wrong platform's semantics. The 8.3
	// short-form variant pins that short paths pass through verbatim too —
	// nothing in the resolution chain may canonicalize them (windows CI
	// runners surface RUNNER~1-style temp roots).
	if runtime.GOOS != "windows" {
		t.Skip("windows volume-letter absoluteness is only observable on GOOS=windows")
	}
	for _, override := range []string{`D:\moai-home-override`, `C:\PROGRA~1\moai-home`} {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("MOAI_HOME", override)

		got, err := MoaiHome()
		if err != nil {
			t.Fatalf("MoaiHome() error = %v, want nil", err)
		}
		if got != override {
			t.Errorf("MoaiHome() = %q, want %q (windows absolute MOAI_HOME honoured verbatim)", got, override)
		}
	}
}

func TestMoaiHome_OverrideEmptyEqualsUnset(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	t.Setenv("MOAI_HOME", "")
	empty, err := MoaiHome()
	if err != nil {
		t.Fatalf("MoaiHome() with empty MOAI_HOME error = %v, want nil", err)
	}

	if err := os.Unsetenv("MOAI_HOME"); err != nil {
		t.Fatalf("unset MOAI_HOME: %v", err)
	}
	unset, err := MoaiHome()
	if err != nil {
		t.Fatalf("MoaiHome() with unset MOAI_HOME error = %v, want nil", err)
	}

	want := filepath.Join(home, ".moai")
	if empty != want {
		t.Errorf("MoaiHome() with MOAI_HOME=\"\" = %q, want %q (empty == unset)", empty, want)
	}
	if unset != want {
		t.Errorf("MoaiHome() with MOAI_HOME unset = %q, want %q", unset, want)
	}
}

func TestMoaiHome_OverrideRelativeDisregarded(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", "relative/moai/root")

	got, err := MoaiHome()
	if err != nil {
		t.Fatalf("MoaiHome() error = %v, want nil", err)
	}
	want := filepath.Join(home, ".moai")
	if got != want {
		t.Errorf("MoaiHome() with relative MOAI_HOME = %q, want %q (relative value disregarded)", got, want)
	}
}

func TestMoaiHome_FallbackUnderHOME(t *testing.T) {
	// AC-MHP-005: with HOME pointing at a temp dir, the fallback resolves
	// under it even where os.UserHomeDir() would consult other variables
	// (Windows). On unix the two coincide; the contract is what is pinned.
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := MoaiHome()
	if err != nil {
		t.Fatalf("MoaiHome() error = %v, want nil", err)
	}
	want := filepath.Join(home, ".moai")
	if got != want {
		t.Errorf("MoaiHome() = %q, want %q (HOME-first fallback)", got, want)
	}
}

func TestMoaiHome_ErrorNeverDots(t *testing.T) {
	// REQ-MHP-006: on resolution failure MoaiHome returns an error with an
	// empty string — never a "."-style relative fallback. On platforms where
	// os.UserHomeDir still resolves (windows USERPROFILE), the success arm
	// must still never produce ".".
	t.Setenv("HOME", "")
	t.Setenv("MOAI_HOME", "")

	got, err := MoaiHome()
	if err != nil {
		if got != "" {
			t.Errorf("MoaiHome() on error = %q, want empty string", got)
		}
		return
	}
	if got == "." || !filepath.IsAbs(got) {
		t.Errorf("MoaiHome() = %q, want an absolute path, never \".\"", got)
	}
}

func TestSubAccessors_OverrideAndFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Platform-native absolute override root (a unix "/tmp/..." literal is
	// not filepath.IsAbs on Windows — see TestMoaiHome_OverrideAbsolute).
	overrideRoot := filepath.Join(os.TempDir(), "moai-home-override")

	cases := []struct {
		name     string
		call     func() (string, error)
		segment  string
	}{
		{"StateDir", StateDir, "state"},
		{"CacheDir", CacheDir, "cache"},
		{"ReleasesDir", ReleasesDir, "releases"},
		{"WorktreesDir", WorktreesDir, "worktrees"},
		{"ProfilesDir", ProfilesDir, "claude-profiles"},
		{"GlmEnvFile", GlmEnvFile, ".env.glm"},
		{"UserSettingsFile", UserSettingsFile, "settings.json"},
		{"UserConfigSectionsDir", UserConfigSectionsDir, "config/sections"},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/fallback", func(t *testing.T) {
			t.Setenv("MOAI_HOME", "")
			got, err := tc.call()
			if err != nil {
				t.Fatalf("%s() error = %v, want nil", tc.name, err)
			}
			want := filepath.Join(home, ".moai", filepath.FromSlash(tc.segment))
			if got != want {
				t.Errorf("%s() = %q, want %q", tc.name, got, want)
			}
		})

		t.Run(tc.name+"/override", func(t *testing.T) {
			t.Setenv("MOAI_HOME", overrideRoot)
			got, err := tc.call()
			if err != nil {
				t.Fatalf("%s() error = %v, want nil", tc.name, err)
			}
			want := filepath.Join(overrideRoot, filepath.FromSlash(tc.segment))
			if got != want {
				t.Errorf("%s() under MOAI_HOME = %q, want %q", tc.name, got, want)
			}
		})
	}
}

func TestSubAccessors_ErrorPassthrough(t *testing.T) {
	// When neither HOME nor os.UserHomeDir can resolve a home, every
	// accessor surfaces the error rather than inventing a path. On windows
	// (USERPROFILE resolvable with HOME empty) the success arm must still
	// produce an absolute path.
	t.Setenv("HOME", "")
	t.Setenv("MOAI_HOME", "")

	calls := map[string]func() (string, error){
		"StateDir":             StateDir,
		"CacheDir":             CacheDir,
		"ReleasesDir":          ReleasesDir,
		"WorktreesDir":         WorktreesDir,
		"ProfilesDir":          ProfilesDir,
		"GlmEnvFile":           GlmEnvFile,
		"UserSettingsFile":     UserSettingsFile,
		"UserConfigSectionsDir": UserConfigSectionsDir,
	}

	for name, call := range calls {
		got, err := call()
		if err != nil {
			if got != "" {
				t.Errorf("%s() on error = %q, want empty string", name, got)
			}
			continue
		}
		if !filepath.IsAbs(got) {
			t.Errorf("%s() = %q, want absolute path", name, got)
		}
	}
}

func TestResolution_PerCall(t *testing.T) {
	// REQ-MHP-005: resolution is per call — a HOME change between two
	// MoaiHome calls in the same process must be observed by the second.
	t.Setenv("HOME", t.TempDir())
	first, err := MoaiHome()
	if err != nil {
		t.Fatalf("first MoaiHome() error = %v", err)
	}

	secondHome := t.TempDir()
	t.Setenv("HOME", secondHome)
	second, err := MoaiHome()
	if err != nil {
		t.Fatalf("second MoaiHome() error = %v", err)
	}

	if first == second {
		t.Errorf("MoaiHome() before/after HOME change = %q / %q, want different values", first, second)
	}
	want := filepath.Join(secondHome, ".moai")
	if second != want {
		t.Errorf("second MoaiHome() = %q, want %q", second, want)
	}
}
