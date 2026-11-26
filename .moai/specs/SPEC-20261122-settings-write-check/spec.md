---
id: SPEC-20261122-settings-write-check
title: "A settings write refuses a value outside the documented shape and never rewrites away what the user's file holds"
status: draft
revision: 3
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# A settings write refuses a value outside the documented shape and never rewrites away what the user's file holds

## Interview

- Q1: 설정 문서 조사 보고서의 항목들을 어떻게 다루나요? → A: 모든 항목을 카드로 만들어요(2026-11-22 사용자 선택, 이 작업의 카드 본문에 기록). 개정 3 의 새 선택(Decisions D-1~D-5)에 대한 사용자의 답은 아직 없어요.

## Intent

설정 파일에 모양이 틀린 값 하나를 쓰면 Claude Code 가 그 파일을 통째로 읽지 않아요: "These settings did not load, so nothing in them is in effect, including hooks, permission rules and sandbox settings"(settings.md, `-p` 실행). 별도 쓰기 경로를 새로 만들지 않고 기존 프로젝트·로컬 설정 쓰기(`claude-settings-plan`/`-apply`, `hooks-plan`/`-apply`, 상태줄 적용이 함께 쓰는 `src/claude.rs` 의 계획·적용 길) 안에 검사를 넣어요. 사용자 범위 출력 스타일 쓰기(`src/home_style.rs`)와 `go-hooks-apply` 의 첫 단계는 이 길을 지나지 않아 범위 밖이에요. 거절은 문서가 정한 모양을 벗어난 값과 내용이 사라지는 다시 쓰기에만 걸고, 문서가 "받되 맞지 않음"·"경고"로 부르는 값과 정책 항목은 먼저 관찰해요. 쓴 뒤의 `claude doctor` 확인과 되돌림은 실제 출력을 관찰한 뒤의 자식 SPEC 이 해요(D-4).

2026-11-26 에 직접 받아 읽은 문서(`https://code.claude.com/docs/en/<이름>`, 받은 시각과 sha256):

