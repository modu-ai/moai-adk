---
id: SPEC-20261125-linked-worktree-profile
title: "A linked worktree reads the project profile of its primary checkout, and a missing profile names the next step"
status: draft
revision: 1
created: 2026-11-25
---

# A linked worktree reads the project profile of its primary checkout, and a missing profile names the next step

## Intent

프로젝트 프로필(`project.yaml`)을 git 이 추적하지 않는 프로젝트에서는 연결 워크트리(linked worktree)에 프로필이 없어, 카드 명령이 워크트리 안에서 `project profile missing` 으로 멈추고 레인이 주 체크아웃을 `--project` 로 가리키는 우회를 써요(다른 프로젝트의 현장 점검 기록, 사용자 확인 2026-11-21). 프로젝트 id 는 프로젝트를 가리키는 값이므로, 프로필이 없는 연결 워크트리는 주 체크아웃 프로필의 id 를 읽어 모든 워크트리가 같은 프로젝트(같은 장부, 같은 HOME 폴더)를 가리키게 해요. 프로필이 정말 없을 때의 오류는 다음 단계를 말해요. 앱의 Go 에서 Rust 로의 이전 문제와는 별개의 공통 결함이에요. 로컬 안내문을 워크트리로 복사하는 일은 다른 카드의 몫이에요.

## Requirements

- REQ-1: When a linked worktree has no profile of its own, the project id reader shall return the `project_id` of its primary checkout's profile.
- REQ-2 (REQ-1): IF the worktree has a profile of its own, THEN the project id reader shall return that profile's id and not the primary's.
- REQ-3 (REQ-1): IF the primary checkout's profile cannot give an id, THEN the project id reader shall return an error and not a missing profile.
- REQ-4 (REQ-1): The project id reader shall find the primary checkout only from the `gitdir:` line of a `.git` file that names `<primary>/.git/worktrees/<name>`, absolute or relative to the worktree, and only when that entry is a real folder whose `gitdir` file names this worktree's `.git` back.
- REQ-5 (REQ-1): IF neither the project nor its primary checkout has a profile, THEN the card commands shall refuse with `project profile missing in <folder>` followed by the next step, which is to run `/moai project` in the main checkout or to pass `--project <folder>`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 합성 주 체크아웃과 `.git` 파일이 있는 연결 워크트리에서 프로필 없는 워크트리가 주 프로필의 id 를 읽고, 자기 프로필이 있으면 그것을, 주 프로필이 깨졌으면 오류를, 서브모듈·깨진 `.git` 파일, 되돌림(`gitdir` 파일)이 없거나 다른 워크트리를 가리키는 항목, 링크인 항목에서는 아무것도 빌리지 않고, 상대 경로 `gitdir:` 도 읽어요; 읽기는 워크트리에 아무것도 쓰지 않아요 · verify: `cargo nextest run --lib -E 'test(project_id::)'`
- AC-2 (REQ-1, REQ-5): `project_id_of` 가 연결 워크트리와 주 체크아웃에 같은 id 를 내고, 프로필이 없으면 폴더와 다음 단계(`/moai project`, `--project <folder>`)를 말하며 계약 분류는 NotFound 로 그대로예요 · verify: `cargo nextest run --lib -E 'test(ledger::command::)'`

## Scope

- write: `src/project_id.rs`(읽기 대체와 단위 시험), `src/ledger/command.rs`(오류 문장과 시험), 이 SPEC 폴더.
- write 하지 않는 것: `src/profile.rs`(`PROFILE_MISSING` 은 여러 곳이 문자열 동치로 읽어요), 프로필 쓰기 경로(`project-apply` 는 자기 폴더에만 써요), 다른 프로젝트·실제 레인 설정, 로컬 안내문 복사.

## Plan

1. RED: `project_id::tests` 의 연결 워크트리 시험 여섯 개를 먼저 써서 새 동작 둘이 실패하는 것을 봐요.
2. GREEN: `project_id::read` 가 자기 프로필이 없을 때만 `.git` 파일로 주 체크아웃을 찾아 그 프로필을 읽어요. `project_id_of` 의 오류 문장에 폴더와 다음 단계를 붙여요.
3. 이 입구 하나가 카드 명령과 프로젝트 HOME 키(`project_home::resolve`)의 공통 읽기라서 호출한 곳은 따로 바꾸지 않아요.
