---
id: SPEC-20261119-property-tests
title: "property tests for the invariants agents miss: HeadVer order, manifest round-trip, reservation state machine, path normalisation (C-100)"
status: draft
revision: 2
created: 2026-11-19
---

# property tests for the invariants agents miss (C-100)

## Intent

카드 C-100 (p1, Verify F): 에이전트가 자주 놓치는 불변식에 속성 시험을 걸어요 — HeadVer 정렬, source-manifest 쓰기의 멱등성, 카드 예약 상태 기계, 쓰기 범위 경로 검사. 시드를 찍는 자체 난수 생성기(새 의존성 없음)로 각 성질마다 1000 사례 이상을 30 초 안에 돌리고, 실패 시 시드와 사례 번호로 재생해요.

## Non-goals

- 새 크레이트 의존성(proptest·quickcheck) — 카드 본문이 리더 승인 없이는 금지.
- src/·Cargo.toml·skills/** 변경 — 시험 파일과 이 SPEC 만.
- 실제 원장·실제 HOME 대상 실행 — 임시 폴더만.
- 버그 수리: 이 SPEC 은 위반을 "잡는지"만 검증해요(위반이 발견되면 별도 카드).

## Requirements

- REQ-1: When a property of the suites runs, it shall execute at least 1000 generated cases within 30 seconds, seeded so `PROP_SEED` and `PROP_ONLY` replay one exact case, and name the seed and the case index in any failure message.
- REQ-2 (REQ-1): IF a property ran fewer than 1000 cases, crossed 30 seconds, or a failing case named neither its seed nor its case index, THEN the suite shall report a test failure (a canary property with a false assertion must fail).
- REQ-3: When two version strings are formatted from generated HeadVer components, `headver::compare` shall order them exactly as the components do — equal components equal; head, then yearweek, then build — with `headver::is_newer` agreeing and a batch sorted by `compare` holding no inversion.
- REQ-4 (REQ-3): IF `compare` answered `None` for a formatted version, ordered a pair against its components, disagreed with `is_newer`, or a sorted batch held an inversion, THEN the property shall fail naming the seed and the case.
- REQ-5: When the source manifest is rendered from a generated entry set, the bytes shall be identical for every permutation of the entries, unchanged by a second render of the same tree, and byte-identical to the merge driver's merged bytes for the same content.
- REQ-6 (REQ-5): IF two permutations of one entry set rendered differently, a re-render moved one byte, or the merge driver's bytes differed from the writer's, THEN the property shall fail naming the seed and the case.
- REQ-7: When random interleavings of take, renew, release, handoff, review, reopen, done and reclaim run against generated cards in a temporary ledger, the ledger shall hold no `review` card without a held reservation (a live lease or an ended, reclaimable one) and no pair of live reservations covering one path.
- REQ-8 (REQ-7): IF an interleaving left a `review` card without a held reservation or two live reservations covering one path, THEN the property shall fail naming the seed and the case.
- REQ-9: When a generated relative path — with `.` steps, `..` climbs, repeated separators, mixed case, and a symlinked temp tree around it — is checked against a write scope, the lexical scope check shall never allow a path whose resolution leaves the scope or answer differently when the tree exists.
- REQ-10 (REQ-9): IF the scope check allowed an escaping path or answered differently with the tree present, THEN the property shall fail naming the seed and the case.

## Acceptance

- AC-1 (REQ-3…REQ-6): headver 정렬 성질과 매니페스트 멱등·순열·드라이버 byte 동일 성질이 각각 1000+ 사례 통과.
  verify: `cargo test --locked --offline --test prop_headver`
  verify: `python3 -m unittest discover -s tests -p 'prop_manifest_test.py'`
- AC-2 (REQ-7, REQ-8): 무작위 연쇄 뒤 두 불변식(review 는 보유 예약·live 예약 경로 무공유)이 항상 성립.
  verify: `cargo test --locked --offline --test prop_reservation`
- AC-3 (REQ-1, REQ-2, REQ-9, REQ-10): 1000 사례·30 초·시드 재생·카나리 실패 전파와 경로 검사 무허가·순수성.
  verify: `cargo test --locked --offline --test prop_scope_paths`
  verify: `cargo test --locked --offline --test prop_headver the_generator_and_the_canary` (하네스 자체 시험)
- AC-4: 이웃 검사(대상 한정; 전량은 PR 의 CI run 이 인용).
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`(로컬은 대상 한정, 전량은 PR CI 인용)

## Scope

- write: `tests/common/prop.rs`, `tests/prop_headver.rs`, `tests/prop_reservation.rs`, `tests/prop_scope_paths.rs`, `tests/prop_manifest_test.py`, `.moai/specs/SPEC-20261119-property-tests/**`
- read-only: `src/headver.rs`, `src/ledger/reservation.rs`, `src/ledger/workflow.rs`, `src/ledger/mod.rs`, `tools/source_manifest.py`, `tools/merge_driver_manifest.py`, `tests/source_manifest_test.py`(Repo 하네스), `tests/common/home.rs`(TempHome)
- forbidden: 실제 HOME·실제 원장, 네트워크, src/·Cargo.toml·skills/** 쓰기, 새 의존성

## Plan

1. tests/common/prop.rs — SplitMix64 생성기·재생·30 초 예산·카나리(완료).
2. prop_headver.rs — 생성 성분→포맷→compare 성질(대칭·추이·is_newer).
3. prop_reservation.rs — TempHome 임시 원장, 카드 3·세션 3, 무작위 연쇄, 불변식 2종.
4. prop_scope_paths.rs — 임시 트리(심볼릭 링크 포함)·생성 경로·covers 무허가·순수성.
5. prop_manifest_test.py — 생성 항목→render 멱등·순열 불변·드라이버 byte 동일(Repo 하네스 재사용).
6. 대상 검사·PR(audit-tier 0, 시험 전용).

## Risks

- 1000 TempHome 사례의 시간 — 사전 풀로 SQLite 비용 축소, 30 초 초과 시 생성 축소가 아니라 풀링으로 해결.
- 연쇄의 합법 상태 공간 모델링 — 오류 응답은 합법적 거부로 무시하고 불변식만 측정.
- Python 드라이버 병합의 git 의존 — `Repo` 하네스가 이미 git 환경을 통제(GIT_CONFIG_NOSYSTEM 등).
- 속성 시험은 기존 코드 검증이라 "RED 먼저" 가 없을 수 있음 — 카나리(거짓 성질은 실패해야 한다)로 검출력을 대신 보여요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 자체 시드 생성기(tests/common/prop.rs, `#[path]` 포함; mod.rs 는 범위 밖) | 새 의존성 금지(카드 본문) | 리더가 크레이트 승인 |
| D-2 | 예약 상태기계는 in-process 라이브러리 호출(tests/ledger_reservation.rs setup 패턴) | 이진 실행 1000회는 30 초 예산 밖 | CLI 구동 축소판 |
| D-3 | 매니페스트 성질은 Python in-process(merged_manifest 직접 호출) | 작성자·드라이버가 Python 도구뿐; Repo 하네스의 저장소 수준 동일성은 기존 source_manifest_test 가 못박음 | Rust 재작성 카드 |
| D-4 | 하네스 자체 시험·카나리는 prop_headver.rs 에 한 곳(prop.rs 중복 컴파일 회피) | `#[path]` 포함이 매 suite 에 컴파일됨 | 세 suite 반복 |
| D-5 | 드라이버의 git 설정 탐침(refuse_settings_git_merges_differently)을 시험 파일 안에서 1회만 수행 | 두 git config 값은 시험 실행 중 불변; 성질은 병합 바이트를 측지, 도구는 미변경 | 매 호출 탐침(예산 초과) |
| D-6 | 예약 원장은 속성 전체에서 하나, 사례 끝에 lease 만료 | 사례당 원장 생성은 30 초 예산 초과(측정: 80.78 초) | 사례당 원장 |

## Result

- AC-1: met — `prop_headver` 3 통과(성분 순서·대칭·자기동일·is_newer 일치 1000사례, 배치 무역전 1000 묶음×8, 0.12 초); `prop_manifest_test` 2 통과(render 멱등·순열 불변·키 정렬, 드라이버 병합이 어느 항목 순서에서도 같은 바이트·canonical 과 동일, 2.43 초; git 설정 탐침은 시험 안 1회 캐시 — D-5, 도구 미변경).
- AC-2: met(1차 감사 FAIL B1 수리 후, -r2 가지) — 첫 구동부는 사례 종료 처리가 lease 만 깎아 과거 held 행이 take 를 계속 막아 공소했어요(reserve 5·review 1, 불변식 짝 비교 0회 — 감사 측정). 수리: fenced 연산이 현재 예약의 holder·generation 을 쓰고, 사례 종료를 시스템 전이(reopen→release, 끝난 lease 는 reclaim→release)로 마감. 최종 실측(기본 시드, 1000 사례): reserve 915·renew 135·release 147·review 126·reopen 15·handoff 98·reclaim 15·lease 만료 1122, 검사기 실행 review 판정 111회·live 짝 비교 115회 — 바닥(reserve≥500·renew≥50·review≥50·handoff≥20·reclaim≥5·review_checks≥50·live_pairs≥100)을 시험이 못박고, 6.4~11.6 초 안에 돌아요.
- AC-3: met — `prop_scope_paths` 1 통과(1.56 초): `.`·`..`·이중 구분자·대소문자·심볼릭 링크 트리에서 탈출 경로 불허가·트리 존재 무의존; 하네스 시험 `the_generator_and_the_canary`(결정성·이름 알파벳·거짓 성질의 실패 전파 PROP_SEED/PROP_ONLY); 최장 성질 19.99 초 < 30 초.
- AC-4: met — clippy `--all-targets -D warnings` 0, `cargo fmt --check`·rustfmt 0, repo_scans 25 OK, ac_command_lint 5 OK(Python 시험이 git 을 구동해 실행), Python prop 묶음 2 OK. 전량 cargo·CI 범위는 묶음 검증·CI 인용.
  개발 중 붉어진 것은 대상 코드가 아니라 시험 생성기 결함 4건(이중 구분자 변형이 한 글자 잘라먹기, 이름 `.`/`..` 충돌, 생성 패턴과 검사 패턴 불일치, Python 해시 재추첨) — 하네스가 스스로 잡아 고쳤고, 카나리가 검출력을 입증해요.
