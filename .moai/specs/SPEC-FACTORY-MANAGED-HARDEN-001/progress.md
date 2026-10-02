# progress.md — SPEC-FACTORY-MANAGED-HARDEN-001

> 단계별 진행 기록. `§E.1` 만 plan 단계(manager-spec)가 채우고, `§E.2`·`§E.3` 은 run 단계(manager-develop), `§E.4` 는 sync 단계(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-after-scope-split (재감사 대기 — 축소 범위 1차)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 12) · plan.md · acceptance.md (AC 13) · design.md (D-1, D-2) · progress.md
tier: M (축소 범위 재평가 결과 유지; 방법은 plan.md §A — 측정 `wc -l` + 가정 기반 줄 수 추정)
measured_tree: 7109e0900
open_clarifications: 0

### 분할 전 감사 3회 (F3·F4·F5 포함 범위)

| 회차 | 판정 | 점수 | audited_sha | 결과 |
|---|---|---|---|---|
| 1 | FAIL | 0.75 | `95dfd85c8` | 차단 D1–D5, 주요 D6–D7, 경미 D8–D10. 개정 0.2.0(`820eff47f`) |
| 2 | FAIL | 0.81 | `820eff47f` | 신규 차단 N1–N4, 경미 N5–N7. 개정 0.3.0(`951f2bfb6`) |
| 3 | FAIL | 0.81 | `951f2bfb6` | 차단 N8–N12(검증층: 경합 오라클, `codex_dialed` 단언, 변이표 지목 불일치, REQ-MH-011 실제 소유자 시험 부재, elicitation 귀속 구멍), 경미 N13–N15. 마지막 회차(3/3) |

판정서 로컬 사본: `.moai/reports/t1409/plan-audit-iter{1,2,3}.md`(gitignore 대상).

### 범위 분할 결정과 그 출처

- **결정**: F5(시그널 처리와 `Start`/`Close` 수명주기)를 새 카드 t1459 로 분리하고, 이 SPEC은 F3·F4 로 줄인다.
- **출처와 날짜**: 운영자 결정, 리더가 중계(2026-10-03). 3차 판정서의 Recommendation 선택지 3("범위 축소")에 해당한다.
- **리더 지시**: 축소된 이 SPEC의 다음 plan-audit 는 **축소 범위 1차(reduced-scope iteration 1)** 로 기록한다. Tier M 의 plan-auditor 상한은 2회이며 다시 상한에 닿으면 연장하지 않고 보고한다.
- 이전 텍스트(F5 포함)는 git 이력 `95dfd85c8`, `820eff47f`, `951f2bfb6` 에서만 볼 수 있고 어디에도 복사하지 않았다.

### 이 개정(0.4.0)이 한 일

**제거된 것(F5 와 함께 사라짐)**: 구 REQ-MH-010(시그널 → 정리), 011(막힌 전달 해제), 012(`Start`/`Close` 수명주기), 013(Windows 시그널·zero-syscall 한정) · 구 AC-MH-009(드라이버 취소), 010(시그널 종단), 011(시그널 도우미 Windows 빌드 중 `launch_signals.go` 부분), 016(수명주기) · 구 AC-MH-004 의 `Close` 동시성 부분, 구 AC-MH-013 의 시그널·`Close` 재현 시험 3개, 구 AC-MH-015 의 시그널 관련 항목 · 변이 m1–m13, m22 와 수명주기 변이 전부 · 구 AC-MH-010·016 하위 케이스 표(§1.3·§1.5) · `managedStepHook` 시험 이음새 전부 · `launch_signals.go`, cross-platform syscall 면제 선언, 부모 REQ-MS-012 예외, 소유자 진입 오류 매핑(R-E), design D-3·D-3.1 수명 순서표.

**번호 재부여(구 → 신)**

| 구 REQ | 신 REQ | 구 AC | 신 AC |
|---|---|---|---|
| 001–005 | 001–005 | 001 | 001 |
| 006–009 | 006–009 | 002 | 002 |
| 010–013 | **제거(F5)** | 003 | 003 |
| 014 (문서) | 010 | 004 | 004(쓰기 직렬화만) |
| 015 (RED 기준선) | 011 | 005 | 005 |
| 016 (비침범) | 012(Windows 빌드·zero-syscall 가드를 합침) | 006 | 006 |
| — | — | 007 | 007 |
| — | — | 008 | 008 |
| — | — | 009, 010 | **제거(F5)** |
| — | — | 011 | 009(Windows 빌드 가드만) |
| — | — | 012 | 010 |
| — | — | 013 | 011 |
| — | — | 014 | 012 |
| — | — | 015 | 013 |
| — | — | 016 | **제거(F5)** |

REQ 16→12, AC 16→13, 번호는 연속(MP-1).

### 3차 결함 N8–N15 처분 (축소 범위 기준)

