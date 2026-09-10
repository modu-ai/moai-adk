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

| 테스트 | 성질 |
|---|---|
| `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile` | 기지 적색 |
| `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition` | 기지 적색 |
| `TestAuditLagUsesBinlagSeam` | 기지 적색 |
| `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite` | 위 건들의 중첩 스위트 파급 |

**[HARD] 이 4건은 테스트 이름으로만 식별한다 — 에러가 지목하는 파일명은 무작위다.**
`expectedBlobs` **map 순회**의 첫 적중을 출력하므로 같은 실패가 실행마다 다른 파일을 지목한다
(lane-6 실측: 3차 실행 `mcp_server.go`, 격리 재실행 `launcher.go`; 이 트리에서는 `launcher.go`).
따라서 아래 인용에 등장하는 파일명은 **귀속 근거가 아니다** — 초판 이 표가 파일명을 키로 삼았던 것을
2026-09-10 리드 통지를 받아 정정한다.

귀속은 파일명이 아니라 재측정으로 세웠다 — **수리를 되돌린 베이스라인에서 다시 쟀다**:

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

**독립 확인**: lane-6이 다른 경로로 같은 결론에 도달했다 — 지목 파일들이 develop과 자기 HEAD 사이에서 blob 동일함을 보여 develop 자체가 빨갛다고 판정했다. 서로 다른 두 경로(수리 되돌린 베이스라인 재현 / blob 동일성)가 같은 4건을 가리킨다. 리드가 이를 배치 공식 기준선으로 고정했다.

**병합 통과 기준(2026-09-10 리드 정책)**: 실패 집합이 위 4건의 **부분집합**이고 이 카드의 테스트가 통과할 것. 5번째 실패가 나오면 병합하지 않는다.


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
4. **`internal/cli` 선재 실패 4건**(§5.4)은 이 카드 소관이 아니지만 develop에 남아 있다. 리드가 t600으로 발행했다. `TestAuditLagUsesBinlagSeam`은 REQ-ABI-006 계약 위반을 주장하는 성질이 달라 분리 판단이 필요할 수 있다.
5. **에러가 지목하는 파일명이 무작위라는 성질 자체**(§5.4)가 별개의 계기 결함이다. `expectedBlobs` map 순회의 첫 적중을 출력하므로, 실패를 파일명으로 귀속하려는 다음 사람은 실행마다 다른 답을 얻는다. 리드가 **t606으로 발행**했다 — t600에 딸린 축이 아니라 **형제 축**이다. t600을 고쳐 넷이 초록이 되어도 파일명 무작위성은 남아 앞으로의 실패에서 같은 오귀속을 만들기 때문이다. 카드 문안에 대조군이 [HARD]로 박혔다: 같은 입력으로 반복 실행해 지목이 흔들리는 것을 먼저 잴 것 — 한 번 실행으로는 무작위성이 안 보인다.

---

## 8. 병합 트리 재측정 (2026-09-10, 리드 슬롯 승인)

### 8.1 흡수

| 항목 | 값 |
|---|---|
| 흡수 대상 | 로컬 `develop` @ `c8203fbf3` |
| 흡수 결과 | `f51f7b9e6` (충돌 없음) |
| 병합 트리 | `d0782bfc2b40c637d01180a452c6d90ddf1f6d8e` |

### 8.2 재측정 — 파이프 없음, 전체 출력 보존

```
$ go test ./internal/cli/... -count=1 -timeout 60m > .moai/reports/t554/remeasure-merged.txt 2>&1
REMEASURE_EXIT=1
```

(래퍼는 exit 0을 보고했으나 명령의 판정은 위 값이다.)

실패 패키지는 `internal/cli` 하나이고 하위 16개는 모두 `ok`다. panic·timeout·빌드 실패는 없다(882.553s < 60m).

**고유 실패 이름 — 중첩 재출력 포함 전수**:

```
$ grep -oE -- '--- FAIL: [A-Za-z0-9_/]+' .moai/reports/t554/remeasure-merged.txt | sort | uniq -c
   1 --- FAIL: TestAuditLagUsesBinlagSeam
   2 --- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition
   2 --- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile
   1 --- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite
```

×2는 `RunsBoundedFocusedSuite`가 중첩 스위트 출력을 재인쇄한 것이다. **고유 이름은 정확히 기지 4건이며 5번째는 없다** — 리드 통과 기준(부분집합)을 충족한다.

### 8.3 이 카드 테스트의 양성 증거

비-verbose 출력은 통과한 테스트를 인쇄하지 않는다. "FAIL 줄이 없다"는 "돌아서 통과했다"를 입증하지 못하므로 별도로 쟀다:

```
$ go test ./internal/cli/ -run 'TestPreCommit' -v -count=1 -timeout 20m > .moai/reports/t554/remeasure-precommit-v.txt 2>&1
PRECOMMIT_V_EXIT=0
--- PASS: TestPreCommitHook_SubmodulePassesClean (0.27s)
--- PASS: TestPreCommitHook_SubmoduleVetBlocks (0.40s)
--- PASS: TestPreCommitHook_OutsideAnyModuleSkips (0.24s)
--- PASS: TestPreCommitHook_RootModuleStillVets (0.29s)
--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)
```

`TestPreCommit*` PASS 39건, FAIL 0건.

### 8.4 델타 판정 — **이 재측정은 병합 근거로 무효다**

재측정 직후 develop을 재판독하니 움직여 있었다: `c8203fbf3` → `d3b7d438d`.

```
$ git diff --name-only c8203fbf3 d3b7d438d -- internal/cli
internal/cli/todo.go
internal/cli/todo_verb_leak_test.go
internal/cli/update/merge/conflict_blind_breadth_test.go
internal/cli/update/merge/conflict_blind_repro_test.go
```

`internal/cli/todo.go`는 **프로덕션 코드**다(t555, `c9a9e8866`, +151/−14 범위). 테스트 파일은 t555·t576에서 왔다.

리드 규칙("흡수 시점 이후 델타가 `internal/cli` 코드를 건드렸으면 다시 흡수하고 재측정을 다시 하라")에 따라 **§8.2–8.3의 결과는 `f51f7b9e6` 트리에 대한 실측으로만 유효하고, 병합 근거로 쓰지 않는다.** 흡수하면 base 기준 판별식이 낡는다.

~~**다음**: 리드가 슬롯을 다시 주면 그 시점 develop을 재흡수하고 재측정한다.~~ — **§9의 규칙 변경으로 대체됨.**

## 9. 흡수 델타 전제 정정과 규칙 변경 (2026-09-10)

### 9.1 리드 전제 정정 — 무효 원인은 둘이다

리드는 흡수 이후 델타를 "t576(`d3b7d438d`)뿐, `internal/cli` 본체와 별개 패키지"로 보고 재측정이 유효하리라 판단했다. 재판독 결과 그 전제는 두 곳에서 틀렸고, 리드가 정정을 수용했다.

**(1) t555 — 본체 직접 적중.** t555 병합은 흡수 기준 이후에 착지했고 같은 구간에 들어 있다:

```
$ git merge-base --is-ancestor c9a9e8866 c8203fbf3; echo $?
1        # 흡수 기준의 조상이 아님 = 흡수 이후 착지
$ git merge-base --is-ancestor c9a9e8866 d3b7d438d; echo $?
0        # 구간 끝에는 포함
$ git show d3b7d438d:internal/cli/todo.go | grep -m1 '^package '
package cli
```

`internal/cli/todo.go`는 이 카드의 훅 파일과 같은 본체 패키지의 **프로덕션 코드**다.

**(2) t609 — 경로 적중 0, 전이 적중.** develop은 이어서 `d1b61005d`로 움직였다.

