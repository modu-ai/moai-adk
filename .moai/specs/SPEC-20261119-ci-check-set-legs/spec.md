---
id: SPEC-20261119-ci-check-set-legs
title: "ci 의 Check set 을 nextest 다리와 나머지 다리로 나눈다"
status: draft
revision: 1
created: 2026-11-19
---

## Intent

C-20 조각 2 실측(카드 노트 2833, 최근 10개 push run): ci 잡 시간의 89%(linux 911초 중 Check set 809초)·96%(macOS 2159초 중 2064초)가 Check set 한 단계이고, 그 안의 nextest(469초/1141초)와 python unittest(289초/881초) 둘이 94–97% 를 차지해요. 두 무거운 단계를 같은 OS 안에서 서로 다른 잡으로 병렬로 돌리면 잡 벽시계가 linux 약 15m→9m, macOS 약 36m→20m 으로 줄어요(리더 결정 노트 2837, 후보 1 채택 — 후보 2 파이썬 병렬화는 보류).

이 SPEC 은 설계만 적어요. 구현은 agent-2 가 ci.yml 을 바꾸는 head(C-57 계열)가 통합에 올라온 뒤 그 위에서 이뤄져요; 그때까지 워크플로 파일은 바꾸지 않아요(리더 노트 2837).

### 잡 이름을 읽는 곳 (이름을 바꿔도 되는 곳과 안 되는 곳)

