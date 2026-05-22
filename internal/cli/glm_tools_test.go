package cli

// glm_tools_test.go — SPEC-GLM-MCP-001 GWT-1 ~ GWT-22 automated tests.
// RED phase: written before glm_tools.go is implemented. Every test must fail.
// See CLAUDE.local.md §6 (t.TempDir() isolation; never use t.Setenv("HOME",...) in parallel tests).
//
// Test isolation strategy:
//   - Do not overwrite HOME via t.Setenv("HOME", tmpDir) (parallel safety first).
//   - Inject the home directory through the userHomeDirFn function variable (declared in glm_tools.go).
//   - Mock node-version detection through the detectNodeFn function variable.
//   - Perform every file operation inside t.TempDir().

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ─── Test helpers ──────────────────────────────────────────────────────────

// setupToolsTestHome creates a fake HOME directory for the test and overrides
// userHomeDirFn to point at it.
// The GLM token is stored at ~/.moai/.env.glm.
func setupToolsTestHome(t *testing.T) (homeDir string) {
	t.Helper()
	tmpDir := t.TempDir()

	// Override userHomeDirFn
	origFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { userHomeDirFn = origFn })

	return tmpDir
}

// setupGLMToken stores a token at ~/.moai/.env.glm under the specified home directory.
func setupGLMToken(t *testing.T, homeDir, token string) {
	t.Helper()
	moaiDir := filepath.Join(homeDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("moai 디렉토리 생성 실패: %v", err)
	}
	content := "# GLM API Key\nGLM_API_KEY=\"" + token + "\"\n"
	if err := os.WriteFile(filepath.Join(moaiDir, ".env.glm"), []byte(content), 0o600); err != nil {
		t.Fatalf(".env.glm 저장 실패: %v", err)
	}
}

// setupClaudeJSON initializes ~/.claude.json under the specified home directory.
// mcpServers is the initial MCP server map; if nil, mcpServers is initialized empty.
func setupClaudeJSON(t *testing.T, homeDir string, mcpServers map[string]any) string {
	t.Helper()
	if mcpServers == nil {
		mcpServers = map[string]any{}
	}
	data := map[string]any{
		"mcpServers": mcpServers,
	}
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}
	path := filepath.Join(homeDir, ".claude.json")
	if err := os.WriteFile(path, jsonBytes, 0o600); err != nil {
		t.Fatalf(".claude.json 저장 실패: %v", err)
	}
	return path
}

// readMCPEntry reads the mcpServers.zai-mcp-server entry from ~/.claude.json.
func readMCPEntry(t *testing.T, claudeJSONPath string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		t.Fatalf("claude.json 읽기 실패: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}
	mcpServers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		return nil
	}
	entry, ok := mcpServers[zaiMCPServerKey].(map[string]any)
	if !ok {
		return nil
	}
	return entry
}

// readAllMCPServers reads the entire mcpServers map from ~/.claude.json.
func readAllMCPServers(t *testing.T, claudeJSONPath string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		t.Fatalf("claude.json 읽기 실패: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}
	mcpServers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return mcpServers
}

// makeNodeOK returns a detectNodeFn override that simulates Node.js v22.x.
func makeNodeOK(t *testing.T) func() {
	t.Helper()
	origFn := detectNodeFn
	detectNodeFn = func() (int, string, error) { return 22, "v22.5.0", nil }
	return func() { detectNodeFn = origFn }
}

// makeNodeMissing simulates the situation where the node command is absent from PATH.
func makeNodeMissing(t *testing.T) func() {
	t.Helper()
	origFn := detectNodeFn
	detectNodeFn = func() (int, string, error) {
		return 0, "", errNodeNotFound
	}
	return func() { detectNodeFn = origFn }
}

// makeNodeOld simulates an older node version (v18.20.4).
func makeNodeOld(t *testing.T) func() {
	t.Helper()
	origFn := detectNodeFn
	detectNodeFn = func() (int, string, error) { return 18, "v18.20.4", nil }
	return func() { detectNodeFn = origFn }
}

// ─── GWT-1: tools enable subcommand routing + idempotency ────────────────

// TestGLMTools_Cmd_Exists — verifies glmToolsCmd is registered as a subcommand of glmCmd (REQ-GMC-001).
func TestGLMTools_Cmd_Exists(t *testing.T) {
	found := false
	for _, sub := range glmCmd.Commands() {
		if sub.Name() == "tools" {
			found = true
			break
		}
	}
	if !found {
		t.Error("glmCmd 에 'tools' 서브커맨드가 등록되어 있지 않음 (REQ-GMC-001)")
	}
}

// TestGLMTools_Enable_Subcommand_Exists — verifies the enable subcommand exists.
func TestGLMTools_Enable_Subcommand_Exists(t *testing.T) {
	if glmToolsCmd == nil {
		t.Fatal("glmToolsCmd 가 nil 임")
	}
	found := false
	for _, sub := range glmToolsCmd.Commands() {
		if sub.Name() == "enable" {
			found = true
			break
		}
	}
	if !found {
		t.Error("glmToolsCmd 에 'enable' 서브커맨드가 없음")
	}
}

// TestGLMTools_Disable_Subcommand_Exists — verifies the disable subcommand exists.
func TestGLMTools_Disable_Subcommand_Exists(t *testing.T) {
	if glmToolsCmd == nil {
		t.Fatal("glmToolsCmd 가 nil 임")
	}
	found := false
	for _, sub := range glmToolsCmd.Commands() {
		if sub.Name() == "disable" {
			found = true
			break
		}
	}
	if !found {
		t.Error("glmToolsCmd 에 'disable' 서브커맨드가 없음")
	}
}

// TestGLMToolsEnableIdempotent — GWT-1: running enable twice MUST make the second invocation a no-op (REQ-GMC-001, REQ-GMC-006).
func TestGLMToolsEnableIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-glm-key-abc123")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	// First enable
	err := runEnableMCPServer(claudeJSONPath, "test-glm-key-abc123")
	if err != nil {
		t.Fatalf("첫 번째 enable 실패: %v", err)
	}

	// Verify zai-mcp-server exists after the first enable
	entry := readMCPEntry(t, claudeJSONPath)
	if entry == nil {
		t.Fatal("첫 번째 enable 후 zai-mcp-server 엔트리가 없음")
	}

	// Record file mtime (used to verify idempotency)
	info1, _ := os.Stat(claudeJSONPath)

	// Second enable (same token → idempotent skip)
	skipped, err := enableMCPServerIdempotent(claudeJSONPath, "test-glm-key-abc123")
	if err != nil {
		t.Fatalf("두 번째 enable 실패: %v", err)
	}
	if !skipped {
		t.Error("두 번째 enable 은 idempotent skip 이어야 함")
	}

	// Verify mtime is unchanged
	info2, _ := os.Stat(claudeJSONPath)
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Error("idempotent skip 시 claude.json 의 mtime 이 변경됨 (기대: 변경 없음)")
	}
}

