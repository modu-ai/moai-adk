---
id: SPEC-AUDIT-MODEL-CONVERGE-001
title: "Make workflow.audit.model real — one resolver turns the audit model token and gates into a backend plan, audit_multi and the plan/sync auditors follow it, and a required cross-model backend that cannot answer fails the gate by name"
version: "0.1.4"
status: in-progress
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

- 0.1.4 — 2026-10-02 — iteration 4 = delta confirmation over the Tier L ceiling
  of 3, leader-approved 10-02 (same criterion as t1411); D1 and D2 of
  plan-audit iteration 3 repaired (REQ-ACV-016 and its checker rule: a required
  entry counts only with a `verdict` of exactly `pass` or `fail`; REQ-ACV-007 now
  names the reader rule of REQ-ACV-020 beside the deadline). No requirement or
  criterion added, none renumbered.
- 0.1.3 — 2026-10-02 — final revision for plan-audit iteration 2 (FAIL 0.81, tied
  with iteration 1; report `.moai/reports/t1423/plan-audit-iter2.md`; its defects
  are cited as PA2-D1..D9). This revision REDUCES scope. Operator decisions D12
  and D13: the verb's `--check-session` mode that read the persisted
  `audit-multi` store is replaced by a pure checker fed the `audit_multi` result
  the auditor already holds (PA2-D1 and PA2-D2 cannot arise any more — there is
  no store read, no session id, no tree or freshness binding to build);
  requirements REQ-ACV-024 and -021 are folded into the sync requirement, and the
  regression-guard requirements 018/019 with their criteria are dropped; the
  per-test hermetic helper becomes one `TestMain` scrub in `internal/cli`.
  Mechanical fixes: the legacy path now needs a POSITIVE old-binary signature —
  any other failure to run the verb is a PASS-blocking Gap (PA2-D3) — and
  must call `audit_multi` without a `gates` argument; the "no outage window"
  overclaim is replaced by the true per-session statement (PA2-D4); the
  read-only requirement is restated against the binary's measured start-up side
  effect (PA2-D5); the codex turn reader's stream-closed arm joins the context
  arms (PA2-D6); the all-off plan dispatch is named as a residual risk (PA2-D7);
  the weakest criteria are re-anchored (PA2-D9). Counts: requirements 24 -> 20,
  acceptance criteria 22 -> 20, run-phase files 38 -> 36, milestones 7.
  Every cited source line was re-read on this tree before editing (research.md
  §R.6); all held.
- 0.1.2 — 2026-10-02 — amendment for plan-audit iteration 1 (FAIL 0.81): result
  check and unreadable-config state added, deadline moved to `defaults.go`,
  raw-read requirement, `claude` row per D7', verb moved into `moai verify`,
  legacy path for an absent verb, `IsBinding` dropped.
- 0.1.1 — 2026-10-02 — amendment before plan-audit: a codex-leg deadline and the
  sync 4-dimension verdict kept with an added `audit_multi` call.
- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1423; worktree
  `.moai/worktrees/t1423`, branch `WT-audit-model-convergence`, HEAD `c50da9c2f`).
  Tier L: spec.md + plan.md + acceptance.md + design.md + research.md. Operator
  decisions D1-D3 (§B.1) are binding input and are not re-opened here.

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

### A.2 Verified basis (this tree, HEAD `739556db5`, code identical to `c50da9c2f`; commands and outputs in research.md §R.1, §R.5 and §R.6)

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
token-chosen `required` is as binding as a gate written out. A read-only verb,
`moai verify audit-plan`, prints the plan, reports an unreadable configuration as
its own non-passing state, and can compare the plan with an `audit_multi` result
the caller passes in, so that a result from a server that predates the plan is
not trusted. The two auditors call that verb first and follow it; they keep the
pre-change path, named as a Gap, only on a positive old-binary signature. A codex
review turn yields a verdict only if it completed, and the codex leg of
`audit_multi` gets a deadline, so a hung or cut codex becomes an unmet gate. On
the sync side the 4-dimension verdict stays the Claude input and the orchestrator
adds one `audit_multi` call after a clean PASS, the verdict being binding only
when both pass. The stale "deferred" markers go, and this repository's committed
config opts in with `audit.model: multi` and the claude pin aligned to `high`.

