# SPEC-HARNESS-TOKEN-OPT-001 — Implementation Plan

> Plan-phase artifact. Derived from `spec.md` §A-§H. Ordered by decision-reversibility (highest-change-likelihood first), per CLAUDE.md §7 Rule 1.

## §A. Context

A 5-lens parallel audit recovered ~18,400 tokens/turn of always-loaded rule weight plus 30-120s of redundant re-execution wall-clock per run-phase completion. The user approved "apply all" (P0+P1, 7 recommendations). This plan decomposes the work into 7 milestones (M0-M6) over 6 rule files + CLAUDE.local.md, with a cross-cutting template-mirror + neutrality-verification milestone (M6) that runs after the rule edits land.

Worktree: `.claude/worktrees/token-opt/` (branch `worktree-token-opt`, base `origin/main` HEAD `08eef9a0f`).

## §B. Known Issues

- **K1** — IK restatement count discrepancy: prompt cites "~88 prose restatements across 9+ files"; grep finds 53 raw string matches across 12 files. M3 must enumerate each occurrence and classify before cutting. The discrepancy is likely the prompt counting sub-clause restatements vs. raw string matches.
- **K2** — `goal-directive.md` lazy companion destination filename: prompt suggests `goal-directive-detail.md`. The session-handoff sidecar pattern uses `-examples.md`. For goal-directive, the lazy content is "detail / integration notes / native prohibition rationale", so `-detail.md` is the better suffix. Decision: `goal-directive-detail.md`.
- **K3** — `session-handoff.md` Diet destination: prompt allows existing `session-handoff-examples.md` OR a new `session-handoff-diet.md`. The examples sidecar is already lazy and already carries AP-D-style catalogue content (see its "Diet Constraints" reference). Decision: extend `session-handoff-examples.md` with the AP-D-001..005 + 9-item pre-emit checklist + V0 Abort Gate Doctrine, avoiding a third sidecar file.
- **K4** — A9 default inversion is a doctrinal edit, not a code edit. The "default" lives in the prose of `agent-common-protocol.md` §Parallel Execution. The invert is: change the prose so the consume-path is the default and the re-execute path is the explicit fallback. No Go code changes; no hook changes.

## §C. Pre-flight (verified before any edit)

- [x] Worktree HEAD at `origin/main` base (`08eef9a0f`, `0 0` divergence — verified 2026-08-11).
- [x] All 6 mirrored target files exist in both local and template trees (CLAUDE.local.md correctly has no template mirror).
- [x] SPEC ID regex PASS; no existing SPEC-HARNESS-TOKEN-OPT collision.
- [x] Baselines measured (see §D Constraints table in spec.md).

## §D. Constraints

- **Parity**: local and template byte-identical for every mirrored file.
- **Neutrality**: zero SPEC-ID / REQ-token / SHA / internal-date / audit-citation in template source.
- **do_not_touch**: see spec.md §B Out of Scope + REQ-HTO-010 grep sentinels.
- **No new code**: pure rule-content edits + one prose default inversion.
- **Frontmatter**: each `paths:`-scoped rule carries YAML frontmatter `--- paths: <globs> ---` (verified against existing examples: `cadence-bridge.md`, `native-invocation-model.md`, `session-handoff-examples.md`).

## §E. Self-Verification (plan-phase; the run-phase §E is owned by manager-develop)

- Plan-audit (independent `plan-auditor` agent) will run before Implementation Kickoff Approval.
- This plan's §F milestones cover all 10 REQs (HTO-001..010) with traceability.
- AC coverage: 18 acceptance criteria (see `acceptance.md`) cover all 10 REQs.

## §F. Milestones

> Ordered by decision-reversibility: highest-change-likelihood decisions (M3 SSOT consolidation across 12 files; M1 goal-directive split) first, mechanical edits (M0 paths batch; M4 A9 invert; M5 CLAUDE.local.md local) last, cross-cutting mirror (M6) at the end.

### M0 — paths: restrictions batch (rec #1, #4)

**REQs**: REQ-HTO-001 (verification-batch-pattern.md), REQ-HTO-002 (nav-tokens.md).

