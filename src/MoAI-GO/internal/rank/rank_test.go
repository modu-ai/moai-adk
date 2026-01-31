package rank

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCredentialsPath(t *testing.T) {
	path, err := credentialsPath()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
	if filepath.Base(path) != "credentials.json" {
		t.Errorf("expected credentials.json, got %s", filepath.Base(path))
	}
}

func TestConfigPath(t *testing.T) {
	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", filepath.Base(path))
	}
}

func TestSaveAndLoadCredentials(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	creds := &Credentials{
		APIKey:    "test-api-key-123",
		Username:  "testuser",
		UserID:    "uid-456",
		CreatedAt: "2026-01-31T00:00:00Z",
	}

	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.APIKey != creds.APIKey {
		t.Errorf("expected APIKey %s, got %s", creds.APIKey, loaded.APIKey)
	}
	if loaded.Username != creds.Username {
		t.Errorf("expected Username %s, got %s", creds.Username, loaded.Username)
	}
	if loaded.UserID != creds.UserID {
		t.Errorf("expected UserID %s, got %s", creds.UserID, loaded.UserID)
	}

	// Verify file permissions
	path, pathErr := credentialsPath()
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestDeleteCredentials(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	creds := &Credentials{APIKey: "to-delete", Username: "user"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatal(err)
	}

	if err := DeleteCredentials(); err != nil {
		t.Fatalf("DeleteCredentials failed: %v", err)
	}

	_, err := LoadCredentials()
	if err == nil {
		t.Error("expected error after deleting credentials")
	}
}

func TestDeleteCredentials_NotExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Should not error when file doesn't exist
	if err := DeleteCredentials(); err != nil {
		t.Fatalf("DeleteCredentials should not error for non-existent file: %v", err)
	}
}

func TestIsLoggedIn(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if IsLoggedIn() {
		t.Error("should not be logged in with no credentials")
	}

	creds := &Credentials{APIKey: "valid-key", Username: "user"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatal(err)
	}

	if !IsLoggedIn() {
		t.Error("should be logged in after saving credentials")
	}
}

func TestIsLoggedIn_EmptyAPIKey(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	creds := &Credentials{APIKey: "", Username: "user"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatal(err)
	}

	if IsLoggedIn() {
		t.Error("should not be logged in with empty API key")
	}
}

func TestLoadRankConfig_NotExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := LoadRankConfig()
	if err != nil {
		t.Fatalf("LoadRankConfig should return empty config when file missing: %v", err)
	}
	if len(cfg.ExcludeProjects) != 0 {
		t.Errorf("expected empty exclusion list, got %d items", len(cfg.ExcludeProjects))
	}
}

func TestSaveAndLoadRankConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &RankConfig{
		ExcludeProjects: []string{"/path/to/project1", "/path/to/project2"},
	}
	if err := SaveRankConfig(cfg); err != nil {
		t.Fatalf("SaveRankConfig failed: %v", err)
	}

	loaded, err := LoadRankConfig()
	if err != nil {
		t.Fatalf("LoadRankConfig failed: %v", err)
	}
	if len(loaded.ExcludeProjects) != 2 {
		t.Errorf("expected 2 exclusions, got %d", len(loaded.ExcludeProjects))
	}
}

func TestAddExclusion(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := AddExclusion("/my/project"); err != nil {
		t.Fatalf("AddExclusion failed: %v", err)
	}

	cfg, err := LoadRankConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ExcludeProjects) != 1 {
		t.Fatalf("expected 1 exclusion, got %d", len(cfg.ExcludeProjects))
	}

	// Adding same path again should be idempotent
	if err := AddExclusion("/my/project"); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadRankConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ExcludeProjects) != 1 {
		t.Errorf("expected 1 exclusion after duplicate add, got %d", len(cfg.ExcludeProjects))
	}
}

func TestRemoveExclusion(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := AddExclusion("/project-a"); err != nil {
		t.Fatal(err)
	}
	if err := AddExclusion("/project-b"); err != nil {
		t.Fatal(err)
	}

	if err := RemoveExclusion("/project-a"); err != nil {
		t.Fatalf("RemoveExclusion failed: %v", err)
	}

	cfg, err := LoadRankConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ExcludeProjects) != 1 {
		t.Fatalf("expected 1 exclusion after remove, got %d", len(cfg.ExcludeProjects))
	}
	if cfg.ExcludeProjects[0] != "/project-b" {
		t.Errorf("expected /project-b remaining, got %s", cfg.ExcludeProjects[0])
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("test-key")
	if client.apiKey != "test-key" {
		t.Errorf("expected apiKey test-key, got %s", client.apiKey)
	}
	if client.baseURL != defaultBaseURL {
		t.Errorf("expected baseURL %s, got %s", defaultBaseURL, client.baseURL)
	}
}

func TestNewClientWithURL(t *testing.T) {
	client := NewClientWithURL("key-abc", "http://localhost:9999")
	if client.apiKey != "key-abc" {
		t.Errorf("expected apiKey key-abc, got %s", client.apiKey)
	}
	if client.baseURL != "http://localhost:9999" {
		t.Errorf("expected custom baseURL, got %s", client.baseURL)
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := generateState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected 64 char state, got %d", len(state1))
	}

	state2, err := generateState()
	if err != nil {
		t.Fatal(err)
	}
	if state1 == state2 {
		t.Error("two generated states should be different")
	}
}

