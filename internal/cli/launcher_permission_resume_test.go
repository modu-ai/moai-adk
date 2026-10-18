package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Launch scenarios that run more than once in one project, or that start from a
// pre-seeded settings.local.json (card t1866 audit repair round 1). They share
// the helpers below and permissionModeArg from launcher_permission_default_test.go.

// permissionProject creates an isolated MoAI project (chdir'd into) with an
// isolated profile store and a stubbed claude binary.
func permissionProject(t *testing.T) string {
	t.Helper()
	root := fakeMoaiProject(t)
	isolateProfileStore(t)
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-stub")))
	return root
}

// setProjectBackend writes llm.yaml so the project resolves to the GLM backend
// (glm true) or removes it so the project resolves to the Claude backend.
func setProjectBackend(t *testing.T, root string, glm bool) {
	t.Helper()
	sections := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	path := filepath.Join(sections, "llm.yaml")
	if !glm {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove llm.yaml: %v", err)
		}
		return
	}
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(path, []byte("llm:\n  team_mode: glm\n"), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}

// launchInProject launches through the exec seam in the current project and
// returns the child args.
func launchInProject(t *testing.T, extraArgs []string) []string {
	t.Helper()
	var args []string
	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(bin string, a, env []string) error {
		args = a
		return nil
	}
	if err := launchClaudeDefault("", extraArgs); err != nil {
		t.Fatalf("launch: %v", err)
	}
	return args
}

func settingsLocalPathFor(root string) string {
	return filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)
}

// readSettingsDoc returns settings.local.json as a map (empty when absent).
func readSettingsDoc(t *testing.T, root string) map[string]any {
	t.Helper()
	doc := map[string]any{}
	raw, err := os.ReadFile(settingsLocalPathFor(root))
	if err != nil {
		return doc
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("settings.local.json is not JSON: %v", err)
	}
	return doc
}

func settingsDefaultMode(doc map[string]any) string {
	perms, _ := doc["permissions"].(map[string]any)
	mode, _ := perms["defaultMode"].(string)
	return mode
}

// writeSettingsDoc replaces settings.local.json with doc.
func writeSettingsDoc(t *testing.T, root string, doc map[string]any) {
	t.Helper()
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(settingsLocalPathFor(root)), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.WriteFile(settingsLocalPathFor(root), raw, 0o644); err != nil {
		t.Fatalf("write settings.local.json: %v", err)
	}
}

// setSettingsEnv sets one env key in settings.local.json, keeping the rest.
func setSettingsEnv(t *testing.T, root, key, value string) {
	t.Helper()
	doc := readSettingsDoc(t, root)
	env, _ := doc["env"].(map[string]any)
	if env == nil {
		env = map[string]any{}
	}
	env[key] = value
	doc["env"] = env
	writeSettingsDoc(t, root, doc)
}

// TestLaunch_GLMRemovesStaleAutoFromSettingsLocal pins that a launcher-written
// permissions.defaultMode: auto does not survive into a GLM launch, resume or
// not: the resume path skips the sync step, so it needs its own removal.
func TestLaunch_GLMRemovesStaleAutoFromSettingsLocal(t *testing.T) {
	cases := []struct {
		name      string
		extraArgs []string
		envCont   bool
	}{
		{name: "fresh GLM launch"},
		{name: "GLM launch with --continue", extraArgs: []string{"--continue"}},
		{name: "GLM launch with -c", extraArgs: []string{"-c"}},
		{name: "GLM launch with DO_CLAUDE_CONTINUE=true", envCont: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := permissionProject(t)

			// An earlier Claude launch with an unset profile writes auto.
			launchInProject(t, nil)
			if got := settingsDefaultMode(readSettingsDoc(t, root)); got != "auto" {
				t.Fatalf("precondition: Claude launch left defaultMode %q, want auto", got)
			}

			setProjectBackend(t, root, true)
			if tc.envCont {
				setSettingsEnv(t, root, "DO_CLAUDE_CONTINUE", "true")
			}
			args := launchInProject(t, tc.extraArgs)

			if got := permissionModeArg(args); got != "" {
				t.Errorf("--permission-mode = %q in args %v, want none on the GLM backend", got, args)
			}
			doc := readSettingsDoc(t, root)
			if got := settingsDefaultMode(doc); got == "auto" {
				t.Errorf("settings.local.json still carries permissions.defaultMode auto after a GLM launch: %v", doc)
			}
		})
	}
}

