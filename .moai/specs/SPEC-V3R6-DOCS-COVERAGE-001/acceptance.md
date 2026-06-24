# acceptance.md — SPEC-V3R6-DOCS-COVERAGE-001

> Acceptance criteria matrix. Each AC uses per-locale digit-boundary-anchored grep verification per applied lesson `feedback_digit_boundary_locale_grep_4parity`. Severity: MUST (blocking) / SHOULD (non-blocking).

---

## §D. AC Matrix

### AC-001 — Canonical count "32" replaces "31" across all 4 locales [MUST]

**Given** the docs-site 4 locales (en/ko/ja/zh) currently state "31" for the skill count
**When** the run-phase reconciliation is complete
**Then** digit-boundary-anchored grep for "31" + skill-adjacent keyword returns 0 matches per locale:

```bash
# Per-locale verification (NEVER glob — per-locale discipline)
for loc in en ko ja zh; do
  # en uses "skill", ko "스킬", ja "スキル", zh "技能"
  case $loc in
    en) kw="skill";;
    ko) kw="스킬";;
    ja) kw="スキル";;
    zh) kw="技能";;
  esac
  count=$(grep -rnE '(^|[^0-9])31([^0-9]|$)' "docs-site/content/$loc/" 2>/dev/null | grep -ic "$kw")
  echo "$loc: $count residual '31'+skill matches (expected 0)"
done
```

**PASS criterion**: each locale prints `0`. Any non-zero = FAIL.

### AC-002 — Canonical count "32" present in all 4 locales [MUST]

**Given** the reconciliation introduces "32" as the canonical count
**When** the run-phase is complete
**Then** digit-boundary-anchored grep for "32" + skill-adjacent keyword returns ≥1 match per locale:

```bash
for loc in en ko ja zh; do
  case $loc in
    en) kw="skill";;
    ko) kw="스킬";;
    ja) kw="スキル";;
    zh) kw="技能";;
  esac
  count=$(grep -rnE '(^|[^0-9])32([^0-9]|$)' "docs-site/content/$loc/" 2>/dev/null | grep -ic "$kw")
  echo "$loc: $count '32'+skill matches (expected ≥1)"
done
```

**PASS criterion**: each locale prints `≥1`. Zero = FAIL (count was not introduced).

### AC-003 — `moai-domain-humanize` present in en/ko/zh Domain category [MUST]

**Given** `moai-domain-humanize` is absent from all docs locales
**When** the run-phase adds it to the Domain category listing
**Then** grep for "humanize" returns ≥1 match per locale (en/ko/ja/zh):

```bash
for loc in en ko ja zh; do
  count=$(grep -rl 'humanize' "docs-site/content/$loc/" 2>/dev/null | wc -l | tr -d ' ')
  echo "$loc: $count files mentioning 'humanize' (expected ≥1)"
done
```

**PASS criterion**: each locale prints `≥1`. Zero = FAIL (humanize still missing).

### AC-004 — Domain category sub-count = 9 across en/ko/zh [MUST]

**Given** the docs state Domain has 8 skills (wrong; actual 9)
**When** the run-phase corrects the sub-count
**Then** the Domain category header/label shows 9 in each of en/ko/zh:

```bash
# en: "Domain (Domain Expertise) - 9 skills" or "Domain ... 9"
# ko: "Domain ... 9개" equivalent
# zh: "Domain(9)" or "Domain 9 个" — format-robust regex (iter-2 D5 broadened from Domain\(9\) to Domain.*9|9.*Domain)
for loc in en ko zh; do
  case $loc in
    en) count=$(grep -rnE 'Domain.*9|9.*Domain|Domain \(.*\) - 9' "docs-site/content/$loc/advanced/skill-guide.md" | wc -l | tr -d ' ');;
    ko) count=$(grep -rnE 'Domain.*9|Domain.*9개' "docs-site/content/$loc/advanced/skill-guide.md" | wc -l | tr -d ' ');;
    zh) count=$(grep -rnE 'Domain.*9|9.*Domain' "docs-site/content/$loc/advanced/skill-guide.md" | wc -l | tr -d ' ');;
  esac
  echo "$loc: $count Domain=9 matches (expected ≥1)"
done
```

**PASS criterion**: each locale prints `≥1`.

### AC-005 — Category sub-count sum = 31 specialized + 1 umbrella = 32 [MUST]