func TestFindAvailablePort(t *testing.T) {
	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("findAvailablePort failed: %v", err)
	}
	if port < MinPort || port > MaxPort {
		t.Errorf("port %d outside range [%d, %d]", port, MinPort, MaxPort)
	}
}

func TestContainsHookRef(t *testing.T) {
	tests := []struct {
		command  string
		hookName string
		want     bool
	}{
		{"python /path/to/session_end__rank_submit.py", "session_end__rank_submit.py", true},
		{"bash -l -c 'python /path/to/session_end__rank_submit.py'", "session_end__rank_submit.py", true},
		{"python /path/to/other_hook.py", "session_end__rank_submit.py", false},
		{"", "session_end__rank_submit.py", false},
	}

	for _, tt := range tests {
		got := containsHookRef(tt.command, tt.hookName)
		if got != tt.want {
			t.Errorf("containsHookRef(%q, %q) = %v, want %v", tt.command, tt.hookName, got, tt.want)
		}
	}
}

func TestHookInstallAndUninstall(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Initially not installed
	if IsHookInstalled() {
		t.Error("hook should not be installed initially")
	}

	// Install
	if err := InstallHook(); err != nil {
		t.Fatalf("InstallHook failed: %v", err)
	}

	if !IsHookInstalled() {
		t.Error("hook should be installed after InstallHook")
	}

	// Verify hook script exists
	hookDir, err := globalHookDir()
	if err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(hookDir, HookFileName)
	if _, statErr := os.Stat(hookPath); os.IsNotExist(statErr) {
		t.Error("hook script file should exist")
	}

	// Verify settings.json was created
	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.json should exist: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("settings.json should be valid JSON: %v", err)
	}

	// Install again (idempotent)
	if err := InstallHook(); err != nil {
		t.Fatalf("second InstallHook should be idempotent: %v", err)
	}

	// Verify only one entry
	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	hooks := settings["hooks"].(map[string]any)
	entries := hooks["SessionEnd"].([]any)
	if len(entries) != 1 {
		t.Errorf("expected 1 hook entry after idempotent install, got %d", len(entries))
	}

	// Uninstall
	if err := UninstallHook(); err != nil {
		t.Fatalf("UninstallHook failed: %v", err)
	}

	if IsHookInstalled() {
		t.Error("hook should not be installed after UninstallHook")
	}

	// Hook script should be removed
	if _, statErr := os.Stat(hookPath); !os.IsNotExist(statErr) {
		t.Error("hook script file should be removed")
	}
}

func TestUninstallHook_NoSettings(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Should not error when settings doesn't exist
	if err := UninstallHook(); err != nil {
		t.Fatalf("UninstallHook should not error without settings: %v", err)
	}
}

func TestBuildHookCommand(t *testing.T) {
	cmd := buildHookCommand("/path/to/hook.py")
	if len(cmd) == 0 {
		t.Error("hook command should not be empty")
	}
	if !containsHookRef(cmd, "hook.py") {
		t.Errorf("hook command should reference the hook file: %s", cmd)
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		count int64
		want  string
	}{
		{0, "0"},
		{500, "500"},
		{1500, "1.5K"},
		{1_500_000, "1.5M"},
	}

	for _, tt := range tests {
		got := FormatTokens(tt.count)
		if got != tt.want {
			t.Errorf("FormatTokens(%d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestRankMedal(t *testing.T) {
	if RankMedal(1) != "[1st]" {
		t.Errorf("expected [1st] for position 1, got %s", RankMedal(1))
	}
	if RankMedal(2) != "[2nd]" {
		t.Errorf("expected [2nd] for position 2, got %s", RankMedal(2))
	}
	if RankMedal(3) != "[3rd]" {
		t.Errorf("expected [3rd] for position 3, got %s", RankMedal(3))
	}
	if RankMedal(4) != "" {
		t.Errorf("expected empty for position 4, got %s", RankMedal(4))
	}
}

func TestGenerateHookScript(t *testing.T) {
	script := generateHookScript()
	if len(script) == 0 {
		t.Error("hook script should not be empty")
	}
	if !contains(script, "submit_session_hook") {
		t.Error("hook script should reference submit_session_hook")
	}
	if !contains(script, "moai_adk.rank.hook") {
		t.Error("hook script should reference moai_adk.rank.hook module")
	}
}

// --- New tests for improved coverage ---

func TestAuthURL(t *testing.T) {
	url := AuthURL()
	expected := "https://rank.mo.ai.kr/api/auth/cli"
	if url != expected {
		t.Errorf("AuthURL() = %q, want %q", url, expected)
	}
}

func TestReadSettings_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(path, []byte("not valid json {{{"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := readSettings(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadSettings_ValidExisting(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")
	content := `{"key": "value", "number": 42}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	settings, err := readSettings(path)
	if err != nil {
		t.Fatalf("readSettings failed: %v", err)
	}
	if settings["key"] != "value" {
		t.Errorf("expected key=value, got %v", settings["key"])
	}
}

func TestReadSettings_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")
	settings, err := readSettings(path)
	if err != nil {
		t.Fatalf("readSettings should return empty map for non-existent file: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected empty map, got %d entries", len(settings))
	}
}

func TestWriteSettings_CreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "nested", "settings.json")
	settings := map[string]any{"test": true}
	if err := writeSettings(path, settings); err != nil {
		t.Fatalf("writeSettings failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file should exist: %v", err)
	}
	var loaded map[string]any
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("should be valid JSON: %v", err)
	}
	if loaded["test"] != true {
		t.Errorf("expected test=true, got %v", loaded["test"])
	}
}

func TestUninstallHook_SettingsWithNoHooks(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"some_setting": "value"}`
	if err := os.WriteFile(settingsPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UninstallHook(); err != nil {
		t.Fatalf("UninstallHook should succeed with settings missing hooks: %v", err)
	}
}

func TestUninstallHook_InvalidJSON(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	err = UninstallHook()
	if err == nil {
		t.Error("expected error for invalid JSON settings")
	}
}

func TestRemoveSessionEndHook_NoSessionEnd(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{"type": "command", "command": "echo pre"},
			},
		},
	}
	removeSessionEndHook(settings)
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks section should remain")
	}
	if _, ok := hooks["PreToolUse"]; !ok {
		t.Error("PreToolUse hooks should be preserved")
	}
}

func TestRemoveSessionEndHook_NoHooksSection(t *testing.T) {
	settings := map[string]any{
		"some_key": "some_value",
	}
	removeSessionEndHook(settings)
	if _, ok := settings["hooks"]; ok {
		t.Error("hooks should not be created by removeSessionEndHook")
	}
}

func TestRemoveSessionEndHook_PreservesOtherEntries(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"SessionEnd": []any{
				map[string]any{"type": "command", "command": "echo other"},
				map[string]any{"type": "command", "command": "python /path/to/session_end__rank_submit.py"},
			},
		},
	}
	removeSessionEndHook(settings)

	hooks := settings["hooks"].(map[string]any)
	entries, ok := hooks["SessionEnd"].([]any)
	if !ok {
		t.Fatal("SessionEnd should still exist with remaining entries")
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(entries))
	}
}

