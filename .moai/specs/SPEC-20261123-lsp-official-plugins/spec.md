---
id: SPEC-20261123-lsp-official-plugins
title: "LSP only through the official Claude Code plugins, and /moai project recommends official plugins"
status: draft
revision: 1
created: 2026-11-23
---

# LSP 는 Claude Code 공식 플러그인만, `/moai project` 는 공식 플러그인을 추천

## Intent

사용자 지시(2026-11-23)예요.

- 모든 LSP(코드 탐색 서버)는 `rust-analyzer-lsp` 처럼 Claude Code 공식 플러그인만 써요. Homebrew 로 서버를 깔지 않아요.
- LSP 가 깔려 있지 않은 프로젝트는 필요한 때 설치 안내를 하고, 사용자가 고른 뒤 설치하고 이어가요. 이 규칙은 지침에 적어요.
- 개발에 쓸 만한 공식 플러그인이 있으면 `/moai project` 에서 사용자에게 안내하고, 고른 것만 추가해요.

### Sources (read 2026-11-23, 이 기계의 공식 마켓 복사본 `claude-plugins-official`)

- LSP 플러그인 12개: `clangd-lsp`, `csharp-lsp`, `gopls-lsp`, `jdtls-lsp`, `kotlin-lsp`, `lua-lsp`, `php-lsp`, `pyright-lsp`, `ruby-lsp`, `rust-analyzer-lsp`, `swift-lsp`, `typescript-lsp`.
- 각 플러그인의 `.claude-plugin/plugin.json` `lspServers` 가 PATH 에서 부르는 실행 파일: `clangd`, `csharp-ls`, `gopls`, `jdtls`, `kotlin-lsp`, `lua-language-server`, `intelephense`, `pyright-langserver`, `ruby-lsp`, `rust-analyzer`, `sourcekit-lsp`, `typescript-language-server`.
- 각 README 의 설치 경로: Rust 는 `rustup component add rust-analyzer`(권장), Go 는 `go install golang.org/x/tools/gopls@latest`, Python 은 `pipx install pyright`·`pip install pyright`·`npm install -g pyright`, TypeScript 는 `npm install -g typescript-language-server typescript`, PHP 는 `npm install -g intelephense`, Ruby 는 `gem install ruby-lsp`, C# 은 `dotnet tool install --global csharp-ls`. 나머지(`clangd`, `jdtls`, `kotlin-lsp`, `lua-language-server`, `sourcekit-lsp`)의 README 는 Homebrew 나 시스템 패키지, 수동 내려받기를 적어요.
- 추천 후보로 쓰는 공식 플러그인 이름 `github`, `playwright`, `security-guidance` 는 같은 마켓의 `marketplace.json` 에 있어요.

## Non-goals

- 프로젝트 쪽에서 플러그인을 켜는 일, `~/.claude/settings.json` 을 고치는 일은 하지 않아요. 설치는 `claude plugin install` 이 해요.
- 플러그인이 켜져 있는지(enabled)는 `project-detect` 가 읽지 않아요. HOME 을 읽지 않는다는 규칙이 그대로예요. 세션이 `claude plugin list` 로 읽어요.
- 마켓 목록을 네트워크로 새로 받지 않아요.

## Risks

- 공식 README 가 Homebrew 만 적는 서버(`clangd`, `jdtls`, `kotlin-lsp`, `lua-language-server`, `sourcekit-lsp`)는 규칙 2 와 부딪혀요. 세션이 설치하지 않고 사용자에게 선택지로 보여 주고, 사용자가 고른 경로를 따라요.
- `lsp_plugins` 의 PATH 확인은 HOME 밖의 폴더 이름을 읽어요. 파일 이름만 보고 내용은 읽지 않으며, 프로젝트 파일(`project.yaml`)에는 싣지 않아요.
- `detected.lsp` 의 Kotlin·C# 서버 이름이 바뀌어서(`kotlin-language-server` → `kotlin-lsp`, `omnisharp` → `csharp-ls`) 이 두 언어의 기존 프로필은 다음 `project-plan` 에서 변경으로 보여요.

## Requirements

