---
id: SPEC-FACTORY-MIXED-HOOK-001
document: progress
status: completed
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-MIXED-HOOK-001

## §A Status

- Current status: `completed`.
- Card: `t1074` (`picked`).
- Worktree: `WT-factory-mixed-hook`.
- Plan baseline: `758314007` (local develop matched when authoring began).
- Implementation baseline: `8d8e101bf`; M1 implementation commit: `cb099897a`.
- Sync gate: independent audit `PASS` (85/100), blocking findings 0.
- Remaining non-blocking gates: repository-wide integration CI, Windows process-tree LIVE, and package-wide coverage confirmation.

## §B Plan artifacts

| Artifact | Status |
|---|---|
| spec.md | authored |
| plan.md | authored |
| acceptance.md | authored |
| research.md | authored |
| design.md | authored |
| progress.md | authored |

## §B.1 User-requested operational lane status plan change

- The user explicitly added operational lead visibility into real worker lane status. `operational-lane-status-addendum.md` is now normative through minimal cross-references in `spec.md`, `plan.md`, and `acceptance.md`; REQ-FMH-OPS-001..008 and AC-FMH-OPS-001..006 were initially `PENDING` and have final unit/LIVE evidence in the 2026-09-23 section below.
- The earlier independent plan-audit `PASS` predates this scope addition and does not cover the OPS requirements or criteria. An independent plan re-audit is required before the OPS implementation is treated as audit-ready; prior run-phase evidence below remains unchanged and cannot be credited to the new criteria.
- OPS delta plan audit iteration 1 returned `FAIL` at `0.81` with D1–D3: production launcher-chain proof, six criterion-level RED-now cells, and a fresh existing-15-AC regression gate were missing. The plan documents now address those three findings; independent re-audit remains required and no OPS PASS is claimed.

## §C Plan audit

### Iteration 1

- Verdict: FAIL, merge-blocking.
- Findings: non-canonical GEARS clauses, forbidden sibling status fields, abbreviated REQ references, missing executable RED ledger, live `NOT_RUN` false-pass, and missing exclusion structure.
- Resolution: all findings corrected; strict lint and mutant gates added.

### Iteration 2 — final

- Independent auditor verdict: `PASS`.
- Score: `0.94 / 1.00` (Tier L threshold `0.85`).
- Findings: empty; merge-blocking findings: none.
- Auditor-observed evidence:
  - `go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json` → `[]`, exit 0 (with isolated `GOCACHE`).
  - `git diff --check` → stdout empty, exit 0.
  - Fifteen RED-now named-test presence probes → stdout empty, exit 1 at tree `758314007d8c696ff1af377dc8cdc46d76368314`.
  - Synthetic live events: positive passes, `skip` and `NOT_RUN` fail.

## §D Run-phase evidence

- M1 committed as `cb099897a` (`feat(t1074): M1 bind canonical factory runs and peers`).
- Observed scoped tests at that commit: `internal/cli` → `ok ... 4.491s`; `internal/factorymsg` → `ok ... 1.310s`.
- M2/M3 committed as `6bde8412c` (`feat(t1074): deliver durable mixed factory messaging`).
- AC-FMH-001 through AC-FMH-009 exact JSON/JQ gates passed; logs are `../../reports/t1074/ac01.jsonl` through `ac09.jsonl`.
- M4 repair added a native Darwin process-start probe, one-pass linked-worktree canonicalization, explicit Codex MCP factory env forwarding, Codex launcher attribution scrubbing/backend export, tmux factory-env propagation, and distinct live owner identities.
- AC-FMH-010 now passes with separate real Codex model contexts, unique owner PIDs, no injected Codex session UUID, and two explicit acknowledgements. `ac10.jsonl` records `acknowledged=2` and the exact JQ gate printed `true`.
- AC-FMH-011 through AC-FMH-014 initially returned `Failed to authenticate: OAuth session expired and could not be refreshed` on the direct-Claude harness. All four now pass through the built `moai glm` path with real mixed/GLM nonce receipts and idle-boundary evidence.
- AC-FMH-015 now passes without changing its thresholds. The final matrix measured empty full-hook p50 `31.216917ms`, p95 `33.143167ms`, every inspection cell below 200ms, and bounded contention at `57.022917ms` with truthful `SQLITE_BUSY` degradation. The exact JQ gate printed `true`.

## §D.1 Codex cwd redesign decision

- Latest official source/docs and a local Codex `0.155.1` probe establish that interactive `/cd` can keep the visible conversation flow while rotating the physical session/thread UUID.
- `t1074` therefore owns the stable logical lane/current-endpoint broker seam only.
- `t1082` owns card worktree creation and `/cd` or headless `cwd` handoff through `BOUND`.
- `t1075` depends on t1082 and may wake only the current bound endpoint.

## §E Audit-ready signals

- Plan-phase: audit-ready; independent `PASS` at score `0.94`.
- Run-phase: M1/M2/M3 complete; all scoped AC-FMH-001..015 and AC-FMH-OPS-001..006 rows pass.
- Sync-phase: independent audit `PASS` (85/100), blocking findings 0; SPEC status is `completed`.

## §E.2 Run-phase evidence

| Criterion / invariant | Actual output | Status |
|---|---|---|
| AC-FMH-001..009 | Exact named Go JSON gates each produced final `true`; `ac01.jsonl`..`ac09.jsonl`. | PASS |
| AC-FMH-010 | Exact live gate printed `true`; separate Codex contexts completed nonce round trip with `acknowledged=2`. | PASS |
| AC-FMH-011 | `m5-reg-ac11-glm.jsonl`: built `moai glm` LIVE verified nonce `057d60face93a6325425c3330f1914de`, `acknowledged=2`, exact predicate `{"pass":true,"fail":false,"skip":false,"not_run":false}`. | PASS |
| AC-FMH-012 | `m5-reg-ac12-glm-rerun.jsonl`: GLM↔Codex nonce `8e74ffdc2458360dfdf1a9a3fc9c552f`, `acknowledged=2`. | PASS |
| AC-FMH-013 | `m5-reg-ac13-glm.jsonl`: GLM↔GLM nonce `8103c0a80c989e669bc79b9c6af17ea4`, `acknowledged=2`. | PASS |
| AC-FMH-014 | `m5-reg-ac14-glm.jsonl`: idle pending truth and next-turn receipt verified. | PASS |
| AC-FMH-015 | Exact benchmark gate printed `true`; empty p95 `33.143167ms`, contention `57.022917ms`, all inspection cells below 200ms. | PASS |
| Canonical/legacy isolation | Actual linked-worktree fixture converged; foreign run/project list/claim/read/receipt failed; legacy sentinel bytes unchanged. | PASS |
| Template parity | Project and embedded Stop hook entries are synchronous and parity test passed. | PASS |
| Repository-wide verdict | Scoped packages outside `hook`/`cli` passed. Separate full `hook` and `cli` attempts both reached the explicit 180-second local timeout; this is a GAP, not a pass or an attributed regression. | PENDING |

## §E.3 Run-phase Audit-Ready Signal

### OPS 추가 구현 — 2026-09-22, 부분 완료 / LIVE GAP

**Claim**

- 기존 `peers`와 `Status`에 slot 순서의 `lanes`를 추가했다. 실제 process identity probe로 endpoint를 판정하고, 별도 작업 관측이 없는 `task_state`는 `unknown`으로 유지한다. 새 저장소는 만들지 않았다.
- 실제 등록된 MCP status handler는 active run을 검증하고 기존 broker만 연다. 없는 broker를 생성하지 않으며 조회 오류를 빈 성공으로 바꾸지 않는다. 리더 SessionStart 안내는 실제 `factory_msg_status({"run_id":...})`로 연결된다.
- 단위/MCP 검증은 통과했지만 OPS 전체 완료나 AC-FMH-OPS-001/005/006 LIVE PASS를 주장하지 않는다.

**Evidence**

실행 위치는 아래 Baseline-attribution의 WT다. Go 검증은 한 invocation 안에서 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY`를 실행했으며 `GOCACHE=/tmp/t1074-go-cache`를 사용했다.

```text
go test ./internal/factorymsg -run '^TestFactoryLaneRoster(StateTruth|ProjectIsolation)$' -count=1 -timeout=90s
RED: roster_test.go:39: lanes=0, want 3
RED: roster_test.go:64: project/run leak: []
GREEN: ok  github.com/modu-ai/moai-adk/internal/factorymsg  0.512s

