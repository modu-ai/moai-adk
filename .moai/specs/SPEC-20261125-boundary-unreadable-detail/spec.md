---
id: SPEC-20261125-boundary-unreadable-detail
title: "The boundary hook names the unreadable part, reads a syntax check as read-only and judges the command a moai line runs after --"
status: draft
revision: 2
created: 2026-11-25
---

# The boundary hook names the unreadable part, reads a syntax check as read-only and judges the command a moai line runs after --

## Intent

다른 저장소(mo.ai.kr 리더의 덧붙임 8 P와 9 S)의 감사 세션이 경계 훅(`moai hook bash-boundary`, `src/hook_bash_boundary.rs`)에 거절당하면서 어느 구문 때문인지 듣지 못해 스크래치가 쌓이고 같은 명령을 여러 번 고쳐 썼어요: `unreadable` 거절 문구는 "읽을 수 없는 값이 있다"는 일반 문장뿐이에요. 이 SPEC 은 세 가지를 바꿔요. 첫째, `unreadable` 거절이 읽지 못한 명령 조각을 이름으로 말해요. 둘째, 읽기만 하는 구문 검사(`bash -n`, `zsh -n`, `sh -n`)를 통과시켜요: `-n` 은 스크립트를 실행하지 않고 파싱만 해요. 셋째, 실측에서 드러난 우회로를 막아요: 같은 `rm -rf /opt/…` 가 직접 쓰면 `delete-outside` 로 거절되는데 `moai heavy-slot-acquire … -- rm -rf /opt/…` 로 감싸면 통과해요. 훅의 래퍼 표에 `moai` 가 없어서 감싼 명령이 판단되지 않아요. 카드 본문이 "감싸기가 검사를 건너뛰면 그쪽을 막는 것이 먼저"라고 했으므로 이 SPEC 의 첫 줄은 이 우회로예요. 감사 스크래치 정리를 위해 삭제 경계를 넓히는 일은 이 SPEC 이 하지 않아요(D-1).

실측(2026-11-25, 설치된 `moai 4.2648.46-dev`; 훅 입력으로 명령 글자만 보냈고 아무것도 실행하지 않았어요; 이 저장소의 나무 끝으로 다시 재현하는 일은 첫 RED 시험이에요):

| 명령(작업 폴더) | 결과 |
| --- | --- |
| `rm -rf /opt/…` (체크아웃 안) | 거절 `delete-outside` |
| `moai heavy-slot-acquire … -- rm -rf /opt/…`, `-- git -C /opt/… reset --hard`, `-- sh -c "rm -rf /opt/…"` | 통과 (우회) |
| `bash -n 파일`, `zsh -n 파일`, `sh -n 파일` | 거절 `unreadable` (`-n` 이 허용 글자표에 없어요) |
| `rm -rf audit-x-scratch`, `rm -rf node_modules/.vite` (체크아웃 안) | 통과 |
| 같은 두 명령 (체크아웃 밖의 폴더) | 거절 `delete-outside` (경계가 의도한 동작) |
| `/bin/zsh 파일`, `bash 파일`, 리터럴 `cd … \|\| exit` 세 줄, heredoc `cat > 파일` | 통과 (보고된 거절이 재현되지 않아요) |
| 변수로 경로를 짠 `for` 루프 | 거절 `unreadable` (원칙: 변수는 따라가지 않아요) |

## Interview

