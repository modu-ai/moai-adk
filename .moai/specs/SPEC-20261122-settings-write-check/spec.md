---
id: SPEC-20261122-settings-write-check
title: "Every settings file moai writes is backed up first, merged key by key, re-read and validated before an atomic swap, doctor-checked after, and rolled back and named on rejection"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# Every settings file moai writes is backed up first, merged key by key, re-read and validated before an atomic swap, doctor-checked after, and rolled back and named on rejection

## Intent

설정 쓰기 검사. 공식 문서의 사실(2026-11-22 받은 원문에서 확인): 사용자·프로젝트·로컬 설정 파일은 "a file whose JSON or top-level shape fails validation is rejected as a whole and reported, and an individual entry that fails, such as a malformed permission rule, is skipped with a warning while the rest of the file applies"(managed-settings.md), 그리고 `-p`는 깨진 파일을 조용히 건너뛰어요(settings.md) — 즉 moai가 하나의 잘못된 값을 쓰면 사용자의 설정 파일 전체가 조용히 죽어요. 훅 매처가 배열이면 그 파일의 다른 훅까지 깨져요("If the array is under `PreToolUse` or `PermissionRequest`, none of that file's other hooks load either" — debug-your-config.md). `Write(<path>)`·`NotebookEdit(<path>)` 규칙은 아무 호출도 못 맞추고 `Edit(<path>)`가 정답이에요(agent-sdk__permissions.md); 와일드카드가 하위 명령 앞에 있으면 `git -c …`까지 허용해요(errors.md 의 구성 경고); `env` 값은 셸이 처리하지 않고 빈 문자열은 해제(env-vars.md); `attribution` 은 객체 꼴(settings-reference.md). 프로젝트 설정은 3-3 목록의 키를 못 정해요(`defaultMode: "auto"`·`"bypassPermissions"`, `env` 의 `HOME`·`TMPDIR`·`CLAUDE_CONFIG_DIR`…, `pluginConfigs`, `remoteControlAtStartup: true`, `denyInsteadOfPrompting`, `crossSessionInbound` 는 더 엄격하게만 — settings.md). `claude doctor` 는 세션 없이 설정 진단을 읽기만 출력해요(cli-reference.md). 이미 있는 것: 3중 병합 계획과 충돌 토큰(`src/claude.rs:709` plan_settings), 교체 확정 때의 원본 보관(`src/claude.rs` 의 SETTINGS_BACKUPS), 원자 교체(`src/agents.rs:1568` write_replace → apply::publish_*), 소유 영수증(manifests/claude-settings.json). 없는 것: 모든 쓰기의 항상 백업, 바꿔 끼우기 전 되읽기·금지 목록 검사, 쓴 뒤 doctor 확인과 되돌림, 안내 한 줄. 이 조각은 SPEC 만 쓰고(코드는 이어지는 조각), 테스트는 Rust 만이며 진짜 사용자 설정 폴더는 건드리지 않아요.

## Non-goals

- 설정 파일에 어느 키를 어떤 값으로 쓸지(쓰는 내용의 목록) — 각 내용 카드(스타일·상태줄·훅·env 프로필).
- 관리(managed) 설정 파일의 쓰기·검증 — moai는 관리 파일을 쓰지 않아요.
- `~/.claude.json`(글로벌 구성)의 쓰기 — 별도 파일, 별도 규칙.
- 훅 실행 검사(preflight)와 `.claudeignore`·Read 거절 규칙 — 각자의 카드.
- 설정 폴더 여러 벌의 선택·설치 흐름 — 설치·업데이트 카드; 이 SPEC 은 한 파일을 쓰는 순간의 검사만 정해요.

## Requirements

