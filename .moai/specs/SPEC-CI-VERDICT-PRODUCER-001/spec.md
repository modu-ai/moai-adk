---
id: SPEC-CI-VERDICT-PRODUCER-001
title: "CI verdict producer: a moai CLI verb that records remote CI conclusions per head SHA as on-disk evidence the escalation detector's contradictory-evidence CI limb consumes, making AC-AE-012(c) of SPEC-AUTONOMY-ESCALATION-001 re-judgeable as written"
version: "0.1.0"
status: in-progress
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/escalation, internal/cli"
lifecycle: spec-anchored
tags: "ci,verdict,escalation,contradictory-evidence,detector,producer"
tier: M
related_specs: [SPEC-AUTONOMY-ESCALATION-001]
---

# SPEC-CI-VERDICT-PRODUCER-001 — CI verdict producer and detector CI limb

## §A — History

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-26 | Initial plan-phase draft (card t1268, Tier M). |

## §B — Problem

SPEC-AUTONOMY-ESCALATION-001 REQ-AE-010 lists three disagreeing-verdict shapes that trip class
`contradictory-evidence`. The third — "a recorded CI failure for a head whose local verification
was recorded as passing" — is dead today: **no producer anywhere records CI verdicts**, and the
detector hardcodes the limb as never-observed:

- `internal/escalation/operational.go:25-28` — `notObservedCIVerdict = "ci verdict (no recorded CI verdict producer)"`, an unconditional constant.
- `internal/escalation/operational.go:325` — `classContradictoryEvidence()` seeds every record's and every checkpoint's `not_observed` list with that constant, so AC-AE-012(c) cannot be judged PASS or FAIL; it is structurally UNVERIFIED.

Debt provenance: card t1235's verdict (`/Users/goos/MoAI/moai-adk-go/.moai/reports/t1235/verdict.md`, "Debt — no CI verdict producer") recorded AC-AE-012(c) UNVERIFIED and issued card t1268. The lead ruled against amending SPEC-AUTONOMY-ESCALATION-001; **this card makes the criterion re-judgeable AS WRITTEN** — it builds the producer the criterion's wording presupposes, and (in the same card, because the producer alone still leaves the limb inert) implements the detector's CI limb against that producer's records.

## §C — Requirements (GEARS)

### §C.1 Producer

- **REQ-CV-001** (Event-driven) — When the CI verdict producer verb is invoked for a head SHA in a project tree, the producer shall record one CI verdict record for that head at `.moai/state/ci-verdicts/<head-sha>.json` under the tree it runs in, carrying the CI conclusion observed for that head.
- **REQ-CV-002** (Event-driven) — When the producer is invoked with an offline input file naming the head, conclusion, and run identifier, it shall record the verdict from that file without invoking any network client; the recorded bytes shall be indistinguishable in schema from a fetched verdict.
- **REQ-CV-003** (Event-detected) — When the network client (`gh`) is absent, unauthenticated, or its query fails, the producer shall print a clear one-line message naming the fault and exit 0 without writing any record (fail-open house discipline, per `internal/cli/doctor.go` `gh` handling precedent); a failed observation never writes a fabricated verdict.
- **REQ-CV-004** (Ubiquitous) — The producer shall write records whose schema is: `head_sha` (the full SHA the verdict judges — the record's filename and its pinned head are the same value), `conclusion` (one of `success`, `failure`, `neutral`), `run_id` (the CI run identifier, empty when the backend reports none), `observed_at` (RFC 3339 timestamp of the observation), and `producer` (the writing verb's identity); writes shall be atomic (temp-file-then-rename), and re-recording the same head with the same conclusion is an idempotent byte-equivalent rewrite (last-writer-wins on differing content).
- **REQ-CV-005** (Unwanted) — The producer shall not invoke the escalation detector, any checkpoint, or any hook path; it writes evidence files and nothing else. The question "what invokes `escalation.Checkpoint` in production" (t1235 design Q2) is **separately unresolved and stays decoupled** from this SPEC.

### §C.2 Detector CI limb

- **REQ-CV-006** (Ubiquitous) — The local verification pass evidence source for the CI limb shall be the verify snapshot store (`.moai/state/verify/snapshots/`, keyed by HEAD SHA per `internal/verify/store.go`): a head has a recorded local pass when a snapshot whose key's head portion equals that head contains at least one check entry recorded with exit code 0. No second local-pass source is invented.
- **REQ-CV-007** (Event-driven) — When a commit checkpoint runs and a CI verdict record exists whose `head_sha` equals the checkpoint's head, whose `conclusion` is `failure`, and whose head also has a recorded local verification pass per REQ-CV-006, the detector shall trip class `contradictory-evidence`, writing exactly one record via the existing `contradiction(...)` path (fingerprint `(class, source, pair)`; the `freshEvidence` content-hash gate governs re-tripping per REQ-AE-020/AC-AE-023 of SPEC-AUTONOMY-ESCALATION-001).
- **REQ-CV-008** (Event-detected) — When a checkpoint runs with a CI verdict record whose `conclusion` is `success` or `neutral`, whose `head_sha` differs from the checkpoint head, or when no CI verdict record exists for the checkpoint head, the detector shall write no record and shall label every limb it could not complete under `not_observed` per REQ-AE-022 of SPEC-AUTONOMY-ESCALATION-001 — a CI failure recorded for a different head, and a missing local pass at the checkpoint head, are each named in `not_observed`; no CI record at all yields exactly the limb-(e) behavior already specified (listed as not-observed, never as agreement). A `success` or `neutral` conclusion observed at the checkpoint head completes the CI observation and is NOT listed under `not_observed` — though a missing local pass at that head is still listed, since the limb's other half remains unobserved. (`neutral` is a completed observation, same as `success`: it is an observed verdict that does not contradict the local pass, and listing it as not-observed would keep the limb permanently flagged on every skipped or cancelled CI run — noise without any disagreement signal.)
- **REQ-CV-009** (Event-driven) — When the same CI-record content has already tripped the limb's fingerprint, a later checkpoint shall write no duplicate record; the `freshEvidence` content-hash gate treats identical bytes as consumed evidence, so an idempotent producer re-run (REQ-CV-004) cannot re-trip a resolved record without new evidence.