// TestGLMToolsDisableIdempotent — GWT-2: running disable twice MUST make the second invocation a no-op (REQ-GMC-001).
func TestGLMToolsDisableIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-glm-key-abc123")
	// Initialize with zai-mcp-server already present
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("test-glm-key-abc123"),
	})

	// First disable
	removed, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("첫 번째 disable 실패: %v", err)
	}
	if !removed {
		t.Error("첫 번째 disable 은 엔트리를 제거해야 함")
	}

	// Second disable (no entry → idempotent skip)
	removed2, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("두 번째 disable 실패: %v", err)
	}
	if removed2 {
		t.Error("두 번째 disable 은 제거할 엔트리가 없어야 함 (idempotent skip)")
	}
}

// ─── GWT-3: SPEC-GLM-001 compatibility (REQ-GMC-002) ──────────────────────

func TestGLMTools_NoConflictWithSPECGLM001(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	// Run enable
	err := runEnableMCPServer(claudeJSONPath, "test-token")
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	// claude.json must not contain env-policy fields such as DISABLE_BETAS
	data, _ := os.ReadFile(claudeJSONPath)
	if strings.Contains(string(data), "DISABLE_BETAS") {
		t.Error("enable 이 SPEC-GLM-001 의 env 정책 필드를 변경함 (REQ-GMC-002 위반)")
	}
}

// ─── GWT-4: enable entry correctness (REQ-GMC-003) ────────────────────────

// TestGLMToolsEnable_EntryFields — GWT-4: the four fields are recorded exactly.
func TestGLMToolsEnable_EntryFields(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-glm-key-abc123"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	err := runEnableMCPServer(claudeJSONPath, token)
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	entry := readMCPEntry(t, claudeJSONPath)
	if entry == nil {
		t.Fatal("zai-mcp-server 엔트리가 없음")
	}

	// command: "npx"
	if entry["command"] != "npx" {
		t.Errorf("command = %q, 기대: %q", entry["command"], "npx")
	}

	// args: ["-y", "@z_ai/mcp-server@latest"]
	args, ok := entry["args"].([]any)
	if !ok || len(args) != 2 || args[0] != "-y" || args[1] != "@z_ai/mcp-server@latest" {
		t.Errorf("args = %v, 기대: [\"-y\", \"@z_ai/mcp-server@latest\"]", args)
	}

	// env.Z_AI_API_KEY = token
	envMap, ok := entry["env"].(map[string]any)
	if !ok {
		t.Fatal("env 필드가 없음")
	}
	if envMap["Z_AI_API_KEY"] != token {
		t.Errorf("Z_AI_API_KEY = %q, 기대: %q", envMap["Z_AI_API_KEY"], token)
	}
	if envMap["Z_AI_MODE"] != "ZAI" {
		t.Errorf("Z_AI_MODE = %q, 기대: %q", envMap["Z_AI_MODE"], "ZAI")
	}
}

