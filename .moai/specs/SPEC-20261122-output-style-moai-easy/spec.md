---
id: SPEC-20261122-output-style-moai-easy
title: "The MoAI-Easy output style reads for Claude Code only and its installer takes an unchanged file as its own"
status: draft
revision: 1
created: 2026-11-22
---

# The MoAI-Easy output style reads for Claude Code only and its installer takes an unchanged file as its own

## Intent

사용자 선택 2026-11-22: 리더 보고서(출력 스타일 개선과 moai-adk용 Claude mods 설계, 2-3절)의 개선된 스타일 문장을 제품 템플릿 `skills/moai/references/home/output-style.md` 에 반영해요. 배너는 `MoAI 🗿` 이고 Codex·"모든 도구" 문구가 없어요. 아울러 스타일 설치기(`src/home_style.rs`)가 지금 템플릿 렌더와 같거나 과거에 배포한 렌더와 같은 파일을 MoAI 소유로 받아서, 손으로 옮겨 둔 설치본이 "충돌"에 갇히지 않게 해요(지금은 소유 기록이 옛 판을 가리키면 다음 템플릿 변경이 영영 적용되지 않아요).

공식 문서 근거(2026-11-22 원문 `.md` 전체 읽음: output-styles 204줄, settings 344줄, memory 667줄): 스타일은 지시일 뿐 강제가 아니에요. 맞춤 스타일은 `keep-coding-instructions: true` 가 없으면 기본 소프트웨어 공학 지침을 빼요. 사용자 범위 위치는 `~/.claude/output-styles` 이고, 값 `outputStyle` 은 대소문자를 가려요. 파일은 시작할 때 읽어서 고치면 재시작해야 반영돼요. 프로젝트 `.claude/output-styles` 도 읽히지만 moai 는 쓰지 않아요(사용자 결정, 아래 Decisions D-8).

## Non-goals

- Codex 대상(`~/.codex/AGENTS.md` 의 `moai-output-style` 블록)은 건드리지 않아요. Codex 제거 카드들이 걷어요. 같은 소스 문장을 아직 읽으므로 그때까지 Codex 블록도 새 문장을 담아요.
- `outputStyle` 설정 규칙(없을 때만 넣고 다른 값은 충돌)과 `settings.json` 바이트 보존은 바꾸지 않아요.
- 스타일 이름 `MoAI-Easy`, 설치 경로, 영수증·되돌리기 형식은 바꾸지 않아요.
- 설치 소유 기록 파일의 위치(`~/.moai-cli/home-style/`)는 다른 카드(경로 이전)의 몫이에요. 이 SPEC 의 시험은 위치를 `moai_cli::home_root::path` 로 물어요.
- 설치하지 않는 옛 템플릿 `templates/claude/output-styles/*` 정리, 프로필(`CLAUDE_CONFIG_DIR`)별 복사본, mods 는 이 SPEC 밖이에요.

## Requirements

