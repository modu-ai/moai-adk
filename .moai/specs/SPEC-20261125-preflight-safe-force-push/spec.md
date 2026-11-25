---
id: SPEC-20261125-preflight-safe-force-push
title: "tool-preflight 가 덮어쓸 수 없는 새 ref 전용 lease 푸시와 이 컴퓨터의 경로로 가는 강제 푸시에는 확인을 묻지 않아요"
status: draft
revision: 1
created: 2026-11-25
branch: spec/preflight-safe-force-push-20261125
---

# tool-preflight 가 덮어쓸 수 없는 새 ref 전용 lease 푸시와 이 컴퓨터의 경로로 가는 강제 푸시에는 확인을 묻지 않아요

## Intent

`moai tool-preflight` 의 확인 규칙(`bash_confirm`, `src/hook_guards.rs:288-296`)은 강제 표지가 하나라도 있으면 어떤 푸시든 확인을 물어요(`CONFIRM_FORCE_PUSH`, `src/hook_guards.rs:256`). 조사 에이전트가 임시 저장소에서 `--force-with-lease` 글만으로 질문이 뜨는 것을 보았고, 사용자가 2026-11-24T23:35:48Z 에 선택지 "안전한 경우만 통과"를 골랐어요(이 일의 카드 본문). 이 SPEC 은 실제 원격을 덮어쓸 수 없는 두 경우만 묻지 않게 해요. 실제 원격을 덮어쓰는 강제 푸시의 확인과 `main`/`master` 강제 푸시의 거절(`FORCE_PUSH`, deny)은 그대로예요. 권한 훅을 만지므로 tier 2 이고, 승인은 사용자 몫이에요.

git 2.54 를 임시 저장소(원격은 bare 저장소)에서 돌려 읽은 사실이에요(2026-11-25).

- `--force-with-lease=work: up work`: 원격에 `work` 가 없으면 만들어지고(`[new branch]`), 있고 기록이 다시 쓰였으면 `stale info` 로 거절돼요. 그러니 빈 기대값(`<ref>:`)의 lease 는 새로 만들기 전용이에요.
- lease 가 다른 ref 를 이름 지으면(`--force-with-lease=other: up work`) `work` 는 강제되지 않아 non-fast-forward 로 거절돼요. lease 는 이름 지은 ref 만 강제해요.
- `--mirror --force-with-lease=work: up` 은 원격의 `other` 를 지우고 ref 를 덮어요. 새로 만들기 전용 lease 가 있어도 `--mirror` 는 안전하지 않아요.
- `+work` 는 lease 와 상관없이 강제 갱신이에요(`forced update`).
- `git remote get-url --push --all <이름>` 은 `insteadOf` 바꿔 쓰기와 `pushurl` 을 거친 실제 푸시 주소를 주고, 없는 이름은 `No such remote` 로 실패해요.

## Non-goals

