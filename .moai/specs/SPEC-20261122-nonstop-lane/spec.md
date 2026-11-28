---
id: SPEC-20261122-nonstop-lane
title: "레인을 멈추지 않고 돌리는 선택: 저장한 선택 하나, 만든 루프 프롬프트 하나, 원장 감시는 `moai card wait`"
status: draft
revision: 6
created: 2026-11-22
---

# 레인을 멈추지 않고 돌리는 선택

## Intent

2026-11-22 사용자 요청: 레인을 멈추지 않고 돌리는 길을 moai 에 들어 있는 선택으로 내놓는 일이에요. 선택은 추천 항목을 앞에 둔 질문으로 받고, 프로필에 저장하고, 다시 묻지 않고 명령으로 바꿔요. non-stop 이면 레인은 Claude Code 의 self-paced `/loop` 를 moai 가 만든 프롬프트로 돌려요. Claude Code 의 `/loop`, Monitor, 그리고 사용자가 고를 때의 `/goal` 만 쓰고, 기록은 원장이며 Monitor 밖에서는 폴링하지 않아요.

개정 5 가 필요한 까닭은 개정 3 뒤에 바뀐 것이 다섯 가지이기 때문이에요.

1. 레인 시작 말이 바뀌었어요. `/moai agent` 는 모드만 바꾸고, 일은 교차 세션 메시지로 와요(2026-11-25 사용자 결정, `skills/moai/references/flows/team-messages.md`).
2. 등록된 리더가 있으면 레인은 다음 카드를 스스로 잡지 않아요(`flows/next-flow`, `workflows/work.md`). 개정 3 의 프롬프트는 READY 뒤에 바로 다음 카드를 잡게 해요.
3. 호스트의 루프 만료가 7일에서 30일로 바뀌었어요(Claude Code 2.1.330).
4. 원장 감시 명령 `moai card wait` 가 이미 있어요. 따로 만들 `lane-run watch` 가 필요 없어요.
5. 2번 조각(`show`, `set`, `prompt`, `goal`)은 착지했고(머지 `6f4117db80`), 3번 조각(자체 `watch`, 개정 4)은 브랜치 `spec/20261122-nonstop-lane-3`(head `5499d32b704e`)에 보류된 채 착지하지 않았어요.

개정 6 은 개정 5 가 열어 둔 선택의 답을 반영해요. 선택은 `/moai project` 와 레인 시작 두 곳에서 물어요(사용자 선택, 노트 10281). 나머지는 개정 5 의 기본값이에요.

### 읽은 공식 문서

`https://code.claude.com/docs/en/<이름>.md` 를 2026-11-26T09:11:03Z 에 받아 읽었어요.

| 문서 | sha256 |
| --- | --- |
| scheduled-tasks | `8bede12eb3c0d3253205d9cdf7324bbdc015169a3c42c1914e58c9a7d9831c2b` |
| tools-reference | `fed5908233528b94f388ac06eef60268158aa6ec775b5943d7c1c26fc3bfb2f2` |
| goal | `a967e8db746060a3efbe6496b088f7cc5185785b172c6167f01e9d74341bb5d5` |
| cross-session-messaging | `d68cba02a8597a58825ae954e0c25265c0108e3926c13fb5b0c72028b70a37f0` |

이 SPEC 이 기대는 사실이에요.

