---
id: SPEC-AUDIT-ENGINE-CONVERGENCE-001
title: "Audit engine convergence repairs — fix-confirm delta by default with a verified delta chain, unrun legs classified instead of failed, receipts that survive evidence-only commits, one operator path to clear a rejection, honest auditor prompts, and a repair-scope rule"
version: "0.3.0"
status: draft
created: 2026-11-03
updated: 2026-11-03
author: manager-spec (card t2017)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/auditreceipt, internal/auditverdict, internal/runtime, internal/closure, .claude/agents/moai, .moai/docs"
lifecycle: spec-anchored
tags: "audit, convergence, delta-scope, delta-chain, second-review, receipt-binding, rejection, operator-approval, inconclusive, repair-scope, card-t2017"
tier: M
card: t2017
related_specs: [SPEC-MCP-NEUTRALITY-001, SPEC-AUDIT-MODEL-CONVERGE-001, SPEC-AUDIT-CEILING-REPAIR-001, SPEC-AUDIT-GATE-INTEGRITY-001, SPEC-AUDIT-BUILD-IDENTITY-001, SPEC-WF-AUDIT-GATE-001, SPEC-RECEIPT-REUSE-001, SPEC-WORKTREE-STATE-ROOT-001, SPEC-AUTONOMY-CLOSURE-001]
---

# SPEC-AUDIT-ENGINE-CONVERGENCE-001 — Audit engine convergence repairs

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-11-03 | manager-spec (card t2017) | Initial plan-phase draft. Six repair items (B1-B5, A1). Every code citation was re-read on tree `cb30e05b3e5f64774b8dabd87e80c24332f4551d`; mismatches are recorded in §B. |
| 0.2.0 | 2026-11-03 | manager-spec (card t2017) | Leader ruling R287 applied. Q1: the delta default stands, a delta round can pass only on a verified chain plus a BLOCKING-closure confirmation, and a gated delta call files a second-review record (REQ-AEC-004, -005). Q3: `INCONCLUSIVE` label with its consumer rules (REQ-AEC-007). Q6/Q7: CLI verb on the human path, no MCP path; the hook deny is recorded as a Gap. B3: the reports allowlist is narrowed to the audit-artifact filename family, and `progress.md` §E.2/§E.3 edits are pinned as negative cases (REQ-AEC-009, -010). REQ 14 → 16, AC sub-criteria 15 → 16; both are now at the Tier M ceiling. |
| 0.3.0 | 2026-11-03 | manager-spec (card t2017) | Leader ruling R288 applied. The B3 negative case narrows to `progress.md` §E.2: a rewrite of §E.3 after a PASS keeps the receipt (REQ-AEC-009, -010; pinned inside AC-AEC-003a/003b, counts unchanged at 16/16). The Q3 reading of the round counter is confirmed. The cooperating-process threat model and the auditor-written BLOCKING ids are recorded as residual risk (§E). B4 stays in this SPEC. |

## §A Background

