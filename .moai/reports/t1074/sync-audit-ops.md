# Sync Audit — t1074 OPS

- SPEC: `SPEC-FACTORY-MIXED-HOOK-001`
- Subject: OPS delta (`REQ-FMH-OPS-001..008`, `AC-FMH-OPS-001..006`)
- Baseline: `WT-factory-mixed-hook` at `19e9fc68d2e6b37222042b6bc7568dc0a9a52f23` plus uncommitted OPS changes
- Verdict: **FAIL**
- Weighted score: **69/100**
- Merge-blocking: **YES** — F1 evidence-parser defect and missing production LIVE proof for AC-FMH-OPS-001/006

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | Exact four-name unit/MCP/hook JSON gate produced `true`, but AC-FMH-OPS-001/006 have no successful production LIVE run and F1 violates their owned-`call_id` evidence predicate. |
| Security (25%) | 75/100 | PASS | Safe project/run filtering and read-only behavior tests passed; F1 is a Medium evidence-integrity weakness, not a proven Critical/High production trust-boundary exploit. |
| Craft (20%) | 75/100 | PASS | Existing measured changed-function coverage is `85.7%`/`86.7%`; formatting and diff checks were clean. The missing empty-ID adversarial case lowers test completeness. |
| Consistency (15%) | 100/100 | PASS | Existing `peers`, `factory_msg_status`, process-identity probe, SessionStart registration, MCP registry, and locale patterns are reused without a second roster. |

Must-pass firewall: Functionality requires every AC to pass. AC-FMH-OPS-001 and AC-FMH-OPS-006 are unverified, so the overall verdict is FAIL regardless of the weighted score.

## Per-AC Verdict

| AC | Verdict | Basis |
|---|---|---|
| AC-FMH-OPS-001 | **FAIL / UNVERIFIED** | Production 1 lead + 2 agent SessionStart-before-prompt and lead-owned MCP result did not complete. F1 also weakens the owned-result parser. |
| AC-FMH-OPS-002 | **PASS** | Exact unit gate verifies deterministic slot order, current generation, real live fingerprint, stale mismatch, dead/unknown mapping, and `task_state=unknown`. |
| AC-FMH-OPS-003 | **PASS** | Exact project/run-isolation unit gate and MCP foreign-sentinel assertion pass. The later production phase-2 proof remains unobserved but is additionally required by AC006, not needed to replace the §6.1 AC003 gate. |
| AC-FMH-OPS-004 | **PASS** | Repeated actual MCP handler calls preserve serialized `peers`, `messages`, and `dead_letters`; an absent broker returns an error and remains absent. |
| AC-FMH-OPS-005 | **PASS** | SessionStart notice names `factory_msg_status({"run_id":...})`, the tool is registered, the actual handler responds, and the test rejects `factory_msg_list`. |
| AC-FMH-OPS-006 | **FAIL / UNVERIFIED** | No successful installed-binary/restarted-MCP/process-tree/cleanup production run exists. The only recorded attempt stopped before registration due Codex state DB permissions. |

## Findings

### F1 — Empty `call_id` is accepted as owned MCP evidence

- Severity: **Medium**
- Classification: **blocking**
- Confidence: **High; mechanically reproduced**
- Location: `internal/cli/factory_operational_live_test.go:322`, `internal/cli/factory_operational_live_test.go:340`, `internal/cli/factory_operational_live_test.go:348`
- Impact: `operationalMCPResultCount` inserts a status call under `calls[""]` and later accepts a missing-ID output through the same empty key. Thus a result without an attributable call identity is accepted by the evidence parser and can contribute a distinct entry to its minimum-call count. This violates the explicit owned-`call_id` proof requirement for AC-FMH-OPS-001/006.
- Required fix: reject missing or whitespace-only `call_id` before recording both `function_call` and `function_call_output`; add regression cells for missing/blank call IDs and for a valid call combined with an empty-ID pair at `minimum=2`.

Mechanical reproduction, using a read-only Go overlay outside the repository:

```text
GOCACHE=/tmp/t1074-sync-empty-callid-cache MOAI_HOME=/tmp/t1074-sync-empty-callid-home go test -overlay=/tmp/t1074_empty_callid_overlay.json ./internal/cli -run '^TestAuditOperationalEvidenceRejectsEmptyCallID$' -count=1
--- FAIL: TestAuditOperationalEvidenceRejectsEmptyCallID (0.00s)
    t1074_empty_callid_audit_test.go:9: empty call_id/result pair accepted as owned MCP evidence
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 0.733s
FAIL
```

No other code defect was reproduced in the reviewed OPS delta.

## Claim

The implementation correctly extends the existing broker status with a project/run-scoped, slot-sorted current roster. It keeps endpoint and task state separate, does not infer activity, reuses SessionStart registration, and exposes the real read-only MCP handler in lead guidance. The code-level OPS surface passes the four exact §6.1 gates. The card is nevertheless not acceptable because the production LIVE criteria are unverified and the LIVE evidence parser accepts an empty ownership identifier.

## Evidence

### Exact §6.1 unit/MCP/hook gate

The initial wrapper used zsh's reserved variable `status` and exited after the Go JSON file was written. This wrapper error is not treated as a test result. The persisted JSON was then evaluated with the canonical non-empty jq predicate:

```text
jq -se '([.[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$")))] | length)==4 and ([.[] | select(.Action=="skip" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$")))] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0' /tmp/t1074-sync-ops-unit.jsonl
true
```

Actual named PASS rows:

```text
TestFactoryLaneRosterStateTruth
TestFactoryLaneRosterProjectIsolation
TestFactoryMsgStatusReadOnlyRoster
TestFactoryLeadNoticeUsesOperationalStatus
```

### Existing parser race gate

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-sync-parser-home GOCACHE=/tmp/t1074-sync-parser-cache go test -race ./internal/cli -run '^TestFactoryOperationalEvidenceRejectsUnownedResult$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  1.949s
```

This PASS does not cover the empty-`call_id` counterexample in F1.

### Format and diff hygiene

```text
gofmt -l <changed OPS Go files>; git diff --check
<empty output>
exit 0
```

### Production LIVE evidence status

The current `progress.md` records the only attempted LIVE execution as a failure before the first peer registered:

```text
PRODUCTION_ARGV .../moai codex -f terminal_pid=60892
GAP: production SessionStart before prompt: want=1 got=0
Cause: failed to initialize state runtime at /Users/goos/.codex: failed to open state DB at /Users/goos/.codex/state_5.sqlite: error returned from database: (code: 8) attempt to write a readonly database
--- FAIL: TestFactoryLiveOperationalRosterBeforePrompt (35.03s)
FAIL  github.com/modu-ai/moai-adk/internal/cli  35.759s
exit 1
```

This audit did not rerun an authenticated external Codex session. The recorded failure is historical run evidence read from the subject tree, not a fresh PASS.

## Baseline-attribution

- Audit working tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- Branch: `WT-factory-mixed-hook`
- Start HEAD: `19e9fc68d2e6b37222042b6bc7568dc0a9a52f23`
- Subject: the uncommitted files listed by `git status --short` at that HEAD.
- Fresh measurements in this audit: four-name JSON/JQ gate result, parser race gate, formatting/diff check, and the `/tmp` overlay counterexample.
- Prior-run-only evidence: coverage figures and the failed production LIVE attempt recorded in `progress.md`; these were not promoted to fresh execution evidence.
- No implementation or canonical SPEC file was modified. Only this audit artifact was added.

## Gaps

- AC-FMH-OPS-001 and AC-FMH-OPS-006 lack a successful production run proving exact launcher argv, three real SessionStart owner identities before prompt, lead-owned restarted MCP PID/binary, same roster response, phase-2 isolation, and complete child cleanup.
- The failed LIVE run did not reach authentication, MCP restart, model tool call, or cleanup sentinels. It cannot establish those predicates.
- Windows runtime behavior was not exercised; the LIVE fixture is intentionally Darwin/Linux-only.
- The unit test uses injected process states for dead/unknown truth-table cells. This is permitted by the §6.1 unit gate but is not production evidence.
- The first unit wrapper's `status=$?` assignment failed because `status` is read-only in zsh. The generated JSON and canonical jq readback were valid; the wrapper failure itself proves nothing about product behavior.
- No broad package or repository-wide test was run, per the audit scope.

## Residual-risk

- `OpenExistingWithDeadline` checks file existence before opening without SQLite `mode=ro`; the normal absent-broker test passes, but the check/open race was not mechanically exercised and is therefore not reported as a defect.
- The process-tree harness samples descendants every 100 ms. A daemon that forks and reparents entirely between samples could evade inventory; no such survivor was reproduced.
- Actual Codex TUI startup, SessionStart timing, MCP automatic restart, rollout event shape, and Darwin `lsof` binary attribution remain dependent on the missing LIVE run.

## Iteration History

- Iteration 1: independent OPS sync audit. Four exact unit gates and the existing parser race test passed; F1 was newly reproduced with an out-of-tree overlay; production LIVE remained a GAP. Verdict: FAIL.

## Required next gate

1. Fix F1 and rerun its negative parser matrix.
2. Rerun the exact four-name §6.1 gate.
3. With explicit authorization, run both addendum §6.2 production LIVE commands and require every non-empty jq predicate and cleanup sentinel to pass.

---

## Iteration 2 — F1 Delta Re-audit

- Delta verdict: **PASS — F1 RESOLVED**
- Current overall SPEC verdict: **FAIL**
- Current merge blocker: production LIVE evidence GAP for AC-FMH-OPS-001/006 only
- Subject baseline: `WT-factory-mixed-hook` at `19e9fc68d2e6b37222042b6bc7568dc0a9a52f23` plus the uncommitted F1 fix

### Claim

F1의 재현 경로였던 빈 `call_id` 호출 등록과 빈 `call_id` 출력 수용이 모두 차단됐다. `both-empty`, `call-empty`, `output-empty` 세 회귀 셀이 기존 parser 계약 테스트에 추가됐고 좁은 race 실행이 통과했다. F1에 관한 코드 blocker는 해소됐다.

### Evidence

직접 판독한 변경:

```text
internal/cli/factory_operational_live_test.go:340
if p.Type == "function_call" && p.CallID != "" && ...

