package cli

// glm_tools_test.go — SPEC-GLM-MCP-001 GWT-1 ~ GWT-22 자동화 테스트
// RED 단계: glm_tools.go 구현 전 작성. 모든 테스트가 실패해야 함.
// 참고: CLAUDE.local.md §6 (t.TempDir() 격리, 병렬 테스트 절대 t.Setenv("HOME",...) 금지)
//
// 테스트 격리 전략:
//   - HOME 은 t.Setenv("HOME", tmpDir) 로 덮어쓰지 않음 (병렬 안전성 우선)
//   - userHomeDirFn 함수 변수를 통해 home 디렉토리를 주입 (glm_tools.go 에서 선언)
//   - detectNodeFn 함수 변수를 통해 node 버전 감지 모킹
//   - 모든 파일 조작은 t.TempDir() 내부에서 수행

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

// ─── 테스트 헬퍼 ───────────────────────────────────────────────────────────

// setupToolsTestHome 은 테스트용 가짜 HOME 디렉토리를 생성하고
// userHomeDirFn 을 해당 디렉토리로 오버라이드한다.
// GLM 토큰은 ~/.moai/.env.glm 에 저장된다.
func setupToolsTestHome(t *testing.T) (homeDir string) {
	t.Helper()
	tmpDir := t.TempDir()

	// userHomeDirFn 오버라이드
	origFn := userHomeDirFn
	userHomeDirFn = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { userHomeDirFn = origFn })

	return tmpDir
}

// setupGLMToken 은 지정한 홈 디렉토리의 ~/.moai/.env.glm 에 토큰을 저장한다.
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

// setupClaudeJSON 은 지정한 홈 디렉토리의 ~/.claude.json 을 초기화한다.
// mcpServers 는 초기 MCP 서버 맵이다. nil 이면 빈 mcpServers 로 초기화.
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

// readMCPEntry 는 ~/.claude.json 의 mcpServers.zai-mcp-server 엔트리를 읽는다.
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

// readAllMCPServers 는 ~/.claude.json 의 전체 mcpServers 맵을 읽는다.
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

// makeNodeOK 은 Node.js v22.x 를 시뮬레이션하는 detectNodeFn 오버라이드를 반환한다.
func makeNodeOK(t *testing.T) func() {
	t.Helper()
	origFn := detectNodeFn
	detectNodeFn = func() (int, string, error) { return 22, "v22.5.0", nil }
	return func() { detectNodeFn = origFn }
}

// makeNodeMissing 은 node 명령이 PATH 에 없는 상황을 시뮬레이션한다.
func makeNodeMissing(t *testing.T) func() {
	t.Helper()
	origFn := detectNodeFn
	detectNodeFn = func() (int, string, error) {
		return 0, "", errNodeNotFound
	}
	return func() { detectNodeFn = origFn }
}

// makeNodeOld 는 node 구버전(v18.20.4)을 시뮬레이션한다.
func makeNodeOld(t *testing.T) func() {
	t.Helper()
	origFn := detectNodeFn
	detectNodeFn = func() (int, string, error) { return 18, "v18.20.4", nil }
	return func() { detectNodeFn = origFn }
}

// ─── GWT-1: tools enable subcommand 라우팅 + idempotent ──────────────────

// TestGLMTools_Cmd_Exists — glmToolsCmd 가 glmCmd 의 서브커맨드로 등록됨을 검증 (REQ-GMC-001)
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

// TestGLMTools_Enable_Subcommand_Exists — enable 서브커맨드가 존재함을 검증
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

// TestGLMTools_Disable_Subcommand_Exists — disable 서브커맨드가 존재함을 검증
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

