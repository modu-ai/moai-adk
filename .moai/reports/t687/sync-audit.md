# t687 Sync Audit — scripts.lint를 Node lint 축으로

- 대상 커밋: `add08e10c` (branch `WT-scripts-lint`, base `e4ecbf854` = 로컬 develop 선행)
- 감사 시점 트리: 이 워크트리 HEAD `add08e10c`, `git status` clean
- 초판 판정: **FAIL** (F1 — 차단 결함 1건. 카드 스코프 자체의 구현·시험은 성립하나, 이 커밋이 도달 가능한 프로젝트 형상에서 기존 Go lint 축을 조용히 대체·소멸시키는 회귀를 심었다)
- **재판정(`2730c5520` 델타 후): PASS** — §8 참조.

## 1. 재측정 근거 (이 실행, 이 트리)

| 주장 | 근거 (명령 + verbatim 출력) |
|---|---|
| 패키지 테스트 통과 | `go test -count=1 ./internal/hook/quality/` → `ok github.com/modu-ai/moai-adk/internal/hook/quality 20.237s` (첫 시도는 `(cached)`라 `-count=1`로 재실행) |
| Lint 셀렉터 sweep 공백 없음 | `go test -count=1 -run 'Lint' -v ./internal/hook/quality/` → 최상위 `--- PASS|FAIL|SKIP` 23건(레인 보고 5+18=23과 일치), FAIL 0건, 종료 `ok ... 3.348s`. RUN 라인 36 = 최상위 23 + 서브테스트 13 |
| 린트 청결 | `golangci-lint run ./internal/hook/quality/` → `0 issues.` (exit 0) |
| vet / fmt / build | `go vet ./internal/hook/quality/` → 청결. `gofmt -l internal/hook/quality/` → 출력 없음. `go build ./...` → 성공 |
| F1 재현 (아래 §3) | `/tmp` 픽스처 + 가짜 바이너리 A/B, 하단 그대로 인용 |

레인이 보고한 RED(base `e4ecbf854`에서 5 FAIL)는 재현하지 않았다 — 레인 보고로 귀속한다. 내 baseline은 전부 GREEN `add08e10c`이다.

## 2. [HARD] 안전 주장 검증 — 주석이 아니라 코드로 읽음

`resolveNodeLintSteps`가 만든 `{name: "npm run lint", binary: "npm", args: ["run","lint"], optional: true}` 스텝은 테이블 스텝과 완전히 같은 경로(`executeStep` → `runStep`)로 흐른다. gate.go:586 `executeStep(ctx, step, dir, g.config.LintTimeout)` 확인.

