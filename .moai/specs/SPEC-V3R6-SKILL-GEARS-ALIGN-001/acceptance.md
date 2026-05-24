---
id: SPEC-V3R6-SKILL-GEARS-ALIGN-001
title: "Acceptance Criteria — moai-workflow-spec SKILL GEARS 우선 정렬"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai-workflow-spec, .claude/skills/moai-foundation-core/modules, .claude/agents/core, internal/template/templates"
lifecycle: spec-anchored
tags: "gears, ears, skill, guide, alignment, acceptance, wave-6, v3.0.0"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001]
related_specs: [SPEC-V3R6-GEARS-MIGRATION-001]
---

# SPEC-V3R6-SKILL-GEARS-ALIGN-001 — Acceptance Criteria

## Verification Doctrine

- Each AC is **binary** (PASS / FAIL).
- Each AC includes a verification command executable from project root `/Users/goos/MoAI/moai-adk-go/`.
- Commands use `grep`, `wc -l`, `diff`, and `moai spec lint`. No subjective assessment.
- Where applicable, the expected output is shown.

## Traceability Matrix (REQ ↔ AC)

| REQ | AC | Coverage |
|-----|------|----------|
| REQ-SGA-001 | AC-SGA-002, AC-SGA-005 | GEARS primary in SKILL.md + comparison table |
| REQ-SGA-002 | AC-SGA-002 | EARS Five Patterns section reframed |
| REQ-SGA-003 | AC-SGA-003, AC-SGA-009 | spec-ears-format.md deprecation + cross-link |
| REQ-SGA-004 | AC-SGA-008 | 6-month backward-compat note retained |
| REQ-SGA-005 | AC-SGA-006 | Compound clause example with non-"the system" subject |
| REQ-SGA-006 | AC-SGA-007 | manager-spec.md description + 4-locale keywords contain "GEARS" |
| REQ-SGA-007 | AC-SGA-004 | 5 files × 2 locations byte-identical |
| REQ-SGA-008 | AC-SGA-001 | Self-lint LegacyEARSKeyword 0 on this spec.md |
| REQ-SGA-009 | AC-SGA-013 | New IF/THEN examples absent; legacy retained with deprecation marker |
| REQ-SGA-010 | AC-SGA-005 | Generalized-subject substitution pattern with ≥2 non-"the system" examples |
| REQ-SGA-011 | AC-SGA-010 | 12-field canonical frontmatter + tier:M + related_specs |
| REQ-SGA-012 | AC-SGA-011 | h3 Out of Scope sub-section present in all 3 artifacts |
| (cross-cutting) | AC-SGA-012 | PRESERVE 14 pre-existing untracked + 6 modified/deleted files intact |

100% REQ coverage; AC-SGA-001 .. AC-SGA-013 = 13 binary ACs.

---

## AC-SGA-001 — Self-lint LegacyEARSKeyword count = 0 on this SPEC

**REQ**: REQ-SGA-008
**Intent**: Self-dogfood — this SPEC's own REQs use GEARS notation only, no `IF/THEN`.

**Verification command** (from project root):

```bash
moai spec lint .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md 2>&1 | grep -c "LegacyEARSKeyword"
```

**Expected output**: `0`

**Pass criterion**: Output is exactly `0`. Any non-zero count is FAIL.

**Alternate verification** (if `moai spec lint` CLI is unavailable):

```bash
go test -run TestEARSModalityRule ./internal/spec/ -v 2>&1 | tail -20
```

---

## AC-SGA-002 — SKILL.md contains GEARS five-pattern comparison table

**REQ**: REQ-SGA-001, REQ-SGA-002
**Intent**: GEARS presented as primary notation alongside EARS legacy in SKILL.md.

**Verification command**:

```bash
grep -c "GEARS" .claude/skills/moai-workflow-spec/SKILL.md
```

**Expected output**: `>= 5` (occurrences of "GEARS" in SKILL.md after edit)

**Pass criterion**: Output `>= 5`. Baseline 0 (zero "GEARS" occurrences pre-edit). Post-edit must include: (a) GEARS Five Patterns section heading, (b) comparison table header row, (c) 5-pattern enumeration with GEARS column, (d) cross-link to docs-site `#gears-notation` anchor, (e) reference to predecessor SPEC-V3R6-GEARS-MIGRATION-001.

**Secondary check** — comparison table structure:

```bash
grep -A 8 "GEARS.*EARS\|EARS.*GEARS" .claude/skills/moai-workflow-spec/SKILL.md | grep -c "Ubiquitous\|WHEN\|WHILE\|WHERE\|IF"
```

**Expected output**: `>= 5` (one row per pattern).

---

## AC-SGA-003 — spec-ears-format.md contains IF/THEN deprecation banner

**REQ**: REQ-SGA-003
**Intent**: Authors loading the spec-format reference see deprecation upfront.

**Verification command**:

```bash
grep -c "DEPRECATED\|deprecated\|GEARS\|6.month\|backward.compat" .claude/skills/moai-foundation-core/modules/spec-ears-format.md
```

**Expected output**: `>= 3` (at minimum, one DEPRECATED/deprecated, one GEARS, one 6-month/backward-compat marker)

**Pass criterion**: Output `>= 3`. Baseline 0 pre-edit.

**Secondary check** — banner near document top:

```bash
head -30 .claude/skills/moai-foundation-core/modules/spec-ears-format.md | grep -c "DEPRECATED\|deprecated\|GEARS"
```

**Expected output**: `>= 1` (banner appears in first 30 lines).

---

## AC-SGA-004 — Template mirror byte-identical parity for 5 files

**REQ**: REQ-SGA-007
**Intent**: Template-First Rule [HARD] CLAUDE.local.md §2.

**Verification command** (5 file pairs):

```bash
for f in \
    .claude/skills/moai-workflow-spec/SKILL.md \
    .claude/skills/moai-workflow-spec/references/reference.md \
    .claude/skills/moai-workflow-spec/references/examples.md \
    .claude/skills/moai-foundation-core/modules/spec-ears-format.md \
    .claude/agents/core/manager-spec.md; do
  diff -q "$f" "internal/template/templates/$f" || echo "DRIFT: $f"
done
```

**Expected output**: Empty (no `DRIFT:` lines).

**Pass criterion**: Zero `DRIFT:` lines emitted. Any drift is FAIL.

**Note**: If a file contains runtime-managed lines (none expected for these 5 markdown files), substitute `diff -q` with semantic comparison documented in the plan.

---

## AC-SGA-005 — SKILL.md contains generalized-subject substitution guidance with ≥2 non-"the system" examples

**REQ**: REQ-SGA-010
**Intent**: Authors learn that `<subject>` can be substituted with any noun.

**Verification command**:

```bash
grep -c "The skill shall\|The agent shall\|The component shall\|The service shall\|The function shall\|generalized subject\|<subject>" .claude/skills/moai-workflow-spec/SKILL.md
```

**Expected output**: `>= 3` (at minimum: one "generalized subject" or `<subject>` description + two distinct non-"the system" example phrases).

**Pass criterion**: Output `>= 3` AND grep shows at least 2 distinct non-"the system" subject phrases.

**Secondary check**:

```bash
grep -oE "The [a-z]+ shall" .claude/skills/moai-workflow-spec/SKILL.md | sort -u | wc -l
```

**Expected output**: `>= 3` (e.g., "The system shall", "The skill shall", "The agent shall" — at least 3 distinct subjects).

---

## AC-SGA-006 — Compound clause example present in SKILL.md or examples.md

**REQ**: REQ-SGA-005
**Intent**: GEARS unified compound pattern is demonstrated.

**Verification command**:

```bash
grep -cE "Where .*While .*When|Where .*When|While .*When" \
    .claude/skills/moai-workflow-spec/SKILL.md \
    .claude/skills/moai-workflow-spec/references/examples.md \
    2>&1 | grep -v ":0$" | wc -l
```

**Expected output**: `>= 1` (at least one file contains a compound clause with chained `Where`/`While`/`When` modifiers).

**Pass criterion**: At least one of the two files contains a compound pattern example.

**Secondary check** — example follows full unified form:

```bash
grep -E "^\*\*(Where|While|When)\*\*|^.*\\[Where.*\\]\\[While.*\\]\\[When.*\\]|Where .* While .* When .* shall" \
    .claude/skills/moai-workflow-spec/SKILL.md \
    .claude/skills/moai-workflow-spec/references/examples.md \
    | head -5
```

**Expected output**: At least 1 line matching unified form.

---

## AC-SGA-007 — manager-spec.md description + 4-locale keywords contain "GEARS"

**REQ**: REQ-SGA-006
**Intent**: Agent dispatcher and keyword-based invocation include GEARS triggers.

