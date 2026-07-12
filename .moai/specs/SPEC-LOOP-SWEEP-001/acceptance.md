# acceptance.md — SPEC-LOOP-SWEEP-001

> [HARD] Each AC is a **single discriminating check** — NO vacuous compound OR
> (`A|B|C ≥ N` where one item satisfies N alone). Every `0 → ≥1` AC had its
> baseline verified as **0** on the current tree this iteration (recorded inline).
> **Preservation ACs** (present → present) are explicitly labelled as such — they
> assert an existing invariant is NOT regressed, and are distinct from
> discriminating `0 → ≥1` ACs. Doctrine-reconciliation ACs use per-file grep (one
> file per AC). Template-mirror ACs use per-file checks.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Kind / Baseline → Post |
|----|-----|----------|------------------------|
| AC-LSW-001 | REQ-LSW-001 | loop = goal preset + condition | discriminating: 0 → ≥1 |
| AC-LSW-002 | REQ-LSW-002 | existing scan lenses preserved | preservation: present |
| AC-LSW-002b | REQ-LSW-002 | NEW review-lens additions to the queue | discriminating: 0 → ≥1 |
| AC-LSW-003 | REQ-LSW-003 | no-invented AND empty-queue-exit (conjunctive) | discriminating: both 0 → ≥1 |
| AC-LSW-004 | REQ-LSW-004 | ceiling precedence preserved | preservation: present |
| AC-LSW-004b | REQ-LSW-004 | verdict + Step 1.5 preserved (split checks) | preservation: present |
| AC-LSW-005 | REQ-LSW-005 | `sweep-residue` additive exit_kind | discriminating: 0 → ≥1 |
| AC-LSW-006 | REQ-LSW-006 | loop consumes review lenses (loop.md side) | discriminating: 0 → ≥1 |
| AC-LSW-007 | REQ-LSW-007 | review.md cross-ref; commit-scoped diff | discriminating: 0 → ≥1 + scoped |
| AC-LSW-008 | REQ-LSW-008 | fix pipeline contract intact | preservation: present |
| AC-LSW-008b | REQ-LSW-008 | fix residue "enters the loop queue" | discriminating: 0 → ≥1 |
| AC-LSW-009 | REQ-LSW-009 | loop.md "goal engine + preset" framing | discriminating: 0 → ≥1 |
| AC-LSW-009b | REQ-LSW-009 | fix.md "goal engine + preset" framing | discriminating: 0 → ≥1 |
| AC-LSW-010 | REQ-LSW-010 | goal-directive.md loop row names goal preset | discriminating: 0 → ≥1 |
| AC-LSW-011 | REQ-LSW-011 | cadence loop-NOT-eligible preserved | preservation: present |
| AC-LSW-012 | REQ-LSW-012 | NEW alias-KEEP decision + justification | discriminating: 0 → ≥1 |
| AC-LSW-013 | REQ-LSW-013 | Go help "SPEC-lifecycle" + tests green | discriminating: 0 → ≥1 + test |
| AC-LSW-014 | REQ-LSW-014 | per-file mirror parity + neutral + build | per-file + exit 0 |

### AC-LSW-001 — loop = goal preset + condition

```bash
grep -ic "goal preset" .claude/skills/moai/workflows/loop.md                       # expect ≥1 (baseline 0)
grep -ic "queue drained\|diagnostics clean" .claude/skills/moai/workflows/loop.md  # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): both 0. PASS when loop.md defines `/moai loop`
as a **goal preset** AND names the "issue queue drained + diagnostics clean"
completion condition.

### AC-LSW-002 — existing scan lenses preserved (preservation)

```bash
grep -ic "LSP\|lint\|test" .claude/skills/moai/workflows/loop.md   # existing diagnostic lenses retained
```
**Preservation AC** (present → present): loop.md IS the current 4-tool loop, so
LSP/lint/test pre-match (~25×). This AC asserts the existing diagnostic lenses are
NOT dropped in the rewrite. It is NOT a discriminating check — the NEW review-lens
additions are pinned separately in AC-LSW-002b.

### AC-LSW-002b — NEW review-lens additions to the queue

```bash
grep -ic "review lens\|security lens\|@MX lens" .claude/skills/moai/workflows/loop.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0. PASS when loop.md adds the review-lens
queue suppliers (security + `@MX` review lenses) as NEW queue sources beyond the
existing LSP/lint/test lenses.

### AC-LSW-003 — no-invented-improvements AND empty-queue exit (two conjunctive checks)

