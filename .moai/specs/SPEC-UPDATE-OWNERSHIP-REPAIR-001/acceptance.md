---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "acceptance criteria"
version: "0.1.0"
created: 2026-10-18
---

# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — 인수 기준

## 예산과 규칙

- Tier L 상한: 요구사항 25, 인수 기준 25. 이 파일은 **19** 개의 기준 (AC-OWR-01 ~ -19) 을 가진다. 요구사항은 15 개이며 각각 하나 이상의 기준을 가진다 (spec.md § E).
- 판정 명령은 모두 **단일 호출 형태**이다 (파이프, 리다이렉션, `&&` 없음). 시험 명령의 `TMPDIR` 는 `/private/tmp` 아래 카드 토큰이 없는 이름 (예: `/private/tmp/own-fixture-d1`) 이다. 모든 시험은 임시 HOME 과 `t.TempDir()` 만 쓴다. 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체는 쓰지 않는다.
- **RED 셀 규약.** 판정 시험은 plan 단계에 아직 트리에 없다. 따라서 각 기준의 RED-now 셀 — 명령, 원문 stdout, 종료 코드, 고정한 트리 SHA, 실패 이유 — 는 plan.md 의 RED 작업이 구현 전에 기록한다 (progress.md §E.2). 그 기록이 있기 전의 기준은 회귀 가드이며 릴리스 게이트가 아니다. 아래 "RED 이유" 는 그 셀에 적힐 것으로 예상하는 이유이다.
- **제로 매치 셀렉터는 통과도 RED 도 아니다.** `-run` 이 아무 시험도 고르지 못해 `[no tests to run]` 과 `ok` 를 내면 해당 기준의 실패이다. 판정은 `--- PASS: <시험 이름>` 줄이 보이고 쓸어낸 시험 수가 기대와 같을 때만 성립한다.
- **변이 시험.** 보존 기준마다 "이 변이가 통과하면 기준이 얕다" 는 변이를 적었다. RED 단계에서 변이를 구현이 아닌 시험 쪽 입력 변형으로 한 번 돌려 시험이 붉어지는지 확인한다.
- 공통 시험 헬퍼: `buildMigrationFixture`, `installMigrationUserCounterparts`, `profileReachSandbox`, `enableRemovalForTest`, `runUpdateFlowWithProgress`, `removalBackupCopy` (모두 t1668 이 이 트리에 가져온 것), 그리고 plan.md M0 이 만드는 `migrate_ownership_helpers_test.go`.

## 인수 기준

### AC-OWR-01 — 판정 함수의 네 분류 *(maps REQ-OWR-001)* (항목 (1))

- **Given** 같은 스킬 루트 아래에: 매니페스트에 `template_managed` 로 등록되고 바이트가 기록된 `TemplateHash` 와 같은 파일 (m), 같은 등록이지만 바이트가 다른 파일 (u), `user_created` 로 등록된 파일 (c), 매니페스트 항목이 없는 파일 (x).
- **When** 판정 함수 `classifyOwnership` 가 네 파일을 판정한다.
- **Then** 분류는 m=`managed`, u=`user-modified`, c=`user-created`, x=`unregistered` 이다. 분류 상수의 집합은 정확히 이 네 값이며 시험이 그 수를 단언한다. 사용자 대응물이 확인되고 제거 스위치가 픽스처에서 켜진 경우에만 m 이 `removal-eligible` 이고 나머지는 `keep(<이유>)` 이다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipVerdictClasses$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipVerdictClasses (…)`, 종료 코드 0.
- RED 이유: 판정 함수가 아직 없어 빌드 실패 (`undefined: classifyOwnership`, 종료 코드 1).
- 변이: x 를 `user-modified` 로 분류하는 구현 — 분류 단언이 붉어진다.
- 게이트: TDD RED 마일스톤 (M1a) 의 릴리스 게이트.

### AC-OWR-02 — 보존 픽스처 1: 템플릿과 바이트가 같은 사용자 작성 파일 *(maps REQ-OWR-002, REQ-OWR-003)* (항목 (6))

- **Given** 매니페스트 항목이 없고 바이트가 내장 템플릿의 같은 경로 파일과 동일한 파일 두 개 — `.claude/skills/moai-user-authored/SKILL.md` 와 `.agents/skills/moai-user-authored/SKILL.md` — 그리고 확인된 사용자 대응물, 픽스처에서 켠 제거 스위치.
- **When** 마이그레이션이 실행된다.
- **Then** 두 파일은 바이트 그대로 남는다. 분류는 `unregistered`, 이유는 `no-provenance-record` 이고 잔존 행의 빠진 증거 항목이 `deployed-version`, `original-deployed-path`, `prior-manifest-record` 셋 모두이다. 제거 목록과 실행 백업의 `removed-assets/` 에 두 파일이 없다. 매니페스트에 두 파일의 새 항목이 생기지 않는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveByteEqualUserFile$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipPreserveByteEqualUserFile (…)`, 종료 코드 0.
- RED 이유: 이유 코드와 빠진 증거 항목이 아직 기록되지 않아 단언 실패 (파일 보존 자체는 현재도 성립한다 — `TestMigrationClassifiesUnregisteredMirrorCopy`).
- 변이: 바이트 일치와 확인된 대응물만으로 `managed` 로 보고 제거하는 구현 — 파일이 사라져 붉어진다.
- 게이트: TDD RED 마일스톤 (M1a) 의 릴리스 게이트.

