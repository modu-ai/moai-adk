# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 수용 기준 (0.2.0)

## §A 측정 규칙

- 수용 기준은 9개(Tier M 상한 16)다. 상한 셈은 `grep -c '^#### AC-' .moai/specs/SPEC-AGENTIC-FACTORY-IDENTITY-001/acceptance.md` → 9. 저장소 계수기 정규식 `AC-(?:[A-Z0-9]+-)*[0-9]+`는 다른 SPEC id도 세므로, 이 파일에서 상위 id를 쓸 때는 바로 뒤에 `[REF]`를 붙인다.
- 판정: 시험 명령은 시험 이름을 고르고, 출력에 고른 시험의 `--- PASS: <이름>`이 있고 `[no tests to run]`·`[no test files]`가 없으며, 하위 사례 PASS 줄 수가 기대 수와 같을 때만 PASS다. 종료 코드 0만으로는 PASS가 아니다.
- **채택 상태**: 9개 모두 채택 전이다. run 단계가 의도한 단언 때문의 `EXPECTED_RED`(명령, 원문 출력, 종료 코드, 트리 SHA)를 기록해야 채택된다. `setup failed`·컴파일 오류는 올바른 RED 이유가 아니다.
- **초록 경로 원칙(iteration-1 D5)**: 어떤 기준도 그것이 단언하는 것을 만드는 마일스톤보다 먼저 초록이 되지 않는다. 마일스톤은 plan.md.
- **시험 경계**: 사례마다 "다른 거절 조건은 모두 통과하는 상태"에서 조건 하나만 어긴다. 그래서 M3 전 사례는 유효한 운영자 확인을 부여기 안에 직접 넣는 시험용 경계("유효 확인 주입")를 쓸 수 있다. 실제 운영 경로에서 확인은 design §4로만 생긴다. AC-AFI-002는 이 경계를 쓰지 않는다.
- **실제 프로세스와 주입 프로세스 표**: 계보 사례는 시험이 실제 자식 프로세스를 띄워 확인한다(Linux·macOS·Windows CI 각각). 띄운 프로세스는 `t.Cleanup`에 종료를 등록하거나 `timeout`으로 감싼다. 판정 불가·재사용 사례는 프로세스 표를 주입해 만든다.
- **대조군**: 거절 사례 묶음마다 거절 조건 하나만 빼면 성공하는 대조군을 둔다("모두 거절" 변형을 잡기 위해).
- 사유 문자열은 design §6이 닫힌 집합이다. 상위 게이트로 가면 모두 `producer-not-allowed`다(상위 design §2.1b).

## §B 증거 원장

```
id: EV-AFI-REF-0
command: ls internal/agenticfactory
stdout:
ls: internal/agenticfactory: No such file or directory
exit_code: 1
tree_sha: 8dc835756b09f8e8dd4669b1ad7b47a46b15e9f3
classification: 대상 부재 참고 관측 — EXPECTED_RED 아님
```

0.2.0 작성 트리(`3275706cc9`)에서 다시 실행하지 않았다. 어느 기준의 RED-now 칸도 되지 못한다.

## §C 수용 기준

#### AC-AFI-001 — 부여 권한과 control ref 기록·재구축

- 주 REQ: maps REQ-AFI-001. 초록: M1.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_GrantAuthority_ControlRefOnly$' -count=1 -v` → 하위 사례 PASS 6.
- 사례: (대조군) 유효 확인 주입으로 세션 S에 감사 역할 부여 → control ref에 부여 기록, readback 일치. (G1) ControlService가 아닌 쓰기 주체가 control ref 시험 ref에 S2 부여 기록을 써 넣음 → 부여 목록에 S2 없음. (G2) 로컬 캐시에만 S2 부여 기록 → 부여 목록에 S2 없음. (G3) 대조군 뒤 재시작, S가 살아 있고 결속 같음 → 재구축 뒤 S의 부여 유효. (G4) 재시작, S의 fingerprint가 바뀜 → 부여 `grant-reverify-failed` 표시, 기록은 지우지 않음. (G5) 확인 없이 부여 요청 → 부여 0. Coordinator·agent에 대한 부여 요청 → 거절(등록만 됨).
- 변형 단언: 캐시를 권한으로 읽는 구현(G2), 재시작에서 부여를 버리는 0.1.0 구현(G3), 재검증 없이 이어받는 구현(G4), Coordinator에 부여하는 구현(G5)이 FAIL.
- RED 이유: 부여기가 없어 대조군 단언 실패. 초록 경로: M1(lane-18, control ref CAS 위에).

