# Implementation Plan — SPEC-V3R5-LINT-CLEAN-001

## 1. Strategy Overview

This SPEC reduces the `moai agent lint --strict` baseline from 176 (post-Mega-Sprint W0, captured 2026-05-19 on `main` `02b2bb0a3`) to 0, through four sequential cleanup phases. Each LCLN-Phase targets a contiguous D-category subset, with mechanical/low-risk phases first to build CI signal confidence before tackling semantic-medium-risk phases.

**Terminology** (per `spec.md` §2.0 Glossary): "LCLN-Phase N" (Phase 1..4) = this SPEC's internal cleanup phases; "Mega-Sprint Wave W0..W4" = v3.5.0 release-roadmap SPECs. The plan-auditor iter-1 nomenclature "W1-LCLN..W4-LCLN" is retired in v0.2.0.

Per CLAUDE.local.md §6 and agent-common-protocol Time Estimation rule, no calendar estimates are used — priority labels and phase ordering only.

### High-level phase decomposition (recommended sequencing: **sequential**)

Canonical per-rule baseline (recomputed 2026-05-19, see `research.md` §1.2):
- LR-08=83, LR-06=29, LR-01=28, LR-03=25, LR-12=6, LR-02=3, LR-05=2 → **176 total**

Of the 176 findings, exactly **4** reside on `.claude/agents/moai/expert-mobile.md` (LR-02 ×1 at line 12, LR-03 ×1 at line 2, LR-06 ×1 at line 3, LR-08 ×1 at line 16 — empirically verified). LCLN-Phase 2 deletes this file, clearing those 4 in a single operation. LCLN-Phases 1 and 3 therefore exclude `expert-mobile.md` from their edit scope to avoid editing a file scheduled for deletion.

| LCLN-Phase | Target D-Categories | Target Rules | Findings Reduced | Sum-Check Derivation | Risk | Reversibility |
|------------|---------------------|--------------|-------------------|----------------------|------|---------------|
| Phase 1 | D1 (Frontmatter Hygiene) + D2 (Description Hygiene) + D4 (Worktree Discipline) | LR-03 + LR-12 + LR-06 + LR-05 (excluding `expert-mobile.md`) | **60** | (LR-03: 25−1=24) + (LR-12: 6) + (LR-06: 29−1=28) + (LR-05: 2) = 60 | Low | High |
| Phase 2 | D7 (Residual Live-Surface Drift) | All rules on `.claude/agents/moai/expert-mobile.md` (live file deletion) | **4** | LR-02 ×1 + LR-03 ×1 + LR-06 ×1 + LR-08 ×1 = 4 | Low | Medium (file deletion; restorable via git revert) |
| Phase 3 | D3 (Tool Boundary) | LR-01 + LR-02 (excluding `expert-mobile.md`, already deleted in Phase 2) | **30** | (LR-01: 28) + (LR-02: 3−1=2) = 30 | Medium | High |
| Phase 4 | D5 (Preload Drift, W4-resolvable subset) | LR-08 minus (the Phase-2 mobile contribution AND the W2-deferred residual) | **70** | LR-08 universe (83) − Phase-2 mobile contribution (1) − W2-deferred post-Phase-2 residual (12) = 70 | Medium | High |
| (Mega-Sprint W2 dissolves) | D5 (Preload Drift, W2-overlapping residual) | LR-08 referencing `moai-domain-{backend,frontend,database}` skills OR affecting `expert-{backend,frontend,mobile}` agents | **12** (post-Phase-2 residual) | Canonical W2-deferred set = 13; Phase 2 deletes mobile-LR-08, so 12 remain after this SPEC | — | Resolved automatically when Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001 lifecycle COMPLETEs |

