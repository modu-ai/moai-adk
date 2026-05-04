# SPEC-V3R2-WF-004 Deep Research (Phase 0.5)

> Research artifact for `Agentless Fixed-Pipeline Classification for Utility Subcommands`.
> Companion to `spec.md` (v0.2.0). Authored against worktree `feature/SPEC-V3R2-WF-004`.

## HISTORY

| Version | Date       | Author                        | Description                                                              |
|---------|------------|-------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-02 | MoAI Plan Workflow (run-phase research)| Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5 |

---

## 1. Goal of Research

Substantiate `spec.md` §1 (Goal) and §4 (Assumptions) with concrete file:line evidence so that the run phase can implement REQ-WF004-001..015 against a known-good baseline. The research answers four questions:

1. Where does the *control flow* of each utility subcommand actually live?
2. Does any of the five utilities currently call `Agent()` for **control flow** (vs delegated step execution)?
3. How does the existing pre/post hook fixed pipeline (R6 §2) already approximate Agentless?
4. Which artifacts must be touched to publish the classification matrix and CI guard required by REQ-WF004-005 and REQ-WF004-013?

---

## 2. Architectural Anatomy of `/moai {fix,coverage,mx,codemaps,clean}`

### 2.1 Slash command surface (`.claude/commands/moai/`)

All five user-invocable surfaces are **thin wrappers** that route into a single `moai` skill, per the Thin Command Pattern (`.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md` §1; `internal/template/commands_audit_test.go:11-50`).

| Subcommand   | File                                                    | Body (lines 5-7) |
|--------------|---------------------------------------------------------|---|
| `/moai fix`      | `.claude/commands/moai/fix.md:1-7`           | `Use Skill("moai") with arguments: fix $ARGUMENTS` |
| `/moai coverage` | `.claude/commands/moai/coverage.md:1-7`      | `Use Skill("moai") with arguments: coverage $ARGUMENTS` |
| `/moai mx`       | `.claude/commands/moai/mx.md:1-7`            | `Use Skill("moai") with arguments: mx $ARGUMENTS` |
| `/moai codemaps` | `.claude/commands/moai/codemaps.md:1-7`      | `Use Skill("moai") with arguments: codemaps $ARGUMENTS` |
| `/moai clean`    | `.claude/commands/moai/clean.md:1-7`         | `Use Skill("moai") with arguments: clean $ARGUMENTS` |

Implication: There is **no `internal/cli/{fix,coverage,codemaps,clean}.go` orchestrator** for these five subcommands; the only Go-side `moai` CLI commands with overlapping names are utility tools (e.g., `internal/cli/mx.go:1-29` exposes `moai mx query` for tag sidecar lookups, not the `/moai mx` slash command flow). Control flow for the slash-command experience lives entirely in the workflow skill markdown that Claude executes.

### 2.2 Workflow skill control flow (`.claude/skills/moai/workflows/`)

Each utility skill declares an explicit phase ordering with [HARD] delegation directives. The phase sequence — already deterministic — is the latent Agentless contract that REQ-WF004-003 formalizes.

#### 2.2.1 `/moai fix` — `.claude/skills/moai/workflows/fix.md`

- Flow declaration (`fix.md:33`): `Parallel Scan -> Classify -> Fix -> Verify -> Report`
- Phase 1 Parallel Scan (`fix.md:46-124`): three deterministic scanners (LSP / AST-grep / Linter) launched via `Bash run_in_background`. No `Agent()` invocation in the scan phase itself.
- Phase 2 Classification (`fix.md:126-132`): pure rule-based bucketing into Levels 1-4. No LLM in control flow.
- Phase 2.5 Pre-Fix MX Context Scan (`fix.md:134-153`): Read-only enrichment of the issue list with @MX tag context. Deterministic.
- Phase 3 Auto-Fix (`fix.md:156-170`): `[HARD] Agent delegation mandate` — work executors (`expert-backend`, `expert-frontend`, `expert-refactoring`, `expert-debug`) are invoked **per Level**, but the Level-to-agent mapping (`fix.md:159-163`) is a static lookup table, not an LLM-decided dispatch. This is **Agentless-compatible** because the LLM picks neither the next phase nor the executor; the contract picks both.
- Phase 4 Verification + 4.5 MX Update + 4.6 Dead Code Cleanup (`fix.md:172-225`): deterministic re-run, tag mutation, and optional delegation to the `clean` workflow.
- Resume contract (`fix.md:241-254`): snapshot-driven resume already mirrors Agentless's stateless re-entry pattern.

