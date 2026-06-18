# Progress — SPEC-V3R6-SKILL-DECISION-HEURISTICS-001

> Lifecycle progress tracker. §E section skeleton emitted at plan-phase; run/sync/Mx evidence populated by downstream agents per the artifact ownership matrix.

---

## §A. Phase Status

| Phase | Status | Owner | Commit |
|---|---|---|---|
| Plan | complete | manager-spec | _<plan-phase commit>_ |
| Run | complete | manager-develop | e7a4d5a4f (FF-pushed to origin/main) |
| Sync | complete | orchestrator-direct | b6b2243f9 |
| Mx | complete | orchestrator-direct | a9f7adda9 |

---

## §B. Milestone Tracker

| Milestone | Status | Notes |
|---|---|---|
| M1 — moai-foundation-core + moai-workflow-spec edits | complete | DH sections (5 heuristics each, §-pointers) + provenance one-liners (core: pending-form; spec: AP-SRN-004 2026-05-25 64310df3f) |
| M2 — moai-foundation-cc + moai-meta-harness edits | complete | cc: DH (5) + provenance (CONST-V3R5-038 pending-form); meta-harness: DH (5) deliverable (a) ONLY — M2.5 provenance N/A (0 evolvable markers, no fabrication) |
| M3 — frontmatter lint + grep verification gate | complete | grep 4/4, line-count ≤13 all PASS, additive-only diff (54 ins / 0 del), markers 3/3/3/0 unchanged, template source untouched, FU-1 flag present |

---

## §C. AC Status

| AC ID | Status | Evidence |
|---|---|---|
| AC-SDH-001 | PASS | `grep -rc "## Decision Heuristics"` → 1/1/1/1 = 4 total (1 per file) |
| AC-SDH-002 | PASS | 5 `if X → default Y` heuristics per section (each within 3-5 bound), each with `(<- §...)` pointer |
| AC-SDH-003 | PASS | AC-SDH-003 awk line-count: core=8, spec=8, cc=8, meta=13 — all ≤13 PASS (binary) |
| AC-SDH-004 | PASS | `git diff \| grep '^-'` content-deletion check EMPTY; 54 insertions / 0 deletions; every heuristic carries `(<- §section)` pointer |
| AC-SDH-005 | PASS-WITH-DEBT | 3 evolvable-bearing skills bound (core pending-form, spec AP-SRN-004 dated, cc CONST-V3R5-038 pending-form); meta-harness N/A (0 markers, §F A-3 + M3.6) |
| AC-SDH-006 | PASS | only cited date 2026-05-25 + commit 64310df3f verified present in memory (AP-SRN-004); core/cc use pending-form (no dates) |
| AC-SDH-007 | PASS | skill frontmatter unchanged (M3.1a NONE for all 4); spec lint 0 errors (1 expected StatusGitConsistency warning resolved by this draft→in-progress transition) |
| AC-SDH-008 | PASS | `grep -nE 'FU-1\|Template mirror sync\|template-source mirror' plan.md` → 2 hits (L89 M3.7, L112 FU-1) |

---

## §D. Follow-Up Flags

| FU ID | Status | Notes |
|---|---|---|
| FU-1 — Template mirror sync (`internal/template/templates/.claude/skills/`) | open | Separate step after this SPEC's LOCAL edits land (C-4 scope protection). See plan.md §I. |
| FU-2 — Extend Decision Heuristics to remaining skills | open | Out of scope; future SPEC if pilot validates the device. |
| FU-3 — Provenance refresh from newer memory incidents | open | Out of scope; periodic refresh SPEC. |

---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts complete (iter-2): spec.md (§A-§J, 9 REQs in GEARS notation), plan.md (§A-§J, 3 milestones M1-M3 with M3.6/M3.7 added iter-2, 3 follow-up flags FU-1..3), acceptance.md (8 ACs, 5 closure gates, 8 GWT scenarios, 7 edge cases, M3.1-M3.7 test commands), progress.md (this file, §E skeleton).

SPEC ID self-check: `SPEC-V3R6-SKILL-DECISION-HEURISTICS-001` — decomposition: SPEC ✓ | V3R6 ✓ | SKILL ✓ | DECISION ✓ | HEURISTICS ✓ | 001 ✓ → PASS.

