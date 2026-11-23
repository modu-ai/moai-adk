---
id: SPEC-20261122-model-fast-pair
title: "the model choice never pairs Opus 4.8 with fast mode, which the API refuses from 2026-11-30"
status: draft
revision: 2
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

- AC-1 (REQ-1): a test walks every `ModelRow` of every tool and fails on any id matching the family · verify: `cargo nextest run --test model_effort_settings`
- AC-2 (REQ-2): a hand-edited launcher file with (a) an Opus 4.8 model entry and (b) an entry holding a third `fastMode` key ends with both key paths in `unsupported_entries`, and the resolve is refused (a saved entry the table does not accept is never resolved) · verify: `cargo nextest run --test model_effort_settings`
- AC-3 (REQ-3): the old settings object read by `claude_settings_merge` (the merge candidates of the backed-up Go-generation file) that pairs an Opus 4.8 model value with a fast-mode key gets its `model` row `blocked` with the fixed reason and no carried value, while the same model value alone stays `mergeable`; the module never writes, so the file stays as it was · verify: `cargo nextest run --lib -E 'test(fast_pair)'` (unit tests in `src/claude_settings_merge.rs`, named with `fast_pair`)
- AC-4 (REQ-4): the built model question offers no family id for every tool and role · verify: `cargo nextest run --test model_effort_questions`

## Scope

- `src/model_effort_settings.rs` — only if a test shows the family pin (REQ-1) or the entry refusal (REQ-2) is missing; the OMP-removal card has landed (revision 2: the table and `check_entry` read on 2026-11-23 already hold both, so REQ-1, REQ-2 and REQ-4 are pin tests that pass at once and are mutation-checked instead of RED-first).
- `src/claude_settings_merge.rs` — the pair refusal in `model_row`/`rows_of` (REQ-3), with its unit tests in the file's `mod tests`.
- `tests/model_effort_settings.rs` — AC-1 and AC-2; `tests/model_effort_questions.rs` — AC-4. (Revision 1 named `tests/model_effort_cli.rs`, which does not exist.)

### Sources, rechecked in revision 2

The shell clock of the session read `Mon Nov 23 15:07:15 UTC 2026` (`date -u`) at the first fetch; the session's own date was 2026-11-24, and the two were not reconciled. A read time without a measured clock is marked unverified.

- Claude Code fast mode page, `https://code.claude.com/docs/en/fast-mode` and `.md` (read whole, 2026-11-23T15:07:15Z): "On the Anthropic API, fast mode is supported on Fable 5.2, Opus 5.5, Opus 5, and Opus 4.8, but not on Sonnet, Haiku, or other models." and "Fast mode for Opus 4.8 was deprecated on October 30, 2026, and will be removed on November 30, 2026. If Opus 4.8 is still your model with fast mode on after that date, the API rejects your requests." Claude Code v2.1.280 and later default to Opus 5.5 for fast mode.
- Claude API fast mode page, `https://platform.claude.com/docs/en/build-with-claude/fast-mode.md` (read whole, read time unverified, after 2026-11-23T15:07:15Z): "Removal date: November 30, 2026. After removal, requests to `claude-opus-4-8` with `speed: "fast"` will return an error. To continue using fast mode, migrate to Claude Opus 5.5." Its front matter lists `claude-fable-5-2`, `claude-mythos-5-2`, `claude-opus-5-5`, `claude-opus-5` and `claude-opus-4-8` as the models that take fast mode; Opus 4.8 is "Deprecated".
- Claude Code changelog 2.1.316, `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` (read in part, characters 0 to 300000, read time unverified): "Deprecated fast mode for Opus 4.8 as of October 30, 2026; it will be removed on November 30, 2026. After removal, fast mode requests on Opus 4.8 return an error and do not fall back to standard Opus 4.8." (fragmented relay; the trailing clause is not verified verbatim).
- Claude Code settings reference, `https://code.claude.com/docs/en/settings-reference.md` (read in part, characters 0 to about 400000 of 460800, read time unverified): the keys are `fastMode` (boolean, "false: fast mode stays off") and `fastModePerSessionOptIn` (boolean, "a saved `fastMode: true` no longer turns fast mode on at session start"); they are the only fast-mode keys of the index table.
- Result of the recheck: adopt. The date and the refusal are as revision 1 states; the supported-model list is wider than revision 1 knew (Fable 5.2, Opus 5.5, Opus 5, Opus 4.8; the API page adds Mythos 5.2), and moai's table names none of the Opus 4.8 family, so the pin stays correct and Opus 5.x ids stay unaffected (D-2). Unverified: the rest of the changelog and the last 60000 characters of the settings page.

## Decisions

- D-1: moai still never turns fast mode on or off (report section 17 ⑦); this SPEC adds a refusal, not a lever.
- D-2: the family match is the Opus 4.8 ids alone (`opus-4-8`, `opus-4.8`, any case); Opus 5.x ids stay unaffected.
- D-3: the refusal carries no expiry in code — after 2026-11-30 the pair is gone at the API and the guard is simply always right; a dated toggle would put a clock in a pure table.
- D-4: the code slice starts after the OMP-removal card lands (its reservation holds the model files); the questions module needs no edit — its options come from the table, and REQ-4 pins the outcome, not the file. Revision 2: the OMP removal has landed on NeXT (`src/model_effort_settings.rs` read 2026-11-23 names no retired tool in its table), so the slice may start once this revision is approved.
- D-5 (revision 2): REQ-3 keeps revision 1's rule, a fast-mode key being present at all; the docs read on 2026-11-23 say `fastMode: false` keeps fast mode off, so that one case is refused though harmless; the rule stays simple because the refused value only stays in the backup, and a narrower rule is a follow-up, not this slice.

## Plan

1. The SPEC lands first (this revision) and waits for the leader's review.
2. After the OMP-removal card lands: the family pin and the entry-refusal tests (AC-1, AC-2, AC-4), the merge pair refusal (AC-3), one affected run at READY with the same-head receipt.

## Risks

- The date bound (2026-11-30) is close; the guard is small and the tests are pure, so the slice is one sitting once the model files are free.
- The merge refusal must not learn to quote values: its reason is a fixed string (the module's existing rule).

## Result

Not started (the SPEC is the first slice).
