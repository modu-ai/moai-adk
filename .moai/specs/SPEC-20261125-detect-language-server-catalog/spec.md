---
id: SPEC-20261125-detect-language-server-catalog
title: "project-detect reads the language servers and their official plugins from the catalog, not from a table in the engine"
status: draft
revision: 1
created: 2026-11-25
parent: SPEC-20261124-project-detect-monorepo
---

# project-detect reads the language servers and their official plugins from the catalog, not from a table in the engine

## Intent

사용자 지시(2026-11-24): MoAI 는 모든 개발 언어와 도구를 지원해야 하고 언어·도구 하드코딩은 금지예요. 부모 SPEC `SPEC-20261124-project-detect-monorepo` 가 manifest 인식을, 자식 SPEC 둘(`SPEC-20261125-detect-code-extension-catalog`, `SPEC-20261125-detect-database-dependency-catalog`)이 코드 확장자 표와 데이터베이스 의존성 표를 데이터 카탈로그와 프로젝트 확장 파일로 옮겼어요. 부모 카드의 남은 기준 하나가 언어 서버 표예요: `src/project_detect.rs` 의 `LSP_PLUGINS` 상수 표(언어 열네 개 → 서버 이름, Claude Code 공식 LSP 플러그인, 실행 파일)가 아직 엔진 안에 있어서, 표에 없는 언어(예: Elixir, Zig)는 코드 파일이 감지돼도 `detected.lsp` 와 `lsp_plugins` 에 서버가 나타나지 않고 새 언어마다 MoAI 를 고쳐야 해요.

이 SPEC 은 그 표 하나만 같은 카탈로그 파일의 `language_servers` 행으로 옮겨요. 읽기·병합·검증·거절·보고 장치(배송 행 우선, 목록마다 한도 64, 읽을 수 없으면 확장 파일 통째 거절)는 앞선 SPEC 들이 만든 것을 그대로 써요. 읽는 사람: 레인, 이 변경의 감사자, `/moai project` 의 사용자.

## Non-goals

- `commands_fact` 의 명령 템플릿과 `project_scale_detect` 의 작업공간 표지는 이 SPEC 이 아니에요. 부모 카드의 다른 기준이 한 조각씩 맡아요.
- 공식 마켓플레이스 이름(`LSP_MARKETPLACE`, `claude-plugins-official`)은 언어마다 다른 값이 아니라 하나의 사실이라 엔진에 남겨요. 이 SPEC 은 언어 → 서버 대응만 옮겨요.
- 플러그인 설치와 실행은 이 SPEC 이 하지 않아요. `lsp_plugins` 는 PATH 에 실행 파일이 있는지 읽기만 하고, 설치는 지금처럼 PROJECT 경로에서 사용자의 선택을 받은 뒤에만 해요(공통 규칙 25).
- 언어가 어디서 왔는지(manifest 행, 코드 확장자 행, 확장 파일)는 바꾸지 않아요. `languages` 의 계산은 그대로이고 그 언어의 서버를 찾는 출처만 바뀌어요.
- 한도 상수(`MAX_DEPTH`, `MAX_ENTRIES`, `MAX_ITEMS`, `MAX_FILE`)와 걷기 순서·예산은 바꾸지 않아요.
- 부모 SPEC 과 앞선 자식 SPEC 들의 본문과 승인된 digest 는 건드리지 않아요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: The shipped catalog shall hold the languages and the server name, official plugin and executable of each that the earlier `LSP_PLUGINS` table held, so that a project with one of those languages reports the same `detected.lsp` and `lsp_plugins` as before. · failure: none — a unit test pins the rows and a CLI test runs one project per language.
- REQ-2: When the extension file holds `language_servers` rows, the command shall report the server of a row in `detected.lsp` and its plugin in `lsp_plugins` for each detected language that the row names, the shipped row prevailing for a language that both hold. · failure: none — the extension file is optional.
- REQ-3 (REQ-2): IF a `language_servers` row lacks one of `language`, `server`, `plugin` and `executable`, holds a value that cannot be recorded, a key the catalog does not know, or the file holds more than 64 such rows, THEN the command shall refuse the whole extension file, use the shipped rows alone and report the problem in `manifest_catalog`.
- REQ-4: When the extension file is read, the command shall count its `language_servers` rows with its `manifests`, `code_extensions` and `database_dependencies` rows in `manifest_catalog.rows`. · failure: none — the number is a report, not a decision.
- REQ-5 (REQ-2): IF two `language_servers` rows of the extension file give one language two different servers, plugins or executables, THEN the command shall refuse the whole extension file, use the shipped rows alone and report a problem that names the language; the identical row more than once is not a conflict.

