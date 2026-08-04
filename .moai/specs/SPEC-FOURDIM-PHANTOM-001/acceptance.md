# acceptance.md — SPEC-FOURDIM-PHANTOM-001

## §A Reference

- spec.md §B REQ-FP-001..010 (GEARS requirements — the requirement layer).
- plan.md §C (design decisions), §D (design tension), §F (milestones).

## §B Verification Philosophy

Every AC row is binary-testable (PASS / FAIL), written as Given-When-Then. Evidence is a command actually run + its verbatim output, never a summary (`verification-claim-integrity.md` §2). A FAIL verdict on a phantom mechanism MUST cite the literal `probe_command` and `actual_matches: 0` — the reviewer can re-run the exact command and observe the zero.

## §C Test Surface

- Unit: the verdict-phase JS function — feed it synthetic judge + probe_results payloads, assert the verdict branch and payload shape.
- Integration: run `sync-audit-4dim.js` end-to-end against a fixture SPEC carrying `claimed_mechanisms[]`; cover the phantom-FAIL, happy-path-passthrough, and precedence cases.

## §D Acceptance Criteria Matrix

### AC-FP-001 — Phantom mechanism yields FAIL with named phantom in payload

**Given** a SPEC under audit declares `claimed_mechanisms: [{name: "input-validation-guard", probe_command: "grep -r 'ValidateInput' <diff-paths>", expected_match_substring: "ValidateInput"}]`, and the diff / produced files contain zero occurrences of `ValidateInput` (actual_matches == 0), and all four judges returned finite non-zero scores,
**When** the verdict phase runs,
**Then** the verdict is `FAIL` AND the payload carries `phantom_mechanisms: [{name: "input-validation-guard", probe_command: "...", expected_match_substring: "ValidateInput", actual_matches: 0}]` AND the verdict does NOT fall through to the harmonic-mean computation.

**Severity**: MUST-pass. **Trace**: REQ-FP-001, REQ-FP-002, REQ-FP-004, REQ-FP-007.

### AC-FP-002 — Happy-path passthrough: real mechanism passes the guard unchanged

**Given** a SPEC under audit declares `claimed_mechanisms: [{name: "input-validation-guard", probe_command: "...", expected_match_substring: "ValidateInput"}]`, and the diff / produced files contain ≥1 occurrence of `ValidateInput` (actual_matches > 0), and the four judges returned finite non-zero scores with a harmonic mean ≥ threshold,
**When** the verdict phase runs,
**Then** the verdict is `PASS` AND the payload carries `harmonic_mean` AND the payload does NOT carry any `phantom_mechanisms[]` entry for `input-validation-guard` (the guard fell through silently).

**Severity**: MUST-pass. **Trace**: REQ-FP-006.

### AC-FP-003 — Precedence: null-judge (INCOMPLETE) fires before the phantom guard

**Given** at least one judge returned null / non-finite score, and a claimed mechanism has actual_matches == 0,
**When** the verdict phase runs,
**Then** the verdict is `INCOMPLETE` (the null-judge guard fires first) AND the payload names the missing dimension(s) in `missing[]` AND the payload does NOT carry `phantom_mechanisms[]` (the phantom guard was never reached).

**Severity**: MUST-pass. **Trace**: REQ-FP-003, REQ-FP-008.

### AC-FP-004 — Precedence: zero-score (FAIL) fires before the phantom guard

**Given** all four judges returned finite scores, at least one score is 0, and a claimed mechanism has actual_matches == 0,
**When** the verdict phase runs,
**Then** the verdict is `FAIL` naming the zero-scored dimension(s) in `zero_scored[]` AND the payload does NOT carry `phantom_mechanisms[]` (the phantom guard sits AFTER the zero-score guard).

**Severity**: MUST-pass. **Trace**: REQ-FP-003, REQ-FP-008.

### AC-FP-005 — Per-mechanism baseline attribution in the FAIL payload

**Given** a phantom-FAIL verdict was returned,
**When** the payload is inspected,
**Then** every entry in `phantom_mechanisms[]` carries `{name, probe_command, expected_match_substring, actual_matches: 0}` — the `probe_command` is the verbatim command that was run and `actual_matches: 0` is the observed output, per `verification-claim-integrity.md` §2.

**Severity**: MUST-pass. **Trace**: REQ-FP-007.

### AC-FP-006 — Deterministic verdict logic (no LLM decides FAIL)

**Given** the verdict-phase JS function and synthetic `probe_results` payloads,
**When** the function is invoked twice with identical inputs,
**Then** the verdict and payload are byte-identical (the `actual_matches == 0` comparison is pure JS, never an LLM judgment) AND the function reads no wall-clock and draws no random value.

**Severity**: MUST-pass. **Trace**: REQ-FP-005.

### AC-FP-007 — Context agent executes probes against the actual write surface

**Given** a SPEC declares a `probe_command` whose `expected_match_substring` exists in the diff.patch paths / produced files but NOT in the SPEC's declared `target_surface` intent (the mechanism was implemented at a different path than declared),
**When** the Context agent executes the probe,
**Then** `actual_matches > 0` (the probe hit the actual write surface, not the declared intent) — the phantom guard does NOT falsely FAIL a real mechanism that lives at an undeclared path.

**Severity**: MUST-pass. **Trace**: REQ-FP-002, lesson `feedback_lsel_f1_actual_write_surface`.

