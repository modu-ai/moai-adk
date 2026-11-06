# t2053 진행 기록 (run 단계, manager-develop / cycle_type=tdd)

- 카드: t2053 — `moai update` 의 워크트리 마이그레이션이 수 분간 무출력으로 멈춘 것처럼 보이는 문제
- 브랜치: WT-update-progress, 기준 커밋 67b4734d6b
- 범위: 진행 표시 + 시간 제한 + 해시 축소 (운영자 승인 범위, `brief.md` 의 Implementer brief)

## 1. 커밋

| SHA | 제목 |
|---|---|
| 0ba4130416 | perf(worktree): stop hashing ignored files, bound subprocesses, show live progress [t2053] |
| 5a51a89b01 | feat(cli): label silent update/init waits and bound init's git probe [t2053] |
| (이 파일) | docs(t2053): run-phase progress record [t2053] |

작업 트리의 `.moai/config/sections/language.yaml`, `llm.yaml` 변경은 이 카드의 것이 아니며 스테이징·수정하지 않았다.

## 2. 파일별 변경

| 파일 | 변경 |
|---|---|
| `internal/config/defaults.go` | 이름 붙은 기본값 6개 추가: `DefaultWorktreeGitTimeout`(2분), `DefaultWorktreeMoveTimeout`(5분), `DefaultWorktreeLsofTimeout`(30초), `DefaultWorktreeCommandWaitDelay`(2초), `DefaultInitGitProbeTimeout`(5초), `DefaultUserLockNoticeDelay`(1초). 값은 측정값이 아니라 정한 값이다(주석에 명시). |
| `internal/cli/worktree/migrate.go` | (1) 매니페스트: `treeManifest` 에 `IgnoredDigest` 추가. 추적되지 않고 무시되지 않은 파일(`ls-files --others --exclude-standard -z`)은 기존대로 내용을 전부 SHA-256 으로 해시(심볼릭 링크 처리 불변). 무시된 파일(`--others --ignored --exclude-standard -z`)은 경로 + 파일 종류 + 크기를 `Lstat` 만으로 기록하고 열지 않는다. (2) 시간 제한: `boundedOutput`(`exec.CommandContext` + `WaitDelay`), `ErrCommandTimeout`, 패키지 변수 `lsofCommandTimeout` / `gitCommandTimeout` / `gitMoveTimeout`. lsof, 모든 git 읽기 호출, `git worktree move`(정·역방향)에 적용. (3) 실패 안전 처리: `captureManifestChecked` / `captureTreeManifest` 는 git 호출이 시간 초과되면 즉시 중단하고 오류를 돌려준다. `migrateOneTree` 는 이전 매니페스트가 불완전하면 트리를 `skipped`(이유에 "timed out")로 두고 옮기지 않는다. 이동 후 매니페스트가 불완전하거나 다르면 기존 역순서로 되돌린다. `git worktree move` 가 시간 초과된 뒤 목적지가 생겨 있으면 식별자 검사를 거치는 역이동 경로를 탄다. (4) 진행 표시: `runMigrationEngineWith` 가 `treeProgress` 를 받는다(`runMigrationEngine` 은 nil 로 위임). 적용 모드에서만 개수를 알리고 트리마다 한 줄. `migrate --apply`(비 JSON)는 stderr, 업데이트 통합은 `out` 에 쓴다. 드라이런과 `--json` 은 무출력. |
| `internal/cli/worktree/migrate_progress.go` (신규) | `treeProgress`: `tui.ProgressLine` 재사용. TTY 에서는 한 줄을 제자리 갱신, 비 TTY 에서는 트리마다 평문 한 줄(이스케이프 없음). 끝에 `Done` 으로 줄을 닫아 이어지는 최종 행이 새 줄에서 시작하게 한다. 새 스피너 구현은 만들지 않았다. |
| `internal/cli/init_gitdetect.go` | `git remote`, `git remote get-url origin` 을 `gitDetectOutput`(`exec.CommandContext`, `gitDetectTimeout`) 로 실행. `detectGitConfigReport` 가 시간 초과 이유를 함께 돌려주고, `detectGitConfig` 는 기존 두 값 시그니처를 유지한다. 시간 초과 시 manual / github 기본값(기존 실패 경로와 동일한 안전한 값)으로 물러난다. |
| `internal/cli/init.go` | `detectGitConfigReport` 를 호출하고 이유가 있으면 `p.Warn` 으로 보고. |
| `internal/cli/user_asset_phase.go` | `acquireUserLockNotified`: 잠금을 `userLockNoticeDelay`(기본 1초) 안에 못 잡으면 "다른 프로세스가 moai user lock 을 쥐고 있음 + 최대 대기 시간" 을 한 번만 출력. 잠금이 바로 잡히면 무출력. `ensureUserAssetsLocked`(init)와 `runUserAssetUpdatePhase`(update) 모두 적용. |
| `internal/cli/update_progress.go` (신규) | `runUpdateProgressStep`, `pruneBackupsWithProgress` (모두 `tui.ProgressLine`). |
| `internal/cli/update.go` | 바이너리 업데이트 확인(`IsUpdateAvailable` 구간)과 마이그레이션 이전 하우스키핑(일반 흐름과 `--binary` 흐름)에 라벨 추가. |
| `internal/cli/update_template_sync.go` | 오래된 백업 정리(`PruneBackups`)를 `pruneBackupsWithProgress` 로 감쌈. |

## 3. 추가한 테스트 (실패하는 테스트 먼저)

RED 확인: 구현 전 `go vet ./internal/cli/worktree/` 와 `go vet ./internal/cli/` 가 `undefined: lsofCommandTimeout`, `gitCommandTimeout`, `userLockNoticeDelay`, `acquireUserLockNotified`, `gitDetectTimeout` 등으로 실패했다(컴파일 단계 RED).

`internal/cli/worktree/migrate_bounded_test.go`
- `TestCaptureManifest_IgnoredContentIsNotRead` — 무시된 파일의 같은 크기 내용 변경은 매니페스트를 바꾸지 않는다(읽지 않음의 증거이자 알려진 한계의 고정).
- `TestCaptureManifest_IgnoredUnreadableFile_DeletionAndSizeChangeDetected` — mode 000 무시 파일도 삭제·크기 변경은 탐지, 변경 없는 재측정은 동일.
- `TestCaptureManifest_IgnoredTypeChangeDetected` — 일반 파일을 같은 길이의 심볼릭 링크로 바꾸면 탐지.
- `TestCaptureManifest_UntrackedContentChangeStillDetected` — 추적되지 않고 무시되지 않은 파일의 같은 크기 변경은 여전히 탐지.
- `TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof` — 멈춘 lsof 는 `ErrCommandTimeout` 로 반환.
- `TestPlanMigration_SkipsEveryTreeWhenLsofTimesOut` — lsof 가 답하지 못하면 트리는 이동 계획에 오르지 않는다.
- `TestMigrateApply_ManifestGitTimeout_TreeNotMoved` — git `status` 가 멈추면 `moved == 0`, 행은 `skipped` + "timed out", 트리는 원위치.
- `TestMigrateApply_MoveTimeout_TreeStaysAndFails` — `git worktree move` 가 멈추면 `failed` + "timed out", 목적지 없음, 원위치 등록 유지.
- `TestUpdateConsolidation_NonTTYProgress` — 비 TTY 에서 대상 2개일 때 개수 안내 후 `(1/2)`, `(2/2)` 가 최종 행보다 먼저, 이스케이프 없음, 최종 요약 행 불변.
- `TestUpdateConsolidation_DryRunPrintsNoProgress`, `TestMigrateVerb_ApplyShowsLiveLines`, `TestMigrateVerb_ApplyJSONStaysClean`.