**Verification command** (description block):

```bash
grep -c "GEARS" .claude/agents/core/manager-spec.md
```

**Expected output**: `>= 5` (description mention + EN/KO/JA/ZH keyword lines = 5 minimum).

**Pass criterion**: Output `>= 5`.

**Secondary check** — keyword sections (4 locales):

```bash
grep -E "^\s*(EN|KO|JA|ZH):" .claude/agents/core/manager-spec.md | grep -c "GEARS"
```

**Expected output**: `4` (one per locale: EN/KO/JA/ZH).

---

## AC-SGA-008 — 6-month backward-compatibility note retained in spec-ears-format.md

**REQ**: REQ-SGA-004
**Intent**: Existing 88 SPEC authors retain context during v3.0.0 transition.

**Verification command**:

```bash
grep -c "6.month\|6 months\|backward.compat\|backward-compatibility" .claude/skills/moai-foundation-core/modules/spec-ears-format.md
```

**Expected output**: `>= 1`

**Pass criterion**: Output `>= 1`. Note must reference the predecessor SPEC-V3R6-GEARS-MIGRATION-001 or v3.0.0 release.

---

## AC-SGA-009 — Cross-link to docs-site `#gears-notation` anchor present in SKILL.md

**REQ**: REQ-SGA-003 (cross-link clause)
**Intent**: Authors discoverably navigate from skill body to docs-site migration guide.

**Verification command**:

```bash
grep -c "moai-plan/#gears-notation\|moai-plan.md#gears-notation\|#gears-notation" .claude/skills/moai-workflow-spec/SKILL.md
```

**Expected output**: `>= 1`

**Pass criterion**: Output `>= 1`. URL must point to an existing docs-site anchor.

**Secondary check** — anchor exists in docs-site:

```bash
grep -c "{#gears-notation}\|#gears-notation" docs-site/content/en/workflow-commands/moai-plan.md
```

