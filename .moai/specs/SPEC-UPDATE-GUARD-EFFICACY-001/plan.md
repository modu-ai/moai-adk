# SPEC-UPDATE-GUARD-EFFICACY-001 — 구현 계획

## §A 맥락

선행 SPEC-UPDATE-DATA-SURVIVAL-001은 `moai update` 경로의 데이터 유실 가드를 세우면서, 각 마일스톤의 `#### Gaps` 절에 **자기 가드의 한계를 스스로 기록**했다. 그 기록에서 네 건이 이월 부채로 남았고 이 SPEC이 그것만 해소한다.

부채 네 건은 성격이 다르다.

- **D4 (stat 이음매)**는 프로덕션 타입 결정이다. 이음매의 모양(패키지 변수 vs 구조체 필드 vs 함수 파라미터)을 정해야 하고, 그 결정이 이후 테스트 작성 방식을 규정한다. **가장 되돌리기 어렵고, 리뷰에서 바뀔 가능성이 가장 크다.**
- **D3 (HOME 이음매 균일화)**도 프로덕션 변경이지만 이음매는 이미 존재하며(`userHomeDirFn`) 호출부를 그리로 옮기는 기계적 작업이다. 다만 **D2의 가드 범위를 넓혀 주므로 D2보다 먼저 와야 한다.**
- **D2 (판정 기계 독립성)**는 판정 설계 결정이다. 코드보다 AC가 주된 산출물이며, D3가 먼저 끝나면 같은 가드로 네 호출부를 한꺼번에 덮을 수 있다.
- **D1 (보존 가드 확장)**은 테스트 추가로, 넷 중 가장 기계적이고 되돌리기 쉽다. **맨 뒤로 미룬다.**

이 순서는 "바뀔 가능성이 큰 결정을 앞에 두어 사람의 검토가 거기에 집중되게 한다"는 원칙을 따른다.

### §A.1 선행 SPEC 기술의 정정

spec.md §1.3에 기록한 대로, D1의 "`runCleanReinstall`·`BackupMoaiConfig` 0.0%"는 **가드 스코프 한정 수치**이며 두 함수 모두 패키지 전체 테스트에서는 이미 구동된다(plan-phase에서 호출부를 직접 세어 확인). 진짜 공백은 커버리지가 아니라 **사용자 소유 디렉터리 바이트 동일성 어서션의 도달 범위**다. `snapshotDir` 헬퍼가 리포 전체에서 `update_safety_test.go` 한 파일에서만 쓰인다는 것이 그 증거다.

따라서 M4는 "커버리지 올리기"가 아니라 "보존 어서션을 두 경로에 추가하기"이며, AC-UGE-009/010의 `(d)` 항목(어서션 존재 grep)이 커버리지 수치와 별개로 요구되는 이유가 이것이다.

## §B 알려진 이슈

| # | 이슈 | 영향 | 대응 |
|---|---|---|---|
| B1 | `userHomeDirFn`은 패키지 변수이고 `glm_tools_test.go`의 `setupToolsTestHome`이 47개 호출부에서 재할당한다 | 어느 한쪽이 `t.Parallel()`을 부르면 데이터 경쟁 | NFR-UGE-001로 고정, AC-UGE-015가 기계적으로 감시 |
| B2 | 신설 stat 이음매도 패키지 변수라면 같은 위험을 진다 | 동일 | 같은 NFR·AC가 커버 (AC-UGE-015 둘째 grep이 파일 목록을 동적으로 잡음) |
| B3 | AC-UGE-002의 Windows 런타임 통과는 macOS에서 관측 불가 | 판정이 "소스에 스킵 없음 + windows 컴파일 성공"까지만 증명 | CI 잡 결과를 인용하거나 Gaps에 명시 (§E DoD 4항) |
| B4 | AC-UGE-008의 `sed` 삭제 범위는 run-phase가 쓸 실제 주입 코드 형태에 의존 | 반증 명령이 그대로는 안 먹을 수 있음 | 관측 대상(probe 삭제 여부)은 불변이므로 `sed` 표현만 조정 |
| B5 | 이 브랜치는 배타적으로 점유되지 않는다 | 검증과 push 사이에 타인의 머지가 낄 수 있음 | 커밋 직전 `git rev-parse --short HEAD` + `git branch --show-current` 재확인 |