`internal/cli/update_init_bounded_test.go`
- `TestAcquireUserLockNotified_PrintsNoticeOncePastThreshold` — 주입한 짧은 임계값을 넘기면 안내가 정확히 한 번, 최대 대기 시간 포함.
- `TestAcquireUserLockNotified_QuietWhenLockIsFree` — 잠금이 비어 있으면 임계값을 한참 넘겨도 무출력.
- `TestDetectGitConfig_GitTimeoutFallsBackToManualWithReason`, `TestDetectGitConfigReport_NoReasonOnPlainAbsence`.
- `TestRunBinaryUpdateStep_ShowsCheckProgress`, `TestRunUpdateProgressStep_LabelsThenCompletes`, `TestPruneBackupsWithProgress_LabelsAndReports`.

위생: 모든 픽스처는 `t.TempDir()` 아래. `TMPDIR` 는 `/var/folders/...`(홈·저장소 밖), `GOTMPDIR` 는 비어 있음. 네트워크 없음. 멈추는 가짜 `git`/`lsof` 는 `exec sleep 30` 으로, 검증 대상 시간 제한(300ms)이 종료시킨다. 배경 부하를 만들지 않았다.

## 4. 검증 (레인 한정, 전체 `internal/cli` 스위트는 실행하지 않음)

출력 파일은 스크래치패드 `/private/tmp/claude-501/-Users-goos--moai-worktrees-projects-moai-adk-go-t2053/4f29bfd8-9bc1-4f88-996a-4e8756921769/scratchpad/` 에 있다(머신 로컬, 인용 대상 아님). 아래는 그 출력에서 옮긴 마지막 줄이다.

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | (출력 없음) |
| `go vet ./internal/cli/... ./internal/config/...` | 0 | (출력 없음) |
| `gofmt -l internal/` | 0 | (출력 0바이트) |
| `go test -count=1 ./internal/cli/worktree/... ./internal/cli/printer/... ./internal/config/...` | 0 | `ok  github.com/modu-ai/moai-adk/internal/cli/worktree  281.567s` / `ok  .../internal/cli/printer  0.418s` / `ok  .../internal/config  12.152s` / `ok  .../internal/config/atomicfile  0.591s` / `ok  .../internal/config/toolpolicy  14.310s` |
| `go test -count=1 -v -run 'TestAcquireUserLockNotified\|TestDetectGitConfig\|TestRunBinaryUpdateStep\|TestRunUpdateProgressStep\|TestPruneBackupsWithProgress' ./internal/cli/` | 0 | `ok  github.com/modu-ai/moai-adk/internal/cli  3.216s` — `=== RUN` 20줄, `--- PASS` 13줄, `--- FAIL` 0줄 |
| `go test -count=1 -v -run 'TestAcquireUserLock\|TestUpdatePhase\|TestRunUpdate\|TestUpdate\|UserAsset\|TestInit.*Git\|TestEnsureUserAssets\|Housekeeping\|TestRunUserAsset' ./internal/cli/` (관련 회귀) | 0 | `ok  github.com/modu-ai/moai-adk/internal/cli  124.055s` — `=== RUN` 201줄, `--- PASS` 138줄 |
| `go test -count=1 -run 'TemplateSync\|Backup\|Prune\|Retention\|CleanReinstall\|DryRun\|Binary' ./internal/cli/` (정리 라벨 회귀) | 0 | `ok  github.com/modu-ai/moai-adk/internal/cli  107.386s` |
| `go test -count=1 ./internal/tui/...` (ProgressLine 재사용 확인) | 0 | `ok  github.com/modu-ai/moai-adk/internal/tui  0.339s` 외 하위 패키지 ok |
| `golangci-lint run ./internal/cli/worktree/... ./internal/config/... ./internal/cli/` | 0 | `0 issues.` |

중간에 린트가 `migrate_bounded_test.go:337:5: QF1001: could apply De Morgan's law` 를 한 건 지적해 조건식을 고친 뒤 위 최종 결과를 얻었다.

## 5. 알려진 한계

1. **같은 크기의 내용 변경 미탐지(핵심 한계)**: 무시된 파일(`node_modules`, 빌드 산출물 등)의 내용이 같은 크기로 바뀌어도 매니페스트가 바뀌지 않는다. 경로 없어짐, 크기 변경, 종류 변경(일반 파일/링크/디렉터리/특수 파일)만 잡는다. `treeManifest` 주석, `hashIgnoredMetadata` 의 `@MX:NOTE`, 테스트로 명시했다. 이전 방식(카드 t1640 리뷰 P2f)은 이것까지 잡았으나 무시된 파일은 재생성 가능한 산출물이라는 전제로 맞바꿨다.
2. **이전 빌드의 저널과 롤백 호환**: 매니페스트 구조가 바뀌어(`IgnoredDigest` 추가, `ContentDigest` 범위 축소) 이전 빌드가 쓴 저널의 `manifest` 값과는 `--rollback` 의 정착된 이동 매니페스트 비교가 일치하지 않는다. 그 경우 롤백은 안전한 쪽(`refused`, "manifest mismatch")으로 멈춘다. 이 변경 이전에 적용된 저널을 롤백할 일이 있으면 별도 판단이 필요하다.
3. **`git worktree move` 시간 초과의 불확정성**: git 이 중간에 종료되면 이동 여부가 불확실하다. 목적지가 존재하면 식별자 검사를 거친 역이동 경로를 타고, 없으면 `failed` 로 둔다. 목적지가 존재하는 부분 이동 상태 자체를 직접 재현하는 테스트는 만들지 않았다(재현이 불안정해질 수 있어서). 분기 로직은 코드 리뷰 대상이다.
4. **애니메이션 없음**: 브리프가 허용한 `tui.ProgressLine` 을 썼다. 트리마다 한 줄을 제자리 갱신하지만 프레임 애니메이션은 없다. 한 트리 내부에서 시간이 오래 걸리는 구간은 이제 시간 제한이 상한을 둔다. `printer.Spinner` 는 stderr 전용이고 update 경로는 `out` 에 쓰는 `ProgressLine` 규약으로 통일되어 있어 쓰지 않았다.
5. 값(2분/5분/30초/5초/1초)은 측정값이 아니라 정한 값이다. 실제 600개 워크트리 환경에서 재측정하지 않았다.
6. 검증 훅: `/verify` 스킬은 `moai gate`(전체 lint+test)를 가리키지만, 레인 규칙(전체 스위트 금지)에 따라 이 세션에서는 돌리지 않았다. 커밋 시점의 pre-commit 훅이 같은 게이트를 독립적으로 다시 실행한다. 두 커밋 모두 훅이 막지 않았다.
7. 리포지토리 전체 테스트 판정은 통합 브랜치의 CI 가 소유하며 이 보고 시점에는 미정(PENDING)이다.

