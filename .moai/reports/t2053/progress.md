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