## §C 선행 조건 (run-phase 진입 전 필수)

### C1 — 리베이스가 선행 조건이다

`origin/main`(plan-phase 기준 `a64548a2a`)에는 선행 SPEC의 **M5·M6가 아직 없다**. 두 마일스톤은 PR #1275(`feat/SPEC-UPDATE-DATA-SURVIVAL-001`, HEAD `457e1013f`)에 실려 자동 머지를 기다린다. 이 SPEC이 손대는 파일 중 다음 셋은 **그 PR 위에서만 존재하거나 그 PR이 만든 형태**다.

- `internal/cli/update_safety_test.go` (M5가 전면 재작성)
- `internal/cli/update_preserve_partial_test.go` (M6가 신설)
- `internal/cli/update_preserve_inventory.go`의 `restored %d/%d before failure` 보고 (M6)

**run-phase 진입 절차**:

```bash
# 1) PR #1275 머지 확인
gh pr view 1275 --json state,mergedAt

# 2) 머지 후 origin/main 을 가져와 이 브랜치에 반영
git fetch origin main
git -C <이 워크트리> merge origin/main     # 또는 rebase — 브랜치 정책에 따름

# 3) 세 파일이 실제로 존재하는지 확인 (도달성 확인, 추정 금지)
test -f internal/cli/update_safety_test.go && echo safety-ok
test -f internal/cli/update_preserve_partial_test.go && echo partial-ok
grep -c 'restored %d/%d before failure' internal/cli/update_preserve_inventory.go   # 기대: 3
```

세 확인 중 하나라도 실패하면 **run-phase에 진입하지 않는다.** PR #1275가 아직 안 들어온 것이므로 대기한다.

### C2 — baseline 재측정

acceptance.md §B의 수치는 plan-phase에서 PR 브랜치 워크트리를 대상으로 실측한 값이다. 리베이스 후 트리는 다를 수 있으므로 **§B의 모든 측정 명령을 다시 실행**하고, 값이 달라지면 달라진 값을 baseline으로 삼는다. 옮겨 적은 수치는 baseline이 아니다.

### C3 — 템플릿 경로 확인

어떤 마일스톤이라도 `internal/template/templates/` 아래를 고쳐야 하는 상황이 되면, **그 사실을 먼저 보고하고 멈춘다**(NFR-UGE-002). 이 SPEC은 로컬 개발 코드와 테스트만 다룬다.

## §D 제약

- **Go 관례**: `snake_case.go` 파일명, `fmt.Errorf("…: %w", err)` 래핑. 툴체인이 강제.
- **테스트 격리**: 모든 임시 디렉터리는 `t.TempDir()`. 절대 경로 해석은 `filepath.Abs`, `filepath.Join(cwd, userPath)` 금지.
- **병렬 금지**: 패키지 변수를 재할당하는 테스트는 `t.Parallel()` 호출 금지(NFR-UGE-001).
- **크로스 플랫폼**: `GOOS=windows GOARCH=amd64 go build ./...` 통과 필수.
- **시간 추정 금지**: 마일스톤은 우선순위와 순서로만 기술한다.

## §E 자기 검증

각 마일스톤 종료 시 다음을 관측하고 **실제 출력을 progress.md에 인용**한다.

1. 해당 마일스톤의 AC 전부 (판정 명령 + 관측 출력)
2. A.4 반증이 걸린 AC는 실패 관측 기록
3. `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...`
4. `go test -count=1 ./internal/cli/...`
5. 마지막 마일스톤에서만: `go test ./...` 전체 + `golangci-lint run --timeout=3m` + `go vet ./...`
6. AC-UGE-013 (템플릿 중립성) — 매 마일스톤

읽기 전용 검증은 한 턴에 병렬 Bash로 묶는다.

## §F 마일스톤 (되돌리기 어려운 결정 순)

### M1 — stat 이음매 도입 (REQ-UGE-001, 002, 003)

**왜 먼저인가**: 프로덕션 타입 결정이며, 이음매의 모양이 이후 모든 테스트 작성 방식을 규정한다. 리뷰에서 뒤집힐 가능성이 가장 크다.

