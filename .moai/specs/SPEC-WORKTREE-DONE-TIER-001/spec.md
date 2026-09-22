---
id: SPEC-WORKTREE-DONE-TIER-001
title: "moai worktree done L1 tier guard — make the L2-only doctrine true in code"
version: "0.1.0"
status: completed
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.1.5 target"
module: "internal/cli/worktree"
lifecycle: spec-anchored
tier: M
tags: "worktree,done,tier-guard,L1,L2,data-loss,doctrine,card-t1073"
---

# SPEC-WORKTREE-DONE-TIER-001 — moai worktree done L1 tier guard

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-22 | manager-spec | Initial plan-phase SPEC (card t1073). Code guard makes the `[HARD]` L2-only doctrine true for `done`; minimal doctrine precision edit included. |

## §A Problem Statement

`.claude/rules/moai/workflow/worktree-integration.md` line 52 carries a `[HARD]` claim:

> "moai worktree verbs are L2-only. An L1 tree under `.claude/worktrees/` is never registered with moai worktree, so `done`, `clean`, and `recover` cannot act on it — `moai worktree done` on an L1 tree is a category error, not a disposal."

where L1 = `<repo>/.claude/worktrees/<name>` and L2 = `~/.moai/worktrees/<project>/<SPEC>` (glossary table, lines 44-45 of the same file).

**The claim is refuted by code.** `internal/cli/worktree/done.go` — `runDone` (interactive, line 124) and `runDoneWorktreeCleanup` (`--auto` core, line 51) — resolves the target via `WorktreeProvider.List()` (= `git worktree list --porcelain`, unfiltered, `internal/core/git/worktree.go:76`), matches `wt.Branch == branchName`, then calls `WorktreeProvider.Remove(targetPath, force)`. There is NO path predicate, NO tier constant, and NO registry consult anywhere in `internal/cli/worktree` (10 non-test files). An L1 tree resolves and removes exactly like an L2 tree.

**Measurement quality note (recorded per the verification-claim invariant).** The originating card's grep citation used a BRE pattern with literal pipes — an unfirable pattern shape whose 0-row result was a pattern artifact, not evidence. The plan-phase re-measurement on this tree (0314801c2) used a valid ERE alternation (`paths\.|WorktreesDir|\.claude/worktrees|\.moai/worktrees`): still 0 rows over the package, and the SAME pattern fires in 9 files elsewhere (`internal/cli/glm.go`, `session_worktree.go`, `cc.go`, `launch_chain.go`, `hook.go`, `launcher.go`, `spawn.go`, `worktree_branch_flag.go`, `internal/session/registry.go`) — positive control passed, so the silence is measured absence, not blindness. Positive controls in the run phase MUST reuse the same pattern shape.

**Why the chosen disposal is a code guard, not a doctrine edit.** The doctrine is the desired behavior (an L1 tree is session-scoped; its only copy of unmerged card work must not be silently disposable by a bulk-cleanup verb). Making the doctrine true in code is safer than weakening the doctrine, and the L1 remedy paths remain always available (see REQ-002 rationale), so nothing is stranded by the guard.

## §B Current Protections and the Gap

The only existing protections for an L1 tree, none tier-based:

1. **Anchor guard** — `session.LiveAnchoredSessions` (`internal/session/anchor.go:49`; registry + PID-alive-or-heartbeat-fresh liveness; refuses without `--force`). Silent for a COOLED tree (session exited).
2. **Git worktree lock** — held by a live session, auto-released when dead.
3. **Git's own dirty/untracked refusal** without `--force` (mapped to `ErrWorktreeDirty`, `worktree.go:121-122`).

A **cooled, clean L1 tree** has NO protection. The loss-shape matrix is pinned in §D and encoded as RED-now reproduction tests in acceptance.md.

## §C Requirements (GEARS)

### REQ-001 — L1 tier refusal (Ubiquitous + Event-driven)

**When** a `moai worktree done` invocation resolves a target worktree whose canonicalized path lies under `<repoRoot>/.claude/worktrees/`, **the `done` command shall refuse the removal with a non-zero exit before invoking `WorktreeProvider.Remove`, in BOTH the interactive path (`runDone`) and the auto path (`runDoneWorktreeCleanup`).**

