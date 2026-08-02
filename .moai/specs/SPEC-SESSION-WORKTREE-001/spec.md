---
id: SPEC-SESSION-WORKTREE-001
title: "Automatic worktree isolation for moai init / moai profile / moai web, with worktree-scoped git config and PR-merge auto-cleanup"
version: "0.2.2"
status: in-progress
created: 2026-08-03
updated: 2026-08-03
author: manager-spec
priority: P1
phase: "v3.0.3 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "worktree,parallel-sessions,init,profile,web,isolation,git-config,pr-merge-cleanup"
tier: L
---

## §A — History

- **2026-08-03** — plan-phase v0.1.0 authored. Source: user feature request ("moai init, moai web에서 병렬 세션 작업시 충돌이 없도록 worktree 기본으로 사용 여부를 off로 기본으로 하고 on하면 항상 [WT] ... 같이 wt 사용해서 충돌없이 개발이 병렬로 진행이 되도록 기능을 추가하자."). Four design forks resolved (§2.2). Tier M (multi-file, internal/cli/, no constitutional change).
- **2026-08-03** — plan-phase v0.2.0 amendment. User requested three additions: (1) extend trigger surface to `moai profile`; (2) full git-config tier (safe.directory + profile identity + init.defaultBranch + opt-in options); (3) PR-merge-triggered auto-cleanup of `[WT]` worktrees. REQ/AC budget grew from 15/15 to 24/24, exceeding the Tier M ceiling (16/16) — **tier escalated M → L**, design.md + research.md added. Two new design forks resolved: profile-isolation scope (§2.9) and PR-merge detection mechanism (§2.11).
- **2026-08-03** — plan-phase v0.2.1 iter-2 audit-fix. iter-1 plan-audit FAIL 0.69. All forks now decided (no more blocker rounds): D1 — REQ-SW-019 identity source corrected from the (non-existent) profile `User.Email` field to the user's global git config (`~/.gitconfig`, the live source for BOTH `user.name` AND `user.email`); D2 — all four `[NEEDS CLARIFICATION]` markers resolved (Q1/Q2 documented M1 deferrals, Q3 on-touch trigger, Q4 drop `core.hooksPath`); D3 — research.md §C.3 documents the existing `defaultBranchDetectorFunc`/`detectBranch` in `internal/workflow/worktree_orchestrator.go` (REQ-SW-020 is complementary, not redundant); D4 — AC-SW-022 concrete on-touch trigger at `moai worktree list` / `moai session start`; D5 — path-distinguishing stderr notices folded into REQ-SW-009/022 bodies; D6 — REQ-SW-014 relabelled Event-driven; D7 — REQ-SW-021 SHALL (not SHOULD). Tier L retained (file surface still ≥10, REQ/AC still 24/24). _(Historical record — the v0.2.1 D4 trigger points `moai worktree list` / `moai session start` were discovered at v0.2.2 M8 pre-flight to be non-existent in the codebase; see the v0.2.2 entry immediately below for the correction.)_
- **2026-08-03** — run-phase v0.2.2 inline-fix (D-NEW-1 non-transition frontmatter correction). M8 pre-flight discovered the v0.2.1 Q3 trigger points named two commands that do NOT exist in the codebase: `moai worktree list` is INTENTIONALLY RETIRED (`internal/cli/worktree/root_test.go:61` `TestWorktreeCmd_RetiredSubcommands` pins `list` as retired; `worktree --help` directs users to `git worktree list`), and `moai session start` NEVER EXISTED (`internal/cli/session.go:53-59` registers exactly `register, heartbeat, deregister, list, purge, current, doctor` — no `start` command or alias). The corrected trigger pair is **`moai session register` + `moai session list`** (both verified existing: `session.go:66` `register`, `session.go:132` `list`). `session register` is the genuine session-start write path (the SessionStart hook calls it); `session list` is the natural user-inspection on-touch. CLI-level wiring preserves separation of concerns (cleanup logic stays out of the SessionStart hook). design.md §C.3's "reuses existing invocation points" claim becomes TRUE with this pair. Correction applied consistently across spec.md HISTORY, plan.md (§A / §F M8 / §I Q3), acceptance.md (AC-SW-022 / EC-13), and design.md (§C.3 / §D). REQ/AC count unchanged at 24/24; Tier stays L; status stays `in-progress` (non-transition inline-fix, NOT an amendment — no `amendment_of`, no `## Amendments` section). Q3 rationale ("on-touch, fires when user present, stale-until-next-invocation acceptable") PRESERVED — only the command names change.

