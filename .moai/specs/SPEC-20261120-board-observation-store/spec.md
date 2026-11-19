---
id: SPEC-20261120-board-observation-store
title: "BOARD 02 — observation store and collector after leader-round"
status: draft
revision: 1
created: 2026-11-20
---

# 관측 저장과 수집 — leader-round 뒤의 한 걸음 (BOARD 02)

## Intent

기준 문서 `~/.moai/reports/moai-adk/collab-board-final-design-2026-11-20.md`(3판) 5.3·6.1절, 승인 결정 4(위치·보관)·3(카드
번호 비공개). 원장(영구)과 관측(30일)을 갈라: 프로젝트 폴더 안의 **별도 관측 저장 파일**에 관측을 쌓고,
MODE-LEADER 7단계 `supervise` 에서 `leader-round` **바로 뒤의 한 걸음**으로 수집해요. 원장 스키마 v1 은 그대로(표·색인·
journal_mode 무변경), 읽기 명령은 아무것도 쓰지 않고 수집을 깨우지 않아요. BOARD 01 의 읽기 방식(짧은 대기로 읽기 연결,
읽기 트랜잭션 안에서 정해진 개수, 연결을 닫은 뒤 화면)과 커서 계약(불투명 문자열, 받은 그대로 돌려줌, `--after` 는 기다리지
않음)에 맞춰요.

## Non-goals

- 원장 v2(표·색인 추가) — 승인 범위 밖(결정 6).
- 타임라인 구현·화면 — BOARD 01(읽기 전용) 영역; 이 저장이 읽기 API 를 제공할 뿐.
- 깨우기 전달 자체와 그 화면 표시 — BOARD 05(운용) 영역; `wake@1` 행을 쓰는 것만 이 카드.
- `run-start@1`·`run-exit@1`(런처 시작·종료 기록) — 생존은 잠금 조회가 알려 줘 뒤로 미룸.
- 로그 원문·비밀값·대화 원문 저장 — 참조와 해시만.
- 실제 원장·실제 HOME·호스트 훅 설치 — 모두 금지.

## Requirements

- REQ-1: When the observation store is created, it shall live inside the project's own folder as one file beside the ledger, with the ledger folder's rules applied again: owner-only permissions (0700) and a refusal when a link stands anywhere on the path · failure: none — deleting the whole store and collecting again must leave the ledger file unchanged.
- REQ-2: When an observation row is appended, the row shall carry the store's own sequence, the source, the source's event id, the scope, the observed time, the kind, the content hash and the content, with the pair of source and source event id unique · failure: none — feeding the same observation file twice adds no rows, and the same id with different content is kept apart as a conflict.
- REQ-3: When a kind is first fixed, it shall be one of lane-observation@1, pr-state@1, local-check@1, bundle-run@1, gap@1, wake@1, with its version never changing afterwards · failure: none — an unknown kind is refused, not guessed into a row.
- REQ-4: When observations age past the retention of 30 days, they shall be deleted from the store while the ledger stays permanent · failure: none — pruning never writes the ledger file.
- REQ-5: When the collector runs, it shall hold the project's single collection lock and end as skipped without waiting when the lock is already held · failure: none — the overlapping second collection is a skip, never an error and never a queue.
- REQ-6: When MODE-LEADER's supervise step runs leader-round, the collector shall run as the one step right after it and wake on nothing else · failure: none — read commands never wake it, no host hook is installed, and the observation file it reads keeps schema v1 with no added fields.
- REQ-7: When the collector stores what it gathered, the cursor advance and the observation appends shall be one transaction inside the store only, with host calls (Orca, GitHub) made before that transaction and a collection failure leaving the ledger content and every card state unchanged · failure: none — the timeline then shows the last observation time and unknown.
- REQ-8: When the cursor is deleted and collection runs again, the result shall be the same as with it · failure: none — the cursor only reduces work; an empty range becomes a gap@1 row from a bounded re-snapshot, never a guess, and the four PR values (number, state, review verdict, head SHA) are re-observed each round.

## Acceptance

- AC-1 (REQ-2, REQ-8): 같은 관측 파일을 두 번 넣어도 행이 늘지 않고, 저장을 지우고 다시 모으면 결과가 같아요.
  verify: `cargo test --locked --offline --test board_observation_store_cli`(시험 파일 이름은 후보)
- AC-2 (REQ-5, REQ-7): 수집 전후로 원장 파일의 내용이 바이트까지 같고, 겹치는 수집 잠금은 둘째를 건너뜀으로 끝내요.
  verify: `cargo test --locked --offline --test board_observation_store_cli`(수집 잠금·원장 무변경 시험)
- AC-3 (REQ-1, REQ-3, REQ-4, REQ-6): 위치·권한(0700)·링크 거절, 종류 여섯 고정, 30일 보관, 커서 삭제 후 재실행 동일, 관측 파일 스키마 v1 무변경.
  verify: `cargo test --locked --offline --test board_observation_store_cli`(저장 규칙 시험) + 불변식 8 질의(관측 저장 삭제·재수집 뒤 원장 무변경)

## Scope

