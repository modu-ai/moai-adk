---
id: SPEC-AUDIT-MODEL-CONVERGE-001
title: "Make workflow.audit.model real — one resolver turns the audit model token and gates into a backend plan, audit_multi and the plan/sync auditors follow it, and a required cross-model backend that cannot answer fails the gate by name"
version: "0.1.1"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1423)
priority: P1
phase: "v3.2.0 target"
module: "internal/config, internal/cli, internal/auditreceipt, internal/runtime"
lifecycle: spec-anchored
tags: "audit, cross-model, convergence, codex, glm, claude-anchor, fail-closed, resolver, plan-auditor, sync-auditor, card-t1423"
tier: L
card: t1423
depends_on: [SPEC-AUDIT-MULTI-MODEL-001, SPEC-CODEX-AUDIT-GATE-AXES-001, SPEC-AUDIT-SNAPSHOT-001]
related_specs: [SPEC-MOAI-MCP-SERVER-001, SPEC-V3R6-AUDIT-MODEL-PIN-001, SPEC-MCP-WORKTREE-ROOT-001, SPEC-MCP-WORKTREE-UNTRACKED-001, SPEC-AUDIT-BUILD-IDENTITY-001, SPEC-AGENT-TIER-001]
---

# SPEC: make `workflow.audit.model` real and wire cross-model convergence into plan-audit and sync-audit

## HISTORY

- 0.1.1 — 2026-10-02 — amendment before plan-audit (card t1423), applying
  operator decisions D4-D6 (§B.1). D4: the `claude` token reading is adopted at
  confidence 0.57 and flagged for the plan-audit to re-weigh (OQ-1 closed as a
  flagged assumption). D5 reverses the earlier default: the codex leg of
  `audit_multi` gets a named deadline (new REQ-ACV-020, AC-ACV-020, M3). D6
  reverses the earlier default: the 4-dimension sync verdict is NOT demoted;
  the orchestrator additionally calls `audit_multi` after a clean 4-dimension
  PASS when codex is required, and the sync verdict is binding only when both
  pass (REQ-ACV-015 rewritten, new REQ-ACV-021, AC-ACV-014 rewritten, new
  AC-ACV-021). REQ-ACV-007 gains a deadline carve-out. Requirements 19 -> 21,
  acceptance criteria 19 -> 21 (Tier L ceilings 25 / 25); the file count in §E
  is re-derived (about 29 -> about 40).
- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1423; worktree
  `.moai/worktrees/t1423`, branch `WT-audit-model-convergence`, HEAD `c50da9c2f`).
  Tier L: spec.md + plan.md + acceptance.md + design.md + research.md. Operator
  decisions D1-D3 (§B.1) are binding input and are not re-opened here. Three
  corrections to the card intake are recorded in research.md §R.2 (the config
  gates already have more readers than the intake counted; the distributed
  template already carries the claude pin at `high`; the 900 s figure is the
  Stop-hook budget, not an `audit_multi` bound).

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

### A.2 Verified basis (this tree, HEAD `c50da9c2f`; commands and outputs in research.md §R.1)

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
  `audit.gates` block. The engine default is deliberately not an opt-in.
- The committed `.moai/config/sections/workflow.yaml` carries only the three
  backend pins; its claude pin says `effort: medium` while the Go default is
  `high` (`internal/config/closed_sets.go:106-107`).

### A.3 What this SPEC changes, in one paragraph

A single pure resolver turns the tree's `audit.model` token and `audit.gates`
block into a per-backend plan. `audit_multi` follows that plan when the caller
passes no gates; the three existing fail-closed readers read gates through the
same resolver, so a token-chosen `required` is as binding as a gate written
out; a read-only CLI verb prints the plan; the two auditors call that verb
first and follow it; the codex leg of `audit_multi` gets a named deadline so a
hung codex becomes an unmet gate; on the sync side the 4-dimension verdict stays
the Claude input and the orchestrator adds one `audit_multi` call after a clean
PASS, the verdict being binding only when both pass; the stale "deferred"
markers go; and this repository's committed config opts in with
`audit.model: multi` and the claude pin aligned to `high`.

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
- **D4 (confidence 0.57 — low).** An explicit `audit.model: claude` means Claude
  alone (claude required, codex and glm off). Adopted at confidence 0.57;
  **plan-audit please re-weigh**. Evidence and the consequence of the wrong
  reading are in plan.md §B OQ-1.