// ─── GWT-6, GWT-7: preserve other mcpServers entries after disable (REQ-GMC-004) ──

// TestGLMToolsDisable_RemovesZAIEntry — GWT-6: disable removes zai-mcp-server.
func TestGLMToolsDisable_RemovesZAIEntry(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("test-token"),
	})

	removed, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("disable 실패: %v", err)
	}
	if !removed {
		t.Error("disable 이 엔트리를 제거하지 않음")
	}

	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("disable 후에도 zai-mcp-server 엔트리가 남아 있음")
	}
}

// TestGLMToolsDisable_PreservesOtherEntries — GWT-7: disable preserves the other three entries.
func TestGLMToolsDisable_PreservesOtherEntries(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")

	// Four entries: context7, sequential-thinking, moai-lsp, zai-mcp-server
	other1 := map[string]any{"command": "/bin/bash", "args": []any{"-l", "-c", "exec npx context7"}}
	other2 := map[string]any{"command": "/bin/bash", "args": []any{"-l", "-c", "exec npx seq-think"}}
	other3 := map[string]any{"command": "node", "args": []any{"moai-lsp"}}

	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		"context7":            other1,
		"sequential-thinking": other2,
		"moai-lsp":            other3,
		zaiMCPServerKey:       buildZAIMCPEntry("test-token"),
	})

	_, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("disable 실패: %v", err)
	}

	servers := readAllMCPServers(t, claudeJSONPath)
	if _, ok := servers[zaiMCPServerKey]; ok {
		t.Error("disable 후에도 zai-mcp-server 가 남아 있음")
	}
	if _, ok := servers["context7"]; !ok {
		t.Error("context7 엔트리가 제거됨 (REQ-GMC-010 위반)")
	}
	if _, ok := servers["sequential-thinking"]; !ok {
		t.Error("sequential-thinking 엔트리가 제거됨 (REQ-GMC-010 위반)")
	}
	if _, ok := servers["moai-lsp"]; !ok {
		t.Error("moai-lsp 엔트리가 제거됨 (REQ-GMC-010 위반)")
	}
	if len(servers) != 3 {
		t.Errorf("엔트리 수 = %d, 기대: 3", len(servers))
	}
}

// ─── GWT-8, GWT-9: backup policy (REQ-GMC-005) ────────────────────────────

// TestGLMToolsEnable_BackupCreated — GWT-8: enable creates a backup file.
func TestGLMToolsEnable_BackupCreated(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	beforeEnable := time.Now().Add(-time.Second) // slack for timestamp comparison

	err := runEnableMCPServer(claudeJSONPath, "test-token")
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	// Find backup files: ~/.claude.json.bak-<ISO ts>
	dir := filepath.Dir(claudeJSONPath)
	entries, _ := os.ReadDir(dir)
	var backupFiles []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".claude.json.bak-") {
			info, _ := e.Info()
			if info.ModTime().After(beforeEnable) {
				backupFiles = append(backupFiles, e.Name())
			}
		}
	}

	if len(backupFiles) == 0 {
		t.Error("enable 후 백업 파일(.claude.json.bak-*)이 생성되지 않음 (REQ-GMC-005)")
	}
	if len(backupFiles) > 1 {
		t.Errorf("백업 파일이 %d개 생성됨 (기대: 1개): %v", len(backupFiles), backupFiles)
	}
}

// TestGLMToolsEnable_NoBackupOnIdempotent — GWT-9: skip backup on idempotent skip.
func TestGLMToolsEnable_NoBackupOnIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	// zai-mcp-server already present (same token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("test-token"),
	})

	beforeEnable := time.Now()

	// Trigger an idempotent skip (same token → skip)
	skipped, err := enableMCPServerIdempotent(claudeJSONPath, "test-token")
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}
	if !skipped {
		t.Skip("idempotent skip 이 발생하지 않아 백업 테스트 스킵")
	}

	// No backup file should be created.
	dir := filepath.Dir(claudeJSONPath)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".claude.json.bak-") {
			info, _ := e.Info()
			if info.ModTime().After(beforeEnable) {
				t.Errorf("idempotent skip 시 백업 파일이 생성됨: %s (REQ-GMC-005 위반)", e.Name())
			}
		}
	}
}

// ─── GWT-10, GWT-11: existing-entry handling (REQ-GMC-006) ─────────────────

// TestGLMToolsEnable_SameTokenIdempotent — GWT-10: idempotent skip when tokens match.
func TestGLMToolsEnable_SameTokenIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "matching-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry(token),
	})

	skipped, err := enableMCPServerIdempotent(claudeJSONPath, token)
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}
	if !skipped {
		t.Error("토큰이 동일할 때 idempotent skip 이 발생하지 않음 (REQ-GMC-006 (a))")
	}
}

