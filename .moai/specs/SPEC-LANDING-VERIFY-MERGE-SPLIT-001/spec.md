---
id: SPEC-LANDING-VERIFY-MERGE-SPLIT-001
title: "Split landing verification from merge: re-measure completeness gate, ready-to-land view, heavy-slot capacity and dead-owner release, landing metrics, and gated re-measure reuse"
version: "0.1.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/factory, internal/cli, internal/config, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "landing, remeasure, completeness-gate, reuse, ready-to-land, slot-lease, capacity, instrumentation, card-t1872, git-flow"
tier: L
related_specs: [SPEC-MERGE-WINDOW-QUEUE-001, SPEC-MERGE-WINDOW-QUEUE-002, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-HEAVY-TEST-SLOT-001, SPEC-VERIFY-RUN-REUSE-001]
---

# SPEC-LANDING-VERIFY-MERGE-SPLIT-001

Card t1872 (Class C, P0). Operator directive 10-18: uninterrupted lanes, and verification split from merge. Baseline and code-level evidence: `research.md`. Design, threat analysis, and the split recommendation: `design.md`. Open product-level questions: `decision-index.md`.

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Plan-phase draft. Six-part scope per the card body (the leader added part (6), the completeness gate, as the FIRST milestone, 10-18). Tier L; REQ 22 / AC 22. |

## 1. Problem

Under the git-flow window path (`git-strategy.yaml` `workflow: git-flow`; `research.md` §1), every landing of one card moves the integration tip, and every other waiting `merge-ready` card must then re-absorb, re-run its HEAD confirmation audit, and re-run its re-measure before `moai factory complete` accepts it. Measured at authoring (`research.md` §5): 16 cards in `merge-ready`; landing intervals median 34.0 min, p90 158.5 min over the last 40 landings. A single heavy-slot holder serializes the re-measures.

A premise discovered while scoping the relief is the more important finding. The re-measure record the gate consumes proves little: for a script-wrapped command it records command and exit code only (`research.md` §3). Landing t1680's tree `1a3dc230…` was gated by one `gate-b.sh` record with `structured_required=false`; which calls were required, and that all of them passed on that tree, is nowhere mechanically recorded. Relaxing the base-equals-tip rule on top of such a record would let a record that was never complete travel across trees. The completeness gate (part 6) therefore comes first, and the reuse relaxation (part 1) is the last thing to open.

## 2. Terms

- **Candidate tree** — the tree of the card branch tip being landed. **Integration tip** — the tip of the configured integration branch.
- **Manifest** — the per-card list of required verification calls (§3 REQ-LVM-001). **Call** — one manifest entry. **Attempt** — one execution of a call; it has a unique attempt ID.
- **Package terminal event** — the `go test -json` event with `Action` of `pass`, `fail` or `skip` and no `Test` field, for one `Package`.
- **Gate-complete on tree T** — every manifest call has a completed, exit-0 attempt on T whose package terminal events cover the call's whole package set with zero missing and none failed (REQ-LVM-004).
- **Scope derivation (b2cd)** — the standing landing scope: the Go packages the card changed plus the Go packages that import them, each run whole with no `-run`; the single recorded narrowing allowance is `internal/cli` (rule 35e9: `-run` over tests the card adds or changes and tests that reference identifiers it changes). These rules are machine-local decision-board rulings today; REQ-LVM-001 and REQ-LVM-007 commit them to the tree.
- **Moved set** — the commits and files between a record's base and the current integration tip.

## 3. Requirements (GEARS)

### Part 6 — re-measure completeness gate (first milestone)

**REQ-LVM-001 (event).** When a card's candidate tree is prepared for landing and the completeness gate is enabled, the manifest generator shall write a manifest for that card listing, for every Go package the card changed (test-only changes included) and for every direct importer of a changed package (importers counted by non-test and test imports alike), one call that runs that package whole with `go test -count=1 -p 1 -json`, and shall record for each call its identifier, command, and resolved package set; a call for `internal/cli` alone may carry a `-run` narrowing, and shall name it as the 35e9 allowance.

**REQ-LVM-002 (event).** When a manifest call is executed through the re-measure verb, the verb shall write an attempt record carrying a unique attempt ID, the manifest call identifier, the command, the exit code, the per-package terminal event (pass, fail or skip) for every package in the stream, the candidate tree SHA, the card HEAD SHA, the base, an environment fingerprint (GOOS, GOARCH, CGO_ENABLED, GOFLAGS, toolchain version, build tags, and the declared environment variables), the start and end times, and a termination class.