go test ./internal/cli -run '^(TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$' -count=1 -timeout=90s
RED: status initialized absent broker instead of preserving query failure
RED: notice has no read-only operational call
GREEN: ok  github.com/modu-ai/moai-adk/internal/cli  2.759s

go test -json ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$' -count=1 -timeout=90s > /tmp/t1074-lane-status-unit.jsonl
jq -se '([.[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$")))] | length)==4 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0' /tmp/t1074-lane-status-unit.jsonl
true
exit 0

MOAI_FACTORY_BENCH=1 go test ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryCanonicalNamespaceAndIsolation|TestFactoryRunSelectionAtomicSlotsAndArgv|TestFactorySessionGenerationOwnership|TestFactoryEnvelopeIdempotencyAndStaleAck|TestFactoryCrashRecoveryExplicitReceipt|TestFactoryDeadLetterAndLegacyIsolation|TestFactoryHookContextAndContinuationSafety|TestFactoryHookZeroTurnAndCapabilityTruth|TestFactoryBrokerTrustBoundaries|TestFactoryHookBenchmarkBudget)$' -count=1 -timeout=240s
ok  github.com/modu-ai/moai-adk/internal/factorymsg  1.839s
ok  github.com/modu-ai/moai-adk/internal/cli  1.953s
ok  github.com/modu-ai/moai-adk/internal/hook  5.236s

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryOperationalEvidenceRejectsUnownedResult)$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/factorymsg  1.618s
ok  github.com/modu-ai/moai-adk/internal/cli  4.183s
ok  github.com/modu-ai/moai-adk/internal/hook  1.638s [no tests to run]

go test -race ./internal/cli -run '^TestFactoryOperationalEvidenceRejectsUnownedResult$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  2.013s

GOCACHE=/tmp/t1074-go-cache go vet ./internal/factorymsg ./internal/cli ./internal/hook
stdout/stderr empty, exit 0

최종 parser 보강 후 같은 race 명령 재실행:
ok  github.com/modu-ai/moai-adk/internal/cli  1.946s

git diff --check
stdout/stderr empty, exit 0
```

위 hook 패키지의 `[no tests to run]`은 해당 패키지에 gate 이름이 없기 때문이다. 4개 필수 이름은 JSON gate에서 실제 pass 개수를 확인했다. SessionStart handler 연결 검증은 cli 패키지의 실제 in-process MCP 서버 테스트가 수행하며, 이를 authenticated LIVE 증거로 대체하지 않는다.

첫 production LIVE 시도:

```text
MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=operational-roster-before-prompt GOCACHE=/tmp/t1074-go-cache go test ./internal/cli -run '^TestFactoryLiveOperationalRosterBeforePrompt$' -count=1 -timeout=240s -v
BUILT_BINARY /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestFactoryLiveOperationalRosterBeforePrompt2447418386/002/moai sha256=2b7f31039803628c9fa1d97be1830b89d06eafd7c0fc2a661f099f30c7f044bf
PRODUCTION_ARGV /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestFactoryLiveOperationalRosterBeforePrompt2447418386/002/moai codex -f terminal_pid=60892
GAP: production SessionStart before prompt: want=1 got=0
Cause: failed to initialize state runtime at /Users/goos/.codex: failed to open state DB at /Users/goos/.codex/state_5.sqlite: error returned from database: (code: 8) attempt to write a readonly database
--- FAIL: TestFactoryLiveOperationalRosterBeforePrompt (35.03s)
FAIL  github.com/modu-ai/moai-adk/internal/cli  35.759s
exit 1

ps -axo pid=,ppid=,command= | rg '60892|TestFactoryLiveOperationalRosterBeforePrompt2447418386'
zsh:1: operation not permitted: ps
exit 1
```

**Baseline-attribution**

- WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`, branch `WT-factory-mixed-hook`, 시작/재측정 HEAD `19e9fc68d2e6b37222042b6bc7568dc0a9a52f23`의 미커밋 OPS 변경이다. t1078은 수정하지 않았다. 커밋·병합·queue 변경은 하지 않았다.
- 위 LIVE hash는 최초 진단 harness 실행 시 만든 임시 binary의 관측값이다. 후속 harness 수정 뒤 binary/hash나 LIVE 통과를 뜻하지 않는다.
- 후속 harness에는 primary 정확히 1 lead+2 agents의 프롬프트 전 owner 검증, lead-owned MCP 종료/재시작 PID 및 loaded binary 대조, 해당 session rollout의 `call_id`/`run_id` 연계 응답, 반복 조회 DB snapshot, phase 2 별도 production lead 격리, 실제 worker 종료 후 dead 확인, 지속적인 descendant 추적과 cleanup 검증을 추가했다. 이 경로는 코드/컴파일 수준이며 재LIVE하지 않았다.

**Gaps**

- 부모의 최종 정정: acceptance.md:152의 수동 등록 금지는 AC001/005/006의 실세션 PASS에 적용되며, addendum:124는 최초 production 세 행과 구분된 stale 보조 행을 허용한다. 따라서 AC002 stale는 `TestFactoryLaneRosterStateTruth`에서 실제 살아 있는 OS PID와 의도적으로 불일치한 fingerprint로 검증한다. 자연 PID 재사용을 기다리는 비결정적 LIVE 실패 조건은 제거했다. 이 단위 보조 행은 production AC001/005/006 증거에 절대 포함하지 않는다. production harness는 직접 RegisterPeer/DB seed/가짜 probe 없이 primary 1+2를 검증하고, 별도 production lead를 phase 2로 구분해 격리를 검증한다.
- 기존 5개 authenticated LIVE 회귀, 새 두 LIVE gate의 전체 실행, 새 MCP 재시작 응답·전체 cleanup의 실제 관측은 미완료다. 첫 실행은 프롬프트 전에 환경 오류로 끝났으며 작업 프롬프트·인증 입력·의도적인 모델 payload 전송을 하지 않았다.
- 최종 harness 재실행은 부모의 별도 승인 대상이다. sandbox 밖에서는 사용자 Codex state DB와 session rollout이 갱신되고 `ps` 및 macOS `lsof`로 process/binary 정보를 조회한다. 프롬프트 단계에서는 synthetic fixture의 run ID/endpoint 정보가 인증된 모델 서비스로 전송될 수 있다. 사용자 credentials/DB를 복사하거나 수정해 우회하지 않았다.

**Residual-risk**

- 추가 5-name scoped coverage는 `factorymsg 22.3%`, `cli 5.9%`, 함수별 `handleFactoryMsgStatus 85.7%`, `Status 52.4%`, `laneRoster 86.7%`였다. 이후 위 기존 10-name 회귀와 5-name OPS/parser 검증을 합쳐 `MOAI_HOME=/tmp/t1074-final-regression-home MOAI_FACTORY_BENCH=1 GOCACHE=/tmp/t1074-go-cache go test ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryCanonicalNamespaceAndIsolation|TestFactoryRunSelectionAtomicSlotsAndArgv|TestFactorySessionGenerationOwnership|TestFactoryEnvelopeIdempotencyAndStaleAck|TestFactoryCrashRecoveryExplicitReceipt|TestFactoryDeadLetterAndLegacyIsolation|TestFactoryHookContextAndContinuationSafety|TestFactoryHookZeroTurnAndCapabilityTruth|TestFactoryBrokerTrustBoundaries|TestFactoryHookBenchmarkBudget|TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryOperationalEvidenceRejectsUnownedResult)$' -coverprofile=/tmp/t1074-combined-cover.out -count=1 -timeout=240s`를 실행했다. 실제 출력: `factorymsg 1.861s coverage: 61.8%`, `cli 3.932s coverage: 6.0%`, `hook 5.209s coverage: 1.2%`, exit 0. `GOCACHE=/tmp/t1074-go-cache go tool cover -func=/tmp/t1074-combined-cover.out` 출력: `handleFactoryMsgStatus 85.7%`, `Status 85.7%`, `laneRoster 86.7%`. 전체 패키지 coverage 85%를 주장하지 않는다. 최초 cover 표시 명령은 기본 Go cache 읽기 권한 오류였으며 명시적 GOCACHE로 재측정했다. combined 첫 실행은 MOAI_HOME을 빠뜨려 기존 benchmark의 `chmod /Users/goos/.moai: operation not permitted`로 실패했으며 위 isolated home 재실행 결과와 구분한다.
- Codex TUI의 신뢰/인증 화면, 실제 MCP 자동 재연결 가능 여부, rollout 형식, process owner 관계와 descendant cleanup은 실제 재실행으로 확인해야 한다. UI나 재연결이 기대와 다르면 테스트가 실패하며 성공 sentinel을 대체하지 않는다.
- phase 2 production fixture는 별도 프로젝트에서 자연 생성된 run을 사용한다. 같은 run 문자열의 프로젝트 격리는 단위 테스트로만 관측했다.
- 승인 후 실행 명령은 위 LIVE 명령과 `MOAI_FACTORY_LIVE_CASE=operational-launcher-chain ... -run '^TestFactoryLiveOperationalLauncherChain$'`이며, 반드시 이번 WT에서 환경 scrub과 함께 실행한다. 현재 전체 run 상태는 불완료다.

