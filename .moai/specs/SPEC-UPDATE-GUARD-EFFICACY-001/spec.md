---
id: SPEC-UPDATE-GUARD-EFFICACY-001
title: update 서브시스템 가드 실효성 보강 — 판정 명령의 기계 독립성과 가드 도달 범위
version: 0.1.0
status: draft
created: 2026-08-02
updated: 2026-08-02
author: manager-spec
priority: P2
phase: plan
module: internal/cli
lifecycle: spec-anchored
tags: "cli, update, guard-efficacy, test-seam, falsifiability, cross-platform, coverage"
tier: M
depends_on: [SPEC-UPDATE-DATA-SURVIVAL-001]
---

# SPEC-UPDATE-GUARD-EFFICACY-001 — update 서브시스템 가드 실효성 보강

## HISTORY

| 날짜 | 버전 | 변경 |
|---|---|---|
| 2026-08-02 | 0.1.0 | 최초 작성. SPEC-UPDATE-DATA-SURVIVAL-001이 §E.3에 남긴 이월 부채 4건(D1~D4)을 해소 대상으로 삼는다. |

## §1 배경

SPEC-UPDATE-DATA-SURVIVAL-001은 `moai update` 경로의 데이터 유실을 막는 가드를 여러 개 세웠고, 각 마일스톤의 `#### Gaps` 절에 **가드가 실제로 무엇을 덮고 무엇을 못 덮는지**를 스스로 기록해 두었다. 그 기록에서 네 건이 이월 부채로 남았다. 이 SPEC은 그 네 건만 다룬다.

부채의 공통 성격은 "가드가 없다"가 아니라 **"가드는 있는데 그 가드가 주장하는 만큼 넓지 않거나, 판정 명령이 코드가 아니라 실행 환경에 의존한다"**는 것이다. 따라서 이 SPEC의 산출물은 대부분 테스트와 판정 명령이며, 프로덕션 코드 변경은 최소 두 곳(주입 가능한 stat 이음매, HOME 해석 이음매 통일)에 한정된다.

### §1.1 베이스라인 주의 — 이 SPEC이 겨냥하는 코드는 아직 `origin/main`에 없다

`origin/main`(`a64548a2a`)에는 SPEC-UPDATE-DATA-SURVIVAL-001의 M5·M6가 **아직 없다**. 두 마일스톤은 PR #1275(`feat/SPEC-UPDATE-DATA-SURVIVAL-001`)에 실려 자동 머지를 기다리는 중이다. 이 SPEC이 손대는 `update_safety_test.go`, `update_preserve_partial_test.go`, `update_preserve_inventory.go`의 `restored %d/%d` 보고는 모두 그 PR 위에서만 존재한다.

따라서 **run-phase 진입 전에 PR #1275 머지 후의 `origin/main`으로 리베이스하는 것이 선행 조건**이다. 자세한 절차는 plan.md §C를 따른다.

### §1.2 실측 근거

