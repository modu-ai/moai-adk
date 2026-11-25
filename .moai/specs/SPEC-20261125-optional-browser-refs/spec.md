---
id: SPEC-20261125-optional-browser-refs
title: "The bundled Aside and ego(lite) references follow the installed official guides and never need the tool"
status: draft
revision: 1
created: 2026-11-25
---

# The bundled Aside and ego(lite) references follow the installed official guides and never need the tool

## Intent

`ops-tools` 번들 팩의 선택 도구 레퍼런스(Aside, ego(lite) 브라우저)는 도구가 없어도 설치와 작업이 평소대로 돼야 하고, 도구가 있을 때는 설치된 공식 안내를 따라야 해요. 지금 번들의 `moai-ref-aside-browser` 는 `aside mcp` 로 MCP 서버를 등록하라는 안내(`moai mcp add aside …`)와 옛 JavaScript 호출 설명을 본문에 두고, ego(lite) 용 레퍼런스는 팩에 없어요. 그래서 이 SPEC 은 두 글을 같은 모양으로 맞춰요: Aside 글은 `aside` 명령줄만 쓰고, 새 ego(lite) 글을 더하고, 두 글 모두 "언제 쓰는가, 어디까지 허용하는가, 도구가 없을 때 안전하게 멈추는 법"만 말해요. 도구의 사용법은 설치된 공식 스킬과 CLI 안내를 가리키기만 해요.

도구 사실은 2026-11-25 15:19Z 에 조사 에이전트가 이 Mac 에서 읽기 전용 명령으로 읽었어요. ego(lite) 쪽은 `ego-browser --version` 이 0.6.3.17(chromium 156.0.8078.96), `ego-browser skill` 이 "ego-browser usage guide v2.2.1 (macos)" 이고, Aside 쪽은 `aside --version` 이 1.26.1121.1917, `aside guide` 가 "Aside CLI 1.26.1121.1917 · Skill version 5"(4593 바이트)예요. 같은 도움말과 안내 글을 계획 에이전트가 같은 날 늦게 한 번 더 읽어서 같은 내용을 확인했어요(정확한 시각은 기록하지 않았어요). 두 안내가 말하지 않는 줄(사용자가 넘겨받거나 권한이 거절되면 멈춤, 탭 재사용 금지, `ego-browser import` 금지, 업그레이드 금지, 완료 배지 규칙)은 MoAI 자신의 규칙이고, 글이 그렇게 밝혀요.

바뀌는 것: Aside 글(MCP 등록 안내 제거, 안내 가리키기, 명령 집합 고정), 새 ego(lite) 글, `ops-tools` 팩 데이터 한 항목, 이를 고정하는 새 시험 파일. 바뀌지 않는 것: `skills/moai` 의 워크플로 글과 어댑터, 사용자의 도구 설정과 MCP 설정, 도구 설치.

### Sources

읽기 전용 명령이고 읽은 곳은 이 Mac 이에요. 출력은 명령의 표준 출력이라 체크섬은 기록하지 않았어요.

