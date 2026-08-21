# Plan — SPEC-MCP-WORKTREE-ROOT-001

> Tier S: two artifacts, criteria inline in `spec.md` §3. The two decisions most
> likely to be argued come first — whether `spec_progress` belongs in scope, and
> whether `audit_multi` is worth a seam widening — because those are the places
> this plan spends more than it was handed.

## §A Context

Five MCP handlers resolve the project root through `resolveProjectDir()`, which
prefers `CLAUDE_PROJECT_DIR`, which Claude Code sets to the primary checkout even
for a worktree session. The repair gives the caller a way to say which tree it
means. The caller knows: an agent with Bash runs `git rev-parse --show-toplevel`
and gets its own worktree.

Surfaces, measured in this worktree at `edcbf593c`:

| File | Line | What changes |
|---|---|---|
| `internal/cli/mcp_server.go` | 526, 583, 597 | `project_root` input on `spec_progress` / `spec_audit` / `spec_drift` |
| `internal/cli/mcp_codex.go` | 1170 | `project_root` overrides the `"cwd"` handed to codex |
| `internal/cli/mcp_convergence.go` | 353, 485 | `backendCallFn` widens to carry the root through `audit_multi`'s fan-out; injected doubles follow |
| `internal/cli/mcp_server.go` | tool registration block (~361-400) | declare the new optional input |
| `.claude/rules/moai/core/moai-mcp-tools.md` + mirror | — | record the parameter and who passes it |
| `.claude/agents/moai/{plan-auditor,sync-auditor}.md` + mirrors | — | tell the auditor to pass its own tree |

## §B Known issues going in

- **`resolveProjectDir()` is shared.** The temptation is to fix it once at the
  bottom. That would move the five undecided consumers at the same time, in a
  direction nobody has established is right for them. The parameter goes in at
  the handler, above the shared function.
- **A validation that falls back silently reintroduces the bug.** REQ-3 exists
  because the obvious ergonomic choice — ignore a bad path, use the default — is
  the one choice that makes a typo look like success.
- **The stale-cwd branch is a trap for the follow-up.** Removing the env branch
  without also fixing the fallback lands on an MCP server's spawn-time cwd, which
  is wrong in the same direction. Record it; do not touch it.
- **Template-First applies.** Any `.claude/` edit needs its mirror and
  `make build` in the same commit.
- **`moai spec audit` (CLI) must not change.** It is already correct. A change
  there would be scope creep on top of a working path.

## §C Pre-flight

1. `git rev-parse --show-toplevel` from this worktree returns the worktree path —
   this is the value callers will pass, so confirm it is what it should be.
2. `go build ./...` clean at the base commit, so a later failure is attributable.
3. `.moai/reports/t171/cause-confirmation.md` readable from this branch — the
   verdict table (M3) extends its call-site list.

## §D Constraints

- Additive and optional. No existing caller changes behaviour.
- No edit to `resolveProjectDir()`'s body.
- Both directions asserted on every new test — a check that only ever passes
  proves nothing, which is what let this defect sit unnoticed.
- Path validation rejects rather than falls back.

## §E Design decisions

### D1 — the parameter goes on the tool, not in the resolver

The caller is the only party that knows which tree it means. The MCP server
cannot infer it: its cwd is stale by construction and its environment names the
project. Pushing the answer down into `resolveProjectDir()` would require the
server to guess; taking it as an input requires the caller to state a fact it
already holds.

### D1b — `audit_multi` carries the parameter rather than being swapped out

REQ-6 exists to check the repair against the symptom that motivated it, and that
symptom happened on `audit_multi`. Pointing REQ-6 at `codex_audit` instead would
have been the smaller edit and would have made AC-6 a check of a different
surface under the same name. The cost is real — `backendCallFn` has no root
parameter, so the seam and its doubles widen — and it is stated in `spec.md` §2
so it can be traded away deliberately rather than by omission.

### D2 — `spec_progress` is included, and this is the one scope excess

Dispatched scope named `spec_audit`, `spec_drift`, and the codex `cwd`.
`spec_progress` is the SPEC scanner reached through the identical call, used by
the agents that author SPECs. Repairing the readers while leaving the scanner
blind would mean an auditor that sees a worktree SPEC and an author that does
not. One line, same test shape. Flagged in `spec.md` §2 for veto rather than
folded in quietly.

### D3 — reject a bad `project_root` instead of falling back

Stated as its own requirement because the alternative is actively harmful: a
caller that mistypes its worktree path would be silently returned to auditing the
primary — this SPEC's exact defect, arriving through the mechanism intended to
fix it, and reporting success while it happens.

### D4 — the verdict table is a deliverable, not a note

The undecided consumers are deferred, and a deferral is only cheap if the
next card does not have to re-derive the survey. The table's value is in the
"what evidence would settle this" column; without it the table is a list of
shrugs.

## §F Milestones

### M1 — `project_root` on the three SPEC tools

Input declaration, handler plumbing, validation. Ends with AC-1a, AC-2, AC-3 met
for `spec_audit`, `spec_drift`, `spec_progress`.

### M2 — `project_root` on the codex path and through `audit_multi`

`mcp_codex.go:1170` takes the same parameter with the same validation.
`audit_multi` accepts it and forwards it: `backendCallFn` (`mcp_convergence.go:353`)
widens to carry a root, the call site at `:485` passes it, and the injected test
doubles follow the signature. Ends with **AC-1b, AC-2, and AC-3** met there — the
parameter-present direction included, which is what iter-1 found missing.

### M3 — the call-site verdict table

Every `resolveProjectDir()` call site gets a row: artifact, correct scope,
evidence, disposition. The goal rows carry the suspected shared cause with the
two prior defects, marked unverified. Ends with AC-4 and AC-5.

### M4 — docs, mirrors, build

`moai-mcp-tools.md` records the parameter and states that the caller supplies
`git rev-parse --show-toplevel`; the two auditor agent bodies say to pass it.
Mirrors updated, `make build` clean.

### M5 — the post-repair check against the motivating symptom

Re-run `audit_multi` from a worktree with `project_root` set, record which tree
each backend read and whether the lane-9 symptoms recur. Ends with AC-6.

## §G Anti-patterns

- Fixing `resolveProjectDir()` "while we are in there". Five consumers with
  undecided verdicts ride on it — four MCP tools and the convergence state
  directory.
- A validation that falls back to the default on a bad path (D3).
- A test that only exercises the parameter-present path. AC-1a requires the
  absent-parameter direction too.
- The mirror of that mistake, which iter-1 actually caught: asserting only the
  absent and invalid directions on the codex path and never the present one.
  AC-1b exists because that is the surface the whole card came from.
- Changing the CLI's `--base-dir` or `baseDir` default. That path is correct.
- Writing the verdict table with "unknown" rows that do not say what would settle
  them.

## §H Cross-references

- `.moai/reports/t171/cause-confirmation.md` — the measurement
- `.claude/rules/moai/core/moai-mcp-tools.md` — the tool catalogue to update
- `.claude/rules/moai/workflow/worktree-integration.md` — the L1/L2 worktree
  tiers this defect surfaces in
