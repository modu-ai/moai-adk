---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "research — moai update ownership verdict and legacy-migration repair"
version: "0.3.0"
created: 2026-10-18
---

# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — 코드 조사 기록 (research.md)

> 이 파일은 plan 시점 HEAD `cdd72a814` 의 조사 기록이다. 버전 번호는 SPEC 묶음 (spec.md `0.3.0`) 과 맞췄을 뿐 본문 수치(요구사항·인수 기준 개수 등)는 갱신하지 않았다. 현재 개수와 이후 측정은 spec.md, acceptance.md § L 이 가진다.

## A. 목적과 방법

카드 t1865 가 전달한 현장 진단 일곱 항목과 "해시 일치만으로는 관리 소유권을 확정하지 않는다"는 조건 (6) 에 대해, 각 증상을 만드는 코드의 위치와 현재 동작, 그리고 그 위치를 덮는 기존 시험을 읽기 전용으로 확인한 기록이다. 현장 진단 자체는 사용자가 전달한 것이며 ADK 에서 재측정한 값이 아니다. 이 문서의 모든 코드 좌표는 아래 기준 트리의 것이고, 다른 트리에서는 다시 측정해야 한다.

- 기준 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, 브랜치 `WT-update-ownership-repair`, HEAD `cdd72a814` (t1668 브랜치 `WT-t1619-spec-b` 의 `e1e163320` 병합을 포함).
- 판정에 쓴 도구 빌드: 트리에서 직접 빌드한 `/private/tmp/own-fixture-bin/moai` (설치본이 아니다). 명령은 `go build -o /private/tmp/own-fixture-bin/moai ./cmd/moai`. 이 바이너리의 버전 문자열은 ldflags 없이 빌드해 기본값 `moai-adk v3.1.3` 을 출력하므로, 빌드 출처의 증거는 버전 문자열이 아니라 "HEAD `cdd72a814` 에서 빌드했다"는 이 기록이다.
- 시험 실행의 임시 디렉터리: `TMPDIR=/private/tmp/own-fixture-d1` (HOME 과 저장소 밖, 카드 토큰을 담지 않은 이름).
- 모든 조사는 읽기 전용이다. 이 조사에서 코드, 시험, 설정은 바뀌지 않았다. 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체는 쓰지 않았다.

## B. 기준선 — 기존 시험 11개가 모두 통과한다

다음 명령으로 현재 동작을 확인했다 (쓸어낸 시험 11개, 빈 쓸기 아님).

```
cd <tree> && TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestMigrationClassifiesUnregisteredMirrorCopy|TestMigrationKeepsSharedCopyUnderProfileReach|TestMigrationRemovesSharedCopyWithoutProfileReach|TestEmptiedSkillDirRemoved|TestMigrationBackupBeforeRemoval|TestMigrationDefaultKeepsConfirmedProjectCopy|TestMigrationRemovalGate|TestMigrationIdempotent|TestMigrationPreservesExistingMirrorEntries|TestRF2F3b_MigrationPreservesUntracked|TestPreviewOmitsRemovalCandidateWhileSwitchOff)$' -count=1 -v
```

관측된 출력 (원문 발췌, 종료 코드 0):

```
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (2.38s)
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (2.22s)
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (1.80s)
--- PASS: TestMigrationBackupBeforeRemoval (4.06s)
--- PASS: TestEmptiedSkillDirRemoved (4.43s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (1.68s)
--- PASS: TestMigrationRemovalGate (1.45s)
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.01s)
--- PASS: TestMigrationIdempotent (2.84s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (3.06s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	25.915s
```

t1668 이 가져온 제거 기계 (백업 후 제거, 빈 스킬 디렉터리 정리, 제거 게이트) 는 이 트리에서 이미 동작한다. 이 SPEC 은 그 기계를 다시 만들지 않는다.

## C. 가장 중요한 구조 발견 — 소유권 판정이 네 곳에 흩어져 있다

같은 질문 ("이 파일은 템플릿이 배포한 관리 파일인가, 사용자 파일인가") 을 서로 다른 근거와 어휘로 답하는 코드가 네 곳이다.

