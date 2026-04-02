package template

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

// renderTemplate loads a template from embedded FS and renders it with the given context.
func renderTemplate(t *testing.T, tmplPath string, ctx *TemplateContext) string {
	t.Helper()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, tmplPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", tmplPath, err)
	}

	tmpl, err := template.New(tmplPath).Funcs(templateFuncMap).Parse(string(data))
	if err != nil {
		t.Fatalf("Parse template error: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, ctx); err != nil {
		t.Fatalf("Execute template error: %v", err)
	}

	return buf.String()
}

func testContext(platform string) *TemplateContext {
	return NewTemplateContext(
		WithPlatform(platform),
		WithSmartPATH("/usr/local/bin:/usr/bin"),
		WithGoBinPath("/usr/local/go/bin"),
		WithHomeDir("/home/test"),
	)
}

// --- settings.json.tmpl tests ---

func TestSettingsTemplateValidJSON(t *testing.T) {
	platforms := []string{"darwin", "linux", "windows"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			ctx := testContext(platform)
			output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

			trimmed := strings.TrimSpace(output)
			if !json.Valid([]byte(trimmed)) {
				t.Fatalf("rendered settings.json is not valid JSON for platform %s:\n%s", platform, trimmed)
			}
		})
	}
}

func TestSettingsTemplateRequiredHooks(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("missing hooks section")
	}

	requiredEvents := []string{"SessionStart", "PreCompact", "SessionEnd", "PreToolUse", "PostToolUse", "Stop"}
	for _, event := range requiredEvents {
		if _, ok := hooks[event]; !ok {
			t.Errorf("missing required hook event %q", event)
		}
	}
}

func TestSettingsTemplateRequiredEnvVars(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	env, ok := settings["env"].(map[string]any)
	if !ok {
		t.Fatal("missing env section")
	}

	requiredKeys := []string{
		"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS",
		"CLAUDE_CODE_FILE_READ_MAX_OUTPUT_TOKENS",
		"ENABLE_TOOL_SEARCH",
		"MOAI_CONFIG_SOURCE",
		"PATH",
	}
	for _, key := range requiredKeys {
		if _, ok := env[key]; !ok {
			t.Errorf("missing required env var %q", key)
		}
	}
}

func TestSettingsTemplatePlatformHookCommands(t *testing.T) {
	t.Run("darwin_uses_direct_command", func(t *testing.T) {
		ctx := testContext("darwin")
		output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

		// Darwin should NOT use bash prefix (check raw JSON with escaped quotes)
		if strings.Contains(output, `bash \"$CLAUDE_PROJECT_DIR`) {
			t.Error("darwin should not use bash prefix for hook commands")
		}
		if !strings.Contains(output, `\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"`) {
			t.Error("darwin should have direct path to hook script")
		}
	})

	t.Run("windows_uses_bash_prefix", func(t *testing.T) {
		ctx := testContext("windows")
		output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

		if !strings.Contains(output, `bash \"$CLAUDE_PROJECT_DIR`) {
			t.Error("windows should use bash prefix for hook commands")
		}
	})
}

func TestSettingsTemplateOutputStyle(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	if !strings.Contains(output, `"outputStyle": "MoAI"`) {
		t.Error("settings should contain outputStyle: MoAI")
	}
}

func TestSettingsTemplateNoSandbox(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := settings["sandbox"]; ok {
		t.Error("template should not include sandbox section (causes test failures across language ecosystems)")
	}
}

func TestSettingsTemplateAttribution(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	attr, ok := settings["attribution"].(map[string]any)
	if !ok {
		t.Fatal("missing attribution section")
	}
	commit, _ := attr["commit"].(string)
	if !strings.Contains(commit, "MoAI") {
		t.Errorf("attribution.commit should contain MoAI, got %q", commit)
	}
	pr, _ := attr["pr"].(string)
	if !strings.Contains(pr, "MoAI") {
		t.Errorf("attribution.pr should contain MoAI, got %q", pr)
	}
}

