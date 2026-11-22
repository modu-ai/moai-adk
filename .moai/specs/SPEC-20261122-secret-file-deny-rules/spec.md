---
id: SPEC-20261122-secret-file-deny-rules
title: "Key files found in a project are offered as Read deny rules in the root project settings, with the .claudeignore entries carried over, only on the user's pick"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# Key files found in a project are offered as Read deny rules in the root project settings, with the .claudeignore entries carried over, only on the user's pick

## Intent

비밀 파일 거절 규칙. 공식 문서의 사실(2026-11-22 받은 원문에서 확인): (1) "If your project has a `.claudeignore` file, it has no effect, so move its entries into `Read` deny rules"(permissions.md) — `.claudeignore` 는 아무 효과가 없어요. (2) "`@` 언급은 훅을 거치지 않음": "Your `PreToolUse` hooks don't run when Claude Code includes a file you reference with `@`, because it makes no tool call"(common-workflows.md) — 훅으로 읽기를 막는 길은 `@` 언급을 놓쳐요; 거절 규칙은 `@` 언급·붙여넣은 그림 경로·디렉터리 목록에도 최선을 다해 적용돼요(permissions.md). (3) 거절 규칙은 모든 모드에서 이겨요("deny, then ask, then allow"), 작업공간 신뢰 단계를 기다리지 않아요("deny and ask rules aren't affected, since they only restrict" — permissions.md). (4) 작업트리 세션은 주 체크아웃 루트의 설정 파일을 읽으므로 "the deny rules need a second copy in the repository root's `.claude/settings.json` so worktree sessions pick them up"(large-codebases.md). (5) `Read(path)` 거절은 같은 경로의 Edit·Write(새 파일 만들기 포함)까지 막아요; `Write(path)`·`Glob(path)`·`Grep(path)` 꼴은 아무것도 못 맞춰요(permissions.md). (6) 어떤 파일이 자격 증명인지: secure-deployment 의 표(`.env`·`*.pem`·`*.key`·`*-service-account.json`·`.npmrc`·`.pypirc` …). 무인 레인이 이 저장소의 비밀 파일을 읽는 사고가 실제로 일어났고(안내 문구 규칙의 유래), 이 조각은 그 문을 닫아요. 규칙은 프로젝트 설정이므로 미리보기 뒤 사용자의 선택으로만 쓰여요(가드 파일 규칙). 첫 조각은 SPEC 만 쓰고, 테스트는 Rust 만이에요.

## Non-goals

- 운영체체 수준의 차단(샌드박스) — Read 거절은 "arbitrary subprocesses that read or write files indirectly" 에 닿지 않아요(permissions.md 경고); 샌드박스는 별개의 기능이에요.
- NotebookEdit 경로 — `Read` 거절이 덮지 않는 한 곳(permissions.md); 위험에 기록만 해요.
- 사용자 범위·전역 경로(`~/`·`//`) 거절 규칙 — 이 카드는 프로젝트 안의 파일이에요.
- 훅 실행 전 검사와 `.claudeignore` 파일 자체의 문법 — 파일은 그대로 두고 규칙으로 옮기는 것만.
- 프로젝트 규모별 생성·벤더 폴더 거절(큰 저장소 카드)과 키 파일 질의(의사 카드).

## Requirements

- REQ-1: When the project setup or update previews a project, the planner shall list every Read deny rule it proposes for the key files found in that project, writing nothing before the user's pick. · failure: none — the preview names the matched files and their proposed rules and touches no settings file.
- REQ-2: The planner shall build its proposals from one documented pattern list of credential file shapes (`.env` variants, `*.pem`, `*.key`, `*-service-account.json`, `.npmrc`, `.pypirc` and the like), marking the patterns that matched files in this project. · failure: none — a pattern that matched nothing proposes no rule.
- REQ-3: When the project has a `.claudeignore` file, the planner shall offer each of its entries as the equivalent Read deny rule in the same plan, keeping the entry's text verbatim, negations included. · failure: none — the plan names the file and shows the one-to-one mapping.
- REQ-4: The applied rules shall land in the repository root's project settings file so worktree sessions pick them up. · failure: none — a rule that would land anywhere else is a finding.
- REQ-5: The write of the rules shall go through the checked settings writer (the settings write check) with its backup, merge, validation and doctor steps. · failure: none — a direct settings write that skips the check is a finding.
- REQ-6: The planner shall propose only `Read(path)` rule forms, never `Write(path)`·`Glob(path)`·`Grep(path)` shapes that match no call. · failure: none — a proposed rule in a never-matching shape is a build error of the plan.
- REQ-7: Retiring the `.claudeignore` file shall be a second explicit pick offered in the same preview, and the file is never deleted without it. · failure: none — the rules apply whether or not the file is retired, and the preview says the file has no effect either way.
- REQ-8: The preview shall state that deny rules apply without the workspace-trust step and in headless runs. · failure: none — the statement is one line in the preview.
- REQ-9: Every test of the planner shall run in a temporary project tree with a temporary config folder, writing no real user settings. · failure: none — a test that touches a real settings file is a finding.

