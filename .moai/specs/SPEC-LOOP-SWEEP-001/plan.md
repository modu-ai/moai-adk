# plan.md — SPEC-LOOP-SWEEP-001

> Tier M. Docs-heavy (no new Go engine). Epic AGENTIC-CORE, SPEC 3 of 3.
> depends_on SPEC-GOAL-ENGINE-001. Shared findings:
> `../SPEC-ANALYZE-FIRST-ROUTING-001/research.md`.

## §A — Context

- **Work location**: repo root `/Users/goos/MoAI/moai-adk-go`.
- **Affected surfaces** (docs/config; one Go help-text string):
  - `.claude/skills/moai/workflows/loop.md` (rewrite → goal preset).
  - `.claude/skills/moai/workflows/fix.md` (residue-handoff TEXT only).
  - `.claude/skills/moai/workflows/review.md` (1-2 para cross-ref only).
  - `.claude/rules/moai/workflow/goal-directive.md` (table row).
  - `.claude/rules/moai/workflow/cadence-bridge.md` (loop-NOT-eligible statement).
  - `.claude/rules/moai/workflow/spec-workflow.md` (Subcommand Classification loop row).
  - `CLAUDE.md:46` (wording).
  - `.claude/skills/moai-workflow-loop/SKILL.md` (preset architecture).
  - `internal/cli/loop.go` (help-text string rename ONLY — no behavior change).
  - `internal/template/templates/**` mirrors.