- write(구현 조각, 리더가 PLAN 을 읽은 뒤): 관측 저장 모듈(이름 후보 `src/board/observation.rs`), 수집기와 `leader-round` 뒤 걸음(호출부), 시험 파일(이름 후보 `tests/board_observation_store_cli.rs`), 이 SPEC, 운용 문서의 저장·보관 문단
- read-only: `src/ledger/folders.rs`(폴더 규칙), `src/ledger/open.rs`(짧은 대기 읽기), `src/leader_round.rs`, BOARD 01 의 타임라인 코드(착지 뒤 계약 정렬)
- forbidden: 원장 스키마·표·색인·journal_mode 변경, 읽기 명령의 어떤 쓰기, 호스트 훅 설치, 실제 원장·실제 HOME, 카드 번호를 SPEC id·가지·커밋·PR 에 사용, 타인 회수 3회 한도·4시간 임대 변경

## Plan

1. 조각 1(이번, 코드 없음): SPEC + PLAN 노트 — 리더가 읽고 조각 2 를 정해요.
2. 조각 2(저장): 파일·권한·링크 거절, 행 모양·유일성·멱등, 종류 여섯, 30일 보관, 커서+추가 한 트랜잭션.
3. 조각 3(수집기): leader-round 뒤 걸음(예약 겹침 card check 먼저), PR 네 값, 수집 잠금, gap@1 제한된 재스냅샷, 실패 격리, 깨우기 키(원장 이벤트 id, 중복 답 하나).
4. 조각 4(시험 마감): 기준 1..3 시험, 불변식 8 질의, 이웃 검사.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 저장은 원장 폴더 규칙(src/ledger/folders.rs)을 그대로 다시 적용 | 설계 5.3 [현재] 항 | 별도 규칙 만들기 |
| D-2 | 커서 전진과 관측 추가는 저장 안의 한 트랜잭션, 원장과는 묶지 않음 | 설계 5.3 | 원장 트랜잭션에 넣기(스키마 변경 필요 — 금지) |
| D-3 | 커서는 BOARD 01 계약 그대로 불투명 문자열, 계산·정렬하지 않음 | 설계 5.1 커서 계약 | 해석하는 커서 |
| D-4 | 수집 잠금은 프로젝트마다 하나, 잡혀 있으면 기다리지 않고 건너뜀 | 설계 5.3 수집기 표 | 대기열 |
| D-5 | 종류 여섯만 먼저 고정하고 버전은 영구 | 설계 5.3 | 종류 추가는 되돌림(새 종류 추가는 가능, 버전 변경은 불가) |
| D-6 | 호스트 호출(Orca·GitHub)은 저장 트랜잭션 밖에서 먼저 읽기 | 설계 5.3 수집기 표 | 트랜잭션 안 호출(금지) |

## Risks

- `leader-round` 호출부(src/leader_round.rs 또는 그 상단)의 예약 겹침 — 착수 전 card check 로 확인하고 겹치면 리더에게 조정 요청.
- BOARD 01(C-138) 착지 전에는 타임라인 합치기 읽기가 없어요 — 저장 읽기 API 계약을 먼저 고정하고 01 이 맞춰요(또는 반대).
- 30일 보관 삭제가 수집 트랜잭션과 겹치면 길어져요 — 삭제는 잠금 안에서 짧게, 대상이 많으면 나눠서(구현 시험에서 재요).

## Result

조각 1(SPEC·PLAN): 커밋 `1dfc1e7c14`(SPEC — push 뒤 git ls-remote 재독 동일, spec check 0 errors 0 warnings), PLAN 노트 plan-board02-20261119T213427Z. 조각 2(저장): 커밋 `c0d7d0bb25`(모듈+시험), base `035f62dd1d`.

- RED: 모듈이 없어 컴파일 실패 — `couldn't read tests/../src/board/observation.rs`, 종료 101(기록).
- GREEN: `cargo test --locked --offline --test board_observation_store_cli` 9 passed 0 failed(위험 경계 시험 포함: 링크 거절·실패 묶음의 커서 무변·30일 prune 의 원장 무변·읽기 무변).
- 변이 2/2 KILLED: M1 종류 검증 끄기 → `the_six_kinds_are_fixed_and_an_unknown_kind_is_refused`·`a_failed_batch_leaves_the_cursor_where_it_was` 실패; M2 멱등 검사 우회 → `feeding_the_same_observation_file_twice_adds_no_rows`·`the_same_id_with_different_content_is_kept_apart_as_a_conflict`·`deleting_the_cursor_and_committing_again_gives_the_same_result` 3개 실패. 각각 git checkout 으로 복원해 작업 폴더 깨끗.
- 검사: cargo fmt --check 0, clippy --offline --locked --all-targets -- -D warnings 0(미사용 필드·상수 format! 3건 수리 뒤), home_isolation_scan 18 통과, spec check 0/0, spec_cli 18 통과, python 네 모듈(repo_scans·source_manifest·local_check·skill_references) 0, git diff --check base...HEAD 0.
- lib 선언 한 줄(pub mod board;)은 src/lib.rs 의 예약이 풀리면 다음 조각에서 — 이 조각의 시험은 경로 포함으로 모듈을 직접 돌렸다(결정 노트 decision-lib-decl-deferred). wake@1 은 이 카드에서 행 저장만.
