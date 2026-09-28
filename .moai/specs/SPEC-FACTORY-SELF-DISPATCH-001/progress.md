# progress.md — SPEC-FACTORY-SELF-DISPATCH-001 (card t1240)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-27
- artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, this skeleton (Tier L)
- baseline: worktree `.claude/worktrees/t1240`, branch `WT-factory-self-dispatch`, base develop `ed506740b`
- vocabulary source: SPEC-ROLE-NAMING-CODE-001 committed text on `WT-role-naming-code` at `d39a1dc09`
- run gate: REQ-SD-001 — run starts only after t1256 lands on develop (plan.md §D.3)
- plan_audit: iter-1 FAIL 0.66 on `0cb1d127b` (`.moai/reports/t1240/plan-audit-iter1.md`) → v0.2.0 revision (D1-D12, O1-O6, O10, O11 resolved; decisions in plan.md §B B1-B10); iter-2 FAIL 0.79 (`.moai/reports/t1240/plan-audit-iter2.md`) → v0.3.0 narrow fix (N1-N4, P1-P6); iter-3 FAIL 0.83 (`.moai/reports/t1240/plan-audit-iter3.md`) → leader-approved override fix v0.4.0 (R1-R3); iter-4 → leader final decision: OD-1/OD-2 PERMITTED (self-dispatch lane mode only), plan closed PASS-WITH-DEBT, v0.5.0; v0.5.1 adds the local lane-protocol rule §6 as fourth doctrine deliverable

## §E.2 Run-phase Evidence

### Pre-flight (plan.md §C, 2026-09-28)