### F1 audit repair — 빈 call_id의 소유 관계 거부

**Claim / Baseline-attribution:** `WT-factory-mixed-hook@19e9fc68d2e6b37222042b6bc7568dc0a9a52f23`의 기존 미커밋 OPS 변경에서 F1만 수정했다. 빈 `function_call.call_id`는 calls에 등록하지 않고 빈 output ID도 인정하지 않는다. 양쪽 빈 ID 및 한쪽만 빈 ID 회귀 3개를 먼저 추가했다. 구현은 조건 2개로 충분하여 추가 refactor는 하지 않았다. 이번 수정 파일은 두 parser test 파일과 이 progress 문서뿐이며 커밋하지 않았다.

**Evidence:** 아래 모든 Go 명령은 WT에서 `unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY &&` 뒤에 실행했다.

```text
MOAI_HOME=/tmp/t1074-f1-home GOCACHE=/tmp/t1074-go-cache go test ./internal/cli -run '^TestFactoryOperationalEvidenceRejectsUnownedResult$' -count=1 -timeout=90s
--- FAIL: TestFactoryOperationalEvidenceRejectsUnownedResult (0.00s)
    factory_operational_evidence_test.go:28: got true want false: {"type":"response_item","payload":{"type":"function_call","name":"mcp__moai__factory_msg_status","arguments":"{\"run_id\":\"r\"}","call_id":""}}
        {"type":"response_item","payload":{"type":"function_call_output","call_id":"","output":"{\"lanes\":[{\"slot\":\"lead\"}]}"}}
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 0.735s
FAIL
exit 1

MOAI_HOME=/tmp/t1074-f1-home GOCACHE=/tmp/t1074-go-cache go test -race ./internal/cli -run '^TestFactoryOperationalEvidenceRejectsUnownedResult$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  1.941s
exit 0

MOAI_HOME=/tmp/t1074-f1-gate-home GOCACHE=/tmp/t1074-go-cache go test -json ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$' -count=1 -timeout=90s > /tmp/t1074-f1-unit.jsonl
jq -se '([.[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$")))] | length)==4 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0' /tmp/t1074-f1-unit.jsonl
true
exit 0

git diff --check
stdout/stderr empty, exit 0
```

**Gaps / Residual-risk:** F1의 빈 ID 연결만 검증했다. authenticated LIVE는 재실행하지 않았고 앞서 기록한 LIVE GAP은 그대로다.

```yaml
run_complete_at: null
run_commit_sha: afcee4757
run_status: fail
ac_pass_count: 11
ac_fail_count: 4
preserve_list_post_run_count: 1
l44_pre_commit_fetch: not-run
l44_post_push_fetch: not-applicable-no-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-run
  reason: "M4 live blockers and benchmark regression leave the run incomplete"
total_run_phase_files: 28
m1_to_mN_commit_strategy: "M1 cb099897a; M2/M3 6bde8412c; M4 harness 45285bf1b; M4 repair/evidence afcee4757"
```

### UserPromptSubmit pending rebind — 2026-09-22

**Claim**

- 비어 있지 않은 정상 `UserPromptSubmit`은 inbox 조회 전에 launcher가 만든 `launch_pending` endpoint를 실제 session UUID로 rebind한다. empty/whitespace prompt는 bind를 호출하지 않는다.
- 이미 동일 owner PID/process-start, run, backend, role, slot에 bound된 endpoint의 후속 prompt는 peer row를 다시 쓰지 않으면서 inbox 조회를 계속한다.
- LIVE harness의 후속 prompt 완료 판정은 TUI 문자열을 보지 않고 해당 session rollout의 `response_item` 중 assistant `output_text`가 예상 문자열과 정확히 같은 경우만 인정한다.

**Evidence**

RED는 production 변경 전에 아래 exact selector로 관측했다.

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/t1074-up-red-cache go test -race ./internal/hook -run '^(TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=90s
--- FAIL: TestFactoryUserPromptSubmitRebindsLaunchPendingPeer
    factory_messages_test.go:323: factory endpoint is launch-pending
--- FAIL: TestFactoryBoundUserPromptSubmitDoesNotRewritePeer
    factory_messages_test.go:341: factory endpoint is launch-pending
FAIL
FAIL github.com/modu-ai/moai-adk/internal/hook
```

첫 bind와 동일 hook invocation에서 inbox가 bind 뒤에 조회됨을 구분하도록 anticipated actual session/generation 대상 test-only envelope를 준비한 뒤 exact GREEN을 재관측했다. 이 fixture는 production LIVE broker 우회 증거로 사용하지 않는다.

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/t1074-up-order-cache go test -race ./internal/hook -run '^(TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/hook  3.029s
exit 0
```

8-selector JSON/JQ gate는 fresh `MOAI_HOME=$(mktemp -d /tmp/t1074-eight.XXXXXX)`에서 실행했다.

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=$T1074_HOME GOCACHE=/tmp/t1074-eight-cache go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=120s -json | jq -s '([.[] | select(.Test != null and .Action == "pass") | .Test] | unique | length) == 8 and ([.[] | select(.Test != null and (.Action == "fail" or .Action == "skip"))] | length) == 0'
true
exit 0
```

관련 hook race와 LIVE evidence parser/fixture selector도 fresh isolated home에서 통과했다.

```text
go test -race ./internal/cli ./internal/hook -run '^(TestFactoryOperationalEvidenceRejectsUnownedResult|TestOperationalAssistantOutputRequiresOwnedExactText|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer|TestFactoryHookContextAndContinuationSafety|TestFactoryHookZeroTurnAndCapabilityTruth)$' -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/cli  1.983s
ok  github.com/modu-ai/moai-adk/internal/hook  4.722s

go test -race ./internal/cli -run '^(TestFactoryOperationalFixtureUsesProductionInit|TestFactoryOperationalTrustBootstrap|TestFactoryOperationalHookTrustOnce|TestFactoryOperationalTerminalHasGeometry|TestFactoryOperationalCleanupWaitsForExit|TestFactoryOperationalDirectoryTrustOnce|TestFactoryOperationalEvidenceRejectsUnownedResult|TestOperationalAssistantOutputRequiresOwnedExactText)$' -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/cli  16.443s

git diff --check
stdout/stderr empty, exit 0
```

**Baseline-attribution**

- 실행 위치: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`.
- branch/HEAD: `WT-factory-mixed-hook@8c5d9be99`와 그 위 사용자 승인 미커밋 t1074 변경.
- 모든 Go 검증은 같은 compound invocation 안에서 provider 환경을 scrub하고 새 `MOAI_HOME`을 사용했다. repository-wide suite, commit, push는 수행하지 않았다.

**Gaps**

- 수정된 UserPromptSubmit 계약의 authenticated `TestFactoryLiveOperationalLauncherChain`은 이 실행에서 `NOT_RUN`이다. 따라서 `USERPROMPT_REBIND_OK lead,agent-1,agent-2`, `BOUND_PROMPT_NO_REWRITE_OK lead,agent-1,agent-2`, MCP restart/isolation/dead-owner/cleanup의 새 계약 LIVE PASS를 주장하지 않는다.
- LIVE 실행은 실제 Codex 세션/rollout과 인증된 모델 turn을 사용하므로 권한 있는 orchestrator가 별도로 수행한다.