- **D5 (confidence 0.92).** The codex leg of `audit_multi` gets a named
  deadline (REQ-ACV-020); its expiry is an unmet gate naming codex, never a
  silent downgrade.
- **D6 (confidence 0.89).** The `sync-audit-4dim` verdict stays the Claude input
  and keeps its A3 fast path; after a clean four-dimension PASS, when codex is
  required, the orchestrator additionally calls `audit_multi`; the sync verdict
  is binding only when both the four-dimension PASS and the `audit_multi`
  convergence pass (REQ-ACV-015, REQ-ACV-021).

### B.2 Decisions this SPEC takes (design.md carries the reasoning)

- **B.2.1 Token semantics.** `multi` and the empty token both mean "claude
  required, codex required, glm advisory"; they differ in *source*: an empty
  token is the engine default (not an opt-in, fail-open as today), `multi` is an
  operator choice (fail-closed). `claude`, `codex` and `glm` mean that backend
  alone (the other two off), matching the existing skill table "codex reviews
  alone". The `claude` reading is decision D4 (confidence 0.57), a flagged
  assumption for the plan-audit to re-weigh (plan.md §B OQ-1).
- **B.2.2 Surface.** The read-only surface is a CLI verb, `moai audit plan`, not
  a new MCP tool. design.md §D.4 states the measured trade-off.
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
- **B.2.5 Codex leg deadline (D5).** The deadline equals the existing Claude
  stage limit (5 minutes), derived in design.md §D.10; no new number is
  introduced.

## §C Requirements

Verification layer: `acceptance.md`. The requirement layer below is GEARS.

- **REQ-ACV-001** (Ubiquitous) — The audit plan resolver shall map the empty
  `audit.model` token and each token of the closed set (`claude`, `codex`, `glm`,
  `multi`) to a plan that assigns each of the three backends (`claude`, `codex`,
  `glm`) exactly one gate from `off`, `advisory`, `required`, as the table in
  design.md §D.1 states; the empty token shall map to the distributed default
  profile (claude required, codex required, glm advisory).

- **REQ-ACV-002** (Ubiquitous) — The resolver shall settle each backend's gate
  by this precedence, highest first: the gate the caller supplied for that
  backend; the tree's `audit.gates.<backend>` value; the gate the tree's
  `audit.model` token assigns; the distributed default. It shall record, for
  each backend, which of those four sources decided the gate.

- **REQ-ACV-003** (When) — **When** the tree's `audit.model` is outside the
  closed set, or any `audit.gates.<backend>` value is outside `off`, `advisory`,
  `required`, the resolver shall return an error whose text names the offending
  key and value and lists the accepted values, and neither `audit_multi` nor
  `moai audit plan` shall answer that error by applying the distributed default.

- **REQ-ACV-004** (Ubiquitous) — The resolver shall be a pure function of the
  audit configuration value and the caller-supplied gates: it shall read no file,
  environment variable or clock, and it shall live in a package that both
  `internal/cli` and `internal/auditreceipt` import.

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
  nor `audit.gates` and the call supplies no `gates`, `audit_multi` shall behave
  exactly as before this SPEC: claude required, codex required, glm advisory,
  and a result whose serialization carries no member the pre-change result did
  not carry; the one difference is a codex leg that outlasts the deadline of
  REQ-ACV-020, which previously had no bound.

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

- **REQ-ACV-010** (Ubiquitous) — The CLI verb `moai audit plan` shall print the
  resolved backend plan of one tree as JSON and exit 0, and shall execute no
  audit, invoke no backend, create no file, and read only the tree's
  `workflow.yaml` section file.

- **REQ-ACV-011** (When) — **When** the tree's configuration fails the resolver's
  validation, `moai audit plan` shall exit 1 with the REQ-ACV-003 error text on
  stderr and no plan on stdout; **when** the tree has no `workflow.yaml` or no
  `audit` block it shall print the distributed default plan marked as default.

- **REQ-ACV-012** (Where) — **Where** the caller names a tree with
  `--project-root`, `moai audit plan` shall resolve that tree — a config-orphaned
  worktree taking its primary checkout's configuration through the existing
  rules — and otherwise the tree containing its working directory.

- **REQ-ACV-013** (When) — **When** `plan-auditor` or `sync-auditor` starts an
  audit, it shall run `moai audit plan` for its own tree before reaching a
  verdict, and shall call `audit_multi` without a `gates` argument whenever the
  tree's configuration sets a backend other than `claude` to `advisory` or
  `required` (the plan marks that entry explicit), folding the result as
  `moai-ref-cross-model-audit` states; where no such backend is set — the
  distributed default included — it shall keep the single-model path it follows
  today; neither agent definition nor that skill shall instruct choosing a
  backend by interpreting the `audit_model` value.

