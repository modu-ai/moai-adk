---
id: SPEC-20261121-project-state-move
title: "One state folder: the update flow moves .moai-cli to .moai with a preview, a token, a receipt and a revert"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# One state folder: the update flow moves .moai-cli to .moai with a preview, a token, a receipt and a revert

## Intent

오늘 두 상태 폴더가 공존해요: `state_root::dir` 는 유효 마커가 있는 프로젝트만 `.moai` 로 읽고(src/state_root.rs:566-572), 마커가 없으면 예전 세대의 `.moai-cli` 를 답해요. `.moai/` 는 "마커를 마지막에 쓰는 이동"으로만 생겨요(go-state 가 Go 내용을 옮길 때, adopt_new 가 새 프로젝트에 표지를 쓸 때) — 예전 Rust 세대 `.moai-cli` 를 그대로 쓰는 프로젝트가 `.moai` 로 넘어가는 이동은 아직 없어요. 이 조각이 그 이동을 만들어요: `/moai update` 미리보기에 이동 단계를 얹고(go-state 의 plan → 토큰 적용 → revert 계약과 update_cleanup 의 영수증 계약을 따라), 적용 뒤엔 `.moai-cli` 를 남기지 않아요. 리더 결정(2026-11-20 노트 4593): 이 조각의 배선 파일이 활동 카드 손안이라(src/lib.rs·src/help.rs·src/project_update.rs·src/state_root.rs), 준비를 먼저 하고 배선은 `.moai` 레이아웃 규칙 카드와 update 미리보기 카드가 착지한 뒤 잇는다 — 이 SPEC 의 이동·무시 규칙 세부는 그 레이아웃 카드가 내리는 규칙을 따라요.

## Non-goals

- Go 세대 `.moai/` 내용의 이전(go-state 가 이미 하고), 그 되돌리기의 변경.
- `.moai-cli/worktrees` 등 옛 worktree 자리의 이동: 설계 보고서가 정리 후보로만 보게 해요 — 이동하지 않고 `keep` 사유로 보여요.
- clone 공유 상태(`<git-common>` 아래)와 `~/.moai-cli`: 카드 본문이 뒤 조각으로 남긴 것.
- `.moai-cli` 안 무엇을 지우는가: 이 조각은 지우지 않고 옮기거나 두거만 해요.

## Requirements

- REQ-1: The state-move plan shall list every top-level entry of the earlier `.moai-cli/` with an action of `move` or `keep` and one reason each, reading only, for a project whose state-root verdict names the earlier generation. · failure: none — a plan over the entry limit keeps the whole list and says so.
- REQ-2: The plan shall give an entry whose target name already exists under `.moai/` the action `keep` with the collision named, so a Go-generation leftover is never overwritten by a move. · failure: none — the plan does not read the contents of a Go-generation `.moai/` beyond the top-level names it reports.
- REQ-3: The state-move apply shall run only with the plan's confirm token recomputed at apply time, move each move entry under its target name, write the state-root marker last, and record a manifest of original and new paths with a receipt whose status starts pending. · failure: none — a changed folder refuses the apply and asks for a fresh preview.
- REQ-4: When nothing keeps, the apply shall remove the emptied earlier state folder, so the state-root folder answer becomes .moai for every later read; a keep entry leaves the folder in place and the receipt names it. · failure: none — the removal is of an empty folder only.
- REQ-5: The revert shall take a stamp, read that receipt, move the recorded entries back to their original paths, remove the marker and the receipt's `.moai` additions it made, and mark the receipt `reverted`. · failure: none — the revert refuses a receipt that is not `committed`.
- REQ-6: While the state-move step of the project update preview is ready, the steps that read the state folder shall wait for the move with their own wait reason, and the step itself prints the apply command of its plan. · failure: none — the preview writes nothing.
- REQ-7: The plan shall refuse a link or a special file and name it as a keep entry, and the move applies the same refusal at apply time. · failure: none — the checks run in both the preview and the apply.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-7): a ledger project with `.moai-cli/` state and a Go-generation `.moai/` leftover shows every entry with its action, the colliding entry `keep` with the collision, a link `keep`, and nothing written.
  verify: `cargo test --locked --offline --test state_move_cli` — the preview scenes
- AC-2 (REQ-3, REQ-4): the apply with the token moves the entries, writes the marker last, empties and removes `.moai-cli/`, and the receipt reads `committed` with the manifest; `state_root::dir` answers `.moai` after.
  verify: `cargo test --locked --offline --test state_move_cli` — the apply scenes
- AC-3 (REQ-5): the revert moves the entries back, removes the marker, marks the receipt `reverted`, and a second revert refuses.
  verify: `cargo test --locked --offline --test state_move_cli` — the revert scenes
- AC-4 (REQ-6): the project update preview shows the state-move step with its apply command and makes the state-reading steps wait while it is `ready`.
  verify: `cargo test --locked --offline --test project_update_plan_cli` — the step scene
- AC-5: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-project-state-move/**`, `src/state_move.rs` 또는 그 선언집이 놓이는 자리, `src/ops/state_move_*.rs`, `tests/state_move_cli.rs`, `docs/design-review/appendix/operations/state_move*.md`, 그리고 배선이 이어질 때 `src/project_update.rs`(예약에 그때 추가)
- read-only: `src/state_root.rs`, `src/update_cleanup.rs`, `src/go_state.rs`
- forbidden: 항목 삭제, 링크·특수 파일 추적, Go 세대 `.moai/` 내용 덮어쓰기, 카드 번호의 출력 사용

## Plan

1. 준비(리더 결정 (a)): 이 SPEC 과 op 파일 넷·명령 문서·CLI 시험 초안을 예약 스코프 안에서 써요.
2. 배선 이어가기: 모듈 선언집(src/lib.rs 또는 프로젝트 update 모듈)과 update 미리보기 단계가 착지하는 대로 예약에 추가하고 RED → GREEN 으로 잇는다.
3. READY 전체 검사 → 감사는 리더에게 → READY. 뮤테이션은 토큰·마커-마지막 순서를 건드리는 형태로 상한 안에서.

## Risks

- 이동·무시 규칙의 세부는 같은 상태 폴더 가족의 레이아웃 카드가 정하는 규칙에 달려요: 이 SPEC 은 계약 수준(아이템·사유·충돌·마커-마지막)으로 고정하고 세부 목록은 그 카드 착지를 읽어 확정해요.
- `.moai-cli/worktrees` 같은 큰 자리는 옮기지 않아요: running 세션이 읽는 자리를 옮기면 망가져요 — `keep` 사유로 보이고, 정리는 설계 보고서의 뒤 걸음이에요.
- 두 폴더가 섞인 프로젝트(`mixed`·`interrupted` 판정)는 state_root 이 이미 막아요: 이 조각은 그 판정을 그대로 존중하고 자체 판단을 더하지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이동 계약은 update_cleanup 의 plan→토큰→영수증→revert 계약과 go_state 의 신중한 이동(rename 거부 사유·마커 마지막)을 따른다 — 새 계약을 만들지 않는다. | src/update_cleanup.rs:1-24, src/go_state.rs:3814, src/state_root.rs:630 | 카드 열림 |
| D-2 | 배선 파일의 착지 순서는 리더 결정을 따른다 — 준비물은 예약 스코프 안에서 먼저 쓴다. | 노트 4593 | 카드 열림 |

## Result

(status draft — written when the slice lands)
