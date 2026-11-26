---
id: SPEC-20261125-detect-database-dependency-catalog
title: "project-detect reads the database dependency names and their kinds from the catalog, not from a table in the engine"
status: draft
revision: 1
created: 2026-11-25
parent: SPEC-20261124-project-detect-monorepo
---

# project-detect reads the database dependency names and their kinds from the catalog, not from a table in the engine

## Intent

사용자 지시(2026-11-24): MoAI 는 모든 개발 언어와 도구를 지원해야 하고 언어·도구 하드코딩은 금지예요. 부모 SPEC `SPEC-20261124-project-detect-monorepo` 가 manifest 인식을, 형제 SPEC `SPEC-20261125-detect-code-extension-catalog` 가 코드 확장자 표를 데이터 카탈로그와 프로젝트 확장 파일로 옮겼어요. 부모 카드의 남은 기준 하나가 데이터베이스 의존성 표예요: `src/project_detect.rs` 의 `DB_DEPENDENCIES` 상수 표(의존성 이름 스물두 개 → 데이터베이스 종류 다섯 개)가 아직 엔진 안에 있어서, 표에 없는 의존성(예: `sqlx`, `npgsql`, `diesel`)은 `detected.databases` 에 나타나지 않고 새 의존성마다 MoAI 를 고쳐야 해요.

이 SPEC 은 그 표 하나만 같은 카탈로그 파일의 `database_dependencies` 행으로 옮겨요. 읽기·병합·검증·거절·보고 장치(배송 행 우선, 목록마다 한도 64, 읽을 수 없으면 확장 파일 통째 거절)는 부모와 형제가 만든 것을 그대로 써요. 읽는 사람: 레인, 이 변경의 감사자, `/moai project` 의 사용자.

## Non-goals

- 언어 서버 표(`LSP_PLUGINS`), `commands_fact` 의 명령 템플릿, `project_scale_detect` 의 작업공간 표지는 이 SPEC 이 아니에요. 부모 카드의 다른 기준이 한 조각씩 맡아요.
- `prisma_database`(`prisma/schema.prisma` 의 provider → 종류 대응)는 이 SPEC 이 아니에요. 의존성 이름 표가 아니라 파일 내용을 읽는 별도 경로이고, 기준 5 의 글은 `DB_DEPENDENCIES` 표만 말해요. 후속으로 둬요(Risks).
- 의존성 이름을 모으는 방식(`dependency_names`: Cargo.toml, package.json, Python, go.mod 를 읽고 소문자로 바꿈)과 한도 상수(`MAX_DEPTH`, `MAX_ENTRIES`, `MAX_ITEMS`, `MAX_FILE`)는 바꾸지 않아요. 판정의 출처만 바뀌어요.
- 의존성 이름의 정규화(예: `psycopg2_binary` 와 `psycopg2-binary`)는 하지 않아요. 지금처럼 모은 이름을 소문자로만 비교해요.
- 부모 SPEC 과 형제 SPEC 의 본문과 승인된 digest 는 건드리지 않아요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: The shipped catalog shall hold the database dependency names and the kind of each that the earlier `DB_DEPENDENCIES` table held, so that a project declaring one of them reports the same kind as before. · failure: none — a unit test pins the pairs and a CLI test runs one project per dependency.
- REQ-2: When the extension file holds `database_dependencies` rows, the command shall add the kind of a row to `detected.databases` for each dependency the project declares that the row names, the shipped row prevailing for a dependency that both hold. · failure: none — the extension file is optional.
- REQ-3 (REQ-2): IF a `database_dependencies` row names no dependency, has no kind or a kind that cannot be recorded, holds a dependency name outside lower-case letters, digits and `- _ . @ /`, without a letter or digit, or longer than 120 bytes, a key the catalog does not know, or more than 64 rows or more than 64 dependencies in one row, THEN the command shall refuse the whole extension file, use the shipped rows alone and report the problem in `manifest_catalog`.
- REQ-4: When the extension file is read, the command shall count its `database_dependencies` rows with its `manifests` and `code_extensions` rows in `manifest_catalog.rows`. · failure: none — the number is a report, not a decision.
- REQ-5 (REQ-2): IF two `database_dependencies` rows of the extension file give one dependency two different kinds, THEN the command shall refuse the whole extension file, use the shipped rows alone and report a problem that names the dependency; the same kind for one dependency more than once is not a conflict.

