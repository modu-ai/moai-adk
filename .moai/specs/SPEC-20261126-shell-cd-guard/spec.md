---
id: SPEC-20261126-shell-cd-guard
title: "A MoAI session keeps its shell folder: a cd inside a linked worktree, or from the main checkout into one, is refused with the other way in one line"
status: approved
revision: 4
created: 2026-11-26
approval:
  by: user
  at: 2026-11-27
  basis: "USER-APPROVAL 노트 10711 (개정 4, 사용자가 직접 골랐어요 2026-11-27, 리더 세션의 질문 도구 \"기본값으로 승인 (Recommended)\"; 리더가 노트에 적었고 레인은 그 노트를 인용만 해요). 승인은 개정 4 의 글 그대로예요: D-12 등록된 레인도 예외 없음, D-14 객체만 쓰는 git 명령은 지금처럼 거절, D-11 읽을 수 없는 목표는 허용, D-13 탈출구 없음. 앞 개정의 근거는 노트 10637"
  digest: "sha256:74287255680dae72001282c283db8daedf2325f913715623ef49e3b0429b729d"
---

# A MoAI session keeps its shell folder: a cd inside a linked worktree, or from the main checkout into one, is refused with the other way in one line

## Intent

레인과 감사 세션은 카드마다 자기 연결 워크트리에서 일해요. Bash 에서 `cd` 를 쓰면 세션의 작업 폴더가 공유 체크아웃이나 다른 워크트리로 옮겨 가고, 밤새(2026-11-25~26) 같은 일이 여러 번 일어났어요(`docs/later.md`). 글 규칙만으로는 막히지 않으니 원인인 훅에서 막아요: 워크트리 세션의 Bash 명령이 세션 셸에서 `cd`·`pushd`·`popd` 를 실행하려 하면 tool-preflight 가 거절하고 대신 쓸 길을 한 줄로 알려요.

개정 4(mo.ai.kr 리더 보고 2026-11-27 16:15 KST, 설치본 57-dev): 같은 규칙에 반대쪽 입구가 비어 있었어요. 기본(주) 체크아웃에서 띄운 세션이 `cd <연결 워크트리> && git …` 를 한 번 쓰면 셸 폴더가 그 워크트리로 옮겨지고, 그 순간부터 훅은 이 세션을 "자기 워크트리의 세션"으로 읽어요(`own_checkout` 이 입력의 `cwd` 가 든 체크아웃을 세션 체크아웃으로 보기 때문이에요). 그러면 돌아가는 `cd` 는 이 SPEC 의 규칙이 막고, 주 체크아웃을 향한 `git -C … commit-tree` 는 `git-state` 가 막아서 세션을 다시 띄우는 것 말고 회복 수단이 없어요. 개정 4 는 들어가는 `cd` 를 막고(REQ-6~8), 이미 갇힌 세션에 복구 안내를 주고(REQ-9), 객체만 쓰는 git 하위 명령의 취급을 사용자 결정으로 올려요(REQ-11).

## Non-goals

- 쓰기 경계(H3)·삭제·git 판정의 범위는 바꾸지 않아요(기본 선택에서; 결정 D-14 에서 사용자가 넓히기를 고르면 그 한 곳만 바뀌어요).
- `moai hook bash-boundary` 입구의 판정 표와 훅 배선, 설정 생성기는 바꾸지 않아요(같은 조건).
- R-5 의 거짓 거절을 없애는 렉서 수정은 이 SPEC 에 넣지 않아요(후속 작업).
- 이미 갇힌 세션의 셸 폴더를 훅이 되돌리지는 않아요(훅은 명령 글만 읽어요).

## Read forms

정적 명령 가드는 "이미 허용되는 명령"의 목록이 없으면 리뷰마다 새 지적이 나와서, 새 규칙보다 앞서 현재 동작을 적어요. 출처는 `src/hook_bash_boundary.rs` 의 `GIT_READ`·`classify_git`·`moves_session` 읽기와 `tests/hooks_cd_guard_cli.rs` 의 `MOVES`·`STAYS`, `tests/hook_bash_boundary_cli.rs` 의 고정 행이에요(NeXT 8b6ab17634 에서 읽었어요). 이 목록의 허용 줄은 개정 4 가 끝난 뒤에도 허용이고, AC-14 가 표로 고정해요. 설치된 훅에 실제 명령을 넣어 보는 확인은 이 레인에서 `HOME` 을 바꾸는 호출이 호스트 가드에 거절돼서 돌리지 못했어요(not_run, 거절 글은 카드 노트에 그대로 있어요). 아래는 코드와 시험 표를 읽은 결과예요.