```bash
# (a) no work outside the scanned queue:
grep -ic "no invented\|outside the scanned queue" .claude/skills/moai/workflows/loop.md   # expect ≥1 (baseline 0)
# (b) empty queue → immediate exit:
grep -ic "empty queue.*exit\|immediate exit" .claude/skills/moai/workflows/loop.md        # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): both 0. PASS when BOTH (a) AND (b) are ≥1
(conjunctive — this replaces the former non-conjunctive OR that could pass on one
half alone). This is the critical no-invented-improvements guard.

### AC-LSW-004 — ceiling precedence preserved (preservation)

```bash
grep -c "max_iterations" .claude/skills/moai/workflows/loop.md   # precedence rule retained (baseline ≥1)
```
**Preservation AC** (present → present): the iteration-ceiling precedence rule
(CLI `--max` > ralph.yaml `loop.max_iterations` > workflow.yaml
`loop_prevention.max_iterations`) MUST survive the rewrite.

### AC-LSW-004b — 5-section verdict + Step 1.5 preserved (split, not OR)

```bash
# Split the former 4-element preservation OR into individual presence checks:
grep -ic "5-section" .claude/skills/moai/workflows/loop.md      # ceiling-exit verdict retained (present)
grep -ic "Step 1.5" .claude/skills/moai/workflows/loop.md       # independent final pass retained (present)
grep -ic "loop-verdict" .claude/skills/moai/workflows/loop.md   # verdict-file persistence retained (present)
```
**Preservation AC** (present → present, split): each of the 5-section ceiling-exit
verdict, the Step 1.5 independent final pass, and the `loop-verdict-<id>.json`
persistence is checked INDIVIDUALLY (a single OR would be blind to a partial
regression). PASS when all three remain present after the rewrite.

### AC-LSW-005 — `sweep-residue` additive exit_kind

```bash
grep -ic "sweep-residue" .claude/skills/moai/workflows/loop.md          # expect ≥1 (baseline 0)
grep -n '"exit_kind".*sweep-residue' .claude/skills/moai/workflows/loop.md  # additive enum placement
```
Baseline (verified this iteration): `sweep-residue` = 0. PASS when loop.md names
the `sweep-residue` exit_kind value added ADDITIVELY (a fourth value alongside the
base `ceiling | manual-residue | one-shot-residue`). The bare `exit_kind` token is
dropped from the grep (it pre-matches the existing enum).

### AC-LSW-006 — loop consumes review lenses (loop.md side of the boundary)

```bash
grep -ic "review lens\|consumes review" .claude/skills/moai/workflows/loop.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0. PASS when loop.md documents that `/moai loop`
consumes review lenses as queue suppliers. `read-only` is dropped from the grep (it
pre-matches an unrelated line — vacuous otherwise).

### AC-LSW-007 — review.md cross-ref, commit-scoped diff (no behavior change)

```bash
# (a) review.md gains the cross-ref (baseline 0):
grep -ic "loop.*queue supplier\|consumed by.*loop\|layered" .claude/skills/moai/workflows/review.md   # expect ≥1
# (b) behavior-unchanged: commit-scoped diff shows ONLY the cross-ref paragraph changed.
#     (A no-ref `git diff --stat` is always empty after commit → false-PASS; use the run-phase commit SHA.)
git show --stat <review-md-commit-sha> -- .claude/skills/moai/workflows/review.md
```
Baseline (verified this iteration): (a) 0. PASS when review.md gains a 1-2
paragraph cross-ref AND the commit-scoped `git show --stat <sha>` (the run-phase
author cites the actual commit SHA that touched review.md) shows only the cross-ref
addition — no lens/pipeline/behavior lines changed.

### AC-LSW-008 — fix pipeline contract intact (preservation)

```bash
grep -c "No LLM-driven control flow\|Agentless fixed-pipeline" .claude/skills/moai/workflows/fix.md   # ≥2 (unchanged)
```
**Preservation AC** (present → present): the fix.md Agentless / no-LLM-control-flow
pipeline-contract lines MUST remain intact (fix behavior UNCHANGED — REQ-LSW-008).

### AC-LSW-008b — fix residue text updated ("enters the loop queue")

```bash
grep -ic "enters the loop queue" .claude/skills/moai/workflows/fix.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0. PASS when the fix.md residue-handoff text is
updated to state residue enters the loop queue. The bare `/moai loop` token is
dropped from the grep (it pre-matches fix.md 4× — vacuous otherwise).

### AC-LSW-009 — loop.md "goal engine + preset" framing

```bash
grep -ic "goal engine + preset\|goal-preset" .claude/skills/moai/workflows/loop.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0. PASS when loop.md re-expresses the taxonomy
using the NEW "goal engine + preset" framing. The bare `quadrant` token is kept
ONLY as a secondary preservation signal (it pre-matches loop.md:54/60), not the
discriminating check.

### AC-LSW-009b — fix.md "goal engine + preset" framing

```bash
grep -ic "goal engine + preset\|goal-preset" .claude/skills/moai/workflows/fix.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0 (fix.md `quadrant` already = 3, so a
`quadrant` grep would be vacuous; the NEW framing token is the discriminating
check). PASS when the fix.md sibling quadrant note adopts the "goal engine + preset"
framing consistently with loop.md (AC-LSW-009).

### AC-LSW-010 — goal-directive.md loop row NAMES goal preset

```bash
grep -ic "/moai loop.*goal preset\|goal preset.*/moai loop" .claude/rules/moai/workflow/goal-directive.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0 (the bare `/moai loop` pre-matches the
existing Ralph-Engine row 8× — vacuous; the row must NOW name it a goal preset).
PASS when the goal-directive.md loop row describes `/moai loop` as a goal preset
(distinct from native `/goal` and `/moai goal`).