- REQ-1: The shipped output style source shall name no host other than Claude Code and no "every coding tool" wording.
- REQ-2: The shipped output style source shall carry the picked sections Voice, Rhythm, Size, Banner, Progress board and Format with the four steps, the evidence rule and the short-answer exception.
- REQ-3: The shipped output style source shall state the banner first line as `<emoji> MoAI 🗿 <label> ─…` with the 46 `─` closing line and no star marker.
- REQ-4: The rendered style file shall hold a description without a double quote so that its frontmatter stays valid YAML.
- REQ-5: WHEN the style file equals the current render and the record names another text, the plan shall mark it `merge` so that the apply records it as MoAI's own. · failure: none
- REQ-6: WHEN the style file equals an earlier shipped render, the plan shall mark it `merge` so that the apply replaces it with the current render. · failure: none
- REQ-7: WHEN the style file equals no shipped render and not the text MoAI last wrote, the plan shall mark it `conflict` so that the apply leaves its bytes. · failure: none
- REQ-8: The list of earlier shipped renders shall hold the sha256 of every style file text moai-cli rendered before the current one and never the current one.
- REQ-9: WHEN an apply that adopted or replaced a style file is reverted, the revert shall give back the before bytes and the previous record. · failure: none
- REQ-10: The installer shall write the style file only below the host's config folder and nothing below a project folder.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 템플릿에 `Codex`·`every coding tool` 가 없고, 여섯 절 제목과 네 단계·증거·짧은 답 예외·`MoAI 🗿` 배너 줄·`▓▓▓▓▓▓░░░░`·46 `─` 가 있어요 · verify: `cargo test --offline --locked --test home_style_cli the_shipped_source_names_only_claude_code_and_keeps_the_picked_sections`
- AC-2 (REQ-3): 배포 스타일 파일 네 개에 `MoAI 🗿` 가 있고 맨 `★` 와 옛 배너 문구가 없어요 · verify: `cargo test --offline --locked --test style_banner_scan`
- AC-3 (REQ-4): 렌더된 앞머리의 description 값에 큰따옴표가 없고 키 순서가 `name`, `description`, `keep-coding-instructions` 예요 · verify: `cargo test --offline --locked --test home_style_cli the_style_frontmatter_is_valid_yaml`
- AC-4 (REQ-5): 임시 HOME 에서 소유 기록이 옛 `★` 판을 가리키고 파일이 현재 렌더와 같으면 plan 이 `merge` 이고, apply 뒤 파일 바이트는 그대로이며 기록이 현재 렌더를 가리키고, 템플릿이 바뀐 다음 plan 이 `merge`(충돌 아님)예요 · verify: `cargo test --offline --locked --test home_style_cli a_style_file_equal_to_the_current_text_is_taken_as_owned`
- AC-5 (REQ-6): 임시 HOME 에서 소유 기록이 없거나 다른 판을 가리킬 때 파일이 옛 배포 렌더 세 벌 각각과 같으면 plan 이 `merge` 이고 apply 뒤 파일이 현재 렌더와 같아요 · verify: `cargo test --offline --locked --test home_style_cli a_style_file_equal_to_an_earlier_shipped_text_is_updated`
- AC-6 (REQ-7): 사용자가 고친 파일(현재 렌더·옛 렌더·소유 기록 판 위에 한 줄을 더한 것)은 plan 이 `conflict` 이고 apply 가 바이트를 그대로 둬요 · verify: `cargo test --offline --locked --test home_style_cli a_real_user_edit_stays_a_conflict_on_every_known_text`
- AC-7 (REQ-8): 옛 렌더 목록의 항목이 64자리 16진수·서로 다름이고, 각 항목이 시험 픽스처 파일의 sha256 과 같으며, 현재 렌더의 sha256 이 시험에 못 박혀 있고 목록에 없어요(템플릿이 바뀌면 시험이 실패해서 나가는 렌더를 목록에 더하게 해요) · verify: `cargo test --offline --locked --lib home_style::tests::the_earlier_renders_are_pinned_and_never_hold_the_current_one`
- AC-8 (REQ-9): 채택·교체를 적용한 뒤 되돌리면 파일 바이트와 소유 기록이 적용 전과 같아요 · verify: `cargo test --offline --locked --test home_style_cli reverting_an_adopted_or_replaced_style_file_gives_back_bytes_and_record`
- AC-9 (REQ-10): 프로젝트 폴더에 `.claude/output-styles` 가 있어도 plan 의 대상 경로가 모두 HOME 의 호스트 폴더 아래이고 apply 가 프로젝트 폴더를 쓰지 않아요 · verify: `cargo test --offline --locked --test home_style_cli the_style_is_never_written_below_a_project_folder`
- AC-10 (REQ-1…REQ-10): 이웃 시험이 그대로 통과해요 · verify: `cargo test --offline --locked --test update_user_layer_cli` 와 `cargo test --offline --locked --test host_home_cli` 와 `cargo test --offline --locked --test home_style_cli`

## Scope

- write: `.moai/specs/SPEC-20261122-output-style-moai-easy/**`, `skills/moai/references/*/output-style.md`(`home` 폴더의 한 파일), `skills/moai/references/source-manifest.json`(그 파일 항목의 해시와 사유 한 곳), `src/home_style.rs`, `tests/home_style_cli.rs`, `tests/fixtures/home-style/**`
- read-only: `src/ops/home_style_*.rs`, `src/update.rs`, `src/update_tux.rs`, `src/host_home.rs`, `tests/style_banner_scan.rs`, `tests/update_user_layer_cli.rs`, `tests/host_home_cli.rs`
- forbidden: Codex 대상 바꾸기, `outputStyle` 규칙 바꾸기, 사용자 HOME 쓰기(시험은 임시 HOME 만), 새 동작 이름 더하기(`update_tux` 가 아는 `create`·`merge`·`conflict`·`unchanged` 만 써요)

## Plan

1. 이 SPEC 을 먼저 커밋해요(슬라이스 1).
2. RED(슬라이스 2): `tests/home_style_cli.rs` 와 `tests/fixtures/home-style/render-1.md`~`render-3.md`(옛 배포 렌더 세 벌을 git 이력에서 꺼낸 것, 목록의 sha256 과 맞춰요)와 `src/home_style.rs` 단위 시험을 먼저 더하고, 템플릿·설치기가 옛 모습인 채로 실패하는 것을 눈으로 봐요.
3. GREEN(슬라이스 2): 템플릿을 보고서 2-3절 본문으로 바꾸되 앞머리 없이(D-1), `DESCRIPTION` 을 큰따옴표 없이(D-2), 설치기에 옛 렌더 목록과 채택·교체 규칙(D-3, D-4)을 넣고 `python3 -I tools/source_manifest.py write` 로 매니페스트를 갱신해요.
4. 확인: 위 verify 명령, `cargo test --offline --locked --test update_user_layer_cli`, `host_home_cli`, `style_banner_scan`, `python3 -m unittest tests.source_manifest_test tests.skill_references_test`, `moai spec check`, `git diff --check`, fmt, clippy `--all-targets`, trailer-check, `moai check affected` 한 번.

## Risks