| 세션 | 명령 | 지금 | 개정 4 뒤 |
| --- | --- | --- | --- |
| 주 체크아웃(MoAI 프로젝트) | `cd src`, `cd <주>/src`, `cd /tmp/x`, `pushd src`, `popd`, `cd`, `cd -`, `cd "$X"` | 허용(REQ-4) | 허용 |
| 주 체크아웃 | `cd <연결 워크트리>`, `cd <연결 워크트리>/src`, `pushd <연결 워크트리>`, `ls && cd <연결 워크트리>`, `X=1 cd <연결 워크트리>`, `builtin cd <연결 워크트리>`, `eval 'cd <연결 워크트리>'` | 허용 | **거절(REQ-6)** |
| 주 체크아웃 | `(cd <연결 워크트리> && ls)`, `v=$(cd <연결 워크트리> && pwd)`, `bash -c 'cd <연결 워크트리>'`, `echo "cd <연결 워크트리>"` | 허용 | 허용(REQ-8, D-3) |
| 주 체크아웃 | `git -C <연결 워크트리> status`·`log -1`·`diff`·`rev-parse HEAD`·`cat-file -p HEAD` 와 `GIT_READ` 전체 | 허용(읽기) | 허용 |
| 주 체크아웃 | `git add`·`commit`·`fetch`·`restore`·`rm --cached`(자기 체크아웃) | 허용 | 허용 |
| 주 체크아웃 | `git -C <연결 워크트리> add`·`commit`·`checkout …`(다른 체크아웃의 쓰기) | `git-state` 거절 | 같음 |
| 연결 워크트리 | `cd src`, `cd <다른 곳>`, `pushd`, `popd`(어떤 `cd` 든) | 거절(REQ-1) | 거절 + 복구 안내(REQ-9) |
| 연결 워크트리 | `MOVES`·`STAYS` 표의 모든 줄(`(cd src && ls)`, `bash -c 'cd src'`, `git -C /tmp status` 등) | 표대로 | 같음 |
| 연결 워크트리 | `git -C <주>`·`<다른 워크트리>` 의 `status`·`log -1`·`diff`·`rev-parse HEAD` | 허용(읽기) | 허용 |
| 연결 워크트리 | `git -C <주>` 의 `commit-tree`·`hash-object -w`·`mktree`·`mktag`·`write-tree` | `git-state` 거절(`GIT_WRITE`) | D-14 선택에 따라 |

`GIT_READ` 가 읽기로 읽는 하위 명령은 `status log diff show rev-parse rev-list ls-files ls-tree cat-file describe blame grep shortlog whatchanged name-rev merge-base merge-tree for-each-ref show-ref ls-remote check-ignore check-attr var version help count-objects fsck verify-commit verify-tag diff-tree diff-index diff-files range-diff cherry show-branch annotate difftool archive bundle format-patch request-pull get-tar-commit-id column stripspace check-ref-format interpret-trailers mailinfo mailsplit check-mailmap fmt-merge-msg verify-pack show-index` 이고, 형태에 따라 읽기인 것은 `branch`(`--show-current`·`-l`·`--list`·`-a`·`-r`·`-v`), `tag`(`-l`·인자 없음), `remote`(`show`·`get-url`), `stash`(`list`·`show`), `worktree list`, `reflog`(`expire`·`delete` 제외), `symbolic-ref`(인자 하나), `lfs`(`version`·`env`·`ls-files`·`status`·`logs`·`locks`·`pointer`)예요. `hash-object` 는 `-w` 가 없어도 `GIT_WRITE` 에 있어서 지금은 거절이에요(읽기 형태가 아니에요).

## Interview

- Q1: 어떤 세션이 대상인가요? -> A: 요청의 "격리된 레인 세션"을 세션 체크아웃이 MoAI 프로젝트의 연결 워크트리인 세션으로 읽었어요(D-2, 위임 승인 때 리더가 바꿀 수 있어요).
- Q2: 어떤 `cd` 를 막나요? -> A: 세션 셸의 작업 폴더를 옮기는 것만요(요청: 맨 앞, `&&` 뒤, `;` 뒤; 문자열 안의 글자는 제외). 서브셸과 다른 셸의 문자열은 옮기지 않아요(D-3).
- Q3: 폴더를 옮기지 않는데 거절되는 남은 입력은 어떻게 하나요? -> A: 알려진 거짓 거절로 적고 테스트로 고정해요. 코드는 바꾸지 않아요(리더 결정, 노트 10265; 두 번째 감사 노트 10263). 범위는 D-7 이에요.
- Q4: (개정 4) 이 사고를 같은 SPEC 의 개정으로 다루나요, 후속 SPEC 으로 다루나요? -> A: 개정 4 예요(D-9). 새 규칙이 승인된 REQ-4 와 AC-3 의 "주 체크아웃 세션은 거절하지 않아요"와 직접 부딪쳐서 한 문서에서 고쳐야 모순된 요구 둘이 남지 않아요.
- Q5: (개정 4) 들어가는 `cd` 를 무엇을 보고 막나요? -> A: 명령이 실행되기 전, 세션 폴더가 주 체크아웃일 때 `cd`·`pushd` 의 글자로 읽히는 목표가 같은 저장소의 연결 워크트리 안이면 막아요(D-10, D-11). 시작 폴더를 따로 기록하지 않고 `CLAUDE_PROJECT_DIR` 도 믿지 않아요.
- Q6: (개정 4) 그 워크트리에 등록된 레인은 예외인가요? -> A: 기본은 예외 없음이에요. 지금 레인은 주 체크아웃에서 등록하고 `EnterWorktree` 로 들어가서 등록부의 `worktree` 가 목표를 가리키지 못해요. 사용자 결정이에요(D-12).
- Q7: (개정 4) 객체만 쓰는 git 하위 명령은요? -> A: 기본은 지금처럼 거절이에요. 넓히는 선택지와 대가는 사용자 결정이에요(D-14).

