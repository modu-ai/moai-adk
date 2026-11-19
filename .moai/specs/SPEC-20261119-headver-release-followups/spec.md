---
id: SPEC-20261119-headver-release-followups
title: "Release follow-ups: malformed tags are refused, the build cap is explicit, the order test has no silent green"
status: draft
revision: 1
created: 2026-11-19
---

# Release follow-ups: malformed tags are refused, the build cap is explicit, the order test has no silent green

## Intent

C-93 의 tier 2 감사(카드 노트 2759, PASS)의 비차단 후속 세 가지: (N1) `release.yml` 의 "번호는 오르기만" 검사가 `moai-adk-v*` 태그의 모양을 엄격히 보지 않아서 — head 가 10자리를 넘거나 앞자리 0, build 가 2^63 을 넘으면 bash 산술이 감겨 검사를 조용히 건너뛰고, 잘못된 모양의 태그는 아예 무시돼요 — 협력자가 밀어 넣은 태그도 거절하지 않았어요. (N2) SPEC REQ-2 가 실제 강제 범위보다 강하게 말했어요(첫 시도 검사는 plan 잡에만 있고, build 잡만 다시 실행하면 같은 커밋·같은 번호라 무해). (N3) `tests/update_headver.rs` 의 같은 주 비교 시험이 번호 없는 빌드에서 조용히 통과했어요.

## Non-goals

- build 잡에 첫 시도 검사를 더하기(N2 의 다른 갈래 — 이 조각은 REQ-2 문구로 닫음).
- 태그 ruleset·보호 설정, 실제 태그·릴리스 게시(카드 AC-4 가 가드한 사용자 선택 영역).
- 다른 tag 계열(`v3.1.3`·`go-v1.2.3`·`moai-cli-v*`)의 무시 동작 — 그대로 유지해요.

## Requirements

- REQ-1: While a tag under `refs/tags/` starts with `moai-adk-v`, the release plan shall refuse the run unless the tag matches `moai-adk-v{head}.{yearweek}.{build}` exactly — the head `0` or up to 10 digits without a leading zero, the yearweek exactly 4 digits, the build from 1 without a leading zero.
- REQ-2 (REQ-1): IF the matched build is over 4294967295, THEN the release plan shall refuse the run, so bash arithmetic never wraps the comparison.
- REQ-3: While `MOAI_BUILD_NUMBER` and `MOAI_LOCAL_BUILD` give a build no usable number (or the number is 1), the update-order test shall assert the real order — the installed release of the same week reads newer, so the plan reports `going_back` — instead of returning early.
- REQ-4: Where SPEC-20261119-headver-release-workflow's REQ-2 says more than the enforced scope, the follow-up branch shall reword it to the scope the plan enforces.

## Acceptance

- AC-1 (REQ-1): 잘못된 `moai-adk-v*` 태그 열 모양(head 앞자리 0·head 10자리 초과·head 11자리인데 앞자리 0 없음·주차 5자리·build 앞자리 0·build 없음·접미사·네 번째 점 부분 등)이 각각 "does not match moai-adk-v" 로 거부되고, 다른 계열은 여전히 막지 않아요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.release_workflow_test -k tag`
- AC-2 (REQ-2): build 4294967296 은 상한 거절("over 4294967295"), 4294967295 는 "not above" 거절로 이름이 불려요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.release_workflow_test -k cap`
- AC-3 (REQ-3): 번호 없는 빌드에서 같은 주의 릴리스가 `going_back` 으로 읽혀요(조용한 통과 제거).
  verify: `cargo test --locked --offline --test update_headver`
- AC-4 (REQ-4): SPEC-20261119-headver-release-workflow 의 REQ-2 가 실제 범위를 말하고 `moai spec check` 가 0 오류예요.
  verify: `moai spec check SPEC-20261119-headver-release-workflow`

## Scope

- write: `.github/workflows/release.yml`, `tests/release_workflow_test.py`, `tests/update_headver.rs`, `.moai/specs/SPEC-20261119-headver-release-followups/spec.md`, `.moai/specs/SPEC-20261119-headver-release-workflow/spec.md`(REQ-2 문구만)
- read-only: `tests/github_assets_test.py`, `src/headver.rs`, `tools/repo_scans.json`
- forbidden: 실제 태그·릴리스 게시, ruleset·보호 설정, 실제 HOME, 네트워크(시험은 fixture 저장소)

