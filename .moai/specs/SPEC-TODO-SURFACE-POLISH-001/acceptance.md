# acceptance.md — SPEC-TODO-SURFACE-POLISH-001 (카드 t1349)

모든 AC는 `maps REQ-...` 정규형으로 REQ에 대응하고 실행 가능한 검증 명령을 가진다.
RED-now 셀(②⑥)은 run 단계 첫 커밋에서 명령·출력·exit code·트리 SHA를 채워 §E.2에
고정한다(verification-completeness §2 두 칸 채택). fixture 스토어는 전부
`t.TempDir()` + 홈 시임 주입 — 실제 홈 DB·primary `.moai/` 는 대상이 아니다.
모든 `-run` 패턴은 양 끝 앵커 완결(`^...$`) — 부분 앵커는 더 긴 이름까지 골라
녹색을 위장한다(spec-lint VacuousTestAssertion).

## §D AC Matrix

- **AC-TSP-001** (maps REQ-TSP-001, REQ-TSP-002 — show 기본·부재·전문, 3 시나리오)
  - **AC-TSP-001a**: Given queued 카드 1장인 fixture 큐, When
    `moai todo show <id>`, Then stdout 한 줄에 id·`live`·상태·landing·스탬프·
    **전문 본문**(탭/개행 평탄화, 길이 보존)이 순서대로 있다.
  - **AC-TSP-001b**: Given 없는 id, When `moai todo show t99999`, Then stdout은
    `t99999\tabsent` 이고 last_seq 이하이면 stderr 보충 한 줄이 있다.
  - **AC-TSP-001c**: Given 개행·탭 포함 본문 카드, When show, Then 본문 필드가
    절단 없이 평탄화돼 나온다(`todoPRCell` 경로).
  - 검증: `go test ./internal/cli/ -run '^TestTodoShow$'`(단일 이름 전방위 앵커;
    시나리오별 서브테스트는 착지 시 `^TestTodoShow$/^scenario$` 형태로 확정).
- **AC-TSP-002** (maps REQ-TSP-003 — 읽기 전용·고지·등록 부수 규약)
  - Given fixture 큐, When show 실행 전후, Then 스토어 파일 mtime+sha256
    불변(읽기 전용), stdout은 고지 유무와 무관 바이트 동일, 소스에
    `"show": true` 라인 존재(`grep -n '"show": true' internal/cli/todo.go`),
    `AddCommand` 등록 확인(`grep -n 'newTodoShowCmd()' internal/cli/todo.go`).
  - 검증: `go test ./internal/cli/ -run '^TestTodoShowReadOnly$'` + 상기 grep 2종
    (각각 적중 1 이상).
- **AC-TSP-010** (maps REQ-TSP-010 — -f 본문 verbatim, RED→GREEN)
  - **RED-now(run 단계 포착)**: Given 시임 홈 fixture, When
    `moai todo add "-f hello world"`, Then 오늘은 생성 실패 — 실패 명령·출력
    전문·exit code·트리 SHA를 §E.2에 고정. RED 사유: `-f`에 등록된 숏핸드가
    없어(`todo.go:617-618`) pflag가 본문을 플래그로 소비.
  - **GREEN**: 같은 명령이 Then 카드를 생성하고 stdout `<id> <pos>`, 저장된
    text가 `"-f hello world"`와 정확히 일치.
  - 검증: `go test ./internal/cli/ -run '^TestTodoAddLeadingDashText$'`.
- **AC-TSP-011** (maps REQ-TSP-011 — 플래그 회귀)
  - Given fixture 큐, When `add --pick "x"`, `add --force "dup"`(분석기 중복
    판정 fixture), `add --classification-file <json> "y"`, Then 각각 today와
    동일한 결과(stdout 형식·잠금 원자성). `--` 구분자 형태와 t69 낙하·t203 가드
    회귀 포함.
  - 검증: `go test ./internal/cli/ -run '^(TestTodoAddPick|TestTodoAddForce|TestTodoAddClassificationFile|TestTodoBareFallthrough)$'`
    (anchored 열거형 — 비공허).
