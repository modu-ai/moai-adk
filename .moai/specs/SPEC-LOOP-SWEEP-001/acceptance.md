# acceptance.md — SPEC-LOOP-SWEEP-001

> Each AC is a single discriminating check. Doctrine-reconciliation ACs use
> per-file grep (one file per AC, not a vacuous multi-file OR). Template-mirror ACs
> use per-file checks. `/moai fix` + `/moai review` behavior-unchanged ACs assert
> the ABSENCE of behavior edits.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Baseline → Post |
|----|-----|----------|-----------------|
| AC-LSW-001 | REQ-LSW-001 | loop.md defines loop as goal preset + condition | absent → present |
| AC-LSW-002 | REQ-LSW-002 | scan builds finite queue; default + `--lens` | absent → present |
| AC-LSW-003 | REQ-LSW-003 | no-invented-improvements + empty-queue exit | 0 → ≥1 |
| AC-LSW-004 | REQ-LSW-004 | ceiling precedence + verdict + Step 1.5 preserved | present → present |
| AC-LSW-005 | REQ-LSW-005 | extended `exit_kind` for sweep | absent → present |
| AC-LSW-006 | REQ-LSW-006 | loop consumes review lenses; review read-only | 0 → ≥1 |
| AC-LSW-007 | REQ-LSW-007 | review.md↔loop.md cross-ref; no review behavior change | present + unchanged |
| AC-LSW-008 | REQ-LSW-008 | fix unchanged; residue text updated | text-only |
| AC-LSW-009 | REQ-LSW-009 | four-quadrant re-expressed (goal engine + presets) | present |
| AC-LSW-010 | REQ-LSW-010 | goal-directive.md loop preset row | absent → present |
| AC-LSW-011 | REQ-LSW-011 | cadence-bridge loop-NOT-eligible statement | present |
| AC-LSW-012 | REQ-LSW-012 | spec-workflow loop row + alias decision | updated |
| AC-LSW-013 | REQ-LSW-013 | Go loop help-text renamed; behavior unchanged | text + tests |
| AC-LSW-014 | REQ-LSW-014 | mirrors + neutral + make build | per-file + exit 0 |

### AC-LSW-001 — loop = goal preset + condition

```bash
grep -in "goal preset\|preset" .claude/skills/moai/workflows/loop.md | head
grep -in "queue drained\|diagnostics clean" .claude/skills/moai/workflows/loop.md | head
```
PASS when loop.md defines `/moai loop` as a goal preset with the "issue queue
drained + diagnostics clean" completion condition.

### AC-LSW-002 — finite scan queue + lenses

```bash
grep -in "finite\|issue queue\|--lens" .claude/skills/moai/workflows/loop.md | head
grep -in "LSP\|lint\|test failures\|security\|@MX" .claude/skills/moai/workflows/loop.md | head
```
PASS when loop.md documents a finite scan queue from default lenses (LSP, lint,
test failures, review lenses security + `@MX`) plus opt-in
`--lens clean|simplify|coverage`.

### AC-LSW-003 — no-invented-improvements + empty-queue exit

```bash
grep -in "no invented\|outside the queue\|empty queue" .claude/skills/moai/workflows/loop.md | head
```
PASS when loop.md states no work outside the scanned queue AND an empty queue
causes immediate exit.

### AC-LSW-004 — ceiling/verdict/Step1.5 preserved

```bash
grep -c "max_iterations" .claude/skills/moai/workflows/loop.md   # precedence retained
grep -in "5-section\|Step 1.5\|ceiling-exit verdict\|loop-verdict" .claude/skills/moai/workflows/loop.md | head
```
PASS when the ceiling precedence rule, the 5-section ceiling-exit verdict, the
`loop-verdict-<id>.json` persistence, and the Step 1.5 independent final pass are
all still present after the rewrite.

### AC-LSW-005 — extended exit_kind

```bash
grep -in "sweep-residue\|exit_kind" .claude/skills/moai/workflows/loop.md | head
```
PASS when loop.md names the sweep-preset `exit_kind` value (default proposed
`sweep-residue`) added additively to the base enum
(`ceiling | manual-residue | one-shot-residue`).

### AC-LSW-006 — loop consumes review lenses; review read-only

```bash
grep -in "review lens\|consumes review\|read-only" .claude/skills/moai/workflows/loop.md | head
```
PASS when loop.md documents that loop consumes review lenses as queue suppliers
AND standalone `/moai review` stays read-only/report-only.

