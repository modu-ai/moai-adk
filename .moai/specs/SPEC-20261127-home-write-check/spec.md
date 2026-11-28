---
id: SPEC-20261127-home-write-check
title: "HOME 상태에 쓰는 모든 명령이 공통 쓰기 검사 하나를 거쳐요"
status: draft
revision: 2
created: 2026-11-27
approval:
  by: user
  at: 2026-11-28
  basis: "승인 노트 10957(USER-APPROVAL, 개정 2, 커밋 764e16be90): 사용자가 2026-11-28 에 리더 세션의 질문 도구에서 직접 고른 '승인 (Recommended)'과 D-1·D-17, D-7, D-9 의 기본값 답(선택은 리더가 원장에 옮긴 글이고 이 레인이 직접 보지는 않았어요). digest 는 그 커밋의 파일에서 moai spec show 로 읽은 approval_digest 예요"
  digest: "sha256:369375d0678a72d17c3c86dad7e8659ab1ea48833d7f43cce4093d7d0585e7eb"
---

# HOME 상태에 쓰는 모든 명령이 공통 쓰기 검사 하나를 거쳐요

## Intent

카드 C-613 은 C-199(HOME 해석기를 `~/.moai` 하나로 고정) 앞에 놓인 순서 게이트예요. 사용자 결정(2026-11-27, C-199 노트 10937 답 1, "쓰기 검사를 먼저 넣음 (Recommended)")이 그 이유예요. 지금 해석기(`home_root::path`)는 Go 세대·mixed·interrupted·refuse 로 판정된 `~/.moai` 를 만나면 `~/.moai-cli` 로 돌려서, 그런 HOME 에 새 파일을 섞어 쓰지 않아요. 고정하고 나면 그 돌림이 없어져서, HOME 에 쓰는 곳들이 v3(Go) 사용자의 폴더에 곧바로 써요. 오늘 그 판정을 스스로 거절하는 쓰기는 둘뿐이에요: `home_root::backup_root`(`src/home_root.rs:59-69`)와 레저의 `check_home_ready`(`src/ledger/open.rs:239-256`). 나머지 쓰기 지점은 검사가 없어요(아래 Plan 의 표, 2026-11-27 `6bd9720827` 에서 읽음).

이 SPEC 은 그 쓰기 지점들이 부르는 검사 함수 하나(`home_write::check`)를 만들고, 명령을 쓰기 전에 멈춰 세운 뒤 `moai update` 의 정리 단계를 안내해요. 정리 명령 자체(`moai update --confirm-home` 등)와 미리보기·읽기 명령은 막지 않아요. 새 쓰기 지점이 검사를 빼먹으면 실패하는 소스 스캔을 함께 넣어요. 이 카드가 랜딩되기 전에는 C-199 의 해석기 고정(REQ-16)이 랜딩되지 않아요.

판정 기준은 `state_root::classify_home` 의 세대 이름이에요. 거절은 `go`, `mixed`, `interrupted`, `refuse` 넷이고 `none`, `empty`, `unknown`, `rust`, `rust-first` 는 통과해요. `unknown` 이 통과해야 하는 까닭은 Go 정리(retire) 직후의 `~/.moai` 가 `unknown` 또는 `empty` 라서, 정리를 마친 사용자가 다시 쓸 수 있어야 하기 때문이에요.

이 SPEC 은 tier 2(HOME, 게이트)이고 코드 전에 사용자의 승인이 필요해요. 이 초안은 코드를 바꾸지 않아요.

## Non-goals

- 해석기를 `~/.moai` 로 고정하는 일(C-199)과 이동기 `home-install-*` 삭제. 이 SPEC 은 그 카드가 쓸 검사만 먼저 놓아요.
- `moai update` 정리 단계(`--confirm-home`, `home-retire-*`)의 동작 변경. 이 단계는 막지 않아요.
- `backup_root` 와 `check_home_ready` 의 글과 동작 변경. 그대로 둬요.
- 새 명령, 새 도움말 항목, 설정 변경.
- 프로젝트 쪽 `<project>/.moai` 쓰기와 `~/.claude` 안의 쓰기 자체(`moai update` 가 한 번에 쓰는 묶음의 입구 검사는 REQ-5).
- 실제 `~/.moai`, `~/.moai-cli` 읽기. 모든 시험은 임시 HOME 이에요.

## Requirements

<!--
Rule 1: exactly one `shall` per requirement line; a second behavior goes on its own REQ line (spec check refuses it as req_gears).
Rule 2: every When requirement needs an IF/THEN requirement that names it in parentheses, or its line ends with "· failure: none — reason" (spec check refuses it as req_failure_path).
-->
<!-- This check starts in observe: record the would-block count and the false positives in the Result, and leave the switch to enforce to a recorded decision. -->
- REQ-1: The `home_write::check` function shall classify the HOME it is given with `state_root::classify_home` and refuse the generations `go`, `mixed`, `interrupted` and `refuse` and no other.
- REQ-2: The refusal of `home_write::check` shall say that nothing was written, name the way out that clears the refused state for its generation and its reason, and use no internal term.
- REQ-3: When a command in the write-point table writes HOME state, the command shall call `home_write::check` before its first write.
- REQ-4 (REQ-3): IF `home_write::check` refuses, THEN the command shall stop with a non-zero exit before it writes anything under HOME.
- REQ-5: When `moai update` is applied, the update shall call `home_write::check` once before its first write of any kind.
- REQ-6 (REQ-5): IF `home_write::check` refuses, THEN the update shall stop as a whole with a non-zero exit before it writes under HOME or in the Claude settings folder.
- REQ-7: A plan command or a preview command or a read-only command shall not call `home_write::check`.
- REQ-8: A clean-up command in the exempt table shall not call `home_write::check`.
- REQ-9: A command in the write-point table shall write exactly what it wrote before this change on a HOME of generation `none`, `empty`, `unknown`, `rust` or `rust-first`.
- REQ-10: The refusals of `backup_root` and of `check_home_ready` shall keep their text and their behaviour.
- REQ-11: The `home_write_scan` test shall report one line for each production use of a HOME-root function that its table does not list, for each entry row whose function lacks the check, and for each table row that the source no longer has.
- REQ-12: The `home_write_scan` test shall fail on a reported line only in its enforce mode, which a recorded decision switches on.

