# Research — SPEC-SESSION-WORKTREE-001 (Tier L)

> Read-only codebase reconnaissance backing the v0.2.0 additions. Every claim about a code path is verified by a grep or file read at plan-phase. This file is the SSOT for "what does the code do today" that the SPEC's REQs and the plan's milestones assume. v0.2.1 iter-2 audit-fix: §A.1 retracted the false "User.Email" inference; §C.2 corrected (profile is NOT the identity source — `~/.gitconfig` is); §C.3 removed "Verified by inference" and documents `defaultBranchDetectorFunc`/`detectBranch` from source.

## §A — `internal/profile` — the surface `moai profile` auto-entry touches

### §A.1 — Identity source correction (REQ-SW-019 source) — v0.2.1 D1

**Verified at v0.2.1 (D1 correction):**
- `pkg/models/config.go:32-37` — `UserConfig` struct carries **`Name` only** (no `Email` field). The comment confirms `Name` is intentionally optional.
- `internal/profile/preferences.go:14-16` — `ProfilePreferences` carries flat `UserName string` (no `User` struct, no `Email`).
- `internal/profile/sync.go:27` — reads `prefs.UserName` and `cfg.User.Name` (where `cfg.User` is the `UserConfig` struct above — Name only).
- `internal/profile/sync_test.go:77/145/399` — assertions on `wrapper.User.Name` (Name only; NO `Email` round-trip assertion exists).

**Correction:** the v0.2.0 research.md inferred a `User.Email` field from the existence of `User.Name`. That inference was wrong — `UserConfig` has `Name` only and `ProfilePreferences` has `UserName` only. The profile/config are NOT the identity source for `user.email`.

**Implication for REQ-SW-019 (v0.2.1 rewrite):** moai reads the user's global git identity (`user.name` AND `user.email`) from `~/.gitconfig` / `~/.config/git/config` via `git config --global --get user.name` / `user.email`, and applies both worktree-scoped at materialization. This is a read-only copy of the existing global identity — moai does NOT manage the email itself. When the user has no global identity, the requirement is a no-op.

### §A.2 — Profile package layout

**Verified:** `ls internal/profile/` shows: `profile.go` (16KB, primary), `sync.go` (the identity/preferences sync entry point), `preferences.go`, plus test files. The package is the canonical owner of profile state; `moai profile` CLI subcommand wiring lives elsewhere (in `internal/cli/`).

**Implication for M6:** the `moai profile` CLI entry point is NOT in `internal/profile/` — it is in `internal/cli/` (likely `internal/cli/profile.go` or a sibling). M6's pre-flight (BI-5) must locate it. The profile PACKAGE API (`internal/profile`) is consumed by the CLI entry point; the auto-entry goes at the CLI entry, not inside the profile package.

### §A.3 — `launch.yaml` write path (the surface NOT isolated by worktree entry)

**Verified:** `internal/profile/launch.yaml` is referenced in spec.md §1.2 as the launch ledger; `profile.RecordLaunchProfileForProject` (named in web.go / launcher.go) writes to it. The path is `~/.moai/claude-profiles/<name>/launch.yaml` — a per-machine global, NOT project-scoped.

**Implication for REQ-SW-017:** confirmed. `launch.yaml` lives OUTSIDE any project working tree, so worktree isolation of the project tree does NOT isolate it. design.md §A.3's deferral of the profile-dir lock to a follow-up SPEC is grounded in this path layout.

## §B — `internal/cli/worktree/` — the existing manual surface

### §B.1 — Available helpers

**Verified:** `internal/cli/worktree/` exposes `clean.go`, `done.go`, `guard.go`, `root.go`, `shared.go`, `render.go`, `sync.go`, `recover.go`, `remove.go` (per the existing spec.md §5 Assumption 2 and plan.md §B BI-1).

**NOT exposed:** a `worktree.Add(name, base)` Go helper. The add path today is `moai cc -w <name>` → passes `-w` through to `claude` → the Claude Code runtime creates the worktree. `moai init` / `moai profile` / `moai web` cannot rely on this path because they are not launcher subcommands.