## Requirements

- REQ-1: When a Bash call in a worktree session of a MoAI project would run `cd`, `pushd` or `popd` in the session shell, the tool-preflight hook shall deny the call with a one-line reason that names the other ways.
- REQ-2 (REQ-1): IF the letters are inside a quoted string, a comment, a here-document body or an argument of another command, THEN the hook shall allow the call under this rule, except for the known false denials of R-5.
- REQ-3 (REQ-1): IF the `cd` runs in a subshell, a command substitution or a shell string given to another shell, THEN the hook shall allow the call under this rule, except for the known false denials of R-5.
- REQ-4 (REQ-1): IF the session checkout is the primary checkout, the folder is no MoAI project or the command cannot be split into commands, THEN the hook shall allow the call under REQ-1 (revision 4: a primary-checkout session keeps every `cd` that does not enter a linked worktree; REQ-6 holds the entering `cd`).
- REQ-5: The reason of the boundary's `git-state` deny shall leave out the advice to `cd` into a worktree.
- REQ-6 (revision 4): When a Bash call in a primary-checkout session of a MoAI project would run `cd` or `pushd` in the session shell with a target that resolves to a folder inside a linked worktree of the same repository, the tool-preflight hook shall deny the call with a one-line reason that names the other ways, whatever the permission mode.
- REQ-7 (REQ-6): IF the target is not literal text (a variable, a substitution, `-`, a bare `cd`, a `CDPATH` search, a stack index) or names no folder, THEN the hook shall allow the call under REQ-6 (D-11); a literal target is an absolute path or a path relative to the session folder or to the folder an earlier `cd` of the same line left, with `~` and `..`, read in the forms REQ-1 reads the command words.
- REQ-8 (REQ-6): IF the `cd` runs in a subshell, a command substitution or a shell string given to another shell, THEN the hook shall allow the call under REQ-6, as REQ-3 does for the worktree session.
- REQ-9 (revision 4): The reason of REQ-1 shall add, in the same line, how a session that started in the main checkout and was moved by a shell `cd` gets out: end the session and start it again from the main checkout; the sentence names `EnterWorktree` as the way in, adds no `git -C`, no subshell `cd` and no new backticked form (D-8, D-13).
- REQ-10 (REQ-6): By the default of D-12 the hook shall not exempt a registered lane or leader session from REQ-6.
- REQ-11 (revision 4, default of D-14): The boundary's `git-state` deny shall keep refusing `commit-tree`, `hash-object`, `mktree`, `mktag` and `write-tree` that `-C`, `--git-dir` or `--work-tree` aim at a checkout that is not the session's own (AC-12 pins that answer).

## Acceptance

