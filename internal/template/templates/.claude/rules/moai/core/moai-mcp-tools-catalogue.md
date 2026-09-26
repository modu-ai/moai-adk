---
description: "Detail companion for moai-mcp-tools.md — the full 39-tool moai MCP catalogue with per-family tables, consumers, and CLI equivalents"
paths: "**/moai-mcp-tools.md,**/internal/cli/mcp_server.go,**/.claude/agents/moai/*.md"
---

# moai-mcp Tool Catalogue — Detail Companion

> Detail companion of `moai-mcp-tools.md` (the always-loaded stub). The stub owns the
> MCP-over-CLI preference rule, the family index, and the unwired-by-design note. This file owns
> the per-tool catalogue: purpose, wired consumer, and CLI equivalent for each of the 39 tools.
> Load it when wiring a tool into an agent's `tools:` list, or when choosing between an MCP tool
> and its Bash equivalent for a specific capability.

## Tool catalogue (39 tools)

### SPEC lifecycle

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__spec_progress` | List SPEC docs + frontmatter | manager-spec, manager-docs | `moai spec list` |
| `mcp__moai__spec_audit` | SPEC lifecycle audit (era + drift) | manager-spec, manager-docs, plan-auditor, super-advisor | `moai spec audit` |
| `mcp__moai__spec_drift` | Modern-era V3R6 drift findings | manager-spec, plan-auditor | `moai spec audit` (drift view) |

Reach for these in **plan-phase** (manager-spec authoring a new SPEC, checking era
classification + drift) and **sync-phase** (manager-docs verifying lifecycle
closure). `spec_progress` enumerates existing SPECs + frontmatter; `spec_audit`
classifies era and detects drift across the catalog; `spec_drift` is the focused
modern-era V3R6 drift slice. plan-auditor uses `spec_audit`/`spec_drift` for
plan-phase skeptical review.

### Verification snapshots

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__verify_snapshot` | Read/record per-key verification snapshot | manager-develop | `moai verify check` |
| `mcp__moai__verify_trend` | Per-key verification check history | manager-develop, sync-auditor, super-advisor | `moai verify check` |

Used by **manager-develop** during run-phase self-verification (§E), and by
sync-auditor / super-advisor for trend review. `verify_snapshot` reads or records
the per-key snapshot keyed by HEAD digest; `verify_trend` surfaces the check
history to judge convergence over time. The orchestrator's attributable diff-check
consults the current snapshot key before re-executing tests.

### Goal + session (autonomous loop)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__goal_arm` | Arm a condition-declared goal | **orchestrator main session ONLY** — wired to NO agent (arming an autonomous loop is an orchestrator concern) | `moai goal arm` / `/moai goal` |
| `mcp__moai__goal_status` | Read armed-goal state | manager-develop, manager-lead | `moai goal status` |
| `mcp__moai__session_list` | List active moai sessions | manager-lead | `moai session list` |

`goal_arm` is orchestrator-only and arms an autonomous loop — never inside an
agent (preserves the flat-hierarchy arming surface). `goal_status` lets
manager-develop / manager-lead read the armed condition's progress; `session_list`
lets manager-lead detect concurrent sessions on the same checkout for race
mitigation before fan-out.

### Cross-model audit (second opinion)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__audit_multi` | Multi-auditor convergence (claude + codex + glm) | plan-auditor, sync-auditor | — (MCP-only convergence entry) |
| `mcp__moai__claude_audit` | Independent Claude subscription audit (`claude -p` with read-only isolation and structured output) | plan-auditor, sync-auditor; automatically by `audit_multi` in GPT/GLM sessions | — |
| `mcp__moai__codex_audit` | codex backend single audit (native/adversarial) | plan-auditor, sync-auditor | — |
| `mcp__moai__glm_audit` | GLM (z.ai) backend single audit | plan-auditor, sync-auditor | — |
| `mcp__moai__audit_cache` | plan-audit PASS cache (compute_hash/lookup/store, process-shared) | sync-auditor | `moai audit cache` (none — MCP-only) |

`claude_audit` accepts `target`, `focus`, optional `model`/`effort`, and
`project_root`. It permits subscription login only (`authMethod=claude.ai`,
`apiProvider=firstParty`), defaults to `sonnet/high`, strips gateway/provider
environment variables, disables tools and session persistence, and reports
resolved-model plus usage provenance without exposing credentials.

