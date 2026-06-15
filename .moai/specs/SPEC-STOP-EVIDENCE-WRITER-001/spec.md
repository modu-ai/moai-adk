---
id: SPEC-STOP-EVIDENCE-WRITER-001
title: "Record-time Evidence Writer — Activates the Stop Evidence Gate in Production"
version: "0.1.0"
status: implemented
created: 2026-06-16
updated: 2026-06-16
author: GOOS행님
priority: P1
phase: "v0.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hooks, post-tool-use, telemetry, evidence-writer, evidence-gate, activation, tdd"
era: V3R6
tier: M
---

# SPEC-STOP-EVIDENCE-WRITER-001 — Record-time Evidence Writer

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-16 | manager-spec | Initial draft. Plan-phase artifacts (spec/plan/acceptance/design/research). Tier M, cycle_type=tdd. Direct follow-up to SPEC-STOP-EVIDENCE-GATE-001 (IMP-02/03, completed). Closes the GATE-001 knowingly-dormant gap by populating record-time evidence fields (`IsTestPass`/`IsTestFail`/`PathKind`) so the dormant Stop evidence gate fires in production. |

---

## A. Background — The Gate Activation This SPEC Delivers

This SPEC is the **direct follow-up** to `SPEC-STOP-EVIDENCE-GATE-001` (IMP-02/03, status: completed, Mx close commit `59561c92d`). GATE-001 built the Stop-hook verification-evidence gate (`runEvidenceGate` in `internal/hook/stop.go`) plus the session-ledger inference layer (`internal/hook/session_ledger.go`), and it added the `IsTestPass` / `IsTestFail` / `PathKind` `omitempty` fields to `telemetry.UsageRecord` (`internal/telemetry/types.go`). But GATE-001 knowingly shipped the gate as a **production no-op** because no writer ever populates those evidence fields.

### A.1 The doctrine this line serves

The whole STOP-EVIDENCE line serves `.claude/rules/moai/core/verification-claim-integrity.md` — the **no-unobserved-verification-claim invariant** ("evidence absent ≠ evidence of success"). The gate's purpose is to advisory-warn when a session claims success (a skill/agent ran) but no genuine test-pass evidence was observed in that session.

### A.2 [HARD] The dormancy gap GATE-001 left open (verified ground truth)

GATE-001 was honest that its gate is dormant. The dormancy has three mechanical causes, all in the single production `UsageRecord` writer:

1. The only production `UsageRecord` writer is `logSkillUsage` (`internal/hook/post_tool_metrics.go:47`). It constructs the record with `Outcome: telemetry.OutcomeUnknown` **hardcoded** (`post_tool_metrics.go:81`), and NEVER sets `PathKind`, `IsTestPass`, or `IsTestFail` (they default to empty/false). `Phase: "none"` and `DurationMs: 0` are also hardcoded.
2. `logSkillUsage` fires at **Skill *invocation* time** (PostToolUse of the `Skill` tool). At that moment no test result exists yet — so it is structurally impossible for `logSkillUsage` itself to know `IsTestPass`. Evidence MUST be captured from **other** tool events across the session.
3. The PostToolUse `Handle` method (`post_tool.go:145-253`) records a `UsageRecord` ONLY for `input.ToolName == "Skill"` (via `logSkillUsage`) and a separate task-metrics line for `"Agent"/"Task"`. **Bash, Edit, and Write PostToolUse events write NO `UsageRecord` at all today** — they only feed LSP / AST / MX validation. So there is currently no record carrying test-pass or code-change evidence.

Consequently, in the current stream `hasSuccessClaim` (Outcome ∈ {success, partial}) is always false, no record carries a binary signal, and `inferPathKind` returns `unknown` for effectively every real session → the gate never fires.

### A.3 [HARD] What this SPEC delivers — genuine activation, not a partial field-fill

[HARD] This SPEC makes the gate **genuinely active in production**, NOT a minimal field-fill that leaves it partially no-op. The user-confirmed scope is Comprehensive (Tier M). The deliverable expands the PostToolUse writer to record **evidence-bearing tool events** so that the session ledger the gate already consumes carries real evidence:

- **(W1) Bash test-command evidence** — `When` a `Bash` PostToolUse event runs a recognized test command (`go test`, `pytest`, `cargo test`, `npm test`, `jest`, etc.), the writer **shall** record a `UsageRecord` with `IsTestPass` / `IsTestFail` derived from the command text plus the observed exit/output signal.
- **(W2) Edit/Write path-kind evidence** — `When` an `Edit` or `Write` PostToolUse event modifies a file, the writer **shall** record a `UsageRecord` with `PathKind` classified as `code-change` (source file) or `docs-only` (docs / markdown / config) from the file path.

The existing `buildSessionLedger` → `inferPathKind` / `BinaryPass` / `BinaryFail` chain already wires these fields into the gate's `evaluateEvidence`. No gate-side logic change is required — the activation is purely on the **write side**.

### A.4 [HARD] Falsifiable activation point — avoids the GATE-001 dormant-scaffold trap

[HARD] GATE-001's lesson `L_plan_auditor_value_realization_unfalsifiable` applies: a gate/scaffold SPEC must be auditable as to whether it actually fires in production. This SPEC's activation is **falsifiable end-to-end** and is encoded in acceptance.md AC-SEW-009:

> **Given** a session in which a code-change-bearing tool event ran (Edit/Write of a `.go` file → `PathKind=code-change`) AND a test event was observed as a fail (or no test was observed), with a success claim present, **When** the Stop hook's `runEvidenceGate` runs against that session's ledger, **Then** it emits the advisory finding to stderr. **Given** instead a session where an observed test-pass was recorded (`IsTestPass=true`), **Then** the gate emits NO finding.

If any part of the writer remained dormant, the AC would not pass — so the activation cannot silently regress into a no-op. The writer's production caller is the existing `postToolHandler.Handle` PostToolUse path (`post_tool.go`), which is exercised on every tool call in a live session; there is no separate opt-in flag gating activation.

### A.5 Scope discipline — minimum that genuinely activates the gate

This SPEC implements the **minimum** scope that genuinely activates the gate (Enforce Simplicity / Karpathy Surgical Changes). It does NOT add speculative configurability beyond what activation needs: no per-language test-runner config file, no tunable path-kind regex registry, no new outcome states. The test-command recognizer and path classifier are fixed taxonomies (extensible by editing the constant lists, not by config). The success-claim mechanism reuses the existing `Outcome` flow; this SPEC adds evidence, not a new claim source.

---

## B. Requirements (GEARS notation)

> GEARS subject is generalized. The `<subject>` here is "the PostToolUse evidence writer" (the new code added to `internal/hook/post_tool.go` + a new evidence-writer file), unless otherwise stated.

### B.1 Core writer requirements

- **REQ-SEW-001** (Event-driven): **When** a `PostToolUse` event arrives with `ToolName == "Bash"`, the evidence writer **shall** inspect the Bash `command` (from `input.ToolInput`) and, **where** the command is a recognized test command, record a `telemetry.UsageRecord` carrying the derived `IsTestPass` / `IsTestFail` binary evidence.
- **REQ-SEW-002** (Event-driven): **When** a `PostToolUse` event arrives with `ToolName == "Edit"` or `ToolName == "Write"`, the evidence writer **shall** classify the target `file_path` into a `PathKind` (`code-change` | `docs-only`) and record a `telemetry.UsageRecord` carrying that `PathKind`.
- **REQ-SEW-003** (Ubiquitous): The evidence writer **shall** detect test pass/fail from a Bash command + its observed result using a pure, side-effect-free classifier that takes the command text and the tool result, and returns `(isTest bool, isPass bool, isFail bool)`.
- **REQ-SEW-004** (Ubiquitous): The evidence writer **shall** classify a file path into `code-change` vs `docs-only` using a pure, side-effect-free classifier whose taxonomy treats source-code extensions as `code-change` and documentation/markdown/text/config-doc paths as `docs-only`.
- **REQ-SEW-005** (Ubiquitous): The evidence writer **shall** persist evidence records through the existing `telemetry.RecordSkillUsage` path (same daily JSONL store, same session-keyed layout) so that `telemetry.LoadBySession` — which the GATE-001 gate already calls — returns them. The writer **shall not** create a new on-disk store.

