---
id: SPEC-AUDIT-MODEL-CONVERGE-001
title: "Make workflow.audit.model real — one resolver turns the audit model token and gates into a backend plan, audit_multi and the plan/sync auditors follow it, and a required cross-model backend that cannot answer fails the gate by name"
version: "0.1.2"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1423)
priority: P1
phase: "v3.2.0 target"
module: "internal/config, internal/cli, internal/auditreceipt"
lifecycle: spec-anchored
tags: "audit, cross-model, convergence, codex, glm, claude-anchor, fail-closed, resolver, plan-auditor, sync-auditor, card-t1423"
tier: L
card: t1423
depends_on: [SPEC-AUDIT-MULTI-MODEL-001, SPEC-CODEX-AUDIT-GATE-AXES-001, SPEC-AUDIT-SNAPSHOT-001]
related_specs: [SPEC-MOAI-MCP-SERVER-001, SPEC-V3R6-AUDIT-MODEL-PIN-001, SPEC-MCP-WORKTREE-ROOT-001, SPEC-MCP-WORKTREE-UNTRACKED-001, SPEC-AUDIT-BUILD-IDENTITY-001, SPEC-AGENT-TIER-001]
---

# SPEC: make `workflow.audit.model` real and wire cross-model convergence into plan-audit and sync-audit

## HISTORY

- 0.1.2 — 2026-10-02 — amendment for plan-audit iteration 1 (FAIL 0.81, threshold
  0.85; report `.moai/reports/t1423/plan-audit-iter1.md`), applying the leader's
  rulings and operator decisions D7'-D11 (§B.1). The audit's defects are cited
  here as PA1-D1..D11 to keep them apart from the operator decisions. Every
  cited line was re-verified on this tree before editing (research.md §R.5);
  all of them held. Changes: PA1-D1 (silent downgrade against an older MCP
  server) — new REQ-ACV-022, a result check with a consumer for `plan_source`;
  PA1-D2 (unreadable config) — new REQ-ACV-023, a distinct non-passing state;
  PA1-D3 (caller-cancelled codex leg) — REQ-ACV-020 now covers any end of the
  leg context; PA1-D4 (deadline value and home) — the codex-leg limit moves to
  `internal/config/defaults.go`, derived from `DefaultCodexAuditTimeout`, and
  the 900 s Stop-hook arithmetic is dropped (those hooks read a persisted
  result and never bind `audit_multi`); PA1-D5 (raw reads) — folded into
  REQ-ACV-004 and REQ-ACV-007 with pins-only and explicit-`claude` fixtures;
  PA1-D6 — decision D7' (confirmed): an explicit `claude` token is Claude
  required and explicit with codex and glm left at the non-explicit default,
  superseding D4; PA1-D7 (rollout) — corrected by the leader: the new path
  applies only after merge, build, install and MCP reconnect, and until then the
  auditors and the sync step keep the pre-change path (new REQ-ACV-014), named
  as a Gap, with the activation text as the card's last milestone; PA1-D9 —
  decision D9: the Go `IsBinding` change and its AC are dropped, the sync rule
  lives in `sync.md` prose plus the machine-readable `audit_multi` result;
  decision D10 (leader override): the read-only verb is `moai verify audit-plan`,
  not a new top-level `moai audit` group; PA1-D8/D10/D11 — decision D11 and the
  optional items (tests-of-record wording, a wider leak sweep, a behavioural RED
  for AC-ACV-009, requirement wording, the import-weight decision, the
  `/moai review` leftover). Counts: requirements 21 -> 24, acceptance criteria
  21 -> 22, run-phase files about 40 -> about 38, milestones 7 (re-ordered).
- 0.1.1 — 2026-10-02 — amendment before plan-audit (card t1423), applying
  operator decisions D4-D6: the `claude` reading at confidence 0.57 (since
  superseded, see 0.1.2), a codex-leg deadline (REQ-ACV-020), and the sync
  4-dimension verdict kept with an added `audit_multi` call.
- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1423; worktree
  `.moai/worktrees/t1423`, branch `WT-audit-model-convergence`, HEAD `c50da9c2f`).
  Tier L: spec.md + plan.md + acceptance.md + design.md + research.md. Operator
  decisions D1-D3 (§B.1) are binding input and are not re-opened here. Three
  corrections to the card intake are recorded in research.md §R.2.

