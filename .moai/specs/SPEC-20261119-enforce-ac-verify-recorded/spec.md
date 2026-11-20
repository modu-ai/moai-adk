---
id: SPEC-20261119-enforce-ac-verify-recorded
title: "A turn on a SPEC branch ends only after each AC's verify command has a recorded run at the current commit"
status: draft
revision: 2
created: 2026-11-19
parent: SPEC-20261119-enforce-spec-hooks
---

# A turn on a SPEC branch ends only after each AC's verify command has a recorded run at the current commit

## Intent

카드 C-86 의 AC-3: "작업이 끝났다"는 말이 증거 없이 나가면 안 돼요. 지침 파일에 "verify 를 돌려라"고 적는 것은 권고일 뿐이라(카드 근거 METR 2025), 돌렸다는 기록이 없으면 턴이 끝나지 않게 규칙으로 막아요. 지금 엔진은 AC 의 `verify:` 명령이 돌았는지 어디에도 기록하지 않아요(`src/verify.rs` 는 아무것도 쓰지 않아요). 이 조각은 두 가지를 해요: tool-observe 가 SPEC 작업 가지에서 AC 의 verify 명령과 같은 Bash 호출이 성공하면 `ac_run` 한 줄을 엔진 로그에 남기고, Stop 규칙 `ac_verify_recorded` 가 이 세션이 파일을 쓴 뒤 현재 커밋에 기록이 없는 AC 가 있으면 턴 종료를 거절해요(모드 off/observe/enforce, 기본 observe).

이 조각은 엔진 쪽만 만들어요. 생성되는 Claude·Codex 설정은 아직 `PostToolUse Bash` 를 `hook-receipt` 로, `Stop` 은 아무것도 안 부르고 있어서(`src/hooks.rs` 의 `HOOK_SET`), 그 연결은 별도 조각이에요(Non-goals, Risks).

## Interview

- Q1: S3(AC-3 Stop hook)는 어느 방식으로 하나요? → A: 선택 B(리더 결정 note 2711): tool-observe 가 AC verify 명령과 같은 성공한 Bash 호출에 `ac_run(spec, ac, head)` 한 줄을 남기고, Stop 규칙 `ac_verify_recorded` 가 현재 head 에 기록 없는 verify 가 있으면 완료 보고를 거절해요. A(신뢰 실행기 기록)와 C(Stop 이 직접 실행)는 고르지 않았어요.
- Q2: 모드와 호스트는요? → A: off/observe/enforce, 기본 observe, Claude·Codex 만이에요(OMP 는 Stop 이 연결돼 있지 않아요). enforce 를 기본으로 올리는 일은 별도 결정이에요(note 2711).
- Q3: `STOP_RULES` 에 규칙을 더해도 되나요? → A: 네. note 2711 이 Issue #2391 의 "결정 기록과 함께만" 조건에 맞는 결정 기록이에요(범주 none, 되돌릴 수 있음).
- Q4: 호스트 설정에 PostToolUse→tool-observe 와 Stop→turn-validation 을 연결하는 일은 이 조각에 넣나요? → A: 아직 몰라요. 카드 노트 `s3-wiring-finding` 으로 리더에게 선택지를 물었고, 기본안은 별도 조각(S3b)이에요. 이 조각은 연결 없이 엔진 쪽만 만들어요.

## Non-goals

- 생성되는 호스트 설정(`HOOK_SET`)에 `PostToolUse` → `tool-observe` 와 `Stop` → `turn-validation` 을 연결하기: 사용자 설정에 쓰이는 생성물과 그것을 고정한 시험을 바꾸는 일이라 별도 조각(S3b)이에요(Risks).
- enforce 를 기본 모드로 올리기: 별도 결정이에요(note 2711).
- Stop hook 이 verify 명령을 직접 실행하기(선택 C): 엔진의 "Stop 에서 하위 프로세스를 쓰지 않는다" 원칙과 부딪치고 시간 초과 위험이 있어요.
- 신뢰 실행기가 verify 실행 기록을 디스크에 쓰게 하기(선택 A): 에이전트가 `cargo test` 를 직접 돌리면 기록이 안 남아 규칙이 계속 막아요.
- OMP: Stop 이 연결돼 있지 않아요(`turn-validation is not wired for OMP`는 그대로예요).
- `manual:`·`unknown:` AC 와 자리표시자(`<…>`)가 든 verify 값은 묻지 않아요.