### AC-LSW-007 — review.md cross-ref, no behavior change

```bash
grep -in "loop\|queue supplier\|layered" .claude/skills/moai/workflows/review.md | head
# Behavior-unchanged assertion: review.md contract sections are not edited.
git diff --stat .claude/skills/moai/workflows/review.md   # only the cross-ref para changed
```
PASS when review.md gains a 1-2 paragraph cross-ref AND the diff shows only the
cross-ref addition (no lens/pipeline/behavior lines changed).

### AC-LSW-008 — fix unchanged; residue text updated

```bash
# fix.md pipeline contract unchanged (Agentless, no-LLM-control-flow lines intact):
grep -c "No LLM-driven control flow\|Agentless fixed-pipeline" .claude/skills/moai/workflows/fix.md   # ≥2 (unchanged)
# residue handoff text updated to state residue enters the loop queue:
grep -in "enters the loop queue\|/moai loop" .claude/skills/moai/workflows/fix.md | head
```
PASS when the fix.md pipeline-contract lines are intact AND the residue-handoff
text is updated to say residue enters the loop queue.

### AC-LSW-009 — four-quadrant re-expressed

```bash
grep -in "goal engine\|preset\|quadrant" .claude/skills/moai/workflows/loop.md | head
grep -in "quadrant" .claude/skills/moai/workflows/fix.md | head
```
PASS when the quadrant sibling notes in loop.md and fix.md consistently frame the
taxonomy as "goal engine + presets."

### AC-LSW-010 — goal-directive.md loop preset row

```bash
grep -in "/moai loop" .claude/rules/moai/workflow/goal-directive.md | head
```
PASS when goal-directive.md carries a `/moai loop` row describing it as a goal
preset (distinct from native `/goal` and `/moai goal`).

### AC-LSW-011 — cadence loop-NOT-eligible

```bash
grep -in "loop" .claude/rules/moai/workflow/cadence-bridge.md | grep -in "not\|ineligible\|read-only" | head
```
PASS when cadence-bridge.md continues to state `/moai loop` is NOT cadence-eligible
(and loop is NOT added to the recipe catalog).

### AC-LSW-012 — spec-workflow loop row + alias decision

```bash
grep -in "loop\|--mode loop\|Multi-Agent" .claude/rules/moai/workflow/spec-workflow.md | head
```
PASS when the Subcommand Classification loop row is updated AND the
`run --mode loop` alias disposition (keep, per plan § Open Decisions default) is
stated with justification.

### AC-LSW-013 — Go loop help text renamed, behavior unchanged

```bash
grep -in "SPEC-lifecycle\|lifecycle controller" internal/cli/loop.go | head
go test ./internal/loop/... ./internal/ralph/... 2>&1 | tail -3   # green (behavior unchanged)
```
PASS when the `internal/cli/loop.go` help text clarifies it is the SPEC-lifecycle
controller AND the loop Go tests still pass (string-only change).

### AC-LSW-014 — mirrors + neutrality + build

```bash
test -f internal/template/templates/.claude/skills/moai/workflows/loop.md && echo MIRROR_OK
grep -rn "SPEC-LOOP-SWEEP\|SPEC-GOAL-ENGINE\|AGENTIC-CORE\|REQ-LSW" internal/template/templates/.claude/ | wc -l   # expect 0
make build ; echo "exit=$?"
```
PASS when every changed `.claude/` file has a mirror, the neutrality grep is 0,
and `make build` exits 0.

## §D.1 Definition of Done

- All 14 ACs PASS.
- `/moai fix` + `/moai review` behavior unchanged (text-only edits verified by diff).
- `internal/loop`/`internal/ralph` logic untouched; loop Go tests green.
- loop stays NOT cadence-eligible.
- All template mirrors updated + neutral + `make build` green.

## §D.2 Edge cases

- **Empty scan queue**: loop exits immediately (AC-LSW-003) — no ceiling burn.
- **`--lens coverage` with coverage disabled**: coverage lens is opt-in; when the
  project has no coverage gate, the lens supplies no queue items (documented).
- **Review lens finds issues loop cannot auto-fix**: those enter the queue as
  manual-residue at ceiling exit (existing `manual-residue` path, unchanged).
- **Depends_on unmet**: run-phase entry blocked unless GOAL-ENGINE is `completed`
  or `--ignore-deps` + logged rationale (spec-workflow Depends_on pre-flight).
- **`run --mode loop` alias**: kept (default decision); both routes resolve to the
  sweep preset (AC-LSW-012).
