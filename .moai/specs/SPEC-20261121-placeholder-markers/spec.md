---
id: SPEC-20261121-placeholder-markers
title: "Values the launcher writes carry a placeholder marker and the project interview always asks them"
status: draft
revision: 1
created: 2026-11-21
---

# Values the launcher writes carry a placeholder marker and the project interview always asks them

## Intent

설치 흐름 점검에서 나온 틈이에요. 처음 실행하면 런처가 프로필에 `purpose`(자리표시 문장), `hosts`(실행한 호스트 하나), `development_methods`(`unspecified`)를 채워요(`src/launcher_init.rs`의 `default_profile`). 그런데 `/moai project`의 질문 은행은 이 항목들을 "프로필에 없을 때" 물어요(`interview.md`의 `purpose`·`method`·`hosts` 행). 값이 이미 들어 있으니 묻지 않게 되고, 그래서 Codex를 쓰려는 사람이 먼저 `moai cc`를 실행하면 `hosts=[claude]`로 굳어 Codex 단계가 모두 건너뛰어져요(`src/project_hosts.rs`).

사용자 결정(2026-11-20): 런처가 쓴 값에는 자리표시 표식이 붙고, `/moai project`는 표식이 붙은 항목을 늘 묻는다. 호스트는 설치된 도구를 모두 추천한다.

표식은 새로 만들지 않고 프로필의 `decisions`에 이미 있는 `source: "default"` 결정을 써요. 런처는 도구·개발 방법·호스트에는 이미 이 결정을 쓰고, 목적에는 쓰지 않아요. 이 SPEC의 첫 조각은 그 빈 곳(목적)을 채우고 표식 규칙을 시험으로 고정해요. 질문 쪽(스킬 글)과 설치된 호스트 목록은 다른 카드가 쥔 파일이 풀려야 해서 다음 조각이에요.

## Non-goals

- 사용자나 저장소가 이미 답한 항목(`source: user` 또는 `repo`)은 다시 묻지 않아요.
- `Profile::is_placeholder`(프로필 전체가 자리표시일 때 AGENTS.md 생성을 막는 규칙)는 바꾸지 않아요.
- 프로필 스키마와 `decisions`의 항목 수 한도는 바꾸지 않아요.
- 질문 은행의 다른 행(검사 명령, 버전 관리 등)은 바꾸지 않아요.

## Requirements

- REQ-1: When the launcher writes a first-launch profile, the launcher shall record one `source: default` decision for each of purpose, tools, development methods and hosts · failure: none — the profile stays valid and the values stay as they are today.
- REQ-2: The marker of a first-launch value shall be the decision of that topic with `source: "default"` and the id `auto-init-purpose`, `auto-init-tools`, `auto-init-methods` or `auto-init-hosts`, never a new profile field.
- REQ-3: When `/moai project` finds a topic with a `source: default` decision, `/moai project` shall ask that item even though the profile holds a value · failure: none — an unanswered item is recorded as HOLD and asked again on the next run.
- REQ-4: When `/moai project` asks the hosts item, `/moai project` shall recommend every installed host · failure: none — the user may pick fewer.
- REQ-5: When the user answers a marked item, the plan shall replace that topic's `source: default` decision with the user's decision · failure: none — a topic keeps its marker until it is answered.

## Acceptance

- AC-1 (REQ-1, REQ-2): the first-launch profile holds exactly one `source: default` decision for each of `purpose`, `tools`, `development_methods` and `hosts`, with the ids of REQ-2; the `hosts` decision names the launching host. verify: `cargo test --locked --offline --test placeholder_markers`
- AC-2 (REQ-1): the profile is still valid for every launching host, every folder name the launcher accepts, and `Profile::is_placeholder` still answers true for it. verify: `cargo test --locked --offline --test placeholder_markers`; `cargo test --locked --offline --test placeholder_profile_guard`
- AC-3 (REQ-3, REQ-4, REQ-5): `/moai project` reads a profile written by a first launch and asks purpose, method and hosts although each holds a value, recommends every installed host, and drops the markers it has answered. verify: a temp-HOME run of `moai project-detect` and `moai project-plan` on a first-launch profile (second piece, needs the question and workflow files)
- AC-4: light checks. verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write now (this piece): `src/launcher_init.rs`, `tests/placeholder_markers.rs`, `tests/placeholder_profile_guard.rs`, `.moai/specs/SPEC-20261121-placeholder-markers/**`
- write later (second piece, held by other cards today): `skills/moai/workflows/project.md`, `skills/moai/references/internal/interview.md`, `skills/moai/references/source-manifest.json`, `src/project_detect.rs`, and `src/profile.rs` if the plan has to drop a marker there
- forbidden: keys and the real HOME, card numbers in a SPEC id, branch or commit, writes to another card's branch

## Plan

1. Test first: `tests/placeholder_markers.rs` reads the launcher's first-launch profile through a temporary-HOME launch and asserts the four markers and their ids (RED: no `purpose` marker yet).
2. `src/launcher_init.rs`: `default_profile` adds the `purpose` decision; the four ids come from one list so the test and the code cannot drift.
3. Run the checks and write each command and exit code into Result.
4. Second piece after the other cards return their files: the `Ask when` cells of the `purpose`, `method` and `hosts` rows, the interview step of `project.md`, the installed hosts in `project-detect`, and the marker drop in the plan.

## Risks

- A marker without the question change does nothing visible; the first piece alone changes only what the profile file holds. The leader should land the second piece with it.
- The `decisions` list has a size limit (`DECISION_ITEM_LIMIT`); four markers fit well inside it.
- Rollback: one revert commit (one decision entry and one test file).

## Decisions

- 2026-11-21: the marker is the `source: default` decision that the launcher already writes for tools, development methods and hosts, not a new field, so no schema or reader changes. Reversible, category `none`, decided by agent-17.
- 2026-11-21: the work splits in two because the question files are held by other cards; the first piece is self-contained and safe to land alone. Reversible, category `none`, decided by agent-17.

## Result

(Filled after the implementation.)
