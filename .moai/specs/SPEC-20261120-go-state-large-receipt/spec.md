---
id: SPEC-20261120-go-state-large-receipt
title: "go-state-revert finishes with a receipt above 256 KiB, and a refused receipt save is never reported as a stale plan"
status: draft
revision: 1
created: 2026-11-20
---

# go-state-revert finishes with a receipt above 256 KiB, and a refused receipt save is never reported as a stale plan

## Intent

한 프로젝트에서 `moai go-state-revert` 가 미리보기에는 충돌이 없었는데도 `stale_plan`("target changed during apply; temporary file discarded")으로 멈췄어요. 같은 시도를 두 번 하는 동안 복원 사본이 하나씩 지워졌고, 상태는 `interrupted`, 영수증은 `pending` 그대로라서 그 뒤의 `update` 가 막혔어요(원본은 `.moai-backup/` 과 git HEAD 에 있어 안전했어요).

원인은 둘이에요(팁 `1c9f1a1f95c58db9ee7e843f0dc0557a84b120be` 의 코드를 읽은 결과).

1. 크기 한도가 쓰기와 읽기에서 달라요. 영수증은 처음 쓸 때 크기 검사 없이 새로 만들고(`publish_exclusive_mode`), 읽을 때는 16 MiB 한도(`MANIFEST_LIMIT`)로 읽어요. 그런데 그 뒤의 모든 저장은 `publish_replace` → `stage` 를 거치고, `stage` 가 기존 영수증을 다시 읽을 때 `read_opt` 의 256 KiB(`RECEIPT_LIMIT`) 한도에 걸려요. 복원 항목이 1,319개인 실제 영수증은 262,362바이트라서 한도를 넘었어요. 이 크기 실패를 `stage` 가 "target changed during apply" 로 돌려주고, `contract.rs` 는 그 문구를 `stale_plan` 으로 분류해요. 그래서 크기 문제가 "계획이 낡았다"로 보였어요.
2. 순서가 거꾸로예요. `revert` 는 복원 사본을 먼저 지우고 그 다음에 영수증을 저장해요. 저장이 실패하면 지우기는 이미 끝난 뒤인데도 되돌리기는 멈춰요. 도움말은 "All or nothing" 이라고 해요.

같은 저장 경로를 `go-state-apply` 도 써요. 복원 사본마다 영수증을 다시 저장하므로, 영수증이 256 KiB 를 넘는 순간 apply 도 같은 가짜 `stale_plan` 으로 멈춰요(RED 시험으로 확인: 긴 폴더 이름의 복원 항목 1,000개짜리 프로젝트에서 apply 가 `target changed during apply; temporary file discarded` 로 멈췄어요). 그래서 수리는 apply 와 revert 가 함께 쓰는 저장 한 곳에 들어가요.

이 SPEC 은 이 둘을 고쳐요. 영수증 하나에 한도 하나(쓰는 쪽이 지키고 읽는 쪽이 같은 값을 써요)를 두고, 크기 실패는 자기 이름의 오류로 말하고, 되돌리기는 첫 삭제 전에 영수증을 저장할 수 있는지 먼저 확인해요. 이미 지워진 복원 사본을 이미 지워진 것으로 읽는 복구 길(`revert_plan` 의 `skip`)은 코드에 이미 있으니 시험으로 고정해요.

## Non-goals

- 모든 파일의 256 KiB 한도(`read_opt`)를 올리는 일. 그 한도는 AGENTS.md 와 설정 파일 등 다른 대상의 규칙이고, 시험이 고정해요(`codex_config`, `agents_local_merge`). 영수증만 자기 한도를 써요.
- 영수증을 쪼개거나 압축하는 일(복원 목록을 해시 파일로 분리하는 설계). 16 MiB 한도는 실제 프로젝트 영수증의 60배 넘는 여유예요; 쪼개기는 한도가 부족하다고 밝혀질 때의 다음 일이에요.
- 되돌리기에서 삭제 순서를 영수증 먼저로 바꾸는 일. 영수증을 먼저 저장하면 중단 시 지워지지 않은 사본이 기록에서 빠져 영영 남아요; 첫 삭제 전 저장 확인과, 이미 지워진 사본을 `skip` 으로 읽는 기존 길이면 어느 순서의 중단에서도 다시 실행해 이어갈 수 있어요.
- `update` 와 다른 `publish_replace` 사용처의 동작 변경. 검토만 하고 결과를 `## Result` 에 적어요(Plan 4).