- REQ-1: The core rules shall require that code navigation uses only the official Claude Code LSP plugins named `<language>-lsp@claude-plugins-official` · failure: none — a rule text
- REQ-2: The core rules shall forbid installing a language server with Homebrew and name the language-tool route that the plugin's README lists (rustup, npm, pip or pipx, go install, gem, dotnet tool) as the way · failure: none — a rule text; a server whose README lists no such route is reported to the user as a choice and not installed by the session
- REQ-3: When a language the work needs has no official LSP plugin installed or no server on PATH, the session shall show the install commands as a selection before it installs anything · failure: none — the REQ-4 covers the unpicked or failed install
- REQ-4 (REQ-3): IF the user does not pick the install or the install fails, THEN the session shall go on with its search and read tools and say in its report that no LSP was used
- REQ-5: The core rules shall make the install of a plugin or a server the user's selection (a `guarded` setting) that ends with the session continuing after `/reload-plugins` has run · failure: none — a rule text
- REQ-6: The PROJECT workflow shall hold a step that shows, for each detected language, its official LSP plugin and whether its server is installed · failure: none — a step text
- REQ-7: The PROJECT workflow shall recommend official plugins only from facts found in the same run, with one reason line each · failure: none — a step text; a recommendation without a fact found in the run is not made
- REQ-8: The PROJECT workflow shall install only the plugins the user picked, with `claude plugin install <name>@claude-plugins-official`, and read the result back · failure: none — a step text; an unpicked or unanswered question installs nothing
- REQ-9: When `project-detect` runs, it shall report `lsp_plugins` beside `detected`, one row per official plugin the detected languages call for, with the plugin name, its marketplace, the languages, the server executable and `installed` · failure: none — a project with no detected language reports an empty list
- REQ-10 (REQ-9): IF PATH is unset or one of its folders cannot be read, THEN `project-detect` shall report `installed` false for the executables it did not find
- REQ-11: The `installed` check shall only look at the file names of the PATH folders: it writes nothing, starts no process and uses no network · failure: none — a read-only check
- REQ-12: The `detected.lsp` list shall name the server of the official plugin of each detected language · failure: none — the table is one table that REQ-9 reads too

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): `core-rules.md` 의 새 규칙이 공식 플러그인만, Homebrew 금지와 언어 도구 경로, 없을 때 선택 뒤 설치, 거절·실패 때 LSP 없이 이어감, 사용자 선택(guarded), `/reload-plugins` 뒤 이어감을 담고, 고정 번호 규칙(3, 7, 19–21)은 그대로예요 · verify: `cargo test --offline --locked --test project_lsp_plugins_cli core_rules_state_the_official_lsp_plugin_rule`
- AC-2 (REQ-6, REQ-7, REQ-8): `project.md` 가 LSP 단계(`lsp_plugins`·`claude plugin list`)와 근거 있는 추천 단계(이유 한 줄, 고른 것만 `claude plugin install <name>@claude-plugins-official`, 다시 읽어 확인)를 담고, 감지는 HOME 을 읽지 않는다는 문장에 PATH 의 파일 이름만 본다는 예외를 적어요 · verify: `cargo test --offline --locked --test project_lsp_plugins_cli project_workflow_has_the_lsp_and_recommendation_steps`
- AC-3 (REQ-9, REQ-12): 임시 폴더의 Rust·TypeScript·Kotlin 프로젝트에서 `lsp_plugins` 가 플러그인별 한 줄(`rust-analyzer-lsp`, `typescript-lsp` 에 두 언어, `kotlin-lsp`)이고, 서버가 PATH 임시 폴더에 있는 것만 `installed` true 이고, `detected.lsp` 가 공식 플러그인의 서버 이름이에요 · verify: `cargo test --offline --locked --test project_lsp_plugins_cli lsp_plugins_name_the_official_plugin_and_the_path_state`
- AC-4 (REQ-10): PATH 가 비었거나 없는 폴더뿐이거나 PATH 환경 변수가 없어도 `project-detect` 가 성공하고 모든 `installed` 가 false 예요 · verify: `cargo test --offline --locked --test project_lsp_plugins_cli an_unreadable_or_missing_path_reads_as_not_installed`
- AC-5 (REQ-11): `project-detect` 전후로 프로젝트 폴더와 PATH 임시 폴더의 파일과 내용이 같고, PATH 의 실행 파일은 실행되지 않아요(실행하면 표식 파일을 남기는 가짜 서버로 확인) · verify: `cargo test --offline --locked --test project_lsp_plugins_cli the_path_check_writes_nothing_and_starts_nothing`
- AC-6 (REQ-1, REQ-6, REQ-9, REQ-12): 바뀐 글과 코드를 읽는 기존 테스트가 그대로 초록이에요 · verify: `cargo test --offline --locked --test project_detect_cli --test core_rules_2876`
- AC-7 (REQ-1, REQ-6): 스킬 패키지 파일이 바뀌어서 소스 매니페스트의 sha256 이 맞고 패키지 테스트가 그대로 초록이에요 · verify: `cargo test --offline --locked --test product_cli skill_package::`