| # | 위치 (HEAD `cdd72a814`) | 대상 루트 | 근거 | 결과 어휘 | 운영 경로에서 쓰이는가 |
|---|--------------------------|-----------|------|-----------|------------------------|
| 1 | `internal/cli/migrate_project_assets.go:128-360` `migrateProjectCommonAssets` | `.claude/skills/`, `.agents/skills/`, `.claude/agents/moai/`, `.codex/agents/moai/` (`:31-36`). `.claude/commands/` 는 없다 | 프로젝트 매니페스트 항목 (`Provenance`, `TemplateHash`) + 현재 바이트 + 사용자 대응물 (`userCounterpartConfirmed` `:377-452`) | 보존/유지/제거 (보고 문자열) | 예 — Backup 단계 안에서 호출 (`internal/cli/update_template_sync.go:817`) |
| 2 | `internal/cli/update/migrate_classify.go:172` `ClassifyMigration` | `.claude/skills`, `.claude/commands`, `.agents/skills` (`:80-84`) | 템플릿 운반 여부 (template carriage) + 매니페스트 건강도 | identical/modified/foreign | 아니오 — 비시험 호출자를 grep 으로 찾지 못했다. 트리거 `runUpdateMigrationTrigger` 는 판정 없이 `migrateOptedOut` 만 돌려준다 (`internal/cli/update_migrate.go:141`) |
| 3 | `internal/cli/update/reconcile_classify.go:282` `classifyManagedFile` | 관리 청소 대상 + 공통 자산 루트를 PreserveOnly 로 (`internal/cli/update_reconcile.go:52-89`) | 템플릿 운반 + 매니페스트 `CurrentHash` 일치 | template-owned/user-modified/user-owned/stale | 예 — 기본 경로 `reconcileManagedRoots` (`update_template_sync.go:609`) |
| 4 | `internal/cli/update_dryrun_preview.go:120` `migrationRemovalCandidate` | 공통 자산 루트 | `Provenance==TemplateManaged` + `TemplateHash` + 바이트 일치만 | 후보/비후보 | 예 — `--dry-run`. 단 `removalEnabled` 가 켜진 때만 후보를 알린다 (`:84`) |

네 판정은 서로 다른 입력을 본다. 구체적인 어긋남 두 가지.

