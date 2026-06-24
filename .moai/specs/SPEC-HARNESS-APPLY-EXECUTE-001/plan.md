# Implementation Plan — SPEC-HARNESS-APPLY-EXECUTE-001

## §A. Context

Self-Harness 로드맵 P2 "observer/gate activation" 1차. `Applier.Apply()`(프로덕션 caller 0)를 처음으로 live 경로에 배선하는 thin caller(opt-in Go execute verb)를 추가한다. Tier M, cycle_type=tdd. user decision A (opt-in, Path S 유지) 확정.

## §B. Known Issues / 사전 검증된 ground-truth

(manager-spec가 코드를 직접 탐색하여 확인한 사실 — design.md에서 정확 line 재확인 권장)

1. `Applier.Apply()` 프로덕션 caller = 0 (applier.go:247). 첫 caller가 본 SPEC.
2. `safety.NewPipeline` 프로덕션 caller = 0 (pipeline.go:55). `PipelineConfig.AutoApply` 필드 존재 (pipeline.go:51).
3. L5 동작: `autoApply==false` → 항상 PendingApproval (pipeline.go:147~153); `autoApply==true` → Approved (pipeline.go:155~157).
4. `NewApplierWithRegressionGate(manifestPath, baselinePath)` (applier.go:84) + `.WithOutcomeObserver(obs)` (applier.go:102) fluent 배선.
5. `harness.NewObserver(logPath)` (observer.go:31).
6. `RecordOutcome` 타겟: `usage-log.jsonl` (outcome.go:57~73, `RecordExtendedEvent` append path, EventType `EventTypeApplyOutcome`).
7. baseline const: `.moai/harness/measurements-baseline.yaml` (regression_gate.go:98 주석 + 117 NewBaselineStore).
8. canary nil-safe: `baselineScore([])` = 0 (canary.go:34), `defaultProjectedScorer` meaningful proposal → baseline+0.02 (canary.go:56~62), drop=0이라 L2 미reject.
9. 경로 const (harness.go): `harnessDefaultProposalDir`(:40), `harnessDefaultSnapshotBase`(:37), `harnessDefaultLogPath`(:34). manifest는 `<learning-history>/manifest.jsonl` (applier.go:41 주석).
10. verb-factory 패턴: `propose.go`/`install.go` (package `internal/cli/harness`), boundary guard `TestPropose_NoAskUserQuestion` (propose_boundary_test.go) — 디렉터리 내 모든 `.go`에서 `AskUserQuestion(` 부재 검증.
11. 기존 `apply` verb (`harness.go:203 newHarnessApplyCmd` → `runHarnessApply` :219)는 payload-only. package `cli`(boundary guard 미적용 디렉터리).
12. harness.yaml `auto_apply: false` (line 116).

## §C. Pre-flight (run-phase 진입 전 manager-develop 확인)

- [ ] `go build ./...` exit 0 (baseline)
- [ ] `go test ./internal/cli/harness/... ./internal/harness/...` green (baseline)
- [ ] design.md §D Decision 표의 verb shape 최종값 재확인 (option (b) 채택)
- [ ] 위 §B의 line 번호 grep 재확인 (코드 이동 가능성)

## §D. [HARD] Constraints (FROZEN — DO-NOT-MODIFY)

| # | 불변식 | 근거 |
|---|--------|------|
| C1 | `harness.yaml`의 `auto_apply: false` 디스크 값 변경 금지 (in-memory AutoApply=true만) | §B.2 spec.md |
| C2 | `applier.go` 로직 변경 금지 (배선만, 호출만) | §D.3 Exclusions |
| C3 | `regression_gate.go` / `outcome.go` / `pipeline.go` / `canary.go` / `lineage.go` 로직 변경 금지 | §D.3 Exclusions |
| C4 | 기존 `apply` payload-only 동작 보존 (`--execute` 부재 시) | REQ-AEX-003 |
| C5 | Path S (skill Edit) default 유지 — 전환/대체 금지 | §D.1 Exclusions |
| C6 | `internal/cli/harness/` 내 `AskUserQuestion(` 부재 (C-HRA-008) | REQ-AEX-016 |
| C7 | "prevents regressions" 과대광고 금지 — telemetry framing만 | §A.2 / §D.5 |

## §E. Self-Verification (manager-develop가 각 milestone 후 실행)

```bash
# 1. 전체 빌드 + 테스트
go build ./... && go test -count=1 ./internal/cli/harness/... ./internal/harness/...
# 2. cross-compile (CLI는 cross-platform 필수)
GOOS=windows GOARCH=amd64 go build ./...
# 3. C-HRA-008 boundary guard (신규 execute.go 포함)
go test -run TestPropose_NoAskUserQuestion ./internal/cli/harness/...
# 4. autoApply contract invariant grep (AutoApply: true 존재, harness.yaml 디스크 mutate 부재)
grep -n 'AutoApply: *true' internal/cli/harness/execute.go
grep -rn 'auto_apply' internal/cli/harness/ | grep -v '_test.go'   # 디스크 write 부재 확인
# 5. lint
golangci-lint run --timeout=2m ./internal/cli/harness/... ./internal/harness/...
```

## §F. Milestones (Tier M)

### M1 — execute.go verb-factory + RunExecute 골격 (RED→GREEN)
- 신규 `internal/cli/harness/execute.go`: `NewExecuteCmd()` (export, `--id` 필수 flag, `--project-root`) + `RunExecute(opts ExecuteOptions) error`.
- proposal loader: `.moai/harness/proposals/<id>.json` → `harness.Proposal` (REQ-AEX-004). 부재 시 exit 1 (REQ-AEX-012).
- RED: `execute_test.go` — proposal 부재 시 user error, ID 정상 로드.
- AC 매핑: AC-AEX-001, AC-AEX-002, AC-AEX-012.

