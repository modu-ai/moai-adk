---
id: SPEC-LANDING-VERIFY-MERGE-SPLIT-001
title: "Re-measure completeness gate for landing: per-card required-command manifest, attempt records with package terminal events, whole-manifest verdict on one tree, sync-audit and SHA-bound receipt binding to the current HEAD, and the ordered enable switch"
version: "0.3.0"
status: draft
created: 2026-10-18
updated: 2026-10-19
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/factory, internal/factorylane, internal/auditreceipt, internal/cli, internal/config, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "landing, remeasure, completeness-gate, manifest, attempt-record, audit-head-binding, receipt-commit, card-t1872, git-flow"
tier: L
related_specs: [SPEC-MERGE-WINDOW-QUEUE-001, SPEC-MERGE-WINDOW-QUEUE-002, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-VERIFY-RUN-REUSE-001]
---

# SPEC-LANDING-VERIFY-MERGE-SPLIT-001

Card t1872 (Class C, P0), narrowed to **piece A** (board `9dbe`) and revised after plan-audit iteration 1 (FAIL 0.73) under the leader's rulings of board `0577`. Both are recorded verbatim in `decision-index.md`. Evidence: `research.md`. Design and threat analysis: `design.md`.

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Six-part draft; Tier L, 22 REQ / 22 AC. |
| 0.2.0 | 2026-10-19 | Narrowed to piece A per board `9dbe`; 12 REQ / 12 AC. |
| 0.3.0 | 2026-10-19 | Iteration 1 fixes: receipt commit field and its writer (D1, P4); attempt-command binding, twelfth reject class, per-package test counts (D2, D10); zero-Go cards (D3); manifest producer surface (D4); switch moved out of the branch, landing binary prerequisite, no post-audit backfill, recovery path (D5, D6, P1-P3, P5); embed owners and deleted packages (D9); milestone headings made machine-readable (CN-4). 15 REQ / 15 AC. |

## 1. Problem

The landing path accepts a re-measure record when one command exited 0 on the candidate tree (`internal/factory/integration_remeasure.go:135` `ValidateRemeasureRecord`; for a command that is not `go test -json` it keeps command and exit code only, :185-193, :698). Nothing records which calls were required, that each ran to the end, or that all of them passed on one tree. Observed (`research.md` §3): landed tree `1a3dc230c35ca0e2a997fc0fc3ac9444eb1f9198` (card t1680) was gated by one script-wrapped record with `structured_required=false`; the other required call's result lived only in a progress note.

For a SPEC card the merge-readiness sync-audit condition reads only `sync_status: complete` from progress (`internal/factorylane/merge.go` `checkSyncAudit`), so an earlier PASS can stand for the audit of a later HEAD. The operator decided (board `2c51`) that `complete` binds the verdict, and the required backend's receipt, to the card's current HEAD. An audit receipt records no commit today: `auditreceipt.Receipt` is `receipt_id, tool, tree_root, root_source, created_at, codex_verdict, gate_unmet, codex_error` (`internal/auditreceipt/store.go:166-177`) and its only production writer sets none of a commit (`internal/cli/mcp_audit_receipt.go:70-80`); board `0577` P4 approves adding the audited commit to new receipts.

## 2. Terms

- **Candidate tree / HEAD** — tree and commit of the card branch tip being landed. **Integration root** — the shared root the landing verbs resolve for the record store and configuration (`integrationLockRoot()`; configuration is read as `config.NewLoader().Load(filepath.Join(root, ".moai"))`, the pattern of `candidateCIEnabled`, `internal/cli/integration_merge.go:210-216`).
- **Manifest** — the per-card list of required verification calls. **Call** — one entry. **Attempt** — one execution of a call, with an attempt ID.
- **Package terminal event** — the `go test -json` event with `Action` pass, fail or skip, no `Test` field, for one `Package`.
- **Gate-complete on tree T** — every manifest call has a counted attempt on T whose package terminal events cover the call's whole package set, zero missing, none failed, with a non-zero executed-test count per package unless the manifest recorded `no_test_files` for it (REQ-LVM-004).
- **b2cd scope** — the standing landing scope: the Go packages the card changed (test-only and package-directory non-Go changes included) plus the Go packages that import them, each run whole with no `-run`; the one recorded narrowing allowance is `internal/cli` (rule 35e9). These are machine-local board rulings; this SPEC commits them to the tree.
- **Structured form** — one `moai integration remeasure --call <call-id> -- go test … -json …` invocation per manifest call.
- **Bound receipt** — a receipt whose audited-commit field equals the card HEAD (REQ-LVM-009/010).

