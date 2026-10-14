# SPEC-MCP-NEUTRALITY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-14
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L) + decision-index.md (decision_gate: on)
plan_audit_verdict: PASS-WITH-DEBT (r3, overall_score 0.91; trajectory 0.625 → 0.825 → 0.91)
plan_audit_sha: 9be172fce7
plan_artifact_hash: 61225de45bf1ed864b5af169632dca8db6499c1c4ba6216aaaee3937f599270c
plan_audit_receipt: rcpt-7323ca7bef2c885633529458 (codex required-gate pass, 0 findings)
plan_audit_verdict_file: .moai/reports/t1632/plan-audit-r3.md
debts: N3 (serial-slot instance red at M1 RED), N4 (AC-MN-023 reuse-refusal + restart-persistence depth, M6), N5 (AC-MN-012 "without stalling" quantification, M3)

## §E.2 Run-phase Evidence

### M1 — structured refusals + identity (card t1632, branch WT-3-2-0-2, HEAD 6199704240)

Commits (oldest first): `ebba213f0` (refusal contract + FactoryRoleOperator + ErrRunNotActive sentinel + spec.md draft→in-progress), `9e7a4a9a9` (handler wiring + instruments), `92cabf6ef` (observer split), `6199704240` (Codex env allowlist). Push state: nothing pushed (lane protocol — merged-local comes later); `git status -sb` = `## WT-3-2-0-2` with no upstream and a clean tree.

E8 — verbatim RED outputs (captured BEFORE each GREEN; final pre-GREEN run):

```
$ go test ./internal/cli -run '^(TestMCPRefusal_SerialSlotStructured|TestMCPFactoryActor_OwnedAndStaleRun|TestMCPOperatorMark_Required|TestTodoClaimMCP_Mirror)$' -count=1   → exit 1
--- FAIL: TestMCPRefusal_SerialSlotStructured (2.22s)
    mcp_refusal_test.go:135: refusal rode the error channel (is_error true); text=factory_next: factory next: refused serial-slot: another serial card is in flight and holds the serial slot
--- FAIL: TestMCPFactoryActor_OwnedAndStaleRun (9.65s)      [actor lane-2 arm: the stage SUCCEEDED as env lane-1 — the actor argument was ignored; stale-run arm: toolErr "factory card not found"]
--- FAIL: TestMCPOperatorMark_Required (12.95s)             [unmarked decide/todo_claim/todo_add arms SUCCEEDED — env absence granted operator]
--- FAIL: TestTodoClaimMCP_Mirror (0.58s)
    todo_claim_test.go:410: refusal rode the error channel (is_error true); text=todo_claim: moai claim: refused — lane boundary: ...
```

```
$ go test ./internal/hook -run '^TestFailureObserver_ExcludesPolicyRefusals$' -count=1   → exit 1
    failure_observer_refusal_test.go:79: inbox rows = 3, want exactly 1 (the genuine error only):
    [tool_failure:mcp__moai__factory_next:UnknownFailure, tool_failure:mcp__moai__todo_claim:UnknownFailure, tool_failure:Bash:ExitError]
```