// TestLaunch_GLMResumeKeepsOtherDefaultModes: only a value equal to auto is
// removed on a GLM resume; any other defaultMode is the user's and stays.
func TestLaunch_GLMResumeKeepsOtherDefaultModes(t *testing.T) {
	root := permissionProject(t)
	setProjectBackend(t, root, true)
	writeSettingsDoc(t, root, map[string]any{"permissions": map[string]any{"defaultMode": "plan"}})

	launchInProject(t, []string{"--continue"})

	if got := settingsDefaultMode(readSettingsDoc(t, root)); got != "plan" {
		t.Errorf("defaultMode = %q after a GLM resume, want plan untouched", got)
	}
}

// TestLaunch_ClaudeResumeDoesNotRewriteSettingsLocal: on the Claude backend a
// resume leaves a hand-edited defaultMode alone (the sync runs on fresh launches
// only), even though an unset profile would default to auto, and adds no default
// --permission-mode flag to the child (the auto default is for new sessions only,
// card t1866 operator exception b18d). The child is the pinned stub, which a
// --continue launch starts through exec.Command(...).Run().
func TestLaunch_ClaudeResumeDoesNotRewriteSettingsLocal(t *testing.T) {
	for _, hand := range []string{"plan", "auto"} {
		t.Run(hand, func(t *testing.T) {
			root := permissionProject(t)
			writeSettingsDoc(t, root, map[string]any{"permissions": map[string]any{"defaultMode": hand}})
			argsFile := pinArgsRecordingClaude(t)

			if err := launchClaudeDefault("", []string{"--continue"}); err != nil {
				t.Fatalf("launch: %v", err)
			}

			if got := settingsDefaultMode(readSettingsDoc(t, root)); got != hand {
				t.Errorf("defaultMode = %q after a Claude resume, want the hand-edited %q", got, hand)
			}
			if args := recordedChildArgs(t, argsFile); hasArg(args, "--permission-mode") {
				t.Errorf("child args %v carry --permission-mode on a resume with an unset profile, want none", args)
			}
		})
	}
}

// pinArgsRecordingClaude pins a stub claude binary that writes its arguments, one
// per line, to a file in its own temp directory (never the project's .claude
// directory) and exits 0. It returns the file path; the file exists only if the
// binary ran.
func pinArgsRecordingClaude(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "claude-args")
	bin := filepath.Join(dir, "claude-args-stub")
	script := "#!/bin/sh\n: >> '" + argsFile + "'\nfor a in \"$@\"; do printf '%s\\n' \"$a\" >> '" + argsFile + "'; done\nexit 0\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write args stub: %v", err)
	}
	t.Setenv(config.EnvClaudeBin, bin)
	return argsFile
}

// recordedChildArgs returns the arguments the args-recording stub saw. It fails
// the test when the stub never ran, so an absent flag is never read from a child
// that did not start.
func recordedChildArgs(t *testing.T, argsFile string) []string {
	t.Helper()
	raw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("the pinned claude binary never ran (args file %s): %v", argsFile, err)
	}
	return strings.Fields(string(raw))
}

func hasArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

// TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag: the auto default
// for a profile that never chose a mode applies to NEW sessions only. A resume
// (typed --continue / -c, or DO_CLAUDE_CONTINUE=true in settings.local.json) on the
// Claude backend passes no --permission-mode, so the child follows
// permissions.defaultMode in settings.local.json (card t1866, operator exception b18d).
func TestLaunch_ClaudeResumeUnsetProfileAddsNoDefaultPermissionFlag(t *testing.T) {
	cases := []struct {
		name      string
		extraArgs []string
		envCont   bool
	}{
		{name: "--continue", extraArgs: []string{"--continue"}},
		{name: "-c", extraArgs: []string{"-c"}},
		{name: "DO_CLAUDE_CONTINUE=true", envCont: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := permissionProject(t)
			if tc.envCont {
				setSettingsEnv(t, root, "DO_CLAUDE_CONTINUE", "true")
			}
			argsFile := pinArgsRecordingClaude(t)

			if err := launchClaudeDefault("", tc.extraArgs); err != nil {
				t.Fatalf("launch: %v", err)
			}

			args := recordedChildArgs(t, argsFile)
			if !hasArg(args, "--continue") {
				t.Fatalf("child args %v lack --continue: this was not a resume", args)
			}
			if hasArg(args, "--permission-mode") {
				t.Errorf("child args %v carry --permission-mode on a resume with an unset profile, want none", args)
			}
		})
	}
}