func TestSettingsTemplateStatusLine(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	sl, ok := settings["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("missing statusLine section")
	}
	if sl["type"] != "command" {
		t.Errorf("statusLine.type = %v, want %q", sl["type"], "command")
	}
	if sl["command"] != ".moai/status_line.sh" {
		t.Errorf("statusLine.command = %v, want %q", sl["command"], ".moai/status_line.sh")
	}
}

func TestSettingsTemplatePermissions(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Fatal("missing permissions section")
	}

	if perms["defaultMode"] != "acceptEdits" {
		t.Errorf("permissions.defaultMode = %v, want %q", perms["defaultMode"], "acceptEdits")
	}

	allow, ok := perms["allow"].([]any)
	if !ok {
		t.Fatal("permissions.allow is not an array")
	}
	if len(allow) < 50 {
		t.Errorf("expected at least 50 allow entries, got %d", len(allow))
	}

	deny, ok := perms["deny"].([]any)
	if !ok {
		t.Fatal("permissions.deny is not an array")
	}
	if len(deny) < 10 {
		t.Errorf("expected at least 10 deny entries, got %d", len(deny))
	}
}

func TestSettingsTemplateCleanupPeriod(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	val, ok := settings["cleanupPeriodDays"]
	if !ok {
		t.Fatal("missing cleanupPeriodDays")
	}
	if val != float64(30) {
		t.Errorf("cleanupPeriodDays = %v, want 30", val)
	}
}

func TestSettingsTemplateNewFields(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify new boolean fields
	boolFields := map[string]bool{
		"enableAllProjectMcpServers": true,
		"respectGitignore":           true,
	}
	for field, want := range boolFields {
		val, ok := settings[field]
		if !ok {
			t.Errorf("missing field %q", field)
			continue
		}
		if val != want {
			t.Errorf("%s = %v, want %v", field, val, want)
		}
	}
}

func TestSettingsTemplateAllHookEvents(t *testing.T) {
	t.Parallel()

	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("missing hooks section")
	}

	allEvents := []string{
		"SessionStart", "PreCompact", "SessionEnd",
		"PreToolUse", "PostToolUse", "Stop",
		"SubagentStop", "PostToolUseFailure", "Notification",
		"SubagentStart", "UserPromptSubmit", "PermissionRequest",
		"TeammateIdle", "TaskCompleted",
	}
	for _, event := range allEvents {
		if _, ok := hooks[event]; !ok {
			t.Errorf("missing hook event %q", event)
		}
	}
}