// TestGLMToolsEnableIdempotent — GWT-1: enable 두 번 실행 시 두 번째는 no-op (REQ-GMC-001, REQ-GMC-006)
func TestGLMToolsEnableIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-glm-key-abc123")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	// 첫 번째 enable
	err := runEnableMCPServer(claudeJSONPath, "test-glm-key-abc123")
	if err != nil {
		t.Fatalf("첫 번째 enable 실패: %v", err)
	}

	// 첫 번째 이후 zai-mcp-server 존재 확인
	entry := readMCPEntry(t, claudeJSONPath)
	if entry == nil {
		t.Fatal("첫 번째 enable 후 zai-mcp-server 엔트리가 없음")
	}

	// 파일 mtime 기록 (idempotent 검증용)
	info1, _ := os.Stat(claudeJSONPath)

	// 두 번째 enable (same token → idempotent skip)
	skipped, err := enableMCPServerIdempotent(claudeJSONPath, "test-glm-key-abc123")
	if err != nil {
		t.Fatalf("두 번째 enable 실패: %v", err)
	}
	if !skipped {
		t.Error("두 번째 enable 은 idempotent skip 이어야 함")
	}

	// mtime 변경 없음 확인
	info2, _ := os.Stat(claudeJSONPath)
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Error("idempotent skip 시 claude.json 의 mtime 이 변경됨 (기대: 변경 없음)")
	}
}

// TestGLMToolsDisableIdempotent — GWT-2: disable 두 번 실행 시 두 번째는 no-op (REQ-GMC-001)
func TestGLMToolsDisableIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-glm-key-abc123")
	// zai-mcp-server 가 이미 있는 상태로 초기화
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("test-glm-key-abc123"),
	})

	// 첫 번째 disable
	removed, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("첫 번째 disable 실패: %v", err)
	}
	if !removed {
		t.Error("첫 번째 disable 은 엔트리를 제거해야 함")
	}

	// 두 번째 disable (엔트리 없음 → idempotent skip)
	removed2, err := disableMCPServerSafe(claudeJSONPath)
	if err != nil {
		t.Fatalf("두 번째 disable 실패: %v", err)
	}
	if removed2 {
		t.Error("두 번째 disable 은 제거할 엔트리가 없어야 함 (idempotent skip)")
	}
}

// ─── GWT-3: SPEC-GLM-001 호환성 (REQ-GMC-002) ─────────────────────────────

func TestGLMTools_NoConflictWithSPECGLM001(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	// enable 실행
	err := runEnableMCPServer(claudeJSONPath, "test-token")
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	// claude.json 에 DISABLE_BETAS 등 환경변수 관련 필드가 없어야 함
	data, _ := os.ReadFile(claudeJSONPath)
	if strings.Contains(string(data), "DISABLE_BETAS") {
		t.Error("enable 이 SPEC-GLM-001 의 env 정책 필드를 변경함 (REQ-GMC-002 위반)")
	}
}

// ─── GWT-4: enable 엔트리 정확성 (REQ-GMC-003) ───────────────────────────

// TestGLMToolsEnable_EntryFields — GWT-4: 4개 필드가 정확히 기록됨
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

// ─── GWT-6, GWT-7: disable 후 다른 mcpServers 엔트리 보존 (REQ-GMC-004) ──

// TestGLMToolsDisable_RemovesZAIEntry — GWT-6: disable 시 zai-mcp-server 제거
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

// TestGLMToolsDisable_PreservesOtherEntries — GWT-7: disable 시 다른 3개 엔트리 보존
func TestGLMToolsDisable_PreservesOtherEntries(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")

	// 4개 엔트리: context7, sequential-thinking, moai-lsp, zai-mcp-server
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

// ─── GWT-8, GWT-9: 백업 정책 (REQ-GMC-005) ───────────────────────────────

// TestGLMToolsEnable_BackupCreated — GWT-8: enable 시 백업 파일 생성
func TestGLMToolsEnable_BackupCreated(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	beforeEnable := time.Now().Add(-time.Second) // 타임스탬프 비교용 여유분

	err := runEnableMCPServer(claudeJSONPath, "test-token")
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}

	// 백업 파일 탐색: ~/.claude.json.bak-<ISO ts>
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

// TestGLMToolsEnable_NoBackupOnIdempotent — GWT-9: idempotent skip 시 백업 생략
func TestGLMToolsEnable_NoBackupOnIdempotent(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	// zai-mcp-server 가 이미 있는 상태 (같은 토큰)
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("test-token"),
	})

	beforeEnable := time.Now()

	// idempotent skip 트리거 (같은 토큰 → skip)
	skipped, err := enableMCPServerIdempotent(claudeJSONPath, "test-token")
	if err != nil {
		t.Fatalf("enable 실패: %v", err)
	}
	if !skipped {
		t.Skip("idempotent skip 이 발생하지 않아 백업 테스트 스킵")
	}

	// 백업 파일이 생성되지 않아야 함
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