## 6. 범위 밖 (손대지 않음)

- `internal/cli/worktree/sweep_cwd_posix.go` 등 마이그레이션이 아닌 다른 `lsof`/`git` 호출, `guard.go`, `list.go`, `shared.go` 의 `exec.Command` (브리프는 `migrate.go` 와 `init_gitdetect.go` 로 한정).
- `moai bundle add|remove`, `resumeInitializedProject` 의 사용자 자산 잠금 대기(브리프는 update 와 init 한정).
- `update_clean_install.go` 의 클린 재설치 경로 백업 정리(`[clean-reinstall]` 접두 출력 서식이 달라 제외; 일반 템플릿 동기화 경로만 라벨).
- `.moai/reports` 처리: 파일 개수만 세고 내용을 읽지 않으므로 그대로 둠.
- 푸시, PR, 팩토리 큐: 건드리지 않음.

## 7. 리뷰 반영 (codex 리뷰 지적 2건)

두 건 모두 조정자가 도달 가능함을 확인했고, 실패하는 테스트를 먼저 작성해 단언 실패(RED)를 본 뒤 고쳤다.

### F1. 워크트리 목록 조회 실패가 "0건 성공"으로 읽힘

- 원인: `planMigrationFor` 가 `git worktree list` 오류(이제 `gitCommandTimeout` 로 도달 가능)와 파싱 오류를 삼키고 `nil` 을 돌려줘, 적용이 후보를 한 번도 보지 못했는데도 `0 moved, 0 skipped, 0 failed` 를 찍었다.
- 수정: `planMigrationFor` 가 `([]migratePlan, error)` 를 돌려주고(`list the registered worktrees: ...` / `parse the registered worktree list: ...`), `runMigrationEngineWith` 가 그 오류를 그대로 위로 올린다. `migrate` 동사와 업데이트 통합이 모두 오류를 보고하며, 업데이트는 기존 규약대로 이를 경고 행으로 낮춘다(`update.go` 코드는 바꾸지 않았다). 드라이런도 같은 경로라 같은 오류를 보고한다.
- 테스트(RED 확인: 7개 하위 케이스가 "returned success" 단언으로 실패): `TestMigrationEngine_WorktreeListTimeoutIsAnError`(드라이런/적용), `TestMigrateVerb_WorktreeListTimeoutFails`(`--apply`/드라이런), `TestUpdateConsolidation_WorktreeListTimeoutFails`(드라이런/적용). 모두 오류에 "timed out" 이유가 있고 성공 요약 행이 찍히지 않음을 확인한다.

### F2. 역이동이 파일시스템 이동 완료 뒤 오류를 내면 위치를 잘못 가정

- 원인: `m6ReverseMove` 가 `reverseGitMove` 오류를 곧 "트리는 아직 목적지에 있다"로 간주해 re-key / 카드 rebind 를 건너뛰었다. git 이 이름 바꾸기 완료 뒤 시간 초과로 종료되면 트리는 이미 원위치인데 기록은 목적지를 가리킨다.
- 수정: 오류 시 `treeMovedBack` 으로 실제 위치를 식별자로 다시 확인한다. 원위치(source)가 저널의 식별자를 가지고 있고 **그리고** 목적지가 더 이상 그 식별자를 갖지 않을 때만 "되돌아감"으로 보고 `RekeyWorktreeEntryInverse` + `rebindCardForPath` 후 `reversed` 행을 남기고 `true` 를 돌려준다. 목적지가 여전히 식별자를 가지거나, 두 경로 모두 읽을 수 없는 모호한 상태는 기존 pending / failed 경로를 유지한다(실패 안전).
- 테스트(RED 확인: `a reverse move that completed before git reported the timeout was not treated as moved back`): `TestReverseMove_ErrorAfterMoveCompleted_IsTreatedAsMovedBack`(시임이 실제로 이동한 뒤 시간 초과 오류를 반환; 등록 위치, `reversed` 행, 카드가 원위치로 rebind 됨을 확인), `TestReverseMove_ErrorWhileTreeStaysAtDestination_StaysPending`, `TestReverseMove_AmbiguousLocation_StaysPending`.

### 재검증 (레인 한정)

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | 출력 없음 |
| `go vet ./internal/cli/... ./internal/config/...` | 0 | 출력 없음 |
| `gofmt -l internal/` | 0 | 출력 0바이트 |
| `go test -count=1 -v -run 'TestMigrationEngine_WorktreeList\|TestMigrateVerb_WorktreeList\|TestUpdateConsolidation_WorktreeList\|TestReverseMove_\|TestPlanMigration_' ./internal/cli/worktree/` | 0 | `ok .../internal/cli/worktree 20.747s` (`=== RUN` 13줄, 전부 PASS) |
| `go test -count=1 ./internal/cli/worktree/...` (패키지 전체) | 0 | `ok .../internal/cli/worktree 240.660s` |
| `go test -count=1 -v -run 'Worktree.*Migrat\|Migrat.*Worktree\|TestAcquireUserLockNotified\|TestDetectGitConfig\|TestRunBinaryUpdateStep\|TestRunUpdateProgressStep\|TestPruneBackupsWithProgress' ./internal/cli/` | 0 | `ok .../internal/cli 2.856s` (`=== RUN` 20줄, PASS 20, FAIL 0) |
| `golangci-lint run ./internal/cli/worktree/... ./internal/cli/` | 0 | `0 issues.` |

### 2차 리뷰 반영 (G1, G2: 위 한계 두 건)

2차 codex 리뷰가 위 "한계" 두 건을 P2 로 다시 지적했다. 실패하는 테스트를 먼저 작성해 단언 실패를 확인한 뒤 고쳤다.

