---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "implementation plan"
version: "0.3.1"
created: 2026-10-18
---

# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — 구현 계획

## A. 컨텍스트

카드 t1865, Class C, P1. 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, 브랜치 `WT-update-ownership-repair`, plan-phase 출발 HEAD `cdd72a814` (t1668 의 `e1e163320` 병합 포함). 근거 문서: spec.md (요구사항 17, 인수 기준 21), research.md (코드 조사), design.md (데이터 모델), decision-index.md (결정 지점), acceptance.md (판정 명령).

개발 모드: RED-GREEN-REFACTOR (TDD). 모든 마일스톤은 실패하는 시험을 먼저 기록하고 (RED, 원문 출력 + 종료 코드 + 실패 이유 + 트리 SHA) 구현으로 닫는다. 제로 매치 셀렉터 (`[no tests to run]` + `ok`) 는 RED 도 PASS 도 아니다.

### Tier 분류: L

| 근거 | 값 |
|------|-----|
| 인수 기준 수 | 21 — Tier M 상한 16 초과 (Tier L 상한 25 이내) |
| 요구사항 수 | 17 — Tier M 상한 16 초과 (Tier L 상한 25 이내), 데이터 모델 결정이 포함됨 |
| 영향 파일 | 운영 코드 8~9 + 시험 16 = 24~25 개 안팎 (아래 § F 의 표) — Tier L 구간 (> 15) |
| 규모 | 새 코드와 시험을 합쳐 1,000 줄을 넘길 것으로 본다 (판정 함수, 증거 해석, 빈 디렉터리 정리, 잔존 행 렌더, 시험 파일 16개 — § F 의 시험 열을 센 값) |
| 구조 | 영속 매니페스트 데이터 모델 변경 (파일별 배포 버전, 이전 세대 보존) 과 네 곳의 소유권 판정 어휘 통합 |

Tier L 의 plan-audit 통과 기준은 0.85 이다. 감사가 반복 상한에 닿으면 분할선은 M1b + M8 (증거 생산: 배포 버전 기록, 이전 매니페스트 보존, 복원 일치) 과 나머지 (판정, 보존 이유, 잔존 목록, 대소문자, 빈 디렉터리, commands) 사이이다. Q2 는 리더가 옵션 (가) 로 조건부 결정했다 (보드 d-20261018T040918Z-def2) — (가) 가 이 계획의 기록이다. 옵션 (나) 의 뺄셈 지도 (접히는 것과 남는 것의 전체 목록) 는 기록된 대안으로 § G.1 에 남는다.

## B. 알려진 문제와 트리 상태 (이번 실행 관측, HEAD `cdd72a814`)

research.md 가 근거와 좌표를 가진다. 계획에 직접 영향을 주는 사실만 모은다.

1. 소유권 판정이 네 곳에 흩어져 있다: 마이그레이션 `migrate_project_assets.go:128-360`, 휴면 분류기 `update/migrate_classify.go:172` (비시험 호출자 없음), reconcile `update/reconcile_classify.go:282`, 미리보기 `update_dryrun_preview.go:120` (research.md C).
2. 매니페스트에는 파일별 배포 버전도, 이전 세대 보존도 없다 (`internal/manifest/types.go:93-113`; `Version` 은 초기화 때만, `initializer.go:607-608`).
3. 프로필 도달 유지는 사용자 대응물과 무관하게 `KeptBy` 표지를 붙인다 (`migrate_project_assets.go:255-272`). `KeptBy` 의 비시험 소비자는 아직 없다.
4. 마이그레이션 걷기는 이미 비어 있던 디렉터리를 정리하지 않는다 (`:195-197`, `:332-336`). 정리 판단의 근거가 되는 이전 매니페스트 기록이 없다.
5. 심링크 거울과 복사 폴백 거울은 의도적으로 매니페스트에 등록되지 않는다 (`skill_mirror.go:1-21`). 그래서 `.agents/skills` 아래 복사본은 구조적으로 미등록이다.
6. 소문자 `skill.md` 는 맵 조회에서 등록 키 `SKILL.md` 를 놓친다. 기존 이름 바꾸기 장치는 `.claude/skills` 의 배포되는 스킬만 덮는다.
7. 기존 시험 11개가 이 트리에서 통과한다 (research.md B). 이 계획의 모든 마일스톤은 그 11개와 t1668 의 시험을 GREEN 으로 유지한다.
8. 이 트리에 `internal/config/userroots.go` 는 없다. `internal/config/protected_zone.go` 는 있고 건드리지 않는다.
9. 이 트리의 Go 소스에는 `claude-skills@` 문자열이 없다 (acceptance.md § L-7). 기존 런치 원장 판정이 프로필 설정 디렉터리 항목 자체에 `os.Stat` 한다 (`launchCandidateIsUsable`, `profile.go:439`, § L-8) — 내용은 열지 않는다. 이 한 호출만이 REQ-OWR-005 의 이름 붙은 예외이다 (리더 판정 OQ-5). 갱신 경로 호출 닫힘에는 `internal/profile` 이 있고 내보낸 경로 도우미 (`GetProfileDir` 등) 가 있다 (§ L-11, L-12). 프로필 무접근 검증 설계는 design.md D-8.
12. 같은 버전의 일반 재실행은 조정 단계 앞에서 건너뛴다 (`update_template_sync.go:1239-1248`; 이번 실행 측정 § L-13). 여러 번 실행하는 기준은 `--force` 로 구동하고 도달을 단언한다 (design.md D-10).
13. 리더 결정 d-20261018T040918Z-def2: Q2 = (가), 조건부, 설계 결정일 뿐 (사용자 보안 승인 아님, 실제 홈 접근 승인 아님, Kickoff 사전 승인 아님). (나) 뺄셈 지도는 기록된 대안이다.
10. `probeCaseSensitiveFS` 는 비시험 호출자가 없다 (§ L-9). `moai update --dry-run` 전체 진입은 `update.go:361` 이고 기존 시험 헬퍼 `runDryRunUpdateInProject` 가 구동한다 (design.md D-9).
11. 템플릿 commands 17개는 모두 frontmatter 에 `name:` 이 없고 `description:` 은 있다. 에이전트 12개는 모두 `name:` 이 있다 (§ L-3). 템플릿 감사 시험 (`commands_audit_test.go`) 이 요구하는 것은 `description` 과 `allowed-tools` 이다.

