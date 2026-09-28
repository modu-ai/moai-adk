# acceptance.md — SPEC-TODO-STALE-STORE-001 (카드 t1307)

## §A 검증 원칙

모든 AC는 명령 또는 파일 검사로 독립 재현 가능해야 한다. 테스트는 `t.TempDir()` + 시임
주입으로 실제 홈 DB·primary 체크아웃 `.moai/`를 격리한다. 전체 스위트 판정은 CI 몫.

## §D AC Matrix

| AC | 마일스톤 | 요구사항 | 검증 형태 |
|---|---|---|---|
| AC-TSS-001 | M1 | REQ-TSS-001 | 테스트 + 재현 명령 |
| AC-TSS-002 | M1 | REQ-TSS-002, REQ-TSS-005 | 테스트(stdout 바이트 비교) |
| AC-TSS-003 | M1 | REQ-TSS-003 | 테스트(해시/mtime 불변) |
| AC-TSS-004 | M1 | REQ-TSS-004 | 코드 검사(단일 검출기 호출 그래프) |
| AC-TSS-010 | M2 | REQ-TSS-010 | 테스트(3상태 table-driven) |
| AC-TSS-011 | M2 | REQ-TSS-011 | `go test -run '^TestBinaryLag$'` |
| AC-TSS-012 | M2 | REQ-TSS-012 | 골든 스냅샷 파일 존재 + 골든 테스트 녹색 |
| AC-TSS-013 | M2 | REQ-TSS-013 | 테스트(검증 중 파일 불변) |
| AC-TSS-020 | M3 | REQ-TSS-020 | progress.md §E.2 측정 기록 존재 |
| AC-TSS-021 | M3 | REQ-TSS-021 | 게이트 전제조건 테스트(확인 없으면 무삭제) |
| AC-TSS-022 | M3 | REQ-TSS-022 | progress.md §E.2 증거 기록 존재 |

## §D.1 수용 시나리오 (Given-When-Then)

- **AC-TSS-001** 스테일 스토어 고지 발화 — **Given** 홈 DB `meta.last_seq`=1305와 레거시
  스토어 `meta.last_seq`=661이 공존하는 임시 프로젝트, **When** `moai todo list`를
  실행하면, **Then** stderr에 레거시 스토어 경로와 양쪽 seq 값을 밝히는 한 줄이 나오고
  종료 코드는 0이다.
- **AC-TSS-002** stdout 무변경 + 무발산 무고지 — **Given** 같은 임시 프로젝트,
  **When** `moai todo list --json`을 (a) 레거시 스토어 있음/발산 (b) 레거시 스토어 없음
  두 상태에서 실행하면, **Then** (a)(b)의 stdout이 바이트 동일하고, (b)에서 stderr에
  스테일 고지가 없으며, last_seq가 같은 상태에서도 고지가 없다.
- **AC-TSS-003** 읽기-경로 순수성 — **Given** 레거시 스토어와 홈 DB가 공존하는 임시
  프로젝트, **When** 읽기 동사(bare/list, why, pr, history)를 모두 실행하면, **Then**
  두 DB 파일의 sha256과 mtime이 실행 전과 동일하고 새 파일이 생기지 않는다(마커 부재
  포함).
- **AC-TSS-004** 단일 검출기 — **Given** M1 착지 트리, **When** 고지 경로와 doctor
  점검의 호출 그래프를 검사하면, **Then** 둘 다 동일한 kanban 팩 조회 함수를 호출하고
  둘째 프로브(자체 SQL/파일 판독)가 없다.
- **AC-TSS-010** doctor 3상태 — **Given** 임시 트리 3개(발산/홈 DB 부재/레거시 부재),
  **When** doctor 점검 함수를 실행하면, **Then** 각각 실패(양쪽 seq 포함 메시지) /
  실패 / PASS로 판정된다.
- **AC-TSS-011** binary_lag 쌍 — **Given** M2 착지 커밋, **When**
  `go test ./internal/cli/ -run '^TestBinaryLag$'`을 실행하면, **Then** 전부 통과하고
  `namesAddedAfterBaseline`에 새 점검 상수의 bare 키가 존재한다.
- **AC-TSS-012** 골든 동기화 — **Given** M2 착지 커밋, **When**
  `go test ./internal/cli/ -run '^TestDoctorGolden$'`을 실행하면, **Then** 통과하고 재생성된
  골든 파일이 같은 커밋에 포함돼 있다(커밋에 testdata/*.golden 델타 존재).
- **AC-TSS-013** doctor 읽기 전용 — **Given** 발산 상태의 임시 트리, **When** doctor
  점검을 실행하면, **Then** 어떤 DB 파일의 sha256/mtime도 변하지 않고 lock 파일이
  남지 않는다.
- **AC-TSS-020** 측정 먼저 — **Given** M3 시작, **When** 처분 절차를 진행하면,
  **Then** progress.md §E.2에 대상 4개(0바이트 쌍 2 + 백업 쌍 2) 각각의 측정 기록
  (크기·참조 판정·보존 필요성 결론)이 삭제 판정보다 먼저 기록돼 있다.
- **AC-TSS-021** 확인 게이트 전제조건 — **Given** 운영자/리드의 명시적 확인 기록이
  없는 상태, **When** 처분 절차를 실행하면, **Then** 대상 파일이 하나도
  삭제되지 않았음이 파일 존재 검사로 증명되고 측정 기록만 남는다. 삭제 실행의
  **전제조건은 확인 기록 자체**다(사후 비고 아님).
- **AC-TSS-022** 처분 증거 — **Given** 확인 기록이 존재하고 처분이 실행된 상태,
  **When** progress.md §E.2를 읽으면, **Then** 삭제된 경로, 삭제 전 sha256(또는 보존
  이동 목적지), 확인 주체와 시점이 기록돼 있다.

## §D.2 엣지 케이스

- 레거시 스토어가 0바이트이거나 SQLite가 아닌 파일 → "판독 불가"로 고지(발산과 구분),
  doctor는 실패가 아닌 별도 상태로 보고.
- 레거시 디렉터리가 둘 다(`todo`+`kanban`) 존재 → 각각 보고.
- 홈 DB와 레거시 스토어 last_seq가 같음 → 무고지, doctor PASS.
- `MOAI_HOME` override 환경 — 검출기가 override를 존중해야 함(기존 `StateDirForRoot`
  로직과 동일하게).

## §D.3 품질 게이트 (TRUST 5)

- Tested: 신규/변경 패키지 커버리지 85% 이상, 테스트 출력 인용.
- Readable/Unified: `golangci-lint run` + `gofmt` 클린(CI 버전 기준 — 카드 t1235/t1271
  교훈: 레인 lint는 CI 판 golangci 버전으로).
- Secured: 읽기 경로 쓰기 금지 검증(AC-TSS-003/013)이 보안 성격의 게이트.
- Trackable: 커밋에 카드 id(t1307) + SPEC id 명시, Conventional Commits.

## §D.4 Definition of Done

1. AC-TSS-001..022 전수 녹색, 각 AC의 재현 명령과 출력이 기록됨.
2. binary_lag 쌍(AC-TSS-011)이 커밋 단위로 함께 착지돼 있음.
3. 잔존 저장소 4개의 상태가 "측정됨/확인됨/처분됨(또는 보존)"으로 progress.md에 정리됨.
4. `go vet ./...` 통과, 변경 패키지 `go test` 녹색, CI 전체 판정 녹색.
