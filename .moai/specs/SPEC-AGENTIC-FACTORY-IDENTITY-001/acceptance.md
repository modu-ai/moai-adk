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
- **F2b 정정(시험 존재 점검, `4cdaba5d7e` 기준)**: 아홉 기준이 고르던 `TestIdentity_*` 시험은 동결 후보 `4cdaba5d7e`에 하나도 없고, `internal/agenticfactory/identity` 패키지도 없다(§F). 고를 시험이 없는 명령에는 고른 시험의 `--- PASS:` 줄이 생길 수 없으므로, 위 판정 규칙상 PASS가 될 수 없다(그 명령이 실제로 무엇을 출력하는지는 실행하지 않아 관측하지 않았다). 그래서 아홉 기준 모두 명령을 두지 않고 "미충족 — 구현 없음 / 미충족 — 시험 없음"으로 적었다. 예전 이름과 하위 사례 수는 "목표 시험 이름(아직 없음, 실행 대상 아님)"으로 남겼다. 지금 기대 PASS 수는 모든 기준에서 0이다. 존재 확인은 `git grep`·`git ls-tree`로만 했고 아무것도 실행하지 않았다. lane-18이 진행 중인 IDENTITY 구현(`6b945fd844` HostFacts, 리더 전달, 이 작성자는 읽지 않음)이 `4cdaba5d7e` 뒤에 시험을 이미 더했을 수 있다. 이 정정은 lane-17이 승인했다(이 SPEC은 lane-17 소관).

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
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L01·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(부여기 없음). 미충족 — 시험 없음: 아래 사례 전부(대조군, G1~G5).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_GrantAuthority_ControlRefOnly`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 6. 지금 기대 PASS 0.
- 사례: (대조군) 유효 확인 주입으로 세션 S에 감사 역할 부여 → control ref에 부여 기록, readback 일치. (G1) ControlService가 아닌 쓰기 주체가 control ref 시험 ref에 S2 부여 기록을 써 넣음 → 부여 목록에 S2 없음. (G2) 로컬 캐시에만 S2 부여 기록 → 부여 목록에 S2 없음. (G3) 대조군 뒤 재시작, S가 살아 있고 결속 같음 → 재구축 뒤 S의 부여 유효. (G4) 재시작, S의 fingerprint가 바뀜 → 부여 `grant-reverify-failed` 표시, 기록은 지우지 않음. (G5) 확인 없이 부여 요청 → 부여 0. Coordinator·agent에 대한 부여 요청 → 거절(등록만 됨).
- 변형 단언: 캐시를 권한으로 읽는 구현(G2), 재시작에서 부여를 버리는 0.1.0 구현(G3), 재검증 없이 이어받는 구현(G4), Coordinator에 부여하는 구현(G5)이 FAIL.
- RED 이유: 부여기가 없어 대조군 단언 실패. 초록 경로: M1(lane-18, control ref CAS 위에).

#### AC-AFI-002 — 운영자 인증은 서버가 관측한 것만

- 주 REQ: maps REQ-AFI-002. 초록: M3.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L02·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(확인 절차 없음). 미충족 — 시험 없음: 아래 사례 전부(대조군, A1~A7).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_OperatorAuth_ServerObservedOnly`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 8. 지금 기대 PASS 0.
- 사례: (대조군) 요청자에 제어 단말이 있고(프로세스 표 관측) 운영자가 ControlService 단말에서 맞게 답함 → 부여, `auth_binding.method`는 서버 관측값. (A1) 답 없음 → 부여 0, `operator-auth-missing`. (A2) 대상 세션이 끝난 뒤의 binding → 만료. (A3) 이미 쓴 binding으로 두 번째 부여 → 재사용 거절. (A4) binding의 session id·역할·Issue가 대상과 다름 → 거절. (A5) 요청자에 제어 단말이 없는데 본문에 `method: interactive-tty`·단말 주장 → `no-controlling-terminal`, 본문 값은 판정 근거로 남지 않음. (A6) 맞는 답이 요청자 연결로 옴 → `operator-auth-missing`(iteration-1 D1 (다)). (A7) 같은 답이 ControlService 단말과 요청자 연결로 동시에 옴 → 단말의 답만 쓰이고 binding 하나.
- 변형 단언: 클라이언트 주장을 읽는 구현(A5), 요청자 연결의 답을 받는 구현(A6)이 FAIL.
- RED 이유: 확인 절차 없음. 초록 경로: M3.