## C. 사전 점검 (Pre-flight)

- [ ] **P0 — 판정 명령이 쓰는 `TMPDIR` 디렉터리를 만든다**: `mkdir -p /private/tmp/own-fixture-d1`. 이 단계가 없으면 `t.TempDir()` 이 실패한다 (acceptance.md 의 모든 판정 명령이 이 경로를 쓴다). 이름에 카드 토큰이 없고 `$HOME`·저장소 밖이다. P0 는 M0 의 첫 단계로도 실행하며 실행 결과를 progress.md § E.2 에 남긴다.
- [ ] 선행 카드 t1668 의 착지 상태 확인: `git log --oneline -n 5` 에 `e1e163320` 병합이 보이는지, t1668 이 main 에 착지했다면 main 을 이 트리에 다시 병합했는지 (착지는 t1668 뒤).
- [ ] `go build ./...` 성공.
- [ ] 기준선 시험 (research.md B 의 11개) GREEN: `TMPDIR=/private/tmp/own-fixture-d1 go test ./internal/cli/ -run '^(TestMigrationClassifiesUnregisteredMirrorCopy|TestMigrationKeepsSharedCopyUnderProfileReach|TestMigrationRemovesSharedCopyWithoutProfileReach|TestEmptiedSkillDirRemoved|TestMigrationBackupBeforeRemoval|TestMigrationDefaultKeepsConfirmedProjectCopy|TestMigrationRemovalGate|TestMigrationIdempotent|TestMigrationPreservesExistingMirrorEntries|TestRF2F3b_MigrationPreservesUntracked|TestPreviewOmitsRemovalCandidateWhileSwitchOff)$' -count=1 -v` — 쓸어낸 시험 수가 11 인지 확인한다. (이번 plan 실행 측정, 트리 `15db036a3`: 11 PASS, `ok ... 44.908s`, 종료 코드 0 — acceptance.md § L-4.)
- [ ] `git status --short` 가 깨끗한지, `git rev-parse --short HEAD` 와 `git branch --show-current` 를 매 커밋 직전에 다시 읽는다.
- [ ] 이 트리의 건너뛰기 관문 측정 재확인: 같은 픽스처에 일반 실행 1 은 건너뛰지 않고 일반 실행 2 는 건너뛴다 (acceptance.md § L-13). 여러 번 실행하는 모든 시험은 `--force` 헬퍼 `runUpdateFlowForced` 와 도달 단언(규칙 7)을 쓴다.
- [ ] 시험 격리 확인: 새 시험은 `t.TempDir()` + `t.Setenv("HOME", …)` + `CLAUDE_CONFIG_DIR` 명시 지정, `profile.BaseDirOverride` 를 임시 디렉터리로 두고 복원한다 (`profileReachSandbox` 패턴, `migrate_profile_reach_test.go:29-40`). `TMPDIR` 는 `/private/tmp` 아래 카드 토큰이 없는 이름.

## D. 제약

1. 리더 구속 제약 (spec.md § F 1) 과 t1668 안전 게이트 (§ F 2) 가 모든 마일스톤에 우선한다. 프로필 루트 열거·쓰기·공급이나 F2 보호가 필요해지는 순간 구현하지 않고 멈춰 blocker 보고를 낸다.
2. 시험 제외, 제한시간 증가, 가드/감사 우회, 거절 기록 삭제, 승인창 선택 금지. 가드나 감사가 추가 승인을 요구하면 막힘으로 보고한다.
3. 커밋은 명시한 경로만 스테이징한다 (`git add -A`, `git add .`, `git commit -a` 금지). 브랜치 상태를 바꾸지 않는다.
4. 한 번에 한 파일씩, 의존성 순서로 (Multi-File Decomposition). 같은 작업 트리에는 쓰는 에이전트가 하나만 있다.
5. 새 판정 코드는 새 파일에 둔다 (`migrate_ownership*.go`, `update_prior_manifest.go`). `migrate_project_assets.go` 와 `update_template_sync.go` 의 변경은 호출 줄 교체·추가로 좁힌다 — t1668 이 같은 파일을 더 고치면 병합 충돌이 이 두 파일에 모인다. 이 노출은 줄이되 없애지 못한다: `update_template_sync.go` 는 M1b·M3·M6(조건부) 세 곳, `migrate_project_assets.go` 는 M1a·M2·M3·M5 네 곳에서 호출 줄이 바뀐다 (M8 은 시험 전용으로 줄였다). t1668 착지 뒤 main 병합에서 이 두 파일의 충돌 가능성은 수용된 잔여 위험이다 (decision-index.md 수용 부채).
6. 템플릿 파일(`internal/template/templates/**`)은 바뀌지 않는다. 따라서 Template-First 재생성(`make build`)의 대상이 아니다. 매니페스트 필드 추가는 Go 코드 변경이며 템플릿이 아니다.

## E. 자가 검증

run 단계 종료 시 progress.md §E.2 에 다음을 원문으로 기록한다.

- AC-OWR-01 ~ -21 의 판정 명령 원문 출력과 종료 코드 (acceptance.md). 각 RED 셀: 명령, 원문 stdout, 종료 코드, 고정한 트리 SHA, 실패 이유.
- 기준선 11개 + t1668 시험의 재실행 원문 (회귀 가드).
- `go test -cover` — 터치한 패키지 (`internal/cli`, `internal/manifest`) 85% 이상.
- `golangci-lint run internal/cli/... internal/manifest/...`, `go vet`, `gofmt -l`.
- `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/config/protected_zone.go internal/config/userroots.go` 빈 출력·종료 코드 0 과, 같은 형식의 양성 대조군 (`-- .moai/specs`) 이 한 줄 이상 (REQ-OWR-005 의 파일 불변 증거. main 병합이 두 파일을 바꿔도 거짓 적색이 되지 않는다 — acceptance.md AC-OWR-08).
- AC 승격 기록: 각 AC 의 시험 커밋 A → 셀 기록 커밋 B → 구현 커밋 C 의 SHA 세 개 (acceptance.md 「등급과 RED 셀 규약」).
- 갭 목록 갱신. 빈 Gaps 선언은 사실일 때만.

