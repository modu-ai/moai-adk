# SPEC-V3R2-WF-004 Implementation Plan (Phase 1B)

> Implementation plan for Agentless Fixed-Pipeline Classification.
> Companion to `spec.md` v0.2.0 and `research.md` v0.1.0.
> Authored against worktree `feature/SPEC-V3R2-WF-004` at `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-004`.

## HISTORY

| Version | Date       | Author                      | Description                                                              |
|---------|------------|-----------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-02 | MoAI Plan Workflow (Phase 1B) | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B |

---

## 1. Plan Overview

### 1.1 Goal restatement

Codify the classification stated in `spec.md` REQ-WF004-001..002:

- **Pipeline (Agentless, fixed 3-phase)**: `fix`, `coverage`, `mx`, `codemaps`, `clean` (5 utility subcommands)
- **Multi-Agent (open-ended orchestration)**: `plan`, `run`, `sync`, `design` (4 implementation subcommands)

The classification is a **declaration-level breaking change** (BC-V3R2-007 contractual flip): the five utility skills already follow a deterministic phase order (research.md §2.2, §4), so the implementation work is *publishing the contract* and *guarding against future regression* — not changing runtime behavior.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md:60-65`. Concretely:

- **RED**: Write failing audit tests (REQ-WF004-013 enforcement, REQ-WF004-011/014 sentinel checks) before any skill body or rule file is touched. Confirm tests fail in CI.
- **GREEN**: Add classification matrix to `spec-workflow.md` (REQ-WF004-005), add Pipeline Contract headers to 5 utility skills (REQ-WF004-003 + sentinel strings for REQ-WF004-011/014), making tests pass.
- **REFACTOR**: Consolidate the 5 Pipeline Contract headers if they share boilerplate; cross-link from skill bodies to the matrix anchor.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Subcommand Classification matrix (new section) | `.claude/rules/moai/workflow/spec-workflow.md` (new `## Subcommand Classification (Pipeline vs Multi-Agent)` section) | REQ-WF004-001, REQ-WF004-002, REQ-WF004-005 |
| Pipeline Contract headers (5 files) | `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` | REQ-WF004-003, REQ-WF004-004, REQ-WF004-006, REQ-WF004-007, REQ-WF004-008, REQ-WF004-009, REQ-WF004-011, REQ-WF004-015 |
| Mode-flag rejection clauses (4 files) | `.claude/skills/moai/workflows/{plan,run,sync,design}.md` | REQ-WF004-010, REQ-WF004-014 |
| Agentless audit test (new file) | `internal/template/agentless_audit_test.go` | REQ-WF004-013 (CI guard for `AGENTLESS_CONTROL_FLOW_VIOLATION`) |
| Mode-flag content audit test (extend existing or new) | `internal/template/agentless_audit_test.go` (subtests) | REQ-WF004-011 (sentinel `MODE_FLAG_IGNORED_FOR_UTILITY`), REQ-WF004-014 (sentinel `MODE_PIPELINE_ONLY_UTILITY`) |
| CHANGELOG entry | `CHANGELOG.md` (Unreleased section) | Trackable (TRUST 5) |

Embedded-template parity is a **HARD** requirement: every change to `.claude/skills/moai/workflows/*.md` and `.claude/rules/moai/workflow/spec-workflow.md` must also be applied to the corresponding `internal/template/templates/.claude/...` source-of-truth file (per `CLAUDE.local.md` §2 Template-First Rule). `make build` regenerates `internal/template/embedded.go` after template edits.

---

## 2. Milestone Breakdown (M1-M5)

Each milestone is **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule). Dependencies are explicit.

### M1: Test scaffolding (RED phase) — Priority P0

Reference: `internal/template/commands_audit_test.go:11-50` (audit-test scaffold pattern).

Owner role: `expert-backend` (Go test) or direct manager-tdd execution.

Scope:
1. Create `internal/template/agentless_audit_test.go` mirroring `commands_audit_test.go` structure.
2. Test func `TestAgentlessUtilityNoLLMControlFlow` walks `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` in the embedded FS.
3. For each file, scan body (excluding code blocks) for the forbidden-pattern regex set per research.md §6.2:
   - `Use the .* subagent to (decide|determine|choose|select|orchestrate|route|dispatch)`
   - `Use the .* subagent to (plan|design) the (pipeline|workflow|next phase|sequence)`
   - `delegate to .* (orchestrator|router|dispatcher|controller)`
   - `manager-strategy.*subagent.*(branch|fork|conditional)`
