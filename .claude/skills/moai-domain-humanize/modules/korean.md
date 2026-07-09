# Humanize — Korean (한국어)

This module catalogues the "AI tells" of Korean (한국어) model output — the recurring habits that make a Korean text read as machine-written — and the rules for rewriting it into natural, human-sounding prose (윤문). It covers two genre surfaces:

1. **Prose layer** (categories A–J): columns, reports, blog posts, formal documents — the lexical and structural tells of running text.
2. **Copy layer** (categories A-20…A-25, L, M): marketing headlines, CTAs, landing pages, brand storytelling, and slide decks — the structural and rhetorical tells that appear even when the vocabulary is clean.

All instruction prose is in English; the before/after rewrite examples are in Korean because the tells are language-specific and cannot be demonstrated in any other language. The supreme, non-negotiable rule is **meaning preservation**: a rewrite that changes facts, numbers, named entities, quotations, or causal claims is a failure regardless of how natural it reads.

Many prose categories descend from the translationese (번역투) patterns long studied in Korean translation practice — the unnatural constructions that appear when English syntax is mapped one-to-one onto Korean. AI-generated Korean reproduces the same translationese, plus chat-model formatting habits of its own.

## Prose Layer — Detection Categories (A–J)

Each row names the category, the tell, why a reader perceives it as machine authorship, and the dominant severity tier. Subcategory IDs are stable handles: A-1…A-19 (A-17 reserved), B-1…B-4, C-1…C-12, D-1…D-7, E-1…E-8, F-1…F-5, G-1…G-3, H-1…H-4, I-1…I-7, J-1…J-4.

| Category | Tell | Why it reads as AI | Severity |
|----------|------|--------------------|----------|
| **A — 번역투 (Translationese / Calque)** | English-syntax carry-over: `~을 통해`, `~에 대하여`, `~에 있어서`, double passives (`~되어진다`), `가지고 있다` have-verbs, mandatory pronouns (`그/그녀/그것/그들`), deep left-branching relative clauses, stacked particles (`~에서의`, `~으로의`). | English-trained models map source structure one-to-one onto Korean instead of using native subject-drop, verb-centric phrasing. Pronoun density in model output runs several times higher than in non-translated Korean. | S1 (A-1, A-2, A-3, A-7, A-8, A-16); rest S2 |
| **B — 영어 인용·용어 과다 (English Citation / Term Excess)** | Reflexive parenthetical glossing (`인공지능(AI)`), untranslated jargon (`framework`, `leverage`), over-long English quotations, `~라고 알려진` known-as calques. | Models gloss every term and leave English untranslated — an encyclopedic register a human writer would not sustain. | S2 (B-1…B-3); S3 (B-4) |
| **C — 구조적 AI 패턴 (Structural / Layout)** | Mechanical `첫째/둘째/셋째`, bullet-list overuse, schematic headings, colon-subtitle headings (`### 서론: …`), numbered `1) 2) 3)` indexing, emoji spam, a comma after every connective ending (`발전하지만,`), high overall comma ratio. | The visual chat-model signature: rigid scaffolding plus comma and emoji density human Korean prose lacks. The comma-after-connective habit is the single strongest separator between human and model Korean. | S1 (C-1, C-5, C-11); rest S2 |
| **D — AI 특유 관용구 (Signature Phrasemes)** | `결론적으로`, `시사하는 바가 크다`, `간과할 수 없다`, `크게 세 가지로 나눌 수 있다`, hype adjectives (`혁신적인`, `압도적`, `폭발적`), personified abstract subjects, the `X에서 Y로` conversion formula, `~할 때입니다` closing formulas. | These exact phrases recur across model outputs as filler — they signal summarization-by-template rather than authored thought. | S1 (D-1…D-4); S2 (D-5…D-7) |
| **E — 리듬·문장 길이 균일성 (Rhythmic Uniformity)** | Low sentence-length variance, repeated identical endings (`~이다. ~이다.`), uniform 3–4-sentence paragraphs, simple-sentence monotony with no complex clauses, long comma-chained clauses, politeness-level drift mid-document, mechanically uniform compound-word spacing. | Humans vary cadence — short and long sentences interleave, endings shift, registers hold steady. Models regress to a flat mean and lose the politeness-level contract with the reader. | S2 (E-1…E-8) |
| **F — 과도한 수식·중복 (Modifier / Redundancy Excess)** | `매우/정말/극히` intensifier addiction, synonym double-modifiers (`중요하고 핵심적인`), function+role compounds (`역할과 기능`), Sino-Korean suffix spam (`-성/-적/-화`), `-적 N` abstract chains (`기술적 토대`). | Models pad with redundant intensifiers and abstract suffixes to sound authoritative, producing bloated noun phrases where plain predication would do. | S2 (F-1…F-5) |
| **G — 과도한 Hedging (Hedging Abuse)** | `~할 수 있을 것으로 보인다`, `~인 것으로 판단된다`, stacked softeners (`~할 가능성이 있을 수 있다`), safety-balance lexicon (`양쪽 모두`, `신중하게`, `균형`). | Alignment-trained models over-hedge to avoid commitment; stacked epistemic softeners and both-sides vocabulary are hallmarks of model caution, not human assertion. | S2 (G-1…G-3) |
| **H — 접속사 남발 (Conjunction Glut)** | Sentence-initial `또한/따라서/즉/나아가` on most sentences, `하지만`+`그러나` over-mixing, repeated `이는 ~` deixis, redefinition `즉` overuse. | Models over-signpost logical flow with head-of-sentence connectives a human writer leaves implicit. | S2 (H-1…H-4) |
| **I — 형식명사·의존명사 과다 (Formal-Noun Nominalization)** | `것이다` ending overuse, `점/바/수/데` repetition, `~라는 것`, `~할 필요가 있다` recommendation endings, `~능력` abstract-noun chains, corporate boilerplate first-person (`당사는/저희는 ~합니다`). | Nominalization replaces direct verbs with bound-noun constructions — the bureaucratic register models favor over plain Korean predication. | S2 (I-1…I-7) |
| **J — 시각 장식 남용 (Visual Ornament Abuse)** | Excessive **bold**, scare-quote overuse, em-dash (—) asides, parenthetical-aside spam. | Markdown-heavy decoration — bold on every keyword, dashes everywhere — is the formatting fingerprint of chat-model output. | S2 (J-1…J-4) at density; S3 when isolated |

