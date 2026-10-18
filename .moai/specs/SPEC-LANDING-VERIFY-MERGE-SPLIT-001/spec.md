---
id: SPEC-LANDING-VERIFY-MERGE-SPLIT-001
title: "Re-measure completeness gate for landing: per-card required-command manifest, attempt records with package terminal events, whole-manifest verdict on one tree, sync-audit binding to the current HEAD, and the ordered enable switch"
version: "0.2.0"
status: draft
created: 2026-10-18
updated: 2026-10-19
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/factory, internal/factorylane, internal/cli, internal/config, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "landing, remeasure, completeness-gate, manifest, attempt-record, audit-head-binding, card-t1872, git-flow"
tier: L
related_specs: [SPEC-MERGE-WINDOW-QUEUE-001, SPEC-MERGE-WINDOW-QUEUE-002, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-VERIFY-RUN-REUSE-001]
---

# SPEC-LANDING-VERIFY-MERGE-SPLIT-001

Card t1872 (Class C, P0), narrowed to **piece A** by the leader's ruling of 2026-10-18 (board `9dbe`, recorded verbatim in `decision-index.md`). The SPEC id is unchanged. Baseline and code-level evidence: `research.md`. Design and threat analysis: `design.md`.

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Six-part draft (gate, slot auto-release and capacity, metrics, ready view, reuse); Tier L, 22 REQ / 22 AC. |
| 0.2.0 | 2026-10-19 | Narrowed to piece A per board `9dbe`: baseline M0, the completeness gate, the Q13 binding of the sync-audit verdict to the current HEAD, and the ordered enable switch. Slot work, the ready view, metrics and the reuse relaxation moved to the follow-up section (no ids). Tier L re-judged (§6). 12 REQ / 12 AC. |

## 1. Problem

The landing path accepts a re-measure record when one command exited 0 on the candidate tree (`internal/factory/integration_remeasure.go:135` `ValidateRemeasureRecord`; for a command that is not `go test -json` it records command and exit code only, :185-193 and :698). Nothing records which calls were required, that each ran to the end, or that all of them passed on the same tree. Observed (`research.md` §3): landed tree `1a3dc230c35ca0e2a997fc0fc3ac9444eb1f9198` (card t1680) was gated by one script-wrapped record with `structured_required=false`; the other required call's result lived only in a progress note. The weakness is also visible in the repository's own stored records.

Separately, for a SPEC card the merge-readiness sync-audit condition reads only `sync_status: complete` from progress (`internal/factorylane/merge.go` `checkSyncAudit`), so an earlier PASS can stand for the audit of a later HEAD (`research.md` §6, §2). The operator decided (board `2c51`) that `complete` binds the verdict to the card's current HEAD.

This SPEC adds the missing evidence layer and turns it on in this repository, in a fixed order.

## 2. Terms

- **Candidate tree / HEAD** — the tree and commit of the card branch tip being landed.
- **Manifest** — the per-card list of required verification calls (REQ-LVM-001). **Call** — one manifest entry. **Attempt** — one execution of a call, with a unique attempt ID.
- **Package terminal event** — the `go test -json` event with `Action` pass, fail or skip, no `Test` field, for one `Package`.
- **Gate-complete on tree T** — every manifest call has a counted attempt on T whose package terminal events cover the call's whole package set, zero missing, none failed (REQ-LVM-004).
- **b2cd scope** — the standing landing scope: the Go packages the card changed (test-only changes included) plus the Go packages that import them, each run whole with no `-run`; the one recorded narrowing allowance is `internal/cli` (rule 35e9: `-run` over tests the card adds or changes and tests that reference identifiers it changes). These are machine-local board rulings today; this SPEC commits them to the tree (REQ-LVM-001, REQ-LVM-007).
- **Structured form** — one `moai integration remeasure -- go test … -json …` invocation per manifest call, as opposed to a script that runs several calls and prints no recognized package report.

## 3. Requirements (GEARS)

