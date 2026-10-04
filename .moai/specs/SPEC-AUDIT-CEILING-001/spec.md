---
id: SPEC-AUDIT-CEILING-001
title: "Machine-enforced plan-audit repetition ceiling and required-backend admission"
version: "0.2.0"
status: draft
created: 2026-10-04
updated: 2026-10-04
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: "internal/runtime"
lifecycle: spec-anchored
tags: "audit-gate, plan-audit, ceiling, fail-closed, autonomy"
tier: L
related_specs: [SPEC-WF-AUDIT-GATE-001, SPEC-AUDIT-MODEL-CONVERGE-001, SPEC-AUDIT-SNAPSHOT-001, SPEC-SYNC-PARALLEL-DOCS-001]
---

# SPEC-AUDIT-CEILING-001 — Machine-enforced plan-audit repetition ceiling and required-backend admission

## §A Background and Motivation

The plan-audit retry ceiling exists only as prose in the plan-auditor agent body
(`.claude/agents/moai/plan-auditor.md` § Retry Loop Contract) and as two
configuration keys in `.moai/config/sections/harness.yaml`
(`plan_audit_tier_ceilings`, `plan_audit_ceiling_policy`) that no Go code reads
(`internal/config/loader.go` acknowledges both as orphans with the note
"prose-consumed by the plan-auditor agent body; no Go reader"). Operator
measurement on card t1500: 405 plan audits across 160 SPECs, 35% of SPECs
audited 3+ times, maximum 11 iterations — against configured ceilings of 1/2/3.
A prose ceiling cannot refuse anything; only the CLI can.

Two governing texts also disagree on what happens when the run-entry audit
verdict is negative. `.claude/skills/moai/workflows/run/phase-execution.md`
Step 4c instructs the orchestrator to present an "Override and proceed — skip
the gate (sets `--skip-audit` implicitly, records BYPASSED)" option via
AskUserQuestion, while `.claude/rules/moai/workflow/auto-semantics.md` §9
records "plan-audit bypass flags | RETIRED into the default path" and §7 makes
a negative or inconclusive verdict at an authority gate fail-closed.

Finally, the multi-backend convergence result (`ConvergenceResult`,
`internal/cli/mcp_convergence.go`) already computes per-required-backend
verdicts and an overall verdict, but the exported plan-audit verdict file
format (`.moai/docs/audit-artifact-convention.md` § What) carries no receipt of
that convergence, and the shared admission predicate
(`internal/auditverdict/verdict.go`) inspects only the auditor's own label,
score, must-pass, blocking, and hash fields. Operator-reported instances
(cards t1469, t1482; not re-verified from artifacts in this plan phase — see
`research.md` Gaps): a convergence `fail` on a required backend accompanied by
an auditor-own PASS was admitted at run entry.

This SPEC gives the ceiling teeth (a CLI-computed refusal plus a policy
outcome decided without a question), hard-blocks run entry on required-backend
failure regardless of the auditor's own verdict, and reconciles the two
conflicting passages into one fail-closed rule.

## §B Requirements (GEARS)

Requirement id prefix: `REQ-ACE` (Audit Ceiling Enforcement).

### §B.1 Work item 1 — per-SPEC audit-round counter and ceiling policy

- **REQ-ACE-001** (Ubiquitous): The CLI shall derive a SPEC's plan-audit round count from durable iteration evidence on disk — the iteration-scoped plan-audit verdict files of `.moai/docs/audit-artifact-convention.md` (`plan-audit-iter<N>.md` family) and the `<SPEC-ID>-review-<N>.md` iteration stream — counting one iteration once even when the same iteration appears in both families, and never from in-process memory. The dedupe identity is the pair (SPEC id, iteration number): two evidence files naming the same SPEC and the same iteration number are one round, and different iteration numbers are different rounds even when the verdict label, score, and audited hash are identical. A convention-family file belongs to the SPEC named in its report header; the counter resolves a SPEC's evidence from `.moai/reports/<SPEC-ID>/` plus every `.moai/reports/<card-id>/` directory whose plan-audit iteration files name that SPEC.

