# acceptance.md — SPEC-USERASSET-DEPLOY-GUARD-001

## §A 범위와 판정 원칙

- 모든 AC는 기계 판정 가능하다: 판정 명령(Go 테스트 또는 빌드 게이트)과 기대 결과를 명시한다.
- **두 칸 채택 규율**(`verification-completeness.md` §2): 각 AC는 RED-now 칸(구현 전 관측될 실패 — 런 페이즈 M0에서 본 트리에서 관측하고 verbatim 출력을 progress.md §E.2에 기록)과 green-path 칸(어느 마일스톤이 무엇으로 뒤집는가)을 함께 가진다. 본 문서 작성 시점(2026-10-09, HEAD db0c514d3)에는 **어떤 재현 테스트도 실행하지 않았다** — RED-now 칸의 "관측 예정" 표기는 미실행을 정직하게 뜬다.
- RED는 바른 이유로 적색이어야 한다: 각 AC의 RED-now 칸은 실패 이유(어느 단정이 깨지는가)를 명시한다 — 공구 실패(TOOL_FAILURE)나 무관 회귀는 RED가 아니다.
- 원장 1:1 추적: §D 매트릭스. 원장 항목→REQ→AC→테스트→마일스톤.

## §B 품질 게이트 (TRUST 5)

- Tested: 아래 AC 전부 + 영향 패키지 회귀 스위트.
- 판정 명령(레인-로컬, 영향 패키지 한정):
  - `go test ./internal/userassets -count=1`
  - `go test ./internal/cli -run '^(TestMigrationClassifiesUnregisteredMirrorCopy|TestDoctorAgentEmissionUncomparedNotOK|TestSkillsDisableResolvesUserInstalledSkill|TestInitResumeAfterUserAssetEnsureFailure|TestUpdateCancelKeepsProjectAssetsIntact|TestDoctorUserInstallLoadFailureReportedSeparately|TestDoctorLockCheckVacuousNotReportedMatch|TestDoctorWorkflowRootSelectedByRequiredFile|TestDoctorFallbackKeepsProjectScopeL1)$' -count=1`
  - `go test ./internal/template -count=1`
  - `go test ./internal/web -count=1`
  - `go test ./internal/hook -count=1`
  - `GOOS=windows go build ./internal/userassets ./internal/cli` (windows 패리티 게이트)
- Unified: gofmt/go vet 청결.
- 커버리지: REQ-NFR-001 — 영향 패키지 85% 이상, 저널·잠금·변환 critical 경로 90% 이상 (`go test -cover`).

## §C 인수 기준 (Given-When-Then)

> 각 AC: [RED-now — M0에서 관측, HEAD db0c514d3] → [green path — 지정 마일스톤].

### AC-001 — 잔존 .guard 마커의 소유자 사망 후 회복 (원장 1/10-P1 · REQ-LOCK-001/002 · M2)
- **Given** 마커(`path+".acquire-guard"` 또는 windows `path+".guard"`)를 보유한 채 죽은 소유자가 잔존 마커를 남긴 상태
- **When** `go test ./internal/userassets -run '^TestGuardMarkerReclaimAfterOwnerDeath$' -count=1`
- **Then** 새 획득자가 소유자 사망을 확인하고 마커를 회수해 획득이 성공한다 (exit 0)
- RED-now: 현재 마커는 무소유권·무회수라 획득이 타임아웃 실패로 RED (관측 예정 — M0)
- green path: M2 — PID 기록+회수 구현 후 GREEN

### AC-002 — 첫 스테이징 전 회수 항목 병합 (원장 3 · REQ-JRN-001 · M1)
- **Given** 회수 대상 항목을 가진 기존 저널 + 새 설치 런
- **When** `go test ./internal/userassets -run '^TestJournalStageCarriesRecoveredEntries$' -count=1`
- **Then** 첫 `WriteJournal` 시점의 스테이지가 회수 항목을 이미 포함한다 (중단 주입 어느 지점에서도 항목 손실 없음 — REQ-JRN-001 동시 판정)
- RED-now: 현행은 반입이 :283-299(적용 뒤)라 첫 스테이징(239) 시점 파일에 회수 항목이 없어 RED (관측 예정)
- green path: M1

### AC-003 — refresh 시 반입 항목 해시 갱신 (원장 6a · REQ-JRN-002 · M1)
- **Given** 회수 항목의 대상 파일이 런 안에서 shipped 바이트로 refresh된 상태
- **When** `go test ./internal/userassets -run '^TestJournalRefreshUpdatesCarriedHash$' -count=1`
- **Then** 반입된 저널 항목의 `ExpectedSHA256`가 refresh된 값과 일치한다
- RED-now: 현행 반입(:297)은 구 해시를 그대로 적어 RED (관측 예정)
- green path: M1

