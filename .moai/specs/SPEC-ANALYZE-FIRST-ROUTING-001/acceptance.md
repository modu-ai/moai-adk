# acceptance.md — SPEC-ANALYZE-FIRST-ROUTING-001

> Each AC is a **single discriminating check** (no vacuous compound greps).
> Router/CLAUDE.md-registration ACs are pinned SEPARATELY from agent-body ACs.
> Template-mirror ACs use per-file existence + content checks. Baseline 0 → post
> ≥ 1 where applicable.

## §D — Acceptance Criteria Matrix

| AC | REQ | Verifies | Baseline → Post |
|----|-----|----------|-----------------|
| AC-AFR-001 | REQ-AFR-001 | §2 window: 5 ordered stages (circled markers) | 0 → ≥1 |
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
| AC-AFR-015 | REQ-AFR-015 | mirror CONTENT parity (not just existence) | per-file content |
| AC-AFR-016 | REQ-AFR-016 | no internal SPEC ID in mirrors | 0 |
| AC-AFR-017 | REQ-AFR-017 | make build succeeds | exit 0 |

### AC-AFR-001 — §2 five ORDERED stages present (circled markers)

```bash
# Window to §2, then assert ordered-stage markers (proves 5 ORDERED stages, not just the phrase).
sec=$(awk '/^## 2\. Request Processing Pipeline/,/^## 3\./' CLAUDE.md)
echo "$sec" | grep -c "intent analysis"          # expect ≥1 (baseline 0)
echo "$sec" | grep -cE "①|②|③|④|⑤"              # expect ≥1 ordered markers (baseline 0)
```
Baseline (verified this iteration): both `intent analysis` and the circled markers
are 0 inside the §2 window. PASS when the §2 window carries an ORDERED enumeration
(circled markers ①-⑤) naming intent analysis → context-sufficiency →
execution-plan composition → approval gates → execute/verify/iterate.

### AC-AFR-002 — stale phrasing removed

```bash
grep -c "Detect technology keywords for agent matching" CLAUDE.md   # expect 0
```
PASS when the exact stale phrase count is 0.

### AC-AFR-003 — language-independent statement (scoped to §2 window)

```bash
# Scope to §2 (root-file `conversation_language` pre-matches at L9/62/202 — vacuous otherwise).
awk '/^## 2\. Request Processing Pipeline/,/^## 3\./' CLAUDE.md \
  | grep -ic "any input language\|language-independent"   # expect ≥1
```
Baseline (verified this iteration): 0 inside the §2 window. PASS when the §2 window
contains at least one sentence stating intent analysis applies to any input
language (not only English `/moai` tokens). `conversation_language` is deliberately
dropped from the grep — it pre-matches elsewhere in CLAUDE.md and would make the
check vacuous.

### AC-AFR-004 — enumeration + Kickoff gate live INSIDE §2 (attachment anchor)

This AC guards the SPEC's own HARD compat invariant (REQ-AFR-004), so it must
assert the enumeration + gate anchor live INSIDE §2 — NOT anywhere in CLAUDE.md
(`Phase 0.95` / `Kickoff` pre-match tombstone text at L122/264, and the root §2
currently has NEITHER).

```bash
sec=$(awk '/^## 2\. Request Processing Pipeline/,/^## 3\./' CLAUDE.md)
echo "$sec" | grep -c "intent analysis"                              # expect ≥1 (baseline 0)
echo "$sec" | grep -ci "Implementation Kickoff Approval\|approval gate"  # expect ≥1 (baseline 0)
echo "$sec" | grep -cE "①|②|③|④|⑤"                                  # expect ≥1 (baseline 0)
```
Baseline (verified this iteration): all three are 0 inside the §2 window. PASS when
all three are ≥1 inside §2 — the ordered enumeration (circled markers) AND a
reference to the Implementation Kickoff Approval gate both survive inside the §2
rewrite (HARNESS-EVOLVE Phase −1/Ω anchor).

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

### AC-AFR-007 — `lint` routes to fix, stripped from gate (two pinned lines)

```bash
# (a) the gate-cue line (SKILL.md:81 `... routes to **gate**`) must NOT contain `lint` post-fix.
grep -n "routes to \*\*gate\*\*" .claude/skills/moai/SKILL.md | grep -c "lint"   # expect 0
# (b) the fix-cue line (SKILL.md:83 `... routes to **fix**`) MUST contain `lint`.
grep -n "routes to \*\*fix\*\*" .claude/skills/moai/SKILL.md | grep -c "lint"    # expect ≥1
```
Baseline (verified this iteration): the collision is REAL — the gate-cue line (a)
currently DOES contain `lint` (count 1 → post 0, the discriminating removal), and
the fix-cue line (b) already contains `lint` (count ≥1, preservation). PASS when
(a) is 0 AND (b) is ≥1 — `lint` stripped from the gate-cue line, retained on the
fix-cue line (the `gate` bucket keeps "format, check, pre-commit, quality gate").

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
numbers). This is a delta check, not an absolute threshold (per plan.md § Settled
Decisions — the 7K figure is advisory, the delta is the HARD gate).

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

### AC-AFR-015 — mirror CONTENT parity (not just existence)

A bare `test -f` mirror-existence check passes on a stale mirror. This AC verifies
the mirrors carry the REFORMED content, per-file:

```bash
# (a) all 9 agent mirrors carry the diet line (baseline 0):
grep -l "Match user intent language-independently" internal/template/templates/.claude/agents/moai/*.md | wc -l   # expect 9
# (b) the CLAUDE.md mirror stale phrase removed (baseline 1 → post 0):
grep -c "Detect technology keywords for agent matching" internal/template/templates/CLAUDE.md   # expect 0
# (c) the CLAUDE.md mirror carries the language-independent statement (baseline 0):
grep -ic "any input language\|language-independent" internal/template/templates/CLAUDE.md   # expect ≥1
```
Baseline (verified this iteration): (a) 0, (b) 1, (c) 0. PASS when (a)=9, (b)=0,
(c)≥1 — the mirrors carry the reformed content, not a stale copy. AC-AFR-017
(`make build`) is **necessary-not-sufficient** and is paired with these content
checks (a green build does not prove the mirror content changed).

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