## §A Context

### A.1 Problem

The model that authors a SPEC or a change must not be the only judge of it. The
repository already ships the machinery for a cross-model audit — the
`audit_multi` fan-out over the Claude, codex and GLM backends, per-backend gates,
audit receipts, a SubagentStop guard that refuses an auditor PASS the receipt
store cannot corroborate — but the setting that is supposed to switch it on,
`workflow.audit.model`, is read by nobody. `plan-auditor` and `sync-auditor`
are told in prose to pick a backend "per the project's `audit_model`", which
they can only do by reading and interpreting a YAML file; and on the sync side
the 4-dimension workflow (all Claude judges) is the *binding* verdict on the
clean path, so the cold `sync-auditor` — the only agent that calls
`audit_multi` — is not spawned at all. As shipped, the plan and sync audits of a
Claude-authored change are judged by Claude only.

Target profile for plan-audit and sync-audit: **codex required, GLM advisory,
Claude as the anchor** — the profile the `audit_multi` tool already defaults to.

### A.2 Verified basis (this tree, HEAD `53a42f013`, code identical to `c50da9c2f`; commands and outputs in research.md §R.1 and §R.5)

- No non-test Go code reads `workflow.audit.model`. `activeAuditBackend`
  (`internal/cli/mcp_audit.go:52`) is referenced only by tests and comments, and
  `internal/cli/mcp_audit.go:31` still declares `multiConvergenceImplemented =
  false` ("deferred to SPEC-AUDIT-MULTI-MODEL") although `runMultiAudit`
  (`internal/cli/mcp_convergence.go:666`) exists.
- `audit_multi` takes its gates from the tool-call `gates` argument only, with
  defaults claude required, codex required, glm advisory
  (`internal/cli/mcp_audit_multi.go:124-134`).
- Fail-closed behaviour already exists for a gate the tree's `workflow.yaml`
  EXPLICITLY sets `required`: the convergence result fails with `gate_unmet`
  (`internal/cli/mcp_convergence.go:888`), the single-backend `codex_audit` does
  the same (`internal/cli/mcp_codex.go:1974`), and codex audits are recorded as
  receipts (`internal/auditreceipt/store.go:208,236`). All three read the RAW
  `audit.gates` block — through `workflowAuditPins` (which swallows a parse
  error) and a private struct in the receipt store — never a default-merged
  configuration. The engine default is deliberately not an opt-in.
- The Go default pairs `Model: claude` with the default gates
  (`internal/config/defaults.go:1272-1304`), so a default-merged read would
  hand `claude` to every project.
- The committed `.moai/config/sections/workflow.yaml` carries only the three
  backend pins; its claude pin says `effort: medium` while the Go default is
  `high` (`internal/config/closed_sets.go:106-107`).

### A.3 What this SPEC changes, in one paragraph

A single pure resolver, fed only by the tree's raw audit values, turns the
`audit.model` token and `audit.gates` block into a per-backend plan.
`audit_multi` follows that plan when the caller passes no gates; the three
existing fail-closed readers read gates through the same resolver, so a
token-chosen `required` is as binding as a gate written out; a read-only verb,
`moai verify audit-plan`, prints the plan, reports an unreadable configuration as
its own non-passing state, and can check a persisted `audit_multi` result against
the plan so that a result from an older server is not trusted; the two auditors
call that verb first and follow it, and keep the pre-change path, named as a
Gap, only while the verb itself does not exist; the codex leg of `audit_multi`
gets a deadline so a hung codex becomes an unmet gate; on the sync side the
4-dimension verdict stays the Claude input and the orchestrator adds one
`audit_multi` call after a clean PASS, the verdict being binding only when both
pass; the stale "deferred" markers go; and this repository's committed config
opts in with `audit.model: multi` and the claude pin aligned to `high`.

## §B Decisions

### B.1 Operator decisions (binding input — do not re-open)

Recorded in `.moai/reports/t1423/progress.md` (answers and confidence).