## Requirements

- REQ-1: When a Bash tool call succeeds in a work tree on the branch `spec/<yyyymmdd>-<slug>` and its command, with its white space collapsed, equals one command of a `verify:` value of that branch's SPEC, the tool-observe duty shall write one `ac_run` line with the SPEC id, the AC id, the command's position in that value, the full HEAD commit, the session and the time to the engine log.
- REQ-2 (REQ-1): IF the branch, the SPEC, the HEAD commit or the engine log cannot be read or written, THEN the tool-observe duty shall write no `ac_run` line and let the call through.
- REQ-3: When a session ends a turn on a SPEC work branch after a tool call of its own wrote a file, the rule `ac_verify_recorded` shall ask it to keep working if a `verify:` command of that branch's SPEC has no `ac_run` line for the current HEAD commit written at or after that write.
- REQ-4 (REQ-3): IF the branch is not a SPEC work branch, or the SPEC, the HEAD commit or the engine log cannot be read, or the session wrote no file, or no AC of the SPEC has a runnable `verify:` command, THEN the rule `ac_verify_recorded` shall let the turn end.
- REQ-5: The rule `ac_verify_recorded` shall act by its mode in `[hooks.modes]`: `off` does nothing, `observe` (the default) writes one `would_block` line and lets the turn end, and `enforce` keeps the session working with the Stop answer, within the existing MoAI stop-block cap.
- REQ-6: The rule `ac_verify_recorded` shall leave out an AC whose `verify:` value holds an unfilled `<…>` placeholder and an AC that has a `manual:` or `unknown:` reason.
- REQ-7: The capabilities report of the Claude and Codex hosts shall name the rule `ac_verify_recorded`, its default mode and its named limits.
- REQ-8 (REQ-1): IF the Bash call was started in the background or its response carries a background task id, or the host reports it interrupted, or the Claude input does not name the event `PostToolUse`, THEN the tool-observe duty shall write no `ac_run` line and let the call through.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-8): 성공한 Bash 호출이 AC 의 verify 명령과 공백만 달리 같으면 `ac_run` 줄이 하나 남고(SPEC id·AC id·전체 커밋·세션·시간), 다른 명령·실패한 호출·끝까지 돌지 않은 호출(백그라운드로 시작했거나 백그라운드 작업 id 가 있거나 중단됐거나 Claude 입력에 이벤트 이름 `PostToolUse` 가 없는 것)·SPEC 작업 가지가 아닌 곳·읽을 수 없는 SPEC 이나 커밋에서는 아무것도 남지 않고 호출은 그대로 통과해요.
  verify: `cargo test --locked --offline --test hooks_cli ac_verify_run`
- AC-2 (REQ-3, REQ-4, REQ-6): 이 세션이 파일을 쓴 뒤 현재 커밋에 `ac_run` 이 없는 verify 명령(AC 하나에 명령이 여럿이면 명령마다)이 있으면 Stop 이 이유(AC id 들과 SPEC id)와 함께 거절하고, 쓴 적이 없거나 기록이 모두 있거나 읽을 수 없거나 자리표시자·`manual:`·`unknown:` 뿐이면 그냥 끝나요.
  verify: `cargo test --locked --offline --test hooks_cli ac_verify_rule_stop`
- AC-3 (REQ-5): 모드 `off` 는 아무것도 쓰지 않고, 기본·`observe` 는 `would_block` 한 줄만 쓰고 끝나고, `enforce` 는 거절하되 연속 한도(2) 뒤에는 보내 줘요.
  verify: `cargo test --locked --offline --test hooks_cli ac_verify_rule_modes`
- AC-4 (REQ-7): `hook-dispatch --capabilities` 의 Claude·Codex 가 규칙 이름·기본 모드·한계를 말하고 OMP 는 그대로예요.
  verify: `cargo test --locked --offline --test hooks_cli ac_verify_rule_capabilities`