Ten audit rounds on card t2012 (PR #1965) stalled honest work. The causes were in the audit engine, not in the work under audit:

- A fix-confirm re-audit re-reviewed the whole card instead of the repair, so findings that predated the repair came back every round. The engine already has an explicit `delta_base` input (card t1708), but nothing fills it in, and on `audit_multi` the codex leg is told the delta scope in prompt prose only.
- A required leg that could not run (quota, timeout, authentication) ended the round as if a defect had been found. The exported verdict vocabulary has no non-FAIL label for "the review did not happen", and both auditor prompts tell the auditor that an unmet gate yields `verdict: fail`.
- An audit receipt went stale the moment the lane committed its own audit report or progress record, because the freshness check compares the head commit exactly.
- A rejection mark can be cleared only by the same auditor role producing a proven PASS. There is no operator path, so a rejection caused by an engine fault blocked every phase-entry spawn until the engine itself recovered.

The operator decided on 2026-11-03 to repair six defects. The leader relayed the follow-up product rulings as leader ruling R287 (`decision-index.md`). This SPEC does not land in the rc.39 candidate (`945cdb60fc`); it targets the next release candidate.

## §B Cited-location verification (read on `cb30e05b3e`)

The source record (leader record R283) cited code locations. Each was read in this tree before any requirement was written. The leader accepted the four corrections below in R287.

| Item | Claimed location | What the code shows | Result |
|---|---|---|---|
| B1 | `internal/cli/audit_delta_scope.go:27` | The `delta_base` tool-schema description: explicit optional input; codex adversarial and the `audit_multi` codex leg are "told in the prompt only". No default exists. | Match |
| B1 | `internal/cli/audit_delta_scope.go:144-157` | Lines 144-147 are the tail of `deltaScopeNote`; 149-155 is `codexDeltaScopeClause`. The prose-only text itself is at line 106. | Match (range straddles two functions) |
| B1 | `internal/cli/mcp_convergence.go:1059` (ruling R287) | `if cfg.CardID != "" && cfg.DeltaBase == ""` around `appendSecondReviewRecord`. Today a call with a `card_id` and no delta files one second-review record in the card's evidence home, and a delta call files none ("a delta call is not full-scope review coverage"). | Read; REQ-AEC-005 changes it |
| B1 | (not cited) | The `audit_multi` codex leg is `performCodexAudit` (`mcp_convergence.go:631`), which always sends a `turn/start` adversarial turn. `codex_audit` native mode already binds a delta mechanically through `params["target"]` (`mcp_codex_audit.go:2323-2330`). | Added context |
| B1 | (not cited) | The closure reader `SelectSecondReview` (`internal/closure/readiness.go:109-200`) keeps a record whose target is `baseBranch`, whose scope head equals its head, and whose changed-file count is at least 1 (`:151-155`), then picks the newest by `recorded_at`. It has no notion of a delta chain, so it would accept a delta-scope record as coverage. | Added context; the chain check therefore lives on the writing side (REQ-AEC-004) |
| B2 | `internal/cli/mcp_convergence.go:1123-1155` | `enforceRequiredGateUnmet` already keeps an unmet required leg at overall `inconclusive` (REQ-MN-021) and never turns it into `fail`. | Mismatch: this code does not label the leg FAIL |
| B2 | `internal/auditverdict/verdict.go:361` | `if v != "pass"` in `admitReceipt`: a required backend recorded `fail` or `inconclusive` refuses admission with the reason "required backend X recorded inconclusive". It refuses; it does not relabel. | Partial: a refusal reason, no error class |
| B2 | `internal/auditreceipt/store.go:1563` | The `CauseReceiptAuditNotRun` arm, already distinct from `CauseReceiptVerdictFail`. | Partial: distinct cause exists, no error class |
| B2 | (not cited) | The FAIL label actually arises from the exported verdict vocabulary `verdict: <PASS\|PASS-WITH-DEBT\|FAIL>` (`.moai/docs/audit-artifact-convention.md:87`) and from both auditor prompts (B5). Claude legs already carry backend error codes (`internal/cli/mcp_claude.go:32-44`, in `Provenance.ErrorCode`); codex and GLM legs carry none. The plan-audit round counter counts every iteration file "even when the verdict label … [is] identical" (`internal/runtime/audit_counter.go:10`), so it does not read the label today. | Added context |
| B3 | `internal/auditreceipt/store.go:1605` | `head, digest, sErr := TreeState(r.TreeRoot)`; line 1608 compares `r.HeadCommit != head \|\| r.TreeDigest != digest`. `TreeState` (`store.go:221`) folds the head into the digest, so any commit moves both. Untracked files under `.moai/reports/` are already excluded (`isRuntimeManagedStatePath`, `store.go:385-392`); a tracked report or progress commit is not. | Match |
| B4 | `internal/hook/audit_receipt_guard.go` | Rejections are cleared only inside `checkAuditorStop`: a parsed FAIL clears `CauseVerdictLineMissing` (`:185-190`), and a proven PASS clears the role's receipt-kind rejections in that tree (`:225-235`). `checkAuditReceiptSpawn` (`:429-464`) denies phase-entry spawns while any is outstanding. | Match |
| B5 | `.claude/agents/moai/sync-auditor.md:203`, `.claude/agents/moai/plan-auditor.md:253` | "…returns `verdict: fail` with a non-empty `gate_unmet`…". | Match. Also wrong: sync-auditor line 201 and plan-auditor line 251 ("the overall verdict is `fail`"). |
| B5 | `internal/cli/mcp_codex_audit.go:2448` | Inside `applyGateUnmet` (`:2428`): the comment "could-not-run stays not-performed — verdict unchanged". Exception: line 2441 sets `fail` when the gate configuration cannot be read. | Match, with the exception the correction keeps |
| B5 | Template mirrors | The sync-auditor mirror is byte-identical; the plan-auditor mirror already differs at line 64 (out of scope, R287). | Added context |

## §C Requirements (GEARS)

### B1 — Fix-confirm delta scope

- **REQ-AEC-001** (When) — **When** an `audit_multi` call names a `card_id`, carries no `delta_base`, and targets `baseBranch`, the audit engine shall propose as the delta base the `audited_sha` of the newest exported verdict file in that card's evidence directory, provided that file carries `verdict: FAIL` and exactly one `audited_sha` that is an ancestor of HEAD; **when** the card has no FAIL verdict (including a card's first review), or the newest verdict is not a FAIL, or its `audited_sha` is absent, conflicting, or not an ancestor, the engine shall review the full scope; and the result shall name the verdict file and SHA it proposed, or state why it proposed none.
- **REQ-AEC-002** (When) — **When** an `audit_multi` call carries an explicit `delta_base` SHA, the engine shall propose that SHA instead of the verdict file's; **when** the call carries `delta_base: "full"`, the engine shall review the full resolved base scope exactly as a call without `delta_base` does today.
- **REQ-AEC-003** (While) — **While** a delta base is in effect for an `audit_multi` call, the codex leg shall review exactly `delta_base...HEAD` through the review request's own target or diff input rather than through prompt prose, and the result shall not describe the codex scope as prose-only.
- **REQ-AEC-004** (When) — **When** an `audit_multi` call that names a `card_id` has a proposed delta base, the engine shall review the delta scope only if both of the following hold, and shall otherwise review the full scope and state which condition failed:
  - **(i) Unbroken chain from server-written state.** The card's second-review records, read from the card evidence home, form a chain that starts at a full-scope record. Each later record's delta base equals the previous record's head SHA. Each record cites an audit receipt that exists in the receipt store with tool `audit_multi`, the same tree root, and a head commit equal to the record's head SHA. The proposed base equals the head SHA of the chain's last record. No part of the chain is read from a call argument or from a verdict file.
  - **(ii) BLOCKING-closure confirmation.** The call carries a confirmation naming every `blocking_defect:` id of the newest FAIL verdict file, each with a closing commit inside `delta_base..HEAD`. The number of ids equals that file's `blocking_count`. The confirmation is passed to every leg as part of the review request.
  The audit artifact convention and its mirror shall define the `blocking_defect: <id>` line, one per BLOCKING defect.