**Risk**: Low. `paths:` frontmatter is an existing pattern; the restriction is additive (file becomes lazy, never deleted).

**File-by-file changes**:

1. `.claude/rules/moai/workflow/verification-batch-pattern.md`
   - Add YAML frontmatter:
     ```
     ---
     description: "Read-only verification batching rationale + class taxonomy (loads at run/sync-phase completion)"
     paths: "**/.moai/specs/**,**/.claude/skills/moai/workflows/run.md,**/.claude/skills/moai/workflows/sync/**,**/verification-batch-pattern.md,**/agent-common-protocol.md"
     ---
     ```
   - Compress the A9 attributable diff-check section (currently ~80 lines) to a ~5-line thin pointer: "The A9 attributable diff-check pattern + fallback contract lives in `agent-common-protocol.md` §Parallel Execution → Attributable diff-check doctrinal switch (the SSOT). This file owns only the *why* (grouping rationale + class taxonomy)."
   - Preserve the WHY/Grouping/Class taxonomy body (that's this file's actual ownership).

2. `.claude/rules/moai/workflow/nav-tokens.md`
   - Add YAML frontmatter:
     ```
     ---
     description: "Author-facing NAV:DEC / NAV:SYM binding-token reference for the BAS integration layer"
     paths: "**/.moai/project/*.md,**/.moai/docs/**/*.md,**/*.go,**/nav-tokens.md"
     ---
     ```
   - Body unchanged otherwise.

**Template mirror map**: both files mirrored 1:1 to `internal/template/templates/.claude/rules/moai/workflow/`.

**Verification command**:
- `head -5 .claude/rules/moai/workflow/verification-batch-pattern.md` shows `paths:` line.
- `head -5 .claude/rules/moai/workflow/nav-tokens.md` shows `paths:` line.
- `diff .claude/rules/moai/workflow/verification-batch-pattern.md internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md` empty.
- `diff .claude/rules/moai/workflow/nav-tokens.md internal/template/templates/.claude/rules/moai/workflow/nav-tokens.md` empty.

### M1 — goal-directive.md split (rec #3)

**REQs**: REQ-HTO-003.

**Risk**: Medium. R1 (lazy companion reachability when goal is armed). Mitigation: `paths:` glob includes `.moai/state/goal/**` + goal workflow skill tree.

**File-by-file changes**:

1. `.claude/rules/moai/workflow/goal-directive.md` — REDUCE to always-loaded stub (~2K tokens):
   - Keep: `## What It Is` (condensed — 1 paragraph on `/moai goal` semantics + arm-only property).
   - Keep: `## Goal-Presentation Timing` (the arm-only invariant — load-bearing).
   - Keep: `## Hard Preconditions for Every Recommendation` (4 bullets, condensed).
   - Keep: `## Proactive Recommendation Triggers` — T1-T4 as one-liners (drop templates; templates live in detail companion).
   - Add: 1-line pointer at top — "Full detail (condition templates, Comparing Approaches table, Integration Notes, Native /goal Prohibition rationale) in `goal-directive-detail.md`."
   - Drop from stub (move to detail companion): `## Comparing Autonomous-Continuation Approaches` table, `## Writing an Effective Condition`, T1-T4 condition templates, `## MoAI Integration Notes`, `## Native /goal Prohibition`, `## Cross-References` verbose list (condensed to 3 essential refs in stub).

2. `.claude/rules/moai/workflow/goal-directive-detail.md` — NEW lazy companion:
   - Frontmatter:
     ```
     ---
     description: "Detail companion for goal-directive.md — condition templates, comparing-approaches table, integration notes, Native /goal Prohibition rationale"
     paths: "**/.moai/state/goal/**,**/.claude/skills/moai/workflows/goal.md,**/goal-directive.md,**/goal-directive-detail.md"
     ---
     ```
   - Body: all content moved from the stub (above) + a clear "this is the detail companion of goal-directive.md" header.

**Template mirror map**: `goal-directive.md` mirrored 1:1; `goal-directive-detail.md` created new in template source.

