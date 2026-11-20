---
id: SPEC-20261120-worktree-cap-and-cleanup
title: "Worktree cap per project and cleanup of finished worktrees"
status: draft
revision: 1
created: 2026-11-20
---

# Worktree cap per project and cleanup of finished worktrees

## Intent

사용자 결정(2026-11-20, 리더 대화의 선택)이에요. 에이전트가 카드마다 worktree 를 만들고 치우지 않아서 moai-adk 에는 674개(옛 위치 `.claude/worktrees` 154개, `.moai-cli/worktrees` 181개, 약 200G), mo.ai.kr 에는 456개가 쌓였어요. 이 SPEC 은 모든 MoAI 사용자에게 통하는 두 가지를 정해요: (1) 프로젝트당 worktree 개수의 상한(기본 40, 설정 가능)과 지금 개수를 미리보기에 보여 주기, (2) 끝난 worktree(카드가 done 이거나 가지가 통합 ref 에 병합)를 정리 후보로 미리보기에 올리고, 사용자가 고른 것만 항목마다 안전을 다시 확인한 뒤 지우기. 삭제가 들어가니 사용자 결정 그대로 "미리보기 → 항목별 선택 → 적용 → 되돌림 기록" 순서만 허용해요. 상한은 보여 주기만 하고 만들기를 막거나 지우기를 시키지 않아요. 새로 만들지 않고 있던 `optimize-plan`(`src/optimize.rs`)과 `worktree-gc-plan`(`src/worktree_gc.rs`)의 판정을 그대로 쓰고, 그것이 옛 위치를 "보존"으로 막는 한 가지만 이 SPEC 의 사용자 결정으로 푸는 새 층을 얹어요. 공간은 `du` 가 아니라 할당된 블록(`st_blocks`)과 파일 시스템 여유 바이트(`df`)로 재요(APFS 복제는 `du` 가 두 번 세요).

## Non-goals

- 상한을 어기면 worktree 만들기를 막거나 자동으로 지우는 일: 하지 않아요. 상한은 미리보기의 숫자와 경고예요.
- Codex 가 만든 worktree 와 Go 세대의 `~/.moai/worktrees/projects/`: 계속 보존이에요. 이 SPEC 은 그 둘의 보존 규칙을 풀지 않아요.
- 빌드 폴더(`~/.moai/build`)와 보고서 폴더 정리: 다른 카드의 몫이에요.
- `worktree-gc-apply`(cargo 출력 줄이기와 병합된 worktree 지우기)의 동작 바꾸기: 그대로 두고 새 명령 `optimize-apply` 를 더해요.
- 카드 원장 쓰기: 이 SPEC 의 읽기와 적용은 원장을 읽기만 해요.

## Requirements