Mapping to 3-phase contract per `spec.md` §4:
- localize ← Phase 1 Parallel Scan + Phase 2 Classification + Phase 2.5 MX Context Scan
- repair ← Phase 3 Auto-Fix
- validate ← Phase 4 Verification (+ Phase 4.5/4.6 incidental side effects)

#### 2.2.2 `/moai coverage` — `.claude/skills/moai/workflows/coverage.md`

- Flow declaration (`coverage.md:33`): `Measure Coverage -> Identify Gaps -> Generate Tests -> Verify -> Report`
- Phase 1 Measurement (`coverage.md:44-64`): `[HARD] Delegate coverage measurement to the expert-testing subagent.` Single-agent delegation per phase, no LLM dispatch.
- Phase 2 Gap Analysis (`coverage.md:66-110`): `[HARD] Delegate gap analysis to the expert-testing subagent.` MX tag pre-scan informs P1/P2 priority deterministically.
- Phase 3 Test Generation (`coverage.md:112-141`): TDD/DDD strategy chosen by `quality.yaml development_mode` config — config-driven, not LLM-driven (`coverage.md:118-128`).
- Phase 4 Verification (`coverage.md:143-148`): Deterministic re-measure + diff.
- Phase 5 Report (`coverage.md:150-178`): Static template + AskUserQuestion next-steps.

Mapping: localize ← Phase 1+2; repair ← Phase 3; validate ← Phase 4.

#### 2.2.3 `/moai mx` — `.claude/skills/moai/workflows/mx.md`

This is the closest pre-existing match to Agentless. The skill declares an explicit 3-Pass scan (`mx.md:71-149`):

- Phase 0 Codebase Discovery (`mx.md:73-110`): language detection by indicator file, project context loading from `.moai/project/{tech,structure,product}.md`.
- Pass 1 Full File Scan (`mx.md:112-122`): glob → fan-in → complexity → priority queue P1-P4. Deterministic.
- Pass 2 Selective Deep Read (`mx.md:124-138`): Read-only enrichment for P1+P2 files.
- Pass 3 Batch Edit (`mx.md:140-149`): one Edit per file, all tags inserted in single op.

No `Agent()` delegation appears in any pass; the work is performed by the orchestrator using `Glob`/`Grep`/`Read`/`Edit` tools. Mapping is direct: localize ← Pass 1; repair ← Pass 3 (Pass 2 is preparatory read for Pass 3); validate ← post-edit MX scan run by `postToolHandler` (see §3 below).

#### 2.2.4 `/moai codemaps` — `.claude/skills/moai/workflows/codemaps.md`

- Flow declaration (`codemaps.md:33`): `Explore Codebase -> Analyze Architecture -> Generate Maps -> Verify -> Report`
- Phase 1 (`codemaps.md:42-64`): `[HARD] Delegate codebase exploration to the Explore subagent.`
- Phase 2 (`codemaps.md:66-82`): `[HARD] Delegate architecture analysis to the manager-docs subagent.`
- Phase 3 (`codemaps.md:84-103`): `[HARD] Delegate map generation to the manager-docs subagent.`
- Phase 4 (`codemaps.md:105-110`): Verification — orchestrator-side existence and consistency checks.
- Phase 5 (`codemaps.md:112-123`): Report + AskUserQuestion next-steps.

Mapping: localize ← Phase 1; repair ← Phase 2+3; validate ← Phase 4.

#### 2.2.5 `/moai clean` — `.claude/skills/moai/workflows/clean.md`

- Flow declaration (`clean.md:33`): `Static Analysis -> Usage Graph -> Classification -> Safe Removal -> Test Verification -> Report`
- Phase 1 (`clean.md:44-65`): `[HARD] Delegate static analysis to the expert-refactoring subagent.`
- Phase 2 (`clean.md:67-97`): `[HARD] Delegate usage graph analysis to the expert-refactoring subagent.` MX tag cross-check (`clean.md:85-93`) is rule-driven reclassification.
- Phase 3 (`clean.md:99-130`): User approval via AskUserQuestion (orchestrator side).
- Phase 4 (`clean.md:132-150`): `[HARD] Delegate removal to the expert-refactoring subagent.`
- Phase 5 (`clean.md:152-166`): `[HARD] Delegate test verification to the expert-testing subagent.`
- Phase 5.5 (`clean.md:168-174`): MX tag cleanup.

Mapping: localize ← Phase 1+2; repair ← Phase 4; validate ← Phase 5 (+ 5.5 incidental).