- AC-1 (REQ-1): 임시 HOME 의 MoAI 프로젝트에 `git worktree add` 로 만든 연결 워크트리에서 `moai hook-dispatch --host claude --event tool-preflight` 가 `cd x`, `ls && cd x`, `ls; cd x`, `{ cd x; }`, `if true; then cd x; fi`, `X=1 cd x`, `builtin cd x`, `eval 'cd x'`, `pushd x`, `popd` 를 `default` 와 `bypassPermissions` 두 모드에서 거절하고, 이유가 한 줄이며 절대 경로·부속 명령 뒤에 경로를 주는 `--manifest-path`·git 단일 명령·객체 id 읽기를 말해요(개정 3: `git -C` 와 서브셸은 레인 규칙이 금지한 꼴이라 글에서 빠졌어요, D-8) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_worktree_session_cd_is_denied)'`
- AC-2 (REQ-2, REQ-3): 같은 워크트리 세션에서 `echo "cd x"`, `grep 'cd ' a.txt`, `printf '%s' 'a && cd x'`, `ls # cd x`, here-document 본문의 `cd x`, `ls cd`, `(cd x && ls)`, `v=$(cd x && pwd)`, `bash -c 'cd x'`, `sh -c "cd x"`, `command -v cd`, `/usr/bin/cd x` 는 거절되지 않아요(R-5 의 입력은 여기에 들지 않아요) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_cd_that_does_not_move_the_session_is_allowed)'`
- AC-3 (REQ-4): 개정 4 에서 들어가지 않는 `cd` 로 좁혔어요. 주 체크아웃 세션의 `cd src`·`cd <주>/src`·`pushd src`·`popd`·`cd`, MoAI 프로필이 없는 폴더의 연결 워크트리, 나눌 수 없는 명령은 REQ-1 로 거절되지 않아요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_session_outside_a_worktree_keeps_cd)'`
- AC-4 (REQ-5): `git-state` 거절 이유에 `cd into it` 이 남지 않아요 · verify: `rg -c 'cd into it' src` (일치 없음, exit 1)
- AC-5 (REQ-1, REQ-2): 기존 경계 시험(`moai hook bash-boundary` 의 허용·거절 표)은 바뀌지 않고 통과해요 · verify: `cargo nextest run --test hook_bash_boundary_cli`
- AC-6 (REQ-2, REQ-3): 같은 워크트리 세션에서 R-5 의 대표 입력 네 가지는 `default` 와 `bypassPermissions` 두 모드에서 session-cd 규칙으로 거절돼요(알려진 동작을 고정) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_known_false_denial_stays_denied)'`
- AC-7 (REQ-1): 거절 이유가 권하는 cargo 꼴은 cargo 가 받아들여요: 이유의 백틱 안 `cargo` 꼴은 부속 명령(`test`)이 플래그 앞에 오고, `<path>` 자리에 임시 crate 를 넣어 다른 폴더에서 그대로 돌리면(`--no-run --offline` 만 덧붙임) cargo 가 종료 0 으로 끝나요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(the_session_cd_refusal_cargo_form_is_one_cargo_accepts)'`
- AC-8 (REQ-6): 카드 기준 1 의 앞쪽이에요. 임시 HOME 의 MoAI 프로젝트(주 체크아웃, 연결 워크트리 `wt-own`, 주 체크아웃 안에 만든 연결 워크트리 `.claude/worktrees/x`)에서 세션 폴더가 주 체크아웃이거나 그 하위 폴더일 때 `moai hook-dispatch --host claude --event tool-preflight` 가 `cd <wt-own>`, `cd <wt-own>/src`, `cd .claude/worktrees/x`, `pushd <wt-own>`, `ls && cd <wt-own>`, `cd src && cd <wt-own>`, `X=1 cd <wt-own>`, `builtin cd <wt-own>`, `eval 'cd <wt-own>'` 를 `default` 와 `bypassPermissions` 두 모드에서 거절하고, 이유가 한 줄이며 `moai-boundary/session-cd: ` 로 시작하고 `EnterWorktree` 를 말해요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_main_checkout_cd_into_a_worktree_is_denied)'`
- AC-9 (REQ-7, REQ-8): 같은 세션에서 `cd src`, `cd .`, `cd <주>/src`, `cd <연결 워크트리가 아닌 폴더>`, `cd "$W"`, `cd $W`, `cd -`, `cd`, `(cd <wt-own> && ls)`, `v=$(cd <wt-own> && pwd)`, `bash -c 'cd <wt-own>'`, `echo "cd <wt-own>"`, `git -C <wt-own> status`, 다른 저장소의 연결 워크트리로 가는 `cd` 는 REQ-6 으로 거절되지 않아요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_main_checkout_cd_that_does_not_enter_a_worktree_is_allowed)'`
- AC-10 (REQ-9): 카드 기준 2 예요. 연결 워크트리 세션의 `cd src` 거절 이유가 한 줄이고, 셸 `cd` 로 옮겨진 세션이 세션을 끝내고 주 체크아웃에서 다시 띄우라는 말과 `EnterWorktree` 를 담고, `git -C`·서브셸 `( cd`·`pushd` 를 권하지 않고, 백틱 안의 꼴은 cargo 꼴 하나뿐이에요. AC-7 의 cargo 확인은 이 글에도 그대로 통과해요 · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(the_session_cd_refusal_tells_a_moved_session_how_to_get_out)'`
- AC-11 (REQ-10): 카드 기준 1 의 뒤쪽이에요. 같은 임시 HOME 의 세션 등록부에 살아 있는 `agent` 항목(같은 세션 id)을 두어도 AC-8 의 `cd <wt-own>` 은 거절돼요(기본 선택의 동작을 고정; D-12 에서 다른 선택을 고르면 이 시험이 그 선택의 예외 시험으로 바뀌어요) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_registered_session_is_not_exempt_from_the_entering_cd_rule)'`
- AC-12 (REQ-11): 카드 기준 3 이에요. 연결 워크트리 세션에서 `git -C <주> commit-tree -m x HEAD^{tree}`, `git -C <주> hash-object -w x`, `git -C <주> mktree`, `git -C <주> mktag`, `git -C <주> write-tree` 가 `default` 와 `bypassPermissions` 모드에서 `moai-boundary/git-state` 로 거절되고, 같은 하위 명령을 `-C` 없이 쓰는 줄은 거절되지 않아요(기본 선택의 동작을 고정; D-14 에서 B·C 를 고르면 이 시험이 그 선택의 허용 시험으로 바뀌어요) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(an_objects_only_git_command_toward_another_checkout_stays_denied)'`
- AC-13 (REQ-1, REQ-2, REQ-3): 위 "Read forms" 표의 허용 줄 전부가 해당 세션에서 개정 4 코드로도 거절되지 않아요(`git -C` 읽기 줄, 자기 체크아웃의 쓰기 줄, `MOVES`·`STAYS` 표) · verify: `cargo nextest run --test hooks_cd_guard_cli -E 'test(the_read_forms_stay_allowed)'`