### AC-LSW-011 — cadence loop-NOT-eligible preserved (preservation)

```bash
grep -in "loop" .claude/rules/moai/workflow/cadence-bridge.md | grep -i "not.*eligible\|MUST NOT"   # present (baseline ≥1, at line 39)
```
**Preservation AC** (present → present): cadence-bridge.md ALREADY states
`/moai loop` is NOT cadence-eligible (line 39: "Not cadence-eligible — these MUST
NOT appear... `/moai loop`..."). This AC asserts the SPEC does NOT weaken that
statement AND does NOT add loop to the recipe catalog. PASS when the
NOT-cadence-eligible statement remains present and loop is absent from the recipe
table.

### AC-LSW-012 — NEW alias-KEEP decision + justification

```bash
grep -ic "both routes resolve to\|alias.*keep\|keep.*alias" .claude/rules/moai/workflow/spec-workflow.md   # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): 0 (the `Multi-Agent` / `--mode loop`
classification already exists at spec-workflow.md:91 — a grep for those is vacuous;
the NEW alias-KEEP decision is the discriminating check). PASS when the loop row
states the settled alias-KEEP decision (both `/moai run --mode loop` and `/moai
loop` resolve to the goal-preset sweep) with its backward-compat justification.
The already-true Multi-Agent classification is asserted-unchanged, not re-counted.

### AC-LSW-013 — Go loop help text renamed, behavior unchanged

```bash
grep -ic "SPEC-lifecycle\|lifecycle controller" internal/cli/loop.go   # expect ≥1 (baseline 0)
go test ./internal/loop/... ./internal/ralph/... 2>&1 | tail -3        # green (behavior unchanged)
```
Baseline (verified this iteration): `SPEC-lifecycle|lifecycle controller` = 0
(current help: `Short: "Manage the Ralph feedback loop lifecycle"`). PASS when the
`internal/cli/loop.go` help text clarifies it is the SPEC-lifecycle controller
(distinct from the `/moai loop` sweep skill) AND the loop/ralph Go tests still pass
(string-only change, no behavior regression).

### AC-LSW-014 — per-file mirror parity (all changed mirrored files) + neutrality + build

The SPEC changes 8 template-mirrored `.claude/` files (all verified mirrored this
iteration; NONE are dev-only per CLAUDE.local.md §2). Assert each mirror exists:

```bash
for f in \
  CLAUDE.md \
  .claude/skills/moai/workflows/loop.md \
  .claude/skills/moai/workflows/fix.md \
  .claude/skills/moai/workflows/review.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/cadence-bridge.md \
  .claude/rules/moai/workflow/spec-workflow.md \
  .claude/skills/moai-workflow-loop/SKILL.md ; do
  test -f "internal/template/templates/$f" && echo "MIRROR $f OK" || echo "MISSING $f"
done
# neutrality: no internal SPEC ID leaked into any mirror body
grep -rn "SPEC-LOOP-SWEEP\|SPEC-GOAL-ENGINE\|AGENTIC-CORE\|REQ-LSW" internal/template/templates/.claude/ internal/template/templates/CLAUDE.md | wc -l   # expect 0
make build ; echo "exit=$?"
```
PASS when the loop prints no `MISSING` line (all 8 mirrors exist and carry the
reformed content — the run-phase author additionally spot-checks a reformed token
per mirror, e.g. `grep -c "goal preset" internal/template/templates/.claude/skills/moai/workflows/loop.md`),
the neutrality grep is 0, and `make build` exits 0. NOTE `internal/cli/loop.go` is
Go source, NOT a template mirror — it is verified by AC-LSW-013, not here.

## §D.1 Definition of Done

- All 17 ACs PASS (AC-LSW-001..014 with 002b / 004b / 008b / 009b split-outs).
- `/moai fix` + `/moai review` behavior unchanged (AC-LSW-008 preservation +
  AC-LSW-007 commit-scoped diff).
- `internal/loop`/`internal/ralph` logic untouched; loop/ralph Go tests green (AC-LSW-013).
- loop stays NOT cadence-eligible (AC-LSW-011 preservation).
- All 8 template mirrors updated + neutral + `make build` green (AC-LSW-014).

## §D.2 Edge cases

- **Empty scan queue**: loop exits immediately (AC-LSW-003 clause b) — no ceiling burn.
- **`--lens coverage` with coverage disabled**: coverage lens is opt-in; with no
  coverage gate the lens supplies no queue items (documented).
- **Review lens finds issues loop cannot auto-fix**: those enter the queue as
  manual-residue at ceiling exit (existing `manual-residue` path, unchanged).
- **Depends_on unmet**: run-phase entry blocked unless GOAL-ENGINE is `completed`
  or `--ignore-deps` + logged rationale (spec-workflow Depends_on pre-flight).
- **`run --mode loop` alias**: KEEP (settled); both routes resolve to the sweep
  preset (AC-LSW-012).
