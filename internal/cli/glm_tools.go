package cli

// glm_tools.go — SPEC-GLM-MCP-001: `moai glm tools enable|disable` 서브커맨드
//
// @MX:NOTE: [AUTO] Z.AI 공식 @z_ai/mcp-server 를 ~/.claude.json mcpServers 에 등록/해제하는 CLI
// @MX:NOTE: [AUTO] 토큰은 기존 loadGLMKey() 헬퍼 재사용 (GLM_API_KEY from ~/.moai/.env.glm)
// @MX:NOTE: [AUTO] glm.go 의 SPEC-GLM-001 env 정책 (DISABLE_BETAS/DISABLE_PROMPT_CACHING) 과 완전 독립
//
// @MX:WARN: [AUTO] ~/.claude.json 에 atomic write (temp file + rename) 사용
// @MX:REASON: 비원자적 쓰기는 Claude Code 세션 중 파일 손상 가능, POSIX rename atomicity 로 방어
//
// @MX:ANCHOR: [AUTO] runEnableMCPServer, disableMCPServerSafe — 외부 테스트와 서브커맨드에서 직접 호출
// @MX:REASON: GWT 시나리오 22개 모두 이 두 함수를 통해 테스트됨, 시그니처 변경 시 전체 테스트 영향

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ─── 상수 ──────────────────────────────────────────────────────────────────

const (
	// zaiMCPServerKey 는 mcpServers 에서 사용하는 Z.AI MCP 서버의 키 이름
	zaiMCPServerKey = "zai-mcp-server"

	// zaiNPXPackage 는 npx 로 실행할 패키지 이름
	zaiNPXPackage = "@z_ai/mcp-server@latest"

	// nodeMinMajorVersion 은 npx 실행에 필요한 최소 Node.js major 버전
	nodeMinMajorVersion = 22
)

// 지원 도구명 목록
var validToolNames = map[string]bool{
	"vision":    true,
	"websearch":  true,
	"webreader":  true,
	"all":        true,
}

// errNodeNotFound 는 PATH 에서 node 를 찾지 못했을 때 반환하는 센티넬 에러
var errNodeNotFound = errors.New("no Node.js executable found on PATH")

// ─── 테스트 주입 지점 (함수 변수) ─────────────────────────────────────────

// userHomeDirFn 은 홈 디렉토리 조회 함수 변수 (테스트 오버라이드용)
var userHomeDirFn = userHomeDir

// detectNodeFn 은 node 버전 감지 함수 변수 (테스트 오버라이드용)
// 반환값: (major 버전 정수, 버전 문자열 e.g. "v22.5.0", 에러)
var detectNodeFn = detectNodeVersion

// ─── Cobra 커맨드 정의 ─────────────────────────────────────────────────────

// glmToolsCmd — `moai glm tools` 루트 커맨드
var glmToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Z.AI MCP 서버 도구 관리 (enable/disable)",
	Long: `Z.AI MCP 서버를 Claude Code 에 등록하거나 해제합니다.

사전 조건:
  - Node.js >= v22.0.0 (npx 실행용)
  - GLM API 키 설정: moai glm setup <api-key>

서브커맨드:
  enable  [vision|websearch|webreader|all]   Z.AI MCP 서버 등록
  disable [vision|websearch|webreader|all]   Z.AI MCP 서버 해제

예시:
  moai glm tools enable all
  moai glm tools disable all
  moai glm tools enable vision --scope project`,
}

// glmToolsEnableCmd — `moai glm tools enable` 커맨드
var glmToolsEnableCmd = &cobra.Command{
	Use:   "enable [vision|websearch|webreader|all]",
	Short: "Z.AI MCP 서버를 ~/.claude.json 에 등록",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGLMToolsEnable,
}

// glmToolsDisableCmd — `moai glm tools disable` 커맨드
var glmToolsDisableCmd = &cobra.Command{
	Use:   "disable [vision|websearch|webreader|all]",
	Short: "Z.AI MCP 서버를 ~/.claude.json 에서 해제",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGLMToolsDisable,
}

