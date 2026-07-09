# Copy Review — Post-Generation QA Gate

This module turns the skill into a **review-only QA gate** for copy produced by another tool or workflow — a design tool's result copy, a landing-page generator's hero section, a deck builder's slide titles, or copy lines pasted out of a chat session. It detects formulaic AI copy against per-language formula dictionaries plus the language modules' tell catalogues, and **proposes** fixes; it never applies them.

This module reuses the skill's shared machinery: the S1/S2/S3 severity model, the copy-mode grade table, and the fact-anchor preservation guard all live in `SKILL.md` and apply here unchanged. This module defines no parallel grading system.

---

## Review-Only Gate Mode

When the skill is invoked as a post-generation QA gate, it operates in **review-only** mode — distinct from the skill's default rewrite mode:

| Mode | What it returns | When |
|------|-----------------|------|
| Default rewrite mode (SKILL.md Output Contract) | The humanized text + change report | The user asked for the text to be fixed |
| **Review-only gate mode (this module)** | A review report with fix proposals — the original text is returned untouched | The copy came from another tool/workflow and the user reviews before applying |

Rules of the gate:

- **Detect and propose — never auto-apply.** No rewrite is applied to the source copy; every proposal waits for the user's selection. Automatic batch application is prohibited even when every finding is S1.
- **The user reviews before application.** After the user accepts proposals, the accepted rewrites flow through the skill's default rewrite mode (copy mode), where the fact-anchor preservation guard verifies them.
- **Report everything, including uncertain findings.** Context-dependent findings are reported with their context question attached rather than silently dropped or silently decided.

---

## Six-Stage Review Pipeline

### Stage 1 — Copy collection

Accept the copy under review in one of four input classes:

| Input class | Example |
|-------------|---------|
| Single copy text | One headline, CTA label, or tagline pasted directly |
| Copy bundle | Hero + feature blurbs + CTA + footer collected as one set |
| Full exported document text | Text extracted from an exported page, PDF, or deck |
| Pasted chat extract | Copy lines the user picked out of a generation-tool chat |

For bundles and full documents, record each copy item's slot (hero headline, subheadline, feature card, CTA, footer) — severity judgments are slot-sensitive.

### Stage 2 — Language detection and dictionary routing

Detect the language of each copy item and route it to its formula dictionary below AND to the matching language module's copy-layer catalogue (`modules/korean.md`, `modules/english.md`, `modules/japanese.md`, `modules/chinese.md`). For mixed-language bundles, route per item; for a single mixed-language item, route per span.

### Stage 3 — Pattern matching

Match each item against, in order:

1. The per-language formula dictionary in this module (slot formulas).
2. The language module's copy-layer catalogue (structural and rhetorical tells).
3. The display-copy genre rules in `modules/design-copy.md` when the surface is a landing page, card deck, or slide set.

Near-decisive entries (S1, and S2 at their stated escalation conditions) are flagged directly. Context-dependent entries (S3, context-gated S2) are queued for Stage 4 instead of being flagged outright.

### Stage 4 — Context inference

Context-dependent findings resolve against three inputs:

- **Industry register** — the same polish adverb can be routine in one industry's copy register and formulaic in another's.
- **Brand tone** — a deliberately energetic brand legitimately uses intensity that would read as filler for a reserved brand.
- **Audience / persona** — ease-of-use framing that suits a consumer audience can undercut copy aimed at expert buyers.

When industry, brand tone, or audience were not provided with the request, do NOT decide silently: mark the finding **context needed** in the report and state the question the user must answer. Never guess a brand's intent.

### Stage 5 — Fix proposal

Emit one fix proposal per confirmed finding, in the Fix-Proposal Format below.

### Stage 6 — Review report emission

Assemble all findings into the Review Report Template below and return it with the untouched original copy.

---

## Fix-Proposal Format

Every finding carries exactly these four parts:

```
Original:  <the exact copy span, quoted verbatim>
Reason:    <the named pattern — a dictionary ID or catalogue category ID — plus one line of why>
Alternatives (at least 3):
  1. <rewrite candidate>
  2. <rewrite candidate>
  3. <rewrite candidate>
Preferred: <one pick> — <one-line justification>
```

Rules:

- **Alternatives are concrete-fact-oriented.** Each candidate replaces the formula with a mechanism, a measure, or a named outcome.
- **Use explicit placeholders when no real fact is available**: `[number]`, `[persona]`, `[task]`, `[outcome]`, `[product]`. Never invent a number, name, or statistic to fill a slot — the fact-anchor preservation guard applies to proposals too.
- **Anchors survive verbatim.** Any real number, date, price, or proper noun in the original appears character-intact in every alternative.