4. Test func `TestUtilitySkillsContainModeFlagIgnoredSentinel` asserts each of the 5 utility skills contains the exact sentinel string `MODE_FLAG_IGNORED_FOR_UTILITY`.
5. Test func `TestImplementationSkillsContainPipelineRejectionSentinel` asserts each of the 4 implementation skills (`plan.md`, `run.md`, `sync.md`, `design.md`) contains the exact sentinel string `MODE_PIPELINE_ONLY_UTILITY`.
6. Run `go test ./internal/template/ -run TestAgentless` — confirm RED (all three tests fail because sentinels and classification headers are not yet added).

Verification gate before advancing to M2: at least 5 of the 5 utility skill subtests fail with "missing sentinel" or "no Pipeline Contract section" message; at least 4 of the 4 implementation skill subtests fail similarly.

[HARD] No implementation code in M1 outside of test files.

### M2: Pipeline Contract headers in 5 utility skills (GREEN, part 1) — Priority P0

Owner role: `manager-docs` for content authoring; `expert-backend` for any embedded-template parity build step.

Scope:
1. For each of `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md`, insert a new section after the existing `## Supported Flags` section titled:

   ```markdown
   ## Pipeline Contract (Agentless Classification)

   This subcommand is classified as **Agentless fixed-pipeline** per SPEC-V3R2-WF-004.
   It executes a deterministic 3-phase contract: **localize → repair → validate**.

   - **Phase mapping**: localize ← <Phase 1+2 references>; repair ← <Phase 3 reference>; validate ← <Phase 4 reference>
   - **No LLM-driven control flow**: Agent() invocations exist for executor delegation within phases (e.g., `expert-testing` for measurement) but never select the next phase.
   - **No-op exit**: When the localize phase finds zero targets, the pipeline exits with status `no-op` and exit code 0, skipping repair and validate.
   - **Fail-fast**: When repair encounters an unresolvable error, the pipeline terminates and reports the error. There is no multi-agent fallback.
   - **`--mode` flag handling**: Any `--mode` flag passed to this subcommand is ignored. The system logs `MODE_FLAG_IGNORED_FOR_UTILITY` at info level and proceeds with the fixed pipeline.
   - **Repeatability**: Even when the parent invocation supplies `--mode loop`, the pipeline runs once per command invocation. Re-entry requires explicit user re-invocation.

   See `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification` for the full matrix.
   ```

2. Customize the `<Phase X reference>` slots per research.md §2.2 mapping:
   - `fix.md`: localize ← Phase 1+2+2.5; repair ← Phase 3; validate ← Phase 4 (Phase 4.5/4.6 incidental)
   - `coverage.md`: localize ← Phase 1+2; repair ← Phase 3; validate ← Phase 4
   - `mx.md`: localize ← Pass 1+2; repair ← Pass 3; validate ← post-edit MX scan
   - `codemaps.md`: localize ← Phase 1; repair ← Phase 2+3; validate ← Phase 4
   - `clean.md`: localize ← Phase 1+2; repair ← Phase 4; validate ← Phase 5+5.5

3. Mirror each edit into `internal/template/templates/.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md`.
4. Run `make build` to regenerate `internal/template/embedded.go`.

Reference for sentinel string convention: `MODE_FLAG_IGNORED_FOR_UTILITY` is the canonical info-level log key per `spec.md` REQ-WF004-011 (and shared with WF-003 ecosystem). The string must appear verbatim in the skill body so the M1 sentinel test passes.

Verification: `go test ./internal/template/ -run TestUtilitySkillsContainModeFlagIgnoredSentinel` turns GREEN.

### M3: Mode-flag rejection clauses in 4 implementation skills (GREEN, part 2) — Priority P0

Owner role: `manager-docs`.

