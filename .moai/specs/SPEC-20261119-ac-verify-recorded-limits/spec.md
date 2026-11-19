---
id: SPEC-20261119-ac-verify-recorded-limits
title: "The AC verify record states its limits, logs the Stop release, reads shell syntax and bounded files, and pins the host shapes (card C-130 slice 2)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-enforce-ac-verify-recorded
---

# The AC verify record states its limits, logs the Stop release, reads shell syntax and bounded files, and pins the host shapes (card C-130 slice 2)

## Intent

C-86 S3(SPEC-20261119-enforce-ac-verify-recorded)의 규칙 `ac_verify_recorded` 는 tier 2 감사에서 PASS 를 받았지만, 감사의 비차단 N1~N11 중 문서(N1)는 C-130 조각 1 이 닫았고, 나머지는 코드가 사용자에게 말하지 않는 한계이거나 입력이 이상할 때의 모습이에요: Stop 은 프로세스 폴더를 읽는데 관찰은 입력의 cwd 를 읽어요(N2), 공백 접기가 따옴표·줄바꿈·NBSP 안까지 접어요(N3), 하위 폴더에서 돌린 같은 명령도 기록돼요(N4), 4096바이트를 넘는 verify 명령은 채울 수 없는데 Stop 은 계속 요구해요(N5), 건너뜀 한도에 시험이 없어요(N6), HEAD·packed-refs 읽기에 한도가 없어서 큰 파일은 느리고 FIFO 면 멈춰요(N7), detached HEAD·sha256·reftable·symref 는 조용히 꺼져요(N8), 자리표시자 판정이 셸 문법을 놓쳐요(N9), 한도 뒤 풀어 줌이 로그 없이 조용해요(N11), 그리고 `run_in_background`·`interrupted` 의 다른 모양(문자열 "true", 1, snake_case)은 기록으로 들어가요(N-r1). 독자: `[hooks.modes]` 를 켜는 사용자, 이 규칙을 `enforce` 로 올릴지 정하는 리더, 훅 코드를 고치는 다음 레인.

이 SPEC 은 카드 C-130 의 기준 2~5 를 한 곳에 정해요. 판단이 필요한 둘(N4 하위 폴더, N10 쓰기 판정)은 정한 것과 그 이유를 적어요. 기준 3·4 의 일부는 `src/hook_guards.rs` 를 써야 하고 그 파일은 지금 C-112 가 예약하고 있어서, 그 부분은 예약이 풀린 뒤에 같은 SPEC 안에서 해요(Plan 의 단계 순서).

## Non-goals

- `ac_verify_recorded` 를 `enforce` 로 올리는 일과 기본 모드 변경: 사용자 몫이에요(N10 은 제안 노트로만, D-3).
- 실제 Claude·Codex·OMP 호스트를 띄워서 하는 확인: 호스트의 문서에서 읽은 모양을 시험으로 고정하고, 실제 호스트 실행은 이 SPEC 이 하지 않아요.
- 로그 줄의 새 종류 추가: Stop 의 풀어 줌은 이미 있는 `rule_release` 줄을 써요(D-1).
- 설정 파일 읽기(`src/hook_config.rs`)와 `tests/hook_dispatch_cli.rs` 의 변경: C-86 이 예약하고 있어요.

## Requirements

