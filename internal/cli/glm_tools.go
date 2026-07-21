package cli

// glm_tools.go — SPEC-GLM-MCP-001: `moai glm tools enable|disable` subcommands
//
// @MX:NOTE: [AUTO] CLI that registers/deregisters the official Z.AI @z_ai/mcp-server in ~/.claude.json mcpServers
// @MX:NOTE: [AUTO] Reuses the existing loadGLMKey() helper for the token (GLM_API_KEY from ~/.moai/.env.glm)
// @MX:NOTE: [AUTO] Fully independent from glm.go's SPEC-GLM-001 env policy (DISABLE_BETAS)
//
// @MX:WARN: [AUTO] Uses atomic write (temp file + rename) when modifying ~/.claude.json
// @MX:REASON: Non-atomic writes can corrupt the file mid-Claude Code session; defended via POSIX rename atomicity
//
// @MX:ANCHOR: [AUTO] runEnableMCPServer, disableMCPServerSafe — called directly by external tests and subcommands
// @MX:REASON: All 22 GWT scenarios exercise these two functions; signature changes affect the entire test surface

import (
	"bytes"
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

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// ─── Constants ─────────────────────────────────────────────────────────────

const (
	// zaiMCPServerKey는 mcpServers 에서 사용하는 Z.AI vision MCP 서버 키 이름이다 (stdio npx).
	zaiMCPServerKey = "zai-mcp-server"

	// zaiWebSearchPrimeKey는 web search HTTP MCP 서버 키 이름이다.
	// 언더스코어 표기로 등록하여 canonical 툴 참조 mcp__web_search_prime__webSearchPrime 를 생성한다 (design.md §B.3).
	zaiWebSearchPrimeKey = "web_search_prime"

	// zaiWebReaderKey는 web reader HTTP MCP 서버 키 이름이다.
	// 언더스코어 표기로 등록하여 canonical 툴 참조 mcp__web_reader__webReader 를 생성한다 (design.md §B.3).
	zaiWebReaderKey = "web_reader"

	// zaiNPXPackage는 npx 로 실행하는 vision 서버 패키지 이름이다.
	zaiNPXPackage = "@z_ai/mcp-server@latest"

	// zaiWebSearchPrimeURL는 web search HTTP MCP 엔드포인트이다 (CLAUDE.local.md §14 하드코딩 금지 → const 추출).
	zaiWebSearchPrimeURL = "https://api.z.ai/api/mcp/web_search_prime/mcp"

	// zaiWebReaderURL는 web reader HTTP MCP 엔드포인트이다 (CLAUDE.local.md §14 하드코딩 금지 → const 추출).
	zaiWebReaderURL = "https://api.z.ai/api/mcp/web_reader/mcp"

	// zaiBearerHeaderTemplate는 HTTP MCP 서버의 Authorization 헤더 값 템플릿이다.
	// ${Z_AI_API_KEY} 리터럴은 Claude Code 가 .mcp.json/.claude.json 헤더에서 확장한다 (design.md §B.4 R2).
	zaiBearerHeaderTemplate = "Bearer ${Z_AI_API_KEY}"

	// nodeMinMajorVersion는 npx 실행에 필요한 최소 Node.js major 버전이다 (vision 서버 전용).
	nodeMinMajorVersion = 22
)

// 지원 도구명 목록
var validToolNames = map[string]bool{
	"vision":    true,
	"websearch": true,
	"webreader": true,
	"all":       true,
}

// toolServerKeys는 각 도구명이 등록하는 서버 키 집합을 반환한다 (REQ-GWR-C1~C4).
// vision→[zai-mcp-server], websearch→[web_search_prime], webreader→[web_reader], all→3개 전부.
func toolServerKeys(toolName string) []string {
	switch toolName {
	case "vision":
		return []string{zaiMCPServerKey}
	case "websearch":
		return []string{zaiWebSearchPrimeKey}
	case "webreader":
		return []string{zaiWebReaderKey}
	case "all":
		return []string{zaiMCPServerKey, zaiWebSearchPrimeKey, zaiWebReaderKey}
	default:
		return nil
	}
}

// toolSetNeedsNode는 요청한 도구 집합이 npx vision 서버(Node 필요)를 포함하는지 판정한다 (REQ-GWR-C8).
// vision 또는 all 일 때만 true — websearch/webreader 는 HTTP 서버라 Node 게이트가 무의미하다.
func toolSetNeedsNode(toolName string) bool {
	for _, key := range toolServerKeys(toolName) {
		if key == zaiMCPServerKey {
			return true
		}
	}
	return false
}

// toolServerLabels는 각 서버 키에 대응하는 사용자 표시용 상세 라벨(한국어)이다.
// enable 성공 메시지의 "활성화된 도구" 블록을 실제 등록된 서버 집합에서 데이터-구동으로 생성할 때 사용한다 (REQ-GWR-C5, design.md §B.6).
var toolServerLabels = map[string]string{
	zaiMCPServerKey:      "Vision (이미지 OCR, 스크린샷 분석)",
	zaiWebSearchPrimeKey: "Web Search (실시간 웹 검색)",
	zaiWebReaderKey:      "Web Reader (웹 페이지 내용 읽기)",
}

// toolServerShortLabels는 각 서버 키에 대응하는 짧은 라벨이다.
// disable "제거된 도구" 한 줄 메시지를 콤마로 결합할 때 사용한다 (REQ-GWR-C5).
var toolServerShortLabels = map[string]string{
	zaiMCPServerKey:      "Vision",
	zaiWebSearchPrimeKey: "Web Search",
	zaiWebReaderKey:      "Web Reader",
}

// errNodeNotFound is the sentinel error returned when node is not found on PATH
var errNodeNotFound = errors.New("no Node.js executable found on PATH")

// ─── Test injection points (function variables) ───────────────────────────

// userHomeDirFn is the home-directory lookup function variable (for test override)
var userHomeDirFn = userHomeDir

// detectNodeFn is the node-version detection function variable (for test override)
// Return values: (major version int, version string e.g. "v22.5.0", error)
var detectNodeFn = detectNodeVersion

// claudeJSONGuardPreLockHook is a test injection point invoked between the
// prep-phase read and the in-lock compare-and-publish inside mutateClaudeJSON-
// Atomic. Production code leaves it nil (no-op). Tests set it to simulate a
// concurrent non-cooperating writer (e.g., a live Claude Code process that does
// not respect the .lock file) so the compare-and-retry path is exercised
// deterministically rather than relying on goroutine scheduling.
var claudeJSONGuardPreLockHook func(configPath string)

// claudeJSONGuardMaxRetries bounds the number of times mutateClaudeJSONAtomic
// re-applies the mutation when a concurrent writer keeps changing the file
// between the prep read and the in-lock compare. After this ceiling the guard
// writes best-effort with a stderr warning (fail-open — never block a session
// start indefinitely). SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-003.
const claudeJSONGuardMaxRetries = 3

// claudeJSONGuardRetryDelay is the backoff between compare-and-retry attempts.
const claudeJSONGuardRetryDelay = 15 * time.Millisecond

// ─── Cobra command definitions ─────────────────────────────────────────────

// glmToolsCmd — `moai glm tools` root command
var glmToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Manage Z.AI MCP server tools (enable/disable)",
	Long: `Register or unregister the Z.AI MCP server with Claude Code.

Prerequisites:
  - Node.js >= v22.0.0 (required to run npx)
  - GLM API key configured: moai glm setup <api-key>

Subcommands:
  enable  [vision|websearch|webreader|all]   Register the Z.AI MCP server
  disable [vision|websearch|webreader|all]   Unregister the Z.AI MCP server

Examples:
  moai glm tools enable all
  moai glm tools disable all
  moai glm tools enable vision --scope project`,
}