#### AC-AFI-003 — 등록 세션과 그 자손의 요청 거절

- 주 REQ: maps REQ-AFI-003. 초록: M2.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L03·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(계보 검사기 없음). 미충족 — 시험 없음: 아래 사례 전부(대조군, D1~D4).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_Ancestry_RegisteredDescendantRefused`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 5. 지금 기대 PASS 0.
- 사례: (대조군) 기준점 셸의 자식(등록 세션 아님)이 부여 요청 → 계보 통과. (D1) 등록된 agent 세션 자신이 자기 부여 요청 → `requester-is-session-descendant`. (D2) agent가 다른 세션 S에 대한 부여 요청 → 같음. (D3) agent의 자식(등록 안 됨, pty 붙임)이 부여 요청 → 같음. (D4) agent의 자식이 세션 등록 요청 → 같음.
- 변형 단언: 요청자 본인만 보고 자손을 보지 않는 구현(D3), 기준점 도달을 자손 판정보다 먼저 보는 구현(D3, 기준점 아래의 agent 자손이 통과)이 FAIL.
- RED 이유: 계보 검사기 없음. 초록 경로: M2.

#### AC-AFI-004 — 기준점 하나와 고아 경로

- 주 REQ: maps REQ-AFI-003. 초록: M2.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L04·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(기준점 기록 없음). 미충족 — 시험 없음: 아래 사례 전부(대조군, O1~O5).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_Ancestry_OperatorAnchorAndOrphans`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 6. 지금 기대 PASS 0.
- 사례: (대조군) 기준점 셸의 손자 → 통과. (O1) agent가 두 번 fork해 pid 1(Linux)·launchd(darwin)에 입양된 프로세스의 부여 요청 → `requester-outside-operator-anchor`. (O2) Windows에서 부모가 끝나 `ParentProcessID`가 살아 있는 프로세스를 가리키지 않음 → 같은 코드. (O3) Windows·darwin에서 `ParentProcessID`가 가리키는 pid가 재사용되어 부모 시작 시각이 자식보다 늦음(주입 표) → 같은 코드. (O4) 기준점과 무관한 다른 터미널 셸의 자식 → 같은 코드. (O5) 고아 프로세스의 세션 등록 요청 → 등록은 되지만 그 세션의 이후 부여 요청은 `requester-outside-operator-anchor`.
- 변형 단언: 플랫폼마다 다른 코드를 내는 구현(O1~O3 코드 동일 단언), 고아를 판정 불가로 보내는 구현(D10 순서)이 FAIL.
- RED 이유: 기준점 기록 없음. 초록 경로: M2.

#### AC-AFI-005 — 판정 불가와 본문 pid

