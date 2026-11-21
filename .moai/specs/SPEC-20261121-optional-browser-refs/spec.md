---
id: SPEC-20261121-optional-browser-refs
title: "The bundled Aside and ego(lite) references follow the installed official guides and never need the tool"
status: draft
revision: 1
created: 2026-11-21
---

# The bundled Aside and ego(lite) references follow the installed official guides and never need the tool

## Intent

선택 도구(Aside, ego(lite) 브라우저) 레퍼런스 3 조각 중 첫 조각이에요. 사용자가 2026-11-21 에 확인한 규칙: 도구의 정본은 설치된 공식 스킬·CLI 가 말하는 현재 버전의 안내예요. MoAI 글은 "언제 쓰는가, 어디까지 허용하는가, 도구가 없거나 오프라인이거나 권한이 거절됐을 때 안전하게 멈추는 법"만 말해요. 공식 스킬을 통째로 복제하거나 설치하거나 사용자의 도구 설정·MCP 설정을 바꾸지 않아요.

지금 번들 팩의 `moai-ref-aside-browser` 는 옛 안내를 남겨요: 도구 이름·플래그를 본문에 적어 두고, `moai mcp add aside …` 로 MCP 서버를 등록하라고 권해요. 이 저장소의 어댑터(`skills/moai/references/adapters/aside.md`)는 이미 "CLI 만, MCP 는 안 쓴다"가 사용자 결정이에요. 그래서 번들 글을 그 결정에 맞춰 정리하고, 같은 모양의 `moai-ref-ego-browser` 를 더해요(설치된 `ego-browser skill` 이 정본이에요: 이 Mac 에서 2.2.1, 브라우저 0.6.3.17). 두 도구 모두 없어도 설치와 작업은 평소대로 돼요.

## Non-goals

- 공식 스킬의 본문·API 목록을 복제하지 않아요. 도구를 설치·업그레이드하지 않아요.
- 사용자의 도구 설정, 개인 MCP 설정, 브라우저 프로필·쿠키 가져오기를 건드리지 않아요.
- 새 MCP 서버, 새 Rust 하위 명령, 훅 덮어쓰기를 만들지 않아요.
- `skills/moai` 의 워크플로 글(웹 페이지 절)과 어댑터는 바꾸지 않아요(이미 `aside guide`, `ego-browser skill` 을 가리켜요).
- 도구가 "멈춤·대기·끝남" 표시를 낸 것을 M-Flow 의 완료로 보지 않아요(완료는 원장·감사·로컬 병합이에요).

## Requirements

- REQ-1: The bundled Aside reference shall name the `aside` command line as its only way in.
- REQ-2: The bundled Aside reference shall give no MCP registration advice.
- REQ-3: The bundled Aside reference shall send the reader to the installed `aside guide` for the flags of the installed version instead of listing them.
- REQ-4: The bundled ego(lite) reference shall name `ego-browser skill` as the guide of the installed version.
- REQ-5: The bundled ego(lite) reference shall not copy the interface of that guide.
- REQ-6: The bundled ego(lite) reference shall say when to use the tool and how far it may go.
- REQ-7: When the tool is missing, offline or refused a permission, each reference shall tell the reader to record the step as `not_run` with the reason and to go on without the tool.
- REQ-8 (REQ-7): IF the tool answers a probe or a call with a failure for any other reason, THEN each reference shall tell the reader to treat the tool as absent.
- REQ-9: Each reference shall forbid installing or upgrading the tool, importing browser data and changing the user's tool or MCP settings.
- REQ-10: Each reference shall say that a tool's stopped, idle or done badge is not completion of the card.
- REQ-11: The `ops-tools` pack data shall list the ego(lite) reference as a skill entry beside the Aside one.
- REQ-12 (REQ-11): IF neither tool is installed, THEN `moai bundle add ops-tools` shall still install both references.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `moai-ref-aside-browser` 가 `aside guide` 를 가리키고, CLI 만 쓴다고 말하고, `moai mcp add` 와 "Optional registration" 이 없어요.
  verify: `cargo test --locked --offline --test optional_browser_refs_cli the_aside_reference_is_cli_only_and_points_to_the_installed_guide`
