---
id: SPEC-20261125-preflight-safe-force-push
title: "tool-preflight 가 덮어쓸 수 없는 새 ref 전용 lease 푸시와 이 컴퓨터의 경로로 가는 강제 푸시에는 확인을 묻지 않아요"
status: approved
revision: 3
created: 2026-11-25
branch: spec/preflight-safe-force-push-r2-20261125
approval:
  by: leader
  at: 2026-11-26
  basis: "리더 위임 승인 노트 9562 (사용자 위임 2026-11-25 ~14:50Z \"나머진 네가 알아서 추천 해주는데로 선택해서\", 리더 판단이고 사용자 직접 선택은 아님; 방향은 사용자가 2026-11-24T23:35:48Z 에 고른 \"안전한 경우만 통과\"); 리더가 착지 직전 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:6d76e2813e844acc0acaaf48b65383d3675ab45583b64fb6e4161732ab019fac"
---

## Interview

- Q1: 덮어쓸 수 없는 강제 푸시에도 확인을 물을까요? → A: "안전한 경우만 통과" (사용자 선택, 2026-11-24T23:35:48Z; 이 SPEC 의 방향이에요)

# tool-preflight 가 덮어쓸 수 없는 새 ref 전용 lease 푸시와 이 컴퓨터의 경로로 가는 강제 푸시에는 확인을 묻지 않아요

## Intent

`moai tool-preflight` 의 확인 규칙(`bash_confirm`, `src/hook_guards.rs:288-296`)은 강제 표지가 하나라도 있으면 어떤 푸시든 확인을 물어요(`CONFIRM_FORCE_PUSH`, `src/hook_guards.rs:256`). 조사 에이전트가 임시 저장소에서 `--force-with-lease` 글만으로 질문이 뜨는 것을 보았고, 사용자가 2026-11-24T23:35:48Z 에 선택지 "안전한 경우만 통과"를 골랐어요(이 일의 카드 본문). 이 SPEC 은 실제 원격을 덮어쓸 수 없는 두 경우만 묻지 않게 해요. 실제 원격을 덮어쓰는 강제 푸시의 확인과 `main`/`master` 강제 푸시의 거절(`FORCE_PUSH`, deny)은 그대로예요. 권한 훅을 만지므로 tier 2 이고, 승인은 사용자 몫이에요.

git 2.54 를 임시 저장소(원격은 bare 저장소)에서 돌려 읽은 사실이에요(2026-11-25).

- `--force-with-lease=work: up work`: 원격에 `work` 가 없으면 만들어지고(`[new branch]`), 있고 기록이 다시 쓰였으면 `stale info` 로 거절돼요. 그러니 빈 기대값(`<ref>:`)의 lease 는 새로 만들기 전용이에요.
- lease 가 다른 ref 를 이름 지으면(`--force-with-lease=other: up work`) `work` 는 강제되지 않아 non-fast-forward 로 거절돼요. lease 는 이름 지은 ref 만 강제해요.
- `--mirror --force-with-lease=work: up` 은 원격의 `other` 를 지우고 ref 를 덮어요. 새로 만들기 전용 lease 가 있어도 `--mirror` 는 안전하지 않아요.
- `+work` 는 lease 와 상관없이 강제 갱신이에요(`forced update`).
- `git remote get-url --push --all <이름>` 은 `insteadOf` 바꿔 쓰기와 `pushurl` 을 거친 실제 푸시 주소를 주고, 없는 이름은 `No such remote` 로 실패해요.
- git 은 푸시 줄에 직접 적은 주소에도 `url.<base>.insteadOf` 와 `url.<base>.pushInsteadOf` 를 적용해요. 그래서 글자가 `/…` 인 주소도 설정이 그 접두를 네트워크 주소로 바꿔 쓰면 실제 푸시는 네트워크로 가요(revision 2 에서 더한 사실, 보안 리뷰 지적과 리더 결정 9318).
- 설정 `remote.<이름>.push=+refs/heads/*:refs/heads/*` 가 있으면 줄에 ref 를 한 단어(`main`)로 적은 푸시는 `--force-with-lease=main:`(빈 기대값, `main` 이 원격에 있음)가 붙어도 `main` 을 `forced update` 로 덮어요. 설정이 만든 강제 앞에서 lease 의 기대값 검사가 막지 못해요. 줄에 `src:dst` 를 적으면(`main:main`) 설정은 적용되지 않아요. `remote.<이름>.mirror=true` 는 줄에 refspec 이 없을 때 덮고, refspec 이 있으면 git 이 `--mirror can't be combined with refspecs` 로 죽어요. 설정이 없으면 같은 줄은 `non-fast-forward` 나 `stale info` 로 거절돼요(2026-11-26 git 2.54, 임시 bare 저장소 열 가지 줄, revision 3 에서 더한 사실, 감사 노트 9476 과 보안 리뷰).