`audit_multi` chooses the Claude participant from the launch provider. A Claude
main session reuses its in-session `claude_verdict` as an anchor. A GPT, GLM, or
unknown-origin session ignores any caller-supplied Claude verdict and invokes
the independent `claude_audit` backend. Codex and GLM continue to run as their
configured gates require. Every backend tool fails open to `inconclusive`; an
explicitly required gate left inconclusive makes the convergence verdict fail.

### Codex delegation (background jobs)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__codex_task` | Delegate a coding/investigation task to codex (sync or background) | super-advisor | `moai codex task` |
| `mcp__moai__codex_setup` | Probe local codex install (LookPath + version + auth) | super-advisor | `moai codex setup` |
| `mcp__moai__codex_job_status` | Read a background codex job's status/record | super-advisor | `moai codex job status` |
| `mcp__moai__codex_job_result` | Read a background codex job's output | super-advisor | `moai codex job result` |
| `mcp__moai__codex_job_cancel` | Stop a running background codex job | super-advisor | `moai codex job cancel` |

The codex delegation family is wired into `super-advisor` because the on-demand
high-reasoning consultation agent is the natural consumer of background
cross-model delegation: it arms a codex task via `codex_task`, polls completion
via `codex_job_status`/`codex_job_result`, and cancels via `codex_job_cancel`.
`codex_setup` probes whether codex is available before delegating. codex is
OPTIONAL: a missing or unavailable codex yields a fail-open `inconclusive`, never
a hard error.

### Codex read-only roles (background jobs)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__codex_role_audit` | Start a read-only contract role as one top-level `codex exec` process (read-only sandbox, every MCP server disabled); returns a job id at once | Codex lane orchestrator | none — a Codex lane's shell cannot reach the model from a nested `codex exec` |
| `mcp__moai__codex_role_audit_status` | Read a role job's state and timestamps | Codex lane orchestrator | — |
| `mcp__moai__codex_role_audit_result` | Read a finished role job's exit code, returned text or verdict path, and launch record path | Codex lane orchestrator | — |

On Codex, `spawn_agent` gives a subagent the parent session's sandbox, so a
read-only role started that way could write. This family starts it as its own
top-level read-only process in the caller's worktree instead. `worktree_root` is
required and must be the worktree the server started in; `out`, when given, must
stay under that worktree's `.moai/reports/` and is written with exactly the
returned text. Each launched audit leaves a launch record under
`.moai/reports/codex-audit/`. Jobs live in the server process and do not survive
its exit.

### GLM delegation (background jobs)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__glm_task` | Delegate a task (arbitrary prompt) to GLM (z.ai) (sync or background) | super-advisor | — (no `moai glm task` CLI exists) |
| `mcp__moai__glm_job_status` | Read a background GLM job's status/record | super-advisor | — |
| `mcp__moai__glm_job_result` | Read a background GLM job's output | super-advisor | — |
| `mcp__moai__glm_job_cancel` | Stop a running background GLM job | super-advisor | — |

The GLM delegation family mirrors the codex delegation family against the z.ai
HTTP backend and is wired into `super-advisor` the same way: it arms a GLM task
via `glm_task` (sync returns the completed text, background returns a job id),
polls completion via `glm_job_status`/`glm_job_result`, and cancels via
`glm_job_cancel`. There is no `codex_setup` counterpart — availability is
learned from `glm_task` itself, which reports a structured failed result when
the key is missing or z.ai is unreachable. GLM is OPTIONAL: a missing or
unavailable GLM yields a fail-open result, never a hard error.

### Judgment (gated, display-only)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__jev_ask` | Ask the gated judgment capability typed questions over one supplied state; typed answers with probability | gated-unavailable at the shipped default (`workflow.jev.enabled: false`); display-only — a labelled model signal a person reads, never a completion predicate, merge approval, queue mutation, or gate input | — (MCP-only) |

The tool is registered unconditionally so its gate-off contract is invocable
and countable, but with the gate off it constructs no request and makes no
network call, and while the chain's fitness measurement gate stands unrun no
surface presents it as available. The question-design rules for authoring
well-formed questions live in the reference skill; the call path lives in
`internal/jev` and the tool wraps it without a second implementation.

