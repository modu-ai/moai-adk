---
id: SPEC-20261121-gitignore-nested
title: "The repository .gitignore test follows the tracked profile, and the update plan reads nested .gitignore files"
status: draft
revision: 1
created: 2026-11-21
---

# The repository .gitignore test follows the tracked profile, and the update plan reads nested .gitignore files

## Intent

`.moai/` 레이아웃 조각 1(SPEC-20261120-moai-layout)의 감사가 남긴 둘이에요.

1. 저장소 `.gitignore` 시험: 사용자가 `/moai project` 로 프로필을 추적하기로 하고 저장소 `.gitignore` 에 `!.moai/project.yaml` 을 더했어요(커밋 bcfb1bc5b7). 그런데 `the_repository_gitignore_tracks_specs_and_ignores_the_rest` 는 아직 `.moai/project.yaml` 이 무시된다고 기대해서 tip 에서 혼자 돌려도 실패해요. `.gitignore` 가 아니라 시험을 고쳐요: 저장소는 SPEC 과 함께 `.moai/project.yaml` 도 추적해요. 나머지 기대는 그대로예요.
2. 계획의 `now visible: <경로>`(감사 N5): `newly_visible`(`src/agents.rs`)은 루트 `.gitignore` 와 제외 파일만 읽는 임시 저장소에 묻기 때문에, 아래 폴더의 `.gitignore`(예: `.moai/tools/.gitignore`)가 여전히 무시하는 파일도 "새로 보여요"로 말해요. 넘치게 말하는 쪽이라 안전하지만 틀린 말이에요. 아래 폴더의 `.gitignore` 도 같이 읽어요.

## Non-goals

- 저장소 `.gitignore` 를 바꾸지 않아요(사용자의 줄을 그대로 둬요).
- `now visible` 줄의 모양·상한(스무 개)과 블록의 줄을 바꾸지 않아요.
- 하위 폴더의 `.gitignore` 를 고치거나 지우지 않아요(읽기만 해요).
- 쓰는 쪽을 `local/` 로 옮기는 일(조각 2)은 하지 않아요.

## Requirements

- REQ-1: The repository's own `.gitignore` test shall expect `.moai/project.yaml` and the SPEC files to be tracked and every other state file and every credential or secret-shaped file to be ignored.
- REQ-2: When the project update plan lists the files it makes visible, it shall leave out a file that a `.gitignore` in a folder above that file still ignores.
- REQ-3 (REQ-2): IF a `.gitignore` above a candidate file cannot be read as text, THEN the plan shall still name the files as visible and add one line naming that `.gitignore`.

## Acceptance

- AC-1 (REQ-1): 저장소 `.gitignore` 시험이 tip 에서 통과해요: `.moai/project.yaml` 은 무시되지 않고 SPEC 파일도 무시되지 않으며, 나머지 상태 파일·자격 증명 모양 파일은 무시돼요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_repository_gitignore_tracks_specs_and_ignores_the_rest`
- AC-2 (REQ-2): `tools/` 의 `.gitignore` 가 `*.tmp` 를 무시하면 계획이 `tools/a.tmp` 를 말하지 않고 `tools/b.sh` 와 그 `.gitignore` 파일은 말해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_file_a_nested_gitignore_still_ignores_is_not_named_as_visible`
- AC-3 (REQ-3): 읽을 수 없는 하위 `.gitignore`(UTF-8 이 아닌 글)가 있으면 파일들을 그대로 말하고 그 `.gitignore` 를 이름으로 대는 한 줄이 더해져요.
  verify: `cargo test --locked --offline --test gitignore_block_cli an_unreadable_nested_gitignore_is_named_and_the_files_stay_listed`

## Scope

- write: `.moai/specs/SPEC-20261121-gitignore-nested/spec.md`, `src/agents.rs`(`newly_visible` 와 `user_rule_hits` 의 하위 `.gitignore` 읽기), `tests/gitignore_block_cli.rs`
- read-only: `.gitignore`, `src/state_root.rs`
- forbidden: `.gitignore` 바꾸기, 블록의 줄·`now visible` 의 모양과 상한 바꾸기

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: AC-1 은 tip 에서 이미 실패해요(감사 기록). AC-2·AC-3 시험을 먼저 써서 실패하는 것을 봐요.
3. GREEN: 시험의 기대를 고치고(AC-1), `newly_visible` 이 후보 파일마다 위쪽 폴더의 `.gitignore` 를 모아 임시 저장소의 같은 자리에 써서 `check-ignore` 에 함께 읽히게 해요(AC-2). 읽을 수 없으면 줄을 더해요(AC-3).
4. 확인: `gitignore_block_cli` 전체, `--lib`, 골든 `product_cli`, fmt, clippy, 변이, `moai check affected`.

## Risks

- 후보 파일마다 위쪽 폴더를 오르며 `.gitignore` 를 읽어서 파일이 아주 많으면 읽기가 늘어요: 폴더마다 한 번만 읽고(집합으로 중복을 없애요) 읽기는 후보 파일의 폴더 사슬에만 해요.
- 하위 `.gitignore` 의 `!` 줄이 블록의 줄과 겹치면 임시 저장소의 git 이 자기 순서로 판단해요: 실제 git 과 같은 규칙(깊은 폴더가 앞서요)을 써요.
- 임시 저장소에 쓰는 `.gitignore` 는 사용자 프로젝트의 것을 복사한 글이라 외부 입력이에요: 크기 한도 없이 읽지 않고 1 MiB 를 넘으면 읽을 수 없는 것으로 쳐서 줄을 더해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시험을 고치고 저장소 `.gitignore` 는 그대로 둔다 | 리더 지시·사용자 결정: 프로필 추적은 사용자가 한 일이에요 | 해당 없음 |
| D-2 | 하위 `.gitignore` 는 임시 저장소의 같은 상대 경로에 복사해 git 에 묻는다 | 규칙을 직접 풀이하지 않고 git 의 판단을 쓰는 지금 방식 그대로예요 | 복사 빼기 |
| D-3 | 읽을 수 없거나 1 MiB 를 넘는 하위 `.gitignore` 는 건너뛰고 한 줄로 알린다 | 넘치게 말하는 쪽이 안전해요. 조용히 건너뛰면 왜 말했는지 알 수 없어요 | 줄 빼기 |

## Result

(구현 뒤에 채워요.)
