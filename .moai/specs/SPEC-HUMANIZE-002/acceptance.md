# Acceptance Criteria — SPEC-HUMANIZE-002

Each AC declares its verification **Method**: `mechanical` (deterministic command → PASS/FAIL), `hybrid` (mechanical proxy + one-line manual confirmation), or `manual` (scenario / evidence cross-check). Commands assume repo root `/Users/goos/MoAI/moai-adk-go` and:

```bash
T=internal/template/templates/.claude/skills/moai-domain-humanize
L=.claude/skills/moai-domain-humanize
BASE=39c74d77787621b6645aebe81e470277ba3c97cb   # plan baseline SHA
```

Portability note: script-range greps (`[가-힣]`, kana/Han classes) require GNU grep (`ggrep`) or ripgrep; on stock BSD grep substitute `rg -q '<pattern>'`. Section windowing uses flag-based awk (never `sed -n` line windows).

## §D AC Matrix

| AC | REQ | Verification | Method | Severity |
|----|-----|--------------|--------|----------|
| AC-H2-001 | REQ-H2-001 | `modules/copy-review.md` + `modules/design-copy.md` exist in BOTH trees | mechanical | MUST |
| AC-H2-002 | REQ-H2-001 | Template skill-directory count unchanged (28 @ baseline) | mechanical | MUST |
| AC-H2-003 | REQ-H2-002 | Consolidation: no `*slop*`-named file in skill dir; exactly one copy-review module | mechanical | MUST |
| AC-H2-004 | REQ-H2-003/004/005 | copy-review workflow completeness: 6-stage pipeline + fix-proposal format (original/reason/≥3 alternatives/preferred) + report template | hybrid | MUST |
| AC-H2-005 | REQ-H2-006 | Review-only gate mode present; explicit no-auto-apply statement | hybrid | MUST |
| AC-H2-006 | REQ-H2-007 | `CRS-1`…`CRS-7` all present, each with a before/after example | mechanical | MUST |
| AC-H2-007 | REQ-H2-008 | Per-language dictionary minimums: CR-EN ≥8, CR-KO ≥6, CR-JA ≥4, CR-ZH ≥4 (unique IDs) | mechanical | MUST |
| AC-H2-008 | REQ-H2-009 | Zero "Tier 1/2" severity vocabulary; S1+S2+S3 all in use | mechanical | MUST |
| AC-H2-009 | REQ-H2-010 | Dedup: zero re-definition table rows for existing v1.1.0 IDs; ≥3 cross-references present | mechanical | MUST |
| AC-H2-010 | REQ-H2-011/012 | design-copy genre rules: DCG IDs + landing structure/repair table + short-form rules | mechanical | MUST |
| AC-H2-011 | REQ-H2-013 | 4 per-language adaptation blocks; KO numeric limits not transferred verbatim | hybrid | MUST |
| AC-H2-012 | REQ-H2-014 | Target-script examples per language section (script presence/absence) | mechanical | MUST |
| AC-H2-013 | REQ-H2-015 | JA/ZH grounding traceability (map-to-v1.1.0 or sources-section note) | manual | MUST |
| AC-H2-014 | REQ-H2-016 | Shared machinery reuse: S1/S2/S3 + fact-anchor guard referenced; no parallel grade table | mechanical | MUST |
| AC-H2-015 | REQ-H2-017 | SKILL.md genre-module routing added; Language Routing rows intact | mechanical | MUST |
| AC-H2-016 | REQ-H2-018 | SKILL.md 1.2.0 (metadata + footer) + catalog.yaml 1.2.0 + hash regenerated clean | mechanical | MUST |
| AC-H2-017 | REQ-H2-019 | 4 existing language modules byte-frozen vs baseline (both trees) | mechanical | MUST |
| AC-H2-018 | REQ-H2-020 | Template ↔ local byte parity for the skill dir | mechanical | MUST |
| AC-H2-019 | REQ-H2-021 | Neutrality: 7 forbidden classes → 0 (scoped to skill dir; humanize absent from real leak-test violations) | mechanical | MUST |
| AC-H2-020 | REQ-H2-022 | License: `Apache-2.0` unchanged; zero `MIT License` tokens; im-not-ai credit intact | mechanical | MUST |
| AC-H2-021 | REQ-H2-023 | No programming-language primacy claims in new modules | manual | MUST |
| AC-H2-022 | REQ-H2-003..009/016 | Scenario S1–S3: 4-language review-only gate smoke (KO, EN, JA+ZH) | manual (scenario) | MUST |
| AC-H2-023 | REQ-H2-004/016 | Scenario S4: fact-anchor preservation + placeholder discipline in proposals | manual (scenario) | MUST |
| AC-H2-024 | REQ-H2-013 | Scenario S5: short-form genre adaptation — EN measured natively, not by KO char limit | manual (scenario) | MUST |

