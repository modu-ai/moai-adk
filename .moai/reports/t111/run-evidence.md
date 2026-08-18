# t111 — moai update 무손실 근본 수정 — 실행 증거

Worktree: `.claude/worktrees/t111` · Branch: `WT-t111` · Base: `5c3141372` (origin/release/v3.1.1)

## 1. Claim (주장)

1. **근본 수정 이행 (후보 절충안 — "선백업 후 삭제")**: `CleanMoaiManagedPaths`가 각 관리 뿌리를 삭제하기 전, 임베드 템플릿 FS가 같은 상대경로로 운반하지 **않는** 모든 regular 파일을 run 단위 백업(`.moai-backups/<timestamp>/pre-clean/<root>/...`)으로 복사한 뒤 삭제한다. 백업 실패는 삭제 **이전에** 전체를 중단한다(REQ-UDS-008의 `MigrateLegacyMemoryDir` 패턴 일반화).
2. **템플릿 관리 파일은 백업하지 않는다** — 직후 배포 단계가 같은 경로에 다시 쓰므로 유일 사본이 위태롭지 않고, 백업은 "없어질 파일"로 제한된다(2026-08-15 실측 12건 소멜 — `.moai/config/astgrep-rules`, dev-only rules 등 — 이 클래스 차단).
3. **임베드 FS 로드 실패 시 중단**: 호출부가 `template.EmbeddedTemplates()` 실패를 에러로 돌려 클린업이 "백업 없이 삭제"로 진행되지 않게 한다.
4. **후보 선택 근거**: 후보 1(마커)은 사전 지식 요구·사고 후 무의미, 후보 2(매니페스트만 삭제)는 stale 옛 템플릿 잔존, 후보 3(preserve_paths)은 운영 부담 + 기본 비면 재발 — 세 후보의 결함을 피하면서 이 파일의 기존 패턴(REQ-UDS-008)을 재사용하는 절충을 채택.

## 2. Evidence (증거 — 명령 + 출력)

```
$ go build ./...
BUILD ALL OK

$ go vet ./internal/cli/            # 시그니처 갱신 후
(출력 없음 — 통과)

$ go test ./internal/cli/update/deploy/
ok  github.com/modu-ai/moai-adk/internal/cli/update/deploy  0.860s
  # 신규 5종 포함: BackupsUnmanagedFiles / ConfigTreeReachesBackup /
  # SettingsFileBackedUp / GlobMatchBackedUp / BackupFailureAbortsRemoval
  # 기존 전부(에러 경로 5종·정상 경로·마이그레이션) 통과

$ go test ./internal/cli/ -run "Update|Clean|Deploy"
ok  github.com/modu-ai/moai-adk/internal/cli  17.353s

$ go test ./internal/cli/ -run "Destructive" -v
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.02s)   # 드리프트 가드
--- PASS: TestBackup_OnDiskBeforeFirstDestructiveStep
--- PASS: TestBackupSubsystem_DestructiveSurfaces
--- PASS: TestRecoveryGuard_SilentBeforeDestructiveRegion

$ golangci-lint run ./internal/cli/update/deploy/...
0 issues.
```

**파괴적 사이트 레지스트리**: RemoveAll 3곳이 `CleanMoaiManagedPaths` 본문에서 `backupThenRemove`로 이동 — 레지스트리 행을 옮기고 Protection 문구를 t111 계약으로 갱신. 정적 스캔 가드(TestDestructiveTargetRegistry_CoversAllSites) PASS로 재분류 정확성 기계 검증.

**시그니처 갱신 파급**: `CleanMoaiManagedPaths(root, out)` → `(root, out, tmplFS fs.FS)`. 호출부 1곳(운영) + 테스트 20곳(9개 파일) 갱신. `fs.FS`는 호출부가 주입 — deploy 패키지 leaf 성질 유지(template 미 import), 테스트는 `fstest.MapFS`로 관리/비관리 분할을 입력으로 제어.

## 3. Baseline-attribution (baseline 귀속)

모든 측정은 본 워크트리(`.claude/worktrees/t111`, branch `WT-t111`)에서, base `5c3141372`(= origin/release/v3.1.1 HEAD)에 본 카드 diff를 적용한 트리에 대해 이번 실행으로 관측했다. 주의: 최초 조사를 primary(main 체크아웃)의 `deploy.go`로 읽어 1회 편집 실패 — release 판은 t40 이후 `ManagedCleanTargets` 공유 구조였음. 재독 후 release 판 기준으로 수정했다.

## 4. Gaps (미검증)

- **운영자 요구 1(폴백 큐 adopt) — 리드 판정으로 잔여 없음 (양카드 분담)**: t106 재작업 커밋 `f5297037f`(작성 시점 기준 `WT-t106` 소유, release 미착지 — `git merge-base --is-ancestor` 부정, tip에서 `adoptLocalTodoQueue` 부재 실측)이 `adoptLocalTodoQueue` + 전용 테스트(`TestTodoQueue_FallbackAdoptsExistingLocalQueue`, 2 queued+1 picked+last_seq+spec_id 컷오버 동일)로 폴백 큐 adopt를 담당. 요구 1은 **t106(폴백 큐 adopt) + t111(선백업-후삭제로 update 경로 자료 손실 차단)의 양카드 분담으로 해소** — t111이 이중 구현하지 않는다. t106 착지 시점에 요구 1의 전체가 성립한다.
  - 경로 정정(리드 회신 반영): t106 폴백은 `~/.moai/kanban/` → `~/.moai/todo/`로 재작업 완료(소유 커맨드가 `moai todo`라는 명명 근거).
  - 스테일 정정: 본 카드 작성 시 fetch값 `5c3141372`는 낡았음 — 리뷰 시점 tip은 `5ed668566`(t85+t94 `162f74d99`, t95 `49697cfb4`, t85↔t97 정합 `5ed668566` 착지). 통합은 fetch 후 새 tip 기준.
- **cli 전체 스위트 미실행** — 로컬 부하 규율(레인 타깃 테스트 → CI가 전수 판정). `-run "Update|Clean|Deploy"` 조각만.
- **`moai update` 통합 실행 미검증** — dev 프로젝트에서 update 실행 금지 규율(CLAUDE.local.md §13 정신). 계약(백업 후 삭제·중단 순서)은 deploy 단위 테스트 5종이 커버.

## 5. Residual-risk (잔여 위험)

- **백업은 복사본일 뿐 자동 복원 아님** — 사용자는 `.moai-backups/<timestamp>/pre-clean/`에서 수동 복원. 안내 문구는 progress 출력 1줄(`backed up N unmanaged file(s)`).
- **백업 보존 기간** — `CleanupOldBackups`의 보존 pruning은 pre-clean 백업에도 적용될 수 있음(restore point 수명 정책 의도대로). 즉시 복원 권장.
- **비 regular 파일(심볼릭 링크 등)은 백업에서 제외** — `copyTree`와 동일 규칙(보존 복사가 링크를 따라 백업 트리 밖으로 나가지 않음). 링크 자체가 유실될 수 있음.
- **메모리 보존 경로와의 중복**: `.claude/settings.json` 등 mergeable 집합의 in-memory 보존(REQ-UDS-001/002)은 그대로이고, t111 백업은 같은 파일의 디스크 사본을 추가로 남김 — 중복이지만 무해(파일 1개 단위).