**Residual-risk**

- rollout 형식 또는 Codex hook lifecycle이 실제 설치 버전에서 달라지면 LIVE parser가 성공 sentinel 없이 실패한다. parser 단위 테스트는 user echo, 부분 문자열, 비-`output_text`를 거부하지만 실제 authenticated rollout을 대체하지 않는다.
- repository-wide CI verdict는 integration branch CI 소유이며 현재 `PENDING`이다.

### M5 및 operational final evidence — 2026-09-23

**Claim**

- 기준 HEAD `8c5d9be993ad45d5c91ae8a95af904b27c935599`의 사용자 승인 t1074 작업 트리에서 atomic pending bind/authoritative turn 정책과 exact 8-selector gate가 race 포함 PASS했다.
- LIVE12..14는 실제 MCP call이 없어서가 아니라 당시 fail-closed evidence parser가 current Codex output shape와 turn 간 누적을 해석하지 못해 실패했다. LIVE15는 trust TUI flake로 모델 chain 이전에 실패했다. 이 진단 실패를 최종 PASS로 소급하지 않는다.
- 최종 LIVE16은 실제 production `moai codex -f` lead와 agent 두 개를 실행해 prompt 전 pending, 정상 UserPromptSubmit bind, 후속 prompt 무쓰기, lead-owned MCP, project isolation, dead owner, cleanup, no bypass를 모두 관측하고 PASS했다.
- M5 AC-FMH-010은 nonce `75bb84b10c8ee1397752a7c03a3681ed`, `acknowledged=2`로 PASS했다. AC-FMH-011..013도 built `moai glm` 경로에서 각각 nonce와 `acknowledged=2`로 PASS했고 AC-FMH-014 idle-boundary도 PASS했다. 이후 independent sync audit도 `PASS`(85/100, blocking finding 0)로 닫혔다.

**Evidence**

Atomic race/unit gate:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/t1074-evidence-atomic-cache go test -race ./internal/factorymsg ./internal/hook -run '^(TestBindLaunchPendingCannotOverwriteConcurrentAuthoritativePeer|TestBindLaunchPendingRejectsOwnerMismatch|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=90s -v
--- PASS: TestBindLaunchPendingCannotOverwriteConcurrentAuthoritativePeer (0.78s)
--- PASS: TestBindLaunchPendingRejectsOwnerMismatch (0.06s)
PASS
ok  github.com/modu-ai/moai-adk/internal/factorymsg  2.171s
--- PASS: TestFactorySessionStartRebindsLaunchPendingPeer (0.45s)
--- PASS: TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding (0.94s)
--- PASS: TestFactoryUserPromptSubmitRebindsLaunchPendingPeer (0.64s)
--- PASS: TestFactoryBoundUserPromptSubmitDoesNotRewritePeer (0.66s)
PASS
ok  github.com/modu-ai/moai-adk/internal/hook  4.219s
```

SessionStart operational notice fixture의 구계약 회귀 RED와 실제 launcher 선행 조건을 반영한 GREEN:

```text
go test ./internal/cli -run '^TestFactoryLeadNoticeUsesOperationalStatus$' -count=1 -timeout=30s -v
mcp_factory_msg_test.go:237: SessionStart peer missing: {"Acknowledged":0,"Capability":"hook-boundary","Claimed":0,"DeadLetter":0,"NextDelivery":"pending-until-next-turn","Pending":0,"lanes":[]}
--- FAIL: TestFactoryLeadNoticeUsesOperationalStatus (1.03s)

