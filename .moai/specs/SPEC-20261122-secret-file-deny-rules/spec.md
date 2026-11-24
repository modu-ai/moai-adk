---
id: SPEC-20261122-secret-file-deny-rules
title: "Key files found in a project are offered as Read deny rules in the root project settings, with the .claudeignore entries carried over, only on the user's pick"
status: draft
revision: 2
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# Key files found in a project are offered as Read deny rules in the root project settings, with the .claudeignore entries carried over, only on the user's pick

## Intent

비밀 파일 거절 규칙. 공식 문서의 사실(2026-11-22 받은 원문에서 확인): (1) "If your project has a `.claudeignore` file, it has no effect, so move its entries into `Read` deny rules"(permissions.md) — `.claudeignore` 는 아무 효과가 없어요. (2) "`@` 언급은 훅을 거치지 않음": "Your `PreToolUse` hooks don't run when Claude Code includes a file you reference with `@`, because it makes no tool call"(common-workflows.md) — 훅으로 읽기를 막는 길은 `@` 언급을 놓쳐요; 거절 규칙은 `@` 언급·붙여넣은 그림 경로·디렉터리 목록에도 최선을 다해 적용돼요(permissions.md). (3) 거절 규칙은 모든 모드에서 이겨요("deny, then ask, then allow"), 작업공간 신뢰 단계를 기다리지 않아요("deny and ask rules aren't affected, since they only restrict" — permissions.md). (4) 공유 프로젝트 설정은 세션의 주 작업 디렉터리에서 읽혀요(settings.md) — 작업트리는 자기 체크아웃이므로, 규칙이 작업트리에 닿으려면 저장소 루트의 설정 파일이 git 에 추적돼야 해요; 이 저장소는 지금 `.claude/` 를 무시해서(.gitignore) 그 경우예요. (5) `Read(path)` 거절은 같은 경로의 Edit·Write(새 파일 만들기 포함)까지 막아요; `Write(path)`·`Glob(path)`·`Grep(path)` 꼴은 아무것도 못 맞춰요(permissions.md). (6) 같은 파일 안에서 `!` 항목은 자기 앞의 상대 거절을 다 파내요(permissions.md). (7) 어떤 파일이 자격 증명인지: secure-deployment 의 표(`.env`·`*.pem`·`*.key`·`*-service-account.json`·`.npmrc`·`.pypirc` …). 무인 레인이 이 저장소의 비밀 파일을 읽는 사고가 실제로 일어났고(안내 문구 규칙의 유래), 이 조각은 그 문을 닫아요. 규칙은 프로젝트 설정이므로 미리보기 뒤 사용자의 선택으로만 쓰여요. 첫 조각은 SPEC 만 쓰고, 테스트는 Rust 만이에요.

## Non-goals

- 운영체제 수준의 차단(샌드박스) — Read 거절은 "arbitrary subprocesses that read or write files indirectly" 에 닿지 않아요(permissions.md 경고); 샌드박스는 별개의 기능이에요.
- NotebookEdit 경로 — `Read` 거절이 덮지 못하는 한 곳(permissions.md); 위험에 기록만 해요.
- 사용자 범위·전역 경로(`~/`·`//`) 거절 규칙 — 이 카드는 프로젝트 안의 파일이에요.
- 훅 실행 전 검사와 `.claudeignore` 파일 자체의 문법 — 파일은 그대로 두고 규칙으로 옮기는 것만.
- 프로젝트 규모별 생성·벤더 폴더 거절(큰 저장소 카드)과 키 파일 질의(의사 카드).
- `.gitignore` 의 `.claude/` 무시를 좁히는 변경 자체 — 별개의 변경(무시 좁히기)이 전제를 바꿔요; 이 SPEC 은 감지와 알림만.

## Requirements

