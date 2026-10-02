# progress.md — SPEC-FACTORY-MANAGED-HARDEN-001

> 단계별 진행 기록. `§E.1` 만 plan 단계(manager-spec)가 채우고, `§E.2`·`§E.3` 은 run 단계(manager-develop), `§E.4` 는 sync 단계(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-after-audit-iter2 (재감사 대기, 마지막 회차)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 16) · plan.md · acceptance.md (AC 16) · design.md (D-1..D-3, 순서 정본 D-3.1) · progress.md
tier: M
measured_tree: 7109e0900
open_clarifications: 0

### plan-audit iteration 1 — FAIL 0.75 (audited_sha 95dfd85c8)

판정서: `.moai/reports/t1409/plan-audit-iter1.md`(로컬 사본, gitignore 대상). 차단 D1–D5, 주요 D6–D7, 경미 D8–D10. 개정(0.2.0, 커밋 `820eff47f`)이 그 처분이었다.

| 결함 | 처분 | 개정 내용 |
|---|---|---|
| D1 RED 산출물 경로 불착지 | fixed | 산출물을 추적되는 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 로 이동(측정: `git check-ignore -v` 대 `.moai/reports/…` 는 `.gitignore:235` exit 0, SPEC 경로는 출력 없음 exit 1; 감사 캐시 해시 목록 밖). `.moai/reports/t1409/` 는 로컬 사본뿐 |
| D2 `Close`/`Start` 경합 | fixed(2차에서 순서 재정립) | design D-3 에 수명주기 규칙, REQ-MH-012 개정, AC-MH-016 신설, spec §F 경계 — 2차 감사가 N1·N2 로 이어짐 |
| D3 표 이스케이프 파이프로 시험 0개 선택 | fixed | 파이프 문자를 담은 명령은 표 밖 fenced 블록, 선택 수 66 대 0 재측정 |
| D4 시험 기반 AC의 RED-now 부재 | fixed (옵션 b) | "plan 시점 미채택, M1 이 채택", 재현 시험 8개 |
| D5 MP-6 D8 문면 | fixed | spec.md C.3·§D 의 cross-platform exemption(EXCL-syscall) |
| D6 부모 REQ-MS-012 문면과 운영 문서 | fixed | 명시적 예외 선언, AC-MH-012 가 운영 문서 문장 정정 확인 |
| D7 elicitation 전제 오귀속·조용한 루프 | fixed(2차에서 귀속 순서 재정립) | 전제 정정, stderr 한 줄, 턴 단위 실패 처리 — 2차 감사가 N3 으로 이어짐 |
| D8·D9·D10 | fixed / disclosed | 고정 앵커, "after the priming turn", 호출 수·시험 수 정정, 쓰기 한계·두 번째 시그널·후손 프로세스 공시 — 쓰기 한계 문장은 2차에서 N7 로 정정 |

### plan-audit iteration 2 — FAIL 0.81 (audited_sha 820eff47f)

판정서: `.moai/reports/t1409/plan-audit-iter2.md`(로컬 사본). 점수는 Tier M 통과선 0.80 을 넘었으나 blocking N1–N4 가 FAIL 을 강제했다. 1차 결함 D1·D3·D5·D6·D8–D10 은 해소 확인, D4 는 PASS-with-debt, D2·D7 은 N1–N4 로 이어짐.

**리더 결정(2차 후)**: 정확히 **한 번 더**(3차, 마지막) 감사를 승인했다. 3차에서 FAIL 이면 연장하지 않는다. 델타 범위는 N1–N4 + 경미(N5–N7) + 회귀. 진입 조건: 3차 판정이 PASS 계열이고 blocking 0 이면 plan→run Kickoff 를 autonomous 형태(독립 감사 교차 증거 + 판정 기록)로 진행한다(auto-semantics §9.1; keep-set 사례는 운영자 응답을 유지).

개정(0.3.0)의 처분(리더 지시 — 순서 문제는 문장 땜질이 아니라 한 번에 못 박는다):

