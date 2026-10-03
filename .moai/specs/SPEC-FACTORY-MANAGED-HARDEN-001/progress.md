# progress.md — SPEC-FACTORY-MANAGED-HARDEN-001

> 단계별 진행 기록. `§E.1` 만 plan 단계(manager-spec)가 채우고, `§E.2`·`§E.3` 은 run 단계(manager-develop), `§E.4` 는 sync 단계(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-after-reduced-scope-audit-iter2 (0.5.1, 재감사 대기 — 운영자가 한 번 더 연 상한 안의 개정, 범위 N1·N2 한정)
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

### 축소 범위 감사 (reduced-scope iteration)

| 회차 | 판정 | 점수 | audited_sha | 결과 |
|---|---|---|---|---|
| 축소 1 | FAIL | 0.75 | `bca1e0629` | 차단 D1–D7(전부 검증층: 채택 경로, 변이 지목, 오라클 순서, 시간 상한, 로그 sink 경합), 선택 D8–D13. 개정 0.5.0(이 커밋) |

판정서 로컬 사본: `.moai/reports/t1409/plan-audit-iter4.md`(gitignore 대상). 위 "3차 결함" 표의 m14·m17 문구는 분할 시점의 기록이고 현행 변이 범위는 mu1–mu20 이다(아래 "축소 범위 감사 2차" 절).

### 이 개정(0.5.0)이 한 일 — 축소 범위 1차 결함 D1–D13 처분

원칙: 새 REQ·새 AC 없음(REQ 12·AC 13 유지). 한 단계 편집으로 적용되지 않거나 정직하게 붉어지지 않는 항목은 변이 표에서 빼 "불변 가드, 변이 미채택"(acceptance §2.5)으로 옮기고 DoD 3 을 적용 가능하고 가를 수 있는 변이로 줄였다.

| 결함 | 처분 | 근거와 위치 |
|---|---|---|
| D1 #2 `codex_failed_or_interrupted` 미채택 | **fixed(채택)** | M1 에 `TestManagedCodexNonCompletedTurnIsolated`(`failed`·`interrupted` 각각, 세 번째 턴 전달 단언)를 추가(재현 시험 5→6)하고 변이 mu15 가 #2 와 그 시험을 지목(acceptance §1.3·§2.1·§2.4) |
| D2 mu14 의 #4·#8 지목 오류 | **fixed + 이관** | mu14 지목을 #5–#7·#9 로 줄이고 그 행들이 `driveManagedFactorySession` 을 통과해 `/exit` 가 nil 을 내면 "기대한 오류가 아님"으로 붉어진다고 명시. #4 는 G2 로 이관(우선 턴은 루프 앞에서 무조건 반환, 기존 `TestManagedDriverFailureBranches` 가 고정). #8 은 `startTurn` 수준 행으로 두고 변이 mu16(`ctx.Done()` 분기가 표식 오류를 반환)이 지목 |
| D3 mu2 "11개 각각" | **fixed** | mu2 를 결과 응답 7종 + 미지 method 로 좁히고(8개, 오류 3종은 초록 명시), 오류 코드 교환 변이 mu3(`-32000` ↔ `-32601`, 4개)을 추가 |
| D4 mu17 한 단계 불가·AC-MH-008 공허, mu5 설계 교체 | **moved to guard** | mu17 삭제, AC-MH-008 은 G1(기준 트리에서도 초록, 드라이버에 store 핸들 없음 `managed_factory_session.go:300`, 재배달 증거는 기존 `TestDispatchResultExactlyOnce/lost_receipt_redelivery`, 이 개정에서 재실행 7개 PASS `ok 2.167s`). 소비자 쪽 계수기(구 mu5)는 G7 로 이관(DoD 아님, 가장 작은 적용 형태만 기술) |
| D5 로그 이음새 `-race` 주장 오류 | **fixed** | design D-1 에 세 규칙: (a) sink 는 쓰기·스냅숏이 한 잠금 아래인 타입, 시험의 `String()` 직접 접근 금지, (b) 클라이언트를 시작한 모든 시험이 종료·읽기 고루틴 종료 대기 후 `t.Cleanup` 이 포인터 복원, (c) 로그 대기가 있는 행의 명령에 `-race`. 원자 포인터는 포인터 로드만 보호한다고 정정 |
| D6 시간 상한 부재 | **fixed** | acceptance 머리말 규약: 블로킹 호출은 고루틴에서 돌리고 시험 고루틴이 5초 watchdog 으로 실패시키며, 정리에서 가짜 서버·연결을 닫고, 드라이버 시험은 `/exit` 종료자를 쓴다. AC-MH-002 의 5초가 이 watchdog 상한 |
| D7 로그 오라클 순서 의존 | **fixed** | design D-1: 로그 줄을 답장보다 먼저 쓴다. 모든 로그 단언은 상한 있는 폴링, "정확히 한 번"은 500 ms 조용한 구간 값으로 고정. 변이 mu4(로그 줄 생략)가 AC-MH-001 11개 전부를 지목 |
| D8 REQ-006/008 관계·귀속 규칙 혼재 | **fixed** | REQ-MH-006 에 "until REQ-MH-008 applies", REQ-MH-008 에 "notwithstanding REQ-MH-006" 과 상한 도달 시 로그 줄을 쓰고 반환함을 명시. 귀속 규칙 상세는 design.md D-1 결정 3 으로 가리키고 REQ-MH-006 에는 한 절만 남김. 새 REQ 는 만들지 않음 |
| D9 줄 번호·상수 2개 | **fixed** | 비교 상수를 `moaiMCPServerKey`(`mcp_server.go:57`)로 못 박고 같은 값 `moaiMCPServerName`(`:54`)의 존재를 적음 |
| D10 늦은 완료 프레임·쓰기 뮤텍스 | **fixed(문면) + G5** | `X == prevTurnID` 인 `turn/completed` 는 창을 닫지 않는다 한 줄과 "뮤텍스는 `WriteJSON` 호출만 감싼다, 답장은 상태 잠금을 푼 뒤 쓴다" 한 줄. 프로토콜 동작 미관측이라 시험 행·변이 없음(G5) |
| D11 CHANGELOG 부모 엔트리 정정 | **fixed** | acceptance §1.2 에 `grep -c "후속 카드 t1409 대상" CHANGELOG.md` → `0`(exit 1) 행. 기준 트리 측정값 1 |
| D12 템플릿 스위트 임대 | **fixed** | 템플릿 스위트를 AC 근거에서 빼고 plan M4 의 슬롯 임대(`moai slot acquire --resource internal-template-suite --max-duration 15m`) 안 스모크로만 둠. 영향 범위 근거 유지 |
| D13 사소한 문면 | **fixed** | (a) AC-MH-001 Then 에 `item/tool/call` 의 `success:false` 응답 형태, (b) 로그·오류의 method 를 `%q` 인용, (c) `id: null` 은 분류되지 않고 버려진다고 한 줄 |

**0.5.0 시점의 변이 표:** mu1–mu18(0.5.1 에서 mu19·mu20 이 더해져 현행은 mu1–mu20, acceptance §2.4). **불변 가드 G1–G8:** acceptance §2.5. 변이가 아닌 점검 둘: C1(RED·수리 커밋 분리, `merge-base --is-ancestor` 쌍)과 P1(상수를 일시적으로 2 로 바꿔도 AC-MH-007 시험 PASS). DoD 3 은 0.5.1 에서 mu1–mu20 전부와 P1·C1 로 갱신됐다.

**개정 시점에 이 개정이 실제로 재측정한 것:** `go test ./internal/cli -list '^.*(Managed|managed).*$'` 선택 수 66, `TestDispatchResultExactlyOnce` 7개 PASS(`ok … 2.167s`), 기준 `CHANGELOG.md` 의 `후속 카드 t1409 대상` 1건, `mcp_server.go:57`·`managed_factory_session.go:300-304` 줄 확인.

### 축소 범위 감사 2차와 상한 재개방

| 회차 | 판정 | 점수 | audited_sha | 결과 |
|---|---|---|---|---|
| 축소 2 | FAIL | 0.78 | `d43a52dac` | 차단 N1(mu18 이 `below_ceiling_continues` 를 못 붉힘, 시나리오 부재)·N2(#11 미채택, 채택 회계 문장 모순), 선택 N3–N9(이 개정에서 건드리지 않음). MP-1..MP-9 전부 통과, D1–D13 해소. 개정 0.5.1(이 커밋) |

- **상한 재개방**: Tier M 의 plan-auditor 상한(2회)에 닿았으나 운영자가 **정확히 한 번 더** 열었다. 출처: 리더 중계(2026-10-03). 범위는 N1·N2 로 한정, REQ·design·plan 구조는 건드리지 않는다. 판정서 로컬 사본: `.moai/reports/t1409/plan-audit-iter5.md`(gitignore 대상).
- **현행 변이 범위는 mu1–mu20**(acceptance §2.4). 위 0.5.0 절의 "mu1–mu18" 서술은 그 시점의 기록이다.

### 이 개정(0.5.1)이 한 일 — N1·N2 처분

| 결함 | 처분 | 근거 |
|---|---|---|
| N1(a) mu18 지목 과장 | **fixed** | mu18 은 `at_ceiling_returns` 하나만 지목. 연속 실패 k 에서 `k >= N` 와 `k > N` 은 `k == N` 에서만 갈린다 |
| N1(b) `below_ceiling_continues` 시나리오 부재 | **fixed** | AC-MH-007 Given-When-Then 에 세 하위 케이스를 모두 정의(N−1번 실패 뒤 반환하지 않고 다음 성공 턴 전달 후 `/exit` 로 nil; `success_resets` 순서: N−1 실패, 성공, 실패 1) |
| N1(c) 붉히는 변이 지정 | **fixed(싼 길)** | mu13 의 붉어지는 목록에 `below_ceiling_continues` 추가. mu13 의 편집 결과가 기준 코드 `managed_factory_session.go:345-347` 의 첫 오류 반환과 같으므로 그 줄을 따라 걸었다: N−1번 표식 실패 시나리오의 첫 실패에서 `DeliverTurn` 오류가 그대로 반환돼 드라이버가 `/exit` 에 닿지 못한다 → 반환값이 nil 이 아니고 받은 턴이 모자라 붉다. `success_resets` 는 mu17(성공이 횟수를 되돌리지 않으면 N−1 뒤 성공 뒤 실패가 N 번째가 되어 반환)이 죽이고 mu18 은 죽이지 못함을 확인해 귀속 유지 |
| N2(i) 면제 (2) 오기와 #11–#17 컴파일 | **fixed** | 면제 (2) 를 17행 전수 채택표로 재작성: 이 시험은 M3 에서 쓰며 새 심볼(`armTurn()`, 표식 오류, 로그 sink)을 불러 기준 트리에서 컴파일되지 않음. #17 은 mu12 소속(가드 아님), 가드는 #4(G2) 하나 |
| N2(ii) #11 채택 | **fixed** | 새 변이 mu19(`turn/completed` 가 창을 닫지 않음: 후행 요청이 열린 창에서 규칙 (3)으로 T1 에 귀속돼 `turn=T1`, #11 의 `turn=none` 대기가 5초 상한에서 붉음; `waitTurn` nil 단언은 판정이 이미 완료 이벤트에 실려 초록). design D-1 결정 3 의 완료 처리와 귀속 규칙 (3)을 따라 걸어 한 줄 편집임을 확인 |
| N2(iii) `broker_declined=0` 변이 | **fixed(변이 채택)** | #11 에 `broker_declined=0` 단언을 더하고, 그 값 단언만이 잡는 변이를 mu20(닫힌 창 분기도 거부 수를 올림, 줄은 `turn=none … broker_declined=1`)으로 채택. 단언만 두고 변이를 안 두면 그 단언의 붉음이 관측되지 않아 미채택이므로 가드로 두지 않고 변이로 올렸다 |
| N2(iv) G7 전제 오류 | **fixed** | G7 은 소비자 쪽 계수기 설계 교체에만 남기고, #11 의 채택이 한 단계 변이 mu19·mu20 이 운반한다고 고침 |
| N2(v) §2.5 머리말 | **fixed** | "이 목록 밖은 mu1–mu18 이 덮는다" 를 실제 채택 경로(M1 RED / §2.4 변이 / §2.2 구조 확인 / C1)별 서술로 바꿈. AC-MH-002·003 은 M1 RED 로만, AC-MH-010 은 grep 으로만 채택되므로 변이 문장은 거짓이었다 |

DoD 3 = mu1–mu20 전부와 P1·C1. 가드 목록 G1–G8 은 항목 수 변화 없음(G7 문구만 정정). 변이 적용은 run 단계 몫이고 이 개정의 모든 "붉어진다" 지목은 기준 코드와 design.md 를 따라 걸은 문서 대조이며 실행이 아니다.

### t1410 의존 정리 (plan §B, spec §F 와 동일 내용)

F8·F9·F13 모두 이 카드의 F3·F4 변경에 의존하지 않는다(줄 번호 측정, 결론은 추론). F8 은 `Start`/`Close` 를 구조적으로 바꾸는 t1459 와 같은 줄 영역을 만진다. 공유 시험 파일 `managed_codex_factory_test.go` 에서 텍스트 충돌 가능(추론).

### 미측정(개정 시점)

설치된 `moai` 바이너리가 이 트리보다 뒤처져 있으므로 `moai spec lint` 출력은 지연 빌드의 증거일 뿐 이 트리 빌드의 판정이 아니다. elicitation 귀속의 읽기 고루틴 대 소비자 경합은 plan 작성자가 재현·관측하지 않았다(소스 순서와 설계 논증만). 이 개정(0.5.0)의 오라클·변이표·시험 규약(로그 선기록과 폴링, 단일 sink, 5초 watchdog, 500 ms 조용한 구간, mu1–mu18 의 "붉어지는 케이스" 지목)은 문서상으로만 고쳤고 어느 것도 실행해 보지 않았다(재현 시험이 아직 없다) — 실행 관측은 run 의 M1·M2·M3 몫이다. 연속 실패 상한 3 은 UNMEASURED(저장소의 "최대 3회 재시도" 관례에서 가져옴). 라이브 Codex 관측(`serverName` 값 포함)은 run 진입 조건이 아닌 이름 붙은 Gap. Tier 의 줄 수 추정은 가정이다.

## §E.2 Run-phase Evidence

### M1 — RED 기준선 (card t1409, manager-develop, cycle_type=tdd)

원문 출력과 시험별 붉은 이유는 추적 파일 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 에 있다. 이 블록은 명령·관측·귀속만 인용한다. 귀속은 모두 `(this run, this tree, HEAD fe79bfa0e)` 이다.

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 사전: 빌드 | `go build ./...` | 출력 없음 | 0 |
| 사전: Windows 빌드 | `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음 | 0 |
| 선택 수 (전) | `go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 `grep -c '^Test'` 로 센 값(원문 명령은 파이프 없이 acceptance.md §1.1 AC-MH-013 블록) | `66` | — |
| 선택 수 (후) | 같은 명령 | `72` (최상위 6개 추가) | — |
| 시험 6개 선택 확인 | `go test ./internal/cli -list '^(TestManagedCodexServerRequestPolicy\|…)$'` (acceptance.md 와 같은 6개 이름) | 이름 6줄 + `ok  github.com/modu-ai/moai-adk/internal/cli  1.238s` | 0 |
| RED 1 | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` | `--- FAIL: TestManagedCodexServerRequestPolicy (2.02s)`, 하위 11개 `--- FAIL`(`no answer to server request 101..111 … within 2s`) | 1 |
| RED 2 | `go test ./internal/cli -run '^TestManagedCodexTurnSurvivesServerRequest$' -count=1 -v` | `DeliverTurn still blocked after 5s …`(during_turn), `… got no answer within 5s`(between_turns) | 1 |
| RED 3 | `go test ./internal/cli -run '^TestManagedCodexServerRequestIDCollision$' -count=1 -v` | `call result "", want the genuine response …`, `call error = managed codex app server connection closed …` | 1 |
| RED 4 | `go test -race ./internal/cli -run '^TestManagedCodexDeclinedBrokerElicitationFailsTurn$' -count=1 -v` | 하위 3개 `DeliverTurn = nil for a turn whose MoAI broker elicitation was declined, want a non-nil error` | 1 |
| RED 5 | `go test -race ./internal/cli -run '^TestManagedDriverIsolatesTurnFailure$' -count=1 -v` | `driver returned managed Factory turn failed: overloaded after one failed turn …`, 전달된 턴 2개(기대 3개) | 1 |
| RED 6 | `go test -race ./internal/cli -run '^TestManagedCodexNonCompletedTurnIsolated$' -count=1 -v` | `driver returned managed codex turn fake-turn-2 ended as failed …`, `op-interrupted`·`op-normal` 미전달 | 1 |
| 기존 managed 시험(새 6개 제외) | `go test ./internal/cli -run '^.*(Managed\|managed).*$' -skip '^(…새 시험 6개…)$' -count=1 -v` (red-baseline.md 에 전체 원문) | `--- PASS` 63, `--- SKIP` 3, `--- FAIL` 0, `ok  github.com/modu-ai/moai-adk/internal/cli  41.464s` | 0 |
| 정적 | `go vet ./internal/cli` / `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` / `golangci-lint run --timeout=5m ./internal/cli/...` | 출력 없음 / 출력 없음 / `0 issues.` | 0 / 0 / 0 |
| 형식 | `gofmt -l` 을 이 카드가 건드린 두 시험 파일에 | 출력 없음 | 0 |

모든 `go test` 는 환경 정리 접두(`unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test …`)를 붙인 단일 호출이다. 붉은 이유는 모두 acceptance.md §2.1 이 정한 이유(소유자가 서버 요청에 답하지 않음, id 충돌·문자열 id, 비완료·`is_error` 턴에서 드라이버가 반환, 거부된 브로커 elicitation 턴이 nil)와 일치하고 컴파일 오류·패키지 타임아웃·하네스 실패로 붉은 것은 없다.

미관측: 시험 6개의 GREEN 경로(M2·M3 몫), 로그 이음새 단언(이음새는 M2/M3 이 도입), 라이브 Codex. `gofmt -l internal/cli` 는 기준 트리의 기존 파일 `internal/cli/mcp_claude.go` 한 줄을 낸다(이 카드가 만든 것이 아니다).

### M2 — F3 서버 요청 응답 (card t1409, manager-develop, cycle_type=tdd)

귀속은 모두 `(this run, this tree, HEAD f042a05a0 위의 M2 작업 트리, 커밋 전)` 이다. 모든 `go test` 는 환경 정리 접두(`unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test …`)를 붙인 단일 호출이다. 변경 파일: `internal/cli/managed_codex_factory.go`(프레임 분류·`id` 원문 보존·읽기 고루틴 응답·정책표·쓰기 뮤텍스·턴 창·로그 이음새 `managedLogOutput`), `internal/cli/managed_hardening_test.go`(로그 sink·정책 시험의 로그 단언·`TestManagedCodexConcurrentWrites`·`TestManagedCodexCompletionEventCarriesBrokerVerdict`). `defaults.go`·`Close`·`Start()` 자원 생성부·`store.go`·부모 SPEC 은 건드리지 않았다.

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 사전: 빌드 | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음 / 출력 없음 | 0 / 0 |
| 사전: 선택 수 | `go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 `grep -c '^Test'` 로 센 값 | `72` (M1 직후) → M2 뒤 `74` (`ConcurrentWrites`·`CompletionEventCarriesBrokerVerdict` 추가) | — |
| GREEN: M1 시험 3개 | `go test -race ./internal/cli -run '^(TestManagedCodexServerRequestPolicy\|TestManagedCodexTurnSurvivesServerRequest\|TestManagedCodexServerRequestIDCollision)$' -count=1 -v` | 세 시험 `--- PASS`, 하위 11 + 2 + 2 개 PASS, `ok … 3.347s` (M1 RED 원문: `red-baseline.md`, 같은 세 시험이 `--- FAIL`) | 0 |
| GREEN: AC-MH-001..004 + 판정 시험 | `go test -race ./internal/cli -run '^(TestManagedCodexServerRequestPolicy\|TestManagedCodexTurnSurvivesServerRequest\|TestManagedCodexServerRequestIDCollision\|TestManagedCodexConcurrentWrites\|TestManagedCodexCompletionEventCarriesBrokerVerdict)$' -count=1 -v` (선행 `-list` 로 이름 5개 확인) | 5개 최상위 `--- PASS`(정책 11 하위, 경합 200 라운드 `PASS (0.38s)`, 판정 3 하위), `WARNING: DATA RACE` 0건, `ok … 3.852s` | 0 |
| M3 시험 3개는 계속 RED | `go test -race ./internal/cli -run '^(TestManagedCodexDeclinedBrokerElicitationFailsTurn\|TestManagedDriverIsolatesTurnFailure\|TestManagedCodexNonCompletedTurnIsolated)$' -count=1 -v` | 세 시험 `--- FAIL`, 이유는 M1 과 같음(`DeliverTurn = nil …`, `turns delivered` 2개, `op-interrupted`·`op-normal` 미전달) | 1 |
| 회귀 | `go test -race ./internal/cli -run '^.*(Managed\|managed).*$' -skip '^(…M3 시험 3개…)$' -count=1 -v` | `-list` 74개 중 M3 3개 제외, `--- FAIL` 0건, `--- SKIP` 3건(기존 3건: `TestManagedCodexFakeAppServer`, `TestManagedCodexFactoryBrokerLive`, `TestManagedLoopbackChild`), `ok  github.com/modu-ai/moai-adk/internal/cli  131.919s` | 0 |
| 정적 | `gofmt -l internal/cli/managed_codex_factory.go internal/cli/managed_hardening_test.go` / `go vet ./internal/cli` / `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` / `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음 ×4 | 0 ×4 |
| lint | `golangci-lint run --timeout=5m ./internal/cli/...` (`golangci-lint has version v2.1.6 built with go1.26.8`) | 1차 `QF1001: could apply De Morgan's law` 2건(`noteElicitation` 조건식) → `switch` 로 재작성 후 `0 issues.` | 1 → 0 |
| 경계 | `grep -rn 'syscall\.' internal/cli/managed_*.go` (양성 대조 `grep -n 'syscall\.' internal/cli/mcp_server.go` → `:123` 1줄) / `grep -rn 'AskUserQuestion' internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go` | 출력 0줄 / 출력 0줄 | 1 / 1 |

**TDD 순서 정직 기록**: 앞 세 시험(M1)은 RED 가 `red-baseline.md` 에 이미 있다. 프로덕션 코드를 먼저 쓰고 그 뒤에 새 단언(로그 줄 단언, `TestManagedCodexConcurrentWrites`, 판정 시험)을 붙였으므로 그 단언들의 RED 를 "코드 없는 트리"에서는 보지 못했다 — 대신 변이로 관측했다: 로그 줄 단언의 RED = mu4(11개 전부 붉음), 경합 시험의 RED = mu5(acceptance.md §2.1 면제 (1)이 정한 채택 경로). 판정 시험(`CompletionEventCarriesBrokerVerdict`)은 SPEC 이 정하지 않은 M2 보조 시험이며 변이로 채택하지 않았다(M3 의 AC-MH-006 #11–#17 이 정본).

**변이 확인 mu1–mu5** (한 번에 하나, 편집기로 적용·복구; 복구 뒤 `git diff --stat` 와 `cmp`(변이 전에 떠 둔 사본) 로 변이 전 상태 복원을 확인. 변이 전 `git diff --stat`: `managed_codex_factory.go | 260 +++…`, `managed_hardening_test.go | 240 +++…`, `2 files changed, 461 insertions(+), 39 deletions(-)`; 복구 뒤 mu1·mu2·mu3·mu5 는 같은 줄·`cmp` 출력 없음, mu4 는 `cmp` 출력 없음만 확인했다. 변이 뒤 lint 지적(`noteElicitation` 의 De Morgan 2건)을 고치려 그 함수 조건식을 `switch` 로 다시 썼고(어느 변이도 그 줄을 건드리지 않음), 그 뒤에 위 표의 회귀·M3 RED·정적 검사를 돌렸다). 명령은 모두 `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v`(mu5 만 `'^TestManagedCodexConcurrentWrites$'`), 트리 HEAD `f042a05a0` + M2 작업 트리, exit 1.

| 변이 | 편집(한 곳) | 결정 출력 | 붉은 케이스(관측) |
|---|---|---|---|
| mu1 | 결정 다섯 종류의 응답 값을 `accept`/`approved` 로 | `result {"decision":"accept"}, want {"decision":"decline"}` 등, 최상위 `answer to request 101 grants approval: {"decision":"accept"}` | 하위 5개: `command_execution_approval`, `file_change_approval`, `mcp_elicitation`, `legacy_apply_patch_approval`, `legacy_exec_command_approval`. 나머지 6개 `--- PASS` |
| mu2 | 모든 method 를 `-32000` 오류로 답함(`answerServerRequest` 한 줄) | `answered with error {Code:-32000 …}, want result …`, 미지 method: `error code -32000 …, want -32601` | 하위 8개(결과 응답 7종 + `unknown_method`) 붉음, 오류 3종(`tool_request_user_input`, `chatgpt_token_refresh`, `attestation_generate`) `--- PASS` |
| mu3 | `-32000` ↔ `-32601` 상수값 교환 | `error code -32601 (…), want -32000` ×3, `error code -32000 (…), want -32601` | 하위 4개(`tool_request_user_input`, `chatgpt_token_refresh`, `attestation_generate`, `unknown_method`)만 붉음, 나머지 7개 `--- PASS` |
| mu4 | 응답 로그 줄을 쓰지 않음(`managedLogf` → `_ = line`) | `log has no line "Factory server request answered: \"…\" -> …" within 5s` ×11, `--- FAIL: TestManagedCodexServerRequestPolicy (55.60s)` | 하위 11개 전부 붉음(각 5.0 s 상한 — 시험 자신의 폴링 상한) |
| mu5 | 연결 쓰기 뮤텍스 제거(`write` 의 `Lock`/`Unlock` 두 줄) | `WARNING: DATA RACE`, 쓰는 쪽 스택 `managedCodexAppClient.write()` ← `answerServerRequest()`(읽기 고루틴)와 `write()` ← `call()`(시험 고루틴), `gorilla/websocket (*Conn).beginMessage`·`flushFrame` 읽기/쓰기 경합 (출력이 길어 도구가 중간을 잘라 `--- FAIL` 줄 자체는 확인하지 못함) | `TestManagedCodexConcurrentWrites` 붉음(첫 실행에서, 재시도 불필요). exit 1 |

**미관측(Gaps)**: ① mu5 의 `--- FAIL: TestManagedCodexConcurrentWrites` 줄(도구 출력 잘림; 경합 보고와 exit 1 은 관측) ② 라이브 Codex(`serverName` 값 포함, design.md D-1 Gap) ③ `X == prevTurnID` 중복 완료 프레임 방어(G5, 시험 행 없음) ④ 쓰기 막힘 상한(공시한 한계) ⑤ 회귀 `--- PASS` 최상위 개수는 직접 세지 않았다(`-v` 출력을 파일로 돌릴 수 없는 레인 가드 때문; `-list` 74개 − `-skip` 3개 = 71개 실행, SKIP 3, FAIL 0 에서 산술로만 도출) ⑥ AC-MH-013 의 config/guardstate/template 인접 패키지는 M4 몫이라 이 마일스톤에서 돌리지 않았다.

### M3 — F4 턴 단위 실패 격리와 elicitation 판정 연결 (card t1409, manager-develop, cycle_type=tdd)

귀속은 모두 `(this run, this tree, HEAD 0f852b527 위의 M3 작업 트리, 커밋 전)` 이다. 모든 `go test` 는 환경 정리 접두(`unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test …`)를 붙인 단일 호출이고 출력은 파일로 돌려 읽었다(스크래치 파일은 인용 대상이 아니며 필요한 줄만 아래에 옮겼다). 변경 파일: `internal/cli/managed_factory_session.go`(표식 오류 `errManagedTurnFailed`, `pumpManagedStreamTurn` 의 `result.is_error` 표식, 드라이버 배달 루프의 격리·연속 횟수·상한·로그), `internal/cli/managed_codex_factory.go`(`observe` 가 완료 이벤트의 판정을 `brokerFailed` 표에 담고 `waitTurn` 이 비완료 종료와 판정을 표식 오류로 반환), `internal/config/defaults.go`(`DefaultManagedSessionMaxConsecutiveTurnFailures = 3`, `DefaultManagedCodexTurnTimeout` 바로 뒤, UNMEASURED 주석), `internal/cli/managed_hardening_test.go`(새 시험 4개 + M1 시험 둘의 단언 보강). `Close`·`Start()` 자원 생성부·`store.go`·`factoryMoAIMCPApprovalArgs`·부모 SPEC·`managedSession` 인터페이스·`driveManagedFactorySession` 시그니처는 건드리지 않았다.

**TDD 순서 정직 기록.** (1) RED-now 구조 확인: 상수를 더하기 전 `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` → 출력 없음, exit 1. (2) 시험이 컴파일되도록 **심볼 선언만** 먼저 더했다(상수 값과 미사용 `errManagedTurnFailed` 변수; 어떤 동작도 바꾸지 않음). (3) 시험을 모두 쓴 뒤 그 트리(HEAD `0f852b527` + 심볼 선언)에서 RED 를 관측했다. (4) 그 다음에야 표식 부착·드라이버·`waitTurn` 코드를 썼다. 심볼 선언은 RED 관측 전에 있었고 동작 코드는 후에 있었다.

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 사전: 빌드 | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음 / 출력 없음 | 0 / 0 |
| 사전: 선택 수 | `go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 `grep -c '^Test'` 로 센 값 | `74` (기대 74) | — |
| 시험 7개 선택 확인 | `go test ./internal/cli -list '^(TestManagedTurnFailureClassification\|TestManagedBrokerNameMatchesApprovalArgs\|TestManagedDriverConsecutiveFailureCeiling\|TestManagedFailedTurnLeavesClaimUntouched\|TestManagedDriverIsolatesTurnFailure\|TestManagedCodexDeclinedBrokerElicitationFailsTurn\|TestManagedCodexNonCompletedTurnIsolated)$'` | 이름 7줄 + `ok … 1.793s` | 0 |

**RED (M2 트리 + 심볼 선언, 동작 코드 없음)** — `go test -race ./internal/cli -run '^(위 7개)$' -count=1 -v`, exit 1, `FAIL … 26.410s`. 붉은 것은 전부 시험 소유 단언 또는 시험 소유 5 s 폴링 상한이고 컴파일 오류·패키지 타임아웃은 없었다.

| 시험 / 하위 케이스 | 결정 출력(원문 줄) | 붉은 이유 |
|---|---|---|
| `TestManagedCodexDeclinedBrokerElicitationFailsTurn` 3개 | `DeliverTurn = nil for a turn whose MoAI broker elicitation was declined, want a non-nil error` | 기준 `waitTurn` 이 `completed` 턴에 nil |
| `TestManagedDriverIsolatesTurnFailure` | `driver returned managed Factory turn failed: overloaded after one failed turn, want it to keep delivering …` · `turns delivered = [… "op-fails"], want [… "op-fails" "op-succeeds"]` · `log has no line starting "Factory turn failed (1/3 consecutive): " within 5s` | 드라이버가 첫 오류에서 반환, 로그 없음 |
| `TestManagedCodexNonCompletedTurnIsolated` | `driver returned managed codex turn fake-turn-2 ended as failed after a non-completed turn …` · `turn "op-interrupted" was never delivered` · `turn "op-normal" was never delivered` | 드라이버가 `failed` 에서 반환 |
| 분류 #1 `stream_is_error_after_priming` | `an is_error stream result = managed Factory turn failed: overloaded, want a turn-scoped error (errors.Is errManagedTurnFailed)` + 드라이버 계속·로그 단언 | 표식 없음, 드라이버 반환 |
| 분류 #2 `codex_failed_or_interrupted` | `a turn that ended failed = … want a turn-scoped error` / `… interrupted …` | 표식 없음 |
| 분류 #3 `codex_moai_elicitation_turn_completed` | `DeliverTurn = <nil>, want a turn-scoped error for a completed turn with a declined broker elicitation` | 판정 미연결 |
| 분류 #12 `normal_turn_after_declined_turn` | `the declined turn T1 = <nil>, want a turn-scoped error` (이 판의 #12 는 이후 변경됨 — 아래 발견 1) | 판정 미연결 |
| 분류 #13 `elicitation_after_turn_start_response` | `a request before turn/started = <nil>, want it to fail T2 with a turn-scoped error` | 판정 미연결 |
| 분류 #14 `two_requests_in_one_turn` | `two requests in one turn = <nil>, want exactly one turn-scoped failure` | 판정 미연결 |
| 분류 #17 `string_jsonrpc_id_request_counts` | `T2 = <nil>, want a turn-scoped failure: the JSON-RPC id form does not matter to attribution` | 판정 미연결 |
| 연속 실패 `at_ceiling_returns` | `driver error "managed Factory turn failed: overloaded #1" does not state the count ("3 consecutive")` · `2 turns delivered, want the priming turn plus 3` · `no "Factory turn failed (3/3 consecutive): " line …` | 첫 오류에서 반환, 횟수·로그 없음 |
| 연속 실패 `below_ceiling_continues` | `driver returned … overloaded #1 after 2 failures (below the ceiling), want it to deliver the next turn and end nil on /exit` | 첫 오류에서 반환 |
| 연속 실패 `success_resets` | `0 lines of "Factory turn failed (1/3 consecutive): ", want 2 …` | 로그·리셋 없음 |

RED 에서 **이미 초록이었던 것(기준 드라이버가 모든 오류를 반환하고 표식·판정이 아직 없어 "nil 이 기대인 행"과 "세션 치명 행"이 자명하게 통과)**: 분류 #4(G2 가드), #5·#6·#7·#9(mu14 가 채택), #8(mu16), #10(mu9), #11(mu19·mu20), #15(mu10), #16(mu11), `TestManagedBrokerNameMatchesApprovalArgs`(G3 가드), `TestManagedFailedTurnLeavesClaimUntouched`(AC-MH-008 G1 가드 — 라벨대로 수리 전에도 초록). 이 행들의 채택은 아래 변이 확인이다.

**GREEN (같은 7개 시험, 동작 코드 작성 뒤)** — `go test -race ./internal/cli -run '^(위 7개)$' -count=1 -v`, exit 0, `ok  github.com/modu-ai/moai-adk/internal/cli  7.817s`: 최상위 `--- PASS` 7개, 분류 하위 17개·연속 실패 하위 3개·`DeclinedBrokerElicitationFailsTurn` 하위 3개 PASS, `WARNING: DATA RACE` 0건. (이후 #12 를 조정해 분류 시험만 다시 돌려 `ok … 3.147s`, exit 0.)

**변이 확인 mu6–mu20 + P1** (한 번에 하나, 편집기로 적용·복구. 시험 파일 `managed_hardening_test.go` 의 변이 전 사본과 프로덕션 3파일의 변이 전 사본을 작업 트리 밖에 떠 두었고, 복구마다 네 파일 모두 `cmp` 출력 없음과 `git diff --stat` 동일(`4 files changed, 654 insertions(+), 13 deletions(-)`)을 확인했다). 명령은 모두 `go test -race ./internal/cli -run '^(위 7개)$' -count=1 -v`(mu9 만 M2 의 `TestManagedCodexCompletionEventCarriesBrokerVerdict` 를 더함), exit 1.

| 변이 | 편집(한 곳) | 결정 출력 | 붉은 케이스(관측) |
|---|---|---|---|
| mu6 | `armTurn` 의 `c.brokerDeclined = 0` 삭제 | `the next, normal turn T2 = managed Factory turn failed: … completed but a MoAI broker elicitation was declined during it, want nil (armTurn resets the count)` | #12 만 |
| mu7 | `noteElicitation` 조건에 `c.turnID == ""` 추가 | `a request before turn/started = <nil>, want it to fail T2 …` | #13, M1 `…FailsTurn/request_before_turn_started` |
| mu8 | `declined := c.brokerDeclined > 0` → `== 1` | `two requests in one turn = <nil>, want exactly one turn-scoped failure` | #14, M1 `…FailsTurn/two_requests_in_one_turn` |
| mu9 | `noteElicitation` 의 `serverName != moaiMCPServerKey` 비교 삭제 | `log has no "serverName=\"other-server\" turn=none broker_declined=0"` (시험 소유 5 s 폴링 상한) | #10, M2 `…CompletionEventCarriesBrokerVerdict/other_server_request` |
| mu10 | `noteElicitation` 의 `requestTurn != "" && requestTurn == c.prevTurnID` 갈래(3줄) 삭제 — design.md D-1:67 귀속 규칙 (2)의 비교(아래 발견 2) | `a late request of the previous turn was not logged uncounted within 5s` (5 s 폴링 상한) | #15 만 |
| mu11 | `noteElicitation` 의 `requestTurn != c.turnID` 갈래(3줄) 삭제 | `a request for another turn was not logged uncounted within 5s` (5 s 폴링 상한) | #16 만 |
| mu12 | `waitTurn` 의 `brokerFailed := c.brokerFailed[turnID]` → `:= false` | `DeliverTurn = <nil>, want a turn-scoped error for a completed turn …`, `T2 = <nil>, want a turn-scoped failure …` | #3, #13, #14, #17, M1 `…FailsTurn` 3개. #11·#12·#15·#16 초록 |
| mu13 | 드라이버 `!errors.Is(err, errManagedTurnFailed)` → `err != nil` (첫 오류에서 반환, `:345-347` 과 같은 결과) | `driver returned managed Factory turn failed: overloaded after one failed turn …` | AC-MH-005, #1, `below_ceiling_continues`; 추가로 M1 두 시험과 `at_ceiling_returns`·`success_resets` 도 붉음(SPEC 이 허용한 "붉을 수 있음") |
| mu14 | 드라이버 분류 조건을 `if false` 로(표식 없는 오류도 로그 후 계속) | `driver returned <nil>, want the session-fatal error containing "managed session output closed"` 등 | #5, #6, #7, #9 만 (#4 는 G2 로 초록) |
| mu15 | `waitTurn` 비완료 종료 오류에서 `%w` 표식 제거 | `a turn that ended failed = managed codex turn T1 ended as failed, want a turn-scoped error` / `… interrupted …` | #2(`failed`·`interrupted` 각각), M1 `TestManagedCodexNonCompletedTurnIsolated` |
| mu16 | `waitTurn` 의 `ctx.Done()` 분기가 `fmt.Errorf("%w: %w", errManagedTurnFailed, ctx.Err())` 반환 | `a per-turn timeout managed Factory turn failed: context deadline exceeded carries the turn-scoped marker, want session-fatal` | #8 만 |
| mu17 | 성공 턴의 `consecutiveFailures = 0` 삭제 | `driver returned 3 consecutive managed Factory turn failures, last: … one more, want a success to reset the count so the last failure is 1/3` | `success_resets` 만 |
| mu18 | `consecutiveFailures >= N` → `> N` | `driver returned nil after 3 consecutive turn-scoped failures, want the last failure` | `at_ceiling_returns` 만 |
| mu19 | `noteTurnCompleted` 의 `c.open = false` 삭제 | `the trailing request was not logged with turn=none within 5s` (5 s 폴링 상한) | #11 만 (`waitTurn` nil 단언은 초록 — 판정이 먼저 이벤트에 실렸다) |
| mu20 | 닫힌 창 분기에서 브로커 요청이면 `c.brokerDeclined++` (3줄 삽입, 발견 3) | `the trailing request was counted: want serverName="moai" turn=none broker_declined=0` (값 단언, 0.01 s) | #11 만 (mu19 와 달리 `turn=none` 대기는 통과) |
| P1 | `DefaultManagedSessionMaxConsecutiveTurnFailures` 를 일시적으로 2 로 | `--- PASS: TestManagedDriverConsecutiveFailureCeiling` 와 하위 3개 PASS, `ok … 2.915s`, exit 0 | 드라이버가 숫자를 박지 않았음 |

구조 확인(DoD 4 의 `grep`): `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/config/defaults.go` → `:122`(주석)·`:130`(정의), exit 0 (RED-now 는 출력 없음 exit 1). `grep -n DefaultManagedSessionMaxConsecutiveTurnFailures internal/cli/managed_factory_session.go` → `:364`(로그)·`:365`(비교), exit 0. 이 상수를 참조하는 파일은 `defaults.go`·`managed_factory_session.go`·`managed_hardening_test.go` 셋뿐. `managed_factory_session.go` 에 맨 숫자 `3` 없음(`grep -nE '[^A-Za-z0-9_"/.-]3[^0-9]'` → exit 1).

**정적·경계·회귀**

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 형식 | `gofmt -l` 을 건드린 4파일에 | 출력 없음 | 0 |
| vet | `go vet ./internal/cli ./internal/config` | 출력 없음 | 0 |
| Windows | `GOOS=windows GOARCH=amd64 go build ./...` / `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` | 출력 없음 / 출력 없음 | 0 / 0 |
| lint | `golangci-lint run --allow-serial-runners --timeout=10m ./internal/cli/... ./internal/config/...` (`golangci-lint has version v2.1.6 built with go1.26.8`) | `0 issues.` (첫 호출은 `Error: parallel golangci-lint is running` 으로 거부되어 exit 3 — 다른 세션의 실행 때문이며 재시도는 직렬 대기 플래그) | 0 |
| config | `go test ./internal/config -count=1` | `ok  github.com/modu-ai/moai-adk/internal/config  35.366s` | 0 |
| 경계 | `grep -rn 'syscall\.' internal/cli/managed_*.go` (양성 대조 `grep -n 'syscall\.' internal/cli/mcp_server.go` → `:123` 1줄) / `grep -n 'AskUserQuestion' internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go` | 출력 0줄 / 출력 0줄 | 1 / 1 |
| 선택 수 | `go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 `grep -c '^Test'` 로 센 값 | `78` (74 + 새 4: `TestManagedTurnFailureClassification`, `TestManagedBrokerNameMatchesApprovalArgs`, `TestManagedDriverConsecutiveFailureCeiling`, `TestManagedFailedTurnLeavesClaimUntouched`) | — |
| 회귀 | `go test -race ./internal/cli -run '^.*(Managed\|managed).*$' -count=1 -v` | 최상위 `--- PASS` 75, `--- SKIP` 3(`TestManagedCodexFakeAppServer`, `TestManagedCodexFactoryBrokerLive`, `TestManagedLoopbackChild`), `--- FAIL` 0, `WARNING: DATA RACE` 0, 끝 줄 `ok  github.com/modu-ai/moai-adk/internal/cli  45.286s` (75 + 3 = 78) | 0 |

**발견(SPEC 을 고치지 않고 기록)**
1. **mu12 대 #12.** #12 의 첫 판에는 선행 조건으로 "T1 이 표식 오류로 실패"를 단언했고, 그러면 mu12 가 #12 도 붉혔다 — acceptance.md §2.4 mu12 행은 #12 가 초록이어야 한다고 적었다. #12 에서 T1 단언을 빼(T1 의 실패는 #3·#13·#14 가 고정) SPEC 대로 맞춘 뒤 mu12 와 mu6 을 다시 돌렸다: mu12 → #3·#13·#14·#17 만, mu6 → #12 만. 위 표의 mu6·mu12 줄은 재실행 결과다.
2. **mu10 의 두 `prevTurnID` 비교(N3).** design.md D-1:67(귀속 규칙 (2))의 `requestTurn == c.prevTurnID` 를 골랐다. `noteTurnCompleted` 의 `id == c.prevTurnID`(design.md:68, G5)는 시험 행·변이를 두지 않는 방어라서 변이하지 않았다.
3. **mu20 은 한 줄 편집이 아니다(N11).** 내 구조에서는 닫힌 창과 비브로커 요청이 한 `if` 를 공유하므로 3줄 삽입으로 적용했다.
4. **#8 의 마감.** design.md 는 50ms 마감 컨텍스트라 했는데 150ms 를 썼다 — `-race` 부하에서도 가짜 서버가 `turn/start` 에 먼저 답해 `waitTurn` 의 `ctx.Done()` 분기(mu16 의 표적)가 마감을 맞게 하려는 선택이다. 마감 안에 답이 안 오면 `call()` 의 `ctx.Err()` 가 같은 `context.DeadlineExceeded` 를 내므로 단언은 통과하되 mu16 탐지력이 약해진다(관측: mu16 은 붉었다).
5. **블로킹 호출 목록(N5).** `waitTurn` 은 `hardenWaitTurn`(고루틴 + 5 s watchdog)으로 부른다. 연속 실패 시험·분류 시험의 드라이버 호출은 `hardenRunDriver`(5 s watchdog, stdin 끝 `/exit`)다.
6. **AC-MH-007 의 `-race`(N10).** `success_resets` 가 로그 줄을 읽으므로 이 시험은 `-race` 로 돌렸다.
7. **#7 `write_failure`** 는 실제 `pumpManagedStreamTurn` 을 닫힌 stdin 으로 돌려 실제 쓰기 오류를 만든다(기존 `managedClosedPipe` 재사용).
8. **`TestManagedFailedTurnLeavesClaimUntouched`(G1)** 는 수리 전 드라이버가 실패 턴 오류를 그대로 반환해도 통과하도록(`nil` 또는 표식 오류 허용) 썼다 — 불변 가드이고, 단언은 행 `claimed`·`acknowledged` 0·원래 claim token 으로 `ReadBody` 성공이다.

**미관측(Gaps)**: ① 운영자 stdout 에 로그가 쓰이지 않는다는 AC-MH-005 의 문장은 단언하지 않았다(`os.Stdout` 전역을 가로채지 않는다는 이 SPEC 의 규약; 로그 경로는 `managedLogf` 하나이고 `os.Stdout` 을 참조하지 않는다는 것은 소스 판독) ② 실제 codex 세션에서 거부 응답 뒤 모델 행동과 요청의 `serverName` 값(design.md D-1 Gap) ③ `X == prevTurnID` 중복 완료 프레임 방어(G5, 시험 행 없음) ④ 쓰기 막힘 상한(공시한 한계) ⑤ AC-MH-013 의 guardstate·template 인접 패키지와 AC-MH-009 의 전체 재측정은 M4 몫이라 이 마일스톤에서 돌리지 않았다 ⑥ mu1–mu5 는 M2 몫이고 이 마일스톤에서 다시 돌리지 않았다.

### M4 — 게이트와 증거 (card t1409, manager-develop, cycle_type=tdd)

귀속은 모두 `(this run, this tree, HEAD 1963ec376, 작업 트리 깨끗함)` 이다. 이 마일스톤은 프로덕션·시험 코드를 바꾸지 않았고 이 블록(progress.md)만 쓴다. 모든 `go test` 는 환경 정리 접두(`unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test …`)를 붙인 단일 호출이며 출력은 파일로 돌려 읽었다(스크래치 파일은 인용 대상이 아니다 — 필요한 줄만 아래에 옮겼고 파일 자체의 소실은 Residual-risk 에 적는다). 변이 확인 mu1–mu20 과 섭동 확인 P1 은 M2·M3 블록에 원문이 있고 이 마일스톤에서 다시 하지 않았다.

**항목 1 — 정적·빌드**

| 명령 | 관측 | exit |
|---|---|---|
| `gofmt -l internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go internal/cli/managed_hardening_test.go internal/cli/managed_codex_factory_test.go internal/config/defaults.go` | 출력 없음 | 0 |
| `go build ./...` | 출력 없음 | 0 |
| `go vet ./internal/cli ./internal/config` | 출력 없음 | 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음 | 0 |
| `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` | 출력 없음 | 0 |
| `golangci-lint run --allow-serial-runners --timeout=10m ./internal/cli/... ./internal/config/...` (`golangci-lint has version v2.1.6 built with go1.26.8`) | `0 issues.` | 0 |

**항목 2 — 스코프 회귀 (plan.md §C C-1)**

| 명령 | 관측 | exit |
|---|---|---|
| `go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 `grep -c '^Test'` 로 센 값(표 안 `\|` 는 원문 명령의 `|` 이다; 원문은 acceptance.md 머리말 블록) | `78`, 끝 줄 `ok  github.com/modu-ai/moai-adk/internal/cli  1.385s` (`[no tests to run]` 아님) | 0 |
| `go test ./internal/cli -run '^.*(Managed\|managed).*$' -count=1 -race -v` 로그를 `grep -c` 로 센 값 | 최상위 `--- PASS` 75, `--- SKIP` 3, `--- FAIL` 0, `DATA RACE` 0 (75 + 3 = 78); SKIP 은 `TestManagedCodexFakeAppServer`, `TestManagedCodexFactoryBrokerLive`, `TestManagedLoopbackChild`; 끝 줄 `PASS` · `ok  github.com/modu-ai/moai-adk/internal/cli  41.231s` | 0 |

**항목 3 — 인접 패키지**

| 명령 | 관측 | exit |
|---|---|---|
| `go test ./internal/config -count=1` | `ok  github.com/modu-ai/moai-adk/internal/config  9.791s` | 0 |
| `go test ./internal/guardstate -count=1` | 실패 시험은 정확히 하나: `--- FAIL: TestCensus_SetDifferenceEmptyBothDirections`, `census_test.go:80: disk\manifest: .github/workflows/workflow-parse-guard.yaml exists on disk with no manifest entry`, `census_test.go:89: declared 19 entries against 20 workflow files` | 1 |
| `go test ./internal/template -count=1` (슬롯 임대 안; 임대 획득 출력 `slot internal-template-suite acquired by 050d2b26-7816-450b-b2d8-dc197ebd878e until 2026-10-03T03:01:52Z`) | `ok  github.com/modu-ai/moai-adk/internal/template  232.918s` (스모크, AC 근거 아님; 기준 트리 414.986s 와 시간이 다른 것은 머신 부하 차이이며 비교 대상이 아니다) | 0 |
| `moai slot release --resource internal-template-suite` 후 `moai slot status --resource internal-template-suite` | `slot internal-template-suite released (was 050d2b26-7816-450b-b2d8-dc197ebd878e)` · `slot internal-template-suite: free` | 0 / 0 |

guardstate 의 한 건은 **사전 적색**이다: plan.md §C 가 기준 트리(HEAD `fe79bfa0e`)에서 같은 시험명과 같은 두 메시지(`19 entries against 20 workflow files`)를 이미 측정해 두었고, 이번 실행의 메시지가 그와 글자 그대로 같다. 이 카드는 `.github/workflows/` 와 매니페스트를 건드리지 않았다(`git diff --stat fe79bfa0e HEAD` 의 파일 8개에 둘 다 없음). 다른 guardstate 시험의 실패는 출력에 없다(비상세 모드에서 실패만 인쇄됨 — 실패한 시험은 위 하나).

**항목 4 — 커밋 그래프 점검 C1 (AC-MH-011)**

| 명령 | 관측 | exit |
|---|---|---|
| `git log --reverse --format=%h:%s fe79bfa0e..HEAD` | `f042a05a0:test(SPEC-FACTORY-MANAGED-HARDEN-001): M1 RED baseline for F3 F4 (card t1409)` · `0f852b527:fix(…): M2 answer server-originated codex requests (card t1409)` · `1963ec376:fix(…): M3 isolate turn failures with a ceiling (card t1409)` | 0 |
| `git merge-base --is-ancestor f042a05a0 0f852b527` / 역방향 | (출력 없음) / (출력 없음) | 0 / 1 |
| `git merge-base --is-ancestor f042a05a0 1963ec376` / 역방향 | (출력 없음) / (출력 없음) | 0 / 1 |
| `git show --stat --format=%h f042a05a0` | 5개 파일: `managed_hardening_test.go` +694(재현 시험 6개 전부 여기), `managed_codex_factory_test.go` 27행 변경, `red-baseline.md` +157, `progress.md` 25행 변경, `spec.md` 2행 변경; 프로덕션 `.go` 없음 | 0 |
| `git ls-files .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` | 그 경로 한 줄 | 0 |
| `git check-ignore -v .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` | 출력 없음 | 1 |
| 양성 대조: `git check-ignore -v .moai/reports/t1409/red-baseline.md` | `.gitignore:235:.moai/reports/*	.moai/reports/t1409/red-baseline.md` | 0 |

**항목 5 — 경계·불변 grep (acceptance.md §5, AC-MH-009·012)**

| 명령 | 관측 | exit |
|---|---|---|
| `grep -rn 'syscall\.' internal/cli/managed_*.go` | 출력 없음 | 1 |
| 양성 대조 `grep -n 'syscall\.' internal/cli/mcp_server.go` | `123:	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` | 0 |
| `grep -rn 'AskUserQuestion' internal/cli/managed_*.go` (acceptance.md §5 범위; 이 형태는 비테스트·테스트 파일 모두 0행) | 출력 없음 | 1 |
| `grep -rnE 'merge-window\|Decider\|T29b\|T29c\|handover' internal/cli/managed_*.go` (원문 `\|` 는 `|`) | 출력 없음 | 1 |
| `git diff --stat fe79bfa0e HEAD -- internal/factorymsg .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001` | 출력 없음 | 0 |
| `git diff --stat develop...HEAD -- .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001 internal/factorymsg/store.go` (AC-MH-012 원문) | 출력 없음 | 0 |
| 양성 대조 `git diff --stat fe79bfa0e HEAD -- internal/cli` | 4개 파일, `1798 insertions(+), 20 deletions(-)` | 0 |
| `git diff -U0 fe79bfa0e HEAD -- internal/cli/managed_factory_session.go` (`managedSession` 인터페이스 `:75` 와 `driveManagedFactorySession` 시그니처 `:308` 불변) | 헝크 네 개: `errManagedTurnFailed` 선언(`:59` 뒤 추가), `pumpManagedStreamTurn` 의 `fmt.Errorf("%w: %s", …)`, 드라이버 본문 내부 `consecutiveFailures` 변수·분기·리셋 — 인터페이스 선언과 시그니처 줄은 어느 헝크에도 없다 | 0 |

주의: `grep -rl 'AskUserQuestion' internal/cli --include='*.go' --exclude='*_test.go'` (패키지 전체, 비테스트)는 37개 파일을 냈다(exit 0). 이는 이 카드와 무관한 기존 문자열 참조이며 acceptance.md §5 의 범위는 `managed_*.go` 이므로 위 표의 `managed_*.go` 형태를 판정으로 인용한다(패키지 전체 0행을 주장하지 않는다).

**항목 6 — M3 커밋 메시지 정정 (`--amend` 금지에 따라 이 추적 기록이 커밋 그래프의 주장을 대체한다)**

`1963ec376` 의 본문은 한 문장에 부정확한 두 구절을 담았다. 원문 인용: "Each new case was observed RED before the fix, and mutants mu6-mu20 plus the ceiling perturbation P1 were applied one at a time, each going red on the named cases only."

- (a) "Each new case was observed RED before the fix" 는 틀렸다. 분류 #4–#11, #15, #16, `TestManagedBrokerNameMatchesApprovalArgs`, `TestManagedFailedTurnLeavesClaimUntouched` 는 수리 전에 이미 초록이었고, 변이(채택) 또는 불변 가드(G1·G2·G3)로 채택됐다. 정확한 기록은 위 M3 블록의 "RED 에서 이미 초록이었던 것" 문단과 mu 표다.
- (b) "each going red on the named cases only" 도 틀렸다. mu9, mu12, mu13 은 SPEC(acceptance.md §2.4)이 허용한 범위에서 지목 케이스 외의 케이스도 붉혔다(mu9 는 M2 `…CompletionEventCarriesBrokerVerdict/other_server_request` 를 더 붉혔고, mu12 는 M1 `…FailsTurn` 3개 등을, mu13 은 M1 두 시험과 `at_ceiling_returns`·`success_resets` 를 더 붉혔다). 정확한 기록은 위 M3 mu 표의 "붉은 케이스(관측)" 열이다.

이 정정은 `1963ec376` 의 변이·RED 관측 사실 자체를 바꾸지 않는다. 그 문장이 담은 일반화만 위 두 갈래로 좁힌다.

**Claim**: M4 정적·빌드·스코프·인접 패키지·커밋 그래프·경계 게이트가 HEAD `1963ec376` 에서 통과하며 실패는 사전 적색 한 건(guardstate)뿐이다. **Evidence**: 위 표의 명령·출력·exit. **Baseline-attribution**: `(this run, this tree, HEAD 1963ec376)`. **Gaps**: ① 라이브 Codex(실제 세션의 거부 응답 뒤 모델 행동, 요청 `serverName` 값 — design.md D-1 Gap) 미관측 ② 로그가 운영자 stdout 에 쓰이지 않는다는 문장은 단언 시험이 없고 소스 판독(`managedLogf` 는 `os.Stdout` 을 참조하지 않음)뿐 ③ `X == prevTurnID` 중복 완료 프레임 방어(G5)는 시험 행이 없음 ④ 쓰기 막힘 상한(공시한 한계)과 `id: null` 프레임(공시한 한계) 미해결 ⑤ 설치된 `moai` 바이너리가 이 트리보다 뒤처져 있어 `moai spec lint` 출력은 증거로 쓰지 않았다(지연 빌드의 증거일 뿐) ⑥ acceptance.md §1 의 AC 매트릭스 전체를 M4 에서 한꺼번에 다시 돌리지는 않았다 — AC-MH-001..008 의 GREEN 은 M2·M3 블록의 원문과 위 스코프 회귀(75 PASS, 0 FAIL)로 덮이며, AC-MH-010 은 sync 단계 몫이다 ⑦ `internal/template` 스위트는 AC 근거가 아닌 스모크로만 돌렸다(plan.md §F M4). **Residual-risk**: 스크래치 로그 파일(`…/scratchpad/m4_*.txt`)은 추적되지 않아 이 기록의 인용 줄만 남는다; 사전 적색은 develop 쪽에서 수리되기 전까지 인접 패키지 전체 통과 주장을 막는다; 연속 실패 상한 3 은 UNMEASURED 그대로다.

### Repair R1 (sync-audit F1) — 레거시 승인 요청의 거부 응답 형태 (card t1409, manager-develop, cycle_type=tdd)

**잘못된 값의 출처.** 레거시 승인 요청 두 종(`applyPatchApproval`, `execCommandApproval`)에 `{"decision":"denied"}` 를 답하게 한 값은 plan 단계 design.md D-1 응답 정책 표에서 왔다. 구현이 그 표를 그대로 옮겼고, plan 감사 3회는 enum 이름(denied)만 맞춰 보고 변형(variant)의 모양을 대조하지 않아 놓쳤다. codex 0.160.0 스키마에서 `ReviewDecision` 의 `denied` 변형은 문자열이 아니라 필수 키 `denied` 를 가진 객체이고 그 값은 필수 문자열 `rejection` 을 가진 객체다. 유효한 거부는 `{"decision":{"denied":{"rejection":"<문자열>"}}}` 이며, 에이전트가 실행하지 않되 세션은 계속 이어 가게 하는 의미라 이 SPEC 이 원하는 뜻과 같다. 문자열 변형 `abort` 는 턴을 끊으므로 쓰지 않는다. 승인 계열은 어떤 경로로도 답하지 않는다.

**변경 범위.** 운영 코드 `internal/cli/managed_codex_factory.go`(상수 `managedLegacyApprovalRejection` 하나와 객체 빌더 `managedLegacyDeniedResult`, 정책 표의 두 항목, 로그 토큰 `denied` 는 그대로), 시험 `internal/cli/managed_hardening_test.go`, 스키마 사본 `internal/cli/testdata/codex-0.160.0/`. `.moai/specs/**` 의 다른 파일, CHANGELOG, 문서는 건드리지 않았다.

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 1 RED (정책 시험) | `go test -race ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1 -v` (환경 정리 접두 포함, 상수만 먼저 추가하고 표는 그대로 둔 트리) | `--- FAIL …/legacy_apply_patch_approval`, `--- FAIL …/legacy_exec_command_approval`, 사유 `result {"decision":"denied"} is not the {decision:{denied:{rejection}}} object`; 나머지 9개 하위 `--- PASS` | 1 |
| 2 RED (새 스키마 가드, 수리 전) | `go test -race ./internal/cli -run '^TestManagedServerRequestPolicyMatchesCodexSchema$' -count=1 -v` | `answer {"decision":"denied"} to applyPatchApproval violates ApplyPatchApprovalResponse.json`, `… to execCommandApproval violates ExecCommandApprovalResponse.json`; 두 하위만 `--- FAIL`, 나머지 9개 종류와 `validator_catches_the_bare_denied_string` 은 `--- PASS` | 1 |
| 3 GREEN | `go test -race ./internal/cli -run '^(TestManagedCodexServerRequestPolicy\|TestManagedServerRequestPolicyMatchesCodexSchema)$' -count=1 -v` | `--- PASS` 25, `--- FAIL` 0, `ok  github.com/modu-ai/moai-adk/internal/cli  4.271s` | 0 |
| 4a 변이: `applyPatchApproval` 을 다시 문자열 `"denied"` 로 | 같은 명령 | 정책 시험 `…/legacy_apply_patch_approval` 과 스키마 가드 `…/legacy_apply_patch_approval` 둘 다 `--- FAIL` (다른 하위는 초록) | 1 |
| 4b 변이: `execCommandApproval` 을 `"abort"` 로 | 같은 명령 | 정책 시험 `…/legacy_exec_command_approval` 만 `--- FAIL`. 스키마 가드는 초록이었다(`abort` 는 스키마상 유효하고 정책이 틀린 경우이므로 정확한 응답 단언이 잡는다 — 지시의 예상 "둘 다 붉음" 과 다른 관측이며, 가드가 정책 값을 대신 지키지 않는다는 이 시험의 설계 그대로다) | 1 |
| 4c 변이: `item/commandExecution/requestApproval` 의 `decline` 을 `accept` 로 | 같은 명령 | 정책 시험 `…/command_execution_approval` 만 `--- FAIL`(mu1 의 동작 유지). 스키마 가드는 초록(`accept` 도 스키마상 유효) | 1 |
| 변이 복원 확인 | `git diff --stat` | 변이마다 복원 직후 `managed_codex_factory.go 18 +++-`, `managed_hardening_test.go 180 +++…-`(변이 전과 동일) | — |
| 5 정적 | `gofmt -l` 두 파일 / `go build ./...` / `go vet ./internal/cli ./internal/config` / `GOOS=windows GOARCH=amd64 go build ./...` / `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` / `golangci-lint run --allow-serial-runners ./internal/cli/...` (v2.1.6) | 출력 없음 ×5 / `0 issues.` | 모두 0 |
| 선택 수 | `go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 `grep -c '^Test'` 로 센 값 | `79` (78 + 새 가드 1) | 0 |
| managed 전체 | `go test ./internal/cli -run '^.*(Managed\|managed).*$' -count=1 -race -v` | `--- PASS` 218(하위 포함), `--- SKIP` 3(`TestManagedCodexFakeAppServer`, `TestManagedCodexFactoryBrokerLive`, `TestManagedLoopbackChild`), `--- FAIL` 0, `DATA RACE` 0, `ok  github.com/modu-ai/moai-adk/internal/cli  47.055s` | 0 |
| config | `go test ./internal/config -count=1` | `ok  github.com/modu-ai/moai-adk/internal/config  11.219s` | 0 |
| 경계 | `grep -c 'syscall\.' internal/cli/managed_*.go` / `grep -c 'AskUserQuestion' internal/cli/managed_*.go` (비시험 파일 2개) | 모두 `:0`; 양성 대조: `internal/cli/codex_direct_posix.go` 에 `syscall.` 존재, `internal/cli/mcp_server.go` 에 `AskUserQuestion` 2건 | — |
| 스키마 사본 재생성 | `codex app-server generate-json-schema --out <트리 밖 폴더>` (codex-cli 0.160.0) 후 `cmp` | 사본 8개 모두 바이트 동일 | 0 |

모든 `go test` 는 환경 정리 접두(`unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test …`)를 붙인 단일 호출이다. 귀속은 모두 `(this run, this tree, HEAD f4f47dc51)` 이다. 이 블록과 같은 커밋에서 HEAD 가 움직이므로 위 HEAD 는 수리 직전 부모다.

**검증기 경로.** `go.mod` 에 이미 직접 의존으로 있는 `github.com/santhosh-tekuri/jsonschema/v6`(운영 코드 `internal/codextools/registry.go` 도 사용)를 썼다. 새 의존도 `go.mod`/`go.sum` 변경도 없다. 사본 스키마는 draft-07 이고 내부 `#/definitions` 참조만 가져 로더가 필요 없다. 시험은 `codex` 바이너리도 네트워크도 쓰지 않는다. 사본: `ApplyPatchApprovalResponse.json`(4278 B), `ExecCommandApprovalResponse.json`(4279 B), `CommandExecutionRequestApprovalResponse.json`(3202 B), `FileChangeRequestApprovalResponse.json`(1158 B), `PermissionsRequestApprovalResponse.json`(7068 B), `McpServerElicitationRequestResponse.json`(741 B), `DynamicToolCallResponse.json`(2010 B), `JSONRPCError.json`(893 B), 그리고 생성 명령과 재생성 조건을 적은 `README.md`.

**다른 정책 항목.** 새 스키마 가드에서 레거시 두 항목 외의 정책 항목은 하나도 붉지 않았다(결과 응답 6종은 각 스키마를, 오류 응답 3종과 미등록 메서드 폴백은 `JSONRPCError.json` 을 통과). 따라서 이 수리는 SPEC 정책 표의 다른 줄을 바꾸지 않는다. 가드는 표 키와 `hardenPolicyRows`·결과 스키마 사상의 일치도 단언한다(표에 종류가 늘면 매핑이 없어 붉어진다).

**Gaps.** ① 라이브 Codex 에서 레거시 승인 요청에 새 거부 객체를 답했을 때의 모델 행동은 관측하지 않았다(스키마 유효성만 관측). ② 오류 프레임 검증은 와이어에서 읽은 `id`·`code`·`message` 로 프레임을 재조립해 대조한 것이다(`data` 키는 소유자가 쓰지 않는다). ③ `abort`·`accept` 처럼 스키마상 유효한 오답은 스키마 가드가 아니라 정확한 응답 단언(AC-MH-001)만 잡는다 — 4b·4c 로 관측. ④ design.md D-1 표, acceptance.md, spec.md 의 `denied` 서술은 이 위임이 고치지 않았다(후속 위임 몫). **Residual-risk**: 사본은 0.160.0 기준이라 최소 지원 codex 버전이 올라가면 재생성하지 않는 한 가드는 옛 스키마를 계속 신뢰한다(README 에 명시). 거부 문구 상수의 영어 문장은 서버가 모델에 그대로 전달할 수 있다 — 식별자·경로·비밀은 넣지 않았다. 사전 적색(guardstate)은 이 수리 범위 밖이라 재측정하지 않았다.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-03
tree: .moai/worktrees/t1409
branch: WT-managed-session-hardening
head_at_signal: 1963ec376 (측정한 트리; 이 신호를 담는 M4 증거 커밋은 자기 SHA 를 적을 수 없으므로 `git log -1` 로 읽는다)
cycle_type: tdd
milestones: M1 `f042a05a0` (RED 기준선) · M2 `0f852b527` (F3 서버 요청 응답) · M3 `1963ec376` (F4 턴 격리와 상한) · M4 (이 신호를 담은 증거 커밋)
tests: managed 스코프 `-list` 78, `-race -v` 로 PASS 75 · SKIP 3 · FAIL 0 · DATA RACE 0 (SKIP 3 건은 기존: `TestManagedCodexFakeAppServer`, `TestManagedCodexFactoryBrokerLive`, `TestManagedLoopbackChild`)
mutation: mu1–mu20 과 P1 은 M2·M3 블록에 원문(M4 에서 재실행 없음)
gates: gofmt·build·vet·windows build/vet·golangci-lint v2.1.6 모두 exit 0 (§E.2 M4 항목 1)
adjacent: `./internal/config` ok; `./internal/template` ok 232.918s (임대 안에서 실행·반납); `./internal/guardstate` 는 **사전 적색 1건** `TestCensus_SetDifferenceEmptyBothDirections`(`declared 19 entries against 20 workflow files`, 기준 트리 `fe79bfa0e` 에서 plan.md §C 가 먼저 측정) — 이 카드의 변경이 아니다
commit_graph: C1 통과 — RED `f042a05a0` 는 `0f852b527`·`1963ec376` 의 조상이고 역방향은 아니다(§E.2 M4 항목 4)
correction_note: M3 커밋 `1963ec376` 메시지의 두 구절 정정은 §E.2 M4 항목 6 이 대체한다
open_gaps: 라이브 Codex 미관측 · F5(시그널 처리·`Start`/`Close` 수명주기)는 카드 t1459 · 문서·CHANGELOG(AC-MH-010)는 sync 단계 · 막힌 쓰기 상한과 `id: null` 프레임은 공시한 한계
next: sync 단계(manager-docs)에서 AC-MH-010 문서·CHANGELOG 와 sync 감사

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-03
superseded_sync_commit_sha: 13eab5d4c (첫 종결의 sync 커밋. 재종결이 대체했다. 도구가 현재 종결의 `sync_commit_sha` 를 읽도록 키 이름을 `sync_commit_sha` 에서 바꿨고, 값은 이 줄을 쓸 때 `pending-backfill` 이던 것을 실제 SHA 로 채웠다)
tree: .moai/worktrees/t1409
branch: WT-managed-session-hardening
head_at_signal: 3ff4392d4 (측정한 트리; 이 신호를 담는 sync 커밋은 자기 SHA 를 적을 수 없으므로 `sync_commit_sha` 는 정규 플레이스홀더 `pending-backfill` 이고 실제 SHA 는 뒤따르는 커밋에서 backfill 한다. 비워 두지 않았다) (이후 backfill 됨: 위 `superseded_sync_commit_sha` 줄과 재종결 블록의 `sync_commit_sha` 줄 참조)
owner: manager-docs (sync-phase)
ac_source: .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/acceptance.md (tier M, 결과 `resolved`, 비어 있지 않음)
docs_changed: `.moai/docs/factory-managed-session.md` · `CHANGELOG.md` · `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/spec.md`(frontmatter `status:` 한 줄뿐) · 이 파일의 §E.4
docs_not_changed: 템플릿 미러 없음(`internal/template/templates/.moai/docs/factory-managed-session.md` 부재 확인), plan.md · acceptance.md · design.md · red-baseline.md · 부모 SPEC 디렉터리 · `.go` 파일 전부
changelog_entry_position: `[Unreleased]` → `### Added` 의 첫 항목 하나(SPEC-AUDIT-MODEL-CONVERGE-001 엔트리 바로 위). 부모 SPEC-FACTORY-MANAGED-SESSION-001 엔트리의 "위 (3)의 세 항목은 …" 문장 하나를 정정(전체 SPEC id 를 쓰지 않고 `FACTORY-MANAGED-HARDEN-001` 로 가리켜 이 SPEC id 의 줄 수를 1 로 유지)
frontmatter_status_transitions: spec.md `in-progress → implemented → completed` 를 이 단일 sync 커밋에 병합(frontmatter 에는 `status: completed` 로 기록), `updated: 2026-10-03` 은 이미 같은 날짜라 값 변경 없음. 스키마(§Artifact Statelessness)상 plan.md · acceptance.md · design.md · progress.md 는 `status:` 필드를 갖지 않으므로 바꿀 것이 없다(측정: acceptance.md frontmatter 에 `status:` 없음, progress.md 는 frontmatter 없음)
b12_self_test_a: 사전 `grep -c 'SPEC-FACTORY-MANAGED-HARDEN-001' CHANGELOG.md` → `0`(기준 트리 `git grep -c … fe79bfa0e` 도 출력 없음·exit 1); 사후 → `1`
b12_self_test_b: B12 카운터(acceptance.md, tier M) → stdout `13`, stderr `live=13 excluded=0 ambiguous=0`, exit 0; REQ `grep -oE '^- \*\*REQ-MH-[0-9]{3}\*\*' spec.md` 고유 12개(001..012); CHANGELOG 엔트리는 12 / 13 으로 적었다
b12_self_test_c: CHANGELOG 가 인용한 경로 전부 `ls` 로 존재 확인(`internal/cli/managed_codex_factory.go`, `managed_factory_session.go`, `managed_hardening_test.go`, `managed_codex_factory_test.go`, `internal/config/defaults.go`, `.moai/docs/factory-managed-session.md`, `progress.md`, `red-baseline.md`; 템플릿 미러만 `No such file or directory` — 기대대로)
canary_compliance_check: 해당 없음 — 이 SPEC 은 자기 sync 가 시험해야 하는 전방 정책을 정의하지 않는다

귀속은 모두 `(this run, this tree, HEAD 3ff4392d4 위의 sync 작업 트리, 커밋 전)` 이다. 이 sync 는 프로덕션·시험 코드를 바꾸지 않았고 `go test` 는 선택 수 확인(`-list`) 한 번만 돌렸다. 스크래치 파일은 인용 대상이 아니며 필요한 줄만 아래에 옮겼다.

### 증거

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 사전 상태 | `git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse --short HEAD` / `git status --short` | `…/.moai/worktrees/t1409` / `WT-managed-session-hardening` / `3ff4392d4` / 출력 없음 | 0 |
| 선택 수 | `unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 파일로 돌려 `grep -c '^Test'` (원문 명령은 acceptance.md §1.1 AC-MH-013 블록; 표 안의 `\|` 는 원문의 `|` 이다) | `78`, 끝 줄 `ok  github.com/modu-ai/moai-adk/internal/cli  1.412s`, `no tests to run` 0건 | 0 |
| 커밋 전 변경 목록 | `git status --short` | ` M .moai/docs/factory-managed-session.md` · ` M .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/spec.md` · ` M CHANGELOG.md` (§E.4 편집 뒤 `progress.md` 가 더해져 네 경로) | 0 |
| AC-MH-010 문서 앵커(부재) | `grep -c "턴 하나가 실패하면 세션 전체가 끝난다"` / `"서버가 먼저 보내는 승인·질문 요청에는 답하지 않는다"` / `"후속 카드 t1409 대상"` 를 `.moai/docs/factory-managed-session.md` 에 | `0` / `0` / `0`, 세 개 모두 exit 1 | 1 |
| 부재 앵커의 양성 대조 | `git grep -c <같은 세 문구> fe79bfa0e -- .moai/docs/factory-managed-session.md` | 각 `fe79bfa0e:.moai/docs/factory-managed-session.md:1` (기준 트리에서 세 문구가 실제로 줄을 냈다) | 0 |
| AC-MH-010 문서 앵커(존재) | 같은 파일에 `grep -c` | `런처는 시그널을 처리하지 않는다` 1 · `t1459` 1 · ``관리 계층은 .syscall. 을 쓰지 않는다`` 1 · `TTL까지 재배달` 1 · `턴 타임아웃은 세션을 끝낸다` 1 · `거부된 elicitation` 1 · `쓰기 데드라인` 1 · `실제 codex 세션에서는 관측하지 않았다` 1 | 0 |
| AC-MH-010 CHANGELOG | `grep -c` 를 `CHANGELOG.md` 에 | `SPEC-FACTORY-MANAGED-HARDEN-001` → `1` · `t1459` → `2`(새 엔트리와 부모 정정 문장, 줄 수) · `후속 카드 t1409 대상` → `0`(exit 1) | 0 / 0 / 1 |
| CHANGELOG 부재의 양성 대조 | `git grep -c "후속 카드 t1409 대상" fe79bfa0e -- .moai/docs/factory-managed-session.md CHANGELOG.md` / `git grep -c "SPEC-FACTORY-MANAGED-HARDEN-001" fe79bfa0e -- CHANGELOG.md` | `fe79bfa0e:.moai/docs/factory-managed-session.md:1` · `fe79bfa0e:CHANGELOG.md:1` / 출력 없음(0) | 0 / 1 |
| 부모·브로커·`internal` 불변 | `git diff --stat fe79bfa0e -- internal/factorymsg .moai/specs/SPEC-FACTORY-MANAGED-SESSION-001` / `git diff --stat 3ff4392d4 -- internal` | 출력 없음 / 출력 없음 | 0 / 0 |
| 설치된 `moai` 의 지연 | `moai version` / `git merge-base --is-ancestor 45600e4ee 3ff4392d4` | `v3.2.0-rc.26   archive/t1401-293-g45600e4ee   built 2026-10-02T14:42:32Z` / 출력 없음 | 0 / 0 |
| `moai spec lint SPEC-FACTORY-MANAGED-HARDEN-001` | (위 지연 빌드로 실행) | `✓ No findings — all SPEC documents are valid` | 0 |

`moai spec lint` 의 통과는 **지연 빌드의 증거**일 뿐이다: 설치된 빌드의 커밋 `45600e4ee` 가 이 트리 HEAD 의 엄격한 조상이라(`is-ancestor` exit 0) 이 트리가 담은 점검을 다 갖췄다고 할 수 없다. 이 트리 빌드의 판정이 아니다.

### MX 태그 검증 (sync 하위 단계; sync 에서는 `.go` 를 고칠 수 없어 태그를 더하거나 바꾸지 않고 소견만 적는다)

`.moai/config/sections/mx.yaml` 의 `fan_in_anchor` 는 3, `anchor_per_file` 3, `warn_per_file` 5, `note_per_file` 10 이다. 대상은 `internal/cli/managed_codex_factory.go` 와 `internal/cli/managed_factory_session.go` 이다.

| 파일 | 현재 태그(읽기 확인) | 소견 |
|---|---|---|
| `managed_codex_factory.go` | `@MX:NOTE`(응답 정책표, 이 카드가 더함) · `@MX:WARN`+`@MX:REASON`(`read()`, 기존) | **`managedCodexAppClient.write` 는 `@MX:ANCHOR` 후보다.** 비테스트 호출 지점이 3곳이라 fan_in 3 에 닿는다(`answerServerRequest` 의 `c.write`, `call` 의 `c.write`, `Start` 의 `s.client.write`; grep 확인). 모든 연결 쓰기가 이 한 길을 지난다는 것(gorilla/websocket 은 동시 쓰기를 허용하지 않고 읽기 고루틴과 호출 루프가 둘이라는 점)이 계약이다. 태그 없음. 태그를 달면 파일당 ANCHOR 한도 3 안이다 |
| `managed_codex_factory.go` | — | `answerServerRequest`(호출 1곳), `armTurn`(1곳), `noteTurnStarted`·`noteTurnCompleted`·`noteElicitation`(각 1곳)은 fan_in 1 이라 ANCHOR 대상이 아니다. `noteElicitation` 의 귀속 규칙(직전 턴 id 와 불일치 id 는 세지 않음)은 시험으로만 고정된 비자명한 규칙이라 `@MX:NOTE` 후보다(선택). 새 `go` 문(고루틴)은 이 카드가 더하지 않았다 — `go s.client.read()`(기존)는 이미 `@MX:WARN` 아래에 있다 |
| `managed_factory_session.go` | `@MX:NOTE`·`@MX:SPEC`(기존, 이 카드와 무관) | **`errManagedTurnFailed` 는 `@MX:NOTE` 후보다** — 이 오류를 세 곳에서만 감싸고 드라이버는 `errors.Is` 로만 격리한다는 계약이 변수 주석에 있으나 태그가 없다. 호출이 아니라 변수라 fan_in 규칙의 대상은 아니다(참조 4곳). `managedLogf`(비테스트 호출 2곳)는 fan_in 3 미만 |

판정: 새 코드에서 **ANCHOR 한 건(`write`)** 과 NOTE 두세 건이 프로토콜에 비추어 권고된다. 이 sync 는 `.go` 편집이 금지돼 있어 적용하지 않았고, 리더가 후속 커밋이나 `/moai mx` 에서 결정한다. 이 소견은 `grep` 으로 센 호출 지점 수와 `mx.yaml` 임계값을 읽은 것이며 `moai mx` 스캐너를 돌려 얻은 것이 아니다.

### SPEC 과 코드의 불일치(보고만 하고 고치지 않음)

1. **응답 정책표의 종류 수.** 지시문은 "정책표에 11종"이라 했으나 코드의 `managedServerRequestPolicies` 맵은 이름 붙은 **10종**이고 표에 없는 method 를 `managedUnknownRequestPolicy`(`-32601`)가 따로 받는다. 시험 하위 케이스가 11개인 것은 10종 + 미지 method 1종이다(acceptance.md AC-MH-001 도 "10종 + 미지 method 하나"). 문서와 CHANGELOG 는 코드대로 적었다.
2. **#8 마감 시간.** design.md 는 50ms 마감 컨텍스트라 했으나 구현 시험은 150ms 를 썼다(§E.2 M3 발견 4 에 이미 기록).
3. **상태 전이 대상 산출물 수.** 에이전트 본문은 `spec.md`·`plan.md`·`acceptance.md`·`progress.md` 네 곳에 상태를 적는다고 하나, `spec-frontmatter-schema.md` § Artifact Statelessness 는 `plan.md`·`acceptance.md`·`design.md`·`research.md` 에 `status:` 를 두지 말라고 한다. 이 SPEC 에서는 `status:` 를 가진 산출물이 `spec.md` 하나뿐이라 그것만 바꿨다.

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim**: AC-MH-010(운영 문서와 CHANGELOG)이 충족되고, F3·F4 는 해결로, 시그널 처리 공백(F5)은 카드 t1459 를 가리키는 미해결 한계로 적혔으며, 남은 한계가 고정 앵커 문자열과 함께 문서에 있다. REQ 12·AC 13 의 개수가 CHANGELOG 와 일치한다. **Evidence**: 위 표의 명령·관측·exit. **Baseline-attribution**: `(this run, this tree, HEAD 3ff4392d4 위의 sync 작업 트리, 커밋 전)`. 커밋 뒤 점검(`git diff --stat 3ff4392d4 HEAD`, spec.md 의 diff, 부모 SPEC 불변, 커밋 본문 트레일러)은 이 신호를 담은 커밋 자신에 대한 것이라 커밋 전에는 얻을 수 없으며 sync 보고서에 따로 적는다.

**Gaps**: ① 라이브 Codex 미관측 — 거부 응답 뒤 모델 행동, elicitation 요청의 `serverName` 값, 그리고 문서·CHANGELOG 가 한계로 적은 "거부된 elicitation 이 영수증을 막는" 시나리오가 실제로 일어나는지 ② `X == prevTurnID` 중복 완료 프레임 방어(G5)는 시험 행이 없음 ③ 막힌 쓰기 상한과 `id: null` 프레임은 공시한 한계로 남음(소스 판독) ④ 로그가 운영자 stdout 에 쓰이지 않는다는 문장은 단언 시험이 없음(§E.2 M3 Gaps ①) ⑤ 이 sync 는 `go test` 를 선택 수 확인 외에 돌리지 않았다 — 75 PASS·3 SKIP·0 FAIL 은 §E.2 M4 항목 2 의 manager-develop 기록을 인용한 것이지 이 sync 의 재측정이 아니다 ⑥ `moai spec lint` 는 지연 빌드라 이 트리의 판정이 아니다 ⑦ MX 소견은 스캐너 실행이 아니라 호출 지점 grep 과 임계값 판독이다 ⑧ 커밋 뒤 점검 결과는 이 파일에 담기지 않는다(위 귀속 문단) ⑨ 문서의 서술(예: "연속 N번 실패하면 마지막 오류를 반환해 세션을 끝낸다", 한계 문장들)은 `internal/cli/managed_codex_factory.go`·`managed_factory_session.go`·`defaults.go` 를 이 sync 가 읽은 것에 근거하고 라이브 세션으로 확인한 것이 아니다.

**Residual-risk**: ① 계획 감사 부채 N3–N12 는 optional 로 열린 채 닫히지 않았다(plan-audit-iter6.md: PASS-WITH-DEBT 0.81, Tier M 기준 0.80) ② `internal/guardstate` 의 사전 적색 `TestCensus_SetDifferenceEmptyBothDirections`(`declared 19 entries against 20 workflow files`)는 기준 트리 `fe79bfa0e` 에서 이미 있었고 이 카드와 무관하지만 develop 쪽 수리 전에는 인접 패키지 전체 통과 주장을 막는다 ③ 연속 실패 상한 3 은 UNMEASURED 관례 값이다 ④ 실제 codex 관측이 없어 elicitation 귀속 규칙(`serverName` 이 `moai` 가 아니면 세지 않음)이 실제 값과 어긋나면 격리 판정이 의도와 달라질 수 있다 ⑤ 독 메시지와 거부된 elicitation 의 TTL 까지 재배달은 이 SPEC 이 닫지 않았다 ⑥ M3 커밋 `1963ec376` 본문의 두 구절은 `--amend` 금지로 그대로 남고 §E.2 M4 항목 6 이 정정한다 ⑦ 스크래치 증거 파일은 추적되지 않아 이 신호의 인용 줄만 남는다 ⑧ t1408(TUI 부착)·t1410(F8·F9·F13)·t1459(F5)는 후속 카드로 남아 있다.

### §E.4 재종결 블록 — sync-audit F1 수리 뒤의 두 번째 종결 (manager-docs, card t1409)

sync_status: audit-ready
sync_complete_at: 2026-10-03
sync_commit_sha: c9087a7ff
superseded_first_close: 13eab5d4c (위 §E.4 블록이 신호를 담은 첫 sync 커밋이며 독립 sync 감사 `.moai/reports/t1409/sync-audit.md` 가 FAIL(F1)을 냈다. 이 재종결이 대체한다. 위 블록의 `sync_commit_sha: pending-backfill` 은 backfill 되지 않은 채 남은 첫 종결의 값이고, 현재 종결의 값은 이 블록의 위 줄이다)
tree: .moai/worktrees/t1409
branch: WT-managed-session-hardening
head_at_signal: 3b12e75a1 (측정한 트리; 이 신호를 담는 재종결 커밋은 자기 SHA 를 적을 수 없으므로 `sync_commit_sha` 는 정규 플레이스홀더 `pending-backfill` 이고 실제 SHA 는 뒤따르는 progress.md 전용 커밋에서 backfill 한다. 비워 두지 않았다) (이후 backfill 됨: 위 `sync_commit_sha` 줄 참조)
owner: manager-docs (sync-phase, 재종결)
card_commits: M1 `f042a05a0` · M2 `0f852b527` · M3 `1963ec376` · M4 `3ff4392d4` · 첫 sync `13eab5d4c`(대체됨) · MX 주석 `f4f47dc51` · 수리 R1 `2796a98d3` · 개정 `3b12e75a1` · 재종결(이 블록을 담은 커밋) · backfill(그 다음 커밋)
ac_source: .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/acceptance.md (tier M, 결과 `resolved`, 비어 있지 않음)
docs_changed: `.moai/docs/factory-managed-session.md`(서버 요청 응답 행의 레거시 거부 객체·스키마 가드 문장, 실제 codex 미관측 한계에 레거시 거부 객체 절, 로그 줄 위조 한계 항목, 확인 방법에 스키마 가드 시험) · `CHANGELOG.md`(기존 단일 엔트리를 제자리에서 수정) · `spec.md`(frontmatter `status:` 한 줄뿐) · 이 파일의 §E.4 재종결 블록
docs_not_changed: 템플릿 미러 없음(`internal/template/templates/.moai/docs/factory-managed-session.md` 부재 확인), plan.md · acceptance.md · design.md · red-baseline.md · 부모 SPEC 디렉터리 · `.go` 와 `internal/` 아래 전부
changelog_entry_position: 새 엔트리가 아니라 기존 단일 엔트리(`[Unreleased]` → `### Added` 의 첫 항목)를 제자리에서 고쳤다. 커밋 목록·선택 수 79·레거시 승인 거부 객체·스키마 가드·감사 F1 수리 경위·한계 ②(레거시 거부 객체는 스키마 유효성만 관측)와 ⑧(로그 줄 위조, F2)을 반영했다
frontmatter_status_transitions: spec.md `in-progress → completed`(개정 종결). 개정 커밋 `3b12e75a1` 이 `status: in-progress` 로 돌렸고 이 재종결이 `status: completed` 로 되돌린다. `updated: 2026-10-03` 은 이미 같은 날짜라 값 변경 없음. `amendment_of` 와 나머지 frontmatter 는 건드리지 않았다
b12_self_test_a: `grep -c 'SPEC-FACTORY-MANAGED-HARDEN-001' CHANGELOG.md` 편집 전 1(줄 29 하나), 편집 후 1 — 제자리 수정이라 새 줄이 없다
b12_self_test_b: B12 카운터(acceptance.md, tier M) → stdout `13`, stderr `live=13 excluded=0 ambiguous=0`, exit 0; `grep -oE '^- \*\*REQ-MH-[0-9]{3}\*\*' spec.md | sort -u | wc -l` → 12; CHANGELOG 엔트리는 12 / 13 으로 유지
b12_self_test_c: CHANGELOG 가 인용한 경로 `ls` 확인 — `internal/cli/testdata/codex-0.160.0/`(JSON 8개 + README.md), `internal/cli/managed_hardening_test.go`(스키마 가드 시험 `TestManagedServerRequestPolicyMatchesCodexSchema` 가 547행), `managed_codex_factory.go`·`managed_factory_session.go`·`managed_codex_factory_test.go`·`internal/config/defaults.go`·`.moai/docs/factory-managed-session.md`·`progress.md`·`red-baseline.md` 도 이 재종결의 `ls` 한 번(exit 0)으로 존재 확인; 템플릿 미러만 `No such file or directory` — 기대대로
canary_compliance_check: 해당 없음 — 이 SPEC 은 자기 sync 가 시험해야 하는 전방 정책을 정의하지 않는다

귀속은 모두 `(this run, this tree, HEAD 3b12e75a1 위의 재종결 작업 트리, 커밋 전)` 이다. 이 재종결은 프로덕션·시험 코드를 바꾸지 않았고 `go test` 는 선택 수 확인(`-list`) 한 번만 돌렸다. 스크래치 파일은 인용 대상이 아니며 필요한 줄만 아래에 옮겼다.

#### 재종결 증거

| 항목 | 명령 | 관측 | exit |
|---|---|---|---|
| 사전 상태 | `git rev-parse --show-toplevel` / `git branch --show-current` / `git rev-parse --short HEAD` / `git status --short` | `…/.moai/worktrees/t1409` / `WT-managed-session-hardening` / `3b12e75a1` / 출력 없음 | 0 |
| 선택 수 | `unset MOAI_KANBAN … MOAI_KANBAN_BACKEND && go test ./internal/cli -list '^.*(Managed\|managed).*$'` 를 파일로 돌려 `grep -c '^Test'`(원문 명령은 acceptance.md §1.1 AC-MH-013 블록) | `79`, 끝 줄 `ok  github.com/modu-ai/moai-adk/internal/cli  1.408s`, `no tests to run` 0건, `^TestManagedServerRequestPolicyMatchesCodexSchema$` 1건 | 0 |
| 코드 대조 | `grep -n 'managedLegacyApprovalRejection\|managedLegacyDeniedResult\|applyPatchApproval\|execCommandApproval' internal/cli/managed_codex_factory.go` | 정책표 360·361행이 `managedLegacyDeniedResult()` 를 가리키고 368–369행이 `{"decision":{"denied":{"rejection":managedLegacyApprovalRejection}}}` 를 만든다. 문자열 `"denied"` 는 와이어 값에 없고 로그 토큰 `denied` 로만 남는다 | 0 |
| 스키마 사본 | `ls internal/cli/testdata/codex-0.160.0/` | JSON 8개(`ApplyPatchApprovalResponse`, `CommandExecutionRequestApprovalResponse`, `DynamicToolCallResponse`, `ExecCommandApprovalResponse`, `FileChangeRequestApprovalResponse`, `JSONRPCError`, `McpServerElicitationRequestResponse`, `PermissionsRequestApprovalResponse`) + `README.md` | 0 |
| 새 의존 없음 | `git diff --stat HEAD -- go.mod go.sum` / `grep -n 'santhosh-tekuri/jsonschema' go.mod` | 출력 없음 / `86: github.com/santhosh-tekuri/jsonschema/v6 v6.0.2`(기존 직접 의존) | 0 |
| MX ANCHOR | `grep -n '@MX:ANCHOR' internal/cli/managed_codex_factory.go` / `git show --stat f4f47dc51` | 224행 `@MX:ANCHOR: [AUTO] the single path for every connection write …` 와 225·226행 `@MX:REASON`·`@MX:SPEC` / `managed_codex_factory.go \| 4 ++++` | 0 |
| ANCHOR fan_in | `grep -n '\.write(' internal/cli/managed_codex_factory.go internal/cli/managed_factory_session.go` (비시험 파일) | 410(`answerServerRequest`) · 479(`call`) · 715(`Start` 의 `initialized`) — 3곳 | 0 |
| 변경 목록(커밋 전) | `git status --short` | ` M .moai/docs/factory-managed-session.md` · ` M .moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/spec.md` · ` M CHANGELOG.md` (이 블록을 쓴 뒤 `progress.md` 가 더해져 네 경로) | 0 |
| AC-MH-010 문서 앵커(부재) | `grep -c "턴 하나가 실패하면 세션 전체가 끝난다"` / `"서버가 먼저 보내는 승인·질문 요청에는 답하지 않는다"` / `"후속 카드 t1409 대상"` 를 `.moai/docs/factory-managed-session.md` 에 | `0` / `0` / `0` | 1(각각) |
| 부재 앵커의 양성 대조 | `git grep -c "후속 카드 t1409 대상" fe79bfa0e -- .moai/docs/factory-managed-session.md CHANGELOG.md` | `fe79bfa0e:.moai/docs/factory-managed-session.md:1` · `fe79bfa0e:CHANGELOG.md:1` | 0 |
| AC-MH-010 문서 앵커(존재) | 같은 파일에 `grep -c` | `런처는 시그널을 처리하지 않는다` 1 · `t1459` 1 · ``관리 계층은 .syscall. 을 쓰지 않는다`` 1 · `TTL까지 재배달` 1 · `턴 타임아웃은 세션을 끝낸다` 1 · `거부된 elicitation` 1 · `쓰기 데드라인` 1 · `실제 codex 세션에서는 관측하지 않았다` 1 | 0 |
| AC-MH-010 CHANGELOG | `grep -c` 를 `CHANGELOG.md` 에 | `SPEC-FACTORY-MANAGED-HARDEN-001` → `1` · `t1459` → `2` · `후속 카드 t1409 대상` → `0`(exit 1) | 0 / 0 / 1 |
| 템플릿 미러 부재 | `ls internal/template/templates/.moai/docs/factory-managed-session.md` | `No such file or directory` | 1 |
| gofmt (감사 F5) | `gofmt -l internal/cli internal/config` / `git diff --stat 7109e0900 HEAD -- internal/cli/mcp_claude.go internal/config/slice.go` | `internal/cli/mcp_claude.go` · `internal/config/slice.go` 두 줄 / 출력 없음 | 0 / 0 |
| `moai spec lint SPEC-FACTORY-MANAGED-HARDEN-001` | (지연 빌드로 실행) | `✓ No findings — all SPEC documents are valid` | 0 |
| 설치된 `moai` 의 지연 | `moai version` / `git merge-base --is-ancestor 0732cc699 HEAD` / `git merge-base --is-ancestor HEAD 0732cc699` | `v3.2.0-rc.27   archive/t1401-504-g0732cc699   built 2026-10-03T03:34:50Z` / exit 1 / exit 1 | 0 / 1 / 1 |

(표 안의 `\|` 는 원문 명령에서 `|` 한 글자다 — 표 셀에서 파이프를 이스케이프한 것이다. `grep` 의 `\|` 교번은 원문에서도 `\|` 이다.)

`moai spec lint` 의 통과는 **빌드 출처가 제한된 증거**일 뿐이다: 설치된 빌드의 커밋 `0732cc699` 는 이 트리 HEAD 의 조상이 아니고 HEAD 도 그 조상이 아니라(둘 다 exit 1, 갈라진 이력) 이 트리가 담은 점검을 갖췄다고도 말할 수 없다. 이 트리 빌드의 판정이 아니다.

#### MX 태그 검증 (재종결 하위 단계)

첫 종결 블록의 MX 소견 가운데 `managedCodexAppClient.write` 의 `@MX:ANCHOR` 후보는 커밋 `f4f47dc51` 이 해소했다(위 증거: 224–226행, fan_in 3곳 grep). 그 블록의 나머지 소견은 그대로 열려 있다: `errManagedTurnFailed`(변수, `managed_factory_session.go:65`)와 `noteElicitation` 의 귀속 규칙은 `@MX:NOTE` 후보(선택), `managedLogf` 는 fan_in 3 미만. 이 재종결은 `.go` 를 고치지 않았고 `moai mx` 스캐너를 돌리지 않았다 — 호출 지점 grep 과 태그 grep 만 읽었다.

#### 감사 소견의 처분 (`.moai/reports/t1409/sync-audit.md`)

- **F1 [blocking]** 수리됐다: 코드·시험은 `2796a98d3`, SPEC 정정(design.md D-1·acceptance.md AC-MH-001·013·spec.md HISTORY)은 개정 `3b12e75a1`, 운영 문서·CHANGELOG 문구는 이 재종결. 재감사는 델타 감사(F1 한정)가 맡으며 이 재종결은 그 판정이 아니다.
- **F2 [minor, 고치지 않음]** 서버가 보낸 `turnId` 가 elicitation 로그 줄에 따옴표 없이 실린다. 지금 줄 번호는 `internal/cli/managed_codex_factory.go:397`(`turn=%s`, 감사가 읽을 때는 383행)이고, `Factory turn failed (%d/%d consecutive): %v` 줄(`managed_factory_session.go:364`)도 오류 문구를 따옴표 없이 싣는다. SPEC 이 로그 문법을 `turn=<id>` 로 고정해 이 카드는 고치지 않았다. 운영 문서 한계 항목과 CHANGELOG 한계 ⑧에 공시했다. 실행 재현은 없다(소스 판독).
- **F3 [leader 의 프로세스 소견, 이 재종결이 다루지 않음]** plan→run Kickoff 를 PASS-WITH-DEBT(0.81) 위에서 자율로 취한 읽기는 리더가 추인하거나 기각한다. manager-docs 가 판단할 일이 아니다.
- **F4 [minor, 기록만]** `f4f47dc51` 은 레인 오케스트레이터가 직접 만든 주석 4줄 커밋이라 `Authored-By-Agent` 트레일러가 없다. 감사가 이미 규칙 위반이 아님을 확인했다. `@MX:ANCHOR` 가 `write()` 에 있다는 사실만 위 MX 항목에서 다시 읽어 확인했다.
- **F5 [minor, 고치지 않음]** `gofmt -l internal/cli internal/config` 가 이 카드가 건드리지 않은 두 파일 `internal/cli/mcp_claude.go`, `internal/config/slice.go` 를 여전히 나열한다(위 증거: 기준 트리 `7109e0900` 대비 두 파일의 diff 가 비어 있음). AC-MH-013 의 "출력 0행" 문면과 이 트리의 사전 상태는 어긋난 채 남는다. 같은 명령의 출력에는 이 카드가 건드린 파일이 하나도 없다(출력은 위 두 줄뿐) — 이 재종결에서 측정한 것이다.
- **F6 [info]** 공시된 한계는 감사가 그대로 확인했다. 이 재종결도 한계 문장을 줄이지 않았다.

**Claim**: AC-MH-010(운영 문서와 CHANGELOG)이 레거시 승인 거부 객체 수리 뒤의 상태로 다시 충족되고, F3·F4 는 해결로, 시그널 처리 공백(F5)은 카드 t1459 를 가리키는 미해결 한계로 적혔다. 문서·CHANGELOG 의 선택 수는 79 이고 REQ 12·AC 13 의 개수는 유지된다. **Evidence**: 위 표의 명령·관측·exit. **Baseline-attribution**: `(this run, this tree, HEAD 3b12e75a1 위의 재종결 작업 트리, 커밋 전)`. 커밋 뒤 점검(`git diff --stat 3b12e75a1 HEAD`, spec.md diff, `internal` 불변, 부모 SPEC 불변, 트레일러)은 이 신호를 담은 커밋 자신에 대한 것이라 커밋 전에는 얻을 수 없으며 sync 보고서에 따로 적는다.

**Gaps**: ① 라이브 Codex 미관측 — 레거시 승인 요청에 돌려주는 새 거부 객체는 vendoring 한 스키마에 대한 유효성만 관측했고, 어떤 실제 codex 세션도 그 객체를 받아 본 적이 없다. 거부 응답 뒤 모델 행동, elicitation 요청의 `serverName` 값, "거부된 elicitation 이 영수증을 막는" 시나리오도 마찬가지다 ② `X == prevTurnID` 중복 완료 프레임 방어(G5)는 시험 행이 없음 ③ 막힌 쓰기 상한과 `id: null` 프레임은 공시한 한계로 남음(소스 판독) ④ 이 재종결은 `go test` 를 선택 수 확인 외에 돌리지 않았다 — 수리 후 managed 스코프 전체 PASS 218(하위 포함)·SKIP 3·FAIL 0·DATA RACE 0 은 §E.2 R1 의 manager-develop 기록을 인용한 것이지 이 재종결의 재측정이 아니다 ⑤ `moai spec lint` 는 빌드 출처가 제한돼 이 트리의 판정이 아니다 ⑥ MX 소견은 스캐너 실행이 아니라 grep 과 임계값 판독이다 ⑦ 커밋 뒤 점검 결과는 이 파일에 담기지 않는다 ⑧ 문서의 서술은 `internal/cli/managed_codex_factory.go`·`managed_factory_session.go`·`defaults.go` 와 스키마 사본을 이 재종결이 읽은 것에 근거하고 라이브 세션으로 확인한 것이 아니다 ⑨ 이 파일에는 `sync_commit_sha:` 줄이 둘이다(첫 종결의 정규 플레이스홀더와 이 재종결의 값). `internal/spec/era.go` 의 `extractProgressField` 는 **첫 일치만** 읽으므로 도구가 읽는 값은 첫 종결의 `pending-backfill` 이다 — 비어 있지 않아 era 분류와 drift 판정은 같지만, 도구가 현재 종결의 SHA 를 읽는다고 말할 수는 없다. 첫 종결 블록을 그대로 두라는 지시와 어긋나 고치지 않았고 보고한다.

**Residual-risk**: ① 델타 재감사가 남아 있다 — 이 재종결은 감사 판정이 아니다 ② 계획 감사 부채 N3–N12 는 optional 로 열린 채 닫히지 않았다 ③ `internal/guardstate` 의 사전 적색 `TestCensus_SetDifferenceEmptyBothDirections` 는 이 재종결에서 재측정하지 않았다 ④ 연속 실패 상한 3 은 UNMEASURED 관례 값이다 ⑤ 스키마 사본은 codex 0.160.0 기준이라 최소 지원 버전이 올라가면 재생성하지 않는 한 가드가 옛 스키마를 계속 믿는다(사본 README 에 명시) ⑥ 거부 문구 상수는 영어 한 줄이며 서버가 모델에 그대로 전달할 수 있다 ⑦ 독 메시지와 거부된 elicitation 의 TTL 까지 재배달, 쓰기 데드라인 부재, `id: null` 프레임은 닫히지 않았다 ⑧ M3 커밋 `1963ec376` 본문의 두 구절은 `--amend` 금지로 그대로 남고 §E.2 M4 항목 6 이 정정한다 ⑨ t1408·t1410·t1459 는 후속 카드로 남아 있다.
