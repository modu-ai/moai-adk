# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-27
plan_iteration: 2
spec_id: SPEC-TEMPLATE-DATE-NEUTRALITY-002
tier: L
baseline_head: 760f09f73
iteration_1_commit: f5d3a93bf
branch: spec/template-date-2025
```

### Artifacts

| Artifact | Status |
|---|---|
| `spec.md` | authored — 24 requirements, 9 out-of-scope sub-sections |
| `plan.md` | authored — 6 milestones, 10 known issues, PRESERVE list |
| `acceptance.md` | authored — 32 criteria + §G known limitations, all baselines executed |
| `design.md` | authored — coupled-ordering, masking hazard, taxonomy rationale |
| `research.md` | authored — full measurement record, 1 refuted hypothesis, §K partition derivation |
| `classify.sh` | committed — year-widened classifier, reproduces 74 / 48 / 34 |
| `progress.md` | this file |

Tier L artifact set (5 files) complete; `classify.sh` is the committed measurement instrument (REQ-TDN2-008) and `progress.md` is emitted at every Tier — neither counts toward the Tier artifact total.

### Counts

| Metric | Value |
|---|---:|
| Requirements (`REQ-TDN2-001` … `REQ-TDN2-024`) | 24 |
| Acceptance criteria (`AC-TDN2-001` … `AC-TDN2-032`) | 32 |
| Occurrence-class rows in scope | 74 |
| Distinct findings | 48 |
| Distinct files carrying a finding | 34 |
| REMOVE rows (`DC-2a`) | 28 across 22 files |
| PRESERVE rows (`DC-2b`) | 13 |
| Per-row adjudicated rows (`DC-5`) | 33 |
| Dual-category findings | 4 |
| Open questions carried to M2 | 3 |
| Deferred / blocking questions | 0 |

### Measurement provenance

Every count traces to a command recorded in `research.md`, executed from the worktree root at `760f09f73`. The measurement instrument — a year-widened replica of the predecessor's committed classifier — was validated by reproducing the predecessor's own post-remediation residual of 88 `202[6-9]` rows (`carved out = 100 − k`, `k = 12`).

Two of the task brief's stated figures were re-verified and confirmed exactly (74 occurrences / 48 findings / 34 files; 10 frontmatter-shaped `updated:` lines). One stated hypothesis was **refuted**: the 10 `updated:` lines do not belong to the frontmatter category and are not carved by its structural gate, because all 10 sit at column 0 while the gate requires indentation. See `research.md` §D.

### Open questions (non-blocking, resolved at M2)

| # | Question | Rows | Recorded default |
|---|---|---:|---|
| 1 | Are version-history records factual records worth keeping, or internal history that should not ship? The predecessor's "Known cosmetic residue" note (its `spec.md` §5) is the constraining precedent and is now cited. | 14 | none — per-row; splits into a 2-row bullet form carrying the precedent and a 12-row table form that does not |
| 2 | Does a `Created:` stamp follow the prose-stamp REMOVE rule, or is it distinguishable? | 3 | none — per-row; either outcome is covered (dual-category finding, routed through AC-016) |
| 3 | **Re-posed as an editing instruction**: what is the residual line after the mid-line stamp is excised from each of the 2 named `COMPOSITE` rows? | 2 | candidate residual `Version: 5.0.0 \| Enterprise Ready:` — M2 confirms or replaces |

None blocks plan-phase completion; each has its measurement attached in `research.md` §J and an owning milestone in `plan.md` M2.

Question 3's earlier phrasing ("does removal constitute a placeholder substitution?") was mis-posed — REQ-TDN2-009 already answers it. The real gap was that the two rows were never located, so the shape of the edit could not be reviewed. Both are now named with file and line in `research.md` §J item 3 and §K.

---

## §E.1.1 Iteration-2 audit resolution

Plan-audit iteration 1 returned **FAIL 0.67** against the Tier L 0.85 threshold. All seven must-pass criteria passed and all 17 executable baselines reproduced exactly; the failure was confined to the acceptance and traceability layers. Resolution:

| Finding | Severity | Resolution |
|---|---|---|
| D1 — AC-016 unsatisfiable under a permitted M2 outcome | critical | Target recomputed from the M2 adjudication (`actual == expect` per file). AC-012's adjudication branch preserved. |
| D2 — AC-011 half-dead pattern, no control | major | Replaced with a single-stage unanchored command (catches all 3 shapes incl. mid-line). **Baseline corrected `0` → `9`** after measurement found 9 pre-existing ISO-format doc lines. AC-008 named as control. |
| D3 — AC-012 line-number dependency | major | Re-anchored on the `fenced` rationale marker now required by REQ-TDN2-012. AC-031 added to enforce the §A.4 claim mechanically. |
| D4 — no known-limitations disclosure; edit-scope criterion dropped | major | §G added with 5 disclosed gaps; AC-029 restores the predecessor's edit-scope criterion with `catalog.yaml` pre-excluded. |
| D5 — AC-001 traced to wrong REQ | minor | REQ column `022` → `008`; command re-anchored on the `DATE_RE` line (target `1`, was mis-stated). |
| D6 — 4 requirements uncovered | major | AC-029 (REQ-023), AC-030 (REQ-017), AC-031 (REQ-020), AC-032 (REQ-024) added; REQ-022 disclosed as unverifiable in §G item 1. |
| D7 — HIST question missing its precedent | major | Predecessor's residue note cited in `research.md` §J item 1 and `plan.md` M2. |
| D8 — "path suffix" wrong | minor | Corrected to exact relative path (`entry.File == relPath`) in `plan.md` M4 and `design.md` §C. |
| D9 — fourth year-bearing site omitted | minor | `research.md` §F now tables 4 sites; REQ-TDN2-024 brings the cross-SPEC doc comment into scope; AC-032 verifies the owning test stays green. |
| D10 — three `Where` used as data conditions | minor | REQ-003 / 012 / 015 converted to `When`. |
| Unverified 1-2 — sub-shape partition, COMPOSITE rows | — | `research.md` §K enumerates all 33 `DC-5` rows per code; both COMPOSITE rows named with file, line, and exact text. |
| Unverified 3 — classifier not committed | — | `classify.sh` committed; re-run reproduces 74 rows / 48 findings / 34 files, DC-2a 28 / DC-2b 13 / DC-5 33. |
| Unverified 4 — phantom lint flag | — | No artifact cites `moai spec lint --path`; verified by grep. |

**Two defects were self-inflicted during the repair and caught before commit**: AC-001's target was `1` but the committed `classify.sh` contains the token twice (functional line + header comment), so the command was re-anchored on `^DATE_RE=`; and AC-031's whole-file grep was tripped by this SPEC's own explanatory prose quoting the offending shape, so the prose was reworded and the false-positive mode disclosed.

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