## Prose Severity Rules

- **S1 — Decisive (remove on single occurrence).** One instance is enough to confirm machine authorship. Members: A-1, A-2, A-3, A-7, A-8, A-16; C-1, C-5, C-11; D-1, D-2, D-3, D-4.
- **S2 — Strong (allow 1–2, remove at 3+).** One or two instances pass as natural variation; the third triggers removal. Members: the remaining A subcategories (A-4, A-5, A-6, A-9…A-15, A-18, A-19); B-1…B-3; C-2, C-3, C-4, C-6…C-10, C-12; D-5, D-6, D-7; all of E, F, G, H, I; J-1…J-4 at density.
- **S3 — Weak (only problematic when layered).** Harmless in isolation; contributes only when stacked with other tells. Members: B-4; J-category items when they appear alone rather than at density.

Notes on the inventory:
- **A-17 is a reserved slot** (mechanical `-들` plural marker on inanimate/abstract nouns): the linguistic basis is strong, but positive corpus instances are lacking, so no detector gates on it. The ID stays reserved so downstream tooling keeps a stable numbering.
- **E-7 (politeness-level consistency loss)** applies to dialogue and quoted speech: mixing 해라/해요/합쇼체 levels within one document or conversation. Formal single-register documents are exempt.
- **E-8 (uniform compound-word spacing)** is an auxiliary signal only — Korean spacing norms vary by compound, so it never counts alone.
- **I-7 (corporate boilerplate first-person)** gates on the combined density of `당사는/저희는` subjects plus formal `~합니다` closures in business copy, CS replies, and press-release registers.

## Prose Rewrite Examples

Each pair shows the machine-flavored original, the human-sounding rewrite, and the subcategory addressed. Numbers, named entities, and quotations are preserved exactly.

### Category A — 번역투

1. `데이터 분석을 통해 인사이트를 얻는다` → `데이터를 분석해 인사이트를 얻는다` (A-2: drop the `~을 통해` calque; use a native verb)
2. `강한 경쟁력을 가지고 있다` → `경쟁력이 강하다` (A-7: replace the have-construction with direct predication)
3. `지수는 지친 표정이었다. 그는 앉았다. 그는 한숨을 쉬었다.` → `지수는 지친 표정으로 앉아 한숨을 쉬었다.` (A-16: zero-anaphora — Korean drops recoverable pronouns)
4. `본사 2층에서의 워크숍` → `본사 2층에서 연 워크숍` (A-19: unstack the double particle)

