---
id: SPEC-20261124-project-detect-monorepo
title: "project-detect finds the sub-package manifests of a monorepo in any language, and a profile apply never drops recorded manifests unseen"
status: approved
revision: 2
created: 2026-11-24
approval: {by: leader, at: 2026-11-24, basis: "leader decision note 8296 of card C-256 (answers lane note 8295, 2026-11-24T06:46:32Z, under the user's delegation of 2026-11-23T13:57Z, not a direct user pick), approving revision 2 at head c68d78dfb7 with D-5 and D-6 as written; recorded by the lane, not approved by it", digest: "sha256:254b58336e0c19df493e36ab21afc1eff565761931ee1a8445c5d087e638dc15"}
---

# project-detect finds the sub-package manifests of a monorepo in any language, and a profile apply never drops recorded manifests unseen

## Interview

- Q1: MoAI 는 어떤 개발 언어를 지원해야 하고, 감지 규칙을 언어에 묶어도 되나요? -> A: "MoAI 는 Rust 만이 아니라 모든 개발 언어를 지원해야 하고 언어·도구 하드코딩은 금지예요 — 감지 규칙도 특정 언어에 묶지 말아요."(사용자 지시 2026-11-24, 이 카드의 디스패치 문장; 같은 지시가 카드 C-150 의 노트 8151·8158 에도 있어요)
- Q2: 이 SPEC 의 설계 선택(카탈로그와 확장, 숨김 폴더 건너뛰기, 한 SPEC 으로 묶기, 이 SPEC 의 범위)은? -> A: 아직 답이 없어요. 승인 요청 노트(`LANE-BLOCKED kind=decision`)가 Decisions 표의 D-1 ~ D-4 를 묻고, 답은 승인 때 이 항목에 옮겨 적어요.

## Intent

보고(mo.ai.kr 리더, 2026-11-21, moai-cli 4.2647.12): 모노레포(`apps/*/package.json`, `tsconfig.json`, `apps/id/go.mod` …)에서 `moai project-detect .` 가 루트 `package.json` 하나만 찾고, 프로필을 다시 적용하면 `detected.manifests` 가 55 개에서 1 개로 줄어요. 같은 저장소를 읽기만 하는 실행으로 원인을 재현했어요(카드 노트 4789): 하위 manifest 탐색(`SPEC-20261119-nested-manifests`)이 폴더를 이름 순으로 깊이 먼저 훑다가 항목 5000 개에서 조용히 멈추는데, `apps/` 보다 앞서는 숨김 폴더(`.moai-backup`, `.moai-backups`, `.vercel`, 합쳐 약 15 만 항목)가 예산을 다 써요. 적용이 목록을 줄이는 것을 미리보기가 말하지 않은 것도 같은 사고의 일부예요.

세 가지를 고쳐요.

1. 탐색: 얕은 폴더를 모두 읽은 뒤에 한 단계 깊은 폴더로 가고, 이름이 `.` 으로 시작하는 폴더는 걷지 않고, 목록이 부분적이면 `manifest_walk` 로 알려요. 세 규칙 모두 어떤 언어 이름도 쓰지 않아요.
2. 인식: 무엇이 manifest 인지(파일 이름이나 접미사 → 생태계·언어·도구)와 걷지 않을 폴더 이름(`node_modules`, `target`, `vendor` …)을 코드의 상수 표에서 데이터 카탈로그로 옮기고, 프로젝트가 자기 행을 덧붙일 수 있게 해요. 엔진은 카탈로그 행만 보고 파일 이름을 알지 못하니, 카탈로그에 없는 언어의 manifest 는 코드 변경 없이 행 하나로 더해져요. 공개 구현 둘이 같은 구조예요: Renovate 는 manager 마다 `managerFilePatterns` 를 선언하고 사용자 설정의 패턴을 "added to the default values" 로 더하고(`docs.renovatebot.com`, 읽은 때 2026-11-24T04:05Z), GitHub Linguist 는 `languages.yml` 의 `filenames`·`extensions` 데이터로 언어를 알고 `vendor.yml` 의 정규식 데이터로 vendored 경로를 걸러요(`github-linguist/linguist`, 같은 때). ripgrep 도 숨김 폴더를 기본으로 건너뛰어요(`docs.rs/ignore` WalkBuilder::hidden "enabled by default").
3. 보여 주기: 프로필 계획이 `detected` 키를 적용하면서 기록돼 있던 manifest 를 새 목록에서 뺄 때, 빠지는 이름을 계획과 업데이트 미리보기에 보여 줘요. 적용은 계획의 digest 확인을 거치고 digest 가 그 목록을 덮으니, 보여 준 것이 사용자의 결정이에요.

읽는 사람: 레인, 이 변경의 감사자, `/moai project`·`/moai update` 의 사용자.

## Non-goals