Worked example:

```
Original:  "Reimagine your workflow with AI-powered automation"
Reason:    english.md ENC-1 (aspirational-verb headline); the tech base is also offered as the benefit
Alternatives (at least 3):
  1. "[product] automates the [task] you repeat every day"
  2. "Cut the time you spend on [task] by [number]"
  3. "[outcome] automation built for [persona]"
Preferred: 1 — verb-first and names the concrete task instead of promising reimagination.
```

---

## Review Report Template

```markdown
# Copy Review Report

## At a Glance

| Item | Result |
|------|--------|
| Copy items reviewed | N |
| S1 findings (decisive) | N |
| S2 findings (strong) | N |
| S3 / context-gated findings | N |
| Context needed | N |
| Total fix proposals | N |

## S1 Findings (fix required)

### <slot — e.g. hero headline>
<fix proposal in the Fix-Proposal Format>

## S2 Findings (fix recommended)

### <slot>
<fix proposal>

## Context Needed

- <finding> — resolve by answering: <industry / brand tone / audience question>

## Follow-Up

1. Apply the accepted proposals through this skill's default rewrite mode (copy mode, fact-anchor guard active).
2. For display-surface copy, run the genre checklist in modules/design-copy.md.
3. Re-run this gate on the revised copy set.
```

The summary table always leads; a reader decides in ten seconds whether the copy set ships or goes back.

---

## Severity Mapping

The formula dictionaries below use ONLY the shared S1/S2/S3 severity model from `SKILL.md`. Source vocabularies that classify formulas as "near-always" versus "context-dependent" map as follows:

- **Near-always formulas** → **S1** when a single occurrence at a headline, hero, or CTA slot is decisive; **S2** when density or stacking is what convicts.
- **Context-dependent expressions** → **S3** (downgrade-only contributor), or **context-gated S2** when Stage 4 context (industry, brand tone, audience) confirms the mismatch.

Grading of a reviewed set follows the copy-mode grade table in `SKILL.md` — this module adds no grade table of its own.

---

## Repair-Strategy Playbook (CRS-1…CRS-7)

Seven named repair strategies. Fix proposals reference them by ID so the user can see *how* an alternative was constructed, not just *what* it says.

### CRS-1 — Voice concretization

Replace generalized claims with a specific actor, case, or result. The writer's judgment and experience must be visible.

- Before: 이 기능은 매우 효율적입니다.
- After: 이 기능을 붙인 뒤 우리 팀 주간 회의가 30분 줄었습니다.

### CRS-2 — Rhythm variation

Break uniform sentence length. One short sentence makes the emphasis point; the next sentence carries the context.

- Before: This tool improves productivity significantly. It also reduces manual work considerably. It helps teams collaborate more effectively.
- After: Most of the busywork disappears. What's left is the part your team actually wants to do — reviewing edge cases and shipping.

### CRS-3 — Transition-word diet

Delete stacked connectives; keep a transition only where the logic genuinely turns.

- Before: 또한 가격이 합리적입니다. 더불어 설치가 간단합니다. 한편 지원도 빠릅니다.
- After: 가격이 합리적이고 설치는 5분이면 끝납니다. 지원 답변은 하루를 넘긴 적이 없습니다.

### CRS-4 — List-to-prose conversion

Three or fewer items read better as one sentence; keep a list only for genuine comparison or ordered steps.

- Before: • 정기 미팅 실시 • 명확한 역할 분담 • 피드백 문화 조성
- After: 매주 15분 스탠드업만 남기고, 담당 범위를 문서로 공유하고, 결과물에 솔직한 피드백을 주고받았습니다.

### CRS-5 — Specificity injection

Swap vague quantities for checkable figures — or explicit placeholders when the real figure is not in the source.

- Before: 많은 사용자들이 만족하고 있습니다.
- After: 설문 응답자 [number]명 중 [number]명이 재구매 의사를 밝혔습니다.

### CRS-6 — Opening rewrite

The generated opening is almost always disposable. Open with the reader's question, an unexpected fact, or the conclusion itself.

- Before: In today's fast-paced business environment, effective communication is more important than ever.
- After: We cut our meetings in half last quarter. Work got faster, not slower.

### CRS-7 — Closing rewrite

Never close by restating the body. Close on the next action, an open question, or one line addressed to the reader.

- Before: 지금까지 살펴본 바와 같이 자동화는 매우 중요한 의미를 가집니다.
- After: 오늘 반복한 작업 하나만 골라 자동화해 보세요. 내일 아침이 달라집니다.

