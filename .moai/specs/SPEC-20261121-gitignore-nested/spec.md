---
id: SPEC-20261121-gitignore-nested
title: "The update plan reads nested .gitignore files before it says a file becomes visible"
status: draft
revision: 2
created: 2026-11-21
---

# The update plan reads nested .gitignore files before it says a file becomes visible

## Intent

`.moai/` 레이아웃 조각 1(SPEC-20261120-moai-layout)의 감사가 남긴 N5 예요. 계획의 `now visible: <경로>` 를 만드는 `newly_visible`(`src/agents.rs`)은 루트 `.gitignore` 와 제외 파일만 읽는 임시 저장소에 묻기 때문에, 아래 폴더의 `.gitignore`(예: `.moai/tools/.gitignore`)가 여전히 무시하는 파일도 "새로 보여요"로 말해요. 넘치게 말하는 쪽이라 안전하지만 틀린 말이에요. 위쪽 폴더의 `.gitignore` 도 같이 읽어요.

(저장소 `.gitignore` 시험을 추적되는 프로필에 맞추는 일은 SPEC-20261121-gitignore-tracked-profile 이 맡아요. 이 SPEC 은 그 시험을 바꾸지 않아요.)

## Non-goals

- 저장소 `.gitignore` 와 그 시험을 바꾸지 않아요.
- `now visible` 줄의 모양·상한(스무 개)과 블록의 줄을 바꾸지 않아요.
- 하위 폴더의 `.gitignore` 를 고치거나 지우지 않아요(읽기만 해요).
- 쓰는 쪽을 `local/` 로 옮기는 일(조각 2)은 하지 않아요.

## Requirements

- REQ-1: When the project update plan lists the files it makes visible, it shall leave out a file that a `.gitignore` in a folder above that file still ignores.
- REQ-2 (REQ-1): IF a `.gitignore` above a candidate file cannot be read as text, THEN the plan shall still name the files as visible and add one line naming that `.gitignore`.

## Acceptance

- AC-1 (REQ-1): `tools/` 의 `.gitignore` 가 `*.tmp` 를 무시하면 계획이 `tools/a.tmp` 와 더 깊은 `tools/sub/deep/z.tmp` 를 말하지 않고, 아래 폴더의 `.gitignore`(`*.o`)는 그 폴더 밖 파일에 적용되지 않으며, `tools/b.sh` 와 `.gitignore` 파일들은 말해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_file_a_nested_gitignore_still_ignores_is_not_named_as_visible`
- AC-2 (REQ-1): 추적되는 하위 `.gitignore`(후보가 아니에요)가 몇 폴더 아래의 유일한 후보에도 적용돼요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_above_a_deeper_file_still_applies`
- AC-3 (REQ-2): 읽을 수 없는 하위 `.gitignore`(UTF-8 이 아닌 글, 1 MiB 를 넘는 글)가 있으면 파일들을 그대로 말하고 그 `.gitignore` 를 이름으로 대는 한 줄이 더해져요. 딱 1 MiB 는 읽혀요.
  verify: `cargo test --locked --offline --test gitignore_block_cli an_unreadable_nested_gitignore_is_named_and_the_files_stay_listed`
