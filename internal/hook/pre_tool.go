// Resolution: KEEP — security scan, secrets, reflective write; emit PermissionDecision.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/quality"
	"github.com/modu-ai/moai-adk/internal/hook/security"
	"golang.org/x/text/unicode/norm"
)

// HARNESS_FROZEN_* sentinels are emitted when a harness-learner subagent attempts to
// modify a FROZEN zone file via Write/Edit. These sentinel strings appear in the deny
// reason returned to Claude Code (PermissionDecision deny), enabling orchestrator-side
// pattern matching without AskUserQuestion interaction.
//
// Vision §3.4: 8 frozen sentinel types cover the full FROZEN zone taxonomy.
// W3 is the first runtime implementer of these sentinels (zone-registry SSOT from W1).
//
// [HARD] No AskUserQuestion calls. Deny reason is emitted as sentinel string.
// Orchestrator handles user notification via blocker report pattern (CLAUDE.md §8).
//
// @MX:ANCHOR: [AUTO] HARNESS_FROZEN_* sentinel catalog — PreToolUse W3 first implementer.
// @MX:REASON: [AUTO] fan_in >= 3: pre_tool.go (emit), integration_test.go (verify), sentinel_catalog_test.go (CI guard)
const (
	// SentinelHarnessFrozenAgent is emitted when a harness-learner attempts to modify
	// a .claude/agents/moai/ path (frozen agent definition files).
	SentinelHarnessFrozenAgent = "HARNESS_FROZEN_AGENT_VIOLATION"

	// SentinelHarnessFrozenSkill is emitted when a harness-learner attempts to modify
	// a .claude/skills/moai-* path (frozen skill definition files).
	SentinelHarnessFrozenSkill = "HARNESS_FROZEN_SKILL_VIOLATION"

	// SentinelHarnessFrozenRule is emitted when a harness-learner attempts to modify
	// a .claude/rules/moai/ path (frozen rule files).
	SentinelHarnessFrozenRule = "HARNESS_FROZEN_RULE_VIOLATION"

	// SentinelHarnessFrozenCommand is emitted when a harness-learner attempts to modify
	// a .claude/commands/ path (frozen command files).
	SentinelHarnessFrozenCommand = "HARNESS_FROZEN_COMMAND_VIOLATION"

	// SentinelHarnessFrozenHook is emitted when a harness-learner attempts to modify
	// a .claude/hooks/ path (frozen hook handler files).
	SentinelHarnessFrozenHook = "HARNESS_FROZEN_HOOK_VIOLATION"

	// SentinelHarnessFrozenOutputStyle is emitted when a harness-learner attempts to modify
	// a .claude/output-styles/ path (frozen output style files).
	SentinelHarnessFrozenOutputStyle = "HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION"

	// SentinelHarnessFrozenInstruction is emitted when a harness-learner attempts to modify
	// the main CLAUDE.md or CLAUDE.local.md instruction files.
	SentinelHarnessFrozenInstruction = "HARNESS_FROZEN_INSTRUCTION_VIOLATION"

	// SentinelHarnessFrozenConfig is emitted when a harness-learner attempts to modify
	// a .moai/project/brand/ path or other frozen config area.
	SentinelHarnessFrozenConfig = "HARNESS_FROZEN_CONFIG_VIOLATION"

	// harnessLearnerIdentity is the agent name used by the harness-learner subagent.
	// Write/Edit from this agent to FROZEN paths is checked via frozenZonePrefixes.
	harnessLearnerIdentity = "harness-learner"
)

// SecurityPolicy defines tool access control rules for PreToolUse events.
type SecurityPolicy struct {
	// BlockedTools is a list of tool names that are always blocked.
	BlockedTools []string

	// DenyPatterns are regex patterns for files that should NEVER be modified.
	DenyPatterns []*regexp.Regexp

	// AskPatterns are regex patterns for files that require user confirmation.
	AskPatterns []*regexp.Regexp

	// DangerousBashPatterns are regex patterns for dangerous Bash commands.
	DangerousBashPatterns []*regexp.Regexp

	// AskBashPatterns are regex patterns for Bash commands requiring confirmation.
	AskBashPatterns []*regexp.Regexp

	// SensitiveContentPatterns are regex patterns for sensitive data in content.
	SensitiveContentPatterns []*regexp.Regexp

	// AllowedExternalPaths are absolute directory paths outside the project that
	// are permitted for file access. This bypasses the project-boundary check
	// for specific well-known directories (e.g., ~/.claude/plans/).
	AllowedExternalPaths []string
}