---

## Formula Dictionary — Korean

Slot formulas native to Korean marketing copy. Structural tells are already catalogued in `modules/korean.md` and are cross-referenced, never re-defined here: dash-contrast headline → M-1, particle/noun-fragment title → M-2, from-X-to-Y opener → M-3, founder-myth arc → L-6, unverifiable narrative statistics → L-7, third-person expository voice in appeal copy → A-25. The entries below are net-new formula patterns.

| ID | Formula pattern | Detection signal | Severity | Rewrite direction (example) |
|----|-----------------|------------------|----------|------------------------------|
| CR-KO-1 | 자화자찬 형용사 슬롯 — "혁신적인 [X]", "차세대 [X]", "재정의하는 [X]" | Self-praise adjective directly modifying the product noun at a headline slot | S2 — escalates to S1 when two or more stack in one headline ("혁신적인 AI 기반의 차세대 …") | State the mechanism or measured result: "혁신적인 마케팅 자동화" → "주간 보고서 작성 30분이 0분이 됩니다" |
| CR-KO-2 | 기술 기반 과시 — "AI 기반의 [X]" | The technology base presented as the benefit itself, with no user-visible outcome | S2 | Name what the technology does: "AI 기반의 리뷰 분석" → "리뷰 3,000건을 1분 안에 분류합니다" |
| CR-KO-3 | 초월 약속 공식 — "지금까지 없던 [X]", "한 차원 높은 [X]", "당신의 [X]를 변화시킬" | Unprecedented/elevation promise with an interchangeable product slot and no checkable claim | S1 at a headline or hero slot | Replace the promise with the concrete differentiator: "지금까지 없던 자동화" → "설정 없이 첫 실행까지 5분" |
| CR-KO-4 | 시대 선언 공식 — "이제는 [X]의 시대", "새로운 패러다임" | Era-declaration frame opening a headline; any product noun fits the slot | S1 | Drop the era frame and state the claim: "이제는 자동화의 시대" → "손으로 붙여넣던 일을 오늘 끊으세요" |
| CR-KO-5 | 부정 해소 호소 — "더 이상 [X]에 시달리지 마세요" | Negative-relief appeal formula at a headline or CTA slot | Context-gated S2 — pain-point framing is legitimate when the pain is named concretely; the bare formula is the tell | Name the relief in checkable terms: "더 이상 야근에 시달리지 마세요" → "마감 정산이 [number]분 만에 끝납니다" |
| CR-KO-6 | 매끄러움 부사 군집 — "원활하게", "손쉽게", "직관적으로", "단숨에" | Polish adverbs replacing evidence of ease; convicts on density, not single use | S3 — density-gated | Replace with a measurable ease claim: "손쉽게 시작하세요" → "가입부터 첫 결과까지 30초" |
| CR-KO-7 | 형용사·부사 나열 슬로건 — "빠르게. 쉽게. 정확하게. 강력하게." | Three or more interchangeable modifiers listed as the whole message | S2 | Trade the list for one verifiable outcome: → "응답 시간이 12시간에서 30분으로 줄었습니다" |
| CR-KO-8 | 리스티클 유인 상투구 — "꼭 알아야 할 [X]", "[X] 완벽 가이드", "[X]의 세계로 빠져보세요" | Listicle-teaser filler at a cover or title slot; the invitation-calque variants also fall under the prose translationese category in `modules/korean.md` | S2 (covers and titles) | Promise something concrete: "마케터가 꼭 알아야 할 자동화" → "마케터가 이번 주에 자동화할 수 있는 업무 3가지" |

Cross-language note: the Korean rendering of the contrastive-negation headline ("이제 [X]는 그만. [Y]로.") is judged under the same manufactured-contrast logic as its English counterpart (english.md ENC-2) — treat it as CR-KO-5-adjacent and require the concrete relief.

---

## Formula Dictionary — English

Slot formulas for English marketing copy. Families already catalogued in `modules/english.md` are cross-referenced, never re-defined here: the aspirational-verb headline family ("Unleash your [X]", "Transform the way you [X]", "Supercharge [X]", "Empower your team to [X]", "Revolutionize [X]", "Reimagine your [X]") → ENC-1; the contrastive-negation headline ("No more [X]. Just [Y].", "It's not just [X] — it's [Y]") → ENC-2; the interchangeable-modifier trio ("Fast. Simple. Scalable.") → ENC-3; body-prose focal verbs ("leverage", "synergy") → the prose layer of `modules/english.md`. The entries below are net-new formula patterns.

