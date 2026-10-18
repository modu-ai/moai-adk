---
id: SPEC-LANDING-VERIFY-MERGE-SPLIT-001
title: "Landing audit binding: SHA-bound audit receipts, the sync-audit verdict and its receipt bound to the card's current HEAD at complete, a committed-branch gate switch, and the ordered enable (piece A2)"
version: "0.9.0"
status: draft
created: 2026-10-18
updated: 2026-10-19
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/auditreceipt, internal/auditverdict, internal/factory, internal/factorylane, internal/homestate, internal/cli, internal/config, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "landing, audit-head-binding, receipt-commit, gate-switch, committed-blob-gate, card-t1872, git-flow"
tier: L
related_specs: [SPEC-MERGE-WINDOW-QUEUE-001, SPEC-MERGE-WINDOW-QUEUE-002, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-VERIFY-RUN-REUSE-001]
---

# SPEC-LANDING-VERIFY-MERGE-SPLIT-001

Card t1872 (Class C, P0), **piece A2** after the split ruled by board `d20e`: the audit binding and the switch kernel stay here and land first; the re-measure completeness gate (piece A1) moved to card t1875 and SPEC-LANDING-COMPLETENESS-GATE-001, which is blocked until this card is merged-local. Piece A2 ships the audit binding and the switch reader with the key off; the key is switched on only after A2 and A1 have both landed (decision Q27, board `6404`). The rulings are recorded verbatim in `decision-index.md`. Evidence: `research.md` (§12 proves A2 independent of A1 by code). Design and threat analysis: `design.md`. The text of the whole 18-requirement SPEC before the split is preserved at commit `9713b6757d` (`git show 9713b6757d:.moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/spec.md`, and the same for `acceptance.md`, `design.md`, `plan.md`).

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Six-part draft; Tier L, 22 REQ / 22 AC. |
| 0.2.0 | 2026-10-19 | Narrowed to piece A per board `9dbe`; 12 REQ / 12 AC. |
| 0.3.0 | 2026-10-19 | Iteration 1 fixes: receipt commit field, attempt-command binding, zero-Go cards, producer verbs, switch outside the branch, recovery path; 15 REQ / 15 AC. |
| 0.4.0 | 2026-10-19 | Iteration 2 fixes: manifest currency by re-derivation; gate state read from the integration branch's committed blob with fail-modes (board `470c`); verdict locator and `audit_receipt:` grammar with a non-vacuous refusal; one equals-HEAD predicate; receipt tree check, entry-point capture, mixed-citation rule. 18 REQ / 18 AC. |
| 0.5.0 | 2026-10-19 | Iteration 3 delta (board `5d4d`), ids unchanged: equals HEAD with no tolerance; receipt looked up at the audited tree's store root and compared with the audited tree's own path; integration-branch resolution fail-modes; SPEC-less cards bound through `verdict.md`; audit target recorded. 18 REQ / 18 AC. |
| 0.6.0 | 2026-10-19 | Single ceiling exception (board `4812`), ids unchanged: effective GOFLAGS rule, module files select the whole module, `auditreceipt.Canonical` named, refusal semantics, receipt-side empty-diff mutant. Plan audit iteration 5 FAIL 0.86 at `9713b6757d`; the three blockers sit in the completeness gate. 18 REQ / 18 AC. |
| 0.7.0 | 2026-10-19 | Split per board `d20e`: this SPEC keeps REQ-LVM-007, 008, 009, 010, 012, 017, 018 (ids stable); REQ-LVM-001-006, 011, 013-016 and the plan-audit blockers B1-B3 moved to SPEC-LANDING-COMPLETENESS-GATE-001 (§7). Adjusted only where an A1 concept appeared: REQ-007 (narrowing-allowance clause removed), REQ-012 (manifest wording replaced; the audit-side lane guidance of the old REQ-011 attached), REQ-017 (consumers named for A2, observability through `moai integration status`), REQ-008 (placement before the window). Independence proof by code in `research.md` §12. 7 REQ / 7 AC. |
| 0.8.0 | 2026-10-19 | Rulings of board `6404`, ids and counts unchanged: the key is switched on only after A2 and A1 have both landed (Q27; spec §2, C3, C10, REQ-LVM-012 wording); the audit binding stays in `complete` only and the direct merge path binding is an acceptance requirement of A1 (Q28; §7 row 12, §8, design R6). |
| 0.9.0 | 2026-10-19 | Plan-audit A2 round 1 (FAIL 0.81, `plan-audit-a2.md`) and board `9b0e`, ids and counts unchanged (7 REQ / 7 AC): the characterization test of today's `complete` lands in the M0 commit, ahead of the commit that edits `factoryCompleteCard` (AC-LVM-007, AC-LVM-012 graph legs); the gate-state reader resolves `refs/heads/<branch>^{commit}` once so a same-named tag cannot shadow the branch (REQ-LVM-017, new test `TestSameNamedTagDoesNotShadowBranch`); the audit row has a named sink; M4 regenerates the Codex agent copy with `make agents-emit`; the REQ id gap is waived by an explicit clause (§3, row Q30); terms "audited tree" and abbreviated `audited_sha` defined; the hand-off gains three inherited obligations. |