## Requirements

- REQ-1: The go-state receipt writer shall refuse a receipt larger than the shared receipt limit of 16 MiB before it writes any file, for go-state-apply and go-state-revert alike, and name the limit in its message.
- REQ-2: When go-state-apply or go-state-revert saves its receipt after a step, the save shall replace an existing receipt of any size up to the shared receipt limit.
- REQ-3 (REQ-2): IF that save fails, THEN go-state-apply and go-state-revert shall report the real cause of the failure and never the stale-plan kind.
- REQ-4: The apply engine shall report a replace that is refused because the existing file is larger than the limit of that replace with a message that names the limit and has the kind invalid_input, never stale_plan.
- REQ-5: Before its first write go-state-revert shall prove that it can save the receipt by saving the unchanged receipt once.
- REQ-6: When go-state-revert runs again on a receipt that lists a restored copy which is already gone, the revert shall treat that copy as removed and carry on to the end. · failure: none — a missing copy needs no action, so nothing can fail; a copy that is still there is judged by its digest as before.
- REQ-7: The help text of go-state-revert shall say that a conflict or a changed backup refuses the revert before anything is written, and that a stop part-way is recorded and continued by running the revert again.

## Acceptance

- AC-1 (REQ-2, REQ-5): a project whose go-state receipt grows above 256 KiB during the apply (1,000 restored entries with long folder names) applies completely and then reverts completely: the revert plan is current, every restored copy is removed, the Go folder is back and the receipt state is reverted.
  verify: `cargo test --locked --offline --test go_state_cli a_receipt_above_256_kib_reverts_completely`
- AC-2 (REQ-6): a restored copy that an earlier stopped revert already removed is not a conflict: the revert plan reads it as skip and the run finishes.
  verify: `cargo test --locked --offline --test go_state_cli a_restored_copy_removed_by_a_stopped_revert_is_not_a_conflict`
- AC-3 (REQ-3, REQ-5): when the receipt cannot be saved (its folder is not writable), the revert stops before it removes anything, every file is still in place, and the error is not the stale-plan kind.
  verify: `cargo test --locked --offline --test go_state_cli a_refused_receipt_save_leaves_every_file_and_is_not_a_stale_plan`
- AC-4 (REQ-4): a replace of a file larger than the limit of that replace is refused with a message that names the limit and does not contain the stale-plan phrases; the temporary file is gone.
  verify: `cargo test --locked --offline --lib apply::tests::a_replace_over_the_size_limit_names_the_limit`
- AC-5 (REQ-4): the contract's table of current messages lists both new size messages as invalid_input and keeps the target-changed message as stale_plan.
  verify: `cargo test --locked --offline --lib contract::tests::every_current_message_has_a_deliberate_kind`
- AC-6 (REQ-1): a receipt above the shared receipt limit is refused before any file is written, and the message names the limit.
  verify: `cargo test --locked --offline --lib go_state::tests::a_receipt_over_the_limit_is_refused_before_any_write`
- AC-7 (REQ-7): `moai go-state-revert --help` says what REQ-7 says.
  verify: `cargo test --locked --offline --test go_state_cli the_revert_help_says_what_a_stop_part_way_leaves`