- REQ-1: When `optimize-plan` runs, the preview shall report the project's worktree count, the cap and whether the count is over the cap, writing nothing.
- REQ-2 (REQ-1): IF no cap is configured, THEN the preview shall use 40 and say the cap is the default.
- REQ-3 (REQ-1): IF the configured cap is zero or not a whole number, THEN the preview shall refuse and name the setting.
- REQ-4: When the card of a linked worktree is done or its HEAD is merged into the integration ref, the preview shall list the worktree as a cleanup candidate with the reason, its allocated bytes and its location.
- REQ-5 (REQ-4): IF a worktree has a tracked change, an untracked file, an ignored file that is not regenerable, a git lock, a live card lease or a process working inside it, THEN the preview shall list it as kept with every reason.
- REQ-6: The preview shall include worktrees under `.claude/worktrees` and `.moai-cli/worktrees` and worktrees that sit inside another worktree's `target/` folder.
- REQ-10: The preview shall keep every worktree of the Go generation's folder and of Codex as kept.
- REQ-7: When picked worktrees are applied with each pick's item digest from the preview and a permission reference, `optimize-apply` shall remove each picked worktree with `git worktree remove` without force, the deepest path first, after a fresh check of that item.
- REQ-8 (REQ-7): IF a picked item's digest differs from the fresh one, the item is no longer a candidate, or its removal fails, THEN `optimize-apply` shall leave that worktree, report the reason and go on to the next pick.
- REQ-12: IF the permission reference is missing or unusable, THEN `optimize-apply` shall remove nothing.
- REQ-9: When `optimize-apply` finishes, it shall write one record with the path, branch, HEAD and restore command of each removed worktree and report the reclaimed space as the change of the file system's free bytes.
- REQ-11 (REQ-9): IF the record cannot be written, THEN `optimize-apply` shall report the failure together with the path of every worktree it removed.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소에 연결 worktree 셋을 만들고 상한을 2 로 주면 미리보기가 개수 4(주 체크아웃 포함), 상한 2, 출처 `flag`, 초과 `true` 를 말하고, 상한을 주지 않으면 40 과 출처 `default` 를 말해요 · verify: `cargo test --locked --offline --test worktree_cap_cleanup the_preview_counts_worktrees_against_the_cap`
- AC-2 (REQ-3): 상한 0, 음수, 숫자가 아닌 값은 미리보기가 거절하고 설정 이름을 말해요 · verify: `cargo test --locked --offline --test worktree_cap_cleanup a_cap_that_is_not_a_positive_whole_number_is_refused`
- AC-3 (REQ-4): 병합된 가지의 깨끗한 worktree 와 done 카드의 worktree 가 이유(`merged` 또는 `card_done`)와 할당 바이트와 위치와 함께 후보로 나오고, 병합되지 않았고 카드도 끝나지 않은 worktree 는 후보가 아니에요 · verify: `cargo test --locked --offline --test worktree_cap_cleanup finished_worktrees_are_candidates_with_reason_and_bytes`
- AC-4 (REQ-5): 추적 변경, 추적 안 된 파일, 되살릴 수 없는 무시 파일, git 잠금, 살아 있는 카드 임대, 안에서 일하는 프로세스가 있는 worktree 는 각각 이유를 달고 유지로 나오고, 적용이 그 항목을 거절해요 · verify: `cargo test --locked --offline --test worktree_cap_cleanup an_unsafe_worktree_is_kept_with_every_reason`
- AC-5 (REQ-6, REQ-10): `.claude/worktrees` 와 `.moai-cli/worktrees` 아래 worktree 가 후보로 나오고, 다른 worktree 의 `target/` 안에 중첩된 worktree 는 부모 이름(`nested_in`)과 함께 나오며, Go 세대 폴더와 Codex worktree 는 보존으로 남아요 · verify: `cargo test --locked --offline --test worktree_cap_cleanup old_locations_and_nested_worktrees_are_listed_and_preserved_ones_stay`
- AC-6 (REQ-7, REQ-8, REQ-12): 고른 항목만 지워지고(`--pick 경로@항목다이제스트`), 중첩된 worktree 는 부모보다 먼저 지워지고, 허가 참조가 없으면 아무것도 지우지 않고, 항목 다이제스트가 다르면 그 항목만 `digest_mismatch` 로 거절되고, 미리보기 뒤에 더러워진 항목도 거절되며 다음 항목은 계속 처리되고, 가지는 모두 남아요 · verify: `cargo test --locked --offline --test optimize_apply_cli apply_removes_only_the_picked_clean_worktrees_inside_out`
- AC-7 (REQ-9, REQ-11): 적용이 끝나면 지운 항목마다 경로·가지·HEAD·복구 명령이 든 기록 파일이 `.git` 아래에 생기고, 그 복구 명령으로 worktree 가 되살아나며, 회수한 공간은 `df` 여유 바이트의 변화로 나오고, 기록을 쓸 수 없으면 지운 경로를 모두 대며 실패를 알려요 · verify: `cargo test --locked --offline --test optimize_apply_cli apply_writes_a_record_that_restores_the_worktree`

## Scope