### AC-004 — 파일별 즉시 WriteCompleted 영속화 (원장 10 · REQ-JRN-003 · M1)
- **Given** 다중 파일 설치 런 + 루프 중단 주입(첫 파일 성공 후)
- **When** `go test ./internal/userassets -run '^TestWriteCompletedPersistedPerFile$' -count=1`
- **Then** 중단 시점의 저널 파일에 첫 파일의 `WriteCompleted=true`가 이미 영속되어 있다 (재개 런이 기설치 파일을 collision으로 오 판정하지 않는다)
- RED-now: 현행은 루프 후 일괄 영속화(300-304)라 중단 시점에 플래그 부재로 RED (관측 예정)
- green path: M1

### AC-005 — 미인식 스키마 저널 거절 (원장 10a-재정식 · REQ-JRN-004 · M1)
- **Given** `schema_version`이 현재 값(1)이 아닌 저널 파일
- **When** `go test ./internal/userassets -run '^TestJournalUnknownSchemaRefused$' -count=1`
- **Then** `LoadJournal`이 진단과 함께 오류를 돌려준다 (조용한 디코드 금지)
- RED-now: 현행 LoadJournal은 버전 게이트 없이 디코드해 RED (관측 예정)
- green path: M1

### AC-006 — Codex 역할 TOML 참조 변환 (원장 7a · REQ-CNV-001 · M3)
- **Given** `.claude/rules` 참조를 포함하는 역할 TOML 소스 + Codex 루트 설치
- **When** `go test ./internal/userassets -run '^TestCodexAgentTOMLReferencesConverted$' -count=1`
- **Then** 설치된 TOML의 참조가 대상 하니스 배포 면으로 변환되어 있다 (원문 복사 아님)
- RED-now: 현행 fileTarget은 원문 복사라 RED (관측 예정 — 메커니즘 관측됨, 내용 주장은 재현에서 확정)
- green path: M3

### AC-007 — Codex 사용자 스킬의 Claude 전용 참조 제거 (원장 13a · REQ-CNV-001 · M3, 재현 조건부)
- **Given** CLAUDE.md·`.claude/rules/moai/` 참조를 포함하는 스킬 디렉터리 소스
- **When** `go test ./internal/userassets -run '^TestCodexUserSkillInstallNoClaudeOnlyRefs$' -count=1`
- **Then** Codex 루트에 설치된 스킬 사본에 미배포 참조가 없다 (또는 REQ-CNV-002의 파일 단위 보고가 나온다)
- RED-now/적용 여부: **미검증 릴레이** — M0 재현이 확정할 때까지 구현 착수 금지. 재현이 반박하면 본 AC는 범위 조정(plan-audit 회신) 대상
- green path: M3 (재현 확정 시)

### AC-008 — Codex 역할 TOML 무변환 좌표 확정 (원장 13b · REQ-CNV-001 · M3, 재현 조건부)
- **Given** 원장 13b의 주장 좌표(install.go:584 표류 — 실제 복사 경로는 :392)
- **When** M0 재현 재확정 후 `go test ./internal/userassets -run '^TestCodexRoleTOMLNotVerbatim$' -count=1`
- **Then** :392 경로 설치물이 변환된다 — AC-006과 동일 결함으로 확정되면 본 AC는 AC-006에 병합·소관 종료로 표기
- 상태: relayed-unverified — 좌표 표류 명시(카드 본문). 재현이 동일 결함으로 확정하는 것이 본 AC의 1차 목적
- green path: M3 (병합 확정 시 AC-006의 GREEN으로 갈음)

### AC-009 — 동결 가드의 사용자 경로 삭제 거절 (원장 9a · REQ-GRD-001/002 · M4)
- **Given** 보호 집합에 포함된 사용자 설치 루트 아래의 관리 파일(예: plan-auditor.md)
- **When** `go test ./internal/hook -run '^TestFrozenGuardDeniesUserPathDelete$' -count=1`
- **Then** 삭제 시도가 거절된다 (allow 통과 금지)
- RED-now: 사용자 루트가 보호 집합(config 적재)에 없다는 주장의 재현 — RED (관측 예정; 집합 불포함 자체는 relayed)
- green path: M4