- 주 REQ: maps REQ-AFI-003. 초록: M2.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L05·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(계보 검사기 없음). 미충족 — 시험 없음: 아래 사례 전부(I1~I4, P1, P2).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_Ancestry_IndeterminateAndBodyPid`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 6. 지금 기대 PASS 0.
- 사례: (I1) 요청자 fingerprint 읽기 실패(주입) → `requester-ancestry-indeterminate`. (I2) 중간 고리의 부모 fingerprint가 권한 거부 → 같음. (I3) 깊이 상한 초과(주입) → 같음. (I4) 플랫폼 구현 없음(주입) → 같음. (P1) 요청 본문에 기준점 셸의 pid를 적고 실제 연결은 agent 자식에서 옴 → 연결 pid로 판정해 `requester-is-session-descendant`. (P2) 본문 pid가 연결 pid와 같아도 판정 기록에 본문 값이 근거로 남지 않음.
- 변형 단언: 판정 불가를 통과로 보는 구현(I1~I4), 본문 pid를 쓰는 구현(P1)이 FAIL.
- RED 이유: 계보 검사기 없음. 초록 경로: M2.

#### AC-AFI-006 — 세션 신원 결속과 PID 재사용

- 주 REQ: maps REQ-AFI-004. 초록: M1.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L06·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(등록기 없음). 미충족 — 시험 없음: 아래 사례 전부(대조군, B1~B5).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_SessionBinding_ReobservedOnEveryRegistration`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 6. 지금 기대 PASS 0.
- 사례: (대조군) 등록한 세션이 같은 프로세스에서 증거 등록 → 받음. (B1) 다른 프로세스가 같은 session id로 증거 등록 → `identity-mismatch`. (B2) 같은 pid, 다른 fingerprint(재사용) → `identity-mismatch`. (B3) 살아 있는 프로세스에 묶인 session id로 두 번째 등록 → `session-already-bound`. (B4) 결속 프로세스가 `dead` → `identity-mismatch`. (B5) Linux 원천에서 같은 초 재사용(fingerprint 문자열 같음, 주입) → 결속 검사는 통과함을 **알려진 한계로 단언**하고, 같은 시험에서 그 프로세스가 기준점 사슬에 닿지 않으면 부여가 거절됨을 함께 단언(design §3.3 보완).
- 변형 단언: pid만 비교하는 구현(B2)이 FAIL. B5는 한계를 숨기는 구현(같은 초 재사용을 "탐지했다"고 주장하는 출력)을 잡기 위해 기대값을 명시한다.
- RED 이유: 등록기 없음. 초록 경로: M1.

#### AC-AFI-007 — Issue 단위 역할 배제

- 주 REQ: maps REQ-AFI-005. 초록: M4.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L07·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없다(배제 규칙 없음). 미충족 — 시험 없음: 아래 사례 전부(X1~X5).
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_RoleExclusion_PerIssueWriter`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 5. 지금 기대 PASS 0.
- 사례(유효 확인 주입): (X1) Issue A의 writer였던 세션에 A의 코드 리뷰 부여 → `was-writer`. (X2) 같은 세션에 Issue B(그 세션이 writer가 아님)의 코드 리뷰 부여 → 허용. (X3) A의 코드 리뷰 역할을 가진 세션에 A의 감사 부여 → `role-already-held`. (X4) A의 리뷰 세션에 B의 감사 부여 → 허용. (X5) 이전 attempt writer(현재는 아님)에 A의 QA 부여 → `was-writer`.
- 변형 단언: run 단위로 막는 0.1.0 구현은 X2·X4에서, 현재 attempt만 보는 구현은 X5에서 FAIL.
- RED 이유: 배제 규칙 없음. 초록 경로: M4.

#### AC-AFI-008 — 생산자 × 증거 종류 생성 행렬

- 주 REQ: maps REQ-AFI-006. 초록: M4.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L08·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없고, `runner.Attributor`를 구현한 테스트 밖 코드가 없다(§F 원장 L13·L14; `runner/runner.go:72-74` 주석: 구현은 identity SPEC 몫이며 아직 없다). 미충족 — 시험 없음: 행렬 48칸 전부와 E1~E3.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_AttributionMatrix_Generated`(하위 사례 목표 48), `TestIdentity_AttributionMatrix_ExtraConditions`(하위 사례 목표 3). 제안 패키지 `internal/agenticfactory/identity/`. 지금 기대 PASS 0.
- 행렬: design §5의 생산자 7 × 종류 6 = 42칸(허용 8, `role-kind-mismatch` 34) + 미등록 생산자 × 6 = 6(`not-registered`). 시험은 표에서 사례를 **생성**하고 생성된 사례 수 48을 단언한다(빈 행렬이 통과하지 않도록).
- 추가 조건: (E1) 허용 칸이지만 생산자가 그 Issue의 writer였음 → `was-writer`. (E2) Coordinator가 그 변경을 직접 구현함 → `was-writer`. (E3) 운영자 기록이 audit 종류 → `role-kind-mismatch`(행렬에 있으나 이유를 명시).
- 변형 단언: "역할이 있으면 아무 종류나 허용" 구현은 34칸에서, "모두 거절" 구현은 허용 8칸에서 FAIL.
- RED 이유: 귀속기 없음. 초록 경로: M4.