- **REQ-AEC-005** (When) — **When** an `audit_multi` call names a `card_id` and reviewed a delta scope admitted under REQ-AEC-004, the engine shall file a second-review record for it. The record's scope base shall be the delta base, it shall be marked as a delta record, and it shall name the audit receipt of the chain record it extends. A full-scope call shall file its record exactly as today, and a delta call that REQ-AEC-004 sent to full scope is a full-scope call.

### B2 — An unrun required leg is classified, never failed

- **REQ-AEC-006** (When) — **When** an audit leg returns `inconclusive` because the review could not be performed, the leg's entry in `per_backend_verdicts` and the audit receipt shall carry an `error_class` from the closed set `quota`, `timeout`, `auth`, `unavailable`, `other`; the leg's `verdict` shall stay `inconclusive`, and a leg whose verdict is `pass` or `fail` shall carry no `error_class`.
- **REQ-AEC-007** (When) — **When** an audit ends with no backend reporting a defect and at least one required leg inconclusive, the exported verdict file shall carry the label `INCONCLUSIVE` and one `required_backend: <backend> inconclusive <error_class>` line per unrun leg, and shall not carry `FAIL`. Every consumer shall read `INCONCLUSIVE` as not-performed, neither PASS nor FAIL:
  - the shared admission rule shall refuse it exactly as it refuses FAIL, and the refusal reason shall name each unrun backend and its error class;
  - the plan-audit round counter shall not count an `INCONCLUSIVE` iteration file and shall keep counting FAIL and PASS-family files as today.
