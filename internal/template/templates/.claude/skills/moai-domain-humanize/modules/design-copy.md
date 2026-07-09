# Display Copy — Genre Rules for Landing Pages and Short-Form Cards/Slides

Genre rules for **display-surface copy**: landing pages, slide and card decks, and design-tool result copy. This module is loaded IN ADDITION to the matching language module — the language module owns the tell catalogue; this module owns genre structure. All severity judgments use the shared S1/S2/S3 model, and every rewrite obeys the fact-anchor preservation guard (both defined in `SKILL.md`).

---

## Landing-Page Rules

### Structure Constraints (DCG-1…DCG-4)

| ID | Element | Rule |
|----|---------|------|
| DCG-1 | Headline | One core benefit only. Say what the reader **gets**, not how the product works. Length budget per the language adaptation blocks below. |
| DCG-2 | Subheadline | Complements the headline with the *how* or the differentiator — never restates it. Length budget per language. |
| DCG-3 | Body | Short sentences, short paragraphs (limits per language). Every feature line converts to a benefit: feature "automatic classification" → benefit "saves [number] hours a week". |
| DCG-4 | CTA | Verb-first label, short (limit per language). Reassurance microcopy sits under the button ("[number]-second signup, no card required" — real facts only, never invented). |

### Vague-Claim → Concrete-Mechanism Repair Table (DCG-5)

Every vague claim repairs to a named mechanism, a metric, or a piece of evidence. Anti-pattern classes are language-neutral; the formula-level instances per language live in the `modules/copy-review.md` dictionaries.

| Anti-pattern class | Repair direction |
|--------------------|------------------|
| "revolutionary solution" self-praise | Name the mechanism plus its measured effect: "automatic classification cuts manual sorting by [number]%" |
| "cutting-edge technology" tech boast | Name the technology plus its user-visible effect: "[named model/engine]-based real-time translation" |
| "fast" with no number | Quantify: "[number]× faster than the previous step", "processes [number] items per minute" |
| "everyone loves it" satisfaction claim | Cite a checkable figure: "[score]/5.0 average across [number] reviews" |
| "your life changes today!" urgency | Evidence-based promise: "decide after the [number]-day free trial" |
| "industry's best" superlative | Name the comparison basis: "ranked first in its category on [named review platform]" |

### Required Copy Element Checklist (DCG-6)

- [ ] Concrete numbers appear (%, multiples, prices, or time saved)
- [ ] Testimonials carry a real name + role — if no real testimonial exists, omit the section entirely (fabricating one violates the fact-anchor guard)
- [ ] The FAQ pre-empts the main purchase objections
- [ ] A guarantee/safety line sits near the CTA ("free trial", "no card required" — only when true)
- [ ] The headline delivers its value inside a 10-second read

### Tone Profile (DCG-7)

When reviewing or authoring landing copy, emit a tone profile so the copy set can be judged as one voice:

```
Tone profile:
- Primary emotion target: <e.g. trust, competence>
- Secondary emotion: <e.g. curiosity, relief>
- Intended reader takeaway: "<one sentence the reader should think>"
- Average sentence length: <in the language's native measure>
- Avoid: exclamation stacking, urgency imperatives, self-praise adjectives
```

---

## Short-Form Card/Slide Rules (DCG-8…DCG-12)

| ID | Rule |
|----|------|
| DCG-8 | **Cover economy** — the cover carries one hook within the language's native length budget (see adaptation blocks). A cover that needs two readings has failed. |
| DCG-9 | **One thought per line** — body slides put exactly one idea on each line; the line break IS the reading rhythm. Do not pack two claims into one line. |
| DCG-10 | **Ending slide** — no formulaic reader-address closer (the "was this helpful?" class). End on a checklist, a one-line summary, or one resonant sentence. |
| DCG-11 | **Numeral style** — data points use numerals, not spelled-out numbers (per-language conventions in the blocks below). Numbers get the visual emphasis (bold or large type), not extra adjectives. |
| DCG-12 | **Exclamation restraint** — at most one exclamation mark across a card/slide set. The punch comes from the fact, not the punctuation. |

---

## Per-Language Adaptations

Language-dependent parameters — length limits, ending-form conventions, script and numeral conventions — are stated per language **in that language's own native measure**. A limit defined for one language MUST NOT be applied verbatim to another: a Korean character budget says nothing about an English word budget, and vice versa.

### Korean