## §B — Problem

`moai init`, `moai profile`, and `moai web` currently operate inside the **shared primary checkout**. Sessions invoking these subcommands in parallel collide on shared mutable surfaces:

1. **`.claude/settings.local.json`** — `moai init` (and the launcher flow `moai web` ultimately hands off to) writes teammate / GLM env state here (CLAUDE.local.md §22.3). The web console's write surface also flows through settings.go helpers.
2. **`.moai/config/sections/*.yaml`** — `moai init` regenerates templates; the web console persists profile / language / statusline edits; `moai profile` mutates profile-scoped config.
3. **`.moai/state/`** — launch ledger, handoff pending record, active-sessions registry, context-usage snapshot.
4. **Web console profile/ledger** — `internal/profile/launch.yaml` carries a single global `last_profile` plus a per-project map (SPEC-PROFILE-MEMORY-001); a concurrent second console on the same project races on write.
5. **Profile directory + git identity** — `moai profile` mutates `~/.moai/claude-profiles/<name>/launch.yaml` AND the profile's `UserName` preference (`internal/profile/preferences.go` `ProfilePreferences.UserName`, a flat string — verified; there is NO `User` struct and NO `Email` field on the profile). The user's full git identity (`user.name` AND `user.email`) lives in `~/.gitconfig` (global git config), NOT in the profile. Two parallel `moai profile` sessions race on the profile dir AND on the per-profile launch ledger; worktree-scoped identity at materialization is sourced read-only from `~/.gitconfig` (REQ-SW-019).
6. **TCP port 3041** — the web console binds loopback only; a second session hits `ensurePortFree` and either reclaims (killing the first) or errors.
7. **`moai` binary runtime state** — any shared cache the binary writes under the shared root.

The web entry point already acknowledges this: `internal/cli/web.go` L79-81 emits `emitWorktreeAdvisory` warning the console runs in the shared primary checkout (per SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-009). This SPEC converts that advisory from prose into an optional automatic worktree entry, and additionally equips the worktree with a worktree-scoped git configuration so each isolated session commits under the right identity on a usable worktree (git 2.35.2+ `safe.directory` non-negotiable).

## §C — Scope

**In scope** (NEW SPEC):

1. A config-gated + env-overridable activation flag (default OFF — see Fork B) that, when ON, routes `moai init`, `moai profile`, and `moai web` into a fresh worktree before any shared-state mutation.
2. Branch naming carrying the literal `[WT]` marker so isolation branches are greppable and traceable (Fork C).
3. Disposal semantics consistent with CLAUDE.local.md §22.8's opt-in posture (Fork D), extended with PR-merge-triggered auto-cleanup.
4. Coexistence — not replacement — with the existing `moai cc -w`, `moai worktree`, and `Agent(isolation: "worktree")` surfaces.
5. Worktree-scoped git configuration applied at materialization: `safe.directory` (essential), the user's global git identity (`user.name`/`user.email`, sourced read-only from `~/.gitconfig`), `init.defaultBranch=main` for `moai init`, and opt-in options (`commit.gpgsign`/`user.signingkey`, `core.autocrlf=input` / `fetch.prune=true` / `push.default=current`). The v0.2.0 `core.hooksPath` opt-in was removed at v0.2.1 (hook isolation is out of scope).
6. PR-merge detection that auto-removes a `[WT]` worktree whose branch merged to its base (`gh pr view` primary, `git branch --merged` fallback with documented squash-merge blindness).

**Out of scope — explicit non-goals** (see §3 Exclusions).

## §1 — Background and Context

### §1.1 — Existing surfaces this SPEC composes with (NOT replaces)