## F. 파일 소유와 의존

| 마일스톤 | 운영 코드 | 시험 |
|----------|-----------|------|
| M0 | — | `internal/cli/migrate_ownership_helpers_test.go` (`runUpdateFlowForced` — `--force` 로 조정에 닿게 하고 `skipped` 와 새 `manifest.prior.json` 유무를 돌려줌; 공유 픽스처: 등록/미등록/수정/사용자 작성 파일, 이전 세대 매니페스트 빌더, 프로필 샌드박스와 카나리, 디렉터리 mtime 을 담는 트리 해시, 경로 지정 사용자 대응물 설치 도우미), `internal/cli/testdata/profile_callsite_baseline.txt` (프로필 심볼·토큰·`internal/profile` 안 파일시스템 동사의 (파일, 심볼) 별 호출 지점 개수 — `go list -deps ./internal/cli` 닫힘을 손대지 않은 트리로 측정해 얼린 기준선, 예외 행 하나 포함) |
| M1a | `internal/cli/migrate_ownership.go` (신규: 분류, 처분, 이유 코드, `classifyOwnership`), `internal/cli/migrate_ownership_evidence.go` (신규: `provenanceSource` 인터페이스, 해석기 `provenanceEvidence`, 원천을 얻는 패키지 수준 변수), `internal/cli/migrate_project_assets.go` (판정 호출 교체) | `internal/cli/migrate_ownership_test.go`, `internal/cli/migrate_ownership_evidence_test.go` |
| M1b | `internal/manifest/types.go`, `internal/manifest/manifest.go`, `internal/cli/update_prior_manifest.go` (신규: `priorManifestWriter` 변수와 원자적 이전 매니페스트 쓰기 — 임시 쓰기와 이름 바꾸기는 `priorManifestTempWrite`·`priorManifestRename` 관측점 변수로만, 보존 실패 결과 전달, 이전 세대 원천 읽기), `internal/cli/migrate_project_assets.go` (제거 암을 「후보 수집 → 전체 백업 확인 → 일괄 제거」 로 나눔; REQ-OWR-016), `internal/cli/update_template_sync.go` (호출 줄만: Backup 단계 보존 호출, 배포 단계 저장의 `Version`·`DeployedAt` 갱신 호출) | `internal/manifest/manifest_deployed_version_test.go`, `internal/cli/update_prior_manifest_test.go`, `internal/cli/update_prior_manifest_io_test.go` (AC-OWR-21) |
| M2 | `internal/cli/migrate_ownership.go`, `internal/cli/migrate_project_assets.go` (호출 줄) | `internal/cli/migrate_ownership_profile_test.go`, `internal/cli/migrate_ownership_noprofileaccess_test.go` (심볼 단위 스캔 + 변이 대조군 + 정규화 카나리) |
| M3 | `internal/cli/migrate_ownership.go`, `internal/cli/migrate_project_assets.go` (행 수집 호출), `internal/cli/update_template_sync.go` (렌더 호출 줄) | `internal/cli/migrate_ownership_remaining_test.go` |
| M4 | `internal/cli/migrate_ownership.go` (별칭 확정: 파일시스템 모드 프로브 없이 관측된 동일성으로) | `internal/cli/migrate_ownership_case_test.go` |
| M5 | `internal/cli/migrate_ownership.go`, `internal/cli/migrate_project_assets.go` (정리 단계 호출) | `internal/cli/migrate_ownership_emptydir_test.go` |
| M6 | (조건부, M0 의 탐침 결과에 따라) `internal/cli/update_template_sync.go` 의 호출 줄 또는 없음 | `internal/cli/update_commands_discovery_test.go` |
| M7 | `internal/cli/update_dryrun_preview.go` | `internal/cli/migrate_ownership_preview_test.go` |
| M8 | — (운영 코드 변경 없음: 이전 매니페스트 보존은 M1b 가 만든다. 복원 일치는 시험 전용 절차) | `internal/cli/migrate_ownership_restore_test.go` |
| M9 | — (수렴 검증) | `internal/cli/migrate_ownership_recovery_test.go` |
| M10 | — | `internal/cli/migrate_ownership_gate_test.go` |

의존하는 t1668 코드: `backupRemovedAsset`, `pruneEmptiedSkillDir`, `migrationRunContext`, `verifyMigrationFinalState`, `removalEnabled`, 시험 헬퍼 `enableRemovalForTest`, `removalBackupCopy`, `runUpdateFlowWithProgress`, `buildMigrationFixture`, `installMigrationUserCounterparts`, `profileReachSandbox`, 그리고 기존 dry-run 헬퍼 `runDryRunUpdateInProject`·`hashDryRunTree`. 이 이름들이 바뀌면 이 계획의 해당 마일스톤을 다시 쓴다.

## G. 마일스톤

순서 원칙 (결정 번복 가능성이 높은 것 먼저): 영속 데이터 모델과 판정 어휘 (M1) → 프로필 이유 (M2) → 잔존 목록 형태 (M3) → 대소문자 (M4) → 빈 디렉터리 (M5) → commands (M6) → 미리보기 통일 (M7) → 복원 일치 (M8) → 수렴 검증 (M9) → 게이트와 품질 측정 (M10, 기계적, 마지막).

**모든 마일스톤의 RED→GREEN 규약** (acceptance.md 「등급과 RED 셀 규약」): 해당 AC 의 시험 파일만 담은 커밋 A → A 의 트리 SHA 에서 판정 명령을 돌려 명령·원문 stdout·종료 코드·트리 SHA·실패 이유를 progress.md § E.2 에 기록하는 커밋 B → 구현 커밋 C. 이 순서는 커밋 그래프가 증명한다 (verification-claim-integrity.md § 2.3). RED 는 시험이 존재하는 상태의 실패 (빌드 실패 또는 단언 실패) 여야 하며 제로 매치 셀렉터는 RED 가 아니다.

### M0 — 기준선과 공유 픽스처 — Priority High

