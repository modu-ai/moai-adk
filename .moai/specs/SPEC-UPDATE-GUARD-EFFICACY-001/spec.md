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

### §1.1 베이스라인 — PR #1275는 머지되었고 이 브랜치에 반영되었다

이 SPEC이 겨냥하는 코드(SPEC-UPDATE-DATA-SURVIVAL-001의 M5·M6)는 PR #1275로 **이미 `origin/main`에 들어왔다**.

```
$ gh pr view 1275 --json state,mergedAt,mergeCommit
{"mergeCommit":{"oid":"83610e03e…"},"mergedAt":"2026-08-01T15:56:27Z","state":"MERGED"}
$ git rev-parse --short origin/main
83610e03e
```

이 SPEC의 plan 브랜치는 `origin/main`을 머지해 `merge-base`가 `83610e03e`로 올라왔고, 타깃 파일 세 개(`update_safety_test.go`, `update_preserve_partial_test.go`, `update_home_radius_test.go`)와 `restored %d/%d` 보고 3건이 모두 실재함을 확인했다.

> **정정 이력**: 이 절의 최초 판은 "PR #1275가 아직 머지되지 않았다"고 적었다. 저작 시점에 이미 거짓이었다(머지 시각이 저작 시각보다 앞선다). plan.md §C1의 리베이스 절차 자체는 유효하며, run-phase는 진입 시점에 그 절차로 상태를 **재확인**한다 — 이 문장을 근거로 확인을 생략하지 않는다.

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

### 부채 D3 — HOME 해석 이음매가 update 서브시스템 전체에 균일하지 않다 (근거 재작성)

`ensureGlobalSettingsEnv`만 `userHomeDirFn` 이음매를 쓰고, update 서브시스템의 다른 세 호출부는 `userHomeDir()`를 직접 부른다.

> **정정 — 삭제 반경 누수가 아니라 테스트 격리 누수다.** 이 절의 최초 판은 "HOME 삭제 반경이 이 세 곳으로 새어 나간다"고 적었다. **거짓이다.** 세 호출부를 직접 읽어 확인한 결과, `homeDir`은 오직 두 곳으로만 흐른다.
>
> ```
> internal/cli/update_clean_install.go:410-411   homeDir, _ := userHomeDir()
>                                                goBinPath := detectGoBinPathForUpdate(homeDir)
> internal/cli/update_template_sync.go:227-228   (동일)
> internal/cli/update_template_sync.go:272-273   (동일)
> ```
>
> 둘 다 **렌더링 입력**이다 — `template.WithHomeDir(homeDir)`와 `template.WithGoBinPath(goBinPath)`. 세 호출부 어디에도 `os.RemoveAll`이 없다. (`update_clean_install.go:322`의 `os.RemoveAll(abs)`는 `projectRoot` 기반 `deprecated` 경로 루프이며 `homeDir`과 무관하다.) 따라서 "삭제 반경"을 근거로 세운 최초 REQ는 성립하지 않았고, plan.md의 "M2 없이는 M3 가드가 한 곳만 고정한다"는 주장도 거짓이었다 — M2가 있든 없든 M3는 한 곳을 고정한다.

**실제 위험은 테스트 격리 누수다.** 테스트가 `userHomeDirFn`을 주입해도 이 세 호출부는 여전히 프로세스 `$HOME`을 읽는다. 그 결과 테스트 안의 템플릿 렌더링이 **운영자의 실제 홈 디렉터리에 좌우된다.** 구체적으로 `detectGoBinPathForUpdate` → `gobin.Detect(homeDir)`는 `GOBIN`/`GOPATH`가 비면 `filepath.Join(homeDir, "go", "bin")`로 떨어지므로, 렌더된 `settings.json`의 `PATH`와 `status_line.sh`의 폴백 경로에 운영자의 실제 홈 절대 경로가 박힌다. 이것은 CLAUDE.local.md §22.4가 기록한 "머신 종속 절대 경로가 렌더 산출물에 굳는다"는 위험과 같은 계열이며, 주입한 픽스처가 아니라 실행 머신이 테스트 결과를 결정한다는 뜻이다.

- **REQ-UGE-004** — update 서브시스템의 HOME 해석 호출부(`update_clean_install.go`의 deploy 컨텍스트 구성부, `update_template_sync.go`의 Validate Templates 단계와 Deploy Templates 단계)는 `userHomeDirFn` 이음매를 경유해야 한다(shall). 목적은 **테스트 격리**다 — 주입된 홈이 렌더링 입력까지 일관되게 도달하게 한다.
- **REQ-UGE-005** — **When** 테스트가 `userHomeDirFn`을 주입하고 이 세 경로 중 하나에 도달하면, 렌더링 결과는 주입된 홈을 따라야 하며(shall) 프로세스 `$HOME`을 따르지 않아야 한다(shall not).

