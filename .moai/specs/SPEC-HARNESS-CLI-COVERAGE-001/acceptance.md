# SPEC-HARNESS-CLI-COVERAGE-001 — 인수 기준 (acceptance.md)

> 모든 AC는 anchored & measurable(greppable assertion). 측정 baseline은 spec.md §B(77.9%)를
> SSOT로 한다. 패키지 경로: `github.com/modu-ai/moai-adk/internal/cli/harness`.

## §A — 측정 명령 (재현 가능)

```bash
# 패키지 coverprofile + total 행
go test -coverprofile=/tmp/hcc.out ./internal/cli/harness/...
go tool cover -func=/tmp/hcc.out | tail -1          # total: ... ≥ 90.0%

# 함수별 커버리지 (gap 확인)
go tool cover -func=/tmp/hcc.out | grep -E 'NewInstallCmd|runExecuteCommand|NewExecuteCmd|RunInstall|runPropose'

# 프로덕션 무수정 검증
git diff --name-only internal/cli/harness/ | grep -vE '_test\.go$'   # (출력 없음 = PASS)

# 경계 + 회귀 가드
go test -run 'TestPropose_NoAskUserQuestion' ./internal/cli/harness/...
go test -run 'TestRunExecute_RegressionGateMeasuresProjectRoot' ./internal/cli/harness/...
```

---

## §B — 인수 기준 (AC-HCC-*)

### AC-HCC-001 — 패키지 coverage ≥90% [MUST]

- **Given** `internal/cli/harness` 패키지에 본 SPEC의 신규 테스트가 추가된 상태
- **When** `go test -coverprofile=/tmp/hcc.out ./internal/cli/harness/...` 실행 후
  `go tool cover -func=/tmp/hcc.out | tail -1`을 측정하면
- **Then** total 행이 **≥ 90.0%**를 보고한다 (baseline 77.9%에서 ≥12.1pp 증가).
- 검증: `go tool cover -func=/tmp/hcc.out | tail -1` 출력의 백분율 ≥ 90.0
- (REQ-HCC-001)

### AC-HCC-002 — `NewInstallCmd` ≥87.5% (reachable ceiling) [MUST]

- **Given** M1 테스트(RunE 클로저 success / default-cwd / RunInstall-error)가 추가된 상태
- **When** 함수별 커버리지를 측정하면
- **Then** `install.go:98 NewInstallCmd`가 **≥87.5%**를 보고한다 (baseline 33.3%).
- 검증: `go tool cover -func=/tmp/hcc.out | grep 'NewInstallCmd'` → ≥87.5%
- 노트: 87.5%는 reachable ceiling이다. 24 statements 중 3개는 표준 Go 테스트로 도달 불가한
  documented residual(os.Getwd 실패 129-131 / filepath.Abs 실패 137-139 / MarkFlagRequired panic
  168-169)로, 이 3개가 함수를 capping한다. 측정 87.5%가 재조정된 threshold를 충족한다. (§D.4 / EX-2)
- (REQ-HCC-002 / REQ-HCC-007 / REQ-HCC-008)

### AC-HCC-003 — `RunInstall` 100% [MUST]

- **Given** M2 테스트(empty-root error + ScaffoldHarnessDir 실패) + 기존 happy/InjectMarker-error
  테스트가 모두 존재
- **When** 함수별 커버리지를 측정하면
- **Then** `install.go:60 RunInstall`이 **100%**를 보고한다 (baseline 80.0%).
- 검증: `go tool cover -func=/tmp/hcc.out | grep 'RunInstall'` → 100.0%
- (REQ-HCC-004 / REQ-HCC-013 / REQ-HCC-014)

### AC-HCC-004 — `runExecuteCommand` ≥90% [MUST]