- **REQ-ACV-014** (When) — **When** the `audit_multi` result carries a
  `gate_unmet`, or `moai audit plan` cannot be reached, the auditor shall name
  the unmet backend (or the unreachable surface) in its Gaps and Residual-risk,
  shall not reach PASS on the strength of the remaining backends, and shall cite
  the returned `audit_receipt` ids in its `AUDIT-VERDICT` line.

- **REQ-ACV-015** (Where) — **Where** the tree's configuration sets a backend
  other than `claude` to `required` **and** the 4-dimension sync-audit verdict is
  a clean PASS (the existing binding predicate holds), the sync-phase
  orchestration shall call `audit_multi` for the tree, passing as the Claude
  input a verdict synthesized from the workflow output (the 4-dimension verdict,
  its harmonic mean and its findings — never judge reasoning text), and shall
  treat the sync verdict as binding only when both the 4-dimension PASS and the
  `audit_multi` convergence pass with no `gate_unmet`; the machine binding
  predicate shall carry that second condition. The 4-dimension verdict shall not
  be demoted to non-binding for any other reason, and where no non-Claude
  backend is required the existing binding rule shall apply unchanged.

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

- **REQ-ACV-020** (Ubiquitous) — The codex leg of `audit_multi` shall run under
  a named deadline held in one constant beside the other leg limits and equal to
  the existing Claude leg limit (5 minutes); **when** the deadline expires, the
  leg shall be abandoned, its process stopped, and its entry returned as
  `inconclusive` with a summary naming the codex timeout — never a verdict
  parsed from partial output — so that a `required` codex gate becomes unmet and
  names codex (REQ-ACV-008).

- **REQ-ACV-021** (When) — **When** the 4-dimension PASS is clean and codex is
  required, the orchestrator's statement of which verdict is binding shall carry
  the plan's cross-model requirement, the `audit_multi` outcome (`pass`, `fail`,
  or unavailable), the `gate_unmet` backends and the `audit_receipt` ids; **when**
  `audit_multi` returns a `gate_unmet`, or cannot be called, the sync verdict
  shall be not-binding with the missing backend (or "audit_multi unreachable")
  named and shall not be replaced by the 4-dimension verdict alone; the
  4-dimension fast path shall stay in force for the Claude half, at the cost of
  one `audit_multi` call per sync when codex is required.

### C.1 Traceability

REQ-ACV-001 → AC-ACV-001 · REQ-ACV-002 → AC-ACV-002 · REQ-ACV-003 → AC-ACV-003 ·
REQ-ACV-004 → AC-ACV-004 · REQ-ACV-005 → AC-ACV-005 + AC-ACV-010 · REQ-ACV-006 →
AC-ACV-006 + AC-ACV-007 · REQ-ACV-007 → AC-ACV-008 + AC-ACV-018 + AC-ACV-020 · REQ-ACV-008 →
AC-ACV-009 + AC-ACV-020 · REQ-ACV-009 → AC-ACV-010 · REQ-ACV-010 → AC-ACV-011 · REQ-ACV-011 →
AC-ACV-011 · REQ-ACV-012 → AC-ACV-011 · REQ-ACV-013 → AC-ACV-012 · REQ-ACV-014 →
AC-ACV-013 · REQ-ACV-015 → AC-ACV-014 + AC-ACV-021 · REQ-ACV-016 → AC-ACV-015 · REQ-ACV-017 →
AC-ACV-016 + AC-ACV-018 · REQ-ACV-018 → AC-ACV-017 · REQ-ACV-019 → AC-ACV-019 · REQ-ACV-020 → AC-ACV-020 · REQ-ACV-021 → AC-ACV-021 +
AC-ACV-014
(bodies and `**Covers**` clauses in `acceptance.md`).

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
  Those hooks are unchanged.
- A new deadline for any backend leg other than the codex leg of `audit_multi`
  (REQ-ACV-020). The Claude leg (5 minutes) and the GLM leg (120 seconds) keep
  their limits, and the single-backend `codex_audit` handler gets no deadline
  from this SPEC.

### Out of Scope — the distributed default and other audit surfaces

- Changing the distributed template default for `audit.model`, or adding
  `audit.model` / `audit.gates` keys to the distributed template
  `workflow.yaml`.