- AC-4 (REQ-2): 링크인 하위 `.gitignore` 는 따라가지 않고(가리키는 곳이 프로젝트 밖일 수 있어요) 읽을 수 없는 것으로 이름을 대요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_nested_gitignore_that_is_a_link_is_named_and_not_followed`

## Scope

- write: `.moai/specs/SPEC-20261121-gitignore-nested/spec.md`, `src/agents.rs`(`newly_visible`, 새 `nested_gitignores`, `user_rule_hits` 가 부르는 `rule_hits_with_nested`), `tests/gitignore_block_cli.rs`
- read-only: `.gitignore`, `src/state_root.rs`
- forbidden: `.gitignore` 바꾸기, 블록의 줄·`now visible` 의 모양과 상한 바꾸기

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: AC-1·AC-3 시험을 먼저 써서 실패하는 것을 봐요.
3. GREEN: `newly_visible` 이 후보 파일마다 위쪽 폴더의 `.gitignore` 를 모아(폴더마다 한 번) 임시 저장소의 같은 자리에 써서 `check-ignore` 에 함께 읽히게 해요. 읽을 수 없으면 줄을 더해요.
4. 확인: `gitignore_block_cli` 전체, 변이, `moai check affected`.

## Risks

- 후보 파일마다 위쪽 폴더를 오르며 `.gitignore` 를 읽어서 파일이 아주 많으면 읽기가 늘어요: 폴더마다 한 번만 읽어요(집합으로 중복을 없애요).
- 하위 `.gitignore` 의 `!` 줄이 블록의 줄과 겹치면 임시 저장소의 git 이 자기 순서로 판단해요: 실제 git 과 같은 규칙(깊은 폴더가 앞서요)이에요.
- 임시 저장소에 쓰는 `.gitignore` 는 사용자 프로젝트의 것을 복사한 글이라 외부 입력이에요: 1 MiB 를 넘으면 읽지 않고 한 줄로 알려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 하위 `.gitignore` 는 임시 저장소의 같은 상대 경로에 복사해 git 에 묻는다 | 규칙을 직접 풀이하지 않고 git 의 판단을 쓰는 지금 방식 그대로예요 | 복사 빼기 |
| D-2 | 읽을 수 없거나 1 MiB 를 넘거나 링크인 하위 `.gitignore` 는 건너뛰고 한 줄로 알린다 | 넘치게 말하는 쪽이 안전해요. 조용히 건너뛰면 왜 말했는지 알 수 없어요 | 줄 빼기 |

## Result

- AC-1: met — `cargo test --locked --offline --test gitignore_block_cli a_file_a_nested_gitignore_still_ignores_is_not_named_as_visible` 통과(`tools/a.tmp`·`sub/deep/z.tmp` 는 말하지 않고, `sub/x.o` 도 말하지 않으며, `tools/top.o`(아래 폴더 규칙 밖)와 `.gitignore` 두 개는 말해요. 미리보기가 쓴 것이 없어요). 변이 "아래 폴더 규칙을 넘기지 않기"에 실패했어요.
- AC-2: met — `… a_nested_gitignore_above_a_deeper_file_still_applies` 통과(추적되는 `tools/.gitignore` 가 `sub/deep/z.tmp` 하나에 적용). 변이 "읽는 폴더를 가장 깊은 것으로 줄이기"에 실패했어요(처음엔 `.gitignore` 자신이 후보라 이 변이가 살아남아 시험을 추적되는 파일로 바꿨어요).
- AC-3: met — `… an_unreadable_nested_gitignore_is_named_and_the_files_stay_listed` 통과(UTF-8 아닌 글: 줄 하나와 파일들, 딱 1 MiB: 읽힘, 1 MiB + 1: 줄 하나). 변이 "한도 비교를 `<=` 에서 `<` 로"와 "알림 줄 빼기"에 실패했어요.
- AC-4: met — `… a_nested_gitignore_that_is_a_link_is_named_and_not_followed` 통과. 변이 "링크도 읽기"에 실패했어요.
- 확인: `cargo test --locked --offline --test gitignore_block_cli` 42 개 통과, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `cargo fmt --check` 0. 변이 5 개 모두 시험이 잡았어요(첫 시도에서 5번은 변이 코드가 컴파일되지 않아 무효였고 고쳐 다시 돌렸어요).
- 돌리지 않은 것: 전체 `moai check affected`(조각 1 이 병합된 뒤 그 위 머지에서 돌려 병합 준비 노트에 줄째로 적어요), `agents` 의 `--lib` 시험과 골든(`product_cli`)은 그 검사에 들어 있어요.