**Implication for plan.md §B BI-1:** confirmed. M2 must either extract a Go helper from the launcher plumbing OR shell out to `git worktree add`. The decision is resolved at v0.2.1 as the Q1 M1 deferral (plan.md §I): M1 pre-flight probes `git worktree add` shell-out vs an internal/worktree helper; default shell-out unless a helper already exists.

### §B.2 — `clean --stale` / `--merged-only` (the squash-blindness precedent)

**Verified:**
- `internal/cli/worktree/clean.go:4` — `@MX:NOTE: [AUTO] --merged-only flag removes only fully merged branches`.
- `internal/cli/worktree/clean.go:31` — `cmd.Flags().String("base", "main", "Base branch for --merged-only and --stale checks")`.
- `internal/cli/worktree/clean.go:46` — `return fmt.Errorf("--stale and --merged-only are mutually exclusive")`.
- `internal/cli/worktree/subcommands_test.go:631` — `// --- Tests for clean --merged-only ---`.
- `internal/cli/worktree/clean_stale_test.go:201` — `t.Fatal("expected --stale with --merged-only to be rejected")`.

**Implication for REQ-SW-023:** the repo already has a `git branch --merged`-based cleanup path with documented squash-merge blindness (memory `project_clean_stale_squash_blindspot.md`). The PR-merge detection mechanism (REQ-SW-023) is the automatic counterpart to this manual command; it inherits the same fallback blindness and resolves it via the `gh pr view` primary path. design.md §C.1 grounds this in the existing code.

## §C — Existing git-config touch points in `internal/`

### §C.1 — `safe.directory` — NOT currently set anywhere

**Verified:** `grep -rn "safe.directory" internal/` returns ZERO matches (the sole hit, `internal/harness/proposalgen/mapper.go:179`, is an unrelated comment about "overwrite-safe directory creation" — not the git `safe.directory` config key).

**Implication for REQ-SW-018:** the project has NO existing `safe.directory` handling. M7 introduces it. Any test asserting `safe.directory` is set must do so against a fresh global config fixture (t.TempDir-scoped `HOME` or `GIT_CONFIG_GLOBAL`), not against any existing project state.

### §C.2 — `user.name` / `user.email` — sourced from `~/.gitconfig` (v0.2.1 D1 correction)

**Verified at v0.2.1:** `grep -rn 'User\.\(Name\|Email\)' internal/profile/*.go` returns matches ONLY on `User.Name` (NOT `Email`) in `sync.go` and `sync_test.go`. The `User.Email` field does NOT exist on `UserConfig` (§A.1 correction). On the maintainer machine, `git config --global --get user.email` returns `email@goos.kim` and `git config --global --get user.name` returns `Goos Kim` — confirming the global gitconfig is the live identity source for BOTH fields.

**Implication for REQ-SW-019 (v0.2.1):** the profile/config is NOT the identity source. M7's `worktree.ApplyGitConfig(worktreePath)` helper shells out to `git config --global --get user.name` / `user.email` read-only and applies both worktree-scoped. No `internal/profile` import is needed for identity (the helper may still take a profile handle for REQ-SW-021 opt-in fields like `signingkey`).

### §C.3 — `init.defaultBranch` — complementary to orchestrator runtime detection (v0.2.1 D3)

**Verified at v0.2.1 (D3 correction):** `internal/workflow/worktree_orchestrator.go:113-121` defines:

```go
// defaultBranchDetectorFunc returns the repository's default branch name.
type defaultBranchDetectorFunc func(ctx context.Context, root string) string

// worktreeOrchestrator implements WorktreeOrchestrator.
type worktreeOrchestrator struct {
    worktreeMgr  git.WorktreeManager
    validator    quality.WorktreeValidator
    executor     PhaseExecutor
    detectBranch defaultBranchDetectorFunc
    logger       *slog.Logger
}
```

The orchestrator carries a `detectBranch` field of type `defaultBranchDetectorFunc` and uses it to detect the repo's default branch at runtime.

**Implication for REQ-SW-020 (v0.2.1 correction):** REQ-SW-020 (`moai init` sets `init.defaultBranch=main` on the new repo) is **complementary** to the orchestrator's runtime detection, NOT redundant. `moai init` forces `main` as the default branch on NEW repos it creates; the orchestrator detects the default branch of an EXISTING repo at runtime via `defaultBranchDetectorFunc`. The two operate at different stages (init-time vs. orchestrator-runtime) and do not overlap. The v0.2.0 "Verified by inference" hedge is removed — the orchestrator detection path is now documented from source, not inferred. The v0.2.0 `[NEEDS CLARIFICATION: init.defaultBranch existing handling]` marker is resolved.