### REQ-002 — Not bypassable by --force (Unwanted)

**The L1 tier refusal shall not be bypassable by `--force`.**

Rationale: the doctrine's L1 remedy — the session-end keep/remove prompt, or `git worktree unlock <path>` + `git worktree remove <path>` — remains always available, so an explicit L1 disposal is never stranded; the guard only removes the silent bulk-verb path.

### REQ-003 — Refusal message (Event-driven)

**When** the tier refusal fires, **the `done` command shall print to stderr a message that (a) names the target path, (b) states that L1 trees under `.claude/worktrees/` are session-scoped, and (c) gives the two remedy forms verbatim (session-end keep/remove prompt; `git worktree unlock <path>` + `git worktree remove <path>`), in the style of the existing `lockGuidance` message (done.go:116-122).** Both `--auto` and interactive modes exit non-zero (auto-mode failures already exit non-zero by design — t41 c).

### REQ-004 — Canonicalized path comparison against a CWD-independent repo root (Event-driven)

**When** the tier predicate evaluates a target path, **the `done` command shall resolve the MAIN repository root from the TARGET path itself — never from the process CWD — by running `git -C <targetPath> rev-parse --path-format=absolute --git-common-dir` and taking the parent directory of the returned common dir as the main root** (git >= 2.31, the `--path-format` floor). **For older git, the command shall fall back to the FIRST `worktree` stanza of `git -C <targetPath> worktree list --porcelain` (the main worktree, present in every porcelain output).** The L1 prefix is then `<mainRoot>/.claude/worktrees/`, compared against the canonicalized target path (`filepath.EvalSymlinks` best-effort with fallback to the raw path on error — macOS `/tmp` vs `/private/tmp` lesson).

Rationale: `git rev-parse --show-toplevel` (the existing `gitRepoRootFunc`) returns the LINKED WORKTREE's own path when invoked from inside one, so a CWD-anchored prefix (`<worktree>/.claude/worktrees/`) never matches the real L1 targets and the guard fails open in the primary operating mode — `done` invoked from inside a worktree still finds and removes the L1 tree via the repo-global `List()` (plan-audit iter-1 D1, measured live). The resolver stays a test seam (overridable function), but its DEFAULT is the target-derived mechanism above.

### REQ-005 — L2 flow unchanged (Event-driven)

**When** the resolved target lies OUTSIDE `.claude/worktrees/` (the L2 shape, e.g. `~/.moai/worktrees/...` resolved to `feature/SPEC-<ID>` via `resolveSpecBranch`, shared.go:32), **the `done` command shall behave exactly as before this SPEC** — including the deployed sync flow `moai worktree done SPEC-{ID} --auto --delete-branch` (`.claude/skills/moai/workflows/sync/delivery.md` "Execute cleanup equivalent to `moai worktree done SPEC-{ID} --auto --delete-branch`" — local copy line 383, template mirror line 358) — with no refusal, no extra output, and no behavior change for the anchor guard, lock guidance, or launch-ledger pruning.

### REQ-006 — Reproduction-first two-cell adoption (Ubiquitous)

**The run phase shall, before landing the guard, pin the loss-shape scenarios A1/A2/A3/C of §D as failing reproduction tests (real git repos in `t.TempDir`, never the project tree), each pinned to the pre-fix tree SHA with the verbatim command, output, and exit code, in a new test file `internal/cli/worktree/done_l1_tier_guard_test.go`** following the existing patterns in `done_test.go` / `done_anchor_test.go` (DI fake `WorktreeProvider` for pure logic; real git repos for git-semantics cells).

### REQ-007 — Doctrine precision edit (Ubiquitous)

**The SPEC shall carry ONE bounded precision edit to `worktree-integration.md` line 52 (and its byte-identical template mirror at `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`) so the sentence states:** `done` refuses L1 trees by code (tier guard, this SPEC), `clean`'s `--merged-only` WT- sweep is the documented exception (lines 64-67 of the same file — an internal tension that predates this card), and `recover` is unchanged. The edit stays within a few lines. `kanban-dispatch.md`'s "`moai worktree done` closes L2 trees only" claims become TRUE post-guard and are edited only if factually wrong.

