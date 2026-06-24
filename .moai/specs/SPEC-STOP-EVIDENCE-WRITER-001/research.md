# Research — SPEC-STOP-EVIDENCE-WRITER-001

> Codebase investigation grounding the record-time evidence-writer design. All claims cite `file:line` evidence verified by Read/grep during plan-phase.

## §1 The telemetry / hook pipeline (current state)

### §1.1 The single production UsageRecord writer

`internal/hook/post_tool_metrics.go` — `logSkillUsage(input *HookInput)`:

- Line 47: function entry. Resolves project root via `resolveProjectRoot(input)` (line 48); returns early if empty (line 49-55) — **no-project-root skip** (PRESERVE per REQ-SEW-013).
- Lines 59-67: parses `skillToolInput` from `input.ToolInput`; returns early if `si.Skill == ""` — **no-skill-id skip** (PRESERVE).
- Lines 72-82: constructs `telemetry.UsageRecord` with `Outcome: telemetry.OutcomeUnknown` **hardcoded** (line 81), `Phase: "none"` (line 79), `DurationMs: 0` (line 80). `IsTestPass`/`IsTestFail`/`PathKind` are never set → default false/false/"".
- Lines 84-90: `telemetry.RecordSkillUsage(projectRoot, r)`; errors handled with `slog.Warn`, never returned — **best-effort fail-open** (PRESERVE).

This confirms ground-truth item #1: the only writer hardcodes `Outcome=Unknown` and sets no evidence fields.

### §1.2 The UsageRecord data model is already evidence-ready

`internal/telemetry/types.go`:

- Lines 25-27: `IsTestPass bool json:"is_test_pass,omitempty"`, `IsTestFail bool json:"is_test_fail,omitempty"`, `PathKind string json:"path_kind,omitempty"` — added by GATE-001 (the SPEC reference comment at line 18-24 names `SPEC-STOP-EVIDENCE-GATE-001` and explicitly defers record-time population to the writer successor: *"record-time 채움(populate)은 본 SPEC scope 밖이다 — 후속 writer SPEC SPEC-STOP-EVIDENCE-WRITER-001 이 code-change 세션에 set 한다"*).
- Lines 32-36: `PathKindDocsOnly = "docs-only"`, `PathKindCodeChange = "code-change"`, `PathKindUnknown = "unknown"` — the fixed taxonomy constants the classifier must emit.
- Lines 44-50: `OutcomeSuccess`/`OutcomePartial`/`OutcomeError`/`OutcomeUnknown` — the existing outcome constants the writer reuses (no new state needed, satisfies §A.5 scope discipline).

This confirms ground-truth item #2: the data model is ready; only the writer is missing. **`internal/telemetry/types.go` is NOT modified by this SPEC.**

### §1.3 The PostToolUse Handle flow — where evidence events are (not) recorded

`internal/hook/post_tool.go` — `postToolHandler.Handle(ctx, input)` (lines 145-253):

- Lines 169-171: `if input.ToolName == "Agent" || input.ToolName == "Task"` → `logTaskMetrics(input)` (task metrics, NOT a UsageRecord).
- Line 176: `logCacheUsage(input)` (prompt-cache telemetry, separate stream).
- **Lines 180-182**: `if input.ToolName == "Skill"` → `logSkillUsage(input)`. **This is the ONLY UsageRecord write in the whole Handle flow.**
- Lines 189-191: `Write`/`Edit` + diagnostics → LSP collect (no UsageRecord).
- Lines 201-211: `Write`/`Edit` + analyzer → AST scan (no UsageRecord).
- Lines 214-216: `Write`/`Edit` + mxValidator → MX validation (no UsageRecord).
- Lines 220-222: `Write`/`Edit` → `runMemoryAudit` (no UsageRecord).
- Line 225: `writeHookMetric` (slow-hook metrics, separate `.moai/observability/` stream).