internal/cli/factory_operational_live_test.go:348
if p.Type == "function_call_output" && p.CallID != "" && calls[p.CallID] {
```

직접 판독한 회귀 셀:

```text
internal/cli/factory_operational_evidence_test.go:20  both-empty -> false
internal/cli/factory_operational_evidence_test.go:21  call-empty -> false
internal/cli/factory_operational_evidence_test.go:22  output-empty -> false
```

독립 재실행:

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-sync-parser-iter2-home GOCACHE=/tmp/t1074-sync-parser-iter2-cache go test -race ./internal/cli -run '^TestFactoryOperationalEvidenceRejectsUnownedResult$' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  1.910s
exit 0
```

작성자가 보고한 OPS four-name jq `true`는 이번 F1-only delta에서 재실행하지 않았으며, iteration 1의 독립 `true` 결과와 동일 영역에 대한 비회귀 참고값으로만 남긴다.

### Baseline-attribution

- 시작/종료 HEAD: `19e9fc68d2e6b37222042b6bc7568dc0a9a52f23`
- 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- 검증 범위: F1 parser guard와 정확한 단일 race 테스트
- 구현 코드와 canonical SPEC 문서는 수정하지 않았다. 이 iteration 부록만 감사 보고서에 추가했다.

### Gaps

- addendum §6.2 production LIVE 두 행은 실행하지 않았다.
- AC-FMH-OPS-001/006의 실제 launcher, SessionStart owner, restarted lead MCP, phase-2 isolation, cleanup 증거는 여전히 없다.
- broad test와 기존 4-name OPS gate는 이번 delta 범위에서 재실행하지 않았다.

### Residual-risk

F1은 닫혔지만 실제 Codex rollout이 생성하는 `call_id`·output 형식과 lead-owned MCP 재시작 사슬은 LIVE에서만 최종 확인할 수 있다. 따라서 코드 finding은 0건이지만 Functionality must-pass 방화벽은 AC-FMH-OPS-001/006 증거가 착지할 때까지 계속 FAIL이다.