## Acceptance

- AC-1 (REQ-1): 아홉 세대 각각의 임시 HOME 에서 `check` 가 거절 넷만 `Err` 로 답하고, `backup_root` 와 `check_home_ready` 가 같은 열 가지 HOME 에서 같은 거절 집합을 내요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_write)'`
- AC-2 (REQ-2): 거절 글이 "아무것도 쓰지 않았어요"로 시작하고, `go`·`mixed` 는 `moai update` 의 Go 정리 단계(`--confirm-home`)를, `interrupted` 는 옛 폴더를 손으로 되돌리는 길(HOME 단계에 풀어 주는 명령이 없어요, Plan 의 표)을, `refuse` 는 이유 다섯(링크, 폴더가 아닌 것, 더 새로운 표식, 읽을 수 없는 표식, 폴더 모양 오류)마다 그 이유가 고칠 것을 말하고, "verdict" 같은 내부 말이 없어요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(the_refusal_text_names_the_way_out_per_class)'`
- AC-3 (REQ-3, REQ-4): 표의 명령이 `go`, `mixed`, `interrupted`, `refuse` 임시 HOME 에서 모두 0 이 아닌 종료 코드로 멈추고, 실행 전후 `TempHome::snapshot` 이 같고(새 `.moai-cli` 폴더가 없어요), 거절 글이 AC-2 의 글이에요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(a_refused_home_stops_every_listed_write_command_before_its_first_write)'`
- AC-4 (REQ-5, REQ-6): 거절 HOME 에서 `moai update --confirm <digest>` 가 통째로 멈춰 HOME 과 Claude 설정 폴더(`.claude`)에 아무것도 쓰지 않고, 모델 단계(`update_models::stage_with`)도 같은 글로 멈춰요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(update_confirm_stops_as_a_whole_on_a_refused_home)'`
- AC-5 (REQ-7): 같은 임시 HOME 에서 계획·미리보기·읽기 명령(`moai update` 확인 없이, `*-plan`, `ledger-status`, `key-status`)이 종료 0 으로 답하고 snapshot 이 같아요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(plan_and_preview_commands_still_answer_on_a_refused_home)'`
- AC-6 (REQ-8): Go HOME 에서 `home-retire-plan` 과 `home-retire-apply` 가 거절 없이 돌고, 그 뒤 같은 `key-set` 이 통과해요(정리 뒤 `unknown` 이나 `empty` 가 통과함을 같이 보여요). · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(a_go_home_can_use_the_clean_up_step_and_then_write)'`
- AC-7 (REQ-9): `none`, `empty`, `unknown`, `rust`, `rust-first` HOME 에서 표의 명령이 이 변경 전과 같은 종료 코드와 같은 파일을 만들어요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(a_ready_home_is_written_as_before_by_every_listed_command)'`
- AC-8 (REQ-10): `backup_root` 와 `check_home_ready` 의 거절 글이 바이트 단위로 그대로이고 기존 레저 테스트가 통과해요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(the_four_existing_refusals_keep_their_exact_text)'`
- AC-9 (REQ-11): 스캔이 표에 없는 사용, 검사를 부르지 않는 입구 행, 소스에 없는 표 행을 각각 한 줄로 알리고, 주석·문자열·`#[cfg(test)]` 안은 세지 않아요. · verify: `cargo nextest run --offline --locked --test home_write_scan -E 'test(an_unlisted_root_use_is_reported_and_enforce_fails)'`
- AC-10 (REQ-11, REQ-12): 체크인된 소스가 표와 같고(관찰 모드는 보고하고 통과, enforce 는 다르면 실패), 스캔이 레지스트리에 올라 있어요. · verify: `cargo nextest run --offline --locked --test home_write_scan -E 'test(every_home_root_use_in_product_code_is_classified)'`
- AC-11 (REQ-4): 위험한 길 실행. 실제 바이너리를 임시 HOME 에서 돌려, 거절 넷, HOME 의 상태 폴더(`.moai`)가 HOME 안 폴더를 가리키는 링크, 링크 대상이 다른 임시 폴더인 경우, Go HOME 안에 열린 파일이 있는 경우에도 멈추고, 모든 snapshot 이 같고 링크 대상이 그대로예요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(the_dangerous_path_runs_stop_and_change_nothing)'`
- AC-12 (REQ-9): 이 카드의 범위로 `moai check affected` 를 맨몸으로 한 번 돌린 결과 요약이 `result=PASS` 이고, 거절 HOME 에서 쓰기 명령을 돌리던 기존 테스트는 통과하거나 의도해서 고쳐요(고친 것은 Result 에 적어요). · verify: `moai check affected <base>..<head>`
- AC-13 (REQ-2, REQ-8): 거절에서 풀림까지 한 줄로 이어지는지 `interrupted` 로 보여요. `.moai-backup` 만 있고 `.moai` 가 없는 HOME 에서 `key-set` 이 거절되고, 글이 말한 대로 시험이 `.moai-backup` 을 `.moai` 로 되돌린 뒤(사용자의 손 단계) `moai update` 미리보기가 정리 단계와 토큰을 보이고, `--confirm-home <토큰>` 만 적용하면 같은 `key-set` 이 통과해요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(an_interrupted_home_gets_out_by_the_stated_steps_and_then_writes)'`
- AC-14 (REQ-2): `refuse` 이유 다섯 각각의 HOME 에서 글이 그 이유가 고칠 것을 말하고, 시험이 그 한 가지를 고치면(링크를 지우고 폴더로 두기, 파일을 옆으로 옮기기, 읽을 수 없는 표식 파일을 옆으로 옮기기; 더 새로운 표식과 폴더 모양 오류는 글만 확인) 같은 명령의 거절이 풀려요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(each_refuse_reason_names_what_to_fix_and_the_fix_clears_it)'`
- AC-15 (REQ-5, REQ-6, REQ-8): `go`·`mixed` 임시 HOME 에서 `moai update`(확인 없이)의 미리보기가 정리 단계의 토큰과 칠 명령(`--confirm-home <토큰>`)을 보이고, `moai update --confirm <digest> --confirm-home <토큰>` 을 함께 주면 아무것도 쓰기 전에 멈추고 글이 "정리 단계만 먼저"라고 말하고, `--confirm-home <토큰>` 만 주면 적용되고, 그 뒤 새 미리보기의 `--confirm <digest>` 가 통과해요. · verify: `cargo nextest run --offline --locked --test home_write_cli -E 'test(a_go_home_update_runs_the_clean_up_first_and_then_confirms)'`

