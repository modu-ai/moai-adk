# Humanize — Chinese (中文, Simplified)

This module catalogues AI-text "tells" specific to Simplified-Chinese model output and the rules for humanizing (윤문) it without distorting meaning. Detectors are unreliable for Chinese (academic Chinese detectors score ~40% on ChatGPT-level text; the OpenAI classifier reports ~26% sensitivity), so this catalogue is built for **pattern-based human/agent editing, not as a detection oracle**. A single tell rarely proves AI authorship — confidence comes from **clustering** (multiple S1/S2 patterns co-occurring), which is why several categories gate on repetition and overlap rather than single occurrence.

All instruction prose here is English; every before/after rewrite example is in Simplified Chinese, because the tells are language-specific and cannot be demonstrated in English.

## Detection Categories

| Category | Tell | Why it reads as AI | Severity |
|----------|------|--------------------|----------|
| **CN-A · 机械逻辑连接词** (Mechanical Logical Connectors) | Formulaic discourse markers stacked at sentence/paragraph heads: 首先、其次、再者、最后；然而、此外、因此、由此可见、综上所述、值得注意的是、不可否认、毋庸置疑、总而言之、换句话说、基于此、在此基础上。 | LLMs scaffold a "fake logic chain" with explicit connectors to look structured. Humans rely far more on implicit cohesion; this dense connector layering is the single most-cited Chinese AI tell. | **S1** |
| **CN-B · 三段式 / 总—分—总 模板结构** (Three-Part Template Structure) | Rigid 总—分—总 (intro–body–conclusion) skeleton; "提出问题—分析问题—解决问题"; sections literally titled 挑战与展望 / 机遇与挑战。 | The model defaults to a safe outline template. Different models given the same prompt produce near-identical openings and skeletons — a fingerprint humans don't share. | **S2** |
| **CN-C · 互联网黑话 / 抽象名词** (Corporate Jargon & Abstract Nouns) | 赋能、抓手、闭环、生态、底层逻辑、全维度、深度剖析、深入探讨、维度、量子纠缠 used metaphorically. | These abstract, image-less buzzwords are statistically "safe" high-probability tokens. Editor's rule: if you can't draw it as a picture, don't write it — AI fails this test constantly. | **S1** |
| **CN-D · 强行升华 / 正能量结尾** (Forced Elevation & Positive-Energy Endings) | Endings that inflate a mundane topic to grand moral/historical significance: 在当今……的时代、不仅……更是一种……、彰显了、具有重要意义、为……注入了新的活力、让我们携手共创美好未来。 | RLHF biases the model toward an agreeable, "eager pupil seeking praise" tone. The unprompted uplifting summary is one of the most recognizable giveaways. | **S1** |
| **CN-E · 句式节奏单调** (Monotonous Sentence Rhythm / Low Burstiness) | Uniform sentence length and structure; every sentence roughly the same weight; no short punchy sentences mixed with long ones. | Low perplexity/burstiness — AI regresses to mean sentence length. Human writing is uneven (不均匀); AI rhythm is a flat robotic drone. | **S2** |
| **CN-F · 对偶 / 排比 修辞过载** (Antithesis & Parallelism Overload) | Rule-of-three lists (形容词、形容词、形容词 or 短语、短语、短语); ornate antithetical couplets ("昔日先贤上下求索……；今朝我辈接续奋斗……")。 | Largest measured rhetorical gap — AI averages ~4 对偶句 vs students' ~0.67; 排比 ~1.5 vs near-zero. Over-balanced rhetoric reads as machine-perfect. | **S2** |
| **CN-G · 翻译腔 / 书面化僵硬** (Translationese & Over-Formality) | 进行了……的探讨、对……进行分析、有着、无可替代、起到了……的作用; nominalized verbs; English-syntax calques. | Heavy noun ratio and verb-nominalization mirror translated/academic register, not natural spoken Chinese. | **S2** |
| **CN-H · 套话填充词 / 学术高频词** (Filler Clichés & Academic High-Frequency Words) | 至关重要、极具潜力、无与伦比、充满活力、深入研究了、需要指出的是、显而易见、众所周知；grand concept-words 智慧、时代、人生、力量、成就。 | Corpus studies find roughly 1 in 7 papers carry these LLM signature words. Empty intensifiers with no concrete referent. | **S2** |
| **CN-I · 格式滥用** (Formatting Abuse — Bold / Lists / Emoji / Em-dash) | Excessive **粗体强调**; over-reliance on bullet lists for prose; emoji in headings; markdown residue (反引号、星号) leaking into plain text. | Inherited from README/manual/marketing training data. Community cleanup projects flag bold abuse and emoji-in-headings as primary structural tells. | **S2** |
| **CN-J · "不是……而是" / 二元对立句式** (Negation-Contrast Frame) | 这不是……而是……; 与其说……不如说……; 不仅是……更是…… used reflexively as a rhetorical default. | A signature "ghost" pattern — appears unbidden as the model's go-to way to manufacture insight/contrast. | **S3** |
| **CN-K · 引用与事实造假** (Fabricated Citations & Attribution) | Invented ISBN/DOI, broken hyperlinks, `utm_source=chatgpt.com` in URLs, vague "专家指出""研究表明" with no specific source, ghost references. | Hallucinated authority markers. High-severity because it is verifiable and damaging. | **S1** |