#### AC-AFI-002 — 운영자 인증은 서버가 관측한 것만

- 주 REQ: maps REQ-AFI-002. 초록: M3.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_OperatorAuth_ServerObservedOnly$' -count=1 -v` → 하위 사례 PASS 8.
- 사례: (대조군) 요청자에 제어 단말이 있고(프로세스 표 관측) 운영자가 ControlService 단말에서 맞게 답함 → 부여, `auth_binding.method`는 서버 관측값. (A1) 답 없음 → 부여 0, `operator-auth-missing`. (A2) 대상 세션이 끝난 뒤의 binding → 만료. (A3) 이미 쓴 binding으로 두 번째 부여 → 재사용 거절. (A4) binding의 session id·역할·Issue가 대상과 다름 → 거절. (A5) 요청자에 제어 단말이 없는데 본문에 `method: interactive-tty`·단말 주장 → `no-controlling-terminal`, 본문 값은 판정 근거로 남지 않음. (A6) 맞는 답이 요청자 연결로 옴 → `operator-auth-missing`(iteration-1 D1 (다)). (A7) 같은 답이 ControlService 단말과 요청자 연결로 동시에 옴 → 단말의 답만 쓰이고 binding 하나.
- 변형 단언: 클라이언트 주장을 읽는 구현(A5), 요청자 연결의 답을 받는 구현(A6)이 FAIL.
- RED 이유: 확인 절차 없음. 초록 경로: M3.

#### AC-AFI-003 — 등록 세션과 그 자손의 요청 거절

- 주 REQ: maps REQ-AFI-003. 초록: M2.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_Ancestry_RegisteredDescendantRefused$' -count=1 -v` → 하위 사례 PASS 5.
- 사례: (대조군) 기준점 셸의 자식(등록 세션 아님)이 부여 요청 → 계보 통과. (D1) 등록된 agent 세션 자신이 자기 부여 요청 → `requester-is-session-descendant`. (D2) agent가 다른 세션 S에 대한 부여 요청 → 같음. (D3) agent의 자식(등록 안 됨, pty 붙임)이 부여 요청 → 같음. (D4) agent의 자식이 세션 등록 요청 → 같음.
- 변형 단언: 요청자 본인만 보고 자손을 보지 않는 구현(D3), 기준점 도달을 자손 판정보다 먼저 보는 구현(D3, 기준점 아래의 agent 자손이 통과)이 FAIL.
- RED 이유: 계보 검사기 없음. 초록 경로: M2.

#### AC-AFI-004 — 기준점 하나와 고아 경로

- 주 REQ: maps REQ-AFI-003. 초록: M2.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_Ancestry_OperatorAnchorAndOrphans$' -count=1 -v` → 하위 사례 PASS 6.
- 사례: (대조군) 기준점 셸의 손자 → 통과. (O1) agent가 두 번 fork해 pid 1(Linux)·launchd(darwin)에 입양된 프로세스의 부여 요청 → `requester-outside-operator-anchor`. (O2) Windows에서 부모가 끝나 `ParentProcessID`가 살아 있는 프로세스를 가리키지 않음 → 같은 코드. (O3) Windows·darwin에서 `ParentProcessID`가 가리키는 pid가 재사용되어 부모 시작 시각이 자식보다 늦음(주입 표) → 같은 코드. (O4) 기준점과 무관한 다른 터미널 셸의 자식 → 같은 코드. (O5) 고아 프로세스의 세션 등록 요청 → 등록은 되지만 그 세션의 이후 부여 요청은 `requester-outside-operator-anchor`.
- 변형 단언: 플랫폼마다 다른 코드를 내는 구현(O1~O3 코드 동일 단언), 고아를 판정 불가로 보내는 구현(D10 순서)이 FAIL.
- RED 이유: 기준점 기록 없음. 초록 경로: M2.