- **REQ-AEC-008** (When) — **When** an `audit_multi` call names legs to re-run, the engine shall re-run only those legs and shall carry every other leg's verdict from the convergence result the server itself persisted for the same session, the same tree state (head commit and tree digest), and the same scope; **when** no such persisted result exists or any of those three differ, the engine shall run every leg; and a carried verdict shall never be read from a caller-supplied argument.

### B3 — Receipt binding across evidence-only commits

- **REQ-AEC-009** (While) — **While** checking a receipt's tree-state binding, the receipt store shall treat the receipt as fresh when all of the following hold:
  - the receipt's head commit is an ancestor of the current HEAD;
  - every change between the two commits is either (a) to a file under `.moai/reports/<id>/` whose name belongs to the audit-artifact family (`plan-audit*.md`, `sync-audit*.md`, `verdict*.md`, `card-review.md`, `second-review.jsonl`), or (b) to a `progress.md` directly inside a `.moai/specs/<SPEC-ID>/` directory, with no changed line under its `§E.2` heading (changes under `§E.1`, `§E.3`, and `§E.4`, including a full rewrite of `§E.3`, are allowed);
  - the uncommitted state outside those changes equals the uncommitted state the receipt judged.
- **REQ-AEC-010** (Ubiquitous) — The receipt store shall treat a receipt as stale whenever any condition of REQ-AEC-009 does not hold, including a change to any other file under `.moai/reports/`, a `progress.md` change touching even one line under `§E.2`, a path that merely begins with the allowed prefix text, a rename into or out of the allowed set, and a receipt written before this change that carries no separate uncommitted-state record; a receipt whose head equals the current HEAD shall be judged exactly as today.

### B4 — One operator path to clear a rejection

- **REQ-AEC-011** (Ubiquitous) — The audit receipt subsystem shall provide exactly one input path that clears a receipt-kind rejection mark on operator approval: one CLI verb on the human path, reusing the agent-environment marker check, the interactive-terminal check, and the typed confirmation of the contract signing path. It shall not provide an MCP tool, a hook path, a configuration key, or a file drop that clears a rejection on approval, and no process shall read an approval record as authority to clear a mark.
- **REQ-AEC-012** (When) — **When** the verb runs with any agent-environment marker set, with standard input that is not an interactive terminal, naming a rejection mark that does not exist in the current tree, naming a mark recorded for another tree, naming an evidence file outside `.moai/reports/` or one whose content does not name the mark's SPEC ID, or receiving a typed confirmation that does not repeat the mark's agent type and SPEC ID, the verb shall refuse, write nothing, and remove nothing.
- **REQ-AEC-013** (When) — **When** every check of REQ-AEC-012 passes, the verb shall append one approval record to an append-only log inside the existing audit-receipt store — naming the operator (the typed name and the git identity), the mark's fields as read, the evidence file path and its SHA-256, the tree's HEAD, and the UTC time — and shall then remove exactly the one named mark file and no other.

### B5 — Auditor prompts state the real behavior

- **REQ-AEC-014** (Ubiquitous) — The plan-auditor and sync-auditor definitions and their template mirrors shall state that an explicitly required gate left unmet keeps the verdict `inconclusive` with a non-empty `gate_unmet`, yields no PASS, and is exported as `INCONCLUSIVE`; that `fail` is returned only when the gate configuration cannot be read; and shall not state that an unmet gate returns or makes the verdict `fail`.

### A1 — Repair-scope rule