### AC-OWR-03 — 보존 픽스처 2: 사용자가 수정한 파일 *(maps REQ-OWR-002)* (항목 (1))

- **Given** 매니페스트에 `template_managed` 로 등록된 뒤 사용자가 바이트를 바꾼 스킬 파일 (`TemplateHash` 와 다름), 그리고 `user_modified` 로 등록된 에이전트 파일, 제거 스위치 켬.
- **When** 마이그레이션이 실행된다.
- **Then** 두 파일은 바이트 그대로 남고, 분류는 `user-modified`, 이유는 `user-modified` 이며, 잔존 행에 오른다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipPreserveUserModifiedFile$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipPreserveUserModifiedFile (…)`, 종료 코드 0.
- RED 이유: 분류와 이유 코드가 잔존 행으로 기록되지 않아 단언 실패 (보존 자체는 `TestRF2F3_MigrationPreservesPostInstallEdits` 가 이미 지킨다).
- 변이: 현재 바이트 대신 기록된 `Provenance` 만 보고 `managed` 로 취급 — 수정된 파일이 제거되어 붉어진다.
- 게이트: TDD RED 마일스톤 (M1a).

### AC-OWR-04 — 보존 픽스처 3: 출처 불명 미등록 파일 *(maps REQ-OWR-003)* (항목 (6))

- **Given** 매니페스트 항목이 없고 어떤 템플릿 파일과도 바이트가 다른 파일들 — `.agents/skills/moai-custom-x/SKILL.md`, `.claude/agents/moai/custom-agent.md` — 그리고 템플릿 이름과 같은 경로이지만 바이트가 다른 `.agents/skills/moai-plan/extra.md`.
- **When** 마이그레이션이 `--verbose` 와 함께 실행된다.
- **Then** 모두 보존된다. 각 파일의 잔존 행은 분류 `unregistered`, 이유 `no-provenance-record`, 빠진 증거 항목을 가진다. `--verbose` 출력이 각 파일 경로와 이유를 이름으로 댄다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipUnregisteredUnknownOriginKeptWithReason$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipUnregisteredUnknownOriginKeptWithReason (…)`, 종료 코드 0.
- RED 이유: 현재 출력은 경로만 대고 이유 코드와 빠진 증거 항목이 없다.
- 게이트: TDD RED 마일스톤 (M1a).

### AC-OWR-05 — 양성 픽스처 4: 증거가 완비된 파일만 managed *(maps REQ-OWR-002, REQ-OWR-015)* (항목 (6))