// ─── GWT-10, GWT-11: 기존 엔트리 처리 (REQ-GMC-006) ─────────────────────

// TestGLMToolsEnable_SameTokenIdempotent — GWT-10: 토큰 일치 시 idempotent skip
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

// TestGLMToolsEnable_DifferentTokenRefuse — GWT-11: 토큰 불일치 시 거부 + --force 안내
func TestGLMToolsEnable_DifferentTokenRefuse(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	newToken := "new-token-B"
	setupGLMToken(t, homeDir, newToken)
	claudeJSONPath := setupClaudeJSON(t, homeDir, map[string]any{
		zaiMCPServerKey: buildZAIMCPEntry("old-token-A"),
	})

	err := runEnableMCPServer(claudeJSONPath, newToken)
	// 토큰 불일치 시 반드시 에러 반환 (REQ-GMC-006 (b))
	if err == nil {
		t.Fatal("토큰 불일치 시 에러가 반환되어야 함 (REQ-GMC-006 (b))")
	}
	if !strings.Contains(err.Error(), "force") && !strings.Contains(err.Error(), "--force") {
		t.Errorf("에러 메시지에 --force 안내가 없음: %v", err)
	}

	// claude.json 변경 없음 확인
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		envMap, _ := entry["env"].(map[string]any)
		if envMap != nil && envMap["Z_AI_API_KEY"] != "old-token-A" {
			t.Error("토큰 불일치 거부 시 claude.json 이 변경됨 (REQ-GMC-006 (b) 위반)")
		}
	}
}

// ─── GWT-12: 토큰 부재 → enable 거부 (REQ-GMC-007) ───────────────────────

// TestGLMToolsEnable_NoToken_Rejected — GWT-12: GLM_AUTH_TOKEN 부재 시 거부
func TestGLMToolsEnable_NoToken_Rejected(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	// .env.glm 미생성 → 토큰 없음
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	err := runEnableMCPServer(claudeJSONPath, "") // 빈 토큰
	if err == nil {
		t.Fatal("토큰 부재 시 에러가 반환되어야 함 (REQ-GMC-007)")
	}

	// claude.json 변경 없음 확인
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("토큰 부재 시에도 zai-mcp-server 엔트리가 추가됨 (REQ-GMC-007 위반)")
	}
}

// ─── GWT-13: --scope project 옵션 (REQ-GMC-008) ──────────────────────────

// TestGLMToolsEnable_ScopeProject — GWT-13: --scope project 시 .mcp.json 에 기록
func TestGLMToolsEnable_ScopeProject(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)

	// 프로젝트 루트 = 별도 tmpDir
	projectRoot := t.TempDir()
	mcpJSONPath := filepath.Join(projectRoot, ".mcp.json")

	err := runEnableMCPServerScoped(mcpJSONPath, token)
	if err != nil {
		t.Fatalf("--scope project enable 실패: %v", err)
	}

	// .mcp.json 에 zai-mcp-server 엔트리 존재
	data, err := os.ReadFile(mcpJSONPath)
	if err != nil {
		t.Fatalf(".mcp.json 읽기 실패: %v", err)
	}
	if !strings.Contains(string(data), zaiMCPServerKey) {
		t.Errorf(".mcp.json 에 %s 엔트리가 없음", zaiMCPServerKey)
	}

	// ~/.claude.json 변경 없음 확인 (user scope 미터치)
	userServers := readAllMCPServers(t, claudeJSONPath)
	if _, ok := userServers[zaiMCPServerKey]; ok {
		t.Error("--scope project 사용 시 ~/.claude.json 이 변경됨 (REQ-GMC-008 위반)")
	}
}

