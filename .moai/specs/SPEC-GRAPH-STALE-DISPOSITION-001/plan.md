---
id: SPEC-GRAPH-STALE-DISPOSITION-001
title: "Graph freshness gate stale-verdict disposition — implementation plan"
version: "0.1.0"
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
tier: S
---

# plan.md — SPEC-GRAPH-STALE-DISPOSITION-001

## §A Context

Card t493 asked whether `moai graph check`'s stale verdicts (mx-index `value=361`, edges `value=3`) indicate a defect. The lane's investigation concluded **correct reporting — no defect**: both artifacts are UNTRACKED runtime artifacts stamped 2026-08-27 (commit `d2cba5e21`) in the develop worktree, and the observed drift exactly equals the source evolution since those stamps. The mx-index threshold of 1 is a correct signal; the "threshold mismatch with the workflow" hypothesis is disproven (spec.md §1.4). This SPEC's run phase turns that completed investigation into tracked evidence: one disposition report, one regeneration demonstration, one mutant RED proof — in this card worktree only.

The verdict is already reached; the run phase is an evidence-formalization phase, not an investigation phase. No production code changes exist in scope.

## §B Known Issues

- **Artifacts absent in this worktree** (measured 2026-09-07): `.moai/state/mx-index.json` and `.moai/project/graph/edges.meta.json` exist only in the develop worktree. The card-worktree transition is absent→fresh; the mutant RED check carries the stale-detection proof. ACs are written against this reality.
- **Exit-code pipe artifact**: `$?` after `moai graph check | head` reflects `head`, not the check. Verdict strings are the judgment basis (spec.md REQ-008).
- **The regenerable artifacts are 27 MB+ and the mx scan walks the full inventory** — both commands are expected to take real time; run them with an adequate Bash timeout and never concurrently.
- **t478 lineage hazard**: a bare `edges.meta.json` re-stamp produces the cheapest green and is a false green (restamp-discriminator principle). The plan mandates `moai graph build` (full recompute) only.
- **Binary lag [coordinator-added]**: the installed `moai` (v3.2.0-rc.0, build `e79c010b8`, 2026-09-03) predates the two commits that landed the t478 restamp-discriminator predicate (`2649fe296`, `46f6a3236` — `internal/graph/check.go` + `internal/cli/graph_check.go`). A PATH-invoked build cannot judge this tree's freshness behavior — every graph operation MUST run via `./bin/moai` after `make build` (REQ-008 / AC-008; VCI §2.2).

## §C Pre-flight

Measured in this worktree on 2026-09-07 (plan-phase, this session):

| Item | State |
|------|-------|
| HEAD | `df74b3c9d` |
| Branch | `WT-graph-mxindex-edges` (card t493) |
| `.moai/state/mx-index.json` | absent |
| `.moai/project/graph/edges.meta.json` | absent |
| `.moai/reports/t493/` | absent (created by run phase) |

Pre-flight commands for the run phase (re-measure, do not trust this table):

```bash
git rev-parse --short HEAD && git branch --show-current
ls .moai/state/mx-index.json .moai/project/graph/edges.meta.json 2>&1
make build                        # tree-built binary — REQUIRED before any graph op (AC-008)
./bin/moai graph check 2>&1 | head -40   # baseline per AC-001; verdict strings, not $?
```

All subsequent graph invocations in M1 use `./bin/moai` by path, never the installed `moai`.

## §D Constraints

1. **Worktree isolation [HARD]**: all writes and executions happen in this card worktree only. The develop worktree is never written — not even to "helpfully" refresh its artifacts (REQ-001, AC-007).
2. **No production changes [HARD]**: no `internal/`/`pkg/`/`cmd/` edits, no `gate.yaml` edits, no threshold constants (REQ-002).
3. **Tracked write surface**: `.moai/reports/t493/verdict.md` + the SPEC artifacts under `.moai/specs/SPEC-GRAPH-STALE-DISPOSITION-001/` only. Regenerated runtime artifacts are UNTRACKED and stay uncommitted.
4. **Mutant hygiene [HARD]**: exactly one file mutated at a time, by reverse-applying a real committed change; restored with `git checkout -- <file>` immediately after the RED observation; `git status --porcelain` verified clean of mutants before the card completes (REQ-006, AC-005).
5. **Commit messages**: every commit on this branch carries the card id `t493` in the message (Conventional Commits, English), e.g. `feat(SPEC-GRAPH-STALE-DISPOSITION-001): M1 disposition report + regeneration demo (t493)`.
6. **Verdict-string basis**: layer verdicts are cited from output text; exit codes recorded only as annotated secondary observations (REQ-008).
7. **No messages to other sessions** — results return in the lane's final report; the lead recommendation travels inside the disposition report (REQ-007(e)).
8. **Tree-built binary provenance [HARD, coordinator-added]**: every cited graph measurement is produced by `./bin/moai` (built via `make build` from this tree at run-phase start); the installed `moai` from PATH is never the judging build, because it predates `2649fe296`/`46f6a3236` (REQ-008, AC-008).

