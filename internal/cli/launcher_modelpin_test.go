package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// modelFixture names the knobs of one launch for card t1441. Each knob is one
// precedence level (or one way a level can be broken); an empty string leaves
// the level unset.
type modelFixture struct {
	args         []string // explicit launch args, e.g. --model haiku
	defaultProf  bool     // launch with no profile; CLAUDE_CONFIG_DIR still names the dir
	profileModel string   // preferences.yaml model
	envModel     string   // ANTHROPIC_MODEL
	localModel   string   // project .claude/settings.local.json "model"
	userModel    string   // user-scope settings.json "model"
	userRaw      string   // raw user-scope settings.json body (overrides userModel)
	projectPin   string   // project .claude/settings.json "model"
}

// launchModel drives the real runUnifiedLaunch flow with claude stubbed and
// returns the --model value passed ("" when no flag) plus the launcher notices.
//
// NOTE: callers must not use t.Parallel() — the helper chdirs and overrides
// package-level seams (findProjectRootFn, execOrSpawnClaudeFunc,
// launcherStderr, profile.BaseDirOverride) and sets environment variables.
func launchModel(t *testing.T, f modelFixture) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	for _, d := range []string{".moai", ".claude"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeModelJSON := func(path, model string) {
		if err := os.WriteFile(path, []byte("{\n  \"model\": \""+model+"\"\n}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if f.projectPin != "" {
		writeModelJSON(filepath.Join(tmpDir, ".claude", "settings.json"), f.projectPin)
	}
	if f.localModel != "" {
		writeModelJSON(filepath.Join(tmpDir, ".claude", "settings.local.json"), f.localModel)
	}

	profileBase := t.TempDir()
	origBase := profile.BaseDirOverride
	t.Cleanup(func() { profile.BaseDirOverride = origBase })
	profile.BaseDirOverride = profileBase

	// The config dir the launch uses: a named profile's dir (EnsureDir points
	// CLAUDE_CONFIG_DIR at it), or an explicit CLAUDE_CONFIG_DIR for the default
	// profile.
	configDir := filepath.Join(profileBase, "mo.ai.kr")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	prefs := "user_name: tester\neffort_level: medium\n"
	if f.profileModel != "" {
		prefs += "model: " + f.profileModel + "\n"
	}
	prefsDir := configDir
	profileName := "mo.ai.kr"
	if f.defaultProf {
		prefsDir = profileBase
		profileName = ""
	}
	if err := os.WriteFile(filepath.Join(prefsDir, "preferences.yaml"), []byte(prefs), 0o644); err != nil {
		t.Fatal(err)
	}
	switch {
	case f.userRaw != "":
		if err := os.WriteFile(filepath.Join(configDir, "settings.json"), []byte(f.userRaw), 0o644); err != nil {
			t.Fatal(err)
		}
	case f.userModel != "":
		writeModelJSON(filepath.Join(configDir, "settings.json"), f.userModel)
	}

	if f.defaultProf {
		t.Setenv(config.EnvClaudeConfigDir, configDir)
	} else {
		t.Setenv(config.EnvClaudeConfigDir, "")
	}
	t.Setenv(config.EnvAnthropicModel, f.envModel)

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

	if err := runUnifiedLaunch(profileName, "claude", f.args); err != nil {
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

// TestLaunchModelPrecedence pins the order of card t1441: --model > profile
// model > ANTHROPIC_MODEL > project settings.local.json > user-scope /model
// value > project pin (nothing passed, pin announced). Each row sets every
// LOWER level too, so a row only passes when its level really beats them. The
// two Claude-Code-owned levels (env, settings.local.json) are asserted as
// "nothing passed": the launcher must stay out of their way.
func TestLaunchModelPrecedence(t *testing.T) {
	cases := []struct {
		name        string
		fx          modelFixture
		wantModel   string
		wantNotes   []string
		wantNoNotes []string
	}{
		{
			name: "1 explicit --model beats everything",
			fx: modelFixture{args: []string{"--model", "haiku"}, profileModel: "sonnet[1m]",
				envModel: "fable", localModel: "sonnet", userModel: "opus", projectPin: "sonnet"},
			wantModel: "haiku", wantNoNotes: []string{"user /model", "project pin"},
		},
		{
			name: "2 profile model beats env, local, user and pin",
			fx: modelFixture{profileModel: "fable[1m]", envModel: "haiku",
				localModel: "sonnet", userModel: "opus", projectPin: "sonnet"},
			wantModel: "fable[1m]", wantNoNotes: []string{"user /model", "project pin"},
		},
		{
			name:      "3 ANTHROPIC_MODEL beats local, user and pin: nothing passed, no notice",
			fx:        modelFixture{envModel: "haiku", localModel: "sonnet", userModel: "opus", projectPin: "sonnet"},
			wantModel: "", wantNoNotes: []string{"user /model", "project pin"},
		},
		{
			name:      "4 settings.local.json beats user /model and the pin: nothing passed, no notice",
			fx:        modelFixture{localModel: "haiku", userModel: "opus", projectPin: "sonnet"},
			wantModel: "", wantNoNotes: []string{"user /model", "project pin"},
		},
		{
			name:      "5 user-scope /model beats the project pin and is announced",
			fx:        modelFixture{userModel: "opus", projectPin: "sonnet"},
			wantModel: "opus", wantNotes: []string{"model: opus (user /model)"},
			wantNoNotes: []string{"project pin"},
		},
		{
			name:      "6 nothing chosen: no --model, pin named with the ways to change it",
			fx:        modelFixture{projectPin: "sonnet"},
			wantModel: "",
			wantNotes: []string{`"model": "sonnet"`, "moai profile setup", "settings.local.json", "project pin"},
		},
		{
			name:      "7 default profile reads the user-scope settings of CLAUDE_CONFIG_DIR",
			fx:        modelFixture{defaultProf: true, userModel: "opus", projectPin: "sonnet"},
			wantModel: "opus", wantNotes: []string{"model: opus (user /model)"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model, stderr := launchModel(t, tc.fx)
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

// TestLaunchModelUserScopeFailsOpen pins fail-open: an absent, empty, blank,
// non-string, or malformed user-scope "model" never errors, never passes a
// model, and falls through to the project pin with its notice.
func TestLaunchModelUserScopeFailsOpen(t *testing.T) {
	for name, raw := range map[string]string{
		"malformed json":    "{not json",
		"empty file":        "",
		"no model key":      "{\"theme\": \"dark\"}\n",
		"empty model":       "{\"model\": \"\"}\n",
		"blank model":       "{\"model\": \"   \"}\n",
		"non-string model":  "{\"model\": 5}\n",
		"null model":        "{\"model\": null}\n",
		"top-level array":   "[\"opus\"]\n",
		"nested model only": "{\"modelSettings\": {\"model\": \"opus\"}}\n",
	} {
		t.Run(name, func(t *testing.T) {
			fx := modelFixture{projectPin: "sonnet"}
			if raw != "" {
				fx.userRaw = raw
			}
			model, stderr := launchModel(t, fx)
			if model != "" {
				t.Errorf("--model = %q, want none\n--- stderr ---\n%s", model, stderr)
			}
			if !strings.Contains(stderr, "project pin") {
				t.Errorf("fail-open must announce the project pin\n--- stderr ---\n%s", stderr)
			}
		})
	}
}

// TestResolveLaunchModelFallbackGLMExcluded pins the GLM exclusion directly: a
// GLM backend never takes the user-scope /model value (slot aliases route the
// session) and never prints the Claude-only project-pin notice.
func TestResolveLaunchModelFallbackGLMExcluded(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "settings.json"), []byte("{\"model\": \"opus\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".claude", "settings.json"), []byte("{\"model\": \"sonnet\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeConfigDir, configDir)
	t.Setenv(config.EnvAnthropicModel, "")

	var glm strings.Builder
	if got := resolveLaunchModelFallback("", true, root, "", &glm); got != "" {
		t.Errorf("glm backend took model %q, want none", got)
	}
	if strings.Contains(glm.String(), "project pin") || strings.Contains(glm.String(), "user /model") {
		t.Errorf("glm backend printed a Claude-only notice: %q", glm.String())
	}

	// Control: the same fixture on the Claude backend does take the value.
	var claude strings.Builder
	if got := resolveLaunchModelFallback("", false, root, "", &claude); got != "opus" {
		t.Errorf("claude backend model = %q, want opus (control)", got)
	}
}

// TestUserScopeSettingsPathFollowsConfigDir pins level 5's location: the
// profile-specific CLAUDE_CONFIG_DIR when set, ~/.claude otherwise.
func TestUserScopeSettingsPathFollowsConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.EnvClaudeConfigDir, dir)
	if got, want := userScopeSettingsPath(), filepath.Join(dir, "settings.json"); got != want {
		t.Errorf("with CLAUDE_CONFIG_DIR: path = %q, want %q", got, want)
	}

	home := t.TempDir()
	t.Setenv(config.EnvClaudeConfigDir, "")
	t.Setenv("HOME", home)
	if got, want := userScopeSettingsPath(), filepath.Join(home, ".claude", "settings.json"); got != want {
		t.Errorf("without CLAUDE_CONFIG_DIR: path = %q, want %q", got, want)
	}
}
