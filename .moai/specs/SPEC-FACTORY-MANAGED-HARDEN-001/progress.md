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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