- 코드 파일의 언어 감지(`EXTENSIONS`), 언어 서버 표(`LSP_PLUGINS`), 데이터베이스 의존성 표(`DB_DEPENDENCIES`), `commands_fact` 의 빌드·테스트 명령 템플릿, `project_scale_detect` 의 작업공간 표지(`MARKER_FILES`)도 언어에 묶인 표예요. 이 SPEC 은 manifest 인식과 걷기 한 줄기만 데이터로 옮기고 나머지는 후속 카드로 적어요(Decisions D-4).
- 한도 상수(`MAX_DEPTH` 3, `MAX_ENTRIES` 5000, `MAX_ITEMS` 64)를 바꾸지 않아요. 한도를 넘으면 조용히 줄이지 않고 알려요. 판 2 는 `.gitignore` 총량 한도 하나(`MAX_IGNORE_TOTAL`, Decisions D-6)만 더해요.
- 하위 manifest 의 내용 읽기, 심볼릭 링크 따라가기, 프로세스 시작은 하지 않아요. 걷기는 이름만 읽어요.
- 기록된 manifest 목록을 몰래 합치거나 되돌리지 않아요. 보여 주기만 해요.
- 스킬 글(`skills/**`)과 `CLAUDE.md`, 인터뷰·언어·`project-apply` 입력 읽기(카드 C-150 의 파일 `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`)는 건드리지 않아요.
- 미착지 가지 `spec/20261121-detect-walk-breadth`(머리 `c095753cd1`)와 `spec/20261121-plan-shows-dropped-manifests`(머리 `2f3784219a`)는 지우지 않고 복구 참고로 남겨요. 이 SPEC 이 두 가지의 설계와 시험을 이어받아 새 팁에서 다시 구현해요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: The manifest walk shall read every folder of one depth before it reads any folder one level deeper.
- REQ-2: The manifest walk shall not enter a folder whose name starts with a dot or that the catalog lists as a folder to skip. · failure: none — an ignore file past a total of 1048576 bytes read counts as no lines, which lets more manifests in and hides none.
- REQ-3: When `project-detect` reads a project folder, the command shall list in `detected.manifests` each file whose name equals a catalog file name or ends with a catalog suffix, by its path relative to the project root. · failure: none — an unreadable folder is left out and `manifest_walk` says so.
- REQ-4: The shipped catalog shall hold the twelve root manifest names and the tool names that the earlier tables held, with the same language and tools for each, so that `detected.languages`, `detected.manifests` and `tools` keep their earlier values for those names.
- REQ-5: When the project folder holds a catalog extension file, the command shall add its rows to the shipped rows, the shipped row prevailing for a file name or suffix that both hold. · failure: none — the extension file is optional.
- REQ-6 (REQ-5): IF the extension file is a link, larger than the file limit, not valid, or holds a row or a folder to skip with a value that cannot be recorded (a name made of dots alone is one), THEN the command shall use the shipped rows alone and report the problem in `manifest_catalog.problem` without writing anything.
- REQ-7: When `project-detect` reports the facts, the command shall report beside `detected` a `manifest_walk` object holding `complete` (false when the walk stopped at `MAX_ENTRIES` or could not read a folder or an entry of one) and `over_cap` (the number of manifests found that `detected.manifests` leaves out because of `MAX_ITEMS`). · failure: none — the object is always reported.
- REQ-8 (REQ-7): IF `manifest_walk.complete` is false or `manifest_walk.over_cap` is more than zero, THEN `next_actions` shall carry a line saying the manifest list is partial and must not be recorded over an earlier list without showing the user what drops.
- REQ-9: The profile plan that applies the `detected` key over a recorded profile shall carry `dropped_manifests`, the sorted recorded `detected.manifests` entries that the new list leaves out, exactly when that list is not empty. · failure: none — a plan without the field is a plan that drops nothing.
- REQ-10: The plan digest shall cover `dropped_manifests` when it is not empty and keep its earlier value when it is empty.
- REQ-11: The profile step of the update plan shall say in its `reason` how many recorded manifests would drop and carry the sorted names in its `preview`.
- REQ-12 (REQ-10): IF a plan's `dropped_manifests` is not what the current files give, THEN `project-apply` shall refuse the plan as stale or altered and write nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2): 이름 순으로 앞서는 큰 폴더(깊이 3 에 파일 6000 개)가 있어도 뒤따르는 `apps/x/package.json` 이 나열되고 `manifest_walk.complete` 는 false 예요. `.cache/pkg/package.json`, `.vercel/x/package.json`, `.moai-backup/y/go.mod` 는 나열되지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_large_early_folder_does_not_hide_a_later_packages_manifest hidden_folders_are_not_walked_for_manifests`
- AC-2 (REQ-3, REQ-4): 카드의 임시 모노레포 — `apps/a/package.json`, `apps/b/tsconfig.json`, `apps/id/go.mod` 와 루트 `package.json` — 가 모두 나열되고, `tools` 와 `languages` 가 전과 같은 값이에요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_temporary_monorepo_lists_every_sub_manifest`
- AC-3 (REQ-2): 카탈로그가 걷지 않을 폴더로 올린 이름의 폴더(`node_modules`)와 카탈로그 확장이 올린 이름의 폴더는 걷지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli folders_the_catalog_lists_are_not_walked`
- AC-4 (REQ-3, REQ-5): 카탈로그에 없는 언어(`apps/svc/mix.exs`)는 확장 행 하나로 `detected.manifests` 와 `languages` 에 나타나고, 확장 파일을 지우면 사라져요. 엔진 코드는 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_row_in_the_extension_file_adds_a_language_without_code`
- AC-5 (REQ-5, REQ-6): 확장 행이 같은 이름의 배송 행을 바꾸려 해도 배송 행은 그대로예요. 링크이거나 한도를 넘거나 잘못된 확장 파일은 `manifest_catalog.problem` 으로 알려지고 배송 행만 쓰이며 트리가 바뀌지 않아요. 한도는 양쪽 경계에서 지켜져요: 확장 파일은 262144 바이트까지 읽히고 262145 바이트는 거절되며, 행·건너뛸 폴더·행 안의 파일·접미사·도구는 64 개까지 읽히고 65 개는 거절돼요. 점만으로 된 이름(`.`, `..`)은 어느 칸에서도 거절돼요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli an_extension_file_adds_rows_and_a_bad_one_is_reported_and_ignored the_extension_file_limit_holds_at_both_edges the_extension_row_and_list_limits_hold_at_both_edges a_name_of_only_dots_is_refused_in_the_extension_file`
- AC-6 (REQ-7, REQ-8): 작은 프로젝트는 `manifest_walk` 가 `{"complete": true, "over_cap": 0}` 이고 `next_actions` 에 부분 목록 줄이 없어요. manifest 70 개는 64 개가 나열되고 `over_cap` 6 에 줄이 하나 있어요. 예산 소진도 줄 하나예요. 읽을 수 없는 폴더(권한 000)가 있으면 읽을 수 있는 폴더의 manifest 는 그대로 나열되고 `complete` 는 false 에 줄이 하나예요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli the_manifest_walk_fact_reports_completeness_and_the_cap a_partial_list_adds_one_next_action_and_a_complete_one_adds_none a_folder_that_cannot_be_read_makes_the_walk_incomplete`
- AC-7 (REQ-9): 기록된 `detected.manifests` 가 `apps/a/package.json`, `apps/b/go.mod`, `package.json` 이고 새 목록이 `package.json` 뿐이면 `project-plan` 이 앞의 둘을 정렬해 `dropped_manifests` 로 보여요. 목록이 그대로이거나 늘기만 하면 필드가 없고, 손으로 고쳐서 충돌인 `detected` 키는 덮어쓰지 않으니 필드도 없어요 · verify: `cargo test --locked --offline --test project_update_drops_cli the_project_plan_names_the_recorded_manifests_a_smaller_detected_list_drops a_list_that_only_grows_or_stays_drops_nothing a_hand_edited_detected_key_drops_nothing`
- AC-8 (REQ-10, REQ-12): 목록이 비면 digest 는 전과 같은 값이고(박아 둔 기존 시험이 그대로 통과), `dropped_manifests` 를 지운 계획은 `project-apply` 가 "stale or altered plan" 으로 거절하고 프로필 파일이 그대로예요 · verify: `cargo test --locked --offline --test project_update_drops_cli a_plan_with_the_dropped_list_removed_is_refused_at_apply`
- AC-9 (REQ-11): `project-update-plan` 의 `profile` 단계가 `ready` 이고 `reason` 에 "2 recorded manifests would drop" 이 있고 `preview.dropped_manifests` 가 그 둘이에요 · verify: `cargo test --locked --offline --test project_update_drops_cli the_update_plan_profile_step_says_how_many_manifests_would_drop`
- AC-10: 이웃 시험이 그대로 통과해요 · verify: `cargo test --locked --offline --test project_detect_cli --test project_scale_detect_lib --test profile_apply_engine --test profile_plan_cli --test project_update_plan_cli --test home_isolation_scan` 과 `cargo test --locked --offline --lib project_detect`
- AC-12 (REQ-2): 크기가 정확히 262144 바이트인 `.gitignore`(루트와 하위 폴더)는 읽혀 제외가 적용되고 262145 바이트는 읽히지 않아요. 약 200 KB 짜리 `.gitignore` 가 일곱 폴더에 있으면 총량(1 MiB) 안의 다섯은 적용되고 여섯째부터는 줄이 없는 것으로 세요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli the_ignore_file_limit_holds_at_both_edges the_ignore_files_the_walk_reads_have_a_total_size_cap`
- AC-11 (REQ-4): 실측 — 보고된 저장소를 읽기만 하는 `project-detect` 로 돌려 `manifests` 가 55 개, `manifest_walk` 가 `{complete: true, over_cap: 0}` 이에요(고치기 전 1 개). 그 저장소를 읽을 수 없는 실행이면 `not_run` 으로 적어요 · verify: `moai project-detect <그 저장소>` 를 리더나 레인이 읽기 전용으로 실행