## Scope

- write: `src/hook_bash_boundary.rs`(세션 셸의 `cd` 를 찾는 함수와 `git-state` 문장), `src/hook_engine.rs`(`tool_preflight_deny` 의 호출 몇 줄, reports 커밋 가드 작업의 예약과 겹쳐 line-lease), `tests/hooks_cd_guard_cli.rs`, 이 SPEC 폴더.
- write 하지 않는 것: `src/hooks.rs`(설정 생성기; 처음 잡은 범위에 적혔지만 필요 없고 다른 작업이 예약), `moai hook bash-boundary` 의 판정 표, 사용자의 설정 파일, 훅 배선, 권한 모드.
- 개정 4: 이 개정을 쓰는 카드의 범위는 SPEC 폴더뿐이에요. 코드는 승인 뒤 같은 카드가 쓰고, 그때의 쓰기 범위는 위 목록과 같아요(`src/hook_bash_boundary.rs`, `src/hook_engine.rs` 호출부, `tests/hooks_cd_guard_cli.rs`). D-14 에서 B·C 를 고르면 `git-state` 판정 표와 `tests/hook_bash_boundary_cli.rs` 의 고정 행이 더해지고 그 선택은 새 범위를 SPEC 에 적은 뒤에만 써요.

## Plan

1. RED: `tests/hooks_cd_guard_cli.rs` 를 실제 `git worktree add` 픽스처와 임시 HOME 으로 쓰고 실패를 봐요.
2. GREEN: 기존 렉서(`lex`)로 괄호 깊이 0 의 명령에서 대입·예약어·`command`/`builtin` 을 건너뛴 첫 단어가 `cd`·`pushd`·`popd` 인지 보는 함수(`eval` 의 리터럴 문자열은 같은 깊이로)를 더하고, tool-preflight 의 guard 뒤에서 부르며, `git-state` 문장을 고쳐요.
3. 가드 뮤턴트, `moai check affected --card`, 임시 폴더의 실제 `hook-dispatch` 실행 한 번, READY.