- **D1 (confidence 0.99).** One Go resolver turns config (`audit.model` +
  `audit.gates`) into the concrete backend plan. `audit_multi` falls back to that
  plan when the caller passes no `gates`. A read-only surface exposes the
  resolved plan; the auditor agents call it first and follow the measured plan
  instead of interpreting prose. Explicit call arguments still win over config.
- **D2 (confidence 0.96).** This repository's committed
  `.moai/config/sections/workflow.yaml` gets `audit.model: multi` (codex
  required, GLM advisory, Claude anchor). The distributed template default stays
  `claude`.
- **D3 (confidence 0.92).** Fail-closed with the missing backend named. A
  required backend (codex) that is unavailable — timeout, quota, missing key —
  leaves the gate unmet; the verdict names the backend; there is no automatic
  downgrade to a Claude-only verdict.
- **D4 (superseded).** The earlier reading "explicit `claude` = Claude alone"
  (confidence 0.57) is superseded by D7'.
- **D5 (confidence 0.92).** The codex leg of `audit_multi` gets a deadline
  (REQ-ACV-020); its expiry is an unmet gate naming codex, never a silent
  downgrade. Home and value per plan-audit iteration 1: design.md §D.10.
- **D6 (confidence 0.89).** The `sync-audit-4dim` verdict stays the Claude input
  and keeps its A3 fast path; after a clean four-dimension PASS, when codex is
  required, the orchestrator additionally calls `audit_multi`; the sync verdict
  is binding only when both pass (REQ-ACV-015, -021, -024).
- **D7' (confidence 0.22 — below the 0.5 gate; CONFIRMED by the factory leader).**
  An explicit `audit.model: claude` means the Claude gate is `required` and
  explicit, while codex and GLM stay at the non-explicit default (today's
  `audit_multi` behaviour; `cross_model_active` stays false). An operator who
  wants Claude alone sets `audit.gates.codex: off` and `audit.gates.glm: off`.
  It supersedes D4.
- **D8 (confidence 0.94, corrected by the leader).** Rollout is behaviour, not
  only ordering. The new path applies only once the card is merged into develop,
  a build is installed from develop and the MCP server is reconnected; until
  then the auditors and the sync step keep the pre-change path (REQ-ACV-014).
  The activation text is the card's last milestone.
- **D9 (confidence 0.86).** The Go `IsBinding` change and its acceptance
  criterion are dropped; the sync rule lives in `sync.md` prose plus the
  machine-readable `audit_multi` result (REQ-ACV-015, -021, -022, -024).
- **D10 (confidence 0.34 — below the gate; the leader OVERRIDES the original
  spelling).** The read-only verb lives in the existing `moai verify` namespace
  as `moai verify audit-plan`; no top-level `moai audit` group is created
  (design.md §D.4 states why that spelling).
- **D11 (confidence 0.75).** `/moai review` Phase 3.5 stays unchanged and is
  recorded as a known leftover prose interpreter (§D, §G R-9).

### B.2 Decisions this SPEC takes (design.md carries the reasoning)

- **B.2.1 Token semantics.** `multi` and the empty token both mean "claude
  required, codex required, glm advisory"; they differ in *source*: an empty
  token is the engine default (not an opt-in, fail-open as today), `multi` is an
  operator choice (fail-closed). `codex` and `glm` mean that backend alone (the
  other two off), matching the existing skill table "codex reviews alone".
  `claude` is decision D7': Claude required and explicit, the rest the default.
  The asymmetry is deliberate — `claude` is the Go default value and may be
  persisted untouched, so a Claude-alone reading would silently remove codex and
  GLM from `audit_multi`.
- **B.2.2 Surface.** The read-only surface is a CLI verb, not a new MCP tool
  (design.md §D.4 states the measured trade-off).
- **B.2.3 What counts as explicit.** A `required` gate counts as explicitly
  configured when it comes from the tree's `audit.gates` or from its
  `audit.model` token; a gate that only the caller's tool argument sets keeps
  today's fail-open reading (byte-identical for existing callers).
- **B.2.4 Sync verdict (D6).** The 4-dimension verdict is not demoted. Where a
  non-Claude backend is required and the 4-dimension verdict is a clean PASS,
  the orchestrator — not the workflow script, whose agents are read-only —
  calls `audit_multi` with a verdict synthesized from the workflow output as
  the Claude input; the sync verdict is binding only when both pass. The
  script's logic and output are unchanged; only its header comment and
  `meta.description` text are edited (design.md §D.8).
