# Design — SPEC-SESSION-WORKTREE-001 (Tier L)

> Architecture decisions for the three v0.2.0 additions: (1) `moai profile` trigger + isolation-scope fork, (2) worktree-scoped git-config layering, (3) PR-merge auto-cleanup detection. Derives from spec.md §2.8-§2.12 and plan.md §F M6/M7/M8; contradicts nothing in spec.md (spec.md is canonical). v0.2.1 iter-2 audit-fix: §A.1 corrected (profile carries `UserName` only, NOT `User.Email` — identity sourced from `~/.gitconfig`); §B.1 tier 2 corrected; §B.3 hooksPath resolved (dropped); §C.3 trigger resolved (on-touch); §D Tier re-check (Tier L retained).

## §A — Addition 1: `moai profile` trigger + isolation-scope fork

### §A.1 — The fork, stated honestly

`moai profile` mutates TWO surfaces:

1. **The project tree** — profile-scoped config, project-level settings.
2. **The global profile dir** at `~/.moai/claude-profiles/<name>/` — `launch.yaml` (the launch ledger) and the profile's `UserName` preference (`internal/profile/preferences.go` `ProfilePreferences.UserName`, a flat string — verified; the profile carries NO `User` struct and NO `Email` field).

A git worktree isolates (1) — the worktree has its own working tree of the project. A git worktree CANNOT isolate (2) — `~/.moai/claude-profiles/` is a per-machine global, not a per-project file. Worktree-isolating the project tree does nothing to the profile dir.

**Note on git identity (v0.2.1 D1 correction).** The user's full git identity (`user.name` AND `user.email`) lives in `~/.gitconfig` (global git config), NOT in the profile. The profile's `UserName` (a single `Name` string, no `Email`) is the profile's display name, not the git commit identity. REQ-SW-019 sources the worktree-scoped identity read-only from `~/.gitconfig`. This is a verified correction — `pkg/models/config.go` L32-37 confirms `UserConfig` is `Name`-only, and `internal/profile/preferences.go` L14-16 confirms `ProfilePreferences` is flat `UserName`.

### §A.2 — Resolution

**Profile auto-entry scopes to the project; the profile dir is NOT isolated.** REQ-SW-017 codifies this. The worktree gives the project side of `moai profile` the same isolation `moai init` / `moai web` get, and the global profile dir continues to be a single-writer surface shared across all sessions.

### §A.3 — Why not a profile-dir lock in THIS spec?

A file lock on `~/.moai/claude-profiles/<name>/launch.yaml` IS the correct isolation for the global ledger surface. But:

- It is a **different isolation mechanism** (advisory file lock, e.g. `flock` or `go-internal/lockedfile`), not a worktree.
- It applies to a **different code path** (`internal/profile.RecordLaunchProfileForProject` and friends), not to the CLI auto-entry path this SPEC owns.
- Bundling it would conflate two concerns and grow this SPEC further past the Tier L ceiling.

The honest posture is: this SPEC delivers worktree isolation for the project surface, and explicitly defers the profile-dir lock to a follow-up SPEC (spec.md §4 Forward References). M6 emits a stderr notice during `moai profile` auto-entry stating the profile dir is NOT isolated, so a user running two parallel `moai profile` invocations on the same profile name is warned, not silently racing.

### §A.4 — What M6 implements, concretely

- Locate the profile-mutating subverbs of `moai profile` (the entry function and the subverbs that write state — NOT read-only verbs like `profile list`). EC-7 binds this.
- Apply the M2 auto-entry pattern: activation → already-in-worktree skip → branch name `[WT]-<session-short>-profile` → materialize → fail-back → apply worktree-scoped git config (M7) → run the profile logic against the project worktree.
- Emit the "profile dir NOT isolated" stderr notice during auto-entry.
- Do NOT touch `internal/profile/launch.yaml`'s write path.

## §B — Addition 2: Worktree-scoped git-config layering

### §B.1 — The layering order (load-bearing)

The four config tiers (REQ-SW-018 / 019 / 020 / 021) are applied in this order at materialization:

```
1. safe.directory       (REQ-SW-018, MUST, --global --add)
2. user.name/email      (REQ-SW-019, MUST, sourced read-only from ~/.gitconfig, applied --worktree or local .git/config)
3. init.defaultBranch   (REQ-SW-020, MUST for init only, on the new repo — complementary to orchestrator runtime detection)
4. options              (REQ-SW-021, SHALL/opt-in, --worktree or local .git/config — hooksPath removed v0.2.1)
```

