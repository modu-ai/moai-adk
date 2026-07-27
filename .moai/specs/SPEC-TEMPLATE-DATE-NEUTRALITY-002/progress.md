# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-27
spec_id: SPEC-TEMPLATE-DATE-NEUTRALITY-002
tier: L
baseline_head: 760f09f73
branch: spec/template-date-2025
```

### Artifacts

| Artifact | Status |
|---|---|
| `spec.md` | authored — 23 requirements, 9 out-of-scope sub-sections |
| `plan.md` | authored — 6 milestones, 10 known issues, PRESERVE list |
| `acceptance.md` | authored — 28 criteria, all baselines executed and recorded |
| `design.md` | authored — coupled-ordering, masking hazard, taxonomy rationale |
| `research.md` | authored — full measurement record, 1 refuted hypothesis |
| `progress.md` | this file |

Tier L artifact set (5 files) complete; `progress.md` is emitted at every Tier and is not counted in the Tier total.

### Counts

| Metric | Value |
|---|---:|
| Requirements (`REQ-TDN2-001` … `REQ-TDN2-023`) | 23 |
| Acceptance criteria (`AC-TDN2-001` … `AC-TDN2-028`) | 28 |
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
| 1 | Are version-history records factual records worth keeping, or internal history that should not ship? | 14 | none — per-row |
| 2 | Does a `Created:` stamp follow the prose-stamp REMOVE rule, or is it distinguishable? | 3 | none — per-row |
| 3 | Does removing a mid-line stamp from a composite footer constitute a placeholder substitution? | 2 | none — per-row |

None blocks plan-phase completion; each has its measurement attached in `research.md` §J and an owning milestone in `plan.md` M2.

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