- **B.2.5 The check lives in the verb.** The auditors are model-driven, so the
  rule "verify the result against the plan" is given a mechanical form:
  `moai verify audit-plan --check-session <id>` reads the result the server
  persisted for that session and judges it against the plan (REQ-ACV-022). It
  is what gives `plan_source` a consumer.
- **B.2.6 Unreachable is not unmet.** Only the absence of the verb itself falls
  back to the pre-change path. A configured required backend that does not
  answer, an unreadable configuration, and a result that fails the check stay
  fail-closed (REQ-ACV-014 states the line).

## §C Requirements

Verification layer: `acceptance.md`. The requirement layer below is GEARS.

- **REQ-ACV-001** (Ubiquitous) — The audit plan resolver shall map the empty
  `audit.model` token and each token of the closed set (`claude`, `codex`, `glm`,
  `multi`) to a plan that assigns each of the three backends (`claude`, `codex`,
  `glm`) exactly one gate from `off`, `advisory`, `required`, as the table in
  design.md §D.1 states: the empty token to the distributed default profile
  (claude required, codex required, glm advisory); `multi` to the same gates
  as an operator choice; `codex` and `glm` to that backend required and the
  other two off; and `claude` to the Claude gate required and explicit with the
  codex and glm entries left at the default profile and marked default.

- **REQ-ACV-002** (Ubiquitous) — The resolver shall settle each backend's gate
  by this precedence, highest first: the gate the caller supplied for that
  backend; the tree's `audit.gates.<backend>` value; the gate the tree's
  `audit.model` token assigns; the distributed default. It shall record, for
  each backend, which of those four sources decided the gate.

- **REQ-ACV-003** (When) — **When** the tree's `audit.model` is outside the
  closed set, or any `audit.gates.<backend>` value is outside `off`, `advisory`,
  `required`, the resolver shall return an error whose text names the offending
  key and value and lists the accepted values, and neither `audit_multi` nor
  `moai verify audit-plan` shall answer that error by applying the distributed
  default.

- **REQ-ACV-004** (Ubiquitous) — The resolver shall be a pure function of the
  audit values the tree's `workflow.yaml` section file itself carries and the
  caller-supplied gates: it shall read no file, environment variable or clock,
  it shall never be fed from a default-merged configuration (a merged load pairs
  `audit.model: claude` with the default gates and would change every project),
  and it shall live in a package that both `internal/cli` and
  `internal/auditreceipt` import.

- **REQ-ACV-005** (Ubiquitous) — The `audit_multi` fallback, the explicit-
  required gate enforcement of the convergence result, the single-backend
  `codex_audit` unmet-gate rule, and the receipt gate predicate shall each obtain
  their gate values from the resolver, so that one tree configuration yields one
  reading on every surface.

- **REQ-ACV-006** (When) — **When** an `audit_multi` call omits the `gates`
  argument, or omits the key for a backend, the handler shall apply, for each
  omitted backend, the gate the resolver assigns for the audited tree; a gate the
  call supplies shall win over the tree's configuration for that backend.

- **REQ-ACV-007** (Where) — **Where** the audited tree sets neither `audit.model`
  nor `audit.gates` — including a `workflow.yaml` whose `audit` block carries
  only the backend pins, the shape the distributed template ships — and the call
  supplies no `gates`, `audit_multi` shall behave exactly as before this SPEC:
  claude required, codex required, glm advisory, and a result whose
  serialization carries no member the pre-change result did not carry; the one
  difference is a codex leg that outlasts the deadline of REQ-ACV-020, which
  previously had no bound.

- **REQ-ACV-008** (When) — **When** a `required` gate that came from the tree's
  `audit.model` token or `audit.gates` is left without a verdict by its backend —
  binary missing, key missing, quota refusal, timeout, unauthenticated, or
  malformed answer — the convergence result shall carry `overall_verdict: fail`,
  a `gate_unmet` naming each unmet backend, and a `residual_risk_note` naming
  them; it shall not be replaced by a verdict built from the remaining backends,
  and the backend's own `per_backend_verdicts` entry shall stay `inconclusive`.

