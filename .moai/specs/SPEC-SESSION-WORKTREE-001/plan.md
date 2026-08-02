# Plan — SPEC-SESSION-WORKTREE-001

> **Tier L** implementation plan (escalated from Tier M at v0.2.0; Tier L retained at v0.2.1). Milestones ordered by decision-reversibility (highest-change-likelihood first). Architecture decisions for the three v0.2.0 additions live in `design.md`; codebase reconnaissance for `internal/profile`, `internal/cli/worktree/`, and existing git-config touch points lives in `research.md`. v0.2.1 is the iter-2 audit-fix: all four v0.2.0 `[NEEDS CLARIFICATION]` markers are resolved (no more blocker rounds).

## §A — Context

See `spec.md` §A/§B/§C. Three collision-prone subcommands (`moai init`, `moai profile`, `moai web`) currently run in the shared primary checkout. This SPEC adds an opt-in (default-OFF) automatic worktree entry before any shared-state mutation, leaving today's behavior byte-identical when the feature is off (REQ-SW-001). At materialization the worktree is additionally equipped with a worktree-scoped git config: `safe.directory` (essential) + the user's global git identity (`user.name` / `user.email`, sourced read-only from `~/.gitconfig` — NOT from the profile, which carries `Name` only) + `init.defaultBranch=main` for `moai init` + opt-in options (`commit.gpgsign` / `user.signingkey` + sane defaults; `core.hooksPath` was removed at v0.2.1). On disposal, an opt-in PR-merge cleanup path auto-removes merged `[WT]` worktrees, triggered on-touch at `moai worktree list` / `moai session start` (Q3 resolved).

The feature **composes with** the existing worktree surfaces (`moai cc -w`, `moai worktree`, `Agent(isolation: "worktree")`) and **reuses** the `-w` plumbing rather than introducing a parallel implementation. The user's `[WT]` marker is the load-bearing branch-name token distinguishing these branches from the orchestrator's `auto-<session-short>-<spec-id>` scheme and from user-authored feature branches.