### M2 — Apply 파이프라인 배선 + autoApply contract (RED→GREEN)
- `safety.NewPipeline(PipelineConfig{AutoApply: true, ViolationLogPath, RateLimitPath})` 구성 (REQ-AEX-005).
- `NewApplierWithRegressionGate(manifestPath, baselinePath).WithOutcomeObserver(NewObserver(usageLogPath))` 배선 (REQ-AEX-008/009).
- 경로 resolve (project root 상대 canonical 경로, 절대경로 규칙 — REQ-AEX-009; design.md §B/§F.1 4개 경로).
- nil/empty sessions 전달 (REQ-AEX-011).
- RED: production Pipeline이 `AutoApply: true`로 구성됨 검증 (AC-AEX-007 non-vacuous production-wiring — tautology `_NeverPending` 금지), 4개 경로 canonical resolve 검증 (AC-AEX-009), 정상 apply → nil.
- AC 매핑: AC-AEX-005, AC-AEX-007, AC-AEX-008, AC-AEX-009, AC-AEX-011.

### M3 — Error surface → exit code mapping (RED→GREEN)
- `*ApplyPendingError` (AutoApply=true 하 invariant 위반) → exit 2 (REQ-AEX-014).
- rejection error (L1~L4) → exit 1 (REQ-AEX-012).
- `*ApplyRegressionError` → exit 1 (REQ-AEX-013).
- measurement-exec wrapped error → exit 2 (REQ-AEX-015).
- `errors.As`로 타입 분기 (errjoin 선례 — joined error walk 가능).
- RED: stub `SafetyEvaluator`(exported interface) 주입으로 각 에러 타입 강제 발생 → 기대 exit code. AC-AEX-014는 `DecisionPendingApproval` 반환 stub로 `*ApplyPendingError`를 강제 발생시켜 error-classification 분기가 exit 2 (INVARIANT VIOLATION)로 매핑함을 실증 (tautology "AutoApply=true는 Pending 안 냄" 금지 — design.md §F.1 T2).
- AC 매핑: AC-AEX-012(L1~L4 rejection exit 1), AC-AEX-013, AC-AEX-014, AC-AEX-015.

### M4 — `apply --execute` UX 위임 + payload-only 보존 (RED→GREEN)
- 기존 `newHarnessApplyCmd` (harness.go)에 `--execute` + `--id` flag 추가. flag 존재 시 `harnesscli.RunExecute`로 위임; 부재 시 기존 `runHarnessApply` payload-only (REQ-AEX-003).
- `newHarnessRouterCmd()`에 `harnesscli.NewExecuteCmd()` 등록 (REQ-AEX-002) — `apply --execute` UX와 `execute` sub-verb 둘 다 노출 (design.md §D 최종 결정에 따름).
- RED: `--execute` 부재 시 payload-only 동작 불변 (회귀 테스트), `--execute` 존재 시 Go 경로 호출.
- AC 매핑: AC-AEX-003, AC-AEX-004.

### M5 — first apply-outcome telemetry 검증 + MX 태그 + 정직 framing (RED→GREEN→REFACTOR)
- 통합 테스트: 정상 apply → `usage-log.jsonl`에 `apply_outcome` line 1건 생성 (REQ-AEX-010).
- MX 태그: `runExecute`/`NewExecuteCmd`에 `@MX:NOTE`(autoApply contract) + `@MX:WARN`+`@MX:REASON`(in-memory only, 디스크 불변). `Apply`/`NewPipeline` ANCHOR caller 목록 갱신.
- godoc + HONEST FRAMING 주석 (telemetry framing, "prevents regressions" 금지).
- AC 매핑: AC-AEX-006, AC-AEX-007, AC-AEX-010, AC-AEX-016.

## §G. Anti-Patterns (회피)

- **AP-1**: harness.yaml을 디스크에서 `auto_apply: true`로 쓰기 → C1 위반.
- **AP-2**: `apply` verb를 harness.go(package cli, boundary guard 밖)에 execute 로직 통째로 추가 → boundary guard 커버리지 약화 (design.md §D option (a) 단점). execute 로직은 `internal/cli/harness/`에 둔다.
- **AP-3**: regression_gate/outcome 알고리즘 "개선" → C2/C3 위반 (thin caller만).
- **AP-4**: CHANGELOG에 "회귀 방지" 문구 → §A.2 위반.
- **AP-5**: inert comment-filter grep idiom. `grep -n` 출력은 `N:` prefix를 붙이므로 column-0 형태 `^[[:space:]]*//` (및 `^\s*//`)는 **inert** (comment 줄이 `-v`를 통과해 살아남음 — SEC-HARDEN line 반복 + 본 SPEC plan-audit iter1 D2 실증). → 반드시 `-n`-aware 형태 `^[^:]*:[0-9]*:[[:space:]]*//` 사용 (AC-AEX-016과 통일) + exact-token + test-name trailing-`$` (acceptance.md §3 grep 규약 참조).

## §H. Cross-References

- spec.md §B (autoApply contract), §D (Exclusions), §E (MX)
- acceptance.md (16 AC matrix [MUST 14 / SHOULD 2, SSOT] + `-n`-aware grep 규약)
- design.md (verb shape decision + §E error→exit-code + §F sessions nil-safe + §F.1 [D4] 2-tier 테스트 seam 확정)
- research.md (코드 ground-truth + §F 해소된 설계 결정 + 선행 SPEC 계보)
