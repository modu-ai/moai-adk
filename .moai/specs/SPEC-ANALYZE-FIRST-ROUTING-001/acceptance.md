# acceptance.md — SPEC-ANALYZE-FIRST-ROUTING-001

> Each AC is a **single discriminating check** (no vacuous compound greps).
> Router/CLAUDE.md-registration ACs are pinned SEPARATELY from agent-body ACs.
> Template-mirror ACs use per-file existence + content checks. Baseline 0 → post
> ≥ 1 where applicable.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Baseline → Post |
|----|-----|----------|-----------------|
| AC-AFR-001 | REQ-AFR-001 | §2 enumerates 5 ordered stages | absent → present |
| AC-AFR-002 | REQ-AFR-002 | stale keyword phrase removed | present → 0 |
| AC-AFR-003 | REQ-AFR-003 | language-independent statement present | 0 → ≥1 |
| AC-AFR-004 | REQ-AFR-004 | pipeline enumeration + Kickoff ref preserved | present → present |
| AC-AFR-005 | REQ-AFR-005 | P3 language-exemplar clause present | 0 → ≥1 |
| AC-AFR-006 | REQ-AFR-006 | P1 fast-path unchanged | present → present |
| AC-AFR-007 | REQ-AFR-007 | `lint` single-bucket (fix) | (see AC body) |
| AC-AFR-008 | REQ-AFR-008 | "MUST INVOKE" removed from all 9 agents | 8 files → 0 files |
| AC-AFR-009 | REQ-AFR-009 | language-independence line in all 9 agents | 0 → 9 files |
| AC-AFR-010 | REQ-AFR-010 | manager-design gains scope prose | absent → present |
| AC-AFR-011 | REQ-AFR-011 | post-diet char total < pre-diet baseline | baseline → smaller |
| AC-AFR-012 | REQ-AFR-012 | foundation-core related-skills matches template | drift → aligned |
| AC-AFR-013 | REQ-AFR-013 | skill-authoring triggers section reflects semantic-matching | (see body) |
| AC-AFR-014 | REQ-AFR-014 | no dead `team/*.md` ref in SKILL.md | (see body) |
| AC-AFR-015 | REQ-AFR-015 | each changed file has a mirror | per-file |
| AC-AFR-016 | REQ-AFR-016 | no internal SPEC ID in mirrors | 0 |
| AC-AFR-017 | REQ-AFR-017 | make build succeeds | exit 0 |

### AC-AFR-001 — §2 five-stage enumeration present

```bash
grep -c "intent analysis" CLAUDE.md   # expect ≥1 (stage ①)
# AND the ordered pipeline enumeration is present in §2 (numbered/ordered list of 5 stages).
```
PASS when the §2 section carries an ordered enumeration naming intent analysis →
context-sufficiency → execution-plan composition → approval gates →
execute/verify/iterate.

### AC-AFR-002 — stale phrasing removed

```bash
grep -c "Detect technology keywords for agent matching" CLAUDE.md   # expect 0
```
PASS when the exact stale phrase count is 0.

### AC-AFR-003 — language-independent statement

```bash
grep -in "any input language\|language-independent\|conversation_language" CLAUDE.md | head
```
PASS when §2 contains at least one sentence stating intent analysis applies to any
input language (not only English `/moai` tokens).

### AC-AFR-004 — enumeration + Kickoff preserved (attachment anchor)

```bash
grep -c "Phase 0.95\|Implementation Kickoff Approval\|Kickoff" CLAUDE.md   # expect ≥1
```
PASS when the §2 rewrite retains the ordered pipeline enumeration AND a reference
to the Implementation Kickoff Approval gate (REQ-AFR-004 — HARNESS-EVOLVE anchor).

### AC-AFR-005 — P3 language-exemplar clause

```bash
grep -in "exemplar\|not.*literal\|any conversation_language" .claude/skills/moai/SKILL.md | head
```
PASS when the Intent Router P3 body states cue words are English exemplars, not
literal-match requirements, and intent is classified for any `conversation_language`.

### AC-AFR-006 — P1 fast-path unchanged

```bash
grep -n "Intent Router" .claude/skills/moai/SKILL.md   # section still present
```
PASS when the P1 first-word subcommand fast-path text is unchanged (diff of the P1
block against pre-edit shows no P1 semantic change; only P3 is extended).

### AC-AFR-007 — `lint` routes to fix (single bucket)

```bash
# Discriminating check: lint appears in the fix bucket, NOT the gate bucket.
# Read the P3 body and confirm `lint` is listed under fix only.
grep -n "lint" .claude/skills/moai/SKILL.md
```
PASS when `lint` is a cue for the `fix` bucket and is NOT also listed under `gate`
(the `gate` bucket keeps "quality gate / pre-commit / check"). If the pre-flight
finds `lint` was never dual-membership, this AC PASSES by confirming single
membership (no regression introduced).