func TestSettingsTemplateNewHookStructure(t *testing.T) {
	t.Parallel()

	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	hooksSection, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("missing hooks section")
	}

	newEvents := []struct {
		event      string
		scriptName string
	}{
		{"SubagentStop", "handle-subagent-stop.sh"},
		{"PostToolUseFailure", "handle-post-tool-failure.sh"},
		{"Notification", "handle-notification.sh"},
		{"SubagentStart", "handle-subagent-start.sh"},
		{"UserPromptSubmit", "handle-user-prompt-submit.sh"},
		{"PermissionRequest", "handle-permission-request.sh"},
		{"TeammateIdle", "handle-teammate-idle.sh"},
		{"TaskCompleted", "handle-task-completed.sh"},
	}

	for _, ne := range newEvents {
		t.Run(ne.event, func(t *testing.T) {
			t.Parallel()

			eventData, ok := hooksSection[ne.event]
			if !ok {
				t.Fatalf("missing hook event %q", ne.event)
			}

			// Hook config must be an array (Claude Code expects array of hook groups)
			hookGroups, ok := eventData.([]any)
			if !ok {
				t.Fatalf("%q: expected array of hook groups, got %T", ne.event, eventData)
			}
			if len(hookGroups) == 0 {
				t.Fatalf("%q: hook groups array is empty", ne.event)
			}

			// Each hook group must have a "hooks" array
			group, ok := hookGroups[0].(map[string]any)
			if !ok {
				t.Fatalf("%q: hook group is not an object, got %T", ne.event, hookGroups[0])
			}

			hooksArr, ok := group["hooks"].([]any)
			if !ok {
				t.Fatalf("%q: missing or invalid 'hooks' array in hook group", ne.event)
			}
			if len(hooksArr) == 0 {
				t.Fatalf("%q: hooks array is empty", ne.event)
			}

			// Each hook entry must have command, timeout, and type fields
			hookEntry, ok := hooksArr[0].(map[string]any)
			if !ok {
				t.Fatalf("%q: hook entry is not an object, got %T", ne.event, hooksArr[0])
			}

			// Verify command contains the correct shell script name
			command, ok := hookEntry["command"].(string)
			if !ok {
				t.Fatalf("%q: missing or non-string 'command' field", ne.event)
			}
			if !strings.Contains(command, ne.scriptName) {
				t.Errorf("%q: command %q does not contain expected script name %q", ne.event, command, ne.scriptName)
			}

			// Verify timeout is 5 (number)
			timeout, ok := hookEntry["timeout"]
			if !ok {
				t.Fatalf("%q: missing 'timeout' field", ne.event)
			}
			if timeout != float64(5) {
				t.Errorf("%q: timeout = %v, want 5", ne.event, timeout)
			}

			// Verify type is "command"
			hookType, ok := hookEntry["type"].(string)
			if !ok {
				t.Fatalf("%q: missing or non-string 'type' field", ne.event)
			}
			if hookType != "command" {
				t.Errorf("%q: type = %q, want %q", ne.event, hookType, "command")
			}
		})
	}
}

func TestSettingsTemplateNewHooksPlatformCompatibility(t *testing.T) {
	t.Parallel()

	newEvents := []struct {
		event      string
		scriptName string
	}{
		{"SubagentStop", "handle-subagent-stop.sh"},
		{"PostToolUseFailure", "handle-post-tool-failure.sh"},
		{"Notification", "handle-notification.sh"},
		{"SubagentStart", "handle-subagent-start.sh"},
		{"UserPromptSubmit", "handle-user-prompt-submit.sh"},
		{"PermissionRequest", "handle-permission-request.sh"},
		{"TeammateIdle", "handle-teammate-idle.sh"},
		{"TaskCompleted", "handle-task-completed.sh"},
	}

	platforms := []string{"darwin", "linux", "windows"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			t.Parallel()

			ctx := testContext(platform)
			output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

			var settings map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			hooksSection, ok := settings["hooks"].(map[string]any)
			if !ok {
				t.Fatal("missing hooks section")
			}

			for _, ne := range newEvents {
				t.Run(ne.event, func(t *testing.T) {
					eventData, ok := hooksSection[ne.event]
					if !ok {
						t.Fatalf("missing hook event %q", ne.event)
					}

					hookGroups := eventData.([]any)
					group := hookGroups[0].(map[string]any)
					hooksArr := group["hooks"].([]any)
					hookEntry := hooksArr[0].(map[string]any)
					command := hookEntry["command"].(string)

					switch platform {
					case "darwin", "linux":
						expected := `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/` + ne.scriptName + `"`
						if !strings.Contains(command, expected) {
							t.Errorf("%s/%s: command %q does not contain expected path %q", platform, ne.event, command, expected)
						}
						// Should NOT have bash prefix
						if strings.HasPrefix(command, "bash ") {
							t.Errorf("%s/%s: command should not have bash prefix, got %q", platform, ne.event, command)
						}
					case "windows":
						if !strings.HasPrefix(command, "bash ") {
							t.Errorf("windows/%s: command should have bash prefix, got %q", ne.event, command)
						}
						if !strings.Contains(command, ne.scriptName) {
							t.Errorf("windows/%s: command %q does not contain script name %q", ne.event, command, ne.scriptName)
						}
					}
				})
			}
		})
	}
}

