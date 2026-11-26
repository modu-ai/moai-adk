---
id: SPEC-20261124-glm-key-scope
title: "The glm key pick resolves its scope through key_vault and answers a Shared key outside a project"
status: approved
revision: 1
created: 2026-11-24
parent: SPEC-20261121-glm-key-at-pick
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8521 (2026-11-24T11:44:07Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 8520 에 답함: 개정 1 승인, D-1·D-2 그대로, 기준 1 해석 확인); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:db52bb981f895d4acf0874d4b6bdbb09fe59c2d114d67520bcd89ee981895c3b"
---

# The glm key pick resolves its scope through key_vault and answers a Shared key outside a project

## Intent

카드 C-457 은 C-248(SPEC-20261121-glm-key-at-pick)의 READY 노트 8362 와 감사 N6(노트 8382, 8487)이 남긴 두 후속이에요.

1. `src/model_effort_questions.rs:620` 의 `glm_key_scope` 는 `src/key_vault.rs:203` 의 `project_scope` 와 같은 계산(identity 의 `clone_id`·`clone_anchor`·`project_id` 를 읽어 `folder_name` 으로 만드는 14줄)을 복제해요. 이유는 `project_scope` 가 `identity(dir)` 를 불러서 프로세스 HOME 을 읽기 때문이에요(SPEC-20261124-identity-home-arg 가 `identity_in(path, home)` 를 더했어요). 두 곳이 어긋나면 모델 단계가 런처가 읽지 않는 폴더에 키를 저장해요. `key_vault` 에 HOME 을 받는 `project_scope_in` 을 두고 복제를 없애요.
2. 프로젝트 밖(git 작업 트리가 아닌 곳)에서는 모델 단계가 키 질문을 건너뛰고 `moai key-set glm --prompt` 안내 한 줄만 보여요. 설치 흐름은 보통 프로젝트 밖에서 돌아서 설치 보고서 G8(glm 을 고르면 그 자리에서 키를 받아요)의 일부가 비어요. 프로젝트 밖에서는 Shared 범위(런처가 프로젝트 항목 다음에 읽는 폴더, `moai key-set --shared` 가 쓰는 곳)에 저장해요.

## Interview

- Q1: 프로젝트 밖에서는 키를 어디에 저장하나요? → A: 질문을 하고 Shared 범위에 저장해요. 질문에 '모든 프로젝트가 같이 쓰는 키'라는 말을 넣고 `moai key-remove glm --shared` 로 되돌려요(리더 결정 노트 8521, D-1).
- Q2: identity 를 풀지 못하면요? → A: 지금처럼 건너뛰어요. 런처가 그 경우 Shared 키를 읽지 못해요(노트 8521, D-2).
- Q3: 기준 1 의 "C-248 키 시험은 그대로"는 무슨 뜻인가요? → A: 단언은 그대로 두고, `glm_key_scope` 호출 세 곳만 기대 범위를 `project_scope_in` 으로 계산하게 바꿔요(노트 8521).
- Q4: 모델 단계와 시험 세 개가 C-449 예약에 있는데요? → A: 리더가 C-449 에 give-back 을 요청했고, 돌려받으면 `card scope` 로 넓혀요. 그동안 key_vault 쪽부터 해요(노트 8521, 8522).

## Non-goals

- 키 보관소의 저장 방식, 폴더 이름 규칙(`folder_name`)과 `project_scope` 의 답 변경: 같은 입력에 같은 답이에요.
- 프로젝트 안에서의 동작 변경: 프로젝트 항목에 저장하는 것은 그대로예요.
- `identity` 를 풀지 못하는 오류(깨진 git, 거절된 링크)를 Shared 로 돌리는 것: 런처의 `read_here` 가 같은 경우에 Shared 를 읽지 않고 거절하므로 Shared 에 넣어도 런처가 못 찾아요. 이 경우는 지금처럼 건너뛰기예요.
- 질문 문장의 다국어, 첫 `moai glm` 실행의 `key_absent` 동작.

## Requirements

- REQ-1: The `key_vault::project_scope_in` function shall resolve the project scope of a folder from the HOME it is given and from no environment variable.
- REQ-2: The `key_vault::project_scope` function shall answer what `project_scope_in` answers for the environment HOME. · failure: none — the answer of every `key-*` command and of the launcher does not change
- REQ-3: The glm key step shall resolve its scope through `key_vault::project_scope_in` and through no copy of that rule.
- REQ-4: WHEN the glm key step runs outside a git work tree, it shall ask the key question and store the answer in the Shared scope. · failure: none — the answer lands where `moai key-set glm --shared` puts it
- REQ-5: WHEN the glm key step runs outside a git work tree, the question shall say that the key is shared by every project. · failure: none — the words only describe the Shared scope
- REQ-6: IF the workspace identity of the current folder cannot be resolved, THEN the glm key step shall skip the key question with one line naming `moai key-set glm --prompt` and keep the model pick. · failure: none — nothing is written to the vault

## Acceptance