## 1. Problem

The sync-audit verdict that gates a landing is bound only at stage time. The T13 stage guard admits a verdict whose `audited_sha` is a prefix of the evidence SHA recorded then (`internal/homestate/card_evidence_readers.go:124`, `parseAuditVerdict`), merge-readiness condition (a) reads only `sync_status: complete` from progress (`internal/factorylane/merge.go:245`), and `moai factory complete` gates on the card's stage and lease (`internal/cli/factory_card.go:2096-2109`) without reading the verdict. A commit made after the audit therefore lands under the earlier PASS. The operator decided (board `2c51`) that `complete` binds the verdict, and the required backend's receipt, to the card's current HEAD.

An audit receipt records no commit today (`internal/auditreceipt/store.go:166-177`; writer `internal/cli/mcp_audit_receipt.go:60-100`, called from `mcp_codex.go:2112, 2163` and `mcp_convergence.go:928`); board `0577` P4 approves adding the audited commit to new receipts. No Go code parses a receipt id from a verdict file (`ParseVerdictLine`, `store.go:1223`, reads the final-message `AUDIT-VERDICT:` line), and measured sync-audit verdict files of merge-ready SPEC cards cite receipts in mixed forms (`research.md` §10); the sync workflow already names an orchestrator binding-statement line `audit_receipt: <ids|none>` (`.claude/skills/moai/workflows/sync.md:91`), which this SPEC makes strict and machine-read at landing. The completeness half of the original problem — a re-measure record that one script-wrapped command can satisfy — is the subject of SPEC-LANDING-COMPLETENESS-GATE-001.

## 2. Terms

- **Candidate tree / HEAD** — tree and commit of the card branch tip being landed. **Integration branch** — the branch named by the held window record or else by the common-root integration configuration (`configuredIntegrationBranch()`, `internal/cli/integration.go:291`). A project with neither has not opted in to the git-flow window, and for the gate that means off (REQ-LVM-017); the merge verb's own fallback to `develop` is not used to decide the gate.
- **The gate** — the condition `workflow.landing_gate.enabled` read from the integration branch's committed content (REQ-LVM-017). While on, `moai factory complete` applies the audit binding of REQ-LVM-008/010/018. In this SPEC the gate has no other effect; SPEC-LANDING-COMPLETENESS-GATE-001 adds the completeness checks and the direct-merge-path binding under the same key. This SPEC ships the key off, and the key is switched on only after both pieces have landed (C10).
- **Audited tree** — the card's worktree directory whose HEAD commit is the card's HEAD; REQ-LVM-009 captures its HEAD at audit entry and REQ-LVM-010 compares a receipt's `tree_root` with its canonical path.
- **Equals HEAD** — the audited commit is exactly the card's current HEAD commit, written as the full 40-hex or as an abbreviation of 12 to 40 hex digits that is a prefix of HEAD (the form the stage reader accepts, `internal/homestate/card_evidence_readers.go:21`); a shorter string, or one naming another commit, does not equal HEAD, and an abbreviation is not a tolerance of the diff. There is no tolerance: an ancestor of HEAD, whether its diff to HEAD is empty, patch-identical or confined to `.moai/reports/<card-id>/`, does not equal HEAD (boards `0577` P3 and `5d4d`).
- **Bound receipt** — a receipt whose audited-commit field equals HEAD (REQ-LVM-009/010).

## 3. Requirements (GEARS)