- AC-2 (REQ-4, REQ-5, REQ-6): `moai-ref-ego-browser` 가 있고 `ego-browser skill` 을 정본으로 가리키며, 언제·어디까지를 말하고, 공식 API 호출 줄(`await page.`, `task.`)이 없고 짧아요(120 줄 이하).
  verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ego_reference_points_to_the_installed_guide_without_copying_it`
- AC-3 (REQ-7, REQ-8, REQ-9, REQ-10): 두 글 모두 `not_run`, 설치·업그레이드·가져오기 금지, 설정 불변, 배지는 완료가 아님을 말해요.
  verify: `cargo test --locked --offline --test optional_browser_refs_cli both_references_state_the_safe_stop_the_limits_and_the_badge_rule`
- AC-4 (REQ-11): `packs.json` 의 `ops-tools` 가 두 항목을 경로가 맞게 가지고, 폴더가 실제로 있고, 이름이 SKILL.md 머리말과 같아요.
  verify: `cargo test --locked --offline --test optional_browser_refs_cli the_ops_tools_pack_lists_both_references_with_matching_names`
- AC-5 (REQ-12): 도구가 없는 깨끗한 HOME 에서 `moai bundle add ops-tools` 가 성공하고 두 글이 설치돼요.
  verify: `cargo test --locked --offline --test optional_browser_refs_cli adding_ops_tools_installs_both_references_with_no_browser_tool_present`
- AC-6: 이웃 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test bundle_user_assets_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-optional-browser-refs/spec.md`, `templates/bundle-packs/.claude/skills/moai-ref-aside-browser/SKILL.md`, `templates/bundle-packs/.claude/skills/moai-ref-ego-browser/SKILL.md`, `templates/bundle-packs/packs.json`, `tests/optional_browser_refs_cli.rs`(새 파일)
- read-only: `skills/moai/references/adapters/aside.md`, `skills/moai/workflows/`
- forbidden: 공식 스킬 복제, 도구 설치, 사용자 설정, `skills/moai` 바꾸기

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: `tests/optional_browser_refs_cli.rs` 의 시험이 지금 코드에서 실패하는 것을 봐요.
3. GREEN: Aside 글을 CLI 정본으로 정리하고, ego(lite) 글과 `packs.json` 항목을 더해요.
4. 확인: 새 시험, `bundle_user_assets_cli`, 번들·manifest 를 읽는 이웃 시험, 변이, `moai check affected`. 선택: 로컬 페이지에서 `ego-browser` 한 번 확인(도구가 있을 때만).

## Risks

- `packs.json` 의 `sha256` 열은 코드가 읽지 않는 옛 값이고 파일 바이트에서 재현되지 않아요(`Entry` 구조체에 이 열이 없어요). 바꾸는 항목과 새 항목에는 `SKILL.md` 의 SHA-256 을 넣고, 옛 항목은 그대로 둬요.
- 글에 도구 이름·옵션을 적으면 버전이 바뀔 때 낡아요: 옵션 목록 대신 "설치된 안내를 먼저 읽는다"를 적어 둬요. 이 글이 말하는 옵션은 도구 도움말에서 확인한 `--permission` 하나예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | MCP 등록 안내를 지운다(대신 CLI 만 쓴다고 쓴다) | 어댑터의 사용자 결정과 이 카드의 확인 규칙이에요. 사용자가 직접 넣은 MCP 항목은 그대로 둬요 | 글 되돌리기 |
| D-2 | ego(lite) 글은 어댑터가 아니라 번들 팩에 둔다 | 선택 도구이고 Aside 글이 이미 같은 팩(`ops-tools`)에 있어요 | 항목 빼기 |

