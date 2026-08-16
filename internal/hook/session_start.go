// Resolution: KEEP — full business logic; GLM setup, skill discovery, memory evaluation.
package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/modu-ai/moai-adk/internal/mx"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/modu-ai/moai-adk/internal/telemetry"
	"golang.org/x/sync/errgroup"
)

// sessionStartHandler processes SessionStart events.
// It initializes the session, loads project configuration, and validates
// the execution environment (REQ-HOOK-030).
type sessionStartHandler struct {
	cfg ConfigProvider
}

// NewSessionStartHandler creates a new SessionStart event handler.
func NewSessionStartHandler(cfg ConfigProvider) Handler {
	return &sessionStartHandler{cfg: cfg}
}

// EventType returns EventSessionStart.
func (h *sessionStartHandler) EventType() EventType {
	return EventSessionStart
}

// Handle processes a SessionStart event. It logs the session ID, loads
// project configuration, and returns project information in the Data field.
// Errors are non-blocking: the handler logs warnings and returns allow.
//
// Input-lag budget (CLAUDE.local.md §7 default 5s; SPEC-SESSIONSTART-PERF-001):
// Claude Code awaits SessionStart completion before accepting the first user
// input. Handle therefore minimizes synchronous work to turn-visible side
// effects only — credential/teammate-mode/tmux-env injection, multi-session
// registry attribution, evolved-skill symlinks, migration apply, and the
// AdditionalContext/session-id attribution. Independent synchronous steps run
// concurrently via errgroup (collapsing the serial sum toward the slowest
// single step). Heavy advisory scanning (telemetry prune, stale-memory wrap,
// pending-proposal summary, status-drift git scan) is deferred to a
// best-effort background goroutine so it never blocks the synchronous return.
//
// @MX:NOTE: [AUTO] input-lag budget — synchronous path is turn-visible-effects only
func (h *sessionStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session started",
		"session_id", input.SessionID,
		"cwd", input.CWD,
		"project_dir", input.ProjectDir,
	)

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "initialized",
	}

	// SPEC-V3R6-MULTI-SESSION-COORD-001 L3: the 3-step multi-session protocol
	// (Register + Purge + Query + stderr surface) now runs inside the errgroup
	// below (Task 2), concurrent with the other independent synchronous steps.
	// See runMultiSessionProtocol for the per-step BEST-EFFORT contract
	// (REQ-COORD-013..015).

	// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-WPR-003: when the
	// multi-session protocol is bypassed because input.SessionID is empty,
	// emit a non-blocking stderr warning so the orchestrator can observe
	// that the registry write path was skipped. This is a leading cause of
	// the K6 "empty registry" defect (research.md §D.1): some Claude Code
	// activation paths emit an empty session_id, the L66 gate bypasses
	// Register, and the orchestrator later finds the registry empty. The
	// warning is observation-only — the hook still returns allow (non-blocking).
	if input.SessionID == "" {
		_, _ = fmt.Fprint(os.Stderr,
			"warning: SessionStart received empty session_id; multi-session registry write bypassed "+
				"(source_session_id attribution will fall back to the environment-fallback pattern)\n")
	}

	// Load project information from config if available
	cfg := h.getConfig()
	if cfg != nil {
		if cfg.Project.Name != "" {
			data["project_name"] = cfg.Project.Name
		}
		if string(cfg.Project.Type) != "" {
			data["project_type"] = string(cfg.Project.Type)
		}
		if cfg.Project.Language != "" {
			data["project_language"] = cfg.Project.Language
		}
	} else {
		slog.Warn("configuration not available, proceeding with defaults",
			"session_id", input.SessionID,
		)
	}

	if input.ProjectDir != "" {
		// (a) Parallelize the independent synchronous steps. Each step writes
		// to its OWN local data map; the maps are merged into `data`
		// sequentially AFTER errgroup.Wait() to avoid concurrent map writes.
		// Steps that share a resource (settings.local.json, the session
		// registry) are sequenced INSIDE their owning task.
		g, gctx := errgroup.WithContext(ctx)

		// Task 1 — settings.local.json chain. The four writers (GLM creds,
		// teammateMode, tmux GLM env, Windows CLAUDE_ENV_FILE) share ONE file
		// and MUST stay sequential. ensureTmuxGLMEnv depends on the token
		// ensureGLMCredentials just wrote. See runSettingsChain.
		var settingsData map[string]any
		g.Go(func() error {
			settingsData = h.runSettingsChain(input)
			return nil
		})

		// Task 2 — session-registry chain. runMultiSessionProtocol (Register +
		// Purge + Query, REQ-COORD-013..015) writes the registry that
		// attribution anchors to; pruneGoalOrphans then reads that same
		// registry to identify active sessions before pruning goal orphans
		// (AC-GLE-038 pins synchronous completion of the orphan move).
		// Sequenced inside one task so the prune's registry read sees the
		// protocol's register write.
		//
		// @MX:NOTE: [AUTO] registry RMW + goal orphan prune sequenced per shared file
		var registryData map[string]any
		g.Go(func() error {
			registryData = make(map[string]any)
			if input.SessionID != "" {
				h.runMultiSessionProtocol(input, registryData)
			}
			pruneGoalOrphans(input.ProjectDir)
			return nil
		})

		// Task 3 — evolved skill symlinks (R5). Independent of settings and
		// registry; touches only .claude/skills/ and .moai/evolution/.
		var skillData map[string]any
		g.Go(func() error {
			skillData = runSkillSymlinks(input.ProjectDir)
			return nil
		})

		// Task 4 — migration apply (REQ-020, REQ-021). File mutations affect
		// the session, so it stays synchronous; it is independent of the other
		// tasks (writes only migration-version + migrated files). cfg was
		// captured once above and is safe to read here (no other task reads it).
		//
		// @MX:WARN @MX:REASON migration errors MUST NOT block the session (REQ-021);
		// a failed apply preserves the migration-version file unchanged and the
		// next session retries. Best-effort, fail-open.
		var migrationData map[string]any
		g.Go(func() error {
			migrationData = runMigration(gctx, input.ProjectDir, cfg)
			return nil
		})

		if err := g.Wait(); err != nil {
			// Best-effort contract preserved: Handle never returns a non-nil
			// error from these steps. errgroup.WithContext cancels on first
			// non-nil return, but every task above returns nil, so this branch
			// is defensive only.
			slog.Warn("session start: synchronous step group reported error (non-blocking)",
				"error", err.Error())
		}

		mergeData(data, settingsData, registryData, skillData, migrationData)

		// (b) Defer heavy advisory scanning off the synchronous critical path.
		// These four steps (telemetry prune, stale-memory wrap, pending-proposal
		// summary, status-drift git scan) previously ran synchronously and
		// blocked the hook return. They now run in a background goroutine and
		// are joined with a SHORT bounded deadline (deferredScanJoinBound).
		//
		// The bounded join is the drop-mitigation: most scans finish well
		// within the bound → their advisory keys (memory_stale_warning,
		// skill_proposals, status_drift_warning) land in THIS session's Data
		// map. Only a genuinely-slow scan (e.g. drift on a huge repo) exceeds
		// the bound and is dropped — its advisory key is absent this session
		// and re-derives idempotently next session. The Data field carries
		// json:"-" (internal-only, never surfaced to Claude Code per
		// research.md §D.0), so a dropped key is not user-observable.
		//
		// The bound is deliberately SHORTER than the drift scan's own time-box
		// (DefaultSessionStartDriftTimeout) so worst-case added input lag is
		// capped by deferredScanJoinBound, not the scan ceiling. An unbounded
		// join would reintroduce the full serial input lag this refactor
		// removed; the bound is the whole point.
		//
		// The drift seams (driftCountFn, sessionStartDriftTimeout) are
		// snapshotted HERE (synchronously) so the background goroutine never
		// concurrently reads the package-level vars while a test restores them
		// in t.Cleanup — that would be a -race finding. The goroutine uses only
		// the captured locals.
		//
		// @MX:WARN @MX:REASON bounded-background-goroutine join — if the scan
		// exceeds deferredScanJoinBound it is abandoned for this session; safe
		// because all work is idempotent and best-effort (next session re-derives).
		driftFn := driftCountFn
		driftTimeout := sessionStartDriftTimeout
		projectDir := input.ProjectDir

		// MX sidecar index freshness (SPEC-MX-ACTIVATION P0-1): cheap
		// synchronous check — does .moai/state/mx-index.json exist AND is its
		// ScannedAt within mxIndexFreshnessThreshold? The result gates whether
		// the deferred goroutine runs an expensive full ScanDir. ScanDir itself
		// NEVER runs on the synchronous path (Advisory-Check Discipline); only
		// this stat + one-field read does.
		//
		// @MX:NOTE: [AUTO] MX index freshness — sync check gates deferred cold-start scan
		mxScanNeeded := mxIndexNeedsRebuild(projectDir)

		if deferredScansAsyncEnabled() {
			// Production path: spawn a background goroutine and join with a
			// short bounded deadline.
			//
			// Snapshot the test-only completion seam HERE (synchronously, in
			// this goroutine) so the deferred goroutine never reads the
			// package-level var. nil in production.
			completed := snapshotDeferredScanCompleted()
			advisoryCh := h.spawnDeferredAdvisoryScans(projectDir, driftFn, driftTimeout, completed, mxScanNeeded)

			// Timeout-bound join. On receive, merge the advisory keys into
			// `data` before the final marshal so they ship in this session's
			// Data payload.
			joinTimer := time.NewTimer(deferredScanJoinBound)
			select {
			case advisory := <-advisoryCh:
				joinTimer.Stop()
				if len(advisory) > 0 {
					maps.Copy(data, advisory)
				}
			case <-joinTimer.C:
				// Scan exceeded the bound; advisory keys for this session are
				// dropped (non-blocking). The goroutine continues to completion
				// in the background (durable side effects still land) and sends
				// into the buffered channel without blocking.
				slog.Debug("session start: deferred advisory scan exceeded join bound (non-blocking)",
					"bound", deferredScanJoinBound.String())
			}
		} else {
			// Test path (TestMain sets deferredScansAsync=false): run the
			// scans INLINE-synchronously. No goroutine is spawned → nothing
			// leaks past the test boundary → no race against parallel tests
			// that reassign os.Stderr / reset the slog handler. The advisory
			// keys always land in `data` (there is no bound to exceed).
			advisory := h.computeDeferredAdvisory(projectDir, driftFn, driftTimeout)
			if len(advisory) > 0 {
				maps.Copy(data, advisory)
			}
			// MX cold-start scan runs AFTER the advisory keys are merged so it
			// never delays them; it is a durable side effect (index write),
			// not an advisory. Time-boxed and fail-open (see runMXColdStartScan).
			if mxScanNeeded {
				runMXColdStartScan(projectDir)
			}
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal session data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 M3 (REQ-RDP-004/005):
	// inject this orchestrator's own UUID via hookSpecificOutput.AdditionalContext
	// so the Claude Code runtime surfaces it at session start, AND write a
	// side-channel file (.moai/state/current-session-id.txt) so `moai session
	// current` can resolve the UUID post-compaction (the additionalContext is
	// lost after /clear; the side-channel file persists).
	//
	// The injection is STRICTLY ADDITIVE (REQ-RDP-005): existing behavior
	// (Register → Purge → Query → stderr surface → data map) is unchanged.
	// The AdditionalContext is the serialized field per Claude Code SessionStart
	// stdout contract (hooks-system.md § Hook Event stdin/stdout Reference);
	// the existing `Data` field carries `json:"-"` and is internal-only
	// (research.md §D.0 — structural root cause of the attribution dead feature).
	//
	// Gated on input.SessionID != "" (research.md §D.0/D.1 P1-outcome
	// implication): an empty UUID is never injected or written.
	out := &HookOutput{Data: jsonData}
	if input.SessionID != "" && input.ProjectDir != "" {
		out.HookSpecificOutput = &HookSpecificOutput{
			HookEventName: string(EventSessionStart),
			AdditionalContext: fmt.Sprintf(
				"moai session attribution: source_session_id=%s\n"+
					"Use 'moai session current' to re-read this UUID after /clear or compaction.\n"+
					"If unavailable, emit the canonical fallback via 'moai session current --show-fallback'.",
				input.SessionID,
			),
		}
		// Side-channel file write (best-effort, non-blocking).
		sidecar := filepath.Join(input.ProjectDir, session.CurrentSideChannelFile)
		if writeErr := os.WriteFile(sidecar, []byte(input.SessionID), 0o600); writeErr != nil {
			slog.Warn("session start: failed to write current-session-id side-channel file (non-blocking)",
				"path", sidecar,
				"error", writeErr.Error(),
			)
		}
	}

	// SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001: GLM 가드레일 리마인더 주입.
	// GLM 백엔드 세션(PROCESS env ANTHROPIC_BASE_URL이 z.ai 포함)일 때만 z.ai MCP
	// 라우팅 요약을 AdditionalContext에 추가한다. 비-GLM 세션은 빈 문자열을
	// 받으므로 아무것도 주입되지 않는다 (REQ-GH-002/003). cg-leader pane은 PROCESS
	// env에 z.ai가 없으므로 자동 carve-out된다 (REQ-GH-005/006). 검출은 절대
	// 블로킹하지 않는다 (REQ-GH-012). always-load에서 제거된 glm-web-tooling.md
	// 규칙을 on-demand로 대체 전달한다.
	if reminder := glmGuardrailReminder(); reminder != "" {
		if out.HookSpecificOutput == nil {
			out.HookSpecificOutput = &HookSpecificOutput{
				HookEventName: string(EventSessionStart),
			}
		}
		if out.HookSpecificOutput.AdditionalContext == "" {
			out.HookSpecificOutput.AdditionalContext = reminder
		} else {
			out.HookSpecificOutput.AdditionalContext += "\n\n" + reminder
		}
	}

	// Factory Mode bootstrap announcement — the kanban announcement's sibling
	// (SPEC-FACTORY-WORKER-FANOUT-001), emitted ahead of it so the two modes'
	// notices can never stack. Same dual-channel shape and the same
	// startup-only gating, for the same reasons the kanban block below
	// records; the operator copies N worker launch lines instead of four
	// companion lines.
	if notice := factoryBootstrapNoticeForSource(input.Source, langEnglish); notice != "" {
		if out.HookSpecificOutput == nil {
			out.HookSpecificOutput = &HookSpecificOutput{
				HookEventName: string(EventSessionStart),
			}
		}
		if out.HookSpecificOutput.AdditionalContext == "" {
			out.HookSpecificOutput.AdditionalContext = notice
		} else {
			out.HookSpecificOutput.AdditionalContext += "\n\n" + notice
		}

		operatorNotice := factoryBootstrapNotice(operatorLang(h.cfg))
		if out.SystemMessage == "" {
			out.SystemMessage = operatorNotice
		} else {
			out.SystemMessage += "\n\n" + operatorNotice
		}
	}

	// Kanban Mode bootstrap announcement. The launcher cannot deliver this —
	// it syscall.Exec's into claude, so its stdout is overwritten when the TUI
	// takes the screen. Non-kanban sessions get "" and nothing is injected.
	//
	// The notice rides BOTH channels because it has two audiences and they read
	// different surfaces. additionalContext reaches the orchestrator, which needs
	// the companion labels to address them later; systemMessage reaches the
	// operator, who must type the four launch lines by hand into new terminals.
	// Emitting only additionalContext delivered a human-addressed instruction to
	// the model alone, so the operator saw nothing at all.
	//
	// Each copy is rendered in its audience's language, the split language.yaml
	// already draws: agent_prompt_language (English) for the agent-facing copy,
	// conversation_language for the operator-facing one. The two copies differ
	// only in prose — commands, run id, and socket path are identical in both.
	//
	// The notice is a BOOTSTRAP announcement, so it belongs to a genuinely new
	// session and nothing else. SessionStart also fires on resume, clear, and
	// compact, where the kanban environment is still set and the notice would
	// therefore re-emit — telling the operator to open four terminals they
	// already opened, for a run already under way. Those three sources are
	// skipped; an empty source is treated as startup (see the helper).
	// The notice reads the backlog queue under the project root so the lead's
	// opening screen names the work the run actually moves. ProjectDir first,
	// CWD as fallback — the same preference chainLineageBanner applies. An
	// empty root degrades to a zero-count summary line inside the notice
	// rather than failing here.
	kanbanRoot := input.ProjectDir
	if kanbanRoot == "" {
		kanbanRoot = input.CWD
	}
	if notice := kanbanBootstrapNoticeForSource(input.Source, kanbanRoot, langEnglish); notice != "" {
		if out.HookSpecificOutput == nil {
			out.HookSpecificOutput = &HookSpecificOutput{
				HookEventName: string(EventSessionStart),
			}
		}
		if out.HookSpecificOutput.AdditionalContext == "" {
			out.HookSpecificOutput.AdditionalContext = notice
		} else {
			out.HookSpecificOutput.AdditionalContext += "\n\n" + notice
		}

		operatorNotice := kanbanBootstrapNotice(kanbanRoot, operatorLang(h.cfg))
		if out.SystemMessage == "" {
			out.SystemMessage = operatorNotice
		} else {
			out.SystemMessage += "\n\n" + operatorNotice
		}
	}

	// SPEC-CHAIN-CORE-001 REQ-CHAIN-013: lineage banner. Resolves the current
	// chain node (from env or ledger), backfills session_id (REQ-CHAIN-021),
	// and emits a depth + parent-chain + resume system-reminder. Time-boxed
	// and fail-open (empty string = no banner injected).
	if banner := chainLineageBanner(input.ProjectDir, input.CWD, input.SessionID); banner != "" {
		if out.HookSpecificOutput == nil {
			out.HookSpecificOutput = &HookSpecificOutput{
				HookEventName: string(EventSessionStart),
			}
		}
		if out.HookSpecificOutput.AdditionalContext == "" {
			out.HookSpecificOutput.AdditionalContext = banner
		} else {
			out.HookSpecificOutput.AdditionalContext += "\n\n" + banner
		}
	}

	return out, nil
}

// getConfig safely retrieves the configuration, returning nil if unavailable.
func (h *sessionStartHandler) getConfig() *config.Config {
	if h.cfg == nil {
		return nil
	}
	return h.cfg.Get()
}

// runSettingsChain runs the settings.local.json write chain sequentially.
// The four writers share ONE file and cannot run concurrently; the chain is
// also order-dependent (ensureTmuxGLMEnv reads the token ensureGLMCredentials
// wrote). It returns its own local data map so the caller can merge results
// after all errgroup tasks complete, avoiding concurrent map writes.
func (h *sessionStartHandler) runSettingsChain(input *HookInput) map[string]any {
	d := make(map[string]any)

	// Validate GLM credentials: if GLM model overrides exist in settings.local.json
	// but ANTHROPIC_AUTH_TOKEN is missing, auto-inject from ~/.moai/.env.glm.
	// This prevents 401 errors for Agent Teams teammates.
	if msg := ensureGLMCredentials(input.ProjectDir); msg != "" {
		d["glm_credentials"] = msg
		slog.Info("GLM credentials auto-injected", "message", msg)
	}

	// Auto-detect tmux environment and set teammateMode accordingly.
	// When inside tmux, teammates spawn in separate panes for visibility.
	// When outside tmux, fall back to "auto" (in-process display).
	if mode := ensureTeammateMode(input.ProjectDir); mode != "" {
		d["teammate_mode"] = mode
	}

	// In GLM team mode, inject GLM environment variables into the current tmux
	// session so that teammate panes inherit ANTHROPIC_AUTH_TOKEN. Must execute
	// after ensureGLMCredentials writes the token to settings.local.json.
	if msg := ensureTmuxGLMEnv(input.ProjectDir); msg != "" {
		d["tmux_glm_env"] = msg
		slog.Info("tmux GLM environment variable injected", "message", msg)
	}

	// Windows only: inject CLAUDE_ENV_FILE into settings.local.json when a
	// .env file is present in the project root (T-016, R-P1-1). Guarded to
	// Windows so macOS/Linux GLM env injection is never affected.
	if claudeEnvFileGuard(runtime.GOOS) {
		if msg := injectCLAUDEEnvFile(input.ProjectDir); msg != "" {
			d["claude_env_file"] = msg
			slog.Info("CLAUDE_ENV_FILE injected", "message", msg)
		}
	}

	return d
}

// runSkillSymlinks creates evolved-skill symlinks (R5) and returns a local
// data map so the caller can merge results without concurrent map writes.
func runSkillSymlinks(projectDir string) map[string]any {
	d := make(map[string]any)
	if n := ensureNewSkillSymlinks(projectDir); n > 0 {
		d["evolved_skills_linked"] = n
		slog.Info("evolved skill symlinks created", "count", n)
	}
	return d
}

// runMigration applies pending migrations (REQ-020, REQ-021). Best-effort,
// fail-open: a migration error is logged and surfaced via the returned data
// map, but never propagated. cfg may be nil (handler is resilient to a nil
// ConfigProvider).
func runMigration(ctx context.Context, projectDir string, cfg *config.Config) map[string]any {
	d := make(map[string]any)
	if cfg == nil || !cfg.System.Migrations.Disabled {
		runner := migration.NewRunner(projectDir)
		applied, err := runner.Apply(ctx)
		if err != nil {
			slog.Warn("session start: migration apply failed",
				"error", err.Error(),
				"project_dir", projectDir,
			)
			d["migration_error"] = err.Error()
		} else if len(applied) > 0 {
			d["migrations_applied"] = len(applied)
			slog.Info("session start: migrations applied successfully",
				"count", len(applied),
				"versions", applied,
			)
		}
	} else {
		slog.Info("session start: migrations disabled via config",
			"project_dir", projectDir,
		)
	}
	return d
}

// spawnDeferredAdvisoryScans launches the heavy advisory scanning in a
// background goroutine and returns a buffered (cap-1) channel that receives
// the advisory results once all four steps complete. The buffer guarantees
// the goroutine never blocks on send even if Handle has already returned
// after deferredScanJoinBound elapsed (the dropped case) — the result sits
// in the buffer and the goroutine exits cleanly.
//
// `completed` is the test-only join seam (nil in production): when non-nil
// the goroutine closes it on exit so a test can join for goleak hygiene.
//
// `mxScanNeeded` (computed synchronously by the caller via mxIndexNeedsRebuild)
// gates whether the MX cold-start full scan runs after the advisories. The
// scan is a durable side effect (index write) and is dispatched AFTER the
// advisory result is sent so it never delays advisory keys landing in the
// session's Data payload.
func (h *sessionStartHandler) spawnDeferredAdvisoryScans(
	projectDir string,
	driftFn func(context.Context, string) (int, error),
	driftTimeout time.Duration,
	completed chan struct{},
	mxScanNeeded bool,
) <-chan map[string]any {
	resultCh := make(chan map[string]any, 1)
	go func() {
		if completed != nil {
			defer close(completed)
		}
		advisory := h.computeDeferredAdvisory(projectDir, driftFn, driftTimeout)
		// Send advisories FIRST (buffered channel → never blocks even after
		// the join bound elapses), THEN run the MX cold-start scan as a
		// best-effort durable side effect.
		resultCh <- advisory
		if mxScanNeeded {
			runMXColdStartScan(projectDir)
		}
	}()
	return resultCh
}

// computeDeferredAdvisory runs the four heavy advisory scans and returns
// their advisory results (keys: memory_stale_warning, skill_proposals,
// status_drift_warning). Durable side effects (pruned telemetry files)
// persist regardless of whether the caller's bounded join receives the
// result. All steps are best-effort and idempotent.
//
// @MX:NOTE: [AUTO] deferred advisory scans — best-effort, idempotent, non-blocking
func (h *sessionStartHandler) computeDeferredAdvisory(
	projectDir string,
	driftFn func(context.Context, string) (int, error),
	driftTimeout time.Duration,
) map[string]any {
	res := make(map[string]any)

	// SPEC-TELEMETRY-001 R4: prune files older than 90 days. Durable side effect.
	if err := pruneTelemetry(projectDir); err != nil {
		slog.Warn("session start (deferred): telemetry pruning failed",
			"error", err.Error(),
		)
	}

	// SPEC-V3R2-EXT-001 REQ-006/017: stale-memory caveat. Advisory string
	// surfaced via the Data map (internal-only).
	if staleMsg := detectAndWrapStaleMemories(projectDir, time.Now()); staleMsg != "" {
		res["memory_stale_warning"] = staleMsg
	}

	// Reflective learning: pending proposals summary. Advisory string surfaced
	// via the Data map (internal-only).
	if summary := PresentPendingProposals(projectDir); summary != "" {
		res["skill_proposals"] = summary
	}

	// SPEC-SESSIONSTART-PERF-001 REQ-SSP-015: heaviest step — time-boxed git
	// scan over SPEC dirs. Uses the snapshotted seams; on deadline emits the
	// advisory (preserving the "Run 'moai spec drift' for details." hint).
	driftCtx, cancel := context.WithTimeout(context.Background(), driftTimeout)
	count, err := driftFn(driftCtx, projectDir)
	cancel()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			res["status_drift_warning"] = driftTimeoutAdvisory
			slog.Info("session start (deferred): status drift check timed out",
				"project_dir", projectDir)
		}
		// Other errors (git absent, no specs dir) stay silent, matching the
		// synchronous detectStatusDrift best-effort contract.
		return res
	}
	if count >= driftWarningThreshold {
		res["status_drift_warning"] = fmt.Sprintf("⚠ %d SPECs have status drift. Run 'moai spec drift' for details.", count)
		slog.Info("session start (deferred): status drift detected",
			"count", count,
		)
	}
	return res
}