// compilePatterns compiles a list of pattern strings into regexp objects.
func compilePatterns(patterns []string) []*regexp.Regexp {
	result := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile("(?i)" + p) // Case-insensitive
		if err != nil {
			slog.Warn("failed to compile security pattern", "pattern", p, "error", err)
			continue
		}
		result = append(result, re)
	}
	return result
}

// DefaultSecurityPolicy returns a SecurityPolicy with comprehensive security patterns
// ported from the Python pre_tool__security_guard.py implementation.
func DefaultSecurityPolicy() *SecurityPolicy {
	// Files that should NEVER be modified
	denyPatterns := []string{
		// Secrets and credentials
		`secrets?\.(json|ya?ml|toml)$`,
		`credentials?\.(json|ya?ml|toml)$`,
		`\.secrets/.*`,
		`secrets/.*`,
		// SSH and certificates
		`\.ssh/.*`,
		`id_rsa.*`,
		`id_ed25519.*`,
		`\.pem$`,
		`\.key$`,
		`\.crt$`,
		// Git internals
		`\.git/.*`,
		// Cloud credentials
		`\.aws/.*`,
		`\.gcloud/.*`,
		`\.azure/.*`,
		`\.kube/.*`,
		// Token files
		`\.token$`,
		`\.tokens/.*`,
		`auth\.json$`,
	}

	// Files that require user confirmation
	askPatterns := []string{
		// Lock files
		`package-lock\.json$`,
		`yarn\.lock$`,
		`pnpm-lock\.ya?ml$`,
		`Gemfile\.lock$`,
		`Cargo\.lock$`,
		`poetry\.lock$`,
		`composer\.lock$`,
		`Pipfile\.lock$`,
		`uv\.lock$`,
		// Critical configs
		`tsconfig\.json$`,
		`pyproject\.toml$`,
		`Cargo\.toml$`,
		`package\.json$`,
		`docker-compose\.ya?ml$`,
		`Dockerfile$`,
		`\.dockerignore$`,
		// CI/CD configs
		`\.github/workflows/.*\.ya?ml$`,
		`\.gitlab-ci\.ya?ml$`,
		`\.circleci/.*`,
		`Jenkinsfile$`,
		// Infrastructure
		`terraform/.*\.tf$`,
		`\.terraform/.*`,
		`kubernetes/.*\.ya?ml$`,
		`k8s/.*\.ya?ml$`,
	}

	// Dangerous Bash commands that should NEVER be executed
	dangerousBashPatterns := []string{
		// Database deletion commands - Supabase
		`supabase\s+db\s+reset`,
		`supabase\s+projects?\s+delete`,
		`supabase\s+functions?\s+delete`,
		// Database deletion commands - Neon
		`neon\s+database\s+delete`,
		`neon\s+projects?\s+delete`,
		`neon\s+branch\s+delete`,
		// Database deletion commands - PlanetScale
		`pscale\s+database\s+delete`,
		`pscale\s+branch\s+delete`,
		// Database deletion commands - Railway
		`railway\s+delete`,
		`railway\s+environment\s+delete`,
		// Database deletion commands - Vercel
		`vercel\s+env\s+rm`,
		`vercel\s+projects?\s+rm`,
		// SQL dangerous commands
		`DROP\s+DATABASE`,
		`DROP\s+SCHEMA`,
		`TRUNCATE\s+TABLE`,
		// Unix dangerous file operations
		`rm\s+-rf\s+/`,
		`rm\s+-rf\s+~`,
		`rm\s+-rf\s+\*`,
		`rm\s+-rf\s+\.\*`,
		`rm\s+-rf\s+\.git\b`,
		`rm\s+-rf\s+node_modules\s*$`,
		// Windows dangerous file operations (CMD)
		`rd\s+/s\s+/q\s+[A-Za-z]:\\`,
		`rmdir\s+/s\s+/q\s+[A-Za-z]:\\`,
		`del\s+/f\s+/q\s+[A-Za-z]:\\`,
		`rd\s+/s\s+/q\s+\\\\`,
		`rd\s+/s\s+/q\s+\.git\b`,
		`del\s+/s\s+/q\s+\*\.\*`,
		`format\s+[A-Za-z]:`,
		// Windows dangerous file operations (PowerShell)
		`Remove-Item\s+.*-Recurse\s+.*-Force\s+[A-Za-z]:\\`,
		`Remove-Item\s+.*-Recurse\s+.*-Force\s+~`,
		`Remove-Item\s+.*-Recurse\s+.*-Force\s+\$env:`,
		`Remove-Item\s+.*-Recurse\s+.*-Force\s+\.git\b`,
		`Clear-Content\s+.*-Force`,
		// Git dangerous commands
		`git\s+push\s+.*--force\s+origin\s+(main|master)`,
		`git\s+branch\s+-D\s+(main|master)`,
		// Cloud infrastructure deletion
		`terraform\s+destroy`,
		`pulumi\s+destroy`,
		`aws\s+.*\s+delete-`,
		`gcloud\s+.*\s+delete\b`,
		// Azure CLI dangerous commands
		`az\s+group\s+delete`,
		`az\s+storage\s+account\s+delete`,
		`az\s+sql\s+server\s+delete`,
		// Docker dangerous commands
		`docker\s+system\s+prune\s+(-a|--all)`,
		`docker\s+image\s+prune\s+(-a|--all)`,
		`docker\s+container\s+prune`,
		`docker\s+volume\s+prune`,
		`docker\s+network\s+prune`,
		`docker\s+builder\s+prune\s+(-a|--all)`,
		// Classic dangerous patterns
		`:\(\)\{\s*:\|:&\s*\};:`, // Fork bomb
		`mkfs\.`,
		`>\s*/dev/sda`,
		`dd\s+if=/dev/zero\s+of=/dev/sda`,
	}

	// Bash commands that require user confirmation
	askBashPatterns := []string{
		// Database reset/migration
		`prisma\s+migrate\s+reset`,
		`prisma\s+db\s+push\s+--force`,
		`drizzle-kit\s+push`,
		// Git force operations (non-main branches)
		`git\s+push\s+.*--force`,
		`git\s+reset\s+--hard`,
		`git\s+clean\s+-fd`,
		// Package manager cache clear
		`npm\s+cache\s+clean`,
		`yarn\s+cache\s+clean`,
		`pnpm\s+store\s+prune`,
	}

	// Content patterns that indicate sensitive data
	sensitiveContentPatterns := []string{
		`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`,
		`-----BEGIN\s+CERTIFICATE-----`,
		`sk-[a-zA-Z0-9]{32,}`,       // OpenAI API keys
		`ghp_[a-zA-Z0-9]{36}`,       // GitHub tokens
		`gho_[a-zA-Z0-9]{36}`,       // GitHub OAuth tokens
		`glpat-[a-zA-Z0-9\-]{20}`,   // GitLab tokens
		`xox[baprs]-[a-zA-Z0-9\-]+`, // Slack tokens
		`AKIA[0-9A-Z]{16}`,          // AWS access keys
		`ya29\.[a-zA-Z0-9_\-]+`,     // Google OAuth tokens
	}

	// Resolve allowed external paths that bypass the project-boundary check.
	var allowedExternal []string
	if home, err := os.UserHomeDir(); err == nil {
		allowedExternal = append(allowedExternal, filepath.Join(home, ".claude"))
		allowedExternal = append(allowedExternal, filepath.Join(home, ".moai"))
	}

	return &SecurityPolicy{
		BlockedTools:             []string{},
		DenyPatterns:             compilePatterns(denyPatterns),
		AskPatterns:              compilePatterns(askPatterns),
		DangerousBashPatterns:    compilePatterns(dangerousBashPatterns),
		AskBashPatterns:          compilePatterns(askBashPatterns),
		SensitiveContentPatterns: compilePatterns(sensitiveContentPatterns),
		AllowedExternalPaths:     allowedExternal,
	}
}

