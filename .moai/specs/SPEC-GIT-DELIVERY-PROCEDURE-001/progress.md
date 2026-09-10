# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.2.0 — the SPEC was split by operator decision T1 (via lead, 2026-09-11); 0.1.3 was committed as `87988e946` and is the input for card t658.
- **Tier**: M (re-classified from the reduced counts; spec.md §C.4). **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: HEAD `87988e946`; the four scope files (local and template), `manager-git.toml`, the guarding test files, and `Makefile` have no diff against `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` (`git diff --stat` empty, exit 0), so plan-time counts on these copies are base counts.
- **Audit count**: restarts at iteration 1 for this reduced SPEC. Earlier iterations on the full scope: iteration 1 FAIL 0.67 (`.moai/reports/t622/plan-audit.md`), iteration 2 FAIL 0.75 (`.moai/reports/t622/plan-audit-iter2.md`). No audit ran on 0.1.3.
- **Pre-write self-checks**: SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → `PASS` (re-run for 0.2.0 in the return). Frontmatter 12 canonical fields present.
- **Scope retained**: AC-11 fetch ordering (manager-git.md:156, agent-common-protocol.md Pre-Spawn), SX-R04 merge method (delivery.md 343/355, manager-git.md:114), OD-2 = B (delivery.md 335-338 and 348-349, doc-execution.md 34-36), plus parity, `.toml` regeneration, neutrality, always-loaded-rule-last ordering, and a Frozen non-touch guard.
- **Numbering**: original IDs kept; moved (REQ 007-010, 016-023; AC 007-010, 017-024) and withdrawn (011, 012) numbers stay as one-line placeholders so REQ-GDP-001..024 and AC-GDP-001..025 each appear exactly once (plan-auditor MP-1, `.claude/agents/moai/plan-auditor.md:138`). Renumbering was rejected because it would give one token two meanings between this document and commit `87988e946`.
- **Counts**: REQ 24 numbers — judged 10 (001-006, 013-015, 024), moved 12, withdrawn 2. AC 25 numbers — judged 11 (001-006, 013-016, 025), moved 12, withdrawn 2.
- **Frozen non-touch check (plan time)**: `[ZONE:Frozen]` lines in the scope files — agent-common-protocol.md line 17 only (local and template); manager-git.md, delivery.md, doc-execution.md 0. zone-registry.md entries whose `file:` is manager-git.md, delivery.md, or doc-execution.md: 0. Frozen entries for agent-common-protocol.md: CONST-V3R2-006 (line 13), 036 and 038 (line 17), 037 (line 52), all `#user-interaction-boundary`. Retained edit sites (acp 290-305, manager-git 114 and 156, delivery 335-338/343/348-349/355, doc-execution 34-36) lie outside every Frozen block and registered clause.
- **Test coverage decision (AC-GDP-013)**: `workflowOptMirroredPaths` and `lateBranchMirroredPaths` contain none of the scope files, so `TestRuleTemplateMirrorDrift` and `TestLateBranchTemplateMirror` are deselected. `sanitizedPairPaths` (`sanitized_pair_parity_test.go:71`) contains agent-common-protocol.md → `TestSanitizedPairParity`; `TestTemplateNoInternalContentLeak` (`internal_content_leak_test.go:1535`) walks the whole template root. manager-git.md, delivery.md, doc-execution.md have diff as their only parity guard.
- **Post-write verification (0.2.0, template copies identical to `b412f8a33`)**:
  - MP-1: `**REQ-GDP-NNN**` definitions in spec.md equal REQ-GDP-001..024 exactly once each (diff against the generated sequence exit 0). Judged 10, moved placeholders 12, withdrawn 2. AC `### AC-GDP-NNN — ` headings 11.
  - AC-GDP-001: Synchronization section paragraphs with fetch and rev-list 1, auto-fail 1, list-group 0.
  - AC-GDP-002: Pre-Spawn block 9 lines; fetch-without-rev-list 1; standalone rev-list at block line 5 (exit 0); joined line 0; rev-list count 1; session command 1; interpretation table rows 8.
  - AC-GDP-004: delivery.md `gh pr merge --squash --delete-branch` 2; `merge_method` 0.
  - AC-GDP-005: `gh pr merge … --squash` lines — manager-git.md 32 and 114, delivery.md 343 and 355, manager-git.toml 26 and 108; agent-common-protocol.md 0, doc-execution.md 0; default-description sentence in manager-git.md 1.
  - AC-GDP-006: delivery Step 3.4 section 52 lines, default-merge hits at section lines 8 and 20; doc-execution subsection 9 lines, hit at line 8; `manager-[g]it[.]md` mentions 0 and 0; opt-in sentences 1 and 1.
  - AC-GDP-013: local/template diff — manager-git.md exit 0, agent-common-protocol.md exit 0, delivery.md exit 1 (`275c275`, `278c278`, `479,480c479`), doc-execution.md exit 1 (`138,143d137`). `func TestSanitizedPairParity(` and `func TestTemplateNoInternalContentLeak(` each exist once; no longer names share those prefixes.
  - AC-GDP-014: `gh pr merge <PR> --<merge_method> --delete-branch` in manager-git.md 0 and in manager-git.toml 0 (base still `--squash` at toml:108).
  - AC-GDP-015: controls on spec.md — SPEC-ID lines 8, REQ lines 45, date lines 12; SHA-shaped tokens 49 (87988e946 x35, b412f8a33 x5, b7447cb90 x3, 980ccdc56 x1, c352330d3 x1, and the diff hunk strings 275c275, 278c278, 480c479, 143d137). Fixture: letter tokens 5, digit tokens 2.
  - AC-GDP-025: the four registered Frozen clauses — 1 each in local and template agent-common-protocol.md; `[ZONE:Frozen]` lines — agent-common-protocol.md 1 (L and T), the other three files 0 (L and T); zone-registry.md entries naming the other three files 0 (L and T). Mutant (template copy with "MUST NOT prompt" lowercased) → CONST-V3R2-036 clause count 0, diff exit 1 with 2 changed lines containing `[ZONE:Frozen]` → FAIL.
- **plan_status**: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
