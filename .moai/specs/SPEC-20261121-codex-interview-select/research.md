# Research: the Default-mode availability of `request_user_input`, re-checked 2026-11-21

SPEC-20261121-codex-interview-select. The facts behind the Codex CLI row change of
`skills/moai/references/internal/questions.md`: the install-flow report (2026-11-21 §7) observed that
a `moai codex` session in Default mode asks the interview as plain-text numbered choices, and the
user pick (leader decision note, 2026-11-20) was to re-check the current official Codex
documentation and either apply real choice prompts or document the limit. This note records that
re-check. The earlier note `SPEC-20261118-codex-question-tool/research.md` (read 2026-11-18 from
tag `rust-v0.170.0`) stays the detailed source map; this note only adds what changed — and what did
not.

Conclusion first: the answer to "does `request_user_input` work in Default mode?" is still **no, not
as a shipped capability**. Default mode lists the tool only when the feature flag
`features.default_mode_request_user_input` is on; the flag is under development and off by default in
the released 0.171.0; the official config reference does not document it; the `moai codex` launcher
does not set it. With the flag on, the tool works but a Default-mode request that goes unanswered
resolves by itself after about a minute, so even the flag-on path needs the no-answer rule. The
documentation branch of the user pick applies.

## The official documentation (no flag documented)

- url: https://learn.chatgpt.com/docs/config-file/config-basic
- read: 2026-11-21T02:43:24Z
- sha256: f356e0e77d02370adb1991c9117a76ebe90635a6148663c7793b574dbbf8ed13 (fetched 2026-11-21T02:43:24Z, 440814 bytes)
- quote: the page's feature-flags table names `apps`, `goals`, `hooks`, `fast_mode`, `memories`,
  `multi_agent`, `remote_plugin`, `shell_snapshot`, `shell_tool`, `unified_exec`, `web_search`,
  `web_search_cached`, `web_search_request` — and nothing else; the page says "This table lists
  common user-facing flags, not every internal or under-development feature."
- kind: docs

`default_mode_request_user_input` has zero occurrences in the fetched page (`grep -c` printed 0, run
2026-11-21T02:43:24Z), so the official documentation still does not present the flag as a supported
setting. This matches the earlier note's finding (read 2026-11-18T06:15:45Z) that the docs never
mention the tool at all.

## The installed binary (0.171.0, same defaults as 0.170.0)

- command: `codex --version` and `codex features list`
- read: 2026-11-21T02:43:22Z
- output: `codex-cli 0.171.0` and `default_mode_request_user_input          under development  false`

- command: `strings` over `/Users/goos/.local/bin/codex` (read-only), searched for the
  Default-mode permission instruction
- read: 2026-11-21T02:43:23Z
- output: `Never use the `request_user_input` tool for permission requests or permission-related escalations`
- kind: source (installed binary)

So the row's citation moves from "installed Codex 0.170.0 (read from its binary, 2026-11-19)" to
"installed Codex 0.171.0 (read from its binary, 2026-11-21)"; the sentence itself is unchanged.

## The released source at tag `rust-v0.171.0` (published 2026-11-20T00:44:47Z)

- quote: `key: "default_mode_request_user_input",` with `stage: Stage::UnderDevelopment,` and
  `default_enabled: false,` (lines 2163–2168)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.171.0/codex-rs/features/src/lib.rs
- read: 2026-11-21T02:40:06Z (fetched 2026-11-21T02:38Z, re-hashed below)
- sha256: 5b191163de22c539b7a587d1d79db2d6b872812790cfb83da3506c62a2a55ff3 (fetched 2026-11-21T02:43Z re-fetch verified, 89300 bytes)
- kind: source
- quote: `pub const fn allows_request_user_input(self) -> bool {` (line 700; the body matches only
  `Self::Plan`)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.171.0/codex-rs/protocol/src/config_types.rs
