# progress.md — SPEC-CONTEXT-ENGINE-RIGHTSIZE-001

> Tier M progress tracker. Skeleton emitted at plan-phase; §E.2-§E.4 are placeholder headings only (per canonical §E skeleton generation policy — era.go parses literal `§E.2`/`§E.3`/`§E.4` tokens). manager-spec populates only §E.1; run/sync evidence is owned by manager-develop / manager-docs respectively.

---

## §A. Status

- **Phase**: plan (manager-spec)
- **Status**: draft
- **Tier**: M (3 artifacts: spec.md + plan.md + acceptance.md)
- **Plan-phase commit**: pending

---

## §B. Milestone Tracker

| Milestone | Title | Status | Commit SHA |
|---|---|---|---|
| M1 | Expressive transition: code_comments line | ⬜ pending | — |
| M2 | Tool Selection consolidation + informational reframing | ⬜ pending | — |
| M3 | Template mirror synchronization + §25 neutralization | ⬜ pending | — |
| M4 | Regression verification (A-group + C-group + lint parity) | ⬜ pending | — |

Legend: ⬜ pending / 🟡 in-progress / ✅ done / ⚠️ blocked

---

## §C. Decision Log

| Date | Decision | Rationale | Authority |
|---|---|---|---|
| 2026-07-28 | Conservative "B-group only" scope | GOOS decision; preserve all A-group Frozen + C-group mechanical guardrails | GOOS |
| 2026-07-28 | Tier M classification | 3 files, no Go changes, no architectural decisions deferred | manager-spec |
| 2026-07-28 | M1.3 = verification-only (no defect) | Direct grep confirmed SSOT reference already exists at `plan-auditor.md:~144` | manager-spec (per `feedback_defect_claim_verification`) |
| 2026-07-28 | M1.4 (CLAUDE.md Tool Selection) out of scope | Direct grep confirmed CLAUDE.md has no such section | manager-spec |

---

## §D. Baseline Captures (pre-edit, locked at plan-phase)

| Baseline | Value | Source |
|---|---|---|
| `[HARD]` in CLAUDE.md | 15 | `grep -c '\[HARD\]' CLAUDE.md` |
| `[ZONE:Frozen]` in `.claude/rules/moai/` | 66 across 13 files | `grep -rc '\[ZONE:Frozen\]'` |
| `[ZONE:Evolvable]` in `.claude/rules/moai/` | 98 | `grep -rc '\[ZONE:Evolvable\]'` |
| `MUST` in `.claude/rules/moai/` | 269 | `grep -rc '\bMUST\b'` |
| `NEVER` in `.claude/rules/moai/` | 14 | `grep -rc '\bNEVER\b'` |
| `moai-constitution.md` "Use X instead of Y" bullets | 5 | `grep -c '^- Use .* instead of'` |
| `moai-constitution.md` "English comments" line | 1 (line ~77) | `grep -n 'English comments'` |
| `plan-auditor.md` SSOT cross-reference | observable (line ~144) | `grep -n 'agent-common-protocol.md.*Tool Selection by Task'` |

---

## §E.1 Plan-phase Audit-Ready Signal

- **plan_status**: audit-ready (plan-auditor PASS 0.84; D1/D2/AC-CER-006b defects corrected + directly verified 2026-07-29)
- **plan_complete_at**: 2026-07-29
- **artifact_set**: spec.md + plan.md + acceptance.md + progress.md (Tier M)
- **frontmatter_validated**: 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + optional (tier, related_specs)
- **SPEC ID regex**: PASS (`SPEC-CONTEXT-ENGINE-RIGHTSIZE-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`)
- **out_of_scope_section**: present (8 `### Out of Scope — <topic>` H3 sub-headings)
- **applied_lessons_linked**: feedback_claimed_correction_never_applied, feedback_defect_claim_verification, feedback_local_template_sync_neutralize_first, feedback_guard_observation_must_be_falsifiable, feedback_guard_signal_proves_call_not_effect, feedback_plan_commit_subject_feat_prefix, feedback_shared_checkout_concurrent_commit_race, feedback_index_level_commit_shared_checkout

---

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop populates>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop populates>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates>_

---