// TestLaunch_ClaudeResumeUnsetProfileFollowsHandWrittenDefaultMode: a hand-edited
// permissions.defaultMode: plan survives a Claude resume with an unset profile
// byte for byte, and no --permission-mode flag overrides it at run time.
func TestLaunch_ClaudeResumeUnsetProfileFollowsHandWrittenDefaultMode(t *testing.T) {
	root := permissionProject(t)
	writeSettingsDoc(t, root, map[string]any{"permissions": map[string]any{"defaultMode": "plan"}})
	before, err := os.ReadFile(settingsLocalPathFor(root))
	if err != nil {
		t.Fatalf("read settings.local.json: %v", err)
	}
	argsFile := pinArgsRecordingClaude(t)

	if err := launchClaudeDefault("", []string{"--continue"}); err != nil {
		t.Fatalf("launch: %v", err)
	}

	if args := recordedChildArgs(t, argsFile); hasArg(args, "--permission-mode") {
		t.Errorf("child args %v carry --permission-mode, which would override the hand-written plan", args)
	}
	after, err := os.ReadFile(settingsLocalPathFor(root))
	if err != nil {
		t.Fatalf("read settings.local.json after launch: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("settings.local.json changed: before %q, after %q", before, after)
	}
}

// TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode: only the unset-profile
// DEFAULT is skipped on a resume. A stored profile mode (auto included) and a
// typed --permission-mode are still forwarded to the child.
func TestLaunch_ClaudeResumeStillForwardsChosenPermissionMode(t *testing.T) {
	cases := []struct {
		name      string
		stored    string
		extraArgs []string
		want      string
	}{
		{name: "stored auto profile", stored: "auto", extraArgs: []string{"--continue"}, want: "auto"},
		{name: "typed plan", extraArgs: []string{"--permission-mode", "plan", "--continue"}, want: "plan"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			permissionProject(t)
			if tc.stored != "" {
				if err := profile.WritePreferences("default", profile.ProfilePreferences{PermissionMode: tc.stored}); err != nil {
					t.Fatalf("seed profile: %v", err)
				}
			}
			argsFile := pinArgsRecordingClaude(t)

			if err := launchClaudeDefault("", tc.extraArgs); err != nil {
				t.Fatalf("launch: %v", err)
			}

			args := recordedChildArgs(t, argsFile)
			if got := permissionModeArg(args); got != tc.want {
				t.Errorf("--permission-mode = %q in child args %v, want %q forwarded on the resume", got, args, tc.want)
			}
		})
	}
}

// TestLaunch_LegacyBypassWithEmptyProfileModeKeepsBypassAndWritesNoOverride:
// DO_CLAUDE_BYPASS=true with an empty profile mode launches bypassPermissions and
// leaves no permissions.defaultMode override (the default auto does not apply).
func TestLaunch_LegacyBypassWithEmptyProfileModeKeepsBypassAndWritesNoOverride(t *testing.T) {
	root := permissionProject(t)
	writeSettingsDoc(t, root, map[string]any{
		"env":         map[string]any{"DO_CLAUDE_BYPASS": "true"},
		"permissions": map[string]any{"defaultMode": "auto"},
	})

	args := launchInProject(t, nil)

	if got := permissionModeArg(args); got != "bypassPermissions" {
		t.Errorf("--permission-mode = %q in args %v, want bypassPermissions", got, args)
	}
	if got := settingsDefaultMode(readSettingsDoc(t, root)); got != "" {
		t.Errorf("permissions.defaultMode = %q, want the override removed", got)
	}
}