### AC-FP-008 — Composable: 4-dimension enum FROZEN

**Given** the phantom guard is wired in,
**When** the DIMENSIONS array is inspected,
**Then** it contains exactly `['Functionality', 'Security', 'Craft', 'Consistency']` — the phantom guard did NOT add a 5th dimension; it is a verdict-phase guard, not a judge.

**Severity**: MUST-pass. **Trace**: REQ-FP-008.

### AC-FP-009 — Template-First mirror parity

**Given** the live `.claude/workflows/sync-audit-4dim.js` has been edited to wire the phantom guard,
**When** `diff .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` is run,
**Then** the two files are byte-identical AND `make build` regenerated the embedded catalog (exit 0).

**Severity**: MUST-pass. **Trace**: REQ-FP-010, CLAUDE.local.md §2.

### AC-FP-010 — Template neutrality §25

**Given** the distributed JS (live + mirror),
**When** `grep -E 'SPEC-FOURDIM-PHANTOM-001|REQ-FP-|IMP-5|2026-08' internal/template/templates/.claude/workflows/sync-audit-4dim.js` is run,
**Then** zero matches (the distributed JS carries no SPEC IDs / REQ tokens / IMP-5 references / internal dates; the guard is named "phantom-mechanism guard" generically).

**Severity**: MUST-pass. **Trace**: REQ-FP-009, CLAUDE.local.md §25.

### AC-FP-011 — Empty `claimed_mechanisms[]` makes the phantom guard a no-op

**Given** a SPEC under audit declares no `claimed_mechanisms[]` (the field is absent or set to an empty array),
**When** the verdict phase runs,
**Then** the phantom-mechanism guard is a no-op (the payload carries NO `phantom_mechanisms[]` entry, no probe is executed) AND the verdict falls through to the harmonic-mean computation unchanged.

**Severity**: MUST-pass. **Trace**: REQ-FP-006.

### AC-FP-012 — Probe execution error / missing `actual_matches` routes to `evidence_gaps[]`

**Given** a declared probe whose execution returned an error or produced no `actual_matches` (the probe command errored, the file was not found, or the Context agent could not produce a count),
**When** the verdict phase runs,
**Then** the affected mechanism is reported in `evidence_gaps[]` (naming the mechanism, the probe_command, and the error/missing-data reason) AND the verdict is NOT a spurious PASS or FAIL based on the missing data — the mechanism is neither treated as phantom (which would require `actual_matches == 0`) nor as verified (which would require `actual_matches > 0`).

**Severity**: MUST-pass. **Trace**: REQ-FP-005.

## §E Severity Scale

- **MUST-pass** — failure blocks merge. AC-FP-001 through AC-FP-012 are all MUST-pass at this Tier (S).
- **SHOULD-pass** — failure triggers a debt marker but does not block merge. None at this Tier.
- **NICE-to-have** — failure is a tracking item. None at this Tier.

## §F Indirect Verification

The phantom guard's correctness is indirectly verified by:
- The existing harmonic-mean tests (unchanged — the guard composes, not replaces).
- The existing null-judge and zero-score tests (unchanged — the guard sits after both).
- AC-FP-003 + AC-FP-004 explicitly assert the precedence invariant.

## §G Closure Gates (Definition of Done)

- All 12 MUST-pass AC rows show PASS with cited command + verbatim output.
- Live + mirror byte-identical, `make build` exit 0.
- §25 neutrality grep returns zero matches in the distributed JS.
- `sync-audit-4dim.js` carries no SPEC IDs / REQ tokens / IMP-5 / dates / SHAs.
- progress.md §E.2 run-phase evidence populated by manager-develop; §E.3 audit-ready signal populated; §E.4 sync-phase signal populated by manager-docs.

## §H Forward-Looking Checks (post-merge)

- Monitor whether any audit raises a phantom-FAIL for a mechanism the SPEC author believed was real — if so, the probe-fidelity question (plan.md §D Design Tension) may warrant revisiting the dedicated 6th probe-agent alternative.
- Monitor whether Context-agent payload size becomes a concern at high probe-declaration counts (>10 per SPEC) — if so, consider paginating probes or moving to a dedicated probe agent.
- **Vacuous-probe / false-negative (self-deception) monitor** — the inverse failure mode from a phantom-FAIL: a SPEC under audit declares a trivially-matching probe whose outcome is pre-determined to PASS (e.g. an `expected_match_substring` that appears in boilerplate, or a `probe_command` whose grep pattern is certain to match regardless of whether the mechanism actually operates), making the phantom guard vacuously PASS. This is a DISTINCT failure class from the phantom-FAIL this guard catches: a phantom-FAIL catches a declared-but-absent mechanism (implementation/drift deception); a vacuous probe catches a probe whose match is structurally guaranteed (SPEC-author self-deception at authoring time, not runtime drift). Probe-fidelity validation — judging whether a declared probe is faithful to the mechanism it claims to verify — is a future-SPEC concern and/or a sync-auditor responsibility, NOT this phantom guard's job; the guard executes probes deterministically and is agnostic to the probe's faithfulness. Recorded as documented Tier-S debt: at this tier the guard ships without a probe-fidelity check, and a follow-up SPEC (or a sync-auditor skeptical read of declared probes) is the place that concern is revisited.
