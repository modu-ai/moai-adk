# SPEC-AGENTIC-FACTORY-LOCAL-001 — 구현 계획 (0.1.0)

순서는 바뀔 가능성이 큰 결정부터 둔다(데이터 모델·API, 병합 큐, 흐름). 기계적인 단계(퇴역·상수)는 뒤에 둔다. 시간 예측은 쓰지 않는다. 단계 묶음은 리더의 구현 계약 §7(P1~P4)을 따른다.

## §A 맥락

- 원천: 사용자 결정 D1~D4(상위 DEC-AF-30~33)와 리더의 구현 계약(spec §A.1). 기준 커밋 `37b67f6d39`.
- 계약 머리말의 규칙: push 없음, PR 없음, 패키지당 writer 하나, 시험 먼저(RED 커밋이 GREEN보다 먼저), 바뀐 패키지만 확인.
- 원격 게시 주기(D5)는 정해지지 않았다(상위 OD-31). 아무 단계도 push하지 않는다.

## §B 알려진 위험

- 의존 관계의 원천이 계약에 없다(OD-34). Claim의 "의존 카드가 DONE" 조건이 무엇을 읽을지 정해지기 전에는 REQ-AFL-010의 그 부분을 구현할 수 없다.
- `Heartbeat`가 호출자 시각(`nextWake *time.Time`)을 받는 것과 I4(DB 시계)의 관계가 미정이다(OD-35 (f)). REQ-AFL-004는 사건 시각만 DB 시계로 묶었다.
- 퇴역(REQ-AFL-015)의 시작 조건 "새 경로가 green"의 판정 기준이 미정이다(OD-35 (e)).
- 상위 REQ 다섯 개(005·010·013·016·018)는 Issue 의존과 독립 부분이 섞여 처분이 열려 있다(상위 OD-32).

## §C 마일스톤 (바뀔 가능성이 큰 순서)

### M1 — `ledger/api.go` 형식 전용 커밋과 원장 본체 (Priority High, lane-18, 계약 P1)

- 계약 §3 API 형식만 담은 커밋을 먼저 낸다(다른 lane이 이것에 맞춰 짠다).
- 이어 Open·schema·Claim·Move·Heartbeat·Reap·locks·Snapshot과 시험.
- 초록 대상: AC-AFL-001~012, 018.

### M2 — 병합 큐 (Priority High, lane-22, 계약 P1)

- 후보 worktree 검증(잠금 밖) → `merge-queue:landing` → CAS update-ref → `landed` 쪽지 → DONE → Release. 재시작 뒤 git 조상 관계로 맞춤.
- 초록 대상: AC-AFL-014.

### M3 — 흐름 읽기 모형과 보드 (Priority Medium, lane-20 + lane-21, 계약 P1)

- `ledger.Reader`에 맞춘 flow 패키지, 가짜 Reader로 시험하는 읽기 전용 보드.
- 초록 대상: AC-AFL-019, 020.

### M4 — importer dry-run (Priority Medium, lane-22, 계약 P1)

- 원천은 읽기만 한다. dry-run이 먼저.
- 초록 대상: AC-AFL-016.

### M5 — `moai af`, 실제 원장 위 훅 heartbeat, 카드 흐름 통합 (Priority Medium, lane-22 + lane-20, 계약 P2)

- 동사 이름은 OD-33이 정해진 뒤.
- 초록 대상: AC-AFL-013, 017, 021.

### M6 — Issue·원격 코드 퇴역 (Priority Low, lane-22 배선 + lane-18 claim/controlref, 계약 P2)

- 새 경로가 green이 된 뒤. import를 먼저 지우고 패키지는 마지막에, 패키지마다 검토된 커밋 하나.
- 초록 대상: AC-AFL-015.

## §D 제약

- 이 SPEC의 범위 밖(spec §I): 원격 게시, 배포·held 카드 처리(계약 P3·P4), 규칙 문서, 옛 원장 직접 수정.
- 패키지당 writer 하나(계약 §4). lane-19는 모든 커밋을 통합 전에 읽기 전용으로 검토한다.

## §E 자체 확인

- AC마다 시험 이름이 정확히 하나인지, 그 이름이 아직 없을 때 실행 명령을 두지 않았는지.
- RED 커밋과 GREEN 커밋이 따로 있는지(커밋 그래프로).

## §F 상위 SPEC과의 연결

- 상위 SPEC-AGENTIC-FACTORY-001 0.7.0 §J 처분표가 이 SPEC의 REQ-AFL id를 가리킨다. 상위 DEC-AF-06/07/17/18(개정)·DEC-AF-30~33, OD-31~35.
