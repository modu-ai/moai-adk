# t295 — Launcher Existing-Branch Worktree Path (2026-09-02, lane-7)

## Claim

The launcher gains a sanctioned creation path for worktrees on EXISTING branches —
`moai cc -w <name> --branch <existing>` (also accepted by `cg` and `glm`) — closing the
tool gap that forced a raw `git worktree add` (a [HARD] launcher-route violation) every
time the gitflow chain provisioned the develop integration worktree. The created tree is
registered in the shared worktree state file so `moai worktree clean`'s anchor check sees it.

## Design decisions (the card's two judgment items)

1. **Add the launcher path (yes), not a docs-only exception.** The card's mutant guard names
   exactly this: documenting the exception while the tool stays mute would re-invent the same
   judgment for the next person. The surface is the launcher (where entry and creation fuse
   naturally), not a revived `moai worktree new` — that verb was retired deliberately by
   #1278 and the worktree subcommand family keeps zero creation verbs.
2. **Long-lived trees need no new registry machinery.** The registry that `clean`/`done`/
   `recover` consult is the session-anchor pair (integration lock ∪ `.moai/state/worktrees.json`,
   REQ-WR-019 lock-∪-registry). The launcher-created tree registers in the SAME state file the
   WorktreeCreate hook maintains (`internal/hook/worktree_registry.go`), so protection comes
   from the existing anchor mechanism, not from a new registry class with one consumer.
   Disposal stays L1: session-end keep/remove prompt or manual `git worktree remove` — `done`
   remains L2-only by design.

## Evidence

Measured 2026-09-02 in worktree `.claude/worktrees/t295`, branch `WT-existing-branch`,
based on local develop `2660bcd09` (the develop tip at entry — note it advanced past the
dispatch's `b7462203a` between the t313 and t295 worktree creations; absorbed via
fast-forward merge).

| Check | Command | Result |
|---|---|---|
| New tests | `go test ./internal/cli -run 'WorktreeBranch\|SplitWorktreeBranchFlag\|LauncherWorktreeMaterialize' -count=1` | ok |
| Full touched package (cli) | `go test ./internal/cli -count=1` | `ok ... 454.167s` |
| Full touched package (hook) | `go test ./internal/hook -count=1` | `ok ... 34.369s` |
| Lint | `golangci-lint run ./internal/cli/... ./internal/hook/...` | `0 issues.` |
| Cross-platform build | `GOOS=windows go build ./internal/cli/...` + `GOOS=linux ...` | both exit 0 |
| Template mirror parity | `cp` template → local, `diff -q` | `MIRROR-IDENTICAL` |
| `make build` after template edit | — | exit 0 (catalog.yaml re-stamped) |

**End-to-end smoke (freshly built `bin/moai`, this worktree):**

- POSITIVE: `./bin/moai cc -w zzz-smoke --branch zzz-smoke-branch -- --version` → rc=0.
  Tree created at `.claude/worktrees/zzz-smoke`; `git -C <tree> symbolic-ref --short HEAD`
  → `zzz-smoke-branch` (the EXISTING branch, not a fresh `worktree-zzz-smoke`); registry
  entry written (`agent_name: "launcher"`); the backend launched into the tree and exited
  cleanly. Smoke residue (tree, branch, registry entry) removed after the measurement.
- NEGATIVE: `./bin/moai cc -w zzz-smoke --branch develop -- --version` (develop already
  checked out in the lead's integration worktree) → fails loud with git's message naming the
  holding worktree: `fatal: 'develop' is already used by worktree at
  '.../.claude/worktrees/develop'`, wrapped with the full `--branch:` context. No partial state.

**Integration test** (`TestLauncherWorktreeMaterializeReal_Integration`, real git in a
throwaway repo): existing-branch checkout verified; missing branch → hard error and NO tree
created (the `WorktreeManager.Add` auto-create-with-`-b` trap is closed by an explicit
`show-ref` pre-check — a typo must never silently cut a branch); second run → idempotent
re-entry with an `already exists` notice.

## Files

| File | Change |
|---|---|
| `internal/cli/worktree_branch_flag.go` | NEW — flag parse (`splitWorktreeBranchFlag`), short-name enforcement, materializer (seam), `show-ref` pre-check, registry registration |
| `internal/cli/worktree_branch_flag_test.go` | NEW — parse table, no-op, wiring/stripping, bad-usage, error propagation, real-git integration |
| `internal/cli/cc.go` / `cg.go` / `glm.go` | call the resolver in the shared pre-launch position (before `normalizeWorktreeFlag`); help text + example |
| `internal/hook/worktree_registry.go` | `RegisterLauncherWorktree` exported wrapper (same state file, same mutex) |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | doctrine amended: the launcher `--branch` flag is the one sanctioned existing-branch creation path; the [HARD] launcher rule names it alongside `EnterWorktree`/`-w` (card-id-free, template-neutral) |
| `.claude/rules/moai/workflow/worktree-integration.md` | local mirror (byte-identical) |
| `.claude/rules/local/gitflow-lane-protocol.md` | §2: provisioning path named (`moai cc -w develop --branch develop`) |
| `CLAUDE.local.md` | §4.1 integration procedure: first-provisioning vs re-entry lines |
| `CHANGELOG.md` | [Unreleased] Added entry |

## Baseline-attribution

All measurements in this run, this tree (`WT-existing-branch` @ develop `2660bcd09` + the
commits carrying this change). The full-package cli run (454s) is the heaviest single
measurement; per dispatch rule 5 no full-suite (`go test ./...`) was run locally — CI on the
develop push is the full-suite verdict.

## Gaps

- `claude --worktree <short-name>` re-entry semantics are read from documented doctrine
  (kanban-dispatch.md rename-safe clause, worktree-integration.md) and this repo's daily
  operating pattern, not from Claude Code source; the smoke exercised it behaviorally
  (rc=0, correct tree) without asserting the runtime's internals.
- Windows/Linux: cross-BUILD verified; tests were NOT run on those platforms (CI covers the
  PR head). The materializer shells out to git with a plain argv (no shell), which is the
  same portability shape as the existing WorktreeCreate hook path.
- `moai worktree clean`'s anchor check was not driven live against a launcher-created tree —
  the registration writes the same state-file shape the hook writes, and clean's reader is
  covered by its own tests; the composed path (launcher writes → clean reads) is verified by
  format identity, not by a live sweep.

## Residual-risk

- A `--branch` value pointing at a branch checked out in ANOTHER worktree fails at git's
  layer with a message naming the holder — correct, but the error arrives after the
  pre-checks, not as a pre-validated refusal. Acceptable: git's message is precise.
- The smoke ran `bin/moai cc` against this worktree's real settings surface
  (`.claude/settings.local.json`); moai cc's mutations there are idempotent by design and
  the file is untracked.
- Codex launcher (`codex_launcher.go`) deliberately did NOT gain the flag — its entry-token
  surface is narrower and out of the card's need; extending it is a one-call-site change if
  ever wanted.
