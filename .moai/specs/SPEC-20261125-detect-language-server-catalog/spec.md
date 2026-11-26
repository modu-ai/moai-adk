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
- `src/project_scale.rs` 의 `PLUGINS` 표(언어 열네 개 → 플러그인, 자기만의 마켓플레이스 상수)는 이 SPEC 이 아니에요. `project-scale-plan` 이 쓰는 별개의 하드코딩 표라서 후속으로 둬요: 그때까지 프로젝트 확장 파일로 더한 언어는 `detected.lsp` 와 `lsp_plugins` 에는 나타나지만 `project-scale-plan` 의 플러그인 선택은 그 언어를 거절해요(Risks).
- 플러그인 설치와 실행은 이 SPEC 이 하지 않아요. `lsp_plugins` 는 PATH 에 실행 파일이 있는지 읽기만 하고, 설치는 지금처럼 PROJECT 경로에서 사용자의 선택을 받은 뒤에만 해요(공통 규칙 25).
- 언어가 어디서 왔는지(manifest 행, 코드 확장자 행, 확장 파일)는 바꾸지 않아요. `languages` 의 계산은 그대로이고 그 언어의 서버를 찾는 출처만 바뀌어요.
- 한도 상수(`MAX_DEPTH`, `MAX_ENTRIES`, `MAX_ITEMS`, `MAX_FILE`)와 걷기 순서·예산은 바꾸지 않아요.
- 부모 SPEC 과 앞선 자식 SPEC 들의 본문과 승인된 digest 는 건드리지 않아요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: The shipped catalog shall hold the languages and the server name, official plugin and executable of each that the earlier `LSP_PLUGINS` table held, so that a project with one of those languages reports the same `detected.lsp` and `lsp_plugins` as before. · failure: none — a unit test pins the rows and a CLI test runs one project per language.
- REQ-2: When the extension file holds `language_servers` rows, the command shall report the server of a row in `detected.lsp` and its plugin in `lsp_plugins` for each detected language that the row names, the shipped row prevailing for a language that both hold. · failure: none — the extension file is optional.
- REQ-3 (REQ-2): IF a `language_servers` row lacks one of `language`, `server`, `plugin` and `executable`, holds a value that cannot be recorded, a `plugin` with an `@` in it, a key the catalog does not know, or the file holds more than 64 such rows, THEN the command shall refuse the whole extension file, use the shipped rows alone and report the problem in `manifest_catalog`.
- REQ-4: When the extension file is read, the command shall count its `language_servers` rows with its `manifests`, `code_extensions` and `database_dependencies` rows in `manifest_catalog.rows`. · failure: none — the number is a report, not a decision.
- REQ-5 (REQ-2): IF two `language_servers` rows of the extension file give one language two different servers, plugins or executables, THEN the command shall refuse the whole extension file, use the shipped rows alone and report a problem that names the language; the identical row more than once is not a conflict.
- REQ-6 (REQ-2): IF the extension file gives one plugin two different executables, between its own rows or against a shipped row, THEN the command shall refuse the whole extension file, use the shipped rows alone and report a problem that names the plugin; the same executable for one plugin more than once is not a conflict.

## Acceptance

