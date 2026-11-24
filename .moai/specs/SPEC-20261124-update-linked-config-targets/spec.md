---
id: SPEC-20261124-update-linked-config-targets
title: "moai update plans a config file that is a link into the default Claude folder, and says why and what next when a target is blocked"
status: draft
revision: 1
created: 2026-11-24
---

# moai update plans a config file that is a link into the default Claude folder, and says why and what next when a target is blocked

## Intent

사용자가 Orca 세션 안에서 `/moai update` 를 해도 `settings.json` 이 안 바뀐다고 했어요. 빌드 4.2648.24 로 같은 환경(`CLAUDE_CONFIG_DIR` 이 Orca 프로필 폴더)에서 미리보기를 보면 스타일 대상 두 개가 모두 `blocked` 예요(카드 C-478 의 측정).

1. 설정 대상(`settings.json`)은 `<프로필>/settings.json is a symlink; refused` 예요. 이 파일은 `~/.claude/settings.json` 으로 가는 링크예요(사용자 결정 2026-11-24). `src/home_style.rs` 의 `current_file` 이 링크를 무조건 거절해요.
2. 스타일 파일 대상은 `<프로필> is not the default claude folder … nothing is written for it` 예요. `output-styles` 폴더가 기본 폴더의 같은 폴더로 가는 링크라서 파일은 `~/.claude/output-styles/MoAI-Easy.md` 에 닿는데도, 소유 기록 이름이 `CLAUDE_CONFIG_DIR` 폴더로 정해져서(`record_name`) 이 HOME 의 상태 폴더가 이 세대 것이 아니면 막혀요.
3. `notices` 가 비어 있어서 왜 안 바뀌는지, 무엇을 하면 되는지 사용자가 몰라요.

같은 미리보기를 `CLAUDE_CONFIG_DIR` 없이 보면 `~/.claude` 의 두 파일이 `unchanged` 예요. 그래서 사용자 눈에는 "update 가 settings.json 을 못 바꾼다" 로 보여요.

이 SPEC 은 (a) 기본 폴더 안의 보통 파일로 곧장 가는 링크는 그 파일을 같은 소유권 규칙으로 계획하고, (b) 막힌 대상은 화면과 `notices` 가 이유와 다음 걸음 한 줄을 말하게 하고, (c) `keys`·`switch` 가 아직 `~/.moai-cli` 를 가리키는 까닭을 밝혀요.

읽는 쪽은 `moai update` 를 돌리는 사용자, 이 변경을 구현하는 레인, 감사하는 세션이에요.

## Non-goals

- 링크 자체를 바꾸거나 만드는 일이에요. 링크는 그대로 둬요.
- 기본 폴더 밖을 가리키는 링크, 링크 사슬, 깨진 링크를 따라가는 일이에요. 지금처럼 거절해요.
- 기본 폴더 안의 링크(`~/.claude/settings.json` 자신이 링크인 경우)를 따라가는 일이에요. 리다이렉트는 `CLAUDE_CONFIG_DIR` 폴더에서만 따라가요.
- `src/host_home.rs` 의 `place_in`(폴더 링크 규칙)과 `src/update.rs` 의 화면 코드를 바꾸는 일이에요. 폴더 링크는 지금 그대로 쓰고, 화면은 이미 있는 `notices` 와 `blocked` 줄을 그대로 써요.
- HOME 상태 폴더를 `~/.moai-cli` 에서 `~/.moai` 로 옮기는 일이에요. 열린 카드 C-199 의 몫이에요(D-5).
- 진짜 `~/.claude`, Orca 프로필, 진짜 HOME 을 읽거나 쓰는 일이에요. 모든 테스트는 임시 HOME 만 자식 프로세스에게 줘요.

## Requirements

