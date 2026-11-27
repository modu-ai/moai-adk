---
id: SPEC-20261126-check-affected-trailers-flaky
title: "check affected 가 트레일러 단계를 영수증에 싣고 플레이키 원장이 옛 상태 폴더를 만들지 않는다"
status: draft
revision: 2
created: 2026-11-26
---

# check affected 가 트레일러 단계를 영수증에 싣고 플레이키 원장이 옛 상태 폴더를 만들지 않는다

## Intent

두 가지 같은 뿌리의 문제를 한 카드에서 고쳐요. 둘 다 "잘못된 상태가 영수증이나 작업 트리에 남아서 감사나 제출에서야 드러나는" 꼴이에요.

1. 트레일러: 레인 커밋의 트레일러 문단이 나뉘거나 호스트 attribution 줄이 섞이면 `Moai-Session` 이 읽히지 않는데도 `moai trailer-check` 가 `failing 0` 으로 통과해요. 읽은 원인(2026-11-26, NeXT 끝 `398b66bda0`): git 은 마지막 문단에 트레일러가 아닌 줄이 있으면 그 문단 전체를 트레일러 블록으로 읽지 않아서 `git interpret-trailers --parse` 가 빈 결과를 내요(`%(trailers:key=…)` 를 읽는 `trailer-check`, `trailers --list`, `Moai-Spec` 읽기 전부 같아요). 반면 착지 검사의 `audit_gate::sessions_of` 는 마지막 문단에서 `Moai-Session:` 로 시작하는 줄을 그대로 읽어요. 그래서 검사기가 못 보는 세션을 착지 검사가 믿어요. 8168개 커밋 전체(현재 HEAD 에서 닿는 것)에서 이 방향의 불일치는 0건이라 새 실패를 만들 위험이 낮아요(awk 로 흉내 낸 수, Rust 파서 실행이 아니에요).
2. 플레이키 원장: `check affected` 가 플레이키 이름을 기록할 때 `state_root::path` 로 원장 위치를 풀어요. 이 저장소의 워크트리처럼 `.moai/` 에 마커가 없고 `.moai-cli/` 가 추적 중인 곳에서는 `.moai-cli/check-flaky.jsonl` 로 풀려서(읽은 근거: `kept_only` 가 `none` 으로 분류하고 `.moai-cli` 가 있으니 `dir()` 가 옛 폴더를 골라요) 추적도 무시도 안 되는 새 파일이 남고 `card submit` 의 깨끗한 트리 검사를 막아요. 같은 일이 하루에 두 번 있었어요.

3. 상태 폴더 판정(개정 2): 원장을 `.moai/local/` 에 두면, 마커가 없는 프로젝트에서 판정 규칙(`kept_only`)이 `local` 폴더를 몰라서 `classify` 가 `none` 에서 `unknown` 으로 바뀌고 업데이트 미리보기의 상태 이전 단계가 `blocked` 가 돼요. 2026-11-26 에 일회용 폴더 세 배치(옛 폴더 + 보존 이름만 있는 `.moai/`, 옛 폴더만, 상태 폴더 없음)에 원장 파일을 손으로 만들어 설치된 `moai project-update-plan` 으로 확인했어요. 셋째 배치는 원장을 `.moai/` 바로 아래에 쓰던 이전 코드에서도 같았어요.

리더 결정 노트의 방향: 원장은 `.moai/` 아래에 쌓고 이미 있는 옛 원장은 읽기만 하고, `state_root::path` 자체는 공유 인터페이스라 바꾸지 않아요. 판정 문제는 `kept_only` 가 `local` 폴더를 알게 해서 고쳐요.

## Non-goals