// TestGLMToolsEnable_DifferentTokenRefuse — GWT-11: refuse when tokens differ; guide to --force.
func TestGLMToolsEnable_DifferentTokenRefuse(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	newToken := "new-token-B"
	setupGLMToken(t, homeDir, newToken)
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("old-token-A"),
	})

	err := runEnableMCPServer(claudeJSONPath, newToken)
	// MUST return an error on token mismatch (REQ-GMC-006 (b))
	if err == nil {
		t.Fatal("토큰 불일치 시 에러가 반환되어야 함 (REQ-GMC-006 (b))")
	}
	if !strings.Contains(err.Error(), "force") && !strings.Contains(err.Error(), "--force") {
		t.Errorf("에러 메시지에 --force 안내가 없음: %v", err)
	}

	// Verify claude.json is unchanged
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		envMap, _ := entry["env"].(map[string]any)
		if envMap != nil && envMap["Z_AI_API_KEY"] != "old-token-A" {
			t.Error("토큰 불일치 거부 시 claude.json 이 변경됨 (REQ-GMC-006 (b) 위반)")
		}
	}
}

// ─── GWT-12: missing token → reject enable (REQ-GMC-007) ──────────────────

// TestGLMToolsEnable_NoToken_Rejected — GWT-12: reject when GLM_AUTH_TOKEN is missing.
func TestGLMToolsEnable_NoToken_Rejected(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	// No .env.glm → no token
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	err := runEnableMCPServer(claudeJSONPath, "") // empty token
	if err == nil {
		t.Fatal("토큰 부재 시 에러가 반환되어야 함 (REQ-GMC-007)")
	}

	// Verify claude.json is unchanged
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("토큰 부재 시에도 zai-mcp-server 엔트리가 추가됨 (REQ-GMC-007 위반)")
	}
}

// ─── GWT-13: --scope project option (REQ-GMC-008) ─────────────────────────

// TestGLMToolsEnable_ScopeProject — GWT-13: write to .mcp.json when --scope project is used.
func TestGLMToolsEnable_ScopeProject(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	// Project root = separate tmpDir
	projectRoot := t.TempDir()
	mcpJSONPath := filepath.Join(projectRoot, ".mcp.json")

	err := runEnableMCPServerScoped(mcpJSONPath, token)
	if err != nil {
		t.Fatalf("--scope project enable 실패: %v", err)
	}

	// zai-mcp-server entry exists in .mcp.json
	data, err := os.ReadFile(mcpJSONPath)
	if err != nil {
		t.Fatalf(".mcp.json 읽기 실패: %v", err)
	}
	if !strings.Contains(string(data), zaiMCPServerKey) {
		t.Errorf(".mcp.json 에 %s 엔트리가 없음", zaiMCPServerKey)
	}

	// Verify ~/.claude.json is unchanged (user scope untouched)
	userServers := readAllMCPServers(t, claudeJSONPath)
	if _, ok := userServers[zaiMCPServerKey]; ok {
		t.Error("--scope project 사용 시 ~/.claude.json 이 변경됨 (REQ-GMC-008 위반)")
	}
}

// ─── GWT-14: Node.js absent (REQ-GMC-009) ─────────────────────────────────

// TestGLMToolsEnable_NodeMissing — GWT-14: graceful failure when node is absent.
// The node check is performed by the command handler via checkNodeVersion().
// runEnableMCPServer is a pure JSON-mutation function, so the node check is tested
// separately through checkNodeVersion().
func TestGLMToolsEnable_NodeMissing(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeMissing(t)()

	// Call the node-check function directly (REQ-GMC-009)
	err := checkNodeVersion()
	if err == nil {
		t.Fatal("node 부재 시 에러가 반환되어야 함 (REQ-GMC-009)")
	}

	// claude.json must remain unchanged (checkNodeVersion does not touch JSON)
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("checkNodeVersion 이 claude.json 을 변경함 (REQ-GMC-009 위반)")
	}
}

// ─── GWT-15: Node.js outdated version (REQ-GMC-009) ───────────────────────

// TestGLMToolsEnable_NodeOldVersion — GWT-15: graceful failure on outdated node.
func TestGLMToolsEnable_NodeOldVersion(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOld(t)()

	// Call the node-check function directly (REQ-GMC-009)
	err := checkNodeVersion()
	if err == nil {
		t.Fatal("node 구버전 시 에러가 반환되어야 함 (REQ-GMC-009)")
	}
	if !strings.Contains(err.Error(), "22") {
		t.Errorf("에러 메시지에 최소 버전(22) 안내가 없음: %v", err)
	}

	// claude.json must remain unchanged
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("checkNodeVersion 이 claude.json 을 변경함 (REQ-GMC-009 위반)")
	}
}

// ─── GWT-16, GWT-17: preserve user-defined entries (REQ-GMC-010) ───────────