- **PRESERVE**: `/moai fix` pipeline contract; `/moai review` behavior;
  `internal/loop`/`internal/ralph` logic; the `loop-verdict-<id>.json` base
  `exit_kind` enum owner (extend, don't reassign); ceiling precedence + Step 1.5.

## §A.5 PRESERVE / EXTEND list

| Path | Disposition |
|------|-------------|
| `loop.md` | REWRITE (goal preset; PRESERVE ceiling/verdict/Step1.5) |
| `fix.md` residue-handoff text | EXTEND (text only) |
| `review.md` | EXTEND (cross-ref para only, no behavior change) |
| `goal-directive.md` | EXTEND (table row) |
| `cadence-bridge.md` | EXTEND (loop-NOT-eligible statement) |
| `spec-workflow.md` Subcommand Classification | EDIT (loop row + alias decision) |
| `CLAUDE.md:46` | EDIT (wording) |
| `moai-workflow-loop/SKILL.md` | EXTEND (preset architecture) |
| `internal/cli/loop.go` | EDIT (help-text string ONLY) |
| `internal/loop`, `internal/ralph` | PRESERVE (logic untouched) |
| `internal/template/templates/**` | MIRROR (D5) |

## §B — Known Issues (filtered, docs-heavy Tier M)

- **B6 — spec-lint heading**: this spec.md uses `### §D.1 Out of Scope — <topic>`
  H3 sub-headings (satisfies `OutOfScopeRule`).
- **B2 — cross-SPEC conflict**: do NOT change `/moai fix`/`/moai review` behavior;
  do NOT reassign the `loop-verdict-<id>.json` base `exit_kind` enum owner
  (loop.md owns `ceiling|manual-residue`; fix.md added `one-shot-residue`; this
  SPEC adds the sweep value ADDITIVELY).
- **B8/B10 — hygiene/scope**: help-text rename in `internal/cli/loop.go` must be a
  STRING change only; run the loop Go tests to confirm no behavior regression.
- **Cadence coordination (research.md §D.3)**: REQ-LSW-011 keeps loop
  cadence-ineligible — do NOT add loop to `cadence-bridge.md`'s recipe catalog.
- **Template neutrality (§25)**: mirrors carry no internal SPEC ID.

## §C — Pre-flight (run before editing)

```bash
# 1. Confirm loop.md anchors (goal-based quadrant, ceiling precedence, verdict)
grep -n "goal-based\|ceiling\|loop-verdict\|Step 1.5\|max_iterations" .claude/skills/moai/workflows/loop.md | head
# 2. Confirm fix.md residue-handoff block (text-only edit target)
grep -n "one-shot-residue\|/moai loop\|residue" .claude/skills/moai/workflows/fix.md | head
# 3. Confirm review.md is read-only (no behavior change target)
grep -n "read-only\|report" .claude/skills/moai/workflows/review.md | head
# 4. Confirm spec-workflow loop row + run --mode loop alias
grep -n "loop\|--mode loop\|Multi-Agent" .claude/rules/moai/workflow/spec-workflow.md | head
# 5. Confirm cadence-bridge loop eligibility statement
grep -n "loop\|cadence-eligible\|read-only" .claude/rules/moai/workflow/cadence-bridge.md | head
# 6. Confirm Go loop help text string + tests
grep -n "Short:\|Long:\|Use:" internal/cli/loop.go | head
go test ./internal/loop/... ./internal/ralph/... 2>&1 | tail -3
# 7. Confirm the depends_on target status
grep -n "^status:" .moai/specs/SPEC-GOAL-ENGINE-001/spec.md
```

## §D — Constraints

- Loop is a PRESET on the `SPEC-GOAL-ENGINE-001` engine — build NO new engine code.
- `/moai fix` and `/moai review` behavior UNCHANGED (text-only edits).
- PRESERVE ceiling precedence, 5-section verdict, Step 1.5, memory-pressure guard.
- Extend `exit_kind` ADDITIVELY (do not reassign the base enum owner).
- `internal/cli/loop.go` change is a help-text STRING only (loop Go tests green).
- `internal/loop/` and `internal/ralph/` MUST remain untouched — `git diff --exit-code
  --stat internal/loop/ internal/ralph/` exits 0 (empty diff) post-run-phase (AC-LSW-013b).
  This is the "build NO new engine code" boundary guard — tests-green alone is
  insufficient (a behavior-preserving refactor could pass tests while crossing the
  no-new-engine-code line).
- loop stays NOT cadence-eligible (REQ-LSW-011).
- Template-First mirrors + §25 neutrality + `make build`.

## §E — Self-Verification (plan-phase audit-ready)

Run-phase completion verified by `acceptance.md`. Plan-phase audit-ready recorded
in `progress.md` §E.1.

## §F — Milestones (priority-ordered)

- **M1 — loop.md rewrite (D1)**: loop = goal preset; scan → finite queue; default
  + `--lens` opt-in; no-invented-improvements boundary; empty-queue exit; PRESERVE
  ceiling/verdict/Step1.5; extend `exit_kind`. Priority High.
- **M2 — review + fix relationship (D2, D3)**: loop.md consumes review lenses;
  review.md cross-ref paragraph (no behavior change); fix.md residue-handoff text
  update. Priority High.
- **M3 — doctrine reconciliation (D4)**: four-quadrant taxonomy re-expression;
  goal-directive.md row; cadence-bridge loop-ineligible; spec-workflow loop row +
  `run --mode loop` alias decision; CLAUDE.md:46 wording; moai-workflow-loop
  SKILL.md preset architecture. Priority Medium.
- **M4 — Go loop help-text rename (D4)**: `internal/cli/loop.go` help string only;
  loop Go tests green. Priority Medium.
- **M5 — Template mirror + build (D5)**: mirror all changed `.claude/` files;
  `make build`; neutrality grep. Priority High (gate).

Ordering: M1 → M2 → M3 → M4 → M5.

## §G — Anti-Patterns

- Building a new loop engine (loop is a PRESET on GOAL-ENGINE — §D.1).
- Changing `/moai fix`/`/moai review` behavior (text-only per REQ-LSW-007/008).
- Reassigning the `loop-verdict-<id>.json` base `exit_kind` enum owner.
- Adding loop to cadence recipes (REQ-LSW-011 keeps it ineligible).
- Changing `internal/loop`/`internal/ralph` LOGIC (help-text string only). The
  empty-diff guard `git diff --exit-code --stat internal/loop/ internal/ralph/`
  (AC-LSW-013b) MUST exit 0 — even a behavior-preserving refactor inside these
  directories crosses the "no new engine code" boundary.
- Vacuous doctrine-edit claim — per-file grep for each reconciliation surface.

## §H — Cross-References

- Shared `research.md` §C.1 (two loop surfaces), §C.2 (fix), §C.3 (quadrants),
  §D.3 (cadence).
- `loop.md` (goal-based quadrant, ceiling precedence, 5-section verdict, Step 1.5).
- `fix.md` (`one-shot-residue`, residue-recommends-loop text).
- `goal-directive.md` (autonomous-continuation comparison table).
- `spec-workflow.md` § Subcommand Classification (loop = Multi-Agent alias).
- `SPEC-GOAL-ENGINE-001` (the engine loop presets on).

## § Deferred (NOT run-phase scope)

- **Go `moai loop` / goal-engine unification** (research.md §C.1). This SPEC only
  renames help text + documents the two surfaces. Follow-up SPEC.
- **docs-site 4-locale** loop documentation. Follow-up SPEC.
- **Full `--lens` catalog beyond clean|simplify|coverage**. Follow-up.

## § Settled Decisions (iteration-2 — clarifications resolved via AskUserQuestion)

- **DECISION (`run --mode loop` alias)** — RESOLVED: **KEEP** the alias
  (backward-compat; `/moai run --mode loop` is a historical entry point and
  retiring it is a breaking change). Both `/moai run --mode loop` and `/moai loop`
  resolve to the goal-preset sweep. The `spec-workflow.md` § Subcommand
  Classification loop row is updated to state this, with the justification. Folded
  into REQ-LSW-012 as settled.
- **DECISION (sweep `exit_kind` value name)** — RESOLVED: `"sweep-residue"`
  (parallel to `one-shot-residue`), added **ADDITIVELY** to the base
  `ceiling | manual-residue | one-shot-residue` enum. Downstream consumers of
  `loop-verdict-<id>.json` gain a fourth valid `exit_kind` value; the base enum
  owner is unchanged. Folded into REQ-LSW-005 as settled.
