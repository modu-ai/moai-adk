# SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001 — Acceptance Criteria

> Canonical AC enumeration. Each MUST-PASS AC has ≥ 1 Given-When-Then scenario. Edge cases, quality gates, and Definition of Done at the end.

## §D. AC Matrix

| AC ID | REQ covered | Severity | Milestone | Verification surface |
|-------|-------------|----------|-----------|----------------------|
| AC-LEDGER-001 | REQ-LEDGER-001, REQ-LEDGER-003, REQ-LEDGER-006 | MUST-PASS | M1 | `.claude/rules/moai/core/agent-common-protocol.md` has `### Ledger Closure` with clauses (a), (b), (c), (d) |
| AC-LEDGER-002 | REQ-LEDGER-002 | MUST-PASS | M2 | `.claude/hooks/moai/team-ac-verify.sh` exit-2 path emits `ledger_note` JSON field |
| AC-LEDGER-004 | REQ-LEDGER-004 | MUST-PASS | M3 | `.claude/output-styles/moai/moai.md` §8 Error Recovery banner has "Interrupt Closure" annotation |
| AC-LEDGER-005 | (gap-closure) | MUST-PASS | M4 | grep reproducibility — `ledger` / `synthetic` / `dangling tool_use` vocabulary introduced into active doctrine |
| AC-LEDGER-006 | REQ-LEDGER-005 (scope boundary) | MUST-PASS | M1, M4 | §Ledger Closure and P0 SPEC's Recovery-Signal Carve-Out in distinct sections |
| AC-LEDGER-007 | REQ-LEDGER-006 (book citation) | MUST-PASS | M1 | book1 ch04 + ch07 cited in §Ledger Closure |

## §D.1 Severity Model

- **MUST-PASS**: blocks 4-phase close. All of AC-LEDGER-001, 002, 004, 005, 006, 007 (6 MUST-PASS ACs) are MUST-PASS. Note: there is intentionally NO `AC-LEDGER-003`; the ID sequence jumps `001 → 002 → 004` by design. REQ-LEDGER-003 is verified by AC-LEDGER-001 clause (c) (see §D.2 traceability note below).
- **SHOULD-PASS**: no SHOULD-PASS ACs in this SPEC (Tier M minimal scope).
- **NICE-TO-HAVE**: none.

## §D.2 AC↔REQ Traceability (100% coverage)

| REQ | Covered by |
|-----|-----------|
| REQ-LEDGER-001 | AC-LEDGER-001 (clause a) |
| REQ-LEDGER-002 | AC-LEDGER-002 |
| REQ-LEDGER-003 | AC-LEDGER-001 (clause c) |
| REQ-LEDGER-004 | AC-LEDGER-004 |
| REQ-LEDGER-005 | AC-LEDGER-006 |
| REQ-LEDGER-006 | AC-LEDGER-001 (clause d) + AC-LEDGER-007 |

No REQ is uncovered. No AC is orphaned.

> **Traceability note on REQ-LEDGER-003 (no standalone AC-LEDGER-003):** REQ-LEDGER-003 (TeammateIdle exit-2 task closure) is verified by **AC-LEDGER-001 clause (c)** — it appears as one clause inside the §Ledger Closure subsection's four-clause enumeration. There is intentionally **no standalone AC-LEDGER-003**; the AC ID sequence jumps `001 → 002 → 004` by design. This matches spec.md §E matrix (the REQ-LEDGER-003 row cites `AC-LEDGER-001 (clause c)`). The two files agree — no dangling `AC-LEDGER-003` reference exists in either spec.md §E or acceptance.md §D.2.

## §D.3 Given-When-Then Scenarios

### AC-LEDGER-001 — §Ledger Closure subsection has clauses (a)-(d)

**Given** `.claude/rules/moai/core/agent-common-protocol.md` is the target file for M1.
**When** the run-phase reads the file and greps for `### Ledger Closure`.
**Then** the grep returns exactly 1 hit, and the subsection body contains all four clauses:
- (a) synthetic result / ledger-closing artifact on aborted `Agent()` delegation (REQ-LEDGER-001),
- (b) team-ac-verify.sh exit-2 `ledger_note` reference (REQ-LEDGER-002),
- (c) TeammateIdle exit-2 task closure — re-assign or close, not left open (REQ-LEDGER-003),
- (d) cross-references to book1 ch04, book1 ch07, and `session-handoff.md` Block 3-4 (REQ-LEDGER-006).

**Evidence command** (run-phase re-verification):
```bash
grep -n '### Ledger Closure' .claude/rules/moai/core/agent-common-protocol.md
# expect: 1 hit
awk '/### Ledger Closure/,/^### [A-Z]/' .claude/rules/moai/core/agent-common-protocol.md | \
  grep -E 'synthetic|ledger-closing|ledger_note|TeammateIdle|book1 ch04|book1 ch07|session-handoff'
# expect: ≥ 6 hits (one per clause + cross-refs)
```

