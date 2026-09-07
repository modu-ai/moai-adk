# SPEC-CC-GD124-001 — Sync-Phase Verdict (card t491, lane-2)

Date: 2026-09-07 · Branch `WT-cc-upstream-sweep` · Worktree `.claude/worktrees/t491`

## Claim

SPEC-CC-GD124-001 (Claude Code upstream drift repair GD-1/2/3/4 — `context-window-management.md` + `cross-session-messaging.md`, each × local/template mirror) is closed: 4 files repaired, AC 9/9 PASS, `status: completed` on the single sync close commit, `sync_commit_sha` backfilled in the follow-up commit. Card scope is GD-1/2/3/4 only — GD-5/7/9 untouched (per SPEC § Out of Scope), GD-6 operator-held, GD-8 no-action.

## Evidence

- **What landed** — 4 target files: `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` (template) + local `.claude/rules/moai/workflow/context-window-management.md`; `cross-session-messaging.md` pair likewise. Run commits: `c67019383` (M1, cw pair + draft→in-progress), `79aa4f1c4` (M2, csm pair), `e710163c9` (M3, §E.2/§E.3). Close commit: `397db0209`; backfill commit: `d2be0362d` (follow-up).
- **AC 9/9 PASS** — attributed to (a) manager-develop §E.2 verbatim outputs (`.moai/specs/SPEC-CC-GD124-001/progress.md`, run at HEAD `79aa4f1c4`) and (b) the lane orchestrator's independent re-measure on 2026-09-07 (all AC greps reproduced; cw pair `diff -q` rc=0; csm pair exactly 1 hunk; `make build` rc=0; `go test ./internal/template/...` rc=0).
- **Absorb guidance (measured 2026-09-07)** — `origin/develop` = `ace1c5440`, 91 commits ahead of branch base `615d18c1f`, ZERO of them touching the 4 target paths. Expected conflict-free absorb. Expected delta at the integration window: 4 rule files + the SPEC directory (`.moai/specs/SPEC-CC-GD124-001/`) + the evidence directory (`.moai/reports/t491/`). Re-read the develop tip at window time regardless (the tip moves; this snapshot does not).

## Lead-report items (corrections to the sweep's verdicts)

1. **csm mirror parity is "exactly 1 known hunk", NOT rc=0.** The single `diff` hunk is the `> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).` blockquote, present in the local copy only — intentional template-neutrality divergence (SPEC IDs are forbidden template content per template isolation doctrine §25.1 class C1; introduced by `9ef2b91e1`, card t187). AC7 asserts exactly-1-hunk, not byte-identity, for this pair.
2. **GD-4 verified stronger than the sweep's partial verdict.** The sweep recorded a partial class-level note; this card verified the full class-level statement (2026-09-07): since v2.1.248 same-machine messaging works with feature-flag fetching off on every provider — the four named env flags remain flag-evaluation disables with no per-flag claims (AC6, REQ-008).

## CHANGELOG disposition

`grep -c 'SPEC-CC-GD124-001' CHANGELOG.md` → `0`. No entry emitted, per repo convention (since card t484, 2026-09-06): CHANGELOG entries are NOT emitted for card SPECs; history lives in each SPEC's HISTORY/Amendments sections. No evidence found that the convention changed.

## Baseline-attribution

All AC evidence: this worktree (`.claude/worktrees/t491`), run phase at `79aa4f1c4`/`e710163c9`, lane re-measure 2026-09-07, close at this commit. develop-tip measurement: `ace1c5440` via `git fetch origin develop` + `git rev-parse`, 2026-09-07.

## Gaps

- `go test ./...` (full suite) NOT run locally — lane-local scope discipline; verdict rests on `./internal/template/...` + develop CI at window time.
- develop-absorb "conflict-free" is a prediction from path-disjointness (0 of 91 commits touch the 4 paths), not an observed merge.
- Lint (E5) N/A — prose-only change, no Go source touched.

## Residual-risk

- `origin/develop` advances between now and the window; a future commit could touch the target paths (unlikely but the re-read-at-window-time rule covers it).
- The csm Origin-line hunk must be preserved as divergence — an over-eager future mirror (verbatim `cp`) would erase it and silently break the template-neutrality guard.
