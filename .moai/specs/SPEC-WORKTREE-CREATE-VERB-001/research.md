# SPEC-WORKTREE-CREATE-VERB-001 — Research

> Card t1070 (class C). Baseline tree: worktree `.claude/worktrees/t1070`, branch `WT-worktree-verb`, HEAD `0314801c2`, aligned to local `develop`. All facts below were measured in THIS tree; file:line anchors are from this baseline.

## A. Measured code facts

### Fact 1 — `internal/cli/worktree/` carries NO creation verb

Files in `internal/cli/worktree/`: `clean.go`, `done.go`, `guard.go`, `recover.go`, `remove.go`, `render.go`, `launch_ledger.go`, `root.go`, `shared.go`, `sync.go` (10 non-test files, + tests). No file implements creation. (`shared.go` carries helpers extracted from the retired `worktree new` for the clean/sync/done consumers, and `sync.go` is `moai worktree sync` — both verified creation-free by the iter-1 plan-audit.) `moai worktree new` is RETIRED: `.claude/rules/moai/development/branch-origin-protocol.md` § Retired records that the internal/bodp library, `moai worktree new`, and its `--base`/`--from-current` flags were all retired; a worktree is entered with `moai cc -w <name>`.

### Fact 2 — `materializeSessionWorktree` is the existing harness-neutral creation path

- `internal/cli/session_worktree.go:203` — `materializeSessionWorktree`: runs `git worktree add -b <branch> <destDir> [<base>]` (args built by `gitWorktreeAddArgs` at `:255`).
- Base branch selection: `config.LoadWorktreeBaseBranch(projectRoot)` per SPEC-WORKTREE-BASEREF-001 (REQ-WBR-010/011 — no-operand fallback keeps old behavior).
- Post-create: best-effort `init.defaultBranch=main` + M7 worktree-scoped config tiers (`applyWorktreeGitConfig`).
- Gating: default-OFF via `config.SessionWorktreeEnabled(cfg)`.
- Consumption today: only via `enterSessionWorktree` (`:163`) as a side effect of `moai init` / `moai web` / `moai profile`. Never exposed as a verb.

### Fact 3 — `moai cc -w` delegates creation OUTSIDE moai

`moai cc -w <name>` delegates creation to the `claude` binary it syscall.Execs (Claude Code's native `EnterWorktree`). The creation capability therefore lives outside moai and is not harness-neutral: a Codex lane or a script cannot use it.

### Fact 4 — `moai codex -w` resolves, never creates

- `internal/cli/codex_launcher.go:272` — `resolveCodexWorktreeDir` resolves L1 (`.claude/worktrees/<value>`) and L2 (absolute paths via `resolveWorktreeL2Path`) but never creates.
- Missing tree → structured error: `worktree %q does not exist ... moai codex resolves an existing worktree and never creates one — create it first, then re-run this command`.
- The comment block at `codex_launcher.go:257-267` explicitly records the gap: "Consuming the flag means moai would have to implement worktree CREATION to match cc, which belongs to the worktree tooling that already owns it."

### Fact 5 — doctrine demands launcher entry, but only one harness has a creation path

`.claude/rules/moai/workflow/kanban-dispatch.md` § Isolation, `AGENTS.md` §3, and `.claude/rules/local/gitflow-lane-protocol.md` §1 all mandate: enter through the launcher, never bare `git worktree add`. Yet Codex lanes, scripts, and harness-neutral provisioning have NO authorized creation path today. The gitflow integration chain provisions the develop worktree via `moai cc -w develop --branch develop` (launcher existing-branch flag, landed 2026-09-02) — a claude-harness-only path.

## B. Retired-verb history (why "revive" is not "restore")

The retired `moai worktree new` was removed together with the bodp library and its `--base`/`--from-current` flags. Reviving a verb under that name (option 가) means a NEW command wired to the CURRENT plumbing (`materializeSessionWorktree`), not a restoration of the retired implementation. The retired flags' semantics must not be silently reintroduced.

## C. Decision axis — RESOLVED: option 가 (`moai worktree new`)

The card's [HARD] clause required measurement before choosing between:

- **(가)** Revive `moai worktree new` as a first-class verb under the existing `internal/cli/worktree/` namespace.
- **(나)** Narrow the surface to `moai codex -w --create` (extend the codex launcher's resolve path with creation).

**Resolution evidence:** the 2026-09-22 operator relay records a live Codex lane entering the already-existing t1070 tree through `moai codex -w`; the resolve-existing half is therefore present. The factory lane design (t1082) and script provisioning need a harness-neutral creator before either Claude or Codex enters. Option 나 would expose creation only through the Codex launcher and leave that consumer asymmetry intact. Option 가 is selected, but only as a new thin adapter to Fact 2; none of the retired implementation, BODP layer, or retired flags returns.

## D. Constraints carried into the SPEC

- Reuse-first: wire `materializeSessionWorktree` plumbing + `resolveWorktreeL2Path` validation + `LoadWorktreeBaseBranch` into the chosen verb surface.
- L1 placement: `.claude/worktrees/<name>`; L1/L2 boundary per `worktree-integration.md` § Terminology Glossary.
- CLI subagent boundary (`internal/cli/CLAUDE.md` C-HRA-008 / REQ-PGN-012): no interactive prompts; static guard test (TestNew_NoAskUserQuestion pattern).
- The default-OFF session-worktree gate (REQ-SW-001 byte-identical baseline) must not be silently flipped.
- Errors in English (`error_messages: en`); comments in English (`code_comments: en`).

## E. Gaps

- Live Codex session observation: closed by the operator relay recorded in `progress.md` §E.2; the launcher was observed resolving the existing t1070 tree. A missing-tree launch was not re-run in this lane because the committed resolve-only regression test already preserves that negative contract.
- Exact coverage of `session_worktree.go` tests under the new wiring: not measured at plan-phase (run-phase scope).