// MergeExtraPatterns appends additional security patterns from external config
// to the policy. Extra patterns extend (never replace) the built-in defaults
// per SOLID O-Principle (REQ-SEC-003, REQ-SEC-004).
func (p *SecurityPolicy) MergeExtraPatterns(extra *security.ExtraSecurityConfig) {
	if extra == nil {
		return
	}
	if len(extra.Security.ExtraDangerousBashPatterns) > 0 {
		p.DangerousBashPatterns = append(p.DangerousBashPatterns, compilePatterns(extra.Security.ExtraDangerousBashPatterns)...)
	}
	if len(extra.Security.ExtraDenyPatterns) > 0 {
		p.DenyPatterns = append(p.DenyPatterns, compilePatterns(extra.Security.ExtraDenyPatterns)...)
	}
	if len(extra.Security.ExtraAskPatterns) > 0 {
		p.AskPatterns = append(p.AskPatterns, compilePatterns(extra.Security.ExtraAskPatterns)...)
	}
	if len(extra.Security.ExtraSensitiveContentPatterns) > 0 {
		p.SensitiveContentPatterns = append(p.SensitiveContentPatterns, compilePatterns(extra.Security.ExtraSensitiveContentPatterns)...)
	}
}

// preToolHandler processes PreToolUse events.
// It enforces security policies by checking tool names against blocklists
// and scanning tool input for dangerous patterns (REQ-HOOK-031, REQ-HOOK-032).
// Optionally integrates with SecurityScanner for AST-based security scanning.
type preToolHandler struct {
	cfg        ConfigProvider
	policy     *SecurityPolicy
	scanner    *security.SecurityScanner
	projectDir string
}