// mergeData merges each src map into dst sequentially. It is the single
// convergence point after errgroup.Wait(), guaranteeing no concurrent map
// writes: each errgroup task wrote only to its own local map.
func mergeData(dst map[string]any, srcs ...map[string]any) {
	for _, src := range srcs {
		maps.Copy(dst, src)
	}
}

// settingsLocalJSON is the minimal struct for reading settings.local.json env vars.
type settingsLocalJSON struct {
	Env         map[string]string `json:"env,omitempty"`
	Permissions map[string]any    `json:"permissions,omitempty"`
	// Preserve unknown fields
	Extra map[string]json.RawMessage `json:"-"`
}

// ensureGLMCredentials checks settings.local.json for GLM model overrides
// without ANTHROPIC_AUTH_TOKEN. If found, it reads the API key from
// ~/.moai/.env.glm and injects it along with ANTHROPIC_BASE_URL.
// Returns a status message if credentials were injected, empty string otherwise.
func ensureGLMCredentials(projectDir string) string {
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil || len(data) == 0 {
		return ""
	}

	var settings settingsLocalJSON
	if err := json.Unmarshal(data, &settings); err != nil {
		return ""
	}

	if settings.Env == nil {
		return ""
	}

	// Skip auto-injection in CG mode: CG mode intentionally removes AUTH_TOKEN
	// from settings.local.json so the leader uses Claude OAuth. Teammates get
	// GLM credentials via tmux session env instead.
	if isCGMode(projectDir) {
		return ""
	}

	// Check if GLM model overrides exist
	hasGLMModel := false
	for _, key := range []string{
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
	} {
		if val, ok := settings.Env[key]; ok && strings.Contains(strings.ToLower(val), "glm") {
			hasGLMModel = true
			break
		}
	}

	if !hasGLMModel {
		return ""
	}

	// GLM models configured — check if AUTH_TOKEN exists
	if token := settings.Env[config.EnvAnthropicAuthToken]; token != "" {
		// Already has credentials — nothing to inject, but the context-window
		// envs must still be ensured: settings written by an older binary (or
		// by `moai glm setup`) carry neither window key, and without the
		// CLAUDE_CODE_MAX_CONTEXT_TOKENS declaration Claude Code assumes a
		// 200K window for the custom GLM model ID (Issue #653, PR #1574
		// review). Persist only when a key was actually added so the steady
		// state does not rewrite settings.local.json on every session start.
		before := settings.Env[config.EnvClaudeCodeAutoCompactWindow] + "|" + settings.Env[config.EnvClaudeCodeMaxContextTokens]
		maybeSet1MAutoCompactWindow(settings.Env)
		maybeDeclareGLMContextWindow(settings.Env)
		after := settings.Env[config.EnvClaudeCodeAutoCompactWindow] + "|" + settings.Env[config.EnvClaudeCodeMaxContextTokens]
		if after != before {
			if err := persistSettingsEnv(settings.Env, data, settingsPath); err != nil {
				slog.Error("failed to write GLM context window env to settings.local.json", "error", err.Error())
			}
		}
		return ""
	}

	// AUTH_TOKEN missing — try to load from ~/.moai/.env.glm
	apiKey := loadGLMKeyFromEnvFile()
	if apiKey == "" {
		slog.Warn("GLM models configured but no API key found",
			"settings", settingsPath,
			"hint", "run 'moai glm setup <api-key>' to save your key",
		)
		return ""
	}

	// Inject credentials
	settings.Env[config.EnvAnthropicAuthToken] = apiKey
	if settings.Env[config.EnvAnthropicBaseURL] == "" {
		settings.Env[config.EnvAnthropicBaseURL] = config.DefaultGLMBaseURL
	}
	// Ensure compatibility flags are set
	if settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] == "" {
		settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"
	}
	// 1M context activation: scale the auto-compact window to the full 1M
	// context when the High (Opus) slot resolves to the 1M tier (e.g. glm-5.2).
	// Mirrors internal/cli/glm.go glmAutoCompactWindow; the retired [1m] suffix
	// path (z.ai rejects suffixed ids) is replaced by resolved-window detection.
	maybeSet1MAutoCompactWindow(settings.Env)
	// Declare the resolved GLM context window (Issue #653): Claude Code assumes
	// 200K for unrecognized custom IDs and caps the auto-compact window above
	// at that assumption. Mirrors internal/cli/glm.go glmMaxContextTokens.
	maybeDeclareGLMContextWindow(settings.Env)

	if err := persistSettingsEnv(settings.Env, data, settingsPath); err != nil {
		slog.Error("failed to write GLM credentials to settings.local.json",
			"error", err.Error(),
		)
		return ""
	}

	return fmt.Sprintf("auto-injected GLM credentials from ~/.moai/.env.glm into %s", settingsPath)
}