### AC-010 — 충돌 사전 판정의 FIFO 무차단 (원장 8c · REQ-COL-001 · M5)
- **Given** 설치 대상 경로에 FIFO가 존재
- **When** `go test ./internal/userassets -run '^TestCollisionPrecheckSkipsFifoWithoutBlock$' -count=1 -timeout 30s`
- **Then** 판정이 유한 시간 내에 비차단 분류로 끝난다 (타임아웃 없음)
- RED-now: 현행 `os.ReadFile`(211)은 FIFO에서 무한 블록해 타임아웃 RED (관측 예정)
- green path: M5

### AC-011 — rename 직전 부모 재검증 (원장 12 · REQ-COL-002 · M5)
- **Given** confined 쓰기의 검증-이후-rename 사이에 부모 symlink 교체 주입
- **When** `go test ./internal/userassets -run '^TestConfinedWriteRevalidatesParentBeforeRename$' -count=1`
- **Then** rename이 거부되고 외부 경로에 기록이 발생하지 않는다
- RED-now: 현행 재검증(725-728)은 CreateTemp(729) 전이라 swap 창이 열려 RED (관측 예정)
- green path: M5

### AC-012 — 미등록 미러 사본의 분류·보고 (원장 4 · REQ-SRF-001 · M6)
- **Given** 매니페스트 항목 없는 `.agents/skills` 미러 사본 + 확인된 사용자 측 대응물
- **When** `go test ./internal/cli -run '^TestMigrationClassifiesUnregisteredMirrorCopy$' -count=1`
- **Then** 사본이 분류·보고되고 사용자 대응물 확인 시 제거된다 (영구 보존 아님)
- RED-now: 현행 미등록 항목은 untouched 보존(:119-122)이라 RED (관측 예정)
- green path: M6

### AC-013 — doctor emission 미비교 집계 수정 (원장 5-잔여 · REQ-SRF-002 · M6)
- **Given** 구 `.codex/agents` 배출물 비교에서 비교 불가 항목 존재
- **When** `go test ./internal/cli -run '^TestDoctorAgentEmissionUncomparedNotOK$' -count=1`
- **Then** compared 0/12 상태가 OK로 보고되지 않는다 (누락 보고)
- RED-now: 현행 집계가 미비교를 허용해 RED (관측 예정 — doctor_agentemit_embed.go:135-160)
- green path: M6

### AC-014 — skills disable의 사용자 계층 폴백 (원장 8d+5-증상 · REQ-SRF-003 · M6)
- **Given** 프로젝트 미러가 없고 사용자 설치 스킬만 존재
- **When** `go test ./internal/cli -run '^TestSkillsDisableResolvesUserInstalledSkill$' -count=1`
- **Then** 사용자 계층 면이 해석되어 비활성화가 가능하다 (absent/Nothing-to-disable 무시 아님)
- RED-now: 생 면(skills.go)의 미러 전용 해석으로 RED (관측 예정 — 좌표 퇴역 재표현분)
- green path: M6

### AC-015 — 동명 에이전트 행의 분리·양 행 적용 (원장 8a · REQ-SRF-004 · M6)
- **Given** 같은 이름의 사용자 에이전트와 프로젝트 에이전트
- **When** `go test ./internal/web -run '^TestAgentFormDuplicateNameRowsBothApplied$' -count=1`
- **Then** 폼 행이 스코프 구분되어 렌더되고 POST가 두 행의 편집을 모두 적용한다
- RED-now: 홈 우선 스캔(:123-132)+이름 키 PostFormValue(:282-283)로 둘째 편집 소실 RED (관측 예정)
- green path: M6

### AC-016 — depends_on 클로저 설치 (원장 8b · REQ-SRF-005 · M6)
- **Given** `depends_on`을 선언한 선택 번들
- **When** `go test ./internal/userassets -run '^TestBundleDependsOnClosureInstall$' -count=1`
- **Then** 의존 번들의 자산이 함께 설치되고 보존 집합이 의존 번들을 포함한다 (prune 절반은 M0 변별 결과를 따른다)
- RED-now: 수집 경로(collectEntries)에 클로저 확장 부재로 RED (관측 예정)
- green path: M6

### AC-017 — update 취소 시 프로젝트 자산 불변 (원장 11a · REQ-SRF-006 회귀 가드 · M6)
- **Given** 확인창에서 취소되는 update 런
- **When** `go test ./internal/cli -run '^TestUpdateCancelKeepsProjectAssetsIntact$' -count=1`
- **Then** 프로젝트 관리 자산이 바이트 그대로 유지된다
- 상태: HEAD에서 이미 수리 관측 — 본 AC는 **처음부터 GREEN이어야 하는 회귀 가드**다. RED가 관측되면 수리 이행(relocation)의 퇴행이므로 별도 결함으로 리더 보고
- green path: M6 (가드 테스트 상륙)