## 3. Requirements (GEARS)

- **REQ-LVM-001 (event).** When the manifest producer runs for a card whose gate is enabled, it shall write a manifest listing: one call that runs each changed Go package whole (`go test -count=1 -p 1 -json`), where a package counts as changed when a file under its directory changed, test-only, embedded and testdata files included; one such call for each direct importer of a changed package, by non-test, test and external-test imports; one such call for each package that embeds a file belonging to a changed package directory; for every changed non-Go file outside a Go package directory, the reader calls found by a search of the repository's test sources, or an explicit zero-readers record naming the search command and its empty result; and, for a package the card deleted, a record comparing the base and candidate package graphs instead of a call. Each call shall record its identifier, exact command, and resolved package set; a call for `internal/cli` alone may carry a `-run` narrowing and shall name it as the 35e9 allowance.
- **REQ-LVM-002 (event).** When a manifest call is executed in the structured form, the re-measure verb shall write an attempt record carrying a unique attempt ID, the call identifier, the exact command argv, the exit code, for every package in the stream its terminal event and its count of executed tests, the candidate tree SHA, the card HEAD SHA, the base, an environment fingerprint (GOOS, GOARCH, CGO_ENABLED, GOFLAGS, toolchain version, build tags, declared environment variables), the start and end times, the judging build identity, and a termination class.
- **REQ-LVM-003 (event).** When an attempt ends by signal, by timeout, or with a stream that ends on a truncated line, the verb shall classify it as aborted and retain its record, and an aborted attempt shall never count as a pass; a later execution of the same call shall be a new attempt with a new ID that does not overwrite the aborted record, and only the most recent completed attempt per call on the same tree shall count.
- **REQ-LVM-004 (ubiquitous).** The completeness verifier shall report gate-complete only when every call of the card's manifest has a counted attempt on the candidate tree, each counted attempt exited 0 with a terminal event for every package of the call's resolved set (zero missing), no failure, and a non-zero executed-test count per package unless the manifest recorded `no_test_files` for that package; for a manifest of zero calls it shall require the evidence block of REQ-LVM-013 instead; the most recent record alone shall never decide the verdict.
- **REQ-LVM-005 (event).** When any of the following holds, the verifier shall refuse readiness: (1) the manifest is missing; (2) a call has no counted attempt; (3) a call's only record is a script-wrapped command with no recognized package report; (4) the stream lacks the terminal event of an expected package; (5) the stream is cut mid-line; (6) the attempt is aborted; (7) the attempt names a different tree or HEAD than the candidate; (8) the environment fingerprint differs from the manifest's; (9) a changed non-Go input has neither reader call nor zero-readers record; (10) the b2cd scope cannot be derived; (11) a counted package executed zero tests where the manifest does not allow it; (12) the attempt's command differs from the manifest call's command — exact argv equality of the `go test` segment after dropping leading environment assignments and an `unset … &&` scrub segment, any other segment, pipe or compound refusing, and a `-run` selector accepted only on the `internal/cli` call that names 35e9.
- **REQ-LVM-006 (ubiquitous).** The completeness verifier shall be the single verifier consulted by the merge step, by `moai factory complete`, by merge-readiness condition (d) and by `moai integration manifest verify`, and while the gate is enabled a record from a command that carries no recognized package report shall not satisfy any of them.
- **REQ-LVM-007 (state).** While `workflow.landing_gate.enabled` is false or absent in the integration root's configuration, the landing path shall behave exactly as it does today; the key shall ship false in the distributed template; a copy of the key in a card branch's configuration shall have no effect; and the narrowing allowance list shall ship containing `internal/cli` only and shall not be widened by this SPEC.
- **REQ-LVM-008 (event).** When `moai factory complete` evaluates a SPEC card while the gate is enabled, it shall refuse unless the card's sync-audit verdict names the card's current HEAD as its audited commit and is admitted by the shared verdict predicate with the configured required backends; the audited commit may differ from HEAD only under the card's own evidence directory `.moai/reports/<card-id>/` (the allowance `internal/factorylane/merge_specless.go` already applies), no other path is exempt, and a verdict for any earlier commit shall not substitute, even when the change since is empty or patch-identical.
- **REQ-LVM-009 (event).** When an audit call records a receipt, the receipt writer shall stamp the audited commit — the HEAD of the audited tree read before the backends run — as an additive field, shall leave the field empty when that tree's HEAD differs when the receipt is written, and shall never rewrite or delete an existing receipt; a receipt without the field shall remain a valid record and shall be read as unbound.
- **REQ-LVM-010 (event).** When `moai factory complete` evaluates a card while the gate is enabled, it shall refuse unless every receipt the sync-audit verdict cites is readable, is a codex or multi-backend audit receipt with an empty unmet-gate field and a pass verdict, and carries an audited commit equal to the card's current HEAD; an absent, unreadable, older-HEAD, empty-commit or legacy receipt shall each refuse.
- **REQ-LVM-011 (ubiquitous).** The landing rules shall name, in the committed rule files and their template mirrors, the structured form and the commands a lane runs (`moai integration manifest derive`, `moai integration remeasure --call`, `moai integration manifest verify`), the restart of an aborted call as a new attempt, the rule that no evidence or progress backfill commit lands after the audited SHA (bookkeeping is committed before the final audit), and the prerequisite that the governed landing commands are invoked by path from a binary built from the landing HEAD's main and that the receipt-writing server runs a build containing REQ-LVM-009.
- **REQ-LVM-012 (ubiquitous).** The branch carrying this SPEC's milestones shall never enable the gate in the repository's own configuration; the enabling shall be a separate later commit on the integration branch after the SPEC has landed and a binary built from the landing HEAD is in use, changing the one configuration path only; and the reversal of that key shall be landable as an ordinary one-path card through the zero-Go path of REQ-LVM-013 with the audit binding of REQ-LVM-008/010 intact, not through any bypass.
- **REQ-LVM-013 (event).** When a card changes no Go package and the reader search finds no reader call, the manifest shall be empty with an evidence block — the changed-path list, the `go list` proof that no Go package directory contains a changed path, and the reader-search result — and the verifier shall recompute the changed-path list from the repository and admit the card only when it equals the recorded list; a missing, stale or inconsistent evidence block shall refuse, and the audit binding of REQ-LVM-008/010 shall apply unchanged.
- **REQ-LVM-014 (event).** When a lane runs `moai integration manifest derive --card <id>`, the producer shall write the manifest to `.moai/state/remeasure/manifests/<card-id>.json` under the integration root, shall be re-run after every re-absorb (a regenerated manifest replaces the old one, and attempts bound to the superseded manifest identity stop counting), and `moai integration manifest show` and `moai integration manifest verify` shall print the manifest and the verifier's verdict with the failing class named, the latter exiting non-zero on refusal.
- **REQ-LVM-015 (ubiquitous).** The gate shall not skip or weaken any audit check, shall not accept an earlier PASS for a later HEAD, shall not allow a manifest call to select fewer tests than its whole package outside the named `internal/cli` allowance, shall not write the integration branch except through the existing merge step, shall not widen any default permission or narrowing allowance, and shall treat every undecidable reading as refusal.