--- PASS: TestFactoryLeadNoticeUsesOperationalStatus (1.06s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  1.960s
```

동일 8-selector non-race/race JSON gate는 각각 `true`였다.

```text
go test ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=120s -json | jq -s '<exact eight-pass/no-fail/no-skip expression>'
true

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '<same exact eight selectors>' -count=1 -timeout=120s -json | jq -s '<exact eight-pass/no-fail/no-skip expression>'
true
```

M5 unit artifact:

```text
jq -s 'any(.[]; .Action=="pass" and .Test!=null)' .moai/reports/t1074/m5-reg-unit.jsonl
true
```

M5 AC-FMH-010 `.moai/reports/t1074/m5-reg-ac10.jsonl`:

```text
live separate-context nonce/receipt verified: case=codex-codex nonce=75bb84b10c8ee1397752a7c03a3681ed acknowledged=2
--- PASS: TestFactoryLiveCodexCodex (130.32s)
```

M5 AC-FMH-011..014의 각 artifact `m5-reg-ac11.jsonl`..`m5-reg-ac14.jsonl`:

```text
Failed to authenticate: OAuth session expired and could not be refreshed
```

Operational 진단 chain:

```text
LIVE12 /tmp/t1074-root-lane-status-chain-live12.jsonl:
lead session rollout has no attributable factory_msg_status call/result
--- FAIL: TestFactoryLiveOperationalLauncherChain (87.80s)

LIVE13 /tmp/t1074-root-lane-status-chain-live13.jsonl:
lead session rollout has no attributable factory_msg_status call/result
--- FAIL: TestFactoryLiveOperationalLauncherChain (86.15s)

LIVE14 /tmp/t1074-root-lane-status-chain-live14.jsonl:
MCP_RESTART_OK old=96980 new=97161
LEAD_MCP_ROSTER_OK lead,agent-1,agent-2
lead session rollout has no attributable factory_msg_status call/result
--- FAIL: TestFactoryLiveOperationalLauncherChain (102.53s)

LIVE15 /tmp/t1074-root-lane-status-chain-live15.jsonl:
trust bootstrap did not reach post-trust idle terminal
Hooks need review
--- FAIL: TestFactoryLiveOperationalLauncherChain (50.89s)
```

LIVE12/13은 current `custom_tool_call_output.output`의 block/JSON-array shape를 거부했고, LIVE14는 첫 turn 2 calls와 다음 turn 2 calls를 누적하지 못했다. 각 실패 뒤 parser는 attributable call/output/completed ownership을 유지한 채 해당 shape와 distinct-turn 누적만 수용하도록 좁게 보강됐다.

최종 `/tmp/t1074-root-lane-status-chain-live16.jsonl`:

```text
PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent
LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2
NO_SESSION_UUID_BEFORE_TURN_OK
NO_BYPASS_OK
PROMPT_COMPOSER_READY_OK lead,agent-1,agent-2
USERPROMPT_REBIND_OK lead,agent-1,agent-2
BOUND_ROSTER_OK lead,agent-1,agent-2
BOUND_PROMPT_NO_REWRITE_OK lead,agent-1,agent-2
MCP_RESTART_OK old=30634 new=30792
LEAD_MCP_ROSTER_OK lead,agent-1,agent-2
PROJECT_ISOLATION_OK primary=001-eef421b0 secondary=006-c5096921 foreign_session=
DEAD_OWNER_OK
NO_SYNTHETIC_TURN_OR_BYPASS_OK
CLEANUP_OK
--- PASS: TestFactoryLiveOperationalLauncherChain (74.22s)
ok  github.com/modu-ai/moai-adk/internal/cli  74.976s
```

Exact LIVE predicate:

```text
jq -s 'any(.[]; .Action=="pass" and .Test=="TestFactoryLiveOperationalLauncherChain") and ([.[]|select(.Action=="skip")]|length==0) and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length==0)' /tmp/t1074-root-lane-status-chain-live16.jsonl
true
```

최종 criterion 판정:

| Criterion | Actual output | Status |
|---|---|---|
| AC-FMH-001..009 | `m5-reg-unit.jsonl` selector artifact; predicate `true`. | PASS |
| AC-FMH-010 | nonce `75bb84b10c8ee1397752a7c03a3681ed`, `acknowledged=2`. | PASS |
| AC-FMH-011 | `m5-reg-ac11-glm.jsonl`: built `moai glm` rerun, real Codex↔GLM nonce `057d60face93a6325425c3330f1914de`, `acknowledged=2`. | PASS |
| AC-FMH-012 | `m5-reg-ac12-glm-rerun.jsonl`: GLM↔Codex nonce round trip, `acknowledged=2`. | PASS |
| AC-FMH-013 | `m5-reg-ac13-glm.jsonl`: GLM↔GLM nonce round trip, `acknowledged=2`. | PASS |
| AC-FMH-014 | `m5-reg-ac14-glm.jsonl`: idle pending truth and next-turn receipt. | PASS |
| AC-FMH-015 | M5 unit artifact의 benchmark selector PASS. | PASS |
| AC-FMH-OPS-001..006 | atomic unit/race와 LIVE16 exact predicate/sentinels. | PASS |

**Baseline-attribution**

- 실행 위치: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`, branch `WT-factory-mixed-hook`.
- 기준 HEAD: `8c5d9be993ad45d5c91ae8a95af904b27c935599`; 결과는 그 기준과 현재 사용자 승인 t1074 미커밋 변경에 귀속한다.
- LIVE12..16은 해당 `/tmp/t1074-root-lane-status-chain-live*.jsonl`에, M5는 `.moai/reports/t1074/m5-reg-*.jsonl`에 각각 귀속한다.

**Gaps**

- provider-backed live rows와 independent sync audit은 모두 PASS했다. integration-branch CI만 범위 밖 `PENDING`이다.
- LIVE16 원본은 `/tmp/t1074-root-lane-status-chain-live16.jsonl`에 있으며 tracked report에는 판정에 필요한 sentinel만 전사했다.
- Windows 실제 process-tree LIVE와 repository-wide integration-branch CI verdict는 관측하지 않았다. CI verdict는 현재 `PENDING`이다.

**Residual-risk**

- current Codex rollout shape가 다시 바뀌면 fail-closed evidence parser가 거부할 수 있다. LIVE16은 현재 관측 shape만 증명한다.
- trust TUI는 외부 UI 상태에 민감하다. LIVE15 flake 뒤 LIVE16이 PASS했지만 UI 변화로 재발할 수 있다.
- provider 및 client UI 변화는 live 재현성에 영향을 줄 수 있으므로 integration CI와 후속 live에서 회귀할 수 있다.

### POSIX launch 실패 시 provisional rollback — 2026-09-22

**Claim**

- POSIX direct 경로는 cwd 진입에 성공한 뒤에만 launch-pending row를 등록한다. 따라서 `chdir` 실패는 broker row를 만들지 않는다.
- `syscall.Exec` 실패 시 launcher가 방금 돌려받은 slot/project/run/session/generation/PID/process-start 전체가 동일한 launch-pending row만 삭제한다. hook rebind 또는 새 generation으로 바뀐 row는 삭제하지 않는다.
- 이 보강은 LIVE5 절의 “실패 row를 자동 삭제하지 않는다”는 residual-risk를 대체한다. 정상 exec 성공과 Windows Start/Wait 경로는 변경하지 않았다.

**Evidence**

rollback API 부재 및 실패 row 잔존 RED:

```text
internal/factorymsg/launch_pending_rollback_test.go:29:23: s.RollbackLaunchPending undefined
internal/factorymsg/launch_pending_rollback_test.go:41:23: s.RollbackLaunchPending undefined
internal/factorymsg/launch_pending_rollback_test.go:56:23: s.RollbackLaunchPending undefined
FAIL github.com/modu-ai/moai-adk/internal/factorymsg [build failed]
--- FAIL: TestCodexDirectPOSIXChdirFailureDoesNotRegisterPending
    codex_launcher_exec_posix_test.go:154: error=register factory launch-pending endpoint: NO_ACTIVE_FACTORY
--- FAIL: TestCodexDirectPOSIXExecFailureRollsBackExactPending
    codex_launcher_exec_posix_test.go:185: exec failure left pending rows: [{Slot:lead ... BindingState:launch_pending ...}]
FAIL github.com/modu-ai/moai-adk/internal/cli
```

exact rollback GREEN:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=$(mktemp -d /tmp/t1074-rollback-green.XXXXXX) GOCACHE=/tmp/t1074-rollback-green-cache go test -race ./internal/factorymsg ./internal/cli -run '^(TestRollbackLaunchPendingDeletesOnlyExactProvisionalOwner|TestCodexDirectPOSIXChdirFailureDoesNotRegisterPending|TestCodexDirectPOSIXExecFailureRollsBackExactPending)$' -count=1 -timeout=60s
ok  github.com/modu-ai/moai-adk/internal/factorymsg  1.542s
ok  github.com/modu-ai/moai-adk/internal/cli  2.898s
```

POSIX owner-preservation 회귀와 rollback 확대 selector:

```text
go test -race ./internal/factorymsg ./internal/cli -run '^(TestRollbackLaunchPendingDeletesOnlyExactProvisionalOwner|TestCodexDirectPOSIXExecPreservesFactoryOwner|TestCodexDirectPOSIXChdirFailureDoesNotRegisterPending|TestCodexDirectPOSIXExecFailureRollsBackExactPending|TestCodexDirectLaunch|TestCodexLauncherRegistersLaunchPending)$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/factorymsg  1.371s
ok  github.com/modu-ai/moai-adk/internal/cli  4.543s
```

8-selector gate, Windows compile, whitespace check:

```text
go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '<exact eight selectors>' -count=1 -timeout=120s -json | jq -s '<eight-pass/no-fail/no-skip expression>'
true

GOOS=windows GOARCH=amd64 GOCACHE=/tmp/t1074-windows-rollback-cache go test -c ./internal/cli -o /tmp/t1074-cli-windows-rollback.test.exe
stdout/stderr empty, exit 0

git diff --check
stdout/stderr empty, exit 0
```

**Baseline-attribution**

- RED/GREEN과 회귀 검증은 `WT-factory-mixed-hook@8c5d9be99` 및 현재 사용자 승인 t1074 미커밋 변경에서, provider 환경 scrub과 fresh `MOAI_HOME`으로 관측했다.
- mismatched process-start와 이미 bound된 row가 보존되고 exact pending만 삭제되는 것은 `TestRollbackLaunchPendingDeletesOnlyExactProvisionalOwner`가 동일 broker에서 직접 비교했다.

**Gaps**

- rollback 보강 뒤 authenticated LIVE는 이 실행에서 재실행하지 않았다. 실제 세 lane rebind와 cleanup sentinel은 orchestrator 재실행 대상이다.
- Windows는 compile만 관측했고 Windows 실제 process lifecycle은 실행하지 않았다.

**Residual-risk**

- exec 실패와 동시에 broker가 쓰기 불능이면 rollback도 실패할 수 있다. 이 경우 반환 error에 exec/rollback 오류가 함께 포함되고, exact-owner 조건 때문에 다른 generation이나 bound endpoint를 지우지는 않는다.
- repository-wide CI verdict는 integration branch CI 소유이며 현재 `PENDING`이다.

### Authenticated LIVE 재현과 prompt readiness 보강 — 2026-09-22

**Claim**

- 재현 로그에서 세 production launcher와 세 `launch_pending` endpoint까지는 실제로 관측됐다. 실패 시점의 `FACTORY_READY_1`은 assistant 응답이 아니라 TUI의 `›` composer에 남아 있었고, 동일 로그에는 assistant/rollout 성공 증거가 없었다. 따라서 이 실행은 `UserPromptSubmit` product 결함을 입증하지 않으며, prompt가 ready 전 입력되어 Enter가 제출로 처리되지 않은 harness timing 결함으로 귀속한다.
- LIVE harness는 세 terminal 모두에서 loading/trust 화면보다 뒤에 나타난 `Ask Codex to do anything` composer를 확인한 후에만 첫 prompt를 보낸다. readiness는 세 terminal을 하나의 45초 유한 deadline으로 함께 기다린다. 느린 최초 trust bootstrap도 기존 30초 경계 대신 45초의 유한 deadline을 사용한다.
- 실제 hook subprocess 실행이 관측되지 않았으므로 이번 보강에서는 production hook/launcher 코드를 추가 변경하지 않았다.

**Evidence**

실제 재현 로그 `/tmp/t1074-root-lane-status-chain-live2.jsonl`의 구조적 출력은 다음과 같다. prompt 본문 외 민감 payload는 기록하지 않았다.

```text
TRUST_BOOTSTRAP_OK ... no_factory_state=true
PRODUCTION_ARGV .../moai codex -f terminal_pid=8316
PRODUCTION_ARGV .../moai codex -f agent terminal_pid=8451
PRODUCTION_ARGV .../moai codex -f agent terminal_pid=8711
PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent
LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2
NO_SESSION_UUID_BEFORE_TURN_OK
NO_BYPASS_OK
UserPromptSubmit did not rebind terminal 0
... model: loading ...
... › Reply with exactly FACTORY_READY_1 and do not call any tool. ...
CLEANUP_OK terminal_pid=8711 descendants=16
CLEANUP_OK terminal_pid=8451 descendants=16
CLEANUP_OK terminal_pid=8316 descendants=17
```

readiness predicate와 지연된 실제 subprocess output 경계 검증:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=$T1074_HOME GOCACHE=/tmp/t1074-readiness-green-cache go test -race ./internal/cli -run '^(TestFactoryOperationalPromptReadinessGate|TestFactoryOperationalPromptReadinessProcessBoundary)$' -count=1 -timeout=30s -v
--- PASS: TestFactoryOperationalPromptReadinessGate (0.00s)
--- PASS: TestFactoryOperationalPromptReadinessProcessBoundary (0.21s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  2.442s
```

관련 fixture/parser 묶음과 기존 8-selector gate도 재검증했다.

```text
go test -race ./internal/cli -run '^(TestFactoryOperationalPromptReadinessGate|TestFactoryOperationalTrustBootstrap|TestFactoryOperationalFixtureUsesProductionInit|TestFactoryOperationalHookTrustOnce|TestFactoryOperationalTerminalHasGeometry|TestFactoryOperationalCleanupWaitsForExit|TestFactoryOperationalDirectoryTrustOnce|TestFactoryOperationalEvidenceRejectsUnownedResult|TestOperationalAssistantOutputRequiresOwnedExactText)$' -count=1 -timeout=120s
ok  github.com/modu-ai/moai-adk/internal/cli  16.423s

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=120s -json | jq -s '<exact eight-pass/no-fail/no-skip expression>'
true
exit 0
```

**Baseline-attribution**

- LIVE 재현은 orchestrator가 같은 WT의 수정 전 readiness harness로 실행한 `/tmp/t1074-root-lane-status-chain-live2.jsonl`에 귀속한다.
- 수정/검증 위치는 `WT-factory-mixed-hook@8c5d9be99`와 사용자 승인 t1074 미커밋 변경이다. provider 환경 scrub 및 fresh `MOAI_HOME`을 사용했다.

**Gaps**

- readiness 보강 뒤 authenticated LIVE는 이 실행에서 재실행하지 않았다. 새 `PROMPT_COMPOSER_READY_OK lead,agent-1,agent-2` 이후 실제 UserPromptSubmit bind와 나머지 sentinel은 orchestrator 재실행 대상이다.
- hook subprocess가 실제 실행된 상태의 input/env/owner 결과는 실패 로그에 존재하지 않는다. 따라서 합성 중간-process owner mismatch를 실제 Codex 결함으로 일반화하거나 production 변경 근거로 사용하지 않았다.

**Residual-risk**

- 설치 Codex가 composer 문구를 바꾸면 readiness gate는 prompt를 조기 전송하지 않고 deadline failure로 끝난다.
- readiness 이후 실제 hook owner가 launcher pending PID와 다르다는 증거가 새 LIVE에서 나타나면 별도 process-owner 수정이 필요하다. 현재는 그 결함을 주장할 관측이 없다.
- repository-wide CI verdict는 integration branch CI 소유이며 현재 `PENDING`이다.

### LIVE3 입력 handoff 재현과 split-submit 보강 — 2026-09-22

**Claim**

- readiness 1차 보강 뒤 authenticated LIVE3에서도 `PROMPT_COMPOSER_READY_OK` 직후 prompt text가 composer에 남고 60초 동안 bind/rollout이 생기지 않았다. 따라서 단일 `prompt+CR` write가 ready 화면 전환 경계에서 Enter를 유실할 수 있는 harness 결함으로 범위를 더 좁혔다.
- 이 단계의 harness는 composer readiness를 500ms 동안 연속 관측한 뒤 prompt text만 먼저 write하고 raw stream에서 exact `› <prompt>` 재구성을 시도했다. LIVE4에서 Codex TUI가 cursor-diff만 내보내 이 가정이 성립하지 않음이 확인되었고, 이 방식은 아래 LIVE4 보강으로 대체됐다.
- 실제 hook 실행 증거가 여전히 없으므로 product hook/launcher 구현은 추가 변경하지 않았다.

**Evidence**

LIVE3 로그 `/tmp/t1074-root-lane-status-chain-live3.jsonl`:

```text
LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2
NO_BYPASS_OK
PROMPT_COMPOSER_READY_OK lead,agent-1,agent-2
UserPromptSubmit did not rebind terminal 0
... › Reply with exactly FACTORY_READY_1 and do not call any tool. ...
CLEANUP_OK terminal_pid=13997 descendants=16
CLEANUP_OK terminal_pid=13803 descendants=16
CLEANUP_OK terminal_pid=13691 descendants=17
```

combined-write mutant RED:

```text
go test ./internal/cli -run '^TestFactoryOperationalPromptSubmitWaitsForRenderedComposer$' -count=1 -timeout=30s -v
=== RUN   TestFactoryOperationalPromptSubmitWaitsForRenderedComposer
    factory_operational_fixture_test.go:168: write prompt text: first write="Reply with exactly FACTORY_READY_1 and do not call any tool.\r"
--- FAIL: TestFactoryOperationalPromptSubmitWaitsForRenderedComposer (0.00s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 0.727s
```

split-write 및 안정화 GREEN:

```text
go test -race ./internal/cli -run '^(TestFactoryOperationalPromptReadinessGate|TestFactoryOperationalPromptReadinessProcessBoundary|TestFactoryOperationalPromptSubmitWaitsForRenderedComposer|TestFactoryOperationalTrustBootstrap|TestFactoryOperationalHookTrustOnce)$' -count=1 -timeout=60s -v
--- PASS: TestFactoryOperationalPromptReadinessProcessBoundary (0.71s)
--- PASS: TestFactoryOperationalPromptSubmitWaitsForRenderedComposer (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  2.837s

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=120s -json | jq -s '<exact eight-pass/no-fail/no-skip expression>'
true
exit 0
```

**Baseline-attribution**

- LIVE3 실패는 orchestrator가 readiness 1차 보강 tree에서 실행한 `/tmp/t1074-root-lane-status-chain-live3.jsonl`에 귀속한다.
- RED/GREEN은 `WT-factory-mixed-hook@8c5d9be99`와 현재 사용자 승인 t1074 미커밋 변경에서, provider 환경 scrub 및 fresh `MOAI_HOME`으로 관측했다.

**Gaps**

- split-submit 보강 뒤 authenticated LIVE는 이 실행에서 재실행하지 않았다. 실제 UserPromptSubmit 실행, bind, assistant rollout 및 후속 sentinel은 orchestrator 재실행 대상이다.
- LIVE3에 hook subprocess 실행 증거가 없으므로 owner resolution product 결함은 주장하지 않는다.

**Residual-risk**

- exact composer glyph 재구성 방식은 LIVE4에서 부적합 판정을 받아 아래 output-activity/settle 방식으로 대체됐다.
- 실제 제출 뒤 hook이 실행되지만 bind가 실패하는 새 증거가 나타나면 그때 input/env/owner를 별도 진단해야 한다.
- repository-wide CI verdict는 integration branch CI 소유이며 현재 `PENDING`이다.

### LIVE4 raw ANSI delta와 activity-settle handoff — 2026-09-22

**Claim**

- LIVE4는 readiness를 통과했지만 prompt write 뒤 Codex TUI가 완성된 `› prompt` 문자열 대신 cursor-diff/status redraw만 내보내 exact 렌더 재구성이 10초 deadline에 실패했다. 이는 product hook 실행 전 harness 관측 가정의 실패다.
- terminal emulator를 추가하지 않았다. helper는 prompt text를 한 write로 보낸 뒤 terminal output이 실제로 변했는지와 terminal PID가 계속 live인지 확인하고, 500ms의 bounded settle interval이 지난 후 CR을 별도 write한다. output activity가 없거나 process가 죽으면 CR을 보내지 않고 bounded tail/identity 진단으로 실패한다.
- product hook/launcher는 변경하지 않았다.

**Evidence**

LIVE4 `/tmp/t1074-root-lane-status-chain-live4.jsonl`:

```text
LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2
NO_BYPASS_OK
PROMPT_COMPOSER_READY_OK lead,agent-1,agent-2
submit first prompt to terminal 0: prompt text was not rendered before deadline
terminal_tail="... cursor/status redraw ..."
CLEANUP_OK terminal_pid=73554 descendants=16
CLEANUP_OK terminal_pid=73245 descendants=16
CLEANUP_OK terminal_pid=73126 descendants=16
```

combined-write mutant RED:

```text
go test ./internal/cli -run '^TestFactoryOperationalPromptSubmitWaitsForActivityAndSettle$' -count=1 -timeout=30s -v
=== RUN   TestFactoryOperationalPromptSubmitWaitsForActivityAndSettle
    factory_operational_fixture_test.go:169: write prompt text: first write="Reply with exactly FACTORY_READY_1 and do not call any tool.\r"
--- FAIL: TestFactoryOperationalPromptSubmitWaitsForActivityAndSettle (0.00s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 0.792s
```

split/order/delay GREEN과 기존 gate:

```text
go test -race ./internal/cli -run '^(TestFactoryOperationalPromptReadinessGate|TestFactoryOperationalPromptReadinessProcessBoundary|TestFactoryOperationalPromptSubmitWaitsForActivityAndSettle|TestFactoryOperationalTrustBootstrap|TestFactoryOperationalHookTrustOnce)$' -count=1 -timeout=60s -v
--- PASS: TestFactoryOperationalPromptReadinessProcessBoundary (0.71s)
--- PASS: TestFactoryOperationalPromptSubmitWaitsForActivityAndSettle (0.52s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  3.379s

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=120s -json | jq -s '<exact eight-pass/no-fail/no-skip expression>'
true
exit 0
```

**Baseline-attribution**

- LIVE4 failure is attributed to the orchestrator-run `/tmp/t1074-root-lane-status-chain-live4.jsonl` on the preceding exact-render harness.
- RED/GREEN은 `WT-factory-mixed-hook@8c5d9be99`와 현재 사용자 승인 t1074 미커밋 변경에서 provider 환경 scrub 및 fresh `MOAI_HOME`으로 관측했다.

**Gaps**

- activity-settle 보강 뒤 authenticated LIVE는 이 실행에서 재실행하지 않았다. 실제 CR 제출, UserPromptSubmit bind 및 나머지 LIVE sentinel은 orchestrator 재실행 대상이다.
- LIVE4 역시 product hook subprocess가 실행되기 전 실패했으므로 hook input/env/owner 결함 증거가 아니다.

**Residual-risk**

- 일반 status spinner도 output activity로 인정한다. 이는 prompt echo를 raw delta에서 재구성하지 않는 최소 계약이며, 실제 제출 여부는 후속 UserPromptSubmit bind/rollout gate가 판정한다.
- terminal PID가 live여도 입력 focus가 다른 곳이면 후속 bind gate가 실패한다. helper 자체는 그 실패를 성공으로 바꾸지 않는다.
- repository-wide CI verdict는 integration branch CI 소유이며 현재 `PENDING`이다.

### LIVE5 실제 UserPromptSubmit와 POSIX owner stamp — 2026-09-22

**Claim**

- LIVE5에서는 정상 prompt 제출, 실제 `UserPromptSubmit`, model response가 모두 관측됐지만 lead가 `launch_pending`에 머물렀다. launcher는 direct Codex PID를 pending owner로 등록했으나 child env에서 `MOAI_SESSION_PID`를 제거했고, hook ancestry는 MacosSeatbelt wrapper에서 멈출 수 있었다.
- official Codex UserPromptSubmit payload field는 `prompt`다. 현재 tracked captured golden도 non-empty `prompt`, 실제 `session_id`, `hook_event_name=UserPromptSubmit`을 갖는다. 따라서 입력 key normalization이 이번 실패 원인이 아니다.
- POSIX direct 기본 경로는 현재 moai PID/fingerprint로 pending을 등록하고, inherited `MOAI_SESSION_PID`를 그 PID로 교체하며, 지정 cwd로 이동한 뒤 `syscall.Exec`으로 Codex를 실행한다. moai PID가 Codex PID로 유지되므로 hook은 sandbox ancestry 추측 없이 같은 owner를 해석한다. Windows는 기존 child Start/identity/register/Wait 경로를 유지한다. 주입된 `codexDirectLaunchFn` test seam은 변경하지 않았다.

**Evidence**

LIVE5 `/tmp/t1074-root-lane-status-chain-live5.jsonl`의 경계 출력:

```text
LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2
NO_BYPASS_OK
PROMPT_COMPOSER_READY_OK lead,agent-1,agent-2
... FACTORY_READY_1 ...
UserPromptSubmit did not rebind terminal 0
CLEANUP_OK terminal_pid=4936 descendants=16
CLEANUP_OK terminal_pid=4727 descendants=16
CLEANUP_OK terminal_pid=4603 descendants=20
```

tracked Codex golden field 확인:

```text
jq -c '{session_id_present:(.session_id|length>0),prompt_nonempty:(.prompt|length>0),hook_event_name}' .moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads/UserPromptSubmit.json
{"session_id_present":true,"prompt_nonempty":true,"hook_event_name":"UserPromptSubmit"}
```

POSIX subprocess RED:

```text
go test ./internal/cli -run '^TestCodexDirectPOSIXExecPreservesFactoryOwner$' -count=1 -timeout=30s -v
codex_launcher_exec_posix_test.go:35: MOAI_SESSION_PID="" want 41995
--- FAIL: TestCodexDirectPOSIXExecPreservesFactoryOwner
FAIL
```

POSIX exec GREEN과 관련 launcher race:

```text
go test -race ./internal/cli -run '^TestCodexDirectPOSIXExecPreservesFactoryOwner$' -count=1 -timeout=30s -v
--- PASS: TestCodexDirectPOSIXExecPreservesFactoryOwner (1.77s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  3.787s

go test -race ./internal/cli -run '^(TestCodexDirectPOSIXExecPreservesFactoryOwner|TestCodexChildEnvScrubsForeignAttribution|TestCodexVerbRouting_LaunchCountsPerVerb|TestCodexApp_LaunchedProgramsClosedSet|TestFactoryCodexSpawnRegistersLaunchPendingPeer|TestFactoryLauncherRegistersLaunchPendingPeers)$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  4.626s
```

subprocess helper는 exec 뒤 `MOAI_SESSION_PID == os.Getpid()`, exact cwd/argv, pending owner PID/process-start가 현재 process와 같음, inherited stale `MOAI_SESSION_PID=999999`가 교체됨을 함께 assertion한다.

```text
GOOS=windows GOARCH=amd64 go test -c ./internal/cli -o /tmp/t1074-cli-windows.test.exe
stdout/stderr empty, exit 0

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=120s -json | jq -s '<exact eight-pass/no-fail/no-skip expression>'
true
exit 0
```

**Baseline-attribution**

- LIVE5 failure is attributed to the orchestrator-run `/tmp/t1074-root-lane-status-chain-live5.jsonl` on the pre-exec direct launcher.
- RED/GREEN과 cross-platform compile은 `WT-factory-mixed-hook@8c5d9be99`와 현재 사용자 승인 t1074 미커밋 변경에서 provider 환경 scrub 및 fresh `MOAI_HOME`으로 관측했다.

**Gaps**

- POSIX exec 보강 뒤 authenticated LIVE는 이 실행에서 재실행하지 않았다. 실제 세 lane rebind와 나머지 LIVE sentinel은 orchestrator 재실행 대상이다.
- Windows는 compile만 관측했고 Windows 실제 process tree LIVE는 실행하지 않았다.

**Residual-risk**

- POSIX `syscall.Exec` 성공 뒤에는 launcher defer가 실행되지 않는다. factory run metadata와 pending 등록은 exec 전에 완료되며, session cleanup은 exec된 Codex process lifecycle에 귀속된다.
- `chdir` 또는 exec 자체가 실패하면 이미 기록한 pending row는 launcher process 종료 뒤 dead로 판정된다. 자동 삭제로 실패를 숨기지 않는다.
- repository-wide CI verdict는 integration branch CI 소유이며 현재 `PENDING`이다.

### Final run-phase status readback — 2026-09-23

- 위 LIVE1..15 절은 당시 관측을 보존한 진단 이력이다. 최종 판정은 `M5 및 operational final evidence — 2026-09-23` 절과 `.moai/reports/t1074/verdict.md`가 담당한다.
- AC-FMH-001..015와 AC-FMH-OPS-001..006은 모두 PASS다. independent sync audit도 `PASS`(85/100, blocking finding 0)로 닫혀 status는 `completed`다.
- 최종 operational 근거는 `/tmp/t1074-root-lane-status-chain-live16.jsonl`의 exact predicate `true`와 `USERPROMPT_REBIND_OK`, `BOUND_PROMPT_NO_REWRITE_OK`, `MCP_RESTART_OK`, `LEAD_MCP_ROSTER_OK`, `PROJECT_ISOLATION_OK`, `DEAD_OWNER_OK`, `NO_SYNTHETIC_TURN_OR_BYPASS_OK`, `CLEANUP_OK`, test PASS다.
- 기준 HEAD는 `8c5d9be993ad45d5c91ae8a95af904b27c935599`와 사용자 승인 t1074 변경이며 전체 SPEC status는 `completed`다. repository-wide CI verdict는 범위 밖 `PENDING`이다.

### AC-FMH-011 built `moai glm` remediation — 2026-09-23

**Claim**

- factory LIVE의 `claude` backend는 더 이상 `claude -p`를 직접 선택하지 않는다. 같은 fixture에서 freshly built한 `moai glm`을 실행하고, print/MCP arguments를 `--` 뒤에 전달한다.
- canonical LIVE의 isolated `MOAI_HOME`을 유지하면서 운영자 GLM credential은 메모리에서만 읽어 기존 `MOAI_TEST_GLM_KEY` child seam으로 전달한다. key는 로그, fixture file, `.mcp.json`에 쓰지 않는다.
- legacy AC10..14 fixture의 수동 peer identity와 production GLM launcher registration이 충돌하지 않도록 launcher process env에서는 factory attribution을 제거하고, non-secret broker/peer attribution은 turn별 strict `.mcp.json`의 MCP server env로만 전달한다.

**Evidence**

Command-route RED:

```text
internal/cli/factory_live_command_test.go:15:16: f.modelCommand undefined
FAIL  github.com/modu-ai/moai-adk/internal/cli [build failed]
```

Focused race GREEN:

```text
go test -race ./internal/cli -run '^(TestFactoryLiveClaudeCommandRoutesThroughBuiltMoaiGLM|TestFactoryLiveOperatorGLMKeyRestoresIsolatedHome|TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig)$' -count=1 -timeout=60s -v
--- PASS: TestFactoryLiveClaudeCommandRoutesThroughBuiltMoaiGLM (0.00s)
--- PASS: TestFactoryLiveOperatorGLMKeyRestoresIsolatedHome (0.00s)
--- PASS: TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  2.036s
```

첫 sandbox 실행은 Codex app-server가 `Operation not permitted`로 차단되어 코드 판정에 사용하지 않았다. 첫 escalated diagnostic은 GLM credential/launch까지 도달했지만 기존 manual bound peer와 production pending 등록이 충돌해 `Factory logical lane has a live owner`로 실패했다. 위 MCP-only attribution 분리 뒤 AC-FMH-011 재실행:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_TEST_GLM_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=codex-claude MOAI_HOME=/tmp/t1074-ac11-glm-json-home GOCACHE=/tmp/t1074-ac11-glm-json-cache go test -json ./internal/cli -run '^TestFactoryLiveCodexClaude$' -count=1 -timeout=240s > .moai/reports/t1074/m5-reg-ac11-glm.jsonl
live separate-context nonce/receipt verified: case=codex-claude nonce=057d60face93a6325425c3330f1914de acknowledged=2
--- PASS: TestFactoryLiveCodexClaude (112.72s)
predicate: {"pass":true,"fail":false,"skip":false,"not_run":false}
```

후속 rows는 각각 별도 bounded JSONL artifact로 순차 실행했다. AC12 최초 diagnostic은 GLM lead가 첫 turn일 때 fixture root에 `.moai` marker가 없어 실패했으며, fixture setup에 marker를 추가한 뒤 별도 rerun artifact가 PASS했다.

```text
.moai/reports/t1074/m5-reg-ac12-glm-rerun.jsonl
live separate-context nonce/receipt verified: case=claude-codex nonce=8e74ffdc2458360dfdf1a9a3fc9c552f acknowledged=2
--- PASS: TestFactoryLiveClaudeCodex (96.53s)

.moai/reports/t1074/m5-reg-ac13-glm.jsonl
live separate-context nonce/receipt verified: case=claude-claude nonce=8103c0a80c989e669bc79b9c6af17ea4 acknowledged=2
--- PASS: TestFactoryLiveClaudeClaudeCompletionSeparation (106.70s)

.moai/reports/t1074/m5-reg-ac14-glm.jsonl
idle pending truth and next-turn receipt verified: message=c26d88099996c5d86ab3c4d06808abc1
--- PASS: TestFactoryLiveHookBoundaryIdleTruth (48.05s)
```

각 final artifact의 parsed result는 `pass=true`, `fail=false`, `skip=false`, `not_run=false`였다. 최초 AC12 diagnostic은 `.moai/reports/t1074/m5-reg-ac12-glm.jsonl`에 FAIL로 별도 보존했다.

**Baseline-attribution**

- 실행 위치는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`, branch `WT-factory-mixed-hook`, baseline HEAD `8c5d9be993ad45d5c91ae8a95af904b27c935599`와 현재 사용자 승인 t1074 변경이다.
- PASS는 direct Claude가 아니라 built fixture `moai glm -- ...` 경로에 귀속한다. nonce와 receipt는 실제 Codex lead 및 GLM worker model contexts에서 관측됐다.

**Gaps**

- scoped AC/OPS live gap은 남지 않았다.
- repository-wide integration-branch CI verdict는 `PENDING`이다.

**Residual-risk**

- AC-FMH-011은 legacy real-model fixture의 manual peer contract를 보존하기 위해 factory attribution을 MCP subprocess에만 명시한다. production launcher pending/bind 계약은 별도 operational LIVE16이 담당한다.
- GLM provider model/approval 동작은 외부 서비스 상태에 영향을 받으므로 후속 integration CI나 재실행에서 회귀할 수 있다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-23
sync_commit_sha: 875efc28b
sync_status: complete
audit_verdict: PASS
audit_score: 85
audit_blocking_defects: 0
audit_dimensions:
  functionality: 100
  security: 100
  craft: 25
  consistency: 100
evidence_path: .moai/reports/t1074/sync-audit-final.md
b12_self_test_a: pass
b12_self_test_b: pass
b12_self_test_c: pass
changelog_entry_position: "[Unreleased] → Added (섹션 선두)"
frontmatter_status_transitions:
  spec_md: in-progress → completed
  plan_md: no-status-field
  acceptance_md: no-status-field
  progress_md: in-progress → completed
canary_compliance_check: not-applicable
readme_docs_site_entry: none
```

### 동기화 판정

- 독립 감사: `PASS`, 85/100, 병합 차단 finding 0.
- 완료 근거: 21개 live AC(AC-FMH-001..015, AC-FMH-OPS-001..006) 모두 scoped unit/LIVE evidence에서 PASS.
- README/docs-site 변경 없음: 이번 카드는 기존 CLI 표면의 factory 통신·상태 동작을 구현하며 별도 사용자 문서 표면을 추가하지 않는다.

### 완료 후에도 열려 있는 비차단 항목

- repository-wide integration-branch CI는 이 worktree의 scoped 검증으로 대체하지 않았으며 `PENDING`이다.
- Windows는 cross-compile만 관측했고 실제 Windows process-tree LIVE는 실행하지 않았다.
- 독립 감사의 Craft 25/100은 scoped selector가 package-wide 85% coverage를 증명하지 못한 데 따른 명시적 coverage GAP이다.