| 결함 | 처분 | 근거 |
|---|---|---|
| N8 경합 오라클(`*_racing_start`) | **F5 와 함께 제거** | 수명주기 AC-MH-016 에만 있던 결함 |
| N9 `codex_dialed` 쓰기 단언 | **F5 와 함께 제거** | 같음 |
| N10 변이 m3·m7·m8·m9·m13 | **F5 와 함께 제거** | 수명주기·시그널 변이 |
| N10 변이 m14 | **fixed** | AC-MH-006 #11 을 "같은 쓰기 묶음의 `turn/completed(T1)` 직후 요청 + 읽기 고루틴이 요청까지 처리할 때까지 소비자(`waitTurn`) 호출 보류"로 재작성. 가장 작은 이음새: 로그 대상 주입(별도 훅 없음) — 읽기 고루틴의 처리 완료를 그 요청 로그 줄(`turn=none`)로 안다. 소비자 쪽 계수기 설계(mu5)는 `waitTurn` 이 오류를 내 붉어짐 |
| N10 변이 m17(요청마다 실패로 세기) | **declined(동치 변이) + 관측 가능한 대체** | 한 턴은 `DeliverTurn` 오류 하나만 내므로 차이가 없다. 관측값으로 로그 줄의 `broker_declined=<k>` 를 택하고 가를 수 있는 변이 mu8(`== 1` 판정)을 채택. #14 가 둘째 요청 줄 `broker_declined=2` 와 실패 1회를 단언 |
| N10 변이 m19(상수 `3` 박기) | **fixed(방법 변경)** | 한 단계 변이로는 동치. 섭동 확인(상수를 2 로 바꿔도 AC-MH-007 시험 PASS, 드라이버가 3 을 박았다면 붉어짐)과 드라이버의 상수 참조 `grep` 으로 방어(acceptance §2.4·§6 DoD 4) |
| N11 REQ-MH-011 실제 소유자 시험 부재 | **F5 와 함께 제거** | REQ 자체가 F5 |
| N12 귀속 규칙 `turn/started` 이전 구멍 | **fixed** | `armTurn` 이 직전 완료 턴 id(`prevTurnID`)를 기억하고 지우지 않음. 규칙과 약속 문장을 design D-1 결정 3 의 한 문장으로 통일(acceptance §4 가 같은 문장을 가리킴). 행 추가: #15(`turn/started` 이전 직전 턴 id → nil), #16(확정된 현재 턴과 불일치 id → nil), #17(문자열 JSON-RPC id 요청도 귀속·계수·답장 id 보존). AC-MH-006 14→17 하위 케이스 |
| N13(a) 훅 위치·`close-before-lock`·`codex_timeout` 이음새 | **부분 제거 + fixed** | 훅 위치 항목은 F5 와 함께 제거. `codex_timeout`(#8)은 이미 컨텍스트 인자를 받는 `startTurn` 에 50ms 마감 컨텍스트를 줘 이음새 없이 해결(design D-2) |
| N13(b) O20 defer 순서 | **F5 와 함께 제거** | |
| N13(c) REQ-MH-012 문면 vs F8 | **F5 와 함께 제거** | |
| N13(d) 변이→마일스톤 배정 | **fixed** | acceptance §2.4 표에 모든 유지 변이의 마일스톤(M2/M3/M4)을 지정 |
| N13(e) 경합 격자 근거 | **F5 와 함께 제거** | |
| N14 Tier 선언 | **fixed(재평가)** | Tier M 유지, 방법(측정 `wc -l` + 가정 기반 추정)과 경계 상태를 plan §A 에 명시 |
| N15(a) 훅 설정 race | **fixed(이음새 축소)** | 훅이 모두 사라져 남은 이음새는 로그 대상 `atomic.Pointer[io.Writer]` 하나 — 원자적이라 `-race` 안전, 시험은 병렬 금지 |
| N15(b) stderr 줄 잡는 법 | **fixed** | 주입 가능한 쓰기 대상(위 원자 포인터). `os.Stderr` 전역 교체 안 함 |

### t1410 의존 정리 (plan §B, spec §F 와 동일 내용)

F8·F9·F13 모두 이 카드의 F3·F4 변경에 의존하지 않는다(줄 번호 측정, 결론은 추론). F8 은 `Start`/`Close` 를 구조적으로 바꾸는 t1459 와 같은 줄 영역을 만진다. 공유 시험 파일 `managed_codex_factory_test.go` 에서 텍스트 충돌 가능(추론).

### 미측정(개정 시점)

설치된 `moai` 바이너리가 이 트리보다 뒤처져 있으므로 `moai spec lint` 출력은 지연 빌드의 증거일 뿐 이 트리 빌드의 판정이 아니다. elicitation 귀속의 읽기 고루틴 대 소비자 경합은 plan 작성자가 재현·관측하지 않았다(소스 순서와 설계 논증만). 라이브 Codex 관측(`serverName` 값 포함)은 run 진입 조건이 아닌 이름 붙은 Gap. Tier 의 줄 수 추정은 가정이다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