- REQ-1: When the Stop rule runs, the capabilities listing shall state, as limits of `ac_verify_recorded`, that Stop reads the process folder while the observation reads the input's cwd, that a verify command over 4096 bytes can never be recorded, that a detached HEAD, a sha256 object format, a reftable and a symbolic ref switch the rule off silently, and that a run from a subfolder of the work tree counts.
- REQ-2 (REQ-1): IF a session has already been blocked `MOAI_STOP_BLOCK_CAP` times in a row and a rule would have blocked again, THEN turn-validation shall let the turn end and write one `rule_release` line that names the rule and the session.
- REQ-3: The command comparison shall keep the white space inside quotes, a newline and a no-break space as they are and fold only a run of spaces and tabs outside quotes into one space.
- REQ-4 (REQ-3): IF a verify command holds a `<name>` written as an unquoted placeholder, THEN the verify list shall leave it out.
- REQ-5: The rule shall read HEAD, the branch's ref file and the packed-refs file as regular files only and no more than a fixed number of bytes each.
- REQ-6 (REQ-5): IF the file at one of those paths is a FIFO or larger than the bound, THEN the rule shall treat the commit as unreadable without waiting for the file.
- REQ-7: The rule shall ignore a Bash call that carries a started-in-background or interrupted mark in any of the spellings the hosts document or the code cannot tell apart from them (the string "true", the number 1, snake_case and camelCase keys).
- REQ-8 (REQ-7): IF the engine log has reached its size limit, THEN the rule shall read no lines, so that the turn ends.
- REQ-9: The `hook.dispatch` row of `docs/rust-product-path.md` shall say that the Stop rule writes the `rule_release` line, that the Stop rule asks about the process folder's SPEC, and that any mark other than the plain `false` or `null` reads as unfinished.
- REQ-10: The Stop rule shall count a write of the session only when its outcome is not a failure and its path is not outside the project folder, so that a write with an unknown outcome or with no path still counts.

## Acceptance

- AC-1 (REQ-1): `hook-dispatch --event session-check` capabilities 의 `limits` 가 프로세스 폴더·4096바이트·detached HEAD 와 sha256·reftable·symref·하위 폴더 한계를 말해요. · verify: `cargo test --locked --offline --test hooks_cli the_ac_verify_limits_name_the_process_folder_the_command_size_the_unreadable_heads_and_the_subfolder`
- AC-2 (REQ-2): 세션이 이미 두 번 막혔고 규칙이 또 막으려 하면 턴이 끝나게 두고 `rule_release` 한 줄(규칙 id·세션)을 남겨요; 규칙이 막지 않으려 하면 줄이 없어요. · verify: `cargo test --locked --offline --test hooks_cli the_stop_release_after_the_cap_writes_one_rule_release_line`; `cargo test --locked --offline --test hooks_cli no_rule_release_line_is_written_when_no_rule_would_block`
- AC-3 (REQ-3): 따옴표 안의 공백 두 개와 줄바꿈과 NBSP 는 접히지 않고, 따옴표 밖의 공백·탭 연속은 한 칸이 돼요. · manual: `these tests are written after C-112 releases src/hook_guards.rs (card C-130 slice 2, part 2); until then the criterion is open`
- AC-4 (REQ-4): 따옴표 없는 `<name>` 자리표시자가 든 명령은 목록에서 빠지고 `< in.txt > out.txt` 와 따옴표 안 `'<b>'` 는 남아요. · manual: `these tests are written after C-112 releases src/hook_guards.rs (card C-130 slice 2, part 2); until then the criterion is open`
- AC-5 (REQ-5, REQ-6): HEAD·ref 파일·`packed-refs` 가 FIFO 이거나 한도보다 크면 커밋을 읽을 수 없는 것으로 보고 기다리지 않아요(FIFO 시험은 제한 시간 안에 끝나요). · manual: `these tests are written after C-112 releases src/hook_guards.rs (card C-130 slice 2, part 2); until then the criterion is open`
- AC-6 (REQ-7): 문자열 `"true"`·숫자 1·snake_case·camelCase 로 온 백그라운드·중단 표시는 기록하지 않아요. · verify: `cargo test --locked --offline --test hooks_cli ac_verify_run_ignores_a_call_marked_unfinished_in_every_documented_spelling`
- AC-7 (REQ-8): 엔진 로그가 한도에 닿으면 규칙은 줄을 읽지 않고 턴이 끝나게 두고, 그 시험이 있어요. · verify: `cargo test --locked --offline --test hooks_cli the_ac_verify_rule_lets_the_turn_end_when_the_log_is_full`
- AC-8 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): 이웃 검사: 훅 시험과 문서·SPEC 검사가 그대로 통과해요. · verify: `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test hook_dispatch_cli`; `cargo test --locked --offline --test doc_references`; `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`
- AC-9 (REQ-9): 문서의 `hook.dispatch` 행이 Stop 규칙의 `rule_release` 줄, 프로세스 폴더 한계, "unfinished" 읽기, OMP 한계를 말해요(그 행 안에서만 읽어요). · verify: `sh -c 'row=$(grep -E "^\| .hook\.dispatch. \(rc\.2" docs/rust-product-path.md | head -1); for w in "the Stop rule" "writes the same line" "process folder" "reads as unfinished" "sends only"; do printf %s "$row" | grep -qF -- "$w" || { echo "row lacks: $w"; exit 1; }; done'`
- AC-10 (REQ-10): 실패한 Edit 와 프로젝트 밖 쓰기(절대 경로·`..` 로 나가는 경로)는 세지 않고, 프로젝트 쓰기·결과 모름(Codex `apply_patch`)·경로 없는 줄은 세요. · verify: `cargo test --locked --offline --test hooks_cli ac_verify_rule_counts_only_writes_that_did_not_fail_and_are_inside_the_project`

