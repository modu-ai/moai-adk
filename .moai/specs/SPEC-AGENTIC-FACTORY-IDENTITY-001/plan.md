# SPEC-AGENTIC-FACTORY-IDENTITY-001 — 구현 계획

> Tier M(spec.md, plan.md, acceptance.md)에 design.md를 더했다. 검사 절차가 보안 계약의 본체라 따로 적었다(spec-frontmatter-schema § Artifact Statelessness: Tier와 무관하게 허용). research.md는 두지 않는다. 이 트리에서 읽은 사실은 spec.md §B에 있다. 시간 추정은 쓰지 않는다. 순서와 우선순위만 적는다.

## §A 맥락

- 상위 SPEC-AGENTIC-FACTORY-001의 게이트는 리뷰·감사·QA·coordinator 증거를 이 SPEC의 귀속 조회(REQ-AFI-007, design §6)로만 판정한다. 이 SPEC이 `completed`가 되기 전에는 상위의 잠금 조항(상위 §C.5.1 DEC-AF-02)에 따라 그 증거가 모두 귀속 불가로 막힌다.
- 대상 패키지 `internal/agenticfactory/identity/`는 아직 없다(acceptance §B EV-AFI-REF-0).

## §B 알려진 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| 연결 상대 pid를 OS에서 얻는 방법이 이 트리에 없음(OQ-AFI-2) | REQ-AFI-003·004 전체가 막힘 | M1 착수 전 연결 방식을 정한다. 정하지 못하면 M1은 시작하지 않는다 |
| Windows의 pid 재사용과 부모 정보 | 계보가 엉뚱한 프로세스로 이어짐 | 고리마다 시작 시각 비교(design §3.2-3), 어기면 판정 불가 |
| 실제 프로세스를 띄우는 시험의 잔여 프로세스 | CI 부하, 다른 시험 간섭 | `t.Cleanup` 종료 등록 또는 `timeout` 래퍼. 백그라운드 부하 금지 |
| 운영자 터미널 입력 주입(OQ-AFI-1) | 계보 검사로 막히지 않음 | 잔여 위험으로 기록. 이 SPEC은 더 강한 인증을 구현하지 않음 |
| 제어 단말 읽기의 플랫폼 차이 | 서버 관측 인증이 플랫폼마다 다르게 동작 | 구현이 없는 플랫폼은 거절(기본 부여 없음) |

## §C 마일스톤 (결정이 뒤집힐 가능성이 큰 순서)

### M1 — 연결 방식과 세션 신원 결속 (Priority High)
- OQ-AFI-2 결정(연결 방식, 연결 상대 pid 원천). 이것이 바뀌면 이후 마일스톤이 모두 바뀐다.
- 등록기: 관측 pid와 시작 시각으로 결속, 증거 등록마다 재관측(REQ-AFI-004).
- 초록: AC-AFI-006.

### M2 — 계보 검사와 부여 권한 (Priority High)
- 운영자 기준점 기록(OQ-AFI-3 결정 포함), 사슬 걷기, 판정 불가 처리(design §3).
- 부여기의 권한 경계: ControlService 이번 기동 목록만(REQ-AFI-001).
- 초록: AC-AFI-001, AC-AFI-003, AC-AFI-004, AC-AFI-005.

### M3 — 운영자 인증, 역할 분리, 종류 일치 (Priority High)
- 서버 관측 제어 단말, 확인 절차, ack의 결속·일회성·만료(REQ-AFI-002).
- 한 세션 한 역할(REQ-AFI-005), 증거 종류와 역할 일치(REQ-AFI-006).
- 초록: AC-AFI-002, AC-AFI-007, AC-AFI-008.

### M4 — 상위로 내보내는 조회 (Priority Medium)
- 읽기 전용 귀속 조회와 결과 모양(REQ-AFI-007, design §6). 상위의 대역(stub)과 같은 모양인지 확인.
- 초록: AC-AFI-009.

## §D 제약

- 기존 코드(`internal/session`, `internal/homestate`, `internal/contract`)는 고치지 않는다. 재사용은 호출만 한다.
- 시험은 범위를 좁혀 돌린다(`go test ./internal/agenticfactory/identity/ -run '^...$'`). 전체 저장소 시험은 CI가 돌린다.
- 운영자 확인이 필요한 사례는 시험 경계(유효 ack 주입)로 대신한다. 실제 운영자 대화 경로의 대기 한도는 OQ-AFI-1에 남는다.

## §E 자체 확인

- 9개 AC 각각 `EXPECTED_RED` → `PASS` 기록(acceptance §A 규칙).
- AC-AFI-004·005 실제 프로세스 사례는 Linux·macOS·Windows CI에서 PASS.
- 변형 점검표(acceptance §D)의 각 행이 RED 기록으로 확인되었는지 표시.

## §F 상위 SPEC과의 연결

- 상위가 받는 것: `Attribute(evidence_id)` 조회와 이유 값 목록(design §6). 이 모양이 바뀌면 상위의 대역과 AC-AF-018 게이트 사례도 바뀌어야 한다.
- 상위의 잠금 해제: 이 SPEC의 `completed`(acceptance §E). 이 SPEC은 상위 문서를 고치지 않는다.