**Given** the 6-category sub-counts must sum to 31 specialized
**When** the run-phase corrects Domain to 9
**Then** the per-category sub-counts are mechanically consistent with the primary-source `find` outputs recorded in research.md §1.2 (Foundation=4, Workflow=10, Domain=9, Reference=5, Meta/Harness=2, Design=1). The arithmetic-sum invariant (4+10+9+5+2+1 = 31) is implied by per-category correctness — no separate manual arithmetic is required.

```bash
# Mechanical verification — per-category sub-count check against primary source.
# Anchor locale: zh `advanced/skill-guide.md` L61 carries the most detailed single-line breakdown
# "Foundation(4)、Workflow(10)、Domain(8)、Reference(5)、Meta-Harness(2)、Design(1)".
# Post-correction, every category that appears in any locale's breakdown MUST match the
# primary-source find count for that prefix. Domain is the load-bearing correction (8→9);
# the other 5 categories are already correct and serve as the baseline invariants.

# Per-prefix primary-source counts (reproduce from research.md §1.2):
FOUNDATION=$(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-foundation-*' | wc -l | tr -d ' ')   # expected 4
WORKFLOW=$(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-workflow-*' | wc -l | tr -d ' ')       # expected 10
DOMAIN=$(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-domain-*' | wc -l | tr -d ' ')           # expected 9
REFERENCE=$(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-ref-*' | wc -l | tr -d ' ')           # expected 5
META_HARNESS=$(($(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-meta-*' | wc -l) + $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-harness-*' | wc -l)))  # expected 2
DESIGN=$(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-design-*' | wc -l | tr -d ' ')           # expected 1
SPECIALIZED=$((FOUNDATION + WORKFLOW + DOMAIN + REFERENCE + META_HARNESS + DESIGN))
echo "specialized sum = $SPECIALIZED (expected 31)"

# Docs-side consistency: every locale that carries a Domain sub-count MUST show 9 (not 8).
# AC-004 already enforces Domain=9 per-locale; the arithmetic-sum invariant below is the
# complementary mechanical check that the per-category values add to 31 specialized.
# Run after AC-004 PASS; if AC-004 passes for all 3 locales AND the primary-source sum is 31,
# the docs cannot carry a breakdown that sums to anything other than 31.
echo "if AC-004 PASS (Domain=9 in en/ko/zh) AND specialized sum == 31 → AC-005 PASS"
```

**PASS criterion**: primary-source `SPECIALIZED` sum equals `31` AND AC-004 passes for all 3 locales. The combination mechanically implies no locale carries a breakdown summing to a different specialized total. (iter-2 D2: replaced iter-1 "manual arithmetic verification" with this mechanical grep+find assertion.)

### AC-006 — ja AND ko locale structural rewrite: 9 fictional categories eliminated [MUST]

> **iter-2 scope expansion (2026-06-18):** iter-1 scoped this AC to ja only. Independent re-derivation confirmed ko `advanced/skill-guide.md` carries the identical 9-fictional-category / 37-fictional-name taxonomy as ja. This AC now runs a per-locale loop over BOTH ja and ko. Per `feedback_digit_boundary_locale_grep_4parity`, the loop is per-locale — NEVER globbed across both locales (globbing would mask per-locale drift).

**Given** ja AND ko `advanced/skill-guide.md` each carry 9 categories including nonexistent Language/Platform/Library/Framework/Design-Tools plus 37 nonexistent skill-name references each
**When** the run-phase replaces both taxonomies with the canonical 6-category structure
**Then** for EACH of ja and ko (run the sub-checks independently per locale):

```bash
# Per-locale loop over BOTH structurally-divergent locales (iter-2: ja + ko)
for loc in ja ko; do
  echo "=== $loc ==="

  # Sub-check A: 9-category claim eliminated
  case $loc in
    ja) nine=$(grep -c '9カテゴリ\|9 カテゴリ' "docs-site/content/$loc/advanced/skill-guide.md");;
    ko) nine=$(grep -c '9.*카테고리\|9개 카테고리\|9 카테고리' "docs-site/content/$loc/advanced/skill-guide.md");;
  esac
  echo "$loc: 9-category claim count = $nine (expected 0)"

  # Sub-check B: 6-category claim introduced
  case $loc in
    ja) six=$(grep -c '6カテゴリ\|6 カテゴリ' "docs-site/content/$loc/advanced/skill-guide.md");;
    ko) six=$(grep -c '6.*카테고리\|6개 카테고리\|6 카테고리' "docs-site/content/$loc/advanced/skill-guide.md");;
  esac
  echo "$loc: 6-category claim count = $six (expected ≥1)"

  # Sub-check C: fictional skill-name prefixes eliminated (37 each → 0)
  fic=$(grep -cE 'moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context' "docs-site/content/$loc/advanced/skill-guide.md")
  echo "$loc: fictional-name count = $fic (expected 0; was 37 pre-correction)"

  # Sub-check D: all 3 missing canonical categories introduced (Reference, Meta/Harness, Design — not "Design Tools")
  case $loc in
    ja) refs=$(grep -cE '^### (Reference|Meta/Harness|Design )' "docs-site/content/$loc/advanced/skill-guide.md");;
    ko) refs=$(grep -cE '^### (Reference|Meta/Harness|Design )' "docs-site/content/$loc/advanced/skill-guide.md");;
  esac
  echo "$loc: canonical-only category headers (Reference/Meta-Harness/Design) = $refs (expected 3)"
done
```