**Verification command**:
- `wc -c .claude/rules/moai/workflow/goal-directive.md` < 12000 (target ~8K-10K from baseline 25755).
- `wc -c .claude/rules/moai/workflow/goal-directive-detail.md` reports the relocated content size.
- `head -5 .claude/rules/moai/workflow/goal-directive-detail.md` shows `paths:` line.
- `grep -c "arm-only\|Goal-Presentation Timing" .claude/rules/moai/workflow/goal-directive.md` ≥ 1 (invariant preserved).

### M2 — session-handoff.md Diet lazy move (rec #6)

**REQs**: REQ-HTO-004.

**Risk**: Medium. R5 (CLAUDE.local.md §5/§7 references) does NOT apply here; risk is preserving the 5-section emission surface. Mitigation: explicit grep sentinels in REQ-HTO-010.

**File-by-file changes**:

1. `.claude/rules/moai/workflow/session-handoff.md`:
   - Keep always-loaded: 6-block Canonical Format + Cut-line Marker Specification + Localization Table (en/ko inline) + 5 Triggers + Emission-Time Save Obligation + Auto-Injected Resume Flow invariants + Auto-Memory Integration + Output Surface + Worktree-Anchored pointer.
   - Cut from always-loaded (move to sidecar): full §Diet Constraints (AP-D-001..005 verbose catalogue + 9-item pre-emit checklist), full §V0 Abort Gate Doctrine detail.
   - Replace with inline: 2 concrete AP-D examples + a 1-line pointer — "Full Diet Constraints catalogue (AP-D-001..005 + 9-item pre-emit checklist) and V0 Abort Gate Doctrine in `session-handoff-examples.md`."

2. `.claude/rules/moai/workflow/session-handoff-examples.md` (existing, already lazy via `paths: "**/session-handoff.md"`):
   - Append new sections: `## Diet Constraints (Full Catalogue)` (AP-D-001..005 verbose + 9-item pre-emit checklist), `## V0 Abort Gate Doctrine (Full Detail)`.
   - Update its `paths:` if needed (already matches `**/session-handoff.md`, which is the right trigger).

**Template mirror map**: both files mirrored 1:1.

**Verification command**:
- `wc -c .claude/rules/moai/workflow/session-handoff.md` < 25000 (target ~22K-24K from baseline 26267; modest reduction — most weight is already in the always-loaded 6-block contract).
- `grep -c "✂──── 여기부터 복사" .claude/rules/moai/workflow/session-handoff.md` = 1 (Cut-line Marker Spec preserved).
- `grep -c "AP-D-001\|AP-D-002" .claude/rules/moai/workflow/session-handoff-examples.md` ≥ 2 (catalogue relocated).

### M3 — Implementation Kickoff Approval SSOT consolidation (rec #5)

**REQs**: REQ-HTO-005.

**Risk**: HIGHEST. R2 (over-cutting a load-bearing local restatement). Mitigation: enumerate-and-classify BEFORE any cut; plan-auditor reviews classification.

**Highest-touch milestone: 12 files contain 45 measured occurrences of "Implementation Kickoff Approval" (baseline `grep -rn "Implementation Kickoff Approval" .claude/rules/moai/ CLAUDE.md | wc -l` = 45, run 2026-08-11).**

**File-by-file procedure** (one pass per file):

1. `grep -n "Implementation Kickoff Approval" <file>` to enumerate every occurrence.
2. Classify each occurrence as:
   - **A — Canonical §E**: the SSOT statement in `orchestration-mode-selection.md` (line 16 + §E Anti-Patterns line 188). PRESERVE.
   - **B — Load-bearing local restatement**: the file's own logic depends on the gate (e.g., `agent-common-protocol.md` §Super-Advisor Escalation references the gate in its E1-E4 triggers; `goal-directive.md` Goal-Presentation Timing references the gate's ordering). PRESERVE up to 1 canonical-quality restatement per file.
   - **C — Redundant restatement**: prose that restates the mandate for emphasis without changing the file's logic. CUT → replace with `Per the Implementation Kickoff Approval mandatory-restoration invariant (orchestration-mode-selection.md §E).`

