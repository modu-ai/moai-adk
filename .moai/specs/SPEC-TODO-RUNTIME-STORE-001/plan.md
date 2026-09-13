# Plan — SPEC-TODO-RUNTIME-STORE-001

## 범위와 게이트

2026-09-12 운영자 승인: 직전 root의 “내부 저장 기반 먼저 → 해당 baseline의 실패 검증 → 통과 후 외부 연결” 검증 순서 예외 질문에 사용자가 **“승인!!!”**이라고 명시 응답했다. t648 내부 저장 기반에 한해 verification-order gate를 narrowly BYPASSED 처리한다. 기존 감사 FAIL은 그대로이며 PASS로 변경하지 않는다. 상세 범위·금지·만료는 ../../reports/t648/staged-verification-approval.md를 따른다.

t648 단일 카드의 첫 persistence slice다. umbrella SPEC-TODO-UNIFIED-001의 AC001/003/011/014/015 일부를 담당하며 전체 충족으로 대체하지 않는다. 이 child의 4 AC 각각 독립 plan gate를 충족해야 구현 가능하다. 실제 RED가 없는 조건은 미채택/UNVERIFIED며 실행 담당이 채운 후 재감사한다.

## 고정 API/DTO 계약

새 함수 존재를 컴파일 조건으로 삼지 않는다. 현재 존재하는 아래 함수의 인수·반환 signature를 보존한다.

- RecordFactoryRunStart(root, runID, backend, specID string) error
- RecordFactoryCardAssignment(root, runID, cardID, owner, specID string) error
- RecordFactoryCardState(root, runID, cardID, owner, specID, state, eventKind string) error
- BacklogStore.LoadPure() (*BacklogRecord, error)
- BacklogStore.Mutate(func(*BacklogRecord) error) error

실제 BacklogPathForRoot/EnginePath가 가리키는 DB를 사용한다. RecordFactoryCardAssignment가 공유하는 RecordFactoryCardState도 함께 전환하여 다른 이벤트가 legacy mirror를 다시 만들지 않는다. factory_runtime.go의 호출자 signature는 변경하지 않는다. 저장 오류를 무시하는 CLI adapter 수정은 후속 lifecycle child의 AC011까지 추적하며 이 child를 그 adapter 전환 완료라고 보고하지 않는다.

LoadPure의 기존 JSON에 runtime 객체를 추가한다. tests는 기존 API→json.Marshal→map/DTO readback을 사용해 신규 Go 타입 없이 RED를 작성할 수 있다.

```json
{"runtime":{"runs":[{"run_id":"fixture-run","backend":"claude","manifest_json":"{}"}],"assignments":[{"run_id":"fixture-run","card_id":"t1","owner_label":"worker-1","reported_state":"picked","event_kind":"card.assigned","provenance_json":"{}"}]}}
```

manifest_json/provenance_json은 기존 captureFactoryProvenance 결과의 JSON 문자열이며 실제 spec_id/spec_sha256/git_commit 값을 검사한다. run 시작은 card-level SPEC snapshot을 주장하지 않는 기존 계약을 보존한다. runs는 run_id, assignments는 run_id+card_id 정렬이다. runtime이 없으면 두 배열 모두 빈 배열이다. 동일 key 재호출은 upsert이며 immutable 완료 이력이 아니다. 이것은 실행 seed/최신 할당 기록이며 owner 권한 토큰이나 최종 완료 증거가 아니다.

기존 기록 API가 run 없이 card state를 받으면 기존과 같이 암묵 실행 seed를 만든다. live 또는 archive의 해당 card_id가 없으면 오류로 거절하고 run도 남기지 않는다. root별 DB가 프로젝트 경계이며 UUID는 후속 child에서 도입한다.

## SQLite 구현 선택과 안전

- Todo core meta.schema_version=1은 유지하고 별도 key runtime_schema_version=1을 사용한다. 미래 core/확장 version은 DDL 전에 거절한다. Factory의 meta나 runs/cards 테이블을 가져오지 않는다.
- 최소 두 namespaced 테이블 todo_runtime_runs, todo_runtime_assignments. key는 각각 run_id와 (run_id, card_id). 호환 reader는 확장 부재를 빈 배열로 취급한다.
- BoardLock은 기존 Mutate와 동일한 것을 사용한다. runtime-only write도 같은 lock 안에서 SQL transaction을 연다.
- MaxOpenConns(1) 환경에서는 transaction이 연결을 보유한 동안 별도 db.Query를 호출하지 않는다. 같은 tx.Query/Exec로 카드 존재·암묵 run·assignment를 처리한다. rows를 닫기 전에 후속 DB query를 요청하지 않는다.
- 기존 writeRecordArchive가 runtime을 delete/rewrite하지 않도록 별도 테이블 보존. readRecord의 snapshot에 runtime SELECT를 포함한다. 기존 card editor가 들고 온 오래된 runtime 사본을 다시 덮어쓰지 않는다.
- 확장 DDL/version stamp는 한 transaction. 외부 Git·provenance 수집은 lock 밖에서 실행한다. 단순 경로 변경으로 FactoryDB를 Todo에 연결하지 않는다.
- mixed old/new binary가 실행 기록 통일을 보장한다고 주장하지 않는다. 기존 Factory history는 보존하되 신규 writer가 mirror를 추가하지 않는다. 운영 구 바이너리 중단/이전은 umbrella 후속 운영 점검이다.

