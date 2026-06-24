# SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001 — Implementation Plan

> **Tier M** (standard) — 3 core artifacts + progress.md §E skeleton. Plan-phase only (CON-4); no run-phase execution in this delivery.

## §A. Context

Sprint 15 harness-books application cohort, P1a entry. The cohort applies externally grounded invariants from `github.com/wquguru/harness-books` (book1 ch04 + ch07) to MoAI-ADK orchestration doctrine. This SPEC introduces the `ledger` / `synthetic result` / `dangling tool_use` vocabulary into active doctrine — a verified grep-zero-hit gap (see spec.md §A.1).

The P0 sibling `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` owns the §Hook Invocation Surface Recovery-Signal Carve-Out in the same target file (`agent-common-protocol.md`). This SPEC's §Ledger Closure subsection goes in a **distinct** section (REQ-LEDGER-005) to avoid collision.

## §B. Known Issues

| Issue | Status | Impact on this SPEC |
|-------|--------|---------------------|
| team-ac-verify.sh has no exit-2 path today | Open (this SPEC introduces minimal stub) | M2 must add both the reject-path trigger AND the `ledger_note` field, but the trigger is a minimal stub (full AC verification deferred per §X.1) |
| P0 sibling SPEC not yet merged to `agent-common-protocol.md` | Open | Collision check (AC-LEDGER-006) runs regardless of merge order — this SPEC edits a different section |

## §C. Pre-flight (run-phase entry verification)

Before M1 edit, run-phase MUST verify:

1. **Section anchors present** — `grep -nE '^## User Interaction Boundary|^### Hook Invocation Surface|^### Blocker Report Format' .claude/rules/moai/core/agent-common-protocol.md` returns 3 hits. The `### Ledger Closure` subsection will be added as a sibling under `## User Interaction Boundary`, immediately after `### Blocker Report Format`.
2. **No existing `Ledger Closure` section** — `grep -n 'Ledger Closure' .claude/rules/moai/core/agent-common-protocol.md` returns 0 hits (idempotency check — no prior partial edit).
3. **team-ac-verify.sh current state** — hook has no `exit 2` and no `ledger_note` today. M2 adds both; the trigger condition is a minimal stub per §X.1.
4. **moai.md §8 Error Recovery banner current state** — banner exists at line ~587; M3 adds a small annotation below the banner's existing options list (A/B/C/D). No structural rewrite.
5. **P0 sibling collision check** — `grep -n 'Recovery-Signal\|Recovery Carve-Out\|Carve-Out' .claude/rules/moai/core/agent-common-protocol.md` may return 0 hits (P0 not merged yet) or N hits (P0 merged). Either way, the §Ledger Closure subsection must NOT be placed inside `### Hook Invocation Surface`. Verify by line range: `### Ledger Closure` line number < `### Hook Invocation Surface` line number OR `### Ledger Closure` is under a different H2.

## §D. Constraints (carried from spec.md §C)

- CON-1: The MEANING of exit 2 in team-ac-verify.sh remains "reject completion" (unchanged). This SPEC ADDS a minimal reject path (currently ABSENT from the hook — verified at plan-time the hook has zero exit-2 branches) that emits exit 2 with a `ledger_note` field. The trigger is a minimal stub per §X.1 (full AC-verification deferred). This is a behavioral ADDITION (new code path), honestly described — not a no-op field-only addition.
- CON-2: moai.md §8 = small annotation only.
- CON-3: §Ledger Closure in NEW subsection, distinct from §User Interaction Boundary prose AND §Hook Invocation Surface.
- CON-4: Plan-phase only — NO run-phase in this delivery.
- CON-5: `## §X. Out of Scope` + h3 pattern (avoid MissingExclusions lint ERROR).
- CON-6: book1 ch04 + ch07 cited.

## §E. Self-Verification (plan-phase deliverable completeness)

- [x] spec.md frontmatter: 12 canonical fields + `era: V3R6` + `depends_on` (P0 sibling).
- [x] spec.md uses GEARS notation (no `IF/THEN`).
- [x] spec.md has `## §X. Out of Scope` with h3 sub-sections (CON-5).
- [x] spec.md cites book1 ch04 + ch07 (CON-6).
- [x] plan.md (this file) carries pre-flight checks (§C), milestones (§F), anti-patterns (§G), cross-references (§H).
- [x] acceptance.md has ≥ 2 Given-When-Then scenarios per MUST-PASS AC + grep reproducibility + scope-boundary AC.
- [x] progress.md §E skeleton emitted with all 5 placeholder headings (§E.1..§E.5).

## §F. Milestones

> Milestones are priority-ordered work steps, NOT time estimates. Per `.claude/rules/moai/development/sprint-round-naming.md`, these are Milestones (M1..M4) within this single-SPEC Tier M lifecycle.

### M1 — agent-common-protocol.md §Ledger Closure

**Priority**: High (first milestone — content foundation).

**Deliverable**: New `### Ledger Closure` subsection under `## User Interaction Boundary`, placed immediately after `### Blocker Report Format` (collision-free per pre-flight §C.5).

