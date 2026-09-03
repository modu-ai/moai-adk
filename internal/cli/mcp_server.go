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
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/goal"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
	"github.com/modu-ai/moai-adk/internal/runtime"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/sessionmsg"
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
	// Stamp this process's build identity so `moai doctor` can detect a host
	// still talking to a previously-installed build (mcp_server_runtime.go).
	// Best-effort by contract: a failed stamp never blocks serving.
	if recordPath, err := writeMCPServerRuntimeRecord(resolveProjectDir()); err == nil {
		defer removeMCPServerRuntimeRecord(recordPath)
	}
	s := newMoaiMCPServer()
	// ServeStdio blocks until the stdin stream closes; the goal.go blocking-RunE
	// pattern. opts remain extensible (error logger, etc.) without API churn.
	return server.ServeStdio(s)
}

// newMoaiMCPServer constructs the MCP server instance, advertises the moai
// identity + version, and registers the core tool handlers. REQ-MCP-001/004/005.
//
// The per-tool enablement map is read ONCE at construction (SPEC-MCP-CONSOLE-001
// REQ-C-2) from `.moai/config/sections/mcp.yaml` (default all-enabled) and a
// disabled tool is not registered at all, so it never appears in tools/list.
func newMoaiMCPServer() *server.MCPServer {
	// The advertised version carries the FULL build stamp (semver + commit +
	// build date), not the bare semver. A host that reconnects after a
	// reinstall shows the same semver either way, so the commit is the only
	// field that makes a version skew visible in the `initialize` result.
	s := server.NewMCPServer(
		moaiMCPServerName,
		version.GetFullVersion(),
		server.WithToolCapabilities(true),
		server.WithInstructions(moaiMCPServerInstructions()),
	)
	registerMoaiMCPTools(s, resolveProjectDir())
	return s
}

// moaiMCPServerInstructions returns the server `instructions` string sent in
// the `initialize` result. It names the running build so an operator reading
// the host's MCP panel can compare it against `moai version` without calling
// a tool, and states the restart requirement explicitly — a rebuilt binary
// does not replace an already-spawned server process.
func moaiMCPServerInstructions() string {
	return fmt.Sprintf(
		"moai self-hosted MCP server, build %s (pid %d). "+
			"This process was spawned by the host and keeps running the build it started with: "+
			"after reinstalling the moai binary, reconnect the server so the host respawns it, "+
			"otherwise newly added tools stay absent from tools/list. "+
			"Run `moai doctor` to compare the running server against the installed binary.",
		version.GetFullVersion(), os.Getpid(),
	)
}

