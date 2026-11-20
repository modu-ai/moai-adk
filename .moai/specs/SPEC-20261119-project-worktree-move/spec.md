---
id: SPEC-20261119-project-worktree-move
title: "The PROJECT route reports a worktree inside the Go .moai folder with the worktree-move choice"
status: draft
revision: 1
created: 2026-11-19
---

# The PROJECT route reports a worktree inside the Go .moai folder with the worktree-move choice

## Intent

카드 C-59 의 기준 3 이에요: "Go .moai 안의 worktree 를 worktree 이동 선택으로 보고". tip(035f62dd1d)에서 읽은 사실이에요. `moai go-state-plan` 은 Go 세대의 `.moai/` 아래에 git worktree 가 등록되어 있으면 그 폴더를 `leave` 로 두고 이유에 "git worktree(s) are registered under .moai/ … (worktree-move-plan, #2723)" 를 적어요(`src/go_state.rs` 의 `go_folder` 항목). 폴더를 옮기면 git 이 그 worktree 를 잃기 때문이에요. 그런데 PROJECT 라우트(`skills/moai/workflows/project.md` 단계 2)는 이 항목을 사용자에게 어떻게 보고하고 어떤 선택을 주는지 말하지 않아요. 읽기 전용 `moai worktree-move-plan` 과, 사용자가 고른 뒤에만 도는 `moai worktree-move-apply` 는 이미 있어요(`src/worktree_move.rs`, `src/ops/worktree_move_plan.rs`, `src/ops/worktree_move_apply.rs`, 시험 `tests/worktree_move.rs`). `project.md` 는 이 apply 를 닫는 커밋 단계(8)의 목록에서만 이름을 불러요.
이 SPEC 은 `project.md` 의 글과 그 글이 인용하는 이름·이유가 코드와 같다는 시험만 다뤄요. 새 Rust 명령은 없어요.

## Non-goals

- 기준 2(MCP plan/apply)는 하지 않아요: 리더 결정 2521 과 note 3437.
- `worktree-move-plan`·`worktree-move-apply`·`go-state-plan` 의 코드, 이유 문구, 인자 바꾸기: 읽기만 해요.
- 관측 파일(`holds`)을 만드는 새 도구: 쓰지 않아요. 글은 봤던 것만 파일에 쓰라고 말해요.

## Requirements

- REQ-1: When `moai go-state-plan` reads the Go `.moai` folder as `leave` because git worktrees are registered under it, the PROJECT workflow shall report that reason to the user in the plan's own words and name the worktree move as the way to let the folder move. · failure: none — the text is checked for the quoted reason and for the named way, and a missing sentence fails that test.
- REQ-2: The PROJECT workflow shall run the read-only `moai worktree-move-plan <project>` and show its items, `totals` and each item's reasons before it asks anything about moving a worktree.
- REQ-3: The PROJECT workflow shall ask one selection for the worktree move, with the move first and `leave them where they are` second. · failure: none — the text is checked for the order of the two options, and a swapped order fails that test.
- REQ-4: The PROJECT workflow shall run `moai worktree-move-apply` only after the user picked the move. · failure: none — the text is checked for the apply coming after the pick, and a missing sentence fails that test.
- REQ-5 (REQ-2): IF the plan reads every worktree as `blocked` with `hold_check_not_run` because no holds observation was given, THEN the PROJECT workflow shall report the move as unavailable with that reason.
- REQ-6 (REQ-5): IF the PROJECT workflow has read no branch holds in the run, THEN it shall not write an empty holds list.
- REQ-7: The PROJECT workflow shall name the plan's reason, the blocked reason, the plan fields and the apply arguments exactly as the code prints and accepts them. · failure: none — the names are checked against the source files by a test.

## Acceptance

