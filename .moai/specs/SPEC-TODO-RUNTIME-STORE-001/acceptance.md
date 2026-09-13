# Acceptance — SPEC-TODO-RUNTIME-STORE-001

- **AC-TRS-001**: Given 실제 Git 프로젝트와 Git 없는 일반 폴더의 live 카드·보관 카드와 빈 legacy Factory DB 및 SPEC 없는 경우를 포함한 run/card 원본 인수. When 기존 RecordFactoryRunStart, RecordFactoryCardAssignment, RecordFactoryCardState를 실제로 호출하고 공개 LoadPure 결과를 JSON으로 읽는다. Then runtime.runs에 실제 run_id/backend/manifest가 있고 runtime.assignments에 실제 run_id/card_id/owner_label/reported_state/event_kind/provenance가 있다. legacy runs/cards 수는 증가하지 않고 live/archive membership은 원래대로다. 같은 run·card 재기록은 중복 행 대신 최신 기록 하나이며 run 없이 assignment를 요청하면 실행 seed와 assignment가 함께 기록된다. (maps REQ-TRS-001)
- **AC-TRS-002**: Given 실행과 할당이 있는 checkpointed DB 및 활성 WAL fixture. When 기존 LoadPure를 반복하고 JSON 카드·runtime 및 영속 DB/schema를 비교한다. Then 모든 실제 값이 같은 snapshot에 있으며 읽기 전후 영속 내용이 불변이다. runtime 없는 구 큐는 빈 runs/assignments를 반환하고 읽기만으로 확장 DDL을 실행하지 않는다. (maps REQ-TRS-002)
- **AC-TRS-003**: Given 실제 실행 기록과 카드가 있는 DB 및 assignment 쓰기를 중단시키는 SQLite trigger fixture. When 기존 Mutate로 카드 text를 바꾸고 기존 RecordFactoryCardAssignment와 경쟁시키거나 실행 seed 뒤 assignment insert를 abort한다. Then 성공한 카드 text와 실행 기록이 모두 남고 abort한 요청의 implicit run/assignment는 둘 다 남지 않으며 오류가 호출자에게 반환된다. 기록 후 기존 whole-record write를 다시 실행해도 runtime은 삭제되지 않는다. (maps REQ-TRS-003)
- **AC-TRS-004**: Given core schema1의 기존 큐, 미래 core/확장 version, 퇴역 원본, 확장 설치 실패 fixture. When 기존 runtime 기록 API 및 LoadPure를 실행한다. Then 구 큐 카드/last_seq/findings/archive를 보존하고 확장은 전체 전후 중 하나다. 미래 version·퇴역 writer는 오류이며 영속 DB bytes를 바꾸지 않는다. Factory meta/table을 Todo DB에 복사하지 않는다. (maps REQ-TRS-004)

## RED ledger 상태

**현재 승인 상태:** 사용자가 2026-09-12 “승인!!!”으로 [단계별 검증 순서](../../reports/t648/staged-verification-approval.md)를 명시 승인했다. 내부 저장 기반에 한해 verification-order를 narrowly BYPASSED한 것이며 AC·85% 품질·권한·이전 안전 의무는 면제하지 않는다. 아래 ledger의 “예외 미승인”과 “미관측”은 승인 전 해당 측정 시점의 기록이다. 현재 안전성 검증은 진행 중이며 이 문구 수정은 AC 완료나 감사 PASS 판정이 아니다.

승인 전 추가 실측 기록은 ../../reports/t648/runtime-store-red.md를 참조한다. 당시 public readback 부재·legacy runs/cards 증가는 AC001 일부의 실제 RED이며 4AC 전체 채택이 아니었다. assignment transaction rollback/경쟁과 신규 schema 설치 원자성도 당시 미관측이었다. 당시에는 plan gate 예외가 승인되지 않았으므로 첫 child를 착수 가능 PASS로 표시하지 않았다. 이후 검증 순서 승인은 위 현재 승인 상태를 따른다.

현재 .moai/reports/t648/red-baseline.md의 TestTodoUnifiedFactoryAssignmentDoesNotWriteLegacyCardMirror는 AC-TRS-001 중 legacy cards 미작성 조건만 관측한 semantic RED다. runs 미작성/공개 runtime positive readback 및 AC002–004는 미채택·UNVERIFIED다. 전체 AC001로 확대하지 않는다. 실행 담당이 명령·원문 stdout·exit·HEAD·실행개수를 채운 후 독립 감사한다.

기존 pure reader/future schema/retired/whole-record 회귀는 보존 검사다. 신규 extension/실행 값 검사와 구분하고 신규 RED 대신 사용하지 않는다. 모든 4 AC는 필수이며 기존 회귀라는 이유로 종료 조건에서 제외하지 않는다.

## 실행 조건

실제 Git·카드·DB fixture, t.TempDir, scrub된 환경, cleanup 보장 자식 프로세스만 사용한다. timeout·컴파일·0test·trigger설치 실패는 semantic RED가 아니다. 변경 로직 coverage와 race 및 실제 public API 오류 전파를 측정한다. full suite/운영전환 성공을 scoped검사로 주장하지 않는다.