```
$ git diff --name-only d3b7d438d d1b61005d
.claude/settings.json
.moai/config/sections/tool-policy.yaml
.moai/reports/t609/verdict.md
internal/template/templates/.claude/settings.json.tmpl
$ git diff --name-only d3b7d438d d1b61005d | grep -c '^internal/cli/'
0
```

`internal/cli` 경로 필터는 빈 출력을 냈으나 `git diff --name-only`는 차이가 있든 없든 exit 0이므로, 필터 없는 전체 목록(4파일)으로 대조해 **양성 부재**로 확인했다. 그러나 경로 적중 0은 영향 없음이 아니다:

```
$ go list -deps ./internal/cli > .moai/reports/t554/cli-deps.txt 2>&1; echo $?
0        # 548개
$ grep -x 'github.com/modu-ai/moai-adk/internal/template' .moai/reports/t554/cli-deps.txt
github.com/modu-ai/moai-adk/internal/template
```

`settings.json.tmpl`은 `//go:embed all:templates`로 `internal/template`의 **컴파일 산출물**에 들어가고, `internal/cli`가 그 패키지를 가져가므로 본체가 싣는 바이트가 바뀐다.

**대조 — t576은 무영향.** `go list -deps`에 `internal/cli/update/merge`도 적중하지만 t576이 들인 것은 `_test.go` 2개뿐이다. 테스트 파일은 다른 패키지가 import하는 패키지에 포함되지 않으므로 `internal/cli`가 가져가는 컴파일 내용은 불변이다. 의존 적중만으로는 판정이 서지 않는다는 것을 이 대조가 보인다.

리드는 "경로 필터 기반 델타 판정은 `//go:embed` 전이를 못 본다"를 이 배치의 **다섯째 계기 결함**으로 채택했다 — 에러 없이 `0`이라는 그럴듯한 값을 돌려주는 형태다.

### 9.2 규칙 변경 — 레인은 영향 범위만, 전체 스위트는 리드가 한 번

병합마다 병합 트리에서 `internal/cli` 전체를 재는 방식은 구조적으로 끝나지 않는 경쟁이었다. 대기열의 어느 병합이든 본체를 직접, 또는 템플릿을 통해 전이로 건드리면 앞선 측정이 무효가 된다. `AGENTS.md` §4는 로컬 검증을 변경 범위로 좁히고 전체 스위트는 CI에 맡기라고 한다.

| 주체 | 새 규칙 |
|---|---|
| 레인 | 병합 트리에서 **이 카드가 영향 줄 수 있는 테스트만** 잰다 |
| 리드 | 일괄 push 직전 develop tip에서 `internal/cli` 전체를 **한 번** 돈다. 실패 ⊆ 기지 4건이면 통과, 5번째가 나오면 병합된 카드 사이에서 이분 탐색. push 뒤 CI가 전 매트릭스를 돈다 |

### 9.3 §8의 위상

§8.2의 882초 전체 스위트(차집합 ∅)와 §8.3의 `-v` 측정은 **흡수 기준 `f51f7b9e6` 트리의 관측**이다. 병합 근거가 아니라 참고로만 남긴다.

### 9.4 이 카드의 남은 절차

창 순번: t612 → t556 → t611 → t549 → **t554**. 슬롯 불요.

1. 창을 받으면 로컬 develop tip을 흡수한다.
2. 병합 트리에서 `go test ./internal/cli/ -run 'TestPreCommit' -v -count=1`을 파일 리다이렉트·파이프 없는 종료 코드로 잰다. 본체 컴파일이 따르므로 다른 레인의 `internal/cli` 컴파일이 없는지 사전 확인한다(프로세스는 패턴 매칭이 아니라 실행 파일 이름으로 거른다).
3. 통과 기준: `TestPreCommit*` FAIL 0, 신규 두 테스트를 이름으로 확인.
4. develop 워크트리에서 `--no-ff` 병합 → `moai integration release` → 로컬 병합 SHA 보고. push 없음.