// glmToolsEnableCmd — `moai glm tools enable` command
var glmToolsEnableCmd = &cobra.Command{
	Use:   "enable [vision|websearch|webreader|all]",
	Short: "Register the Z.AI MCP server in ~/.claude.json",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGLMToolsEnable,
}

// glmToolsDisableCmd — `moai glm tools disable` command
var glmToolsDisableCmd = &cobra.Command{
	Use:   "disable [vision|websearch|webreader|all]",
	Short: "Unregister the Z.AI MCP server from ~/.claude.json",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGLMToolsDisable,
}

func init() {
	// --scope flag: write to .mcp.json when scope is project (REQ-GMC-008)
	glmToolsEnableCmd.Flags().String("scope", "user", "등록 범위: user (기본, ~/.claude.json) 또는 project (.mcp.json)")
	glmToolsDisableCmd.Flags().String("scope", "user", "해제 범위: user (기본) 또는 project (.mcp.json)")

	// --force flag: force overwrite on token mismatch (REQ-GMC-006)
	glmToolsEnableCmd.Flags().Bool("force", false, "기존 토큰이 달라도 강제 덮어쓰기")

	glmToolsCmd.AddCommand(glmToolsEnableCmd, glmToolsDisableCmd)

	// Because glmCmd has DisableFlagParsing=true, the manual routing requires an explicit "tools" case.
	// init() here only registers the subcommand; runGLM in glm.go dispatches args[0]=="tools".
	glmCmd.AddCommand(glmToolsCmd)
}

