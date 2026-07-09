# Research — SPEC-HUMANIZE-001 Copy-Genre AI Tells (4 languages)

This file consolidates the evidentiary basis for the copy layer. It carries forward three ephemeral scratchpad research files (English / Japanese / Chinese copy-slop research) plus the Korean port source, into the repo. **The sourced/unsourced separation is load-bearing: any pattern in a module's main catalogue must trace to a Verified Source below; unsourced hypotheses are quarantined in §5 and MUST NOT be promoted into a module.**

Scope of the copy genre: marketing landing pages, headlines, CTAs, taglines, brand/founder storytelling, slide headlines — the structural/rhetorical AI-tell surface the prose catalogues do not cover.

---

## §1 Korean Copy Layer — Faithful Port (source: `general-humanize-korean` taxonomy)

The Korean copy layer is a **faithful port** of the source taxonomy, not independent research. Instruction prose is re-authored in English (REQ-HUM-010); before/after examples stay Korean; internal version annotations and `_source_anchor:` internal-spec citations are stripped (§25). Source-anchor provenance is summarized here (kept OUT of the shipped module per §25).

### §1.1 A-20…A-25 — Copy-genre translationese (extends prose category A)

| ID | Tell | Severity | Source provenance (not shipped) |
|----|------|----------|---------------------------------|
| A-20 | Machine/system verb calque 굴러가다/굴리다 (← EN `run/roll/operate`); "자동화가 굴러간다/굴리다" | S1 (decisive in automation/system context) | Korean copy-genre field observation; 4/14 sample recurrence + industry-expansion samples |
| A-21 | Abstract-noun terminal → verb terminal ("하나의 흐름으로/이 됩니다" ← EN `become X / into X / as one X`) | S1 (highest copy-genre frequency; 6/14 recurrence) | Korean copy-genre field observation |
| A-22 | Agency/collaboration verb calque "나 대신/나와 함께 일합니다" (← EN `works for/with me`) | S2 (allow 1-2; overlaps personification D-5) | Korean copy-genre field observation; 2/14 recurrence |
| A-23 | Metaphor calque (X is the engine/heart/wings of Y; get in sync; on the same page) | S2 (2+ per doc); 1 allowed in a slogan headline | Korean copy-genre field observation ("비유" user requirement) |
| A-24 | Adverb calque sentence-initial "더는" (← EN `no longer`) | S2 (emotional/appeal copy only) | Korean copy-genre field observation |
| A-25 | 3rd-person expository calque — missing 2nd-person appeal (keeps "자동화가/시스템이 ~합니다" instead of "여러분이/여러분의…") | S2 (appeal headlines/CTA only; NOT informational FAQ/spec copy) | Korean copy-genre field observation |

### §1.2 L-1…L-8 — Storytelling / brand-narrative slop (new category L)

| ID | Tell | Severity |
|----|------|----------|
| L-1 | Cliché opener + mechanical narrative arc ("어느 날", "그렇게 ~는 시작되었습니다") | phrase S1 / structure S2 |
| L-2 | Cliché emotion / sentimentality ("눈물이 흘러내렸다", "정말 감동적인 순간") — showing-not-telling fix | S2 |
| L-3 | Moral-lesson compulsion ("이 경험을 통해 ~을 배웠습니다") | phrase S1 / structure S2 |
| L-4 | Cliché transition ("하지만 그때", "그 순간", "운명처럼") | S1 |
| L-5 | Too-smooth / predictable ending (all conflict resolved, "우리는 성공했다") | S2 |
| L-6 | Founder-myth template ("작은 원룸에서 시작해", garage origin) | S2 |
| L-7 | Fake specificity / unverifiable stats ("114편의 논문", "후기 2042건 분석") | S2 |
| L-8 | Ad-like customer-testimonial story (uniform format, professional cosplay) | S2~S3 |

### §1.3 M-1…M-3 — Slide / presentation structural slop (new category M)

| ID | Tell | Severity |
|----|------|----------|
| M-1 | Dash-contrast headline ("X — Y", e.g. "복붙에서 위임으로 — 목표만 주면") | S1 |
| M-2 | Particle/noun-ending fragment headline (조사·체언 종결, no predicate — "성공의 열쇠") | S1 |
| M-3 | "A에서 B로" transition-formula opener ("엑셀에서 노션으로") | S1 |

