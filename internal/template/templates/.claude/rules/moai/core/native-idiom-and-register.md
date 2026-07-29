# Native-Idiom & Register Policy (Non-English Locales)

> The language-quality invariant for non-English output. Always-loaded.
> Owns: anti-calque rules + conversational/artifact register split. Cross-references `moai-constitution.md` § Response Language and the `moai-domain-humanize` skill.

## The Invariant

[ZONE:Evolvable] [HARD] When `conversation_language ≠ en`, every user-facing surface — chat replies, reports, README, docs-site, generated sites, `AskUserQuestion` text — MUST read as natural native prose, NOT as English mapped one-to-one onto the target language. Translation-style calques (direct word-for-word carry-over of English syntax, metaphor, and figurative stock) are prohibited; native idiom is required.

English is the source language. This policy is **conditional**: it imposes zero overhead on English sessions, because the calque hazard exists only on the source→target direction.

## Why calques survive (the mechanism)

A calque survives generation because the conceptual skeleton is English-shaped and the model emits the nearest dictionary equivalent without checking whether that word carries the same figurative sense in the target language. English technical and marketing prose favors architectural and geometric metaphors ("pillars", "axes", "three-legged stool", "budget defense"); these metaphors are vivid inside English idiom but become awkward literal translations in Korean / Japanese / Chinese, where the register is more direct. The absence of a mechanical de-calque step in the default path is what lets the calque through — which is why this rule exists and why heavy artifacts route through `moai-domain-humanize`.

## Two registers — do not conflate

| Surface | Register | Rule |
|---|---|---|
| **Chat / conversation** (replies to the user) | Colloquial native register (Korean: 해요체). Persona-dependent: MoAI-Easy leans colloquial; MoAI stays professional. | Speak to the user the way a native colleague would speak. No calqued figurative nouns. |
| **Artifacts** (reports, README, docs-site, generated sites) | Clean native written register (문어) — professional, native idiom, de-calqued. | The prose a competent native engineer writes. NOT colloquial. NOT calqued. |

A report in colloquial register is wrong (too casual). A report in calqued register is also wrong (translationese). Artifacts want clean native written prose.

## Calque hazard list (illustrative — Korean)

The authoritative per-locale catalogue is the `moai-domain-humanize` skill, Category A (번역투 / Translationese / Calque) and A-23 (metaphor calque). The table below is a short pointer, not a duplicate — when in doubt, defer to the skill's catalogue.

| Calque (avoid) | Native idiom (prefer) |
|---|---|
| 3축 / 세 축 / 세 가지 기둥 / "Three Axes" (docs-site pillar headings, README sections) | 세 가지 핵심 / 세 가지 핵심 가치 |
| 7대 기둥 ("seven pillars") | 7가지 핵심 차별점 / 일곱 가지 강점 |
| 검증경제 ("verification economy") | 검증 비용을 줄이는 방식 |
| 예산방어 ("budget defense") | 예산 초과 전에 중단하기 |
| 회로차단기 ("circuit breaker", token context) | 토큰 예산 가드 / 자동 중단기 |

Deliberately-coined brand terms (e.g. "토크노믹스" / "tokenomics") and established loanwords (e.g. "라우팅" / "routing") are NOT calques — they are intentional vocabulary. The prohibition binds only to figurative and structural carry-over.

## Mechanism — when to invoke humanize

[ZONE:Evolvable] [HARD] Heavy non-English artifacts (multi-paragraph reports, README rewrites, docs-site pages, generated sites) MUST pass through the `moai-domain-humanize` skill as a final phase before delivery, scoped to the active locale's module (`modules/korean.md` / `japanese.md` / `chinese.md`). Single-turn chat replies apply this rule inline (no skill invocation needed) — the rule above is the inline standard.

## Pre-emit self-check (non-English output only)

- [ ] Did I read `conversation_language`? If `en`, this policy does not apply — stop here.
- [ ] Chat reply: colloquial native register, with no English-mapped figurative nouns in headings or body?
- [ ] Artifact: clean native written register — no calqued metaphors (축 / 기둥 / 검증경제 type), no English-syntax carry-over?
- [ ] Heavy artifact: has the `moai-domain-humanize` final pass run (or is it scheduled)?

## Cross-references

- `.claude/rules/moai/core/moai-constitution.md` § Response Language — the conversation_language requirement this policy specializes.
- `.claude/skills/moai-domain-humanize/` — the per-locale calque catalogue (Category A) and the post-edit pass machinery.
- `.claude/output-styles/moai/*.md` § Language Rules — per-persona register (MoAI-Easy: colloquial; MoAI / MoAI-Learn: professional).

---

Version: 1.0.0
Classification: Always-loaded language-quality invariant — non-English conditional.
