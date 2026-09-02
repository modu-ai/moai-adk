---
id: SPEC-INBOX-DRAIN-GAP-001
title: "Acceptance criteria — distributed lessons-inbox lifecycle"
version: "0.1.0"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
tier: M
---

# SPEC-INBOX-DRAIN-GAP-001 — Acceptance Criteria

All ACs are machine-verifiable. Verification vehicles: Go tests in `internal/hook` (t.TempDir-isolated) and CLI tests against fixture projects. No AC relies on a maintainer-machine live state.

## §A. Acceptance Scenarios (Given-When-Then)

### AC-IBX-001 — Write-time cap rotates an over-cap inbox (REQ-IBX-001, REQ-IBX-003)

**Given** a fixture project with no `.moai/state/lsel/` directory and a `lessons-inbox.jsonl` seeded over the cap constant, **When** the collector appends one stub, **Then** a rotated archive generation exists containing the pre-append content, the live file is fresh, and the live file size stays under cap + one-stub margin.

- Vehicle: Go test, `internal/hook` (t.TempDir).
- PASS = rotation observed on disk + size assertion holds + append landed.

### AC-IBX-002 — Curator stand-down (REQ-IBX-002; NFC-4 parity proof)

**Given** a fixture project with `.moai/state/lsel/` present and a `lessons-inbox.jsonl` seeded over the cap, **When** the collector appends stubs repeatedly, **Then** no archive generation is ever created and the live file grows unbounded (byte-identical append behavior to pre-SPEC).

- Vehicle: Go test, `internal/hook`.
- PASS = zero archives after over-cap appends + every append landed.

### AC-IBX-003 — Archive retention bound (REQ-IBX-004)

**Given** a fixture project without the marker where rotation fires more than twice, **When** the third and later rotations execute, **Then** at most 2 archive generations exist on disk after each rotation.

- Vehicle: Go test, `internal/hook`.
- PASS = archive count ≤ 2 after ≥ 4 forced rotations.

### AC-IBX-004 — CLI status (REQ-IBX-005)

**Given** a fixture project, **When** `moai inbox status` runs, **Then** it exits 0 and its output contains the live size, line count, cap distance, archive-generation count, and the ownership-regime token (`curator` or `cap-managed`).

- Vehicle: Go CLI test against a fixture project root.
- PASS = exit 0 + all five fields matched in output.

### AC-IBX-005 — CLI manual drain, standard install (REQ-IBX-006)

**Given** a fixture project without the marker and a non-empty inbox, **When** `moai inbox drain` runs, **Then** it exits 0, the live file is rotated (archive exists), and rotation statistics are printed.

- Vehicle: Go CLI test.
- PASS = exit 0 + on-disk rotation + stats present.

### AC-IBX-006 — CLI drain refusal on curator machines (REQ-IBX-007)

**Given** a fixture project with `.moai/state/lsel/` present and a non-empty inbox, **When** `moai inbox drain` runs, **Then** it exits non-zero, its notice names the wrapper-mediated drain path (`session_drain.sh`), and the inbox bytes are unchanged.

- Vehicle: Go CLI test.
- PASS = exit ≠ 0 + notice substring + byte-identical inbox (hash before == hash after).

### AC-IBX-007 — Schema version field (REQ-IBX-008)

**Given** the collector appends a stub, **When** the line is parsed as JSON, **Then** it carries an integer version field equal to 1; and given a pre-upgrade line without the field, **When** parsed by the SPEC's reader, **Then** it resolves as version 1.

- Vehicle: Go golden-JSON test + reader unit test.
- PASS = golden line has `"v":1` (or equivalent field name fixed in M1) + absence-tolerant parse.

### AC-IBX-008 — Fail-open on rotation failure (REQ-IBX-009)

**Given** a fixture where rotation is sabotaged (e.g., archive destination made undeletable/unrenamable), **When** an over-cap append executes, **Then** no panic occurs, the failure is logged, and the append is still attempted best-effort on the existing file.

- Vehicle: Go test, `internal/hook`.
- PASS = no panic + warning logged + append outcome matches fail-open semantics (attempted; success or logged write-failure both acceptable — the session must never block).

### AC-IBX-009 — Scope guard: no update-path or wiring changes (REQ-IBX-010)

**Given** the SPEC's merged change set, **When** `git diff --name-only <base>...` is inspected, **Then** no path under `internal/cli/update/` appears and `internal/template/templates/.claude/settings.json.tmpl` is unchanged (no new hook entries anywhere in the template tree).

- Vehicle: run-phase E4 evidence (git diff path filter, exit-code-checked).
- PASS = both path assertions hold with grep exit code 1 (no matches).

### AC-IBX-010 — Concurrency across the cap boundary (REQ-IBX-001, NFC-1)

