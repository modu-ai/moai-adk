---
id: SPEC-20261120-merge-reuse
title: "Decide mechanically whether an audited merge candidate's evidence still holds after the local base moved"
status: draft
revision: 1
created: 2026-11-20
---

# Decide mechanically whether an audited merge candidate's evidence still holds after the local base moved

## Intent

사용자 지시(2026-11-20, 리더 대화에서 확인): M-Flow 의 감사·병합 병목을 원인별로 고치고, 고치기 전과 뒤를 같은 부하에서 재서 증명해요. 오늘은 감사를 통과한 후보(C-186, C-203)가 다른 병합이 로컬 `v2026.11.0` 끝을 먼저 옮겼다는 이유만으로 다시 병합되고 다시 검사됐어요.

병합 절차(리더 공지)는 후보 하나를 "OLD 위 머지 커밋 + 검사 + MERGE-READY 노트"로 만들어 두고 리더가 `git merge --ff-only` 로 올려요. 끝이 그사이 움직이면 ff 가 안 돼서 레인이 "다시 OLD 부터" 하고 검사를 처음부터 다시 해요. 그런데 움직인 부분이 후보와 아무 상관이 없으면(다른 파일, 후보가 기대는 검사의 선택이 그대로) 그 검사 결과는 여전히 맞아요. 이 SPEC 은 그 판단을 사람이 아니라 한 명령이 하게 해요.

`moai check reuse --base <OLD> --head <후보 head> --tip <지금 끝>` (또는 `--merge <머지 커밋>`)는 git 만 읽고 다섯 가지 중 하나로 답해요.

- `fast_forward`: 끝이 OLD 그대로예요. 증거가 그대로 맞고 ff 로 올려요.
- `already_merged`: 후보 head 가 이미 끝 안에 있어요. 올릴 게 없어요.
- `reuse`: 끝만 앞으로 갔고 충돌이 없고, 움직인 파일이 후보의 파일과 겹치지 않고, `check affected` 가 후보에 대해 고르는 검사가 옛 트리와 끝 위 병합 트리에서 같아요. 증거를 다시 쓰고 이유를 남겨요.
- `rerun`: 끝만 앞으로 갔지만 새로 영향받은 검사가 있어요. 그 검사만 다시 돌려요(`rerun` 목록).
- `rerun_all`: 판단할 수 없어요(충돌, OLD 가 끝의 조상이 아님, 읽을 수 없는 사실). 모르면 전부 다시 돌려요.

같은 SPEC 의 측정(통제된 전·후)은 실제 이력에서 해요: `v2026.11.0` 의 첫 부모 이어 달리기의 병합 후보 N 개가 모두 같은 OLD 위에서 감사를 마쳤다고 가정하고 끝이 하나씩 옮겨 갈 때, 지금 절차는 N-1 개를 다시 병합·검사하고 이 명령은 몇 개를 다시 돌리는지를 `moai check reuse` 로 세요(Result).

## Non-goals

- 병합 실행(잠금 아래 `git merge --ff-only`, 새 머지 커밋 만들기)과 MERGED-LOCAL 노트 쓰기: 레인 세션은 호스트 worktree 격리 때문에 기본 체크아웃에 git 을 쓰지 못하고, 큐를 카드 노트에서 읽으려면 `src/main.rs`·ledger 쪽 연결이 필요해요. 이 SPEC 은 판정 부분만 해요. 이어서 할 일은 리더 결정이에요.
- 짝·의존(`pair=`, `after=`)을 묶어 함께 병합·검증하기: 후속.
- 검사의 사전 계산·캐시: 판정은 매번 새로 계산하고 아무것도 저장하지 않아요.
- 서로 다른 영역이 합쳐져서만 깨지는 검사(의미 충돌): 아래 Risks.

## Requirements