## 4. Acceptance index

AC-LVM-001 to AC-LVM-015 map one-to-one to REQ-LVM-001 to REQ-LVM-015 (`acceptance.md`, Given-When-Then with command, exit-code/stdout criterion and a RED-now ledger).

## 5. Constraints and prerequisites

- C1. Git-flow window path only; the github-flow pull-request edge (`internal/cli/factory_card.go:2071-2073`) is not touched.
- C2. Existing audit checks are consumed as they are; this SPEC adds readers, the receipt commit field and bindings, never a bypass.
- C3. The gate ships off in the template and stays off on this SPEC's branch (REQ-LVM-012). Piece A lands under the current rules; its own landing is not the gate's first live use.
- C4. No time estimates; milestones are ordered by dependency.
- C5. **Prerequisites, not steps of this card** (board `0577` P1): a gate counts as ON only where the judging build contains it. The governed commands (remeasure, `complete`, merge step) are invoked by path from a binary built from the landing HEAD's main, never through a PATH-resolved installed binary that may lag; no reinstall of the installed binary is implied. The MCP server that writes receipts must run a build containing REQ-LVM-009 before a receipt is accepted as bound.
- C6. Cards already merge-ready at the switch are not re-audited for it: each still takes the HEAD-confirmation audit at landing (board `2c51`), and its landing re-measure is done in the structured form once the key is on. Measured transition picture: `research.md` §10.