**CRITICAL FINDING (confirms ground-truth item #3 + the orchestrator's critical open question):** `Bash`, `Edit`, and `Write` PostToolUse events write **NO `UsageRecord`** today. They feed only LSP/AST/MX/memory-audit. Therefore the writer scope MUST **expand** the Handle flow to add a new evidence-writer call on the Bash/Edit/Write branch. This is an *additive* insertion alongside the existing observers (REQ-SEW-012).

### §1.4 The Bash command + result are available at PostToolUse time

`internal/hook/types.go` — `HookInput` (lines 180-279):

- Line 190: `ToolInput json.RawMessage json:"tool_input,omitempty"` — for a Bash event this carries `{"command": "...", ...}`. For Edit/Write it carries `{"file_path": "...", ...}` (used already by `runAstScan` line 266 and `runMemoryAudit` line 554).
- Line 192: `ToolResponse json.RawMessage json:"tool_response,omitempty"` — the PostToolUse result (carries exit/output; `extractDurationMs` at `post_tool_duration.go:80` already parses `duration_ms` from it).
- Line 191: `ToolOutput json.RawMessage json:"tool_output,omitempty"` — legacy output field.
- Line 182: `SessionID` — for the `UsageRecord.SessionID` correlation (REQ-SEW-016).

So both the command text (W1) and the file path (W2) and the result signal are present in the PostToolUse `HookInput`. No new event subscription is needed — only an additive handler branch.

## §2 The gate the writer activates (PRESERVE — read-side, unchanged)

### §2.1 The gate entry point

`internal/hook/stop.go` — `stopHandler.Handle` line 80: `runEvidenceGate(projectDir, input.SessionID)`. This is the GATE-001 gate, inserted additively before the final return (PRESERVE).

### §2.2 The ledger chain the writer feeds

`internal/hook/session_ledger.go`:

- Line 29: `runEvidenceGate` calls `telemetry.LoadBySession(projectRoot, sessionID)` — the same read path the writer's records land in (REQ-SEW-005 / REQ-SEW-016 close the loop).
- Lines 58-78: `buildSessionLedger` aggregates `SuccessClaims` (Outcome ∈ {success, partial}, lines 67-69), `BinaryPass` (any `IsTestPass`, lines 70-72), `BinaryFail` (any `IsTestFail`, lines 73-75).
- Lines 88-116: `inferPathKind` — **(1) explicit `PathKind` wins** (lines 90-94). So once the writer sets `PathKind="code-change"`, `inferPathKind` returns it directly without falling back to the `Phase`/`AgentType` inference (lines 97-112) which is unreliable today (`Phase` is always `"none"`).
- Lines 138-167: `evaluateEvidence` — the activation hinge. Returns a `Finding` ONLY when: `PathKind` is NOT docs-only/unknown (lines 139-142) AND `SuccessClaims > 0` (lines 144-146) AND a binary signal was observed (lines 148-152) AND `BinaryPass == false` (lines 154-157). This is exactly the **code-change + success-claim + no-observed-pass** shape.

**Design consequence:** to make `evaluateEvidence` reach a non-nil Finding, the writer must produce (within one session): (a) at least one record with `PathKind="code-change"`, (b) at least one record with `SuccessClaims > 0` i.e. `Outcome ∈ {success, partial}`, and (c) at least one record with `IsTestFail=true` (a binary signal observed) while NO record has `IsTestPass=true`. Conversely, a record with `IsTestPass=true` sets `BinaryPass` → `evaluateEvidence` returns nil (backed). This is the falsifiable activation pair (REQ-SEW-017, AC-SEW-009). The `Outcome` assignment that produces the success claim is resolved in design.md §3.

### §2.3 The telemetry write/read storage (PRESERVE)

`internal/telemetry/recorder.go`:

- Lines 38-69: `RecordSkillUsage(projectRoot, r)` — appends one JSONL line to `<projectRoot>/.moai/evolution/telemetry/usage-YYYY-MM-DD.jsonl`, mutex-guarded, dir auto-created. The writer reuses this path (REQ-SEW-005 — no new store).
- Lines 73-105: `LoadBySession(projectRoot, sessionID)` — reads today + yesterday JSONL, filters by `SessionID`. The writer's records are session-keyed, so this returns them at Stop time.

## §3 cycle_type & house conventions

- `.claude/rules/moai/languages/go.md`: errors wrapped with `%w`; `_test.go` naming; coverage ≥85% target; table-driven tests preferred; `go test -race` for goroutine code (the writer has none — pure classifiers + one append-write).
- `internal/hook/CLAUDE.md`: hooks are best-effort, ≤5s budget; C-HRA-008 (no AskUserQuestion in `internal/hook/`); `$CLAUDE_PROJECT_DIR` resolution via `resolveProjectRoot` (already used by `logSkillUsage`); `slog.Warn`/`slog.Debug` for diagnostics (never block).
- The new code mirrors `logSkillUsage`'s structure exactly: parse input → `resolveProjectRoot` skip → classify → `RecordSkillUsage` with `slog.Warn` on error. This minimizes novelty (Surgical Changes).

## §4 GATE-001 lessons carried forward

- `L_plan_auditor_value_realization_unfalsifiable` — the writer's value (does the gate now fire?) MUST be grep/test-verifiable in production, not just in a unit test. AC-SEW-009 asserts the end-to-end gate emission with the writer's records in the ledger (no separate opt-in flag; the production caller is the live PostToolUse path). This is why this SPEC is NOT itself a dormant scaffold — its deliverable IS the production firing.
- `L_manager_docs_false_backfill_report` — the motivating defect class. The gate this writer activates targets the **code-session false-success** shape (NOT the originating sync-phase docs-only shape, which is exempt). Restated here to keep the activation honest about what shape it catches (spec.md §C.5).

## §5 Risks surfaced

- **R1** — Bash test-result signal ambiguity. The PostToolUse result shape for Bash exit codes is not perfectly standardized across Claude Code versions. Mitigation: graceful degradation (REQ-SEW-007 / REQ-SEW-013) — when the result is unparseable, set neither binary flag (absence ≠ fail), so the gate stays conservative rather than producing a false finding.
- **R2** — A single session mixes docs-only and code-change events. `inferPathKind` already resolves this conservatively (explicit `code-change` wins over `docs-only` when present — lines 96-109). The writer just records both; the ledger reconciliation is GATE-001's job (PRESERVE).
- **R3** — Recording an evidence record on EVERY Bash/Edit/Write would bloat the JSONL. Mitigation: the writer records only **evidence-bearing** events (a recognized test command, or a path that classifies to a known kind) — non-test Bash and unknown-extension paths produce no record (REQ-SEW-008 / design.md §2.3). This keeps write volume proportional to genuine evidence.