### Category B — 영어 인용·용어 과다

1. `이 프레임워크(framework)를 leverage하여 생산성을 높인다` → `이 도구를 활용해 생산성을 높인다` (B-1 + B-2: drop the redundant gloss and the untranslated verb)
2. `RAG(검색 증강 생성)는 LLM(거대 언어 모델)의 hallucination(환각)을 줄인다` → `검색 증강 생성은 거대 언어 모델의 환각을 줄인다` (B-1: gloss once at most, then write Korean)

### Category C — 구조적 패턴

1. `발전하지만, 대응은 더디다` → `발전하지만 대응은 더디다` (C-11: delete the comma after the connective ending — the strongest single separator)
2. `### 서론: 제조업의 미래, AI에 달려있다` → `### 제조업의 미래` (C-10: strip the colon-subtitle heading)
3. `✅ 효율 개선 🚀 비용 절감 💡 핵심 인사이트` → `효율을 개선하고 비용을 줄인다. 핵심은 다음과 같다.` (C-5: remove emoji, restore prose)

### Category D — 관용구

1. `결론적으로, 이는 매우 중요하다고 할 수 있다` → `이 변화는 비용 구조를 바꾼다` (D-1 + D-2: replace summary filler with a concrete claim)
2. `두 진영의 충돌이 질문을 던집니다` → `두 회사가 부딪히면서 질문이 하나 남았습니다` (D-5: return the personified abstract subject to a real agent)
3. `'지식 전달자'에서 '학습 조력자'로` → `교사는 더 이상 지식만 전달하지 않는다. 학생 곁에서 학습을 돕는다.` (D-7: unroll the conversion formula into direct assertion)

### Category E — 리듬·문장 길이 균일성

1. `시장이 성장한다. 기업이 대응한다. 소비자가 반응한다.` → `시장은 빠르게 큰다. 기업들은 뒤늦게, 그러나 한꺼번에 움직이기 시작했다. 소비자는 이미 떠난 뒤였다.` (E-1 + E-2 + E-4: break the uniform length, ending, and simple-sentence monotony)

### Category F — 수식·중복

1. `매우 중요하고 핵심적인 요소` → `핵심 요소` (F-1 + F-2: cut the intensifier and the synonym double-modifier)
2. `근본적 관점에서 구조적 변화가 필연적이다` → `구조가 근본부터 바뀐다` (F-4: unwind the `-적 N` abstract chain)

### Category G — Hedging

1. `효율을 높일 수 있을 것으로 보인다` → `효율이 높아진다` (G-1: assert rather than hedge)
2. `개선될 가능성이 있을 수 있다` → `개선될 수 있다` (G-2: collapse stacked softeners to one)

### Category H — 접속사 남발

1. `또한 비용이 절감된다. 따라서 효율이 높아진다. 즉, 경쟁력이 생긴다.` → `비용이 줄면 효율이 오르고, 결국 경쟁력으로 이어진다.` (H-1: drop the head-of-sentence connective chain; let one sentence carry the flow)

### Category I — 형식명사 (명사화)

1. `주목할 점은 비용이 크게 줄었다는 것이다` → `비용이 크게 줄었다` (I-1 + I-2: drop the `점 … 것이다` nominalization frame)
2. `당사는 고객 만족을 최우선으로 생각합니다. 저희 제품은 엄격한 품질 관리를 거쳐 제공됩니다.` → `우리는 고객 만족을 먼저 봅니다. 제품은 출고 전 전수 검사를 거칩니다.` (I-7: break the corporate boilerplate first-person + formal-closure loop)

### Category J — 시각 장식 남용

1. `**핵심**은 **속도**이며, 이는 **비용**과 직결된다 — 즉, **효율**의 문제다` → `핵심은 속도다. 속도가 곧 비용이고, 결국 효율의 문제다.` (J-1 + J-3: strip keyword-by-keyword bold and the em-dash aside)

---

## Copy Layer (Korean)

Genre scope: marketing headlines, CTAs, landing-page copy, brand/founder storytelling, customer-testimonial narratives, and slide/deck titles. The copy genre carries AI tells that are **structural and rhetorical, not lexical** — a headline can be free of every prose-layer word tell and still read machine-written. Copy mode also changes the over-editing guard: the change-rate limit is replaced by the **fact-anchor preservation guard** (see the shared SKILL.md guardrails) — numbers, dates, prices, proper nouns, and legal notation stay intact 100%, and the core promise/benefit keeps its meaning, while expression and sentence structure may be freely rewritten.

