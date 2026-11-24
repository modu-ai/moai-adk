---
id: SPEC-20261125-detect-code-extension-catalog
title: "project-detect reads the code file extensions and their languages from the catalog, not from a table in the engine"
status: draft
revision: 1
created: 2026-11-25
parent: SPEC-20261124-project-detect-monorepo
---

# project-detect reads the code file extensions and their languages from the catalog, not from a table in the engine

## Intent

사용자 지시(2026-11-24): MoAI 는 모든 개발 언어를 지원해야 하고 언어·도구 하드코딩은 금지예요. 부모 SPEC `SPEC-20261124-project-detect-monorepo` 가 manifest 인식을 데이터 카탈로그와 프로젝트 확장 파일로 옮겼고, 부모 카드의 남은 기준 하나가 코드 파일의 언어 감지예요: `src/project_detect.rs` 의 `EXTENSIONS` 상수 표(확장자 열일곱 개 → 언어 열네 개)가 아직 엔진 안에 있어서, 표에 없는 언어(예: `.ex` Elixir, `.zig` Zig)는 코드 파일만 있는 프로젝트에서 `detected.languages` 에 나타나지 않고 새 언어마다 MoAI 를 고쳐야 해요.

이 SPEC 은 그 표 하나만 같은 카탈로그 파일의 `code_extensions` 행으로 옮겨요. 읽기·병합·검증 장치(배송 행 우선, 한도 64, 읽을 수 없으면 확장 파일 통째 거절)는 부모가 만든 것을 그대로 써요. 읽는 사람: 레인, 이 변경의 감사자, `/moai project` 의 사용자.

## Non-goals

- 언어 서버 표(`LSP_PLUGINS`), 데이터베이스 의존성 표(`DB_DEPENDENCIES`), `commands_fact` 의 명령 템플릿, `project_scale_detect` 의 작업공간 표지는 이 SPEC 이 아니에요. 부모 카드의 다른 기준이 한 조각씩 맡아요.
- 한도 상수(`MAX_DEPTH`, `MAX_ENTRIES`, `MAX_ITEMS`, `MAX_FILE`)와 걷기 순서·예산을 바꾸지 않아요. 코드 파일 걷기는 그대로이고 확장자 판정의 출처만 바뀌어요.
- 확장자 비교 방식을 바꾸지 않아요: 마지막 점 뒤를 소문자 그대로 정확히 비교해요(`X.RS` 는 언어를 더하지 않아요).
- 새 언어의 언어 서버·명령은 정하지 않아요. `code_extensions` 로 더한 언어는 `languages` 에만 나타나고 `lsp` 는 `LSP_PLUGINS` 에 행이 있을 때만 나와요(다음 조각의 몫).
- 부모 SPEC 의 본문과 승인된 digest 는 건드리지 않아요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: The shipped catalog shall hold the code file extensions and the language of each that the earlier `EXTENSIONS` table held, so that a code file with one of them adds the same language as before. · failure: none — a unit test pins the pairs and a CLI test runs one file per extension.
- REQ-2: When the extension file holds `code_extensions` rows, the command shall add the language of a row to `detected.languages` for each code file that carries one of the row's extensions, the shipped row prevailing for an extension that both hold. · failure: none — the extension file is optional.
- REQ-3 (REQ-2): IF a `code_extensions` row names no extension, holds a language or an extension that cannot be recorded, an extension with a dot in it, a key the catalog does not know, or more than 64 rows or more than 64 extensions in one row, THEN the command shall refuse the whole extension file, use the shipped rows alone and report the problem in `manifest_catalog`.
- REQ-4: When the extension file is read, the command shall count its `code_extensions` rows with its `manifests` rows in `manifest_catalog.rows`. · failure: none — the number is a report, not a decision.

## Acceptance