- 옛 렌더 목록은 사람이 관리해요: 템플릿이나 앞머리 형식을 바꾸고 목록에 나가는 렌더를 안 더하면 그 렌더를 가진 설치본이 충돌에 갇혀요. 현재 렌더의 sha256 을 시험에 못 박아서 바꿀 때 시험이 실패하게 해요(AC-7).
- 채택(`merge`)은 같은 바이트를 다시 써요. `update_tux` 미리보기는 이를 "갱신할 것"으로 세어요. 파일 바이트는 바뀌지 않지만 미리보기 숫자에는 한 건이 보여요. 새 동작 이름을 만들면 다른 카드의 미리보기 코드를 건드려서 이름을 늘리지 않았어요(D-3).
- 한국어 어투 한 문장(해요체)이 일반 제품 템플릿에 들어가요. 사용자가 보고서 2-3절을 골랐어요. 되돌리기는 그 문장을 지우는 한 줄 변경이에요(D-6).
- 문서는 `CLAUDE_CONFIG_DIR` 가 설정·세션 기록·플러그인을 옮긴다고만 적고 `output-styles` 는 이름을 대지 않아요. 프로필 폴더가 자기 `output-styles` 를 읽는지는 문서로 확인하지 못했어요. 이 SPEC 은 설치 위치를 지금처럼 `host_home` 이 정하는 폴더로 두고 프로필별 복사는 하지 않아요.
- 옛 소유 기록이 `~/.moai-cli/home-style/` 에 있어요. 위치가 이전되면 그 기록을 읽는 곳이 바뀌어요. 시험은 `home_root::path` 로 위치를 물어서 이전에 끌려가요.
- `docs/claude-adapter-contract.md` 는 "MoAI 는 `outputStyle` 을 스스로 정하지 않는다"고 적었는데 `moai update` 는 비어 있을 때 넣어요. 이 SPEC 은 문서를 고치지 않아요(범위 밖); 감사가 보면 따로 알려요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 템플릿은 보고서 2-3절의 본문만 담고 앞머리는 설치 코드가 붙인다 | `source()` 가 앞머리로 시작하는 소스를 거부하고 설명을 코드에 두는 구조예요(보고서 2-4 둘째 줄) | 템플릿 한 파일의 되돌림 |
| D-2 | `DESCRIPTION` 은 `Plain words, four steps (understand, plan, do, check), evidence before done, MoAI 🗿 banners and a progress board.` 로 큰따옴표를 뺀다 | 제안 파일의 `"done"` 은 설치 코드가 값을 큰따옴표로 감싸서 앞머리를 깨요(`the_style_frontmatter_is_valid_yaml` 이 막아요) | 상수 한 줄 |
| D-3 | 채택과 옛 렌더 교체는 새 동작 이름 없이 `merge` 로 표시한다 | `update.rs`·`update_tux.rs` 가 아는 이름은 `create`·`append`·`merge`·`set`·`unchanged`·`conflict`·`blocked`·`skipped` 뿐이에요 | 분기 한 곳 |
| D-4 | 옛 배포 렌더는 sha256 목록으로 가지고, 현재 렌더가 같은 파일은 소유 기록과 상관없이 채택한다 | git 이력의 렌더 세 벌: 최초 `15578b330b`(앞머리 값 따옴표 없음·`★`) sha256 `93e53607dd194da86abfbe7ea993303fff1d17af1a1a1ae6868788938ed785d2`, YAML 수정 `235f7cafd6`(`★`) `d4e1851a65eb0ed21d16de888925dc6ae8a7bcebcae45973e1b8cce884d75e67`, 배너 `02af21e018`(`🗿`, 현재 팁 `62e457ceaf` 의 렌더) `690c534220f9bb1614e27d608e56306da973e2fe7b53e2bc1a4d42d7d2b9c65e`; 이 세 벌은 `git show <커밋>:src/home_style.rs` 의 `DESCRIPTION` 과 형식, `git show <커밋>:skills/moai/references/home/output-style.md` 로 계산했어요. 마지막 것이 사용자가 오늘 손으로 옮긴 설치본이에요 | 목록 항목과 분기의 제거 |
| D-5 | 앞머리 `keep-coding-instructions: true` 를 유지한다 | 공식 문서: 맞춤 스타일은 이 값이 없으면 기본 소프트웨어 공학 지침을 빼요. 이 스타일은 대화 표현만 바꾸고 코딩 방식은 바꾸지 않아요 | 한 줄 |
| D-6 | 템플릿의 한국어 해요체 문장은 사용자가 고른 대로 둔다 | 보고서 2-3절과 카드 본문(사용자 선택 2026-11-22) | 그 문장 삭제 |
| D-7 | Codex 대상은 이 SPEC 에서 건드리지 않는다 | 카드 본문: "Codex targets go with the Codex removal cards" | 해당 없음 |
| D-8 | 스타일 파일은 호스트 설정 폴더(`CLAUDE_CONFIG_DIR` 또는 `~/.claude`)에만 쓰고 프로젝트 `.claude/output-styles` 에는 쓰지 않는다 | 사용자 결정 2026-11-22(전역 설치): 일반 moai 내용은 전역에 두고 프로젝트 폴더는 쓰지 않아요; 개인 말투를 팀 저장소에 커밋하게 되는 것도 막아요 | 해당 없음 |
