---
id: SPEC-20261122-settings-write-check
title: "A settings write observes a value outside the documented shape and a rewrite that loses content before anything refuses them"
status: draft
revision: 4
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# A settings write observes a value outside the documented shape and a rewrite that loses content before anything refuses them

## Interview

- Q1: 설정 문서 조사 보고서의 항목들을 어떻게 다루나요? → A: 모든 항목을 카드로 만들어요(2026-11-22 사용자 선택, 이 작업의 카드 본문에 기록).
- Q2: 모양이 틀린 값과 내용이 사라지는 다시 쓰기를 착지부터 막나요? → A: 아니요. 처음에는 알림과 기록만 두고, 막기로 바꾸는 일은 기록한 결정으로 해요(리더 결정 카드 노트 9968, 2026-11-26: 사용자 위임은 사용자가 직접 정한 "새 게이트는 observe 로 시작"을 뒤집지 못한다는 판단). 바로 막는 예외를 둘지는 사용자에게 물어요(답 없음).
- Q3: 검사 위치, 기능 스위치, doctor 확인, 추가 요구는요? → A: Decisions D-1, D-3, D-4, D-5 의 기본값 A 예요(같은 노트, 리더 결정이고 사용자의 직접 선택은 아니에요).

## Intent

설정 파일에 모양이 틀린 값 하나를 쓰면 Claude Code 가 그 파일을 통째로 읽지 않아요: "These settings did not load, so nothing in them is in effect, including hooks, permission rules and sandbox settings"(settings.md, `-p` 실행). 별도 쓰기 경로를 새로 만들지 않고 기존 프로젝트·로컬 설정 쓰기(`claude-settings-plan`/`-apply`, `hooks-plan`/`-apply`, 상태줄 적용이 함께 쓰는 `src/claude.rs` 의 계획·적용 길) 안에 검사를 넣어요. 사용자 범위 출력 스타일 쓰기(`src/home_style.rs`)와 `go-hooks-apply` 의 첫 단계는 이 길을 지나지 않아 범위 밖이에요.

이 SPEC 의 새 검사는 모두 observe 로 시작해요(`internal/sdd` 규칙 11): 계획과 적용의 답에 "막았을 것"(`would_refuse`)을 알리고, 파일을 바꾼 적용마다 한 줄을 기록해요. 계획, 계획 digest, 쓰는 내용은 지금과 같아요. 이미 있는 거절은 그대로예요. 막기로 바꾸는 조건과 시점은 D-6 이에요. 쓴 뒤의 `claude doctor` 확인과 되돌림은 실제 출력을 관찰한 뒤의 자식 SPEC 이 해요(D-4).

2026-11-26 에 직접 받아 읽은 문서(`https://code.claude.com/docs/en/<이름>`, 받은 시각과 sha256):

