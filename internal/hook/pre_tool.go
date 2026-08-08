// Resolution: KEEP — security scan, secrets, reflective write; emit PermissionDecision.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
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
	// a frozen config area (e.g. system-level .moai/config/sections/*.yaml).
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
	cfg     ConfigProvider
	policy  *SecurityPolicy
	scanner *security.SecurityScanner
	// projectDir is the resolved project root. Tests may set it directly; the
	// production constructors leave it empty and let projectRoot resolve it.
	projectDir string
	projectRootResolver
}

// NewPreToolHandler creates a new PreToolUse event handler with the given security policy.
// projectDir is resolved on first use via resolveProjectRootFromEnv: CLAUDE_PROJECT_DIR
// env var first, then os.Getwd() fallback with slog.Warn cwd_fallback:true marker
// (REQ-HCWA-005, REQ-HCWA-008).
func NewPreToolHandler(cfg ConfigProvider, policy *SecurityPolicy) Handler {
	return &preToolHandler{
		cfg:                 cfg,
		policy:              policy,
		projectRootResolver: projectRootResolver{caller: "NewPreToolHandler"},
	}
}

// NewPreToolHandlerWithScanner creates a PreToolUse handler with AST-based security scanning.
// If scanner is nil or unavailable, falls back to pattern-based security only.
// projectDir is resolved on first use via resolveProjectRootFromEnv (REQ-HCWA-005, REQ-HCWA-008).
func NewPreToolHandlerWithScanner(cfg ConfigProvider, policy *SecurityPolicy, scanner *security.SecurityScanner) Handler {
	// Validate scanner availability
	if scanner != nil && !scanner.IsAvailable() {
		slog.Info("ast-grep not available, security scanning disabled")
		scanner = nil
	}

	return &preToolHandler{
		cfg:                 cfg,
		policy:              policy,
		scanner:             scanner,
		projectRootResolver: projectRootResolver{caller: "NewPreToolHandlerWithScanner"},
	}
}

