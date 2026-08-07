// Package cli — `moai mcp-server` subcommand (SPEC-MOAI-MCP-SERVER-001 M1).
//
// mcp_server.go implements a thin stdio JSON-RPC MCP server over the SAME
// internal/ core the CLI Cobra handlers call (same-core-two-surfaces, C1). Every
// tool handler is a thin wrapper: it maps JSON-RPC params → the internal/
// function's args, calls that function, and shapes the result into the tool's
// declared JSON Schema. No core logic is forked into the MCP layer (AP-1).
//
// The SDK (mark3labs/mcp-go) is the only thing that touches the transport; the
// internal/ core is untouched by design, so swapping the SDK re-touches only
// this handler-shape layer (design.md §5, R3; progress.md §G.1/G.2).
//
// @MX:ANCHOR: [AUTO] moai mcp-server — thin stdio JSON-RPC wrapper over internal/
// @MX:REASON: [AUTO] the single MCP entry point; every core tool handler fans into a verified internal/ fn (session/spec/goal/verify/runtime). Same-core-two-surfaces invariant C1.
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/runtime"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/modu-ai/moai-adk/internal/verify"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// moai-mcp-server identity + .mcp.json entry constants. These stay generic by
// construction (REQ-MCP-017 / §14 hardcoding prevention): PATH-resolved command,
// fixed subcommand, no absolute path / model name / org identifier.
const (
	// moaiMCPServerName is the server name advertised in `initialize` (ClientInfo
	// server name). Generic — matches the distributed binary name.
	moaiMCPServerName = "moai"
	// moaiMCPServerKey is the mcpServers key under which the entry is provisioned
	// in .mcp.json (design.md §6.2 locked neutral entry).
	moaiMCPServerKey = "moai"
	// moaiMCPServerCommand is the PATH-resolved command (NO absolute path).
	moaiMCPServerCommand = "moai"
	// moaiMCPServerSubcommand is the single fixed arg provisioning the server.
	moaiMCPServerSubcommand = "mcp-server"
)

// newMCPServerCmd builds the `moai mcp-server` subcommand: a long-running
// stdio JSON-RPC server modeled on the goal.go blocking-RunE pattern. It is
// registered under root.go AddCommand alongside the other subcommands.
//
// IMPORTANT: this CLI surface MUST NOT invoke AskUserQuestion (subagent
// boundary, C-HRA-008 / REQ-MCP-014). MCP tool handlers return structured
// results only; the orchestrator owns all user interaction.
func newMCPServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   moaiMCPServerSubcommand,
		Short: "Run the moai self-hosted MCP server (stdio JSON-RPC)",
		Long: `Run a long-running stdio JSON-RPC MCP server exposing moai's core
read/status/audit tools to MCP-capable hosts (Cursor, Cline, Zed, another LLM).

The server is a THIN WRAPPER over the same internal/ core the moai CLI uses:
every tool handler calls an existing internal/ function, so the CLI and MCP
surfaces share one source of truth and can never diverge.

Provisioning the server into a project's .mcp.json is opt-in (default off);
this subcommand only RUNS the server. Use ` + "`moai init`" + ` / ` + "`moai web`" + ` to
provision the entry (M4).`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runMCPServer()
		},
	}
}

// runMCPServer builds the moai MCP server and serves it over stdio (blocking,
// until the stdio stream closes). REQ-MCP-001. ServeStdio owns its context
// internally (derived from os signals); there is no ctx to thread here.
func runMCPServer() error {
	s := newMoaiMCPServer()
	// ServeStdio blocks until the stdin stream closes; the goal.go blocking-RunE
	// pattern. opts remain extensible (error logger, etc.) without API churn.
	return server.ServeStdio(s)
}

// newMoaiMCPServer constructs the MCP server instance, advertises the moai
// identity + version, and registers the core tool handlers. REQ-MCP-001/004/005.
func newMoaiMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		moaiMCPServerName,
		version.GetVersion(),
		server.WithToolCapabilities(true),
	)
	registerMoaiMCPTools(s)
	return s
}

