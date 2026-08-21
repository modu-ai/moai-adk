---
id: SPEC-MCP-WORKTREE-ROOT-001
title: "Let an MCP caller name its own tree, so a worktree SPEC stops being invisible"
version: "0.1.0"
status: draft
created: 2026-08-22
updated: 2026-08-22
author: lane-7
priority: P1
phase: "v3.1.3"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "mcp, worktree, audit, project-root, spec-scanner"
---

# SPEC-MCP-WORKTREE-ROOT-001

## §1 Problem

Every SPEC written in Factory or Kanban mode is authored inside a worktree. The
auditors that are supposed to read those SPECs reach them through the MCP tools.
Those tools cannot see the worktree.

The mechanism is one function. `internal/cli/session.go:242`:

```go
func resolveProjectDir() string {
    if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" { return dir }
    if cwd, err := os.Getwd(); err == nil { return cwd }
    return ""
}
```

Claude Code sets `CLAUDE_PROJECT_DIR` to the **project** root, not to the
worktree the session is working in. Measured across four live `moai mcp-server`
processes in two repositories — every server whose cwd was a worktree carried the
primary checkout in that variable:

| MCP server cwd | `CLAUDE_PROJECT_DIR` |
|---|---|
| `moai-adk-go/.claude/worktrees/t165` | `/Users/goos/MoAI/moai-adk-go` |
| `moai-adk-go/.claude/worktrees/t89` | `/Users/goos/MoAI/moai-adk-go` |
| `mo.ai.kr/.claude/worktrees/t134` | `/Users/goos/MoAI/mo.ai.kr` |

The env branch wins, so the cwd is never consulted and the outcome is
deterministic: **an MCP audit issued from a worktree session audits the primary
checkout.** A SPEC that exists only on the card's branch is not in the catalogue
the auditor reads, so it is neither audited nor reported missing — it is absent,
silently.

Two things this problem is NOT, both checked:

- **It is not the audit engine.** `internal/spec/audit.go` defaults `baseDir` to
  a plain `"."` and reads the tree it runs in. A probe SPEC created only in a
  worktree was seen by `moai spec audit` run there. The CLI has been correct all
  along; only the MCP path is affected.
- **It is not a stale-cwd bug**, although one is hiding behind it. An MCP server
  is a long-lived subprocess and cannot follow `EnterWorktree`, so its cwd goes
  stale on a worktree switch. That branch is unreachable today because the
  environment variable is always set, but a repair that removed only the env
  branch would land on the stale cwd instead of on the right answer.

Full measurement: `.moai/reports/t171/cause-confirmation.md`.

## §2 Scope

**In scope — give the caller a way to say which tree it means.** The MCP tools
whose subject is the caller's working tree accept an optional `project_root`; the
calling agent passes its own `git rev-parse --show-toplevel`. Absent or empty,
behaviour is exactly what it is today.

| Tool | Handler | Why it is in scope |
|------|---------|--------------------|
| `spec_audit` | `mcp_server.go:583` | the reported defect |
| `spec_drift` | `mcp_server.go:597` | same call, same blindness |
| `spec_progress` | `mcp_server.go:526` | see the note below |
| `codex_task` / codex audit `cwd` | `mcp_codex.go:1170` | the backend reviews a tree; the wrong tree is what lane-9 saw |

**Note on `spec_progress` — added beyond the dispatched scope, flagged for veto.**
The scoping decision named the audit family as `spec_audit`, `spec_drift`, and the
codex `cwd`. `spec_progress` is `spec.ListDocs` — the SPEC scanner itself, reached
through the identical `resolveProjectDir()` call, and it is what `manager-spec`
and `manager-docs` use to enumerate SPECs. Its blindness is the same defect by
construction, and the fix is the same one line. Leaving it out would ship an audit
surface that is half repaired: the auditor would see the worktree's SPEC while the
authoring agents still would not. It is included on that reasoning, and it is the
one place this SPEC exceeds what was dispatched — drop it if that call is wrong.

### 2.1 Out of Scope

- **Changing `resolveProjectDir()` itself.** Its other consumers are not all
  wrong, and a shared-function change would move them all at once. REQ-4 produces
  the per-call-site verdicts instead; the repairs they imply belong to a
  follow-up card.
- **The four undecided consumers** — `goal_arm` / `goal_status`
  (`mcp_server.go:466`, `:483`), `verify_snapshot` / `verify_trend` (`:539`,
  `:569`), and the convergence state directory (`mcp_convergence.go:561`). They
  are surveyed, not repaired.
- **Non-MCP consumers.** `goal.go:159`/`:188`, `memory.go:165`/`:287`, and
  `launcher_blockcap_infinite.go:139` run as ordinary CLI invocations in the
  session's own tree, where the same env variable is what the operator expects.
  Nothing measured here shows a defect in them.
