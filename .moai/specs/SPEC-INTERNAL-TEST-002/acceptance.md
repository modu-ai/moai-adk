---
id: SPEC-INTERNAL-TEST-002
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
---

# SPEC-INTERNAL-TEST-002 — acceptance

## §D AC Matrix

> Headline gate 참조: 본 SPEC은 **per-debt gates only** (spec.md §A "Headline gate 범위" 단락). `go test ./internal/... exit 0`는 2개 `internal/web` i18n 실패(`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`)로 인해 본 SPEC 머지 후에도 달성되지 않으며, 이를 claim하지 않는다 (Out of Scope — internal/web i18n 2개 신규 실패).

| AC ID | Milestone | Gate type | Severity | Bind REQ | Description |
|-------|-----------|-----------|----------|----------|-------------|
| AC-TEST-007 | M1 | per-debt MUST-PASS | P0 | REQ-TEST-007 | 6× internal/cli golden testdata 재생성 후 6 TestDoctor/TestStatus 테스트 동시 PASS (ARCH-001 M0 prerequisite) |
| AC-TEST-008 | M2 | per-debt MUST-PASS (verify-only) | P2 | REQ-TEST-008 | `TestBuild_WritesContextUsageWithSessionID -count=10` 연속 PASS (commit 794bb4f84 already-resolved evidence 재생성) |
| AC-TEST-009 | M3 | per-debt MUST-PASS | P1 | REQ-TEST-009 | `internal/constitution` package coverage 67.5% → ≥ 85% (pipeline.go 8함수 + human_oversight.go 2함수 integration test 추가) |

## §D.1 Severity / Traceability

- **P0 (M1, AC-TEST-007)**: ARCH-001 plan-audit BLOCKING defect D1 직결. 본 AC가 PASS하는 것은 ARCH-001 run-phase 진입의 **필요 조건 (necessary not sufficient)** (spec.md §A "ARCH-001 선행 관계"). ARCH-001 AC-ARCH-001(whole-repo `go test ./... exit 0`) 달성을 위해 후속 web-i18n SPEC(TEST-003) 완료가 추가로 필요 — 본 SPEC M1만으로는 잔여 2개 internal/web i18n FAIL로 whole-repo non-green.
- **P1 (M3, AC-TEST-009)**: coverage floor 85% (CLAUDE.local.md §6). TEST-001 DEBT-TEST-003 정식 청산. integration test 저작 effort 가장 집중.
- **P2 (M2, AC-TEST-008)**: 이미 해결된 debt(794bb4f84). evidence 재생성만 수행. 코드 변경 0건.

## §D.2 Given-When-Then Scenarios

### AC-TEST-007 — internal/cli golden testdata 재생성 (M1)

**Given** repository HEAD에 아래 stale 상태가 존재한다:
- `pkg/version/version.go:8` → `Version = "v3.0.0-rc8"`
- `internal/cli/testdata/doctor-{light,dark,nocolor}.golden:13` → stale `v3.0.0-rc6` 3건
- `internal/cli/testdata/status-{light,dark,nocolor}.golden:6` → stale `v3.0.0-rc6` 3건
- 6개 테스트(`TestDoctor_Current_Light/Dark`, `TestDoctor_NoColor`, `TestStatus_Current_Light/Dark`, `TestStatus_NoColor`)가 `doctor_golden_test.go:91,110,128` 및 `status_golden_test.go:142,155,167`에서 FAIL 중

**When** run-phase가 아래 명령을 실행한다:
```bash
UPDATE_GOLDEN=1 go test ./internal/cli/ \
  -run 'TestDoctor_Current_Light|TestDoctor_Current_Dark|TestDoctor_NoColor|TestStatus_Current_Light|TestStatus_Current_Dark|TestStatus_NoColor' \
  -count=1
```

**Then**:
1. 위 6개 golden testdata 파일이 현재 `v3.0.0-rc8` 버전 문자열로 재생성된다 (renderer 출력과 byte-identical). **regen-blindness 방지 기계 검증 (필수)**: `grep -l "v3.0.0-rc8" internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden | wc -l` → `6` (6개 파일 전부가 rc8을 carry 확인 — 단순 선언이 아닌 binary assertion). **byte-level sanity (권장)**: `git diff internal/cli/testdata/ | grep -E '^[+-].*v3\.0\.0-rc[0-9]' | sort -u` → 정확히 6개의 rc6→rc8 단일 문자 diff만 표시 (layout/이모지/구조 변경 0건 — C-1 renderer 무접촉 확인).
2. `UPDATE_GOLDEN` env를 제외한 재실행 시 6개 테스트 전부 PASS:
```bash
go test ./internal/cli/ \
  -run 'TestDoctor_Current_Light|TestDoctor_Current_Dark|TestDoctor_NoColor|TestStatus_Current_Light|TestStatus_Current_Dark|TestStatus_NoColor' \
  -count=1
# → ok github.com/modu-ai/moai-adk/internal/cli ... (no FAIL)
```
3. `git diff --stat`이 `internal/cli/testdata/` 하위 6개 파일만 표시 (renderer.go / doctor.go / status.go 무접촉, C-1 준수)

