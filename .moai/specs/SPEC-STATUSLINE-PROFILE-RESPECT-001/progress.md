---
id: SPEC-STATUSLINE-PROFILE-RESPECT-001
title: "Progress — statusline opt-out honored end-to-end + subtree profile resolution"
version: "0.2.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "progress"
lifecycle: spec-anchored
tags: "progress, statusline, forge, opt-out, launch-ledger"
tier: M
---

# Progress — SPEC-STATUSLINE-PROFILE-RESPECT-001

## §A. Current Phase

**plan-phase complete; kickoff gate PASSED (2026-08-27)** — decisions D1-D5
folded, artifacts at v0.2.0. Status: draft (the `draft → in-progress`
transition belongs to manager-develop on the first run-phase commit). Run
phase may begin.

## §B. Artifact Status

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/spec.md` | draft v0.2.0 (11 REQs; REQ-009 DEFERRED) |
| plan.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/plan.md` | draft v0.2.0 (M0-M4, M6, M7; M5 DEFERRED; 0 open markers) |
| acceptance.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/acceptance.md` | draft v0.2.0 (11 ACs; AC-009 DEFERRED; seam placement pinned §D) |
| progress.md | this file | §E skeleton emitted |

## §C. Milestone Tracker

| Milestone | Scope | Status | Evidence |
|-----------|-------|--------|----------|
| M1 | Profile subtree resolution, READ path (REQ-006/007/008) | pending | — |
| M2 | Segment-gated refresh spawn (REQ-001/003) | pending | — |
| M3 | Explicit-override spawn early-out (REQ-002) | pending | — |
| M4 | Display-honesty characterization (REQ-005) | pending | — |
| M5 | Write-side ledger normalization (REQ-009) | **DEFERRED** → follow-up card (kickoff D1) | — |
| M6 | Test seams + verification closure (REQ-010/011) | pending | — |
| M7 | Operational edit: `forge: none` in this repo's statusline.yaml (NOT code) | pending | — |

## §D. Blockers / Open Decisions

None open — kickoff gate passed 2026-08-27. Resolutions on record:

- D1/D2 — ancestor-walk READ-path only (`ResolveLaunchProfileForProject` miss
  path); write-side normalization (REQ-009/AC-009/M5) DEFERRED to follow-up
  card "launch-ledger write-side subtree normalization".
- D2(policy) — anonymous-session default-off REJECTED; explicit config only.
- New — operational `forge: none` edit accepted → plan.md M7 (§2.3
  update-wipe caveat noted inline).
- D3 — spawn-counter seam pinned: post-freshness, pre-`isSelfInvocable`
  (acceptance.md §D preamble).
- D4 — AC-005 fixture wording corrected (absent/corrupt cache ⇒ Available=false).
- D5 — REQ-004 scoped to the recognized-override branch (none→github).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex check: executed 2026-08-27, output `PASS` for
  `SPEC-STATUSLINE-PROFILE-RESPECT-001` (pattern
  `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`); no existing SPEC directory collision.
- Frontmatter: 12 canonical fields present in spec.md per
  `.claude/rules/moai/development/spec-frontmatter-schema.md` (tier: M added as
  optional field; status: draft — the only transition manager-spec performs).
- Out of Scope: 5 `### Out of Scope — <topic>` H3 sub-headings, each with `-`
  bullets (satisfies `OutOfScopeRule`).
- Requirements notation: GEARS (capability-gate `Where`, state-driven `While`,
  event-driven `When`, unwanted `shall not`, ubiquitous). No `IF/THEN`.
- Open clarification markers: 0 — the single marker (plan.md M5, decision D1)
  was resolved at the 2026-08-27 kickoff gate and replaced by a recorded
  verbatim decision digest; AC-009/REQ-009 carry the DEFERRED outcome for
  traceability.

## §E.2 Run-phase Evidence

Run phase executed 2026-08-27 in worktree `.claude/worktrees/t293`, branch
`WT-statusline-profile`, base `3abde7053`. Commits: `62485c918` (M1),
`5a193fa4c` (M2+M3), `d615bf374` (M4), plus the M7/§E.2 closure commit. All
evidence below is (this run, this tree) against the post-M4 HEAD unless noted.

### AC matrix