M-category noun-phrase boundary: complete metadata noun-phrase titles ("2026년 Q1 사업 보고") are ALLOWED; standalone predicate-less fragments needing explanation ("성공의 열쇠") are the M-2 tell.

### §1.4 Korean copy-mode guard + dual grading (source SKILL.md)

The source SKILL.md 4 철칙 carry copy-mode variants:
- **철칙 1 (meaning invariance)**: in copy mode = fact-anchor (numbers/dates/prices/proper nouns/legal notation 100%) + core promise/benefit meaning preserved; expression & sentence structure MAY be rewritten.
- **철칙 4 (over-editing)**: prose mode uses the 30%/50% change-rate guard; copy mode REPLACES it with the fact-anchor preservation guard.
- **Dual grading** (source Phase 5): prose-mode grade (residual S1/S2 + change-rate 10-25% band) vs copy-mode grade (residual S1 incl. M-1…M-3 + fact-anchor loss + self-verification).

---

## §2 English Copy Layer — ENC-1…ENC-9 (web-researched, NOT ported)

ID scheme: `ENC-` (English Copy) prefix, deliberately distinct from the prose `EN-A…EN-J` — it (1) mirrors the Korean genre separation, (2) makes "same rhetorical move, different genre surface" explicit (several ENC tells are copy instances of a prose category), (3) lets a detector scope copy rules to copy contexts. Where an ENC tell is the copy instance of a prose category, the parent is named.

| ID | Tell | Severity | Parent / evidence |
|----|------|----------|-------------------|
| ENC-1 | Aspirational verb + abstract-object headline (`Unleash your potential`, `Elevate your workflow`, `Transform the way you X`, `Supercharge X`) | S2 (whole-phrase `unleash your potential` / `join the revolution` ≈ S1 singleton) | parent EN-A; oliviacal blacklist, landing-page search |
| ENC-2 | Contrastive-negation headline (`It's not just X — it's Y`, `Not X. Y.`, `More than a platform — it's a movement`) | S1 (strongest contemporary copy tell) | parent EN-B; gc.ai, blakestockton, firstcall #1 |
| ENC-3 | Tricolon tagline / value-prop trio (`Fast. Simple. Scalable.`, `Bold. Bright. Better.`) | S2 | parent EN-C; onlygoodcontent #2, firstcall #2 |
| ENC-4 | Landscaping opener (`In today's fast-paced digital world`, `In the ever-evolving landscape of X`) | S1 | parent EN-G/EN-A; oliviacal, firstcall #3 |
| ENC-5 | Poster / wellness-mug closer (aphoristic bow-tied platitude endings) | S2 | parent EN-G/EN-J; onlygoodcontent #8, how-to-tell search |
| ENC-6 | Hollow strategy-speak value prop (`Empowering innovation.`, `Delivering business outcomes.` — gerund + abstract noun, no object/metric) | S2 | parent EN-D+EN-A; onlygoodcontent #5, landing-page search |
| ENC-7 | Generic/hyperbolic CTA microcopy (hype: `Join the Revolution`, `Unleash Your Potential`; bland verbless: `Get Started`, `Submit`) | S2 hype labels; S3 bland labels | CTA-microcopy WebSearch aggregate (copy.ai/matchbox/woobox) — corroboration-only; **backstopped by fetched ENC-1** [†] |
| ENC-8 | Audience-straddle opener (`Whether you're a beginner or a pro, …`) | S2 | alyssawiens, how-to-tell search |
| ENC-9 | Confirmational-authority opener (`The truth is…`, `The reality is…`, `Here's the thing:`) | S2 | relative of EN-E; firstcall #4 |

[†] **ENC-7 evidence footnote (D7).** ENC-7's only direct citation is the CTA-microcopy WebSearch aggregate, which §6.1 labels corroboration-only (not an independently fetched page). ENC-7 is therefore **backstopped by fetched ENC-1**: its diagnostic hype half (`Join the Revolution`, `Unleash Your Potential`) are ENC-1 aspirational-verb phrases placed in a button, and ENC-1 is fetched-and-confirmed via oliviacal (§6.1). The bland half (`Get Started`/`Submit`) is only S3 (a false-positive-guarded downgrade), so ENC-7 does not rest on the aggregate alone. ENC-7 is retained as weak-but-not-unsourced (auditor judgment); it is NOT a §5 quarantined hypothesis.

