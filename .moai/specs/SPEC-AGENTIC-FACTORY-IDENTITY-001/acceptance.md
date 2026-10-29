# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 수용 기준

## §A 측정 규칙

- 수용 기준은 9개(Tier M 상한 16)다. 모두 REQ 하나 이상에 연결된다.
- 각 시험 명령은 `-run '^<이름>$' -v`로 시험 하나를 고른다. 판정은 출력에 `--- PASS: <이름>`이 있고 `[no tests to run]`·`[no test files]`가 없을 때만 PASS다. 종료 코드 0만으로는 PASS가 아니다.
- **채택 상태(0.1.0)**: 9개 모두 채택 전이며 release-blocking이 아니다. run 단계가 각 시험에서 의도한 단언 때문에 실패하는 `EXPECTED_RED`(명령, 원문 출력, 종료 코드, 트리 SHA)를 기록해야 채택된다. 패키지가 없어 생기는 `setup failed`와 컴파일 오류는 올바른 RED 이유가 아니다.
- **시험 경계**: 각 사례는 "다른 거절 조건은 모두 통과하는 상태"에서 **조건 하나만** 어긴다. 그래서 시험은 유효한 운영자 ack를 ControlService의 부여기 안에 직접 넣는 시험용 경계를 쓸 수 있다(실제 운영 경로에서 ack는 §4 확인 절차로만 생긴다). 이 경계를 쓰는 사례는 "유효 ack 주입"이라고 적는다.
- **실제 프로세스와 주입 프로세스 표**: 계보 사례는 시험이 실제 자식 프로세스를 띄워 확인한다(Linux·macOS·Windows CI 각각). 띄운 프로세스는 시험 정리 훅(`t.Cleanup`)에 종료를 등록하거나 `timeout`으로 감싼다. 판정 불가 사례(프로세스 사라짐, 시작 시각 읽기 실패, 권한 거부, 시작 순서 역전)는 프로세스 표를 주입해 만든다.
- **대조군**: 거절 사례 묶음마다 같은 입력에서 거절 조건 하나만 빼면 성공하는 대조군을 둔다("모두 거절" 변형을 잡기 위해).

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

이 관측은 패키지가 아직 없다는 사실만 보인다. 어느 기준의 RED-now 칸도 되지 못한다.

## §C 수용 기준

#### AC-AFI-001 — 부여 권한은 ControlService의 이번 기동 목록뿐 (maps REQ-AFI-001)
- Given 등록된 세션 S(역할 없음, Worker였던 적 없음)와 운영자 기준점에서 나온 요청자
- When (대조군) 운영자가 확인 절차를 거쳐 S에 감사 역할을 부여하고 S가 감사 PASS 증거를 등록한다. (R9) ControlService가 아닌 프로세스가 claim 저장소 파일에 S의 `role_grant` 기록을 직접 써 넣고, 다른 세션 S2(역할 없음)를 그 기록의 대상으로 둔다. (R10) 대조군 뒤 ControlService를 재시작한다. (무인증) 확인 절차 없이 S2에 대한 부여를 요청한다
- Then (대조군) 부여가 성공하고 귀속 조회가 `attributed{kind: audit, role: audit}`를 돌려준다. (R9) S2의 증거는 `unattributable{no-grant}`이고 무결성 위반 기록이 남는다. (R10) 재시작 전 부여는 이어받지 않아 S의 증거는 `unattributable{grant-from-previous-start}`다. (무인증) 부여가 없고 S2의 증거는 `unattributable{no-grant}`다. 어느 사례에서도 역할이 기본값으로 생기지 않는다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_GrantAuthority_ControlServiceCurrentStartOnly$' -count=1 -v`
- RED 이유: 저장소의 부여 기록을 읽어 권한으로 쓰거나, 재시작 때 부여를 이어받거나, 인증 없이 기본 부여를 해 해당 사례가 `attributed`가 된다
- 초록 경로: M2