// ─── GWT-14: Node.js 부재 (REQ-GMC-009) ──────────────────────────────────

// TestGLMToolsEnable_NodeMissing — GWT-14: node 부재 시 graceful fail
// Node 체크는 checkNodeVersion() 을 통해 커맨드 핸들러가 수행.
// runEnableMCPServer 는 순수 JSON 조작 함수이므로 node 체크는 checkNodeVersion() 으로 별도 테스트.
func TestGLMToolsEnable_NodeMissing(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeMissing(t)()

	// Node 체크 함수 직접 호출 (REQ-GMC-009)
	err := checkNodeVersion()
	if err == nil {
		t.Fatal("node 부재 시 에러가 반환되어야 함 (REQ-GMC-009)")
	}

	// claude.json 이 변경되지 않아야 함 (checkNodeVersion 은 JSON 건드리지 않음)
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("checkNodeVersion 이 claude.json 을 변경함 (REQ-GMC-009 위반)")
	}
}

// ─── GWT-15: Node.js 구버전 (REQ-GMC-009) ────────────────────────────────

// TestGLMToolsEnable_NodeOldVersion — GWT-15: node 구버전 시 graceful fail
func TestGLMToolsEnable_NodeOldVersion(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)
	claudeJSONPath := setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOld(t)()

	// Node 체크 함수 직접 호출 (REQ-GMC-009)
	err := checkNodeVersion()
	if err == nil {
		t.Fatal("node 구버전 시 에러가 반환되어야 함 (REQ-GMC-009)")
	}
	if !strings.Contains(err.Error(), "22") {
		t.Errorf("에러 메시지에 최소 버전(22) 안내가 없음: %v", err)
	}

	// claude.json 이 변경되지 않아야 함
	entry := readMCPEntry(t, claudeJSONPath)
	if entry != nil {
		t.Error("checkNodeVersion 이 claude.json 을 변경함 (REQ-GMC-009 위반)")
	}
}

// ─── GWT-16, GWT-17: 사용자 정의 엔트리 보존 (REQ-GMC-010) ──────────────

// TestGLMToolsEnable_PreservesUserDefinedEntries — GWT-16: enable 시 다른 엔트리 보존
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
	// 총 4개 (my-custom-server, context7, sequential-thinking, zai-mcp-server)
	if len(servers) != 4 {
		t.Errorf("서버 수 = %d, 기대: 4", len(servers))
	}
	// my-custom-server 필드 보존 검증
	myServer, ok := servers["my-custom-server"].(map[string]any)
	if !ok {
		t.Fatal("my-custom-server 가 제거됨 (REQ-GMC-010 위반)")
	}
	if myServer["command"] != "my-custom-mcp" {
		t.Errorf("my-custom-server.command 가 변경됨: %v (REQ-GMC-010 위반)", myServer["command"])
	}
}

// TestGLMToolsDisable_PreservesUserDefinedEntries — GWT-17: disable 시 다른 엔트리 보존
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

// ─── GWT-19: atomic write 실패 시 복구 (R7 검증) ─────────────────────────