## Scope

- write: `.moai/specs/SPEC-20261127-home-write-check/spec.md`
- write: `src/home_write.rs` (새 파일: `check`, 오류 타입, 거절 글)
- write: `tests/home_write_scan.rs`, `tests/home_write_cli.rs` (새 파일)
- write: `src/lib.rs`(`pub mod home_write;` 한 줄), `tools/repo_scans.json`(항목 하나): 공유 파일이라 `card line-lease` 로 짧게 쓰고 커밋 뒤 바로 돌려줘요.
- write (조각 2, 업데이트 묶음): `src/update.rs`(`apply_update_in` 의 `adopt_new` 앞, `switch::apply` 호출, `revert_switch` 호출), `src/update_models.rs`, `src/ops/bin_names_apply.rs`
- write (조각 3, 설정·영수증·레저·키 쓰기): `src/agent_install.rs`, `src/home_style.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs`, `src/language.rs`, `src/star_ask.rs`, `src/gate_schedule.rs`, `src/claim.rs`, `src/ledger/open.rs`, `src/ops/key_set.rs`, `src/ops/key_remove.rs`, `src/ops/key_import_apply.rs`, `src/ops/keys_revert.rs`
- write (조각 4, 나머지): `src/user_assets.rs`, `src/hygiene.rs`, `src/session/machine_names.rs`, `src/build_number.rs`, `src/reports_move.rs`, `src/worktree_gc.rs`, `src/home_state_retire.rs`, `src/layout.rs`, `src/hooks.rs`, `src/edit_diagnostics.rs`, `src/recovery.rs`, `src/worktree_move.rs`. 파일마다 잡은 카드가 있어서, 조각이 시작할 때 `moai card check` 를 다시 하고 잡은 카드에 먼저 물어요(Plan 의 순서).
- write: 위 파일을 읽는 기존 테스트 중 거절 HOME 에서 쓰기 명령을 돌리던 것(`tests/update_conflict_baseline_cli.rs` 의 ac12 등)은 AC-12 의 `check affected` 가 고른 만큼만.
- read-only: `src/state_root.rs`, `src/home_root.rs`, `tests/common/home.rs`, `tests/home_env_readers.rs`
- forbidden: `src/key_vault.rs`, `src/update_keys.rs`, `src/main.rs`

## Plan

조사 기준: 2026-11-27 17:20Z 전후, NeXT tip `6bd972082704ad90376583852847f3b38cc682bd`, 읽기 전용 에이전트 여섯과 직접 grep. 제품 코드에서 루트 함수(`home_root::path|dir`, `home_workdir::root|projects_root|reports_*|worktrees_root|service_dir`, `home_layout::rust_path`, `layout::default_home_root|switched_dir`)를 쓰는 곳은 테스트 모듈을 빼고 89곳, 파일 42개예요(목록은 레인 scratch `target/jade/c613-sites.txt`, 승인 뒤 구현 조각 1 이 표로 옮겨요). 파일과 줄 번호는 이 tip 의 것이고 구현 전에 다시 세요.

**쓰기 지점 표 (함수 단위, 검사 자리)**

| 묶음 | 쓰기 지점(file:line) | 검사 자리 |
| --- | --- | --- |
| 업데이트 | `src/update.rs:3241`(`apply_update_in`, `adopt_new` 앞), 호스트 영수증 `:735`·`:785`·`:907`·`:812`, `run_receipt.rs:480`, `extra_skill.rs:551/736/789`, `switch.rs:108/449/544` | `apply_update_in` 입구 한 번(REQ-5). 안쪽 부분은 `Under(apply_update_in)` 로 분류, 따로 도는 부분은 조각 3 에서 제 검사 |
| 업데이트(예외) | `update.rs:2685`(`--confirm-switch` 단독), `:3006`(`--revert-switch`), `update_models.rs:33`·`:166`·`:221`(`stage`, 미리보기 뒤 터미널에서도 `update.rs:3043` 으로 열려요) | 호출 앞, 그리고 `stage_with` 첫 문장 |
| 설정·영수증 | `agent_install.rs:74/483/571/941`, `home_style.rs:440/554/1057/1230/1491`, `home_routing.rs:208/398/450/529`, `home_routing/rules.rs:96/120/803/841/891/969`, `language.rs:352`, `star_ask.rs:72/140`, `gate_schedule.rs:373/515`, `layout.rs:53/342/402` | 쓰기 입구 첫 문장 |
| 키 | `key_vault.rs:261/692/702/804/1051/1127` | 명령 쪽(`ops/key_*.rs`)에서 검사. `key_vault.rs` 는 C-199 가 쥐어서 건드리지 않아요 |
| 클레임 | `claim.rs:105`→`Journal::open`(`:826`), claim·worktree-apply·scope·pr-link·handoff 가 모두 거쳐요 | `Journal::open` 한 곳 |
| 레저 | `ledger/open.rs:146 open_write`(모든 `moai card` 쓰기, `card-import-apply`), `ledger/folders.rs:189`·`open.rs:266`·`recover.rs`(이미 `check_ready` 로 막힘) | `open_write` 에 더하고 글은 기존 `HomeNotReady` 그대로 |
| 훅 | `hooks.rs:615`(`append_receipt_line_in`, 호출자 `hook_engine.rs`·`worktree_move.rs:652`·`worktree_slot.rs:486`·`audit_slot.rs:575`·`worktree.rs:671`), `edit_diagnostics.rs:152` | 최선 노력 쓰기: 거절이면 HOME 쓰기를 건너뛰고 종료 코드는 그대로(D-8) |
| 해석기 밖 | `user_assets.rs:287/412/681`, `hygiene.rs:340`, `session/machine_names.rs:30/76/127/229`, `build_number.rs:147/209/273`, `reports_move.rs:410/427/475`, `tidy/apply.rs:98`, `worktree_move.rs:170/304`, `worktree_gc.rs:862`, `session.rs`, `recovery.rs:113/454`, `home_state_retire.rs:375/552` | 조각 4: 줄을 읽은 뒤 검사 또는 `unknown` 행(D-15) |
| 이미 검사 있음 | `update.rs:3784/3926`(`backup_root`), `home_routing/rules.rs:423`(`backup_root`), `home_install_apply.rs:752`, `home_state.rs:762/906`, `ledger` 쓰기 | 그대로(REQ-10) |