아래 수치는 plan-phase에서 `.claude/worktrees/e2-data-survival`(PR #1275 브랜치, HEAD `457e1013f`)를 대상으로 직접 실행해 관측했다.

```
$ go test -run 'TestMoaiUpdate_PreservesUserArea' -covermode=set \
    -coverpkg=./internal/cli/...,./internal/cli/update/... \
    -coverprofile=/tmp/uds-d1-base.out -count=1 ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1.506s	coverage: 5.6% of statements
$ go tool cover -func=/tmp/uds-d1-base.out | grep -E '(CleanMoaiManagedPaths|MigrateLegacyMemoryDir|runCleanReinstall|BackupMoaiConfig)'
update/backup/backup.go:27:	BackupMoaiConfig		0.0%
update/deploy/deploy.go:29:	CleanMoaiManagedPaths		66.7%
update/deploy/deploy.go:144:	MigrateLegacyMemoryDir		26.9%
update_clean_install.go:137:	runCleanReinstall		0.0%
```

```
$ grep -n 'userHomeDir()\|userHomeDirFn' internal/cli/*.go | grep -v '_test.go'
internal/cli/glm.go:994:                    homeDir, _ := userHomeDir()
internal/cli/glm_tools.go:123:              var userHomeDirFn = userHomeDir
internal/cli/homedir.go:14:                 func userHomeDir() (string, error) {
internal/cli/update_clean_install.go:410:   homeDir, _ := userHomeDir()
internal/cli/update_template_sync.go:227:   homeDir, _ := userHomeDir()
internal/cli/update_template_sync.go:272:   homeDir, _ := userHomeDir()
internal/cli/update.go:851:                 homeDir, err := userHomeDirFn()
```

### §1.3 선행 SPEC 기술의 정정 — D1의 "0.0%"는 패키지 전체가 아니라 가드 스코프 수치다

선행 기록은 `runCleanReinstall`과 `BackupMoaiConfig`가 "0.0%로 남아 있다"고 적었고 그 수치 자체는 위와 같이 정확하다. 다만 그 `0.0%`는 **`-run 'TestMoaiUpdate_PreservesUserArea'` 선택자 아래에서의 값**이며, 두 함수 모두 패키지 전체 테스트에서는 이미 여러 테스트가 구동한다(`runCleanReinstall` → `update_clean_install_test.go` 외 6개 파일, `BackupMoaiConfig` → `update_restore_test.go`, `update_recovery_manifest_test.go`).

즉 진짜 공백은 "커버리지 0%"가 아니라 **"사용자 소유 디렉터리의 바이트 동일성(byte-identity)을 주장하는 어서션이 이 두 경로를 지나지 않는다"**이다. 실제로 `snapshotDir` 헬퍼는 리포 전체에서 `update_safety_test.go` 한 파일에서만 쓰인다. `update_clean_install_config_preserve_test.go`는 사용자 *설정값*(settings.json 키, user.yaml 이름, .gitignore 패턴)의 생존은 검증하지만, 사용자 소유 *디렉터리 트리*가 바이트 단위로 불변인지는 검증하지 않는다.

따라서 이 SPEC은 REQ를 "커버리지를 0% 위로 올린다"가 아니라 **"사용자 소유 디렉터리 보존 어서션의 도달 범위를 넓힌다"**로 세운다. 커버리지 수치는 도달 여부의 보조 증거로만 쓴다.

## §2 요구사항 (GEARS)

### 부채 D4 — stat 실패 분기가 Windows에서 검증되지 않는다

`mergeBackPreserveInventory`(`internal/cli/update_preserve_inventory.go`)의 stat 실패 분기는 현재 백업 디렉터리에 `chmod 0o000`을 걸어 도달한다. 이 방식은 POSIX 권한 비트 의미론에 의존하므로 Windows에서 `t.Skip`되고 root로 실행할 때도 `t.Skip`된다. 결과적으로 **Windows CI 잡에서 그 분기는 한 번도 실행되지 않는다.** 선행 SPEC은 이음매 추가를 의도적으로 보류했으나, 사용자가 이음매 도입을 승인했다.

- **REQ-UGE-001** — `mergeBackPreserveInventory`는 주입 가능한 stat 함수를 통해 백업 원본의 존재를 조회해야 한다(shall). 기본값은 `os.Stat`이며, 주입은 테스트 전용이다.
- **REQ-UGE-002** — **When** stat 이음매가 `os.ErrNotExist`가 아닌 오류를 반환하면, `mergeBackPreserveInventory`는 멈춘 파일 이름과 이미 복원된 개수를 담은 오류를 반환해야 한다(shall). 이는 기존 동작의 보존이며 변경이 아니다.
- **REQ-UGE-003** — **Where** 테스트가 stat 이음매를 주입하는 경우, stat 실패 분기를 구동하는 서브테스트는 권한 비트에 의존하지 않아야 하며 어떤 플랫폼에서도 `t.Skip`하지 않아야 한다(shall not skip).

### 부채 D3 — HOME 해석 이음매가 update 서브시스템 전체에 균일하지 않다

`ensureGlobalSettingsEnv`만 `userHomeDirFn` 이음매를 쓰고, update 서브시스템의 다른 세 호출부는 `userHomeDir()`를 직접 부른다. 이 때문에 HOME 반경을 고정하려는 어떤 가드도 그 세 경로에는 닿지 못하며, 향후 그 경로를 검증하려는 테스트는 매번 이음매 작업을 먼저 해야 한다.

- **REQ-UGE-004** — update 서브시스템의 HOME 해석 호출부(`update_clean_install.go`의 deploy 컨텍스트 구성부, `update_template_sync.go`의 Validate Templates 단계와 Deploy Templates 단계)는 `userHomeDirFn` 이음매를 경유해야 한다(shall).
- **REQ-UGE-005** — 이 변경은 호출부의 관측 가능한 동작을 바꾸지 않아야 한다(shall not) — `userHomeDirFn`의 기본값이 `userHomeDir`이므로 주입이 없을 때의 결과는 동일하다.

### 부채 D2 — HOME 반경 판정이 운영자 머신 상태에 의존한다

AC-UDS-013의 판정은 실제 `~/.claude/hooks`의 before/after 스냅샷을 `diff`한다. 그 디렉터리가 없는 머신(측정 시점 메인테이너 머신이 그러했다)에서는 양쪽 스냅샷이 모두 0행이라 **삭제 방향에 대해 공허하게 통과**한다. 생성 방향은 여전히 판별하므로 완전 공허는 아니지만, 판정의 판별력이 코드가 아니라 실행 환경에 좌우된다는 점이 결함이다.

- **REQ-UGE-006** — HOME 반경 가드의 판정 절차는 판정 자체가 만든 고정 HOME 픽스처를 대상으로 삼아야 하며(shall), 운영자 머신의 실제 홈 디렉터리 상태에 판별력이 의존하지 않아야 한다(shall not depend).
- **REQ-UGE-007** — 그 고정 HOME 픽스처는 판정 시작 시점에 비어 있지 않아야 한다(shall not be empty) — 삭제 방향의 판별력은 before 스냅샷이 비어 있지 않을 때만 성립한다.
- **REQ-UGE-008** — **When** 프로덕션 코드가 이음매를 우회해 실제 HOME으로 해석하도록 변형되면, 그 판정은 실패해야 한다(shall fail). 이 반증이 관측되지 않은 판정은 REQ-UGE-006을 만족한 것으로 보지 않는다.

### 부채 D1 — 사용자 영역 보존 가드가 레지스트리 함수 하나만 구동한다

`TestMoaiUpdate_PreservesUserArea`는 `deploy.CleanMoaiManagedPaths` 하나를 의도적으로 구동하고, `MigrateLegacyMemoryDir`는 그 꼬리 호출로 조기 반환 분기만 스쳐 지나간다. `runCleanReinstall`(clean-reinstall 오케스트레이션)과 `BackupMoaiConfig`(백업 단계)에는 사용자 소유 디렉터리 보존 어서션이 닿지 않는다(§1.3 정정 참조).

- **REQ-UGE-009** — 사용자 소유 디렉터리(`.moai/harness/`, `.claude/agents/harness/`, `.claude/skills/harness-*`)의 바이트 동일성을 주장하는 가드는 `runCleanReinstall`을 실제로 구동해야 한다(shall).
- **REQ-UGE-010** — 같은 성격의 가드가 `backup.BackupMoaiConfig`를 실제로 구동해야 한다(shall).
- **REQ-UGE-011** — `MigrateLegacyMemoryDir`의 파괴적 분기 두 갈래 — `.moai/state/` 부재 시의 rename 분기, 양쪽 존재 시의 backup-then-remove 분기 — 는 각각 결정적으로 도달되어야 하며(shall), 두 경우 모두 사용자 소유 디렉터리가 불변임을 주장해야 한다(shall).
- **REQ-UGE-012** — **When** `CleanMoaiManagedPaths`의 `.claude/skills/moai*` 글로브가 사용자 소유 스킬을 포함하도록 넓혀지면, 사용자 영역 보존 가드는 실패해야 한다(shall fail). 선행 SPEC은 이 방향을 반증하지 못한 채 글로브 범위를 읽는 것만으로 위험을 주장했다.
- **REQ-UGE-013** — 각 가드는 변경 전 코드 또는 명시된 변형(mutation)에 대해 **실패하는 것이 관측되어야** 한다(shall be observed failing). 통과만 관측된 가드는 실효성이 증명되지 않은 것으로 본다.

### 비기능 요구사항

- **NFR-UGE-001** — 패키지 수준 변수를 재할당하는 테스트(`userHomeDirFn`, 신설 stat 이음매)는 `t.Parallel()`을 호출하지 않아야 한다(shall not). `userHomeDirFn`은 `glm_tools.go:123`에 선언된 패키지 변수이고 `glm_tools_test.go`의 `setupToolsTestHome`이 47개 호출부에서 같은 변수를 재할당한다. 현재 그 파일에는 `t.Parallel()`이 0건이라 경쟁이 없으나, 어느 한쪽이 병렬화되면 즉시 데이터 경쟁이 된다.
- **NFR-UGE-002** — 이 SPEC의 커밋은 `internal/template/templates/` 아래 어떤 파일도 변경하지 않아야 한다(shall not).
- **NFR-UGE-003** — `GOOS=windows GOARCH=amd64 go build ./...`가 성공해야 한다(shall).
- **NFR-UGE-004** — 기존 성공 경로에 회귀가 없어야 한다(shall not regress) — `go test ./...` 전체가 통과해야 한다.
- **NFR-UGE-005** — Go 관례(`snake_case.go` 파일명, `fmt.Errorf("…: %w", err)` 래핑)를 따라야 한다(shall). 이 항목은 프로젝트 툴체인(`gofmt`/`go vet`/`golangci-lint`)이 강제하므로 별도 AC에 매핑하지 않는다.

## §3 범위 밖 (Out of Scope)

### Out of Scope — update 서브시스템 외부의 HOME 호출부

- `internal/cli/glm.go:994`의 `userHomeDir()` 직접 호출. 이 호출부는 GLM 경로 정규화 헬퍼 안에 있으며 update 서브시스템이 아니다. REQ-UGE-004의 대상이 아니고, 이 SPEC에서 손대지 않는다.
- `internal/cli/glm_tools.go:371`의 `userHomeDirFn()` 호출. 이미 이음매를 경유하고 있으며 변경 불필요.

### Out of Scope — 템플릿 배포 자산

- `internal/template/templates/` 아래 모든 파일. 이 SPEC은 로컬 개발 코드와 테스트만 다룬다. 어떤 마일스톤이라도 이 경로를 건드려야 한다면 그 사실을 먼저 보고하고 멈춘다(NFR-UGE-002).

### Out of Scope — 선행 SPEC의 다른 잔여 위험

- `snapshotDir`의 "누락된 루트를 빈 맵으로 취급" 완화책이 픽스처 생성 실패를 숨길 수 있다는 잔여 위험. 픽스처 쓰기가 `t.Fatal`로 보호되므로 경계가 지어져 있고, 이 SPEC의 부채 목록에 포함되지 않았다.
- 관리 대상 제거 어서션이 "타깃 정체성이 아니라 결과만 고정한다"는 잔여 위험(이름 변경·순서 변경 변형은 통과).
- `mergeBackPreserveInventory`의 `continue` 분기(백업에 없는 항목) 커버리지. 어떤 AC도 요구하지 않았고 이번에도 요구하지 않는다.
- 가드가 `t.TempDir()`만 상대하므로 절대 프로젝트 루트에서만 드러나는 경로 해석 버그를 잡지 못한다는 한계.

### Out of Scope — 커버리지 목표치 자체

- 패키지 커버리지 85% 같은 총량 목표. 이 SPEC은 특정 어서션의 **도달 범위**를 다루지 벤치마크 수치를 올리는 작업이 아니다. 함수별 커버리지는 도달 여부의 보조 증거로만 인용한다.

## §4 성공 기준

acceptance.md의 AC 전부가 PASS이며, 각 AC가 명시한 반증(변경 전 코드 또는 변형에 대한 실패)이 실제로 관측되어야 한다. 통과만 있고 반증이 없는 AC는 미완료로 본다.
