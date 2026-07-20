---
id: SPEC-MOAI-SKILL-DOCTRINE-FIX-001
title: "moai Skill Folder Doctrine Drift Remediation — Implementation Plan"
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: ".claude/skills/moai"
lifecycle: spec-anchored
tags: "skill-doctrine, drift-remediation, gears, template-neutrality, harness, tier-l, agent-catalog"
---

# Implementation Plan: SPEC-MOAI-SKILL-DOCTRINE-FIX-001

## §A. Context

This is a **documentation-correctness** SPEC: the fixes are prose/markdown edits inside `.claude/skills/moai/` (and its `internal/template/templates/` mirror), not Go code or config changes. Tier is **L** (42-file scope, 53 REQs, cross-cutting template-neutrality + CI-regex hardening, 9 parallel write-groups) — Route B (PR route) applies per `spec-workflow.md` Frozen HARD Route A/B (Tier L → manager-git PR route, per REQ-SKF-001/002's own subject matter, applied reflexively to this SPEC's own execution).

**Complexity Tier artifact count**: Tier L → 5 artifacts (spec.md, plan.md, acceptance.md, progress.md, + this SPEC's own scope is broad enough that a `design.md`/`research.md` companion is NOT required — the findings are already fully specified by the audit; no additional design exploration is needed).

## §B. Known Issues / Risks

1. **Template-mirror drift on 3 files (pre-existing, unrelated to this SPEC).** `workflows/run/task-decomposition.md`, `workflows/plan/clarity-interview.md`, and `workflows/review.md` currently DIFFER from their `internal/template/templates/` mirror (confirmed via `diff -q` at plan-authoring time, 2026-07-08). Before applying this SPEC's fixes to these 3 files, run-phase MUST diff local-vs-template to determine whether the local drift already contains unrelated in-flight work that must be preserved, or whether it is stale and should be reconciled from template. **Do not blindly overwrite either side.**
2. **Cross-write-group REQ split (WG-A / WG-B / WG-E), resolved.** REQ-SKF-050 spans 3 write-groups by design (tag `[WG-E,WG-A,WG-B]`, mirroring the pattern already used by REQ-SKF-007 `[WG-H,WG-A]` and REQ-SKF-004 `[WG-C,WG-B]`): WG-E normalizes `team/debug.md`'s spawn instructions, WG-A adds the team-mode pointer to `SKILL.md`'s fix entry, WG-B corrects the `workflows/fix.md:43` naming inconsistency. All 3 sub-clauses are real assigned edits in their owning write-group's M3 milestone row (§F) — none is a "cross-ref only" placeholder. Recommended sequencing (soft, not blocking): land WG-E's `team/debug.md` naming fix before WG-A's `SKILL.md` pointer text is finalized, since the pointer text should name the corrected path; this is a text-content nicety, not a file-write conflict (the 3 write-groups touch 3 disjoint files and can still fan out in the same parallel batch).
3. **CI leak-test regex changes are highest-blast-radius (WG-I).** Changing `internal_content_leak_test.go` regex families risks either (a) false negatives if a new pattern is too narrow, or (b) false positives that fail the CI guard against files this SPEC does NOT intend to touch (i.e., legitimate historical SPEC-ID citations elsewhere in the template tree that are NOT skill-body leaks). WG-I implementation MUST run the full `internal_content_leak_test.go` suite against the CURRENT template tree (before this SPEC's other fixes land) to confirm no unintended regressions, then again AFTER WG-A..WG-H land to confirm the newly-fixed leaks (REQ-SKF-027, REQ-SKF-032, REQ-SKF-035) are now caught retroactively (i.e., the extended regex should have flagged them before the fix).
4. **`harness.md` "no CLI" framing rewrite (REQ-SKF-007) is the highest-severity single edit.** It touches the file's opening paragraphs (lines 3-7, 29-30, 34, 263) which likely anchor several internal cross-references within the same file and in `SKILL.md`. Recommend reading the full file before editing to avoid leaving a stale internal ToC/pointer.

## §C. Pre-flight (MANDATORY before any Write/Edit in run-phase)

For **every** file with a template mirror (see §F write-group tables — mirror status column), run-phase MUST:

1. `diff -q .claude/skills/moai/<path> internal/template/templates/.claude/skills/moai/<path>` — confirm current mirror status (IDENTICAL / DIFFERS).
2. If DIFFERS (only expected for the 3 files in §B.1): read both versions, determine which is authoritative, reconcile BEFORE applying this SPEC's fix.
3. Edit `internal/template/templates/.claude/skills/moai/<path>` FIRST (Template-First Rule, `CLAUDE.local.md` §2).
4. Run `make build`.
5. Sync the rendered result to `.claude/skills/moai/<path>` (copy or `moai update -t`).
6. Re-run `diff -q` to confirm byte-identity is re-established (or document the intentional residual local-only content, if any — none is expected for this SPEC's scope).

For files WITHOUT a template mirror (none identified in this SPEC's 42-file scope — all 37 skill-body files audited have mirrors; only the 5 no-finding files are pass-through and untouched) this step is skipped.

## §D. Constraints

- [HARD] Template-First discipline (§C above) applies to every WG-A..WG-H file edit.
- [HARD] Template Internal-Content Isolation (`CLAUDE.local.md` §25) applies to every prose edit touching a template-mirrored file — no internal SPEC ID / REQ / AC / C-PH / commit-SHA / internal-date leaks introduced or left in place.
- [HARD] No code/config behavior changes (see spec.md §E Exclusions) — this SPEC is documentation-only except for REQ-SKF-053 (test-regex extension, itself a test-file change, not production code).
- [HARD] 16-language neutrality (`CLAUDE.local.md` §15) applies to REQ-SKF-017 and REQ-SKF-049 (language-glob expansion) — all 16 languages listed equally, no "PRIMARY" language bias.
- [HARD] Every `AskUserQuestion` example/reference touched (REQ-SKF-018, REQ-SKF-040, REQ-SKF-041, REQ-SKF-042) must conform to `askuser-protocol.md` (Recommended-label-first, ≤4 options, ToolSearch preload citation).

## §E. Self-Verification (progress.md §E skeleton — placeholder headings only)

This plan-phase authoring populates ONLY `progress.md` §E.1. §E.2/§E.3 (manager-develop, run-phase) and §E.4 (manager-docs, sync-phase) are left as empty placeholder headings per the canonical progress.md §E Skeleton Generation contract — see `progress.md` in this SPEC directory.

## §F. Milestones (priority-ordered, no time estimates)

Milestone ordering is severity-first (M1 CRITICAL → M2 MAJOR → M3 MINOR → M4 CI hardening). Within each milestone, the 9 write-groups (WG-A..WG-I) are **file-disjoint** and safe to fan out in parallel (single-turn multi-`Agent()` spawn, per `agent-common-protocol.md` § Parallel Execution) — no two write-groups touch the same file, so no write-write conflicts across groups. Each write-group should apply ALL of its assigned REQs (across M1/M2/M3) to its files in ONE pass per file, to minimize redundant `make build` / diff-verify cycles per §C.

### M1 — CRITICAL (P0, 7 findings): apply first, blocking

| Write-group | Files | REQs in this milestone |
|---|---|---|
| WG-D | `workflows/run/task-decomposition.md` | REQ-SKF-001 |
| WG-F | `workflows/sync/delivery.md` | REQ-SKF-002 |
| WG-C | `workflows/plan/spec-assembly.md` | REQ-SKF-003 |
| WG-C, WG-B | `workflows/plan/clarity-interview.md`; `workflows/review.md` | REQ-SKF-004 |
| WG-G | `workflows/project/meta-harness.md` | REQ-SKF-005 |
| WG-G | `workflows/project.md`; `workflows/project/doc-generation.md` | REQ-SKF-006 |
| WG-H, WG-A | `workflows/harness.md`; `SKILL.md`; `internal/cli/root.go:157-160` (comment-only, WG-H — see §D exception note) | REQ-SKF-007 |

### M2 — MAJOR (P1, consolidated 28 REQs)

| Write-group | Files | REQs in this milestone |
|---|---|---|
| WG-A | `SKILL.md`, `workflows/moai.md`, `references/reference.md` | REQ-SKF-008, 009, 010, 011, 012, 033 |
| WG-B | `workflows/mx.md`, `workflows/clean.md`, `workflows/feedback.md` | REQ-SKF-016, 017, 018 |
| WG-D | `workflows/run/phase-execution.md`, `workflows/run/context-loading.md`, `workflows/run/task-decomposition.md`, `workflows/run/mode-orchestration.md` | REQ-SKF-019, 020, 021, 022 |
| WG-E | `team/run.md`, `team/glm.md` | REQ-SKF-013, 014, 015 |
| WG-F | `workflows/sync/doc-execution.md`, `workflows/sync/quality-gates-quality.md`, `workflows/sync/quality-gates-context.md` | REQ-SKF-023, 024, 025, 026, 027 |
| WG-G | `workflows/project/meta-harness.md`, `workflows/project.md`, `workflows/project/doc-generation.md` | REQ-SKF-028, 029, 030, 031, 032 |
| WG-H | `workflows/harness.md` | REQ-SKF-034, 035 |

### M3 — MINOR (P2, consolidated 17 REQs)

| Write-group | Files | REQs in this milestone |
|---|---|---|
| WG-A | `SKILL.md`, `workflows/moai.md`, `workflows/plan.md` | REQ-SKF-043, 044, 045, **050** (SKILL.md fix-entry team-mode pointer — real assigned edit, not cross-ref; see D3 resolution) |
| WG-B | `references/mx-tag.md`, `workflows/mx.md`, `workflows/fix.md`, `workflows/review.md` | REQ-SKF-036, 037, 038, 039, **050** (`workflows/fix.md:43` naming fix — real assigned edit, not cross-ref; see D3 resolution) |
| WG-C | `workflows/plan/clarity-interview.md`, `workflows/plan/*` (all 3 files), `workflows/plan/spec-assembly.md` | REQ-SKF-040, 041, 042 |
| WG-D | `workflows/run/task-decomposition.md`, `workflows/run/phase-execution.md`, `workflows/run/context-loading.md`, `workflows/run/mode-orchestration.md` | REQ-SKF-046, 047 |
| WG-E | `team/debug.md`, `team/glm.md` | REQ-SKF-050 (`team/debug.md` spawn-instruction normalization only — the SKILL.md and fix.md sub-clauses belong to WG-A and WG-B respectively), 051 |
| WG-F | `workflows/sync/doc-execution.md`, `workflows/sync/delivery.md`, `workflows/sync/quality-gates-quality.md` | REQ-SKF-052 |
| WG-G | `workflows/project/mode-detection.md`, `workflows/project/doc-generation.md` | REQ-SKF-049 |
| WG-H | `workflows/harness.md` | REQ-SKF-048 |

### M4 — CI Neutrality Hardening (cross-cutting, run LAST after M1-M3 land)

| Write-group | Files | REQs in this milestone |
|---|---|---|
| WG-I | `internal/template/internal_content_leak_test.go`, `.github/workflows/template-neutrality-check.yaml` (review only) | REQ-SKF-053 |

M4 runs last so its "confirm the extended regex now catches the M1-M3 leaks retroactively" verification step (§B.3) has real fixed-content to test against.

## §G. Write-Group File Enumeration (full 42-file accounting)

Every write-group is file-disjoint — no file appears in two write-groups. Mirror status recorded from `diff -q` at plan-authoring time (2026-07-08); re-verify at run-phase start per §C.

**WG-A** (Core Entry/Reference, 4 files, all template-mirrored, all currently IDENTICAL):
`SKILL.md`, `workflows/moai.md`, `references/reference.md`, `workflows/plan.md`

**WG-B** (Tag/Utility, 6 files, all template-mirrored; `workflows/review.md` currently DIFFERS — see §B.1):
`workflows/mx.md`, `references/mx-tag.md`, `workflows/clean.md`, `workflows/feedback.md`, `workflows/fix.md`, `workflows/review.md`

**WG-C** (Plan workflow subfolder, 3 files, all template-mirrored; `workflows/plan/clarity-interview.md` currently DIFFERS — see §B.1):
`workflows/plan/spec-assembly.md`, `workflows/plan/clarity-interview.md`, `workflows/plan/context-discovery.md`

**WG-D** (Run workflow, 5 files, all template-mirrored; `workflows/run/task-decomposition.md` currently DIFFERS — see §B.1):
`workflows/run.md` (no findings — pass-through only; confirmed via plan-audit review D5 — no REQ-SKF-NNN cites this file), `workflows/run/task-decomposition.md`, `workflows/run/phase-execution.md`, `workflows/run/context-loading.md`, `workflows/run/mode-orchestration.md`

**WG-E** (Team, 6 files, all template-mirrored, all currently IDENTICAL):
`team/run.md`, `team/glm.md`, `team/debug.md`, `team/plan.md` (no findings — pass-through only), `team/review.md` (no findings — pass-through only), `team/sync.md` (no findings — pass-through only)

**WG-F** (Sync workflow, 5 files, all template-mirrored, all currently IDENTICAL):
`workflows/sync.md` (no findings — pass-through only), `workflows/sync/delivery.md`, `workflows/sync/doc-execution.md`, `workflows/sync/quality-gates-quality.md`, `workflows/sync/quality-gates-context.md`

**WG-G** (Project workflow, 5 files, all template-mirrored, all currently IDENTICAL):
`workflows/project.md`, `workflows/project/doc-generation.md`, `workflows/project/meta-harness.md`, `workflows/project/mode-detection.md`, `workflows/project/codebase-analysis.md` (no findings — pass-through only)

**WG-H** (Harness, 3 skill files + 1 narrow Go-comment exception, all skill files template-mirrored and currently IDENTICAL):
`workflows/harness.md`, `workflows/harness-build-entry.md` (pass-through verification only — no findings), `workflows/harness-builder.md` (pass-through verification only — no findings), **`internal/cli/root.go:157-160`** (comment-only edit, REQ-SKF-007 sub-clause (b) — the one narrow exception to the `.claude/skills/moai` module scope per spec.md §B; no template-mirror concept applies to this file since it is Go source, not a skill doc)

**WG-I** (CI neutrality guard, 2 files, NOT skill-body files, no template mirror concept applies):
`internal/template/internal_content_leak_test.go`, `.github/workflows/template-neutrality-check.yaml`

**No-finding pass-through files (13 total — reconciled per plan-audit D4/D5)**: 0 changes required; each is enumerated above within its owning write-group with a "(no findings — pass-through only)" annotation. Full list, cross-referenced to spec.md §E: `references/anti-patterns.md`, `references/file-reading-optimization.md`, `workflows/codemaps.md`, `workflows/loop.md`, `workflows/gate.md` (5, no write-group — never had findings, listed only for 42-file completeness), `team/plan.md`, `team/review.md`, `team/sync.md` (3, WG-E), `workflows/sync.md` (1, WG-F), `workflows/project/codebase-analysis.md` (1, WG-G), `workflows/harness-build-entry.md`, `workflows/harness-builder.md` (2, WG-H), `workflows/run.md` (1, WG-D). Verify no regression via a single `diff -q` re-check per file after WG runs complete.

Total: WG-A(4) + WG-B(6) + WG-C(3) + WG-D(5) + WG-E(6) + WG-F(5) + WG-G(5) + WG-H(3 skill files) + no-write-group no-finding(5) = **42 files** (matches audit scope; of these, 29 carry ≥1 REQ and 13 are no-finding pass-through). WG-I (2 files) and WG-H's `root.go` comment target are additive, outside the 42-file skill-body scope.

## §H. Technical Approach

1. **Per write-group, single implementing agent (manager-develop, cycle_type=ddd)** reads all assigned REQs' target files fully (not just the cited line ranges — line numbers drift), applies template-source edits first, then `make build`, then syncs, then re-verifies byte-identity.
2. **Grep-based self-verification** after each write-group's edits (see `acceptance.md` for the exact grep assertions per REQ) — run BEFORE declaring the write-group complete.
3. **WG-I runs last**, after WG-A..WG-H land, so its "regex now catches the fixed leaks retroactively" check (§B.3) has real content.
4. **Cross-write-group coordination**: REQ-SKF-050 splits real assigned edits across WG-E (`team/debug.md`), WG-A (`SKILL.md` pointer), and WG-B (`workflows/fix.md` naming) — see §B.2. A soft (non-blocking) sequencing recommendation: land WG-E's `team/debug.md` naming fix before WG-A's `SKILL.md` pointer text is finalized, since the pointer text should name the corrected path.

## §I. Anti-Patterns (avoid during run-phase)

- Do NOT edit `.claude/skills/moai/**` local files without first editing the `internal/template/templates/` mirror (Template-First Rule violation).
- Do NOT introduce new internal SPEC-ID/REQ/AC/C-PH citations into template-mirrored prose while "fixing" one of the T4 findings — the fix must REMOVE the leak, not relocate it.
- Do NOT bulk-`sed`/blind-overwrite the 3 files noted in §B.1 as currently DIFFERING from template — read both sides first.
- Do NOT widen the `internal_content_leak_test.go` regex so broadly it produces false positives against legitimate non-leak content elsewhere in the template tree — run the full suite before AND after (§B.3).
- Do NOT treat "MAJOR" or "MINOR" REQ counts as literal 1:1 with the original 32/31 finding counts — several REQs consolidate 2-5 sub-findings sharing one root cause and file; the AC matrix in `acceptance.md` enumerates the original findings at the sub-bullet level for traceability.

## §J. Cross-References

- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (Route A/B Frozen HARD) — ground truth for REQ-SKF-001, 002.
- `.moai/config/sections/harness.yaml` — ground truth for REQ-SKF-003, 025.
- `internal/harness/frozen_guard.go` `frozenPrefixes` (4 entries, verified) — ground truth for REQ-SKF-005.
- `internal/cli/harness_route.go` `newHarnessRouterCmd()` + `internal/cli/harness_retirement_test.go` `TestHarnessV3R5VerbSurface` + `internal/cli/root.go:157-166` — ground truth for REQ-SKF-007 (ALL harness verbs are Go-binary Cobra subcommands; no learning-vs-v4 split; `root.go:157-160` carries the stale comment that seeded the doc error).
- `.claude/rules/moai/workflow/team-protocol.md` § Role Matrix — ground truth for REQ-SKF-011, 013, 022.
- `.moai/config/sections/workflow.yaml` `role_profiles` — ground truth for REQ-SKF-011, 013, 022.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` `lifecycle` enum — ground truth for REQ-SKF-024.
- `.claude/agents/moai/sync-auditor.md` rubric — ground truth for REQ-SKF-026.
- `internal/template/internal_content_leak_test.go` — target of REQ-SKF-053.
- `CLAUDE.local.md` §2 (Template-First), §15 (16-language neutrality), §25 (Template Internal-Content Isolation) — cross-cutting constraints for all write-groups.