| 문서 | 받은 시각 | sha256 | 쓴 사실 |
| --- | --- | --- | --- |
| settings.md | 03:58:46Z | 7587a3251756cd151948718b70fff20fd26e3a4b545858d172d0941d57ab8f10 | 스키마가 거부한 값이 있는 파일은 통째로 안 읽힘, 개별 규칙 오류는 건너뜀 |
| settings-reference.md | 03:59:01Z | f87b28f014d323f1d99655f4dd8f329693e8295eec0f1cecf8af0bcc534a5249 | `statusLine`·`outputStyle` 모양, 프로젝트·로컬이 무시하는 env 이름 |
| permissions.md | 03:58:46Z | 687a80f8ae55138a03648fbd54fa1c109291656a512a8ed28c1c3f2825460615 | `Write`·`Glob` 류의 경로 규칙은 받되 맞지 않음, 하위 명령 앞의 `*` |
| env-vars.md | 03:58:46Z | 30cd22fe95a40ced8d175d368d5e915441014255f40965e681d2a3dcbdff317e | env 값은 셸이 처리하지 않고 쓴 그대로 |
| agent-teams.md | 03:58:46Z | 742a5e15697d90b09c48f85b83c61a1acf1d5fe88680c1bf36cd69ad9d5b5c9a | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` 이 팀을 켬 |
| errors.md | 04:01:40Z | e4d74f5e0dc5453ab6b522faabb541e3062722a6ff93912a2fb1cf3944a7a83c | 와일드카드 경고의 판정 예 |

받은 팁(41c16e60ac)의 코드에서 확인한 것: 호출자 설정 쓰기는 관리 키 목록(`src/claude.rs` `MANAGED_SETTINGS`), env 항목 거절(`env_entry_refusal`), 권한 규칙·모드 거절, 3중 병합(`plan_settings`)을 이미 가져요. `statusLine`·`outputStyle` 의 문서 모양을 계획 때 거절하는지는 RED 테스트가 확인해요. 반복 키와 달라지는 숫자 표기를 찾는 검사는 `src/go_settings_clean.rs` 에만 있어요.

## Non-goals

- `claude doctor` 실행, 거부된 쓰기의 되돌림, 제한된 실행기: 실제 출력·종료 코드·읽기 전용 동작을 관찰한 뒤의 자식 SPEC.
- 커밋되는 설정 파일의 자체 점검(env 항목, 작업 폴더 밖 allow, 훅 끄기, sandbox 끄기, 맨 `Edit` allow): 후속 SPEC.
- 사용자 범위 설정 쓰기(`crossSessionInbound` 기본값 포함): 후속 SPEC.
- 모든 쓰기의 백업, 목록 병합, `/status` 안내 줄, Go 훅 제거 단계(`go-hooks-apply`)의 다시 쓰기 보호, 사용자 출력 스타일 쓰기: 후속.
- Windows·텔레메트리·로그인 env 이름: 무시 목록 가운데 저장 위치 이름 묶음만 알려요.
- 이미 있는 미리보기가 `current`/`replace_with` 로 설정 값을 그대로 보여 주는 동작(`src/claude.rs`)은 바꾸지 않고 `docs/later.md` 한 줄로 남겨요.
- 관리 설정 파일과 `~/.claude.json`.

## Requirements

<!-- REQ-1 and REQ-6 block from landing without an observation window: the exception to internal/sdd rule 11 is Decisions D-2. -->
- REQ-1: When a plan would add or change `statusLine` or `outputStyle` to a value outside its documented shape, the project settings writer shall refuse the plan with the key path and the broken rule, issue no plan digest and leave the settings file and its receipt byte-identical (the shape of `statusLine` is an object whose `type` is `command` with a non-empty string `command` and, where present, a number `padding`, a number `refreshInterval` of at least 1 and a boolean `hideVimModeIndicator`; the shape of `outputStyle` is a non-empty string; the project settings writer is the plan and apply path behind `claude-settings-plan` and `claude-settings-apply`, `hooks-plan` and `hooks-apply`, and the status line apply, which writes the project and local settings files). · failure: none — the refusal is the handling of the unwanted case and nothing is written.
- REQ-2: When a plan would add or change a value that an observed rule names, the project settings writer shall list the key path and the rule under `would_refuse` in the plan and apply answers and leave the plan, its digest and the write unchanged (the observed rules are a `Write`, `NotebookEdit`, `MultiEdit`, `Glob` or `Grep` permission rule with a path other than `*`; a `Bash` allow rule whose `*` comes before a later word that determines the command; an `env` value holding `$` or starting with `~`; an `env` name of the storage-location family, which is `CLAUDE_CONFIG_DIR`, `CLAUDE_CODE_TMPDIR`, `HOME`, `TMPDIR`, `TMP`, `TEMP` or an `XDG_` name; and `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` or `CLAUDE_CODE_EXPERIMENTAL_CODE_MODE` with a value other than `0` or empty). · failure: none — observation only; turning an entry into a refusal is a separate recorded decision.
- REQ-3: The project settings writer shall refuse a request that names a top-level key outside the keys a caller may set (`attribution`, `fastModePerSessionOptIn`, `includeGitInstructions`, `disableAllHooks`, `sandbox` and a caller-supplied `hooks` among them), a widening `permissions.defaultMode`, a non-string `env` value or a permission rule that is not a non-empty string, and leave the file byte-identical.
- REQ-4: The project settings writer shall apply REQ-1 and REQ-2 only to a value the plan would add or change, never to one the file already holds or the plan leaves as it is.
- REQ-5: When the earlier-settings merge holds an unfit value, the merge preview shall leave that one value in the backup with its reason and keep every other candidate (an unfit value is a `statusLine` outside the shape of REQ-1, or an `env` entry that turns on agent teams or Code Mode and that the current file does not hold). · failure: none — the left-out value stays in the backup file.
- REQ-6: When a plan or apply would rewrite a settings file that repeats an object key or holds a number the rewrite would spell differently, the project settings writer shall refuse before any byte changes and name the file and either the affected top-level key or the repeated-key rule. · failure: none — the refusal leaves the file as it is.
- REQ-7: While no replacement is confirmed, the project settings writer shall keep every key the plan does not name with its value and order, and leave a key the user changed as it is.
- REQ-8: The project settings writer shall put no settings value into a refusal of REQ-1 or REQ-6, into a `would_refuse` entry or into a left-out reason of REQ-5.

## Acceptance

- AC-1 (REQ-1): a misshapen `statusLine` or `outputStyle` is refused at plan and apply with the key path and rule, no digest is issued, the settings file and receipt stay byte-identical, and a documented valid value plans and applies (temporary HOME) · verify: `cargo test --locked --offline --test settings_write_cli invalid_shapes_refuse_before_the_swap`
- AC-2 (REQ-4): a misshapen or listed value the file already holds blocks no other change and adds no `would_refuse` entry · verify: `cargo test --locked --offline --test settings_write_cli untouched_user_values_do_not_block`
- AC-3 (REQ-2): each listed value, `Bash(git * main)` and `Bash(git -C * status *)` among them, appears under `would_refuse`, `Bash(git status *)`, `Bash(ls *)` and `Edit(docs/**)` do not, the write succeeds and the digest equals the digest without the report · verify: `cargo test --locked --offline --test settings_write_cli questionable_values_are_reported_not_refused`
- AC-4 (REQ-3): a request naming each refused key, a widening mode, a non-string env value or an empty rule is refused and the file stays byte-identical (pins today's behaviour) · verify: `cargo test --locked --offline --test settings_write_cli forbidden_keys_are_never_written`
- AC-5 (REQ-5): the merge leaves out only the unfit status line and the feature switch, and carries a feature switch the current file already holds · verify: `cargo test --locked --offline --lib merge_keeps_unfit_values_in_backup`
- AC-6 (REQ-6): a file with a repeated key or a respelled number is refused by plan and apply and stays byte-identical, while a read of the same file still answers · verify: `cargo test --locked --offline --test settings_write_cli a_rewrite_that_would_lose_content_refuses`
- AC-7 (REQ-7): a valid write keeps every user key, list entry and their order, and a key the user changed stays as it is · verify: `cargo test --locked --offline --test settings_write_cli a_valid_write_keeps_every_user_key`
- AC-8 (REQ-8): a planted credential-shaped value and a planted plain value in a refused, reported or left-out entry appear in no new answer field · verify: `cargo test --locked --offline --test settings_write_cli no_value_reaches_a_new_answer`

## Scope

- write: `.moai/specs/SPEC-20261122-settings-write-check/**`, `src/settings_write.rs`, `src/claude.rs`, `src/claude_settings_merge.rs`, `src/go_settings_clean.rs`, `tests/settings_write_cli.rs`, `docs/rust-product-path.md`, `docs/later.md`
- write: `src/lib.rs`, `tests/areas.toml` (shared file 이라 line-lease)
- read-only: `src/status_line.rs`, `src/hooks.rs`, `src/home_style.rs`, `src/go_hooks.rs`
- forbidden: 테스트가 실제 사용자 설정 폴더를 읽거나 쓰는 일과 실제 `claude` 실행, 기존 거절을 약하게 하는 변경, Python

## Plan

1. RED: `tests/settings_write_cli.rs`(AC-1~4, AC-6~8)와 병합 단위 테스트(AC-5). AC-4 와 AC-7 은 지금 동작을 못박는 것이라 처음부터 초록일 수 있어요.
2. `src/settings_write.rs`(새 모듈, 판정 표는 한 곳) 에 모양 거절과 관찰 보고를 두고 `src/claude.rs` 의 계획 직전에서 바뀌는 값만 넘겨요(AC-1~4, AC-8).
3. 병합 후보에서 어긋난 상태줄과 새로 켜는 기능 스위치를 백업에 남겨요(AC-5). `go_settings_clean.rs` 의 반복 키·숫자 표기 검사를 같은 모듈에서 부르고, 쓰기가 일어날 때만 거절해요(AC-6, AC-7).
4. 글(`docs/rust-product-path.md` 의 설정 쓰기 줄)과 `docs/later.md` 한 줄(미리보기의 값 표시).
5. 새 테스트마다 가드 뮤턴트 하나, 위험 경로 실행(임시 HOME 스크립트: 거절 뒤 바이트 동일, 링크 거절, 반복 키 거절), `moai check affected --card`, READY. Result 에 거절·`would_refuse` 를 세어 적어요: 실제 업데이트 미리보기를 못 돌리면 not_run 으로 적어요.

## Risks

- 이미 키가 겹치거나 숫자 표기가 달라질 파일은 설정 쓰기 단계가 막혀요(`hooks-apply` 도 같은 길). 거절이 파일과 규칙을 이름해요.
- `command` 가 빈 문자열이 아니어야 한다는 것은 문서의 "a `command` string" 에 moai 가 더한 조건이에요.
- 와일드카드 판정은 문서의 예를 따른 근사예요. 거절이 아니라 알림이라 틀려도 쓰기는 같아요.
- `go-hooks-apply` 의 첫 단계는 보호 없이 파일을 다시 인코딩해서 뒤의 검사는 이미 바뀐 바이트를 봐요(후속).
- 문서의 무시 env 목록은 이 SPEC 이 알리는 저장 위치 묶음보다 길어요(내보내기·로그인·대화 시간 변수 등).
- 되돌림이 없는 동안에는 문서 모양 밖의 다른 거부 사유를 쓴 뒤에 알 수 없어요(자식 SPEC).
- Windows 는 not_run 이에요.

## Decisions

추천이 맨 앞이고 기본값이에요.

- D-1 (검사를 두는 곳): (A, 기본·권장) 기존 쓰기 경로 안의 검사로 넣고 별도 쓰기 함수는 만들지 않아요 — 3중 병합·원자 교체·영수증이 이미 있어요. (B) 개정 2 처럼 별도 writer 를 만들어 호출부를 옮겨요 — 지난 코드는 아무도 부르지 않는 두 번째 쓰기 경로였어요.
- D-2 (관찰 없이 바로 막는 것, `internal/sdd` 규칙 11 의 예외): (A, 기본·권장) REQ-1 의 모양 거절과 REQ-6 의 내용 손실 거절은 착지부터 막고 나머지는 `would_refuse` 로 관찰해요. 규칙 11 은 2026-11-20 사용자 결정("새 게이트는 observe 로 시작", `SPEC-20261120-gates-start-in-observe` 의 Intent)에서 온 것이라 이 예외는 사용자의 확인을 받는 것을 권해요. 근거 수치는 이전 세대 배포 템플릿(`settings.json.tmpl`)을 읽어 센 것이고 실행하지 않았어요: REQ-1 거절 0(`statusLine` 은 `command` 꼴에 `refreshInterval` 60, `outputStyle` 은 문자열), `would_refuse` 1(438행의 agent teams env; 병합 경로는 백업에 남겨요). 템플릿이 만드는 `PATH` 값과 REQ-6 은 확인하지 않았고, 실제 업데이트 미리보기는 not_run 이에요. (B) 전부 관찰부터 — 카드 기준 1 이 열려요. (C) 맞지 않는 규칙까지 전부 거절.
- D-3 ("moai 는 agent teams·Code Mode 를 켜지 않는다"): (A, 기본·권장) moai 의 병합은 스위치를 백업에 남기고, 호출자가 직접 요청한 값은 `would_refuse` 로 알려요 — "켜지 않는다"는 병합 경로에서만 지켜져요. (B) 직접 요청도 거절해요 — "절대"를 지키지만 사용자가 일부러 켜는 길도 막혀요. (C) 어디서나 알리기만.
- D-4 (쓴 뒤의 doctor 확인과 되돌림): (A, 기본·권장) 이 SPEC 에서 떼어 자식 SPEC 으로 해요. 먼저 사용자가 임시 폴더에서 실제 `claude doctor` 의 출력·종료 코드·검사하는 파일·부작용을 관찰해요(문서는 "read-only installation and settings diagnostics" 라고만 해요: cli-reference.md, 04:01:40Z, sha256 6f14dee1b0d15f83910653e82fe01c11cf688dc6d3c31790c3a4cd187bac367a). 레인 세션은 실제 `claude` 를 돌리지 않아요. 카드 기준 2 의 되돌림 절반은 열려 있어요. (B) 관찰이 끝날 때까지 이 SPEC 을 멈추고 한 번에 해요. (C) 사용자가 관찰 실행을 세션에 허락해요(사용자만 고를 수 있어요).
- D-5 (이 SPEC 밖으로 보내는 추가 요구): (A, 기본·권장) 커밋되는 설정의 자체 점검과 `crossSessionInbound` 기본값(사용자 결정 2026-11-22)은 각각 후속 카드로 열어요. 자체 점검은 지금 코드가 `env` 를 커밋 파일에 쓰는 동작과 부딪혀 먼저 정할 것이 있어요. (B) 이 SPEC 을 넓혀요 — 요구가 12개를 넘어 나눠야 해요.

## Result

(draft)
