# SPEC-INBOX-DRAIN-GAP-001 — Progress

> Card t280 · Factory Mode lane-15 · Tier M · status: in-progress (run-phase M1 commit, 2026-09-02)

## §E.1 Plan-phase Audit-Ready Signal

_Pending plan-audit — plan-phase artifact set (spec.md / plan.md / acceptance.md / progress.md) created 2026-09-02 at worktree HEAD `131daa290`._

## §E.2 Run-phase Evidence

Run-phase evidence — manager-develop (lane-15), worktree `.claude/worktrees/t280`, branch `WT-inbox-drain-gap`, base `918bacd2c`. TDD discipline (RED→GREEN per milestone); mutant-teeth RED observations recorded per acceptance.md §A to `.moai/reports/t280/`.

### Milestone log

| Milestone | Scope | Tests | Evidence |
|---|---|---|---|
| M1 | Stub schema version field `v:1` + absence-tolerant reader (`InboxStubVersion`); cap constants pinned in `internal/config/defaults.go` (`DefaultInboxMaxBytes` 1 MiB, `DefaultInboxArchiveGenerations` 2) | RED build-failure observed (`m1-red.log`) → GREEN 7/7 PASS incl. pre-existing lessons-inbox baseline (`m1-green.log`) | `.moai/reports/t280/m1-red.log`, `.moai/reports/t280/m1-green.log` |

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