**PASS criterion**: for BOTH ja and ko, every sub-check passes — 9-category claim = 0, 6-category claim ≥1, fictional-name count = 0, canonical-only category headers = 3.

### AC-007 — 4-locale parity: same count + same category structure [MUST]

**Given** all corrections must land simultaneously
**When** the run-phase commits the 4-locale changes
**Then** a parity diff confirms all 4 locales carry the same canonical count (32) and the same 6-category structure:

```bash
# Count parity: all 4 locales have "32" + skill, zero have "31" + skill (AC-001 + AC-002 combined)
# Category parity: all 4 locales list exactly 6 categories (Foundation/Workflow/Domain/Reference/Meta-Harness/Design)
for loc in en ko ja zh; do
  case $loc in
    en) cats=$(grep -cE '^### (Foundation|Workflow|Domain|Reference|Meta/Harness|Design)' "docs-site/content/$loc/advanced/skill-guide.md");;
    ko) cats=$(grep -cE '^### (Foundation|Workflow|Domain|Reference|Meta/Harness|Design)' "docs-site/content/$loc/advanced/skill-guide.md");;
    ja) cats=$(grep -cE '^### (Foundation|Workflow|Domain|Reference|Meta/Harness|Design)' "docs-site/content/$loc/advanced/skill-guide.md");;
    zh) cats=$(grep -cE '^### (Foundation|Workflow|Domain|Reference|Meta/Harness|Design)' "docs-site/content/$loc/advanced/skill-guide.md");;
  esac
  echo "$loc: $cats of 6 canonical category headers (expected 6)"
done
```

**PASS criterion**: each locale prints `6`. The ja locale must match after its structural rewrite.

### AC-008 — Locale-native idiom (no invented CJK) [SHOULD]

**Given** REQ-007 prohibits invented CJK locale idioms
**When** the run-phase introduces "32" in each locale
**Then** each locale uses its native idiom verifiable by a locale-native reader:

| Locale | Expected idiom | Grep anchor |
|--------|---------------|-------------|
| en | "32 skills" | `32 skill` |
| ko | "32개 스킬" | `32개 스킬` |
| ja | "32スキル" | `32スキル` |
| zh | "32个技能" or "32项技能" | `32个技能\|32项技能` |

**PASS criterion**: each locale's idiom is present and uses the locale's established numeric counter (개/スキル/个/项), not a machine-transliterated form.

### AC-009 — primary-source evidence reproducibility [MUST]

**Given** the canonical count must be traceable to `.claude/skills/`
**When** a reviewer runs the primary-source command
**Then** the output matches the count stated in the docs:

```bash
$ find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l
32
```

**PASS criterion**: command output is `32`, matching the docs count.

### AC-010 — spec-lint clean [MUST]

**Given** the SPEC must pass the lint gate
**When** `moai spec lint .moai/specs/SPEC-V3R6-DOCS-COVERAGE-001/spec.md` runs
**Then** the result is 0 findings.

**PASS criterion**: lint exits 0 with 0 findings.

### AC-011 — en/zh in-body fictional-name elimination [MUST]

> **iter-2 addition (2026-06-18):** independent grep found that en and zh `advanced/skill-guide.md` carry correct 6-canonical-category headers but still reference nonexistent skill names inside conceptual illustrations (Mermaid nodes, explicit-invocation code, ASCII trees, frontmatter examples, auto-load comments, closing callouts). en: 10 matches, zh: 11 matches. This AC eliminates those residuals. ja and ko are covered by AC-006 (structural rewrite eliminates their 37 each).

**Given** en AND zh `advanced/skill-guide.md` carry residual references to nonexistent skill names (`moai-lang-*`, `moai-library-*`, `moai-platform-*`, `moai-framework-*`, `moai-foundation-claude/philosopher/context`) inside conceptual illustrations
**When** the run-phase replaces each fictional reference with a real-skill equivalent (e.g., `moai-lang-python` → `moai-domain-backend` with a note that Python patterns ship via `rules/moai/languages/`) or rephrases to avoid naming a nonexistent skill
**Then** the fictional-name regex returns 0 matches per locale:

```bash
# Per-locale loop over en and zh (iter-2 addition; ja+ko covered separately by AC-006)
for loc in en zh; do
  count=$(grep -cE 'moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context' "docs-site/content/$loc/advanced/skill-guide.md")
  echo "$loc: $count fictional-name matches (expected 0; was en=10 / zh=11 pre-correction)"
done
```

**PASS criterion**: en prints `0` AND zh prints `0`. Any non-zero = FAIL (residual fictional name presented as a shipped skill).

---

## §D.1 Severity Summary

| AC | Severity | Blocking? |
|----|----------|-----------|
| AC-001 (no residual "31") | MUST | YES |
| AC-002 ("32" present) | MUST | YES |
| AC-003 (humanize added) | MUST | YES |
| AC-004 (Domain=9) | MUST | YES |
| AC-005 (sub-count sum=31) | MUST | YES |
| AC-006 (ja AND ko structural rewrite) | MUST | YES |
| AC-007 (4-locale parity) | MUST | YES |
| AC-008 (locale-native idiom) | SHOULD | NO |
| AC-009 (primary-source reproducibility) | MUST | YES |
| AC-010 (spec-lint clean) | MUST | YES |
| AC-011 (en/zh in-body fictional-name elimination) | MUST | YES |

**10 MUST + 1 SHOULD = 11 ACs total.**

---

## §D.2 Edge Cases

1. **`update.md` exists in all 4 locales but only en/zh carry the statusline string**: `update.md` is present in en/ko/ja/zh (independently verified). However, only en/zh `update.md` carry the `31 <skill-count>` statusline string; ko/ja `update.md` do not carry a skill-count claim and require no count correction in this page-family. AC-001/AC-002 must not flag ko/ja `update.md` as missing a count — it legitimately carries no skill-count claim. The coverage map (research.md §3.1) documents this. (iter-2 D3: corrected iter-1 prose "update.md is en/zh-only" which was factually wrong about file presence.)
2. **ja AND ko `what-is-moai-adk.md` asymmetry (ja)**: ja currently only line 267 carries "31" (lines 7/48/652 do not). After reconciliation, ja must gain the "32" count in the same structural positions as en/ko/zh to achieve parity (AC-007). This may require adding count claims to ja lines that previously lacked them — the goal is parity, not minimal-edit. (ko `what-is-moai-adk.md` already carries the claim at L7/48/267/652 and needs only count correction.)
3. **ASCII tree comment**: `core-concepts/what-is-moai-adk.md` L652 (en/ko) / L661 (zh) carries "# 31 skill modules" in a directory-tree code block. This is a factual claim inside a code comment and MUST be corrected (it is in scope per AC-001 digit-boundary grep).
4. **`builder-agents.md` count wording**: "31 built-in skills" → "32 built-in skills". The word "built-in" correctly excludes user-owned harness skills (§E.4), so the count 32 refers to template-shipped skills only.

---

## §D.3 Definition of Done

- [ ] All 10 MUST ACs pass (AC-001 through AC-007, AC-009, AC-010, AC-011)
- [ ] AC-008 (SHOULD) verified or explicitly waived with rationale
- [ ] 4-locale parity confirmed (no locale left behind)
- [ ] primary-source `find` output reproducible
- [ ] spec-lint 0 findings
- [ ] No scope overlap with DOCSITE-001's 6 axes (verified via diff inspection)
- [ ] ja AND ko structural rewrites eliminate all fictional skill names (37 each, 74 total)
- [ ] en AND zh in-body fictional-name cleanup eliminates all residuals (10 + 11, 21 total)
- [ ] Commit message follows `feat(SPEC-V3R6-DOCS-COVERAGE-001): ...` convention

---

## §D.4 Forward-Looking Checks

- **docs-truth.md skill-count axis**: after COVERAGE-001 closes, docs-truth.md MAY be updated to add a §6 "Skill Count (32)" axis (separate sync-phase commit, not in COVERAGE-001 scope). This is recorded as a follow-up, not an AC.
- **`humanize` skill description**: adding `moai-domain-humanize` to the Domain category listing is in scope; writing its full description prose is §E.5 out-of-scope (separate SPEC).
- **Per-skill description audit**: the broader question of whether each skill's description matches its current SKILL.md is §E.5 out-of-scope.