**Files in scope** (12, from grep):
- `.claude/rules/moai/core/askuser-protocol.md`
- `.claude/rules/moai/development/coding-standards.md`
- `.claude/rules/moai/workflow/cache-aware-execution.md`
- `.claude/rules/moai/workflow/archived-agent-rejection.md`
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (the SSOT — A-class only)
- `.claude/rules/moai/workflow/dynamic-workflows.md`
- `.claude/rules/moai/workflow/cadence-bridge.md`
- `.claude/rules/moai/workflow/session-handoff-examples.md`
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/rules/moai/workflow/goal-directive.md` (will already be the M1 stub — re-grep after M1 lands)
- `.claude/rules/moai/workflow/spec-workflow.md`
- `CLAUDE.md`

**Pre-cut deliverable**: a classification table (file:line → A/B/C + reason) is appended to this section below. **Decided policy (user-confirmed 2026-08-11 via AskUserQuestion round — "보존 우선 / default-to-preserve")**: when an occurrence is ambiguous between B (load-bearing local) and C (redundant duplicate), classify as B and preserve. Over-cutting is the named hazard R2; the user accepted lower token savings in exchange for zero risk of cutting a load-bearing restatement.

**Classification table** (45 occurrences across 12 files; baseline `grep -rn "Implementation Kickoff Approval" .claude/rules/moai/ CLAUDE.md | wc -l` = 45, run 2026-08-11):

| File | Line | Class | Reason |
|---|---|---|---|
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 14 | **A** | Canonical SSOT header — co-locates progression-mode axis with mandatory-restoration invariant. PRESERVE. |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 16 | **A** | **The canonical statement** — "mandatory and score-independent... per the Implementation Kickoff Approval mandatory-restoration policy." PRESERVE. |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 188 | **A** | §E Anti-Patterns canonical restatement — "violates the Implementation Kickoff Approval mandatory-restoration policy." PRESERVE. |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | 35, 64, 88, 127, 128, 131, 161, 190, 193, 203, 205 (11 occ) | **B** | Within-SSOT-file operational references — the file's own Mode 4/5/6 logic depends on each. PRESERVE (intra-SSOT references, not duplicates). |
| `.claude/rules/moai/core/askuser-protocol.md` | 180 | **B** | Report-Before-Ask Gate exceptions list — cites IK Approval as the canonical example of an exception. Removing would leave the exception list incomplete. PRESERVE. |
| `.claude/rules/moai/development/coding-standards.md` | 129 | **B** | §4 hook-signal doctrine cross-references IK Approval to clarify that hook signals do NOT enforce confirmation. Load-bearing for the file's own scope discipline. PRESERVE. |
| `.claude/rules/moai/workflow/archived-agent-rejection.md` | 149 | **B** | Anti-pattern catalogue entry — the hazard description depends on IK Approval for its "user loses the gate signal" framing. Load-bearing. PRESERVE. |
| `.claude/rules/moai/workflow/cache-aware-execution.md` | 21 | **B** | Non-goals section — explicitly carves the cache-aware rules away from gate semantics. Load-bearing for the file's scope contract. PRESERVE. |
| `.claude/rules/moai/workflow/cadence-bridge.md` | 26 | **B** | The file's CENTRAL doctrine — cadence cannot bypass the gate. Load-bearing. PRESERVE. |
| `.claude/rules/moai/workflow/cadence-bridge.md` | 107 | **C** | Cross-references list bullet — already a pointer; restatable as the 1-line canonical ref. CUT → `Per the Implementation Kickoff Approval mandatory-restoration invariant (orchestration-mode-selection.md §E).` (Borderline; default-to-preserve would mark B, but this bullet carries no local semantic beyond the pointer — defensible C.) |
| `.claude/rules/moai/workflow/dynamic-workflows.md` | 58, 98, 107 (3 occ) | **B** | Each disambiguates the batch/workflow form from the plan-to-implement gate — load-bearing for the file's own scope contract (workflow ≠ approval). PRESERVE. |
| `.claude/rules/moai/workflow/goal-directive.md` | 63, 65, 67, 71, 80, 96, 97, 98, 100 (9 occ) | **B** | Each is load-bearing for the file's Goal-Presentation Timing arm-only invariant, the Hard Preconditions list, the T1-T4 triggers, or the MoAI Integration Notes. M1 will relocate most to the lazy companion `goal-directive-detail.md` regardless; M3 does not cut them. PRESERVE. |
| `.claude/rules/moai/workflow/session-handoff.md` | 88, 92, 109 (3 occ) | **B** | Block 1 SEED-not-permission-grant clause, Block 5 arm-only clause, Auto-Injected Resume Flow invariants — each load-bearing for the paste-ready resume contract. PRESERVE. |
| `.claude/rules/moai/workflow/session-handoff-examples.md` | 49, 92, 291, 296, 304 (5 occ) | **B** | Each is load-bearing for the Block 5 contract, the condition invariants list, or the SEED-not-permission-grant clauses. PRESERVE. |
| `.claude/rules/moai/workflow/spec-workflow.md` | 73, 75, 351 (3 occ) | **B** | Design-phase route contract + Phase 0.5 skip-vs-IK-Approval distinction — each load-bearing for its own passage. PRESERVE. |
| `.claude/rules/moai/workflow/spec-workflow.md` | 355 | **B** | Closing reference in the skip-eligibility passage — ambiguous between B and C; default-to-preserve → B. PRESERVE. |
| `CLAUDE.md` | 40 | **B** | The orchestrator's own pipeline-doc — the Approval Gates bullet cites IK Approval as the plan→run boundary gate. Load-bearing for CLAUDE.md §2. PRESERVE. |

**Classification summary**: A=3 (the canonical §E SSOT in orchestration-mode-selection.md lines 14/16/188), B=41 (load-bearing local restatements, PRESERVE), C=1 (cadence-bridge.md:107, the single clear-cut). Under the user-confirmed default-to-preserve policy, M3 cuts exactly 1 occurrence and designates the §E SSOT; the mandate is preserved everywhere.

**Token-savings revision (honest)**: the audit's original ~2,000 tok/turn projection for M3 assumed a larger C-class set. Under the measured baseline (45/12) and the user's default-to-preserve policy, M3's realized savings are ~50-100 tok/turn (one cut + the SSOT designation). The larger token wins in this SPEC come from M0/M1/M2 paths-restrictions, NOT from M3. This is recorded honestly so sync-audit does not over-count M3's contribution.

**Template mirror map**: every mirrored file above is mirrored 1:1. CLAUDE.md (a project instruction file, distributed via template as `internal/template/templates/CLAUDE.md`) is mirrored 1:1. (CLAUDE.local.md is NOT in this list — it is REQ-HTO-007's local-only concern.)

**Verification command**:
- After M3: `grep -rn "Implementation Kickoff Approval" .claude/rules/moai/ CLAUDE.md | wc -l` ≤ 45 (no-regression floor; under the user-confirmed default-to-preserve policy, most occurrences are class B and remain — the reduction target is modest, not 25). The `< 25` target is the *ambitious* ceiling, achievable only if the C-class set grows during run-phase re-classification.
- `grep -c "Implementation Kickoff Approval mandatory-restoration invariant (orchestration-mode-selection.md §E)" .claude/rules/moai/ CLAUDE.md` ≥ 1 (the cross-reference pointer landed in any C-class cut site; under default-to-preserve this may be as low as 1).
- `grep -c "mandatory and score-independent\|mandatory, score-independent" .claude/rules/moai/workflow/orchestration-mode-selection.md` ≥ 1 (the mandate itself preserved).

### M4 — A9 attributable diff-check default inversion (rec #7)

**REQs**: REQ-HTO-006.

**Risk**: Low with R3 mitigation. The fallback contract is preserved unchanged; only the DEFAULT inverts.

**File-by-file changes**:

1. `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution → Attributable diff-check doctrinal switch:
   - Find the current prose: "The canonical 7-command batch RE-EXECUTES test / lint / vet / cover by default. SPEC-SYNC-PARALLEL-DOCS-001 A9 introduces a doctrinal switch..."
   - INVERT: "The canonical 7-command batch CONSUMES the attributable §E evidence by default (no re-execution) when the three-way attribution match holds. SPEC-SYNC-PARALLEL-DOCS-001 A9 introduces this default-inversion switch... On any mismatch, the batch falls back to re-execution of the affected verification dimension (fallback contract preserved unchanged)."
   - Preserve verbatim: the three-way match predicate (snapshot key / command / output), the fallback-to-re-execution contract, the mismatch reason enum (`snapshot_key_drift` / `command_drift` / `missing_section_e` / `output_drift`), and the VCI §1.1 invariant statement.
   - DO NOT touch the `Attributable diff-check pattern` section in `verification-batch-pattern.md` beyond the M0 thin-pointer (M0 already compressed it).