**Given** N concurrent appenders crossing the cap boundary, **When** the appends execute, **Then** no panic or lost-append error occurs and the final on-disk state is consistent (live + archives account for the appended lines within the documented one-in-flight interleaving tolerance).

- Vehicle: Go test with `-race`, `internal/hook`.
- PASS = race-clean + no panic + bounded final live size.

## §B. Edge Cases

- Empty inbox + `drain` on a standard install → exit 0, "nothing to rotate" stats, no archive created.
- Inbox exactly at cap (boundary equality) → rotation fires (cap is "at or over", per REQ-IBX-003).
- Marker directory created between the stat and the rotate → worst case one rotation lands on a would-be curator machine; the local drain's offset advance (t259 REQ-LDS-007/008) re-syncs on its next wrapper run; residual risk recorded, not guarded (one-stat race, self-healing).
- Windows rename-over-existing → remove-then-rename in the retention chain (NFC-5); verdict is CI windows matrix.
- Archive path already exists from a prior SPEC-less era → the chain overwrite path must be idempotent (delete-then-rename), not an error.

## §C. Quality Gate Criteria

- TRUST 5 Tested: new cap/rotation code covered; `go test -race` on the concurrency path (E5).
- TRUST 5 Secured: rotation operates only inside the project root on fixed names derived from the inbox path — no user-controlled path interpolation; archive deletion is bounded to the numbered-generation pattern.
- TRUST 5 Trackable: conventional commits referencing SPEC-INBOX-DRAIN-GAP-001.
- Harness level: standard (Tier M; no security/OWASP-sensitive surface — the inbox is 0o600 local state).

## §D. AC Matrix and Traceability

| AC | REQ | Severity | Vehicle | Machine-verifiable |
|---|---|---|---|---|
| AC-IBX-001 | REQ-IBX-001, 003 | MUST | Go test (hook) | yes — on-disk assertions |
| AC-IBX-002 | REQ-IBX-002 | MUST | Go test (hook) | yes — zero-archive assertion |
| AC-IBX-003 | REQ-IBX-004 | MUST | Go test (hook) | yes — count assertion |
| AC-IBX-004 | REQ-IBX-005 | MUST | CLI test | yes — exit code + output match |
| AC-IBX-005 | REQ-IBX-006 | MUST | CLI test | yes — exit code + on-disk rotation |
| AC-IBX-006 | REQ-IBX-007 | MUST | CLI test | yes — exit code + byte-hash equality |
| AC-IBX-007 | REQ-IBX-008 | MUST | Go golden test | yes — JSON parse assertions |
| AC-IBX-008 | REQ-IBX-009 | MUST | Go test (hook) | yes — no-panic + log + append |
| AC-IBX-009 | REQ-IBX-010 | MUST | git diff filter | yes — path assertions |
| AC-IBX-010 | REQ-IBX-001, NFC-1 | MUST | Go test `-race` | yes — race-clean run |

### §D.1 Severity policy

All ten ACs are MUST (the cap, stand-down, and refusal semantics are the SPEC's reason to exist; none is a nice-to-have). No SHOULD/MINOR rows.

### §D.2 Traceability

Every REQ-IBX-001..010 maps to at least one AC (see matrix); REQ-IBX-005..007 are the CLI trio verified by the CLI test trio; NFC-1..5 are verified indirectly (D.3).

### §D.3 Indirect verification

- NFC-2 (rotation cost): no dedicated AC — verified by design review (rename-chain is O(1) renames) and implicitly by AC-IBX-010 runtime.
- NFC-4 (curator parity): AC-IBX-002 is its proof.
- NFC-5 (Windows): AC-IBX-001/003 run in the CI windows matrix.
- NFC-3 (fail-open parity): AC-IBX-008.
- Consumer-vacuum premise: recorded in spec.md §B.1 with grep evidence; re-checkable at run phase with the same grep.

### §D.4 Closure gates

- All 10 ACs PASS with cited command output (E1) before `in-progress → implemented`.
- AC-IBX-009 evidence must be regenerated at sync phase against the final merge base (not carried from run phase).

### §D.5 Forward-looking checks

- If a future SPEC ships a user-facing inbox consumer, the schema-version field (REQ-IBX-008) is the migration hook — readers must continue tolerating absence.
- If the curator ever ships beyond the maintainer machine, the stand-down marker semantics (REQ-IBX-002) transfer unchanged — the marker travels with the state dir, not the binary.

## §E. Definition of Done

1. All §A scenarios PASS with verbatim command output cited in progress.md §E.2.
2. E1-E7 planned evidence (plan.md §E) all observed, not assumed.
3. No `internal/cli/update/` or template-tree changes in the final diff (AC-IBX-009 re-verified at sync).
4. Conventional commits reference the SPEC-ID.