**REQ-LVM-001 (event).** When a card's candidate tree is prepared for landing and the completeness gate is enabled, the manifest generator shall write a manifest for that card listing, for every Go package the card changed (test-only and package-directory non-Go changes included) and for every direct importer of a changed package (non-test and test imports alike), one call that runs that package whole with `go test -count=1 -p 1 -json`; for every changed non-Go file outside a Go package directory it shall record the result of a reader search over the repository's test sources (the readers found, each as a call, or an explicit zero-readers record naming the search command and its empty result); and it shall record per call its identifier, command, and resolved package set. A call for `internal/cli` alone may carry a `-run` narrowing and shall name it as the 35e9 allowance.

**REQ-LVM-002 (event).** When a manifest call is executed through the structured form, the re-measure verb shall write an attempt record carrying a unique attempt ID, the call identifier, the command, the exit code, the per-package terminal event for every package in the stream, the candidate tree SHA, the card HEAD SHA, the base, an environment fingerprint (GOOS, GOARCH, CGO_ENABLED, GOFLAGS, toolchain version, build tags, declared environment variables), the start and end times, and a termination class.

**REQ-LVM-003 (event).** When an attempt ends by signal, by timeout, or with a stream that ends on a truncated line, the verb shall classify it as aborted and retain its record, and an aborted attempt shall never count as a pass; a later execution of the same call shall be a new attempt with a new attempt ID that does not overwrite the aborted record, and only the most recent completed attempt per call on the same tree shall count.

**REQ-LVM-004 (ubiquitous).** The completeness verifier shall report gate-complete only when every call of the card's manifest has a counted attempt on the candidate tree, each counted attempt exited 0 with a terminal event for every package of the call's resolved set (zero missing) and none failed, and no counted attempt swept an empty package set where the manifest does not allow it; the most recent record alone shall never decide the verdict.

**REQ-LVM-005 (event).** When any of the following holds, the verifier shall refuse readiness: the manifest is missing; a manifest call has no counted attempt; the call's only record is a script-wrapped command that carries no recognized package report; the stream lacks the terminal event of an expected package; the stream is cut mid-line; the attempt is aborted; the attempt names a different tree or HEAD than the candidate; the attempt's environment fingerprint differs from the manifest's declared fingerprint; a changed non-Go input has neither a reader call nor a zero-readers record; the b2cd scope cannot be derived (`go list` fails or a package set is unresolved); a counted attempt swept an empty set the manifest does not allow.

**REQ-LVM-006 (ubiquitous).** The completeness verifier shall be the single verifier consulted by the merge step, by `moai factory complete`, and by merge-readiness condition (d), and while the gate is enabled a record from a command that carries no recognized package report shall not satisfy any of the three.

**REQ-LVM-007 (state).** While `workflow.landing_gate.enabled` is false or absent, the landing path shall behave exactly as it does today, the key shall ship false in the distributed template, and the narrowing allowance list shall ship containing `internal/cli` only and shall not be widened by this SPEC.