- **REQ-LVM-007 (state).** While the gate is not on (REQ-LVM-017), the landing path shall behave exactly as it does today, and the key `workflow.landing_gate.enabled` shall ship false in the distributed template and in the configuration defaults.
- **REQ-LVM-008 (event).** When `moai factory complete` evaluates a card while the gate is on, it shall locate the card's verdict — for a SPEC card the newest `sync-audit*.md` under `.moai/reports/<card-id>/` (highest `-iter<N>`, a name without it counting as iteration 1, then latest modification); for a SPEC-less card the file `.moai/reports/<card-id>/verdict.md` that its readiness path already reads — shall refuse when none exists, and shall refuse unless the verdict's audited commit equals HEAD and the verdict is admitted by the shared verdict predicate with the configured required backends; a verdict for any earlier commit shall not substitute, even when the change since is empty or patch-identical, and a commit made after the audited commit shall refuse whatever paths it touches; every refusal shall precede the integration window and every record change.
- **REQ-LVM-009 (event).** When an audit call is entered, the entry point shall capture the HEAD of the audited tree and the review target the call names before the backends run, and when the receipt writer records the receipt it shall stamp them as the additive fields `audited_commit` and `audit_target`, leaving `audited_commit` empty when the audited tree's HEAD differs at write time; every call site of the writer (`mcp_codex.go` twice, `mcp_convergence.go` once) shall pass the captured values; the writer shall never rewrite or delete an existing receipt, and a receipt without the fields shall remain a valid record read as unbound.
- **REQ-LVM-010 (event).** When `moai factory complete` evaluates a card while the gate is on, it shall refuse unless at least one receipt cited by the verdict qualifies — looked up at the store root of the audited tree (for a config-orphaned worktree, the primary checkout), readable, a codex or multi-backend audit receipt, with `tree_root` equal to `auditreceipt.Canonical` of the audited tree's path (symlinks resolved, else cleaned — the normalization the writer applies, `internal/auditreceipt/treeroot.go:37`; not the store root), an empty unmet-gate field, codex verdict pass, and an audited commit equal to HEAD — and a failed or inconclusive extra cited receipt shall be recorded without blocking and shall never replace the qualifying one; an absent, unreadable, older-HEAD, empty-diff-ancestor, empty-commit, legacy, other-tree or unmet-gate receipt shall each fail to qualify.
- **REQ-LVM-012 (ubiquitous).** The branch carrying this SPEC's milestones shall never enable the gate in the repository's own configuration; the enabling shall be a separate later commit on the integration branch made only after this SPEC and SPEC-LANDING-COMPLETENESS-GATE-001 have both landed and a binary built from a HEAD containing both is in use, changing the one configuration path only; the reversal of that key shall land as an ordinary one-path card through the gated path, with the audit binding of REQ-LVM-008/010 intact, not through any bypass; and the committed rule files and their template mirrors shall state that the gate state is read from the committed content of the integration branch, the notice `gate_state_unreadable`, that no evidence or progress backfill commit lands after the audited SHA (bookkeeping is committed before the final audit), and the prerequisite that the governed landing commands are invoked by path from a binary built from the landing HEAD's main and that the receipt-writing server runs a build containing REQ-LVM-009.
- **REQ-LVM-017 (event).** When any landing command evaluates whether the gate is on, it shall resolve the integration branch as the held window record's branch, else the configured integration target; with none configured the project has not opted in and the gate shall read off; with one configured whose `refs/heads/<branch>` does not resolve to a commit (a tag of the same name does not count) the command shall refuse the landing with the named error `gate_state_integration_branch_unresolvable` (fail-closed: the merge step cannot run without that branch, so no landing is wedged, and repairing the branch configuration or ref needs no landing); with the branch resolved it shall resolve `refs/heads/<branch>^{commit}` exactly once and use that single commit SHA for the existence check, the blob read and the displayed source, so that a tag of the same name can never shadow the branch, shall read the committed blob `<sha>:.moai/config/sections/workflow.yaml`, consult no working tree, and tell an absent path (an existence check, `git cat-file -e`) from a failed read: an absent path, an absent key or an explicit false shall read off, an explicit true shall read on, and a failed read of a present path, an unparseable file or a non-boolean value shall read on with the notice `gate_state_unreadable` and an audit row appended to `.moai/logs/landing-gate-audit.jsonl` under the integration lock root (one JSON line carrying the event name, the branch, the SHA, the path, the cause and the UTC time) (a landing that satisfies the gate, including the card that repairs the file, is still admitted); a card branch's copy of the key shall have no effect, and an enabling or reverting commit on the integration branch shall take effect when it lands, whatever commit the primary checkout is on; the mutating command `moai factory complete` shall exit non-zero before any state changes when it refuses, and `moai integration status` shall print the gate state with its source as `gate: on|off (source: <branch>:<path> @<sha>)` together with the notice or the named error when one applies, without changing its own exit status.
- **REQ-LVM-018 (state).** While the gate is on, `complete` shall take the cited receipt ids from the located verdict file's single machine line `audit_receipt: rcpt-<id>[,rcpt-<id>…]` (ids matching `rcpt-[a-z0-9]{20,40}`, split on commas, no other characters) for a SPEC card and for a SPEC-less card alike, and shall refuse when the line is missing, duplicated, empty, `none` or malformed, or when a cited id is absent from the store; the sync-auditor protocol and the audit artifact convention, with the template mirror of the protocol and the Codex agent definitions emitted from it, shall state the `audit_receipt machine line`.