- **REQ-AEC-015** (Ubiquitous) — The audit artifact convention and its template mirror shall carry one rule paragraph stating that a repair of an audit FAIL may only remove the defective path, tighten an existing check, or add tests; that a repair needing a new verb, flag, input, store, or rule is delivered as a SPEC amendment with its own acceptance criterion on a separate card; and that work already specified by a SPEC amendment with its own acceptance criterion is outside the limit.

### Cross-cutting

- **REQ-AEC-016** (Ubiquitous) — The repair shall not delete a check, weaken a refusal, replace the installed binary, or add a concurrently writing session or agent to a working tree; every edited file with a template mirror shall change in the same commit as its mirror, followed by `make build` exiting 0.

## §D The A1 / B4 tension, resolved

A1 forbids a repair from adding a new verb, flag, store, or rule without a SPEC amendment and its own acceptance criterion. This SPEC adds:
- a CLI verb (B4);
- two `audit_multi` inputs (the leg re-run list and the BLOCKING-closure confirmation);
- a verdict label and a verdict-file line;
- receipt and second-review record fields.

These are allowed because this SPEC is the amendment A1 asks for: each addition has a requirement here and an acceptance criterion in `acceptance.md`. A1's own text (REQ-AEC-015) carries the clause that makes this explicit, so A1 applies only to repairs made after a FAIL without such an amendment. It does not reach back over any item of this SPEC.

Leader ruling R287 (Q7) settles the second boundary. "No new writer" means no concurrently writing session or agent added to a tree. The B4 verb is an operator invocation that appends one record and removes one mark, so it is not a new writer in that sense (REQ-AEC-016).

## §E Design decisions recorded with this plan

### B1 — what a correct-looking delta chain can still miss (residual risk)

A delta round passes only when the chain is unbroken and the BLOCKING-closure confirmation is complete. Even so:

- **Defects that predate the full review.** A defect that predates the chain's full-scope record and that review did not find stays outside every later delta's scope. No delta round can see it.
- **Interactions.** A defect created by the interaction between a delta change and code outside the delta is visible only if a reviewer reads beyond the diff it was given.
- **The confirmation is self-asserted.** The lane writes it. The engine checks that it is complete and that each closing commit lies in range; whether the commit actually closes the defect is the reviewers' judgment. The confirmation is passed to them so they can check it.
- **`blocking_defect:` ids come from the exported verdict file, which the auditor writes.** An auditor that omits an id from those lines also omits it from the completeness check. The count must equal `blocking_count`, which catches a mismatch but not an auditor that under-reports both.
- **"Server-written" means written by the audit code path, not tamper-proof.** Second-review records and receipts are files. A process with write access to the tree could forge a matching record and receipt pair. The receipt cross-check raises the cost of forgery from one file to two coordinated files; it does not make it impossible. Leader ruling R288 fixes the threat model as a cooperating process. The mitigation is the server-written record-and-receipt pair, the `audited_sha` binding, and the operator-only B4 path. A server-held secret or signature is a residual risk and a follow-up candidate only; no card is issued for it.
- **BLOCKING ids are not server-persisted.** The engine does not store the BLOCKING ids of a FAIL itself; it reads them from the auditor-written verdict file. Leader ruling R288 records this as residual risk. If a plan audit raises server-side storage as BLOCKING, the leader decides on splitting B4 out.

### B2 — `INCONCLUSIVE` and the round counter

Leader ruling R287 (Q3) excludes `INCONCLUSIVE` rounds from the plan-audit ceiling count. The counter does not read the label today (`audit_counter.go:10`), so REQ-AEC-007 changes it to skip `INCONCLUSIVE` files while FAIL files keep counting. The same ruling says "the Q8 default stands", and that default reads "the ceiling ladder is unchanged". The two statements conflict for this one rule, so `decision-index.md` Q8 records the conflict and this SPEC follows the explicit Q3 consumer rule.

### B3 — commit-range allowlist (chosen) versus a code-tree digest (rejected), and the two cases R287 asked to close

Chosen: keep the existing binding and its exact-match path, and relax freshness only when HEAD moved: ancestry, plus changes confined to the allowed set, plus an unchanged uncommitted state outside it. This needs one new receipt field, the uncommitted-state record without the head.