#### AC-AFI-002 — 운영자 인증은 서버가 관측한 것만 (maps REQ-AFI-002)
- Given 계보 검사를 통과하는 요청자와 대상 세션 S(역할 없음)
- When 다음 사례에서 부여를 요청한다. 대조군: 요청자에게 제어 단말이 있고(프로세스 표에서 관측) 운영자가 확인 값을 맞게 답한다. (R8a) 확인 답 없음. (R8b) 대상 세션이 끝난 뒤의 ack(만료). (R8c) 이미 한 번 쓴 ack로 두 번째 부여(재사용). (R8d) ack의 session id 또는 역할이 대상과 다름(불일치). (R13) 요청자에게 제어 단말이 **없지만**(프로세스 표에서 관측) 요청 본문에 `method: interactive-tty`와 `stdin_is_terminal: true`를 적어 보냄
- Then 대조군만 부여되고 `operator_ack.method`에는 서버가 관측한 값이 적힌다. 나머지 다섯 사례는 모두 부여가 없고, 그 역할이 필요한 상위 게이트 대역은 `blocked(operator-ack-missing)`를 기록한다. (R13)에서 요청 본문의 방식 값은 기록 어디에도 판정 근거로 남지 않는다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_OperatorAuth_ServerObservedOnly$' -count=1 -v`
- RED 이유: ack가 있기만 하면 부여하거나, 만료·재사용·대상을 보지 않거나, 클라이언트가 보낸 방식 값을 믿어 해당 사례가 부여된다
- 초록 경로: M3

