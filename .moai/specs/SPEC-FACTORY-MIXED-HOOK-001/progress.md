---
id: SPEC-FACTORY-MIXED-HOOK-001
document: progress
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-MIXED-HOOK-001

## §A Status

- Current status: `in-progress`.
- Card: `t1074` (`picked`).
- Worktree: `WT-factory-mixed-hook`.
- Plan baseline: `758314007` (local develop matched when authoring began).
- Implementation baseline: `8d8e101bf`; M1 implementation commit: `cb099897a`.
- Next gate: rerun AC-FMH-011 through AC-FMH-014 after Claude subscription capacity recovers, then independent sync audit.

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

- The user explicitly added operational lead visibility into real worker lane status. `operational-lane-status-addendum.md` is now normative through minimal cross-references in `spec.md`, `plan.md`, and `acceptance.md`; REQ-FMH-OPS-001..008 and AC-FMH-OPS-001..006 are implementation/verification `PENDING`.
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
- AC-FMH-011 was rerun after the Codex repair: the Codex lead send succeeded, then the Claude worker failed with `OAuth session expired and could not be refreshed`. The project-managed Claude profile was separately probed and authenticated far enough to return HTTP 429 weekly-limit exhaustion. AC-FMH-012 through AC-FMH-014 were not rerun because they require the same unavailable Claude subscription turn; all four rows remain FAIL, never skipped or passed.
- AC-FMH-015 now passes without changing its thresholds. The final matrix measured empty full-hook p50 `31.216917ms`, p95 `33.143167ms`, every inspection cell below 200ms, and bounded contention at `57.022917ms` with truthful `SQLITE_BUSY` degradation. The exact JQ gate printed `true`.

## §D.1 Codex cwd redesign decision

- Latest official source/docs and a local Codex `0.155.1` probe establish that interactive `/cd` can keep the visible conversation flow while rotating the physical session/thread UUID.
- `t1074` therefore owns the stable logical lane/current-endpoint broker seam only.
- `t1082` owns card worktree creation and `/cd` or headless `cwd` handoff through `BOUND`.
- `t1075` depends on t1082 and may wake only the current bound endpoint.

## §E Audit-ready signals

- Plan-phase: audit-ready; independent `PASS` at score `0.94`.
- Run-phase: M1/M2/M3 complete; M4 Codex and performance rows pass, while four Claude-dependent live rows remain blocked by current subscription capacity. Independent sync audit remains pending.

## §E.2 Run-phase evidence

| Criterion / invariant | Actual output | Status |
|---|---|---|
| AC-FMH-001..009 | Exact named Go JSON gates each produced final `true`; `ac01.jsonl`..`ac09.jsonl`. | PASS |
| AC-FMH-010 | Exact live gate printed `true`; separate Codex contexts completed nonce round trip with `acknowledged=2`. | PASS |
| AC-FMH-011 | Codex lead send returned `ok`; Claude worker returned `OAuth session expired and could not be refreshed`. | FAIL |
| AC-FMH-012 | Not rerun after repair because the required Claude turn is unavailable under the same exhausted subscription. | FAIL |
| AC-FMH-013 | Not rerun after repair because the required Claude turns are unavailable under the same exhausted subscription. | FAIL |
| AC-FMH-014 | Not rerun after repair because the required Claude turn is unavailable under the same exhausted subscription. | FAIL |
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