### 2.3 Implementation skill comparison (`plan`, `run`, `sync`, `design`)

For contrast, the four implementation skills explicitly retain LLM-driven control flow and per-phase mode selection per SPEC-V3R2-WF-003:

- `.claude/skills/moai/workflows/plan.md` (36,153 bytes) — Phase 0/0.5/1A/1B/2/3/4 with `manager-spec` agent, conditional plan-auditor entry, annotation cycle (1-6 user iterations), branching by `--mode`.
- `.claude/skills/moai/workflows/run.md` (47,073 bytes) — Phase 0.5 Plan Audit Gate, Phase 1-N driven by `quality.yaml` development_mode (DDD vs TDD), Drift Guard re-planning trigger (`.claude/rules/moai/workflow/spec-workflow.md:106-129`).
- `.claude/skills/moai/workflows/sync.md` (48,933 bytes) — Phase 0.6.1 language detection, multi-agent doc sync via `manager-docs`, PR creation orchestration.
- `.claude/skills/moai/workflows/design.md` (10,565 bytes) — Path A (Claude Design) vs Path B (code-based) router; GAN loop with evaluator-active.

These four are the structural complement — they are open-ended in the `R1 §25` sense: task structure is **not** known in advance.

---

## 3. The pre-existing pre/post hook fixed pipeline (R6 §2)

`spec.md` §1.1 cites R6 §2.1-§2.4 as evidence that the system is already "준-Agentless." Confirmed by direct read:

- `internal/hook/post_tool.go` (509 LOC, R6 §2.2 verdict KEEP) implements **fixed-order** MX validation → LSP check → metrics emit per write op. The order is hard-coded in Go, not LLM-decided. (`r6-commands-hooks-style-rules.md:127-129`).
- `internal/hook/pre_tool.go` (652 LOC) similarly enforces security scan → secrets detection → reflective-write guard before any Write/Edit reaches disk (`r6-commands-hooks-style-rules.md:128`).
- The shell wrapper layer (`r6-commands-hooks-style-rules.md:80-114`) — 26 wrappers, each ~30 LOC — is a deterministic pass-through to `moai hook <event>`. Zero LLM in this path.
- `r6-commands-hooks-style-rules.md:122` confirms 27 handlers registered in `internal/cli/deps.go:152-186` with deterministic dispatch.

Implication: The Agentless pattern's **localize → repair → validate** rhythm is already enforced *transversally* by the hook system on every Write/Edit. The five utility subcommands run *on top of* that pipeline. SPEC-V3R2-WF-004 makes the rhythm explicit at the subcommand layer.

---

## 4. BC-V3R2-007 scope analysis (does any utility currently violate?)

`spec.md` §10 states: *"BC 영향: 없음 (기존 utility subcommand behavior 보존, 분류만 규정)."* Verifying:

| Subcommand | Current LLM-driven control-flow decisions? | Verdict |
|------------|---------------------------------------------|---------|
| `/moai fix`      | Phase ordering hard-coded (`fix.md:33`); agent picked by Level lookup (`fix.md:159-163`); MX context scan optional but rule-gated (`fix.md:151`); team mode is opt-in via `--team` flag, not LLM choice. | NO violation. Compliant with REQ-WF004-004. |
| `/moai coverage` | Phase ordering hard-coded (`coverage.md:33`); single agent (`expert-testing`) handles all phases; mode chosen by `quality.yaml development_mode` (`coverage.md:118`). | NO violation. |
| `/moai mx`       | Pure orchestrator-side execution; no `Agent()` delegation in main flow. `--team` flag (`mx.md:59`) is parallel scan, not control flow. | NO violation. |
| `/moai codemaps` | Phase ordering hard-coded (`codemaps.md:33`); two static agents (`Explore` then `manager-docs`). | NO violation. |
| `/moai clean`    | Phase ordering hard-coded (`clean.md:33`); deterministic agent rotation per phase; user approval via AskUserQuestion is allowed (orchestrator-side, not subagent). | NO violation. |

**Confirmation: NO existing utility subcommand violates REQ-WF004-004.** BC-V3R2-007's *flip* is therefore a **declaration-level** breaking change (the contract is newly asserted), not a behavioral break — consistent with `spec.md` §10. Test mass: codifying the existing behavior, not changing it.

