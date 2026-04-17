package cli

// Coverage gap tests for checkMCPScopeDuplicates, parseMCPJSON, and
// doctor-adjacent functions that were below the 85% threshold.
// All tests use t.TempDir() for isolation.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- parseMCPJSON ---

// TestParseMCPJSON_ValidFile verifies that a well-formed .mcp.json is parsed.
func TestParseMCPJSON_ValidFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `{"mcpServers":{"server1":{},"server2":{}}}`
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := parseMCPJSON(path)
	if len(result) != 2 {
		t.Errorf("expected 2 servers, got %d", len(result))
	}
	if _, ok := result["server1"]; !ok {
		t.Error("server1 missing")
	}
	if _, ok := result["server2"]; !ok {
		t.Error("server2 missing")
	}
}

// TestParseMCPJSON_MissingFile returns nil for a missing file.
func TestParseMCPJSON_MissingFile(t *testing.T) {
	t.Parallel()

	result := parseMCPJSON("/nonexistent/path/.mcp.json")
	if result != nil {
		t.Errorf("expected nil for missing file, got %v", result)
	}
}

// TestParseMCPJSON_MalformedJSON returns nil on parse error.
func TestParseMCPJSON_MalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := parseMCPJSON(path)
	if result != nil {
		t.Errorf("expected nil for malformed JSON, got %v", result)
	}
}

// TestParseMCPJSON_EmptyServers returns empty map when mcpServers is empty.
func TestParseMCPJSON_EmptyServers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{}}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := parseMCPJSON(path)
	if len(result) != 0 {
		t.Errorf("expected 0 servers, got %d", len(result))
	}
}

// TestParseMCPJSON_NullMCPServers handles null mcpServers field.
func TestParseMCPJSON_NullMCPServers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":null}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := parseMCPJSON(path)
	// null unmarshals to nil map — result should have 0 entries.
	if len(result) != 0 {
		t.Errorf("expected 0 servers for null mcpServers, got %d", len(result))
	}
}

// --- checkMCPScopeDuplicates ---

// TestCheckMCPScopeDuplicates_BothEmpty returns OK with "no MCP configuration" msg.
func TestCheckMCPScopeDuplicates_BothEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // No .mcp.json files here or in home.
	check := checkMCPScopeDuplicates(dir, false)

	if check.Name != "MCP Scope Duplicates" {
		t.Errorf("Name = %q", check.Name)
	}
	// Both files missing → OK
	if check.Status != CheckOK {
		t.Errorf("expected OK for both empty, got %q", check.Status)
	}
}