1. 미리보기 판정 (#4) 은 프로필 도달 유지 표지 (`KeptBy`) 도, 사용자 대응물 확인도 보지 않는다. 그래서 제거 스위치를 시험 픽스처에서 켠 상태의 미리보기는 `profile-reach` 로 유지될 항목을 "migration candidate" 로 알릴 수 있다 (코드 읽기로 도출. 이 조합을 실행해 보지는 않았다 — Gaps G-3).
2. 템플릿이 운반하지만 매니페스트 항목이 없는 파일은 #1 에서 "unregistered mirror copy" 로 유지되고, #3 에서는 `ClassUserModified` (병합/충돌 경로) 로 분류된다 (`reconcile_classify.go` 의 템플릿 운반 + 비건강 기록 분기). 둘 다 보존 쪽이라 데이터는 잃지 않지만 어휘가 달라 하나의 "잔존 목록" 으로 모을 수 없다.

## D. 항목별 조사

항목 번호는 카드 본문 범위절의 (1)-(7) 을 따른다. 현장 진단의 증상은 괄호로 붙였다.

### (1) 출처·배포 manifest/hash 근거의 파일별 소유권 판정 (현장: 실행 rc.36-t1822c/ee981a932 dirty, 템플릿 기록 rc.35)

현재 동작: #1 의 판정 순서는 다음과 같다 (`migrate_project_assets.go`).

- 매니페스트 항목 없음 → "unregistered mirror copy" 로 유지 (`:216-226`, `migrationPreservedProjectFiles` 에 기록).
- `user_created` → 건드리지 않음 (`:227-230`).
- 읽기 실패 또는 `TemplateHash` 비어 있음 → 유지 (`:236-241`).
- 현재 바이트의 해시 ≠ 기록된 `TemplateHash` → 사용자 수정으로 유지 (`:243-248`).
- `user_modified` → 유지 (`:249-254`).
- 프로필 지도 프로젝트 + `template_managed` → 유지 + `KeptBy` 표지 (`:255-272`).
- 사용자 대응물 미확인 → 유지 (`:275-290`).
- 루트 경계 심링크 검사 → 거부 (`:291-295`).
- `removalEnabled` 가 꺼짐 → 유지 (`:296-302`).
- 백업 실패 → 제거 거부 (`:307-312`), 성공 후 `os.Remove` (`:313-316`).

매니페스트가 기록하는 것: `FileEntry` 는 `Provenance`, `TemplateHash`, `DeployedHash`, `CurrentHash`, `Parts`, `KeptBy` 뿐이다 (`internal/manifest/types.go:103-113`). **파일별 배포 버전은 없다.** 매니페스트 수준의 `Version`/`DeployedAt` (`types.go:93-96`) 은 프로젝트 초기화 때만 쓰인다 (`internal/core/project/initializer.go:607-608`, 비시험 대입은 이 한 곳뿐). 갱신 배포가 기록하는 것은 `TemplateTreeHash` 이다 (`internal/cli/update_template_sync_guard.go:29-40`, 호출 `update_template_sync.go:672`). 따라서 rc.35 로 기록된 프로젝트를 rc.36 바이너리가 갱신해도 매니페스트에는 "어느 버전이 이 파일을 배포했는가" 가 남지 않는다. 이전 매니페스트 세대를 보존하는 곳도 없다: 실행 백업은 `.moai/config` 와 마이그레이션 제거 사본만 담는다 (`update_disk_backup.go:65-78`, `migrate_project_assets.go:84-104`).

증상과의 연결: 실행 바이너리가 dirty 빌드 (`ee981a932`) 이면 내장 템플릿 트리가 출시 트리와 다를 수 있다. 사용자 대응물 확인은 내장 트리와의 바이트 일치를 요구하므로 (`userCounterpartConfirmed` `:436-451`), "템플릿과 바이트가 같다" 는 사실은 출시된 템플릿과 같다는 뜻이 아니라 이 빌드의 내장 트리와 같다는 뜻일 뿐이다. 바이트 일치만으로 소유권을 정할 수 없는 구체적 이유이다.

기존 시험: `TestMigrationClassification` (`internal/cli/migrate_classify_ac_test.go:41` — 이름만 확인, 이 조사에서 본문을 읽지 않았다), `TestFR3_1_MigrationPreservesOnUnverifiable`, `TestRF2F3_MigrationPreservesPostInstallEdits`, `TestRF2F3b_MigrationPreservesUntracked`, `TestRF2F4_StaleRecordDoesNotAuthorizeDeletion`. 배포 버전/이전 매니페스트 증거를 요구하는 시험은 없다.

### (2) 프로필 도달 예외 (현장: 에이전트 6·스킬 1 이 보존되나 대응 사용자/프로필 파일 없음)

현재 동작: 런치 원장이 프로젝트를 프로필에 매핑하면 (`profile.ResolveLaunchProfileForProject(rootAbs, "") != ""`, `migrate_project_assets.go:181-185`, 정의 `internal/profile/profile.go:461`) 템플릿 일치 `template_managed` 사본마다 `SetKeptBy(..., KeptByProfileReach)` 를 호출한다 (`:255-272`, 상수 `types.go:117`). 코드 주석이 명시하듯 "사용자 대응물 확인은 이 유지를 게이트하지 않는다" (`:259-260`). 즉 대응 사용자 파일이 없어도 표지가 붙는다. 보고 문장은 "profile sessions need this shared copy" 로 같다.

위험: 표지는 시험 서문에서 "3.2-2-5 doctor 가 되찾는다 (reclaim)" 는 약속으로 쓰인다 (`migrate_profile_reach_test.go` 의 `TestMigrationKeepsSharedCopyUnderProfileReach` 주석). 대응 파일이 없는 사본에 같은 표지가 붙으면 나중의 reclaim 이 유일한 사본을 되찾을 대상으로 읽을 수 있다. 다만 `KeptBy` 를 읽는 비시험 코드는 이 트리에 아직 없다 (grep: 비시험 참조는 `manifest.go` 의 `SetKeptBy` 와 이 마이그레이션의 쓰기뿐). 소비자는 후속 카드 몫이다.

프로필 루트: 이 코드는 프로필 디렉터리를 읽지 않는다. 사용자 대응물 확인은 `userassets.RootBySlugDir(homeDir, ...)` 로 `$HOME/.claude/...` 계열 네 루트만 본다 (`internal/userassets/paths.go:36-56`). 프로필 쪽 루트 (`claude-skills@<profile>/` 등) 를 열거하거나 쓰는 코드는 이 트리의 마이그레이션에 없다. 또한 이 트리에는 `internal/config/userroots.go` 가 없다 (`git ls-files | grep -i userroots` 결과 없음). `internal/config/protected_zone.go` 는 존재한다 (428줄, `user-root:<slug>/` 항목 종류 `:70-82`, `:294-335`). 이 SPEC 은 두 파일 모두 건드리지 않는다.

기존 시험: `TestMigrationKeepsSharedCopyUnderProfileReach`, `TestMigrationRemovesSharedCopyWithoutProfileReach`, `TestMigrationReachLeavesUserModifiedCopy`, `TestFileEntryKeptBy` (`internal/cli/migrate_profile_reach_test.go`). 대응 파일이 없는 프로필 지도 사본을 다루는 시험은 없다. `CLAUDE_CONFIG_DIR` 은 시험 샌드박스가 비운다 (`profileReachSandbox` `:29-40`).

### (3) commands 이전 제외 ↔ 재배포 순환 (현장: commands 17개가 이전 대상에서 빠졌다가 재배포)

현재 동작: `.claude/commands/` 는 REQ-020 마이그레이션의 걷기 루트에 없다 (`migrate_project_assets.go:31-36`). commands 는 다음 경로로만 다룬다.

- 관리 청소 대상 목록에 `.claude/commands/moai` 가 있다 (`internal/cli/update/deploy/deploy.go:65-75`). `computeRunCleanTargets` 는 공통 자산 루트만 청소 범위에서 뺀다 (`update_migrate.go:75-97`), 그래서 commands 는 청소 범위에 남는다.
- 기본 경로의 "Clean Managed Paths" 단계는 전면 삭제가 아니라 `reconcileManagedRoots` 이다 (`update_template_sync.go:609`). 전면 삭제 (`CleanMoaiManagedPathsWithTargetsGuarded`, `:559`) 는 구 v1→v2 설정 감지 분기 (`legacyClean`) 에서만 쓰인다.
- 마이그레이션 트리거는 분류 없이 `migrateOptedOut` 을 돌려주므로 (`update_migrate.go:141`) 분류기 #2 가 commands 를 제거 후보로 만드는 경로는 현재 없다.
- 이후 "Deploy Templates" 단계가 commands 를 다시 쓴다 (`update_template_sync.go:618`, 배포 호출 `:641`).

그러므로 현장이 말한 "이전 대상에서 빠졌다가 재배포" 가 어떤 경로를 가리키는지는 저장소 사실만으로 확정되지 않는다. 후보는 (a) 보고에서 commands 가 "마이그레이션 제외" 로 열거된 뒤 같은 실행의 배포 단계가 다시 쓰는 서술상의 순환, (b) `legacyClean` 분기의 전면 삭제 뒤 재배포, (c) reconcile 이 매니페스트 항목 없는 템플릿 운반 commands 를 `ClassUserModified` 로 병합 경로에 보내는 것. 확정은 M0 의 RED 탐침 (plan.md) 이 한다. 이 불확실성은 Gaps G-2 에 기록했다.

기존 시험: `TestUpdateMigratesLegacyProject`, `TestForceUpdatePreservesUserCommonAssets`, `TestUnknownSkillPreservedAcrossLocalUpdates` 등 (`update_migrate_test.go`) 이 commands 픽스처 (`migForeignCommand`, `migAbsentRecCmd` `:37-41`) 를 쓴다. 이 둘 중 `migAbsentRecCmd` (`.claude/commands/moai/todo.md`) 는 "템플릿과 바이트가 같지만 매니페스트 항목이 없는" 모양이라 이 SPEC 의 "바이트 일치 ≠ 소유" 조건과 같은 픽스처 모양이다. "commands 와 agents 가 실행 뒤에도 모두 있는가" 를 직접 단언하는 전용 시험은 이 조사에서 찾지 못했다.

### (4) macOS 대소문자 무구분 경로 (현장: OWASP skill.md/SKILL.md 불일치)

관측 (이 호스트): 파일시스템이 대소문자 무구분이다. `skill.md` 를 만든 뒤 `SKILL.md` 로 존재 확인이 성공한다. `filepath.WalkDir` 은 디스크에 저장된 철자 (`skill.md`) 를 돌려주고, `os.ReadFile(.../SKILL.md)` 는 오류 없이 읽으며, `os.SameFile` 은 두 철자에 대해 `true` 이다 (스크래치 프로그램 `/private/tmp/own-fixture-d3/walk.go`, `probe2.go` 의 출력).

코드 도출: 매니페스트 `Files` 는 정확한 문자열 키의 Go 맵이다 (`internal/manifest/types.go:96`, `GetEntry` `internal/manifest/manifest.go:157` 의 맵 조회). 템플릿이 배포한 키는 `SKILL.md` 이고 (`internal/template/templates/.claude/skills/moai-ref-owasp-checklist/SKILL.md`), 걷기가 돌려준 `skill.md` 는 조회에서 빗나간다. 그 결과 `migrate_project_assets.go:216-226` 이 같은 물리 파일을 "unregistered mirror copy" 로 센다. 사용자 매니페스트 조회와 내장 트리 조회도 같은 이유로 철자가 다르면 빗나간다 (`:400`, `:436`). 이 도출은 맵 조회 의미에서 오며, 이 철자 조합으로 마이그레이션을 실행해 보지는 않았다 (G-4).

이미 있는 장치: `internal/cli/update/skillresidue` 가 소문자 `skill.md` 를 `SKILL.md` 로 바꾸는 이동을 가진다. 원자적 비교체 이름 바꿈 두 번 (`skill.md` → 임시 폴더 → `SKILL.md`) 이고 (`residue.go` 패키지 주석, `internal/cli/update_destructive_registry.go` 의 보호 설명), 확정 뒤에 적용하도록 보류된다 (`update_state_residue.go:146`, `:175`; 호출 `update_template_sync.go:254`, `:369`, `:1246`). 한계: 대상은 `.claude/skills` 아래 폴더뿐이고 `opts.shipped(rel+"/SKILL.md")` 인 스킬뿐이다 (`residue.go:117`, `:139`). `.agents/skills` 아래 복사본의 철자는 다루지 않는다. 대소문자 구분 파일시스템에서 두 철자가 함께 있으면 `reported-both-skill-files` 로 보고만 한다 (`residue.go:123-127`). 파일시스템 대소문자 프로브는 `update_cleanup.go:387` 의 `probeCaseSensitiveFS` 가 있다.

기존 시험: `TestClean_RenamesLowercaseSkillFileWhenCanonicalAbsent`, `TestClean_ReportsButKeepsBothSkillFilesPresent` (`skillresidue/residue_test.go`). 마이그레이션이 철자 변형을 같은 경로로 보는지 단언하는 시험은 없다.

### (5) 관리 파일만 남았던 빈 디렉터리 정리 누락 (현장: 빈 디렉터리 3개)

현재 동작: 빈 스킬 디렉터리 정리는 t1668 이 넣은 `pruneEmptiedSkillDir` (`migrate_project_assets.go:526-559`) 하나이며, 파일을 방금 제거한 직후 (`:332-336`) 에만 호출된다. 후보는 스킬 루트 바로 아래 한 단계의 디렉터리뿐이다 (`:541-550`). 걷기 콜백은 디렉터리 항목에서 아무 일도 하지 않는다 (`:195-197`). 따라서 다음은 정리되지 않는다.

- 이번 실행이 비우지 않은, 이미 비어 있던 디렉터리 (이전 실행이나 중단된 실행의 잔재).
- `removalEnabled` 가 꺼진 운영 기본값에서는 제거가 일어나지 않으므로 정리 호출 자체가 없다.
- 하위 폴더만 비어 있는 스킬 디렉터리 (한 단계 규칙).
- `.agents/skills` 의 빈 복사본 디렉터리 (이 경로는 `.claude/skills` 와 같은 한 단계 규칙을 가지지만 호출 지점이 같아 같은 조건).

`skillresidue.Clean` 이 빈 폴더를 지우는 경우는 "템플릿이 배포하는 스킬" 의 빈 폴더뿐이고, 그 외의 빈 폴더는 사용자 것으로 두며 아무것도 보고하지 않는다 (`residue.go:114-122`).

그러므로 "관리 파일만 남았던" 빈 디렉터리 (매니페스트 기록상 안에 있던 파일이 전부 `template_managed` 이고 지금은 전부 없는 디렉터리) 를 가려내는 근거는 현재 코드에 없다. 필요한 근거는 이전 매니페스트 세대 (안에 무엇이 있었는가) 이다.

기존 시험: `TestEmptiedSkillDirRemoved` (제거 직후 정리), `TestClean_RemovesEmptyFolderOnlyForMoaiSkillNames`.

### (6) 미등록 파일 (현장: `.agents/skills` 313개가 매니페스트 미등록으로 보존)

산출 경로: `.agents/skills` 의 내용은 두 갈래이다.

- 게시된 명령 스킬 17개 `.agents/skills/moai-<command>/SKILL.md` — 정규 배포 파일이며 일반 걷기로 배포되어 `Track` 된다 (`internal/template/published_skills.go:22,29`, `internal/template/deployer.go:294`).
- 정규 스킬 카탈로그의 거울 — `.agents/skills/<name>` 이 `../../.claude/skills/<name>` 의 상대 심링크이고, 심링크를 만들 수 없는 환경에서는 실제 디렉터리 복사본 (`skill_mirror.go` 머리말 `:1-21`, `mirrorOneSkill` `:230-`). 거울 항목은 **의도적으로 매니페스트에 등록하지 않는다** (같은 머리말의 제약 2, 그리고 `deployer.go:310-318`). 또 `harness_fs.go` 의 gpt/both 프로필에서는 카탈로그 스킬이 실제 디렉터리로 `.agents/skills` 에 재배치되고 거울은 꺼진다 (`harness_fs.go:13-19`, `:429-`).

관측 (탐침): `.agents/skills/<name>` 이 심링크면 `filepath.WalkDir` 은 그것을 디렉터리가 아닌 항목 (`IsDir=false`, `Type=L`) 으로 돌려준다 (`probe2.go` 출력). 그래서 심링크 거울은 스킬당 한 행의 "미등록" 으로 보이고, 복사 폴백 거울은 파일마다 한 행이다. 313 이라는 수는 이 두 모양 중 어느 쪽에서 왔는지 저장소만으로 알 수 없다 (G-5).

현재 동작: 매니페스트 항목이 없으면 "removal needs creation provenance or your approval" 문구와 함께 유지되고 (`:216-226`), 시험은 `TestMigrationClassifiesUnregisteredMirrorCopy` (`m0_doctor_red_test.go:27`) 와 `TestMigrationPreservesExistingMirrorEntries` 이다. 즉 보존은 이미 된다. 빠진 것은 (a) 보존 사유가 파일별로 구조화되지 않고 보고 문자열에만 있다는 점, (b) 어느 증거가 부족해서 관리로 못 정했는지 남지 않는다는 점, (c) 증거가 갖춰진 경우를 관리로 인정하는 규칙 자체가 없다는 점이다.

배포기가 매니페스트에 없는 기존 파일을 만났을 때의 동작도 확인했다: 강제 갱신이 아닌 경로에서 "매니페스트에 없는 기존 파일" 은 `UserCreated` 로 기록하고 건너뛴다 (`deployer.go:265-267`). 같은 파일을 다시 만나면 `UserCreated` 항목이 생긴 뒤이므로 "미등록" 이 아니게 된다. 이 경로는 해시 일치를 근거로 관리로 승격하지 않는 쪽의 선행 사례이다.

### (7) 백업·복원 바이트 동일, dry-run 무변경, 멱등, 중단 복구, 최종 잔존 목록

- 백업: `backupRemovedAsset` (`migrate_project_assets.go:84-104`) 이 제거 직전에 `<runDir>/removed-assets/<rel>` 로 바이트 사본을 쓴다. 쓰는 것은 바이트뿐이고 파일 모드는 `defs.FilePerm` 으로 정규화된다 (`:100`). 실행 디렉터리는 Backup 단계의 것을 재사용하고 (`update_template_sync.go:815-817` 의 `runBackupDir = configBackupPath` 와 마이그레이션 호출), 직접 호출이면 새로 만든다 (`update_disk_backup.go:65-78`).
- 복원: `moai update --restore` 는 `.moai/config` 만 복원한다 (`update.go:88`, `:282`). `removed-assets/` 를 되돌리는 코드는 이 트리에 없다. 따라서 "복원 바이트 동일" 은 현재 사본을 되돌려 놓는 수동 절차에 대해서만 말할 수 있다. 되돌린 파일은 매니페스트 항목이 이미 제거된 상태이므로 "미등록" 이 된다 — 이전 매니페스트가 함께 보존되지 않으면 복원이 소유권 정보를 잃는다.
- dry-run: `previewReconciliation` 은 읽기 전용 로더 `update.LoadManifestReadOnly` 를 쓰고 (`update_dryrun_preview.go:49`, `internal/cli/update/reconcile.go:563`) 아무것도 쓰지 않는다. 미리보기와 실행의 판정은 #4 와 #1 로 나뉘어 있다 (C 절).
- 멱등: `TestMigrationIdempotent` 가 첫 실행과 둘째 실행 사이 보관소 집합과 대표 파일을 비교한다. 잔존 목록, 매니페스트 바이트, 트리 해시의 반복 동일성은 단언하지 않는다.
- 중단: 마이그레이션은 제거 → 매니페스트 저장 → (선행 t1668) 정리 순서이다. 저장 전에 멈추면 매니페스트에는 사라진 파일의 항목이 남지만 엔트리 정리 루프가 다음 실행에서 걸러낸다 (`:346-352`). 제거 후 정리 전에 멈추면 빈 디렉터리가 남고 다음 실행의 걷기는 그것을 보지 않는다 (위 (5)). 실행 후 최종 상태 검증은 있다 (`verifyMigrationFinalState` `:573-601`, 호출 `update_template_sync.go:1304`).
- 잔존 목록: 파일별 구조화된 목록은 없다. 보고 문자열과 `migrationPreservedProjectFiles` (절대 경로 집합, 사유 없음, `update_migrate.go:127`) 뿐이다. 운영자 ruling d-20261012T221654Z-9ced 는 보존 파일을 기본으로 개수로, `--verbose` 에서만 경로로 보여 주게 정했다 (`update_dryrun_preview.go:37-39` 주석). 313 행을 기본 출력으로 낼 수는 없다.

### 제거 스위치와 안전 게이트 (건드리지 않는 것)

- `removalEnabled` 기본값은 `false` (`migrate_project_assets.go:38-39`). 비시험 코드에서 `true` 로 대입하는 곳은 없다. 시험 헬퍼 `enableRemovalForTest` 만 켠다 (`migrate_removal_switch_test.go:29-34`).
- 이 SPEC 은 스위치를 켜지 않고, 기본값을 바꾸지 않으며, 켜는 운영 경로를 만들지 않는다.

## E. 증거 원천 목록 (소유권 판정에 쓸 수 있는 것)

| 원천 | 위치 | 담고 있는 것 | 한계 |
|------|------|--------------|------|
| 프로젝트 매니페스트 항목 | `.moai/manifest.json` | `Provenance`, `TemplateHash`, `DeployedHash`, `CurrentHash`, `KeptBy` | 파일별 배포 버전 없음. 거울 항목은 등록되지 않음 |
| 매니페스트 수준 기록 | 같은 파일 | `Version`/`DeployedAt` (초기화 때만), `TemplateTreeHash` (갱신마다) | `Version` 이 갱신 배포를 따라가지 않음 |
| 사용자 매니페스트 | `~/.moai/user-assets.json` | 파일별 `sha256`, `bundle`, `installed_at`, `moai_version` (`internal/userassets/manifest.go` 머리말) | 사용자 쪽 설치의 기록이지 프로젝트 사본의 출처가 아님 |
| 런치 원장 | 프로필 기본 디렉터리의 원장 파일 (`internal/profile/profile.go:461`) | 프로젝트 → 프로필 매핑 | 프로필 루트의 파일 존재/해시는 담지 않음 |
| 실행 백업 | `.moai-backups/<run>/` | `.moai/config` 사본, `removed-assets/` | 이전 매니페스트 세대는 담지 않음 |

## F. 확인하지 못한 것 (Gaps)

- G-1: 현장 진단의 개별 수치 (에이전트 6·스킬 1, commands 17, 빈 디렉터리 3, 313) 는 사용자 전달값이며 이 조사가 재측정하지 않았다.
- G-2: commands 의 "이전 제외 ↔ 재배포" 가 가리키는 경로를 저장소 사실만으로 특정하지 못했다 (D(3) 의 후보 a-c).
- G-3: 미리보기 판정 #4 가 프로필 유지 항목을 "candidate" 로 알리는 조합 (스위치 켜짐 + 프로필 지도 + 표지) 은 코드 읽기로 도출했고 실행하지 않았다.
- G-4: 소문자 `skill.md` 가 마이그레이션 걷기에서 "unregistered" 로 분류된다는 결론은 맵 조회 의미 + 호스트 파일시스템 관측의 조합이며, 마이그레이션을 그 철자로 실행해 확인하지 않았다.
- G-5: `.agents/skills` 313 개가 심링크 거울 행인지 복사 폴백 파일인지 저장소만으로 알 수 없다.
- G-6: `TestMigrationClassification` 등 일부 기존 시험은 이름만 확인했고 본문을 읽지 않았다.
- G-7: 갱신 경로 전체 (`runTemplateSyncWithReporter`) 의 둘째 실행이 트리를 바이트 단위로 바꾸지 않는지는 이 조사에서 측정하지 않았다 (`TestMigrationIdempotent` 는 부분 단언).
- G-8: 대응 사용자/프로필 파일의 "프로필 쪽" 확인은 프로필 루트 읽기를 요구하므로 이 SPEC 의 구속 제약 (a) 아래 조사하지 않았다. t1702 가 보류 중인 영역이다.

## G. 실행한 명령 (읽기 전용 + 위 시험 실행)

- `git rev-parse --short HEAD`, `git branch --show-current`, `ls .moai/specs | grep -i -E "UPDATE|OWNERSHIP"`
- `grep -n` 으로 `migrate_project_assets.go`, `update_template_sync.go`, `update_dryrun_preview.go`, `update_migrate.go`, `deploy.go`, `deployer.go`, `skill_mirror.go`, `skillresidue/residue.go`, `update_cleanup.go`, `update_state_residue.go`, `manifest/*.go`, `types.go` 의 좌표 확인
- `grep -rn "KeptBy\|migrationPreservedProjectFiles\|ClassifyMigration\|EnvClaudeConfigDir\|removed-assets" internal --include='*.go'`
- `git ls-files | grep -i userroots` (결과 없음), `wc -l internal/config/protected_zone.go` (428)
- `ls internal/template/templates/.claude/{commands/moai,agents/moai,skills/moai-ref-owasp-checklist}` (20 / 15 / `SKILL.md` 1)
- 스크래치 프로그램: 대소문자 무구분 파일시스템 확인, `filepath.WalkDir` 철자, `os.SameFile`, 심링크 항목의 `IsDir`
- 기존 시험 11개 실행 (B 절)
- `go build -o /private/tmp/own-fixture-bin/moai ./cmd/moai`, `moai spec --help`