// NewPreToolHandler creates a new PreToolUse event handler with the given security policy.
func NewPreToolHandler(cfg ConfigProvider, policy *SecurityPolicy) Handler {
	projectDir := os.Getenv(config.EnvClaudeProjectDir)
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	return &preToolHandler{cfg: cfg, policy: policy, projectDir: projectDir}
}

// NewPreToolHandlerWithScanner creates a PreToolUse handler with AST-based security scanning.
// If scanner is nil or unavailable, falls back to pattern-based security only.
func NewPreToolHandlerWithScanner(cfg ConfigProvider, policy *SecurityPolicy, scanner *security.SecurityScanner) Handler {
	projectDir := os.Getenv(config.EnvClaudeProjectDir)
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}

	// Validate scanner availability
	if scanner != nil && !scanner.IsAvailable() {
		slog.Info("ast-grep not available, security scanning disabled")
		scanner = nil
	}

	return &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    scanner,
		projectDir: projectDir,
	}
}

// EventType returns EventPreToolUse.
func (h *preToolHandler) EventType() EventType {
	return EventPreToolUse
}

// Handle processes a PreToolUse event. It checks the tool name against the
// blocklist and scans tool input for dangerous patterns. Returns Decision
// "deny" with a reason if the tool is denied, "ask" if user confirmation is
// needed, or "allow" otherwise.
func (h *preToolHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// No policy means allow everything
	if h.policy == nil {
		return NewAllowOutput(), nil
	}

	slog.Debug("checking tool security",
		"tool_name", input.ToolName,
		"session_id", input.SessionID,
	)

	// Check if tool is in the blocked list
	for _, blocked := range h.policy.BlockedTools {
		if strings.EqualFold(input.ToolName, blocked) {
			reason := fmt.Sprintf("tool %q is blocked by security policy", input.ToolName)
			slog.Warn("tool blocked",
				"tool_name", input.ToolName,
				"reason", reason,
			)
			return NewDenyOutput(reason), nil
		}
	}

	// Handle Bash commands
	if input.ToolName == "Bash" && len(input.ToolInput) > 0 {
		// Quality gate for git commit commands (REQ-GATE-001).
		// Executes before security pattern checks so the gate cannot be bypassed.
		if command := h.extractBashCommand(input.ToolInput); quality.IsGitCommit(command) {
			gate := quality.NewQualityGate(h.loadGateConfig())
			passed, output := gate.Run(ctx)
			if !passed {
				slog.Warn("quality gate blocked git commit",
					"reason_summary", firstLine(output),
				)
				return NewDenyOutput(output), nil
			}
		}

		decision, reason := h.checkBashCommand(input.ToolInput)
		if decision != "" {
			slog.Warn("bash command security check",
				"tool_name", input.ToolName,
				"decision", decision,
				"reason", reason,
			)
			if decision == DecisionDeny {
				return NewDenyOutput(reason), nil
			}
			if decision == DecisionAsk {
				return NewAskOutput(reason), nil
			}
		}
	}

	// Handle Write and Edit tools
	if (input.ToolName == "Write" || input.ToolName == "Edit") && len(input.ToolInput) > 0 {
		// Harness-learner FROZEN zone guard (Vision §3.4, W3 first implementer).
		// Must run before general file-access check to emit typed sentinel strings.
		// Uses AgentType (custom agent name from --agent flag) to identify harness-learner.
		if sentinel, reason := h.checkHarnessFrozenZoneFromInput(input.AgentType, input.ToolInput); sentinel != "" {
			slog.Warn("harness frozen zone violation",
				"sentinel", sentinel,
				"agent_id", input.AgentID,
				"tool_name", input.ToolName,
			)
			return NewDenyOutput(reason), nil
		}

		decision, reason := h.checkFileAccess(input.ToolInput, input.ToolName)
		if decision != "" {
			slog.Warn("file access security check",
				"tool_name", input.ToolName,
				"decision", decision,
				"reason", reason,
			)
			if decision == DecisionDeny {
				return NewDenyOutput(reason), nil
			}
			if decision == DecisionAsk {
				return NewAskOutput(reason), nil
			}
		}

		// AST-based security scanning for Write operations
		if input.ToolName == "Write" && h.scanner != nil {
			decision, reason := h.scanWriteContent(ctx, input.ToolInput)
			if decision == DecisionDeny {
				return NewDenyOutput(reason), nil
			}
		}
	}

	return NewAllowOutput(), nil
}