개정 4(승인 뒤에만, 같은 순서): 4. RED: `tests/hooks_cd_guard_cli.rs` 에 AC-8~AC-13 시험을 쓰고 실패를 봐요. 기존 `a_session_outside_a_worktree_keeps_cd` 가 주 체크아웃에서 연결 워크트리로 가는 `cd` 를 허용으로 세고 있으면 AC-3 문구대로 들어가지 않는 `cd` 만 남겨요. 5. GREEN: `session_cd` 옆에 같은 렉서(`lex`)와 `change_dir` 가 읽는 글자 형태로 `cd`·`pushd` 의 목표를 읽는 함수를 더해요(파일만 읽고 프로세스를 띄우지 않아요, 목표가 든 체크아웃은 `checkout_of` 로 찾고 `common` 이 세션 체크아웃과 같으면서 `primary` 가 아닌 것만 거절). 6. `SESSION_CD` 에 REQ-9 의 문장을 더하고 `SESSION_ENTER` 글을 더해요. 7. 가드 뮤턴트(목표 읽기를 항상 허용, `common` 비교 제거, 복구 문장 제거), `moai check affected --card`, 임시 폴더의 실제 `hook-dispatch` 실행 한 번, READY.

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
| D-8 | 거절 글은 `git -C`, 서브셸 `( cd … )`, 맨 `cd <경로>`, `pushd`, `popd` 를 권하지 않아요. 글이 권하는 cargo 꼴은 `cargo test --manifest-path <path>/Cargo.toml` 이에요(플래그는 부속 명령 뒤). | 레인 규칙이 그 꼴들을 금지해서 권하면 레인이 두 번째 거절을 만나요(배치 24 의 C-583 이 글을 고쳤어요). cargo 는 부속 명령 앞의 `--manifest-path` 를 거절해요(cargo 1.94.1, 리뷰 노트 10564). | 글을 되돌리는 개정. |
| D-9 | 이 사고의 수정은 이 SPEC 의 개정 4 로 써요. 후속 SPEC 이 아니에요. | 개정 3 이 코드 착지(0d4dd1dd8b) 뒤에도 이 SPEC 을 고친 선례가 있어요. 새 규칙이 승인된 REQ-4·AC-3 과 부딪쳐서 두 SPEC 이 같은 규칙을 반대로 말하게 되는 것을 막아요. 승인 블록은 개정 3 의 것(노트 10637)을 지웠고 `status: draft` 예요: 개정 4 의 승인은 사용자가 고른 뒤 리더가 적어요. | 새 SPEC 으로 옮기고 이 SPEC 에는 REQ-4 의 한 줄만 고치는 개정. |
| D-10 | 들어가는 `cd` 는 명령이 실행되기 전에, 세션 폴더(`cwd`)가 주 체크아웃일 때의 목표로 판정해요. 시작 폴더를 따로 기록하지 않고 `CLAUDE_PROJECT_DIR` 도 판정에 쓰지 않아요. | 원인은 `own_checkout` 이 입력의 `cwd` 로 세션을 읽는 것이고(`src/hook_bash_boundary.rs`), 갇히는 순간은 그 `cd` 한 번이에요. 그때의 `cwd` 는 아직 주 체크아웃이라 기록 없이 판정돼요. `CLAUDE_PROJECT_DIR` 은 이 레인의 Bash 에서 주 체크아웃 세션일 때는 값이 보이고 워크트리 세션에서는 비어 있었지만(관찰, 훅 입력이 아니에요) 훅 프로세스의 값은 확인하지 못했어요(not_run). | 세션 시작 때 훅이 시작 폴더를 파일로 적고 그것으로 판정하는 개정(상태 쓰기가 생겨요). |
| D-11 | 목표를 글자로 읽을 수 없으면 허용해요(REQ-7). 경계 판정기의 "읽을 수 없으면 거절" 원칙을 이 규칙에는 쓰지 않아요. | 이 규칙은 세션이 갇히는 사고를 막는 편의 가드이고 삭제·보호 브랜치 같은 되돌릴 수 없는 일을 막지 않아요. `cd "$W"` 나 `cd -` 까지 막으면 주 체크아웃 세션의 흔한 `cd` 가 함께 거절돼요. R-2 의 읽지 않는 줄과 같은 처리라서 REQ-4 와 일관돼요. | 읽을 수 없는 목표도 거절하는 개정(주 체크아웃 세션의 변수 `cd` 가 모두 거절돼요). |
| D-12 | **사용자 결정(기본: A).** 들어가는 `cd` 에 등록된 레인 예외를 두지 않아요(REQ-10). 선택지는 아래 표예요. | 카드 제안(가)는 "그 워크트리의 레인으로 등록된 세션은 예외"예요. 이 저장소의 최근 레인은 주 체크아웃에서 등록하고 `EnterWorktree` 로 들어가서, 읽은 최근 `agent` 항목 셋(2026-11-27, `.git/.moai-cli/sessions/*.json` 을 `jq` 로 읽음; lease 가 아직 남았는지는 따지지 않았어요)의 `worktree` 가 모두 주 체크아웃이에요. 11-25 의 옛 항목 하나(`agent-2`)만 연결 워크트리 값이었어요. 그래서 등록부의 `worktree` 는 목표를 가리킨다고 믿을 수 없어요. 레인 규칙은 셸 `cd` 를 이미 금지하고, `EnterWorktree` 는 훅이 판정하지 않는 도구예요. | 표의 B 또는 C 로 바꾸는 개정. |
| D-13 | 거절 글은 한 줄이고 두 가지예요: 연결 워크트리 세션의 `SESSION_CD` 에 복구 문장(REQ-9)을 더하고, 들어가는 `cd` 에는 `SESSION_ENTER` 를 새로 둬요. 둘 다 `moai-boundary/session-cd: ` 로 시작하고 `git -C`·서브셸·`pushd` 를 권하지 않고 백틱은 cargo 꼴 하나뿐이에요(D-8, AC-7 이 그 꼴을 읽어요). "프로젝트 폴더로는 `cd` 를 허용하는 탈출구"는 쓰지 않아요. | 탈출구는 레인(`CLAUDE_PROJECT_DIR` 이 주 체크아웃)에게도 열려서 이 SPEC 이 막으려던 일(레인이 공유 체크아웃으로 `cd`)을 다시 열어요. 복구 문장은 `EnterWorktree` 로 들어간 레인이 따라 하지 않도록 "셸 `cd` 로 옮겨진" 세션만 말해요. 문장 초안은 아래 "글 초안" 이에요. | 탈출구를 두는 개정(레인 보호가 약해져요). |
| D-14 | **사용자 결정(기본: A).** 객체만 쓰는 git 하위 명령(`commit-tree`, `hash-object`, `mktree`, `mktag`, `write-tree`)이 다른 체크아웃을 향한 `-C` 일 때: 지금처럼 거절(REQ-11). 선택지는 아래 표예요. | 이 명령들은 객체 저장소에만 쓰고 ref·인덱스·작업 트리를 건드리지 않지만, 지금은 `GIT_WRITE` 한 줄이라 다른 체크아웃을 향하면 `git-state` 로 거절돼요. 가드를 넓히는 쪽이라 기본을 현재 동작에 둬요. 이 명령들의 `-C` 거절을 고정한 시험은 지금 없어서(`tests/hook_bash_boundary_cli.rs`·`hooks_cd_guard_cli.rs` 를 읽었어요) AC-12 가 처음 고정해요. | 표의 B 또는 C 로 바꾸는 개정. |