- `state_root::path`, `dir`, `create_dir`, `resolves_to_moai` 의 규칙. `classify` 의 판정표는 `local` 폴더를 아는 것 말고는 그대로예요.
- 프로젝트의 `.gitignore` 가 `.moai/local/` 을 가리게 만드는 일(이 저장소는 추적 파일의 `/.moai/*` 가 이미 가려요).
- `Co-Authored-By` 줄이 없는 커밋을 실패로 만드는 일, 트레일러 문단 쪼개짐(`split_trailer_block`)을 범위 모드에서 실패로 올리는 일(이미 착지한 커밋이 새로 실패하지 않게 한 다른 SPEC 의 결정이에요).
- 런처가 호스트 attribution 줄을 따로 두는 일과 흐름 글의 트레일러 규칙 문장을 고치는 일. 이 SPEC 은 검사기 쪽만 맞추고 나머지는 후속 줄로 남겨요.
- `audit_gate::sessions_of` 와 `card land` 의 `authors_of` 를 git 의 읽기로 바꾸는 일(GitHub API 메시지를 읽는 PR 감사 관문에는 git 이 없어요).
- 병합 커밋 판정(범위 모드의 `trailer-check` 는 병합을 보지 않아요).

## Requirements

- REQ-1: When `check affected` runs on a range, the run shall apply the trailer rules of `moai trailer-check` to that range on the checkout's own branch as a cheap in-process step labelled `trailers` · failure: none — pinned by AC-1.
- REQ-2 (REQ-1): IF the trailer rules report a failing finding for the range, THEN the summary line shall carry `trailers=1` and `result=FAIL` with the heavy steps skipped as for any failed cheap step.
- REQ-3: When the trailer rules report no failing finding for the range, the summary line shall stay exactly as it was before this change, without a `trailers` word · failure: none — pinned by AC-2.
- REQ-4: When the last paragraph of a commit message holds a line that begins with `Moai-Spec:`, `Moai-Session:` or `Moai-Verified:` while git reads no value of that key from it, `moai trailer-check` shall report a failing `unread_trailer_line` finding that names the key and the line · failure: none — pinned by AC-3.
- REQ-5: When `check affected` records a flaky name, it shall append the row to the `local` folder of the project's `.moai` folder · failure: none — pinned by AC-4.
- REQ-6: When `check affected` records a flaky name, the earlier state folder shall stay exactly as it was, with no folder or file created or written in it · failure: none — pinned by AC-4.
- REQ-7: When `moai check flaky` reports, it shall count the rows of the new ledger and the rows of an existing ledger in the earlier state folder together, leaving the earlier file byte for byte · failure: none — pinned by AC-5.
- REQ-8 (REQ-5): IF the new ledger cannot be written, THEN the check result shall stay as it would have been without the ledger.
- REQ-9: When a project's `.moai` folder holds a real `local` folder and every other entry is one the kept-only rule already allows, `classify` shall answer `none` as for an absent `.moai` folder.
- REQ-10 (REQ-9): IF the `local` entry is a link or no folder, or a Go marker or an unfinished go-state receipt is present, THEN `classify` shall answer what it answered before this change.
- REQ-11: When the state move applies into a `.moai` folder that already holds a `local` folder, the apply shall move the earlier state beside that folder and record it in the receipt as held before · failure: none — pinned by AC-7.
- REQ-12: When the revert of such an apply runs, it shall leave the `.moai` folder in place with the files of its `local` folder unchanged · failure: none — pinned by AC-7.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소의 `spec/` 가지에서 트레일러 문단이 나뉜 커밋(`Moai-Spec` 줄만 앞 문단에 있어요)과 attribution 줄이 문단에 섞인 커밋(항목 I)을 범위에 두고 `check affected` 를 돌리면 요약이 `trailers=1`, `stopped_at=trailers`, `result=FAIL` 이고 무거운 단계는 건너뛰어요 · verify: `cargo nextest run --locked --offline --test check_affected_trailers a_split_trailer_paragraph_fails_the_receipt_with_trailers_1 an_attribution_line_in_the_trailer_paragraph_fails_the_receipt_with_trailers_1`
- AC-2 (REQ-3): 트레일러가 정상인 범위의 요약은 `trailers` 낱말이 없고 단계와 낱말 순서가 이전과 같아요 · verify: `cargo nextest run --locked --offline --test check_affected_trailers a_clean_range_keeps_the_summary_without_a_trailers_word`
- AC-3 (REQ-4): 마지막 문단에 attribution 줄이 섞인 메시지는 `unread_trailer_line` 으로 실패하고, 정상 블록과 `Signed-off-by` 가 섞인 블록, 병합 커밋의 충돌 주석 문단은 실패하지 않아요 · verify: `cargo nextest run --locked --offline --test card trailer_check_cli::an_attribution_line_in_the_last_paragraph_is_an_unread_trailer_line trailer_check_cli::a_clean_block_and_a_signed_off_by_block_raise_no_unread_finding`
- AC-4 (REQ-5, REQ-6, REQ-8): 옛 폴더가 없는 임시 프로젝트와 옛 폴더가 있고 `.moai` 에 마커가 없는 임시 프로젝트에서 플레이키를 기록하면 `.moai` 아래 `local/check-flaky.jsonl` 에 쌓이고 옛 폴더의 파일은 바이트가 그대로이며 새 파일이 없어요. 기록 뒤에도 `state_root::dir` 의 답은 같아요. 원장을 못 쓰면 결과는 그대로예요 · verify: `cargo nextest run --locked --offline --test check_flaky_ledger the_flaky_ledger_is_written_under_the_local_folder_and_the_earlier_folder_stays_untouched the_state_folder_answer_is_the_same_after_a_flaky_record`
- AC-5 (REQ-7): 옛 원장의 이름 횟수와 새 원장의 횟수가 합쳐져서 `check flaky` 가 보고하고 옛 원장은 바이트가 그대로예요 · verify: `cargo nextest run --locked --offline --test check_flaky_ledger the_flaky_report_counts_both_ledgers_and_reads_the_earlier_one_only`
- AC-6 (REQ-9, REQ-10): 일회용 프로젝트 세 배치(옛 폴더 + 보존 이름만 있는 `.moai`, 옛 폴더만, 상태 폴더 없음)에서 플레이키를 기록하면 `classify` 가 `none` 이고 `project-update-plan` 의 상태 이전 단계가 기록 전과 같아요. 상태 폴더가 없던 배치는 기록 뒤에도 `adopt_new` 가 마커를 써요. `local` 이 링크이거나 파일이면, 또는 Go 마커가 있으면 판정은 이전과 같아요 · verify: `cargo nextest run --locked --offline --test check_flaky_verdict a_flaky_record_leaves_the_state_verdict_and_the_update_preview_as_they_were a_linked_or_plain_file_local_entry_keeps_the_earlier_verdict`
- AC-7 (REQ-11, REQ-12): 옛 폴더가 있고 `.moai/local/` 에 파일이 있는 일회용 프로젝트에서 상태 이전을 적용하고 되돌리면, 적용은 옛 상태를 그 옆으로 옮기고, 되돌린 뒤 `.moai/local/` 의 파일은 바이트가 그대로이며 `.moai` 는 치워지지 않고 옛 상태는 제자리로 돌아와요 · verify: `cargo nextest run --locked --offline --test check_flaky_verdict a_state_move_and_its_revert_leave_a_local_folder_in_place`