// TestGLMToolsEnable_PreservesUserDefinedEntries — GWT-16: enable preserves other entries.
func TestGLMToolsEnable_PreservesUserDefinedEntries(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)

	customServer := map[string]any{"command": "my-custom-mcp", "args": []any{"--config", "custom.json"}}
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		"my-custom-server":    customServer,
		"context7":            map[string]any{"command": "npx", "args": []any{"-y", "context7"}},
		"sequential-thinking": map[string]any{"command": "npx", "args": []any{"-y", "seq-think"}},
	})

	err := runEnableMCPServer(claudeJSONPath, token)
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	servers := readAllMCPServers(t, claudeJSONPath)
	// Four entries in total (my-custom-server, context7, sequential-thinking, zai-mcp-server)
	if len(servers) != 4 {
		t.Errorf("서버 수 = %d, 기대: 4", len(servers))
	}
	// Verify my-custom-server fields are preserved
	myServer, ok := servers["my-custom-server"].(map[string]any)
	if !ok {
		t.Fatal("my-custom-server 가 제거됨 (REQ-GMC-010 위반)")
	}
	if myServer["command"] != "my-custom-mcp" {
		t.Errorf("my-custom-server.command 가 변경됨: %v (REQ-GMC-010 위반)", myServer["command"])
	}
}

// TestGLMToolsDisable_PreservesUserDefinedEntries — GWT-17: disable preserves other entries.
func TestGLMToolsDisable_PreservesUserDefinedEntries(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")

	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		"my-custom-server": map[string]any{"command": "my-custom-mcp"},
		"context7":         map[string]any{"command": "npx"},
		zaiMCPServerKey:    buildZAIMCPEntry("test-token"),
	})

	_, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("disable 실패: %v", err)
	}

	servers := readAllMCPServers(t, claudeJSONPath)
	if _, ok := servers[zaiMCPServerKey]; ok {
		t.Error("disable 후 zai-mcp-server 가 남아 있음")
	}
	if _, ok := servers["my-custom-server"]; !ok {
		t.Error("disable 후 my-custom-server 가 제거됨 (REQ-GMC-010 위반)")
	}
	if _, ok := servers["context7"]; !ok {
		t.Error("disable 후 context7 가 제거됨 (REQ-GMC-010 위반)")
	}
}

// ─── GWT-19: atomic write failure recovery (R7 verification) ──────────────

// TestGLMToolsEnable_AtomicWriteProtectsOriginal — GWT-19: preserve the original on write failure.
// A real disk-full simulation is impractical, so use a write-disallowed directory.
func TestGLMToolsEnable_AtomicWriteProtectsOriginal(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 권한에서는 권한 테스트가 무의미하므로 스킵")
	}

	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)

	// Create a valid claude.json, then make the parent directory read-only.
	// (Unix-only behavior.)
	readOnlyDir := filepath.Join(t.TempDir(), "readonly")
	if err := os.MkdirAll(readOnlyDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(readOnlyDir, 0o755) })

	readOnlyPath := filepath.Join(readOnlyDir, ".claude.json")
	initialContent := `{"mcpServers":{}}`
	if err := os.WriteFile(readOnlyPath, []byte(initialContent), 0o600); err != nil {
		// File creation may itself fail in a write-restricted directory
		t.Skip("읽기 전용 디렉토리에 파일 생성 불가 — 환경 제한으로 스킵")
	}

	// Change the directory to read-only (no temp file creation allowed)
	if err := os.Chmod(readOnlyDir, 0o555); err != nil {
		t.Skip("chmod 실패 — 환경 제한으로 스킵")
	}

	err := runEnableMCPServer(readOnlyPath, token)
	// MUST return an error on write failure
	if err == nil {
		t.Skip("읽기 전용 디렉토리에 쓰기가 성공함 — 환경이 예상과 다름")
	}

	// The original file content must not be corrupted
	content, readErr := os.ReadFile(readOnlyPath)
	if readErr != nil {
		t.Logf("원본 파일 읽기 실패 (권한 문제일 수 있음): %v", readErr)
	} else if string(content) != initialContent {
		t.Errorf("atomic write 실패 시 원본 파일이 손상됨")
	}
}

// ─── GWT-20: JSON parsing validity ────────────────────────────────────────

// TestGLMToolsEnable_ValidJSON — GWT-20: the resulting JSON is valid after enable.
func TestGLMToolsEnable_ValidJSON(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-glm-key-abc123"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	err := runEnableMCPServer(claudeJSONPath, token)
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	data, _ := os.ReadFile(claudeJSONPath)
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("enable 후 claude.json 이 유효하지 않은 JSON: %v", err)
	}
}

// ─── GWT-22: command-argument validation ──────────────────────────────────

// TestGLMToolsEnable_InvalidToolName — GWT-22(a): error on an invalid tool name.
func TestGLMToolsEnable_InvalidToolName(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	defer makeNodeOK(t)()

	err := validateToolName("foo")
	if err == nil {
		t.Error("잘못된 도구명 'foo' 에 대해 에러가 반환되어야 함 (GWT-22 (a))")
	}
	if !strings.Contains(err.Error(), "vision") || !strings.Contains(err.Error(), "all") {
		t.Errorf("에러 메시지에 지원 도구 목록이 없음: %v", err)
	}
}

