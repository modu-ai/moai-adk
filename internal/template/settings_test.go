package template

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
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

	tmpl, err := template.New(tmplPath).Parse(string(data))
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

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
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

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	env, ok := settings["env"].(map[string]interface{})
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

func TestSettingsTemplateSandbox(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	sandbox, ok := settings["sandbox"].(map[string]interface{})
	if !ok {
		t.Fatal("missing sandbox section")
	}
	if sandbox["enabled"] != true {
		t.Error("sandbox should be enabled")
	}
	if sandbox["autoAllowBashIfSandboxed"] != true {
		t.Error("autoAllowBashIfSandboxed should be true")
	}
}

func TestSettingsTemplateAttribution(t *testing.T) {
	ctx := testContext("darwin")
	output := renderTemplate(t, ".claude/settings.json.tmpl", ctx)

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	attr, ok := settings["attribution"].(map[string]interface{})
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

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	sl, ok := settings["statusLine"].(map[string]interface{})
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

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	perms, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		t.Fatal("missing permissions section")
	}

	if perms["defaultMode"] != "default" {
		t.Errorf("permissions.defaultMode = %v, want %q", perms["defaultMode"], "default")
	}

	allow, ok := perms["allow"].([]interface{})
	if !ok {
		t.Fatal("permissions.allow is not an array")
	}
	if len(allow) < 50 {
		t.Errorf("expected at least 50 allow entries, got %d", len(allow))
	}

	deny, ok := perms["deny"].([]interface{})
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

	var settings map[string]interface{}
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

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &settings); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify new boolean fields
	boolFields := map[string]bool{
		"enableAllProjectMcpServers": true,
		"respectGitignore":          true,
		"spinnerTipsEnabled":        true,
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

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &config); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	servers, ok := config["mcpServers"].(map[string]interface{})
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

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &config); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	startup, ok := config["staggeredStartup"].(map[string]interface{})
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

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &config); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expected := "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json"
	if config["$schema"] != expected {
		t.Errorf("$schema = %v, want %q", config["$schema"], expected)
	}
}

// --- BuildSmartPATH and PathContainsDir tests ---

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