#### AC-AFI-005 — 판정 불가와 본문 pid

- 주 REQ: maps REQ-AFI-003. 초록: M2.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_Ancestry_IndeterminateAndBodyPid$' -count=1 -v` → 하위 사례 PASS 6.
- 사례: (I1) 요청자 fingerprint 읽기 실패(주입) → `requester-ancestry-indeterminate`. (I2) 중간 고리의 부모 fingerprint가 권한 거부 → 같음. (I3) 깊이 상한 초과(주입) → 같음. (I4) 플랫폼 구현 없음(주입) → 같음. (P1) 요청 본문에 기준점 셸의 pid를 적고 실제 연결은 agent 자식에서 옴 → 연결 pid로 판정해 `requester-is-session-descendant`. (P2) 본문 pid가 연결 pid와 같아도 판정 기록에 본문 값이 근거로 남지 않음.
- 변형 단언: 판정 불가를 통과로 보는 구현(I1~I4), 본문 pid를 쓰는 구현(P1)이 FAIL.
- RED 이유: 계보 검사기 없음. 초록 경로: M2.

#### AC-AFI-006 — 세션 신원 결속과 PID 재사용

- 주 REQ: maps REQ-AFI-004. 초록: M1.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_SessionBinding_ReobservedOnEveryRegistration$' -count=1 -v` → 하위 사례 PASS 6.
- 사례: (대조군) 등록한 세션이 같은 프로세스에서 증거 등록 → 받음. (B1) 다른 프로세스가 같은 session id로 증거 등록 → `identity-mismatch`. (B2) 같은 pid, 다른 fingerprint(재사용) → `identity-mismatch`. (B3) 살아 있는 프로세스에 묶인 session id로 두 번째 등록 → `session-already-bound`. (B4) 결속 프로세스가 `dead` → `identity-mismatch`. (B5) Linux 원천에서 같은 초 재사용(fingerprint 문자열 같음, 주입) → 결속 검사는 통과함을 **알려진 한계로 단언**하고, 같은 시험에서 그 프로세스가 기준점 사슬에 닿지 않으면 부여가 거절됨을 함께 단언(design §3.3 보완).
- 변형 단언: pid만 비교하는 구현(B2)이 FAIL. B5는 한계를 숨기는 구현(같은 초 재사용을 "탐지했다"고 주장하는 출력)을 잡기 위해 기대값을 명시한다.
- RED 이유: 등록기 없음. 초록 경로: M1.

#### AC-AFI-007 — Issue 단위 역할 배제

- 주 REQ: maps REQ-AFI-005. 초록: M4.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_RoleExclusion_PerIssueWriter$' -count=1 -v` → 하위 사례 PASS 5.
- 사례(유효 확인 주입): (X1) Issue A의 writer였던 세션에 A의 코드 리뷰 부여 → `was-writer`. (X2) 같은 세션에 Issue B(그 세션이 writer가 아님)의 코드 리뷰 부여 → 허용. (X3) A의 코드 리뷰 역할을 가진 세션에 A의 감사 부여 → `role-already-held`. (X4) A의 리뷰 세션에 B의 감사 부여 → 허용. (X5) 이전 attempt writer(현재는 아님)에 A의 QA 부여 → `was-writer`.
- 변형 단언: run 단위로 막는 0.1.0 구현은 X2·X4에서, 현재 attempt만 보는 구현은 X5에서 FAIL.
- RED 이유: 배제 규칙 없음. 초록 경로: M4.