- AC-1 (REQ-1): 이전 표의 열네 언어(`rust python javascript typescript go java kotlin ruby php c cpp csharp swift lua`)마다 그 언어 하나의 코드 파일만 둔 프로젝트는 이전과 같은 서버 이름을 `detected.lsp` 에 내고 `lsp_plugins` 에 같은 플러그인·마켓플레이스·실행 파일을 내요(`javascript` 와 `typescript`, `c` 와 `cpp` 는 서버가 같아서 `detected.lsp` 에서 한 이름으로 합쳐져요). 열네 언어를 한 프로젝트에 모두 두면 `detected.lsp` 는 열두 이름이고 `lsp_plugins` 는 플러그인 이름 순으로 열두 행이에요(`clangd-lsp` 는 `c cpp`, `typescript-lsp` 는 `javascript typescript`; 테스트 환경은 PATH 가 없어서 `installed` 가 모두 `false`). 배송 파일이 읽히지 않아 카탈로그가 비는 일이 없도록, 단위 테스트는 배송 manifest 행과 코드 확장자 행과 의존성 행과 서버 행이 모두 비어 있지 않은 것도 지켜요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli every_shipped_language_server_keeps_its_plugin` and `cargo test --locked --offline --lib the_shipped_catalog_holds_the_earlier_language_servers`
- AC-2 (REQ-2, REQ-4): 카탈로그에 서버가 없는 언어(`lib/a.ex` 가 코드 확장자 행으로 `elixir` 가 된 프로젝트)는 확장 파일의 `language_servers` 행 하나로 `detected.lsp` 에 그 서버가, `lsp_plugins` 에 그 플러그인 행이 나타나고 파일을 지우면 사라져요. `manifest_catalog.rows` 는 네 종류의 행을 함께 세요(네 종류가 한 파일에 있으면 `{"extension":"read","rows":4,"problem":null}`). 서버 행이 있어도 감지되지 않은 언어의 서버는 `detected.lsp` 와 `lsp_plugins` 에 나타나지 않고, 코드 파일 없이 manifest 행으로만 감지된 언어(`apps/svc/mix.exs`)도 그 서버를 얻어요. 확장 파일을 읽어도 트리는 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_language_server_row_adds_a_server_without_code`
- AC-3 (REQ-2): 배송 언어(`rust`)의 서버를 다른 것으로 올리려는 행은 그 언어를 바꾸지 못하고(`detected.lsp` 는 `rust-analyzer` 그대로, `lsp_plugins` 의 `rust-analyzer-lsp` 행도 그대로이고 덮어쓰려던 플러그인 이름의 행은 생기지 않아요) 같은 파일의 새 언어 행은 쓰여요. 배송 언어 행 하나만 든 파일은 `read`·`rows` 1·`problem` null 이에요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_language_server_row_never_replaces_a_shipped_server`
- AC-4 (REQ-3): 읽을 수 없는 `language_servers` 행(`language`·`server`·`plugin`·`executable` 네 값마다: 키가 없음, 빈 값, 공백이 든 값, `/` 가 든 값, 점만 있는 값(`..`), 81 바이트 값, 문자열이 아닌 값; 그리고 `@` 가 든 `plugin`, 행 안의 모르는 키, 목록이 아닌 `language_servers`)이 하나라도 있으면 파일 전체가 `manifest_catalog.extension` `refused` 이고 `problem` 이 문자열이며, 같은 파일의 좋은 `manifests`·`code_extensions`·`database_dependencies` 행도 쓰이지 않고 배송 행만 쓰이며 트리가 바뀌지 않아요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli a_language_server_row_that_cannot_be_read_refuses_the_whole_file`
- AC-5 (REQ-3): 행 64 개(언어가 모두 달라요)와 네 값 각각의 80 바이트는 읽히고, 65 개와 네 값 각각의 81 바이트는 거절돼요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli the_language_server_row_limits_hold_at_both_edges`
- AC-6 (REQ-5): `elixir` 를 서버가 다른 두 행(또는 플러그인만, 실행 파일만 다른 두 행)으로 올리는 파일은 어느 쪽이 앞이든 통째로 `refused` 이고 `rows` 0 이며 `problem` 이 `the language elixir is claimed by two different language server rows` 이고 배송 서버만 나오며 트리가 바뀌지 않아요. 배송 언어(`rust`)를 서버가 다른 두 행으로 올리는 파일도 같아요(충돌 검사는 파일 안에서만 해요). 똑같은 행을 두 번 쓰면 `read`·`rows` 2 이고 서버가 한 번만 나오며, `Elixir` 와 `elixir` 는 서로 다른 이름이라 `read` 예요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli one_language_for_two_servers_in_the_extension_file_is_refused`
- AC-7: 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test project_detect_cli --test project_detect_monorepo_cli --test project_lsp_plugins_cli --test project_scale_detect_lib --test green_test_lock --test path_root_direct_sites --test home_isolation_scan` and `cargo test --locked --offline --lib project_detect`
- AC-8 (REQ-6): 새 언어(`elixir`)의 플러그인이 배송 플러그인(`rust-analyzer-lsp`)과 이름이 같은데 실행 파일이 다르면, 그리고 파일 안의 두 새 언어가 한 플러그인 이름에 실행 파일을 둘 쓰면 파일 전체가 `refused` 이고 `rows` 0 이며 `problem` 이 `the plugin <이름> is claimed with two different executables` 이고 배송 행만 나오며 트리가 바뀌지 않아요. 같은 플러그인·같은 실행 파일을 두 언어가 쓰면(`typescript-lsp` 와 `javascript`·`typescript` 처럼) `read` 예요 · verify: `cargo test --locked --offline --test project_detect_monorepo_cli one_plugin_with_two_executables_is_refused`

## Scope

- write: `.moai/specs/SPEC-20261125-detect-language-server-catalog/**`, `src/project_detect.rs`(`LSP_PLUGINS` 표 삭제, `CatalogFile` 의 새 키, 서버 행 검증·병합·조회, `lsp_plugins_fact` 와 `detect_full` 이 카탈로그를 읽게 함, 단위 테스트), `src/project_detect_catalog.json`(`language_servers` 행), `tests/project_detect_monorepo_cli.rs`(테스트 더하기)
- read-only: `src/project_scale_detect.rs`, `src/project_scale.rs`, `src/ops/project_detect.rs`, `tests/project_detect_cli.rs`, `tests/project_lsp_plugins_cli.rs`, `tests/green_test_lock.rs`, `tests/path_root_direct_sites.rs`, `skills/moai/workflows/project.md`
- forbidden: `commands_fact`·`prisma_database`·`dependency_names`·`MARKER_FILES`·`PLUGINS`(`src/project_scale.rs`) 의 표와 본문, `src/profile.rs`, `src/project_update.rs`, `src/language.rs`, `src/agents.rs`, `src/ops/project_apply.rs`, `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME, 한도 상수, 앞선 SPEC 들의 파일

## Plan

1. SPEC 초안을 쓰고 커밋해요. 승인은 리더가 노트로 해요(위임); 레인은 승인을 기록하지 않고 초안 위에서 RED 부터 구현해요.
2. RED: `tests/project_detect_monorepo_cli.rs` 에 AC-1 ~ AC-6 과 AC-8 의 테스트를 먼저 써서 현재 팁에서 AC-2 ~ AC-6 과 AC-8 이 옳은 이유(확장 파일의 모르는 키 `language_servers` 가 거절됨)로 실패하는 것을 봐요. 나쁜 행 테스트는 고친 파일이 `read` 인 대조를 맨 앞에 두어서 RED 가 그 대조에서 멈추게 하고, 단위 테스트는 현재 팁에서 컴파일돼요(`CATALOG_JSON` 을 `serde_json::Value` 로 읽어요). AC-1 의 CLI 테스트는 이미 맞는 동작을 지켜서 통과해요.
3. GREEN(JSON 의 새 키와 `CatalogFile` 의 새 필드는 같은 커밋에 들어가야 해요: 필드 없이 키만 있으면 `deny_unknown_fields` 때문에 배송 카탈로그가 통째로 비어요): `CatalogFile` 에 `language_servers` 를, `Catalog` 에 서버 행을 더하고, `catalog_file_valid` 에 행 검사(`plugin` 의 `@` 거절 포함)와 한 파일 안의 한 언어 두 서버 거절과 한 플러그인 두 실행 파일 거절을 더하고, 프로젝트 행의 플러그인이 배송 행과 같은 이름에 다른 실행 파일을 쓰는지는 병합 때(`for_project`) 배송 행과 견주어 거절하고, `detect_full` 의 `lsp` 계산과 `lsp_plugins_fact` 가 카탈로그를 읽게 하고(`detect_cli` 의 호출은 `&catalog` 를 넘겨요), `LSP_PLUGINS` 를 지워요. 지워지는 표의 출처 주석(마켓플레이스의 `plugin.json` 에서 2026-11-23 에 읽음, `SPEC-20261123-lsp-official-plugins`)은 새 행 구조체의 문서 주석으로 옮겨요. 배송 파일에 열네 행을 적어요. 확장 병합은 프로젝트 행을 배송 행 뒤에 덧붙이기만 해요: 한 언어의 서버는 먼저 찾은 행이 정해요. `ExtensionReport.rows` 가 네 종류를 함께 세요. 단위 테스트 `tables_name_only_recordable_values` 는 카탈로그 행을 읽게 바꿔요.
4. 확인: 새 테스트, AC-7 의 이웃 테스트, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`(대상 한정), 새 가드 테스트마다 가드 뮤턴트 한 줄(`guard-mutant: <test> | <지운 코드> | exit <n>`), 위험 경로는 읽기 전용이라 해당 없음을 READY 에 적어요, `moai check affected --card` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적고, 카드의 기준 4 를 `moai card criterion` 으로 표시해요. 감사는 이 변경을 쓰지 않은 세션이 맡아요(tier 1).

## Risks

- 배송 `language_servers` 가 읽히지 않으면 `detected.lsp` 와 `lsp_plugins` 가 조용히 비어요. 더 나쁜 것은 검증이 배송 파일 자체를 거절하면 `shipped()` 가 통째로 비어서(`unwrap_or_default`) manifest 행, 코드 확장자 행, 의존성 행까지 함께 사라지는 거예요. 서버 행의 네 값은 모두 `recordable`(영문·숫자·`- _ . @`, 80 바이트까지)을 통과해야 하고 배송 값은 이전 표에서 가져오므로, 단위 테스트가 열네 행과 네 종류의 배송 행이 비어 있지 않음을 지켜요.
- 프로젝트 확장 파일이 새 언어의 플러그인 이름을 적을 수 있어요. `lsp_plugins` 는 그 이름을 보여 주기만 하고(PATH 의 실행 파일 존재 여부만 읽어요) 설치는 PROJECT 경로에서 사용자의 선택을 받은 뒤에만 해요; 마켓플레이스는 `claude-plugins-official` 로 고정이에요. 실행 파일 이름은 PATH 폴더 안에서 찾는 파일 이름이라 `/` 를 거절해서(`recordable`) 경로 탈출이 없어요.
- 프로젝트 확장 파일에 읽을 수 없는 행이 하나라도 있으면 파일 전체를 거절해서 좋은 행도 못 써요. 추측으로 일부만 받는 것보다 안전한 쪽이에요: 거절은 `manifest_catalog.problem` 으로 보여서 사용자가 행을 고쳐요(REQ-3).
- `src/project_scale.rs` 에는 `project-scale-plan` 이 쓰는 별개의 언어 → 플러그인 표 `PLUGINS` 가 있어요(자기만의 마켓플레이스 상수도 있어요). 이 SPEC 뒤에 프로젝트 확장 파일로 더한 언어는 `detected.lsp` 와 `lsp_plugins` 에는 나타나지만 `project-scale-plan` 의 플러그인 선택은 "not a detected language with a code intelligence plugin" 로 거절해요. 후속으로 그 표도 카탈로그를 읽게 해요(`docs/later.md` 한 줄은 리더가 올려요).
- `plugin` 이름은 PROJECT 경로에서 `<plugin>@claude-plugins-official` 로 이어 붙여져서, `@` 가 든 이름이면 고정된 마켓플레이스를 흐릴 수 있어요. 그래서 `plugin` 에서만 `@` 를 거절해요(D-5). 같은 플러그인이 서로 다른 실행 파일을 가지면 `lsp_plugins` 는 한 행에 하나만 보이므로 순서가 답을 정하게 돼요: 그래서 거절해요(D-8).
- 프로젝트가 더한 서버는 `skills/moai/workflows/project.md` 의 서버별 설치 표에 행이 없어요(`skills/**` 는 이 SPEC 의 금지 경로). 후속으로 두어요.
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
| D-5 | 서버 행의 네 값은 모두 `recordable` 규칙을 쓴다(영문·숫자·`- _ . @`, 80 바이트까지, 공백과 `/` 거절). 단 `plugin` 에서는 `@` 도 거절한다(PROJECT 경로가 `<plugin>@claude-plugins-official` 로 이어 붙여요). 대안: 이름마다 따로 규칙 | `server` 는 `detected.lsp` 에 기록되고 `plugin` 과 `executable` 은 보고에 실리며 `executable` 은 PATH 폴더 안의 파일 이름이라 `/` 가 있으면 안 돼요. 이전 표의 열네 행은 모두 이 규칙을 통과해요(단위 테스트가 지켜요). 대문자는 거절하지 않아요: 언어 이름은 카탈로그의 다른 행(manifest·확장자)이 쓴 그대로 비교하므로 정본 대소문자가 없어요(`Elixir` 와 `elixir` 는 다른 두 이름이고 충돌이 아니에요). 기본: 이대로 | 값마다 따로 규칙을 둬요 |
| D-6 | 한 확장 파일 안에서 한 언어가 서로 다른 두 서버 행(서버, 플러그인, 실행 파일 중 하나라도 다름)으로 가면 파일 통째로 거절하고 `problem` 에 그 언어를 적는다. 똑같은 행을 두 번 쓰면 읽는다 | 앞선 자식 SPEC 들의 D-5·D-6 과 같은 규칙이에요: 순서로만 정해지는 답은 사용자가 보는 목록과 적용되는 목록을 달라지게 해요. 기본: 거절 | 앞 행이 이기게 해요 |
| D-7 | 배송 언어만 가진 프로젝트 행은 버려진 것을 답이 말하지 않는다(`read`, `rows` 는 파일의 행 수, `problem` null). 앞선 자식 SPEC 들의 같은 결정과 같은 이유로, 알림이 필요하면 manifest·확장자·의존성·서버 네 종류를 함께 후속으로 정해요 | `manifest_catalog` 의 세 키는 부모 SPEC 의 테스트가 객체 통째로 고정해서 키를 더하면 그 계약이 바뀌어요. 기본: 이대로 | `manifest_catalog` 에 무시된 행 수를 더하고 부모의 테스트를 고쳐요 |
| D-8 | 한 플러그인 이름에 서로 다른 실행 파일이 가면(확장 파일의 두 행끼리, 또는 프로젝트 행과 배송 행 사이) 파일 통째로 거절하고 `problem` 에 그 플러그인을 적는다. 같은 플러그인·같은 실행 파일을 여러 언어가 쓰는 것은 충돌이 아니다 | `lsp_plugins` 는 플러그인마다 한 행이고 실행 파일은 그 플러그인에 닿은 첫 언어(정렬 순)의 것이라, 두 실행 파일이면 순서로 답이 정해져요. D-6 의 근거와 같아요. 배송 표에서 `typescript-lsp` 와 `clangd-lsp` 는 두 언어가 같은 실행 파일을 써서 충돌이 아니에요. 기본: 거절 | 첫 언어의 실행 파일이 이기게 하는 지금 동작으로 되돌려요 |

## Result

기준 팁 NeXT `5a6aa857d3`(한 번 머지, 머지 커밋 `ecc4b813d4`). SPEC 초안 커밋 `4d30f6fba6`, 의존성 조각 머리 `127009ab31` 를 들여온 머지 `db1610152e`, 계획 점검 뒤 다듬은 커밋 `a2b04e3c08`, RED 커밋 `956c415442`, GREEN 커밋 `9ca3a6425f`. 개정은 1 그대로예요.

- 작업 카드가 옮겨졌어요: 이 조각은 처음 카드의 기준 4 로 시작했고, 그 카드가 `--keep-open` 없이 불린 착지로 닫힌 뒤(리더 정정 노트 9733) 남은 일을 후속 카드의 기준 1 로 이어요. Plan 의 "카드의 기준 4" 는 지금 이 후속 카드의 기준 1 이에요.
- 승인: 위임 승인이 노트 9612 에 있어요(사용자의 직접 선택이 아니라 위임에 따른 리더 판단). 노트의 `approval_digest` `sha256:96acd8b1…f60d` 는 `ecc4b813d4` 에서 `moai spec show` 가 낸 값과 같아요. frontmatter 는 `status: draft` 이고 `approval` 키가 없어요.
- 계획 점검 뒤 본문에 더한 것: `plugin` 의 `@` 거절, 한 플러그인 두 실행 파일 거절, AC-7 의 `lsp_plugins` 이웃, 값마다의 경계, `src/project_scale.rs` 의 별개 `PLUGINS` 표를 후속으로 적음.
- Run 1 (RED, `956c415442` 의 테스트만; `cargo nextest ... --test project_detect_monorepo_cli --lib` 필터 실행; `target/kenan/red.log`): `37 tests run: 30 passed, 7 failed`. 새 테스트 일곱이 옳은 이유로 실패했어요 — 확장 파일의 키 `language_servers` 를 몰라서 `refused`(`the file is not a valid catalog`)이거나, 단위 테스트는 배송 파일에 그 키가 없어서 멈췄어요. 현재 동작을 지키는 `every_shipped_language_server_keeps_its_plugin` 은 통과했어요.
- Run 2 (GREEN, 한 슬롯의 `nextest`: `project_detect_monorepo_cli`, `project_detect_cli`, `project_lsp_plugins_cli`, `green_test_lock`, `project_scale_detect_lib`, `path_root_direct_sites`, `home_isolation_scan`, `--lib project_detect`; `target/kenan/green.log`): `122 tests run: 122 passed, 2025 skipped`. AC 의 `verify:` 줄에 적힌 `cargo test` 형태 그대로는 돌리지 않았어요. 로그는 GREEN 커밋보다 먼저 쓰였고, 머지 커밋 `ecc4b813d4` 위에서는 다시 돌리지 않았어요(`git diff 9ca3a6425f ecc4b813d4` 에서 `src/project_detect.rs`, 카탈로그 파일, 위 통합 테스트 파일들은 차이가 없어요).
- `cargo clippy ... --lib --test project_detect_monorepo_cli -- -D warnings`: 종료 0(`target/kenan/clippy.log`). `cargo fmt --all -- --check`: 머지 커밋 `ecc4b813d4` 에서 종료 0.
- AC-1: met — `every_shipped_language_server_keeps_its_plugin` 과 단위 테스트 `the_shipped_catalog_holds_the_earlier_language_servers` 통과(Run 2).
- AC-2: met — `a_language_server_row_adds_a_server_without_code` 통과(Run 2; Run 1 에서는 실패).
- AC-3: met — `a_language_server_row_never_replaces_a_shipped_server` 통과(Run 2; Run 1 에서는 실패).
- AC-4: met — `a_language_server_row_that_cannot_be_read_refuses_the_whole_file` 통과(Run 2; Run 1 에서는 실패). Run 1 은 대조에서 멈춰서, 나쁜 행의 거절은 Run 2 와 가드 뮤턴트가 근거예요.
- AC-5: met — `the_language_server_row_limits_hold_at_both_edges` 통과(Run 2; Run 1 에서는 실패).
- AC-6: met — `one_language_for_two_servers_in_the_extension_file_is_refused` 통과(Run 2; Run 1 에서는 실패).
- AC-7: met — 이웃 테스트가 Run 2 에서 모두 통과했어요. READY 영수증(`moai check affected --card`)은 이 Result 커밋 뒤에 돌리고, 여기가 아니라 카드의 READY 노트에 인용해요.
- AC-8: met — `one_plugin_with_two_executables_is_refused` 통과(Run 2; Run 1 에서는 실패).
- 가드 뮤턴트(tier 1 조각, 디스패치가 줄을 요청했어요; `target/kenan/guard-mutants.txt` 의 여덟 줄 그대로):
  - guard-mutant: the_shipped_catalog_holds_the_earlier_language_servers | shipped cpp row removed | exit 100
  - guard-mutant: every_shipped_language_server_keeps_its_plugin | shipped python executable changed | exit 100
  - guard-mutant: a_language_server_row_adds_a_server_without_code | the project rows are not appended to the catalog | exit 100
  - guard-mutant: a_language_server_row_never_replaces_a_shipped_server | lookup takes the last row instead of the first | exit 100
  - guard-mutant: a_language_server_row_that_cannot_be_read_refuses_the_whole_file | deleted deny_unknown_fields from ServerRow (unknown key in a row accepted) | exit 100
  - guard-mutant: the_language_server_row_limits_hold_at_both_edges | row count limit dropped | exit 100
  - guard-mutant: one_language_for_two_servers_in_the_extension_file_is_refused | a different server for one language is no conflict | exit 100
  - guard-mutant: one_plugin_with_two_executables_is_refused | a project plugin clashing with a shipped plugin is accepted | exit 100
- 뮤턴트 실행 경과: 처음의 29 개짜리 스크립트는 15 개를 처리한 시점에 중단했어요(14 개 죽음, 종료 100; m07 은 패턴이 두 번 나와 건너뜀). bash 3.2 의 큰 문자열 치환이 뮤턴트마다 몇 분씩 걸렸기 때문이에요(`target/kenan/mutants.out`). 줄이 없던 나머지 테스트에는 `sed` 뮤턴트를 하나씩 줬어요(`target/kenan/mutants-rest.out`; m29 는 패턴이 두 줄이라 건너뛰어서 패턴을 좁혀 다시 돌렸어요, `target/kenan/mutants-m29.out`).
- 이웃 검색(`git grep`): 기준 팁 `5a6aa857d3` 의 `src`, `tests`, `skills`, `docs`, `agents`, `tools` 에서 `LSP_PLUGINS` 는 `src/project_detect.rs` 에만 있었고 `ecc4b813d4` 에는 한 곳도 없어요. `lsp_plugins` 사실의 모양은 `tests/project_lsp_plugins_cli.rs` 가 지켜요(고치지 않았고 통과). `tests/path_root_direct_sites.rs` 의 기준선은 그대로예요(통과). `manifest_catalog` 가 나오는 파일은 `src`, `tests`, `skills`, `docs` 중 `src/project_detect.rs` 와 `tests/project_detect_monorepo_cli.rs` 뿐이에요. 서버 이름은 `skills/moai/workflows/project.md`, `skills/moai/references/languages/python.md` 와 몇몇 테스트·픽스처에도 있고, 이 변경은 그 파일들 가운데 `tests/project_detect_monorepo_cli.rs` 만 고쳤어요.
- 위험 경로: 읽기 전용 감지라서 해당 없어요.
- 미충족/한계: `project-scale-plan` 은 확장 파일로 더한 언어를 아직 거절해요(`src/project_scale.rs` 의 `PLUGINS`; 리더 결정 노트 9687 의 (가): 다음 조각의 새 SPEC 으로 표를 하나로 만들어요).
- 미충족/한계: 프로젝트가 더한 서버는 `skills/moai/workflows/project.md` 의 서버별 설치 표에 행이 없어요. `skills` 와 `README*` 에서 `language_servers` 와 `detect-catalog` 를 찾아도 설명하는 글이 없어요.
- 미충족/한계: 마켓플레이스 이름은 Non-goals 의 결정대로 엔진에 남아요(`LSP_MARKETPLACE`; Decisions 표에는 이것만의 번호가 없고 D-5 가 고정된 이름을 전제해요).
- 미충족/한계: Windows·Linux 는 돌리지 않았어요. 전체 `cargo test` 는 돌리지 않았어요. 이전 판 바이너리가 새 키가 든 확장 파일을 거절하는지도 돌려 보지 않았어요.