- **Given** 이전 매니페스트 세대가 경로 P 를 `template_managed` 로, 배포 버전 V, 템플릿 해시 H 와 함께 기록하고, 현재 매니페스트에는 P 의 항목이 없으며 (항목 유실), 디스크의 P 는 해시 H 의 바이트를 가진다.
- **When** 판정 함수가 P 를 판정한다.
- **Then** 분류는 `managed` (증명된) 이고 증거 기록에 배포 버전 V, 원래 배포 경로 P, 이전 기록의 세대가 들어 있다. 기본 처분은 `keep(removal-gate-off)` 이고 `removal-eligible` 은 제거 스위치를 픽스처에서 켜고 사용자 대응물이 확인된 때만이다. **매니페스트는 바뀌지 않는다** (자동 채택 없음). 변이 네 가지가 모두 `managed` 가 아니다: (a) 기록에서 배포 버전을 뺀 경우 → `unregistered` + `evidence-incomplete:deployed-version`, (b) 같은 해시의 기록이 다른 경로 P' 키로만 있는 경우 (재배치된 파일) → `evidence-incomplete:original-deployed-path`, (c) 이전 세대에 P 도, 같은 해시의 다른 경로 기록도 없는 경우 → `no-provenance-record`, (d) 디스크 바이트가 H 와 다른 경우 → `record-hash-mismatch`.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestOwnershipAttestedManagedOnlyWithFullEvidence$' -count=1 -v`
- 기대: `--- PASS: TestOwnershipAttestedManagedOnlyWithFullEvidence (…)` (서브 시험 5개: 양성 1 + 변이 4), 종료 코드 0.
- RED 이유: 증거 해석기가 없어 빌드 실패.
- 비고: decision-index.md Q2 가 (나) 로 풀리면 이 시험은 증거를 시험이 직접 공급하는 판정 함수 수준으로 남고, 운영 경로의 이전 세대 읽기는 빠진다.
- 게이트: TDD RED 마일스톤 (M1b).

### AC-OWR-06 — 배포 버전 기록과 이전 매니페스트 보존 *(maps REQ-OWR-015)* (항목 (6))

- **Given** (a) `internal/manifest` 에서 `Track` 으로 항목을 기록하는 매니페스트 매니저, 그리고 `deployed_version` 필드가 없는 기존 JSON. (b) 갱신 흐름을 한 번 도는 픽스처.
- **When** (a) `Track` 을 부르고 저장·재적재한다. (b) 갱신 흐름이 Backup 단계와 배포 단계를 거친다.
- **Then** (a) 새 항목은 배포 바이너리의 버전을 `deployed_version` 에 가진다. 필드가 없는 기존 JSON 은 그대로 적재되고, 다시 저장해도 필드를 지어내지 않는다. (b) 실행 백업 디렉터리에 `manifest.prior.json` 이 있고 그 바이트는 실행 시작 시점의 매니페스트와 같다. 매니페스트 수준 `Version` 은 배포 버전으로 갱신되고 `DeployedAt` 은 `Version` 이 바뀔 때만 바뀐다.
- 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/manifest/ -run '^TestManifestRecordsDeployedVersion$' -count=1 -v`
- 판정 명령 2: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestPriorManifestRetainedInRunBackup$' -count=1 -v`
- 기대: 두 명령 모두 해당 `--- PASS`, 종료 코드 0.
- RED 이유: 필드와 보존 단계가 없어 단언 실패 또는 빌드 실패.
- 비고: Q2 가 (나) 로 풀리면 이 기준은 삭제된다.
- 게이트: TDD RED 마일스톤 (M1b).

### AC-OWR-07 — 프로필 지도 프로젝트: 대응 파일 없는 사본 *(maps REQ-OWR-004)* (항목 (2))

- **Given** 런치 원장이 프로젝트를 프로필 `work` 에 지도한 샌드박스 (`profileReachSandbox` + `mapProjectToWork`), `template_managed` 로 등록되고 바이트가 `TemplateHash` 와 같은 에이전트 파일 둘과 스킬 파일 하나, 사용자 루트 대응물은 없음 (임시 HOME 이 비어 있음), 제거 스위치를 픽스처에서 켬.
- **When** 마이그레이션이 실행된다.
- **Then** 세 파일은 제자리에 바이트 그대로 남고 제거 목록이 비어 있다. 세 항목의 `KeptBy` 는 `profile-reach` 이다 (Q1 의 결정 그대로). 잔존 행의 이유는 `profile-reach`, 사용자 루트 대응물 상태는 `absent`, 프로필 쪽 대응물 상태는 `not-checked` 이다. 어느 파일도 `removal-eligible` 이 아니다. 변형 1: 사용자 대응물이 확인되면 상태가 `confirmed` 이고 여전히 보존된다 (기존 `TestMigrationKeepsSharedCopyUnderProfileReach` 는 그대로 GREEN). 변형 2: 대응물이 있으나 해시가 다르면 `unconfirmed`.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestProfileMappedCopyKeptWithoutCounterpart$' -count=1 -v`
- 기대: `--- PASS: TestProfileMappedCopyKeptWithoutCounterpart (…)`, 종료 코드 0.
- RED 이유: 대응물 상태와 `not-checked` 가 이유에 기록되지 않아 단언 실패.
- 변이: 프로필 매핑을 무시하는 구현 — 대응물 확인 변형에서 사본이 제거되어 붉어진다.
- 게이트: TDD RED 마일스톤 (M2).

### AC-OWR-08 — 프로필 환경과 프로필 루트 무접근 *(maps REQ-OWR-005)* (항목 (2))

