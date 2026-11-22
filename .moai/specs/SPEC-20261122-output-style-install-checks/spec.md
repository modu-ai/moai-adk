---
id: SPEC-20261122-output-style-install-checks
title: "The update checks that the MoAI-Easy style is the one in force and writes it once per Claude config folder"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-output-style-moai-easy
---

# 스타일 설치 점검과 설정 폴더별 설치

## Intent

리더 노트와 전체 문서 조사 보고서 9-2절, 사용자 결정 2026-11-22(전역 설치)에서 나온 두 가지예요. 하나, `moai update` 가 설치 뒤에 스타일이 실제로 켜지는지 확인하고 알려 줘요(고치지는 않아요): `outputStyle` 값은 대소문자를 가려서 틀리면 Claude Code 가 조용히 기본 스타일을 쓰고, 프로젝트의 `.claude/settings.json` 이나 `.claude/settings.local.json` 에 다른 `outputStyle` 이 있으면 그 프로젝트에서는 그 값이 이겨요(`/output-style` 은 로컬 파일에 써요). 스타일 파일이 깨졌으면 Claude Code 의 `/status` 가 그 목록을 보여 주니 그리로 안내해요. 둘, 스타일을 사용자가 쓰는 Claude 설정 폴더마다 한 벌씩, 폴더별 기록과 함께 써요(Orca 프로필처럼 설정 폴더가 여럿일 때). 기록이 하나뿐이면 한 폴더에서 손댄 파일이 다른 폴더의 정상 설치까지 충돌로 만들어요.

부모 SPEC `SPEC-20261122-output-style-moai-easy` 가 템플릿과 소유 규칙을 정해요. 이 SPEC 은 그 위에서 점검·알림과 폴더별 기록만 더해요. 부모의 슬라이스 2(템플릿과 설치기 소유 규칙)가 먼저 들어간 뒤에 시작해요(같은 파일 `src/home_style.rs`·`tests/home_style_cli.rs` 를 쓰므로 겹침을 피해요).

## Non-goals

- 설정 폴더 목록을 보여 주고 고르는 화면과 고른 목록의 저장은 폴더 선택을 맡는 레이아웃 SPEC들의 몫이에요. 이 SPEC 은 폴더 하나를 받아 계획·적용·기록·되돌리기가 되게 하고, 기본값은 `host_home` 이 정하는 폴더예요.
- 프로젝트의 `.claude/settings*.json` 은 읽기만 해요. 알림만 내고 고치지 않아요.
- 프로필 폴더가 자기 `output-styles` 를 읽는지는 문서로 확인하지 못한 채예요(Risks). 폴더별 설치는 읽는다고 주장하지 않고, 폴더마다 같은 한 벌을 쓸 뿐이에요.
- `outputStyle` 을 고치는 일, 사용자 값을 덮어쓰는 일, `force-for-plugin` 은 하지 않아요.
- 기본 폴더의 기존 기록(`claude.json`)의 이전(`~/.moai-cli/home-style/` 에서 새 자리로)은 경로 이전 카드의 몫이에요. 이 SPEC 은 그것을 있는 자리에서 읽고 쓰는 지금 동작만 이어 가요.
- `~/.moai-cli` 아래에는 새로 아무것도 만들지 않아요(사용자의 상시 지시 2026-11-21). 둘째 폴더부터의 새 기록은 처음부터 현재 상태 루트(`~/.moai`)에 만들어요.

## Requirements

- REQ-1: WHEN the `outputStyle` value of the config folder's `settings.json` is a string that equals `MoAI-Easy` only when the case is ignored, the plan shall mark the setting `conflict` so that its detail says the name is case-sensitive and that Claude Code then uses its default style. · failure: none
- REQ-2: WHEN the `.claude/settings.local.json` or the `.claude/settings.json` of the project holds an `outputStyle` string other than `MoAI-Easy`, the update shall show a notice that names the file and says this project uses its value, with no write to either project file.
- REQ-3 (REQ-2): IF that project settings file cannot be read or is not a JSON object, THEN the update shall go on with its other steps and give no notice for that file.
- REQ-4: WHEN the apply creates or replaces the style file, the update shall tell the user that `/status` in Claude Code lists a style file it could not load. · failure: none
- REQ-5: WHEN the style is installed into a second config folder, the installer shall keep one owned-text record per folder so that an edit of the file in one folder is a conflict only in that folder. · failure: none
- REQ-6: WHEN the style is installed into a second config folder, the installer shall write that folder's record below the state root of this generation so that nothing is created below the earlier state folder `.moai-cli` of the HOME.
- REQ-7 (REQ-6): IF the state root of the HOME is not this generation's, THEN the plan shall mark the targets of the second folder `blocked` with that reason so that no file is written for it.

