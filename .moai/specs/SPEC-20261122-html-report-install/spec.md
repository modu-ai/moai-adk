---
id: SPEC-20261122-html-report-install
title: "moai update installs the html-report skill beside the moai skill in the Claude config folder and shows it as its own preview row"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-html-report-skill
---

# moai update installs the html-report skill beside the moai skill in the Claude config folder and shows it as its own preview row

## Intent

부모 SPEC `SPEC-20261122-html-report-skill` 은 패키지에 두 번째 스킬 폴더 `skills/moai-html-report/` 를 실어요. 사용자 결정은 "moai-adk 설치시 이게 포함이 되어야 한다" 이고, 리더 결정(2026-11-22)이 이 조각을 지금 허락했어요. 다른 작업이 잡고 있는 파일(`src/update.rs`, `src/update_tux.rs`, `templates/bundle-packs/packs.json`, 설치 시험 셋)에는 새 줄만 더하고, 리더가 합친 트리를 다시 확인해요.

지금 `moai update` 는 `SKILL_DIR = "moai"` 한 폴더만 복사하고, 미리보기도 스킬 줄이 하나예요. 이 조각은 기존 줄을 바꾸지 않고, 새 모듈 `src/update/extra_skill.rs` 가 두 번째 폴더의 계획·영수증·복사를 맡아요. 영수증, 소유 판단, 버전 비교는 moai 스킬과 같은 규칙을 따라요. 새 모듈은 `update.rs` 의 자식 모듈이라 그 안의 도우미를 그대로 불러요. 다만 `SKILL_DIR` 를 박아 둔 세 도우미(영수증 경로, 영수증 찾기, 알려진 좋은 사본)는 스킬 이름을 받는 사본을 새 모듈에 둬요.

## Non-goals

- 기존 함수에 스킬 이름 매개변수를 다는 정리는 하지 않아요(기존 줄을 바꿔야 해요). 후속 작업으로 남겨요.
- 리더가 손으로 깐 `moai-domain-html-report` 폴더를 다루지 않아요. 리더가 빌드 17 확인 뒤 지워요.
- `creative` 묶음의 `moai-domain-html-report` 항목은 그대로 둬요. 폴더 이름이 달라서 겹치지 않아요.
- 복구(`recovery-plan`)는 이 스킬의 영수증을 보지 않아요. 그 대신 적용이 실패하면 이 부분이 스스로 쓴 파일을 되돌려요(REQ-6, Decisions D-5).

## Requirements

- REQ-1: When `moai update` runs from a release that carries `skills/moai-html-report`, the preview shall show that folder as its own skills row beside the moai skill · failure: none — a release without the folder shows no row.
- REQ-2: When the update is confirmed, moai shall copy `moai-html-report` into the `skills` folder of the Claude config folder with its receipt under `hosts/moai-html-report/` · failure: none — REQ-3 handles a folder moai did not install.
- REQ-3 (REQ-2): IF the `skills/moai-html-report` folder of the Claude config folder holds a package file that no moai receipt owns, THEN the update shall name that folder and leave it unchanged.
- REQ-4: The bundle base names shall include `moai-html-report` · failure: none — `bundle remove` keeps base names.
- REQ-5 (REQ-2): IF a path that a write of the html-report part goes through has a shape the apply would fail on, THEN the preview shall block that part with a reason that names the path.
- REQ-6 (REQ-2): IF a write of the html-report part fails during the apply, THEN the apply shall undo the writes it made and mark its receipt recovered.
- REQ-7: The confirm digest shall bind the plan of every skill folder beside the moai skill · failure: a package changed after the preview is refused as a stale plan and nothing is written.
- REQ-8 (REQ-6): IF the html-report receipt is pending, THEN the preview shall block that part with a reason that names the receipt and the folder and the way out by hand.
- REQ-9: The html-report receipt shall record whether the top skills folder existed before any apply step of the run · failure: none — the reload line follows it.
- REQ-10: The blocked reason of a symlink or a file at the html-report folder shall name that folder · failure: none.

## Acceptance