| Source | What was read | Read at | Result used here |
| --- | --- | --- | --- |
| `ego-browser --version` | 도구와 브라우저 버전 | 2026-11-25 15:19Z, 조사 에이전트 | 0.6.3.17, chromium 156.0.8078.96 |
| `ego-browser skill` | 설치된 사용 안내 | 2026-11-25 15:19Z, 조사 에이전트 | "ego-browser usage guide v2.2.1 (macos)": 한 목표 한 공간, 프로필은 요청할 때만, 막힌 페이지에서 새 공간으로 우회 금지, 사용자가 넘겨받으면 멈춤, 로그인·CAPTCHA·2차 인증은 사용자 몫, 업그레이드는 사용자 승인 뒤 |
| `ego-browser help` | 명령 목록 | 2026-11-25 15:19Z, 조사 에이전트 | 명령 이름 확인. 안내가 `ego-browser import` 를 말하지 않아서 금지는 MoAI 규칙이에요 |
| `aside --version` | 도구 버전 | 2026-11-25 15:19Z, 조사 에이전트 | 1.26.1121.1917 |
| `aside --help` | 옵션과 하위 명령 | 2026-11-25 15:19Z, 조사 에이전트 | `--permission <MODE>`: ask 와 guard 는 같은 모드(기본 Guard), full-access. "읽기 전용" 모드는 없어요. `aside repl --help` 에는 `--permission` 이 없어요 |
| `aside guide` | 설치된 CLI 안내 | 2026-11-25 15:19Z, 조사 에이전트 | "Aside CLI 1.26.1121.1917 · Skill version 5", 4593 바이트: `aside guide repl` 먼저 읽기, `aside skills show <name>` 먼저 읽기, `aside memory` 는 읽기만, `aside session delete` 전에 묻기. 멈춤 규칙, 탭 재사용 규칙, CLI 설치 명령은 말하지 않아요 |
| `aside skills list` | 설치된 Aside 스킬 | 2026-11-25 15:19Z, 조사 에이전트 | 목록 명령이 있어요 |
| 위 도움말과 안내 글 다시 읽기 | 같은 글 | 같은 날 늦게, 계획 에이전트(정확한 시각은 기록하지 않았어요) | 위 내용과 같아요 |
| 카드 본문의 확인 규칙 | Interview 의 답 | 사용자 확인 2026-11-21, 다른 세션이 전달 | 정본은 설치된 공식 스킬·CLI 의 안내예요. 승인으로 기록하지 않아요 |

## Interview

- Q1: 번들되는 선택 도구(Aside, ego(lite)) 레퍼런스에서 도구 사용법의 정본은 무엇이고 MoAI 글은 무엇을 말하나요? → A: 각 도구의 정본은 설치된 공식 스킬·CLI 가 말하는 현재 버전의 안내예요. MoAI 글은 언제 쓰는가, 어디까지 허용하는가, 도구가 없을 때의 안전한 멈춤만 말해요(사용자 확인 2026-11-21, 카드 본문, 다른 세션이 전달).
- Q2: MoAI 글이 도구나 사용자 설정을 건드려도 되나요? → A: 아니요. 도구와 공식 스킬을 설치하지 않고, 사용자의 도구 설정과 MCP 설정을 바꾸지 않아요(같은 확인, 카드 본문).

## Non-goals

- 공식 스킬이나 안내의 본문, 플래그 목록, API 목록을 복제하지 않아요.
- 도구나 공식 스킬을 설치하거나 업그레이드하지 않아요. 사용자의 도구 설정, 개인 MCP 설정, 브라우저 프로필과 쿠키 가져오기를 건드리지 않아요.
- 새 MCP 서버, 새 Rust 하위 명령, 훅 변경을 만들지 않아요.
- `skills/moai/workflows/run.md`, `skills/moai/workflows/audit.md`, `skills/moai/references/adapters/aside.md` 는 바꾸지 않아요.
- 도구의 멈춤·대기·끝남 배지를 카드의 완료로 보지 않아요(완료는 원장, 독립 감사, 로컬 병합이에요).
- 도구가 설치된 상태에서의 실제 브라우저 조작 확인은 이 SPEC 의 시험이 아니에요.

## Requirements

- REQ-1: The bundled Aside reference shall name the `aside` command line as its only way in and give no MCP registration advice.
- REQ-2: Each bundled reference shall send the reader to the installed tool's own guide and official skill instead of copying its flags or interface.
- REQ-3: Each bundled reference shall name exactly the tool commands that its read-only, probe and forbidden lines need, and no other.
- REQ-4: Each bundled reference shall mark the lines that follow the installed official guide apart from the lines that are MoAI's own rules.
- REQ-5: The bundled ego(lite) reference shall keep the use conditions of the installed official guide: several spaces, a profile or another user's space only when the operator explicitly asks, and a stop when the operator takes control.
- REQ-6: Each bundled reference shall limit its work to the session or space this task opened and never attach to, steer or close one it did not open.
- REQ-7: Each bundled reference shall forbid installing or upgrading the tool and changing the user's tool or MCP settings, and name a stopped, idle or done badge as no completion of the card.
- REQ-8: The bundled ego(lite) reference shall forbid importing browser data.
- REQ-9: Each bundled reference shall tell the reader, when the tool is missing, offline, refused a permission or failing a probe, to record the step as `not_run` with the reason and go on without the tool or another browser.
- REQ-10: The `ops-tools` pack data shall list both references as skill entries, the ego(lite) one beside the Aside one, each with the SHA-256 of its shipped bytes.
- REQ-11: The `moai bundle add ops-tools` command shall install both references, and write no tool or MCP setting, when neither tool is installed.

