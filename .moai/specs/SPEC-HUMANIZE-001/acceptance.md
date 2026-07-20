# Acceptance Criteria — SPEC-HUMANIZE-001

Each AC declares its verification **Method** in the matrix: `mechanical` (a deterministic command yields PASS/FAIL), `hybrid` (a mechanical proxy plus a one-line manual confirmation where explanatory prose defeats a pure grep), or `manual` (a run-through scenario or evidence cross-check with no single deterministic command). Grep commands assume repo root `/Users/goos/MoAI/moai-adk-go`. The skill files are the five existing files (NO NOTICE.md is created — v0.3.0 decision; REQ-HUM-015 instead REMOVES the dangling NOTICE.md pointers):
`internal/template/templates/.claude/skills/moai-domain-humanize/{SKILL.md,modules/korean.md,modules/english.md,modules/japanese.md,modules/chinese.md}` (verify the mirror at `.claude/skills/moai-domain-humanize/...` too after sync). Note: only `SKILL.md` carries YAML frontmatter; the four modules are plain markdown with no frontmatter fence — this matters for the AC-HUM-008 date check.

## §D AC Matrix

| AC | REQ | Verification | Method | Severity |
|----|-----|--------------|--------|----------|
| AC-HUM-001 | REQ-HUM-001 | Re-authored korean.md carries prose A–J + copy layer (A-20…A-25, L-1…L-8, M-1…M-3); no port claim | mechanical | MUST |
| AC-HUM-002 | REQ-HUM-002 | English copy layer present (ENC-1…ENC-9) | mechanical | MUST |
| AC-HUM-003 | REQ-HUM-003 | Japanese copy layer present (JA-10…JA-14) | mechanical | MUST |
| AC-HUM-004 | REQ-HUM-004 | Chinese copy layer present (CN-L…CN-Q) | mechanical | MUST |
| AC-HUM-005 | REQ-HUM-005/006 | SKILL.md copy-mode fact-anchor guard + dual grading tables | hybrid | MUST |
| AC-HUM-006a | REQ-HUM-007(b) | JA-10 table-row severity = S2 + frequency framing AND row severity ≠ S1 (positive+negative) | mechanical | MUST |
| AC-HUM-006b | REQ-HUM-007(a) | No ENC table row defines a removable fragment tell (negative) AND "fragment" framed as high-false-positive (positive) | mechanical | MUST |
| AC-HUM-006c | REQ-HUM-007(c) | 排比/对偶 boundary states content-first/template-first (positive) + manual confirm not a bare-count decisive rule | hybrid | MUST |
| AC-HUM-007 | REQ-HUM-008 | templates/ and local .claude/ byte-identical after make build + sync | mechanical | MUST |
| AC-HUM-008 | REQ-HUM-009 | 6 forbidden neutrality classes → 0 (date body-scoped); humanize dir absent from real leak-test | mechanical | MUST |
| AC-HUM-009 | REQ-HUM-010 | Instruction prose English; examples in target language | manual | MUST |
| AC-HUM-010 | REQ-HUM-011 | SKILL.md user-invocable/allowed-tools/attribution preserved; metadata.version bumped | mechanical | MUST |
| AC-HUM-011 | REQ-HUM-012 | Every EN/JA/ZH catalogue pattern traces to a research.md verified source; hypotheses quarantined | manual | MUST |
| AC-HUM-012 | REQ-HUM-013 | catalog.yaml version 1.1.0 + hash matches recompute + routing table updated | mechanical | MUST |
| AC-HUM-013 | REQ-HUM-005/006 | 8-sample matrix (4 lang × prose/copy) each yields a grade + no meaning drift | manual (scenario) | MUST |
| AC-HUM-014 | REQ-HUM-007 | 3 false-positive guards (EN slide fragment / JA one 体言止め / ZH one 排比) NOT flagged | manual (scenario) | MUST |
| AC-HUM-015 | REQ-HUM-014 | Retired-Python known limitation + conservative-judgment instruction present | mechanical | MUST |
| AC-HUM-016 | REQ-HUM-015 | Zero `NOTICE.md` references in the skill dir; courtesy credit present, no license claim | mechanical | MUST |
| AC-HUM-017 | REQ-HUM-016 | Zero `MIT License` tokens in the skill dir; `license: Apache-2.0` unchanged | mechanical | MUST |