- **REQ-ACE-002** (Where the harness config declares `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`): The CLI shall read the tier ceiling and the policy values (`auto_delta_rounds`, `on_final_hit`) through typed Go config structs, validating `on_final_hit` against its documented value (`hold-and-split`; any other value is a config error — the CLI refuses to enforce a policy it cannot name) — the "no Go reader" orphan disposition of both keys in `internal/config/loader.go` is retired, and the config-struct/YAML symmetry audit shall cover the new structs.

- **REQ-ACE-003** (When a plan-audit round is requested for a SPEC whose round count has reached the effective ceiling — the tier ceiling plus `auto_delta_rounds`, per the composition Q2 records — or whose round count has reached the tier ceiling while no eligible auto delta round remains, where eligibility is the prose policy's own condition (`plan-auditor.md` § Retry Loop Contract): a `fix_scope`-anchored diff between the two audited SHAs and an unchanged REQ/AC id set; an ineligible delta or a STOP signal is a final hit): The CLI shall refuse the round, and shall compute and emit the ceiling-policy outcome without asking any question.

- **REQ-ACE-004** (When the refusal of REQ-ACE-003 fires and the latest non-admitted verdict fails admission on the verdict label alone — `overall_score` at or above the tier threshold, `must_pass_failed` = 0, `blocking_count` = 0, and a `plan_artifact_hash` that binds the current plan artifacts — and enumerates at least one finding): The CLI shall produce the debt-admission outcome — record PASS-WITH-DEBT with the verdict's findings enumerated as debt lines (each carrying `dispose_in`), written by the CLI into its outcome record (never by rewriting the auditor's verdict file, which would duplicate decision keys), and admit run entry at the seam that consumes the outcome. Every other admission check keeps its refusing force: a verdict that fails the score threshold, a field check, a duplicate-key check, or hash binding is never debt-admitted (C5).

- **REQ-ACE-005** (When the refusal of REQ-ACE-003 fires and the latest non-admitted verdict records one or more blocking findings and every blocking finding carries a scoped fix anchor): The CLI shall produce the scope-split outcome — persist a hold record together with a split proposal naming the anchored scope, and keep run entry blocked.

- **REQ-ACE-006** (When the refusal of REQ-ACE-003 fires and the latest non-admitted verdict satisfies neither REQ-ACE-004 nor REQ-ACE-005 — including a verdict whose `plan_artifact_hash` does not bind the current plan artifacts, which holds and is never debt-admitted, and a verdict enumerating no findings, for which no debt line can be honestly written): The CLI shall produce the hold-record outcome — persist the hold record and keep run entry blocked.

- **REQ-ACE-007** (Ubiquitous): The CLI shall report every ceiling refusal with the outcome, its reasons, and the evidence paths in its structured output, shall persist the outcome to the SPEC's `progress.md` §G Override and Refusal Record (the same section REQ-ACE-011 writes — outside the plan-artifact hash subject set), and shall complete the entire path without an interactive prompt.

### §B.2 Work item 2 — required-backend failure blocks run entry

- **REQ-ACE-008** (When a plan-audit verdict file is produced from an audit whose tree resolves a non-empty required-backend set — an `audit.gates` entry or an `audit.model` token assignment, resolved the way `resolveAuditGates` resolves them; a tree resolving no required backend requires no receipt, C4): The verdict file shall carry machine-readable receipt lines recording the convergence overall verdict and one line per required backend carrying that backend's verdict — whether the backend passed, failed, or was inconclusive — in the line format `.moai/docs/audit-artifact-convention.md` § What defines for this SPEC.

- **REQ-ACE-009** (When the tree's audit gate configuration marks a backend required and the verdict's receipt records that backend as fail or inconclusive): The shared admission predicate (`internal/auditverdict`) shall refuse the verdict regardless of the verdict's own label — an auditor-own PASS never overrides a required backend's fail, and an unresolvable required backend is not a pass.