func TestSettingsTemplateHookEventCount(t *testing.T) {
	t.Parallel()

	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("missing hooks section")
	}

	const expectedCount = 26 // v2.1.89: PermissionDenied added (25 → 26)
	if len(hooks) != expectedCount {
		t.Errorf("hook event count = %d, want %d; events: %v", len(hooks), expectedCount, hookKeys(hooks))
	}
}

// hookKeys returns sorted keys from a map for diagnostic output.
func hookKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- .mcp.json.tmpl tests ---

func TestMCPTemplateValidJSON(t *testing.T) {
	platforms := []string{"darwin", "linux", "windows"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			ctx := testContext(platform)
			output := renderTemplate(t, ".mcp.json.tmpl", ctx)

			trimmed := strings.TrimSpace(output)
			if !json.Valid([]byte(trimmed)) {
				t.Fatalf("rendered .mcp.json is not valid JSON for platform %s:\n%s", platform, trimmed)
			}
		})
	}
}

func TestMCPTemplateRequiredServers(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".mcp.json.tmpl", ctx)

	var config map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &config); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("missing mcpServers section")
	}

	requiredServers := []string{"context7", "sequential-thinking"}
	for _, name := range requiredServers {
		if _, ok := servers[name]; !ok {
			t.Errorf("missing required MCP server %q", name)
		}
	}
}

func TestMCPTemplatePlatformCommands(t *testing.T) {
	t.Run("darwin_uses_bash", func(t *testing.T) {
		ctx := testContext("darwin")
		output := renderTemplate(t, ".mcp.json.tmpl", ctx)

		if !strings.Contains(output, "/bin/bash") {
			t.Error("darwin should use /bin/bash")
		}
	})

	t.Run("windows_uses_cmd", func(t *testing.T) {
		ctx := testContext("windows")
		output := renderTemplate(t, ".mcp.json.tmpl", ctx)

		if !strings.Contains(output, "cmd.exe") {
			t.Error("windows should use cmd.exe")
		}
	})
}

func TestMCPTemplateStaggeredStartup(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".mcp.json.tmpl", ctx)

	var config map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &config); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	startup, ok := config["staggeredStartup"].(map[string]any)
	if !ok {
		t.Fatal("missing staggeredStartup section")
	}
	if startup["enabled"] != true {
		t.Error("staggeredStartup.enabled should be true")
	}
	if startup["delayMs"] != float64(500) {
		t.Errorf("staggeredStartup.delayMs = %v, want 500", startup["delayMs"])
	}
	if startup["connectionTimeout"] != float64(15000) {
		t.Errorf("staggeredStartup.connectionTimeout = %v, want 15000", startup["connectionTimeout"])
	}
}

func TestMCPTemplateSchema(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".mcp.json.tmpl", ctx)

	var config map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &config); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expected := "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json"
	if config["$schema"] != expected {
		t.Errorf("$schema = %v, want %q", config["$schema"], expected)
	}
}

// --- BuildSmartPATH and PathContainsDir tests ---

// TestBuildSmartPATH_StableAcrossTerminalPATH is a regression test for issue #467:
// "moai update hardcodes Linux PATH into settings.json, breaking MCP servers on macOS".
// BuildSmartPATH must NOT capture the terminal PATH; it must produce a stable,
// platform-appropriate result regardless of what is currently in the PATH env var.
//
// Note: This invariant applies to non-WSL2 environments only. In WSL2, BuildSmartPATH
// intentionally captures /mnt/ paths from the terminal PATH to preserve access to
// Windows executables (issue #495). The test is skipped in WSL2 environments.
func TestBuildSmartPATH_StableAcrossTerminalPATH(t *testing.T) {
	// Skip in WSL2: the output legitimately changes when terminal PATH changes
	// because /mnt/ entries are captured to preserve Windows executable access.
	if IsWSL2() {
		t.Skip("WSL2: BuildSmartPATH captures /mnt/ paths from terminal PATH (issue #495)")
	}

	path1 := BuildSmartPATH()

	// Simulate running from a different environment (e.g., CI on Linux with minimal PATH)
	t.Setenv("PATH", "/tmp/fake-linux-path:/some/other/fake:/usr/bin:/bin")
	path2 := BuildSmartPATH()

	if path1 != path2 {
		t.Errorf("BuildSmartPATH must not capture terminal PATH (issue #467):\ngot1: %s\ngot2: %s", path1, path2)
	}
}