**REQ id gap waiver (board `d-20261018T173648Z-9b0e`, `9b0e`; decision row Q30).** The identifiers REQ-LVM-001 to 006, 011 and 013 to 016, and the matching AC-LVM-001 to 006, 011 and 013 to 016, are absent from this SPEC on purpose. Split `d20e` moved those requirements to SPEC-LANDING-COMPLETENESS-GATE-001 (card t1875); they are listed in the §7 hand-off table and are reserved there, never renumbered, deleted or reused. The leader ruled that the gap is a documented exemption from strict id contiguity, because renumbering would rewrite the A1 hand-off, this SPEC and every criterion reference.

## 4. Acceptance index

AC-LVM-007, 008, 009, 010, 012, 017 and 018 map one-to-one to the requirements of the same numbers (`acceptance.md`, Given-When-Then with command, exit-code/stdout criterion and a RED-now ledger).

## 5. Constraints and prerequisites

- C1. Git-flow window path only; the github-flow pull-request edge (`internal/cli/factory_card.go:2071-2073`) is not touched, and its readiness caller (`factory_card_pr.go:229`) is left unwired.
- C2. Existing audit checks are consumed as they are; this SPEC adds a reader, the receipt commit field and one binding, never a bypass.
- C3. The gate ships off in the template and stays off on this SPEC's branch (REQ-LVM-012). Piece A2 lands under the current rules and ships the audit binding and the switch reader with the key off; its own landing is not the gate's first live use, and neither is the landing of A1 that follows it.
- C4. No time estimates; milestones are ordered by dependency.
- C5. **Prerequisites, not steps of this card** (board `0577` P1): a gate counts as on only where the judging build contains it. The governed commands (`complete`, the merge step) are invoked by path from a binary built from the landing HEAD's main, never through a PATH-resolved installed binary that may lag; no reinstall is implied. The MCP server that writes receipts must run a build containing REQ-LVM-009 before a receipt is accepted as bound.
- C6. Cards already merge-ready at the switch (made only after both pieces have landed, C10) are not re-audited for it: each still takes the HEAD-confirmation audit at landing (board `2c51`) and, with the gate on, needs a verdict carrying the `audit_receipt` machine line and a bound receipt, which that new audit produces. Measured transition picture: `research.md` §10.
- C7. The integration branch name is read from the common-root configuration as ruled (board `470c`); that read is not the gate key and is unchanged by this SPEC.
- C8. **Independence from piece A1.** This SPEC introduces no manifest, attempt, verifier or producer-verb type and reads none; the completeness checks of A1 consume this SPEC's gate-state reader and its binding, not the reverse (`research.md` §12). Piece A1 is blocked until this card is merged-local because both edit `factoryCompleteCard`.
- C9. When the gate is on, the audit binding is the only gate effect of this SPEC: the re-measure record check of `complete` (`factory_card.go:2281`) and of the merge step (`integration_merge_step.go:321-325`) stays exactly as today until A1 lands.
- C10. **The key is switched on only after both pieces have landed** (decision Q27, board `6404`): a separate commit on main, judged by a build that contains both pieces (board `0577` P1). Reason: with A2 alone on, `complete` is bound while `moai integration merge --card` still bypasses the binding, a gate with a known side door (design R6); A1 adds the binding to the direct merge path (§7 row 12), so the gate never ships on with the side door.
- C11. The receipt rule of REQ-LVM-010 requires a codex or multi-backend receipt with codex verdict pass, so the gate is meaningful only where the audit plan requires codex (this repository: `moai verify audit-plan` reports `enforced_required: ["codex"]`); on a tree that does not require codex a gate-on `complete` would refuse every card, which is why the key ships off in the distributed template.