// ─── Command handlers ──────────────────────────────────────────────────────

// runGLMToolsEnable — runs `moai glm tools enable [tool-name]`
func runGLMToolsEnable(cmd *cobra.Command, args []string) error {
	toolName := "all"
	if len(args) > 0 {
		toolName = args[0]
	}

	// (a) Validate tool name (REQ-GMC-001, GWT-22)
	if err := validateToolName(toolName); err != nil {
		return err
	}

	// (b) Load GLM token (REQ-GMC-007, GWT-12)
	token := loadGLMKey()
	if token == "" {
		return fmt.Errorf(
			"GLM API 키가 설정되지 않았습니다\n\n" +
				"토큰 등록 방법:\n" +
				"  moai glm setup <api-key>\n\n" +
				"Z.AI API 키는 https://bigmodel.cn 에서 발급받을 수 있습니다",
		)
	}

	// (c) Validate Node.js version — vision(npx) 서버를 포함할 때만 (REQ-GWR-C8, REQ-GMC-009)
	// websearch/webreader 전용 enable 은 HTTP 서버라 Node 게이트를 적용하지 않는다.
	if toolSetNeedsNode(toolName) {
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
	}

	// (d) Determine scope (REQ-GMC-008, GWT-13)
	scope, _ := cmd.Flags().GetString("scope")
	force, _ := cmd.Flags().GetBool("force")

	configPath, err := resolveConfigPath(scope)
	if err != nil {
		return fmt.Errorf("설정 파일 경로 결정 실패: %w", err)
	}

	// (e) Enable after handling existing entry + force flag (REQ-GWR-C, REQ-GMC-006)
	if force {
		if err := runEnableMCPServerForTool(configPath, toolName, token); err != nil {
			return err
		}
	} else {
		skipped, err := enableMCPServerIdempotentForTool(configPath, toolName, token)
		if err != nil {
			return err
		}
		if skipped {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Z.AI MCP 서버가 이미 활성화되어 있습니다 (변경 없음)")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "비활성화: moai glm tools disable "+toolName)
			return nil
		}
	}

	// (f) Emit success message — 실제 등록된 서버 집합 기반 (REQ-GWR-C5, GWT-5)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Z.AI MCP 서버 활성화 완료")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "활성화된 도구:")
	for _, line := range registeredToolLines(toolName) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  - "+line)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "주의: Pro 플랜 ($9/월) 이상에서 모든 도구가 활성화됩니다")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Claude Code 를 재시작해야 MCP 서버가 로드됩니다")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "비활성화: moai glm tools disable "+toolName)

	return nil
}

// registeredToolLines는 도구명에 대응하는 서버 상세 라벨 목록을 결정적 순서로 반환한다 (REQ-GWR-C5).
// enable 성공 메시지에서 실제 등록된 서버만 정확히 나열하기 위해 사용한다.
func registeredToolLines(toolName string) []string {
	var lines []string
	for _, key := range toolServerKeys(toolName) {
		if label, ok := toolServerLabels[key]; ok {
			lines = append(lines, label)
		}
	}
	return lines
}

// registeredToolShortLabels는 도구명에 대응하는 짧은 라벨 목록을 결정적 순서로 반환한다 (REQ-GWR-C5).
// disable 메시지의 한 줄 콤마 결합에 사용한다.
func registeredToolShortLabels(toolName string) []string {
	var lines []string
	for _, key := range toolServerKeys(toolName) {
		if label, ok := toolServerShortLabels[key]; ok {
			lines = append(lines, label)
		}
	}
	return lines
}

// runGLMToolsDisable — runs `moai glm tools disable [tool-name]`
func runGLMToolsDisable(cmd *cobra.Command, args []string) error {
	toolName := "all"
	if len(args) > 0 {
		toolName = args[0]
	}

	// Validate tool name (REQ-GMC-001, GWT-22)
	if err := validateToolName(toolName); err != nil {
		return err
	}

	// Determine scope (REQ-GMC-008)
	scope, _ := cmd.Flags().GetString("scope")
	configPath, err := resolveConfigPath(scope)
	if err != nil {
		return fmt.Errorf("설정 파일 경로 결정 실패: %w", err)
	}

	// Remove entry — 도구명에 대응하는 서버만 제거 (REQ-GWR-C6, GWT-6, GWT-7)
	removed, err := disableMCPServerForTool(configPath, toolName)
	if err != nil {
		return err
	}

	if !removed {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "활성화된 Z.AI MCP 서버 엔트리가 없습니다 (변경 없음)")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Z.AI MCP 서버 비활성화 완료")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  제거된 도구: "+strings.Join(registeredToolShortLabels(toolName), ", "))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Claude Code 를 재시작해야 변경사항이 반영됩니다")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "재활성화: moai glm tools enable "+toolName)

	return nil
}