// TestGLMToolsEnable_AtomicWriteProtectsOriginal — GWT-19: write 실패 시 원본 보존
// 실제 디스크 풀 시뮬레이션은 어려우므로, 쓰기 불가 디렉토리를 사용.
func TestGLMToolsEnable_AtomicWriteProtectsOriginal(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 권한에서는 권한 테스트가 무의미하므로 스킵")
	}

	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)

	// 정상 claude.json 을 만든 후, 부모 디렉토리를 읽기 전용으로 만듦
	// (Unix 에서만 동작)
	readOnlyDir := filepath.Join(t.TempDir(), "readonly")
	if err := os.MkdirAll(readOnlyDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(readOnlyDir, 0o755) })

	readOnlyPath := filepath.Join(readOnlyDir, ".claude.json")
	initialContent := `{"mcpServers":{}}`
	if err := os.WriteFile(readOnlyPath, []byte(initialContent), 0o600); err != nil {
		// 쓰기 전용 디렉토리라 파일 자체 생성도 안 될 수 있음
		t.Skip("읽기 전용 디렉토리에 파일 생성 불가 — 환경 제한으로 스킵")
	}

	// 디렉토리를 읽기 전용으로 변경 (temp file 생성 불가)
	if err := os.Chmod(readOnlyDir, 0o555); err != nil {
		t.Skip("chmod 실패 — 환경 제한으로 스킵")
	}

	err := runEnableMCPServer(readOnlyPath, token)
	// 쓰기 실패 시 에러 반환 필수
	if err == nil {
		t.Skip("읽기 전용 디렉토리에 쓰기가 성공함 — 환경이 예상과 다름")
	}

	// 원본 파일 내용이 손상되지 않음
	content, readErr := os.ReadFile(readOnlyPath)
	if readErr != nil {
		t.Logf("원본 파일 읽기 실패 (권한 문제일 수 있음): %v", readErr)
	} else if string(content) != initialContent {
		t.Errorf("atomic write 실패 시 원본 파일이 손상됨")
	}
}

// ─── GWT-20: JSON 파싱 유효성 ────────────────────────────────────────────

// TestGLMToolsEnable_ValidJSON — GWT-20: enable 후 결과 JSON 이 유효함
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

// ─── GWT-22: 명령 인자 검증 ───────────────────────────────────────────────

// TestGLMToolsEnable_InvalidToolName — GWT-22(a): 잘못된 도구명 시 에러
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

// TestGLMToolsEnable_ValidToolNames — GWT-22: 유효한 도구명 검증
func TestGLMToolsEnable_ValidToolNames(t *testing.T) {
	validNames := []string{"vision", "websearch", "webreader", "all"}
	for _, name := range validNames {
		if err := validateToolName(name); err != nil {
			t.Errorf("유효한 도구명 %q 에 대해 에러가 반환됨: %v", name, err)
		}
	}
}

// ─── Node.js 버전 감지 헬퍼 테스트 ──────────────────────────────────────

// TestDetectNodeVersion_Parse — 버전 문자열 파싱 검증
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

// ─── 백업 파일명 ISO 타임스탬프 형식 테스트 ─────────────────────────────

// TestBackupFilenameFormat — 백업 파일명이 ISO 타임스탬프 형식임을 검증
func TestBackupFilenameFormat(t *testing.T) {
	ts := time.Date(2026, 5, 10, 1, 35, 0, 0, time.UTC)
	name := buildBackupFilename(ts)
	// 기대 패턴: .claude.json.bak-2026-05-10T01-35-00Z
	expected := ".claude.json.bak-2026-05-10T01-35-00Z"
	if name != expected {
		t.Errorf("백업 파일명 = %q, 기대: %q", name, expected)
	}
	// 콜론이 없어야 함 (파일명 안전)
	if strings.Contains(name, ":") {
		t.Errorf("백업 파일명에 콜론이 포함됨: %q", name)
	}
}

// ─── 상수 및 엔트리 빌더 테스트 ─────────────────────────────────────────

// TestZAIMCPEntryBuilder — buildZAIMCPEntry 가 정확한 구조를 반환함
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

// ─── resolveConfigPath 테스트 ────────────────────────────────────────────

// TestResolveConfigPath_UserScope — user scope 시 ~/.claude.json 반환
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

// TestResolveConfigPath_ProjectScope — project scope 시 .mcp.json (cwd 기준)
func TestResolveConfigPath_ProjectScope(t *testing.T) {
	// cwd 를 TempDir 로 변경
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

	// macOS 에서는 /var → /private/var symlink 때문에 EvalSymlinks 비교 필요
	resolvedPath, _ := filepath.EvalSymlinks(filepath.Dir(path))
	resolvedTmp, _ := filepath.EvalSymlinks(tmpDir)
	if resolvedPath != resolvedTmp {
		t.Errorf("path 디렉토리 = %q, 기대: %q", resolvedPath, resolvedTmp)
	}
	if filepath.Base(path) != ".mcp.json" {
		t.Errorf("파일명 = %q, 기대: .mcp.json", filepath.Base(path))
	}
}

