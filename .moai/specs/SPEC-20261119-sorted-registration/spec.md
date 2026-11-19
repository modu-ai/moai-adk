---
id: SPEC-20261119-sorted-registration
title: "Fewer same-place conflicts: generated ops table and sorted shared lists"
status: draft
revision: 1
created: 2026-11-19
---

# Fewer same-place conflicts: generated ops table and sorted shared lists

## Intent

PR 히스토리 분석(2026-11-17)에서 `src/main.rs`(62, 같은 자리 삽입 21)·`src/lib.rs`(23, mod/use 20)·
`docs/rust-product-path.md`(45, 표 행 30)이 충돌 상위권이에요. 카드 C-75 는 계획 우선이라 이 SPEC 의
전달물은 측정이 들어간 계획(`research.md`)이에요. 구현 슬라이스(조각 스키마·행 이주·생성기)는 계획의
권고를 리더가 고른 뒤 후속 SPEC 으로 가요. `src/ops/` 추가 충돌은 이미 build.rs 레지스트리(#2617)로
해제돼 있어요.

## Non-goals

- 생성기·행 이주의 구현: `research.md` C-1 의 조각 형식 결정(리더)이 먼저예요.
- `src/lib.rs`·`src/main.rs` 정렬: 측정이 이득 없음을 보여 C-2 로 넘겼어요.
- `src/ops/` 레지스트리 변경(#2617 이 이미 목표 달성).

## Requirements

- REQ-1: When `research.md` is read, it shall hold a before/after table for at least ten historical merge pairs whose parents still conflict today in at least one of `src/main.rs`, `src/lib.rs`, `docs/rust-product-path.md`, each pair naming its merge commit, the file, the before conflict count and the after count under each proposed design, and the method that produced them.
- REQ-2 (REQ-1): IF a proposed design has no measured pair behind it in the table, THEN the plan shall not present that design as measured anywhere.
- REQ-3: When the plan's conclusions are read, they shall state a recommendation for the docs table with its measured gain, an explicit statement about sorting `src/lib.rs` and `src/main.rs` with its measured result, and the follow-up slices with what each needs decided.
- REQ-4 (REQ-3): IF a follow-up slice needs a decision that is not this lane's to make, THEN the plan shall name which decision it waits on.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): research.md 의 쌍이 10 이상이고 모든 after 가 before 를 넘지 않으며, 결론 절이 C-1(권고와 측정 이득)·C-2(정렬 측정 결과)·C-3(이미 해제된 것)을 이름해요.
  verify: `cargo test --offline --locked --test ops_table_gen_cli`
- AC-2 (REQ-1): 측정 명령이 오늘 나무에서 재실행되어 같은 12쌍과 같은 수를 내요(스크립트가 target/ 에 있고, research.md 방법 절이 명령을 이름해요).
  verify: `zsh target/find_conflicted_merges.sh` + `python3 target/simulate_sorted.py`(작업나무에서, 커밋 안 함)

## Scope

- write: `.moai/specs/SPEC-20261119-sorted-registration/**`, `tests/ops_table_gen_cli.rs`
- read-only: `src/main.rs`, `src/lib.rs`, `build.rs`, `docs/rust-product-path.md`, `docs/design-review/appendix/operations/*.md`, `target/`(측정 스크립트, 커밋 안 함)
- forbidden: 실제 HOME·원장, `src/ops/` 레지스트리 변경, main.rs/lib.rs 재정렬(측정이 반대), 문서 표의 이번 슬라이스 변경

## Plan

1. 병합쌍 탐색·측정 스크립트를 target/ 에 쓰고 12쌍을 뽑아 before/after 를 측정했어요(완료).
2. research.md 를 쓰고 이 SPEC 을 계획-우선으로 다시 썼어요(완료).
3. `tests/ops_table_gen_cli.rs` 로 research.md 의 표·결론을 핀해요(RED→GREEN).

## Risks

- merge-tree 재판정이 당시와 다를 수 있어요 — research.md 방법 절이 "오늘 같은 나무에서"라고 밝혀요.
- 리더가 C-1 대신 행 유지를 고르면 이 계획의 후속 슬라이스가 바뀌어요 — 대안은 C-1 절의 비용 표기로 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이 SPEC 은 계획(측정)만 전달하고 구현은 후속으로 | 카드 본문 "Plan first", 측정이 정렬 구현을 반대(C-2) | 리더가 C-1 을 고르면 후속 SPEC 에서 구현 |
| D-2 | after 설계는 정렬과 행-제거 둘 다 측정 | 정렬만으로는 12쌍 중 2쌍, 행-제거는 표 8쌍 중 7쌍 해제 — 근거 비교 필요 | 측정 스크립트 변형으로 추가 설계 비교 |
| D-3 | 표 원료를 조각으로 옮기는 형식은 이 SPEC 이 정하지 않음 | 조각에 구조화 블록이 없어 형식 결정이 리더 몫(research.md C-1 비용) | 후속 SPEC 의 REQ |

## Result

- AC-1: met — `cargo test --offline --locked --test ops_table_gen_cli` 3 passed/0 failed(2026-11-19, worktree moai-adk-c75): 12쌍(≥10) 전 after ≤ before, C-1·C-2·C-3 문장과 7/8 수치, 모든 행이 병합 커밋·공유 파일을 이름함.
- AC-2: met(재실행 가능) — `zsh target/find_conflicted_merges.sh` → 12쌍, `python3 target/simulate_sorted.py` → 위 표와 같은 수(2026-11-19 두 번 실행). 스크립트는 target/(무시 경로, 커밋 안 함), 방법 절이 명령을 이름해요.
- 남은 것: research.md C-1 의 조각 형식 결정(리더) → 후속 SPEC 으로 생성기·행 이주 구현. lib.rs·main.rs 정렬은 측정상 기각(C-2).
