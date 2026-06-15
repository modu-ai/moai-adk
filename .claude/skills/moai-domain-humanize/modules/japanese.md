# Humanize — Japanese (日本語)

This module catalogues AI-generated-text tells specific to Japanese (日本語) and the rules for humanizing (윤문 / post-editing) such text into natural, native prose. It is tuned for output from ChatGPT, Gemini, and Claude.

> Note on detectors: English-trained AI-detection tools are unreliable on Japanese and frequently produce false positives on ordinary 敬語 (polite forms). Prefer human-and-heuristic inspection over automated detector scores when setting confidence. The "read aloud" (声に出して読む) test is the single most reliable rhythm check for Japanese.

## Detection Categories

| Category | Tell | Why it reads as AI | Severity |
|----------|------|--------------------|----------|
| JA-01 冗長定型助動詞 (Redundant auxiliary constructions) | Padded polite endings: 「〜することができます」「〜を実現します」「〜の最適化」「〜することが望ましいでしょう」 | LLMs pick the longest "safe / 無難" form; a human writes 「利用できます」, not 「利用することができます」. The padding is statistically over-represented. | S2 |
| JA-02 ヘッジ・断定回避 (Hedging / assertion-avoidance) | 「〜と考えられます」「〜と思われます」「〜の可能性があります」「〜と言われています」「一般的に〜とされています」 | The model never holds full confidence and lacks a source, so it hedges constantly. Humans either assert or cite; AI does neither. | S2 |
| JA-03 構造宣言・定型接続 (Structural declarations & formulaic connectives) | Paragraph-initial 「まず」「次に」「さらに」「また」「したがって」; structure announcements 「以下で解説します」「結論から言うと」「大きく分けると」「論点は3つあります」; 「これにより」 | AI defaults to a textbook "declare structure → enumerate → caveat → encourage" scaffold; connectives cluster at sentence heads with mechanical regularity. 「これにより」 is a notorious giveaway. | S1 |
| JA-04 文末単調 (Monotone sentence endings) | Long runs of identical 「〜です。〜ます。」 with no variation — no 体言止め, no 「〜でしょう」, no 「〜かもしれません」 | Human prose varies cadence; AI keeps a flat, uniform ending rhythm that reads as 稚拙 (childish) yet over-polished. | S2 |
| JA-05 リズム均一・文長一定 (Uniform rhythm & sentence length) | All sentences roughly equal length; no 緩急 (short/long alternation); every paragraph closes "cleanly"; 「〜ではなく〜」 repeated | Statistical generation produces 平板 (flat), evenly-metered text. Humans write in bursts and pauses. | S3 |
| JA-06 抽象・具体欠如 (Abstraction / lack of concrete detail) | Generic filler 「多くの企業で導入が推進されています」「継続することが成功の鍵です」; no place names, proper nouns, numbers, dates, or named cases | The model produces broadly-applicable, surface-level content with no lived specifics ("既製品感" / "マニュアル感"). | S2 |
| JA-07 因果薄弱・同義反復 (Weak causality & tautological repetition) | Sentences end on 「〜が重要です」「〜が大切です」「〜がポイントです」 with no reason; restating one idea with slight rephrase (「柔軟な対応が必要です。適応力を高めることが重要です。」) | AI asserts importance without grounding and pads length by paraphrasing itself. | S1 (tautology) / S2 (bare 重要です) |
| JA-08 記号・整形クセ (Symbol & formatting artifacts) | Emoji at list heads (✅ 💡 🚀 ✨); Markdown bold (**keyword**), 「#」見出し記号; em-dash 「—」 in place of Japanese 「…」; over-frequent 1. 2. 3. lists & 箇条書き／チェックリスト; parenthetical responsibility-dodging 「（〜の場合）」 | Direct chat-interface / Markdown leakage. A bulleted prose article reads as a 生成物 (machine product) at a glance — described in sources as 致命傷 (fatal) for readable content. | S1 (emoji / em-dash / Markdown in prose) / S2 (excess bullets) |
| JA-09 カタカナ語多用・翻訳調 (Katakana overload & translationese) | Excess loanwords where 和語／漢語 fit; literal-translation phrasing (the "thought in English first, then translated" feel); mishandled idioms read literally; 「本質的な課題」「〜の最適化」 | Models are English-data-heavy and pivot internally through English, producing 直訳調 (literal-translation tone) and importing katakana a native writer would localize. | S3 (katakana) / S1 (broken idiom) |

## Severity Rules

Severity controls when a flagged span must be removed or rewritten.

- **S1 — critical. Remove on a single occurrence.**
  - JA-03: 「結論から言うと」「以下で解説します」「これにより」 → remove on any single occurrence; paragraph-head 「まず／次に／さらに」 chains → trigger when 3+ run consecutively.
  - JA-07 (tautology): same-idea restatement → remove on single occurrence.
  - JA-08: emoji, 「—」, **bold**, 「#」 appearing in body prose → remove on sight.
  - JA-09 (broken idiom): an idiom read or rendered literally → fix on single occurrence.

- **S2 — strong. Allow 1–2; remove starting at the 3rd occurrence.**
  - JA-01: allow 1–2 padded constructions; remove at 3+.
  - JA-02: allow 1–2 hedges; remove at 3+.
  - JA-04: trigger when ≥3 identical endings run consecutively.
  - JA-06: allow some generality; flag a passage that has zero concrete anchors.
  - JA-07 (bare 重要です): unsupported 「重要です／ポイントです」 → remove at 3+.
  - JA-08 (excess bullets): bullet overuse in what should be prose → reduce at 3+.

