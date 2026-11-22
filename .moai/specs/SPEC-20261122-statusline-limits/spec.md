---
id: SPEC-20261122-statusline-limits
title: "The third line of the status line shows the limits as the host gives them, and the render stops doing work nobody sees"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-statusline-flow
---

# 상태줄 3번째 줄: 호스트가 준 한도 그대로, 보이지 않는 일은 하지 않기

## Intent

부모 SPEC `SPEC-20261122-statusline-flow` 의 줄 역할 중 **3번째 줄(한도)**이에요. 사용자가 고른 것(리더 결정 기록 5935·5942): `CW 74% | 5H 23% | 7D 21% (6d 21h) | 💰 1.23 | 버전` 모양, `CW` 는 호스트의 사용 비율 그대로, `5H` 는 70% 를 넘을 때만, `💰` 는 한도 정보가 없을 때만, 모든 아이콘 뒤에 글자나 숫자와 띄어쓰기, moai 가 받은 입력을 Orca 래퍼에도 넘기기, 생성 설정의 `refreshInterval: 30`.

이 SPEC 이 함께 닫는 것은 호스트와 맞지 않는 곳이에요(리더 조사 보고서 2장, 공식 문서 `statusline.md` 2026-11-21T23:52Z 받음).

| 지금 | 문제 | 이 SPEC |
|---|---|---|
| `CTX` 를 자동 압축 기준 85% 로 다시 계산 | 호스트가 25% 라는데 29% 로 보여요(실측). `/context` 와 달라요. 공식 문서는 `used_percentage` 를 쓰라고 해요 | REQ-2 |
| `rate_limits` 가 있는 렌더마다 `~/.moai-cli/state/usage-limits.json` 에 써요 | 폐지된 폴더이고 "읽기만 한다" 규칙과도 어긋나요. 레인이 이 파일을 읽어요(`lane_screen.rs:746`, `ops/usage_limit.rs:23`) | REQ-7, D-5 |
| Go 디코더 흉내(`godecode.rs`)가 필드 하나만 형식이 달라도 줄 전체를 대체 문구로 바꿔요 | 문서의 입력 필드는 늘고 있어요(`spend_limit`, `pr.kind`, `prompt_id`) | REQ-10 |
| 화면에 안 나오는 값(캐시 적중률, 세션 시간, 출력 스타일, 목표 파일)을 모아요 | 헛일 | REQ-11 |
| Orca 상태줄 래퍼가 프로젝트 설정에 가려서 moai 창의 Orca 가 사용량을 못 받아요 | 설정 우선순위 | REQ-8 |
| 설정에 `refreshInterval` 이 없어요 | 쉬는 동안 원장·시간 값이 멈춰요 | REQ-9 |

## Non-goals

- 줄 1·2 의 내용(`where`, 부모 SPEC).
- `5H` 를 호스트가 보내지 않을 때 만들어 내는 것. 사용자가 들은 "5시간 한도가 없어질 수 있다"는 이야기는 공식 문서 확인(2026-11-21T23:52Z)에서 `rate_limits.five_hour` 가 아직 있고 "창마다 따로 없을 수 있다"고 적혀 있어요. 그래서 `5H` 는 호스트가 보낼 때만 그리고, 없을 때를 위한 코드는 따로 필요 없어요.
- 비용 계산, 청구액 안내. `cost.total_cost_usd` 는 호스트의 "추정" 값이고 실제 청구와 다를 수 있다고 문서가 말해요; 줄은 숫자만 보여요.
- Orca 쪽 코드. 이 SPEC 은 Orca 의 래퍼를 실행만 해요.

## Requirements