- REQ-1: When moai writes any settings file, the writer shall keep a byte-exact backup of the existing file under the state root's backup folder before the first byte of the new file is written. · failure: none — a write with no place to keep the backup is refused, and the backup of a file that does not exist yet is recorded as absent rather than skipped.
- REQ-2: The writer shall change only the keys the plan names, keeping every other user key and every user list entry and merging lists by adding entries without removing the user's. · failure: none — a plan whose merge drops a user key or list entry is refused before the swap.
- REQ-3: Before the swap the writer shall re-read the produced bytes and parse them as one JSON object, refusing the write when the round-trip fails. · failure: none — a candidate that does not parse back is never written.
- REQ-4: The writer shall refuse, before any write, an entry on the forbidden list: a hook matcher that is not one string, a `Write(path)` or `NotebookEdit(path)` rule where `Edit(path)` is the form that governs, a command rule with a wildcard before the rest of the command, a bare `Edit` in a committed project allowlist, and an `attribution` value that is not an object. · failure: none — each refusal names the entry and the rule it broke.
- REQ-5: The writer shall refuse an `env` value that is not a string, warning on a value that looks like shell expansion and naming that settings `env` values are literal and an empty string means unset. · failure: none — the warning text carries no credential material.
- REQ-6: When the target is a project-scope file, the writer shall refuse a key a project file cannot set and name the scope that may set it. · failure: none — the refused keys are the documented project-scope exclusions (`permissions.defaultMode` of `auto` or `bypassPermissions`, `autoMode`, an `env` name a project file ignores, `pluginConfigs`, `remoteControlAtStartup: true`, `permissions.denyInsteadOfPrompting`, `skipDangerousModePermissionPrompt`, a `crossSessionInbound` loosening, global-config-only keys).
- REQ-7: After the swap the writer shall run `claude doctor` read-only and, when its diagnostics name a rejected entry for the written file, roll the file back from the backup byte-exactly and name the rejected entry. · failure: none — a doctor that cannot run leaves the write in place with one notice naming the skipped check.
- REQ-8: A refused or rolled-back write shall leave the previous file byte-identical, with the refusal named in the answer. · failure: none — the old file's bytes after a refusal equal its bytes before the plan.
- REQ-9: After a successful write the writer shall print one guide line naming the `/status` Setting sources row as the place to confirm the file loaded. · failure: none — the line prints once per write.
- REQ-10: Every test of the writer shall run against a temporary config folder and a stand-in `claude` for the doctor check, never the real user settings and never a real session. · failure: none — a test that touches the real config folder is a finding.
- REQ-11: The Result of this SPEC shall record, from the real update previews of the release it lands in, the count of writes the new refusals would have refused, so a false positive is seen before users meet it. · failure: none — the count is recorded with the landing.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4, REQ-8): a write whose candidate carries an invalid value (an array hook matcher, a `Write(path)` rule, a wildcard-first command rule, a bare `Edit` in a project allowlist, a non-object `attribution`) is refused before the swap, each refusal names the entry, and the old file stays byte-identical, in a temporary config folder.
  verify: `cargo test --locked --offline --test settings_write_cli invalid_values_refuse_before_the_swap` — red first in the code slice
- AC-2 (REQ-6): a project-scope write naming a key a project file cannot set is refused with the key and the scope that may set it named.
  verify: `cargo test --locked --offline --test settings_write_cli project_only_keys_refuse_in_project_scope` — red first
- AC-3 (REQ-1, REQ-7, REQ-8): a write whose doctor names a rejected entry is rolled back from the backup byte-exactly and the rejected entry is named, through a stand-in `claude`.
  verify: `cargo test --locked --offline --test settings_write_cli doctor_rejection_rolls_back_byte_exact` — red first
- AC-4 (REQ-2, REQ-9): a valid write keeps every user key and list entry it did not plan and prints the one guide line.
  verify: `cargo test --locked --offline --test settings_write_cli valid_write_keeps_user_keys` — red first
- AC-5 (REQ-5): a non-string `env` value refuses and a value that looks like shell expansion warns without echoing any credential material.
  verify: `cargo test --locked --offline --test settings_write_cli env_values_refuse_and_warn` — red first
- AC-6: the wider READY set passes.
  verify: `cargo test --locked --offline --bins` and `--lib`

## Scope