// ─── Core logic functions ──────────────────────────────────────────────────

// resolveConfigPath returns the config file path based on scope.
// scope "project" → .mcp.json (relative to cwd); otherwise → ~/.claude.json (REQ-GMC-008)
func resolveConfigPath(scope string) (string, error) {
	if scope == "project" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("현재 디렉토리 조회 실패: %w", err)
		}
		return filepath.Join(cwd, ".mcp.json"), nil
	}
	// Default: user scope → ~/.claude.json
	home, err := userHomeDirFn()
	if err != nil {
		return "", fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
	}
	return filepath.Join(home, ".claude.json"), nil
}

// validateToolName verifies that the tool name is valid (REQ-GMC-001, GWT-22)
func validateToolName(name string) error {
	if validToolNames[name] {
		return nil
	}
	return fmt.Errorf(
		"알 수 없는 도구명: %q\n지원 도구명: vision, websearch, webreader, all",
		name,
	)
}

// buildZAIMCPEntry는 vision(zai-mcp-server) stdio npx 엔트리를 구성한다 (REQ-GMC-003, REQ-GWR-C3).
// env.Z_AI_API_KEY 에는 resolve 된 토큰을 그대로 기록한다 (HTTP 헤더와 달리 npx env 는 Claude Code 가 확장하지 않음).
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

// buildZAIHTTPEntry는 web_search_prime / web_reader 같은 HTTP MCP 엔트리를 구성한다 (REQ-GWR-C1/C2).
// Authorization 헤더에는 리터럴 ${Z_AI_API_KEY} 를 기록한다 — Claude Code 가 런타임에 확장한다 (design.md §B.4 R2).
func buildZAIHTTPEntry(url string) map[string]any {
	return map[string]any{
		"type": "http",
		"url":  url,
		"headers": map[string]string{
			"Authorization": zaiBearerHeaderTemplate,
		},
	}
}

// buildZAIMCPEntries는 도구명을 받아 등록할 {서버키: 엔트리} 집합을 반환한다 (REQ-GWR-C1~C4).
// vision→npx stdio 1개, websearch/webreader→HTTP 1개, all→3개 전부.
func buildZAIMCPEntries(toolName, token string) map[string]map[string]any {
	entries := map[string]map[string]any{}
	for _, key := range toolServerKeys(toolName) {
		switch key {
		case zaiMCPServerKey:
			entries[key] = buildZAIMCPEntry(token)
		case zaiWebSearchPrimeKey:
			entries[key] = buildZAIHTTPEntry(zaiWebSearchPrimeURL)
		case zaiWebReaderKey:
			entries[key] = buildZAIHTTPEntry(zaiWebReaderURL)
		}
	}
	return entries
}

// buildBackupFilename generates a backup file name (REQ-GMC-005)
// Format: .claude.json.bak-<ISO ts> (colons replaced with hyphens for filename safety)
func buildBackupFilename(t time.Time) string {
	ts := t.UTC().Format("2006-01-02T15-04-05Z")
	return ".claude.json.bak-" + ts
}

// readClaudeJSON reads and parses JSON from configPath.
// Returns an empty struct ({}) when the file does not exist.
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

// writeClaudeJSONBytes writes pre-marshaled bytes to configPath atomically via the
// consolidated writeFileAtomic helper (SPEC-CLIFIX-CONCURRENCY-001 M3). It is the
// bytes-level publication step used by the guarded RMW (mutateClaudeJSONAtomic),
// so the guard can write pre-marshaled bytes inside the lock without re-running
// the marshal under the lock. The 0600 perm preserves the credential-bearing
// file-mode contract for ~/.claude.json.
//
// @MX:NOTE: [AUTO] M3 consolidation — writeClaudeJSONAtomic deleted (sole caller inlined the marshal); this bytes-level seam delegates to writeFileAtomic so mutateClaudeJSONAtomic keeps the marshal OUTSIDE the lock (acceptance.md §C large-file constraint).
// @MX:SPEC: SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-004
func writeClaudeJSONBytes(configPath string, data []byte) error {
	return writeFileAtomic(configPath, data, 0o600)
}