- AC-1 (REQ-1): `project.md` 단계 2 가 `go-state-plan` 의 "git worktree(s) are registered under .moai/" 이유를 그 말 그대로 보고하고 worktree 이동을 폴더가 옮겨지게 하는 길로 이름 붙인다는 것을 말해요.
  verify: `cargo test --locked --offline --test project_worktree_move_choice go_state_worktree_reason_is_reported_with_the_worktree_move_choice`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5, REQ-6): `project.md` 가 읽기 전용 `worktree-move-plan` 을 먼저 돌려 항목·`totals`·이유를 보이고, 이동을 먼저 두는 선택 하나를 묻고, 고른 뒤에만 `worktree-move-apply` 를 돌리며, `hold_check_not_run` 이면 이동을 쓸 수 없다고 보고하고 관측하지 않은 빈 holds 를 쓰지 않는다는 것을 말해요.
  verify: `cargo test --locked --offline --test project_worktree_move_choice worktree_move_is_previewed_then_asked_once_then_applied_only_after_the_pick`
- AC-3 (REQ-7): 글이 인용하는 이유·필드·인자 이름이 `src/go_state.rs`, `src/worktree_move.rs`, `src/ops/worktree_move_apply.rs` 에 글자 그대로 있어요.
  verify: `cargo test --locked --offline --test project_worktree_move_choice quoted_names_are_the_ones_the_code_prints`
- AC-4 (REQ-1…REQ-7): 이웃 검사.
  verify: `cargo test --locked --offline --test project_flow_closing`; `cargo test --locked --offline --test project_preview_order`; `cargo test --locked --offline --test worktree_move`; `cargo test --locked --offline --test go_state_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test doc_references`; `python3 -m unittest tests.source_manifest_test tests.skill_references_test`

## Scope