## Scope

- write: `.moai/specs/SPEC-20261124-project-detect-monorepo/**`, `src/project_detect.rs`(걷기, 카탈로그 읽기, `detect`·`detect_cli`), `src/project_detect_catalog.json`(새 파일, `include_str!` 로 넣어요), `src/ops/project_detect.rs`(설명 글), `src/project_scale_detect.rs`(`SKIP_DIRS` 를 쓰는 줄만), `src/profile.rs`(`Plan` 의 새 필드, `plan`, `plan_digest`), `src/project_update.rs`(`profile_step` 의 `reason`·`preview`), `tests/project_detect_monorepo_cli.rs`(새 파일), `tests/project_update_drops_cli.rs`(새 파일)
- write by line lease(다른 카드도 쓰는 공용 파일): `tests/areas.toml` 과 `tests/fixtures/state_folder_literals/allow.tsv`(새 시험 파일 둘의 등록 줄; 새 파일이 없어도 `check affected` 의 `test_tables` 가 요구해요), `tests/project_detect_cli.rs`(박아 둔 기대가 바뀔 때만; 카드 C-150 이 줄 임대로 써요)
- read-only: `src/ops/project_plan.rs`(계획을 그대로 내보내요), `src/project_scale.rs`, `src/state_root.rs`, `skills/moai/workflows/project.md`
- forbidden: `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`(카드 C-150), `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME, 한도 상수

## Plan

1. SPEC 승인을 기다려요(승인 전에는 코드·시험을 쓰지 않아요). 승인 뒤 카드에 SPEC 을 잇고 범위를 위 write 목록으로 넓혀요.
2. RED: `tests/project_detect_monorepo_cli.rs` 와 `tests/project_update_drops_cli.rs` 에 AC-1 ~ AC-9 시험을 먼저 써서 새 팁에서 실패하는 것을 봐요(옛 가지의 시험 다섯과 넷을 가져와 새 팁에 맞춰요; 시험은 `TempHome` 으로 바이너리를 실행해요).
3. GREEN, 인식: `src/project_detect_catalog.json`(배송 행 = 지금의 열두 manifest 와 도구, 건너뛸 폴더 이름) 을 만들고 `MANIFESTS`, `MANIFEST_TOOLS`, `SKIP_DIRS` 의 쓰임을 카탈로그 읽기(한 번 파싱해 두는 `OnceLock`)로 바꿔요. 확장 파일은 상태 폴더 아래 한 파일이고 읽기 규칙은 `read_in` 과 같아요(정규 파일, 링크 아님, `MAX_FILE` 이하). 표에 더하는 행은 언어 근거(공식 문서 주소와 읽은 때)를 SPEC Result 에 적어요.
4. GREEN, 걷기: `walk_manifests` 를 폴더 줄(queue)로 바꿔 한 깊이씩 읽고, 점 이름 폴더와 카탈로그의 폴더를 건너뛰고, 예산 소진을 `complete` 로 돌려줘요. `detect` 가 `over_cap` 을 세고 `detect_cli` 가 `manifest_walk`, `manifest_catalog` 와 안내 줄을 내요.
5. GREEN, 보여 주기: `profile::plan` 이 `detected` 가 적용될 때 줄어드는 항목을 구해 `Plan.dropped_manifests` 에 넣고 digest 에 더해요(비어 있으면 digest 를 전과 같게 해요). `profile_step` 이 `reason` 과 `preview` 에 보여요.
6. 확인: 새 시험 파일 둘, AC-10 의 이웃 시험, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`(대상 한정), 새 가드 시험마다 가드 뮤턴트 한 줄(`guard-mutant: <test> | <지운 코드> | exit <n>`), 위험 경로 실행(읽기 전용이라 해당 없음을 READY 에 적어요), `moai check affected` 를 READY 에서 한 번, 감싸지 않고.
7. SYNC: Result 에 AC 별 met·not_run 을 적고, 카드 기준을 `moai card criterion` 으로 표시하고, 후속 카드를 열 항목(D-4)을 카드 노트로 남겨요. 감사는 이 변경을 쓰지 않은 세션이 맡아요(tier 1; `src/profile.rs` 가 tier-2 바닥에 있으면 전체 감사이고 `moai bundle check` 가 알려요).