- write: `.moai/specs/SPEC-20261122-settings-write-check/**`
- later code slices on this card (joined by card scope when their turn come): the settings writer in `src/claude.rs` (the plan/apply path around `plan_settings`, `SETTINGS_BACKUPS` and the `write_replace` calls) or its extracted module under `src/`, the test target `tests/settings_write_cli.rs` with its stand-in `claude` fixture, and the one guide line in the apply answer
- read-only: `src/agents.rs` (the atomic `write_replace`), `src/apply.rs`, `.moai/specs/SPEC-20261122-headless-run-contract/**` (the stand-in claude pattern)
- forbidden: writes to the real user config folder or the real global config file anywhere in tests; a real `claude` session in tests; force pushes; pushes of the integration branch

## Plan

1. This slice: this SPEC, committed on its branch, READY note with head and base; a moved tip is not re-merged.
2. Code slice A: the writer's new gate red tests first — backup-before-write, re-read round-trip, the forbidden list, project-only keys (AC-1, AC-2, AC-5).
3. Code slice B: the doctor check with rollback and the guide line (AC-3, AC-4), stand-in `claude` fixture shared with the headless-run tests.
4. Code slice C: the existing update and project-setup call sites adopt the checked writer; the would-refuse count of the real update previews goes into the Result (REQ-11).

## Risks

- 금지 목록이 너무 넓으면 정상 업데이트가 막혀요: REQ-11 의 실측 횟수가 첫 릴리스 전 거짓 양성을 보여요.
- `claude doctor` 는 버전에 따라 출력이 달라져요: 되돌림 판정은 doctor 가 "거부된 항목"을 그 파일로 지목할 때만 일어나고, 판정 규칙은 코드 조각에서 문장 단위로 정해요(테스트가 고정).
- 백업 폴더가 사용자의 정리 대상이 될 수 있어요: 백업은 상태 루트의 영수증 폴더 아래(기존 SETTINGS_BACKUPS 와 같은 자리)로, 정리 카드의 기준값을 따라요.
- 프로젝트 전용 키 목록은 문서가 늘어요: 목록은 코드 한 곳에 두고 문서 인용을 붙여요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 쓰기 순서는 보고서 3-4의 여섯 단계를 그대로: 백업 → 병합 → 되읽기·금지 검사 → 원자 교체 → doctor → 안내. 이미 있는 3중 병합·원자 교체·영수증은 다시 만들지 않고, 없는 단계만 채운다. | 리더 보고서 3-4, 기존 코드(`src/claude.rs:709`, `src/agents.rs:1568`) | 카드 열림 |
| D-2 | 백업은 교체 확정 때만이 아니라 모든 쓰기에 남긴다(기존 파일이 없으면 부재로 기록). | settings.md 의 "rejected as a whole" 위험, 기존 SETTINGS_BACKUPS 확장 | 카드 열림 |
| D-3 | 되돌림 판정: doctor 진단이 쓴 파일을 향한 거부 항목을 이름할 때만. doctor 를 못 돌리면 쓰기는 유지하고 검사 건너뜀을 한 줄 알림(실패가 쓰기를 되돌리게 하지 않는다). | cli-reference.md "read-only installation and settings diagnostics", 무인 흐름 안전 | 카드 열림 |
| D-4 | 금지 목록은 문서가 이름하는 것만: 배열 매처, Write/NotebookEdit 경로 규칙, 하위 명령 앞 와일드카드, 커밋되는 allowlist 의 맨 `Edit`, 객체 아닌 attribution. 그 밖의 정책은 늘리지 않는다. | debug-your-config.md, agent-sdk__permissions.md, errors.md, settings-reference.md | 카드 열림 |
| D-5 | 프로젝트 범위 전용 키 거부는 settings.md 3-3 표를 코드 한 곳의 목록으로. | settings.md "Troubleshoot a setting that doesn't apply" | 카드 열림 |
| D-6 | 테스트는 임시 설정 폴더 + 스탠드인 claude(doctor 출력 가짜)만. | "테스트는 Rust 만" 결정, 카드 기준 1(임시 홈) | 카드 열림 |

## Result

(status draft — written when the slice lands)