**G1. 식별자만으로는 "원위치로 돌아감"을 확정할 수 없음**
- 원인: 역이동이 이름 바꾸기 직후 중단되면 원위치 디렉터리는 식별자를 갖지만 git 의 `worktrees/<id>/gitdir` 은 여전히 목적지를 가리킬 수 있다. 이 상태에서 `reversed` 를 찍으면 기록과 git 등록이 어긋난다.
- 수정: `treeMovedBack(root, source, destination, ident)` 가 세 조건을 모두 요구한다. (1) 원위치가 저널 식별자를 가짐(빈 식별자는 거부), (2) 목적지는 더 이상 갖지 않음, (3) 한도가 있는 `git worktree list --porcelain` 에서 원위치(정규 경로 비교)가 등록 경로로 나옴. (3)이 아니면 한도가 있는 `git worktree repair <source>` 를 한 번 실행하고 다시 확인한다. 그래도 등록이 원위치를 가리키지 않으면(repair 가 멈추거나 실패) 기존 pending 경로를 유지한다. 등록이 목적지를 가리키는 동안에는 `reversed` 가 기록되지 않는다.
- 테스트(RED: `registration does not name the source only after the repair`, `reported moved back while git still registers the destination`): `TestReverseMove_StaleRegistration_RepairConverges`(시임이 디렉터리만 옮기고 시간 초과 오류; repair 가 수렴시켜 `reversed`, 등록 = 원위치), `TestReverseMove_StaleRegistration_UnrepairableNeverReversed`(repair 가 멈추면 `false`, 등록은 목적지, `reversed` 행 없음).

**G2. `--rollback` 에도 같은 사후 재확인 적용**
- 원인: `runMigrateRollback` 은 `reverseGitMove` 오류를 곧바로 pending / 실패로 처리해, 이동이 끝난 뒤 시간 초과가 나도 re-key, 카드 rebind, `restored` 기록이 빠졌다.
- 수정: 같은 `treeMovedBack`(G1 의 등록 확인 포함)을 재사용한다. 되돌아간 것이 확인되면 기존 흐름대로 re-key → rebind → `restored` 로 이어지고, 아니면 이전처럼 `pending` 행과 실패 집계를 남긴다.
- 테스트(RED: `rollback failed although the move had completed`): `TestRollback_ReverseErrorAfterMoveCompleted_RecordsRestored`(요약 `restored 1 tree(s)`, 마지막 행 `restored`, 카드가 원위치로 rebind), `TestRollback_StaleRegistrationUnrepairable_StaysPending`(등록이 목적지에 남으면 오류로 끝나고 `restored` 행 없음. 이 케이스는 수정 전에도 통과하는 안전 가드다).

**테스트 안정화**: 전체 패키지 실행 중 `TestMigrateApply_ManifestGitTimeout_TreeNotMoved` 가 부하 때문에 한 번 실패했다(실제 `git worktree list` 가 300ms 한도를 넘김. F1 수정 이후 목록 시간 초과는 오류로 올라오므로 드러난 것). 그 테스트의 한도를 2초로 늘려 `status` 만 한도에 걸리게 했다.

**2차 재검증 (레인 한정)**

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | 출력 없음 |
| `go vet ./internal/cli/... ./internal/config/...` | 0 | 출력 없음 |
| `gofmt -l internal/` | 0 | 출력 0바이트 |
| `go test -count=1 -v -run 'TestReverseMove_\|TestRollback_' ./internal/cli/worktree/` | 0 | `ok .../internal/cli/worktree 17.054s` (신규 4건 + 기존 5건 모두 PASS) |
| `go test -count=1 ./internal/cli/worktree/...` (패키지 전체) | 0 | `ok .../internal/cli/worktree 321.279s` (첫 시도는 위 부하 플레이크로 1건 실패, 한도를 고친 뒤 재실행) |
| `go test -count=1 -v -run 'Worktree.*Migrat\|Migrat.*Worktree\|TestAcquireUserLockNotified\|TestDetectGitConfig\|TestRunBinaryUpdateStep\|TestRunUpdateProgressStep\|TestPruneBackupsWithProgress' ./internal/cli/` | 0 | `ok .../internal/cli 4.313s` (`=== RUN` 20줄, PASS 20, FAIL 0) |
| `golangci-lint run ./internal/cli/worktree/... ./internal/cli/` | 0 | `0 issues.` |

### 한계 (이번 수정 관련)

- (해소) 롤백 경로와 등록 불일치는 위 2차 리뷰 반영(G1, G2)에서 처리했다.
- 남은 한계: `git worktree repair` 가 실패하거나 멈추는 경우는 pending 으로 남고 다음 실행(롤백 재실행 또는 다음 업데이트)이 수렴시키도록 두었다. 이 경로에서 repair 의 성공은 `git worktree list` 재조회로만 판정하며, 등록 외의 git 내부 상태(예: 잠금 파일)는 보지 않는다.

### 3차 리뷰 반영 (H1, H2)

3차 codex 리뷰가 P2 두 건을 지적했다. 두 건 모두 실패하는 테스트를 먼저 작성해 RED 를 확인한 뒤 고쳤다.

**H1. 정방향 이동이 이름 바꾸기 뒤 등록 갱신 전에 끊긴 경우**
- 원인: `git worktree move` 가 디렉터리 이름 바꾸기를 마치고 git 등록을 고치기 전에 오류나 시간 초과로 끝나면, 목적지가 저널 식별자를 갖고 원위치는 비어 있는데 git 은 여전히 원위치를 등록 경로로 가진다. 이때 역이동은 `not a working tree` 로 실패하고, `treeMovedBack` 도(원위치가 없으므로) `--rollback` 도 같은 이유로 수렴하지 못했다.
- 수정: 역이동 직전에 `forwardMoveUnregistered(root, source, destination, ident)` 로 이 상태를 판정한다. 조건은 (1) 식별자가 비어 있지 않음, (2) 원위치가 없음, (3) 목적지가 저널 식별자를 가짐, (4) 한도가 있는 `git worktree list --porcelain` 에서 등록 경로가 원위치이고 목적지가 아님. 이 상태이면 한도가 있는 `git worktree repair <destination>` 를 한 번 실행하고(`repairRegistration`), 등록이 목적지로 바뀐 것을 다시 조회로 확인한 뒤에만 역이동을 진행한다. repair 가 실패하거나 멈추거나 확인되지 않으면 `m6ReverseMove` 는 `pending` 행을 남기고 `false`(트리는 `failed`)를 돌려주며, `--rollback` 은 `pending` 행과 실패 집계를 남긴다. 판정 조건에 맞지 않는 상태(목록 조회 실패 포함)는 기존 경로 그대로 간다.
- 시임: 정방향 이동 호출을 `forwardGitMove` 로 분리했다(동작 변화 없음). 테스트는 이 시임으로 "이름만 바꾸고 시간 초과 오류" 를 재현한다.
- 테스트(RED 출력: `reversed=0 failed=2, want 2 reversed and 0 failed`, `rollback failed on a half-moved tree: ... reverse move failed (exit status 128)`):
  - `TestMigrateApply_ForwardMoveInterruptedAfterRename_ConvergesBack` — 적용 두 트리 모두 `reversed`, 원위치에 있고 원위치로 등록됨, `moved` 0.
  - `TestMigrateApply_ForwardMoveInterrupted_RepairHangs_NeverSuccess` — repair 가 멈추면 `failed` 2, `moved`/`reversed` 0.
  - `TestRollback_ForwardMoveInterrupted_RepairsThenRestores` — 반쯤 옮겨진 트리의 pending 저널이 롤백으로 수렴해 마지막 행 `restored`, 등록은 원위치만.
  - `TestRollback_ForwardMoveInterrupted_RepairHangs_StaysPending` — repair 가 멈추면 롤백이 오류로 끝나고 `restored` 행 없음(이 케이스는 수정 전에도 통과하는 안전 가드).