- write: 이 SPEC, `src/worktree_cap.rs`(새), `src/ops/optimize_plan.rs`(출력에 개수·상한·후보를 더함), `src/ops/optimize_apply.rs`(새), `tests/worktree_cap_cleanup.rs`(새), `tests/optimize_apply_cli.rs`(새), `tests/worktree_cleanup_safety.rs`(새, 감사 뒤 안전 시험), `docs/design-review/appendix/operations/optimize_apply.md`(새), `src/lib.rs`(모듈 한 줄)와 `src/help.rs`(도움말 행 — 새 명령은 도움말 표에 행이 있어야 `help_rules` 시험이 통과해요)
- 아직 쓸 수 없는 곳(다른 카드가 잡고 있어요, 범위 확장이 필요해요): 상한 설정 키를 읽는 `src/profile.rs`, `skills/moai/workflows/optimize.md` 의 안내 문단, `docs/design-review/appendix/operations/optimize_plan.md` 의 새 키 설명
- read-only: `src/optimize.rs`, `src/worktree_gc.rs`, `src/worktree_report.rs`, `src/ledger/**`, `src/lane_registry.rs`
- forbidden: `.claude/` 와 Go 세대의 `worktrees/projects/` 폴더 안에 쓰기, `git worktree remove --force`, `rm -rf`, 가지 삭제, 원장 쓰기, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME 과 실제 저장소의 worktree(시험은 임시 저장소만)

## Plan

1. 조각 1(이번): SPEC + PLAN 노트, 범위 확장 요청.
2. 조각 2(미리보기): `src/worktree_cap.rs` 의 순수 판정(개수·상한, 후보 판정, 중첩, 이유 모으기)과 `optimize-plan` 출력의 `worktree_cap` 과 `cleanup` 키. 상한은 `--worktree-cap` 과 HOME 설정 파일 `[worktrees] cap` 으로 받고, 프로필 키는 범위가 열린 뒤 이어요. 시험은 AC-1~5.
3. 조각 3(적용): `optimize-apply` 와 기록, AC-6·7.
4. 조각 4(마감): 변이 시험, 이웃 검사, 운영 문서, 프로필 키(범위가 열리면).

## Risks