## Acceptance

- AC-1 (REQ-1): 설정 폴더 `settings.json` 의 `outputStyle` 이 `moai-easy`·`MOAI-EASY` 이면 plan 의 설정 대상이 `conflict` 이고 detail 에 대소문자와 기본 스타일이 나오며, 정확히 `MoAI-Easy` 면 `unchanged` 예요 · verify: `cargo test --offline --locked --test home_style_cli the_setting_plan_says_the_style_name_is_case_sensitive`
- AC-2 (REQ-2, REQ-3): 임시 프로젝트의 로컬 설정이나 프로젝트 설정에 다른 `outputStyle` 이 있으면 알림이 그 파일 이름과 함께 나가고 두 파일의 바이트가 그대로이며, 읽을 수 없거나 JSON 객체가 아닌 파일은 알림 없이 나머지 단계가 끝까지 가요 · verify: `cargo test --offline --locked --test home_style_cli a_project_setting_that_hides_the_style_gives_a_notice_and_writes_nothing` 와 `cargo test --offline --locked --test home_style_cli an_unreadable_project_setting_gives_no_notice`
- AC-3 (REQ-4): 스타일 파일을 만들거나 바꾸는 update 의 미리보기 한 줄에 `/status` 안내가 있고, 바뀌지 않는 update 에는 없어요 · verify: `cargo test --offline --locked --test home_style_cli the_update_points_to_status_after_a_style_write`
- AC-4 (REQ-5): 임시 설정 폴더 두 곳에 설치하면 기록이 두 개이고, 한 폴더의 파일을 고치면 그 폴더만 충돌이며, 한 폴더를 되돌려도 다른 폴더의 파일과 기록이 그대로예요; 기본 폴더의 기록 파일 이름은 이전과 같아요 · verify: `cargo test --offline --locked --test home_style_cli two_config_folders_keep_two_records`
- AC-5 (REQ-6): 임시 HOME(그 HOME 의 `.moai` 가 이 세대 것)에서 둘째 설정 폴더에 설치하면 새 기록이 그 HOME 의 `.moai` 아래 `home-style` 폴더에 생기고, 설치 앞뒤로 그 HOME 의 `.moai-cli` 폴더의 목록(파일 이름과 sha256)이 같아요(없던 폴더는 설치 뒤에도 없어요) · verify: `cargo test --offline --locked --test home_style_cli installing_into_a_second_config_folder_creates_nothing_under_the_earlier_state_folder`
- AC-6 (REQ-7): 임시 HOME 의 `.moai` 가 이 세대 것이 아니면(표지 없음, 이전 `.moai-cli` 만 있음) 둘째 폴더의 대상이 `blocked` 이고 이유가 나오며 어느 파일도 쓰이지 않고, 첫(기본) 폴더의 계획은 그대로예요 · verify: `cargo test --offline --locked --test home_style_cli a_second_folder_is_blocked_while_the_state_root_is_not_this_generations`
- AC-7 (REQ-1…REQ-7): 이웃 시험이 그대로 통과해요 · verify: `cargo test --offline --locked --test update_user_layer_cli` 와 `cargo test --offline --locked --test host_home_cli` 와 `cargo test --offline --locked --test home_style_cli`

## Scope

- write: `.moai/specs/SPEC-20261122-output-style-install-checks/**`, `src/home_style.rs`, `src/update.rs`(스타일 단계가 알림을 싣는 곳만), `src/update_tux.rs`(스타일 줄과 알림 줄만), `tests/home_style_cli.rs`, `tests/fixtures/home-style/**`
- read-only: `src/ops/home_style_*.rs`, `src/host_home.rs`, `src/home_layout.rs`, `src/home_root.rs`, `src/state_root.rs`, `tests/update_user_layer_cli.rs`, `tests/host_home_cli.rs`
- forbidden: 프로젝트 `.claude` 쓰기, `outputStyle` 값 고치기, 사용자 HOME 쓰기(시험은 임시 HOME 만), Codex 대상, 새 동작 이름 더하기

## Plan

