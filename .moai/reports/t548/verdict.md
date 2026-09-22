# Card t548 — Verdict: DELIVERED (pending lead merge-window + remote landing)

SPEC: SPEC-CODEX-DISABLE-EXIT-001 · Tier S · Class C · lane-9 · 2026-09-08

## Claim

Card t548 delivered per the card's three [HARD] rules: (1) the boundary-intent adjudication was performed BEFORE any code change — operator decision Outcome B (Skipped guard-refusal exits non-zero; Unchanged stays 0; absent-inputs stay 0), recorded in `progress.md` §E.1 and gated by plan-audit PASS 0.938; (2) Unchanged and Skipped are distinct outcomes enforced by per-branch tests (AC-CDE-002, RED observed on the pre-change tree); (3) the scope axis was decided first (verb-local, spec.md §D) with the `moai clean --codex-skills` distinction documented, so no cross-verb inconsistency was created. The M2 adjudication recorded in §E.1 precedes the first implementation commit (AC-CDE-007).

## Evidence

- **Caller census (card's mandatory first investigation)**: zero production in-repo callers — 7 search commands with outputs in `progress.md` §E.1; the card's literal form `moai codex skills disable` was design-rejected (SPEC-CODEX-SKILL-DISABLE-001/plan.md:74) and the real surface is `moai skills disable <name> --codex`.
- **Plan audit**: PASS 0.938 (Tier S bar 0.75), `.moai/reports/t548/plan-audit.md`, re-stamped at `d1742de07` after the D1 maps-clause fix; D2 routed as the binding RED-before-GREEN execution order.
- **RED-first (D2 held)**: `TestRunCodexSkillDisableSkippedExitsNonZero` observed FAIL on the pre-change tree `c6f6193e4` — verbatim evidence `.moai/state/verify/t548/red-skipped-nonzero.txt`; implementation sources byte-identical to baseline `a4855f0b2` in that window.
- **Lane re-verification (this run)**: `go test ./internal/cli/ -run 'TestRunCodexSkillDisable' -count=1` → `ok` (GREEN re-observed independently); `git diff --stat a4855f0b2 -- internal/cli/codex_skills_prune.go` → empty (AC-CDE-005); `moai spec lint .moai/specs/SPEC-CODEX-DISABLE-EXIT-001/spec.md` → `✓ No findings`; `go run ./cmd/moai skills disable --help` → per-class exit-code table present (AC-CDE-006); `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (dev-attributed E2).
- **AC matrix**: 7/7 PASS (`progress.md` §E.2, attribution triples per A9).
- **Commits on `WT-codex-disable-exit`** (baseline `a4855f0b2`): `5e7907746` plan artifacts · `ffc7a5ef9` M2 decision record · `d1742de07` D1 fix · `c6f6193e4` audit report (re-stamped) · `554f93e35` RED-first test · `c7a9e0830` M3 implementation · `5154bf2b8` M5 help docs · `6f3433a32` sync 3-phase close · `7ba44589c` SHA backfill. spec.md `status: completed`; `sync_commit_sha: 6f3433a32`.

## Baseline-attribution

All measurements in this report were taken in this run, in this worktree (`.claude/worktrees/t548`, branch `WT-codex-disable-exit`), against baseline `a4855f0b2`, by lane-9 or its attributed subagents with the command and output recorded in `progress.md` §E.1–§E.4.

## Gaps

- **E3 coverage**: `internal/cli` package coverage is 81.4% — below the 85% package target at a PRE-EXISTING level (untouched helpers); the touched runner is 95.0% covered with all new lines covered. Not a regression of this card; the package-level gap is not owned here.
- **docs-site**: the verb remains undocumented there (census finding) — separate work, out of this card's scope.
- **Integration verdict**: local verification is scoped to `internal/cli` per the lane-local rule; the full-suite verdict is `origin/develop` CI after the lead's batch push — pending by construction until the merge window.

## Residual-risk

- Out-of-repo scripts treating a Skipped refusal as success would break on upgrade — the census cannot see outside this repo; mitigated by the CHANGELOG entry and the `--help` contract statement.
- gopls reports spurious `undefined` diagnostics on this worktree (workspace-scoping noise over sibling-file symbols); real builds and tests pass — do not chase.