**제외 표 (정리 묶음, 막지 않아요, REQ-8)**: `update --confirm-home`·`home-retire-apply/revert`(쓰기는 `~/.moai-go-retired/<UTC>/`; Go 항목을 `~/.moai` 에서 이름만 바꿔 내보내요, 스스로 `interrupted`·`refuse` 를 막아요, `src/home_retire.rs:254`), `home-state-apply/revert`(스스로 판정을 검사해요: apply 는 `go` 만, revert 는 `rust` 만), `home-install-apply/revert`(C-199 결정 B 로 지워질 때까지 제외, D-5), `go-state-*`(프로젝트 쪽). 행마다 `~/.moai` 안에 쓰는지 밖에 쓰는지 적고, 안이면 C-199 에 알려요. `cleanup-apply/revert`·`update --confirm-cleanup` 와 `home-archive-*` 는 제외 표에 두지 않고 검사를 걸어요(D-3, D-4).

**하나의 함수**: `src/home_write.rs` 의 `pub fn check(home: &Path) -> Result<(), HomeWriteRefused>`. 오류는 `HomeWriteRefused { generation, reason }` 이고 `Display`, `std::error::Error`, `From<HomeWriteRefused> for String` 을 가져요(`Result<T, String>` 함수가 많아서 `?` 로 이어져요). 판정은 `backup_root`·`check_home_ready` 와 같은 `classify_home(home).generation` 이에요. HOME 폴더가 없거나 폴더가 아니면 통과시켜서 각 쓰기가 제 오류를 그대로 내요(레저만 "the HOME folder does not exist" 를 그대로 가져요). 환경에서 HOME 을 읽지 않아요. 거절 글은 아래 "세대별로 푸는 길" 표의 마지막 칸을 그대로 쓰고(승인 뒤 글자를 다듬어요), 영어와 한국어 두 판으로 두며 `cli_text_citations` 가 막는 카드·결정 번호는 넣지 않아요. 모두 "아무것도 쓰지 않았어요"로 시작하고 다 읽고 나면 사용자가 칠 명령이나 손으로 할 일이 한 가지로 정해져요.

**세대별로 푸는 길** (코드를 읽어 확정했어요, 2026-11-28 `6bd9720827`). 거절 글이 가리키는 것은 여기서 실제로 그 상태를 푸는 것뿐이에요.

| 세대 | 그 판정을 만드는 표식 | 푸는 길(읽은 곳) | 거절 글이 말할 것 |
| --- | --- | --- | --- |
| `go`, `mixed` | Go 표식 파일(`user-assets.json`, `state/active-sessions.json`, `credentials.yaml`)이 있고 이 세대 표식이 없거나(`go`, `state_root.rs:640`) 둘 다 있어요(`mixed`, `:633`) | `moai update` 가 미리보기에 정리 단계와 토큰을 보여 줘요(`update.rs:2596`, `update_home::plan`) → `moai update --confirm-home <토큰>` 만 적용해요(`update.rs:2678-2684`, `home_retire.rs:208`이 `go`·`mixed`·`rust-first` 에서 움직여요). 끝나고 나면 HOME 이 어느 세대가 되는지는 읽은 것만 말해요: 남는 것은 `keep` 항목(사용자 데이터와 이 세대 항목)뿐이라 거절 네 세대는 아니에요(AC-15 가 실제로 돌려 확인해요). Go 세션이 살아 있거나 잠금·목록을 읽을 수 없으면 정리가 막혀요(`home_retire.rs:235-247`: `go_session_live`, `session_lock_unknown` 등) | "Go 세대 파일이 남아 있어요(<표시>). `moai update` 를 실행해 보이는 정리 단계의 명령 `moai update --confirm-home <토큰>` 만 먼저 실행하고, 끝나면 `moai update` 를 다시 실행해 주세요. 미리보기가 Go 세션이 돌고 있다고 하면 그 세션을 닫고 다시 해 주세요" |
| `interrupted` | `.moai-backup`(또는 `.moai-backup-<UTC>`)이 있는데 `.moai` 가 없는 HOME(`state_root.rs:354-365`, `:582-584`), 또는 `.moai` 가 끝나지 않은 `go-state` 영수증을 품은 HOME(`:369-373`) | HOME 단계에 이 상태를 푸는 명령이 코드에 없어요. `home-state-revert` 는 `rust` 만 받고(`home_state.rs:915-921`), `--confirm-home` 은 `interrupted`·`refuse` 에서 스스로 막혀요(`home_retire.rs:254`), `go-state-revert` 는 `go-state-receipt.json` 이 있어야 읽어요(`go_state.rs:5492`)인데 HOME 쪽 `home-state-apply` 는 다른 영수증 `home-state-receipt.json` 을 써요(`home_state.rs:700`). 그래서 첫째 모양(`.moai-backup` 만 있음)은 사용자가 `.moai-backup` 이름을 `.moai` 로 손으로 되돌리면 HOME 이 `go` 로 읽히고 위 정리 길로 이어져요. 둘째 모양(`go-state` 영수증)은 `moai go-state-revert-plan --backup <UTC> <폴더>` 가 푸는 길이에요 | 첫째 모양: "~/.moai 를 옮기다 멈춘 HOME 이에요. Go 폴더가 `~/.moai-backup`(또는 `~/.moai-backup-<UTC>`)에 그대로 있어요. 그 폴더 이름을 `~/.moai` 로 되돌린 뒤 `moai update` 를 실행해 주세요." 둘째 모양: "끝나지 않은 옮기기 기록이 `~/.moai` 에 있어요. `moai go-state-revert-plan --backup <UTC> <폴더>` 로 되돌릴 것을 먼저 확인해 주세요" |
| `refuse` | (1) `.moai` 가 링크(`state_root.rs:607`) (2) 폴더도 링크도 아님(`:608`) (3) 더 새로운 표식(`:610-617`) (4) 읽을 수 없는 표식(`:618-621`) (5) schema-2 표식인데 폴더 모양이 틀림(`:622-631`, 이유는 `home_layout::check_root_in` 이 줘요) | 손으로 고치는 길뿐이에요. `moai update` 의 정리 단계는 막혀요(`home_retire.rs:254`) | 이유마다: (1) "`~/.moai` 가 링크예요. 링크는 따라가지 않아요. 링크를 지우거나 진짜 폴더로 바꿔 주세요" (2) "`~/.moai` 가 파일이에요. 그 파일을 옆으로 옮겨 주세요" (3) "`~/.moai` 를 더 새로운 moai 가 썼어요. 그 moai 를 쓰거나 폴더를 옆으로 옮겨 주세요" (4) "`~/.moai` 안의 표식 파일을 읽을 수 없어요. 그 파일을 고치거나 옆으로 옮겨 주세요" (5) 폴더 모양 오류의 이유를 그대로 보여 주고 "그 항목을 고치거나 옮겨 주세요" |