## Scope

- write: `src/check_affected.rs`
- write: `src/check_flaky.rs`
- write: `src/trailers.rs`
- write: `src/ops/check.rs`
- write: `src/state_root.rs`
- write: `src/go_state.rs`
- write: `tests/check_flaky_verdict.rs`
- write: `tests/check_affected_trailers.rs`
- write: `tests/check_affected_rerun.rs`
- write: `tests/check_flaky_ledger.rs`
- write: `tests/card/trailer_check_cli.rs`
- write: `.moai/specs/SPEC-20261126-check-affected-trailers-flaky/**`
- write (shared file, only by `moai card line-lease`): `tools/repo_scans.json`, `tests/areas.toml`
- read-only: `src/audit_gate.rs`, `src/card/submit.rs`, `src/card/land/check.rs`, `src/project_update.rs`
- forbidden: `src/gate.rs`, `Cargo.toml`, `skills/**`

## Plan

1. SPEC 초안을 커밋해요. 개정 1의 변경은 감사 바닥 목록(tier 2) 밖이었고, 개정 2는 `src/state_root.rs` 와 `src/go_state.rs` 를 고쳐 tier 2 라서 승인을 받은 뒤에 그 두 파일을 고쳐요.
2. RED 한 커밋에 모든 시험을 담아요: 트레일러 단계 실패 둘과 통과 하나(`tests/check_affected_trailers.rs`), 새 검사 항목 둘(`tests/card/trailer_check_cli.rs`), 원장 시험 셋(`tests/check_flaky_ledger.rs`), 그리고 옛 위치를 말하는 기존 시험 `ac5_an_unwritable_ledger_does_not_change_the_result` 를 새 위치로 고쳐요.
3. GREEN: `check_affected.rs` 에 `table_step` 모양의 `trailers` 단계(실패가 있을 때만 계획해서 통과 영수증의 모양이 안 바뀌어요)를 `run_with` 와 `list` 양쪽에, `trailers.rs` 에 `unread_trailer_line` 항목을, `check_flaky.rs` 의 쓰기 경로를 `.moai/local/` 로, 읽기는 새 원장과 옛 원장 둘을 `EARLIER_PROJECT_DIR` 로 읽어요.
4. 시험 파일을 `tools/repo_scans.json`(규칙과 검사)과 `tests/areas.toml`(첫 낱말이 이미 있으면 생략)에 line-lease 로 등록해요.
5. 개정 2 의 승인 뒤에 `tests/check_flaky_verdict.rs` 의 시험을 먼저 쓰고(RED), `kept_only` 가 진짜 `local` 폴더를 허용하게(보존 이름이 없어도), 상태 이전의 영수증이 그 폴더를 "전부터 있던 이름"으로 기록하게 고쳐요(GREEN).
6. 영수증(`check affected --card`), 위험 경로 실행(일회용 폴더에서 상태 이전과 되돌리기, 링크 `local`)과 guard-mutant 줄을 제출에 담아요.