## Risks

- 숨김 폴더 안의 진짜 패키지(`.devcontainer/package.json` 등)는 나열되지 않아요. 보고된 피해는 숨김 폴더가 예산을 먼저 써서 실제 패키지를 가리는 쪽이 훨씬 컸어요. 루트 manifest 는 숨김 규칙과 무관하게 그대로 읽고, 확장 파일은 숨김 폴더를 걷는 행을 만들지 않아요.
- 예산은 항목 5000 개 그대로예요. 숨김이 아닌 큰 폴더가 깊이 3 이내에 있으면 `complete` 가 false 로 나와요: 조용히 줄어들던 것이 알려지는 것으로 바뀔 뿐 목록이 늘지는 않아요(REQ-9, REQ-10).
- 걷지 않을 폴더 이름도 카탈로그 데이터라서, 카탈로그에 없는 생태계의 출력 폴더(`.gitignore` 에 없고 점 이름도 아닌 것)는 걷고 그 안의 manifest 를 나열해요. 프로젝트의 `.gitignore` 와 확장 파일의 폴더 행이 그 길이에요. 알려진 한계로 두고 후속으로 적어요.
- `Plan` 에 필드가 하나 더해져요. 이 구조체는 `deny_unknown_fields` 라서 옛 계획 JSON 은 그대로 읽히고(필드가 없으면 빈 목록), 새 계획을 옛 바이너리가 읽으면 거절돼요. 계획은 같은 실행에서 만들고 쓰는 단기 값이라 문제가 없어요. 카드 C-150 이 `project-apply` 의 입력 읽기를 바꾸는데, `dropped_manifests` 는 `Plan` 의 알려진 필드라서 그 변경과 겹치지 않아요. 둘이 같은 날 착지하면 착지하는 쪽이 `project_apply_plan_result` 시험을 한 번 돌려요.
- `SKIP_DIRS` 의 쓰임을 읽기 함수로 바꾸면 코드 파일 걷기(`code_languages`)와 `project_scale_detect` 가 같은 목록을 계속 써요. 값은 같고 출처만 카탈로그예요. 그 두 걷기의 순서와 예산은 바뀌지 않아요.
- 새 상태 폴더 파일(확장)과 새 시험 파일이 `test_area_layout` 과 `state_folder_literal_scan` 의 표에 줄을 더해야 해요. Plan 6 의 `check affected` 가 단계 `test_tables` 에서 먼저 알려요.
- 카탈로그 JSON 에 틀린 행이 있으면 모든 감지가 영향을 받아요: 파싱 실패는 `detect` 를 실패시키지 않고 단위 시험이 배송 파일의 모든 값이 기록 가능한지(`recordable`)와 열두 이름이 모두 있는지를 지켜요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | manifest 인식을 데이터 카탈로그로 옮기고 프로젝트 확장 파일(행을 더하기만 함)을 둔다. 대안: (b) 배송 카탈로그만(확장은 후속), (c) 코드 표 유지 | 사용자 지시 "언어·도구 하드코딩 금지, 모든 개발 언어"; 카탈로그만으로는 새 언어마다 MoAI 를 고쳐야 해요. Renovate 는 사용자 패턴을 기본값에 "더하고" Linguist 는 데이터 파일과 `.gitattributes` 덮어쓰기를 둬요(조사 2026-11-24). 기본: (a) | 확장 읽기를 빼요(카탈로그는 그대로) |
| D-2 | 걷지 않을 폴더 = 점 이름 폴더 전부 + 카탈로그의 폴더 이름 + 프로젝트 무시 목록(지금과 같음). 이름 목록을 코드에 늘리지 않는다 | 보고된 큰 폴더가 모두 숨김 폴더였어요. ripgrep 은 숨김을 기본으로 건너뛰어요. 기본: 이대로 | 점 이름 건너뛰기를 빼요 |
| D-3 | 탐색 고침과 "줄어드는 목록 보여 주기"를 한 SPEC·한 카드에서 한다(옛 두 가지의 설계를 이어받음). 필요한 `src/profile.rs`·`src/project_update.rs` 는 지금 풀려 있어요(C-105 완료, C-226 보류에 예약 없음, 이 실행에서 읽음). 대안: 탐색만 먼저 착지하고 보여 주기는 두 번째 SPEC | 카드 기준 1 이 둘을 한 문장으로 묶고, 둘 다 같은 사고의 두 부분이에요. 기본: 한 SPEC | 둘째 부분을 새 SPEC 으로 분리해요 |
| D-4 | 이 SPEC 은 manifest 인식과 걷기만 데이터로 옮긴다. 코드 파일 확장자 표, 언어 서버 표, 데이터베이스 표, 빌드·테스트 명령 템플릿, 작업공간 표지 표는 후속 카드로 적는다 | 한 SPEC 이 감지 전체를 다시 쓰면 `lsp_plugins`·`commands`·`scale` 을 쓰는 다른 카드와 부딪히고 감사가 어려워져요(규칙 20). 기본: 후속 카드 | 후속 카드를 이 SPEC 에 합쳐요 |
| D-5 | 하위 manifest 의 언어도 `languages` 와 `lsp` 에 반영한다. `SPEC-20261119-nested-manifests` 의 D-3("하위 manifest 는 언어에 반영하지 않는다")을 이 SPEC 이 대체한다. `databases` 는 루트 한정 그대로 | 첫 감사의 리더 결정(2026-11-24T06:15:51Z, 사용자 위임 2026-11-23T13:57Z, 노트 "AUDIT #256"의 N2): AC-4 가 확장 행 언어의 반영을 요구하고 사용자 규칙 "모든 언어 지원"과 맞아요. 결과: 하위 manifest 만 있는 기존 모노레포 프로필은 `/moai update` 에서 `languages`·`lsp` 차이로 `ready` 가 돼요(적용 전에 미리보기에서 보여요). 기본: 반영 | 하위 manifest 를 `languages` 에서 빼요(한 줄) |
| D-6 | `.gitignore` 읽기는 폴더를 큐에서 꺼낼 때 하고 한 걸음의 총량을 1 MiB(`MAX_IGNORE_TOTAL`, 파일 한도의 4 배)로 막는다. 총량을 넘는 파일은 줄이 없는 것으로 세어 더 읽기만 하고 덜 읽지 않는다. 기존 한도 상수(`MAX_DEPTH`, `MAX_ENTRIES`, `MAX_ITEMS`, `MAX_FILE`)는 그대로 | 첫 감사의 N1: 폴더마다 최대 256 KiB 를 큐에 넣을 때 읽어 한 깊이 전체를 메모리에 두면 300 폴더에서 RSS 430 MB 였어요. 읽기를 꺼낼 때로 옮기는 것만으로는 하위 폴더가 있는 폴더의 목록이 자식이 나올 때까지 남아서 총량 한도를 같이 둬요. 기본: 4 배 | 총량 한도를 빼요(`budgeted_ignore_lines` 의 한 조건) |