- 구현 중 겪은 일: 롤백 쪽 배선을 처음 패턴 치환으로 넣었다가 들여쓰기 불일치로 적용되지 않았고, 롤백 테스트 하나가 실패해서 발견했다. 치환 후에는 `grep` 으로 두 호출 지점을 확인했다.

**H2. 이전 빌드가 쓴 저널이 `--rollback` 에서 "manifest mismatch"**
- 원인: 지난 수정이 매니페스트 JSON 에 `IgnoredDigest` 를 추가해 해시 입력이 달라졌다. 이전 빌드가 쓴 저널의 다이제스트는 새 알고리즘과 절대 일치하지 않아, 변하지 않은 트리도 롤백이 거부됐다.
- 수정(최소): 저널 단계에 `manifest_version`(`omitempty`)을 두고 새 저널은 `manifestFormatVersion = 2` 를 기록한다(`settle` 이 매니페스트가 있을 때 설정). 롤백은 `manifestForJournal(tree, version)` 으로 비교한다. 버전이 없는(0) 단계는 `captureLegacyManifest` 로 이전 방식 그대로 계산한다: 필드 구성은 `IgnoredDigest` 를 뺀 이전 JSON 모양, `ContentDigest` 는 `ls-files --others -z` 와 무시 목록의 합집합 전체의 내용 SHA-256(`hashUntrackedContent` 재사용, 정렬과 중복 제거가 같아 바이트가 동일). git 호출이 시간 초과이면 빈 문자열을 돌려주므로 기록된 값과 절대 같지 않다. 일반 버전 관리 틀은 만들지 않았다.
- 검증용 복제본: 테스트 파일 `migrate_legacy_manifest_test.go` 에 기준 커밋 `67b4734d6b` 의 알고리즘을 식별자만 바꿔 그대로 복사했다. 프로덕션 `captureLegacyManifest` 와 구현이 독립이므로, 두 구현이 같은 다이제스트를 낸다는 것 자체가 호환성 증거다.
- 테스트(RED 출력: `rollback of an unchanged tree from a legacy journal failed: ... manifest mismatch`, `a settled step carries manifest_version 0, want >= 2`):
  - `TestRollback_LegacyJournalManifest_UnchangedTreeMatches` — 이전 방식 저널, 변하지 않은 트리(무시 파일 포함)는 `restored`.
  - `TestRollback_LegacyJournalManifest_ChangedTreeRefused` — 무시 파일의 같은 크기 내용 변경, 추적되지 않는 새 파일 추가 둘 다 `manifest mismatch` 로 거부, 마지막 행 `refused`. 이전 빌드가 잡던 변경을 이전 방식 검증이 계속 잡는다는 증거다(이 케이스는 수정 전에도 통과하는 안전 가드).
  - `TestMigrateApply_JournalRecordsManifestVersion` — 새 저널의 정산된 단계는 `manifest_version >= 2`.

**3차 재검증 (레인 한정)**

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | 출력 없음 |
| `go vet ./internal/cli/... ./internal/config/...` | 0 | 출력 없음 |
| `gofmt -l internal/` | 0 | 출력 0바이트 |
| `go test -count=1 -v -run 'TestMigrateApply_ForwardMove\|TestRollback_ForwardMove\|TestRollback_LegacyJournal\|TestMigrateApply_JournalRecords' ./internal/cli/worktree/` | 0 | `ok .../internal/cli/worktree 28.901s` (신규 7건, 하위 2건 PASS) |
| `go test -count=1 ./internal/cli/worktree/... ./internal/cli/printer/... ./internal/config/...` | 0 | `ok .../internal/cli/worktree 270.020s`, printer, config 계열 모두 ok |
| `go test -count=1 -v -run 'Update\|InitGit\|Progress\|UserLock\|Migrat' ./internal/cli/` | 0 | `ok .../internal/cli 256.295s` (`--- PASS` 362, `--- FAIL` 0) |
| `golangci-lint run ./internal/cli/worktree/... ./internal/cli/ ./internal/config/...` | 0 | `0 issues.` |

### 한계 (3차 수정 관련)

- 이전 빌드 저널의 롤백 검증은 이전 방식의 장단점을 그대로 가진다(무시 파일의 내용까지 읽으므로 느릴 수 있으나, 롤백은 드문 복구 경로이고 한도가 있는 git 호출만 쓴다). 새 저널은 무시 파일의 같은 크기 내용 변경을 잡지 못한다는 기존 한계를 유지한다.
- `forwardMoveUnregistered` 는 목록 조회가 실패하면 "해당 없음" 으로 답하고 기존 역이동 경로로 간다. 그 경로가 실패하면 pending 으로 남는다(성공으로 보고되지 않는다).
- `git worktree repair <destination>` 의 성공은 `git worktree list` 재조회로만 판정한다.
- `/verify` 스킬은 실행하지 않았다(전체 스위트를 돌리기 때문).

### 4차 리뷰 반영 (J1)

4차 codex 리뷰의 P2 한 건이다. 실패하는 테스트를 먼저 작성해 RED 를 확인한 뒤 고쳤다.

**J1. 역이동이 이름만 바꾸고 끊긴 뒤 `--rollback` 이 "성공"으로 끝남**
- 원인: 역이동이 디렉터리를 원위치로 돌려놓은 뒤 시간 초과로 끝나고 등록 repair 도 실패하면, 트리는 원위치에 있고 git 등록은 목적지를 가리키며 단계는 pending 으로 남는다. 이 뒤의 `--rollback` 은 `s.NewPath` 가 없다는 이유로 단계를 건너뛰어, 등록, 상태 파일 re-key, 카드 바인딩을 고치지 않은 채 성공으로 끝났다.
- 수정: `runMigrateRollback` 에서 목적지가 없을 때 곧바로 건너뛰지 않는다. 판정 순서는 다음과 같다.
  1. 단계에 식별자가 없거나 원위치도 없으면 기존처럼 건너뛴다(손댈 대상이 없음).
  2. 원위치가 단계 식별자를 갖지 않으면(읽기 실패 포함) `refused` 행을 남기고 실패로 센다.
  3. 원위치가 식별자를 갖고 git 이 이미 원위치로 등록하고 있으면 re-key 와 rebind 의 역연산만 멱등으로 실행한다. 둘 다 이미 수렴한 상태면 아무 행도 쓰지 않는 "할 일 없음"이다. 역연산이 실패하면 실패로 센다.
  4. 등록이 원위치가 아니면 기존 `treeMovedBack`(한도가 있는 `git worktree repair <source>` 한 번 + 재조회)을 재사용한다. 확인되지 않으면(repair 실패, 멈춤, 모호) `pending` 행을 남기고 실패로 세어 롤백이 0 이 아닌 코드로 끝난다.
  5. 확인되면 re-key, rebind 역연산 뒤 `restored` 행을 기록한다.