## §B Decisions

### B.1 Operator decisions (binding input — do not re-open)

Recorded in `.moai/reports/t1423/progress.md` (answers and confidence).

- **D1 (0.99).** One Go resolver turns config (`audit.model` + `audit.gates`)
  into the concrete backend plan. `audit_multi` falls back to that plan when the
  caller passes no `gates`. A read-only surface exposes the resolved plan; the
  auditor agents call it first and follow the measured plan instead of
  interpreting prose. Explicit call arguments still win over config.
- **D2 (0.96).** This repository's committed `workflow.yaml` gets
  `audit.model: multi` (codex required, GLM advisory, Claude anchor). The
  distributed template default stays `claude`.
- **D3 (0.92).** Fail-closed with the missing backend named. A required backend
  (codex) that is unavailable — timeout, quota, missing key — leaves the gate
  unmet; the verdict names the backend; no automatic downgrade to a Claude-only
  verdict.
- **D4 (superseded)** by D7'.
- **D5 (0.92).** The codex leg of `audit_multi` gets a deadline (REQ-ACV-020);
  its expiry is an unmet gate naming codex. Home and value: design.md §D.10.
- **D6 (0.89).** The `sync-audit-4dim` verdict stays the Claude input and keeps
  its A3 fast path; after a clean PASS, when codex is required, the orchestrator
  additionally calls `audit_multi`; binding only when both pass (REQ-ACV-017).
- **D7' (0.22 — below the gate; CONFIRMED by the factory leader).** An explicit
  `audit.model: claude` is the Claude gate `required` and explicit, with codex
  and GLM left at the non-explicit default; `cross_model_active` stays false. An
  operator who wants Claude alone sets `audit.gates.codex: off` and
  `audit.gates.glm: off`.
- **D8 (0.94, corrected by the leader).** The new path applies only after the
  card is merged, a build from develop is installed, and each session's MCP
  server is reconnected; the pre-change path keeps running while the verb is
  absent (REQ-ACV-015); the activation text is the card's last milestone.
- **D9 (0.86).** The Go `IsBinding` change is dropped; the sync rule lives in
  `sync.md` prose plus the machine-readable result.
- **D10 (leader override).** The verb is `moai verify audit-plan` (no top-level
  `moai audit` group).
- **D11 (0.75).** `/moai review` Phase 3.5 stays unchanged (a known leftover
  prose interpreter, §D, §G R-9).
- **D12 (0.78).** The result check is a PURE checker: the verb compares the
  resolved plan with the `audit_multi` result JSON the caller passes in. It reads
  no `.moai/state/audit-multi/` store, takes no session id, and needs no
  freshness or tree binding (REQ-ACV-016).
- **D13 (0.82).** All proposed cuts adopted: one requirement per behaviour,
  REQ-ACV-018/019 and their criteria dropped, one `TestMain` scrub instead of a
  per-test helper (design.md §D.14).

### B.2 Decisions this SPEC takes (design.md carries the reasoning)

- **B.2.1 Token semantics.** `multi` and the empty token both mean "claude
  required, codex required, glm advisory"; they differ in *source*: an empty
  token is the engine default (not an opt-in, fail-open as today), `multi` is an
  operator choice (fail-closed). `codex` and `glm` mean that backend alone (the
  other two off). `claude` is decision D7'. The asymmetry is deliberate —
  `claude` is the Go default value and may be persisted untouched.
- **B.2.2 Surface.** The read-only surface is a CLI verb, not a new MCP tool
  (design.md §D.4).
- **B.2.3 What counts as explicit.** A `required` gate counts as explicitly
  configured when it comes from the tree's `audit.gates` or from its
  `audit.model` token; a gate that only the caller's tool argument sets keeps
  today's fail-open reading.
- **B.2.4 Sync verdict (D6).** The 4-dimension verdict is not demoted. The
  orchestrator — not the workflow script, whose agents are read-only — calls
  `audit_multi` after a clean PASS when codex is required, with a verdict
  synthesized from the workflow output as the Claude input. The script is not
  edited (design.md §D.8).