## §D.1 Verification Commands

### AC-H2-001 — new modules exist in both trees
```bash
for f in "$T"/modules/copy-review.md "$T"/modules/design-copy.md \
         "$L"/modules/copy-review.md "$L"/modules/design-copy.md; do
  test -f "$f" && echo "OK $f" || echo "MISSING $f"
done   # PASS: 4× OK
```

### AC-H2-002 — skill-directory count unchanged
```bash
ls -d internal/template/templates/.claude/skills/*/ | wc -l   # PASS: 28 (baseline @ $BASE)
```

### AC-H2-003 — consolidation, no source-named files
```bash
find "$T" "$L" -iname '*slop*' -o -iname '*cd-slop*' | grep . && echo FAIL || echo PASS
ls "$T"/modules/   # PASS: exactly {korean,english,japanese,chinese,copy-review,design-copy}.md
```

### AC-H2-004 — copy-review workflow completeness (hybrid)
```bash
f="$T"/modules/copy-review.md
grep -qiE 'pipeline|stage' "$f" && grep -qi 'language detection' "$f" \
  && grep -qi 'context' "$f" && grep -qiE 'report' "$f" \
  && grep -qiE 'alternativ' "$f" && grep -qiE 'preferred' "$f" && echo TOKENS-PASS || echo TOKENS-FAIL
```
- Manual confirm: 6 stages enumerated in order; fix-proposal format shows original / reason / ≥3 alternatives / preferred with justification; report template carries a severity summary table.

### AC-H2-005 — review-only gate mode (hybrid)
```bash
f="$T"/modules/copy-review.md
grep -qi 'review-only' "$f" && grep -qiE 'auto-appl' "$f" && echo TOKENS-PASS || echo TOKENS-FAIL
```
- Manual confirm: the mode statement says the gate detects + proposes and does NOT auto-apply (user reviews before application), and is distinguished from the skill's default rewrite mode.

### AC-H2-006 — CRS playbook
```bash
f="$T"/modules/copy-review.md; ok=1
for i in 1 2 3 4 5 6 7; do grep -q "CRS-$i" "$f" || { echo "MISSING CRS-$i"; ok=0; }; done
[ $ok -eq 1 ] && echo PASS || echo FAIL
# Before/after presence proxy (manual spot-check accompanies):
grep -ciE 'before|after|→' "$f"
```

### AC-H2-007 — per-language dictionary minimums
```bash
f="$T"/modules/copy-review.md
for LNG in EN:8 KO:6 JA:4 ZH:4; do
  lang=${LNG%%:*}; min=${LNG##*:}
  n=$(grep -oE "\bCR-$lang-[0-9]+\b" "$f" | sort -u | wc -l | tr -d ' ')
  [ "$n" -ge "$min" ] && echo "$lang PASS ($n ≥ $min)" || echo "$lang FAIL ($n < $min)"
done
```

### AC-H2-008 — severity remap, no Tier vocabulary
```bash
grep -nEi '\btier ?[12]\b' "$T"/modules/copy-review.md "$T"/modules/design-copy.md \
  && echo "TIER-VOCAB FAIL" || echo "TIER-VOCAB PASS (0)"
f="$T"/modules/copy-review.md
grep -q 'S1' "$f" && grep -q 'S2' "$f" && grep -q 'S3' "$f" && echo "SEVERITY PASS" || echo "SEVERITY FAIL"
```