// ─── Cobra 커맨드 통합 테스트 ────────────────────────────────────────────

// TestGLMToolsEnableCmd_Success — cobra enable 커맨드 전체 경로 테스트
func TestGLMToolsEnableCmd_Success(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "integration-test-token"
	setupGLMToken(t, homeDir, token)
	setupClaudeJSON(t, homeDir, nil)
	defer makeNodeOK(t)()

	// GLM API 키를 loadGLMKey() 가 읽도록 테스트 env 설정
	// loadGLMKey() 는 MOAI_TEST_GLM_KEY 또는 .env.glm 를 읽음
	// setupGLMToken 이 이미 .env.glm 을 설정했으므로 바로 사용 가능

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	glmToolsEnableCmd.SetOut(outBuf)
	glmToolsEnableCmd.SetErr(errBuf)

	// enable "all" 실행
	err := glmToolsEnableCmd.RunE(glmToolsEnableCmd, []string{"all"})
	if err != nil {
		t.Fatalf("enable 커맨드 실패: %v", err)
	}

	// 성공 메시지 확인
	output := outBuf.String()
	if !strings.Contains(output, "활성화") && !strings.Contains(output, "enable") && !strings.Contains(output, "Z.AI") {
		t.Errorf("성공 메시지에 활성화 관련 문자열 없음: %q", output)
	}
}

// TestGLMToolsDisableCmd_Success — cobra disable 커맨드 전체 경로 테스트
func TestGLMToolsDisableCmd_Success(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "integration-test-token"
	setupGLMToken(t, homeDir, token)
	// zai-mcp-server 가 이미 있는 상태
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

// TestGLMToolsDisableCmd_NothingToDisable — disable 할 엔트리 없을 때 no-op
func TestGLMToolsDisableCmd_NothingToDisable(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	setupGLMToken(t, homeDir, "test-token")
	setupClaudeJSON(t, homeDir, nil) // 빈 mcpServers

	outBuf := new(strings.Builder)
	glmToolsDisableCmd.SetOut(outBuf)

	err := glmToolsDisableCmd.RunE(glmToolsDisableCmd, []string{"all"})
	if err != nil {
		t.Fatalf("disable no-op 실패: %v", err)
	}
	output := outBuf.String()
	if !strings.Contains(output, "없") && !strings.Contains(output, "no") && !strings.Contains(output, "없습니다") {
		// "없음" 이 포함된 메시지 또는 no-op 메시지 확인
		t.Logf("no-op 출력: %q (OK - no error)", output)
	}
}

// TestGLMToolsEnableCmd_NoToken — 토큰 없을 때 에러 반환
// runGLMToolsEnable 은 loadGLMKey() 를 호출하는데, 이 함수는 MOAI_TEST_GLM_KEY env 를 먼저 확인한다.
// 테스트 격리: MOAI_TEST_GLM_KEY 를 빈 문자열로 설정 + HOME 을 .env.glm 없는 tmpDir 로 오버라이드.
func TestGLMToolsEnableCmd_NoToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())     // loadGLMKey() 의 getGLMEnvPath() 가 빈 DIR 를 보도록
	t.Setenv("MOAI_TEST_GLM_KEY", "") // loadGLMKey() 의 테스트 키 env 비워둠
	defer makeNodeOK(t)()

	err := glmToolsEnableCmd.RunE(glmToolsEnableCmd, []string{"all"})
	if err == nil {
		t.Fatal("토큰 없을 때 에러가 반환되어야 함")
	}
}

// TestGLMToolsEnableCmd_BadNodeVersion — 구버전 node 시 에러 반환
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

// TestGLMToolsEnableCmd_InvalidTool — 잘못된 도구명 에러
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

