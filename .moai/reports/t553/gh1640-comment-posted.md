Both defects have been repaired on the development branches. **Neither repair is in any released version yet** — this comment records where the work landed, not that it shipped.

Re-measured 2026-09-08 in a linked worktree at `origin/develop` = `3ac58b5a1`.

## You were right at report time

The report was filed against v3.1.2 (`a1b1ca696`), and the repairs postdate it. Measured ancestry:

```
git merge-base --is-ancestor 1f9deed0c v3.1.2    -> 1 (absent)
git merge-base --is-ancestor db1ac0afa v3.1.2    -> 1 (absent)
git tag --list 'v3*' --sort=-v:refname | head -1 -> v3.1.2
```

So on the revision you measured, the documented `project_root` parameter genuinely was not registered. The end-to-end consequence you recorded — a probe SPEC in the moved-to worktree returning an empty catalog entry with count 184 / probe 0 / warnings 0, no error and no warning — was a correct observation of that revision.

## Defect 2 — `project_root` on the catalog tools

Landed in `SPEC-MCP-WORKTREE-ROOT-001`:

| commit | scope |
|---|---|
| `1f9deed0c` | `project_root` on the three SPEC tools |
| `2c0efade0` | the codex path and `audit_multi` |
| `34de07740` | post-repair check, and the defect that check found in the repair |
| `db1ac0afa` | canonicalize `project_root`, so a symlink cannot outlive a boundary |
| `21734f9e9` | the verify tools honor it; catalog responses carry `_root` provenance |

The resolver lives at `internal/cli/mcp_project_root.go`. An unusable path is **rejected** rather than replaced by the default (`:160` / `:163` / `:187`) — silent fallback is deliberately not what happens, because it would return a caller who mistyped its own worktree path to acting on the primary checkout and report success. An accepted path is canonicalized, symlinks resolved (`:181`).

Twelve tools now declare the parameter — including `verify_snapshot` and `verify_trend`, which your patch draft also proposed. `audit_multi` takes it with pass-through semantics (absent supplies no root at all, rather than substituting a default), so an existing caller's backends receive exactly what they received before.

Scoped tests green on `3ac58b5a1`:

```
go test ./internal/cli/ -run 'ProjectRoot|project_root|Worktree' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  16.218s
```

## Defect 1 — the env stamp going stale across worktree switches

Repaired by `f1b379434` — a PostToolUse branch for `EnterWorktree` / `ExitWorktree` that re-stamps the env file and relocates the session registry entry. This is your proposed patch item (a), landed independently.

One correction to the mechanism, which does not change your conclusion: the MCP `project_root` fallback reads `CLAUDE_PROJECT_DIR`, not `MOAI_PROJECT_DIR`. At `a1b1ca696` the resolver already read `CLAUDE_PROJECT_DIR` (`internal/cli/session.go:243`), and `MOAI_PROJECT_DIR` has a producer but no Go consumer — `internal/hook/cwd_changed.go` stamps it, and the comment at `:70` records "No Go code consumes MOAI_PROJECT_DIR yet (verified 2026-09-02)". The substance of the defect stands unchanged, because `CLAUDE_PROJECT_DIR` also names the primary checkout for a session working inside a worktree, which is exactly the wrong-tree read you measured. Only the variable named as the carrier was off by one.

Your `unset MOAI_PROJECT_DIR && <command>` workaround is therefore not the load-bearing one for the MCP path; passing `project_root` explicitly is, once a build carrying the repair is available.

## Where the repairs sit relative to `main`

```
1f9deed0c  in main
2c0efade0  in main
db1ac0afa  in main
21734f9e9  NOT in main (development branch only)
f1b379434  NOT in main (development branch only)
```

## What is still open

The repairs have landed but **have not been cut into a release**; the newest tag is still `v3.1.2`, which carries none of them. No release number is promised here — this comment cites only what landed and where. The issue stays open until a build carrying these commits is published.

Two items from your "verification still needed" list are not closed by this re-measurement, and I am not claiming otherwise:

1. Whether PostToolUse actually fires on `EnterWorktree` in a live session — the landed hook depends on that event, and what I verified is the hook's existence, not its live firing.
2. The end-to-end parameter path was exercised by tests here, not by live MCP calls from a moved worktree.

Thank you for the measurement discipline in the original report — the control group, the traced event counts, and the count-184 / probe-0 / warnings-0 figure are what made the silent-omission shape legible.