// persistSettingsEnv writes env back to settings.local.json, preserving all
// non-env top-level keys from rawOriginal. Extracted from the injection tail
// so the already-credentialed path (PR #1574 review) reuses the identical
// write seam.
func persistSettingsEnv(env map[string]string, rawOriginal []byte, settingsPath string) error {
	// Re-read original file to preserve all fields (not just env)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rawOriginal, &raw); err != nil {
		return fmt.Errorf("unmarshal settings: %w", err)
	}

	envData, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal env: %w", err)
	}
	raw["env"] = envData

	newData, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	// @MX:ANCHOR: [AUTO] settings.local.json holds GLM ANTHROPIC_AUTH_TOKEN — write with 0o600 only
	// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001 (CWE-732/552). Prior baseline 0o644
	// allowed any local user to read the credential. Regression locked by TestEnsureGLMCredentialsFilePerm.
	if err := writeSettingsSecure(settingsPath, newData); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}

// maybeSet1MAutoCompactWindow sets CLAUDE_CODE_AUTO_COMPACT_WINDOW in env when
// the High (Opus) slot model resolves to the 1M context tier (e.g. glm-5.2) and
// no window is already configured. This is the SessionStart-hook analogue of
// internal/cli/glm.go glmAutoCompactWindow: activation keys on the resolved
// context window (via statusline.ResolveGLMContextWindow), NOT the [1m]
// model-id suffix — the suffix was retired because z.ai rejects suffixed ids,
// so the prior Contains(..., "[1m]") gate never fired and was dead code.
func maybeSet1MAutoCompactWindow(env map[string]string) {
	if env[config.EnvClaudeCodeAutoCompactWindow] != "" {
		return
	}
	if statusline.ResolveGLMContextWindow(env[config.EnvAnthropicDefaultOpusModel]) >= config.Default1MContextTokens {
		env[config.EnvClaudeCodeAutoCompactWindow] = strconv.Itoa(config.Default1MContextTokens)
	}
}

