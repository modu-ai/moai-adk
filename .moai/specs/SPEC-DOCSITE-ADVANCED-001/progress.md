# progress.md — SPEC-DOCSITE-ADVANCED-001

> Plan-phase skeleton. The §E.1 section is populated by manager-spec at
> plan-phase; §E.2-§E.4 are placeholder headings (NOT populated at plan-phase)
> — they belong to manager-develop (run-phase) and manager-docs (sync-phase)
> per the canonical SPEC Artifact Ownership matrix.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-13
plan_artifact_set: Tier-L-6-files
plan_files:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
  - design.md
  - research.md
plan_notes: |
  6-page × 4-locale docs-site content expansion + pre-existing _meta.yaml
  parity debt fix. Tier L (single SPEC, NOT an Epic). All 6 page sources
  verified substantially-ready at plan-phase; zero blockers. M1 is the
  parity-debt pre-fix (hard precondition for clean M5 registration).
  Route A (Hybrid Trunk main-direct) recommended.
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---

## §F Phase 4 Mode Selection

_<pending orchestrator decision before first run-phase Agent() spawn>_

The orchestrator fills this section before the first M1 delegation, per
`.claude/rules/moai/workflow/orchestration-mode-selection.md` §D logging
contract. Recommendation for this SPEC:

- **Input parameters**: tier=L, scope=29 files (24 new + 4 _meta + 1 main.yaml),
  domain count=1 (docs-site only), file language mix=markdown 100%,
  concurrency benefit=LOW (mechanical uniform transform, NOT coding-heavy but
  also NOT research-heavy).
- **Recommended mode**: Mode 5 (sub-agent sequential) — the 6 pages can be
  authored sequentially under manager-develop with the oss-docs harness
  specialists (content-author + locale-translator) delegated per page. Mode 4
  (parallel) is justifiable for the 4-locale derivation fan-out within a single
  page, but not for the page-to-page sequence. Mode 6 (workflow) is NOT
  recommended — the work is not high-volume mechanical (24 files but each
  requires substantive authoring, not a uniform transform).
- **Justification**: per Anthropic's coding-task parallelism caveat (most
  coding tasks involve fewer truly parallelizable tasks than research), the
  per-page authoring is sequential; the per-locale derivation within a page is
  parallelizable but low-volume (3 derived locales per page).

---

## §H Recursive Self-Diagnosis Log

_<empty — populated at run-phase only if a DIAGNOSE-PATCH-VERIFY loop fires>_

## §I Token Accounting

_<pending sync-close — populated by token-accounting mechanism at sync-close>_