## §D.1 Verification commands (Method annotated per AC — see matrix)

### AC-HUM-001 — Re-authored korean.md: prose + copy layer present, no port claim
```bash
f=internal/template/templates/.claude/skills/moai-domain-humanize/modules/korean.md
# Copy-layer + prose-layer category presence (IDs align with the maintainer's taxonomy)
grep -Eq 'A-20' "$f" && grep -Eq 'A-25' "$f" && grep -Eq 'L-1' "$f" && grep -Eq 'L-8' "$f" && grep -Eq 'M-1' "$f" && grep -Eq 'M-3' "$f" \
  && grep -Eq 'A-1' "$f" && echo "CATEGORIES PASS" || echo "CATEGORIES FAIL"
# NEGATIVE (v0.3.0): the re-authored module must NOT claim to be a port of im-not-ai
grep -niE 'faithful port|ported from' "$f" && echo "PORT CLAIM FAIL" || echo "NO PORT CLAIM PASS"
```

### AC-HUM-002 — English copy layer present
```bash
grep -Eq 'ENC-1' internal/template/templates/.claude/skills/moai-domain-humanize/modules/english.md && \
grep -Eq 'ENC-9' internal/template/templates/.claude/skills/moai-domain-humanize/modules/english.md && echo PASS || echo FAIL
```

### AC-HUM-003 — Japanese copy layer present
```bash
grep -Eq 'JA-10' internal/template/templates/.claude/skills/moai-domain-humanize/modules/japanese.md && \
grep -Eq 'JA-14' internal/template/templates/.claude/skills/moai-domain-humanize/modules/japanese.md && echo PASS || echo FAIL
```

### AC-HUM-004 — Chinese copy layer present
```bash
grep -Eq 'CN-L' internal/template/templates/.claude/skills/moai-domain-humanize/modules/chinese.md && \
grep -Eq 'CN-Q' internal/template/templates/.claude/skills/moai-domain-humanize/modules/chinese.md && echo PASS || echo FAIL
```

### AC-HUM-005 — SKILL.md copy-mode guard + dual grading
```bash
f=internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md
grep -qi 'fact.anchor' "$f" && grep -qi 'copy mode' "$f" && grep -qi 'prose' "$f" && echo PASS || echo FAIL
# Manual: confirm TWO grading tables (prose-mode + copy-mode) exist in the shared section.
```

### AC-HUM-006a — JA-10 frequency-gated S2, not presence-S1 (mechanical: positive + negative)
Authoring contract: the JA-10 **detection-table row** severity cell reads `S2` with a frequency/over-reliance framing; the presence-vs-S1 discussion lives in the SEPARATE 体言止め boundary-analysis subsection, NOT the table row (so the negative grep is not defeated by explanatory prose).
```bash
JA=internal/template/templates/.claude/skills/moai-domain-humanize/modules/japanese.md
row=$(grep -E '^\|.*JA-10' "$JA")                       # the detection-table row for JA-10
# POSITIVE: row carries S2 + a frequency/over-reliance/ratio framing
echo "$row" | grep -qE '\bS2\b' \
  && grep -niE 'JA-10|体言止め' "$JA" | grep -qiE 'over-reliance|frequency|ratio|過剰|3 consecutive|≥ *3|default' \
  && POS=1 || POS=0
# NEGATIVE: the JA-10 table row's severity cell is NOT S1 (presence-based)
echo "$row" | grep -qE '\bS1\b' && NEG=1 || NEG=0
[ "$POS" = 1 ] && [ "$NEG" = 0 ] && echo PASS || echo FAIL
```