- **Given** M3 테스트(success stdout emit + InheritedFlags success 경로)가 추가된 상태
- **When** 함수별 커버리지를 측정하면
- **Then** `execute.go:355 runExecuteCommand`가 **≥90%**를 보고한다 (baseline 62.5%).
- 검증: `go tool cover -func=/tmp/hcc.out | grep 'runExecuteCommand'` → ≥90.0%
- (REQ-HCC-003 / REQ-HCC-009 / REQ-HCC-010)

### AC-HCC-005 — `runPropose` ≥90% [MUST]

- **Given** M4 테스트(non-dry-run WriteProposals + default-flag fallback)가 추가된 상태
- **When** 함수별 커버리지를 측정하면
- **Then** `propose.go:78 runPropose`가 **≥90%**를 보고한다 (baseline 78.1%).
- 검증: `go tool cover -func=/tmp/hcc.out | grep 'runPropose'` → ≥90.0%
- (REQ-HCC-005 / REQ-HCC-011 / REQ-HCC-012)

### AC-HCC-006 — `NewExecuteCmd` ≥69.2% (reachable ceiling; os.Exit/panic residual) [MUST]

- **Given** M3 테스트로 `runExecuteCommand`가 success/inherit 경로로 호출되어 RunE 클로저의
  비-os.Exit statement가 도달된 상태
- **When** 함수별 커버리지를 측정하면
- **Then** `execute.go:298 NewExecuteCmd`가 **≥69.2%**를 보고하며, 미커버 잔존은 os.Exit body
  (329-334 = 3 stmts)와 MarkFlagRequired panic(344-345 = 1 stmt)에 한정된다.
- 검증: `go tool cover -func=/tmp/hcc.out | grep 'NewExecuteCmd'` → ≥69.2%
- 노트: 69.2%는 reachable ceiling이다. 13 statements 중 4개(os.Exit body 3 + panic 1)는 표준 Go
  테스트로 도달 불가한 documented residual이다. RunE success path(327 진입 + 335 return-nil)는
  커버되며, residual은 os.Exit+panic에 한정된다. (§D.4 / EX-2)
- (REQ-HCC-006)

### AC-HCC-007 — 격리 [MUST]

- **Given** 본 SPEC의 모든 신규 테스트
- **When** 테스트 실행 중 파일시스템 write가 발생하면
- **Then** 모든 write는 `t.TempDir()` 내부에서만 발생하고 프로젝트 root는 mutate되지 않는다.
- 검증(a): `git status --porcelain` — 테스트 실행 후 추적 파일 변경 없음(신규 코드 외).
- 검증(b): os.Chdir을 사용하는 subtest는 `t.Parallel()`을 호출하지 않으며 cwd를 복원한다 —
  소스 grep: `grep -A5 'os.Chdir' internal/cli/harness/*_test.go`에서 동일 함수 내 `t.Parallel()` 부재.
- (REQ-HCC-016 / REQ-HCC-018)

### AC-HCC-008a — C-HRA-008 경계 무회귀 [MUST]

- **Given** 본 SPEC의 신규 테스트
- **When** `go test -run 'TestPropose_NoAskUserQuestion' ./internal/cli/harness/...`를 실행하면
- **Then** PASS한다 (어떤 신규 테스트도 `AskUserQuestion(` 호출을 도입하지 않음).
- 검증: 위 `go test -run 'TestPropose_NoAskUserQuestion'` exit 0 — 이 Go 테스트가 canonical
  authoritative guard다(`strings.Contains(src, "AskUserQuestion(")` open-paren 매칭으로 실제 호출만
  탐지하며, godoc prose나 테스트 자체 이름에 오탐하지 않는다).
- 노트(D1 정정): 이전 secondary `grep -rn 'AskUserQuestion\|mcp__askuser'` bullet은 제거되었다 —
  해당 grep은 cobra `Long:` 문자열 내 godoc 텍스트와 경계 테스트 자체 이름에 매칭되어 clean tree에서도
  9 매치를 반환하는 false-fail이었다. open-paren 호출 형태만 검사하는 위 Go 테스트가 sound guard다.