## Non-goals

- `main`/`master` 강제 푸시 거절, `git reset --hard`·`git clean -f` 확인, 직접 PR 병합 확인(모드 층)은 바꾸지 않아요.
- 네트워크 드라이브에 마운트된 경로(NFS·SMB)는 이 컴퓨터의 경로로 읽어요. 마운트 여부는 판정하지 않아요(D-4).
- `remote.<이름>.push` 와 `remote.<이름>.mirror` 는 새 ref 전용 lease 에 한해 폴더의 설정에서 읽고, 하나라도 있거나 읽을 수 없으면 물어요(REQ-7). `+` 를 만드는 설정은 `remote.<이름>.push` 이고 `push.default` 는 읽지 않아요(강제 표지를 만든다는 근거를 못 찾았어요, 이 SPEC 에서 재지 않았어요: unverified). `GIT_CONFIG_*` 환경 변수, `git -C <폴더>`·`-c`·`--git-dir` 로 바뀐 저장소는 읽지 않고 물어요(REQ-3, REQ-7).
- 줄 안에 `cd`·파이프·환경 변수 앞붙임이 있는 줄의 원격 이름 조회: 어느 저장소의 설정을 읽어야 하는지 알 수 없어 물어요. 경로를 직접 적은 푸시도 같아요: 바꿔 쓰기 설정을 읽을 수 있을 때만 판정하고, 읽을 수 없으면 물어요.
- `~/` 로 시작하는 경로, 윈도 경로(`C:\`), `//host/share`: 이 컴퓨터의 경로로 읽지 않고 물어요.
- 새 모드(`[hooks.modes]`)를 만들지 않아요: 이 규칙은 기존 확인 규칙처럼 모드가 없어요.

## Requirements

- REQ-1: When a `git push` has every force marker in the create-only lease form `--force-with-lease=<ref>:` (a non-empty ref, an empty expected value, the option or a prefix git reads as it) and has none of `-f`, `--force`, `--force-if-includes`, a `+` refspec, `--mirror`, `--all`, `--branches`, `--prune` or `--delete`, names the remote and at least one refspec, and no `remote.<name>.push` or `remote.<name>.mirror` setting exists (REQ-7), the tool-preflight guard shall not ask for confirmation. · failure: none — a lease that is not create-only, or any other force marker beside it, is not this case and falls to the confirmation as before.
- REQ-2: When a forced `git push` names a remote that is a path on this computer (`/…`, `./…`, `../…`, `file:///…`, `file://localhost/…`) which no `url.<base>.insteadOf` or `url.<base>.pushInsteadOf` setting rewrites, or a remote name whose every push URL is such a path, the tool-preflight guard shall not ask for confirmation.
- REQ-3 (REQ-2): IF the remote cannot be judged to be on this computer (its name has no push URL, the lookup fails or prints nothing, any push URL is not one of those forms, a written path starts with a `url.<base>.insteadOf` or `url.<base>.pushInsteadOf` value, those settings cannot be read, there is no folder to read them in, the line holds another command, does not start with `git`, or `git` carries its own options before `push`), THEN the guard shall ask for confirmation as before.
- REQ-4: The guard shall ask for confirmation of every other forced `git push` as before.
- REQ-5: The guard shall refuse a force push to `main` or `master` as before, whatever its lease or remote.
- REQ-6: The `capabilities` answer of the hook engine shall state the two cases that are not asked about.
- REQ-7 (REQ-1): IF a `remote.<name>.push` or `remote.<name>.mirror` setting exists in the settings git reads in the folder, those settings cannot be read, there is no folder to read them in, the folder is no repository, the line holds another command or does not start with `git`, `git` carries its own options before `push`, or the hook's environment changes which settings git reads, THEN the guard shall ask for confirmation of a create-only lease push as before.

## Acceptance

- AC-1 (REQ-1, REQ-7): 폴더가 저장소이고 push·mirror 설정이 없을 때 새 ref 전용 lease(`--force-with-lease=work: up work`, 접두 `--force-w=work:`, 여러 개)는 묻지 않아요. 값이 있는 lease(`=work:abc`)·값 없는 `--force-with-lease`·`-f`·`--force`·`--force-if-includes`·`+work`·`--mirror`·`--all`·`--prune`·`--delete` 와 섞이면, refspec 이 없는 줄(`--force-with-lease=work: up`)이면, `remote.<이름>.push`·`remote.<이름>.mirror` 설정이 하나라도 있으면(값이 `+` 가 아니어도), 폴더 없이 부른 줄(`bash_confirm`)·저장소가 아닌 폴더·`git -c`·`git -C`·`cd x &&`·`FOO=1` 앞붙임이면 지금처럼 물어요 · verify: `cargo test --locked --offline --lib push_confirm_lease_`
- AC-2 (REQ-2, REQ-3): 폴더를 줄 때(`bash_confirm_at`) 경로를 직접 적은 푸시와 `get-url` 이 로컬 경로를 주는 원격 이름(임시 bare 저장소)은 묻지 않고, 네트워크 주소(`https://`, `ssh://`, `git@host:`, `git://`)·없는 원격 이름·`file://otherhost/`·`git -C x push`·`cd x && git push`·`FOO=1 git push`·그 접두와 맞는 `insteadOf` 나 `pushInsteadOf` 설정이 있는 경로(두 설정 모두)·폴더 없이 부른 경로 푸시(`bash_confirm`)·설정을 읽지 못하는 폴더는 물어요 · verify: `cargo test --locked --offline --lib push_confirm_remote_`
- AC-3 (REQ-1…REQ-5): 위험 경로를 실제로 돌려요. 임시 저장소에서 훅 명령(`moai hook-dispatch --host claude --event tool-preflight`)에 Bash 입력을 먹여, 묻지 않는 두 경우(`--force-with-lease=work:`, 로컬 bare 원격으로 `git push -f`)는 `ask` 가 없고, 네트워크 원격(`git remote add` 로 `https://` 주소)·값 있는 lease·`--mirror` 는 `ask`, `main` 으로의 강제 푸시는 `deny` 예요 · verify: `cargo test --locked --offline --test hooks_cli safe_force_push_`
- AC-4 (REQ-6): host `claude` 의 `capabilities` 가 두 경우를 말해요(Codex 와 OMP 는 2026-11-21 에 철회돼 host 는 하나예요) · verify: `cargo test --locked --offline --test hooks_cli safe_force_push_capabilities`
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
3. `push` 갈래에서 강제 표지가 있을 때만: 전부 새 ref 전용 lease 이고 줄이 원격과 refspec 을 적었고 푸시를 바꾸는 설정이 없다고 읽혔으면(REQ-1, REQ-7: 줄이 한 `git` 명령이고 폴더가 있고 전역 옵션이 없고 환경이 깨끗하고 저장소이고 `git config --null --get-regexp '^remote\..*\.(push|mirror)$'` 가 종료 1 에 빈 출력) 통과, 아니면 원격이 이 컴퓨터의 경로이면(REQ-2) 통과, 아니면 물어요. 원격 이름 조회는 `git remote get-url --push --all`(환경 정리 `scrub_git_env`, 입력 닫음, 오류 닫음)이고, 직접 적은 경로의 바꿔 쓰기 조회는 `git config --null --get-regexp '^url\..*\.(push)?insteadof$'` 예요. 둘 다 줄이 `git` 한 명령이고 전역 옵션이 없고 폴더가 있을 때만 해요. 아니면 경로도 물어요.
4. `CONFIRM_COMMANDS` 문구에 두 경우를 더해 `capabilities` 가 말하게 해요.
5. 위험 경로 실행(AC-3)과 가드마다 뮤턴트 한 번(조건을 지우면 테스트가 실패하는지), 이웃 테스트, READY 에서 `check affected --card` 한 번.

## Risks

- 이 컴퓨터의 경로 판정은 글자 모양과 바꿔 쓰기 설정의 접두만 봐요. 네트워크 드라이브에 마운트된 경로로의 강제 푸시는 묻지 않게 돼요(D-4). 원격 이름은 설정의 바꿔 쓰기를 거친 `get-url` 결과를 써서, 이름이 로컬이어도 `insteadOf` 가 네트워크로 바꾸면 물어요. 직접 적은 경로는 접두가 맞는 `insteadOf`·`pushInsteadOf` 가 있거나 설정을 읽을 수 없으면 물어요(D-6).
- 로컬 경로의 저장소로 가는 강제 푸시는 그 저장소에 다른 사람의 푸시 안 한 작업이 있으면 확인 없이 덮어요. 사용자가 고른 범위(D-1)라 이 SPEC 이 좁히지 않아요. 좁히려면 사용자 선택 항목으로 올려요(작업 폴더·임시 폴더 안만 / 지금대로).
- 새 ref 전용 lease 는 푸시를 바꾸는 설정이 없을 때만 통과해요(D-7). 설정을 파싱하지 않고 하나라도 있으면 물어서, `remote.<이름>.push` 를 쓰는 사용자의 lease 줄은 `+` 가 없어도 물어요. 사용자가 고른 "안전한 경우만"의 쪽으로 틀려요.
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
| D-6 | 직접 적은 경로도 글자 모양만 믿지 않아요: 폴더가 있고 한 `git` 명령일 때 `insteadOf`·`pushInsteadOf` 접두와 맞으면 물어요, 폴더가 없거나 설정을 읽지 못하면 물어요 | 리더 결정 노트 9318(2026-11-25, 사용자 위임·보안 리뷰 8897 #1): git 이 직접 적은 주소에도 바꿔 쓰기를 적용해 글자가 로컬이어도 네트워크로 갈 수 있어요. revision 1 의 "글자 모양으로 판정"을 대신해요 | 글자 모양만 판정(revision 1) |
| D-7 | 새 ref 전용 lease 는 줄이 원격과 refspec 을 적고, 줄이 한 `git` 명령·폴더·깨끗한 환경이고, `remote.*.push`·`remote.*.mirror` 설정이 하나도 없다고 읽혔을 때만 통과해요. 읽을 수 없으면 물어요 | 감사 FAIL 노트 9476 과 보안 리뷰, 진짜 git 2.54 측정(Intent): 설정 `+` 가 있으면 한 단어로 적은 ref 는 lease 가 붙어도 덮여요. "줄에 refspec 이 있으면 안전"(감사 제안)은 틀렸어요. 설정을 파싱하지 않고 존재만 봐서 틀릴 곳이 적어요 | refspec 이 있으면 통과(감사 제안), `src:dst` 형만 통과(대표 줄 `up work` 도 물어 카드 목적이 사라져요) |

## Result

구현 head 는 카드의 READY 노트에 있어요(영수증도 거기에 있어요). 아래 검사는 코드 head bcf2a24631 에서 SPEC 의 `verify:` 줄 그대로 돌렸고(revision 3, 그 뒤 커밋은 SPEC 문서만 바꿔요) 종료 코드는 모두 0이에요. 옛 head 92dafd5c51 의 위험 경로와 뮤턴트 11개 줄은 아래에 그대로 두고, revision 3 의 수리는 맨 아래 "수리" 줄에 있어요.

- AC-1 met: `cargo test --locked --offline --lib push_confirm_lease_` 2개 통과(`push_confirm_lease_only_the_create_only_form_is_not_asked`, `push_confirm_lease_is_asked_when_settings_can_force`).
- AC-2 met: `cargo test --locked --offline --lib push_confirm_remote_` 2개 통과(`push_confirm_remote_on_this_computer_is_not_asked`, `push_confirm_remote_written_path_is_judged_by_insteadof`). revision 2 대로 직접 적은 로컬 경로는 `insteadOf`·`pushInsteadOf` 설정이 그 접두로 시작할 때, 폴더가 저장소가 아닐 때, 폴더가 없을 때(`bash_confirm`) 물어요. verify 줄에 없는 가드 테스트 `push_confirm_env_names_that_change_git_settings_are_found` 1개를 더했고 `--lib push_confirm_` 는 모두 4개 통과예요.
- AC-3 met: `cargo test --locked --offline --test hooks_cli safe_force_push_` 6개 통과(revision 3 의 `safe_force_push_lease_is_asked_when_settings_force` 포함). 위험 경로는 임시 저장소와 로컬 bare 원격에서 진짜 `git push` 와 진짜 `moai-cli hook-dispatch --host claude --event tool-preflight` 명령(head 92dafd5c51 빌드)으로 돌렸어요. 새 ref 전용 lease 는 새 ref 에 성공했고, 그 ref 가 생긴 뒤 같은 lease 는 git 이 `stale info` 로 거절했고, `git push -f` 는 bare 경로로 성공했어요. 훅 답은 결정 없음이 `--force-with-lease=work: loc work`, `--force-with-lease=work: net work`, `-f loc work`, `-f <bare 경로> work` 이고, `ask` 가 `-f net work`, `--force-with-lease=work:abc net work`, `--force-with-lease net work`, `-f --force-with-lease=work: net work`, `--force-with-lease=work: --mirror net`, `cd <repo> && git push -f loc work`, `insteadOf` 규칙이 접두로 잡는 직접 적은 경로, 훅 환경에 `GIT_CONFIG_COUNT` 가 있는 모든 푸시이고, `deny` 가 `-f loc main`, `--force-with-lease=main: loc main` 이에요. bypassPermissions 세션에서는 `-f net work` 가 `deny` 이고 안전한 두 경우는 결정이 없어요. 임시 폴더는 지웠어요.
- AC-4 met: `cargo test --locked --offline --test hooks_cli safe_force_push_capabilities` 1개 통과. host 는 `claude` 하나뿐이에요(Codex 와 OMP 는 2026-11-21 에 철회). revision 2 가 옛 "host 세 개" 문구를 바로잡았어요.
- AC-5 met: `--lib hook_` 68개, `--test hooks_cli` 61개, `--test statusline_guard_false_positive` 3개, `--test hook_dispatch_cli` 53개 통과.
- 엔진: `src/hook_engine.rs` 의 `tool_preflight` 가 `bash_confirm_at(command, Some(&dir))` 를 부르고 `CONFIRM_FORCE_PUSH` 의 bypass 모드 deny 는 그대로예요. `CONFIRM_COMMANDS` 가 두 경우를 말해요.
- 자동 보안 리뷰 뒤에 더한 것(Non-goals 의 `GIT_CONFIG_*` 줄 안): `remote_is_on_this_computer` 는 훅 환경에 `GIT_CONFIG_COUNT`·`PARAMETERS`·`GLOBAL`·`SYSTEM`·`NOSYSTEM`, `GIT_CONFIG`, `GIT_DIR`, `GIT_COMMON_DIR`, `GIT_WORK_TREE`, `GIT_CONFIG_KEY_n`·`VALUE_n` 이 있으면 물어요(조회는 이 값들을 지우지만 실제 푸시는 지우지 않아서예요). 테스트는 `safe_force_push_env_that_changes_config_is_asked` 예요.
- 지킴이 뮤턴트 11개(한 줄씩 정확히 고치고, 하나씩 돌리고, git 으로 되돌리고, 끝난 뒤 트리가 깨끗함)가 모두 컴파일 오류가 아니라 단언 실패로 죽었어요. M1 넓히는 옵션 검사(mirror·all·branches·prune·delete), M2 `+` 참조 명세 검사, M3 lease 기대값이 비었는지 검사, M4 `-f`·`-d` 글자, M5 `git` 한 명령 조건, M6 `git` 전역 옵션 없음 조건, M7 경로의 `//` 검사, M8 폴더의 저장소 검사, M9 `insteadOf` 접두 맞춤, M10 lease ref 이름이 비지 않았는지 검사는 `--lib push_confirm_` 로 죽었고, M11 은 환경 가드 호출을 지운 것이라 `--test hooks_cli safe_force_push_env_that_changes_config_is_asked` 로 죽었어요.
- 수리(revision 3, 감사 FAIL 노트 9476 과 REPAIR-FINDING 노트 9512, 코드 커밋 3e4e6e8c6a 와 bcf2a24631): 감사는 refspec 이 없는 lease 줄이 설정이 만든 강제를 지나친다고 보았고 `positionals.len() >= 2` 를 제안했어요. 그 조건은 넣었지만(`create_only_lease_push`) 진짜 git 2.54 로 재니 충분하지 않았어요(`target/danger/probe.sh`, 일회용 로컬 bare): 설정 `remote.loc.push=+refs/heads/*:refs/heads/*` 에서 `--force-with-lease=main: loc main` 이 `main` 을 덮고(H), 평범한 `git push loc main` 도 덮고(C), `src:dst` 를 쓴 줄(`main:topic`, `main:main`)은 설정이 적용되지 않았어요(D, J). 설정이 없으면 같은 줄은 거절돼요(A, E, F, G). 그래서 REQ-7 과 D-7: 새 ref 전용 lease 는 `push_settings_may_force` 가 거짓일 때만 통과해요(한 `git` 명령, 폴더, 전역 옵션 없음, 깨끗한 환경, 저장소, `git config --null --get-regexp '^remote\..*\.(push|mirror)$'` 가 종료 1 에 빈 출력). 위험 경로(`target/danger/run2.out`, 복원한 소스로 재빌드한 `moai-cli`, 진짜 훅 명령): 설정이 없으면 `--force-with-lease=work: net work` 는 결정 없음이고, `remote.net.push=+refs/heads/*:refs/heads/*` 와 `remote.net.mirror=true` 에서는 네 줄(`net work`, `net`, `-u net`, 원격 없음)이 모두 `ask`, 설정을 지우면 다시 결정 없음, 훅 환경의 `GIT_CONFIG_COUNT` 와 `git -c remote.net.push=… push` 는 `ask` 예요. 뮤턴트(`target/repair/summary.txt`, 한 줄씩 정확히 고치고 하나씩 돌리고 git 으로 되돌림, 모두 단언 실패 exit 101, 끝난 뒤 트리 깨끗함): R1 `positionals.len() < 2` 조건 삭제는 단위 시험 `push_confirm_lease_only_the_create_only_form_is_not_asked` 가 죽였고(hooks_cli 시험은 설정 가드가 먼저 물어서 R1 만으로는 못 잡아요), R4 설정 가드 호출 삭제는 단위 시험 둘과 hooks_cli `safe_force_push_lease_is_asked_when_settings_force` 가, R5 종료 코드 1 판단 삭제는 단위 `push_confirm_lease_is_asked_when_settings_can_force` 와 같은 hooks_cli 시험이, R6 저장소 확인, R7 전역 옵션 조건은 같은 단위 시험이, R8 환경 검사는 `safe_force_push_env_that_changes_config_is_asked` 가, R2 환경 이름 목록은 `push_confirm_env_names_that_change_git_settings_are_found` 가, R3 푸시 URL 의 로컬 모양 검사는 `safe_force_push_remote_name_is_judged_by_its_push_url` 이 죽였어요. 보안 리뷰 두 건 중 "줄에 refspec 이 있어도 설정이 강제한다"는 맞았고(위 측정 H) 고쳤어요. "mirror 설정과 refspec"은 git 이 `--mirror can't be combined with refspecs` 로 죽어서 덮이지 않지만(위험 경로 run2.out), 이제 설정 존재만으로 물어요.
- not_run과 빈틈: 전체 `cargo test` 와 clippy 는 이 Result 에 없어요(READY 영수증의 `check affected --card` 가 맡아요). 진짜 원격과 진짜 HOME 은 쓰지 않았어요. 네트워크 원격으로의 `git push --mirror`(강제 표지 없음)는 이 변경 전에도 묻지 않았고(`push_force_option` 은 NeXT 끝의 같은 함수예요) 후속 후보예요. 이 컴퓨터의 저장소로 가는 강제 푸시는 다른 사람의 푸시 안 한 작업을 묻지 않고 덮을 수 있어요. 사용자가 고른 범위예요(D-1, 보안 리뷰 노트 8897 #2). 마운트된 네트워크 경로는 로컬로 읽어요(D-4).