AC matrix (M1-mapped rows; every row's command run at HEAD 6199704240, this tree):

| AC | Status | Command (verbatim) | Observed output |
|----|--------|--------------------|-----------------|
| AC-MN-001 (serial-slot instance; card gate 1) | PASS | `go test ./internal/cli -run '^TestMCPRefusal_SerialSlotStructured$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` — refusal is_error false, `refused:true, code:"serial-slot"`, text block carries the same JSON |
| AC-MN-001 class flip (EV-001 instrument) | PASS | same run selector incl. `TestTodoClaimMCP_Mirror` | `ok` — the lane-session todo_claim refusal is structured, CLI-identical text in the reason |
| AC-MN-002 (observer split; card gate 2) | PASS | `go test ./internal/hook -run '^TestFailureObserver_ExcludesPolicyRefusals$' -count=1` | `ok` — 2 refusal payloads + 1 genuine error → exactly 1 inbox row (the genuine Bash failure); usage-log gains 1 tool_failure for the same seam reason |
| AC-MN-004 (actor owned + stale-run) | PASS | `go test ./internal/cli -run '^TestMCPFactoryActor_OwnedAndStaleRun$' -count=1` | `ok` — actor lane-2 on a lane-1 card → `owned`, is_error false; actor lane-1 on a retired run → `stale-run`, is_error false; actor lane-1 on the active run proceeds |
| AC-MN-005 (positive operator mark) | PASS | `go test ./internal/cli -run '^TestMCPOperatorMark_Required$' -count=1` | `ok` — unmarked session refuses (unknown attribution, no closed-set code); lane session refuses with the CLI-identical boundary text; `MOAI_FACTORY_ROLE=operator` proceeds (factory_decide, todo_claim, todo_add) |
| AC-MN-003 (card gate 3) | NOT M1 | — | M3 flips it (plan §F M3 measurement harness) |

Verification batch at HEAD 6199704240 (this tree, this run):

- E2 builds: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (darwin/arm64 also verified).
- E3 coverage (change-scoped, per plan D8; full-matrix judgment belongs to CI): `mcp_refusal.go` 11 funcs average **94.2%** (toolRefusal 75% = the unreachable marshal-error branch; mcpRenewActorLabel 75%); `internal/hook` observer additions: recordToolFailureEvent 94.1%, mcpPolicyRefusalMark 100%, mcpRefusalInValue 90.9%, mcpRefusalInRawText 90.0%. Instrument: `go test -coverprofile` over the affected families.
- E4 boundary grep: `grep -rn "AskUserQuestion" internal/cli internal/hook` — the only hits are pre-existing documentation strings (harness.go "does not directly call", mcp_server.go result-shaping comment); zero in the M1 diff.
- E5 lint: `golangci-lint run` — internal/cli **0 issues** (isolated-cache run after waiting out a sibling session's lint lock), internal/hook + internal/config + internal/codexwiring + internal/factorymsg **0 issues**. No NEW findings, no baseline findings surfaced in the touched packages.
- Affected families green (-count=1): TestSD_AC010/014(×2)/015/016/024, TestFactoryMsg*/TestFactoryMCPIdentity*/TestFactoryLeadNotice*, TestFactoryNextNominateMCPParity, TestFactoryNominate/Quota/Complete sets, TestQWR_* (quota worktrees), TestTodoClaimMCP_Mirror, TestTodoAddMCPCarriesPresentation, `./internal/factorymsg` package, `./internal/hook` failure family, `./internal/config` + `./internal/codexwiring` packages.

Interpretation notes for the audit:

1. **Identity-class refusals carry no code** (REQ-MN-003's "refused with refused: true"): the closed set (plan D3) names no member for unknown-attribution / lane-boundary / not-a-lane, and inventing one is forbidden — the payload omits `code`. The five named classes carry their codes. EV-001's green prose says "refused:true+code"; the serial-slot instance (AC-MN-001's named class) carries `code: serial-slot`; the mirror arm asserts the structured shape with the code-omitting identity payload.
2. **Refusal tokens outside the closed set** (dropped, held, recorded, quota-hold, hold-marker, blocked, backend-skip, foreign-worktree, unknown-card) and the Codex merge-edge refusal keep today's error-channel shape — REQ-MN-001 names exactly five classes; broadening is a D3 SPEC amendment.
3. **Argument errors stay isError:true** per design D3 step 2 (missing project_root, blank card, malformed run id, legacy actor label) — the pinned wire shape.
4. **The observer exclusion covers both writes** (usage-log event + lessons-inbox stub) at the one seam: a policy refusal is not a malfunction for either consumer; AC-MN-002's metric (the inbox) is asserted directly.
5. **CLI/MCP gate divergence is sanctioned by plan D1**: factory_card.go is off-limits (parallel repair t1696), so the positive operator mark and the actor argument live at the MCP layer; the CLI verbs absorb after t1696 lands (lane board conditional task #5).
6. **Pre-existing load-sensitive flake classified**: TestFactoryLeaseQueueLockStallBounded failed once inside a 10-minute broad union run (timing assertion over by 19.7 ms; the union also hit go's 10m package timeout); isolated run and -count=3 pass. Not touched by this diff (CLI lease wait budget path).
7. **Existing behavior-pinning tests updated with the change** (the EV-018 pattern): the freeze snapshot (AC-SD-022) re-baselined with provenance, the MCP parity inputs pin gained `actor`, TestSD_AC016's unmarked success arm now stamps the operator mark, and the stage parity twin's refusal arm moved onto fresh fixtures (it previously passed by matching an illegal-transition error on both sides — the holder refusal it names requires a card still at plan).

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Input parameters:
- tier: L (25 REQ / 25 AC, 6 milestones M1-M6)
- scope (file count): ~15-25 Go source + test files across internal/cli (mcp_*), internal/hook, internal/factorymsg, internal/auditreceipt, internal/codexwiring
- domain count: 3+ (MCP server layer, hook observer, messaging broker, audit receipts) — but coding-heavy
- file language mix: ~100% Go (+ markdown SPEC artifacts)
- concurrency benefit: LOW (coding-heavy — sequential milestone dependencies; M1 refusal contract is consumed by M2-M6)
- Agent Teams prereqs: not applicable (no operator request)

Mode evaluation:
- direct: not selected — semantic multi-file implementation, not trivial
- fanout: not selected — coding-heavy work violates Anthropic's coding-task parallelism caveat; milestones share the refusal contract and have inter-file dependencies
- sweep: not selected — semantic new-code work, not mechanical-uniform transformation
- serial: **selected** — one manager-develop spawn per milestone, sequential M1→M6

Decision: serial

Justification: the milestone chain is a layered contract build — M1's structured-refusal + actor-argument result shape is consumed by every later milestone (M2 project_root, M3 gate sessions, M6 audit integrity). Parallel spawns would each need the finished M1 contract and would contend for the same files (mcp_server.go dispatch tables). Anthropic's guidance: coding tasks involve fewer truly parallelizable tasks than research; serial is the safe default. Kickoff gate met in autonomous form (audit-cross: plan-audit-r3 PASS-WITH-DEBT 0.91 + codex pass rcpt-7323ca7bef2c885633529458; decision record .moai/reports/t1632/kickoff-decision.md; artifacts unchanged since verdict).
