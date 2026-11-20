---
id: SPEC-20261121-gitignore-tracked-profile
title: "The repository .gitignore test follows the tracked project profile"
status: draft
revision: 1
created: 2026-11-21
---

# The repository .gitignore test follows the tracked project profile

## Intent

사용자가 `/moai project` 로 프로필을 추적하기로 하고 저장소 `.gitignore` 에 `!.moai/project.yaml` 을 더했어요(커밋 bcfb1bc5b7, 사용자 결정 "내가 한 것, 그대로 둠"). 그런데 `tests/gitignore_block_cli.rs` 의 `the_repository_gitignore_tracks_specs_and_ignores_the_rest` 는 아직 `.moai/project.yaml` 이 무시된다고 기대해서 tip 에서 혼자 돌려도 실패하고, 그래서 모든 병합 후보가 막혀요. `.gitignore` 가 아니라 시험을 고쳐요: 저장소는 SPEC 과 함께 `.moai/project.yaml` 도 추적해요. 나머지 기대는 그대로예요.

## Non-goals

- 저장소 `.gitignore` 를 바꾸지 않아요(사용자의 줄을 그대로 둬요).
- 다른 시험과 `src/` 를 바꾸지 않아요.

## Requirements

- REQ-1: The repository's own `.gitignore` test shall expect `.moai/project.yaml` and the SPEC files to be tracked and every other state file and every credential or secret-shaped file to be ignored.

## Acceptance

- AC-1 (REQ-1): 저장소 `.gitignore` 시험이 tip 에서 통과해요: `.moai/project.yaml` 과 SPEC 파일은 무시되지 않고, 나머지 상태 파일·자격 증명 모양 파일은 무시돼요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_repository_gitignore_tracks_specs_and_ignores_the_rest`

## Scope

- write: `.moai/specs/SPEC-20261121-gitignore-tracked-profile/spec.md`, `tests/gitignore_block_cli.rs`
- read-only: `.gitignore`
- forbidden: `.gitignore` 바꾸기, `src/` 바꾸기

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: 시험이 tip 에서 `.moai/project.yaml is not ignored` 로 실패하는 것을 봐요.
3. GREEN: `.moai/project.yaml` 을 "무시되지 않아야 하는" 목록으로 옮겨요.
4. 확인: `gitignore_block_cli` 전체, 변이(목록 옮기기를 되돌리면 실패), `moai check affected`.

## Risks

- 시험이 `.gitignore` 를 컴파일 때 읽어 들이므로(`include_str!`), 사용자가 그 줄을 다시 빼면 이 시험이 다시 실패해요: 그건 결정이 바뀐 것이라 시험도 그때 같이 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시험을 고치고 저장소 `.gitignore` 는 그대로 둔다 | 리더 지시·사용자 결정: 프로필 추적은 사용자가 한 일이에요 | 해당 없음 |

## Result

(구현 뒤에 채워요.)