// registerMoaiMCPTools registers the M1 core read/status/audit tool surface
// (design.md §3). Each tool declares its JSON Schema (REQ-MCP-004) and registers
// a thin-wrapper handler that calls the verified internal/ integration point
// (REQ-MCP-005). Read-only tools carry the read-only hint annotation (§G.2).
func registerMoaiMCPTools(s *server.MCPServer) {
	// --- session_list → session.QueryActiveWork (registry.go:254) ---
	s.AddTool(mcp.NewTool(
		"session_list",
		mcp.WithDescription("List active moai sessions (optional spec_id filter). Wraps session.QueryActiveWork."),
		mcp.WithString("spec_id", mcp.Description("Optional SPEC-ID filter; empty lists all active sessions.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSessionList)

	// --- goal_status → goal.LoadGoal (state.go:57) ---
	s.AddTool(mcp.NewTool(
		"goal_status",
		mcp.WithDescription("Read the armed-goal state for a session. Wraps goal.LoadGoal."),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("Session id whose goal state is read.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGoalStatus)

	// --- goal_arm → goal.NewGoal (schema.go:115) + goal.SaveGoal (state.go:76) ---
	s.AddTool(mcp.NewTool(
		"goal_arm",
		mcp.WithDescription("Arm a condition-declared goal for a session. Wraps the goal.NewGoal + goal.SaveGoal composition (same internal/goal core the `moai goal arm` CLI uses)."),
		mcp.WithString("condition", mcp.Required(), mcp.Description("Goal condition text (a shell command optionally suffixed 'exits <N>', or a transcript-referencing model claim).")),
		mcp.WithString("session_id", mcp.Description("Session id to arm under; defaults to the resolved current session.")),
		mcp.WithInteger("max_turns", mcp.Description("Turn ceiling (0 = infinite, requires max_duration > 0). Default 30 when omitted.")),
		mcp.WithInteger("max_duration", mcp.Description("Wall-clock bound in seconds since arm time (required when max_turns = 0).")),
	), handleGoalArm)

	// --- spec_progress → spec.ListDocs (listdocs.go:36) — the SPEC scanner ---
	s.AddTool(mcp.NewTool(
		"spec_progress",
		mcp.WithDescription("List SPEC documents + frontmatter under a project root (SPEC lifecycle distribution source). Wraps spec.ListDocs."),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecProgress)

	// --- verify_snapshot → verify.Load (store.go:38) + verify.RecordCheck (store.go:107) ---
	s.AddTool(mcp.NewTool(
		"verify_snapshot",
		mcp.WithDescription("Read (or record into) the per-key verification snapshot. Wraps verify.Load (+ verify.RecordCheck when a check is supplied). First CLI/MCP surface for verify."),
		mcp.WithString("key", mcp.Required(), mcp.Description("Snapshot key (HEAD:digest form).")),
		mcp.WithString("command", mcp.Description("When set, RECORD a check entry via verify.RecordCheck instead of reading.")),
		mcp.WithInteger("exit_code", mcp.Description("Exit code for a recorded check (used with 'command').")),
	), handleVerifySnapshot)

	// --- verify_trend → verify.Load (store.go:38) — the per-key check history ---
	s.AddTool(mcp.NewTool(
		"verify_trend",
		mcp.WithDescription("Read the per-key verification check history (trend). Wraps verify.Load, surfacing the Checks sequence."),
		mcp.WithString("key", mcp.Required(), mcp.Description("Snapshot key whose check trend is read.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleVerifyTrend)

	// --- spec_audit → spec.Audit (audit.go:156) ---
	s.AddTool(mcp.NewTool(
		"spec_audit",
		mcp.WithDescription("Run the SPEC lifecycle audit (era classification + drift detection). Wraps spec.Audit."),
		mcp.WithString("filter_spec", mcp.Description("Optional SPEC-ID filter (exact match).")),
		mcp.WithString("filter_era", mcp.Description("Optional era filter (e.g. V3R6).")),
		mcp.WithBoolean("include_grandfathered", mcp.Description("Surface grandfathered (V2.x/V3R2-R4/V3R5) SPECs as INFO.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecAudit)

	// --- spec_drift → spec.Audit (audit.go:156), modern-era drift view ---
	s.AddTool(mcp.NewTool(
		"spec_drift",
		mcp.WithDescription("Read SPEC lifecycle drift (modern-era V3R6 drift findings only). Wraps spec.Audit, filtered to drift."),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecDrift)

	// --- audit_cache → runtime.AuditCache (audit_cache.go:50) ---
	s.AddTool(mcp.NewTool(
		"audit_cache",
		mcp.WithDescription("Operate the plan-audit PASS cache (read-only reuse): compute_hash a SPEC dir, or lookup a cached verdict. Wraps runtime.AuditCache ComputeHash/Lookup."),
		mcp.WithString("op", mcp.Required(), mcp.Description("Operation: 'compute_hash' or 'lookup'.")),
		mcp.WithString("spec_dir", mcp.Description("SPEC directory (for op=compute_hash).")),
		mcp.WithString("spec_id", mcp.Description("SPEC id (for op=lookup).")),
		mcp.WithString("hash", mcp.Description("Plan-artifact hash (for op=lookup).")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleAuditCache)

	// --- M2 codex backend (design.md §3 M2) ---

	// codex_audit → codex binary shellout (REQ-MCP-006 / AC-MCP-007). mode=native
	// dispatches codex review/start; mode=adversarial dispatches turn/start + the
	// adversarial-review prompt. Output adopts review-output.schema.json (§G.4).
	// Fail-open on missing codex (VerdictInconclusive). The codex binary is
	// OPTIONAL + experimental (R1).
	s.AddTool(mcp.NewTool(
		"codex_audit",
		mcp.WithDescription("Run a codex code review. mode=native → codex review/start; mode=adversarial → codex turn/start + an adversarial-review prompt. Returns a review-output schema (verdict/summary/findings/next_steps). codex is OPTIONAL; a missing or unavailable codex yields verdict 'inconclusive' (fail-open)."),
		mcp.WithString("mode", mcp.Enum(codexModeNative, codexModeAdversarial), mcp.Description("Audit mode: 'native' (codex review/start) or 'adversarial' (codex turn/start + red-team prompt). Defaults to native.")),
		mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), mcp.Description("What codex reviews: 'uncommittedChanges' or 'baseBranch'.")),
		mcp.WithString("focus", mcp.Description("Adversarial-only focus area (e.g. 'concurrency', 'auth').")),
		mcp.WithString("model", mcp.Description("Optional model override (resolved via the model/effort SSOT in M3).")),
		mcp.WithOutputSchema[ReviewOutput](),
	), handleCodexAudit)

	// codex_setup → Go probe (REQ-MCP-007 / AC-MCP-008): exec.LookPath("codex")
	// + codex --version + auth-provider classification + the enable_review_gate
	// toggle. NO Node bridge. Read-only.
	s.AddTool(mcp.NewTool(
		"codex_setup",
		mcp.WithDescription("Probe the local codex installation: exec.LookPath + codex --version + auth-provider classification (ChatGPT / apiKey / provider / unknown) + the current enable_review_gate toggle state. Pure Go — no Node bridge."),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexSetup)

	// --- M3 GLM backend (design.md §3 M3) ---

	// glm_audit → z.ai GLM API direct call (REQ-MCP-009 / AC-MCP-011). Calls
	// the Anthropic-compatible https://api.z.ai/api/anthropic/v1/messages
	// endpoint directly — NOT the z.ai MCP server, NOT any gateway. Output
	// adopts the shared review-output.schema.json (§G.4). Fail-open on a
	// missing/unauthenticated/erroring/malformed GLM (VerdictInconclusive →
	// claude fallback, REQ-MCP-012). Model/effort resolved via the SSOT
	// (ResolveAgentModelEffort, REQ-MCP-013) when no explicit model is given.
	s.AddTool(mcp.NewTool(
		"glm_audit",
		mcp.WithDescription("Run a GLM (z.ai) code review. Calls the z.ai Anthropic-compatible endpoint directly and returns a review-output schema (verdict/summary/findings/next_steps). GLM is OPTIONAL; a missing/unauthenticated/erroring GLM yields verdict 'inconclusive' (fail-open → fall back to the active auditor)."),
		mcp.WithString("focus", mcp.Description("Optional focus area (e.g. 'concurrency', 'auth', 'secret handling').")),
		mcp.WithString("model", mcp.Description("Optional GLM model override; omitted ⇒ resolved via the model/effort SSOT (ResolveAgentModelEffort).")),
		mcp.WithOutputSchema[ReviewOutput](),
	), handleGLMAudit)
}

// ─── thin-wrapper handlers (C1: each calls the SAME internal/ fn the CLI calls) ───

// handleSessionList wraps session.QueryActiveWork (registry.go:254).
func handleSessionList(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	specID := req.GetString("spec_id", "")
	entries, err := session.QueryActiveWork(specID)
	if err != nil {
		return toolErr("session_list", err), nil
	}
	return toolJSON("session_list", map[string]any{
		"spec_id":  specID,
		"count":    len(entries),
		"sessions": entries,
	}), nil
}

// handleGoalStatus wraps goal.LoadGoal (state.go:57).
func handleGoalStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := req.GetString("session_id", "")
	if sessionID == "" {
		sessionID = statusSessionID("") // inherit the CLI's read-path resolver
	}
	g, err := goal.LoadGoal(resolveProjectDir(), sessionID)
	if err != nil {
		return toolErr("goal_status", err), nil
	}
	if g == nil {
		return toolJSON("goal_status", map[string]any{"session_id": sessionID, "armed": false}), nil
	}
	return toolJSON("goal_status", map[string]any{"session_id": sessionID, "armed": true, "goal": g}), nil
}

// handleGoalArm wraps the goal.NewGoal (schema.go:115) + goal.SaveGoal
// (state.go:76) composition — the SAME internal/goal core the `moai goal arm`
// CLI RunE composes. It inherits the CLI's core rule enforcement
// (parseCondition classification; the infinite-goal fail-closed: max_turns == 0
// requires max_duration > 0) WITHOUT re-implementing internal/goal logic. The
// CLI-only stderr warnings / flag UX are presentation, not core.
func handleGoalArm(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root := resolveProjectDir()
	if root == "" {
		return toolErr("goal_arm", fmt.Errorf("cannot resolve project root (set CLAUDE_PROJECT_DIR or run from the project directory)")), nil
	}
	conditionText := req.GetString("condition", "")
	if conditionText == "" {
		return toolErr("goal_arm", fmt.Errorf("condition must not be empty")), nil
	}
	maxTurns := req.GetInt("max_turns", -1) // -1 sentinel = omitted
	maxDuration := req.GetInt("max_duration", 0)

	// Inherited rule (same predicate the CLI applies): an infinite goal
	// (max_turns == 0) REQUIRES a real wall-clock bound. Recorded-only cost-cap
	// does not satisfy it. The CLI surfaces this as stderr text + non-zero exit;
	// the MCP tool surfaces it as a structured IsError result.
	if maxTurns == 0 && maxDuration <= 0 {
		return toolErr("goal_arm", fmt.Errorf("max_turns 0 (infinite) requires max_duration > 0 as the real wall-clock bound")), nil
	}

	sessionID := req.GetString("session_id", "")
	if sessionID == "" {
		sessionID, _ = resolveArmSessionID("") // inherit the CLI arm-path resolver
	}

	cond := parseCondition(conditionText) // same classifier the CLI uses
	g := goal.NewGoal(sessionID, conditionText, []goal.Condition{cond})
	if maxTurns >= 0 {
		g.Ceiling.MaxTurns = maxTurns
	}
	g.Ceiling.MaxDuration = maxDuration
	if err := goal.SaveGoal(root, g); err != nil {
		return toolErr("goal_arm", err), nil
	}
	return toolJSON("goal_arm", map[string]any{
		"action":           "arm",
		"session_id":       sessionID,
		"condition_type":   string(cond.Type),
		"ceiling_maxturns": g.Ceiling.MaxTurns,
	}), nil
}

// handleSpecProgress wraps spec.ListDocs (listdocs.go:36) — the SPEC scanner.
func handleSpecProgress(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	records, err := spec.ListDocs(resolveProjectDir())
	if err != nil {
		return toolErr("spec_progress", err), nil
	}
	return toolJSON("spec_progress", map[string]any{
		"count": len(records),
		"specs": records,
	}), nil
}

// handleVerifySnapshot wraps verify.Load (store.go:38); when a `command` arg is
// supplied it records via verify.RecordCheck (store.go:107).
func handleVerifySnapshot(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root := resolveProjectDir()
	key := req.GetString("key", "")
	if key == "" {
		return toolErr("verify_snapshot", fmt.Errorf("key must not be empty")), nil
	}
	if cmd := req.GetString("command", ""); cmd != "" {
		entry := verify.CheckEntry{
			CheckID:  cmd,
			Command:  cmd,
			ExitCode: req.GetInt("exit_code", 0),
		}
		snap, err := verify.RecordCheck(root, key, entry)
		if err != nil {
			return toolErr("verify_snapshot", err), nil
		}
		return toolJSON("verify_snapshot", map[string]any{"action": "record", "snapshot": snap}), nil
	}
	snap, err := verify.Load(root, key)
	if err != nil {
		return toolErr("verify_snapshot", err), nil
	}
	return toolJSON("verify_snapshot", map[string]any{"action": "load", "snapshot": snap}), nil
}

// handleVerifyTrend wraps verify.Load (store.go:38), surfacing the Checks trend.
func handleVerifyTrend(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key := req.GetString("key", "")
	if key == "" {
		return toolErr("verify_trend", fmt.Errorf("key must not be empty")), nil
	}
	snap, err := verify.Load(resolveProjectDir(), key)
	if err != nil {
		return toolErr("verify_trend", err), nil
	}
	checks := []verify.CheckEntry{}
	if snap != nil {
		checks = snap.Checks
	}
	return toolJSON("verify_trend", map[string]any{"key": key, "checks": checks}), nil
}

// handleSpecAudit wraps spec.Audit (audit.go:156).
func handleSpecAudit(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	opts := spec.AuditOptions{
		BaseDir:              resolveProjectDir(),
		FilterSpec:           req.GetString("filter_spec", ""),
		FilterEra:            req.GetString("filter_era", ""),
		IncludeGrandfathered: req.GetBool("include_grandfathered", false),
	}
	result, err := spec.Audit(opts)
	if err != nil {
		return toolErr("spec_audit", err), nil
	}
	return toolJSON("spec_audit", result), nil
}

// handleSpecDrift wraps spec.Audit (audit.go:156), filtered to modern-era drift.
func handleSpecDrift(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := spec.Audit(spec.AuditOptions{BaseDir: resolveProjectDir()})
	if err != nil {
		return toolErr("spec_drift", err), nil
	}
	return toolJSON("spec_drift", map[string]any{
		"total_specs":    result.TotalSpecs,
		"modern_clean":   result.ModernEraClean,
		"drift_findings": result.DriftFindings,
	}), nil
}

// handleAuditCache wraps runtime.AuditCache ComputeHash/Lookup (audit_cache.go).
// The cache is in-memory, per server process (read-only reuse — design.md §3).
func handleAuditCache(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cache := runtime.NewInMemoryCache()
	op := req.GetString("op", "")
	switch op {
	case "compute_hash":
		specDir := req.GetString("spec_dir", "")
		if specDir == "" {
			return toolErr("audit_cache", fmt.Errorf("op=compute_hash requires spec_dir")), nil
		}
		hash, err := cache.ComputeHash(specDir)
		if err != nil {
			return toolErr("audit_cache", err), nil
		}
		return toolJSON("audit_cache", map[string]any{"op": op, "hash": hash}), nil
	case "lookup":
		specID := req.GetString("spec_id", "")
		hash := req.GetString("hash", "")
		entry, ok := cache.Lookup(specID, hash, time.Now())
		return toolJSON("audit_cache", map[string]any{"op": op, "hit": ok, "entry": entry}), nil
	default:
		return toolErr("audit_cache", fmt.Errorf("unknown op %q (want compute_hash|lookup)", op)), nil
	}
}

// ─── .mcp.json provisioning (AC-MCP-006, opt-in default-off C6) ───────────────

// buildMoaiMCPServerEntry returns the neutral .mcp.json server entry for the
// moai self-hosted MCP server (design.md §6.2 locked shape, §G.5). It is the
// sibling of buildZAIMCPEntry (glm_tools.go:391) but carries NO env block — the
// moai entry invokes the user's own binary with a fixed subcommand; secrets
// (only when a backend is enabled, M3) stay ${VAR} literals, never resolved
// values (C3 / REQ-MCP-011).
func buildMoaiMCPServerEntry() map[string]any {
	return map[string]any{
		"command": moaiMCPServerCommand,
		"args":    []string{moaiMCPServerSubcommand},
	}
}

// provisionMoaiMCPServerEntryAt is the path-explicit provisioning seam (testable
// without scope resolution). It inserts exactly one neutral moai entry under
// root["mcpServers"][moaiMCPServerKey] via the guarded atomic RMW
// (mutateClaudeJSONAtomic — the SAME flock+RMW infrastructure the zai MCP
// provisioning uses). Opt-in default-off (C6): never called unless the user
// explicitly provisions.
//
// M1 ships ONLY this path-explicit seam. The scope-resolving wrapper
// (resolveConfigPath(scope) → this seam) is DEFERRED TO M4, where it will be
// wired into the `moai init` / `moai web` call sites alongside the Template-First
// reversal (plan.md §F M4, design.md §6.3). It is intentionally absent here so
// M1 carries no dead/uncalled code (scope discipline, Agent Core Behavior #5).
func provisionMoaiMCPServerEntryAt(configPath string) error {
	return mutateClaudeJSONAtomic(configPath, func(root map[string]any) (bool, error) {
		servers, _ := root["mcpServers"].(map[string]any)
		if servers == nil {
			servers = map[string]any{}
		}
		// Idempotent: if an identical entry is already present, skip the write.
		// Comparison is marshal-based so it survives the JSON round-trip
		// ([]string from buildMoaiMCPServerEntry vs []any after a disk
		// read/unmarshal — a type-assertion on []string alone would miss the
		// re-read case and needlessly rewrite every time).
		if existing, ok := servers[moaiMCPServerKey].(map[string]any); ok && mcpEntryEqual(existing, buildMoaiMCPServerEntry()) {
			root["mcpServers"] = servers
			return false, nil
		}
		servers[moaiMCPServerKey] = buildMoaiMCPServerEntry()
		root["mcpServers"] = servers
		return true, nil
	})
}

// ─── result-shaping helpers (structured results, never AskUserQuestion) ───────

// toolJSON shapes a successful structured result with a text fallback.
func toolJSON(tool string, data any) *mcp.CallToolResult {
	return mcp.NewToolResultStructured(data, tool+": ok")
}

// toolErr shapes a domain error as a structured IsError result (NOT a Go error),
// preserving the subagent boundary: the tool call succeeds at the JSON-RPC layer
// while surfacing the failure to the caller (REQ-MCP-014).
func toolErr(tool string, err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: tool + ": " + err.Error()},
		},
	}
}
