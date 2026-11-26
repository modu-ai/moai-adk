---
id: SPEC-20261125-update-run-receipt
title: "One confirmed moai update leaves a run receipt that names what is left and reads the written skill files back"
status: draft
revision: 1
created: 2026-11-25
parent: SPEC-20261125-update-screen-facts
---

# One confirmed moai update leaves a run receipt that names what is left and reads the written skill files back

## Intent

`moai update` 는 스킬 복사본을 먼저 쓰고, 그 뒤의 부분(agents, routing, style, 링크, hooks, 키, 정리)이 따로 실패하거나 거절돼도 스킬은 이미 적용된 채로 끝나요(`src/update.rs` 의 `update_command_core`). 한 번의 실행을 통틀어 말해 주는 영수증이 없어서, 반쯤 적용된 상태를 끝 화면이 말하지 못하고 사용자는 무엇이 남았는지, 다시 실행하면 되는지 알 수 없어요. 또 쓴 스킬 파일이 정말 의도한 바이트인지 디스크에서 다시 읽어 확인하지도 않아요.

이 SPEC 은 확인된 실행 하나가 끝날 때 결과에 `run` 기록 하나를 담고(남은 부분의 이름, 일부러 남긴 부분, 쓴 스킬 파일의 해시 대조), 무언가를 쓴 실행이면 그 기록을 HOME 상태 폴더의 파일 하나로 남기고, 끝 화면을 그 기록에서 그리게 해요. `update` 재설계의 S7 조각이에요(순서와 근거는 `SPEC-20261125-update-screen-facts` 가 속한 update 전체 분석).

## Non-goals

- 스킬 밖의 부분(agents, routing, style, bin_names, hooks, keys)이 쓴 파일의 해시 대조: 그 부분들은 결과에 파일 경로와 기대 해시가 아니라 `host:file` 꼬리표를 주고 각자 영수증을 가져서, 이 기록은 `not_compared` 로 이름과 영수증 id 만 남겨요. 후속 조각의 몫이에요.
- 실패한 부분을 자동으로 다시 시도하거나 되돌리는 일: 남은 부분은 사용자가 원인을 고친 뒤 `moai update` 를 다시 실행해서 이어요.
- 적용 순서, 확인 절차(plan digest), 부분별 결과의 모양을 바꾸는 일: `mode`, `applied` 와 부분별 `*_result` 는 그대로예요.
- `skills/moai/workflows/update.md` 가 새 `run` 키를 읽도록 쓰는 일: 지침 문장은 별도 조각이에요.
- 설치된 사용자 상태를 읽거나 바꾸는 시험: 시험은 임시 HOME 에서만 돌아요.

## Requirements