- **B.2.5 Unreachable is not unmet.** Only a positive old-binary signature
  falls back to the pre-change path. A configured required backend that does not
  answer, an unreadable configuration, a verb that fails to run for any other
  reason, and a result that fails the check all stay fail-closed, by name.

## §C Requirements

Verification layer: `acceptance.md`. The requirement layer below is GEARS; one
requirement states one behaviour.

- **REQ-ACV-001** (Ubiquitous) — The audit plan resolver shall map the empty
  `audit.model` token and each token of the closed set (`claude`, `codex`, `glm`,
  `multi`) to a plan that assigns each of the three backends (`claude`, `codex`,
  `glm`) exactly one gate from `off`, `advisory`, `required`, as the table in
  design.md §D.1 states: the empty token to the distributed default profile
  (claude required, codex required, glm advisory); `multi` to the same gates as
  an operator choice; `codex` and `glm` to that backend required and the other
  two off; and `claude` to the Claude gate required and explicit with the codex
  and glm entries left at the default profile and marked default.

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
  caller-supplied gates: it shall read no file, environment variable or clock, and
  it shall never be fed from a default-merged configuration (a merged load pairs
  `audit.model: claude` with the default gates and would change every project).

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
  serialization carries no member the pre-change result did not carry; the
  differences are those of REQ-ACV-020 — its deadline, which a codex leg
  previously had none of, and its reader rule: a codex turn that ends without
  `turn/completed` (a closed stream or an ended context) is `inconclusive`
  instead of a verdict from partial text, for every codex caller.

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
  print the resolved backend plan of one tree as JSON and exit 0 — the tree named
  by `--project-root`, otherwise the tree the `moai verify` group resolves
  (`$CLAUDE_PROJECT_DIR`, then the working directory, which in a worktree-isolated
  session names the primary checkout), a config-orphaned worktree taking its
  primary checkout's configuration through the existing rules; a tree with no
  `workflow.yaml` or no `audit` block gets the distributed default plan marked as
  default. The verb shall invoke no audit backend and shall write no state,
  receipt, configuration or audit file of the audited tree; the CLI's own
  start-up directory creation under the working directory is a pre-existing
  property of every non-trivial `moai` command, measured and named in
  design.md §D.5.

- **REQ-ACV-011** (When) — **When** the tree's configuration fails the resolver's
  validation, `moai verify audit-plan` shall exit 1 with the REQ-ACV-003 error
  text, prefixed `audit-plan:`, on stderr and no plan on stdout.

- **REQ-ACV-012** (When) — **When** the tree's `workflow.yaml` exists but cannot
  be read or parsed, `moai verify audit-plan` shall report the distinct state
  `config_status: unreadable` with `cross_model_active: "unknown"`, reading the
  file through the loader that reports the failure (never through the reader that
  turns it into an absent configuration), and shall print no default plan.

- **REQ-ACV-013** (When) — **When** `plan-auditor` or `sync-auditor` starts an
  audit, it shall determine its own tree's toplevel and run
  `moai verify audit-plan` with `--project-root` set to it before reaching a
  verdict, and shall call `audit_multi` without a `gates` argument, with
  `project_root`, whenever the plan reports a backend other than `claude` as
  active; where the plan reports none, the distributed default and an explicit
  `claude` token included, it shall keep the single-model path it follows today;
  neither agent definition nor the cross-model skill shall instruct choosing a
  backend by interpreting the `audit_model` value, except in the labelled legacy
  path of REQ-ACV-014.

- **REQ-ACV-014** (When) — **When** running the verb prints the `moai verify`
  group's own help — the positive signature of a binary that predates the verb:
  text containing `Shared diagnostic snapshot contract` and neither
  `config_status` nor an `audit-plan:` line — the auditors and the sync step shall
  follow the pre-change path, with no new failure, calling `audit_multi` without a
  `gates` argument whenever the tree's `workflow.yaml` sets an audit model other
  than `claude` or any `audit.gates` key, and shall name "plan surface
  unreachable, legacy path used" in the verdict's Gaps.

