---
id: SPEC-20261126-shell-cd-guard
title: "A worktree session's shell cd is refused with the other way in one line"
status: draft
revision: 1
created: 2026-11-26
---

# A worktree session's shell cd is refused with the other way in one line

## Intent

레인과 감사 세션은 카드마다 자기 연결 워크트리에서 일해요. Bash 에서 `cd` 를 쓰면 세션의 작업 폴더가 공유 체크아웃이나 다른 워크트리로 옮겨 가고, 밤새(2026-11-25~26) 같은 일이 여러 번 일어났어요(`docs/later.md`). 글 규칙만으로는 막히지 않으니 원인인 훅에서 막아요: 워크트리 세션의 Bash 명령이 세션 셸에서 `cd`·`pushd`·`popd` 를 실행하려 하면 tool-preflight 가 거절하고 대신 쓸 길을 한 줄로 알려요.

## Non-goals

- 쓰기 경계(H3)·삭제·git 판정의 범위는 바꾸지 않아요.
- `moai hook bash-boundary` 입구의 판정 표와 훅 배선, 설정 생성기는 바꾸지 않아요.

## Interview

- Q1: 어떤 세션이 대상인가요? -> A: 요청의 "격리된 레인 세션"을 세션 체크아웃이 MoAI 프로젝트의 연결 워크트리인 세션으로 읽었어요(D-2, 위임 승인 때 리더가 바꿀 수 있어요).
- Q2: 어떤 `cd` 를 막나요? -> A: 세션 셸의 작업 폴더를 옮기는 것만요(요청: 맨 앞, `&&` 뒤, `;` 뒤; 문자열 안의 글자는 제외). 서브셸과 다른 셸의 문자열은 옮기지 않아요(D-3).

## Requirements

- REQ-1: When a Bash call in a worktree session of a MoAI project would run `cd`, `pushd` or `popd` in the session shell, the tool-preflight hook shall deny the call with a one-line reason that names the other ways.
- REQ-2 (REQ-1): IF the letters are inside a quoted string, a comment, a here-document body or an argument of another command, THEN the hook shall allow the call under this rule.
- REQ-3 (REQ-1): IF the `cd` runs in a subshell, a command substitution or a shell string given to another shell, THEN the hook shall allow the call under this rule.
- REQ-4 (REQ-1): IF the session checkout is the primary checkout, the folder is no MoAI project or the command cannot be split into commands, THEN the hook shall allow the call under this rule.
- REQ-5: The reason of the boundary's `git-state` deny shall leave out the advice to `cd` into a worktree.

## Acceptance

- AC-1 (REQ-1): 임시 HOME 의 MoAI 프로젝트에 `git worktree add` 로 만든 연결 워크트리에서 `moai hook-dispatch --host claude --event tool-preflight` 가 `cd x`, `ls && cd x`, `ls; cd x`, `{ cd x; }`, `if true; then cd x; fi`, `X=1 cd x`, `builtin cd x`, `eval 'cd x'`, `pushd x`, `popd` 를 `default` 와 `bypassPermissions` 두 모드에서 거절하고, 이유가 한 줄이며 절대 경로·`git -C`·객체 id 읽기·서브셸을 말해요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_worktree_session_cd_is_denied)'`
- AC-2 (REQ-2, REQ-3): 같은 워크트리 세션에서 `echo "cd x"`, `grep 'cd ' a.txt`, `printf '%s' 'a && cd x'`, `ls # cd x`, here-document 본문의 `cd x`, `ls cd`, `(cd x && ls)`, `v=$(cd x && pwd)`, `bash -c 'cd x'`, `sh -c "cd x"`, `command -v cd`, `/usr/bin/cd x` 는 거절되지 않아요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_cd_that_does_not_move_the_session_is_allowed)'`
- AC-3 (REQ-4): 주 체크아웃 세션, MoAI 프로필이 없는 폴더의 연결 워크트리, 나눌 수 없는 명령은 이 규칙으로 거절되지 않아요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_session_outside_a_worktree_keeps_cd)'`
- AC-4 (REQ-5): `git-state` 거절 이유에 `cd into it` 이 남지 않아요 · verify: `rg -c 'cd into it' src` (일치 없음, exit 1)
- AC-5 (REQ-1, REQ-2): 기존 경계 시험(`moai hook bash-boundary` 의 허용·거절 표)은 바뀌지 않고 통과해요 · verify: `cargo nextest run --test hook_bash_boundary_cli`

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

## Risks

- R-1: 워크트리 안의 하위 폴더로 가는 정상적인 `cd` 도 거절돼요. 서브셸이나 절대 경로가 대신이에요.
- R-2: 계산된 명령어(`$C x`), 별칭·함수 안의 `cd`, zsh 의 폴더 이름만 입력(autocd)은 판정하지 않아요.
- R-3: 세션 폴더가 이미 워크트리 밖이면 이 규칙은 적용되지 않아요(기존 경계 판정은 그대로).
- R-4: 새 빌드를 리더가 설치한 뒤에야 레인 세션에 적용돼요.
