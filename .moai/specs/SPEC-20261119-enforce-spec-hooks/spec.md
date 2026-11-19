---
id: SPEC-20261119-enforce-spec-hooks
title: "The GREEN test lock also holds against Bash deletions and moves, and a denied call leaves its reason in the audit line"
status: draft
revision: 3
created: 2026-11-19
---

# The GREEN test lock also holds against Bash deletions and moves, and a denied call leaves its reason in the audit line

## Intent

카드 C-86 은 "지침 파일은 권고일 뿐이니 강제하라"(사용자 2026-11-17)는 큰 카드이고, 코드를 읽어 보니 AC 대부분이 이미 부분 구현돼 있어요(card note 2277). 이 SPEC 은 그 첫 조각(S1)만 해요: 카드 AC-2 의 둘째 절반(RED 기록이 적은 시험 파일을 GREEN 동안 편집·삭제하지 못하게)에서 지금 뚫려 있는 길 하나와, AC-5 의 감사 로그 빈틈 하나예요.

- `tdd::lock_refusal` 은 `Write`·`Edit`·`MultiEdit`·`NotebookEdit`·`apply_patch`(Delete 포함)만 막아요. Bash 로 `rm tests/x.rs`, `git rm`, `mv`, `unlink` 를 하면 잠긴 시험이 GREEN 동안 사라지거나 바뀌어도 hook 이 몰라요(ImpossibleBench 가 말한 "시험을 고쳐서 통과"의 쉬운 길).
- `record_deny` 가 deny 마다 `hook-engine.jsonl` 에 한 줄을 남기지만 어느 guard 가 막았는지(이유)는 없어요.

## Non-goals

- 카드의 AC-1(인터뷰 기록), AC-3(Stop hook), AC-4(병합 관문): 각각 별도 SPEC·PR 이에요. AC-1·AC-3 은 이 카드의 쓰기 범위 밖 파일이 필요해 범위 요청을 했어요(card note 2278 D-B).
- "SPEC 이 approved 여야 쓰기 허용"으로 조이기(AC-2 의 첫 절반): 지금 hook 은 SPEC 파일이 있으면 통과해요. 승인은 사용자의 대화 중 선택만 줄 수 있어서 조이면 모든 레인이 멈춰요. 정책 결정이라 리더에게 물었고(note 2278 D-A) 기본값은 그대로 두기예요.
- 셸이 파일을 바꾸는 나머지 길은 이 조각도 막지 못해요. 이름으로: 리다이렉트(`>`), `tee`, `sed -i`, 스크립트(`bash cleanup.sh`, `python3 -c …`), 변수로 만든 경로(`F=…; rm $F`), `xargs`, `find -exec`·`find -delete`, 확장 글롭(`@(..)`, `+(..)`), `~` 확장, 중괄호 256개 초과·4단계 초과, `cd` 로 폴더를 바꾼 뒤의 상대 경로. REQ-6 이 `capabilities` 에 같은 이름들을 적고, 시험 `green_lock_bash_named_limits_are_not_guarded` 와 `green_lock_bash_brace_limits_are_the_named_ones` 가 이 형태들이 지금 통과한다는 사실을 고정해요(하나를 막게 되면 시험과 문구를 같이 바꿔요). `!`, `{ …; }`, `if`·`while`·`for` 몸통, 백틱, `busybox`, `git -C dir`, 작업 트리 위 폴더는 이제 읽어요(REQ-10).
- 실제 `~/.claude` 설정과 hook 의 변경, OMP(편집 경로를 번역하지 않아요).
- `src/tdd.rs` 의 잠금 규칙 바꾸기: 이 조각은 `lock_refusal` 을 그대로 불러요.

## Requirements