- **REQ-ACV-009** (Where) — **Where** the resolver sets codex to `required` from
  the tree's configuration, the codex receipt machinery shall treat the codex gate
  as explicitly required: a receipt shall be recorded for every codex audit,
  `audit_receipt` shall be returned (including on an unmet-gate failure), and the
  SubagentStop guard shall demand a valid receipt for an auditor PASS.

- **REQ-ACV-010** (Ubiquitous) — The CLI verb `moai verify audit-plan` shall
  print the resolved backend plan of one tree as JSON and exit 0, and shall
  execute no audit, invoke no backend, create no file, and read only the tree's
  `workflow.yaml` section file and, when asked to check a result, the one
  convergence result persisted for the named session.

- **REQ-ACV-011** (When) — **When** the tree's configuration fails the resolver's
  validation, `moai verify audit-plan` shall exit 1 with the REQ-ACV-003 error
  text, prefixed `audit-plan:`, on stderr and no plan on stdout; **when** the
  tree has no `workflow.yaml` or no `audit` block it shall print the distributed
  default plan marked as default.

- **REQ-ACV-012** (Where) — **Where** the caller names a tree with
  `--project-root`, `moai verify audit-plan` shall resolve that tree — a
  config-orphaned worktree taking its primary checkout's configuration through
  the existing rules; **where** it does not, the verb shall resolve the tree the
  `moai verify` group resolves (`$CLAUDE_PROJECT_DIR`, then the working
  directory), which in a worktree-isolated session names the primary checkout.

- **REQ-ACV-013** (When) — **When** `plan-auditor` or `sync-auditor` starts an
  audit, it shall determine its own tree's toplevel and run
  `moai verify audit-plan` with `--project-root` set to it before reaching a
  verdict, and shall call `audit_multi` without a `gates` argument, with
  `project_root` and a `session_id`, whenever the plan reports a backend other
  than `claude` as active; where the plan reports none, the distributed default
  and an explicit `claude` token included, it shall keep the single-model path it
  follows today; neither agent definition nor the cross-model skill shall
  instruct choosing a backend by interpreting the `audit_model` value, except
  in the labelled legacy path of REQ-ACV-014.

- **REQ-ACV-014** (When) — **When** the verb cannot be reached — the invocation
  does not produce the verb's output contract (a JSON plan object carrying
  `config_status` on stdout, or an `audit-plan:` error line on stderr), as with
  an installed binary that predates the verb — the auditors and the sync step
  shall follow the pre-change path exactly, with no new failure and no outage,
  and shall name "plan surface unreachable, legacy path used" in the verdict's
  Gaps; this fallback applies only to the absence of the verb itself: a
  configured required backend that does not answer (REQ-ACV-008), an unreadable
  configuration (REQ-ACV-023) and a result that fails the check of REQ-ACV-022
  stay fail-closed.

- **REQ-ACV-015** (Where) — **Where** the tree's configuration sets a backend
  other than `claude` to `required` **and** the 4-dimension sync-audit verdict is
  a clean PASS (the existing binding rule holds), the sync-phase orchestration
  shall call `audit_multi` for the tree, passing as the Claude input a verdict
  synthesized from the workflow output (the 4-dimension verdict, its harmonic
  mean and its findings — never judge reasoning text), and shall treat the sync
  verdict as binding only when both the 4-dimension PASS and the `audit_multi`
  convergence pass and the result check of REQ-ACV-022 holds. The 4-dimension
  verdict shall not be demoted to non-binding for any other reason, and where no
  non-Claude backend is required the existing binding rule shall apply
  unchanged.

- **REQ-ACV-016** (Ubiquitous) — The source tree shall carry no
  `multiConvergenceImplemented` constant, no "deferred to SPEC-AUDIT-MULTI-MODEL"
  comment, and no test that pins the old sentinel; the web-console
  no-forked-interpreter guard shall name the live resolver symbol and shall fail
  when that symbol is absent from non-test source.

- **REQ-ACV-017** (Ubiquitous) — This repository's committed
  `.moai/config/sections/workflow.yaml` shall set `workflow.audit.model: multi`
  and the claude audit pin to `{claude-opus-5-5, high}`, with its header comment
  stating the same effort; the distributed template `workflow.yaml` and the Go
  default `audit.model` (`claude`) shall be unchanged.