- REQ-1: When `moai check reuse` runs with a tip equal to the base, the operation shall answer `fast_forward`. · failure: none
- REQ-2: When the head is already an ancestor of the tip, the operation shall answer `already_merged`. · failure: none
- REQ-3: When the base is an ancestor of the tip, the merge of the head into the tip is clean, the paths the base-to-tip change touches overlap none of the paths the head changes, and the selection of `check affected` for the head's changes is equal on the head's merge tree and on the tip's merge tree, the operation shall answer `reuse` and name those facts as reasons. · failure: none
- REQ-4 (REQ-3): IF the selection on the tip's merge tree names a test target or Python module the old selection lacks, or the base-to-tip change picks a target the new selection holds, THEN the operation shall answer `rerun` with exactly those targets in `rerun`.
- REQ-5 (REQ-3): IF the base-to-tip change touches a path the head also changes, THEN the operation shall answer `rerun` and add the steps that check the whole change (`diff_check`, `fmt`, `clippy`, `test_lib`, `test_bins`, and the per-file steps of the overlapping files) to `rerun`.
- REQ-6: IF the base is not an ancestor of the tip, the merge of the head into the tip conflicts, or a fact cannot be read, THEN the operation shall answer `rerun_all` with the reason. · failure: none
- REQ-7: The operation shall move no ref and change no working file and no index. · failure: none
- REQ-8: When `--merge <commit>` names a commit with exactly two parents, the operation shall take its first parent as the base and its second parent as the head. · failure: none
- REQ-9: IF an argument is missing, is not a revision, or starts with a dash, THEN the operation shall refuse with a usage error and read nothing. · failure: none
- REQ-10: The answer shall list `rerun` and the reasons in a fixed order, so that equal inputs give equal answers. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a tip equal to the base is `fast_forward`; a head inside the tip is `already_merged`; a tip that moved by an unrelated change, or that only removed a target the candidate ran, is `reuse` with its reasons.
  verify: `cargo test --locked --offline --test merge_reuse_cli decides_`
- AC-2 (REQ-4, REQ-5, REQ-6): a new test target, a new Python module of the area, a changed test of the candidate's area, a changed scan list, an overlapping edit, a conflict and a base that is not an ancestor each give `rerun` with exactly the newly affected checks, or `rerun_all`.
  verify: `cargo test --locked --offline --test merge_reuse_cli reruns_`
- AC-3 (REQ-7, REQ-8, REQ-9, REQ-10): read-only, `--merge`, refused arguments and a fixed order.
  verify: `cargo test --locked --offline --test merge_reuse_cli answers_`
