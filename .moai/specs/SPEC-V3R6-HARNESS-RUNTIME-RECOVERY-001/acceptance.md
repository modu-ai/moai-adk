# Acceptance — SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001

> **Canonical AC enumeration** for the runtime recovery doctrine SPEC.
> Each AC traces to one or more REQs in spec.md §F and binds to a milestone in plan.md §F.
> Evidence MUST be observed (per `.claude/rules/moai/core/verification-claim-integrity.md`); inferred PASS is a Gap, not a Claim.

## §D. AC Matrix

| AC ID | Severity | REQ trace | Milestone | Evidence type |
|-------|----------|-----------|-----------|---------------|
| AC-RR-001 | MUST | REQ-RR-001 | M1 | Rule file §1 names the 4-element withheld-recoverable-error set verbatim |
| AC-RR-002 | MUST | REQ-RR-002, REQ-RR-003, REQ-RR-004 | M1 | Rule file §2 contains the 4-rung ladder table with moai-adk artifact cross-references + cheapest-first ordering rule |
| AC-RR-003 | MUST | REQ-RR-005 | M1 | Rule file §3 lists all 5 circuit-breaker invariants, each traceable to book1 ch06 §5 framing |
| AC-RR-004 | MUST | REQ-RR-006, REQ-RR-007 | M2 | agent-common-protocol.md §Hook Invocation Surface contains a "Recovery-Signal Carve-Out" subsection (guidance, not mechanical enforcement); doctrine rule §4 mirrors it |
| AC-RR-005 | MUST | REQ-RR-008 | M3 | zone-registry.md contains exactly one new `CONST-V3R6-0XX` entry; `moai constitution list` returns it |
| AC-RR-006 | SHOULD | REQ-RR-006 | M1 | Rule §4 names both `sync-phase-quality-gate.sh` and `status-transition-ownership.sh` and states the exit-0 carve-out as a **documentation-only policy recommendation** (the current hooks do NOT mechanically enforce it; runtime-layer enforcement deferred to future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001) |
| AC-RR-007 | SHOULD | REQ-RR-009 | M1 | Rule §5 cross-references session-handoff.md, verification-claim-integrity.md, and book1 ch03/ch06 |
| AC-RR-008 | MUST | REQ-RR-011 | M2 | grep confirms NO `Ledger` / `Ledger Closure` heading added to agent-common-protocol.md outside §Hook Invocation Surface (anti-collision with P1a; REQ-RR-011 scope boundary) |
| AC-RR-009 | MUST | (Non-Goals) | M4 | spec.md §D / §E explicitly exclude Go runtime reimplementation; no `internal/` Go code added |
| AC-RR-010 | MUST | REQ-RR-010 | M4 | Gap-closing grep returns ≥1 hit per RECOVERY-LADDER term only: `reactive-compact`, `death-spiral`, `withheld-recoverable`, `circuit-breaker`. Error-type vocabulary (`max_output_tokens`, `media_size`, `compact-failure`) is EXCLUDED (pre-existing coverage in hooks-system.md). |
| AC-RR-011 | MUST | REQ-RR-012 | M1 | Rule body contains a sentence stating the agent MUST consult the doctrine + apply the cheapest-first ladder before concluding the turn failed (normative agent obligation) |

### §D.1 Severity definitions

- **MUST**: blocks close. A MUST AC that fails MUST be resolved before sync-phase.
- **SHOULD**: does not block close, but the gap MUST be recorded in `progress.md` §E.2 Residual-risk as a forward-looking follow-up.

### §D.2 Traceability

Every REQ in spec.md §F has at least one AC. REQ → AC mapping:

- REQ-RR-001 → AC-RR-001
- REQ-RR-002 → AC-RR-002
- REQ-RR-003 → AC-RR-002
- REQ-RR-004 → AC-RR-002
- REQ-RR-005 → AC-RR-003
- REQ-RR-006 → AC-RR-004, AC-RR-006
- REQ-RR-007 → AC-RR-004
- REQ-RR-008 → AC-RR-005
- REQ-RR-009 → AC-RR-007
- REQ-RR-010 → AC-RR-010
- REQ-RR-011 → AC-RR-008
- REQ-RR-012 → AC-RR-011