## Scope

- write: `.moai/specs/SPEC-20261119-ac-verify-recorded-limits/spec.md`, `src/hook_engine.rs`, `tests/hooks_cli.rs`, `docs/rust-product-path.md`, `.moai/specs/SPEC-20261119-enforce-ac-verify-recorded/spec.md`
- write (after C-112 releases it): `src/hook_guards.rs`
- read-only: `src/hook_config.rs`, `src/hooks.rs`, `tests/hook_dispatch_cli.rs`
- forbidden: 실제 HOME·실제 프로젝트의 훅 설정, 네트워크, 새 의존성, `src/main.rs`

## Plan

1. 이 SPEC 의 `spec check` 를 확인하고, 호스트 모양 조사(`research-c130-host-shapes`)의 문서 인용으로 REQ-7 의 낱말을 확정해요.
2. RED 먼저: AC-1·AC-2·AC-6·AC-7 의 시험(`tests/hooks_cli.rs`)을 쓰고 실패하는 것을 봐요. 그다음 `src/hook_engine.rs` 의 `AC_VERIFY_LIMITS`·Stop 풀어 줌 줄·`ran_to_its_end` 를 고쳐서 GREEN.
3. `docs/rust-product-path.md` 46줄에 Stop 풀어 줌(`rule_release`)과 새 한계를 적고 조각 1 의 AC 명령이 통과하는지 봐요.
4. `src/hook_guards.rs` 가 풀리면: AC-3·AC-4·AC-5 의 단위 시험을 먼저 쓰고(RED), `collapse_whitespace`·`verify_commands`·`head_sha` 를 고쳐요.
5. 변이 시험, fmt·clippy, 로컬 v2026.11.0 병합, 여섯 검사·`--bins`·등재 검사 전부·`skill_references_test`, push, 새 감사 에이전트(읽기 전용)의 감사.

## Risks

- 공백 접기를 줄이면(REQ-3) 전에 기록되던 명령이 안 맞을 수 있어요. 안 맞는 쪽은 "기록 안 됨"이라 `enforce` 에서 다시 돌리게 할 뿐 규칙을 풀지는 않아요. 시험으로 양쪽(SPEC 에 쓴 명령과 호출한 명령)이 같은 접기를 쓰는지 고정해요.
- 호스트가 보내는 모양을 문서에서만 읽어요(REQ-7). 문서에 없는 모양은 "끝나지 않은 것으로 본다"가 안전하지만 실제 호스트가 평범한 성공에 truthy 낱말을 보낼 수 있으면 과잉 차단이 생겨요: 조사에서 "문서에 없음"으로 나온 낱말은 넣지 않고 한계로 적어요.
- `src/hook_guards.rs` 의 예약이 늦어지면 AC-3·AC-4·AC-5 가 열린 채 이 SPEC 이 일부만 닫혀요: 카드 C-130 의 기준을 하나씩 표시하고, 남은 조각은 노트로 적어요.
- Stop 의 풀어 줌 줄이 한 세션의 `rule_release` 개수 세기(`rule_blocks_in` 은 `rule_block` 만 세요)와 섞이지 않는지: 시험이 한 세션의 두 줄 종류를 따로 세게 고정해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Stop 규칙의 한도 뒤 풀어 줌은 새 줄 종류가 아니라 기존 `rule_release` 줄(규칙 id·사유·세션)로 남긴다 | 합치기 확인의 풀어 줌이 이미 같은 뜻의 줄을 쓰고(`src/hook_engine.rs` `apply_rule_mode`), 새 종류를 더하면 조각 1 의 문서 행과 AC 목록이 또 바뀌어요(레인 agent-16, 리더 확인 대기) | 다른 이름의 줄로 바꾸고 문서 행의 낱말을 함께 바꿔요 |
| D-2 | 하위 폴더에서 돌린 같은 명령은 기록한다(한계로 적고 시험으로 고정) | 작업 트리 루트를 요구하면 하위 폴더에서 시작한 세션은 `enforce` 에서 기록이 없어 매번 막히고(한도로만 풀려요), 하위 폴더 실행이 만드는 위험은 상대 경로의 다른 결과인데 그 경우 명령이 실패해서 기록이 안 돼요(레인 agent-16, 리더 확인 대기) | 기록 조건에 "cwd 가 작업 트리 루트"를 더하고 시험을 뒤집어요 |
| D-3 | 쓰기 판정은 실패가 아닌 쓰기 가운데 경로가 프로젝트 안인 것만 센다(결과 모름·경로 없음은 셈). `enforce` 승격은 이 SPEC 이 정하지 않는다 | 리더 결정(C-130 노트 3387, 제안 노트 3384 의 (b)+좁힘): 모름을 빼면 규칙이 조용히 약해져요. 승격은 사용자 결정이에요 | 필터 두 줄을 빼면 이전 판정(모든 write 줄을 셈)으로 돌아가요 |

