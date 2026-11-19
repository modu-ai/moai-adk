---
id: SPEC-20261119-nested-manifests
title: "project-detect: manifests below the root and their tools (C-61)"
status: draft
revision: 1
created: 2026-11-19
---

# project-detect: manifests below the root and their tools (C-61)

## Intent

카드 C-61 (p2): `moai project-detect` 가 루트 아래 중첩 manifest(`plugins/<name>/package.json` 등)를 정해진 깊이까지 찾아 경로로 나열하고(AC-1), 그 manifest 이름이 부르는 도구를 도구 질문의 추천 옵션으로 내놓아요(AC-2). 관찰된 결함(2026-11-17, goose-money-plugins의 `/moai project`): `detected.manifests = []` 인데 `plugins/moai-atp/package.json`·`plugins/moai-atp/server/package.json` 이 git 추적 중이었고, 도구(node·npm·typescript·python)를 인터뷰가 전부 다시 물어야 했어요.

## Non-goals

- 카드 AC-3(끝난 인터뷰가 `development_methods` 를 `unspecified` 로 남기는 것, 질문 은행 문구) — `skills/**` 와 manifest 가 필요해 C-47 뒤 카드(note 2315).
- 중첩 manifest 의 **내용** 읽기(의존성·데이터베이스·스크립트) — `databases` 는 #2535 계약대로 루트 한정.
- `skills/**` 문구 갱신(`skills/moai/workflows/project.md` 의 "root only" 줄 포함) — 이 카드의 쓰기 범위 밖; C-47 이 갱신.
- 프로세스 시작(git 호출 포함), HOME·전역 제외 목록 읽기, 심볼릭 링크 추적, 언어 서버 설치 여부 판단.

## Requirements

- REQ-1: When `project-detect` reads a project folder, it shall find every manifest file named in `MANIFESTS` that sits at folder depth at most `MAX_DEPTH` below the root, and shall list each one in `detected.manifests` by its path relative to the project root (a manifest at the root by its file name alone).
- REQ-2 (REQ-1): IF a manifest sits in a folder deeper than `MAX_DEPTH`, or in a folder named in `SKIP_DIRS`, or in a folder that the project's own ignore lines exclude (the `.gitignore` of the root and of each folder the walk enters, and `.git/info/exclude` when `.git` is a real folder — each read as a regular non-linked file of at most `MAX_FILE` bytes, evaluated as git does: the deepest list with a matching line decides, and a folder a non-negated line excludes is not entered), or in a folder reached only through a symbolic link, THEN the command shall not list it.
- REQ-3: The manifest walk shall read file and folder names only; it shall not read the bytes of a file below the root, and an ignore list it cannot read (a link, over the size limit, not UTF-8) shall count as no lines.
- REQ-4: When a folder holds nothing but nested manifests (no root manifest, no code file), `detected.project_kind` shall be `existing`.
- REQ-5: When `project-detect` reports the facts, it shall report beside `detected` (never inside it) a `tools` object with one entry per tool that the name of a detected manifest implies (`package.json` → `node`,`npm`; `tsconfig.json` → `typescript`; `Cargo.toml` → `cargo`; `pyproject.toml`,`setup.py`,`requirements.txt` → `python`; `go.mod` → `go`; `pom.xml` → `maven`; `build.gradle`,`build.gradle.kts` → `gradle`; `Gemfile` → `ruby`; `composer.json` → `php`), keyed by tool name in sorted order with each value the sorted manifest paths that imply it, at most `MAX_ITEMS` entries; no bytes of a nested manifest shall appear anywhere in the output.
- REQ-6: `validate` shall accept a `manifests` entry that is a relative path whose `/`-separated components are each recordable (the whole path at most 160 bytes, no `..`, no absolute path, no backslash, no empty component, no trailing separator) and shall refuse every other value there, with every other list checked exactly as before.

## Acceptance

- AC-1 (REQ-1…REQ-4): 중첩 manifest 가 경로로 나열돼요(루트 manifest 이름과 섞여 정렬, `BTreeSet` 순). 깊이 경계 시험: 깊이 3 폴더의 `package.json` 은 찾고 깊이 4 는 안 찾아요. 무시 목록(루트·중첩 `.gitignore`, 부정 줄 포함)과 `SKIP_DIRS` 가 고른 폴더, 심볼릭 링크 폴더는 읽지 않고, 중첩 manifest 만 있는 폴더는 `existing` 이에요.
  verify: `cargo test --locked --offline --test project_detect_cli nested_manifests_are_listed_with_their_paths`
  verify: `cargo test --locked --offline --test project_detect_cli the_manifest_depth_limit_is_three_folder_levels`
  verify: `cargo test --locked --offline --test project_detect_cli folders_the_ignore_lists_or_the_walk_skip_are_not_read`
  verify: `cargo test --locked --offline --test project_detect_cli a_linked_nested_folder_is_not_followed`
  verify: `cargo test --locked --offline --test project_detect_cli nested_manifests_are_capped_like_every_list`