### Detection Categories (copy translationese — A-20…A-25)

These extend the prose translationese category A into the copy genre: calques that surface specifically in headlines, CTAs, and landing copy.

| ID | Tell | Why it reads as AI | Severity |
|----|------|--------------------|----------|
| A-20 | Machine/system verb calque `굴러가다/굴리다` with an automation/system subject ("자동화가 굴러갑니다", "자동화를 굴리는 네 가지") — the English run/roll/operate sense mapped onto Korean | "일이 굴러간다" is natural colloquial Korean for general work, but a machine or automation subject with `굴러가다` copies the English machine-operation verb; models repeat it across automation-product copy | S1 (decisive with an automation/system subject) |
| A-21 | Abstract-noun sentence closure — the sentence ends on an abstract noun phrase ("흩어진 일이 **하나의 흐름이 됩니다**", "업무가 **하나의 경험이 됩니다**") copying English become-X/into-X noun endings | Korean closes naturally on a predicate verb; models carry over the English noun-final frame, and it is the single most frequent copy-genre calque | S1 |
| A-22 | Agency/collaboration verb calque "나 대신 일합니다 / 나와 함께 일합니다" (works-for-me / works-with-me) | Korean expresses delegated benefit with the `-어 주다` benefactive ("대신 일해 줍니다") and drops the first-person pronoun; the literal frame keeps English argument structure | S2 (allow 1–2; overlaps personification D-5 — apply the D-5 fix first) |
| A-23 | Metaphor calque — English figurative stock ("X는 Y의 엔진/심장/날개", "같은 페이지를 본다", "손발을 맞추다" for sync) transplanted literally | Korean either has its own idiom or prefers a concrete verb; a transplanted metaphor reads as imported rhetoric | S2 (one metaphor in a slogan headline is allowed) |
| A-24 | Sentence-initial adverb calque `더는` ("더는 혼자가 아닙니다") for English no-longer | Korean opens this appeal naturally with `이제는/이젠`; sentence-initial `더는` is the no-longer frame carried over | S2 (emotional/appeal copy only) |
| A-25 | Third-person expository voice where Korean copy wants second-person appeal — "자동화가/시스템이 ~합니다" throughout, with zero reader address (`여러분/당신`) | English product copy defaults to third-person description; effective Korean appeal copy converts to second-person address, and the missing conversion leaves a distant, catalog-like tone | S2 (appeal headlines/CTA only — informational copy such as FAQ, specs, prices legitimately stays third-person) |

### Detection Categories (storytelling / brand narrative — L-1…L-8)

Brand stories, founder narratives, and testimonial-style copy carry their own machine habits: stock arcs, unverifiable feeling, and template myths.

| ID | Tell | Why it reads as AI | Severity |
|----|------|--------------------|----------|
| L-1 | Cliché opener + mechanical narrative arc — "어느 날", "그렇게 ~는 시작되었습니다" openers riding a predictable conflict→resolution→lesson arc | Models converge on the same stock arc across topics; the opener alone flags the template | phrase S1 / structure S2 |
| L-2 | Stock emotion / sentimentality — "눈물이 흘러내렸다", "정말 감동적인 순간이었다" and similar unverifiable feeling statements | The tell is not naming an emotion but reaching for the same unverifiable stock expressions instead of showing a concrete scene, action, or quotation | S2 |
| L-3 | Moral-lesson compulsion — "이 경험을 통해 ~을 배웠습니다" closing every episode with an explicit lesson | Human narratives often end on action or consequence; the compulsory stated lesson is a template artifact | phrase S1 / structure S2 |
| L-4 | Cliché transition — "하지만 그때", "그 순간", "운명처럼" bending the story at stock pivot phrases | The same pivot phrases recur across model narratives regardless of content | S1 |
| L-5 | Too-smooth ending — every conflict resolved, "우리는 성공했다" | Real ventures keep open problems; the fully-resolved ending signals generated narrative | S2 |
| L-6 | Founder-myth template — "작은 원룸에서 시작해", garage-origin humble-beginnings arc | The humble-origin myth without verifiable lived detail is a stock skeleton | S2 |
| L-7 | Fake specificity / unverifiable statistics — "수많은 논문 분석", "후기 2042건 분석", credential cosplay | Precision theater: numbers and expertise that cannot be checked simulate authority | S2 |
| L-8 | Ad-like customer-testimonial story — uniform format ("13년차 직장맘도 인정한"), professional cosplay, exaggerated effect | Testimonials in one repeated mold with unverifiable personas read as agency-written | S2–S3 |