// readClaudeJSONRaw reads configPath and returns the raw bytes alongside the
// parsed root map. The guard (mutateClaudeJSONAtomic) uses the raw bytes for the
// in-lock content compare (detecting concurrent writers) while the parsed root
// feeds the prep-phase apply. Returns (nil, empty-map, nil) when the file does
// not exist, matching readClaudeJSON semantics.
func readClaudeJSONRaw(configPath string) ([]byte, map[string]any, error) {
	// The prep read runs OUTSIDE the guard's lock, so it races another writer's
	// atomic replace. On Windows that read transiently fails with a sharing
	// violation, which would surface as a spurious "설정 파일 읽기 실패" rather
	// than the lost update the guard exists to prevent.
	data, err := atomicfile.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, map[string]any{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}
	if len(data) == 0 {
		return data, map[string]any{}, nil
	}
	root, err := parseClaudeJSON(data)
	if err != nil {
		return nil, nil, err
	}
	return data, root, nil
}

// parseClaudeJSON parses JSON bytes into a root map. Returns a non-nil empty map
// for a JSON null body so downstream callers can safely range/assign.
func parseClaudeJSON(data []byte) (map[string]any, error) {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}
	if root == nil {
		root = map[string]any{}
	}
	return root, nil
}

// mutateClaudeJSONAtomic performs a locked, atomic read-modify-write of the
// ~/.claude.json file at configPath (SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-003).
// It guards against concurrent writers (a live Claude Code process or another
// moai invocation) by combining two layers:
//
//  1. An advisory flock on a sibling .lock file — serializes cooperating
//     (moai-vs-moai) writers. Reuses the same internal/lockfile Lock/Unlock
//     family as mutateSettingsLocal (no second lock convention).
//  2. A content compare-and-retry inside the lock — detects non-cooperating
//     writers (e.g., Claude Code, which does not respect the .lock file) and
//     re-applies the mutation to the fresh state so their changes are not lost.
//
// The expensive marshal work is performed OUTSIDE the lock (prep phase); inside
// the lock the guard only re-reads, compares, conditionally re-applies, and
// publishes via temp+rename. This bounds the critical section so a large
// ~/.claude.json (many MCP registrations) does not block a live Claude Code
// process for a human-noticeable duration.
//
// The guard fails OPEN: when flock is unavailable (network filesystems) it
// proceeds with compare-and-retry alone and emits a stderr warning; it never
// blocks a session start indefinitely.
//
// apply receives the freshly-read root map, mutates it in-place, and returns
// changed=true when a write is needed (changed=false signals an idempotent
// skip — no write, no backup). A non-nil error aborts the RMW. apply MUST be
// safe to call multiple times (prep phase + each retry) — it only reads root
// and sets keys, never touches external state.
//
// @MX:ANCHOR: [AUTO] Guarded RMW seam for ~/.claude.json — mirrors mutateSettingsLocal (settings.go:62)
// @MX:REASON: fan_in = 2 (runEnableMCPServerForTool, enableMCPServerIdempotentForTool); the only locked-RMW entry point for ~/.claude.json
func mutateClaudeJSONAtomic(configPath string, apply func(root map[string]any) (changed bool, err error)) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// ── PREP PHASE (outside lock): read, apply, marshal ──────────────────
	prepData, prepRoot, err := readClaudeJSONRaw(configPath)
	if err != nil {
		return err
	}
	prepChanged, err := apply(prepRoot)
	if err != nil {
		return err
	}
	if !prepChanged {
		return nil // idempotent skip — no write needed
	}
	prepOut, err := json.MarshalIndent(prepRoot, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 직렬화 실패: %w", err)
	}

	// Test injection: simulate a concurrent external writer between the prep
	// read and lock acquisition. Production code leaves the hook nil.
	if claudeJSONGuardPreLockHook != nil {
		claudeJSONGuardPreLockHook(configPath)
	}

	// ── GUARDED PUBLISH (inside lock) ────────────────────────────────────
	return withClaudeJSONLock(configPath, func(locked bool) error {
		if !locked {
			_, _ = fmt.Fprintf(os.Stderr,
				"Warning: could not lock %s; proceeding without cross-process guard (best-effort)\n", configPath)
		}

		// Compare the on-disk state against the prep-phase read. If a concurrent
		// writer changed the file, re-apply the mutation on the fresh state so
		// their change is not lost (bounded retry). On the common uncontended
		// path the bytes match immediately and we skip straight to publish.
		publishOut := prepOut
		compareData := prepData
		for attempt := 0; ; attempt++ {
			// A non-cooperating writer (Claude Code itself) can be replacing the
			// file even while we hold the lock, so this read races too.
			data, rerr := atomicfile.ReadFile(configPath)
			if rerr != nil && !os.IsNotExist(rerr) {
				return fmt.Errorf("설정 파일 읽기 실패: %w", rerr)
			}
			if bytes.Equal(data, compareData) {
				break // no concurrent change — prep marshal is valid
			}
			if attempt >= claudeJSONGuardMaxRetries {
				_, _ = fmt.Fprintf(os.Stderr,
					"Warning: %s changed concurrently after %d retries; writing best-effort (a concurrent edit may be overwritten)\n",
					configPath, attempt)
				break
			}
			// Concurrent writer changed the file — re-apply on the fresh state.
			freshRoot, ferr := parseClaudeJSON(data)
			if ferr != nil {
				return ferr
			}
			freshChanged, ferr := apply(freshRoot)
			if ferr != nil {
				return ferr
			}
			if !freshChanged {
				return nil // concurrent writer already achieved the desired state
			}
			mo, merr := json.MarshalIndent(freshRoot, "", "  ")
			if merr != nil {
				return fmt.Errorf("JSON 직렬화 실패: %w", merr)
			}
			publishOut = mo
			compareData = data // fresh state becomes the compare baseline
			time.Sleep(claudeJSONGuardRetryDelay)
		}

		// Backup the current on-disk state, then publish atomically.
		if err := backupClaudeJSON(configPath); err != nil {
			return fmt.Errorf("백업 생성 실패: %w", err)
		}
		return writeClaudeJSONBytes(configPath, publishOut)
	})
}