### AC-H2-009 — dedup: no re-definitions + cross-reference evidence
```bash
# NEGATIVE: no table row in the NEW modules DEFINES an existing v1.1.0 ID (first cell)
grep -nE '^\|[[:space:]]*(ENC-[0-9]|JA-1[0-4]|CN-[LMNOPQ]|A-2[0-5]|L-[1-8]|M-[1-3])[[:space:]]*\|' \
  "$T"/modules/copy-review.md "$T"/modules/design-copy.md && echo "REDEF FAIL" || echo "REDEF PASS (0)"
# POSITIVE: ≥3 distinct existing IDs are cross-referenced (dedup actually happened)
grep -ohE '\b(ENC-[0-9]|JA-1[0-4]|CN-[LMNOPQ]|A-2[0-5]|L-[1-8]|M-[1-3])\b' \
  "$T"/modules/copy-review.md "$T"/modules/design-copy.md | sort -u | wc -l   # PASS: ≥ 3
```

### AC-H2-010 — design-copy genre rules
```bash
f="$T"/modules/design-copy.md
grep -qE '\bDCG(-[A-Z]{2})?-[0-9]+\b' "$f" && echo "ID PASS" || echo "ID FAIL"
grep -qiE 'headline' "$f" && grep -qiE 'CTA' "$f" && echo "LANDING PASS" || echo "LANDING FAIL"
grep -qiE 'card|slide' "$f" && grep -qiE 'cover' "$f" && echo "SHORTFORM PASS" || echo "SHORTFORM FAIL"
```

### AC-H2-011 — 4-language adaptation blocks (hybrid)
```bash
f="$T"/modules/design-copy.md
for lang in Korean English Japanese Chinese; do
  grep -qE "^#+ .*$lang" "$f" && echo "$lang PASS" || echo "$lang FAIL"
done
```
- Manual confirm: language-dependent parameters (length limits, ending rules) state a native measure per language; no EN/JA/ZH block repeats a Hangul character count verbatim.

### AC-H2-012 — target-script examples per language section
```bash
f="$T"/modules/copy-review.md
sec() { awk -v h="$1" 'BEGIN{f=0} $0 ~ "^#+ .*"h {f=1; next} /^## /{f=0} f' "$f"; }
sec Korean   | rg -q '[가-힣]'        && echo "KO-script PASS" || echo "KO-script FAIL"
sec Japanese | rg -q '[ぁ-んァ-ヶ]'    && echo "JA-script PASS" || echo "JA-script FAIL"
sec Japanese | rg -q '[가-힣]'        && echo "JA-hangul FAIL" || echo "JA-hangul PASS (none)"
sec Chinese  | rg -q '[\p{Han}]'      && echo "ZH-script PASS" || echo "ZH-script FAIL"
sec Chinese  | rg -q '[가-힣ぁ-んァ-ヶ]' && echo "ZH-purity FAIL" || echo "ZH-purity PASS (none)"
sec English  | rg -q '[가-힣ぁ-んァ-ヶ]' && echo "EN-purity FAIL" || echo "EN-purity PASS (none)"
```
(Mechanically proves JA/ZH sections are not Korean-catalogue clones at the script level; semantic independence is AC-H2-013.)

### AC-H2-013 — JA/ZH grounding traceability (manual)
Reviewer walks every `CR-JA-*` / `CR-ZH-*` entry: each either (a) cross-references an existing v1.1.0 tell (JA-10…14 / CN-L…Q / prose), or (b) carries a grounding note resolving to the module's sources section. Reviewer additionally compares each JA/ZH entry against the KO dictionary to confirm it is not a translated KO clone (language-native phenomenon, native example).

### AC-H2-014 — shared machinery reuse
```bash
grep -qiE 'fact.anchor' "$T"/modules/copy-review.md && echo "ANCHOR PASS" || echo "ANCHOR FAIL"
# NEGATIVE: no parallel grade table in new modules (grades live in SKILL.md)
grep -nE '^\|[[:space:]]*\*{0,2}[ABCD]\*{0,2}[[:space:]]*\|' \
  "$T"/modules/copy-review.md "$T"/modules/design-copy.md && echo "GRADE-TABLE FAIL" || echo "GRADE-TABLE PASS (0)"
```

