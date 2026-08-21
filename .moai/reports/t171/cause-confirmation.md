# t171 — cause confirmation, before any plan is written

The card names step (a) as confirming the cause, and says the fix direction is
decided only after. This is that step. It changes the card: **one of the two
measured premises is false as stated, and the other one is true for a different
reason than the card proposes — a reason that makes it wider and easier to fix.**

Measured in `.claude/worktrees/t171` (`WT-audit-worktree-blind`, base
`1519f2660`), with `CLAUDE_PROJECT_DIR` unset in the shell (`printenv` exit 1).

---

## Premise 1 — "`moai spec audit` resolves the catalog to the primary" — FALSE

### The measurement

A probe SPEC directory carrying a real `spec.md` was created in this worktree
only, then the audit was run from inside the worktree:

```
$ ls -d .moai/specs/*/ | wc -l                                     630
$ find .moai/specs -mindepth 2 -maxdepth 2 -name spec.md | wc -l   628
$ moai spec audit --json | jq .total_specs                         628
   SPEC-WTPROBE-999 present in the audit JSON:                     true
```

The audit **saw the worktree-only probe**. Its `total_specs` equals this
worktree's count of directories that actually carry a `spec.md` — 628 — not its
count of directories, 630.

### Why the original reading went wrong

The card records lane-7's observation as `total_specs 628 = primary 체크아웃
디렉터리 수 (워크트리는 629)`. Counting the same way today:

| Tree | dirs with `spec.md` |
|---|---|
| primary checkout | **627** |
| `.claude/worktrees/t83` (where lane-7 measured) | **628** |

628 is **t83's own count**, not the primary's — the primary has never been 628.
What lane-7 compared was a *directory* count against a *SPEC* count, and the
one-off gap between them is the directories carrying no `spec.md`. The
observation was real; the label on the number was wrong.

### The code agrees

`internal/spec/audit.go` `Audit()`:

```go
baseDir := opts.BaseDir
if baseDir == "" {
    baseDir = "."
}
specsDir := filepath.Join(baseDir, ".moai", "specs")
```

A plain relative `.`. No `FindProjectRoot`, no git common-dir walk, no
`CLAUDE_PROJECT_DIR`. The CLI passes `--base-dir` through and defaults it empty,
so `moai spec audit` reads the tree it is standing in. **The common-dir
upward-resolution hypothesis in the card is disproved** — that code path does not
exist here.

---

## Premise 2 — "the backend runs in the primary, not the worktree" — TRUE, and the cause is `CLAUDE_PROJECT_DIR`

The card attributes this to the same catalog bug. It is not. The CLI and the MCP
surface resolve the project root through **two different functions**, and only
the CLI's is tree-relative.

| Surface | Resolution | In a worktree |
|---|---|---|
| `moai spec audit` (CLI) | `baseDir = "."` | correct |
| `mcp__moai__spec_audit` / `spec_drift` (`mcp_server.go:583`, `:597`) | `resolveProjectDir()` | **primary** |
| `mcp__moai__codex_*` (`mcp_codex.go:1170`, `"cwd": resolveProjectDir()`) | `resolveProjectDir()` | **primary** |

`internal/cli/session.go:242`:

```go
func resolveProjectDir() string {
    if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" { return dir }
    if cwd, err := os.Getwd(); err == nil { return cwd }
    return ""
}
```

### The measurement that pins it

The MCP server is a long-lived subprocess (`.mcp.json` → `moai mcp-server`),
spawned once per session. Reading the live ones on this machine — cwd via `lsof`,
environment via `ps eww`:

| PID | server cwd | `CLAUDE_PROJECT_DIR` in its env |
|---|---|---|
| 76584 | `moai-adk-go/.claude/worktrees/t165` | `/Users/goos/MoAI/moai-adk-go` |
| 12377 | `moai-adk-go/.claude/worktrees/t89` | `/Users/goos/MoAI/moai-adk-go` |
| 55780 | `mo.ai.kr/.claude/worktrees/t134` | `/Users/goos/MoAI/mo.ai.kr` |
| 52841 | `mo.ai.kr` | `/Users/goos/MoAI/mo.ai.kr` |

**`CLAUDE_PROJECT_DIR` is the primary checkout in every case, including the three
servers whose cwd is a worktree.** Two different repositories, three different
worktrees, same result. Claude Code exports the *project* root to MCP
subprocesses, not the worktree the session is working in.

Because `resolveProjectDir()` checks that variable **first**, the cwd is never
consulted. Every MCP-path project-root resolution in a worktree session returns
the primary — deterministically, from the first tool call, with no dependence on
when the server was spawned.

PID 76584 is this session's own server, and PID 12377 is lane-9's `t89` session —
which is exactly the tree lane-9 reported a backend reading from the wrong place.

### The second, smaller hazard behind it

The `Getwd()` fallback is stale for a different reason: a subprocess cannot
follow `EnterWorktree`. This session launched in `t165` and is now in `t171`, and
its server's cwd is still `t165`. That branch is currently unreachable here
because the environment variable is always set, but it is wrong in the same
direction, so a repair that only removes the env branch would land on a stale
value rather than a correct one.

---

## What this changes about the card

- **The impact claim survives and widens.** Auditor agents reach the audit
  through the MCP tools (`plan-auditor` and `sync-auditor` both list
  `mcp__moai__spec_audit` / `spec_drift`), so plan-phase SPECs authored in a
  worktree are invisible to them. But the same seam feeds every other
  `resolveProjectDir()` consumer on the MCP path — goal state
  (`mcp_server.go:466`), `spec.ListDocs` (`:526`), `verify.Load` (`:569`), and the
  convergence state directory (`mcp_convergence.go:561`). Audit is where it was
  noticed, not where it stops.
- **Both proposed fix directions need re-aiming.** "Scan the worktree's catalog
  when a worktree is detected" targets `spec.Audit`, which is already correct. A
  `--project-root` argument would have to be threaded to the **MCP tool inputs**;
  the CLI already has `--base-dir` and it already works.
- **The repair belongs at `resolveProjectDir()`**, which is one function with a
  known call-site list, not at the audit engine.

## What is NOT established

- **No repair is proposed here.** Per the card, direction is decided after the
  cause is confirmed; that is where this document stops. Whether the right answer
  is to prefer a git-worktree-aware resolution over `CLAUDE_PROJECT_DIR`, to
  accept a per-tool root argument, or to keep the primary deliberately for some
  consumers (convergence state may genuinely want one shared location) is a
  design decision with a blast radius across all the call sites above.
- **The lane-9 codex/GLM symptoms are not reproduced.** A `verdict` field
  contradicting its own summary, and a hallucinated REST API, are consistent with
  a backend reading the wrong tree, but nothing measured here demonstrates that
  causal chain. They remain reported symptoms.
- **Whether every consumer is actually wrong.** Being handed the primary is a
  defect for an audit and for a codex `cwd`; it may be correct for something that
  wants one shared project-level location. Each call site needs its own verdict
  before a single change is made to the shared function.
- **Full-suite state.** Nothing was built or tested in this worktree; this is a
  read-and-measure pass only.
