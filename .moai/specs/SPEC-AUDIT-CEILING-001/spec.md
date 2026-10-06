---
id: SPEC-AUDIT-CEILING-001
title: "Machine-enforced plan-audit repetition ceiling and required-backend admission"
version: "0.5.0"
status: draft
created: 2026-10-04
updated: 2026-10-06
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

The multi-backend convergence result (`ConvergenceResult`,
`internal/cli/mcp_convergence.go`) already computes per-required-backend
verdicts and an overall verdict, but the exported plan-audit verdict file
format (`.moai/docs/audit-artifact-convention.md` § What) carries no receipt of
that convergence, and the shared admission predicate
(`internal/auditverdict/verdict.go`) inspects only the auditor's own label,
score, must-pass, blocking, and hash fields. Operator-reported instances
(cards t1469, t1482; not re-verified from artifacts in this plan phase — see
`research.md` Gaps): a convergence `fail` on a required backend accompanied by
an auditor-own PASS was admitted at run entry.

This SPEC gives the ceiling teeth: a CLI-computed per-SPEC iteration counter,
a policy outcome decided without a question when the cap is reached, and a
hard block on run entry when a required backend fails regardless of the
auditor's own verdict. Operator decision D9 (card t1500, 2026-10-06) narrowed
this SPEC to exactly those three concerns after the plan phase consumed its
3+1 audit budget; the run-gate doc reconciliation this SPEC previously carried
(the `phase-execution.md` / `auto-semantics.md` §9 rewrites, former REQ-ACE-013
and REQ-ACE-014) is split off to a follow-up card, and the cross-card
re-audit counting refinement (former iter2 D22 machinery, iter3 D33) is
deferred with it — see §E and decision-index.md for the disposition record.

## §B Requirements (GEARS)

Requirement id prefix: `REQ-ACE` (Audit Ceiling Enforcement).

### §B.1 Work item 1 — per-SPEC audit-round counter and ceiling policy

- **REQ-ACE-001** (Ubiquitous): The CLI shall derive a SPEC's plan-audit round count from durable iteration evidence on disk — the iteration-scoped plan-audit verdict files of `.moai/docs/audit-artifact-convention.md` (`plan-audit-iter<N>.md` family) and the `<SPEC-ID>-review-<N>.md` iteration stream — counting one iteration once even when the same iteration appears in both families, and never from in-process memory. The dedupe identity is the pair (SPEC id, iteration number): two evidence files naming the same SPEC and the same iteration number are one round regardless of which family or card directory recorded them; different iteration numbers are different rounds even when the verdict label, score, and audited hash are identical. A file whose iteration number is unreadable is never collapsed into another file (fail-counted). A convention-family file belongs to the SPEC named in its report header; the counter resolves a SPEC's evidence from `.moai/reports/<SPEC-ID>/` plus every `.moai/reports/<card-id>/` directory whose plan-audit iteration files name that SPEC.

- **REQ-ACE-002** (Where the harness config declares `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`): The CLI shall read the tier ceiling and the policy values (`auto_delta_rounds`, `on_final_hit`) through typed Go config structs, validating `on_final_hit` against its documented value (`hold-and-split`; any other value is a config error — the CLI refuses to enforce a policy it cannot name) — the "no Go reader" orphan disposition of both keys in `internal/config/loader.go` is retired, and the config-struct/YAML symmetry audit shall cover the new structs.

- **REQ-ACE-003** (When a plan-audit verdict that fails the shared predicate's admission is presented for admission at a run-entry admission seam — the kickoff evaluator or the card-transition guard, the LIVE seams of design.md §5 — for a SPEC whose round count, the presented round included, has reached the effective ceiling — the tier ceiling plus `auto_delta_rounds`, per the composition Q2 records — or has reached the tier ceiling while no eligible auto delta round remains, where eligibility is the prose policy's own condition (`plan-auditor.md` § Retry Loop Contract): a `fix_scope`-anchored diff between the two audited SHAs and an unchanged REQ/AC id set; an ineligible delta or a STOP signal is a final hit): The CLI shall refuse the round's admission at that seam — the engine evaluates before the verdict reaches the admitting consumer, and a verdict that satisfies every admission check is not this requirement's subject: it follows REQ-ACE-013, never a refusal — and shall compute and emit the ceiling-policy outcome without asking any question.