1. P0 (`mkdir -p /private/tmp/own-fixture-d1`) 를 실행한다. 기준선 시험 11개 + t1668 시험을 실행해 GREEN 원문을 progress.md § E.2 에 남긴다 (쓸어낸 수 확인).
2. 공유 시험 헬퍼를 `migrate_ownership_helpers_test.go` 에 만든다: 상태별 파일 빌더 (등록 managed, 등록 user-modified, 등록 user-created, **내장 템플릿의 실재 경로에서 읽은 바이트로 만든 미등록 파일**, 템플릿 바이트와 다른 출처 불명 파일), 이전 세대 매니페스트 빌더, 프로필 샌드박스 (`CLAUDE_CONFIG_DIR` 지정/해제) 와 카나리 프로필 디렉터리 빌더 (세 상태), 경로를 지정하는 사용자 대응물 설치 도우미, 트리 해시 계산 (경로 + 바이트 + 모드 + **디렉터리 mtime**, 런 백업 디렉터리 제외 옵션; 실행 전에 모든 디렉터리 mtime 을 고정 과거 시각으로 맞추는 도우미).
3. 프로필 호출 지점 기준선을 측정해 `testdata/profile_callsite_baseline.txt` 에 얼린다 (손대지 않은 트리, 측정 명령과 트리 SHA 를 파일 머리에 적는다): `go list -deps ./internal/cli` 닫힘 (이 트리 142개 패키지, § L-11) 의 비시험 파일에서 `internal/profile` 내보낸 식별자 호출 지점·문자열 토큰·상수 아닌 환경 변수 읽기를 (파일, 심볼) 별로 세고, `internal/profile` 안에서는 함수 단위 파일시스템 동사 호출을 센다. 이름 붙은 예외 행 하나 (`launchCandidateIsUsable` · `os.Stat` · 1) 를 기준선에 표시한다. 종전 19개 파일 토큰 기준선 (§ L-10) 은 정보로 남는다.
3a. `runUpdateFlowForced` 를 만든다 (`--force`, `runTemplateSyncWithProgress` 의 `skipped` 와 실행 백업의 새 `manifest.prior.json` 유무를 돌려줌). 건너뛰기 관문 측정 (§ L-13) 을 progress.md § E.2 에 재기록한다.
4. 헬퍼는 실제 홈을 읽거나 쓰지 않는다. 임시 HOME 만.

산출물: 기준선 원문, 공유 헬퍼, 기준선 데이터. 터치 파일: 시험 1개 + 데이터 1개.

### M1 — 소유권 판정과 증거 모델 — Priority High (번복 가능성 최대)

**M1a — 판정 함수, 이유 코드, 증거 해석기 (스키마 변경 없음)**

1. RED: `TestOwnershipVerdictClasses` (AC-OWR-01), `TestOwnershipPreserveByteEqualUserFile` (AC-OWR-02, 전제 단계 + 보존 + 승격 변이 대조군), `TestOwnershipPreserveUserModifiedFile` (AC-OWR-03), `TestOwnershipUnregisteredUnknownOriginKeptWithReason` (AC-OWR-04), `TestOwnershipAttestedManagedOnlyWithFullEvidence` (AC-OWR-05, 증거를 시험이 직접 공급 — 변이 네 가지). 모두 새 판정 함수와 이유 코드를 단언하므로 현재 트리에서 빌드 단계에서 실패하거나 단언에서 실패해야 한다 — 실패 이유를 RED 셀에 적는다.
2. GREEN: `migrate_ownership.go` 에 소유권 분류, 처분 (`removal-eligible` 은 spec.md § B 의 정의), 이유 코드 (design.md D-1, D-3), `classifyOwnership` 를 만든다. `migrate_ownership_evidence.go` 에 `provenanceSource` 인터페이스와 해석기 `provenanceEvidence` (design.md D-2), 그리고 운영 경로가 원천을 얻는 패키지 수준 변수를 만든다 (이 단계의 운영 원천은 항상 "증거 없음" — 이전 세대를 읽는 원천은 M1b). `migrate_project_assets.go:198-302` 의 인라인 순서를 이 함수 호출로 바꾼다. 처분은 현재와 같아야 한다 — 특성화 시험(기준선 11개)이 GREEN 으로 남는 것이 증거다.
3. 보호 장치: 판정 함수는 파일 시스템을 읽기만 한다 (쓰기 없음). 미리보기가 같은 함수를 부를 수 있어야 한다 (M7).

**M1b — 배포 버전 기록, 이전 매니페스트 보존, 이전 세대 원천 (decision-index.md Q2 (가) 확인 뒤)**

1. RED: `TestManifestRecordsDeployedVersion` (`internal/manifest`), `TestPriorManifestRetainedInRunBackup`, `TestPriorGenerationSourceBoundedAndTolerant` (AC-OWR-06 (a)(b)(c)), `TestBackupFailureRemovesNothingThroughProductionPath` (AC-OWR-20, 전체 흐름), `TestPriorManifestAtomicWriteFailures`, `TestPriorGenerationCorruptIsNoEvidence`, `TestLegacyManifestRoundTrip`, `TestDowngradeDropsEvidenceNotSafety`, `TestPriorGenerationsAreLookupBoundOnly` (AC-OWR-21), 그리고 AC-OWR-05 (h) 의 운영 원천 두 건.
2. GREEN: `FileEntry` 에 `deployed_version` (omitempty) — 이번 실행이 배포한 항목에만 찍고 배포하지 않은 항목은 채우지도 지우지도 않는다 (재추적 경로 보존), 갱신의 배포 단계 저장에서 매니페스트 수준 `Version` 을 갱신하고 `DeployedAt` 은 `Version` 이 바뀔 때만 갱신 (`recordTemplateTree` 와 같은 자리, `update_template_sync.go:672-673`; design.md D-3 항목 1-2; 저장은 기존 `Save` 의 임시 파일 + 이름 바꾸기). Backup 단계가 시작 시점 매니페스트의 **원본 바이트**를 `<runDir>/manifest.prior.json` 으로 **임시 파일 + 이름 바꾸기로** 보존하고 (`priorManifestWriter` 변수 경유; 임시 쓰기와 이름 바꾸기는 `priorManifestTempWrite`·`priorManifestRename` 변수로만 수행 — design.md D-3 항목 3) 그 성공 여부를 실행 문맥에 넘긴다. 마이그레이션 제거 암은 「후보 수집 → 모든 바이트 사본과 매니페스트 사본 확인 → 일괄 제거」 로 나뉘어, 어느 백업이든 실패하면 그 실행의 제거가 0건이다 (REQ-OWR-016, design.md D-3 항목 3-5). 이전 세대 원천 (최근 순으로 최대 `priorManifestGenerations` 개) 을 `update_prior_manifest.go` 에 만들어 M1a 의 패키지 수준 변수에 꽂는다.
3. 호환: `deployed_version` 이 없는 기존 항목은 읽기에서 그대로 유효하다 (기존 동작 유지). 증거 부족은 미등록 파일의 `managed` 승격만 막는다.
4. Q2 는 (가) 로 결정되었다 (이 계획의 기록). (나) 였다면 M1b 를 접었을 것이다 — 그 기록된 대안의 목록은 § G.1.
5. 보존 기간: `priorManifestGenerations` (8) 는 조회 한도일 뿐이다. 이 마일스톤은 삭제·정리·보존 기간 단축을 더하지 않고 `internal/cli/update/backup/` 를 건드리지 않는다 (design.md D-3 항목 8, AC-OWR-21 (e)).

