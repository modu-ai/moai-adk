---
id: SPEC-20261120-board-observation-ops
title: "BOARD 05 — observation replay, retention, restart, wake records, screen tidy"
status: draft
revision: 1
created: 2026-11-20
parent: SPEC-20261120-board-observation-store
---

# 관측 운용 — 재생·보관·중단 뒤 재시작·깨우기 기록·화면 정리 (BOARD 05)

## Intent

기준 문서 `~/.moai/reports/moai-adk/collab-board-final-design-2026-11-20.md`(3판) 5.3절(수집기 표의 재시작·실패·깨우기 줄)·6.1절(BOARD 05 행)·7절(실패 주입). 관측 저장(`SPEC-20261120-board-observation-store`, BOARD 02)과 수집기·타임라인 위에서 **운용 반쪽**을 정해요: 관측 저장을 지우고 다시 모아도 결과가 같고(재생), 30일 보관이 수집 길에서 돌며 원장은 영구이고, 수집을 강제로 끊고 다시 돌려도 중복·누락 표시가 맞으며, 관측 저장이 없거나 깨져도 타임라인이 원장만으로 동작하고, 깨우기 전달 실패가 타임라인에 보여요. 이 조각(1)은 SPEC·PLAN 까지 — 시험은 수집기(관측 저장 카드 조각 3)와 타임라인(BOARD 01)이 착지한 뒤 그 계약에 맞춰 돌아가요.

## Non-goals

- 수집기·타임라인 자체의 구현 — 각각 관측 저장 카드 조각 3·BOARD 01.
- 원장 스키마 v2, 깨우기 전달 경로의 변경 — 승인 범위 밖.
- 로그 원문·비밀값·대화 원문의 어떤 기록 — 참조와 해시만(관측 저장 SPEC 과 같은 금지).
- 실제 원장·실제 HOME·호스트 훅.

## Requirements

- REQ-1: When the observation store is deleted and collection runs again over the same sources, the resulting rows shall be the same as before the deletion · failure: none — replay never duplicates and never invents; rows older than the retention cutoff are gone by age, not by accident.
- REQ-2: When retention runs on the collection path, pruning shall report how many observations went and leave the ledger file byte-for-byte as it was · failure: none — a failed prune stops no other collection step.
- REQ-3: When collection is killed mid-run and runs again, the duplicate and gap marks shall come out correct · failure: none — the same id with the same content is a no-op, the same id with different content stays a named conflict, and a hole becomes a gap@1 row from a bounded re-snapshot, never a guess.
- REQ-4: When the observation store is missing or broken, the timeline shall still answer from the ledger alone and mark the observation side as the last observation time with unknown · failure: none — a broken store never blocks the ledger read and never crashes the renderer.
- REQ-5: When a wake is delivered twice, the answer note shall be one, keyed by the originating ledger event id, with an unconfirmed delivery recorded as a wake@1 row carrying the failure and shown in the timeline · failure: none — the transporter's acknowledgement is not a delivery confirmation.
- REQ-6: When the timeline renders observation-derived marks, gaps and unknowns shall be one tidy line each with the observation time · failure: none — no raw observation content spills into the screen.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 수집을 강제로 끊고 다시 돌려도 중복과 누락 표시가 맞고, 저장을 지우고 다시 모은 결과가 같으며, 보관이 지나간 행만 지우고 원장은 바이트까지 그대로.
  verify: `cargo test --locked --offline --test board_observation_ops_cli`(시험 파일 이름은 후보; 수집기·타임라인 착지 뒤 실행 가능)
- AC-2 (REQ-4): 관측 저장이 없거나 깨진 상태에서 타임라인이 원장만으로 동작하고 관측쪽을 마지막 관측 시각·미확인으로 표시.
  verify: 같은 시험 파일의 저장 결손 시험
- AC-3 (REQ-5, REQ-6): 깨우기 두 번에 답 노트 하나, 전달 실패가 wake@1 행과 타임라인에 보이며, 표시는 관측 시각이 든 한 줄.
  verify: 같은 시험 파일의 깨우기·표시 시험

## Scope

- write(구현 조각, 리더가 PLAN 을 읽은 뒤): 운용 시험 파일(이름 후보 `tests/board_observation_ops_cli.rs`), 깨우기 기록·표시 정리 코드(수집기·타임라인 착지 뒤 그 계약에 맞춰), 이 SPEC, 운용 문서 문단
- read-only: `src/board/observation.rs`(BOARD 02), 수집기(관측 저장 카드 조각 3, 착지 뒤), 타임라인(BOARD 01, 착지 뒤), C-141 이 정하는 pr-state@1·bundle-run@1 필드 이름
- forbidden: 원장 스키마·표·색인 변경, 읽기 명령의 어떤 쓰기, 호스트 훅 설치, 실제 원장·실제 HOME, 카드 번호를 SPEC id·가지·커밋·PR 에 사용, 타인 회수 3회 한도·4시간 임대 변경

## Plan

1. 조각 1(이번, 코드 없음): SPEC + PLAN 노트 — 리더가 읽고 조각 2 를 정해요.
2. 조각 2(운용 시험): 7절의 세 주입(중단 뒤 복구·깨우기 두 번·깨우기 전달 실패) + 저장 결손·재생 동일성·보관 — 수집기와 타임라인이 착지한 뒤 RED 로 먼저.
3. 조각 3(기록·표시 정리): 깨우기 실패의 wake@1 기록과 타임라인 한 줄 표시 — BOARD 01 렌더러와 계약을 맞춰.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 재생 동일성은 저장 수준 멱등(관측 저장 SPEC REQ-2) 위에서 수집 수준으로 잰다 | 같은 근거를 두 층에 다시 증명하지 않는다 | 저장 시험만으로 |
| D-2 | 깨우기 답의 하나 보장은 답 노트 ref 의 키(원인 원장 이벤트 id)로 | 설계 5.3 깨우기 줄 | 별도 중복 제거 층 |
| D-3 | 저장 결손 시 타임라인은 원장만으로 답하고 관측쪽 표시는 마지막 관측 시각·미확인 | 설계 5.3 실패 줄 | 빈 관측으로 채우기 |
| D-4 | 표시 정리는 BOARD 01 렌더러와 한 계약(관측 시각 든 한 줄) | 6.1 화면 정리 | 표 형태 |

## Risks

- 수집기·타임라인 착지 전에는 AC 전부가 표면 없음 RED — 착지 순서를 리더가 정해요.
- 깨우기 전달 확인의 정의(무엇이 "확인"인지)가 7절 시험에서 처음 고정돼요 — 잘못 정하면 실패가 안 보여요.
- C-141 의 필드 이름이 늦으면 pr-state·bundle-run 재생 시험이 늦어져요.

## Result

(채우는 중 — 이 조각은 SPEC·PLAN 까지: 커밋·push 기록을 여기에 적어요.)