### AC-AFR-008 — "MUST INVOKE" removed from all 9 agents

```bash
grep -l "MUST INVOKE" .claude/agents/moai/*.md | wc -l   # expect 0
```
PASS when the file count is 0 (baseline was 8).

### AC-AFR-009 — language-independence line in all 9 agents

```bash
grep -l "Match user intent language-independently" .claude/agents/moai/*.md | wc -l   # expect 9
```
PASS when the line appears in all 9 agent files (baseline 0). (If a verbatim
variant is used, the AC uses that exact string — one canonical sentence, checked
per-file.)

### AC-AFR-010 — manager-design scope prose

```bash
grep -c "language-independently\|Match user intent" .claude/agents/moai/manager-design.md   # expect ≥1
```
PASS when `manager-design.md` gains the concise scope/trigger prose it lacked.

### AC-AFR-011 — post-diet char total strictly smaller

```bash
# Compare against the pre-flight baseline captured in §C.
wc -c .claude/agents/moai/*.md | tail -1
```
PASS when the re-measured total agent-body char count is strictly less than the
pre-diet baseline recorded in progress.md §E.2 (the run-phase author records both
numbers). This is a delta check, not an absolute threshold (per plan.md § Open
Decisions — the 7K figure is advisory).

### AC-AFR-012 — foundation-core related-skills aligned

```bash
diff <(grep "related-skills" .claude/skills/moai-foundation-core/SKILL.md) \
     <(grep "related-skills" internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md)
```
PASS when the local `related-skills` value equals the template value (empty diff),
and no longer references `moai-foundation-context`.

### AC-AFR-013 — skill-authoring triggers section updated

```bash
grep -in "triggers:.*optional\|semantic matching\|not a matcher" .claude/rules/moai/development/skill-authoring.md | head
```
PASS when the `triggers:` section states it is optional metadata reflecting
model-side semantic matching (not a literal matcher).

### AC-AFR-014 — no dead team/*.md reference

```bash
# For each team/*.md path referenced in SKILL.md, confirm it resolves; else it was removed.
for p in $(grep -o "team/[A-Za-z0-9_-]*\.md" .claude/skills/moai/SKILL.md | sort -u); do
  test -f ".claude/skills/moai/workflows/$p" -o -f ".claude/skills/moai/$p" && echo "RESOLVES $p" || echo "DEAD $p"
done
```
PASS when the loop prints no `DEAD` line (every remaining `team/*.md` reference
resolves, or all dead ones were removed).

### AC-AFR-015 — per-file mirror existence

```bash
# For each changed .claude/ file, assert the template mirror exists.
for f in CLAUDE.md .claude/skills/moai/SKILL.md .claude/agents/moai/manager-spec.md; do
  test -f "internal/template/templates/$f" && echo "MIRROR $f OK" || echo "MISSING $f"
done
# (full list per plan.md §A.5 — the run-phase author iterates all changed files)
```
PASS when every changed `.claude/` file (CLAUDE.md is mirrored at
`internal/template/templates/CLAUDE.md`) has an existing mirror.

### AC-AFR-016 — mirror neutrality (no internal SPEC ID)

```bash
grep -rn "SPEC-ANALYZE-FIRST\|SPEC-GOAL-ENGINE\|SPEC-LOOP-SWEEP\|AGENTIC-CORE\|REQ-AFR" \
  internal/template/templates/.claude/ internal/template/templates/CLAUDE.md | wc -l   # expect 0
```
PASS when the count is 0 (no Epic/SPEC/REQ token leaked into any mirror body).

### AC-AFR-017 — make build succeeds

```bash
make build ; echo "exit=$?"
```
PASS when `exit=0` (templates recompiled).

## §D.1 Definition of Done

- All 17 ACs PASS.
- No Go source modified (`git diff --name-only | grep -E '\.go$'` is empty).
- `run.md § Run-phase Autonomy` section unmodified (owned by AUTONOMY-RUN-GOAL-001).
- CLAUDE.md §2 ordered enumeration preserved (AC-AFR-004).
- All template mirrors updated + neutral + `make build` green.

## §D.2 Edge cases

- **Verbatim-variant of the language-independence line**: AC-AFR-009 pins ONE
  canonical sentence. If the run-phase author chooses a variant, the AC string is
  updated to that exact variant (single SSOT sentence), still checked per-file.
- **`lint` never was dual-membership**: AC-AFR-007 degrades to a single-membership
  confirmation (no regression), documented in progress.md §E.2.
- **manager-design already partially has scope prose**: AC-AFR-010 still requires
  the canonical language-independence line to be present.