### D-12 선택지: 들어가는 `cd` 의 예외

| 선택 | 판정하는 것 | 대가 |
| --- | --- | --- |
| A(기본) | 예외 없음. 레인은 `EnterWorktree` 로 들어가요. | 셸 `cd` 로 자기 워크트리에 들어가려는 레인도 거절돼요(레인 규칙은 이미 금지). 카드 기준 1 의 "등록된 레인은 예외"는 이 선택에서 시험으로 보이지 않고, AC-11 이 예외 없음을 고정해요. |
| B | 살아 있는 `agent` 모드 등록 항목의 `session` 이 훅 입력의 `session_id` 와 같으면 예외(파일만 읽어요: `mode_of` 와 같은 읽기). | 최근 레인의 등록부 `worktree` 가 주 체크아웃이라 목표와 맞춰 볼 수 없어요. 그래서 등록된 모든 레인이 어느 워크트리로든 `cd` 할 수 있게 돼요(구멍). 목표와 `worktree` 가 같을 때만 예외로 좁히면(B') 구멍은 없지만 최근 레인에는 아무 효과가 없어요. 훅 입력 `session_id` 와 등록부 `session` 값의 꼴이 같은지는 확인하지 못했어요(not_run). |
| C | 카드 예약(잡은 세션 id 와 `--worktree`)이 목표 워크트리를 가리키는 세션만 예외. | 가장 정확하지만 훅이 원장 데이터베이스를 열어야 해요. 지금 어떤 훅도 원장을 읽지 않고(git 가드는 파일만 읽어요), 지연과 실패 방식이 새로 생겨요. |

### D-14 선택지: 객체만 쓰는 git 하위 명령

| 선택 | 허용하는 것 | 대가 |
| --- | --- | --- |
| A(기본) | 없음. 지금처럼 `git-state` 거절. | 갇힌 세션이 주 체크아웃 객체 저장소에 객체를 쓰는 `commit-tree` 를 못 써요(세션을 다시 띄워요). 판정 표와 AC-5 는 그대로예요. |
| B | `commit-tree`·`mktree`·`mktag`·`hash-object`(`-w` 유무와 상관없이)를 `-C` 목표의 `common` 이 세션 체크아웃과 같을 때만 허용. 연결 워크트리는 한 객체 저장소를 나눠 쓰므로 객체가 같은 곳에 쌓여요. `write-tree` 는 목표의 인덱스를 읽어 그 체크아웃의 상태를 읽으므로 빼요. | 판정 표(`classify_git` 의 쓰기 목록)가 바뀌어 AC-5 의 고정 행에 새 행이 더해져요. `commit-tree` 뒤의 `update-ref` 는 그것대로 판정돼서(보호 브랜치·다른 체크아웃의 `ref`) 거절이 남아요. 다른 저장소를 향한 `-C` 는 계속 거절이에요. |
| C | 위 다섯 가지를 어느 체크아웃이든 허용. | 다른 저장소의 객체 저장소에도 쓰게 돼서 "세션 체크아웃이 아닌 것의 상태" 경계가 이 명령들에 한해 사라져요. `write-tree` 가 들면 다른 체크아웃의 인덱스를 읽고 그 내용을 객체로 써요. 판정 표와 AC-5 가 바뀌어요. |

### 글 초안(D-13)

- `SESSION_ENTER`: `moai-boundary/session-cd: this command would move the shell of a session in the main checkout into a linked worktree, and a shell cd cannot take it back out. Stay where you are: name files by their absolute path, give a path argument to a command that takes one (`cargo test --manifest-path <path>/Cargo.toml`), or enter the worktree with EnterWorktree.`
- `SESSION_CD` 에 더하는 문장(기존 글 끝에 이어요): `If a shell cd (not EnterWorktree) already moved a session that started in the main checkout into this worktree, the shell cannot go back: end the session and start it again from the main checkout.`

## Risks

- R-1: 워크트리 안의 하위 폴더로 가는 정상적인 `cd` 도 거절돼요. 서브셸이나 절대 경로가 대신이에요.
- R-2: 계산된 명령어(`$C x`), 별칭·함수 안의 `cd`, zsh 의 폴더 이름만 입력(autocd)은 판정하지 않아요.
- R-3: 세션 폴더가 이미 워크트리 밖이면 이 규칙은 적용되지 않아요(기존 경계 판정은 그대로).
- R-4: 새 빌드를 리더가 설치한 뒤에야 레인 세션에 적용돼요.
- R-6(개정 4): 들어가는 `cd` 규칙도 R-2·R-5 의 읽지 않는 줄과 D-11 의 읽을 수 없는 목표(`cd "$W"`)는 통과시켜서 세션이 여전히 갇힐 수 있어요. 옛 빌드(R-4)도 같아요. 그래서 복구 문장(REQ-9)이 필요하고 AC-10 이 그 글을 시험해요.
- R-7(개정 4): 주 체크아웃 세션이 연결 워크트리를 셸 `cd` 로 읽는 일(예: 리더가 레인의 워크트리를 살펴보기)은 거절돼요. `git -C <경로>` 읽기, 절대 경로, 서브셸 `( cd … )` 가 대신이에요(Read forms 표).
- R-8(개정 4): 이미 갇힌 세션은 이 규칙으로 풀리지 않고, 훅은 명령 글만 읽어서 셸 폴더를 되돌리지 못해요. 복구는 세션을 다시 띄우는 것이고 글이 그렇게 말해요.
- R-9(개정 4): 목표 읽기는 글자 그대로의 경로만 따라가요. 링크를 거친 목표는 `checkout_of` 가 풀어서 보지만, 명령이 실행되는 사이에 바뀌는 폴더(경쟁 상태)는 판정하지 않아요.
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

### 개정 3: 거절 글의 cargo 꼴

배치 24 의 글 수정에 대한 리뷰 노트 10564 의 논블로킹 둘을 고쳤어요. 글이 권한 `cargo --manifest-path <path>/Cargo.toml` 은 cargo 가 거절하는 꼴이라(플래그는 부속 명령 뒤) `cargo test --manifest-path <path>/Cargo.toml` 로 고쳤고, AC-1 문구를 실제 글에 맞췄어요(D-8). 코드 head 1380fe8dfc(base 8933f238cd), 승인은 노트 10637(사용자가 직접 골랐어요, 2026-11-27).

- AC-1: met — `cargo nextest run --test hooks_cd_guard_cli -E 'test(a_worktree_session_cd_is_denied)'` 1/1 통과(종료 0, 개정 3 문구 그대로)
- AC-7: met — `cargo nextest run --test hooks_cd_guard_cli -E 'test(the_session_cd_refusal_cargo_form_is_one_cargo_accepts)'` 1/1 통과(종료 0). 글을 고치기 전에는 같은 테스트가 단언 실패, 종료 100(`tdd red` 기록, 커밋 4640b4dba4).
- AC-2, AC-3, AC-5, AC-6 와 `hooks_cd_guard_cli` 전체: `cargo nextest run --offline --locked --test hooks_cd_guard_cli --test hook_bash_boundary_cli` 45/45 통과(`hook_bash_boundary_cli` 38, `hooks_cd_guard_cli` 7), AC-4 `rg -c 'cd into it' src` 종료 1.
- 가드 뮤턴트 1개: 글의 `cargo test --manifest-path` 에서 `test` 를 지우면(Edit, 되돌림) `the_session_cd_refusal_cargo_form_is_one_cargo_accepts` 종료 100.

### 개정 4: 승인됨, 코드 쓰기 전

SPEC 글은 커밋 f47f5a33bb 이고 승인은 노트 10711(사용자가 직접 골랐어요, 개정 4 글 그대로)이에요. 승인 digest 는 그 커밋의 파일에서 `moai spec show` 로 읽은 값이에요. 이 줄을 쓸 때 코드와 시험은 아직 없어요. 개정 3 까지의 AC-1~AC-7 결과는 위 그대로이고 AC-3 문구만 좁혀졌어요.

- AC-8: not_run — 코드 없음(`a_main_checkout_cd_into_a_worktree_is_denied`)
- AC-9: not_run — 코드 없음(`a_main_checkout_cd_that_does_not_enter_a_worktree_is_allowed`)
- AC-10: not_run — 코드 없음(`the_session_cd_refusal_tells_a_moved_session_how_to_get_out`)
- AC-11: not_run — 코드 없음(`a_registered_session_is_not_exempt_from_the_entering_cd_rule`)
- AC-12: not_run — 코드 없음(`an_objects_only_git_command_toward_another_checkout_stays_denied`)
- AC-13: not_run — 코드 없음(`the_read_forms_stay_allowed`)
- 읽은 것(2026-11-27, NeXT 8b6ab17634): `src/hook_bash_boundary.rs` 의 `own_checkout`·`session_cd`·`moves_session`·`change_dir`·`classify_git`·`GIT_READ`·`GIT_WRITE`·`D5`, `src/hook_engine.rs` 의 `tool_preflight_deny`, `src/lane_registry.rs` 의 `SessionEntry`·`mode_of`, 시험 표 `tests/hooks_cd_guard_cli.rs`·`tests/hook_bash_boundary_cli.rs`, 등록부 항목 넷(`jq`, 읽기만).
- not_run: 설치된 훅에 실제 명령을 넣는 확인(호스트 가드가 `HOME` 을 바꾸는 호출을 거절했어요), 훅 프로세스의 `CLAUDE_PROJECT_DIR` 값, 훅 입력 `session_id` 와 등록부 `session` 의 꼴 비교.