func TestRemoveSessionEndHook_DeletesKeyWhenEmpty(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"SessionEnd": []any{
				map[string]any{"type": "command", "command": "python /path/to/session_end__rank_submit.py"},
			},
		},
	}
	removeSessionEndHook(settings)

	hooks := settings["hooks"].(map[string]any)
	if _, ok := hooks["SessionEnd"]; ok {
		t.Error("SessionEnd key should be removed when all entries are removed")
	}
}

func TestAddSessionEndHook_WithExistingOtherHooks(t *testing.T) {
	settings := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{"type": "command", "command": "echo pre"},
			},
		},
	}
	addSessionEndHook(settings, "python /path/to/session_end__rank_submit.py")

	hooks := settings["hooks"].(map[string]any)
	if _, ok := hooks["PreToolUse"]; !ok {
		t.Error("PreToolUse hooks should be preserved")
	}
	entries, ok := hooks["SessionEnd"].([]any)
	if !ok {
		t.Fatal("SessionEnd should exist after addSessionEndHook")
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 SessionEnd entry, got %d", len(entries))
	}
}

func TestAddSessionEndHook_NoHooksSection(t *testing.T) {
	settings := map[string]any{
		"some_setting": "value",
	}
	addSessionEndHook(settings, "python /path/to/session_end__rank_submit.py")

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks section should be created")
	}
	entries, ok := hooks["SessionEnd"].([]any)
	if !ok {
		t.Fatal("SessionEnd should exist")
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestAddSessionEndHook_Idempotent(t *testing.T) {
	settings := map[string]any{}
	cmd := "python /path/to/session_end__rank_submit.py"

	addSessionEndHook(settings, cmd)
	addSessionEndHook(settings, cmd)

	hooks := settings["hooks"].(map[string]any)
	entries := hooks["SessionEnd"].([]any)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after idempotent add, got %d", len(entries))
	}
}

func TestSaveCredentials_InvalidDir(t *testing.T) {
	tmpHome := t.TempDir()
	// Create a regular file where the .moai directory should be,
	// so MkdirAll fails trying to create .moai/rank/
	blockingFile := filepath.Join(tmpHome, ".moai")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmpHome)

	creds := &Credentials{APIKey: "test", Username: "user"}
	err := SaveCredentials(creds)
	if err == nil {
		t.Error("expected error when directory creation is blocked by file")
	}
}

func TestSaveRankConfig_InvalidDir(t *testing.T) {
	tmpHome := t.TempDir()
	blockingFile := filepath.Join(tmpHome, ".moai")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmpHome)

	cfg := &RankConfig{ExcludeProjects: []string{"/test"}}
	err := SaveRankConfig(cfg)
	if err == nil {
		t.Error("expected error when directory creation is blocked by file")
	}
}