- **Given** `CLAUDE_CONFIG_DIR` 을 픽스처 디렉터리 `profileDir` 로 지정한 환경. `profileDir` 아래에 `skills/x/SKILL.md` 와 `agents/y.md` 센티널 파일이 있고 디렉터리 모드는 0o000 이다. AC-OWR-07 과 같은 프로젝트.
- **When** 마이그레이션과 갱신 흐름이 실행된다.
- **Then** 결과가 AC-OWR-07 과 같고 오류가 없다 (읽기 시도가 있었다면 권한 오류가 났을 것이다). 실행 전후 `profileDir` 의 이름·바이트·수정 시각 요약이 같다. 별도로 정적 시험이 `migrate_ownership*.go` 와 `migrate_project_assets.go` 의 소스에 `claude-skills@`, `claude-agents@` 문자열이 없고 프로필 기본 디렉터리를 `ReadDir`/`Glob` 하는 호출이 없음을 단언한다. 파일 불변은 `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go` 가 종료 코드 0 인 것으로 확인한다.
- 판정 명령 1: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestProfileEnvDoesNotReachProfileRoot|TestNoProfileRootAccessInOwnershipSources)$' -count=1 -v`
- 판정 명령 2: `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go`
- 기대: 명령 1 은 두 시험의 `--- PASS`, 명령 2 는 종료 코드 0 (출력 없음).
- RED 이유: 두 시험이 아직 트리에 없어 지금 실행하면 제로 매치 (`[no tests to run]`) 이고 이는 RED 가 아니다. RED 셀은 시험을 먼저 만든 뒤, 프로필 루트를 읽는 변이 입력 (아래 변이) 으로 두 시험이 붉어지는 것을 기록한다.
- 변이: 판정 코드에 `os.ReadDir(os.Getenv("CLAUDE_CONFIG_DIR"))` 를 끼운 구현 — 0o000 디렉터리 읽기 오류와 정적 스캔 양쪽에서 붉어진다.
- 게이트: TDD RED 마일스톤 (M2).

### AC-OWR-09 — commands: 한 실행에 한 처분 *(maps REQ-OWR-006)* (항목 (3))

- **Given** 템플릿 commands 를 네 상태로 나눈 픽스처 — 건강한 매니페스트 항목 (managed), 사용자 수정, 템플릿과 바이트가 같지만 항목 없음 (`migAbsentRecCmd` 모양), 사용자 작성 `.claude/commands/moai-user-cmd.md`.
- **When** 전체 갱신 흐름 (`runUpdateFlowWithProgress`) 이 실행된다.
- **Then** reconcile 계획의 네 분류 목록 (template-owned, user-modified, user-owned, stale) 이 서로소이다 — 한 경로가 둘 이상의 처분에 나타나지 않는다. managed 는 현재 템플릿 바이트로 갱신되고, 사용자 수정은 운영자의 바이트가 보존되거나 `.moai-new` 곁파일과 함께 병합되며, 항목 없는 바이트 동일 파일은 제자리에 남고, 사용자 작성 파일은 건드려지지 않는다. 실행 뒤 템플릿이 운반하는 모든 commands 파일이 존재한다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCommandsSingleDispositionAndPresent$' -count=1 -v`
- 기대: `--- PASS: TestCommandsSingleDispositionAndPresent (…)`, 종료 코드 0.
- RED 이유: 현장의 "이전 제외 후 재배포" 가 이 트리에서 재현되는지가 이 시험의 첫 관측이다 (research.md G-2). 재현되면 서로소 단언이나 존재 단언이 붉어지고, 재현되지 않으면 이 기준은 회귀 가드로 남아 RED 셀에 "재현 안 됨" 을 적고 리더에게 보고한다.
- 게이트: M6 의 측정 결과에 따라 릴리스 게이트 또는 회귀 가드.

### AC-OWR-10 — 중단 뒤에도 commands/agents 탐색 가능 *(maps REQ-OWR-006)* (항목 (3))

- **Given** 앞선 실행이 `.claude/commands/moai` 와 `.claude/agents/moai` 를 비운 채 죽은 상태 (픽스처가 만든다). harness 별로 claude, both 두 경우.
- **When** 갱신이 다시 실행된다.
- **Then** 두 디렉터리는 템플릿이 운반하는 집합과 이름·바이트가 같고, 각 파일의 frontmatter 에 비어 있지 않은 `name`/`description` 이 있다. 실행 백업에서 사라진 사용자 파일은 없다. both 에서는 `.codex/agents/moai/*.toml` 도 템플릿 집합과 같다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCommandsAndAgentsDiscoverableAfterInterruptedRun$' -count=1 -v`
- 기대: `--- PASS: TestCommandsAndAgentsDiscoverableAfterInterruptedRun (…)` (서브 시험 2개), 종료 코드 0.
- RED 이유: 전용 시험이 없다 (research.md D(3)). 회귀 가드로 시작해 구현 변화가 이 집합을 깨는 변이로 붉어지는 것을 RED 셀에 기록한다.
- 게이트: 회귀 가드 (탐색 회귀 없음의 증거).

