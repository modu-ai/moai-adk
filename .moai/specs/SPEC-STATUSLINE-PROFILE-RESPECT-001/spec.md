---
id: SPEC-STATUSLINE-PROFILE-RESPECT-001
title: "Statusline forge/github opt-out honored end-to-end; worktree sessions resolve the enclosing project's launch profile"
version: "0.2.0"
status: implemented
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/statusline;internal/cli;internal/profile"
lifecycle: spec-anchored
tags: "statusline, forge, github, opt-out, poll-suppression, launch-ledger, profile, worktree"
tier: M
issue_number: 1675
---

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-08-27 | 0.1.0 | manager-spec | Initial plan-phase authoring (factory card t293, branch WT-statusline-profile, base origin/main 3abde7053). Fixes modu-ai/moai-adk#1675. |
| 2026-08-27 | 0.2.0 | manager-spec | Kickoff-gate decisions folded: D1/D2 ancestor-walk pinned to the READ path (`ResolveLaunchProfileForProject` miss path), write-side normalization (REQ-009) DEFERRED to follow-up card; D2(policy) anonymous default-off REJECTED; D3 spawn-counter seam placement pinned (acceptance.md §D); D4 AC-005 fixture wording corrected; D5 REQ-004 scoped to the recognized-override branch; operational `forge: none` edit added to plan.md M7. |

## §A. Overview

GitHub issue #1675 reports that after the operator opted out of GitHub counts in
the statusline (an order dated 2026-08-17, recorded in the operator's other
project), the fallback profile kept showing counts / polling anyway — "폴백
프로필이 설정을 무시". Investigation in this worktree established the actual
mechanics; the premise "config is ignored" is the symptom of two independent
code-level gaps, not a config-read override.

### §A.1 Verified findings (ground truth for this SPEC)

- **F1 — the opt-out keys are simply absent where the renderer reads them.** The
  operator's 2026-08-17 opt-out (`segments.github: false` + `forge: none`) lives
  in the *other* project's `statusline.yaml`. The file the renderer reads for a
  session inside this repository (`<root>/.moai/config/sections/statusline.yaml`,
  33 lines) carries neither key, so no opt-out exists to ignore. Any fix must
  make the opt-out keys work when present — it cannot invent them where absent.
- **F2 — render path.** `internal/cli/statusline.go` (`runStatusline`) resolves
  the project root by walking up for a `.moai` dir, then
  `loadStatuslineFileConfig` reads ONLY `{theme, segments}`. An absent file
  yields `nil`, and the Builder then falls back to all-segments-enabled.
  `statusline.forge` is read separately by `forgeOverride`
  (`internal/statusline/forge.go`) from the same file.
- **F3 — suppression already exists but is not spawn-gating.**
  `GitHubCounts.Suppressed` is set when the explicit override names no forge,
  the remote host is unrecognized, or the forge CLI is missing
  (`internal/statusline/github.go`, `forge.go`). However
  `Builder.Build` (`internal/statusline/builder.go`) calls
  `maybeRefreshGitHubCounts(boardRoot)` unconditionally — it never consults the
  segment config, and `maybeRefreshGitHubCounts` consults only cache
  availability/age. Consequences: (a) `segments.github: false` hides the pair
  but still spawns refresh children; (b) an explicit `forge: none` with an
  absent/stale cache still spawns one child per TTL, which re-derives
  suppression it could have read from the config.
- **F4 — profile resolution is exact-path only.**
  `~/.moai/claude-profiles/launch.yaml` `projects[]` registers exact paths
  (it even lists individual past worktrees). `lookupProjectKey`
  (`internal/profile/profile.go:297`) does exact-match plus `os.SameFile` alias
  detection — NO ancestor walking. A fresh worktree
  (`.claude/worktrees/<new-id>`) has no entry, so
  `ResolveLaunchProfileForProject` returns `""` and the session launches on an
  anonymous/fallback config dir even though the enclosing project IS registered.
  This is the "폴백 프로필" in the issue title: the fallback profile is the one
  whose statusline config never carried the opt-out.
- **F5 — display honesty.** `renderForgePair` (`renderer.go`) renders four
  states: `7/3` fetched (including honest `0/0` when a fetch genuinely returned
  zeros), `-/-` unknown (cache unavailable), and `""` for both gated-off
  (segment disabled) and suppressed (no forge will answer). The issue reported
  seeing `0/0` during failed polling; per the code, `0/0` can only render from a
  cache written by a successful fetch (a stale cache legitimately keeps zeros
  after later rate-limited refreshes). This SPEC pins that contract in an AC
  rather than changing it.