- **삭제 기능이에요.** 항목마다 지우기 직전에 새로 확인하고(미리보기 이후에 바뀌었을 수 있어요), `--force` 를 쓰지 않아서 git 이 더러운 worktree 를 직접 거절해요. 지워지는 것은 worktree 폴더와 그 안의 되살릴 수 있는 무시 파일뿐이고 가지는 남아요. 복구 명령은 기록에 있어요(무시 파일 캐시는 되돌릴 수 없어요).
- **옛 위치 보존 규칙을 푸는 것은 사용자 결정이에요.** `src/worktree_gc.rs` 는 `.claude/worktrees` 아래를 보존으로 막아요(`preserved_location`). 이 SPEC 은 카드의 사용자 결정(옛 위치 포함)대로 그 한 이유만 새 층에서 풀고, 나머지 이유는 그대로 둬요. Codex 와 Go 세대 폴더는 풀지 않아요. 그 파일을 다른 카드가 잡고 있어서, 새 층은 그 계획의 행(`reasons`, `untracked`, `ignored_outside_build`, `change_age_secs`)을 다시 읽어 판정해요. 그 행의 모양이 바뀌면 이 층이 깨지니 시험이 그 모양을 고정해요.
- 카드 상태는 원장에서 읽어요. 원장이 없거나 읽을 수 없으면 카드 사실은 "관측 안 됨"이고 병합 경로만 써요. 원장이 있는데 읽을 수 없으면 후보를 만들지 않고 이유를 말해요(fail closed).
- 프로세스 검사는 작업 디렉터리만 봐요(`worktree_gc::process_cwds`). 파일만 열어 둔 프로세스는 못 봐요. 적용 직전에 한 번 더 읽지만 같은 한계예요.
- 공간 합계는 파일마다 `st_blocks` 를 더하므로 APFS 복제 파일은 여러 번 세어져 합이 실제 회수보다 커요. 그래서 회수 공간은 합이 아니라 `df` 여유 바이트의 변화로만 말해요.
- 후보가 수백 개면 걷기가 길어요. 미리보기는 시간 예산(`--budget-ms`)을 넘으면 나머지를 "측정 안 됨"으로 표시해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기존 `worktree-gc-plan` 행을 재사용하고 새 층에서 다시 판정해요(`src/worktree_gc.rs` 를 고치지 않아요) | 그 파일이 다른 카드에 잡혀 있어요. 새로 만들면 판정 규칙이 둘이 돼요 | `worktree_gc.rs` 에 옛 위치 허용 인자를 더하고 새 층을 걷어내기 |
| D-2 | 상한은 `--worktree-cap` 플래그, 그다음 HOME 의 `hook-config.toml` 의 `[worktrees] cap`(양의 정수), 없으면 40 상수 순서로 읽고 출처(`flag`·`config`·`default`)를 미리보기가 말해요. 프로필 키 `worktree_cap` 은 프로필 파일의 범위가 열린 뒤 이어요 | 카드 기준 2("설정에서 읽고 기본 40"); 프로필 파일은 다른 조각이 쥐고 있어서 리더가 HOME 설정 파일 키로 먼저 가라고 정했어요 | 키를 빼고 플래그만 두기 |
| D-3 | 옛 위치(`.claude/worktrees`, `.moai-cli/worktrees`)의 보존을 풀어요. Codex 와 Go 세대 폴더는 풀지 않아요 | 카드 본문의 사용자 결정(옛 위치 포함); 리더가 같은 기준으로 손으로 정리한 기록(2026-11-20 리더 정리 보고) | 새 층에서 옛 위치 행을 다시 보존으로 되돌리기 |
| D-4 | 적용의 "자기 토큰"은 미리보기가 후보마다 주는 `item_digest`(경로·가지·HEAD·이유·후보 종류에 묶여요) + 호출자가 준 허가 참조예요. 호출은 `--pick 경로@item_digest` 를 되풀이해요. 전체 계획 하나의 다이제스트로 묶지 않는 이유: 한 항목이 미리보기 뒤에 바뀌어도 나머지는 계속 처리해야 해서요 | 기존 `worktree-gc-apply` 는 계획 전체를 한 다이제스트로 묶어요; 이 명령은 사용자가 고른 항목마다 안전을 다시 보므로 항목 단위가 맞아요. 허가 참조는 허가 검증기가 아니고(`permission_reference_ok`) 승인은 호스트와 사용자의 몫이에요 | 없음 |
| D-5 | 되돌림 기록은 `<git 공통 폴더>/.moai-cli/worktree-cleanup/<now_unix>-<digest12>.json` 에 써요(추적되지 않는 `.git` 아래; 시각은 읽은 시계의 Unix 초) | 세션 레지스트리가 쓰는 `<공통 폴더>/.moai-cli/` 와 같은 자리; 보고서 폴더 이동 카드와 겹치지 않아요 | 보고서 폴더(`~/.moai/reports/<project>/`)로 옮기기 |
| D-6 | `git worktree remove` 는 `--force` 없이 써요. 되살릴 수 있는 무시 파일만 남은 worktree 는 `worktree_gc` 의 `ignored_safe` 판정을 따라요 | AGENTS.local.md 의 "never force" 와 `worktree_gc` 의 데이터 손실 방지 규칙 | 없음 |
| D-7 | 상한은 보여 주기만 해요(만들기를 막지 않고 자동 삭제도 없어요) | 카드 기준 2("미리보기가 초과를 말해요") | 후속 카드로 상한 강제 |
| D-8 | 새 층은 gc 행에서 버리는 이유(`preserved_location`(옛 위치만), 끝난 카드의 `not_merged`·`merged_unknown`, 센서스 이유)를 자기 규칙으로 다시 세워요: 끝난 카드여도 가지 없는(detached) worktree 는 어떤 ref 도 그 HEAD 를 가리킬 때만 후보이고(`unmerged_detached_commits`), 가지가 있는 worktree 는 지운 뒤에도 가지가 커밋을 지켜요; 하루가 안 된 빈 worktree 는 `empty_worktree` 로 유지해요 | 감사 B1·B3: 병합 안 된 detached 커밋과 막 만든 레인 worktree 를 지울 수 있었어요 | 규칙을 `worktree_gc.rs` 의 한 곳으로 옮기기(그 파일 카드가 풀린 뒤) |
| D-9 | worktree 목록은 `git worktree list --porcelain -z` 로 읽고, 한 경로에 행이 둘이거나 NUL 목록이 모르는 경로(줄바꿈이 든 경로가 gc 의 줄 단위 목록을 쪼갠 흔적)나 제어 문자가 든 경로는 판정하지 않고 유지해요. 적용도 한 경로에 정확히 한 행일 때만 지워요 | 감사 B2: 줄바꿈 경로가 이웃 worktree 의 가짜 행을 만들어 지우게 할 수 있었어요. 공용 파서(`worktree_report.rs`)는 이 카드 범위 밖이라 그대로 두고 후속으로 알려요 | 공용 파서를 `-z` 로 고치고 이 방어 걷기 |
| D-10 | 후보는 MoAI 의 위치로 제한하지 않아요: 이 저장소에 연결된 worktree 이면 어디에 있든, 병합됐거나 끝난 카드의 것이고 모든 안전 이유가 없을 때 후보예요. 지우는 것은 사용자가 고른 항목뿐이고(`--pick`), 가지는 남고, 복구 명령이 기록에 남아요 | 감사 N3: 사용자가 손으로 만든 worktree 도 같은 규칙(깨끗함·병합됨)이면 같은 대접이 일관돼요; 위치로 가르면 옛 위치 허용(D-3)과 갈라져요 | 후보를 `location` 이 project·claude 인 것으로 제한 |