// TestCheckMCPScopeDuplicates_ProjectOnly_NoDup returns OK when only project file exists.
func TestCheckMCPScopeDuplicates_ProjectOnly_NoDup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create project .mcp.json
	content := `{"mcpServers":{"my-server":{}}}`
	if err := os.WriteFile(filepath.Join(dir, ".mcp.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	check := checkMCPScopeDuplicates(dir, false)
	// No global file → no duplicates
	if check.Status != CheckOK {
		t.Errorf("expected OK, got %q: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "no duplicates") {
		t.Errorf("message should say 'no duplicates': %q", check.Message)
	}
}

// TestCheckMCPScopeDuplicates_DuplicatesFound returns Warn.
// Not parallel: uses t.Setenv which requires non-parallel parent.
func TestCheckMCPScopeDuplicates_DuplicatesFound(t *testing.T) {
	// We need to control both the project dir and the global home dir.
	// Use a fake HOME so parseMCPJSON reads from our controlled path.
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome) // Windows fallback

	projectDir := t.TempDir()
	globalClaudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(globalClaudeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	serverName := "duplicate-server"
	projectContent := `{"mcpServers":{"` + serverName + `":{}}}`
	globalContent := `{"mcpServers":{"` + serverName + `":{}}}`

	if err := os.WriteFile(filepath.Join(projectDir, ".mcp.json"), []byte(projectContent), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalClaudeDir, ".mcp.json"), []byte(globalContent), 0o644); err != nil {
		t.Fatalf("write global: %v", err)
	}

	check := checkMCPScopeDuplicates(projectDir, false)
	if check.Status != CheckWarn {
		t.Errorf("expected Warn for duplicate, got %q: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, serverName) {
		t.Errorf("message should contain server name %q: %s", serverName, check.Message)
	}
}

// TestCheckMCPScopeDuplicates_VerboseAddsDetail verifies that verbose mode
// adds a Detail field when duplicates are found.
// Not parallel: uses t.Setenv.
func TestCheckMCPScopeDuplicates_VerboseAddsDetail(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome)

	projectDir := t.TempDir()
	globalClaudeDir := filepath.Join(fakeHome, ".claude")
	_ = os.MkdirAll(globalClaudeDir, 0o755)

	shared := `{"mcpServers":{"shared":{}}}`
	_ = os.WriteFile(filepath.Join(projectDir, ".mcp.json"), []byte(shared), 0o644)
	_ = os.WriteFile(filepath.Join(globalClaudeDir, ".mcp.json"), []byte(shared), 0o644)

	check := checkMCPScopeDuplicates(projectDir, true)
	if check.Detail == "" {
		t.Error("verbose duplicate check should have Detail")
	}
}

// TestCheckMCPScopeDuplicates_NoDuplicates_HasCount verifies message format
// when both scopes exist but have no overlapping names.
// Not parallel: uses t.Setenv.
func TestCheckMCPScopeDuplicates_NoDuplicates_HasCount(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome)

	projectDir := t.TempDir()
	globalClaudeDir := filepath.Join(fakeHome, ".claude")
	_ = os.MkdirAll(globalClaudeDir, 0o755)

	projectContent := `{"mcpServers":{"proj-server":{}}}`
	globalContent := `{"mcpServers":{"global-server":{}}}`
	_ = os.WriteFile(filepath.Join(projectDir, ".mcp.json"), []byte(projectContent), 0o644)
	_ = os.WriteFile(filepath.Join(globalClaudeDir, ".mcp.json"), []byte(globalContent), 0o644)

	check := checkMCPScopeDuplicates(projectDir, false)
	if check.Status != CheckOK {
		t.Errorf("expected OK, got %q: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "1 project") {
		t.Errorf("message should contain '1 project': %q", check.Message)
	}
}

// --- readSettingsLocalForLaunch ---

// TestReadSettingsLocalForLaunch_MissingFile returns empty map when file absent.
// Not parallel: uses os.Chdir which is process-global state.
func TestReadSettingsLocalForLaunch_MissingFile(t *testing.T) {
	// Run from a temp dir with no .claude/settings.local.json.
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	result := readSettingsLocalForLaunch()
	if len(result) != 0 {
		t.Errorf("expected empty map for missing file, got %v", result)
	}
}

// TestReadSettingsLocalForLaunch_ValidFile reads DO_CLAUDE_ vars from env section.
// Not parallel: uses os.Chdir.
func TestReadSettingsLocalForLaunch_ValidFile(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	claudeDir := filepath.Join(tmpDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	settings := map[string]any{
		"env": map[string]string{
			"DO_CLAUDE_MODEL":   "claude-opus-4-7",
			"DO_CLAUDE_BYPASS":  "true",
			"DO_CLAUDE_CHROME":  "false",
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), data, 0o644)

	result := readSettingsLocalForLaunch()
	if result["DO_CLAUDE_MODEL"] != "claude-opus-4-7" {
		t.Errorf("DO_CLAUDE_MODEL = %q, want %q", result["DO_CLAUDE_MODEL"], "claude-opus-4-7")
	}
	if result["DO_CLAUDE_BYPASS"] != "true" {
		t.Errorf("DO_CLAUDE_BYPASS = %q, want %q", result["DO_CLAUDE_BYPASS"], "true")
	}
}

// TestReadSettingsLocalForLaunch_MalformedJSON returns empty map.
// Not parallel: uses os.Chdir.
func TestReadSettingsLocalForLaunch_MalformedJSON(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	claudeDir := filepath.Join(tmpDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte("{bad json"), 0o644)

	result := readSettingsLocalForLaunch()
	if len(result) != 0 {
		t.Errorf("expected empty map for malformed JSON, got %v", result)
	}
}

// --- resolveSymlinks ---

// TestResolveSymlinks_NonExistentPath returns the original path on error.
func TestResolveSymlinks_NonExistentPath(t *testing.T) {
	t.Parallel()

	path := "/nonexistent/path/that/does/not/exist"
	result := resolveSymlinks(path)
	if result != path {
		t.Errorf("resolveSymlinks(%q) = %q, want original path", path, result)
	}
}

// TestResolveSymlinks_ExistingPath returns a non-empty resolved path.
func TestResolveSymlinks_ExistingPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := resolveSymlinks(dir)
	if result == "" {
		t.Error("resolveSymlinks of existing dir should return non-empty")
	}
}

// --- syncPermissionModeToSettingsLocal ---

// TestSyncPermissionModeToSettingsLocal_SetsBypassPermissions writes defaultMode.
func TestSyncPermissionModeToSettingsLocal_SetsBypassPermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.local.json")

	if err := syncPermissionModeToSettingsLocal(path, "bypassPermissions"); err != nil {
		t.Fatalf("error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var s SettingsLocal
	_ = json.Unmarshal(data, &s)
	if s.Permissions["defaultMode"] != "bypassPermissions" {
		t.Errorf("defaultMode = %v, want bypassPermissions", s.Permissions["defaultMode"])
	}
}

// TestSyncPermissionModeToSettingsLocal_AcceptEditsRemovesOverride removes defaultMode.
func TestSyncPermissionModeToSettingsLocal_AcceptEditsRemovesOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.local.json")

	// First set bypass.
	_ = syncPermissionModeToSettingsLocal(path, "bypassPermissions")
	// Then clear to acceptEdits.
	if err := syncPermissionModeToSettingsLocal(path, "acceptEdits"); err != nil {
		t.Fatalf("error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var s SettingsLocal
	_ = json.Unmarshal(data, &s)
	if _, ok := s.Permissions["defaultMode"]; ok {
		t.Error("defaultMode should be absent for acceptEdits mode")
	}
}

// TestSyncPermissionModeToSettingsLocal_EmptyModeRemovesOverride empty string = remove.
func TestSyncPermissionModeToSettingsLocal_EmptyModeRemovesOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.local.json")

	// Set then clear.
	_ = syncPermissionModeToSettingsLocal(path, "auto")
	if err := syncPermissionModeToSettingsLocal(path, ""); err != nil {
		t.Fatalf("error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var s SettingsLocal
	_ = json.Unmarshal(data, &s)
	if _, ok := s.Permissions["defaultMode"]; ok {
		t.Error("defaultMode should be absent for empty mode")
	}
}

// TestSyncPermissionModeToSettingsLocal_PreservesExistingEnv verifies other fields
// in settings.local.json are not lost.
func TestSyncPermissionModeToSettingsLocal_PreservesExistingEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.local.json")

	// Write existing settings with env.
	initial := SettingsLocal{
		Env: map[string]string{"MY_KEY": "my_value"},
	}
	initialData, _ := json.MarshalIndent(initial, "", "  ")
	_ = os.WriteFile(path, initialData, 0o600)

	if err := syncPermissionModeToSettingsLocal(path, "auto"); err != nil {
		t.Fatalf("error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var s SettingsLocal
	_ = json.Unmarshal(data, &s)
	if s.Env["MY_KEY"] != "my_value" {
		t.Errorf("MY_KEY lost: %q", s.Env["MY_KEY"])
	}
	if s.Permissions["defaultMode"] != "auto" {
		t.Errorf("defaultMode = %v, want auto", s.Permissions["defaultMode"])
	}
}

// --- buildEnvForLaunch ---

// TestBuildEnvForLaunch_EmptyEffort returns base unchanged.
func TestBuildEnvForLaunch_EmptyEffort(t *testing.T) {
	t.Parallel()

	base := []string{"A=1", "B=2"}
	result := buildEnvForLaunch("", base)
	if len(result) != len(base) {
		t.Errorf("expected same slice length, got %d", len(result))
	}
}

// TestBuildEnvForLaunch_AddsNewEntry adds CLAUDE_CODE_EFFORT_LEVEL when absent.
func TestBuildEnvForLaunch_AddsNewEntry(t *testing.T) {
	t.Parallel()

	base := []string{"A=1", "B=2"}
	result := buildEnvForLaunch("high", base)
	found := false
	for _, e := range result {
		if e == "CLAUDE_CODE_EFFORT_LEVEL=high" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("CLAUDE_CODE_EFFORT_LEVEL=high not found in %v", result)
	}
}

// TestBuildEnvForLaunch_ReplacesExisting replaces existing CLAUDE_CODE_EFFORT_LEVEL.
func TestBuildEnvForLaunch_ReplacesExisting(t *testing.T) {
	t.Parallel()

	base := []string{"CLAUDE_CODE_EFFORT_LEVEL=low", "A=1"}
	result := buildEnvForLaunch("max", base)

	count := 0
	for _, e := range result {
		if strings.HasPrefix(e, "CLAUDE_CODE_EFFORT_LEVEL=") {
			count++
			if e != "CLAUDE_CODE_EFFORT_LEVEL=max" {
				t.Errorf("expected CLAUDE_CODE_EFFORT_LEVEL=max, got %q", e)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 CLAUDE_CODE_EFFORT_LEVEL entry, got %d", count)
	}
}

// TestBuildEnvForLaunch_PreservesOtherVars other env vars remain untouched.
func TestBuildEnvForLaunch_PreservesOtherVars(t *testing.T) {
	t.Parallel()

	base := []string{"PATH=/usr/bin", "HOME=/home/user", "CLAUDE_CODE_EFFORT_LEVEL=medium"}
	result := buildEnvForLaunch("xhigh", base)

	seen := make(map[string]bool)
	for _, e := range result {
		k := strings.SplitN(e, "=", 2)[0]
		seen[k] = true
	}

	if !seen["PATH"] {
		t.Error("PATH not preserved")
	}
	if !seen["HOME"] {
		t.Error("HOME not preserved")
	}
	if !seen["CLAUDE_CODE_EFFORT_LEVEL"] {
		t.Error("CLAUDE_CODE_EFFORT_LEVEL not present")
	}
}

// TestSyncBypassToSettingsLocal_TrueSetsBypass backward-compat wrapper.
func TestSyncBypassToSettingsLocal_TrueSetsBypass(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.local.json")

	if err := syncBypassToSettingsLocal(path, true); err != nil {
		t.Fatalf("error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var s SettingsLocal
	_ = json.Unmarshal(data, &s)
	if s.Permissions["defaultMode"] != "bypassPermissions" {
		t.Errorf("defaultMode = %v, want bypassPermissions", s.Permissions["defaultMode"])
	}
}

// TestSyncBypassToSettingsLocal_FalseRemovesBypass removes override.
func TestSyncBypassToSettingsLocal_FalseRemovesBypass(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.local.json")

	_ = syncBypassToSettingsLocal(path, true)
	if err := syncBypassToSettingsLocal(path, false); err != nil {
		t.Fatalf("error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var s SettingsLocal
	_ = json.Unmarshal(data, &s)
	if _, ok := s.Permissions["defaultMode"]; ok {
		t.Error("defaultMode should be absent after bypass=false")
	}
}