func init() {
	// --scope 플래그: project 시 .mcp.json 에 기록 (REQ-GMC-008)
	glmToolsEnableCmd.Flags().String("scope", "user", "등록 범위: user (기본, ~/.claude.json) 또는 project (.mcp.json)")
	glmToolsDisableCmd.Flags().String("scope", "user", "해제 범위: user (기본) 또는 project (.mcp.json)")

	// --force 플래그: 토큰 불일치 시 강제 덮어쓰기 (REQ-GMC-006)
	glmToolsEnableCmd.Flags().Bool("force", false, "기존 토큰이 달라도 강제 덮어쓰기")

	glmToolsCmd.AddCommand(glmToolsEnableCmd, glmToolsDisableCmd)

	// glmCmd 의 DisableFlagParsing=true 이므로 수동 라우팅에 "tools" 케이스 추가 필요.
	// glm.go 의 runGLM 에서 args[0]=="tools" 를 처리하도록 init() 에서 subcommand 등록만 함.
	glmCmd.AddCommand(glmToolsCmd)
}

// ─── 커맨드 핸들러 ─────────────────────────────────────────────────────────

// runGLMToolsEnable — `moai glm tools enable [도구명]` 실행
func runGLMToolsEnable(cmd *cobra.Command, args []string) error {
	toolName := "all"
	if len(args) > 0 {
		toolName = args[0]
	}

	// (a) 도구명 검증 (REQ-GMC-001, GWT-22)
	if err := validateToolName(toolName); err != nil {
		return err
	}

	// (b) GLM 토큰 로드 (REQ-GMC-007, GWT-12)
	token := loadGLMKey()
	if token == "" {
		return fmt.Errorf(
			"GLM API 키가 설정되지 않았습니다\n\n" +
				"토큰 등록 방법:\n" +
				"  moai glm setup <api-key>\n\n" +
				"Z.AI API 키는 https://bigmodel.cn 에서 발급받을 수 있습니다",
		)
	}

	// (c) Node.js 버전 검증 (REQ-GMC-009, GWT-14, GWT-15)
	major, versionStr, err := detectNodeFn()
	if err != nil {
		if errors.Is(err, errNodeNotFound) {
			return fmt.Errorf(
				"no Node.js executable found on PATH\n\n"+
					"최소 요구 버전: >= v%d.0.0\n\n"+
					"설치 방법:\n"+
					"  https://nodejs.org/ 에서 다운로드 또는\n"+
					"  nvm install %d",
				nodeMinMajorVersion, nodeMinMajorVersion,
			)
		}
		return fmt.Errorf("node 버전 확인 실패: %w", err)
	}
	if major < nodeMinMajorVersion {
		return fmt.Errorf(
			"감지된 Node.js 버전이 너무 낮습니다: %s, 최소 요구 >= v%d.0.0\n\n"+
				"업그레이드 방법:\n"+
				"  https://nodejs.org/ 에서 최신 버전 다운로드 또는\n"+
				"  nvm install %d",
			versionStr, nodeMinMajorVersion, nodeMinMajorVersion,
		)
	}

	// (d) scope 결정 (REQ-GMC-008, GWT-13)
	scope, _ := cmd.Flags().GetString("scope")
	force, _ := cmd.Flags().GetBool("force")

	configPath, err := resolveConfigPath(scope)
	if err != nil {
		return fmt.Errorf("설정 파일 경로 결정 실패: %w", err)
	}

	// (e) 기존 엔트리 + force 처리 후 enable (REQ-GMC-003, REQ-GMC-006)
	if force {
		if err := runEnableMCPServer(configPath, token); err != nil {
			return err
		}
	} else {
		skipped, err := enableMCPServerIdempotent(configPath, token)
		if err != nil {
			return err
		}
		if skipped {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Z.AI MCP 서버가 이미 활성화되어 있습니다 (토큰 일치 — 변경 없음)")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "비활성화: moai glm tools disable all")
			return nil
		}
	}

	// (f) 성공 메시지 출력 (REQ-GMC-003, GWT-5)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Z.AI MCP 서버 활성화 완료")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "활성화된 도구:")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  - Vision (이미지 OCR, 스크린샷 분석)")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  - Web Search (실시간 웹 검색)")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  - Web Reader (웹 페이지 내용 읽기)")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "주의: Pro 플랜 ($9/월) 이상에서 모든 도구가 활성화됩니다")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Claude Code 를 재시작해야 MCP 서버가 로드됩니다")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "비활성화: moai glm tools disable all")

	return nil
}