### Detection Categories (slide / presentation structure — M-1…M-3)

Slide titles and deck headlines have structural tells detectable at the sentence-shape level; one occurrence is decisive.

| ID | Tell | Why it reads as AI | Severity |
|----|------|--------------------|----------|
| M-1 | Dash-contrast headline — "X — Y" split by a dash ("복붙에서 위임으로 — 목표만 주면") | The two-part dash headline is a generated-deck signature; Korean headline craft integrates or splits the halves | S1 |
| M-2 | Particle/noun-ending fragment headline — the title ends on a particle or bare noun with no predicate ("성공의 열쇠") | A standalone fragment that needs explanation to complete it reads as template output; Korean titles carry a predicate or form a complete informational noun phrase | S1 |
| M-3 | "A에서 B로" transition-formula opener ("엑셀에서 노션으로") | The from-X-to-Y slot formula opens generated decks with interchangeable nouns and no concrete claim | S1 |

**Noun-phrase boundary for M-2**: complete metadata noun-phrase titles are ALLOWED — "2026년 Q1 사업 보고" is a full informational title, not a fragment. The tell is the standalone predicate-less fragment that cannot stand without explanation ("성공의 열쇠", "변화의 시작").

### Copy Severity Rationale

- **S1**: A-20 (with automation/system subject), A-21, L-1/L-3/L-4 phrase forms, and all of M-1…M-3 — one occurrence in a headline or deck confirms machine authorship, because a human copywriter essentially never produces these shapes by accident at a title position.
- **S2**: A-22…A-25, L-1/L-3 structure forms, L-2, L-5…L-8 — these appear occasionally in genuine human copy; density and co-occurrence convict.
- **Cross-tell stacking**: a generated Korean landing section typically stacks A-21 (noun closure) + M-1 (dash headline) + A-25 (third-person voice). Resolve the S1 tells first, then re-score the S2 stack.

### Copy Rewrite Examples

1. `자동화는 24시간 굴러갑니다` → `자동화는 24시간 알아서 돌아갑니다` (A-20: replace the machine-verb calque with a native operation verb)
2. `흩어진 일이 하나의 흐름이 됩니다` → `흩어진 일이 하나로 이어집니다` (A-21: close on a verb, not an abstract noun)
3. `자동화가 나 대신 일합니다` → `자동화가 대신 일해 줍니다` (A-22: benefactive `-어 주다`, drop the pronoun)
4. `이 도구는 팀의 엔진입니다` → `이 도구가 팀을 움직입니다` (A-23: return the metaphor to a direct verb)
5. `더는 혼자가 아닙니다` → `이제는 혼자가 아닙니다` (A-24: native adverb for the no-longer frame)
6. `자동화가 반복 업무를 처리합니다` (헤드라인) → `여러분의 반복 업무를 AI에게 맡기세요` (A-25: convert third-person description to second-person appeal — appeal copy only)
7. `어느 날, 작은 사무실에서 우리의 이야기는 그렇게 시작되었습니다` → `첫 주문이 들어온 날, 사무실에는 책상이 하나뿐이었다` (L-1: replace the stock opener with a concrete scene)
8. `가슴이 뭉클해졌고, 정말 감동적인 순간이었습니다` → `편지에는 '다시 사겠다'는 문장이 세 번 반복돼 있었다` (L-2: show the verifiable detail instead of naming the feeling)
9. `이 경험을 통해 저희는 진심의 소중함을 배웠습니다` → `그 실패 이후 반품 절차를 바꿨다. 지금은 환불의 90%가 하루 안에 끝난다` (L-3: end on action and consequence, not a stated lesson)
10. `복붙에서 위임으로 — 목표만 주면` → `목표만 주면, 나머지는 자동화가 처리합니다` (M-1 + M-3: dissolve the dash contrast and the transition formula into one complete claim)
11. `성공의 열쇠` → `자동화가 성공의 열쇠입니다` (M-2: give the fragment a predicate)

### High-False-Positive Signals (copy — do NOT flag)

