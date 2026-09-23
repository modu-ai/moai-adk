---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: progress
created: 2026-09-22
updated: 2026-09-24
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## §A Status

- Current SPEC status: `in-progress` (M1 commit, manager-develop). The plan-era lines below are kept as written.
- Current phase: run, M1, M2, M3a, M3b and M4 complete (M4 code HEAD `c16d1422a`; follow-up tests `a0d4755a7` on develop absorb `09e620d2d`); M5 (LIVE) blocked, both LIVE rows NOT_RUN = LIVE FAIL, gate-quality test delivered (see § M5 lane record).
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

### M2 — Interactive and headless relocation adapters (2026-09-23, manager-develop, cycle_type=tdd)

**Boundary.** Lane decision option A, recorded in plan.md at `3c6725fbb`: both adapters stop at `SWITCH_PENDING_*`. M2 writes no BOUND, tombstone, BOUND receipt, peer change, or dispatch release. AC-FLH-003/004 named tests are not written in M2 (M3 owns them); the M2 adapter tests use other names. AC-FLH-014 closes in M2.

**Scope delivered.**

- `internal/factorymsg/handoff.go`: CAS transitions `MarkHandoffSwitchPendingInteractive` / `MarkHandoffSwitchPendingHeadless`, each allowed only from `WT_READY` and only for its own stored mode (`transitionHandoff` now reads `state,mode` in the same transaction). New NACK reasons `TARGET_READBACK_MISMATCH`, `RELOCATION_RPC_FAILED`, `RELOCATION_EVIDENCE_INVALID`; method constants `thread/fork`, `thread/start`.
- `internal/factorymsg/handoff_relocation.go` (new): `lane_handoff_relocations` table in the existing broker schema (no new DB). `RecordHeadlessRelocation` re-reads state/mode/nonce inside its write transaction, accepts only `SWITCH_PENDING_HEADLESS`, validates the evidence, and inserts once. The validator refuses `SessionStart`, `turn/start`, `turn/steer`, `wrong_method_thread_start` (stored history + `thread/start`), `wrong_method_thread_fork`, missing `thread/started`, empty or non-new thread id, lineage not naming the source, and any cwd/branch/HEAD mismatch against the reservation.
- `internal/cli/mcp_codex.go` (existing client): `codexMethodThreadFork`, `codexNotifyThreadStarted`; `codexInitialize` extracted from `openCodexSessionResolved` (same summaries); `awaitCodexResponseObserving` (the existing parser, now handing non-matching lines to an observer; `awaitCodexResponse` delegates to it); `runCodexThreadRelocation` issues exactly one of `thread/fork {threadId,cwd}` / `thread/start {cwd}` over `codexSession`, reuses `conn.close()` for bounded teardown, and waits for `thread/started` of the returned id. No new transport.
- `internal/cli/factory_lane_handoff_switch.go` (new): `switchLaneHandoffInteractive` (idle check → `SWITCH_PENDING_INTERACTIVE` → operator guidance with `/cd <absolute target>` and the nonce; no app-server contact) and `switchLaneHandoffHeadless` (idle check → controller readback guard → `SWITCH_PENDING_HEADLESS` → fork/start → controller readback → `RecordHeadlessRelocation`). Non-idle, RPC failure, readback failure, and invalid evidence each record a NACK on the handoff. Production client `codexLaneHandoffAppServer` is bounded by `config.DefaultCodexHandoffRelocationTimeout` (60s, new in `internal/config/defaults.go`).
- `internal/cli/factory_lane_handoff.go`: `laneHandoffAppServer` now returns `codexThreadRelocation`; git subprocesses go through the `handoffCommand` seam so a test can observe every process started.

#### RED evidence (verbatim, captured before any M2 implementation)

HEAD `3c6725fbb`, log `.moai/reports/t1082/run-m2-red.log`:

```text
3c6725fbb
# github.com/modu-ai/moai-adk/internal/factorymsg [github.com/modu-ai/moai-adk/internal/factorymsg.test]
internal/factorymsg/handoff_switch_test.go:57:33: undefined: HeadlessRelocation
internal/factorymsg/handoff_switch_test.go:58:9: undefined: HeadlessRelocation
internal/factorymsg/handoff_switch_test.go:59:11: undefined: RelocationMethodThreadFork
internal/factorymsg/handoff_switch_test.go:73:17: s.MarkHandoffSwitchPendingHeadless undefined (type *Store has no field or method MarkHandoffSwitchPendingHeadless)
internal/factorymsg/handoff_switch_test.go:77:16: s.MarkHandoffSwitchPendingInteractive undefined (type *Store has no field or method MarkHandoffSwitchPendingInteractive)
internal/factorymsg/handoff_switch_test.go:81:17: s.MarkHandoffSwitchPendingInteractive undefined (type *Store has no field or method MarkHandoffSwitchPendingInteractive)
internal/factorymsg/handoff_switch_test.go:95:18: hs.MarkHandoffSwitchPendingInteractive undefined (type *Store has no field or method MarkHandoffSwitchPendingInteractive)
internal/factorymsg/handoff_switch_test.go:98:16: hs.MarkHandoffSwitchPendingHeadless undefined (type *Store has no field or method MarkHandoffSwitchPendingHeadless)
internal/factorymsg/handoff_switch_test.go:103:45: undefined: NackRelocationRPCFailed
internal/factorymsg/handoff_switch_test.go:116:35: undefined: HeadlessRelocation
internal/factorymsg/handoff_switch_test.go:116:35: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg [build failed]
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/factory_lane_handoff_switch_test.go:80:7: undefined: codexMethodThreadFork
internal/cli/factory_lane_handoff_switch_test.go:86:20: undefined: codexMethodThreadFork
internal/cli/factory_lane_handoff_switch_test.go:142:95: undefined: laneHandoffSwitch
internal/cli/factory_lane_handoff_switch_test.go:143:9: undefined: laneHandoffSwitch
internal/cli/factory_lane_handoff_switch_test.go:189:14: undefined: switchLaneHandoffInteractive
internal/cli/factory_lane_handoff_switch_test.go:223:14: undefined: switchLaneHandoffInteractive
internal/cli/factory_lane_handoff_switch_test.go:247:16: undefined: switchLaneHandoffHeadless
internal/cli/factory_lane_handoff_switch_test.go:254:47: undefined: codexMethodThreadFork
internal/cli/factory_lane_handoff_switch_test.go:260:28: f.store.HeadlessRelocationFor undefined (type *factorymsg.Store has no field or method HeadlessRelocationFor)
internal/cli/factory_lane_handoff_switch_test.go:264:23: undefined: factorymsg.HeadlessRelocation
internal/cli/factory_lane_handoff_switch_test.go:264:23: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
exit=1
```

RED reason: adapter, relocation client, and relocation store code absent. The RED is compile-level for all M2 tests, including AC-FLH-014.

**Mutant checks (guards observed red).** Each mutant was applied, run, and reverted:

| Mutant | Command | Result | Log |
|---|---|---|---|
| forbidden literal `tmux send-keys` in the interactive guidance | `go test ./internal/cli -run '^TestFactoryLaneHandoffNoPrivateControl$'` | FAIL, static half names `factory_lane_handoff_switch.go:130:3: literal ...`, exit 1 | `run-m2-mutant-static-literal.log` |
| interactive switch calls `StartThread` | same | FAIL `interactive app-server sessions = 1, want 0 (no model turn)`, exit 1 | `run-m2-mutant-interactive-appserver.log` |
| client always issues `thread/start` (history dropped) | `go test ./internal/cli -run '^(TestFactoryLaneHandoffNoPrivateControl\|TestLaneHandoffHeadlessSwitchForksStoredHistory)$'` | FAIL `handoff NACK RELOCATION_EVIDENCE_INVALID: wrong_method_thread_start` (3 failures), exit 1 | `run-m2-mutant-wrong-method.log` |

#### GREEN evidence

| Check | Command (scrubbed env, one invocation each) | Result |
|---|---|---|
| M2 tests | `go test -race -v ./internal/factorymsg ./internal/cli -run '^(TestHandoffSwitchPendingIsModeBound\|TestHeadlessRelocationEvidenceRejectsNonEvidence\|TestLaneHandoff.*Switch.*\|TestFactoryLaneHandoffNoPrivateControl)$'` | both packages `ok`; 11 parent PASS, 41 PASS lines, 0 FAIL/SKIP (`run-m2-green.log`) |
| race, factorymsg | `go test -race -count=1 ./internal/factorymsg/...` | `ok`, exit 0 (`run-m2-race-factorymsg.log`) |
| race, cli subset | `go test -race -count=1 ./internal/cli -run '^(TestFactoryLaneHandoff.*\|TestLaneHandoff.*\|TestSessionWorktree.*\|TestCodex.*\|TestMoaiMCPServer_RegistrationMatchesCatalog\|TestFactoryMsgStatusReadOnlyRoster\|TestFactoryLeadNoticeUsesOperationalStatus)$' -v` with `MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS` also unset | `ok`, exit 0; 278 PASS, 0 FAIL, 5 SKIP (pre-existing env-gated `TestCodexLive_*`), 0 DATA RACE (`run-m2-race-cli.log`) |
| race, handoff incl. supplementary tests | `go test -race -count=1 -run 'LaneHandoff' -coverprofile=… ./internal/cli/` | `ok`, exit 0 (`run-m2-cover-cli.log`) |
| vet | `go vet ./internal/factorymsg/... ./internal/cli/... ./internal/config/...` | exit 0 |
| lint | `golangci-lint run ./internal/factorymsg/... ./internal/cli/... ./internal/config/...` | `0 issues.`, exit 0 |
| build | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 / exit 0 (`run-m2-build.log`) |

#### AC matrix for M2

| AC | Status | Acceptance command | Exit / output | Evidence |
|---|---|---|---|---|
| AC-FLH-014 | PASS | acceptance.md verbatim (`./internal/cli ./internal/hook`), prefixed with the lane env scrub | `true`, exit 0 | `.moai/reports/t1082/ac14.jsonl`; re-run on the final tree after the supplementary tests |
| AC-FLH-003 | NOT_RUN | named test is M3 (option A) | — | adapter behavior covered by `TestLaneHandoffInteractiveSwitch*` (different names) |
| AC-FLH-004 | NOT_RUN | named test is M3 (option A) | — | adapter behavior covered by `TestLaneHandoffHeadlessSwitch*` (different names) |

M1 regression (acceptance.md verbatim, lane env scrub prefix): AC-FLH-001 `true`, AC-FLH-002 `true`, AC-FLH-016 `true`, AC-FLH-017 `true`, AC-FLH-019 (order vii only) `true` — each exit 0; jsonl files overwritten in place.

#### Coverage (file-level, statement-weighted, from the cover profiles)

| File | Profile | Covered |
|---|---|---|
| `internal/factorymsg/handoff_relocation.go` (new) | `go test -count=1 -coverprofile ./internal/factorymsg/` (package 70.4%, M1 68.0%) | 46/48 = 95.8% |
| `internal/factorymsg/handoff.go` | same | 86/100 = 86.0% |
| `internal/cli/factory_lane_handoff_switch.go` (new) | `go test -race -count=1 -run 'LaneHandoff' -coverprofile ./internal/cli/` | 66/76 = 86.8% |
| `internal/cli/factory_lane_handoff.go` | same | 80/100 = 80.0% (M1 code; unchanged figure) |
| `internal/cli/mcp_codex.go` card-added (`codexInitialize`, `runCodexThreadRelocation`, `awaitCodexResponse*`) | `go test -count=1 -run '(LaneHandoff\|Codex)' -coverprofile ./internal/cli/` | 57/64 = 89.1% |

