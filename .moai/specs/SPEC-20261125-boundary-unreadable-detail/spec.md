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