- 정리: 정방향 롤백과 이 경로가 같은 re-key/rebind 꼬리를 쓰도록 `rekeyAndRebindBack` 으로 뺐다(오류 문구는 기존과 동일).
- 테스트(RED 출력: `rollback summary: "restored 0 tree(s), 0 refused, 0 pending"`, `rollback reported success with the registration unfixed`):
  - `TestRollback_RenamedBackStaleRegistration_RepairsAndRestores` — 요약 `restored 1 tree(s)`, 마지막 행 `restored`, 등록은 원위치만, 카드가 원위치로 rebind.
  - `TestRollback_RenamedBackRepairHangs_ExitsNonZeroNoRestored` — repair 가 멈추면 오류로 끝나고 `restored` 행 없음.
  - `TestRollback_ConvergedStep_IsNothingToDo` — 이미 수렴한 단계는 `restored 0 tree(s)`, `restored` 행 없음, 성공(수정 전에도 통과하는 안전 가드).

**4차 재검증 (레인 한정)**

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | 출력 없음 |
| `go vet ./internal/cli/... ./internal/config/...` | 0 | 출력 없음 |
| `gofmt -l internal/` | 0 | 출력 0바이트 |
| `go test -count=1 -v -run 'TestRollback_\|TestReverseMove_' ./internal/cli/worktree/` | 0 | `ok .../internal/cli/worktree 27.113s` (신규 3건 포함 전부 PASS) |
| `go test -count=1 ./internal/cli/worktree/... ./internal/cli/printer/... ./internal/config/...` | 0 | `ok .../internal/cli/worktree 261.861s`, printer, config 계열 모두 ok |
| `go test -count=1 -v -run 'Update\|InitGit\|Progress\|UserLock\|Migrat' ./internal/cli/` | 0 | `ok .../internal/cli 232.712s` (`--- PASS` 362, `--- FAIL` 0) |
| `golangci-lint run ./internal/cli/worktree/... ./internal/cli/ ./internal/config/...` | 0 | `0 issues.` |

### 한계 (J1 관련)

- 원위치도 목적지도 없는 단계는 여전히 건너뛴다(트리가 어디에도 없으므로 손댈 대상이 없다). 식별자가 빈 단계도 기존대로 건너뛴다. 이 엔진이 쓴 저널에는 항상 식별자가 있으므로 손으로 쓴 저널에만 해당한다.
- 오너의 지시에 따라 이번 수정으로 리뷰 대응을 멈추고, 상태 전수 자체 감사를 `.moai/reports/t2053/self-audit.md` 에 별도로 두었다(커밋 전).

## 8. 통합 수정 (자체 감사 G1–G13 + 독립 감사, 오너 지시 "전부 수정")

수정 커밋은 세 묶음이다. 묶음 A `3451c486b5`(저널·verify·락), 묶음 B `f604a2a47e`(이동·역이동·신원·롤백), 묶음 C `b640917185`(재실행 보고·매니페스트·경로 표기). 규칙은 항목마다 같다. 먼저 상태를 재현하는 테스트를 쓰고, 실패를 확인한 뒤 고친다. 재현되지 않으면 코드를 건드리지 않고 "재현 안 됨"으로 적는다. 이번 라운드에서 재현되지 않은 항목은 없다. C10 은 가설이었으나 상태 파일과 카드 바인딩 양쪽에서 재현되어 고쳤다.

### 항목별 결과

