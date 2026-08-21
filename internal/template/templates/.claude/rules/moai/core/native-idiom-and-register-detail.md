---
description: "Detail companion for native-idiom-and-register.md — why calques survive generation, and the illustrative Korean calque hazard table"
paths: "**/native-idiom-and-register.md,**/.claude/skills/moai-domain-humanize/**"
---

# Native Idiom & Register — Detail Companion

> Detail companion of `native-idiom-and-register.md` (the always-loaded stub). The stub owns the
> invariant, the two-register split, the humanize-invocation rule, and the pre-emit self-check.
> This file owns the generation mechanism behind the hazard and the worked Korean examples. Load it
> when a calque is suspected and the substitution is not obvious.

## Why calques survive (the mechanism)

A calque survives generation because the conceptual skeleton is English-shaped and the model emits the nearest dictionary equivalent without checking whether that word carries the same figurative sense in the target language. English technical and marketing prose favors architectural and geometric metaphors ("pillars", "axes", "three-legged stool", "budget defense"); these metaphors are vivid inside English idiom but become awkward literal translations in Korean / Japanese / Chinese, where the register is more direct. The absence of a mechanical de-calque step in the default path is what lets the calque through — which is why this rule exists and why heavy artifacts route through `moai-domain-humanize`.


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


---

Classification: Lazy companion — mechanism and illustrative examples only. The authoritative
per-locale catalogue is the `moai-domain-humanize` skill; every rule stays in
`native-idiom-and-register.md`.