// withClaudeJSONLock acquires an advisory flock on a sibling .lock file and
// invokes fn. When flock is unavailable (network filesystems, or cross-process
// on Windows where only process-local mutexes exist) it invokes fn without the
// lock (fail-open) — fn receives locked=false so it can emit a warning. The
// guard NEVER blocks a session start indefinitely: a crashed lock-holder
// releases the flock automatically (POSIX flock is tied to process lifetime),
// and unsupported-FS / open errors fall through to the no-lock path immediately.
func withClaudeJSONLock(configPath string, fn func(locked bool) error) error {
	lockPath := configPath + ".lock"
	lockF, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"Warning: could not open lock file %s (%v); proceeding without cross-process guard\n", lockPath, err)
		return fn(false)
	}
	defer func() { _ = lockF.Close() }()

	if err := lockfile.Lock(lockF); err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"Warning: could not lock %s (%v); proceeding without cross-process guard\n", configPath, err)
		return fn(false)
	}
	defer func() { _ = lockfile.Unlock(lockF) }()
	return fn(true)
}

// backupClaudeJSON backs up the current contents of configPath (REQ-GMC-005)
// Does not back up when the file does not exist.
func backupClaudeJSON(configPath string) error {
	data, err := atomicfile.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil // No backup needed when the file does not exist
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

// runEnableMCPServer는 vision(zai-mcp-server) 엔트리를 configPath 에 추가한다 (legacy 시그니처 보존).
// 토큰 불일치 시 에러 (REQ-GMC-006 (b)), 변경 전 백업 (REQ-GMC-005), 다른 엔트리 무변경 (REQ-GMC-010).
// 내부적으로 per-tool dispatch(runEnableMCPServerForTool)에 위임한다 (REQ-GWR-C 리팩터).
//
// @MX:ANCHOR: [AUTO] Core entry point for all GWT scenarios — signature changes affect every test
// @MX:REASON: fan_in = 6 (glmToolsEnableCmd, enableMCPServerIdempotent, runEnableMCPServerScoped, test helpers)
func runEnableMCPServer(configPath string, token string) error {
	return runEnableMCPServerForTool(configPath, "vision", token)
}

// runEnableMCPServerForTool는 도구명에 맞는 서버 집합을 configPath 에 등록한다 (--force 경로, REQ-GWR-C1~C4).
// vision npx 엔트리에 한해 토큰 불일치를 검사한다 (HTTP 엔트리는 토큰을 직접 보유하지 않음).
// 변경 전 백업 (REQ-GMC-005), 다른 mcpServers 엔트리 무변경 (REQ-GMC-010).
//
// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-003: the RMW is guarded by
// mutateClaudeJSONAtomic (flock + compare-retry) so concurrent writes from a
// live Claude Code process are not lost. The apply closure is re-callable:
// the token-mismatch check and mcpServers mutation are re-evaluated against
// the fresh in-lock state on each retry.
func runEnableMCPServerForTool(configPath, toolName, token string) error {
	if token == "" {
		return fmt.Errorf(
			"GLM API 키가 설정되지 않았습니다\n" +
				"  moai glm setup <api-key>",
		)
	}

	entries := buildZAIMCPEntries(toolName, token)

	return mutateClaudeJSONAtomic(configPath, func(root map[string]any) (bool, error) {
		mcpServers := getMCPServers(root)

		// vision npx 엔트리에 한해 기존 토큰 불일치 검사 (REQ-GMC-006)
		if _, want := entries[zaiMCPServerKey]; want {
			if existing, ok := mcpServers[zaiMCPServerKey].(map[string]any); ok {
				existingToken := extractTokenFromEntry(existing)
				if existingToken != token {
					return false, fmt.Errorf(
						"기존 zai-mcp-server 엔트리에 다른 토큰이 설정되어 있습니다\n"+
							"  현재 토큰: %s...%s\n"+
							"  새 토큰:   %s...%s\n\n"+
							"강제 덮어쓰기: moai glm tools enable --force",
						maskPartial(existingToken), maskPartial(existingToken)[len(maskPartial(existingToken))-4:],
						maskPartial(token), maskPartial(token)[len(maskPartial(token))-4:],
					)
				}
			}
		}

		// 요청한 서버 집합 등록 (REQ-GWR-C, REQ-GMC-010 — 다른 엔트리 무변경)
		for key, entry := range entries {
			mcpServers[key] = entry
		}
		root["mcpServers"] = mcpServers
		return true, nil // changed → backup + publish handled by the guard
	})
}

// enableMCPServerIdempotent는 vision 엔트리에 대한 idempotent enable 이다 (legacy 시그니처 보존).
// per-tool dispatch(enableMCPServerIdempotentForTool)에 위임한다.
func enableMCPServerIdempotent(configPath string, token string) (bool, error) {
	return enableMCPServerIdempotentForTool(configPath, "vision", token)
}

// enableMCPServerIdempotentForTool는 도구명에 맞는 서버 집합을 idempotent 하게 등록한다 (REQ-GWR-C7).
// Return values: (skipped bool, err error)
//   - skipped=true: 요청한 모든 서버가 이미 동일하게 등록됨 → 변경 없음, 백업 없음 (REQ-GMC-006 (a), GWT-9)
//   - skipped=false: 하나라도 신규/변경이 필요해 등록을 수행함
//
// vision npx 엔트리에 한해 토큰 불일치 시 에러를 반환한다 (REQ-GMC-006 (b)).
//
// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-003: the RMW is guarded by
// mutateClaudeJSONAtomic (flock + compare-retry). The skipped flag reflects the
// LAST apply invocation (prep or the final retry) so it stays correct even when
// a concurrent writer achieves the desired state mid-guard.
func enableMCPServerIdempotentForTool(configPath, toolName, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf(
			"GLM API 키가 설정되지 않았습니다\n" +
				"  moai glm setup <api-key>",
		)
	}

	entries := buildZAIMCPEntries(toolName, token)
	var skipped bool

	err := mutateClaudeJSONAtomic(configPath, func(root map[string]any) (bool, error) {
		mcpServers := getMCPServers(root)

		// vision npx 엔트리 토큰 검사 — 불일치 시 에러 (REQ-GMC-006 (b))
		if _, want := entries[zaiMCPServerKey]; want {
			if existing, ok := mcpServers[zaiMCPServerKey].(map[string]any); ok {
				if extractTokenFromEntry(existing) != token {
					return false, fmt.Errorf(
						"기존 zai-mcp-server 엔트리에 다른 토큰이 설정되어 있습니다\n" +
							"강제 덮어쓰기: moai glm tools enable --force",
					)
				}
			}
		}

		// 요청한 서버 중 신규/변경이 필요한 것이 하나라도 있는지 판정
		needsWrite := false
		for key, entry := range entries {
			existing, ok := mcpServers[key].(map[string]any)
			if !ok || !mcpEntryEqual(existing, entry) {
				needsWrite = true
				break
			}
		}

		if !needsWrite {
			// 모든 서버가 이미 동일하게 등록됨 → idempotent skip (백업 없음)
			skipped = true
			return false, nil
		}

		// 요청한 서버 집합 등록
		for key, entry := range entries {
			mcpServers[key] = entry
		}
		root["mcpServers"] = mcpServers
		skipped = false
		return true, nil // changed → backup + publish handled by the guard
	})

	return skipped, err
}

