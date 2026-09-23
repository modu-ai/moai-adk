---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: progress
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## §A Status

- Current SPEC status: `in-progress` (M1 commit, manager-develop). The plan-era lines below are kept as written.
- Current phase: run, M1 complete; M2 not started (see §E.2).
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

## §E.2 Run-phase Evidence

### M1 — Durable reservation and worktree provenance (2026-09-23, manager-develop, cycle_type=tdd)

**Run plan head (N11).** The original plan milestones did not carry the REQ-FLH-016..018 work or the t1074 `RegisterPeer` transaction change. They are now placed as follows: REQ-FLH-016 (reservation reads the source row inside its own `BEGIN IMMEDIATE`, `ENDPOINT_LAUNCH_PENDING` NACK) is M1; REQ-FLH-017 (launcher registration finalizes an unfinished handoff inside `RegisterPeer`'s transaction) and REQ-FLH-018 (UserPromptSubmit registration rejected with `ENDPOINT_HANDOFF_PENDING` inside the same transaction), together with the t1074 `RegisterPeer` transaction change, are M3. M1 did not modify `RegisterPeer`.

**Scope delivered.** Handoff row, event, and tombstone tables in the existing broker schema (`internal/factorymsg/handoff.go`, appended to the `schema` executed by `OpenWithDeadline`; no new DB, no daemon). `Store.ReserveHandoff` pins project/run/lane/card/SPEC/source endpoint tuple/handoff generation/nonce/develop pin/target path/target branch in one write transaction; `MarkHandoffWTReady` / `NackHandoff` are nonce-and-state CAS transitions; a partial unique index enforces one unfinished handoff per lane. The cli controller `prepareLaneHandoff` (`internal/cli/factory_lane_handoff.go`) resolves canonical identity through `homestate.CanonicalProjectRoot` and the run through `factorymsg.ResolveActiveRun`, applies fail-closed admission, pins local `develop`, calls the existing materializer `materializeSessionWorktree` once, renames the branch to `WT-<slug>`, and verifies path, HEAD == pin, develop unmoved, branch, branch uniqueness, and a clean target before `WT_READY`. Creation-base drift becomes `NACK`/`BASE_DRIFT` with the target preserved. The N6/N10 re-entry rule reconstructs `WT_READY` without a materializer call and otherwise NACKs `TARGET_DIRTY`, `BRANCH_COLLISION`, `BASE_DRIFT` in that order.

**AC-FLH-019 order (vii) fixture.** Adversary A used the acceptance-allowed fixture: a raw `BEGIN IMMEDIATE` on A's own production-`Open` handle (production DSN) that rewrites the slot row to launch-pending and holds it. The REQ-FLH-017 finalize step inside `RegisterPeer` does not exist until M3.

#### RED evidence (verbatim, captured before any implementation)

First RED — AC-FLH-019 order (vii), HEAD `92c932cac`, log `.moai/reports/t1082/run-m1-red-ac19-vii.log`:

```text
$ go test -race ./internal/factorymsg -run '^TestFactoryLaneHandoffRebindVsUserPromptRegisterRace$' -count=1 -timeout=180s
# github.com/modu-ai/moai-adk/internal/factorymsg [github.com/modu-ai/moai-adk/internal/factorymsg.test]
internal/factorymsg/lane_handoff_race_test.go:143:8: undefined: Handoff
internal/factorymsg/lane_handoff_race_test.go:149:21: vStore.ReserveHandoff undefined (type *Store has no field or method ReserveHandoff)
internal/factorymsg/lane_handoff_race_test.go:149:41: undefined: HandoffReservation
internal/factorymsg/lane_handoff_race_test.go:151:21: undefined: HandoffModeInteractive
internal/factorymsg/lane_handoff_race_test.go:201:21: undefined: HandoffNackReason
internal/factorymsg/lane_handoff_race_test.go:202:27: undefined: NackEndpointLaunchPending
internal/factorymsg/lane_handoff_race_test.go:203:70: undefined: NackEndpointLaunchPending
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg [build failed]
FAIL
exit=1
```

RED reason: reservation code absent. Second RED — AC-FLH-001/002/016/017, log `.moai/reports/t1082/run-m1-red-cli.log`: build failed on `undefined: laneHandoffRequest`, `undefined: prepareLaneHandoff`, `undefined: factorymsg.NackEndpointLaunchPending` (exit 1) — controller code absent.

**Discrimination check (read-before-BEGIN mutant).** After GREEN, `ReserveHandoff` was temporarily changed to read the source row through `s.db` before `BeginTx`. The (vii) test then failed with exactly the predicted defect — `subject result = {... Source:{... SessionUUID:src-uuid Generation:2 ...} State:RESERVED ...} err=<nil>, want NACK ENDPOINT_LAUNCH_PENDING` (exit 1, log `.moai/reports/t1082/run-m1-mutant-read-before-begin.log`). The original was restored and re-verified.

#### GREEN evidence

| Check | Command (scrubbed env, one invocation each) | Result |
|---|---|---|
| AC-019 (vii) | `go test -race -v ./internal/factorymsg -run '^TestFactoryLaneHandoffRebindVsUserPromptRegisterRace$'` | PASS; log lines `RACER_HANDLES_DISTINCT=2`, `V_blocked_observed=true`, `A_commit_at < V_return_at` |
| cli M1 tests | `go test ./internal/cli -run '^(TestFactoryLaneHandoffLaunchPendingSourceNack\|...AdmissionFailClosed\|...DevelopPinAndTraceability\|...CreationBaseDriftRejected)$' -v` | 4/4 PASS, `.moai/reports/t1082/run-m1-green-cli.log` |
| race, factorymsg | `go test -race -count=1 ./internal/factorymsg/...` | `ok` (in `.moai/reports/t1082/run-m1-race.log`; re-run with coverage `.moai/reports/t1082/run-m1-cover-factorymsg.log`) |
| race, cli subset | `go test -race -count=1 ./internal/cli -run '^(TestFactoryLaneHandoff.*\|TestSessionWorktree.*\|TestFactoryMsgStatusReadOnlyRoster\|TestFactoryLeadNoticeUsesOperationalStatus\|TestMoaiMCPServer_RegistrationMatchesCatalog)$'` | exit 0, 17 PASS |
| vet | `go vet ./internal/factorymsg/... ./internal/cli/...` | exit 0 |
| lint | `golangci-lint run ./internal/factorymsg/... ./internal/cli/...` | exit 0, `0 issues.` |
| build | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 / exit 0 |

#### AC matrix for M1

| AC | Status | Acceptance command | Exit / output | Evidence |
|---|---|---|---|---|
| AC-FLH-001 | PASS | acceptance.md verbatim (`./internal/factorymsg ./internal/cli`) | `true`, exit 0 | `.moai/reports/t1082/ac01.jsonl` — 13 subtests incl. re-entry clean/dirty/branch/pin |
| AC-FLH-002 | PASS | acceptance.md verbatim (`./internal/cli`) | `true`, exit 0 | `.moai/reports/t1082/ac02.jsonl` |
| AC-FLH-016 | PASS | acceptance.md verbatim (`./internal/cli`) | `true`, exit 0 | `.moai/reports/t1082/ac16.jsonl` — main-creation mutant, develop-moves-during-create, matching control |
| AC-FLH-017 | PASS (see gap on spies) | acceptance.md verbatim (`./internal/factorymsg ./internal/cli`) | `true`, exit 0 | `.moai/reports/t1082/ac17.jsonl` |
| AC-FLH-019 | NOT PASS — order (vii) only | acceptance.md verbatim (`-race ./internal/factorymsg ./internal/hook`) | gate prints `true`, exit 0 | `.moai/reports/t1082/ac19.jsonl`. The gate's `true` is NOT an AC-FLH-019 PASS: orders (i)–(vi) and the 200 unforced iterations need the M3 rebind and `RegisterPeer` changes. The test logs `AC_FLH_019_ORDERS_COVERED=vii` |

#### Gaps

- AC-FLH-019 orders (i)–(vi) and the unforced iterations: not written, pending M3. AC-FLH-018 not started (M3).
- AC-FLH-017 app-server spy: preparation has no relocation step in M1, so the spy is structurally unreachable; the count 0 becomes a live guard only when M2 wires the headless switch. The "rollback spy" is realized as byte-identity of the provisional row (including `updated_at`) plus the controller having no access to `RollbackLaunchPending`; no call-counting rollback spy exists.
- AC-FLH-017 "dispatch release markers": no release-marker surface exists before M3; the test counts `messages` rows (0) instead.
- `validate()` branches in `handoff.go` were not each driven by an individual RED; `TestLaneHandoffStoreLifecycle` was added after GREEN as supplementary coverage.
- Package coverage: `internal/factorymsg` 68.0% (handoff.go functions 66.7–100%, uncovered lines are DB-error branches); `factory_lane_handoff.go` 61.9–100% per function from the handoff tests only. The package-wide 85% target is not met for `internal/factorymsg`; its pre-M1 baseline was not measured.

#### Residual risks

- `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` and `TestFactoryBoundUserPromptSubmitDoesNotRewritePeer` (package `hook`) are flaky on the unmodified HEAD `92c932cac` factorymsg as well: 3/5 runs failed with `factory endpoint is launch-pending` (`.moai/reports/t1082/run-m1-baseline-hook-HEAD-x5.log`); with M1 1/5 and 1/5 (env-scrubbed). Not attributable to M1; cause not investigated (out of scope). `TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding` passed 15/15.
- `TestEnsureGLMCredentials` and `TestEnsureGLMCredentialsFilePerm` fail on HEAD in this environment too (`.moai/reports/t1082/run-m1-baseline-hook-HEAD.log`).
- The controller's git facts (dirty source, branch, path) are read outside the broker transaction; a concurrent actor can change them between the check and the reservation. Post-create verification re-reads them, so a race produces a NACK, not a false `WT_READY`.
- The materializer resolves the repository from the process cwd; the controller verifies the created path equals `<canonical primary>/.claude/worktrees/<card-id>` and NACKs otherwise.