**Content** (prose + REQ recap, ≤ 80 lines):
- Opening: names the ledger-closure invariant (book1 ch04 "账本闭环") and the dangling-tool_use hazard.
- Clause (a): REQ-LEDGER-001 — synthetic result on aborted Agent() delegation.
- Clause (b): REQ-LEDGER-002 — team-ac-verify.sh exit-2 `ledger_note` field (cross-reference to hook).
- Clause (c): REQ-LEDGER-003 — TeammateIdle exit-2 task closure.
- Clause (d): REQ-LEDGER-006 — cross-references (book1 ch04, ch07, session-handoff.md Block 3-4).
- Scope-boundary note: "This subsection is distinct from the §Hook Invocation Surface Recovery-Signal Carve-Out owned by SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001."

**Exit criteria**: AC-LEDGER-001 (clauses a-d present), AC-LEDGER-006 (collision-free), AC-LEDGER-007 (book1 cited).

### M2 — team-ac-verify.sh `ledger_note` field

**Priority**: High (second milestone — hook-side companion to clause (b)).

**Deliverable**: Edit `.claude/hooks/moai/team-ac-verify.sh`:
- Add a minimal reject path that emits exit 2 with a JSON object containing `ledger_note`. The trigger is a minimal stub (per §X.1 — full AC verification is out of scope). Suggested stub: a `--reject` test flag OR a documented TODO comment marking where future AC verification will emit exit 2.
- On the existing allow/dormant paths, NO change — they continue to exit 0 without `ledger_note` (the field is only meaningful on reject).
- The `ledger_note` value is a short human-readable rejection reason.

**Exit criteria**: AC-LEDGER-002 (`grep -n 'ledger_note' .claude/hooks/moai/team-ac-verify.sh` returns ≥ 1 hit; exit-2 path emits the field). Exit-code semantics unchanged (CON-1).

### M3 — moai.md §8 Error Recovery banner annotation

**Priority**: Medium (third milestone — render-surface reminder).

**Deliverable**: Edit `.claude/output-styles/moai/moai.md` §8 Error Recovery banner — add a small "Interrupt Closure" annotation. Suggested placement: immediately below the banner's A/B/C/D options list, before the closing `──────`. Example annotation (≤ 3 lines):

```
🤖 MoAI ★ Error ──────────────────────────────
❌ [what broke]
🔍 [root cause if known]
🔧 Recovery options via AskUserQuestion:
  A. Retry as-is  B. Alt approach  C. Pause  D. Abort+preserve
📎 Interrupt Closure: if an Agent() delegation was aborted (not merely failed),
   reference the synthetic ledger-closing artifact above before retrying —
   do not proceed as if the delegation returned cleanly.
──────────────────────────────────────────────
```

**Exit criteria**: AC-LEDGER-004 (banner has "Interrupt Closure" annotation). No structural rewrite of §8 (CON-2).

### M4 — Lint clean + grep reproducibility + scope-boundary verification

**Priority**: Medium (final milestone — quality gate).

**Deliverable**:
- Run `moai spec lint .moai/specs/SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001/spec.md` — expect exit 0.
- Run the phrase-targeted grep (AC-LEDGER-005) — expect ≥ 1 hit per term post-edit (`ledger`, `synthetic`, `dangling tool_use` introduced into active doctrine).
- Run the scope-boundary grep (AC-LEDGER-006) — expect `### Ledger Closure` and the P0 SPEC's Recovery-Signal Carve-Out in distinct sections.
- Run `moai spec audit --json` on the SPEC — confirm era `V3R6`, no drift findings.

**Exit criteria**: All 6 MUST-PASS ACs (AC-LEDGER-001, 002, 004, 005, 006, 007) independently re-verified with live grep/read evidence.

## §G. Anti-Patterns

- **AP-L-001 — Embedding §Ledger Closure inside §Hook Invocation Surface prose**: violates REQ-LEDGER-005 / CON-3 / AC-LEDGER-006. The subsection goes under `## User Interaction Boundary` as a sibling, not nested inside §Hook Invocation Surface.
- **AP-L-002 — Full AC verification logic in M2**: violates §X.1. The reject-path trigger is a minimal stub; full verification is a follow-up SPEC.
- **AP-L-003 — Structured ledger artifact (JSON schema, state file)**: violates §X.3. The artifact is a prose summary.
- **AP-L-004 — Structural rewrite of moai.md §8**: violates CON-2. M3 adds a small annotation; the banner's A/B/C/D structure is preserved.
- **AP-L-005 — Changing team-ac-verify.sh exit-code semantics**: violates CON-1. Exit 2 = reject (unchanged); only `ledger_note` is added.
- **AP-L-006 — Overclaiming grep-zero-hit at run-time if bare `ledger` appears elsewhere**: spec.md §A.3 documents the honest baseline; AC-LEDGER-005 uses the phrase-targeted grep.

## §H. Cross-References

- spec.md §A.4 — source material (book1 ch04 + ch07).
- spec.md §B — REQ-LEDGER-001..006 (GEARS).
- spec.md §X — Out of Scope (§X.1..§X.5).
- acceptance.md §D — AC-LEDGER-001, 002, 004, 005, 006, 007 (6 ACs; full Given-When-Then). REQ-LEDGER-003 is verified by AC-LEDGER-001 clause (c) — no standalone AC-003.
- progress.md §E.1 — plan-phase audit-ready signal (populated by manager-spec at plan-phase close).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 — the ledger-closing artifact must be a real summary, not a fabricated success (binds REQ-LEDGER-001 truthfulness).
- `.claude/rules/moai/workflow/session-handoff.md` Block 3-4 — persistence-layer analogue (REQ-LEDGER-006 cross-reference).
- P0 sibling: `.moai/specs/SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001/` — collision-avoidance dependency.
