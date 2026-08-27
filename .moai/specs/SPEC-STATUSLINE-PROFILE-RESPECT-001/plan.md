---
id: SPEC-STATUSLINE-PROFILE-RESPECT-001
title: "Plan — statusline forge/github opt-out honored end-to-end + subtree-aware profile resolution"
version: "0.2.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/statusline;internal/cli;internal/profile"
lifecycle: spec-anchored
tags: "plan, statusline, forge, opt-out, launch-ledger, worktree"
tier: M
---

# Plan — SPEC-STATUSLINE-PROFILE-RESPECT-001

## §A. Context

Fixes modu-ai/moai-adk#1675 from factory card t293 (branch `WT-statusline-profile`,
base origin/main 3abde7053). Two independent root causes, both verified in this
worktree (see spec.md §A.1):

1. **Spawn-gating gap** — `Builder.Build` (`internal/statusline/builder.go`)
   calls `maybeRefreshGitHubCounts(boardRoot)` without consulting the segment
   config, and `maybeRefreshGitHubCounts` consults only cache availability/age —
   never the explicit forge override. So both opt-out levers suppress rendering
   but not polling.
2. **Profile resolution gap** — `lookupProjectKey`
   (`internal/profile/profile.go:297`) is exact-path + `os.SameFile` only; a
   fresh worktree has no `launch.yaml` entry, so sessions fall back to an
   anonymous profile whose statusline config never carried the operator's
   opt-out. This is the "폴백 프로필" of the issue title.

Development mode: DDD (quality.yaml `constitution.development_mode`) —
characterization first (the renderForgePair four-state contract and the
exact-match resolver are existing behavior to PRESERVE), then IMPROVE.

## §B. Known Issues (inputs, not outputs of this SPEC)

- The local `.moai/config/sections/statusline.yaml` in this repo carries no
  `forge` key and no `github` segment key — tests must construct their own
  config fixtures; do not depend on repo-local config.
- Past worktrees are individually registered in the operator's `launch.yaml`;
  subtree matching must coexist with those entries without duplicate warnings.
- `moai update` wipes hand-edited files under `.moai/config` (separate card
  family) — any manual opt-out in a template-managed path can vanish on update.
  Overlap noted; not fixed here.

## §C. Pre-flight (run-phase M0)

1. Baseline: `go test ./internal/statusline/... ./internal/profile/...` —
   record pass/fail before any edit.
2. Confirm seams available: `forgeOverride` reads a real file path (inject via
   `boardRoot` fixture under `t.TempDir()`); `maybeRefreshGitHubCounts` spawn is
   guarded by `isSelfInvocable` (test binary is not `moai` → spawn is a no-op
   under `go test`). Verify this guard holds for the new gating tests.
3. Confirm `ResolveLaunchProfileForProject` signature and the ledger fixture
   pattern used by `internal/profile/profile_test.go` /
   `project_key_casing_test.go` — reuse their fixture style.

## §D. Constraints

- See spec.md §C (C1–C5). Additionally:
- **D1** — Segment gating must reach the Builder without a new Options field if
  one already exists that carries `SegmentConfig` (it does — `Options.SegmentConfig`
  feeds `isSegmentEnabled`); prefer reusing `isSegmentEnabled(SegmentGitHub)`
  over plumbing a new flag.
- **D2** — No behavior change for sessions in registered roots with no subtrees
  (pure additive matching in the miss path of `lookupProjectKey`).

## §E. Self-Verification (per-milestone)

- After each milestone: `go test ./internal/<pkg>/...` for the touched package
  only (per repo policy — no local `go test ./...`), plus
  `GOOS=windows GOARCH=amd64 go build ./...` before commit.
- Grep guard: new tests contain no `gh ` exec and no real network — verify by
  reviewing that every `exec.Command` in new test code targets fixtures.

## §F. Milestones

Ordered by decision-reversibility (most likely to change first). No time
estimates; priority labels only.

### M1 — Profile subtree resolution semantics (READ path) [Priority: High]

**Kickoff decision D1 (recorded)**: ancestor-walk is implemented inside
`ResolveLaunchProfileForProject`'s miss path — i.e. after `lookupProjectKey`
(exact match + `os.SameFile`) misses — NOT inside `lookupProjectKey` itself.
`lookupProjectKey` keeps its exact/alias semantics untouched (C5). On a miss,
walk the session directory's ancestors; for each ancestor that IS a ledger key
(respecting path-segment boundaries), prefer the deepest. Deliver REQ-006,
REQ-007, REQ-008. Characterization tests for the existing exact-match/alias
behavior come first (PRESERVE).

**Decision shape to review**: walk-up-from-cwd (O(depth) stat calls, bounded by
filesystem depth) vs. iterate-ledger-check-containment (O(entries)). Walk-up is
preferred: ledger size is unbounded, depth is not.

### M2 — Segment-gated refresh spawn [Priority: High]

User-visible behavior change (polling stops). Gate the
`maybeRefreshGitHubCounts` call in `Builder.Build` on
`isSegmentEnabled(SegmentGitHub)`; the Builder already holds the segment
config. Deliver REQ-001 (+ REQ-003 characterization: nil segments map ⇒ spawn
still happens — preserve current behavior).

