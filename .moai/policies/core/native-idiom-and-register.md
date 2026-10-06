# Native-Idiom & Register Policy (Non-English Locales)

> The language-quality invariant for non-English output. Always-loaded.
> Owns: the anti-calque invariant and the humanize trigger. Cross-references `moai-constitution.md` § Response Language and the `moai-domain-humanize` skill.

## The Invariant

[ZONE:Evolvable] [HARD] When `conversation_language ≠ en`, every user-facing surface — chat replies, reports, README, docs-site, generated sites, `AskUserQuestion` text — MUST read as natural native prose, NOT as English mapped one-to-one onto the target language. Translation-style calques (direct word-for-word carry-over of English syntax, metaphor, and figurative stock) are prohibited; native idiom is required.

English is the source language. This policy is **conditional**: it imposes zero overhead on English sessions, because the calque hazard exists only on the source→target direction.

Two registers, not one: chat replies take the colloquial native register, artifacts (reports, README, docs-site, generated sites) the clean native written register. A report in colloquial register is wrong; so is one in calqued register.

## Mechanism — when to invoke humanize

[ZONE:Evolvable] [HARD] Heavy non-English artifacts (multi-paragraph reports, README rewrites, docs-site pages, generated sites) MUST pass through the `moai-domain-humanize` skill as a final phase before delivery, scoped to the active locale's module (`modules/korean.md` / `japanese.md` / `chinese.md`). Single-turn chat replies apply this rule inline (no skill invocation needed) — the rule above is the inline standard.

## Cross-references

- `.claude/rules/moai/core/moai-constitution.md` § Response Language — the conversation_language requirement this policy specializes.
- `.claude/skills/moai-domain-humanize/` — the per-locale calque catalogue (Category A) and the post-edit pass machinery.
- `native-idiom-and-register-detail.md` — the lazy companion. Load it for § Why calques survive (the mechanism) · § Calque hazard list · § Two registers — do not conflate (the per-surface register table) · § Pre-emit self-check (non-English output only).

---

Version: 1.0.0
Classification: Always-loaded language-quality invariant — non-English conditional.
