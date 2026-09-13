---
id: SPEC-GATEWAY-ENVELOPE-REPAIR-001
title: "Reasoning-envelope repair — acceptance criteria"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t708)
tier: M
---

# Acceptance — SPEC-GATEWAY-ENVELOPE-REPAIR-001

## §A Verification philosophy

Every AC is binary-testable. The controlling principle: the repair is verified by the UNCHANGED validator — no AC may be satisfied by repair-side logic that reimplements, previews, or relaxes the binding. Characterization ACs (AC-EVR-001..003) are GREEN at arrival; behavioral ACs (AC-EVR-004..010) are TDD RED→GREEN.

## §D AC Matrix

### Validator lock (characterization)

- **AC-EVR-001** — Validator accept-set invariance:
  - **Given** the t672/t703 characterization suites at the absorbed develop base, **When** the card's full diff is applied and the gateway translate + receipt suites run, **Then** every pre-existing pass/fail outcome is identical and the card's new characterization tests prove: an envelope-bearing replay is accepted, and a stripped-envelope replay with surviving markers is rejected with `CauseReasoning`.
  - Evidence: `go test ./internal/gateway/translate/ ./internal/gateway/receipt/ -count=1` (verbatim output) + the new test names.

- **AC-EVR-002** — Error-contract byte-identity:
  - **Given** the frozen `historyReplayGuidance` sentence and the t672/t703 reason clauses, **When** any rejection in the characterization matrix fires, **Then** the 400 body strings match golden byte-identical (no new clause, no rewording).
  - Evidence: golden-string assertions in the characterization tests, green.

- **AC-EVR-003** — Refusal shapes stay rejected:
  - **Given** marker-missing, digest-mismatch, public-content-changed, and source-gone histories, **When** each is replayed, **Then** each is rejected by the unchanged Check with its existing classification, unchanged by this card.
  - Evidence: the characterization matrix cells, green.

### Repair path (TDD)

- **AC-EVR-004** — Byte-exact verbatim injection:
  - **Given** a stripped-envelope history whose family transcript retains the issued carriers and whose boundaries carry self-attesting markers, **When** the repair runs, **Then** every injected carrier is byte-identical to the transcript-recorded carrier (`sha256(raw) == marker.opaque_sha256` asserted per boundary) and no other byte of the replayed history changes.
  - Evidence: TDD test asserting per-boundary digest equality + full-history byte-diff confined to injected `redacted_thinking` blocks.

- **AC-EVR-005** — Position exactness:
  - **Given** the same history, **When** the repair runs, **Then** injection occurs only at boundaries with a self-attesting surviving marker and only at the original positions; a boundary whose marker was stripped is left untouched and aborts the repair (refusal, AC-EVR-007 path).
  - Evidence: TDD test with a marker-stripped fixture — repair refuses, zero modification.

- **AC-EVR-006** — Explicit invocation only:
  - **Given** a classified `CauseReasoning` 400, **When** no user invocation occurs, **Then** no history, payload, transcript, or record is modified (automatic-surgery test: rejection handling produces a byte-identical state), and **When** the user invokes the repair, the repaired replay proceeds to the unchanged Check.
  - Evidence: TDD tests, both branches.

- **AC-EVR-007** — Refusal on non-repairable preconditions:
  - **Given** each of: source-gone (no verbatim carrier in transcript), digest mismatch between carrier and marker, missing marker, Prefix-class public-content mismatch, **When** the repair is invoked, **Then** each case refuses with zero modification, preserves non-destructive state, and surfaces the classified guidance unchanged.
  - Evidence: four refusal TDD cells, green.

- **AC-EVR-008** — Single-shot with durable termination:
  - **Given** a repair attempt whose retry is still rejected, **When** the same conversation is presented again (including from a fresh process reading the durable record), **Then** the repair path performs no further injection and surfaces the guidance unchanged.
  - Evidence: TDD test simulating a fresh process against the recorded attempt marker.

- **AC-EVR-009** — Non-destructive aside + provenance:
  - **Given** any completed or refused repair attempt, **When** the operation ends, **Then** the unmodified history is recoverable from the preserved aside and the durable record carries the per-boundary digest + position provenance.
  - Evidence: TDD test reading back the aside and the record.