- AC-8 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261120-go-state-large-receipt`

## Scope

- write: `.moai/specs/SPEC-20261120-go-state-large-receipt/**`, `src/apply.rs` (a replace with its own limit and a size error of its own), `src/contract.rs` (the table of current messages names the new ones), `src/go_state.rs` (the receipt writer, the save check in `revert`, unit tests), `src/main.rs` (the help sentence only), `tests/go_state_cli.rs`.
- read-only: every other file, the real HOME.
- tier 2 (deletions, project state): an independent audit of the exact head.
- order of the files: SPEC, red tests, `src/apply.rs` and `src/contract.rs`, then `src/go_state.rs` and `src/main.rs` when their holders release them.

## Plan

1. SPEC and the red tests: a receipt above 256 KiB that must apply and revert completely (fixture with 1,000 restored entries with long folder names through the real go-state-apply), a stopped revert with already removed copies, a receipt folder that cannot be written, the apply-engine size error, the contract classes, the receipt over the limit, the help sentence.
2. `src/apply.rs`: `read_opt_limited` and `publish_replace_limited(path, before, after, limit)`; `publish_replace` keeps its 256 KiB limit through them; `stage` returns the size error ("target larger than N bytes; refused") instead of "target changed during apply" when the re-read fails on size; `read_opt` keeps its message. `src/contract.rs`: the table of current messages lists the new ones as invalid_input (the existing `refused` rule already classes them; the table pins that on purpose).
3. `src/go_state.rs`: one `RECEIPT_MAX` (the value of `MANIFEST_LIMIT`) for the writer and the reader; `save_receipt` refuses above it and replaces through the limit-aware publish; `revert` saves the unchanged receipt once before its first write. `src/main.rs`: the help sentence.
4. Review of the other `publish_replace` users that write files larger than 256 KiB (`update.rs`, `agents_local_merge.rs`, `archive.rs`, the home layers), recorded in `## Result`: the size message is the only change they get.
5. The unnarrowed READY set, then an independent audit.

## Risks

- `src/go_state.rs` and `src/main.rs` are reserved by other cards when this SPEC is written; the plan keeps those edits small, in regions the other branches do not touch (the other branch changes the file near lines 1803 and 2689; this change near 3358 and 4710).
- The save check writes the receipt once more with the same bytes: a new file with the same content, so the receipt's mtime moves. Nothing reads that mtime today (read in the code: `read_receipt` reads the content only).
- Some test may pin the old "target changed during apply" text for a size case; the lib and integration targets of `apply`, `contract` and `go_state` run in full at READY to find it.
- A 16 MiB receipt still reads whole into memory; the limit is the reader's own and unchanged.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the go-state receipt uses one limit of 16 MiB (the reader's existing value), enforced by the writer; all other apply targets keep 256 KiB | the card text asks for one limit that the writer enforces and every reader shares; changing the global limit would change AGENTS.md and settings rules that tests pin | lower the constant; a receipt above it is then refused at the writer |
| D-2 | `revert` checks that the receipt can be saved before its first write, instead of saving the receipt before each removal | the card text allows either; saving first leaves an unrecorded copy after a stop, the check plus the existing `skip` reading leaves nothing behind | reorder the two lines in `revert` |

Both are reversible and in none of the five protected categories (money, legal, external contract, secret, real data); the lane chose them inside the options the dispatch gave and records them here.

## Result

- 구현(2026-11-20): `src/apply.rs` — `read_within(path, limit, named_size)` 와 `publish_replace_limited(path, before, after, limit)`; `publish_replace` 는 그 위에서 256 KiB 를 그대로 지키고, `stage` 가 다시 읽기에서 크기로 막히면 `target larger than N bytes; refused` 를 돌려줘요("changed during apply" 가 아니에요). `read_opt` 의 문구는 그대로예요(`codex_config` 시험이 고정). `src/go_state.rs` — 영수증 한도 `RECEIPT_MAX`(16 MiB, 읽기 쪽 값)를 `save_receipt` 와 `read_receipt` 가 같이 써요; `save_receipt` 는 한도를 넘는 영수증을 아무것도 만들기 전에 `receipt larger than 16 MiB; refused; nothing was written` 으로 거절하고, 교체는 한도 인지 저장으로 해요; `revert` 는 첫 쓰기 전에 바뀌지 않은 영수증을 한 번 저장해 보여요. `src/contract.rs` — 현재 메시지 표에 두 새 메시지를 `invalid_input` 으로 올렸어요(기존 `refused` 규칙이 이미 분류해서 새 규칙은 필요 없었어요).
- 시험: RED 먼저(통합 4개 중 3개와 단위 1개 실패, 나머지 1개는 코드에 이미 있던 `skip` 길을 고정하는 시험이라 통과): 크기가 "target changed during apply" 로 나옴, 저장이 실패했는데 파일이 이미 바뀜, 도움말 문장 없음, 그리고 복원 항목이 많은 프로젝트에서 `go-state-apply` 자체가 같은 가짜 `stale_plan` 으로 멈춤. GREEN 뒤: lib `apply::tests`·`contract::tests`·`go_state::tests::a_receipt_over` 22 passed(종료 0), `tests/go_state_cli.rs` 65개 중 64 passed — 실패 1개는 AC-7(도움말)이에요. AC-1 시험은 긴 폴더 이름의 복원 항목 1,000개로 영수증이 262,144바이트를 넘게 만들고(시험이 크기를 직접 단언) apply 와 revert 를 끝까지 돌려요; 단독 75.8초(디버그 빌드, 저장마다 디스크 동기화가 항목 수에 비례).
- 변이 5개(커밋 뒤 실행): 크기 오류 분기 제거, 영수증 저장을 기본 한도로 되돌림, 첫 쓰기 전 저장 확인 제거, 쓰기 쪽 경계 `>` 를 `>=`, 읽기 쪽 경계 `>` 를 `>=` — 모두 적용되고 모두 시험 실패(종료 101, 컴파일 오류 아님: 각 로그에서 실패한 시험 이름을 읽음)로 잡혔어요; 실행 뒤 트리는 깨끗했어요.
- 검사(이 head, 전용 target 폴더, heavy-slot): `git diff --check` 0, `cargo fmt --check` 0, `rustfmt --edition 2021 --check src/ops/*.rs` 0, `cargo clippy --offline --locked --all-targets -- -D warnings` 0, `--lib` 전체 1623 passed, `--bins` 0, 영역 통합 시험 93개 대상(go_state·agents·update·home·project·github·omp·recovery·archive·generation 등) 중 `go_state_cli` 만 101(위의 도움말 한 개), 그 시험만 건너뛴 `go_state_cli` 는 64 passed(종료 0), `tools/repo_scans.json` 스캔 39개 전부 0, `python3 -m unittest tests.local_check_test` 0, `moai spec check` 0, `trailer-check` 0.
- (d) 검토: 다른 `publish_replace` 사용처 42곳을 읽었어요. 17곳은 `apply::read_opt`, 설정 파일류(`go_hooks`·`go_settings_clean`·`update_hooks`)는 `claude::read_existing` 으로 같은 262,144 한도에서 먼저 거절해서("not a bounded regular file") 교체까지 가지 않아요. 큰 한도로 읽어 `before` 로 넘기며 가짜 stale 을 냈던 곳은 `go_state` 영수증뿐이었어요. `agents_local_merge` 는 자기 크기 검사가 있고, `spec_close` 는 크기 검사가 없지만(256 KiB 넘는 SPEC 은 없다고 보고 손대지 않음) 이제 크기가 걸리면 새 메시지로 나와요.
- 남은 것: AC-7(도움말 문장)은 `src/main.rs` 가 다른 레인의 예약에 있어서(예약 회수가 그 폴더의 커밋 안 된 캐시 파일 때문에 거절됨) 그 레인이 문장을 넣기로 했어요(리더 결정); 시험 `the_revert_help_says_what_a_stop_part_way_leaves` 는 이 head 에서 빨강이고 그 문장이 들어오는 head 와 함께 통합돼야 초록이에요. 돌리지 않은 것: 문제를 신고한 실제 프로젝트(영수증 262,362바이트)의 직접 재현, 전체 `cargo test`(영역 대상 93개와 lib·bins 만 돌림), CI, Windows. 이미 복원 사본이 지워진 프로젝트의 복구 길은 `revert_plan` 이 지워진 사본을 `skip` 으로 읽는 기존 코드이고, AC-2 시험이 그 길을 고정해요(새 코드는 아님).