- AC-2 (REQ-5): `tools` 객체가 `detected` 옆에, 중첩 manifest 경로와 함께(정렬·`MAX_ITEMS` 상한) 나오고 manifest 내용(토큰)은 출력 어디에도 안 나타나며 `detected` 안에는 들어가지 않아요. `databases` 는 루트 한정을 유지(#2535 시험 갱신).
  verify: `cargo test --locked --offline --test project_detect_cli nested_manifests_feed_the_tools_fact`
  verify: `cargo test --locked --offline --test project_detect_cli nested_manifests_are_listed_but_databases_stay_at_the_root`
- AC-3 (REQ-6): `validate` 는 경로 모양의 `manifests` 값만 받고(구성 요소별 검사) 나머지 값은 전과 같이 거부해요.
  verify: `cargo test --locked --offline --lib project_detect`
- AC-4: 이웃 검사(대상 한정; 전량은 PR 의 CI run 이 인용).
  verify: `cargo test --locked --offline --test project_detect_cli`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`(로컬은 대상 한정, 전량은 PR CI 인용)

## Scope

- write: `src/project_detect.rs`, `src/ops/project_detect.rs`, `tests/project_detect_cli.rs`, `.moai/specs/SPEC-20261119-nested-manifests/spec.md`
- read-only: `src/worktree_include.rs`(무시-목록 문법 참고), `skills/moai/workflows/project.md`(문구 갱신은 C-47)
- forbidden: 실제 HOME·전역 제외 목록, 네트워크, 프로세스 시작, `skills/**` 쓰기, 원장 수편

## Plan

1. RED: `tests/project_detect_cli.rs` 에 시험 먼저(임시 저장소만) — 위 AC 의 verify 시험들. 현 코드에서 실패해야 하고, 그 실패를 기록해요.
2. GREEN: `src/project_detect.rs` — `nested_manifests` 걷기(이름만·깊이·예산 `MAX_ENTRIES`), 국소 무시-목록 판단(루트+폴더별 `.gitignore`, `.git/info/exclude`), `recordable_relpath` 검증, `tools` 팩트(`detected` 옆). `src/ops/project_detect.rs` 문서 줄 한 줄.
3. 이웃 검사·PR(audit-tier 1)·감사.

## Risks

- 무시-목록 파일을 못 읽으면 규칙 없음으로 취급해 "더 읽는" 방향으로만 실패해요(과잉 목록 가능, 누출 없음 — 이름만 읽음).
- `SKIP_DIRS` 이름과 같은 진짜 소스 폴더(예: `build/`)는 #2496 계약 그대로 건너뛰어요.
- `manifests` 값이 루트 이름 → 경로 혼합으로 바뀌어, 저장된 profile 의 `detected` 는 다음 `project-plan` 적용 때 새 모양으로 바뀌어요(모양은 같고 값이 추가되는 것).
- `skills/moai/workflows/project.md` 의 "root only" 설명이 이 변화 뒤 낡아요(C-47 이 갱신; PR 에 명시).
- `worktree_include` 의 매처 재사용이 범위 밖 파일이라 국소 구현을 먼저 해요(D-1); 중복이 생기면 리더가 범위를 넓혀 통합할 수 있어요. 리더 결정(note 2352 답): 국소 구현을 유지하고 공유 매처 통합은 병합 뒤 후속 조각이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 무시-목록 판단은 `src/project_detect.rs` 안에 국소 구현; `worktree_include` 매처의 `pub(crate)` 노출은 리더 질문 후 | 카드 쓰기 범위 밖 파일은 쓰지 않는다(AGENTS.md); 진행 우선 | 리더가 범위를 넓히면 공유 모듈로 통합 |
| D-2 | `tools` 팩트는 `detected` 옆 객체(키=도구, 값=경로 목록) | `commands` 팩트(#2639)와 같은 자리·같은 원칙; profile 의 `detected` 모양 불변 | 배열 형태 |
| D-3 | 중첩 manifest 는 `languages`·`lsp`·`databases` 에 반영하지 않음 | AC-1·AC-2 범위; `databases` 루트 한정은 #2535 계약 유지 | 반영 추가 카드 |
| D-4 | `moai spec new` 미사용(설치 바이너리에 없음) — SPEC 파일을 도구로 직접 작성 | 훅이 요구한 것은 가지의 SPEC 존재 | 소행으로 되돌리기 |

## Result

- AC-1: — (진행 중)
- AC-2: — (진행 중)
- AC-3: — (진행 중)
- AC-4: — (진행 중)