- **REQ-ACV-018** (Ubiquitous) — The resolver shall activate no backend other
  than `claude`, `codex` and `glm`, and `audit_multi` under a plan activating all
  three shall keep the existing concurrent fan-out, so that one call's elapsed
  time is bounded by its slowest backend leg and not by the sum of the legs.

- **REQ-ACV-019** (Where) — **Where** the launch provider is not `claude`,
  `audit_multi` under a plan that gates the Claude backend shall ignore any
  supplied `claude_verdict` and obtain the Claude verdict from the independent
  Claude subscription backend, exactly as before this SPEC; the plan shall add no
  other path.

- **REQ-ACV-020** (Ubiquitous) — The codex leg of `audit_multi` shall end within
  a bounded time derived from the repository's existing codex audit bound, and
  **whenever** it ends by that deadline or by the caller's cancellation, its
  entry shall be `inconclusive` with a summary naming which of the two ended it,
  and shall carry no verdict derived from any partial output, so that a
  `required` codex gate becomes unmet and names codex (REQ-ACV-008).

- **REQ-ACV-021** (When) — **When** the 4-dimension PASS is clean and codex is
  required, the orchestrator's statement of which verdict is binding shall carry
  the plan's cross-model requirement, the `audit_multi` outcome (`pass`, `fail`,
  or unavailable), the `gate_unmet` backends, the `audit_receipt` ids and the
  result-check outcome.

- **REQ-ACV-022** (When) — **When** an auditor or the sync orchestrator holds an
  `audit_multi` result for a tree whose plan lists enforced-required backends, it
  shall run `moai verify audit-plan --check-session <id>` and require, for every
  enforced-required backend, a non-`inconclusive` `per_backend_verdicts` entry
  whose effective gate is `required`, and, where the plan reports any
  config-sourced gate, a `plan_source` of `config` on the result; an absent
  entry, an `inconclusive` entry, a different gate or a missing `plan_source`
  shall be an unmet gate named in Gaps and shall not yield PASS, so that a
  result from a server that predates the plan is never trusted.

- **REQ-ACV-023** (When) — **When** the tree's `workflow.yaml` exists but cannot
  be read or parsed, `moai verify audit-plan` shall report the distinct state
  `config_status: unreadable` with `cross_model_active: "unknown"`, reading the
  file through the loader that reports the failure (never through the reader that
  turns it into an absent configuration), and the auditors and the sync
  orchestrator shall record the unreadable file and its cause as a Gap and shall
  not yield PASS on that audit.

- **REQ-ACV-024** (When) — **When** the result check fails, or `audit_multi`
  returns a `gate_unmet` or cannot be called while the plan is reachable and
  requires a non-Claude backend, the sync verdict shall be not binding with the
  missing backend (or "audit_multi unreachable") named, and shall not be
  replaced by the 4-dimension verdict alone.

### C.1 Traceability

REQ-ACV-001 → AC-ACV-001 · REQ-ACV-002 → AC-ACV-002 · REQ-ACV-003 → AC-ACV-003 ·
REQ-ACV-004 → AC-ACV-004 + AC-ACV-008 · REQ-ACV-005 → AC-ACV-005 + AC-ACV-010 ·
REQ-ACV-006 → AC-ACV-006 + AC-ACV-007 · REQ-ACV-007 → AC-ACV-008 + AC-ACV-018 +
AC-ACV-020 · REQ-ACV-008 → AC-ACV-009 + AC-ACV-020 · REQ-ACV-009 → AC-ACV-010 ·
REQ-ACV-010 → AC-ACV-011 · REQ-ACV-011 → AC-ACV-011 · REQ-ACV-012 → AC-ACV-011 ·
REQ-ACV-013 → AC-ACV-012 · REQ-ACV-014 → AC-ACV-013 · REQ-ACV-015 → AC-ACV-014 ·
REQ-ACV-016 → AC-ACV-015 · REQ-ACV-017 → AC-ACV-016 + AC-ACV-018 · REQ-ACV-018 →
AC-ACV-017 · REQ-ACV-019 → AC-ACV-019 · REQ-ACV-020 → AC-ACV-020 · REQ-ACV-021 →
AC-ACV-014 · REQ-ACV-022 → AC-ACV-021 · REQ-ACV-023 → AC-ACV-022 · REQ-ACV-024 →
AC-ACV-014 (bodies and `**Covers**` clauses in `acceptance.md`).

