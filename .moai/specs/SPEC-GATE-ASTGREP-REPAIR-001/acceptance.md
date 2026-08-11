---
id: SPEC-GATE-ASTGREP-REPAIR-001
title: "Narrow the Over-Broad ast-grep Error-Wrapping Rule and Scope the moai gate ast-grep Scan — Acceptance"
version: "0.1.0"
status: in-progress
created: 2026-08-11
updated: 2026-08-11
author: GOOS
priority: P1
phase: "v3.1.0 target"
module: "internal/cli, internal/hook/quality, .moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "astgrep, moai-gate, error-wrapping, false-positives, gate-config, config-behavior-consistency"
tier: L
---

# 인수 기준 — SPEC-GATE-ASTGREP-REPAIR-001

> 본 파일은 **검증 계층**이다. 각 AC는 `AC-XXX` 라벨링된 `Given … When … Then …` 시나리오이며 binary-testable해야 한다. GEARS 의무는 **요구사항 계층**(spec.md §B `REQ-GAR-*`)에 있다; 본 파일에서 GEARS를 재진술하지 않는다.

## §A AC 매트릭스 (AC Matrix)

### A.1 D1 — ast-grep 룰 정교화

#### AC-GAR-001 — 비-에러 단일 리터럴 반환 미매칭

**Given** 테스트 fixture Go 파일이 다음 반환문을 포함한다:
```go
func f() int { return 0 }
func g() bool { return true }
func h() error { return nil }
func i() string { return "" }
```
**When** ast-grep이 `go-error-not-wrapped` 룰로 해당 파일을 스캔한다 (`sg scan --config .moai/config/astgrep-rules/sgconfig.yml --rule go-error-not-wrapped <fixture>` 또는 동등한 단일-룰 호출).
**Then** 매칭된 위반은 **0건**이다 (REQ-GAR-001).

#### AC-GAR-002 — `error` 타입 반환 매칭 유지

**Given** 테스트 fixture가 다음을 포함한다:
```go
func f() error {
    err := doSomething()
    if err != nil {
        return err   // wrapping 없이 반환 — 진짜 위반
    }
    return nil
}
```
**When** 동일 룰로 스캔.
**Then** `return err` 라인이 **정확히 1건** 매칭된다 (REQ-GAR-002). `return nil`은 매칭되지 않는다.

#### AC-GAR-003 — autofix 비-에러 미적용

**Given** AC-GAR-001의 fixture.
**When** `sg scan --fix` (또는 동등)로 autofix를 시도한다.
**Then** fixture의 `return 0` / `return nil` / `return true` / `return ""`가 `fmt.Errorf("TODO: operation: %w", ...)`로 변경되지 **않는다** (바이트 동일). (REQ-GAR-003).

**Given-When-Then (D1 통합 시나리오)**:
- Given: 로컬 dogfood 트리의 기존 룰이 `return $ERR` 패턴이었다.
- When: 룰이 정교화된 후(`kind`/`inside`/`constraints` 가드 적용) 동일 fixture를 스캔.
- Then: 진짜 `error` 반환만 매칭; 비-에러 반환 0건; autofix는 진짜 위반에만 적용.

### A.2 D2 — 스캔 범위 / dedup

#### AC-GAR-004 — worktree 경로 제외

**Given** `.claude/worktrees/wt-foo/internal/cli/gate.go`가 존재한다 (메인 트리의 사본).
**When** `moai gate` 또는 `sg scan --config <sgconfig.yml> <project-root>`를 실행한다.
**Then** `.claude/worktrees/**` 경로 아래의 매칭은 결과에 **포함되지 않는다** (REQ-GAR-004).

**검증 명령**:
```bash
moai gate 2>&1 | grep -c "\.claude/worktrees/" || echo 0
# 기대값: 0 (worktree 경로가 결과에 등장하지 않음)
```

#### AC-GAR-005 — vendor / 테스트 파일 제외

**Given** `vendor/` 디렉터리와 `internal/cli/gate_test.go`가 존재한다.
**When** `moai gate`를 실행.
**Then** `vendor/**`와 `*_test.go` 파일의 매칭은 결과에 **포함되지 않는다** (REQ-GAR-004).