**Sum-check** (this SPEC's contribution): 60 (Phase 1) + 4 (Phase 2) + 30 (Phase 3) + 70 (Phase 4) = **164 findings reduced**.
**Residual after Phase 4 merges**: 176 − 164 = **12 findings** (= the 12-element post-Phase-2 W2-deferred residual).
**Final 0/0 baseline**: achieved when Mega-Sprint W2 (CORE-SLIM-001) lifecycle COMPLETEs, dissolving the 12-finding residual.

**Canonical W2-deferred set size** (per plan-auditor iter-1 Finding 1 + empirical recount on 2026-05-19): **13 elements** (set total, computed as union of "drift skill ∈ {moai-domain-backend, moai-domain-frontend, moai-domain-database}" ∪ "affected agent ∈ {expert-backend, expert-frontend, expert-mobile}"). Of these 13, 1 (LR-08 on live `expert-mobile.md`) is cleared in-flight by Phase 2; the remaining **12** persist as the post-this-SPEC residual awaiting Mega-Sprint W2 dissolution.

### Decision: Sequential sequencing (recommended)

Alternatives considered:

- **Sequential (recommended)**: Phase 1 → Phase 2 → Phase 3 → Phase 4. Each phase's baseline JSON feeds the next phase's delta check. Lower coordination overhead, cleaner audit trail. Trade-off: slower wall-clock completion.
- **Parallel-where-safe**: Phase 1 ‖ Phase 2 ‖ Phase 4, then Phase 3 sequential. Theoretically safe because:
  - Phase 1 touches frontmatter `effort:` field + description field
  - Phase 2 deletes a single file (`expert-mobile.md`)
  - Phase 3 touches body text (literal AskUserQuestion)
  - Phase 4 touches frontmatter `skills:` array
  - However, Phase 2 and Phase 4 both touch `expert-mobile.md` indirectly (Phase 2 deletes it; Phase 4 would add `skills:` if not sequenced). Sequencing Phase 2 before Phase 4 prevents wasted work.

**Recommendation**: Sequential. Rationale:

1. The total finding reduction (164) is concentrated; parallelism saves limited wall-clock at the cost of merge-conflict risk.
2. Phase 2's `expert-mobile.md` deletion changes the file count, which would invalidate any in-flight parallel baseline JSON.
3. Each phase's delta verification benefits from a stable predecessor baseline.
4. CI signal per phase is preserved (easier root-cause for any regression).
5. 1-2 person team (CLAUDE.local.md §18 cadence) does not gain from parallel branches.

## 2. Task Decomposition

### LCLN-Phase 1 (D1 + D2 + D4: frontmatter & description hygiene)

| Task | Maps to AC | File(s) Touched | Risk | Reversibility |
|------|-----------|-----------------|------|---------------|
| T1.0 | AC-LCLN-001.1 | None — baseline capture (`./bin/moai agent lint --strict --format=json > .moai/state/lint-baseline-pre-LCLN-P1.json`) | None | N/A |
| T1.1 | AC-LCLN-P1-LR03 | All template agents missing `effort:` field, **excluding `.claude/agents/moai/expert-mobile.md`** (queued for Phase 2 deletion). 24 findings = 25 total − 1 mobile. Add `effort: <enum>` per SPEC-V3R2-ORC-003 canonical matrix. | Low | High |
| T1.2 | AC-LCLN-P1-LR12 | `evaluator-active.md`, `expert-refactoring.md`, `plan-auditor.md` (template + live, 6 findings). Correct `effort: high` → `effort: xhigh` per canonical matrix. | Low | High |
| T1.3 | AC-LCLN-P1-LR06 | All agents with `description:` containing `--deepthink flag:` boilerplate, **excluding `expert-mobile.md`**. 28 findings = 29 total − 1 mobile. Remove the boilerplate substring without altering the surrounding sentence's intent. | Low | High |
| T1.4 | AC-LCLN-P1-LR05 | `manager-develop.md` template + live (2 findings). Add `isolation: worktree` to YAML frontmatter. | Low | High |
| T1.5 | AC-LCLN-P1-build, AC-LCLN-007.1 | Run `make build` to regenerate `internal/template/embedded.go`. Verify template-first invariant: edits mirror across template + live, and `embedded.go` is present in PR diff. | Low | Medium (requires Go toolchain; embedded.go is a generated artifact) |
| T1.6 | AC-LCLN-002.1, 002.2, 002.3 | None — verification only. Capture `.moai/state/lint-baseline-post-LCLN-P1.json`, compute diff vs pre-LCLN-P1, assert NEW=0 and total decreased by exactly 60 (= 24 + 6 + 28 + 2). | None | N/A |
| T1.7 | AC-LCLN-003.1, 003.2, 003.3, 003.4 | Run orthogonal lints (`moai spec lint --strict`, `golangci-lint run ./...`, `moai workflow lint`); run Frozen Guard script (design.md §1.2). All MUST exit 0. | None | N/A |

### LCLN-Phase 2 (D7: residual live-surface drift)

| Task | Maps to AC | File(s) Touched | Risk | Reversibility |
|------|-----------|-----------------|------|---------------|
| T2.0 | AC-LCLN-001.1 | None — capture `.moai/state/lint-baseline-pre-LCLN-P2.json` | None | N/A |
| T2.1 | AC-LCLN-P2-mobile, AC-LCLN-007.1 | Delete `.claude/agents/moai/expert-mobile.md` (live file). Template counterpart was hard-deleted in Mega-Sprint W0 SPEC-V3R5-CLAUDE-REFRESH-001 REQ-CLR-004. **Post-condition (per plan-auditor iter-1 Finding 13)**: Run `./bin/moai update --dry-run` and confirm no agent registry changes beyond the `expert-mobile` deletion (i.e., no live regression and no spurious template-live drift). | Low | Medium (file deletion; recovery via git revert) |
| T2.2 | AC-LCLN-002.x | Capture post-LCLN-P2 baseline, assert NEW=0 and exactly 4 findings cleared (= LR-02 ×1 + LR-03 ×1 + LR-06 ×1 + LR-08 ×1 on expert-mobile.md). | None | N/A |
| T2.3 | AC-LCLN-003.x | Run orthogonal lints + Frozen Guard. | None | N/A |

### LCLN-Phase 3 (D3: tool-boundary hygiene)

| Task | Maps to AC | File(s) Touched | Risk | Reversibility |
|------|-----------|-----------------|------|---------------|
| T3.0 | AC-LCLN-001.1 | Capture `.moai/state/lint-baseline-pre-LCLN-P3.json` | None | N/A |
| T3.1 | AC-LCLN-P3-LR01 | All agents with literal `AskUserQuestion` token in body text (28 findings; 14 template + 14 live). For each occurrence, rewrite the surrounding paragraph to either:<br>(a) replace `AskUserQuestion` with phrase "orchestrator's user-interaction channel" + cross-reference `[askuser-protocol.md](.claude/rules/moai/core/askuser-protocol.md)`, OR<br>(b) if the line is genuinely about orchestrator escalation, restructure as a "blocker report" instruction directing the subagent to return a structured report.<br>Each edit is small (1-3 lines), no semantic loss. | Medium | High (text-only) |
| T3.2 | AC-LCLN-P3-LR02 | `builder-harness.md` template + live (2 remaining findings after Phase 2 cleared `expert-mobile.md`-LR-02). Remove `Agent` from `tools:` CSV list. Verify the agent doesn't actually invoke `Agent` in body; if it does, restructure to delegate via "return blocker report" pattern. | Medium | High |
| T3.3 | AC-LCLN-P3-build, AC-LCLN-007.1 | Run `make build` to regenerate embedded.go. Verify template-first invariant. | Low | Medium |
| T3.4 | AC-LCLN-002.x | Capture post-LCLN-P3 baseline, assert NEW=0 and exactly 30 findings cleared (= 28 LR-01 + 2 LR-02). | None | N/A |
| T3.5 | AC-LCLN-003.x | Run orthogonal lints + Frozen Guard. | None | N/A |

### LCLN-Phase 4 (D5: skill preload drift, W4-resolvable subset)

| Task | Maps to AC | File(s) Touched | Risk | Reversibility |
|------|-----------|-----------------|------|---------------|
| T4.0 | AC-LCLN-001.1 | Capture `.moai/state/lint-baseline-pre-LCLN-P4.json` | None | N/A |
| T4.1 | AC-LCLN-P4-drift-classification, AC-LCLN-005.2 | Re-classify the current LR-08 findings (after Phases 1, 2, 3 merge): split into W4-resolvable (skill ∉ {moai-domain-backend, moai-domain-frontend, moai-domain-database} AND agent ∉ {expert-backend, expert-frontend}) and W2-deferred (residual). Document the W2-deferred subset (expected 12 elements after Phase 2 clears 1 from the canonical 13-element set) in `.moai/state/lint-w2-deferred.json`. Bound check: count ∈ [11, 16] per AC-LCLN-005.2. | Low | N/A (classification only) |
| T4.2 | AC-LCLN-P4-add-skills | For each W4-resolvable LR-08 finding (70 total = 83 − 1 mobile − 12 W2-deferred-residual), add the missing skill to the agent's `skills:` YAML array. The linter's "drift" definition is "skill present in some agents of a category but not all"; resolution is to either (a) add the skill to the outlier, or (b) — if the skill is genuinely irrelevant to the agent — remove it from the over-including agents (only when justified). Default: prefer (a). Per finding, document choice in commit message. | Medium | High (frontmatter array edits) |
| T4.3 | AC-LCLN-P4-build, AC-LCLN-007.1 | Run `make build`. Verify template-first invariant. | Low | Medium |
| T4.4 | AC-LCLN-002.x | Capture post-LCLN-P4 baseline, assert NEW=0 and total ≈ 12 (the W2-deferred post-Phase-2 residual). | None | N/A |
| T4.5 | AC-LCLN-003.x | Run orthogonal lints + Frozen Guard. | None | N/A |
| T4.6 | AC-LCLN-005 | Verify `.moai/state/lint-w2-deferred.json` count equals the residual finding count (12 ± upstream drift). Document expectation that Mega-Sprint W2 SPEC will dissolve these. | None | N/A |

### Task ordering rationale

- **Phase 1 first** — broadest reduction (60 findings), lowest semantic risk. Establishes CI confidence early.
- **Phase 2 second** — file deletion is isolated; doing it before Phases 3-4 prevents wasted work on a soon-to-be-deleted file. Also clears the live/template asymmetry that LCLN-Phase 1 and 3 must work around.
- **Phase 3 third** — text-level edits to body content; depends on Phase 1's frontmatter cleanup being stable so that LR-01 fix doesn't accidentally re-introduce LR-06 boilerplate.
- **Phase 4 last** — most-entangled with Mega-Sprint W2; performing it last allows the Phase 4 task T4.1 (re-classification) to use the latest post-Phase-3 baseline, minimizing W2-deferred subset miscounts.

## 3. Milestones

Per CLAUDE.local.md §6, no time estimates. Priority labels and phase ordering only.

Sum-check column verifies that each milestone's post-merge total equals (176 − cumulative reduction). 176 − 60 = 116; 116 − 4 = 112; 112 − 30 = 82; 82 − 70 = 12.

| Milestone | Priority | Trigger Condition | Tasks Included | Post-merge `moai agent lint --strict` total (sum-check) |
|-----------|----------|-------------------|----------------|------|
| M1 — Frontmatter & Description Cleanup | P1 | LCLN-Phase 1 PR merged | T1.0–T1.7 | **116** (= 176 − 60) |
| M2 — Live-Surface Residual Cleared | P1 | LCLN-Phase 2 PR merged | T2.0–T2.3 | **112** (= 116 − 4) |
| M3 — Tool-Boundary Hygiene | P1 | LCLN-Phase 3 PR merged | T3.0–T3.5 | **82** (= 112 − 30) |
| M4 — Preload Drift Resolved (pre-W2) | P1 | LCLN-Phase 4 PR merged | T4.0–T4.6 | **12** (= 82 − 70; equals W2-deferred post-Phase-2 residual) |
| M5 — Mega-Sprint W2 COMPLETE → 0/0 baseline | P0 (gating release) | SPEC-V3R5-CORE-SLIM-001 lifecycle COMPLETE | (cross-SPEC observation, no tasks in this SPEC) | **0** (= 12 − 12) |

Sum-check arithmetic: 60 + 4 + 30 + 70 = 164 reductions within this SPEC's scope. Residual = 176 − 164 = 12 = the W2-deferred post-Phase-2 set. Mega-Sprint W2 dissolves the remaining 12, taking the total to 0. Verification: any consistent renumbering of the per-phase reduction MUST preserve the sum-check 60 + 4 + 30 + 70 = 164 and 176 − 164 − 12 = 0.

## 4. Technical Approach

### 4.1 LCLN-Phase 1 — Frontmatter & Description Cleanup

**T1.1 (LR-03 missing `effort:` field)**

Reference canonical matrix from `internal/cli/agent_lint.go` LR-12 documentation + SPEC-V3R2-ORC-003 effort matrix. Default mapping (verify by reading each agent's `description:` to confirm classification):

| Agent role pattern | Canonical `effort:` value |
|---|---|
| manager-* (orchestration depth) | `xhigh` |
| expert-* (domain depth) | `high` |
| builder-* (template depth) | `high` |
| evaluator-active | `xhigh` |
| plan-auditor | `xhigh` |
| claude-code-guide | `medium` |
| researcher | `high` |

For each agent file in the LR-03 finding list:

```
1. Read the file with Read tool.
2. Locate the YAML frontmatter block (between `---` delimiters at file start).
3. Use Edit to add `effort: <value>\n` after the existing `name:` or `model:` line.
4. Verify with grep that the field was added correctly.
```

Per CLAUDE.local.md §22.5, no template edits should change settings.local.json or runtime-managed keys; `effort:` is a static frontmatter field, safe.

**T1.2 (LR-12 effort drift)**

Three agents currently have `effort: high` but should be `effort: xhigh`:

- `evaluator-active.md` → `xhigh`
- `expert-refactoring.md` → `xhigh`
- `plan-auditor.md` → `xhigh`

Edit both template and live copy. Verification: `grep -E '^effort:' <file>` returns the new value.

**T1.3 (LR-06 `--deepthink flag:` boilerplate)**

Inspect a sample agent (`expert-backend.md`) to identify the exact boilerplate string:

```bash
grep -A 1 "description:" .claude/agents/moai/expert-backend.md | head -5
```

Expected pattern: `description: "..." --deepthink flag: ..."`. Use Edit to remove the literal substring `--deepthink flag: ` and any redundant trailing punctuation. Preserve the surrounding sentence's primary intent.

Do this surgically per file — no global find-replace, because the surrounding text varies. Per CLAUDE.local.md §1 [HARD] Approach-First: report the planned edit pattern to the user before mass-execution (handled in run-phase, not plan-phase).

**T1.4 (LR-05 missing `isolation: worktree`)**

`manager-develop.md` template + live. Add `isolation: worktree` to frontmatter under SPEC-V3R2-ORC-004 (already mandated in `.claude/rules/moai/workflow/worktree-integration.md` per CLAUDE.local.md §22.1 worktree advisory rules). Edit pattern:

```yaml
# before
permissionMode: <X>
# after
permissionMode: <X>
isolation: worktree
```

**T1.5 (build)** — `make build` regenerates `internal/template/embedded.go`. This is a generated artifact; commit it as part of the LCLN-Phase 1 PR but do not hand-edit.

### 4.2 LCLN-Phase 2 — Live-Surface Residual

**T2.1 (delete live `expert-mobile.md`)**

```bash
git rm .claude/agents/moai/expert-mobile.md
# Post-condition (per plan-auditor iter-1 Finding 13):
./bin/moai update --dry-run > /tmp/update-dry-run-P2.txt 2>&1
grep -E '^(\+|\-|MODIFY|DELETE)' /tmp/update-dry-run-P2.txt | grep -v 'expert-mobile.md' && \
  echo "BLOCK: moai update --dry-run shows non-mobile registry changes" && exit 1
echo "OK: no live registry regression beyond expert-mobile.md deletion"
```

No template edit needed (template was deleted in Mega-Sprint W0). This is purely a sync-divergence cleanup.

Verification: `[ ! -f .claude/agents/moai/expert-mobile.md ] && echo "PASS"`.

### 4.3 LCLN-Phase 3 — Tool-Boundary

**T3.1 (LR-01 literal `AskUserQuestion` in body)**

For each finding (file:line), Read the surrounding ±10 lines to understand context. Apply one of:

- **Pattern A (paraphrase + cross-ref)**: If the line documents the orchestrator's behavior, rewrite to:
  ```
  Before: "The orchestrator uses AskUserQuestion to collect ..."
  After:  "The orchestrator uses its user-interaction channel (see `.claude/rules/moai/core/askuser-protocol.md`) to collect ..."
  ```

- **Pattern B (subagent escalation)**: If the line incorrectly instructs the subagent to use `AskUserQuestion`, rewrite to:
  ```
  Before: "If input is missing, call AskUserQuestion to ask the user."
  After:  "If input is missing, return a structured blocker report to the orchestrator (per `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format)."
  ```

Per [HARD] rule from `.claude/rules/moai/core/askuser-protocol.md` § Subagent Prohibitions: subagents MUST NOT invoke AskUserQuestion. Pattern B is the strict semantically-correct rewrite; Pattern A is for documentation-only references.

**T3.2 (LR-02 `Agent` in `tools:` CSV)**

Three findings: `builder-harness.md` template + live, plus `expert-mobile.md` (already resolved in Phase 2). Phase 3 addresses the 2 builder-harness findings only.

For `builder-harness.md`: remove `Agent` from `tools:` CSV. If the agent body actually invokes `Agent`, restructure to delegate via blocker-report pattern (Pattern B above). Verify the agent's actual capabilities are not lost.

### 4.4 LCLN-Phase 4 — Preload Drift

**T4.1 (classification)**

Run baseline capture. For each LR-08 finding:

```
1. Extract the drift skill name from .message
2. Extract the affected agent's category from path or frontmatter
3. If skill ∈ {moai-domain-backend, moai-domain-frontend, moai-domain-database} → W2-deferred
4. If agent ∈ {expert-backend, expert-frontend, expert-mobile} → W2-deferred
5. Else → W4-resolvable
```

Output to `.moai/state/lint-w2-deferred.json` (committed to repo for audit trail).

**T4.2 (add skills)**

For each W4-resolvable finding:

```
1. Read agent frontmatter to locate `skills:` array
2. Use Edit to insert the missing skill name as a new YAML array element
3. Preserve YAML array formatting (- skill-name pattern)
```

Example for `manager-spec.md` missing `moai-foundation-core`:

```yaml
# before
skills:
  - moai-workflow-spec
  - moai-foundation-thinking
# after
skills:
  - moai-workflow-spec
  - moai-foundation-thinking
  - moai-foundation-core
```

Per CLAUDE.local.md token-budget concerns: each added skill is ~5K body tokens (Progressive Disclosure Level 2). Adding 14 instances of `moai-foundation-core` to all manager agents is the largest token-budget change in this SPEC. Mitigation: `moai-foundation-core` is already widely loaded; the drift is in metadata declaration, not actual runtime behavior. Token impact is metadata-only.

## 5. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| Edit accidentally modifies FROZEN-zone file | Low | High (constitutional violation) | Frozen Guard pre-flight script (design.md §1.2); each LCLN-Phase PR includes a `git diff --name-only ..main \| grep -E '<frozen-pattern>'` audit step (must return empty); guarded-paths set is the design.md §1.1 + Appendix A superset, not just zone-registry Frozen entries |
| W2-deferred classification miscount → introduces NEW finding in wrong phase | Medium | Medium (CI fails LCLN-Phase PR) | T4.1 explicit classification step bound-checks the result is in [11, 16] (= 13 ± 3 upstream drift tolerance); rollback LCLN-Phase PR if `NEW != 0`, re-classify |
| `make build` generates non-deterministic `embedded.go` | Low | Medium (PR diff polluted) | `make build` is deterministic in this codebase (verified per Mega-Sprint W0 PR #1006); commit the generated file alongside template changes |
| Skill addition (Phase 4) introduces token-budget regression in agent runtime | Low | Low | `moai-foundation-core` is already loaded by every workflow; metadata declaration matches runtime. Token impact measurable but not regression |
| Orthogonal lint regresses during a phase | Low | High (gates LCLN-Phase merge) | C5 mandates orthogonal lint check before LCLN-Phase PR opens; CI runs all four lints |
| LCLN-Phase PR rebase conflict with concurrent Mega-Sprint W1/W3/W4 SPECs | Medium | Low (rebase trivial since file scopes are orthogonal) | Per AC-LCLN-005.1: rebase onto latest `main` before each LCLN-Phase PR opens |
| Live `expert-mobile.md` deletion breaks user-project agent dispatch | Low | Low | Per CLAUDE.local.md §21 dev-only commands isolation: `expert-mobile` was Mega-Sprint W0-retired from template; live copy is the residual sync drift. User projects do not have `expert-mobile` after `moai init` post-W0. T2.1 post-condition (`moai update --dry-run`) provides safety net per plan-auditor iter-1 Finding 13. |
| `.golangci.yml` introduced concurrently | Low | Medium (re-baseline required) | Per C8: at each LCLN-Phase entry, detect file presence; if found, halt and re-plan |

## 6. Decision Log

| ID | Decision | Rationale | Alternatives Considered |
|----|----------|-----------|--------------------------|
| D1 | Sequential LCLN-Phase ordering (Phase 1 → 2 → 3 → 4) | Cleaner audit trail, no merge-conflict risk, simpler delta verification | Parallel-where-safe (Phase 1 ‖ 2 ‖ 4, then Phase 3) — rejected due to Phase 2/Phase 4 file overlap (`expert-mobile.md`) |
| D2 | Mega-Sprint W2-defer the 13-element LR-08 subset (12 after Phase-2 self-clear) | Resolving them here is wasted work — Mega-Sprint W2 SPEC retires the agents/skills entirely. Avoid duplicate edits. | Resolve all in this SPEC — rejected; violates DRY across SPECs |
| D3 | Use `effort: xhigh` for `evaluator-active`/`expert-refactoring`/`plan-auditor` (LR-12) | SPEC-V3R2-ORC-003 canonical matrix mandates `xhigh` per the agent's reasoning depth tier | Use `effort: high` and relax LR-12 — rejected; violates REQ-LCLN-012 |
| D4 | Phase 4 default = add missing skill (Pattern (a)), NOT remove from over-including agents | Skill preload drift is a metadata declaration issue; runtime behavior is already convergent. Adding to outliers preserves runtime semantics. Removing from over-includers would change runtime behavior unnecessarily. | Remove from over-includers when irrelevant — kept as escape hatch per finding |
| D5 | AC tree placement in `acceptance.md` (separate file), summary tree in `spec.md` §4 | Per user instruction: 5 artifacts including separate `acceptance.md`. Spec.md remains canonical AC declaration; acceptance.md provides shell-command-level verification | Co-locate full tree in spec.md — rejected per user instruction |
| D6 | Delta verification mechanism reuses Mega-Sprint W0 AC-CLR-008 precedent (`NEW=0` against pre-phase baseline) | Proven mechanism, identical jq diff command pattern, low cognitive overhead | Custom delta scheme — rejected for consistency |
| D7 | No `.golangci.yml` adoption in this SPEC | Out of scope; Go lint is already at 0 under default config | Adopt curated `.golangci.yml` — defer to separate future SPEC; C8 documents the conditionality |
| D8 | `make build` artifacts committed to LCLN-Phase PRs | Per CLAUDE.local.md §2: `internal/template/embedded.go` is generated but committed (consistent with Mega-Sprint W0 pattern) | Add `embedded.go` to `.gitignore` — rejected; would break user-project `moai init` consumption |
| D9 | Phase naming: "LCLN-Phase N" not "W<N>-LCLN" (plan-auditor iter-1 Finding 4) | Disambiguates from Mega-Sprint Wave WN nomenclature (W0..W4). Prevents reader confusion when this SPEC and Mega-Sprint W-series cross-reference each other. | Continue with "W<N>-LCLN" — rejected; auditor surfaced conflation as a D5 traceability defect |
| D10 | AC-LCLN-005.2 bound tightened from `[6, 18]` to `[11, 16]` (plan-auditor iter-1 Finding 7) | Empirical canonical = 13; ±3 covers concurrent Mega-Sprint W1/W2 perturbation of the skill/agent file set. Lower bound 6 (from skill-only narrow interpretation) is incompatible with the union semantics actually used. | Keep `[6, 18]` — rejected; too lax and inconsistent with `research.md` §2.2 derivation |

## 7. Verification Strategy (cross-reference to acceptance.md)

After each LCLN-Phase merges to `main`:

1. **Phase-specific AC**: Run phase-targeted lint count assertion (e.g., post-Phase-1 LR-03 + LR-12 + LR-06 + LR-05 contributions should be ≤ originating count minus 60 reduction target).
2. **Delta AC**: Run pre-phase baseline diff, assert NEW=0 (AC-LCLN-002.x).
3. **Orthogonal AC**: Run all four lints, assert three remain at 0 (moai spec lint, golangci-lint, moai workflow lint) and the fourth (moai agent lint) is on its planned trajectory (AC-LCLN-003.x).
4. **Frozen Guard AC**: Run Frozen Guard script (design.md §1.2), assert OK (AC-LCLN-003.4).
5. **Template-first AC**: Verify template + live edits mirror, embedded.go in PR diff (AC-LCLN-007.1).

See `acceptance.md` for full shell-command specifications per AC leaf.

## 8. Notes for Run-Phase

- Implementation is delegated to `manager-develop` via `/moai run SPEC-V3R5-LINT-CLEAN-001` per quality.yaml `development_mode` setting. The DDD/TDD cycle applies to template-side edits; for pure frontmatter additions (D1, D4, D5) the "improve" step is dominant. For body-text edits (D3) the "preserve behavior" step is paramount — no semantic loss.
- The orchestrator should open LCLN-Phases as separate PRs (one PR per LCLN-Phase) for cleaner CI signal and reviewer cognitive load.
- Each LCLN-Phase PR should include in its description: the phase number, the reduction target, the delta verification output, a Frozen Guard audit confirmation, and (for Phase 2) the `moai update --dry-run` confirmation.
- Per CLAUDE.local.md §18.3: chore/fix PRs use squash merge. Each LCLN-Phase PR squashes to one commit on `main`.