### AC-H2-015 — SKILL.md routing integration
```bash
S="$T"/SKILL.md
grep -q 'modules/copy-review.md' "$S" && grep -q 'modules/design-copy.md' "$S" && echo "ROUTE PASS" || echo "ROUTE FAIL"
for m in korean english japanese chinese; do
  grep -q "modules/$m.md" "$S" || echo "LANG-TABLE MISSING $m"
done   # PASS: no MISSING lines
```

### AC-H2-016 — version bumps + hash
```bash
grep -qE 'version:[[:space:]]*"?1\.2\.0"?' "$T"/SKILL.md && echo "SKILL-VER PASS" || echo "SKILL-VER FAIL"
grep -A5 'name: moai-domain-humanize' internal/template/catalog.yaml | grep -q 'version: 1.2.0' \
  && echo "CATALOG-VER PASS" || echo "CATALOG-VER FAIL"
# Hash regenerated by make build; after the M3 commit:
git status --porcelain internal/template/catalog.yaml   # PASS: empty (committed hash == recomputed)
```

### AC-H2-017 — existing language modules byte-frozen
```bash
git diff --stat "$BASE"..HEAD -- \
  "$T"/modules/korean.md "$T"/modules/english.md "$T"/modules/japanese.md "$T"/modules/chinese.md \
  "$L"/modules/korean.md "$L"/modules/english.md "$L"/modules/japanese.md "$L"/modules/chinese.md
# PASS: empty output (0 files changed)
```

### AC-H2-018 — template ↔ local parity
```bash
diff -rq "$T" "$L" && echo "PARITY PASS" || echo "PARITY FAIL"
```

### AC-H2-019 — neutrality: 7 forbidden classes → 0 (scoped)
```bash
DIR="$T"; fail=0
grep -rEn 'SPEC-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[0-9]{3}' "$DIR" && fail=1                  # (1) SPEC IDs
grep -rEn '\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b' "$DIR" && fail=1                          # (2) REQ/AC tokens
grep -rhoE '\b[0-9a-f]{7,40}\b' "$DIR" | grep -E '[a-f]' && fail=1                       # (3) SHAs (benign all-hex words: eyeball)
awk 'f>=2; /^---[[:space:]]*$/{f++}' "$DIR/SKILL.md" | grep -En '20[0-9]{2}-[0-9]{2}-[0-9]{2}' && fail=1  # (4) body dates
for m in "$DIR"/modules/*.md; do grep -En '20[0-9]{2}-[0-9]{2}-[0-9]{2}' "$m" && fail=1; done
grep -rEn '(moai-writer|moai-coworker|moai-marketer|moai-designer|moai-officer|moai-code):' "$DIR" && fail=1  # (5) plugin namespaces
grep -rEn '3-point sync|plugin\.json|marketplace metadata' "$DIR" && fail=1              # (6) sync comments
grep -rEn '\.moai/(reports|research)/' "$DIR" && fail=1                                  # (7) internal paths
[ $fail -eq 0 ] && echo "NEUTRALITY PASS" || echo "NEUTRALITY FAIL"
# Advisory (not whole-repo-green gate): humanize dir absent from the real leak-test violation list
go test ./internal/template/ -run TestTemplateNoInternalContentLeak 2>&1 \
  | grep 'moai-domain-humanize' && echo "HUMANIZE LEAK — investigate" || echo "humanize dir clean"
```

### AC-H2-020 — license unchanged
```bash
grep -rn 'MIT License' "$T" && echo "MIT FAIL" || echo "MIT PASS (0)"
grep -qE '^license:[[:space:]]*Apache-2\.0[[:space:]]*$' "$T"/SKILL.md && echo "LICENSE PASS" || echo "LICENSE FAIL"
grep -rqi 'im-not-ai' "$T"/SKILL.md && echo "CREDIT PASS" || echo "CREDIT FAIL"
```