> **의도된 동작 변화임을 명시한다.** REQ-UGE-005는 "동작이 안 바뀐다"가 **아니다**. 이음매를 주입한 테스트에 대해서는 동작이 **바뀐다** — 렌더링이 주입된 홈을 따라간다. 그것이 이 마일스톤의 목적이며 회귀가 아니다. 주입이 없는 프로덕션 실행에서는 `userHomeDirFn`의 기본값이 `userHomeDir`이므로 결과가 동일하다. 따라서 판정은 "전체 스위트가 여전히 green"이 아니라 **주입된 홈을 따라가는지를 직접 관측**하는 것이어야 한다(AC-UGE-006).

### 부채 D2 — HOME 반경 판정이 운영자 머신 상태에 의존한다

AC-UDS-013의 판정은 실제 `~/.claude/hooks`의 before/after 스냅샷을 `diff`한다. 그 디렉터리가 없는 머신(측정 시점 메인테이너 머신이 그러했다)에서는 양쪽 스냅샷이 모두 0행이라 **삭제 방향에 대해 공허하게 통과**한다. 생성 방향은 여전히 판별하므로 완전 공허는 아니지만, 판정의 판별력이 코드가 아니라 실행 환경에 좌우된다는 점이 결함이다.

- **REQ-UGE-006** — HOME 반경 가드의 판정 절차는 판정 자체가 만든 고정 HOME 픽스처를 대상으로 삼아야 하며(shall), 운영자 머신의 실제 홈 디렉터리 상태에 판별력이 의존하지 않아야 한다(shall not depend).
- **REQ-UGE-007** — 그 고정 HOME 픽스처는 판정 시작 시점에 비어 있지 않아야 한다(shall not be empty) — 삭제 방향의 판별력은 before 스냅샷이 비어 있지 않을 때만 성립한다.
- **REQ-UGE-008** — **When** 프로덕션 코드가 이음매를 우회해 실제 HOME으로 해석하도록 변형되면, 그 판정은 실패해야 한다(shall fail). 이 반증이 관측되지 않은 판정은 REQ-UGE-006을 만족한 것으로 보지 않는다.

### 부채 D1 — 사용자 영역 보존 가드가 레지스트리 함수 하나만 구동한다

`TestMoaiUpdate_PreservesUserArea`는 `deploy.CleanMoaiManagedPaths` 하나를 의도적으로 구동하고, `MigrateLegacyMemoryDir`는 그 꼬리 호출로 조기 반환 분기만 스쳐 지나간다. `runCleanReinstall`(clean-reinstall 오케스트레이션)과 `BackupMoaiConfig`(백업 단계)에는 사용자 소유 디렉터리 보존 어서션이 닿지 않는다(§1.3 정정 참조).

- **REQ-UGE-009** — 사용자 소유 디렉터리(`.moai/harness/`, `.claude/agents/harness/`, `.claude/skills/harness-*`)의 바이트 동일성을 주장하는 가드는 `runCleanReinstall`을 실제로 구동해야 한다(shall). 그 가드는 `runCleanReinstall`의 파괴적 표면 — `scanDeprecatedPaths`가 만든 목록에 대한 `os.RemoveAll(abs)` 루프 — 을 지나야 한다.
- **REQ-UGE-010** — 백업 서브시스템의 **파괴적 표면**에 대해 보존 가드가 있어야 한다(shall). 표면은 둘이다.
  - (a) `BackupMoaiConfig`의 실패 롤백 `os.RemoveAll(backupDir)` — 그 실행이 만든 백업 디렉터리만 지워야 하며 형제 백업이나 사용자 영역을 지워서는 안 된다(shall not).
  - (b) `CleanupOldBackups`의 회전 삭제 — **가장 최신 `keepCount`개가 살아남아야** 하며 가장 오래된 초과분만 지워져야 한다(shall).