func TestLoadCredentials_InvalidJSON(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := credentialsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not json {{{"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err = LoadCredentials()
	if err == nil {
		t.Error("expected error for invalid JSON credentials")
	}
}

func TestLoadRankConfig_InvalidYAML(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	// Write content that triggers a YAML parse error
	if err := os.WriteFile(path, []byte("{{{\ninvalid:\n  - [\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = LoadRankConfig()
	if err == nil {
		t.Error("expected error for invalid YAML config")
	}
}

func TestIsHookInstalled_InvalidJSON(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false for invalid JSON settings")
	}
}

func TestIsHookInstalled_NoHooksKey(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"other": "value"}`), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false when no hooks key")
	}
}

func TestIsHookInstalled_InvalidHooksStructure(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	// hooks is a string instead of an object
	if err := os.WriteFile(settingsPath, []byte(`{"hooks": "not an object"}`), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false when hooks is not an object")
	}
}

func TestIsHookInstalled_NoSessionEnd(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"hooks": {"PreToolUse": []}}`), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false when no SessionEnd key")
	}
}

func TestIsHookInstalled_InvalidSessionEnd(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	// SessionEnd is a string instead of an array
	if err := os.WriteFile(settingsPath, []byte(`{"hooks": {"SessionEnd": "not an array"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false when SessionEnd is not an array")
	}
}

func TestIsHookInstalled_EmptySessionEnd(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"hooks": {"SessionEnd": []}}`), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false when SessionEnd is empty")
	}
}

func TestIsHookInstalled_DifferentHook(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"hooks": {"SessionEnd": [{"type": "command", "command": "echo hello"}]}}`
	if err := os.WriteFile(settingsPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if IsHookInstalled() {
		t.Error("should return false when SessionEnd has different hook")
	}
}

func TestBuildHookCommand_PlatformSpecific(t *testing.T) {
	hookPath := "/home/user/.claude/hooks/moai/session_end__rank_submit.py"
	cmd := buildHookCommand(hookPath)

	if !contains(cmd, hookPath) {
		t.Errorf("command should contain hook path: %s", cmd)
	}
	if !contains(cmd, "python") {
		t.Errorf("command should contain python: %s", cmd)
	}
	if runtime.GOOS != "windows" {
		if !contains(cmd, "${SHELL:-/bin/bash}") {
			t.Errorf("non-Windows command should include shell wrapper: %s", cmd)
		}
	}
}

func TestInstallHook_WithExistingSettings(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	existing := `{"some_key": "some_value"}`
	if err := os.WriteFile(settingsPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := InstallHook(); err != nil {
		t.Fatalf("InstallHook failed with existing settings: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	if settings["some_key"] != "some_value" {
		t.Error("existing settings should be preserved")
	}
	if _, ok := settings["hooks"]; !ok {
		t.Error("hooks should be added")
	}
}

func TestDeleteCredentials_WithInvalidDir(t *testing.T) {
	tmpHome := t.TempDir()
	// Create a file where .moai directory should be so path resolution works
	// but the file doesn't exist (normal case already tested)
	// Here we test that credentialsPath itself works when HOME is set
	t.Setenv("HOME", tmpHome)

	// Ensure no error when credentials file does not exist
	err := DeleteCredentials()
	if err != nil {
		t.Fatalf("DeleteCredentials should not error for non-existent path: %v", err)
	}
}

func TestLoadCredentials_NoFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, err := LoadCredentials()
	if err == nil {
		t.Error("expected error when credentials file does not exist")
	}
}

func TestInstallHook_InvalidSettingsJSON(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	err = InstallHook()
	if err == nil {
		t.Error("expected error when settings.json contains invalid JSON")
	}
}

func TestWriteSettings_MkdirAllError(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a file that blocks directory creation
	blockingFile := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(blockingFile, "subdir", "settings.json")
	settings := map[string]any{"test": true}
	err := writeSettings(path, settings)
	if err == nil {
		t.Error("expected error when parent directory creation fails")
	}
}

func TestWriteSettings_MarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")
	// Functions cannot be marshaled to JSON
	settings := map[string]any{
		"bad": func() {},
	}
	err := writeSettings(path, settings)
	if err == nil {
		t.Error("expected error for unmarshable value")
	}
}

func TestReadSettings_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a directory where the file should be - ReadFile on a directory
	// returns an error that is NOT os.IsNotExist
	dirPath := filepath.Join(tmpDir, "settings.json")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatal(err)
	}
	_, err := readSettings(dirPath)
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}

func TestInstallHook_HookDirCreationBlocked(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Block .claude directory creation with a file
	blockingFile := filepath.Join(tmpHome, ".claude")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0644); err != nil {
		t.Fatal(err)
	}

	err := InstallHook()
	if err == nil {
		t.Error("expected error when hook directory creation is blocked")
	}
}

func TestDeleteCredentials_RemoveError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := credentialsPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	// Create a non-empty directory at the credentials path.
	// os.Remove on a non-empty directory returns ENOTEMPTY, not IsNotExist.
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	innerFile := filepath.Join(path, "inner.txt")
	if err := os.WriteFile(innerFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	err = DeleteCredentials()
	if err == nil {
		t.Error("expected error when credentials path is a non-empty directory")
	}
}

func TestLoadRankConfig_ReadError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	// Create a directory where config.yaml should be - ReadFile on directory
	// returns a non-NotExist error
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}

	_, err = LoadRankConfig()
	if err == nil {
		t.Error("expected error when config path is a directory")
	}
}

func TestUninstallHook_WithHookScript(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install hook first, then verify uninstall removes the script file too
	if err := InstallHook(); err != nil {
		t.Fatalf("InstallHook failed: %v", err)
	}

	hookDir, err := globalHookDir()
	if err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(hookDir, HookFileName)
	if _, statErr := os.Stat(hookPath); os.IsNotExist(statErr) {
		t.Fatal("hook script should exist after install")
	}

	if err := UninstallHook(); err != nil {
		t.Fatalf("UninstallHook failed: %v", err)
	}

	// Verify hook script was removed
	if _, statErr := os.Stat(hookPath); !os.IsNotExist(statErr) {
		t.Error("hook script should be removed after uninstall")
	}

	// Verify SessionEnd was cleaned from settings
	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	hooks, ok := settings["hooks"].(map[string]any)
	if ok {
		if _, hasSessionEnd := hooks["SessionEnd"]; hasSessionEnd {
			t.Error("SessionEnd should be removed from settings after uninstall")
		}
	}
}

func TestAddExclusion_LoadError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	// Write invalid YAML so LoadRankConfig fails
	if err := os.WriteFile(path, []byte("{{{\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err = AddExclusion("/some/project")
	if err == nil {
		t.Error("expected error when LoadRankConfig fails")
	}
}

func TestRemoveExclusion_LoadError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	// Write invalid YAML so LoadRankConfig fails
	if err := os.WriteFile(path, []byte("{{{\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err = RemoveExclusion("/some/project")
	if err == nil {
		t.Error("expected error when LoadRankConfig fails")
	}
}

func TestInstallHook_WriteScriptError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	hookDir, err := globalHookDir()
	if err != nil {
		t.Fatal(err)
	}
	// Create the hook directory so MkdirAll succeeds
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Make it read-only so WriteFile for the script fails
	if err := os.Chmod(hookDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if chmodErr := os.Chmod(hookDir, 0755); chmodErr != nil {
			t.Logf("cleanup chmod: %v", chmodErr)
		}
	})

	err = InstallHook()
	if err == nil {
		t.Error("expected error when hook script write fails")
	}
}

func TestUninstallHook_WriteSettingsFileReadOnly(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install hook so there is content to uninstall
	if err := InstallHook(); err != nil {
		t.Fatal(err)
	}

	settingsPath, err := globalSettingsPath()
	if err != nil {
		t.Fatal(err)
	}
	// Make settings.json read-only so writeSettings fails
	if err := os.Chmod(settingsPath, 0444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if chmodErr := os.Chmod(settingsPath, 0644); chmodErr != nil {
			t.Logf("cleanup chmod: %v", chmodErr)
		}
	})

	err = UninstallHook()
	if err == nil {
		t.Error("expected error when settings.json is read-only")
	}
}

func TestUninstallHook_RemoveScriptError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install hook first
	if err := InstallHook(); err != nil {
		t.Fatal(err)
	}

	hookDir, err := globalHookDir()
	if err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(hookDir, HookFileName)
	// Replace hook script with a non-empty directory so os.Remove fails
	// with ENOTEMPTY (not IsNotExist)
	if err := os.Remove(hookPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hookPath, 0755); err != nil {
		t.Fatal(err)
	}
	innerFile := filepath.Join(hookPath, "inner.txt")
	if err := os.WriteFile(innerFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	err = UninstallHook()
	if err == nil {
		t.Error("expected error when hook script removal fails")
	}
}

func TestIsHookInstalled_MatchingHookPresent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Install hook so a matching entry exists
	if err := InstallHook(); err != nil {
		t.Fatal(err)
	}

	if !IsHookInstalled() {
		t.Error("should return true when matching hook is installed")
	}
}

func TestSaveCredentials_MarshalIndentPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Save with all fields populated to exercise full marshal path
	creds := &Credentials{
		APIKey:    "key-123",
		Username:  "user",
		UserID:    "uid-456",
		CreatedAt: "2026-01-31T00:00:00Z",
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatal(err)
	}

	// Verify the file content is proper JSON
	path, err := credentialsPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Credentials
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.APIKey != "key-123" || loaded.UserID != "uid-456" {
		t.Errorf("unexpected loaded credentials: %+v", loaded)
	}
}

func TestSaveRankConfig_MarshalPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &RankConfig{
		ExcludeProjects: []string{"/a", "/b", "/c"},
	}
	if err := SaveRankConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadRankConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.ExcludeProjects) != 3 {
		t.Errorf("expected 3 exclusions, got %d", len(loaded.ExcludeProjects))
	}
}

// =============================================================================
// httptest-based tests for network-dependent functions
// =============================================================================

// newMockRankServer creates an httptest.Server that responds at /users/rank
// with the given status code and body.
func newMockRankServer(statusCode int, body any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/rank" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"missing authorization"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		switch v := body.(type) {
		case string:
			fmt.Fprint(w, v)
		default:
			json.NewEncoder(w).Encode(v)
		}
	}))
}

// validRankResponse returns a UserRankResponse suitable for test assertions.
func validRankResponse() *UserRankResponse {
	return &UserRankResponse{
		Username: "testuser",
		Ranks: map[string]RankInfo{
			"daily": {
				Position:          1,
				TotalParticipants: 100,
				CompositeScore:    95.5,
			},
		},
		Tokens: TokenStats{
			Total:  50000,
			Input:  30000,
			Output: 20000,
		},
	}
}

// --- GetRank tests ---

func TestGetRank_Success(t *testing.T) {
	server := newMockRankServer(http.StatusOK, validRankResponse())
	defer server.Close()

	client := NewClientWithURL("test-api-key", server.URL)
	resp, err := client.GetRank()
	if err != nil {
		t.Fatalf("GetRank returned unexpected error: %v", err)
	}
	if resp.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", resp.Username)
	}
	daily, ok := resp.Ranks["daily"]
	if !ok {
		t.Fatal("expected daily rank entry")
	}
	if daily.Position != 1 {
		t.Errorf("expected daily position 1, got %d", daily.Position)
	}
	if daily.TotalParticipants != 100 {
		t.Errorf("expected 100 total participants, got %d", daily.TotalParticipants)
	}
	if daily.CompositeScore != 95.5 {
		t.Errorf("expected composite score 95.5, got %f", daily.CompositeScore)
	}
	if resp.Tokens.Total != 50000 {
		t.Errorf("expected total tokens 50000, got %d", resp.Tokens.Total)
	}
	if resp.Tokens.Input != 30000 {
		t.Errorf("expected input tokens 30000, got %d", resp.Tokens.Input)
	}
	if resp.Tokens.Output != 20000 {
		t.Errorf("expected output tokens 20000, got %d", resp.Tokens.Output)
	}
}

func TestGetRank_VerifiesAuthHeader(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(validRankResponse())
	}))
	defer server.Close()

	client := NewClientWithURL("my-secret-key", server.URL)
	_, err := client.GetRank()
	if err != nil {
		t.Fatalf("GetRank returned unexpected error: %v", err)
	}
	if capturedAuth != "Bearer my-secret-key" {
		t.Errorf("expected Authorization header 'Bearer my-secret-key', got %q", capturedAuth)
	}
}

func TestGetRank_HTTPError(t *testing.T) {
	server := newMockRankServer(http.StatusUnauthorized, `{"error":"invalid token"}`)
	defer server.Close()

	client := NewClientWithURL("bad-key", server.URL)
	_, err := client.GetRank()
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should contain status code 401: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("error should contain response body: %v", err)
	}
}

func TestGetRank_HTTPErrorVariousStatuses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"forbidden", http.StatusForbidden, `{"error":"access denied"}`},
		{"not found", http.StatusNotFound, `{"error":"not found"}`},
		{"server error", http.StatusInternalServerError, `{"error":"internal server error"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockRankServer(tt.statusCode, tt.body)
			defer server.Close()

			client := NewClientWithURL("key", server.URL)
			_, err := client.GetRank()
			if err == nil {
				t.Fatalf("expected error for status %d", tt.statusCode)
			}
			expected := fmt.Sprintf("%d", tt.statusCode)
			if !strings.Contains(err.Error(), expected) {
				t.Errorf("error should contain status code %s: %v", expected, err)
			}
		})
	}
}

