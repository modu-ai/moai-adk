# Implementation Plan — SPEC-V3R4-HARNESS-001

This document is the Wave-level implementation plan for the foundation SPEC of the self-evolving harness v2 system. All priorities use P0/P1/P2/P3 labels per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation; no time estimates are used.

---

## 1. Overview

This SPEC is the foundation for the eight-SPEC self-evolving harness v2 decomposition. Its scope is intentionally narrow: lock in the CLI-retirement contract, the slash-command-only lifecycle contract, the 5-Layer Safety preservation, the FROZEN zone immutability, and the AskUserQuestion-only orchestrator contract — all as architectural baselines that downstream SPECs 002-008 will build upon.

The plan decomposes into five Waves (A through E). Each Wave is independently PR-able in a fully-implemented downstream SPEC, but this foundation SPEC only ships SPEC artifacts (four markdown files); the implementation Waves are scaffolded here as forward-looking guidance for run-phase delegation, not as in-scope deliverables of this manager-spec PR.

| Wave | Title | Priority | Owning Run-Phase Agent | Acceptance Gate |
|------|-------|----------|------------------------|-----------------|
| Wave A | Skill body and workflow consolidation | P0 | `manager-develop` | `/moai:harness <verb>` reachable through skill workflow body without binary invocation. |
| Wave B | Subagent reorganization and contract preservation | P0 | `manager-develop` | `moai-harness-learner` and meta-harness subagents inherit V3R4 contracts verbatim. |
| Wave C | Hook contract and observer baseline preservation | P1 | `manager-develop` (delegating to `expert-devops`) | PostToolUse observer remains operational; `learning.enabled: false` no-op verified. |
| Wave D | CLI deprecation markers and CI guard | P0 | `manager-develop` (delegating to `expert-backend`) | CI test fails on hypothetical re-registration; `internal/cli/harness.go` annotated. |
| Wave E | Documentation and superseded SPEC follow-up coordination | P1 | `manager-docs` and `manager-git` | CHANGELOG, release notes drafted; V3R3 status-transition commit deferred to post-merge follow-up. |

All five Waves are documented here as the run-phase roadmap. This `/moai plan` invocation produces the four SPEC artifacts only; run-phase work happens in `/moai run SPEC-V3R4-HARNESS-001` after this plan PR merges.

---

## 2. Architecture Overview

### 2.1 Current State (pre-V3R4)

```
+--------------------------------------------------------------+
| User                                                         |
+--+------------------------------+----------------------------+
   |                              |
   | terminal                     | Claude Code session
   |                              |
   v                              v
+-------------------+   +-------------------------+
| moai (Go binary)  |   | /moai:harness slash cmd |
| - root.go         |   | - thin wrapper          |
| - harness.go      |   |   (Skill("moai")        |
|   (newHarnessCmd  |   |    arguments: harness)  |
|    defined, NOT   |   +-----------+-------------+
|    registered)    |               |
+-------------------+               v
                          +---------------------+
                          | moai skill          |
                          | workflows/harness.md|
                          +-----------+---------+
                                      |
                                      v
                          +-----------------------+
                          | moai-harness-learner  |
                          | (subagent)            |
                          +-----------+-----------+
                                      |
                          +-----------v-----------+
                          | .moai/harness/        |
                          | - usage-log.jsonl     |
                          | - learning-history/   |
                          | - proposals/          |
                          +-----------------------+
```