- Q1: 감사 스크래치 정리(`rm -rf audit-…-scratch`, 시험 캐시)를 어떻게 할까요? 선택지: (가) 삭제 경계는 그대로: 체크아웃 안의 상대 경로 삭제는 이미 통과하고, 체크아웃 밖은 경계가 의도한 거절이에요. 감사 스크래치는 워크트리 아래 `target/<이름>/` 에 두는 규칙을 따르게 하고 거절 문구가 이를 알려요 (나) 프로필에 "스크래치 뿌리"를 선언하면 그 아래 삭제를 허용해요 (삭제 경계를 넓히는 설정) (다) 리더용 정리 명령 `moai audit-gc` 를 새 카드로 만들어요. → A: 아직 묻지 않았어요. 추천은 (가)예요(D-1): 삭제·권한 경계를 넓히는 쪽은 신중하게 하라는 카드의 지시와 실측(체크아웃 안은 이미 통과)이 같은 쪽을 가리켜요. (나)와 (다)는 사용자 선택이에요.
- Q2: `moai` 줄의 `--` 뒤를 판단하는 범위는 어디까지일까요? → A: 아직 묻지 않았어요. 추천: `moai` 와 `moai-cli` 로 시작하는 모든 줄에서 첫 `--` 뒤의 단어를 하나의 명령으로 판단해요(D-2): 지금 그렇게 실행하는 곳은 `heavy-slot-acquire` 와 카드의 한 줄 임대(`card line-lease`) 둘이에요. 명령 이름을 표로 고르면 새 감싸기 명령이 생길 때 같은 구멍이 다시 열려요.
- Q3: `-n` 을 허용 글자표에 더할까요? → A: 아직 묻지 않았어요. 추천: 예(D-3): `-n` 은 bash 와 zsh 에서 파싱만 하고 실행하지 않는 옵션이고, 글자를 더해도 `-c` 문자열은 지금처럼 판단돼요(`-n -c 'rm -rf /opt/…'` 는 여전히 `delete-outside`).

## Non-goals

- 삭제 경계(체크아웃 뿌리, 임시 폴더 뿌리, `.git`, 다른 체크아웃)를 넓히는 일과 새 정리 명령(D-1 선택지 (나), (다)).
- 변수를 따라가는 일: 훅의 원칙("변수는 따라가지 않아요")은 그대로예요.
- 스크립트 파일의 내용을 읽는 일(`bash 파일` 은 지금처럼 통과해요).
- `delete-outside` 와 다른 거절 문구의 바꿈: 이 SPEC 은 `unreadable` 문구에만 조각을 더해요.
- 보고된 거절 중 재현되지 않은 구문(`/bin/zsh 파일`, 리터럴 세 줄 스크립트, heredoc `cat > 파일`)의 수정: 같은 거절이 다시 나면 새 문구가 어느 조각인지 말해 줘요(D-5).

## Requirements

- REQ-1: When the boundary hook denies a command as `unreadable`, the hook shall name in its reason the innermost command it could not read, cut to 80 characters. · failure: none — the class word and the general sentence stay as they were.
- REQ-2: IF the unreadable command holds a secret-shaped word or a control character, THEN the hook shall name the command word alone.
- REQ-3: IF the command text is over the size limit or cannot be split into commands, THEN the hook shall say that in its reason instead of naming a command.
- REQ-4: When a shell is started with the `-n` letter, the hook shall read that letter as an allowed option. · failure: none — a `-c` string given with it is still judged as before.
- REQ-5: When a `moai` or `moai-cli` command line holds a `--`, the hook shall judge the words after the first `--` as a command of their own, with every rule it applies to that command typed bare. · failure: none — a pathspec after `--` names no runner and passes as before.
- REQ-6: The hook shall judge every command that is not a `moai` line that runs a command exactly as before this SPEC, the delete boundary, the checkout roots and the other four classes included.
- REQ-7: IF a `moai` line that names `heavy-slot-acquire` or `line-lease` holds a word that an expansion decides before its first literal `--`, or has no literal `--`, or has an expansion in the place of its subcommand, THEN the hook shall deny it as `unreadable`. · revision 2: 보안 검토가 짚은 구멍(변수가 `--` 로 펼쳐지면 뒤의 명령이 판단되지 않았어요)을 막고, 다른 래퍼(`nice $X …`)가 이미 하는 닫힘과 같아요.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 변수로 경로를 짠 반복문의 `unreadable` 거절이 읽지 못한 명령 조각을 80자 안에서 말해요; 비밀 모양 단어나 제어 문자가 든 조각은 명령 단어만 말해요; 너무 긴 명령과 쪼갤 수 없는 구문은 그렇다고 말해요 · verify: `cargo test --locked --offline --test hook_bash_boundary_cli an_unreadable_deny_names_the_part_it_could_not_read`
  verify: `cargo test --locked --offline --test hook_bash_boundary_cli a_secret_shaped_or_control_fragment_names_the_command_word_alone`
  verify: `cargo test --locked --offline --test hook_bash_boundary_cli an_oversized_or_unsplittable_command_says_so`