- **AC-TSP-020** (maps REQ-TSP-020 — 기본 한도 100 확정)
  - Given live 60행 fixture, When bare `moai todo`(기본 한도 100), Then 60행
    전부 렌더되고 withheld stderr가 없다. Given 120행(한도 100 초과), When
    기본, Then 100행만 렌더되고 withheld 한 줄이 잔여 20수를 밝힌다.
  - 검증: `go test ./internal/cli/ -run '^TestTodoListDefaultLimit$'`.
- **AC-TSP-021** (maps REQ-TSP-021 — 한도 계약 보존)
  - `--limit 0` 무제한, `--json` 한도 무시, `--limit -1` 거절
    (`todo.go:753-755` 문구 유지), `--dropped` 보기 — today와 동일 판정.
  - 검증: `go test ./internal/cli/ -run '^(TestTodoListLimitZero|TestTodoListJSONIgnoresLimit|TestTodoListNegativeLimit|TestTodoListDroppedOnly)$'`.
- **AC-TSP-030** (maps REQ-TSP-030 — 렌더 완전성 불변식)
  - Given queued·picked·dropped·hold가 섞인 fixture 스토어, When 기본 렌더와
    `--dropped` 렌더 각각, Then 렌더 id 집합 == 필터 만족 행 집합(집합 비교 —
    순서 무관). hold 행이 기본 보기에 포함됨을 명시 단언.
  - 검증: `go test ./internal/kanban/ ./internal/cli/ -run '^TestTodoRenderCompleteness$'`.
- **AC-TSP-031** (maps REQ-TSP-031 — 진단 기록)
  - progress.md §E.2에 t1338 관측 진단 섹션이 존재한다 — 재현 시도 결과(본
    plan 시점: 미재현), v1/v2 CHECK 제약에서 hold 쓰기 행동 확인 결과, 결론.
  - 검증: `grep -c 't1338' .moai/specs/SPEC-TODO-SURFACE-POLISH-001/progress.md`
    ≥ 1 (run 단계 판정 시점).
- **AC-TSP-040** (maps REQ-TSP-040 — 검출기 확장)
  - Given 유령 fixture(backlog.json 홈형+로컬형, .migrated, 세션 JSON 2개,
    그리고 정상 큐), When `InspectStaleLocalStores`, Then 팩트에 각 클래스별
    경로·바이트 크기가 별개 사실로 채워지고, SQLite 발산 판정과 서로 오염하지
    않는다. 유령 파일 sha256 불변(읽기 전용).
  - 검증: `go test ./internal/kanban/ -run '^TestInspectStaleLocalStoresGhostClasses$'`.
- **AC-TSP-041** (maps REQ-TSP-041 — 첫 발견 1회 안내)
  - Given 유령 존재 fixture, When 읽기 동사 1회, Then stderr 1줄 안내; When
    2회째, Then 침묵. 마커 파일이 살아 있는 상태 디렉터리에 생기고 유령
    디렉터리에는 어떤 파일도 추가되지 않는다. stdout 바이트 동일.
  - 검증: `go test ./internal/cli/ -run '^TestGhostNoticeOnce$'`.
- **AC-TSP-042** (maps REQ-TSP-042 — doctor 유령 목록 점검 + 쌍 + 골든)
  - 점검 3상태(부재 OK/존재 WARN/판독 불가 FAIL) 판정 + **같은 커밋에** 상수
    등록·allowlist·골든. 명령(anchored 열거형, 비공허):
    - `go test ./internal/cli/ -run '^(TestBinaryLag_AllowlistKeysAreLiveNames|TestBinaryLag_DoctorCheckNameSetIsUnchanged)$'`
    - `go test ./internal/cli/ -run '^(TestDoctorGolden_Light|TestDoctorGolden_Dark|TestDoctorGolden_NoColor)$'`
    - `go test ./internal/cli/ -run '^TestDoctorTodoGhostInventory$'`(신설 판정
      테스트 — 실제 이름은 착지 시 확정, 패턴 열거는 같은 커밋 갱신).
