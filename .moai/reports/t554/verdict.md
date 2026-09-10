# t554 — pre-commit `go vet` 모듈 해상도 (GH #1641 + #1679)

- 카드: t554 (Class B — plan 생략, run → sync)
- 워크트리: `.claude/worktrees/t554`, 브랜치 `WT-vet-module`
- base: 로컬 `develop` @ `d060e0d13`

---

## 1. 묶음 판정 검증 ([HARD] 배차문 요구)

**판정: 두 이슈는 같은 뿌리가 맞다. 쪼갤 필요 없음.**

근거는 추정이 아니라 제보 채널 자체의 기록이다.

| 이슈 | 상태 | 관측 |
|---|---|---|
| #1641 | CLOSED | 제보자(MEMBER)가 2026-09-10 코멘트에서 "#1679 과 같은 결함"이라 명시하고 #1679로 통합해 닫았다 |
| #1679 | OPEN | 같은 증상(`go: cannot find main module`)을 검증 매트릭스 T0–T4B로 확장 기술 |

두 이슈의 근본 원인은 동일하다 — staged 파일의 패키지 경로를 **저장소 최상단 기준**으로 만들고 `go vet`을 최상단 CWD에서 한 번 돌린다.

**인접 관측은 이 카드 소관이 아니다**: #1679 "Adjacent observation 1"(heavy gate 언어 마커 탐지)은 #1680으로 분리됐고 형제 카드 t559(lane-3)가 다룬다. 이 카드는 fast-subset shell 템플릿 축만 만진다.

## 2. 뿌리 수리는 이미 착지해 있었다

`SPEC-PRECOMMIT-VET-MONOREPO-001`(status: completed, 커밋 `45c60e1a3`)이 모듈별 vet(`_moai_module_root` 상향 탐색 + 모듈 루트별 서브셸 `cd`)을 이미 구현했고, 회귀 테스트 2건이 초록이다.

```
$ go test ./internal/cli/ -run 'TestPreCommitHook_Submodule' -v -count=1
--- PASS: TestPreCommitHook_SubmodulePassesClean (0.46s)
--- PASS: TestPreCommitHook_SubmoduleVetBlocks (0.48s)
```

즉 #1679 매트릭스의 **T0/T1/T2/T4A/T4B는 이미 충족**이다.

## 3. 남아 있던 결함 — 매트릭스 T3

**Claim**: 루트에 `go.mod`이 없고 staged `.go` 파일이 **어떤 모듈에도 속하지 않을 때**, 착지분은 여전히 커밋을 전면 차단하며 그 원인을 vet 결함이라고 거짓 보고한다.

**Evidence** (수리 전 constant 기준, probe 테스트):

```
=== RUN   TestProbe_T554_OutsideAnyModule
    PROBE T3 exit=1 stderr=
        [pre-commit] FAILED: go vet reported issues in the staged packages (module root: .).
        [pre-commit] Hint: cd . && go vet ./scripts
        [pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit
--- FAIL: TestProbe_T554_OutsideAnyModule (1.68s)
```

fixture: 루트 `go.mod` 없음 / 모듈은 `submod/` / staged 파일은 `scripts/tool.go`.
#1679 매트릭스 T3의 기대값은 `exit 0` (skip, no false block)이다.

**원인**: `_moai_module_root`는 위로 훑어 `go.mod`을 하나도 못 찾으면 `return 1`로 그 사실을 알린다. 그런데 호출부 두 곳이 그 신호를 버리고 `|| _mr="."`로 루트로 되돌린다. 루트에 모듈이 없으니 `go vet ./scripts`는 **반드시** 실패한다.

**착지 SPEC의 `REQ-PVM-002`를 뒤집는 판단이다.** 그 조항은 루트 폴백을 의도적으로 택하며 "최악이어도 종전 동작이므로 커버리지가 줄지 않는다"를 근거로 삼았다. 그 전제는 **루트에 `go.mod`이 있을 때만 참**이다. 루트에 모듈이 없는 배치에서는 폴백이 커버리지를 만들 수 없고, 거짓 메시지를 단 확정 차단만 만든다 — 위 실측이 그것이다.

## 4. 수리

`|| _mr="."` → `|| continue` (두 호출부, 쌍둥이 두 파일 동일 편집).

- 루트에 `go.mod`이 **있으면** 탐색이 종전대로 `.`을 반환하므로 단일 모듈 배치의 동작·커버리지는 바이트 단위로 불변이다.
- 루트에 `go.mod`이 **없으면** 어떤 모듈로도 vet할 수 없으므로 건너뛰어도 잃는 커버리지가 0이고, 거짓 차단만 사라진다.
- staged 파일이 전부 모듈 밖이면 `MODROOTS`가 비어 `for` 루프가 아예 돌지 않고 exit 0.

편집 파일 (`TestPreCommitTemplateMatchesConstant`가 바이트 동일성을 기계적으로 강제):

- `internal/cli/hook_install_precommit.go` (`preCommitHookContent` 상수)
- `internal/template/templates/.git_hooks/pre-commit` (템플릿 쌍둥이)
- `internal/cli/hook_install_precommit_test.go` (회귀 2건)

## 5. 검증

### 5.1 신규 회귀 + 인접 가드

```
$ go test ./internal/cli/ -run 'TestPreCommitHook_OutsideAnyModuleSkips|TestPreCommitHook_RootModuleStillVets|TestPreCommitTemplateMatchesConstant|TestPreCommitHook_Submodule' -v -count=1
--- PASS: TestPreCommitHook_SubmodulePassesClean (0.46s)
--- PASS: TestPreCommitHook_SubmoduleVetBlocks (0.48s)
--- PASS: TestPreCommitHook_OutsideAnyModuleSkips (0.36s)
--- PASS: TestPreCommitHook_RootModuleStillVets (0.53s)
--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.913s
```