func TestGetRank_InvalidJSON(t *testing.T) {
	server := newMockRankServer(http.StatusOK, "this is not json {{{")
	defer server.Close()

	client := NewClientWithURL("key", server.URL)
	_, err := client.GetRank()
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "parse rank response") {
		t.Errorf("error should mention parsing failure: %v", err)
	}
}

func TestGetRank_ConnectionError(t *testing.T) {
	// Use a URL where nothing is listening
	client := NewClientWithURL("key", "http://127.0.0.1:1")
	_, err := client.GetRank()
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("error should mention connection failure: %v", err)
	}
}

func TestGetRank_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write empty body (EOF for JSON decoder)
	}))
	defer server.Close()

	client := NewClientWithURL("key", server.URL)
	_, err := client.GetRank()
	if err == nil {
		t.Fatal("expected error for empty response body")
	}
}

// --- LoginWithAPIKey tests ---

func TestLoginWithAPIKey_Success(t *testing.T) {
	server := newMockRankServer(http.StatusOK, validRankResponse())
	defer server.Close()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	old := defaultBaseURL
	defaultBaseURL = server.URL
	defer func() { defaultBaseURL = old }()

	creds, err := LoginWithAPIKey("test-api-key-999")
	if err != nil {
		t.Fatalf("LoginWithAPIKey returned unexpected error: %v", err)
	}
	if creds.APIKey != "test-api-key-999" {
		t.Errorf("expected APIKey test-api-key-999, got %s", creds.APIKey)
	}
	if creds.Username != "testuser" {
		t.Errorf("expected Username testuser, got %s", creds.Username)
	}
	if creds.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}

	// Verify credentials were saved to disk
	loaded, loadErr := LoadCredentials()
	if loadErr != nil {
		t.Fatalf("LoadCredentials failed after LoginWithAPIKey: %v", loadErr)
	}
	if loaded.APIKey != "test-api-key-999" {
		t.Errorf("saved APIKey mismatch: expected test-api-key-999, got %s", loaded.APIKey)
	}
	if loaded.Username != "testuser" {
		t.Errorf("saved Username mismatch: expected testuser, got %s", loaded.Username)
	}
}

