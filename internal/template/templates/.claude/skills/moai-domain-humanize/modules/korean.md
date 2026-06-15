# Humanize — Korean (한국어)

This module describes how to detect and remove AI-generated "tells" from Korean (한국어) text and rewrite it into natural, human-sounding prose (윤문). It defines ten detection categories (A–J), three severity tiers (S1/S2/S3), a four-grade quality rubric (A/B/C/D), and over-editing guardrails that protect meaning. All instruction prose is in English; the rewrite examples are in Korean because the AI tells are language-specific and cannot be demonstrated in any other language.

The supreme, non-negotiable rule is **meaning preservation**: a rewrite that changes facts, numbers, named entities, quotations, or causal/logical claims is a failure regardless of how natural it reads.

## Detection Categories

The catalogue spans ten categories. Each row gives the tell, why a reader perceives it as AI authorship, and the dominant severity tier for that category. Subcategory IDs follow the verbatim inventory: A-1…A-19 (A-17 held), B-1…B-4, C-1…C-12, D-1…D-7, E-1…E-7, F-1…F-5, G-1…G-3, H-1…H-4, I-1…I-6, J-1…J-4.

| Category | Tell | Why it reads as AI | Severity |
|----------|------|--------------------|----------|
| **A — 번역투 (Translationese / Calque)** | Literal English-syntax carry-over: `~을 통해`, `~에 대하여`, double passives (`~되어진다`), `가지고 있다` have-verbs, mandatory pronouns (`그/그녀/그것/그들`), deep left-branching relative clauses, double-postposition binding (`~에서의`). | English-trained models map source-language structure 1:1 onto Korean instead of using native subject-drop, verb-centric phrasing. Spacing, morphology, and punctuation artifacts compound the foreign feel. | S1 (A-1, A-2, A-3, A-7, A-8, A-16); rest S2 |
| **B — 영어 인용·용어 과다 (English Citation / Term Excess)** | Reflexive parenthetical glossing (`인공지능(AI)`), untranslated jargon (`framework`, `leverage`), over-long English quotations. | Models default to glossing every term and leaving English untranslated — an encyclopedic, hedged register a human writer would not sustain. | S2 (B-1…B-3); S3 (B-4) |
| **C — 구조적 AI 패턴 (Structural / Layout)** | Mechanical `첫째/둘째/셋째`, bullet-list overuse, repeated section headings, colon-subtitle headings (`### 서론: …`), emoji spam, comma-after-connective (`발전하지만,`), high overall comma ratio. | The visual chat-model signature: rigid scaffolding plus comma and emoji density that human prose lacks (humans ≈26% comma sentences vs AI 50%+). | S1 (C-1, C-5, C-11); rest S2 |
| **D — AI 특유 관용구 (Signature Phrasemes)** | `결론적으로`, `시사하는 바가 크다`, `간과할 수 없다`, `크게 세 가지로 나눌 수 있다`, hype adjectives (`혁신적인`, `압도적`, `폭발적`), the `X에서 Y로` conversion formula. | These exact phrases recur across model outputs as filler — they signal summarization-by-template rather than authored thought. | S1 (D-1…D-4); S2 (D-5…D-7) |
| **E — 리듬·문장 길이 균일성 (Rhythmic Uniformity)** | Low sentence-length variance (all 30–50 chars), repeated identical endings (`~이다. ~이다.`), uniform 3–4-sentence paragraphs, short-sentence monotony, mixed politeness levels. | Humans vary cadence (short and long sentences interleaved); models regress to a uniform mean length and ending. Expert panels note deeper but less-varied structure as a tell. | S2 (E-1…E-7) |
| **F — 과도한 수식·중복 (Modifier / Redundancy Excess)** | `매우/정말/극히` adverb addiction, synonym double-modifiers (`중요하고 핵심적인`), function+role compounds (`역할과 기능`), `-적 N` abstract chains (`기술적 토대`), `-성/-화` suffix spam. | Models pad with redundant intensifiers and Sino-Korean abstract suffixes to sound authoritative, producing bloated noun phrases. | S2 (F-1…F-5) |
| **G — 과도한 Hedging (Hedging Abuse)** | `~할 수 있을 것으로 보인다`, `~인 것으로 판단된다`, triple softening (`~할 가능성이 있을 수 있다`), safety-balance lexicon (`양쪽 모두`, `신중하게`, `균형`). | RLHF-trained models over-hedge to avoid commitment; stacked epistemic softeners are a hallmark of model caution, not human assertion. | S2 (G-1…G-3) |
| **H — 접속사 남발 (Conjunction Glut)** | Sentence-initial `또한/따라서/즉/나아가` on most sentences, `하지만`+`그러나` over-mixing, repeated `이는 ~` deixis, redefinition `즉` overuse. | Models over-signpost logical flow with head-of-sentence connectives a human would leave implicit. | S2 (H-1…H-4) |
| **I — 형식명사·의존명사 과다 (Formal-Noun Nominalization)** | `것이다` ending overuse, `점/바/수/데` repetition, `~라는 것`, `~할 필요가 있다` recommendation endings, `~능력` abstract-noun chains. | Nominalization (명사화) replaces direct verbs with bound-noun constructions — the bureaucratic register models favor over plain Korean predication. | S2 (I-1…I-6) |
| **J — 시각 장식 남용 (Visual Ornament Abuse)** | Excessive **bold**, scare-quote overuse, em-dash (—) abuse, parenthetical-aside spam. | Markdown-heavy decoration (bold on every keyword, dashes everywhere) is the formatting fingerprint of chat-model output. | S2 (J-1…J-4) at density; S3 when isolated |