### §2.1 Severity rationale
- **S1**: ENC-2 (contrastive negation) + ENC-4 (landscaping) — structural moves a human copywriter essentially never produces by accident at a headline/hero position; one occurrence decisive. Whole-phrase singletons `unleash your potential` / `join the revolution` behave as S1.
- **S2**: ENC-1, ENC-3, ENC-5, ENC-6, ENC-7 (hype), ENC-8, ENC-9 — appear in genuine human copy occasionally; density + co-occurrence convict.
- **S3**: bland-CTA half of ENC-7 (`Get Started`, `Submit`) + single em-dash — so common in competent human copy they downgrade a grade only when reinforcing an S1/S2 finding.
- **Cross-tell**: a hero section is often ENC-4 opener → ENC-2 negation headline → ENC-3 tricolon subhead → ENC-1 CTA. Resolve the S1 first, then re-score the S2 stack.

### §2.2 Slide/headline structural mapping (partial transfer, honest false-positive flags)
- **Dash-contrast slide headline `X — Y` (Korean M-1)**: transfers — but as a *contrastive* structure (ENC-2), not merely because it contains a dash. Decisive signal = negation + em-dash + buzzword together, NOT the dash alone.
- **`A → B` / "From X to Y" transition headline (Korean M-3)**: transfers. Severity S2 (humans use "From zero to hero" legitimately). Diagnostic only when abstract and content-free.
- **Predicate-less fragment headline (Korean M-2 noun-ending)**: does **NOT** cleanly transfer. English headlines/slide titles are naturally fragmentary ("Q1 Revenue", "Our Approach", "Why It Matters") in fully human decks — a HIGH-false-positive signal, NOT a standalone removable category. Only contributes when the fragment is also an ENC-1 buzzword noun-phrase ("Unlocking Growth"). Sourcing limitation (stated for honesty): dedicated slide-title AI-structure research in English is thin; the transferable slide claims are inherited from the general copy/headline sources, not a slide-specific corpus study.

### §2.3 English high-false-positive signals (modifiers/downgrades, never standalone removals)
- Em-dash (—): the most-hyped, least-reliable signal; flag only above ~30% of sentences AND with other tells.
- `Get Started`/`Submit`/`Sign Up` CTAs: ubiquitous in human UI; only hype variants are diagnostic.
- Rule of three generally: a legitimate ancient device; only mechanical period-split near-equal-length repeated trios convict.
- Fragment/verbless headlines: natural English headline register.
- A single buzzword ("transform", "unlock"): one occurrence is weak; ENC-1 is density-governed.
- "From X to Y" titles: legitimate for genuine before/after narratives.

---

## §3 Japanese Copy Layer — JA-10…JA-14 (web-researched, NOT ported)

Continues the `JA-` numeric scheme (source prose is JA-01…JA-09). Critical prior finding: 体言止め (noun-ending) is a legitimate, prestigious Japanese copywriting device, so the Korean M-2 "predicate-less fragment = slop" tell does NOT transfer 1:1. The Japanese tell is over-use / default-rhythm, not presence.

| ID | Tell | Severity | Evidence (verified) |
|----|------|----------|---------------------|
| JA-10 | 体言止め過剰依存 (noun-ending over-reliance) — every headline/most sentences end in 体言止め/言い切り as the *default* rhythm | **S2 (frequency-based, NOT presence)** | tachibannna, fereple |
| JA-11 | 英語式コロン見出し (English-style sentence-final/heading colon `見出し：`, often + half-width space) | S1 (on sight) | nahouemura |
| JA-12 | ダッシュ対比・記号代用 (em-dash `—` where Japanese uses 「……」/中黒/full sentence) | S1 for the punctuation artifact; S2 in copy | cooqieinc (EN mechanism) + existing JA-08 |
| JA-13 | 定型訴求フレーズ (formulaic value-prop CTA `〜を実現します`/`〜を可能にします` + abstract ownerless benefit, no number/name/outcome) | S2 (allow 1-2; flag at 3+) | ai-souken + existing JA-01 |
| JA-14 | ブランド/創業ストーリーテンプレート (founder-myth arc, creator self-narration, no 一次情報) | S2 (S1 for fabricated beats) | twelfth |