Frontmatter: 12 canonical fields present, `era: V3R6`, `status: draft`, `created: 2026-06-18`, `updated: 2026-06-18`, `tags: "skills,craft,heuristics,harness"` (comma-separated string, not labels array), `priority: P3`, `lifecycle: spec-anchored`. version bumped 0.1.0 → 0.2.0 (iter-2).

Honest scope reframing documented in spec.md §A: the 4 target skills carry no inline `AP-*` codes today; deliverable (b) binds provenance to the existing evolvable rationalization/red-flag rows instead.

iter-2 honest evolvable baseline (spec.md §F A-3, verified 2026-06-18 via `grep -c 'moai:evolvable-start' SKILL.md`): moai-foundation-core=3, moai-workflow-spec=3, moai-foundation-cc=3, **moai-meta-harness=0**. Deliverable (b) is N/A for moai-meta-harness (deliverable (a) only); no evolvable content fabricated.

iter-2 resolved 5 plan-auditor defects (iter-1 FAIL 0.74 → iter-2 pending re-audit): D1 §H Out of Scope h3 (MissingExclusions ERROR → lint clean); D2 §F A-3 + REQ-SDH-005/AC-SDH-005/M2.5/M3.6 honest moai-meta-harness N/A; D3 AC-SDH-008 added for REQ-SDH-009 (was "(meta)"); D4 binary ≤13 PASS / ≥14 VIOLATION threshold across C-2/M3.3/AC-SDH-003; D5 REQ-SDH-009 subject `[<plan.md>]` → `[plan-phase artifacts]`. REQ count 9 (unchanged); AC count 7→8.

Ready for Implementation Kickoff Approval (CLAUDE.local.md §19.1) before any run-phase delegation.

---

## §E.2 Run-phase Evidence

Run-phase executed by manager-develop (cycle_type=additive-doc; the "test" is the M3 grep/lint verification gate — no Go code in this SPEC). All 4 LOCAL SKILL.md edited; template mirror (FU-1) deferred out of scope per C-4.

### AC PASS/FAIL Matrix (run-phase verified, verbatim command output)

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-SDH-001 | PASS | `grep -rc "## Decision Heuristics" .claude/skills/moai-{foundation-core,workflow-spec,foundation-cc,meta-harness}/SKILL.md` | `foundation-core:1  workflow-spec:1  foundation-cc:1  meta-harness:1` → 4 total |
| AC-SDH-002 | PASS | manual count of `- If ... default ...` lines per DH section | 5 heuristics each (within 3-5 bound); each ends with `(<- §<section>)` pointer |
| AC-SDH-003 | PASS | `awk '/^## Decision Heuristics/{flag=1;next} /^## /{flag=0} flag' "$f" \| wc -l` | core=8, spec=8, cc=8, meta=13 — all `≤13 → PASS` (binary) |
| AC-SDH-004 | PASS | `git diff <4 skills> \| grep -E '^-' \| grep -vE '^---\|^-+$\|^-\s*$'` | (empty) — no content deletions; `--stat` = 54 insertions(+), 0 deletions |
| AC-SDH-005 | PASS-WITH-DEBT | provenance one-liners appended INSIDE existing red-flags blocks | core: pending-form; spec: `AP-SRN-004 — recurred on 2026-05-25 in Sprint 10 paste-ready chore (commit 64310df3f)`; cc: CONST-V3R5-038 pending-form; meta-harness: N/A (0 markers) |
| AC-SDH-006 | PASS | `grep -hoE "recurred on [0-9]{4}-..." <3 skills> \| sort -u` then memory grep | only date `2026-05-25` cited; `grep -l "64310df3f" ~/.claude/projects/.../memory/*.md` returns matches → verified, no fabrication |
| AC-SDH-007 | PASS | per-file frontmatter diff grep + `moai spec lint .../spec.md` | skill frontmatter diff = NONE (all 4); spec lint `0 error(s), 1 warning(s)` (StatusGitConsistency, resolved by draft→in-progress) |
| AC-SDH-008 | PASS | `grep -nE 'FU-1\|Template mirror sync\|template-source mirror' plan.md` | L89 (M3.7) + L112 (FU-1) → 2 hits in plan.md §I |

### Invariant checks