- REQ-1: When a confirmed update run finishes, the update result shall carry one `run` record that lists every part of that run with its end state. · failure: none — the record only reads the finished result, and a part that failed is one of its lines.
- REQ-2 (REQ-1): The `run` record shall name each part a rerun can finish in `remaining` and each part left as it was on purpose in `held`.
- REQ-3 (REQ-2): Where `remaining` is not empty, the `run` record shall have status `partial` and name `moai update` as its next command.
- REQ-4 (REQ-2): When the `run` record has no remaining part, no held part and no written file that fails its read-back, the `run` record shall have status `complete`. · failure: none — it only sets a status.
- REQ-5 (REQ-1): When a confirmed run wrote something, the update shall write the `run` record to `installations/<install id>/update-run.json` in the HOME state folder, replacing the earlier file atomically.
- REQ-6 (REQ-5): IF a confirmed run wrote nothing, THEN the update shall leave the HOME state folder as it was and still return the `run` record in its result. · failure: none — only the file is left out.
- REQ-7 (REQ-5): IF the record cannot be written, THEN the update shall name the cause in `receipt_error` of the `run` record and keep the apply result unchanged. · failure: none — a missing receipt never turns an applied run into a failed one.
- REQ-8 (REQ-1): When a run wrote skill files, the update shall read each written file of the moai skill and of the skill folders beside it back from disk and compare its sha256 with the `after_sha256` the result names.
- REQ-9 (REQ-8): IF a written file does not read back as the digest the result named, THEN the `run` record shall mark that file with its check and have status `partial`. · failure: none — the checks are `missing`, `unreadable`, `not_a_file`, `too_large`, `unsafe_path` (a result path that leaves its skill folder is never read) and `mismatch`.
- REQ-10 (REQ-8): The `run` record shall list every part that wrote files it does not read back in `not_compared` with that part's own receipt, and claim no comparison for it.
- REQ-11 (REQ-1): When the end screen of a confirmed update is drawn, the screen shall take its head, its remaining-parts line with the command that continues the run, and its hash line from the `run` record. · failure: none — a result without a `run` record is drawn from its parts as before.
- REQ-12: The update result shall keep `mode`, `applied` and every per-part result as they were before this SPEC. · failure: none — `run` is an added key.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): 임시 HOME 에서 스킬이 쓰인 뒤 agents 부분에 실패를 주입하면 결과의 `run` 이 `partial` 이고 `remaining` 이 `["agents"]` 이며 `next` 가 `moai update` 이고, `update-run.json` 이 디스크에 같은 내용으로 남아요. 원인을 고치고 다시 실행하면 agents 가 적용되고 `run` 이 `complete` 이며 파일도 `complete` 로 바뀌고 세 번째 계획은 적용할 것이 없어요 · verify: `cargo test --locked --offline --test update_run_receipt_cli a_failure_behind_the_skills`
- AC-2 (REQ-8, REQ-10): 임시 HOME 의 새 설치에서 쓴 스킬 파일마다(moai 스킬과 옆 스킬 폴더) 검사가 하나씩 있고 모두 `match` 이며, 기록의 관찰 해시와 기대 해시가 디스크 바이트의 해시와 같고, `not_compared` 가 agents 를 이름으로 말해요 · verify: `cargo test --locked --offline --test update_run_receipt_cli every_written_skill_file`
- AC-3 (REQ-6): 쓸 것이 없는 확인된 실행은 `run.wrote` 가 false 이고 `receipt` 키가 없으며 HOME 의 파일 바이트가 실행 전과 같아요 · verify: `cargo test --locked --offline --test update_run_receipt_cli a_run_with_nothing_to_write`
- AC-4 (REQ-11): 끝 화면이 `run` 기록에서 그려져요. 남은 부분이 있는 실행은 "남은 부분" 줄과 이어서 실행하는 명령을 보이고, 깨끗한 실행은 읽어서 해시가 모두 맞다는 줄을 보이며, 부분들이 모두 깨끗해도 기록의 파일 하나가 쓴 대로 읽히지 않으면 머리가 일부만 적용으로 바뀌어요 · verify: `cargo test --locked --offline --test update_run_receipt_cli the_end_screen_is_drawn`
- AC-5 (REQ-2, REQ-3, REQ-4, REQ-7, REQ-9): 기록을 만드는 규칙의 단위 시험이 통과해요: 깨끗한 결과는 `complete` 이고 파일을 남기지 않음, 뒤 부분의 실패는 `remaining`, 충돌과 보존된 파일과 레이아웃 밖 링크는 `held`, 게이트가 건너뛴 부분은 `remaining`, 파일 대조의 `match`·`mismatch`·`missing`, 폴더 밖으로 나가는 경로는 읽지 않음 · verify: `cargo test --locked --offline --lib update::run_receipt`
- AC-6 (REQ-12): 이웃 시험이 그대로 통과해요: 사용자 계층 update, 끝 화면 사실, update 가드 · verify: `cargo test --locked --offline --test update_user_layer_cli`
  verify: `cargo test --locked --offline --test update_tux_golden`

## Scope

- write: `.moai/specs/SPEC-20261125-update-run-receipt/spec.md`
- write: `src/update/run_receipt.rs`, `src/update.rs`, `src/update_tux.rs`
- write: `tests/update_run_receipt_cli.rs`
- write: `tools/repo_scans.json` (shared file: only through `card line-lease`, the test is listed as a package reader)
- read-only: `src/apply.rs`, `src/home_root.rs`, `skills/moai/workflows/update.md`
- forbidden: 실제 HOME, 설치된 moai, 다른 프로젝트, 부분별 `*_result` 의 모양, `skills/moai/**` 의 글, 카드 번호를 이 파일 밖의 코드나 주석에 쓰는 일

## Plan