> **REQ-UGE-010 재정박 근거.** 이 REQ의 최초 판은 "`BackupMoaiConfig`를 구동하는 보존 가드"였다. 코드를 읽어 보니 `BackupMoaiConfig`의 주 경로는 **순수 추가(copy)** 라서, 그 위에 세운 "사용자 영역 불변" 가드는 구조적으로 거의 공허하다 — 아무것도 지우지 않는 함수가 아무것도 지우지 않았음을 확인하는 꼴이다. 실제로 파괴적인 것은 위 두 표면이며, 특히 (b)는 **과거에 실제로 뒤집혔던 버그**다(`backup.go`의 주석: "A prior revision deleted `backups[keepCount:]` — the newest — which destroyed the most recent restore points on every rotation"). 그 역사적 반전이 그대로 반증 변형(canary)이 된다. 부채 D1이 지목한 "백업 단계에 보존 어서션이 없다"는 이렇게 해소해야 의미가 있다.
- **REQ-UGE-011** — `MigrateLegacyMemoryDir`의 파괴적 분기 두 갈래 — `.moai/state/` 부재 시의 rename 분기, 양쪽 존재 시의 backup-then-remove 분기 — 는 각각 결정적으로 도달되어야 하며(shall), 두 경우 모두 사용자 소유 디렉터리가 불변임을 주장해야 한다(shall).
- **REQ-UGE-012** — **When** `CleanMoaiManagedPaths`의 `.claude/skills/moai*` 글로브가 사용자 소유 스킬을 포함하도록 넓혀지면, 사용자 영역 보존 가드는 실패해야 한다(shall fail). 선행 SPEC은 이 방향을 반증하지 못한 채 글로브 범위를 읽는 것만으로 위험을 주장했다.
- **REQ-UGE-013** — 이 SPEC이 **새로 만들거나 고치는 모든 가드**는 변경 전 코드 또는 명시된 변형(mutation)에 대해 **실패하는 것이 관측되어야** 한다(shall be observed failing). 통과만 관측된 가드는 실효성이 증명되지 않은 것으로 본다.

  구속 대상 가드는 **여섯** 개이며, 각각 자기 반증 AC를 가진다. 반증 없는 가드는 하나도 없어야 한다.

  | 가드 | 반증 AC | 변형(canary) |
  |---|---|---|
  | M1 stat 이음매 가드 | AC-UGE-003 | 변경 전 `update_preserve_inventory.go` overlay |
  | M2 렌더링 격리 가드 | AC-UGE-006F | 세 호출부를 `userHomeDir()`로 되돌리는 overlay |
  | M3 HOME 반경 가드 | AC-UGE-008 | 테스트의 이음매 주입을 `userHomeDir`로 치환 |
  | M4 `runCleanReinstall` 가드 | AC-UGE-009F | `scanDeprecatedPaths`에 사용자 경로 추가 |
  | M4 백업 회전 가드 | AC-UGE-010F | 회전 슬라이스를 역사적 버그 형태로 반전 |
  | M4 `MigrateLegacyMemoryDir` 가드 | AC-UGE-011F | backup-then-remove의 백업 단계 제거 |

> **정정 이력**: 최초 판은 REQ-UGE-013을 "각 가드"라고 썼으나 §D 추적 매트릭스는 세 AC(003/008/012)에만 매핑했고, 그중 AC-UGE-012는 **기존** 가드를 겨냥했다. 결과적으로 M4가 새로 만드는 세 가드에 반증이 하나도 없었다 — 자기 §A.4를 스스로 어긴 상태였다. 위 표가 그 공백을 메운다.

### 비기능 요구사항

- **NFR-UGE-001** — 패키지 수준 변수를 재할당하는 테스트(`userHomeDirFn`, 신설 stat 이음매)는 `t.Parallel()`을 호출하지 않아야 한다(shall not). `userHomeDirFn`은 `glm_tools.go:123`에 선언된 패키지 변수이고 `glm_tools_test.go`의 `setupToolsTestHome`이 47개 호출부에서 같은 변수를 재할당한다. 현재 그 파일에는 `t.Parallel()`이 0건이라 경쟁이 없으나, 어느 한쪽이 병렬화되면 즉시 데이터 경쟁이 된다.
- **NFR-UGE-002** — 이 SPEC의 커밋은 `internal/template/templates/` 아래 어떤 파일도 변경하지 않아야 한다(shall not).
- **NFR-UGE-003** — `GOOS=windows GOARCH=amd64 go build ./...`가 성공해야 한다(shall).
- **NFR-UGE-004** — 기존 성공 경로에 회귀가 없어야 한다(shall not regress) — `go test ./...` 전체가 통과해야 한다.
- **NFR-UGE-005** — Go 관례(`snake_case.go` 파일명, `fmt.Errorf("…: %w", err)` 래핑)를 따라야 한다(shall). 이 항목은 프로젝트 툴체인(`gofmt`/`go vet`/`golangci-lint`)이 강제하므로 별도 AC에 매핑하지 않는다.

### 수용된 부채 — 패키지 수준 이음매의 누적

이 SPEC은 `internal/cli`에 **두 번째** 패키지 수준 테스트 이음매(`osStatFn`)를 추가한다. 첫 번째는 기존 `userHomeDirFn`이다. 그 대가는 **영구적 비병렬성**이다 — 두 이음매를 재할당하는 테스트는 영원히 `t.Parallel()`을 쓸 수 없고, 그 제약은 이음매가 존재하는 한 갚아지지 않는다.

AC-UGE-015는 이 비용을 **감시**할 뿐 **상환 계획이 아니다.** 이것을 알려진·수용된 부채로 여기 기록해 둔다.

- **왜 수용하는가**: 대안(구조체 필드, 파라미터 추가)은 `mergeBackPreserveInventory`의 시그니처와 호출부를 모두 바꾸며, 이 SPEC의 목적(분기 도달성)에 필요하지 않다. 이 패키지는 이미 같은 대가를 치르고 있어 *새로운* 종류의 제약이 아니다.
- **상환은 이 SPEC의 범위가 아니다**: 이음매 두 개를 하나의 주입 가능한 구조체로 묶는 리팩터링은 별도 SPEC 소관이다. 세 번째 패키지 수준 이음매가 필요해지는 시점이 그 리팩터링의 진입 조건으로 적절하다.
- **재발견 방지**: 이후 감사에서 "왜 이 패키지는 병렬 테스트를 못 쓰는가"가 다시 제기되면 이 절을 근거로 인용한다.

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
