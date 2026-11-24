---
id: SPEC-20261124-glm-key-scope
title: "The glm key pick resolves its scope through key_vault and answers a Shared key outside a project"
status: draft
revision: 1
created: 2026-11-24
parent: SPEC-20261121-glm-key-at-pick
---

# The glm key pick resolves its scope through key_vault and answers a Shared key outside a project

## Intent

카드 C-457 은 C-248(SPEC-20261121-glm-key-at-pick)의 READY 노트 8362 와 감사 N6(노트 8382, 8487)이 남긴 두 후속이에요.

1. `src/model_effort_questions.rs:620` 의 `glm_key_scope` 는 `src/key_vault.rs:203` 의 `project_scope` 와 같은 계산(identity 의 `clone_id`·`clone_anchor`·`project_id` 를 읽어 `folder_name` 으로 만드는 14줄)을 복제해요. 이유는 `project_scope` 가 `identity(dir)` 를 불러서 프로세스 HOME 을 읽기 때문이에요(SPEC-20261124-identity-home-arg 가 `identity_in(path, home)` 를 더했어요). 두 곳이 어긋나면 모델 단계가 런처가 읽지 않는 폴더에 키를 저장해요. `key_vault` 에 HOME 을 받는 `project_scope_in` 을 두고 복제를 없애요.
2. 프로젝트 밖(git 작업 트리가 아닌 곳)에서는 모델 단계가 키 질문을 건너뛰고 `moai key-set glm --prompt` 안내 한 줄만 보여요. 설치 흐름은 보통 프로젝트 밖에서 돌아서 설치 보고서 G8(glm 을 고르면 그 자리에서 키를 받아요)의 일부가 비어요. 프로젝트 밖에서는 Shared 범위(런처가 프로젝트 항목 다음에 읽는 폴더, `moai key-set --shared` 가 쓰는 곳)에 저장해요.

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
