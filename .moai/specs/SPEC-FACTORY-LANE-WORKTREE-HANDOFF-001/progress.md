---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## §A Status

- Current SPEC status: `draft`.
- Current phase: plan artifact authoring only.
- Card/worktree/branch: `t1082` / `.claude/worktrees/t1082` / `WT-factory-lane-worktree-handoff`.
- Plan subject HEAD: `bf39a539d97f49edf3b11517ee7c982239c60df3`.
- Implementation: NOT STARTED.
- Implementation Kickoff Approval: NOT REQUESTED / NOT GRANTED.
- Independent plan audit: prior review interrupted before final verdict; revised plan requires a fresh audit.

## §B Plan-phase artifacts

| Artifact | Plan-phase status |
|---|---|
| spec.md | authored as draft |
| plan.md | authored as draft |
| acceptance.md | authored as draft |
| design.md | authored as draft |
| research.md | authored as draft |
| progress.md | plan evidence only |

## §C Plan-phase evidence

### Claim

- t1082 worktree/branch와 t1074 dependency ancestry를 읽었다.
- 실제 creation-base drift가 reflog에 남아 있음을 관측했다.
- 기존 MCP catalog/status narrow baseline은 현재 tree에서 통과했다.
- 20개 t1082 named tests(AC 19 + 공통 gate-quality 1; `grep -E '^unset ' acceptance.md | grep -oE 'Test[A-Za-z]+' | sort -u | wc -l` → 20)는 현재 모두 부재하여 plan RED 상태다.
- Strict SPEC lint는 빈 finding 배열을 반환했고, 18 REQ/19 AC heading과 양방향 trace reference를 확인했다.
- 사용자 승인에 따라 interactive next-normal-turn SessionStart와 headless official returned-thread-ID direct BOUND를 mode별 계약으로 분리했다.
- 모든 acceptance jq gate가 child/subtest/package의 `fail` 또는 `skip`을 전역 거부하도록 강화됐다.

### Evidence

```text
git -C .claude/worktrees/t1082 status --short --branch
## WT-factory-lane-worktree-handoff

git -C .claude/worktrees/t1082 rev-parse --short HEAD
bf39a539d

git -C .claude/worktrees/t1082 reflog --date=iso --format='%h %gd %gs' -12
bf39a539d ... merge(t1082): absorb t1074 factory dependency
3f3ffbb57 ... merge develop: Fast-forward
2213871af ... Branch: renamed refs/heads/t1082 to refs/heads/WT-factory-lane-worktree-handoff

git -C .claude/worktrees/t1082 merge-base --is-ancestor 3f3ffbb57 bf39a539d
exit 0

git -C .claude/worktrees/t1082 merge-base --is-ancestor 8c5d9be99 bf39a539d
exit 0

go test ./internal/mcp -run '^(TestMoaiMCPTools_CatalogSize|TestMoaiMCPTools_FourteenWriteCapable|TestMoaiMCPTools_NoDuplicateNames|TestMoaiMCPToolNames_MatchesCatalog)$' -count=1 -timeout=90s
ok github.com/modu-ai/moai-adk/internal/mcp 0.238s

go test ./internal/cli -run '^(TestMoaiMCPServer_RegistrationMatchesCatalog|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$' -count=1 -timeout=90s
ok github.com/modu-ai/moai-adk/internal/cli 2.716s

go run ./cmd/moai spec lint SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 --strict --json
[]

rg -c '^### REQ-FLH-[0-9]{3} ' spec.md
15

rg -c '^### AC-FLH-[0-9]{3} ' acceptance.md
16
```

16개 exact function selector는 각각 stdout 0 bytes, exit 1이었다. 전체 이름과 criterion별 명령은 `acceptance.md`의 RED-now ledger에 있다.

```text
Codex 0.155.1 app-server bounded probe:
initialize + thread/start only; no turn/start
official thread ID returned
after 5s: factory_db=false, peers=0, current-session-id absent

codex app-server generate-json-schema --out <tmp>
exit 0
ThreadForkParams.properties.cwd = ["string", "null"]
ThreadForkParams.required = ["threadId"]
```

### Baseline-attribution

