# SPEC Review Report: SPEC-FACTORY-MIXED-HOOK-001

Iteration: 2/3
Verdict: **PASS**
Overall Score: **1.00**
Plan Artifact Hash: `fc5cedb1d205b687fa1bc3e76571dd7d90438393b7752d29f7b931cdc25d345c`
Auditor Version: `plan-auditor/v1`

Reasoning context ignored per M1 Context Isolation. 이번 재감사는 iteration 1의 D1–D3 delta와 매 iteration 필수인 ordering consistency만 다시 판정했다.

## Claim

직전 감사의 blocking finding D1–D3가 모두 해소되었다. 계획은 이제 exact production `moai codex -f` launcher chain을 강제하고, 신규 OPS 6개 criterion의 current-tree RED-now를 제공하며, 기존 AC-FMH-001..015 전체를 post-M5 기준선에서 non-empty regression gate로 다시 실행하도록 요구한다.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: delta가 번호 체계를 변경하지 않았다. 기존 `REQ-FMH-001..012`와 OPS `REQ-FMH-OPS-001..008`의 연속성은 iteration 1 PASS를 유지한다.
- [PASS] MP-2 GEARS format compliance: `spec.md:113`과 addendum `:69-75`의 변경된 OPS 요구는 `Production proof SHALL...` 및 한국어 동등 규범형으로 표현된다.
- [PASS] MP-3 YAML frontmatter validity: strict lint exit 0; informational `OwnershipTransitionUnmeasured` 1건 외 error/warning은 없다.
- [N/A] MP-4 language neutrality: 단일 Go runtime/CLI 기능이다.
- [PASS] MP-5 D7 cross-SPEC reconciliation: delta에 retired/superseded/archived SPEC 참조가 추가되지 않았다.
- [PASS] MP-6 D8 cross-platform discipline: `spec.md`에서 `syscall` 검색 결과가 없다.
- [PASS] MP-7 clarification gate: `plan.md`와 `research.md`에서 `[NEEDS CLARIFICATION` 검색 결과가 없다.
- [PASS] MP-8 RED-now: `acceptance.md:43-55`가 OPS 6개 각각에 command, `<empty>` stdout, exit 1, current subject HEAD를 기록한다. 여섯 명령을 이 감사에서 그대로 재실행하여 모두 동일한 `<empty>`, exit 1을 관측했다. 기존 15개 원장은 `acceptance.md:66-86`에서 historical-only로 명확히 분리되었다.
- [PASS] MP-9 ordering consistency: canonical `M1→M2→M3→M4→M5`와 M5 내부 addendum `M1→M2→M3→M4`가 유지된다. 프롬프트 전 SessionStart 등록 후 lead MCP 조회, 그 후 worker 종료/dead 판정이라는 addendum 순서는 milestone 배치와 양립한다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 1.00 | exact production argv와 금지 우회, identity assertions가 단일 해석으로 고정됨 | `spec.md:113`; `plan.md:60`; addendum `:71,116-124` |
| Completeness | 1.00 | D1 live chain, D2 OPS RED-now, D3 기존 15개 regression surface 모두 존재 | `acceptance.md:43-116`; addendum `:154-180` |
| Testability | 1.00 | exact pass count, skip/NOT_RUN 거부, sentinel, process identity 및 cleanup assertions가 이진 판정 가능 | addendum `:158-180` |
| Traceability | 1.00 | OPS 6 AC와 8 REQ mapping은 유지되고 각 RED/GREEN selector가 criterion에 연결됨 | `acceptance.md:32-55` |

단순 평균: `1.00`. Tier L PASS 기준 `0.85`를 충족하며 must-pass 실패가 없다.

## Defects Found

No defects found in the D1–D3 delta.

## Regression Check

- D1 `OPS-LIVE-PRODUCTION-PATH` — **RESOLVED**: addendum `:71,116-124,128-152,164-172`가 한 built-tree `moai codex -f` lead와 두 `moai codex -f agent` worker의 exact argv/process tree를 요구한다. direct `codex`/`codex exec`, manual `MOAI_SESSION_PID`, direct `RegisterPeer`, DB/pre-seed, 별도 owner process를 금지하며, 실제 Codex owner PID/fingerprint와 SessionStart peer row 일치, lead-owned restarted MCP 조회, 전체 child cleanup을 gate에 포함한다.
- D2 `OPS-RED-NOW-MISSING` — **RESOLVED**: `acceptance.md:43-55`에 OPS 6 AC별 current-tree RED-now 행이 추가되었다. 실제 재실행 결과 여섯 selector 모두 stdout `<empty>`, exit 1이었다.
- D3 `LEGACY-15-REGRESSION-GATE-ABSENT` — **RESOLVED**: `plan.md:61`과 `acceptance.md:88-116`은 post-M5 tree에 귀속되는 10개 unit/benchmark + 5개 real-session gate를 요구한다. 과거 로그 상속을 금지하고 실행 불가 live row를 FAIL/GAP으로 유지한다.

## Evidence

### E1 — 시작 freeze hash

Command:

```text
shasum -a 256 spec.md plan.md acceptance.md progress.md operational-lane-status-addendum.md
```

Observed output:

```text
bd7dddae43556446533622757642f471f115e74d7eb02f50a01b4db9fb03dd18  spec.md
fb71e45e67036ce6cb143b863a07dcc0d2df32e2436ba5b000866addf7efbc1a  plan.md
4e8edd54f9132e4f4e4cd106cd811d9d123979bcff5c80e8fed83bbcce7f6fa5  acceptance.md
81e11d3716ee723a4f45d3efdb018b3c2cde8bef2758bc89056faa858861f607  progress.md
6ca613f572db0ee63079d2ab13606984b57f40138f00005cade29a63639bb748  operational-lane-status-addendum.md
```

Combined five-file hash: `fc5cedb1d205b687fa1bc3e76571dd7d90438393b7752d29f7b931cdc25d345c`.

### E2 — OPS 6개 RED-now 재실행

Each cited command was run verbatim in shape: `rg -n -F 'func <TestName>(' internal`.

Observed output:

```text
TestFactoryLiveOperationalRosterBeforePrompt: stdout=<empty>, exit=1
TestFactoryLaneRosterStateTruth: stdout=<empty>, exit=1
TestFactoryLaneRosterProjectIsolation: stdout=<empty>, exit=1
TestFactoryMsgStatusReadOnlyRoster: stdout=<empty>, exit=1
TestFactoryLeadNoticeUsesOperationalStatus: stdout=<empty>, exit=1
TestFactoryLiveOperationalLauncherChain: stdout=<empty>, exit=1
```

### E3 — 기존 10개 non-live/benchmark regression gate 실제 실행

Command: `acceptance.md:92-96`의 exact selector와 jq predicate를 사용하되 감사 소유권을 지키기 위해 로그만 `/tmp/t1074-iter2-m5-reg-unit.jsonl`로 바꾸어 실행했다.

Observed output:

```text
true
TestFactoryBrokerTrustBoundaries
TestFactoryCanonicalNamespaceAndIsolation
TestFactoryCrashRecoveryExplicitReceipt
TestFactoryDeadLetterAndLegacyIsolation
TestFactoryEnvelopeIdempotencyAndStaleAck
TestFactoryHookBenchmarkBudget
TestFactoryHookContextAndContinuationSafety
TestFactoryHookZeroTurnAndCapabilityTruth
TestFactoryRunSelectionAtomicSlotsAndArgv
TestFactorySessionGenerationOwnership
```

열 개의 고유 top-level PASS event가 존재했고 skip/`NOT_RUN`은 predicate를 통과하지 않았다.

### E4 — 기존 5개 live regression gate의 non-empty 구조

Selector lookup output:

```text
TestFactoryLiveCodexCodex rc=0 internal/cli/factory_live_test.go:22
TestFactoryLiveCodexClaude rc=0 internal/cli/factory_live_test.go:25
TestFactoryLiveClaudeCodex rc=0 internal/cli/factory_live_test.go:28
TestFactoryLiveClaudeClaudeCompletionSeparation rc=0 internal/cli/factory_live_test.go:31
TestFactoryLiveHookBoundaryIdleTruth rc=0 internal/cli/factory_live_test.go:35
```

`acceptance.md:98-116`은 각 test를 별도 `go test -run ^name$`로 실행하고 정확한 `Action=pass`를 요구하며 `skip`과 `NOT_RUN`을 거부한다. 다섯 결과 중 하나라도 실패하면 `set -e`로 gate 전체가 실패한다.

### E5 — strict lint 및 정적 건전성

Command:

```text
GOCACHE=/tmp/t1074-ops-plan-audit-iter2-cache go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json
git diff --check -- <five frozen artifacts>
```

Observed output:

```text
[{"severity":"info","code":"OwnershipTransitionUnmeasured",...}]
LINT_RC=0
DIFF_CHECK_RC=0
```

`git diff --check`은 stdout이 없었다.

## Baseline-attribution

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- Branch/HEAD: `WT-factory-mixed-hook@99d77dd25d3f2b1f6bae68d5fbccb5f8f4cf3bca`
- Freeze hash는 parent가 제공한 다섯 값과 시작 시 모두 일치했다.
- 이 감사가 작성한 repository 파일은 `plan-audit-ops-iter2.md` 하나뿐이다. 테스트 로그는 `/tmp`에만 기록했다.

## Gaps

- OPS production launcher live tests는 아직 구현 전 RED이므로 실행하지 않았다.
- 기존 5개 provider-backed live regression test도 이 plan delta 감사에서 재실행하지 않았다. 해당 gate의 selector 존재와 non-empty/skip 거부 구조만 검증했다.
- 따라서 이 PASS는 계획의 실행 가능성과 검증 계약에 대한 판정이며 runtime 구현 PASS가 아니다.

## Residual-risk

- 실제 PTY/process-tree 제어와 provider 가용성은 run phase에서만 확인된다. 새 gate는 실패 시 상속 PASS를 금지하므로 이 위험을 숨기지 않는다.
- `task_state=unknown`은 명시적 활동 관측이 없는 현재 범위의 의도된 truthful 결과다.

## Iteration history

- Iteration 1: FAIL 0.81 — D1 production path 미고정, D2 OPS RED-now 부재, D3 기존 15 AC 회귀 gate 부재.
- Iteration 2: PASS 1.00 — D1–D3 모두 해결; ordering regression 없음.

## Recommendation

계획은 run phase로 진행 가능하다. 구현 완료 판정은 addendum §6의 두 OPS live gate와 `acceptance.md`의 기존 15개 current-baseline regression gate가 실제로 통과한 뒤에만 내린다.