#### AC-AFI-009 — 읽기 전용 귀속 조회와 닫힌 사유

- 주 REQ: maps REQ-AFI-007. 초록: M4.
- 명령: 없음. `4cdaba5d7e`에 이 기준이 고를 시험이 없다(§F 원장 L09·L11).
- 미충족 — 구현 없음: `internal/agenticfactory/identity` 패키지가 없고, `runner.Attributor`를 구현한 테스트 밖 코드가 없다(§F 원장 L13·L14). 미충족 — 시험 없음: Q1~Q4 전부.
- 목표 시험 이름(아직 없음, 실행 대상 아님): `TestIdentity_AttributionQuery_ReadOnlyClosedReasons`(제안 패키지 `internal/agenticfactory/identity/`). 시험이 생긴 뒤의 하위 사례 목표 4. 지금 기대 PASS 0.
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

1. AC-AFI-001~009가 같은 HEAD에서 PASS(기대 PASS 수 일치). `4cdaba5d7e` 기준 기대 PASS 수는 아홉 기준 모두 0이다(§F). 각 기준의 하위 사례 목표 수는 목표 시험이 생기고 그 시험을 고르는 명령이 다시 적힌 뒤에만 기대 PASS 수가 된다.
2. 각 AC의 RED-now 칸이 올바른 이유의 `EXPECTED_RED`.
3. OQ-AFI-1에 대한 운영자 위험 수용 기록(상위 OD-4). 이것 없이는 `completed`가 아니다.
4. 독립 감사 무부채 PASS(새 감사는 운영자 승인 대상).
5. 상위 잠금 해제(상위 M7)는 이 정의 밖이며, 이 SPEC 완료 뒤 상위가 한다.

## §F 시험 이름 점검 (4cdaba5d7e 기준)

- 존재 확인은 `git grep`·`git ls-tree`로만 했다. 아무것도 실행하지 않았다(`go test` 0회). 체크아웃·트리 전환·쓰기는 없었다.
- 출력이 없는 명령의 종료 코드는 이 실행에서 따로 관측하지 않았다. "출력 없음"은 일치하는 줄이 없다는 뜻으로만 적는다.
- 부재 주장은 적힌 정규식이 닿는 범위로 한정한다. 같은 형태가 다른 이름에서는 실제로 걸린다는 것을 L11·L12로 확인했다(양성 대조).
- lane-18이 IDENTITY 항목을 구현하고 있다(`6b945fd844` HostFacts, 리더 전달, 이 작성자는 읽지 않음). 그 작업이 `4cdaba5d7e` 뒤에 시험을 이미 더했을 수 있으며, 이 표는 그 커밋을 판정하지 않는다.
- 감사 보고의 "`runner.Attributor`·`runner.Facts` 구현 없음"은 리더가 전달한 것이다. 이 작성자가 L13·L14로 따로 본 결과, 두 인터페이스는 `runner/runner.go:75`·`:123`에 선언만 있고 메서드 구현은 `_test.go` 파일에만 있다(L14). `effect/ghidentity/ghidentity.go:133`의 `IssueFacts`는 `context`를 받고 `effect.IssueFacts`를 돌려주는 다른 서명이라 `runner.Facts` 구현이 아니다. `runner.go:121-122` 주석은 `Facts` 구현이 forge 코드에 산다고 적으므로, `Facts` 부재는 이 SPEC 기준의 미충족 사유로 쓰지 않았다.

### F.1 기준별 상태