// scanWriteContent scans the content to be written using AST-based security scanner.
// Creates a temporary file with the content, scans it, and returns the result.
// Returns (decision, reason) where decision is "deny" or "" for allow.
func (h *preToolHandler) scanWriteContent(ctx context.Context, toolInput json.RawMessage) (string, string) {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return "", ""
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return "", ""
	}

	content, ok := parsed["content"].(string)
	if !ok || content == "" {
		return "", ""
	}

	// Check if file extension is supported for scanning
	ext := filepath.Ext(filePath)
	if !security.IsSupportedExtension(ext) {
		slog.Debug("file extension not supported for security scanning",
			"file_path", filePath,
			"extension", ext,
		)
		return "", ""
	}

	// Create temporary file with the content
	tmpFile, err := os.CreateTemp("", "moai-security-scan-*"+ext)
	if err != nil {
		slog.Warn("failed to create temp file for security scan", "error", err)
		return "", ""
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	if _, err := tmpFile.WriteString(content); err != nil {
		slog.Warn("failed to write content to temp file", "error", err)
		return "", ""
	}
	_ = tmpFile.Close() // Close before scanning

	// Scan the temporary file
	result, err := h.scanner.ScanFile(ctx, tmpFile.Name(), h.projectDir)
	if err != nil {
		slog.Warn("security scan failed", "error", err)
		return "", ""
	}

	if result == nil || !result.Scanned {
		return "", ""
	}

	// Check for error-severity findings (REQ-HOOK-131)
	if h.scanner.ShouldAlert(result) {
		report := h.scanner.GetReport(result, filePath)
		reason := fmt.Sprintf("Security vulnerabilities detected in %s:\n%s", filepath.Base(filePath), report)
		slog.Warn("security scan blocked write operation",
			"file_path", filePath,
			"error_count", result.ErrorCount,
			"warning_count", result.WarningCount,
		)
		return DecisionDeny, reason
	}

	// Log warnings but allow the operation
	if result.WarningCount > 0 {
		slog.Info("security scan found warnings",
			"file_path", filePath,
			"warning_count", result.WarningCount,
		)
	}

	return "", ""
}

// extractBashCommand parses the command string from Bash tool input JSON.
// Returns "" if the input cannot be parsed or lacks a command field.
func (h *preToolHandler) extractBashCommand(toolInput json.RawMessage) string {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return ""
	}
	command, _ := parsed["command"].(string)
	return command
}