// registerMoaiMCPTools registers the M1 core read/status/audit tool surface
// (design.md §3). Each tool declares its JSON Schema (REQ-MCP-004) and registers
// a thin-wrapper handler that calls the verified internal/ integration point
// (REQ-MCP-005). Read-only tools carry the read-only hint annotation (§G.2).
func registerMoaiMCPTools(s *server.MCPServer, projectDir string) {
	// SPEC-MCP-CONSOLE-001 M1 (REQ-C-2): per-tool enablement consulted at
	// registration time. A disabled tool is NOT registered and does NOT appear
	// in tools/list. Default all-enabled (owner decision 2026-08-12) — a
	// missing/unreadable/absent mcp.yaml registers every tool.
	enabled := readMCPToolEnablement(projectDir)
	// add gates one tool registration on the per-tool enablement map. The name
	// argument MUST match both the mcp.NewTool name and the catalog
	// (internal/mcp MoaiMCPTools) — the guard test
	// TestMoaiMCPServer_RegistrationMatchesCatalog enforces set equality.
	add := func(name string, tool mcp.Tool, handler server.ToolHandlerFunc) {
		if enabled[name] {
			s.AddTool(tool, handler)
		}
	}

	// --- session_list → session.QueryActiveWork (registry.go:254) ---
	add("session_list", mcp.NewTool(
		"session_list",
		mcp.WithDescription("List active moai sessions (optional spec_id filter). Wraps session.QueryActiveWork."),
		mcp.WithString("spec_id", mcp.Description("Optional SPEC-ID filter; empty lists all active sessions.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSessionList)

	// --- goal_status → goal.LoadGoal (state.go:57) ---
	add("goal_status", mcp.NewTool(
		"goal_status",
		mcp.WithDescription("Read the armed-goal state for a session. Wraps goal.LoadGoal."),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("Session id whose goal state is read.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGoalStatus)

	// --- goal_arm → goal.NewGoal (schema.go:115) + goal.SaveGoal (state.go:76) ---
	add("goal_arm", mcp.NewTool(
		"goal_arm",
		mcp.WithDescription("Arm a condition-declared goal for a session. Wraps the goal.NewGoal + goal.SaveGoal composition (same internal/goal core the `moai goal arm` CLI uses)."),
		mcp.WithString("condition", mcp.Required(), mcp.Description("Goal condition text (a shell command optionally suffixed 'exits <N>', or a transcript-referencing model claim).")),
		mcp.WithString("session_id", mcp.Description("Session id to arm under; defaults to the resolved current session.")),
		mcp.WithInteger("max_turns", mcp.Description("Turn ceiling (0 = infinite, requires max_duration > 0). Default 30 when omitted.")),
		mcp.WithInteger("max_duration", mcp.Description("Wall-clock bound in seconds since arm time (required when max_turns = 0).")),
	), handleGoalArm)

	// --- spec_progress → spec.ListDocs (listdocs.go:36) — the SPEC scanner ---
	add("spec_progress", mcp.NewTool(
		"spec_progress",
		mcp.WithDescription("List SPEC documents + frontmatter under a project root (SPEC lifecycle distribution source). Wraps spec.ListDocs."),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecProgress)

	// --- verify_snapshot → verify.Load (store.go:38) + verify.RecordCheck (store.go:107) ---
	add("verify_snapshot", mcp.NewTool(
		"verify_snapshot",
		mcp.WithDescription("Read (or record into) the per-key verification snapshot. Wraps verify.Load (+ verify.RecordCheck when a check is supplied). First CLI/MCP surface for verify."),
		projectRootOption(),
		mcp.WithString("key", mcp.Required(), mcp.Description("Snapshot key (HEAD:digest form).")),
		mcp.WithString("command", mcp.Description("When set, RECORD a check entry via verify.RecordCheck instead of reading.")),
		mcp.WithInteger("exit_code", mcp.Description("Exit code for a recorded check (used with 'command').")),
	), handleVerifySnapshot)

	// --- verify_trend → verify.Load (store.go:38) — the per-key check history ---
	add("verify_trend", mcp.NewTool(
		"verify_trend",
		mcp.WithDescription("Read the per-key verification check history (trend). Wraps verify.Load, surfacing the Checks sequence."),
		projectRootOption(),
		mcp.WithString("key", mcp.Required(), mcp.Description("Snapshot key whose check trend is read.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleVerifyTrend)

	// --- spec_audit → spec.Audit (audit.go:156) ---
	add("spec_audit", mcp.NewTool(
		"spec_audit",
		mcp.WithDescription("Run the SPEC lifecycle audit (era classification + drift detection). Wraps spec.Audit."),
		mcp.WithString("filter_spec", mcp.Description("Optional SPEC-ID filter (exact match).")),
		mcp.WithString("filter_era", mcp.Description("Optional era filter (e.g. V3R6).")),
		mcp.WithBoolean("include_grandfathered", mcp.Description("Surface grandfathered (V2.x/V3R2-R4/V3R5) SPECs as INFO.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecAudit)

	// --- spec_drift → spec.Audit (audit.go:156), modern-era drift view ---
	add("spec_drift", mcp.NewTool(
		"spec_drift",
		mcp.WithDescription("Read SPEC lifecycle drift (modern-era V3R6 drift findings only). Wraps spec.Audit, filtered to drift."),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecDrift)

	// --- audit_cache → runtime.AuditCache (audit_cache.go:50) ---
	add("audit_cache", mcp.NewTool(
		"audit_cache",
		mcp.WithDescription("Operate the plan-audit PASS cache (process-shared): compute_hash a SPEC dir, lookup a cached verdict, or store a PASS verdict. Wraps runtime.AuditCache ComputeHash/Lookup/Store. NOTE: cache is in-memory per moai mcp-server process; the plan-audit gate (separate process) does not populate it."),
		mcp.WithString("op", mcp.Required(), mcp.Description("Operation: 'compute_hash', 'lookup', or 'store'.")),
		mcp.WithString("spec_dir", mcp.Description("SPEC directory (for op=compute_hash).")),
		mcp.WithString("spec_id", mcp.Description("SPEC id (for op=lookup, op=store).")),
		mcp.WithString("hash", mcp.Description("Plan-artifact hash (for op=lookup, op=store).")),
		mcp.WithString("auditor_version", mcp.Description("Auditor identifier (for op=store).")),
		mcp.WithString("report_path", mcp.Description("Audit report path (for op=store).")),
		// SPEC-CODEX-WIRING-001 (REQ-CW-011 / spec §A.4): audit_cache is a
		// catalog-READ tool; declaring read-only keeps its effective annotation
		// equal to the catalog classification so the capability-based `writes`
		// approval mode does not prompt for it.
		mcp.WithReadOnlyHintAnnotation(true),
	), handleAuditCache)

	// --- M2 codex backend (design.md §3 M2) ---

	// codex_audit → codex binary shellout (REQ-MCP-006 / AC-MCP-007). mode=native
	// dispatches codex review/start; mode=adversarial dispatches turn/start + the
	// adversarial-review prompt. Output adopts review-output.schema.json (§G.4).
	// Fail-open on missing codex (VerdictInconclusive). The codex binary is
	// OPTIONAL + experimental (R1).
	add("codex_audit", mcp.NewTool(
		"codex_audit",
		mcp.WithDescription("Run a codex code review. mode=native → codex review/start; mode=adversarial → codex turn/start + an adversarial-review prompt. Returns a review-output schema (verdict/summary/findings/next_steps). codex is OPTIONAL; a missing or unavailable codex yields verdict 'inconclusive' (fail-open)."),
		mcp.WithString("mode", mcp.Enum(codexModeNative, codexModeAdversarial), mcp.Description("Audit mode: 'native' (codex review/start) or 'adversarial' (codex turn/start + red-team prompt). Defaults to native.")),
		mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), mcp.Description("What codex reviews: 'uncommittedChanges' or 'baseBranch'. For 'baseBranch' the branch name is resolved SERVER-SIDE and cannot be supplied here — it is read from the reviewed tree, the remote default head first and then 'main', the same chain the GLM backend uses so both review the same change. A tree where neither resolves returns 'inconclusive' naming that cause rather than reviewing something else.")),
		mcp.WithString("focus", mcp.Description("Adversarial-only focus area (e.g. 'concurrency', 'auth').")),
		mcp.WithString("model", mcp.Description("Optional model override (resolved via the model/effort SSOT in M3).")),
		projectRootOption(),
		mcp.WithOutputSchema[ReviewOutput](),
		// SPEC-CODEX-WIRING-001 (REQ-CW-011 / spec §A.4): catalog-READ tool —
		// see the audit_cache note.
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexAudit)

	// codex_setup → Go probe (REQ-MCP-007 / AC-MCP-008): exec.LookPath("codex")
	// + codex --version + auth-provider classification + the enable_review_gate
	// toggle. NO Node bridge. Read-only.
	add("codex_setup", mcp.NewTool(
		"codex_setup",
		mcp.WithDescription("Probe the local codex installation: exec.LookPath + codex --version + auth-provider classification (ChatGPT / apiKey / provider / unknown) + the current enable_review_gate toggle state. Pure Go — no Node bridge."),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexSetup)

	// --- codex task delegation + job lifecycle (SPEC-CODEX-PHASE2-001 M5) ---
	//
	// Each tool name is written as a quoted literal here, matching the shape
	// "codex_audit" / "codex_setup" use above. The handlers carry the same names
	// as constants (codexTaskToolName, codexJob*ToolName); the two cannot drift,
	// because TestCodexJobTools_RegistrationShape looks each registered tool up
	// BY the constant.
	//
	// The read-only hints below are stated explicitly, including the `false`
	// ones. mcp.NewTool already seeds ReadOnlyHint to false, so the write tools
	// would carry the right value without the call — but a hint that is correct
	// only by inheriting a library default is indistinguishable from one nobody
	// considered, and REQ-CX2-013 is a statement about what each tool DOES.

	// codex_task → handleCodexTask (REQ-CX2-006/007/008). Drives a codex turn
	// with the caller's prompt. NOT read-only: it starts a turn, and with the
	// project opted in via workflow.codex.task.allow_write it can modify the
	// working tree. Fail-open on a missing codex.
	add("codex_task", mcp.NewTool(
		"codex_task",
		mcp.WithDescription("Delegate a coding or investigation task to codex. Returns the completed output when background is false, or a job id (observable via codex_job_status / codex_job_result, stoppable via codex_job_cancel) when background is true. Writing the working tree requires the project opt-in workflow.codex.task.allow_write; without it the turn runs read-only and the result says the write was not honored. codex is OPTIONAL; a missing or unavailable codex yields a structured fail-open result."),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The task to give codex.")),
		mcp.WithBoolean("background", mcp.Description("Run the turn in the background and return a job id immediately. Background jobs live inside this server process and do not survive its exit.")),
		mcp.WithBoolean("write", mcp.Description("Request that codex be allowed to modify the working tree. Honored only when the project has opted in (workflow.codex.task.allow_write: true); otherwise the turn runs read-only and the result states the refusal.")),
		mcp.WithBoolean("resume_last", mcp.Description("Continue the most recently recorded codex thread for this project instead of opening a new one. When no thread is recorded, a new one is opened and the result says so.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleCodexTask)

	// codex_job_status → handleCodexJobStatus (REQ-CX2-009). Reads one record.
	add("codex_job_status", mcp.NewTool(
		"codex_job_status",
		mcp.WithDescription("Read a codex job's record (status, timestamps, thread id, turn id, pid, mode, request summary). An unknown or unreadable job is a structured result, not a failed call."),
		mcp.WithString(codexJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background codex_task.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexJobStatus)

	// codex_job_result → handleCodexJobResult (REQ-CX2-010). Reads one record;
	// never waits for the turn.
	add("codex_job_result", mcp.NewTool(
		"codex_job_result",
		mcp.WithDescription("Read a codex job's output. A job in a terminal status returns its recorded output; a job still running returns its current status without blocking the caller."),
		mcp.WithString(codexJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background codex_task.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexJobResult)

	// codex_job_cancel → handleCodexJobCancel (REQ-CX2-011/012). NOT read-only:
	// it interrupts a turn and may terminate the process this server spawned for
	// the job. It signals ONLY that process — a record this server lifetime did
	// not start is refused rather than signalled.
	add("codex_job_cancel", mcp.NewTool(
		"codex_job_cancel",
		mcp.WithDescription("Stop a running codex job: sends turn/interrupt on that job's own session and, if the turn does not end within a bounded grace window, terminates the codex process this server spawned for it. A job this server lifetime did not start is refused rather than signalled; an already-terminal job returns its status and sends nothing."),
		mcp.WithString(codexJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background codex_task.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleCodexJobCancel)

	// --- GLM task delegation + job lifecycle ---
	//
	// Mirrors the codex task/job family above against the GLM (z.ai) backend:
	// same job model (queued → running → terminal), same in-process background
	// execution, same fail-open posture. GLM is reached over HTTPS through the
	// mcp_glm.go plumbing — there is no spawned process, so cancel revokes the
	// job's context instead of signalling a pid.
	//
	// The read-only hints are stated explicitly, including the `false` ones,
	// for the same reason as the codex block above: a hint that is correct only
	// by inheriting a library default is indistinguishable from one nobody
	// considered.

	// glm_task → handleGLMTask. Delegates a prompt to GLM. NOT read-only: it
	// starts a model call and writes a job record under .moai/state/glm-jobs/.
	// Fail-open on a missing key or an unavailable z.ai.
	add("glm_task", mcp.NewTool(
		"glm_task",
		mcp.WithDescription("Delegate a task (arbitrary prompt) to GLM (z.ai). Returns the completed text output when background is false, or a job id (observable via glm_job_status / glm_job_result, stoppable via glm_job_cancel) when background is true. Background jobs live inside this server process and do not survive its exit. GLM is OPTIONAL; a missing key or an unavailable z.ai yields a structured fail-open result, never a tool error."),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The task to give GLM.")),
		mcp.WithBoolean("background", mcp.Description("Run the task in the background and return a job id immediately. Background jobs live inside this server process and do not survive its exit.")),
		mcp.WithString("model", mcp.Description("Optional GLM model override; omitted ⇒ resolved via the model/effort SSOT (ResolveAgentModelEffort).")),
		mcp.WithString("system", mcp.Description("Optional system prompt for the task.")),
		mcp.WithInteger("max_tokens", mcp.Description("Optional response token bound; defaults to the server-side glm_task ceiling.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleGLMTask)

	// glm_job_status → handleGLMJobStatus. Reads one record.
	add("glm_job_status", mcp.NewTool(
		"glm_job_status",
		mcp.WithDescription("Read a GLM job's record (status, timestamps, model, request summary). An unknown or unreadable job is a structured result, not a failed call."),
		mcp.WithString(glmJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background glm_task.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMJobStatus)

	// glm_job_result → handleGLMJobResult. Reads one record; never waits for
	// the call.
	add("glm_job_result", mcp.NewTool(
		"glm_job_result",
		mcp.WithDescription("Read a GLM job's output. A job in a terminal status returns its recorded output; a job still running returns its current status without blocking the caller."),
		mcp.WithString(glmJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background glm_task.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMJobResult)

	// glm_job_cancel → handleGLMJobCancel. NOT read-only: it cancels a running
	// job and records the cancellation. A job this server lifetime did not
	// start is refused rather than cancelled.
	add("glm_job_cancel", mcp.NewTool(
		"glm_job_cancel",
		mcp.WithDescription("Stop a running GLM job: records the cancellation and revokes the context its in-flight z.ai request runs under, then waits a bounded grace window for the job to end. A job this server lifetime did not start is refused rather than cancelled; an already-terminal job returns its status and sends nothing."),
		mcp.WithString(glmJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background glm_task.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleGLMJobCancel)

	// --- M3 GLM backend (design.md §3 M3) ---

	// glm_audit → z.ai GLM API direct call (REQ-MCP-009 / AC-MCP-011). Calls
	// the Anthropic-compatible https://api.z.ai/api/anthropic/v1/messages
	// endpoint directly — NOT the z.ai MCP server, NOT any gateway. Output
	// adopts the shared review-output.schema.json (§G.4). Fail-open on a
	// missing/unauthenticated/erroring/malformed GLM (VerdictInconclusive →
	// claude fallback, REQ-MCP-012). Model/effort resolved via the SSOT
	// (ResolveAgentModelEffort, REQ-MCP-013) when no explicit model is given.
	add("glm_audit", mcp.NewTool(
		"glm_audit",
		mcp.WithDescription("Run a GLM (z.ai) code review. Calls the z.ai Anthropic-compatible endpoint directly and returns a review-output schema (verdict/summary/findings/next_steps). GLM is OPTIONAL; a missing/unauthenticated/erroring GLM yields verdict 'inconclusive' (fail-open → fall back to the active auditor)."),
		mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), mcp.Description("What GLM reviews: 'uncommittedChanges' (default) or 'baseBranch'. The named change is collected as a diff and sent in the request — GLM has no filesystem, so a target that yields no diff returns 'inconclusive' rather than a verdict.")),
		mcp.WithString("focus", mcp.Description("Optional focus area (e.g. 'concurrency', 'auth', 'secret handling').")),
		mcp.WithString("model", mcp.Description("Optional GLM model override; omitted ⇒ resolved via the model/effort SSOT (ResolveAgentModelEffort).")),
		projectRootOption(),
		mcp.WithOutputSchema[ReviewOutput](),
		// SPEC-CODEX-WIRING-001 (REQ-CW-011 / spec §A.4): catalog-READ tool —
		// see the audit_cache note.
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMAudit)

	// audit_multi → SPEC-AUDIT-MULTI-MODEL-001 multi-auditor convergence
	// (REQ-AMM-009 / REQ-AMM-010 / AC-AMM-012 / AC-AMM-013). Thin wrapper over
	// runMultiAudit (mcp_convergence.go) — does NOT re-implement the
	// codex/glm backends (C1). The claude_verdict argument is the always-available
	// anchor; gates is an optional per-auditor override (defaults: claude+codex
	// required, glm advisory).
	add(auditMultiToolName, mcp.NewTool(
		auditMultiToolName,
		mcp.WithDescription("Run the multi-auditor convergence engine over a claude_verdict anchor + optional codex/glm backend verdicts. Returns a ConvergenceResult (overall verdict + per-auditor verdicts + residual_risk_note). Backend fan-out reuses the existing codex/glm handlers — no re-implementation."),
		mcp.WithObject("claude_verdict", mcp.Description("The always-available claude anchor verdict (review-output schema: verdict/summary/findings/next_steps).")),
		mcp.WithString("target", mcp.Description("Optional review target (file path, diff ref, or scope label).")),
		mcp.WithString("focus", mcp.Description("Optional focus area (e.g. 'concurrency', 'auth', 'secret handling').")),
		mcp.WithObject("gates", mcp.Description("Optional per-auditor gate override (keys: claude/codex/glm; values: off|advisory|required). Defaults: claude+codex required, glm advisory.")),
		mcp.WithString("session_id", mcp.Description("Optional session id for per-session convergence state persistence (.moai/state/audit-multi/<session>.json). Empty ⇒ persistence no-op.")),
		// Pass-through semantics: this fan-out handed its backends no root
		// before the parameter existed, so an absent parameter must keep
		// handing them none rather than substituting a default.
		projectRootPassthroughOption(),
		// SPEC-CODEX-WIRING-001 (REQ-CW-011 / spec §A.4): catalog-READ tool —
		// see the audit_cache note.
		mcp.WithReadOnlyHintAnnotation(true),
	), handleAuditMulti)

	// --- Session messaging broker (SPEC-CODEX-SESSION-MSG-001 M2, design.md §6) ---
	//
	// Four SYMMETRIC tools over internal/sessionmsg: any session kind
	// (claude|codex) registers, lists, sends, polls — the broker is the
	// transport for Claude↔Codex messaging on one machine. Every description
	// carries the discipline short-form (REQ-CSM-014): the tool description
	// loads into the session context, so it is the surface a Codex reader
	// actually sees. Read-only hints are stated explicitly, including the
	// false ones — only session_msg_list is read-only.

	// session_msg_register → handleSessionMsgRegister. Idempotent
	// kind+name registration; mutates the agent record (heartbeat), so NOT
	// read-only.
	add("session_msg_register", mcp.NewTool(
		"session_msg_register",
		mcp.WithDescription("Register this session in the local message broker (kind: claude|codex + name) and get a stable agentId; re-registering the same kind+name refreshes the heartbeat and returns the same id. "+sessionMsgDisciplineShortForm),
		mcp.WithString("kind", mcp.Required(), mcp.Enum(sessionmsg.KindClaude, sessionmsg.KindCodex), mcp.Description("Session kind: 'claude' or 'codex'.")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Short stable name identifying this session to peers (same kind+name re-registers idempotently).")),
		mcp.WithString("description", mcp.Description("Optional human-readable description stored on the agent record.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleSessionMsgRegister)

	// session_msg_list → handleSessionMsgList. Reads the agent roster only.
	add("session_msg_list", mcp.NewTool(
		"session_msg_list",
		mcp.WithDescription("List registered message-broker agents (agentId, name, kind, online flag, pending count). "+sessionMsgDisciplineShortForm),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSessionMsgList)

	// session_msg_send → handleSessionMsgSend. Writes the recipient's
	// pending mailbox; NOT read-only.
	add("session_msg_send", mcp.NewTool(
		"session_msg_send",
		mcp.WithDescription("Send a short message to another registered agent's mailbox (they receive it on their next session_msg_poll). "+sessionMsgDisciplineShortForm),
		mcp.WithString("from_agent_id", mcp.Required(), mcp.Description("Sender agentId (from session_msg_register).")),
		mcp.WithString("to_agent_id", mcp.Required(), mcp.Description("Recipient agentId; an unknown id returns a structured error listing the known agents.")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Message text (bounded by the broker's text ceiling).")),
		mcp.WithObject("data", mcp.Description("Optional structured JSON payload appended as a data part.")),
		mcp.WithString("context_id", mcp.Description("Optional A2A contextId carried on the message envelope.")),
		mcp.WithString("task_id", mcp.Description("Optional A2A taskId carried on the message envelope.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleSessionMsgSend)

	// session_msg_poll → handleSessionMsgPoll. Claims a batch out of the
	// caller's mailbox and applies ack_ids deletions; mutates mailbox state,
	// so NOT read-only.
	add("session_msg_poll", mcp.NewTool(
		"session_msg_poll",
		mcp.WithDescription("Claim pending messages from this agent's mailbox (at-least-once; unacked claims return after the claim TTL) and optionally ack processed message ids. "+sessionMsgDisciplineShortForm),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("Polling agent's agentId (from session_msg_register).")),
		mcp.WithArray("ack_ids", mcp.Items(map[string]any{"type": "string"}), mcp.Description("Optional messageIds to acknowledge (delete from the claimed set).")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleSessionMsgPoll)
	// --- M5 code-query tools (SPEC-V3R6-GRAPH-FRESHNESS-001 REQ-GF-017..019).
	// Signature-level answers from the code-derived layer; every response
	// carries tree+commit provenance. Read-only hints per the audit-family
	// precedent.
	add("graph_file_api", mcp.NewTool(
		"graph_file_api",
		mcp.WithDescription("List a source file's exported declarations with signatures — no source bodies. Answers name the tree root + commit they were computed from."),
		mcp.WithString("file", mcp.Required(), mcp.Description("Repository-relative file path (e.g. internal/graph/check.go).")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGraphFileAPI)

	add("graph_find_code", mcp.NewTool(
		"graph_find_code",
		mcp.WithDescription("Search the code-derived edge layer for a symbol: callee sites (where it is called) and caller observations, each with its resolution grade and per-edge resolution confidence (1.0 extracted / 0.95 intra-package / 0.85 inferred; absent on pre-upgrade artifacts). Answer names the tree root + commit."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Symbol name to search for (e.g. 'CheckFreshness').")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGraphFindCode)

	add("graph_trace_calls", mcp.NewTool(
		"graph_trace_calls",
		mcp.WithDescription("Traverse code-call edges from a symbol: callers (who reaches it) and callees (what it calls), up to depth hops over the code-derived layer, each edge carrying its resolution confidence (absent on pre-upgrade artifacts). Answer names the tree root + commit."),
		mcp.WithString("symbol", mcp.Required(), mcp.Description("Symbol name to trace from (e.g. 'RefreshIndex').")),
		mcp.WithInteger("depth", mcp.Description("Traversal depth in hops (default 1, capped at 8).")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGraphTraceCalls)

	// SPEC-GRAPH-REPORT-001 REQ-GR-001: the A→B reachability query. The
	// description restates the shared 8-hop cap exactly as graph_trace_calls
	// does (REQ-GR-002) — a static string cannot reference the const, which
	// remains the enforced bound (codequery.go maxTraceDepth).
	add("graph_shortest_path", mcp.NewTool(
		"graph_shortest_path",
		mcp.WithDescription("Shortest call path from one symbol to another over code-call edges (capped at 8 hops), each hop carrying its line and resolution confidence (absent on pre-upgrade artifacts). Endpoints are file:function node ids or bare symbol names resolving to exactly one node; no-path and ambiguous-name answers are structured results naming both endpoints and the cap. Answer names the tree root + commit."),
		mcp.WithString("from", mcp.Required(), mcp.Description("Path start: a node id (file:function) or a bare symbol name resolving to exactly one node.")),
		mcp.WithString("to", mcp.Required(), mcp.Description("Path end: a node id (file:function) or a bare symbol name resolving to exactly one node.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGraphShortestPath)
}

// readMCPToolEnablement reads the per-tool enablement map from
// `.moai/config/sections/mcp.yaml` (SPEC-MCP-CONSOLE-001 REQ-C-2). The default
// is ALL-ENABLED (owner decision 2026-08-12), so the truth table is fail-OPEN —
// the opposite of the codex gates' fail-CLOSED posture:
//
//   - projectDir empty                      → all enabled (empty map ⇒ every key absent ⇒ enabled)
//   - file missing / unreadable             → all enabled
//   - YAML parse error                      → all enabled
//   - `mcp.tools` block absent              → all enabled
//   - `mcp.tools.<name>.enabled: false`     → that tool DISABLED
//   - `mcp.tools.<name>.enabled: true`      → that tool enabled
//   - `<name>` key absent entirely          → that tool enabled (default)
//
// The key path is NESTED under the file's `mcp:` root, matching the shape the
// schema seam (settings.mcpFields) writes via yamlpatch and the shape the
// shipped template deploys. Only the literal `enabled:` bool is read; unknown
// keys under `mcp.tools.<name>` are ignored.
//
// This stays a small hand-rolled read rather than a config.Loader.Load call for
// the same reason readCodexReviewGateEnabled does: registration runs once at
// server construction and should touch exactly one file with no
// defaults/env-override machinery, and each failure path here returns the
// all-enabled default explicitly rather than blurring "unreadable" into a
// populated default. The shape is pinned against the real loader by
// TestReadMCPToolEnablement_AgreesWithConfigLoader (deferred to a follow-up;
// the codex reader's TestReviewGateReaders_AgreeWithConfigLoader is the
// precedent shape).
//
// @MX:NOTE: [AUTO] fail-OPEN default (all-enabled) — the inverse posture of the
// codex review-gate / task-allow-write readers, which fail CLOSED. The owner
// decision (2026-08-12) chose all-enabled so SPEC-B agent grants stay functional
// out of the box; a too-permissive default is surfaceable via REQ-C-3's
// write-capable distinction, which is a better failure mode than silent inert tools.
// @MX:SPEC: SPEC-MCP-CONSOLE-001
func readMCPToolEnablement(projectDir string) map[string]bool {
	enabled := make(map[string]bool, 21)
	// Default: every catalog tool enabled. Keys absent from mcp.yaml stay true.
	for _, name := range mcpcat.MoaiMCPToolNames() {
		enabled[name] = true
	}
	if projectDir == "" {
		return enabled
	}
	configPath := filepath.Join(projectDir, ".moai", "config", "sections", "mcp.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return enabled // file missing/unreadable → all enabled (default)
	}
	var doc struct {
		MCP struct {
			Tools map[string]struct {
				Enabled *bool `yaml:"enabled"`
			} `yaml:"tools"`
		} `yaml:"mcp"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return enabled // parse error → all enabled (default)
	}
	for name, t := range doc.MCP.Tools {
		if t.Enabled != nil {
			enabled[name] = *t.Enabled
		}
	}
	return enabled
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
func handleGoalArm(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	// Inherited rule (same predicate the CLI applies): a mechanical condition
	// whose first word resolves to no command can never exit 0, so arming it
	// buys a goal that blocks every turn-end to the ceiling. Both arm paths call
	// parseCondition, so both need the gate — a CLI-only check would leave the
	// MCP path arming exactly what the CLI refuses.
	// An explicit `cmd:` prefix exempts the condition here exactly as it does on
	// the CLI path (declaredMechanical) — the two arm surfaces must not disagree
	// about which conditions are armable.
	if cond.Type == goal.ConditionMechanical && !declaredMechanical(conditionText) {
		if tok, bad := unrunnableCommandToken(ctx, cond.Cmd); bad {
			return toolErr("goal_arm", unrunnableConditionError("goal_arm", tok, cond.Cmd)), nil
		}
	}
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
func handleSpecProgress(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("spec_progress", err), nil
	}
	records, err := spec.ListDocs(root)
	if err != nil {
		return toolErr("spec_progress", err), nil
	}
	return toolJSON("spec_progress", map[string]any{
		"count": len(records),
		"specs": records,
		"_root": rootProvenanceMap(root, source),
	}), nil
}

// handleVerifySnapshot wraps verify.Load (store.go:38); when a `command` arg is
// supplied it records via verify.RecordCheck (store.go:107).
func handleVerifySnapshot(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("verify_snapshot", err), nil
	}
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
		return toolJSON("verify_snapshot", map[string]any{"action": "record", "snapshot": snap, "_root": rootProvenanceMap(root, source)}), nil
	}
	snap, err := verify.Load(root, key)
	if err != nil {
		return toolErr("verify_snapshot", err), nil
	}
	return toolJSON("verify_snapshot", map[string]any{"action": "load", "snapshot": snap, "_root": rootProvenanceMap(root, source)}), nil
}

// handleVerifyTrend wraps verify.Load (store.go:38), surfacing the Checks trend.
func handleVerifyTrend(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("verify_trend", err), nil
	}
	key := req.GetString("key", "")
	if key == "" {
		return toolErr("verify_trend", fmt.Errorf("key must not be empty")), nil
	}
	snap, err := verify.Load(root, key)
	if err != nil {
		return toolErr("verify_trend", err), nil
	}
	checks := []verify.CheckEntry{}
	if snap != nil {
		checks = snap.Checks
	}
	return toolJSON("verify_trend", map[string]any{"key": key, "checks": checks, "_root": rootProvenanceMap(root, source)}), nil
}

// handleSpecAudit wraps spec.Audit (audit.go:156).
func handleSpecAudit(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr("spec_audit", err), nil
	}
	opts := spec.AuditOptions{
		BaseDir:              root,
		FilterSpec:           req.GetString("filter_spec", ""),
		FilterEra:            req.GetString("filter_era", ""),
		IncludeGrandfathered: req.GetBool("include_grandfathered", false),
	}
	result, auditErr := spec.Audit(opts)
	if auditErr != nil {
		return toolErr("spec_audit", auditErr), nil
	}
	return toolJSON("spec_audit", result), nil
}

// handleSpecDrift wraps spec.Audit (audit.go:156), filtered to modern-era drift.
func handleSpecDrift(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("spec_drift", err), nil
	}
	result, err := spec.Audit(spec.AuditOptions{BaseDir: root})
	if err != nil {
		return toolErr("spec_drift", err), nil
	}
	return toolJSON("spec_drift", map[string]any{
		"total_specs":    result.TotalSpecs,
		"modern_clean":   result.ModernEraClean,
		"drift_findings": result.DriftFindings,
		"_root":          rootProvenanceMap(root, source),
	}), nil
}

// moaiAuditCache is the process-shared AuditCache backing the audit_cache MCP
// tool. It persists across calls within ONE moai mcp-server process so that a
// store in one call is visible to a lookup in a later call (the prior per-call
// NewInMemoryCache made lookup ALWAYS return hit:false).
//
// @MX:DEBT: single-process in-memory cache only
// @MX:CEILING: single moai mcp-server process lifetime, sequential calls
// @MX:UPGRADE: the plan-audit gate (audit_gate.go) runs in a separate /moai run
//
//	process and does NOT populate this cache; true cross-process verdict reuse
//	requires lookup to read the durable daily-report cache (AuditReporter).
//
// @MX:REASON: [AUTO] fan_in >= 3 (compute_hash/lookup/store ops + tests)
var moaiAuditCache = runtime.NewInMemoryCache()

// handleAuditCache wraps runtime.AuditCache ComputeHash/Lookup/Store over the
// process-shared moaiAuditCache. The store op lets a caller record a PASS
// verdict that a later lookup (same moai mcp-server process) returns as a hit.
func handleAuditCache(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	op := req.GetString("op", "")
	switch op {
	case "compute_hash":
		specDir := req.GetString("spec_dir", "")
		if specDir == "" {
			return toolErr("audit_cache", fmt.Errorf("op=compute_hash requires spec_dir")), nil
		}
		hash, err := moaiAuditCache.ComputeHash(specDir)
		if err != nil {
			return toolErr("audit_cache", err), nil
		}
		return toolJSON("audit_cache", map[string]any{"op": op, "hash": hash}), nil
	case "lookup":
		specID := req.GetString("spec_id", "")
		hash := req.GetString("hash", "")
		entry, ok := moaiAuditCache.Lookup(specID, hash, time.Now())
		return toolJSON("audit_cache", map[string]any{"op": op, "hit": ok, "entry": entry}), nil
	case "store":
		// Record a PASS verdict so a later lookup (same process) returns a hit.
		// Only PASS verdicts are cached (audit_cache.go Store guards on VerdictPass).
		specID := req.GetString("spec_id", "")
		hash := req.GetString("hash", "")
		if specID == "" || hash == "" {
			return toolErr("audit_cache", fmt.Errorf("op=store requires spec_id and hash")), nil
		}
		moaiAuditCache.Store(specID, hash, &runtime.AuditResult{
			Verdict:        runtime.VerdictPass,
			AuditAt:        time.Now(),
			AuditorVersion: req.GetString("auditor_version", ""),
			ReportPath:     req.GetString("report_path", ""),
		})
		return toolJSON("audit_cache", map[string]any{"op": op, "stored": true, "spec_id": specID, "hash": hash}), nil
	default:
		return toolErr("audit_cache", fmt.Errorf("unknown op %q (want compute_hash|lookup|store)", op)), nil
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