- AC-5 (REQ-1…REQ-7): 이웃 시험과 린트.
  verify: `cargo test --locked --offline --lib hook_`

## Scope

- write: `src/hook_engine.rs`, `src/hook_guards.rs`, `src/hook_config.rs`, `tests/hooks_cli.rs`, `tests/hook_dispatch_cli.rs`, `tests/green_test_lock.rs`, `.moai/specs/SPEC-20261119-enforce-ac-verify-recorded/spec.md`
- read-only: `src/spec_doc.rs`, `src/hooks.rs`, `src/verify.rs`, `src/hook_config.rs` (`[hooks.modes]` is S4's), `tests/hook_status_cli.rs`, `docs/design-review/appendix/operations/spec.md`
- forbidden: the real host settings and the real HOME (tests use a temporary HOME), `src/hooks.rs` (the generated hook set: S3b), `src/verify.rs` (it stays read-only), the Stop duty on OMP, any change that lets a hook allow what it refused before outside the new rule

## Plan

1. 시험 먼저(`tests/hooks_cli.rs`, 임시 저장소·임시 HOME, 손으로 만든 `.git` 에 `HEAD`·ref 파일): 관찰 쪽 `ac_verify_run*`, Stop 쪽 `ac_verify_rule_stop*`·`ac_verify_rule_modes*`·`ac_verify_rule_capabilities*`; 실패(RED)를 봐요.
2. `src/hook_guards.rs`: 파일 시스템만 읽는 `head_sha`(HEAD 가 가리키는 ref 의 커밋, `packed-refs` 포함)와 `branch_spec_commands`(가지의 SPEC 을 읽어 AC 별 verify 명령 목록).
3. `src/hook_engine.rs`: `tool_observe` 가 `ac_run` 줄을 더하고, `StopRule`·`turn_validation_with` 가 모드를 읽고, 규칙 `ac_verify_recorded` 를 `STOP_RULES` 에 더하고, 능력 보고에 이름·기본 모드·한계를 적어요.
4. 이웃 시험, mutation(위험한 것부터, 상한 12), SPEC `## Result`.

## Risks

- 연결 없음: 생성되는 Claude·Codex 설정에는 `tool-observe` 도 `turn-validation` 도 호출이 없어서(`src/hooks.rs` 의 `HOOK_SET` 은 `PostToolUse Bash → hook-receipt`), 이 조각만으로는 실제 세션에서 규칙이 돌지 않아요. 연결은 별도 조각(S3b)이고, 그때까지 시험은 `hook-dispatch` 를 직접 불러 증명해요(`hooks-status` 의 "no host run is proven" 와 같은 한계).
- 호스트의 실패 신호는 아직 실제 세션에서 확인하지 못했어요: Claude 에서 0 아닌 종료가 `PostToolUseFailure` 로 오는지, 백그라운드 호출이 `run_in_background`·`backgroundTaskId` 를 그 이름으로 싣는지는 문서와 감사 재현에 기대요. 모르는 모양은 성공으로 치지 않고(이벤트 이름이 없으면 기록 안 함) 실제 확인은 S3b 에서 해요.
- 커밋하면 `HEAD` 가 바뀌어 기록이 옛 커밋 것이 돼요: verify 를 돌린 뒤 커밋하면 규칙이 새 커밋에서 다시 돌리라고 해요(리더 결정이 "현재 head 의 기록"이에요). 같은 트리를 다시 증명하는 비용이에요.
- 더러운 작업 트리와 셸로 쓴 파일: "이 세션이 쓴 뒤"는 Write·Edit 류 도구 호출만 세요. `sed -i` 같은 셸 쓰기는 쓰기로 안 읽혀서 그 뒤 기록이 있어도 최신이 아닐 수 있어요.
- 명령이 같아야 인정: `cargo test … | tail` 처럼 뒤를 붙이거나 환경 변수를 앞에 붙인 호출은 verify 명령과 달라서 기록이 안 남아요(공백만 정리해 비교해요). SPEC 에 `verify:` 를 적을 때 그대로 돌릴 수 있는 한 줄로 적어야 해요.
- 기록은 한 줄이라 조작 가능: 에이전트가 `ac_run` 줄을 로그에 직접 쓸 수 있어요(엔진 로그는 일반 파일). 이 규칙은 실수와 게으름을 막지 속임수를 막지 않아요. 속임수는 감사자가 막아요.
- 로그 한도: 엔진 로그는 크기가 제한돼 있어서(`RECEIPT_LOG_LIMIT`), 한도를 넘으면 더는 쓰지 못하고 읽지도 못해요. 읽을 줄이 없으면 규칙은 "이 세션이 쓴 적 없음"으로 보고 턴을 끝내요(조용히 풀려요). 관찰만 하는 기본 모드에서는 영향이 없고, enforce 로 올릴 때 로그 회전이 먼저 있어야 해요(별도 결정).
- 순서는 시간으로: 프로젝트 쪽과 HOME 쪽 로그를 합쳐 읽어 줄 순서가 시간 순서가 아니라서 `unix_ms` 로 비교해요. 시계가 거꾸로 가면 틀릴 수 있어요.
- 같은 SPEC 의 AC 가 많을 때: 거절 이유는 AC id 를 다섯 개까지만 적어요.
- 감사(C-86 S3, 노트 2880·2945)가 찾은 한계는 SPEC-20261119-ac-verify-recorded-limits 가 한 곳에서 정해요: Stop 은 프로세스 폴더를 읽고 관찰은 호출이 말하는 cwd 를 읽어요(N2), 하위 폴더에서 돌려도 기록돼요(N4, 결정 D-2), 4096바이트가 넘는 verify 명령은 기록할 수 없는데 Stop 은 두 번 막고 세 번째 턴 끝에서 풀어 줄 때까지 계속 요구해요(N5), detached HEAD·sha256·reftable·symref 는 조용히 꺼져요(N8), 한도를 넘어 풀어 주는 일은 `rule_release` 줄을 남겨요(N11), 호스트가 보내는 "끝나지 않음" 표지는 문서에 있는 낱말을 방어적으로 읽어요(N-r1). 이 SPEC 의 `capabilities` 한계 문장이 그 목록을 사용자에게 보여 줘요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 선택 B: 관찰이 `ac_run` 을 남기고 Stop 규칙이 확인 | 리더 결정 note 2711; 신뢰 실행기 기록(A)은 직접 돌린 `cargo test` 를 못 보고 Stop 이 직접 실행(C)은 하위 프로세스 금지 원칙과 부딪쳐요 | A, C |
| D-2 | HEAD 는 파일 시스템으로 읽는다(`.git/HEAD` → ref 파일 → `packed-refs`), `git` 을 부르지 않는다 | Stop 은 하위 프로세스를 쓰지 않는 원칙이고, 같은 함수를 관찰과 Stop 이 써서 두 쪽의 커밋이 같아요 | `git rev-parse`(관찰만) |
| D-3 | "이 세션이 쓴 뒤"를 `unix_ms` 로 비교하고, 기록은 어느 세션의 것이든 인정 | 같은 커밋의 증명은 누가 돌렸든 같아요; 세션이 안 쓴 턴(대화만)은 묻지 않아요 | 줄 순서(두 로그를 합치면 시간 순서가 아님), 세션 별 기록만 인정 |
| D-4 | 규칙은 모든 Stop 규칙처럼 `[hooks.modes]` 의 모드를 가지고, 기본은 observe | 리더 note 2346·2711; 기본을 enforce 로 올리는 일은 별도 결정 | 모드 없이 항상 거절 |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. macOS arm64, 이 worktree(`spec/20261119-enforce-ac-verify-recorded`, S4 r2 head `b4ad1405bd` 위), 빌드 폴더 `~/.moai/build/moai-adk-86-s3`, `CARGO_INCREMENTAL=0`, cargo cache `/tmp/moai-rc-cargo-agent-3`. 시험 개수는 `cargo test --test <이름>` 을 파일마다 돌려 그 파일 로그의 `test result` 줄에서 읽었어요(`target/c86s3/n-*.log`, 종료 코드는 `target/c86s3/nb-exits.txt`, 2026-11-19T06:20Z 까지의 마지막 실행).

RED 먼저: 구현 전에 새 시험 6개 중 4개(`ac_verify_rule_capabilities_…`, `ac_verify_run_is_recorded_…`, `ac_verify_rule_modes_…`, `ac_verify_rule_stop_asks_…`)가 실패하는 것을 봤어요(`target/c86s3/red-exits.txt`). 나머지 둘(`ac_verify_run_ignores_…`, `ac_verify_rule_stop_lets_the_turn_end`)은 "기록하지 않는다", "끝낸다"를 고정하는 시험이라 구현 전에도 통과해요; 그 둘이 실제로 잡는지는 변이 시험으로 보였어요.

- AC-1: met — `cargo test --locked --offline --test hooks_cli ac_verify_run` 3개 통과(verify 명령과 공백만 달리 같은 성공한 Bash 호출은 `ac_run` 줄 하나: SPEC·AC·명령 위치·전체 커밋·세션 16자·시간, 한 AC 의 명령 둘은 위치 0·1, observe 줄도 그대로 쓰고 로그에 `cargo test` 문구가 없어요; 뒤에 단어·파이프·환경 변수 접두·`;` 를 붙인 명령, 실패한 호출(`PostToolUseFailure`), 빈 명령, `manual:`·`unknown:` 문장, SPEC 작업 가지가 아닌 곳, 다른 SPEC 이름의 가지, SPEC 파일·커밋 ref·40자 hex 가 없을 때, `.moai-cli` 가 없을 때는 줄이 없고 호출은 그대로 통과하고 아무것도 새로 만들지 않아요). 수리 1(tier 2 감사 B1, SPEC revision 2, REQ-8): 새 시험 `ac_verify_run_ignores_calls_that_did_not_finish` 를 먼저 써서 실패를 봤고(`run_in_background` 호출에서 `ac_run` 이 기록됨, `target/c86s3/r1-red.log`) `ran_to_its_end` 로 고친 뒤 통과해요: `run_in_background:true`(작업 id 있음·없음)·`backgroundTaskId` 만 있는 응답·`interrupted:true`·`hook_event_name` 없는 입력은 `ac_run` 이 0줄이고 호출은 통과하고 observe 줄 5개는 남아요; `run_in_background:false`·`interrupted:false` 인 `PostToolUse` 는 1줄 기록돼요. 변이 5개(`run_in_background` 안 읽기, 작업 id 안 읽기, `interrupted` 안 읽기, 이벤트 이름 안 요구, 검사 호출 빼기)가 모두 KILLED(`target/c86s3/mutants-r1.txt`, 복원 clean). `src/hook_engine.rs` 의 거짓 주석("the turn-validation rules are empty")도 고쳤어요. 이 수리가 못 하는 것: 호스트가 실제로 0 아닌 종료를 `PostToolUseFailure` 로, 백그라운드 호출을 그 이름의 필드로 보내는지는 실제 세션에서 확인하지 못했어요(S3b, Risks).
- AC-2: met — `… --test hooks_cli ac_verify_rule_stop` 2개 통과(이 세션이 쓰기 전에는 끝남; 쓴 뒤 `AC-1`·`AC-2`·`AC-6` 을 이름 대고 거절, `AC-3`(manual)·`AC-4`(자리표시자)·`AC-5`(unknown)는 안 대요; 기록하면 줄어들고, AC-6 은 명령 둘 다 해야 풀리고, 모두 하면 끝남; 이후 쓰기는 앞의 기록을 무효로 하고, 새 커밋(ref 파일 변경)도 무효로 하고, 쓴 적 없는 다른 세션은 안 물으며 그 세션의 실행이 첫 세션에게 인정돼요; 일곱 모양(가지가 main, SPEC 없음, 커밋 없음, 커밋이 hex 아님, 인정되는 명령이 없는 SPEC, Acceptance 없는 SPEC, 다른 SPEC 이름)은 enforce 에서도 끝나고 `stop_block`·`would_block` 줄이 없어요).
- AC-3: met — `… --test hooks_cli ac_verify_rule_modes` 통과(설정 없음·빈 파일·`[hooks.modes]` 만·`observe`·`"yes"`·다른 규칙만·깨진 TOML 은 모두 observe: 종료 0·빈 stdout·`would_block` 한 줄(rule `ac_verify_recorded`, reason 에 `AC-1`)·`stop_block` 없음; `off` 는 observe 줄 말고 아무것도 안 써요; `enforce` 는 `stop_hook_active` 면 바로 끝나고, 아니면 두 번 거절한 뒤 세 번째 턴이 끝나고 `stop_block` 줄 2개).
- AC-4: met — `… --test hooks_cli ac_verify_rule_capabilities` 통과(Claude·Codex: `rules` 1, `rule_ids` `["ac_verify_recorded"]`, `modes` 가 off·observe·enforce·`[hooks.modes]`·규칙 이름을 말하고, `limits` 가 "do not call tool-observe or turn-validation", "a commit changes HEAD", "shell writes (sed -i) are not writes", "must equal the verify command", "plain file" 을 말해요; OMP 는 "not wired" 그대로이고 옛 "STOP_RULES is empty" 말은 없어요).
- AC-5: met — `--test hooks_cli` 55, `hook_dispatch_cli` 56, `hook_status_cli` 10, `spec_before_code_hook` 9, `green_test_lock` 14, `ac_command_lint` 5, `git_env_scrub` 3, `spec_cli` 17 모두 통과; `cargo test --locked --offline --lib` 전체 1455 통과(새 단위 시험 3개 포함: `head_sha_reads_ref_files_and_packed_refs`, `head_sha_follows_a_linked_worktree_to_the_common_folder`, `verify_commands_split_spans_and_leave_out_placeholders`); `cargo fmt --check` 종료 0; `cargo clippy --locked --offline --all-targets -- -D warnings` 종료 0; `python3 -m unittest tests.ci_scope_test` 종료 0. 알림: 한 번의 `--lib` 전체 실행에서 `decide::outbound::tests::a_64_kib_field_of_a_repeated_key_prefix_is_screened_quickly` 가 실패했어요(기계 부하가 높을 때의 시간 의존 시험, 이 변경이 건드리지 않은 파일); 그 시험만 다시 돌리면 통과했고 마지막 전체 실행도 통과했어요.
- spec check: `moai spec check SPEC-20261119-enforce-ac-verify-recorded` 오류 0·경고 1(`section_unknown`, `## Interview`): 이 가지의 바이너리에는 `## Interview` 절이 아직 알려져 있지 않아요(S2 의 spec 검사는 다른 가지에서 감사 PASS 중이고 합쳐지지 않았어요). 합쳐지면 사라져요.
- mutation: 11개를 돌려 11개 모두 KILLED(`target/c86s3/mutants.txt`): Z1 앞부분만 같아도 기록, Z2 공백을 안 줄임, Z3 실패한 호출도 기록, Z4 커밋을 안 비교, Z5 시간을 안 비교, Z6 명령 위치를 안 비교, Z7 다른 세션의 쓰기도 인정, Z8 모드를 안 읽음, Z9 `off` 가 enforce 처럼 동작, Z10 자리표시자 명령도 묻기, Z11 쓴 적 없는 세션도 묻기. 돌린 뒤 `git status` 는 clean 이었어요. 시험이 못 잡는 곳(돌리지 않은 것): 4096 바이트를 넘는 명령의 건너뜀(`AC_RUN_COMMAND_LIMIT`)과 1 MiB 를 넘는 SPEC 의 건너뜀(`AC_SPEC_LIMIT`)은 시험이 없어요(성능 보호 한도).
- 이 조각 밖(돌리지 않았고 맞다고도 하지 않아요): 생성되는 Claude·Codex 설정이 `tool-observe` 와 `turn-validation` 을 부르게 하는 연결(S3b, 카드 노트 `s3-wiring-finding`)이라 이 규칙은 실제 세션에서 아직 돌지 않아요(Risks). `src/hooks.rs:79` 의 "turn-validation rules are empty" 주석은 이 조각 쓰기 범위 밖이라 그대로 남아 낡았어요(S3b 에서 고쳐요). 실제 `~/.claude` 설정과 실제 HOME 은 건드리지 않았어요(시험은 임시 HOME).