## Result

기준 팁 NeXT `e9c6b93400`. RED 커밋 `412167526c`(새 시험 13 개 중 11 개가 옳은 이유로 실패하고 줄어들지 않는 경우를 지키는 가드 2 개만 통과), GREEN 커밋 `46d34c30be`.

- AC-1: met — `cargo nextest run --locked --test project_detect_monorepo_cli` 의 `a_large_early_folder_does_not_hide_a_later_packages_manifest` 와 `hidden_folders_are_not_walked_for_manifests` 통과(RED 에서는 manifests 가 `[]`, 숨김 폴더 세 개의 manifest 가 유입).
- AC-2: met — `a_temporary_monorepo_lists_every_sub_manifest` 통과: 루트 `package.json` 과 `apps/a/package.json`, `apps/b/tsconfig.json`, `apps/id/go.mod` 가 나열되고 `tools`(go, node, npm, typescript)와 `languages`(go, javascript, typescript)가 예상대로, 내용은 읽지 않고(깨진 JSON 도 통과) 트리는 그대로.
- AC-3: met — `folders_the_catalog_lists_are_not_walked` 통과(`node_modules`, `target`, `vendor` 는 배송 카탈로그의 이름, `generated-out` 은 확장 파일의 이름, `kept` 는 걷고 나열).
- AC-4: met — `a_row_in_the_extension_file_adds_a_language_without_code` 통과: `mix.exs`(파일 이름 행)와 `App.csproj`(접미사 행)가 확장 행만으로 나열되고 `languages`·`tools` 에 반영, 확장 파일을 지우면 사라져요. 엔진은 새 이름을 알지 못해요.
- AC-5: met — `an_extension_file_adds_rows_and_a_bad_one_is_reported_and_ignored` 통과: 배송 행의 이름을 가진 확장 행은 배송 행을 바꾸지 못하고, 잘못된 JSON·알 수 없는 키·기록할 수 없는 값 둘·한도(256 KiB)를 넘는 파일·링크 다섯은 `manifest_catalog.extension` 이 `refused` 이고 `problem` 이 문자열이며 배송 행만 쓰이고 트리는 그대로.
- AC-6: met — `the_manifest_walk_fact_reports_completeness_and_the_cap`(작은 프로젝트 `{complete: true, over_cap: 0}`, 70 개는 64 개 나열에 `over_cap` 6)와 `a_partial_list_adds_one_next_action_and_a_complete_one_adds_none`(예산 소진 1 줄, 상한 초과 1 줄, 완전한 목록 0 줄) 통과.
- AC-7: met — `cargo nextest run --locked --test project_update_drops_cli` 의 `the_project_plan_names_the_recorded_manifests_a_smaller_detected_list_drops`, `a_list_that_only_grows_or_stays_drops_nothing`, `a_hand_edited_detected_key_drops_nothing` 통과.
- AC-8: met — `a_plan_with_the_dropped_list_removed_is_refused_at_apply` 통과("stale or altered plan", 프로필 파일 그대로). 목록이 비면 digest 가 전과 같다는 것은 빈 목록을 옛 튜플 그대로 직렬화하는 코드(`plan_digest`)와 이웃 시험 통과로 보여요. 옛 digest 값을 박은 새 시험은 만들지 않았어요.
- AC-9: met — `the_update_plan_profile_step_says_how_many_manifests_would_drop` 통과(`reason` 에 "2 recorded manifests would drop", `preview.dropped_manifests` 가 그 둘).
- AC-10: met — 한 번의 `cargo nextest run --locked` 로 `project_detect_cli`, `project_scale_detect_lib`, `profile_apply_engine`, `profile_plan_cli`, `project_update_plan_cli`, `home_isolation_scan`, `path_root_direct_sites`, `state_folder_literal_scan`, `test_area_layout` 과 lib 의 `project_detect`·`profile`·`project_update`·`project_scale` 시험을 돌려 212 개 통과, 실패 0. `cargo fmt --check` 종료 0, `cargo clippy --locked --offline --lib --bins --test project_detect_monorepo_cli --test project_update_drops_cli -- -D warnings` 종료 0(경고 0). 새 lib 단위 시험 셋: `the_shipped_catalog_holds_the_earlier_rows_with_recordable_values`, `extension_rows_are_added_and_never_replace_a_shipped_row`, `a_catalog_file_with_an_unrecordable_value_is_refused`.
- AC-10 보충: 첫 `moai check affected`(head `041ec0b00b`)가 전체 3102 개 중 `green_test_lock every_detected_language_has_test_patterns` 하나만 실패(`test=100`)했어요. 이 시험이 소스의 `const MANIFESTS` 줄을 읽었는데 그 표가 카탈로그 파일로 옮겨졌기 때문이에요. 시험이 카탈로그의 `language` 값을 받도록 한 곳만 고쳤어요(`tests/green_test_lock.rs`, 카드 범위를 이 파일 하나만큼 넓힘). 처음에는 시험이 카탈로그 파일을 `include_str!` 로 읽게 했다가 두 번째 `check affected`(head `09527b9c64`)에서 `repo_scans_table every_include_str_reader_of_a_repository_file_has_an_always_scan` 이 새 읽기를 잡았어요(`test=100`). 공유 스캔 표를 건드리는 대신 `project_detect::shipped_manifest_languages()` 를 두고 시험이 그걸 쓰게 했고, 두 파일과 감지 시험 86 개가 통과해요. 이 고침으로 "감지기가 말하는 모든 언어에 시험 경로 행이 있다"는 검사는 그대로 유지되고, 확장 파일 행의 언어는 프로젝트마다 달라서 이 검사의 대상이 아니에요.
- AC-11: not_run — 보고된 저장소는 이 레인이 읽을 권한이 없는 실제 프로젝트라서 돌리지 않았어요. 55 개 실측은 리더나 그 저장소를 읽을 수 있는 세션이 읽기 전용 `moai project-detect` 로 재요. 이 SPEC 은 그 구조(큰 숨김 폴더가 `apps/` 앞에서 예산을 쓰는 것)를 시험 `a_large_early_folder_does_not_hide_a_later_packages_manifest` 로 재현해요.
- 가드 뮤턴트 13 개를 하나씩 코드에 넣고 새 두 시험 파일을 돌려 모두 컴파일됐고 지정한 시험이 실패했어요(종료 100): 폭 우선→깊이 우선, 점 이름 건너뛰기 빼기, 카탈로그 폴더 건너뛰기 빼기, 예산 소진 `complete` 를 항상 true, `over_cap` 계산 한 칸 어긋남, 부분 목록 조건 `||` 를 `&&` 로, 확장 파일 크기 한도 빼기, 확장 값 검사 빼기, 배송 행 우선 규칙 빼기, 하위 manifest 언어 반영 빼기, digest 에서 목록 빼기, 줄어드는 목록의 적용 키 조건 빼기, 업데이트 단계 reason 의 목록 빼기. 변이한 파일은 모두 원래 바이트로 되돌렸고 트리는 깨끗해요.
- 위험 경로: 읽기 전용(삭제·리셋·프로세스 시작·권한 규칙 없음)이라 해당 없음. 확장 파일 읽기는 링크를 따르지 않고 크기를 제한하며 시험이 링크·초과 파일을 지켜요.