**Template mirror map**: `agent-common-protocol.md` mirrored 1:1.

**Verification command**:
- `grep -c "CONSUMES the attributable §E evidence" .claude/rules/moai/core/agent-common-protocol.md` ≥ 1.
- `grep -c "fallback to re-execution\|fallback-to-re-execution" .claude/rules/moai/core/agent-common-protocol.md` ≥ 3 (fallback contract preserved).
- `grep -c "VCI §1.1 invariant holds on every path" .claude/rules/moai/core/agent-common-protocol.md` ≥ 1.

### M5 — CLAUDE.local.md consolidation (rec #2)

**REQs**: REQ-HTO-007.

**Risk**: Low. Local-only file; no template mirror; no distributed-user impact.

**File-by-file changes**:

1. `CLAUDE.local.md`:
   - Consolidate §18-27 stub tails into a single `## References` section (~10 lines, one bullet per external `.moai/docs/X.md` pointer).
   - Move §5 (Version Management, ~92 lines) to `.moai/docs/version-management.md` (new file) — leave a 1-line pointer in CLAUDE.local.md `## References`.
   - Move §7 (Hook Development, ~57 lines) to `.moai/docs/hook-development.md` (new file) — leave a 1-line pointer.
   - Update internal cross-references in CLAUDE.local.md that currently point to §5/§7.