Cross-cutting (NOT new IDs): copy amplifies JA-09 katakana overload; JA-04/JA-08 already own monotone です・ます endings and em-dash/emoji/Markdown artifacts respectively.

### §3.1 Severity rationale
- JA-11 = S1 by parity with existing JA-08 (English-imported punctuation/Markdown artifact); near-deterministic leak.
- JA-12 = S1 for raw punctuation artifact, S2 in body copy (a human MIGHT choose 「──」; require repetition).
- JA-10 = S2 (frequency, not presence) — because 体言止め is legitimate; flag when 体言止め is the *default* (≥3 consecutive lines end 体言止め, OR every headline in a set is 体言止め, OR it replaces varied endings throughout). Read-aloud (声に出して読む) test is decisive.
- JA-13 = S2 (inherits prose JA-01 threshold; a single concrete 「〜を実現します」can be legitimate).
- JA-14 = S2, escalating to S1 on fabrication (fix = DELETE the invented beat, never invent a replacement).

### §3.2 体言止め Boundary Analysis (CRITICAL — the JA non-transfer finding)
体言止め is a legitimate, long-standing Japanese copywriting technique (fereple: "ライティングに欠かせない手法" usable in キャッチコピー/広告文/タイトル; effects = 強調 + リズム). A detector flagging noun-endings *by presence* will fire constantly on skilled human copy — a guaranteed false-positive machine.
- **Skilled (do NOT flag)**: ONE 体言止め at the single emphasis point amid varied endings (tachibannna: humans use it 狙って/deliberately).
- **AI slop (flag as JA-10)**: 体言止め/言い切り as the *default* ending (fine-tuning reward over-produces it; fereple: 使い過ぎ → ぶつ切り/流れが悪くなる).
- **Does Korean M-2 transfer? — NO.** (1) The noun-ending half does not transfer — 体言止め is native/prestigious; only the over-use/monotony axis carries over, calibrated far more permissively than Korean. (2) The particle-ending half has no clean Japanese analog (the nearest 助詞終わりの倒置・余韻 is itself a legitimate device); so the Korean particle-fragment tell should NOT be ported at all. Net: JA-10 is over-reliance (frequency-gated S2), NOT the Korean M-2 presence rule (S1). Porting M-2 verbatim would systematically mis-flag competent Japanese copywriting.

### §3.3 Japanese high-false-positive signals (do NOT flag)
- A single strategic 体言止め amid varied endings.
- 体言止め in bullet lists, captions, spec sheets, completed metadata titles ("2026年Q1 事業報告").
- Established katakana loanwords with no natural 和語/漢語 substitute (アプリ, サイト, メール).
- Non-sentence colons (time 14:00, ratios 3:1, fixed labels).
- Literary 2-em dash 「──」 used deliberately for 余韻.
- 「〜を実現します」when the claim is concrete and true in a genuine spec line.
- 3人称 informational copy (FAQ, spec, price, dates, business-registration info) — legitimately impersonal.

---

## §4 Chinese Copy Layer — CN-L…CN-Q (web-researched, NOT ported)

Continues the `CN-` letter scheme after CN-K. Copy inverts the prose clustering logic: a headline/CTA/slogan is often a single line with no room to cluster, so several thresholds shift toward single-occurrence decisiveness.