Observation: `newHarnessCmd` exists but is not wired into `rootCmd.AddCommand(...)`. The slash command thin wrapper (PR #908, commit `452aa638f`) already routes through the skill workflow without invoking the binary. The foundation SPEC formalizes both decisions.

### 2.2 Target State (post-V3R4 foundation)

```
+--------------------------------------------------------------+
| User                                                         |
+--+------------------------------+----------------------------+
   |                              |
   | terminal                     | Claude Code session
   |                              |
   v                              v
+-------------------+   +-------------------------+
| moai (Go binary)  |   | /moai:harness slash cmd |
| - root.go         |   | - thin wrapper          |
|   (no harness     |   |   (Skill("moai")        |
|    subcommand)    |   |    arguments: harness)  |
| - harness.go      |   +-----------+-------------+
|   (DEPRECATED     |               |
|    marker only)   |               v
+-------------------+   +-------------------------+
   | "unknown        |   | moai skill            |
   |  command"       |   | workflows/harness.md  |
                        | - status              |
                        | - apply (orchestrator |
                        |   AskUserQuestion gate)|
                        | - rollback            |
                        | - disable             |
                        +-----------+-----------+
                                    |
                                    v
                      +-------------------------+
                      | moai-harness-learner    |
                      | (subagent — no          |
                      |  AskUserQuestion calls) |
                      +-----------+-------------+
                                  |
                      +-----------v-----------+
                      | .moai/harness/        |
                      | - usage-log.jsonl     |
                      | - learning-history/   |
                      |   - snapshots/        |
                      |   - tier-promotions/  |
                      |   - frozen-guard-     |
                      |     violations.jsonl  |
                      | - proposals/          |
                      +-----------------------+

+--------------------------------------------------------------+
| 5-Layer Safety (FROZEN — design constitution §5)             |
| L1 Frozen Guard | L2 Canary | L3 Contradiction | L4 Rate     |
| Limiter         | L5 Human Oversight (AskUserQuestion)       |
+--------------------------------------------------------------+

+--------------------------------------------------------------+
| FROZEN zones (path-prefix protected, REQ-HRN-FND-006)        |
| .claude/agents/moai/**                                       |
| .claude/skills/moai-*/**                                     |
| .claude/rules/moai/**                                        |
| .moai/project/brand/**                                       |
+--------------------------------------------------------------+
```

Key changes:
- CLI verb path is now formally retired (REQ-HRN-FND-001) with CI guard (REQ-HRN-FND-002).
- Slash command verbs are all reachable through the skill body alone (REQ-HRN-FND-003).
- Tier-4 application gate uses `AskUserQuestion` issued by the orchestrator only (REQ-HRN-FND-004, -015).
- 5-Layer Safety and FROZEN zones are preserved verbatim (REQ-HRN-FND-005, -006).

### 2.3 Lifecycle Phase Mapping

The harness lifecycle has five phases. Each phase has a clear ownership boundary:

| Phase | Owner | Trigger | Output |
|-------|-------|---------|--------|
| Generation | `moai-meta-harness` skill (called from `/moai project` Phase 5+) | Project init or explicit invocation | `.claude/skills/my-harness-*/`, `.claude/agents/my-harness/*`, `.moai/harness/main.md` |
| Observation | PostToolUse hook (`internal/hook/`) | Every tool invocation when `learning.enabled: true` | `.moai/harness/usage-log.jsonl` (JSONL append) |
| Learning | Tier classifier (in `moai-harness-learner` skill body) | Cumulative count reaches 1 / 3 / 5 / 10 thresholds | `.moai/harness/learning-history/tier-promotions.jsonl` |
| Evolution | `moai-harness-learner` proposal generator | Tier-4 reached after Canary/Contradiction passes | `.moai/harness/proposals/<ISO-DATE>-tier4-NNN.json` |
| Application | MoAI orchestrator + 5-Layer Safety + `AskUserQuestion` | User invokes `/moai:harness apply` | Snapshot at `.moai/harness/learning-history/snapshots/<ISO-DATE>/` + actual modification + `applied/` log |

The foundation SPEC preserves all five phases as they exist in the V3R3 implementation. Downstream SPECs 002-008 evolve the algorithms within Observation / Learning / Evolution; the Application phase's AskUserQuestion gate remains immutable.

---

## 3. Wave Decomposition (Run-Phase Roadmap)

This section describes the run-phase Wave structure. The plan-phase PR (this manager-spec session) ships only the SPEC artifacts; the Wave implementations are downstream `/moai run` work.

### 3.1 Wave A — Skill Body and Workflow Consolidation (P0)

**Goal**: Ensure `.claude/skills/moai/workflows/harness.md` and the `moai` skill body fully implement the four lifecycle verbs (status, apply, rollback, disable) without invoking any Go binary.

**Owner**: `manager-develop`

**Inputs**:
- This SPEC's REQ-HRN-FND-003 (slash-command verb coverage)
- Existing `.claude/commands/moai/harness.md` thin wrapper (already CLI-free per PR #908)
- Existing `.claude/skills/moai/workflows/harness.md` workflow body

**Tasks** (detailed in `tasks.md` as T-A1 through T-A4):
- T-A1: Audit existing `.claude/skills/moai/workflows/harness.md` for any binary invocation references. Remove if found.
- T-A2: Ensure each verb (status, apply, rollback, disable) is documented in the workflow body with explicit AskUserQuestion checkpoints for Tier-4 application (REQ-HRN-FND-004).
- T-A3: Add an explicit cross-reference from the workflow body to the canonical AskUserQuestion protocol (`.claude/rules/moai/core/askuser-protocol.md`).
- T-A4: Add cross-reference annotations to the moai-harness-learner SKILL.md noting its subordination to the V3R4 contract introduced by this SPEC.

**Acceptance** (Wave A complete when):
- AC-HRN-FND-002 passes (slash command verbs reachable without Go binary).
- Static check `grep -nE 'moai harness' .claude/commands/moai/harness.md .claude/skills/moai/workflows/harness.md` returns zero invocation matches.

**Risk**:
- Workflow body bloat — the four verbs all in one file could exceed the 500-line skill cap. Mitigation: progressive disclosure to `modules/harness-status.md`, `modules/harness-apply.md`, etc. if needed.

### 3.2 Wave B — Subagent Reorganization and Contract Preservation (P0)

**Goal**: Re-affirm that `moai-harness-learner` and `moai-meta-harness` subagents inherit the V3R4 contracts (no AskUserQuestion calls, structured blocker reports on missing input) verbatim from the orchestrator-subagent protocol.

**Owner**: `manager-develop`

**Inputs**:
- This SPEC's REQ-HRN-FND-015 (subagent AskUserQuestion prohibition)
- Existing `.claude/agents/moai/moai-harness-learner.md` (or its V3R3 home if not yet relocated)
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary

**Tasks** (T-B1 through T-B3):
- T-B1: Audit `moai-harness-learner` skill body for any `AskUserQuestion` invocation. If found (none expected), remove and replace with a blocker-report return template.
- T-B2: Add a top-of-body annotation in `moai-harness-learner` and `moai-meta-harness` SKILL.md citing this SPEC's REQ-HRN-FND-015 and the orchestrator-subagent boundary.
- T-B3: Verify the existing `.claude/agents/my-harness/*` agent definitions also respect the contract (they should already, since they were authored under the V3R3 PROJECT-HARNESS-001 SPEC).

**Acceptance** (Wave B complete when):
- AC-HRN-FND-003 passes (Tier-4 AskUserQuestion gate by orchestrator only).
- `grep -rn 'AskUserQuestion' .claude/agents/moai/moai-harness-learner.md .claude/agents/my-harness/*.md` returns zero invocation matches.

**Risk**:
- Cross-PR drift: another SPEC may modify these subagents concurrently. Mitigation: pin Wave B to immediately follow Wave A in the run-phase ordering.

### 3.3 Wave C — Hook Contract and Observer Baseline Preservation (P1)

**Goal**: Verify the PostToolUse hook continues to emit observations to `.moai/harness/usage-log.jsonl`, becomes a no-op when `learning.enabled: false`, and does not regress any V3R3 behavior.

**Owner**: `manager-develop` (delegating to `expert-devops` for hook inspection)

**Inputs**:
- This SPEC's REQ-HRN-FND-009 (observer no-op when disabled)
- This SPEC's REQ-HRN-FND-010 (PostToolUse baseline preserved)
- Existing PostToolUse hook implementation in `internal/hook/` and `.claude/hooks/moai/`

**Tasks** (T-C1 through T-C3):
- T-C1: Inspect `internal/hook/post_tool_use.go` (or equivalent) to confirm the observer path. Confirm `learning.enabled` is checked at the top.
- T-C2: Add or confirm a unit test verifying no-op behavior when `learning.enabled: false`.
- T-C3: Verify observation entry schema matches the JSONL format required by REQ-HRN-FND-010.

**Acceptance** (Wave C complete when):
- AC-HRN-FND-007 passes (observer no-op when disabled).
- AC-HRN-FND-008 passes (PostToolUse baseline and 4-tier ladder preserved).

**Risk**:
- Hook execution latency exceeding 100ms (existing budget) — out of scope for this SPEC. Documented in existing SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-001.

### 3.4 Wave D — CLI Deprecation Markers and CI Guard (P0)

**Goal**: Annotate `internal/cli/harness.go` with deprecation comments, write a Go test enforcing REQ-HRN-FND-002 (CLI re-registration prevention), and ensure `go test ./internal/cli/...` passes.

**Owner**: `manager-develop` (delegating to `expert-backend` for Go test authoring)

**Inputs**:
- This SPEC's REQ-HRN-FND-001 (CLI retirement)
- This SPEC's REQ-HRN-FND-002 (CI guard)
- Existing `internal/cli/harness.go` and `internal/cli/root.go`

**Tasks** (T-D1 through T-D4):
- T-D1: Add a deprecation header comment block at the top of `internal/cli/harness.go` referencing `SPEC-V3R4-HARNESS-001` and `BC-V3R4-HARNESS-001-CLI-RETIREMENT`.
- T-D2: Write a new Go test file `internal/cli/harness_retirement_test.go` that asserts `rootCmd.Commands()` does not contain any command with `Use: "harness"`. The test MUST reference this SPEC's REQ-HRN-FND-002 in its failure message.
- T-D3: Run the existing `internal/cli/harness_test.go` test suite to confirm no regressions in the CLI factory itself (the factory function remains defined for backward compatibility with internal callers that may inspect it).
- T-D4: Add a Markdown comment at the top of `internal/cli/root.go` near the command registration block noting the harness exclusion intent.

**Acceptance** (Wave D complete when):
- AC-HRN-FND-001 passes (CLI verb path retirement and CI guard).
- The new Go test passes locally and in CI.

**Risk**:
- The deprecation comment may be removed by a future formatting pass. Mitigation: the comment includes the SPEC ID and a deprecation tag string that grep-based audits can detect.

### 3.5 Wave E — Documentation and Superseded SPEC Follow-Up Coordination (P1)

**Goal**: Update CHANGELOG.md and release notes to announce the V3R4 foundation, prepare the follow-up coordination plan for transitioning the three V3R3 SPECs to `status: superseded`, and document the migration story for users.

**Owner**: `manager-docs` (CHANGELOG and release notes) and `manager-git` (V3R3 status-transition follow-up commit)

**Inputs**:
- This SPEC's REQ-HRN-FND-013 (V3R3 mutation prevention within this PR)
- The supersedes relationship declared in this SPEC's frontmatter

**Tasks** (T-E1 through T-E3):
- T-E1: Draft a CHANGELOG.md entry for v2.20.0 (or target release) describing BC-V3R4-HARNESS-001-CLI-RETIREMENT and citing this SPEC.
- T-E2: Draft a `.moai/release/RELEASE-NOTES-v2.20.0.md` section explaining the migration story (slash command continues to work; CLI verb path returns "unknown command"; no usage-log migration; downstream SPECs 002-008 enumerated).
- T-E3: Author a `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md` document specifying the V3R3 status-transition commit. This document is the spec-side input for the post-merge `manager-git` invocation.

**Acceptance** (Wave E complete when):
- CHANGELOG.md and release notes contain the V3R4 foundation announcement.
- `follow-up.md` exists with explicit V3R3 status-transition instructions.

**Risk**:
- Release notes may need to be re-targeted if the release version slips. Mitigation: use a placeholder `<TARGET-RELEASE>` token in `follow-up.md` and resolve at PR-merge time.

---

## 4. File-Level Changes (Run-Phase Reference)

This section summarizes the file-level changes that the run-phase Waves A-E will produce. The plan-phase PR (this manager-spec session) only produces the four SPEC artifacts under `.moai/specs/SPEC-V3R4-HARNESS-001/`.

| File | Operation | Wave | Notes |
|------|-----------|------|-------|
| `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` | Created | Plan | This SPEC's main document. |
| `.moai/specs/SPEC-V3R4-HARNESS-001/plan.md` | Created | Plan | This file. |
| `.moai/specs/SPEC-V3R4-HARNESS-001/acceptance.md` | Created | Plan | AC definitions. |
| `.moai/specs/SPEC-V3R4-HARNESS-001/tasks.md` | Created | Plan | Task breakdown. |
| `.claude/skills/moai/workflows/harness.md` | Modified (annotations only) | Wave A | Add V3R4 cross-references; verify no binary invocation. |
| `.claude/skills/moai-harness-learner/SKILL.md` | Modified (annotations only) | Wave B | Top-of-body annotation citing REQ-HRN-FND-015. Text-only; behavior unchanged. |
| `.claude/skills/moai-meta-harness/SKILL.md` | Modified (annotations only) | Wave B | Top-of-body annotation citing the V3R4 contract. Text-only. |
| `.claude/commands/moai/harness.md` | Verified unchanged | Wave A | Already CLI-free per PR #908. |
| `internal/hook/post_tool_use.go` (or equivalent) | Verified | Wave C | Confirm `learning.enabled` gate. |
| `internal/cli/harness.go` | Modified (deprecation comment) | Wave D | Top-of-file annotation; behavior unchanged. |
| `internal/cli/root.go` | Modified (comment near registration) | Wave D | Annotate harness exclusion. |
| `internal/cli/harness_retirement_test.go` | Created | Wave D | New CI guard test. |
| `CHANGELOG.md` | Modified | Wave E | Add v2.20.0 entry. |
| `.moai/release/RELEASE-NOTES-v2.20.0.md` | Created (or appended) | Wave E | Migration story. |
| `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md` | Created | Wave E | V3R3 status-transition instructions for `manager-git`. |

**Files explicitly NOT modified in this PR** (per REQ-HRN-FND-013):
- `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`
- `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`
- `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md`
- `.claude/rules/moai/design/constitution.md` (FROZEN)

---

## 5. Test Strategy

### 5.1 Static Verification

- `grep -nE 'moai harness' .claude/commands/moai/harness.md .claude/skills/moai/workflows/harness.md .claude/skills/moai/SKILL.md` returns zero invocation matches.
- `grep -rn 'AskUserQuestion' .claude/agents/moai/moai-harness-learner.md .claude/agents/my-harness/` returns zero invocation matches.
- `git diff main..HEAD --name-only | grep -E 'SPEC-V3R3-(HARNESS-001|HARNESS-LEARNING-001|PROJECT-HARNESS-001)/'` returns zero matches.
- `git diff main..HEAD -- .claude/rules/moai/design/constitution.md` returns zero non-comment diff.

### 5.2 Unit Tests

- `internal/cli/harness_retirement_test.go` — new test asserting `rootCmd.Commands()` does not include `Use: "harness"`.
- `internal/hook/post_tool_use_test.go` — verify no-op when `learning.enabled: false` (Wave C; existing test may already cover this).
- `internal/harness/*_test.go` — existing tests run unchanged; no regressions.

### 5.3 Integration Tests

- Manual verification in a Claude Code session:
  - `/moai:harness status` renders without binary invocation.
  - Simulated Tier-4 application triggers `AskUserQuestion` from the orchestrator.
  - Simulated frozen-guard violation appends to the audit log without raising user-visible errors.

### 5.4 Plan-Auditor Run

- `Agent(subagent_type: "plan-auditor")` is invoked at Phase 2.5 of this `/moai plan` session with the four SPEC artifacts as input.
- Pass threshold: 0.80 minimum. Below that, the auditor's findings are addressed in iterative drafts (max 3 iterations).
- If iteration 3 still fails, this manager-spec session escalates the findings to the orchestrator as a blocker report.

---

## 6. Risks (Top 5)

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|------------|
| Downstream SPECs 002-008 introduce mechanisms that drift from the FROZEN re-assertions | Medium | High | Each downstream SPEC's plan-auditor run cites REQ-HRN-FND-005 and REQ-HRN-FND-006 as binding constraints. The contract precedes the implementations. |
| The CI guard test (Wave D) becomes a maintenance burden if cobra's API changes | Low | Medium | The test asserts on `rootCmd.Commands()`, which is stable cobra public API. Mitigation: pin to cobra version when authored; review at major version bumps. |
| Users running `moai harness status` in shell scripts experience silent failures after upgrade | Medium | Low | Release notes (Wave E) document the migration. The CLI returns `unknown command` with exit code 1, which is a clear signal in script contexts. |
| The Tier-4 rate limit of 1/week (REQ-HRN-FND-012) causes user frustration if too conservative | Medium | Low | The expansion path (REQ-HRN-FND-018) is explicit. Power-users have the disable escape valve (REQ-HRN-FND-009). |
| Plan-auditor flags ambiguity in REQ-HRN-FND-017 (Reflexion-evaluator conflict resolution) because Reflexion is not yet implemented | Low | Medium | REQ-HRN-FND-017 is intentionally a contract assertion preceding the downstream implementation. The auditor's role is to verify the contract is unambiguous, not to verify the implementation. |

---

## 7. Dependencies

### 7.1 Inbound Dependencies (this SPEC depends on)

- `.claude/rules/moai/design/constitution.md` — design constitution FROZEN file. This SPEC re-asserts §2 and §5 verbatim.
- `.claude/rules/moai/core/agent-common-protocol.md` — orchestrator-subagent boundary. This SPEC inherits the AskUserQuestion contract.
- `.claude/rules/moai/core/askuser-protocol.md` — canonical AskUserQuestion protocol.
- `.claude/rules/moai/workflow/spec-workflow.md` — SPEC Phase Discipline.

### 7.2 Outbound Dependencies (SPECs that depend on this)

- `SPEC-V3R4-HARNESS-002` (multi-event observer) — blocked until this SPEC merges.
- `SPEC-V3R4-HARNESS-003` (embedding-cluster detection) — blocked.
- `SPEC-V3R4-HARNESS-004` (Reflexion self-critique) — blocked.
- `SPEC-V3R4-HARNESS-005` (principle-based scoring) — blocked.
- `SPEC-V3R4-HARNESS-006` (multi-objective scoring + auto-rollback) — blocked.
- `SPEC-V3R4-HARNESS-007` (skill library auto-organization) — blocked.
- `SPEC-V3R4-HARNESS-008` (cross-project lesson federation, deferred) — blocked.

### 7.3 Co-temporal Dependencies (none)

This SPEC has no co-temporal dependencies. All downstream SPECs enter plan-phase only after this SPEC merges.

---

## 8. Out of Scope (Explicit List)

This SPEC is the foundation. The following are out of scope and are deferred to the named downstream SPECs:

1. Multi-event observation expansion → `SPEC-V3R4-HARNESS-002`.
2. Embedding-cluster pattern detection → `SPEC-V3R4-HARNESS-003`.
3. Reflexion self-critique loop → `SPEC-V3R4-HARNESS-004`.
4. Principle-based self-scoring → `SPEC-V3R4-HARNESS-005`.
5. Multi-objective effectiveness measurement → `SPEC-V3R4-HARNESS-006`.
6. Voyager skill library → `SPEC-V3R4-HARNESS-007`.
7. Cross-project lesson federation → `SPEC-V3R4-HARNESS-008` (deferred; privacy-sensitive).
8. Physical deletion of `internal/cli/harness.go` and `internal/cli/harness_test.go` — deferred to a follow-up SPEC after downstream Waves merge.
9. Migration tooling for existing `.moai/harness/usage-log.jsonl` — explicitly out of scope (users begin v2 with whatever state they have).
10. GUI / dashboard for evolution history — out of scope.

---

## 9. Execution Order (Plan-Phase)

This `/moai plan SPEC-V3R4-HARNESS-001` session executes the following phases:

| Phase | Activity | Status |
|-------|----------|--------|
| Phase 0 | ultrathink deep research (read brain artifacts, V3R3 SPECs, design constitution) | Complete |
| Phase 1 | Branch setup: `plan/SPEC-V3R4-HARNESS-001` from `feat/cmd-harness-slash-wrapper` HEAD | Complete |
| Phase 2 | Draft `spec.md` with 18 EARS-format REQs | Complete |
| Phase 2.5 | Invoke `plan-auditor` subagent (max 3 iterations) | Pending |
| Phase 2.75 | Pre-commit quality gate (markdown lint, frontmatter validation) | Pending |
| Phase 2.8 | Draft `acceptance.md` with 12 ACs | Complete |
| Phase 2.9 | Draft `plan.md` (this file) | Complete |
| Phase 2.10 | Draft `tasks.md` | Pending |
| Phase 3 | Commit + delegate PR creation to `manager-git` via `Agent()` | Pending |

The run-phase Waves A-E happen in `/moai run SPEC-V3R4-HARNESS-001` after the plan PR merges.