Scope:
1. For each of `.claude/skills/moai/workflows/{plan,run,sync,design}.md`, insert a brief rejection clause near the existing flag/mode discussion. Recommended placement: in or after the "Supported Flags" section, or near the Phase 0 section.

   ```markdown
   ## Mode Flag Compatibility

   Per SPEC-V3R2-WF-004, this subcommand is multi-agent (open-ended) and supports the
   `--mode {autopilot|loop|team}` axis defined in SPEC-V3R2-WF-003. The `pipeline` mode
   is **not valid** for this subcommand; passing `--mode pipeline` here triggers
   `MODE_PIPELINE_ONLY_UTILITY` (the same error key used by WF-003 REQ-WF003-016).

   See `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification` for the
   full subcommand × mode matrix.
   ```

2. Mirror each edit into `internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync,design}.md`.
3. Run `make build` to regenerate embedded.go.

Verification: `go test ./internal/template/ -run TestImplementationSkillsContainPipelineRejectionSentinel` turns GREEN.

[HARD] M3 must NOT modify the existing flow declarations, phase orderings, or agent delegation directives in plan/run/sync/design skills. Insert-only.

### M4: Subcommand Classification matrix in `spec-workflow.md` (GREEN, part 3) — Priority P1

Owner role: `manager-docs`.

Scope:
1. Insert a new section `## Subcommand Classification (Pipeline vs Multi-Agent)` in `.claude/rules/moai/workflow/spec-workflow.md` immediately after the existing `## Phase Overview` table (around line 17, before `## Plan Phase`).

   Body template:
   ```markdown
   ## Subcommand Classification (Pipeline vs Multi-Agent)

   Source: SPEC-V3R2-WF-004. Each MoAI subcommand is classified along the
   *control-flow style* axis. The classification governs which agents are spawned,
   how the `--mode` flag is interpreted, and which CI guards apply.

   | Subcommand   | Class          | 3-phase contract (localize → repair → validate)                | `--mode` honored? | Reference                                                    |
   |--------------|----------------|-----------------------------------------------------------------|-------------------|--------------------------------------------------------------|
   | `/moai fix`      | Pipeline (Agentless) | Parallel Scan + Classify + MX context → Auto-Fix → Verify        | No (info log)     | `.claude/skills/moai/workflows/fix.md`                       |
   | `/moai coverage` | Pipeline (Agentless) | Measure + Gap Analysis → Test Generation → Verify                 | No (info log)     | `.claude/skills/moai/workflows/coverage.md`                  |
   | `/moai mx`       | Pipeline (Agentless) | Pass 1 + Pass 2 → Pass 3 → Post-edit scan                         | No (info log)     | `.claude/skills/moai/workflows/mx.md`                        |
   | `/moai codemaps` | Pipeline (Agentless) | Explore → Analyze + Generate → Verify                             | No (info log)     | `.claude/skills/moai/workflows/codemaps.md`                  |
   | `/moai clean`    | Pipeline (Agentless) | Static Analysis + Usage Graph → Safe Removal → Test Verification  | No (info log)     | `.claude/skills/moai/workflows/clean.md`                     |
   | `/moai plan`     | Multi-Agent    | n/a — open-ended (`autopilot` / `loop` / `team` per WF-003)        | Yes (rejects `pipeline`) | `.claude/skills/moai/workflows/plan.md`               |
   | `/moai run`      | Multi-Agent    | n/a — open-ended                                                  | Yes (rejects `pipeline`) | `.claude/skills/moai/workflows/run.md`                |
   | `/moai sync`     | Multi-Agent    | n/a — open-ended                                                  | Yes (rejects `pipeline`) | `.claude/skills/moai/workflows/sync.md`               |
   | `/moai design`   | Multi-Agent    | n/a — open-ended (`autopilot` / `import` / `team` per WF-003)      | Yes (rejects `pipeline`) | `.claude/skills/moai/workflows/design.md`             |

   ### Pipeline Class — Contract

   Pipeline-classified subcommands MUST satisfy:
   - Three deterministic phases (localize → repair → validate); no LLM dispatcher selects the next phase.
   - `Agent()` invocations are permitted only as **executor delegation within a phase** (e.g., `expert-testing` runs the coverage tool); never to decide phase order.
   - When localize finds zero targets, exit with status `no-op` and exit code 0.
   - When repair encounters an unresolvable error, fail-fast (no multi-agent fallback).
   - The CI guard `internal/template/agentless_audit_test.go` enforces the no-LLM-dispatch rule via static text scan.

   ### Out of scope of this matrix

   `/moai feedback`, `/moai review`, `/moai e2e` are *not* yet classified.
   See `spec.md` §1.2 (Non-Goals) — they are deferred to a future SPEC.

   ### Cross-references

   - `--mode` flag matrix: SPEC-V3R2-WF-003 (sibling SPEC, defines `autopilot|loop|team|pipeline`).
   - Pipeline regression guard: `internal/template/agentless_audit_test.go` (REQ-WF004-013).
   - Pattern source: `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6 (Agentless).
   - Research source: `.moai/design/v3-redesign/research/r1-ai-harness-papers.md` §25 (Xia et al. 2024).
   ```

2. Mirror into `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`.
3. Run `make build`.

Reference: `spec-workflow.md:9-15` (existing Phase Overview table) defines the insertion anchor.

Decision per research.md §5: publish in `spec-workflow.md`, **not** a new `workflow-modes.md`. Rationale documented in research.md §5; this plan honors that decision.

### M5: Documentation sync + cross-links — Priority P2

Owner role: `manager-docs`.

Scope:
1. Add CHANGELOG entry under `## [Unreleased]`:
   ```
   ### Added
   - SPEC-V3R2-WF-004: Agentless fixed-pipeline classification for utility subcommands (`fix`, `coverage`, `mx`, `codemaps`, `clean`). Multi-agent classification preserved for `plan`, `run`, `sync`, `design`. Subcommand × class matrix published in `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification`. CI guard `TestAgentlessUtilityNoLLMControlFlow` enforces no-LLM-dispatch in utility skills.
   ```