- **REQ-ACE-004** (When the refusal of REQ-ACE-003 fires and the latest non-admitted verdict fails admission on the verdict label alone — `overall_score` at or above the tier threshold, `must_pass_failed` = 0, `blocking_count` = 0, and a `plan_artifact_hash` that binds the current plan artifacts — and enumerates at least one finding): The CLI shall produce the debt-admission outcome — record PASS-WITH-DEBT with the verdict's findings enumerated as debt lines (each carrying `dispose_in`), written by the CLI into its outcome record (never by rewriting the auditor's verdict file, which would duplicate decision keys), and admit run entry at the seam that consumes the outcome. Every other admission check keeps its refusing force: a verdict that fails the score threshold, a field check, a duplicate-key check, or hash binding is never debt-admitted, and neither is a verdict carrying a REQ-ACE-009 or REQ-ACE-010 receipt refusal — a required backend recorded fail or inconclusive, a receipt absent or missing a configured backend's line, or a malformed receipt line (work item 1 never converts what work item 2 refuses; the consuming seam re-runs the full predicate — design.md §2) (C5).

- **REQ-ACE-005** (When the refusal of REQ-ACE-003 fires and the latest non-admitted verdict records one or more blocking findings and every blocking finding carries a scoped fix anchor): The CLI shall produce the scope-split outcome — persist a hold record together with a split proposal naming the anchored scope, and keep run entry blocked.

- **REQ-ACE-006** (When the refusal of REQ-ACE-003 fires and the latest non-admitted verdict satisfies neither REQ-ACE-004 nor REQ-ACE-005 — including a verdict whose `plan_artifact_hash` does not bind the current plan artifacts, which holds and is never debt-admitted, and a verdict enumerating no findings, for which no debt line can be honestly written): The CLI shall produce the hold-record outcome — persist the hold record and keep run entry blocked. The hold record names its release path — the REQ-ACE-005 split/new-SPEC route or an operator decision recorded in `progress.md` §G — so a hold is never a silent dead end.

- **REQ-ACE-007** (Ubiquitous): The CLI shall report every ceiling refusal and every required-backend refusal it emits (REQ-ACE-009 through REQ-ACE-011) with the outcome, its reasons, and the evidence paths in its structured output, shall persist each to the SPEC's `progress.md` §G Override and Refusal Record (the same section REQ-ACE-011 writes — outside the plan-artifact hash subject set), and shall complete the entire path without an interactive prompt.

### §B.2 Work item 2 — required-backend failure blocks run entry

- **REQ-ACE-008** (When a plan-audit verdict file is produced from an audit whose tree resolves a non-empty required-backend set — an `audit.gates` entry or an `audit.model` token assignment, resolved through the `resolveAuditGates` resolution subject to the config-error disposition of REQ-ACE-010; a tree resolving no required backend requires no receipt, C4): The verdict file shall carry machine-readable receipt lines recording the convergence overall verdict and one line per required backend carrying that backend's verdict — whether the backend passed, failed, or was inconclusive — in the line format `.moai/docs/audit-artifact-convention.md` § What defines for this SPEC. The projection onto that two-valued `convergence_overall` and three-valued `required_backend` vocabulary is defined for every label the auditor's own verdict enum can carry (design.md §3): an own label the shared predicate's `AdmitLabel` reads as passing — PASS or PASS-WITH-DEBT — projects to `pass` for the backend it ran and `convergence_overall: pass`; FAIL or FAIL_WARNED projects to `fail`; INCONCLUSIVE projects to `inconclusive` on the backend line and `fail` on `convergence_overall` (a required backend recorded fail or inconclusive refuses under REQ-ACE-009/010, so the projection never upgrades silently — the raw own label stays readable in the verdict body, which the receipt never rewrites). The receipt producer shall cover both audit shapes the trigger names: a multi-model audit projects the convergence result's overall verdict and per-backend entries (design.md §3), and a single-model audit writes `convergence_overall` from its own verdict under the same projection plus one `required_backend` line for the backend it actually ran — a required backend the audit did not cover stays absent from the receipt and refuses under REQ-ACE-010 (correct fail-closed; iter3 D32).

- **REQ-ACE-009** (When the tree's audit gate configuration marks a backend required and the verdict's receipt records that backend as fail or inconclusive): The shared admission predicate (`internal/auditverdict`) shall refuse the verdict regardless of the verdict's own label — an auditor-own PASS never overrides a required backend's fail, and an unresolvable required backend is not a pass.