1. 이 SPEC 을 먼저 커밋해요(슬라이스 1).
2. 부모의 슬라이스 2 가 들어간 뒤: `update.rs` 가 스타일 단계를 만드는 곳과 `update_tux.rs` 의 스타일 줄·차단 줄을 끝까지 읽어서 알림이 놓일 자리를 정해요. 미리보기는 지금 스타일 `conflict` 를 숫자로만 세므로 알림 줄이 새로 필요해요.
3. RED: REQ-1…REQ-5 의 시험과 픽스처를 먼저 쓰고 실패를 눈으로 봐요.
4. GREEN: 설정 대상 detail(REQ-1), 프로젝트 설정 읽기와 알림(REQ-2, REQ-3), 쓴 뒤 `/status` 안내(REQ-4), 폴더를 인자로 받는 계획·적용·기록·되돌리기와 폴더별 기록 이름(REQ-5, D-1).
5. 확인: 위 verify 명령, `moai spec check`, `git diff --check`, fmt, clippy `--all-targets`, trailer-check, `moai check affected` 한 번.

## Risks

- 프로필 폴더가 자기 `output-styles` 를 읽는지는 문서로 확인하지 못했어요(`CLAUDE_CONFIG_DIR` 는 설정·세션 기록·플러그인을 옮긴다고만 적혀 있어요). 폴더별 설치는 읽는다고 주장하지 않고, 읽는지는 프로필 하나로 `/status` 를 열어 사람이 확인해요. 이 SPEC 의 시험은 폴더 안의 바이트와 기록만 증명해요.
- 알림 줄은 `update_tux.rs` 에 새로 생겨요. 이 파일은 다른 카드의 미리보기 코드와 같은 파일이라 겹치면 병합 충돌이 나요. 스타일 줄과 알림 줄 두 곳만 건드리고 다른 줄의 형식은 그대로 둬요.
- 기본 폴더의 기록은 `~/.moai-cli` 아래에 있는 HOME 이 많고 새 폴더의 기록은 `~/.moai` 아래에 생겨서, 한 설치의 기록이 두 곳에 흩어져요. 경로 이전 카드가 기본 폴더의 기록을 옮기면 한곳이 돼요. 그때까지 두 곳을 읽는 코드는 기록 이름 함수 하나뿐이에요(D-1).
- 이전 세대 HOME 에서는 둘째 폴더 설치가 막혀요(REQ-7). 경로 이전이 끝난 HOME 에서만 여러 폴더가 돼요. 사용자는 막힘 이유로 그 사실을 보아요.
- 프로젝트 설정이 어디에 있는지는 update 가 도는 폴더에 달려요. 프로젝트가 아닌 곳에서 돌면 읽을 파일이 없고 알림도 없어요(REQ-3 과 같은 길).
- `/output-style` 이 쓰는 로컬 파일 때문에 알림이 사람마다 다를 수 있어요. 알림은 값이 있고 다를 때만 나가요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 폴더별 기록: 기본 폴더(`host_home` 이 정한 곳)의 기록은 지금 자리·이름 그대로(`<지금의 상태 폴더>/home-style/claude.json`, 읽고 쓰는 동작은 바뀌지 않고 이전은 경로 카드가 해요)이다. 다른 설정 폴더의 새 기록은 처음부터 현재 상태 루트에 `home_layout::rust_path(home, "home-style")/claude-<폴더 경로 sha256 의 앞 12자리 16진수>.json` 로 쓴다. 이 세대의 상태 루트가 아닌 HOME 에서는 둘째 폴더를 `blocked`(REQ-7) | 사용자의 상시 지시 2026-11-21: `~/.moai-cli` 아래에 새로 아무것도 만들지 않아요(리더 검토). `home-style` 은 HOME 배치표(`src/home_layout.rs`)에 이미 이 세대 이름으로 있어요. 이전 세대 HOME 에 쓰면 `~/.moai-cli` 도, Go 의 `~/.moai` 도 건드리게 되어서 막아요. 12자리 16진수는 `Moai-Session` 의 작업트리 키와 같은 모양이에요 | 기록 이름 함수 한 곳 |
| D-2 | `outputStyle` 이 대소문자만 다르면 `conflict` 로 두고 고치지 않으며, 말은 detail 한 줄이다 | 공식 문서: 값은 대소문자를 가리고 틀리면 조용히 기본 스타일을 써요. 사용자 값은 "사용자의 선택"이라는 기존 규칙이 먼저예요 | detail 분기 한 곳 |
| D-3 | 프로젝트 설정에서 `outputStyle` 을 읽기만 하고 알림만 낸다 | 보고서 9-2: `/output-style` 이 `.claude/settings.local.json` 에 써서 사용자 값을 이겨요. 사용자 결정: 프로젝트 `.claude` 에는 쓰지 않아요 | 알림 줄과 읽기 함수 삭제 |
| D-4 | 깨진 스타일 파일은 우리가 검사하지 않고 `/status` 로 안내한다 | 공식 변경 기록 2.1.327: `/status` 가 깨진 스타일 파일을 보여 줘요(보고서 9-2) | 안내 한 줄 삭제 |