Caveat for run phase: REQ-WF004-013 (`AGENTLESS_CONTROL_FLOW_VIOLATION`) is a **regression guard** for future PRs that might introduce Agent() into control-flow position — i.e., a `Use the X subagent to decide which phase runs next` pattern. Today's skills do not exhibit that; the test must be written so it stays green now and turns red if such a pattern is ever introduced.

---

## 5. Workflow-modes publication target

`spec.md` §5.1 REQ-WF004-005 specifies the matrix MUST land in `.claude/rules/moai/workflow/workflow-modes.md` **or** its merge target `spec-workflow.md` per R6 §4.5 recommendation.

State of the worktree as of `git log --oneline -1` = base-of-`origin/main` (`a3be99e67`):

- `.claude/rules/moai/workflow/workflow-modes.md` does **NOT** yet exist (`ls .claude/rules/moai/workflow/` returns 6 files: `context-window-management.md`, `moai-memory.md`, `mx-tag-protocol.md`, `spec-workflow.md`, `team-pattern-cookbook.md`, `worktree-integration.md`).
- `.claude/rules/moai/workflow/spec-workflow.md:1-300` already documents Plan/Run/Sync phase contracts and Phase 0.5 Plan Audit Gate. It is the natural single-source location.

Decision (recommendation for run phase): publish the classification matrix as a **new section inside `spec-workflow.md`** (not a separate `workflow-modes.md`). Rationale:

- Avoids the R6 §4.5 merge-target ambiguity that `spec.md` §8 risks table flagged.
- Keeps the workflow contract single-file (one source of truth, one merge conflict surface).
- SPEC-V3R2-WF-003 also writes its `--mode` matrix here per its own §11 Constraints; co-locating WF-003 + WF-004 matrices avoids cross-file desync.

Run-phase agent should add a section `## Subcommand Classification (Pipeline vs Multi-Agent)` to `spec-workflow.md` near the existing `## Phase Overview` table (around line 9-17). Cross-link from each of the 9 skill files via a one-line "see `spec-workflow.md#subcommand-classification`" reference.

---

## 6. CI guard for AGENTLESS_CONTROL_FLOW_VIOLATION (REQ-WF004-013)

### 6.1 Existing audit-test scaffold

`internal/template/commands_audit_test.go:1-60` already implements an embedded-fs walker that:

1. Collects all `.md` / `.md.tmpl` files under `.claude/commands` in the embedded template FS (`commands_audit_test.go:23-39`).
2. Validates frontmatter + body line-count per file (Thin Command Pattern check) (`commands_audit_test.go:53+`).

This is the canonical pattern for static rule enforcement against the embedded skill/command corpus. The new test follows the same shape, walking `.claude/skills/moai/workflows/{fix,coverage,mx,codemaps,clean}.md` and asserting **absence of LLM-dispatch patterns**.

### 6.2 Forbidden patterns (test will fail if any matches in the 5 utility skills)

The test should fail if any of the following appear in the *body* of a utility skill (excluding code blocks and unambiguous executor delegation in [HARD] blocks):

| Forbidden pattern (regex, case-insensitive) | Why it's a control-flow agent |
|----------------------------------------------|-------------------------------|
| `Use the .* subagent to (decide|determine|choose|select|orchestrate|route|dispatch)` | LLM picks next step |
| `Use the .* subagent to (plan|design) the (pipeline|workflow|next phase|sequence)` | LLM plans the pipeline itself |
| `delegate to .* (orchestrator|router|dispatcher|controller)` | Same as above, alternative phrasing |
| `manager-strategy.*subagent.*(branch|fork|conditional)` | Strategy agent in control-flow position |

Allowed (must NOT trigger): `Use the {expert-X|Explore|manager-docs} subagent to {scan|analyze|generate|fix|test|verify|measure|remove}` — these are **executor** delegations within a deterministic phase.

### 6.3 Recommended test location

New file: `internal/template/agentless_audit_test.go`. Mirror structure of `commands_audit_test.go:1-60`. Single test func `TestAgentlessUtilityNoLLMControlFlow` plus subtests per skill file.

### 6.4 `--mode` flag rejection tests

REQ-WF004-011 (utility ignores `--mode`) and REQ-WF004-014 (implementation rejects `--mode pipeline`) are **runtime** assertions. Since utility subcommands have no Go handler (they are skill-delegated), enforcement must live in the workflow skill body itself: each utility skill must add a Pipeline contract header that documents flag-ignore behavior. The corresponding integration test is therefore a **skill-content audit** (the skill body must contain the documented sentinel string `MODE_FLAG_IGNORED_FOR_UTILITY`), not a Go runtime test.