| ID | Formula pattern | Detection signal | Severity | Rewrite direction (example) |
|----|-----------------|------------------|----------|------------------------------|
| CR-EN-1 | Tech-base-as-benefit — "Powered by AI", "AI-powered [X]" as the lead claim | The technology base offered as the benefit itself, with no user-visible outcome attached | S2 | Name what the technology does, with a measure: "AI-powered analytics" → "Sorts [number] reviews a minute so you read only the outliers" |
| CR-EN-2 | Futurity self-praise adjectives — "Next-generation [X]", "Cutting-edge [X]", "State-of-the-art [X]" | Futurity/novelty adjective directly modifying the product noun at a headline slot | S2 — escalates when two or more stack in one headline | State the concrete capability: "Next-generation deployment platform" → "Deploys on every merge, rolls back in one click" |
| CR-EN-3 | Intersection formula — "Where [X] meets [Y]" | The meets-frame headline with two interchangeable abstract nouns | S1 at a headline or hero slot | Say what the combination actually does: "Where design meets automation" → "Design once — the layout rebuilds itself for every screen" |
| CR-EN-4 | Future-declaration formula — "Built for the future of [X]", "The future of [X] is here" | Future-of frame carrying no present, checkable claim | S1 at a headline or hero slot | Make the claim present-tense and verifiable: "The future of invoicing is here" → "Invoices reconcile themselves the day they arrive" |
| CR-EN-5 | Appositive tagline — "[X], simplified.", "[X], reimagined.", "[X], unleashed." | Noun + comma + past-participle tagline with an interchangeable product slot | S2 | Say what got simpler and by how much: "Expenses, simplified." → "Expense reports in [number] minutes, straight from your inbox" |
| CR-EN-6 | Colon-payoff headline — "[Headline]: The [X] that finally [Y]" | Colon staging plus a "finally" payoff that adds no checkable content | S2 | Drop the staging; keep the claim: "Meet Flow: The tool that finally fixes meetings" → "Cuts your weekly meeting load by [number] hours" (the sentence-final/heading colon in Japanese copy is catalogued separately at japanese.md JA-11) |
| CR-EN-7 | Unverifiable scale claim — "Trusted by thousands of [teams/customers]", "Loved by users worldwide" | Scale/trust claim with no number, name, or source | S2 | Use a checkable figure or placeholder: → "Trusted by [number] teams, including [named customer]" (narrative analogs: korean.md L-7, chinese.md CN-Q) |
| CR-EN-8 | Superlative self-praise cluster — "Game-changing", "Best-in-class", "World-class" | Superlatives replacing evidence; convicts on density, not single use | S3 — density-gated | Replace with the comparison basis: "Best-in-class support" → "Median support reply under [number] minutes" |
| CR-EN-9 | Polish-adverb cluster — "Seamlessly", "Effortlessly", "Intuitively" | Ease adverbs replacing evidence of ease; the same adverb can be routine in one industry's register and formulaic in another's | S3 — context-gated S2 via Stage 4 | Replace with a measurable ease claim: "Integrates seamlessly" → "Connects to [product] in one click — no setup steps" |

---

## Formula Dictionary — Japanese

Entries in this dictionary are copy-slot formula instances of tells already catalogued in `modules/japanese.md`; the parent category is named per row, following the same parent-naming convention that module uses for its own copy layer. Structural tells stay owned by their catalogue entries and are cross-referenced, never re-defined: the noun-ending frequency gate → JA-10, the imported dash → JA-12, the bare formulaic promise endings 「〜を実現します」「〜を可能にします」 → JA-13, the brand-story arc → JA-14.