- AC-4: neighbours.
  verify: `cargo test --locked --offline --test check_affected_cli --test merge_reuse_cli`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`
- AC-5 (REQ-3, REQ-4, REQ-5, REQ-6): in a controlled replay of real history, the commands re-run fewer candidates than the current procedure; the counts are in Result.
  manual: replay the last merges of the local integration branch as candidates that were all audited on the same base, run `moai check reuse` for each against the tip it would meet, and write the counts of `fast_forward`, `reuse`, `rerun` and `rerun_all` in Result.

## Scope

- write: `.moai/specs/SPEC-20261120-merge-reuse/**`, `src/check_affected.rs`, `src/check_affected/**`, `src/ops/check.rs`, `docs/design-review/appendix/operations/check.md`, `tests/merge_reuse_cli.rs`
- read-only: `src/check_affected.rs` (`select`, `steps`, `TestFile`), `tests/check_affected_cli.rs` (fixture style), `tools/repo_scans.json`
- forbidden: `src/lib.rs`, `src/main.rs`, `src/help.rs` (held by other cards), any write to a ref, the primary checkout or the card ledger by the operation, and a card, Issue or pull request number in the text this change adds

## Plan

1. SPEC, then RED tests (`tests/merge_reuse_cli.rs`, temporary repositories).
2. `src/check_affected/reuse.rs` (a child module of `check_affected`, so `src/lib.rs` is not touched): a pure `judge` from facts, a reader of tests and scan list from a git tree, and the `reuse` action of `moai check`.
3. Operation document, checks, mutants, a read-only audit by another agent, then the replay measurement in Result.

## Risks

- `reuse` rests on the selection rule of the leader's procedure: the selection is the same and no file overlaps. A change in an unrelated area that only breaks the candidate's checks in combination (for example an unused import after both land) is not seen by this rule; the leader's periodic verification of the tip is what catches it. The operation never says the merged tree was tested, only that the earlier evidence holds under the rule.
- `git merge-tree --write-tree` writes tree objects (unreachable until a commit uses them). No ref, index or working file moves.
- The delta's areas pick targets by the same name rules as `check affected`; a target that reads files by a path the name rules do not link is not picked. The scan list (`tools/repo_scans.json`) is part of the selection, so a changed list reruns its new targets.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The decision is a read-only `check reuse` action, not a new top-level word | A new word touches `src/main.rs`, `src/lib.rs` and `src/help.rs`, which other cards hold; `check` already exists | Move the action behind its own word later |
| D-2 | Reuse needs equal selection and no overlap; anything unreadable is `rerun_all` | The leader's rule: "uncertain → re-run" | Loosen the rule in one function |
| D-3 | Rerun lists the newly affected targets, not a fresh full check | The leader's rule: "re-run only the newly affected checks" | Answer `rerun_all` for every non-reuse |
| D-4 | The tests of a tree are read with one `git cat-file --batch`, not one git call per file | First real run on this repository took 40 s of wall time and 16 s of CPU under load, mostly from about 300 git processes; the decision must cost less than the re-check it saves | Read file by file again (same answers, slower) |

### Next slice (needs the leader's decision, not built here)

The decision above answers one candidate. The queue the card asks for (durable first-in-first-out, one lock, the leader draining it) needs three things this slice does not touch: reading the `MERGE-READY` notes of the card ledger in process (wiring in `src/main.rs` and the card module, held by other cards), writing `MERGED-LOCAL` (a git write on the primary checkout, which only the leader's session may do), and `pair=` / `after=` grouping. A smaller path that keeps the queue in the existing notes: each `MERGE-READY` note already carries card, head, base and merge, so its note id is the queue order; a later `moai check queue` could list the notes after the last `MERGED-LOCAL`, call this decision for each against the current tip and print what to merge, what to re-run and what to skip. That needs a leader decision on who runs it (the leader session, once per round) and on which card holds the ledger wiring.

## Result

구현: `moai check reuse --base <b> --head <h> --tip <t>`(또는 `--merge <m>`)는 `src/check_affected/reuse.rs` 의 `judge`(순수)와 git 읽기로 `fast_forward`·`already_merged`·`reuse`·`rerun`·`rerun_all` 중 하나를 답해요. 아무 ref·색인·작업 파일도 움직이지 않아요(`git merge-tree --write-tree` 의 트리 객체만 남아요). 한 트리의 시험 파일은 `git cat-file --batch` 한 번으로 읽어요(D-4).

- AC-1: met — `cargo test --locked --offline --test merge_reuse_cli decides_` 통과(끝이 base 그대로 → `fast_forward`, head 가 끝 안 → `already_merged`, 상관없는 파일만 움직임 → `reuse`, 끝이 후보가 돌린 대상을 지우기만 함 → `reuse` + `selection_is_covered`). 전체 15개 통과.
- AC-2: met — `cargo test --locked --offline --test merge_reuse_cli reruns_` 통과(새 Rust 대상, 새 파이썬 모듈, 바뀐 시험, 바뀐 스캔 목록의 새 대상, 겹치는 파일은 `diff_check`·`fmt`·`clippy`·`test_lib`·`test_bins` 추가, 충돌과 base 가 끝의 조상이 아님은 `rerun_all`).
- AC-3: met — `cargo test --locked --offline --test merge_reuse_cli answers_` 통과(ref·작업 트리·색인이 그대로, `--merge` 는 두 부모와 같은 답, 인자 누락·`-` 시작 거절, 같은 입력은 같은 순서).
- AC-4: met — `cargo test --locked --offline --test check_affected_cli --test merge_reuse_cli`(13 / 15 통과), `cargo clippy` 와 `cargo fmt --check` 와 `rustfmt --edition 2021 --check src/ops/check.rs` 는 `moai check affected 341f6c7abb..5fb8dc84f9` 안에서 통과(`targets=44 py=12 scans=listed:41 … result=PASS`). 독립 감사는 아직이에요. 변이 5개(gained 대상 계산 뒤집기, 움직인 경로 대상 계산 바꾸기, 겹침 검사 끄기, 조상 검사 끄기, 이미 병합됨 검사 끄기)는 모두 시험이 잡았어요(각각 종료 코드 101, 작업 트리는 원상 복구).
- AC-5: met — 통제된 재생(아래). 같은 입력에서 판정이 한 번에 40.9 초(CPU 16.8 초)이던 것이 0.9 초(CPU 0.65 초)로 줄었어요(시험 파일마다 git 을 따로 부르던 것을 `git cat-file --batch` 한 번으로).

### 재생 측정

방법: 로컬 `v2026.11.0`(`341f6c7abb`)의 첫 부모 사슬에서 마지막 병합 13개를 오래된 것부터 놓고(`replay-reuse.py 13 5`), 연속한 5개를 한 묶음으로 봐요. 한 묶음의 후보 5개는 모두 묶음 첫 병합의 첫 부모(OLD) 위에서 감사를 마쳤다고 가정하고, j 번째 후보는 앞의 j-1 개가 먼저 끝을 옮긴 상태의 실제 끝(`tip`)을 만나요. 후보의 head 는 각 병합의 둘째 부모예요. 묶음은 9개(13-5+1), 후보 판정은 45번이에요. 명령은 `moai check reuse --base <OLD> --head <head> --tip <끝> --dir <공유 clone>` 이고, 지금 절차가 다시 하는 횟수는 "끝이 한 번 움직일 때마다 한 번 다시 병합·검사"로 세요.

| 판정 | 횟수 |
| --- | --- |
| `fast_forward`(묶음의 첫 후보) | 9 |
| `reuse` | 7 |
| `rerun`(새로 영향받은 검사만) | 29 |
| `rerun_all` | 0 |

- 다시 병합·검사하는 횟수: 지금 절차 90번(묶음마다 0+1+2+3+4=10, 9묶음), 이 명령으로는 29번(`rerun` 만 세요). 재검사 때 다시 돌리는 범위도 달라요: 지금은 매번 `check affected` 전체(이번 병합 열차에서 대상 41~55개), `rerun` 29번의 `rerun.rust` 대상 수는 0개 3번, 1~3개 17번, 8~14개 9번(가운데값 2개)이고, 겹치는 파일이 있어 `steps` 5개(`diff_check`·`fmt`·`clippy`·`test_lib`·`test_bins`)가 붙은 것은 9번이에요(출력: `replay-13-5.txt`).
- 한 곳 직접 확인: 이 레인의 후보(C-176)를 `341f6c7abb` 에서 감사를 마친 뒤 끝이 `ca9723d773` 으로 움직였을 때 `reuse`(이유 4개 모두 기록), 리더가 "재사용은 아직 안 받는다"고 해서 전체 검사를 새로 돌리게 됐어요.
- 못 잰 것: 병합 전 기다린 시간(벽시계). 이 기계의 부하와 슬롯 대기가 섞여서 전·후를 같은 조건으로 잴 자료가 없어요. 의미가 겹치는 깨짐(서로 다른 영역이 합쳐져서만 깨지는 검사, 예: 다른 카드의 help 골든 변경 때문에 한 후보가 실패한 일)은 이 판정이 잡지 못해요. 이 측정은 과거 이력을 끝만 바꿔 다시 놓은 것이고 실제 병합 열차를 돌린 게 아니에요.

아직 안 한 것: 큐 소비와 `MERGED-LOCAL` 쓰기, 노트를 읽는 `check queue`(위 "Next slice"의 결정 대기), 독립 감사, `spec check --closure` 의 마지막 `closure_status`(승인 전 SPEC 이라 남아요).