func TestLoginWithAPIKey_APIError(t *testing.T) {
	server := newMockRankServer(http.StatusUnauthorized, `{"error":"unauthorized"}`)
	defer server.Close()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	old := defaultBaseURL
	defaultBaseURL = server.URL
	defer func() { defaultBaseURL = old }()

	_, err := LoginWithAPIKey("bad-key")
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
	if !strings.Contains(err.Error(), "invalid API key") {
		t.Errorf("error should mention invalid API key: %v", err)
	}
}

func TestLoginWithAPIKey_SaveError(t *testing.T) {
	server := newMockRankServer(http.StatusOK, validRankResponse())
	defer server.Close()

	tmpHome := t.TempDir()
	// Block .moai directory creation so SaveCredentials fails
	blockingFile := filepath.Join(tmpHome, ".moai")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmpHome)

	old := defaultBaseURL
	defaultBaseURL = server.URL
	defer func() { defaultBaseURL = old }()

	_, err := LoginWithAPIKey("test-key")
	if err == nil {
		t.Fatal("expected error when SaveCredentials fails")
	}
	if !strings.Contains(err.Error(), "failed to save credentials") {
		t.Errorf("error should mention save failure: %v", err)
	}
}

// --- GetStatus tests ---

func TestGetStatus_NotLoggedIn(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	status, err := GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned unexpected error: %v", err)
	}
	if status.LoggedIn {
		t.Error("expected LoggedIn to be false when no credentials exist")
	}
}