- **프로세스 그룹 격리**: `runStep`이 `isolateProcessGroup(cmd)` 호출 (gate.go:1485) + `cmd.Cancel`이 그룹 종료 (1487-1490) + 모든 exit 경로에서 `defer terminateProcessGroup` (1523-1527). 기제 본체 `step_process_group.go`.
- **타임아웃**: `context.WithDeadline(ctx, stepDeadline)`, stepDeadline은 호출자가 넘긴 `g.config.LintTimeout`에서 산출 (1459-1467). 조상 취소와 스텝 예산의 귀속 구분 로직도 그대로 적용.
- **stepEnv git scrub**: `cmd.Env = stepEnv()` (1514) → `step_git_env.go`의 `gitenv.Env()` → `gitenv.Scrub`. GIT_DIR/GIT_INDEX_FILE 계열 제거 (GH #1691 계승).
- **WaitDelay**: `cmd.WaitDelay = stepWaitGrace` (1486) — 파이프 홀딩 후손 대비.

세 가지 모두 상속 확인. 이 부분은 카드 주장대로다.

## 3. 발견 결함

### F1 [High] [blocking] gate.go:472, 578 — resolveNodeLintSteps에 언어 가드가 없어 비(非)Node 툴체인의 lint 축을 대체한다 (검증된 회귀)

`resolveNodeLintSteps`는 도구종별 판별 없이 스텝 슬라이스 전체를 받는다. `detectToolchains`는 루트에서 첫 매치만 취하므로(matchToolchainAt, 테이블 순서상 Go가 Node보다 앞), **go.mod와 package.json(scripts.lint 포함)이 같은 루트에 있으면** 그 루트는 Go 엔트리 하나가 차지하고, 그 Go 엔트리의 `golangci-lint` 스텝이 `npm run lint`로 **교체**된다. 형제 함수 `resolveNodeTestStep`은 `step.name != nodeTestStepName` 가드로 정확히 이 문제를 막고 있는데(1061행) — lint 해석기만 그 패턴이 빠져 있다. 대체 자체는 서머리에 공지되지 않는다(golangci-lint 행이 아예 생성되지 않음).

검증 — 가짜 바이너리 A/B (같은 fakes, 같은 `.golangci.yml`, package.json 유무만 다름):

- 가짜 `golangci-lint`는 exit 1 + `FAKE_GOLANGCI_LINT_INVOKED` 출력, 가짜 `npm`은 exit 0.
- **대조군** (go.mod + .golangci.yml + main.go, package.json 없음):
  ```
  quality gate failed: golangci-lint
    - golangci-lint: executed in 4ms ... — golangci-lint run
  ```
  config-gated Go lint 축이 정상 동작함을 확인.
- **실험군** (같은 트리 + `package.json` `{"scripts":{"lint":"echo own-lint-ran"}}`):
  ```
  quality gate steps (4 configured):
    - go vet: skipped — ...
    - typecheck: skipped — ...
    - npm run lint: executed in 4ms ... — npm run lint
    - go test: skipped — ...
  ```
  게이트 exit 0. `.golangci.yml`이 있는데도 golangci-lint는 한 번도 호출되지 않았다(호출됐다면 exit 1로 게이트가 실패했을 것). **Go 프로젝트의 lint 커버리지가 조용히 사라지고 npm 스크립트가 그 자리를 차지한다.**

도달 가능성: 루트 go.mod + 루트 package.json은 npm workspaces·husky·문서 툴링을 루트에 두는 하이브리드 저장소의 실제 형상이다. 이 커밋 이전에는 그 형상에서 golangci-lint가 돌았다 — 커밋이 만든 회귀다.

**필요 수리**: `resolveNodeLintSteps` 내부에 Node 테이블 판별 가드 1줄 — `resolveNodeTestStep`의 명칭 기반 가드와 대칭으로 `if len(steps) == 0 || steps[0].name != "eslint" { return steps, "", false }` (또는 `steps[0].binary != "npx"`). 가드를 함수 안에 두면 시딩(472)과 실행(578) 두 호출점이 한 번에 덮인다. 회귀 시험은 위 A/B 그대로: go.mod + .golangci.yml + package.json(scripts.lint) + exit-1 가짜 golangci-lint → 게이트는 실패해야 한다(또는 최소한 golangci-lint 행이 executed로 남아야 한다).

### F2 [Low] [blocking 아님] gate_scripts_lint_test.go:14-16 — 파일 머리 주석이 구현과 반대로 말한다

주석: "The replaced entries stay visible in the run summary as skips". 실제 구현(및 `TestScriptsLintSupersedesConfigEntries`의 단언 `recordFor("biome")` == nil)은 정반대 — 단순 대체 경로에서는 대체된 엔트리가 시딩되지 않아 서머리에 **어떤 행도 남지 않고**, 스킵 행은 watch-prone 경로에만 생긴다. gate.go:467-471의 주석은 구현과 일치한다. 테스트 파일 머리만 옛 설계(스킵 행 유지)를 서술한다. 수리: 주석을 구현에 맞춰 2줄 수정. (F1의 "대체가 공지되지 않는다"는 성격 판단과는 별개 — F1 수리 후에도 이 주석은 거짓말 상태로 남는다.)

### F3 [Low] [optional] gate.go:579-584 — watch-prone 스킵 행이 실행 시점에 늦게 생겨 표 순서가 흔들린다

다른 행들은 시딩 순서인데 watch-prone 행만 실행 시점 생성이라 요약 표에서 위치가 뒤로 밀린다. 표시 순서의 미관 문제일 뿐 정보 손실 없음. 수리 여부는 재량.

### F4 [Low] [optional] gate.go:1193 — watch-prone 판별이 테스트 스텝용 술어를 그대로 재사용한다

`nodeScriptWatchProne`은 `--watch`/`--watchAll` 토큰과 bare vitest만 잡는다(REQ-HGT-002). `chokidar`/`nodemon`/`eslint-watch`처럼 토큰 없이 감시하는 lint 스크립트는 판별을 통과해 실행되고 LintTimeout(60s)까지 매달려 false red가 된다. 기존 테스트 스텝이 같은 한계를 이미 감수 중인 **계승된 범위**라 이 카드의 결함으로 보지 않는다 — 별도 후속 카드 감.

### 검토 후 결함 아님으로 판정한 것들

- **이중 해석(시딩·실행) 불일치**: 두 호출점이 같은 `dt.root`/`tcs[i].root`(감지 시 바인딩, 빈 값 불가능)에서 같은 파일을 읽는다. 실행 도중 manifest 수정이라는 병적 상황 외에는 불일치 불가능.
- **t685 다중 툴체인 상호작용**: `detectToolchains`가 언어별 1엔트리를 강제하므로 Run당 Node 해석은 최대 1회. 확인.
- **패키지 테이블 변이 없음**: pass-through는 슬라이스를 그대로 돌려주고 대체 경로는 새 슬라이스 리터럴 — `toolchains` 테이블에 쓰기 없음. `resolveToolchainAt`의 clone(859-861)도 기존 그대로.
- **`optional: true`로 npm 부재 시**: LookPath 실패 → `markSkipped` 가시 행 + 게이트 통과. 모든 optional 스텝과 같은 정책, 일관됨.
- **보안 표면**: 스크립트 텍스트는 판별에만 읽히고 요약에는 `%q`로 인용 — 셸 주입면 신규 없음. `npm run lint`의 신뢰 수준은 기존 `npm test` 해석과 동일.

## 4. 차원 점수 (0-10, 조화평균)

| 차원 | 점수 | 근거 한 줄 |
|---|---|---|
| Functionality (40%) | 5 | 카드 스코프(6시험, RED→GREEN, 불변성 핀)는 성립 — 그러나 검증된 F1 회귀로 도달 가능한 형상에서 기존 검사가 소멸 |
| Security (25%) | 9 | 격리·타임아웃·scrub 전부 상속 확인(§2), 주입면 신규 없음, 신뢰 수준 기존 npm test와 동일 |
| Craft (20%) | 8 | 가짜 바이너리 격리·real-npm e2e·불변성 보존증명까지 갖춘 시험 — F2의 거짓 주석 감점 |
| Consistency (15%) | 6 | 형제 해석기 `resolveNodeTestStep`의 언어 가드 패턴 결여가 F1 그 자체 — 명명·주석·상수 스타일은 기존과 일치 |

조화평균 = 4 / (1/5 + 1/9 + 1/8 + 1/6) ≈ **6.6 / 10 → FAIL** (must-pass인 Functionality 실패가 전체를 강제)

## 5. 미검증 (Gaps)

- `-race` 미실행, 패키지 커버리지 수치 미측정(레인도 수치 미보고), Windows 경로 미실행(시험이 GOOS skip — 문서화됨).
- base `e4ecbf854`에서의 RED 재현 안 함 — 레인 보고 귀속(§1).
- 게이트의 graph-freshness·ast-grep 축은 픽스처에서 advisory/not-configured로만 관측.

## 6. 잔여 위험

- F1 수리 후에도 "대체가 일어났다"는 사실은 서머리에 공지되지 않는다(F2와 표리). 대체 공시 행을 원하면 별도 결정.
- 실행 도중 package.json 수정이라는 병적 상황에서 시딩·실행 해석이 갈라지면 요약 행 불일치 가능 — 실질 위험 미미.
- 스크립트 텍스트는 `%q`로 요약에 노출 — 스크립트에 시크릿을 박아둔 프로젝트에서는 요약 유출면이 생김(기존 npm test 행도 같은 성격, 신규 아님).

## 7. 판정

**FAIL.** 차단 결함 F1 하나: `resolveNodeLintSteps`에 Node 테이블 판별 가드를 추가하고(§3 F1의 1줄 수리 + A/B 회귀 시험 1건), F2의 파일 머리 주석을 구현에 맞춘 뒤 재감사한다. 재감사 범위는 이 결함 delta에 한정한다 — 카드 스코프의 나머지(6시험, [HARD] 안전 상속, 불변성)는 이번 판정에서 확인됐다.

## 8. 재판정 — 델타 `2730c5520` (지적 범위 한정 재감사)

델타: `2730c5520` ("fix(quality): guard scripts.lint resolution to the Node lint axis"), `add08e10c` 위 커밋 1개, 파일 2개 — 범위 이탈 없음.

**(1) 가드**: `resolveNodeLintSteps` 선두에 `if len(steps) == 0 || steps[0].binary != "npx" { return steps, "", false }` — 제안 그대로 함수 안에 두어 시딩·실행 두 호출점 동시 커버, 빈 슬라이스 선방. `npx`는 테이블 전체에서 Node lint 축(gate.go 212/224/234행)에만 존재 — 가드가 다른 언어를 오식별하지 않음을 grep으로 확인. 형제 `resolveNodeTestStep`의 가드와 같은 자리·같은 패턴.

**(2) 재측정 (이 실행, 트리 `2730c5520`)**:

- `go test -count=1 ./internal/hook/quality/` → `ok ... 20.562s`
- `-run 'GoToolchainKeepsGolangciLintBesideScriptsLint' -v` → `=== RUN` 1건 / `--- PASS` 1건 (공백 sweep 없음)
- `golangci-lint run ./internal/hook/quality/` → `0 issues.` / `go vet` 청결 / `gofmt -l` 출력 없음
- **공개 표면 A/B 재실행 (1차 감사의 원 실험 그대로, 재빌드 바이너리)**:
  - 하이브리드(go.mod+package.json+.golangci.yml+main.go, 가짜 golangci-lint exit 1): `EXIT=1`, `quality gate failed: golangci-lint`, `golangci-lint: executed`, `npm run lint` 행 부재 — 1차 감사의 침묵 대체 관측(exit 0 + npm run lint executed)의 정확한 역전. **F1 봉합.**
  - 하이브리드 무 .go 변형: `golangci-lint: skipped — no source file ...(.go)` — Go 축이 자기 규칙(sourceExts 스킵)을 따르는 정상 스킵.
  - 대조군(package.json 없음): `EXIT=1`, `quality gate failed: golangci-lint` — 불변.
  - Node 전용(package.json만): `EXIT=0`, `npm run lint: executed` — 가드가 카드 고유 경로를 막지 않음.

**(3) F2**: 머리 주석이 "Superseded entries are not seeded: for such a project the configured lint axis is exactly the declared command."로 수정 — 구현·시험과 일치.

**(4) 갱신 점수 (0-10)**:

| 차원 | 초판 | 갱신 | 근거 |
|---|---|---|---|
| Functionality (40%) | 5 | 9 | F1 회귀가 단위 시험+공개 표면 양쪽에서 봉합 확인, 카드 AC 전항 충족 |
| Security (25%) | 9 | 9 | 델타는 읽기 전용 판별 로직 — 보안 표면 변화 없음 |
| Craft (20%) | 8 | 9 | F2 주석 해소 + RED 근거가 커밋 메시지에 핀되고 본 감사가 양방향 독립 재현 |
| Consistency (15%) | 6 | 9 | 가드가 형제 해석기 패턴과 동일 자리·동일 형태로 정렬 |

조화평균 = 4 / (4 × 1/9) = **9.0 / 10 → PASS**

**(5) 잔여**: F3(요약 행 순서 미관)·F4(감시형 스크립트 판별 한계 — 계승 범위)는 optional로 후속 카드 권고. 미검증: `-race`, 커버리지 수치, Windows(시험 skip 문서화). base RED는 커밋 메시지 핀에 더해, 본 감사가 1차(add08e10c 결함 재현)와 2차(델타 봉합) 양방향을 독립으로 관측해 보강됐다.
