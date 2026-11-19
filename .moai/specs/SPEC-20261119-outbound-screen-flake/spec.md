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

측정 환경: 이 레인 worktree(`/Users/goos/.moai/worktrees/moai-adk-111`), 디버그 빌드, 이 커밋의 부모 `fb3521bdd0c9d67793ae05b5c89a4732af92ed13` 코드 상태(이 커밋은 문서만). 각 항목에 명령과 그 실행에서 관측한 수치를 적어요.

- AC-1 RED(옛 벽시계 단언의 부하 실패 재현): 스피너 32개 병행 부하에서 옛 테스트 `a_64_kib_field_of_a_repeated_key_prefix_is_screened_quickly` 가 6회 실행 6회 실패(본체 2.39–6.31 s, 2026-11-19T06:26Z, load 평균 ~37). C-93 tier 2 감사의 관찰(부하 ~38, 3.12 s, lib 전량 중 유일 실패)과 같은 결함 — 벽시계 단언은 부하에서 깨져요.
- AC-1 GREEN(수리, 커밋 `6350ebb75e`): 14개 키 모양 각각에 대해 (a) 같은 프로세스에서 텍스트 바이트를 한 번 도는 in-crate 선형 기준선(`black_box` 로 접기 방지)을 재고 (b) `screen` 비용이 그 기준선의 `COST_RATIO_LIMIT` 배 미만임을 단언. 같은 부하에서 통과. 벽시계 단언과 2초 한도는 테스트에서 완전히 제거됨(REQ-1).
- AC-1 변이(REQ-2): `screen` 에 이차 재스캔(각 위치에서 전체 재판정)을 주입 → 같은 테스트 실패, exit 101 `"ratio 21,316 at limit 15,000"`(KILL). `COST_RATIO_LIMIT=20,000` 시험에서도 24,489 로 KILL — 두 한도 모두 변이를 잡아요. 원복 `git diff --exit-code -- src/decide/outbound.rs` → CLEAN(제품 코드 무변경).
- `COST_RATIO_LIMIT` 산정(커밋 `fb3521bdd0`): 실측(디버그, heavy-slot) — 정상 선형 비율 범위 6–1,104, 이차 재스캔 변이 비율은 변이 실행 두 번에서 21,316(한도 15,000)·24,489(한도 20,000)로 관측. 기준선과 변이의 바이트당 비용이 같아 변이 비율의 구조적 상한은 텍스트 절반 길이(≈32,768) 근처예요. 15,000 으로 고정: 정상 최대(1,104)의 13.6배 위, 변이 실측 24,489 의 1.6배 아래(21,316 기준 1.4배), 상한의 약 2.2배 아래 — 선형 쪽 여유 10배 이상, 변이 쪽 바닥은 1.4배 이상(Risks 의 노이즈 흡수).
- AC-2(lib 20연속 부하): `zsh target/load20-lib.sh` — 스피너 32개 병행 + `cargo test --locked --offline --lib` 20회 → **20/20 통과**(매 회 1453 passed; 회당 155–207 s; 2026-11-19T06:59–08:02Z; 부하 30–78).
- AC-3 이웃 검사(위 코드 상태에서 2026-11-19T09:41–09:47Z 재실행, 전부 exit 0): `cargo fmt --check` / `cargo clippy --locked --offline --all-targets -- -D warnings`(25.2 s) / `cargo test --locked --offline --test home_isolation_scan`(18 passed, 0 failed) / `python3 -m unittest tests.repo_scans_test tests.source_manifest_test`(68 tests OK, 79.6 s) / `./target/debug/moai-cli spec check SPEC-20261119-outbound-screen-flake --project .`(errors 0, warnings 0) / `moai trailer-check origin/v2026.11.0..HEAD`(3 commits, failing 0 — 이 Result 커밋 뒤 4 commits 로 재실행해 failing 0 확인).
- 남은 것: lib 전량·전체 cargo test 는 이 묶음의 PR CI run 이 인용해요(AC-3 전언; tier 0).