### AC-TEST-008 — statusline env-isolation verify-only (M2)

**Given** repository HEAD가 commit `794bb4f84` (2026-07-09 "test(statusline): isolate context-usage test from ambient GLM env")를 포함한다:
- `internal/statusline/builder_test.go:1815-1818`에 4개 `t.Setenv` 호출 존재 (MOAI_STATUSLINE_CONTEXT_SIZE / ANTHROPIC_DEFAULT_OPUS_MODEL / SONNET / HAIKU 모두 ""로 격리)
- TEST-001 §E.4 sync 시점(2026-07-08)에는 이 fix가 없어 deterministic FAIL이었으나, 현재 HEAD에서는 이미 PASS 상태

**When** run-phase가 아래 명령을 실행한다 (동일 명령을 ambient GLM env 주입 상태에서도 반복):
```bash
go test -run TestBuild_WritesContextUsageWithSessionID -count=10 ./internal/statusline/
# 그리고 GLM env 주입 상태에서 반복:
ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2 ANTHROPIC_DEFAULT_SONNET_MODEL=glm-5.2 \
  ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-5.2 \
  go test -run TestBuild_WritesContextUsageWithSessionID -count=3 ./internal/statusline/
```

**Then**:
1. 첫 번째 명령: 10 consecutive runs 전부 PASS (`ok ... no FAIL`)
2. 두 번째 명령(GLM env 주입): 3 consecutive runs 전부 PASS (`ok ... no FAIL`) — isolation이 ambient env를 정상 차단 증명
3. 본 AC의 run-phase 보고는 evidence-only (코드 변경 0건, C-2 준수). commit-scope 검증은 2가지 허용 form 중 하나: (a) `git diff <M2-commit>^..<M2-commit> -- internal/statusline/` → empty (where `<M2-commit>` is the SHA recorded in progress.md §E.2 for M2); (b) M2가 M1/M3와 동일 commit에 bundle된 경우(plan.md §F 허용), run-phase range form으로 대체: `git log --name-only <TEST-002-run-start>..HEAD -- internal/statusline/*.go | grep -v _test.go` → empty (본 SPEC run-phase range에서 internal/statusline/ production 파일 무수정). `HEAD~..HEAD` 단일-commit form은 M2가 독립 commit이 아닐 때 M1/M3 변경까지 노출하므로 사용 금지.

### AC-TEST-009 — internal/constitution coverage 67.5%→85% (M3)

**Given** repository HEAD의 `internal/constitution` package coverage가 67.5%이며, 아래 함수들이 0% 또는 near-0%이다:
- `pipeline.go`: `NewPipeline` 0%, `Execute` 0%, `createLogEntry` 0%, `applyAmendment` 0%, `acquireLock` 0%, `releaseLock` 0%, `updateSourceFile` 0%, `updateRegistryClause` 0% (8함수)
- `human_oversight.go`: `Approve` 14.3%, `printDiff` 0%

**When** run-phase가 `internal/constitution/{pipeline_test.go,human_oversight_test.go}` (또는 동등 파일)에 integration test를 추가한 뒤 아래 명령을 실행한다:
```bash
go test -coverprofile=cover.out ./internal/constitution/
go tool cover -func=cover.out | tail -1
# 그리고 함수별 확인:
go tool cover -func=cover.out | grep -E 'pipeline\.go|human_oversight\.go'
```

**Then**:
1. `total: ... 85.0%` 이상 (≥ 85%, CLAUDE.local.md §6 floor)
2. `pipeline.go` 8함수 전부 0% 탈출 (각 함수 ≥ 1 path 커버)
3. `human_oversight.go::printDiff` 0% 탈출, `Approve` 14.3% → 유의미 상승
4. 모든 신규 테스트가 `t.TempDir()` isolation 사용 (C-5 준수, 실제 `.claude/rules/moai/core/` 무변경)
5. `pipeline.go` / `human_oversight.go` production 코드 무변경 (C-3 준수 — `git diff HEAD~..HEAD -- internal/constitution/*.go | grep -v _test.go` empty)

## §D.3 Edge Cases