- **AC-TSP-043** (maps REQ-TSP-043 — 문서 정정 + 결정 기록)
  - 템플릿 원본에 홈 DB 경로 서술이 있고 구 레이아웃 "산 큐" 서술이 없다:
    `grep -c 'backlog.db' internal/template/templates/.moai/docs/todo-queue-storage.md`
    ≥ 1 이면서 홈 경로(`~/.moai/db/`) 언급 존재. 정정본이 브랜치에 추적됨
    (`git ls-files internal/template/templates/.moai/docs/todo-queue-storage.md`
    1행). 결정 기록은 plan §B + progress.
- **AC-TSP-050** (maps REQ-TSP-050 — 라벨 일회 이관)
  - Given `lead` 3행 + `worker-67` 2행 + 정식 `lane-1` 1행 fixture DB, When
    마이그레이션 1회, Then 레거시 0행(lead→정식 leader 표기 3행,
    worker-67→`lane-67` 2행 — §D 확정 목표 표기), `todo_identities` 등 타 표는
    바이트 불변, 전후 행수가 progress에 기록.
  - 검증: `go test ./internal/kanban/ -run '^TestOwnerLabelMigration$'`.
- **AC-TSP-051** (maps REQ-TSP-051 — doctor 라벨 드리프트 점검)
  - 레거시 존재 fixture → WARN(건수 보고), 0행 fixture → OK. 상수 등록 +
    allowlist + 골든 한 커밋 — binary_lag 명령은 AC-TSP-042의 열거형 재사용,
    판정 테스트 `^TestDoctorOwnerLabelDrift$`(실제 이름 착지 시 확정).
- **AC-TSP-052** (maps REQ-TSP-052 — 쓰기 정규화)
  - Given 마이그레이션 착지 후, When owner_label을 새로 기록하는 경로 실행,
    Then 기록값이 정식 어휘(레거시 표기 0행 유지).
  - 검증: `go test ./internal/kanban/ -run '^TestOwnerLabelWritesCanonical$'`.

## §D.1 심각도·추적

- 기능 정합(①②③): AC-TSP-001~011, 020~021 — FAIL 시 run 반복.
- 구조 불변식(⑥): AC-TSP-030 — FAIL 시 렌더 결함으로 간주, M1 재개.
- 고지 계약: AC-TSP-041 stdout 바이트 동일성 위반은 기계 표면 오염 — 즉시 수리.
- 규율 묶음: AC-TSP-042/051의 binary_lag·골든 누락은 착지 결함(t1251 선례).
- 간접 검증: grep·go test 명령은 전부 anchored 열거형/전방위 anchored로
  비공허 — 0 적중은 적색이어야 한다(t1307 plan-audit D3/D4).

## §D.2 품질 게이트

- `go vet ./internal/cli/... ./internal/kanban/...` 0 오류.
- 변경 패키지 커버리지 85% 이상(quality.yaml 기준).
- `GOOS=windows GOARCH=amd64 go build ./...` 통과(경로 처리 신규 코드 있음).
- spec-lint: 본 SPEC 디렉터리 finding 0(MissingExclusions 포함).

## §D.3 Definition of Done

1. §D AC 전수 PASS(run §E.1 매트릭스에 명령+출력 전사).
2. 진단 기록(AC-TSP-031)·이관 기록(AC-TSP-050)이 progress §E.2에 존재.
3. §D.2 게이트 전수 통과.
4. sync 단일 커밋이 `implemented → completed` 전이를 운반(3-phase close).

## Sync 노트

acceptance.md를 나중에 제자리 개정할 때는 **AC 스냅숏을 같은 커밋에** 싣는다
(카드 t1122/t1148/t1154 교훈 — 개정과 기준선 분리 금지).
