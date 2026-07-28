---
id: SPEC-I18N-GOVERNANCE-001
title: "Web Console i18n Catalogue Governance"
version: "0.1.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-29
author: manager-spec
priority: P2
phase: "v3.0.2 target"
module: "internal/web (i18n catalogue + untranslated-value governance tests)"
lifecycle: spec-anchored
tags: "i18n, web-console, governance, allowlist, translation, quality-gate"
tier: M
---

# SPEC-I18N-GOVERNANCE-001 — Web Console i18n Catalogue Governance

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | Initial draft — plan-phase authoring (spec/plan/acceptance) |

## §A Context

`internal/web/assets/i18n.js` is the MoAI web console's interface-chrome dictionary: a single `window.MOAI_I18N` object literal carrying four locale blocks (`en`, `ko`, `ja`, `zh`), loaded offline via `go:embed` and applied by `applyI18n()` in `app.js`, which swaps the text of every `[data-i18n]` element.

The catalogue is **structurally** well-tested and **semantically** ungoverned. `internal/web/i18n_test.go` verifies that the dictionary is embedded, that it is served from `/static/i18n.js`, that rendered `data-i18n` keys exist somewhere in the file, and that design-review keys are absent. `internal/web/webux_followup_test.go` adds per-locale assertions for one key family. **No test compares a translated value to its English counterpart.** A contributor can add a key to all four locales by pasting the English string three times and every existing test stays green.

### A.1 Measured baseline (worktree base, `origin/main`)

All figures below were measured directly from `internal/web/assets/i18n.js` at the SPEC's worktree base by evaluating the object literal and comparing locale maps. They are the plan-phase ground truth; run-phase re-measures before acting.

| Measurement | Value |
|---|---|
| Locales | `en`, `ko`, `ja`, `zh` |
| Key count | `en` = 340; `ko` / `ja` / `zh` = 350 each |
| `en` keys absent from a non-`en` locale | 0 (forward coverage is complete) |
| Non-`en` keys absent from `en` | 10 — the entire `agentdesc.*` family, in each of `ko` / `ja` / `zh` |
| Values byte-identical to `en` | exactly 25 per locale |
| Identical-set intersection across `ko` / `ja` / `zh` | 24 |
| Identical-set union across `ko` / `ja` / `zh` | 27 |
| Per-locale divergence in the identical sets | exactly one key each — the locale's own `lang.opt.<locale>` |
| Additional matches under trim + casefold normalization | 1 per locale — `mp.col.effort` (`en` = `"Effort"`, all three = `"effort"`) |
| Values containing quote / backslash / newline escapes | 0 |
| Keys outside `[A-Za-z0-9._-]` | 4 — `f.report.format.opt.html+md`, `f.model.opt.{opus,sonnet,fable}[1m]` |
| Template-tree mirror of `i18n.js` | none (the catalogue is `internal/web`-only, not distributed) |

### A.2 Why a naive identity rule is unusable

The obvious mechanical rule — *a non-`en` value equal to its `en` value is untranslated* — flags 25 keys per locale, of which the large majority are **correct**. Enum literals (`md`, `html+md`, `manual`, `auto`, `1h`, `5m`, `off`), model names (`Opus 4.8 (200K)`, `Haiku`), convention names (`Angular`, `Karma`, `Conventional commits`), a product name (`MoAI-Loop`), and an initialism (`LLM`) are all locale-invariant by intent. Shipping that rule as-is would produce roughly two dozen false positives per locale and would be switched off within a week.

Two members of the identical set are **not** false positives:

- `mp.tier.empty` = `"(runtime default: medium)"` — English prose. Its sibling `mp.tier.default` **is** translated in every locale (`ko` = `"(런타임 기본값: medium — llm.performance_tier 미설정)"`), so the omission is an inconsistency, not a decision.
- `board.badge.mustfix` = `"MUST-FIX"` — borderline; see §A.4.

The governance mechanism therefore cannot be an identity rule alone. It must be an identity rule **plus an explicit, reasoned allowlist of intentionally-untranslated keys**, so that the check fails only on keys nobody has justified.

### A.3 Why the `lang.opt.*` family must not be handled by the allowlist

`lang.opt.ko` in the `ko` block is `"한국어"` — identical to the `en` block, because the `en` block itself renders language names as endonyms (`lang.opt.ko` = `"한국어"`, `lang.opt.ja` = `"日本語"`). This is the single reason the three locales' identical sets diverge at all: remove the `lang.opt.*` family and the sets become byte-identical at 24 keys each.