산출물: AC-OWR-01 ~ -06, -20, -21 GREEN. 터치 파일: 위 표.

### M2 — 프로필 도달 유지 이유 — Priority High

1. RED: `TestProfileMappedCopyKeptWithoutCounterpart` (AC-OWR-07) — 에이전트 여러 개와 스킬 1개를 가진 프로필 지도 프로젝트, 사용자 루트 대응물 없음. 사본은 제자리에 남고, 제거 스위치를 픽스처에서 켜도 제거되지 않고, 잔존 행에 `profile-reach` + 대응물 상태 `absent` + 프로필 쪽 `not-checked` 가 있어야 한다. `TestProfileEnvDoesNotReachProfileRoot` (차분 카나리), `TestNoProfileRootAccessInUpdateSources` (기준선 대비 소스 스캔) (AC-OWR-08, design.md D-8).
2. GREEN: 프로필 유지 분기 (`:255-272`) 는 `KeptBy` 쓰기를 그대로 두고 (Q1 DECIDED), 대응물 상태 판정 (`userCounterpartConfirmed` 재사용) 을 이유에 붙인다. 프로필 디렉터리를 여는 코드는 추가하지 않는다.
3. 중단 조건: 구현 중 프로필 쪽 파일을 읽어야만 요구가 충족되는 지점이 나오면 멈추고 blocker 보고를 낸다 (decision-index.md Q5 — 리더 판정 2026-10-18 로 읽지 않기로 확정됨).

산출물: AC-OWR-07, -08 GREEN.

### M3 — 잔존 목록 — Priority High

1. RED: `TestRemainingListKeeps313UnregisteredCopies` (AC-OWR-14) — **내장 템플릿의 실재 파일 바이트로 만든** 복사 폴백 거울 흉내 313개 (`.claude/skills` 의 315개 파일에서 정렬 순으로 313개를 `.agents/skills/<같은 상대 경로>` 에 둔다; 전제 단계가 각 원본의 내장 조회 성공과 바이트 일치를 먼저 단언한다). 전부 보존되고, 매니페스트는 바이트 단위로 변하지 않고 (허용된 배포 기록 제외), 잔존 행 313개가 이유 `no-provenance-record` 와 빠진 증거 항목을 가지며, 기본 렌더는 이유별 개수 한 줄, `--verbose` 렌더는 파일별 행이다. 바이트 일치 승격 변이 대조군도 같은 픽스처에서 붉게 잡힌다.
2. GREEN: 마이그레이션이 잔존 행을 모아 `migrationRunContext` (또는 반환값) 로 돌려주고, 갱신 흐름이 터미널 참조 행으로 렌더한다 (decision-index.md Q3 (가)). 정렬은 경로 사전순으로 결정적이다.

산출물: AC-OWR-14 GREEN.

### M4 — 대소문자 변형 — Priority Medium

1. RED: `TestCaseVariantResolvesToRegisteredEntry` (파일시스템 대소문자 모드를 주입해 양쪽 모드를 같은 시험에서 단언), `TestCaseVariantOnHostFilesystem` (실제 파일시스템에서 **시험 자신의 임시 디렉터리에서** 프로브한 결과에 맞는 기대를 단언 — 시험을 건너뛰지 않는다), `TestCaseVariantCounterpartAndMirrorSpelling` (AC-OWR-11, -12).
2. GREEN: `resolveCaseAlias` 를 `migrate_ownership.go` 에 둔다 (design.md D-4). 판정 함수가 걷기 경로로 정확한 키를 먼저 찾고, 빗나가면 `strings.EqualFold` 키 후보 + `os.SameFile` 로 별칭을 확정한다. 파일시스템 모드를 따로 판정하지 않는다 — 대소문자 구분 파일시스템에서는 `SameFile` 이 거짓이라 별칭이 성립하지 않는다. 이름 바꾸기는 하지 않는다 (Q4 (가)).
3. 운영 코드는 `probeCaseSensitiveFS` 를 부르지 않는다 (비시험 호출자가 없는 함수이고 프로젝트 루트에 쓴다 — dry-run 무변경과 충돌, design.md D-4). 종전 계획의 "기존 프로브 재사용" 문장은 이 단계와 모순이어서 철회했다.

산출물: AC-OWR-11, -12 GREEN.

### M5 — 관리 파일만 있던 빈 디렉터리 정리 — Priority Medium