// TestDetectNodeVersion_Real — 실제 node 가 PATH 에 있을 때 (환경 의존적)
func TestDetectNodeVersion_Real(t *testing.T) {
	// node 가 실제로 있는지 먼저 확인
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

// ─── REQ-GMC-002: 오케스트레이션 호환성 테스트 (수동 시뮬레이션) ─────────

// ─── extractTokenFromEntry 분기 커버리지 ─────────────────────────────────

// TestExtractTokenFromEntry_MapStringAny — map[string]any 타입 env
func TestExtractTokenFromEntry_MapStringAny(t *testing.T) {
	entry := map[string]any{
		"env": map[string]any{"Z_AI_API_KEY": "my-token"},
	}
	token := extractTokenFromEntry(entry)
	if token != "my-token" {
		t.Errorf("token = %q, 기대: my-token", token)
	}
}

// TestExtractTokenFromEntry_MapStringString — map[string]string 타입 env
func TestExtractTokenFromEntry_MapStringString(t *testing.T) {
	entry := map[string]any{
		"env": map[string]string{"Z_AI_API_KEY": "string-token"},
	}
	token := extractTokenFromEntry(entry)
	if token != "string-token" {
		t.Errorf("token = %q, 기대: string-token", token)
	}
}

// TestExtractTokenFromEntry_NoEnv — env 없는 경우
func TestExtractTokenFromEntry_NoEnv(t *testing.T) {
	entry := map[string]any{"command": "npx"}
	token := extractTokenFromEntry(entry)
	if token != "" {
		t.Errorf("token = %q, 기대: empty", token)
	}
}

// TestMaskPartial_Short — 짧은 토큰 마스킹
func TestMaskPartial_Short(t *testing.T) {
	if maskPartial("abc") != "****" {
		t.Errorf("짧은 토큰 마스킹 실패: %q", maskPartial("abc"))
	}
}

// TestMaskPartial_Long — 긴 토큰 마스킹
func TestMaskPartial_Long(t *testing.T) {
	result := maskPartial("sk-12345678")
	if result != "sk-1****" {
		t.Errorf("마스킹 결과 = %q, 기대: sk-1****", result)
	}
}

// ─── writeClaudeJSONAtomic 에러 경로 ─────────────────────────────────────

// TestWriteClaudeJSONAtomic_BadDir — 쓸 수 없는 디렉토리 에러
func TestWriteClaudeJSONAtomic_BadDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 권한에서는 권한 테스트 불가")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Windows에서는 /nonexistent-dir 경로가 드라이브 루트로 해석됨")
	}
	// 존재하지 않는 경로에 쓰기 시도 (임시 파일 생성 시도 실패 예상)
	badPath := filepath.Join("/nonexistent-dir-xyz", ".claude.json")
	err := writeClaudeJSONAtomic(badPath, map[string]any{})
	if err == nil {
		t.Error("잘못된 경로에 쓰기가 성공함 (에러 기대)")
	}
}

// ─── readClaudeJSON 에러 경로 ─────────────────────────────────────────────

// TestReadClaudeJSON_InvalidJSON — 잘못된 JSON 파일
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

// TestReadClaudeJSON_NotExists — 파일 없을 때 빈 구조 반환
func TestReadClaudeJSON_NotExists(t *testing.T) {
	root, err := readClaudeJSON("/tmp/nonexistent-claude.json")
	if err != nil {
		t.Fatalf("파일 없을 때 에러가 반환됨: %v", err)
	}
	if root == nil {
		t.Error("빈 구조가 반환되어야 함")
	}
}

// ─── REQ-GMC-002: 오케스트레이션 호환성 테스트 (수동 시뮬레이션) ─────────
func TestGLMTools_OrthogonalToGLMMode(t *testing.T) {
	homeDir := setupToolsTestHome(t)
	token := "test-token"
	setupGLMToken(t, homeDir, token)

	// 프로젝트 루트에 settings.local.json 생성
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

	// settings.local.json 이 변경되지 않아야 함
	data, _ := os.ReadFile(settingsPath)
	if string(data) != initialSettings {
		t.Errorf("enable 이 settings.local.json 을 변경함 (REQ-GMC-002 위반)")
	}
}

// ─── autoEnableMCPServer 테스트 ──────────────────────────────────────────────

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