## 6. Tier judgment

Tier L. By file count alone this exceeds the Tier M limit; by production lines alone it sits within M, so the file count decides.

- Production: receipt schema and writer ≈ 60-90 lines; gate-state reader (new file in `internal/factory`) ≈ 150-200; audit binding, citation parser and receipt qualification (new files in `internal/factorylane`) ≈ 300; the shared newest-verdict helper moved to `internal/auditverdict` ≈ 40; `complete` wiring, `integration status` display, configuration type and defaults ≈ 110. Estimate ≈ 650-800 lines.
- Tests: 55 named tests (`acceptance.md`) at ≈ 25-35 lines each with fixtures ≈ 1,400-1,950 lines. The estimate is derived from the test count, not measured.
- Files: new production 4, new test files ≈ 7 (the M0 characterization test file included), edited production 9 (`auditreceipt/store.go`, `cli/mcp_audit_receipt.go`, `cli/mcp_codex.go`, `cli/mcp_convergence.go`, `cli/factory_card.go`, `cli/integration.go`, `config/types.go` with its defaults, `homestate/card_evidence_readers.go`, the template `workflow.yaml`), hand-edited rule and protocol text 5 (`factory-dispatch-mechanics.md` and mirror, `sync-auditor.md` and mirror, `.moai/docs/audit-artifact-convention.md`), generated or copied 2 (the template `.codex/agents/moai/sync-auditor.toml`, emitted by `make agents-emit`, and the local `.codex/agents/moai/sync-auditor.toml` refreshed from it). About 27 files — above the Tier M limit of 15.
- Counts: 7 requirements and 7 criteria against the Tier L ceiling of 25 each.

## 7. Moved to SPEC-LANDING-COMPLETENESS-GATE-001 (card t1875)

Piece A1 (the re-measure completeness gate) moved by board `d20e`. It is blocked until this card is merged-local and starts from the pre-split texts at `9713b6757d`. The reserved identifiers and what they carried:

| # | Moved id | Requirement, one line | Criterion that moved with it |
|---|---|---|---|
| 1 | REQ-LVM-001 | manifest derivation: changed packages whole, importers, embed owners, non-Go reader search, deleted packages, module files select the whole module | AC-LVM-001 |
| 2 | REQ-LVM-002 | attempt record: id, argv, exit code, per-package terminal event and test count, tree, HEAD, base, environment fingerprint, build identity, termination class | AC-LVM-002 |
| 3 | REQ-LVM-003 | abort versus restart: an aborted attempt never passes; a rerun is a new attempt | AC-LVM-003 |
| 4 | REQ-LVM-004 | whole-manifest verdict on one tree, never the last record alone | AC-LVM-004 |
| 5 | REQ-LVM-005 | the twelve attempt-level reject classes (command binding, GOFLAGS rule, empty sweep) | AC-LVM-005 |
| 6 | REQ-LVM-006 | one verifier consulted by the merge step, `complete`, condition (d) and `manifest verify` | AC-LVM-006 |
| 7 | REQ-LVM-011 | lane guidance for the structured form: `manifest derive`, `remeasure --call`, `manifest verify`, restart as a new attempt (the audit-side half is kept here in REQ-LVM-012 and REQ-LVM-018) | AC-LVM-011 |
| 8 | REQ-LVM-013 | zero-Go card: empty manifest with derivation evidence, never for module-file changes | AC-LVM-013 |
| 9 | REQ-LVM-014 | manifest producer verbs and stored location | AC-LVM-014 |
| 10 | REQ-LVM-015 | safety invariants of the gate (no weakened check, no widened allowance; the audit leg is carried by REQ-LVM-008 here) | AC-LVM-015 |
| 11 | REQ-LVM-016 | manifest currency by re-derivation under the recorded environment | AC-LVM-016 |
| 12 | (new, inherited from decision Q28, board `6404`) | the audit binding on the direct merge path: merge-readiness condition (a) (`factorylane/merge.go:174`, `:245`) and the merge step's landing check (`factory/integration_merge_step.go:321-325`) apply the same binding function as `complete` (equals HEAD, bound receipt, citation line), so a lane that skips `complete` and calls `moai integration merge --card` is refused by the same causes. A2 binds `complete` only and accepts the residual (design R6) because the key stays off until A1 has landed | a new criterion of A1 (A2 ids are not reused) |