### AC-018 — init 실패 후 재개 경로 (원장 11b · REQ-SRF-007 · M6)
- **Given** 사용자 자산 ensure 실패로 끝난 직후의 프로젝트
- **When** `go test ./internal/cli -run '^TestInitResumeAfterUserAssetEnsureFailure$' -count=1`
- **Then** 재실행이 "already initialized" 거절 대신 부족분 완성 재개를 제공한다
- RED-now: :944-948 실패 후 재실행이 :905에서 거절되는 흐름으로 RED (관측 예정)
- green path: M6

### AC-019 — doctor의 적재 실패 별도 보고+읽기 전용 (원장 2a · REQ-DOC-001 · M7)
- **Given** 파손된 사용자 매니페스트
- **When** `go test ./internal/cli -run '^TestDoctorUserInstallLoadFailureReportedSeparately$' -count=1`
- **Then** 실패 등급이 별도 행으로 보고되고 복구 사본이 덮어쓰이지 않는다 (OK 위장 없음)
- RED-now: 임의 적재 오류가 CheckOK로 위장(:134-137)해 RED (관측 예정)
- green path: M7

### AC-020 — 공허 잠금 비교의 정직 보고 (원장 2b · REQ-DOC-002 · M7)
- **Given** 잠금 비교 가능한 면이 없는 프로젝트 트리(예: .claude/settings.json만 존재)
- **When** `go test ./internal/cli -run '^TestDoctorLockCheckVacuousNotReportedMatch$' -count=1`
- **Then** "project tree matches the lock file" 대신 공허 조건이 보고된다
- RED-now: 루트 게이트(:165) 통과 파일이 0개여도 OK(:184)로 RED (관측 예정)
- green path: M7

### AC-021 — 워크플로 루트의 필수 파일 기반 선택 (원장 9b · REQ-DOC-003 · M7)
- **Given** 빈 프로젝트 workflows 디렉터리 + 정상 사용자 워크플로
- **When** `go test ./internal/cli -run '^TestDoctorWorkflowRootSelectedByRequiredFile$' -count=1`
- **Then** 필수 파일 존재 기준으로 사용자 워크플로가 선택되고 L4가 정상 판정된다
- RED-now: 디렉터리 존재만으로 프로젝트 선택(:57)해 사용자 워크플로 읽기 실패 RED (관측 예정)
- green path: M7

### AC-022 — 폴백 시 L1 프로젝트 유지 (원장 7b · REQ-DOC-004 · M7)
- **Given** 폴백이 발동한 상태(프로젝트 workflows 부재·사용자 워크플로 존재)
- **When** `go test ./internal/cli -run '^TestDoctorFallbackKeepsProjectScopeL1$' -count=1`
- **Then** L1은 프로젝트 skillsDir를 검사하고 L6 참조 해석이 skillsDir 교체로 오 판정되지 않는다 (L6:FAIL·L1:PASS 오류 재현 소멸)
- RED-now: 통째 교체(:58)로 RED (관측 예정)
- green path: M7

### AC-023 — 커버리지 게이트 (REQ-NFR-001 · 전 마일스톤)
- **When** §B의 스코프 명령에 `-cover`를 결합해 측정한다 — 예: `go test -cover ./internal/userassets ./internal/template ./internal/web -count=1` (internal/cli는 위 스코프 명령에 `-cover` 결합)
- **Then** 영향 패키지 85% 이상, 저널·잠금·변환 critical 경로 90% 이상

### AC-024 — windows 패리티 빌드 게이트 (REQ-LOCK-002 · M2)
- **When** `GOOS=windows go build ./internal/userassets ./internal/cli`
- **Then** 빌드 성공 (exit 0) — windows 전용 잠금 가족 수정이 크로스 컴파일로 판정 가능

### AC-025 — 설치 스크립트 실행 권한 보존 (원장 6b · REQ-COL-003 · M5)
- **Given** 설치 대상에 실행 가능한 스크립트 자산(navigator-audit.sh 재현 형태)이 포함
- **When** `go test ./internal/userassets -run '^TestConfinedWritePreservesExecBit$' -count=1`
- **Then** 설치된 .sh의 파일 모드가 0755로 기록된다 (실행 권한 보존 — 0644 하락 없음)
- RED-now: 현행 confinedWrite의 0o644 하드코딩(install.go:743, 관측됨)으로 RED (관측 예정 — M0)
- green path: M5