1. RED: `TestManagedOnlyEmptyDirsPruned` (AC-OWR-13) — 세 경우: 이번 실행이 비운 디렉터리, 이전 실행이 비웠던 디렉터리, 하위 폴더만 남은 중첩 빈 디렉터리. 반례: 출처 없는 빈 디렉터리는 유지 + 이유 기록, 사용자 파일이 있는 디렉터리 유지, 스킬 루트 자체는 정리하지 않음.
2. GREEN: 걷기 뒤에 별도 정리 단계를 둔다. 후보 디렉터리마다 이전 매니페스트 기록(시작 시점 매니페스트 + 보존된 이전 세대)에서 그 아래 기록된 파일이 전부 `template_managed` 이고 전부 디스크에 없음을 확인한 뒤 아래쪽부터 `os.Remove` 한다. `pruneEmptiedSkillDir` (t1668) 는 재사용하고 다시 쓰지 않는다. 에이전트 디렉터리와 루트는 정리 대상이 아니다.
3. 정리 단계는 제거 스위치와 무관하다 (제거가 아니라 비어 있는 디렉터리의 정리). 그러나 파일을 지우지는 않는다. `os.Remove` 는 빈 디렉터리만 지운다.

산출물: AC-OWR-13 GREEN.

### M6 — commands 한 처분과 탐색 보증 — Priority Medium (조건부)

1. 먼저 측정: `TestCommandsSingleDispositionAndPresent` 의 RED 단계가 commands 픽스처(템플릿 commands 17개를 managed, user-modified, 미등록 바이트 동일, 사용자 작성으로 나눔) 를 **두 변형**으로 구동해 현장의 "이전 제외 후 재배포" 가 재현되는지 기록한다 (research.md G-2 의 후보 a-c). **V1 같은 버전**: 기록과 바이너리 버전이 같다 (`runUpdateFlowWithProgress` 경로). **V2 낡은 기록**: 프로젝트의 템플릿 버전 기록과 매니페스트 `Version`·`TemplateTreeHash` 가 실행 중인 바이너리보다 오래된 릴리스의 값이다 — 현장 증상이 rc.35 템플릿 기록 + rc.36 바이너리였다. V2 는 `versionMatchSkipsSync` 를 거치는 `runUpdate` 진입으로 구동하고 `version.Version` 은 기존 헬퍼가 쓰는 시임으로 둔다. 한 변형만 돌리고 "재현 안 됨" 으로 닫지 않는다. V2 를 기존 시임만으로 구동할 수 없으면 그 사실과 이유를 RED 셀에 적고 한계로 남긴다 (acceptance.md 갭 G-A).
2. 재현되면: 원인 경로에 따라 한 실행에 한 처분만 남기도록 고치고 (파일 소유: 원인 파일), 재현되지 않으면 AC-OWR-09, -10 은 회귀 가드로 남기고 REQ-OWR-006 의 존재 보증만 시험으로 닫은 뒤 리더에게 "재현 안 됨" (어느 변형을 돌렸는지 포함) 을 보고한다.
3. `TestCommandsAndAgentsDiscoverableAfterInterruptedRun` (AC-OWR-10) — 중단 후 복구한 실행 뒤에도 `.claude/commands/moai/*.md`, `.claude/agents/moai/*.md` (harness 에 따라 `.codex/agents/moai/*.toml`) 가 템플릿이 운반하는 집합과 같다. commands 와 agents 의 frontmatter 조건은 분리한다 (commands: `description`·`allowed-tools`, agents: `name`·`description`).

산출물: AC-OWR-09, -10.

### M7 — 미리보기 판정 통일과 dry-run — Priority Medium

1. RED: `TestDryRunChangesNothingAndMatchesVerdict` (AC-OWR-16) — `moai update --dry-run` 이 디스패치하는 분기 전체를 `runDryRunUpdateInProject` 로 구동하고, 실행 전 디렉터리 mtime 을 고정 과거 시각으로 맞춘 트리 해시가 전후 같으며 `.moai-backups` 미생성, 제거 스위치를 픽스처에서 켠 미리보기의 후보 집합 == 판정 함수의 `removal-eligible` 집합. 프로필 지도 + `KeptBy` 항목은 후보에 없어야 한다 (research.md G-3 의 불일치). 판정·미리보기 새 코드에 쓰기 가능 호출이 없음을 소스 스캔으로 단언한다.
2. GREEN: `previewReconciliation` 의 `migrationRemovalCandidate` (`:120`) 호출을 판정 함수 호출로 바꾼다. 읽기 전용 로더 `LoadManifestReadOnly` 를 유지한다.

산출물: AC-OWR-16 GREEN.

### M8 — 백업·복원 바이트 동일 — Priority Medium

1. RED: `TestRemovedAssetBackupRestoreByteIdentical` (AC-OWR-15) — 픽스처에서 제거 스위치를 켜고 실행 (실행 1, 일반 — 시작 매니페스트는 비정규 모양), 같은 픽스처의 독립 복사본에 스위치를 끄고 같은 흐름을 돌린 통제 실행 T0 를 함께 둔다. 단언은 AC-OWR-15 와 같다: 제거된 각 파일의 백업 사본 sha256 == 제거 전 sha256; `<runDir>/manifest.prior.json` 바이트 == 실행 전 매니페스트 파일 바이트; 복원 영역 R = {제거된 파일 전부, 매니페스트 파일} 을 사본으로 되돌린 트리 T2 에서 R 의 각 파일 바이트 == 실행 전 바이트; 나머지 차이 센티널 — T2 와 T0 의 차이 경로 집합 D 가 비어 있지 않고 D ⊆ E (E = `.moai/manifest.json`, `.moai-backups/**`, `.moai/logs/**`, `.moai/state/**`, 정규화 뒤) 이며 E 밖 차이는 붉다. 되돌린 전체 트리의 해시가 실행 전과 같다는 조건은 요구하지 않는다 — 전체 흐름이 `llm.yaml` 을 다시 쓰고 템플릿을 배포하므로 어떤 올바른 구현도 통과하지 못한다 (AC-OWR-15 서두, 감사 N4·N13).
2. GREEN: 시험 안의 복원 절차(사본 되돌리기)는 시험 전용이며 운영 하위 명령이 아니다. 운영 코드 변경 없음 — 이전 매니페스트 보존은 M1b 가 만든다 (Q2 (나) 이면 파일 바이트 일치만 단언한다).

산출물: AC-OWR-15 GREEN.

### M9 — 반복 갱신 멱등과 중도 중단 복구 — Priority High