func TestBuildSmartPATH_NonEmpty(t *testing.T) {
	path := BuildSmartPATH()
	if path == "" {
		t.Error("BuildSmartPATH returned empty string")
	}
}

func TestBuildSmartPATH_EssentialDirs(t *testing.T) {
	path := BuildSmartPATH()
	sep := string(os.PathListSeparator)
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	if homeDir == "" {
		t.Skip("cannot determine home directory")
	}

	// Use filepath.Join for cross-platform path construction
	localBin := filepath.Join(homeDir, ".local", "bin")
	goBin := filepath.Join(homeDir, "go", "bin")

	if !PathContainsDir(path, localBin, sep) {
		t.Errorf("PATH should contain %s", localBin)
	}
	if !PathContainsDir(path, goBin, sep) {
		t.Errorf("PATH should contain %s", goBin)
	}
}

func TestPathContainsDir_Cases(t *testing.T) {
	tests := []struct {
		name    string
		pathStr string
		dir     string
		sep     string
		want    bool
	}{
		{"exact match", "/usr/bin:/usr/local/bin", "/usr/bin", ":", true},
		{"no match", "/usr/bin:/usr/local/bin", "/opt/bin", ":", false},
		{"trailing slash on entry", "/usr/bin/:/usr/local/bin", "/usr/bin", ":", true},
		{"trailing slash on dir", "/usr/bin:/usr/local/bin", "/usr/bin/", ":", true},
		{"no false positive on prefix", "/usr/local/bin2:/usr/bin", "/usr/local/bin", ":", false},
		{"empty path", "", "/usr/bin", ":", false},
		{"single entry match", "/usr/bin", "/usr/bin", ":", true},
		{"single entry no match", "/usr/bin", "/opt/bin", ":", false},
		{"windows separator", `C:\Go\bin;C:\Users\bin`, `C:\Go\bin`, ";", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathContainsDir(tt.pathStr, tt.dir, tt.sep)
			if got != tt.want {
				t.Errorf("PathContainsDir(%q, %q, %q) = %v, want %v",
					tt.pathStr, tt.dir, tt.sep, got, tt.want)
			}
		})
	}
}

// --- IsWSL2 tests ---

// TestIsWSL2_EnvVar verifies that IsWSL2 returns true when WSL_DISTRO_NAME is set.
func TestIsWSL2_EnvVar(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	if !IsWSL2() {
		t.Error("IsWSL2() should return true when WSL_DISTRO_NAME is set")
	}
}

// TestIsWSL2_EmptyEnvVar verifies that IsWSL2 returns false when WSL_DISTRO_NAME is empty
// and procVersionPath points to a non-existent file (simulating non-WSL2 Linux).
func TestIsWSL2_EmptyEnvVar(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "")

	// Override procVersionPath to a non-existent file so the test is deterministic
	// regardless of the host environment (prevents skip on actual WSL2).
	orig := procVersionPath
	procVersionPath = filepath.Join(t.TempDir(), "nonexistent")
	t.Cleanup(func() { procVersionPath = orig })

	if IsWSL2() {
		t.Error("IsWSL2() should return false when WSL_DISTRO_NAME is empty and proc/version is absent")
	}
}

// --- WSL2 BuildSmartPATH tests (regression tests for issue #495) ---

