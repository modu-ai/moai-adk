# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 구현 계획 (0.2.0)

> Tier M(spec.md, plan.md, acceptance.md)에 design.md를 더했다. 검사 절차가 보안 계약의 본체라 따로 적었다(spec-frontmatter-schema § Artifact Statelessness: Tier와 무관하게 허용). research.md는 두지 않는다. 이 트리에서 읽은 사실은 spec.md §B에 있다. 시간 추정은 쓰지 않는다. 순서와 우선순위만 적는다.

## §A 맥락

- 상위 SPEC-AGENTIC-FACTORY-001의 병합 게이트는 리뷰·감사·QA·Coordinator 증거를 이 SPEC의 귀속 조회(REQ-AFI-007, design §8)로만 판정한다. 귀속 실패는 상위에서 모두 `producer-not-allowed` 하나로 간다(상위 design §2.1b).
- 상위 잠금(상위 spec §C.4, design §11.2)은 컴파일 시점 상수다. 이 SPEC의 `completed`만으로는 풀리지 않는다. 해제는 상위 M7에서 lane-18이 상수 한 곳을 바꾸는 코드 변경으로 하며, 이 SPEC의 `completed`와 상위 OD-4(운영자 위험 수용) 기록이 그 선행 조건이다.
- 권위 원천은 control ref와 Issue다. 부여 기록은 control ref에만 있고 로컬 캐시는 재구축 전용이다(REQ-AFI-001).
- 대상 패키지 `internal/agenticfactory/identity/`는 아직 없다(acceptance §B EV-AFI-REF-0, 0.2.0 트리에서 재측정 안 함).

## §B 알려진 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| 연결 상대 pid를 OS에서 얻는 방법이 이 트리에 없음(OQ-AFI-2) | REQ-AFI-003·004 전체가 막힘 | M2 착수 전 연결 방식을 정한다. 정하지 못하면 M2는 시작하지 않는다 |
| Linux `ps lstart`는 1초 해상도라 같은 초 PID 재사용을 못 잡음 | 결속 검사가 같은 프로세스로 오인 | 알려진 한계로 단언(AC-AFI-006 B5), 기준점 사슬 검사가 보완. 순서 비교는 darwin·Windows에서만 |
| 고아·입양 프로세스 | 기준점 밖 요청이 통과할 위험 | 모든 플랫폼에서 `requester-outside-operator-anchor` 하나(design §3.2 순서표) |
| 실제 프로세스를 띄우는 시험의 잔여 프로세스 | CI 부하, 다른 시험 간섭 | `t.Cleanup` 종료 등록 또는 `timeout` 래퍼. 백그라운드 부하 금지 |
| 기준점 밖 중계 경로(비자손 중계, OQ-AFI-1) | 계보 검사로 막히지 않음 | 잔여 위험으로 기록(spec §F.3). 상위 OD-4 수용 기록 없이는 `completed` 불가 |
| 제어 단말 읽기의 플랫폼 차이 | 서버 관측 인증이 플랫폼마다 다르게 동작 | 구현이 없는 플랫폼은 `no-controlling-terminal`로 거절(기본 부여 없음) |
| control ref 쓰기가 상위 M1에 묶임 | 부여 기록을 쓸 곳이 없음 | M1은 상위 M1의 CAS 쓰기 뒤에 착수 |

## §C 마일스톤 (결정이 뒤집힐 가능성이 큰 순서)

### M1 — control ref 부여 권한과 세션 등록 (Priority High, lane-18)

- 부여 기록은 control ref에만, ControlService만 쓴다. 로컬 캐시는 권위가 아니다(REQ-AFI-001, design §2.2).
- 재시작 재구축: 부여는 버리지 않고, 결속을 재검증해 실패하면 `grant-reverify-failed`로 표시(design §7).
- 등록기: 관측 pid와 fingerprint로 결속, 증거 등록마다 재관측(REQ-AFI-004, design §2.1). Coordinator·agent는 등록만 되고 역할 부여 대상이 아니다.
- 선행: 상위 M1의 control ref CAS 쓰기.
- 초록: AC-AFI-001, AC-AFI-006.

### M2 — 요청자 계보와 운영자 기준점 (Priority High, lane-18)

- OQ-AFI-2 결정(연결 방식, 연결 상대 pid 원천). 본문 pid는 판정 근거가 아니다(iteration-1 D6).
- 운영자 기준점 하나(design §3.1), 사슬 걷기와 순서표(design §3.2), 플랫폼 표(design §3.3).
- 판정 불가는 거절(`requester-ancestry-indeterminate`).
- 초록: AC-AFI-003, AC-AFI-004, AC-AFI-005.

### M3 — 서버 관측 운영자 확인 (Priority High, lane-18)

- 요청자의 제어 단말은 서버가 프로세스 표에서 관측한다. 운영자 답은 ControlService 자신의 단말에서만 받는다(design §4, iteration-1 D1).
- binding의 결속(session id·역할·Issue)·일회성·만료.
- 초록: AC-AFI-002.

### M4 — 역할 배제·귀속 행렬·읽기 전용 조회 (Priority Medium, lane-18 + lane-22)

- Issue 단위 `was-writer`와 `role-already-held`(REQ-AFI-005).
- 생산자 7 × 종류 6 행렬과 미등록 6, 생성 사례 48(design §5, iteration-1 D7).
- 읽기 전용 귀속 조회와 닫힌 사유(design §6.2, §8). 상위 대역과 같은 모양인지 대조.
- 초록: AC-AFI-007, AC-AFI-008, AC-AFI-009.

### M5 — 완료 관문 (Priority High, 운영자 결정 필요)

- 선행: AC-AFI-001~009 같은 HEAD PASS, 독립 감사 무부채 PASS(새 감사는 운영자 승인 대상), 상위 OD-4 운영자 위험 수용 기록.
- 이 SPEC은 상위 잠금 상수를 바꾸지 않는다. 해제는 상위 M7(lane-18)이며, 같은 착지에서 상위 AC-AF-011 [REF]의 L1·L2를 해제 뒤 형태로 다시 쓴다.
- 초록 대상 없음(관문).

## §D 제약

- 기존 코드(`internal/session`, `internal/homestate`, `internal/contract`)는 고치지 않는다. 재사용은 호출만 한다.
- 시험은 범위를 좁혀 돌린다(`go test ./internal/agenticfactory/identity/ -run '^...$'`). 전체 저장소 시험은 CI가 돌린다.
- M3 전 사례는 시험 경계(유효 확인 주입)로 운영자 확인을 대신한다. 실제 운영 경로의 확인은 design §4로만 생긴다.
- 어느 AC도 그것이 단언하는 것을 만드는 마일스톤보다 먼저 초록이 되지 않는다(acceptance §A).

## §E 자체 확인

- 9개 AC 각각 `EXPECTED_RED` → `PASS` 기록(acceptance §A 규칙).
- AC-AFI-003~005 실제 프로세스 사례는 Linux·macOS·Windows CI에서 PASS.
- AC-AFI-008 생성 사례 수 48 단언.
- 변형 점검표(acceptance §D)의 각 행이 RED 기록으로 확인되었는지 표시.

## §F 상위 SPEC과의 연결

- 상위가 받는 것: 귀속 조회와 이유 값 목록(design §6.2). 이 모양이 바뀌면 상위의 게이트 단위 대역과 상위 AC-AF-011 [REF] 사례도 바뀌어야 한다.
- 상위 잠금 해제: 상위 M7, lane-18. 선행 조건은 이 SPEC `completed` + 상위 OD-4. 이 SPEC은 상위 문서를 고치지 않는다.