| 문서 | 받은 시각 | sha256 | 쓴 사실 |
| --- | --- | --- | --- |
| settings.md | 03:58:46Z | 7587a3251756cd151948718b70fff20fd26e3a4b545858d172d0941d57ab8f10 | 스키마가 거부한 값이 있는 파일은 통째로 안 읽힘, 개별 규칙 오류는 건너뜀 |
| settings-reference.md | 03:59:01Z | f87b28f014d323f1d99655f4dd8f329693e8295eec0f1cecf8af0bcc534a5249 | `statusLine`·`outputStyle` 모양, 프로젝트·로컬이 무시하는 env 이름 |
| permissions.md | 03:58:46Z | 687a80f8ae55138a03648fbd54fa1c109291656a512a8ed28c1c3f2825460615 | `Write`·`Glob` 류의 경로 규칙은 받되 맞지 않음, 하위 명령 앞의 `*` |
| env-vars.md | 03:58:46Z | 30cd22fe95a40ced8d175d368d5e915441014255f40965e681d2a3dcbdff317e | env 값은 셸이 처리하지 않고 쓴 그대로 |
| agent-teams.md | 03:58:46Z | 742a5e15697d90b09c48f85b83c61a1acf1d5fe88680c1bf36cd69ad9d5b5c9a | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` 이 팀을 켬 |
| errors.md | 04:01:40Z | e4d74f5e0dc5453ab6b522faabb541e3062722a6ff93912a2fb1cf3944a7a83c | 와일드카드 경고의 판정 예 |

받은 팁(41c16e60ac)의 코드에서 확인한 것:

- 호출자 설정 쓰기는 관리 키 목록(`src/claude.rs` `MANAGED_SETTINGS`), env 항목 거절(`env_entry_refusal`), 권한 규칙·모드 거절, 3중 병합(`plan_settings`)을 이미 가져요.
- 반복 키와 달라지는 숫자 표기를 찾는 검사는 `src/go_settings_clean.rs` 에만 있어요(`repeats_a_key`, `numbers_that_change`).
- 막았을 것을 한 줄씩 남기는 길은 이미 있어요: 훅 엔진의 observe 모드가 `src/hooks.rs` `append_receipt_line` 으로 `would_block` 줄을 써요.
- 이전 설정 병합(`src/claude_settings_merge.rs`)은 `type` 이 `command` 인 짧은 상태줄만 후보로 내고 `padding` 같은 나머지 항목의 모양은 보지 않아요. 팀·Code Mode env 이름을 가리는 코드는 `src` 에서 찾지 못했어요.
- `statusLine`·`outputStyle` 의 문서 모양을 계획 때 보는지는 RED 테스트가 확인해요.

## Non-goals

- 새 검사를 막기로 바꾸는 일: D-6 의 관찰 뒤 기록한 결정과 개정한 SPEC 이 해요.
- `claude doctor` 실행, 거부된 쓰기의 되돌림, 제한된 실행기: 실제 출력·종료 코드·읽기 전용 동작을 관찰한 뒤의 자식 SPEC.
- 커밋되는 설정 파일의 자체 점검(env 항목, 작업 폴더 밖 allow, 훅 끄기, sandbox 끄기, 맨 `Edit` allow): `docs/later.md` 한 줄로 남겨요.
- 사용자 범위 설정 쓰기(`crossSessionInbound` 기본값 포함): `docs/later.md` 한 줄로 남겨요.
- 모든 쓰기의 백업, 목록 병합, `/status` 안내 줄, Go 훅 제거 단계(`go-hooks-apply`)의 다시 쓰기 보호, 사용자 출력 스타일 쓰기: 후속.
- Windows·텔레메트리·로그인 env 이름: 무시 목록 가운데 저장 위치 이름 묶음만 알려요.
- 이미 있는 미리보기가 `current`/`replace_with` 로 설정 값을 그대로 보여 주는 동작(`src/claude.rs`)은 바꾸지 않고 `docs/later.md` 한 줄로 남겨요.
- 관리 설정 파일과 `~/.claude.json`.

## Requirements

- REQ-1: When a plan would add or change `statusLine` or `outputStyle` to a value outside its documented shape, the project settings writer shall list the key path and the broken rule under `would_refuse` in the plan and apply answers and leave the plan, its digest and the write unchanged (the shape of `statusLine` is an object whose `type` is `command` with a non-empty string `command` and, where present, a number `padding`, a number `refreshInterval` of at least 1 and a boolean `hideVimModeIndicator`; the shape of `outputStyle` is a non-empty string; the project settings writer is the plan and apply path behind `claude-settings-plan` and `claude-settings-apply`, `hooks-plan` and `hooks-apply`, and the status line apply, which writes the project and local settings files; an answer without an entry carries no `would_refuse` field). · failure: none — observation only, so the write goes on as it does today.
- REQ-2: When a plan would add or change a value that an observed rule names, the project settings writer shall list the key path and the rule under `would_refuse` in the plan and apply answers and leave the plan, its digest and the write unchanged (the observed rules are a `Write`, `NotebookEdit`, `MultiEdit`, `Glob` or `Grep` permission rule with a path other than `*`; a `Bash` allow rule whose `*` comes before a later word that determines the command; an `env` value holding `$` or starting with `~`; an `env` name of the storage-location family, which is `CLAUDE_CONFIG_DIR`, `CLAUDE_CODE_TMPDIR`, `HOME`, `TMPDIR`, `TMP`, `TEMP` or an `XDG_` name; and `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` or `CLAUDE_CODE_EXPERIMENTAL_CODE_MODE` with a value other than `0` or empty). · failure: none — observation only.
- REQ-3: The project settings writer shall refuse a request that names a top-level key outside the keys a caller may set (`attribution`, `fastModePerSessionOptIn`, `includeGitInstructions`, `disableAllHooks`, `sandbox` and a caller-supplied `hooks` among them), a widening `permissions.defaultMode`, a non-string `env` value or a permission rule that is not a non-empty string, and leave the file byte-identical.
- REQ-4: The project settings writer shall apply REQ-1 and REQ-2 only to a value the plan would add or change, never to one the file already holds or the plan leaves as it is.
- REQ-5: When the earlier-settings merge holds a feature switch, the merge preview shall leave that one `env` entry in the backup with its reason and keep every other candidate (a feature switch is an `env` entry that turns on agent teams or Code Mode and that the current file does not hold). · failure: none — the left-out value stays in the backup file.
- REQ-6: When a plan or apply would rewrite a settings file that repeats an object key or holds a number the rewrite would spell differently, the project settings writer shall list the settings file and either the affected top-level key or the repeated-key rule under `would_refuse` and leave the plan, its digest and the write unchanged. · failure: none — observation only, so the rewrite goes on as it does today.
- REQ-7: While no replacement is confirmed, the project settings writer shall keep every key the plan does not name with its value and order, and leave a key the user changed as it is.
- REQ-8: The project settings writer shall put no settings value into a `would_refuse` entry, into an observation line or into a left-out reason of REQ-5.
- REQ-9: When an apply changes a settings file, the project settings writer shall append one line to the settings observation log with the time, the settings file and the rule id and key path of each `would_refuse` entry (the settings observation log is `receipts/settings-write.jsonl` under the state folder of the project the apply runs on, never under HOME; an apply without an entry writes the line with an empty list; a plan writes no line).
- REQ-10 (REQ-9): IF the observation line cannot be appended, THEN the project settings writer shall finish the apply as it would otherwise and say `observation_recorded: false` in the answer.

## Acceptance

- AC-1 (REQ-1): a misshapen `statusLine` or `outputStyle` appears under `would_refuse` with the key path and rule in plan and apply, the plan issues its digest, the apply writes the planned bytes, and a documented valid value gets no `would_refuse` field (temporary HOME) · verify: `cargo test --locked --offline --test settings_write_cli invalid_shapes_are_reported_not_refused`
- AC-2 (REQ-4): a misshapen or listed value the file already holds adds no `would_refuse` entry and changes nothing about another change · verify: `cargo test --locked --offline --test settings_write_cli untouched_user_values_are_not_reported`
- AC-3 (REQ-2): each listed value, `Bash(git * main)` and `Bash(git -C * status *)` among them, appears under `would_refuse`, `Bash(git status *)`, `Bash(ls *)` and `Edit(docs/**)` do not, and the write succeeds · verify: `cargo test --locked --offline --test settings_write_cli questionable_values_are_reported_not_refused`
- AC-4 (REQ-3): a request naming each refused key, a widening mode, a non-string env value or an empty rule is refused and the file stays byte-identical (pins today's behaviour) · verify: `cargo test --locked --offline --test settings_write_cli forbidden_keys_are_never_written`
- AC-5 (REQ-5): the merge leaves out only the feature switch, keeps the other candidates, and carries a feature switch the current file already holds · verify: `cargo test --locked --offline --lib merge_keeps_feature_switches_in_backup`
- AC-6 (REQ-6): a rewrite of a file with a repeated key or a respelled number appears under `would_refuse` in plan and apply and writes the same bytes as today, and a plan that rewrites nothing gets no entry · verify: `cargo test --locked --offline --test settings_write_cli a_rewrite_that_would_lose_content_is_reported`
- AC-7 (REQ-7): a valid write keeps every user key, list entry and their order, and a key the user changed stays as it is · verify: `cargo test --locked --offline --test settings_write_cli a_valid_write_keeps_every_user_key`
- AC-8 (REQ-8): a planted credential-shaped value and a planted plain value in a reported or left-out entry appear in no new answer field and in no observation line · verify: `cargo test --locked --offline --test settings_write_cli no_value_reaches_a_new_answer_or_line`
- AC-9 (REQ-9, REQ-10): each changing apply adds one line, with and without entries; a plan and an unchanged apply add none; with a link at the log path the apply still writes, says `observation_recorded: false` and leaves the link's target as it was · verify: `cargo test --locked --offline --test settings_write_cli every_changing_apply_is_recorded`

## Scope

- write: `.moai/specs/SPEC-20261122-settings-write-check/**`, `src/settings_write.rs`, `src/claude.rs`, `src/claude_settings_merge.rs`, `src/go_settings_clean.rs`, `tests/settings_write_cli.rs`, `docs/rust-product-path.md`, `docs/later.md`
- write: `src/lib.rs`, `tests/areas.toml` (shared file 이라 line-lease)
- read-only: `src/status_line.rs`, `src/hooks.rs`, `src/home_style.rs`, `src/go_hooks.rs`
- forbidden: 테스트가 실제 사용자 설정 폴더를 읽거나 쓰는 일과 실제 `claude` 실행, 기존 거절을 약하게 하는 변경, Python

## Plan

1. RED: `tests/settings_write_cli.rs`(AC-1~4, AC-6~9)와 병합 단위 테스트(AC-5). AC-4 와 AC-7 은 지금 동작을 못박는 것이라 처음부터 초록일 수 있어요.
2. `src/settings_write.rs`(새 모듈, 판정 표는 한 곳)에 모양·관찰 규칙과 보고를 두고 `src/claude.rs` 의 계획·적용 답에 붙여요. 계획 digest 의 입력은 바꾸지 않아요(AC-1~4, AC-8).
3. `go_settings_clean.rs` 의 반복 키·숫자 표기 검사를 같은 모듈에서 부르고 다시 쓰기가 있을 때만 알려요(AC-6, AC-7). 파일을 바꾼 적용 뒤에 `src/hooks.rs` 의 기존 줄 쓰기 `append_receipt_line_in` 을 HOME 없이(`None`) 불러 프로젝트 상태 폴더의 `receipts/settings-write.jsonl` 에 기록해요(AC-9).
4. 병합 후보에서 새로 켜는 기능 스위치만 백업에 남겨요(AC-5).
5. 글(`docs/rust-product-path.md` 의 설정 쓰기 줄)과 `docs/later.md` 세 줄(미리보기의 값 표시, `crossSessionInbound` 기본값, 커밋 설정 자체 점검)을 line-lease 로 써요.
6. 새 테스트마다 가드 뮤턴트 하나, 위험 경로 실행(임시 HOME 스크립트: 알림이 있어도 쓰기가 지금과 같은 바이트, 기록 줄에 값 없음, 기록 경로의 링크를 따라가지 않음), `moai check affected --card`, READY.

## Risks

- 관찰 기간에는 모양 틀린 값이 지금처럼 써지고, Claude Code 가 그 파일을 통째로 안 읽을 수 있어요. 내용이 사라지는 다시 쓰기도 지금처럼 일어나요(확인한 교체가 아니면 원본 보관이 없어요). 답이 파일과 규칙을 알려요.
- 기록은 로그 크기 상한이 차거나 상태 폴더를 못 쓰면 빠져요. 답의 `observation_recorded: false` 가 알려요.
- 기록은 적용한 체크아웃의 상태 폴더에만 있어서 워크트리마다 따로이고, 워크트리를 지우면 함께 사라져요.
- 파일을 바꾸는 적용마다 파일 하나가 새로 생기거나 늘어요. 적용을 부르는 기존 테스트(`src/claude.rs` 단위 테스트, 훅·상태줄 CLI 테스트)가 달라지는지 `check affected` 로 확인해요.
- `command` 가 빈 문자열이 아니어야 한다는 것은 문서의 "a `command` string" 에 moai 가 더한 조건이에요.
- 와일드카드 판정은 문서의 예를 따른 근사예요. 알림이라 틀려도 쓰기는 같아요.
- `go-hooks-apply` 의 첫 단계는 보호 없이 파일을 다시 인코딩해서 뒤의 검사는 이미 바뀐 바이트를 봐요(후속).
- 문서의 무시 env 목록은 이 SPEC 이 알리는 저장 위치 묶음보다 길어요(내보내기·로그인·대화 시간 변수 등).
- 되돌림이 없는 동안에는 문서 모양 밖의 다른 거부 사유를 쓴 뒤에 알 수 없어요(자식 SPEC).
- Windows 는 not_run 이에요.

## Decisions

추천이 맨 앞이고 기본값이에요.

- D-1 (검사를 두는 곳): (A, 리더 결정) 기존 쓰기 경로 안의 검사로 넣고 별도 쓰기 함수는 만들지 않아요 — 3중 병합·원자 교체·영수증이 이미 있어요. (B) 개정 2 처럼 별도 writer 를 만들어 호출부를 옮겨요 — 지난 코드는 아무도 부르지 않는 두 번째 쓰기 경로였어요.
- D-2 (모양 거절과 내용 손실 거절의 시작 방식): (A, 리더 결정) observe 로 시작해요: 알림과 기록만. 되돌리기: D-6 의 전환. (B) 착지부터 막는 예외 — 사용자만 고를 수 있어요(2026-11-20 사용자 결정 "새 게이트는 observe 로 시작"의 예외). 고르면 REQ-1·REQ-6 을 거절로 개정해요. 근거 수치는 이전 세대 배포 템플릿(`settings.json.tmpl`)을 읽어 센 것이고 실행하지 않았어요: 모양 거절 대상 0(`statusLine` 은 `command` 꼴에 `refreshInterval` 60, `outputStyle` 은 문자열), 관찰 규칙 대상 1(438행의 agent teams env; 병합 경로는 백업에 남겨요). 템플릿이 만드는 `PATH` 값과 내용 손실은 확인하지 않았고, 실제 업데이트 미리보기는 not_run 이에요. (C) A 에 더해, 내용이 사라지는 다시 쓰기 직전에 원본을 백업 폴더에 보관해요 — 아무것도 막지 않고 손실을 되돌릴 수 있지만 요구 하나가 늘어요.
- D-3 ("moai 는 agent teams·Code Mode 를 켜지 않는다"): (A, 리더 결정) moai 의 병합은 새로 켜는 스위치를 백업에 남기고, 호출자가 직접 요청한 값은 `would_refuse` 로 알려요 — "켜지 않는다"는 병합 경로에서만 지켜져요. 이 병합 동작은 이 SPEC 에서 착지부터 바뀌는 하나예요: 요청을 거절하지 않고 후보 하나를 백업에 두며, 직접 요청 길은 열려 있어요. (B) 직접 요청도 거절해요 — "절대"를 지키지만 사용자가 일부러 켜는 길도 막혀요. (C) 어디서나 알리기만(병합도 지금처럼 후보로 내요).
- D-4 (쓴 뒤의 doctor 확인과 되돌림): (A, 리더 결정) 이 SPEC 에서 떼어 자식 SPEC 으로 해요. 먼저 사용자가 임시 폴더에서 실제 `claude doctor` 의 출력·종료 코드·검사하는 파일·부작용을 관찰해요(문서는 "read-only installation and settings diagnostics" 라고만 해요: cli-reference.md, 04:01:40Z, sha256 6f14dee1b0d15f83910653e82fe01c11cf688dc6d3c31790c3a4cd187bac367a). 레인 세션은 실제 `claude` 를 돌리지 않아요. 카드 기준 2 의 되돌림 절반은 열려 있어요. (B) 관찰이 끝날 때까지 이 SPEC 을 멈추고 한 번에 해요. (C) 사용자가 관찰 실행을 세션에 허락해요(사용자만 고를 수 있어요).
- D-5 (이 SPEC 밖으로 보내는 추가 요구): (A, 리더 결정) 커밋되는 설정의 자체 점검과 `crossSessionInbound` 기본값(사용자 결정 2026-11-22)은 후속 일로 두고, 지금은 `docs/later.md` 줄로 남겨요. 자체 점검은 지금 코드가 `env` 를 커밋 파일에 쓰는 동작과 부딪혀 먼저 정할 것이 있어요. (B) 이 SPEC 을 넓혀요 — 요구가 12개를 넘어 나눠야 해요.
- D-6 (관찰 기간과 막기로 바꾸는 조건): (A, 기본·권장) 시작은 이 SPEC 의 착지예요. 점검 시점은 착지 7일 뒤와 다음 릴리스 준비 가운데 먼저 오는 때이고, 리더가 이 기계에서 적용한 체크아웃들의 `receipts/settings-write.jsonl`(상태 폴더 아래)을 `jq` 로 세어 적용 수, 규칙별 `would_refuse` 수, 거짓 양성 수를 `## Result` 에 적어요. 규칙(모양, 내용 손실, 관찰 규칙 각각)을 막기로 바꾸는 조건은 모두예요: 기간 안에 실제 적용 기록이 5건 이상, 그 규칙의 항목을 하나씩 확인해 거짓 양성 0, 답이 막힐 파일과 규칙(고칠 곳)을 알려 줌, 전환 결정이 기록됨. 하나라도 모자라면 observe 를 이어 가고 다음 릴리스 준비 때 다시 세며, 두 번의 점검에도 적용 기록이 모자라면 리더가 그 수와 함께 사용자에게 결정을 물어요. 전환은 개정한 SPEC 으로 해요. (B) 날짜 없이 기록이 충분하다고 판단할 때 — 점검이 밀려요. (C) 적용 수 조건 없이 7일 뒤 거짓 양성만 — 관찰한 것이 없어도 통과해요.

## Result

(draft)