| ID | Tell | Severity | Evidence (verified) |
|----|------|----------|---------------------|
| CN-L | 否定式煽情标题 (negation-contrast headline 这不是…而是…/不仅是…更是…) | S2 (elevated from prose CN-J's S3 for the headline genre) | huxiu, 36kr |
| CN-M | 破折号对比标题 ("X — Y" em-dash headline) | S2 | 36kr; em-dash-norms search |
| CN-N | 落地页万能公式标题 (专为X打造的Y / 开启X之旅 / 解锁全新Y slot-fill templates) | S2 | woshipm, chuhaizhinan |
| CN-O | 强行升华·口号式行动结尾 (让我们携手共创美好未来) | S1 (matches prose CN-D) | 36kr; office-ai (search only) |
| CN-P | 品牌/创业故事模板骨架 (黄金圈 Why-How-What + 三幕式 + 反差构建 + origin myth) | S2 (gate on ≥2 skeleton markers) | chuhaizhinan |
| CN-Q | 虚假具体性·可疑叙事数据 (forced sensory 多模态锚点 + precise-but-unsourced business numbers) | S2 (→S1 when the number is verifiable and material) | csdn-bbs, chuhaizhinan |

### §4.1 Severity rationale
- CN-L is S2 not S3: in prose "不是…而是" reads as natural Chinese in isolation (prose CN-J S3), but in a headline slot it is the whole message — genre lifts severity one step.
- CN-O is S1 (inherits CN-D forced elevation); decisive on a single unprompted uplifting slogan.
- CN-M/CN-N/CN-P/CN-Q are S2. CN-P gates on co-occurrence of ≥2 skeleton markers (any one beat can be legitimate). CN-Q escalates toward S1 when a suspect number is verifiable and material; like CN-K, never replace a fabricated stat with an invented one — delete or downgrade.

### §4.2 对偶/排比 Boundary Analysis (CRITICAL — the ZH non-transfer finding)
Parallelism is a prized, classical, ethnically-native device in Chinese slogans, so a blanket "3+ 排比 = AI" rule (prose CN-F) over-fires badly on copy. The line is **content-driven vs template-driven, not count** (Wikipedia 排比; haoad123):
- 排比 = 3+ clauses of *similar* structure; legitimately reuses words; benefit = 加强语势/节奏感.
- 对偶 = stricter 2-clause couplet, 对称严谨, avoids repeats; hallmark of the 对联 tradition — native, prized.
- Skilled 排比 example: 万科 "感谢冰峰，感谢风暴，感谢悬崖，感谢缺氧。" — concrete, rhythmic, content-driven.
- AI over-balance (36kr; woshipm): balance assembled BEFORE content (先套模板再填内容, 形式凌驾于意义); 用对称句式稀释信息密度; predictable/落窠臼.
- **Operational boundary**: (1) content-first vs template-first — does each parallel member carry a distinct concrete fact, or interchangeable filler around one abstract idea? (2) information density — skilled 排比 concentrates, AI 排比 dilutes. (3) novelty — crafted is 不落窠臼. (4) count is a weak signal on copy — a single crafted 对偶/三段排比 is expected native craft; flag only when blocks stack (3+) OR balance dilutes concrete info. This is the highest-false-positive area.

### §4.3 Dash-contrast applicability (narrow transfer)
- The em-dash-overuse tell transfers (36kr: practitioners "拉黑破折号" as an AI fingerprint + false-positive warnings).
- But "X — Y" binary-contrast is a translationese import: Chinese 破折号 is full-width double-em —— (GB/T 15834); sanctioned functions are 解释说明 / 话题转换 / 声音延长 / 列举 — NONE is English's binary antithesis. So an "X — Y" contrast headline reads doubly-AI.
- **False-positive guard**: a legitimate —— for 解释说明 ("大会堂的枢纽——中央大厅") or 话题转换 must NOT be flagged. Keep CN-M scoped to the binary-contrast headline shape, NOT to every dash.

### §4.4 Chinese high-false-positive signals (do NOT flag standalone)
- A single crafted 对偶 couplet or one 三段 排比 slogan (对联 tradition; 万科 example).
- 破折号 for 解释说明/话题转换/声音延长 (GB/T 15834-sanctioned).
- Verb-first urgency CTA copy (立即购买/免费试用) — standard good practice, not an AI tell.
- Precise numbers that ARE verifiable (real survey n, real dates, citable source).
- Information copy (FAQ/spec/price/business info) — legitimately 3rd-person, no elevation.
- Platform mis-flagging is rampant; require a cluster before asserting AI authorship.

---

## §5 Quarantined Unsourced Hypotheses (NOT catalogue entries — do NOT promote)

These are kept strictly separate. An unsourced pattern MUST NEVER appear in a module's main catalogue (REQ-HUM-012).

### §5.1 English (unsourced)
- **H1** — Alliterative feature-name triads ("Plan. Prioritize. Ship." with forced alliteration). No source isolating alliteration vs the general tricolon (ENC-3).
- **H2** — Fake-precision social proof ("Join 10,000+ teams", "Trusted by 500+ companies"). Whether the round-number "+N" line is specifically an AI copy tell vs a decades-old human convention could not be sourced.
- **H3** — Emoji-led hero bullets / sparkle CTA (`✨ Let's do this!`). Sparkle-emoji is a chat tell (prose EN-I), but no source pins it to landing-page/CTA copy.
- **H4** — Second-person imperative overload ("Imagine…", "Picture this…", "Now think about…" density as a headline tell). Adjacent to oliviacal's "fake experience tells" but no copy-genre source on imperative density.

### §5.2 Japanese (unsourced)
1. "X — Y" bipartite contrast HEADLINE structure as inherently an AI tell in Japanese copy (Korean M-1 analog) — only the punctuation-artifact part (JA-12) is verified; the structural claim is unconfirmed.
2. LP "定番の型" structural template (ファーストビュー→課題提起→…→CTA) as a distinct AI tell — the general "無難/画一的" output IS verified (ai-souken, nahouemura), but the LP-structure-specific framing is unconfirmed.
3. Colon + half-width space (`：␣`) as a standalone high-precision signature — the colon tell is verified (nahouemura), the half-width-space refinement is not.
4. "A から B へ" transition-opener formula (Korean M-3 analog) in Japanese slides/copy — no fetched source; likely real by analogy but unconfirmed.
5. 創業神話の捏造ビート ("創業者が幼少期から…") as a specific recurring template — the general fabrication risk is folded into JA-14's S1 escalation; the "幼少期" cliché as a named pattern is unverified.

### §5.3 Chinese (unsourced)
- CTA button-copy stereotypes (立即开启你的 AI / 开启你的 X 之旅 / 解锁全新体验) — every source treats verb-first CTA copy as good practice, not an AI tell.
- Third-person vs missing second-person address (Korean A-25 analog: 系统/AI 会… instead of 你/您…) — no Chinese source names it.
- Fake customer-testimonial slop (Korean L-8 analog) — partially subsumed by CN-Q; a standalone testimonial-format tell is unsourced.
- "这价格要啥自行车"-style forced-casual authenticity as its own inverse tell — mentioned as a *fix* in a 去-AI-味 guide, not catalogued as a tell.

---

## §6 Verified Sources (two tiers: fetched-and-confirmed URLs, plus clearly-marked corroborating aggregates)

Tier 1 = a URL that was fetched and confirmed to contain the cited claim (the primary evidence for every catalogue entry). Tier 2 = a WebSearch aggregate used ONLY to corroborate a claim already verified in a Tier-1 fetched page — an aggregate is never the sole basis for a main-catalogue entry. The one entry whose direct citation is a Tier-2 aggregate (ENC-7) is backstopped by a Tier-1 fetched source per the §2 [†] footnote.

### §6.1 English
- gc.ai/blog/ai-writing-pattern-to-know-contrastive-negation — "contrastive negation" ("not X but Y"), 6 marketing examples → ENC-2.
- blakestockton.com/dont-write-like-ai-1-101-negation — "It's Not X, it's Y" ubiquitous tell → ENC-2.
- firstcall-web.com/5-sentence-patterns-that-make-ai-copy-instantly-recognizable — Contrast Framing (ENC-2), Triadic (ENC-3), Landscaping (ENC-4), Confirmational Authority (ENC-9), Motivational Closes (ENC-5).
- onlygoodcontent.com/post/spot-the-tell-8-signs-ai-wrote-that-copy — trio "Fast. Simple. Scalable." (ENC-3), "Empowering innovation" (ENC-6), contrast formulas (ENC-2), poster/mug closers (ENC-5).
- oliviacal.com/post/ai-writing-tells — verb blacklist (unleash, empower, ignite…), openers ("In today's fast-paced digital world") → ENC-1, ENC-4.
- alyssawiens.com/2025/03/27/how-can-you-tell-if-writing-is-ai-generated — em-dash unreliability; "Whether X or Y" → ENC-8 + em-dash false-positive.
- WebSearch aggregates (landing-page clichés; CTA microcopy copy.ai/matchbox/woobox; how-to-tell/em-dash 30% caveat) → corroborate ENC-1/6/7 + slide "From X to Y". (Aggregates used only to corroborate quotes already verified in fetched pages above.)

### §6.2 Japanese
1. note.com/tachibannna/n/n21a0b3a2ffa1 — AI overuses 体言止め/言い切り; humans use it 狙って → JA-10.
2. cooqieinc.com/blog/how-americans-detect-ai-writing — em-dash a distinctive AI signature in English → JA-12 mechanism.
3. ai-souken.com/article/detecting-chatgpt-generated-text — katakana/business-loanword overuse, weak causality, absence of first-hand experience → JA-13, JA-09 amplification.
4. note.com/nahouemura/n/n9e519ef21b17 — 「日本語では文末にコロンをほとんど使わない」; bullets 「取扱説明書のよう」→ JA-11.
5. fereple.com/writers-apc/stop-talking — 体言止め essential for キャッチコピー but 使い過ぎ → 違和感 → JA-10 boundary.
6. twelfth.jp/ai/brand-story — AI brand story cold/generic without 一次情報 → JA-14.
(Corroborating, NOT independently fetched — not cited as evidence: mieru-ca, yosca, kenkyo.ai, malna, webrandum + kakuyomu/ncode writing guides on ダッシュ conventions.)

### §6.3 Chinese
- 36氪 (36kr.com/p/3824601267196037) — 跑鞋 "这不仅仅是…而是…" (CN-L); em-dash "拉黑破折号" fingerprint (CN-M); forced elevation "将次要事件拔高为分水岭时刻" (CN-O); 排比 overload (CN-F boundary).
- 虎嗅 (huxiu.com/article/4823754.html) — "不是…而是" AI 八股文 (CN-L).
- CSDN (bbs.csdn.net/…/100167916) — 多模态锚点 forced sensory detail; 冲突密度阈值; 虚假具体性 (CN-Q).
- 出海指南 (chuhaizhinan.com/2025/03/29/brand-story-guideline) — 黄金圈 Why-How-What; 反差构建; 细节颗粒度; 万能公式 (CN-P, CN-N, CN-Q).
- 人人都是产品经理 (woshipm.com/share/6060495.html) — 空洞堆砌形容词; 段末总结套路; 冷冰冰无人味 (CN-N, corroborates genre transfer).
- 维基百科 排比 (zh.wikipedia.org/zh-hans/排比) — 排比 (3+, similar-not-identical) vs 对偶 (2, strict, avoids repeats) → boundary.
- 好文案 (haoad123.com/article/3209.html) — 排比 in ad copy; 万科 slogan as prized example → boundary skilled example.
(Search-corroborated, NOT independently fetched, cited only as secondary support: neican.ai 去味指南; zhihu AI标点用法; 1zhengji 广告词对偶排比; office-ai.cn 让我们共创; 36kr 高考作文实测; GB/T 15834 破折号规范 blogs.)

### §6.4 Korean (port source — provenance, not shipped)
The Korean copy layer's provenance is the `general-humanize-korean` taxonomy (`ai-tell-taxonomy.md` §A-20…A-25 / §L / §M) + its SKILL.md copy-mode guard + dual grading tables. That taxonomy cites its own academic anchors (arXiv 2402.01536 ACM C&C'24 Homogenization; arXiv 2501.19361; arXiv 2604.03136 StoryScope; longblack.co brand storytelling; YTN fact-check; StoryBrand) plus Korean copy-genre field observation of 14 before/after samples. These provenance citations are summarized here and stripped from the shipped Korean module per §25.

---

## §7 Three Non-Transfer Findings (summary — the load-bearing design constraint)

1. **Korean M-2 → English: does NOT transfer.** English headlines are natively fragmentary; fragment-ness is a high-false-positive signal, not a removable tell (§2.2, §2.3). ENC layer has NO standalone fragment-headline category.
2. **Korean M-2 → Japanese: does NOT transfer as written.** 体言止め is a prestigious native device; JA-10 is a frequency-gated S2 (over-reliance), not a presence S1; the Korean particle-ending half has no Japanese analog and is not ported (§3.2).
3. **Chinese 对偶/排比: content-first vs template-first, not count.** Parallelism is prized classical craft; the boundary is information density and content-drivenness, not occurrence count; dash-contrast transfers only to the binary-contrast headline shape, since 破折号 —— marks explanation/topic-shift, never binary contrast (§4.2, §4.3).

Consequence encoded as REQ-HUM-007: mechanically translating the Korean copy layer into the other three modules is PROHIBITED — it would cause systematic false positives.