## Acceptance

- AC-1 (REQ-1, REQ-2): the Aside reference names `aside guide`, says it is CLI only, points to the official `aside-browser` skill, and carries no MCP registration line, no `/moai e2e` relation and no Playwright-style call line · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_aside_reference_is_cli_only_and_points_to_the_installed_guide`
- AC-2 (REQ-3): the Aside reference names exactly the seven read-only guide commands, the `aside --version` probe, `aside exec` and the five forbidden commands, and each forbidden-command line says never · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_aside_reference_names_exactly_the_commands_its_safe_lines_allow`
- AC-3 (REQ-4, REQ-6): the Aside guide section holds only lines the installed guide states, and the section of MoAI's own rules says so and keeps the stop, session, tab and permission rules · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_aside_reference_marks_the_guides_lines_apart_from_its_own_rules`
- AC-4 (REQ-7, REQ-9): the Aside reference records the step as `not_run`, treats Aside as absent without switching browsers, forbids installing, upgrading and changing settings, and calls the badge no completion · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_aside_reference_stops_safely_when_the_tool_is_missing`
- AC-5 (REQ-2): the ego(lite) reference names `ego-browser skill` and the official `ego-browser` skill, has the `When to use` and `How far it may go` headings, carries no official API call line and stays within 120 lines · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ego_reference_points_to_the_installed_guide_without_copying_it`
- AC-6 (REQ-3): the ego(lite) reference names exactly the three read-only guide commands, `ego-browser upgrade` and `ego-browser import`, and the last two lines say never · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ego_reference_names_exactly_the_commands_its_safe_lines_allow`
- AC-7 (REQ-4, REQ-5, REQ-6): the ego(lite) guide section keeps the official use conditions and the section of MoAI's own rules says so, keeps the never-upgrade rule as stricter than the guide and keeps the operator's tabs protected · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ego_reference_keeps_the_official_conditions_and_its_own_rules_apart`
- AC-8 (REQ-7, REQ-8, REQ-9): the ego(lite) reference records the step as `not_run`, treats ego(lite) as absent without switching browsers, forbids installing, upgrading, importing and changing settings, and calls the badge no completion · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ego_reference_stops_safely_when_the_tool_is_missing`
- AC-9 (REQ-3): the command set comparison reports a reference that names one extra command and one that rewords a command as different from the allowed set · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_command_set_comparison_catches_an_extra_and_a_reworded_command`
- AC-10 (REQ-10): the `ops-tools` pack lists both references once each as skill entries with the right root, path and empty skill dependencies, the ego(lite) entry right after the Aside one, a `sha256` equal to the SHA-256 of each shipped `SKILL.md`, and a front matter name equal to the entry name · verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ops_tools_pack_lists_both_references_with_matching_names_and_hashes`
- AC-11 (REQ-11): `moai bundle add ops-tools` on a fresh temporary HOME with no browser tool on PATH succeeds, installs both files byte-equal to the shipped ones and creates no tool or MCP setting file · verify: `cargo test --locked --offline --test optional_browser_refs_cli adding_ops_tools_installs_both_references_with_no_browser_tool_present`
- AC-12: the neighbouring bundle install tests keep passing · verify: `cargo test --locked --offline --test bundle_user_assets_cli`
- AC-13: the new test file, which reads shipped files, is covered by the repo scan registry · verify: `cargo test --locked --offline --test repo_scans_table every_file_reader_has_an_always_scan_or_an_exemption`

## Scope

- write: `.moai/specs/SPEC-20261125-optional-browser-refs/spec.md`
- write: `tests/optional_browser_refs_cli.rs`
- write: `templates/bundle-packs/.claude/skills/moai-ref-aside-browser/SKILL.md`
- write: `templates/bundle-packs/.claude/skills/moai-ref-ego-browser/SKILL.md`
- write: `templates/bundle-packs/packs.json`
- write: `tools/repo_scans.json`
- read-only: `skills/moai/references/adapters/aside.md`
- read-only: `skills/moai/workflows/run.md`
- read-only: `skills/moai/workflows/audit.md`
- read-only: `tests/launcher_reads_models.rs`
- forbidden: `skills/moai/**`

## Plan

1. SPEC 을 채우고 `moai spec check SPEC-20261125-optional-browser-refs` 를 돌려요.
2. RED: `tests/optional_browser_refs_cli.rs`(시험 11 개, AC-1..AC-11 하나씩)를 쓰고 `cargo fmt --all` 을 돌려요. 기대하는 RED 는 10 개 실패, 1 개 통과(AC-9 는 순수 픽스처라 처음부터 통과해요)이고, 컴파일 오류는 RED 가 아니에요. 출력은 `target/optional-browser-refs/red.out` 에 저장해요.
3. `tools/repo_scans.json` 의 `reads-shipped-files` 검사에 새 시험 파일을 늘리기만 하는 두 줄을 더해요(`moai card line-lease` 로, `git diff --numstat` 가 `2 1`).
4. Aside `SKILL.md` 를 정리해요(110 줄 이하).
5. ego(lite) `SKILL.md` 를 새로 만들어요(120 줄 이하, `.claude/` 가 `.gitignore` 에 걸려서 `git add -f` 로 추가해요).
6. 마지막에 `packs.json`: Aside 항목의 `sha256` 을 바꾸고 `ops-tools` 에서 그 바로 뒤에 `moai-ref-ego-browser` 항목을 같은 모양으로 더해요(`git diff --numstat` 가 `9 1`). 해시는 `shasum -a 256` 으로 구해요.
7. 새 시험, `bundle_user_assets_cli`, `repo_scans_table` 을 돌리고 구현이 끝난 뒤 moai-docs 가 `## Result` 를 써요. 이 단계는 구현 에이전트의 일이 아니에요.

## Risks

- `packs.json` 의 `sha256` 열은 코드가 읽지 않아서 틀려도 설치가 깨지지 않아요. 그래서 AC-10 의 시험이 파일 바이트의 SHA-256 과 맞는지 직접 비교해요. 글을 고친 뒤 해시를 다시 구하지 않으면 이 시험이 실패해서 알아채요.
- 글에 도구 이름과 옵션을 적으면 도구 버전이 바뀔 때 낡아요. 그래서 옵션 목록 대신 설치된 안내를 먼저 읽으라고 쓰고, 시험은 명령 이름의 집합만 고정해요. 도구가 바뀌어 안내의 줄이 달라지면 "공식 안내의 줄" 절이 낡을 수 있고, 이는 새 SPEC 으로 다시 읽어요.
- 시험이 "문장이 있다"만 보면 명령 하나를 더 적은 글도 통과해요. 명령 묶음을 집합으로 통째로 비교하고(AC-2, AC-6) 비교 함수 자체를 AC-9 의 픽스처가 고정해요.
- 새 시험 파일이 상태 폴더 이름이나 트리 순회 표지를 쓰면 저장소 전체 스캔이 걸려요. 시험은 상태 폴더 이름을 실행 시점에 조립하고 순회 표지를 쓰지 않아요.
- 새 ego(lite) `SKILL.md` 가 `.gitignore` 의 `.claude/` 에 걸려 커밋에서 빠질 수 있어요. `git add -f` 로 추가하고 `git status` 로 확인해요.
- `tools/repo_scans.json` 은 공유 파일이에요. 줄 임대로만 쓰고 커밋 직후에 돌려줘요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Aside 글에서 MCP 등록 안내를 지우고 `aside` 명령줄만 쓴다고 적어요 | 카드 본문의 확인 규칙(사용자 확인 2026-11-21)과 저장소 어댑터 `skills/moai/references/adapters/aside.md` 의 "CLI 만" 결정이에요. 사용자가 직접 넣은 MCP 항목은 그대로 둬요 | 글의 MCP 절을 되살려요 |
| D-2 | ego(lite) 글은 어댑터가 아니라 `ops-tools` 번들 팩에 둬요 | 선택 도구이고 Aside 글이 이미 같은 팩에 있어요. 어댑터와 `skills/moai` 는 이 SPEC 이 바꾸지 않아요 | `packs.json` 항목과 폴더를 빼요 |
| D-3 | 업그레이드는 공식 안내보다 엄격하게 "하지 않는다"로 적고, 글이 "엄격하다"고 밝혀요 | 카드 본문 규칙: 도구를 설치하거나 업그레이드하지 않아요. 설치된 `ego-browser skill` 은 사용자 승인 뒤 `ego-browser upgrade` 를 허용해서 그대로 두면 공식 문장과 어긋나 보여요 | 사용자 승인 뒤 허용하는 문장으로 바꿔요 |
| D-4 | 글이 이름 대는 명령 묶음을 시험이 집합으로 통째로 비교해요 | 문장 존재만 보는 시험은 명령 하나를 더한 변형을 통과시켜요(앞선 같은 주제 작업의 감사에서 확인했어요). 집합 비교는 하나를 더하거나 말을 바꾸면 실패해요 | 부분 문자열 시험으로 되돌려요 |
| D-5 | 금지 명령(`aside mcp`, `aside update`, `aside skills install`, `aside settings`, `aside host use`, `ego-browser import`)은 공식 안내보다 엄격하게 둬요 | 카드 본문 규칙: 공식 스킬 설치와 사용자 설정·MCP 설정 변경이 없어요. 두 안내 모두 이 명령들의 사용을 막지는 않아서 글이 MoAI 자신의 규칙이라고 밝혀요 | 금지 줄을 빼요 |
| D-6 | `tools/repo_scans.json` 은 시험 파일 이름과 검사 인자만 늘리는 두 줄로 줄 임대로 고쳐요 | 공유 파일은 범위에 넣지 않고 줄 임대로 쓰는 저장소 규칙이에요. 새 시험이 배포 파일을 읽어서 AC-13 의 등록이 필요해요 | 두 줄을 되돌려요 |

## Result

모든 명령은 브랜치 `spec/optional-browser-refs-20261125` 의 워크트리에서 이번 실행 중에 돌렸고, 출력은 `target/optional-browser-refs/` 아래에 있어요. 공식 안내를 읽은 때는 Intent 의 `### Sources` 표에 날짜와 함께 적혀 있고, 두 레퍼런스 각각의 세 줄은 MoAI 자신의 규칙이며 글이 그렇게 밝혀요.

RED 와 단계별 결과:

- RED(시험 파일만 커밋한 시점, `red.out`): `cargo test --locked --offline --test optional_browser_refs_cli` exit 101, 1 passed(AC-9 의 픽스처 시험), 10 failed. 실패 이유는 ego 파일 없음, 옛 Aside 글, 낡은 팩 해시로 모두 맞는 이유예요.
- Aside 글을 고친 뒤(`step4.out`): 5 passed, 6 failed. ego 글을 더한 뒤(`step5.out`): 9 passed, 2 failed(`packs.json` 이 필요한 팩 해시 시험과 설치 시험). `packs.json` 을 고친 뒤(`step6.out`): `optional_browser_refs_cli` 11 passed, 0 failed, exit 0. AC-1..AC-11 이 시험 하나씩이에요.

AC 별 결과:

- AC-1: met — `cargo test --locked --offline --test optional_browser_refs_cli` 한 번 실행(`gates.out`, `moai heavy-slot-acquire --name milo --max 3 --` 안에서 exit 0)에서 `the_aside_reference_is_cli_only_and_points_to_the_installed_guide` ok, 그 파일 11 passed, 0 failed.
- AC-2: met — 같은 실행에서 `the_aside_reference_names_exactly_the_commands_its_safe_lines_allow` ok.
- AC-3: met — 같은 실행에서 `the_aside_reference_marks_the_guides_lines_apart_from_its_own_rules` ok.
- AC-4: met — 같은 실행에서 `the_aside_reference_stops_safely_when_the_tool_is_missing` ok.
- AC-5: met — 같은 실행에서 `the_ego_reference_points_to_the_installed_guide_without_copying_it` ok.
- AC-6: met — 같은 실행에서 `the_ego_reference_names_exactly_the_commands_its_safe_lines_allow` ok.
- AC-7: met — 같은 실행에서 `the_ego_reference_keeps_the_official_conditions_and_its_own_rules_apart` ok.
- AC-8: met — 같은 실행에서 `the_ego_reference_stops_safely_when_the_tool_is_missing` ok.
- AC-9: met — 같은 실행에서 `the_command_set_comparison_catches_an_extra_and_a_reworded_command` ok. 이 시험은 RED 때도 통과한 픽스처 시험이에요.
- AC-10: met — 같은 실행에서 `the_ops_tools_pack_lists_both_references_with_matching_names_and_hashes` ok. `packs.json` 의 두 해시는 커밋된 `SKILL.md` 바이트의 `shasum -a 256` 값과 같아요.
- AC-11: met — 같은 실행에서 `adding_ops_tools_installs_both_references_with_no_browser_tool_present` ok.
- AC-12: met — 같은 묶음 실행(`optional_browser_refs_cli`, `bundle_user_assets_cli`, `home_isolation_scan`, `repo_scans_table`, `test_area_layout`, `ac_command_lint`, `state_folder_literal_scan`, `launcher_reads_models`, `gates.out`)에서 `bundle_user_assets_cli` 6 passed, 0 failed.
- AC-13: met — 같은 묶음 실행에서 `repo_scans_table` 33 passed, 0 failed, `every_file_reader_has_an_always_scan_or_an_exemption` ok.

같은 묶음 실행의 나머지: `home_isolation_scan` 15, `test_area_layout` 19, `ac_command_lint` 15, `state_folder_literal_scan` 10, `launcher_reads_models` 12 passed, 모두 0 failed, 전체 exit 0.

그 밖에 돌린 것:

- `cargo fmt --all` exit 0, 바뀐 것 없음. `git diff --check` exit 0.
- `moai trailer-check c5c0eda5259cb04d6dcc7a51acf7a94f62556a7d..HEAD` ok(커밋 6 개, 발견 0 건).
- `moai spec check --all`: 이 문서를 고치기 전에는 459 개 SPEC, 오류 0, 경고 50(기준선은 458 개, 오류 0, 경고 49 이고 하나 더 늘어난 것이 이 SPEC 의 `## Sources` 제목이었어요). `## Sources` 를 `## Intent` 아래 `### Sources` 로 옮긴 뒤에는 459 개 SPEC, 오류 0, 경고 49.
- 공유 파일 `tools/repo_scans.json` 은 `moai card line-lease` 로 정확히 두 줄만 늘렸어요(`git diff --numstat` 가 `2 1`: 규칙 한 줄, `--test` 옵션 하나). `packs.json` 의 `git diff --numstat` 는 `9 1` 이고 Aside 해시가 바뀌고 ego 항목이 더해졌어요.

Not run: `moai check affected --card`(READY 에서 돌리고 READY 노트에 인용해요), clippy 와 `product_cli` 의 base-names 시험(둘 다 check affected 와 함께 와요), 뮤턴트 증명(tier 1 이라 요구되지 않아요), ego(lite) 나 Aside 를 설치한 상태의 실제 사용, 독립 감사.