## Result

- AC-1: met — `cargo test --locked --offline --test optional_browser_refs_cli the_aside_reference_is_cli_only_and_points_to_the_installed_guide` 통과(`aside guide` 를 가리키고 "CLI only" 가 있으며 `moai mcp add`, "Optional registration", `--args mcp` 가 없어요). 고치기 전에는 `no pointer to the installed guide` 로 실패했어요(RED).
- AC-2: met — `… the_ego_reference_points_to_the_installed_guide_without_copying_it` 통과(`ego-browser skill` 이 정본, "When to use"·"How far it may go" 절, `await page.`·`task.page(`·`taskSpace(`·`task.finish(` 없음, 70 줄 안팎으로 120 줄 이하). 고치기 전에는 파일이 없어서 실패했어요.
- AC-3: met — `… both_references_state_the_safe_stop_the_limits_and_the_badge_rule` 통과(두 글 모두 `not_run`, "as absent", 설치·업그레이드 금지, 사용자 설정·MCP 설정 불변, 배지는 완료가 아님, ego 글은 금지하는 `ego-browser import` 를 이름으로 대요).
- AC-4: met — `… the_ops_tools_pack_lists_both_references_with_matching_names` 통과(경로·종류·루트·빈 `depends_skills`·64 자 `sha256`, 머리말 이름이 항목 이름과 같아요).
- AC-5: met — `… adding_ops_tools_installs_both_references_with_no_browser_tool_present` 통과(깨끗한 HOME 과 PATH 에서 `bundle add ops-tools` 성공, 두 글이 출하 바이트 그대로 설치돼요).
- AC-6: met — `cargo test --locked --offline --test bundle_user_assets_cli` 6 개 통과, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `cargo fmt` 적용, `python3 -m unittest tests.repo_scans_test tests.ci_scope_test` 53 개 통과.
- 변이 5 개 모두 시험이 잡았어요(Aside 글에 MCP 등록 줄 되살리기, ego 글에서 `not_run` 지우기, ego 글에 공식 API 줄 복사, ego 머리말 이름 바꾸기, 팩 항목 이름 바꾸기). 처음 한 번은 새 글이 `.gitignore` 의 `.claude/` 에 걸려 커밋에 빠진 채였고 복원이 실패해 변이 흔적이 쌓였어요: 파일을 원본으로 되돌려 `packs.json` 에 적은 해시와 맞는 것을 확인하고(`git add -f` 로 추가), 메모리에 둔 원본으로 복원하는 방식으로 다섯을 다시 돌렸어요.
- 확인한 사실(2026-11-21, 이 Mac): `ego-browser skill` 은 "ego-browser usage guide v2.2.1 (macos)", `ego-browser --version` 은 0.6.3.17, `aside --version` 은 1.26.1115.1556 이에요. 글이 쓰는 도구 사실은 이 출력에서 왔어요: 업그레이드는 사용자 승인, 사람이 해야 할 일(로그인·사람 확인·2차 인증)은 넘겨주고 멈춤, 프로필은 사용자가 지정하지 않으면 건드리지 않음, `ego-browser import` 는 브라우저 프로필 가져오기, Aside 는 `--permission ask|guard|full-access`(`ask` 와 `guard` 는 같은 모드, 기본값).
- `packs.json` 의 `sha256` 열은 코드가 읽지 않는 옛 값이고 바이트에서 재현되지 않아요: 바꾼 Aside 항목과 새 ego 항목에는 `SKILL.md` 의 SHA-256 을 넣었어요.
- 돌리지 않은 것: 도구가 설치된 상태에서의 실제 사용(로컬 페이지 확인은 이 조각의 시험이 아니에요), 전체 `moai check affected`(병합 준비 노트에 줄째로 적어요), `skills/moai` 의 어댑터·워크플로 글(이 조각에서 바꾸지 않았어요: 이미 `aside guide`, `ego-browser skill` 을 가리켜요).