### AC-HUM-006b — English no removable fragment category (mechanical: negative + positive)
```bash
EN=internal/template/templates/.claude/skills/moai-domain-humanize/modules/english.md
# NEGATIVE: no ENC-N detection-table row defines a removable fragment/verbless/predicate-less/noun-ending HEADLINE
#   category. The family token must be ADJACENT (either side) to a headline-shaped noun (headline|title|slide|fragment);
#   a bare "verbless"/"fragment" token in a non-headline context does NOT trip.
# AUTHORING-CONTRACT NOTE (mirrors the AC-006a table-row note — records the ENC-7 collision as the reason for scoping):
#   ENC-7's own definition is "bland verbless: `Get Started`, `Submit`" (research.md §2, ENC-7) — a CTA-microcopy tell,
#   NOT a removable headline category. A bare-`verbless` alternation false-FAILs that legitimate row (verified: NEG flips
#   1→0 under the scoped pattern). Likewise ENC-3's "three parallel fragments" is a tricolon description, not a
#   fragment-headline category. The run-phase author MUST keep ENC-7's "bland verbless: Get Started/Submit" phrasing intact
#   (safe under this pattern) and MUST NOT author any ENC row defining a removable predicate-less/verbless/fragment
#   HEADLINE/TITLE/SLIDE category (Korean M-2 does not transfer — REQ-HUM-007a).
NEGPAT='(predicate-less|noun-ending|verbless|fragment)[ -]+(headline|title|slide|fragment)|(headline|title|slide)[ -]+(predicate-less|noun-ending|verbless|fragment)'
grep -E '^\|.*ENC-[0-9]' "$EN" | grep -qiE "$NEGPAT" && NEG=1 || NEG=0
# POSITIVE: "fragment"/"verbless" is addressed as a high-false-positive / natural register (not a removable category)
grep -niE 'fragment|verbless' "$EN" | grep -qiE 'false.positive|natural|NOT a (standalone|removable)|human-authored' && POS=1 || POS=0
[ "$NEG" = 0 ] && [ "$POS" = 1 ] && echo PASS || echo FAIL
```

### AC-HUM-006c — Chinese 排比/对偶 content-first boundary (hybrid: mechanical positive + manual negative)
```bash
CN=internal/template/templates/.claude/skills/moai-domain-humanize/modules/chinese.md
# POSITIVE (mechanical): the 排比/对偶 boundary is stated as content-first / content-driven / template-first / info-density
grep -niE '排比|对偶' "$CN" | grep -qiE 'content-first|content-driven|template-first|信息密度|稀释|distinct concrete fact' && echo "POS PASS" || echo "POS FAIL"
```
- **Manual negative** (grep-defeating): confirm the decisive test is NOT a bare occurrence count. A pure `3+` grep is unreliable because "count is a weak signal" / "flag only when blocks stack (3+)" legitimately contains `3+`. A reviewer confirms the 对偶/排比 boundary subsection makes content-drivenness (not count) the decisive criterion, with count demoted to a weak signal.

### AC-HUM-007 — byte-identity after make build + sync
```bash
diff -rq internal/template/templates/.claude/skills/moai-domain-humanize/ .claude/skills/moai-domain-humanize/ \
  && echo "PASS IDENTICAL" || echo "FAIL DIFFER"
```

### AC-HUM-008 — template neutrality: 6 forbidden classes → 0 (reachable; mirrors the real CI guard's forbidden classes, scoped to the humanize dir)

