---
id: SPEC-20261126-shell-cd-guard
title: "A worktree session's shell cd is refused with the other way in one line"
status: approved
revision: 2
created: 2026-11-26
approval:
  by: leader
  at: 2026-11-26
  basis: "리더 GOAL-APPROVAL 노트 10298 (사용자 목표 2026-11-26 \"불포함 카드 까지 모두 다 구현을 해서 릴리즈를 하자. 카드 우선순위도 정해서 계속해서 모든 카드 구현 완료를 하자\" 에 따른 리더 판단, 사용자 직접 선택 아님): 판 2 의 방향은 결정 10265 와 보충 10289 예요"
  digest: "sha256:3092b8f1d45f47cdf1cb498007cd60c23e1c78741569436810a7a5fdc7c2f219"
---

# A worktree session's shell cd is refused with the other way in one line

## Intent

레인과 감사 세션은 카드마다 자기 연결 워크트리에서 일해요. Bash 에서 `cd` 를 쓰면 세션의 작업 폴더가 공유 체크아웃이나 다른 워크트리로 옮겨 가고, 밤새(2026-11-25~26) 같은 일이 여러 번 일어났어요(`docs/later.md`). 글 규칙만으로는 막히지 않으니 원인인 훅에서 막아요: 워크트리 세션의 Bash 명령이 세션 셸에서 `cd`·`pushd`·`popd` 를 실행하려 하면 tool-preflight 가 거절하고 대신 쓸 길을 한 줄로 알려요.

## Non-goals

- 쓰기 경계(H3)·삭제·git 판정의 범위는 바꾸지 않아요.
- `moai hook bash-boundary` 입구의 판정 표와 훅 배선, 설정 생성기는 바꾸지 않아요.
- R-5 의 거짓 거절을 없애는 렉서 수정은 이 SPEC 에 넣지 않아요(후속 작업).

## Interview

- Q1: 어떤 세션이 대상인가요? -> A: 요청의 "격리된 레인 세션"을 세션 체크아웃이 MoAI 프로젝트의 연결 워크트리인 세션으로 읽었어요(D-2, 위임 승인 때 리더가 바꿀 수 있어요).
- Q2: 어떤 `cd` 를 막나요? -> A: 세션 셸의 작업 폴더를 옮기는 것만요(요청: 맨 앞, `&&` 뒤, `;` 뒤; 문자열 안의 글자는 제외). 서브셸과 다른 셸의 문자열은 옮기지 않아요(D-3).
- Q3: 폴더를 옮기지 않는데 거절되는 남은 입력은 어떻게 하나요? -> A: 알려진 거짓 거절로 적고 테스트로 고정해요. 코드는 바꾸지 않아요(리더 결정, 노트 10265; 두 번째 감사 노트 10263). 범위는 D-7 이에요.

## Requirements

- REQ-1: When a Bash call in a worktree session of a MoAI project would run `cd`, `pushd` or `popd` in the session shell, the tool-preflight hook shall deny the call with a one-line reason that names the other ways.
- REQ-2 (REQ-1): IF the letters are inside a quoted string, a comment, a here-document body or an argument of another command, THEN the hook shall allow the call under this rule, except for the known false denials of R-5.
- REQ-3 (REQ-1): IF the `cd` runs in a subshell, a command substitution or a shell string given to another shell, THEN the hook shall allow the call under this rule, except for the known false denials of R-5.
- REQ-4 (REQ-1): IF the session checkout is the primary checkout, the folder is no MoAI project or the command cannot be split into commands, THEN the hook shall allow the call under this rule.
- REQ-5: The reason of the boundary's `git-state` deny shall leave out the advice to `cd` into a worktree.

## Acceptance