**REQ-LVM-003 (event).** When an attempt ends by signal, by timeout, or with a stream that lacks a terminal event for any expected package or ends on a truncated line, the verb shall classify the attempt as aborted, shall retain its record, and shall never count it as a pass; a later execution of the same call shall be a new attempt with a new attempt ID, and only the most recent completed attempt per call on the same tree shall count.

**REQ-LVM-004 (ubiquitous).** The completeness verifier shall report gate-complete only when every call of the card's manifest has a counted attempt on the same candidate tree, each attempt's terminal events cover the call's entire package set with zero missing packages and no failure, and no counted attempt swept an empty set where the manifest does not allow it; the last record alone shall never decide the verdict.

**REQ-LVM-005 (event).** When any of the following holds, the verifier shall refuse readiness and refuse reuse: the manifest is missing; a manifest call has no counted attempt; an attempt's stream is partial; an attempt is aborted; an attempt names a different tree or HEAD than the candidate; the attempt's environment fingerprint differs from the manifest's declared fingerprint; a non-Go input the card changed has no verified reader call; the overlap decision is undecidable.

**REQ-LVM-006 (ubiquitous).** The completeness verifier shall be the single verifier consulted by the merge step, by `moai factory complete`, by merge-readiness condition (d), and by the ready-to-land view, and where the gate is enabled a record from a command that carries no recognized package report shall not satisfy it.

**REQ-LVM-007 (state).** While `workflow.landing_gate.manifest_required` is false or absent, the landing path shall behave exactly as it does today, and the key shall ship false; the narrowing allowance list shall ship containing `internal/cli` only and shall not be widened by this SPEC.

### Ordering

**REQ-LVM-008 (ubiquitous).** The re-measure reuse relaxation (REQ-LVM-016 to REQ-LVM-020) shall be inert unless the completeness gate is enabled and the source record is gate-complete, and the run-phase commits shall land in the commit-graph order M1 (gate) before every other behavioural milestone and the reuse milestone last.

### Part 5 — instrumentation

**REQ-LVM-009 (event).** When a card lands, becomes ready, has a re-measure executed, or is considered for reuse, the landing path shall append one event row to a landing-event log, and the metrics verb shall report the landing interval distribution, the ready-wait time per card, the re-measure execution count, the reuse count, and the shadow reuse-eligibility count; a failure to write the log shall never block or alter a landing.

### Part 4 — dead-owner release

**REQ-LVM-010 (event).** When a slot verb or the slot guard reads a lease whose recorded owner process is gone, the slot subsystem shall remove the record, shall append an `auto-release` audit row naming the removed holder, and shall treat an owner that cannot be resolved (pid 0 or negative) as live.

### Part 3 — heavy-slot capacity

**REQ-LVM-011 (state).** While a resource's configured capacity is 1, or no capacity is configured, the slot subsystem shall admit at most one holder exactly as today; where the capacity is N greater than 1, it shall admit up to N distinct live holders and shall refuse the next with the held exit code.

**REQ-LVM-012 (event).** When an acquire would become the second or later concurrent holder, the slot subsystem shall admit it only if the one-minute load average and the available memory, read at that moment, satisfy the configured thresholds, and shall refuse it when a reading is unavailable, unparseable, or the thresholds are unset.

**REQ-LVM-013 (event).** When an acquire declares a package set that intersects the configured wall-clock-bound package list while another live holder has declared an intersecting set, the slot subsystem shall refuse it with the held exit code regardless of free capacity.

### Part 2 — ready-to-land view

**REQ-LVM-014 (event).** When the operator or leader requests the ready set, the ready verb shall evaluate every `merge-ready` card read-only and shall list a card as ready only if the card branch contains the current integration tip (or the reuse decision of REQ-LVM-016 holds for it), the gate is complete on the candidate tree, a valid re-measure record exists for that tree, the audit evidence names the card's current HEAD, and the merge dry-run is conflict-free; every card not listed shall carry the first failing predicate.

**REQ-LVM-015 (ubiquitous).** The ready view shall be derived at read time and shall persist no readiness marker that any landing verb reads; `moai factory complete` shall re-evaluate every gate itself, and an audit verdict whose audited SHA is not the card's current HEAD shall make the card not ready, even where the patch content is identical.