## 6. Tier judgment (re-judged with the receipt work)

Tier L. Arithmetic from measured analogues and the touch list:

- Nearest analogue: the re-measure record module, `integration_remeasure.go` 740 + `cli/integration_remeasure.go` 131 + tests 470 + 261 + 132 = 1,734 lines. The gate generalizes it (manifest + attempts + verifier + producer verbs) and adds the audit binding beside `factorylane/merge.go` (471) / `merge_specless.go` (154), and the receipt writer and schema (`auditreceipt/store.go`, `cli/mcp_audit_receipt.go`). Estimate: production ≈ 1,150-1,450 lines, tests ≈ 1,700-2,100 lines — above the Tier M ceiling of 1,000.
- Files: new production ≈ 5; new test files ≈ 8 (factory, factorylane, auditreceipt, cli); edited production ≈ 12 (`integration_remeasure.go`, `integration_merge_step.go`, `factory_card.go`, `factory_merge.go`, `cli/integration_remeasure.go`, a new `cli` manifest verb file, `factorylane/merge.go`, `factorylane/merge_specless.go`, `auditreceipt/store.go`, `cli/mcp_audit_receipt.go`, config types/defaults/loader); config ≈ 2 (template default, defaults); rules ≈ 2 files plus 2 mirrors. About 33 files — above the Tier M limit of 15.
- Counts: 15 requirements and 15 criteria against the Tier L ceiling of 25 each.

## 7. Follow-up pieces (not in this SPEC)

Recorded per board `9dbe`; each is issued as its own card only after piece A lands. None is specified here.

- A slot piece: removal of the lease record of a dead owning session; no new liveness semantics; capacity stays 1; the wall-clock package list stays empty until measured.
- A readiness-and-metering piece: a derived ready view, landing-event metrics, a shadow counter of admissible reuse; it depends on piece A.
- The re-measure reuse relaxation is refused for now and not built; no allowed non-Go glob list and no hermetic-package list is guessed.

## 8. Exclusions

### Out of Scope — landing paths and policy

- The github-flow pull-request delivery edge, CodeRabbit gating, and CI-side verification.
- Skipping or shortening the HEAD confirmation audit for any reason, patch-id equality included.
- Train stacking, a second writer of the integration branch, queue admission, lane leasing and card classification.
- Re-auditing cards because of the switch itself, and reinstalling the installed binary.

### Out of Scope — verification content

- Choosing which tests a card should add; the manifest records what the b2cd rule requires.
- Forwarding the child stdout of `moai integration remeasure` (the known swallow defect); a script-wrapped call does not satisfy the gate while it is enabled.
- Reusing a record across a moved integration tip, and any base-equals-tip relaxation.

### Out of Scope — slot subsystem and views

- Slot capacity, load and memory headroom, wall-clock-bound package exclusion, a slot wrapper verb, any change to lease liveness semantics.
- A ready-to-land view and landing metrics.
