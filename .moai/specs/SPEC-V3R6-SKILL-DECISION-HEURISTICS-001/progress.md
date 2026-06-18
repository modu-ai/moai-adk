# Progress — SPEC-V3R6-SKILL-DECISION-HEURISTICS-001

> Lifecycle progress tracker. §E section skeleton emitted at plan-phase; run/sync/Mx evidence populated by downstream agents per the artifact ownership matrix.

---

## §A. Phase Status

| Phase | Status | Owner | Commit |
|---|---|---|---|
| Plan | complete (this commit) | manager-spec | _<pending run-phase>_ |
| Run | not started | manager-develop | _<pending>_ |
| Sync | not started | manager-docs | _<pending>_ |
| Mx | not started | manager-docs / orchestrator | _<pending>_ |

---

## §B. Milestone Tracker

| Milestone | Status | Notes |
|---|---|---|
| M1 — moai-foundation-core + moai-workflow-spec edits | not started | _<pending run-phase>_ |
| M2 — moai-foundation-cc + moai-meta-harness edits | not started | _<pending run-phase>_ |
| M3 — frontmatter lint + grep verification gate | not started | _<pending run-phase>_ (M3.6 honest-evolvable N/A + M3.7 template-follow-up flag added iter-2) |

---

## §C. AC Status

| AC ID | Status | Evidence |
|---|---|---|
| AC-SDH-001 | not started | _<pending run-phase>_ |
| AC-SDH-002 | not started | _<pending run-phase>_ |
| AC-SDH-003 | not started | _<pending run-phase>_ |
| AC-SDH-004 | not started | _<pending run-phase>_ |
| AC-SDH-005 | not started | _<pending run-phase>_ (3 evolvable-bearing skills only; N/A for moai-meta-harness) |
| AC-SDH-006 | not started | _<pending run-phase>_ |
| AC-SDH-007 | not started | _<pending run-phase>_ |
| AC-SDH-008 | not started | _<pending run-phase>_ (REQ-SDH-009 template follow-up flag; plan-phase doc-existence check) |

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

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_

---

Version: 0.2.0 (plan-phase iter-2 — 5 plan-auditor defects resolved)
Status: draft
Last Updated: 2026-06-18