### B.2 Test pass/fail detection requirements

- **REQ-SEW-006** (Ubiquitous): The test-command recognizer **shall** recognize the canonical multi-language test commands `go test`, `pytest`, `cargo test`, `npm test`, `npm run test`, `jest`, `vitest`, `go test ... -run`, and `pnpm test` / `yarn test` as test commands.
- **REQ-SEW-007** (State-driven): **While** the Bash tool result is available (`input.ToolResponse`/`input.ToolOutput` carrying an exit code or stdout), the recognizer **shall** derive `IsTestPass=true` from a zero-exit / pass signal and `IsTestFail=true` from a non-zero-exit / fail signal; **where** the result is unavailable or ambiguous, it **shall** set neither flag (REQ-SEW-013 graceful degradation — absence ≠ fail).
- **REQ-SEW-008** (Unwanted behavior): The recognizer **shall not** set `IsTestPass` or `IsTestFail` for a non-test Bash command (e.g. `go build`, `ls`, `git status`); a non-test command yields a record with neither binary flag (or no evidence record at all, per design.md §2.3).

### B.3 Path-kind classification requirements

- **REQ-SEW-009** (Ubiquitous): The path classifier **shall** classify source-code file extensions (`.go`, `.py`, `.ts`, `.js`, `.rs`, `.java`, `.kt`, `.cs`, `.rb`, `.php`, `.ex`, `.cpp`, `.scala`, `.r`, `.dart`, `.swift`) as `PathKind="code-change"`.
- **REQ-SEW-010** (Ubiquitous): The path classifier **shall** classify documentation / prose / config-doc paths (`.md`, `.mdx`, `.txt`, `.rst`, `CHANGELOG`, `README`, `.moai/specs/` markdown, `docs/` markdown) as `PathKind="docs-only"`.
- **REQ-SEW-011** (Event-driven): **When** a path matches neither the code nor the docs taxonomy (e.g. an unknown extension), the path classifier **shall** classify it as `PathKind` empty/unknown so the downstream gate treats the session conservatively (no false code-change claim).

### B.4 Behavior-preservation requirements (additive, fail-open)

- **REQ-SEW-012** (Ubiquitous): The evidence writer **shall** be additive to the existing `postToolHandler.Handle` flow — it **shall not** alter the existing `logSkillUsage` (Skill), `logTaskMetrics` (Agent/Task), LSP-diagnostics, AST-scan, MX-validation, memory-audit, or `writeHookMetric` behavior, and **shall not** change the `HookOutput` JSON contract.
- **REQ-SEW-013** (Ubiquitous): The evidence writer **shall** be best-effort and fail-open — it **shall** preserve `logSkillUsage`'s skip semantics (no-project-root skip via `resolveProjectRoot`, no-evidence skip), handle all errors with `slog.Warn` / `slog.Debug` and never return an error, and never block or fail the PostToolUse hook.
- **REQ-SEW-014** (State-driven): **While** the PostToolUse hook executes, the evidence writer **shall** stay within the ≤5s MoAI hook budget — it **shall** perform only an in-memory classification plus one append-write per evidence-bearing event (no test re-execution, no network, no full-repo scan).
- **REQ-SEW-015** (Unwanted behavior): The evidence writer **shall not** invoke `AskUserQuestion` or `mcp__askuser__*` (C-HRA-008 subagent boundary holds for hook code).
- **REQ-SEW-016** (Ubiquitous): The evidence writer **shall** populate the record's `SessionID` from `input.SessionID` so that `telemetry.LoadBySession(projectRoot, sessionID)` correlates the evidence to the gate-evaluated session.

### B.5 Success-claim wiring requirement (the activation hinge)

- **REQ-SEW-017** (Event-driven): **When** the evidence writer records a code-change-bearing or test-bearing `UsageRecord`, it **shall** set `Outcome` such that the GATE-001 gate's success-claim path can fire — i.e. an observed test-pass record carries `Outcome=success` and an observed test-fail record carries an outcome that does NOT count as a success claim — so that the existing `evaluateEvidence` chain (which keys on `SuccessClaims` + `BinaryPass`/`BinaryFail`) reaches a non-nil finding on the targeted code-session false-success shape. The exact `Outcome` assignment is resolved in design.md §3.