// runGLMToolsDisable — `moai glm tools disable [도구명]` 실행
func runGLMToolsDisable(cmd *cobra.Command, args []string) error {
	toolName := "all"
	if len(args) > 0 {
		toolName = args[0]
	}

	// 도구명 검증 (REQ-GMC-001, GWT-22)
	if err := validateToolName(toolName); err != nil {
		return err
	}

	// scope 결정 (REQ-GMC-008)
	scope, _ := cmd.Flags().GetString("scope")
	configPath, err := resolveConfigPath(scope)
	if err != nil {
		return fmt.Errorf("설정 파일 경로 결정 실패: %w", err)
	}

	// 엔트리 제거 (REQ-GMC-004, GWT-6, GWT-7)
	removed, err := disableMCPServerSafe(configPath)
	if err != nil {
		return err
	}

	if !removed {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "활성화된 Z.AI MCP 서버 엔트리가 없습니다 (변경 없음)")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Z.AI MCP 서버 비활성화 완료")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  제거된 도구: Vision, Web Search, Web Reader")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Claude Code 를 재시작해야 변경사항이 반영됩니다")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "재활성화: moai glm tools enable all")

	return nil
}

// ─── 핵심 로직 함수 ────────────────────────────────────────────────────────

// resolveConfigPath 는 scope 에 따라 설정 파일 경로를 반환한다.
// scope "project" → .mcp.json (cwd 기준), 그 외 → ~/.claude.json (REQ-GMC-008)
func resolveConfigPath(scope string) (string, error) {
	if scope == "project" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("현재 디렉토리 조회 실패: %w", err)
		}
		return filepath.Join(cwd, ".mcp.json"), nil
	}
	// 기본: user scope → ~/.claude.json
	home, err := userHomeDirFn()
	if err != nil {
		return "", fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
	}
	return filepath.Join(home, ".claude.json"), nil
}

// validateToolName 은 도구명이 유효한지 검증한다 (REQ-GMC-001, GWT-22)
func validateToolName(name string) error {
	if validToolNames[name] {
		return nil
	}
	return fmt.Errorf(
		"알 수 없는 도구명: %q\n지원 도구명: vision, websearch, webreader, all",
		name,
	)
}

// buildZAIMCPEntry 는 Z.AI MCP 서버 엔트리를 구성한다 (REQ-GMC-003)
func buildZAIMCPEntry(token string) map[string]any {
	return map[string]any{
		"command": "npx",
		"args":    []string{"-y", zaiNPXPackage},
		"env": map[string]string{
			"Z_AI_API_KEY": token,
			"Z_AI_MODE":    "ZAI",
		},
	}
}

// buildBackupFilename 은 백업 파일명을 생성한다 (REQ-GMC-005)
// 형식: .claude.json.bak-<ISO ts> (콜론을 하이픈으로 대체하여 파일명 안전)
func buildBackupFilename(t time.Time) string {
	ts := t.UTC().Format("2006-01-02T15-04-05Z")
	return ".claude.json.bak-" + ts
}

// readClaudeJSON 은 configPath 에서 JSON 을 읽고 파싱한다.
// 파일이 없으면 빈 구조({})를 반환한다.
func readClaudeJSON(configPath string) (map[string]any, error) {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}
	return root, nil
}

// writeClaudeJSONAtomic 은 configPath 에 JSON 을 atomic write 한다 (REQ-GMC-005, R7)
// atomic write: 임시 파일 → os.Rename (POSIX 원자성 보장)
func writeClaudeJSONAtomic(configPath string, root map[string]any) error {
	jsonBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 직렬화 실패: %w", err)
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".claude-json-*.tmp")
	if err != nil {
		return fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // 실패 시 임시 파일 정리

	if _, err := tmp.Write(jsonBytes); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("임시 파일 쓰기 실패: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("임시 파일 닫기 실패: %w", err)
	}

	// 원자적 rename
	if err := os.Rename(tmpName, configPath); err != nil {
		return fmt.Errorf("파일 교체 실패: %w", err)
	}
	return nil
}

// backupClaudeJSON 은 configPath 의 현재 내용을 백업한다 (REQ-GMC-005)
// 파일이 없으면 백업하지 않는다.
func backupClaudeJSON(configPath string) error {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil // 파일 없으면 백업 불필요
	}
	if err != nil {
		return fmt.Errorf("백업 대상 파일 읽기 실패: %w", err)
	}

	backupName := buildBackupFilename(time.Now())
	backupPath := filepath.Join(filepath.Dir(configPath), backupName)

	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return fmt.Errorf("백업 파일 생성 실패: %w", err)
	}
	return nil
}