위 표에서 `interrupted` 만 사용자가 낼 수 있는 HOME 단계의 명령이 없어요. 새 되돌리기 명령을 만드는 것은 이 카드의 범위가 아니에요(D-7). 정리가 `moai update` 흐름 안에서 막다른 길이 되지 않게 하는 순서는 이래요: 거절 글은 항상 정리 단계만 먼저라고 말하고, `--confirm` 과 `--confirm-home` 을 함께 준 호출은 입구 검사가 아무것도 쓰기 전에 멈추며(D-17, AC-15), 정리가 끝난 HOME 은 통과하는 세대라 `--confirm` 이 그다음에 돌아요. `home-archive-*` 는 정리 길과 독립이에요: 정리(`--confirm-home`)는 `~/.moai-cli` 를 보관하지 않아도 돌고(`home_retire.rs` 에서 `.moai-cli` 폴더를 읽는 곳을 못 찾았어요: grep 에 걸리는 것은 `:738` 의 파일 이름 `moai-cli.json` 하나뿐이에요), 거절 HOME 에서 보관을 하려는 사용자는 같은 거절 글로 정리 먼저를 안내받아요. C-199 가 보관 단계를 `moai update` 흐름에 거는 한 줄(`src/update.rs`)은 그 카드의 자리예요. 걸 때는 정리 뒤에 보관이 오게 해야 막다른 길이 되지 않아요(이 SPEC 의 요구가 아니라 C-199 에 알리는 줄이에요).

**소스 스캔 `tests/home_write_scan.rs`**: 선례 `tests/home_env_readers.rs`(토큰 걷기로 둘러싼 함수를 잡고, 상수 표와 정확한 개수로 비교, 관찰 기본). 행은 `(파일, 함수, 개수, 종류)` 이고 종류는 `Entry`(함수 몸에 `home_write::check(` 나 이름 붙은 제 관문 `check_ready`·`check_home_ready`·`backup_root` 가 있어야 하고 `cli:` 이름이 `tests/home_write_cli.rs` 에 있거나 `no_runtime: <이유>`), `Under`(`Entry` 행을 가리켜요), `Read`, `WayOut`(안/밖 기록), `Resolver`, `BestEffort`(검사 필요)예요. 토큰은 모듈 이름이 붙은 호출만 세요(맨 이름 `use` 는 따로 보고해요). `MOAI_DIR`·`HOME_STATE_DIR` 직접 이음은 같은 문장에 `home` 이 든 이름이 있을 때만 세는 어림이라 관찰 모드에서 오탐을 기록해요(이 상수는 프로젝트 쪽에서 60개 파일가량이 써요). 주석, 문자열, `#[cfg(test)]` 는 건너뛰어요. 최소 개수 아래면 실패해요. 모드는 `MOAI_HOME_WRITE_CHECK=enforce` 가 켤 때만 실패하고, 기본 관찰은 한 줄씩 stderr 에 `home-write-check observe: ...` 로 써요. 등록은 `tools/repo_scans.json` 에 `home-env-readers`(`:3128`) 항목을 본뜬 하나이고 시험은 `home_write_scan` 이에요. 새 테스트 파일 이름의 첫 낱말 `home` 은 `tests/areas.toml` 에 이미 있어요.