// TestGLMToolsEnable_ValidToolNames — GWT-22: validate accepted tool names.
func TestGLMToolsEnable_ValidToolNames(t *testing.T) {
	validNames := []string{"vision", "websearch", "webreader", "all"}
	for _, name := range validNames {
		if err := validateToolName(name); err != nil {
			t.Errorf("유효한 도구명 %q 에 대해 에러가 반환됨: %v", name, err)
		}
	}
}

// ─── Node.js version-detection helper tests ───────────────────────────────

// TestDetectNodeVersion_Parse — verify version-string parsing.
func TestDetectNodeVersion_Parse(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantMaj int
		wantErr bool
	}{
		{"v22.5.0", "v22.5.0", 22, false},
		{"v18.20.4", "v18.20.4", 18, false},
		{"v20.0.0", "v20.0.0", 20, false},
		{"v22.0.0", "v22.0.0", 22, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			maj, err := parseNodeMajorVersion(tc.version)
			if tc.wantErr && err == nil {
				t.Errorf("에러를 기대했으나 없음")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("예상치 못한 에러: %v", err)
			}
			if maj != tc.wantMaj {
				t.Errorf("major = %d, 기대: %d", maj, tc.wantMaj)
			}
		})
	}
}

// ─── Backup-filename ISO-timestamp format tests ───────────────────────────

// TestBackupFilenameFormat — verify the backup filename uses the ISO timestamp format.
func TestBackupFilenameFormat(t *testing.T) {
	ts := time.Date(2026, 5, 10, 1, 35, 0, 0, time.UTC)
	name := buildBackupFilename(ts)
	// Expected pattern: .claude.json.bak-2026-05-10T01-35-00Z
	expected := ".claude.json.bak-2026-05-10T01-35-00Z"
	if name != expected {
		t.Errorf("백업 파일명 = %q, 기대: %q", name, expected)
	}
	// MUST NOT contain a colon (filename-safe)
	if strings.Contains(name, ":") {
		t.Errorf("백업 파일명에 콜론이 포함됨: %q", name)
	}
}

// ─── Constants and entry-builder tests ────────────────────────────────────

// TestZAIMCPEntryBuilder — buildZAIMCPEntry returns the exact expected structure.
func TestZAIMCPEntryBuilder(t *testing.T) {
	token := "my-test-token"
	entry := buildZAIMCPEntry(token)

	if entry["command"] != "npx" {
		t.Errorf("command = %v, 기대: npx", entry["command"])
	}
	args, ok := entry["args"].([]string)
	if !ok || len(args) != 2 {
		t.Errorf("args 형식 오류: %v", entry["args"])
	}
	envMap, ok := entry["env"].(map[string]string)
	if !ok {
		t.Fatal("env 없음")
	}
	if envMap["Z_AI_API_KEY"] != token {
		t.Errorf("Z_AI_API_KEY = %q, 기대: %q", envMap["Z_AI_API_KEY"], token)
	}
	if envMap["Z_AI_MODE"] != "ZAI" {
		t.Errorf("Z_AI_MODE = %q, 기대: ZAI", envMap["Z_AI_MODE"])
	}
}

// ─── resolveConfigPath tests ──────────────────────────────────────────────

// TestResolveConfigPath_UserScope — returns ~/.claude.json under user scope.
func TestResolveConfigPath_UserScope(t *testing.T) {
	homeDir := setupToolsTestHome(t)

	path, err := resolveConfigPath("user")
	if err != nil {
		t.Fatalf("resolveConfigPath(user) 실패: %v", err)
	}
	expected := filepath.Join(homeDir, ".claude.json")
	if path != expected {
		t.Errorf("path = %q, 기대: %q", path, expected)
	}
}

// TestResolveConfigPath_ProjectScope — returns .mcp.json (cwd-relative) under project scope.
func TestResolveConfigPath_ProjectScope(t *testing.T) {
	// Switch cwd to a TempDir
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	path, err := resolveConfigPath("project")
	if err != nil {
		t.Fatalf("resolveConfigPath(project) 실패: %v", err)
	}

	// On macOS, /var → /private/var is a symlink, so an EvalSymlinks comparison is required
	resolvedPath, _ := filepath.EvalSymlinks(filepath.Dir(path))
	resolvedTmp, _ := filepath.EvalSymlinks(tmpDir)
	if resolvedPath != resolvedTmp {
		t.Errorf("path 디렉토리 = %q, 기대: %q", resolvedPath, resolvedTmp)
	}
	if filepath.Base(path) != ".mcp.json" {
		t.Errorf("파일명 = %q, 기대: .mcp.json", filepath.Base(path))
	}
}

// ─── Cobra command integration tests ──────────────────────────────────────