## Severity Rules

Severity sets the removal threshold for each category. Counts are measured across the whole text being edited.

**S1 — critical, remove on single occurrence.** Categories: **CN-A, CN-C, CN-D, CN-K**.
- CN-A: remove on a single head-of-sentence connector cluster (e.g. 首先…其次…综上所述).
- CN-C: remove on a single image-less buzzword (赋能/抓手/闭环-class).
- CN-D: remove on a single unprompted uplifting moral summary.
- CN-K: remove on a single fabricated DOI/ISBN/utm-tagged URL or ghost-expert attribution. Never replace a fabricated citation with an invented one — only delete or downgrade to an unattributed claim.

**S2 — strong, allow 1–2, remove at 3+.** Categories: **CN-B, CN-E, CN-F, CN-G, CN-H, CN-I**.
- CN-B: allow loose structure; remove the template only when 总—分—总 *and* a titled 挑战与展望/机遇与挑战 section co-occur.
- CN-E: flag at sustained uniformity; require sentence-length variation once 3+ paragraphs run at the same rhythm.
- CN-F: allow 1–2 rhetorical figures; remove at 3+ 对偶/排比.
- CN-G: allow 1–2 nominalized/calque constructions; remove at 3+.
- CN-H: allow 1–2 intensifiers; remove at 3+ empty 至关重要-class words.
- CN-I: allow purposeful formatting; remove bold/list/emoji abuse at 3+ instances.

**S3 — weak, problematic only when overlapping.** Category: **CN-J**.
- CN-J reads as natural Chinese in isolation. Treat it as a tell only when it overlaps a CN-D or CN-F cluster (i.e. it is being used to manufacture forced contrast or balanced rhetoric). It affects the grade only when stacked on an S1/S2 cluster.

**Clustering principle:** Prioritize removing co-occurring tells. A passage where CN-A + CN-D + CN-F appear together is a far stronger signal than any one of them alone, and should be the first target of an edit pass.

## Rewrite Examples

### CN-A · Mechanical Connectors
- **AI:** 首先，我们需要明确目标。其次，制定详细计划。最后，综上所述，执行是关键。
- **人:** 先把目标定清楚，再排出计划——剩下的就看执行了。

- **AI:** 然而，值得注意的是，此外，由此可见这一策略行之有效。
- **人:** 这个策略确实管用。

### CN-C · Corporate Jargon / Abstract Nouns
- **AI:** 这套方案能够赋能团队，打造业务闭环，实现全维度增长。
- **人:** 这套方案让团队多接了三成订单，从下单到回款一条龙跑通了。

- **AI:** 我们要找到撬动增长的抓手，构建可持续的生态。
- **人:** 我们得找个能真正带量的渠道，先把第一批客户留住。