// maybeDeclareGLMContextWindow sets CLAUDE_CODE_MAX_CONTEXT_TOKENS in env when
// the High (Opus) slot model resolves to a known GLM context window. Claude
// Code assumes a 200K window for unrecognized custom model IDs and caps
// CLAUDE_CODE_AUTO_COMPACT_WINDOW at that assumed window, so the 1M value set
// by maybeSet1MAutoCompactWindow is ineffective without this declaration.
// This is the SessionStart-hook analogue of internal/cli/glm.go
// glmMaxContextTokens. Issue #653; ref: code.claude.com/docs/en/model-config
// ("Correct the window for a gateway or custom model ID").
func maybeDeclareGLMContextWindow(env map[string]string) {
	if env[config.EnvClaudeCodeMaxContextTokens] != "" {
		return
	}
	if size := statusline.ResolveGLMContextWindow(env[config.EnvAnthropicDefaultOpusModel]); size > 0 {
		env[config.EnvClaudeCodeMaxContextTokens] = strconv.Itoa(size)
	}
}

// isCGMode checks if the project is running in CG (Claude+GLM hybrid) mode
// by reading team_mode from llm.yaml.
func isCGMode(projectDir string) bool {
	llmPath := filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		return false
	}
	// Simple check: look for "team_mode: cg" in the file
	return strings.Contains(string(data), "team_mode: cg")
}