- **REQ-ACE-010** (When a required backend is configured and the verdict carries no convergence receipt, or a receipt that omits a configured required backend's line): The admission predicate shall refuse the verdict — the fail-closed default per `auto-semantics.md` §7 (disposition confirmed or amended by decision-index Q4). Where no required backend is configured, a receipt-less verdict stays admissible (C4).

- **REQ-ACE-011** (When an operator overrides a required-backend refusal): The CLI shall accept only an explicit override input that names the backend and carries an acknowledgement note, shall itself write the override record to the SPEC's `progress.md` §G Override and Refusal Record — a section outside the plan-artifact hash subject set; never `decision-index.md`, whose edit would break the hash binding of every later verdict — and to the audit-trail log of REQ-ACE-012, and shall provide no silent-skip path.

- **REQ-ACE-012** (Ubiquitous): The CLI shall append a durable audit-trail log entry for every ceiling refusal (REQ-ACE-003) and every required-backend refusal or override (REQ-ACE-009 through REQ-ACE-011).

### §B.3 Work item 3 — one rule for the negative-verdict path, and the gate inventory

- **REQ-ACE-013** (Ubiquitous): The run-workflow gate text (`.claude/skills/moai/workflows/run/phase-execution.md` Step 4) shall present exactly the fail-closed semantics of `auto-semantics.md` §7 in its FAIL and INCONCLUSIVE question branches — no override-and-proceed option, no BYPASSED recording on a negative verdict, and no proceed-with-acknowledgement option — and shall name the complete non-block vocabulary of the negative-verdict path: the ceiling-policy outcomes of REQ-ACE-003 through REQ-ACE-006, the logged operator override of REQ-ACE-011, the existing `EnvSkipAudit` bypass (verdict BYPASSED; disposition per the `auto-semantics.md` §9 RETIRED row), and the grace-window FAIL_WARNED warn-only path — exactly one outcome vocabulary, every member named.

- **REQ-ACE-014** (Ubiquitous): The gate inventory in `auto-semantics.md` §9 shall carry a disposition row for every gate named in §D.2 of this SPEC, classified using only the existing disposition vocabulary (KEEP / AUTONOMOUS / RETIRED / PRESERVED EQUIVALENT FORM / capability switch) and citing its doctrine source by file and section name — the C2 risk-tier scheme of the source research note is not imported.

- **REQ-ACE-015** (Ubiquitous): Every deployed-file edit this SPEC makes — rules, skills, docs, and config sections — shall land in its `internal/template/templates/` mirror in the same change.

- **REQ-ACE-016** (Ubiquitous): Where this SPEC enforces a ceiling or a refusal, the enforcement shall be a CLI decision (a nonzero exit status or a recorded refusal), never prose-only guidance.

## §C Constraints

- **C1 — one policy, two consumers.** The ceiling-policy semantics stay consistent with `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier escalation paragraph and `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract; this SPEC adds the machine consumer of the existing policy keys, not a third rule text. Where the prose and the CLI disagree after implementation, the CLI is authoritative and the prose is corrected in the same change.
- **C2 — Template-First.** Constraint REQ-ACE-015 applies to every milestone.
- **C3 — no interactive prompt in the CLI path.** The CLI runs in subagent context; the refusals of REQ-ACE-003/009/010 are machine outputs, and any operator decision arrives through the orchestrator, not through the CLI.
- **C4 — backward compatibility.** A SPEC whose frontmatter carries no `tier:` field is treated as Tier L (existing rule); the legacy `<SPEC-ID>-review-<N>` naming stays counted beside the convention family; existing verdict files without receipt lines remain admissible wherever no required backend is configured.
- **C5 — admission thresholds untouched.** The per-tier plan PASS thresholds (S 0.75 / M 0.80 / L 0.85), the must-pass and blocking-count semantics, and the existing `Admit` checks keep their meaning; this SPEC only adds refusal causes and the receipt fields.

## §D Success Criteria

### §D.1 Acceptance criteria

The authoritative acceptance surface is `acceptance.md` §D (22 criteria,
`AC-ACE-001` … `AC-ACE-022`, Given-When-Then form, severity-classified, with
RED-now baselines where the defect is observable on the pre-implementation
tree and explicit new-test declarations otherwise). Every REQ in §B maps to at
least one AC there.

### §D.2 Gate inventory rows to add (REQ-ACE-014 scope)

The following 11 rows are added to `auto-semantics.md` §9. Row 2 REPLACES the
existing `plan-audit bypass flags` row (the reconciled wording), so the §9
inventory goes from 10 disposition rows to 20 — 10 − 1 replaced + 11 added.
Each row cites its source by file and section name; the disposition is
classified from the repo's own doctrine at implementation time:

1. SPEC quality gate (plan Phase 15) — `.claude/skills/moai/workflows/plan/spec-assembly.md` § Phase 15: SPEC Quality Gate
2. Plan-audit negative/inconclusive verdict at run entry (replaces the retired `plan-audit bypass flags` row) — `.claude/skills/moai/workflows/run/phase-execution.md` § Step 4 as rewritten by this SPEC
3. Run semantic failure (race, deadlock, panic, assertion) — `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy, semantic-failure paragraph
4. Sync GATE 1 pre-sync quality — `.claude/skills/moai/workflows/sync.md` gate table (`gate-sync-1`) + `workflows/sync/quality-gates-context.md`
5. Sync GATE 2 documentation scope — `.claude/skills/moai/workflows/sync.md` gate table (`gate-sync-2`) + `workflows/sync/doc-execution.md`
6. Sync deployment-readiness breaking-change acknowledgement — `.claude/skills/moai/workflows/sync/quality-gates-context.md` § Deployment Readiness
7. CI autofix iteration cap — `.claude/rules/moai/workflow/ci-autofix-protocol.md` § Iteration Cap
8. CI autofix semantic failure — `.claude/rules/moai/workflow/ci-autofix-protocol.md` § Semantic Failure — No Auto-Patch
9. Destructive command confirmation — `.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier Doctrine (3)
10. Hook block (`decision: block`) disposition — `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface
11. Pre-spawn / pre-edit divergence or foreign session — `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check / § Pre-Edit Sync Check

## §E Out of Scope

### Out of Scope — audit quality thresholds

- The per-tier plan PASS thresholds (S 0.75 / M 0.80 / L 0.85), the must-pass criterion semantics, the blocking-finding semantics, and dimension scoring are not changed by this SPEC.

### Out of Scope — auditor iteration behavior

- The plan-auditor's STOP-on-score-regression signal, delta-audit eligibility rules (`fix_scope` anchor arithmetic), and report format are not changed; this SPEC consumes their outputs, it does not retune them.

### Out of Scope — queue admission and push policy

- Nothing in this SPEC touches gtd/kanban queue admission, card pick, push batching, or PR policy.

### Out of Scope — sync-audit ceilings

- Only the plan-audit repetition ceiling is machine-enforced; a sync-audit iteration ceiling is a separate future concern.

### Out of Scope — the research note's invented tier scheme

- The source research note's C2 "Tier 0-3" gate taxonomy is not imported into `auto-semantics.md` §9; rows are classified with the inventory's existing disposition vocabulary only.

### Out of Scope — pre-existing mirror drift

- The measured pre-existing drift between `.moai/config/sections/harness.yaml`, `phase-execution.md`, and `plan-auditor.md` and their template mirrors predates this SPEC (all three measured DIFF at f2f815008, research.md §3); this SPEC must not encode that drift as expected state, and repairing unrelated drift is not its scope.

### Out of Scope — the plan_audit_global orphan key

- The third acknowledged config orphan, `plan_audit_global` (`internal/config/loader.go:344`), keeps its "no Go reader" disposition; only the two ceiling keys of REQ-ACE-002 are given a reader by this SPEC.

## §F Dependencies and Related SPECs

- `SPEC-WF-AUDIT-GATE-001` — the audit gate (verdict enum, the `GateConfig` library) this SPEC hardens; its `Invoke` seam carries no production caller today (research.md §2), so this SPEC's enforcement lands at the live admission seams (the kickoff evaluator and the card-transition guard).
- `SPEC-AUDIT-MODEL-CONVERGE-001` — the convergence result and required-gate semantics whose receipt this SPEC binds into the verdict file.
- `SPEC-AUDIT-SNAPSHOT-001` — hash-only cache validity and the plan-artifact hash subject set the admission predicate binds.
- `SPEC-SYNC-PARALLEL-DOCS-001` — origin (A6) of the `plan_audit_tier_ceilings` map this SPEC gives a Go reader.

## §G Risks

- **Receipt-absence false blocks**: trees that configure a required backend but run auditors that never emit receipts would block every run entry. Mitigated by REQ-ACE-008 (the auditor export mandate extends to receipt lines) and flagged for operator confirmation in decision-index Q4.
- **Double counting across families**: the same iteration recorded in both naming families could inflate the count. Mitigated by REQ-ACE-001's one-iteration-once rule (identity: SPEC id + iteration number — label, score, and hash are deliberately NOT the identity, so genuinely repeated audits on unchanged artifacts keep counting toward the ceiling).
- **Verdict enum extension**: consumers switch on known `Verdict` values and default unknown ones to INCONCLUSIVE (`audit_gate.go` Step 4 default branch); adding a ceiling-blocked verdict value requires updating every switch, or the refusal silently downgrades.
- **Label conversion masking auditor judgment**: the debt-admit outcome records PASS-WITH-DEBT over the auditor's own non-passing label. Bounded three ways: only a label-only admission failure qualifies (score, field, duplicate-key, and hash checks keep refusing), at least one finding must exist to enumerate as debt, and the outcome record names the converting rule so the auditor's original label stays readable beside it.
- **Config symmetry audits**: the new Go structs must join `audit_struct_yaml_symmetry_test.go` or CI fails on the very keys this SPEC adopts.

## §H Amendments

- v0.2.0 (2026-10-04): plan-audit iter1 repair (card t1500) — closed the defect delta of `.moai/reports/t1500/plan-audit-iter1.md` (FAIL 0.69; 18 findings, 13 blocking): D2 debt-admit re-gated to label-only failures (hash mismatch and no-findings verdicts hold; AC-ACE-017), D3 receipt projects every required backend's verdict with the REQ-ACE-008 trigger moved to required-configured trees (AC-ACE-018/019), D4 refusal rerouted to the live admission seams (`GateConfig.Invoke` measured production-dead; AC-ACE-022), D5 dedupe identity = SPEC id + iteration number (AC-ACE-020), D6 delta eligibility encoded (AC-ACE-021), D7 at-ceiling condition added to REQ-ACE-004..006, D8 one negative-verdict outcome vocabulary naming the override, `EnvSkipAudit`, and FAIL_WARNED, D9 AC-ACE-015 rescoped region-scoped with a named known-FAIL list and research §3 corrected, D10 override record target named (`progress.md` §G, outside the hash set), D11 M1 override input removed, D12 AC-ACE-014 replace arithmetic (10 − 1 + 11 = 20) with per-named-row verification, D13 `.moi/` → `.moai/` (×3), D14 AC-ACE-016 remapped to REQ-ACE-007 and AC-ACE-022 added for REQ-ACE-016, D15 M3 gated on Q2/Q3/Q5 with kickoff-amendable defaults, D16 §D.3 renumbered §D.2, D17 `on_final_hit` validated on load and the M2 harness.yaml mirror step dropped, D18 per-key receipt repeat rule stated. No REQ id changed (16 REQs); the AC set extended 16 → 22 (Tier L ceiling 25).
- v0.1.0 (2026-10-04): initial plan-phase authoring (card t1500).