**Template mirror map**: NONE. `CLAUDE.local.md` is explicitly excluded from templates per CLAUDE.local.md §2. The new `.moai/docs/version-management.md` and `.moai/docs/hook-development.md` files are also local-only (`.moai/docs/` is not in the template distribution set per CLAUDE.local.md §2 Protected Directories).

**Verification command**:
- `wc -c CLAUDE.local.md` < 32000 (target ~28K-32K from baseline 42164).
- `grep -c "^## References$" CLAUDE.local.md` = 1.
- `grep -c "^## §5\.\|^## 5\." CLAUDE.local.md` = 0 (§5 moved out).
- `ls -la .moai/docs/version-management.md .moai/docs/hook-development.md` both exist.

### M6 — Template mirror + make build + §25 neutrality verification (cross-cutting)

**REQs**: REQ-HTO-008, REQ-HTO-009, REQ-HTO-010.

**Risk**: Low. Mechanical mirror + build + CI guard.

**Pre-existing byte drift closed by this milestone (not attributable to this SPEC's edits)**: M6's verbatim mirror also reconciles 3 files whose local and template byte-counts have drifted on `origin/main` independent of this SPEC:
- `verification-batch-pattern.md`: local 8405 / template 8470 (+65 bytes template-heavy).
- `orchestration-mode-selection.md`: local 29399 / template 28324 (−1075 bytes template-light).
- `agent-common-protocol.md`: local 38182 / template 38196 (+14 bytes template-heavy).

After M0-M4 land and M6 mirrors verbatim, all three drift deltas collapse to zero (byte-identical parity). The sync-audit must NOT misattribute the pre-existing drift closure to this SPEC's token-reduction work — only the M0-M4 edits are this SPEC's contribution; the drift reconciliation is a side-effect of the mirror step. Recorded here so the delta attribution is auditable.

**Procedure** (runs after M0-M4 land; M5 is local-only and excluded):

1. For each mirrored target file (6 files: verification-batch-pattern.md, nav-tokens.md, goal-directive.md, goal-directive-detail.md [NEW], session-handoff.md, session-handoff-examples.md, agent-common-protocol.md, + the 9 mirrored files in M3):
   - `cp .claude/rules/moai/<file> internal/template/templates/.claude/rules/moai/<file>` (verbatim mirror).
2. Run `make build` to regenerate `internal/template/catalog.yaml` (templates embedded via `//go:embed all:templates`).
3. Commit `internal/template/catalog.yaml` alongside the template source.
4. Run the 5-item §25.3 pre-commit self-check (`.moai/docs/template-internal-isolation-doctrine.md`):
   - [ ] No SPEC-ID (`SPEC-HARNESS-TOKEN-OPT-001`) in any template source.
   - [ ] No REQ-token (`REQ-HTO-`) in any template source.
   - [ ] No audit citation in any template source.
   - [ ] No internal work date (2026-08-11) beyond routine doc dates.
   - [ ] No commit SHA in any template source.
5. Run CI guard locally: `go test ./internal/template/...` (covers `internal_content_leak_test.go` + `split_namespace_test.go`).
6. Verify `.github/workflows/template-neutrality-check.yaml` would pass (the workflow runs on push; local equivalent is the test above).

**do_not_touch grep sentinels** (REQ-HTO-010):
- `grep -c "no unobserved-claim\|no-unobserved-claim" .claude/rules/moai/core/verification-claim-integrity.md` ≥ 1.
- `grep -c "AskUserQuestion is the only user-facing question channel\|AskUserQuestion is the.*exclusive.*channel" .claude/rules/moai/core/askuser-protocol.md` ≥ 1.
- `grep -c "ToolSearch(query: \"select:AskUserQuestion\")" .claude/rules/moai/core/askuser-protocol.md` ≥ 1.
- `grep -c "mandatory and score-independent\|mandatory, score-independent" .claude/rules/moai/workflow/orchestration-mode-selection.md` ≥ 1.
- `grep -c "any-mismatch → re-execute" .claude/rules/moai/core/agent-common-protocol.md` ≥ 1.
- `grep -c "Functionality 40\|Functionality.*40" .claude/rules/moai/ -r` ≥ 1 (sync-auditor weights — locate the file at run-phase; do_not_touch is preservation, file is not edited by this SPEC).
- `grep -c "verbatim command + observed output + baseline-attribution\|verbatim command.*observed output" .claude/rules/moai/development/manager-develop-prompt-template.md` ≥ 1.
- `grep -c "✂──── 여기부터 복사" .claude/rules/moai/workflow/session-handoff.md` = 1.

## §G. Anti-Patterns (named for traceability)

- **AP-HTO-001** — Cutting a load-bearing IK restatement (R2). Mitigation: classify-before-cut; default-to-preserve when ambiguous; plan-auditor reviews the classification.
- **AP-HTO-002** — Mirror drift between local and template. Mitigation: `diff` after each mirror; M6 final parity check.
- **AP-HTO-003** — SPEC-ID / REQ-token leakage into template source. Mitigation: §25.3 5-item checklist + CI guard.
- **AP-HTO-004** — Weakening the A9 fallback contract. Mitigation: REQ-HTO-006 preserves the fallback contract verbatim; grep sentinel `any-mismatch → re-execute`.
- **AP-HTO-005** — Lazy companion unreachable (R1). Mitigation: `paths:` glob includes goal-state files AND goal workflow skill tree.
- **AP-HTO-006** — Editing `verification-claim-integrity.md` §1-§3. HARD prohibition — out of scope.

## §H. Cross-References

- spec.md §B (Out of Scope — the do_not_touch catalogue).
- spec.md §C REQ-HTO-001..010 (the requirement layer this plan derives from).
- `acceptance.md` AC-HTO-001..018 (the verification layer).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields (frontmatter validation).
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 + §25.3 (neutrality catalogue + 5-item pre-commit self-check).

---

**Decided policy (2026-08-11, user-confirmed via AskUserQuestion round — "보존 우선 / default-to-preserve")**: the M3 IK restatement classification is resolved. The full A/B/C table for the 45 measured occurrences across 12 files is in §F.M3 above. Summary: A=3 (canonical §E SSOT), B=41 (preserve), C=1 (cut). M3 cuts exactly 1 occurrence and designates the SSOT; the mandate is preserved everywhere. No further clarification needed before Implementation Kickoff Approval.
