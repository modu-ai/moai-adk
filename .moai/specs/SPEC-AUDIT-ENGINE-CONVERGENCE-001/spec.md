---
id: SPEC-AUDIT-ENGINE-CONVERGENCE-001
title: "Audit engine convergence repairs — fix-confirm delta by default with a verified delta chain, unrun legs classified instead of failed, receipts that survive evidence-only commits, one operator path to clear a rejection, honest auditor prompts, and a repair-scope rule"
version: "0.5.2"
status: draft
created: 2026-11-03
updated: 2026-11-03
author: manager-spec (card t2017)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/auditreceipt, internal/auditverdict, internal/runtime, internal/closure, internal/homestate, internal/template/agentemit/source, .moai/docs"
lifecycle: spec-anchored
tags: "audit, convergence, delta-scope, delta-chain, second-review, receipt-binding, rejection, operator-approval, inconclusive, repair-scope, card-t2017"
tier: M
card: t2017
related_specs: [SPEC-MCP-NEUTRALITY-001, SPEC-AUDIT-CEILING-001, SPEC-AUDIT-CEILING-002, SPEC-AUDIT-MODEL-CONVERGE-001, SPEC-AUDIT-CEILING-REPAIR-001, SPEC-AUDIT-GATE-INTEGRITY-001, SPEC-AUDIT-BUILD-IDENTITY-001, SPEC-WF-AUDIT-GATE-001, SPEC-RECEIPT-REUSE-001, SPEC-WORKTREE-STATE-ROOT-001, SPEC-AUTONOMY-CLOSURE-001]
---

# SPEC-AUDIT-ENGINE-CONVERGENCE-001 — Audit engine convergence repairs

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-11-03 | manager-spec (card t2017) | Initial plan-phase draft. Six repair items (B1-B5, A1). Every code citation was re-read on tree `cb30e05b3e5f64774b8dabd87e80c24332f4551d`; mismatches are recorded in §B. |
| 0.2.0 | 2026-11-03 | manager-spec (card t2017) | Leader ruling R287 applied. Q1: the delta default stands; on the `card_id` path a delta round can pass only on a verified chain plus a BLOCKING-closure confirmation, and a gated delta call files a second-review record (REQ-AEC-004, -005). Q3: `INCONCLUSIVE` label with its consumer rules (REQ-AEC-007). Q6/Q7: CLI verb on the human path, no MCP path; the hook deny is recorded as a Gap. B3: the reports allowlist is narrowed to the audit-artifact filename family, and `progress.md` §E.2/§E.3 edits are pinned as negative cases (REQ-AEC-009, -010). REQ 14 → 16, AC sub-criteria 15 → 16; both are now at the Tier M ceiling. |
| 0.3.0 | 2026-11-03 | manager-spec (card t2017) | Leader ruling R288 applied. The B3 negative case narrows to `progress.md` §E.2: a rewrite of §E.3 after a PASS keeps the receipt (REQ-AEC-009, -010; pinned inside AC-AEC-003a/003b, counts unchanged at 16/16). The Q3 reading of the round counter is confirmed. The cooperating-process threat model and the auditor-written BLOCKING ids are recorded as residual risk (§E). B4 stays in this SPEC. |
| 0.4.0 | 2026-11-03 | manager-spec (card t2017) | Plan-audit round 1 repair (`.moai/reports/t2017/plan-audit-1.md`, FAIL 0.75, audited at `36148aeec9`). D1: every chain record, the starting record included, needs a pass or fail from every required backend and a receipt that did run; no FAIL verdict means full scope; an explicit `delta_base` with a `card_id` goes through the same gate. D2: the closure set is the union of the BLOCKING ids of every FAIL verdict tied to the chain; both auditor sources must write `blocking_count` and `blocking_defect:` lines on a FAIL. D3/D4: the four `INCONCLUSIVE` consumers, the last-line form, and a reconciliation with SPEC-AUDIT-CEILING-001 REQ-ACE-008 and SPEC-AUDIT-CEILING-002 REQ-ACR-005 (§G). D5: B5 targets the agent sources and `make agents-emit`. D6: REQ-AEC-014 narrowed to "no backend reported a defect". D7: EV-10 re-measured. D8: the session result gains head, digest and scope; a chain-admitted delta persists. D9: B4 residual risk restated; marker set made identical. D10: the §E.2 boundary on both versions; heading changes stale; only the cited SPEC's `progress.md`. D11-D13: the codex delta leg stays adversarial, citation gaps, `Latest` defined. Leader ruling R290: this card is in the rc.39 integration candidate. REQ and AC counts unchanged at 16/16. |
| 0.5.0 | 2026-11-03 | manager-spec (card t2017) | Plan-audit round 2 repair (`.moai/reports/t2017/plan-audit-2.md`, delta audit of 0.4.0, FAIL, audited at `cc100c2909`), under leader ruling R308. F1 (accepted): `make agents-emit` renders the plan-auditor and sync-auditor only into the template sets (`internal/template/agentemit/sets.go:48-50`, `:62-63`), so M6 updates the two local `.claude/agents/moai/` definitions by hand to match the corrected sources, and AC-AEC-005 checks them; the local definitions were re-measured (EV-20). F2 (accepted): a closure verified at an earlier chain link is inherited; only closures new since the previous link are checked against the current delta (REQ-AEC-004 (ii), -005; AC-AEC-001d fixture and mutant). F3 (accepted): a record with a required backend entry absent counts as not run (REQ-AEC-004 (i); AC-AEC-001c fixtures and mutant). F4 (accepted as a wording repair; the order claim was not reproduced): the multi-file `grep -c` notes now say the per-file order may vary, record where it was observed, and claim no cause. REQ and AC counts unchanged at 16/16. |
| 0.5.1 | 2026-11-03 | manager-spec (card t2017) | Leader ruling R313 applied: the local Codex copies `.codex/agents/moai/plan-auditor.toml` and `sync-auditor.toml` are in scope. All four local definitions (the two `.claude/agents/moai/` Markdown files and the two `.codex/agents/moai/` TOML files) are updated by hand and checked by AC-AEC-005 (REQ-AEC-014, §E B5, plan M6). EV-20 re-measured over the four files on the committed tree `cc100c2909` with `git grep`; each TOML copy carries the same two wrong lines as its `.claude` counterpart and no `blocking_defect:` line. A mutant row now fails a fix that touches only the `.claude` copies. The `CarriedFour` expansion stays t2018 input. REQ and AC counts unchanged at 16/16. |
| 0.5.2 | 2026-11-03 | manager-spec (card t2017) | Plan-audit round 3 repair (`.moai/reports/t2017/plan-audit-3.md`, delta audit round 2 of 2, FAIL, audited at `44b21e558e`), sized by leader ruling R320. G1 (accepted): inheritance of a closure is decided per naming, not per id; a later FAIL that names an id again voids its earlier inheritance, and the id then needs a closing commit inside the current `delta_base..HEAD` after the latest naming (one sentence in REQ-AEC-004 (ii); plan.md M1 (line 63) swept to match, per verification-completeness.md §3). AC-AEC-001d gains a re-report fixture, and a mutant row fails an implementation that inherits by id regardless of a later re-naming. F1, F3, F4 text unchanged. REQ and AC counts unchanged at 16/16. |