// autoEnableMCPServer attempts to enable Z.AI MCP server during GLM launch.
// Non-blocking: warns on stderr but never returns error.
// Skips if MOAI_GLM_NO_AUTO_TOOLS=1, no token, or already enabled with same token.
func autoEnableMCPServer() {
	if os.Getenv("MOAI_GLM_NO_AUTO_TOOLS") == "1" {
		return
	}

	token := loadGLMKey()
	if token == "" {
		return
	}

	configPath, err := resolveConfigPath("user")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: MCP auto-enable config path resolution failed: %v\n", err)
		return
	}

	skipped, err := enableMCPServerIdempotent(configPath, token)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Z.AI MCP auto-enable failed: %v\n", err)
		_, _ = fmt.Fprintln(os.Stderr, "  MCP tools (Vision, Web Search, Web Reader) may not be available.")
		_, _ = fmt.Fprintln(os.Stderr, "  Manual enable: moai glm tools enable all")
		return
	}

	if !skipped {
		_, _ = fmt.Fprintln(os.Stderr, "Z.AI MCP tools auto-enabled (Vision, Web Search, Web Reader)")
		_, _ = fmt.Fprintln(os.Stderr, "  Disable: moai glm tools disable all")
	}
}

// runEnableMCPServerScoped performs enable against project scope (.mcp.json) (REQ-GMC-008)
func runEnableMCPServerScoped(mcpJSONPath string, token string) error {
	return runEnableMCPServer(mcpJSONPath, token)
}