- AC-1 (REQ-1): 임시 HOME 의 MoAI 프로젝트에 `git worktree add` 로 만든 연결 워크트리에서 `moai hook-dispatch --host claude --event tool-preflight` 가 `cd x`, `ls && cd x`, `ls; cd x`, `{ cd x; }`, `if true; then cd x; fi`, `X=1 cd x`, `builtin cd x`, `eval 'cd x'`, `pushd x`, `popd` 를 `default` 와 `bypassPermissions` 두 모드에서 거절하고, 이유가 한 줄이며 절대 경로·`git -C`·객체 id 읽기·서브셸을 말해요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_worktree_session_cd_is_denied)'`
- AC-2 (REQ-2, REQ-3): 같은 워크트리 세션에서 `echo "cd x"`, `grep 'cd ' a.txt`, `printf '%s' 'a && cd x'`, `ls # cd x`, here-document 본문의 `cd x`, `ls cd`, `(cd x && ls)`, `v=$(cd x && pwd)`, `bash -c 'cd x'`, `sh -c "cd x"`, `command -v cd`, `/usr/bin/cd x` 는 거절되지 않아요(R-5 의 입력은 여기에 들지 않아요) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_cd_that_does_not_move_the_session_is_allowed)'`
- AC-3 (REQ-4): 주 체크아웃 세션, MoAI 프로필이 없는 폴더의 연결 워크트리, 나눌 수 없는 명령은 이 규칙으로 거절되지 않아요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_session_outside_a_worktree_keeps_cd)'`
- AC-4 (REQ-5): `git-state` 거절 이유에 `cd into it` 이 남지 않아요 · verify: `rg -c 'cd into it' src` (일치 없음, exit 1)
- AC-5 (REQ-1, REQ-2): 기존 경계 시험(`moai hook bash-boundary` 의 허용·거절 표)은 바뀌지 않고 통과해요 · verify: `cargo nextest run --test hook_bash_boundary_cli`
- AC-6 (REQ-2, REQ-3): 같은 워크트리 세션에서 R-5 의 대표 입력 네 가지는 `default` 와 `bypassPermissions` 두 모드에서 session-cd 규칙으로 거절돼요(알려진 동작을 고정) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_known_false_denial_stays_denied)'`

## Scope

- write: `src/hook_bash_boundary.rs`(세션 셸의 `cd` 를 찾는 함수와 `git-state` 문장), `src/hook_engine.rs`(`tool_preflight_deny` 의 호출 몇 줄, reports 커밋 가드 작업의 예약과 겹쳐 line-lease), `tests/hooks_cd_guard_cli.rs`, 이 SPEC 폴더.
- write 하지 않는 것: `src/hooks.rs`(설정 생성기; 처음 잡은 범위에 적혔지만 필요 없고 다른 작업이 예약), `moai hook bash-boundary` 의 판정 표, 사용자의 설정 파일, 훅 배선, 권한 모드.

## Plan

1. RED: `tests/hooks_cd_guard_cli.rs` 를 실제 `git worktree add` 픽스처와 임시 HOME 으로 쓰고 실패를 봐요.
2. GREEN: 기존 렉서(`lex`)로 괄호 깊이 0 의 명령에서 대입·예약어·`command`/`builtin` 을 건너뛴 첫 단어가 `cd`·`pushd`·`popd` 인지 보는 함수(`eval` 의 리터럴 문자열은 같은 깊이로)를 더하고, tool-preflight 의 guard 뒤에서 부르며, `git-state` 문장을 고쳐요.
3. 가드 뮤턴트, `moai check affected --card`, 임시 폴더의 실제 `hook-dispatch` 실행 한 번, READY.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 판정은 `hook_bash_boundary` 의 새 함수이고 tool-preflight 에서 불러요. `moai hook bash-boundary` 의 walker 는 바꾸지 않아요. | walker 에 넣으면 자기 워크트리에서 `cd` 를 허용하는 고정 사례 13건(`tests/hook_bash_boundary_cli.rs`)이 뒤집혀요. 설치된 배선은 tool-preflight 예요(이 저장소 `.claude/settings.json`). | walker 의 새 판정으로 옮기고 표를 고치는 개정. |
| D-2 | 대상 세션은 세션 체크아웃(`cwd` 를 담은 체크아웃, 저장소가 다르면 `CLAUDE_PROJECT_DIR` 의 것; H2 와 같은 정의)이 MoAI 프로젝트의 연결 워크트리인 세션이고, 권한 모드와 상관없어요. | 요청: "격리된 레인 세션". 레인·감사 세션은 연결 워크트리에서 일하고 주 체크아웃 세션(리더)은 그대로예요. | `bypassPermissions` 만, 또는 등록된 레인 세션만으로 좁히는 개정. |
| D-3 | 서브셸 `( … )`, 명령 치환, 다른 셸에 준 문자열(`bash -c`)은 허용해요. | 세션 셸의 폴더가 옮겨지지 않아요. | 막는 개정. |
| D-4 | `pushd`·`popd` 도 `cd` 처럼 다뤄요. | 같은 결과(작업 폴더 이동)예요. | 두 단어를 빼요. |
| D-5 | 파이프의 각 단계는 세션 셸로 보고 거절해요. | zsh 는 마지막 단계를 현재 셸에서 돌려요; 나머지 단계의 `cd` 는 쓸모가 없어요. | 마지막 단계만 보는 개정. |
| D-6 | `hook_engine.rs` 의 `capabilities` 가드 목록에는 항목을 더하지 않아요(상태 보고가 새 가드를 적게 말할 뿐 더 말하지 않아요). | `SPEC-20261126-reports-staged-guard` 작업이 같은 배열 두 줄을 고치는 중이에요. | 그 착지 뒤 후속 작업에서 한 줄 더하기. |
| D-7 | R-5 의 입력은 폴더를 옮기지 않아도 거절될 수 있어요. REQ-2·REQ-3 의 예외예요. | 리더 결정(노트 10265): 막아야 할 것을 놓치지 않는 거절 쪽 빈틈이고, 일부러 만든 입력에서만 나와요. | 렉서가 그 부분을 읽게 고치고 AC-6 의 입력을 허용 목록으로 옮기는 개정. |