// projectRoot returns the project root, resolving it on first use.
func (h *preToolHandler) projectRoot() string {
	return h.resolve(&h.projectDir)
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
	// No policy means allow everything (subject to the same permission-mode
	// awareness as the "no dangerous pattern found" path below — a nil
	// policy trivially finds nothing dangerous).
	if h.policy == nil {
		return NewSafeDefaultOutput(permissionModeOf(input)), nil
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

	// gateNotice carries a passing quality gate's non-blocking notice (for
	// example an ast-grep step that skipped because sg is absent) to this
	// handler's response. It rides the hook's structured output rather than
	// slog: the `moai hook` path installs a discarding handler, so a log
	// record here would be silent by construction.
	var gateNotice string

	// Handle Bash commands
	if input.ToolName == "Bash" && len(input.ToolInput) > 0 {
		command := h.extractBashCommand(input.ToolInput)

		// F5 mechanical: --no-verify bypass defense (SPEC-PRETOOL-GATE-MOVE-001
		// REQ-PGM-006). Under defaultMode: bypassPermissions, the PreToolUse
		// deny is the SOLE blocking mechanism between Claude Code and a
		// --no-verify bypass of the relocated git pre-commit gate. M1.b
		// confirmed git suppresses the pre-commit hook entirely under
		// --no-verify. Checked BEFORE the IsGitCommit gate block because a
		// compound Bash command (e.g. `git add . && git commit --no-verify`)
		// does not start with "git commit" and would bypass IsGitCommit's
		// start-anchored regex. The documented bypass is SKIP_MOAI_PRECOMMIT=1
		// (REQ-PGM-010), not --no-verify.
		//
		// Substring match on both "git commit" and "--no-verify" is sufficient:
		// --no-verify is a standalone flag that never legitimately appears inside
		// a git commit argument value (message, pathspec), and the conservative
		// over-block on contrived echo/printf payloads is safe (user can override
		// the deny). Sub-millisecond, no regex engine needed.
		if command != "" && strings.Contains(command, "--no-verify") && strings.Contains(command, "git commit") {
			reason := "git commit --no-verify would bypass the MoAI-ADK pre-commit gate " +
				"(the relocated heavy quality gate). The commit was NOT created. " +
				"Remove --no-verify, or use SKIP_MOAI_PRECOMMIT=1 for the documented bypass."
			slog.Warn("pre-tool: git commit --no-verify denied (bypass defense)")
			return NewDenyOutput(reason), nil
		}

		// Quality gate for git commit commands (REQ-GATE-001).
		// Executes before security pattern checks so the gate cannot be bypassed.
		//
		// @MX:WARN: [AUTO] tier-aware commit gate — K-6 deny-before-tier ordering invariant
		// @MX:REASON: [AUTO] SPEC-STOPCHAIN-TRIM-001 REQ-005 + REQ-007; the destructive-pattern denylist (checkBashCommand below) runs UNCONDITIONALLY after this block regardless of tier. Editing this branch to widen the tier predicate OR moving checkBashCommand after a tier-early-return would let an unattended fully-autonomous session bypass a deny — the load-bearing safety invariant of the autonomy tier.
		//
		// SPEC-STOPCHAIN-TRIM-001 REQ-005 (A11): the synchronous vet+lint+test
		// commit gate is tier-aware. At MOAI_AUTONOMY_TIER ∈ {automatic,
		// fully-autonomous} the gate is OFF — the commit proceeds without the
		// verification tax (the user approved unattended operation at
		// Implementation Kickoff). At semi-auto (or unset) the gate retains its
		// full behavior. K-6 / AP-1: the destructive-pattern denylist
		// (checkBashCommand below) runs UNCONDITIONALLY after this block
		// regardless of tier — the tier branch relaxes ONLY this verification
		// gate, never a deny. config.AutonomyTier() fails safe to semi-auto on
		// unset/empty/invalid (REQ-003 backward compat), so a session that does
		// not opt in keeps the gate ON.
		if quality.IsGitCommit(command) && !config.IsAutonomyTierCommitGateOff(config.AutonomyTier()) {
			gate := quality.NewQualityGate(h.loadGateConfig())
			passed, output := gate.Run(ctx)
			if !passed {
				slog.Warn("quality gate blocked git commit",
					"reason_summary", firstLine(output),
				)
				return NewDenyOutput(output), nil
			}
			gateNotice = output
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

	// Branch-state guard (SPEC-WORKTREE-BRANCH-GUARD-001 +
	// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001). Runs after checkBashCommand so the
	// existing dangerous-pattern deny takes precedence, and before the
	// default-allow fall-through. Gated by Workflow.BranchGuard.Enabled
	// (default false) at the call site — when disabled, NO git rev-parse
	// subprocess runs (the isPrimaryCheckout cost is avoided) and the guard is
	// inert. When enabled, denies branch-state-changing git commands ONLY in
	// the primary checkout when the invoking agent is not exempt; fails OPEN on
	// any git-context uncertainty (REQ-WBG-012). The exemption logic
	// (MOAI_BRANCH_GUARD_EXEMPT + manager-git identity) is unchanged and is
	// consulted only on the enabled path (REQ-6 backward compat).
	if input.ToolName == "Bash" && len(input.ToolInput) > 0 && h.branchGuardEnabled() {
		if decision, reason := checkBranchState(input, h.projectRoot()); decision == DecisionDeny {
			slog.Warn("branch guard denied",
				"tool_name", input.ToolName,
				"session_id", input.SessionID,
				"reason", reason,
			)
			return NewDenyOutput(reason), nil
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

		// Pre-Edit Sync Check advisory (SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001 M4, REQ-PES-004).
		// Read-only, fail-open: surfaces foreign active sessions on this checkout
		// via stderr + .moai/logs/preedit-session-guard.log WITHOUT blocking the
		// edit. The return is intentionally discarded (advisory only — never Deny).
		// This is the mechanical companion to the procedural Pre-Edit Sync Check
		// doctrine (agent-common-protocol.md § Pre-Edit Sync Check).
		checkForeignSessionAdvisory(input, h.projectRoot())

		// AST-based security scanning for Write operations
		if input.ToolName == "Write" && h.scanner != nil {
			decision, reason := h.scanWriteContent(ctx, input.ToolInput)
			if decision == DecisionDeny {
				return NewDenyOutput(reason), nil
			}
		}
	}

	// Agent/Task spawn model observation. Placed LAST — after every existing
	// deny path — so inserting it cannot displace an established decision
	// (the Bash dangerous-pattern deny, the branch guard, the file-access
	// deny all still win). The observation and advisory layers always run;
	// the deny below is opt-in via Workflow.AgentModelGuard.Enabled and fires
	// only on the mismatch verdict. Since v2.1.63 Claude Code renamed
	// Task → Agent; accept both.
	var agentAdvisory string
	if input.ToolName == "Agent" || input.ToolName == "Task" {
		decision, reason, advisory := h.checkAgentModel(input)
		if decision == DecisionDeny {
			return NewDenyOutput(reason), nil
		}
		agentAdvisory = advisory
	}

	out := NewSafeDefaultOutput(permissionModeOf(input))
	if gateNotice != "" {
		out.SystemMessage = gateNotice
	}
	if agentAdvisory != "" {
		out.SystemMessage = agentAdvisory
	}
	return out, nil
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
	result, err := h.scanner.ScanFile(ctx, tmpFile.Name(), h.projectRoot())
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

// branchGuardEnabled reports whether the Main-Checkout Branch-State Guard is
// opted in via Workflow.BranchGuard.Enabled (SPEC-WORKTREE-BRANCH-GUARD-OPTIN-
// 001 REQ-1/REQ-3). Defensive: a nil ConfigProvider or nil Config returns false
// (the distributed default — guard inert), so a misconfigured hook never
// accidentally enables the deny path. Reading the flag at the call site (NOT
// inside checkBranchState) avoids threading a config arg through the signature
// and keeps the isPrimaryCheckout subprocess cost entirely off the disabled
// path (plan.md D1).
func (h *preToolHandler) branchGuardEnabled() bool {
	if h.cfg == nil {
		return false
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return false
	}
	return cfg.Workflow.BranchGuard.Enabled
}

// loadGateConfig reads gate configuration from the config provider.
// Falls back to DefaultGateConfig when the config is not available.
func (h *preToolHandler) loadGateConfig() *quality.GateConfig {
	var qcfg *quality.GateConfig

	if h.cfg == nil {
		qcfg = quality.DefaultGateConfig()
	} else {
		cfg := h.cfg.Get()
		if cfg == nil {
			qcfg = quality.DefaultGateConfig()
		} else {
			gate := cfg.Gate
			qcfg = &quality.GateConfig{
				Enabled:     gate.Enabled,
				SkipTests:   gate.SkipTests,
				VetTimeout:  gate.VetTimeoutDuration(),
				LintTimeout: gate.LintTimeoutDuration(),
				TestTimeout: gate.TestTimeoutDuration(),
			}
			// Map gate.disabled_steps through verbatim (issue #1265): the
			// runner reads a FALSE value as "skip this step", so normalising
			// the values here would silently stop the skip from applying.
			if len(gate.DisabledSteps) > 0 {
				qcfg.DisabledSteps = make(map[string]bool, len(gate.DisabledSteps))
				maps.Copy(qcfg.DisabledSteps, gate.DisabledSteps)
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
		}
	}

	// Project context + Go build tags apply uniformly across every config path
	// (including the DefaultGateConfig fallback when no config provider is
	// loaded), so a project requiring non-default build tags (e.g. goolm) is
	// vetted/linted/tested under them regardless of config availability.
	qcfg.ProjectDir = h.projectRoot()
	qcfg.GoBuildTags = readGoBuildTags(qcfg.ProjectDir)
	return qcfg
}

// readGoBuildTags reads the project Go build tags from
// <dir>/.moai/config/build-tags (first non-empty, non-comment line) when
// present. Returns "" when the file is absent or unreadable so the gate
// behaves as before for projects that do not declare build tags. The value
// is passed verbatim to go vet/test (-tags=<value>) and golangci-lint
// (--tags=<value>); Go accepts comma- or space-separated tags.
func readGoBuildTags(dir string) string {
	if dir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, ".moai", "config", "build-tags"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
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

	// SPEC-INTERNAL-SECURITY-001 REQ-SEC-007: resolve symlinks before the
	// project-boundary check and DenyPatterns matching. A project-internal
	// symlink (e.g. notes.txt -> ~/.ssh/id_rsa) passes the lexical boundary
	// check and matches no deny pattern on its unresolved form, but the Write
	// tool then follows the link and overwrites the real secret (CWE-61).
	// EvalSymlinks resolves the real target so both downstream checks see the
	// actual destination. This mirrors the file_changed.go resolve-recheck
	// pattern (SPEC-SEC-HARDEN-004 / REQ-SEC4-004).
	//
	// Unlike file_changed.go (which fails-closed on EvalSymlinks error), this
	// is the CRITICAL path: a Write to a not-yet-existing path (new file) MUST
	// still succeed. When EvalSymlinks returns an error (not-exist for a
	// new-file Write, or otherwise unresolvable), fall back to the unresolved
	// path and do NOT deny (NFR-SEC-003 behavior preservation, AC-SEC-007c).
	resolvedSymlink := false
	if realPath, evalErr := filepath.EvalSymlinks(resolvedPath); evalErr == nil {
		resolvedPath = realPath
		resolvedSymlink = true
	}

	// Check if path is within project directory
	if projectDir := h.projectRoot(); projectDir != "" {
		projectAbs, absErr := filepath.Abs(projectDir)
		if absErr != nil {
			// Cannot resolve project directory, skip boundary check
			slog.Debug("cannot resolve project directory", "error", absErr)
		} else {
			// Normalize projectAbs via EvalSymlinks ONLY when resolvedPath was
			// also resolved, so the boundary comparison stays symmetric. If
			// resolvedPath fell back to its unresolved form (new-file Write),
			// keep projectAbs unresolved too — otherwise a symlinked project
			// prefix (macOS /var -> /private/var) would make a legitimate
			// in-project new-file look like an escape (false-positive deny,
			// NFR-SEC-003 violation). Mirrors file_changed.go's normRoot.
			if resolvedSymlink {
				if resolvedProject, evalErr := filepath.EvalSymlinks(projectAbs); evalErr == nil {
					projectAbs = resolvedProject
				}
			}
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

	// Check content for sensitive data. Write carries the payload in "content";
	// Edit carries the replacement in "new_string" (REQ-SEC-008). Both must be
	// scanned so a secret injected via Edit cannot bypass the deny that the
	// identical Write content would trigger.
	if toolName == "Write" || toolName == "Edit" {
		contentField := "content"
		if toolName == "Edit" {
			contentField = "new_string"
		}
		content, ok := parsed[contentField].(string)
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

// isAllowedExternalPath checks whether the given absolute path falls under one
// of the policy's built-in AllowedExternalPaths directories, OR under a
// directory the user registered via Claude Code's permissions.additionalDirectories
// (.claude/settings.json / settings.local.json). The latter is read at call time
// so /add-dir takes effect without restarting the session (issue #1162).
func (h *preToolHandler) isAllowedExternalPath(resolvedPath string) bool {
	for _, allowed := range h.allowedExternalPaths() {
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

// allowedExternalPaths returns the policy's built-in allowlist plus any
// directories the user registered in Claude Code settings under
// permissions.additionalDirectories (issue #1162). Both .claude/settings.json
// and .claude/settings.local.json are consulted, relative to the project dir,
// so /add-dir (which writes there) is honored by the path guard.
func (h *preToolHandler) allowedExternalPaths() []string {
	var paths []string
	if h.policy != nil {
		paths = append(paths, h.policy.AllowedExternalPaths...)
	}
	paths = append(paths, loadAdditionalDirectories(h.projectRoot())...)
	return paths
}

// loadAdditionalDirectories reads permissions.additionalDirectories from the
// project's Claude Code settings (.claude/settings.json and
// .claude/settings.local.json) and returns the absolute directory paths.
// Missing files, malformed JSON, or an absent key yield an empty result — the
// hook never blocks on a settings read failure. Relative entries are resolved
// against the project directory.
func loadAdditionalDirectories(projectDir string) []string {
	if projectDir == "" {
		return nil
	}
	var dirs []string
	for _, name := range []string{"settings.json", "settings.local.json"} {
		path := filepath.Join(projectDir, ".claude", name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Permissions struct {
				AdditionalDirectories []string `json:"additionalDirectories"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			continue
		}
		for _, d := range doc.Permissions.AdditionalDirectories {
			if d == "" {
				continue
			}
			if !filepath.IsAbs(d) {
				d = filepath.Join(projectDir, d)
			}
			if abs, err := filepath.Abs(d); err == nil {
				dirs = append(dirs, abs)
			}
		}
	}
	return dirs
}