Why scoped, not whole-repo: `go test ./internal/template/ -run TestTemplateNoInternalContentLeak` is currently RED for a **pre-existing, unrelated** leak (4 SPEC-ID tokens in `templates/.claude/rules/moai/core/agent-common-protocol.md`), so REQ-HUM-009 is verified against the humanize skill dir only — never whole-repo green. Why the date class is body-scoped: the real guard's narrow tier does NOT check dates (its `S1-internal-date` class is strict-tier only, gated behind `MOAI_TEMPLATE_LEAK_STRICT=1`), so a legitimate `SKILL.md` frontmatter `created:`/`updated:` date MUST NOT count. Classes (2) and (3) mirror the guard's `S3-req-ac-token-any-prefix` and `S2` (`requireHexLetter`) classes.

```bash
DIR=internal/template/templates/.claude/skills/moai-domain-humanize
fail=0
# (1) internal SPEC IDs — anywhere in the skill (frontmatter carries none). Broader than the real guard's
#     narrow C1 family on purpose: it must also catch a generic SPEC-HUMANIZE-001 leak in the ported content.
grep -rEn 'SPEC-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[0-9]{3}' "$DIR" && fail=1
# (2) internal REQ/AC tokens — mirrors the guard's S3-req-ac-token-any-prefix skill-body class
grep -rEn '\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b' "$DIR" && fail=1
# (3) commit SHAs (D6 — the class the v0.1.0 AC omitted) — 7-40 hex requiring >=1 [a-f] letter (mirrors guard S2 requireHexLetter).
#     KNOWN benign false-positive: an all-hex English word >=7 chars (e.g. "defaced") matches; disambiguate by eye if it fires.
grep -rhoE '\b[0-9a-f]{7,40}\b' "$DIR" | grep -E '[a-f]' && fail=1
# (4) internal version annotations (Korean "vN.N 신규/확장/신설" style)
grep -rEn 'v[0-9]+\.[0-9]+ *(신규|확장|신설)' "$DIR" && fail=1
# (5) .moai/reports research paths
grep -rEn '\.moai/reports/research' "$DIR" && fail=1
# (6) internal work dates. SKILL.md is the ONLY file with YAML frontmatter, whose created:/updated: are
#     legitimate skill metadata → strip the frontmatter block and grep the BODY only. The four modules carry
#     NO frontmatter, so ANY ISO date there is a body/work-date leak → grep the whole file.
awk 'f>=2; /^---[[:space:]]*$/{f++}' "$DIR/SKILL.md" | grep -En '20[0-9]{2}-[0-9]{2}-[0-9]{2}' && fail=1
for f in "$DIR"/modules/*.md; do
  [ -f "$f" ] || continue
  grep -En '20[0-9]{2}-[0-9]{2}-[0-9]{2}' "$f" && fail=1
done
[ "$fail" -eq 0 ] && echo "NEUTRALITY PASS" || echo "NEUTRALITY FAIL"
# Expected on a correctly-authored tree: NEUTRALITY PASS (reachable to 0 — a normal frontmatter date no longer counts).
```

Advisory tie to the real CI guard (NOT a whole-repo-green gate): the humanize dir MUST be absent from the guard's violation list; the pre-existing `agent-common-protocol.md` failure is unrelated and out of this SPEC's scope.
```bash
go test ./internal/template/ -run TestTemplateNoInternalContentLeak 2>&1 \
  | grep 'moai-domain-humanize' && echo "HUMANIZE LEAK — investigate" || echo "humanize dir clean (unrelated pre-existing failures ignored)"
```

### AC-HUM-009 — English instruction prose / target-language examples
```bash
# Spot-check: each module's copy-layer instruction sentences are English; before/after examples are in the target script.
# Manual review: korean.md examples in Hangul, japanese.md in kana/kanji, chinese.md in Simplified Han, english.md in English.
```