## Risks

- R-1: 워크트리 안의 하위 폴더로 가는 정상적인 `cd` 도 거절돼요. 서브셸이나 절대 경로가 대신이에요.
- R-2: 계산된 명령어(`$C x`), 별칭·함수 안의 `cd`, zsh 의 폴더 이름만 입력(autocd)은 판정하지 않아요.
- R-3: 세션 폴더가 이미 워크트리 밖이면 이 규칙은 적용되지 않아요(기존 경계 판정은 그대로).
- R-4: 새 빌드를 리더가 설치한 뒤에야 레인 세션에 적용돼요.
- R-5: 알려진 거짓 거절(zsh 기준으로 세션 폴더가 안 옮겨지는데 거절될 수 있어요).
  - 명령 치환 `$( … )` 안의 주석이나 here-document 본문에 `)` 가 있고, 그 뒤에 `cd`·`pushd`·`popd` 와 짝 없는 `(` 가 이어지는 줄. 렉서가 그 `)` 를 치환의 끝으로 읽고, `(` 때문에 괄호 수가 다시 맞아요. macOS bash 3.2 는 here-document 쪽에서 치환을 실제로 일찍 끝내 폴더를 옮기므로 그 셸에서는 맞는 거절이에요.
  - here-document 끝 표시에 `$` 가 들어 있는 줄(`<<EOF$X`)의 본문. 렉서가 `$X` 를 버려서, 본문에 `EOF` 만 있는 줄이 나오면 그 뒤의 `cd` 를 명령으로 읽어요.

## Result

코드 head c2a322d3087808512ed612e5e605533f43de6e9a(base 91dd561a5aa2740aad0fae713a2e51a74525e7fc), 2026-11-26T06:22Z 까지 확인. AC 명령은 9b1daccb7b(그 뒤 변경은 시험 파일에 시험 하나 추가)에서 `verify` 에 쓴 그대로 실행했어요. 코드를 쓰기 전에는 AC-1 시험이 `"cd src"` 허용으로 실패했고 AC-2·AC-3 시험은 통과했어요.

- AC-1: met — `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_worktree_session_cd_is_denied)'` 1/1 통과
- AC-2: met — `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_cd_that_does_not_move_the_session_is_allowed)'` 1/1 통과
- AC-3: met — `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_session_outside_a_worktree_keeps_cd)'` 1/1 통과
- AC-4: met — `rg -c 'cd into it' src` exit 1
- AC-5: met — `cargo nextest run --test hook_bash_boundary_cli` 38/38 통과