**REQ-LVM-008 (event).** When `moai factory complete` evaluates a SPEC card while the gate is enabled, it shall refuse unless the card's sync-audit verdict names the card's current HEAD as its audited commit (differing from HEAD at most under the card's own evidence directory, the allowance the SPEC-less path already applies in `internal/factorylane/merge_specless.go`), and an audit verdict for any earlier commit shall not substitute for it, even when the change since then is empty or patch-identical.

**REQ-LVM-009 (event).** When `moai factory complete` evaluates a card while the gate is enabled, it shall refuse unless the required audit backend's receipt for the card's current HEAD exists and reads pass; a receipt from an earlier HEAD, an absent receipt, or an unreadable receipt shall refuse.

**REQ-LVM-010 (ubiquitous).** The landing rules shall state the structured form and the per-call procedure a lane follows (derive the manifest, run each call as its own invocation, restart an aborted call as a new attempt, check the verdict before `complete`), in the committed rule files and their template mirrors.

**REQ-LVM-011 (ubiquitous).** The configuration key that enables the gate in this repository shall be switched on in a commit that changes that one configuration path and nothing else, and that commit shall have the baseline commit, the gate commit, the audit-binding commit, and the guidance commit as ancestors, with the negative and positive gate tests passing on the tree it produces; the baseline commit shall precede every other behavioural commit and shall change only the baseline file.

**REQ-LVM-012 (ubiquitous).** The gate shall not skip or weaken any audit check, shall not accept an earlier PASS for a later HEAD, shall not exclude a test from a call, shall not write the integration branch except through the existing merge step, shall not widen any default permission or narrowing allowance, and shall treat every undecidable reading as refusal.

## 4. Acceptance index

AC-LVM-001 to AC-LVM-012 map one-to-one to REQ-LVM-001 to REQ-LVM-012 (`acceptance.md`, Given-When-Then with command, exit-code/stdout criterion and a RED-now ledger).

## 5. Constraints

- C1. Git-flow window path only; the github-flow pull-request edge (`internal/cli/factory_card.go:2071-2073`) is not touched.
- C2. Existing audit checks are consumed as they are; the SPEC adds readers and one binding (REQ-LVM-008/009), never a bypass.
- C3. The gate ships off in the distributed template; the in-repo switch is the last milestone (REQ-LVM-011).
- C4. No time estimates; milestones are ordered by dependency and by the switch rule.
- C5. Cards already merge-ready when the switch lands are refused until their calls are re-run in the structured form; this is the decided cost (decision-index Q7), mitigated by the guidance milestone preceding the switch.

## 6. Tier judgment (re-judged for piece A)

Tier L. Arithmetic, from measured analogues and the touch list:

- Nearest analogue: the re-measure record module, `internal/factory/integration_remeasure.go` 740 lines + `internal/cli/integration_remeasure.go` 131 + their tests 470 + 261 + 132 = 1,734 lines. The gate generalizes that module (manifest + attempts + verifier) and adds the audit binding beside `internal/factorylane/merge.go` (471) / `merge_specless.go` (154): estimated production ≈ 1,000-1,300 lines and tests ≈ 1,400-1,800 lines, i.e. well above the Tier M ceiling of 1,000 lines.
- Files: new production files ≈ 4 and new test files ≈ 6 in `internal/factory`/`internal/factorylane`; edited production ≈ 9 (`integration_remeasure.go`, `integration_merge_step.go`, `factory_card.go`, `factory_merge.go`, `cli/integration_remeasure.go`, `factorylane/merge.go`, `factorylane/merge_specless.go`, config types/defaults/loader); config ≈ 3 (template default, repo config, defaults); rules ≈ 4 (two rule files and their mirrors); CLI/config tests ≈ 3. About 29 files, above the Tier M ceiling of 15.
- Requirement and criterion counts (12 and 12) sit under the Tier L ceiling of 25 and would also fit Tier M's 16; the tier follows LOC and files, as the table states, and the artifact set is the Tier L five files plus the decision index.

## 7. Follow-up pieces (not in this SPEC)

Recorded per board `9dbe`; each is issued as its own card only after piece A lands. None is implemented, specified, or acceptance-checked here.

- A slot piece: removal of the lease record of a dead owning session, and nothing that takes over an active lease; capacity stays 1 and the wall-clock package list stays empty until measured.
- A readiness-and-metering piece: a derived ready view, landing-event metrics, and a shadow counter of how often a re-measure reuse would have been admissible; it depends on piece A.
- The re-measure reuse relaxation is refused for now and not built; no allowed non-Go glob list and no hermetic-package list is to be guessed.

## 8. Exclusions

### Out of Scope — landing paths and policy

- The github-flow pull-request delivery edge, CodeRabbit gating, and CI-side verification.
- Skipping or shortening the HEAD confirmation audit for any reason, patch-id equality included.
- Train stacking, a second writer of the integration branch, queue admission, lane leasing and card classification.

### Out of Scope — verification content

- Choosing which tests a card should add; the manifest records what the b2cd rule requires.
- Forwarding the child stdout of `moai integration remeasure` (the known swallow defect); a script-wrapped call simply does not satisfy the gate while it is enabled.
- Reusing a record across a moved integration tip, and any base-equals-tip relaxation.

### Out of Scope — slot subsystem and views

- Slot capacity, load and memory headroom, wall-clock-bound package exclusion, a slot wrapper verb, and any change to lease liveness semantics.
- A ready-to-land view and landing metrics.