## §D — Success Criteria

All acceptance criteria in `acceptance.md` PASS; specifically: the extended `TestContradictoryEvidenceTrips` limb (c) exercises a real same-head trip (this is the mechanical form of AC-AE-012(c) re-judgement — the extended test IS the re-judgement run); `go.mod` carries no new dependency; `GOOS=windows GOARCH=amd64 go build ./...` passes.

## §E — Out of Scope

### Out of Scope — Detector invocation wiring (t1235 Q2)

- Nothing in this SPEC decides what invokes `escalation.Checkpoint` in production. The producer writes evidence only; that unresolved question remains t1235's debt and is explicitly decoupled.

### Out of Scope — t1235 audit findings

- t1235's own audit findings (e.g. F3, records agent-writable) are that card's debts. This SPEC neither absorbs nor re-opens them.

### Out of Scope — SPEC-AUTONOMY-ESCALATION-001 amendment

- That SPEC is a read-only reference. No requirement, AC, or frontmatter of it is modified; the criterion AC-AE-012(c) becomes judgeable through new evidence, not through rewording.

### Out of Scope — Cross-tree record mirroring and new machinery

- Records are consumed by a detector running in the same tree (`r.root`) the producer wrote them in. No cross-tree sync, no mirroring daemon, no new gate, and no new external dependency (no `go.mod` change) — the producer shells out to the already-house-standard `gh` CLI exactly as `internal/cli/doctor.go`, `internal/cli/todo_pr.go`, and `internal/cli/branch_protection.go` already do.

## §F — Coordination Premises

| Premise | Disposition |
|---------|-------------|
| t1235 design Q2 (nothing calls `escalation.Checkpoint` in production) | Unresolved elsewhere; REQ-CV-005 pins the decoupling. |
| t1235 audit findings (F3 records agent-writable, etc.) | Not this card's debt; not absorbed. |
| Lanes never push in this repo's git-flow (lead batch-pushes `origin/develop`) | Grounds the producer's natural caller: the lead, after its batch push (§G Q1). |

## §G — Design Decisions (the four settled questions)

**Q1 — Who reads CI results (producer invocation surface).** CHOSEN: a moai CLI verb whose natural caller is the **lead session after its batch push** — the only actor in this repo's git-flow that pushes, and therefore the only actor that can observe remote CI for a pushed head. REJECTED: a hook (hooks fire per tool call inside lane worktrees; lanes cannot push, so no hook ever has a fresh remote verdict); integration into `escalation.Checkpoint` (would absorb t1235 Q2 — forbidden by REQ-CV-005); manager-git PR machinery (the release-path surface, not the develop CI verdict surface).

**Q2 — Where records go.** CHOSEN: `.moai/state/ci-verdicts/<head-sha>.json` under the tree the producer runs in — one JSON file per judged head, following the class-5 precedent (`.moai/state/audit-multi/*.json`) and the gitignored `.moai/state/` evidence namespace. Primary-vs-worktree visibility is answered by the same-head condition: records land where CI is judged (the lead's tree at the pushed head); a detector in any tree consumes a record only when that record's pinned head equals the detector's own checkpoint head (REQ-CV-007/008), so a stale worktree simply fails the head match. No cross-tree mirroring (rejected — new machinery without a demonstrated consumer).

**Q3 — What format.** CHOSEN: the five-field JSON schema of REQ-CV-004, mirroring how `internal/escalation/operational.go` already parses sibling evidence files (`convergenceFile` — plain `encoding/json` struct with json tags, glob + `os.ReadFile` + `json.Unmarshal`, unreadable files listed not-observed and skipped). REJECTED: YAML/Markdown records (no sibling precedent in the class-5 evidence directory); a richer CI-API-shaped dump (drags `gh`'s schema into the detector's parse path).

**Q4 — Freshness / same-head semantics.** CHOSEN: the record pins the head it judges; the detector consumes it only on a head match; trip requires local pass + CI failure at the SAME head. No-trip cases: head mismatch and a missing local pass are each listed under `not_observed` (REQ-CV-008); a `success` or `neutral` conclusion at the head completes the observation and is NOT listed (REQ-CV-008's rationale — an observed non-contradicting verdict is not an unobserved limb); no CI record at all → the limb-(e) status quo is preserved byte-for-byte. Re-tripping is governed by the existing `freshEvidence` content-hash gate (REQ-CV-009), which is what makes the idempotent producer re-run safe.

## §H — Cross-References

- `depends_on` is deliberately omitted: SPEC-AUTONOMY-ESCALATION-001 reads `status: implemented`, and the strict dependency-fulfillment rule (fulfilled only at `status: completed`) would hard-block run-phase on it; the relation is carried non-blocking via the `related_specs` frontmatter field instead.
- `acceptance.md` — AC-CV-001..008 Given-When-Then scenarios and traceability.
- `plan.md` — milestones, file targets, pre-flight; `research.md` — anchor citations and rejected alternatives.
- SPEC-AUTONOMY-ESCALATION-001 — REQ-AE-010 (the trip list this SPEC completes), REQ-AE-022 (not-observed labeling), REQ-AE-020/AC-AE-023 (freshEvidence gate), AC-AE-012(c) (the criterion made re-judgeable).