- A new MCP tool, a change to the MCP tool catalogue, or any change to the
  per-tool permission entries (the surface is a CLI verb; design.md §D.4).
- The judges and the logic of `.claude/workflows/sync-audit-4dim.js` — they stay
  Claude Explore agents, the script's output shape is unchanged, and the script
  does not call `audit_multi` (its agents are read-only and `audit_multi` is
  catalogued write-capable); only the script's header comment and
  `meta.description` text are edited (REQ-ACV-021, design.md §D.8).
- `/moai review` Phase 3.5: it gates on the cross-model skill, which this SPEC
  updates; its own text is not edited.
- Validating the strings a caller passes in the `gates` tool argument (only the
  tree's configuration is validated, REQ-ACV-003), and making a caller-supplied
  `required` count as explicit for fail-closed enforcement (B.2.3).
- The fail-open reading of an unreadable `workflow.yaml` (existing posture).
  `moai audit plan` reports the condition; `audit_multi` keeps today's behaviour.

## §E Tier classification

**Tier L.** Reasoning (file count and milestones, per `spec-workflow.md` § SPEC
Complexity Tier), re-derived at 0.1.1 from the actual change: the run phase
touches about 40 distinct files —

- 12 Go source files: the new resolver and the new CLI verb; the `audit_multi`
  handler, the convergence engine (plan source and the codex-leg deadline), the
  codex constant, the worktree-root gate reader, the receipt store, and the
  binding predicate; and the stale-deferral cleanup (`mcp_audit.go`,
  `audit_models.go`, `closed_sets.go`, `schema_sections.go`);
- 14 test files and one golden fixture;
- 10 mirrored documents and scripts, each as a local and a template copy:
  `plan-auditor.md`, `sync-auditor.md`, the cross-model skill, `sync.md`, and
  `sync-audit-4dim.js` (comment and description text only);
- 4 others: the committed `workflow.yaml`, the two emitted `.codex` agent TOMLs,
  and the skill catalogue hash —

against the Tier M ceiling of 15. Seven milestones; the estimated change is on
the order of 1,400-1,700 lines, most of it tests and documents. Requirements 21
and acceptance criteria 21 sit under the Tier L ceilings of 25 and 25
independently. The orchestrator's `manager-lead` entry predicate (at least 3
milestones AND at least 10 files) is met; the choice of run-phase mode is the
orchestrator's (plan.md §E).

## §G Gaps and Residual Risks

- **R-1 — Binary lag window.** The token is read by the *installed* binary and
  the long-lived MCP server. Until a build carrying this SPEC is installed and
  the MCP server reconnects, `model: multi` in the committed yaml changes
  nothing, and `moai audit plan` is an unknown command. REQ-ACV-014 makes the
  unreachable surface a named Gap rather than a silent pass.
- **R-2 — Self-application.** Once installed, every plan-audit and sync-audit
  PASS in this repository needs a valid codex receipt; a codex outage blocks the
  pipeline by design (D3). The operator relaxes it by setting
  `audit.gates.codex: advisory` (config precedence, REQ-ACV-002).
- **R-3 — Test environment leak.** The committed `model: multi` becomes visible
  to any test that resolves its project root from `CLAUDE_PROJECT_DIR`, which
  lane sessions export. Milestone ordering (hermetic tests first, yaml last) and
  AC-ACV-018 close this.
- **R-4 — Unreadable config.** A `workflow.yaml` that fails to parse reads as
  "no audit config" on the `audit_multi` path (existing fail-open posture), which
  silently drops an operator's `multi`. Out of Scope here; the verb reports it.
- **R-5 — The codex deadline is derived, not measured.** REQ-ACV-020 sets the
  codex leg to the Claude stage limit (5 minutes); no real codex adversarial
  review duration was measured. A review that legitimately outlasts it makes a
  required codex gate fail closed, by name. Asked in plan.md §B OQ-8.
- **R-7 — Where the sync binding statement is persisted.** `sync.md` does not
  say where the orchestrator's statement of the binding verdict is stored for
  the sync-audit to re-read; REQ-ACV-021 puts the cross-model line in that
  statement without pinning its file. Asked in plan.md §B OQ-9.
- **R-6 — Not observed in this plan phase.** Live behaviour against real codex
  and GLM backends; whether the settings wizard persists the default `claude`
  token to YAML; whether a Codex-hosted read-only auditor role can execute
  `moai audit plan`; `moai update` preservation of a user-set `audit.model` (only
  the 3-way-merge comments in `internal/cli/update_clean_install.go:495` were
  read). See plan.md §I.