위험 경로 확인: `a_linked_session_folder_is_judged_by_where_it_leads`(임시 HOME 에서 실제 바이너리, 링크를 거친 세션 폴더, 판정한 명령은 실행 안 됨), c2a322d308 에서 `hooks_cd_guard_cli` 4/4 통과. 가드 뮤턴트 7/7 exit 100(반납 노트). 승인은 노트 9869(리더 위임 판단)이고 승인 블록은 착지 때 리더가 써요.

### 수리: 첫 감사 FAIL(노트 10147)

head 05ccfb5e13 은 옮기지 않는 `cd` 세 가지(`$( … )` 안 here-document 본문, `$( … )` 안 주석 뒤, `$((2*(1+1)))` 뒤의 서브셸)를 거절했어요(REQ-2, REQ-3 위반). 렉서가 치환이나 산술을 일찍 끝내면 `)` 가 남으므로, 괄호 짝이 안 맞는 줄은 나눌 수 없는 명령으로 봐요(REQ-4). `moai hook bash-boundary` 의 렉서와 판정은 바꾸지 않았어요.

- 수리 head 8400181c950f2b16f67bf5185981914dfb6be101. 코드를 고치기 전에 AC-2 시험이 첫 번째 입력 거절로 실패했고(나머지 두 입력의 거절은 감사의 실행 확인), 고친 뒤 `hooks_cd_guard_cli`·`hook_bash_boundary_cli` 42/42 통과.
- 같은 head 에서 AC-1·AC-2·AC-3 `verify` 1/1, AC-4 `rg` exit 1, AC-5 38/38 통과.
- 가드 뮤턴트 추가 1개: 괄호 짝 확인을 지우면 `a_cd_that_does_not_move_the_session_is_allowed` exit 100.

알려진 빈틈(R-2 에 적은 것보다 넓어요):

- 줄 전체를 판정하지 않아 앞의 `cd` 도 통과해요: 배열 `a=()`, `(( … ))`, 붙여 쓴 `name()`, `;;`, `${…$(…)}`, `$( … )` 안의 짝 없는 따옴표, 괄호 짝이 안 맞는 줄(`(x)` 꼴이 아닌 `case` 포함; 이 마지막 것은 이번 수리로 생겼어요).
- 띄어 쓴 함수 정의 `name () {`, `function`, `coproc`, 짝이 맞는 `case` 는 그 자리에서 읽기를 멈춰요(앞의 `cd` 는 거절).
- `chdir`, `noglob cd`, `time -p cd` 는 허용하고 `cd x & wait` 는 거절해요(감사의 확인).
- `$( … )` 안 here-document·주석의 괄호 꼴과 끝 표시에 `$` 가 있는 here-document 는 본문의 `cd` 가 거절돼요(개정 2 에서 R-5 에 적고 AC-6 으로 고정).

### 개정 2: 두 번째 감사 FAIL(노트 10263)

두 번째 감사가 괄호 수가 다시 맞는 입력 세 가지의 거절을 확인했어요. 리더 결정(노트 10265)에 따라 코드는 바꾸지 않고 알려진 거짓 거절(R-5, D-7)로 적었어요. 끝 표시에 `$` 가 있는 갈래는 이번에 실행으로 확인했고, 리더가 같은 곳에 함께 적는 것에 동의했어요(2026-11-26 세션 간 답).

- 고정 테스트 커밋 196a0966676eabd18652ee0b89cde4ea63b89356. 제품 코드는 69e4cce8a6 와 같아요.
- AC-6: met — `a_known_false_denial_stays_denied` 1/1 통과(입력 4개 × 두 모드), 같은 변경에서 `hooks_cd_guard_cli`·`hook_bash_boundary_cli` 43/43 통과. 동작을 고정하는 테스트라 처음 실행부터 통과했어요(RED 기록 없음).
- 셸 실행 확인(`cd` 를 `echo` 로 바꿔서): 주석 입력 둘과 `EOF$X` 입력은 zsh 5.9 와 bash 3.2 모두 그 줄을 실행하지 않아요. here-document 입력은 zsh 5.9 만 본문으로 읽고 bash 3.2 는 실행해요.
- 가드 뮤턴트: 새 가드 코드가 없어 추가하지 않았어요(앞의 8개는 그대로).
