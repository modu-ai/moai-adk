# t1074 final verdict

## Claim

- 기준 HEAD `8c5d9be993ad45d5c91ae8a95af904b27c935599`의 `WT-factory-mixed-hook` 작업 트리에서 AC-FMH-001..010, AC-FMH-015와 AC-FMH-OPS-001..006의 scoped unit/LIVE 증거가 PASS다.
- 최종 authenticated Codex chain인 `/tmp/t1074-root-lane-status-chain-live16.jsonl`은 production argv, prompt 전 세 provisional lane, 첫 정상 UserPromptSubmit rebind, 후속 prompt 무쓰기, lead-owned MCP, 프로젝트 격리, dead-owner, cleanup 및 no-bypass sentinel을 모두 남기고 PASS했다.
- AC-FMH-011..014는 Claude 접근을 freshly built `moai glm -- ...`로 교정한 뒤 실제 Codex↔GLM, GLM↔Codex, GLM↔GLM nonce/receipt와 idle-boundary를 모두 PASS했다.
- 독립 sync audit은 `PASS`(85/100, blocking finding 0)이며 모든 scoped AC/OPS 행이 PASS여서 SPEC status는 `completed`다.

## Evidence

- Atomic owner/bind race gate:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/t1074-evidence-atomic-cache go test -race ./internal/factorymsg ./internal/hook -run '^(TestBindLaunchPendingCannotOverwriteConcurrentAuthoritativePeer|TestBindLaunchPendingRejectsOwnerMismatch|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=90s -v
--- PASS: TestBindLaunchPendingCannotOverwriteConcurrentAuthoritativePeer (0.78s)
--- PASS: TestBindLaunchPendingRejectsOwnerMismatch (0.06s)
ok  github.com/modu-ai/moai-adk/internal/factorymsg  2.171s
--- PASS: TestFactorySessionStartRebindsLaunchPendingPeer (0.45s)
--- PASS: TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding (0.94s)
--- PASS: TestFactoryUserPromptSubmitRebindsLaunchPendingPeer (0.64s)
--- PASS: TestFactoryBoundUserPromptSubmitDoesNotRewritePeer (0.66s)
ok  github.com/modu-ai/moai-adk/internal/hook  4.219s
```

- M5 unit artifact predicate:

```text
jq -se '. as $events | ([ $events[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryCanonicalNamespaceAndIsolation|TestFactoryRunSelectionAtomicSlotsAndArgv|TestFactorySessionGenerationOwnership|TestFactoryEnvelopeIdempotencyAndStaleAck|TestFactoryCrashRecoveryExplicitReceipt|TestFactoryDeadLetterAndLegacyIsolation|TestFactoryHookContextAndContinuationSafety|TestFactoryHookZeroTurnAndCapabilityTruth|TestFactoryBrokerTrustBoundaries|TestFactoryHookBenchmarkBudget)$"))) | .Test ] | unique | length)==10 and (any($events[]; .Action=="skip") | not) and (any($events[]; .Action=="fail") | not) and (any($events[]; ((.Output // "") | contains("NOT_RUN"))) | not)' .moai/reports/t1074/m5-reg-unit.jsonl
true
```

- M5 Codex↔Codex LIVE: `.moai/reports/t1074/m5-reg-ac10.jsonl`:

```text
live separate-context nonce/receipt verified: case=codex-codex nonce=75bb84b10c8ee1397752a7c03a3681ed acknowledged=2
--- PASS: TestFactoryLiveCodexCodex (130.32s)
```

- M5 Claude-dependent LIVE: `.moai/reports/t1074/m5-reg-ac11.jsonl` through `m5-reg-ac14.jsonl` 모두 다음 exact result를 기록했다.

```text
Failed to authenticate: OAuth session expired and could not be refreshed
```

- 위 M5 AC11 artifact는 direct `claude -p` 경로의 과거 실패다. built fixture `moai glm` 경로로 교정한 AC11 재실행 artifact `.moai/reports/t1074/m5-reg-ac11-glm.jsonl`은 다음을 관측했다.

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_TEST_GLM_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=codex-claude MOAI_HOME=/tmp/t1074-ac11-glm-json-home GOCACHE=/tmp/t1074-ac11-glm-json-cache go test -json ./internal/cli -run '^TestFactoryLiveCodexClaude$' -count=1 -timeout=240s > .moai/reports/t1074/m5-reg-ac11-glm.jsonl
live separate-context nonce/receipt verified: case=codex-claude nonce=057d60face93a6325425c3330f1914de acknowledged=2
--- PASS: TestFactoryLiveCodexClaude (112.72s)
predicate: {"pass":true,"fail":false,"skip":false,"not_run":false}
```