**파일**
- `internal/cli/update_preserve_inventory.go` — 패키지 변수 stat 이음매 선언(기본값 `os.Stat`), `mergeBackPreserveInventory`가 그것을 경유
- `internal/cli/update_preserve_partial_test.go` — `stat_failure_*` 서브테스트를 권한 비트 방식에서 이음매 주입 방식으로 교체, `runtime.GOOS`/`os.Geteuid`/`os.Chmod` 가드 제거

**설계 노트**
- 이음매는 `var osStatFn = os.Stat` 형태의 패키지 변수를 기본안으로 한다. 구조체 필드나 파라미터 추가는 `mergeBackPreserveInventory`의 시그니처와 세 호출부를 모두 바꾸므로 비용이 크고, 이 SPEC의 목적(분기 도달성)에 필요하지 않다.
- 패키지 변수를 택하는 대가는 병렬 테스트 금지(NFR-UGE-001)이며, 이 패키지는 이미 `userHomeDirFn`으로 같은 대가를 치르고 있어 새로운 제약이 아니다.
- 기존 세 실패 반환의 문구(`stat backup %s (restored %d/%d before failure)` 등)는 **바꾸지 않는다** — 기존 서브테스트가 그 부분 문자열을 어서트한다.

**AC**: AC-UGE-001, 002, 003, 004

### M2 — HOME 이음매 균일화 (REQ-UGE-004, 005)

**왜 여기인가**: 프로덕션 변경이지만 기계적이다. 그리고 **M3의 가드가 덮을 수 있는 표면을 넓히므로 M3보다 반드시 먼저**여야 한다. M2 없이 M3를 하면 가드는 `ensureGlobalSettingsEnv` 하나만 고정한다.

**파일**
- `internal/cli/update_clean_install.go` — deploy 컨텍스트 구성부의 `userHomeDir()` → `userHomeDirFn()`
- `internal/cli/update_template_sync.go` — Validate Templates 단계와 Deploy Templates 단계, 두 곳

**설계 노트**
- 세 곳 모두 `homeDir, _ := userHomeDir()` 형태로 오류를 버린다. 이음매 교체 시 **오류 처리 방식은 바꾸지 않는다** — 동작 보존이 REQ-UGE-005이고, 오류 처리 변경은 별개 결정이다.
- `internal/cli/glm.go:994`는 범위 밖(spec.md §3). 건드리지 않았음을 AC-UGE-005 (d)가 확인한다.

**AC**: AC-UGE-005, 006

### M3 — 기계 독립적 HOME 반경 판정 (REQ-UGE-006, 007, 008)

**왜 여기인가**: 판정 설계 결정이며, M2가 끝난 뒤라야 네 호출부를 한 가드로 덮을 수 있다.

**파일**
- `internal/cli/update_home_radius_test.go` — sentinel HOME 픽스처 기반으로 재구성, M2가 이음매화한 세 호출부까지 반경 검증 확대
- `acceptance.md` AC-UGE-007 — 판정 절차가 sentinel 트리를 직접 만들고 비어 있지 않음을 확인하는 형태(이미 반영됨)

**설계 노트**
- 판정은 `go test -c`로 테스트 바이너리를 먼저 만들고, **컴파일된 바이너리만** sentinel HOME으로 실행한다. `go test`에 직접 `HOME`을 주면 Go 툴체인의 캐시·모듈 경로까지 옮겨가 느려지거나 네트워크 없이 실패한다.
- 테스트는 `t.Setenv("HOME", …)`를 쓰지 않는다(프로세스 전역 변조 금지, NFR 계열). 격리는 `userHomeDirFn` 주입으로 한다.
- sentinel HOME은 "이음매가 뚫렸을 때 프로덕션이 도달하게 되는 곳"이며, probe 파일의 생사가 반증 신호다(AC-UGE-008).

**AC**: AC-UGE-007, 008

### M4 — 사용자 영역 보존 가드 확장 (REQ-UGE-009 ~ 013)

**왜 마지막인가**: 순수 테스트 추가로 가장 기계적이고 되돌리기 쉽다. 앞 세 마일스톤의 결정에 의존하지도 않는다(독립 실행 가능하나, 검토 집중을 위해 뒤로 뺀다).