### AC-HUM-010 — frontmatter preservation (scoped, per amended REQ-HUM-011) + deliberate changes
```bash
f=internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md
# PRESERVED verbatim:
grep -q 'user-invocable: false' "$f" && grep -q 'allowed-tools:' "$f" && grep -qi 'im-not-ai' "$f" && echo "PRESERVE PASS" || echo "PRESERVE FAIL"
# DELIBERATELY changed (must NOT remain at the old value): metadata.version bumped to 1.1.0
grep -qE 'version:[[:space:]]*"?1\.1\.0"?' "$f" && echo "VERSION BUMP PASS" || echo "VERSION BUMP FAIL"
# (license is UNCHANGED — verified by AC-HUM-017; the `updated:` date change is expected, not a defect.
#  The `im-not-ai` token above now matches the courtesy credit, not the old attribution — both satisfy the grep.)
```

### AC-HUM-011 — evidence traceability (Method: manual — no single deterministic command)
This check is irreducibly manual: a reviewer cross-checks that every `ENC-N` / `JA-1x` / `CN-[L-Q]` catalogue row traces (directly or via `research.md` §6 Verified Sources) to a fetched-and-confirmed URL, and that no `research.md` §5 quarantined hypothesis appears in any module's main catalogue.
```bash
# Advisory aid only (NOT a pass/fail): list the copy-layer category IDs, then eyeball each against research.md §6.
grep -roE '\b(ENC-[0-9]|JA-1[0-4]|CN-[L-Q])\b' internal/template/templates/.claude/skills/moai-domain-humanize/modules/ | sort -u
# The reviewer confirms each ID has a Verified-Sources backing in research.md and no quarantined-hypothesis leak.
```

### AC-HUM-012 — catalog version + hash
```bash
grep -A5 'name: moai-domain-humanize' internal/template/catalog.yaml | grep -q 'version: 1.1.0' && echo "VERSION PASS" || echo "VERSION FAIL"
# HASH: after `make build` (gen-catalog-hashes --all), the committed hash must equal the freshly recomputed hash (no dirty diff on catalog.yaml hash line).
```

### AC-HUM-013 — 8-sample verification (see §D.2 scenarios)
```bash
# Not a single grep: run the 8-sample matrix per §D.2. Each sample → a grade (A/B/C/D) + a meaning-drift check = "no drift".
```

### AC-HUM-014 — false-positive guards (see §D.3 scenarios)
```bash
# Three guard samples, each must be graded NOT-flagged for the guarded pattern.
```

### AC-HUM-015 — retired-Python limitation + conservative instruction
```bash
f=internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md
grep -qi 'conservativ' "$f" && grep -Eqi '30%|50%|threshold' "$f" && echo PASS || echo FAIL
# STRICT: SKILL.md instructs conservative judgment near the change-rate thresholds; spec.md §E records the metric-loss limitation.
```

### AC-HUM-016 — attribution cleanup: zero NOTICE.md references + courtesy credit (REQ-HUM-015, rewritten v0.3.0)
```bash
DIR=internal/template/templates/.claude/skills/moai-domain-humanize
# (a) NO dangling NOTICE.md pointers anywhere in the skill (and no NOTICE.md file)
grep -rn 'NOTICE.md' "$DIR" && echo "NOTICE REF FAIL" || echo "NOTICE REF PASS (0 matches)"
test -f "$DIR/NOTICE.md" && echo "NOTICE FILE FAIL (must not exist)" || echo "NOTICE FILE PASS (absent)"
# Courtesy credit present (inspiration-only, names im-not-ai; no license claim — license absence verified by AC-HUM-017)
grep -rqi 'inspired by' "$DIR" && grep -rqi 'im-not-ai' "$DIR" && echo "CREDIT PASS" || echo "CREDIT FAIL"
```

### AC-HUM-017 — license unchanged: Apache-2.0 as-is, zero MIT-license tokens (REQ-HUM-016, simplified v0.3.0)
```bash
DIR=internal/template/templates/.claude/skills/moai-domain-humanize
S="$DIR/SKILL.md"
# (b) NO "MIT License" claim anywhere in the skill (the courtesy credit carries no license token)
grep -rn 'MIT License' "$DIR" && echo "MIT TOKEN FAIL" || echo "MIT TOKEN PASS (0 matches)"
# (c) frontmatter license field unchanged: exactly Apache-2.0, no compound expression
grep -qE '^license:[[:space:]]*Apache-2\.0[[:space:]]*$' "$S" && echo "LICENSE UNCHANGED PASS" || echo "LICENSE UNCHANGED FAIL"
```