### Part 1 — re-measure reuse (opens last)

**REQ-LVM-016 (event).** When the integration tip has moved past a gate-complete record's base and the card has been re-absorbed by a conflict-free merge, the reuse verb shall decide whether the moved set overlaps the verification scope, and where it does not, shall write a reuse record keyed by the new candidate tree that cites the source record, the source base, the new base, the moved-commit list, the scope inventory digest, the decision, and the reasons.

**REQ-LVM-017 (ubiquitous).** The verification scope shall be the union of: every directory (recursive) of every module-local package in the `go list -deps -test` closure of the packages the source record observed, computed on the new candidate tree and on the source tree; every file the card changed; the module files go.mod, go.sum, go.work and go.work.sum; and the non-Go inputs of REQ-LVM-018.

**REQ-LVM-018 (event).** When the decision cannot be established — the source record carries no observed package list, its command is not a recognized `go test -json` run, the environment fingerprint or build tags differ, the closure cannot be listed, a measured package's non-Go reads are undeclared, a measured package's tests depend on live git state, or the source is itself a reuse record — the reuse verb shall decide overlap; a moved non-Go file outside the closure directories shall count as overlap unless the file matches a declared never-read glob configured by the operator.

**REQ-LVM-019 (event).** When the merge step, `moai factory complete`, or merge-readiness condition (d) meets a reuse record, it shall recompute the moved-commit list and the overlap decision from the repository and shall refuse with the existing re-measure-and-re-acquire code when its result differs from the record or reads overlap; an overlapping moved set shall leave today's refusal behaviour unchanged.

**REQ-LVM-020 (ubiquitous).** A reuse record shall cite a measured, gate-complete record as its source and never another reuse record, and the moved set shall be computed cumulatively from that source's base.

### Safety and documentation

**REQ-LVM-021 (ubiquitous).** The SPEC's mechanisms shall not skip or weaken any audit gate, shall not allow two cards to merge in one window hold, shall not write the integration branch except through the existing merge step, and shall treat every undecidable reading as the refusing outcome.

**REQ-LVM-022 (ubiquitous).** The landing rules shall describe the manifest, the gate, the ready view, the capacity keys, the auto-release, the metrics, and the reuse conditions, in the committed rule files and their template mirrors, edited together after the code milestones.

## 4. Acceptance index (Given-When-Then in `acceptance.md`)

AC-LVM-001 to AC-LVM-022 map one-to-one to REQ-LVM-001 to REQ-LVM-022. Each AC carries a command, an exit-code/stdout criterion, and a RED-now cell.

## 5. Constraints

- C1. Git-flow window path only; the github-flow pull-request edge (`factory_card.go:2071-2073`) is not touched.
- C2. The audit gates (sync-audit verdict, `audited_sha` bind, HEAD confirmation audit) are consumed unchanged; this SPEC adds a reader of them (REQ-LVM-015), never a bypass.
- C3. Every new key ships off or neutral (gate false, capacity 1, reuse closed); enabling any of them in this repository's own config is an operator act (decision-index Q7, Q9).
- C4. No time estimates; milestones are ordered by dependency and by the gating rule of REQ-LVM-008.

## 6. Exclusions

### Out of Scope — landing paths and policy

- The github-flow pull-request delivery edge, CodeRabbit gating, and CI-side verification.
- Skipping or shortening the HEAD confirmation audit for any reason, patch-id equality included.
- Train stacking: merging more than one card per window hold, or batching merges.
- Any second writer of the integration branch; `moai integration merge` stays the single writer.
- Changing queue admission, lane leasing, or card classification.

### Out of Scope — verification content

- Deciding which tests a card should add; the manifest records what the b2cd rule requires, not whether the tests are good.
- Forwarding the child stdout of `moai integration remeasure` (the known swallow defect); script-wrapped calls simply do not satisfy the gate while it is enabled.
- A file-access tracer for measured test processes (considered; not specified — `design.md` §5).

### Out of Scope — slot subsystem

- A slot wrapper verb that runs a command under a lease (`moai slot run`), and binding a lease to a command process rather than the owning session (decision-index Q10).
- The lane registry capacity (`factory_slots.go`), which is a different object from the resource slot lease.
- Changing the integration window's own dead-holder handling, which already clears dead owners.