## Severity Rules

Three tiers govern how aggressively a tell must be removed.

- **S1 — Decisive (remove on single occurrence).** A single instance is enough to confirm AI authorship, so it must always be removed. S1 members are: A-1, A-2, A-3, A-7, A-8, A-16; C-1, C-5, C-11; D-1, D-2, D-3, D-4.
- **S2 — Strong (allow 1–2, remove at 3+).** One or two instances are tolerable as natural variation; the third occurrence triggers removal. S2 covers the remaining A subcategories (A-4, A-5, A-6, A-9–A-15, A-18, A-19); B-1…B-3; C-2, C-3, C-4, C-6, C-7, C-8, C-9, C-10, C-12; D-5, D-6, D-7; all of E, F, G, H, I; and J-1…J-4.
- **S3 — Weak (only problematic when layered).** Harmless in isolation; remove only when stacked with other tells. S3 covers B-4 (`~라고 알려진 / ~로 일컬어지는`); J-category items degrade to S3-equivalent strength when they appear alone rather than at density.

**Held subcategory — A-17** (inanimate/abstract-noun `-들` plural marker): strong academic basis but zero positive corpus instances. Do not gate on it; it is reserved for a later revision pending re-test.

## Rewrite Examples

Each example shows the original AI-flavored Korean, the rewritten human-sounding Korean, and the subcategory addressed. Numbers, named entities, and quotations are preserved exactly. Every one of the ten categories is demonstrated.

### Category A — 번역투

1. `데이터 분석을 통해 인사이트를 얻는다` → `데이터를 분석해 인사이트를 얻는다` (A-2: drop the `~을 통해` calque, use a native verb)
2. `강한 경쟁력을 가지고 있다` → `경쟁력이 강하다` (A-7: replace the `가지고 있다` have-construction with predication)
3. `John은 피곤했다. 그는 앉았다. 그는 한숨을 쉬었다` → `존은 피곤했다. 자리에 앉아 한숨을 쉬었다` (A-16: zero-anaphora — drop redundant pronouns)

### Category B — 영어 인용·용어 과다

1. `이 프레임워크(framework)를 leverage하여 생산성을 높인다` → `이 도구를 활용해 생산성을 높인다` (B-1 + B-2: remove the redundant parenthetical gloss and the untranslated `leverage`)
2. `RAG(검색 증강 생성)는 LLM(거대 언어 모델)의 hallucination(환각)을 줄인다` → `검색 증강 생성은 거대 언어 모델의 환각을 줄인다` (B-1: keep the Korean term, drop reflexive English/괄호 glossing — first mention only if truly needed)

### Category C — 구조적 패턴

1. `발전하지만, 대응은 더디다` → `발전하지만 대응은 더디다` (C-11: delete the comma after the connective ending — over-segmentation signal)
2. `### 서론: 제조업의 미래, AI에 달려있다` → `### 제조업의 미래` (C-10: strip the colon-subtitle heading)
3. `✅ 효율 개선 🚀 비용 절감 💡 핵심 인사이트` → `효율을 개선하고 비용을 줄인다. 핵심은 다음과 같다` (C-5: remove emoji, restore prose)

### Category D — 관용구

1. `결론적으로, 이는 매우 중요하다고 할 수 있다` → `이 변화는 비용 구조를 바꾼다` (D-1 + D-2: replace summary filler with a concrete claim)
2. `DeepSeek-V4의 등장은 효율의 가치를 보여줍니다` → `DeepSeek는 적은 자원으로도 성능이 나온다는 것을 증명했다` (D-5 / A-15: make the agent explicit; product name preserved)

### Category E — 리듬·문장 길이 균일성

1. `시장이 성장한다. 기업이 대응한다. 소비자가 반응한다.` → `시장은 빠르게 큰다. 기업들은 뒤늦게, 그러나 한꺼번에 움직이기 시작했다. 소비자는 이미 떠난 뒤였다.` (E-1 + E-2: break the uniform 길이·`~다` 종결 rhythm with varied sentence length and endings)

### Category F — 수식·중복

1. `매우 중요하고 핵심적인 요소` → `핵심 요소` (F-1 + F-2: cut the intensifier and the synonym double-modifier)
2. `근본적 관점에서 구조적 변화가 필연적이다` → `구조가 근본부터 바뀐다` (F-4: unwind the `-적 N` abstract chain)