1. RED: `TestRepeatedUpdateIsIdempotent` (AC-OWR-17: 실행 2·3 은 `--force`, 도달 단언, 건너뛰는 일반 재실행 사례 분리), `TestInterruptedMigrationRecoversAndConverges` (AC-OWR-18). 중단 모사는 운영 코드에 시험용 끼어들기를 넣지 않고 픽스처로 한다: (a) 한 후보의 실행 백업 경로 자리에 디렉터리를 두어 백업을 쓸 수 없게 만든다 — 그 실행은 제거가 막혀 파일을 하나도 제거하지 않는다 (REQ-OWR-016: 부분 진행은 없다). 단언은 제거 목록이 비어 있고, 보존된 파일이 잔존 행에 이유 `backup-failed` 로 나열됨이다 (AC-OWR-18 (a), AC-OWR-20), (b) 매니페스트를 실행 전 바이트로 되돌려 저장 전 중단을 모사한다, (c) 정리된 빈 디렉터리를 다시 만들어 정리 전 중단을 모사한다. 복구 후 재실행은 중단 없는 실행의 최종 상태와 같아야 한다.
2. GREEN: 대부분 M1-M5 가 닫는다. 남는 차이만 고친다 (예: 되살아난 매니페스트 항목, 남은 빈 디렉터리).

산출물: AC-OWR-17, -18 GREEN.

### M10 — 게이트 무결성과 품질 측정 — Priority Medium (기계적, 마지막)

1. `TestRemovalSwitchStaysOffInProduction` (AC-OWR-19): 비시험 코드에 `removalEnabled = true` 대입이 없음을 소스 스캔으로 단언하고, 기존 `TestMigrationRemovalGate`, `TestMigrationDefaultKeepsConfirmedProjectCopy` 가 GREEN 임을 함께 기록한다.
2. `go test -cover` (터치한 패키지 85% 이상), `golangci-lint run`, `go vet`, `gofmt -l`.
3. 결과를 progress.md § E.2 에 기록하고 § E.3 을 작성한다.

### G.1 Q2 = (나) 일 때의 뺄셈 지도 — 기록된 대안 (감사 D6)

**Q2 는 (가) 로 결정되었다 (리더, 보드 d-20261018T040918Z-def2, 2026-10-18). 이 계획의 기록은 (가) 이며 이 표는 (나) 였을 때의 뺄셈을 기록으로만 남긴다.** (나) 는 아래 "접힘" 열을 제거하고 "축소" 열을 줄이는 것만으로 성립해야 한다. "남음" 열은 손대지 않는다. 이 표는 decision-index.md Q2 의 결과 목록과 같은 내용이며, 어느 쪽이 바뀌면 다른 쪽을 같이 고친다.

| 구분 | 요구사항 | 인수 기준 | 마일스톤·코드 |
|------|----------|-----------|----------------|
| **접힘 (삭제)** | REQ-OWR-015, REQ-OWR-016 (매니페스트 사본 반쪽; 파일 바이트 사본 실패 시 일괄 보존은 남음), REQ-OWR-017 | AC-OWR-06, AC-OWR-20 (매니페스트 사본 하위 사례), AC-OWR-21, AC-OWR-05 (h) | M1b 전체: `deployed_version` 필드, 매니페스트 수준 `Version`/`DeployedAt` 갱신, `manifest.prior.json`, 이전 세대 원천 읽기 (`update_prior_manifest.go`), 보존 실패 전달 |
| **축소** | REQ-OWR-010 → 파일 바이트 사본 반쪽만. REQ-OWR-008 → "이번 실행이 비운" 반쪽만 (t1668 의 `pruneEmptiedSkillDir` 가 이미 한다). REQ-OWR-013 → "비어 있는 관리 디렉터리가 남지 않음" 절은 이번 실행이 비운 것에 한정 | AC-OWR-15 → 매니페스트 사본 단언 제외. AC-OWR-13 → 사례 A·C 삭제, B·D·E·F 는 t1668 코드에 대한 회귀 가드 (새 코드 없음). AC-OWR-18 → 하위 사례 (c) 삭제, (a)(b) 남음. AC-OWR-17 → `DeployedAt` 갱신 변이 삭제 (AC-OWR-17 은 **축소** 열에만 있다 — 남음 열에서 뺐다, 감사 N10) | M5 접힘 (새 코드 없음, 시험은 B·D·E·F 가드로 남김). M8 → 파일 바이트 일치만 |
| **남음 (함수 수준 포함)** | REQ-OWR-001, 002, 003, 004, 005, 006, 007, 009, 011, 012, 014 | AC-OWR-01, 02, 03, 04, **05 (a)-(g) (증거를 시험이 직접 공급하는 함수 수준 — M1a 의 해석기가 필요로 하는 코드는 M1a 에 있으므로 남는다)**, 07, 08, 09, 10, 11, 12, 14, 16, 19 | M0, M1a, M2, M3, M4, M6, M7, M9, M10 |

**(나) 의 운영자 가시 결과**: 현장 증상 (5) 중 이전 실행이 비워 둔 빈 디렉터리 3개는 운영에서 정리되지 않는다. 현장 증상 (6) 의 양성 경로 (증거 완비 파일이 managed 로 판정됨) 는 운영에서 도달할 수 없고 함수 수준에서만 시험된다. 이 둘은 (나) 를 고르는 순간 카드 범위 항목 (5)·(6) 의 일부가 의도적으로 미이행이 된다는 뜻이다.

## H. 안티 패턴