### AC-OWR-11 — 대소문자 변형을 한 파일로 판정 *(maps REQ-OWR-007)* (항목 (4))

- **Given** 등록 키 `.claude/skills/moai-ref-owasp-checklist/SKILL.md` (`template_managed`, 바이트 == `TemplateHash`) 와 디스크의 같은 파일이 `skill.md` 철자로 존재하는 상태. 파일시스템 대소문자 모드를 주입하는 시험과, 실제 호스트 파일시스템에서 도는 시험 둘.
- **When** 판정 함수가 걷기 경로 `.../skill.md` 를 판정한다.
- **Then** 대소문자 무구분 모드 (또는 실제 호스트가 무구분일 때): 등록 항목의 판정이 적용되어 분류는 `managed`, 이유에 `case-variant` 가 붙고, 이 물리 파일의 행은 하나다. 대소문자 구분 모드: 두 철자는 별개 경로이다 — `skill.md` 는 `unregistered`, `SKILL.md` 는 자기 항목의 판정을 받는다. 어느 모드에서도 파일 이름은 바뀌지 않는다. 실제 호스트 시험은 프로브 결과에 맞는 기대를 단언하며 어느 쪽도 건너뛰지 않는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestCaseVariantResolvesToRegisteredEntry|TestCaseVariantOnHostFilesystem)$' -count=1 -v`
- 기대: 두 시험의 `--- PASS` (첫 시험은 서브 시험 2개: 무구분/구분), 종료 코드 0.
- RED 이유: 현재 `skill.md` 는 매니페스트 맵 조회를 놓쳐 `unregistered` 로 센다 (research.md D(4); 이 호스트는 무구분이다).
- 변이: 철자를 무시하고 항상 별칭으로 합치는 구현 — 구분 모드에서 `skill.md` 가 `managed` 가 되어 붉어진다.
- 게이트: TDD RED 마일스톤 (M4).

### AC-OWR-12 — 대소문자 변형: 사용자 매니페스트 키와 `.agents/skills` 철자 *(maps REQ-OWR-007)* (항목 (4))

- **Given** 사용자 매니페스트가 `claude-skills/<name>/SKILL.md` 를 기록하고 프로젝트 파일은 `skill.md` 철자인 대소문자 무구분 모드, 그리고 `.agents/skills/<name>/skill.md` 의 미등록 복사본.
- **When** 마이그레이션이 대응물을 확인하고 미등록 복사본을 판정한다.
- **Then** 대응물 상태가 `confirmed` 이다 (철자 때문에 `unconfirmed` 가 되지 않는다). 미등록 복사본은 하나의 행으로 `no-provenance-record` + `case-variant` 이유를 받고, 이름 바꾸기 없이 바이트 그대로 남는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestCaseVariantCounterpartAndMirrorSpelling$' -count=1 -v`
- 기대: `--- PASS: TestCaseVariantCounterpartAndMirrorSpelling (…)`, 종료 코드 0.
- RED 이유: 사용자 매니페스트 키 조회와 내장 트리 조회가 철자가 다르면 빗나간다 (`migrate_project_assets.go:400`, `:436`).
- 게이트: TDD RED 마일스톤 (M4).

### AC-OWR-13 — 관리 파일만 있던 빈 디렉터리 정리 *(maps REQ-OWR-008)* (항목 (5))