- 모든 Git/code observation은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1082`의 `bf39a539d` 기준이다.
- Go baseline command는 같은 invocation에서 provider env를 scrub하고 `MOAI_HOME=/tmp/t1082-plan-baseline-home`, `GOCACHE=/tmp/t1082-plan-baseline-cache`를 사용했다.
- Official API 사실은 2026-09-22에 OpenAI 공식 app-server 및 ChatGPT Worktrees 문서를 열어 확인했다.

### Gaps

- 개정본에 대한 fresh plan-auditor PASS와 Implementation Kickoff Approval은 아직 없다.
- 모든 t1082 unit/integration/LIVE test는 구현 전이며 NOT_RUN이다.
- 실제 interactive/headless handoff, BOUND receipt, code-write zero counters, nonce exchange는 관측되지 않았다.

### Residual-risk

- 단일 Codex 0.155.1 fixture의 5초 관측은 pre-turn SessionStart의 보편적 부재를 증명하지 않는다. 설계는 그 관측을 일반 법칙으로 사용하지 않고 headless 결합에서 SessionStart 전제를 제거했다.
- Existing `cwd_changed` session-registry relocation과 factory endpoint rebind의 경계가 구현 중 잘못 합쳐질 수 있다.
- Cross-DB write를 도입하면 atomicity가 깨지므로 plan-auditor가 storage boundary를 특별히 검토해야 한다.

## §D Run-phase state

- No implementation files changed by this plan-phase authoring task.
- No run-phase test result may be inferred from the existing t1074 baseline.
- No commit, push, PR, queue mutation, report verdict, or completion claim was performed.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: null
plan_status: draft
plan_audit: not-run
implementation_kickoff_approval: not-granted
```

`audit-ready`는 독립 plan audit PASS와 사용자 승인 전에는 기록하지 않는다.

## §F Phase 4 Mode Selection

- Tier: L.
- Mode selection: not-run; run phase is not authorized.
- This document does not select or start implementation agents.

## §G t1074 landing absorb + premise recheck (2026-09-23)

- Absorbed local develop `08113ff0f` (t1074 landed as `861510fb6`) at merge `a2d8afe84`; one comment-only conflict in `internal/mcp/catalog_test.go` resolved to develop wording.
- Premise recheck: `.moai/reports/t1082/premise-recheck-20260923.md` — 10 premises match; M1 (ResolveLane returns ErrEndpointLaunchPending, no REQ/AC for handoff during launch-pending), M2 (two rebind paths: BindLaunchPending vs RESERVED→BOUND), M3 (run resolver location wording).
- Run phase NOT started; awaiting operator Kickoff gate and a decision on M1/M2.

## §H Plan-audit closure (2026-09-23)

- iter-4 FAIL 0.81 → iter-5 FAIL 0.81 → iter-6 PASS-WITH-DEBT 0.86 (Tier L threshold 0.85, margin 0.0125) at SPEC commit `6ad4824a2`; reports `.moai/reports/t1082/plan-audit-iter{4,5,6}.md`; N1 decision `.moai/reports/t1082/n1-decision-20260923.md`.
- Carried debt (all optional, no BLOCKING): N7 major — reservation source-row read inside its own transaction is not discriminated; absorb as AC-FLH-019 order (vii) at the first RED of run phase. N8 spec.md:133 unconditional "leave unchanged" vs t1074 dead-owner restart. N9 fixture wording (ownerCurrent injection, source seed path, (v) not-current vs live owner). N10 dirty-target NACK reason. N11 plan.md milestones lack REQ-016..018 work incl. t1074 RegisterPeer transaction change. N12 progress.md frontmatter `updated`.
- Run phase NOT started; awaiting operator Implementation Kickoff Approval via lead.

## §I Implementation Kickoff (2026-09-23)

- Approval path: lead approved under the operator's standing delegation (09-23) using Jev (noul) answers kickoff 0.69 / scope_cut 0.12 / reversible 0.57; that path is in tension with CLAUDE.local.md §29 (operator gates are grade 3, never delegated to Jev), which the lead escalated to the operator. The operator then approved the Kickoff directly in the t1082 lane session via AskUserQuestion ("직접 승인하고 run 진행"). The binding approval is the operator's direct answer.
- Conditions (lead): (1) absorb N7 as the first RED — AC-FLH-019 order (vii), reservation reads the source row inside its own transaction; (2) scope unchanged (no Part C cut); (3) fix N8–N12 wording during run, and state the plan-milestone gap (REQ-FLH-016..018 work and the t1074 RegisterPeer transaction change) at the head of the run plan; (4) fill the plan-phase gaps (execution reproduction of the new forced orders, `-race`) with run evidence — live NOT_RUN is not PASS; (5) develop CI Race Test failure `TestCC_FactoryEntryThroughRunCC/-f_lane-2` AMBIGUOUS_FACTORY (factory_test.go:887, run 35802361895) is out of this card's scope.
- Pre-run absorb: local develop `f0fdd88e4` merged at `ec821ce83`, no conflicts; factory-code delta limited to t1097 close-error handling (`cc50115ae`, `06bd358a6`) — SPEC line coordinates in store.go / factory_messages.go shifted, semantics unchanged.