- AC-1 (REQ-1): 두 번째 폴더가 있는 임시 설치에서 미리보기 JSON 이 `extra_skills.parts` 에 그 스킬 한 줄을 가지고, 터미널 화면이 `skills claude` 와 `skills claude moai-html-report` 두 줄을 따로 보여 줘요 · verify: `cargo test --locked --offline --test update_user_layer_cli the_preview_lists_the_html_report_skill_on_its_own_row_beside_the_moai_skill`
- AC-2 (REQ-2): 확정하면 두 폴더가 원본과 바이트 단위로 같고, 설치된 `moai-html-report/SKILL.md` 머리말이 `user-invocable: false` 이며, 영수증 `hosts/moai-html-report/claude.json` 이 committed 이고, 다음 실행은 할 일이 없어요 · verify: `cargo test --locked --offline --test update_user_layer_cli confirm_installs_both_skill_folders_and_the_html_report_is_not_user_invocable`
- AC-3 (REQ-3): moai 가 깔지 않은 같은 이름 폴더가 있으면 그 줄이 막힘이고 이유가 폴더를 이름으로 말하며, 확정 뒤에도 그 바이트가 그대로이고 moai 스킬은 설치돼요 · verify: `cargo test --locked --offline --test update_user_layer_cli an_html_report_folder_moai_did_not_install_is_named_and_left_untouched`
- AC-4 (REQ-2): `CLAUDE_CONFIG_DIR` 가 있으면 두 폴더가 그 아래 `skills` 에 놓이고 임시 홈의 `.claude` 폴더는 생기지 않아요 · verify: `cargo test --locked --offline --test host_home_cli the_html_report_skill_follows_claude_config_dir_beside_the_moai_skill`
- AC-5 (REQ-4): 묶음 목록의 `base_names` 가 패키지 `skills/` 의 모든 폴더 이름을 가져요 · verify: `cargo test --locked --offline --test product_cli skill_package::the_bundle_base_names_hold_every_installed_skill_folder`
- AC-6 (REQ-5): 스킬 폴더 안에 `references` 라는 일반 파일이 있거나, `SKILL.md` 라는 폴더가 있거나, 쓸 수 없는 하위 폴더가 있으면 미리보기가 그 부분을 막고 이유가 그 경로를 말해요. 확정해도 그 폴더에는 아무것도 쓰지 않고 영수증도 생기지 않아요 · verify: `cargo test --locked --offline --test update_user_layer_cli a_path_the_apply_would_fail_on_blocks_the_html_report_part_at_preview`
- AC-7 (REQ-6): 시험 장치(`MOAI_CLI_TEST_FAULT=extra_skill_last_publish`, 디버그 빌드만)가 마지막 게시를 실패시키면, 새로 깐 경우에는 스킬 폴더가 남지 않고, 고쳐 쓴 경우에는 바뀐 두 파일이 원래 바이트로 돌아와요. 두 경우 모두 영수증은 `recovered` 이고, 다음 미리보기는 막히지 않으며, 결과의 되돌림 줄은 `recovery-plan` 을 가리키지 않아요 · verify: `cargo test --locked --offline --test update_user_layer_cli a_failed_html_report_apply_undoes_its_own_files_and_the_next_preview_is_not_blocked`
- AC-8 (REQ-7): 미리보기 뒤에 `skills/moai-html-report` 의 파일 하나를 바꾸고 확정하면 stale 로 거절되고 HOME 의 바이트가 그대로예요(moai 스킬 포함) · verify: `cargo test --locked --offline --test update_user_layer_cli an_html_report_package_changed_after_the_preview_is_refused_with_nothing_written`
- AC-9 (REQ-8): 영수증을 pending 으로 바꾸면 이 버전 폴더와 다음 버전 폴더의 미리보기 모두 그 부분을 막고, 이유가 영수증 경로와 폴더 경로와 손으로 푸는 방법을 말하며 `recovery-plan` 을 말하지 않아요. 확정해도 폴더는 그대로예요 · verify: `cargo test --locked --offline --test update_user_layer_cli a_pending_html_report_receipt_blocks_the_part_and_names_the_way_out`
- AC-10 (REQ-10): 스킬 폴더 자리에 심볼릭 링크나 일반 파일이 있으면 막힘 이유가 그 폴더 경로를 말해요 · verify: `cargo test --locked --offline --test update_user_layer_cli a_symlink_or_a_file_at_the_html_report_folder_is_named_in_the_reason`
- AC-11 (REQ-9): `skills` 폴더가 없던 HOME 에 처음 깔면 영수증의 `top_skills_existed_before` 가 false 이고 결과가 다시 불러오기를 요구해요. 같은 실행의 moai 스킬 적용이 그 폴더를 먼저 만들어도 같아요 · verify: `cargo test --locked --offline --test update_user_layer_cli a_first_html_report_install_records_that_the_skills_folder_was_new_to_this_run`

## Scope

- write: `.moai/specs/SPEC-20261122-html-report-install/**`, `src/update/extra_skill.rs`(새)
- write(리더 결정으로, 새 줄만): `src/update.rs`(모듈 한 줄과 연결 줄 열 개), `src/update_tux.rs`(미리보기·결과 줄), `templates/bundle-packs/packs.json`(`base_names` 한 줄), `tests/update_user_layer_cli.rs`, `tests/host_home_cli.rs`, `tests/product_cli.rs`(새 시험과 도우미)
- read-only: `src/host_home.rs`, `src/user_assets.rs`, `src/recovery.rs`, `src/update_cleanup.rs`
- forbidden: 위 여섯 파일의 기존 줄 고치기·지우기, 실제 HOME, 새 Python 시험, 카드 번호

## Plan