**Expected output**: `>= 1` (anchor exists, baseline from predecessor PR #1046).

---

## AC-SGA-010 — Frontmatter canonical 12-field + tier:M + related_specs

**REQ**: REQ-SGA-011
**Intent**: spec-lint FrontmatterInvalid emits 0 on this SPEC's 3 artifacts.

**Verification command** — required fields present in spec.md:

```bash
for k in id title version status created updated author priority phase module lifecycle tags; do
  grep -c "^${k}:" .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md \
    || echo "MISSING: $k"
done | grep -v "^[1-9]"
```

**Expected output**: Empty (all 12 fields present, count ≥ 1 each).

**Pass criterion**: Zero `MISSING:` lines emitted.

**Secondary check** — `tier: M` and `related_specs` present:

```bash
grep -c "^tier: M\$\|^related_specs:" .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md
```

**Expected output**: `>= 2`

**Tertiary check** — no snake_case alias:

```bash
grep -cE "^(created_at|updated_at|spec_id):" \
    .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md \
    .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/plan.md \
    .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/acceptance.md
```

**Expected output**: `0` (zero snake_case alias usage).

---

## AC-SGA-011 — h3 Out of Scope sub-section present in all 3 artifacts

**REQ**: REQ-SGA-012
**Intent**: spec-lint `MissingExclusions` ERROR avoided (h2 `## Out of Scope` is rejected; h3 form required).

**Verification command**:

```bash
for f in .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md \
         .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/plan.md \
         .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/acceptance.md; do
  grep -c "^### .* Out of Scope" "$f" || echo "MISSING h3 Out of Scope: $f"
done | grep -v "^[1-9]"
```

**Expected output**: Empty (each artifact contains `>= 1` h3 Out of Scope sub-section).

**Pass criterion**: Zero `MISSING h3 Out of Scope:` lines.

**Secondary check** — no h2 `## Out of Scope` standalone (would trigger MissingExclusions ERROR):

```bash
grep -cE "^## Out of Scope\$" \
    .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md \
    .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/plan.md \
    .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/acceptance.md
```

**Expected output**: `0`

---

## AC-SGA-012 — PRESERVE 14 pre-existing untracked + 6 modified/deleted files intact

**REQ**: cross-cutting (B8 Working Tree Hygiene)
**Intent**: This SPEC plan-phase and run-phase do not touch the 14 pre-existing untracked items or 6 modified/deleted files captured in spec.md §5.1 baseline snapshot.

**Verification command** — baseline file count consistency:

```bash
git status --porcelain | wc -l
```

**Expected output**: `21` after plan-phase write (6 modified/deleted + 15 untracked where 1 of 15 is this SPEC's new `.moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/` directory).

**Pass criterion**: `git status --porcelain` count = `21` strictly. Any divergence from the spec.md §5.1 enumeration (in either direction) triggers a blocker.

**Secondary check** — pre-existing modified/deleted files unchanged:

```bash
git status --porcelain | grep -E "^( M| D)" | wc -l
```

**Expected output**: `6` (porcelain column 2 = unstaged modifications/deletions; 5 M + 1 D per spec.md §5.1).

**Tertiary check** — pre-existing untracked items still present:

```bash
git status --porcelain | grep "^?? " | grep -cE "(docs-site/content/.*/book/|\.moai/research/.*-2026-05-22\.md|internal/hook/\.moai/|\.moai/specs/SPEC-V3R6-CODE-COMMENTS-EN-001/|\.moai/specs/SPEC-V3R6-RULES-COMPRESS-001/progress\.md)"
```

**Expected output**: `>= 5` (at least 5 pre-existing untracked baseline patterns still present).

---

## AC-SGA-013 — New IF/THEN examples absent; legacy IF/THEN annotated with deprecation marker

**REQ**: REQ-SGA-009
**Intent**: Authors do NOT receive guidance to write new IF/THEN REQs. Legacy IF/THEN examples (if retained as historical reference) carry deprecation markers.

**Verification command** — new IF/THEN REQ examples absent in primary teaching surfaces:

```bash
grep -E "^### .*\\(.*IF/THEN.*\\)|^- \\*\\*IF\\*\\*.*\\*\\*THEN\\*\\*|IF .* THEN" .claude/skills/moai-workflow-spec/SKILL.md | grep -v "DEPRECATED\|deprecated\|legacy\|backward-compat" | wc -l
```

**Expected output**: `0` (no IF/THEN REQ-style examples without deprecation marker in SKILL.md).

**Pass criterion**: Output `0`. Any IF/THEN reference must carry a deprecation marker.

**Note (S3 plan-auditor finding, 2026-05-23)**: a previous draft included a secondary check requiring `>= 1` IF/THEN deprecation marker hit in `spec-ears-format.md`. Baseline scan confirmed the file contains zero IF/THEN syntactic patterns (the "Unwanted" pattern uses `the system SHALL NOT [action]` instead). The secondary check was unsatisfiable as written and has been removed. REQ-SGA-009 remains conditional: *if* any IF/THEN reference exists, it must carry a deprecation marker; absence of IF/THEN references is itself compliant with REQ-SGA-009.

---

## Verification Order (run-phase exit gate)

1. AC-SGA-010 + AC-SGA-011 — Frontmatter + h3 structure (cheapest, file-level)
2. AC-SGA-001 — Self-lint LegacyEARSKeyword = 0
3. AC-SGA-002 + AC-SGA-005 + AC-SGA-006 — SKILL.md content checks (GEARS table + generalized subject + compound clause)
4. AC-SGA-003 + AC-SGA-008 + AC-SGA-009 — spec-ears-format.md + cross-link
5. AC-SGA-007 — manager-spec.md description + keywords
6. AC-SGA-013 — IF/THEN absence + deprecation markers
7. AC-SGA-004 — Template mirror byte-identical (final, most expensive)
8. AC-SGA-012 — PRESERVE intact (final smoke check)

All 13 ACs must PASS. Any FAIL blocks run-phase completion.

## Definition of Done

- All 13 ACs PASS via the documented verification commands.
- `moai spec lint .moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/spec.md` emits 0 `LegacyEARSKeyword` findings and 0 `FrontmatterInvalid` findings.
- All 5 file pairs (local + template) are byte-identical.
- PRESERVE invariants from AC-SGA-012 hold.
- Plan-auditor PASS threshold `>= 0.80` achieved (Tier M).

## Out of Scope Reaffirmation

### 8.1 Out of Scope (acceptance-level)

- This acceptance.md does NOT prescribe internal SPEC-V3R6-GEARS-SWEEP-001 acceptance criteria — that is the sweep SPEC's own scope.
- This acceptance.md does NOT re-test the predecessor's M2 lint engine behavior — predecessor AC-GM-002 already covers `LegacyEARSKeyword` emit on IF/THEN; this SPEC only asserts that the new guide files do not trigger that warning on their own REQ examples.
