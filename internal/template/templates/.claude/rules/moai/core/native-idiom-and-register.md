# Native-Idiom & Register Policy (Non-English Locales)

> Moved to the detail companion: `native-idiom-and-register-detail.md` ("Native-Idiom & Register Policy (Non-English Locales)").

## The Invariant

[ZONE:Evolvable] [HARD] When `conversation_language ≠ en`, every user-facing surface — chat replies, reports, README, docs-site, generated sites, `AskUserQuestion` text — MUST read as natural native prose, NOT as English mapped one-to-one onto the target language. Translation-style calques (direct word-for-word carry-over of English syntax, metaphor, and figurative stock) are prohibited; native idiom is required.

> Moved to the detail companion: `native-idiom-and-register-detail.md` ("The Invariant").

Two registers, not one: chat replies take the colloquial native register, artifacts (reports, README, docs-site, generated sites) the clean native written register. A report in colloquial register is wrong; so is one in calqued register.

## Mechanism — when to invoke humanize

[ZONE:Evolvable] [HARD] Heavy non-English artifacts (multi-paragraph reports, README rewrites, docs-site pages, generated sites) MUST pass through the `moai-domain-humanize` skill as a final phase before delivery, scoped to the active locale's module (`modules/korean.md` / `japanese.md` / `chinese.md`) — where the skill is installed (optional packs carry it; the default core catalog does not). A project deployed without it applies this rule inline as the humanize pass instead. Single-turn chat replies apply this rule inline (no skill invocation needed) — the rule above is the inline standard.

## Cross-references

> Moved to the detail companion: `native-idiom-and-register-detail.md` ("Cross-references").

---
