---
id: SPEC-V3R5-LATE-BRANCH-001
title: "Late-Branch Workflow — Compact Extract"
version: "0.3.0"
status: completed
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.0.0 — Round 5"
module: ".moai/config/sections/git-strategy + .claude/skills/moai/workflows/plan/spec-assembly + .claude/skills/moai/SKILL.md + .claude/agents/moai/manager-git + .claude/rules/moai/workflow/spec-workflow + internal/template/templates (5 mirrors)"
lifecycle: spec-anchored
tags: "workflow, late-branch, git-strategy, dogfooding, compact, mega-sprint, v3r5"
---

> Auto-extracted compact summary from spec.md / plan.md / acceptance.md. For full context, see those three files. Memory sources: `project_v3r5_late_branch_decision.md`, `feedback_late_branch_workflow.md`, `feedback_no_github_issue_for_specs.md` (v0.1.1 extension).

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.3.0 | 2026-05-20 | GOOS Kim (via MoAI) | Sync-phase complete — status `implemented → completed`. Single-integrated PR pattern (user decision C). docs-site impact: none. Cherry-pick으로 본 SPEC 10 commits만 새 feat branch로 분리, STATUSLINE 4 commits은 main에 보존. See spec.md HISTORY v0.3.0 for full ledger. |
| 0.2.0 | 2026-05-20 | GOOS Kim (via manager-develop ddd) | Run-phase complete — status `draft → implemented`. 7/7 ACs PASS, 6 milestone commits + 1 chore catalog hash sync. See spec.md HISTORY v0.2.0 for full ledger. |
| 0.1.2 | 2026-05-20 | GOOS Kim (via MoAI) | REQ-LB-008 promoted from Optional → Mandatory per plan-auditor iter1 Q2 CRITICAL recommendation. M6 delivers `lateBranchMirroredPaths` + `TestLateBranchTemplateMirror` parallel test. |
| 0.1.1 | 2026-05-20 | GOOS Kim (via MoAI) | Mid-draft policy extension — REQ-LB-009 + AC-LB-007 + EXCL-LB-008 + R-LB-005 + D2 (no-auto-issue policy). 1 new affected file (`.claude/skills/moai/SKILL.md` + mirror). `issue_number` frontmatter field removed (D2). |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial compact extract |

---

## Mission

Formalize the **Late-branch workflow pattern**: SPEC commits accumulate directly on `main`; the `feat/SPEC-*` branch is created ONLY at PR creation time via `git switch -c`; PR squash-merges into main; local main is then reset via `git reset --hard origin/main`. **`mode: team` is preserved** — branch protection + PR/CI gates remain active. User directive (2026-05-20): "1인 메인테이너 환경에서 자동 branch 생성 overhead 제거, PR/CI gate 유지."

---

## 4-Phase Procedure

| Phase | Action | Branch state |
|---|---|---|
| **A** SPEC creation | `git checkout main && git pull origin main` → `/moai plan` → commit SPEC files on main | `main` (no new branch) |
| **B** Implementation | RED/GREEN/REFACTOR commits accumulate on local main | `main` (ahead of origin) |
| **C** PR creation | `git switch -c feat/SPEC-XXX` → push → `gh pr create` → squash merge | `feat/SPEC-XXX` (transient) |
| **D** Local reset | `git checkout main && git reset --hard origin/main && git pull origin main` | `main` (synced) |

---

## Architecture (5 Modifications + 5 Mirrors)

| # | File | Modification |
|---|------|--------------|
| D1 | `.moai/config/sections/git-strategy.yaml` `team` section | 4 keys: `auto_branch`/`auto_pr`/`auto_enabled`: true→false; `prompt_always`: false→true |
| D2 | `.claude/skills/moai/workflows/plan/spec-assembly.md` Phase 3 + Phase 2.5 | Add `if auto_enabled == false: skip branch creation` conditional + "Late-branch" mode display. **v0.1.1**: Phase 2.5 (GitHub Issue Creation) default skip; gated only by explicit `--issue` flag per REQ-LB-009. |
| D3 | `.claude/agents/moai/manager-git.md` | Add `main_late_branch` row to Personal Mode table + new "Late-Branch Invocation Pattern" subsection documenting Phase A→D |
| D4 | `.claude/rules/moai/workflow/spec-workflow.md` | Step 1 entry precondition: `main checkout` requirement when `auto_enabled == false`; Step 4 cleanup: `git reset --hard origin/main` canonical closure |
| D5 | 5 template mirrors under `internal/template/templates/` | byte-equivalent (modulo `.tmpl` vars) propagation of D1-D4 + new `.claude/skills/moai/SKILL.md` mirror (v0.1.1) |
| **NEW v0.1.1** | `.claude/skills/moai/SKILL.md` | `--issue` flag semantics flip: opt-out via `--no-issue` → opt-in via `--issue` (default = skip issue creation) |

---

## D1 Backward-Compatibility Decision