| AC | Status | Test | Evidence |
|----|--------|------|----------|
| AC-001 | PASS | `TestBuilder_SegmentGateSuppressesPairAndSpawn` | `go test ./internal/statusline/ ./internal/profile/ -count=1 -v -run '...'` → exit 0, 12 top-level PASS / 0 FAIL (`ac-matrix.txt`) |
| AC-002 | PASS | `TestMaybeRefreshGitHubCounts_ExplicitNoForgeSpawnsNothing` (+`TestMaybeRefreshGitHubCounts_UnsetOverrideStillSpawns`) | same run, exit 0 |
| AC-003 | PASS | `TestMaybeRefreshGitHubCounts_NoConfigSpawnsOnce`, `TestBuilder_NilSegmentsPreservesSpawn` | same run, exit 0 |
| AC-004 | PASS | `TestForgeOptOut_TwoWayRevert` (paired with AC-002 in same run) | same run, exit 0 |
| AC-005 | PASS | `TestRender_ForgePairFourStateContract` (6 subtests) | same run, exit 0 |
| AC-006 | PASS | `TestResolveLaunchProfileForProject_SubtreeWorktreeResolves` | same run, exit 0 |
| AC-007 | PASS | `TestResolveLaunchProfileForProject_DeepestRegisteredAncestorWins` | same run, exit 0 |
| AC-008 | PASS | `TestResolveLaunchProfileForProject_LexicalPrefixIsNotAnAncestor` | same run, exit 0 |
| AC-009 | DEFERRED (kickoff D1) | — | no run-phase test by design; follow-up card "launch-ledger write-side subtree normalization" |
| AC-010 | PASS | grep review | `grep -n 'exec.Command\|gh \|git remote'` over the 3 new test files → only comment-line hits; zero exec targets (`ac-matrix.txt` sibling greps) |
| AC-011 | PASS | grep review | new code adds no env-var reads; pre-existing refs use `config.EnvClaudeConfigDir` / `config.EnvNoProfileFallback`; paths via `filepath` APIs |

### TDD RED evidence (pre-GREEN, `--count=1`)

- Statusline gates: `go test ./internal/statusline/ -run 'ExplicitNoForgeSpawnsNothing|SegmentGateSuppressesPairAndSpawn|TwoWayRevert'` → exit 1;
  verbatim `spawn attempts = 1, want 0 — an explicit no-forge override must
  gate the refresh child` and `gated render spawn attempts = 1, want 0`
  (`red-statusline.txt`).
- Profile walk: `go test ./internal/profile/ -run 'Subtree|Deepest|Lexical|Stale|ExactMatchBeatsSubtree'` → exit 1;
  verbatim `subtree session resolved to "", want "alpha"` (`red-profile.txt`).
  RED for the right stated reason: gates/walk absent, not fixture breakage
  (AC-008 / miss-fallthrough / exact-match guards were green at RED and stayed
  green — the two-cell pair).

### Closure gates

- `go test ./internal/statusline/... ./internal/profile/... -count=1` → exit 0
  (`final-suite.txt`; both `ok`).
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`final-winbuild.txt`).
- `go vet ./internal/statusline/... ./internal/profile/... ./internal/cli/` → exit 0.
- `gofmt -l` over the 6 touched files → empty.
- Coverage (touched packages): statusline 90.6%; profile 84.3% package /
  `lookupSubtreeProjectKey` 88.9% and `ResolveLaunchProfileForProject` 100.0%
  per-function (`coverage.txt`, `profile-cover.out`).
- M7 operational read-back: `.moai/config/sections/statusline.yaml` line 12
  reads `forge: "none"`; `git status internal/template/` clean (no mirror).

Evidence dir: `.moai/state/verify/t293-dev/` — `m0-baseline.txt`,
`red-statusline.txt`, `red-profile.txt`, `green-m1m3.txt`, `green-m1m3-v2.txt`,
`m4-fourstate.txt`, `coverage.txt`, `profile-cover.out`, `final-suite.txt`,
`final-winbuild.txt`, `ac-matrix.txt`.

### Gaps

- Profile package-level coverage 84.3% is below the 85% bar; the shortfall is
  pre-existing uncovered functions (Delete, GetCurrentName paths) outside this
  SPEC's scope. The SPEC's own additions are covered at 88.9% / 100%.
- No integration run of the real statusline binary against the repo-local
  `forge: none` key (M7 is an operational edit; its code path is covered by
  unit tests against identical fixtures).


## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
