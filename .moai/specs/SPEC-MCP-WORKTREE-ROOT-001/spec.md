---
id: SPEC-MCP-WORKTREE-ROOT-001
title: "Let an MCP caller name its own tree, so a worktree SPEC stops being invisible"
version: "0.4.0"
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
| `codex_audit` | `handleCodexAudit`, `mcp_codex.go:1170` | the backend reviews a tree; the wrong tree is what lane-9 saw |
| `audit_multi` | `mcp_server.go:367`, fan-out `mcp_convergence.go:485` | the surface lane-9 actually used; carries the parameter through to its backends |

**`codex_task` is not in scope.** Row four names `codex_audit` — the tool whose
handler builds the `cwd` at `mcp_codex.go:1170`. `codex_task` delegates a coding
task rather than reviewing a tree, so whether its own root resolution is wrong is
a question this SPEC does not answer; it belongs to REQ-4's survey. An earlier
draft wrote the two names together in one cell, which left the five-tool count
resolvable only by following the line citation.

**Note on `audit_multi` — carrier, and it costs a signature change.** REQ-6 checks
the repair against the symptom that motivated it, and that symptom occurred on
`audit_multi`, not on `codex_audit`. Re-aiming REQ-6 at `codex_audit` — which does
receive the parameter — would have been the cheaper edit, but it would make AC-6 a
check of a different surface wearing the same name. So `audit_multi` accepts
`project_root` and forwards it.

The cost is not one line. Its fan-out seam is
`backendCallFn(ctx, backend, target, focus string)` (`mcp_convergence.go:353`,
invoked at `:485`), which carries no root, so the parameter requires widening that
signature and its injected test doubles. `performCodexAudit`'s own signature
widens with it, since the root has to reach the params map it builds.

**One of those touches is load-bearing and must not be done carelessly.** The
comments at `mcp_convergence.go:348-352` and `:480-483` state an independence
invariant — the seam takes `(ctx, backend, target, focus)` **and nothing else**,
so that `claude_verdict` is *structurally* forbidden from reaching a secondary
backend, with the compiler as the enforcement. Adding a root parameter makes the
literal wording false, so both comments are reworded in the same change: the
invariant they protect is that `claude_verdict` never reaches a backend, not that
the parameter list never grows. Widening the signature without rewording them
leaves a comment asserting a guarantee the code no longer expresses — the shape
in which a real invariant quietly stops being enforced.

That is still contained — one seam, one call site, two comments, the doubles that
implement the signature — but it is more than the SPEC-tool edits, and it lands on
the file-count axis the tier note below already flags. Stated here so the
alternative can be taken instead if the cost is judged wrong.

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
- **The undecided consumers — four MCP tools and the convergence state
  directory**: `goal_arm` / `goal_status` (`mcp_server.go:466`, `:483`),
  `verify_snapshot` / `verify_trend` (`:539`, `:569`), and the convergence state
  directory (`mcp_convergence.go:561`). They are surveyed, not repaired.
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

**Both lists above are illustrative, not exhaustive.** The MCP-path set also holds
at least `mcp_glm.go` (`projectDirResolver`, the GLM `llm.yaml` location) and
`mcp_server.go:105` (the tool-enablement read), and the CLI-side list is longer
than the three files named — those three hold five call sites between them.
Nothing escapes the SPEC on that account: REQ-4's mandate is
grep-derived, so the full set is re-derived at M3 rather than taken from this
prose. The prose names the consumers the reasoning above turned on.

## §3 Requirements and acceptance criteria

Tier S, so criteria sit inline. Six requirements, seven criteria — REQ-1
carries two, one per direction of the redirect.

### REQ-1 — the caller can name its tree

The five in-scope tools SHALL accept an optional string input `project_root`.
When it is non-empty, the handler SHALL use it as the project root instead of
`resolveProjectDir()`.

**AC-1a** — Given an MCP `spec_audit` call carrying `project_root` set to a
worktree that holds a SPEC absent from the primary checkout, when the call
returns, then that SPEC appears in the result. Given the same call without the
parameter, then it does not. **Both directions are asserted**: a test that only
proves the new path works would pass equally if the parameter were ignored and
the primary happened to contain the SPEC.

**AC-1b** — Given a codex audit call carrying `project_root` naming a worktree,
when the review parameters are constructed, then the `cwd` they carry is that
path. This is asserted on the `params` map of **both** codex paths, with no live
backend:

- the single-backend path, `mcp_codex.go:1167-1171`, where `cwd` is today
  `resolveProjectDir()`;
- the `audit_multi` fan-out path, `performCodexAudit` (`mcp_convergence.go:387-394`),
  whose params map carries **no `cwd` key at all** today.