Allowlisting `lang.opt.ko` would suppress the check for that key in **every** locale, so a future copy of the English endonym into `ja` or `zh` would pass silently. The family instead admits a positive, bidirectional invariant that holds today across all sixteen locale × language pairs and needs zero allowlist entries — see REQ-I18NGOV-012 and REQ-I18NGOV-013.

### A.4 The `agentdesc.*` reverse-coverage gap is deliberate

The 10 keys present in `ko` / `ja` / `zh` but absent from `en` are not an oversight. `internal/web/webux_followup_test.go` carries `TestD3AgentDescIsEnExempt`, which **asserts** the `en` block defines no `agentdesc.*` key: English reads the agent `.md` frontmatter description as the server-rendered baseline (the single source), and `applyI18n()` guards its assignment on a non-empty string so an absent key leaves that baseline intact. Duplicating the `.md` text into the dictionary would create a second surface that silently goes stale.

Reverse coverage therefore cannot be a plain key-set equality. It must be equality **modulo a declared registry of `en`-exempt key prefixes**, whose sole initial member is `agentdesc.` — which converts an undocumented asymmetry into a bounded, reviewable one.

## §B Requirements (GEARS)

### B.1 Allowlist artifact

- **REQ-I18NGOV-001** (Ubiquitous): The repository shall carry exactly one untranslated-value allowlist artifact, co-located with the console package under `internal/web/` and located **outside** `internal/web/assets/`, so the allowlist is never exposed by the `/static/` handler.
- **REQ-I18NGOV-002** (Ubiquitous): Each allowlist entry shall carry three fields — the exact catalogue key, a reason drawn from the closed taxonomy of REQ-I18NGOV-003, and a free-text justification naming why the value is locale-invariant.
- **REQ-I18NGOV-003** (Ubiquitous): The reason taxonomy shall be closed and shall consist of exactly three values: `technical-identifier` (the value must match a configuration value, CLI token, or machine-emitted string byte-for-byte), `proper-noun` (a product, brand, vendor, model, or convention name), and `acronym` (a locale-invariant initialism).
- **REQ-I18NGOV-004** (Unwanted behavior): The allowlist shall not accept wildcard, prefix, glob, or regular-expression entries; every entry shall name one exact catalogue key.
- **REQ-I18NGOV-005** (Ubiquitous): Allowlist entries shall be scoped per key and shall apply to every non-`en` locale, because after the `lang.opt.*` family is removed by REQ-I18NGOV-011 the three locales' identical sets are byte-identical; per-locale narrowing shall be introduced only when a measurement demonstrates divergence.
- **REQ-I18NGOV-006** (Ubiquitous): The allowlist artifact shall carry an inline governance contract stating where the allowlist lives, its entry format, who may add an entry, and the assertion a reviewer shall make before accepting one.

### B.2 Detection mechanism

- **REQ-I18NGOV-007** (Ubiquitous): A parser shall produce, for each of the four locale blocks, a key-to-value map from the catalogue, correctly reading keys containing `+`, `[`, and `]`.
- **REQ-I18NGOV-008** (Ubiquitous): The untranslated-value detector shall be a pure function of its two inputs — a parsed catalogue and an allowlist — so a caller can supply a synthetic catalogue and a synthetic allowlist without touching the shipped files.
- **REQ-I18NGOV-009** (Ubiquitous): Value comparison shall normalize both operands by Unicode NFC composition, surrounding-whitespace trimming, and case folding before testing equality.
- **REQ-I18NGOV-010** (Event-driven): **When** the detector finds a non-`en` value whose normalized form equals its `en` counterpart, and the key is neither present in the allowlist nor a member of the `lang.opt.*` family, the detector shall report a violation naming the key and the locale.
- **REQ-I18NGOV-011** (Ubiquitous): The detector shall exclude the `lang.opt.*` key family from the generic identity comparison, delegating that family to the endonym invariants of REQ-I18NGOV-012 and REQ-I18NGOV-013.
- **REQ-I18NGOV-012** (Ubiquitous — endonym self-consistency): For every locale `L`, the value of `lang.opt.<L>` in locale `L` shall equal the value of `lang.opt.<L>` in `en`, because both render the endonym.
- **REQ-I18NGOV-013** (Ubiquitous — exonym distinctness): For every locale `L` and every language code `X` where `X` differs from `L`, the value of `lang.opt.<X>` in locale `L` shall differ from the value of `lang.opt.<X>` in `en`, because locale `L` renders the exonym.
- **REQ-I18NGOV-014** (Unwanted behavior): The allowlist shall not contain any key in the `lang.opt.*` family.

### B.3 Anti-rubber-stamp guards

