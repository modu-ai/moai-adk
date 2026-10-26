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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/homestate"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
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
	cmd := &cobra.Command{
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
			exposure, _ := c.Flags().GetString("exposure")
			if exposure == "" {
				exposure = mcpServerExposureFromEnv()
			}
			return runMCPServerWithIOAndExposure(os.Stdin, os.Stdout, exposure)
		},
	}
	// REQ-MN-009: the gate child's containment overrides pass --exposure gate
	// (the MOAI_MCP_EXPOSURE env is the equivalent input), restricting the
	// served surface to the gate read-only subset.
	cmd.Flags().String("exposure", "", `Serve a restricted surface: "gate" serves only the gate read-only subset (set by the gate child's containment overrides).`)
	return cmd
}

// mcpServerExposureFromEnv reads the exposure input MOAI_MCP_EXPOSURE (the
// flag-free carrier; the gate child's env table can set it).
func mcpServerExposureFromEnv() string {
	return strings.TrimSpace(os.Getenv("MOAI_MCP_EXPOSURE"))
}

// runMCPServer builds the moai MCP server and serves it over stdio (blocking,
// until the stdio stream closes). REQ-MCP-001. The exposure comes from the
// --exposure flag, defaulted by MOAI_MCP_EXPOSURE (REQ-MN-009: the gate
// child's containment overrides pass "gate").
func runMCPServer() error {
	return runMCPServerWithIOAndExposure(os.Stdin, os.Stdout, mcpServerExposureFromEnv())
}

// runMCPServerWithIOAndExposure serves under an explicit exposure (REQ-MN-009:
// "gate" restricts the surface to the gate read-only subset).
func runMCPServerWithIOAndExposure(stdin io.Reader, stdout io.Writer, exposure string) error {
	projectDir := resolveProjectDir()
	admissionLock, lockErr := homestate.AcquireAdmissionLock(projectDir)
	if lockErr != nil {
		return lockErr
	}
	if err := homestate.CheckRuntimeAdmission(projectDir); err != nil {
		_ = admissionLock.Release()
		return err
	}
	// REQ-MN-015: the runtime stamp is BEST-EFFORT — a failed stamp (a
	// read-only state dir, a non-project cwd) never blocks serving. The
	// pre-fix code returned here, which made a read-only project unable to
	// start the server at all.
	if recordPath, err := writeMCPServerRuntimeRecord(projectDir); err == nil {
		defer removeMCPServerRuntimeRecord(recordPath)
		_ = admissionLock.Release()
	} else {
		_ = admissionLock.Release()
	}
	s := newMoaiMCPServerWithExposure(exposure)
	// REQ-MN-015: the call log binds to the serving project — the server run
	// is the only writer; direct-construction tests leave it nil (disabled).
	mcpCallLog = newMCPCallLogWriter()
	// REQ-MN-014: the shutdown pass cancels every background job family —
	// GLM jobs, role-audit jobs — so nothing outlives the server that started
	// it (the codex task/job family left with t1909).
	defer func() {
		stopGLMBackgroundJobs()
		stopCodexRoleAuditJobs()
	}()
	// ServeStdio waits for in-flight tool workers on EOF before cancelling its
	// internal context. A worker stalled in an app-server handshake would keep
	// that wait (and our shutdown cleanup) blocked. Cancel on EOF before waiting.
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	ctx, cancel := context.WithCancel(signalCtx)
	defer cancel()
	reader := &mcpEOFReader{Reader: stdin, cancel: cancel}
	err := server.NewStdioServer(s).Listen(ctx, reader, stdout)
	if reader.eof.Load() && errors.Is(err, context.Canceled) {
		return nil // EOF is normal MCP shutdown, not a command failure.
	}
	return err
}

// mcpEOFReader forwards every byte unchanged and ends the request context when
// the host closes stdin, including while a tool handler is still running.
type mcpEOFReader struct {
	io.Reader
	cancel     context.CancelFunc
	eof        atomic.Bool
	pendingEOF atomic.Bool
}

func (r *mcpEOFReader) Read(p []byte) (int, error) {
	if r.pendingEOF.Swap(false) {
		r.eof.Store(true)
		r.cancel()
		return 0, io.EOF
	}
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		if n > 0 {
			// Process the final bytes before ending an in-flight tool request.
			r.pendingEOF.Store(true)
			return n, nil
		}
		r.eof.Store(true)
		r.cancel()
	}
	return n, err
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
		// REQ-MN-008 wire honesty: the long-running tools emit
		// notifications/message progress logs, so the server declares the
		// logging capability that emission implies.
		server.WithLogging(),
		server.WithInstructions(moaiMCPServerInstructions()),
	)
	registerMoaiMCPTools(s, resolveProjectDir())
	return s
}

// newMoaiMCPServerWithExposure is newMoaiMCPServer with an explicit exposure
// (SPEC-MCP-NEUTRALITY-001 REQ-MN-009): "gate" serves ONLY the gate
// read-only subset; empty serves the full catalogue.
func newMoaiMCPServerWithExposure(exposure string) *server.MCPServer {
	s := server.NewMCPServer(
		moaiMCPServerName,
		version.GetFullVersion(),
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithInstructions(moaiMCPServerInstructions()),
	)
	registerMoaiMCPToolsWithExposure(s, resolveProjectDir(), exposure)
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
// mcpStalenessGuard wraps one handler with the REQ-MN-013 per-call staleness
// contract: a WRITE-CAPABLE tool refuses while the serving build is stale,
// and every STALE result carries the reconnect guidance plus the stale_server
// and _server markers. Fresh results pass through untouched.
func mcpStalenessGuard(name string, handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		info := serverStalenessInfo()
		if !info.Stale {
			return handler(ctx, req)
		}
		guidance := fmt.Sprintf(
			"stale_server: the serving moai build (%s) differs from the installed build (%s) — reconnect the moai MCP server (restart the host session) so the installed build serves",
			info.Serving, info.Installed)
		if mcpCallIsWriteCapable(name, req) {
			res := toolErr(name, fmt.Errorf("the serving moai build is stale — %s", guidance))
			return addMCPStalenessMetadata(res, info, guidance), nil
		}
		res, err := handler(ctx, req)
		if res == nil {
			return res, err
		}
		return addMCPStalenessMetadata(res, info, guidance), err
	}
}