## §D.2 8-Sample Verification Scenarios (Given-When-Then)

The verification surface requires 4 languages × {one prose text, one copy text} = 8 samples run through detect → rewrite → grade.

### Scenario 1 — Korean prose (baseline regression)
- **Given** a Korean prose paragraph carrying prose tells (번역투 `~을 통해`, hedging `~할 수 있을 것으로 보인다`),
- **When** the skill runs in prose mode with `modules/korean.md`,
- **Then** it flags the prose tells, rewrites surgically, and returns a prose-mode grade (A/B/C/D) with numbers/names preserved (no meaning drift).

### Scenario 2 — Korean copy
- **Given** a Korean SaaS headline "자동화는 24시간 굴러갑니다 — 복붙에서 위임으로" (A-20 굴러가다 + M-1 dash-contrast + M-3 X에서 Y로),
- **When** the skill runs in copy mode,
- **Then** it flags A-20/M-1/M-3, rewrites to a fact-anchored human headline, and returns a copy-mode grade with the core promise/benefit preserved (no fact-anchor loss).

### Scenario 3 — English prose
- **Given** an English paragraph with EN-A (`delve`, `tapestry`) + EN-B negative parallelism,
- **When** the skill runs in prose mode with `modules/english.md`,
- **Then** it flags the prose tells and returns a prose-mode grade with no meaning drift.

### Scenario 4 — English copy
- **Given** an English hero section "In today's fast-paced digital world, unleash your potential. It's not just a tool — it's a movement. Fast. Simple. Scalable." (ENC-4 + ENC-1 + ENC-2 + ENC-3),
- **When** the skill runs in copy mode,
- **Then** it flags ENC-4/ENC-1/ENC-2/ENC-3, resolves the S1 (ENC-2/ENC-4) first, rewrites to concrete-claim copy, and grades — preserving any real numbers.

### Scenario 5 — Japanese prose
- **Given** a Japanese paragraph with JA-01 (`〜することができます`) + JA-03 (`これにより`),
- **When** the skill runs in prose mode with `modules/japanese.md`,
- **Then** it flags the prose tells and returns a prose-mode grade with no meaning drift.

### Scenario 6 — Japanese copy
- **Given** a Japanese copy block with three consecutive 体言止め headlines + JA-11 English-style colon,
- **When** the skill runs in copy mode,
- **Then** JA-10 fires on the ≥3-consecutive 体言止め (over-reliance), JA-11 fires on the colon, and the rewrite varies the endings while preserving the offer.

### Scenario 7 — Chinese prose
- **Given** a Chinese paragraph with CN-A (首先…其次…综上所述) + CN-C (赋能/闭环),
- **When** the skill runs in prose mode with `modules/chinese.md`,
- **Then** it flags the prose tells and returns a prose-mode grade with no meaning drift.

### Scenario 8 — Chinese copy
- **Given** a Chinese landing headline "这不仅仅是一双跑鞋，而是对自律生活方式的承诺" (CN-L negation-contrast) + "让我们携手共创美好未来" (CN-O forced elevation),
- **When** the skill runs in copy mode,
- **Then** CN-L (S2 headline) + CN-O (S1) fire, the abstract second clause is replaced with the concrete fact it hid, and any real spec numbers are preserved.

## §D.3 False-Positive Guard Scenarios (must NOT flag)

### Guard 1 — English human slide fragment
- **Given** a human-authored English slide title fragment "Q1 Revenue" / "Our Approach" / "Why It Matters",
- **When** the skill runs the English copy layer,
- **Then** the fragment is NOT flagged as a removable tell (English headlines are natively fragmentary — REQ-HUM-007a).