---

## C. Exclusions (What NOT to Build)

[HARD] The following are explicitly OUT OF SCOPE for this SPEC.

### C.1 Out of Scope — GATE-001 gate logic (PRESERVE)

- Modifying `runEvidenceGate`, `buildSessionLedger`, `inferPathKind`, `evaluateEvidence`, `Finding`, or `SessionLedger` in `internal/hook/session_ledger.go`. These are **PRESERVE** targets — this SPEC adds write-side evidence only; the read-side gate is unchanged.
- Modifying `internal/hook/stop.go` (the `runEvidenceGate` call site) or `internal/telemetry/types.go` (the omitempty fields already exist from GATE-001).

### C.2 Out of Scope — IMP-06 invariant doctrine

- Baseline-integrity attribution, the 5-section evidence-bearing report format, and the no-unobserved-verification-claim invariant codification — owned by `SPEC-EVIDENCE-CLAIM-INVARIANT-001` (completed). This SPEC does not touch `.claude/rules/moai/core/verification-claim-integrity.md`.

### C.3 Out of Scope — advisory→blocking promotion

- Promoting the gate from advisory to blocking (exit-2 activation). GATE-001 keeps exit-2 dormant; this SPEC keeps the advisory/warn-first/fail-open contract unchanged. No `decision: block` is emitted.

### C.4 Out of Scope — configurability beyond activation

- A per-language test-runner configuration file or a tunable path-kind regex registry. The recognizer/classifier taxonomies are fixed constant lists (extensible by code edit, not config). No new `.moai/config/sections/*.yaml` field is added.

### C.5 Out of Scope — the originating sync-phase incident shape

- Capturing the originating sync-phase `L_manager_docs_false_backfill_report` incident (a docs-only / sync git-state false-backfill shape). That shape is `docs-only` EXEMPT by design and is owned by IMP-06 doctrine, NOT this gate. This SPEC targets only the **code-session false-success** shape.

### C.6 Out of Scope — logSkillUsage retrofit

- Changing `logSkillUsage`'s existing hardcoded `Outcome: OutcomeUnknown` / `Phase: "none"` for the Skill event itself. The Skill-invocation record stays as-is; this SPEC adds NEW evidence records on the Bash/Edit/Write branches rather than retrofitting the Skill record (the Skill record cannot know test results — see §A.2 #2).

### C.7 Out of Scope — non-test evidence kinds

- Coverage-delta evidence, lint-clean evidence, or build-success evidence as binary signals. Only test pass/fail (Bash) and path-kind (Edit/Write) are in scope; other evidence kinds are deferred.

---

## D. Affected Files (estimate)

| File | Action | Notes |
|------|--------|-------|
| `internal/hook/post_tool.go` | EXTEND | Add evidence-writer call on the Bash/Edit/Write PostToolUse branch (additive, after existing observers). |
| `internal/hook/evidence_writer.go` | NEW | The pure classifiers (`classifyTestCommand`, `classifyPathKind`) + the `logEvidence(input)` writer (mirrors `logSkillUsage` structure). |
| `internal/hook/evidence_writer_test.go` | NEW | TDD: classifier unit tests + writer integration tests + the falsifiable gate-activation end-to-end test. |
| `internal/hook/post_tool_test.go` | EXTEND (if present) | Behavior-preservation assertions for the existing Handle flow. |

`internal/telemetry/types.go` is NOT modified — the `IsTestPass`/`IsTestFail`/`PathKind` omitempty fields already exist (added by GATE-001). `internal/hook/session_ledger.go` and `internal/hook/stop.go` are PRESERVE.

---

## E. Tier & Methodology

| Field | Value |
|-------|-------|
| Tier | M (estimated 300-1000 LOC, 2-4 files affected — see plan.md §A justification) |
| cycle_type | tdd (per `.moai/config/sections/quality.yaml` development_mode=tdd) |
| Artifact set | 5 files (Tier M baseline 3 + design.md for the record-time evidence-sourcing architecture + research.md for pipeline grounding) |
| plan-auditor PASS threshold | 0.80 (Tier M) |