### CN-D · Forced Elevation / Positive Endings
- **AI:** 在当今这个飞速发展的时代，淄博烧烤的出圈不仅是美食的胜利，更彰显了城市精神的力量，让我们携手共创美好未来。
- **人:** 说到底，淄博烧烤火起来，就是把人情味做到了别人没做到的地步。

- **AI:** 总而言之，情绪管理在现代社会具有至关重要的意义。
- **人:** 地铁上那个憋着没发火的上班族，大概比谁都懂情绪管理值多少钱。

### CN-F · Antithesis / Parallelism Overload
- **AI:** 它是希望的灯塔，是前行的号角，是奋斗的旗帜。昔日先贤上下求索，今朝我辈接续奋斗。
- **人:** 它给了不少人一个继续干下去的理由。

- **AI:** 创新驱动发展，发展引领未来，未来成就梦想。
- **人:** 靠创新先活下来，剩下的慢慢再说。

### CN-E / CN-G · Rhythm + Translationese
- **AI:** 本文对该问题进行了深入的探讨，并对相关因素进行了系统的分析，有着重要的参考价值。
- **人:** 这篇文章我把问题翻了个底朝天，几个关键因素也都掰开看了，值得一读。

### CN-H · Filler Clichés
- **AI:** 这项技术极具潜力，无与伦比，必将充满活力地改变众所周知的行业格局。
- **人:** 这项技术还在早期，但已经能看出来它会动到行业的老规矩。

### CN-I · Formatting Abuse
- **AI:**
  我们的产品具备以下**核心优势**：
  - ✅ **高效**
  - ✅ **稳定**
  - ✅ **智能** 🚀
- **人:** 我们的产品跑得快、不崩、还省心。

### CN-J · Negation-Contrast Frame (only when overlapping)
- **AI:** 这不仅是一次更新，更是一场理念的革命，彰显了我们对未来的承诺。
- **人:** 这次更新主要修了几个老毛病，顺手加了两个常被要求的功能。

## Quality Grading

After humanization, grade the output by residual S1/S2 counts and improvement percentage. Improvement % = proportion of detected tells removed versus the original.

| Grade | Residual S1 | Residual S2 | Improvement | Verdict |
|-------|-------------|-------------|-------------|---------|
| **A** | 0 | ≤ 2 | ≥ 70% | Ship — reads human |
| **B** | 0 | ≤ 4 | ≥ 50% | Acceptable — minor residue |
| **C** | 1–2 | any | < 50% **or** over-edit signal | Second pass required |
| **D** | ≥ 3 | any | severe over-edit (> 50% changed) | Human review required |

**Hard rules:**
- Any residual S1 caps the grade at **C**.
- 3+ residual S1 forces **D**.
- S3 (CN-J) affects the grade only when it overlaps an S1/S2 cluster.

## Over-Editing Guardrails

Humanization removes AI tells; it must never become a rewrite that drifts from the author's meaning or intent.

**Change-ratio gates** (measured by character-level edit distance against the original):
- **> 30% changed → WARN.** Flag for review; likely drifting from author intent. Continue only if every edit maps to a detected tell.
- **> 50% changed → HALT.** Forced termination. Editing has become rewriting; revert to the last safe state and request a narrower scope or a human decision.

**Meaning-preservation rules (all must hold):**
1. **No fact mutation** — never add, drop, or alter named entities, numbers, dates, claims, or citations. The sole exception is *removing* fabricated CN-K citations; never invent replacements.
2. **No stance shift** — the author's position, hedging level, and conclusion stay intact.
3. **Tell-anchored edits only** — every change must trace to a catalogued tell (CN-A…CN-K). Stylistic "improvements" unconnected to a tell are forbidden and are themselves an over-edit signal.
4. **Length parity** — keep output length within ±20% of the source unless a specific tell (CN-B template bloat, CN-H filler) requires cutting.
5. **Register lock** — humanize *within* the source register. Do not push formal text into slang or vice versa; academic text stays academic, just de-AI'd.

**Mode mapping:**
- **Fast mode** (default, short text): single pass — detect → rewrite → self-verify; prioritize S1 removal.
- **Strict mode** (long or high-stakes text): separate stages — detection → rewrite → fidelity audit (meaning-preservation rules 1–5) → naturalness review; produce an itemized per-tell change log.