**Tier escalation M → L (v0.2.0).** v0.1.0 was Tier M (15 REQ / 15 AC). The three user-requested additions grow the budget to 24 REQ / 24 AC (ceiling Tier M = 16/16; Tier L = 25/25) and the file surface to ≥10 (internal/cli/{init,web,profile*}.go + internal/profile/* + internal/config/{defaults,envkeys,loader}.go + internal/cli/worktree/ + new internal/worktree-cleanup helper + tests across these). Per the tier-judgment rule (≥10 files OR needs a design.md → Tier L), this SPEC escalates to Tier L and adds `design.md` + `research.md`.

## §B — Known Issues (pre-flight, to be confirmed at M1)

- [BI-1] `internal/cli/worktree/` exposes `clean`, `done`, `remove`, `guard`, `root`, `shared`, `render`, `sync`, `recover` — but **does NOT expose a `worktree.Add` Go helper**. The add path today is `moai cc -w <name>` → passes `-w` through to `claude` (the Claude Code runtime creates the worktree). For `moai init` / `moai profile` / `moai web` we cannot rely on the `claude` runtime, so M2 must either (a) extract a `worktree.Add(name, base)` helper from the launcher plumbing, or (b) shell out to `git worktree add` directly.
- [BI-2] `findProjectRootFn` (web.go L65) resolves the project root once at entry. If the auto-entry happens BEFORE this call, the call resolves the worktree root (correct). If it happens AFTER, the cached root is stale (the R1 risk). M2 ordering: auto-entry FIRST, then project-root resolution.
- [BI-3] `moai init` may be invoked in a directory that is not yet a git repository. REQ-SW-004's fail-back handles this; M2 must verify `git worktree add` produces a non-zero exit that the fail-back catches (and not a panic or an `os.Exit`).
- [BI-4] `workflow.session_worktree` is a NEW config sub-key. It must be added to `internal/config/defaults.go` and survive the loader's schema validation.
- [BI-5] **Profile subcommand entry point.** M6 must locate `moai profile`'s entry function (likely `internal/cli/profile.go` or similar) and verify shared-state mutation happens after a single point where the auto-entry can be inserted. The profile dir (`~/.moai/claude-profiles/`) is global — REQ-SW-017 scopes isolation to the project, M6 must NOT attempt to worktree-isolate the profile dir.
- [BI-6] **`git config --worktree` availability.** Worktree-scoped config requires git ≥ 2.20 (December 2018). M7 should detect the git version once and fall back to the worktree's local `.git/config` file path if `--worktree` is unsupported on the user's git.
- [BI-7] **safe.directory global pollution.** Each materialized worktree adds one entry to `~/.gitconfig`'s `safe.directory` list. M7 must verify the add-only `--add` form is used (idempotent across re-materialization) and that removal on cleanup unsets the matching entry.
- [BI-8] **`gh` availability for PR-merge detection.** M8 must detect `gh` at runtime; absence routes to the `git branch --merged` fallback (REQ-SW-023). The fallback's squash-merge blindness must be surfaced as a documented limitation, NOT silently.

## §C — Pre-flight (M0)

Before M1:

1. Confirm `internal/config/defaults.go` `WorkflowConfig` struct location and add `SessionWorktree SessionWorktreeConfig` field.
2. Confirm `internal/config/loader.go` reads the new sub-key without a schema-validation error.
3. Confirm `internal/cli/worktree/` does NOT export `Add` (BI-1) by grep — drives the M2 decision.
4. Confirm `moai init`'s current invocation order (does it mutate shared state before or after project-root resolution?).
5. Confirm `moai profile` entry function location and shared-state mutation order (BI-5).
6. Confirm identity source paths (v0.2.1 verified at plan-phase): `pkg/models/config.go` L32-37 `UserConfig` is `Name`-only (no `Email`); `internal/profile/preferences.go` L14-16 `ProfilePreferences.UserName` is a flat string (no `User` struct). The profile is NOT the identity source — moai reads `user.name` / `user.email` read-only from `~/.gitconfig` (REQ-SW-019).
7. Confirm `git config --worktree` support assumption (BI-6) — documented git version requirement, no runtime probe at plan-phase.
8. Confirm `internal/workflow/worktree_orchestrator.go` L113-121 `defaultBranchDetectorFunc` / `detectBranch` exist (D3 — verified at plan-phase; REQ-SW-020 is complementary, not redundant).
8. Confirm `gh` is a runtime dependency the project already assumes elsewhere (search `gh pr` in internal/) — drives M8's "is gh already required?" judgment.

## §D — Constraints

1. **Backward-compatibility is non-negotiable.** Default OFF means byte-identical behavior when unset. REQ-SW-001 / AC-SW-001 / AC-SW-002.
2. **No new user-facing prompt.** CLI subagent boundary (C-HRA-008): `moai init` / `moai profile` / `moai web` MUST NOT call AskUserQuestion. Outcomes are decided from config + env + observable state, reported via stderr + exit codes.
3. **No WorktreeCreate / WorktreeRemove hook registration.** Default git worktree behavior (spec.md §3).
4. **Template neutrality (§25).** The new config sub-key's default value ships in `internal/config/defaults.go` (Go code), not in `internal/template/templates/.moai/config/sections/workflow.yaml`. The distributed template must NOT leak the `[WT]` token or any SPEC-ID.
5. **Cross-platform.** `git worktree add` works on linux/darwin/windows; `GOOS=windows GOARCH=amd64 go build ./...` MUST pass.
6. **Profile dir is NOT isolated.** REQ-SW-017 is load-bearing: worktree isolation scopes to the project, NOT to `~/.moai/claude-profiles/`. The profile-dir lock is a separate follow-up SPEC.
7. **git-config application is worktree-scoped, not global-mutating (except safe.directory).** REQ-SW-019 / REQ-SW-020 / REQ-SW-021 use `git config --worktree` or the worktree's local config file. REQ-SW-019 reads the user's `user.name` / `user.email` read-only from `~/.gitconfig` (the profile is NOT the source — verified `UserConfig` is Name-only). REQ-SW-018 (safe.directory) is the ONE global mutation — it must be `--global` because the dubious-ownership guard reads global config.

## §E — Self-Verification (plan-phase)

- [ ] spec.md frontmatter validated against the 12-field SSOT (`spec-frontmatter-schema.md` § Canonical 12 Required Fields).
- [ ] SPEC ID `SPEC-SESSION-WORKTREE-001` passed the pre-write regex self-check (PASS observed).
- [ ] No duplicate SPEC ID in `.moai/specs/`.
- [ ] GEARS notation used throughout spec.md §2; no residual `IF/THEN`.
- [ ] §3 Out of Scope satisfies `OutOfScopeRule` (≥1 `### Out of Scope — <topic>` H3 with ≥1 `-` bullet) — eight such sub-headings present.
- [ ] Acceptance criteria (acceptance.md) in Given-When-Then; binary-testable.
- [ ] Tier L artifact set complete: spec.md + plan.md + acceptance.md + design.md + research.md + progress.md.

## §F — Milestones

### M1 — Config flag + activation detection

**Reversibility: HIGH** (config schema is the contract every downstream milestone consumes).

- Add `SessionWorktreeConfig` struct to `internal/config/defaults.go` with field `Enabled bool` (default `false`).
- Wire the loader to read `workflow.session_worktree.enabled` from `.moai/config/sections/workflow.yaml`.
- Add env-var override resolution: `MOAI_SESSION_WORKTREE=1` forces on, `=0` forces off, unset falls through to config (REQ-SW-003). Env-var name constant in `internal/config/envkeys.go`.
- Expose pure helper `SessionWorktreeEnabled(cfg) bool` (env resolution + config read) — the single decision point consumed by M2/M3/M6.
- Tests: table-driven over (config, env) pairs.

**Exit gate M1:** helper is the ONLY activation decision site; tests pass on linux + windows cross-build.

### M2 — `moai init` auto-entry

**Reversibility: MEDIUM** (init entry-point shape).

- Resolve BI-1: extract a `worktree.Add(name, baseRef)` Go helper from the launcher plumbing OR shell out to `git worktree add`.
- At `moai init` entry, BEFORE any shared-state mutation:
  1. `SessionWorktreeEnabled(cfg)`? If false, run init unchanged (REQ-SW-001).
  2. Already-in-worktree (REQ-SW-012)? Skip + notice.
  3. Branch name `[WT]-<session-short>-init` (REQ-SW-006 / REQ-SW-007).
  4. Materialize; on failure, fall back + notice (REQ-SW-004).
  5. Apply worktree-scoped git config (see M7 — M2 wires the call site, M7 implements the config helper).
  6. Re-resolve project root against the worktree path (BI-2 / R1 mitigation).
  7. Set `init.defaultBranch=main` on the new repo (REQ-SW-020).
  8. Run the existing init logic against the worktree.
- Tests: RED → GREEN; falsification round-trip proving fail-back fires on non-git dirs.

**Exit gate M2:** `moai init` with feature ON materializes a worktree and runs init inside it; OFF = byte-identical.

### M3 — `moai web` auto-entry + advisory suppression

**Reversibility: MEDIUM** (web entry-point shape, advisory hook).

- Mirror M2's auto-entry at `runWeb` (web.go L64), before `ensurePortFree` and before `emitWorktreeAdvisory`.
- Auto-entry success → suppress `emitWorktreeAdvisory` (REQ-SW-013). Off OR fallback → advisory fires.
- Port-handling unchanged (REQ-SW-015).
- Tests: advisory suppression via captured stdout; port-collision path preserved.

**Exit gate M3:** `moai web` with feature ON materializes worktree, runs console inside, suppresses advisory; OFF = unchanged.

### M4 — Branch naming + session-exit disposal

**Reversibility: LOW** (mechanical, once M2/M3 are in place).

- Implement `<session-short>` resolution: session UUID from the registry, fallback 6-byte random hex.
- Implement session-exit disposal:
  - Default-manual (REQ-SW-008): worktree persists after exit.
  - Opt-in auto-cleanup (REQ-SW-009): on clean exit, `git worktree remove`; non-clean → preserve.
  - Dirty-worktree guard (REQ-SW-010): `git status --porcelain` non-empty → skip + notice.
- Reuse `workflow.worktree.auto_cleanup` (§22.8) — no new flag.

**Exit gate M4:** parallel-safe deterministic branch names; session-exit disposal honors the three cases.

### M5 — Tests, coexistence, anti-regression (M1-M4 surface)

**Reversibility: LOWEST** (test-only for the v0.1.0 surface).

- Coexistence tests (REQ-SW-011).
- Already-in-worktree tests (REQ-SW-012).
- Parallel-isolation tests (REQ-SW-014).
- Anti-regression: feature OFF → today's behavior.
- Falsification round-trips for the three guards (fail-back, dirty-preserve, already-in-worktree skip).

**Exit gate M5:** every AC in acceptance.md up to AC-SW-015 has a GREEN test; every guard has a falsification round-trip on record.

### M6 — `moai profile` auto-entry (Addition 1)

**Reversibility: MEDIUM** (profile entry-point shape; NEW v0.2.0 surface).

- Locate `moai profile`'s entry function (BI-5). Apply the M2 auto-entry pattern with:
  1. Branch name `[WT]-<session-short>-profile` (REQ-SW-007).
  2. Worktree scopes to the **project** the profile is launched against (REQ-SW-017) — NOT to `~/.moai/claude-profiles/`.
  3. Same fail-back / already-in-worktree / advisory rules.
- **Explicit non-action:** M6 MUST NOT attempt to isolate the profile dir. The profile dir remains global; the launch-ledger race is documented as out of scope (§3 Out of Scope — Profile directory lock) and deferred to a follow-up SPEC. M6 emits a stderr notice when `moai profile` is auto-entered naming that the profile dir is NOT isolated by this entry.
- Tests: profile auto-entry observes the project-scoped worktree; profile dir path in notices is the global one; parallel profile invocations produce distinct project worktrees but share the profile dir (documented).

**Exit gate M6:** `moai profile` auto-enters a project-scoped worktree; profile dir race is documented, NOT silently "solved".

### M7 — Worktree-scoped git config (Addition 2)

**Reversibility: MEDIUM-HIGH** (config-layer decision; touches init/profile/web materialization).

- Implement a `worktree.ApplyGitConfig(worktreePath)` helper that:
  1. **safe.directory (REQ-SW-018, MUST-PASS):** `git config --global --add safe.directory <worktree-path>`. Idempotent via `--add` (BI-7).
  2. **Global-gitconfig identity (REQ-SW-019, MUST-PASS):** read `user.name` AND `user.email` from the user's global git config via `git config --global --get user.name` / `user.email` (sourced from `~/.gitconfig` / `~/.config/git/config`, read-only); when non-empty, apply both worktree-scoped (`git -C <worktree> config user.name <...>` + `user.email <...>`). Empty (no global identity) → no-op. Detect git < 2.20 and fall back to writing the worktree's local `.git/config` directly (BI-6). NOTE: the profile is NOT the identity source — `UserConfig` (`pkg/models/config.go` L32-37) is `Name`-only; `ProfilePreferences` (`internal/profile/preferences.go` L14-16) is flat `UserName`.
  3. **init.defaultBranch (REQ-SW-020, MUST-PASS for init):** applied at M2 step 7, not here — M7 exposes the helper, M2 calls it. Complementary to `internal/workflow/worktree_orchestrator.go` `defaultBranchDetectorFunc` (runtime detection of an EXISTING repo's default branch), NOT redundant.
  4. **Options (REQ-SW-021, SHALL/opt-in):** when the profile carries the opt-in fields, apply `commit.gpgsign` + `user.signingkey`, and on new repos created by init the sane defaults `core.autocrlf=input` / `fetch.prune=true` / `push.default=current`. The v0.2.0 `core.hooksPath` opt-in is REMOVED at v0.2.1 (hook isolation is out of scope — spec.md §3 Out of Scope — Hook isolation). Absent fields → silent skip.
- Wire M7's helper into M2's materialization step 5 (init), M3's (web), and M6's (profile).
- Emit a stderr notice at materialization naming the applied identity (REQ-SW-019 / R7 mitigation).
- Tests: each config tier asserted in isolation + together; no-global-identity no-op; git < 2.20 fallback path; safe.directory idempotency across re-materialization.

**Exit gate M7:** worktree materialization carries the four config tiers (essential/identity/init-defaults/options); each tier has a falsification round-trip.

### M8 — PR-merge auto-cleanup (Addition 3)

**Reversibility: MEDIUM** (new cleanup path; touches worktree disposal).

- **Trigger RESOLVED (Q3, v0.2.1): on-touch at `moai worktree list` and `moai session start`.** A periodic background tick (option a) was rejected — adds background-loop lifecycle complexity this SPEC should not own. Piggybacking on session-exit (option c) was rejected — misses crashed-session worktrees. On-touch reuses existing invocation points, fires when the user is present, and the "stale until next moai invocation" latency is acceptable for a default-OFF opt-in feature. See design.md §C.3.
- Implement a `worktree.PrMergeCleanup()` helper invoked on-touch at the two named subcommands:
  1. Enumerate `[WT]-*` worktrees.
  2. For each, check merge state per REQ-SW-023:
     - **Primary:** `gh pr view <branch> --json state` → state == `MERGED`? (correctly handles squash/merge/rebase).
     - **Fallback (gh absent):** `git branch --merged origin/main` → branch in merged list? (squash-merge blind — documented).
  3. Dirty guard (REQ-SW-024): `git status --porcelain` non-empty → preserve + notice.
  4. Remove merged + clean worktrees; unset the matching `safe.directory` entry (R5 mitigation). Emit a stderr notice `removed by PR-merge cleanup: [WT] worktree <name> (branch merged)` distinguishable from the session-exit cleanup notice (REQ-SW-009 / REQ-SW-022).
- Activation: same `workflow.worktree.auto_cleanup` toggle (REQ-SW-022) — turning it ON enables BOTH session-exit cleanup AND PR-merge cleanup.
- Document the squash-merge blindness of the fallback path in the user-facing notice (R6).
- Tests: primary-path removal (gh present, PR-merge notice string asserted); fallback-path removal (gh absent, branch merge method); fallback-path false-negative (gh absent, squash merge — branch NOT removed, notice emitted); dirty-guard preservation; toggle-off = no removal; on-touch trigger fires at `moai worktree list` and `moai session start`; notice-string distinguishes PR-merge from session-exit cleanup.

**Exit gate M8:** PR-merge cleanup fires on-touch when toggle ON; dirty worktrees preserved; squash-blind fallback documented + tested; notice distinguishable from session-exit cleanup.

## §G — Anti-Patterns (do NOT)

- **AP-1 — Do NOT flip the default to ON.** Default OFF is REQ-SW-001 and the one property tested most.
- **AP-2 — Do NOT force a worktree for single-session use.** Opt-in only.
- **AP-3 — Do NOT virtualize the TCP port.** Two parallel `moai web` sessions need distinct `--port` values (REQ-SW-015).
- **AP-4 — Do NOT replace `moai cc -w`, `moai worktree`, or `Agent(isolation: "worktree")`.** Composes (REQ-SW-011).
- **AP-5 — Do NOT extend the trigger surface silently.** REQ-SW-005 scopes this to `init` + `profile` + `web`. Adding `moai cc` / `moai cg` / `moai glm` is a follow-up SPEC.
- **AP-6 — Do NOT register WorktreeCreate / WorktreeRemove hooks.** Default git worktree behavior.
- **AP-7 — Do NOT delete uncommitted changes on auto-cleanup.** REQ-SW-010 / REQ-SW-024.
- **AP-8 — Do NOT call AskUserQuestion from init / profile / web.** CLI subagent boundary (C-HRA-008).
- **AP-9 — Do NOT leak the `[WT]` token or SPEC-ID into `internal/template/templates/`.** §25 template neutrality.
- **AP-10 — Do NOT attempt to worktree-isolate the profile dir.** REQ-SW-017 is load-bearing. The profile dir is global; the launch-ledger race is a separate SPEC (file lock, not worktree).
- **AP-11 — Do NOT mutate the user's global git identity.** REQ-SW-019 reads `user.name` / `user.email` read-only from `~/.gitconfig` and applies both **worktree-scoped** only. The profile is NOT the identity source. The only global mutation is `safe.directory` (REQ-SW-018), which is add-only and path-scoped.
- **AP-12 — Do NOT silently swallow the squash-merge blindness of the `git branch --merged` fallback.** REQ-SW-023 requires the blindness to be documented in the user-facing notice when the fallback fires.
- **AP-13 — Do NOT force the sane defaults onto pre-existing repos.** REQ-SW-021's `core.autocrlf=input` / `fetch.prune=true` / `push.default=current` apply to NEW repos created by `moai init` only. (The v0.2.0 `core.hooksPath` opt-in is removed at v0.2.1 — hook isolation is out of scope.)

## §H — Cross-References

- spec.md §A-H (this plan derives from it; contradictions defer to spec.md).
- acceptance.md (AC enumeration that binds run-phase completion).
- **design.md** (Tier L) — architecture decisions for the three v0.2.0 additions: profile-isolation fork resolution, git-config layering order, PR-merge detection trigger shape.
- **research.md** (Tier L) — codebase reconnaissance for `internal/profile`, `internal/cli/worktree/`, existing git-config touch points, and the `clean --stale` squash-blindness precedent.
- `internal/cli/CLAUDE.md` § Key Patterns (`syscall.Exec` replacement, settings.json mutation helpers).
- `internal/cli/CLAUDE.md` § Conventions (subagent boundary C-HRA-008 — no AskUserQuestion from CLI).
- CLAUDE.local.md §22.8 (worktree auto-toggles default OFF — REQ-SW-008/022 inherit).
- CLAUDE.local.md §14 (hardcoding prevention — env-var names in `envkeys.go`).
- CLAUDE.local.md §25 (template neutrality — AP-9).
- memory `project_clean_stale_squash_blindspot.md` (`git branch --merged` squash blindness — REQ-SW-023 fallback inherits).

## §I — Resolved Questions (v0.2.1 iter-2 audit-fix — all forks decided, no more blocker rounds)

All four v0.2.0 `[NEEDS CLARIFICATION]` markers are resolved at v0.2.1. None triggers an orchestrator AskUserQuestion round.

- **Q1 — `worktree.Add` helper extraction vs shell out (RESOLVED as M1 deferral).** M1 pre-flight probes `git worktree add` shell-out vs an internal/worktree helper; default shell-out unless a helper already exists in `internal/cli/worktree/` (research.md §B.1 confirms `Add` is NOT exposed today). No user-facing decision; M1 records the probe outcome in `progress.md §E.2`.
- **Q2 — `[WT]` branch name — square brackets (RESOLVED as M1 deferral).** M1 runs `git check-ref-format --branch '[WT]-x'`; if exit 0, keep the brackets; if rejected, fall back to `WT-` prefix (no brackets) and document in `progress.md §E.2`. No user-facing decision; the check is mechanical.
- **Q3 — PR-merge cleanup trigger (RESOLVED — on-touch).** Trigger is on-touch at `moai worktree list` and `moai session start`. Periodic background tick rejected (lifecycle complexity); session-exit piggyback rejected (misses crashed-session worktrees). Pinned in design.md §C.3, plan.md M8, AC-SW-022.
- **Q4 — `core.hooksPath` worktree-scoped (RESOLVED — drop).** `core.hooksPath` is REMOVED from REQ-SW-021 at v0.2.1 (option (c) in design.md §B.3). Hook isolation is out of scope — spec.md §3 Out of Scope — Hook isolation. The project's `.claude/hooks/` are MoAI's own harness hooks shared across sibling worktrees; duplicating them per-worktree would diverge harness behavior. Genuine per-worktree hook isolation is a follow-up SPEC.