- **REQ-ACV-015** (When) — **When** the verb cannot give the auditors a plan —
  `config_status: unreadable`, an `audit-plan:` error, or any failure to run it
  other than the signature of REQ-ACV-014 (refused, crashed, timed out,
  malformed output) — the auditors and the sync orchestrator shall record the
  cause as a Gap and shall not yield PASS on that audit, as they would not for an
  unmet required gate.

- **REQ-ACV-016** (When) — **When** an auditor or the sync orchestrator holds an
  `audit_multi` result for a tree whose plan lists enforced-required backends, it
  shall pass the result JSON to `moai verify audit-plan --result '<json>'`, and
  the verb shall compare it with the plan without reading any store: for every
  enforced-required backend, a `per_backend_verdicts` entry whose `verdict` is
  exactly `pass` or `fail` and whose effective gate is `required`, and, where the
  plan reports any config-sourced gate, a `plan_source` of `config` on the result;
  an absent entry, an `inconclusive` entry, an entry whose `verdict` is missing,
  null, empty or any other value, a different gate or a missing `plan_source` is
  an unmet gate, named in the verb's `convergence_check`, which the caller shall
  record in Gaps and not yield PASS on, so that a result from a server that
  predates the plan is never trusted.

- **REQ-ACV-017** (Where) — **Where** the tree's configuration sets a backend
  other than `claude` to `required` **and** the 4-dimension sync-audit verdict is
  a clean PASS, the sync-phase orchestration shall call `audit_multi` for the
  tree, passing as the Claude input a verdict synthesized from the workflow
  output (the 4-dimension verdict, its harmonic mean and its findings — never
  judge reasoning text); the sync verdict shall be binding only when the
  4-dimension PASS, the `audit_multi` convergence pass with an empty `gate_unmet`
  and the REQ-ACV-016 check all hold, and the orchestrator's statement of which
  verdict is binding shall carry the plan's cross-model requirement, the
  `audit_multi` outcome, the `gate_unmet` backends, the `audit_receipt` ids and
  the check outcome; otherwise the verdict shall be not binding with the missing
  backend (or "audit_multi unreachable") named, and shall not be replaced by the
  4-dimension verdict alone; where no non-Claude backend is required the existing
  binding rule applies unchanged.

- **REQ-ACV-018** (Ubiquitous) — The source tree shall carry no
  `multiConvergenceImplemented` constant, no "deferred to SPEC-AUDIT-MULTI-MODEL"
  comment, and no test that pins the old sentinel; the web-console
  no-forked-interpreter guard shall name the live resolver symbol and shall fail
  when that symbol is absent from non-test source.

- **REQ-ACV-019** (Ubiquitous) — This repository's committed
  `.moai/config/sections/workflow.yaml` shall set `workflow.audit.model: multi`
  and the claude audit pin to `{claude-opus-5-5, high}`, with its header comment
  stating the same effort; the distributed template `workflow.yaml` and the Go
  default `audit.model` (`claude`) shall be unchanged; and the `internal/cli`
  test binary shall start with `CLAUDE_PROJECT_DIR` unset, so that no existing test
  can read the committed file through the ambient project directory.

- **REQ-ACV-020** (Ubiquitous) — A codex review turn shall yield a verdict only
  from a turn that completed: a turn that ends any other way — the leg's
  deadline, the caller's cancellation, or the stream closing before
  `turn/completed` — shall be `inconclusive` with the cause named and no verdict
  derived from partial output, and the codex leg of `audit_multi` shall end
  within a bound derived from the repository's existing codex audit bound, so
  that a `required` codex gate then becomes unmet and names codex (REQ-ACV-008).

### C.1 Traceability