| 항목 | 테스트 | RED 관찰 | 수정 커밋 | 상태 |
|---|---|---|---|---|
| C1 저널 원자 쓰기 | `TestJournalWrite_ReplacesTheFileAtomically`, `TestJournalWrite_InterruptedReplaceKeepsThePreviousJournal` | 예 (같은 inode, "interrupted replace reported success") | `3451c486b5` | 수정 |
| C2 `--verify` 거짓 성공 | `TestVerify_PendingStep_IsAFailure`, `_UnreadableJournal_IsAFailure`, `_MovedButUnregistered_IsAFailure`, `_NoJournalDirectory_NothingToVerify`, `_IntactMove_IsSuccess`, `_RolledBackStep_IsNotAFailure`, `_JSON` | 묶음 A 작업 중 관찰. 세션 압축 후 이 기록을 쓰면서 개별 실패 출력은 옮기지 않았다 | `3451c486b5` | 수정 |
| C3 apply 역이동의 stale 메타데이터 | `TestMigrateApply_ReverseWithFailingInverseMetadata_IsNotReversed` | 예 (`reversed = 2, want 0`, `reversed` 행 기록) | `f604a2a47e` | 수정 |
| C4 정방향 오류 전반의 역이동 | `TestMigrateApply_ForwardMoveNonTimeoutErrorAfterRename_Reverses` | 예 (`reversed=0 failed=2`) | `f604a2a47e` | 수정 |
| C5 비었거나 잘린 gitdir | `TestRollback_EmptiedRegistrationFile_RepairsAndRestores` | 예 (`reverse move failed (exit status 128)`; 이 git 에서 빈 gitdir 이 목록에서 빠짐을 확인 후 진행) | `f604a2a47e` | 수정 |
| C6 상대 gitdir 포인터 | `TestTreeIdent_RelativeGitdirPointer` | 예 (토큰이 `../repo/.git/worktrees/tree/` 를 cwd 기준으로 열려 실패) | `f604a2a47e` | 수정 |
| C7 재실행 감지와 저널 경로 | `TestPlan_InterruptedMigration_NamesTheJournal`, `TestApply_InterruptedMigration_IsNotMovedAgain`, `TestUpdateConsolidation_Interrupted_WarnsAndNamesJournal`, `TestApply_FailedTrees_PrintTheJournalPath` | 예 (재실행이 "exited clean", 출력에 저널 경로 없음) | `b640917185` | 수정 |
| C8 행을 먼저 출력하고 오류 반환 | `TestApply_ReportRowsPrecedeTheError`, `TestUpdateConsolidation_ReportRowsPrecedeTheError` | 예 (verb: 출력이 `Error` 로 시작, update: 행 없음) | `b640917185` | 수정 |
| C9 매니페스트의 git 비정상 종료 | `TestApply_ManifestGitFailure_MovesNothing`, `TestCaptureLegacyManifest_GitFailure_NeverMatches`, `TestCaptureTreeManifest_UnbornHEAD_IsNotAnError` | 앞 둘은 예 (`moved = 2`, 레거시 다이제스트가 비지 않음). 셋째는 수정 전에도 통과하는 안전 가드 | `b640917185` | 수정 |
| C10 경로 표기 차이 | `TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove`, `TestApply_CardBoundViaSymlinkSpelling_StillFollowsTheMove` | 예 (상태 항목 미갱신, 카드 경로가 심볼릭 표기 그대로) | `b640917185` | 재현되어 수정 |
| C11 Windows 락 프로세스 간 | 별도 테스트 없음 | 해당 없음 | `3451c486b5` | 수정 (컴파일 확인만, 아래 한계 참고) |
| C12 저널 파일명 충돌 | `TestMigrateApply_JournalNameCollision_NeverOverwrites` | 예 (묶음 A 작업 중 관찰) | `3451c486b5` | 수정 |
| C13 해제 줄 | `TestMigrateApply_ForwardMoveFailsBeforeRename_ReleasesMovedBy`, `TestRollback_ConvergedStep_WritesRestored` | 예 (`MovedBy still protects…`, 이동 목록 파일 자체가 없음) | `f604a2a47e` | 수정 |
| C14 매니페스트 선행 조건 | `TestRollback_FailedStepWithManifest_RefusesAChangedTree` | 예 (변경된 트리를 되돌림) | `f604a2a47e` | 수정 |
| C15 알 수 없는 `manifest_version` | `TestRollback_UnknownManifestVersion_IsRefusedByName` | 예 (`manifest mismatch` 로 표시) | `f604a2a47e` | 수정 |
| C16 오해를 부르는 사유, `--json`, 조용한 건너뜀, `WaitDelay` | `TestReverseAfterFailure_IdentityRefusal_ReasonSaysRefused`, `TestRollback_JSON`, `TestVerify_JSON`, `TestPlan_TreeAlreadyAtItsDestination_IsNotARow` | 예 (사유가 "reverse move failed", JSON 이 아님, 목적지에 있는 트리가 행으로 나옴) | `f604a2a47e`, `3451c486b5`, `b640917185` | 수정. `WaitDelay` 는 `boundedOutput` 에 이미 설정되어 있어 변경 없음 |
| G8 apply/rollback 신원 규칙 공유 | `TestReverse_UnreadableDestinationIdentity_IsRefused` | 예 (마지막 줄이 `pending`, 기대값 `refused`) | `f604a2a47e` | 수정 |
| G10 빈 신원 단계 | `TestRollback_EmptyIdentStep_IsRefusedWithReason` | 예 (`err=<nil>`) | `f604a2a47e` | 수정 |

### 설계 메모

- **C3/C13.** 역이동 안쪽 순서를 `m6ReverseMoveWhy` 로 분리했다(`m6ReverseMove` 는 같은 시그니처의 래퍼). 트리가 원위치로 돌아온 뒤 상태 파일 역 re-key 나 카드 역 rebind 가 실패하면 `pending` 줄과 함께 실패로 끝난다. 정방향 이동이 이름 변경 전에 실패하면 `reversed` 줄(사유 포함)이 `MovedBy` 보호를 풀어 준다.
- **C5.** `forwardMoveUnregistered` 는 이제 "원위치가 목록에 있어야 한다"는 조건을 요구하지 않는다. 원위치 없음, 목적지 신원 일치, 목록 조회 성공, 목적지가 미등록이면 repair 를 시도한다. 목록 조회 자체가 실패하면 예전처럼 `false` 를 답한다.
- **C7.** 계획 단계에서 prunable 이거나 목적지가 있거나 이미 목적지에 있는 레코드에 대해, 이동 목록의 마지막 위치 줄이 `pending` 이면 `interrupted` 행으로 보고하고 저널을 찾아 `--rollback` 을 안내한다. 자동 정방향 완료는 하지 않는다. 종료 코드는 0 이 아니며 update 는 경고 줄만 출력한다. 실패/되돌림이 있는 실행은 이 실행의 저널 경로를 출력한다(JSON 은 행의 `journal` 필드).
- **C10.** 정규화는 존재하는 가장 긴 상위 경로의 심볼릭 링크를 풀고 나머지를 붙이는 방식이다(`worktreeroot.CanonicalSpelling`). 이동이 끝난 뒤에는 원래 경로가 없으므로 `EvalSymlinks` 를 직접 쓸 수 없다. 카드 바인딩은 `homestate` 에 읽기 전용 `CardWorktreeBindings` 를 더해, 정확 일치가 없을 때 정규화 비교로 행이 실제로 가진 표기를 찾아 그 표기에서 rebind 한다.
- **테스트 변경.** `TestRollback_ConvergedStep_IsNothingToDo` 는 C13 에 맞춰 바뀌었다. 요약은 그대로 `restored 0 tree(s)` 이지만(이동한 것이 없으므로 세지 않는다) 이제 마지막 줄이 `restored` 여야 한다.
- **고아 자식 프로세스 가설.** 시간 초과 시 git 의 자식 프로세스가 남을 수 있다는 가설은 문서로만 남긴다. 프로세스 그룹 코드는 추가하지 않았다. `WaitDelay` 가 파이프 해제를 보장하지만 손자 프로세스까지 죽이지는 않는다.

### 한계 (이번 수정 관련)

- **C11 은 darwin 에서 컴파일만 확인했다.** `GOOS=windows go vet` 와 `GOOS=windows go test -c` 는 통과했으나 `LockFileEx` 의 프로세스 간 배제를 Windows 에서 실행해 보지는 못했다.
- **C14 의 부작용.** `failed` 단계가 매니페스트를 가지고 있고 그 실패 원인이 "이동 뒤 매니페스트 불일치"였다면, 이후 `--rollback` 은 그 트리를 `refused` 로 돌려보낸다. 지시대로 `moved` 와 같은 규칙을 적용한 결과이며, 이 경우 사람이 트리를 확인해야 한다.
- 빈 신원 단계는 이 엔진이 쓴 저널에는 없고 손으로 쓴 저널에만 나타난다. 이제 목적지가 있든 없든 사유와 함께 거부되고 롤백은 0 이 아닌 코드로 끝난다.
- `interrupted` 판정은 이동 목록의 마지막 위치 줄이 `pending` 일 때만 성립한다. `refused` 로 끝난 줄은 이 판정에 들어가지 않는다.
- 전체 `internal/cli` 스위트는 레인 규칙에 따라 실행하지 않았다.