## §E Self-Verification

Run-phase verification evidence maps 1:1 to acceptance.md-equivalent ACs (spec.md §3, Tier S inline):

- AC-001/002: `moai graph check` outputs before/after regeneration, recorded verbatim in the report.
- AC-003: the `moai graph build` output's recomputed edge count cited in the report.
- AC-004/005: mutant RED output, restore, and final `git status --porcelain` recorded in the report.
- AC-006: report completeness checklist (REQ-007's five elements).
- AC-007: `git diff --stat <run-start-HEAD>..HEAD` confined to SPEC artifacts + report; untracked inventory shows only the runtime artifacts and the report.
- AC-008: every evidence row names the `./bin/moai` invocation path; the report records that `make build` ran at run-phase start and no cited measurement used the installed PATH build.

Each AC row in the report carries: command, verbatim output, exit code (annotated), and the HEAD it was measured on.

## §F Milestones

Ordered by decision-reversibility: the report's framing (the only human-readable, review-heavy deliverable) leads; the mechanical demonstration and cleanup follow.

### M1 — Disposition report + regeneration demonstration + mutant RED proof (Priority High)

1. `make build` — produce the tree-built judging binary; all graph ops below invoke `./bin/moai` (AC-008).
2. Record the pre-intervention baseline (AC-001).
3. Run `./bin/moai mx scan`, then `./bin/moai graph build` (serial; real regeneration), then `./bin/moai graph check` — capture the fresh transition (AC-002, AC-003).
4. Mutant RED: reverse-apply one real committed change to one inventoried file, re-run `./bin/moai graph check`, record RED (AC-004); restore and re-verify fresh (AC-005).
5. Write `.moai/reports/t493/verdict.md` with all five REQ-007 elements and the AC evidence rows (each naming the `./bin/moai` invocation path and the HEAD measured).
6. Verify clean tree + no out-of-scope writes (AC-006, AC-007).

Exit: all eight ACs evidenced in the report; single milestone — no M2.

## §G Anti-Patterns

- **Bare meta re-stamp** instead of `moai graph build` — the false-green defect this card's lineage exists to prevent (t478 / SPEC-GRAPH-GATE-RESTAMP-001).
- **Reading `$?` after a pipe** as the check's verdict — the reported rc=0 anomaly that motivated this card.
- **Treating the develop-worktree figures (361/3) as card-worktree acceptance numbers** — they are investigation evidence; this worktree's baseline is its own measurement (§1.3).
- **Writing to the develop worktree to "fix" its stale artifacts** — lead-owned integration-window concern; doing so is an AC-007 failure.
- **Leaving a mutant un-reverted** — the card must not complete with any tracked file modified beyond SPEC artifacts + report.
- **Re-litigating the threshold** — the hypothesis is disproven with measurement (spec.md §1.4); reopening it requires new evidence, not this SPEC.
- **Judging this tree with the installed binary** — a PATH-invoked `moai` predating `2649fe296`/`46f6a3236` silently skips the restamp-discriminator predicate; every cited measurement must come from `./bin/moai` after `make build` (AC-008).

## §H Cross-References

- spec.md §1.2 — the lane's verified evidence chain (mx-index intersection = 361; edges source-set commit counts 6/836/780; exit-code semantics).
- spec.md §1.4 — the disproven threshold-mismatch hypothesis and its disproving measurement.
- `internal/cli/graph_check.go:121-131` — `Failed()` / `exitStaleOrAbsent` exit semantics.
- `internal/graph/meta.go` `SourceFingerprintsForEdges` — edges source sets: codemaps / mx-index / specs / reports.
- SPEC-GRAPH-GATE-RESTAMP-001 — restamp discriminator. SPEC-V3R6-GRAPH-FRESHNESS-001/002, SPEC-GRAPH-FRESHNESS-CADENCE-001 — cadence domain (out of scope here).
- Card t493 — dispatch, evidence path `.moai/reports/t493/`, branch `WT-graph-mxindex-edges`.