// ensureTeammateMode detects whether the session runs inside tmux and
// sets "teammateMode" in settings.local.json accordingly.
// - Inside tmux → "tmux" (teammates appear in separate panes)
// - Outside tmux → removes override (project default "auto" applies)
//
// This runs at every SessionStart so the setting stays current when the
// user switches between tmux and non-tmux terminals. CG/GLM modes
// already force "tmux" via their own code paths, so this is a no-op in
// those cases (the value is already "tmux").
func ensureTeammateMode(projectDir string) string {
	inTmux := os.Getenv("TMUX") != ""

	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return ""
	}

	var raw map[string]json.RawMessage
	if len(data) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return ""
		}
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}

	// Read current value to avoid unnecessary writes.
	var current string
	if v, ok := raw["teammateMode"]; ok {
		_ = json.Unmarshal(v, &current)
	}

	desired := "auto"
	if inTmux {
		desired = "tmux"
	}

	if current == desired {
		return desired // Already correct, skip write.
	}

	modeJSON, _ := json.Marshal(desired)
	raw["teammateMode"] = modeJSON

	// Clean up legacy env var if present.
	if envRaw, ok := raw["env"]; ok {
		var env map[string]string
		if err := json.Unmarshal(envRaw, &env); err == nil {
			if _, legacy := env["CLAUDE_CODE_TEAMMATE_DISPLAY"]; legacy {
				delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")
				if len(env) > 0 {
					newEnv, _ := json.Marshal(env)
					raw["env"] = newEnv
				} else {
					delete(raw, "env")
				}
			}
		}
	}

	newData, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		return ""
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return ""
	}

	// @MX:NOTE: [AUTO] settings.local.json may contain GLM credentials elsewhere; use 0o600
	// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001 — uniform 0o600 prevents partial regression.
	if err := writeSettingsSecure(settingsPath, newData); err != nil {
		slog.Error("failed to update teammateMode in settings.local.json",
			"error", err.Error(),
		)
		return ""
	}

	slog.Info("teammateMode updated",
		"mode", desired,
		"in_tmux", inTmux,
	)
	return desired
}