- **C.1 develop_sha**: `a7190891d` (run start read; the leader's dispatch named `f6b1d7e4f`, develop advanced before absorption — REQ-SD-001 binds the SHA actually read).
- **C.2 t1256_landed**: yes — `git show develop:internal/cli/factory.go` → `factoryLaneRoleToken = "lane"` (factory.go:61, legacy `worker`/`agent` tokens rejected in comment); `git show develop:internal/config/envkeys.go` → `EnvFactoryRole = "MOAI_FACTORY_ROLE"` (:337), `FactoryRoleLane = "lane"` (:347).
- **C.3 absorb**: `git merge develop` → `abb815921`, conflicts 0. Citation re-read on absorbed tree: guard anchor `internal/hook/contract_sign_guard.go:133` reads `EnvFactoryRole == FactoryRoleLane` ✓; F1 transition API in `internal/homestate` (card_record.go:72,126 lease/transition family) ✓; `codex_launcher.go` factory-entry refusal seam present (region ~:675) ✓; `session_start_factory.go:74` startup-only gate ✓; `configtoml.go:98,173` env_vars ✓; `exportFactoryLaunchFacts` cc.go:197,221 → kanban.go:539-542 ✓; launcher does not stamp `MOAI_FACTORY_ROLE` (grep 0 in factory.go) — F2 premise holds. Line-number drift: envkeys :323,332→:337,347; kanban.go :514→:539; codex_launcher :701→~:675. **Vocabulary-conflict check (leader directive): NO CONFLICT — SPEC vocabulary matches landed t1256 code; no SPEC revision round required.** Citation corrections land in research.md/design.md during M1 (outside the spec/plan/acceptance ownership boundary).
- **C.4 MCP-server tree follow (B7)**: measured live in this lane session (t1240, worker-62) — the running moai mcp-server resolves its project root from spawn-frozen state; after EnterWorktree the session is "reading another tree: pass project_root = git rev-parse --show-toplevel" (repeated runtime notice, this session). Server does NOT follow the session tree → REQ-SD-024's caller-supplied project_root requirement stands as written.
- **C.5 codex --help working-directory flag**: `codex --help` → `-C, --cd <DIR>` ("Tell the agent to use the specified directory as its working root") — the `-C` spelling stands (R6 resolved).
- **C.6 characterization** (`go test ./internal/homestate/... ./internal/kanban/... -count=1` + `go test ./internal/cli -run '^TestFactoryRoleTokenPinsGuardConstant$' -count=1`, absorbed tree `abb815921`): `ok internal/homestate 39.574s` · `ok internal/kanban 179.483s` · `ok internal/cli 0.785s` — baseline green.

Pre-flight complete (C.1–C.6). Run proceeds to M1.

### M1 — lane predicate, selection, permission boundary (2026-09-28)

Scope: REQ-SD-008..010, -015, -016, -025 (selection half). ACs: AC-SD-008, -009, -010 (CLI half; the
`project_root`/MCP half is M3 and consumes the same `factoryAssertParentCheckout`), -015 (CLI halves;
`todo_add` MCP-tool arms land with the tools in M3), -016 (CLI half; `factory_decide` MCP arm M3), -023.

- **RED (E8)**: tests written first in `internal/cli/factory_self_dispatch_test.go`; pre-implementation
  run `go test ./internal/cli -run '^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1` → build
  failure naming the missing production symbols (`undefined: factoryNextLeaseOnce`,
  `factoryNextWaitSleep`, `factoryAssertParentCheckout`), `FAIL ... [build failed]`.
- **GREEN (E1, env-scrubbed single compound invocations, tree = M1 commit)**:
  `go test ./internal/cli -run '^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1 -v` → `--- PASS` (6 subtests);
  `…AC009…` → `--- PASS` (2 subtests); `…AC010…` → `--- PASS`; `…AC015_LaneQueueAllowlistWalk…` → `--- PASS` (19-verb walk, walked set == tree);
  `…AC015_LabelOnlyIsNotALane…` → `--- PASS`; `…AC016…` → `--- PASS`; `…AC023…` → `--- PASS`.
- **Regression**: `go test ./internal/cli -run '^TestTodo' -count=1` → `ok 308.987s`;
  `… -run '^TestFR_' -count=1` → `ok 35.258s`; the 17 non-`TestTodo`-named runTodo callers → `ok 33.534s`;
  `go test ./internal/homestate/... ./internal/config/... ./internal/spec/... -count=1` → all `ok`;
  `go test ./internal/kanban/... -count=1` → `ok 180.547s`. Slot `internal-test-suite` acquired/released around the multi-package runs.
- **E2 builds**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E5 lint**: `golangci-lint run --timeout=5m` → `0 issues.` (baseline was `0 issues.` — no new findings).
- **E4 boundary**: `grep -n 'AskUserQuestion' internal/cli/factory_card.go internal/cli/todo.go internal/cli/todo_pr.go internal/cli/factory_handoff_recover.go` → no matches (exit 1).
- **E3 coverage** (`go test ./internal/cli -run '^TestSD_' -count=1 -coverprofile`): new M1 functions 80-100%
  (factoryLaneAdmission/Refusal 100, todoRefuseLaneMutation 100, writeTodoPRRows 100, factoryNextSkipForBackend 100,
  factoryNextClaim 90, newFactoryNextCommand 87.9, computeTodoPRRows 88, factoryNextLeaseOnce 80);
  package-wide 7.0% under the AC selector only (whole-package runs prohibited by acceptance.md §B — CI owns the full figure).
- **B10 wait values recorded**: interval 5s (`factoryNextWaitInterval`), default bound 15m
  (`factoryNextWaitBoundDefault`), one flag `--wait-bound <duration>` changes the bound; no-card exit status 3
  (`factoryNextNoCardExit` via `exitCodeError`).
- **Files**: `internal/cli/factory_card.go` (predicates, `next`/`stage`/`complete` verbs, decide guard,
  parent-checkout check, selection+lease), `internal/cli/todo.go` (lane queue guard on the todo
  PersistentPreRunE, allowlist {list, history, why, pr, triage} + bare parent), `internal/cli/todo_pr.go`
  (`computeTodoPRRows`/`writeTodoPRRows` extracted — `next`'s PR/landed line is byte-equal to `todo pr`'s by
  construction), `internal/cli/factory_handoff_recover.go` (verb registration),
  `internal/cli/factory_self_dispatch_test.go` (new AC tests).
- **Citation corrections (research.md/design.md, old → new)**: envkeys :323→:337, `FactoryRoleWorker`"worker" :332→`FactoryRoleLane`"lane" :347;
  guard read :131-133→:132-134; factory_card.go :81,213,302→:444,575,665; codex_launcher refusal :701→:690,
  const :740-745→:666-672, scanner :754-768→:678-681, child env :607-625→:521-535; session_start_factory gate
  :62-67→:73-78, notice builder `factoryWorkerNotice`:201→`factoryLaneNotice`:212; kanban.go :514→:543, :492-497→:520-526.
- **E6**: no push (lane does not push; the leader batch-pushes develop).

## §F Phase 4 Mode Selection

- tier: L · scope: >10 production files across cli/hook/config/kanban/homestate/codexwiring + template rules · domains: 6 (Go CLI, hooks, MCP server, launchers, doctrine rules, env constants) · language mix: Go + Markdown · concurrency benefit: LOW (coding-heavy, sequential milestone chain with shared files)
- direct: not selected — multi-milestone production change. fanout: not selected — coding-heavy (Anthropic coding-task parallelism caveat), milestones share files (factory.go touched by M1/M4/M5). sweep: not selected — semantic multi-rule work, not mechanical uniform transform. agent-team: not selected — no operator request.
- **Decision: serial** (sequential manager-develop delegations per milestone, M1→M7).
- Justification: Tier L coding-heavy chain; plan.md §D.2 file table shows M1/M4/M5 share `factory.go` and M1/M6 share hook files, so parallel writers would race one working tree. Kickoff approved via leader dispatch (operator 09-28 전력 완수 지시 + CLAUDE.local.md §31); skip-eligible plan-audit (PASS-WITH-DEBT 0.90 final iter4, artifacts unchanged since v0.5.1).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