## Result

- 부분 결과(조각 2 앞부분, 기준 2·5): REQ-1·REQ-2·REQ-7·REQ-8·REQ-9 와 AC-1·AC-2·AC-6·AC-7·AC-9 는 닫혔고, REQ-3~REQ-6 와 AC-3·AC-4·AC-5 는 `src/hook_guards.rs`(C-112 가 예약 중)가 필요해서 열려 있어요. 열린 AC 는 `manual:` 로 두어서 있지도 않은 시험을 증거로 말하지 않아요; 예약이 풀리면 시험을 먼저 쓰고(RED) AC 를 `verify:` 로 되돌려요.
- RED→GREEN: 새 시험을 먼저 썼을 때 한계 문구·Stop 풀어 줌 줄·줄바꿈/NBSP 세 시험이 실패했고(왜 실패하는지 메시지로 확인: 한계 낱말 없음, `rule_release` 줄 0개, 줄바꿈 호출이 기록됨), 코드를 고친 뒤 `cargo test --test hooks_cli` 가 61개 통과했어요(줄바꿈 시험은 커밋에서 빼 두었고 `src/hook_guards.rs` 가 풀리면 되돌려요). 한도 가득 시험(`the_ac_verify_rule_lets_the_turn_end_when_the_log_is_full`), 하위 폴더 시험(`a_run_from_a_subfolder_of_the_work_tree_is_recorded`), "풀 줄이 없을 때 풀어 줌 줄 없음" 시험은 지금 동작을 고정하는 핀이라 처음부터 통과했어요(RED 가 아니에요).
- 변이(`target/mut130.sh`, 다섯 개, 파일은 해시로 복원 확인): M1 `run_in_background` 를 `== true` 로 되돌림, M2 `timedOutAfterMs` 무시, M3 Stop 풀어 줌 끄기, M4 응답 모양 검사 끄기, M5 한도 비교를 `>` 로 → 모두 KILLED. 문서 변이(AC-9, 다섯 개)도 모두 KILLED.
- 결정: D-2 는 시험으로 확인했어요: 하위 폴더 실행은 프로세스 폴더의 로그에 기록돼요(`a_run_from_a_subfolder_of_the_work_tree_is_recorded`). N10 은 리더가 (b)+좁힘으로 정했고(노트 3387) REQ-10·AC-10 으로 구현했어요(시험 한 개에 여섯 꼴: 성공한 프로젝트 쓰기·결과 모름·경로 없음은 세고, 실패한 Edit·프로젝트 밖 절대 경로·`..` 로 나가는 경로는 안 세요).
- 알려진 한계: OMP 확장은 `isError` 만 보내서 OMP 의 백그라운드·시간 초과 호출은 완료로 읽혀요(`capabilities` 한계와 문서에 적었고 코드 변경은 `src/hooks.rs` 의 생성 모듈이라 이 조각 밖이에요).