- write: `.moai/specs/SPEC-20261119-project-worktree-move/spec.md`, `tests/project_worktree_move_choice.rs`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/go_state.rs`, `src/worktree_move.rs`, `src/ops/worktree_move_plan.rs`, `src/ops/worktree_move_apply.rs`, `tests/worktree_move.rs`, `tests/go_state_cli.rs`, `tests/project_flow_closing.rs`
- forbidden: 실제 HOME, 사용자의 전역 git 설정, push 이외의 원격 쓰기, 사용자의 worktree 를 시험에서 옮기기

## Plan

1. 시험 먼저: `tests/project_worktree_move_choice.rs` 의 시험 셋을 쓰고, 글 시험 둘이 실패하는 것(RED)을 읽어요. 이름 시험은 코드에 이미 있는 이름만 보므로 처음부터 통과해야 해요.
2. `project.md` 단계 2 에 항목 하나를 더해 GREEN 으로 만들고, 매니페스트 해시는 마지막 커밋에서만 써요(예약 경로를 최소로).
3. 이웃 검사(AC-4)와 READY 전 검사 전부를 돌려요.

## Risks

- 글 시험은 문장을 고정해서 나중에 문구를 고치면 시험도 같이 고쳐야 해요 → 규칙을 이루는 문장(이유 보고, 읽기 전용 미리보기 먼저, 선택 순서, 고른 뒤에만 apply, `hold_check_not_run`)과 인용한 이름만 고정해요. 문장 전체가 아니에요.
- 코드의 이유·필드·인자 이름이 바뀌면 글의 인용이 틀려져요 → AC-3 이 글자까지 맞춰서 바로 실패해요.
- `project.md` 는 기준 1 가지(`spec/20261119-project-preview-order`)도 고쳐요 → 두 가지는 다른 단락을 고쳐서 합쳐도 충돌이 없는지 `git merge-tree` 로 읽어요. 매니페스트 해시는 합칠 때 다시 써야 할 수 있어요.
- 매니페스트 경로는 마지막 커밋 하나에만 예약하고 바로 풀어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 명령이나 새 단계를 만들지 않고 단계 2 에 항목 하나를 더해요 | 두 명령이 이미 있고 `project.md` 는 apply 이름을 단계 8 에서 부르기만 해요; 같은 규칙을 두 곳에 쓰면 어긋나요 | 항목 하나를 지워요 |
| D-2 | 기준 1 감사(note audit-c59a-aaedc57)의 N1·N2 문구 고침이 이 조각에 같이 실려요: Completion Check 의 merge 단계 대기 이유를 backup·install 두 가지로 정확히 쓰고, 단계 5 에 "이미 yes 한 앞 단계에만 해당, 거절하거나 뺀 앞 단계는 그대로 두고 뒤 단계는 이유와 함께 보고"를 더해요; 두 문장은 `tests/project_preview_order.rs` 에 묶여요 | 같은 파일을 한 번만 고치려고, 그리고 기준 1 가지는 이미 감사를 마쳐서 다시 밀지 않으려고 | 두 문장과 시험 줄 둘을 지워요 |
| D-3 | holds 관측 파일은 이 라우트가 본 가지 보류만 적고, 보지 않았으면 이동을 쓸 수 없다고 보고해요 | `worktree-move-plan` 은 관측 없이는 모든 worktree 를 `hold_check_not_run` 으로 막고, 빈 목록은 "막힌 가지 없음을 관측했다"는 뜻이에요 | 문장 하나를 지워요 |

## Result

- AC-1: met — `cargo test --locked --offline --test project_worktree_move_choice go_state_worktree_reason_is_reported_with_the_worktree_move_choice` passes (RED 먼저: 3 중 3 failed, exit 101, 글에 문장이 없어서). 단계 2 가 `go-state-plan` 의 "git worktree(s) are registered under .moai/" 이유를 그 말 그대로 보고하고 worktree 이동을 폴더가 옮겨지게 하는 길로 이름 붙여요.
- AC-2: met — 같은 시험 파일 `worktree_move_is_previewed_then_asked_once_then_applied_only_after_the_pick` passes. 읽기 전용 `worktree-move-plan` 미리보기 → 선택 하나(이동 먼저) → 고른 뒤에만 `worktree-move-apply` 순서, `hold_check_not_run` 이면 이동을 쓸 수 없다고 보고하고 관측하지 않은 빈 holds 를 쓰지 않는다는 것을 말해요.
- AC-3: met — 같은 시험 파일 `quoted_names_are_the_ones_the_code_prints` passes. 인용한 이유·`folder_stays`·`hold_check_not_run`·`totals`·`plan_digest`·`items`·`reasons`·`destination`·`root` 와 apply 인자 네 개가 `src/go_state.rs`, `src/worktree_move.rs`, `src/ops/worktree_move_apply.rs` 에 글자 그대로 있고 git 에 `--force` 를 넘기지 않아요.
- Mutation (committed 먼저, 커밋 275f724a11): 변이 5 개를 `project.md` 에 하나씩 넣고(old 문구가 정확히 1번인지 확인) 컴파일된 시험 바이너리를 돌린 뒤 `git restore` 했어요: M1 선택 순서 뒤집기, M2 apply 가 선택을 기다리지 않음, M3 관측하지 않은 빈 holds 허용, M4 인용한 이유를 코드와 다르게, M5 "Without a pick nothing moves" 삭제. 첫 실행에서 M2 가 살아남았어요(같은 문구가 `go-state-apply` 단락에도 있어서): 시험 문구를 `worktree-move-apply` 줄에 묶고 다시 돌려 5 개 모두 exit 101 로 죽었어요. 복원 뒤 파일이 커밋본과 같아요.
- 기준 1 감사 N1·N2 문구 고침(D-2): `tests/project_preview_order.rs` 의 두 문구 핀이 GREEN 이고, 변이 3 개(Completion Check 를 install 이유만으로 되돌림, 단계 5 의 "거절한 앞 단계는 그대로" 문장 삭제, 그 문장을 "적용한다"로 바꿈)가 모두 exit 101 로 죽었고(커밋 127025a547 뒤), 복원 뒤 파일이 커밋본과 같아요.
- 이웃: `go_state_cli` 61, `project_flow_closing` 10, `project_update_plan_cli` 59, `update_route` 9, `worktree_move` 12 passed(exit 0), `project_preview_order` 3, `project_worktree_move_choice` 3 passed. READY 전 검사 전부(1~7, 등재 명령, 같은 영역 시험)의 exit code 는 카드 note 에 적어요.
- Not run here: 전체 `cargo test`; 실제 worktree 가 `.moai/` 아래 있는 프로젝트에서 `/moai project` 세션이 이 글을 읽는 것(시험은 문구와 코드의 인용이 맞는지만 고정해요).