- **Cover economy (DCG-8)**: measured in characters (자). Around 10자 is the ideal hook length; keep covers under 20자.
- **Headline (DCG-1)**: one benefit; avoid particle/noun-fragment titles and dash-contrast shapes — both are catalogued in `modules/korean.md` (M-2, M-1).
- **Ending variation**: never let ~합니다/~됩니다 run three or more in a row; mix 단정(~한다), 설명(~이다), 질문(~일까?), 한정(~뿐이다). ~하세요/~해보세요 stacking (3+) reads generated.
- **Slot formulas**: the Korean headline formulas (혁신적인/차세대 class) are catalogued in the `modules/copy-review.md` Korean dictionary; the habit-word families live in the `modules/korean.md` prose categories.
- **Numerals (DCG-11)**: Arabic numerals for data — "3가지", not "세 가지".
- **Exclamation (DCG-12)**: at most one per card set; prefer 마침표 or 물음표.
- **Ending slide (DCG-10)**: the "도움이 되셨나요?" class closer is the formulaic reader-address to avoid.

### English

- **Cover economy (DCG-8)**: measured in words. Around 5-7 words is the ideal hook; keep covers under 12 words.
- **Headline (DCG-1)**: 10 words or fewer, one benefit, outcome-first. Terse or verbless headlines are native English craft — do not import another language's fragment rules (see the partial-transfer notes in `modules/english.md`).
- **Subheadline (DCG-2)**: 20 words or fewer. **Body (DCG-3)**: sentences of 20 words or fewer; paragraphs of 3 sentences or fewer. **CTA (DCG-4)**: verb-first, 7 words or fewer.
- **Ending variation**: English has no sentence-ending morphology axis — the native lever is sentence-length variation and avoiding a uniform declarative rhythm (rhythm tells live in the `modules/english.md` prose layer).
- **Numerals (DCG-11)**: numerals for data points and statistics.
- **Exclamation (DCG-12)**: at most one per set; hype CTA labels are catalogued at english.md ENC-7.

### Japanese

- **Cover economy (DCG-8)**: measured in characters (文字). Around 13-15文字 is the classic short-headline economy; judge by the read-aloud rhythm (声に出して読む) rather than a hard cap.
- **Headline (DCG-1)**: one benefit. 体言止め (the noun-ending close) is legitimate, prestigious craft — gate it by frequency per japanese.md JA-10, never by mere presence.
- **Ending variation**: vary sentence endings; monotone です・ます runs are catalogued at JA-04 in `modules/japanese.md`.
- **Punctuation**: no English-style heading colon (JA-11); no imported em-dash where native punctuation serves (JA-12).
- **Numerals (DCG-11)**: half-width Arabic numerals (半角数字) for data.
- **Exclamation (DCG-12)**: at most one per set.

### Chinese

- **Cover economy (DCG-8)**: measured in characters (字). Chinese packs roughly one morpheme per character, so covers run tighter — around 8-16字, judged by information density per character.
- **Headline (DCG-1)**: one benefit. Parallel couplets (对偶/排比) are prized native craft — judge content-first versus template-first per the boundary analysis in `modules/chinese.md`, never by occurrence count.
- **Dash use**: the full-width 破折号 (——) natively marks explanation or topic shift; the binary-contrast headline shape is catalogued at chinese.md CN-M.
- **Ending slide (DCG-10)**: forced-elevation closers (让我们……共创 class) are catalogued at CN-O — end on a concrete next step instead.
- **Numerals (DCG-11)**: Arabic numerals for data and statistics; avoid spelling data numbers in 汉字数字.
- **Exclamation (DCG-12)**: at most one per set (！).

---

## Integration

- **Review path**: when the copy-review pipeline (`modules/copy-review.md`) processes display-surface copy, these genre rules join Stage 3 pattern matching as an additional checklist.
- **Rewrite path**: in the skill's default rewrite mode, apply genre rules under copy mode — the fact-anchor preservation guard in `SKILL.md` governs every rewrite.
- **Boundary**: the language modules stay authoritative for tell catalogues; this module owns genre structure only and defines no severity or grade tables of its own.

---

## Sources

- The landing-page structure rules, the repair table, and the element checklist are adapted from the maintainer's own landing-copy rulebook (same authorship as this skill; direct reuse).
- The Korean short-form card/slide rules are adapted from the maintainer's own Korean card-copy style guide, which draws on published research on detecting LLM-generated Korean text and on public AI-writing style guides.
- The English, Japanese, and Chinese adaptation blocks restate each language-dependent parameter in that language's own native measure (words for English; characters for Japanese and Chinese, each with its own budget and rationale). No Korean numeric limit is transferred verbatim; where a genre tell is already catalogued in a language module, the module ID is cross-referenced instead of re-defined.