// TestLaunch_ExplicitPermissionModeAutoOnCommandLineIsForwardedOnClaude: a typed
// `--permission-mode auto` reaches the child unchanged, once, on the Claude backend.
func TestLaunch_ExplicitPermissionModeAutoOnCommandLineIsForwardedOnClaude(t *testing.T) {
	permissionProject(t)

	args := launchInProject(t, []string{"--permission-mode", "auto"})

	if got := permissionModeArg(args); got != "auto" {
		t.Errorf("--permission-mode = %q in args %v, want auto forwarded unchanged", got, args)
	}
	count := 0
	for _, a := range args {
		if a == "--permission-mode" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("--permission-mode appears %d times in %v, want once", count, args)
	}
}

// TestLaunch_SettingsChosenModeDoesNotPersistDefaultAuto: a profile with no mode
// but a DO_CLAUDE_PERMISSION_MODE in settings.local.json runs that mode and, as
// before t1866, removes the defaultMode override rather than persisting auto.
func TestLaunch_SettingsChosenModeDoesNotPersistDefaultAuto(t *testing.T) {
	root := permissionProject(t)
	writeSettingsDoc(t, root, map[string]any{
		"env":         map[string]any{"DO_CLAUDE_PERMISSION_MODE": "plan"},
		"permissions": map[string]any{"defaultMode": "auto"},
	})

	args := launchInProject(t, nil)

	if got := permissionModeArg(args); got != "plan" {
		t.Errorf("--permission-mode = %q in args %v, want plan", got, args)
	}
	if got := settingsDefaultMode(readSettingsDoc(t, root)); got != "" {
		t.Errorf("permissions.defaultMode = %q, want the override removed (a settings-chosen mode is not the launcher default)", got)
	}
}

// restoreEnvAfter snapshots the process environment and restores it at cleanup:
// applyGLMMode exports the GLM credentials into the test process.
func restoreEnvAfter(t *testing.T) {
	t.Helper()
	snapshot := os.Environ()
	t.Cleanup(func() {
		os.Clearenv()
		for _, kv := range snapshot {
			if k, v, ok := strings.Cut(kv, "="); ok {
				_ = os.Setenv(k, v)
			}
		}
	})
}

