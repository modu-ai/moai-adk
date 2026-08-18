# t76 — TestPreTool_AstGrepSkipReasonSurfaces 선재 결함 조사·수정

- 카드: t76 (칸반 배치 tjv7iy, run 레인)
- 워크트리: `.claude/worktrees/t76` (branch `WT-t76`, base = `release/v3.1.1` @ `36a12cf82`)
- 대상: `internal/hook/pre_tool_astgrep_reason_test.go` — CI 초록 vs 로컬 결정론적 실패 (`:78 SystemMessage is empty`)

## 1. 주장 (Claim)

실패의 원인은 three-frame chain(`RunAstGrepGateV2` → `QualityGate.Run` → `preToolHandler.Handle`) 내부의 전파 결함이 아니라, **테스트가 `MOAI_AUTONOMY_TIER` 환경변수에 대해 밀폐되지 않은 것**이다. 칸반 런처가 컴패니언 세션에 `MOAI_AUTONOMY_TIER=fully-autonomous`를 export하며, 이 티어에서 `pre_tool.go`의 커밋 게이트 블록이 통째로 스킵되어 `gateNotice`(→ `SystemMessage`)가 애초에 설정되지 않는다. CI에는 해당 변수가 없어 `AutonomyTier()`가 semi-auto로 fail-safe → 게이트 실행 → 통과.

## 2. 증거 (Evidence) — 카드에 명시된 3단계 조사 순서대로

### 단계 ① — CI 실행/스킵 여부: **실행됨 (스킵 아님)**

`.github/workflows/ci.yml` 필수 `test` 잡(ubuntu)이 전체 스위트를 돌린다:

```
- name: Run tests with coverage (fast — no race detector)
  run: go test -coverprofile=coverage.out -covermode=atomic ./...
```

`test-race` 잡만 ast-grep을 설치한다("ast-grep is needed because internal/cli rewrite-guard tests shell out to `sg`" 주석). 필수 `test` 잡에는 `sg`가 없어 테스트의 `t.Setenv("PATH", t.TempDir())` 전제와 동일 환경이다. 테스트 파일은 PR #1183(`3e6c92ef7`, SPEC-FALSE-ALLCLEAR-GUARD-001)에서 들어온 성숙한 테스트로, main에서 CI 초록이 실측 기반이다.

### 단계 ② — ast-grep 바이너리·룰셋 분기 개입 여부: **개입 없음 (가설 반박)**

"homebrew 등 폴백 경로가 PATH 밖의 sg를 찾아낸다"는 가설을 검토했으나 반박됐다:

- `internal/astgrep/scanner.go:231-238` — `isSGAvailable()`는 순수 `exec.LookPath(binary)`.
- `scanner.go:129-146`의 `trustedBinaryPrefixes()`(`/usr/bin/`, `/opt/homebrew/bin/`, `$HOME/go/bin/` …)는 **사용자 지정 절대경로의 보안 검증 전용**(`ValidateBinary`)이며 바이너리 탐색 폴백이 아니다.
- 룰셋 분기도 무관 — 이 결함 경로에선 스캔 자체에 도달하지 않는다.
- 이 기계에 `/opt/homebrew/bin/sg`(ast-grep 0.40.5)가 실재하지만, PATH가 비어 있으면 `exec.LookPath`는 이를 찾지 못한다(통과 실행 로그의 `WARN ast-grep (sg) CLI not found` 2회 — 핸들러 체인과 재계산 양쪽에서 unavailable 분기 도달 확인).

### 단계 ③ — SystemMessage 드랍 지점: **three-frame 안이 아니라 진입 전 (티어 분기)**

`internal/hook/pre_tool.go:454`:

```go
if quality.IsGitCommit(command) && !config.IsAutonomyTierCommitGateOff(config.AutonomyTier()) {
    gate := quality.NewQualityGate(h.loadGateConfig())
    ...
    gateNotice = output
}
```

- `internal/config/autonomy.go:46` — `IsAutonomyTierCommitGateOff`는 tier ∈ {automatic, fully-autonomous}에서 true (SPEC-STOPCHAIN-TRIM-001 REQ-005의 **의도된** 동작 — 프로덕션 결함 아님).
- 티어가 OFF면 블록(455-464) 전체가 스킵 → `gateNotice` 미설정 → `pre_tool.go:567-568`의 `out.SystemMessage = gateNotice` 할당 자체가 일어나지 않음 → `:78`에서 "SystemMessage is empty".
- `internal/config/autonomy_test.go:35` 주석이 "The kanban launcher exports MOAI_AUTONOMY_TIER into the …"라고 문서화 — 칸반 컴패니언 세션(리드 포함)은 항상 이 변수를 물고 시작한다.

