# progress.md — SPEC-CONFIG-MODE-MIGRATE-001

> Plan-phase skeleton. The §E.* headings below are the canonical markers the
> `era.go` classifier greps for (literal `§E.2` / `§E.3` / `§E.4` substrings).
> Do NOT author §E.5 (retired Mx-phase marker). Only §E.1 is populated at
> plan-phase; §E.2-§E.4 are placeholder headings left for the run/sync phases.

## §E.1 Plan-phase Audit-Ready Signal

- `plan_status`: _pending plan-auditor_
- `plan_complete_at`: _pending_
- `plan_artifact_count`: 2 (Tier S — spec.md + plan.md; this progress.md is emitted
  at every Tier and is not counted in the Tier total)
- `tier`: S
- `spec_lint_result`: _pending (run before plan-phase commit)_

## §E.2 Run-phase Evidence

_run-phase evidence placeholder — manager-develop populates this section at run-phase.
The literal `§E.2` heading above is the run-evidence START marker the era classifier
detects; do NOT rename or remove it._

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase — `sync_commit_sha:` field is populated by the single sync commit
at sync-phase close (manager-docs)._