#### AC-AFI-008 — 생산자 × 증거 종류 생성 행렬

- 주 REQ: maps REQ-AFI-006. 초록: M4.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_AttributionMatrix_Generated$' -count=1 -v` → 하위 사례 PASS 48. 추가 조건 사례는 `-run '^TestIdentity_AttributionMatrix_ExtraConditions$'` → PASS 3.
- 행렬: design §5의 생산자 7 × 종류 6 = 42칸(허용 8, `role-kind-mismatch` 34) + 미등록 생산자 × 6 = 6(`not-registered`). 시험은 표에서 사례를 **생성**하고 생성된 사례 수 48을 단언한다(빈 행렬이 통과하지 않도록).
- 추가 조건: (E1) 허용 칸이지만 생산자가 그 Issue의 writer였음 → `was-writer`. (E2) Coordinator가 그 변경을 직접 구현함 → `was-writer`. (E3) 운영자 기록이 audit 종류 → `role-kind-mismatch`(행렬에 있으나 이유를 명시).
- 변형 단언: "역할이 있으면 아무 종류나 허용" 구현은 34칸에서, "모두 거절" 구현은 허용 8칸에서 FAIL.
- RED 이유: 귀속기 없음. 초록 경로: M4.

#### AC-AFI-009 — 읽기 전용 귀속 조회와 닫힌 사유

- 주 REQ: maps REQ-AFI-007. 초록: M4.
- 명령: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_AttributionQuery_ReadOnlyClosedReasons$' -count=1 -v` → 하위 사례 PASS 4.
- 사례: (Q1) 조회 전후 control ref 순번·부여 목록·등록 결속이 바이트 동일, 쓰기 호출 0. (Q2) 조회가 돌려준 모든 이유 값이 design §6.2 집합 안. (Q3) 재시작 직후 재검증 전 조회 → 모든 증거 `grant-reverify-failed`. (Q4) 이 SPEC의 조회 구현은 `identity-spec-incomplete`를 스스로 내지 않음(그 값은 상위 잠금 원천의 고정값).
- 상위 연결 확인은 상위 AC-AF-011 [REF]의 해제 뒤 형태(상위 M7)에서 한다. 이 기준은 상위 게이트를 부르지 않는다.
- RED 이유: 조회 없음. 초록 경로: M4.

## §D 변형 점검표 (문서 추론)

실제 변형 코드를 돌려 본 것이 아니다. run 단계가 변형을 만들어 해당 사례가 FAIL하는지 관측해야 한다.

| AC | 변형 | 죽이는 사례 |
|---|---|---|
| 001 | 캐시를 권한으로 | G2 |
| 001 | 재시작에서 재검증 없이 이어받음 | G4 |
| 002 | 클라이언트 주장 사용 | A5 |
| 002 | 요청자 연결의 답 사용 | A6 |
| 003 | 기준점 검사를 자손 검사보다 먼저 | D3 |
| 004 | 플랫폼별 다른 사유 코드 | O1~O3 |
| 005 | 판정 불가를 통과로 | I1~I4 |
| 005 | 본문 pid 사용 | P1 |
| 006 | pid만 비교 | B2 |
| 007 | run 단위 배제 | X2·X4 |
| 008 | 역할만 보고 종류 무시 | 34 mismatch 칸 |
| 008 | 모두 거절 | 허용 8칸 |
| 009 | 조회가 상태를 바꿈 | Q1 |

## §E 완료 정의

1. AC-AFI-001~009가 같은 HEAD에서 PASS(기대 PASS 수 일치).
2. 각 AC의 RED-now 칸이 올바른 이유의 `EXPECTED_RED`.
3. OQ-AFI-1에 대한 운영자 위험 수용 기록(상위 OD-4). 이것 없이는 `completed`가 아니다.
4. 독립 감사 무부채 PASS(새 감사는 운영자 승인 대상).
5. 상위 잠금 해제(상위 M7)는 이 정의 밖이며, 이 SPEC 완료 뒤 상위가 한다.