## Risks

- 새 `trailers` 단계가 이미 착지한 긴 범위를 실패로 만들 수 있어요. 측정(2026-11-26, `moai trailer-check`, 현재 규칙): `e946ae8102..398b66bda0`(병합 14개를 뺀 64개 커밋)은 7건 실패(`--branch NeXT` 로는 5건: `missing_verified` 4, 팁 커밋의 `malformed_spec` 1), 그 직전 30개 구간은 20건(17건)이에요. 그래서 이 단계는 레인 자신의 범위(`card submit` 의 5번째 관문이 같은 규칙을 이미 돌려요)와 묶음 착지의 새 범위에 쓰이고, 착지된 긴 범위에 쓰면 실패가 나올 수 있어요. 단계가 지금의 `trailer-check` 와 같은 규칙만 쓰므로 새 규칙으로 새 실패가 생기는 것은 REQ-4 의 항목 하나뿐이고 그 항목은 8168개 커밋에서 0건이에요.
- `kept_only` 의 변경은 `classify`, `adopt_new`, 상태 이전의 적용과 되돌리기에 닿아요. 사용자가 직접 만든 `.moai/local/` 만 있는 프로젝트도 이제 `none` 으로 읽혀서, `adopt_new` 가 그 폴더에 마커를 쓰고 상태 이전이 옛 상태를 그 옆으로 옮길 수 있어요. 두 길 모두 그 폴더 안의 파일을 지우거나 옮기지 않아요(AC-7 이 고정해요).
- `local` 을 "전부터 있던 이름"으로 기록한 영수증은 이전 빌드의 되돌리기가 거절해요(`receipt names kept folder entries it cannot have`). 그 빌드에서 되돌리려면 새 빌드를 써야 해요.
- `src/go_state.rs` 는 다른 활성 예약과 범위가 겹쳐요. 읽은 그 가지의 변경은 함수 공개 범위를 넓히는 줄이고 이 SPEC 이 고치는 자리와 겹치지 않아요.
- 프로젝트의 `.gitignore` 가 `.moai/local/` 을 가리지 않으면(옛 배치의 사용자 프로젝트) 원장은 여전히 추적 안 된 파일로 보여요. 이전에 `.moai-cli/` 에 쓸 때와 같고 이 SPEC 은 고치지 않아요.
- 새 `trailers` 낱말이 요약 줄을 읽는 코드를 깨지 않는지는 읽어서 확인했어요(`card submit` 의 `pass_summary`, `ready_receipt::judge`, `land-batch` 의 `receipt_covers`, `check_flaky::read_earlier` 는 `check-affected`, `head=`, `base=`, `selection=`, 마지막 `result=` 만 봐요). 그래도 실패 때만 단계를 계획하는 설계라 통과 영수증은 바뀌지 않아요.
- 병합 커밋과 묶음 착지의 분리된 작업 폴더(`HEAD` 가 떨어져 있어요)에서는 가지 이름이 `spec/` 이 아니라서 `Moai-Spec` 누락이 실패하지 않아요. 이 단계는 레인 가지에서는 `card submit` 과 같은 강도이고 묶음 착지에서는 약해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `trailers` 단계는 실패가 있을 때만 계획해요(`table_step`, `endline_step_if_any` 모양). 통과 영수증의 요약 줄과 단계 목록은 그대로예요. | 요약 줄과 단계 목록을 고정한 시험이 열 곳 넘게 있고, 읽는 코드는 새 낱말에 안 깨져요(조사 결과) | 항상 계획하는 쪽으로 바꾸면 그 시험들을 같이 고쳐요 |
| D-2 | 새 항목 `unread_trailer_line` 은 `trailer-check` 의 범위 모드에서 실패해요(`fails: true`). 메시지 파일 모드(커밋 훅)는 이 SPEC 에서 바꾸지 않아요. | 조사: git 의 읽기와 착지 검사의 읽기가 갈라지는 방향의 실제 사례가 8168개 중 0건 | `fails: false` 로 내리면 항목은 알림만 해요 |
| D-3 | 새 원장 위치는 프로젝트의 `.moai` 폴더 아래 `local/check-flaky.jsonl` 이에요. 위치는 `state_root` 의 상수(`MOAI_DIR`, `LOCAL_DIR`)로 만들고, 옛 원장은 `EARLIER_PROJECT_DIR` 로 읽기만 해요. | 리더 결정(원장은 `.moai/` 아래, 옛 원장은 읽기만), 추적 파일 `.gitignore` 의 `/.moai/*` 가 이 경로를 무시해요, `state_root::path` 는 바꾸지 않아요 | `state_root` 의 규칙이 바뀌면 그 함수로 돌아가요 |
| D-4 | `check flaky` 보고는 새 원장과 옛 원장을 합쳐서 세요. | 전환 직후에 이전 횟수를 잃지 않게 | 옛 원장을 버리려면 읽기를 지워요 |
| D-5 | 시험은 폴더 이름을 문자열로 적지 않고 상수와 `ledger_path` 를 써요. | 상태 폴더 문자열 스캔(`allow.tsv`)과 `path_root_direct_sites` 기준선 | — |
| D-6 | `kept_only` 는 진짜 `local` 폴더를 `backups` 처럼 허용하고, 보존 이름이 하나도 없어도 `local` 이 있으면 참이에요. | 리더 결정(원장은 `.moai/` 아래, 판정 규칙이 `local` 폴더를 알게 해요). `local/` 은 기계 로컬 출력의 정해진 자리예요(`LOCAL_DIR`) | 허용 줄을 지우면 판정은 이전과 같아요 |
| D-7 | 상태 이전의 영수증은 전부터 있던 `local` 폴더를 기록해서, 되돌리기가 `.moai` 를 치우지 않아요. | 기록하지 않으면 되돌리기가 `.moai` 를 통째로 옆으로 치워요(읽은 근거: 보존 이름 기록이 비면 `set_aside`) | 두 자리를 되돌려요 |