1. 새 시험을 먼저 썼어요: 단위 시험 7개(`update::run_receipt`)와 임시 HOME 시험 4개(`tests/update_run_receipt_cli.rs`)이고, 시험이 필요로 하는 `MOAI_TEST_FAULT` 지점(`update_part_agents`, `update_part_routing`, `update_part_style`)은 디버그 빌드에서만 켜져요.
2. `src/update/run_receipt.rs`: 결과 하나를 읽어 부분마다 끝 상태(`Done`, `Held`, `Open`)를 정하고, 쓴 스킬 파일을 읽어 해시를 대조하고, `status`·`remaining`·`held`·`files`·`not_compared` 를 가진 기록을 만들어요. 무언가를 쓴 실행만 `installations/<install id>/update-run.json` 으로 원자적으로 교체해요.
3. `src/update.rs`: `update_command_core` 가 `Ok(out)` 를 돌려주기 직전에 `out["run"]` 을 채워요. 게이트가 건너뛴 부분의 이유 문장은 기록과 같은 상수를 써요.
4. `src/update_tux.rs`: 끝 화면의 머리·"남은 부분" 줄·해시 줄을 기록에서 그려요.
5. `tools/repo_scans.json`: 새 시험을 패키지를 읽는 시험으로 올려요(`line-lease`).
6. 승인 뒤 새 head 에서 `check affected --card` 영수증을 다시 돌려 `REPAIRED` 로 알려요.

## Risks

- 새 파일 하나가 HOME 상태 폴더에 생겨요(tier 2): 쓰는 위치는 설치 id 폴더 안 한 파일이고, 그 폴더가 링크면 쓰기를 거절하고, 쓰기 실패는 적용 결과를 바꾸지 않고 `receipt_error` 로만 남아요(REQ-7).
- 해시 대조는 쓴 스킬 파일에만 해당해요. 다른 부분이 쓴 파일은 대조되지 않았다는 사실을 `not_compared` 로 드러내요(REQ-10). 이 구분이 흐려지면 기록이 하지 않은 검증을 한 것처럼 읽혀요.
- 기록의 `held`(충돌, 사용자 파일 보존, 레이아웃 밖 링크)는 다시 실행해도 바뀌지 않아요. 이를 `remaining` 에 넣으면 사용자가 끝나지 않는 재실행을 하게 되어, 기록은 둘을 나누고 `partial` 이어도 `next` 는 `remaining` 이 있을 때만 줘요.
- 처음 설계는 "남은 부분이 있으면 쓴다" 였어요. 그러면 모든 호스트가 막힌 실행도 파일을 쓰게 되어 기존 시험 둘이 지키는 "쓴 것이 없으면 HOME 이 그대로" 와 어긋나서, 쓴 실행만 남기게 했어요(REQ-5, REQ-6).
- 시험 지점은 디버그 빌드 전용이라 릴리스 빌드에는 주입 경로가 없어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기록은 `installations/<install id>/update-run.json` 한 파일이고 무언가를 쓴 실행만 교체해요(REQ-5, REQ-6). 대안: 실행마다 쓰거나 날짜별 파일을 쌓음. | 카드 기준 2 와 기존 시험의 계약("쓴 것이 없으면 HOME 이 그대로"). 승인 전. | `leaves_record` 와 `write_record` 호출을 지우면 결과 속 `run` 만 남아요. |
| D-2 | 일부러 남긴 부분(충돌, 보존된 사용자 파일, 레이아웃 밖 링크)은 `held` 로 나누고 `remaining` 에 넣지 않아요(REQ-2). | 다시 실행해도 같은 결과여서 `next` 를 주면 거짓이 돼요. 승인 전. | `End::Held` 를 `Open` 으로 바꿔요. |
| D-3 | 해시 대조는 쓴 스킬 파일(moai 스킬과 옆 스킬 폴더)만 하고 나머지는 `not_compared` 로 드러내요(REQ-8, REQ-10). 대안: 모든 부분의 파일을 대조. | 다른 부분은 결과에 기대 해시를 주지 않고 각자 영수증을 가져요. 대조 없이 대조했다고 말하지 않는 쪽을 골랐어요. 승인 전. | 해당 부분이 기대 해시를 주면 `check_files` 에 더해요. |
| D-4 | 쓰기 실패는 `receipt_error` 로만 알리고 적용 결과를 바꾸지 않아요(REQ-7). | 스킬은 이미 적용됐으므로 영수증 실패로 실행 전체를 실패로 말하면 거짓이에요. 승인 전. | `finish` 가 오류를 위로 올려요. |