- REQ-1: When the project-setup route previews a project, the planner shall list every Read deny rule it proposes for the key files found in that project, writing nothing before the user's pick and only hinting from the update route. · failure: none — the preview names the matched files and their proposed rules and touches no settings file.
- REQ-2: The planner shall build its proposals from one documented pattern list of credential file shapes (`.env` variants, `*.pem`, `*.key`, `*-service-account.json`, `.npmrc`, `.pypirc` and the like), walking the project by file names only — ignored files included, no symlink followed out of the project — and marking the patterns that matched files in this project. · failure: none — a pattern that matched nothing proposes no rule.
- REQ-3: When the project has a `.claudeignore` file, the planner shall offer each of its entries — comments and blank lines skipped, a trailing `/` kept — as one Read deny rule the user picks one by one, warning that a deny rule also blocks editing the path and marking the entries that match tracked files. · failure: none — the plan names the file and shows one row per entry with its marks.
- REQ-4: The planner's matching shall be by file name only, never opening, reading, hashing or printing a matched file. · failure: none — a matched fixture with mode 000 or a planted secret value proves it: the preview names the file and neither its content nor an opening error appears.
- REQ-5: The plan shall order the key-file rules after every carried-over `.claudeignore` entry and list each `!` carve-out that reaches a key-file rule. · failure: none — a carried-over negation that would carve into a key-file rule is named in the plan before the pick.
- REQ-6: The applied rules shall land in the tracked repository-root project settings file — so worktree sessions pick them up — with the preview detecting an ignored or untracked root settings file and saying the rules will not reach worktrees. · failure: none — the write is refused while the root file stays ignored or untracked and the preview says why.
- REQ-7: The write of the rules shall go through the checked settings writer (the settings write check) with its backup, merge, validation and doctor steps. · failure: none — a direct settings write that skips the check is a finding.
- REQ-8: The planner shall propose only `Read(path)` rule forms, never `Write(path)`·`Glob(path)`·`Grep(path)` shapes that match no call. · failure: none — a proposed rule in a never-matching shape is a build error of the plan.
- REQ-9: Retiring the `.claudeignore` file shall be a second explicit pick offered in the same preview, and the file is never deleted without it. · failure: none — the rules apply whether or not the file is retired, and the preview says the file has no effect either way.
- REQ-10: The preview shall state that deny rules apply without the workspace-trust step and in headless runs whenever project settings load. · failure: none — the same line names the exception: a run started with `--setting-sources user`, or an SDK session without project settings, does not read the rules.
- REQ-11: Every test of the planner shall run in a temporary project tree with a temporary config folder, writing no real user settings. · failure: none — a test that touches a real settings file is a finding.

## Acceptance

- AC-1 (REQ-1, REQ-2): the preview lists the Read deny rules for the key files found in a temporary project and writes nothing without the pick.
  verify: `cargo test --locked --offline --test secret_deny_cli preview_lists_rules_and_writes_nothing` — red first in the code slice
- AC-2 (REQ-3, REQ-9): each `.claudeignore` entry of the project appears as one Read deny rule the user picks one by one with the edit-blocking warning and the tracked-file marks, and retiring the file needs the second explicit pick.
  verify: `cargo test --locked --offline --test secret_deny_cli claudeignore_entries_are_picked_one_by_one` — red first
- AC-3 (REQ-7, REQ-8): after the pick the rules land in the root project settings through the checked writer, every rule read back from the written file in `Read(path)` form.
  verify: `cargo test --locked --offline --test secret_deny_cli rules_land_in_root_settings_through_the_checked_writer` — red first
- AC-4 (REQ-6): with the root settings file ignored by git in the fixture, the preview says the rules will not reach worktrees and the write is refused.
  verify: `cargo test --locked --offline --test secret_deny_cli ignored_root_settings_warns_worktrees` — red first
- AC-5 (REQ-4): a matched fixture file with mode 000 (and one carrying a planted secret value) is never opened — the preview names it and neither its content nor an opening error appears.
  verify: `cargo test --locked --offline --test secret_deny_cli matched_file_is_never_opened` — red first
- AC-6 (REQ-5): the plan orders the key-file rules after the carried-over entries and lists the `!` entry that carves into a key-file rule.
  verify: `cargo test --locked --offline --test secret_deny_cli negation_ordering_is_listed` — red first
- AC-7 (REQ-10): the preview's one line names the no-trust, headless and `--setting-sources user` behavior of deny rules.
  verify: `cargo test --locked --offline --test secret_deny_cli preview_states_trust_and_headless_behavior` — red first
