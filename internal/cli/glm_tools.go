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

// ─── Constants ─────────────────────────────────────────────────────────────

const (
	// zaiMCPServerKey is the key name of the Z.AI MCP server used in mcpServers
	zaiMCPServerKey = "zai-mcp-server"

	// zaiNPXPackage is the package name executed via npx
	zaiNPXPackage = "@z_ai/mcp-server@latest"

	// nodeMinMajorVersion is the minimum Node.js major version required to run npx
	nodeMinMajorVersion = 22
)

// Supported tool name list
var validToolNames = map[string]bool{
	"vision":    true,
	"websearch": true,
	"webreader": true,
	"all":       true,
}

// errNodeNotFound is the sentinel error returned when node is not found on PATH
var errNodeNotFound = errors.New("no Node.js executable found on PATH")

// ─── Test injection points (function variables) ───────────────────────────

// userHomeDirFn is the home-directory lookup function variable (for test override)
var userHomeDirFn = userHomeDir

// detectNodeFn is the node-version detection function variable (for test override)
// Return values: (major version int, version string e.g. "v22.5.0", error)
var detectNodeFn = detectNodeVersion

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

	// (c) Validate Node.js version (REQ-GMC-009, GWT-14, GWT-15)
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

	// (d) Determine scope (REQ-GMC-008, GWT-13)
	scope, _ := cmd.Flags().GetString("scope")
	force, _ := cmd.Flags().GetBool("force")

	configPath, err := resolveConfigPath(scope)
	if err != nil {
		return fmt.Errorf("설정 파일 경로 결정 실패: %w", err)
	}

	// (e) Enable after handling existing entry + force flag (REQ-GMC-003, REQ-GMC-006)
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

	// (f) Emit success message (REQ-GMC-003, GWT-5)
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

	// Remove entry (REQ-GMC-004, GWT-6, GWT-7)
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

// buildZAIMCPEntry constructs a Z.AI MCP server entry (REQ-GMC-003)
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

// writeClaudeJSONAtomic performs an atomic JSON write to configPath (REQ-GMC-005, R7)
// Atomic write: temp file → os.Rename (POSIX atomicity guarantee)
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
	defer func() { _ = os.Remove(tmpName) }() // Clean up the temp file on failure

	if _, err := tmp.Write(jsonBytes); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("임시 파일 쓰기 실패: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("임시 파일 닫기 실패: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpName, configPath); err != nil {
		return fmt.Errorf("파일 교체 실패: %w", err)
	}
	return nil
}

// backupClaudeJSON backs up the current contents of configPath (REQ-GMC-005)
// Does not back up when the file does not exist.
func backupClaudeJSON(configPath string) error {
	data, err := os.ReadFile(configPath)
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

// runEnableMCPServer adds the zai-mcp-server entry to configPath.
// Returns an error on token mismatch (REQ-GMC-006 (b)).
// Creates a backup before modifying the file (REQ-GMC-005).
// Never modifies any other mcpServers entry (REQ-GMC-010).
//
// @MX:ANCHOR: [AUTO] Core entry point for all 22 GWT scenarios — signature changes affect every test
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

	// Acquire the mcpServers map
	mcpServers := getMCPServers(root)

	// Inspect token of the existing entry (REQ-GMC-006)
	if existing, ok := mcpServers[zaiMCPServerKey].(map[string]any); ok {
		existingToken := extractTokenFromEntry(existing)
		if existingToken != token {
			// Token mismatch → reject + --force guidance
			return fmt.Errorf(
				"기존 zai-mcp-server 엔트리에 다른 토큰이 설정되어 있습니다\n"+
					"  현재 토큰: %s...%s\n"+
					"  새 토큰:   %s...%s\n\n"+
					"강제 덮어쓰기: moai glm tools enable --force",
				maskPartial(existingToken), maskPartial(existingToken)[len(maskPartial(existingToken))-4:],
				maskPartial(token), maskPartial(token)[len(maskPartial(token))-4:],
			)
		}
		// Token match → already handled (idempotent: handled by enableMCPServerIdempotent)
	}

	// Create backup (REQ-GMC-005)
	if err := backupClaudeJSON(configPath); err != nil {
		return fmt.Errorf("백업 생성 실패: %w", err)
	}

	// Add the entry (REQ-GMC-003, REQ-GMC-010)
	mcpServers[zaiMCPServerKey] = buildZAIMCPEntry(token)
	root["mcpServers"] = mcpServers

	return writeClaudeJSONAtomic(configPath, root)
}

// enableMCPServerIdempotent performs an idempotent enable.
// Return values: (skipped bool, err error)
//   - skipped=true: already registered with the same token → no change, no backup (REQ-GMC-006 (a))
//   - skipped=false: performs a new registration
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

	// Check existing entry
	if existing, ok := mcpServers[zaiMCPServerKey].(map[string]any); ok {
		existingToken := extractTokenFromEntry(existing)
		if existingToken == token {
			// Token match → idempotent skip (REQ-GMC-006 (a), GWT-9: no backup)
			return true, nil
		}
		// Token mismatch → return error (REQ-GMC-006 (b))
		return false, fmt.Errorf(
			"기존 zai-mcp-server 엔트리에 다른 토큰이 설정되어 있습니다\n" +
				"강제 덮어쓰기: moai glm tools enable --force",
		)
	}

	// New registration → back up then write
	if err := backupClaudeJSON(configPath); err != nil {
		return false, fmt.Errorf("백업 생성 실패: %w", err)
	}

	mcpServers[zaiMCPServerKey] = buildZAIMCPEntry(token)
	root["mcpServers"] = mcpServers

	return false, writeClaudeJSONAtomic(configPath, root)
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

// disableMCPServerSafe removes only the zai-mcp-server entry from configPath (REQ-GMC-004, REQ-GMC-010).
// Does not modify other mcpServers entries.
// Return values: (removed bool, err error) — removed=false when the entry is absent (idempotent)
//
// @MX:ANCHOR: [AUTO] Core implementation of REQ-GMC-004/010 — partial-delete safety
// @MX:REASON: fan_in = 4 (runGLMToolsDisable, many GWT tests)
func disableMCPServerSafe(configPath string) (bool, error) {
	root, err := readClaudeJSON(configPath)
	if err != nil {
		return false, err
	}

	mcpServers := getMCPServers(root)

	if _, ok := mcpServers[zaiMCPServerKey]; !ok {
		// No entry → idempotent skip
		return false, nil
	}

	// Back up then remove (REQ-GMC-005)
	if err := backupClaudeJSON(configPath); err != nil {
		return false, fmt.Errorf("백업 생성 실패: %w", err)
	}

	delete(mcpServers, zaiMCPServerKey)
	root["mcpServers"] = mcpServers

	return true, writeClaudeJSONAtomic(configPath, root)
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