- AC-1 (REQ-1): `project_scope_in(dir, Some(home))` of a git work tree under a temporary HOME answers `Scope::Project` with the folder name `project_scope` makes, of a folder that is no work tree answers `None`, and it reads no `.moai` marker of the process HOME. · verify: `cargo nextest run --offline --locked --lib -E 'test(project_scope_in)'`
- AC-2 (REQ-2): the `key-*` commands and the launcher key read pass unchanged. · verify: `cargo nextest run --offline --locked --test key_vault_cli --test launcher_key_vault --test update_keys_cli`
- AC-3 (REQ-3): `glm_key_scope` is gone from `src/` and `tests/`, and the C-248 key tests pass with their assertions unchanged (only the call that computes the expected scope changes to `project_scope_in`). · verify: `grep -rn "glm_key_scope" src tests` prints nothing and exits 1; `cargo nextest run --offline --locked --test model_effort_questions --test update_model_pick --test update_model_step_reachable_tty`
- AC-4 (REQ-4, REQ-5): on a terminal in a folder under a temporary HOME that is no git work tree, picking glm shows the key question with the shared-by-every-project words, the typed fake key is read from the Shared folder of that HOME, in no other file and on no screen, and the pick is saved. · verify: `cargo nextest run --offline --locked --test update_model_step_reachable_tty -E 'test(a_glm_pick_outside_a_project)'`
- AC-5 (REQ-4): outside a git work tree a Shared glm key that already exists is kept on Enter and replaced on `n`, as a project key is inside a work tree. · verify: `cargo nextest run --offline --locked --test update_model_step_reachable_tty -E 'test(a_glm_pick_outside_a_project)'`
- AC-6 (REQ-6): when the identity does not resolve, the step prints the skip line, writes nothing to the vault and keeps the pick. · verify: `cargo nextest run --offline --locked --lib -E 'test(glm_key_step_skips)'`

## Scope

- write: `.moai/specs/SPEC-20261124-glm-key-scope/**`, `src/key_vault.rs`, `src/model_effort_questions.rs`, `tests/model_effort_questions.rs`, `tests/update_model_pick.rs`, `tests/update_model_step_reachable_tty.rs`. The last four files and `src/model_effort_questions.rs` are in the reservation of C-449; they are added when C-449 gives them back or lands (D-3).
- read-only: `src/identity.rs`, `src/key_prompt.rs`, `src/ops/key_set.rs`, `src/session.rs`
- forbidden: the real key vault and the real HOME (tests use a temporary HOME and a fake key only), a change to a stored entry's format, a key value in any output, `CLAUDE.md`

## Plan

1. RED: a unit test for `project_scope_in` (AC-1), the pty test outside a project that expects the key question and a Shared store (AC-4, AC-5), and a test for the unresolved-identity skip (AC-6).
2. GREEN: `key_vault::project_scope_in(dir, home: Option<&Path>)`; `project_scope(dir)` calls the same body with the identity of the environment HOME; the glm key step calls `project_scope_in(dir, Some(home))`, answers the Shared scope on `None` and skips on `Err`; `glm_key_scope` is deleted and its three test call sites compute the expected scope through `project_scope_in`.
3. Dangerous path run in a throwaway: the pty run outside a project under a temporary HOME shows the key question and the Shared entry, never the real vault.
4. READY: `moai check affected`, guard mutants for the new tests.

## Risks

- A key typed outside a project becomes a Shared key that every project's launcher finds after its own entry. The question says so (REQ-5), and Enter skips.
- The identity of the current folder with the temporary HOME reads that HOME's state; the tests keep to a temporary HOME so the process HOME stays unread (nextest HOME guard).
- Four of the files are held by C-449: the edits there are a few lines each, so the landing merge is small, but the order of the two landings decides who merges.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Outside a git work tree the glm key is stored in the Shared scope. | card C-457 body, the leader's dispatch of 2026-11-24 | delete the Shared entry with `moai key-remove glm --shared` |
| D-2 | An identity that cannot be resolved still skips the question (no Shared fallback). | the launcher's `read_here` refuses that case, so a Shared key would not be found | a later SPEC |
| D-3 | The four held files are added to the reservation after C-449 gives them back, or the edits go through `card line-lease`. | `CLAUDE.md` scope rules; C-449 holds them with no change to them on its branch (read 2026-11-24) | none needed |

## Result

구현 커밋은 `1168e7f047` 이에요(브랜치 `spec/glm-key-scope-kz-20261124`). AC-1~AC-6 은 `d811cd2004` 에서 `verify:` 줄 그대로 실행해 충족을 확인했어요. AC-3 의 테스트 교체 예외는 리더 결정 노트 10379 가 받아들였어요.

### 커밋

- 브랜치는 NeXT `75b6920066` 에서 잘랐어요.
- `dcd19942d1`(SPEC 초안)과 `a9b18ae81f`(`project_scope_in`, 승인 기록, 결정 노트 8521)는 브랜치 `spec/glm-key-scope-20261124` 에 있던 앞선 두 커밋을 옮긴 것이에요.
- `fc226b26c6` 은 중간 단계예요(터미널 테스트와 폴더 seam, 그 시점에는 컴파일하지 않았어요).
- `1168e7f047` 이 구현이에요.