- **The stale-cwd fallback.** Unreachable while the env variable is set. Recorded
  in the verdict table as a hazard the follow-up must not step on; not fixed here.
- **Reproducing the lane-9 backend symptoms.** A codex `verdict` contradicting its
  own summary and a hallucinated GLM REST API are consistent with reading the
  wrong tree, but the causal chain was never demonstrated. AC-6 turns this into a
  post-repair check rather than a claim.

## §3 Requirements and acceptance criteria

Tier S, so criteria sit inline. Six requirements, six criteria.

### REQ-1 — the caller can name its tree

The four in-scope tools SHALL accept an optional string input `project_root`.
When it is non-empty, the handler SHALL use it as the project root instead of
`resolveProjectDir()`.

**AC-1** — Given an MCP `spec_audit` call carrying `project_root` set to a
worktree that holds a SPEC absent from the primary checkout, when the call
returns, then that SPEC appears in the result. Given the same call without the
parameter, then it does not. **Both directions are asserted**: a test that only
proves the new path works would pass equally if the parameter were ignored and
the primary happened to contain the SPEC.

### REQ-2 — silence means today's behaviour

An absent or empty `project_root` SHALL resolve exactly as the tool resolves it
now, through `resolveProjectDir()`, with no change in output.

**AC-2** — Given each in-scope tool called with no `project_root`, when its result
is compared against the same call before this change, then the results match. A
caller that never learns about the parameter sees nothing different.

### REQ-3 — a named root is validated before it is used

A non-empty `project_root` SHALL be rejected unless it resolves to an existing
directory containing a `.moai` directory. The rejection SHALL be a tool error
naming the path, not a silent fallback to the default.

**AC-3** — Given `project_root` pointing at a non-existent path, a file, or a
directory with no `.moai`, when the tool runs, then it returns an error naming
the offending path and does not audit anything. **Why an error rather than a
fallback**: a silent fallback would send a caller that mistyped its own worktree
path back to auditing the primary — reproducing this SPEC's defect through the
mechanism meant to fix it, and reporting success while doing it.

### REQ-4 — the undecided consumers are surveyed, not guessed

A verdict table SHALL be committed under `.moai/reports/t171/` covering every
`resolveProjectDir()` call site, recording for each: the artifact it resolves,
whether project scope or tree scope is correct for it, the evidence for that
reading, and whether it is repaired here or deferred.

**AC-4** — Given the table, when each `resolveProjectDir()` call site found by
grep is looked up, then it has a row, and every row deferred to a follow-up
states what evidence a decision would need. A row reading "unknown" with no
statement of what would settle it fails this criterion — that is a gap wearing a
verdict's clothes.

### REQ-5 — the goal-state rows record a suspected shared cause

The rows for `goal_arm` and `goal_status` SHALL note that two previously recorded
defects — goal keying being unreliable under worktrees, and the multi-session
arm parser — may share this seam, and SHALL state that the connection is
suspected rather than established.

**AC-5** — Given those rows, when read, then the connection is present and marked
as unverified. **Why this is its own requirement**: the follow-up card's cheapest
possible start is knowing that three symptoms may have one cause; the expensive
failure is three cards each fixing a third of it.

### REQ-6 — the repair is checked against the symptom that motivated it

After the repair, `audit_multi` SHALL be re-run from a worktree with
`project_root` naming that worktree, and the outcome recorded — whether the
backends read the intended tree, and whether the lane-9 symptoms recur.

**AC-6** — Given that run, when its output is recorded under `.moai/reports/t171/`,
then it states which tree each backend read and whether the self-contradicting
verdict and the hallucinated API reappeared. **A recurrence does not fail this
criterion** — it would show the symptoms have a second cause, which is a finding
worth having. What fails it is not looking.

## §4 Evidence

- `.moai/reports/t171/cause-confirmation.md` — the measurement this SPEC rests on
- `internal/cli/session.go:242` — `resolveProjectDir()`
- `internal/spec/audit.go:157` — the `baseDir = "."` default that clears the CLI
- `internal/cli/mcp_server.go:526`, `:583`, `:597`; `internal/cli/mcp_codex.go:1170`

## §5 Constraints

- Backward compatibility is not negotiable: the parameter is additive and
  optional, and every existing caller keeps working untouched.
- No change to `resolveProjectDir()`'s body in this SPEC. Its other consumers are
  surveyed here and repaired elsewhere.
- The template mirror and `make build` apply to any changed
  `.claude/rules/.../moai-mcp-tools.md` or agent body, per Template-First.

## §6 HISTORY

- 0.1.0 (2026-08-22) — initial draft. Scope set to Tier S by the lead after cause
  confirmation falsified the card's first premise and re-attributed its second.