// runEnableMCPServer 는 configPath 에 zai-mcp-server 엔트리를 추가한다.
// 토큰 불일치 시 에러 반환 (REQ-GMC-006 (b)).
// 파일 변경 전 백업 생성 (REQ-GMC-005).
// 다른 mcpServers 엔트리는 절대 변경하지 않음 (REQ-GMC-010).
//
// @MX:ANCHOR: [AUTO] 22개 GWT 시나리오의 핵심 진입점 — 시그니처 변경 시 모든 테스트 영향
// @MX:REASON: fan_in = 6 (glmToolsEnableCmd, enableMCPServerIdempotent, runEnableMCPServerScoped, test helpers)
func runEnableMCPServer(configPath string, token string) error {
	if token == "" {
		return fmt.Errorf(
			"GLM API 키가 설정되지 않았습니다\n" +
				"  moai glm setup <api-key>",
		)
	}

	root, err := readClaudeJSON(configPath)
	if err != nil {
		return err
	}

	// mcpServers 맵 확보
	mcpServers := getMCPServers(root)

	// 기존 엔트리 토큰 검사 (REQ-GMC-006)
	if existing, ok := mcpServers[zaiMCPServerKey].(map[string]any); ok {
		existingToken := extractTokenFromEntry(existing)
		if existingToken != token {
			// 토큰 불일치 → 거부 + --force 안내
			return fmt.Errorf(
				"기존 zai-mcp-server 엔트리에 다른 토큰이 설정되어 있습니다\n"+
					"  현재 토큰: %s...%s\n"+
					"  새 토큰:   %s...%s\n\n"+
					"강제 덮어쓰기: moai glm tools enable --force",
				maskPartial(existingToken), maskPartial(existingToken)[len(maskPartial(existingToken))-4:],
				maskPartial(token), maskPartial(token)[len(maskPartial(token))-4:],
			)
		}
		// 토큰 일치 → 이미 처리됨 (idempotent: enableMCPServerIdempotent 에서 처리)
	}

	// 백업 생성 (REQ-GMC-005)
	if err := backupClaudeJSON(configPath); err != nil {
		return fmt.Errorf("백업 생성 실패: %w", err)
	}

	// 엔트리 추가 (REQ-GMC-003, REQ-GMC-010)
	mcpServers[zaiMCPServerKey] = buildZAIMCPEntry(token)
	root["mcpServers"] = mcpServers

	return writeClaudeJSONAtomic(configPath, root)
}

// enableMCPServerIdempotent 는 idempotent enable 을 수행한다.
// 반환값: (skipped bool, err error)
//   - skipped=true: 동일 토큰으로 이미 등록됨 → 변경 없음, 백업 없음 (REQ-GMC-006 (a))
//   - skipped=false: 신규 등록 수행
func enableMCPServerIdempotent(configPath string, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf(
			"GLM API 키가 설정되지 않았습니다\n" +
				"  moai glm setup <api-key>",
		)
	}

	root, err := readClaudeJSON(configPath)
	if err != nil {
		return false, err
	}

	mcpServers := getMCPServers(root)

	// 기존 엔트리 확인
	if existing, ok := mcpServers[zaiMCPServerKey].(map[string]any); ok {
		existingToken := extractTokenFromEntry(existing)
		if existingToken == token {
			// 토큰 일치 → idempotent skip (REQ-GMC-006 (a), GWT-9: 백업 없음)
			return true, nil
		}
		// 토큰 불일치 → 에러 반환 (REQ-GMC-006 (b))
		return false, fmt.Errorf(
			"기존 zai-mcp-server 엔트리에 다른 토큰이 설정되어 있습니다\n"+
				"강제 덮어쓰기: moai glm tools enable --force",
		)
	}

	// 신규 등록 → 백업 후 쓰기
	if err := backupClaudeJSON(configPath); err != nil {
		return false, fmt.Errorf("백업 생성 실패: %w", err)
	}

	mcpServers[zaiMCPServerKey] = buildZAIMCPEntry(token)
	root["mcpServers"] = mcpServers

	return false, writeClaudeJSONAtomic(configPath, root)
}

// runEnableMCPServerScoped 는 project scope (.mcp.json) 에 enable 을 수행한다 (REQ-GMC-008)
func runEnableMCPServerScoped(mcpJSONPath string, token string) error {
	return runEnableMCPServer(mcpJSONPath, token)
}