// TestLaunch_GLMLaunchWithUnwritableConfigDoesNotReceiveAuto: when persisting
// team_mode fails, applyGLMMode has already exported the GLM environment but
// llm.yaml still reads as the Claude backend. The command-selected provider, not
// the disk read, must decide that this is a GLM launch. A read-only sections
// directory (llm.yaml absent) makes persistTeamMode fail for a non-root user.
func TestLaunch_GLMLaunchWithUnwritableConfigDoesNotReceiveAuto(t *testing.T) {
	root := permissionProject(t)
	restoreEnvAfter(t)
	t.Setenv("GLM_API_KEY", "test-key")
	t.Setenv(config.EnvGLMNoAutoTools, "1")
	t.Setenv("TMUX", "")

	sections := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.Chmod(sections, 0o555); err != nil {
		t.Fatalf("chmod sections read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(sections, 0o755) })
	if err := persistTeamMode(root, "glm"); err == nil {
		t.Fatal("precondition: persistTeamMode succeeded in a read-only config directory; this environment cannot reproduce the failed save")
	}

	var args []string
	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(bin string, a, env []string) error {
		args = a
		return nil
	}
	if err := runUnifiedLaunch("", "glm", nil); err != nil {
		t.Fatalf("glm launch: %v", err)
	}

	if len(args) == 0 {
		t.Fatal("the launch never reached the exec seam")
	}
	if got := permissionModeArg(args); got == "auto" {
		t.Errorf("--permission-mode auto reached a GLM launch whose team_mode save failed: %v", args)
	}
	doc := readSettingsDoc(t, root)
	if got := settingsDefaultMode(doc); got == "auto" {
		t.Errorf("settings.local.json carries defaultMode auto after a GLM launch: %v", doc)
	}
}

// TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved: a GLM resume whose stale
// permissions.defaultMode: auto cannot be removed (read-only .claude directory)
// must not start the session. The launch returns an error naming the file, the
// exec seam is never reached, and the file is byte-unchanged (card t1866 F7).
func TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved(t *testing.T) {
	root := permissionProject(t)
	setProjectBackend(t, root, true)
	writeSettingsDoc(t, root, map[string]any{"permissions": map[string]any{"defaultMode": "auto"}})

	path := settingsLocalPathFor(root)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings.local.json: %v", err)
	}

	dir := filepath.Dir(path)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("chmod .claude read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	// The induction must take effect: a probe write into the directory has to fail
	// (it does not for root, which ignores directory permissions).
	probe := filepath.Join(dir, "induction-probe")
	if f, err := os.OpenFile(probe, os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
		_ = f.Close()
		_ = os.Remove(probe)
		t.Fatal("precondition: a file could be created in the read-only .claude directory; this environment cannot reproduce the failed removal")
	}

	// A --continue resume starts its first child through exec.Command(...).Run(),
	// not through the exec seam, so the seam alone cannot see a started session:
	// the pinned stub claude binary records its own start in a marker file.
	marker := pinMarkerClaude(t)

	called := false
	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(bin string, a, env []string) error {
		called = true
		return nil
	}

	err = launchClaudeDefault("", []string{"--continue"})
	if err == nil {
		t.Error("GLM resume returned nil although the stale auto could not be removed")
	} else if !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "auto") {
		t.Errorf("error %q must name %s and say the stale auto could not be removed", err, path)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Error("the session was started (the claude binary ran) with a stale auto still on disk")
	} else if !os.IsNotExist(statErr) {
		t.Fatalf("stat marker: %v", statErr)
	}
	if called {
		t.Error("the session was started (exec seam called) with a stale auto still on disk")
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read settings.local.json after launch: %v", readErr)
	}
	if string(after) != string(before) {
		t.Errorf("settings.local.json changed: before %q, after %q", before, after)
	}
}

// pinMarkerClaude pins a stub claude binary that touches a marker file when it
// is started and exits 0. The marker lives in its own temp directory (never in
// the project's .claude directory, which a test may make read-only). It returns
// the marker path; the file exists only if the binary ran.
func pinMarkerClaude(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "claude-started")
	bin := filepath.Join(dir, "claude-marker-stub")
	script := "#!/bin/sh\ntouch '" + marker + "'\nexit 0\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write marker stub: %v", err)
	}
	t.Setenv(config.EnvClaudeBin, bin)
	return marker
}

// TestLaunch_GLMResumeWithoutAutoStartsTheClaudeBinary is the positive control
// for TestLaunch_GLMResumeFailsWhenStaleAutoCannotBeRemoved: a GLM resume whose
// settings.local.json carries no auto (a different defaultMode, or no file at
// all) does start the pinned stub, so the marker file exists. It proves the
// marker mechanism observes a child start on the --continue path.
func TestLaunch_GLMResumeWithoutAutoStartsTheClaudeBinary(t *testing.T) {
	cases := []struct {
		name string
		seed func(t *testing.T, root string)
	}{
		{"defaultMode plan", func(t *testing.T, root string) {
			writeSettingsDoc(t, root, map[string]any{"permissions": map[string]any{"defaultMode": "plan"}})
		}},
		{"no settings.local.json", func(t *testing.T, root string) {
			if err := os.Remove(settingsLocalPathFor(root)); err != nil && !os.IsNotExist(err) {
				t.Fatalf("remove settings.local.json: %v", err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := permissionProject(t)
			setProjectBackend(t, root, true)
			tc.seed(t, root)
			marker := pinMarkerClaude(t)

			origExec := execOrSpawnClaudeFunc
			t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
			execOrSpawnClaudeFunc = func(bin string, a, env []string) error { return nil }

			if err := launchClaudeDefault("", []string{"--continue"}); err != nil {
				t.Fatalf("GLM resume without auto: %v", err)
			}
			if _, err := os.Stat(marker); err != nil {
				t.Errorf("the resume did not start the pinned claude binary (marker %s): %v", marker, err)
			}
		})
	}
}
