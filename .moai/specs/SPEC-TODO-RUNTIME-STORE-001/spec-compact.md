# Compact — SPEC-TODO-RUNTIME-STORE-001

spec.md/acceptance.md 파생 참고본. API·변경 파일은 plan.md.

- **REQ-TRS-001**: When 기존 실행 시작 또는 카드 할당·상태 기록 요청이 성공하면, the system shall 그 실행과 할당의 실제 값을 프로젝트 Todo DB에 보존하고 별도 Factory 실행·카드 기록을 새로 작성하지 않는다.
- **REQ-TRS-002**: When 공개 읽기 전용 큐 조회가 실행되면, the system shall 카드와 실행·할당 기록을 같은 스냅숏에서 반환하고 영속 상태를 변경하지 않는다.
- **REQ-TRS-003**: When 실행 기록과 기존 카드 변경이 경쟁하거나 기록 중 오류가 발생하면, the system shall 성공한 기존 카드·실행 변경을 보존하고 실패 요청의 부분 기록을 남기지 않는다.
- **REQ-TRS-004**: When 구 큐 또는 지원하지 않는 스키마를 열면, the system shall 지원 구 큐의 카드 내용을 보존하며 확장을 원자적으로 적용하고 지원하지 않는 스키마·퇴역 원본에는 쓰지 않는다.

- **AC-TRS-001**: Given 실제 Git 프로젝트와 Git 없는 일반 폴더의 live 카드·보관 카드와 빈 legacy Factory DB 및 SPEC 없는 경우를 포함한 run/card 원본 인수. When 기존 RecordFactoryRunStart, RecordFactoryCardAssignment, RecordFactoryCardState를 실제로 호출하고 공개 LoadPure 결과를 JSON으로 읽는다. Then runtime.runs에 실제 run_id/backend/manifest가 있고 runtime.assignments에 실제 run_id/card_id/owner_label/reported_state/event_kind/provenance가 있다. legacy runs/cards 수는 증가하지 않고 live/archive membership은 원래대로다. 같은 run·card 재기록은 중복 행 대신 최신 기록 하나이며 run 없이 assignment를 요청하면 실행 seed와 assignment가 함께 기록된다. (maps REQ-TRS-001)
- **AC-TRS-002**: Given 실행과 할당이 있는 checkpointed DB 및 활성 WAL fixture. When 기존 LoadPure를 반복하고 JSON 카드·runtime 및 영속 DB/schema를 비교한다. Then 모든 실제 값이 같은 snapshot에 있으며 읽기 전후 영속 내용이 불변이다. runtime 없는 구 큐는 빈 runs/assignments를 반환하고 읽기만으로 확장 DDL을 실행하지 않는다. (maps REQ-TRS-002)
- **AC-TRS-003**: Given 실제 실행 기록과 카드가 있는 DB 및 assignment 쓰기를 중단시키는 SQLite trigger fixture. When 기존 Mutate로 카드 text를 바꾸고 기존 RecordFactoryCardAssignment와 경쟁시키거나 실행 seed 뒤 assignment insert를 abort한다. Then 성공한 카드 text와 실행 기록이 모두 남고 abort한 요청의 implicit run/assignment는 둘 다 남지 않으며 오류가 호출자에게 반환된다. 기록 후 기존 whole-record write를 다시 실행해도 runtime은 삭제되지 않는다. (maps REQ-TRS-003)
- **AC-TRS-004**: Given core schema1의 기존 큐, 미래 core/확장 version, 퇴역 원본, 확장 설치 실패 fixture. When 기존 runtime 기록 API 및 LoadPure를 실행한다. Then 구 큐 카드/last_seq/findings/archive를 보존하고 확장은 전체 전후 중 하나다. 미래 version·퇴역 writer는 오류이며 영속 DB bytes를 바꾸지 않는다. Factory meta/table을 Todo DB에 복사하지 않는다. (maps REQ-TRS-004)

## Exclusions

### Out of Scope — 후속 기능과 운영 이전

- UUID backfill, 소유권 CAS, slot/resume 이동, 자동 완료/receipt, Graph, 웹 시각화는 후속 child다.
- 기록의 reported_state를 카드 lifecycle 권위로 사용하지 않는다. legacy completed를 카드 완료로 바꾸지 않는다.
- 기존 Factory 과거 이력의 bulk migration, 운영 DB 적용/설치/push/PR/병합/배포를 실행하지 않는다.