- **Given** 이전 매니페스트 세대가 기록한 파일이 전부 `template_managed` 이고 지금은 전부 없는 빈 디렉터리 셋 — (A) 이전 실행이 비운 스킬 디렉터리, (B) 이번 실행이 마지막 파일을 제거해 비운 스킬 디렉터리 (제거 스위치는 픽스처에서 켬), (C) 하위 폴더만 남아 있고 그 하위 폴더도 비어 있는 중첩 디렉터리. 그리고 반례: (D) 이전 기록이 없는 빈 디렉터리, (E) 이전 기록에 `user_created` 파일이 있었던 빈 디렉터리, (F) 사용자 파일이 들어 있는 디렉터리, 스킬 루트, 비어 있는 `.claude/agents/moai`.
- **When** 마이그레이션이 실행된다.
- **Then** A, B, C 는 사라진다 (C 는 아래쪽부터). D 와 E 는 남고 잔존 행의 이유가 `empty-dir-no-provenance` 이다. F 는 내용이 그대로이다. 스킬 루트와 에이전트 디렉터리는 남는다. 정리 단계는 파일을 지우지 않는다 (`os.Remove` 는 빈 디렉터리만 지운다).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestManagedOnlyEmptyDirsPruned$' -count=1 -v`
- 기대: `--- PASS: TestManagedOnlyEmptyDirsPruned (…)` (서브 시험: A, B, C, D, E, F), 종료 코드 0.
- RED 이유: 걷기가 이미 비어 있던 디렉터리를 보지 않고 (`migrate_project_assets.go:195-197`) 정리 판단의 근거도 없다.
- 변이: 비어 있기만 하면 지우는 구현 — D 가 사라져 붉어진다.
- 게이트: TDD RED 마일스톤 (M5).

### AC-OWR-14 — 잔존 목록과 313개 보존 *(maps REQ-OWR-003, REQ-OWR-009)* (항목 (6), (7))

- **Given** 복사 폴백 거울을 흉내 낸 `.agents/skills` 아래 미등록 파일 313개 (템플릿과 바이트가 같다), 확인된 사용자 대응물, 제거 스위치를 픽스처에서 켬. 별도로 사용자 수정 파일 몇 개와 프로필 지도 사본 몇 개가 섞인 두 번째 픽스처.
- **When** 갱신 흐름이 실행되고, 이어서 `--verbose` 로 한 번 더 실행된다.
- **Then** 313개 모두 바이트 그대로 남는다. 매니페스트는 바이트 단위로 바뀌지 않는다 (배포가 기록하는 항목 외에 새 항목이 없다). 잔존 행 313개가 이유 `no-provenance-record` 와 빠진 증거 항목을 가지며 경로 사전순으로 정렬된다. 기본 렌더는 이유별 개수 줄을 내고 개별 경로를 찍지 않으며, `--verbose` 렌더는 파일별 행을 낸다. 두 번째 픽스처의 행은 이유가 섞여 있고 각 행에 이유가 있다 (이유 없는 행이 없다).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemainingListKeeps313UnregisteredCopies$' -count=1 -v`
- 기대: `--- PASS: TestRemainingListKeeps313UnregisteredCopies (…)`, 종료 코드 0.
- RED 이유: 구조화된 잔존 행과 이유별 렌더가 없다 (현재는 보고 문자열과 사유 없는 경로 집합뿐 — `update_migrate.go:127`).
- 변이: 이유 열을 비운 행을 내는 구현 — "이유 없는 행 없음" 단언이 붉어진다. 313개를 기본 출력에 모두 찍는 구현 — 기본 렌더 단언이 붉어진다.
- 게이트: TDD RED 마일스톤 (M3).

### AC-OWR-15 — 백업·복원 바이트 동일 *(maps REQ-OWR-010)* (항목 (7))

- **Given** 제거 스위치를 픽스처에서 켠 전체 갱신 흐름과, 실행 전 트리 해시 (런 백업 디렉터리 제외).
- **When** 흐름이 파일을 제거한 뒤, 백업된 사본을 원위치에 되돌린다 (이 되돌리기는 시험 전용 절차이며 운영 하위 명령이 아니다).
- **Then** 제거된 각 파일의 백업 사본 sha256 이 제거 전 sha256 과 같다. 실행 백업에 `manifest.prior.json` 이 있고 되돌린 매니페스트 바이트가 실행 전 매니페스트와 같다. 되돌린 트리의 해시가 실행 전 트리 해시와 같다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRemovedAssetBackupRestoreByteIdentical$' -count=1 -v`
- 기대: `--- PASS: TestRemovedAssetBackupRestoreByteIdentical (…)`, 종료 코드 0.
- RED 이유: 매니페스트 사본이 실행 백업에 없어 트리 해시 단언이 붉어진다 (파일 사본의 바이트 일치는 `TestMigrationBackupBeforeRemoval` 이 이미 지킨다).
- 비고: Q2 가 (나) 로 풀리면 파일 바이트 일치만 단언한다. 파일 모드는 보존 범위 밖이다 (spec.md § J).
- 게이트: TDD RED 마일스톤 (M8).