For REQ-WF004-014 / AC-WF004-11, similar logic: implementation skills (`plan.md`, `run.md`, `sync.md`, `design.md`) must contain a documented rejection clause for `--mode pipeline` referencing `MODE_PIPELINE_ONLY_UTILITY` (this string is shared with `SPEC-V3R2-WF-003 REQ-WF003-016` — see `.moai/specs/SPEC-V3R2-WF-003/spec.md:158`).

---

## 7. Reference patterns to preserve

The run phase MUST NOT touch the following — they encode the rhythm WF-004 is formalizing:

- **Reference**: `.claude/skills/moai/workflows/fix.md:33` — `Parallel Scan -> Classify -> Fix -> Verify -> Report` flow line. This is the canonical 3-phase contract for `fix`. Header insertion in M2 must NOT replace or contradict this line.
- **Reference**: `.claude/skills/moai/workflows/coverage.md:33` — `Measure Coverage -> Identify Gaps -> Generate Tests -> Verify -> Report` line.
- **Reference**: `.claude/skills/moai/workflows/mx.md:71-149` — 3-Pass scan structure (the Agentless poster child).
- **Reference**: `.claude/skills/moai/workflows/codemaps.md:33` — `Explore Codebase -> Analyze Architecture -> Generate Maps -> Verify -> Report`.
- **Reference**: `.claude/skills/moai/workflows/clean.md:33` — `Static Analysis -> Usage Graph -> Classification -> Safe Removal -> Test Verification -> Report` (5+ phases, but mappable to 3 macro-phases).
- **Reference**: `internal/hook/post_tool.go` (referenced via R6 `r6-commands-hooks-style-rules.md:127-129`) — 509-LOC handler that already runs MX/LSP/metrics in fixed order. WF-004 must not change this.
- **Reference**: `internal/template/commands_audit_test.go:11-50` — Thin Command Pattern audit test scaffold for the new audit test in §6.

---

## 8. Risks (extends `spec.md` §8 with run-phase mitigations)

| Risk | Probability | Impact | File:Line Mitigation Anchor |
|------|-------------|--------|-----------------------------|
| Future PR adds `Agent()` for control-flow in a utility skill | M | H | `internal/template/agentless_audit_test.go` (new, per §6.3) catches via static text scan in CI |
| `spec-workflow.md` insertion conflicts with WF-003 matrix landing | L | M | Run phase coordinates: WF-004 inserts `## Subcommand Classification` heading at line ~17 (immediately after `## Phase Overview` table, before `## Plan Phase`); WF-003 lands its `--mode` matrix as a child section |
| Skill body header bloat from "Pipeline Contract" addition | L | L | Header is ≤10 lines per file; skill bodies (`fix.md` 292 lines, `coverage.md` 209 lines, `mx.md` 226 lines, `codemaps.md` 150 lines, `clean.md` 239 lines) have margin |
| User confused by `/moai fix --mode team` being silently ignored | M | M | REQ-WF004-011 mandates `MODE_FLAG_IGNORED_FOR_UTILITY` *info-level* log; M2 step adds the explicit user-facing note in each utility skill's `## Supported Flags` section |
| Forbidden-pattern regex (§6.2) generates false positives on existing legitimate executor delegation | M | L | Test §6.2 uses *inclusion* allowlist for `{expert-X\|Explore\|manager-docs}` + verb allowlist `{scan\|analyze\|generate\|fix\|test\|verify\|measure\|remove}`; only flags *control-flow* verbs |
| Run phase touches `feedback`/`review`/`e2e` skills accidentally | L | M | `spec.md` §1.2 + §2.2 explicitly exclude these; tasks.md M2 enumerates exactly the 5 in scope |

---

## 9. Recommendations for the Run Phase

