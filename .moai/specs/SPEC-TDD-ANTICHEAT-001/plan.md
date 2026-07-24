# Implementation Plan — SPEC-TDD-ANTICHEAT-001

> Milestones are ordered by decision-reversibility: the highest-change-likelihood decision (the exact **invariant wording**) leads, so human review focuses there; the mechanical mirror/build/verify steps follow.

## §A Context

Plan-phase artifact for a Tier S harness-prose change. Three files each gain a bounded addition; every addition is a dual-write to the operational copy plus its template mirror, then `make build`. The three operational copies are currently byte-identical to their mirrors, and the new prose is neutral, so the copies remain byte-identical after the change (parity is checkable).

## §B Known Issues / Preconditions

- The operational and mirror copies of all three files are byte-identical at baseline (verified).
- The neutrality guard `TestTemplateNoInternalContentLeak` tolerates pre-existing pedagogical placeholders (`AC-XXX-001`, `C-HRA-008`); the neutrality check for this change must target THIS SPEC's specific internal tokens, not generic `SPEC-`/`REQ-` patterns (see acceptance.md AC-TDD-005 note).

## §C Pre-flight

```bash
# Confirm baseline byte-parity of the 3 file pairs before editing
diff -q .claude/skills/moai-workflow-tdd/SKILL.md \
        internal/template/templates/.claude/skills/moai-workflow-tdd/SKILL.md
diff -q .claude/rules/moai/development/manager-develop-prompt-template.md \
        internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
diff -q .claude/agents/moai/manager-develop.md \
        internal/template/templates/.claude/agents/moai/manager-develop.md
# Confirm neutrality guard is green at baseline
go test ./internal/template/ -run TestTemplateNoInternalContentLeak
```

## §D Constraints

Inherit §C of spec.md (C1 Tier S minimal · C2 dual-write · C3 neutrality · C4 English · C5 simplicity bound). Do NOT restructure E1-E7. Do NOT create any new file. Cycle type: `autofix` (prose edits, no Go code / no tests to author).

## §E Self-Verification

Run-phase completion is verified against acceptance.md AC-TDD-001..008. The falsifiability contrast (AC-TDD-007) and the dual-write parity check (AC-TDD-002) are the two must-pass gates.

## §F Milestones

### M1 — Invariant wording (decision-heavy — review focus)

The exact text of the three additions is the highest-change-likelihood decision. Finalize and apply to the three **operational** `.claude/...` copies.

1. **`.claude/skills/moai-workflow-tdd/SKILL.md`** — add a HARD **"Test-First Anti-Cheat"** section (placed after the Verification checklist, promoting the Red Flags / Verification content into two enforced invariants):
   - Invariant i: the RED failure output MUST be observed and shown as completion evidence (verbatim failing-test output captured before GREEN).
   - Invariant ii: any implementation code written before its failing test MUST be deleted and re-derived test-first.
   - Generic mechanism prose only — no internal identifiers.
2. **`.claude/rules/moai/development/manager-develop-prompt-template.md` `§E`** — add exactly one new self-verification item (an **E8**) after E7, requiring the verbatim RED failing-test output captured before GREEN. E1-E7 untouched.
3. **`.claude/agents/moai/manager-develop.md`** —
   - Behavioral Contract **Forbidden** (~L86): append "writing implementation before its failing test".
   - TDD Cycle **STEP 2** (~L169-173): add the RED-evidence-plus-delete-pre-test-code invariant sub-line.

### M2 — Template-mirror dual-write + build (mechanical)

4. Mirror the identical neutral text from M1 into the three template-source copies:
   - `internal/template/templates/.claude/skills/moai-workflow-tdd/SKILL.md`
   - `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md`
   - `internal/template/templates/.claude/agents/moai/manager-develop.md`
5. `make build` (recompiles the binary with embedded templates via `//go:embed all:templates`).

> M1 + M2 together constitute the mandatory per-file dual-write. Each file's operational and mirror copies carry byte-identical new text.

### M3 — Verification (mechanical)

6. Dual-write parity: `diff -q` each of the 3 pairs → identical.
7. Neutrality: `go test ./internal/template/ -run TestTemplateNoInternalContentLeak` → PASS; plus targeted grep of the 3 mirror copies for this SPEC's internal tokens → 0 matches.
8. Falsifiability contrast + simplicity bound per acceptance.md.

## §G Anti-Patterns

- Naive neutrality grep of `SPEC-`/`REQ-`/`AC-` (false-positives on pre-existing pedagogical placeholders already in the CI-clean templates). Target THIS SPEC's specific tokens instead.
- Editing only the operational copy and forgetting the mirror (breaks Template-First; the next `moai update` would overwrite the operational change).
- Restructuring E1-E7 while adding the new item (violates C1/C5 — additive only).
- Building a hook/lint detector for RED (out of scope — this SPEC makes test-first falsifiable via the matrix, not mechanically enforced).

## §H File Impact Map (6 files + build)

| # | File | Copy | Change |
|---|------|------|--------|
| 1 | `.claude/skills/moai-workflow-tdd/SKILL.md` | operational | + HARD Test-First Anti-Cheat section (2 invariants) |
| 2 | `internal/template/templates/.claude/skills/moai-workflow-tdd/SKILL.md` | mirror | identical to #1 |
| 3 | `.claude/rules/moai/development/manager-develop-prompt-template.md` | operational | + one new `§E` item (E8), E1-E7 unchanged |
| 4 | `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` | mirror | identical to #3 |
| 5 | `.claude/agents/moai/manager-develop.md` | operational | Forbidden += entry; STEP 2 += RED-evidence/delete invariant |
| 6 | `internal/template/templates/.claude/agents/moai/manager-develop.md` | mirror | identical to #5 |
| — | `make build` | — | recompile binary with embedded template FS |