---

## Copy Layer (Chinese)

Genre scope: 营销文案 — headlines, slogans, CTAs, 落地页 (landing pages), brand/founder storytelling, and slide titles. The copy layer continues the `CN-` letter scheme after CN-K (CN-L…CN-Q). Copy inverts the prose clustering logic: a headline or slogan is often a single line with no room to cluster, so several thresholds shift toward single-occurrence decisiveness at the headline slot. In copy mode the over-editing guard changes: the change-rate limit is replaced by the **fact-anchor preservation guard** (see the shared SKILL.md guardrails) — numbers, dates, prices, and proper nouns stay intact and the core promise/benefit keeps its meaning, while expression and sentence structure may be rewritten freely.

A critical calibration governs this layer: 对偶 and 排比 are prized classical devices of Chinese rhetoric — the boundary is **content-first versus template-first, never occurrence count**. See the boundary analysis below; this is the highest-false-positive area of the whole catalogue.

### Detection Categories (copy)

| ID | Tell | Why it reads as AI | Severity |
|----|------|--------------------|----------|
| CN-L | 否定式煽情标题 — negation-contrast headline: 这不是……而是……／不仅是……更是…… as the whole message of a headline slot | In prose the frame can read as natural Chinese (prose CN-J is S3), but in a headline it IS the message — the genre lifts severity a step because the contrast is manufactured insight at the highest-visibility position | S2 (genre-elevated from prose CN-J) |
| CN-M | 破折号对比标题 — the "X — Y" dash-contrast headline shape carrying a binary antithesis | Chinese 破折号 (the full-width double-em ——) conventionally marks explanation or topic shift, never binary contrast; the contrast use is an English import, so the shape reads doubly generated. Scoped to the binary-contrast headline shape only — never to every dash | S2 |
| CN-N | 落地页万能公式标题 — slot-fill landing templates: 专为X打造的Y／开启X之旅／解锁全新Y | The universal formula accepts any product noun; interchangeable slots with no concrete claim are the generated-landing signature | S2 |
| CN-O | 强行升华·口号式行动结尾 — forced elevation to grand significance: 让我们携手共创美好未来 | The unprompted uplifting moral summary at a copy ending mirrors prose CN-D; a single occurrence at a slogan/closing slot is decisive | S1 (matches prose CN-D) |
| CN-P | 品牌/创业故事模板骨架 — the brand-story skeleton: 黄金圈 Why-How-What + 三幕式 + 反差构建 + origin myth | Any one beat can be legitimate craft; the co-occurring skeleton (2+ markers) is the template fingerprint | S2 (gate on ≥2 skeleton markers) |
| CN-Q | 虚假具体性·可疑叙事数据 — forced sensory 多模态锚点 plus precise-but-unsourced business numbers | Precision theater simulates authenticity; unverifiable exact figures and stacked sensory detail are generated-narrative markers | S2 (escalates toward S1 when a suspect number is verifiable and material; never replace a fabricated stat with an invented one — delete or downgrade) |

### Severity Rationale (copy)

- CN-L is S2, not S3: the genre lifts it one step from prose CN-J, because a headline slot has no surrounding text to dilute the manufactured contrast.
- CN-O is S1: it inherits the prose CN-D forced-elevation verdict — decisive on a single unprompted uplifting slogan.
- CN-M, CN-N, CN-P, CN-Q are S2. CN-P convicts only on the co-occurrence of two or more skeleton markers. CN-Q escalates toward S1 when the suspect number is both verifiable and material to the claim; like CN-K, the fix is deletion or downgrade, never substitution with an invented figure.
- **Clustering at the section level**: a generated landing section typically stacks CN-N template headline + CN-L contrast subhead + CN-O uplifting close. Resolve the S1 first, then re-score.

### 对偶/排比 Boundary Analysis (critical)

Parallelism is prized, classical, natively Chinese craft — 排比 (three or more clauses of similar structure, legitimately reusing words for momentum) and 对偶 (the stricter two-clause couplet of the 对联 tradition). A blanket count rule over-fires badly on copy. The operational boundary is **content-first versus template-first**:

1. **Content-first (do NOT flag)**: each parallel member carries a distinct concrete fact. The classic crafted 排比 — 感谢冰峰，感谢风暴，感谢悬崖，感谢缺氧。— concentrates real referents into rhythm; this is skilled human copy.
2. **Template-first (flag)**: the symmetry is assembled BEFORE the content (先套模板再填内容) — interchangeable filler orbits one abstract idea, and the balanced form dilutes 信息密度 (information density) instead of concentrating it (用对称句式稀释信息).
3. **Novelty**: crafted parallelism is 不落窠臼 (unhackneyed); generated parallelism is predictable.
4. **Count is a weak signal on copy** — a single crafted 对偶 couplet or one three-part 排比 slogan is expected native craft; occurrence count alone never decides. Stacked blocks of balanced rhetoric matter only when the content test above already indicates template-first assembly.

### Dash-Contrast Applicability (narrow transfer)

- The em-dash-overuse tell transfers to Chinese copy — practitioners treat dash abuse as a generated-text fingerprint — but the **binary-contrast headline shape is the actual tell**, not the dash character.
- The Chinese 破折号 is the full-width double-em ——, and its sanctioned functions are 解释说明 (explanation), 话题转换 (topic shift), 声音延长, and 列举 — none of which is English's binary antithesis. An "X —— Y" contrast headline therefore reads doubly imported.
- **False-positive guard**: a legitimate —— used for 解释说明 ("大会堂的枢纽——中央大厅") or 话题转换 must NOT be flagged. CN-M stays scoped to the binary-contrast headline shape.

### Before/After Rewrite Examples (copy)

- **AI:** 这不仅仅是一双跑鞋，而是对自律生活方式的承诺。
  **人:** 这双跑鞋陪你跑完第一个五公里。
  (CN-L: replace the manufactured contrast with the concrete promise it hid)
- **AI:** 极速体验 —— 全新升级
  **人:** 打开页面只要0.8秒，比上一版快了一半。
  (CN-M: dissolve the dash-contrast headline into a verifiable claim)
- **AI:** 专为现代团队打造的一站式协作平台，开启高效办公之旅。
  **人:** 审批、评论、进度，一个页面全看到——不用再翻三个群。
  (CN-N: swap the slot-fill formula for the actual differentiator; the —— here is legitimate 解释说明)
- **AI:** 让我们携手共创美好未来。
  **人:** 下个月的账单，先从这里省起。
  (CN-O: cut the forced elevation; close on the concrete next step)
- **AI:** 我们从一间车库起步，怀揣改变世界的梦想，历经三幕般的起伏，终于迎来了属于我们的时代。
  **人:** 第一批货是在城中村的出租屋里打包的，胶带用完了三卷。
  (CN-P: replace the story skeleton with one verifiable lived detail)
- **AI:** 经过对2042条用户评价的深度分析，98.7%的用户感受到了指尖传来的温润触感。
  **人:** 上个月的300条评价里，提到"手感"的有41条。
  (CN-Q: replace precision theater with a checkable count)

### High-False-Positive Signals (copy — do NOT flag standalone)

- A single crafted 对偶 couplet or one three-part 排比 slogan — the 对联 tradition is living, prized craft; judge by the content-first test, not by count.
- 破折号 —— used for 解释说明, 话题转换, or 声音延长 — sanctioned native functions.
- Verb-first urgency CTA copy (立即购买／免费试用) — standard good practice, not a generated tell.
- Precise numbers that ARE verifiable (a real survey n, real dates, a citable source).
- Informational copy (FAQ, specs, prices, business-registration blocks) — legitimately third-person, no elevation expected.
- Platform mis-flagging is rampant for Chinese; require a cluster before asserting machine authorship.

---

## Attribution

The category-catalogue structure of this module is inspired by the **im-not-ai (Humanize KR)** project (https://github.com/epoko77-ai/im-not-ai). The Chinese tell patterns here — prose and copy layers alike — were independently researched and authored.