// loadGateConfig reads gate configuration from the config provider.
// Falls back to DefaultGateConfig when the config is not available.
func (h *preToolHandler) loadGateConfig() *quality.GateConfig {
	if h.cfg == nil {
		return quality.DefaultGateConfig()
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return quality.DefaultGateConfig()
	}
	gate := cfg.Gate
	qcfg := &quality.GateConfig{
		Enabled:     gate.Enabled,
		SkipTests:   gate.SkipTests,
		VetTimeout:  gate.VetTimeoutDuration(),
		LintTimeout: gate.LintTimeoutDuration(),
		TestTimeout: gate.TestTimeoutDuration(),
		ProjectDir:  h.projectDir,
	}
	// Map config.AstGrepGateConfig → quality.AstGrepGateConfig (SPEC-SLQG-001).
	ag := gate.AstGrepGate
	qcfg.AstGrepGate = &quality.AstGrepGateConfig{
		Enabled:      ag.Enabled,
		RulesDir:     ag.RulesDir,
		BlockOnError: ag.BlockOnError,
		WarnOnlyMode: ag.WarnOnlyMode,
	}
	// Apply defaults when RulesDir is empty (not set in YAML).
	if qcfg.AstGrepGate.RulesDir == "" {
		qcfg.AstGrepGate.RulesDir = ".moai/config/astgrep-rules"
	}
	return qcfg
}

// firstLine returns the first non-empty line of s, or s itself when there is no newline.
func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

// checkBashCommand checks a Bash command against dangerous and ask patterns.
// Returns (decision, reason) where decision is "deny", "ask", or "" for allow.
func (h *preToolHandler) checkBashCommand(toolInput json.RawMessage) (string, string) {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return "", ""
	}

	command, ok := parsed["command"].(string)
	if !ok || command == "" {
		return "", ""
	}

	// Check dangerous patterns (deny)
	for _, pattern := range h.policy.DangerousBashPatterns {
		if pattern.MatchString(command) {
			return DecisionDeny, fmt.Sprintf("Dangerous command blocked: %s", pattern.String())
		}
	}

	// Check ask patterns (require confirmation)
	for _, pattern := range h.policy.AskBashPatterns {
		if pattern.MatchString(command) {
			return DecisionAsk, "This command may have significant effects. Please confirm."
		}
	}

	return "", ""
}

// checkFileAccess checks file path and content against security patterns.
// Returns (decision, reason) where decision is "deny", "ask", or "" for allow.
func (h *preToolHandler) checkFileAccess(toolInput json.RawMessage, toolName string) (string, string) {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return "", ""
	}

	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return "", ""
	}

	// Resolve path to prevent path traversal attacks
	resolvedPath, err := filepath.Abs(filePath)
	if err != nil {
		return DecisionDeny, "Invalid file path: cannot resolve"
	}

	// Check if path is within project directory
	if h.projectDir != "" {
		projectAbs, absErr := filepath.Abs(h.projectDir)
		if absErr != nil {
			// Cannot resolve project directory, skip boundary check
			slog.Debug("cannot resolve project directory", "error", absErr)
		} else {
			// Normalize both paths to Unicode NFC before comparison.
			// macOS HFS+/APFS stores paths in NFD form, but tools like
			// Claude Code may send paths in NFC form. Without normalization,
			// filepath.Rel produces ".." prefixed results for paths containing
			// non-ASCII characters (e.g., Korean), causing false path traversal errors.
			nfcProject := norm.NFC.String(projectAbs)
			nfcResolved := norm.NFC.String(resolvedPath)

			rel, relErr := filepath.Rel(nfcProject, nfcResolved)
			if relErr != nil || strings.HasPrefix(rel, "..") {
				// Before denying, check if path is under an allowed external directory.
				if !h.isAllowedExternalPath(nfcResolved) {
					return DecisionDeny, "Path traversal detected: file is outside project directory"
				}
			}
		}
	}

	// Normalize path for pattern matching
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	normalizedResolved := strings.ReplaceAll(resolvedPath, "\\", "/")

	// Check deny patterns
	for _, pattern := range h.policy.DenyPatterns {
		if pattern.MatchString(normalizedPath) || pattern.MatchString(normalizedResolved) {
			return DecisionDeny, "Protected file: access denied for security reasons"
		}
	}

	// Check ask patterns
	for _, pattern := range h.policy.AskPatterns {
		if pattern.MatchString(normalizedPath) || pattern.MatchString(normalizedResolved) {
			return DecisionAsk, fmt.Sprintf("Critical config file: %s", filepath.Base(filePath))
		}
	}

	// For Write operations, check content for secrets
	if toolName == "Write" {
		content, ok := parsed["content"].(string)
		if ok && content != "" {
			for _, pattern := range h.policy.SensitiveContentPatterns {
				if pattern.MatchString(content) {
					return DecisionDeny, "Content contains sensitive data (credentials, API keys, or certificates)"
				}
			}
		}
	}

	return "", ""
}