- 소유권 판정을 다섯 번째 장소에 새로 쓰는 것 — 판정 함수 하나만 있어야 한다 (REQ-OWR-001).
- 바이트 일치 (내장 템플릿 트리 포함) 를 `managed` 근거로 쓰는 것 (REQ-OWR-002).
- 미등록 파일을 자동으로 매니페스트에 채택하거나 `UserCreated` 로 바꿔 쓰는 것 (REQ-OWR-003, 이 SPEC 은 매니페스트 채택을 하지 않는다).
- 프로필 디렉터리를 열거하거나 읽는 것 (REQ-OWR-005). 기준선 대비 소스 스캔과 차분 카나리가 지킨다 (design.md D-8). `claude-skills@` 문자열 스캔만으로 지켰다고 주장하는 것 — 그 문자열은 이 트리의 Go 소스에 없어서 어떤 구현도 통과시킨다.
- 0o000 디렉터리 읽기 오류에 기대어 "읽지 않았다" 를 주장하는 것 — root 에서 무력하고 오류를 삼키는 읽기를 통과시킨다.
- 대소문자 시험을 파일시스템에 따라 `t.Skip` 하는 것 — 시험 제외다. 양쪽 모드를 주입으로 단언하고, 실제 파일시스템 시험은 프로브 결과에 맞는 기대를 단언한다.
- 제로 매치 셀렉터를 통과로 보고하는 것.
- 없는 RED 를 있다고 적는 것 — 시험이 아직 없는 AC 에 RED 셀을 미리 채우거나, 시험 커밋과 구현을 한 커밋에 묶어 RED 가 구현 앞에 있었음을 커밋 그래프로 증명하지 못하게 하는 것.
- dry-run 경로에서 "만들었다 지우는" 임시 쓰기 (`probeCaseSensitiveFS` 류) 를 하는 것 — 트리 해시의 디렉터리 mtime 이 잡는다.
- 이전 매니페스트 사본 쓰기 실패를 경고로만 처리하고 제거를 계속하는 것 (REQ-OWR-016).
- 313개를 기본 출력에 모두 찍는 것 (운영자 ruling d-20261012T221654Z-9ced).
- 운영 코드에 중단용 끼어들기(훅)를 넣는 것 — 중단 모사는 픽스처로 한다.
- 제거 스위치를 운영 경로에서 켜거나 기본값을 바꾸는 것.
- 여러 번 실행하는 기준을 `--force` 없이 같은 버전으로 돌려 건너뛴 실행을 통과로 읽는 것 (design.md D-10, 규칙 7).
- 이전 세대를 최종 이름에 직접 쓰거나 손상 세대를 증거로 읽는 것, 이번 실행이 배포하지 않은 항목에 `deployed_version` 을 채우는 것 (REQ-OWR-015, -017).
- 보존 정책 변경이나 새 삭제로 이전 세대를 정리하는 것 (REQ-OWR-017).
- `git add -A`, 브랜치 상태 변경, 다른 트리 (t1668, 1차 체크아웃) 편집.

## I. 열린 항목

- OQ-1: 프로필 지도 프로젝트에서 "대응 파일 해시 일치가 이전/보존을 가른다"는 리더 문장과 DECIDED 옵션 (c) 사이의 해석. **리더 판정 2026-10-18 로 닫혔다**: 보수적 읽기 — 프로필 루트는 읽지 않고, 프로필 쪽 상태는 `not-checked` 로 기록하며, 프로필 지도 사본은 무조건 보존한다 (decision-index.md Q5).
- OQ-2: Q2 는 리더가 옵션 (가) 로 조건부 결정했다 (보드 d-20261018T040918Z-def2, 2026-10-18; 설계 결정일 뿐 사용자 보안 승인·실제 홈 접근 승인·Kickoff 사전 승인 아님). 운영자 가시 잔여 진술 (감사 N7): (가) 는 지금 현장의 `.agents/skills` 313개 파일의 판정을 **바꾸지 않는다** — 증거는 (가) 가 들어온 뒤 배포되는 파일에 대해서만 생기고, 복사 폴백 거울은 애초에 매니페스트에 등록되지 않는다 (design.md D-5). (가) 가 바꾸는 것: 이후 실행이 배포하는 항목에 배포 버전이 찍히고, 매니페스트 수준 `Version` 이 갱신되며, 실행 백업마다 `manifest.prior.json` 이 하나 생기고, 백업이 실패한 실행은 파일을 제거하지 않는다.
- OQ-3: 복사 폴백 거울을 배포 시점에 기록하는 배포기 변경은 이 SPEC 밖이다 (spec.md § J). 이 변경이 없으면 `.agents/skills` 복사본은 영구히 미등록 보존이다.
- OQ-4: commands 의 "이전 제외 ↔ 재배포" 가 재현되지 않을 경우의 처리 (M6). 같은 버전과 낡은 기록 두 변형을 모두 돌린 뒤에만 "재현 안 됨" 으로 닫는다.
- OQ-5: **리더 판정으로 닫혔다 (2026-10-18)**: 기존 `os.Stat` (`launchCandidateIsUsable`, `profile.go:439`) 는 카드 이전부터 있어 REQ-OWR-005 위반이 아니라 「이 카드가 더하지 않는 기존 접촉」 이다. 예외는 좁게 쓴다: (1) 이 SPEC 은 그 함수를 새로 부르지도 동작을 바꾸지도 않는다, (2) AC-OWR-08 의 스캔 예외는 그 **한 함수의 한 호출** 이름이며 (심볼 + 파일:줄) 파일 전체나 패키지 전체가 아니다, (3) 그 함수나 다른 프로필 접촉 심볼의 **새 호출 지점**은 스캔을 붉게 한다 (같은 함수를 한 번 더 부르는 변이 M-c 가 붉다). 정의는 spec.md § B, 요구사항은 REQ-OWR-005.
- OQ-6 (리더에게): 같은 버전의 일반 재실행은 조정 단계 앞에서 건너뛴다 (`update_template_sync.go:1239-1248`; 이번 실행 측정 § L-13). 그래서 복구 (REQ-OWR-013) 와 빈 디렉터리 정리의 이전 실행 사례는 `--force` 나 템플릿·버전 변화가 있는 실행에서만 닿는다. 이 SPEC 은 관문을 바꾸지 않는다 (범위 밖; 바꾸려면 `versionMatchSkipsSync` 를 다루는 별도 마일스톤·카드가 필요하다). 일반 재실행도 복구해야 한다고 판단하면 그 별개 카드를 연다.

## J. 교차 참조

- spec.md § C (요구사항), § D (항목별 판정표), § F (제약).
- research.md C (판정 네 곳), D (항목별 코드 조사), G (명령).
- design.md D-1 ~ D-7.
- decision-index.md Q1 ~ Q5.
- acceptance.md (판정 명령과 기대 출력).