- REQ-1: While a worktree's TDD record is in GREEN or REFACTOR, the preflight guard shall refuse a Bash command whose `rm`, `unlink` or `git rm` names a test path the record locks, or whose `mv` or `git mv` names such a path as a source or as a destination, with the reason the lock already gives.
- REQ-2: IF the target of `rm`, `unlink` or `git rm` does not exist, THEN the guard shall not refuse the command for that target. · failure: none — removing nothing changes no test, and the lock's rule against a new test file is not meant for it.
- REQ-3: Where the worktree has no TDD record, the record is in RED, no named path is locked, or the repository holds no MoAI state, the guard shall allow the command as before.
- REQ-4: The guard shall read relative paths against the hook input's `cwd` and judge a path where it lands (`..` and symlinks resolved), as the file-write lock does.
- REQ-5: When the engine denies a tool call, the line it appends to `hook-engine.jsonl` shall carry the reason sentence of the guard that refused it and no command text, path or file content.
- REQ-6: The `capabilities` answer of the hook engine shall say that the Bash lock covers `rm`, `unlink`, `git rm`, `mv` and `git mv` on literal, glob and brace words, on folders and behind the shell structures it reads, and name each form it does not guard: redirections, `tee`, `sed -i`, scripts, variables, `xargs`, `find -exec`, `find -delete`, extended globs, `~` expansion, more than 256 brace alternatives, braces nested deeper than 4 and a command run after a `cd`.
- REQ-8: The guard shall read a folder that exists and holds a locked test as that test, and a glob (`*`, `?`, `[..]`) or brace (`{a,b}`) word as the locked tests it names or holds.
- REQ-9: The guard shall read each word of those commands as written as well as matched or opened, so that a locked test whose real name holds `[..]`, `{a,b}`, `*` or `?` stays protected.
- REQ-10: The guard shall read those commands also behind `!`, a `{ …; }` group, the `then`, `do` and `else` bodies of `if`, `while` and `for`, backticks, `busybox` or `toybox`, and `git -C <dir>`, and as holding the locked tests when the folder is above the work tree.
- REQ-7 (REQ-5): IF a refusal's sentence is built at run time (the commit-identity rule names an email, the claim rule an Issue number, and a command line that commits in more than one folder names the verbs), THEN the line that REQ-5 describes shall carry `reason: null`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): GREEN 기록이 있는 임시 저장소에서 `rm`·`unlink`·`git rm`·`mv`·`git mv` 가 잠긴 시험 경로를 건드리면 deny, 잠기지 않은 경로·RED 단계·기록 없음·MoAI 상태 없는 저장소·존재하지 않는 `rm` 대상은 allow, 상대 경로와 `..` 는 착지한 곳으로 판단.
  verify: `cargo test --locked --offline --test hooks_cli green_lock_bash`
- AC-2 (REQ-5, REQ-7): deny 한 줄에 reason 이 있고 명령·경로·내용은 없어요; 실행 중에 만든 문장(여러 폴더에 커밋하는 줄)은 `reason: null`.
  verify: `cargo test --locked --offline --test hooks_cli deny_audit_line`
- AC-3 (REQ-6): `capabilities` 가 Bash 잠금의 범위와 못 막는 형태를 이름으로 말하고, 그 형태들이 실제로 통과한다는 사실을 시험이 고정해요.
  verify: `cargo test --locked --offline --test hooks_cli capabilities_name_the_bash_lock_limits`; `cargo test --locked --offline --test hooks_cli green_lock_bash_named_limits`
- AC-5 (REQ-8, REQ-9): 폴더째(`rm -rf tests`, `git rm -r tests`, `mv tests …`), glob(`tests/locked_*.rs`, `t*`, 클래스의 양 끝, `*` 가 빈 글자, 심볼릭 링크 폴더), 중괄호(`tests/{a,b}`)는 deny, 잠긴 시험이 없는 폴더·패턴은 allow. 이름에 `[id]`·`{b,c}`·`*`·`?` 가 든 잠긴 시험은 쓴 그대로도 deny(재감사 B1). `--` 뒤의 `-` 로 시작하는 이름도 경로예요(감사 N1).
  verify: `cargo test --locked --offline --test hooks_cli green_lock_bash_covers`; `cargo test --locked --offline --test hooks_cli green_lock_bash_reads_locked_names`; `cargo test --locked --offline --test hooks_cli green_lock_bash_edges`; `cargo test --locked --offline --test hooks_cli green_lock_bash_reads_dash`