### REQ-008 — Scope confinement (Unwanted)

**The SPEC shall not modify `clean.go`, `remove.go`, `recover.go`, or any anchor logic.** Tier blindness in `remove`/`recover` is recorded as an observation only.

## §D Loss-Shape Matrix (pinned as RED-now tests, then flipped by the guard)

| ID | Scenario | Today (RED-now) | Post-guard |
|----|----------|-----------------|------------|
| A1 | Cooled CLEAN L1 + `done <branch>` (no flags) | Tree removed silently (branch survives; commits safe) | REFUSED |
| A2 | Cooled DIRTY L1 + `done` | Refused by git (`ErrWorktreeDirty`) | Refused by tier guard (message improvement allowed) |
| A3 | Cooled DIRTY L1 + `done --force` | Removed; uncommitted work DESTROYED | STILL REFUSED — not bypassable (REQ-002) |
| A4 | Cooled L1 + `done --delete-branch` on unmerged branch | Tree removed; branch survives (`DeleteBranch` uses safe `-d`, worktree.go:207) | Refused pre-removal |
| C  | Cooled L1 whose only dirtiness is IGNORED files (e.g. `.git/info/exclude` entry) | `done` removes WITHOUT `--force`; ignored files (local .env-type data) silently lost — the one unconditional-loss shape | Refused (L1) |
| B  | L2-shaped tree outside `.claude/worktrees/` + `done SPEC-<ID> --auto --delete-branch` (post-merge) | Removed — deployed legit flow | MUST STILL PASS unchanged (both flags) |

## §E Acceptance Criteria

See `acceptance.md` (AC-001 .. AC-010). Every criterion carries a RED-now cell (pre-fix failure pinned to tree SHA + verbatim command + output + exit code) and a green-path cell naming the milestone that flips it.

## §F Non-Functional Constraints

- Test isolation: all test temp directories under `t.TempDir()`; never the project tree (CLAUDE.local.md §6).
- Code comments in English; error messages in English; conventional-commit style.
- The guard is unconditional — no config flag, no opt-out (a data-loss guard with an opt-out is a hazard, not a guard).
- @MX tags per constitution at the new guard code site (run phase): `@MX:ANCHOR` with `@MX:SPEC: SPEC-WORKTREE-DONE-TIER-001` on the guard function (public decision boundary), `@MX:WARN` with `@MX:REASON` is not warranted (the guard REMOVES a danger); `@MX:NOTE` on the predicate's symlink fallback.

### Out of Scope — clean / remove / recover tier blindness

- `clean --merged-only` intentionally acts on merged `WT-` L1 trees — documented designed behavior (`worktree-integration.md` lines 64-67). OUT OF SCOPE; the doctrine edit (REQ-007) names it as the documented exception instead of changing it.
- `remove` / `recover` remain tier-blind — recorded here as observations only; no code change in this SPEC.

### Out of Scope — L2 ignored-file loss (pre-existing deployed behavior)

- The same ignored-file loss shape (loss-shape C) remains possible on L2 disposal via `done` — this is pre-existing deployed behavior on the L2 flow and is NOT fixed by this SPEC. Recorded as an observation; a future card may address it.

### Out of Scope — registry and listing changes

- No change to `WorktreeProvider.List()` filtering, no L1 registry, no tier metadata on worktree entries. The guard is a pure path predicate at the `done` call sites.

### Out of Scope — anchor guard, lock guidance, launch-ledger pruning

- These three mechanisms are unchanged (REQ-005). The tier guard composes BEFORE the anchor guard in the removal sequence but replaces none of them.

## §G Success Criteria

- All RED-now tests of §D flip green after M2; no regression in the existing `internal/cli/worktree` suite; affected-package tests pass (`go test ./internal/cli/worktree/...`), full-suite verdict left to CI.
- `go vet` + `golangci-lint` clean on the touched package.
- Doctrine edit present in BOTH the local rule and the template mirror, byte-identical wording.