// TestBuildSmartPATH_WSL2 is a table-driven test covering all WSL2-specific
// BuildSmartPATH scenarios. Regression tests for issue #495: "v2.7.9 update causes
// env.PATH to overwrite Windows paths in WSL2, blocking access to powershell.exe
// and other Windows executables".
func TestBuildSmartPATH_WSL2(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("WSL2 PATH tests only apply on Linux")
	}

	sep := string(os.PathListSeparator)

	tests := []struct {
		name            string
		terminalPATH    string
		mustContain     []string
		mustNotContain  []string
		checkNoDup      string // if non-empty, assert this entry appears at most once
		isStabilityTest bool   // if true, test PATH stability (non-WSL2 mode)
	}{
		{
			name:         "preserves_windows_paths",
			terminalPATH: "/usr/bin:/bin:/mnt/c/Windows/System32:/mnt/c/Windows/System32/WindowsPowerShell/v1.0:/mnt/c/Windows",
			mustContain: []string{
				"/mnt/c/Windows/System32",
				"/mnt/c/Windows/System32/WindowsPowerShell/v1.0",
				"/mnt/c/Windows",
			},
		},
		{
			name:           "non_mnt_paths_not_included",
			terminalPATH:   "/home/user/custom/tool:/usr/bin:/mnt/c/Windows/System32",
			mustContain:    []string{"/mnt/c/Windows/System32"},
			mustNotContain: []string{"/home/user/custom/tool"},
		},
		{
			name:         "no_duplicates",
			terminalPATH: "/mnt/c/Windows/System32:/mnt/c/Windows/System32",
			mustContain:  []string{"/mnt/c/Windows/System32"},
			checkNoDup:   "/mnt/c/Windows/System32",
		},
		{
			name:           "user_scoped_paths_excluded",
			terminalPATH:   "/mnt/c/Windows/System32:/mnt/c/Users/alice/AppData/Local/Programs/bin:/mnt/c/Users/alice/AppData/Roaming/npm",
			mustContain:    []string{"/mnt/c/Windows/System32"},
			mustNotContain: []string{"/mnt/c/Users/alice/AppData/Local/Programs/bin", "/mnt/c/Users/alice/AppData/Roaming/npm"},
		},
		{
			name:            "non_wsl2_stable",
			isStabilityTest: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() is intentionally omitted: subtests use t.Setenv which
			// modifies global environment state and cannot run concurrently.

			if tc.isStabilityTest {
				// Verify that on standard Linux (not WSL2), BuildSmartPATH is stable.
				t.Setenv("WSL_DISTRO_NAME", "")
				if IsWSL2() {
					t.Skip("running in actual WSL2 environment")
				}
				path1 := BuildSmartPATH()
				t.Setenv("PATH", "/tmp/fake-path:/some/other:/usr/bin:/bin")
				path2 := BuildSmartPATH()
				if path1 != path2 {
					t.Errorf("non-WSL2 Linux: BuildSmartPATH must be stable across terminal PATH changes:\ngot1: %s\ngot2: %s", path1, path2)
				}
				return
			}

			// Simulate WSL2 environment
			t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
			t.Setenv("PATH", tc.terminalPATH)

			path := BuildSmartPATH()

			for _, want := range tc.mustContain {
				if !PathContainsDir(path, want, sep) {
					t.Errorf("WSL2: SmartPATH should contain %q\nfull SmartPATH: %s", want, path)
				}
			}
			for _, unwanted := range tc.mustNotContain {
				if PathContainsDir(path, unwanted, sep) {
					t.Errorf("WSL2: SmartPATH must not contain %q\nfull SmartPATH: %s", unwanted, path)
				}
			}
			if tc.checkNoDup != "" {
				count := 0
				for _, entry := range strings.Split(path, sep) {
					if strings.TrimRight(entry, "/\\") == tc.checkNoDup {
						count++
					}
				}
				if count > 1 {
					t.Errorf("WSL2: %q should appear at most once, found %d times in %q", tc.checkNoDup, count, path)
				}
			}
		})
	}
}