- ruleset(통합 규칙): 필수 check 이름 `ci` 와 `moai/audit` 만 요구해요. `ci` 는 집계 잡의 GitHub check run 이름(잡 이름이 곧 check 이름)이고, 매트릭스 잡(`ci / …`)은 필수가 아니에요. → 집계 잡 이름을 `ci` 로 유지하면 ruleset 은 한 글자도 안 바뀌어요.
- audit-gate(SPEC-20261117-audit-gate, src/audit_gate.rs): PR head 의 필수 check `ci`(집계)를 읽어요. 다리 이름과 무관해요.
- ci-verdict(tools/ci_superseded.py): push run 의 잡 이름을 `TARGET_PREFIX = "ci / "` 로 잘라 target 을 얻고, macOS 대기-대체를 주석 문자열 `ci-push-<target>` 으로 판정해요. 다리 접미사가 붙으면 이 파서도 같은 커밋에서 고쳐요(REQ-8).
- src/gate.rs:1021: `name.starts_with("ci / ")` 로 tip 의 잡을 분류해요(필수 check 가 아니라 잡으로). 다리 이름도 `ci / ` 로 시작하면 이 분류는 그대로 유효해요.
- D8a concurrency(#2554): push 첫 시도의 macOS 잡이 그룹 `ci-push-<target>` 을 나눠 쓰는데, 같은 run 의 두 다리가 같은 그룹을 쓰면 하나가 대기로 밀려 병렬이 사라져요. → 그룹을 (target, leg) 단위로 `ci-push-<target>-<leg>` 으로 해요(REQ-7), ci_superseded.py 의 주석 판정도 같은 형태로(REQ-8).

## Non-goals

- 후보 2(python unittest 병렬화)·후보 3(ripgrep 캐시)·후보 4(checkout depth)는 이 SPEC 밖(리더: 후보 2 보류, 3·4 는 조각 3 뒤). 후보 5(push run 의 macOS 경로 필터)·6(러너 업)은 사용자 결정 목록.
- 필수 check 이름(`ci`·`moai/audit`), ruleset, 보호 설정, windows-build, targets 잡의 C-76 분류, docs-only 경로(#2527), candidate · audit-gate · moai-label-sync 워크플로 — 전부 무변경.
- local_check.sh 의 기본 동작(환경 변수 없이 전체 단계를 지금 순서대로) — 무변경. 로컬 레인의 사용법은 안 바뀌어요.

## Requirements

- REQ-1: While `LOCAL_CHECK_LEG` is unset or empty, tools/local_check.sh shall run exactly today's stage set in today's order (fetch, test, doctest, clippy, fmt, fmt-ops, unittest, diff-check). · failure: none — the default is pinned by tests.
- REQ-2: When `LOCAL_CHECK_LEG=nextest`, the script shall run exactly fetch, test(nextest), doctest, and no other stage. · failure: none — the stage set is pinned by AC-1.
- REQ-3: When `LOCAL_CHECK_LEG=rest`, the script shall run exactly fetch, clippy, fmt, fmt-ops, unittest, diff-check, and no other stage. · failure: none — the stage set is pinned by AC-1.
- REQ-4: When `LOCAL_CHECK_LEG` holds any other value, the script shall refuse with exit 2 and a stderr line before running any stage. · failure: none — the refusal is the guard itself; AC-1 pins it.
- REQ-5: The check matrix of `.github/workflows/ci.yml` shall be the targets output of the `targets` job (fromJSON) multiplied by the two legs `[nextest, rest]`, with job names `ci / <target> / <leg>`, the runner and timeout taken from the target entry, and `MOAI_BUILD_NUMBER` carried on both legs as today.
- REQ-6: The cargo-nextest install step shall run on the nextest leg only; both legs keep the same pinned actions, the docs-only scope step, the rust-cache step with `save-if` push-only, and `LOCAL_CHECK_LEG` passed to the script through `env:` (never `${{ }}` inside `run:`).
- REQ-7: The D8a concurrency groups on a first push attempt shall be per (macOS target, leg): `ci-push-<target>-<leg>`; a running job is never cancelled and every other job keeps a group of its own, exactly as today.
- REQ-8: tools/ci_superseded.py shall parse the leg-suffixed names (target = the first ` / `-separated segment after `ci / `) and detect the macOS wait-replacement by the annotation string `ci-push-<target>-<leg>`; the state is `superseded` only when every leg of every target passed or was a macOS wait replacement, and anything else stays `fail`.
- REQ-9: The aggregate job shall keep the name `ci`, its `needs` (`check`, `ci-verdict`) and its conditions unchanged, so the required check names (`ci`, `moai/audit`) and the integration ruleset are untouched.
- REQ-10: While `LOCAL_CHECK_MODE` is candidate, a set `LOCAL_CHECK_LEG` shall make the script refuse with exit 2 and a stderr line before running any stage. · failure: none — pinned by AC-1.

## Acceptance

- AC-1 (REQ-1…REQ-4, REQ-10): tests/local_check_test.py — 기본은 8단계 전부를 지금 순서로(기존 행 유지), `LOCAL_CHECK_LEG=nextest` 는 정확히 fetch·test·doctest, `=rest` 는 정확히 fetch·clippy·fmt·fmt-ops·unittest·diff-check, 모르는 값은 2 로 거절하고 pass 줄이 없음, 후보(candidate) 모드에서는 어떤 다리 값이든 2 로 거절(다리 하나만 돈 실행이 전체 통과 증거로 읽히지 않게 — 리더 수리 N1), keep-going · 첫 실패 규칙은 다리 안에서도 그대로.
  verify: `python3 -m unittest tests.local_check_test`
- AC-2 (REQ-5, REQ-6, REQ-9): tests/github_assets_test.py — 스크립트 호출은 매트릭스 다리가 함께 쓰는 정확히 한 곳(`LOCAL_CHECK_LEG` 는 step env 로 `matrix.leg` 를 받고 run 안에는 다리 표현식 없음; 구현 검토 결과 한 단계 정의가 두 다리 잡을 그대로 나눠 실행하므로 "각 다리 한 곳"의 못 박기가 이 모양으로 옮겨졌어요 — Result 참조), 매트릭스가 targets × [nextest, rest], 잡 이름 `ci / <target> / <leg>`, 집계 잡 이름은 `ci`, nextest 설치 step 은 nextest 다리만, 캐시 save-if 는 두 다리 모두 push 만, MOAI_BUILD_NUMBER 는 첫 시도만(기존 행), pinned actions · 권한 · scope step 행은 무변경.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-3 (REQ-7, REQ-8): tests/ci_superseded_test.py — 다리 접미사 이름 파싱, `ci-push-<target>-<leg>` 주석 판정, 전 다리 통과 + macOS 대체 조합은 superseded, 한 다리라도 다른 결과면 fail.
  verify: `python3 -m unittest tests.ci_superseded_test`
- AC-4 (REQ-9): gate 읽기 — `ci / linux-x86_64 / nextest` 같은 이름의 check run 은 tip 잡으로 분류되고 필수 check 로 읽히지 않으며, 필수 `ci` 는 집계만이에요(기존 gate 행 유지; 다리 이름 fixture 한 행은 연기 — Result 의 자체 확인 B1 기록: `tests/gate_collect.rs` 를 C-99 가 예약 중이라 이 가지에서 못 쓰고, `src/gate.rs` 분류가 `ci / ` 접두사라 다리 이름은 기존 행과 같이 읽혀요).
  verify: `cargo test --locked --offline --test gate_check_cli`; `cargo test --locked --offline --test gate_collect`
- AC-5 (REQ-1…REQ-9): 이웃 검사 — 저장소 전수와 SPEC 형식.
  verify: `python3 -m unittest tests.repo_scans_test tests.ci_scope_test`; `moai spec check SPEC-20261119-ci-check-set-legs`
- AC-6 (REQ-5, REQ-7): 관측 — 구현이 통합에 올라간 첫 push run 에서 타깃마다 두 다리 잡이 도는 것과 잡 벽시계를 기록해 조각 2 표(카드 노트 2833)와 나란히 적어요.
  verify: 결과 보고의 run id 로 `gh run view <id> --json jobs` 재읽기

## Scope

- write: `.github/workflows/ci.yml`, `tools/local_check.sh`, `tools/ci_superseded.py`, `tests/local_check_test.py`, `tests/github_assets_test.py`, `tests/ci_superseded_test.py`, `tests/ci_scope_test.py`(decide 스크립트 리터럴 고정 행 — AC-5 검증이 강제), (gate fixture 한 행은 연기: C-99 예약, AC-4 참조), `.moai/specs/SPEC-20261119-ci-check-set-legs/spec.md`
- read-only: `src/gate.rs`(접두사 분류), `tools/ci_scope.py`, `src/audit_gate.rs`
- forbidden: ruleset · 보호 설정 변경, 다른 workflow 파일, agent-2 의 ci.yml head 가 통합에 오르기 전 워크플로 파일 편집(리더 노트 2837), 실제 HOME · 키 · 원장

## Plan

1. 기준: agent-2 의 ci.yml head 가 통합에 올라온 tip 위에서 시작해요(그 전엔 구현을 시작하지 않아요).
2. RED 먼저: AC-1 의 다리 행(기본/nextest/rest/모르는 값)이 실패하는 것부터 — local_check.sh 에 다리 개념이 없다는 것이 첫 실패예요.
3. local_check.sh: `LOCAL_CHECK_LEG` 한 개(기본 전체), 지금의 step 호출 목록을 다리 표로만 갈라요; 요약 · keep-going · finish 줄은 그대로.
4. tools/ci_superseded.py: 이름 파싱(첫 칸 = target)과 주석 형식(다리 포함)을 REQ-8 대로 — ci.yml 과 같은 커밋에.
5. ci.yml: 매트릭스에 leg 차원, 잡 이름, nextest 설치 조건, Check set 의 env 에 LOCAL_CHECK_LEG, D8a 그룹에 leg, 집계와 verdict 잡은 이름 · 조건 그대로.
6. 시험 행(AC-2 · AC-3 · AC-4)과 이웃 검사(AC-5), push 뒤 AC-6 관측 기록.

## Risks

- 다리가 2배라 macOS 대기-대체(D8a)의 대기열이 다리별로 나뉘어요 — 각 다리가 짧아져 벽시계는 줄지만, 병합이 몰릴 때 다리별 대기가 늘어날 수 있어요. AC-6 관측에서 확인해요.
- ci.yml 과 ci_superseded.py 는 한 커밋에 함께 가야 해요 — 어긋나면 대체-취소 run 이 중립 대신 fail 로 읽혀 tip 이 붉게 보여요(REQ-8 + AC-3 이 못박음).
- 매트릭스 곱은 targets 잡의 decide 스크립트가 리터럴로 미리 계산해 내보내요(구현 모양; 정적 leg 차원을 check 잡에 두는 방식은 쓰지 않아요). `matrix.include` 가 fromJSON 출력을 그대로 읽는 문서화된 동작이지만, 첫 run 에서 AC-6 으로 실제 다리 넷이 도는지 확인해요.
- macOS 유일 시험 요약 표(e2e 등)가 두 잡의 요약으로 나뉘어 읽혀요 — 각 다리의 요약이 자기 결과만 보여 주는 것을 AC-2 로 못박아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 다리 고르기는 환경 변수 한 개 `LOCAL_CHECK_LEG`, 기본은 전체 | 리더 노트 2837("환경 변수 한 개, 기본은 지금과 같은 전체") | 플래그를 빼고 ci.yml 을 한 다리로 되돌림 |
| D2 | 값은 `nextest` · `rest` 둘뿐, 모르는 값은 exit 2 거절 | CI 가 문자값을 넣는 곳이라 조용한 전체 재실행보다 큰 소리 실패가 안전 | 모르는 값도 전체로 읽음 |
| D3 | doctest 는 nextest 다리에 | cargo test 계열로 같은 빌드를 쓰고 2–3초라 다리 균형에 영향 없음(github_assets_test 의 nextest·doctest 분리 행을 갱신) | rest 다리로 옮김 |
| D4 | D8a 그룹을 (target, leg) 단위로 | 같은 target 그룹을 두 다리가 나눠 쓰면 병렬이 사라짐(#2554 대기-교체) | target 단위 그룹으로 |
| D5 | 집계 잡 이름 `ci` 유지 — ruleset 무변경 | 리더 노트 2837; ruleset · audit-gate 는 집계 이름만 읽음 | 이름을 바꾸면 ruleset 은 사용자 결정 |
| D6 | ci.yml 과 ci_superseded.py 를 같은 커밋에 | 어긋나면 중립-취소가 fail 로 읽힘(Risks) | 따로 가면 감사에서 FAIL |
| D7 | 구현 다듬기 셋: 호출 한 곳 + 매트릭스 다리 · rest 다리 전용 빌드 단계 · smoke 는 두 다리 모두 | Result 1–3(설계 뒤 사실: REQ-5 매트릭스가 다리를 이미 실음, 파이썬 시험의 실바이너리 요구, 98351d78f9 의 시험 고정) | 각각 단계 복제·빌드 제거(완전 캐시 가정)·smoke 다리 조건으로 |
| D8 | 리더 수리 둘: 후보 모드 다리 거절(REQ-10) · rust-cache key 다리 접두 | 리더 tier 2 감사 뒤 지시(증거 오독 방지·캐시 항목 충돌 방지) | 각각 거절 제거·key 접두 제거 |

## Result

구현 완료(2026-11-19, agent-20, 카드 C-20 세대 2) — 커밋과 시험 증거는 카드 노트가 저번 행이에요. 구현이 설계에서 자세를 다듬은 세 곳:

1. **스크립트 호출 한 곳 + 매트릭스 다리**(AC-2 표현 갱신): REQ-5 가 매트릭스에 다리를 실으므로, Check set 단계는 한 정의로 두 다리 잡이 함께 쓰고 `LOCAL_CHECK_LEG` 는 step env 가 `matrix.leg` 로 받아요(run 안에는 다리 표현식 없음 — REQ-6 의 문자 그대로). 두 곳에 단계를 복제하는 모양보다 요약 분기가 단계 안 `if [ "$LOCAL_CHECK_LEG" ]` 하나로 들어가요.
2. **rest 다리의 바이너리 입력**(설계 때 없던 사실): 파이썬 시험 다수가 `target/debug/moai-cli` 를 실제로 실행해요(tests/self_audit_run.py 머리 주석 참조). 한 몸이던 Check set 에서는 nextest 가 같은 잡에서 먼저 링크해 줬지만, rest 다리는 clippy(check 전용)라 아무것도 링크하지 않아요. 그래서 rest 다리에만 Check set 앞 `Build moai-cli (unittest input)` 단계(`cargo build --locked --bin moai-cli`, 풀 범위만)를 뒀어요 — `--offline` 를 붙이지 않은 이유: 의존성 캐시가 차가운 PR(Cargo.lock 이 base 캐시키를 지난 경우)에서는 아직 아무것도 내려받지 않은 상태라, 필요한 것을 네트워크에서 가져와야 해요.
3. **smoke 단계는 두 다리 모두**: SPEC 작성 뒤 tip 에 올라온 98351d78f9(check 잡의 install·update smoke)은 `if:` 문구가 tests/ci_smoke_step_test.py(이 카드 범위 밖)에 고정돼 있어 다리 조건을 붙일 수 없어요. 영역당 수 초라 중복 비용은 미미하고, 집계 `ci` 요구는 그대로예요.

그 외는 설계 그대로: local_check.sh 는 `LOCAL_CHECK_LEG`(기본 전체·nextest·rest·모르는 값 exit 2)와 `on_leg` 목록으로 단계를 가려요(REQ-1…4); ci_superseded.py 는 첫 ` / ` 조각을 target 으로, 주석은 `ci-push-<target>-<leg>` 로 읽어요(REQ-8); D8a 그룹은 (target, leg) 단위, 집계 잡 이름·needs·조건 무변경(REQ-7·9). AC-6(첫 push run 관측)은 통합 뒤 리더 몫으로 남아요.

**자체 확인 기록(agent-20, 감사 에이전트 audit-c20-legs, 2026-11-19T20:47Z)** — 지정 검증 전부 exit 0. 발견 세 건과 처리: B1(AC-4 의 다리 이름 fixture 한 행이 gate 시험에 없고 Result 에도 미기록) → fixture 파일이 C-99 예약이라 이 가지에서는 쓸 수 없어 연기로 못 박고 이 문서에 기록(리더·C-99 몫; 접두사 분류라 행동은 같음); N1(tests/ci_scope_test.py 변경이 Scope 밖이고 미기록) → Scope 쓰기 목록에 추가; N2(Risks 의 동적×정적 곱 문장이 구현과 불일치) → 구현 모양(리터럴 곱)으로 문장 교체. REQ-1…9 와 Result 세 다듬기는 코드·시험과 일치 확인.

**리더 tier 2 감사와 수리(2026-11-19T20:53Z)** — tier 2 감사 PASS(head 7c1d7ca70b, 차단 0). READY 전 리더 수리 넷 중 이 문서가 이미 담은 것: N4(AC-4·Result 문구)·N5(Scope 의 ci_scope_test) — 위 자체 확인 수리와 같은 내용. 새로 더한 것: 리더 수리 N1 — 후보 모드에서 어떤 다리 값이든 exit 2 로 거절(REQ-10, 다리 하나만 돈 실행이 전체 통과 증거로 읽히지 않게; legs 는 CI in-place 병렬 장치); 리더 수리 N2 — rust-cache key 에 다리 접두(leg-nextest/leg-rest)를 넣어 한 push 의 두 다리 잡이 서로의 캐시 항목을 덮어쓰지 않게. N2 는 REQ-6 의 save-if 무변경과 별개의 key 값이에요.