- REQ-1: The status line shall write its third line as the items context window, five-hour window, seven-day window, cost, Claude Code version and MoAI version in this order and in the widest of the forms of the Layouts section whose width fits the budget of the parent SPEC · failure: none — an item without its input is left out with its separator, the narrowest form is written when none fits and the line is left out when all items are
- REQ-2: The status line shall show the context window as `<battery> CW <n>%` with the stdin `context_window.used_percentage` unchanged and rescaled by nothing · failure: `CW n/a` with no battery icon when the percentage is missing, never `0%`
- REQ-3: When the stdin carries `rate_limits.seven_day`, the status line shall show `<battery> 7D <n>%` followed by its reset countdown in parentheses without the words `resets in` · failure: none — an absent window or one whose reset time has passed is left out and a window without a percentage says `7D n/a`
- REQ-4: When the stdin carries `rate_limits.five_hour` above 70 percent, the status line shall show `🪫 5H <n>%` followed by its reset countdown · failure: none — a window at 70 percent or below or an absent window is left out
- REQ-5: When the stdin carries neither a five-hour nor a seven-day window, the status line shall show `💰 <usd>` with the stdin `cost.total_cost_usd` to two decimals · failure: none — a missing cost is left out and a payload with a window shows no cost
- REQ-6: The status line shall show `🤖 <version>` from the stdin version and `🗿 <version>` of the build or `🗿 <template> -> <build>` when the project's template is older · failure: none — a missing version is left out
- REQ-7: The status line shall write the usage windows it receives to one record file in the cache folder of the status line only when a percentage or a reset time changed · failure: none — nothing is written under the user's home folder and an unwritable folder drops the record with the line unchanged
- REQ-8: When the Orca hook variables are set and the Orca status line script is an executable file and the stdin holds `rate_limits`, the status line shall start that script detached with the received stdin · failure: none — the render never waits for it and a missing script or variable starts nothing
- REQ-9: The status line settings generator shall write the command `moai statusline`, `refreshInterval` of 30 and the `padding` of the status line configuration into the host's `statusLine` block · failure: a block the user wrote is reported and left as it is
- REQ-10: The status line shall read the stdin field by field so that a field of the wrong type or an unknown field changes only itself · failure: none — a payload that is not a JSON object reads as empty and the line says `n/a` where it must
- REQ-11: The status line shall collect only the stdin fields and files that a shown item needs · failure: none — the cache hit ratio, the session duration, the output style, the agent name, the transcript path and the goal file are not read
- REQ-12: The status line shall start in one render at most the process of its git read and the process of the Orca script and write at most the cache and the record · failure: none — a render that needs more drops the extra work and the line is unchanged

## Acceptance

- AC-1 (REQ-1, REQ-2): the line of every case — subscriber with the five-hour window, subscriber without it, API key user, empty window, null percentage — is exactly the string of its golden, plain and with color; `CW` equals the host's percentage with no rescale for 25, 85 and 100 and a missing percentage says `CW n/a`.
  verify: `cargo test --offline --locked --test statusline_cards line_three_items_`
- AC-2 (REQ-3, REQ-4): the icon switch at 70/71, the countdown forms (`6d 21h`, `2h 10m`, `35m`), the hidden window after its reset, `7D n/a`, the five-hour window shown at 71 and hidden at 70 are pinned.
  verify: `cargo test --offline --locked --lib statusline::render::window_tests`
- AC-3 (REQ-5): cost is shown with no window, with only a spend limit and hidden with either window, and rounds to two decimals.
  verify: `cargo test --offline --locked --lib statusline::render::cost_tests`
- AC-4 (REQ-6): the versions, the update arrow and a missing version are pinned.
  verify: `cargo test --offline --locked --lib statusline::render::version_tests`
- AC-5 (REQ-7): a render with a changed percentage writes the record in the cache folder, an unchanged one does not rewrite it, the home folder is the same tree before and after, and the lane reader reads the new record.
  verify: `cargo test --offline --locked --test statusline_home_state usage_record_`
- AC-6 (REQ-8): with a fake Orca script and the hook variables the script starts once with the stdin bytes, with no `rate_limits` or no variable or a non-executable file it does not start, and the render returns before a slow script ends.
  verify: `cargo test --offline --locked --test statusline_counts_refresh orca_pass_through_`
- AC-7 (REQ-9): the generated block carries the three values, a block the user wrote is reported and kept, and a block that still runs the retired command path is reported.
  verify: `cargo test --offline --locked --test claude_settings_cli status_line_block_`