### AC-LEDGER-002 — team-ac-verify.sh exit-2 path emits `ledger_note`

**Given** `.claude/hooks/moai/team-ac-verify.sh` is the target file for M2.
**When** the run-phase greps the hook for `ledger_note`.
**Then** the grep returns ≥ 1 hit, and the surrounding context shows the `ledger_note` field is emitted on the exit-2 (reject) path — NOT on the allow/dormant exit-0 paths.

**Evidence command**:
```bash
grep -n 'ledger_note' .claude/hooks/moai/team-ac-verify.sh
# expect: ≥ 1 hit
grep -n "exit 2" .claude/hooks/moai/team-ac-verify.sh
# expect: ≥ 1 hit (the reject path), and exit 2 still = reject (CON-1 unchanged)
```

**Negative verification (exit-code semantics unchanged)**:
```bash
# Smoke test: dormant path still exits 0 with decision: dormant (team mode off)
echo '{"task":{"metadata":{}}}' | bash .claude/hooks/moai/team-ac-verify.sh | grep -o '"decision":"[a-z]*"'
# expect: "decision":"dormant" (or "allow") — NOT "reject" — when team mode is off
```

### AC-LEDGER-004 — moai.md §8 Error Recovery banner has Interrupt Closure annotation

**Given** `.claude/output-styles/moai/moai.md` §8 Error Recovery banner is the target for M3.
**When** the run-phase greps §8 for "Interrupt Closure".
**Then** the grep returns ≥ 1 hit, the annotation is ≤ 3 lines, and the banner's A/B/C/D option structure is preserved (no structural rewrite per CON-2).

**Evidence command**:
```bash
grep -n 'Interrupt Closure' .claude/output-styles/moai/moai.md
# expect: ≥ 1 hit within the §8 Error Recovery banner block (between line ~587 and the next ###)
# Verify A/B/C/D preserved:
awk '/### Error Recovery/,/^### /' .claude/output-styles/moai/moai.md | grep -c 'A. Retry as-is.*B. Alt approach.*C. Pause.*D. Abort'
# expect: 1 (the options line is intact)
```

### AC-LEDGER-005 — Grep reproducibility (vocabulary introduced, gap closed)

**Given** the plan-time baseline (spec.md §A.1) showed zero hits for the phrase-targeted grep across `.claude/rules/moai/` and `.claude/output-styles/`.
**When** the run-phase re-runs the phrase-targeted grep after M1+M3 land.
**Then** the grep returns ≥ 1 hit for at least `ledger` (in §Ledger Closure) and ≥ 1 hit for `synthetic` (in §Ledger Closure), demonstrating the vocabulary is now in active doctrine.

**Evidence command** (phrase-targeted — the defensible signal per spec.md §A.3):
```bash
# Plan-time baseline (before M1/M3):
grep -rniE 'interruptBehavior|ledger.?clos|synthetic.*tool_result|dangling.*tool_use' \
  .claude/rules/moai/ .claude/output-styles/ 2>/dev/null | wc -l
# expect at plan-time: 0

# Run-time post-edit:
grep -rniE 'ledger.?clos|synthetic.*result|dangling.*tool_use|ledger-closing' \
  .claude/rules/moai/core/agent-common-protocol.md .claude/output-styles/moai/moai.md
# expect post-edit: ≥ 2 hits (one in §Ledger Closure, one in §8 annotation)
```

**Honest-baseline discipline**: if at run-time the bare word `ledger` is found in unrelated financial/accounting prose under `.claude/` (outside the two target files), the claim is NOT "zero hits anywhere" but "zero hits for the orchestration-interrupt sense of ledger closure in the two target surfaces before this SPEC". The AC uses the phrase-targeted grep on the two target files, not the bare-word grep across the whole tree.

### AC-LEDGER-006 — Scope boundary (collision-free with P0 SPEC)

**Given** the P0 sibling `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` owns the §Hook Invocation Surface Recovery-Signal Carve-Out in `agent-common-protocol.md`.
**When** the run-phase greps `agent-common-protocol.md` for both `### Ledger Closure` and the P0 SPEC's `Recovery-Signal` / `Carve-Out` markers.
**Then** the two additions live in distinct `### `-level (or distinct `## `-level) sections — `### Ledger Closure` is NOT inside the `### Hook Invocation Surface` subsection.

**Evidence command**:
```bash
# Locate both additions and verify distinct sections
grep -nE '^### Ledger Closure|^### Hook Invocation Surface|^## User Interaction Boundary' \
  .claude/rules/moai/core/agent-common-protocol.md
# expect: 3 distinct line numbers, with `### Ledger Closure` and `### Hook Invocation Surface`
#         as siblings under `## User Interaction Boundary` (NOT nested).