**Why the second bullet is named separately, and why the seam is not enough.**
An earlier wording asked only that the root "reach the backend fan-out". A
natural test of that — a `backendCall` double asserting what it received — is
satisfied while `performCodexAudit` is never entered, and that function is where
the codex params are actually built. The root could stop at the seam with all
seven criteria green, contradicting §2's claim that `audit_multi` "carries the
parameter through to its backends". Pinning the assertion to the params map
closes that gap.

The second bullet also changes what M2 has to build: `performCodexAudit` does not
merely pass the wrong `cwd`, it passes none, so the repair adds the key rather
than redirecting an existing one. The two codex paths have diverged, and this is
the SPEC's record of it.

**Why this is separate from AC-1a**: the codex `cwd` is the surface lane-9's
symptom appeared on and the surface AC-6's post-repair check reads, yet without
this criterion nothing in the SPEC asserted its parameter-present direction —
only the absent and invalid ones. A redirect nobody checks is a redirect nobody
knows happened.

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
- `internal/cli/mcp_server.go:526`, `:583`, `:597` — the three SPEC-tool handlers
- `internal/cli/mcp_server.go:367` — `audit_multi` registration, whose declared
  inputs carry no `project_root` today
- `internal/cli/mcp_codex.go:1167-1171` — the single-backend codex params, where
  `cwd` is `resolveProjectDir()`
- `internal/cli/mcp_convergence.go:353`, `:485` — `backendCallFn` and its call
  site, the seam the root has to cross
- `internal/cli/mcp_convergence.go:387-394` — `performCodexAudit`, whose params
  map carries no `cwd` key at all
- `internal/cli/mcp_convergence.go:348-352`, `:480-483` — the independence-invariant
  comments the seam widening makes literally false
- `internal/cli/mcp_codex.go:1152` — `handleCodexAudit`, which owns `:1170` and is
  why row four of §2 names `codex_audit` rather than `codex_task`

## §4.1 Tier note

Tier S is claimed on the LOC and requirement/criterion axes, both of which fit
comfortably: the core change is three Go files plus tests, six requirements, seven
criteria. The **file-count** axis does not fit as cleanly — counting template
mirrors and the two auditor agent bodies, the touch surface is roughly eight to
ten paths against the tier's "< 5 files" guidance, and `audit_multi`'s seam
widening adds to it.

The overrun is mirror bookkeeping and one signature, not hidden complexity, so
the tier holds. What would flip it is the follow-up scope arriving early: if the
per-call-site verdicts (REQ-4) turn into repairs inside this card, the file axis
and the LOC axis move together and Tier M is the honest call. Recorded so the
decision is visible rather than assumed.

## §5 Constraints

- Backward compatibility is not negotiable: the parameter is additive and
  optional, and every existing caller keeps working untouched.
- No change to `resolveProjectDir()`'s body in this SPEC. Its other consumers are
  surveyed here and repaired elsewhere.
- The template mirror and `make build` apply to any changed
  `.claude/rules/.../moai-mcp-tools.md` or agent body, per Template-First.

## §6 HISTORY

- 0.4.0 (2026-08-22) — plan-audit iter-3's two optional findings applied rather
  than deferred, because both bite in run-phase. R3: the scope table's fourth row
  named `codex_task` where the operative surface is `codex_audit`
  (`handleCodexAudit` owns `mcp_codex.go:1170`) — a wrong tool name in the scope
  table would have sent the run at the wrong handler. R2: the cost note omitted
  `performCodexAudit`'s own signature and, more importantly, the two independence
  -invariant comments the seam widening makes literally false; they are now named
  as part of the change, with the invariant they protect (`claude_verdict` never
  reaches a backend) stated so the reword preserves it.
- 0.3.0 (2026-08-22) — plan-audit iter-2 finding R1 applied, plus two items
  found by a self-sweep while iter-2 was running. R1: AC-1b asked only that the
  root reach the backend fan-out, which a `backendCall` double satisfies without
  ever entering the function that builds the codex params — the root could stop
  at the seam with every criterion green. AC-1b now pins both codex params maps,
  and in doing so records that the two paths have diverged: the fan-out path
  passes no `cwd` at all, so M2 adds the key rather than redirecting one. The
  self-sweep items: §4's evidence list still named only the pre-D1 surfaces, and
  a "three named" count mixed files with call sites.
- 0.2.0 (2026-08-22) — plan-audit iter-1 findings applied. D1: `audit_multi`
  joins REQ-1 as a forwarding carrier, so AC-6 checks the surface the symptom
  actually occurred on; the seam-widening cost is stated rather than hidden. D2:
  AC-1 splits into 1a/1b so the codex `cwd` parameter-present direction is
  asserted — it was the one direction nothing checked. D3/D4/D5: consumer count
  corrected, both consumer lists marked illustrative with the grep mandate named
  as the exhaustive source, and the tier call recorded with what would flip it.
- 0.1.0 (2026-08-22) — initial draft. Scope set to Tier S by the lead after cause
  confirmation falsified the card's first premise and re-attributed its second.