- read: 2026-11-21T02:40:06Z
- sha256: ce9b26223db7cc4b5cf474e571e4509e8cc52536e159510f153f94de0fcd37e8 (fetched 2026-11-21T02:43Z re-fetch verified, 31651 bytes)
- kind: source
- quote: `mode.allows_request_user_input()` `|| (features.enabled(Feature::DefaultModeRequestUserInput)` and `*mode == ModeKind::Default)` (lines 10–13; the line breaks are joined here)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.171.0/codex-rs/tools/src/tool_config.rs
- read: 2026-11-21T02:40:06Z
- sha256: 1925968949c32190288dbc3a1473520ced9c66da425996a0652dd572b1341aae (fetched 2026-11-21T02:43Z re-fetch verified, 1004 bytes)
- kind: source

Two of the three hashes (`config_types.rs`, `tool_config.rs`) are byte-identical to the hashes the
earlier note recorded for tag `rust-v0.170.0`: the Plan-only gating and the Default-plus-flag tool
gate did not change between 0.170.0 and 0.171.0. The features file grew (new features elsewhere) but
this flag's stage and default are the same as the earlier note quoted.

## The launcher does not set the flag

- command: `rg "default_mode_request_user_input" src skills` in this worktree (head 041536fe07)
- read: 2026-11-21T02:3xZ (re-confirmed with the row edit)
- output: the only hit is `skills/moai/references/internal/questions.md:18` — no `src/` hit
- kind: source (this repository)

So the `moai codex` launcher writes no `features.default_mode_request_user_input` in this version,
which is what the row now says instead of the stale "is to turn it on" attribution (the sentence the
earlier card left open has no card of its own; its closing note said so as open item 2).

## The upstream issues behind the two limits

- quote: "the tool was present in the tool definitions … when I tried to call it, the runtime
  returned: `request_user_input is unavailable in Default mode`"
- url: https://github.com/openai/codex/issues/29104 (state open, created 2026-06-19T14:48:40Z)
- read: 2026-11-21T02:43:26Z (via `gh api repos/openai/codex/issues/29104`; the quoted sentence
  re-read from the body in the same step)
- kind: docs
- quote: "I have `request_user_input` explicitly enabled in Default mode using the
  `default_mode_request_user_input` feature. After updating to 0.147.0: `request_user_input` requests
  made outside Plan mode automatically resolve if I don't respond within a minute or two", citing
  #36410 ("made blocking behavior of `request_user_input` explicit and currently classifies Plan-mode
  requests as blocking while requests from other enabled modes are non-blocking")
- url: https://github.com/openai/codex/issues/37472 (state open, created 2026-08-07T17:03:16Z)
- read: 2026-11-21T02:43:26Z (via `gh api repos/openai/codex/issues/37472`; the quoted sentence
  re-read from the body in the same step)
- kind: docs
- related, same step: #30150 (open) asks to expose the tool outside Plan mode; #34455, #28969 and
  #29702 (cited by #37472) ask for configurable or disableable auto-resolution.
- kind: docs

unverified: whether the auto-resolution timeout of #37472 still measures one to two minutes in
0.171.0 (the issue is open and the timeout is client-side; the earlier note found the core wait
`rx_response.await` has no timeout of its own, so the resolution happens in the client, which a lane
cannot read without a live TUI session). The row therefore records the behavior as issue-reported,
with the read date, and keeps the no-answer rule that is correct under either timing.

## What the re-check changes

1. The row's availability parenthetical: "the `moai codex` launcher is to turn it on — card C-52, not
   in this version" → "the `moai codex` launcher does not turn it on in this version" (the attribution
   was stale: the earlier card closed having delivered only the documentation and the pin test).
2. The row gains the two Default-mode limits with their citations (#29104 refusal without the flag,
   #37472 auto-resolution with it) and the interview consequence: a returned no-answer is not a pick,
   and an interview question is re-asked in plain text, one per message, and waited for.
3. The row's binary citation: 0.170.0 / 2026-11-19 → 0.171.0 / 2026-11-21 (sentence unchanged).
4. Everything else in the row (call shape, Plan-mode behavior, the case-(2) plain-text rule, the
   async ban, the fallback column) stays as it is, because today's sources confirm it unchanged.