### 5.2 뮤테이션 — 수리를 되돌리면 빨간불인가

공허한 초록을 배제하기 위해 폴백을 복원한 뮤턴트에서 재측정:

```
MUTANT applied (fallback restored)
--- FAIL: TestPreCommitHook_OutsideAnyModuleSkips (0.42s)
    hook_install_precommit_test.go:674: expected exit 0 (file outside any module is skipped), got 1
        [pre-commit] FAILED: go vet reported issues in the staged packages (module root: .).
```

`TestPreCommitHook_RootModuleStillVets`는 뮤턴트에서도 초록이다 — 이는 결함이 아니라 설계다. 그 테스트는 폴백을 겨누는 것이 아니라 "전부 건너뛰는" 공허한 구현을 겨눈다(대조군은 불변을 단언한다).

### 5.3 gofmt / vet

```
$ gofmt -l internal/cli/hook_install_precommit_test.go   # 출력 없음
$ go vet ./internal/cli/                                  # exit 0
```

### 5.4 패키지 스위트

```
$ go test ./internal/cli/... -count=1 -timeout 900s > .moai/reports/t554/gotest-cli.txt 2>&1
exit=1        # 파이프 없이 받은 종료 코드
$ grep -c '^ok' .moai/reports/t554/gotest-cli.txt
16
```

pre-commit 계열 테스트는 전부 초록이다:

```
$ go test ./internal/cli/ -run 'TestPreCommit' -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	26.453s
precommit-exit=0
```

**실패 4건이 있다. 모두 선재 실패이며, 추정이 아니라 재측정으로 귀속했다.**

실패 목록 (`.moai/reports/t554/gotest-cli.txt`):

| 테스트 | 지목한 파일 |
|---|---|
| `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile` | `internal/cli/launcher.go` |
| `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition` | `internal/cli/launcher.go` |
| `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite` | 위 두 건의 중첩 스위트 파급 |
| `TestAuditLagUsesBinlagSeam` | `internal/cli/home_state_coverage.go:243/245/251/253` |

넷 다 이 카드가 건드리지 않은 파일을 지목한다. 그러나 "안 건드렸으니 무관하다"는 도달성 논거일 뿐이므로, **수리를 되돌린 베이스라인에서 다시 쟀다**:

```
$ # 훅 두 파일을 수리 이전 상태로 되돌린 뒤
$ go test ./internal/cli/ -run 'TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition|TestAuditLagUsesBinlagSeam' -count=1
baseline-exit=1
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (5.42s)
    home_state_coverage_test.go:520: audited production file changed after coverage tip: internal/cli/launcher.go
--- FAIL: TestAuditLagUsesBinlagSeam (1.57s)
    mcp_build_identity_test.go:643: sweep: baseline hit home_state_coverage.go:243 MISSING — ...
    mcp_build_identity_test.go:648: sweep: NEW ancestry hit home_state_coverage.go:245 — ...
FAIL	github.com/modu-ai/moai-adk/internal/cli	8.462s
```

메시지가 바이트 단위로 동일하다 — base `develop d060e0d13`가 이미 안고 있던 실패다. 증거: `.moai/reports/t554/baseline-4fail.txt`.


### 5.5 임베드 재생성

`make build` 실행 — `catalog.yaml updated successfully (12899 bytes)` + `go build` 완료. 재생성 후 `git status --short`에 추가 변경 없음(임베드는 `//go:embed all:templates`라 별도 생성 파일이 없다).

## 6. Gaps — 관측하지 않은 것

- **전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다** — 레인 검증 부하 규율(CLAUDE.local.md §4)에 따라 건드린 패키지만 재측정했다. 전 패키지 판정은 CI 몫이다.
- **darwin 외 플랫폼 미측정.** 편집은 POSIX `sh` 문법이고 `continue`는 POSIX 필수 내장이라 이식성 위험은 낮다고 보나, 측정하지는 않았다.
- **실제 모노레포에서의 end-to-end `git commit`(#1679 T1b)은 재현하지 않았다.** 근거는 fixture 기반 훅 실행이다.
- **`#1679`가 함께 제안한 "vet 출력 verbatim 노출"(현재 `>/dev/null 2>&1`로 삼킴)은 손대지 않았다.** 모듈 해상도와 다른 축이고 카드 범위 밖이라 후속으로 남긴다 — 아래 §7.

## 7. Residual risk / 후속 제안

1. **삼켜진 vet 진단** — 실패 시 `go vet`의 실제 출력이 `/dev/null`로 사라져 사용자는 이유를 볼 수 없다. #1679가 명시적으로 지적한 항목이며 이번 수리로 거짓 메시지의 빈도는 줄었지만 원인 자체는 남아 있다. 별도 카드 권장.
2. **모듈 루트 경로에 공백이 있으면** `for _mr in $MODROOTS`의 비인용 단어 분리가 깨진다. 착지분에서 넘어온 선재 성질이며 이번 편집이 만들지 않았다.
3. **`.moai/config/build-tags`가 존재하되 유효 줄이 없으면** `set -e` 아래에서 `[ -n "$_bt_line" ] && BT_TAGS=...`가 상태 1로 훅을 중단시킬 수 있다. 역시 선재 성질이며 미측정 — 확인 필요하면 별도 카드.
4. **`internal/cli` 선재 실패 4건**(§5.4)은 이 카드 소관이 아니지만 develop에 남아 있다. `home_state_coverage.go`의 ancestry 비교가 `binlag.Evaluate` 밖으로 나갔다는 `TestAuditLagUsesBinlagSeam`의 지적(REQ-ABI-006)은 별도 카드감으로 보인다 — 리드 판정 요청.