// disableMCPServerSafe는 vision(zai-mcp-server) 엔트리만 configPath 에서 제거한다 (legacy 시그니처 보존).
// per-tool dispatch(disableMCPServerForTool)에 위임한다.
//
// @MX:ANCHOR: [AUTO] Core implementation of REQ-GMC-004/010 — partial-delete safety
// @MX:REASON: fan_in = 4 (runGLMToolsDisable, many GWT tests)
func disableMCPServerSafe(configPath string) (bool, error) {
	return disableMCPServerForTool(configPath, "vision")
}

// disableMCPServerForTool는 도구명에 대응하는 서버만 제거한다 (REQ-GWR-C6, REQ-GMC-010).
// 무관한 mcpServers 엔트리(context7, chrome-devtools 등)는 보존한다.
// Return values: (removed bool, err error) — 요청한 서버가 하나도 없으면 removed=false (idempotent skip).
func disableMCPServerForTool(configPath, toolName string) (bool, error) {
	root, err := readClaudeJSON(configPath)
	if err != nil {
		return false, err
	}

	mcpServers := getMCPServers(root)

	// 제거 대상 중 실제로 존재하는 서버 키 수집
	var present []string
	for _, key := range toolServerKeys(toolName) {
		if _, ok := mcpServers[key]; ok {
			present = append(present, key)
		}
	}

	if len(present) == 0 {
		// 제거할 엔트리 없음 → idempotent skip
		return false, nil
	}

	// Back up then remove (REQ-GMC-005)
	if err := backupClaudeJSON(configPath); err != nil {
		return false, fmt.Errorf("백업 생성 실패: %w", err)
	}

	for _, key := range present {
		delete(mcpServers, key)
	}
	root["mcpServers"] = mcpServers

	jsonBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return true, fmt.Errorf("JSON 직렬화 실패: %w", err)
	}
	return true, writeClaudeJSONBytes(configPath, jsonBytes)
}

// mcpEntryEqual는 두 MCP 엔트리가 의미상 동일한지 판정한다 (idempotency 검사용).
// JSON round-trip 시 []string→[]any, map[string]string→map[string]any 로 변하므로
// JSON 정규화 후 바이트 비교한다 (REQ-GWR-C7).
func mcpEntryEqual(a, b map[string]any) bool {
	ab, err := json.Marshal(normalizeForCompare(a))
	if err != nil {
		return false
	}
	bb, err := json.Marshal(normalizeForCompare(b))
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}

// normalizeForCompare는 []string / map[string]string 를 []any / map[string]any 로 변환해
// JSON unmarshal 결과와 구조를 일치시킨다.
func normalizeForCompare(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = normalizeForCompare(val)
		}
		return out
	case map[string]string:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = val
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = normalizeForCompare(val)
		}
		return out
	case []string:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = val
		}
		return out
	default:
		return v
	}
}

// ─── Internal helper functions ─────────────────────────────────────────────

// getMCPServers extracts the mcpServers map from the root JSON.
// Returns an empty map and sets it on root when absent.
func getMCPServers(root map[string]any) map[string]any {
	if existing, ok := root["mcpServers"].(map[string]any); ok {
		return existing
	}
	m := map[string]any{}
	root["mcpServers"] = m
	return m
}

// extractTokenFromEntry extracts the env.Z_AI_API_KEY value from an MCP entry.
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

// maskPartial masks part of a token (for log display)
func maskPartial(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****"
}

// ─── Node.js version detection ─────────────────────────────────────────────

// checkNodeVersion verifies that Node.js is installed and meets the minimum version (REQ-GMC-009).
// Shared verification function for GWT-14 (missing) and GWT-15 (older version) scenarios.
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

// detectNodeVersion detects the node major version via PATH (REQ-GMC-009).
// Return values: (major int, versionString string, error)
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

// parseNodeMajorVersion extracts the major integer from a version string in the "v22.5.0" format.
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