## Acceptance

- AC-1 (REQ-1): 이전 표의 열네 언어(`rust python javascript typescript go java kotlin ruby php c cpp csharp swift lua`)마다 그 언어 하나의 코드 파일만 둔 프로젝트는 이전과 같은 서버 이름을 `detected.lsp` 에 내고 `lsp_plugins` 에 같은 플러그인·마켓플레이스·실행 파일을 내요(`javascript` 와 `typescript`, `c` 와 `cpp` 는 서버가 같아서 `detected.lsp` 에서 한 이름으로 합쳐져요). 배송 파일이 읽히지 않아 카탈로그가 비는 일이 없도록, 단위 테스트는 배송 manifest 행과 코드 확장자 행과 의존성 행과 서버 행이 모두 비어 있지 않은 것도 지켜요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli every_shipped_language_server_keeps_its_plugin` and `cargo test --locked --offline --lib the_shipped_catalog_holds_the_earlier_language_servers`
- AC-2 (REQ-2, REQ-4): 카탈로그에 서버가 없는 언어(`lib/a.ex` 가 코드 확장자 행으로 `elixir` 가 된 프로젝트)는 확장 파일의 `language_servers` 행 하나로 `detected.lsp` 에 그 서버가, `lsp_plugins` 에 그 플러그인 행이 나타나고 파일을 지우면 사라져요. `manifest_catalog.rows` 는 네 종류의 행을 함께 세요. 확장 파일을 읽어도 트리는 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_language_server_row_adds_a_server_without_code`
- AC-3 (REQ-2): 배송 언어(`rust`)의 서버를 다른 것으로 올리려는 행은 그 언어를 바꾸지 못하고(`detected.lsp` 는 `rust-analyzer` 그대로, `lsp_plugins` 의 `rust-analyzer-lsp` 행도 그대로) 같은 파일의 새 언어 행만 쓰여요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_language_server_row_never_replaces_a_shipped_server`
- AC-4 (REQ-3): 읽을 수 없는 `language_servers` 행(`language`·`server`·`plugin`·`executable` 중 하나가 없음, 빈 값, 공백이나 `/` 가 든 값, 81 바이트 값, 모르는 키, 문자열이 아닌 값)이 하나라도 있으면 파일 전체가 `manifest_catalog.extension` `refused` 이고 `problem` 이 문자열이며, 같은 파일의 좋은 `manifests`·`code_extensions`·`database_dependencies` 행도 쓰이지 않고 배송 행만 쓰이며 트리가 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_language_server_row_that_cannot_be_read_refuses_the_whole_file`
- AC-5 (REQ-3): 행 64 개와 80 바이트 값은 읽히고 65 개와 81 바이트 값은 거절돼요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli the_language_server_row_limits_hold_at_both_edges`
- AC-6 (REQ-5): `elixir` 를 서버가 다른 두 행(또는 플러그인만, 실행 파일만 다른 두 행)으로 올리는 파일은 어느 쪽이 앞이든 통째로 `refused` 이고 `rows` 0 이며 `problem` 이 `the language elixir is claimed by two different language server rows` 이고 배송 서버만 나오며 트리가 바뀌지 않아요. 똑같은 행을 두 번 쓰면 `read` 예요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli one_language_for_two_servers_in_the_extension_file_is_refused`
- AC-7: 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test project_detect_cli --test project_detect_monorepo_cli --test project_scale_detect_lib --test green_test_lock` and `cargo test --locked --offline --lib project_detect`

## Scope

- write: `.moai/specs/SPEC-20261125-detect-language-server-catalog/**`, `src/project_detect.rs`(`LSP_PLUGINS` 표 삭제, `CatalogFile` 의 새 키, 서버 행 검증·병합·조회, `lsp_plugins_fact` 와 `detect_full` 이 카탈로그를 읽게 함, 단위 테스트), `src/project_detect_catalog.json`(`language_servers` 행), `tests/project_detect_monorepo_cli.rs`(테스트 더하기)
- read-only: `src/project_scale_detect.rs`, `src/ops/project_detect.rs`, `tests/project_detect_cli.rs`, `tests/green_test_lock.rs`, `skills/moai/workflows/project.md`
- forbidden: `commands_fact`·`prisma_database`·`dependency_names`·`MARKER_FILES` 의 표와 본문, `src/profile.rs`, `src/project_update.rs`, `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`, `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME, 한도 상수, 앞선 SPEC 들의 파일

## Plan

1. SPEC 초안을 쓰고 커밋해요. 승인은 리더가 노트로 해요(위임); 레인은 승인을 기록하지 않고 초안 위에서 RED 부터 구현해요.
2. RED: `tests/project_detect_monorepo_cli.rs` 에 AC-1 ~ AC-6 의 테스트를 먼저 써서 현재 팁에서 AC-2 ~ AC-6 이 옳은 이유(확장 파일의 모르는 키 `language_servers` 가 거절됨)로 실패하는 것을 봐요. AC-1 의 CLI 테스트는 이미 맞는 동작을 지켜서 통과해요.
3. GREEN: `CatalogFile` 에 `language_servers` 를, `Catalog` 에 서버 행을 더하고, `catalog_file_valid` 에 행 검사와 한 파일 안의 한 언어 두 서버 거절을 더하고, `detect_full` 의 `lsp` 계산과 `lsp_plugins_fact` 가 카탈로그를 읽게 하고 `LSP_PLUGINS` 를 지워요. 배송 파일에 열네 행을 적어요. 확장 병합은 프로젝트 행을 배송 행 뒤에 덧붙이기만 해요: 한 언어의 서버는 먼저 찾은 행이 정해요. `ExtensionReport.rows` 가 네 종류를 함께 세요. 단위 테스트 `tables_name_only_recordable_values` 는 카탈로그 행을 읽게 바꿔요.
4. 확인: 새 테스트, AC-7 의 이웃 테스트, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`(대상 한정), 새 가드 테스트마다 가드 뮤턴트 한 줄(`guard-mutant: <test> | <지운 코드> | exit <n>`), 위험 경로는 읽기 전용이라 해당 없음을 READY 에 적어요, `moai check affected --card` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적고, 카드의 기준 4 를 `moai card criterion` 으로 표시해요. 감사는 이 변경을 쓰지 않은 세션이 맡아요(tier 1).

## Risks

- 배송 `language_servers` 가 읽히지 않으면 `detected.lsp` 와 `lsp_plugins` 가 조용히 비어요. 더 나쁜 것은 검증이 배송 파일 자체를 거절하면 `shipped()` 가 통째로 비어서(`unwrap_or_default`) manifest 행, 코드 확장자 행, 의존성 행까지 함께 사라지는 거예요. 서버 행의 네 값은 모두 `recordable`(영문·숫자·`- _ . @`, 80 바이트까지)을 통과해야 하고 배송 값은 이전 표에서 가져오므로, 단위 테스트가 열네 행과 네 종류의 배송 행이 비어 있지 않음을 지켜요.
- 프로젝트 확장 파일이 새 언어의 플러그인 이름을 적을 수 있어요. `lsp_plugins` 는 그 이름을 보여 주기만 하고(PATH 의 실행 파일 존재 여부만 읽어요) 설치는 PROJECT 경로에서 사용자의 선택을 받은 뒤에만 해요; 마켓플레이스는 `claude-plugins-official` 로 고정이에요. 실행 파일 이름은 PATH 폴더 안에서 찾는 파일 이름이라 `/` 를 거절해서(`recordable`) 경로 탈출이 없어요.
- 프로젝트 확장 파일에 읽을 수 없는 행이 하나라도 있으면 파일 전체를 거절해서 좋은 행도 못 써요. 추측으로 일부만 받는 것보다 안전한 쪽이에요: 거절은 `manifest_catalog.problem` 으로 보여서 사용자가 행을 고쳐요(REQ-3).
- 한 서버가 여러 언어를 맡아요(`typescript-language-server` 는 `javascript` 와 `typescript`, `clangd` 는 `c` 와 `cpp`): 행은 언어마다 하나라서 이 경우는 행이 둘이고 `detected.lsp` 는 이름이 같아 한 줄로 합쳐요(지금 동작 그대로).
- `manifest_catalog.rows` 의 뜻이 네 종류의 행 수 합으로 넓어져요. 한 파일은 목록마다 64 개까지라 `rows` 가 256 까지 나와요. 이 숫자를 보는 것은 테스트뿐이에요(앞선 SPEC 에서 검색으로 읽음; 이 변경에서 다시 검색해 확인해요).
- 표를 소스 글로 읽는 이웃 테스트가 있는지 찾아야 해요: `LSP_PLUGINS` 와 서버 이름을 `tests/`, `skills/`, `docs/` 에서 찾아 읽은 결과를 Result 에 적어요. `tests/skill_references_test.py` 는 Python 테스트라 근거로 쓰지 않아요.
- 새 키가 든 확장 파일을 이전 판 바이너리가 읽으면 모르는 키라서 거절돼요(`deny_unknown_fields`). 팀이 판을 섞어 쓰면 이전 판 쪽은 확장 파일을 못 읽고 배송 행만 써요. 이 SPEC 의 판으로 올리면 풀려요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 언어 서버 표를 새 파일이 아니라 같은 카탈로그 파일의 `language_servers` 행(`language`, `server`, `plugin`, `executable`)으로 옮긴다. 대안: (b) 언어 행을 코드 확장자 행에 붙임(manifest 로만 감지되는 언어를 표현 못 함), (c) 별도 파일 | 카탈로그 읽기·병합·검증·거절·보고가 이미 한 파일을 기준으로 있고 앞선 두 자식 SPEC 도 같은 길을 갔어요. 기본: 이대로 | 키와 행을 지우고 표를 되돌려요 |
| D-2 | 읽을 수 없는 행이 하나라도 있으면 확장 파일을 통째로 거절한다. 대안: 나쁜 행만 건너뜀 | 부모 SPEC 의 REQ-6 과 앞선 자식 SPEC 의 D-2 와 같은 규칙이에요. 기본: 통째 거절 | 거절 대신 건너뜀으로 바꿔요 |
| D-3 | 배송 언어의 서버는 프로젝트 행이 바꾸지 못한다. 같은 파일의 새 언어 행만 쓴다 | 부모의 REQ-5 와 앞선 자식 SPEC 의 D-3 과 같은 규칙이에요. 기본: 이대로 | 프로젝트 행이 이기게 해요 |
| D-4 | 이 조각은 부모 SPEC 의 개정이 아니라 `parent` 로 이은 자식 SPEC 이다 | 부모는 이미 열두 요구사항이고 개정하면 승인된 digest 가 바뀌어요. 앞선 자식 SPEC 들의 D-4 와 같아요. 기본: 자식 SPEC | 부모에 합쳐요 |
| D-5 | 서버 행의 네 값은 모두 `recordable` 규칙을 쓴다(영문·숫자·`- _ . @`, 80 바이트까지, 공백과 `/` 거절). 대안: 이름마다 따로 규칙 | `server` 는 `detected.lsp` 에 기록되고 `plugin` 과 `executable` 은 보고에 실리며 `executable` 은 PATH 폴더 안의 파일 이름이라 `/` 가 있으면 안 돼요. 이전 표의 열네 행은 모두 이 규칙을 통과해요(단위 테스트가 지켜요). 대문자는 거절하지 않아요: 언어 이름은 카탈로그의 다른 행(manifest·확장자)이 쓴 그대로 비교하므로 정본 대소문자가 없어요(`Elixir` 와 `elixir` 는 다른 두 이름이고 충돌이 아니에요). 기본: 이대로 | 값마다 따로 규칙을 둬요 |
| D-6 | 한 확장 파일 안에서 한 언어가 서로 다른 두 서버 행(서버, 플러그인, 실행 파일 중 하나라도 다름)으로 가면 파일 통째로 거절하고 `problem` 에 그 언어를 적는다. 똑같은 행을 두 번 쓰면 읽는다 | 앞선 자식 SPEC 들의 D-5·D-6 과 같은 규칙이에요: 순서로만 정해지는 답은 사용자가 보는 목록과 적용되는 목록을 달라지게 해요. 기본: 거절 | 앞 행이 이기게 해요 |
| D-7 | 배송 언어만 가진 프로젝트 행은 버려진 것을 답이 말하지 않는다(`read`, `rows` 는 파일의 행 수, `problem` null). 앞선 자식 SPEC 들의 같은 결정과 같은 이유로, 알림이 필요하면 manifest·확장자·의존성·서버 네 종류를 함께 후속으로 정해요 | `manifest_catalog` 의 세 키는 부모 SPEC 의 테스트가 객체 통째로 고정해서 키를 더하면 그 계약이 바뀌어요. 기본: 이대로 | `manifest_catalog` 에 무시된 행 수를 더하고 부모의 테스트를 고쳐요 |

## Result

(구현 뒤에 써요.)
