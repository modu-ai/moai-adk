package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// launchModelFixture drives the real runUnifiedLaunch flow for a named profile
// and returns the model argv value ("" when no --model flag was passed) plus
// the launcher's stderr notices. Fixture knobs mirror the four precedence
// levels of card t1441: explicit --model, profile model, the user-scope
// settings.json /model value, and the project's .claude/settings.json pin.
//
// NOTE: callers must not use t.Parallel() — the helper chdirs and overrides
// package-level seams (findProjectRootFn, execOrSpawnClaudeFunc,
// launcherStderr, profile.BaseDirOverride).
func launchModelFixture(t *testing.T, explicitArgs []string, profileModel, userModel, projectPin string) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	for _, d := range []string{".moai", ".claude"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if projectPin != "" {
		body := "{\n  \"model\": \"" + projectPin + "\"\n}\n"
		if err := os.WriteFile(filepath.Join(tmpDir, ".claude", "settings.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	profileBase := t.TempDir()
	origBase := profile.BaseDirOverride
	t.Cleanup(func() { profile.BaseDirOverride = origBase })
	profile.BaseDirOverride = profileBase
	namedDir := filepath.Join(profileBase, "mo.ai.kr")
	if err := os.MkdirAll(namedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	prefs := "user_name: tester\neffort_level: medium\n"
	if profileModel != "" {
		prefs += "model: " + profileModel + "\n"
	}
	if err := os.WriteFile(filepath.Join(namedDir, "preferences.yaml"), []byte(prefs), 0o644); err != nil {
		t.Fatal(err)
	}
	if userModel != "" {
		// EnsureDir points CLAUDE_CONFIG_DIR at the profile dir, so this is the
		// user-scope settings.json that /model writes for the launch.
		body := "{\n  \"model\": \"" + userModel + "\"\n}\n"
		if err := os.WriteFile(filepath.Join(namedDir, "settings.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv(config.EnvClaudeConfigDir, "")

	origDir, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	origRoot := findProjectRootFn
	t.Cleanup(func() { findProjectRootFn = origRoot })
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-stub")))

	var capturedArgs []string
	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(bin string, args []string, env []string) error {
		capturedArgs = args
		return nil
	}

	var stderr strings.Builder
	origErr := launcherStderr
	t.Cleanup(func() { launcherStderr = origErr })
	launcherStderr = &stderr

	if err := runUnifiedLaunch("mo.ai.kr", "claude", explicitArgs); err != nil {
		t.Fatalf("runUnifiedLaunch error: %v", err)
	}

	got := ""
	for i, a := range capturedArgs {
		if a == "--model" && i+1 < len(capturedArgs) {
			got = capturedArgs[i+1]
		}
	}
	return got, stderr.String()
}

// TestLaunchModelPrecedence pins the four-level order of card t1441:
// --model > profile model > user-scope /model value > project pin (nothing
// passed, pin announced). Each row sets every LOWER level too, so a row only
// passes when the higher level really beats them.
func TestLaunchModelPrecedence(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		profileModel string
		userModel    string
		projectPin   string
		wantModel    string
		wantNotes    []string
		wantNoNotes  []string
	}{
		{
			name: "1 explicit --model beats everything", args: []string{"--model", "haiku"},
			profileModel: "sonnet[1m]", userModel: "opus", projectPin: "sonnet",
			wantModel: "haiku", wantNoNotes: []string{"user /model", "project pin"},
		},
		{
			name:         "2 profile model beats user /model and the pin",
			profileModel: "fable[1m]", userModel: "opus", projectPin: "sonnet",
			wantModel: "fable[1m]", wantNoNotes: []string{"user /model", "project pin"},
		},
		{
			name:      "3 user-scope /model beats the project pin and is announced",
			userModel: "opus", projectPin: "sonnet",
			wantModel: "opus", wantNotes: []string{"model: opus (user /model)"},
			wantNoNotes: []string{"project pin"},
		},
		{
			name:       "4 nothing chosen: no --model, pin named with the ways to change it",
			projectPin: "sonnet",
			wantModel:  "",
			wantNotes:  []string{`"model": "sonnet"`, "moai profile setup", "settings.local.json", "project pin"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model, stderr := launchModelFixture(t, tc.args, tc.profileModel, tc.userModel, tc.projectPin)
			if model != tc.wantModel {
				t.Errorf("--model = %q, want %q\n--- stderr ---\n%s", model, tc.wantModel, stderr)
			}
			for _, want := range tc.wantNotes {
				if !strings.Contains(stderr, want) {
					t.Errorf("notice missing %q\n--- stderr ---\n%s", want, stderr)
				}
			}
			for _, bad := range tc.wantNoNotes {
				if strings.Contains(stderr, bad) {
					t.Errorf("notice must not contain %q\n--- stderr ---\n%s", bad, stderr)
				}
			}
		})
	}
}