**임시 HOME 시험 `tests/home_write_cli.rs`**: 실제 바이너리와 `TempHome::command`(깨끗한 환경, HOME 이 임시 폴더)와 `TempHome::snapshot`(정렬된 항목, 바이트, 시각)을 써요. 거절 HOME 은 기존 도우미 `make_go_generation`·`make_mixed_generation`·`make_interrupted_generation`·`make_refused_generation` 이고 `tests/common/home.rs` 는 건드리지 않아요. 통과 HOME(`none`, `empty`, `unknown`, `rust`, `rust-first`)은 `moai_cli::home_root::MOAI_DIR` 와 `state_root::RUST_MARKER` 로 시험 안에서 만들어요(`.moai` 글자를 쓰지 않아요: `state_folder_literal_scan`). 각 행은 종료 코드, 글, 전후 snapshot 같음, 같은 명령이 `rust` HOME 에서는 전과 같이 도는지를 봐요. 표는 약 40 명령(조각 4 에서 `moai --help` 로 확정), 첫 최소 표는 여덟 행: `update --confirm <digest>`, `update-apply`(자리 인자, 검사가 먼저), `key-set glm`(값은 stdin), `key-remove glm --yes`, `language-set --home en`, `star-ask --dismiss`(종료 0 과 `written: false` 와 사유), 이미 레저가 있는 HOME 이 나중에 mixed 가 된 뒤의 `card add`, `update_models::stage_with`(프로세스 안). 업데이트 행은 패키지 폴더를 HOME 밖(`<tmp>/versions/v1/bin/moai`)에 복사해 두고, 미리보기(Go HOME 에는 아무것도 안 써요)에서 digest 를 받아 `--confirm` 을 돌려요.

**위험한 길 실행**: 선례는 `SPEC-20261123-home-state-retire` AC-9 와 `SPEC-20261124-native-role-link-refusal` 이에요. 실제 바이너리를 임시 HOME 에서 돌려 AC-11 의 다섯 경우를 확인하고, 결과를 Result 에 한 문단으로 적어요. 스크립트나 Python 은 쓰지 않아요(Rust 시험만).

**조각과 순서**
1. 조각 0: 이 SPEC 승인(사용자).
2. 조각 1(기준선): `src/home_write.rs`, `src/lib.rs` 한 줄, `tests/home_write_scan.rs`(관찰 모드), 레지스트리 항목, 함수 단위 시험(AC-1). 잡힌 파일이 없고 어떤 게이트도 더하지 않아서 먼저 커밋해 would-block 개수를 기준선으로 남겨요.
3. 조각 2(업데이트 묶음): 먼저 `tests/home_write_cli.rs` 의 `update` 행을 RED 로 두고, `apply_update_in`, `switch`, `update_models` 를 거쳐요. C-604 가 `src/update.rs` 를 쥐고 있어서 그 카드의 질문이 먼저예요.
4. 조각 3(설정·영수증·레저·키): 쓰기 입구 첫 문장에 검사. 이미 레저가 있는 거절 HOME 행이 여기 있어요.
5. 조각 4(나머지와 질문의 기본값): 파일을 잡은 카드(C-199, C-557, C-575, C-410, C-584, C-300 등)에 먼저 물어요.
6. 조각 5(enforce): 관찰 모드가 통합 tip 에서 두 번 연속 0 줄이고, 모든 `Entry` 행에 실행 행이 있거나 `no_runtime` 이유가 있고, 리더가 결정을 기록하면 켜요. 켤 때 환경 변수 분기를 없애고 Result 에 would-block 개수와 오탐을 적어요.

랜딩 순서: 조각 1 → 2 → 3 → 4 → 5 가 모두 C-199 의 해석기 고정(REQ-16)보다 먼저예요. 조각 1 은 잡힌 파일이 없어서 바로 랜딩할 수 있어요.

**겹침**(2026-11-27 원장의 예약): C-199(`key_vault.rs`, `home_root.rs`, `state_root.rs`, `update_keys.rs`, `session.rs`, `worktree_move.rs`, `tests/common/home.rs`, `tests/path_root_direct_sites.rs`), C-557(`hooks.rs`, `hook_engine.rs`), C-604(브랜치가 `update.rs`, `install.rs`, `bin_names.rs` 를 바꿔요), C-575(`recovery.rs`), C-579(`bin_names.rs`, `update/switch.rs`), C-410(`layout.rs`), C-409(`identity.rs`, `home_state.rs`), C-322(`tidy/`), C-584(`session.rs`), C-300(`main.rs`), C-298(`state_root.rs`). 이 SPEC 의 새 파일(`home_write.rs` 와 두 테스트)과 조각 1 은 겹치지 않아요. 구현 시작 전에 `moai card check` 를 다시 해요.

## Risks

- `moai update --confirm` 이 거절 HOME 에서 통째로 멈추면(D-1) 지금 `ac12` 가 지키는 동작(스킬은 깔고 백업 단계만 거절)이 바뀌어요. 그 테스트와 비슷한 테스트를 의도해서 고치고, v3 사용자는 `--confirm-home` 을 먼저 마치고 `--confirm` 을 다시 해야 해요.
- 이미 레저가 있는데 `~/.moai` 가 거절 세대로 판정되면 모든 `moai card` 쓰기가 정리가 끝날 때까지 멈춰요. 이 저장소의 레저 쪽 HOME 이 그렇게 판정되면 리더가 잠길 수 있어요. 레저 행(`card add`)이 정확히 그 경우를 시험해요.
- Go 세대 `~/.moai` 의 사용자는 정리 전까지 `key-set`, `language-set`, 업데이트 설치를 못 써요. 정리가 끝난 HOME 은 거절 네 세대가 아니라서 통과하고(AC-15 가 실행으로 확인해요), 레저 등록만 `registration_blocked` 가 `unknown` 에서 계속 막아요(D-11).
- 정리가 막히는 경우(Go 세션이 살아 있거나 잠금·목록을 읽을 수 없음, `home_retire.rs:235-247`)에는 그 HOME 의 쓰기가 막힌 채 사용자가 그 세션을 닫을 때까지 기다려야 해요. 거절 글이 그 말을 하고, 미리보기가 막힌 이유를 보여 줘요.
- `interrupted`(`.moai-backup` 만 있는 HOME)는 HOME 단계에 푸는 명령이 없어서 손으로 이름을 되돌려야 해요(D-7). 이 SPEC 이 새 명령을 넣지 않는 대신 AC-13 이 그 손 단계가 정리 길로 이어지는지 실제로 확인해요.
- 훅 영수증, 편집 진단, 런처 이름 선점은 호스트의 도구 호출이나 시작을 실패시키면 안 돼요(D-8, D-9).
- 스캔은 낱말 단위라서 `home` 이 아닌 이름의 변수에 든 HOME 과 맨 이름 이음을 놓칠 수 있어요. `Under` 행은 사람의 주장이고 실행 표가 확인해요. 관찰 모드가 오탐을 기록하고 enforce 전에 표를 다듬어요.
- 읽지 못한 쓰기(`worktree_move`, `worktree_gc` 안쪽, `layout::switched_dir_checked` 호출자, `tidy`)는 `unknown` 행으로 두어 놓치지 않게 하되, 안 읽은 것을 안전하다고 적지 않아요(D-15).
- 계획 에이전트가 새로 찾은 사실: `home_state_retire`(`home-archive-*`)는 `classify_home` 을 아예 부르지 않고 `~/.moai/archive` 에 써요(D-3). 모델 단계는 미리보기 뒤에도 터미널에서 열려 키 금고와 `launcher.yaml` 에 써요(`update.rs:3043`).