GLM credential은 운영자 home의 기존 `.env.glm`에서 메모리로만 읽어 child의 기존 `MOAI_TEST_GLM_KEY` seam으로 전달했다. fixture `MOAI_HOME`은 원래 값으로 복구·유지했고 key를 로그나 `.mcp.json`에 기록하지 않았다.

- 남은 GLM live rows도 순차 실행해 JSONL artifact와 exact no-fail/no-skip/no-`NOT_RUN` 판정을 남겼다.

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

- Operational LIVE 진단 이력은 실패도 숨기지 않는다.
  - LIVE12와 LIVE13: `lead session rollout has no attributable factory_msg_status call/result`; current Codex output block/JSON-array shape를 parser가 귀속하지 못했다.
  - LIVE14: `MCP_RESTART_OK`와 `LEAD_MCP_ROSTER_OK`까지 통과한 뒤 두 번째 turn의 추가 calls를 첫 turn과 누적하지 못해 같은 attributable-result 오류로 실패했다.
  - LIVE15: trust TUI가 `Hooks need review`에 남아 `trust bootstrap did not reach post-trust idle terminal`로 실패했다.
  - LIVE16: `/tmp/t1074-root-lane-status-chain-live16.jsonl`의 exact predicate는 `true`였고 다음 sentinel과 최종 PASS를 관측했다.

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

- Final OPS split gates도 동일 baseline에서 exact predicate `true`를 반환했다.

```text
jq -se '([.[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$")))] | length)==8 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select(.Action=="fail")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0' /tmp/t1074-lane-status-unit-race.jsonl
true

jq -se '([.[] | select(.Action=="pass" and .Test=="TestFactoryLiveOperationalRosterBeforePrompt")] | length)==1 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select(.Action=="fail")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0 and any(.[]; (.Output // "") | contains("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")) and any(.[]; (.Output // "") | contains("LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("NO_SESSION_UUID_BEFORE_TURN_OK")) and any(.[]; (.Output // "") | contains("NO_BYPASS_OK"))' /tmp/t1074-lane-status-prompt-live.jsonl
true
```

- Sync finding remediation은 두 회귀를 test-first로 닫았다. `codex_launcher.go`의 `tmux kill-pane/display-message` process primitive를 `spawn.go`의 shared helper로 이동해 기존 injected identity/cleanup seams와 production 동작을 보존했다. Claude/GLM POSIX exec failure는 등록된 exact pending peer를 rollback하고 exec/rollback 오류를 `errors.Join`으로 함께 반환한다.

```text
TestCodexSpecFiles_ExecPrimitivesCodexOnly RED:
codex_launcher.go: process-start first argument "\"tmux\"" is not the expected req.Program
codex_launcher.go: process-start first argument "\"tmux\"" is not the expected req.Program
--- FAIL: TestCodexSpecFiles_ExecPrimitivesCodexOnly

TestClaudePOSIXExecFailureRollsBackPending RED:
exec failure left pending rows: [{Slot:lead Role:lead Backend:claude ... BindingState:launch_pending ...}]
--- FAIL: TestClaudePOSIXExecFailureRollsBackPending

go test -race ./internal/cli -run '^(TestCodexSpecFiles_ExecPrimitivesCodexOnly|TestFactoryCodexSpawnRegistersLaunchPendingPeer|TestClaudePOSIXExecFailureRollsBackPending|TestCodexDirectPOSIXExecFailureRollsBackExactPending)$' -count=1 -timeout=90s -v
--- PASS: TestCodexDirectPOSIXExecFailureRollsBackExactPending (0.56s)
--- PASS: TestCodexSpecFiles_ExecPrimitivesCodexOnly (0.05s)
--- PASS: TestFactoryCodexSpawnRegistersLaunchPendingPeer (0.42s)
--- PASS: TestClaudePOSIXExecFailureRollsBackPending (0.58s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  3.705s

go test -race ./internal/factorymsg ./internal/cli ./internal/hook -run '<exact eight selectors>' -count=1 -timeout=120s -json | jq -s '<eight unique pass/no fail/no skip/no NOT_RUN>'
true
```