1. 새 모듈 `src/update/extra_skill.rs`: 원본은 설치된 `skills/moai` 옆 `skills/moai-html-report`(없으면 이 부분 자체가 없음). 대상은 `host_home::place(Claude, home, "skills/moai-html-report")`. 파일마다 세 갈래 계획과 같은 소유 규칙. 영수증 `installations/<id>/hosts/moai-html-report/claude.json`, 알려진 좋은 사본 `hosts/moai-html-report/claude/known-good`. `--confirm` 은 다시 계획하고 pending 영수증 → 파일 게시 → 확인 → committed. 게시가 실패하면 이 적용이 쓴 파일만 되돌리고 영수증을 recovered 로 둬요. 계획 단계에서 쓰기가 지나는 경로의 모양(파일 자리의 폴더, 폴더 자리의 파일, 쓸 수 없는 폴더, 링크)을 미리 보고 막아요.
2. `src/update.rs` 새 줄: 모듈 선언, `Layer.extra_skills`, 계획 호출, 합친 digest, 미리보기 요약 줄, `writes()`, 미리보기 JSON `extra_skills`(부분이 있을 때만), 적용 호출과 결과 상태, 결과 JSON.
3. `src/update_tux.rs` 새 줄: 미리보기 행, 바뀐 파일 묶음, 최신 화면 합계, 적용 결과 행, 다시 불러오기 표시.
4. `packs.json` 의 `base_names` 에 한 줄.
5. 시험 여섯(AC-1 에서 AC-5, 그리고 두 번째 버전 폴더가 영수증을 이어받는 시험).

## Risks

- 미리보기 JSON 의 최상위 키 목록을 고정한 시험(`tests/update_user_layer_cli.rs` 의 `PREVIEW_KEYS`)이 있어서, `extra_skills` 키는 부분이 있을 때만 나와요. 지금의 시험 준비들은 `skills/moai` 만 복사하므로 그대로 통과해요.
- 새 모듈에 영수증·알려진 좋은 사본 논리 약 160줄이 `update.rs` 와 겹쳐요. 기존 줄을 바꾸지 않는다는 조건의 값이고, 후속 작업이 매개변수로 합쳐요.
- 적용 중에 게시가 실패하면 이 적용이 만든 파일은 지우고(쓴 바이트 그대로일 때만) 고쳐 쓴 파일은 원래 바이트로 돌려놓은 뒤 영수증을 recovered 로 둬요. 되돌리기마저 실패하면(예: 그사이 다른 쓰기가 파일을 바꿈) 영수증이 pending 으로 남아 이 부분이 막혀요. `recovery-plan` 은 `moai` 영수증 폴더만 봐요(`src/recovery.rs`). 그래서 막힘 이유가 영수증 경로, 폴더 경로, 손으로 푸는 방법(폴더를 옆으로 옮기고 영수증을 지움)을 말해요.
- 계획과 확정 사이에 경로 모양이 바뀌면 확정의 다시 계획이 그 차이를 보고 거절해요. 확정의 다시 계획 뒤, 게시 도중에 바뀌는 경우만 되돌리기가 맡아요.
- Go 시절 HOME 의 정리 미리보기(`src/update_cleanup.rs`)가 이 폴더를 Go 형제 스킬로 "keep" 이라 보여 줄 수 있어요. 지우지는 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 영수증은 스킬마다 `hosts/<skill>/` 폴더예요. | moai 스킬과 같은 모양이라 읽는 쪽이 하나의 규칙만 알면 돼요. 중복은 후속 정리. | 카드 열림 |
| D-2 | moai 스킬 복사가 막혀도 html-report 는 따로 계획·설치해요. | 두 폴더는 서로 기대지 않아요. | 카드 열림 |
| D-3 | 미리보기 키 `extra_skills` 는 부분이 있을 때만 나와요. | 기존 키 목록 시험과 옛 릴리스를 그대로 둬요. | 카드 열림 |
| D-4 | 기존 줄은 하나도 바꾸지 않아요. 감사 수리는 예외로 `src/update.rs` 의 두 줄(계획 digest 개수 말, moai 스킬 되돌림 줄)을 고쳐요. | 리더 결정(2026-11-22): 다른 작업이 잡은 파일에는 새 줄만. 감사 N6 과 B1(3)이 두 줄을 고치라고 했어요. | 카드 열림 |
| D-5 | 실패한 적용의 탈출구는 `recovery-plan` 확장이 아니라 적용 안의 되돌리기예요. | 되돌리기는 이 적용이 방금 쓴 파일과 그 바이트를 메모리에 들고 있어서 알려진 좋은 사본을 다시 찾을 필요가 없고, `src/recovery.rs`(읽기 전용)를 바꾸지 않아요. 영수증 상태는 이미 있는 `recovered` 를 써요. | 카드 열림 |
| D-6 | 경로 모양 검사는 계획 단계에서 해요. 적용 단계는 같은 검사를 다시 계획으로 받아요. | 미리보기가 "ok" 라고 한 부분이 확정에서 반쯤 깔리는 일을 막아요(감사 B1). | 카드 열림 |

## Result

(status draft — written when the slice lands)