#### AC-GAR-006 — 가양성 카운트 하락

**Given** M0 baseline 측정값: `moai gate`의 `go-error-not-wrapped` 매칭 ~N0건 (M0에서 실측; 본 plan-phase에서는 22,293건 총가양성 중 일부로 추정하나 run-phase에서 정확 값 캡처).
**When** M1(룰 정교화) + M2(globs) 적용 후 `moai gate`를 실행.
**Then** 동일 룰의 매칭 수가 baseline 대비 **관측 가능하게 하락**한다 (예: ≤ baseline × 0.1, 또는 특정 절대 임계값 이하; M0 측정 후 M2에서 구체적 임계값 확정).

#### AC-GAR-007 — dedup / 중복 위안 단일 보고

**Given** 메인 트리의 `internal/cli/gate.go:57`이 동일 룰로 매칭된다.
**When** globs exclusion(B.4 주 메커니즘) 또는 `parseSGFindings` dedup이 적용된 후.
**Then** 동일 (file, line, rule) 트리플은 결과에 **정확히 1건만** 등장한다 (REQ-GAR-005).

### A.3 D3 — `moai gate` 설정 정합

#### AC-GAR-008 — gate.yaml 로드

**Given** `.moai/config/sections/gate.yaml`이 다음을 설정한다:
```yaml
ast_grep_gate:
  enabled: true
  block_on_error: false
  warn_only_mode: true
```
**When** `internal/cli/gate.go`가 초기화 중 `loadGateSection`을 호출한다 (M3 적용 후).
**Then** `cfg.AstGrepGate.WarnOnlyMode == true`, `cfg.AstGrepGate.BlockOnError == false`, `cfg.AstGrepGate.Enabled == true`로 로드된다 (REQ-GAR-006). `grep -n "loadGateSection\|DefaultGateConfig" internal/cli/gate.go` 출력이 `loadGateSection` 호출을 포함하고 `DefaultGateConfig()` 하드코딩 라인은 제거/대체되어 있다.

#### AC-GAR-009 — warn_only_mode advisory 동작

**Given** gate.yaml `warn_only_mode: true` + ast-grep 위반이 존재한다.
**When** `moai gate`를 실행한다.
**Then** ast-grep 위험은 stderr/stdout에 **리포트되지만** `moai gate`의 exit code는 **0**(advisory, 하드-블록 아님)이다 (REQ-GAR-007). (참고: 다른 서브-게이트 — vet/lint/test — 가 실패하면 본 AC와 무관하게 exit non-zero일 수 있다; 본 AC는 ast-grep 서브-게이트의 독자적 exit 기여만 단언.)

#### AC-GAR-010 — enabled=false 서브-스캔 건너뛰기

**Given** gate.yaml `ast_grep_gate.enabled: false`.
**When** `moai gate`를 실행.
**Then** ast-grep 서브-스캔은 실행되지 **않는다** (ast-grep 관련 출력 0줄, ast-grep 관련 exit 기여 없음) (REQ-GAR-008).

**Given-When-Then (D3 통합 시나리오)**:
- Given: 기존 `moai gate`가 `DefaultGateConfig()` 하드코딩으로 항상 블로킹 모드로 동작.
- When: M3가 `loadGateSection` 호출을 추가.
- Then: gate.yaml의 advisory 설정이 그대로 반영되어 PreToolUse 경로와 정합.

### A.4 전체 스위트 게이트

#### AC-GAR-011 — 전체 테스트 스위트 green

**Given** M1~M3가 모두 적용된 트리.
**When** `go test -count=1 ./...`를 실행.
**Then** exit code는 **0**이다 (모든 패키지 green).

#### AC-GAR-012 — golangci-lint clean (또는 baseline과 동일)

**Given** M1~M3 적용된 트리.
**When** `golangci-lint run --timeout=5m` 실행.
**Then** exit code는 **0**이거나, 발견된 위반은 모두 **M0 baseline에 이미 존재하던 것**(NEW 위반 0건).

#### AC-GAR-013 — cross-platform build