### 막혔던 단계

호스트의 워크트리 격리 가드가 AC-6 단위 테스트 모듈을 셸 heredoc 으로 덧붙이는 명령을 두 번 거절했어요("too complex to verify"). 거절 문구는 blocked 노트 10285 에 있고, Edit/Write 도구로 그 파일을 써도 된다는 사용자 답은 결정 노트 10287 에 기록돼 있어요.

### 동작

- glm 키 단계는 범위를 `key_vault::project_scope_in(dir, Some(home))` 으로 정해요.
- git 작업 트리 안에서는 지금까지의 동작 그대로예요.
- 작업 트리 밖에서는 질문을 하고, 두 질문 모두 모든 프로젝트가 같이 쓰는 키라고 알려요. 조회와 저장에는 Shared 범위를 써요.
- identity 를 풀지 못하거나 현재 디렉터리를 읽지 못하면 `moai key-set glm --prompt` 를 알려 주는 건너뛰기 한 줄만 출력해요. 답을 읽지 않고 아무것도 쓰지 않아요.
- `glm_key_scope` 는 삭제했어요. `1168e7f047` 에서 `grep -rn glm_key_scope src tests` 는 아무것도 출력하지 않고 exit 1 이에요.

### 인수 기준

`d811cd2004` 에서 `verify:` 줄을 적힌 그대로 실행한 결과예요(`target/keziah/verify-457.txt`, 모두 exit 0, AC-3 의 grep 은 출력 없이 exit 1).

- AC-1: met — 2 passed
- AC-2: met — 세 테스트 타깃 30 passed
- AC-3: met — 세 테스트 타깃 56 passed, grep 은 출력 없이 exit 1 이에요. 단언이 바뀐 옛 테스트는 하나뿐이에요: `a_glm_pick_outside_a_project_skips_the_key_and_saves_the_pick` 은 REQ-4 가 바꾸는 동작을 단언해서 `a_glm_pick_outside_a_project_asks_and_stores_a_shared_key`, `a_glm_pick_outside_a_project_keeps_the_shared_key_on_enter`, `a_glm_pick_outside_a_project_replaces_the_shared_key_on_n` 으로 대체했어요. 리더 결정 노트 10379 가 이것을 필요한 예외로 받아들였어요. 다른 옛 테스트에서 바뀐 것은 기대 범위를 계산하는 호출 세 곳뿐이에요.
- AC-4: met — 3 passed
- AC-5: met — AC-4 와 같은 한 번의 실행(3 passed)
- AC-6: met — 2 passed

`1168e7f047` 로 커밋한 변경에서는 lib 와 세 테스트 타깃, lib 테스트 프로필의 `clippy -D warnings` 가 exit 0 이었고, `state_folder_literal_scan` 과 `ac_command_lint` 는 28 passed 였어요.

### RED

프로덕션 코드를 바꾸기 전 결과예요(`target/keziah/red-457.log`).

- `glm_key_step_skips` 테스트 두 개는 새 문구에서 실패했어요. 이전 코드도 이미 건너뛰었기 때문에 그 뒤의 단언까지는 가지 않았어요.
- 새 터미널 테스트 세 개는 테스트 자체의 90초 대기 한도에서 실패했어요.

### 위험 경로

보관소에 키를 쓰거나 바꾸는 경로는 임시 HOME 에서 가짜 키로, 실제 바이너리의 터미널 테스트를 통해서만 실행했어요. 실제 보관소와 실제 HOME 은 열지 않았어요.

### 뮤턴트

`1168e7f047` 에서 하나씩 적용하고 되돌렸어요(끝난 뒤 트리는 HEAD 와 같아요). 모두 exit 100 이에요(`target/keziah/mutant-457-<n>.log`).

| 뮤턴트 | 바꾼 것 | 실패한 테스트 |
| --- | --- | --- |
| M1 | identity 를 풀지 못하거나 폴더를 읽지 못하는 경우를 작업 트리 밖으로 처리 | `glm_key_step_skips` 두 개 |
| M2 | 모든 프로젝트가 같이 쓴다는 문구 삭제 | `a_glm_pick_outside_a_project` 세 개 |
| M3 | 작업 트리 밖에서 `n` 을 입력해도 교체하지 않음 | `a_glm_pick_outside_a_project_replaces_the_shared_key_on_n`(90초 대기 한도) |
| M4 | `project_scope_in` 이 항상 `None` 을 답함 | `project_scope_in_names_the_project_of_a_work_tree_from_the_given_home` |
| M5 | `scope_of_identity` 의 `no_git` 분기 삭제 | `project_scope_in_is_none_outside_a_work_tree` |

### 이 시점에 실행하지 않은 것

- `moai check affected`: READY 때 한 번 실행해서 READY 노트에 기록해요.
- 전체 테스트, Windows 와 Linux 확인, 현재 NeXT 와의 머지.

### 남은 겹침

다른 카드가 `src/key_vault.rs` 도 잡고 있어요. 뒤에 랜딩하는 쪽에서 해결해요.