- AC-2 (REQ-4): `bash -n 파일`, `zsh -n 파일`, `sh -n 파일` 과 `-n` 이 든 묶음 글자는 통과해요; `-n -c 'rm -rf /opt/…'` 는 여전히 `delete-outside` 로 거절돼요 · verify: `cargo test --locked --offline --test hook_bash_boundary_cli a_read_only_syntax_check_is_not_unreadable`
- AC-3 (REQ-5, REQ-7): `moai heavy-slot-acquire … --` 와 `moai card line-lease … --` 뒤의 삭제, git 상태 변경, 보호 브랜치 변경, 훅 건너뛰기가 직접 쓴 같은 명령과 같은 거절을 받아요; 뒤가 스크립트 파일 실행이거나 `--` 뒤 경로 목록이면 통과해요; 변수가 정하는 명령 단어는 `unreadable` 이에요 · verify: `cargo test --locked --offline --test hook_bash_boundary_cli a_moai_line_after_double_dash_is_judged_like_the_bare_command`
  verify: `cargo test --locked --offline --test hook_bash_boundary_cli a_moai_line_after_double_dash_that_names_no_runner_passes`
  verify: `cargo test --locked --offline --test hook_bash_boundary_cli a_moai_line_that_runs_a_command_with_a_computed_word_is_unreadable`
- AC-4 (REQ-6): 체크아웃 안의 상대 경로 삭제는 통과하고 체크아웃 밖의 폴더에서는 거절돼요; 기존 표의 모든 사례와 통과 시험 둘은 그대로예요 · verify: `cargo test --locked --offline --test hook_bash_boundary_cli a_relative_delete_is_judged_by_the_folder_the_session_stands_in`
  verify: `cargo test --locked --offline --test hook_bash_boundary_cli`
  verify: `cargo test --locked --offline --test hook_preflight_bypass_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-boundary-unreadable-detail/spec.md`
- write: `src/hook_bash_boundary.rs`, `tests/hook_bash_boundary_cli.rs`
- read-only: `src/hook_engine.rs`, `src/hook_entry.rs`, `src/hook_latency.rs`, `tests/hook_preflight_bypass_cli.rs`
- forbidden: 삭제 경계의 뿌리와 `delete-outside` 규칙의 변경, 실제 HOME 과 원장, 설치된 `.claude` 설정, 파이썬

## Design

- 조각 이름: 판단기(`Walker`)가 명령 하나를 판단하다 `Unreadable` 을 돌려주면, 아직 기록이 없을 때 그 명령의 단어 글자(따옴표를 푼 글, 확장 표시는 그대로)를 80자로 잘라 기록해요. 안쪽 판단(`-c` 문자열, 명령 치환)이 먼저 돌려주므로 가장 안쪽 명령이 기록돼요. 거절 문구는 `…Unread part: <조각>` 으로 끝나요. `Class` 와 `Verdict` 의 모양은 바꾸지 않고 판단기의 기록 하나만 더해요(약 100곳의 `Err(Class::Unreadable)` 를 건드리지 않아요).
- 크기 한도와 쪼갤 수 없는 구문은 조각 대신 "the command text is over the size limit" 또는 "syntax that cannot be split into commands" 를 말해요.
- 비밀 모양 단어(`crate::secret_shaped`)나 제어 문자가 든 조각은 첫 단어(명령 이름)만 말해요. 훅의 이유는 모델에게 돌아가고 기록에 남기 때문이에요.
- `-n`: `OPTION_LETTERS_MINUS` 에 `n` 을 더해요(`+n` 은 아님). `set -n` 의 글자표도 같은 표를 쓰는데, 뒤 명령은 어차피 글자로 판단돼요.
- 감싸기: `simple()` 의 명령 이름이 `moai` 나 `moai-cli` 이면 `--` 뒤의 단어를 새 명령으로 `simple()` 에 다시 넣어요(`peel` 의 `--` 처리와 같은 모양). `--` 앞의 단어(자기 옵션)는 판단하지 않아요. `--` 가 없으면 지금처럼 통과해요.
- 닫힘(REQ-7, revision 2): 그 줄의 글자 단어가 `heavy-slot-acquire` 나 `line-lease` 를 이름으로 대면 명령을 실행하는 줄이에요. 이런 줄에서 첫 글자 `--` 앞에 확장이 정하는 단어가 있거나(`$X`, `-$X`, 명령 치환이 `--` 로 펼쳐질 수 있어요), 글자 `--` 가 하나도 없는데 확장이 있거나, 하위 명령 자리(첫 옵션 아닌 단어)가 확장이면 `Unreadable` 이에요(`Peel::Dynamic` 과 같은 닫힘). 다른 `moai` 하위 명령(`card note --ref "$R"` 등)은 변수가 있어도 지금처럼 통과해요: 명령을 실행하지 않아요.

