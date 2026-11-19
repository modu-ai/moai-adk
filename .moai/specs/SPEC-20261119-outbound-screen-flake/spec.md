---
id: SPEC-20261119-outbound-screen-flake
title: "outbound: 64 KiB key-prefix screening test pins cost by a same-process ratio, not wall-clock (C-111)"
status: draft
revision: 1
created: 2026-11-19
---

# outbound: 64 KiB key-prefix screening test pins cost by a same-process ratio, not wall-clock (C-111)

## Intent

카드 C-111 (p2, flaky): `a_64_kib_field_of_a_repeated_key_prefix_is_screened_quickly` (`src/decide/outbound.rs:2087`) 이 벽시계 2초 미만을 단언하는데, 2026-11-19 C-93 tier 2 감사에서 lib 전체 실행 중 유일한 실패로 관찰됐어요(부하 평균 ~38, 3.12 s; 단독 실행은 통과). 스캔은 선형이지만 기계가 바쁘면 선형 실행도 2초를 넘어요 — 시간 그 자체에 의존하는 단언이 결함이고, 규칙 C-99 에 따라 재실행 통과는 검증이 아니에요. 카드 기준 1 은 벽시계 대신 같은 프로세스에서 잰 기준선(선형 통과 1회)에 대한 비율 고정을 허용해요. 제품 코드와 카운팅 훅은 필요 없어요 — 테스트 모듈만 고쳐요.

## Non-goals

- `screen()` 과 그 아래 스캔 함수의 알고리즘 변경(이미 선형; 이차 재스캔은 2026-11-17 에 제거됨).
- 카운팅 훅(제품 코드 계측) — 기준선 비율로 충분해요(DEC-2).
- 릴리즈 빌드 성능 측정 — CI 와 감사가 도는 디버그 기준으로만 고정해요.
- 다른 flaky 시험(C-109 가 진행 중)과의 공용 헬퍼 추출.

## Requirements

- REQ-1: The test shall pin the screening cost of every 64 KiB key-prefix field by a ratio against one linear baseline pass over the same text measured in the same process, bounded by `COST_RATIO_LIMIT` alone with no wall-clock limit anywhere in the assertion.
- REQ-2 (REQ-1): IF the screening re-reads the whole text at each occurrence (the quadratic re-scan mutant of the 2026-11-17 defect), THEN the test shall fail on a ratio far above `COST_RATIO_LIMIT`.
- REQ-3: The whole lib target shall pass 20 times in a row while 32 busy processes run beside it.

## Acceptance

- AC-1 (REQ-1, REQ-2): 비율 단언 테스트가 통과하고(14개 키 모양 전부), 이차 재스캔 변이 1회를 `screen` 에 주입했을 때 같은 테스트가 실패해요(주입 → 실패 exit 101 → 원복; 결과는 Result 에 기록).
  verify: `cargo test --locked --offline --lib a_64_kib_field_of_a_repeated_key_prefix_is_screened_quickly`
  verify: `git diff --exit-code -- src/decide/outbound.rs`
- AC-2 (REQ-3): lib 전체 20연속 통과(스피너 32 병행). 명령과 통과 횟수를 Result 에 적어요.
  verify: `zsh target/load20-lib.sh`
- AC-3: 이웃 검사(대상 한정; lib 전량·fmt·clippy 는 아래와 Result, 전체는 묶음 PR 의 CI run 이 인용).
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`
  verify: `cargo test --locked --offline --test home_isolation_scan`
  verify: `moai trailer-check origin/v2026.11.0..HEAD`

## Scope

- write: `src/decide/outbound.rs`(tests 모듈만), `.moai/specs/SPEC-20261119-outbound-screen-flake/spec.md`
- read-only: `src/decide/normalise.rs`(스캔 단계 참고), `src/redact.rs`(참고)
- forbidden: 제품 코드·카운팅 훅 수정(변이는 임시 주입 뒤 원복), 실제 HOME·키·원장, 네트워크, force push·`reset --hard`·훅 해제

## Plan

1. RED: 현 테스트(벽시계 2초)를 스피너 32개 부하에서 여러 번 실행해 기록해요. 재현 못 하면 구조적 근거(벽시계 단언)와 C-93 관찰을 Result 에 적어요.
2. GREEN: 테스트를 비율 단언으로 고쳐요 — 텍스트마다 (a) in-crate 선형 통과 1회 시간을 재고 (b) `screen` 시간이 그 `COST_RATIO_LIMIT` 배 미만임을 단언해요. `COST_RATIO_LIMIT` 은 실측 비율의 최소 10배 여유로 정하고 근거 수치를 주석과 Result 에 남겨요.
3. 변이 1회: `screen` 에 이차 재스캔(각 위치에서 전체 재판정)을 주입 → 같은 테스트 실패 확인 → 원복.
4. AC-2 부하 20연속 실행 후 Result 보충, 이웃 검사, push, 자체 감사(tier 0).

## Risks

- 비율 노이즈: 부하가 두 측정을 함께 느리게 해 비율은 크게 안정이지만 스케줄러 양자 경계에서 수 % 흔들려요 — LIMIT 여유로 흡수해요.
- 디버그 빌드의 std 비대칭: std 의 최적화된 검색을 기준선으로 쓰면 비율이 수천이 되어 판별 폭이 닳아요 — 기준선을 크레이트 내 반복자 루프로 만들어 완화해요(DEC-1).
- 같은 코드를 기준선으로 삼으면 변이도 기준선을 늘려(4×) 변이를 놓쳐요 — 기준선은 변이와 무관한 독립 통과여야 해요(DEC-1).
- 원 flake 의 국소 재현 실패 가능: 부하 32 재현은 감사 환경(부하 38, 병렬 cargo)과 다를 수 있어요 — RED 가 재현 못 하면 Result 에 그대로 적고, 기준 2 의 20연속이 실질 검증이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| DEC-1 | 기준선은 변이와 무관한 in-crate 선형 통과(텍스트 바이트 1회 순회)로, `screen` 자체의 절반 길이 재측정이 아니다 | 같은 코드 기준선은 변이에도 함께 커져 이차를 못 잡아요; 카드 기준 1 이 같은 프로세스 기준선을 허용 | 카운팅 훅 도입 카드 |
| DEC-2 | 카운팅 훅 없음 — 테스트 모듈만 고쳐요 | 카드 범위 "only if needed"; diff 최소로 tier 0 유지 | 훅 추가 후속 카드 |
| DEC-3 | `COST_RATIO_LIMIT` 은 실측 뒤 고정 상수로 박아요(측정 코드는 커밋하지 않음) | 상수가 계측보다 검증 가능; 수치 근거는 주석·Result | 실측이 바뀌면 상수 갱신 |
| DEC-4 | `moai spec take` 미사용(설치 바이너리에 없음, `moai spec --help` 명령 목록 확인) — `git worktree add -b spec/… origin/v2026.11.0` + `card take --branch` 로 같은 결과 | 훅이 요구하는 것은 가지와 SPEC 의 존재; nested-manifests SPEC D-4 선례 | 소행으로 되돌리기 |

## Result

- (실행 뒤 채움)