// TestIsUserScopedWindowsPath verifies that per-user Windows directories are rejected.
func TestIsUserScopedWindowsPath(t *testing.T) {
	tests := []struct {
		entry string
		want  bool
	}{
		{"/mnt/c/Users/alice/AppData/Local/Programs/bin", true},
		{"/mnt/c/Users/alice/AppData/Roaming/npm", true},
		{"/mnt/c/users/bob/appdata/local/bin", true},  // case-insensitive
		{"/mnt/c/Documents and Settings/alice/bin", true},
		{"/mnt/c/Windows/System32", false},
		{"/mnt/c/Windows/System32/WindowsPowerShell/v1.0", false},
		{"/mnt/d/tools/bin", false},
		{"/mnt/c/Program Files/Git/bin", false},
		{"/usr/bin", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.entry, func(t *testing.T) {
			if got := isUserScopedWindowsPath(tt.entry); got != tt.want {
				t.Errorf("isUserScopedWindowsPath(%q) = %v, want %v", tt.entry, got, tt.want)
			}
		})
	}
}

// TestIsWSL2DrivePath verifies the drive-mount path filter used by BuildSmartPATH.
func TestIsWSL2DrivePath(t *testing.T) {
	tests := []struct {
		entry string
		want  bool
	}{
		{"/mnt/c/Windows/System32", true},
		{"/mnt/d/tools/bin", true},
		{"/mnt/z/", true},
		{"/mnt/c", true},
		{"/mnt/wslg/distro", false},  // not a single-letter drive
		{"/mnt/foo/bar", false},       // not a drive letter
		{"/mnt/", false},              // no drive letter
		{"/usr/bin", false},           // not /mnt/
		{"/mnt/C/Windows", false},     // uppercase not matched
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.entry, func(t *testing.T) {
			if got := isWSL2DrivePath(tt.entry); got != tt.want {
				t.Errorf("isWSL2DrivePath(%q) = %v, want %v", tt.entry, got, tt.want)
			}
		})
	}
}

// TestIsWSL2_ProcVersionFallback verifies that IsWSL2 can detect WSL2 via
// procVersionPath when WSL_DISTRO_NAME is not set. It overrides procVersionPath
// with a temp file containing a synthetic WSL2 kernel string, making the
// /proc/version fallback path fully testable without a real WSL2 environment.
func TestIsWSL2_ProcVersionFallback(t *testing.T) {
	// Ensure the env-var fast path is inactive
	t.Setenv("WSL_DISTRO_NAME", "")

	// Write a synthetic /proc/version with a WSL2 kernel string
	tmp := t.TempDir()
	fakeProcVersion := filepath.Join(tmp, "version")
	content := "Linux version 6.6.87.2-microsoft-standard-WSL2 (oe-user@oe-host)"
	if err := os.WriteFile(fakeProcVersion, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Override the package-level path and restore it after the test
	orig := procVersionPath
	procVersionPath = fakeProcVersion
	t.Cleanup(func() { procVersionPath = orig })

	if !IsWSL2() {
		t.Error("IsWSL2() should return true when procVersionPath contains WSL2 kernel string")
	}
}

// TestIsWSL2_ProcVersionFallback_NonWSL verifies that IsWSL2 returns false when
// procVersionPath contains a non-WSL kernel string.
func TestIsWSL2_ProcVersionFallback_NonWSL(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "")

	tmp := t.TempDir()
	fakeProcVersion := filepath.Join(tmp, "version")
	content := "Linux version 6.6.1-generic (buildd@lcy02-amd64-051)"
	if err := os.WriteFile(fakeProcVersion, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	orig := procVersionPath
	procVersionPath = fakeProcVersion
	t.Cleanup(func() { procVersionPath = orig })

	if IsWSL2() {
		t.Error("IsWSL2() should return false for a non-WSL kernel string")
	}
}