### 결정적 실험 (원인 단일 변수 확정)

세션 ambient 값: `MOAI_AUTONOMY_TIER=fully-autonomous`, `MOAI_KANBAN_LABEL=run-tjv7iy`.

RED (ambient 그대로, fix 전):

```
$ go test ./internal/hook/ -run 'TestPreTool_AstGrepSkipReasonSurfaces' -count=1 -v
    pre_tool_astgrep_reason_test.go:78: SystemMessage is empty: the ast-grep skip reason was dropped somewhere in the three-frame chain
--- FAIL: TestPreTool_AstGrepSkipReasonSurfaces (0.00s)
```

원인 검증 (변수 1개만 걷어냄, fix 전):

```
$ unset MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run 'TestPreTool_AstGrepSkipReasonSurfaces' -count=1 -v
2026/08/17 01:54:45 WARN ast-grep (sg) CLI not found; skipping scan binary=sg hint="install from https://ast-grep.github.io/guide/quick-start.html"
--- PASS: TestPreTool_AstGrepSkipReasonSurfaces (0.02s)
```

### 수정 (fix)

프로덕션 코드 무변경. 테스트에 티어 고정 추가(기존 패턴 `stopchain_ac005_006_test.go:36`의 `t.Setenv(config.EnvAutonomyTier, …)` 준수):

```go
t.Setenv(config.EnvAutonomyTier, config.AutonomyTierSemiAuto)
```

(픽스처 주석 블록에 근거 설명 추가, `internal/config` 임포트 추가.)

GREEN (fix 후, ambient — RED 조건 그대로):

```
$ go test ./internal/hook/ -run 'TestPreTool_AstGrepSkipReasonSurfaces' -count=1 -v
2026/08/17 01:59:32 WARN ast-grep (sg) CLI not found; skipping scan binary=sg hint="install from https://ast-grep.github.io/guide/quick-start.html"
--- PASS: TestPreTool_AstGrepSkipReasonSurfaces (0.00s)
```

GREEN (fix 후, tier 스크럽 — 과적합 없음):

```
$ unset MOAI_AUTONOMY_TIER && go test ./internal/hook/ -run 'TestPreTool_AstGrepSkipReasonSurfaces' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	0.438s
```

패키지 전체 (fix 후, ambient):

```
$ go test ./internal/hook/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	22.714s
```

정적 검증: `gofmt -l` 출력 없음, `go vet ./internal/hook/` 통과.

## 3. Baseline 귀속

- 조사·수정 모두 위 워크트리(HEAD `36a12cf82` = 조사 시점 release/v3.1.1 선두)에서 실측. 출력은 해당 실행의 verbatim.
- fix 전 패키지 전체(ambient) 실패는 본 테스트 1개뿐(`--- FAIL` 목록 실측) — 즉 다른 실패와의 얽힘 없음.

## 4. 미검증 (Gaps)

- `-race` 미실시 (CI `test-race` 잡 소관 — 로컬 전체 스위트/레이스 회피 규율).
- internal/hook 외 패키지 미실시 — 변경이 테스트 파일 1개뿐이므로 파급 없음 (전체 판정은 배치 PR의 CI 몫).

## 5. 잔여 위험 (Residual-risk)

- 칸반 컴패니언 환경변수에 민감한 **다른** 테스트가 다른 패키지에 존재할 수 있음 (본 카드 범위 밖).
- 티어 분기와 무관하게, 티어가 automatic/fully-autonomous인 실사용 세션에서 커밋 게이트의 skip-reason 안내가 사용자에게 안 보이는 것은 SPEC-STOPCHAIN-TRIM-001이 승인한 동작임(의도된 절세).

## 부록 — t50 (internal/astgrep 경로 드리프트) 중복 뿌리 기록 (본체 미건드림)

- `AstGrepGateConfig.RulesDir` 기본값은 `.moai/config/astgrep-rules` (`astgrep_gate.go:139`)이나 저장소가 실제 트래킹하는 룰셋은 `.moai/astgrep-rules/`(go/ 5 + security/ 4 + sgconfig.yml) — CLAUDE.local.md §2.3 이전(2026-08-15)과 정합하는 경로 드리프트.
- 본 카드 결함과는 무관: 이 경로가 문제되려면 스캔에 도달해야 하는데, 실패 경로는 게이트 블록 진입 전 차단이므로 RulesDir은 관여하지 않음. t50 본체 수정은 하지 않음.