## RED 준비 및 오류 주입

2026-09-12 최신 test 담당 ledger: runtime-store-red.md의 추가 계약 검사에서 최상위6개 중 semantic RED4/guard PASS2, exit1, 패키지5.398s를 보고했다. test SHA08110f95f366d174d9bbf68987fd843ca7ce10c682690732ccf972c3c18ce7a7. 실제 noGit/noSPEC implicit assignment 성공 후 runtime 없음, 구 큐 빈 runtime 없음, 미래core/extension999에서도 runtime API 성공을 관측했다. AC001 비Git 입력, AC002 빈 runtime, AC004 지원하지 않는 version 거절의 일부 조건에만 연결한다. assignment abort·activeWAL snapshot·extension 설치rollback은 GAP이다.

### 착수 blocker와 결정 경계

첫 child를 다시 이름만 나누어 transaction 안전성 의무를 피하지 않는다. 실행 기록을 쓰는 토대 자체에도 원자성 검증이 필요하므로 새 child 발행만으로 해결되지 않는다. 현 정책상 아직 미구현인 Todo transaction에 오류를 주입해 현재 baseline에서 rollback RED를 관측할 수 없다. 이는 실제 미해소 착수 blocker다.

운영자 결정이 필요한 구체적 대안은 검증 시점의 제한적 변경이다: 내부 저장 기반만 먼저 구현하고, 그 baseline에서 assignment rollback·activeWAL snapshot·동시성의 실제 실패 주입/RED→GREEN을 수행한다. 전부 통과하기 전 CLI/hooks 자동완료 연결과 운영전환은 금지한다. 전체23AC·기존85% 품질 기준·권한·이전 안전 의무는 그대로다. 이는 현재 승인된 예외가 아니며 root가 별도 명시 승인을 받아야 한다. 승인 없이는 생산 구현을 시작하지 않는다.

현재 실측 연결: ../../reports/t648/runtime-store-red.md의 TestTodoRuntimeStorePublicReadbackSurvivesCardEdit는 AC001 actual API 성공→public runtime 부재와 legacy runs/cards 증가의 semantic RED다. edit 전후 확인은 AC003의 보존 필요성을 보여주지만 runtime이 처음부터 없어 보존 성공/rollback RED가 아니다. 미래core bytes와 기존 callback refusal PASS는 보존 guard다. AC003 assignment abort·경쟁 및 AC004 신규extension 설치원자성은 미채택/UNVERIFIED다. 이 child의 전체 gate는 아직 충족되지 않았다.

비Git 추가 검사: 기존 API의 root는 일반 folder도 유효 입력으로 검사하며 Git 없는 경우 manifest의 빈 git_commit을 완료 근거로 사용하지 않는다. 명시 Git capability/정책/UUID/write lease는 후속 공통 실행 단위이며 이 child에 미구현 전제조건으로 끼워넣지 않는다. 현재 test 담당 todo_runtime_red_contract가 actual nonGit/noSPEC/implicit run·구 큐 빈 runtime·미래 확장 거절을 측정 중이다. 실제 ledger 도착 전에는 결과를 발명하지 않는다.

AC001: 실제 Git/spec/card fixture→기존 API 성공→LoadPure JSON에서 실제 인수값 확인→legacy DB의 runs/cards count 변화0 확인. 현재 runtime key 부재를 관측할 수 있지만 table 존재 assertion만으로 대체하지 않는다.
AC002: 같은 실제 입력 이후 공개 JSON readback 및 WAL committed 값을 비교한다. 신규 내용 존재를 확인하기 전에 byte불변만 PASS해도 AC전체충족이 아니다.
AC003: 기존 Mutate로 카드 변경→공개runtime 재조회. rollback은 확장 설치 후 test-only SQL trigger를 todo_runtime_assignments BEFORE INSERT에 설정하여 RAISE(ABORT)하고 기존 assignment API에 새 runID를 전달한다. 오류 도달 count와 새 run/assignment 둘다0을 확인한다. 현재 확장 미존재로 trigger 설치가 실패하면 setup실패/미채택이지 semanticRED가 아니다.
AC004: 기존 미래core/retired 회귀와 신규 extension install/rollback을 분리하여 실행한다. 미래extension key를 fixture에 심고 기존 API가 거절하는지 실제값으로 검사한다. 원본 bytes/카드내용 보존도 비교한다.

## 단계와 소유 파일

High A: 4 AC의 유효 fixture/실제 RED→독립 plan audit. High B: namespaced schema+공통lock/txn+pure DTO. High C: 기존 Factory 기록 API 전환→회귀/race/rollback→독립 sync audit.
생산 경계는 internal/kanban/factory_runtime.go, backlog_store.go, backlog_sqlite.go 및 기존 read/write 분할 파일이다. 대응 테스트 외 CLI/hook/UI를 추가 수정하지 않는다. 기존 homestate.RecordRun/RecordCard는 telemetry/과거 호환 API로 남길 수 있으나 이 생산 경로에서 호출하지 않는다.
