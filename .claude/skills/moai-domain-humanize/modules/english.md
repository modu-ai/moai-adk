# Humanize — English

This module catalogues English-specific "AI tells" — the patterns that make text read as machine-generated — and the rules for removing them without distorting meaning. Use it to detect, grade, and rewrite English prose toward natural, human-authored style.

## Detection Categories

Ten categories of English AI tells. IDs are stable handles for downstream tooling.

| Category | Tell | Why it reads as AI | Severity |
|----------|------|--------------------|----------|
| **EN-A** Focal / Excess Vocabulary | Style-word clusters that spiked after late 2022 — verbs (*delve, underscore, leverage, utilize, foster, harness, streamline, facilitate, navigate, elevate, bolster, showcase, unlock, embark, illuminate*), adjectives (*crucial, pivotal, meticulous, robust, seamless, comprehensive, multifaceted, holistic, cutting-edge, transformative, groundbreaking, notable, commendable*), nouns/metaphors (*tapestry, landscape, realm, mosaic, symphony, beacon, cornerstone, testament, ecosystem*). | Corpus studies show these are measurable LLM "excess vocabulary" — words a human rarely reaches for at this density. *Tapestry* and *delve* are near-stigmatized single-word giveaways. | **S2** (*delve*, *tapestry* are **S1** singletons) |
| **EN-B** Negative Parallelism (Contrastive Reframe) | "It's not just X, it's Y" / "not only X but also Y" / "This isn't a product; it's a revolution" / clipped tail negations ("no guessing", "no fluff"). | The strongest contemporary tell. It stages drama before stating the point and manufactures the illusion of insight while adding little. | **S1** |
| **EN-C** Rule of Three / Mechanical Tricolon | Three parallel items of near-equal length, often identically punctuated: "Fast. Simple. Effective." / "No fluff. No filler. No stress." | Models flatten the rhetorical tricolon into an algorithmic default rhythm. The symmetry is too regular to be organic and recurs fractally. | **S2** |
| **EN-D** Inflated Significance / Editorializing Puffery | "stands as a testament to", "plays a vital/pivotal role", "marks a pivotal moment", "underscores the importance of"; promotional flavor: "boasts", "vibrant", "nestled", "rich cultural heritage", "breathtaking". | Models reflexively inflate the importance of whatever they describe and editorialize value judgments instead of reporting facts neutrally. | **S2** |
| **EN-E** Signposting & Hedging Fillers | "It's important to note that", "When it comes to", "In the realm of", "One must consider", "It is worth mentioning", "In light of this". | Low-information framing scaffolding. Humans say "Note that" or just state the thing; these phrases announce content without delivering it. | **S2** |
| **EN-F** Transition / Conjunction Overuse | Sentence- or paragraph-opening "Moreover," "Furthermore," "Additionally," "In addition," "Therefore," clustered mechanically. | People rarely chain formal connectives this densely. The uniform linking signals template-driven generation rather than organic argument flow. | **S2** |
| **EN-G** Cliché Openers & Closing Rituals | Openers: "In today's fast-paced world", "In today's competitive landscape", "In an era of". Closers: "In conclusion,", "In summary,", "Ultimately,", "By [doing X], your [Y] will [Z]". | Default framing devices and recap rituals that appear across countless AI outputs regardless of topic. | **S1** |
| **EN-H** Vague Attribution / Weasel Words + Participial Tails | "Experts argue", "Industry reports suggest", "Observers note", "Some critics", "has been described as" — no named source. Plus tacked-on participial clauses: "...transforming the industry", "...highlighting the need for further research". | Models fabricate authority and append "-ing" clauses to simulate analytical depth, often introducing unsupported claims. | **S2** |
| **EN-I** Formatting Tells | Excessive **bold** on every key term; Title Case In Headings; H2/H3 headings in short pieces; bullet lists for non-list ideas (bolded stem + colon + reworded elaboration); emoji as bullets/decoration; thematic breaks before headings; curly "smart" quotes. | Default chat-UI formatting habits. Research found ~70% of analyzed ChatGPT messages contained an emoji. Emphasis is mechanical rather than meaning-driven. | **S3** |
| **EN-J** Rhythm Uniformity & Strategic Vagueness | Low sentence-length variance ("too clean"); zero contractions + perfect Oxford-comma grammar; absence of concrete details, proper nouns, dates, or anecdote; universal platitudes ("Change is the only constant"); aphoristic pull-quote endings; sycophantic sprinkles ("Great question!", "Hope this helps!"). | Statistically low perplexity — predictable, even, inoffensive. The absence of specificity and the metronomic rhythm read as "machine-shaped." | **S3** |

> **Punctuation note (em-dash):** Em-dash overuse (especially tight `—like this—`) is a widely discussed tell, but it is a high-false-positive signal — many human writers use it well. Treat it as a **modifier** within EN-I/EN-J rather than a standalone removable category: flag it only when it co-occurs with other tells.

## Severity Rules

- **S1 (critical):** A single occurrence strongly confirms AI authorship. Remove on first occurrence.
- **S2 (strong):** Acceptable at 1–2 instances. Remove at 3+ in the same passage.
- **S3 (weak):** Problematic only when overlapping other tells. Acts as a downgrade signal, not a standalone removal target.

Application notes:
- EN-A is governed by *density*: two or more focal words in a short passage trigger removal, but the stigmatized singletons (*delve*, *tapestry*) are removed on first sight.
- S3 tells (EN-I, EN-J) do not by themselves warrant a rewrite. When 3+ S3 signals overlap and reinforce an S1/S2 finding, they pull the grade down one band.
- When two categories co-occur in one sentence (e.g., EN-B negative parallelism built from EN-A focal words), resolve the S1 first, then re-check whether the S2 still stands after the rewrite.

