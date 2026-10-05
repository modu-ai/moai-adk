# Compact — SPEC-TODO-IDENTITY-001

DELTA-ID-04: iteration-2 감사 FAIL 0.75는 유지한다. D10–D13 post-audit repair complete. Tier M audit ceiling=2 소진 후 사용자가 informed override를 명시 승인했으므로 audit_verdict: BYPASSED이며 Implementation Kickoff Approval: approved다. Run preparation complete; production not yet implemented. 감사 PASS 또는 구현 완료를 뜻하지 않는다. 최신 승인·세션·serial/Standard 선택은 progress.md, 작업 의존성은 tasks.md, 계약 원문은 spec.md/acceptance.md, 검증 계획은 plan.md를 따른다. draft → in-progress 전이는 manager-develop 소유다.

- **REQ-TID-001**: When 승인된 writer가 프로젝트 또는 카드 identity를 최초 발급하면, the system shall canonical lowercase UUIDv7을 발급하고 기존 표시 ID tNN을 유지한다.

- **REQ-TID-002**: While 기존 미발급 데이터를 읽기 전용으로 조회하는 동안, the system shall project_uuid/card_uuid 필드를 생략하지 않고 null로 반환하며 UUID 발급과 영속 DB·schema·bytes 변경을 수행하지 않는다.

- **REQ-TID-003**: When 승인된 writer가 기존 live/archive 데이터를 처리하면, the system shall 필요한 identity와 카드·실행 연결을 원자적으로 반영하고 생성 또는 쓰기 실패 시 부분 변경이 관측되지 않게 한다.

- **REQ-TID-004**: When 카드가 수정·보관·복원·재개방되면, the system shall 이미 발급된 project/card identity를 유지하고 성공한 Add 반환과 공개 큐 읽기에 동일한 non-null 카드 identity를 제공한다.

- **REQ-TID-005**: When 기존 실행·카드 할당 기록이 성공하면, the system shall 실제 Todo DB의 프로젝트와 카드 identity를 runtime에 연결하고 누락·모호한 카드는 새로 만들거나 추정 연결하지 않고 거절한다.

- **REQ-TID-006**: When writer가 경쟁하거나 지원하지 않는 미래 스키마·퇴역 원본을 만나면, the system shall 성공한 identity와 카드·runtime 변경을 유실 없이 보존하고 허용되지 않는 변경은 오류로 거절한다.

- **REQ-TID-007**: The system shall UUIDv7의 timestamp를 정렬 힌트로만 취급하고 생성·완료·소유권·검증 증거·승인의 권위로 사용하지 않는다.

- **AC-TID-001**: Given 독립 Git/비Git 프로젝트 2개와 프로젝트마다 새 카드 2개 입력. When 각 프로젝트에서 Add를 성공시키고 반환값과 LoadPure JSON을 수집한다. Then project UUID 전체 cardinality는 2, card UUID 전체 cardinality는 4이며 두 집합의 합집합 cardinality는 6이다. 모든 값은 비영 canonical lowercase UUIDv7이고 프로젝트별 Add 반환=LoadPure, t1/t2 및 last_seq=2를 유지한다. (maps REQ-TID-001)

- **AC-TID-002**: Given identity 없는 legacy JSON·SQLite 및 활성 WAL의 live/archive fixture. When LoadPure를 반복한다. Then 해당 project_uuid/card_uuid key가 존재하고 값은 null이며 schema·원본 bytes·DB 생성 여부가 변하지 않는다. 이미 발급된 identity는 그대로 읽고 null로 낮추지 않는다. (maps REQ-TID-002)

- **AC-TID-003**: Given legacy live/archive/runtime fixture와 identity 표·stamp가 없는 별도 지원 SQLite fixture. When UUID 생성 오류, 기존 identity insert trigger 오류, 최초 identity stamp insert 오류를 각각 주입한다. Then 각 fault가 실제 도달하고 오류가 반환되며 record·기존 행·표/stamp 존재 여부가 실행 전과 같고 부분 발급은 0이다. 각 fault를 제거한 같은 fixture에서 재시도하면 non-null UUID를 발급하고 후속 성공 writer가 기존 project/card UUID를 바꾸지 않는다. (maps REQ-TID-003)

- **AC-TID-004**: Given 실제 Add 성공 카드. When 반환 item과 LoadPure를 비교하고 Mutate로 text/state/spec 변경→ArchiveCard→RestoreCard→재개방한다. Then 같은 card_uuid와 project_uuid를 유지하고 기존 카드 내용/상태 전이가 보존된다. (maps REQ-TID-004)

- **AC-TID-005**: Given live/archived 카드, missing 카드 및 같은 local ID가 items와 archived_items 양쪽에 존재하는 fixture. When RecordFactoryRunStart/RecordFactoryCardAssignment/RecordFactoryCardState를 호출하고 LoadPure로 읽는다. Then 정상 run/assignment의 project_uuid/card_uuid는 실제 카드와 같고 missing 또는 후보 2개인 카드는 오류와 record/runtime 변경 0으로 거절한다. 정상 카드에 completed를 보고하면 assignment.ReportedState만 반영되고 카드 State는 그대로이며 whole-record 수정 뒤에도 runtime 연결을 유지한다. (maps REQ-TID-005)

- **AC-TID-006**: Given 같은 Todo DB를 여는 별도 BacklogStore/SQLite handle 2개와 barrier 및 future core/runtime/identity version·retired fixture. When barrier에서 동시 Add를 시작하고 별도로 지원하지 않는 identity reader/writer와 core/runtime·retired writer를 실행한다. Then 성공 2건의 카드 2개·last_seq=2·서로 다른 card UUID 2개와 동일 project UUID를 보존하고 거절 경로는 오류와 영속 데이터/schema 변경 0을 반환한다. race 검사는 process-local 보조 검사로만 사용한다. (maps REQ-TID-006)

- **AC-TID-007**: Given 역순 timestamp의 유효 UUIDv7 카드와 RecordFactoryCardState 후의 카드 AddedAt·State 및 assignment OwnerLabel·ProvenanceJSON·ReportedState snapshot. When UUID만 다른 값으로 구성하고 LoadPure를 반복한다. Then UUID를 정확히 projection하며 해당 카드·assignment 업무 필드는 snapshot과 동일하다. UUID 시간은 completed 보고를 카드 State 완료로 승격하거나 owner/provenance를 덮어쓰지 않는다. (maps REQ-TID-007)

## Exclusions

### Out of Scope — 후속 실행·완료 기능

- run_uuid alias 전환, owner token/generation, slot/resume/write lease, completion receipt/recovery/hooks/update, Graph/UI는 이 child에서 구현하지 않는다.
- locator 이동·폴더 복사 충돌 해소·Git 재연결은 후속 단위다. project UUID를 그 기능 완료로 보고하지 않는다.
- 운영 DB/설치/카드 상태/push/PR/병합/배포를 변경하지 않는다. 기존 첫 저장 child의 검증 순서 예외를 재사용하지 않는다.
- ULID 의존성 도입과 기존 발급 UUID의 임의 재발급은 금지한다.