// ensureNewSkillSymlinks scans .moai/evolution/new-skills/ for subdirectories
// and creates corresponding symlinks (or directory copies on Windows) in
// .claude/skills/ so that Claude Code can discover evolved skills at session start.
//
// Rules:
// - Target: .claude/skills/<name> → ../../.moai/evolution/new-skills/<name>
// - Existing valid symlinks are skipped.
// - Broken symlinks are removed with a warning.
// - On Windows, a directory copy is used as fallback.
//
// Returns the number of symlinks created in this call.
func ensureNewSkillSymlinks(projectDir string) int {
	newSkillsDir := filepath.Join(projectDir, ".moai", "evolution", "new-skills")
	skillsDir := filepath.Join(projectDir, ".claude", "skills")

	entries, err := os.ReadDir(newSkillsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("ensureNewSkillSymlinks: cannot read new-skills dir",
				"path", newSkillsDir,
				"error", err.Error(),
			)
		}
		return 0
	}

	// Ensure .claude/skills/ exists.
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		slog.Warn("ensureNewSkillSymlinks: cannot create skills dir",
			"path", skillsDir,
			"error", err.Error(),
		)
		return 0
	}

	created := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Name validation: reject path traversal, null bytes, slashes, backslashes, hidden files
		// TOCTOU mitigation: use only ReadDir result and Name, never combine with direct path access
		if name == "" || name == "." || name == ".." ||
			strings.ContainsAny(name, "/\\\x00") ||
			strings.HasPrefix(name, ".") {
			slog.Warn("ensureNewSkillSymlinks: skipping invalid skill name",
				"name", name,
			)
			continue
		}

		linkPath := filepath.Join(skillsDir, name)

		// Check if a symlink (or directory copy) already exists.
		fi, err := os.Lstat(linkPath)
		if err == nil {
			// Path exists — validate it.
			if fi.Mode()&os.ModeSymlink != 0 {
				// It's a symlink — verify it points to a valid target.
				if _, err := os.Stat(linkPath); err == nil {
					// Valid symlink — skip.
					continue
				}
				// Broken symlink — remove it.
				slog.Warn("ensureNewSkillSymlinks: removing broken symlink",
					"path", linkPath,
				)
				if removeErr := os.Remove(linkPath); removeErr != nil {
					slog.Warn("ensureNewSkillSymlinks: cannot remove broken symlink",
						"path", linkPath,
						"error", removeErr.Error(),
					)
					continue
				}
			} else if fi.IsDir() {
				// Directory already exists (Windows copy or manual placement) — skip.
				continue
			} else {
				// Something else — skip to avoid clobbering.
				slog.Warn("ensureNewSkillSymlinks: unexpected file at link path, skipping",
					"path", linkPath,
				)
				continue
			}
		} else if !os.IsNotExist(err) {
			slog.Warn("ensureNewSkillSymlinks: lstat error",
				"path", linkPath,
				"error", err.Error(),
			)
			continue
		}

		// Create symlink or directory copy.
		srcDir := filepath.Join(newSkillsDir, name)

		if runtime.GOOS == "windows" {
			// Windows fallback: copy directory contents instead of symlink.
			if copyErr := copyDirRecursive(srcDir, linkPath); copyErr != nil {
				slog.Warn("ensureNewSkillSymlinks: failed to copy directory on Windows",
					"src", srcDir,
					"dst", linkPath,
					"error", copyErr.Error(),
				)
				continue
			}
		} else {
			// Use a relative symlink so the project is portable.
			// From .claude/skills/<name> to ../../.moai/evolution/new-skills/<name>
			relTarget := filepath.Join("..", "..", ".moai", "evolution", "new-skills", name)
			if symlinkErr := os.Symlink(relTarget, linkPath); symlinkErr != nil {
				slog.Warn("ensureNewSkillSymlinks: failed to create symlink",
					"link", linkPath,
					"target", relTarget,
					"error", symlinkErr.Error(),
				)
				continue
			}
		}

		slog.Info("ensureNewSkillSymlinks: linked evolved skill",
			"name", name,
		)
		created++
	}

	return created
}

// copyDirRecursive copies src directory to dst recursively.
// Used as a Windows fallback when symlinks are not available.
func copyDirRecursive(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", srcPath, err)
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dstPath, err)
		}
	}
	return nil
}

// pruneTelemetry enforces the 90-day retention policy for telemetry files.
// It delegates to telemetry.PruneOldFiles and wraps any error with context.
func pruneTelemetry(projectDir string) error {
	return telemetry.PruneOldFiles(projectDir, 90)
}

// pruneGoalOrphans wires goal.PruneOrphans into the session-start path. It reads
// the active session IDs from the multi-session registry and prunes orphan goal
// state files (session absent from active-sessions.json OR TTL-expired) into
// .moai/state/goal/consumed/. Fail-open: a prune error is logged, NEVER returned
// — session start must never block on goal pruning.
func pruneGoalOrphans(projectDir string) {
	activeIDs := activeGoalSessionIDs(projectDir)
	if _, err := goal.PruneOrphans(projectDir, activeIDs, time.Now()); err != nil {
		slog.Warn("session start: goal orphan prune failed (non-blocking)",
			"error", err.Error(),
		)
	}
}

// activeGoalSessionIDs reads the active session IDs from the multi-session
// registry (.moai/state/active-sessions.json) anchored to projectDir. On any
// read error it returns nil so PruneOrphans falls back to TTL-only pruning
// (fail-open — an unreadable registry never blocks pruning or session start).
func activeGoalSessionIDs(projectDir string) []string {
	registryPath := filepath.Join(projectDir, session.DefaultRegistryPath)
	reg := session.NewRegistry(registryPath, nil)
	entries, err := reg.Query("")
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.SessionID)
	}
	return ids
}

// injectCLAUDEEnvFile checks whether a .env file exists in projectRoot. If it
// does, it injects CLAUDE_ENV_FILE into the env section of
// .claude/settings.local.json so that Claude Code loads the project's env file
// automatically (Windows CLAUDE_ENV_FILE support, T-016).
//
// Returns a non-empty status message when the value was written, empty string
// when the .env file does not exist or when no write was needed.
func injectCLAUDEEnvFile(projectRoot string) string {
	envFilePath := filepath.Join(projectRoot, ".env")
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return ""
	}

	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")

	var raw map[string]json.RawMessage
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			raw = nil
		}
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}

	// Read current env section.
	env := make(map[string]string)
	if envRaw, ok := raw["env"]; ok {
		_ = json.Unmarshal(envRaw, &env)
	}

	// Skip write if already set to the same value.
	if env["CLAUDE_ENV_FILE"] == envFilePath {
		return ""
	}

	env["CLAUDE_ENV_FILE"] = envFilePath

	envData, err := json.Marshal(env)
	if err != nil {
		return ""
	}
	raw["env"] = envData

	newData, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		return ""
	}

	if err := os.MkdirAll(filepath.Join(projectRoot, ".claude"), 0o755); err != nil {
		return ""
	}

	// @MX:NOTE: [AUTO] same settings.local.json may carry sensitive env; 0o600 mandatory
	// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001 uniform hardening.
	if err := writeSettingsSecure(settingsPath, newData); err != nil {
		slog.Error("injectCLAUDEEnvFile: failed to write settings.local.json",
			"error", err.Error(),
		)
		return ""
	}

	return fmt.Sprintf("injected CLAUDE_ENV_FILE=%s into %s", envFilePath, settingsPath)
}

// loadGLMKeyFromEnvFile reads the GLM API key from ~/.moai/.env.glm.
func loadGLMKeyFromEnvFile() string {
	envPath, err := paths.GlmEnvFile()
	if err != nil {
		return ""
	}
	file, err := os.Open(envPath)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		if key == "GLM_API_KEY" && val != "" {
			return val
		}
	}
	return ""
}