## §D Out of Scope

### Out of Scope — routing implementation to other model lanes

- Sending run-phase implementation (or any task delegation) to codex or GLM
  lanes. The audit pins stay audit-only; `codex_task` and `glm_task` defaults are
  not touched.
- Per-agent model or effort overrides, and any change to the audit pins'
  backend model ids (`claude-opus-5-5`, `gpt-6.1-sol`, `glm-5.3` and their
  efforts) other than the committed claude pin effort of REQ-ACV-017.

### Out of Scope — scheduling, quotas and the review-gate hooks

- Usage-aware scheduling of audits against codex or GLM quotas. This SPEC only
  makes an exhausted quota surface as an unmet gate; it does not predict or
  avoid one.
- The Codex review-gate hook scope fixes (cards t1383 and t1404) and the
  `workflow.multi.review_gate` / `workflow.codex.review_gate` Stop-hook semantics.
  Those hooks are unchanged; the multi-review gate reads a persisted result and
  never runs `audit_multi`.
- A new deadline for any backend leg other than the codex leg of `audit_multi`
  (REQ-ACV-020). The Claude leg (5 minutes) and the GLM leg (120 seconds) keep
  their limits, and the single-backend `codex_audit` handler gets no deadline
  from this SPEC.

### Out of Scope — the distributed default and other audit surfaces

- Changing the distributed template default for `audit.model`, or adding
  `audit.model` / `audit.gates` keys to the distributed template
  `workflow.yaml`.
- A new MCP tool, a change to the MCP tool catalogue, a new top-level `moai audit`
  command group, or any change to the per-tool permission entries (the surface is
  a verb in the existing `moai verify` group; design.md §D.4).
- The Go `FourDimVerdict.IsBinding` predicate (`internal/runtime`): it is not
  changed (decision D9); the sync rule is prose plus the persisted result check.
- The judges and the logic of `.claude/workflows/sync-audit-4dim.js` — they stay
  Claude Explore agents, the script's output shape is unchanged, and the script
  does not call `audit_multi` (its agents are read-only and `audit_multi` is
  catalogued write-capable); only the script's header comment and
  `meta.description` text are edited (REQ-ACV-021, design.md §D.8).
- `/moai review` Phase 3.5 (`.claude/skills/moai/workflows/review.md` lines
  211, 213, 263, 531, 542 and its template copy): it still gates on "the
  project's `audit_model`" in prose. It is a known leftover prose interpreter
  that this SPEC does not edit (decision D11; §G R-9).