# If P0 SPEC has merged, also verify the Carve-Out marker is under Hook Invocation Surface:
grep -n 'Recovery-Signal\|Carve-Out' .claude/rules/moai/core/agent-common-protocol.md
# expect: hits are between `### Hook Invocation Surface` and the next `### `, NOT between
#         `### Ledger Closure` and its next `### `.
```

**Pass condition**: `### Ledger Closure` line number is NOT between the `### Hook Invocation Surface` line number and the next `### ` line number. (In other words: the two additions occupy disjoint `### ` subtrees.)

### AC-LEDGER-007 — book1 ch04 + ch07 cited

**Given** REQ-LEDGER-006 requires cross-references to book1 ch04 and ch07.
**When** the run-phase greps the §Ledger Closure subsection for the citations.
**Then** both `book1 ch04` (or `ch04` / "账本闭环") and `book1 ch07` (or `ch07` / "parent-abort") appear in the subsection body.

**Evidence command**:
```bash
awk '/### Ledger Closure/,/^### [A-Z]/' .claude/rules/moai/core/agent-common-protocol.md | \
  grep -E 'book1 ch04|ch04|账本闭环|book1 ch07|ch07|parent-abort'
# expect: ≥ 2 hits (one ch04, one ch07)
```

## §D.4 Edge Cases

| Edge case | Expected behavior | Covered by |
|-----------|-------------------|-----------|
| Agent() delegation times out (no user interrupt, no parent-abort) | REQ-LEDGER-001 covers timeout as an abort case — orchestrator emits ledger-closing artifact | AC-LEDGER-001 (clause a wording includes "timeout") |
| team-ac-verify.sh runs with team mode OFF | No change — dormant path exits 0 without `ledger_note` (the field is only on the reject path) | AC-LEDGER-002 negative verification |
| P0 sibling SPEC merged BEFORE this SPEC's M1 | AC-LEDGER-006 still passes — §Ledger Closure is placed in a distinct section regardless of merge order | AC-LEDGER-006 |
| P0 sibling SPEC NOT yet merged at M1 time | `grep Recovery-Signal` returns 0 hits — AC-LEDGER-006 pass condition is vacuously satisfied for the P0-side check; the §Ledger Closure placement is still verified as not-inside-§Hook-Invocation-Surface | AC-LEDGER-006 |
| Orchestrator aborted mid-AskUserQuestion round (not mid-Agent()) | Out of scope — REQ-LEDGER-001 covers Agent() delegation abort, not AskUserQuestion round abort (AskUserQuestion is synchronous from the orchestrator's turn perspective) | §X.5 (subagent-side) + REQ-LEDGER-001 scope (Agent() delegation) |
| Bare word `ledger` found in unrelated prose at run-time | AC-LEDGER-005 uses the phrase-targeted grep on the two target files, not the bare-word tree grep — claim is honestly narrowed | AC-LEDGER-005 honest-baseline discipline |

## §D.5 Quality Gate Criteria

- `moai spec lint .moai/specs/SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001/spec.md` exits 0.
- `moai spec audit --json` for this SPEC reports `era: V3R6`, no MUST-FIX drift findings.
- AC↔REQ coverage = 100% (verified in §D.2).
- All 6 MUST-PASS ACs (AC-LEDGER-001, 002, 004, 005, 006, 007) independently re-verified with live grep/read evidence at M4.

## §D.6 Closure Gates (4-phase close)

- [ ] M1 complete — §Ledger Closure added (AC-LEDGER-001, AC-LEDGER-006, AC-LEDGER-007).
- [ ] M2 complete — team-ac-verify.sh `ledger_note` added, exit-code semantics unchanged (AC-LEDGER-002).
- [ ] M3 complete — moai.md §8 annotation added (AC-LEDGER-004).
- [ ] M4 complete — lint clean + grep reproducibility + scope-boundary (AC-LEDGER-005, AC-LEDGER-006).
- [ ] progress.md §E.2 (run-phase evidence) populated by manager-develop.
- [ ] progress.md §E.4 (sync-phase audit-ready signal) + §E.5 (Mx-phase) populated by manager-docs.

## §D.7 Forward-Looking Checks (post-close)

- A follow-up SPEC should add **full AC verification logic** to team-ac-verify.sh (parsing acceptance.md, running evidence commands, emitting exit 2 based on AC failure). Tracked in spec.md §X.1.
- A follow-up SPEC may introduce a **structured ledger artifact** (JSON schema, `.moai/state/ledger.json`). Tracked in spec.md §X.3.
- The §Ledger Closure subsection should be reviewed after the P0 sibling SPEC merges to confirm the two additions remain in distinct sections under load (post-merge AC-LEDGER-006 re-verification).

## §D.8 Indirect Verification

The grep-based evidence commands above are **direct** verification (they read the edited files and assert substring presence). Indirect verification — running the orchestrator under a live abort scenario and observing the ledger-closing artifact in the transcript — is out of scope for this SPEC's run-phase (the artifact is a prose summary, not a runtime data structure; §X.3 excludes the structured form). A future SPEC that introduces a structured artifact would add indirect verification.