리더 결정(노트 8178)과 SPEC 본문의 차이, 그리고 확정한 세부:

- D-4 는 노트 8178 로 고쳐졌어요: 나머지 하드코딩 표(코드 확장자, 언어 서버, DB 의존성, 빌드·테스트 명령 템플릿, 작업공간 표지)는 새 카드가 아니라 카드 C-256 의 기준 3 ~ 7 이고 READY 는 `keep_open` 이에요. 위 Non-goals 와 Plan 7 의 "후속 카드" 는 이 뜻으로 읽어요(본문은 승인된 digest 를 지키려고 바꾸지 않았어요).
- 확장 파일의 자리: 상태 폴더의 `config/detect-catalog.json`(`config/` 는 팀이 git 에 두는 폴더), `schema` 는 1 이에요.
- 하위 manifest 도 `languages`(와 그에 따른 `lsp`)에 반영해요. 카드 C-61 의 D-3 은 하위 manifest 를 반영하지 않기로 했는데, 이 SPEC 의 AC-4 가 확장 행의 언어가 `languages` 에 나타나야 한다고 해서 따랐어요. `databases` 는 루트 한정 그대로예요.
- `manifest_catalog` 는 `extension`(absent, read, refused), `rows`, `problem` 을 내요. `SKIP_DIRS` 상수는 사라지고 `project_detect::skip_dir` 가 호스트·상태 폴더 이름과 배송 카탈로그의 이름을 판정해요(`project_scale_detect` 의 사용 줄만 바뀜).
- 루트의 manifest 는 이름으로 직접 확인해서 예산이나 무시 목록과 무관하게 나열하고, 걷기는 루트의 접미사 행도 찾아요.

