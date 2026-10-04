---
id: SPEC-AUDIT-CEILING-002
title: "CLI-counted plan-audit iteration ceiling with recorded policy outcomes and required-backend run-entry refusal"
version: "0.1.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
author: "manager-spec"
priority: P1
phase: "v3.2.0 target"
module: "internal/runtime, internal/auditverdict, internal/config, internal/cli"
lifecycle: spec-anchored
tags: "audit-gate, plan-audit, ceiling, required-backend, fail-closed"
tier: M
related_specs: [SPEC-WF-AUDIT-GATE-001, SPEC-AUDIT-MODEL-CONVERGE-001]
---

# SPEC-AUDIT-CEILING-002 — CLI-counted plan-audit iteration ceiling with recorded policy outcomes and required-backend run-entry refusal

## §A Background and Motivation

The plan-audit iteration ceiling exists today as configuration and prose only. Measured on this tree at HEAD `e497f6936`: `.moai/config/sections/harness.yaml` carries `plan_audit_tier_ceilings` (S:1, M:2, L:3, line 75) and `plan_audit_ceiling_policy` (`auto_delta_rounds: 1`, `on_final_hit: hold-and-split`, line 82), and neither key has a Go reader — the identifiers `PlanAuditTierCeilings`, `CeilingPolicy`, `AutoDeltaRounds`, `OnFinalHit` match nothing under `internal/` or `cmd/` (grep, exit 1). The enforcing text is the plan-auditor agent body's ceiling-policy paragraph (template mirror `internal/template/templates/.claude/agents/moai/plan-auditor.md:712`). A prose ceiling cannot refuse anything; only code can.

Repetition is real and measured in this repository's own report tree (primary checkout, worktrees excluded, measured 2026-10-04): 327 `plan-audit-iter*.md` files across card evidence directories, with a maximum of 6 iteration files in a single card directory (`.moai/reports/t1152/`) — twice the configured Tier L ceiling of 3, with no code refusing. The operator's card-t1500 decision (D9) cites a wider operator measurement over the same history — 405 plan audits across 160 SPECs, maximum 11 iterations on one SPEC — as the motivation; this plan phase did not re-derive that figure and cites it as the operator's.

The negative-verdict path has two live run-entry seams, both already deciding through the shared admission predicate `internal/auditverdict` — the kickoff evaluator (`internal/contract/kickoff/decide.go:373-376`) and the card-transition guard (`internal/homestate/card_evidence_readers.go:176-192`, reached from `card_audit_kickoff.go:16-36`). The predicate inspects only the verdict's own label, score, must-pass, blocking, and hash fields (`internal/auditverdict/verdict.go:195-233`): a required backend recorded as fail has no machine surface at all (the identifier `required_backend_fail` matches nothing under `internal/` or in `.moai/docs/audit-artifact-convention.md`, grep exit 1). And the audit-gates resolver folds a configuration error into an empty gate set (`internal/cli/mcp_worktree_root.go:127-129`; the code's own comment at line 121 reads "this path fails open"), so a tree whose audit configuration cannot be parsed reads as "not configured".

Operator decision D9 (card t1500) narrowed the never-landed SPEC-AUDIT-CEILING-001 — a Tier L draft that failed three plan-audit iterations (final FAIL 0.81, STOP signal) — to exactly three in-code items: the per-SPEC iteration counter, one recording path for the ceiling-hit policy outcome, and required-backend fail blocking run entry. That draft never reached a terminable status; this SPEC is its narrowing, not a supersession (see §G History).

## §B Requirements (GEARS)