- `main`/`master` 강제 푸시 거절, `git reset --hard`·`git clean -f` 확인, 직접 PR 병합 확인(모드 층)은 바꾸지 않아요.
- 네트워크 드라이브에 마운트된 경로(NFS·SMB)는 이 컴퓨터의 경로로 읽어요. 마운트 여부는 판정하지 않아요(D-4).
- `remote.<이름>.push` 설정이나 `push.default` 가 만드는 강제, `GIT_CONFIG_*` 환경 변수, `git -C <폴더>`·`-c`·`--git-dir` 로 바뀐 저장소는 읽지 않고 물어요(REQ-3).
- 줄 안에 `cd`·파이프·환경 변수 앞붙임이 있는 줄의 원격 이름 조회: 어느 저장소의 설정을 읽어야 하는지 알 수 없어 물어요. 경로를 직접 적은 푸시는 폴더와 상관없이 판정해요.
- `~/` 로 시작하는 경로, 윈도 경로(`C:\`), `//host/share`: 이 컴퓨터의 경로로 읽지 않고 물어요.
- 새 모드(`[hooks.modes]`)를 만들지 않아요: 이 규칙은 기존 확인 규칙처럼 모드가 없어요.

## Requirements

- REQ-1: When a `git push` has every force marker in the create-only lease form `--force-with-lease=<ref>:` (a non-empty ref, an empty expected value, the option or a prefix git reads as it) and has none of `-f`, `--force`, `--force-if-includes`, a `+` refspec, `--mirror`, `--all`, `--branches`, `--prune` or `--delete`, the tool-preflight guard shall not ask for confirmation. · failure: none — a lease that is not create-only, or any other force marker beside it, is not this case and falls to the confirmation as before.
- REQ-2: When a forced `git push` names a remote whose every push URL is a path on this computer (`/…`, `./…`, `../…`, `file:///…`, `file://localhost/…`), the tool-preflight guard shall not ask for confirmation.
- REQ-3 (REQ-2): IF the remote cannot be judged to be on this computer (its name has no push URL, the lookup fails or prints nothing, any push URL is not one of those forms, the line holds another command, does not start with `git`, or `git` carries its own options before `push`), THEN the guard shall ask for confirmation as before.
- REQ-4: The guard shall ask for confirmation of every other forced `git push` as before.
- REQ-5: The guard shall refuse a force push to `main` or `master` as before, whatever its lease or remote.
- REQ-6: The `capabilities` answer of the hook engine shall state the two cases that are not asked about.

## Acceptance

- AC-1 (REQ-1): 새 ref 전용 lease(`--force-with-lease=work:`, 접두 `--force-w=work:`, 여러 개)는 묻지 않고, 값이 있는 lease(`=work:abc`)·값 없는 `--force-with-lease`·`-f`·`--force`·`--force-if-includes`·`+work`·`--mirror`·`--all`·`--prune`·`--delete` 와 섞이면 지금처럼 물어요 · verify: `cargo test --locked --offline --lib push_confirm_lease_`
- AC-2 (REQ-2, REQ-3): 경로를 직접 적은 푸시와 `get-url` 이 로컬 경로를 주는 원격 이름(임시 bare 저장소)은 묻지 않고, 네트워크 주소(`https://`, `ssh://`, `git@host:`, `git://`)·없는 원격 이름·`file://otherhost/`·`git -C x push`·`cd x && git push`·`FOO=1 git push` 는 물어요 · verify: `cargo test --locked --offline --lib push_confirm_remote_`
- AC-3 (REQ-1…REQ-5): 위험 경로를 실제로 돌려요. 임시 저장소에서 훅 명령(`moai hook-dispatch --host claude --event tool-preflight`)에 Bash 입력을 먹여, 묻지 않는 두 경우(`--force-with-lease=work:`, 로컬 bare 원격으로 `git push -f`)는 `ask` 가 없고, 네트워크 원격(`git remote add` 로 `https://` 주소)·값 있는 lease·`--mirror` 는 `ask`, `main` 으로의 강제 푸시는 `deny` 예요 · verify: `cargo test --locked --offline --test hooks_cli safe_force_push_`
- AC-4 (REQ-6): 세 host 의 `capabilities` 가 두 경우를 말해요 · verify: `cargo test --locked --offline --test hooks_cli safe_force_push_capabilities`
- AC-5 (REQ-1…REQ-6): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --lib hook_`
  verify: `cargo test --locked --offline --test hooks_cli`
  verify: `cargo test --locked --offline --test statusline_guard_false_positive`
  verify: `cargo test --locked --offline --test hook_dispatch_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-preflight-safe-force-push/spec.md`, `src/hook_guards.rs`, `src/hook_engine.rs`, `tests/hooks_cli.rs`
- read-only: `src/claim_verify.rs`, `src/identity.rs`, `tests/statusline_guard_false_positive.rs`, `tests/hook_dispatch_cli.rs`
- forbidden: 진짜 HOME 과 그 안의 Claude 설정 폴더, 진짜 원격(테스트는 임시 bare 저장소 로컬 경로만 써요), 거절 규칙(`FORCE_PUSH`)의 약화, `bash_confirm(command)` 의 공개 시그니처(다른 테스트가 써요)

## Plan

1. 테스트 먼저: `src/hook_guards.rs` 의 단위 테스트 표(`push_confirm_lease_`, `push_confirm_remote_`)와 `tests/hooks_cli.rs` 의 `safe_force_push_` 를 쓰고, 지금 코드에서 두 묻지 않는 경우가 `ask` 로 실패하는 것을 먼저 봐요.
2. `bash_confirm(command)` 는 그대로 두고(조회할 폴더가 없으니 원격 이름은 물어요), `bash_confirm_at(command, cwd)` 를 더해요. 엔진(`src/hook_engine.rs` 의 `tool_preflight`)은 입력의 `cwd` 로 이 함수를 불러요.
3. `push` 갈래에서 강제 표지가 있을 때만: 전부 새 ref 전용 lease 이면(REQ-1) 통과, 아니면 원격이 이 컴퓨터의 경로이면(REQ-2) 통과, 아니면 물어요. 원격 이름 조회는 `git remote get-url --push --all`(환경 정리 `scrub_git_env`, 입력 닫음, 오류 닫음)이고, 줄이 `git` 한 명령이고 전역 옵션이 없을 때만 해요.
4. `CONFIRM_COMMANDS` 문구에 두 경우를 더해 `capabilities` 가 말하게 해요.
5. 위험 경로 실행(AC-3)과 가드마다 뮤턴트 한 번(조건을 지우면 테스트가 실패하는지), 이웃 테스트, READY 에서 `check affected --card` 한 번.

## Risks

- 이 컴퓨터의 경로 판정은 글자 모양만 봐요. 네트워크 드라이브에 마운트된 경로로의 강제 푸시는 묻지 않게 돼요(D-4). 원격 이름은 설정의 바꿔 쓰기를 거친 `get-url` 결과를 써서, 이름이 로컬이어도 `insteadOf` 가 네트워크로 바꾸면 물어요.
- 훅이 `git` 자식 프로세스를 새로 띄워요. 읽기 전용 조회이고 네트워크를 쓰지 않아요. 저장소를 못 읽거나 `safe.directory` 때문에 실패하면 묻는 쪽으로 떨어져요(REQ-3).
- 묻지 않는 경우가 늘어나는 변경이라, 두 조건 중 하나를 잘못 읽으면 실제 원격 덮어쓰기가 확인 없이 지나가요. 그래서 판정이 불확실한 모든 경우를 묻는 쪽에 두고, 테스트에 묻는 표를 둬요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 묻지 않는 경우는 두 가지뿐: 새 ref 전용 lease, 이 컴퓨터의 경로로 가는 푸시 | 사용자 선택 2026-11-24T23:35:48Z "안전한 경우만 통과"(이 일의 카드 본문) | 모든 lease 통과, 모두 묻기 |
| D-2 | 새 ref 전용은 `--force-with-lease=<ref>:` 형태에 다른 강제·광범위 옵션이 없을 때만 | 임시 저장소 실험: 있는 ref 는 `stale info` 로 거절, `--mirror` 는 덮어요 | `--force-with-lease` 전부 통과 |
| D-3 | 원격 이름은 `git remote get-url --push --all` 로 풀고, 조회가 불확실하면 물어요 | 바꿔 쓰기와 `pushurl` 을 거친 실제 주소예요. 설정 파일을 직접 읽는 것보다 틀릴 곳이 적어요 | `.git/config` 직접 파싱 |
| D-4 | 마운트된 네트워크 경로도 로컬 경로로 읽어요(한계로 기록) | 마운트 판정은 플랫폼마다 달라 드문 경우를 위해 설계하지 않아요(규칙 20). 후속으로 남겨요 | 경로가 어느 파일시스템인지 읽기 |
| D-5 | `bash_confirm(command)` 시그니처를 두고 `bash_confirm_at` 를 더해요 | 다른 테스트 파일(`statusline_guard_false_positive`)이 공개 함수를 써요 | 시그니처에 cwd 추가 |

## Result

아직 구현 전이에요. 사용자의 SPEC 승인 뒤에 채워요.