돌리지 않은 것: 전체 `cargo test`, 전체 크레이트 `clippy --all-targets`, Windows·Linux, AC-11 의 실제 저장소 실측, `moai check affected`(READY 에서 한 번).

### 판 2 수리 (첫 감사 FAIL, 머리 `8b8572385c`)

감사 노트 "AUDIT #256" 의 막는 것 B1·B2 와 비차단 N1·N4 를 같은 가지에서 고쳤어요. RED 커밋 `85411f8c19`(시험 여섯 개를 더함: 셋은 옛 머리에서 실패, 경계 시험 셋은 이미 맞는 동작을 지켜서 통과), GREEN 커밋 `302a78b266`.

- B1 met: 한도의 양쪽 경계 시험이 생겼어요. `the_extension_file_limit_holds_at_both_edges`(확장 파일 262144 바이트는 읽고 262145 는 거절), `the_extension_row_and_list_limits_hold_at_both_edges`(행·건너뛸 폴더·행 안 파일·접미사·도구 64 개는 읽고 65 개는 거절), `the_ignore_file_limit_holds_at_both_edges`(루트와 하위 폴더의 `.gitignore` 262144 는 적용, 262145 는 아님). 동작은 그대로예요.
- B2 met: 읽을 수 없는 폴더·항목이 있으면 `manifest_walk.complete` 가 false 이고 부분 목록 줄이 하나예요(`a_folder_that_cannot_be_read_makes_the_walk_incomplete`, 권한 000 폴더; 읽을 수 있는 폴더의 manifest 는 그대로 나열, 권한을 돌려 놓으면 complete). REQ-7 문장을 이 뜻으로 넓혔어요.
- N1 met: `.gitignore` 를 큐에 넣을 때가 아니라 꺼낼 때 읽고, 한 걸음의 총량을 1 MiB(`MAX_IGNORE_TOTAL`)로 막아요(`the_ignore_files_the_walk_reads_have_a_total_size_cap`: 약 200 KB 목록 일곱 중 다섯만 적용). 총량을 넘는 목록은 줄이 없는 것으로 세어 더 읽기만 해요. 메모리 실측(RSS)은 이 레인이 다시 재지 않았어요(not_run); 감사가 300 폴더×256 KiB 에서 430 MB 를 쟀던 입력이 이제 최대 1 MiB 의 목록 텍스트만 쥐어요(코드 읽기, 시험으로 총량을 지킴).
- N4 met: 점만으로 된 이름(`.`, `..`)은 확장 파일의 어느 칸에서도 거절돼요(`a_name_of_only_dots_is_refused_in_the_extension_file`; `recordable` 한 줄).
- N2 결정 기록: 하위 manifest 언어의 반영을 Decisions D-5 로 적었어요(리더 결정 2026-11-24T06:15:51Z, `SPEC-20261119-nested-manifests` D-3 대체). 이 반영은 하위 manifest 만 있는 기존 모노레포의 프로필을 `/moai update` 에서 `languages`·`lsp` 차이로 `ready` 로 만들어요.
- N3(상한을 넘거나 기록할 수 없는 경로의 manifest 언어도 셈), N5(오류 문구 둘), N7(HOME-GUARD 한 번 흔들림)은 이 수리에서 건드리지 않았고 카드의 열린 기준 2~7 조각에서 봐요(리더 결정).

가드 뮤턴트 아홉 개를 코드에 하나씩 넣고 시험 파일 하나(`project_detect_monorepo_cli`, 열네 개 중 지정한 시험)로 돌려 모두 컴파일됐고 지정한 시험만 실패했어요(종료 100, 결과 `target/project-detect-monorepo/mutants-repair.json`, 소스는 원래 바이트로 복원): 확장 파일 크기 `<=`→`<`, 행·폴더 수 한도 조건을 `false` 로, 행 안 목록 한도 통째로 빼기, 행 안 파일 한도 `<=`→`<`, 루트 `.gitignore` 크기 한도 `>`→`>=`, 하위 `.gitignore` 크기 한도 `>`→`>=`, 총량 한도 빼기, 읽을 수 없는 폴더에서도 `complete` 를 true 로, 점만 있는 이름 검사 빼기. 읽을 수 없는 항목(`entries` 오류, `file_type` 오류)의 표시는 같은 변수라 위 뮤턴트가 지키고, 그 두 갈래만 따로 시험하지는 않아요(not_run: 만들기 어려운 입력).