1. **Methodology**: Per `.moai/config/sections/quality.yaml development_mode`, this project uses TDD. The CI guard test (REQ-WF004-013) and any skill-body audit (REQ-WF004-011 / REQ-WF004-014) MUST be written first (RED), confirmed failing, and then the skill bodies + classification matrix added (GREEN). Refactor pass at the end consolidates duplication across the 5 utility skill headers.
2. **Worktree discipline**: All run-phase work continues in `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-004`. Branch `feature/SPEC-V3R2-WF-004` is already created. Worktree absolute path is the **only** legal write surface.
3. **MX targets**: Pre-mark `internal/template/agentless_audit_test.go` with `@MX:ANCHOR` (the test becomes the contract enforcer — high fan_in for future regressions). Pre-mark the new `## Subcommand Classification` section anchor in `spec-workflow.md` with `@MX:NOTE` documenting why pipeline vs multi-agent split exists.
4. **Quality gate**: Per `.claude/rules/moai/workflow/spec-workflow.md:172-204` Phase 0.5 Plan Audit Gate, plan-auditor will audit this research + plan + acceptance + tasks before the implementation phase begins. Treat each artifact as audit-ready output.
5. **No Agent() in this run phase**: Ironically, the SPEC's content (Agentless classification) is itself a fitting test of the principle. Run phase should be performed by a single agent (manager-tdd or direct orchestrator execution) with no LLM-driven dispatch — exactly what the SPEC mandates for the targets.
6. **Cross-SPEC sync**: Once WF-004 lands, notify SPEC-V3R2-WF-003 (still draft) that REQ-WF003-016's `MODE_PIPELINE_ONLY_UTILITY` error path is now anchored in the WF-004 matrix. WF-003's run phase can reference `spec-workflow.md#subcommand-classification` directly.

---

## 10. v3 Master design cross-anchors (verbatim)

Per `spec.md` §10 traceability, the master design document expresses the WF-004 contract in three places. Quoting the relevant anchor lines for run-phase reference (paragraphs are short and load-bearing):

- **`docs/design/major-v3-master.md:L1069`** — §11.6 WF-004 definition: defines the SPEC's identity as the "Agentless Fixed-Pipeline Classification" and lists the five subcommands. This is the canonical scope statement.
- **`docs/design/major-v3-master.md:L966`** — §8 BC-V3R2-007 Agentless flip: documents the breaking change row stating that utility subcommands transition from "informally agent-using" to "formally Agentless-classified," explicitly noting BC impact is contractual not behavioral (matches our §4 conclusion above).
- **`docs/design/major-v3-master.md:L993`** — §9 Phase 6 Multi-Mode Workflow: positions WF-004 as a sibling deliverable to WF-003 in the same phase, sharing the `--mode` flag matrix.

Pattern-library `§O-6` and R1 §25 are reproduced inline in `spec.md` §1.1 and need no further extraction.

---

## 11. Citation Summary (file:line anchors used)

1. `spec.md:1-237` (entire SPEC, contract source)
2. `.claude/commands/moai/fix.md:1-7`
3. `.claude/commands/moai/coverage.md:1-7`
4. `.claude/commands/moai/mx.md:1-7`
5. `.claude/commands/moai/codemaps.md:1-7`
6. `.claude/commands/moai/clean.md:1-7`
7. `.claude/skills/moai/workflows/fix.md:33,46-124,126-132,134-153,156-170,172-225,241-254`
8. `.claude/skills/moai/workflows/coverage.md:33,44-64,66-110,112-141,143-148,150-178`
9. `.claude/skills/moai/workflows/mx.md:33,71-149`
10. `.claude/skills/moai/workflows/codemaps.md:33,42-64,66-82,84-103,105-110,112-123`
11. `.claude/skills/moai/workflows/clean.md:33,44-65,67-97,99-130,132-150,152-166,168-174`
12. `.claude/rules/moai/workflow/spec-workflow.md:1-300`
13. `.claude/rules/moai/workflow/spec-workflow.md:106-129` (Drift Guard / Re-planning Gate)
14. `.claude/rules/moai/workflow/spec-workflow.md:172-204` (Plan Audit Gate)
15. `.moai/design/v3-redesign/synthesis/pattern-library.md:127-134` (§O-6)
16. `.moai/design/v3-redesign/research/r1-ai-harness-papers.md:194-199` (§25 Agentless)
17. `.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md:80-114` (§2.1 hooks)
18. `.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md:120-153` (§2.2 Go handlers)
19. `internal/cli/mx.go:1-29` (utility CLI tool, NOT slash-command path — disambiguation)
20. `internal/template/commands_audit_test.go:1-60` (audit-test scaffold pattern)
21. `internal/hook/post_tool.go` (referenced via r6-commands-hooks-style-rules.md:127)
22. `.moai/specs/SPEC-V3R2-WF-003/spec.md:51,157-158,179-180,206` (cross-SPEC mode-flag contract)
23. `.moai/specs/SPEC-V3R2-WF-001/spec.md:1-30` (blocked-by SPEC, status: completed)
24. `docs/design/major-v3-master.md:L1069,L966,L993` (master design anchors)

Total distinct file:line citations: **24** (exceeds the §Hard-Constraints minimum of 10 for research.md).

---

End of research.md.