### AC-H2-021 — no programming-language primacy (manual)
Reviewer scans both new modules: no programming language is named as primary/preferred/default; code examples (if any) are absent or language-agnostic. Natural-language names (Korean/English/Japanese/Chinese) are the expected subject matter.

## §D.2 Scenarios (Given-When-Then)

### S1 — Korean landing hero through the review-only gate (AC-H2-022)
- **Given** the KO hero "혁신적인 AI 기반의 차세대 마케팅 자동화 — 당신의 마케팅을 한 차원 높여 줄 솔루션" (slot-formula hits + dash-contrast),
- **When** copy-review runs in review-only mode with `modules/korean.md` + `modules/copy-review.md`,
- **Then** the report flags `CR-KO-*` formula findings AND cites `korean.md` M-1 for the dash-contrast (cross-reference, not a re-defined entry); each finding carries original / reason / ≥3 alternatives / preferred; NO rewritten text is auto-applied.

### S2 — English hero bundle (AC-H2-022)
- **Given** the EN bundle "Reimagine the way your team collaborates. Powered by AI. Trusted by thousands of teams.",
- **When** copy-review runs in review-only mode,
- **Then** formula findings fire from `CR-EN-*` (with ENC cross-references where the tell is already catalogued, e.g., aspirational-verb family → ENC-1), the unfounded-stat claim is flagged, and every alternative uses a placeholder (`[N] teams`) rather than an invented number.

### S3 — JA + ZH smoke (AC-H2-022)
- **Given** the JA copy 「業務効率化を実現します：あなたのチームのための次世代ツール」 and the ZH copy "专为增长团队打造的下一代营销平台，开启增长之旅",
- **When** copy-review runs per language,
- **Then** the JA report cross-references JA-13 (formulaic value-prop) and JA-11 (English-style colon) plus any native `CR-JA-*` hits; the ZH report cross-references CN-N (slot-fill landing template) plus native `CR-ZH-*` hits; each emits a severity-tagged report.

### S4 — Fact-anchor preservation (AC-H2-023)
- **Given** the KO copy "월 9,900원, 3일 안에 시작 — 지금까지 없던 자동화" (real price + real duration + formula tell),
- **When** the gate proposes fixes for the formula tell,
- **Then** every alternative preserves 9,900원 and 3일 character-intact; the formula span is the only rewritten portion; no invented specifics appear.

### S5 — Short-form genre adaptation (AC-H2-024)
- **Given** a KO card-news cover exceeding the KO cover-economy guidance and an EN slide cover of similar visual density,
- **When** design-copy genre rules run per language,
- **Then** the KO cover is flagged by the KO-native measure, and the EN cover is judged by the EN-native measure (word-based) — the KO character limit is NOT applied to the EN cover.

## §D.4 Quality Gate / Definition of Done

- [ ] All 24 AC checks PASS at their declared Method (16 mechanical, 3 hybrid, 5 manual).
- [ ] `make build` clean; catalog.yaml humanize entry 1.2.0 with regenerated hash; porcelain clean post-commit.
- [ ] `diff -rq` template vs local skill dir → identical.
- [ ] 8 existing language-module files: zero diff vs baseline `39c74d777`.
- [ ] Neutrality 7-class sweep → 0; humanize dir absent from real leak-test violations.
- [ ] License gates: zero MIT tokens; `license: Apache-2.0` unchanged; im-not-ai credit intact.
- [ ] Scenarios S1–S5 each recorded with a graded outcome and no meaning/anchor drift.
- [ ] No Go source modified; no flat files in `.moai/specs/`; commits Conventional + specific-path staged.

## §D.5 Traceability

REQ-H2-001→AC-001/002; 002→003; 003→004,022; 004→004,023; 005→004; 006→005,022; 007→006; 008→007,022; 009→008; 010→009; 011→010; 012→010; 013→011,024; 014→012; 015→013; 016→014,023; 017→015; 018→016; 019→017; 020→018; 021→019; 022→020; 023→021. Method mix (24 checks): 16 mechanical, 3 hybrid (AC-004/005/011), 5 manual (AC-013/021/022/023/024).