**파일**
- `internal/cli/update_safety_test.go` — `snapshotDir`/`writeFixture`/`mapsEqual` 헬퍼 재사용, 신규 가드 3종 추가
  - `TestCleanReinstall_PreservesUserArea` (REQ-UGE-009)
  - `TestBackupMoaiConfig_PreservesUserArea` (REQ-UGE-010)
  - `TestMigrateLegacyMemoryDir_PreservesUserArea` — 서브테스트 `rename_when_state_absent` / `backup_then_remove_when_both_exist` (REQ-UGE-011)

**설계 노트**
- `runCleanReinstall`은 `CleanReinstallOptions`에 `Deployer`/`EmbeddedFS`/`Manifest` 주입 지점이 있고 기존 테스트 7개 파일이 이미 그 방식으로 구동한다. 새 가드도 같은 패턴을 따른다. 다만 v2 fingerprint가 감지되지 않으면 조기 no-op 반환하므로, **픽스처가 v2 신호를 심어야** 실제 파괴 경로를 지난다(`update_clean_install_merge_notice_test.go`가 같은 주의를 이미 기록해 두었다).
- `MigrateLegacyMemoryDir`의 두 분기는 `.moai/state/` 존재 여부로 갈린다. 픽스처에서 그 디렉터리를 만들거나 안 만들어 결정적으로 분기시킨다.
- 사용자 소유 경로는 기존 가드와 동일한 셋을 쓴다: `.moai/harness/`, `.claude/agents/harness/`, `.claude/skills/harness-ios-patterns/`.
- 글로브 확장 반증(AC-UGE-012)은 기존 `TestMoaiUpdate_PreservesUserArea`를 대상으로 하므로 새 테스트가 필요 없다 — 판정 명령만 실행한다.

**AC**: AC-UGE-009, 010, 011, 012, 013, 014, 015

## §G 안티패턴 (이 SPEC에서 금지)

- **AP-1 — `-run` 공허 통과**: `go test -run <pattern>`의 `exit 0`만 인용. 0-매치도 exit 0이다. 존재 grep + `--- PASS:` 행을 함께 관측한다(§A.1).
- **AP-2 — 머신 상태 의존 판정**: 판정의 판별력이 운영자 홈 디렉터리 상태에 좌우되는 것. **이것이 부채 D2 자체**이므로 재생산은 명시적 위반이다.
- **AP-3 — 절대 앵커**: 커밋 SHA 핀, 줄 번호 앵커. base는 `git merge-base`로 계산하고 위치는 내용 토큰으로 지목한다.
- **AP-4 — 도달성을 어서션으로 착각**: 커버리지가 올랐다는 것은 "그 함수를 지났다"까지만 증명한다. "사용자 영역 보존을 주장했다"는 별도로 확인해야 한다(AC-UGE-009/010의 `(d)`).
- **AP-5 — 반증 없는 가드**: 통과만 관측하고 실패는 관측하지 않은 가드. REQ-UGE-013이 금지한다.
- **AP-6 — 공허한 변형 반증**: `sed` 변형이 실제로 파일을 바꿨는지 확인하지 않고 FAIL만 세는 것. AC-UGE-012의 `mutation-applied` 확인이 이것을 막는다.
- **AP-7 — 범위 이탈**: `internal/template/templates/` 수정, `glm.go`의 HOME 호출부 수정, 선행 SPEC의 다른 잔여 위험 처리(spec.md §3).

## §H 상호 참조

- `.moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/progress.md` — 부채 네 건의 원 기록 (§E.3 요약, M4/M5/M6의 `#### Gaps`)
- `.moai/specs/SPEC-UPDATE-DATA-SURVIVAL-001/acceptance.md` — AC-UDS-013(부채 D2의 대상), AC-UDS-014/015(부채 D1의 대상), AC-UDS-016(부채 D4의 대상)
- `internal/cli/CLAUDE.md` — 이 패키지의 서브에이전트 경계·출력 스트림·절대 경로 규약
- `CLAUDE.local.md` §6 (테스트 격리), §14 (하드코딩 방지), §25 (템플릿 내부 콘텐츠 격리)
- `.claude/rules/moai/core/verification-claim-integrity.md` — §A 판정 규율이 구현하는 상위 원칙