- Validating the strings a caller passes in the `gates` tool argument (only the
  tree's configuration is validated, REQ-ACV-003), and making a caller-supplied
  `required` count as explicit for fail-closed enforcement (B.2.3).
- The hook-side receipt predicate's reading of an invalid or unreadable
  `workflow.yaml` (it keeps "not required"); and `audit_multi` called outside the
  two auditors, which keeps today's reading of an unreadable file (§G R-4).

## §E Tier classification

**Tier L.** Reasoning (file count and milestones, per `spec-workflow.md` § SPEC
Complexity Tier), re-derived at 0.1.2 from the actual change: the run phase
touches about 38 distinct files —

- 11 Go source files: the new resolver and the new verb; the `audit_multi`
  handler, the convergence engine (plan source and the codex-leg deadline), the
  codex-bound default in `defaults.go`, the worktree-root gate reader, and the
  receipt store; and the stale-deferral cleanup (`mcp_audit.go`,
  `audit_models.go`, `closed_sets.go`, `schema_sections.go`);
- 13 test files, one of them a golden fixture;
- 10 mirrored documents and scripts, each as a local and a template copy:
  `plan-auditor.md`, `sync-auditor.md`, the cross-model skill, `sync.md`, and
  `sync-audit-4dim.js` (comment and description text only);
- 4 others: the committed `workflow.yaml`, the two emitted `.codex` agent TOMLs,
  and the skill catalogue hash —

against the Tier M ceiling of 15. Dropping the `IsBinding` change removed two
files (source and test) and moving the deadline into `defaults.go` replaced the
`mcp_codex.go` edit one for one. Seven milestones; the estimated change is on the
order of 1,300-1,600 lines, most of it tests and documents. Requirements 24 and
acceptance criteria 22 sit under the Tier L ceilings of 25 and 25 independently,
with one requirement and three criteria of headroom; if the plan-audit asks for
more, the first cut is REQ-ACV-018/019 (regression guards that the existing tests
already hold), not a new tier. The orchestrator's `manager-lead` entry predicate
(at least 3 milestones AND at least 10 files) is met; the choice of run-phase
mode is the orchestrator's (plan.md §E).

## §G Gaps and Residual Risks

- **R-1 — Binary lag and reachability.** The token is read by the *installed*
  binary and the long-lived MCP server. Until a build carrying this SPEC is
  installed and the MCP server reconnects, `model: multi` in the committed yaml
  changes nothing and `moai verify audit-plan` is not a verb (the `verify` group
  prints its own help and exits 0 for an unknown verb — research.md R-25).
  REQ-ACV-014 makes that state the pre-change path with a named Gap, so there is
  no outage window; the leader's rollout sequence is plan.md §J.
- **R-2 — Self-application.** Once installed, every plan-audit and sync-audit
  PASS in this repository needs a valid codex receipt; a codex outage blocks the
  pipeline by design (D3). The operator relaxes it by setting
  `audit.gates.codex: advisory` (config precedence, REQ-ACV-002).
- **R-3 — Test environment leak.** The committed `model: multi` becomes visible
  to any test that resolves its project root from `CLAUDE_PROJECT_DIR`, which
  lane sessions export. Milestone ordering (hermetic tests first, yaml late) and
  AC-ACV-018 over the measured packages close this.
- **R-4 — `audit_multi` outside the auditors.** An unreadable `workflow.yaml`
  still reads as "no audit config" on the `audit_multi` path
  (`workflowAuditPins` swallows the parse error), so a direct caller keeps
  today's fail-open reading. The auditors and the sync step never reach that
  call on such a file: the verb reports the state first and REQ-ACV-023 blocks
  PASS.
- **R-5 — The codex deadline is derived, not measured.** REQ-ACV-020 takes its
  value from `DefaultCodexAuditTimeout` (20 minutes, the bound of the same kind
  of read-only codex audit); no real codex leg duration was measured. A hung
  codex now delays an audit by up to that bound, and a review that outlasts it
  makes a required codex gate fail closed, by name. Asked in plan.md §B OQ-8.
- **R-6 — Not observed in this plan phase.** Live behaviour against real codex
  and GLM backends; whether the settings wizard persists the default `claude`
  token to YAML; whether a Codex-hosted read-only auditor role can execute
  `moai verify audit-plan`; `moai update` preservation of a user-set
  `audit.model` (only the 3-way-merge comments in
  `internal/cli/update_clean_install.go:495` were read). See plan.md §I.
- **R-7 — Where the sync binding statement is persisted.** `sync.md` does not
  say where the orchestrator's statement of the binding verdict is stored for
  the sync-audit to re-read; REQ-ACV-021 puts the cross-model lines in that
  statement without pinning its file. Asked in plan.md §B OQ-9.
- **R-8 — Install before reconnect.** A new CLI with the old MCP server still
  running is the one window in which the verb is reachable and the server is
  not new: the persisted result lacks `plan_source`, so the result check
  (REQ-ACV-022) fails closed by name until the server is reconnected. The
  leader installs and reconnects in one step (plan.md §J).
- **R-9 — A known leftover prose interpreter.** `/moai review` Phase 3.5 still
  reads the `audit_model` token by prose. It is the same defect REQ-ACV-013
  removes from the two auditors, left in place by decision D11.
- **R-10 — Instruction text is not proof of obedience.** The acceptance
  criteria that grep the agent, skill and `sync.md` text prove the text says
  the right thing, not that a model follows it. The mechanical backstops are the
  receipt guard (AC-ACV-010) and the verb's result check (AC-ACV-021).
