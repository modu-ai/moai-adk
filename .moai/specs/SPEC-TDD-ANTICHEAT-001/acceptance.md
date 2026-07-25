# Acceptance Criteria — SPEC-TDD-ANTICHEAT-001

Every criterion is observable. Commands are illustrative; the run-phase agent captures verbatim output as evidence.

## §D Acceptance Matrix

| AC | Requirement | Gate |
|----|-------------|------|
| AC-TDD-001 | skill HARD section + agent + §E additions present | must-pass |
| AC-TDD-002 | dual-write byte-parity (3 pairs identical) | must-pass |
| AC-TDD-003 | `make build` succeeds | must-pass |
| AC-TDD-004 | `§E` E1-E7 preserved, exactly one item added | must-pass |
| AC-TDD-005 | neutrality (CI guard + targeted grep) | must-pass |
| AC-TDD-006 | simplicity bound (only 6 files changed) | must-pass |
| AC-TDD-007 | falsifiability contrast (RED-skip → incomplete matrix) | must-pass |
| AC-TDD-008 | agent Forbidden + STEP 2 invariant present | must-pass |

---

## AC-TDD-001 — Content additions present (REQ-TDD-001..004, 006, 007)

**Given** the three operational files after the change,
**When** their new content is inspected,
**Then** all of the following hold:

- `.claude/skills/moai-workflow-tdd/SKILL.md` contains a HARD "Test-First Anti-Cheat" section stating both invariants: (i) the RED failure output must be observed and shown as evidence; (ii) implementation code written before its failing test is deleted and re-derived test-first.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` `§E` contains one new self-verification item requiring the verbatim RED failing-test output captured before GREEN.
- `.claude/agents/moai/manager-develop.md` contains both the new Forbidden entry and the STEP 2 RED-evidence invariant.

```bash
grep -q "Test-First Anti-Cheat" .claude/skills/moai-workflow-tdd/SKILL.md
grep -qi "RED failure output" .claude/skills/moai-workflow-tdd/SKILL.md
grep -qi "delete" .claude/skills/moai-workflow-tdd/SKILL.md          # invariant ii (deleted and re-derived)
grep -qi "before GREEN" .claude/rules/moai/development/manager-develop-prompt-template.md
```

## AC-TDD-002 — Dual-write byte-parity (REQ-TDD-008, C2)

**Given** the six files after the dual-write,
**When** each operational copy is diffed against its template mirror,
**Then** all three pairs are byte-identical (the new prose is neutral, so operational and mirror stay identical).

```bash
diff -q .claude/skills/moai-workflow-tdd/SKILL.md \
        internal/template/templates/.claude/skills/moai-workflow-tdd/SKILL.md
diff -q .claude/rules/moai/development/manager-develop-prompt-template.md \
        internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
diff -q .claude/agents/moai/manager-develop.md \
        internal/template/templates/.claude/agents/moai/manager-develop.md
# Expected: no output from any diff (identical)
```

## AC-TDD-003 — Build succeeds (REQ-TDD-008, C2)

**Given** the edited template mirrors,
**When** `make build` runs,
**Then** it exits 0 (the binary recompiles with the embedded template FS).

```bash
make build; echo "exit=$?"   # expected exit=0
```

## AC-TDD-004 — §E additive-only (REQ-TDD-005, C1)

**Given** the changed `§E` section,
**When** the E-item headings are counted and E1-E7 are inspected,
**Then** the E1-E7 heading lines are byte-preserved and exactly one new item (E8) is added.

```bash
grep -cE '^\*\*E[0-9]+\.' .claude/rules/moai/development/manager-develop-prompt-template.md
# Baseline count is 7 (E1-E7); expected after change: 8
git diff .claude/rules/moai/development/manager-develop-prompt-template.md \
  | grep -E '^\-' | grep -E 'E[1-7]\.' || echo "no E1-E7 lines removed (additive-only)"