## Acceptance

- AC-1 (REQ-1, REQ-2): the preview lists the Read deny rules for the key files found in a temporary project and writes nothing without the pick.
  verify: `cargo test --locked --offline --test secret_deny_cli preview_lists_rules_and_writes_nothing` — red first in the code slice
- AC-2 (REQ-3, REQ-7): the `.claudeignore` entries of the project appear as the equivalent Read deny rules in the plan, and retiring the file needs the second explicit pick.
  verify: `cargo test --locked --offline --test secret_deny_cli claudeignore_entries_become_read_deny_rules` — red first
- AC-3 (REQ-4, REQ-5, REQ-6): after the pick the rules land in the repository root's project settings through the checked writer, every rule in `Read(path)` form, and a session reading the root file from a worktree-shaped path sees them.
  verify: `cargo test --locked --offline --test secret_deny_cli rules_land_in_root_settings_through_the_checked_writer` — red first
- AC-4 (REQ-8): the preview's one line names the no-trust and headless behavior of deny rules.
  verify: `cargo test --locked --offline --test secret_deny_cli preview_states_trust_and_headless_behavior` — red first
- AC-5: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`

## Scope

- write: `.moai/specs/SPEC-20261122-secret-file-deny-rules/**`
- later code slices on this card (joined by card scope when their turn come): one planner module under `src/` (planned name fixed in the code slice), the test target `tests/secret_deny_cli.rs` with temporary project fixtures, and the preview and apply steps it hooks into on the project-setup route
- read-only: `.moai/specs/SPEC-20261122-settings-write-check/**` (the checked writer it calls), `.moai/specs/SPEC-20261122-headless-run-contract/**` (the stand-in and temp-config test patterns)
- forbidden: writes to any real user settings file or the real `.claudeignore` of this repository outside the user's pick; a direct settings write that skips the checked writer; force pushes; pushes of the integration branch

## Plan

1. This slice: this SPEC, committed on its branch, READY note with head and base; a moved tip is not re-merged.
2. Code slice A (after the settings-write-check SPEC lands): the planner and its preview, red tests first (REQ-1 to REQ-3, REQ-6, REQ-8, AC-1·AC-2·AC-4).
3. Code slice B: the apply step through the checked writer into the root settings and the second pick for retiring `.claudeignore` (REQ-4·REQ-5·REQ-7, AC-3).
4. Code slice C: the project-setup and update routes offer the preview (one entry among their picks).

## Risks

- 패턴 목록이 좁으면 비밀이 새고 넓으면 정상 파일이 막혀요: 목록은 문서 인용 한 곳(secure-deployment 표)에서 와고, 미리보기가 실제로 맞은 파일만 제안해요.
- `Read` 거절은 간접 읽기(스크립트가 파일을 여는 경우)를 못 막아요(permissions.md 경고): 미리보기가 샌드박스를 함께 켜라고 한 줄 권해요; 강제는 별개 기능.
- NotebookEdit 은 `Read` 거절 밖이에요: 위험으로만 기록(규칙을 더 늘리지 않아요).
- `.claudeignore` 항목이 gitignore 문법을 이미 쓰므로 변환은 거의 그대로 복사예요 — 그래서 항목 텍스트를 손대지 않는 규칙(REQ-3)이 와요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 거절 규칙은 `Read(path)` 꼴만 (`Edit` 덮임: 같은 경로의 Edit·Write·새 파일 생성까지 차단). | permissions.md "A `Read` deny rule also blocks the Edit and Write tools on the same path" | 카드 열림 |
| D-2 | 규칙은 저장소 루트의 프로젝트 설정 파일에 — 작업트리가 읽는 자리. | large-codebases.md "a second copy in the repository root's `.claude/settings.json` so worktree sessions pick them up" | 카드 열림 |
| D-3 | 쓰기는 설정 쓰기 검사를 통과해서(백업·병합·검증·doctor); 이 카드는 직접 쓰기를 만들지 않는다. | settings-write-check SPEC, 가드 파일은 사용자 선택 | 카드 열림 |
| D-4 | `.claudeignore` 항목은 텍스트 그대로 `Read(<항목>)` 로 옮긴다(부정 `!` 포함 — Read 규칙도 gitignore 부정을 읽어요); 파일 자체는 두 번째 선택 없이 지우지 않고, 어느 쪽이든 효과가 없음을 미리보기가 말한다. | permissions.md ".claudeignore … has no effect", "A deny or ask pattern that starts with `!` is a gitignore negation" | 카드 열림 |
| D-5 | 미리보기는 거절 규칙이 신뢰 단계 없이·무인 실행에서도 적용됨을 한 줄로 말한다. | permissions.md "deny and ask rules aren't affected, since they only restrict" | 카드 열림 |
| D-6 | 테스트는 임시 프로젝트 나무 + 임시 설정 폴더만. | "테스트는 Rust 만" 결정, 카드 기준 1 | 카드 열림 |

## Result

(status draft — written when the slice lands)
