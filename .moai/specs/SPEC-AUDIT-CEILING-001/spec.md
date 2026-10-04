---
id: SPEC-AUDIT-CEILING-001
title: "Machine-enforced plan-audit repetition ceiling and required-backend admission"
version: "0.1.0"
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

- **REQ-ACE-001** (Ubiquitous): The CLI shall derive a SPEC's plan-audit round count from durable iteration evidence on disk — the iteration-scoped plan-audit verdict files of `.moi/docs/audit-artifact-convention.md` (`plan-audit-iter<N>.md` family) and the `<SPEC-ID>-review-<N>.md` iteration stream — counting one iteration once even when the same iteration appears in both families, and never from in-process memory.

- **REQ-ACE-002** (Where the harness config declares `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`): The CLI shall read the tier ceiling and the policy values (`auto_delta_rounds`, `on_final_hit`) through typed Go config structs — the "no Go reader" orphan disposition of both keys in `internal/config/loader.go` is retired, and the config-struct/YAML symmetry audit shall cover the new structs.

- **REQ-ACE-003** (When a plan-audit round is requested for a SPEC whose round count has reached the effective ceiling — the tier ceiling plus `auto_delta_rounds` delta rounds, per the composition Q2 records): The CLI shall refuse the round, and shall compute and emit the ceiling-policy outcome without asking any question.

- **REQ-ACE-004** (When the latest non-admitted verdict records `must_pass_failed` = 0 and `blocking_count` = 0): The CLI shall produce the debt-admission outcome — record the verdict as PASS-WITH-DEBT with at least one enumerated debt line (each carrying `dispose_in`), and admit run entry under the existing PASS-WITH-DEBT admission rules.

- **REQ-ACE-005** (When the latest non-admitted verdict records one or more blocking findings and every blocking finding carries a scoped fix anchor): The CLI shall produce the scope-split outcome — persist a hold record together with a split proposal naming the anchored scope, and keep run entry blocked.

- **REQ-ACE-006** (When the latest non-admitted verdict satisfies neither REQ-ACE-004 nor REQ-ACE-005): The CLI shall produce the hold-record outcome — persist the hold record and keep run entry blocked.

- **REQ-ACE-007** (Ubiquitous): The CLI shall report every ceiling refusal with the outcome, its reasons, and the evidence paths in its structured output, shall persist the outcome to the SPEC's `progress.md`, and shall complete the entire path without an interactive prompt.

### §B.2 Work item 2 — required-backend failure blocks run entry

- **REQ-ACE-008** (When a plan-audit verdict file is produced from an audit that fanned out to multiple backends): The verdict file shall carry machine-readable receipt lines recording the convergence overall verdict and, for each backend whose gate is `required`, that backend's verdict — in the line format `.moi/docs/audit-artifact-convention.md` § What defines for this SPEC.

- **REQ-ACE-009** (When the tree's audit gate configuration marks a backend required and the verdict's receipt records that backend as fail): The shared admission predicate (`internal/auditverdict`) shall refuse the verdict regardless of the verdict's own label — an auditor-own PASS never overrides a required backend's fail.

- **REQ-ACE-010** (When a required backend is configured and the verdict carries no convergence receipt): The admission predicate shall refuse the verdict — the fail-closed default per `auto-semantics.md` §7 (disposition confirmed or amended by decision-index Q4).

- **REQ-ACE-011** (When an operator overrides a required-backend refusal): The CLI shall accept only an explicit override input that names the backend and carries an acknowledgement note, shall record the override in the SPEC's `progress.md`, the decision record, and the audit-trail log of REQ-ACE-012, and shall provide no silent-skip path.

- **REQ-ACE-012** (Ubiquitous): The CLI shall append a durable audit-trail log entry for every ceiling refusal (REQ-ACE-003) and every required-backend refusal or override (REQ-ACE-009 through REQ-ACE-011).

### §B.3 Work item 3 — one rule for the negative-verdict path, and the gate inventory

- **REQ-ACE-013** (Ubiquitous): The run-workflow gate text (`.claude/skills/moai/workflows/run/phase-execution.md` Step 4) shall present exactly the fail-closed semantics of `auto-semantics.md` §7 for a FAIL or INCONCLUSIVE plan-audit verdict — no operator question branch and no override-and-proceed option — so exactly one rule governs the negative-verdict path, and the surface the ceiling policy (REQ-ACE-003 through REQ-ACE-006) computes is the only non-block outcome.

- **REQ-ACE-014** (Ubiquitous): The gate inventory in `auto-semantics.md` §9 shall carry a disposition row for every gate named in §D.3 of this SPEC, classified using only the existing disposition vocabulary (KEEP / AUTONOMOUS / RETIRED / PRESERVED EQUIVALENT FORM / capability switch) and citing its doctrine source by file and section name — the C2 risk-tier scheme of the source research note is not imported.

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

The authoritative acceptance surface is `acceptance.md` §D (16 criteria,
`AC-ACE-001` … `AC-ACE-016`, Given-When-Then form, severity-classified, with
RED-now baselines where the defect is observable on the pre-implementation
tree). Every REQ in §B maps to at least one AC there.

### §D.3 Gate inventory rows to add (REQ-ACE-014 scope)

The following rows are added to `auto-semantics.md` §9. Each cites its source
by file and section name; the disposition is classified from the repo's own
doctrine at implementation time:

1. SPEC quality gate (plan Phase 15) — `.claude/skills/moai/workflows/plan/spec-assembly.md` § Phase 15: SPEC Quality Gate
2. Plan-audit negative/inconclusive verdict at run entry (the reconciled row replacing the retired bypass question) — `.claude/skills/moai/workflows/run/phase-execution.md` § Step 4 as rewritten by this SPEC
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

- The measured pre-existing drift between `.moi/config/sections/harness.yaml` and its template mirror, and between `phase-execution.md` and its mirror, predates this SPEC; this SPEC must not encode that drift as expected state, and repairing unrelated drift is not its scope.

## §F Dependencies and Related SPECs

- `SPEC-WF-AUDIT-GATE-001` — the audit gate (verdict enum, GateConfig.Invoke) this SPEC hardens.
- `SPEC-AUDIT-MODEL-CONVERGE-001` — the convergence result and required-gate semantics whose receipt this SPEC binds into the verdict file.
- `SPEC-AUDIT-SNAPSHOT-001` — hash-only cache validity and the plan-artifact hash subject set the admission predicate binds.
- `SPEC-SYNC-PARALLEL-DOCS-001` — origin (A6) of the `plan_audit_tier_ceilings` map this SPEC gives a Go reader.

## §G Risks

- **Receipt-absence false blocks**: trees that configure a required backend but run auditors that never emit receipts would block every run entry. Mitigated by REQ-ACE-008 (the auditor export mandate extends to receipt lines) and flagged for operator confirmation in decision-index Q4.
- **Double counting across families**: the same iteration recorded in both naming families could inflate the count. Mitigated by REQ-ACE-001's one-iteration-once rule (identity: verdict label + score + `audited_sha`/plan-artifact hash).
- **Verdict enum extension**: consumers switch on known `Verdict` values and default unknown ones to INCONCLUSIVE (`audit_gate.go` Step 4 default branch); adding a ceiling-blocked verdict value requires updating every switch, or the refusal silently downgrades.
- **Config symmetry audits**: the new Go structs must join `audit_struct_yaml_symmetry_test.go` or CI fails on the very keys this SPEC adopts.

## §H Amendments

- v0.1.0 (2026-10-04): initial plan-phase authoring (card t1500). No amendments.
