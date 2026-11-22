---
id: SPEC-20261122-model-fast-pair
title: "the model choice never pairs Opus 4.8 with fast mode, which the API refuses from 2026-11-30"
status: draft
revision: 1
created: 2026-11-22
---

# the model choice never pairs Opus 4.8 with fast mode, which the API refuses from 2026-11-30

## Intent

From 2026-11-30 the Claude API refuses a request that pairs Opus 4.8 with fast mode ("the API rejects your requests", fast-mode.md; changelog 2.1.316, both quoted in the whole-docs report `~/.moai/reports/moai-adk/claude-code-native-redesign-2026-11-22.md` sections 11 and 14, read whole 2026-11-22; coverage row A16). moai never turns fast mode on or off (user decision, report section 17 ⑦), but three moai surfaces could still hand the host that pair: the model support table, a saved `launcher.models` entry, and a settings template that writes both a model value and a fast-mode key into one settings object. None offers the pair today; this SPEC pins that none ever does, with a refusal where a hand edit or a future template could introduce it.

## Non-goals

- No fast-mode lever is added anywhere: moai keeps never turning fast mode on or off (report section 17 ⑦).
- No clock, no dated toggle: the refusal is unconditional (D-3).
- The questions module (`model_effort_questions`) is not edited: its options come from the table, and its file is held by the OMP-removal card until that lands.

## Requirements

- REQ-1: The support tables of `model_effort_settings` shall name no model id of the Opus 4.8 family in any letter case (opus-4-8 or opus-4.8) · failure: a table row naming the family makes the pinning test fail.
- REQ-2: A `launcher.models` entry shall stay exactly `model` and `effort` — a third key (a fast-mode lever) or an Opus 4.8 model id is never accepted, saved or resolved, and `unsupported_entries` lists the key path · failure: a hand-written entry with a fast-mode key or an Opus 4.8 id is refused or listed, never used.
- REQ-3: A settings merge input that pairs an Opus 4.8 model value with a fast-mode key (fastMode or fastModePerSessionOptIn) shall be refused with a fixed reason that names neither saved text nor a secret and leave the file exactly as it was · failure: a merge input holding the pair is refused and the file stays as it was.
- REQ-4: The model question and the update models stage shall offer only table models, so no question path can name an Opus 4.8 id while REQ-1 holds · failure: the question builder's options contain no family id for any tool or role.

## Acceptance

- AC-1 (REQ-1): a test walks every `ModelRow` of every tool and fails on any id matching the family · verify: `cargo nextest run --test model_effort_cli`
- AC-2 (REQ-2): a hand-edited launcher file with (a) an Opus 4.8 model entry and (b) an entry holding a third `fastMode` key ends with both key paths in `unsupported_entries`, and the resolve falls back to the defaults · verify: same target
- AC-3 (REQ-3): a merge input pairing an Opus 4.8 model value with a fast-mode key is refused with the fixed reason and writes nothing, while a model value alone still merges · verify: the settings-merge target named in the Plan
- AC-4 (REQ-4): the built model question offers no family id for every tool and role · verify: same target as AC-1

## Scope

- `src/model_effort_settings.rs` — the family pin (REQ-1) and the entry refusal (REQ-2); held by the OMP-removal card until it lands, so this slice starts after that landing.
- `src/claude_settings_merge.rs` — the pair refusal at the merge surface (REQ-3).
- `tests/model_effort_cli.rs` (or the merge target's own test file where AC-3 reads better) — the four ACs.

## Decisions

- D-1: moai still never turns fast mode on or off (report section 17 ⑦); this SPEC adds a refusal, not a lever.
- D-2: the family match is the Opus 4.8 ids alone (`opus-4-8`, `opus-4.8`, any case); Opus 5.x ids stay unaffected.
- D-3: the refusal carries no expiry in code — after 2026-11-30 the pair is gone at the API and the guard is simply always right; a dated toggle would put a clock in a pure table.
- D-4: the code slice starts after the OMP-removal card lands (its reservation holds the model files); the questions module needs no edit — its options come from the table, and REQ-4 pins the outcome, not the file.

## Plan

1. The SPEC lands first (this revision) and waits for the leader's review.
2. After the OMP-removal card lands: the family pin and the entry-refusal tests (AC-1, AC-2, AC-4), the merge pair refusal (AC-3), one affected run at READY with the same-head receipt.

## Risks

- The date bound (2026-11-30) is close; the guard is small and the tests are pure, so the slice is one sitting once the model files are free.
- The merge refusal must not learn to quote values: its reason is a fixed string (the module's existing rule).

## Result

Not started (the SPEC is the first slice).