## Decisions

각 결정은 이 레인이 정한 기본값이에요. 사용자가 이미 정한 것은 D-0 하나뿐이에요. 나머지는 승인 때 한 번에 보시고 바꾸실 수 있어요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-0 | 쓰기 검사를 먼저 넣고, 해석기 고정은 이 검사가 랜딩된 뒤에만 해요 | 사용자 결정(C-199 노트 10937 답 1, 2026-11-27) | 사용자만 바꿔요 |
| D-1 | `moai update --confirm` 은 거절 HOME 에서 통째로 멈춰요 (대안: HOME 쪽 쓰기만 멈추고 `~/.claude` 스킬은 깔아요) | 결정의 글("쓰기 전에 멈추고 정리 단계를 안내")을 가장 단순하게 지켜요. 부분 적용은 영수증이 반쯤 남는 상태를 만들어요. `ac12` 와 부딪히는 점은 위험으로 적었어요 | 입구 검사를 HOME 쓰기 앞으로 옮기는 한 조각 |
| D-2 | `moai update` 미리보기 JSON 키는 늘리지 않아요(알림 줄 없음). 순서는 거절 글이 알려 줘요 | 미리보기 출력을 읽는 곳이 많아서 키 추가는 별도 일이에요 | 알림 줄 하나 추가 |
| D-3 | `home-archive-apply/revert` 에 검사를 걸어요 | 판정 검사 없이 `~/.moai/archive` 에 쓰고 `~/.moai-cli` 를 옮겨요. 정리 길이 아니에요(이 SPEC 이 새로 찾았어요). 리더 결정(2026-11-28, 승인 전 초안): 기본값대로. 보관은 정리와 독립이라 정리 먼저가 막다른 길이 되지 않아요(Plan 의 표 아래 문단) | 제외 표로 옮김 |
| D-4 | `cleanup-apply/revert` 와 `update --confirm-cleanup` 에 검사를 걸어요 | 정리 길은 `--confirm-home` 이고, 이 쪽은 `<state>/cleanup` 에 써요(해석기 고정 뒤 `~/.moai` 안). 리더 결정(2026-11-28, 승인 전 초안): 기본값대로 | 제외 표로 옮김 |
| D-5 | `home-install-*` 는 C-199 결정 B 로 지워질 때까지 제외해요 | 지워질 명령이에요. 그 카드가 지울 때 표 행도 같은 변경에서 지워요. 리더 결정(2026-11-28): 그대로 | 검사 걸기 |
| D-6 | `recovery-apply` 에 검사를 걸어요 | 리더 결정(2026-11-28): 기본값대로. 거절 HOME 에서는 업데이트가 통째로 멈춰(D-1) 복구할 것이 생기지 않아요 | 제외 표로 옮김 |
| D-7 | `interrupted` 는 HOME 단계의 되돌리기 명령 없이, 글이 손으로 되돌리는 길(첫째 모양)과 `go-state-revert-plan`(둘째 모양, 영수증이 있을 때)을 말해요. 새 되돌리기 명령은 이 카드에 넣지 않아요 (대안: `home-state-revert` 가 `interrupted` 도 받게 하는 별도 조각) | 코드를 읽어 확정했어요(Plan 의 "세대별로 푸는 길"): HOME 단계의 푸는 명령이 없어요. 새 명령은 이름 바꾸기를 하는 tier 2 코드라서 이 쓰기 검사와 한 카드에 섞지 않아요. 사용자 확인 필요 | 별도 카드로 되돌리기 명령을 만들면 글과 AC-13 의 사용자 손 단계가 그 명령으로 바뀜 |
| D-8 | 훅 영수증과 편집 진단은 거절이면 HOME 쓰기만 건너뛰고 종료 코드는 그대로예요 | 이 쓰기가 호스트의 도구 호출을 실패시키면 안 돼요. 리더 결정(2026-11-28): 그대로. `hooks.rs` 는 C-557 이 쥐고 있어요 | 거절 오류로 바꿈 |
| D-9 | 런처 이름 선점(`machine_names.rs`)은 거절 HOME 에서 평이한 글과 함께 시작을 거절해요. 이름 선점은 `-l`·`-a` 로 시작할 때만 일어나요(`src/main.rs:1167`, `:1247`: `machine_wide` 는 첫 입력이 있을 때만) | 검사 없는 이름 잠금이 중복 이름을 막는 유일한 장치예요. 사용자에게 보이는 영향이 커서 승인 때 따로 확인해 주세요. 사용자 쪽 모습: (a) 기본: 거절 HOME 에서 `moai cc -l` 과 `moai cc -a` 가 "아무것도 쓰지 않았어요 … `moai update` 를 실행해 정리 단계를 마친 뒤 다시 `moai cc -l` 을 실행해 주세요"를 내고 시작하지 않아요. 이름 없는 `moai cc` 는 영향이 없어요. (b) 대안: 이름 선점 없이 시작하고 "이름이 겹치는지 확인하지 못했어요"라는 경고 한 줄만 내요 | (b) 로 바꿈 |
| D-10 | `star-ask --dismiss` 는 종료 0 을 지키고 사유를 `written: false` 로 답해요 | 이 명령의 약속이 "실패해도 0"이에요 | 종료 코드 바꿈 |
| D-11 | `unknown` 은 통과해요. `ledger-register` 는 `registration_blocked` 가 `unknown` 도 막는 지금 모습 그대로예요 | 정리 직후 사용자가 다시 써야 해요. 등록은 별도 문제예요 | `unknown` 거절 추가 |
| D-12 | 검사는 새 파일 `src/home_write.rs` 에 둬요. `home_root.rs`·`state_root.rs` 는 건드리지 않아요 | C-199 가 두 파일을 쥐고 있어요 | 함수를 `home_root.rs` 로 옮김 |
| D-13 | 스캔은 관찰로 시작하고, 통합 tip 에서 두 번 연속 0 줄이고 모든 입구 행에 실행 행이 있고 리더가 결정을 기록하면 enforce 로 바꿔요 | SPEC 틀의 규칙(막는 검사는 관찰에서 시작) | 관찰로 되돌림 |
| D-14 | 모든 조각이 C-199 의 해석기 고정보다 먼저 랜딩해요. 조각 1 이 가장 먼저예요 | D-0 의 순서 게이트 | 순서를 사용자만 바꿔요 |
| D-15 | 읽지 못한 쓰기는 `unknown` 행으로 표에 남겨요(`worktree_move`·`worktree_gc` 안쪽, `layout::switched_dir_checked` 호출자, `tidy`) | 읽지 않은 곳을 안전하다고 적지 않아요 | 읽은 뒤 행을 분류로 바꿈 |
| D-16 | 레저는 `open_write` 에 검사를 더하고 거절 글은 기존 `HomeNotReady` 를 그대로 써요 | REQ-10(기존 거절 글 유지) | 별도 글로 바꿈 |
| D-17 | `moai update --confirm <digest>` 와 `--confirm-home <토큰>` 을 함께 준 호출은 거절 HOME 에서 입구 검사가 아무것도 쓰기 전에 통째로 멈춰요. 정리는 `--confirm-home <토큰>` 만 따로 주고, 그다음에 `--confirm` 을 줘요 (대안: 함께 준 호출에서는 정리 부분을 먼저 적용하고 나머지를 이어서) | 지금 코드는 층을 먼저 쓰고 정리 부분을 그다음에 적용해요(`update.rs` 의 `apply_update_in` 뒤 `home_part`). 입구 검사를 두면 함께 준 호출에서 정리 전에 멈추므로, 글이 "정리 단계만 먼저"라고 말하는 것이 막다른 길을 막는 가장 작은 변경이에요. 순서를 바꾸는 대안은 적용 순서를 건드려 영수증과 되돌리기 순서(`revert_order`)가 바뀌어서 크기가 커요 | 입구 검사 앞에 정리 부분을 먼저 돌리는 한 조각 |