**Given** M1~M3 적용된 트리.
**When** `GOOS=windows GOARCH=amd64 go build ./...` 실행.
**Then** exit code는 **0**이다 (B1 — gate.go 변경이 syscall을 건드리지 않으므로 예상치는 green이지만 기계 검증 필수).

## §B 심각도 / 추적성 (Severity / Traceability)

| AC | 대응 REQ | 심각도 | 검증 명령/파일 |
|----|----------|--------|----------------|
| AC-GAR-001 | REQ-GAR-001 | MUST | `sg scan --rule go-error-not-wrapped <fixture>` 매칭 0건 |
| AC-GAR-002 | REQ-GAR-002 | MUST | 동일, `return err` 1건 매칭 |
| AC-GAR-003 | REQ-GAR-003 | MUST | autofix 후 fixture byte-diff 0 |
| AC-GAR-004 | REQ-GAR-004 | MUST | `moai gate \| grep -c "\.claude/worktrees/"` == 0 |
| AC-GAR-005 | REQ-GAR-004 | MUST | `moai gate \| grep -c "vendor/\|_test.go"` == 0 (또는 동등) |
| AC-GAR-006 | REQ-GAR-001/004/005 | MUST | M0 baseline 대비 카운트 하락 (임계값 M2 확정) |
| AC-GAR-007 | REQ-GAR-005 | SHOULD | (file,line,rule) 트리플 유니크 |
| AC-GAR-008 | REQ-GAR-006 | MUST | `grep loadGateSection internal/cli/gate.go` 매치 추가 |
| AC-GAR-009 | REQ-GAR-007 | MUST | warn_only_mode + 위반 존재 시 exit 0 |
| AC-GAR-010 | REQ-GAR-008 | MUST | enabled=false 시 ast-grep 출력 0줄 |
| AC-GAR-011 | 전체 | MUST | `go test -count=1 ./...` exit 0 |
| AC-GAR-012 | 전체 | MUST | `golangci-lint run` NEW 위반 0 |
| AC-GAR-013 | 전체 | MUST | `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |

## §C 품질 게이트 (Quality Gate / Definition of Done)

- 모든 MUST AC가 PASS (SHOULD AC는 PASS 또는 명시된 이유로 deferred).
- M0 characterization 테스트가 M1~M3 적용 후 GREEN.
- 가양성 카운트가 22,293건 수준에서 관측 가능하게 하락 (진짜 위반만 남음).
- `moai gate`가 gate.yaml의 advisory 설정을 존중 (하드-블록 지양).
- 병렬 세션 레이스 없음 (Pre-Spawn Sync Check 통과).
- Template-First 미러 시 §25 template-neutrality 준수.
- 진짜 `error`-타입 반환 위반은 여전히 매칭/리포트됨 (회귀 없음).

## §D Edge Cases

- **D1 edge**: 멀티값 반환 `return a, err` — 본 룰은 단일 `return $ERR` 패턴이므로 멀티값은 매칭되지 않을 수 있다. plan.md §D.1의 positive fixture는 단일값 `return err`에 집중; 멀티값은 본 SPEC 범위 밖(명시 후보)이며 run-phase에서 회귀 테스트 추가 여부 결정.
- **D2 edge**: 심볼릭 링크된 worktree 경로 — globs `!**/.claude/worktrees/**`가 심볼릭 링크를 탐지하는지 run-phase에서 검증.
- **D3 edge**: gate.yaml이 존재하지 않거나 파싱 불가인 경우 — advisory 기본값으로 폴백 (B5). fallback이 hard-block이 아니어야 한다.
- **D3 edge**: PreToolUse 경로와의 정합 — 두 경로가 동일한 `loadGateSection` 로더를 쓰므로 설정이 일치해야 한다 (run-phase에서 두 경로의 cfg를 비교하는 회귀 테스트 후보).

## §E Forward-Looking Checks

- `moai gate`의 advisory vs blocking 동작이 gate.yaml에 의해 완전히 구동되는지 (하드코딩 fallback조차 advisory 기본).
- ast-grep 버전 업그레이드 시 `globs` 문법이 호환되는지 (B3) — sg 릴리스 노트 모니터링 대상.
- 로컬 dogfood 트리와 distributed baseline의 룰셋 drift (DOGFOOD-CLEANUP-001과의 상호보완).
