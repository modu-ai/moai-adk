package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// launchWithStoredPermission launches through the exec seam with a project whose
// llm.yaml optionally selects the GLM backend, and returns the child args and
// the permissions.defaultMode the launch left in settings.local.json ("" when
// the key is absent). The profile store is the TestMain-isolated one, so an
// unconfigured profile reads as an unset permission_mode.
func launchWithStoredPermission(t *testing.T, glm bool) (args []string, defaultMode string) {
	t.Helper()
	return launchWithProfilePermission(t, glm, "")
}

// launchWithProfilePermission is launchWithStoredPermission with the profile's
// stored permission_mode set to stored ("" leaves the profile unconfigured).
func launchWithProfilePermission(t *testing.T, glm bool, stored string) (args []string, defaultMode string) {
	t.Helper()
	root := fakeMoaiProject(t)
	isolateProfileStore(t)
	if stored != "" {
		if err := profile.WritePreferences("default", profile.ProfilePreferences{PermissionMode: stored}); err != nil {
			t.Fatalf("seed profile: %v", err)
		}
	}
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-stub")))

	if glm {
		sections := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatalf("mkdir sections: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("llm:\n  team_mode: glm\n"), 0o644); err != nil {
			t.Fatalf("write llm.yaml: %v", err)
		}
	}

	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(bin string, a, env []string) error {
		args = a
		return nil
	}
	if err := launchClaudeDefault("", nil); err != nil {
		t.Fatalf("launch: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON))
	if err != nil {
		return args, ""
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("settings.local.json is not JSON: %v", err)
	}
	if perms, ok := doc["permissions"].(map[string]any); ok {
		defaultMode, _ = perms["defaultMode"].(string)
	}
	return args, defaultMode
}

func permissionModeArg(args []string) string {
	for i, a := range args {
		if a == "--permission-mode" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// TestLaunch_UnsetPermissionModeDefaultsToAutoOnClaudeBackend pins the shipped
// default (card t1866, operator decision 2026-10-18): a profile that never chose
// a permission mode launches Claude sessions in auto, on the command line and in
// the settings.local.json override.
func TestLaunch_UnsetPermissionModeDefaultsToAutoOnClaudeBackend(t *testing.T) {
	args, defaultMode := launchWithStoredPermission(t, false)
	if got := permissionModeArg(args); got != "auto" {
		t.Errorf("--permission-mode = %q in args %v, want auto for an unset profile on the Claude backend", got, args)
	}
	if defaultMode != "auto" {
		t.Errorf("settings.local.json permissions.defaultMode = %q, want auto", defaultMode)
	}
}

// TestLaunch_UnsetPermissionModeStaysEmptyOnGLMBackend pins the carve-out: the GLM
// launcher refuses `--permission-mode auto` (glm.go), so an unset profile on the
// GLM backend must receive no permission flag and no override.
func TestLaunch_UnsetPermissionModeStaysEmptyOnGLMBackend(t *testing.T) {
	args, defaultMode := launchWithStoredPermission(t, true)
	if got := permissionModeArg(args); got != "" {
		t.Errorf("--permission-mode = %q in args %v, want none for an unset profile on the GLM backend", got, args)
	}
	if defaultMode != "" {
		t.Errorf("settings.local.json permissions.defaultMode = %q, want absent on the GLM backend", defaultMode)
	}
}

// TestLaunch_StoredAutoIsForwardedOnClaudeBackend: a profile that explicitly chose
// auto (the wizard now pre-selects it) keeps passing the flag on the Claude backend.
func TestLaunch_StoredAutoIsForwardedOnClaudeBackend(t *testing.T) {
	args, defaultMode := launchWithProfilePermission(t, false, "auto")
	if got := permissionModeArg(args); got != "auto" {
		t.Errorf("--permission-mode = %q in args %v, want auto for a stored auto on the Claude backend", got, args)
	}
	if defaultMode != "auto" {
		t.Errorf("settings.local.json permissions.defaultMode = %q, want auto", defaultMode)
	}
}

// TestLaunch_StoredAutoIsNotForwardedOnGLMBackend: the GLM launcher refuses
// `--permission-mode auto` (glm.go checks argv only), so a stored auto must not
// reach the child as a flag or as a settings.local.json override on the GLM backend.
func TestLaunch_StoredAutoIsNotForwardedOnGLMBackend(t *testing.T) {
	args, defaultMode := launchWithProfilePermission(t, true, "auto")
	if got := permissionModeArg(args); got != "" {
		t.Errorf("--permission-mode = %q in args %v, want none for a stored auto on the GLM backend", got, args)
	}
	if defaultMode != "" {
		t.Errorf("settings.local.json permissions.defaultMode = %q, want absent on the GLM backend", defaultMode)
	}
}