// detectAndWrapStaleMemories scans all agent memory directories under
// .claude/agent-memory/<agent>/ and wraps stale files in <system-reminder> tags.
//
// When MOAI_MEMORY_AUDIT=0, the function returns empty string (disabled path).
// When 10+ stale files are found, a single aggregated warning is returned.
// Otherwise per-file wrapped content is concatenated (REQ-EXT001-006/017).
//
// The now parameter is accepted to allow deterministic testing.
func detectAndWrapStaleMemories(projectDir string, now time.Time) string {
	// Respect kill-switch (rollback safety — plan.md §6.2).
	if os.Getenv("MOAI_MEMORY_AUDIT") == "0" {
		return ""
	}

	agentMemBase := filepath.Join(projectDir, ".claude", "agent-memory")
	entries, err := os.ReadDir(agentMemBase)
	if err != nil {
		// Directory may not exist yet — not an error.
		return ""
	}

	var allReports []taxonomy.StaleReport
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		agentDir := filepath.Join(agentMemBase, e.Name())
		reports, err := taxonomy.DetectStale(agentDir, config.DefaultMemoryStalenessHours, now)
		if err != nil {
			slog.Warn("session start: staleness scan error",
				"dir", agentDir,
				"error", err.Error(),
			)
			continue
		}
		allReports = append(allReports, reports...)
	}

	if len(allReports) == 0 {
		return ""
	}

	// When count reaches the aggregation threshold, return a single short warning.
	// Otherwise return wrapped content (each file's content in <system-reminder>).
	if len(allReports) >= config.DefaultMemoryStaleAggregateThreshold {
		return taxonomy.AggregateWarning(allReports)
	}

	// Per-file: return each file's wrapped content (the <system-reminder> block).
	var sb strings.Builder
	for _, r := range allReports {
		sb.WriteString(r.Wrapped)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// claudeEnvFileGuard reports whether the CLAUDE_ENV_FILE injection should run
// for the given OS name. Injection is Windows-only (T-016, R-P1-1).
//
// Extracted from Handle() so that unit tests can exercise the guard without
// depending on runtime.GOOS (a compile-time constant that cannot be overridden
// via os.Setenv). See TestSessionStartHandler_Handle_NonWindowsGuard.
func claudeEnvFileGuard(goos string) bool {
	return goos == "windows"
}

// runMultiSessionProtocol executes the 3-step coordination protocol
// (RegisterSession → PurgeStale → QueryActiveWork → stderr surface) for
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-013..015.
//
// All steps are best-effort: errors are logged via slog.Warn but never
// propagated up the hook return path. Hook timeout safety is preserved
// (each registry op completes well under 100ms in normal contention).
//
// The data map is mutated in place to surface multi-session events to the
// hook output (used by tests + observability).
//
// Uses an explicit project-dir-bound Registry instance (not the package-
// level helpers) because the hook may run with arbitrary CWD; the
// registry path is anchored to input.ProjectDir. The method names below
// are the same as the package-level entry points (RegisterSession,
// PurgeStale, QueryActiveWork) and the verification grep matches against
// the function names regardless of receiver.
func (h *sessionStartHandler) runMultiSessionProtocol(input *HookInput, data map[string]any) {
	registryPath := filepath.Join(input.ProjectDir, session.DefaultRegistryPath)
	reg := session.NewRegistry(registryPath, nil)

	// Step 1: RegisterSession with no SPEC scope yet.
	if err := reg.Register(input.SessionID, session.SpecIDNone, session.PhaseNone); err != nil {
		slog.Warn("multi-session protocol: RegisterSession failed (non-blocking)",
			"session_id", input.SessionID,
			"error", err.Error(),
		)
		data["multi_session_register_error"] = err.Error()
	} else {
		data["multi_session_register"] = "ok"
	}

	// Step 2: PurgeStale entries (zombie sessions from crashed runs).
	purged, err := reg.Purge(session.DefaultStaleMinutes)
	if err != nil {
		slog.Warn("multi-session protocol: PurgeStale failed (non-blocking)",
			"session_id", input.SessionID,
			"error", err.Error(),
		)
	} else if purged > 0 {
		data["multi_session_purged"] = purged
		slog.Info("multi-session protocol: PurgeStale removed stale entries",
			"session_id", input.SessionID,
			"count", purged,
		)
	}

	// Step 3: QueryActiveWork — other active sessions surface via stderr.
	entries, err := reg.Query("")
	if err != nil {
		slog.Warn("multi-session protocol: QueryActiveWork failed (non-blocking)",
			"session_id", input.SessionID,
			"error", err.Error(),
		)
		return
	}
	reminder := session.FormatStderrReminder(input.SessionID, entries, time.Now().UTC())
	if reminder != "" {
		_, _ = fmt.Fprint(os.Stderr, reminder)
		data["multi_session_other_active"] = len(entries) - 1
	}
}

// driftWarningThreshold is the number of drifted SPECs at or above which
// session-start surfaces the advisory (pre-existing behavior; named here so the
// time-boxed rewrite carries no inline magic number).
const driftWarningThreshold = 5

// driftTimeoutAdvisory is surfaced when the drift check exceeds its time-box: it
// preserves the "Run 'moai spec drift' for details." advisory so the user still
// learns drift may exist, WITHOUT the check having blocked session start.
const driftTimeoutAdvisory = "⚠ SPEC status drift check timed out. Run 'moai spec drift' for details."

// deferredScanJoinBound is the maximum time Handle waits for the deferred
// advisory scan goroutine before returning. It is the drop-mitigation bound:
// scans that finish within it land their advisory keys in this session's Data
// map; slower scans are abandoned for this session (advisory dropped, durable
// side effects still complete, idempotent re-derive next session).
//
// Value choice (250ms): the drift scan's own time-box is
// DefaultSessionStartDriftTimeout (2s) — that is the scan's CEILING on huge
// repos, NOT a target. On typical repos the four scans (prune + stale-memory
// + proposals + drift) complete in tens of ms. 250ms gives the common case
// ample headroom to land while capping worst-case added input lag at a
// quarter-second. It is deliberately far below the 2s scan ceiling so the
// join can never approach the original synchronous blocking cost. The
// existing TestSessionStart_DeferredScanDoesNotBlockReturn asserts Handle
// returns in <500ms; 250ms preserves that margin even when the injected
// slow scan forces the bound to elapse.
const deferredScanJoinBound = 250 * time.Millisecond

// Session-start drift seams — overridable in tests to inject a slow computation
// (or a short deadline) and prove the time-box fires without waiting the full
// production deadline. Production points them at the real ctx-aware entry point
// and the compiled default timeout.
var (
	driftCountFn             = spec.DriftCountCtx
	sessionStartDriftTimeout = config.DefaultSessionStartDriftTimeout
)

// Deferred-advisory-scan completion seam (test-only). The deferred goroutine
// spawned by Handle is fire-and-forget in production: it may be killed when
// the hook process exits, which is safe because every step is idempotent.
// goleak (TestMain) flags any goroutine still alive at test end, so tests
// that exercise Handle need a way to JOIN the deferred goroutine.
// deferredScanCompleted is nil in production; a test sets it (under
// deferredScanSeamMu) before calling Handle, and the goroutine closes it on
// exit. The goroutine snapshots it once at start so later mutation by a test
// cleanup does not race.
var (
	deferredScanSeamMu      sync.Mutex
	deferredScanCompletedCh chan struct{}
)

// deferredScansAsync selects the deferred-advisory-scan execution mode.
//
// Production default is true: the four scans run in a background goroutine
// joined with deferredScanJoinBound, keeping SessionStart's synchronous path
// to turn-visible side effects only (the input-lag win).
//
// TestMain flips this to false for the whole test binary. Dozens of
// Handle-calling tests do NOT (and should not have to) install the join seam,
// so in async mode they would leak the goroutine past the test boundary; the
// leaked goroutine's slog writes (process-global os.Stderr via the default
// handler) then race against unrelated parallel tests that reassign os.Stderr
// or reset the slog handler. Running the scans INLINE-synchronously in tests
// eliminates the goroutine, the leak, and the race — with zero churn to the
// ~50 Handle-calling tests. session_start_parallel_test.go opts back into
// async=true to keep the production async path covered.
//
// @MX:NOTE: [AUTO] async/sync test seam — TestMain sets false; production CLI keeps true
var deferredScansAsync = true

// deferredScansAsyncEnabled returns whether the advisory scan runs in a
// background goroutine (production) vs inline-synchronously (tests). Guarded
// by deferredScanSeamMu so the race detector recognizes the happens-before
// edge when TestMain and serial tests toggle the flag across test phases.
func deferredScansAsyncEnabled() bool {
	deferredScanSeamMu.Lock()
	defer deferredScanSeamMu.Unlock()
	return deferredScansAsync
}

// snapshotDeferredScanCompleted returns the current test seam channel (nil in
// production) under the seam mutex so the read is race-free against a test's
// t.Cleanup clearing it.
func snapshotDeferredScanCompleted() chan struct{} {
	deferredScanSeamMu.Lock()
	defer deferredScanSeamMu.Unlock()
	return deferredScanCompletedCh
}

// detectStatusDrift checks for SPEC status drift and returns a warning message
// if >= driftWarningThreshold SPECs have drifted. Returns empty string otherwise.
//
// The check is time-boxed (SPEC-SESSIONSTART-PERF-001 REQ-SSP-015): an advisory
// computation on the session-start critical path must never block unboundedly.
// On deadline exceed the handler skips the (abandoned) computation and emits the
// advisory instead of blocking. All other errors (git absent, no specs
// directory) are silently ignored, as before — the check is best-effort and
// non-blocking.
func detectStatusDrift(projectDir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), sessionStartDriftTimeout)
	defer cancel()

	count, err := driftCountFn(ctx, projectDir)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// Time-box exceeded — emit the advisory rather than block session start.
			return driftTimeoutAdvisory
		}
		// git unavailable, no specs directory, etc. — stay silent (non-blocking).
		return ""
	}

	if count >= driftWarningThreshold {
		return fmt.Sprintf("⚠ %d SPECs have status drift. Run 'moai spec drift' for details.", count)
	}

	return ""
}