### Guard 2 — Japanese one strategic 体言止め
- **Given** the concrete Japanese copy block:

  > 毎月の請求書作成、まだ手作業ですか。このツールなら、40件が10分で終わります。難しい設定はいりません。あなたの残業を減らす、いちばんの近道。

  Ending annotation (deterministic replay): sentence 1 「…まだ手作業ですか。」 ends on the question particle か; sentence 2 「…10分で終わります。」 ends on ます; sentence 3 「…設定はいりません。」 ends on the negative ません; sentence 4 「あなたの残業を減らす、いちばんの近道。」 is the ONE strategic 体言止め — it closes on the noun 近道 (体言) at the emphasis/closing position, amid the three varied verb/question endings above.
- **When** the skill runs the Japanese copy layer,
- **Then** JA-10 does NOT fire — a single strategic 体言止め amid varied endings is legitimate native craft; JA-10's frequency gate (≥3 consecutive 体言止め, or every headline 体言止め, or 体言止め replacing ending variation throughout) is not tripped here (REQ-HUM-007b).

### Guard 3 — Chinese one crafted 排比
- **Given** a Chinese slogan with one crafted 排比 giving each member a distinct concrete fact (e.g. 万科 "感谢冰峰，感谢风暴，感谢悬崖，感谢缺氧。"),
- **When** the skill runs the Chinese copy layer,
- **Then** the 排比 is NOT flagged (content-first parallelism is prized craft — REQ-HUM-007c).

## §D.4 Quality Gate / Definition of Done

- [ ] All 19 AC checks (AC-HUM-001…017, incl. 006a/b/c) PASS at their declared Method.
- [ ] `make build` runs clean; `catalog.yaml` hash regenerated for the bumped version (1.1.0).
- [ ] `diff -rq` between templates/ and local .claude/ copy → IDENTICAL.
- [ ] Neutrality: 6 forbidden classes → 0 (date body-scoped), reachable on a normal-frontmatter tree; humanize dir absent from the real leak-test violation list (pre-existing unrelated failures ignored).
- [ ] Attribution cleanup: `grep -rn 'NOTICE.md'` on the skill dir → 0 matches; NO NOTICE.md file exists; courtesy credit present ("structure inspired by the im-not-ai (Humanize KR) project", no license claim).
- [ ] `license: Apache-2.0` unchanged (no compound expression); `grep -rn 'MIT License'` on the skill dir → 0 matches.
- [ ] 8-sample matrix each returns a grade + "no meaning drift" (manual scenario).
- [ ] 3 false-positive guards each return NOT-flagged (manual scenario; Guard 2 uses the concrete §D.3 sample).
- [ ] SKILL.md `user-invocable`/`allowed-tools`/`license` unchanged; `metadata.version`/`updated` changed exactly as specified; attribution blocks rewritten to the courtesy credit (REQ-HUM-015).
- [ ] No Go source modified; no new SPEC files flattened into `.moai/specs/`.

## §D.5 Traceability

Every REQ maps to ≥1 AC: REQ-001→AC-001; REQ-002→AC-002/006b/011; REQ-003→AC-003/006a/011; REQ-004→AC-004/006c/011; REQ-005/006→AC-005/013; REQ-007→AC-006a/b/c/014; REQ-008→AC-007; REQ-009→AC-008; REQ-010→AC-009; REQ-011→AC-010; REQ-012→AC-011; REQ-013→AC-012; REQ-014→AC-015; REQ-015→AC-016; REQ-016→AC-017. 8-sample verification (AC-013) exercises the whole detect→rewrite→grade path across REQ-001…006. Method mix (19 checks): 13 mechanical, 2 hybrid (AC-005, AC-006c), 4 manual (AC-009, AC-011, AC-013, AC-014) — the 4 manual checks are declared `manual` in the matrix rather than presented as false mechanical checks.