- REQ-1: WHEN a style target in a Claude config folder other than the HOME's default folder is a link that points straight at a regular file inside the default folder, the plan shall plan that file with the ownership rules of a plain file so that the link stays as it is.
- REQ-2 (REQ-1): IF that link points outside the default folder, at another link, at nothing or at something that is not a regular file, THEN the plan shall mark the target `blocked` so that nothing is written.
- REQ-3 (REQ-1): WHEN the apply or the revert writes a target that is reached through a link, it shall check the link again at write time and write the real file so that the link stays as it is.
- REQ-8 (REQ-3): IF the link check at write time fails, THEN the apply or the revert shall refuse that target so that no file is written through it.
- REQ-4 (REQ-1): IF the `outputStyle` key would be set in a file of the default folder that is reached through a link while the style file is written in another folder, THEN the plan shall mark the setting `skipped` with that reason.
- REQ-5: WHEN a style file lies in the default folder (reached through a followed folder link or file link), the update shall keep its owned text in the record of the default folder so that no record of the config folder is needed for it. · failure: none — a record name only decides where a record is read and written
- REQ-6: IF a style target is `blocked`, THEN its detail and a notice shall each name the file, the reason and one next step.
- REQ-7: WHEN `moai update` previews `keys` and `switch` for a HOME, the update shall place both files below the HOME state folder that `home_root::path` names. · failure: none — a preview reads and writes nothing

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): 임시 HOME(이전 세대 상태 폴더 `.moai-cli` 만 있음)에서 프로필의 `settings.json` 과 `output-styles` 가 기본 폴더로 가는 링크이면 plan 의 두 대상이 `create`·`set`(또는 기존 값이면 `merge`·`set`)이고, apply 가 기본 폴더의 두 파일을 쓰며 링크는 그대로이고 소유 기록은 `claude.json` 이고, 되돌리면 기본 폴더 파일이 이전 바이트로 돌아와요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files a_config_file_linked_into_the_default_folder_is_planned_and_written_there`
- AC-2 (REQ-2): 링크가 기본 폴더 밖, 링크 사슬, 깨진 링크, 폴더, 기본 폴더 안의 링크 자신, `..` 로 빠져나가는 상대 링크이면 대상이 `blocked` 이고 어느 파일도 쓰이지 않고 밖의 파일 바이트가 그대로예요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files only_a_link_straight_to_a_regular_file_inside_the_default_folder_is_followed`
- AC-3 (REQ-3, REQ-8): plan 뒤에 링크를 사슬로 바꾸거나 밖으로 돌린 다음 apply 하면 거절하고 어느 파일도 쓰이지 않아요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files a_link_changed_after_the_plan_is_refused_at_write_time`
- AC-4 (REQ-4): 프로필 `settings.json` 이 기본 폴더 파일로 가는 링크이고 프로필의 `output-styles` 는 진짜 폴더이면 설정 대상이 `skipped` 이고 이유가 나오며 기본 폴더 `settings.json` 의 바이트가 그대로예요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files a_linked_setting_waits_for_a_style_file_in_the_default_folder`
- AC-5 (REQ-6): `blocked` 대상의 detail 과, 그 plan 으로 `home_style::notices` 가 만드는 `notices`(`moai update` 가 싣는 목록이에요)가 둘 다 파일 이름, 이유, `next:` 한 줄을 담아요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files a_blocked_style_target_gives_a_reason_and_a_next_step_in_detail_and_notices`
- AC-6 (REQ-7): `.moai-cli` 만 있는 임시 HOME 에서는 keys 의 대상 파일이 `.moai-cli` 아래이고, 이 세대 표지가 있는 임시 HOME 에서는 `.moai` 아래예요. switch 의 링크 폴더는 같은 `home_root::path` 에서 나오고(`src/update/switch.rs` 의 `link_dir`) 기존 시험이 이미 잠가요 · verify: `cargo nextest run --locked --offline --test update_linked_config_files keys_follow_the_home_state_folder` 와 `cargo nextest run --locked --offline --test update_switch_cli`
- AC-7 (REQ-1…REQ-7): 이웃 테스트가 그대로 통과해요 · verify: `cargo nextest run --locked --offline --test home_style_cli --test home_style_ownership_cli --test update_user_layer_cli --test update_tux_golden`

## Scope

- write: `.moai/specs/SPEC-20261124-update-linked-config-targets/**`, `src/home_style.rs`, `tests/update_linked_config_files.rs`
- read-only: `src/host_home.rs`, `src/home_root.rs`, `src/state_root.rs`, `src/update.rs`, `src/update_tux.rs`, `src/update_keys.rs`, `src/update/switch.rs`, `src/key_vault.rs`, `tests/home_style_cli.rs`
- forbidden: 진짜 HOME 의 사용자 Claude 폴더와 Orca 프로필의 읽기·쓰기, 링크를 바꾸거나 만드는 일, 이전 HOME 상태 폴더 아래에 새로 만드는 일

## Plan

1. 이 SPEC 을 먼저 커밋해요.
2. RED: `tests/update_linked_config_files.rs` 에 AC-1…AC-6 테스트를 쓰고 실패를 눈으로 봐요.
3. GREEN: `src/home_style.rs` 에 대상 위치 찾기(`locate`)를 더해요. 링크를 따라가는 규칙(`linked_file`), 소유 기록을 대상이 있는 실제 폴더로 정하기(`record_name`), 설정 대기 규칙(REQ-4), 막힘 detail 의 `next:` 줄과 `notices` 를 만들어요.
4. 확인: 위 verify 명령, 가드 뮤턴트, 위험 경로(링크가 기본 폴더 밖, 링크 사슬, 깨진 링크)를 버리는 폴더에서 바이너리로 직접 돌려요. `moai spec check`, `git diff --check`, `cargo fmt`, 맨몸 `moai check affected` 한 번.

## Risks

- 링크를 따라가면 사용자가 링크로 이은 기본 폴더 파일을 쓰게 돼요. 따라가는 것은 기본 폴더 안의 보통 파일로 곧장 가는 링크뿐이고, 쓰기는 plan 의 digest 와 apply 때 다시 읽은 링크 확인을 모두 통과해야 해요(REQ-3). 소유 규칙은 그대로라서 사용자의 `outputStyle` 값과 손으로 고친 스타일 파일은 `conflict` 로 남아요.
- 설정 파일이 기본 폴더 파일이고 스타일 파일이 프로필에만 있으면 기본 폴더의 Claude Code 가 없는 스타일을 가리켜요. 그래서 이 경우는 `skipped` 예요(REQ-4).
- 기록 이름이 대상의 실제 폴더로 정해지면 `output-styles` 가 기본 폴더로 가는 링크인 프로필은 기본 폴더와 기록 하나를 같이 써요. 같은 실제 파일의 소유 텍스트가 두 기록에 갈라지지 않아요. 부작용으로 이 프로필의 영수증은 프로필 폴더를 적지만 파일은 기본 폴더에 있어요(되돌리기는 같은 프로필 환경에서만 돼요, D-5 of SPEC-20261122-output-style-install-checks).
- 화면에는 막힌 줄과 알림 줄이 둘 다 나와 같은 이유가 두 번 보여요. 화면 코드(`src/update_tux.rs`)는 이 카드 범위 밖이라 그대로 둬요. 한 줄로 합치는 일은 후속으로 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 따라가는 링크는 `CLAUDE_CONFIG_DIR` 폴더 안의 파일 링크로, 대상이 기본 폴더 안의 보통 파일이고 중간에 다른 링크가 없어요. 기본 폴더 안의 어느 보통 파일이든 받아요(같은 이름만이 아니에요) | 카드 C-478 기준 1 의 글자("기본 claude 폴더 안의 보통 파일로 가는 링크")를 그대로 따랐어요. 더 좁히려면 같은 상대 경로만 받으면 돼요 | `linked_file` 의 비교 한 곳 |
| D-2 | 소유 기록은 대상의 실제 폴더로 정해요: 기본 폴더에 닿는 스타일 파일은 기본 폴더의 `claude.json`, 아니면 `CLAUDE_CONFIG_DIR` 폴더의 기록이에요 | 같은 실제 파일은 소유 텍스트 기록이 하나여야 해요. 이 세대 아닌 HOME 의 둘째 폴더 막힘(REQ-7 of SPEC-20261122-output-style-install-checks)은 폴더가 정말 둘째 폴더일 때 그대로예요 | `record_name` 한 곳 |
| D-3 | 기본 폴더로 가는 링크 설정 파일에는 스타일 파일도 기본 폴더에 있을 때만 `outputStyle` 을 넣어요 | 기본 폴더의 Claude Code 도 같은 파일을 읽어서, 스타일 파일이 없으면 없는 스타일을 가리켜요 | `setting_plan` 의 가드 한 곳 |
| D-4 | 막힘 이유와 다음 걸음은 `blocked` 의 detail 에 같이 담고, 같은 글을 `notices` 에도 한 줄로 실어요 | 화면은 이미 blocked 줄과 notices 줄을 그려요. 화면 코드를 건드리지 않아요 | `notices` 의 한 블록 |
| D-5 | `keys.destination` 과 `switch.file` 이 `~/.moai-cli` 를 가리키는 것은 남은 코드가 아니에요. 둘 다 `home_root::path(home)` 로 정해지고, 이 세대 표지가 없는 HOME 은 이전 상태 폴더를 그대로 써요(`src/home_root.rs` 머리말). HOME 합치기는 열린 카드 C-199(기준 2 가 열려 있어요)의 몫이라 새 카드를 만들지 않아요 | 카드 C-478 기준 3 의 "옛 상태 폴더라면 … 따로 카드를 열어요"는 카드 C-199 가 이미 덮어요(카드를 먼저 찾아봐요). AC-6 이 두 경로가 같은 한 함수에서 나온다는 것을 잠가요 | AC-6 테스트 삭제 |

## Result

- AC-1: met — `cargo nextest run --locked --offline --test update_linked_config_files` passed 6 of 6 (임시 HOME, 이전 세대 상태 폴더 `.moai-cli`): 프로필의 `settings.json` 과 `output-styles` 가 기본 폴더로 가는 링크이면 plan 이 두 대상을 `create`·`set` 으로 세우고, apply 가 기본 폴더의 두 파일을 쓰고, 링크는 그대로이고, 소유 기록은 `claude.json` 하나(프로필 폴더 기록 없음)이고, 같은 plan 을 다시 보면 `unchanged` 이고, revert 가 `{"a": 1}` 바이트를 되돌려요. 같은 환경을 버리는 폴더에서 바이너리로 직접 돌렸을 때도 `set` 이고 기본 폴더 `settings.json` 이 `{"a": 1, "outputStyle": "MoAI-Easy"}` 가 됐어요.
- AC-2: met — 같은 실행: 기본 폴더 밖, 링크 사슬, 깨진 링크, 폴더, 기본 폴더 안의 링크 자신, `..` 로 빠져나가는 링크는 모두 `blocked` 이고 밖의 파일 바이트가 그대로예요. 버리는 폴더 직접 실행에서도 밖 링크·사슬·깨진 링크 세 경우가 `blocked`, 밖의 파일과 링크가 그대로였어요.
- AC-3: met — 같은 실행: plan 뒤에 링크를 밖으로 돌리면 apply 가 거절하고 기본 폴더의 두 파일과 밖의 파일에 아무것도 쓰이지 않아요. 이 시험이 잠그는 것은 apply 가 새로 만든 plan 의 digest 가 달라져 거절되는 길이에요. apply 안의 두 번째 링크 확인(`locate` 를 쓰기 직전에 한 번 더)은 두 확인 사이의 경쟁에서만 닿아서 결정적 시험이 없어요(Not done 참고).
- AC-4: met — 같은 실행: 프로필 `settings.json` 이 기본 폴더 파일로 가는 링크이고 프로필의 `output-styles` 가 진짜 폴더이면 설정 대상이 `skipped`(이유에 `default folder`)이고 기본 폴더 `settings.json` 바이트가 그대로예요.
- AC-5: met — 같은 실행: `blocked` 대상의 detail 이 `next:` 한 줄을 담고, `home_style::notices` 가 그 plan 의 `blocked` 대상마다 `<파일> was not written: <detail>` 한 줄을 실어요(이유와 다음 걸음이 같은 글이에요). `moai update` 가 이 목록을 그대로 싣는 줄은 `src/update.rs` 의 기존 코드예요.
- AC-6: met — 같은 실행: `.moai-cli` 만 있는 임시 HOME 에서 keys 의 대상 파일(`key_vault::entry_path`)이 `.moai-cli` 아래이고, 이 세대 표지가 있는 임시 HOME 에서는 `.moai` 아래예요. switch 의 링크 폴더는 `src/update/switch.rs` 의 `link_dir` 가 같은 `home_root::path` 로 정해요(읽어서 확인했고, `update_switch_cli` 의 기존 시험이 `.moai-cli/path/moai` 를 잠가요). 카드 기준 3 의 까닭: 둘 다 `home_root::path(home)` 한 함수로 정해져서, 표지 없는 HOME 이 이전 상태 폴더를 쓰는 건 남은 코드가 아니라 설계예요. 합치기는 열린 카드 C-199 의 몫이라 새 카드를 열지 않았어요.
- AC-7: met — `cargo nextest run --locked --offline --no-fail-fast --test update_linked_config_files --test home_style_cli --test home_style_ownership_cli --test update_user_layer_cli --test update_tux_golden --test update_switch_cli` 92 중 92 통과예요. 맨몸 `moai check affected` 영수증은 READY 노트에 있어요.
- 가드 뮤턴트(편집은 정확히 한 번, 시험 파일 전체를 돌리고, 파일은 메모리에서 되돌리고, 끝난 뒤 트리가 깨끗해요): 8 중 8 이 잡혔고 모두 exit 100 이에요 — `outside_default_folder_is_followed`, `link_chain_is_followed`(두 줄을 함께 빼요), `link_inside_the_default_folder_is_followed`, `linked_setting_without_style_in_default_folder`, `record_of_the_config_folder_for_a_shared_style_file`, `shared_means_the_default_folder_only`, `no_notice_for_a_blocked_target`, `no_next_step_added_to_a_blocked_detail`.
- 위험 경로: 버리는 폴더(워크트리 `target/` 아래, 끝난 뒤 지워요)에서 바이너리로 직접: 밖 링크·링크 사슬·깨진 링크는 `blocked` 이고 링크와 밖의 파일 바이트가 그대로, 보통 파일로 가는 올바른 링크는 `set` 이에요.
- Not done: 쓰기 직전의 두 번째 링크 확인은 경쟁에서만 닿아서 뮤턴트가 살아남아요(시험이 없고, 같은 길의 첫 확인은 AC-3 이 잠가요). 깨진 링크의 막힘 글은 `src/host_home.rs` 의 폴더 링크 규칙이 만들어서 "a link to a folder other than …" 로 읽히고 파일에는 어색해요. 그 파일은 이 카드의 쓰기 범위 밖이라 `next:` 줄만 더했고 글 바꾸기는 후속으로 남겨요. 화면은 막힌 줄과 알림 줄에 같은 이유를 두 번 보여 줘요(Risks).