2. Within each of the 9 affected workflow skill files, add a one-line back-reference at the bottom (or near the Pipeline Contract / Mode Flag Compatibility section):
   ```markdown
   See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification) for the full pipeline-vs-multi-agent contract.
   ```

3. Update `progress.md` for SPEC-V3R2-WF-004 with `run_complete_at` and `run_status: implementation-complete` after M1-M4 land.

4. Verify `go test ./...` passes (full repository test suite per `CLAUDE.local.md` §6 Go Test Execution Rules HARD rule: "After fixing ANY test, run the FULL test suite to catch cascading failures").

5. Verify `make build` produces a clean `internal/template/embedded.go` with all 9 workflow skill changes embedded.

[HARD] No new documents are created in `.moai/specs/` or `.moai/reports/` during M5 — this is a SPEC implementation phase, not a planning phase.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `.claude/skills/moai/workflows/fix.md` | After line 44 (end of `## Supported Flags` table) | Insert `## Pipeline Contract (Agentless Classification)` | M2 / REQ-WF004-003,004,011 |
| `.claude/skills/moai/workflows/coverage.md` | After line 42 (end of `## Supported Flags`) | Insert Pipeline Contract section | M2 |
| `.claude/skills/moai/workflows/mx.md` | After line 59 (end of Flags table) | Insert Pipeline Contract section | M2 |
| `.claude/skills/moai/workflows/codemaps.md` | After line 40 (end of `## Supported Flags`) | Insert Pipeline Contract section | M2 |
| `.claude/skills/moai/workflows/clean.md` | After line 41 (end of `## Supported Flags`) | Insert Pipeline Contract section | M2 |
| `.claude/skills/moai/workflows/plan.md` | After existing flag/mode block (anchor TBD by run-phase agent) | Insert `## Mode Flag Compatibility` section | M3 / REQ-WF004-010,014 |
| `.claude/skills/moai/workflows/run.md` | Same | Same | M3 |
| `.claude/skills/moai/workflows/sync.md` | Same | Same | M3 |
| `.claude/skills/moai/workflows/design.md` | Same | Same | M3 |
| `.claude/rules/moai/workflow/spec-workflow.md` | After line 17 (end of Phase Overview table) | Insert `## Subcommand Classification (Pipeline vs Multi-Agent)` section | M4 / REQ-WF004-005 |
| `CHANGELOG.md` | `## [Unreleased]` section | Add entry | M5 |

### 3.2 To-be-created (new files)

| File | Reason |
|------|--------|
| `internal/template/agentless_audit_test.go` | REQ-WF004-013 CI guard + REQ-WF004-011/014 sentinel verification (M1) |

### 3.3 NOT to be touched (preserved by reference)