REQ-ACV-001 → AC-ACV-001 · REQ-ACV-002 → AC-ACV-002 · REQ-ACV-003 → AC-ACV-003 ·
REQ-ACV-004 → AC-ACV-004 + AC-ACV-008 · REQ-ACV-005 → AC-ACV-005 + AC-ACV-010 ·
REQ-ACV-006 → AC-ACV-006 + AC-ACV-007 · REQ-ACV-007 → AC-ACV-008 + AC-ACV-019 +
AC-ACV-020 · REQ-ACV-008 → AC-ACV-009 + AC-ACV-020 · REQ-ACV-009 → AC-ACV-010 ·
REQ-ACV-010 → AC-ACV-011 · REQ-ACV-011 → AC-ACV-011 · REQ-ACV-012 → AC-ACV-012 ·
REQ-ACV-013 → AC-ACV-013 · REQ-ACV-014 → AC-ACV-014 · REQ-ACV-015 → AC-ACV-012 +
AC-ACV-014 · REQ-ACV-016 → AC-ACV-015 · REQ-ACV-017 → AC-ACV-016 · REQ-ACV-018 →
AC-ACV-017 · REQ-ACV-019 → AC-ACV-018 + AC-ACV-019 · REQ-ACV-020 → AC-ACV-020
(bodies and `**Covers**` clauses in `acceptance.md`).

## §D Out of Scope

### Out of Scope — routing implementation to other model lanes

- Sending run-phase implementation (or any task delegation) to codex or GLM
  lanes. The audit pins stay audit-only; `codex_task` and `glm_task` defaults are
  not touched.
- Per-agent model or effort overrides, and any change to the audit pins'
  backend model ids (`claude-opus-5-5`, `gpt-6.1-sol`, `glm-5.3` and their
  efforts) other than the committed claude pin effort of REQ-ACV-019.

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
  their limits.

### Out of Scope — the distributed default and other audit surfaces

- Changing the distributed template default for `audit.model`, or adding
  `audit.model` / `audit.gates` keys to the distributed template
  `workflow.yaml`.
- A new MCP tool, a change to the MCP tool catalogue, a new top-level `moai audit`
  command group, or any change to the per-tool permission entries (the surface is
  a verb in the existing `moai verify` group; design.md §D.4).
- Any read of the persisted `.moai/state/audit-multi/` store by the verb, and any
  session-id argument (decision D12).
- The Go `FourDimVerdict.IsBinding` predicate (`internal/runtime`) and the
  sync-audit workflow script `.claude/workflows/sync-audit-4dim.js`: neither is
  edited. The script's header comment keeps describing an unconditional happy
  path; `sync.md` is the binding rule (§G R-11).
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
Complexity Tier), re-derived at 0.1.3 from the actual change: the run phase
touches about 36 distinct files —

- 12 Go source files: the new resolver and the new verb; the `audit_multi`
  handler, the convergence engine (plan source and the codex-leg deadline), the
  codex turn reader (`mcp_codex.go`), the codex-bound default in `defaults.go`,
  the worktree-root gate reader and the receipt store; and the stale-deferral
  cleanup (`mcp_audit.go`, `audit_models.go`, `closed_sets.go`,
  `schema_sections.go`);
- 12 test files, one of them a golden fixture;
- 8 mirrored documents, each as a local and a template copy: `plan-auditor.md`,
  `sync-auditor.md`, the cross-model skill and `sync.md`;
- 4 others: the committed `workflow.yaml`, the two emitted `.codex` agent TOMLs,
  and the skill catalogue hash —

against the Tier M ceiling of 15. Compared with 0.1.2 (38 files) this removes the
two copies of `sync-audit-4dim.js`, one test file (the receipt tests move into the
new config-plan test file) and the two per-test hermetic edits (replaced by one
`main_test.go` edit), and adds `mcp_codex.go` (the reader fix). Seven milestones;
the estimated change is on the order of 1,100-1,400 lines, most of it tests and
documents. Requirements 20 and acceptance criteria 20 sit under the Tier L
ceilings of 25 and 25 independently. The orchestrator's `manager-lead` entry
predicate (at least 3 milestones AND at least 10 files) is met; the choice of
run-phase mode is the orchestrator's (plan.md §E).

## §G Gaps and Residual Risks