| ID | Formula pattern | Detection signal | Severity | Parent | Rewrite direction (example) |
|----|-----------------|------------------|----------|--------|------------------------------|
| CR-JA-1 | コロン式見出し — 「[名詞]：[ベネフィット]」 ("業務効率化：チームのための新しいツール") | A heading colon splitting the headline into label + benefit — an imported heading shape | S1 on sight (inherits the parent's verdict) | JA-11 | Restructure natively without the colon: 「業務効率化：導入は簡単」 → 「導入は半日、翌週から定時で帰れます」 |
| CR-JA-2 | 変革・格上げ約束公式 — 「あなたの[X]を変える」「[X]の常識を変える」「ワンランク上の[X]」 | Transformation/elevation promise with an interchangeable product slot and an abstract, ownerless benefit | S2 — allow one; flag when stacked or when the benefit stays abstract | JA-13 (formulaic-promise family) | Promise a concrete outcome: 「あなたの働き方を変えるツール」 → 「議事録の清書がなくなり、毎週2時間戻ってきます」 |
| CR-JA-3 | カタカナ・バズワード群見出し — 「シームレスな[X]」「オールインワン[X]ソリューション」 | Loanword benefit-cluster at a headline slot where 和語/漢語 plus a concrete claim would carry more | S3 — context-gated S2 when two or more cluster in one headline | JA-09 (katakana overload, genre instance) | Replace the cluster with a concrete claim: 「シームレスな連携を実現」 → 「二つのアプリを自動でつなぎます。設定は3分です」 |
| CR-JA-4 | 実績ぼかし公式 — 「多くの企業が導入」「数多くの実績」 | Adoption/scale claim with no number, name, or named case at a trust-signal slot | S2 | JA-06 (abstraction / no concrete detail) | Use a checkable figure or placeholder: 「多くの企業が導入しています」 → 「製造業を中心に[数]社が導入、継続率は[数]%です」 |

---

## Formula Dictionary — Chinese

Entries in this dictionary are copy-slot formula instances of tells already catalogued in `modules/chinese.md`; the parent category is named per row. Structural tells stay owned by their catalogue entries and are cross-referenced, never re-defined: the negation-contrast headline → CN-L, the dash-contrast headline → CN-M, the slot-fill landing family (专为[X]打造的[Y] / 开启[X]之旅 / 解锁全新[Y]) → CN-N, the forced-elevation slogan close → CN-O, the brand-story skeleton → CN-P.

| ID | Formula pattern | Detection signal | Severity | Parent | Rewrite direction (example) |
|----|-----------------|------------------|----------|--------|------------------------------|
| CR-ZH-1 | 重新定义公式 — 「重新定义[X]」 | The redefine-frame headline accepting any product noun, with no concrete claim | S2 — escalates to S1 when stacked with another slot formula in the same hero | CN-N (slot-fill landing family) | State the actual differentiator: 「重新定义团队协作」 → 「审批从三天缩到十分钟，进度一个页面全看到」 |
| CR-ZH-2 | 赋能类黑话标题 — 「赋能[X]」「为[X]赋能」「构建[X]新生态」 | Image-less business jargon carried into a headline slot | S1 (inherits the parent's on-sight verdict) | CN-C (jargon / abstract nouns) | Say what it actually does: 「赋能中小企业增长」 → 「帮中小卖家把上架时间缩短一半」 |
| CR-ZH-3 | 技术·未来卖点公式 — 「AI驱动的[X]」「智能[X]新体验」「下一代[X]」 | The technology base or futurity claim offered as the benefit itself, with no user-visible outcome | S2 | CN-N family (tech/futurity slot variants) | Name the capability with a number: 「AI驱动的客服平台」 → 「客服首次回复不超过30秒」 |
| CR-ZH-4 | 时代宣言公式 — 「开启[X]新时代」「引领[X]新纪元」 | Era-declaration frame at a headline; grand significance with no checkable content | S2 at a headline slot; S1 at a closing slogan slot (there it matches CN-O) | CN-N (开启-family) + CN-D (forced elevation) | Close on the concrete next step: 「开启智能办公新时代」 → 「今天下午的报销单，十分钟批完」 |
| CR-ZH-5 | 无数据规模宣称 — 「数千家企业的信赖之选」「众多用户的共同选择」 | Impressive-but-unnumbered scale claim — the mirror case of precise-but-unsourced numbers | S2 | CN-Q (fake specificity, mirror case) | Use a checkable count or placeholder: 「数千家企业的信赖之选」 → 「[数字]家企业在用，上季度续约率[数字]%」 |

---

## Sources and Grounding

- The review pipeline, the fix-proposal format, the report template, the repair-strategy playbook, and the Korean and English formula dictionaries are adapted from the maintainer's own copy-QA pattern dictionaries and post-generation review guides (same authorship as this skill; direct reuse).
- The Japanese and Chinese dictionaries contain no independently asserted language claims: every CR-JA and CR-ZH entry is a copy-slot formula instance of a tell already catalogued and independently researched in `modules/japanese.md` or `modules/chinese.md`, with the parent category named in the entry's row. An entry that could not be parent-mapped would carry an explicit grounding note in this section instead; currently none is needed.
- Dedup policy: any pattern already catalogued in a language module is cross-referenced by its category ID and never re-defined here. New entries are formula-level patterns the catalogues describe only generically, or do not carry.
- Severity values follow the shared model in `SKILL.md`; the mapping rules from source confidence vocabularies are stated in the Severity Mapping section above.