- (REQ-HCC-017)

### AC-HCC-008b — seam 예외 (조건부) [MUST, 발동 시에만]

- **Given** run-phase 실측에서 `NewExecuteCmd`가 reachable ceiling(69.2%, os.Exit body 3 + panic 1
  residual 제외) 미만이고 다른 도달 가능 statement가 없는 경우(plan.md §D fallback)
- **When** 최소 테스트 seam(injectable exitFunc 등)을 프로덕션 코드에 도입하면
- **Then** 그 변경은 progress.md에 명시 enumerate되고 plan-auditor 정밀 검토 + 사용자 승인 대상으로
  flag된다.
- 검증: **기본 미발동**. 발동 시 `git diff --name-only`에 프로덕션 `.go`가 나타나며 progress.md에
  seam 정당화가 기록된다. 미발동이 default·정상 상태.
- (REQ-HCC-022)

### AC-HCC-009 — documented acceptable residual 열거 [MUST]

- **Given** ≥90% 달성 후 잔존 미커버 statement
- **When** residual을 검토하면
- **Then** 잔존은 아래 원칙적 항목에 **한정**되며 progress.md에 열거된다(accidental ceiling 아님):
  1. `execute.go:329-334` — `NewExecuteCmd` os.Exit 분기 (표준 Go 테스트 도달 불가; 로직은
     `runExecuteCommand`에 추출되어 별도 커버)
  2. `execute.go:344-345` — `NewExecuteCmd` MarkFlagRequired panic (방어적 도달 불가)
  3. `install.go:168-169` — `NewInstallCmd` MarkFlagRequired panic (방어적 도달 불가)
  4. (조건부) `execute.go:116/124/159` — `RunExecute`/`runExecuteWith`의 `os.Getwd()`/`filepath.Abs`
     실패 분기 — 결정론적 유발이 test-only로 불가할 경우 residual로 명시(EX-2/EX-4)
- 검증: progress.md §F(또는 동등 섹션)에 위 4(±1) 항목이 파일:라인과 함께 열거됨.
- (EX-2 / EX-4 / REQ-HCC-015)

### AC-HCC-010 — 프로덕션 무수정 [MUST] (seam 예외 시 AC-HCC-008b로 대체)

- **Given** 본 SPEC의 변경
- **When** `git diff --name-only internal/cli/harness/`를 확인하면
- **Then** 패키지 하위 변경은 `*_test.go` 파일만 포함한다(seam 예외 미발동 시).
- 검증: `git diff --name-only internal/cli/harness/ | grep -vE '_test\.go$'` → 출력 없음.
- (REQ-HCC-021 / EX-1)

### AC-HCC-011 — `--execute` fail-close 무회귀 [MUST]

- **Given** 본 SPEC의 신규 테스트
- **When** `go test -run 'TestRunExecute_RegressionGateMeasuresProjectRoot|TestRunExecute_RED' ./internal/cli/harness/...`를
  실행하면
- **Then** PASS한다 (SPEC-HARNESS-EXECUTE-E2E-001 measurementRoot threading 무회귀).
- 검증: 위 명령 exit 0.
- (REQ-HCC-019)

### AC-HCC-012 — 전체 빌드/lint 무회귀 [MUST]

- **Given** 본 SPEC의 변경
- **When** `go test ./...` + `go vet ./...` + `golangci-lint run`을 실행하면
- **Then** `go test ./...` 전체 green, `go vet ./...` clean, `golangci-lint run` baseline 유지.
- 검증: 세 명령 모두 exit 0 (lint은 baseline 대비 신규 finding 0).
- 추가: cross-platform — `GOOS=windows GOARCH=amd64 go build ./...` + `GOOS=linux GOARCH=amd64 go build ./...` exit 0
  (REQ-HCC 신규 테스트가 hardcoded separator를 도입하지 않음, C6).
- (REQ-HCC-020)

---

## §C — Edge Cases