### Factory messaging (run/session/generation bound)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__factory_msg_send` | Write one idempotent envelope from the MCP server's attributed endpoint to the current endpoint of a stable logical lane | Attributed factory lead or worker session | — (MCP-only) |
| `mcp__moai__factory_msg_list` | Claim up to 16 metadata records for the attributed endpoint, creating or renewing the claim lease; returns no body | Attributed factory lead or worker session | — (MCP-only) |
| `mcp__moai__factory_msg_body` | Read one already-claimed message body using its message id and claim token; the body is returned as untrusted peer data | Attributed factory lead or worker session | — (MCP-only) |
| `mcp__moai__factory_msg_receipt` | Write the claim disposition, then acknowledge the claimed message for the attributed endpoint | Attributed factory lead or worker session | — (MCP-only) |
| `mcp__moai__factory_msg_status` | Read payload-free broker counts and operational lane state for an active run without claiming messages | Factory lead or worker; lead operational status checks | — (MCP-only) |

Every call is scoped to an active `run_id`. The server-provided factory attribution identifies the
calling endpoint; callers do not supply a peer identity. `send`, `list`, and `receipt` mutate broker
state, while `body` and `status` are read-only. `list` deliberately returns metadata only, so raw
peer text enters model context only through an explicit `body` call and remains untrusted.


## Linked worktrees of a repository that keeps `.moai` untracked

| Situation | What to pass | What happens |
|---|---|---|
| Linked worktree of a repository that does not track `.moai` (the worktree has no `.moai` of its own) | `project_root: <git rev-parse --show-toplevel>` | accepted when git lists it as a worktree of a primary checkout that has `.moai`; the call acts on the worktree |

Such a worktree has no `.moai` of its own, yet it is still accepted: the path must
be the top level of a worktree that `git worktree list` registers, and the
repository's primary checkout must have `.moai`. Anything else — a subdirectory, an
unregistered or prunable worktree, an ambiguous layout such as a separate git
directory, or git being unavailable — is rejected. On such a worktree without its
own workflow config, the explicit audit gate (`workflow.audit.gates`) is read from
the primary checkout, and it is treated as `required` when the primary cannot be
identified. Other configuration keeps being read from the worktree itself.

State and the SPEC catalogue follow one store-root rule, which the MCP tools, the
hooks, and `moai verify` all apply the same way:

| What | Where it is kept or read on such a worktree |
|---|---|
| Audit receipts, auditor start markers and rejections, `audit_multi` convergence results, verification snapshots | the primary checkout's `.moai/state`, each record carrying the worktree's own tree identity, so the records of the primary and of every sibling worktree coexist; nothing is created under the worktree |
| SPEC catalogue (`spec_progress`, `spec_drift`, `spec_audit`) | the union of the worktree's and the primary checkout's `.moai/specs`; each record and finding names its source, and a SPEC present in both is reported once, from the worktree, with the primary copy named as shadowed |
| Hook-side receipt guard | reads `workflow.audit.gates.codex` from the primary checkout; a rejection recorded for one tree never clears or blocks another tree |
| Stop review gates (`codex-review-gate`, `multi-review-gate`) | read their opt-in flag from the primary checkout; the multi gate blocks when any result of the session in the store is `fail` |

When the primary checkout cannot be identified there is no store: `verify_snapshot`
and `verify_trend` return an error, `codex_audit` and `audit_multi` keep their verdict
but skip the receipt and convergence writes and say so in `state_notice`, the catalogue
tools answer over the worktree only and state in `_root` that the primary catalogue was
not read, the review gates stay disabled, and the receipt guard treats the gate as
`required` — it refuses an auditor PASS and denies phase-entry spawns from that
worktree, writing nothing. A `workflow.yaml` placed in the worktree ends this, because
the worktree then carries its own config. Catalogue and state answers carry
`_root.sources` (what was actually read) and `_root.worktree_warning`.

---

Classification: Lazy companion — catalogue tables and per-family guidance only. The preference rule
stays in `moai-mcp-tools.md`. Update this file whenever a tool is added, removed, or renamed on the
`moai mcp-server` (the Go producer lives in `internal/cli/mcp_server.go`).