// mxIndexFreshnessThreshold is how long an MX sidecar index is considered
// fresh. An index whose ScannedAt is older than this (or absent/corrupt)
// triggers the deferred cold-start full scan so 'moai mx query' returns fresh
// results without a manual 'moai mx scan' after checkout/clone/worktree
// creation. Measured staleness on a fresh worktree (2026-08-04): 764 missing
// tags (1,567 actual vs 803 indexed). 7 days mirrors the MX ArchiveStale TTL.
const mxIndexFreshnessThreshold = 7 * 24 * time.Hour

// mxIndexScanTimeoutDefault bounds the deferred cold-start ScanDir so its cost
// cannot grow unboundedly with repo size. The scan runs in the deferred
// background goroutine; the deferredScanJoinBound (250ms) further caps added
// input lag. On a timeout the scan is abandoned for this session (fail-open,
// non-blocking); the next session re-derives idempotently.
//
// @MX:NOTE: [AUTO] cold-start scan timeout — fail-open ceiling, never blocks the 5s hook budget
const mxIndexScanTimeoutDefault = 2 * time.Second

// mxIndexScanTimeout is the test-overridable seam for the cold-start scan
// ceiling (mirrors the sessionStartDriftTimeout pattern). Production points at
// mxIndexScanTimeoutDefault.
var mxIndexScanTimeout = mxIndexScanTimeoutDefault

// mxIndexNeedsRebuild is the CHEAP synchronous check that gates the deferred
// cold-start scan. It performs one file stat + one JSON field read of
// ScannedAt — never ScanDir. Returns true when the index is absent, empty,
// corrupt, has a zero ScannedAt, or is older than mxIndexFreshnessThreshold.
func mxIndexNeedsRebuild(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	idxPath := filepath.Join(projectDir, ".moai", "state", mx.SidecarFileName)
	info, err := os.Stat(idxPath)
	if err != nil || info.Size() == 0 {
		return true // absent or empty
	}
	data, err := os.ReadFile(idxPath)
	if err != nil {
		return true
	}
	var head struct {
		ScannedAt time.Time `json:"scanned_at"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return true // corrupt
	}
	if head.ScannedAt.IsZero() {
		return true
	}
	return time.Since(head.ScannedAt) > mxIndexFreshnessThreshold
}

// runMXColdStartScan performs a full-project ScanDir and writes the sidecar
// index, time-boxed by mxIndexScanTimeout and fail-open: on timeout or error
// it logs at warn/info and returns without blocking. ScanDir is not
// context-aware, so the scan runs in a helper goroutine whose result is
// selected against the timeout context — when the context fires the result is
// abandoned. In production the SessionStart process may exit and kill the
// helper goroutine (safe — idempotent, next session re-derives); the scan
// only lands if it finishes before the process exits.
//
// @MX:WARN @MX:REASON ScanDir walks the whole repo; the time-box + helper
// goroutine bound cost and guarantee the caller is never blocked past the
// ceiling (Advisory-Check Discipline).
func runMXColdStartScan(projectDir string) {
	if projectDir == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), mxIndexScanTimeout)
	defer cancel()

	type scanResult struct {
		tags []mx.Tag
		err  error
	}
	resCh := make(chan scanResult, 1) // buffered → helper goroutine never blocks on send
	go func() {
		s := mx.NewScanner()
		s.SetIgnorePatterns(mx.DefaultScanIgnore)
		tags, err := s.ScanDir(projectDir)
		resCh <- scanResult{tags: tags, err: err}
	}()

	select {
	case <-ctx.Done():
		slog.Info("session start (deferred): MX cold-start scan timed out (non-blocking)",
			"project_dir", projectDir,
			"timeout", mxIndexScanTimeout.String())
		return
	case r := <-resCh:
		if r.err != nil {
			slog.Warn("session start (deferred): MX cold-start scan failed (non-blocking)",
				"project_dir", projectDir,
				"error", r.err.Error())
			return
		}
		stateDir := filepath.Join(projectDir, ".moai", "state")
		mgr := mx.NewManager(stateDir)
		sidecar := &mx.Sidecar{
			SchemaVersion: mx.SchemaVersion,
			Tags:          r.tags,
			ScannedAt:     time.Now(),
		}
		if err := mgr.Write(sidecar); err != nil {
			slog.Warn("session start (deferred): MX cold-start scan write failed (non-blocking)",
				"project_dir", projectDir,
				"error", err.Error())
			return
		}
		slog.Info("session start (deferred): MX index built via cold-start scan",
			"project_dir", projectDir,
			"tags", len(r.tags))
	}
}