- **REQ-ACE-010** (When a required backend is configured and the verdict carries no convergence receipt, or a receipt that omits a configured required backend's line; or when the tree's audit configuration exists but cannot be read or parsed): The admission predicate shall refuse the verdict — the fail-closed default per `auto-semantics.md` §7 (disposition confirmed or amended by decision-index Q4). A configuration error is not an empty configuration: a resolution that errors refuses, and the `resolveAuditGates` behavior of folding a resolution error into an empty gate set is not the disposition of this SPEC; only a resolution that genuinely yields no required backend admits a receipt-less verdict (C4).

- **REQ-ACE-011** (When an operator overrides a required-backend refusal): The CLI shall accept only an explicit override input that names the backend and carries an acknowledgement note, shall itself write the override record to the SPEC's `progress.md` §G Override and Refusal Record — a section outside the plan-artifact hash subject set; never `decision-index.md`, whose edit would break the hash binding of every later verdict — and to the audit-trail log of REQ-ACE-012, and shall provide no silent-skip path.

- **REQ-ACE-012** (Ubiquitous): The CLI shall append an audit-trail log entry for every ceiling refusal (REQ-ACE-003) and every required-backend refusal or override (REQ-ACE-009 through REQ-ACE-011); the trail is machine-local state (design.md §5), and the durable carrier of every refusal is the `progress.md` §G record REQ-ACE-007 writes.

### §B.3 Cross-cutting requirements