## §D.3 Given-When-Then Scenarios

### Scenario AC-RR-001 — withheld-recoverable-error set named

**Given** the file `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` exists (post-M1),
**When** a maintainer reads §1,
**Then** the section names exactly four error types — `prompt_too_long (PTL)`, `max_output_tokens`, `media_size`, `compact-failure` — and states they are routed to layered recovery rather than surfaced raw.

**Evidence command**:
```bash
grep -nE 'prompt_too_long|PTL|max_output_tokens|media_size|compact-failure' \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: ≥4 hits, one per error type
```

### Scenario AC-RR-002 — 4-rung cheapest-first ladder

**Given** the doctrine rule §2,
**When** a maintainer reads the ladder table,
**Then** all 4 rungs are present, ordered cheapest-first, each with a moai-adk artifact cross-reference (rung 1 = in-turn self-correction; rung 2 = session-handoff.md; rung 3 = session-handoff.md §Worktree-Anchored; rung 4 = progress.md + PRESERVE), and a cheapest-first ordering rule prohibits jumping rungs.

**Evidence command**:
```bash
grep -nE 'rung|cheapest-first|session-handoff|Worktree-Anchored|progress.md' \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: hits for each rung + the ordering rule
```

### Scenario AC-RR-003 — 5 circuit-breaker invariants

**Given** the doctrine rule §3,
**When** a maintainer reads the invariants list,
**Then** all 5 invariants are present: (1) max-consecutive-autocompact-failure analogue, (2) hasAttemptedReactiveCompact no-self-loop analogue, (3) compact-can-PTL last-resort escape, (4) abort-closes-ledger, (5) narrative-consistency requirement — each citing book1 ch06 §5 framing.

**Evidence command**:
```bash
grep -ncE 'hasAttemptedReactiveCompact|autocompact|truncateHead|abort-closes-ledger|narrative' \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: ≥5 distinct invariant markers
```

### Scenario AC-RR-004 — anti-death-spiral carve-out in render surface + SSOT (documentation-only guidance)

**Given** M2 is complete,
**When** a maintainer reads `agent-common-protocol.md` §Hook Invocation Surface,
**Then** a "Recovery-Signal Carve-Out" subsection documents (as policy GUIDANCE to agents and future hook authors, NOT as mechanically-enforced gate) that Stop/PostToolUse hooks should exit 0 on recovering turns; AND the doctrine rule §4 mirrors the same carve-out. The heading literal "Recovery-Signal Carve-Out" appears in BOTH surfaces (per REQ-RR-007 pinning) so a future runtime-layer hook SPEC can locate it.

**Evidence command**:
```bash
grep -nE 'recovery.signal|carve-out|Recovery-Signal' \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: hits in BOTH files (SSOT + render surface parity). The grep is intentionally
# loose (recovery.signal|carve-out) to avoid brittleness on exact heading capitalization.
```

### Scenario AC-RR-005 — zone-registry CONST entry queryable

**Given** M3 is complete,
**When** a maintainer runs `moai constitution list`,
**Then** the output includes a `CONST-V3R6-0XX` entry whose `clause` names the anti-death-spiral invariant and whose `file` is `runtime-recovery-doctrine.md`.

**Evidence command**:
```bash
grep -nE 'CONST-V3R6-0' .claude/rules/moai/core/zone-registry.md | tail -5
moai constitution list 2>/dev/null | grep -i 'recovery\|death.spiral' || echo "(moai CLI unavailable — record as Gap)"
```

### Scenario AC-RR-006 — both hooks named in the carve-out (SHOULD, documentation-only)

**Given** the doctrine rule §4,
**When** a maintainer greps for the hook filenames,
**Then** BOTH `sync-phase-quality-gate.sh` and `status-transition-ownership.sh` appear in §4 with the exit-0 carve-out statement — stated as a **documentation-only policy recommendation** (the current hooks do NOT parse a recovery signal and therefore cannot mechanically enforce the carve-out; runtime-layer enforcement is deferred to a follow-up SPEC, forward-link: future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001).