The two wrongly-accepted cases from round 1 are closed as follows, choosing whichever keeps the allowlist simplest:

- **Edits to tracked report files that an acceptance criterion cites.** Closed by an exclusion at path level. Only the audit-artifact filename family under `.moai/reports/<id>/` is allowed, and every other file there (evidence ledgers, baselines, repro logs) is excluded. Path-level matching is the simplest rule and needs no knowledge of which files an AC cites.
- **A rewrite of `progress.md` run evidence after a PASS.** Closed by a negative case. `progress.md` cannot be excluded at path level without losing the commit B3 exists to tolerate, the lane's own audit-signal record. Instead, a `progress.md` change is allowed only when no changed line falls under `§E.2` (the run evidence body). Leader ruling R288 sets the boundary: rewriting `§E.3` (the run audit-ready signal) after a PASS is the normal flow and keeps the receipt. AC-AEC-003b pins the `§E.2` refusal and AC-AEC-003a pins the `§E.3` acceptance.
- **Remaining residual.** An evidence file deliberately named into the audit-artifact family, such as an evidence ledger called `verdict-notes.md`, would be accepted.

Rejected: a code-tree digest that ignores the head. It wrongly accepts everything the allowlist accepts, plus history that reaches the same content without descending from the judged commit. It would also change the digest format at an `@MX:ANCHOR` seam.

### B4 — CLI verb on the human path (R287 Q6), no MCP path

The verb reuses the human-path checks in `internal/contract/sign`, applied in order:
1. agent-environment markers;
2. `term.IsTerminal` on standard input, which refuses `</dev/null`;
3. the typed confirmation;
4. the git identity.

An MCP tool is excluded because every MCP call comes from an agent session, so the audited lane could clear its own rejection.

- **Terminal-forgery residual risk.** A local process running with the operator's privileges can allocate a pseudo-terminal (for example with `script` or `expect`) and unset the agent-environment markers. These checks stop the audited party from clearing its own rejection through its normal tool surface. They are not a cryptographic boundary.
- **PreToolUse deny: recorded as a Gap, not built here.** A PreToolUse deny of the verb in agent Bash calls would add one more layer. It is not small inside this SPEC: it needs a new matcher in the branch-guard hook family, an opt-in configuration key following that family's convention, and its own tests. The requirement budget is also at the Tier M ceiling. Doing it here would push the SPEC over the ceiling, so it belongs on a follow-up card.

### A1 — home of the repair-scope paragraph: the audit artifact convention (chosen) versus spec-workflow.md (rejected)

Chosen: `.moai/docs/audit-artifact-convention.md` and its byte-identical template mirror.
- Both auditor definitions already cross-reference the convention, so the auditor judging a fix-confirm round reads the limit where it reads the export rules.
- M2 already edits the same file for the `INCONCLUSIVE` label and the `blocking_defect:` line.
- The convention loads only when cited.

Rejected: `.claude/rules/moai/workflow/spec-workflow.md`. It loads into every session that touches `.moai/specs/**`, and its subject is the phase lifecycle.

## §F Out of Scope

### Out of Scope — release and binary
- Landing in the rc.39 candidate (`945cdb60fc`); this SPEC targets the next release candidate.
- Replacing or reinstalling the installed `moai` binary as part of the repair.

### Out of Scope — audit policy beyond the six items
- A default delta scope or a chain check for `codex_audit`, `glm_audit`, or `claude_audit` called directly, and for an `audit_multi` call without a `card_id`; those keep the t1708 behavior.
- A chain check inside the closure reader `SelectSecondReview`; the chain is enforced where records are written.
- Changing the convergence policy, `overall_verdict` values, `disagreement_flag`, or the per-backend `verdict` enum.
- Clearing served-model (`KindServed`) rejections; only receipt-kind rejections are covered.
- A PreToolUse deny of the B4 verb (recorded as a Gap in §E).

### Out of Scope — pre-existing divergence
- The pre-existing difference between `.claude/agents/moai/plan-auditor.md` and its template mirror at line 64 (leader ruling R287).