**Selected: Option (a) Breaking default change** — template `.tmpl` flips to `auto_enabled: false`. New projects (`moai init` post-v3.5.0) start in Late-branch mode. Existing projects (`moai update`) unaffected because `moai update` preserves user-modified `git-strategy.yaml`. v3.5.0 release notes call out the breaking default. (Options b/c evaluated and rejected — see plan.md §2 D1.)

## D2 Frontmatter `issue_number` Field Removal (added v0.1.1)

**Decision**: Remove `issue_number` field from spec.md frontmatter for new SPECs from v0.1.1 onward. Existing SPECs retain `issue_number` as immutable history (EXCL-LB-008 — no migration). Distinction from D1: D1 = behavior change (auto-branch default); D2 = metadata simplification (no issue_number on new SPECs). Both share **sole-maintainer workflow optimization** motivation. (See plan.md §2 D2.)

---

## EARS REQ Summary (9 = 8 Mandatory + 1 Optional)

- REQ-LB-001 Ubiquitous: `/moai plan` does NOT create branch when `auto_enabled == false`
- REQ-LB-002 Ubiquitous: `manager-spec` commits SPEC files to main directly (no push)
- REQ-LB-003 Event-Driven: `manager-git` detects Late-branch pattern after manual `git switch -c`
- REQ-LB-004 Event-Driven: Phase 3 reads `auto_enabled` and skips branch creation when `false`
- REQ-LB-005 State-Driven: `manager-git` Personal Mode supports `main_late_branch` option
- REQ-LB-006 State-Driven: `spec-workflow.md` Step 4 documents `git reset --hard origin/main`
- REQ-LB-007 Unwanted: System NEVER `git push origin main` during Phase A/B
- REQ-LB-008 Mandatory (promoted v0.1.2): `rule_template_mirror_test.go` extended with `lateBranchMirroredPaths` + `TestLateBranchTemplateMirror` parallel test (M6 delivered)
- **REQ-LB-009 Mandatory (v0.1.1)**: `/moai plan` MUST NOT auto-create GitHub Issue without explicit `--issue` flag

---

## Binary ACs (7)