- AC-8 (REQ-10, REQ-11): a payload with one mistyped field, an unknown field and a duplicated key renders every other item, a non-object renders empty, and none of the removed fields is read.
  verify: `cargo test --offline --locked --lib statusline::input::tolerant_reader_tests`
- AC-9 (REQ-1, REQ-12): every form of the Layouts section is exactly the string written there and changes at its boundary; a fake `git` and a fake Orca script count the processes at most one each and the files written are the cache and the record only.
  verify: `cargo test --offline --locked --test statusline_cards line_three_ladder_`

### Layouts

Widths as `unicode-width` counts them (probe 2026-11-22). Fixture values only.

| case | form | text | cols |
|---|---|---|---|
| subscriber, 5H over 70 | Z0 widest | `🪫 CW 74% \| 🪫 5H 85% (1h 10m) \| 🔋 7D 21% (6d 21h) \| 🤖 2.1.330 \| 🗿 4.2647.16` | 79 |
| | Z2 no versions | `🪫 CW 74% \| 🪫 5H 85% (1h 10m) \| 🔋 7D 21% (6d 21h)` | 51 |
| | Z3 no countdowns | `🪫 CW 74% \| 🪫 5H 85% \| 🔋 7D 21%` | 33 |
| | Z4 narrowest | `🪫 CW 74% \| 🔋 7D 21%` | 21 |
| subscriber, 5H hidden | Z0 | `🪫 CW 74% \| 🔋 7D 21% (6d 21h) \| 🤖 2.1.330 \| 🗿 4.2647.16` | 58 |
| | Z2 | `🪫 CW 74% \| 🔋 7D 21% (6d 21h)` | 30 |
| | Z3, Z4 | `🪫 CW 74% \| 🔋 7D 21%` | 21 |
| update pending | Z0 | `🪫 CW 74% \| 🪫 5H 85% (1h 10m) \| 🔋 7D 21% (6d 21h) \| 🤖 2.1.330 \| 🗿 4.2647.15 -> 4.2647.16` | 92 |
| | Z1 no Claude Code version | `🪫 CW 74% \| 🪫 5H 85% (1h 10m) \| 🔋 7D 21% (6d 21h) \| 🗿 4.2647.15 -> 4.2647.16` | 79 |
| API key user | Z0 | `🪫 CW 41% \| 💰 3.87 \| 🤖 2.1.330 \| 🗿 4.2647.16` | 47 |
| | Z2 | `🪫 CW 41% \| 💰 3.87` | 19 |
| | Z4 | `🪫 CW 41%` | 9 |