| 결함 | 처분 | 개정 내용 |
|---|---|---|
| N1 `started` 게이트 모순과 `MkdirTemp`→`cmd.Start` 누수 창 | fixed (옵션 b) | design D-3.1 의 **한 순서표 O1–O20** 이 정본. 게시(published) ⇔ 지금의 `started`(`cmd.Start` 성공 직후). `MkdirTemp`–토큰 쓰기–엔드포인트–`cmd.Start`–게시가 **L 안 한 임계구역**이라 `Close` 가 중간에 끼어들 수 없고, 토큰 디렉터리는 게시 때까지 `Start` 소유(F8 불변). 느린 단계(준비 폴링·다이얼·핸드셰이크)는 L 밖, 연결 기록(O12)에서 닫힘 재확인. 두 소유자의 `Close` 도착 시점별 결과표 포함. 모순 문장 제거, spec §F 의 F8 경계를 "바뀌는 것/안 바뀌는 것/겹치는 곳"으로 재작성 |
| N2 AC-MH-016 이 변이를 못 가름 | fixed | 훅 하나(`managedStepHook`)를 단계 지점마다 둠(`start-before-lock`, `start-token-written`, `start-child-spawned`, `start-published`, `start-dialed`(다이얼 후·기록 전, 연결을 받음), `start-conn-recorded`, `close-snapshot`, `teardown`). AC-MH-016 하위 케이스 10개와 케이스별 기대(Start·Close 결과, 사후 조건 5종, 정리 1회), `during_handshake` 2초 시험 안 상한(Windows 성립), 경합 반복(codex 40, stream 200)과 그 근거, 결과를 훅 이벤트 순서로 결정하는 오라클, 변이 m1–m23 과 지목 하위 케이스(acceptance §2.4), DoD 3 갱신. plan M1 RED 이유 표와 "옳은 이유의 RED"(시험 자신의 시간 상한 단언은 옳은 이유, 하네스 타임아웃·컴파일 실패는 wrong-reason) 규정을 한 방향으로 통일 |
| N3 elicitation 계수기 순서 | fixed | design D-1 결정 3: 판정을 소비자가 아니라 **읽기 고루틴이 프레임 도착 순서로** 정한다(`armTurn` 이 `turn/start` 쓰기 전에 창을 열고 초기화, `turnId` 우선·없으면 열린 창, 창 밖은 어느 턴에도 안 셈, 판정은 `turn/completed` 이벤트에 실려 옴). AC-MH-006 하위 케이스 14개(원인 10 + (a)–(d) 4)와 변이 m14–m18. REQ-MH-002·006 에 귀속 문구, design/acceptance 의 "버려진다" 단정 제거(시험으로 고정) |
| N4 스트림 시그널 구독이 등록보다 늦음 | fixed | O1: 시그널 컨텍스트·핸들러가 **소유자 진입의 첫 영속 단계**(두 소유자 모두 launch-pending 등록 전). 워처만 세션 생성 뒤 부착(이미 취소됐으면 즉시 `Close`). 규칙 R-E: O1 이후 모든 오류 반환을 컨텍스트가 취소돼 있으면 중단 오류로 매핑(O20 ④). AC-MH-010 하위 케이스 6개(`stream_SIGTERM_after_registration` 신설, 훅 단계명별), 재실행 도우미의 준비 표지(`ready <단계명>`) 대기. REQ-MH-010 개정 |
| N5 spec/design 버전·HISTORY | fixed | spec.md·design.md 0.3.0, HISTORY 0.2.0·0.3.0 행, plan·acceptance 0.3.0 |
| N6 AC-MH-002 "한참 짧음", `moai` 리터럴, `Start` 실패 시 `ctx.Err()` | fixed | AC-MH-002 절대 상한 5초, 서버 이름은 기존 상수 `moaiMCPServerKey` 로 비교하고 승인 인수와의 일치를 고정 시험(`TestManagedBrokerNameMatchesApprovalArgs`)으로, `Start` 오류 → 중단 매핑은 R-E |
| N7 쓰기 한계 문장 | fixed (공시 정정) | `call()` 의 `WriteJSON`(`:212`)이 `select`(`:215`) 앞이라 막힌 쓰기는 턴 타임아웃으로 풀리지 않음. 상한은 `Close`/연결 종료까지. 쓰기 데드라인 상수는 더하지 않기로 판단(근거 명시), design D-3·acceptance §4·REQ-MH-014 문구 정정 |

개정 뒤 REQ 16(추가 0, 이번 개정 문구 수정: REQ-MH-002·006·010·012·014), AC 16(추가 0, 이번 개정 수정: AC-MH-002·006·010·016 및 §1.3–§1.5 표 신설).

미측정(개정 시점): 설치된 `moai` 바이너리가 이 트리보다 뒤처져 있으므로 `moai spec lint` 출력은 지연 빌드의 증거일 뿐 이 트리 빌드의 판정이 아니다. `Close`∥`Start` 경합, 시그널 등록 후 창, elicitation 귀속 경합은 plan 작성자가 재현·관측하지 않았다(소스 순서는 이번 개정에서 기준 트리를 다시 읽어 확인; plan.md §C Gaps). 라이브 Codex 관측(`serverName` 값 포함)은 run 진입 조건이 아닌 이름 붙은 Gap.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