- 간격 없는 `/loop` 는 self-paced 예요. 매 회차 뒤에 Claude 가 `ScheduleWakeup` 으로 1분에서 1시간 사이의 다음 시각을 고르고, `stop: true` 로 루프를 끝내요. 다음 회차를 잡지 않고 끝난 회차에서도 루프가 멈출 수 있어요.
- 반복 작업은 만든 지 30일 뒤에 만료돼요. 2.1.330 전에는 7일이에요.
- self-paced `/loop` 는 `--resume`, `--continue` 로 되살아나지 않아요. 돌던 백그라운드 Bash 와 monitor 도 다시 시작되지 않아요.
- `CLAUDE_CODE_DISABLE_CRON=1` 은 `/loop` 와 `CronCreate` 를 쓸 수 없게 해요.
- self-paced 간격은 Amazon Bedrock, Claude Platform on AWS, Google Cloud's Agent Platform, Microsoft Foundry, 또는 feature flag 가져오기가 꺼진 곳에서 2.1.248 이상이 필요해요.
- Monitor 의 마감은 기본 5분, 최대 30분이에요(`-p` 단일 프롬프트 실행은 최대 10분). 마감에 알림이 한 번 오고, Bash 와 같은 권한 규칙을 써요. 위 제공자들이나 `DISABLE_TELEMETRY`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` 을 켠 곳에서는 2.1.310 이상, Windows 는 Git Bash 가 필요해요.
- `/goal` 은 매 턴 뒤에 작은 모델이 조건을 보는 세션 한정 Stop hook 이에요. 조건은 4,000자까지이고 추가 토큰이 청구되며, 백그라운드 작업이 도는 동안은 평가가 미뤄져요. `disableAllHooks` 나 `allowManagedHooksOnly` 면 쓸 수 없어요.
- 쉬는 세션에 교차 세션 메시지가 오면 새 턴이 시작돼요. 메시지는 보류됐다가 버려질 수 있고, 반복은 속도 제한과 중복 제거를 받아요.

GLM 실행기에서 `ScheduleWakeup` 과 Monitor 가 같게 동작하는지는 실행해 보지 않았어요(확인하지 못함).

## Docs

착지한 뒤 사용자가 읽는 내용이에요.

- `/moai project` 가 Claude Code 프로젝트에서, 저장된 선택이 없으면 한 번 물어요: 레인을 non-stop 으로 돌릴지, 카드 하나씩 돌릴지. 아직 정한 적이 없고 사람이 창에 있으면 `/moai agent` 를 시작할 때도 카드를 잡기 전에 한 번 물어요. 답은 프로젝트 값으로 저장돼요. `moai lane-run set nonstop|single [--lane <name>]` 은 다시 묻지 않고 바꾸고, `moai lane-run show [--lane <name>]` 은 지금 적용되는 선택과 그 까닭을 알려 줘요.
- non-stop 이면 `moai lane-run prompt --lane <name>` 이 Claude Code 의 `/loop` 에 넣을 글을 찍어요. 리더가 있으면 레인은 리더의 배정을 기다리고, 없으면 원장에서 다음 카드를 잡아요. 사람이 보지 않는 동안 토큰을 쓰고 도구를 돌리므로, auto mode 나 moai 명령의 allow 규칙이 필요해요. 루프는 30일 뒤에 끝나고(2.1.330 전에는 7일) resume 으로 되살아나지 않아요.
- `CLAUDE_CODE_DISABLE_CRON=1` 이면 카드 하나씩으로 바뀌고 `show` 가 그렇게 말해요.
- `moai lane-run goal --card <name>` 은 카드 하나의 `/goal` 조건을 찍어요. `/goal` 은 사용자가 직접 입력해요.

## Interview

- Q1: 무엇을 만드나요? -> A: 레인을 멈추지 않고 돌리는 내장 선택이에요. 추천 항목을 앞에 둔 선택으로 묻고, 프로필에 저장하고, 명령으로 바꾸고, 레인마다 따로 둘 수 있어요. Claude Code 의 `/loop`, Monitor, 선택적 `/goal` 만 써요(2026-11-22 사용자 요청, 카드 본문).
- Q2: 레인은 어떻게 시작하고 일은 어떻게 받나요? -> A: `/moai agent` 는 모드만 바꾸고 지시 글을 받지 않아요. 일은 교차 세션 메시지로 와요(2026-11-25 사용자 결정).
- Q3: 카드 글에 남은 `/moai work` 는 어떻게 읽나요? -> A: `/moai agent` 로 읽어요(2026-11-26 사용자 선택, 리더 기록 노트 10152).
- Q4: 선택은 어디에 저장하나요? -> A: 프로필 최상위 `lane_run` 키예요(2026-11-22 리더 결정, 노트 6574, D-2).
- Q5: agent 모드에서 선택이 없으면 자동으로 non-stop 인가요? -> A: 아니에요. 자동 non-stop 기본값은 두지 않고, 답이 없으면 `single` 이에요(2026-11-22 리더 결정, 노트 6069).
- Q6: 선택은 어디서 묻나요? -> A: `/moai project` 와 레인 시작이에요. `/moai project` 에서 정하고, 아직 정한 적이 없고 사람이 창에 있으면 `/moai agent` 를 시작할 때도 한 번 물어요(2026-11-26 사용자 직접 선택, 리더 기록 노트 10281).
- Q7: 나머지 선택은요? -> A: 개정 5 의 기본값이에요: 카드 기준 1 은 다시 적고, 사라진 메시지는 리더의 다음 회차가 되살리고, 감시는 `moai card wait`, 리더를 본 레인은 3600초마다 깨고, 리더를 본 세션은 `free` 를 만료로 읽어요(2026-11-26 리더 결정, 노트 10281).

## Non-goals

- moai 가 `/loop` 나 `/goal` 을 시작하는 일. moai 는 글만 찍어요.
- moai 의 스케줄러, 데몬, 상주 프로세스, 그리고 Monitor 밖의 폴링.
- 새 감시 명령(`lane-run watch`). `moai card wait` 를 그대로 써요.
- 클라우드 스케줄이나 Desktop scheduled task. 루프는 세션 한정이고 만료는 알려 줄 뿐 우회하지 않아요.
- 권한 모드를 바꾸는 일.
- 리더 흐름, 메시지 표, 리더 timer 규칙(`next-flow`, `team-messages`, `internal/loop`)을 바꾸는 일.
- `card next` 의 카드 고르는 방식, `card wait` 의 동작, 프로필 읽기(`src/profile.rs`)를 바꾸는 일.
- 예전 `lane-loop-line` 문구.

## Requirements

- REQ-1: When `/moai project` finishes applying a profile on the Claude Code host and the profile holds no `lane_run` pick, the skill shall ask once, as a selection with the recommended option first, whether lanes run non-stop or one card at a time and save the answer with `moai lane-run set` · failure: none — a saved pick is not asked again, no answer writes nothing, and an update never asks.
- REQ-2: When the saved pick is changed with the `lane-run set` command, the lane-run command shall write the new value for the project or for the named lane without asking again and leave every other profile key as it was · failure: none — a value other than `nonstop` or `single` refuses before any write.
- REQ-3: When the effective pick of a lane is read, the lane-run command shall give the lane's own value before the project's value and report `unset` when neither exists · failure: none — an unreadable profile reports the reason and no pick.
- REQ-4: When the effective pick of a lane is `nonstop`, the lane-run prompt command shall print one plain-text prompt for the self-paced `/loop` that names the lane, the leader check at every wake, the rule to keep the current card, the wait for a dispatch with a leader, `moai card next` without one, the rule that a verdict or a repair request comes first, the blocked-note rule and the rule never to end a turn with neither a card in progress, a watch nor a scheduled wake · failure: none — a lane name that is not accepted refuses before printing.
- REQ-5: When the pick is `single`, the lane-run prompt command shall print no loop prompt and say that no loop is started · failure: none — the answer names the pick it read.
- REQ-6: When the generated prompt is followed and the lane has no card it can work, the prompt shall make the lane print one idle line and then end the loop with `ScheduleWakeup` and `stop: true` where the session has seen no leader, or keep one wake pending at 3600 seconds where it has seen one, until the leader's session-end notice or the user's word · failure: none — the idle line is printed once for the same state.
- REQ-7: While a lane waits for a verdict or an answer, the generated prompt shall name `moai card wait`, run by Monitor with one `--card` for each card the session holds and each card in review that the session authored, as its only watch · failure: none — an unknown card is refused before the wait, and with no card to watch no watch is started.
- REQ-8: When the generated prompt starts the watch, the prompt shall give `moai card wait` a timeout shorter than the Monitor timeout and carry the last note mark into the next start with `--since` · failure: none — a watch that ended without an answer starts again as a first start followed by a read of each watched card.
- REQ-9: When a host fact makes the non-stop pick impossible, the lane-run show command shall name the fact, print the version floors and report the effective pick as `single` · failure: none — the facts are `CLAUDE_CODE_DISABLE_CRON` set to a true value and a loop that Claude Code cannot be asked for, and a fact that cannot be read is reported as unread and never as met.
- REQ-10: When the user picks `/goal` for a card, the lane-run goal command shall print one condition of at most 4,000 characters that names the card, one measurable end state, the check that proves it and a turn limit · failure: none — the command prints the text only and never starts a goal.
- REQ-11: When a session in the agent mode whose effective pick is `nonstop` starts a turn with no loop pending, the skill shall start the host's self-paced loop with the generated prompt, or print the prompt for the user to paste, and say that the loop is gone after a resume and expires 30 days after it starts · failure: none — a lane whose effective pick is `single` starts nothing.
- REQ-12: When the generated prompt, the help text or any answer of the lane-run commands is printed, the text shall hold no card number, no key and no path outside the project · failure: none — the card is a placeholder and the lane name comes from the caller.
- REQ-13: When `/moai project` builds a plan input for a profile that holds `lane_run`, the skill shall carry the saved `lane_run` value in that input unchanged · failure: none — a profile without the key gets none from the input.
- REQ-14: IF a wake of the generated prompt reads the leader entry as `free` after the session has seen a registered leader or received a dispatch, or reads any answer other than `held` or `free`, THEN the prompt shall make the lane take no new card and continue only the cards it holds · failure: none — one ledger note on the card worked last names the answer.
- REQ-15: When a session enters the agent mode with a person at the window on the Claude Code host and `moai lane-run show` reads its pick as `unset`, the skill shall ask the selection of REQ-1 once, before it takes any card, and save the answer as the project's pick with `moai lane-run set` · failure: none — with no person at the window or no answer nothing is written and the lane runs as `single`.

## Acceptance

- AC-1 (REQ-1, REQ-11, REQ-13, REQ-15): the project workflow carries the selection with the recommended option first, the rule that no answer writes nothing and the carried `lane_run` input, the work workflow carries the selection asked at agent-mode start before any card only with a person at the window and an `unset` pick, saved as the project's pick, and the agent-mode prompt rule, and the source manifest lists the changed files · verify: `cargo test --locked --offline --test lane_run_cli skill_text_` and `cargo test --locked --offline --test source_manifest_facts`
- AC-2 (REQ-2, REQ-3, REQ-13): in a temporary project, `lane-run set nonstop` and `lane-run set single --lane <name>` write the project value and the lane value and no other key changes, the effective pick is the lane's before the project's, `unset` is reported when neither exists, a wrong value or a project without a profile refuses with nothing written, and after a `set` a project plan whose input carries the saved value leaves `lane_run` as it was through apply · verify: `cargo test --locked --offline --test lane_run_cli pick_`
- AC-3 (REQ-4, REQ-6, REQ-12, REQ-14): in a temporary project and a temporary HOME, with `nonstop` the prompt names the lane, `moai session-register --show --name leader`, the lapsed-entry rule, the wait for a dispatch, `moai card next`, the blocked-note rule, the never-end-a-turn rule and both idle rules with `stop: true` and 3600 seconds, and holds no card number, no key-shaped text and no absolute path; with `single` no prompt is printed · verify: `cargo test --locked --offline --test lane_run_cli prompt_`
- AC-4 (REQ-7, REQ-8): the prompt names `moai card wait` with `--not-author`, `--timeout 25m`, a Monitor timeout of 30 minutes and `--since`, and names no `lane-run watch`; the tests of the wait command pass · verify: `cargo test --locked --offline --test lane_run_cli prompt_` and `cargo test --locked --offline --test card card_wait_cli::`
- AC-5 (REQ-9): with `CLAUDE_CODE_DISABLE_CRON=1` the show answer names that fact and the effective pick reads `single` although the saved pick is `nonstop`; with it unset the saved pick stands, a fact that cannot be read is reported unread, and the floors name 2.1.248, 2.1.310 and 2.1.330 with the condition of each · verify: `cargo test --locked --offline --test lane_run_cli facts_`
- AC-6 (REQ-10): the goal text is at most 4,000 characters for the longest accepted card name, carries the end state, the check and the turn limit, and the command starts nothing · verify: `cargo test --locked --offline --test lane_run_cli goal_`
- AC-7 (REQ-12): the help text and the operations document of the four commands carry no card number, name no `watch` action and give examples that hold the example rules · verify: `cargo test --locked --offline --bin moai-cli every_command_has_examples`, `cargo test --locked --offline --bin moai-cli help_table_names_every_command_once`, `cargo test --locked --offline --test help_layout` and `cargo test --locked --offline --test lane_run_cli help_`

## Design

### 프롬프트 초안

구현은 아래 글에서 시작해요. 테스트가 고정하는 것은 AC-3, AC-4 의 항목이에요.

```text
Lane {lane}: keep working without stopping, as a self-paced loop with no interval.
At every wake, first run `moai session-register --show --name leader`. held means a leader is registered. free means there is none, unless this session has already seen a registered leader or received a dispatch: then the entry has lapsed. On a lapsed entry, any other answer or a failure, take no new card, continue only the cards you hold, and write one note that names the answer on the card you worked last.
1. Keep the current card: if a card is in progress, continue it first.
2. With a leader registered: after your READY note, hand the card back, send the one completion message and wait for the leader's dispatch; never take the next card on your own. With no leader: when the ledger shows the previous card done or released, run `moai card next`, take the card it names and read its body and notes whole.
3. Watch your cards: those `moai card next` lists under mine, and each card of `moai card list --state review` whose `moai card show` roles.authors holds the source_session_id that `moai session-current` prints. Keep a Monitor, with its timeout at 30 minutes, that runs `moai card wait --card <card> --timeout 25m --not-author <your session id>`, one --card for each watched card. On a first start, start it and then read each watched card with `moai card show`. When it ends, start it again with --since set to the highest note_id it returned, or to the mark it answered on a timeout; if it ended with no answer, do a first start again.
4. A verdict or a repair request comes first.
5. A decision you need goes into a LANE-BLOCKED kind=decision note on its card; do not ask the user from the loop.
6. never end a turn with neither a card in progress, a watch nor a scheduled wake: call ScheduleWakeup (1200 to 1800 seconds while a Monitor runs, 3600 with nothing to watch).
7. When there is no card you can work (a card that waits for a decision does not count), print one idle line (lane {lane} idle: no card to take) once. If this session has seen no leader, stop the watch and end the loop with ScheduleWakeup and stop: true. Otherwise keep one wake pending at 3600 seconds, and end the loop only on the leader's session-end notice or the user's word.
Read and write the ledger only through moai commands. The loop ends 30 days after it starts (7 days before Claude Code 2.1.330) and is not restored by a resume.
```

2026-11-26 에 확인한 명령 출력이에요: `moai card list --state review` 의 목록은 `.result.cards`, 작성 세션은 `moai card show` 의 `.result.roles.authors`, 세션 id 는 `moai session-current` 의 `.result.source_session_id` 예요. `moai card wait` 는 시간 초과 답에 `mark` 를, 노트 답에 `items[].note_id` 를 줘요.

### 선택을 묻는 자리

- `/moai project`: `moai project-apply` 가 끝난 뒤의 별도 선택이에요. 확인 한 번으로 묶는 표의 행이 아니에요.
- `/moai agent` 시작: `moai lane-run show --lane <name>` 의 `saved` 가 `unset` 일 때만, 카드를 잡기 전에 한 번 물어요. "사람이 창에 있다" 는 호스트가 선택 창을 주는 대화형 세션이고, 모드 말이 다른 세션의 메시지가 아니라 그 세션 자신의 입력으로 왔다는 뜻이에요. `moai cc -a` 의 첫 입력도 여기에 들어가요. `work.md` 의 "카드를 일하는 세션은 선택 창을 열지 않는다" 는 규칙은 그대로이고, 카드를 잡기 전의 이 한 번만 따로 적어요.
- `/moai agent` 시작에서 받은 답은 프로젝트 값으로 저장하고, 프로필 파일이 바뀌었고 커밋되지 않았다고 한 줄로 알려요.
- 추천은 프로필의 흐름이 NeXT-Flow(여러 레인용)이면 non-stop, 아니면 카드 하나씩이에요. 미리 골라 두지 않아요.
- 질문 한 줄에 무인 턴에는 auto mode 나 allow 규칙이 필요하고, 루프는 30일 뒤나 resume 뒤에 없어진다고 적어요.
- 레인별 값은 `moai lane-run set --lane` 으로 둬요.
- 다음 `/moai project` 의 plan 입력은 저장된 `lane_run` 을 그대로 실어요. `src/profile.rs` 의 키 판정은 baseline 에 있는 키가 입력에 없으면 지우거나 conflict 로 보기 때문이에요(코드 읽기, 실행 확인은 AC-2).

## Scope

- write: `.moai/specs/SPEC-20261122-nonstop-lane/spec.md`
- write (조각 S1): `src/ops/lane_run.rs`, `tests/lane_run_cli.rs`, `docs/design-review/appendix/operations/lane_run.md`
- write (조각 S2): `skills/moai/workflows/work.md`, `skills/moai/workflows/project.md`, `tests/lane_run_cli.rs`
- shared files, by line lease only (S2): `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`
- read-only: `src/profile.rs`, `src/main.rs`, `src/help.rs`, `skills/moai/references/flows/**`, `skills/moai/references/internal/loop.md`, `tests/card/card_wait_cli.rs`
- forbidden: 실제 `claude` 나 `orca` 프로세스, 모델 호출, 테스트의 네트워크, 실제 HOME

## Plan

1. 이 개정을 승인받아요.
2. 조각 S1(카드 기준 3, 예전 기준 1 을 다시 적은 것): 빨간 테스트 `prompt_`, `facts_`, `help_` 를 먼저 쓰고, 프롬프트, 버전 하한, `prompt` 답의 안내 문장, operation 문서를 바꿔요. clap 요약 줄은 그대로 둬요.
3. 조각 S2(카드 기준 2): 빨간 테스트 `skill_text_` 와 plan 뒤에도 값이 남는 `pick_` 을 먼저 쓰고, `work.md` 와 `project.md` 를 바꾼 다음 source manifest 를 갱신해요. `tools/repo_scans.json` 의 `reads-skills` 행에 `lane_run_cli` 를 넣어요. 두 workflow 파일을 다른 카드가 예약하고 있으면 기다려요.
4. 조각마다: 가드 테스트의 뮤턴트, `moai check affected --card` 한 번, READY. 기준이 열려 있는 동안의 착지는 `--keep-open` 이에요.

tier 1 이에요: 지우기, 권한, 게이트, HOME 을 바꾸지 않아요.

## Risks

- 만료된 리더 등록(실제로 본 일). 2026-11-26T09:09:40Z 에 `moai session-register --show --name leader` 가 `free` 였어요. 등록의 lease 는 06:06:41Z 에 끝났는데, 그 뒤인 08:51Z 무렵에도 리더는 배정 메시지를 보냈어요(등록 세션과 보낸 세션이 같은지는 확인하지 못했어요). `free` 를 곧바로 "리더 없음" 으로 읽으면 레인이 리더 모르게 카드를 잡아요. REQ-14 는 리더를 본 세션에서는 막지만, `/clear` 뒤처럼 본 기억이 없는 세션에서는 막지 못해요.
- `/clear` 뒤에는 세션 id 가 바뀔 수 있어, 예전 세션이 낸 review 카드는 감시 대상에서 빠져요. 그 카드의 수리 요청은 리더의 메시지에 기대요.
- 교차 세션 메시지는 보류되거나 버려질 수 있어요. 감시는 레인이 잡았거나 낸 카드만 보므로, 새 배정이 사라지면 리더의 다음 회차가 되살려야 해요(D-23).
- 쉬는 레인이 한 시간에 한 번 깨는 데에도 토큰이 들어요(D-24).
- `/moai agent` 시작의 질문은 답할 때까지 그 세션을 붙잡을 수 있고, 그동안 배정 메시지가 처리되지 않을 수 있어요. 창을 띄우고 자리를 뜬 경우를 moai 는 알 수 없어요. 답 없는 선택 창이 호스트에서 어떻게 되는지는 확인하지 못했어요.
- 레인 시작에서 저장하면 추적되는 프로필 파일에 커밋되지 않은 변경이 남아요.
- 사람이 보지 않는 루프는 토큰을 쓰고 도구를 돌려요. 질문이 그렇게 알리고, moai 가 선택을 대신하지 않고, 답이 없으면 `single` 이에요.
- 루프, Monitor, `/goal` 은 몇 주 사이에도 바뀐 호스트 기능이에요. 테스트는 moai 의 글과 명령만 고정해요.
- 세션이 스스로 `/loop` 를 시작할 수 있는지는 호스트 설정에 달려 있고 실행해 보지 않았어요. REQ-11 은 붙여 넣을 글을 찍는 길을 함께 둬요.
- 프로필은 여러 moai 명령이 함께 쓰는 파일이에요. 같은 키를 동시에 쓰면 두 값 중 하나가 사라질 수 있어요.

## Decisions

| ID | Decision | Evidence | Alternative |
| --- | --- | --- | --- |
| D-1 | `lane-run` 한 operation 에 `show`, `set`, `prompt`, `goal` 을 둬요. 쓰기는 `set` 뿐이에요. | 착지한 2번 조각. | `watch` 동작 추가(D-18 로 대체). |
| D-2 | 선택은 프로필 최상위 `lane_run` 의 `default` 와 `lanes` 에 저장해요. `set` 은 임시 파일과 rename 으로 쓰고 다른 키와 순서를 지켜요. | 리더 결정(노트 6574). `work_flow` 는 `flow` 를 요구하고 모르는 필드를 거절해요. | `work_flow.lane_run`. |
| D-3 | 선택이 없으면 `single` 로 동작해요. 자동 non-stop 기본값은 없어요. | 시작한 루프는 토큰과 권한을 쓰므로 사람의 동의가 필요해요. 노트 6069. "agent 모드에서는 묻지 않는다" 는 부분은 D-21 로 대체됐어요. | 레인의 자동 non-stop 기본값. |
| D-4 | 프롬프트는 slash 명령이 아닌 일반 글이고, 카드 번호와 카드 묶음을 싣지 않아요. | 예약된 실행은 모델이 스스로 부를 수 있는 skill 만 돌려요. REQ-12. | `/moai agent …` 를 루프 입력으로 씀. |
| D-6 | 사실은 읽은 것만 말해요. 환경 변수는 읽고, 버전은 읽지 않고 하한을 글로 찍어요. | moai 는 호스트 버전을 읽지 않아요. | 버전 추정. |
| D-7 | `/goal` 은 선택 사항이고 카드마다 사용자가 골라요. 기본값이 아니에요. | 별도 모델의 판단과 토큰이 더 들어요. | 루프에 항상 붙임. |
| D-8 | 예전 `lane-loop-line` 은 제거 카드의 일로 남겨요. | 범위 밖. | 함께 제거. |
| D-9 | 레인 이름은 호출자가 넘긴 `session-register` 이름이에요. | moai 는 이름을 만들지 않아요. | 이름 추론. |
| D-13 | `set` 은 프로필의 파일 모드를 지켜요. | 2번 조각 감사. | 0600 으로 씀. |
| D-18 | 감시는 `moai card wait` 를 써요. 보류된 `lane-run watch` 는 버려요. 개정 4 의 D-14~D-17 은 착지하지 않았고 이 개정은 채택하지 않아요. D-5, D-11 은 대체됐어요. | `card wait` 가 설치본에 있어요(help 읽음). 보류본(`5499d32b704e` 코드 읽기)은 시작할 때 이미 있는 노트 뒤에서 출발해 두 실행 사이의 노트를 건너뛰고, 읽기 오류 뒤에는 cursor 를 버리고 새 끝에서 다시 시작하며, 프롬프트가 Monitor 마감(기본 5분)을 지정하지 않아요. | 보류 브랜치를 다시 올려 고침. |
| D-19 | 리더가 등록돼 있으면 레인은 배정을 기다려요. | `flows/next-flow`, `workflows/work.md` 8단계. | READY 뒤 바로 `card next`. |
| D-20 | `/goal` 설정 파일 읽기(D-10)와 모양 규칙을 프로필 읽기로 옮기는 일(D-12)은 하지 않아요. | 그 설정이면 호스트에서 `/goal` 을 쓸 수 없어요(goal 문서). D-12 의 규칙은 `lane-run` 안에서 동작해요. | 남은 일로 계속 둠. |
| D-21 | 선택은 `/moai project` 에서 정하고, 아직 정한 적이 없고 사람이 창에 있으면 `/moai agent` 시작에서도 카드를 잡기 전에 한 번 물어요. | 사용자 직접 선택(노트 10281). 2026-11-22 요청의 두 자리를 지켜요. | `/moai project` 에서만 물음. 리더 창이 레인마다 물음. |
| D-22 | `/moai agent` 시작에서 받은 답은 프로젝트 값으로 저장해요. 레인별 값은 `set --lane` 이에요. | 이 레인의 결정이고 되돌릴 수 있어요. 레인 이름은 시작마다 달라질 수 있어, 레인 값으로 저장하면 이름마다 다시 묻게 돼요. | 답을 그 레인의 값으로 저장. |
| D-23 | 사라진 완료·배정 메시지는 리더의 다음 회차가 원장을 읽어 되살려요. 흐름 글은 바꾸지 않아요. | 리더 결정(노트 10281). | 배정을 카드 노트로도 남김. |
| D-24 | 리더를 본 레인이 할 일이 없으면 3600초마다 한 번 깨요. | 리더 결정(노트 10281). | 몇 번 쉰 뒤 멈춤. 바로 멈추고 메시지에 기댐. |
| D-25 | 무인 루프의 `free` 는, 리더를 한 번도 보지 않은 세션에서만 "리더 없음" 이에요. 본 세션은 만료로 읽고 새 카드를 잡지 않아요. | 리더 결정(노트 10281), Risks 의 본 일. | 언제나 `card next`. 언제나 새 카드를 잡지 않음. |

## Result

2번 조각은 착지했어요(머지 `6f4117db80`, head `049c8a7751`).

조각 S1 은 커밋 세 개예요: 테스트 커밋 `41e478a080`, 소스·문서 커밋 `02d82a7246`, 그리고 `goal` help 예시와 operation 문서의 예시·`set` 설명, 새 테스트 두 개를 담은 커밋 `f69362e14c` 예요.

- 소스를 바꾸기 전의 실행은 15개 통과, 10개 실패, exit 101 이었어요(로그 `target/lamech/s1-red.log`). 이때의 테스트 파일은 커밋 `41e478a080` 의 것과 같지 않아요: "With no card to watch, start no watch." 를 보는 확인이 아직 없었어요. `help_of_lane_run_and_its_subcommands_names_no_card_number_and_no_watch_action` 은 이 실행에서도 통과했어요.
- `f69362e14c` 의 새 테스트 `help_of_goal_shows_an_example_that_runs` 와 `help_operations_document_examples_name_all_four_commands_and_run` 은 바꾼 것과 함께 커밋해서 빨간 실행이 없어요.
- 아래의 met 은 2026-11-26 에 head `f69362e14c` 에서 verify 명령 여덟 개를 다시 실행한 결과이고, 모두 exit 0 이에요(로그 `target/lamech/s1-ac2.log`).

AC 별 상태예요.

- AC-1: not_run — 조각 S2 의 일이에요.
- AC-2: not_run — 조각 S2 의 일이에요. `set` 뒤에 project plan·apply 를 거쳐도 값이 남는지 보는 테스트는 아직 쓰지 않았고, 이미 있는 `pick_` 테스트는 여기서 AC 확인으로 실행하지 않았어요.
- AC-3: met — `cargo test --locked --offline --test lane_run_cli prompt_` exit 0, 11개 통과.
- AC-4: met — `cargo test --locked --offline --test lane_run_cli prompt_` exit 0, 11개 통과. `cargo test --locked --offline --test card card_wait_cli::` exit 0, 8개 통과.
- AC-5: met — `cargo test --locked --offline --test lane_run_cli facts_` exit 0, 3개 통과.
- AC-6: met — `cargo test --locked --offline --test lane_run_cli goal_` exit 0, 3개 통과(이름 필터에 `help_of_goal_shows_an_example_that_runs` 도 걸려요). S1 에서 goal 조건을 만드는 코드는 바뀌지 않았고, `f69362e14c` 가 `goal` 의 help 글에 예시 한 줄을 더했어요.
- AC-7: met — verify 명령 네 개가 모두 exit 0: `cargo test --locked --offline --bin moai-cli every_command_has_examples` 1개, `cargo test --locked --offline --bin moai-cli help_table_names_every_command_once` 1개, `cargo test --locked --offline --test help_layout` 11개, `cargo test --locked --offline --test lane_run_cli help_` 4개 통과.

그 밖의 기록이에요.

- operation 문서에는 네 동작(`show`, `set`, `prompt`, `goal`)을 모두 부르는 `sh` 예시 묶음이 하나 있고, `lane-run goal --help` 에는 `Example:` 줄이 있어요. `src/help.rs` 의 `lane-run` 예시는 `show`, `set`, `prompt` 세 줄(1~3줄 한도 안)이고, `goal` 의 예시는 그 명령의 help 글에 있어요. 새 테스트 두 개는 문서의 예시 줄과 `goal` 의 예시 줄을 임시 프로젝트에서 실행해요.
- 가드 테스트의 뮤턴트는 14줄이고 모두 exit 101 이에요(`target/lamech/s1-mutants.txt`). `02d82a7246` 에 든 한 문장("With no card to watch, start no watch.")을 넣기 전의 바이트에서 돌렸고, 그 문장의 뮤턴트는 없어요. 커밋 `f69362e14c` 와 그 테스트 두 개는 뮤턴트를 돌리지 않았어요(tier 1 이라 요구되지 않아요).
- 코드의 프롬프트가 `## Design` 의 초안과 다른 곳은 그 한 문장이에요. 감시할 카드가 없으면 감시를 시작하지 않는다는 REQ-7 의 failure 절 때문에 넣었어요. 이 비교는 `src/ops/lane_run.rs` 의 `prompt` 와 초안을 읽어서 한 것이에요.
- 실행하지 않은 것: 명령이 찍은 프롬프트와 초안의 diff(HOME 을 설정하는 추가 확인이어서 워크트리 가드가 거부했어요), Windows, Linux, GLM 호스트의 동작이에요.
- 영향 범위 확인(`moai check affected`)의 receipt 는 여기가 아니라 READY 노트에 남겨요.
- 범위 밖이라 남은 일: `src/lane_loop.rs:181`, `docs/design-review/host-session-events-2026-11-09.md` 의 45행과 101행은 아직 루프가 7일 뒤에 만료된다고 적고 있어요.