func TestGetStatus_LoggedInSuccess(t *testing.T) {
	server := newMockRankServer(http.StatusOK, validRankResponse())
	defer server.Close()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Save credentials first
	creds := &Credentials{
		APIKey:   "valid-key",
		Username: "saveduser",
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatal(err)
	}

	old := defaultBaseURL
	defaultBaseURL = server.URL
	defer func() { defaultBaseURL = old }()

	status, err := GetStatus()
	if err != nil {
		t.Fatalf("GetStatus returned unexpected error: %v", err)
	}
	if !status.LoggedIn {
		t.Error("expected LoggedIn to be true")
	}
	// Username comes from the API response, not saved credentials
	if status.Username != "testuser" {
		t.Errorf("expected Username testuser from API, got %s", status.Username)
	}
	if status.Ranks == nil {
		t.Error("expected Ranks to be populated")
	}
	daily, ok := status.Ranks["daily"]
	if !ok {
		t.Fatal("expected daily rank entry")
	}
	if daily.Position != 1 {
		t.Errorf("expected daily position 1, got %d", daily.Position)
	}
	if status.Tokens.Total != 50000 {
		t.Errorf("expected total tokens 50000, got %d", status.Tokens.Total)
	}
}

func TestGetStatus_LoggedInAPIError(t *testing.T) {
	server := newMockRankServer(http.StatusInternalServerError, `{"error":"server down"}`)
	defer server.Close()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	creds := &Credentials{
		APIKey:   "valid-key",
		Username: "saveduser",
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatal(err)
	}

	old := defaultBaseURL
	defaultBaseURL = server.URL
	defer func() { defaultBaseURL = old }()

	status, err := GetStatus()
	if err == nil {
		t.Fatal("expected error when API fails")
	}
	if !status.LoggedIn {
		t.Error("expected LoggedIn to be true even when API fails")
	}
	if status.Username != "saveduser" {
		t.Errorf("expected Username saveduser from saved creds, got %s", status.Username)
	}
	// Ranks should be nil when API fails
	if status.Ranks != nil {
		t.Error("expected Ranks to be nil when API fails")
	}
}

// --- NewClient with defaultBaseURL override tests ---

func TestNewClient_UsesDefaultBaseURL(t *testing.T) {
	server := newMockRankServer(http.StatusOK, validRankResponse())
	defer server.Close()

	old := defaultBaseURL
	defaultBaseURL = server.URL
	defer func() { defaultBaseURL = old }()

	client := NewClient("test-key")
	if client.baseURL != server.URL {
		t.Errorf("expected baseURL %s, got %s", server.URL, client.baseURL)
	}

	// Verify the client actually works against the mock server
	resp, err := client.GetRank()
	if err != nil {
		t.Fatalf("GetRank via NewClient with overridden defaultBaseURL failed: %v", err)
	}
	if resp.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", resp.Username)
	}
}

// --- browserOpener override test ---

func TestBrowserOpener_Override(t *testing.T) {
	var capturedURL string
	oldBrowser := browserOpener
	browserOpener = func(url string) error {
		capturedURL = url
		return nil
	}
	defer func() { browserOpener = oldBrowser }()

	err := openBrowser("https://example.com/auth")
	if err != nil {
		t.Fatalf("openBrowser returned unexpected error: %v", err)
	}
	if capturedURL != "https://example.com/auth" {
		t.Errorf("expected captured URL https://example.com/auth, got %s", capturedURL)
	}
}

func TestBrowserOpener_OverrideError(t *testing.T) {
	oldBrowser := browserOpener
	browserOpener = func(url string) error {
		return fmt.Errorf("browser not available")
	}
	defer func() { browserOpener = oldBrowser }()

	err := openBrowser("https://example.com")
	if err == nil {
		t.Fatal("expected error from overridden browserOpener")
	}
	if !strings.Contains(err.Error(), "browser not available") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Login() OAuth flow tests ---

func TestLogin_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	oldBrowser := browserOpener
	browserOpener = func(authURL string) error {
		// Parse the auth URL to extract state and redirect_uri
		parsed, parseErr := url.Parse(authURL)
		if parseErr != nil {
			return parseErr
		}
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")

		// Simulate the OAuth provider sending a callback with credentials
		go func() {
			// Small delay to ensure the callback server is ready
			time.Sleep(50 * time.Millisecond)
			callbackURL := fmt.Sprintf("%s?state=%s&api_key=oauth-key-123&username=oauthuser&user_id=uid-oauth&created_at=2026-01-31T00:00:00Z",
				redirectURI, state)
			resp, getErr := http.Get(callbackURL)
			if getErr == nil {
				resp.Body.Close()
			}
		}()
		return nil
	}
	defer func() { browserOpener = oldBrowser }()

	creds, err := Login()
	if err != nil {
		t.Fatalf("Login returned unexpected error: %v", err)
	}
	if creds.APIKey != "oauth-key-123" {
		t.Errorf("expected APIKey oauth-key-123, got %s", creds.APIKey)
	}
	if creds.Username != "oauthuser" {
		t.Errorf("expected Username oauthuser, got %s", creds.Username)
	}
	if creds.UserID != "uid-oauth" {
		t.Errorf("expected UserID uid-oauth, got %s", creds.UserID)
	}

	// Verify credentials were saved to disk
	loaded, loadErr := LoadCredentials()
	if loadErr != nil {
		t.Fatalf("LoadCredentials failed after Login: %v", loadErr)
	}
	if loaded.APIKey != "oauth-key-123" {
		t.Errorf("saved APIKey mismatch: expected oauth-key-123, got %s", loaded.APIKey)
	}
}