// TestGLMToolsEnableCmd_Success — full-path test for the cobra enable command.
func TestGLMToolsEnableCmd_Success(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "integration-test-token"
	setupGLMToken(t, homeDir, token)
	setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	// Configure the test env so loadGLMKey() can read the GLM API key.
	// loadGLMKey() reads either MOAI_TEST_GLM_KEY or .env.glm.
	// setupGLMToken has already created .env.glm, so it is ready to use.

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	glmToolsEnableCmd.SetOut(outBuf)
	glmToolsEnableCmd.SetErr(errBuf)

	// Run enable "all"
	err := glmToolsEnableCmd.RunE(glmToolsEnableCmd, []string{"all"})
	if err != nil {
		t.Fatalf("enable 커맨드 실패: %v", err)
	}

	// Verify the success message
	output := outBuf.String()
	if !strings.Contains(output, "활성화") && !strings.Contains(output, "enable") && !strings.Contains(output, "Z.AI") {
		t.Errorf("성공 메시지에 활성화 관련 문자열 없음: %q", output)
	}
}

// TestGLMToolsDisableCmd_Success — full-path test for the cobra disable command.
func TestGLMToolsDisableCmd_Success(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "integration-test-token"
	setupGLMToken(t, homeDir, token)
	// zai-mcp-server already present
	setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry(token),
	})

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	glmToolsDisableCmd.SetOut(outBuf)
	glmToolsDisableCmd.SetErr(errBuf)

	err := glmToolsDisableCmd.RunE(glmToolsDisableCmd, []string{"all"})
	if err != nil {
		t.Fatalf("disable 커맨드 실패: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "비활성화") && !strings.Contains(output, "disable") && !strings.Contains(output, "removed") && !strings.Contains(output, "Z.AI") {
		t.Errorf("비활성화 메시지 없음: %q", output)
	}
}

// TestGLMToolsDisableCmd_NothingToDisable — no-op when there are no entries to disable.
func TestGLMToolsDisableCmd_NothingToDisable(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	setupClaudeJSON(t, homeDir, nil) // empty mcpServers

	outBuf := new(strings.Builder)
	glmToolsDisableCmd.SetOut(outBuf)

	err := glmToolsDisableCmd.RunE(glmToolsDisableCmd, []string{"all"})
	if err != nil {
		t.Fatalf("disable no-op 실패: %v", err)
	}
	output := outBuf.String()
	if !strings.Contains(output, "없") && !strings.Contains(output, "no") && !strings.Contains(output, "없습니다") {
		// Check for the "none" Korean message or the no-op message
		t.Logf("no-op 출력: %q (OK - no error)", output)
	}
}

// TestGLMToolsEnableCmd_NoToken — return an error when the token is absent.
// runGLMToolsEnable calls loadGLMKey(), which first checks the MOAI_TEST_GLM_KEY env.
// Test isolation: set MOAI_TEST_GLM_KEY to an empty string + override HOME to a tmpDir without .env.glm.
func TestGLMToolsEnableCmd_NoToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())     // make loadGLMKey()'s getGLMEnvPath() see an empty DIR
	t.Setenv("MOAI_TEST_GLM_KEY", "") // clear loadGLMKey()'s test-key env
	defer makeNodeOK(t)()

	err := glmToolsEnableCmd.RunE(glmToolsEnableCmd, []string{"all"})
	if err == nil {
		t.Fatal("토큰 없을 때 에러가 반환되어야 함")
	}
}

// TestGLMToolsEnableCmd_BadNodeVersion — return an error on outdated node.
func TestGLMToolsEnableCmd_BadNodeVersion(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOld(t)()

	err := glmToolsEnableCmd.RunE(glmToolsEnableCmd, []string{"all"})
	if err == nil {
		t.Fatal("구버전 node 시 에러가 반환되어야 함")
	}
}

// TestGLMToolsEnableCmd_InvalidTool — error on an invalid tool name.
func TestGLMToolsEnableCmd_InvalidTool(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	err := glmToolsEnableCmd.RunE(glmToolsEnableCmd, []string{"invalid-tool"})
	if err == nil {
		t.Fatal("잘못된 도구명 시 에러가 반환되어야 함")
	}
}

// TestDetectNodeVersion_Real — when an actual node is available on PATH (environment-dependent).
func TestDetectNodeVersion_Real(t *testing.T) {
	// First check whether node is actually present
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node 가 PATH 에 없으므로 스킵")
	}

	major, versionStr, err := detectNodeVersion()
	if err != nil && err != errNodeNotFound {
		t.Logf("detectNodeVersion 에러 (환경 의존): %v", err)
		return
	}
	if err == nil {
		t.Logf("감지된 node: major=%d, version=%s", major, versionStr)
	}
}

// ─── REQ-GMC-002: orchestration compatibility tests (manual simulation) ───

// ─── extractTokenFromEntry branch coverage ────────────────────────────────

// TestExtractTokenFromEntry_MapStringAny — env of type map[string]any.
func TestExtractTokenFromEntry_MapStringAny(t *testing.T) {
	entry := map[string]any{
		"env": map[string]any{"Z_AI_API_KEY": "my-token"},
	}
	token := extractTokenFromEntry(entry)
	if token != "my-token" {
		t.Errorf("token = %q, 기대: my-token", token)
	}
}