## Acceptance

- AC-1 (REQ-1): 이전 표의 스물두 의존성(`postgres tokio-postgres pg psycopg psycopg2 psycopg2-binary asyncpg github.com/jackc/pgx/v5 github.com/lib/pq mysql mysql2 pymysql mysqlclient github.com/go-sql-driver/mysql rusqlite sqlite3 better-sqlite3 mongodb pymongo mongoose redis ioredis`)마다 그것 하나만 선언한 프로젝트는 이전과 같은 종류(`postgres mysql sqlite mongodb redis`)를 내요. 배송 파일이 읽히지 않아 카탈로그가 비는 일이 없도록, 단위 테스트는 배송 manifest 행과 코드 확장자 행과 의존성 행이 모두 비어 있지 않은 것도 지켜요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli every_shipped_database_dependency_keeps_its_kind` and `cargo test --locked --offline --lib the_shipped_catalog_holds_the_earlier_database_dependencies`
- AC-2 (REQ-2, REQ-4): 카탈로그에 없는 의존성(`sqlx`)은 확장 파일의 `database_dependencies` 행 하나로 `databases` 에 나타나고 파일을 지우면 사라져요. `manifest_catalog.rows` 는 `manifests`, `code_extensions`, `database_dependencies` 행을 함께 세요. 확장 파일을 읽어도 트리는 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_database_dependency_row_adds_a_kind_without_code`
- AC-3 (REQ-2): 배송 의존성(`pg`)을 다른 종류로 올리려는 행은 그 의존성을 바꾸지 못하고(`pg` 만 선언한 프로젝트는 `postgres` 만 내요) 같은 행의 새 의존성만 쓰여요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_database_dependency_row_never_replaces_a_shipped_dependency`
- AC-4 (REQ-3): 읽을 수 없는 `database_dependencies` 행(의존성 없음, `kind` 없음, 기록할 수 없는 종류(빈 값, 공백, `/`), 빈 이름, 글자나 숫자가 없는 이름(`.`), 공백이나 `+` 가 든 이름, 대문자가 든 이름, 121 바이트 이름, 모르는 키, 목록이 아닌 값)이 하나라도 있으면 파일 전체가 `manifest_catalog.extension` `refused` 이고 `problem` 이 문자열이며, 같은 파일의 좋은 `manifests` 와 `code_extensions` 행도 쓰이지 않고 배송 행만 쓰이며 트리가 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file`
- AC-5 (REQ-3): 행 64 개와 한 행의 의존성 64 개와 120 바이트 이름은 읽히고, 65 개와 121 바이트 이름은 거절돼요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli the_database_dependency_row_limits_hold_at_both_edges`
- AC-6 (REQ-5): `sqlx` 를 `postgres` 와 `mysql` 두 종류로 올리는 두 행은 어느 쪽이 앞이든 파일 전체가 `refused` 이고 `rows` 0 이며 `problem` 이 `the dependency sqlx is claimed by two kinds: <앞 종류> and <뒤 종류>` 이고 배송 종류만 나오며 트리가 바뀌지 않아요. 같은 종류로 두 번(한 행 안에서든 두 행에서든)은 `read` 예요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli one_dependency_for_two_kinds_in_the_extension_file_is_refused`
- AC-7: 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test project_detect_cli --test project_detect_monorepo_cli --test project_scale_detect_lib --test green_test_lock --test home_isolation_scan` and `cargo test --locked --offline --lib project_detect`

## Scope

- write: `.moai/specs/SPEC-20261125-detect-database-dependency-catalog/**`, `src/project_detect.rs`(`DB_DEPENDENCIES` 표 삭제, `CatalogFile` 의 새 키, 의존성 이름 규칙, 확장 병합·검증, 단위 테스트), `src/project_detect_catalog.json`(`database_dependencies` 행), `tests/project_detect_monorepo_cli.rs`(테스트 더하기)
- read-only: `src/project_scale_detect.rs`, `src/ops/project_detect.rs`, `tests/project_detect_cli.rs`, `tests/green_test_lock.rs`, `skills/moai/workflows/project.md`
- forbidden: `LSP_PLUGINS`·`commands_fact`·`prisma_database` 의 표와 본문, `dependency_names` 와 `python_dependency_names` 의 본문, `src/profile.rs`, `src/project_update.rs`, `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`, `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME, 한도 상수, 부모와 형제 SPEC 의 파일

## Plan

1. SPEC 초안을 쓰고 커밋해요. 승인은 감사 때 리더가 노트로 해요; 레인은 승인을 기록하지 않고 초안 위에서 RED 부터 구현해요.
2. RED: `tests/project_detect_monorepo_cli.rs` 에 AC-1 ~ AC-6 의 테스트를 먼저 써서 현재 팁에서 AC-2 ~ AC-6 이 옳은 이유(확장 파일의 모르는 키 `database_dependencies` 가 거절됨)로 실패하는 것을 봐요. AC-1 의 CLI 테스트는 이미 맞는 동작을 지켜서 통과해요.
3. GREEN: `CatalogFile` 에 `database_dependencies` 를, `Catalog` 에 의존성 행을 더하고, `catalog_file_valid` 에 행 검사와 한 파일 안의 한 의존성 두 종류 거절을 더하고, 의존성 이름 규칙(`dependency_name`)을 따로 둬요(이름에 `/` 와 `.` 가 있어서 `recordable` 을 쓸 수 없어요). `detect` 의 `databases` 계산이 카탈로그를 읽게 하고 `DB_DEPENDENCIES` 를 지워요. 배송 파일에 다섯 행을 적어요. 확장 병합은 프로젝트 행을 배송 행 뒤에 덧붙이기만 해요: 한 의존성의 종류는 먼저 찾은 행이 정해요. `ExtensionReport.rows` 가 세 종류를 함께 세요. 단위 테스트 `tables_name_only_recordable_values` 는 카탈로그 행을 읽게 바꿔요.
4. 확인: 새 테스트, AC-7 의 이웃 테스트, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`(대상 한정), 새 가드 테스트마다 가드 뮤턴트 한 줄(`guard-mutant: <test> | <지운 코드> | exit <n>`), 위험 경로는 읽기 전용이라 해당 없음을 READY 에 적어요, `moai check affected --card` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적고, 카드의 기준 5 와 3 을 `moai card criterion` 으로 표시해요(커밋 뒤에 레인이 해요). 감사는 이 변경을 쓰지 않은 세션이 맡아요(tier 1).

## Risks

- 배송 `database_dependencies` 가 읽히지 않으면 `detected.databases` 가 조용히 비어요. 더 나쁜 것은 검증이 배송 파일 자체를 거절하면 `shipped()` 가 통째로 비어서(`unwrap_or_default`) manifest 행과 코드 확장자 행까지 함께 사라지는 거예요. 의존성 이름에는 `/` 와 `.` 가 있어서(`github.com/jackc/pgx/v5`) `recordable` 은 이름을 거절해요. 그래서 이름은 따로 규칙(`dependency_name`)을 두고, 단위 테스트가 스물두 쌍과 세 종류의 배송 행이 비어 있지 않음을 지켜요.
- 프로젝트 확장 파일에 읽을 수 없는 행이 하나라도 있으면 파일 전체를 거절해서 좋은 행도 못 써요. 추측으로 일부만 받는 것보다 안전한 쪽이에요: 거절은 `manifest_catalog.problem` 으로 보여서 사용자가 행을 고쳐요(REQ-3).
- `prisma_database` 는 의존성 이름 표 밖에서 provider 이름을 종류로 바꾸는 하드코딩이 남아 있어요(`postgresql`·`cockroachdb` → `postgres`, `sqlserver` 등). 기준 5 의 글에는 없어서 이 SPEC 이 옮기지 않아요. 후속으로 정할 때 같은 `database_dependencies` 행으로 표현할 수 있는지 먼저 봐요.
- 배송 표에 `sqlserver` 의존성 행이 없는 것은 이전 표와 같아요. 이 SPEC 은 표의 내용을 늘리지 않고 옮기기만 해요(REQ-1).
- `manifest_catalog.rows` 의 뜻이 세 종류의 행 수 합으로 넓어져요. 한 파일은 목록마다 64 개까지라 `rows` 가 192 까지 나와요. 이 숫자를 보는 것은 테스트뿐이에요(형제 SPEC 에서 검색으로 읽었고 이 변경에서 다시 검색해 확인했어요: Result).
- 표를 소스 글로 읽는 이웃 테스트가 있는지 찾아야 해요: 형제 조각에서는 `tests/green_test_lock.rs` 가 `const EXTENSIONS` 를 글로 읽어서 첫 영수증이 실패했어요. `DB_DEPENDENCIES` 와 의존성 이름을 `tests/`, `skills/`, `docs/` 에서 찾아 읽은 결과를 Result 에 적어요. `tests/skill_references_test.py` 는 Python 테스트라 근거로 쓰지 않아요.
- 새 키가 든 확장 파일을 이전 판 바이너리가 읽으면 모르는 키라서 거절돼요(`deny_unknown_fields`). 팀이 판을 섞어 쓰면 이전 판 쪽은 확장 파일을 못 읽고 배송 행만 써요. 이 SPEC 의 판으로 올리면 풀려요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 의존성 표를 새 파일이 아니라 같은 카탈로그 파일의 `database_dependencies` 행(`kind`, `dependencies`)으로 옮긴다. 대안: (b) 의존성을 manifest 행에 붙임(한 의존성이 여러 ecosystem 에 나타나요: `redis` 는 npm 과 Python 모두), (c) 별도 파일 | 카탈로그 읽기·병합·검증·거절·보고가 이미 한 파일을 기준으로 있고 형제 SPEC 도 같은 길을 갔어요. 기본: 이대로 | 키와 행을 지우고 표를 되돌려요 |
| D-2 | 읽을 수 없는 행이 하나라도 있으면 확장 파일을 통째로 거절한다. 대안: 나쁜 행만 건너뜀 | 부모 SPEC 의 REQ-6 과 형제 SPEC 의 D-2 와 같은 규칙이에요. 기본: 통째 거절 | 거절 대신 건너뜀으로 바꿔요 |
| D-3 | 배송 의존성은 프로젝트 행이 바꾸지 못한다. 같은 행의 새 의존성만 쓴다 | 부모의 REQ-5 와 형제의 D-3 과 같은 규칙이에요. 기본: 이대로 | 프로젝트 행이 이기게 해요 |
| D-4 | 이 조각은 부모의 개정이 아니라 `parent` 로 이은 자식 SPEC 이다 | 부모는 이미 열두 요구사항이고 개정하면 승인된 digest 가 바뀌어요. 형제 SPEC 의 D-4 와 같아요. 기본: 자식 SPEC | 부모에 합쳐요 |
| D-5 | 의존성 이름은 소문자 영문·숫자와 `- _ . @ /` 로만 120 바이트까지, 글자나 숫자를 하나 이상 넣어 쓸 수 있다(`.` 나 `/` 만 있는 이름은 거절). 대문자가 든 행은 거절한다. 대안: 읽을 때 소문자로 바꿈 | 엔진은 모은 의존성 이름을 소문자로 바꿔서 비교해요(`dependency_names` 의 마지막 줄). 대문자 행은 어떤 이름과도 맞지 않아서 조용히 쓸모없어지고, 사용자가 보는 행과 적용되는 행이 달라져요. 거절하면 이유가 `problem` 에 나와요. 이름은 비교에만 쓰이고 기록되지 않아서 `recordable` 대신 이 규칙을 따로 둬요(`/` 와 `.` 가 필요해요). 종류는 `detected.databases` 에 기록되므로 `recordable` 을 그대로 써요. 기본: 거절 | 소문자로 바꾸어 읽게 해요 |
| D-6 | 한 확장 파일 안에서 한 의존성이 서로 다른 두 종류로 가면 파일 통째로 거절하고 `problem` 에 그 의존성을 적는다. 같은 종류로 두 번은 읽는다 | 형제 SPEC 의 D-5 와 같은 규칙이에요: 순서로만 정해지는 답은 사용자가 보는 목록과 적용되는 목록을 달라지게 해요. 같은 종류 반복은 의미가 같아요. 기본: 거절 | 앞 행이 이기게 해요 |
| D-7 | 배송 의존성만 가진 프로젝트 행은 버려진 것을 답이 말하지 않는다(`read`, `rows` 는 파일의 행 수, `problem` null). 형제 SPEC 의 같은 결정(D-6)과 함께, 리더의 노트 8756 이 SPEC 글로 정리하라고 한 두 항목 가운데 하나예요 | `manifest_catalog` 의 세 키는 부모 SPEC 의 테스트가 객체 통째로 고정해서 키를 더하면 그 계약이 바뀌어요. 알림이 필요하면 manifest·확장자·의존성 세 종류를 함께 후속으로 정해요. 기본: 이대로 | `manifest_catalog` 에 무시된 행 수를 더하고 부모의 테스트를 고쳐요 |
| D-8 | 노트 8756 의 다른 항목: 대소문자만 다른 확장자(`ex` 와 `EX`)는 충돌이 아니다. 확장자는 쓴 그대로 비교해서 서로 다른 두 이름이기 때문이다. 의존성 이름에는 같은 경우가 없다 — 대문자는 D-5 로 거절돼요 | 형제 SPEC 은 이 경우를 적지 않았고 그 SPEC 의 digest 는 승인됐으니 고치지 않아요(고치면 새 승인이 필요해요). 코드 동작은 이미 이렇고(형제 SPEC D-5 의 충돌 검사는 쓴 그대로 비교해요), 이 조각은 글로만 남겨요. 기본: 충돌 아님 | 대소문자 무시로 비교하게 하고 형제의 테스트를 고쳐요 |

## Result

기준 팁 NeXT `fd8340b166`. SPEC 초안 커밋 `aa1e33619a`, 계획 점검 뒤 다듬은 커밋 `d95db026d7`, RED 커밋 `6e9e85d63d`, GREEN 커밋 `ed6f0071b9`(아래 실행과 가드 뮤턴트는 이 머리 위에서 돌렸어요). 개정은 1 그대로예요: 아직 승인 전 초안이라 계획 점검의 열두 가지 빈틈을 본문에 바로 접었어요.

- 계획 점검(`moai-plan`)이 찾은 것: 검증이 의존성 이름에 `recordable` 을 쓰면 배송 파일이 거절되고 `shipped()` 가 카탈로그를 통째로 비워요(그래서 `dependency_name` 을 따로 뒀어요). `tests/path_root_direct_sites.rs` 는 `src/project_detect.rs` 를 글로 읽어요(이 변경이 더한 줄에는 옛 상태 폴더 철자가 없어요: 더한 `src` 줄 중 `moai-cli` 가 든 줄 0 개, `git diff fd8340b166..ed6f0071b9 -- src`). `tests/green_test_lock.rs` 와 `tests/project_detect_cli.rs` 는 고칠 곳이 없어요. `tests/skill_references_test.py` 는 Python 테스트라 오래돼서 근거로 쓰지 않아요. 공유 파일은 쓰지 않았어요.
- Run 1 (RED, `cargo nextest run --locked --offline --no-fail-fast --test project_detect_monorepo_cli --lib -E 'binary(project_detect_monorepo_cli) or test(the_shipped_catalog_holds_the_earlier_database_dependencies)'`; `target/kenan/red.log`): `30 tests run: 24 passed, 6 failed`. 새 테스트 여섯이 옳은 이유로 실패했어요 — 확장 파일의 모르는 키 `database_dependencies` 때문에 `manifest_catalog.extension` 이 `refused`(`the file is not a valid catalog`)예요, 단위 테스트는 배송 파일에 `database_dependencies` 키가 없어서 멈췄어요. 이전 쌍을 지키는 `every_shipped_database_dependency_keeps_its_kind` 는 현재 동작을 지켜서 통과해요. `a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file` 는 대조(고친 행)가 먼저 실패해서 나쁜 행 반복에 닿지 못했어요: 나쁜 행마다의 거절은 GREEN 실행과 가드 뮤턴트가 증명해요.
- Run 2 (GREEN, `cargo nextest run --locked --offline --no-fail-fast --test project_detect_monorepo_cli --test project_detect_cli --test green_test_lock --test project_scale_detect_lib --lib -E 'binary(project_detect_monorepo_cli) or binary(project_detect_cli) or binary(green_test_lock) or binary(project_scale_detect_lib) or test(project_detect)'`; `target/kenan/green.log`): `88 tests run: 88 passed, 1996 skipped`.
- `cargo clippy --locked --offline --lib --test project_detect_monorepo_cli -- -D warnings`: 종료 0(`target/kenan/clippy.log`). `cargo fmt --all --check`: 레인이 종료 0 으로 보고했지만 `target/kenan/` 에 그 로그가 없어요; `moai check affected --card` 가 `cargo fmt` 를 함께 돌리니 READY 영수증이 정해요.
- AC-1: met — `every_shipped_database_dependency_keeps_its_kind`(이전 표의 스물두 의존성을 하나씩 따로, 모두 함께는 다섯 종류 `mongodb mysql postgres redis sqlite`)와 단위 테스트 `the_shipped_catalog_holds_the_earlier_database_dependencies`(배송 파일의 의존성 행이 정확히 이전 스물두 쌍이고 이름이 `dependency_name` 규칙을 지키고 종류가 기록 가능하며, 배송 파일이 자기 검증을 통과하고 manifest 행과 코드 확장자 행이 비어 있지 않음) 통과(Run 2).
- AC-2: met — `a_database_dependency_row_adds_a_kind_without_code` 통과: `sqlx` 는 파일 없이 `databases` 가 비고, 행 하나로 `postgres`, `rows` 1, 읽어도 트리가 그대로, 세 종류의 행이 있으면 `rows` 3, 파일을 지우면 다시 비어요(Run 2; RED 에서는 `refused`).
- AC-3: met — `a_database_dependency_row_never_replaces_a_shipped_dependency` 통과: `pg`·`sqlx` 를 `mysql` 로 올리는 행은 `pg` 를 바꾸지 못하고 `sqlx` 만 `mysql` 로 쓰여요(`databases` 는 `mysql postgres`), `pg` 만 선언한 프로젝트는 `postgres` 만 내요, `pg` 만 가진 행은 아무것도 더하지 못하고 `read`·`rows` 1 이에요(Run 2).
- AC-4: met — `a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file` 통과: 대조(고친 행)는 `read`·`rows` 3 이고 manifest·코드·의존성 행이 모두 쓰이며, 읽을 수 없는 행 열다섯 가지(빈 목록, `dependencies` 키 없음, `kind` 키 없음, 빈 종류, 공백 든 종류, `/` 든 종류, 숫자 종류, 빈 이름, `.` 만 있는 이름, 공백 든 이름, `+` 든 이름, 대문자 든 이름, 121 바이트 이름, 모르는 키, 목록 아닌 값)은 모두 `refused`·`rows` 0·`problem` 문자열이고 같은 파일의 좋은 manifest·코드 행도 쓰이지 않고 배송 행만 나오며 트리가 그대로예요(Run 2).
- AC-5: met — `the_database_dependency_row_limits_hold_at_both_edges` 통과: 행 64 개와 한 행의 의존성 64 개와 120 바이트 이름은 `read`, 65 개와 121 바이트 이름은 `refused`·`rows` 0 이에요(Run 2).
- AC-6: met — `one_dependency_for_two_kinds_in_the_extension_file_is_refused` 통과: `sqlx` 를 `postgres`·`mysql` 로 올리는 두 행은 어느 쪽이 앞이든 `refused`·`rows` 0·`problem` 이 `the dependency sqlx is claimed by two kinds: <앞 종류> and <뒤 종류>` 이고 배송 종류만 나오며 트리가 그대로예요; 같은 종류를 한 행 안에서 두 번, 두 행에서 두 번 쓰면 `read` 예요(Run 2).
- AC-7: 판정은 READY 영수증이 해요. 이웃 대상 `project_detect_cli`, `project_scale_detect_lib`, `green_test_lock` 과 `--lib project_detect` 는 Run 2 에서 통과했어요. `home_isolation_scan` 은 여기서 `not_run` 이고 READY 영수증(`moai check affected --card`, 최종 머리 위)이 정해요; 통과로 적지 않아요.
- 가드 뮤턴트(열여덟 개, 모두 죽었어요): 한 번에 한 곳만 고치고 `nextest` 를 같은 필터로 돌려 모두 컴파일되고 지정한 테스트가 FAIL 줄에 들어 있으며 종료 100 이에요. 한 번 돌릴 때마다 `git` 으로 파일을 복원했고 두 실행 뒤 `git status --short` 가 비어 있어요. 실행기는 bash 스크립트 `target/kenan/mutants.sh` 예요(출력 `target/kenan/mutants-run1.out`, `target/kenan/mutants-run2.out`). 첫 실행에서 둘(대문자 허용, 문자 집합 검사 삭제)은 패턴이 단위 테스트에도 있어 스크립트가 건너뛰었어요; 패턴을 좁혀 그 둘만 다시 돌렸고 둘 다 죽었어요(Run 2 출력).
  - guard-mutant: a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file | deleted deny_unknown_fields from DbRow (unknown key in a row accepted) | exit 100
  - guard-mutant: the_database_dependency_row_limits_hold_at_both_edges | name length limit 120 becomes 121 | exit 100
  - guard-mutant: the_database_dependency_row_limits_hold_at_both_edges | name length limit 120 becomes 119 | exit 100
  - guard-mutant: a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file | a row with no dependency accepted | exit 100
  - guard-mutant: a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file | the kind is no longer checked | exit 100
  - guard-mutant: a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file | dependency_name drops the letter-or-digit clause | exit 100
  - guard-mutant: the_database_dependency_row_limits_hold_at_both_edges | row count limit dropped | exit 100
  - guard-mutant: the_database_dependency_row_limits_hold_at_both_edges | dependencies per row limit plus one | exit 100
  - guard-mutant: one_dependency_for_two_kinds_in_the_extension_file_is_refused | conflict compare flipped (same kind refused, two kinds read) | exit 100
  - guard-mutant: one_dependency_for_two_kinds_in_the_extension_file_is_refused | conflict problem omits the dependency name | exit 100
  - guard-mutant: a_database_dependency_row_never_replaces_a_shipped_dependency | lookup takes the last row instead of the first | exit 100
  - guard-mutant: a_database_dependency_row_never_replaces_a_shipped_dependency | detection unions the kind of every row that holds a declared dependency | exit 100
  - guard-mutant: a_database_dependency_row_adds_a_kind_without_code | the project rows are not appended to the catalog | exit 100
  - guard-mutant: a_database_dependency_row_adds_a_kind_without_code | rows no longer counts database dependency rows | exit 100
  - guard-mutant: the_shipped_catalog_holds_the_earlier_database_dependencies | validation uses recordable for dependency names (shipped file refused) | exit 100
  - guard-mutant: every_shipped_database_dependency_keeps_its_kind | shipped ioredis dependency removed from the catalog | exit 100
  - guard-mutant: a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file | dependency_name accepts upper-case letters | exit 100
  - guard-mutant: a_database_dependency_row_that_cannot_be_read_refuses_the_whole_file | dependency_name drops the character set check | exit 100
- 표를 글로 읽는 이웃: `DB_DEPENDENCIES` 를 `src`, `tests`, `skills`, `docs` 에서 다시 찾아 한 곳도 없었어요(`grep -rn`; 표는 `src/project_detect.rs` 에서 지워졌어요). `manifest_catalog` 의 `rows` 를 읽는 곳은 `tests/project_detect_monorepo_cli.rs` 뿐이에요(`src`, `tests`, `skills`, `docs` 를 찾아 확인했고 `rows` 가 세 종류의 합으로 넓어져도 읽는 코드가 없어요).
- 위험 경로: 읽기 전용 감지라서 지우기·reset·프로세스 시작·권한 규칙이 없어요. 테스트는 빌드된 바이너리를 임시 폴더와 임시 HOME 에서 돌려요.
- 카드 기준 표시(`moai card criterion`, 카드 기준 5 와 3)는 이 커밋 뒤에 레인이 해요. SPEC 의 일이 아니에요.
- 남은 것(Gaps): `skills` 와 `docs` 에 확장 파일의 `database_dependencies` 키를 설명하는 글이 없어요; 확장 파일 이름(`detect-catalog`), `code_extensions`, `manifest_catalog` 를 `skills`, `docs`, `README*` 에서 찾아도 한 곳도 없어서 `manifests` 키를 설명하는 글도 찾지 못했어요. 확장 파일의 키를 설명하는 글은 형제 SPEC 과 같은 후속이에요. 버려진 행 수 알림(D-7)과 `prisma_database` 의 provider → 종류 대응은 후속으로 남아요. `tests/skill_references_test.py`(Python)는 범위 밖이에요.
- 돌리지 않은 것: 전체 `cargo test`, 합친 트리 위 스캔, Windows·Linux, 이전 판 바이너리가 새 키가 든 확장 파일을 거절하는지, `home_isolation_scan` 단독.