- **REQ-I18NGOV-015** (Event-driven): **When** an allowlist entry names a key that is absent from the catalogue, or names a key whose value is not identical to `en` in any locale, the detector shall report that entry as an orphan, so a stale exemption cannot survive the deletion or translation of its key.
- **REQ-I18NGOV-016** (Ubiquitous): The allowlist shall carry no more than 30 entries, a bound derived from the 24 measured legitimate members plus headroom, so that blanket growth fails loudly rather than accreting silently.
- **REQ-I18NGOV-017** (Ubiquitous): A negative-control check shall demonstrate that a synthetic, non-allowlisted, genuinely-untranslated key produces a violation, and that the identical input produces no violation once the key is allowlisted.

### B.4 Key-coverage governance

- **REQ-I18NGOV-018** (Ubiquitous): Every key defined in the `en` locale shall be defined in each of the `ko`, `ja`, and `zh` locales, so a key cannot land in a single locale.
- **REQ-I18NGOV-019** (Event-driven): **When** a key defined in a non-`en` locale is absent from `en`, the coverage check shall report a violation unless the key matches a prefix declared in the `en`-exempt prefix registry.
- **REQ-I18NGOV-020** (Ubiquitous): The `en`-exempt prefix registry shall be an explicit, enumerated artifact whose sole initial member is `agentdesc.`, and each member shall carry a justification naming the surface that supplies the English baseline instead.

### B.5 Ownership and reachability

- **REQ-I18NGOV-021** (Ubiquitous): The `i18n.js` header comment shall name the owning governance surface — the allowlist artifact's path — and shall state the obligation that any catalogue change is reviewed against that contract.
- **REQ-I18NGOV-022** (Ubiquitous): All checks introduced by this SPEC shall execute as Go tests in package `web`, so that `go test ./internal/web/...` runs them with no new continuous-integration wiring.
- **REQ-I18NGOV-023** (Ubiquitous): The governance contract shall define a ruling procedure for a borderline value — classify the value against the closed taxonomy, cite mechanical evidence for the classification, and record the outcome as either an allowlist entry with justification or a translation fix.

## §C Constraints (DO NOT VIOLATE)

- **C1** — The `en`-exemption for `agentdesc.*` is preserved. `TestD3AgentDescIsEnExempt` and `TestD3AgentDescKeysInThreeLocales` shall continue to pass unmodified.
- **C2** — No fixture, allowlist entry, or test input shall contain a secret-shaped or key-shaped literal. Synthetic fixtures use inert placeholder keys and short interface words.
- **C3** — The `f.cacheStrategy.session_ttl.*` keys remain in the catalogue. This SPEC neither retires them nor depends on their backing configuration being live; three of them are ordinary `technical-identifier` allowlist candidates and are treated exactly like any other enum literal.
- **C4** — No new runtime behavior. The console renders identically before and after; every artifact added by this SPEC is consumed by tests only.
- **C5** — The catalogue has no template-tree mirror, so no dual-write and no template-neutrality obligation applies.

## §D Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — translation quality assessment

- Judging whether an existing translated value renders the English meaning **correctly**. This SPEC governs whether a value was translated **at all**, which is mechanically decidable; semantic accuracy is not.
- A prior investigation asserted a count of mistranslations in this catalogue. That count was not reproducible at the SPEC's worktree base and is therefore not carried into any requirement. A quality audit, if wanted, is a separate SPEC with a human-review deliverable.

### Out of Scope — `f.cacheStrategy.session_ttl.*` key retirement

- Removing or renaming the `session_ttl` key family in response to any backend retirement. That belongs to whichever SPEC owns the retirement; this SPEC only classifies the keys while they exist.

### Out of Scope — adding locales

- Introducing a fifth locale, or any change to the locale set, the language picker, or the CJK font subsets.

### Out of Scope — other i18n surfaces

- The docs-site four-locale content set, the README locale set, and any CLI-emitted string catalogue. This SPEC governs `internal/web/assets/i18n.js` only.

### Out of Scope — distribution

- Mirroring the catalogue, the allowlist, or the governance contract into `internal/template/templates/`. The console dictionary is not a distributed artifact.

## §E Cross-References

- `internal/web/assets/i18n.js` — the governed catalogue.
- `internal/web/assets/app.js` — `applyI18n()`, the consumer whose absent-key fallback underpins the `en`-exemption.
- `internal/web/i18n_test.go` — the existing structural test suite this SPEC extends rather than replaces.
- `internal/web/webux_followup_test.go` — `localeBlocks()`, the existing locale-block splitter, plus the `agentdesc.*` exemption contract (C1).
- `internal/spec/audit.go`, `internal/cli/spec_audit.go`, `internal/web/board.go` — the `"MUST-FIX"` severity literal that grounds the `board.badge.mustfix` ruling.