### §A.2 What this SPEC builds

1. **Opt-out honored end-to-end (render + polling).** Both opt-out levers
   (`segments.github: false` and `statusline.forge: none`) stop not only the
   rendering but also the detached refresh child. Multiple concurrent factory
   lanes each polling `gh` every 10 minutes was the original rate-limit (429)
   motivation of the 2026-08-17 order.
2. **Two-way revert.** Removing the opt-out brings the pair back within one
   refresh cycle — enforced by paired regression ACs, per the lead's mandate.
3. **Subtree-aware profile resolution (read path only).** A session whose
   project directory lies anywhere inside a registered project subtree —
   including a fresh, unlisted worktree — resolves to that project's named
   profile at read time. Per kickoff decision D1, the ancestor-walk lives in
   `ResolveLaunchProfileForProject`'s miss path, NOT inside `lookupProjectKey`;
   the write-side ledger normalization that would let one registration cover
   future subtrees on the ledger too is DEFERRED to a follow-up card
   (REQ-009).
4. **Test seams, not networks.** All new tests exercise forge resolution,
   spawn-gating, and profile resolution through fakes; no real `gh` /
   `git remote` calls.

## §B. Requirements (GEARS)

### §B.1 Opt-out gating (statusline)

- **REQ-001** (capability gate, spawn): **Where** the project's
  `statusline.yaml` sets `segments.github: false`, the statusline Builder
  shall neither render the forge pair nor invoke the detached GitHub refresh
  child.
- **REQ-002** (capability gate, render path): **Where** the project's
  `statusline.yaml` sets `statusline.forge` to a value naming no forge
  (`none`, `off`, or an unrecognized string), the render path shall resolve the
  counts as Suppressed from configuration alone and shall not spawn a refresh
  child, regardless of cache presence or age.
- **REQ-003** (state-driven, fallback preservation): **While** the project's
  `statusline.yaml` carries no `segments` map, the Builder shall keep the
  all-segments-enabled fallback behavior (characterization — no regression).
- **REQ-004** (event-detected, two-way revert): **When** the `forge` override
  changes from a no-forge value to a recognized forge value (e.g. `none` →
  `github` — the branch AC-004 exercises; kickoff decision D5 scoped the
  requirement to this tested branch), the statusline shall resume rendering
  the pair within one refresh cycle without manual cache deletion.
- **REQ-005** (unwanted, display honesty): The renderer shall not render `-/-`
  for a suppressed or segment-gated checkout, and shall not render `0/0` unless
  a successful forge fetch produced zero counts; an unavailable cache with an
  expected forge shall render `-/-` (unknown, not zero).

### §B.2 Subtree-aware profile resolution

- **REQ-006** (event-driven): **When** a session's project directory lies
  within the subtree of a project registered in the launch ledger — including a
  worktree path not individually registered — profile resolution shall return
  that registered project's profile.
- **REQ-007** (state-driven, depth precedence): **While** two or more
  registered projects are nested within one another, the resolver shall prefer
  the deepest registered ancestor of the session's directory.
- **REQ-008** (unwanted): Profile resolution shall not return a registered
  project's profile for a directory that is merely a lexical prefix match but
  not a true descendant (e.g. `/proj-foo` vs `/proj`; path-segment boundary
  must be respected).
- **REQ-009** (write-side normalization — **DEFERRED**, kickoff decision D1):
  **When** a launch resolves through subtree matching, the ledger write side
  shall record the registered project root as the key rather than the
  transient subtree path, so one registration covers future subtrees without
  ledger growth. **Deferred to a follow-up card** — this SPEC ships read-path
  resolution only (AC-009 stays in the matrix as DEFERRED for traceability;
  follow-up candidate: "launch-ledger write-side subtree normalization").

### §B.3 Test discipline

- **REQ-010** (ubiquitous): The test suite added by this SPEC shall exercise
  forge resolution, spawn-gating, and profile resolution through injected seams
  (exec command factory and cache-file fixtures under `t.TempDir()`), and shall
  not invoke a real `gh` client or a real `git remote` network call.
- **REQ-011** (ubiquitous, conventions): New code in `internal/cli` and
  `internal/profile` shall follow the existing package conventions, including
  env-var references only via `internal/config/envkeys.go` constants and
  cross-platform path handling (`filepath` APIs).