### AC-OWR-16 — dry-run 무변경과 판정 일치 *(maps REQ-OWR-001, REQ-OWR-011)* (항목 (7))

- **Given** AC-OWR-02 ~ -05, -07 의 파일들이 섞인 프로젝트, 제거 스위치를 픽스처에서 켬 (그리고 별도로 끔).
- **When** `moai update --dry-run` 에 해당하는 미리보기 (`previewReconciliation`) 가 실행된다.
- **Then** 실행 전후 트리 해시가 같고 (모든 경로, 바이트, 모드) `.moai-backups` 가 생기지 않는다. 미리보기가 "migration candidate" 로 알린 경로 집합이 판정 함수의 `removal-eligible` 집합과 정확히 같다. 프로필 지도 + `KeptBy` 항목은 후보에 없다. 스위치가 꺼져 있으면 후보가 없다 (기존 `TestPreviewOmitsRemovalCandidateWhileSwitchOff` 는 GREEN 유지).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestDryRunChangesNothingAndMatchesVerdict$' -count=1 -v`
- 기대: `--- PASS: TestDryRunChangesNothingAndMatchesVerdict (…)`, 종료 코드 0.
- RED 이유: 미리보기 판정 (`update_dryrun_preview.go:120`) 은 `KeptBy` 와 대응물 확인을 보지 않아 프로필 지도 항목을 후보로 알린다 (코드 읽기 도출, research.md G-3 — 이 시험이 첫 관측이다).
- 변이: 미리보기가 자기 판정을 유지하는 구현 — 후보 집합 불일치로 붉어진다.
- 게이트: TDD RED 마일스톤 (M7).

### AC-OWR-17 — 반복 갱신 멱등 *(maps REQ-OWR-012)* (항목 (7))

- **Given** AC-OWR-14 의 첫 픽스처와 섞인 픽스처를 갱신 흐름으로 세 번 실행한다 (같은 버전, 변경 없음).
- **When** 실행 1, 2, 3 이 끝난다.
- **Then** 실행 뒤마다 계산한 트리 해시 (제외: `.moai-backups/**`, `.moai/logs/**`, `.moai/state/**`) 가 셋 모두 같고, 잔존 행 목록이 셋 모두 같다. 매니페스트 바이트가 같다 (`DeployedAt` 은 `Version` 이 바뀔 때만 바뀐다 — design.md D-3).
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestRepeatedUpdateIsIdempotent$' -count=1 -v`
- 기대: `--- PASS: TestRepeatedUpdateIsIdempotent (…)`, 종료 코드 0.
- RED 이유: 잔존 행이 없어 단언 불가, 그리고 배포 단계가 매 실행 `DeployedAt` 을 갱신하는 구현(M1b)이 들어오면 붉어진다 — 이 시험이 그 변이를 잡는다.
- 변이: 매 실행 `DeployedAt` 을 갱신하는 구현 — 매니페스트 바이트 불일치로 붉어진다.
- 게이트: TDD RED 마일스톤 (M9).

### AC-OWR-18 — 중도 중단 → 복구 → 재실행 수렴 *(maps REQ-OWR-013)* (항목 (7))

- **Given** 제거 스위치를 픽스처에서 켠 프로젝트와 같은 프로젝트의 중단 없는 기준 실행 결과. 중단 모사 세 가지 (운영 코드에 끼어들기를 넣지 않는다): (a) 둘째 파일의 실행 백업 경로를 쓸 수 없게 만들어 일부만 진행, (b) 마이그레이션 뒤 매니페스트를 실행 전 바이트로 되돌려 저장 전 중단을 모사, (c) 정리된 빈 디렉터리를 다시 만들어 정리 전 중단을 모사.
- **When** 중단 상태에서 복구 (쓸 수 없던 경로를 풀거나 백업 사본을 되돌림) 한 뒤 갱신을 다시 실행한다.
- **Then** 최종 트리 해시와 잔존 행이 중단 없는 기준 실행과 같다. 제거된 각 파일의 백업 사본이 존재한다. 제거된 경로의 매니페스트 항목이 되살아나지 않고, 비어 있는 관리 디렉터리가 남지 않는다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^TestInterruptedMigrationRecoversAndConverges$' -count=1 -v`
- 기대: `--- PASS: TestInterruptedMigrationRecoversAndConverges (…)` (서브 시험 a, b, c), 종료 코드 0.
- RED 이유: (c) 에서 이미 비어 있던 디렉터리가 정리되지 않는다 (AC-OWR-13 의 같은 원인). (a), (b) 는 현재 구조가 이미 수렴하는지의 첫 관측이다.
- 게이트: TDD RED 마일스톤 (M9).

