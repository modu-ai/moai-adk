# progress.md — SPEC-AUTONOMY-GATE-REWIRE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.3.5 (2026-09-27, plan-audit iter-5 delta FAIL 0.841 repaired in place — D47-D55, plan.md §L; report .moai/reports/t1236/plan-audit-5.md); prior 0.3.4 (2026-09-27, M0 re-anchor per lead decisions B1-B5 — plan.md §K, research.md §10.6; REQ 25 / AC 25 unchanged); prior 0.3.3 (operator-approved one-time 4th audit exception limited to D43-D46; base 710530d67); prior 0.3.2 (revision base 1b071a573 = v0.3.1; plan-audit iter-2 FAIL 0.77 → `.moai/reports/t1236/plan-audit-2.md`, dispositions D26-D42 in plan.md §J; iter-1 dispositions in plan.md §I)
- tier: L (spec.md, plan.md, acceptance.md, design.md, research.md)
- requirements: 25 (REQ-GR-001..025, contiguous, no new IDs in v0.3.2) / acceptance criteria: 25 (AC-GR-001..025)
- a1_reference_baseline: 25283ebf8 (SPEC-AUTONOMY-CONTRACT-001 0.5.2, schema owner; conditional interim-rule wording, A3 owns the AC-CONTRACT-016 (t) replacement test); history 8f77d9a33 / 98cb7879d / 4208a3a3b / 652243c72 / 6d98ca466 / 67a2f55cb / 65e0a9167
- a2_reference: A2 0.4.3 (implemented, BASE 7fe658815) — escalation/<class>-<fingerprint>.md + YAML header + revoke kind confirmed at M0; revoke records are status: resolved in A2, so resume blocking is owned by the A3 revoke reader (v0.3.4, B3); earlier withdrawn JSON format d8926ff9a is history
- a2b_reference: t1245 owns push serialization, agent-origin sign deny, decide refusal in MOAI_FACTORY_ROLE=agent sessions (SPEC not yet written; marker kept)
- store: $MOAI_HOME/db/<project-key>/contract/{receipts.jsonl,events.jsonl} (lead decision R10)
- deciders: guided=human; contract autonomous = llm (default) | llm+jev; jev refused (R5); outcome in {approve, reject, human} derived by A3 rules R1-R5; reject and human both = sign refused, human decision required
- linkage (REQ-GR-025): Jev-principle amendment (incl. CLAUDE.local.md §29) + R3 release + A1 interim-rule release in ONE commit via single constant jevDoctrineAmended; assembly per lead decision 2026-09-26 (D32): manager-spec then manager-develop, lane orchestrator commits by explicit pathspec; whole commit held without operator §29 confirmation (D33); proposed §29 text in design.md §11.1
- open_clarifications: none (defaults D-2, D-4 recorded; operator may override)
- lead_confirmations_pending: A1 acceptance of the jevDoctrineAmended gating of signer step (1) (marker, research.md §10.4); a producer of `plan_artifact_hash:` for card-path plan-audit reports before M8 (plan.md §C)
- run_preconditions: t1234 (A1 >= 0.5.2), t1235 (A2, with final escalation format), t1245 (A2b) merged into develop; t1175 merged and absorbed; BASE recorded in §E.2; pre-flight: re-measure reject/human consumers after the A2b SPEC lands
- operator_confirmation_29: approved 2026-09-26, text verbatim as design.md §11.1 (relayed by lead); edit lands only on the develop copy of CLAUDE.local.md in the card worktree, via the develop merge
- autonomous_kickoff_activation: see design.md §7.1 (single source)
- Implementation Kickoff Approval: not requested at plan phase

## §E.2 Run-phase Evidence

### M0 — absorb and re-anchor (2026-09-27, manager-develop) — HALTED with blocker report

- BASE: `7fe658815eb0d4110b9acadad56e5a85bee3ed3f` (`git merge-base develop HEAD`; last absorbed develop commit, t1175 close). Branch HEAD at measurement: `016a3965e` (merge of develop into WT-contract-gate-rewire). `MOAI_GR_BASE=7fe658815`.
- Preconditions (plan.md §C): A1 `SPEC-AUTONOMY-CONTRACT-001` status completed v0.5.2 (internal/contract present); A2 `SPEC-AUTONOMY-ESCALATION-001` status implemented v0.4.3 (internal/escalation present); A2b `SPEC-AUTONOMY-PRECONDITION-001` status completed v0.1.3 (internal/hook/contract_sign_guard.go carries the decide guard); t1175 `SPEC-ALWAYS-LOADED-DIET-002` status completed v0.10.0, absorbed.
- research.md §1.1 recount on this tree: local 33 files / 116 lines, template 32 files / 111 lines (identical to plan-phase); per-file counts identical to §1.2 table; Kickoff-line diff local↔template across the 32 paired files: 0. E/R/H classification unchanged.
- Anchors: all skill-file anchors (rows 6-14) present. Two cited anchors moved by t1175 — `askuser-protocol.md` `### The Five Exceptions` (now `askuser-protocol-reference.md:229`; `askuser-protocol.md` keeps only a 1-paragraph `## Ambiguity Triggers and Exceptions` stub at :206) and the `jev_ask` row of `moai-mcp-tools.md` (row now only in `moai-mcp-tools-catalogue.md:138,216`).
- A2 cross-check: record path/YAML frontmatter/`kind: revoke` match the lead-stated final format, BUT A2 spec §I.2 states a revoke record carries `status: resolved` and never makes a card needs-decision, and `escalation.NeedsDecision` counts only `contract`/`operational` kinds — contradicts REQ-GR-022/AC-GR-023 (revoke → needs-decision, "open record" via the A2 reader) and the REQ-GR-012 second-witness path.
- A1 cross-check: signer step (1) (interim rule) lives in `internal/contract/receipt.go` `ReceiptOutcome` (package `contract`, L208-231), not in `internal/contract/sign/` (design.md §2 row 23 allowlist).
- Constitution baseline: `moai constitution validate` on this tree → exit 1, `FAILED — 9 error(s)` (DRIFT CONST-V3R2-013/014/015/016/017 on CLAUDE.md §7, 033, 049, 152, 153); same with `go run ./cmd/moai`. AC-GR-003's second command (expects exit 0 `OK`) is red at BASE for reasons outside this SPEC's edit allowlist.
- Always-loaded sizes at BASE (`LC_ALL=en_US.UTF-8 wc -m`): CLAUDE.md 14039, askuser-protocol.md 18684, goal-directive.md 5628, moai-mcp-tools.md 3841; orchestration-mode-selection.md 37140.
- Disposition: M0 stop rule fired (cited anchors gone + A2/A1 differences). No M1+ work started; blocker report returned to the orchestrator for manager-spec re-delegation.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