Decision rows that move with them (`decision-index.md`, kept as history here): Q4, Q5, Q15, Q16, Q17, Q25, Q26. Rows Q6 and Q8 (reuse) remain not built (board `e4cc`).

**Inherited acceptance requirements for A1** (plan-audit iteration 5 at `9713b6757d`, blockers B1-B3 and the optional items; A1's first audit must close them):

- **B1 — effective environment under the command's own prefix.** The effective environment of a call is the verb's environment transformed by the command's leading assignments, `env` options and `unset` segments; GOFLAGS, GOENV, GOTOOLCHAIN, GOOS, GOARCH, CGO_ENABLED and tags are taken from that environment (or any leading assignment, `env` option or `unset` naming one of those variables is refused). Acceptance adds the command-prefix negative `GOFLAGS=-run=^TestGood$ go test -count=1 -p 1 -json ./p` set nowhere else, with the failing-package fixture, a mutant row ("observation taken in the verb's own environment") and the named test `TestGateRejectsNarrowingGoflagsInCommandPrefix`. The shell leading-assignment gap is the defect: today `exec.Command("sh", "-c", command)` runs with the inherited environment (`internal/factory/integration_remeasure.go:647`).
- **B2 — allowlist inversion.** Replace the GOFLAGS denylist by an allowlist: the effective GOFLAGS may carry only flags the SPEC names as not altering the test set (`-mod=readonly`, `-mod=vendor`, `-trimpath`, `-buildvcs=false`); every other flag refuses, including `-overlay`, `-tags`, `-exec`, `-toolexec`, `-modfile`, `-vet`, `-race`, `-coverpkg` and any `-test.*` or `--` spelling. State it in the reject class and in the safety-invariant requirement, and add the `-overlay` negative (a failing test hidden by an overlay file).
- **B3 — producer-verb refusals bound to tests.** Add `TestManifestDeriveRefusesNarrowingGoflags` and `TestRemeasureCallRefusesNarrowingGoflagsBeforeRun` (the second asserting that no child process starts) to the criterion of the reject classes (N rises to 17) or to the producer-verb criterion, with the RED-now cell updated.
- **S1 — accepted-cost wording.** State in the design and in Q26 that a `go.mod`/`go.sum` change runs every package of the module including `internal/cli` without the 35e9 narrowing, and which `go.mod` is meant (this repository: one `go.mod`, no `go.work`).
- **O1 — trust boundary sentence.** The gate defends against accidental narrowing, not against an actor who controls the `go` binary, PATH or the record.
- **Narrowing-allowance list (from the old REQ-LVM-007).** The list ships containing `internal/cli` only and is not widened; the test `TestRunNarrowingAllowanceInternalCLIOnly` belongs to A1.
- **Key-off card manifest (from the old REQ-LVM-012).** A card that changes only `.moai/config/sections/workflow.yaml` has a derived manifest holding the reader calls of that file (tests in this repository read it), and lands through the gated path; this SPEC keeps only the audit-binding half.
- **Consumers and read-only exit semantics (from the old REQ-LVM-017).** The merge step, merge-readiness condition (d) and `moai integration manifest verify` read this SPEC's gate-state reader; for read-only evaluations (condition (d), `manifest verify`) the named error is printed as the refusal and the exit is non-zero, nothing being written.

## 8. Exclusions

### Out of Scope — landing paths and policy

- The github-flow pull-request delivery edge, CodeRabbit gating, and CI-side verification.
- Skipping or shortening the HEAD confirmation audit for any reason, patch-id equality included.
- Train stacking, a second writer of the integration branch, queue admission, lane leasing and card classification.
- Re-auditing cards because of the switch itself, reinstalling the installed binary, and moving or fixing the primary checkout.
- Binding the direct merge verb (`moai integration merge --card`) and merge-readiness condition (a) to the audit binding; decision row Q28 assigns that binding to piece A1 (§7 row 12), and the key stays off until it lands.

### Out of Scope — verification content

- The re-measure completeness gate in every part (manifest, attempts, verifier, producer verbs, zero-Go evidence): SPEC-LANDING-COMPLETENESS-GATE-001.
- Forwarding the child stdout of `moai integration remeasure` (the known swallow defect).
- Reusing a record across a moved integration tip, and any base-equals-tip relaxation.

### Out of Scope — slot subsystem and views

- Slot capacity, load and memory headroom, wall-clock-bound package exclusion, a slot wrapper verb, any change to lease liveness semantics.
- A ready-to-land view and landing metrics.