### AC-OWR-19 — 제거 스위치는 꺼진 채 *(maps REQ-OWR-014)* (항목 (1), (7) 게이트)

- **Given** 이 트리의 비시험 Go 소스.
- **When** 소스 스캔 시험이 `removalEnabled` 대입을 찾고, 기존 게이트 시험이 기본 구성에서 돈다.
- **Then** 비시험 코드에서 `removalEnabled` 를 `true` 로 대입하는 곳이 없다 (선언의 `false` 만). `TestMigrationRemovalGate` 와 `TestMigrationDefaultKeepsConfirmedProjectCopy` 가 GREEN 이다.
- 판정 명령: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestRemovalSwitchStaysOffInProduction|TestMigrationRemovalGate|TestMigrationDefaultKeepsConfirmedProjectCopy)$' -count=1 -v`
- 기대: 세 시험의 `--- PASS`, 종료 코드 0.
- RED 이유: 소스 스캔 시험이 아직 없다. 실패쪽 관측은 스위치 기본값을 `true` 로 뒤집은 변이에서 두 기존 시험과 새 시험이 붉어지는 것으로 기록한다.
- 게이트: 회귀 가드 (실패쪽은 변이로 관측).

## 가장자리 사례

- 손상된 매니페스트: 읽기 전용 로더가 nil 을 돌려주면 모든 파일이 `unregistered` 이다 — 전부 보존되고 이유가 기록된다 (제거 없음).
- 사용자 매니페스트 없음 또는 HOME 해석 불가: 사용자 대응물 상태가 `unconfirmed` 이고 모든 사본이 보존된다.
- 루트 밖으로 나가는 심링크: 기존대로 제거를 거부하며 이유가 `symlink-escape` 로 기록된다.
- 읽기 실패 또는 `TemplateHash` 가 빈 항목: 보존, 이유 `read-failed` 또는 `no-template-hash`.
- Windows 경로 구분자: 판정 함수는 슬래시 경로를 키로 쓴다 (`isCommonAssetCleanTarget` 의 정규화와 같은 방향).

## 품질 게이트 (Definition of Done)

- AC-OWR-01 ~ -19 의 RED 셀(명령, 원문 stdout, 종료 코드, 트리 SHA, 이유)과 GREEN 원문이 progress.md §E.2 에 기록되어 있다. 제로 매치로 통과를 주장한 기준이 없다.
- 기준선 11개 시험과 t1668 시험 (`TestMigrationBackupBeforeRemoval`, `TestEmptiedSkillDirRemoved`, `TestMigrationRemovesSharedCopyWithoutProfileReach`, `TestMigrationRemovalGate`) 이 GREEN 이다.
- 터치한 패키지 `go test -cover` 85% 이상. `golangci-lint`, `go vet`, `gofmt` 클린.
- REQ-UPM-031 불변식: 제거가 일어난 실행의 요약이 삭제를 보고한다 (기존 `TestTallyCountsMigrationRemovals` 계승).
- 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체가 없었다 (시험은 임시 HOME 과 `/private/tmp` 아래 `TMPDIR` 만 썼다).
- 갭 목록이 최신이고 빈 Gaps 선언은 사실일 때만이다.

## 갭 목록

- **G-A (AC-OWR-09, -10)**: 현장의 commands "이전 제외 후 재배포" 가 이 트리에서 재현되는지 미측정. M6 의 첫 관측이 소거하거나 AC 를 회귀 가드로 낮춘다.
- **G-B (AC-OWR-11, -12)**: 대소문자 결론은 맵 조회 의미와 호스트 파일시스템 관측의 조합이다. 마이그레이션을 그 철자로 실행한 관측은 RED 셀이 처음 만든다.
- **G-C (AC-OWR-14)**: `.agents/skills` 313개가 심링크 거울 행인지 복사 폴백 파일인지 미확인. 시험은 두 모양을 모두 다룬다 (복사 폴백 313 + 심링크 거울 행).
- **G-D (AC-OWR-05, -06, -15)**: Q2 미결. (나) 가 선택되면 운영 경로의 이전 세대 읽기가 사라진다.
- **G-E (REQ-OWR-005)**: 프로필 쪽 대응물은 읽지 않는다 (Q5). 프로필 쪽 상태는 항상 `not-checked` 이다.
- **G-F (spec.md § J)**: 백업 사본의 파일 모드 정규화로 실행 가능 비트가 복원에서 빠질 수 있다. 바이트 동일만 검증한다.