- AC-6 (REQ-10): `! rm`, `{ rm …; }`, `if`·`while`·`for` 몸통, 백틱, `busybox rm`, `git -C tests rm`, 작업 트리 위 폴더(`rm -rf <부모>`)는 deny, 같은 모양의 안 막혀야 할 줄(`echo '{ rm …; }'`, `git -C tests status`)은 allow.
  verify: `cargo test --locked --offline --test hooks_cli green_lock_bash_reads_wrapped
- AC-4 (REQ-1…REQ-6): 이웃 시험.
  verify: `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test spec_before_code_hook`; `cargo test --locked --offline --test green_test_lock`; `cargo test --locked --offline --lib hook_guards::`; `cargo test --locked --offline --lib hook_engine::`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/hook_guards.rs`, `src/hook_engine.rs`, `tests/hooks_cli.rs`, `.moai/specs/SPEC-20261119-enforce-spec-hooks/spec.md`
- read-only: `src/tdd.rs`, `src/claim_verify.rs`, `src/hooks.rs`, `tests/green_test_lock.rs`, `tests/spec_before_code_hook.rs`, `skills/moai/references/internal/sdd.md`
- forbidden: the real host settings and MoAI state under the user's home (tests use a temporary HOME), `src/tdd.rs`, `src/spec_doc.rs`, any change that lets a hook allow what it refused before

## Plan

1. 시험 먼저(`tests/hooks_cli.rs`): GREEN 기록(`.git/moai/tdd.json`)을 임시 저장소에 직접 쓰고 `moai hook-dispatch --host claude --event tool-preflight` 로 Bash 입력을 먹여 deny/allow 를 읽어요. 없는 함수·필드 때문에 deny 시험이 실패하는 것(RED)을 봐요.
2. `bash_lock_guard(command, cwd)`: `shell_commands` 로 `rm`·`unlink`·`git rm`·`mv`·`git mv` 의 경로를 모아 `tdd::lock_refusal` 에 넘기고, `hook_engine` 의 Bash 분기에서 `bash_guard` 다음에 불러요.
3. `record_deny` 줄에 `reason`.
4. `capabilities` 문구.
5. 이웃 시험, mutation(가드마다 한 번).

## Risks