- **S3 — weak. Only problematic when overlapping with another category.**
  - JA-05: flag only when it overlaps JA-04 (monotone endings) or JA-06 (abstraction).
  - JA-09 (katakana overload): flag only when it overlaps another category.

## Rewrite Examples

### JA-01 冗長定型助動詞
- Before: このツールを利用することができます。
  After: このツールが使えます。
- Before: 業務効率の最適化を実現します。
  After: 業務のムダを減らせます。

### JA-02 ヘッジ・断定回避
- Before: この方法が有効であると考えられます。
  After: この方法は効きます。実際、導入後に問い合わせ対応が半分になりました。
- Before: 一般的に重要だとされています。
  After: 現場では、ここを外すと必ず後で揉めます。

### JA-03 構造宣言・定型接続
- Before: 結論から言うと、以下で3つのポイントを解説します。まず…次に…さらに…
  After: ポイントは3つ。一番効くのは、最初の設定をサボらないことです。…あとは…そして最後に…
- Before: これにより、生産性が向上します。
  After: おかげで、残業がぐっと減りました。

### JA-04 文末単調
- Before: 観光地は混雑しています。ホテルは満室です。飲食店も混雑しています。
  After: 観光地はどこも人だらけ。朝からホテルのロビーには長い列ができ、ランチを過ぎても飲食店は満席が続いていました。

### JA-06 抽象・具体欠如
- Before: 多くの企業で導入が推進されています。
  After: 観光庁は2024年度、西みやこ市内のホテルを対象に、多言語AIチャット導入の補助金を5億円規模で交付しました。
- Before: 継続することが成功の鍵です。
  After: 半年やめずに続けた人だけが、3か月目で数字が動き始めました。

### JA-07 因果薄弱・同義反復
- Before: 観光地の安全対策が重要です。観光客が増えているので、事故予防が必要です。
  After: 外国人観光客は前年比35%増。東山寺周辺では転倒事故が前年の1.5倍に増えています。
- Before: 柔軟な対応が必要です。そのためには適応力を高めることが重要です。
  After: ピーク期は臨時スタッフを入れ、夜間対応は例外的に省く——そう事前に決めておくだけで現場は回ります。

### JA-08 記号・整形クセ
- Before: ✅ **重要ポイント** — まず計画を立てましょう
  After: 大事なのは、まず計画を立てること。
- Before: ステップは以下の通りです：1. 準備 2. 実行 3. 確認
  After: やることは、準備して、実行して、最後に見直す。それだけです。

### JA-09 カタカナ語多用・翻訳調
- Before: 本質的な課題に対するソリューションをコミットします。
  After: 根っこの問題から片付けます。
- Before: 財布が悲鳴をあげているという音が聞こえます。 *(idiom read literally)*
  After: 今月は出費がかさんで、財布が悲鳴をあげています。

## Quality Grading

Apply after the rewrite. **Residual** = count of S1/S2 tells remaining in the output. **Improvement** = (original tell count − residual tell count) / original tell count.

| Grade | Residual S1 | Residual S2 | Improvement | Action |
|-------|-------------|-------------|-------------|--------|
| **A** | 0 | ≤ 2 | ≥ 70% | Accept |
| **B** | 0 | ≤ 4 | ≥ 50% | Accept |
| **C** | 1–2 (or ≥2 over-edit signals) | — | < 50% | Round-2 rewrite |
| **D** | ≥ 3 (or severe over-edit / meaning drift) | — | — | Human review required |

Hard rules:
- Any residual S1 caps the grade at **C** regardless of improvement %.
- Any meaning-distortion flag forces **D**.

## Over-Editing Guardrails

Humanization is post-editing (윤문), not rewriting. The goal is to remove AI tells from the flagged spans while preserving meaning — not to produce a new draft.

**Change-magnitude gates** (measured as character-level edit distance ÷ original length):
- **> 30% changed → WARN.** Surface a diff and confirm the edit is humanization, not a rewrite. Padding-removal (JA-01) legitimately shrinks text, so a length drop alone is not a violation — flag when *meaning-bearing* spans are altered.
- **> 50% changed → HALT.** Stop and require human review. At this point the output is a new draft, not a post-edit.

**Meaning-preservation rules (non-negotiable):**
1. **No fact invention.** Humanizing must not add place names, numbers, dates, or cases that were not in the source. JA-06 wants *real* specifics — if none exist, leave a placeholder or flag for the author; never fabricate. This guards against hallucination during the de-AI pass.
2. **Preserve claims & polarity.** Do not flip assertions, negations, or hedged-vs-asserted stance beyond removing redundant hedges.
3. **Preserve named entities, quotes, figures, and technical terms verbatim.**
4. **Style conversion is allowed but must be consistent.** Keep です・ます調 ⇔ だ・である調 unified across the whole text; never mix mid-document.
5. **Don't over-correct into a new AI tell.** Rewriting AI text *with* an AI re-injects AI-ness; favor surgical edits to the flagged spans over wholesale regeneration.

**Processing modes:**
- **Fast Mode** (default, short text ≤ ~5,000 chars): single pass — detect → rewrite → self-verify.
- **Strict Mode** (long text ≥ ~8,000 chars or when explicitly requested): separate detection / rewrite / dual-validation passes. Recommend a literal read-aloud (声に出して読む) verification for rhythm (JA-04 / JA-05).

---

## Source & License

This module's structure follows the **im-not-ai (Humanize KR)** open-source skill (https://github.com/epoko77-ai/im-not-ai, MIT License — Copyright (c) 2026 epoko77-ai). The Japanese tell patterns here were independently web-researched, not ported from the source. See NOTICE.md.