- **EC-1 (M1 version 재틈)**: M1 머지 후 다음 release에서 version이 rc8→rc9로 올라가면 동일 6개 golden이 다시 stale化 된다. 이는 근본적으로 testdata가 hardcoded version을 carry하기 때문이며, 본 SPEC은 mechanism(`UPDATE_GOLDEN=1`)이 이미 존재하므로 "매 release마다 regen하는 release-runbook 항목"을 추가하는 scope는 아니지만 future-work 후보로 기록.
- **EC-2 (M1 renderer edge)**: 만약 `UPDATE_GOLDEN=1` 실행 시 renderer가 예상과 다른 출력을 내면 (단순 rc6→rc8 단일 문자 diff가 아닌 layout 변경이 감지되면), 그것은 renderer regression이며 본 SPEC scope 밖. blocker report로 반환 (C-1 위반 신호).
- **EC-3 (M2 환경 의존성)**: dev machine에서 GLM env가 자동 주입되는 환경(`moai glm` / `moai cg` 세션)에서 AC-TEST-008의 두 번째 명령이 예상과 다르게 FAIL하면, 794bb4f84의 isolation에 추가 환경 변수 구멍이 있다는 뜻. 이 경우 본 SPEC scope는 "이미 해결됨 재검증"이므로, 새로운 env leak 발견 시 blocker report로 반환한 뒤 TEST-002 scope 확장 또는 TEST-003 위임.
- **EC-4 (M3 lock 경합)**: pipeline.go의 `acquireLock`/`releaseLock`이 실제 file lock을 사용하므로 병렬 테스트 실행 시 교착 가능. `t.Parallel()` 사용 자제 또는 lock 경로를 `t.TempDir()`로 격리 필수 (C-5).
- **EC-5 (M3 source amend 부작용)**: `updateSourceFile` / `updateRegistryClause`가 실제 source 파일에 write를 수행하므로, 테스트가 `t.TempDir()` 밖의 경로를 건드리면 dev tree 오염. 모든 test fixture는 `t.TempDir()` 하위에 한정.

## §D.4 Closure Gates (Definition of Done)

본 SPEC이 3-phase close 되기 위해 필요한 조건:

- [ ] AC-TEST-007 PASS (6개 golden testdata 재생성 + 6 TestDoctor/TestStatus PASS)
- [ ] AC-TEST-008 PASS (10 consecutive + 3 GLM env 주입 연속 PASS)
- [ ] AC-TEST-009 PASS (constitution coverage ≥ 85%, 8 pipeline.go + 2 human_oversight.go 함수 0% 탈출)
- [ ] Cross-platform build (`go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`) exit 0
- [ ] Lint (`golangci-lint run --timeout=2m`) 0 NEW issues (baseline 대비)
- [ ] Working tree 보호 (C-4): IGGDA Path-B 잔류 18+5 파일이 본 SPEC commit에 미포함 (`git log --name-only HEAD~..HEAD`로 확인)
- [ ] PRESERVE list 존중 (§A.5): TEST-001 / ARCH-001 산출물 무변경
- [ ] 진행 보고 `progress.md §E.2 / §E.3`에 AC PASS evidence + commit SHA 기록 (manager-develop run-phase)
- [ ] sync-phase `progress.md §E.4`에 `sync_commit_sha` + 3-phase close (manager-docs)

## §D.5 Forward-Looking Checks (비어 있으면 future debt 표기)

- **FL-1 (version-bump runbook)**: release 시마다 `UPDATE_GOLDEN=1`을 포함하는 runbook 항목이 release-procedure에 없다면 future debt. 본 SPEC은 runbook 편집 범위 밖이나, 기록으로 남김.
- **FL-2 (ARCH-001 re-entry)**: 본 SPEC M1 + 후속 web-i18n SPEC(TEST-003) 머지 후 ARCH-001 plan-audit iter-2가 D1 resolved로 PASS하는지 추적. 본 SPEC만으로는 ARCH-001 AC-ARCH-001(whole-repo `go test ./... exit 0`) 미달성 (잔여 2개 internal/web i18n FAIL)이므로, web-i18n SPEC 완료가 추가 선행 조건. 본 SPEC close 시점에는 미확인 사항.
- **FL-3 (internal/web i18n 2개)**: `TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity` 후속 SPEC(SPEC-INTERNAL-TEST-003 또는 web-i18n 특화)에서 청소 필요. 본 SPEC scope 밖 명시.
- **FL-4 (M3 추가 coverage)**: 본 SPEC이 85%를 달성해도 남은 15% (주로 error path / rare branch)는 future debt 후보. 100% 달성을 본 SPEC이 claim하지 않음.

## §D.6 Indirect Verification (정성 평가)

- 본 SPEC run-phase는 `moai spec audit` 시 V3R6 era로 정상 분류되어야 (progress.md §E skeleton + sync_commit_sha 포함 시 H-4 매칭).
- TEST-001의 `acceptance.md §D.1` "Binding gate MUST-PASS AC-TEST-001 headline" body는 **본 SPEC이 수정하지 않는다** (TEST-001 산출물은 read-only). 본 SPEC은 TEST-001의 headline reinterpretation을 계승(또는 반복)하지 않고 per-debt gate로 한정.

## §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: AC-TEST-007/008/009 전부 기계적 도구로 검증 (`go test`, `go tool cover`)
- **Readable**: 신규 `pipeline_test.go` / `human_oversight_test.go`는 기존 `internal/constitution/*_test.go` 스타일 계승 (TEST-001이 추가한 `canary_test.go`, `contradiction_test.go` 패턴)
- **Unified**: 테스트 코드는 `gofmt` / `goimports` 준수, error wrapping은 `fmt.Errorf("...: %w", err)` (CLAUDE.local.md §3)
- **Secured**: 신규 테스트가 credentials / real fs 경로 무접촉 (`t.TempDir()` 강제)
- **Trackable**: Conventional Commits + `🗿 MoAI` trailer + 본 SPEC plan.md §F Milestone 순서 준수