The forms drop in the order: the Claude Code version (Z1, keeping an update arrow), the MoAI version (Z2, the arrow with it), the countdowns (Z3), the five-hour window (Z4). Cost and the seven-day window are the last items to go.

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-limits/**` (this slice); the code slice writes `src/statusline/{input,state,render,mod,config}.rs` (and removes `godecode.rs`), `src/usage_limit.rs` with its two readers, the settings generator (`src/status_line.rs`, `src/claude.rs`, `src/claude_settings_merge.rs`), the tests and goldens, and the statusline section of the product document.
- forbidden: the network (the Orca script posts to `127.0.0.1` itself; this process opens no socket), a write under the user's home folder, the real HOME in tests.

## Plan

One landing per step, red tests first, after the flow line and `where`: (1) the tolerant reader replaces the Go decoder (behavior-visible only for the three golden cases `a13`, `a14`, `a15`, each read before rewriting) and the unseen values stop being collected; (2) the items and forms of line 3 with `CW` as given, `5H`, `7D` countdown, `💰`, versions; (3) the usage record moves to the cache folder and its two readers are re-pointed in the same commit; (4) the Orca pass-through; (5) the settings generator; the project settings of this checkout still run the retired command path (the leader's finding; an untracked file, so no commit carries it) and are rewritten once by running the generator, noted in the Result; (6) documents (parent REQ-12).

Step 3 starts by listing every reader of `usage-limits.json` (`rg` over `src`, the lane screen and the `usage-limit` operation are the two known) and ends with the lane reader reading the record in the cache folder in a test.

## Risks

1. **The usage record is read by lanes.** The lane screen and a `usage-limit` operation read `usage-limits.json` to renew leases before a limit stops a lane (#2668). Moving the record from the retired home folder to the cache folder under `$TMPDIR` changes where they look; a lane started with another `TMPDIR` would see no record. The limits are account-wide, so any session's record is right, but the folder must be the same for all sessions of the user: the code slice states the folder rule and the reader's fallback (no record means no `near` state, as today when the file is missing).
2. **Dropping case folding and duplicate-key rules changes three golden cases.** The Go decoder folded key case and took the last of a duplicated key; the tolerant reader takes the last value of a duplicated key and matches keys exactly. `a13`, `a14` and `a15` change and are read one by one in the code slice.
3. **The Orca script is not ours.** Its interface is the one read on 2026-11-22: it reads the whole stdin, does nothing without `rate_limits` or without the three hook variables, posts to `127.0.0.1` with a token and throttles itself to one post per fifteen seconds. A change of that script changes the pass-through; this SPEC does not copy its logic, only starts it. One extra detached process per render with `rate_limits` is the cost.
4. **`5H` may disappear from Claude Code.** One secondary report says the five-hour cap was lifted for some top-plan accounts; the official page still lists it. The item is drawn only when sent, so nothing is needed on a change.
5. **A window after its reset.** The host drops a window once its `resets_at` passed, so `7D (0m)` does not occur; the item simply disappears until the host sends it again. A gateway user whose payload has only `rate_limits.spend_limit` sees `💰` (REQ-5) and no window.
6. **The cost is an estimate.** The host says it "may differ from your actual bill" and resets to zero on `/clear`; the doc says so next to the item.

## Decisions

| Id | Decision | Basis | Undo |
|---|---|---|---|
| D-1 | `CW` shows `context_window.used_percentage` as the host gives it; when it is null the fallback of today (the sum of `current_usage` over the window size) stays, without the 85% rescale | user pick 2026-11-22; docs: "use `used_percentage` for the simplest accurate context state" | restore the rescale |
| D-2 | `5H` only strictly above 70 percent; its icon is `🪫` then | user pick (note 5935); the battery icon switches at 70 | one constant |
| D-3 | cost only when neither window is present, including a payload with only a spend limit | user pick "when the host sends no rate limits" (note 5935), read as no usage window; gateway users have a money limit, not a window | show cost with a spend limit too |
| D-4 | the countdown loses `resets in`: `(6d 21h)`, `(2h 10m)`, `(35m)` | the user's own line example in note 5942 | restore the words |
| D-5 | the usage record moves to the cache folder and is written only on change; its readers are re-pointed in the same commit | the home folder `~/.moai-cli` is retired and "reads only" is replaced by the cache rule (note 5935); lanes still need the record | drop the record and the lanes' near-limit check |
| D-6 | the Orca pass-through starts `~/.orca/agent-hooks/claude-statusline.sh` (the file Orca installs) only when it is an executable file, the three hook variables (`ORCA_AGENT_HOOK_PORT`, `ORCA_AGENT_HOOK_TOKEN`, `ORCA_PANE_KEY`) or the endpoint file variable are set, and the stdin has `rate_limits`; stdin is the received bytes; stdout and stderr are discarded; the render never waits | user pick: "moai passes the stdin it receives to Orca's wrapper, loopback" (note 5935); the script read 2026-11-22 | remove the call |
| D-7 | the tolerant reader (`serde_json` values, a field at a time) replaces `godecode.rs` | the doc's input fields keep growing; one mistyped field must not blank the line | none planned |
| D-8 | the generated block carries `refreshInterval: 30` and `padding` from the status line configuration | user pick (note 5935); docs: `refreshInterval` re-runs the command every N seconds, minimum 1 | change the number |

## Result

SPEC 만이에요. 돌린 것: 줄 3 모양 너비 측정(`unicode-width` 프로브), Orca 래퍼 스크립트 읽기(2026-11-22, 코드 변경 없음). 돌리지 않은 것: 레인의 사용량 파일 읽는 곳 전수 조사(계획 3단계의 시작에 해요), 시험들.