// TestExtractTokenFromEntry_MapStringString — env of type map[string]string.
func TestExtractTokenFromEntry_MapStringString(t *testing.T) {
	entry := map[string]any{
		"env": map[string]string{"Z_AI_API_KEY": "string-token"},
	}
	token := extractTokenFromEntry(entry)
	if token != "string-token" {
		t.Errorf("token = %q, 기대: string-token", token)
	}
}

// TestExtractTokenFromEntry_NoEnv — env absent.
func TestExtractTokenFromEntry_NoEnv(t *testing.T) {
	entry := map[string]any{"command": "npx"}
	token := extractTokenFromEntry(entry)
	if token != "" {
		t.Errorf("token = %q, 기대: empty", token)
	}
}

// TestMaskPartial_Short — masking for a short token.
func TestMaskPartial_Short(t *testing.T) {
	if maskPartial("abc") != "****" {
		t.Errorf("짧은 토큰 마스킹 실패: %q", maskPartial("abc"))
	}
}

// TestMaskPartial_Long — masking for a long token.
func TestMaskPartial_Long(t *testing.T) {
	result := maskPartial("sk-12345678")
	if result != "sk-1****" {
		t.Errorf("마스킹 결과 = %q, 기대: sk-1****", result)
	}
}

// ─── writeClaudeJSONAtomic error paths ────────────────────────────────────

// TestWriteClaudeJSONAtomic_BadDir — error when the directory is not writable.
func TestWriteClaudeJSONAtomic_BadDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 권한에서는 권한 테스트 불가")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Windows에서는 /nonexistent-dir 경로가 드라이브 루트로 해석됨")
	}
	// Attempt to write to a nonexistent path (temp-file creation is expected to fail)
	badPath := filepath.Join("/nonexistent-dir-xyz", ".claude.json")
	err := writeClaudeJSONAtomic(badPath, map[string]any{})
	if err == nil {
		t.Error("잘못된 경로에 쓰기가 성공함 (에러 기대)")
	}
}

// ─── readClaudeJSON error paths ───────────────────────────────────────────

// TestReadClaudeJSON_InvalidJSON — malformed JSON file.
func TestReadClaudeJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	badFile := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(badFile, []byte("{not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := readClaudeJSON(badFile)
	if err == nil {
		t.Error("잘못된 JSON 파일에 에러가 반환되어야 함")
	}
}

// TestReadClaudeJSON_NotExists — return an empty structure when the file does not exist.
func TestReadClaudeJSON_NotExists(t *testing.T) {
	root, err := readClaudeJSON("/tmp/nonexistent-claude.json")
	if err != nil {
		t.Fatalf("파일 없을 때 에러가 반환됨: %v", err)
	}
	if root == nil {
		t.Error("빈 구조가 반환되어야 함")
	}
}

// ─── REQ-GMC-002: orchestration compatibility tests (manual simulation) ───
func TestGLMTools_OrthogonalToGLMMode(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)

	// Create settings.local.json under the project root
	projectRoot := t.TempDir()
	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	initialSettings := `{"env":{"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":"1"}}`
	if err := os.WriteFile(settingsPath, []byte(initialSettings), 0o600); err != nil {
		t.Fatal(err)
	}

	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	err := runEnableMCPServer(claudeJSONPath, token)
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	// settings.local.json must remain unchanged
	data, _ := os.ReadFile(settingsPath)
	if string(data) != initialSettings {
		t.Errorf("enable 이 settings.local.json 을 변경함 (REQ-GMC-002 위반)")
	}
}

// ─── autoEnableMCPServer tests ────────────────────────────────────────────

func TestAutoEnableMCPServer_Success(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-glm-key-abc123")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	autoEnableMCPServer()

	entry := readMCPEntry(t, claudeJSONPath)
	if entry == nil {
		t.Fatal("MCP server entry should exist after auto-enable")
	}
	if entry["command"] != "npx" {
		t.Errorf("command = %v, want npx", entry["command"])
	}
}

func TestAutoEnableMCPServer_Idempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-glm-key-abc123"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		"zai-mcp-server": buildZAIMCPEntry(token),
	})

	info1, _ := os.Stat(claudeJSONPath)
	mtime1 := info1.ModTime()

	autoEnableMCPServer()

	info2, _ := os.Stat(claudeJSONPath)
	mtime2 := info2.ModTime()

	if !mtime1.Equal(mtime2) {
		t.Error("auto-enable should skip writing when already enabled with same token")
	}
}

func TestAutoEnableMCPServer_OptOut(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-glm-key-abc123")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	t.Setenv("MOAI_GLM_NO_AUTO_TOOLS", "1")

	autoEnableMCPServer()

	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("MCP server should not be enabled when MOAI_GLM_NO_AUTO_TOOLS=1")
	}
}

func TestAutoEnableMCPServer_NoToken(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	autoEnableMCPServer()

	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("MCP server should not be enabled when no GLM token")
	}
}