**Evidence command**:
```bash
grep -nE 'sync-phase-quality-gate.sh|status-transition-ownership.sh' \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: ≥2 hits (one per hook). AC is SHOULD because the carve-out is guidance,
# not mechanically enforced by the current hooks (see REQ-RR-006 scope binding).
```

### Scenario AC-RR-007 — cross-references present (SHOULD)

**Given** the doctrine rule §5,
**When** a maintainer greps for cross-reference targets,
**Then** the section references `session-handoff.md`, `verification-claim-integrity.md`, and book1 ch03/ch06 (by chapter name or `book1`/`harness-books` literal).

**Evidence command**:
```bash
grep -nE 'session-handoff|verification-claim-integrity|book1|ch03|ch06|harness-books' \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: ≥4 distinct cross-reference targets
```

### Scenario AC-RR-008 — agent-common-protocol.md boundary intact (anti-collision with P1a; REQ-RR-011)

**Given** M2 is complete,
**When** a maintainer greps `agent-common-protocol.md` for `Ledger` headings OUTSIDE §Hook Invocation Surface,
**Then** no new `Ledger` / `Ledger Closure` heading exists outside the §Hook Invocation Surface region (lines 35-63 baseline + the new carve-out subsection). The §Ledger Closure subsection is reserved for P1a SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001. REQ-RR-011 establishes this scope boundary normatively.

**Evidence command**:
```bash
grep -nE '^#+.*Ledger' .claude/rules/moai/core/agent-common-protocol.md
# Expected: ZERO hits (this SPEC does NOT add a Ledger heading; P1a does)
```

### Scenario AC-RR-009 — Non-Goals exclude Go runtime

**Given** spec.md is at plan-phase draft,
**When** a maintainer reads §D / §E,
**Then** both sections explicitly exclude Go runtime reimplementation of CC internals (`queryLoop`, `recoverFromError`, `truncateHeadForPTLRetry`, `hasAttemptedReactiveCompact`); AND no `internal/recovery/` Go package was added at run-phase.

**Evidence command**:
```bash
grep -nE 'queryLoop|recoverFromError|truncateHeadForPTLRetry|hasAttemptedReactiveCompact|NOT.*Go|policy-layer' \
  .moai/specs/SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001/spec.md
ls internal/recovery/ 2>/dev/null && echo "FAIL: Go package added" || echo "PASS: no Go package"
```

### Scenario AC-RR-010 — grep reproducibility closes the recovery-ladder zero-hit gap

**Given** M4 is complete,
**When** a maintainer runs the gap-closing grep across `.claude/rules/moai/`,
**Then** each of the RECOVERY-LADDER terms — `reactive-compact`, `death-spiral`, `withheld-recoverable`, `circuit-breaker` — returns ≥1 hit; the pre-SPEC recovery-ladder zero-hit gap is closed. The error-type vocabulary (`max_output_tokens`, `media_size`, `compact-failure`) is EXCLUDED from this AC: `max_output_tokens` already has pre-existing coverage in `hooks-system.md` (L30/L105/L309 as a StopFailure error type), and the gap this SPEC closes is the *recovery-ladder* vocabulary specifically, not the error-type vocabulary.

**Evidence command**:
```bash
for term in 'reactive-compact' 'death-spiral' 'withheld-recoverable' 'circuit-breaker'; do
  count=$(grep -rEl "$term" .claude/rules/moai/ 2>/dev/null | wc -l | tr -d ' ')
  echo "$term: $count file(s)"
  test "$count" -ge 1 || echo "  FAIL: $term still zero-hit"
done
# Expected: every RECOVERY-LADDER term ≥1. Error-type vocabulary intentionally NOT asserted here.
```

### Scenario AC-RR-011 — agent consult-the-doctrine obligation (REQ-RR-012)

**Given** the doctrine rule is authored (post-M1),
**When** a maintainer greps the doctrine rule for the agent obligation sentence,
**Then** the rule body contains a sentence stating the agent MUST consult the doctrine + apply the cheapest-first ladder (§F.2 / rule §2) BEFORE concluding the turn failed — so the recovery obligation is normatively attached to the agent, not only to the doctrine document.