Filtered profiles: `run-m2-factorymsg.cov`, `run-m2-cli-handoff-card.cov`, `run-m2-cli-card.cov`.

#### Gaps

- AC-FLH-003/004 named tests and their BOUND observation: M3. NOT_RUN is not PASS.
- Live Codex: the relocation client ran only against the in-process fake app-server on the `codexConn` seam (official 0.155.1 response shapes, request lines parsed back from the wire). The real subprocess transport (`realCodexConn` pipes, `readLoop`, kill-after-3s close) was not exercised for `thread/fork`; that is M5 LIVE. NOT_RUN.
- The fake is scripted by me: whether a real 0.155.1 app-server emits `thread/started` after a `thread/fork` response (both orders are tested), and whether `thread.cwd` always equals top-level `cwd`, are not observed.
- `TestLaneHandoffSwitchRefusesSecondSwitch`, `TestLaneHandoffHeadlessSwitchNacksMissingCodexBinary`, and the reverse half of `TestLaneHandoffSwitchModesDoNotCross` were added after GREEN to lift switch-file coverage from 81.6% to 86.8%; they had no individual RED.
- The source lane's activity (idle / active turn / permission wait / interrupting) is an input the caller supplies; M2 does not observe it from the app-server (no `thread/read`).
- `factory_lane_handoff.go` stays at 80.0% (below 85%, M1 code, lowest `createHandoffTarget`).
- Not measured: the full `internal/cli` and `internal/hook` suites (CI owns them).

#### Residual risks

- A NACK after `thread/fork` was issued leaves the forked thread orphaned on the app-server side (design.md §2.1 already records this; cleanup is out of scope).
- The readback is taken before the RPC (guard) and after it (recorded); the target can still change between the recorded readback and the M3 rebind, which re-validates.
- `TestCodexSpawn_RealAssemblyThroughStubTmux` fails when the lane's factory env (`MOAI_KANBAN_BACKEND`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`) is not scrubbed; it passes with them unset (`run-m2-codexspawn-envscrub.log`). Not touched by M2; the prescribed scrub list lacks those three.
- `.moai/reports/t1082/ac008-idempotency-check.md` was modified in this worktree at 10:28:48 by another writer during this run (a "개정 반영" section, +16 lines). Not staged by M2; reported to the lead.

### M3a — Atomic mode-evidence rebind, pre-BOUND gates, stale redirect (2026-09-23, manager-develop, cycle_type=tdd)

**Split.** The lane split plan.md M3 into M3a (AC-FLH-003..008, this section) and M3b (REQ-FLH-017/018 inside t1074 `RegisterPeer`, AC-FLH-018, AC-FLH-019 orders (i)–(vi)). plan.md is unchanged. M3a does not modify `RegisterPeer`.

**Scope delivered (commit `399a95c9c`).**

- `internal/factorymsg/handoff_bind.go` (new): `Store.BindHandoff` — one write transaction (`BEGIN IMMEDIATE` from the `_txlock=immediate` DSN) that reads the handoff by id, refuses a stale nonce or a non-pending state with `STALE_GENERATION`, CASes the lane row against the reserved source tuple (session, generation, PID, process-start; mismatch → `STALE_GENERATION`, no write), validates evidence (mode, card/SPEC, cwd == worktree root == target, branch, HEAD == pin, new non-provisional unused UUID, current owner, headless thread id == recorded relocation), NACKs mismatched evidence in the same transaction, and otherwise writes: tombstone → peer row CAS update to generation+1 → handoff `BOUND` + event → `lane_handoff_receipts` row → dispatch release (`lane_message_releases` per message + `lane_dispatch_releases` marker). A BOUND handoff retried with the same evidence returns the same receipt and writes nothing. `StaleEndpointError{Code, Current}` carries non-secret redirect metadata (slot, session UUID, generation; a provisional token is redacted). `AuthorizeCardWrite` admits code writes only for the endpoint the card's BOUND receipt names.
- `internal/factorymsg/store.go`: `verifyPeer` classifies a non-current identity as `STALE_ENDPOINT` (tombstoned) or `STALE_GENERATION` (current session, older generation), else the unchanged t1074 error; `Claim` (inside its transaction), `ReadBody`, `RecordDisposition`, `Receipt` refuse with `ENDPOINT_HANDOFF_PENDING` while the peer's slot has a non-final handoff (pre-check plus a `NOT EXISTS` clause in the statement); a claim mismatch on a released pre-handoff token is `STALE_GENERATION`; `Send`'s idempotency collision compares a released envelope against its original recipient. Schema gains three tables (same broker DB). Test seam `bindStep`.
- `internal/factorymsg/handoff.go`: `HandoffNackReason` also reports a `StaleEndpointError` code.
- `internal/hook/factory_handoff_bind.go` (new) + one call in `factory_messages.go`: on SessionStart, a session that is not the lane's current endpoint, with a `SWITCH_PENDING_INTERACTIVE` handoff on the slot, is bound through `BindHandoff` from a Git readback of the SessionStart cwd (canonical cwd, toplevel, branch, HEAD) and the t1074 owner identity. No handoff / WT_READY / headless → t1074 path unchanged.
- `internal/cli/factory_lane_handoff_bind.go` (new): `bindLaneHandoffHeadless` — recorded relocation thread id + fresh `readbackHandoffTarget` → `BindHandoff`. Added to the AC-FLH-014 control-file scan list.

#### RED evidence (verbatim, captured before any M3a implementation)

HEAD `b08569bac`, log `.moai/reports/t1082/run-m3a-red.log`, command `go test ./internal/factorymsg ./internal/hook ./internal/cli -run '^(TestFactoryLaneHandoffInteractiveStateMachine|TestFactoryLaneHandoffHeadlessAppServerStateMachine|TestFactoryLaneHandoffAtomicModeEvidenceRebind|TestFactoryLaneHandoffDispatchAfterBound|TestFactoryLaneHandoffStaleEndpointRejected|TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch)$' -count=1 -timeout=180s`, exit 1:

```text
# github.com/modu-ai/moai-adk/internal/factorymsg [github.com/modu-ai/moai-adk/internal/factorymsg.test]
internal/factorymsg/handoff_bind_test.go:118:34: undefined: HandoffBindEvidence
internal/factorymsg/handoff_bind_test.go:119:9: undefined: HandoffBindEvidence
internal/factorymsg/handoff_bind_test.go:187:15: undefined: StaleEndpoint
internal/factorymsg/handoff_bind_test.go:212:9: f.s.bindStep undefined (type *Store has no field or method bindStep)
internal/factorymsg/handoff_bind_test.go:218:19: f.s.BindHandoff undefined (type *Store has no field or method BindHandoff)
internal/factorymsg/handoff_bind_test.go:233:8: f.s.bindStep undefined (type *Store has no field or method bindStep)
internal/factorymsg/handoff_bind_test.go:234:18: f.s.BindHandoff undefined (type *Store has no field or method BindHandoff)
internal/factorymsg/handoff_bind_test.go:268:22: f.s.BindHandoff undefined (type *Store has no field or method BindHandoff)
internal/factorymsg/handoff_bind_test.go:278:31: undefined: HandoffBindEvidence
internal/factorymsg/handoff_bind_test.go:281:51: undefined: HandoffBindEvidence
internal/factorymsg/handoff_bind_test.go:281:51: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg [build failed]
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/factory_lane_handoff_bind_test.go:15:34: undefined: laneHandoffBind
internal/cli/factory_lane_handoff_bind_test.go:21:9: undefined: laneHandoffBind
internal/cli/factory_lane_handoff_bind_test.go:24:92: undefined: factorymsg.HandoffBinding
internal/cli/factory_lane_handoff_bind_test.go:71:14: undefined: bindLaneHandoffHeadless
internal/cli/factory_lane_handoff_bind_test.go:97:17: undefined: bindLaneHandoffHeadless
internal/cli/factory_lane_handoff_bind_test.go:121:12: undefined: bindLaneHandoffHeadless
internal/cli/factory_lane_handoff_bind_test.go:141:12: undefined: bindLaneHandoffHeadless
--- FAIL: TestFactoryLaneHandoffInteractiveStateMachine (3.59s)
    --- FAIL: TestFactoryLaneHandoffInteractiveStateMachine/pending_until_next_turn_then_bound (0.73s)
        factory_handoff_bind_test.go:204: SELECT count(*) FROM lane_handoff_receipts: SQL logic error: no such table: lane_handoff_receipts (1)
    --- FAIL: TestFactoryLaneHandoffInteractiveStateMachine/no_evidence_before_the_cd_guidance (0.70s)
        factory_handoff_bind_test.go:268: SELECT count(*) FROM lane_handoff_receipts: SQL logic error: no such table: lane_handoff_receipts (1)
    --- FAIL: TestFactoryLaneHandoffInteractiveStateMachine/mismatch/branch_moved (0.54s)
        factory_handoff_bind_test.go:296: mismatched SessionStart notice = ""
    --- FAIL: TestFactoryLaneHandoffInteractiveStateMachine/mismatch/head_moved (0.72s)
        factory_handoff_bind_test.go:296: mismatched SessionStart notice = ""
    --- FAIL: TestFactoryLaneHandoffInteractiveStateMachine/mismatch/cwd_not_target (0.49s)
        factory_handoff_bind_test.go:296: mismatched SessionStart notice = ""
    --- FAIL: TestFactoryLaneHandoffInteractiveStateMachine/mismatch/cwd_subdirectory (0.41s)
        factory_handoff_bind_test.go:296: mismatched SessionStart notice = ""
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	4.304s
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

RED reasons: rebind API absent (factorymsg, cli — compile level); the hook ignored the next-turn SessionStart (empty notice) and the receipt table did not exist (hook — runtime level).

Second RED, mid-GREEN (sender-layer AC-FLH-008): after the first `BindHandoff` implementation (in-place release), `go test ./internal/factorymsg ...` failed `handoff_bind_test.go:628: K1 with a different recipient was merged into envelope 93c1c5348fee599e386a8db9c7b39b79` — the release re-addresses K1, so `Send`'s dedupe merged a K1 reuse toward the rebound endpoint. Fixed by comparing a released envelope against its original recipient.

Third RED (legacy broker): `TestClaimOnBrokerWithoutHandoffTables`, log `.moai/reports/t1082/run-m3a-red-legacy-broker.log`, exit 1: `handoff_bind_test.go:728: claim on a pre-handoff broker = [] err=SQL logic error: no such table: lane_handoffs (1)` — the hook hot path opens the broker without schema setup; the new claim gate failed on a broker created before the handoff tables. Fixed: an absent table reads as no handoff.

**Mutant checks (guards observed red; each applied, run, reverted):**

| Mutant | Command | Result | Log |
|---|---|---|---|
| rebind without the source-tuple CAS | `go test ./internal/factorymsg -run '^TestFactoryLaneHandoffAtomicModeEvidenceRebind$'` | FAIL `interactive/source_tuple_moved` + `headless/source_tuple_moved`: `err=<nil>, want a stale-endpoint NACK STALE_GENERATION`, exit 1 | `run-m3a-mutant-no-cas.log` |
| `Claim` without the pre-BOUND gate | `go test ./internal/factorymsg -run '^TestFactoryLaneHandoffDispatchAfterBound$'` | FAIL `claim before BOUND: err=<nil>, want NACK ENDPOINT_HANDOFF_PENDING`, exit 1 | `run-m3a-mutant-no-claim-gate.log` |
| `verifyPeer` without stale classification | `go test ./internal/factorymsg -run '^(TestFactoryLaneHandoffStaleEndpointRejected\|TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch)$'` | FAIL on every old-endpoint op: `err=stale or unregistered peer, want a stale-endpoint NACK STALE_ENDPOINT`, exit 1 | `run-m3a-mutant-no-stale-class.log` |
| `Send` dedupe ignores the original recipient | `go test ./internal/factorymsg -run '^TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch$'` | FAIL `K1 with a different recipient was merged into envelope 7e0f…`, exit 1 | `run-m3a-mutant-send-merges-released.log` |

#### GREEN evidence (final tree, scrubbed env: kanban/factory/provider vars unset in the same invocation)

| Check | Command | Result |
|---|---|---|
| race, factorymsg | `go test -race -count=1 -coverprofile=/tmp/t1082-m3a-factorymsg.cov ./internal/factorymsg/...` | `ok … coverage: 77.1% of statements`, exit 0 (`run-m3a-race-factorymsg.log`) |
| race, hook (full package) | `go test -race -count=1 -coverprofile=/tmp/t1082-m3a-hook.cov ./internal/hook` | `ok … 422.845s coverage: 85.6%`, exit 0, no DATA RACE (`run-m3a-race-hook.log`) |
| race, cli subset | `go test -race -count=1 ./internal/cli -run '^(TestFactoryLaneHandoff.*\|TestLaneHandoff.*\|TestSessionWorktree.*\|TestCodex.*\|TestMoaiMCPServer_RegistrationMatchesCatalog\|TestFactoryMsgStatusReadOnlyRoster\|TestFactoryLeadNoticeUsesOperationalStatus\|TestFactoryMsg.*)$' -v` (no `MOAI_HOME`) | exit 0; 898 PASS, 0 FAIL, 5 SKIP (pre-existing env-gated `TestCodexLive_*`), 0 DATA RACE (`run-m3a-race-cli.log`) |
| vet | `go vet ./internal/factorymsg/... ./internal/hook/... ./internal/cli/...` | exit 0 |
| lint | `golangci-lint run ./internal/factorymsg/... ./internal/hook/... ./internal/cli/...` | `0 issues.`, exit 0 |
| build | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 / exit 0 |
| gofmt | `gofmt -l internal/factorymsg internal/hook internal/cli` | empty |

#### AC matrix for M3a (acceptance.md commands verbatim, prefixed with the lane env scrub)

| AC | Status | Named test (package) | Gate output / exit | Evidence |
|---|---|---|---|---|
| AC-FLH-003 | PASS | `TestFactoryLaneHandoffInteractiveStateMachine` (hook) | `true`, exit 0 | `ac03.jsonl` — pending until next turn then BOUND; no evidence before `/cd` guidance; 4 mismatch NACKs |
| AC-FLH-004 | PASS | `TestFactoryLaneHandoffHeadlessAppServerStateMachine` (cli) | `true`, exit 0 (re-run on final tree) | `ac04.jsonl` — 9 subtests: fork, start, 3 non-idle NACKs, readback mismatch, unreadable target, interactive refused, no relocation evidence |
| AC-FLH-005 | PASS | `TestFactoryLaneHandoffAtomicModeEvidenceRebind` (factorymsg) | `true`, exit 0 | `ac05.jsonl` — both modes: failpoint at `locked`/5 writes all-absent, all-present, idempotent retry, 13–14 mismatch NACKs, stale nonce, CAS loser |
| AC-FLH-006 | PASS | `TestFactoryLaneHandoffDispatchAfterBound` (factorymsg) | `true`, exit 0 | `ac06.jsonl` |
| AC-FLH-007 | PASS | `TestFactoryLaneHandoffStaleEndpointRejected` (factorymsg) | `true`, exit 0 | `ac07.jsonl` — 11 stale ops with byte-level no-mutation snapshot, restart survival, current ops succeed |
| AC-FLH-008 | PASS | `TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch` (factorymsg) | `true`, exit 0 | `ac08.jsonl` |

Regressions (acceptance.md verbatim, final tree): AC-FLH-001 `true`, AC-FLH-002 `true`, AC-FLH-014 `true`, AC-FLH-016 `true`, AC-FLH-017 `true`, AC-FLH-019 (order vii only, `-race`) `true` — each exit 0. AC-FLH-019 is still NOT a PASS (orders (i)–(vi) are M3b).

**t1074 hook tests ×5 (flakiness re-check, §J).** `MOAI_HOME=/tmp/t1082-m3a-hookx5-home go test -count=5 ./internal/hook -run '^(TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer|TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding)$' -v` → 15/15 PASS, 0 FAIL, exit 0; durations 0.60–1.01s (`run-m3a-hook-x5.log`). The M1 flake did not recur in this measurement; still not established either way.

#### Coverage (file-level, statement-weighted, from the final-tree profiles)

| File | Profile | Covered |
|---|---|---|
| `internal/factorymsg/handoff_bind.go` (new) | `/tmp/t1082-m3a-factorymsg.cov` | 136/160 = 85.0% |
| `internal/factorymsg/store.go` (t1074 file, changed) | same | 346/489 = 70.8% (M2 profile `run-m2-factorymsg.cov`: 305/473 = 64.5%) |
| `internal/factorymsg/handoff.go` (changed) | same | 88/102 = 86.3% |
| `internal/hook/factory_handoff_bind.go` (new) | `/tmp/t1082-m3a-hook.cov` | 31/36 = 86.1% |
| `internal/hook/factory_messages.go` (t1074 file, 3 lines changed) | same | 70/83 = 84.3% |
| `internal/cli/factory_lane_handoff_bind.go` (new) | `/tmp/t1082-m3a-cli.cov` | 14/16 = 87.5% |
| `internal/cli/factory_lane_handoff.go` | same | 80/100 = 80.0% — not touched by M3a; unchanged |

Uncovered lines in `handoff_bind.go` are DB-error returns and a commit-ACK-loss read of a missing release marker.

#### Gaps

- AC-FLH-018 and AC-FLH-019 orders (i)–(vi), the 200 unforced iterations, and REQ-FLH-017/018 inside `RegisterPeer`: M3b. NOT_RUN.
- LIVE (AC-FLH-012/013) and the real Codex `/cd` SessionStart ordering: M5. NOT_RUN.
- AC-FLH-008 recipient layer: the test models "repeated delivery" as a lease redelivery before the receipt (delivery 1 executes and records `accepted`, the lease lapses, delivery 2 records `duplicate` and is the one successful receipt). The final disposition is therefore `duplicate`, not `accepted`; "one accepted receipt" is read as "exactly one receipt the broker accepted". Under the current basis no second envelope with the same key can exist, so no other shape of repeated delivery is reachable. Reviewer judgement needed on this reading.
- AC-FLH-007 "stale claim token" is tested both ways: the old endpoint with its token (`STALE_ENDPOINT` at `verifyPeer`) and the current endpoint presenting a released pre-handoff token (`STALE_GENERATION`). A token superseded by a lease redelivery (not by a handoff) keeps the t1074 "claim identity mismatch".
- Headless owner identity: `bindLaneHandoffHeadless` takes the owner PID/process-start from its caller; there is no production caller yet, and M2's relocation client closes its app-server process after the RPC, so which process owns a headless endpoint in production is not settled.
- `TestLaneHandoffBindRefusalsAndRedirects`, `TestClaimOnBrokerWithoutHandoffTables` (had its own RED), and the `target_unreadable_at_bind` / `interactive_handoff_refused` subtests of AC-FLH-004 were added after GREEN to raise coverage; the first and the two subtests had no individual RED.
- Not measured: full `internal/cli` suite (CI owns it).

#### Residual risks

- The hook binds from whatever SessionStart arrives with a non-current session while the handoff is `SWITCH_PENDING_INTERACTIVE`. Until M3b, a UserPromptSubmit of the post-`/cd` session that arrives before that SessionStart still goes through unmodified t1074 `RegisterPeer`, which rotates the endpoint without tombstone or receipt; the later SessionStart then finds a current session and the handoff stays pending. That is exactly REQ-FLH-018's case and is M3b's job.
- A mismatched SessionStart NACKs the handoff; a SessionStart from an unrelated session carrying the lane's factory env would do the same. Recovery is a fresh reservation.
- Pre-BOUND refusal of `Claim` also stops receipt-bearing and status messages for that lane; the hook batch reports `degraded` for the duration.
- `ReadBody`/`RecordDisposition`/`Receipt` make one extra read per call (the pending check); `verifyPeer`'s failure path makes up to two extra reads.

#### What M3b must know

- **Transaction shape.** `BindHandoff` reads, in order and inside one `BEGIN IMMEDIATE`: the handoff row by id → nonce check → BOUND short-circuit → pending-state check → the slot's peer row (CAS against `lane_handoffs.source_*`) → `bindStep("locked")` → evidence validation (NACK written in the same transaction) → writes, each followed by `bindStep`: `tombstone`, `peer`, `bound`, `receipt`, `release` → commit. Every refusal before `locked` writes nothing.
- **Seams.** `Store.bindStep func(step string) error` (unexported, set from package `factorymsg` tests) fires after the rebind holds the write lock (`"locked"`) and after each write; a barrier there holds R's transaction open for AC-FLH-019 order (ii). The REQ-FLH-017/018 seams belong in `RegisterPeer` and are not added here.
- **Refusal vocabulary already in place.** `NackStaleGeneration`, `NackStaleEndpoint`, `NackEndpointHandoffPending` constants exist; `HandoffNackReason` returns the code for both `HandoffNackError` and `StaleEndpointError`. A launcher registration that finalizes a handoff should move it to `NACK`/`STALE_GENERATION`; `BindHandoff` then refuses it with `STALE_GENERATION` and writes nothing (covered by `finalized_handoff_is_stale`).
- **REQ-FLH-018 `STALE_ENDPOINT` on the tombstoned source UUID.** `lane_endpoint_tombstones` has `(slot, session_uuid, generation)`; `RegisterPeer` can look up by `session_uuid` inside its transaction.
- **Open-state predicate.** `openHandoffStates` (SQL fragment) and `handoffPending(ctx, q, slot)` accept a `*sql.Tx`; reuse them inside `RegisterPeer` so the read happens after `BEGIN`.

### M3b — RegisterPeer serialized with the handoff rebind (2026-09-23, manager-develop, cycle_type=tdd)

**Scope delivered.** Base `2c22fb10a`. Commits: `eaeea53f4` (implementation + AC tests), `a26804e48` (step-abort / legacy-broker / handle-probe tests), `6c24dd493` (fix: handoff check moved ahead of the live-owner rule, with its RED test), `409ee64b8` (fail-closed store-error tests). Code HEAD for every measurement below is `409ee64b8` unless a row says otherwise.

- `internal/factorymsg/store.go` — `RegisterPeer` (t1074) gains two t1082 calls inside its own write transaction. A **turn registration** (session UUID not launch-pending) calls `refuseTurnRegistrationDuringHandoff` right after the slot-row read, before the t1074 live-owner rule and the row write. A **launcher provisional registration** (`isLaunchPendingSession(p.SessionUUID)`, the path `RegisterLaunchPending` takes) calls `finalizeHandoffOnLauncherRegistration` after the slot-row write and before commit. A live-owner rejection returns before the write, so it never reaches the finalize step and leaves the handoff unchanged. No new caller flag: the launch-pending session prefix already separates the two paths.
- `internal/factorymsg/handoff_bind.go` — `refuseTurnRegistrationDuringHandoff`: `sessionTombstoned(ctx, tx, uuid)` → `STALE_ENDPOINT` (redirect via `staleError(ctx, tx, …)`); `handoffPending(ctx, tx, slot)` → `ENDPOINT_HANDOFF_PENDING`. This applies even when the PID and process-start equal the owner's. `finalizeHandoffOnLauncherRegistration`: every open handoff on the slot → `NACK`/`STALE_GENERATION` plus an event row. It writes no tombstone, receipt or release. Both reads tolerate a broker created before the handoff tables (`no such table` → none). Race instrumentation: `WithStepHook(ctx, fn)` step observer (fires only while the write lock is held; an error aborts that transaction), step names `StepBindBegun` / `StepReserveInserted` / `StepRegisterHandoffRead` / `StepRegisterFinalize`, and `Store.HandleStats()` / `SharesHandle(a, b)` for the separate-handle proof from package `hook`. `Store.step` now takes ctx; M3a's `bindStep` seam and its step list are unchanged.
- `internal/factorymsg/handoff.go` — `ReserveHandoff` fires `StepReserveInserted` after its RESERVED insert, before commit.
- `internal/hook/factory_messages.go` + `factory_handoff_bind.go` — `registerFactoryHookPeer` maps the two refusals to named notices: `factory handoff pending ENDPOINT_HANDOFF_PENDING: slot=…; … endpoint is unchanged` and `factory endpoint replaced STALE_ENDPOINT: slot=… is current at <session> generation <g>; …`. The hook stays fail-open, and the endpoint is not rotated because the store refused.

**Lead constraint (t1100 / t1112) — compliance.** `staleOrUnregistered` is unchanged and has no new callers. The REQ-FLH-018 `STALE_ENDPOINT` check queries `lane_endpoint_tombstones` through the `*sql.Tx` that `RegisterPeer` already holds (`sessionTombstoned(ctx, tx, …)`), and its redirect is built with `staleError(ctx, tx, …)`, also on the tx. The code is local to the new t1082 functions, and nothing queries `s.db` inside the transaction. `Send` and its idempotency re-lookup (`lane_message_releases` path) are untouched: `git diff 2c22fb10a a26804e48 -- internal/factorymsg/store.go` shows only three hunks — the `RegisterPeer` doc/MX comment, the `launcher` local with the turn-registration call, and the finalize call. `6c24dd493` moves the turn-registration call up; it touches nothing else in `store.go`.

#### RED evidence (verbatim, captured before the RegisterPeer logic existed)

First RED, compile level (tests written first), HEAD `2c22fb10a`, `go test ./internal/factorymsg -run '^TestFactoryLaneHandoffRebindVsUserPromptRegisterRace$' -count=1 -timeout=180s`, exit 1 (`run-m3b-red1.log`):

```text
# github.com/modu-ai/moai-adk/internal/factorymsg [github.com/modu-ai/moai-adk/internal/factorymsg.test]
internal/factorymsg/lane_handoff_race_test.go:91:51: undefined: WithStepHook
internal/factorymsg/lane_handoff_race_test.go:322:24: undefined: StepRegisterHandoffRead
internal/factorymsg/lane_handoff_race_test.go:475:24: undefined: StepReserveInserted
internal/factorymsg/lane_handoff_race_test.go:514:24: undefined: StepReserveInserted
internal/factorymsg/lane_handoff_race_test.go:585:24: undefined: StepRegisterFinalize
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg [build failed]
FAIL
```

Second RED, behavioral: after adding only the seam instrumentation (step names fired at the future read points, no handoff logic). Same command, exit 1 (`run-m3b-red2.log`):

```text
--- FAIL: TestFactoryLaneHandoffRebindVsUserPromptRegisterRace (8.74s)
    lane_handoff_race_test.go:312: AC_FLH_019_ORDERS_COVERED=i,ii,iii,iv,v,vi,vii,unforced200
    --- FAIL: TestFactoryLaneHandoffRebindVsUserPromptRegisterRace/order_i_registration_holds_first_then_rebind (0.32s)
        lane_handoff_race_test.go:340: R_blocked_observed=true
        lane_handoff_race_test.go:345: err=<nil>, want NACK ENDPOINT_HANDOFF_PENDING
    --- FAIL: TestFactoryLaneHandoffRebindVsUserPromptRegisterRace/order_ii_rebind_holds_first_then_registrations (0.37s)
        lane_handoff_race_test.go:387: H_blocked_observed=true
        lane_handoff_race_test.go:414: err=<nil>, want NACK STALE_ENDPOINT
    --- FAIL: TestFactoryLaneHandoffRebindVsUserPromptRegisterRace/order_iv_registration_reads_handoff_inside_its_transaction (0.42s)
        lane_handoff_race_test.go:491: H_blocked_observed=true
        lane_handoff_race_test.go:498: V_commit_at=2026-09-23T13:34:43.492678+09:00 H_return_at=2026-09-23T13:34:43.526105+09:00
        lane_handoff_race_test.go:502: err=<nil>, want NACK ENDPOINT_HANDOFF_PENDING
    --- FAIL: TestFactoryLaneHandoffRebindVsUserPromptRegisterRace/order_v_launcher_registration_finalizes_handoff_in_its_commit (0.40s)
        lane_handoff_race_test.go:530: A_blocked_observed=true
        lane_handoff_race_test.go:537: V_commit_at=2026-09-23T13:34:43.8862+09:00 A_return_at=2026-09-23T13:34:43.921228+09:00
        lane_handoff_race_test.go:549: handoff = RESERVED/ while the row is launch-pending, want NACK/STALE_GENERATION from A's commit
    --- FAIL: TestFactoryLaneHandoffRebindVsUserPromptRegisterRace/unforced_200_iterations (6.32s)
        lane_handoff_race_test.go:631: iteration 31: R rebind: stale or unregistered peer: STALE_GENERATION; lane lane-1 is current at post-cd-uuid generation 3
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg	9.023s
FAIL
```

RED reasons, which are the defect REQ-FLH-017/018 describes: during a non-final handoff, the post-`/cd` UserPromptSubmit registration rotated the endpoint without a tombstone or receipt ((i), (iv), and unforced, where the rotation then made R's CAS fail). The tombstoned `src-uuid` was re-accepted (ii). The launcher registration committed a launch-pending row beside a still-`RESERVED` handoff (v). Orders (iii), (vi), and (vii) passed at this point: (iii) and (vi) are t1074 controls, and (vii)'s reservation-side read was already inside the transaction since M1.

Third RED, hook surface: `go test ./internal/hook -run '^TestFactoryUserPromptRegistrationSurfacesHandoffRefusals$' -count=1`, exit 1 (`run-m3b-red-hook-surface.log`): `factory_handoff_race_test.go:720: pending-handoff notice = "factory messaging degraded: handoff NACK ENDPOINT_HANDOFF_PENDING: lane-1"`. The refusal surfaced only as a generic degraded message.

Fourth RED, check order (found while writing this record: the diff showed the turn-registration check sitting *after* the t1074 live-owner rule), HEAD `a26804e48` + new test, `go test ./internal/factorymsg -run '^TestTurnRegistrationDuringHandoffIsPendingForAnyIdentity$' -count=1`, exit 1 (`run-m3b-red-order-before-live-owner.log`):

```text
--- FAIL: TestTurnRegistrationDuringHandoffIsPendingForAnyIdentity (0.04s)
    handoff_register_test.go:59: err=factory logical lane has a live owner, want NACK ENDPOINT_HANDOFF_PENDING
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg	0.307s
FAIL
```

REQ-FLH-018 names `ENDPOINT_HANDOFF_PENDING` for every turn registration during a non-final handoff. A registration with a different identity while the source owner was live got the generic t1074 message instead. The row was unchanged either way, so the gap was the reason code, not a write. Fixed in `6c24dd493`.

**AC-FLH-018 had no fresh RED.** It was written after the REQ-FLH-017 finalize code had already gone GREEN under AC-FLH-019 order (v). Its discrimination is shown by mutants M3 and M4 below instead.

#### Mutant checks (each applied to the GREEN tree, run, then reverted; restoration verified with `cmp`)

| Mutant | Command (tests scoped) | Result | Log |
|---|---|---|---|
| M1 — no in-transaction handoff read for turn registrations | AC-019 + hook surface test | FAIL (i) `H never reached step "register-handoff-read"`; (ii) `err=<nil>, want NACK STALE_ENDPOINT`; (iv) `err=<nil>, want NACK ENDPOINT_HANDOFF_PENDING`; unforced `iteration 7: R rebind: … STALE_GENERATION`; hook `pending-handoff notice = "factory messaging bound: …"`; exit 1 | `run-m3b-mutant-no-read.log` |
| M2 — the handoff read moved before `BEGIN` (decision taken from that snapshot) | AC-019 order (iv) | FAIL `H_blocked_observed=true` … `err=<nil>, want NACK ENDPOINT_HANDOFF_PENDING`, exit 1 | `run-m3b-mutant-read-before-begin.log` |
| M3 — launcher finalize removed | AC-018 + AC-019 | FAIL AC-019 (v) `handoff = RESERVED/ while the row is launch-pending`; AC-018 (iii)/(iv)/(v) `handoff … = SWITCH_PENDING_HEADLESS/, want NACK/STALE_GENERATION`, (i) + AC-019 (vii) lose their seam, unforced red; exit 1 | `run-m3b-mutant-no-finalize.log` |
| M4 — the finalize decision taken from a read before `BEGIN` | AC-018 + AC-019 | FAIL AC-019 (v) `handoff = RESERVED/ while the row is launch-pending` (and (vii), whose seam is skipped when the stale read sees no handoff); **AC-018 stays green** — its orders never hold an uncommitted handoff against A, so the read position is discriminated by AC-019 (v) only; exit 1 | `run-m3b-mutant-finalize-read-before-begin.log` |
| M5 — tombstone check removed | AC-019 + hook surface test | FAIL (ii) `err=<nil>, want NACK STALE_ENDPOINT`; hook `stale-endpoint notice = "factory messaging bound: … generation=4 …"`; exit 1 | `run-m3b-mutant-no-tombstone.log` |
| M6 — `no such table` tolerance removed from `sessionTombstoned` | `TestRegisterPeerOnBrokerWithoutHandoffTables` | FAIL `turn registration on a pre-handoff broker = {…} err=SQL logic error: no such table: lane_endpoint_tombstones (1)`, exit 1 | inline (session output) |

#### Fixture defect found and fixed during GREEN (AC-018 unforced)

The first AC-018 GREEN runs landed 200/200 on rebind-first, both with the spawn order alternated and with sub-2 ms random jitter added. Timing each racer alone showed racer A taking ~37 ms against B's <1 ms. A temporary probe inside `RegisterPeer` (reverted; `cmp`-verified) measured `BEGIN` at ~8–45 µs and the whole transaction at <300 µs. So the delay was in the test, not the store: `f.relaunch()` resolved `homestate.ProjectKey(root)` inside the racer goroutine, after the start signal. The fix builds both racers' arguments before the start signal. The unforced block then exercised both branches (38 launcher-first / 162 rebind-first in that run), and the jitter was removed.

#### How RegisterPeer tells the two registrations apart

`isLaunchPendingSession(p.SessionUUID)` is the existing signal. `RegisterLaunchPending` always writes `launch-pending:<id>` through `RegisterPeer`. Every turn-hook registration carries the hook's real session UUID, and `BindLaunchPending` rejects a launch-pending UUID. The `launcher` bool is computed once, after the slot-row read. No caller flag was added.

#### GREEN evidence (code HEAD `409ee64b8`; every command prefixed with the lane env scrub in the same invocation)

| Check | Command | Result |
|---|---|---|
| race, factorymsg | `go test -race -count=1 -timeout=600s -coverprofile=/tmp/t1082-m3b-factorymsg.cov ./internal/factorymsg/...` | `ok … 41.493s coverage: 78.6% of statements`, exit 0, 0 DATA RACE (`run-m3b-race-factorymsg.log`) |
| race, hook (full package) | `go test -race -count=1 -timeout=900s -coverprofile=/tmp/t1082-m3b-hook.cov ./internal/hook` on `6c24dd493` (the `409ee64b8` diff is one factorymsg `_test.go` file, so the hook package and its production dependencies are byte-identical) | `ok … 268.122s coverage: 85.6% of statements`, exit 0, 0 DATA RACE, 0 FAIL (`run-m3b-race-hook.log`). The earlier run on `eaeea53f4` (`run-m3b-race-hook-eaeea53f4.log`, `ok … 265.584s`, 0 DATA RACE) is superseded because `6c24dd493` changed `store.go` |
| race, cli subset | `go test -race -count=1 -timeout=900s -coverprofile=/tmp/t1082-m3b-cli.cov ./internal/cli -run '^(TestFactoryLaneHandoff.*\|TestLaneHandoff.*\|TestSessionWorktree.*\|TestCodex.*\|TestMoaiMCPServer_RegistrationMatchesCatalog\|TestFactoryMsgStatusReadOnlyRoster\|TestFactoryLeadNoticeUsesOperationalStatus\|TestFactoryMsg.*)$' -v` (`MOAI_HOME` unset) | `ok … 79.916s`, exit 0; 912 PASS, 0 FAIL, 5 SKIP (pre-existing env-gated `TestCodexLive_*`), 0 DATA RACE (`run-m3b-race-cli.log`) |
| vet | `go vet ./internal/factorymsg/... ./internal/hook/... ./internal/cli/...` | exit 0 |
| lint | `golangci-lint run ./internal/factorymsg/... ./internal/hook/... ./internal/cli/...` | `0 issues.`, exit 0 (`run-m3b-lint.log`); the first run on `eaeea53f4` flagged one `unused` test helper, removed before commit |
| build | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 / exit 0 |
| windows test compile | `GOOS=windows GOARCH=amd64 go vet ./internal/hook ./internal/factorymsg ./internal/cli` | exit 0. The AC-018 live-owner helper re-executes the test binary (`os.Args[0]`), not a Unix `sleep`, so the hook test also builds on the Windows leg |
| gofmt | `gofmt -l internal/factorymsg internal/hook internal/cli` | empty |

#### AC matrix (acceptance.md commands verbatim, run through the lane env scrub; HEAD `409ee64b8`)

| AC | Status | Named test (package) | Gate output / exit | Evidence |
|---|---|---|---|---|
| AC-FLH-018 | PASS | `TestFactoryLaneHandoffRebindVsLaunchBindRace` (hook) | `true`, exit 0 | `ac18.jsonl` — `--- PASS … (58.21s)`; orders (i) `B_blocked_observed=true` + `RACER_HANDLES_DISTINCT=2`, (ii) `A_blocked_observed=true` + `RACER_HANDLES_DISTINCT=2`, (iii)–(v); `UNFORCED_OUTCOMES launcher_first=151 rebind_first=49` |
| AC-FLH-019 | PASS | `TestFactoryLaneHandoffRebindVsUserPromptRegisterRace` (factorymsg) | `true`, exit 0 | `ac19.jsonl` — `AC_FLH_019_ORDERS_COVERED=i,ii,iii,iv,v,vi,vii,unforced200`; `R_blocked_observed` (i), `H_blocked_observed` (ii, iv), `A_blocked_observed` (v), `V_blocked_observed` (vii), each with `RACER_HANDLES_DISTINCT=2`; `UNFORCED_OUTCOMES i=7 ii=193` |

AC-FLH-019 order (vii) now uses the production launcher registration as adversary A, held at `StepRegisterFinalize` inside `RegisterPeer`. This replaces M1's raw-SQL stand-in, which acceptance.md allowed only until the REQ-FLH-017 step existed.

**Regressions** (acceptance.md verbatim, HEAD `409ee64b8`): AC-FLH-001 `true`, 002 `true`, 003 `true`, 004 `true`, 005 `true`, 006 `true`, 007 `true`, 008 `true`, 014 `true`, 016 `true`, 017 `true` — each exit 0 (`ac01.jsonl` … `ac17.jsonl`).

**t1074 hook tests ×5.** `MOAI_HOME=/tmp/t1082-m3b-hookx5-home go test -count=5 ./internal/hook -run '^(TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer|TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding)$' -v` → 15 PASS, 0 FAIL, exit 0, durations 0.54–0.97 s (`run-m3b-hook-x5.log`). The M1 flake did not recur; still not established either way. `TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding` and `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` are in this set: no handoff exists there, so REQ-FLH-018 does not apply and t1074 alias correction holds.

#### Coverage (file-level, statement-weighted, from the profiles above)

| File | Profile | Covered |
|---|---|---|
| `internal/factorymsg/handoff_bind.go` (t1082) | factorymsg | 181/209 = 86.6% (M3a 136/160 = 85.0%) |
| `internal/factorymsg/handoff.go` (t1082) | factorymsg | 91/104 = 87.5% |
| `internal/factorymsg/store.go` (t1074 file) | factorymsg | 356/496 = 71.8%; the M3b-added lines (`launcher`, both calls) are all covered. Uncovered blocks are pre-existing t1074 error returns |
| `internal/hook/factory_handoff_bind.go` (t1082) | hook | 35/41 = 85.4% |
| `internal/hook/factory_messages.go` (t1074 file, 3 lines changed) | hook | 72/85 = 84.7% (M3a 70/83); the new notice branch is covered |
| `internal/cli/factory_lane_handoff.go` (t1082, M1) | cli (`/tmp/t1082-m3b-cli.cov`) | **95/100 = 95.0%** (was 80/100 = 80.0%; `createHandoffTarget` fully covered) |
| `internal/cli/factory_lane_handoff_bind.go` (t1082) | cli | 14/16 = 87.5% (unchanged) |
| `internal/cli/factory_lane_handoff_switch.go` (t1082) | cli | 68/76 = 89.5% (unchanged) |

`factory_lane_handoff.go` was raised by `internal/cli/factory_lane_handoff_create_test.go`: admission refusals (unknown activity, invalid slug, no active run with default deps, missing develop pin, cwd = target with no earlier handoff) and every `createHandoffTarget` provenance refusal (materializer error, other path, missing card branch, develop moved after creation, detached HEAD, branch checked out twice, dirty target), each via a one-defect fake materializer. The five remaining statements have no reachable trigger short of a broken filesystem or database: `filepath.Abs` fails only if `Getwd` fails; `factorymsg.Open`, `HandoffsForLane` (×2) and `NackHandoff` fail only on broker I/O after a successful open.

#### Gaps

- LIVE (AC-FLH-012/013) and the real Codex `/cd` SessionStart ordering: M5. NOT_RUN.
- AC-FLH-018 uses a **headless** handoff. The AC says "`SWITCH_PENDING`" without a mode. With an interactive handoff, order (iv)'s relaunched SessionStart would reach the M3a interactive binder, which NACKs the handoff on the cwd readback mismatch (`TARGET_READBACK_MISMATCH`) before A registers. The expected `NACK`/`STALE_GENERATION`-by-A would then be unreachable for a reason outside REQ-FLH-017. The interactive-relaunch shape is not covered by a race test.
- AC-FLH-018 had no fresh RED (the finalize code was driven by AC-019 (v)); its discrimination is shown by mutants M3/M4 only. Under M4 (finalize read before `BEGIN`), AC-018 stays green — only AC-019 (v) catches that mutant.
- Tests added after GREEN without an individual RED: `TestLaneHandoffPreparationRefusals`, `TestLaneHandoffCreateTargetProvenance` (both pin M1 behavior for coverage), `TestRegisterPeerOnBrokerWithoutHandoffTables` (red under M6), `TestHandoffStepHookAbortsItsTransaction`, `TestRacerHandleProbes`, `TestRegistrationHandoffStoreErrorsFailClosed`.
- The unforced outcome split is a single observation per run (AC-018 151/49, AC-019 7/193). No distribution is asserted; the AC does not require one.
- Not measured: full `internal/cli` suite (CI owns it).

#### Residual risks

- **A turn registration is now refused for the whole non-final window, whatever its identity.** An interactive lane whose operator never runs `/cd` keeps a pending handoff, and every later UserPromptSubmit from a new session (a restart that did not go through the launcher, e.g. a resumed session) is refused with `ENDPOINT_HANDOFF_PENDING` until the handoff is NACKed or abandoned. The launcher path finalizes the handoff (REQ-FLH-017); a non-launcher restart has no such path. This is the specified behavior, not a defect, but no timeout or abandon trigger for a stuck `SWITCH_PENDING_INTERACTIVE` exists yet.
- A tombstoned session UUID is refused permanently (`STALE_ENDPOINT`), including a legitimate resume of the old session after the lane moved.
- `WithStepHook`, `HandleStats` and `SharesHandle` are exported for the hook-package race test. Production callers never pass a step hook; a caller that did could abort handoff transactions.
- The live-owner helper in the AC-018 test is a re-executed test binary killed and reaped by `t.Cleanup`. If the test process crashes before cleanup runs, the helper is expected to see EOF on its stdin pipe and exit. That path was not observed.

## §J Lead follow-ups after M1 (2026-09-23)

### Hook-test flakiness — not established

Two measurements disagree; the defect is **not established**. Re-check when M3 re-runs the hook path; report with failing output if it recurs.

| Who | Tree | Command | When | Result |
|---|---|---|---|---|
| manager-develop (M1) | base `92c932cac` factorymsg (per its note) | not recorded in the log — the log holds output only (`.moai/reports/t1082/run-m1-baseline-hook-HEAD-x5.log`) | log timestamps ~10:13 | `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` + `TestFactoryBoundUserPromptSubmitDoesNotRewritePeer` 3/5 FAIL `factory endpoint is launch-pending`; with M1 1/5. Test durations in that log ~2.3s |
| lane orchestrator | M1 HEAD `17f7b89e2` | `unset <kanban/provider vars> && MOAI_HOME=<scratch>/home2 go test -count=5 ./internal/hook -run '^(TestFactoryUserPromptSubmitRebindsLaunchPendingPeer\|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer\|TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding)$' -v` | after M1 return | 15/15 PASS, durations 0.57–1.04s |

Duration gap (~2.3s vs ~0.6–1.0s) suggests load-dependent timing, not verified.

### factorymsg coverage baseline

| Tree | Command | Coverage |
|---|---|---|
| pre-M1 `92c932cac` (exported via `git archive` to scratch) | `go test -count=1 -cover ./internal/factorymsg/` | 64.5% |
| post-M1 `32b384917` | `go test -count=1 -coverprofile=<scratch>/m1.cov ./internal/factorymsg/` | 68.0% |
| card-added `internal/factorymsg/handoff.go` | from the same profile, statement-weighted | 85.3% (81/95) |
| card-added `internal/cli/factory_lane_handoff.go` | `go test -count=1 -run 'LaneHandoff' -coverprofile=… ./internal/cli/` | 80.0% (80/100); lowest `createHandoffTarget` 61.9% |

M1 raised package coverage (+3.5pp); the package-level shortfall vs 85% predates this card. Whether card-added code reaches 85% is judged before run closes.

### AC-FLH-008 narrowed (lead option a)

SPEC revision `745ae0e6d` (v0.5.2): same-key duplicate only within the same recipient generation; resend after BOUND uses a new key; schema unchanged; sender-side Send vs recipient-side `DispositionDuplicate` separated. Evidence `.moai/reports/t1082/ac008-idempotency-check.md`. Plan-audit hash cache invalidated by this body edit; no re-audit requested.

## §K Lane record after M2 (2026-09-23)

### Coverage, side by side (lead request)

| Tree | factorymsg package | Card-added files |
|---|---|---|
| pre-M1 `92c932cac` (git archive → scratch, `go test -cover ./internal/factorymsg/`) | 64.5% | — |
| post-M1 `32b384917` | 68.0% | handoff.go 85.3% · factory_lane_handoff.go 80.0% |
| post-M2 `b56b41898` (manager-develop measurement, M2 section) | 70.4% | handoff.go 86.0% · handoff_relocation.go 95.8% · factory_lane_handoff_switch.go 86.8% · mcp_codex.go card-added funcs 89.1% · factory_lane_handoff.go 80.0% |

Card-added file below 85%: `internal/cli/factory_lane_handoff.go` (80.0%; lowest func `createHandoffTarget` 61.9%). Raise it or record the reason before run closes. The package-level shortfall predates this card.

### M2/M3 boundary (lane decision, option A)

AC-FLH-003 and AC-FLH-004 both require observing BOUND, and BOUND exists only in M3's atomic rebind (REQ-FLH-008, design.md §6). plan.md moved the headless BOUND and both named tests to M3 at `3c6725fbb`; AC bodies unchanged. Rationale recorded in the local report `.moai/reports/t1082/ac008-idempotency-check.md` (reports are local, not tracked). M3's checklist starts with AC-FLH-003 and AC-FLH-004.

### Idempotency-basis neutrality

SPEC `1e753b7eb` (v0.5.4): statements asserting the key basis is unchanged were replaced by "t1082 does not change the idempotency basis; the basis is owned by t1100". AC-FLH-008 now holds under either basis, and its fixtures must not rely on `sender_session` being in the uniqueness key. The earlier line in this file saying "schema unchanged" (AC-FLH-008 narrowed entry) is historical for v0.5.2.

### Reports untracked

`.moai/reports/t1082/` (48 files) untracked at `db6529b95`: `.moai/reports/*` is ignored on purpose. Reports stay local and are exported to the primary checkout at close, verified with `diff -r`.

## Develop absorb + t1112 alignment (2026-09-23)

State: `git merge --no-ff develop` (MERGE_HEAD `52a486635`, contains t1112 `987fa3e73`) onto `b34625050`. Blocked on one lead decision (below); resolved by lead decision (A), then committed — merge SHA recorded in § Merge commit and merge-tree.

### Conflict resolution
One hunk, `verifyPeerOn` mismatch branch → `return s.staleOrUnregistered(ctx, q, p)`.
- `staleOrUnregistered(ctx, q queryer, p)` runs every query (incl. both `staleError` calls) on the caller's handle; final fallback returns `ErrStalePeer`.
- `rowQuerier` (t1082) and `queryer` (t1112) had the identical shape; `rowQuerier` removed, all uses renamed to `queryer`.
- `(*StaleEndpointError).Is(target) bool { return target == ErrStalePeer }` — the hook t1112 documented on `ErrStalePeer`. `errors.As` / `StaleEndpoint(err)` unchanged.
- `Send` idempotency re-lookup: the `lane_message_releases` original-recipient read extracted into `(*Store).originalRecipient` with identical semantics (Scan on no-row leaves the envelope's recipient). Base lookup untouched for t1100.
- No new stale-check callers.

### RED → GREEN
New tests `internal/factorymsg/stale_align_test.go`: `TestStaleEndpointErrorJoinsErrStalePeerClass`, `TestVerifyPeerOnStaleInsideOpenTxKeepsRedirect`.
RED against the unaligned resolution (hunk = HEAD side, `staleOrUnregistered` on `s.db`), `.moai/reports/t1082/align-red.txt`:
- `errors.Is(stale or unregistered peer: STALE_GENERATION; lane agent-1 is current at b generation 1, ErrStalePeer)=false`
- `stale generation: verifyPeerOn(tx) err=stale or unregistered peer after 2.005905209s — waited on the pool`
GREEN after alignment: both pass; t1112 `TestVerifyPeerOnRunsInsideOpenTx` passes (it was RED on the unaligned resolution: `took 2.001048292s`).

### Blocker — t1112 exact-text pins
Still red after alignment (`.moai/reports/t1082/align-race-factorymsg.txt`): `TestCharacterizeVerifyPeerOutcomes/wrong_generation` and `TestStalePeerOutcomeMatchesErrStalePeer` (`stale_peer_seam_test.go:27`) pin `err.Error() == "stale or unregistered peer"` for a same-session wrong-generation peer. Under t1082 that case is a STALE_GENERATION redirect whose text carries the current endpoint. The two are incompatible; `errors.Is` holds in both tests.
- (A) Relax the two t1112 pins for the wrong-generation case to `errors.Is(ErrStalePeer)` + `StaleEndpoint` redirect (edits another card's tests).
- (B) Make `StaleEndpointError.Error()` return the bare `ErrStalePeer` text (t1112 tests pass untouched). Downside: MCP `factory_msg_send` surfaces errors only as `toolErr` text (`internal/cli/mcp_server.go:966`), so peers lose the current-endpoint hint AC-FLH-007 requires; the hook notice reads fields and is unaffected.

**Lead decision: (A).** Basis: t1112's own contract on `ErrStalePeer` (`store.go:46-50`: a richer stale error joins the class via `Is(target) bool`); (B) would erase AC-FLH-007's current-endpoint hint on the MCP path. Applied: only the wrong-generation case is loosened, in `TestCharacterizeVerifyPeerOutcomes/wrong_generation` and `TestStalePeerOutcomeMatchesErrStalePeer` — it asserts `errors.Is(err, ErrStalePeer)` and a `StaleEndpoint(err)` redirect carrying the current session and generation. Session / pid / process-start / missing-row keep their exact-text assertions; `staleText` is kept. Each changed assertion carries the one-line t1082 AC-FLH-007 comment. After (A): `go test -race -count=1 ./internal/factorymsg/...` exit 0 (`.moai/reports/t1082/align-race-factorymsg-A.txt`); ac05/ac07/ac18/ac19 re-run verbatim → `true`, exit 0.

### Test-only exports (lead item ④)
Kept exported. `internal/hook/factory_handoff_race_test.go` is package `hook` and must call the unexported hook production entry points `registerFactorySessionStartPeer` / `registerFactoryUserPromptPeer` (`internal/hook/factory_messages.go:45,49`) — the AC contract forbids test-only registration paths. It cannot move into `internal/factorymsg` (package `factorymsg` importing `hook` is an import cycle; package `factorymsg_test` cannot reach unexported hook functions), and `export_test.go` symbols are invisible to package `hook`'s test binary. Each of `WithStepHook`, `HandleStats`, `SharesHandle` and the `Step*` constant block now carries a one-line test-only-seam godoc; no non-test caller exists (grep: only the definitions).

### Regressions (uncommitted merge tree)
- `go test -race ./internal/factorymsg`: exit 1 before decision (A) — only the two t1112 text pins above; exit 0 after (A).
- `go test -race -run 'Handoff|Factory' ./internal/hook`: exit 0. `go test -race -run 'LaneHandoff' ./internal/cli`: exit 0.
- `go vet` (factorymsg, hook, cli) 0; `golangci-lint run` 0 (`0 issues.`); `go build ./...` 0; `GOOS=windows go build ./...` 0.
- AC commands verbatim: ac01–ac08, ac14, ac16, ac17, ac18, ac19 → all `true`, exit 0 (`.moai/reports/t1082/ac*.jsonl`).

### Merge commit and merge-tree
Merge commit: `f1073a5888f6511ec012a451df192c0345362d98` (parents `b34625050`, `52a486635`).

`git merge-tree --write-tree --name-only WT-dual-harness-recovery WT-factory-lane-worktree-handoff; echo "exit=$?"`:
```
7859a44bf8d778992e63d83e63a47f168947f04a
internal/factorymsg/store.go

Auto-merging internal/factorymsg/store.go
CONFLICT (content): Merge conflict in internal/factorymsg/store.go
exit=1
```
One conflicted hunk, in `Send`'s idempotency re-lookup (printed tree lines 710-722): t1100 replaces the base lookup with a `sender_slot`-scoped SELECT and compares against the envelope's recipient columns; t1082 keeps the `sender_session` SELECT plus the `originalRecipient` step. This is the planned seam — the likely resolution is t1100's base SELECT with t1082's `originalRecipient` step layered on the comparison — but it was measured only, not resolved (t1100's lane).

## § M4 lane record — Recovery, safety, and compatibility (2026-09-23/24, manager-develop, cycle_type=tdd)

Base `a44c9f30c`. Commits (all name card t1082):

| SHA | Subject |
|---|---|
| `e15be1b70` | fix(factorymsg): refuse tombstoned session UUID on launcher resume bind |
| `1ffab026b` | feat(factory): add `moai factory handoff abandon-lane` operator termination |
| `8fa0d5f5d` | feat(factory): add lane handoff crash points and restart reconciler |
| `c16d1422a` | test(factory): cover abandon store transaction and recovery edges |

Code HEAD for every measurement below is `c16d1422a` unless a row names an earlier commit. Logs are under `.moai/reports/t1082/` (gitignored, local only).

### What changed

- **D1(b), REQ-FLH-010** — `internal/factorymsg/store.go` `BindLaunchPending`: inside its write transaction, before the slot row is read, `sessionTombstoned(ctx, tx, p.SessionUUID)` (session UUID only, any generation); on hit returns `staleError(ctx, tx, NackStaleEndpoint, slot)` and writes nothing. `staleOrUnregistered`'s pair match is not used. The t1074 hunk is one guarded block plus two doc lines. `internal/hook/factory_messages.go`: the SessionStart branch maps a `BindLaunchPending` error through `factoryHandoffRegistrationNotice` before the degraded fallback.
- **REQ-FLH-011 abandon** — `internal/factorymsg/handoff_abandon.go` `Store.AbandonLane(ctx, slot, probe func(int) (string, homestate.ProcessIdentityState))`: one write transaction rereads the lane's open handoff and probes the recorded source owner. Live with the recorded start, Live with an empty start, and Indeterminate refuse with `SOURCE_OWNER_LIVE`; no open handoff refuses with `HANDOFF_NOT_PENDING`; otherwise `ABANDONED`/`OPERATOR_ABANDONED` plus one event row, nothing else. `internal/cli/factory_handoff_recover.go`: `moai factory handoff abandon-lane --slot <slot> [--run <id>]`, wired through the single package variable `abandonLaneProbe` (production value `homestate.ProbeProcessIdentity`).
- **REQ-FLH-011 crash recovery** — `internal/cli/factory_lane_handoff_recover.go`: crash points `reserved`, `created`, `renamed`, `switch-pending`, `relocated`, `bound` behind the test-only `laneHandoffFailpoint`; rebind writes are injected through the existing `WithStepHook`. `recoverLaneHandoff` decides from broker/filesystem/Git facts: `none` (no handoff, NACK, ABANDONED), `finalize` (BOUND: receipt + release + tombstone read back, nothing written; any missing is reported as corruption), `abandoned` (target unregistered `OWNER_UNKNOWN`, dirty `TARGET_DIRTY`, commits beyond the pin `TARGET_UNMERGED`, off-pin `BASE_DRIFT`, other branch `BRANCH_COLLISION`, vanished `TARGET_MISSING`), `resume` (RESERVED → create/rename/verify → WT_READY; WT_READY and SWITCH_PENDING_INTERACTIVE unchanged; SWITCH_PENDING_HEADLESS with recorded relocation → rebind), `nack` (headless with no recorded relocation → `RELOCATION_EVIDENCE_INVALID`; a rebind that NACKs). No path deletes a worktree. `createHandoffTarget`'s provenance tail moved into `verifyCreatedTarget` so recovery reuses it. New store helpers: `AbandonHandoff`, `HandoffReceipt`.

### RED → GREEN per AC

| AC leg | RED (command, observed, log) | GREEN (log) |
|---|---|---|
| AC-FLH-007 turn-registration ×3, BOUND-row | `go test ./internal/factorymsg -run '^TestFactoryLaneHandoffStaleEndpointRejected$' -count=1` at `a44c9f30c` + new test: exit 1, `handoff_bind_test.go:630: err=<nil>, want a stale-endpoint NACK STALE_ENDPOINT` (BOUND-row leg; the three turn-registration legs already passed — M3b behavior, first time in the named test) — `run-m4-red-ac07.log` | `run-m4-green-ac07.log`, exit 0 |
| AC-FLH-007 launcher resume + positive | same leg run alone through a temporary probe test (deleted after): exit 1, `handoff_bind_test.go:713: launcher resume: tombstoned session bound to the launch-pending row` — `run-m4-red-ac07-launcher.log` | subtest `launcher_resume_leg` PASS |
| AC-FLH-003 BOUND notice + launch-pending notice | `go test ./internal/hook -run '^TestFactoryLaneHandoffInteractiveStateMachine$' -count=1 -v`: exit 1; BOUND leg `SessionStart = ""`; launch-pending leg `SessionStart = "factory messaging bound: … generation=5 …"` vs UserPromptSubmit `"… current at  generation 4; …"` — `run-m4-red-ac03.log` | `run-m4-green-ac03.log`, exit 0 |
| AC-FLH-020 | compile RED `undefined: abandonLaneProbe`, `undefined: factorymsg.NackSourceOwnerLive` (`run-m4-red1-ac20.log`); behavioral RED with the seam variable only: exit 1, `result err=unknown flag: --slot, want NACK SOURCE_OWNER_LIVE` in all four states + `HANDOFF_NOT_PENDING` case (`run-m4-red2-ac20.log`) | `run-m4-green-ac20.log`, exit 0 |
| AC-FLH-009 / AC-FLH-010 | compile RED (`run-m4-red1-ac09-ac10.log`); with crash points and a stub reconciler: every crash point reached, every subtest `reconcile: not implemented` (13 + 5), exit 1 (`run-m4-red2-ac09-ac10.log`) | `run-m4-green-ac09-ac10.log`, exit 0 |

The AC-FLH-003 fixture follows the SPEC writer's note: the handoff-bound owner is a helper child that is killed AND reaped (`startStoppableLiveOwner`/`stopAndWait`; `startStoppableOwner` in factorymsg), then polled until the probe no longer reads it current.

### Mutant table (each applied to the GREEN tree and reverted; restoration verified by `shasum -c`)

| Mutant | Caught by | Result | Log |
|---|---|---|---|
| pair-match tombstone (`session_uuid=? AND generation=?` with the input generation) | AC-007 BOUND-row (`:630`), AC-003 both legs (`:269`, `:310`) | exit 1 / exit 1 | `run-m4-mutant-pair-match-tombstone.log` |
| no tombstone read in `BindLaunchPending` | AC-007 `:630`, AC-003 `:269`, `:310` | exit 1 / exit 1 | `run-m4-mutant-no-tombstone-read.log` |
| tombstone read only after a launch-pending row was found | AC-007 `:630`, AC-003 BOUND leg `:269` | exit 1 / exit 1 | `run-m4-mutant-tombstone-after-launch-pending.log` |
| hook not wired (SessionStart bind error → degraded) | AC-003 `:269`, `:310` | exit 1 | `run-m4-mutant-hook-not-wired.log` |
| Indeterminate collapsed into "not current" | AC-020 `:174` in all four states | exit 1 | `run-m4-mutant-indeterminate-collapsed.log` |
| `ownerCurrent` bool seam reused | AC-020 `:174` in all four states | exit 1 | `run-m4-mutant-ownercurrent-bool-seam.log` |
| ABANDONED deletes the worktree | AC-010 `:343` target files changed | exit 1 | `run-m4-mutant-abandon-deletes-worktree.log` |
| dirty check removed | AC-010 dirty `decision = resume/` | exit 1 | `run-m4-mutant-no-dirty-check.log` |
| unmerged commits resumed | AC-010 unmerged | exit 1 | `run-m4-mutant-unmerged-resumed.log` |
| headless without relocation stays pending | AC-009 `after_switch_pending_headless_before_rpc` | exit 1 | `run-m4-mutant-headless-no-relocation-stays-pending.log` |
| BOUND not finalized | AC-009 `after_commit_before_receipt_delivery` | exit 1 (first attempt was a build failure; re-run compiled) | `run-m4-mutant-bound-not-finalized.log` |
| card-id branch not renamed on resume | AC-009 `after_create_before_rename` → `BRANCH_COLLISION` | exit 1 | `run-m4-mutant-no-rename-resume.log` |

### Verification batch (lane env scrubbed in the same invocation)

| Check | Command | Result |
|---|---|---|
| AC gates, acceptance.md verbatim | AC-001..010, 014, 016..020 run through the scrub | 16/16 `true`, exit 0 each (`run-m4-ac-gates.txt`, `ac*.jsonl`) |
| race, factorymsg | `go test -race -count=1 -timeout=600s -coverprofile=… ./internal/factorymsg/...` | `ok … 61.916s coverage: 79.7%`, 0 DATA RACE (`run-m4-race-factorymsg.log`) |
| race, hook (all packages) | `go test -race -count=1 -timeout=900s ./internal/hook/...` at `8fa0d5f5d` | `FAIL internal/hook 543.303s`; 0 DATA RACE; subpackages ok. Failures: `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`, `TestFactoryBoundUserPromptSubmitDoesNotRewritePeer` (`factory endpoint is launch-pending`), `TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift` (timing budget) (`run-m4-race-hook.log`). See attribution below |
| cli subset, race | `go test -race -count=1 -timeout=20m ./internal/cli -run '^(TestFactoryLaneHandoff.*\|TestLaneHandoff.*\|TestSessionWorktree.*\|TestCodex.*\|TestMoaiMCPServer_RegistrationMatchesCatalog\|TestFactoryMsg.*\|TestFactory.*\|TestHandoff.*\|TestRecover.*)$' -v` with `MOAI_FACTORY_WORKER(S)`/`MOAI_KANBAN_BACKEND` also unset | `ok … 204.969s`; 1022 PASS, 0 FAIL, 12 SKIP (env-gated live probes), 0 DATA RACE; `TestMoaiMCPServer_RegistrationMatchesCatalog` PASS (`run-m4-race-cli.log`) |
| vet | `go vet ./internal/factorymsg/... ./internal/hook/... ./internal/cli/...` | exit 0 |
| lint | `golangci-lint run` same packages | `0 issues.`, exit 0 (`run-m4-lint.log`) |
| build | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `GOOS=windows GOARCH=amd64 go vet ./internal/hook ./internal/factorymsg ./internal/cli` | exit 0 ×3 |
| gofmt | `gofmt -l internal/factorymsg internal/hook internal/cli` | empty |

**Hook failures are pre-existing, not M4.** Same four tests ×5 (`-count=5`): M4 tree 15 PASS / 5 FAIL (`run-m4-hook-flake-x5.log`); base `a44c9f30c` exported by `git archive` 11 PASS / 9 FAIL with the same messages (`run-m4-hook-flake-x5-base-a44c9f30c.log`, load average 43 at the time). These are the §J M1 flake and a wall-clock budget test; both fail more often on the base than on M4 in this sample, which does not establish anything about M4 beyond "not introduced".

**First cli run was contaminated by this lane's own factory env.** `TestCodexSpawn_RealAssemblyThroughStubTmux` failed with `MOAI_FACTORY_WORKER=agent-41` in the expected tmux command; with those variables also unset it passes (`run-m4-cli-codexspawn-scrubbed.log`), and the full subset re-run above is clean.

### Coverage (file-level, statement-weighted)

| File | Profile | Covered |
|---|---|---|
| `internal/factorymsg/handoff_abandon.go` (M4) | factorymsg package | 43/50 = 86.0% |
| `internal/factorymsg/handoff_bind.go` | factorymsg | 182/210 = 86.7% |
| `internal/factorymsg/handoff.go` | factorymsg | 91/104 = 87.5% |
| `internal/factorymsg/store.go` (t1074 file) | factorymsg | 367/503 = 73.0% (M3b 71.8%); the new guard block is covered |
| `internal/hook/factory_messages.go` | hook (`8fa0d5f5d`) | 75/87 = 86.2% |
| `internal/hook/factory_handoff_bind.go` | hook | 35/41 = 85.4% |
| `internal/cli/factory_lane_handoff_recover.go` (M4) | cli, `-coverpkg` run of the four M4 tests | 77/91 = 84.6% |
| `internal/cli/factory_handoff_recover.go` | cli subset profile (`8fa0d5f5d`) | 51/59 = 86.4% |
| `internal/cli/factory_lane_handoff.go` | cli subset | 99/104 = 95.2% |
| `internal/cli/factory_lane_handoff_switch.go` | cli subset | 71/79 = 89.9% |
| `internal/cli/factory_lane_handoff_bind.go` | cli subset | 17/19 = 89.5% |

### Gaps

- AC-FLH-015 `TestFactoryLaneHandoffT1074Compatibility` does not exist; M4 re-ran the existing t1074 broker/roster/receipt/SessionStart tests and the MCP catalog invariant instead. The named test, with its 36/14/22 catalog count and zero-new-store inventory, is unwritten.
- AC-FLH-011 `TestFactoryLaneHandoffNoPreBoundWrites` does not exist and was not in the M4 dispatch.
- `factory_lane_handoff_recover.go` is at 84.6%, below 85%; the uncovered blocks are broker/Git I/O error returns.
- The cli coverage figures for the older files come from the `8fa0d5f5d` subset run; the last commit adds tests only.
- Full `internal/cli` suite not run (CI owns it). LIVE AC-FLH-012/013: M5, NOT_RUN.
- `TestAbandonLaneOwnerStates`, `TestAbandonLaneRefusals`, `TestAbandonHandoffAndReceiptReadback`, `TestLaneHandoffRecoveryEdges` were written after GREEN for coverage and have no individual RED.
- The AC-FLH-020 test swaps only `abandonLaneProbe`; the cli production value is never exercised against a real dead owner.

### Residual risks

- `recoverLaneHandoff` has no production caller yet, like the rest of the handoff controller; it is reached only from tests until a command or M5 wiring calls it.
- A recovered SWITCH_PENDING_HEADLESS rebind trusts the caller's `Owner` to be the relocated thread's owner; a wrong owner yields `BINDING_EVIDENCE_INVALID` NACK, not a wrong bind.
- The `BindLaunchPending` refusal keeps the launch-pending row in place until its owner dies (as REQ-FLH-010/017 specify); the lane's roster shows `launch_pending` with an empty session in that window.
- Crash points are a panic-based seam; they model process death for the broker (transactions roll back) but not partial Git operations inside `git worktree add`.

### M4 follow-up — AC-FLH-011/015 named tests, recovery coverage, merged-tree re-measure (2026-09-24)

Tree: the lane absorbed local develop as merge `09e620d2d` (brings `cb625f74b`, which contains t1109 `aae9bafb5`). Commit on top: `a0d4755a7` test(factory): add AC-FLH-011 and AC-FLH-015 named tests, raise recovery coverage. All measurements below are on `a0d4755a7`.

**RED (named tests absent).** `grep -rn "func <name>(" internal --include='*_test.go'` exit 1 for both names, with the positive control `TestFactoryLaneHandoffNoPrivateControl` found (exit 0) (`run-m4b-red-absent.txt`). The acceptance.md gate commands, run verbatim through the lane scrub at `09e620d2d`: AC-FLH-011 `false` exit 1, AC-FLH-015 `false` exit 1 (`run-m4b-red-gates.txt`). The behavior both tests assert already existed, so there is no behavioral RED. Discrimination is shown by the mutants below instead.

**AC-FLH-011 `TestFactoryLaneHandoffNoPreBoundWrites`** (`internal/cli/factory_lane_handoff_nowrite_test.go`). Instruments the handoff Git adapter (`handoffCommand`, argv recorded) and uses the crash-point seam as an observation hook. It measures every state: `reserved`, `created`, `renamed`, `WT_READY`, `switch-pending`, `relocated`, `SWITCH_PENDING_HEADLESS`. At each state it checks:
- the primary branch and HEAD, its status, and its file listing;
- target cleanliness and commits beyond the pin;
- `AuthorizeCardWrite` on the source;
- claim followed by receipt on the source.

After BOUND, one dispatch round runs on the bound endpoint. Observed: `AC_FLH_011_COUNTERS code_writes=0 wrong_cwd_writes=0 primary_switches=0 pre_bound_commits=0 task_acks=0 messages=2/2/0 git_calls=19`.

**AC-FLH-015 `TestFactoryLaneHandoffT1074Compatibility`** (`internal/cli/factory_lane_handoff_compat_test.go`):
- MCP catalog `36/14/22`, and the registered server matches it.
- A full headless handoff to BOUND plus a restart reconcile (`finalize`) run on the run.
- The t1074 launcher register+bind, send, claim, body, disposition, receipt and roster path then runs on `lane-2` in the same run.
- The store inventory under `MOAI_HOME` and the primary (SQLite files and sockets) is unchanged. The handoff tables live in the existing broker file at `BrokerPath`.
- Every subprocess the handoff started is a finished `git`.

Observed: `AC_FLH_015 catalog=36/14/22 stores=2 git_calls=19`.

**Mutants** (applied to the GREEN tree, reverted, restoration asserted by sha256 in the runner):

| Mutant | Test | Result | Log |
|---|---|---|---|
| `AuthorizeCardWrite` always grants | 011 | FAIL `code_writes=7` | `run-m4b-mutant-ac11-authorize-always.log` |
| controller runs `git -C <primary> checkout -b` after reserve | 011 | FAIL `primary_switches=7` | `run-m4b-mutant-ac11-primary-branch-switch.log` |
| rebind release moves no message (`… AND 0`) | 011 | FAIL `messages sent/received/lost = 2/0/2` | `run-m4b-mutant-ac11-release-drops-messages.log` (the first draft of this mutant was malformed SQL, and it failed for that reason. It was rewritten, re-run, and the log was overwritten) |
| controller writes `.moai/handoff-queue.db` | 015 | FAIL `store inventory changed` | `run-m4b-mutant-ac15-parallel-store-file.log` |
| controller starts `git status` without waiting | 015 | FAIL `was never waited: a lingering process` | `run-m4b-mutant-ac15-lingering-subprocess.log` |

**Recovery coverage.** `TestLaneHandoffRecoveryEdges` gains four subtests: inactive run, develop moved after the rename crash (→ `BASE_DRIFT`), card-branch rename collision on resume (→ `BRANCH_COLLISION`), and BOUND without its receipt (→ corruption error; the receipt row is deleted by test-only SQL). `internal/cli/factory_lane_handoff_recover.go`: 81/91 = **89.0%** (was 84.6%). Measured with `go test -coverpkg=./internal/factorymsg,./internal/cli -run '^(TestFactoryLaneHandoff.*|TestLaneHandoff.*)$' ./internal/cli` (`run-m4b-cli-xpkg-cover.log`).

**Hook flakiness on the merged tree.** Same four tests ×5 with `-race`, scrubbed env (`run-m4-hook-flake-merged.log`): 20 PASS / 0 FAIL, exit 0. Load average was 19.5 at start and 18.5 at end. The `session_start_drift_fill_test.go:579` lines are `t.Log`, not failures. No failure remained, so no base (`cb625f74b`) run was needed.

| Tree | Runs | PASS | FAIL | Attribution |
|---|---|---|---|---|
| `8fa0d5f5d` (before absorb) | 20 | 15 | 5 | same messages on base `a44c9f30c` (11/9) → pre-existing |
| `a0d4755a7` (after absorb) | 20 | 20 | 0 | none |

The absorbed develop includes t1109 "launch-pending bind budget split". Whether that commit is why the launch-pending flake is gone is **not established** (one sample, different load).

**Verification batch** (lane env scrubbed in the same invocation):

| Check | Result |
|---|---|
| AC-FLH-001..011, 014..020 via acceptance.md commands verbatim | 18/18 `true`, exit 0 each (`run-m4b-ac-gates.txt`, first line is the HEAD SHA) |
| `go test -race -count=1 ./internal/factorymsg/...` | `ok 44.042s`, 0 DATA RACE (`run-m4b-race-factorymsg.log`) |
| `go vet` factorymsg/hook/cli | exit 0 |
| `golangci-lint run` factorymsg/hook/cli | `0 issues.`, exit 0 (`run-m4b-lint.log`) |
| `GOOS=windows GOARCH=amd64 go build ./...`, `go vet` of the three packages | exit 0 / exit 0 |

**Gaps.**
- AC-FLH-015's Given lists SessionStart fixtures. The named test lives in `internal/cli` and does not drive the hook SessionStart path; the t1074 SessionStart tests in `internal/hook` are covered only by the flake re-run above (4 tests) and the M4 full hook run.
- The AC-FLH-011 filesystem instrument is observational: snapshots taken at each state and compared. Writes are not intercepted at the syscall level, and a write undone between two observations is invisible.
- Neither new named test has a behavioral RED; discrimination rests on the five mutants.
- The full `internal/cli` and `internal/hook` suites were not run on the merged tree (CI owns them).

## § M5 lane record — LIVE mixed-factory verification (2026-09-24, manager-develop, cycle_type=tdd)

Base HEAD `7cbaf966c`. Result: **LIVE FAIL (NOT_RUN on both rows)**. Only the non-LIVE gate-quality test was delivered. No model was called.

### Delivered

- `internal/cli/factory_lane_handoff_live_gate_test.go` — `TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants`. It reads the `jq -se` and `jq -e` programs from the AC-FLH-012 and AC-FLH-013 bash blocks in `acceptance.md` and runs them with the `jq` binary, so the production predicates are tested, not a copy. Per AC: a valid evidence control and a valid event-stream control must pass. Then every leaf field is removed one at a time, `bypass.fixture|mock|direct_register_peer=true`, a child `fail`, a child `skip`, and a `NOT_RUN` output line must all be rejected. AC-FLH-013 also covers `wrong_method_thread_start` (`thread/start`, `forked_from_id=null`). A missing `jq` is `t.Fatal`, not a skip.

### RED → GREEN

| Step | Command | Observed | Log |
|---|---|---|---|
| RED (test absent) | `go test -json ./internal/cli -run '^TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants$' -count=1` | `ok … [no tests to run]`; named PASS count `0` | `run-m5-red-gate.jsonl` |
| GREEN | same, `-v` | PASS; `AC-FLH-012: 2 valid controls passed, 51 mutants rejected`, `AC-FLH-013: … 55 mutants rejected`; jq named-PASS/no-fail/no-skip `true` | `run-m5-gate.jsonl` |
| Mutant: predicate loosened | `acceptance.md` copy in scratchpad with every ` and (.bypass.fixture==false)` removed; the const was pointed at it, then reverted (`grep -c scratchpad` → 0) | exit 1; `mutant missing:bypass.fixture accepted` and `mutant bypass.fixture=true accepted` for both ACs | `run-m5-gate-mutant-loose-fixture.log` |

`go vet ./internal/cli` exit 0; `golangci-lint run --new-from-rev=HEAD ./internal/cli/` → `0 issues.`

### LIVE gate commands (acceptance.md verbatim, lane env and provider credentials scrubbed in the same invocation)

| AC | `go test` exit | `jq -se` gate | Live test PASS count | Evidence JSON | Verdict |
|---|---|---|---|---|---|
| AC-FLH-012 | 0 | `false` | 0 (`TestFactoryLiveCodexCodexWorktreeHandoff` does not exist) | absent | FAIL (NOT_RUN) |
| AC-FLH-013 | 0 | `false` | 0 (`TestFactoryLiveClaudeCodexWorktreeHandoff` does not exist) | absent | FAIL (NOT_RUN) |

Logs: `ac12.jsonl`, `ac13.jsonl`. The go test exit is 0 because the selector matched only the gate test. The `jq -se` predicate is what gives FAIL here.

### Model calls and wall time (lead condition 5)

| Combination | Mode | Model calls | LIVE wall time | Status |
|---|---|---|---|---|
| Codex lead ↔ Codex lane | interactive `/cd` | 0 | 0 s | NOT_RUN |
| Claude lead ↔ Codex lane | headless `thread/fork(cwd)` | 0 | 0 s | NOT_RUN |

Machine-readable copy: `.moai/reports/t1082/m5-live-ledger.json`. `ac12-evidence.json` and `ac13-evidence.json` were **not** written: the SPEC gives that job to the LIVE tests alone, and a hand-written evidence file is exactly the self-report REQ-FLH-014 rejects.

### Authentication check (lead condition 2; isolated homes per condition 3)

| Harness | Command | Output | Exit |
|---|---|---|---|
| codex 0.156.1 | `CODEX_HOME=<scratchpad>/codexhome codex login status` | `Not logged in` | 1 |
| claude 2.1.280 | `CLAUDE_CONFIG_DIR=<scratchpad>/home/.claude claude auth status` | `"loggedIn": false, "authMethod": "none"` | 1 |

Login was not attempted (condition 2). The worktree guard refused a shell-level `HOME=` override, so the claude check isolated only `CLAUDE_CONFIG_DIR`.

### Blockers (why nothing LIVE ran)

1. **No model-call cap in the SPEC.** spec/plan/design/acceptance set a time cap (`go test -timeout=240s` per LIVE command) but no call count. Lead condition 1 says to stop in that case.
2. **Isolation vs. authentication conflict.** Lead condition 3 requires isolated HOME/CODEX_HOME, and in those homes both harnesses are unauthenticated (above). plan.md §Verification 3 requires the "actual installed auth context" of the t1074 harness. Both cannot hold without copying credentials into the isolated home, and that was not done.
3. **No production trigger for a handoff.** `prepareLaneHandoff`, `switchLaneHandoffInteractive`, and `switchLaneHandoffHeadless` have no non-test caller: `grep` finds none outside `_test.go`, and the only production call of `bindLaneHandoffHeadless` is in recovery. No CLI command or MCP tool exists (the MCP catalog is pinned at 36/14/22). AC-FLH-013 says "the lead triggers actual `thread/fork(cwd)`" with `contexts.lead.production_cli==true`. A lead model context has no production way to do that. Calling the Go functions from test code would be the test acting, not the lead. That is a SPEC gap, not a lane decision.
4. **AC-FLH-012 needs a human.** The interactive row needs a real `/cd` and a real next normal turn in a Codex TUI. Driving the TUI is private control (tmux send-keys, AC-FLH-014), so it cannot be scripted. The operator must type `/cd <target>` in a live Codex lane, then send one normal message.
5. **The t1074 harness is not reusable as-is.** `newFactoryLiveFixture` registers peers with `RegisterPeer` directly (fails `bypass.direct_register_peer==false`), and its "claude" backend runs `moai glm`, not Claude.

### Gaps

- The LIVE rows did not run. The real `/cd` SessionStart ordering and the real `thread/fork(cwd)` against stored history remain unobserved.
- The gate test's valid controls are constructed documents. They show that the predicates accept a well-formed document. They say nothing about whether a real run would produce one.
- The event-stream mutants use a synthetic `go test -json` stream, not a captured one.

### Residual risk

- The gate test reads predicates by regex from the first bash block under each AC heading. If `acceptance.md` is restructured so the block moves, the test fails with "no bash block" or "lacks a jq predicate" instead of passing silently. A block that keeps the heading but carries different jq programs is read as the new truth by design.