// mcpCallIsWriteCapable includes operation-level writes on tools whose
// catalogue entry is read-only for its other operations.
func mcpCallIsWriteCapable(name string, req mcp.CallToolRequest) bool {
	if name == "audit_cache" {
		return req.GetString("op", "") == "store"
	}
	return mcpcat.IsWriteCapable(name)
}

// addMCPStalenessMetadata mirrors stale-server identity into both MCP result
// channels. A stale write refusal is an IsError result with a JSON text block;
// successful stale reads keep their original structured payload and guidance.
func addMCPStalenessMetadata(res *mcp.CallToolResult, info serverStaleness, guidance string) *mcp.CallToolResult {
	if res == nil {
		return nil
	}
	var payload map[string]any
	if res.StructuredContent != nil {
		if body, err := json.Marshal(res.StructuredContent); err == nil {
			_ = json.Unmarshal(body, &payload)
		}
	}
	if payload == nil {
		payload = make(map[string]any)
		for _, item := range res.Content {
			block, ok := item.(mcp.TextContent)
			if !ok {
				continue
			}
			if err := json.Unmarshal([]byte(block.Text), &payload); err != nil || payload == nil {
				payload = map[string]any{"error": block.Text}
			}
			break
		}
	}
	payload["stale_server"] = true
	payload["_server"] = map[string]any{"version": info.Version, "commit": info.Commit}
	res.StructuredContent = payload

	body, err := json.Marshal(payload)
	if err == nil {
		text := string(body)
		updatedText := false
		for i, item := range res.Content {
			block, ok := item.(mcp.TextContent)
			if !ok {
				continue
			}
			var existing map[string]any
			if json.Unmarshal([]byte(block.Text), &existing) == nil && existing != nil {
				block.Text = text
				res.Content[i] = block
				updatedText = true
				break
			}
		}
		if !updatedText {
			for i, item := range res.Content {
				block, ok := item.(mcp.TextContent)
				if !ok {
					continue
				}
				block.Text = text
				res.Content[i] = block
				updatedText = true
				break
			}
		}
		if !updatedText {
			res.Content = append(res.Content, mcp.TextContent{Type: "text", Text: text})
		}
	}
	if guidance != "" {
		res.Content = append(res.Content, mcp.TextContent{Type: "text", Text: guidance})
	}
	return res
}

// mcpCatalogueGate is the fail-closed check every registration runs
// (SPEC-MCP-NEUTRALITY-001 REQ-MN-006): the enablement map is catalogue-seeded,
// so a registered name absent from the catalogue gets no enablement key and
// silently never serves — that silence is how todo_claim died (plan B4). A
// catalogue-absent registration fails the construction loudly instead.
func mcpCatalogueGate(name string) {
	if !mcpcat.HasTool(name) {
		panic("moai mcp: tool " + name + " is registered but absent from the internal/mcp catalogue; add the ToolDef — a catalogue-absent registration never serves")
	}
}

func registerMoaiMCPTools(s *server.MCPServer, projectDir string) {
	registerMoaiMCPToolsWithExposure(s, projectDir, "")
}

// moaiMCPExposureGate is the exposure value the gate-session containment
// serves: ONLY the gate read-only subset (plan §F M3's enumerated set).
const moaiMCPExposureGate = "gate"

// moaiGateExposureTools is the enumerated gate-session read-only subset
// (REQ-MN-009, plan §F M3): every served tool whose registered annotation is
// read-only and whose handler performs no write to project state. The
// exposure test asserts the SERVED set equals this set exactly — the
// derivation rule's output is fixed here so a new read-only tool must DECIDE
// whether it is gate-visible, in this one table.
var moaiGateExposureTools = map[string]bool{
	"spec_progress":       true,
	"spec_audit":          true,
	"spec_drift":          true,
	"verify_trend":        true,
	"session_list":        true,
	"audit_cache":         true, // read op only: the store op is refused under this exposure
	"goal_status":         true,
	"graph_file_api":      true,
	"graph_find_code":     true,
	"graph_trace_calls":   true,
	"graph_shortest_path": true,
	"todo_list":           true,
	"jev_ask":             true,
	"codex_setup":         true,
	"glm_job_status":      true,
	"glm_job_result":      true,
}

// moaiCurrentExposure carries the exposure the process's server was
// constructed under (REQ-MN-009). One server per process is the production
// reality, so the handlers that enforce an exposure-dependent rule read this
// package-level value instead of threading the exposure through every
// handler signature.
var moaiCurrentExposure string