- **AC-EVR-010** — No receipt-store access from the repair path:
  - **Given** the repair implementation, **When** statically inspected (declaration/import grep + review), **Then** the repair path imports and calls no symbol from `internal/gateway/receipt` store/manifest read surfaces, and no repair trigger reads request metadata.
  - Evidence: scripted grep recorded in the run evidence, zero hits.

### Non-invasiveness + docs

- **AC-EVR-011** — PRESERVE zero-diff (file-level):
  - **Given** the card's final diff (merge-base develop..HEAD), **When** diffed, **Then** zero changed lines in the §D PRESERVE file list of plan.md (validator, issuance/emission, codec, fork, factory).
  - Evidence: scripted diff-scope assertion output (empty for the listed paths).

- **AC-EVR-012** — Operator documentation exists:
  - **Given** `.moai/docs/gateway-envelope-repair.md`, **When** read, **Then** it documents the shape, procedure, bounds (explicit invocation, single-shot, byte-exact, non-destructive), refusal conditions, non-repairable shapes, and the fork path for lineage misses.
  - Evidence: file exists + section checklist.

## §D.1 Severity and traceability

| REQ | ACs | Severity |
|---|---|---|
| REQ-EVR-001 | AC-EVR-001, AC-EVR-002, AC-EVR-003, AC-EVR-011 | MUST |
| REQ-EVR-002 | AC-EVR-004 | MUST |
| REQ-EVR-003 | AC-EVR-004, AC-EVR-005, AC-EVR-010 | MUST |
| REQ-EVR-004 | AC-EVR-006 | MUST |
| REQ-EVR-005 | AC-EVR-006 | MUST |
| REQ-EVR-006 | AC-EVR-008 | MUST |
| REQ-EVR-007 | AC-EVR-005, AC-EVR-007 | MUST |
| REQ-EVR-008 | AC-EVR-009 | MUST |
| REQ-EVR-009 | AC-EVR-010, AC-EVR-011 | MUST |
| REQ-EVR-010 | AC-EVR-012 | SHOULD |

Every REQ has ≥1 AC; every AC traces to ≥1 REQ. No orphan rows.

## §D.2 Edge cases (exercised in the matrix)

- Subagent-transcript boundaries (carriers in subagent rows — research addendum) — repair scope is the launcher conversation's own history; subagent rows are not injected.
- Reverse switch (luna→sol): symmetric to t703 task-4(a); same mechanism, same handling.
- Multi-boundary repair: several stripped boundaries in one history — all-or-nothing per attempt (a partial injection that cannot complete refuses).
- Carrier present but malformed (base64 corruption in transcript): fails digest self-attestation → refusal.

## §D.3 Quality gates

- Affected packages: `go test ./internal/gateway/translate/ ./internal/gateway/receipt/ ./internal/gateway/conversation/ ./internal/cli/ -count=1` green; coverage ≥85% on touched packages (E3).
- `go vet` + `golangci-lint run --new-from-rev=HEAD`: 0 new issues (E5).
- Cross-platform build exit 0 (E2).
- No local full-suite runs; CI owns the verdict.

## §D.4 Definition of Done

1. All AC-EVR-### PASS with verbatim evidence (E1).
2. §D PRESERVE list byte-identical (AC-EVR-011).
3. plan.md §H Resolution Record dispositions in effect (recorded per the lead conditional-Kickoff directive, 2026-09-13 — these ARE the §D.4-3 dispositions); any NEW bounded question surfacing in run phase is dispositioned before M3 exit.
4. M0 seam adjudication recorded.
5. Documentation deliverable landed (AC-EVR-012).
6. Integration per gitflow lane protocol; sync closed before the develop merge.

## §D.5 Forward-looking checks

- The characterization matrix from M2 remains the regression guard for any future validator change — it must keep passing untouched.
- If t700's SPEC lands with wording that resolves spec.md §4 under Reading A, this SPEC's M3+ is blocked pending re-delegation; M2's matrix remains valid either way.