## §A Background

Ten audit rounds on card t2012 (PR #1965) stalled honest work. The causes were in the audit engine, not in the work under audit:

- A fix-confirm re-audit re-reviewed the whole card instead of the repair, so findings that predated the repair came back every round. The engine already has an explicit `delta_base` input (card t1708), but nothing fills it in, and on `audit_multi` the codex leg is told the delta scope in prompt prose only.
- A required leg that could not run (quota, timeout, authentication) ended the round as if a defect had been found. The exported verdict-file vocabulary (`verdict: <PASS|PASS-WITH-DEBT|FAIL>`) has no non-FAIL label for "the review did not happen", and both auditor prompts tell the auditor that an unmet gate yields `verdict: fail`. `INCONCLUSIVE` already exists elsewhere — the run-phase audit gate's verdict stream (`internal/runtime/audit_gate.go:37`, REQ-WAG-007) and the plan-auditor's own label, which SPEC-AUDIT-CEILING-001 REQ-ACE-008 projects to `fail` on `convergence_overall` (`.claude/agents/moai/plan-auditor.md:224`) — but not in the exported verdict file.
- An audit receipt went stale the moment the lane committed its own audit report or progress record, because the freshness check compares the head commit exactly.
- A rejection mark can be cleared only by the same auditor role producing a proven PASS. There is no operator path, so a rejection caused by an engine fault blocked every phase-entry spawn until the engine itself recovered.

The operator decided on 2026-11-03 to repair six defects. The leader relayed the follow-up product rulings as leader rulings R287 and R288 (`decision-index.md`). Per leader ruling R290 this card is included in the rc.39 integration candidate.

## §B Cited-location verification

Read on `cb30e05b3e`; the round-4 additions were read on `36148aeec9`, whose code equals `cb30e05b3e` (that commit adds only this SPEC directory). The leader accepted the round-1 corrections in R287.

| Item | Claimed location | What the code shows | Result |
|---|---|---|---|
| B1 | `internal/cli/audit_delta_scope.go:27` | The `delta_base` tool-schema description: explicit optional input; codex adversarial and the `audit_multi` codex leg are "told in the prompt only". No default exists. | Match |
| B1 | `internal/cli/audit_delta_scope.go:144-157` | Lines 144-147 are the tail of `deltaScopeNote`; 149-155 is `codexDeltaScopeClause`. The prose-only text itself is at line 106. | Match (range straddles two functions) |
| B1 | `internal/cli/mcp_convergence.go:1059` (ruling R287) | `if cfg.CardID != "" && cfg.DeltaBase == ""` around `appendSecondReviewRecord`. A delta call files no second-review record. | Read; REQ-AEC-005 changes it |
| B1 | `internal/cli/mcp_convergence.go:1046` | The session file persists a delta call only when its overall verdict is fail (card t1873: "a delta pass written here would open the gate after a failed full review"). The persisted `ConvergenceResult` carries no head commit, tree digest, or scope (`:128-170`). | Read; REQ-AEC-005, -008 |
| B1 | (not cited) | The `audit_multi` codex leg is `performCodexAudit` (`mcp_convergence.go:631`), which always sends a `turn/start` adversarial turn. `codex_audit` native mode binds a target only outside adversarial mode (`mcp_codex_audit.go:2323-2330`). | Added context; REQ-AEC-003 keeps the adversarial turn |
| B1 | (not cited) | `SelectSecondReview` (`internal/closure/readiness.go:109`) applies five filters in order — bound (`:138-141`), same contract (`:144-146`), scope covered (`:150-155`), a second model with a pass or fail (`:158-160`), in history (`:165-168`) — then picks the newest by `recorded_at`. It has no notion of a delta chain. | Added context; the chain check lives on the writing side (REQ-AEC-004) |
| B2 | `internal/cli/mcp_convergence.go:1123-1155` | `enforceRequiredGateUnmet` keeps an unmet required leg at overall `inconclusive` (REQ-MN-021) unless the overall already carries a reported fail (`:1144-1146`). | Mismatch: this code does not label the leg FAIL |
| B2 | `internal/auditverdict/verdict.go:361` | `if v != "pass"` in `admitReceipt` refuses a required backend recorded `fail` or `inconclusive`. `admitBase` runs first and refuses `required_backend_fail:` lines (`:289-291`) and `convergence_overall: fail` (`:297-299`) before any backend line is read. `convergence_overall` accepts only `pass` or `fail` (`:155`); `required_backend` is split once at the first space (`:168`). | Partial: refusal, no error class; reconciled in §G |
| B2 | `internal/auditreceipt/store.go:1563` | The `CauseReceiptAuditNotRun` arm (a receipt with a non-empty `gate_unmet` or an inconclusive codex verdict), distinct from `CauseReceiptVerdictFail`. | Partial: distinct cause exists, no error class |
| B2 | (not cited) | Verdict-label consumers: the exported file line (`.moai/docs/audit-artifact-convention.md:87`); the Stop-hook last line `^AUDIT-VERDICT: (PASS\|PASS-WITH-DEBT\|FAIL) …` (`internal/auditreceipt/store.go:173`); the factory card reader `^verdict: (PASS\|PASS-WITH-DEBT\|FAIL)$` (`internal/homestate/card_evidence_readers.go:20`); two round counters (`internal/runtime/audit_counter.go:10`, `internal/runtime/audit_ceiling.go:1186-1191`); the machine-line templates in both auditor sources. `AdmitLabel` (`verdict.go:219-221`) admits only PASS-family labels. | Added context; REQ-AEC-007 |
| B3 | `internal/auditreceipt/store.go:1605` | `head, digest, sErr := TreeState(r.TreeRoot)`; line 1608 compares `r.HeadCommit != head \|\| r.TreeDigest != digest`. Any commit moves both. The check (`CheckCitedReceiptsSince`, `:1514`) receives no SPEC ID today. | Match |
| B4 | `internal/hook/audit_receipt_guard.go` | Rejections are cleared only inside `checkAuditorStop` (`:185-190`, `:225-235`). `checkAuditReceiptSpawn` (`:429-464`) denies phase-entry spawns while any is outstanding. Marks are plain JSON files (`store.go:1255-1260`). | Match |
| B4 | `internal/cli/contract.go:76-78` | The reused agent-environment marker set is exactly `CLAUDECODE` and `CLAUDE_CODE_SESSION_ID`. No Codex session marker exists in `internal/`. | Added context; REQ-AEC-011 |
| B5 | Agent sources `internal/template/agentemit/source/{sync,plan}-auditor.md` | Lines 201/203 (sync) and 251/253 (plan) carry the wrong gate sentences. `make agents-emit` (`Makefile:34`, `:38-39`, `:47-49`) renders these two sources into the template sets only (`internal/template/templates/.claude/agents/moai/` and `internal/template/templates/.codex/agents/moai/`). The project targets are limited to `CarriedFour` (`internal/template/agentemit/sets.go:62-63`), whose comment (`:48-50`) says the other agents' project copies stay hand-managed; the four local definitions carry the same two wrong lines (read on the working tree at `f1cd8da975`, whose agent files equal `cc100c2909`: `.claude/agents/moai/sync-auditor.md:201,203`, `.claude/agents/moai/plan-auditor.md:251,253`, `.codex/agents/moai/sync-auditor.toml:190,192`, `.codex/agents/moai/plan-auditor.toml:248,250`). Only the plan-auditor source carries `blocking_count`; neither carries a `blocking_defect:` line. | Match; target corrected to the sources plus the four local definitions (R308, R313) |
| B5 | `internal/cli/mcp_codex_audit.go:2428-2448` | `applyGateUnmet`: line 2441 sets `fail` when the gate configuration cannot be read; line 2448 keeps an unmet gate inconclusive. | Match |
| B5 | Template mirrors | The rendered plan-auditor and its mirror differ at lines 64 and 150 (`.claude/skills/` against `~/.claude/skills/`); the sync-auditor pair is byte-identical. | Out of scope (R287) |

## §C Requirements (GEARS)

### B1 — Fix-confirm delta scope

- **REQ-AEC-001** (When) — **When** an `audit_multi` call names a `card_id`, carries no `delta_base`, and targets `baseBranch`, the audit engine shall propose as the delta base the `audited_sha` of the newest exported verdict file in that card's evidence directory, provided that file carries `verdict: FAIL` and exactly one `audited_sha` that is an ancestor of HEAD; **when** the card has no FAIL verdict file at all (a first review included), or the newest verdict is not a FAIL, or its `audited_sha` is absent, conflicting, or not an ancestor, the engine shall review the full scope; and the result shall name the verdict file and SHA it proposed, or state why it proposed none.
- **REQ-AEC-002** (When) — **When** an `audit_multi` call carries an explicit `delta_base` SHA, the engine shall propose that SHA instead of the verdict file's, and **when** the call also names a `card_id` that proposal shall pass REQ-AEC-004 exactly as a default proposal does; **when** the call carries `delta_base: "full"`, the engine shall review the full resolved base scope exactly as a call without `delta_base` does today; a call without a `card_id` keeps the t1708 behavior (§F).
- **REQ-AEC-003** (While) — **While** a delta base is in effect for an `audit_multi` call, the codex leg shall keep its adversarial `turn/start` turn and shall carry in that turn the `delta_base...HEAD` diff the server itself collected, not an instruction to compute it; the leg shall not switch to the native review target, and the result shall not describe the codex scope as prose-only.
- **REQ-AEC-004** (When) — **When** an `audit_multi` call that names a `card_id` has a proposed delta base, the engine shall review the delta scope only if both conditions hold, and shall otherwise review the full scope and state which condition failed:
  - **(i) Unbroken, reviewed chain from server-written state.** The card's second-review records, read from the card evidence home, form a chain that starts at a full-scope record. Each later record's delta base equals the previous record's head SHA. Every record, the starting record included, carries a pass or fail verdict for every backend the tree resolves as required; a required backend whose entry is absent from a record counts as not run. Each record cites an audit receipt in the receipt store with tool `audit_multi`, the same tree root, a head commit equal to the record's head SHA, an empty `gate_unmet`, and a codex verdict that is not inconclusive (the `CauseReceiptAuditNotRun` conditions). The proposed base equals the head SHA of the chain's last record. No part of the chain is read from a call argument or from a verdict file.
  - **(ii) BLOCKING-closure confirmation.** The closure set is the union of the `blocking_defect:` ids of every FAIL verdict file in the card's evidence directory whose `audited_sha` equals the head SHA of a chain record. A card with no such FAIL verdict file has no closure basis, and its proposal is reviewed at full scope. Each such file shall carry `blocking_count`, and the number of its `blocking_defect:` lines shall equal it; a missing count or a mismatch means full scope. The call carries a confirmation naming every id in the union. A closure verified at an earlier chain link is inherited: an id named by a FAIL file tied to an earlier record, and listed among the ids confirmed by the chain record whose delta base is that FAIL file's `audited_sha` (REQ-AEC-005), needs no closing commit in the current delta. Inheritance is decided per naming, not per id: when a FAIL file tied to a later chain record names an id again, the earlier inheritance of that id is void, and that id is inherited only if the chain record whose delta base is the latest naming file's `audited_sha` confirmed it, and otherwise needs a closing commit inside the current `delta_base..HEAD` that comes after the latest naming. Every other id, that is each closure new since the previous link, carries a closing commit inside `delta_base..HEAD`. The confirmation is passed to every leg as part of the review request.
  The audit artifact convention and its mirror shall define the `blocking_defect: <id>` line, one per BLOCKING defect, on plan-audit and sync-audit FAIL files alike. The guarantee "a delta round passes only on a verified chain" holds on this `card_id` path and its second-review records only.
- **REQ-AEC-005** (When) — **When** an `audit_multi` call names a `card_id` and reviewed a delta scope admitted under REQ-AEC-004, the engine shall:
  - file a second-review record for it, with the delta base as its scope base, a delta marker, the audit receipt id of the chain record it extends, the ids its confirmation named, and the SHA-256 of each FAIL verdict file whose ids formed the closure set;
  - persist its result to the session file whatever its overall verdict.
  A delta call that REQ-AEC-004 did not admit is a full-scope call: its record and persistence are as today. A delta call without a `card_id` persists only when its overall verdict is fail (t1873, unchanged). No other delta pass is written to the session file.

### B2 — An unrun required leg is classified, never failed

- **REQ-AEC-006** (When) — **When** an audit leg returns `inconclusive` because the review could not be performed, the leg's entry in `per_backend_verdicts` and the audit receipt shall carry an `error_class` from the closed set `quota`, `timeout`, `auth`, `unavailable`, `other`; the leg's `verdict` shall stay `inconclusive`, and a leg whose verdict is `pass` or `fail` shall carry no `error_class`.
- **REQ-AEC-007** (When) — **When** an audit ends with no backend reporting a defect and at least one required leg inconclusive, the exported verdict file shall carry the label `INCONCLUSIVE`. Every consumer shall read `INCONCLUSIVE` as not-performed: never a PASS claim, and never a FAIL. The receipt lines, the shared admission rule, the Stop-hook last line, the factory card reader, and both round counters shall then behave as follows:
  - **Receipt lines.** The file carries `convergence_overall: inconclusive` and one `required_backend: <backend> inconclusive <error_class>` line per unrun leg. It carries no `required_backend_fail:` line for an unrun leg; that line names only a backend that reported fail.
  - **Admission.** The shared admission rule refuses an `INCONCLUSIVE` file exactly as it refuses FAIL, and the refusal reason names each unrun backend and its error class.
  - **Stop-hook last line.** Its grammar accepts `AUDIT-VERDICT: INCONCLUSIVE spec=… receipts=…`, a parsed line that neither records `CauseVerdictLineMissing` nor clears a rejection. Stage entry is blocked exactly as after FAIL, and the reason carries the error class.
  - **Factory card reader.** It accepts `verdict: INCONCLUSIVE` and keeps the card in audit as it does for FAIL, with the class in its reason.
  - **Round counters.** Both plan-audit round counters (`CountAuditRounds` and `CountPlanAuditRounds`) skip `INCONCLUSIVE` iteration files and keep counting FAIL and PASS-family files. An `INCONCLUSIVE` file with the highest iteration number is still selected as the latest verdict, so an older verdict never becomes the decision.

  `AdmitLabel` stays unchanged, because it already refuses a label it does not know. §G records how this amends SPEC-AUDIT-CEILING-001 REQ-ACE-008 and SPEC-AUDIT-CEILING-002 REQ-ACR-005.
- **REQ-AEC-008** (When) — **When** an `audit_multi` call names legs to re-run, the engine shall re-run only those legs and shall carry every other leg's verdict from the result the server itself persisted for the same session. Re-running applies only when that persisted result records the same head commit, the same tree digest, and the same scope (target, resolved base, and delta base), fields the persisted result shall carry for this purpose. **When** no such persisted result exists or any of those differ, the engine shall run every leg. A carried verdict shall never be read from a caller-supplied argument.

### B3 — Receipt binding across evidence-only commits

- **REQ-AEC-009** (While) — **While** checking a receipt's tree-state binding for a citation, the receipt store shall treat the receipt as fresh when all of the following hold:
  - the receipt's head commit is an ancestor of the current HEAD;
  - every change between the two commits is one of:
    - (a) a change to a file under `.moai/reports/<id>/` whose name belongs to the audit-artifact family (`plan-audit*.md`, `sync-audit*.md`, `verdict*.md`, `card-review.md`, `second-review.jsonl`);
    - (b) a change to the `progress.md` of the SPEC that the citing verdict line names (`spec=`), where the `§E.2` section is computed on both the old and the new version, is byte-identical between them, and its heading line is unchanged and in the same position. Changes under `§E.1`, `§E.3`, and `§E.4` are allowed, including a full rewrite of `§E.3`.
  - the uncommitted state outside those changes equals the uncommitted state the receipt judged.
- **REQ-AEC-010** (Ubiquitous) — The receipt store shall treat a receipt as stale whenever any condition of REQ-AEC-009 does not hold. That includes:
  - a change to any other file under `.moai/reports/`;
  - a `progress.md` change touching even one line under `§E.2`;
  - a change, deletion, rename, or move of the `§E.2` heading line;
  - a change to the `progress.md` of a SPEC other than the cited one, or any `progress.md` change when the citation names a card id rather than a SPEC ID;
  - a path that merely begins with the allowed prefix text;
  - a rename into or out of the allowed set;
  - a receipt written before this change that carries no separate uncommitted-state record.
  A receipt whose head equals the current HEAD shall be judged exactly as today.

### B4 — One operator path to clear a rejection

- **REQ-AEC-011** (Ubiquitous) — The audit receipt subsystem shall provide exactly one input path that clears a receipt-kind rejection mark on operator approval: one CLI verb on the human path, reusing the contract signing path's interactive-terminal check, its typed confirmation, and its agent-environment marker set, which is exactly `CLAUDECODE` and `CLAUDE_CODE_SESSION_ID` (no Codex session marker exists in the codebase, so a Codex session is not refused by the marker check). It shall not provide an MCP tool, a hook path, a configuration key, or a file drop that clears a rejection on approval, and no process shall read an approval record as authority to clear a mark.
- **REQ-AEC-012** (When) — **When** any of the following holds, the verb shall refuse, write nothing, and remove nothing:
  - either marker of REQ-AEC-011 is set;
  - standard input is not an interactive terminal;
  - the named rejection mark does not exist in the current tree, or was recorded for another tree;
  - the evidence file is outside `.moai/reports/`, or its content does not name the mark's SPEC ID;
  - the typed confirmation does not repeat the mark's agent type and SPEC ID.
- **REQ-AEC-013** (When) — **When** every check of REQ-AEC-012 passes, the verb shall append one approval record to an append-only log inside the existing audit-receipt store — naming the operator (the typed name and the git identity), the mark's fields as read, the evidence file path and its SHA-256, the tree's HEAD, and the UTC time — and shall then remove exactly the one named mark file and no other.

### B5 — Auditor definitions state the real behavior

- **REQ-AEC-014** (Ubiquitous) — The plan-auditor and sync-auditor agent sources under `internal/template/agentemit/source/`, every copy `make agents-emit` renders from them (the template Claude agent files and the template Codex TOML files), and the four local definitions `.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md`, `.codex/agents/moai/plan-auditor.toml`, and `.codex/agents/moai/sync-auditor.toml`, which the emitter does not render and which are updated by hand to match the corrected source text (leader rulings R308, R313), shall:
  - state that when no backend reported a defect and an explicitly required leg did not run, the result is `inconclusive` with a non-empty `gate_unmet`, yields no PASS, and is exported as `INCONCLUSIVE`;
  - state that when a required backend reported fail, the result stays fail even if another required leg did not run;
  - state that on the single-backend `codex_audit` path an unmet gate keeps the verdict `inconclusive`, and `fail` is returned only when the gate configuration cannot be read;
  - not state that an unmet gate alone returns or makes the verdict `fail`;
  - list `INCONCLUSIVE` in the `verdict:` and `AUDIT-VERDICT:` machine-line templates;
  - oblige the auditor, on every FAIL verdict file (plan-audit and sync-audit alike), to write `blocking_count` and one `blocking_defect: <id>` line per BLOCKING defect.

### A1 — Repair-scope rule

- **REQ-AEC-015** (Ubiquitous) — The audit artifact convention and its template mirror shall carry one rule paragraph stating that a repair of an audit FAIL may only remove the defective path, tighten an existing check, or add tests; that a repair needing a new verb, flag, input, store, or rule is delivered as a SPEC amendment with its own acceptance criterion on a separate card; and that work already specified by a SPEC amendment with its own acceptance criterion is outside the limit.

### Cross-cutting

- **REQ-AEC-016** (Ubiquitous) — The repair shall not delete a check, weaken a refusal, replace the installed binary, or add a concurrently writing session or agent to a working tree; every edited file with a template mirror or a rendered copy shall change in the same commit as it, followed by `make build` (which runs `agents-emit-check`) exiting 0.

## §D The A1 / B4 tension, resolved

A1 forbids a repair from adding a new verb, flag, store, or rule without a SPEC amendment and its own acceptance criterion. This SPEC adds:
- a CLI verb (B4);
- two `audit_multi` inputs (the leg re-run list and the BLOCKING-closure confirmation);
- a verdict label and two verdict-file line shapes;
- receipt, persisted-result, and second-review record fields.

These are allowed because this SPEC is the amendment A1 asks for: each addition has a requirement here and an acceptance criterion in `acceptance.md`. A1's own text (REQ-AEC-015) carries the clause that makes this explicit, so A1 applies only to repairs made after a FAIL without such an amendment. It does not reach back over any item of this SPEC.

Leader ruling R287 (Q7) settles the second boundary. "No new writer" means no concurrently writing session or agent added to a tree. The B4 verb is an operator invocation that appends one record and removes one mark, so it is not a new writer in that sense (REQ-AEC-016).

## §E Design decisions recorded with this plan

### B1 — what a correct-looking delta chain can still miss (residual risk)

A delta round passes only when the chain is unbroken and reviewed, and the BLOCKING-closure confirmation is complete. Even so:

- **Defects that predate the full review.** A defect that predates the chain's full-scope record and that review did not find stays outside every later delta's scope. No delta round can see it.
- **Interactions.** A defect created by the interaction between a delta change and code outside the delta is visible only if a reviewer reads beyond the diff it was given.
- **The confirmation is self-asserted.** The lane writes it. The engine checks that it is complete and that each closing commit lies in range; whether the commit actually closes the defect is the reviewers' judgment. The confirmation is passed to them so they can check it.
- **`blocking_defect:` ids come from the exported verdict files, which the auditor writes.** An auditor that under-reports both the ids and `blocking_count` passes the count check. Leader ruling R288 records this as residual risk. The SHA-256 of each relied-on verdict file is written into the delta record (REQ-AEC-005), so a later edit to that file is detectable. If a plan audit raises server-side id storage as BLOCKING, the leader decides on splitting B4 out.
- **"Server-written" means written by the audit code path, not tamper-proof.** Second-review records and receipts are files. A process with write access to the tree could forge a matching record and receipt pair. Leader ruling R288 fixes the threat model as a cooperating process. The mitigation is the server-written record-and-receipt pair, the `audited_sha` binding, and the operator-only B4 path. A server-held secret or signature is a residual risk and a follow-up candidate only; no card is issued for it.
- **Calls outside the `card_id` path.** A delta from `audit_multi` without a `card_id`, or from `codex_audit` called directly, has no chain gate, and its receipt carries no scope field, so a hook-cited PASS receipt from such a delta is not distinguished from a full one. This is the t1708 behavior (§F).

### B1 — the session file and the Stop gate

The Stop gate reads the session file and checks no scope (t1873). A chain-admitted delta has passed the same gate that keeps an unverified delta out of coverage, so its result is persisted whatever its verdict (REQ-AEC-005): a chain-admitted delta PASS opens the gate after the earlier full FAIL, which is the point of B1. A delta without a `card_id` keeps the t1873 rule, so an unverified delta PASS still never opens it.

### B2 — `INCONCLUSIVE` and the round counter

Leader ruling R287 (Q3) excludes `INCONCLUSIVE` rounds from the plan-audit ceiling count, and leader ruling R288 confirms that "the Q8 default stands" means the S1/M2/L3 numbers are untouched while Q3 decides what is counted. Both counters change. An `INCONCLUSIVE` file is still the latest verdict when its iteration number is highest; admission refuses it, so the card stays blocked rather than falling back to an older verdict.

### B3 — commit-range allowlist (chosen) versus a code-tree digest (rejected)

Chosen: keep the existing binding and its exact-match path, and relax freshness only when HEAD moved: ancestry, plus changes confined to the allowed set, plus an unchanged uncommitted state outside it. This needs one new receipt field (the uncommitted state without the head) and the cited SPEC ID passed into the check, which receives none today (`CheckCitedReceiptsSince`, `store.go:1514`). It relaxes the freshness rule of SPEC-MCP-NEUTRALITY-001 REQ-MN-020 ("refuse … when the stored tree state differs from the current one") for exactly this allowed set (§G).

How the wrongly-accepted cases are closed:

- **Edits to non-artifact report files** — closed by exclusion at path level. Only the audit-artifact filename family under `.moai/reports/<id>/` is allowed.
- **A rewrite of `progress.md` run evidence after a PASS** — closed by a negative case. The `§E.2` section is computed on both versions, so deleting, renaming, or moving its heading cannot slide the evidence body under `§E.1`. Leader ruling R288 sets the boundary: rewriting `§E.3` after a PASS is the normal flow and keeps the receipt.
- **Another SPEC's `progress.md`** — closed by restricting the allowance to the SPEC the citation names.
- **Remaining residual** — an evidence file deliberately named into the audit-artifact family (for example `verdict-notes.md`) would be accepted.

Rejected: a code-tree digest that ignores the head. It wrongly accepts everything the allowlist accepts, plus history that reaches the same content without descending from the judged commit, and it would change the digest format at an `@MX:ANCHOR` seam.

### B4 — CLI verb on the human path (R287 Q6), no MCP path

The verb reuses the contract signing path's checks: the two agent-environment markers, `term.IsTerminal` on standard input (which refuses `</dev/null`), the typed confirmation, and the git identity. An MCP tool is excluded because every MCP call comes from an agent session.

- **What the checks prove, and what they do not.** They prove only that the process has neither marker set and has a terminal on standard input. Under the cooperating-process threat model this is a guard against operator mistakes. It does not stop:
  - an agent's own Bash call that unsets the markers and allocates a pseudo-terminal (for example `env -u CLAUDECODE -u CLAUDE_CODE_SESSION_ID script -q /dev/null moai …`);
  - a Codex session, which sets neither marker and is not detected (whether it provides a terminal was not measured);
  - direct deletion of the rejection mark file under `.moai/state/audit-receipts/rejections/`, which no hook guards.
  It is not a security boundary.
- **PreToolUse deny: recorded as a Gap, not built here.** A deny of the verb in agent Bash calls would add one layer. It needs a new matcher in the branch-guard hook family, an opt-in configuration key, and its own tests, and the requirement budget is at the Tier M ceiling.

### B5 — local agent definitions outside the emitter (leader rulings R308, R313)

The emitter writes project copies only for `CarriedFour` (`internal/template/agentemit/sets.go:62-63`), so `agents-emit-check` neither regenerates nor compares the local plan-auditor and sync-auditor, in either the `.claude` or the `.codex` project set. Leader ruling R308 decides a document-only repair: M6 updates the local definitions by hand to match the corrected source text, and AC-AEC-005 checks them directly. Leader ruling R313 sets the local set to four files: `.claude/agents/moai/{plan,sync}-auditor.md` and `.codex/agents/moai/{plan,sync}-auditor.toml`. Each TOML copy carries the same two wrong lines as its `.claude` counterpart and no `blocking_defect:` line (EV-20). Lifting the `CarriedFour` restriction is not part of this SPEC; it is recorded as input for card t2018 (`progress.md`).

- **Residual: drift after the hand edit.** No check keeps the hand-updated local definitions equal to their sources after M6. AC-AEC-005 observes them once, at M6.

### A1 — home of the repair-scope paragraph

Chosen: `.moai/docs/audit-artifact-convention.md` and its byte-identical template mirror. Both auditor definitions already cross-reference it, and M4 edits the same file. Rejected: `.claude/rules/moai/workflow/spec-workflow.md`, which loads into every session touching `.moai/specs/**` and whose subject is the phase lifecycle.

## §F Out of Scope

### Out of Scope — release and binary
- Replacing or reinstalling the installed `moai` binary as part of the repair. (Per leader ruling R290 this card is part of the rc.39 integration candidate; its landing follows that candidate's batch.)

### Out of Scope — audit policy beyond the six items
- A default delta scope or a chain check for `codex_audit`, `glm_audit`, or `claude_audit` called directly, and for an `audit_multi` call without a `card_id`; those keep the t1708 behavior, and their delta receipts carry no scope field.
- A chain check inside the closure reader `SelectSecondReview`; the chain is enforced where records are written.
- Changing the convergence policy, `overall_verdict` values, `disagreement_flag`, or the per-backend `verdict` enum.
- Clearing served-model (`KindServed`) rejections; only receipt-kind rejections are covered.
- A PreToolUse deny of the B4 verb, and any guard on direct deletion of a rejection mark file (recorded as Gaps in §E).

### Out of Scope — pre-existing divergence
- The pre-existing differences between the rendered plan-auditor and its template mirror at lines 64 and 150 (leader ruling R287).
- Adding the plan-auditor and sync-auditor to the emitter's project targets (`CarriedFour`); this is input for card t2018 (leader ruling R308). The four local definitions are in scope and are fixed by hand (REQ-AEC-014, leader ruling R313).

## §G Reconciliation with completed SPECs

This SPEC does not reopen the status of any completed SPEC. It amends the following rules for the verdict files it produces, and the run phase changes the code and convention that implement them:

| Completed rule | What it says today | What this SPEC sets |
|---|---|---|
| SPEC-AUDIT-CEILING-001 REQ-ACE-008 (projection) | An own `INCONCLUSIVE` label projects to `inconclusive` on the backend line and `fail` on `convergence_overall` | An `INCONCLUSIVE` file carries `convergence_overall: inconclusive`. The parser accepts that value, and admission refuses it with a reason naming each `required_backend` line recorded `inconclusive` and its class. FAIL still projects to `fail` |
| SPEC-AUDIT-CEILING-001 receipt line format | `required_backend: <backend> <pass\|fail\|inconclusive>`, split once at the first space (`verdict.go:168`) | A third token is allowed only after `inconclusive`, and only from the closed class set. A class after `pass` or `fail`, or an unknown class, is a malformed receipt line and refuses |
| SPEC-AUDIT-CEILING-002 REQ-ACR-005 (`required_backend_fail:`) | Written for "a required backend the audit could not admit" | Written only for a backend that reported fail. An unrun leg is carried by its `required_backend … inconclusive <class>` line |
| Admission order (`verdict.go` `admitBase`, `admitReceipt`) | Duplicates → `required_backend_fail` → `convergence_overall: fail` → label → fields → receipt lines | Duplicates → `required_backend_fail` → `convergence_overall: fail` → `convergence_overall: inconclusive` (the reason names each unrun backend and class) → label → fields → receipt lines. With no `required_backend_fail` line and no overall `fail`, the class-naming reason is the one reached |
| SPEC-MCP-NEUTRALITY-001 REQ-MN-020 (freshness) | A receipt is refused when the stored tree state differs from the current one | Relaxed only for the allowed set of REQ-AEC-009; every other difference still refuses |