- 셸 파싱은 한계가 있어요(Non-goals). 이 조각은 `rm`·`unlink`·`git rm`·`mv`·`git mv` 를 글자 그대로·glob·중괄호·폴더째로 읽어 막고, 못 막는 형태는 Non-goals 에 이름으로 적고 `capabilities` 가 같은 이름으로 말하며, 그 형태들이 통과한다는 사실은 시험이 고정해요(첫 감사 #3216 B1 이 "막는 것처럼 읽힌다"고 지적해 고친 부분). 이 이름들 밖의 우회 길이 더 있을 수 있어요(찾지 못했을 뿐이에요).
- `lock_refusal` 의 둘째 조건("새 시험 파일")은 존재하지 않는 경로를 거부해요. `rm` 대상이 없는 경우를 건너뛰지 않으면 `rm -f tests/gone.rs` 가 헛거부돼요(REQ-2).
- deny 줄에 필드를 더하면 줄 모양을 고정한 시험이 깰 수 있어요(AC-4 가 이웃 시험으로 확인해요).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | `lock_refusal` 을 바꾸지 않고 불러요 | 규칙은 한 곳(`tdd.rs`)에 있어야 하고 이 카드의 쓰기 범위 밖이에요 | Bash 용 규칙을 따로 만들기(두 곳이 어긋남) |
| D-2 | 존재하지 않는 `rm` 대상은 건너뛰어요 | 지울 게 없으니 시험이 바뀌지 않아요 | 모두 `lock_refusal` 에 넘기기(헛거부) |
| D-3 | 감사 줄에는 고정 reason 문장만 더하고 명령·경로는 안 남겨요; 실행 중에 만든 문장(신원·claim·여러 폴더 커밋)은 `reason: null` | 고정 문장은 비밀·경로가 없고, 실행 중에 만든 문장은 이메일·Issue 번호처럼 호출에서 온 글을 품을 수 있어요 | 모든 reason 을 그대로 기록(새어 나감), 명령 해시 |
| D-4 | "approved 여야 쓰기 허용"은 이 조각에서 안 해요 | 정책 결정(note 2278 D-A, 기본값 그대로) | 선택 B 또는 C |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. macOS arm64, 이 worktree(`spec/20261119-enforce-spec-hooks`), 빌드 폴더 `~/.moai/build/moai-adk-86`, `CARGO_INCREMENTAL=0`, cargo cache `/tmp/moai-rc-cargo-agent-3`. 이 절은 세 번째 수리(커밋 2ed3a16f93 뒤)의 값이에요. 시험 개수는 `cargo test --test <이름>` 을 파일마다 돌려 그 파일 로그의 `test result` 줄에서 읽었어요.

RED 먼저: 첫 구현은 새 시험 5개 중 4개가 실패한 뒤 통과했어요. 세 번째 수리는 소스 변경을 치워 두고 새 시험을 돌려 `capabilities_name_the_bash_lock_limits`, `green_lock_bash_edges_of_folders_globs_and_moves`, `green_lock_bash_named_limits_are_not_guarded`, `green_lock_bash_reads_locked_names_with_pattern_characters`, `green_lock_bash_reads_wrapped_grouped_and_rebased_commands` 5개가 실패하는 것을 본 뒤(`target/c86/red3.log`) 구현했어요. 두 번째 수리의 시험은 구현 뒤에 썼고 RED 를 보지 못했어요.

- AC-1: met — `cargo test --locked --offline --test hooks_cli green_lock_bash` 통과(GREEN·REFACTOR 에서 `rm`·`unlink`·`git rm`·`mv`·`git mv`·`..`·절대 경로·`sh -c` 거부; 잠기지 않은 경로·없는 `rm` 대상·RED·기록 없음·`.git` 없음 허용; 입력 `cwd` 기준).
- AC-2: met — `deny_audit_line_carries_the_reason_and_no_command_or_path` 통과(두 deny 줄에 서로 다른 고정 reason, 명령·경로 문자열 없음)와 `deny_audit_line_is_null_when_the_sentence_is_built_at_run_time` 통과(`git -C a commit … && git -C b commit …` 거절은 `reason: null`). 신원·claim 규칙의 `null` 경로는 시험이 없어요(설정·lane 상태가 필요해 돌리지 않음). 재감사가 지적한 opt-in 규칙의 `reason` 변이용 시험도 더하지 않았어요(not_run).
- AC-3: met — `capabilities_name_the_bash_lock_limits` 통과(claude·codex·omp 세 host 가 막는 형태와 못 막는 형태 13개의 이름을 말해요: 리다이렉트, `tee`, `sed -i`, 스크립트, 변수, `xargs`, `find -exec`, `find -delete`, 확장 글롭, `~` 확장, 중괄호 256개 초과, 중괄호 4단계 초과, `cd` 뒤 명령)와 `green_lock_bash_named_limits_are_not_guarded`·`green_lock_bash_brace_limits_are_the_named_ones` 통과(이 이름마다 지금 통과하는 줄이 하나 이상 고정돼 있어요). `case` 문은 문구에서 뺐어요: 단순한 `case x in x) rm tests/…;; esac` 가 실제로 거절돼서 "막지 않음"이라고 쓰면 틀려요.
- AC-4: met — `--test hooks_cli` 35, `spec_before_code_hook` 9, `green_test_lock` 14, `hook_dispatch_cli` 56, `hook_status_cli` 10, `ac_command_lint` 5, `--lib hook_` 43 모두 통과, `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `python3 -m unittest tests.ci_scope_test` OK, `moai spec check SPEC-20261119-enforce-spec-hooks` 오류 0·경고 0. 돌리지 않은 것: 전체 `cargo test`(CI 가 돌려요), `--test git_env_scrub`(git 을 새로 띄우는 코드가 없어요), codex·omp host 실행.
- AC-5: met — `green_lock_bash_covers_folders_globs_and_braces`, `green_lock_bash_reads_locked_names_with_pattern_characters`(재감사 표 7행 모두와 같은 모양의 변형: 따옴표 없는 `[id]`, `\[id\]`, `*`·`?` 가 든 이름, 새 시험 경로 `new_[1]_test.rs`·`new_{1,2}_test.rs`), `green_lock_bash_edges_of_folders_globs_and_moves`(작업 트리 루트 폴더, 클래스 양 끝, `*` 가 빈 글자, 심볼릭 링크 폴더, `git mv` 의 새 시험 경로), `green_lock_bash_reads_dash_names_after_double_dash` 통과.
- AC-6: met — `green_lock_bash_reads_wrapped_grouped_and_rebased_commands` 통과(`! rm`, `{ rm …; }`, 괄호 묶음, `if`·`while`·`for` 몸통, 함수 정의, 백틱, `busybox`·`git -C` 의 `rm`·`mv`, 작업 트리 위 폴더 deny; `echo '{ rm … }'`, `echo "then rm …"`, `git -C tests status`, `busybox ls` allow). `rm -rf .`·`..`·`/` 는 정적 `rm` 규칙이 먼저 거절하고 시험이 그 문장을 확인해요. 작업 트리 위 폴더 시험은 작업 트리를 임시 폴더 세 단계 아래(`outer/inner/repo`)에 두어요: Linux 의 임시 폴더는 `/tmp/.tmpXXXX` 라 부모가 최상위 폴더 `/tmp` 가 되어 정적 규칙이 먼저 거절했고(CI run 44446207754, 실패 한 건: `rm -rf /tmp`), 한 폴더만 더 내려가면 정적 규칙을 통과하는 것을 빌드한 binary 로 확인했어요(`rm -rf /tmp` 만 거절, `/tmp/zz`·`/tmp/zz/outer/inner` 통과). Linux 에서 그 시험을 직접 돌리지는 않았어요(CI 가 돌려요).
- mutation: 처음 12개(수리 전 코드)와 첫 수리의 11개는 모두 KILLED(`target/c86/mutants.txt`, `mutants2.txt`). 세 번째 수리 코드에 12개(작업 트리 위 폴더, 중괄호 펼침, 역슬래시, 백틱, `!`, `{`, `then/do/else`, 구조어 전체, `busybox`, `git -C`, 글롭, 원래 철자)를 돌려 10개 KILLED, 2개(역슬래시·원래 철자)는 SURVIVED 였고 동작이 같은 중복 코드(토크나이저가 이미 역슬래시를 벗기고, 쓴 그대로의 단어는 다른 경로로 이미 들어가요)라 코드를 지웠어요(`target/c86/mutants3-s1.txt`, 지운 뒤 이 둘은 다시 돌리지 않았고 위 시험은 다시 통과). 재감사가 살아남았다고 한 N3 의 변이(작업 트리 루트, 클래스 위쪽 끝, `*` 빈 글자, 심볼릭 링크 접두사, `git mv` 새 경로, 중괄호 상한 8)는 시험을 더했지만 그 변이를 다시 돌리지는 않았어요(not_run).
- 이 조각 밖: AC-1(인터뷰 기록)·AC-3(Stop hook)·AC-4(병합 관문)은 S2·S3·S4 로 이어져요(리더 note 2278: 순서 S1 → S4 → S2 → S3).