## Scope

- write: `.moai/specs/SPEC-20261123-lsp-official-plugins/**`
- write: `skills/moai/references/internal/core-rules.md`
- write: `skills/moai/workflows/project.md`
- write: `skills/moai/references/source-manifest.json`
- write: `src/project_detect.rs`
- write: `tests/project_lsp_plugins_cli.rs`
- write: `tests/core_rules_2876.rs` (only the `CORE_RULES_MAX_BYTES` cap, which the new rule makes too small)
- write: `tests/agent_defs_package.rs` (only the pinned PATH sentence of `workflows/project.md`, which this SPEC rewrites)
- write: `tests/repo_scans_table.rs` (only one `EXEMPT` row for the new test file)

## Plan

1. 이 SPEC.
2. 테스트 먼저(RED): `tests/project_lsp_plugins_cli.rs` 의 AC-1 ~ AC-5.
3. `src/project_detect.rs`: `LSP` 표를 `(언어, 서버, 플러그인, 실행 파일)` 표로 바꾸고 `lsp_plugins` 사실과 PATH 파일 이름 확인을 더해요. `.lua` 확장자를 `lua` 로 읽어요.
4. `core-rules.md` 에 규칙 25 를 더해요(기존 번호는 옮기지 않아요).
5. `project.md` 에 LSP 단계와 추천 단계를 더하고, 감지 설명의 `lsp` 문장을 고쳐요.
6. 소스 매니페스트의 sha256 을 새로 써요.
7. 점검: 위 verify 명령과 `cargo clippy`, `cargo fmt --check`, `git diff --check`.

## Result

- AC-1: met — `cargo test --offline --locked --test project_lsp_plugins_cli` ran 5 tests, all passed on this tree (RED first: 5 failed before the change).
- AC-2: met — same run, `project_workflow_has_the_lsp_and_recommendation_steps` passed.
- AC-3: met — same run, `lsp_plugins_name_the_official_plugin_and_the_path_state` passed.
- AC-4: met — same run, `an_unreadable_or_missing_path_reads_as_not_installed` passed.
- AC-5: met — same run, `the_path_check_writes_nothing_and_starts_nothing` passed.
- AC-6: met — `cargo test --offline --locked --test project_detect_cli --test core_rules_2876` exit 0 (28 and 6 tests passed), and `cargo test --offline --locked --lib project_detect` exit 0 (7 passed).
- AC-7: met — `cargo test --offline --locked --test product_cli skill_package::` exit 0 (13 passed) after `python3 tools/source_manifest.py write`.

`check affected` on the first head found three pins that this change moved (`agent_defs_package` read the old PATH sentence, `home_isolation_scan` wanted the new test to run through `TempHome`, `repo_scans_table` wanted its `EXEMPT` row); they were repaired in a second commit and the four test targets pass again.

Repair after the first audit (FAIL, blocker B1): the fake server of `the_path_check_writes_nothing_and_starts_nothing` wrote its marker with `touch`, which the test's PATH (one temporary folder) could not find (exit 127), so the guard could not fail. It now writes the marker with the shell's `: >`, and `a_fake_server_leaves_its_marker_without_a_path` runs it with an empty environment to pin that. Mutant: `executable_on_path` made to run each found server with `--version` — `the_path_check_writes_nothing_and_starts_nothing` failed with "the fake server was started" (the other five passed); the file was restored with `git restore`. Also changed: a doc line on relative PATH entries in `src/project_detect.rs`, the `kotlin-lsp` sentence of project.md step 6a (its README names only Homebrew, so a manual install route is given), the stale 4096-byte wording of the core-rules entry in `source-manifest.json` (cap now 5632), and the `workflows/project.md` sha256 (`tools/source_manifest.py write`). The core-rules.md byte cap has 27 bytes left; the next rule has to raise it again (recorded, not changed).

Also run: `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, `cargo fmt` clean, `tools/smoke/run.sh project` printed `smoke: project pass steps=3`. Not run: Windows build (`cargo check --target x86_64-pc-windows-msvc`); the `executable_on_path` windows branch is read only by the `windows-build` run after landing.