## Baseline-attribution

- 위 결과는 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`, branch `WT-factory-mixed-hook`, HEAD `8c5d9be993ad45d5c91ae8a95af904b27c935599`와 사용자 승인 t1074 미커밋 변경을 대상으로 관측했다.
- LIVE12..16은 각각 `/tmp/t1074-root-lane-status-chain-live12.jsonl`부터 `live16.jsonl`까지의 실제 orchestrator 실행 로그에 귀속한다. LIVE16 이전 실패는 최종 PASS로 소급 변경하지 않는다.
- M5 unit/AC 로그는 `.moai/reports/t1074/m5-reg-*.jsonl`에 귀속한다.

## Per-AC verdict

| AC | Verdict | Evidence |
|---|---|---|
| AC-FMH-001..009 | PASS | `m5-reg-unit.jsonl`; predicate `true` |
| AC-FMH-010 | PASS | `m5-reg-ac10.jsonl`; nonce `75bb84b10c8ee1397752a7c03a3681ed`, `acknowledged=2` |
| AC-FMH-011 | PASS | `m5-reg-ac11-glm.jsonl`; built fixture `moai glm`, nonce `057d60face93a6325425c3330f1914de`, `acknowledged=2` |
| AC-FMH-012 | PASS | `m5-reg-ac12-glm-rerun.jsonl`; nonce `8e74ffdc2458360dfdf1a9a3fc9c552f`, `acknowledged=2` |
| AC-FMH-013 | PASS | `m5-reg-ac13-glm.jsonl`; nonce `8103c0a80c989e669bc79b9c6af17ea4`, `acknowledged=2` |
| AC-FMH-014 | PASS | `m5-reg-ac14-glm.jsonl`; idle pending truth and next-turn receipt PASS |
| AC-FMH-015 | PASS | `m5-reg-unit.jsonl`; benchmark selector PASS |
| AC-FMH-OPS-001..006 | PASS | atomic unit/race gates plus LIVE16 exact predicate and sentinels |

## Gaps

- provider-backed AC-FMH-010..014, operational LIVE16, independent sync audit `PASS`를 모두 관측했다.
- LIVE16 원본은 현재 `/tmp/t1074-root-lane-status-chain-live16.jsonl`에 있으며 이 보고서에는 판정에 필요한 exact sentinel만 보존했다.
- repository-wide integration-branch CI verdict는 `PENDING`이며 이 scoped run이 대신하지 않는다.
- Windows는 compile 경계만 관측됐고 실제 Windows process-tree LIVE는 실행하지 않았다.
- scoped selector는 package-wide 85% coverage를 증명하지 않으므로 coverage는 명시적 비차단 GAP이다.

## Residual-risk

- Codex rollout/event shape가 바뀌면 fail-closed parser가 새 shape를 거부할 수 있다. LIVE12..14가 그 위험을 실제로 드러냈으며 LIVE16은 현재 shape에서만 PASS를 증명한다.
- trust TUI는 외부 UI 상태에 민감하다. LIVE15 flake 뒤 LIVE16은 통과했지만 UI 변화에 따른 재발 가능성은 남는다.
- live provider/TUI 동작은 외부 서비스와 client UI 변화에 영향을 받으므로 integration CI 및 후속 재현에서 다시 실패할 수 있다.