- **EC-1 — os.Chdir 복원 실패**: default-cwd 테스트가 chdir 후 panic하면 후속 테스트 cwd 오염.
  → `t.Cleanup(func(){ os.Chdir(orig) })`로 복원 보장 + non-parallel(AC-HCC-007b).
- **EC-2 — ScaffoldHarnessDir 충돌 유발 정확성**: M2의 regular-file 충돌이 OS별로 다른 에러 메시지를
  내도 wrapped error의 `"scaffold .moai/harness"` prefix로 검증(메시지 본문 의존 금지).
- **EC-3 — WriteProposals 픽스처 actionable 보장**: M4 fixture는 `code_change:` prefix + to_tier
  `recommendation`이어야 `MapPromotions`가 candidates>0를 반환(기존 `TestPropose_AutoFlagWithActionableData`
  픽스처와 동일 형식 재사용).
- **EC-4 — runExecuteCommand success 픽스처**: real Go 모듈(`writeE2EFixtureModule`) + valid
  proposal 필요 — gate-active Apply가 `go test ./...`를 실제 실행하므로 빌드 가능한 모듈이어야 함.

---

## §D — Definition of Done (체크리스트)

- [ ] AC-HCC-001: 패키지 total coverage ≥ 90.0%
- [ ] AC-HCC-002: `NewInstallCmd` ≥87.5% (reachable ceiling; 3 residual)
- [ ] AC-HCC-003: `RunInstall` 100%
- [ ] AC-HCC-004: `runExecuteCommand` ≥90%
- [ ] AC-HCC-005: `runPropose` ≥90%
- [ ] AC-HCC-006: `NewExecuteCmd` ≥69.2% (reachable ceiling; os.Exit body 3 + panic 1 residual)
- [ ] AC-HCC-007: 모든 write가 t.TempDir() 내부 + os.Chdir subtest non-parallel
- [ ] AC-HCC-008a: `TestPropose_NoAskUserQuestion` green
- [ ] AC-HCC-008b: seam 예외 미발동(또는 발동 시 명시+flag)
- [ ] AC-HCC-009: documented residual 4(±1)항목 열거
- [ ] AC-HCC-010: 프로덕션 non-test `.go` 무수정
- [ ] AC-HCC-011: `execute_e2e_test.go` green (fail-close 무회귀)
- [ ] AC-HCC-012: `go test ./...` + vet + lint + cross-build 모두 green

---

## §E — Traceability (AC ↔ REQ ↔ Milestone)

| AC | REQ | Milestone | 검증 방식 |
|----|-----|-----------|-----------|
| AC-HCC-001 | REQ-HCC-001 | M6 | coverprofile total 행 |
| AC-HCC-002 | REQ-HCC-002/007/008 | M1 | func 커버리지 grep (≥87.5% ceiling) |
| AC-HCC-003 | REQ-HCC-004/013/014 | M2 | func 커버리지 grep (100%) |
| AC-HCC-004 | REQ-HCC-003/009/010 | M3 | func 커버리지 grep |
| AC-HCC-005 | REQ-HCC-005/011/012 | M4 | func 커버리지 grep |
| AC-HCC-006 | REQ-HCC-006 | M3 | func 커버리지 grep (≥69.2% ceiling) |
| AC-HCC-007 | REQ-HCC-016/018 | M1 | git status + grep |
| AC-HCC-008a | REQ-HCC-017 | M6 | 경계 가드 test + grep |
| AC-HCC-008b | REQ-HCC-022 | M5/M6(조건부) | git diff + progress 명시 |
| AC-HCC-009 | EX-2/EX-4/REQ-HCC-015 | M5/M6 | progress residual enum |
| AC-HCC-010 | REQ-HCC-021/EX-1 | M6 | git diff --name-only |
| AC-HCC-011 | REQ-HCC-019 | M6 | e2e test green |
| AC-HCC-012 | REQ-HCC-020 | M6 | test/vet/lint/cross-build |