```

## AC-TDD-005 — Neutrality (REQ-TDD-008, C3)

**Given** the three template-mirror copies after the change,
**When** the neutrality CI guard runs AND a targeted grep for this SPEC's internal tokens runs,
**Then** the CI guard passes AND the targeted grep returns zero matches.

> NOTE (false-positive trap): do NOT grep the mirrors for generic `SPEC-`/`REQ-`/`AC-` patterns — the deployed templates already contain CI-tolerated pedagogical placeholders (`AC-XXX-001`, `C-HRA-008`). Neutrality here means the change introduced no *internal* markers. Verify with the authoritative CI guard plus a grep scoped to THIS SPEC's tokens.

```bash
go test ./internal/template/ -run TestTemplateNoInternalContentLeak   # authoritative guard → PASS
# Targeted grep: this SPEC's internal markers must not appear in any mirror copy
grep -rEn "SPEC-TDD-ANTICHEAT-001|REQ-TDD-|AC-TDD-|2026-07|[0-9a-f]{7,40}\b" \
  internal/template/templates/.claude/skills/moai-workflow-tdd/SKILL.md \
  internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md \
  internal/template/templates/.claude/agents/moai/manager-develop.md \
  || echo "clean — 0 internal markers in mirror copies"
```

## AC-TDD-006 — Simplicity bound (C1, C5)

**Given** the run-phase change set,
**When** the working tree is inspected,
**Then** only the six named files (plus build artifacts) are modified — no new file is created in the harness scope.

```bash
git status --porcelain -- .claude internal/template/templates/.claude
# Expected: exactly the 6 target files as modified (M), zero added (A/??) harness files
```

## AC-TDD-007 — Falsifiability contrast (REQ-TDD-008)

**Given** the pre-change `§E` (E1-E7) versus the post-change `§E` (E1-E7 + new RED-evidence item),
**When** a test-after run (RED skipped) attempts to complete the self-verification matrix,
**Then**:

- **Before** the change: a fully-passing matrix is producible with NO RED evidence — test-first is unfalsifiable (a test-after run yields an identical clean matrix).
- **After** the change: the matrix contains a mandatory field for the verbatim pre-GREEN RED failing output. A run that skipped RED has no such output to supply, so the matrix cannot be reported as clean — test-first is now falsifiable.

Verification (structural — the new item is a mandatory deliverable requiring pre-GREEN failing output):

```bash
# The new §E item exists, is phrased as mandatory (MUST), and requires verbatim pre-GREEN RED output
grep -A6 -iE 'RED (failure|failing).*(evidence|output)' \
  .claude/rules/moai/development/manager-develop-prompt-template.md \
  | grep -qi 'MUST\|verbatim\|before GREEN'
```

## AC-TDD-008 — Agent Forbidden + STEP 2 invariant (REQ-TDD-006, REQ-TDD-007)

**Given** `.claude/agents/moai/manager-develop.md` after the change,
**When** the Behavioral Contract Forbidden list and TDD Cycle STEP 2 are inspected,
**Then** the Forbidden list includes "writing implementation before its failing test" and STEP 2 carries the RED-evidence-plus-delete-pre-test-code invariant.

```bash
grep -iE 'implementation before its failing test' .claude/agents/moai/manager-develop.md
grep -A10 'STEP 2' .claude/agents/moai/manager-develop.md | grep -iE 'RED|delete|before its'
```

---

## Definition of Done

- [ ] AC-TDD-001..008 all PASS (must-pass gates)
- [ ] Dual-write parity: 3 pairs byte-identical
- [ ] `make build` exit 0
- [ ] Neutrality CI guard `TestTemplateNoInternalContentLeak` PASS + 0 internal markers in mirrors
- [ ] `§E` additive-only: E1-E7 preserved, exactly one item added
- [ ] Only 6 files changed; no new harness file created
- [ ] Falsifiability: post-change matrix cannot be completed clean without pre-GREEN RED evidence
