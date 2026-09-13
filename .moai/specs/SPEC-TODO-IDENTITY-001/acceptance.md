# Acceptance — SPEC-TODO-IDENTITY-001

- **AC-TID-001**: Given 독립 Git/비Git 프로젝트 2개와 프로젝트마다 새 카드 2개 입력. When 각 프로젝트에서 Add를 성공시키고 반환값과 LoadPure JSON을 수집한다. Then project UUID 전체 cardinality는 2, card UUID 전체 cardinality는 4이며 두 집합의 합집합 cardinality는 6이다. 모든 값은 비영 canonical lowercase UUIDv7이고 프로젝트별 Add 반환=LoadPure, t1/t2 및 last_seq=2를 유지한다. (maps REQ-TID-001)

- **AC-TID-002**: Given identity 없는 legacy JSON·SQLite 및 활성 WAL의 live/archive fixture. When LoadPure를 반복한다. Then 해당 project_uuid/card_uuid key가 존재하고 값은 null이며 schema·원본 bytes·DB 생성 여부가 변하지 않는다. 이미 발급된 identity는 그대로 읽고 null로 낮추지 않는다. (maps REQ-TID-002)

- **AC-TID-003**: Given legacy live/archive/runtime fixture와 identity 표·stamp가 없는 별도 지원 SQLite fixture. When UUID 생성 오류, 기존 identity insert trigger 오류, 최초 identity stamp insert 오류를 각각 주입한다. Then 각 fault가 실제 도달하고 오류가 반환되며 record·기존 행·표/stamp 존재 여부가 실행 전과 같고 부분 발급은 0이다. 각 fault를 제거한 같은 fixture에서 재시도하면 non-null UUID를 발급하고 후속 성공 writer가 기존 project/card UUID를 바꾸지 않는다. (maps REQ-TID-003)

- **AC-TID-004**: Given 실제 Add 성공 카드. When 반환 item과 LoadPure를 비교하고 Mutate로 text/state/spec 변경→ArchiveCard→RestoreCard→재개방한다. Then 같은 card_uuid와 project_uuid를 유지하고 기존 카드 내용/상태 전이가 보존된다. (maps REQ-TID-004)

- **AC-TID-005**: Given live/archived 카드, missing 카드 및 같은 local ID가 items와 archived_items 양쪽에 존재하는 fixture. When RecordFactoryRunStart/RecordFactoryCardAssignment/RecordFactoryCardState를 호출하고 LoadPure로 읽는다. Then 정상 run/assignment의 project_uuid/card_uuid는 실제 카드와 같고 missing 또는 후보 2개인 카드는 오류와 record/runtime 변경 0으로 거절한다. 정상 카드에 completed를 보고하면 assignment.ReportedState만 반영되고 카드 State는 그대로이며 whole-record 수정 뒤에도 runtime 연결을 유지한다. (maps REQ-TID-005)

- **AC-TID-006**: Given 같은 Todo DB를 여는 별도 BacklogStore/SQLite handle 2개와 barrier 및 future core/runtime/identity version·retired fixture. When barrier에서 동시 Add를 시작하고 별도로 지원하지 않는 identity reader/writer와 core/runtime·retired writer를 실행한다. Then 성공 2건의 카드 2개·last_seq=2·서로 다른 card UUID 2개와 동일 project UUID를 보존하고 거절 경로는 오류와 영속 데이터/schema 변경 0을 반환한다. race 검사는 process-local 보조 검사로만 사용한다. (maps REQ-TID-006)

- **AC-TID-007**: Given 역순 timestamp의 유효 UUIDv7 카드와 RecordFactoryCardState 후의 카드 AddedAt·State 및 assignment OwnerLabel·ProvenanceJSON·ReportedState snapshot. When UUID만 다른 값으로 구성하고 LoadPure를 반복한다. Then UUID를 정확히 projection하며 해당 카드·assignment 업무 필드는 snapshot과 동일하다. UUID 시간은 completed 보고를 카드 State 완료로 승격하거나 owner/provenance를 덮어쓰지 않는다. (maps REQ-TID-007)

## DELTA-ID-03 — 관측 범위와 품질

Tier M plan audit ceiling=2가 소진되었다. iteration-2 FAIL 0.75 이후 D10–D13 post-audit repair complete이며 independent re-audit unavailable이다. Explicit informed user override 및 formal Implementation Kickoff Approval approved. audit_verdict는 BYPASSED이며 감사 FAIL 원문은 보존한다. 구현/PASS/in-progress가 아니다.

최신 [RED baseline](../../reports/SPEC-TODO-IDENTITY-001/red-baseline.md)의 7개 selector는 1/3/3/1/3/3/1개, 총 15개 테스트를 실행했고 모두 exit 1이다. D10 전역 cardinality, D11 세 fault와 오류 제거 재시도·반복 writer, D12 실제 후보2개 거절 및 state update, D13 기존 assignment OwnerLabel/ProvenanceJSON deep equality assertion이 추가되어 실행됐다. field 부재·fault reached=0·ambiguous 허용으로 실패했으며 GREEN/rollback 성공을 뜻하지 않는다. 최신 hash/원문/명령은 plan.md E에 귀속한다. 기존 missing/retired 2개 guard PASS와 race semantic RED를 별도로 유지한다.

모든7AC가 종료필수다. 신규/변경 로직85% 이상, 명시된 실패 분기 실행과 race/범위 회귀, 명령의 실제 실행 개수와 원문출력을 보존한다. baseline/source파일hash는 실행시점에 고정한다. 기존 지원 동작의 회귀PASS는 신규기능RED를 대신하지 않는다. 실제 schema/backfill/생성실패 조건이 미도달이면 GAP이며 감사결정 없이 강등하지 않는다.

## Given fixture 안전

t.TempDir와 분리한 MOAI_HOME만 사용한다. subprocess cleanup과 bounded timeout을 등록한다. live/archive/runtime 값을 실제로 구성하고 실패지점 도달count를 검사한다. 성공값 없는 equality test를 건너뛰고 전체PASS로 표시하지 않는다. backup 파일 복원 테스트는 현재 없으며 이 child의 AC003은 transaction rollback으로 한정한다. 별도 backup-file restore 검증은 운영 migration 전 필요한 비차단 후속 작업이며 이 child에서 검증했다고 주장하지 않는다.