// registerMoaiMCPToolsWithExposure registers the tool surface under an
// exposure (REQ-MN-009): the "gate" exposure intersects the enablement map
// with the enumerated read-only subset, so a gate session's moai server
// serves exactly that subset and nothing else.
func registerMoaiMCPToolsWithExposure(s *server.MCPServer, projectDir, exposure string) {
	moaiCurrentExposure = exposure
	// SPEC-MCP-CONSOLE-001 M1 (REQ-C-2): per-tool enablement consulted at
	// registration time. A disabled tool is NOT registered and does NOT appear
	// in tools/list. Default all-enabled (owner decision 2026-08-12) — a
	// missing/unreadable/absent mcp.yaml registers every tool.
	enabled := readMCPToolEnablement(projectDir)
	if exposure == moaiMCPExposureGate {
		gated := make(map[string]bool, len(moaiGateExposureTools))
		for name := range moaiGateExposureTools {
			if enabled[name] {
				gated[name] = true
			}
		}
		enabled = gated
	}
	// add gates one tool registration on the per-tool enablement map. The name
	// argument MUST match both the mcp.NewTool name and the catalog
	// (internal/mcp MoaiMCPTools) — the guard test
	// TestMoaiMCPServer_RegistrationMatchesCatalog enforces set equality, and
	// mcpCatalogueGate fails the construction on a catalogue-absent name.
	//
	// The gate also applies the two surface-wide schema policies of
	// SPEC-MCP-NEUTRALITY-001 REQ-MN-007: every served input schema declares
	// its additional-properties policy (strict), and a read-only tool never
	// carries the SDK's default destructive hint.
	add := func(name string, tool mcp.Tool, handler server.ToolHandlerFunc) {
		mcpCatalogueGate(name)
		tool.InputSchema.AdditionalProperties = false
		if tool.Annotations.ReadOnlyHint != nil && *tool.Annotations.ReadOnlyHint {
			noDestructive := false
			tool.Annotations.DestructiveHint = &noDestructive
		}
		// REQ-MN-015: one row per tool call into the rotated call log — tool,
		// duration, and the refused/error classification (M1's refused
		// marking reused). The log write is best-effort and never blocks the
		// result. Card-review P2-10: the timer starts INSIDE logged — a
		// registration-time start would bill every tool the registration-to-
		// first-call gap as its own duration.
		wrapped := mcpStalenessGuard(name, handler)
		logged := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			started := time.Now()
			res, err := wrapped(ctx, req)
			mcpCallLog.record(name, time.Since(started), res, err)
			return res, err
		}
		if enabled[name] {
			s.AddTool(tool, logged)
		}
	}

	// --- session_list → session.QueryActiveWork (registry.go:254) ---
	add("session_list", mcp.NewTool(
		"session_list",
		mcp.WithDescription("List active moai sessions (optional spec_id filter). Wraps session.QueryActiveWork."),
		mcp.WithString("spec_id", mcp.Description("Optional SPEC-ID filter; empty lists all active sessions.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSessionList)

	// --- goal_status → goal.LoadGoal (state.go:57) ---
	add("goal_status", mcp.NewTool(
		"goal_status",
		mcp.WithDescription("Read the armed-goal state for a session. Wraps goal.LoadGoal."),
		// session_id is OPTIONAL here by enforcement (goal.LoadGoal reads it
		// optionally), so the schema no longer declares it required — the
		// required-consistency half of REQ-MN-007.
		mcp.WithString("session_id", mcp.Description("Session id whose goal state is read.")),
		projectRootOption(),
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
		requiredProjectRootOption(),
	), handleGoalArm)

	// --- spec_progress → spec.ListDocs (listdocs.go:36) — the SPEC scanner ---
	add("spec_progress", mcp.NewTool(
		"spec_progress",
		mcp.WithDescription("List SPEC documents + frontmatter under a project root (SPEC lifecycle distribution source). Wraps spec.ListDocs."+worktreeCatalogueDescNote),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecProgress)

	// --- verify_snapshot → verify.Load (store.go:38) + verify.RecordCheck (store.go:107) ---
	add("verify_snapshot", mcp.NewTool(
		"verify_snapshot",
		mcp.WithDescription("Read (or record into) the per-key verification snapshot. Wraps verify.Load (+ verify.RecordCheck when a check is supplied). First CLI/MCP surface for verify."+worktreeStateDescNote),
		projectRootOption(),
		mcp.WithString("key", mcp.Required(), mcp.Description("Snapshot key (HEAD:digest form).")),
		mcp.WithString("command", mcp.Description("When set, RECORD a check entry via verify.RecordCheck instead of reading.")),
		mcp.WithInteger("exit_code", mcp.Description("Exit code for a recorded check (used with 'command').")),
	), handleVerifySnapshot)

	// --- verify_trend → verify.Load (store.go:38) — the per-key check history ---
	add("verify_trend", mcp.NewTool(
		"verify_trend",
		mcp.WithDescription("Read the per-key verification check history (trend). Wraps verify.Load, surfacing the Checks sequence."+worktreeStateDescNote),
		projectRootOption(),
		mcp.WithString("key", mcp.Required(), mcp.Description("Snapshot key whose check trend is read.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleVerifyTrend)

	// --- spec_audit → spec.Audit (audit.go:156) ---
	add("spec_audit", mcp.NewTool(
		"spec_audit",
		mcp.WithDescription("Run the SPEC lifecycle audit (era classification + drift detection). Wraps spec.Audit."+worktreeCatalogueDescNote),
		mcp.WithString("filter_spec", mcp.Description("Optional SPEC-ID filter (exact match).")),
		mcp.WithString("filter_era", mcp.Description("Optional era filter (e.g. V3R6).")),
		mcp.WithBoolean("include_grandfathered", mcp.Description("Surface grandfathered (V2.x/V3R2-R4/V3R5) SPECs as INFO.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleSpecAudit)

	// --- spec_drift → spec.Audit (audit.go:156), modern-era drift view ---
	add("spec_drift", mcp.NewTool(
		"spec_drift",
		mcp.WithDescription("Read SPEC lifecycle drift (modern-era V3R6 drift findings only). Wraps spec.Audit, filtered to drift."+worktreeCatalogueDescNote),
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
		projectRootOption(),
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
		mcp.WithDescription("Run a codex code review. mode=native → codex review/start; mode=adversarial → codex turn/start + an adversarial-review prompt. Returns a review-output schema (verdict/summary/findings/next_steps). codex is OPTIONAL; a missing or unavailable codex yields verdict 'inconclusive' (fail-open) — EXCEPT where the governing workflow.audit.gates.codex is explicitly 'required', in which case a review that produced no verdict returns verdict 'fail' with a non-empty 'gate_unmet' and isError false, the same fail-closed rule the convergence result applies. The gate is read from the reviewed tree's own workflow config; a linked worktree without its own workflow config takes it from the primary checkout instead, and treats it as 'required' when that primary cannot be identified. A 'required' gate also puts an 'audit_receipt' id on the result: cite it in the auditor verdict line. A PASS the receipt store cannot corroborate is refused by the hook-side receipt guard, which reads the gate the same way — from the primary checkout on a worktree without its own workflow config."+worktreeStateDescNote),
		mcp.WithString("mode", mcp.Enum(codexModeNative, codexModeAdversarial), mcp.Description("Audit mode: 'native' (codex review/start) or 'adversarial' (codex turn/start + red-team prompt). Defaults to native.")),
		mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), mcp.Description("What codex reviews: 'uncommittedChanges' or 'baseBranch'. For 'baseBranch' the branch name is resolved SERVER-SIDE and cannot be supplied here — it is read from the reviewed tree, the configured git_strategy.worktree_base_branch first (when set and resolvable), then the remote default head, then 'main', the same chain the GLM backend uses so both review the same change. A tree where neither resolves returns 'inconclusive' naming that cause rather than reviewing something else.")),
		mcp.WithString("focus", mcp.Description("Adversarial-only focus area (e.g. 'concurrency', 'auth').")),
		mcp.WithString("model", mcp.Description("Optional model override; omitted ⇒ the workflow.audit.codex pin (default gpt-6.1-terra), else codex's own configured default.")),
		mcp.WithString("effort", mcp.Enum(config.ValidCodexAuditEfforts()...), mcp.Description("Optional reasoning-effort override (gpt-6.1-terra's reasoning.effort vocabulary; 'none' disables reasoning); overrides workflow.audit.codex.effort (default xhigh). Transmitted on adversarial (turn/start) reviews.")),
		mcp.WithString("delta_base", mcp.Description(deltaBaseDescription)),
		projectRootOption(),
		mcp.WithOutputSchema[ReviewOutput](),
		// Catalog-WRITE tool (card t904): every call files an audit receipt
		// under the audited tree's .moai/state/audit-receipts/, on EVERY exit
		// path including the binary-missing fail-open — so the review need not
		// run for the write to happen. It read as a catalog-READ tool until the
		// receipt store landed (SPEC-CODEX-AUDIT-GATE-AXES-001) and the
		// annotation did not follow. Under codex's capability-based
		// `default_tools_approval_mode = "writes"` a read-only declaration
		// SKIPS the approval prompt, so the stale `true` suppressed the prompt
		// for a tool that writes — a miss, with no signal.
		mcp.WithReadOnlyHintAnnotation(false),
	), handleCodexAudit)

	// claude_audit → official Claude Code subscription CLI. The process receives
	// only a bounded diff on stdin, with built-in tools, plugins, hooks, MCP, and
	// session persistence disabled. Gateway routing credentials are scrubbed so
	// a GPT/GLM main session reaches first-party Claude instead of recursing.
	add(claudeAuditToolName, mcp.NewTool(
		claudeAuditToolName,
		mcp.WithDescription("Run an independent read-only Claude Code subscription audit over a bounded Git diff. Gateway routing is removed; missing subscription/auth/model capacity returns a structured inconclusive verdict."),
		mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), mcp.Description("Review target: uncommittedChanges or baseBranch.")),
		mcp.WithString("focus", mcp.Description("Optional independent-review focus area.")),
		mcp.WithString("model", mcp.Description("Optional Claude Code model alias or model id; overrides workflow.audit.claude.model.")),
		mcp.WithString("effort", mcp.Enum("low", "medium", "high", "xhigh", "max"), mcp.Description("Optional Claude Code effort; overrides workflow.audit.claude.effort.")),
		projectRootOption(),
		mcp.WithOutputSchema[ReviewOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleClaudeAudit)

	// codex_setup → Go probe (REQ-MCP-007 / AC-MCP-008): exec.LookPath("codex")
	// + codex --version + auth-provider classification + the enable_review_gate
	// toggle. NO Node bridge. Read-only.
	add("codex_setup", mcp.NewTool(
		"codex_setup",
		mcp.WithDescription("Probe the local codex installation: exec.LookPath + codex --version + auth-provider classification (ChatGPT / apiKey / provider / unknown) + the current enable_review_gate toggle state. Pure Go — no Node bridge."),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexSetup)

	// --- Codex read-only role launcher (codex_role_audit start/status/result) ---
	//
	// The MCP route to the same launcher core as `moai codex audit`: a read-only
	// contract role runs as one top-level read-only codex exec process in the
	// background, so the host's tool timeout never cuts an audit short.
	for _, rt := range codexRoleAuditTools() {
		add(rt.tool.Name, rt.tool, rt.handler)
	}

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
		requiredProjectRootOption(),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The task to give GLM.")),
		mcp.WithBoolean("background", mcp.Description("Run the task in the background and return a job id immediately. Background jobs live inside this server process and do not survive its exit.")),
		mcp.WithString("model", mcp.Description("Optional GLM model override; omitted ⇒ the backend default GLM model (glm_audit reads the workflow.audit.glm pin first).")),
		mcp.WithString("system", mcp.Description("Optional system prompt for the task.")),
		mcp.WithInteger("max_tokens", mcp.Description("Optional response token bound; defaults to the server-side glm_task ceiling.")),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleGLMTask)

	// glm_job_status → handleGLMJobStatus. Reads one record.
	add("glm_job_status", mcp.NewTool(
		"glm_job_status",
		mcp.WithDescription("Read a GLM job's record (status, timestamps, model, request summary). An unknown or unreadable job is a structured result, not a failed call."),
		mcp.WithString(glmJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background glm_task.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMJobStatus)

	// glm_job_result → handleGLMJobResult. Reads one record; never waits for
	// the call.
	add("glm_job_result", mcp.NewTool(
		"glm_job_result",
		mcp.WithDescription("Read a GLM job's output. A job in a terminal status returns its recorded output; a job still running returns its current status without blocking the caller."),
		mcp.WithString(glmJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background glm_task.")),
		projectRootOption(),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMJobResult)

	// glm_job_cancel → handleGLMJobCancel. NOT read-only: it cancels a running
	// job and records the cancellation. A job this server lifetime did not
	// start is refused rather than cancelled.
	add("glm_job_cancel", mcp.NewTool(
		"glm_job_cancel",
		mcp.WithDescription("Stop a running GLM job: records the cancellation and revokes the context its in-flight z.ai request runs under, then waits a bounded grace window for the job to end. A job this server lifetime did not start is refused rather than cancelled; an already-terminal job returns its status and sends nothing."),
		mcp.WithString(glmJobIDArg, mcp.Required(), mcp.Description("The job id returned by a background glm_task.")),
		requiredProjectRootOption(),
		mcp.WithReadOnlyHintAnnotation(false),
	), handleGLMJobCancel)

	// --- M3 GLM backend (design.md §3 M3) ---

	// glm_audit → z.ai GLM API direct call (REQ-MCP-009 / AC-MCP-011). Calls
	// the Anthropic-compatible https://api.z.ai/api/anthropic/v1/messages
	// endpoint directly — NOT the z.ai MCP server, NOT any gateway. Output
	// adopts the shared review-output.schema.json (§G.4). Fail-open on a
	// missing/unauthenticated/erroring/malformed GLM (VerdictInconclusive →
	// claude fallback, REQ-MCP-012). With no explicit model the
	// workflow.audit.glm pin applies, else the backend default GLM model.
	add("glm_audit", mcp.NewTool(
		"glm_audit",
		mcp.WithDescription("Run a GLM (z.ai) code review. Calls the z.ai Anthropic-compatible endpoint directly and returns a review-output schema (verdict/summary/findings/next_steps). GLM is OPTIONAL; a missing/unauthenticated/erroring GLM yields verdict 'inconclusive' (fail-open → fall back to the active auditor)."),
		mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), mcp.Description("What GLM reviews: 'uncommittedChanges' (default) or 'baseBranch'. The named change is collected as a diff and sent in the request — GLM has no filesystem, so a target that yields no diff returns 'inconclusive' rather than a verdict.")),
		mcp.WithString("focus", mcp.Description("Optional focus area (e.g. 'concurrency', 'auth', 'secret handling').")),
		mcp.WithString("model", mcp.Description("Optional GLM model override; omitted ⇒ the backend default GLM model (glm_audit reads the workflow.audit.glm pin first).")),
		projectRootOption(),
		mcp.WithOutputSchema[ReviewOutput](),
		// SPEC-CODEX-WIRING-001 (REQ-CW-011 / spec §A.4): catalog-READ tool —
		// see the audit_cache note.
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMAudit)

	// codex_review / glm_review → the on-demand self-review tools
	// (SPEC-CODEX-REVIEW-OWNERSHIP-001 REQ-CRO-007..010): any session asks codex
	// or GLM to review its own card diff or its uncommitted changes. They are
	// NOT audit tools — the result is advisory, files no audit receipt and is
	// never subject to the required audit gate — so both are catalog-READ tools
	// (no write to the reviewed tree or to .moai/state) and carry the read-only
	// hint, which keeps codex's `writes` approval mode from prompting for them.
	add("codex_review", mcp.NewTool(
		"codex_review",
		mcp.WithDescription("Ask codex to review YOUR OWN change. The result is ADVISORY (non-binding): it files no audit receipt, is never subject to a required audit gate, and cannot stand in for an audit verdict or for the leader's evidence read. scope=card reviews the card diff — resolved by the same scope resolver the turn-end review gate uses, with the merge base recomputed on every call; a tree that is not a card worktree returns 'inconclusive' without reviewing anything. scope=uncommitted reviews the uncommitted changes of the named tree (in the primary checkout that is the whole shared working tree; there is no path restriction). A missing codex, an error, or an empty change returns verdict 'inconclusive' (fail-open). Every result names its scope, base (the merge base SHA for scope=card), backend, and the canonical tree reviewed."),
		mcp.WithString(selfReviewScopeArg, mcp.Required(), mcp.Enum(selfReviewScopeCard, selfReviewScopeUncommitted), mcp.Description("REQUIRED, no default. 'card' = the card diff from the recomputed merge base; 'uncommitted' = the uncommitted changes of the named tree.")),
		mcp.WithString(selfReviewModelArg, mcp.Description("Optional model override; omitted ⇒ codex's own configured default. The workflow.audit.codex pin is NOT applied.")),
		projectRootOption(),
		mcp.WithOutputSchema[SelfReviewOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleCodexReview)

	add("glm_review", mcp.NewTool(
		"glm_review",
		mcp.WithDescription("Ask GLM (z.ai) to review YOUR OWN change. The result is ADVISORY (non-binding): it files no audit receipt, is never subject to a required audit gate, and cannot stand in for an audit verdict or for the leader's evidence read. GLM has no filesystem, so the change is collected as a diff and sent in the request: scope=card is the diff from the recomputed merge base (same resolver as the turn-end review gate; a non-card tree returns 'inconclusive' without a request), scope=uncommitted is the diff of the named tree against HEAD, staged changes included. Runtime-managed paths are excluded; untracked files cannot be part of a diff and are named in 'excluded_untracked'; 'truncated' is set when the diff was cut at the size cap. A missing key, an error, or an empty diff returns verdict 'inconclusive' (fail-open) without asking GLM for a verdict."),
		mcp.WithString(selfReviewScopeArg, mcp.Required(), mcp.Enum(selfReviewScopeCard, selfReviewScopeUncommitted), mcp.Description("REQUIRED, no default. 'card' = the card diff from the recomputed merge base; 'uncommitted' = the uncommitted changes of the named tree.")),
		mcp.WithString(selfReviewModelArg, mcp.Description("Optional GLM model override; omitted ⇒ the delegation default model. The workflow.audit.glm pin is NOT applied.")),
		projectRootOption(),
		mcp.WithOutputSchema[SelfReviewOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleGLMReview)

	// audit_multi → SPEC-AUDIT-MULTI-MODEL-001 multi-auditor convergence
	// (REQ-AMM-009 / REQ-AMM-010 / AC-AMM-012 / AC-AMM-013). Thin wrapper over
	// runMultiAudit (mcp_convergence.go) — does NOT re-implement the
	// Claude/codex/GLM backends (C1). A Claude-origin session may supply its
	// in-session claude_verdict anchor; GPT/GLM/unknown origins ignore it and run
	// an independent subscription-backed Claude audit. Gates are optional
	// per-auditor overrides (defaults: claude+codex required, glm advisory).
	add(auditMultiToolName, mcp.NewTool(
		auditMultiToolName,
		mcp.WithDescription("Run source-aware Claude/Codex/GLM audit convergence. Claude-origin sessions may reuse an in-session claude_verdict; GPT/GLM/unknown origins ignore caller Claude output and run a fresh subscription-backed Claude audit. Returns a ConvergenceResult with per-backend source and provenance."+worktreeStateDescNote),
		mcp.WithObject("claude_verdict", mcp.Description("Conditional Claude-main anchor (review-output schema). Ignored for GPT/GLM/unknown launch providers, which run an independent Claude subscription audit.")),
		mcp.WithString("target", mcp.Description("Optional review target (file path, diff ref, or scope label).")),
		mcp.WithString("focus", mcp.Description("Optional focus area (e.g. 'concurrency', 'auth', 'secret handling').")),
		mcp.WithString("codex_effort", mcp.Enum(config.ValidCodexAuditEfforts()...), mcp.Description("Optional per-call reasoning-effort override for the codex leg (gpt-6.1-terra's reasoning.effort vocabulary; 'none' disables reasoning); overrides the workflow.audit.codex.effort pin for THIS invocation only. Other backends are unaffected.")),
		mcp.WithObject("gates", mcp.Description("Optional per-auditor gate override (keys: claude/codex/glm; values: off|advisory|required). Defaults: claude+codex required, glm advisory.")),
		mcp.WithString("session_id", mcp.Description("Optional session id for per-session convergence state persistence (.moai/state/audit-multi/<session>.json). Empty ⇒ persistence no-op.")),
		// A4 (SPEC-AUTONOMY-CLOSURE-001 REQ-CLOSURE-012): binds this fan-out
		// to the card — one second-review record lands in the card evidence
		// directory. Empty ⇒ byte-identical pre-change behavior.
		mcp.WithString("card_id", mcp.Description("Optional card id binding this review as the card's second review (writes second-review.jsonl into the card evidence directory). Empty ⇒ no record.")),
		mcp.WithString("delta_base", mcp.Description(deltaBaseDescription)),
		// Pass-through semantics: this fan-out handed its backends no root
		// before the parameter existed, so an absent parameter must keep
		// handing them none rather than substituting a default.
		projectRootPassthroughOption(),
		// Catalog-WRITE tool (card t904): the fan-out files an audit receipt
		// like codex_audit does, and additionally persists its ConvergenceResult
		// to .moai/state/audit-multi/<session_id>.json whenever session_id is
		// supplied. Same stale-`true` history and same consequence as
		// codex_audit — see the note there.
		mcp.WithReadOnlyHintAnnotation(false),
	), handleAuditMulti)

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

	// SPEC-FACTORY-SELF-DISPATCH-001 M3 (REQ-SD-014): the four factory card
	// tools and the two todo tools, each a thin wrapper over the same
	// function its CLI counterpart calls (mcp_factory_card.go /
	// mcp_todo.go).
	registerFactoryCardMCPTools(add)
	registerTodoMCPTools(add)

	// --- jev_ask → jev.Client.Ask (internal/jev) — the gated judgment wrapper.
	// A thin caller over the ONE call path (no transport code here), inert
	// behind workflow.jev.enabled (shipped default false): gate off → no client
	// constructed, no network call, a gated-unavailable result. Registration is
	// unconditional so the tool is countable and its gate-off contract is
	// invocable; the per-tool mcp.yaml enablement map still applies above.
	registerJevAskTool(add)
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
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("session_list", err), nil
	}
	entries, err := session.QueryActiveWork(specID)
	if err != nil {
		return toolErr("session_list", err), nil
	}
	return toolJSONWithRoot("session_list", map[string]any{
		"spec_id":  specID,
		"count":    len(entries),
		"sessions": entries,
	}, root, source), nil
}

// handleGoalStatus wraps goal.LoadGoal (state.go:57).
func handleGoalStatus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := req.GetString("session_id", "")
	if sessionID == "" {
		sessionID = statusSessionID("") // inherit the CLI's read-path resolver
	}
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("goal_status", err), nil
	}
	g, err := goal.LoadGoal(root, sessionID)
	if err != nil {
		return toolErr("goal_status", err), nil
	}
	if g == nil {
		return toolJSONWithRoot("goal_status", map[string]any{"session_id": sessionID, "armed": false}, root, source), nil
	}
	return toolJSONWithRoot("goal_status", map[string]any{"session_id": sessionID, "armed": true, "goal": g}, root, source), nil
}

// handleGoalArm wraps the goal.NewGoal (schema.go:115) + goal.SaveGoal
// (state.go:76) composition — the SAME internal/goal core the `moai goal arm`
// CLI RunE composes. It inherits the CLI's core rule enforcement
// (parseCondition classification; the infinite-goal fail-closed: max_turns == 0
// requires max_duration > 0) WITHOUT re-implementing internal/goal logic. The
// CLI-only stderr warnings / flag UX are presentation, not core.
func handleGoalArm(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// REQ-MN-004: a mutating tool names its tree — the caller's project_root
	// is required, confined, and anchors the goal state the arm writes.
	root, source, err := resolveRequiredToolProjectRoot(req)
	if err != nil {
		return toolErr("goal_arm", err), nil
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
	// Inherited rule: a mechanical condition that can only ever fail buys a goal
	// that blocks every turn-end to the ceiling. This is the SAME function the
	// CLI arm path calls, not a parallel copy — the two arm surfaces disagreeing
	// about which conditions are armable is the defect class issue #1660 was
	// reported against, so the gate has exactly one implementation.
	if err := armTimeConditionGate(ctx, "goal_arm", conditionText, cond); err != nil {
		return toolErr("goal_arm", err), nil
	}
	g := goal.NewGoal(sessionID, conditionText, []goal.Condition{cond})
	if maxTurns >= 0 {
		g.Ceiling.MaxTurns = maxTurns
	}
	g.Ceiling.MaxDuration = maxDuration
	// REQ-MN-010: the goal state write is a project-state write — serialized
	// through the one server-wide write mutex.
	releaseGoal := mcpStateWriteLock()
	if err := goal.SaveGoal(root, g); err != nil {
		releaseGoal()
		return toolErr("goal_arm", err), nil
	}
	releaseGoal()
	return toolJSONWithRoot("goal_arm", map[string]any{
		"action":           "arm",
		"session_id":       sessionID,
		"condition_type":   string(cond.Type),
		"ceiling_maxturns": g.Ceiling.MaxTurns,
	}, root, source), nil
}

// handleSpecProgress wraps spec.ListDocs (listdocs.go:36) — the SPEC scanner.
func handleSpecProgress(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("spec_progress", err), nil
	}
	// A config-orphaned worktree answers over the union of its own and its
	// primary checkout's catalogue (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-011..015).
	if view := resolveCatalogueView(root); view.orphaned {
		docs, err := view.unionSpecDocs()
		if err != nil {
			return toolErr("spec_progress", err), nil
		}
		return toolJSON("spec_progress", map[string]any{
			"count":    len(docs),
			"specs":    docs,
			"shadowed": view.shadowedPrimaryCopies(),
			"_root":    view.rootBlock(source),
		}), nil
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
	// A config-orphaned worktree keeps its snapshots in its primary checkout's
	// store; with that primary unidentifiable there is no store and the call
	// fails rather than reading or writing another root
	// (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-002/004/006).
	store, err := auditreceipt.StoreRoot(root)
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
		// REQ-MN-010: the snapshot record is a project-state write — the one
		// server-wide write mutex serializes it against the other writers.
		release := mcpStateWriteLock()
		snap, err := verify.RecordCheck(store, key, entry)
		release()
		if err != nil {
			return toolErr("verify_snapshot", err), nil
		}
		return toolJSON("verify_snapshot", map[string]any{"action": "record", "snapshot": snap, "_root": stateRootBlock(root, source, store)}), nil
	}
	snap, err := verify.Load(store, key)
	if err != nil {
		return toolErr("verify_snapshot", err), nil
	}
	return toolJSON("verify_snapshot", map[string]any{"action": "load", "snapshot": snap, "_root": stateRootBlock(root, source, store)}), nil
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
	store, err := auditreceipt.StoreRoot(root) // the writer's store (REQ-WSR-006)
	if err != nil {
		return toolErr("verify_trend", err), nil
	}
	snap, err := verify.Load(store, key)
	if err != nil {
		return toolErr("verify_trend", err), nil
	}
	checks := []verify.CheckEntry{}
	if snap != nil {
		checks = snap.Checks
	}
	return toolJSON("verify_trend", map[string]any{"key": key, "checks": checks, "_root": stateRootBlock(root, source, store)}), nil
}

// handleSpecAudit wraps spec.Audit (audit.go:156).
func handleSpecAudit(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("spec_audit", err), nil
	}
	opts := spec.AuditOptions{
		BaseDir:              root,
		FilterSpec:           req.GetString("filter_spec", ""),
		FilterEra:            req.GetString("filter_era", ""),
		IncludeGrandfathered: req.GetBool("include_grandfathered", false),
	}
	if view := resolveCatalogueView(root); view.orphaned {
		merged, findings, err := view.unionAudit(opts)
		if err != nil {
			return toolErr("spec_audit", err), nil
		}
		out := toJSONMap(merged)
		out["drift_findings"] = findings
		out["shadowed"] = view.shadowedPrimaryCopies()
		out["_root"] = view.rootBlock(source)
		return toolJSON("spec_audit", out), nil
	}
	result, auditErr := spec.Audit(opts)
	if auditErr != nil {
		return toolErr("spec_audit", auditErr), nil
	}
	// spec_audit carries the _root block like the other catalogue tools; the
	// audit result's own fields keep their names and values
	// (SPEC-MCP-WORKTREE-UNTRACKED-001 REQ-MWU-013).
	return toolJSON("spec_audit", withRootBlock(result, rootProvenanceMap(root, source))), nil
}

// handleSpecDrift wraps spec.Audit (audit.go:156), filtered to modern-era drift.
func handleSpecDrift(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	root, source, err := resolveToolProjectRootWithSource(req)
	if err != nil {
		return toolErr("spec_drift", err), nil
	}
	if view := resolveCatalogueView(root); view.orphaned {
		merged, findings, err := view.unionAudit(spec.AuditOptions{})
		if err != nil {
			return toolErr("spec_drift", err), nil
		}
		return toolJSON("spec_drift", map[string]any{
			"total_specs":    merged.TotalSpecs,
			"modern_clean":   merged.ModernEraClean,
			"drift_findings": findings,
			"shadowed":       view.shadowedPrimaryCopies(),
			"_root":          view.rootBlock(source),
		}), nil
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

// stateAuditCacheFor binds the disk-persisted store to the call's project root
// (REQ-MN-020). The project root was resolved and validated by the caller.
func stateAuditCacheFor(root string) *runtime.StateAuditCache {
	return runtime.NewStateAuditCache(root)
}

// handleAuditCache wraps runtime.AuditCache ComputeHash/Lookup/Store over the
// process-shared moaiAuditCache. The store op lets a caller record a PASS
// verdict that a later lookup (same moai mcp-server process) returns as a hit.
func handleAuditCache(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	op := req.GetString("op", "")
	resolveRoot := resolveToolProjectRootWithSource
	if op == "store" {
		resolveRoot = resolveRequiredToolProjectRoot
	}
	root, source, rootErr := resolveRoot(req)
	if rootErr != nil {
		return toolErr("audit_cache", rootErr), nil
	}
	// REQ-MN-020: the store is the project's own — entries persist under the
	// named tree's state root, carrying verdict-file evidence.
	cache := stateAuditCacheFor(root)
	// REQ-MN-009: under the gate exposure audit_cache is the READ member —
	// its store operation is one of the mutating operations the exclusion
	// names, and a gate session's server refuses it.
	if op == "store" && moaiCurrentExposure == moaiMCPExposureGate {
		return toolErr("audit_cache", fmt.Errorf("op=store is not available to gate sessions")), nil
	}
	switch op {
	case "compute_hash":
		specDir := req.GetString("spec_dir", "")
		if specDir == "" {
			return toolErr("audit_cache", fmt.Errorf("op=compute_hash requires spec_dir")), nil
		}
		// A RELATIVE spec_dir names the validated root, not the server's
		// working directory (sync-audit r3 F3): a server whose cwd sits in
		// another tree must not return that tree's plan hash while "_root"
		// claims the requested one — the same anchoring the store applies to
		// a relative report_path.
		if !filepath.IsAbs(specDir) {
			specDir = filepath.Join(root, specDir)
		}
		hash, err := cache.ComputeHash(specDir)
		if err != nil {
			return toolErr("audit_cache", err), nil
		}
		return toolJSONWithRoot("audit_cache", map[string]any{"op": op, "hash": hash}, root, source), nil
	case "lookup":
		specID := req.GetString("spec_id", "")
		hash := req.GetString("hash", "")
		entry, ok := cache.Lookup(specID, hash, time.Now())
		return toolJSONWithRoot("audit_cache", map[string]any{"op": op, "hit": ok, "entry": entry}, root, source), nil
	case "store":
		// Record a PASS verdict whose evidence is an existing verdict file
		// (REQ-MN-020): a store whose report file is missing or unreadable is
		// an error, never a stored entry.
		specID := req.GetString("spec_id", "")
		hash := req.GetString("hash", "")
		if specID == "" || hash == "" {
			return toolErr("audit_cache", fmt.Errorf("op=store requires spec_id and hash")), nil
		}
		// REQ-MN-010: the cache store is a project-state write — serialized
		// through the one server-wide write mutex.
		releaseCache := mcpStateWriteLock()
		storeErr := cache.StoreVerdict(specID, hash, &runtime.AuditResult{
			Verdict:        runtime.VerdictPass,
			AuditAt:        time.Now(),
			AuditorVersion: req.GetString("auditor_version", ""),
			ReportPath:     req.GetString("report_path", ""),
		})
		releaseCache()
		if storeErr != nil {
			return toolErr("audit_cache", storeErr), nil
		}
		return toolJSONWithRoot("audit_cache", map[string]any{"op": op, "stored": true, "spec_id": specID, "hash": hash}, root, source), nil
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

// ─── project-state write serialization (REQ-MN-010) ───────────────────────────

// mcpStateWriteMu serializes the project-state writes the tool handlers
// perform under the caller's .moai/state — the goal state, the verification
// snapshots, and the audit cache. Concurrent tool calls on one server hold
// this ONE mutex across their read-modify-write sections, so interleaved
// writers cannot corrupt or double-write the same artifact (REQ-MN-010).
var mcpStateWriteMu sync.Mutex

// mcpStateWriteLock takes the project-state write mutex and returns the
// release func — the handlers' defer shape.
func mcpStateWriteLock() func() {
	mcpStateWriteMu.Lock()
	return mcpStateWriteMu.Unlock
}

// ─── result-shaping helpers (structured results, never AskUserQuestion) ───────

// toolJSON shapes a successful structured result whose TEXT block carries the
// same JSON as the structured content (SPEC-MCP-NEUTRALITY-001 REQ-MN-008
// wire honesty): the old "<tool>: ok" fallback told a text-reading caller
// nothing while the payload sat in the other block.
func toolJSON(tool string, data any) *mcp.CallToolResult {
	body, err := json.Marshal(data)
	if err != nil {
		// The payload is not marshalable — the text block degrades to the
		// old fallback rather than lying with a JSON-looking string.
		return mcp.NewToolResultStructured(data, tool+": ok")
	}
	return mcp.NewToolResultStructured(data, string(body))
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