**Decision shape to review**: gate at the call site (builder) vs. inside
`maybeRefreshGitHubCounts` (needs a new parameter). Call-site gating keeps the
function's signature honest — it cannot know segment config today.

### M3 — Explicit-override spawn early-out [Priority: High]

In `maybeRefreshGitHubCounts`, before the spawn decision, read the explicit
forge override (a config read already priced into `resolveGitHubCounts`'s
contract — same constant-cost class) and return without spawning when the
override names no forge. Deliver REQ-002. Note the subtlety: this must NOT
suppress spawning when the override is unset (auto-detect path stays with the
child, which owns the `git remote` cost).

### M4 — Display-honesty contract pinned [Priority: Medium]

No production change expected. Characterization tests pinning the four
`renderForgePair` states (fetched `7/3`, honest `0/0`, unknown `-/-`, gated/
suppressed `""`), including the stale-zero cache case (successful fetch wrote
zeros; later refreshes rate-limited ⇒ `0/0` remains honest). Deliver REQ-005
and the F5 verification the issue asked for.

### M5 — Write-side ledger normalization [Priority: — DEFERRED, no run-phase work]

**Kickoff decision D1 (recorded verbatim-digest)**: "ancestor-walk moves to
the READ path only … write-side ledger normalization (REQ-009/AC-009/M5
portion) becomes DEFERRED → follow-up card." Consequences, made structurally
coherent:

- No run-phase milestone exists for REQ-009 in this SPEC; it is retained in
  spec.md §B and acceptance.md AC-009 as **DEFERRED** for traceability only.
- Follow-up candidate card: "launch-ledger write-side subtree normalization"
  (key the `projects` entry on the registered root when a launch resolved via
  subtree match, stopping the per-worktree ledger growth already observed in
  the operator's `launch.yaml`).
- Because M1 no longer entails REQ-009, kickoff question D2-as-scoped
  ("does M1 drag the write side in?") is fixed by the same move: it does not.

### M6 — Test seams, docs touch-ups, full verification [Priority: Medium]

Mechanical closure: shared test fixtures (config-file writer, cache writer,
exec fake where a seam is introduced), `@MX` tag updates on touched anchors,
`GOOS=windows` build, per-package test runs. Deliver REQ-010, REQ-011.

### M7 — Operational config edit: `forge: none` in THIS repo [Priority: Medium — OPERATIONAL EDIT, NOT CODE]

Accepted kickoff deliverable (decision #3). As the final step of the run phase
(after M2/M3 land and pass), add `forge: none` under the existing `statusline:`
block of THIS repository's `.moai/config/sections/statusline.yaml` — a local
operational opt-out for factory-lane polling, exercising the fixed code path
on our own tree. Verified by reading the key back (acceptance.md §D.3).

**§2.3 caveat (inline, load-bearing)**: this file lives under the
template-managed `.moai/config` root, which `moai update`
(`CleanMoaiManagedPaths`) wipes and re-deploys — a hand-added key silently
vanishes on the next update. The edit is operational and knowingly ephemeral;
it is NOT a fix for the wipe family (separate cards own that). Do not mirror
it into `internal/template/templates/`.

### Explicitly NOT in the milestone list

Template wipe-on-update redesign, `__no_such_profile__` chase, Forge rename,
anonymous-session default-off (REJECTED at kickoff — decision #2, explicit
config only), write-side ledger normalization (DEFERRED — M5 above).

## §G. Risks / Anti-Patterns

- **RISK (overlap, not owned)**: `moai update` deletes hand-added opt-out keys
  under `.moai/config` — even a correct code fix appears to "regress" after an
  update if the operator re-adds keys in a template-managed path. Mention in
  the sync-phase docs note; fix belongs to the wipe-family cards.
- **RISK (merge surface)**: sibling cards t215/t211 (PR #1621) touched
  statusline files; keep hunks minimal and named (C4) so conflicts review
  cleanly.
- **ANTI-PATTERN**: do not "fix" the anonymous profile by special-casing the
  literal `__no_such_profile__` string anywhere (F4 gap; the fix is resolution,
  not string matching).
- **ANTI-PATTERN**: do not move `git remote` / PATH lookups onto the render
  path to decide suppression — that breaks C2 (constant-cost render).
- **ANTI-PATTERN**: do not make `Suppressed` a latch that outlives its
  evidence; the existing not-a-latch comment contract must survive REQ-002/004.

## §H. Cross-References

- spec.md §B REQ-001..REQ-011; acceptance.md AC-001..AC-011
- Kickoff-gate decisions (all RESOLVED 2026-08-27): D1 read-path-only
  ancestor-walk + REQ-009 DEFERRED (M1, M5); D2(policy) anonymous default-off
  REJECTED (explicit config only); D3 spawn-counter seam placement pinned
  (acceptance.md §D preamble); D4 AC-005 fixture wording; D5 REQ-004 scoped
  to the recognized-override branch; operational `forge: none` edit → M7.