## Plan

1. RED 먼저: release_workflow_test 의 잘못된 태그 거부·상한 거부 시험과 update_headver 의 번호 없는 갈래 단언 — 구현 전 실패.
2. release.yml 의 태그 순환: 모양 검사(엄격 정규식)+32비트 상한+거절 문구.
3. update_headver.rs: build < 2 갈래가 going_back 을 단언.
4. 이웃: github_assets_test·ci_scope_test·lib 전체·fmt·clippy·설치본 trailer-check → push → 자기 감사(다른 문맥) + tier 2 요청 → LANE-READY-INTEGRATION.

## Risks

- 태그 모양을 엄격히 하면 과거에 밀어 넣은 잘못된 태그가 릴리스를 막아요 — 거절 문구가 태그와 이유를 이름 지어 사람이 지우게 해요.
- `going_back` 단언은 시간 의존이 아닌 버전 문자열 비교라 재현성이 높아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 모양 정규식: head `(0|[1-9][0-9]{0,9})`, 주 `[0-9]{4}`, build `[1-9][0-9]{0,9}` + 4294967295 상한 | N1 의 두 구멍(2^63 감김·무시)을 한 자리에서; 워크플로 자체 태그가 늘 이 모양 | 상한 완화 |
| D-2 | N2 는 REQ-2 문구로 닫음(build 잡 검사 추가가 아님) | 같은 커밋·번호라 무해하다는 감사 판단 — 행동 추가보다 문서 정렬이 안전 | build 잡 검사 추가 |
| D-3 | N3 는 build < 2 갈래에 실단언(going_back) | 조용한 통과 대신 분기마다 참인 주장 | 조기 귀환 복원 |

## Result

- AC-1: met — RED 소급: 구 워크플로에서 새 시험 둘 실패(exit 1); GREEN: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.release_workflow_test` 29 tests OK(2026-11-19, 이 worktree, C-93 head 5ca027dd4c 기반) — 여덟 잘못된 모양이 각각 "does not match moai-adk-v" 로 거부되고 `v3.1.3` 등 다른 계열은 계속 막지 않아요.
- AC-2: met — 같은 실행의 `test_a_build_number_over_the_32_bit_cap_is_refused`: 4294967296 은 "over 4294967295" 거부, 4294967295 는 "not above" 거절로 이름이 불려요.
- AC-3: met — `cargo test --locked --offline --test update_headver` 5 passed(로컬 build 0 → 번호 없는 갈래가 preflight blocked·`moai update --allow-older` 이유를 단언 — 조용한 통과 제거).
- AC-4: met — `moai spec check` 두 SPEC 모두 errors 0, warnings 0(REQ-2 문구는 실제 강제 범위: plan 이 번호를 매기는 유일 잡, build-only 재실행은 같은 커밋·번호).
- 이웃과 변이: github_assets_test 39 OK, lib 전체 1451 passed, `cargo fmt --check` 0, `cargo clippy --offline --locked --all-targets -- -D warnings` 0; 변이 M1(엄격 모양을 구 정규식으로 되돌림) → 잘못된 태그 시험 FAILED exit 1(KILLED), 복원 `diff -q` 클린.
- tier 2 감사 수리(AUDIT C-93 head b168766f10, FAIL B1, 감사 판정은 카드 노트 3230, 이 수리를 알린 카드 노트는 3279): 변이 M2(태그 정규식이 네 번째 부분을 허용)와 M10(head 자릿수 제한 제거)이 29 tests OK 로 살아남았어요. `test_a_malformed_moai_adk_tag_is_refused` 에 입력 둘을 더했어요 — 네 부분 태그 `moai-adk-v4.2648.7.1`, 앞자리 0 이 없는 11자리 head `moai-adk-v12345678901.2648.7`(기존 긴 head 둘은 앞자리 0 이라 길이 제한이 고정되지 않았어요). 증거(임시 복사본에서, 저장소 파일은 불변): 새 시험은 GREEN 29 OK; M2 를 넣으면 새 시험 FAILED(`bad='moai-adk-v4.2648.7.1'`)인데 옛 시험은 29 OK; M10(`{0,9}` → `*`)을 넣으면 새 시험 FAILED(`bad='moai-adk-v12345678901.2648.7'`)인데 옛 시험은 29 OK. 두 변이 KILLED. 워크플로 동작은 바뀌지 않았어요(시험 입력만 추가).