| Invariant | Status | Evidence |
|-----------|--------|----------|
| FL-2 marker balance preserved | PASS | start==end per file: core 3/3, spec 3/3, cc 3/3, meta 0/0 (unchanged from §F A-3 baseline) |
| M3.6 meta-harness deliverable (b) honest N/A | PASS | `grep -c 'moai:evolvable-start' meta-harness/SKILL.md` = 0 (no fabrication); core/spec/cc each = 3 (provenance appended INSIDE existing blocks) |
| EC-6 template source untouched | PASS | `git status --porcelain internal/template/templates/.claude/skills/` = (empty) |
| C-5 additive-only | PASS | `git diff --stat` = 4 files, 54 insertions(+), 0 deletions; all `+` content lines |
| C-2 ≤13-line scroll bound | PASS | all 4 DH sections ≤13 lines (binary threshold) |

### Provenance honesty ledger (REQ-SDH-006 / C-3)

| Skill | Anti-pattern family | Form used | Memory-verified? |
|-------|---------------------|-----------|------------------|
| moai-foundation-core | orchestrator-self-execution / delegation-bypass | pending-form (`observed recurrence, provenance pending in memory`) | N/A — no date cited |
| moai-workflow-spec | AP-SRN-004 (Wave→Round retired naming) | dated: `recurred on 2026-05-25 ... (commit 64310df3f)` | YES — `64310df3f` + `2026-05-25` present in memory |
| moai-foundation-cc | frontmatter CSV-format drift (CONST-V3R5-038) | pending-form | N/A — no date cited |
| moai-meta-harness | (deliverable b N/A — 0 evolvable markers) | none | N/A |

No fabricated dates/SPEC-IDs/SHAs. Pending-form is the only fallback used per C-3.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-19
run_commit_sha: "e7a4d5a4fe13bff7179832b1925ed7b89d784f40"  # FF-pushed d78ed9885..e7a4d5a4f to origin/main
run_status: implemented
ac_pass_count: 8           # AC-SDH-001..004, 006, 007, 008 PASS + AC-SDH-005 PASS-WITH-DEBT
ac_fail_count: 0
ac_pass_with_debt_count: 1 # AC-SDH-005 (SHOULD; 2 of 3 provenance bindings use honest pending-form)
preserve_list_post_run_count: 0   # no PRESERVE-list files modified outside the 4 skills + progress.md
new_warnings_or_lints_introduced: 0   # skill frontmatter unchanged; only expected StatusGitConsistency on spec.md (resolved by transition)
cross_platform_build: not_applicable  # no Go code in this SPEC (additive-doc cycle)
total_run_phase_files: 6   # 4 SKILL.md (LOCAL) + spec.md (frontmatter transition) + progress.md (§E evidence)
m1_to_mN_commit_strategy: single-commit  # M1+M2+M3 bundled (Tier S additive-doc, no build/test gate between milestones)
status_transition: "draft → in-progress (this run-phase commit; updated 2026-06-18 → 2026-06-19)"
template_mirror_followup: open  # FU-1 deferred per C-4 (plan.md §I)
```

---

## §E.4 Sync-phase Audit-Ready Signal

- **sync_complete_at**: 2026-06-19
- **sync_artifacts**: CHANGELOG.md `[Unreleased] → Added` entry; spec.md frontmatter `in-progress → implemented`.
- **change_scope**: docs-only — no Go/template change. FU-1 template mirror remains deferred (C-4 scope protection; forward follow-up, not blocking close).
- **owner**: orchestrator-direct (GLM manager-docs spawn context-limit fallback per `feedback_glm_orchestrator_direct_sync_mx`).
- **spec-lint**: 0 errors.

sync_commit_sha: b6b2243f9

---

## §E.5 Mx-phase Audit-Ready Signal

- **mx_complete_at**: 2026-06-19
- **frontmatter_transition**: `implemented → completed`.
- **@MX tags**: N/A (additive markdown craft device — 4 SKILL.md `## Decision Heuristics` sections; no Go code, no high fan_in invariant surface).
- **4-phase close**: plan (plan-phase) → run (e7a4d5a4f) → sync (b6b2243f9) → mx (this commit) — V3R6 era closed.

mx_commit_sha: a9f7adda90ed82396184fa5f3763431e5778b1d2

---


Version: 0.2.0 (run-phase complete — M1/M2/M3, 8 AC PASS incl. 1 PASS-WITH-DEBT)
Status: in-progress
Last Updated: 2026-06-19