## Plan

1. 이 SPEC 의 승인을 기다려요(tier 2: 삭제·권한 경계). 승인 뒤에 구현해요.
2. RED 먼저: `tests/hook_bash_boundary_cli.rs` 의 사례 표에 AC-1 ~ AC-4 의 시험을 써서 이 나무 끝의 빌드로 실패시켜요(우회가 이 나무 끝에서도 재현되는지가 첫 시험이에요).
3. `src/hook_bash_boundary.rs`: 판단기 기록, 문구, `n` 글자, `moai` 감싸기를 시험 하나씩 초록으로 만들어요.
4. `cargo fmt --all` 뒤 `moai check affected --card` 를 한 번 돌려요. READY 에 `guard-mutant:` 줄(아래)과 위험 경로 실행(훅 판단만 하는 임시 세계에서 감싼 삭제 글자를 보내요: 명령은 실행되지 않아요)을 담아요.

변이(READY 에 줄마다 `guard-mutant: <시험> | <뺀 코드> | exit <n>`): (1) `moai` 감싸기 분기를 빼면 AC-3 이 실패, (2) 감싸기에서 첫 `--` 뒤 대신 전체 단어를 넘기면 AC-3 의 통과 사례가 실패, (3) `n` 글자를 빼면 AC-2 가 실패, (4) `-n -c` 에서 `-c` 문자열 판단을 건너뛰면 AC-2 의 거절 사례가 실패, (5) 조각 기록을 빼면 AC-1 이 실패, (6) 비밀 모양 검사를 빼면 AC-1 둘째 시험이 실패, (7) 80자 자르기를 빼면 AC-1 이 실패, (8) 명령을 실행하는 `moai` 줄의 확장 단어 닫힘(REQ-7)을 빼면 AC-3 의 `a_moai_line_that_runs_a_command_with_a_computed_word_is_unreadable` 가 실패.

## Risks

