---
id: SPEC-CODEX-PHASE2-001
title: "Codex Phase 2 — task delegation + job lifecycle tools over the delivered app-server JSON-RPC session client"
version: "0.5.0"
status: in-progress
created: 2026-08-10
updated: 2026-08-10
author: manager-spec
priority: P2
phase: "v3.2 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, mcp, json-rpc, job-lifecycle, fail-open, model-effort-ssot"
depends_on: [SPEC-MOAI-MCP-SERVER-001]
---

# SPEC-CODEX-PHASE2-001 — Codex Phase 2 MCP tool surface

## HISTORY

- 2026-08-10 (run-phase, M0 closure, v0.5.0) — **M0 executed**; its record is in `progress.md` §E.2, read from the protocol JSON Schema that `codex app-server generate-json-schema --out <dir>` emits against the pinned `codex-cli 0.146.1`, not from an inferred handshake. All four probe items came back **present** — `turn/interrupt` exists, `TurnStartParams` requires `threadId` so a second turn on an existing thread is the method's normal shape, `model` and `effort` are both turn fields, and `sandboxPolicy` is expressible per turn (`readOnly` | `workspaceWrite` | `dangerFullAccess` | `externalSandbox`). **No `plan.md` §D M0 item-4 degrade path fires**, and REQ-CX2-011 / AC-CX2-014 stand. Two observations nonetheless forced amendments, both applied here; no requirement and no acceptance criterion was added or removed (the set remains REQ-CX2-001..015 and AC-CX2-001..016). **Amendment 1 — `turnId`.** `turn/interrupt` params are `{threadId, turnId}` with **both required**; REQ-CX2-003's record shape carried no `turnId` (the token appeared zero times across all four artifacts), which made REQ-CX2-011 unimplementable — the record could not supply the second required argument. The value is obtainable from the `turn/started` notification, whose `Turn` has a required `id`. REQ-CX2-003 now records `turnId` and names that source, REQ-CX2-011 names both arguments it sends, and AC-CX2-005 / AC-CX2-014 decode and assert it. **Amendment 2 — sticky `sandboxPolicy`.** Both `effort` and `sandboxPolicy` are documented verbatim as overriding "for this turn **and subsequent turns**" — they are sticky on the thread, not turn-scoped. Under the recorded in-process reusable-session decision plus `resume_last` thread reuse (REQ-CX2-008), a task that opted into writes would leave its thread write-enabled for a later task that did not: a route around the `workflow.codex.task.allow_write` gate, whose check happens at request time while the effect outlives the request. REQ-CX2-007 now requires `sandboxPolicy` to be transmitted explicitly on every turn (`readOnly` when not opted in) rather than relying on a per-turn default or inherited thread state, AC-CX2-010 gained a two-turn reused-thread arm asserting exactly that, and `plan.md` M3 carries the hazard. `effort`'s identical stickiness was **considered and deliberately left untreated**: REQ-CX2-002 already requires the resolved value to be carried "into the params actually transmitted to codex" and resolution is per request, so every turn carries a freshly-resolved effort; and inherited effort is a cost/quality drift rather than a crossing of a safety boundary, so it does not earn a MUST criterion against the 16-AC ceiling.
- 2026-08-10 (plan-phase, iter-4, v0.4.0) — Advisory polish applied after plan-audit iteration 2 returned **PASS at 0.84** against the Tier M threshold 0.80 (report: `.moai/reports/plan-audit/SPEC-CODEX-PHASE2-001-review-2.md`). The PASS was recorded against v0.3.0; nothing here changes the substance the auditor scored, and no requirement or acceptance criterion was added or removed. Five of the six advisory findings were taken. **N1**: AC-CX2-016's read-only-hint assertion is restated against the *value* of `*Annotations.ReadOnlyHint` rather than its presence — `mcp.NewTool` seeds every tool with `ToBoolPtr(false)` (mcp-go v0.57.0 `mcp/tools.go:856`), so a nil-check would have false-failed a correctly-registered write tool. **N2**: the same AC's schema assertion is restated as per-tool declared properties, because `NewTool` always seeds a non-empty `InputSchema` and the original clause therefore could not fail. **N3**: AC-CX2-010's `codex_setup` arm now states literal expected values per config state instead of deferring to "the gate's own read", removing a self-referential oracle. **N5**: `progress.md`'s stale v0.2.0 line was corrected. **N6**: degrade path (b) now states why AC-CX2-002 is deliberately excluded from REQ-CX2-001's amended AC set, so the partial set is not mistaken for the id-arithmetic error D1 fixed. **N4** (degrade paths (c)/(d) under-specifying their knock-on collapse) was **not** taken — M0's closure criterion already requires the real amendment to land before M1 starts, which makes the degrade text guidance rather than an exhaustive script.
- 2026-08-10 (plan-phase, iter-3, v0.3.0) — Closes the four MUST-FIX findings of plan-audit iteration 1 (verdict FAIL, 0.73 against the Tier M 0.80 threshold). **D1**: `plan.md` §D M0's interrupt degrade path pointed at AC-CX2-011 (which verifies REQ-CX2-008 `resume_last`); it now points at AC-CX2-014, the AC actually bound to REQ-CX2-011 per the `acceptance.md` §D table, and states the degraded form explicitly. Degrade paths for probe items (b)-(d) were added alongside it. **D2**: the three previously-uncheckable normative clauses gained binary checks — `codex_setup`'s `allow_write` field (a Go-test arm on AC-CX2-010), REQ-CX2-015's typed-field/default placement, and REQ-CX2-013's per-tool JSON Schema + read-only hint (both in AC-CX2-016's batch). **D3**: M0 gained an explicit closure criterion, a pinned probe version (`codex-cli 0.146.1`), and a probe-record destination (`progress.md` §E.2 — no `research.md`, which would breach the Tier M artifact set), plus a matching DoD bullet. **D4**: AC-CX2-016's registration check replaced a line-counting `grep -c` alternation, which could pass with three of four tools missing, with a per-tool existence loop. **S2**: §A.2's "zero matches over Markdown" claim was narrowed to zero *implementation* matches, since three tokens legitimately appear in the predecessor's out-of-scope prose. `REQ-CX2-007`'s `codex_setup` inspectability clause was **kept** (the write-mode fork chose the config key over an env var because the key is visible to `codex_setup`) and is now claimed by M3 and checked by AC-CX2-010. No requirement and no acceptance criterion was added or removed: the set remains REQ-CX2-001..015 and AC-CX2-001..016.
- 2026-08-10 (plan-phase, iter-2, v0.2.0) — The two open `plan.md` §D M0 design forks were resolved by user decision and recorded as decisions; M0 now carries the protocol probe alone. **Background job execution model** = (i) an in-process goroutine inside the long-lived `moai mcp-server` process, accepting the loss of every in-flight job on server exit and deferring detached-subprocess durability to a later upgrade. **Write-mode opt-in surface** = the config key `workflow.codex.task.allow_write`, following the existing `workflow.codex.review_gate.enabled` shape, read fail-closed with a distributed default of `false` and inspectable via `codex_setup`. REQ-CX2-003 / REQ-CX2-007 / REQ-CX2-015 and risk R2 were amended for consistency with those two decisions; no requirement was added or removed (the set remains REQ-CX2-001..015).
- 2026-08-10 (plan-phase, iter-1) — Initial Tier M authoring. Picks up the follow-up SPEC that `SPEC-MOAI-MCP-SERVER-001` (status: `completed`) declared in its `spec.md:47` and `design.md:81` Out-of-Scope entries. **The declared deferral list is stale**: one of its four items (the native Go JSON-RPC client for `codex app-server`) was substantially delivered afterwards by PR #1430 (`d39e3cdc6`, `fix(codex-gate): speak the real app-server JSON-RPC protocol so the gate can actually BLOCK`). §A.1 below records the verified delivered/remaining split against `origin/main`; this SPEC scopes only the genuinely-missing remainder.

## §A. Verified baseline — what already exists

Every claim in this section was read from `origin/main` (`e55b48def`), not from the predecessor SPEC's prose. The predecessor's deferral list is treated as a hypothesis, and it did not survive contact with the tree.

### §A.1 Delivered — the native app-server JSON-RPC session client

`internal/cli/mcp_codex.go` on `origin/main` (730 lines) already carries a real, session-oriented JSON-RPC client for `codex app-server`. It is not a stub and it is not a shellout-with-one-line-of-JSON:

| Capability | Location (`internal/cli/mcp_codex.go`, `origin/main`) |
|---|---|
| Session-spawn seam + production runner (stdin/stdout pipes, stderr discarded) | `codexSessionRunner` / `realCodexSessionRunner.start` — L188-216 |
| NDJSON connection with a concurrent reader goroutine (8 MB scanner buffer) | `realCodexConn.readLoop` — L229-242 |
| `send` / `recv` / `close` (stdin close, 3 s wait, then kill) | L244-269 |
| Request framing | `writeCodexRequest` — L362-369 |
| id-matched response await, JSON-RPC `error` arm surfaced verbatim | `awaitCodexResponse` — L374-395; `codexIDMatches` (int **and** string ids) — L398-411 |
| Mandatory handshake: `initialize` (with `clientInfo`) → `thread/start` → `threadId` | `runCodexReviewRPC` L313-340; `extractThreadID` L414-427 |
| Protocol-correct param shaping (`threadId` injection; internally-tagged `target`; `turn/start` `input[]` wrapping) | `buildCodexReviewParams` L433-446; `coerceCodexReviewTarget` L453-463 |
| Asynchronous notification stream read until `turn/completed` | `awaitCodexTurnReview` L468-508 |
| Verdict synthesis from codex's free-form review prose | `codexFindingBullet` L536; `synthesizeReviewOutput` L543-554 |

The transport, the handshake, the framing, the id matching, the error arm, and the async turn loop are therefore **NOT in scope** for this SPEC. Re-specifying them would be a defect.

### §A.2 Not delivered — verified by absence

`git grep` over `origin/main` returns **zero implementation matches** — no Go source, no registration site, no config or JSON surface — for every one of: `codex_task`, `codex_job_status`, `codex_job_result`, `codex_job_cancel`, `codex_transfer`, `turn/interrupt`, `resume_last`, `.moai/state/codex-jobs`. Three of those tokens (`codex_task`, `codex_job_status`, `codex_transfer`) do appear in Markdown, but only inside the predecessor's own out-of-scope prose (`SPEC-MOAI-MCP-SERVER-001/spec.md:47`, `design.md:81`) — the very lines this SPEC cites in §I as the deferral it picks up. Prose naming a deferred tool is not an implementation of it. Tool registration in `internal/cli/mcp_server.go` `registerMoaiMCPTools` (L105-242) carries exactly two codex tools — `codex_audit` (L191-199) and `codex_setup` (L204-208).

### §A.3 Three verified structural gaps in the delivered client

These are the parts of the client that a task/job surface actually needs and that the review-gate use case never exercised:

- **G1 — the session is single-shot and synchronous.** `runCodexReviewRPC` (L306) opens a session, drives one turn, and `defer conn.close()` (L311) tears the subprocess down on return. The `threadId` obtained at L337 is a local variable, discarded on return. There is no way to issue a second turn on the same thread, and no way to start a turn that outlives the call.
- **G2 — there is no cancellation path.** The only bound is the caller's `context` deadline (the Stop hook pins `config.DefaultCodexReviewGateTimeout` = 900 s, `internal/config/defaults.go:264`, applied at `internal/cli/codex_review_gate.go:87`). No `turn/interrupt` method constant exists, and `close()` is a blunt stdin-close-then-kill.
- **G3 — the `model` argument is inert.** `codex_audit` declares a `model` parameter (`mcp_server.go:197`) and `handleCodexAudit` reads it into `params["model"]` (`mcp_codex.go:586`, `:596`), but `buildCodexReviewParams` (L433) constructs a **fresh** map containing only `threadId` plus `target`/`input` — the model value is dropped and never reaches codex. The codex path also never calls the model/effort SSOT, unlike the GLM sibling which resolves through `template.ResolveAgentModelEffort` at `internal/cli/mcp_glm.go:134`. The existing guard `TestMCPAudit_NoDirectFrontmatterRead` (`internal/cli/mcp_audit_test.go:146`) passes on `mcp_codex.go` only because that file reads no frontmatter at all — passing by abstinence, not by wiring.

## §B. User Story

**As a** MoAI orchestrator (or an external MCP host) that wants to delegate a bounded coding or investigation task to codex and keep working while it runs,

**I want** a `codex_task` tool that drives a codex turn with my prompt, plus a job surface (`codex_job_status` / `codex_job_result` / `codex_job_cancel`) backed by a durable per-job record,

**so that** the `/codex:rescue` capability is reachable declaratively from the same local stdio server as `codex_audit`, a long task does not block the calling turn, and a task that goes wrong can be stopped rather than waited out.

## §C. Scope Summary

This SPEC delivers four new MCP tools and the session/job plumbing they require, on top of the client verified in §A.1. It closes G1 (thread reuse + detached turns), G2 (`turn/interrupt` cancellation), and G3 (model/effort SSOT wiring), because each is a precondition of a usable task surface rather than a separable cleanup.

### Out of Scope — `codex_transfer`

- `codex_transfer` is **excluded**, and this is a deliberate reversal of the expectation set by the predecessor's deferral list rather than an oversight.
- The predecessor already deferred it twice on the grounds that it is undocumented and version-coupled (`SPEC-MOAI-MCP-SERVER-001/design.md:81`), and the tree carries zero references to it (§A.2) — there is no method name, no param shape, and no observed response to specify against.
- PR #1430 is the evidence that specifying against an unobserved codex method is expensive: four protocol facts (mandatory `initialize`, required `threadId`, internally-tagged `target`, async NDJSON completion) had to be discovered by hand-probing codex-cli 0.146.1 because the documented surface was wrong. Writing requirements for `codex_transfer` today would be an unverified-premise claim of exactly that shape.
- It blocks nothing: the `/codex:rescue` replacement is `codex_task`, which is in scope here.
- **Re-entry condition**: a `codex_transfer` SPEC becomes writable once the method, its params, and its result are observed against a pinned codex-cli version and recorded in a research artifact.

### Out of Scope — structured findings extraction

- `synthesizeReviewOutput` (`mcp_codex.go:543`) always returns `Findings: []Finding{}` and derives a coarse `pass`/`fail` from a finding-bullet regex. Populating the `Finding` array (severity / file / line / confidence) and replacing the regex heuristic is an **audit-quality** improvement to the already-shipped `codex_audit` surface, not a task-surface requirement.
- It is left to a follow-up SPEC so that this one does not carry a second, independently-riskier verdict-parsing change alongside the job lifecycle.

### Out of Scope — CLI mirrors of the new tools

- No `moai codex task` / `moai codex jobs` Cobra subcommand is added. The tools are MCP-only in this SPEC, matching the `verify_snapshot` / `verify_trend` MCP-first precedent recorded as DQ-1 in the predecessor's `design.md:235`.

### Out of Scope — convergence and gate behavior

- `audit_multi` convergence (`internal/cli/mcp_convergence.go`), the `audit_model` / `audit_gate` config semantics, and the Stop-hook review gate's ALLOW/BLOCK contract (`internal/cli/codex_review_gate.go:50-109`) are untouched. The gate keeps calling `runCodexReviewRPC` and keeps its fail-open behavior.

### Out of Scope — new template or distributed config surface

- No file is added or modified under `internal/template/templates/`. The new tools introduce no `.claude/` or `.moai/` template artifact; job records live in `.moai/state/`, which is runtime state and already gitignored (`.gitignore:207`, `:275`). Thresholds land in `internal/config/defaults.go` and env names, if any, in `internal/config/envkeys.go` — both Go, neither a template surface. Template-First mirroring and §25 neutrality therefore do not apply to this SPEC.

## §D. Requirements (GEARS)

> Domain prefix `REQ-CX2-NNN`. Every requirement resting on existing code cites the verified `origin/main` location from §A.

### M1 — Reusable session + model/effort SSOT (closes G1 partially, G3)

**REQ-CX2-001** (Ubiquitous) The codex session client shall expose a reusable session handle that retains the `threadId` obtained from `thread/start`, so that a second turn can be issued on the same thread without repeating the `initialize` + `thread/start` handshake.

**REQ-CX2-002** (State-driven) **While** resolving the codex model and effort for a request, the codex backend shall resolve through `template.ResolveAgentModelEffort` — the same interpreter the GLM sibling uses at `internal/cli/mcp_glm.go:134` — and shall carry the resolved value into the params actually transmitted to codex, and shall not read agent frontmatter or `llm.agent_overrides` directly.

### M2 — Job registry (durable record)

**REQ-CX2-003** (Event-driven) **When** a codex task is started in background mode, the server shall create a job record at `.moai/state/codex-jobs/<job-id>.json` carrying at minimum: job id, status, creation and last-update timestamps, the codex `threadId`, the codex `turnId` of the turn the job is driving, the pid of the codex process this server spawned for the job, the requested mode, and a summary of the request. The `turnId` shall be read from the `turn/started` server notification, whose `turn.id` is the value `turn/interrupt` requires alongside `threadId` (M0 probe, `progress.md` §E.2); recording it is what makes REQ-CX2-011 implementable. The background job runs in-process (a goroutine within the running server), so the recorded pid is one spawned in the current server lifetime; the record shall carry no reattachment metadata for a later server lifetime.

**REQ-CX2-004** (Ubiquitous) A job record's `status` shall be one of `queued`, `running`, `completed`, `failed`, `cancelled`; each transition shall be written atomically, and **where** the state directory is unwritable the tool shall return a structured error result rather than panicking or aborting the server process.

**REQ-CX2-005** (Unwanted) A job record shall not contain codex credentials, API keys, resolved secret values, or a copy of the process environment.

### M3 — `codex_task`

**REQ-CX2-006** (Event-driven) **When** `codex_task` is invoked, the server shall drive a codex `turn/start` carrying the caller's prompt, returning the completed task output when `background` is false and a job id when `background` is true.

**REQ-CX2-007** (Capability gate) **Where** the caller requests `write`, the tool shall permit codex to modify the working tree only when the project has explicitly opted in via the config key `workflow.codex.task.allow_write`, read fail-closed (a missing file, a parse error, or an absent block all read as not opted in); otherwise it shall run the turn in a non-writing mode and shall state in its result that the write request was not honored. The distributed default shall be opt-out (no writes), and the opt-in's current state shall be inspectable through `codex_setup`'s result as an `allow_write` field, alongside the existing `enable_review_gate` field. Because `sandboxPolicy` is **sticky on the thread** — the protocol documents it as overriding the policy "for this turn and subsequent turns" (M0 probe, `progress.md` §E.2) — the tool shall set `sandboxPolicy` explicitly on **every** turn it starts, transmitting the non-writing value `readOnly` on any turn that has not opted in, rather than omitting the field or relying on a per-turn default or on the thread's inherited state; a thread reused under REQ-CX2-008 `resume_last` shall therefore never inherit a write-enabled policy from an earlier turn.

**REQ-CX2-008** (State-driven) **While** `resume_last` is set, `codex_task` shall reuse the most recently recorded `threadId` for the project instead of opening a new thread; **when** no such thread is recorded, it shall open a new thread and report in its result that no prior thread was resumed.

### M4 — Job control

**REQ-CX2-009** (Event-driven) **When** `codex_job_status` is invoked with a job id, the server shall return that job's record, and **when** the id is unknown it shall return a structured not-found result rather than an error that aborts the tool call.

**REQ-CX2-010** (Event-driven) **When** `codex_job_result` is invoked for a job that has reached a terminal status, the server shall return the recorded task output; for a non-terminal job it shall return the current status without blocking the caller.

**REQ-CX2-011** (Event-driven) **When** `codex_job_cancel` is invoked for a running job, the server shall send `turn/interrupt` on that job's session carrying both of the arguments the method requires — the job record's `threadId` and its `turnId` (REQ-CX2-003) — and **when** the codex process does not exit within a bounded grace window, shall terminate the process it spawned and record the job as `cancelled`.

**REQ-CX2-012** (Unwanted) `codex_job_cancel` shall not signal or terminate any process the server did not itself spawn for that job.

### Cross-cutting

**REQ-CX2-013** (Ubiquitous) The four new tools shall be registered in `registerMoaiMCPTools` (`internal/cli/mcp_server.go:105`) with a declared JSON Schema for their inputs, and the read-only tools among them shall carry the read-only hint annotation, consistent with the existing tool surface.

**REQ-CX2-014** (Unwanted) None of the new tools or their supporting code shall invoke `AskUserQuestion` or emit a free-form user-facing question; a missing input, an unknown job, or a refused write is returned as a structured result for the orchestrator to translate.

**REQ-CX2-015** (Capability gate) **Where** this SPEC introduces a threshold, timeout, environment-variable name, or config key, it shall be defined in Go — a constant in `internal/config/defaults.go`, an env name in `internal/config/envkeys.go`, or, for the `workflow.codex.task.allow_write` key, a typed field in `internal/config/types.go` with its distributed default (`false`) in `internal/config/defaults.go` — and no file shall be added or modified under `internal/template/templates/`.

## §E. Acceptance Criteria

Enumerated as `AC-CX2-001..016` in `acceptance.md`, bound to milestones M1-M4 plus cross-cutting.

## §F. Constraints (non-functional)

- **C1 — fail-open preserved.** A missing, unauthenticated, hung, or protocol-rejecting codex never hard-blocks: the existing `inconclusiveReview` / `VerdictInconclusive` path (`mcp_codex.go:82`, `:108`) and the Stop gate's ALLOW-on-inconclusive behavior (`codex_review_gate.go:96-109`) are unchanged by this SPEC.
- **C2 — subagent boundary.** MCP tools return structured results; the orchestrator owns all user interaction.
- **C3 — secret hygiene.** No codex credential is written into a job record, a `.mcp.json` entry, or any tool result. Codex auth material stays in `~/.moai/.env.codex`.
- **C4 — model/effort SSOT.** `template.ResolveAgentModelEffort` is the sole interpreter; the existing guard `TestMCPAudit_NoDirectFrontmatterRead` (`mcp_audit_test.go:146`) must keep passing, and must now pass because the resolver is wired rather than because nothing is read.
- **C5 — same-core-two-surfaces.** Job and session logic is written once; no logic is forked between an MCP handler and any future CLI mirror.
- **C6 — no distributed-surface expansion.** No new template file, no new third-party `.mcp.json` entry.
- **C7 — no regression in the review gate.** `runCodexReviewRPC`'s existing call sites (`handleCodexAudit`, `HandleCodexReviewGate`) keep their current behavior and signature semantics.

## §G. Dependencies

- `SPEC-MOAI-MCP-SERVER-001` (`status: completed`) — delivers the server scaffold, the tool registration surface, `ReviewOutput`, the fail-open contract, and the session client of §A.1.
- Existing packages: `internal/cli` (`mcp_codex.go`, `mcp_server.go`, `codex_review_gate.go`, `session.go` `resolveProjectDir`), `internal/config` (defaults/envkeys), `internal/template` (`ResolveAgentModelEffort`, `profile_matrix.go:385`).
- Runtime, optional: the `codex` CLI binary. Absent codex leaves every new tool on the fail-open path.

## §H. Risks

- **R1 — codex app-server remains experimental and undocumented.** `turn/interrupt`, thread reuse, and any write-mode flag are asserted from the same undocumented surface that PR #1430 had to hand-probe. Mitigation: an M0-style probe against a pinned codex-cli version precedes the requirements that depend on a method name; fail-open covers a method that does not exist.
- **R2 — a long-lived codex subprocess outliving its caller.** Background jobs hold a subprocess whose lifetime is no longer bounded by a single tool call; under the in-process execution model it is bounded by the `moai mcp-server` process lifetime instead. Mitigation: a bounded job-level ceiling reusing the `DefaultCodexReviewGateTimeout` precedent, plus REQ-CX2-011 cancellation and REQ-CX2-012 pid ownership (a same-lifetime check, since every recorded pid was spawned by the running server).
- **R3 — write-mode blast radius.** `codex_task` with writes lets an external MCP host mutate the working tree through moai. Mitigation: REQ-CX2-007 opt-in with a distributed default of no writes.
- **R4 — job-record concurrency.** Two sessions on the same checkout share `.moai/state/codex-jobs/`. Mitigation: per-job files (no shared index), atomic write per REQ-CX2-004.

## §I. Cross-References

- Predecessor: `.moai/specs/SPEC-MOAI-MCP-SERVER-001/` — `spec.md:45-47` (Out of Scope — Codex Phase 2 tools), `design.md:79-82` (out-of-scope tools).
- Baseline commit: PR #1430 `d39e3cdc6` — the app-server JSON-RPC protocol fix that delivered §A.1.
- Sibling: `SPEC-AUDIT-MULTI-MODEL-001` — convergence, untouched here.
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