- `일이 굴러간다` with a general-work subject — natural colloquial Korean; A-20 requires a machine/automation/system subject.
- Complete metadata noun-phrase titles ("2026년 Q1 사업 보고") — full informational titles, not M-2 fragments.
- Informational copy (FAQ, specs, prices, dates, business-registration blocks) written in third person — A-25 applies to appeal copy only.
- One deliberate metaphor in a slogan headline — A-23 allows a single figurative anchor.
- Real, verifiable numbers (an actual survey n, a dated result) — L-7 targets unverifiable precision, not precision itself.

---

## Quality Grading

Grade **after** the rewrite. Prose mode and copy mode grade differently (shared definitions live in SKILL.md; this is the Korean-module summary).

**Prose mode** — residual S1/S2 counts plus improvement % (proportion of detected tells removed without introducing new ones):

| Grade | Criteria | Action |
|-------|----------|--------|
| **A** | 0 residual S1, ≤2 residual S2, ≥70% improvement | Pass |
| **B** | 0 residual S1, ≤4 residual S2, ≥50% improvement | Pass (Strict-mode validation recommended) |
| **C** | 1–2 residual S1, OR 2+ over-editing signals | Second rewrite round |
| **D** | ≥3 residual S1, OR severe over-editing | Hold — human review |

**Copy mode** — residual S1 (including the copy-layer S1 tells A-20/A-21/M-1…M-3 and the L phrase-S1 forms) plus fact-anchor integrity; no change-rate band:

| Grade | Criteria | Action |
|-------|----------|--------|
| **A** | 0 residual S1, 0 fact-anchor loss, self-verification passed | Pass |
| **B** | 0 residual S1, ≤1 conservative fact-anchor concern | Pass with note |
| **C** | 1 residual S1, OR self-verification partially failed | Second rewrite round |
| **D** | 2+ residual S1, OR 2+ fact-anchor losses | Hold — human review |

An "over-editing signal" (prose mode) is one discrete guardrail trip: an estimated change rate above 30%, a flagged meaning drift, or a register shift not required by any detected tell. Two or more such signals cap the grade at C.

## Over-Editing Guardrails

Meaning preservation overrides every stylistic improvement. Naturalness must never be bought with a change in meaning.

**Prose mode — change-rate guard.** Target band roughly 5–30% of the text changed. Above ~30%: WARN and justify every edit against a detected tell. Above ~50%: HALT and roll back — content corruption is likely. Because the change rate is estimated rather than mechanically computed, judge borderline cases conservatively: treat "around the threshold" as over it.

**Copy mode — fact-anchor guard (replaces the change-rate guard).** Numbers, dates, prices, proper nouns, and legal notation stay character-intact; the core promise/benefit keeps its meaning. Expression and sentence structure may be rewritten freely — a headline rewrite legitimately changes most of its characters while preserving every anchor.

**Meaning-preservation rules (both modes):**
- Numeric data, figures, and statistics stay character-intact.
- Proper nouns, product names, legal text, quotations, and citations are preserved exactly.
- Causal relations and logical claims remain equivalent.
- Never add content to resolve ambiguity — no unwarranted expansion, no invented specifics.

**Rewriting anchors (prose).** Apply native-Korean transformations consistently: by-passive → active; double passive → single; three or more anaphoric pronouns → majority zero-anaphora; heavy pre-noun modifier stacks → split or post-appositive; have/make/take calques → native verb (`회의를 가지다` → `회의를 했다`); four or more identical endings → varied endings; English nominalization literals → verbal form.

## Modes (Fast / Strict)

- **Fast mode** (default, source up to ~5,000 chars): a single pass — detect, rewrite the flagged spans, self-verify against the meaning-preservation rules. Auto-escalate to Strict when the source exceeds ~8,000 chars.
- **Strict mode** (long or high-stakes text, or on request): staged — detect → surgical rewrite → content-fidelity audit (numbers, named entities, claims unchanged) → naturalness review. Allow up to 3 rewrite rounds; if S1 tells remain after the third, hold for human review rather than over-editing.

The "read-aloud" test is the most reliable rhythm check for Korean (categories E and the cadence half of J): if a sentence cannot be read aloud naturally, it still carries a tell.

---

## Attribution

The category-catalogue structure of this module is inspired by the **im-not-ai (Humanize KR)** project (https://github.com/epoko77-ai/im-not-ai). The Korean catalogue itself — categories, severities, and all examples — is an original work.