- 훅은 안전 경계라 거절을 줄이는 모든 변경이 약화처럼 보일 수 있어요: `-n` 은 실행하지 않는 옵션이고 `-c` 문자열은 그대로 판단하며, 감싸기 규칙은 거절을 늘리는 쪽이에요. 감사는 작성하지 않은 세션의 전체 감사여야 해요.
- 거절 문구에 사용자가 쓴 명령 조각이 들어가요: 80자 한도와 비밀 모양·제어 문자 규칙으로 줄이지만, 비밀 모양이 아닌 값은 그대로 나가요.
- `moai` 줄의 `--` 뒤를 명령으로 읽으면 `moai` 가 명령이 아닌 것을 `--` 뒤에 받는 경우(경로 목록)에서 첫 단어가 실행 프로그램 이름일 때만 판단돼요: 그 밖에는 `wrap_spec` 이 없어 통과해요.
- 이 실측은 설치된 빌드 46 에서 한 것이고 나무 끝(`NeXT` c3e73024ac)의 빌드는 아니에요: 첫 RED 시험이 이 나무에서 다시 재현해요.
- 보고된 거절 세 가지(`/bin/zsh 파일`, 리터럴 세 줄, heredoc `cat > 파일`)는 재현되지 않았고, 보고자의 정확한 명령과 작업 폴더 없이는 원인을 알 수 없어요(`not_run`).
- 이 변경의 시험은 훅이 글자만 판단하는 임시 세계에서 하고 어떤 명령도 실행하지 않아요.
- 변수가 `--` 로 펼쳐지는 우회(`moai heavy-slot-acquire --name x $SEP rm -rf …`)는 revision 1 에 없었고 자동 보안 검토가 짚었어요: REQ-7 로 닫아요. 대가는 명령을 실행하는 두 하위 명령에서 `--name "$X"` 같은 변수 사용이 거절된다는 점이에요(글자로 쓰면 통과해요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 감사 스크래치 정리를 위한 삭제 경계 확대는 하지 않아요: 체크아웃 안은 이미 통과하고 밖은 의도한 거절이에요 | 카드: 관문을 넓히는 쪽은 신중하게; 실측 | 프로필의 스크래치 뿌리 설정(나) 또는 `moai audit-gc` 카드(다), 사용자 선택 |
| D-2 | `moai`/`moai-cli` 로 시작하는 모든 줄에서 첫 `--` 뒤를 명령으로 판단해요 | 새 감싸기 명령이 생겨도 구멍이 안 열려요 | 명령 이름 표(`heavy-slot-acquire`, `card line-lease` 둘)로 좁히기 |
| D-3 | `-n` 을 `-` 글자표에만 더해요 | 파싱만 하고 실행하지 않아요; `-c` 판단은 그대로 | `n` 한 글자를 빼요 |
| D-4 | 조각은 판단기의 기록 하나로 남기고 `Class`/`Verdict` 모양은 바꾸지 않아요 | 약 100곳을 건드리지 않아요 | 오류 값에 조각을 싣는 큰 변경 |
| D-5 | 재현되지 않은 보고 구문 세 가지는 이 SPEC 이 고치지 않고, 새 문구가 다음 거절에서 조각을 말해요 | 재현 없이 훅을 느슨하게 하지 않아요 | 보고자의 정확한 명령과 폴더로 새 RED 시험 |
| D-6 | 명령을 실행하는 두 하위 명령의 줄에서만 확장 단어를 `unreadable` 로 닫아요(revision 2, 보안 검토의 지적) | `nice $X …` 가 이미 닫혀 있어 같은 수준이에요; 모든 `moai` 줄을 닫으면 `moai card note --ref "$R"` 같은 정상 호출이 막혀요 | 모든 `moai` 줄의 확장을 닫기(더 엄격, 정상 호출이 거절돼요) |

## Result

구현 레인이 2026-11-25에 채웠어요(session `539952cc-1f67-4d6a-9025-b28ca0566b19`). 감사가 다시 확인해요. 기준 NeXT `c3e73024ac`, 가지 `spec/boundary-unreadable-detail-20261125`, 구현 커밋 `d85bc35a99`(감싸기), `dd012213d6`(`-n`), `184b353559`(확장 단어 닫힘), `3a3ccd3070`(조각 이름), `308ad51d55`(글롭·중괄호 닫힘), `11d6c51303`(변이 보강), `a5383bb4a5`(AC-4 시험). 이 SPEC 은 revision 2 로 승인됐어요(보안 검토의 첫 지적으로 REQ-7·D-6 을 더했어요). macOS arm64, 시험은 heavy-slot 한 칸에서 `cargo test --locked --offline` 으로 돌렸어요. `moai check affected --card` 영수증은 카드의 제출 노트에 있어요(head 가 바뀌지 않게 여기에는 숫자를 싣지 않아요).

- AC-1: met — `an_unreadable_deny_names_the_part_it_could_not_read`, `a_secret_shaped_or_control_fragment_names_the_command_word_alone`, `an_oversized_or_unsplittable_command_says_so` 각각 1 passed. 조각은 `Walker` 의 기록 하나이고 가장 안쪽 명령이 남아요(`sh -c 'rm -rf $X'` 는 안쪽 `rm -rf $X`, 큰따옴표 `sh -c "rm -rf $X"` 는 바깥 셸이 `$X` 를 풀기 때문에 바깥 명령이 이름이 돼요). 80자 안(자른 표시 `…` 포함), 비밀 모양 단어나 제어 문자가 든 조각은 명령 단어만, 명령 단어 자체가 그렇다면 `(command word withheld)`. 어휘 분석기 낱말에 `raw`(원문 조각)를 더해서 확장이 쓴 그대로 보여요(`rm -rf "$HOME/x"`; 이 SPEC 의 "따옴표를 푼 글" 보다 따옴표가 남는 쪽이 더 많이 알려줘서 골랐어요).
- AC-2: met — `a_read_only_syntax_check_is_not_unreadable` 1 passed(`bash -n`, `zsh -n`, `sh -n`, `/bin/zsh -n`, `bash -ne` 통과; `bash -n -c "rm -rf …"` 와 `sh -nc "rm -rf …"` 는 `delete-outside`; `bash +n`, `bash -z` 는 `unreadable`). `the_n_letter_loosens_no_other_command` 1 passed(`sudo -n`, `nice -n`, `xargs -n`, `env -n` 의 판단이 그대로이고, `set -n; rm -rf …` 는 이전에 `set -n` 자체가 `unreadable` 이었다가 이제 뒤 명령이 `delete-outside` 로 판단돼요). 기존 시험의 한 행을 바꿨어요: `ac18_the_fifth_audit_table_denied_and_allowed_shapes` 의 "표 밖 글자" 예시 `zsh -n -c 'ls'` 는 REQ-4 가 일부러 허용으로 바꾼 사례라서 `zsh -z -c 'ls'` 로 바꿨고, 같은 보호(표 밖 글자는 거절)가 그대로 고정돼요.
- AC-3: met — `a_moai_line_after_double_dash_is_judged_like_the_bare_command`, `a_moai_line_after_double_dash_that_names_no_runner_passes`, `a_moai_line_that_runs_a_command_with_a_computed_word_is_unreadable`, `a_moai_line_that_runs_a_command_with_a_glob_or_brace_word_is_unreadable` 각각 1 passed. 첫 시험은 감싸지 않은 설치 빌드가 아니라 이 나무에서 먼저 실패했어요(감싼 18줄이 모두 통과, 코드 뒤 초록). 확인 명령에 없는 글롭·중괄호 시험은 두 번째 보안 검토의 지적(`{--,}`, `??` 가 `--` 로 펼쳐지는 우회)을 막아요. 알려진 한계: `--` 뒤 첫 단어가 변수인 줄(`moai card note x -- "$TEXT"`)은 맨 명령처럼 `unreadable` 이에요(D-2 의 결과이고 이 SPEC 의 규칙이에요).
- AC-4: met — `a_relative_delete_is_judged_by_the_folder_the_session_stands_in` 1 passed(체크아웃 안에서 `rm -rf audit-x-scratch`·`node_modules/.vite` 통과, 체크아웃 밖의 평범한 폴더와 다른 저장소 폴더에서는 `delete-outside`, `../x` 는 `delete-outside`; 특성화 시험이라 코드는 바뀌지 않았어요). 전체 대상 `hook_bash_boundary_cli` 38 passed, `hook_preflight_bypass_cli` 6 passed.
- 변이: 11줄을 한 번에 돌렸어요(소스를 바꿔 시험하고 `git checkout` 으로 되돌렸고, 변이마다 `git status` 가 비어 있는 것을 읽었어요). 10개가 exit 101 로 잡혔어요. 살아남은 1개(감싸기에서 첫 `--` 뒤 대신 전체 단어를 넘기기, `…names_no_runner_passes` 가 exit 0)는 `--` 앞에 삭제·git 명령 단어를 둔 사례 2개를 더해 exit 101 로 잡혀요(`11d6c51303`). 줄 전체는 카드의 제출 노트에 있어요.
- 위험 경로: 방금 빌드한 훅 진입점(`moai hook bash-boundary`)에 글자만 보내 확인했어요, 아무것도 실행하지 않았어요. 감싼 `rm -rf /opt/…` 는 `delete-outside`, `{--,}` 로 `--` 를 만든 줄은 `unreadable` 과 `Unread part: …`, `bash -n 파일` 은 결정 없음, `for d in a b; do rm -rf $BASE/$d; done` 은 `Unread part: rm -rf $BASE/$d`.
- 한계(`not_run`): 보고된 거절 세 가지(`/bin/zsh 파일`, 리터럴 세 줄, heredoc `cat > 파일`)는 재현되지 않았고(D-5), 이 SPEC 은 그것을 고치지 않았어요. 새 문구가 다음 거절에서 조각을 말해요. `ac9_reasons_carry_the_class_and_no_command_byte` 는 `unreadable` 한 줄만 바꿨어요: 그 거절의 문구는 `Unread part:` 꼬리에서만 조각을 말하고, 나머지 네 분류는 여전히 명령 글자를 싣지 않아요.