- AC-8: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`

## Scope

- write: `.moai/specs/SPEC-20261122-secret-file-deny-rules/**`
- later code slices on this card (joined by card scope when their turn come): one planner module under `src/` (planned name fixed in the code slice), the test target `tests/secret_deny_cli.rs` with temporary project fixtures, and the preview and apply steps it hooks into on the project-setup route
- read-only: `.moai/specs/SPEC-20261122-settings-write-check/**` (the checked writer it calls), `.moai/specs/SPEC-20261122-headless-run-contract/**` (the stand-in and temp-config test patterns), `.gitignore` (the root-settings tracked check reads it)
- forbidden: writes to any real user settings file or the real `.claudeignore` of this repository outside the user's pick; a direct settings write that skips the checked writer; opening, reading or hashing a matched key file; force pushes; pushes of the integration branch

## Plan

1. This slice: this SPEC (revision 2 answers the independent review), committed on its branch, READY note with head and base; a moved tip is not re-merged.
2. Code slice A (after the settings-write-check SPEC lands): the planner and its preview, red tests first (REQ-1 to REQ-5, REQ-8, REQ-10, AC-1·AC-2·AC-4 to AC-7).
3. Code slice B: the apply step through the checked writer into the root settings and the second pick for retiring `.claudeignore` (REQ-6·REQ-7·REQ-9, AC-3).
4. Code slice C: the project-setup route offers the preview (one entry among its picks) and the update route hints at it.

## Risks

- 패턴 목록이 좁으면 비밀이 새고 넓으면 정상 파일이 막혀요: 목록은 문서 인용 한 곳(secure-deployment 표)에서 와고, 미리보기가 실제로 맞은 파일만 제안해요.
- `Read` 거절은 간접 읽기(스크립트가 파일을 여는 경우)를 못 막아요(permissions.md 경고): 미리보기가 샌드박스를 함께 켜라고 한 줄 권해요; 강제는 별개 기능.
- NotebookEdit 은 `Read` 거절이 덮지 못해요: 위험으로만 기록(규칙을 더 늘리지 않아요).
- `.claudeignore` 항목이 gitignore 문법을 이미 쓰므로 변환은 거의 그대로 복사예요 — 그래서 항목 텍스트를 손대지 않는 규칙(REQ-3)이 와요.
- 루트 설정이 무시되는 저장소(이 저장소가 지금 그래요)에서는 규칙이 작업트리에 닿지 않아요: 미리보기가 감지해 알리고 쓰기를 거부해요; `.gitignore` 무시 좁히기 변경이 그 전제를 바꿀 때까지 이 저장소의 적용은 미뤄져요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 거절 규칙은 `Read(path)` 꼴만 (`Edit` 덮임: 같은 경로의 Edit·Write·새 파일 생성까지 차단) — 그래서 항목마다 "편집도 막힘" 경고가 붙어요. | permissions.md "A `Read` deny rule also blocks the Edit and Write tools on the same path"; 리뷰 6130 (2) | 카드 열림 |
| D-2 | 규칙은 git 에 추적되는 저장소 루트의 프로젝트 설정 파일에 — 작업트리가 읽는 자리; 무시되거나 추적 안 된 루트 파일은 감지해 알리고 쓰기를 거부한다. | settings.md 주 작업 디렉터리 읽기, 리뷰 6130 (1), 이 저장소 .gitignore 의 `.claude/` 무시 | 카드 열림 |
| D-3 | 쓰기는 설정 쓰기 검사를 통과해서(백업·병합·검증·doctor); 이 카드는 직접 쓰기를 만들지 않는다. | settings-write-check SPEC, 가드 파일은 사용자 선택 | 카드 열림 |
| D-4 | `.claudeignore` 항목은 텍스트 그대로 `Read(<항목>)` 후보가 되지만 하나씩 사용자가 고른다(주석·빈 줄 건너뛰기, 끝 `/` 유지); 파일 자체는 두 번째 선택 없이 지우지 않는다. | permissions.md ".claudeignore … has no effect"; 리뷰 6130 (2)·N3 | 카드 열림 |
| D-5 | 키 파일 규칙은 옮겨 온 항목 뒤에 둔다 — 같은 파일의 `!` 항목이 자기 앞의 상대 거절을 파내므로; 파내기가 키 파일 규칙에 닿으면 계획이 이름한다. | permissions.md "A deny or ask pattern that starts with `!` is a gitignore negation… carves the paths it matches out of the `path` or `./path` rules listed before it"; 리뷰 6130 (3) | 카드 열림 |
| D-6 | 미리보기는 거절 규칙이 신뢰 단계 없이·프로젝트 설정을 읽는 무인 실행에서도 적용됨을 한 줄로 말한다(예외 포함: `--setting-sources user`·SDK 프로젝트 설정 없음). | permissions.md "deny and ask rules aren't affected"; permissions.md 841; 리뷰 6130 N2 | 카드 열림 |
| D-7 | 짝 맞춤은 이름으로만 — 플래너는 맞은 파일을 열지·읽지·해시하지·출력하지 않는다. | 리뷰 6130 (4) | 카드 열림 |
| D-8 | 테스트는 임시 프로젝트 나무 + 임시 설정 폴더만. | "테스트는 Rust 만" 결정, 카드 기준 1 | 카드 열림 |

## Result

(status draft — written when the slice lands)