| AC | Verification |
|---|---|
| AC-LB-001 | `yq` on 4 `team` keys: returns false/false/false/true |
| AC-LB-002 | `grep -c 'auto_enabled'` + `grep -c 'Late-branch'` in spec-assembly.md: both ≥1 |
| AC-LB-003 | `grep -c 'main_late_branch'` + `grep -c 'Late-Branch Invocation Pattern'` + `grep -cE '^(Phase A|B|C|D)'` in manager-git.md: ≥1, ≥1, ≥4 |
| AC-LB-004 | `grep -c 'main checkout'` + `grep -c 'git reset --hard origin/main'` in spec-workflow.md: both ≥1 |
| AC-LB-005 | `diff` returns 0 lines for 3 .md mirror pairs + `yq` consistent on .yaml.tmpl + `go test ./internal/template/...` PASS |
| AC-LB-006 | E2E scenario: Phase A→D leaves `git status --porcelain` empty AND `main == origin/main`. Dogfooding (this SPEC's own PR cycle) is primary verification path. |
| AC-LB-007 (v0.1.1) | `grep -c "gh issue create"` in spec-assembly.md returns 0, OR every occurrence is gated by an explicit `--issue` flag check |

---

## Edge Cases (4)

- EC-LB-001 Concurrent SPEC work: BODP gate detects dirty tree → AskUserQuestion (a/b/c)
- EC-LB-002 Accidental `git push origin main` during Phase B: branch protection blocks; orchestrator recommends Phase C `git switch -c`
- EC-LB-003 Phase D omitted: Phase A precondition catches divergence; recommends fetch + reset
- EC-LB-004 CI block during Phase C: fixup commits OR abort PR + fix on main + `--force-with-lease` push (admin-squash-override pattern preserved)

---

## Constraints (4)

- C-LB-001: Local main MAY be ahead of `origin/main` during Phase B/C (commit-not-yet-pushed). Pattern explicitly relies on this.
- C-LB-002: One SPEC at a time on a given checkout. Parallel SPECs → worktree pattern (EXCL-LB-005).
- C-LB-003: Phase D `git reset --hard origin/main` MANDATORY post-squash. Skipping causes next `git pull` conflict (R-LB-002).
- C-LB-004: `mode: team` PRESERVED. Branch protection 4 required checks + PR/CI gates remain active. Late-branch only changes WHEN branch is created, not WHETHER PR exists.

---

## Risks (5)

- R-LB-001: Accidental `git push origin main` during Phase A/B. Mitigation: REQ-LB-007 + `manager-git.md` caveat + future pre-push hook SPEC (EXCL-LB-002).
- R-LB-002: Phase D omitted. Mitigation: Step 4 documentation + memory + orchestrator next-invocation detection.
- R-LB-003: Backward compat for existing users. Mitigation: `moai update` preserves user config + v3.5.0 release notes + `prompt_always: true` teaches at first SPEC creation.
- R-LB-004: Template mirror drift. Mitigation: AC-LB-005 sync-PR enforcement + existing `internal/template/rule_template_mirror_test.go` CI guard.
- R-LB-005 (v0.1.1): PR template `closes #N` references break for SPECs lacking `issue_number`. Mitigation: `closes #N` lines remain optional; SPECs without issue_number omit close reference; no tooling change required.

---

## Exclusions

### Out of Scope (8 Exclusions)

- EXCL-LB-001: Migration tool (auto-rewrite existing `git-strategy.yaml`)
- EXCL-LB-002: pre-push hook impl (follow-on SPEC)
- EXCL-LB-003: `main_direct`/`develop_direct` PR-less mode
- EXCL-LB-004: Worktree pattern extension (already in SPEC-V3R4-WORKTREE)
- EXCL-LB-005: Parallel multi-SPEC on single checkout (worktree required)
- EXCL-LB-006: `mode: manual`/`personal` Late-branch activation (only `mode: team` in scope)
- EXCL-LB-007: GUI/IDE plugin integration
- EXCL-LB-008 (v0.1.1): Migration tool retroactively removing `issue_number` from existing SPEC frontmatter (historical SPECs retain `issue_number` as immutable history)

---

## File Touch List (run-phase)

### MODIFIED existing (10 = 5 source + 5 template mirror)

Local:
- `.moai/config/sections/git-strategy.yaml`
- `.claude/skills/moai/workflows/plan/spec-assembly.md`
- `.claude/skills/moai/SKILL.md` (NEW v0.1.1)
- `.claude/agents/moai/manager-git.md`
- `.claude/rules/moai/workflow/spec-workflow.md`

Template mirrors (`internal/template/templates/`):
- `.moai/config/sections/git-strategy.yaml.tmpl`
- `.claude/skills/moai/workflows/plan/spec-assembly.md`
- `.claude/skills/moai/SKILL.md` (NEW v0.1.1)
- `.claude/agents/moai/manager-git.md`
- `.claude/rules/moai/workflow/spec-workflow.md`

### NEW (0)
No new files. Pure configuration + documentation SPEC.

### CI guards (existing, no new)
- `internal/template/rule_template_mirror_test.go` (mirror parity — possibly extended per REQ-LB-008)

---

## AC Count + Dependency Status

| Metric | Value |
|--------|-------|
| EARS REQs | 8 mandatory + 1 Optional (REQ-LB-001..009) |
| Binary ACs | 7 (AC-LB-001..007) |
| Edge Cases | 4 (EC-LB-001..004) |
| Risk Mitigations | 5 (R-LB-001..005) |
| Constraints | 4 (C-LB-001..004) |
| Exclusions | 8 (EXCL-LB-001..008) |
| Total verification surface | 16 (7 ACs + 4 ECs + 5 Rs) |
| REQ ↔ AC traceability | 100% (each REQ maps to ≥1 AC; full table in acceptance.md §1) |

Dependencies:
- **Soft**: W3 HARNESS-AUTONOMY-001 (COMPLETE) — used as structural reference template
- **Soft**: SPEC-V3R5-CONSTITUTION-DUAL-001 (COMPLETE) — admin-squash-override pattern reference for EC-LB-004
- **Orthogonal**: SPEC-V3R4-WORKTREE — worktree pattern boundary (parallel SPECs)
- **Blocks** (potential): SPEC-V3R5-MAIN-PUSH-GUARD-* (follow-on pre-push hook SPEC, EXCL-LB-002)

---

## M1-M5 Milestone List

| Milestone | Priority | Dependencies | Deliverables |
|-----------|----------|--------------|--------------|
| M1 Config switch | P0 | none | git-strategy.yaml + .tmpl (4 keys flipped) |
| M2 Skill body Phase 3 | P0 | M1 | spec-assembly.md Phase 3 conditional + "Late-branch" mode display |
| M3 Agent body | P0 | M1+M2 | manager-git.md `main_late_branch` row + Late-Branch Invocation Pattern subsection |
| M4 Rule update | P0 | M3 | spec-workflow.md Step 1 + Step 4 |
| M5 Template mirror parity | P0 | M1-M4 | 4 mirrors byte-equivalent to local |

Sequential within single run-phase (`manager-develop` cycle_type=ddd, ANALYZE-PRESERVE-IMPROVE per modification — modifying existing files). ~80-150 LOC total across 8 files (4 local + 4 template).

---

## Self-Application (Dogfooding)

This SPEC's plan-phase commits land directly on `main` as the first concrete demonstration of the pattern. `plan/SPEC-V3R5-LATE-BRANCH-001` branch is NOT created during plan-phase drafting — it is created at plan-PR time via `git switch -c plan/SPEC-V3R5-LATE-BRANCH-001`. AC-LB-006 dogfooding verification is empirically achieved by completing the full plan-PR → run-PR → sync-PR cycle with clean Phase D resets.

---

## Recommended Next Action

Invoke **plan-auditor subagent** for iter 1 independent review of this SPEC's plan-phase artifacts (5 files). Anticipated audit dimensions per W3 iter1 pattern: D1 Brief Quality, D2 Phase Decomposition, D3 Risk Management, D4 Frontmatter Compliance, D5 Exclusion Discipline, D6 Lint Baseline. Target: PASS ≥ 0.85.