## Interview

- Q1: 쓰기 검사를 해석기 고정 전에 넣나요, 그대로 두나요? → A: 먼저 넣어요. 사용자가 "쓰기 검사를 먼저 넣음 (Recommended)"을 골랐어요(C-199 노트 10937 답 1). 이 SPEC 은 그 별도 SPEC 이고 따로 승인을 받아요.
- Q2: 거절은 어느 세대인가요? → A: `go`, `mixed`, `interrupted`, `refuse`(`backup_root` 와 `check_home_ready` 가 지금 거절하는 넷과 같아요). 이 레인이 정했고 REQ-1·AC-1 이 두 기존 관문과 같은 결과임을 시험해요.
- Q3: `unknown` 은요? → A: 통과(D-11). 정리 직후 HOME 이라서요. 이 레인이 정했어요.
- Q4: 정리 명령은 막나요? → A: 아니요(REQ-8). 제외 표에 이유를 적어요. `home-archive-*`, `cleanup-*` 는 정리 길이 아니라고 읽어 검사를 걸었어요(D-3, D-4). 리더가 기본값대로로 정했어요(2026-11-28, 승인 전 초안).
- Q5: 업데이트는 통째로 멈추나요? → A: 그래요(D-1). 사용자 결정의 글을 가장 단순하게 지키는 쪽이고 `ac12` 가 바뀌어요. 승인 때 확인이 필요한 항목이에요.
- Q6: 새 스캔이 다른 카드의 착지를 막나요? → A: 처음에는 막지 않아요. 관찰로 시작해요(D-13, `home-env-readers` D-13 과 같은 방식).
- Q7: 나머지 결정은 누가 정했나요? → A: D-2~D-17 의 기본값은 이 레인이 정했고(D-3, D-4, D-5, D-6, D-8 은 리더가 2026-11-28 에 승인 전 초안의 기본값대로 확인했어요) 승인 때 한 번에 확인해요. 사용자가 직접 정한 것은 Q1 하나예요.
- Q8: `interrupted` 와 `refuse` 의 HOME 은 쓰기가 멈춘 채 나갈 길이 있나요? → A: `refuse` 는 이유별로 손으로 고치는 길이 있고 글이 고칠 것을 말해요. `interrupted` 는 HOME 단계의 푸는 명령이 코드에 없어서(Plan 의 "세대별로 푸는 길") 손으로 되돌리는 길을 말해요. 그 길은 `moai update` 로 이어지고(AC-13), 새 되돌리기 명령을 만들지는 사용자가 골라요(D-7).
- Q9: 정리와 업데이트를 한 번에 줄 수 있나요? → A: 함께 준 호출은 정리 전에 멈추고 글이 정리만 먼저라고 말해요(D-17, AC-15). 사용자가 다음에 칠 명령은 미리보기 화면의 `--confirm-home <토큰>` 이에요.

## Result

초안이에요. 구현 전이고, 승인은 사용자가 리더를 거쳐 해요. 이 레인은 승인을 기록하지 않아요.
