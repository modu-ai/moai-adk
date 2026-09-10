---
id: SPEC-STATUSLINE-PROFILE-RESPECT-001
title: "Acceptance — statusline opt-out honored end-to-end + subtree profile resolution"
version: "0.2.0"
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/statusline;internal/cli;internal/profile"
lifecycle: spec-anchored
tags: "acceptance, statusline, forge, opt-out, launch-ledger"
tier: M
---

# Acceptance — SPEC-STATUSLINE-PROFILE-RESPECT-001

## §D. AC Matrix

All scenarios run under `go test` with fixtures in `t.TempDir()`; every AC is
binary-testable via test output. "Config fixture" = a
`.moai/config/sections/statusline.yaml` written by the test. "Cache fixture" =
`.moai/state/github/counts.json` written by the test. "Spawn seam" = the
`isSelfInvocable` guard plus a spawn-counter seam; assertions on "no child
spawns" are made against that counter, never against a real process.

**Spawn-counter seam placement (kickoff decision D3, pinned).** The counter
sits in `maybeRefreshGitHubCounts` (`internal/statusline/github.go`, the
~113-116 region) AFTER the TTL/freshness check (and, once M3 lands, after the
explicit-override check) and BEFORE `isSelfInvocable`. Two-cell adoption
discipline this placement enables:

- **RED under current code, under `go test`**: today the function has no
  segment/override gating, so with a stale-or-absent cache the flow reaches
  the counter on every render — AC-001/AC-002 assertions of "zero spawn
  attempts" fail before M2/M3, proving the tests bite.
- **GREEN after M2/M3**: the new gates return before the counter, so the same
  assertions pass — the delta is the gate, nothing else.

Why entry-placement fails to discriminate: a counter at function entry counts
*calls*, not *spawn attempts* — the TTL-freshness early return fires below it,
so "call reached, no spawn intended" and "spawn attempted, blocked by
`isSelfInvocable`" produce identical entry counts. Under `go test` the
`isSelfInvocable` guard always blocks the real exec (test binary), so entry
placement cannot distinguish pre-spawn gating from spawn-blocking — exactly
the distinction REQ-001/REQ-002 must observe. Post-freshness placement counts
precisely "would have spawned".

### Opt-out gating (REQ-001..REQ-005)

- **AC-001** [REQ-001, MUST] Given a config fixture with
  `segments: { github: false }` and a stale-or-absent cache, When the Builder
  renders, Then the output contains no forge pair AND the spawn counter records
  zero refresh-child invocations.
- **AC-002** [REQ-002, MUST] Given a config fixture with `forge: none` (and
  separately, in a second case, `forge: off`) and an absent cache, When the
  Builder renders, Then the resolved counts carry `Suppressed: true` from the
  config read alone AND the spawn counter records zero invocations.
- **AC-003** [REQ-003, MUST] Given a project root whose statusline.yaml is
  absent (or carries no `segments` map) and a stale cache, When the Builder
  renders, Then the spawn counter records one invocation (all-enabled fallback
  preserved — characterization).
- **AC-004** [REQ-004, MUST] Given the AC-002 state (suppressed via
  `forge: none`), When the fixture is rewritten to `forge: github` with a
  fake forge resolution succeeding, Then within one refresh cycle the pair
  renders with fetched values AND no cache deletion was performed. (Paired
  two-way regression with AC-002, mandated by the lead.)
- **AC-005** [REQ-005, MUST] Given an absent or corrupt cache file (yields
  `Available=false` — the field is `json:"-"` and derived from a successful
  unmarshal, `github.go:41,86`, kickoff decision D4) and no suppression, When
  the pair renders, Then the output is exactly `-/-` (unknown). Given a cache
  fixture holding zeros written by a simulated successful fetch, Then the
  output is exactly `0/0` (honest zeros). Given `Suppressed: true` or
  `segments.github: false`, Then no pair renders at all.

### Subtree profile resolution (REQ-006..REQ-009)

- **AC-006** [REQ-006, MUST] Given a ledger fixture registering
  `/proj` → profile `alpha` and a session directory `/proj/.claude/worktrees/t999`
  (no entry of its own), When `ResolveLaunchProfileForProject` resolves for
  that directory, Then it returns `alpha`.
- **AC-007** [REQ-007, MUST] Given a ledger registering both `/proj` → `alpha`
  and `/proj/sub` → `beta`, When resolving for `/proj/sub/inner`, Then the
  result is `beta` (deepest registered ancestor wins).
- **AC-008** [REQ-008, MUST] Given a ledger registering `/proj` → `alpha` and a
  session directory `/proj-other` (shared prefix, different path segment), When
  resolving, Then no subtree match occurs (exact-match semantics alone apply).
- **AC-009** [REQ-009, **DEFERRED** — kickoff decision D1] Given a launch that
  resolved via subtree match, When the ledger is written, Then the `projects`
  key recorded is the registered project root, not the session's subtree path.
  **DEFERRED with REQ-009 to a follow-up card** (candidate: "launch-ledger
  write-side subtree normalization"); retained here for REQ↔AC traceability
  only — no run-phase test in this SPEC.

### Test discipline (REQ-010..REQ-011)

- **AC-010** [REQ-010, MUST] Given the full new test set, When inspected and
  run, Then zero tests invoke a real `gh` binary or a real `git remote`
  network call (exec targets are fakes/fixtures only), verified by test-code
  review plus the absence of network-dependent skips.
- **AC-011** [REQ-011, MUST] Given the diff, When reviewed, Then all new
  env-var references use `internal/config/envkeys.go` constants and all paths
  use `filepath` APIs (grep-verifiable).

## §D.1 Severity and Traceability

All ACs are MUST except where marked. Traceability: AC-001→REQ-001,
AC-002→REQ-002, AC-003→REQ-003, AC-004→REQ-004, AC-005→REQ-005,
AC-006→REQ-006, AC-007→REQ-007, AC-008→REQ-008, AC-009→REQ-009 (DEFERRED,
same-named follow-up candidate), AC-010→REQ-010, AC-011→REQ-011. The 11:11
REQ↔AC mapping is preserved; no 12th AC was added for the kickoff decisions.

## §D.2 Edge Cases

- Explicit override set but forge CLI missing → child-owned suppression path
  unchanged (REQ-002 only claims the config-read case).
- Override value that is a typo (e.g. `forge: gitbub`) → treated as
  suppress-or-spawn exactly as today (unrecognized value names no forge).
- Ledger entry pointing to a non-existent directory (stale worktree path) →
  subtree walk must tolerate `os.Stat` failure without matching it.
- Windows path separators — subtree boundary checks use `filepath.Separator`.

## §D.3 Closure Gates (Definition of Done)

- `go test ./internal/statusline/... ./internal/profile/...` green (this run,
  this tree, output cited).
- `GOOS=windows GOARCH=amd64 go build ./...` succeeds.
- Two-way pair (AC-002 + AC-004) both green in the same run.
- No new YAML keys; no production change to `renderForgePair` (M4 expected
  test-only — if a change proves necessary, it returns to plan review).
- **Operational key read-back (M7, kickoff decision #3)**: this repo's
  `.moai/config/sections/statusline.yaml` reads back `forge: none` under the
  `statusline:` block after the run phase (one `grep`/`yaml` read citing the
  key), with the §2.3 update-wipe caveat acknowledged (plan.md M7).

## §D.4 Forward-Looking Checks

- The wipe-on-update risk (spec.md §G / plan.md §G RISK) is surfaced in the
  sync-phase docs note so the operator knows hand-added keys in
  template-managed paths can vanish on `moai update` — separate card family.