Requirement id prefix: `REQ-ACR` (Audit Ceiling Reduced — distinct from the abandoned draft's `REQ-ACE` prefix so no grep collides across the two).

### §B.1 Work item 1 — the CLI counts plan-audit iterations per SPEC

- **REQ-ACR-001** (Ubiquitous): The system shall compute a SPEC's plan-audit round count from durable iteration evidence on disk, one recorded file one round — a `plan-audit.md` file contributing one round and each `plan-audit-iter<N>.md` file contributing one round — across the evidence directories given for the SPEC (its `.moai/reports/<SPEC-ID>/` directory when present, plus every directory explicitly listed in the invocation), and the system shall never derive the count from in-process memory or session state, and a round file whose iteration suffix does not parse as a positive integer shall make the count an error, never a silent skip.

- **REQ-ACR-002** (Where the harness configuration declares `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`): The system shall read the tier ceilings and the ceiling policy through typed Go configuration structs, retiring the no-Go-reader orphan disposition of both keys, and shall resolve the SPEC's ceiling by the SPEC's tier read under the shared predicate's rule (`auditverdict.SpecTier`: tier absent or unknown resolves to L), comparing the computed round count against that ceiling.

### §B.2 Work item 2 — one path records the ceiling-hit policy outcome

- **REQ-ACR-003** (When the system computes a SPEC's round count at or above its resolved ceiling): The system shall evaluate and record the ceiling policy outcome through exactly one code path without asking any question, and the recorded disposition shall be exactly one of `debt-proceed`, `split`, `hold` — the latest verdict admitting as `PASS-WITH-DEBT` under the shared predicate records `debt-proceed`, a non-admitted verdict under the shipped policy value `hold-and-split` records `hold` together with the split-proposal reference (one record carrying both halves of the shipped value), a non-admitted verdict under any other or unreadable policy value records `hold`, and every record carries the round count, the ceiling, the latest verdict label, and the evidence paths.

- **REQ-ACR-004** (When a SPEC's round count is at or above its resolved ceiling and its latest verdict passes the full shared admission predicate with label `PASS`): The system shall not apply the ceiling — no outcome record is written and the verdict admits under the shared predicate unchanged, because the ceiling caps repetition, not a healthy result.

### §B.3 Work item 3 — required-backend fail blocks run entry in code

- **REQ-ACR-005** (When a plan-audit verdict file records one or more `required_backend_fail: <backend>` machine lines): The shared admission predicate (`internal/auditverdict`) shall refuse the verdict regardless of the verdict's own label, naming the recorded backends in the refusal reason, and the check shall be unconditional (a sync-audit verdict carrying the line is malformed evidence and refuses), and the line's producer shall be the exporting auditor — from the multi-model convergence result's per-backend verdicts, or from a single-backend audit's own review — while required backends the audit did not cover carry no line and the absence of lines is not a pass signal because the existing label and field checks remain the primary admission gates, and the line format shall be defined in `.moai/docs/audit-artifact-convention.md` § What.

- **REQ-ACR-006** (When the audit-gates resolution of a tree errors): The resolver (`resolveAuditGates`, `internal/cli/mcp_worktree_root.go`) shall keep the error distinct from an empty not-configured result, so that a resolution error is never folded into an empty gate set and the resolver's caller surfaces the error, retiring the fail-open disposition the current code records at its "this path fails open" comment (line 121).

## §C Constraints

- **C1 — no interactive question.** The ceiling path of REQ-ACR-003 asks nothing of anyone: no `AskUserQuestion`, no prompt, no free-form question. The recording path's only caller in this SPEC is the CLI verb (§D D7); the existing prose gate texts are not edited by this SPEC.
- **C2 — Template-First.** Every deployed-file edit lands with its `internal/template/templates/` mirror in the same change; the convention document is template-distributed.
- **C3 — one file one round, no dedupe identity.** The counter has no collapse rules. In a directory mixing `plan-audit.md` with `iter<N>` files, each file counts — a possible over-count whose error direction is the safe one (an early ceiling is fail-closed). Counting the same iteration number in two directories counts two rounds; automatic cross-card discovery of a SPEC's evidence is out of scope (§E).
- **C4 — additive configuration.** The new fields are additive: the harness section is a dedicated loader entry point outside `Loader.Load()` (measured: `internal/config/audit_registry.go:73-75`), so field coverage lands in the harness loader's own test surface (`internal/config/loader_harness_extended_test.go` — the harness section has no `TestStructYAMLSymmetry_*` case, measured by `go test -list`); the shipped-key inventory already carries both keys (`internal/config/testdata/shipped_key_inventory.yaml:807-810`); no existing field changes shape.
- **C5 — cross-platform.** The new code paths use no `syscall` and no OS-conditional builds.

## §D Decisions

- **D1 — counter input family.** The counter counts the card-scoped `plan-audit` family (the dominant convention per `.moai/docs/audit-artifact-convention.md` § Where; 327 files measured in this repository) plus the SPEC-scoped directory. The legacy `<SPEC-ID>-review-<N>` stream (169 files measured) lives under `.moai/reports/plan-audit/`, a directory that same convention marks FORBIDDEN — counting it would count a forbidden surface, so it is not an input.
- **D2 — outcome enum.** Exactly the three operator-named tokens (`debt-proceed`, `split`, `hold`). The shipped policy value `hold-and-split` maps to one record: disposition `hold` carrying the split-proposal reference. The `split` token exists for a policy value that selects it; no shipped value selects it today.
- **D3 — the healthy-PASS arm.** Decided explicitly (the abandoned draft's iter3 defect D31 held healthy PASSes by omission): REQ-ACR-004 records that a clean PASS at the ceiling is not a ceiling outcome at all.
- **D4 — absence is not a pass signal.** The abandoned draft's iter3 defect D32 left single-model audits unwired; here the `required_backend_fail` line is an independent blocking signal with a producer defined for BOTH the multi-model convergence path and the single-backend audit's own review, and the line's absence refuses nothing — the existing gates stay primary.
- **D5 — fail-closed resolver.** The abandoned draft's iter3 defect D21 measured the fold-error-into-empty behavior; REQ-ACR-006 retires it.
- **D6 — id prefix.** `REQ-ACR` cannot collide with the abandoned draft's `REQ-ACE` ids in any grep, log, or audit.
- **D7 — the recording path's caller.** The CLI verb (`moai spec ceiling`, `internal/cli/spec_ceiling.go`) is the recording path's only caller this SPEC builds. Wiring `GateConfig.Invoke` or the plan-auditor agent body to it is a follow-up card (§E); the path is a user- and agent-invocable CLI surface, not dead code.

## §E Exclusions

### Out of Scope — what this SPEC does not do

- Cross-card re-audit dedupe and automatic card-to-SPEC evidence discovery (the abandoned draft's D33 territory) — a follow-up card.
- Receipt formats beyond the single `required_backend_fail` line (the abandoned draft's REQ-ACE-008/009 receipt system) — the reduced item needs one line, not a receipt.
- Per-run operator override surface and audit-trail log (the abandoned draft's REQ-ACE-011/012).
- Gate-text reconciliation in `run/phase-execution.md` and `auto-semantics.md` §9 (the abandoned draft's work item 3).
- Wiring the plan-auditor agent body's ceiling section, or the audit gate (`GateConfig.Invoke`), to the new CLI verb.
- In-code `auto_delta_rounds` eligibility computation (the `fix_scope` diff check stays prose-consumed; the field is parsed and carried by the configuration structs, and only `on_final_hit` drives outcome selection in this SPEC).

## §F Risks

- **F1 — evidence coverage.** The ceiling applies only to exported evidence; an audit that exports nothing counts zero rounds. The convention's export-as-completion mandate is the coverage carrier; this SPEC does not strengthen it.
- **F2 — machine-local record.** `.moai/state/` is gitignored (`.gitignore:400`), so the outcome record is machine-local state, not a cross-machine durable carrier; the durable carriers stay the existing verdict and progress surfaces.
- **F3 — false blocks at deployment.** SPECs whose historical iteration counts already exceed their tier ceiling will hit the ceiling path on their next non-admitted verdict. The path records a policy outcome instead of asking a question, which is the mitigation this SPEC ships; a bulk grandfathering pass is out of scope.

## §G History

- 2026-10-04, v0.1.0 (card t1500, operator decision D9): narrowed restart of SPEC-AUDIT-CEILING-001 — that draft (Tier L, 16 REQ / 22 AC, old-card branch only) failed three plan-audit iterations (0.69 → 0.83 → 0.81, STOP signal) and was never superseded because it never reached a terminable status. Its final-round lessons are folded in: D31 → REQ-ACR-004; D32 → REQ-ACR-005's producer clause; D21 → REQ-ACR-006; D34 (stale describing surfaces) → every cross-reference in this SPEC was measured on this tree at HEAD `e497f6936`; the structural lesson (16 REQ / 22 AC cannot survive adversarial rounds) → Tier M, 6 REQ, 8 AC.