### Category G — Hedging

1. `효율을 높일 수 있을 것으로 보인다` → `효율이 높아진다` (G-1: assert rather than hedge)
2. `개선될 가능성이 있을 수 있다` → `개선될 수 있다` (G-2: collapse stacked softeners to one)

### Category H — 접속사 남발

1. `또한 비용이 절감된다. 따라서 효율이 높아진다. 즉, 경쟁력이 생긴다.` → `비용이 줄면 효율이 오르고, 결국 경쟁력으로 이어진다.` (H-1: drop the head-of-sentence `또한/따라서/즉` chain and let one sentence carry the flow)

### Category I — 형식명사 (명사화)

1. `주목할 점은 비용이 크게 줄었다는 것이다` → `비용이 크게 줄었다` (I-1 + I-2: drop the `점 … 것이다` nominalization frame)
2. `조직 차원의 혁신이 필요하다` → `조직이 혁신을 주도한다` (I-5: turn the abstract noun into verbal predication)

### Category J — 시각 장식 남용

1. `**핵심**은 **속도**이며, 이는 **비용**과 직결된다 — 즉, **효율**의 문제다` → `핵심은 속도다. 속도가 곧 비용이고, 결국 효율의 문제다.` (J-1 + J-3: strip the keyword-by-keyword bold and the em-dash aside)

## Quality Grading

Grade the rewrite against the residual S1/S2 counts and the improvement ratio. "Improvement %" is the reduction in weighted tell count between the original and the rewrite.

| Grade | Criteria | Action |
|-------|----------|--------|
| **A** | Zero S1, ≤2 S2, ≥70% improvement | Pass |
| **B** | Zero S1, ≤4 S2, ≥50% improvement | Pass (Strict-mode deeper validation recommended) |
| **C** | 1–2 S1 remaining, OR 2+ over-editing signals | Re-run a second rewrite round |
| **D** | ≥3 S1 remaining, OR severe over-editing | Hold and report — human review recommended |

An "over-editing signal" is one discrete guardrail trip: a change rate above 30%, a flagged meaning drift, or a register/politeness shift not required by any detected tell. Two or more such signals cap the grade at C.

## Over-Editing Guardrails

Meaning preservation is the top rule and overrides every stylistic improvement. Naturalness must never be bought with a change in meaning.

**Change-rate metric.** Compute Levenshtein distance ÷ source length. The target band is **5–30%**.

| Threshold | Trigger | Behavior |
|-----------|---------|----------|
| **> 30% changed** | `변경률 30% 초과` | Warning — flag the passage for possible over-editing of tone or voice |
| **> 50% changed** | `변경률 50% 초과` | Forced stop and automatic rollback — content corruption is likely (`변경률 50% 초과 시 자동 롤백`) |

**Meaning-preservation rules (the non-negotiable top rule, `의미 불변이 최상위 불문율`):**

- Numeric data, figures, and statistics stay **character-intact** — never altered.
- Proper nouns and product names (for example GPT-4, Claude 3, DeepSeek-V4), legal text, quotations, and citations are preserved exactly.
- Causal relations and logical claims must remain equivalent (의미 동등성).
- For ambiguous passages, do **not** add or elaborate content — no unwarranted expansion.

**Rewriting playbook anchors.** Apply native-Korean transformations consistently: by-passive → active; double-passive → single; three or more anaphoric pronouns → 50%+ zero-anaphora; pre-noun modifier of three or more phrases → split or post-appositive; have/make/take calques → native verb (`회의를 가지다` → `회의를 했다`); four or more identical sentence endings → varied endings; `-tion/-ment/-ness` literals → verbal form; `Mr./Dr.` → Korean honorific or omit.

## Modes (Fast / Strict)

The two operating modes trade speed for validation depth. Both honor the meaning-preservation rules above.

- **Fast mode** (default, source up to ~5,000 chars): a single pass that detects, rewrites the flagged spans, and self-validates against the meaning-preservation checklist. Auto-escalate to Strict when the source exceeds ~8,000 chars.
- **Strict mode** (`--strict`, or source ≥ ~8,000 chars, or high-stakes text): staged — detect → surgical rewrite → content-fidelity audit (numbers, named entities, claims unchanged) → naturalness review. Allow up to 3 rewrite rounds; if S1 tells remain after the third round, hold and report for human review rather than over-editing.

The "read-aloud" test is the most reliable rhythm check for Korean (categories E and the cadence half of J): if a sentence cannot be read aloud naturally, it still carries a tell.

---

## Source & License

This Korean module is a faithful port of the AI-tell taxonomy from the **im-not-ai (Humanize KR)** open-source skill (https://github.com/epoko77-ai/im-not-ai), used under the **MIT License — Copyright (c) 2026 epoko77-ai**. See NOTICE.md for the full attribution.