// checkHarnessFrozenZoneFromInput extracts file_path from JSON tool input and delegates to checkHarnessFrozenZone.
func (h *preToolHandler) checkHarnessFrozenZoneFromInput(agentID string, toolInput json.RawMessage) (string, string) {
	if agentID == "" {
		return "", ""
	}
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return "", ""
	}
	filePath, _ := parsed["file_path"].(string)
	if filePath == "" {
		return "", ""
	}
	// Normalize to forward slashes for prefix matching.
	normalized := strings.ReplaceAll(filepath.ToSlash(filePath), "\\", "/")
	return h.checkHarnessFrozenZone(agentID, normalized)
}

// frozenZonePrefixes maps file path prefixes (relative, forward-slash) to their sentinel constant.
// Checked by checkHarnessFrozenZone for Write/Edit operations from the harness-learner agent.
// Vision §3.4: 8 frozen zone categories derived from zone-registry (W1 SSOT).
var frozenZonePrefixes = []struct {
	prefix   string
	sentinel string
}{
	{".claude/agents/moai/", SentinelHarnessFrozenAgent},
	{".claude/skills/moai-", SentinelHarnessFrozenSkill},
	{".claude/rules/moai/", SentinelHarnessFrozenRule},
	{".claude/commands/", SentinelHarnessFrozenCommand},
	{".claude/hooks/", SentinelHarnessFrozenHook},
	{".claude/output-styles/", SentinelHarnessFrozenOutputStyle},
	{".moai/project/brand/", SentinelHarnessFrozenConfig},
}

// frozenInstructionFiles lists CLAUDE.md variants guarded by HARNESS_FROZEN_INSTRUCTION_VIOLATION.
var frozenInstructionFiles = []string{"CLAUDE.md", "CLAUDE.local.md"}

// checkHarnessFrozenZone returns (sentinel, deny-reason) when the file path falls inside
// a FROZEN zone. Returns ("", "") when the path is not frozen.
// Only applies when the caller identity is harnessLearnerIdentity (harness-learner).
func (h *preToolHandler) checkHarnessFrozenZone(agentID, normalizedPath string) (string, string) {
	if agentID != harnessLearnerIdentity {
		return "", ""
	}
	// Check instruction files (basename match).
	base := filepath.Base(normalizedPath)
	for _, inst := range frozenInstructionFiles {
		if base == inst {
			reason := fmt.Sprintf("%s: harness-learner cannot modify frozen instruction file %s",
				SentinelHarnessFrozenInstruction, base)
			return SentinelHarnessFrozenInstruction, reason
		}
	}
	// Check prefix-based frozen zones.
	for _, fz := range frozenZonePrefixes {
		if strings.HasPrefix(normalizedPath, fz.prefix) {
			reason := fmt.Sprintf("%s: harness-learner cannot modify frozen path %s",
				fz.sentinel, normalizedPath)
			return fz.sentinel, reason
		}
	}
	return "", ""
}

// isAllowedExternalPath checks whether the given absolute path falls under
// one of the policy's AllowedExternalPaths directories.
func (h *preToolHandler) isAllowedExternalPath(resolvedPath string) bool {
	if h.policy == nil {
		return false
	}
	for _, allowed := range h.policy.AllowedExternalPaths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		nfcAllowed := norm.NFC.String(absAllowed)
		rel, err := filepath.Rel(nfcAllowed, resolvedPath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(rel, "..") {
			return true
		}
	}
	return false
}