## Result

AC-1~7 은 모두 충족이에요(근거는 아래 명령과 결과, 2026-11-20 가지 머리 `019243625a`).

| AC | 시험 | 결과 |
| --- | --- | --- |
| 1 | `--test worktree_cap_cleanup the_preview_counts_worktrees_against_the_cap` | 통과 |
| 2 | `--test worktree_cap_cleanup a_cap_that_is_not_a_positive_whole_number_is_refused` | 통과 |
| 3 | `--test worktree_cap_cleanup finished_worktrees_are_candidates_with_reason_and_bytes` | 통과 |
| 4 | `--test worktree_cap_cleanup an_unsafe_worktree_is_kept_with_every_reason` | 통과 |
| 5 | `--test worktree_cap_cleanup old_locations_and_nested_worktrees_are_listed_and_preserved_ones_stay` | 통과 |
| 6 | `--test optimize_apply_cli apply_removes_only_the_picked_clean_worktrees_inside_out` | 통과 |
| 7 | `--test optimize_apply_cli apply_writes_a_record_that_restores_the_worktree` | 통과 |

설정 파일 출처(`[worktrees] cap`)는 `worktree_cap::tests` 의 단위 시험이 덮어요(값 12 의 출처 `config`, 잘못된 값 거절).

READY 검사 종료 코드(`CARGO_TARGET_DIR` 는 이 worktree 전용): diff-check 0, fmt 0, ops rustfmt 0, clippy `--all-targets -D warnings` 0, `--lib` 전체 0(1609 통과), `--bins` 0, 영역 시험 대상(`worktree_*`, `optimize_*`) 0, 저장소 전체 검사용 시험(`git_env_scrub`, `home_isolation_scan`, `windows_unix_gate_scan`, `product_cli`) 0, `tools/repo_scans.json` 검사 39개 중 실패 0, 파이썬 `repo_scans_test`·`source_manifest_test`·`local_check_test` 0, `trailer-check` 0. macOS 에서 도는 `cfg(macos)` 시험은 101: 실패는 `update_command_cli::in_a_terminal_an_applied_update_asks_the_model_stage_and_saves_after_a_yes` 하나이고, 이 가지의 기준 커밋에서도 같이 실패해요(이 변경과 무관). 그 한 줄을 고치는 커밋 `728fc79a16` 이 로컬 `v2026.11.0` 에 이미 있어요.

뮤턴트(고의로 코드를 망가뜨려 시험이 잡는지 보기) 5개를 돌렸고 모두 의도한 시험 하나가 잡았어요: 안쪽부터 지우기 순서, 항목 다이제스트 비교, 지우기 직전 새 판정, 살아 있는 임대 유지, 보존 위치 범위.

알려진 한계: 프로세스 검사는 작업 디렉터리만 봐요. 할당 바이트 합은 APFS 복제(clone)를 겹쳐 셀 수 있어서 실제로 비워진 공간은 `df` 변화로만 알 수 있어요. 미리보기가 `gc` 계획을 두 번 돌려요(`optimize::plan` 과 후보 판정). 후보 미리보기는 카드 원장이 있어야 하고 없으면 `ledger_observed` 가 거짓이에요. 프로필 키 `worktree_cap` 은 프로필 파일의 범위가 열린 뒤 이어요.