### 재검증 (레인 한정)

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | 출력 없음 |
| `go vet ./internal/cli/... ./internal/config/... ./internal/hook/ ./internal/homestate/ ./internal/worktreeroot/` | 0 | 출력 없음 |
| `GOOS=windows go vet ./internal/cli/worktree/...` | 0 | 출력 없음 |
| `GOOS=windows go test -c -o /dev/null ./internal/cli/worktree/` | 0 | 출력 없음 |
| `gofmt -l internal/` | 0 | 출력 0바이트 |
| `go test -count=1 ./internal/cli/worktree/... ./internal/cli/printer/... ./internal/config/... ./internal/worktreeroot/... ./internal/homestate/...` | 0 | `ok .../internal/cli/worktree 338.970s`, 나머지 ok |
| `go test -count=1 -run 'Worktree\|Rekey\|Registry' ./internal/hook/` | 0 | `ok .../internal/hook 10.143s` |
| `go test -count=1 -v -run 'Update\|InitGit\|Progress\|UserLock\|Migrat' ./internal/cli/` | 0 | `ok .../internal/cli 180.650s` (`--- PASS` 362, `--- FAIL` 0) |
| `golangci-lint run ./internal/cli/worktree/... ./internal/cli/ ./internal/config/... ./internal/hook/ ./internal/homestate/ ./internal/worktreeroot/` | 0 | `0 issues.` |

### 5차 리뷰 반영 (K1, K2)

수정 커밋 `b74bc4a26b`.

**K1. 같은 초에 만든 저널의 순서가 뒤집힘 (C12 에서 유입)**
- 원인: 충돌 저널 `<stamp>-1.json` 은 사전순으로 `<stamp>.json` 보다 앞선다(`-` 가 `.` 보다 작다). `--verify` 는 정렬 목록의 마지막을, `journalRecording` 은 뒤에서부터 찾았으므로 둘 다 더 오래된 저널을 최신으로 골랐다.
- 수정: 저널 디렉터리를 읽는 곳이 두 곳뿐이어서(`runMigrateVerify`, `journalRecording`) 둘 다 공용 `orderedJournals` 를 거친다. 이름을 `<stamp>` 와 숫자 접미사로 나눠 접미사 없음 = 0 으로 정렬한다(`-2` 가 `-10` 앞). 파일명 규칙은 바꾸지 않아 기존 저널이 그대로 읽힌다.
- 테스트(RED 출력: `order = 20260101T000000-10.json,20260101T000000-2.json,20260101T000000.json,...`, `journalRecording = .../20260101T000000.json, want the newer`, `verify judged the older journal`):
  - `TestOrderedJournals_CollisionSuffixSortsAfterTheBareName`
  - `TestJournalRecording_PicksTheSecondJournalOfTheSameSecond`
  - `TestVerify_PicksTheSecondJournalOfTheSameSecond` (오래된 저널은 `pending`, 새 저널은 `moved`; 새 저널이 이겨야 성공)

**K2. 읽을 수 없는 신원을 "없음"으로 취급해 `reversed` 로 보호를 풂 (C6/C13 에서 유입)**
- 원인: 정방향 이동 오류 뒤, 목적지 신원이 일치하지 않으면(읽기 실패 포함) 곧바로 `reversed` 줄을 써서 `MovedBy` 보호를 풀었다. 상대 gitdir 포인터가 풀리지 않는 순간에는 원위치도 없는데 보호가 풀린다.
- 수정: `forwardFailureLocation` 이 위치를 판정한다. 목적지가 있고 신원이 일치하면 역이동 순서, 원위치가 존재하고 단계 신원을 갖고 있으며 목적지가 없거나 읽을 수 있는 다른 신원일 때만 `reversed` 로 보호를 푼다. 그 밖의 모든 경우(목적지 신원을 읽을 수 없음, 경로를 stat 할 수 없음, 어느 쪽도 신원이 없음)는 양 끝을 적은 `pending` 줄을 쓰고 트리를 실패로 센다.
- 같은 패턴 감사 결과: `reversed`/`restored` 줄을 쓰는 곳을 모두 확인했다. `m6ReverseMoveWhy`, 롤백의 목적지 존재 분기, `treeMovedBack`, 롤백의 수렴 분기는 이미 신원 일치를 요구했다. 고친 곳은 둘이다. 롤백이 목적지를 stat 할 수 없을 때 부재로 보고 `restored` 를 쓰던 것, 롤백이 원위치를 stat 할 수 없을 때 "양쪽 다 없음"으로 조용히 건너뛰던 것. 둘 다 `ErrNotExist` 일 때만 부재로 보고, 나머지는 `pending` 으로 유지해 롤백이 0 이 아닌 코드로 끝난다.
- 테스트(RED 출력: `a \`reversed\` line was written for a tree whose location is unknown`, `MovedBy released the protection`, `rollback treated an uninspectable destination as absent`, `rollback skipped a step whose source it could not inspect`):
  - `TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending` — 시임이 실제 git 이동 뒤 목적지의 `.git` 을 풀리지 않는 포인터로 바꾸고 오류를 반환한다. 원위치 없음. 결과는 `pending`, 두 트리 실패, 포인터 복구 뒤 `MovedBy` 가 true.
  - `TestRollback_UnstatableDestination_IsNotTreatedAsAbsent`, `TestRollback_UnstatableSource_IsNotTreatedAsGone` (상위 디렉터리를 `chmod 000`; root 에서는 건너뜀).

**5차 재검증 (레인 한정)**

| 명령 | 종료 코드 | 마지막 줄 |
|---|---|---|
| `go build ./...` | 0 | 출력 없음 |
| `go vet ./internal/cli/... ./internal/config/... ./internal/hook/ ./internal/homestate/ ./internal/worktreeroot/` | 0 | 출력 없음 |
| `GOOS=windows go vet ./internal/cli/worktree/...` | 0 | 출력 없음 |
| `GOOS=windows go test -c -o /dev/null ./internal/cli/worktree/` | 0 | 출력 없음 |
| `gofmt -l internal/` | 0 | 출력 0바이트 |
| `go test -count=1 ./internal/cli/worktree/...` | 0 | `ok .../internal/cli/worktree 343.620s` |
| `go test -count=1 -v -run 'Update\|InitGit\|Progress\|UserLock\|Migrat' ./internal/cli/` | 0 | `ok .../internal/cli 255.731s` (`--- PASS` 362, `--- FAIL` 0) |
| `golangci-lint run ./internal/cli/worktree/... ./internal/cli/ ./internal/config/... ./internal/hook/ ./internal/homestate/ ./internal/worktreeroot/` | 0 | `0 issues.` |

한계: 위치 불명으로 `pending` 이 된 트리는 사람이 `--rollback` 으로 수렴시켜야 한다. 이 라운드에서 `internal/hook`, `internal/homestate`, `internal/worktreeroot` 의 코드는 바뀌지 않아 vet 만 다시 돌렸고 테스트는 4차 정리 때의 결과를 따른다.