- AC-1 (REQ-1): 이전 표의 열일곱 확장자(`rs py js jsx ts tsx go java kt rb php c h cpp cs swift lua`)마다 파일 하나씩 둔 프로젝트는 열네 언어(`c cpp csharp go java javascript kotlin lua php python ruby rust swift typescript`)와 그 언어 서버를 이전과 같이 내요. 확장자는 소문자 그대로 비교해서 `X.RS` 는 언어를 더하지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli every_shipped_code_extension_keeps_its_language` and `cargo test --locked --offline --lib the_shipped_catalog_holds_the_earlier_code_extensions`
- AC-2 (REQ-2, REQ-4): 카탈로그에 없는 언어(`lib/a.ex`)는 확장 파일의 `code_extensions` 행 하나로 `languages` 에 나타나고 파일을 지우면 사라져요. `manifest_catalog.rows` 는 `manifests` 행과 `code_extensions` 행을 함께 세요. 확장 파일을 읽어도 트리는 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_code_extension_row_adds_a_language_without_code`
- AC-3 (REQ-2): 배송 확장자(`rs`)를 다른 언어로 올리려는 행은 그 확장자를 더하지 못하고 같은 행의 새 확장자만 쓰여요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_code_extension_row_never_replaces_a_shipped_extension`
- AC-4 (REQ-3): 읽을 수 없는 `code_extensions` 행(확장자 없음, 언어나 확장자를 기록할 수 없음, 확장자 안의 점, 모르는 키, 목록이 아닌 값)이 하나라도 있으면 파일 전체가 `manifest_catalog.extension` `refused` 이고 `problem` 이 문자열이며, 같은 파일의 좋은 `manifests` 행도 쓰이지 않고 배송 행만 쓰이며 트리가 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_code_extension_row_that_cannot_be_read_refuses_the_whole_file`
- AC-5 (REQ-3): 행 64 개와 한 행의 확장자 64 개는 읽히고 65 개는 거절돼요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli the_code_extension_row_limits_hold_at_both_edges`
- AC-6: 이웃 시험이 그대로 통과해요 · verify: `cargo test --locked --offline --test project_detect_cli --test project_detect_monorepo_cli --test project_scale_detect_lib --test home_isolation_scan` and `cargo test --locked --offline --lib project_detect`

## Scope

- write: `.moai/specs/SPEC-20261125-detect-code-extension-catalog/**`, `src/project_detect.rs`(`EXTENSIONS` 표 삭제, `CatalogFile` 의 새 키, `code_languages`, 확장 병합·검증, 단위 시험), `src/project_detect_catalog.json`(`code_extensions` 행), `tests/project_detect_monorepo_cli.rs`(시험 더하기), `tests/green_test_lock.rs`(`every_detected_language_has_test_patterns` 가 지워진 `EXTENSIONS` 를 글로 읽던 줄을 카탈로그 함수로 바꿔요; 첫 영수증이 알려 줬어요)
- read-only: `src/project_scale_detect.rs`, `src/ops/project_detect.rs`(카탈로그 형식을 설명하는 글이 없어서 바꿀 곳이 없어요), `tests/project_detect_cli.rs`, `skills/moai/workflows/project.md`
- forbidden: `LSP_PLUGINS`·`DB_DEPENDENCIES`·`commands_fact` 의 표와 본문(다음 조각), `src/profile.rs`, `src/project_update.rs`, `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`, `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME, 한도 상수, 부모 SPEC 의 파일

## Plan

1. SPEC 승인을 기다려요. 승인 전에는 코드를 쓰지 않아요(RED 시험은 승인과 무관하게 먼저 써도 되지만 커밋은 SPEC 초안 커밋 뒤예요).
2. RED: `tests/project_detect_monorepo_cli.rs` 에 AC-1 ~ AC-5 의 시험을 먼저 써서 현재 팁에서 AC-2 ~ AC-5 가 옳은 이유(확장 파일의 모르는 키 `code_extensions` 가 거절됨)로 실패하는 것을 봐요. AC-1 의 CLI 시험은 이미 맞는 동작을 지켜서 통과해요.
3. GREEN: `CatalogFile` 에 `code_extensions` 를, `Catalog` 에 확장자 행을, `catalog_file_valid` 에 행 검사를 더하고 `code_languages` 가 카탈로그를 받게 하고 `EXTENSIONS` 를 지워요. 배송 파일에 열네 행을 적어요. 확장 병합은 배송 행의 확장자를 가진 확장자를 떨구고 남은 것이 없는 행을 버려요. `ExtensionReport.rows` 가 두 종류를 함께 세요.
4. 확인: 새 시험, AC-6 의 이웃 시험, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`(대상 한정), 새 가드 시험마다 가드 뮤턴트 한 줄(`guard-mutant: <test> | <지운 코드> | exit <n>`), 위험 경로는 읽기 전용이라 해당 없음을 READY 에 적어요, `moai check affected` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적고, 부모 카드의 기준 3 을 `moai card criterion` 으로 표시해요. 감사는 이 변경을 쓰지 않은 세션이 맡아요(tier 1).

## Risks

- 배송 `code_extensions` 가 읽히지 않으면 코드 파일로 얻던 언어가 조용히 사라져요. 단위 시험이 열일곱 쌍과 모든 값의 기록 가능성(`recordable`)과 "언어 서버 표의 모든 언어에 확장자 행이 있음"을 지키고, 배송 파일이 읽히지 않으면 `shipped()` 가 비는 동작은 부모 SPEC 과 같아요.
- 프로젝트 확장 파일에 읽을 수 없는 행이 하나라도 있으면 파일 전체를 거절해서 좋은 행도 못 써요. 추측으로 일부만 받는 것보다 안전한 쪽이에요: 거절은 `manifest_catalog.problem` 으로 보여서 사용자가 행을 고쳐요(REQ-3).
- 같은 확장자를 여러 언어가 쓰는 경우(`.h` 는 C 와 C++ 모두)는 한 확장자에 한 언어예요. 배송 행이 먼저이고, 프로젝트가 다른 언어로 올릴 수는 없어요(AC-3). 알려진 한계로 두어요.
- `manifest_catalog.rows` 의 뜻이 "manifest 행 수"에서 "행 수 전체"로 넓어져요. `src` 와 `skills` 에서 `manifest_catalog` 를 만드는 곳은 `src/project_detect.rs` 한 줄이고 읽는 곳은 없어요(검색으로 읽음); 이 숫자를 보는 것은 시험뿐이에요.
- 표를 소스 글로 읽는 이웃 시험이 있었어요: `tests/green_test_lock.rs` 가 `const EXTENSIONS` 를 찾아 줄마다 읽어서 표를 지우자 첫 영수증이 실패했어요. 같은 시험을 `shipped_code_languages()` 호출로 바꿔요(부모의 manifest 조각이 `shipped_manifest_languages()` 로 한 것과 같아요). `tests/skill_references_test.py` 도 `EXTENSIONS` 를 글로 읽지만 부모 조각이 `MANIFESTS` 를 지운 때부터 이미 같은 방식으로 깨져 있고 Python 시험이라 이 저장소의 근거로 쓰지 않아요; 건드리지 않고 Result 에 적어요.
- 새 키가 든 확장 파일을 이전 판 바이너리가 읽으면 모르는 키라서 거절돼요(`deny_unknown_fields`). 팀이 판을 섞어 쓰면 이전 판 쪽은 확장 파일을 못 읽고 배송 행만 써요. 이 SPEC 의 판으로 올리면 풀려요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 코드 확장자 표를 새 파일이 아니라 같은 카탈로그 파일의 `code_extensions` 행(`language`, `extensions`)으로 옮긴다. 대안: (b) 확장자를 manifest 행에 붙임(manifest 가 없는 언어는 표현 못 함), (c) 별도 파일(읽기·검증 장치를 둘로 늘림) | 카탈로그 읽기·병합·검증·거절·보고가 이미 한 파일을 기준으로 있어요; 한 파일이면 사용자가 배울 것이 하나 늘지 않아요. 기본: 이대로 | 키와 행을 지우고 표를 되돌려요 |
| D-2 | 읽을 수 없는 행이 하나라도 있으면 확장 파일을 통째로 거절한다(좋은 행도 쓰지 않음). 대안: 나쁜 행만 건너뜀 | 부모 SPEC 의 REQ-6 과 같은 규칙이에요. 읽을 수 없는 데이터 행을 추측하지 않고, 일부만 받으면 사용자가 보는 목록과 적용되는 목록이 달라져요. 기본: 통째 거절 | 거절 대신 건너뜀으로 바꿔요 |
| D-3 | 배송 확장자는 프로젝트 행이 바꾸지 못한다. 같은 행의 새 확장자만 쓴다 | 부모 SPEC 의 REQ-5 와 같은 규칙이에요(Renovate 는 사용자 패턴을 기본값에 더해요, 2026-11-24 조사). 기본: 이대로 | 프로젝트 행이 이기게 해요 |
| D-4 | 이 조각은 부모 SPEC 의 개정 4 가 아니라 `parent` 로 이은 자식 SPEC 이다 | 부모는 이미 열두 요구사항이라 더하면 `spec check` 가 "12 개 넘음" 경고를 내고(읽음), 개정하면 승인된 digest 가 바뀌어 새 승인이 필요해요. 부모 카드는 열린 채로 기준 3 만 이 SPEC 의 시험으로 닫혀요. 기본: 자식 SPEC | 부모에 합쳐요 |

## Result

기준 팁 NeXT `0babbe3566`. SPEC 초안 커밋 `3940dc570c`, RED 커밋 `045184f564`, GREEN 커밋 `ff06b823d8`.

- Run 1 (RED, `cargo nextest run --locked --offline --no-fail-fast --test project_detect_monorepo_cli --lib -E 'binary(project_detect_monorepo_cli) or test(the_shipped_catalog_holds_the_earlier_code_extensions)'`): 23 개 중 18 통과, 5 실패. 실패는 옳은 이유예요 — 확장 파일이 모르는 키 `code_extensions` 때문에 `the file is not a valid catalog` 로 거절돼요(셋), 한도 시험의 첫 단정(64 행이 `read` 여야 함)이 `refused` 예요, 배송 파일에 `code_extensions` 키가 없어요. 이전 쌍을 지키는 `every_shipped_code_extension_keeps_its_language` 는 현재 동작을 지켜서 통과해요.
- Run 2 (GREEN, `cargo nextest run --locked --offline --no-fail-fast --test project_detect_monorepo_cli --test project_detect_cli --lib -E 'binary(project_detect_monorepo_cli) or binary(project_detect_cli) or test(project_detect)'`): 60 개 모두 통과.
- AC-1: met — `every_shipped_code_extension_keeps_its_language`(열일곱 확장자를 하나씩 따로, 모두 함께 열네 언어와 열두 언어 서버, `A.RS` 는 언어 없음)와 단위 시험 `the_shipped_catalog_holds_the_earlier_code_extensions`(배송 파일의 열일곱 쌍이 정확히 이전 표와 같고 값이 기록 가능하고 확장자에 점이 없음) 통과(Run 2).
- AC-2: met — `a_code_extension_row_adds_a_language_without_code` 통과: `lib/a.ex`·`lib/b.exs` 가 행 하나로 `elixir`, `rows` 1, 두 종류의 행이 있으면 `rows` 3, 읽기가 트리를 바꾸지 않고, 파일을 지우면 언어가 사라져요(Run 2; RED 에서는 `refused`).
- AC-3: met — `a_code_extension_row_never_replaces_a_shipped_extension` 통과: `rs` 를 `zig` 로 올리는 행은 `rs` 를 바꾸지 못하고 같은 행의 `zig` 만 쓰이며, `rs` 만 가진 행은 아무것도 더하지 못해요(Run 2).
- AC-4: met — `a_code_extension_row_that_cannot_be_read_refuses_the_whole_file` 통과: 대조(고친 행)는 `read`·`rows` 2 이고, 열 가지 읽을 수 없는 행(빈 목록, `extensions` 키 없음, 빈 언어, 공백 든 언어, 공백 든 확장자, 앞 점, 가운데 점, 모르는 키, 목록 아닌 값, 숫자 언어)은 모두 `refused`·`rows` 0·`problem` 문자열이고 좋은 `manifests` 행도 쓰이지 않고 배송 언어만 나오고 트리가 그대로예요(Run 2).
- AC-5: met — `the_code_extension_row_limits_hold_at_both_edges` 통과: 행 64 개와 한 행의 확장자 64 개는 `read`, 65 개는 `refused`(Run 2).
- AC-6: READY 노트의 영수증이 정해요(`moai check affected` 가 이웃 시험을 고르고 돌려요; 영수증은 이 Result 를 담은 머리에서 돌려서 이 줄을 나중에 고치지 않아요). 이 Result 를 쓰는 때는 Run 2 의 `project_detect_cli` 와 `--lib project_detect` 만 통과를 읽었고 `project_scale_detect_lib`, `home_isolation_scan` 은 아직 `not_run` 이에요.
- 가드 뮤턴트 열 개를 한 슬롯 호출에서 하나씩 넣고 같은 필터로 돌려 열 개 모두 컴파일되고 지정한 시험이 실패했어요(종료 100, 편집은 정확히 한 번씩, 소스는 원래 바이트로 복원, 이후 `git status --short` 비어 있음; `target/c256ext/mutants.out`): `project_code_rows_ignored`, `project_extension_replaces_shipped`(병합의 떨구기와 조회 순서를 함께 바꿈 — 떨구기 한 가지만으로는 조회가 배송 행을 먼저 찾아서 관찰되지 않아요), `dot_in_extension_accepted`, `row_without_extensions_accepted`, `code_rows_not_validated`, `row_limit_65_accepted`, `extension_limit_65_accepted`, `rows_count_manifests_only`, `extension_compared_lowercase`, `shipped_extension_missing`.
- 첫 영수증(`moai check affected 0babbe3566..d574f5ba9a`, 93 targets)은 FAIL 이었어요(`test=100`, 3258 개 중 1 개 실패): `green_test_lock` 의 `every_detected_language_has_test_patterns` 가 지워진 `const EXTENSIONS` 를 글로 찾다가 `None` 에서 멈췄어요. 같은 머리에 대한 재실행이 아니에요 — 시험을 카탈로그 함수 `shipped_code_languages()` 로 바꾼 다른 머리의 새 영수증이에요. Run 3 (`cargo nextest run --locked --offline --no-fail-fast --test green_test_lock --test project_detect_monorepo_cli --lib -E 'binary(green_test_lock) or binary(project_detect_monorepo_cli) or test(project_detect)'`): 46 개 모두 통과.
- 보고만 하는 것: `tests/skill_references_test.py` 의 `detector_languages()` 는 `MANIFESTS` 와 `EXTENSIONS` 표를 소스 글로 찾는데 둘 다 카탈로그로 옮겨져서 더는 맞지 않아요(앞 표는 부모 조각이 옮겼어요). Python 시험이고 이 SPEC 의 범위 밖이라 돌리지도 고치지도 않았어요; 리더가 고칠 곳을 정해 주세요.
- 위험 경로: 읽기 전용 감지라서 지우기·reset·프로세스 시작·권한 규칙이 없어요. 시험은 빌드된 바이너리를 임시 폴더와 임시 HOME 에서 돌려요.
- 돌리지 않은 것: 전체 `cargo test`, Windows·Linux, 부모의 AC-11 실측.