- **R-1 — Binary lag and reachability, stated truthfully.** While the verb is
  absent (an installed binary that predates it) the auditors and the sync step
  keep the pre-change path with a named Gap, so no audit is blocked by the missing
  verb. After a build carrying this SPEC is installed, the CLI is resolved per
  call but **every live lane and leader session keeps its own MCP server**, which
  predates the install: until that session reconnects or restarts, its
  `audit_multi` results lack `plan_source`, so the REQ-ACV-016 check fails closed
  BY NAME for that session's audits. That is a bounded, named, per-session
  window — not an absence of one. `moai doctor` compares a running server with the
  installed binary and is the documented check (plan.md §J).
- **R-2 — Self-application.** Once installed, every plan-audit and sync-audit
  PASS in this repository needs a valid codex receipt; a codex outage blocks the
  pipeline by design (D3). The operator relaxes it by setting
  `audit.gates.codex: advisory` (config precedence, REQ-ACV-002).
- **R-3 — Test environment leak.** The committed `model: multi` becomes visible
  to any test that resolves its project root from `CLAUDE_PROJECT_DIR`, which
  lane sessions export. One scrub in `internal/cli`'s `TestMain` closes it for the
  package whose code resolves the root that way; `internal/hook` already
  scrubs it (research.md R-31).
- **R-4 — `audit_multi` outside the auditors.** An unreadable `workflow.yaml`
  still reads as "no audit config" on the `audit_multi` path
  (`workflowAuditPins` swallows the parse error), so a direct caller keeps
  today's fail-open reading. The auditors and the sync step never reach that
  call on such a file: the verb reports the state first and REQ-ACV-015 blocks
  PASS.
- **R-5 — The codex deadline is derived, not measured.** REQ-ACV-020 takes its
  value from `DefaultCodexAuditTimeout` (20 minutes); no real codex leg duration
  was measured. A hung codex delays an audit by up to that bound, and a review
  that outlasts it makes a required codex gate fail closed, by name. Asked in
  plan.md §B OQ-8.
- **R-6 — Not observed in this plan phase.** Live behaviour against real codex
  and GLM backends; whether the settings wizard persists the default `claude`
  token to YAML; whether a Codex-hosted read-only auditor role can execute
  `moai verify audit-plan` (it would then take the Gap path of REQ-ACV-015, not
  the legacy path); `moai update` preservation of a user-set `audit.model`. See
  plan.md §I.
- **R-7 — Where the sync binding statement is persisted.** `sync.md` does not
  say where the orchestrator's statement of the binding verdict is stored;
  REQ-ACV-017 puts the cross-model lines in that statement without pinning its
  file. Asked in plan.md §B OQ-9.
- **R-8 — The checker trusts the caller's transcription.** The verb checks the
  result JSON the auditor passes; it cannot know the auditor copied it faithfully.
  It is a backstop against the accident (an older server's result shape), not
  against a model that fabricates a digest; the independent mechanical backstop is
  the receipt guard (AC-ACV-010).
- **R-9 — A known leftover prose interpreter.** `/moai review` Phase 3.5 still
  reads the `audit_model` token by prose, the defect REQ-ACV-013 removes from the
  two auditors, left in place by decision D11.
- **R-10 — Instruction text is not proof of obedience.** The criteria that grep
  the agent, skill and `sync.md` text prove the text says the right thing, not
  that a model follows it. They are tests-of-record for instruction text only; the
  mechanical backstops are the receipt guard (AC-ACV-010) and the verb's checker
  and output (AC-ACV-011, -012, -015).
- **R-11 — A stale script header.** `sync-audit-4dim.js` still describes the
  4-dimension verdict as binding on the happy path without the cross-model
  condition; the rule is in `sync.md`. Editing the script was cut from this SPEC.
- **R-12 — All-off plan dispatch.** A plan with every backend `off`
  (`model: codex` with `gates.codex: off`) has `cross_model_active: false`, and
  the single-model path then sends a GPT/GLM main session to `claude_audit`,
  running a Claude review the plan turned off. Named, not fixed (no new
  requirement).
- **R-13 — The reader fix reaches every codex review path.** The codex turn
  reader has one caller (`runTurn`), shared by `codex_audit`, the codex review
  gate and `codex_task`; making a closed stream or an ended context
  `inconclusive` changes those paths too. Existing codex test families must stay
  green (AC-ACV-020).