#### AC-AFI-003 — 등록 세션이 낸 부여 요청은 그것만으로 거절 (maps REQ-AFI-003)
- Given 역할이 없고 Worker였던 적이 없는 새 대상 세션 T, 그리고 T에 묶였고 쓰지 않았고 만료되지 않은 유효 ack(유효 ack 주입). 요청자는 셋이다: (대조군) 운영자 기준점의 직속 자식이며 등록 세션이 아닌 프로세스, (R1) 현재 attempt의 Worker 세션 프로세스 W 자신, (R2) 다른 등록 세션 S2(리뷰 역할 보유) 프로세스
- When 세 요청자가 **같은 대상 T, 같은 역할, 같은 모양의 유효 ack**로 감사 역할 부여를 요청한다(사례마다 새 유효 ack)
- Then 대조군은 부여된다. (R1)·(R2)는 거절되고 거절 기록 코드가 `requester-is-session-descendant`이며, 부여 목록과 T의 역할은 바뀌지 않는다. 세 사례에서 다른 조건(대상, ack, 역할 분리)은 모두 같으므로, 대조군과의 차이는 요청자뿐이다
- 추가: (R1')은 (R1)과 같지만 대상이 W 자신이다. 거절 기록 코드가 역시 `requester-is-session-descendant`여야 한다(역할 분리 검사보다 요청자 검사가 먼저임을 확인)
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_RegisteredSessionRequester_RefusedEvenWithValidAck$' -count=1 -v`
- RED 이유: 요청자 검사가 없어 (R1)·(R2)가 부여되거나, 다른 검사(인증·역할 분리)에만 기대어 거절 기록 코드가 다르게 나온다
- 초록 경로: M2

#### AC-AFI-004 — 등록 세션의 미등록 자손이 낸 요청은 거절 (maps REQ-AFI-003)
- Given Worker 세션 W와 대상 세션 T(역할 없음), T에 묶인 유효 ack(유효 ack 주입). 시험은 실제 프로세스를 띄운다: (R11a) W의 직속 자식 C1(등록 안 됨), (R11b) W의 손자 C2(등록 안 됨)에 pty를 붙이고 제어 단말을 가진 상태, (R11c) C2와 같지만 요청 본문에 `method: interactive-tty`를 적음. 대조군: 운영자 기준점의 자식이며 W와 무관한 프로세스에 같은 pty 조건
- When 각 프로세스가 T에 대한 감사 역할 부여를 요청한다
- Then 대조군만 부여된다. (R11a)~(R11c)는 모두 거절되고 거절 기록 코드가 `requester-is-session-descendant`다. pty와 제어 단말, 방식 값은 결과를 바꾸지 않는다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_SessionDescendantRequester_RefusedWithPty$' -count=1 -v`(Linux·macOS·Windows CI 각각. Windows는 pty 대신 콘솔이 붙은 자식으로 같은 사례를 만든다)
- RED 이유: 요청자 본인만 등록 세션인지 보고 조상을 보지 않아 C1·C2가 부여된다
- 초록 경로: M2

#### AC-AFI-005 — 계보를 판정할 수 없거나 기준점에 닿지 않으면 거절 (maps REQ-AFI-003)
- Given 대상 세션 T와 유효 ack(유효 ack 주입). 사례: (R12a) 사슬 중간 프로세스가 걷는 도중 사라짐(주입 표), (R12b) 조상 하나의 시작 시각 읽기 실패(주입 표), (R12c) 부모의 시작 시각이 자식보다 늦음 — PID 재사용(주입 표), (R12d) 다른 사용자 프로세스라 읽기 권한 거부(주입 표), (R12e) 깊이 상한 초과(주입 표), (R12f) Worker가 자식을 두 번 fork해 고아로 만든 프로세스(실제 프로세스, 부모가 pid 1·launchd로 바뀜)가 요청. 대조군: 같은 표에서 문제 고리 하나만 정상으로 바꾼 사슬
- When 각 요청자가 부여를 요청한다
- Then 대조군만 부여된다. (R12a)~(R12e)는 거절되고 기록 코드가 `requester-ancestry-indeterminate`다. (R12f)는 거절되고 기록 코드가 `requester-outside-operator-anchor`다. 어느 사례도 기본 부여로 이어지지 않는다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_AncestryIndeterminateOrOrphaned_Refused$' -count=1 -v`
- RED 이유: 판정할 수 없는 고리를 "자손 아님"으로 읽거나, 기준점 도달을 확인하지 않아 고아 프로세스가 부여된다
- 초록 경로: M2

#### AC-AFI-006 — 세션 신원 결속 (maps REQ-AFI-004)
- Given 리뷰 역할을 받은 세션 S2(결속 pid P2, 시작 시각 F2)
- When (대조군) S2가 증거를 등록한다. (R6) 다른 프로세스 P3가 S2의 session id를 복사해 증거를 등록한다. (R7a) S2가 끝난 뒤 같은 pid P2·다른 시작 시각의 프로세스가 S2의 session id로 등록한다(주입 표). (R7b) 살아 있는 S2의 session id로 다른 프로세스가 다시 세션 등록을 요청한다. (R7c) 이미 session id를 가진 프로세스가 두 번째 등록을 요청한다. (R7d) 요청 본문에 P2를 적었지만 연결 상대 pid는 P3다
- Then 대조군만 받아들여지고 `attributed`다. (R6)·(R7a)·(R7d)의 증거는 `unattributable{identity-mismatch}`, (R7b)·(R7c)의 등록은 거절된다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_SessionBinding_ObservedPidAndStartOnly$' -count=1 -v`
- RED 이유: 요청에 적힌 session id나 pid를 믿거나, 시작 시각을 비교하지 않아 해당 사례가 받아들여진다
- 초록 경로: M1

#### AC-AFI-007 — 한 세션 한 역할 (maps REQ-AFI-005)
- Given 계보 검사를 통과하는 요청자와 유효 ack(유효 ack 주입)
- When 다음 대상에 감사 역할 부여를 요청한다: (R3) 현재 attempt의 Worker 세션, (R4) 같은 run의 이전 attempt에서 Worker였던 세션, (R5) 이미 코드 리뷰 역할을 가진 세션, (R5b) Coordinator 세션. 대조군: 역할이 없고 Worker였던 적이 없는 세션
- Then 대조군만 부여된다. 네 사례는 모두 거절되고 그 세션의 기존 역할(있으면)은 그대로다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_OneRolePerSessionPerRun$' -count=1 -v`
- RED 이유: 현재 attempt의 Worker만 비교하거나, 이미 역할이 있는지 보지 않아 해당 사례가 부여된다
- 초록 경로: M3

#### AC-AFI-008 — 증거 종류와 역할의 일치 (maps REQ-AFI-006)
- Given 다섯 역할(코드 리뷰, 보안 리뷰, 감사, QA, Coordinator)을 서로 다른 다섯 세션에 정상 부여했다. 인증·신원·HEAD는 모두 유효하다
- When (대조군 다섯) 각 세션이 자기 역할의 증거를 등록한다. 교차 사례: (K1) 코드 리뷰 세션이 보안 리뷰 증거 등록, (K2) 보안 리뷰 세션이 코드 리뷰 증거 등록, (K3) 코드 리뷰 세션이 감사 PASS 등록, (K4) 감사 세션이 QA 판정 등록, (K5) QA 세션이 감사 PASS 등록, (K6) Coordinator 세션이 감사 PASS 등록, (K7) 감사 세션이 coordinator 인계 등록. Worker 사례: (W1) 현재 Worker가 코드 리뷰 등록, (W2) 보안 리뷰 등록, (W3) QA 등록, (W4) 이전 attempt Worker였던 세션이 코드 리뷰 등록(이 세션은 역할 분리 때문에 역할이 없으므로, 시험은 역할 분리 검사를 끈 대역으로 이 세션에 코드 리뷰 역할이 있는 상태를 만들어 귀속기의 `was-worker` 검사만 시험한다)
- Then 대조군 다섯은 `attributed`이고 `kind`·`role`이 각각 맞다. K1~K7은 `unattributable{role-kind-mismatch}`, W1~W3은 `unattributable{no-grant}`, W4는 `unattributable{was-worker}`다. 상위 게이트 대역은 이들을 `blocked(producer-not-allowed)`로 기록한다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_EvidenceKindMatchesGrantedRole$' -count=1 -v`
- RED 이유: "Worker가 아닌 어떤 역할이든 있으면" 귀속하는 구현이라 K1~K7이 `attributed`가 된다
- 초록 경로: M3

#### AC-AFI-009 — 상위로 내보내는 읽기 전용 조회 (maps REQ-AFI-007)
- Given AC-AFI-008 대조군과 교차 사례가 등록된 상태
- When 상위 게이트 쪽에서 귀속 조회를 증거마다 한 번씩 부르고, 같은 조회를 두 번 더 부른다
- Then 결과가 AC-AFI-008의 귀속과 같고 매번 같다. 조회 전후로 부여 목록, 세션 결속, claim 저장소 순번이 바뀌지 않는다. 조회 경로에는 부여·등록 호출이 0건이다. 결과 모양이 design §6의 두 갈래 중 하나이며 `unattributable`에는 `reason`이 있다
- 검증: `go test ./internal/agenticfactory/identity/ -run '^TestIdentity_AttributionQuery_ReadOnlyAndStable$' -count=1 -v`
- RED 이유: 조회가 없거나, 조회가 부여·순번을 바꾸거나, 이유 없는 `unattributable`을 돌려준다
- 초록 경로: M4

## §D 변형(mutant) 점검표

아래 표는 각 AC의 **본문 문장**에서 만든 것이다. AC마다 연결된 REQ 조항을 떼어 그 조항만 어기는 그럴듯한 잘못된 구현을 적고, AC의 어느 단언이 그것을 실패시키는지 적었다. **이 표는 문서 위의 추론이다.** 변형이 실제로 실패하는지는 run 단계가 각 시험의 `EXPECTED_RED`로 확인하기 전까지 확인되지 않은 주장이다.

| AC | REQ 조항 | 변형 | AC가 잡는 이유 |
|---|---|---|---|
| AFI-001 | 001 ControlService만 부여 | 저장소의 `role_grant` 기록을 읽어 권한으로 씀 | (R9) S2 증거가 `unattributable{no-grant}`여야 함 |
| AFI-001 | 001 이번 기동 목록만 | 재시작 때 저장소에서 부여를 다시 읽음 | (R10) `grant-from-previous-start` |
| AFI-001 | 001 기본 부여 없음 | 인증이 없으면 기본으로 역할 부여 | (무인증) 사례 |
| AFI-001 | 001 정상 부여는 됨 | 모든 부여 거절 | 대조군 `attributed` |
| AFI-002 | 002 서버 관측 방식만 | 요청 본문의 `method`·`stdin_is_terminal`을 믿음 | (R13) 제어 단말 없는 요청자는 부여 없음 |
| AFI-002 | 002 확인 답 필요 | ack 레코드만 있으면 부여 | (R8a) |
| AFI-002 | 002 만료 | 만료를 보지 않음 | (R8b) |
| AFI-002 | 002 일회성 | 재사용 허용 | (R8c) |
| AFI-002 | 002 대상 결속 | 대상 필드를 비교하지 않음 | (R8d) |
| AFI-002 | 002 게이트 사유 | 부여 실패를 `producer-not-allowed`로 기록 | 게이트 대역이 `operator-ack-missing`이어야 함 |
| AFI-003 | 003 등록 세션 요청 거절 | 요청자 검사가 없고 인증·역할 분리 검사만 있음 | (R1)·(R2)는 대상·ack가 유효하므로 그 검사로는 거절되지 않음 → 부여되어 실패 |
| AFI-003 | 003 등록 세션 요청 거절(단독성) | 요청자 검사가 있지만 역할 분리 뒤에 둠 | (R1') 거절 기록 코드가 `requester-is-session-descendant`여야 함 |
| AFI-003 | 003 정상 요청자는 통과 | 모든 요청 거절 | 대조군 부여 |
| AFI-004 | 003 자손 거절 | 요청자 본인만 등록 세션인지 비교 | (R11a)·(R11b) 자식·손자 |
| AFI-004 | 003 pty·방식 값 무관 | 제어 단말이 있으면 계보 검사 생략 | (R11b)·(R11c) |
| AFI-004 | 003 정상 pty 요청자는 통과 | pty가 붙은 요청 모두 거절 | 대조군(같은 pty 조건) 부여 |
| AFI-005 | 003 판정 불가는 거절 | 읽기 실패한 고리를 건너뜀 | (R12a)·(R12b)·(R12d) |
| AFI-005 | 003 고리마다 시작 시각 검증 | pid만 따라 올라감 | (R12c) 시작 순서 역전 |
| AFI-005 | 003 깊이 상한 | 무한히 걷거나 상한 초과를 통과로 읽음 | (R12e) |
| AFI-005 | 003 기준점 도달 필요 | 부정 조건(자손 아님)만 검사 | (R12f) 고아 프로세스 |
| AFI-005 | 003 정상 사슬은 통과 | 주입 표가 있으면 모두 거절 | 대조군(문제 고리 하나만 정상) |
| AFI-006 | 004 관측 pid 결속 | 요청에 적힌 session id만 믿음 | (R6) |
| AFI-006 | 004 시작 시각 비교 | pid만 비교 | (R7a) |
| AFI-006 | 004 중복 등록 거절 | 같은 session id 재등록 허용 | (R7b)·(R7c) |
| AFI-006 | 004 연결 상대 pid | 요청 본문의 pid를 씀 | (R7d) |
| AFI-007 | 005 현재 Worker 제외 | 검사 없음 | (R3) |
| AFI-007 | 005 이전 attempt Worker 제외 | 현재 attempt만 비교 | (R4) |
| AFI-007 | 005 한 세션 한 역할 | 기존 역할을 보지 않음 | (R5) |
| AFI-007 | 005 Coordinator 제외 | Worker만 제외 | (R5b) |
| AFI-008 | 006 종류 일치 | "Worker가 아닌 어떤 역할이든" 귀속 | K1~K7 |
| AFI-008 | 006 Worker 제외 | 역할만 보고 Worker 이력은 안 봄 | W4 `was-worker` |
| AFI-008 | 006 정상 귀속 | 모두 `unattributable` | 대조군 다섯 |
| AFI-009 | 007 읽기 전용 | 조회가 순번이나 부여를 바꿈 | 전후 불변, 부여·등록 호출 0건 |
| AFI-009 | 007 결과 모양·이유 | 이유 없는 `unattributable` | `reason` 단언 |
| AFI-009 | 007 일관성 | 조회마다 다른 결과 | 세 번 같은 결과 |

**남은 생존 변형(이 AC 집합으로 잡지 못하는 것)**: 운영자 터미널에 키 입력을 주입해 운영자 기준점 아래에서 요청을 만드는 구현 밖 공격은 어느 AC도 잡지 못한다. 이것은 구현의 변형이 아니라 위협 모형의 한계이며 OQ-AFI-1로 남는다. AC-AFI-008 W4는 대역(역할 분리 검사를 끈 상태)을 써야만 만들 수 있다. 그래서 실제 경로에서 `was-worker` 검사가 빠져도, 역할 분리(AC-AFI-007 R4)가 막는 한 결과는 드러나지 않는다. 두 검사를 겹으로 두는 이유다.

## §E 완료 정의

- 9개 기준 모두 이름 붙은 시험의 `EXPECTED_RED` 기록 → `PASS` 기록이 같은 HEAD 계열에서 있다.
- AC-AFI-004·005의 실제 프로세스 사례가 Linux·macOS·Windows CI 각각에서 PASS다. 한 플랫폼이라도 구현이 없으면 그 플랫폼의 요청은 모두 판정 불가로 거절되며, 그 상태는 완료가 아니다.
- 독립 감사 무부채 PASS가 있다.
- 이 SPEC이 `completed`가 되면 상위 SPEC-AGENTIC-FACTORY-001의 잠금 조항(DEC-AF-02)이 풀린다. 이 SPEC은 상위 문서를 고치지 않는다.