Tier 1 must run BEFORE any `git` operation inside the worktree — without `safe.directory`, git 2.35.2+ refuses to read the worktree at all. Tier 2 must run BEFORE the user's first commit so the identity is correct from commit #1; it is sourced read-only from `~/.gitconfig` via `git config --global --get user.name` / `user.email` (the profile is NOT consulted — verified `UserConfig` is Name-only). Tier 3 is `moai init`-specific (it sets the default branch of the NEW repo being created inside the worktree, not the worktree's own config); it is COMPLEMENTARY to `internal/workflow/worktree_orchestrator.go` `defaultBranchDetectorFunc` which detects the default branch of an EXISTING repo at runtime — the two do not overlap. Tier 4 is opt-in and last so its absence doesn't block the MUST tiers; `core.hooksPath` was removed at v0.2.1 (see §B.3).

### §B.2 — Why `safe.directory` is the ONE global mutation

Every other config tier in this SPEC is worktree-scoped (`git config --worktree` or the worktree's local `.git/config` file). `safe.directory` is the exception because git's dubious-ownership guard reads it from global config (`~/.gitconfig`); a worktree-scoped `safe.directory` is ignored by the guard. So the global `--add` is non-negotiable. Mitigations:

- `--add` form is idempotent (re-materialization adds the entry once, not N times — AC-SW-018 verifies).
- Path-scoped (each entry allows exactly one worktree path, not a blanket allow).
- Removal on cleanup unsets the matching entry (R5 mitigation, M8 + M4 session-exit both participate).

### §B.3 — `core.hooksPath` — RESOLVED (dropped at v0.2.1, Q4)

The v0.2.0 design carried `core.hooksPath` worktree-scoped as a fork with three options (inherit / per-worktree-seed / drop). **Q4 is RESOLVED at v0.2.1: drop (`core.hooksPath` removed from REQ-SW-021).** The project's `.claude/hooks/` are MoAI's own harness hooks (SessionStart, Stop, PreToolUse, etc.) shared across worktrees of the same project; duplicating them per-worktree (option b) would diverge harness behavior across sibling worktrees. Option (a) — inherit — is what dropping achieves (the worktree inherits the user's global + project hooks unchanged). Genuine per-worktree hook isolation is a follow-up SPEC, not this one. See spec.md §3 Out of Scope — Hook isolation.

### §B.4 — git < 2.20 fallback (EC-9)

`git config --worktree` requires git 2.20+. For older git, M7 detects the version once and writes directly to the worktree's local `.git/config` file (`<worktree>/.git/config`). This is functionally equivalent for the config keys this SPEC sets (they are all simple scalars). A stderr notice names the fallback so the user knows their git is old.

## §C — Addition 3: PR-merge auto-cleanup — trigger shape + detection mechanism

### §C.1 — Detection mechanism (resolved)

Two candidate mechanisms:

| Mechanism | Requires | Squash-merge? | Merge-merge? | Rebase-merge? |
|---|---|---|---|---|
| `gh pr view <branch> --json state` | `gh` CLI + auth | DETECTED (state == MERGED) | DETECTED | DETECTED |
| `git branch --merged origin/main` | nothing | BLIND (branch looks unmerged) | DETECTED | BLIND (no merge commit) |

The repo already hit the squash-blindness hazard in `moai worktree clean --stale` (memory `project_clean_stale_squash_blindspot.md`): `git branch --merged` could not detect squash-merged branches, so 45 candidates yielded only 12 removable.

**Resolution: `gh pr view` primary, `git branch --merged` fallback.** REQ-SW-023 codifies this. The fallback's blindness is documented in the user-facing notice when it fires, NOT silently swallowed (AP-12). This mirrors the repo's existing pattern and is honest about the trade-off.

### §C.2 — Why not eliminate the gh dependency?

A direct GitHub API call (`api.github.com/repos/.../pulls/<branch>`) would remove the `gh` dependency. But:

- It requires a stored token (new credential management surface).
- It requires mapping branch → PR number (extra API call).
- `gh` is already a soft dependency elsewhere in the project (the orchestrator uses `gh pr create`, `gh pr checks`).

Eliminating `gh` here would be a point-solution that doesn't simplify the project's overall `gh` dependency. Deferred to a follow-up SPEC (spec.md §4).

### §C.3 — Trigger shape — RESOLVED (on-touch at v0.2.1, Q3)

The v0.2.0 design carried the PR-merge cleanup trigger as a fork with three options (periodic background tick / on-touch / piggyback session-exit). **Q3 is RESOLVED at v0.2.1: on-touch at `moai worktree list` and `moai session start`.**

- **(a) Periodic background tick** — REJECTED. Adds background-loop lifecycle complexity (goroutine startup, shutdown, signal handling) this SPEC should not own; the user's intent ("pr 머지 된 워크트리는 자동 정리") does not map cleanly to a periodic sweep cadence.
- **(b) On-touch at `moai worktree list` / `moai session start`** — ADOPTED. Reuses existing invocation points; fires when the user is present; no new background loop. The "stale until next moai invocation" latency is acceptable for a default-OFF opt-in feature.
- **(c) Piggyback on session-exit cleanup (M4)** — REJECTED. Misses `[WT]` worktrees whose sessions crashed (no clean exit); those would never be cleaned until another session happens to exit with `auto_cleanup` ON.

The on-touch trigger is pinned in plan.md M8 and verified by AC-SW-022's trigger invariant.

### §C.4 — Dirty-guard carry-over

The dirty guard (REQ-SW-010 / REQ-SW-024) is the SAME check on both cleanup paths: `git status --porcelain` non-empty → preserve. M8 reuses M4's dirty-check helper, not a duplicate. AC-SW-024 binds this for the PR-merge path.

## §D — Tier judgment recorded (v0.2.0) and re-checked (v0.2.1)

**v0.2.0 escalation.** The three additions collectively pushed REQ count 15→24 and AC count 15→24, exceeding the Tier M ceiling (16/16). File surface: `internal/cli/{init,web,profile*}.go` + `internal/config/{defaults,envkeys,loader}.go` + `internal/cli/worktree/` + a new worktree-cleanup helper + tests across these = ≥10 files. Per the tier rule (≥10 files OR needs a design.md → Tier L), this SPEC escalated M → L at v0.2.0 and adds this design.md + research.md. The ceiling is 25/25 at Tier L; this SPEC sits at 24/24.

**v0.2.1 re-check.** The `~/.gitconfig` source change (D1) removes the profile-struct identity surface (M7 no longer imports `internal/profile` for identity) but does NOT shrink the file surface below 10 and does NOT reduce the REQ/AC count. The run-phase still touches:

1. `internal/config/defaults.go` — `SessionWorktreeConfig` struct (M1)
2. `internal/config/loader.go` — sub-key wiring (M1)
3. `internal/config/envkeys.go` — env-var constant (M1)
4. `internal/cli/init.go` (or sibling) — `moai init` auto-entry (M2)
5. `internal/cli/web.go` — `moai web` auto-entry + advisory suppression (M3)
6. `internal/cli/profile.go` (or sibling) — `moai profile` auto-entry (M6)
7. `internal/cli/worktree/` — `Add` extraction + new cleanup helper (M2/M8)
8. New `internal/cli/worktree/pr_merge_cleanup.go` (or sibling) — PR-merge cleanup helper (M8)
9. New `internal/cli/worktree/git_config.go` (or sibling) — `ApplyGitConfig` helper (M7)
10. `internal/cli/worktree/list.go` + `internal/cli/session/start.go` (or siblings) — on-touch trigger wiring (M8, Q3)
11. Tests across all of the above (multiple `_test.go` files)

File surface remains ≥10; REQ/AC count stays at 24/24 (D1 rewrites REQ-SW-019, Q4 shrinks REQ-SW-021 scope but does not remove the REQ, D5 folds notices into REQ-SW-022/009 bodies — none of these changes the count). **Tier L retained at v0.2.1.**

## §E — Cross-references

- spec.md §2.8-§2.12 (REQ-SW-016 through REQ-SW-024 — the requirements these decisions operationalize).
- plan.md §F M6/M7/M8 (the milestones that implement these decisions).
- plan.md §I (open questions carried forward to Implementation Kickoff Approval).
- acceptance.md §C AC-SW-016 through AC-SW-024 (the binary tests that verify these decisions).
- memory `project_clean_stale_squash_blindspot.md` (the squash-blindness precedent informing §C.1).
- research.md (codebase reconnaissance backing every claim about `internal/profile`, `internal/cli/worktree/`, and existing git-config touch points).