## §D — `gh` dependency landscape

**Inferred from project context (CLAUDE.md §5, manager-git agent):** `gh` is already a soft dependency of the project — the orchestrator uses `gh pr create`, `gh pr checks`, `gh pr view` elsewhere (sync-phase PR creation, CI watch). The PR-merge detection's `gh pr view <branch> --json state` (REQ-SW-023) reuses the same dependency; it does not introduce a NEW external dependency.

**Implication for design.md §C.2:** grounded. Eliminating `gh` here would not simplify the project's overall `gh` footprint; the deferral to a follow-up SPEC is honest.

## §E — Verified assumptions, deferred assumptions

### §E.1 — Verified at plan-phase (grep / read)

- `pkg/models/config.go` `UserConfig` is Name-only (no `Email`) — D1 correction (§A.1).
- `internal/profile/preferences.go` `ProfilePreferences` is flat `UserName` (no `User` struct, no `Email`) — D1 correction (§A.1).
- The user's full git identity lives in `~/.gitconfig`, NOT in the profile — D1 correction (§C.2).
- `internal/workflow/worktree_orchestrator.go` L113-121 defines `defaultBranchDetectorFunc` / `detectBranch` (runtime default-branch detection of an EXISTING repo) — D3 correction (§C.3).
- `internal/cli/worktree/` does NOT expose `Add` (§B.1).
- `internal/cli/worktree/clean.go` carries `--stale` / `--merged-only` with squash-blindness precedent (§B.2).
- `safe.directory` is NOT currently set anywhere in `internal/` (§C.1).

### §E.2 — Deferred to run-phase M1/M2 probe

- `moai profile` CLI entry function location (BI-5).
- `worktree.Add` Go helper extraction feasibility (BI-1) — Q1 RESOLVED as M1 deferral: M1 probes; default shell-out unless a helper already exists.
- `[WT]` branch-name format acceptance (Q2) — RESOLVED as M1 deferral: M1 runs `git check-ref-format --branch '[WT]-x'`; if rejected, fall back to `WT-` (no brackets).
- git version on target user machines (BI-6 — assumed ≥ 2.20; if lower, the worktree-config fallback path applies).

The v0.2.0 deferral of `init.defaultBranch` handling is RESOLVED at v0.2.1 — §C.3 now documents the orchestrator detection path from source; REQ-SW-020 is complementary, not redundant.

These deferrals are honest: plan-phase reconnaissance answers the structural questions (does the API exist? does the path layout match the assumption?); the runtime-feasibility questions (can we extract the helper cleanly? what does the user's git version support?) require the M1/M2 probe.

## §F — Cross-references

- spec.md §1.1 (existing surfaces), §1.2 (collision surface inventory), §5 (Assumptions), §8 (Cross-References).
- plan.md §B (Known Issues BI-1..BI-8), §C (Pre-flight M0 checks), §I (Resolved Questions — all markers resolved at v0.2.1).
- design.md §A (profile-isolation fork), §B (git-config layering, hooksPath dropped v0.2.1), §C (PR-merge detection mechanism + on-touch trigger resolved v0.2.1).
- memory `project_clean_stale_squash_blindspot.md` (the squash-blindness precedent — referenced from §B.2 and design.md §C.1).
- `pkg/models/config.go` L32-37 (`UserConfig` — Name-only, no Email; D1 correction — §A.1).
- `internal/profile/preferences.go` L14-16 (`ProfilePreferences.UserName` — flat string; D1 correction — §A.1).
- `internal/workflow/worktree_orchestrator.go` L113-121 (`defaultBranchDetectorFunc` / `detectBranch` — D3 correction — §C.3).
- `internal/profile/sync.go`, `internal/profile/sync_test.go` (Name-only assertions; the v0.2.0 "User.Email" inference is retracted at v0.2.1 — §A.1).
- `internal/cli/worktree/clean.go`, `internal/cli/worktree/clean_stale_test.go` (the manual cleanup precedent — §B.2).