행동 축과 시험 축을 따로 적는다(F2d, 표지의 뜻은 상위 `acceptance.md` §G 첫머리).

- 표지 `잠금 유지 — REQ-AF-011에 따른 의도된 상태(미충족 아님)`는 이 표에 쓰지 않는다. 아홉 기준은 모두 이 SPEC 자신의 동작(부여, 인증, 계보, 결속, 배제, 귀속 행렬, 귀속 조회)이고, 상위 잠금의 동작이 아니다. 잠금 상태(`unattributable{identity-spec-incomplete}` → 게이트 `producer-not-allowed`)는 상위 REQ-AF-011과 AC-AF-011 [REF]의 것이며, Q4도 그 값을 이 SPEC의 구현이 스스로 내지 않는다는 단언이다.
- 근거: 상위 REQ-AF-011(상위 `spec.md` 150행) 원문 "the lock SHALL be lifted only by a single code change that lands together with SPEC-AGENTIC-FACTORY-IDENTITY-001 `completed`, its acceptance criteria green, and the operator's recorded risk acceptance of its relay-path residual risk". 곧 이 아홉 기준의 동작은 잠금 해제의 선행 조건이며, 잠금 중에는 운영 경로에 쓰이지 않는다.
- 감사 판정 문구(audit verdict wording relayed by the leader, not read by this SPEC's author): 'F2: production runner.guard always rejects identity-unknown and the merge gate is always blocked by AttributionLocked; fail-closed'. 리더 판정(relayed; lane-19 observation)은 병합 게이트 쪽을 rc.39 범위에서 의도된 상태로 본다. 이 표는 그 판정이 아니라 이 SPEC 구현의 부재를 적는다.
- 의도된 상태라도 시험이 없으면 `시험 없음`으로 남는다. 이 표의 시험 축 값은 F2b 그대로다.

| AC | 행동 축 | 시험 축 | 증거(원장 id: 명령 → 첫 출력 줄) |
|---|---|---|---|
| AC-AFI-001 | 미충족 — 구현 없음 | 시험 없음 | L01 → 출력 없음; L10 → identity 없음 |
| AC-AFI-002 | 미충족 — 구현 없음 | 시험 없음 | L02 → 출력 없음; L10 |
| AC-AFI-003 | 미충족 — 구현 없음 | 시험 없음 | L03 → 출력 없음; L10 |
| AC-AFI-004 | 미충족 — 구현 없음 | 시험 없음 | L04 → 출력 없음; L10 |
| AC-AFI-005 | 미충족 — 구현 없음 | 시험 없음 | L05 → 출력 없음; L10 |
| AC-AFI-006 | 미충족 — 구현 없음 | 시험 없음 | L06 → 출력 없음; L10 |
| AC-AFI-007 | 미충족 — 구현 없음 | 시험 없음 | L07 → 출력 없음; L10 |
| AC-AFI-008 | 미충족 — 구현 없음(실제 귀속 조회, `runner.Attributor` 테스트 밖 구현 없음) | 시험 없음 | L08 → 출력 없음; L10; L13·L14(테스트 밖 `Attribute` 구현 없음) |
| AC-AFI-009 | 미충족 — 구현 없음(같음) | 시험 없음 | L09 → 출력 없음; L10; L13·L14 |

셈: 명령이 있는 기준 0, 고른 시험 0, 지금 기대 PASS 0. 다른 패키지의 기존 시험(L11에 나온 `claim`·`effect`·`homestate` 등의 Identity 이름 시험)은 본문을 읽지 않았으므로 어느 사례에도 연결하지 않았다.

### F.2 원장

```
L01: git grep -n -E '^func (TestIdentity_GrantAuthority_ControlRefOnly)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L02: git grep -n -E '^func (TestIdentity_OperatorAuth_ServerObservedOnly)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L03: git grep -n -E '^func (TestIdentity_Ancestry_RegisteredDescendantRefused)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L04: git grep -n -E '^func (TestIdentity_Ancestry_OperatorAnchorAndOrphans)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L05: git grep -n -E '^func (TestIdentity_Ancestry_IndeterminateAndBodyPid)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L06: git grep -n -E '^func (TestIdentity_SessionBinding_ReobservedOnEveryRegistration)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L07: git grep -n -E '^func (TestIdentity_RoleExclusion_PerIssueWriter)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L08: git grep -n -E '^func (TestIdentity_AttributionMatrix_Generated|TestIdentity_AttributionMatrix_ExtraConditions)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L09: git grep -n -E '^func (TestIdentity_AttributionQuery_ReadOnlyClosedReasons)\(' 4cdaba5d7e -- '*_test.go'
     → 출력 없음
L10: git ls-tree --name-only 4cdaba5d7e internal/agenticfactory/
     → internal/agenticfactory/browser   (16줄, identity 없음: browser checkpoint claim core ctxmodel e2e effect forge gitref intake mirror monitor reconcile resident review runner)
L11: git grep -n -E '^func (TestIdentity_|Test[A-Za-z]*Identity)' 4cdaba5d7e -- '*_test.go'
     → 4cdaba5d7e:internal/agenticfactory/claim/controlref_test.go:236:func TestReleasedIdentityIsRefusedByEveryCache(t *testing.T) {
     (양성 대조: 이 형태는 걸린다. 결과 중 TestIdentity_ 로 시작하는 줄은 없다)
L12: git grep -c -E '^func Test' 4cdaba5d7e -- internal/agenticfactory/runner
     → 4cdaba5d7e:internal/agenticfactory/runner/adopt_test.go:3
L13: git grep -n -E 'Attributor|type Facts' 4cdaba5d7e -- internal/agenticfactory/runner
     → 4cdaba5d7e:internal/agenticfactory/runner/attribution_test.go:12:// fakeAttributor stands in for the read-only attribution lookup of
L14: git grep -n -E 'func \([^)]*\) (IssueFacts|Attribute)\(' 4cdaba5d7e -- internal/agenticfactory
     → 4cdaba5d7e:internal/agenticfactory/e2e/real_enforcer_test.go:34:func (f *fakeIdentity) IssueFacts(_ context.Context, _ string, issue int) (effect.IssueFacts, error) {
     (13줄. `Attribute` 구현 4줄은 모두 _test.go. 테스트 밖 줄은 effect/ghidentity/ghidentity.go:133 하나뿐이며 서명이 runner.Facts와 다르다)
```

리더가 전달한 같은 사실: `git ls-tree --name-only 4cdaba5d7e internal/agenticfactory/identity` → 출력 없음, `git grep -c -E '^func TestIdentity_' 4cdaba5d7e -- '*_test.go'` → 출력 없음. 이 실행의 L10·L11과 어긋나지 않는다.

### F.3 아직 없는 시험 이름

모두 제안 패키지 `internal/agenticfactory/identity/`, 구현 담당 lane-18(리더 전달).

- `TestIdentity_GrantAuthority_ControlRefOnly` — AC-AFI-001, 하위 사례 목표 6
- `TestIdentity_OperatorAuth_ServerObservedOnly` — AC-AFI-002, 목표 8
- `TestIdentity_Ancestry_RegisteredDescendantRefused` — AC-AFI-003, 목표 5
- `TestIdentity_Ancestry_OperatorAnchorAndOrphans` — AC-AFI-004, 목표 6
- `TestIdentity_Ancestry_IndeterminateAndBodyPid` — AC-AFI-005, 목표 6
- `TestIdentity_SessionBinding_ReobservedOnEveryRegistration` — AC-AFI-006, 목표 6
- `TestIdentity_RoleExclusion_PerIssueWriter` — AC-AFI-007, 목표 5
- `TestIdentity_AttributionMatrix_Generated` — AC-AFI-008, 목표 48
- `TestIdentity_AttributionMatrix_ExtraConditions` — AC-AFI-008, 목표 3
- `TestIdentity_AttributionQuery_ReadOnlyClosedReasons` — AC-AFI-009, 목표 4
