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

- AC-1 (REQ-1, REQ-2, REQ-3): a tip equal to the base is `fast_forward`; a head inside the tip is `already_merged`; a tip that moved by an unrelated change is `reuse` with its reasons.
  verify: `cargo test --locked --offline --test merge_reuse_cli decides_`
- AC-2 (REQ-4, REQ-5, REQ-6): a new test target, a changed test of the candidate's area, a changed scan list, an overlapping edit, a conflict and a base that is not an ancestor each give `rerun` with exactly the newly affected checks, or `rerun_all`.
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

## Result

(written after the checks)