func TestLogin_InvalidState(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	oldBrowser := browserOpener
	browserOpener = func(authURL string) error {
		parsed, parseErr := url.Parse(authURL)
		if parseErr != nil {
			return parseErr
		}
		redirectURI := parsed.Query().Get("redirect_uri")

		// Send callback with WRONG state to trigger invalid state error
		go func() {
			time.Sleep(50 * time.Millisecond)
			callbackURL := fmt.Sprintf("%s?state=wrong-state&api_key=key&username=user",
				redirectURI)
			resp, getErr := http.Get(callbackURL)
			if getErr == nil {
				resp.Body.Close()
			}
		}()
		return nil
	}
	defer func() { browserOpener = oldBrowser }()

	_, err := Login()
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
	if !strings.Contains(err.Error(), "invalid state") {
		t.Errorf("error should mention invalid state: %v", err)
	}
}

func TestLogin_ErrorInCallback(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	oldBrowser := browserOpener
	browserOpener = func(authURL string) error {
		parsed, parseErr := url.Parse(authURL)
		if parseErr != nil {
			return parseErr
		}
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")

		// Send callback with error parameter
		go func() {
			time.Sleep(50 * time.Millisecond)
			callbackURL := fmt.Sprintf("%s?state=%s&error=access_denied",
				redirectURI, state)
			resp, getErr := http.Get(callbackURL)
			if getErr == nil {
				resp.Body.Close()
			}
		}()
		return nil
	}
	defer func() { browserOpener = oldBrowser }()

	_, err := Login()
	if err == nil {
		t.Fatal("expected error for access_denied callback")
	}
	if !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("error should mention access_denied: %v", err)
	}
}

func TestLogin_MissingAPIKey(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	oldBrowser := browserOpener
	browserOpener = func(authURL string) error {
		parsed, parseErr := url.Parse(authURL)
		if parseErr != nil {
			return parseErr
		}
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")

		// Send callback with valid state but no api_key
		go func() {
			time.Sleep(50 * time.Millisecond)
			callbackURL := fmt.Sprintf("%s?state=%s&username=user",
				redirectURI, state)
			resp, getErr := http.Get(callbackURL)
			if getErr == nil {
				resp.Body.Close()
			}
		}()
		return nil
	}
	defer func() { browserOpener = oldBrowser }()

	_, err := Login()
	if err == nil {
		t.Fatal("expected error for missing API key in callback")
	}
	if !strings.Contains(err.Error(), "no API key") {
		t.Errorf("error should mention missing API key: %v", err)
	}
}

func TestLogin_SaveCredentialsError(t *testing.T) {
	tmpHome := t.TempDir()
	// Block .moai directory creation so SaveCredentials fails
	blockingFile := filepath.Join(tmpHome, ".moai")
	if err := os.WriteFile(blockingFile, []byte("blocking"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmpHome)

	oldBrowser := browserOpener
	browserOpener = func(authURL string) error {
		parsed, parseErr := url.Parse(authURL)
		if parseErr != nil {
			return parseErr
		}
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")

		go func() {
			time.Sleep(50 * time.Millisecond)
			callbackURL := fmt.Sprintf("%s?state=%s&api_key=key&username=user&user_id=uid",
				redirectURI, state)
			resp, getErr := http.Get(callbackURL)
			if getErr == nil {
				resp.Body.Close()
			}
		}()
		return nil
	}
	defer func() { browserOpener = oldBrowser }()

	_, err := Login()
	if err == nil {
		t.Fatal("expected error when SaveCredentials fails")
	}
	if !strings.Contains(err.Error(), "failed to save credentials") {
		t.Errorf("error should mention save failure: %v", err)
	}
}

func TestLogin_BrowserOpenError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	oldBrowser := browserOpener
	browserOpener = func(authURL string) error {
		// Simulate browser failure; Login prints message but still waits
		// We need to also send a valid callback so the test does not hang
		parsed, parseErr := url.Parse(authURL)
		if parseErr != nil {
			return fmt.Errorf("browser error")
		}
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")

		go func() {
			time.Sleep(50 * time.Millisecond)
			callbackURL := fmt.Sprintf("%s?state=%s&api_key=key-after-fail&username=user",
				redirectURI, state)
			resp, getErr := http.Get(callbackURL)
			if getErr == nil {
				resp.Body.Close()
			}
		}()
		return fmt.Errorf("browser error")
	}
	defer func() { browserOpener = oldBrowser }()

	// Login should still succeed because the callback was sent
	creds, err := Login()
	if err != nil {
		t.Fatalf("Login should succeed even if browser fails: %v", err)
	}
	if creds.APIKey != "key-after-fail" {
		t.Errorf("expected APIKey key-after-fail, got %s", creds.APIKey)
	}
}