## §D 추적성 매트릭스 (원장 → REQ → AC → 테스트 → 마일스톤)

| 원장 | REQ | AC | 테스트 | 마일스톤 |
|---|---|---|---|---|
| 1, 10-P1 | REQ-LOCK-001, 002 | AC-001, AC-024 | TestGuardMarkerReclaimAfterOwnerDeath | M2 |
| 2 | REQ-DOC-001, 002 | AC-019, AC-020 | TestDoctorUserInstallLoadFailureReportedSeparately, TestDoctorLockCheckVacuousNotReportedMatch | M7 |
| 3 | REQ-JRN-001 | AC-002 | TestJournalStageCarriesRecoveredEntries | M1 |
| 4 | REQ-SRF-001 | AC-012 | TestMigrationClassifiesUnregisteredMirrorCopy | M6 |
| 5-잔여 | REQ-SRF-002, 003 | AC-013, AC-014 | TestDoctorAgentEmissionUncomparedNotOK, TestSkillsDisableResolvesUserInstalledSkill | M6 |
| 6a | REQ-JRN-002 | AC-003 | TestJournalRefreshUpdatesCarriedHash | M1 |
| 6b | REQ-COL-003 | AC-025 | TestConfinedWritePreservesExecBit | M5 |
| (커버리지 게이트) | REQ-NFR-001 | AC-023 | go test -cover (§B 결합) | 전 마일스톤 |
| 7a | REQ-CNV-001, 002 | AC-006 | TestCodexAgentTOMLReferencesConverted | M3 |
| 7b | REQ-DOC-004 | AC-022 | TestDoctorFallbackKeepsProjectScopeL1 | M7 |
| 8a | REQ-SRF-004 | AC-015 | TestAgentFormDuplicateNameRowsBothApplied | M6 |
| 8b | REQ-SRF-005 | AC-016 | TestBundleDependsOnClosureInstall | M6 |
| 8c | REQ-COL-001 | AC-010 | TestCollisionPrecheckSkipsFifoWithoutBlock | M5 |
| 8d | REQ-SRF-003 | AC-014 | TestSkillsDisableResolvesUserInstalledSkill | M6 |
| 9a | REQ-GRD-001, 002 | AC-009 | TestFrozenGuardDeniesUserPathDelete | M4 |
| 9b | REQ-DOC-003 | AC-021 | TestDoctorWorkflowRootSelectedByRequiredFile | M7 |
| 10 | REQ-JRN-003 | AC-004 | TestWriteCompletedPersistedPerFile | M1 |
| 10a | REQ-JRN-004 (재정식) | AC-005 | TestJournalUnknownSchemaRefused | M1 |
| 11a | REQ-SRF-006 (가드) | AC-017 | TestUpdateCancelKeepsProjectAssetsIntact | M6 |
| 11b | REQ-SRF-007 | AC-018 | TestInitResumeAfterUserAssetEnsureFailure | M6 |
| 12 | REQ-COL-002 | AC-011 | TestConfinedWriteRevalidatesParentBeforeRename | M5 |
| 13a | REQ-CNV-001, 002 | AC-007 | TestCodexUserSkillInstallNoClaudeOnlyRefs | M3 (조건부) |
| 13b | REQ-CNV-001 | AC-008 | TestCodexRoleTOMLNotVerbatim | M3 (조건부·병합 후보) |
| (windows 패리티 게이트) | REQ-LOCK-002 | AC-024 | GOOS=windows go build | M2 |

## §E 간접 검증과 폐쇄 게이트

- 간접 검증: AC-017(회귀 가드)은 결함이 아닌 불변의 고정이며, RED 관측 시 퇴행 결함으로 리더 보고한다.
- 폐쇄 게이트: §C 전 AC GREEN + §B 게이트 통과 + 열린 질문(spec.md §7)의 plan-audit 회신 기록 + windows 빌드 게이트.

## §F Definition of Done

1. M0 RED 배터리 관측 기록이 progress.md §E.2에 존재한다 (verbatim).
2. §C 전 AC가 지정 마일스톤에서 GREEN으로 전환됐다 (예외: AC-007/008의 재현 반박 경로는 plan-audit 회신에 따라 조정·기록).
3. §B 게이트 전부 통과, AC-023/024 포함.
4. 상태 전이: draft → in-progress (M1 첫 커밋, manager-develop 소관) → implemented → completed (sync 커밋, manager-docs 소관).