// disableMCPServerSafe 는 configPath 에서 zai-mcp-server 엔트리만 제거한다 (REQ-GMC-004, REQ-GMC-010).
// 다른 mcpServers 엔트리는 변경하지 않음.
// 반환값: (removed bool, err error) — 엔트리가 없으면 removed=false (idempotent)
//
// @MX:ANCHOR: [AUTO] REQ-GMC-004/010 의 핵심 구현 — 부분 삭제 안전성
// @MX:REASON: fan_in = 4 (runGLMToolsDisable, 다수의 GWT 테스트)
func disableMCPServerSafe(configPath string) (bool, error) {
	root, err := readClaudeJSON(configPath)
	if err != nil {
		return false, err
	}

	mcpServers := getMCPServers(root)

	if _, ok := mcpServers[zaiMCPServerKey]; !ok {
		// 엔트리 없음 → idempotent skip
		return false, nil
	}

	// 백업 후 제거 (REQ-GMC-005)
	if err := backupClaudeJSON(configPath); err != nil {
		return false, fmt.Errorf("백업 생성 실패: %w", err)
	}

	delete(mcpServers, zaiMCPServerKey)
	root["mcpServers"] = mcpServers

	return true, writeClaudeJSONAtomic(configPath, root)
}

// ─── 내부 헬퍼 함수 ────────────────────────────────────────────────────────

// getMCPServers 는 root JSON 에서 mcpServers 맵을 추출한다.
// 없으면 빈 맵을 반환하고 root 에 설정한다.
func getMCPServers(root map[string]any) map[string]any {
	if existing, ok := root["mcpServers"].(map[string]any); ok {
		return existing
	}
	m := map[string]any{}
	root["mcpServers"] = m
	return m
}

// extractTokenFromEntry 는 MCP 엔트리의 env.Z_AI_API_KEY 값을 추출한다.
func extractTokenFromEntry(entry map[string]any) string {
	envAny, ok := entry["env"]
	if !ok {
		return ""
	}
	switch env := envAny.(type) {
	case map[string]any:
		if v, ok := env["Z_AI_API_KEY"].(string); ok {
			return v
		}
	case map[string]string:
		return env["Z_AI_API_KEY"]
	}
	return ""
}

// maskPartial 은 토큰의 일부를 마스킹한다 (로그 표시용)
func maskPartial(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****"
}

// ─── Node.js 버전 감지 ────────────────────────────────────────────────────

// checkNodeVersion 은 Node.js 가 설치되어 있고 최소 버전 이상인지 검증한다 (REQ-GMC-009).
// GWT-14 (부재), GWT-15 (구버전) 시나리오의 공통 검증 함수.
func checkNodeVersion() error {
	major, versionStr, err := detectNodeFn()
	if err != nil {
		if errors.Is(err, errNodeNotFound) {
			return fmt.Errorf(
				"no Node.js executable found on PATH\n"+
					"최소 요구 버전: >= v%d.0.0\n"+
					"설치: https://nodejs.org/ 또는 nvm install %d",
				nodeMinMajorVersion, nodeMinMajorVersion,
			)
		}
		return fmt.Errorf("node 버전 확인 실패: %w", err)
	}
	if major < nodeMinMajorVersion {
		return fmt.Errorf(
			"감지된 Node.js 버전이 너무 낮습니다: %s, 최소 요구 >= v%d.0.0\n"+
				"업그레이드: https://nodejs.org/ 또는 nvm install %d",
			versionStr, nodeMinMajorVersion, nodeMinMajorVersion,
		)
	}
	return nil
}

// detectNodeVersion 은 PATH 에서 node 의 major 버전을 감지한다 (REQ-GMC-009).
// 반환값: (major int, versionString string, error)
func detectNodeVersion() (int, string, error) {
	path, err := exec.LookPath("node")
	if err != nil {
		return 0, "", errNodeNotFound
	}
	_ = path

	out, err := exec.Command("node", "--version").Output() //nolint:gosec
	if err != nil {
		return 0, "", fmt.Errorf("node --version 실행 실패: %w", err)
	}

	versionStr := strings.TrimSpace(string(out))
	major, err := parseNodeMajorVersion(versionStr)
	if err != nil {
		return 0, versionStr, fmt.Errorf("node 버전 파싱 실패 (%q): %w", versionStr, err)
	}

	return major, versionStr, nil
}

// parseNodeMajorVersion 은 "v22.5.0" 형식의 버전 문자열에서 major 정수를 추출한다.
func parseNodeMajorVersion(version string) (int, error) {
	v := strings.TrimPrefix(version, "v")
	parts := strings.SplitN(v, ".", 2)
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("버전 문자열 파싱 실패: %q", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("major 버전 정수 변환 실패: %w", err)
	}
	return major, nil
}