**Evidence command**:
```bash
grep -nE 'consult|cheapest-first|before concluding' \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md
# Expected: ≥1 hit containing the agent-consult obligation (REQ-RR-012)
```

## §D.4 Edge Cases

- **E1 — `moai constitution list` unavailable in CI**: AC-RR-005's second evidence command may return "(moai CLI unavailable)". Record as a **Gap** (not a FAIL) and fall back to the `grep CONST-V3R6-0` evidence; the grep alone satisfies the AC.
- **E2 — parallel SPEC allocates `CONST-V3R6-001` first**: M3 re-verifies the highest V3R6 numeric at start and allocates the next free ID. The AC is satisfied by *any* valid `CONST-V3R6-0XX`, not by a specific number.
- **E3 — book1 chapter numbering drift**: AC-RR-007 accepts either chapter number (`ch03`/`ch06`) or chapter title ("Query Loop is the heartbeat" / "Errors and recovery") or repo literal (`harness-books` / `book1`). Do not over-constrain to one form.
- **E4 — error-type vocabulary out of scope for AC-RR-010**: AC-RR-010 asserts the RECOVERY-LADDER vocabulary only (`reactive-compact`, `death-spiral`, `withheld-recoverable`, `circuit-breaker`). The error-type vocabulary (`max_output_tokens`, `media_size`, `compact-failure`) is intentionally EXCLUDED because `max_output_tokens` already has pre-existing coverage in `hooks-system.md` as a StopFailure error type. The gap this SPEC closes is the recovery-ladder vocabulary specifically.

## §D.5 Indirect Verification

Where direct grep is insufficient (e.g., "does the doctrine actually bind agent behavior?"), indirect verification applies:

- **Doctrine binding (documentation-only)**: the carve-out's presence in BOTH the SSOT rule and the render surface (`agent-common-protocol.md`) is the indirect evidence that the doctrine binds AGENT BEHAVIOR and FUTURE hook authors — a reader of either surface reaches the same guidance. AC-RR-004 enforces this parity. The binding is policy/doctrine guidance only; the current hooks do NOT mechanically enforce it (REQ-RR-006/007 scope binding).
- **Cheapest-first enforcement**: the ladder table + the ordering rule (REQ-RR-004) together indirectly enforce that an agent cannot skip rungs; no runtime check is possible at the policy layer (and per §D, none is wanted).

## §D.6 Closure Gates

**Definition of Done** (Tier M, policy-layer SPEC):

- [ ] All 9 MUST ACs (AC-RR-001/002/003/004/005/008/009/010/011) PASS with observed evidence (no inferred PASS per verification-claim-integrity.md).
- [ ] The 2 SHOULD ACs (AC-RR-006 documentation-only carve-out naming; AC-RR-007 cross-references) either PASS or are recorded in `progress.md` §E.2 Residual-risk.
- [ ] `moai spec lint` on the SPEC directory is clean (or `lint.skip` debt is documented with rationale).
- [ ] No Go code added under `internal/` (AC-RR-009 boundary).
- [ ] No hook script bodies modified (anti-pattern AP-RR-003) — the carve-out is documentation-only policy guidance (REQ-RR-006/007).
- [ ] No `Ledger` heading added outside §Hook Invocation Surface (AC-RR-008 / REQ-RR-011 boundary).

## §D.7 Forward-Looking Checks

- **FL-1**: When P1a `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001` lands, its `§Ledger Closure` subsection in `agent-common-protocol.md` MUST NOT collide with this SPEC's "Recovery-Signal Carve-Out" subsection. Both subsections live under §Hook Invocation Surface; they are sibling subsections, not nested.
- **FL-2**: A future runtime-layer SPEC (`SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`) MAY propose a hook that parses `stopReason` to mechanically enforce the recovery-signal carve-out. This SPEC explicitly defers mechanical enforcement to that follow-up; this SPEC's ACs (esp. AC-RR-006, which is SHOULD and documentation-only) MUST NOT be read as requiring current-hook mechanical enforcement.
- **FL-3**: If book1 is revised and the named principles move chapters, the cross-references in the doctrine rule §5 should be updated in a follow-up chore commit, not a new SPEC (citation maintenance, not doctrine change).