- **REQ-ACE-013** (When a plan-audit verdict that satisfies every admission check of the shared predicate — label, score, must-pass, blocking, hash binding, duplicate keys, and the REQ-ACE-009/010 receipt checks — is presented for admission at a run-entry admission seam for a SPEC at any ceiling state — the effective ceiling reached (the tier ceiling plus `auto_delta_rounds`), or the tier ceiling reached on a final hit with no eligible delta round): The CLI shall admit the round at that seam — an admission-clean verdict is never held at any ceiling state — and shall record a ceiling-reached outcome in its outcome record — the ceiling caps repetition, not a healthy result; a fully-passing verdict presented at any ceiling state is never held by omission (iter3 D31; v4-iter1 V4-D1's second face closed here) — and shall complete the path without asking any question.

- **REQ-ACE-014** (Ubiquitous): Every deployed-file edit this SPEC makes — rules, skills, docs, and config sections — shall land in its `internal/template/templates/` mirror in the same change.

- **REQ-ACE-015** (Ubiquitous): Where this SPEC enforces a ceiling or a refusal, the enforcement shall be a CLI decision (a nonzero exit status or a recorded refusal), never prose-only guidance.

## §C Constraints

- **C1 — one policy, two consumers.** The ceiling-policy semantics stay consistent with `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier escalation paragraph and `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract; this SPEC adds the machine consumer of the existing policy keys, not a third rule text. Where the prose and the CLI disagree after implementation, the CLI is authoritative and the prose is corrected in the same change.
- **C2 — Template-First.** Constraint REQ-ACE-014 applies to every milestone.
- **C3 — no interactive prompt in the CLI path.** The CLI runs in subagent context; the refusals of REQ-ACE-003/009/010 are machine outputs, and any operator decision arrives through the orchestrator, not through the CLI.
- **C4 — backward compatibility.** A SPEC whose frontmatter carries no `tier:` field is treated as Tier L (existing rule); the legacy `<SPEC-ID>-review-<N>` naming stays counted beside the convention family; existing verdict files without receipt lines remain admissible wherever no required backend is configured.
- **C5 — admission thresholds untouched.** The per-tier plan PASS thresholds (S 0.75 / M 0.80 / L 0.85), the must-pass and blocking-count semantics, and the existing `Admit` checks keep their meaning; this SPEC adds refusal causes, the receipt fields, and two bounded admission outcomes — the REQ-ACE-004 debt-admit outcome (converts only label-only failures at the ceiling) and the REQ-ACE-013 pass-through outcome (admits a fully-passing verdict at the ceiling) — both disclosed in §G.

## §D Success Criteria

### §D.1 Acceptance criteria

The authoritative acceptance surface is `acceptance.md` §D (21 criteria,
`AC-ACE-001` … `AC-ACE-021`, Given-When-Then form, severity-classified, with
RED-now baselines where the defect is observable on the pre-implementation
tree and explicit new-test declarations otherwise). Every REQ in §B maps to at
least one AC there. The v0.4.0 id map (former → current, for history
traceability): REQ-ACE-001..012 and REQ-ACE-013 (new, the D31 pass arm) keep
their subjects; former REQ-ACE-015 → REQ-ACE-014 (mirrors); former REQ-ACE-016
→ REQ-ACE-015 (CLI decision); former REQ-ACE-013/014 (run-gate doc
reconciliation) are deleted by the D9 scope cut; former AC-ACE-015 → AC-ACE-014,
former AC-ACE-022 → AC-ACE-015, former AC-ACE-013/014 (doc-text criteria) are
deleted; AC-ACE-016..021 keep their numbers and subjects.

## §E Out of Scope

### Out of Scope — audit quality thresholds

- The per-tier plan PASS thresholds (S 0.75 / M 0.80 / L 0.85), the must-pass criterion semantics, the blocking-finding semantics, and dimension scoring are not changed by this SPEC.

### Out of Scope — auditor iteration behavior

- The plan-auditor's STOP-on-score-regression signal, delta-audit eligibility rules (`fix_scope` anchor arithmetic), and report format are not changed; this SPEC consumes their outputs, it does not retune them.

### Out of Scope — run-gate doc reconciliation (former work item 3, split off per D9)

- The `phase-execution.md` Step 4 FAIL/INCONCLUSIVE question-branch rewrite and the `auto-semantics.md` §9 gate-inventory expansion (former REQ-ACE-013/014) are follow-up card material. This SPEC leaves the two governing texts as they stand; reconciling them into one fail-closed rule is the follow-up's job. Consequently this SPEC does not edit `phase-execution.md`, `auto-semantics.md`, or their mirrors.

### Out of Scope — cross-card re-audit counting refinement (deferred per D9)

- The audited-state dedupe resolution and the cross-card never-collapsed rule (former iter2 D22 machinery, iter3 D33) are deferred: under REQ-ACE-001's (SPEC id, iteration number) identity, a no-repair re-audit recorded under a new card at a reused iteration number collapses into the earlier round and does not advance the count. This ceiling blindness to audit-without-repair churn is a known, accepted limitation of the narrowed scope, recorded in §G risk 6 and design.md §1, to be revisited by a follow-up card.

### Out of Scope — queue admission and push policy

- Nothing in this SPEC touches gtd/kanban queue admission, card pick, push batching, or PR policy.

### Out of Scope — sync-audit ceilings

- Only the plan-audit repetition ceiling is machine-enforced; a sync-audit iteration ceiling is a separate future concern.

### Out of Scope — pre-existing mirror drift

- The measured pre-existing drift between `.claude/agents/moai/plan-auditor.md` and its template mirror predates this SPEC (measured DIFF at f2f815008 and re-measured DIFF at 69a085b2d, research.md §3); this SPEC must not encode that drift as expected state, and repairing unrelated drift is not its scope.

### Out of Scope — the plan_audit_global orphan key

- The third acknowledged config orphan, `plan_audit_global` (`internal/config/loader.go:344`), keeps its "no Go reader" disposition; only the two ceiling keys of REQ-ACE-002 are given a reader by this SPEC.

## §F Dependencies and Related SPECs

- `SPEC-WF-AUDIT-GATE-001` — the audit gate (verdict enum, the `GateConfig` library) this SPEC hardens; its `Invoke` seam carries no production caller today (research.md §2), so this SPEC's enforcement lands at the live admission seams (the kickoff evaluator and the card-transition guard).
- `SPEC-AUDIT-MODEL-CONVERGE-001` — the convergence result and required-gate semantics whose receipt this SPEC binds into the verdict file.
- `SPEC-AUDIT-SNAPSHOT-001` — hash-only cache validity and the plan-artifact hash subject set the admission predicate binds.
- `SPEC-SYNC-PARALLEL-DOCS-001` — origin (A6) of the `plan_audit_tier_ceilings` map this SPEC gives a Go reader.

## §G Risks

- **Receipt-absence false blocks**: trees that configure a required backend but run auditors that never emit receipts would block every run entry. Mitigated by REQ-ACE-008 with the producer assigned in plan.md M1 (the plan-auditor agent body's export step writes the receipt lines — from the convergence result for multi-model audits, from its own verdict and the backend it ran for single-model audits, per iter3 D32) and flagged for operator confirmation in decision-index Q4.
- **Double counting across families**: the same iteration recorded in both naming families could inflate the count. Mitigated by REQ-ACE-001's one-iteration-once rule (identity: the (SPEC id, iteration number) pair — label, score, and hash are NOT identity components; different iteration numbers always count, and the same iteration number collapses across families and across card directories alike).
- **Verdict enum extension**: consumers switch on known `Verdict` values and default unknown ones to INCONCLUSIVE (`audit_gate.go` Step 4 default branch); adding a ceiling-blocked verdict value requires updating every switch, or the refusal silently downgrades.
- **Label conversion masking auditor judgment**: the debt-admit outcome records PASS-WITH-DEBT over the auditor's own non-passing label. Bounded three ways: only a label-only admission failure qualifies (score, field, duplicate-key, and hash checks keep refusing), at least one finding must exist to enumerate as debt, and the outcome record names the converting rule so the auditor's original label stays readable beside it.
- **Pass-through at the ceiling admitting an unhealthy SPEC**: the REQ-ACE-013 arm admits any verdict that clears the full predicate at any ceiling state. The exposure is bounded by the predicate itself — a verdict that clears score, must-pass, blocking, hash, and receipt checks is admission-clean on its own merits; the ceiling-reached outcome record names the SPECs that entered this way so the population stays auditable.
- **Config symmetry audits**: the new Go structs must join `audit_struct_yaml_symmetry_test.go` or CI fails on the very keys this SPEC adopts.

## §H Amendments

- v0.5.0 (2026-10-06): plan-audit v4-iter1 repair (card t1500; fresh-series iter1, FAIL 0.81) — closed the defect delta of `.moai/reports/t1500/plan-audit-v4-iter1.md` (3 blocking + 2 optional), delta-scoped to the verdict's fix_scope (spec.md#REQ-ACE-003/#REQ-ACE-008/#REQ-ACE-013, acceptance.md#AC-ACE-003/#AC-ACE-008/#AC-ACE-015, design.md#s2-ladder/#s3-receipt-schema). V4-D1 closed: REQ-ACE-003's trigger scoped to verdicts that FAIL the shared predicate's admission (a verdict satisfying every admission check follows REQ-ACE-013, never a refusal — design §2's ladder is now the normative reading in the REQ text); REQ-ACE-013 extended to any ceiling state (effective ceiling, or tier ceiling on a final hit) closing the second face — the D31 defect class no longer survives at the tier-ceiling final-hit boundary; AC-ACE-003 and AC-ACE-015 carry the pass-through exclusion arm. V4-D2 closed: the receipt projection rule defined in REQ-ACE-008 and design.md §3 for every label of the auditor's own verdict enum ({PASS, PASS-WITH-DEBT} → pass; {FAIL, FAIL_WARNED} → fail; {INCONCLUSIVE} → inconclusive on the backend line, fail on convergence_overall) — no silent upgrade, raw label stays readable in the verdict body; AC-ACE-008 carries the PASS-WITH-DEBT end-to-end arm (`TestParseReceiptPassWithDebtProjection`). V4-D3 closed: AC-ACE-008's export-path verification split into per-arm asserted counts — the multi-model projection instruction grep AND a new single-model instruction grep (`backend it actually ran`) each with its own RED-now cell, so removing either arm turns the criterion red. Optionals folded: V4-O1 (design.md §1's "i.e. at max-N" gloss qualified "under contiguous numbering"), V4-O2 (progress.md interlock pointer Q2-Q6 → Q2-Q5, noting Q0 decided / Q6 fell away). No REQ id changed (15 REQs); the AC count stayed 21 (extensions, not additions).
- v0.4.0 (2026-10-06): re-plan under operator decision D9 (card t1500) after the plan phase consumed its 3+1 audit budget (iter1 FAIL 0.69 → iter2 FAIL 0.83 → iter3 FAIL 0.81 + STOP). Scope narrowed to the card's three concerns: the CLI-side per-SPEC audit counter (REQ-ACE-001/002), the single no-question policy-outcome path at the cap (REQ-ACE-003..007), and required-backend fail blocking run entry (REQ-ACE-008..012). iter3 defect dispositions: D31 fixed in-scope (new REQ-ACE-013 pass-through arm + AC-ACE-013), D32 fixed in-scope (single-model receipt producer arm in REQ-ACE-008 + AC-ACE-008), D33 deferred (REQ-ACE-001's identity simplified back to the plain (SPEC id, iteration number) pair — the D22 audited-state machinery and the never-collapsed clause are removed by this scope cut; the contradiction is dissolved by simplification, the ceiling-blindness it exposed is recorded as §G/§E accepted limitation), D34 fixed in-scope (§G risk 2 rewritten to the simplified identity), D35/D36 dropped with their surfaces (former REQ-ACE-013/014 doc-reconciliation criteria deleted). Scope cut removes the phase-execution.md and auto-semantics.md edits (former REQ-ACE-013/014, §D.2 row list, AC-ACE-013/014 and their ledgers); REQ ids renumbered contiguously 001-015 (former 015→014, 016→015, new 013 = pass arm); AC ids 001-021 (former AC-ACE-015→014, AC-ACE-022→015, new AC-ACE-013; former AC-ACE-013/014 deleted). Edge case 9 (cross-card same-N different states) removed with the deferred surface. Baselines re-measured in this tree at 69a085b2d. Tier stays L (multi-subsystem, gate-semantics; REQ 15/25, AC 21/25 within L ceilings) — re-tier judgment recorded in progress.md §E.1.
- v0.3.0 (2026-10-04): plan-audit iter2 repair (card t1500) — closed the defect delta of `.moai/reports/t1500/plan-audit-iter2.md` (FAIL 0.83; 6 blocking D19-D24, 6 optional D25-D30): D19 receipt producer assigned (plan.md M1 — the plan-auditor agent body's export step writes the receipt lines from the convergence result it already receives, mirror included; export-path AC added to AC-ACE-008), D20 the REQ-ACE-009/010 receipt refusals joined REQ-ACE-004's never-debt-admitted set and the consuming seam's admission semantics stated (the seam re-runs the full predicate with one label conversion — design.md §2; AC-ACE-004 negative arm), D21 config-error disposition named fail-closed (REQ-ACE-010 third trigger arm; design.md §4; AC-ACE-010 invalid-config arm + `TestAdmitConfigErrorRefused`), D22 dedupe identity resolved within one audited state so cross-card renumbering cannot stall the count (REQ-ACE-001; design.md §1 cross-card rule; AC-ACE-001 cross-card arm; §C edge 9), D23 REQ-ACE-003's trigger reworded to the admission-seam event the design delivers and the kickoff-evaluator seam given its AC arm (AC-ACE-003/AC-ACE-022), D24 decision-index Q5 rewritten to the label-only predicate. Optionals folded: D25 (§9 total-count + per-member vocabulary greps, LEDGER-ACE-013-A/014-B), D26 (EnvSkipAudit disposition pointer → §D.2 row 2; M4/design §9 restatement scoped to the question branches), D27 (five stale cross-references), D28 (package-wide selector listing + §A exception note), D29 (AC-ACE-008/011 reclassified RB with E8 declarations; 005/012 RG rationale stated), D30 (hold-release path in REQ-ACE-006; C5 wording; REQ-ACE-007 extended to required-backend refusals; REQ-ACE-012 "durable" corrected). No REQ id changed (16 REQs); the AC count stayed 22 (extensions, not additions).
- v0.2.0 (2026-10-04): plan-audit iter1 repair (card t1500) — closed the defect delta of `.moai/reports/t1500/plan-audit-iter1.md` (FAIL 0.69; 18 findings, 13 blocking): D2 debt-admit re-gated to label-only failures (hash mismatch and no-findings verdicts hold; AC-ACE-017), D3 receipt projects every required backend's verdict with the REQ-ACE-008 trigger moved to required-configured trees (AC-ACE-018/019), D4 refusal rerouted to the live admission seams (`GateConfig.Invoke` measured production-dead; AC-ACE-022), D5 dedupe identity = SPEC id + iteration number (AC-ACE-020), D6 delta eligibility encoded (AC-ACE-021), D7 at-ceiling condition added to REQ-ACE-004..006, D8 one negative-verdict outcome vocabulary naming the override, `EnvSkipAudit`, and FAIL_WARNED, D9 AC-ACE-015 rescoped region-scoped with a named known-FAIL list and research §3 corrected, D10 override record target named (`progress.md` §G, outside the hash set), D11 M1 override input removed, D12 AC-ACE-014 replace arithmetic (10 − 1 + 11 = 20) with per-named-row verification, D13 `.moi/` → `.moai/` (×3), D14 AC-ACE-016 remapped to REQ-ACE-007 and AC-ACE-022 added for REQ-ACE-016, D15 M3 gated on Q2/Q3/Q5 with kickoff-amendable defaults, D16 §D.3 renumbered §D.2, D17 `on_final_hit` validated on load and the M2 harness.yaml mirror step dropped, D18 per-key receipt repeat rule stated. No REQ id changed (16 REQs); the AC set extended 16 → 22 (Tier L ceiling 25).
- v0.1.0 (2026-10-04): initial plan-phase authoring (card t1500).