## §C. Constraints

- **C1** — No new YAML keys beyond the two existing opt-out levers
  (`segments.github`, `statusline.forge`). No new config surface.
- **C2** — The render path must remain constant-cost (single small-file reads;
  the `git remote` / PATH lookups stay on the refresh-child side of the line,
  per the existing comment contract in `resolveGitHubCounts`).
- **C3** — The refresh child's timestamp-before-fetch and
  suppression-verdict-through-write semantics are preserved (staleness
  degradation, not blank segments).
- **C4** — Change surface is narrowly named: `internal/statusline/github.go`,
  `internal/statusline/builder.go`, `internal/statusline/forge.go` (if
  needed), `internal/cli/statusline.go` (only if segment gating must surface
  the flag into the Builder), `internal/profile/profile.go` + tests. Sibling
  cards t215/t211 (PR #1621) touched statusline files earlier; keep the diff
  easy to review against that baseline.
- **C5** — Profile subtree matching must not break the existing case-variant /
  `os.SameFile` alias behavior or the duplicate-entry warning.

## §E. Out of Scope

### Out of Scope — template wipe-on-update family (CleanMoaiManagedPaths)

- Redesigning `moai update`'s deletion of local-only files under
  `.moai/config/**` (the reason a hand-added opt-out key can silently vanish on
  update). Separate cards own that family; noted as an overlapping risk in
  plan.md §G only.
- Adding any protection list / skip-marker mechanism to
  `internal/cli/update/deploy/deploy.go`.

### Out of Scope — the `__no_such_profile__` directory-name mystery

- Chasing the origin of the literal anonymous config-dir name
  `~/.moai/claude-profiles/__no_such_profile__`. It is environment-inherited
  and outside this repository; documented as an accepted Gap in §F.

### Out of Scope — forge-surface growth

- Renaming `GitHubCounts` / `RefreshGitHubCounts` to `Forge*` (existing
  `@MX:DEBT` with a named upgrade trigger — a third forge or next cache
  versioning).
- Adding a third forge, or changing `GitHubCountsTTL` / fetch budgets.

### Out of Scope — implicit default-off behavior

- Turning the github pair off by default for anonymous/unregistered sessions
  absent explicit operator config. Whether to add that is an open kickoff
  decision (see plan.md §F, D2); this SPEC ships no implicit default change.
- Auto-writing opt-out keys into any project's `statusline.yaml`
  (migration tooling, wizard changes, template default flips).

### Out of Scope — statusline config surface changes

- Any new segment keys, theme changes, or web-console statusline panel work.

## §F. Gaps and Accepted Risks (accepted at plan-phase)

- **Gap — `__no_such_profile__` provenance.** Live factory lanes today run with
  `CLAUDE_CONFIG_DIR=~/.moai/claude-profiles/__no_such_profile__`. The literal
  name's producer is outside this repository. The code-level fix (subtree
  resolution) removes the dependence on that mystery directory, so its origin
  is accepted as unexplained. Not verified: whether any other behavior keys on
  that exact name.
- **Gap — operator's other-project config.** The 2026-08-17 opt-out in
  `/Users/goos/MoAI/mo.ai.kr/.moai/config/sections/statusline.yaml` was read
  during investigation but is NOT modified by this SPEC (different repository,
  operator-owned).
- **Residual risk — worktree registration drift.** Past worktrees listed
  individually in `launch.yaml` remain valid entries; subtree matching must
  not warn about them as duplicates. Covered by test, but ledger shapes in the
  wild were sampled, not exhaustively enumerated.

## §G. Cross-References

- Issue: modu-ai/moai-adk#1675
- `internal/statusline/github.go` — `GitHubCounts`, `resolveGitHubCounts`,
  `maybeRefreshGitHubCounts`, `RefreshGitHubCounts`
- `internal/statusline/forge.go` — `forgeOverride`, `resolveForge`
- `internal/statusline/renderer.go` — `renderForgePair` (four-state contract)
- `internal/statusline/builder.go` — unconditional
  `maybeRefreshGitHubCounts` call site (the REQ-001/REQ-002 gap)
- `internal/cli/statusline.go` — `loadStatuslineFileConfig` ({theme, segments})
- `internal/profile/profile.go` — `lookupProjectKey`,
  `ResolveLaunchProfileForProject`
- Related: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (segment map as the only
  lever), SPEC-V3R3-STATUSLINE-FALLBACK-001 (render fallback lineage)