| Surface | Owner | What it does | Composition rule |
|---|---|---|---|
| `moai cc -w <name>` (launcher flag) | `internal/cli/launcher.go` | New-session launcher that passes `-w` through to `claude`; accepts short names (L1, `.claude/worktrees/<name>/`) AND absolute L2 paths (`~/.moai/worktrees/<project>/...`). | This SPEC's auto-entry reuses the same `-w` plumbing. The auto-entry is a **detection + delegation** layer on top, not a parallel worktree implementation. |
| `moai worktree {new,clean,done,remove}` | `internal/cli/worktree/` | Persistent L2 worktree lifecycle (manual). | Disposal of an auto-created `[WT]` worktree uses the same `done` / `remove` verbs — no new verb added. |
| `Agent(isolation: "worktree")` | Claude Code runtime | L1 ephemeral session-scoped worktree for write-heavy sub-agents. | Orthogonal — runs inside whichever checkout this SPEC routes to. |
| `emitWorktreeAdvisory` (web.go L79-81) | `internal/cli/web.go` | Today: prints an advisory when the console is launched in the shared checkout. | When auto-entry is ON and succeeds, the advisory is suppressed (the collision hazard was avoided by construction). When auto-entry is OFF, the advisory continues to fire unchanged. |
| Main-checkout branch guard | `internal/hook/branch_guard.go` | Mechanically denies branch-state-mutating git ops in the primary checkout. | Worktree paths are already exempt from the deny (see `main-checkout-branch-guard.md` § Mechanical Enforcement); the auto-entry path inherits this exemption. |
| `auto-<session-short>-<spec-id>` naming | `worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation | Orchestrator-side ad-hoc worktree naming for foreign-session race mitigation. | Distinct from this SPEC: that scheme fires at orchestrator runtime; this SPEC fires at CLI launch. The two share the deterministic-naming spirit but not the same token (the user's `[WT]` marker is the load-bearing prefix here). |
| `moai worktree clean --stale` / `--merged-only` | `internal/cli/worktree/clean.go` | Existing manual cleanup of stale/merged worktrees. | The PR-merge auto-cleanup (REQ-SW-022/023) is the automatic counterpart; the squash-merge blindness the manual command already hits (memory: `project_clean_stale_squash_blindspot.md`) is inherited by the automatic path's fallback. |
| `internal/profile` `UserConfig` + global gitconfig | `pkg/models/config.go` `UserConfig.Name` (Name-only, no Email — verified L32-37) + `~/.gitconfig` (`user.name` / `user.email`) | moai reads the user's git identity from `~/.gitconfig` (global git config) at worktree materialization; the profile carries only `Name` (no `Email`), so the profile is NOT the identity source. | REQ-SW-019 sources `user.name` AND `user.email` from `~/.gitconfig` read-only and applies both worktree-scoped. |

### §1.2 — Collision surface inventory (verified against current code)

| Surface | Path | Written by | Today's isolation |
|---|---|---|---|
| settings.local.json | `.claude/settings.local.json` | `moai cc/glm/cg` (runtime), web console (profile section) | shared (none) |
| moai config sections | `.moai/config/sections/*.yaml` | `moai init` (template regen), web console (profile/lang/statusline), `moai profile` | shared |
| profile dir | `~/.moai/claude-profiles/<name>/` | `moai profile` (launch.yaml, profile `UserName` preference) | **global per machine — NOT project-scoped** |
| launch ledger | `~/.moai/claude-profiles/<name>/launch.yaml` | `profile.RecordLaunchProfileForProject` (web.go, launcher.go, profile) | global per profile dir |
| handoff pending | `.moai/state/handoff/pending.json` | orchestrator (session-handoff.md) | shared |
| sessions registry | `.moai/state/active-sessions.json` | SessionStart hook | shared |
| TCP 3041 | loopback | web console `ensurePortFree` | single-binding only |

The project-scoped surfaces (1, 2, 5, 6, 7) are simultaneously de-risked if the **whole `moai init` / `moai profile` / `moai web` invocation runs inside an isolated worktree** — the worktree's separate working tree + separate `.moai/state/` is the isolation boundary. The **profile directory** (`~/.moai/claude-profiles/`, surfaces 3+4) is intentionally global per-machine — worktree isolation of the project tree does NOT isolate it. This is the profile-isolation fork resolved in §2.9: profile auto-entry scopes to the project, the profile dir is out of scope.

## §2 — Requirements (GEARS notation)

### §2.1 — Activation posture (Fork B)

**REQ-SW-001 (Ubiquitous, default-off invariant)** — The `moai init`, `moai profile`, and `moai web` subcommands SHALL behave identically to the current shared-checkout behavior when the session-worktree feature is unset or `false`.

**REQ-SW-002 (Capability gate, config flag)** — **Where** `workflow.session_worktree.enabled` is `true`, the `moai init`, `moai profile`, and `moai web` subcommands SHALL materialize an isolated worktree and execute the remainder of the subcommand inside it before any shared-state mutation.

**REQ-SW-003 (Capability gate, env override)** — **Where** the environment variable `MOAI_SESSION_WORKTREE=1` is set, the activation SHALL be forced on regardless of the config flag; where `MOAI_SESSION_WORKTREE=0` is set, the activation SHALL be forced off regardless of the config flag. Env wins over config.

**REQ-SW-004 (Event-detected)** — **When** an isolated worktree cannot be materialized (e.g., not a git repository, `git worktree add` fails, disk full), the invocation SHALL fall back to the shared-checkout behavior and emit a non-blocking stderr notice naming the failure reason. A worktree materialization failure MUST NOT abort the user's invocation.

### §2.2 — Trigger surface (Fork A)

**REQ-SW-005 (Capability gate, scoped surface)** — **Where** the session-worktree feature is on, the auto-entry SHALL apply to exactly three subcommands: `moai init`, `moai profile`, and `moai web`. Other subcommands (`moai cc`, `moai cg`, `moai glm`, `moai update`, `moai worktree`, ...) SHALL NOT auto-enter a worktree under this SPEC. The extension to other subcommands is deferred to a follow-up SPEC.

### §2.3 — Branch naming (Fork C)

**REQ-SW-006 (Event-driven)** — **When** a session-worktree is materialized, the branch name SHALL carry the literal prefix `[WT]` so the branch is greppable and distinguishable from the orchestrator's `auto-<session-short>-<spec-id>` scheme and from user-authored feature branches.

**REQ-SW-007 (Ubiquitous, deterministic name body)** — The branch name body SHALL follow the form `[WT]-<session-short>-<subcommand>` where `<session-short>` is the first 8 characters of the current session UUID (falling back to a 6-byte random hex when no session id is available), and `<subcommand>` is `init`, `profile`, or `web`.

### §2.4 — Disposal (Fork D)

**REQ-SW-008 (Capability gate, default-manual)** — **Where** the existing `workflow.worktree.auto_cleanup` flag (CLAUDE.local.md §22.8) is `false` (the distributed default), the session-worktree SHALL persist after the subcommand exits, and disposal SHALL be the user's responsibility via `moai worktree done` / `remove`.

**REQ-SW-009 (Capability gate, opt-in auto-cleanup on session exit)** — **Where** `workflow.worktree.auto_cleanup` is `true`, the session-worktree SHALL be removed on clean subcommand exit (exit 0); a non-clean exit SHALL preserve the worktree for post-mortem. The cleanup SHALL emit a stderr notice `removed by session-exit cleanup: [WT] ...` distinguishable from the PR-merge cleanup notice (REQ-SW-022).

**REQ-SW-010 (Unwanted)** — The session-worktree feature SHALL NOT delete uncommitted working-tree changes in the worktree on auto-cleanup. If uncommitted changes exist at exit time, auto-cleanup SHALL be skipped with a notice pointing the user at the worktree path.

### §2.5 — Coexistence (Fork composition)

**REQ-SW-011 (Ubiquitous)** — The feature SHALL compose with, not replace, `moai cc -w <name>`, `moai worktree`, and `Agent(isolation: "worktree")`. A user who manually enters a worktree with `moai cc -w foo` and then invokes `moai web` inside it SHALL NOT trigger a nested session-worktree.

**REQ-SW-012 (Event-detected)** — **When** `moai init` / `moai profile` / `moai web` is invoked from inside an existing worktree (detected via `git rev-parse --git-dir != git rev-parse --git-common-dir`), the auto-entry SHALL be skipped with an info notice.

### §2.6 — Advisory suppression

**REQ-SW-013 (State-driven)** — **While** the feature is ON **and** the worktree was successfully materialized, the `emitWorktreeAdvisory` notice in `web.go` SHALL be suppressed.

### §2.7 — Concurrency / isolation guarantee

**REQ-SW-014 (Event-driven, isolation boundary)** — **When** two `moai web` (or `moai init` + `moai web`, or two `moai profile`) invocations run in parallel with the feature ON, the two invocations SHALL land in two different worktrees with two different branch names, two different `.moai/state/` trees, and two different `.claude/settings.local.json` write surfaces.

**REQ-SW-015 (Event-detected, port collision under isolation)** — **When** two parallel `moai web --port P` invocations both target port P, the second to start SHALL still report the port conflict via `ensurePortFree` exactly as today (a worktree does NOT virtualize the loopback port).

### §2.8 — Profile trigger (Addition 1, Fork A extension)

**REQ-SW-016 (Capability gate, profile trigger)** — **Where** the session-worktree feature is on, the auto-entry SHALL additionally apply to `moai profile`. The profile subcommand's auto-entry follows the same activation, naming (`[WT]-<session-short>-profile`), disposal, and fail-back rules as `moai init` / `moai web`.

### §2.9 — Profile isolation scope (Addition 1, fork resolved)

**REQ-SW-017 (Ubiquitous, profile project-scope isolation)** — When `moai profile` is auto-entered, the worktree isolation SHALL scope to the **project** the profile is launched against (the working tree), NOT to the global profile directory at `~/.moai/claude-profiles/`. Worktree isolation DOES NOT isolate the global profile directory or the per-profile launch ledger; the ledger race on `~/.moai/claude-profiles/<name>/launch.yaml` is explicitly out of scope for this SPEC (see §3 Out of Scope — Profile directory lock). The recommended posture is honest: worktree isolation helps the project side of `moai profile`, and a separate profile-dir lock (file lock on `launch.yaml`) is the correct isolation for the global ledger — but it is NOT what this SPEC delivers.

### §2.10 — Worktree-scoped git config (Addition 2)

**REQ-SW-018 (Event-driven, safe.directory MUST-PASS)** — **When** a session-worktree is materialized, moai SHALL register the worktree path via `git config --global --add safe.directory <worktree-path>` (or the worktree-scoped equivalent) so the git 2.35.2+ dubious-ownership guard does not reject the worktree. This requirement is **MUST-PASS**; without `safe.directory` the worktree is unusable on git 2.35.2+ (any subsequent `git` operation inside the worktree exits non-zero with "detected dubious ownership").

**REQ-SW-019 (Event-driven, global-gitconfig identity MUST-PASS)** — **When** a session-worktree is materialized, moai SHALL read the user's global git identity (`user.name` AND `user.email`, via `git config --global --get user.name` / `user.email`, sourced from `~/.gitconfig` / `~/.config/git/config`) and apply both worktree-scoped (`git -C <worktree> config user.name <...>` / `user.email <...>`) so the isolated session commits under the user's normal git identity. This is a read-only copy from the existing global identity — moai does NOT manage the email itself, and the profile/config (which carry only `Name`, no `Email` — verified `pkg/models/config.go` L32-37 + `internal/profile/preferences.go` L14-16) are NOT the source. When the user has no global git identity configured, this requirement is a no-op (git inherits unchanged).

**REQ-SW-020 (Event-driven, init defaultBranch MUST-PASS for init)** — **When** `moai init` creates a new project inside a session-worktree, moai SHALL set `init.defaultBranch = main` on the new repository so the first commit lands on `main` (not the git built-in default `master`). This requirement binds `moai init` only; `moai profile` and `moai web` operate on existing trees and are out of scope for this requirement.

**REQ-SW-021 (Capability gate, options opt-in)** — **Where** the active profile opts in via profile fields (`signingkey`, etc.), the session-worktree SHALL additionally apply worktree-scoped `commit.gpgsign=true` + `user.signingkey=<key>`, and on new repos created by `moai init` the sane defaults `core.autocrlf=input` / `fetch.prune=true` / `push.default=current`. The former v0.2.0 `core.hooksPath` opt-in is removed at v0.2.1 (hook isolation is out of scope — see §3 Out of Scope — Hook isolation). A profile that does not carry the opt-in fields skips each option silently.

### §2.11 — PR-merge auto-cleanup (Addition 3, Fork D extension)

**REQ-SW-022 (Capability gate, PR-merge cleanup activation)** — **Where** `workflow.worktree.auto_cleanup` is `true`, the feature SHALL additionally auto-remove a `[WT]` worktree whose branch has been merged to its base. Turning the toggle ON enables BOTH the session-exit cleanup (REQ-SW-009) AND the PR-merge cleanup; the toggle is the single activation knob for both. With the toggle OFF (the distributed default), PR-merge cleanup is also off. The cleanup SHALL emit a stderr notice `removed by PR-merge cleanup: [WT] worktree <name> (branch merged)` distinguishable from the session-exit cleanup notice (REQ-SW-009).

**REQ-SW-023 (Event-detected, PR-merge detection mechanism)** — **When** checking whether a `[WT]` worktree's branch is merged, moai SHALL use `gh pr view <branch> --json state` as the primary mechanism (state == `MERGED`), and fall back to `git branch --merged origin/main` when `gh` is unavailable. The fallback is **squash-merge blind**: a squash-merged branch looks unmerged to `git branch --merged`, so squash-merged `[WT]` branches will NOT be removed by the fallback path. This blindness is documented and cross-references the same hazard the repo's `moai worktree clean --stale` command already hit (memory: `project_clean_stale_squash_blindspot.md`). The primary (`gh`) path sees squash merges correctly because `gh pr view` reports the PR's actual merge state regardless of merge method.

**REQ-SW-024 (Unwanted, PR-merge dirty guard)** — The PR-merge auto-cleanup SHALL NOT remove a worktree whose `git status --porcelain` is non-empty. Dirty worktrees are preserved with a stderr notice (this carries REQ-SW-010's safety to the PR-merge path).

## §3 — Out of Scope

### Out of Scope — Trigger surface extension

- Auto-entering a worktree under `moai cc` / `moai cg` / `moai glm` / `moai update` / `moai doctor` / `moai spec` is NOT implemented by this SPEC. The user named `init` + `profile` + `web`; extending to other subcommands is a follow-up SPEC (see §4 Forward References).
- `moai cc -w <name>` and `moai worktree new|done|remove|clean` are the explicit manual surfaces and are NOT replaced.

### Out of Scope — Worktree subsystem redesign

- The L1 (`.claude/worktrees/`) vs L2 (`~/.moai/worktrees/<project>/`) distinction is NOT redrawn. The feature reuses whichever path style the existing `-w` plumbing picks.
- The WorktreeCreate / WorktreeRemove hook contracts are NOT altered.

### Out of Scope — Port virtualization

- TCP port 3041 is a loopback resource. The feature does NOT virtualize the port (REQ-SW-015).

### Out of Scope — Profile memory isolation

- The global launch ledger at `~/.moai/claude-profiles/<name>/launch.yaml` (SPEC-PROFILE-MEMORY-001) is intentionally global per-machine. The feature does NOT per-worktree-ize profile memory.

### Out of Scope — Profile directory lock

- The profile directory (`~/.moai/claude-profiles/<name>/`) is NOT project-scoped and is NOT isolated by worktree entry. Two parallel `moai profile` sessions on the same profile name race on `launch.yaml` and on the profile's git identity regardless of worktree isolation. The correct isolation for the profile dir is a file lock, NOT a worktree — this is deferred to a separate SPEC. REQ-SW-017 documents this scope boundary honestly.

### Out of Scope — Branch guard reconfiguration

- The Main-Checkout Branch Guard is NOT touched. Worktree paths are already exempt from its deny; the feature inherits the exemption.

### Out of Scope — Auto-cleanup default flip

- CLAUDE.local.md §22.8 sets `AutoCleanup` / `AutoCreate` / `AutoMerge` all to default `false`. This SPEC does NOT flip any of them. REQ-SW-008 / REQ-SW-022 inherit the default-OFF posture; opt-in (REQ-SW-009 / REQ-SW-022) requires the user to enable `workflow.worktree.auto_cleanup: true`.

### Out of Scope — gh dependency elimination

- PR-merge detection (REQ-SW-023) depends on `gh` for the primary path. When `gh` is absent, the fallback (`git branch --merged`) is squash-merge blind. Eliminating the `gh` dependency (e.g. by parsing the GitHub API directly via a stored token) is a follow-up SPEC, NOT this one.

### Out of Scope — Global sane-defaults enforcement on existing repos

- REQ-SW-021's global sane defaults (`core.autocrlf=input` / `fetch.prune=true` / `push.default=current`) apply to **new repos created by `moai init`**. Forcing these defaults onto pre-existing user repositories is out of scope — it would mutate user state the SPEC was not asked to touch.

### Out of Scope — Profile-vs-profile / per-project git identity

- moai reads the user's global git identity (`user.name` AND `user.email`) from `~/.gitconfig` and applies it worktree-scoped (REQ-SW-019). Per-profile divergence (a personal profile committing under one identity and a work profile under another) is OUT OF SCOPE — the profile/config carry only `Name` (no `Email`), so a profile-keyed identity is not implementable without a schema change this SPEC does not make. Users wanting per-profile identity divergence MUST manage their global gitconfig themselves (or via a separate tool). The worktree inherits the user's normal global git identity unchanged.

### Out of Scope — Hook isolation (`core.hooksPath`)

- The v0.2.0 `core.hooksPath` opt-in (worktree-scoped hook isolation) is REMOVED at v0.2.1. The project's `.claude/hooks/` are MoAI's own harness hooks (SessionStart, Stop, PreToolUse, etc.) shared across worktrees of the same project; duplicating them per-worktree would diverge harness behavior across sibling worktrees. Worktrees inherit the user's global + project hooks unchanged. Genuine per-worktree hook isolation is a follow-up SPEC, not this one.

## §4 — Forward References

- Follow-up SPEC (to be filed if adopted): extend session-worktree auto-entry to `moai cc` / `moai cg` / `moai glm` (Fork A extension path).
- Follow-up SPEC: profile-directory lock (file lock on `~/.moai/claude-profiles/<name>/launch.yaml`) — the correct isolation for the surface this SPEC deliberately leaves unisolated (§3 Out of Scope — Profile directory lock).
- Follow-up SPEC (optional): eliminate the `gh` dependency in PR-merge detection via direct GitHub API + stored token.
- Composes with **SPEC-WORKTREE-BRANCH-GUARD-001** (branch guard exempts worktree paths) and **SPEC-WORKTREE-ENTRY-STRATEGY-001** (EnterWorktree-first policy).
- Inherits the §22.8 default-OFF worktree auto-toggles posture from CLAUDE.local.md.

## §5 — Assumptions

1. `moai init` / `moai profile` / `moai web` are invoked from inside a git working tree (or `moai init` is creating one). The fail-back REQ-SW-004 handles the non-git case.
2. The existing `internal/cli/worktree/` package exposes reusable helpers (worktree add / list / remove) that the auto-entry path can call. (Run-phase M2 must extract or add a `worktree.Add` helper — see plan.md §B BI-1.)
3. The `[WT]` token in branch names does not collide with any existing branch-naming convention in the repo.
4. The user's global git identity (`user.name` / `user.email`) is configured in `~/.gitconfig` (or `~/.config/git/config`); REQ-SW-019 reads this read-only at materialization. When the user has no global git identity, the requirement is a verified no-op. The profile/config (carrying `Name` only — no `Email`) are NOT the identity source.
5. `git config --worktree` is supported on the user's git version (git 2.20+, released 2018 — safe assumption). For `safe.directory` the path-scoped form is used so global state is not polluted beyond the add-only entry.

## §6 — Risks

- **R1 — Hidden coupling on absolute paths.** If `moai init` / `moai profile` / the web console resolves the project root once at startup and caches it, an auto-entry mid-startup could leave stale pointers. Mitigation: M3 verifies the project root is re-resolved AFTER worktree entry.
- **R2 — Worktree-materialization latency on cold caches.** A fresh `git worktree add` on a large repo can take seconds. Mitigation: the materialization step prints a one-line progress notice; REQ-SW-004 ensures the user's invocation is never blocked indefinitely.
- **R3 — `[WT]` branch pollution of `git branch` listings.** Mitigation: REQ-SW-008 / REQ-SW-009 / REQ-SW-022 + a follow-up `moai worktree clean --wt` sweep.
- **R4 — Env-var leak across nested sessions.** A parent that exports `MOAI_SESSION_WORKTREE=1` will propagate to subshells. This is the intended env-override behavior (REQ-SW-003).
- **R5 — safe.directory global pollution.** `git config --global --add safe.directory <path>` accumulates one entry per `[WT]` worktree in `~/.gitconfig`. Over time this list grows. Mitigation: PR-merge / session-exit auto-cleanup (REQ-SW-009 / REQ-SW-022) SHOULD additionally unset the `safe.directory` entry on removal; if cleanup is OFF, the entries accumulate (low harm — they are add-only path allowlists, not authority grants).
- **R6 — PR-merge detection false-negative under squash-merge without `gh`.** REQ-SW-023's fallback path cannot detect squash-merged branches. Mitigation: documented; user can install `gh`, run `moai worktree clean --stale` periodically, or switch the repo's merge method to `merge` / `rebase`.
- **R7 — Profile identity divergence confusion.** A user with a personal profile active in one worktree and a work profile active in another may be surprised that the two worktrees commit under different identities. Mitigation: REQ-SW-019's stderr notice at materialization names the applied identity.

## §7 — Acceptance Criteria Matrix (summary)

The full AC enumeration lives in `acceptance.md`. The summary mapping:

| REQ | AC(s) | Tier-L coverage |
|---|---|---|
| REQ-SW-001 (default-off invariant) | AC-SW-001, AC-SW-002 | GWT: identical behavior when feature off |
| REQ-SW-002 (config flag on) | AC-SW-003 | GWT: worktree materialized when config on |
| REQ-SW-003 (env override) | AC-SW-004a, AC-SW-004b | GWT: env wins over config both directions |
| REQ-SW-004 (fail-back) | AC-SW-005 | GWT: non-git → fall back + notice |
| REQ-SW-005 (scoped surface) | AC-SW-006 | GWT: `moai cc` does NOT auto-enter |
| REQ-SW-006 + REQ-SW-007 (branch naming) | AC-SW-007 | GWT: `[WT]-<8hex>-<subcommand>` shape |
| REQ-SW-008 (default-manual disposal) | AC-SW-008 | GWT: worktree persists after exit by default |
| REQ-SW-009 (opt-in auto-cleanup on exit) | AC-SW-009 | GWT: auto-cleanup on exit 0 when flag on |
| REQ-SW-010 (preserve uncommitted) | AC-SW-010 | GWT: dirty worktree preserved |
| REQ-SW-011 (coexistence) | AC-SW-011 | GWT: manual `-w` + web = no nesting |
| REQ-SW-012 (already-in-worktree skip) | AC-SW-012 | GWT: skip + notice |
| REQ-SW-013 (advisory suppression) | AC-SW-013 | GWT: advisory suppressed when materialized |
| REQ-SW-014 (parallel isolation) | AC-SW-014 | GWT: two parallel webs, distinct worktrees |
| REQ-SW-015 (port collision preserved) | AC-SW-015 | GWT: port conflict still reported |
| REQ-SW-016 (profile trigger) | AC-SW-016 | GWT: `moai profile` auto-enters worktree |
| REQ-SW-017 (profile project-scope isolation) | AC-SW-017 | GWT: profile auto-entry scopes to project, NOT profile dir |
| REQ-SW-018 (safe.directory MUST) | AC-SW-018 | GWT: `safe.directory` registered on materialization |
| REQ-SW-019 (global-gitconfig identity MUST) | AC-SW-019 | GWT: worktree commits under user's global user.name/email read-only from ~/.gitconfig |
| REQ-SW-020 (init.defaultBranch MUST for init) | AC-SW-020 | GWT: new repo's first branch is `main` (complementary to orchestrator runtime detection) |
| REQ-SW-021 (options opt-in SHALL) | AC-SW-021 | GWT: opt-in fields applied (gpgsign + autocrlf/prune/push.default; hooksPath removed v0.2.1); absence = silent skip |
| REQ-SW-022 (PR-merge activation) | AC-SW-022 | GWT: toggle ON enables both exit + PR-merge cleanup |
| REQ-SW-023 (PR-merge detection mechanism) | AC-SW-023 | GWT: `gh pr view` primary, `git branch --merged` fallback (squash-blind documented) |
| REQ-SW-024 (PR-merge dirty guard) | AC-SW-024 | GWT: dirty worktree preserved on PR-merge cleanup |

## §8 — Cross-References

- `worktree-integration.md` § Terminology Glossary (L1/L2), § Worktree Selection Rules, § Parallel-Session Branch Conflict Auto-Isolation (`auto-<session-short>-<spec-id>` scheme — distinct from this SPEC's `[WT]` marker).
- `main-checkout-branch-guard.md` § Mechanical Enforcement (worktree paths exempt).
- `CLAUDE.local.md` §22.8 (web worktree auto-toggles default OFF), §22.9 (branch guard default OFF).
- `internal/cli/web.go` L79-81 — `emitWorktreeAdvisory` hook point (REQ-SW-013).
- `internal/cli/launcher.go` — unified launch path (`moai cc -w` plumbing reused).
- `internal/cli/worktree/` — existing `clean.go` (incl. `--stale` / `--merged-only` — squash-blindness precedent for REQ-SW-023's fallback), `done.go`, `guard.go`, `root.go` (disposal verbs reused).
- `pkg/models/config.go` L32-37 — `UserConfig` (Name-only, no Email) — verified, the reason REQ-SW-019 sources from `~/.gitconfig` instead of the profile.
- `internal/profile/preferences.go` L14-16 — `ProfilePreferences.UserName` (flat string) — profile carries no `User` struct and no `Email`.
- `internal/workflow/worktree_orchestrator.go` L113-121 — `defaultBranchDetectorFunc` / `detectBranch` — runtime default-branch detection (context for REQ-SW-020, see research.md §C.3).
- `internal/profile/sync.go` L27 + `internal/profile/sync_test.go` L77/L145/L399 — `UserConfig.Name` references (Name-only; the v0.2.0 "User.Email" inference is retracted at v0.2.1).
- memory `project_clean_stale_squash_blindspot.md` — `git branch --merged` squash-merge blindness precedent (REQ-SW-023 fallback inherits this).
- SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009 — the advisory hook point's origin).
- SPEC-WORKTREE-ENTRY-STRATEGY-001 (EnterWorktree-first policy — this SPEC's auto-entry is a CLI-launch-time analog).
- SPEC-PROFILE-MEMORY-001 (launch ledger per-project map — naturally worktree-distinct by absolute path; the global profile dir is NOT, see §3 Out of Scope — Profile directory lock).