The following files are referenced by tests but MUST NOT be modified by this SPEC's run phase. They define the existing rhythm WF-004 is *codifying* — their bodies are evidence that the classification is true today.

- `.claude/skills/moai/workflows/fix.md:33` — flow declaration `Parallel Scan -> Classify -> Fix -> Verify -> Report` (preserve)
- `.claude/skills/moai/workflows/coverage.md:33` — `Measure Coverage -> Identify Gaps -> Generate Tests -> Verify -> Report` (preserve)
- `.claude/skills/moai/workflows/mx.md:71-149` — 3-Pass scan structure (preserve)
- `.claude/skills/moai/workflows/codemaps.md:33` — `Explore Codebase -> Analyze Architecture -> Generate Maps -> Verify -> Report` (preserve)
- `.claude/skills/moai/workflows/clean.md:33` — `Static Analysis -> Usage Graph -> Classification -> Safe Removal -> Test Verification -> Report` (preserve)
- `internal/hook/post_tool.go` — fixed-order MX/LSP/metrics pipeline (preserve)
- `internal/hook/pre_tool.go` — security/secrets/reflective-write guard (preserve)
- `internal/template/commands_audit_test.go:1-60` — Thin Command Pattern audit (preserve as scaffold reference)

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability and research.md §11, the following anchors are load-bearing and cited verbatim throughout this plan:

1. `docs/design/major-v3-master.md:L1069` — §11.6 WF-004 definition
2. `docs/design/major-v3-master.md:L966` — §8 BC-V3R2-007 Agentless flip
3. `docs/design/major-v3-master.md:L993` — §9 Phase 6 Multi-Mode Workflow
4. `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6
5. `.moai/design/v3-redesign/research/r1-ai-harness-papers.md` §25
6. `.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md:80-153` (§2 hooks audit)
7. `.claude/rules/moai/workflow/spec-workflow.md:9-17` (Phase Overview anchor)
8. `.claude/rules/moai/workflow/spec-workflow.md:172-204` (Plan Audit Gate)
9. `.moai/specs/SPEC-V3R2-WF-003/spec.md:120,158,205-206` (cross-SPEC mode-flag contract)
10. `.moai/specs/SPEC-V3R2-WF-001/spec.md:1-30` (blocked-by SPEC, status: completed)
11. `internal/template/commands_audit_test.go:11-50` (audit-test scaffold)
12. `.claude/skills/moai/workflows/fix.md:33,159-163` (deterministic agent dispatch table)
13. `.claude/skills/moai/workflows/mx.md:71-149` (3-Pass canonical Agentless example)
14. `.moai/config/sections/quality.yaml:1-3` (`development_mode: tdd` for run phase)

Total: 14 distinct file:line anchors (exceeds the §Hard-Constraints minimum of 10 for plan.md).

---

## 4. Technology Stack Constraints

Per `spec.md` §2.2 Out of Scope, **no new technology** is introduced:

- No new Go modules or external libraries.
- No new directory structure (uses existing `.claude/skills/`, `.claude/rules/`, `internal/template/`).
- No language-specific tooling beyond the existing Go test framework.
- Embedded-template machinery (`go:embed` in `internal/template/embed.go`) is reused as-is.

The only additive surfaces are:
- One new Go test file (`internal/template/agentless_audit_test.go`) using the standard `testing` package.
- One new section heading in `spec-workflow.md`.
- One new Pipeline Contract section per utility skill (5 instances).
- One new Mode Flag Compatibility section per implementation skill (4 instances).
- One CHANGELOG entry.

---

## 5. Risk Analysis & Mitigations

Extends `spec.md` §8 risks with concrete file-path mitigations.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| `spec-workflow.md` section insertion conflicts with concurrent SPEC-V3R2-WF-003 work | M | M | Run-phase coordination: WF-004 lands its `## Subcommand Classification` section first (this SPEC blocks WF-003 per `spec.md` §9.2). WF-003's run phase appends `--mode` axis details as a sub-section of WF-004's matrix. |
| Forbidden-pattern regex generates false positives on legitimate executor delegation | M | L | Regex set in `agentless_audit_test.go` matches *control-flow verbs only* (decide/orchestrate/route/dispatch); allowed verbs (scan/analyze/fix/verify/measure) are not flagged. Test §6.2 in research.md documents the verb allowlist. |
| Skill body bloat from new sections | L | L | Each new section is ≤25 lines. Skill body sizes (`fix.md` 292, `coverage.md` 209, `mx.md` 226, `codemaps.md` 150, `clean.md` 239 lines per research.md §8) have ample margin. |
| User confused by `/moai fix --mode team` being silently ignored | M | M | M2 Pipeline Contract section explicitly documents the ignore behavior with the `MODE_FLAG_IGNORED_FOR_UTILITY` key. The info-level log itself is delivered by the orchestrator at runtime (no Go-side runtime change in this SPEC's scope). |
| Run phase touches `feedback`/`review`/`e2e` skills accidentally | L | M | Tasks (this plan §6, tasks.md M2/M3) enumerate exactly 5 utility + 4 implementation skills. The classification matrix (§4 of spec-workflow.md insertion) explicitly documents that the 3 unlisted are out of scope. |
| Embedded-template mismatch (skill edited but `internal/template/templates/` not synced) | M | H | M2/M3/M4 each end with `make build` step. CI runs `go test ./...` which exercises embedded FS. Per `CLAUDE.local.md` §2 Template-First Rule HARD constraint. |
| BC-V3R2-007 perceived as behavioral break by downstream consumers | L | M | research.md §4 confirms NO existing utility violates REQ-WF004-004 — the BC is contractual, not runtime. spec.md §10 explicitly documents "BC 영향: 없음 (기존 utility subcommand behavior 보존, 분류만 규정)." |
| Pipeline regression CI test becomes flaky on Windows runners | L | L | `go test` on embedded FS uses platform-neutral `fs.WalkDir`. Pattern: `commands_audit_test.go:25-39`. Same code paths run identically on linux-amd64, macos-arm64, windows-amd64 (verified in existing Test (Ubuntu/macOS/Windows) CI matrix). |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md:627` (mx_plan in plan.md MANDATORY).

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/template/agentless_audit_test.go:TestAgentlessUtilityNoLLMControlFlow` | `@MX:ANCHOR fan_in=5 - SPEC-V3R2-WF-004 REQ-WF004-013 enforcer; guards 5 utility skills against LLM-dispatch regression. Touching this test signature affects the contract for fix/coverage/mx/codemaps/clean.` | The test *is* the contract enforcer. Future PRs that change utility skill bodies will fail or pass this test — high downstream impact. |
| `.claude/rules/moai/workflow/spec-workflow.md` `## Subcommand Classification` section | `@MX:ANCHOR fan_in=9 - Subcommand classification single source of truth; cross-referenced by 9 workflow skills. Changes here affect all workflow contracts.` | The matrix is referenced by every workflow skill via the cross-link added in M5. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| Each of `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` Pipeline Contract section header | `@MX:NOTE - Agentless classification per SPEC-V3R2-WF-004; localize→repair→validate contract. See spec-workflow.md#subcommand-classification.` | Carries the WHY for future readers — explains why this skill cannot adopt LLM-driven phase dispatch. |
| `internal/template/agentless_audit_test.go` package-level doc comment | `@MX:NOTE - Audit suite for SPEC-V3R2-WF-004 Agentless contract. Three tests: TestAgentlessUtilityNoLLMControlFlow (REQ-WF004-013), TestUtilitySkillsContainModeFlagIgnoredSentinel (REQ-WF004-011), TestImplementationSkillsContainPipelineRejectionSentinel (REQ-WF004-014).` | Documents the test scope so future maintainers do not delete or reduce coverage. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `.claude/skills/moai/workflows/fix.md` Phase 3 Auto-Fix section header (existing) | `@MX:WARN @MX:REASON - Future PRs may be tempted to add LLM-driven Level-to-agent dispatch here. The current static lookup table (lines 159-163) MUST remain a fixed mapping. Any LLM-decided dispatch fails TestAgentlessUtilityNoLLMControlFlow.` | Most likely point of future regression per research.md §6. Pre-warns maintainers. |
| `.claude/skills/moai/workflows/clean.md` Phase 4 Safe Removal section | `@MX:WARN @MX:REASON - Phase 4 delegates to expert-refactoring subagent for *executor* role. Do NOT extend this delegation to choose between Phase 4 and Phase 5; that would violate REQ-WF004-004.` | Clean has 5+ phases — most temptation surface for cross-phase dispatch. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

This SPEC is a classification publication, not a feature build. No `@MX:TODO` markers are planned — all work converges to GREEN within M1-M4. Any `@MX:TODO` introduced during implementation must be resolved before M5 (per `.claude/rules/moai/workflow/mx-tag-protocol.md` GREEN-phase resolution rule).

### 6.5 MX tag count summary

- @MX:ANCHOR: 2 targets
- @MX:NOTE: 6 targets (1 per utility skill section + 1 audit test package doc)
- @MX:WARN: 2 targets
- @MX:TODO: 0 targets
- **Total**: 10 MX tag insertions planned across 8 distinct files

---

## 7. Worktree Discipline

[HARD] All run-phase work for SPEC-V3R2-WF-004 MUST execute in:

```
/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-004
```

Branch: `feature/SPEC-V3R2-WF-004` (already created, base of `origin/main` HEAD `a3be99e67`).

[HARD] All Read/Write/Edit tool invocations MUST use absolute paths under this worktree root. Tool calls referencing `/Users/goos/MoAI/moai-adk-go/` (the main session checkout) are PROHIBITED for this SPEC.

[HARD] `make build` and `go test ./...` execute inside the worktree with `cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-004 && make build` etc. Per `CLAUDE.local.md` §15 worktree isolation conventions.

The run phase SHOULD NOT enter the worktree via `cd` from a long-running shell; instead each Bash invocation uses an absolute-path command or `cd <worktree> && ...`.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:172-204`). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.2.0 schema** — `spec.md` frontmatter has all 9 required fields (`id`, `title`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`). Verified by reading `spec.md:2-11` (lines 1-24 inclusive of v0.2.0 migration).
- [x] **C2: HISTORY entry for v0.2.0** — `spec.md:30-33` HISTORY table has v0.2.0 row with description.
- [x] **C3: 15 EARS REQs across 5 categories** — `spec.md:104-157` (Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Complex 3).
- [x] **C4: 13 ACs all map to REQs (100% coverage)** — `spec.md:163-175`. Each AC explicitly cites the REQ(s) it maps to.
- [x] **C5: BC scope clarity** — `spec.md:223` ("BC 영향: 없음") + research.md §4 (verified per-utility no-violation table).
- [x] **C6: File:line anchors ≥10** — research.md §11 (24 anchors), this plan.md §3.4 (14 anchors).
- [x] **C7: Exclusions section present** — `spec.md:66-73` Out of Scope explicitly enumerates 7 exclusions.
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml:1-3`.
- [x] **C9: mx_plan section** — this plan §6 (10 MX tag insertions across 4 categories).
- [x] **C10: Risk table with mitigations** — `spec.md:191-197` + this plan §5 (8 risks, file-anchored mitigations).
- [x] **C11: Worktree absolute path discipline** — this plan §7 (3 HARD rules).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, and YAML/Markdown templates. No Go function bodies or executable code.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md §1-13.
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md §M1-M5 (manager-tdd / expert-backend / manager-docs assignments).
- [x] **C15: Cross-SPEC consistency** — WF-003 sibling SPEC `MODE_PIPELINE_ONLY_UTILITY` sentinel string matches; WF-001 blocked-by dependency status: completed.

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: Create `agentless_audit_test.go` with 3 tests. Confirm RED.
2. **M2 (P0)**: Add Pipeline Contract sections to 5 utility skills (+ embedded template parity + `make build`). Confirm 1 of 3 tests turns GREEN.
3. **M3 (P0)**: Add Mode Flag Compatibility sections to 4 implementation skills (+ template parity + `make build`). Confirm 2 of 3 tests turn GREEN.
4. **M4 (P1)**: Add `## Subcommand Classification` section to `spec-workflow.md` (+ template parity + `make build`). Confirm 3 of 3 tests GREEN, full `go test ./...` passes.
5. **M5 (P2)**: Add CHANGELOG entry, cross-link footers in 9 skills, MX tags per §6, update `progress.md`, final `go test ./...` + `make build`.

Total milestones: 5. Total file edits (existing): 11. Total file creations (new): 1. Total CHANGELOG entries: 1.

---

End of plan.md.