## Rewrite Examples

> The before/after pairs below are in **English** because they demonstrate English-specific AI tells.

### EN-A — Focal / Excess Vocabulary
- **Before:** "This study delves into the intricate tapestry of cellular mechanisms."
  **After:** "This study examines how the cellular mechanisms interact."
- **Before:** "We leveraged a robust framework to facilitate seamless integration."
  **After:** "We used a reliable framework to connect the two systems."
- **Before:** "The findings underscore the transformative potential of the approach."
  **After:** "The findings show the approach can change how teams work."

### EN-B — Negative Parallelism
- **Before:** "It's not just about efficiency — it's about transformation."
  **After:** "The change saves time and reshapes how the team works."
- **Before:** "This isn't a dashboard. It's a command center."
  **After:** "The dashboard puts approvals, comments, and status in one place."
- **Before:** "The options come from the selected item, no guessing."
  **After:** "The options come from the selected item, so users don't have to guess."

### EN-C — Rule of Three / Tricolon
- **Before:** "Our tool is fast, simple, and effective."
  **After:** "Our tool runs queries in under a second and needs no setup."
- **Before:** "No fluff. No filler. No stress."
  **After:** "It gets to the point and is easy to use."

### EN-D — Inflated Significance / Puffery
- **Before:** "This release stands as a testament to our unwavering commitment to innovation."
  **After:** "This release adds offline sync and fixes the export bug."
- **Before:** "Nestled in the vibrant heart of the region, the town boasts a rich cultural heritage."
  **After:** "The town sits in the river valley and has three 18th-century churches."

### EN-E — Signposting / Hedging Fillers
- **Before:** "It's important to note that, when it comes to security, one must consider access control."
  **After:** "For security, control who can access the data."
- **Before:** "In the realm of testing, it is worth mentioning that coverage matters."
  **After:** "Test coverage matters."

### EN-F — Transition Overuse
- **Before:** "Moreover, the API is fast. Furthermore, it scales well. Additionally, it is secure."
  **After:** "The API is fast, scales well, and is secure."

### EN-G — Cliché Openers / Closing Rituals
- **Before:** "In today's fast-paced digital landscape, businesses must adapt."
  **After:** "Most teams now ship updates weekly, so processes have to keep up."
- **Before:** "In conclusion, by following these steps, your project will succeed."
  **After:** "Run these four checks before each release and you'll catch most regressions."

### EN-H — Vague Attribution / Participial Tails
- **Before:** "Experts argue this is the best approach, highlighting its many benefits."
  **After:** "The 2025 PostgreSQL benchmark found this approach 30% faster on writes."
- **Before:** "The feature shipped early, underscoring the team's dedication."
  **After:** "The feature shipped two weeks early."

## Quality Grading

Grade **after** revision, on residual tell counts plus percent improvement (the proportion of detected tells removed without introducing new ones).

| Grade | Criteria | Action |
|-------|----------|--------|
| **A** | 0 residual S1 **and** ≤2 residual S2 **and** ≥70% improvement. Reads as human-authored. | Ship. |
| **B** | 0 residual S1 **and** ≤4 residual S2 **and** ≥50% improvement. Acceptable; minor polish remains. | Ship or optional polish pass. |
| **C** | 1–2 residual S1 **or** 50–69% improvement **or** over-edit warning triggered (>30% changed). | Trigger a second pass. |
| **D** | 3+ residual S1 **or** <50% improvement **or** over-edit halt triggered (>50% changed) **or** meaning drift detected. | Request human review; do not auto-ship. |

S3 tells do not affect the grade directly. They downgrade by one band only when 3+ overlap and reinforce an S1/S2 finding.

## Over-Editing Guardrails

Humanizing refines *how* something is said, never *what* it says. These guardrails keep an edit from sliding into a rewrite.

**Change-budget gates** (token/word-level diff against the original):
- **>30% changed → WARN.** Surface a caution; the rewrite is drifting from "edit" toward "regenerate." Cap at Grade C until justified.
- **>50% changed → HALT.** Stop and require human confirmation. Above this you are rewriting the argument, not humanizing it.

**Meaning-preservation rules** (mandatory):
1. **Anchor facts first.** Before editing, fix the claims, numbers, names, dates, and certainty level that must not change.
2. **Edit at sentence/phrase level**, not whole-document regeneration — this preserves meaning, terminology, and structure.
3. **Add no new facts.** Never invent specifics to replace vagueness; if real specifics aren't available, simplify rather than fabricate.
4. **Drop no load-bearing facts.** Removing an inflated wrapper must not remove the substantive claim inside it.
5. **Preserve genuine certainty/hedging.** Don't turn a calibrated "may" into a false absolute.
6. **Preserve technical terminology** in technical material — don't "humanize" a precise API name into prose.
7. **Editor-pass diff check.** After rewriting, compare facts, tone, certainty, and examples against the original; if anything drifted, revert that edit.

**Operating modes:**
- **Fast Mode** (default, short text): single pass — detect → rewrite → self-verify against the meaning-preservation checklist.
- **Strict Mode** (long or high-stakes text, or `--strict`): staged — detect → surgical rewrite → fact-fidelity audit → naturalness review — with a second pass whenever the result lands at Grade C.

---

## Source & License

This module's structure follows the **im-not-ai (Humanize KR)** open-source skill (https://github.com/epoko77-ai/im-not-ai, MIT License — Copyright (c) 2026 epoko77-ai). The English tell patterns here were independently web-researched, not ported from the source. See NOTICE.md.
