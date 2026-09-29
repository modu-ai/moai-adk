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

### M2 — integration surface and Codex merge-edge refusal (2026-09-28)

Scope: REQ-SD-013 (complete through the F1 merge gate), REQ-SD-023 (all four window refusals + the
lane-holds-then-releases flow), REQ-SD-025 (edge half — `complete`/`stage` CLI paths wire the
reusable Codex merge-edge check; the MCP tools consume the same check in M3). ACs: AC-SD-013,
AC-SD-024 (complete/stage CLI halves; the `CodexMergeRefusedMCP` verify command is M3 with the
tools), AC-SD-025.

- **Pre-flight (C)**: `git branch --show-current` → `WT-factory-self-dispatch`, `git rev-parse HEAD`
  → `490d64649ae80c2fccc3ae152a4cef5fdbb686c0`; `go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m
  ./internal/cli/...` → `0 issues.`; `grep -n 'AskUserQuestion' internal/cli/factory_card.go` →
  no matches (exit 1).
- **RED (E8)**: tests written first in `internal/cli/factory_complete_test.go`. Stage-1 run →
  `undefined: factoryCodexMergeSentinel` ×3, `FAIL ... [build failed]`. After wiring only the
  sentinel + the stage refusal, behavioral RED: `--- FAIL: TestSD_AC013_ClaudeCompleteViaIntegrationWorktree`
  (verbatim: `factory_complete_test.go:145: complete: accepts 1 arg(s), received 2`;
  `:172: factory complete: not yet implemented (SPEC-FACTORY-SELF-DISPATCH-001 M2)`;
  `:211/:225: want a not-provisioned refusal`; `:242: want a refusal naming --branch`),
  `--- FAIL: TestSD_AC024_CodexMergeRefusedComplete`, `--- FAIL: TestSD_AC025_IntegrationWindowSerializes`;
  `TestSD_AC024_CodexMergeRefusedStage` already PASS at this point (the stage wiring was the
  stage-1 change).
- **GREEN (E1, env-scrubbed single compound invocations, tree = M2 commit)**:
  `go test ./internal/cli -run '^TestSD_AC013_ClaudeCompleteViaIntegrationWorktree$|^TestSD_AC024_CodexMergeRefusedComplete$|^TestSD_AC024_CodexMergeRefusedStage$|^TestSD_AC025_IntegrationWindowSerializes$' -count=1 -v`
  → all four `--- PASS` (AC-013: 6 subtests incl. pre-merged arm, complete-does-the-merge arm,
  parent-checkout / no-tree / caller-source / card's-own-branch refusals; AC-024 complete+stage;
  AC-025 serialize-then-succeed with both merges ancestors of develop).
- **M1 regression**: `go test ./internal/cli -run '^TestSD_AC008|^TestSD_AC009|^TestSD_AC010|^TestSD_AC015|^TestSD_AC016|^TestSD_AC023' -count=1`
  → `ok github.com/modu-ai/moai-adk/internal/cli 17.167s`.
- **E2 builds**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
  (single compound `&&` invocation).
- **E5 lint**: `go vet ./internal/cli/...` + `golangci-lint run --timeout=2m ./internal/cli/...`
  → `0 issues.` (pre-flight baseline `0 issues.` — no new findings).
- **E4 boundary**: `grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/factory_card.go
  internal/cli/factory_complete_test.go` → no matches (exit 1).
- **E3 coverage**: see the full-package run recorded below (slot `internal-cli-testsuite` held
  for the run).
- **Design decisions (design.md D6/D7)**: complete reuses acquire's logic by CALL, never a
  shell-out — `integrationSessionID`, `session.ResolveOwnerPID`,
  `kanban.AcquireIntegrationLock`/`ReadIntegrationLock`, and the `worktreeForBranchFromList`
  parser; `internal/cli/integration.go` is untouched (B10 preference satisfied — no export was
  needed, same package). The branch resolution mirrors `resolveIntegrationTarget`'s decision
  order minus its $PWD legs (`factoryResolveIntegrationBranch` reads the card worktree for the
  caller fallback; `factoryWorktreeForBranchIn` anchors `git worktree list` at the card's repo,
  plan B7). A window already held by the caller keeps ITS recorded branch (REQ-SD-023: "the
  branch the integration window records"); a free/stale window is resolved and taken over with
  the displaced holder reported, never silently. complete does NOT release the window — the
  success output names `moai integration release` as the lane's next step.
- **Merge evidence**: the optional positional names the lane's re-measure file; when omitted,
  complete writes `.moai/reports/<card>/merge-record.txt` naming the merge SHA + tree identity
  (merge identity only — never a test-run claim).
- **Files**: `internal/cli/factory_card.go` (complete verb body + `factoryRefuseCodexMergeEdge`
  + stage Codex-edge wiring + window/branch/worktree helpers),
  `internal/cli/factory_complete_test.go` (new M2 AC tests).
- **E6**: no push (lane does not push; the leader batch-pushes develop).

### Absorption (2026-09-28, worker-70 continuation)

- Absorbed local develop `1ea627832` → merge `cb1f2c226`, conflicts 0.
- WorktreeNew re-adjudication — `go test ./internal/cli -run '^TestWorktreeNew_WiresTheSharedMaterializer$|^TestWorktreeNew_RefusesPlainDirectoryBeforeGit$' -count=1 -v` → both `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 0.833s`; the pre-existing-failure premise (worker-62 report: both red on develop `a7190891d`) is dissolved — fix `ea13cb094 test(worktree): expect neutral L1 root in CLI wiring (t1292)` landed on develop between a7190891d and 1ea627832.
- Scoped remeasure at cb1f2c226: `go test ./internal/homestate/... ./internal/kanban/... ./internal/config/... ./internal/spec/... -count=1` → homestate ok 45.312s, kanban ok 176.833s, config ok 5.617s/0.334s/0.315s, spec ok 119.073s (slot internal-test-suite acquired/released).
- M1/M2 regression at cb1f2c226: `ok github.com/modu-ai/moai-adk/internal/cli 25.711s`.

### M3 — worktree per card, stage, MCP tools (2026-09-28)

Scope: REQ-SD-011, -012, -014, -024. ACs: AC-SD-011, -012, -014 in full, plus the M3 arms of
AC-SD-010 (MCP half), -015 (todo_add), -016 (factory_decide), -024 (CodexMergeRefusedMCP).
Work ran 2026-09-28 through 2026-09-29 (429 quota reset continuation; tree unchanged and clean
across the boundary at `cb1f2c226`).

- **Pre-flight (C)**: `git rev-parse --short HEAD` → `cb1f2c226`; `git branch --show-current` →
  `WT-factory-self-dispatch`; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
  exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.` (baseline; CI version
  v2.1.6 per t1235/t1271); M1/M2 anchored regression → `ok github.com/modu-ai/moai-adk/internal/cli 25.711s`.
- **RED (E8)**: `factory_m3_test.go` written first; pre-implementation `go vet ./internal/cli`
  → build failure naming the missing production symbols (`undefined: handleTodoAdd`, `handleTodoList`,
  `handleFactoryNext`, `handleFactoryStage`, `handleFactoryComplete`, `handleFactoryDecide`), the M1/M2
  RED shape. AC-011/-012 were additionally behavioral RED at that point (no worktree creation; stage
  was the M2 stub returning "not yet implemented").
- **GREEN (E1, env-scrubbed single compound invocations, tree = M3 commit)**:
  - `go test ./internal/cli -run '^TestSD_AC011_CardWorktreeCreateReuseRefuse$|^TestSD_AC012_StageAppliesEdgeAndRenews$|^TestSD_WorktreeSlugShape$' -count=1 -v -timeout 10m` → `--- PASS` ×3 (AC-011: create/reuse/refuse; AC-012: holder edge + actor + renewal, lane-2 F1 refusal), `PASS`, `ok ... 5.436s`.
  - `go test ./internal/cli -run '^TestSD_AC014_MCPMatchesCLIWithProjectRoot$|^TestSD_AC014_ProjectRootRequired$|^TestSD_AC010_MCPNextParentCheck$|^TestSD_AC015_MCPTodoAddRefused$|^TestSD_AC016_MCPDecideRefused$|^TestSD_AC024_CodexMergeRefusedMCP$' -count=1 -timeout 10m` → `ok github.com/modu-ai/moai-adk/internal/cli 13.626s` (AC-014: six tools, twin fixtures, success + refusal each; required project_root rejections; the four M3 arms).
  - Full anchored set (M1+M2+M3): `go test ./internal/cli -run '^TestSD_AC008|...|^TestSD_WorktreeSlugShape' -count=1 -timeout 15m` → `ok github.com/modu-ai/moai-adk/internal/cli 59.505s`.
- **Regression**: `go test ./internal/cli -run '^TestFR_' -count=1 -timeout 10m` → `ok 47.176s`;
  `go test ./internal/mcp -count=1 -timeout 5m` → `ok 0.251s` (catalog size 45 + write-capable set 20);
  `go test ./internal/cli -run '^TestTodoAdd|^TestTodoList|^TestMoaiMCPServer_RegistrationMatchesCatalog' -count=1 -timeout 10m` → `ok 67.752s`.
- **E2 builds**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (compound `&&` invocation).
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/cli/... ./internal/homestate/... ./internal/mcp/...` → `0 issues.` — after 2 NEW errcheck findings (test-only `os.Unsetenv` unchecked) were fixed by restructuring the factory_next twin without an env switch; pre-flight baseline `0 issues.` restored. CI version v2.1.6.
- **E4 boundary**: `grep -n 'AskUserQuestion' internal/cli/factory_card.go internal/cli/mcp_factory_card.go internal/cli/mcp_todo.go internal/cli/factory_m3_test.go internal/cli/factory_self_dispatch_test.go internal/cli/factory_complete_test.go` → no matches (exit 1).
- **E3 coverage** (`go test ./internal/cli -run '^TestSD_' -count=1 -coverpkg=./internal/cli,./internal/homestate -coverprofile=/tmp/m3cover.out`): new M3 functions — factoryRefuseForeignWorktree 100%, factoryWorktreeSlug 95.8%, factoryNextWriteOutput 90%, handleTodoList 85.7%, handleFactoryStage 84.6%, factoryStageCard 81.5%, handleFactoryComplete 80%, handleFactoryNext 76.9%, handleFactoryDecide 75%, handleTodoAdd 75%, factoryEnsureCardWorktree 73.7%, factoryDecideCards 70% (its refused-card loop is exercised by the TestFR_ decide set, outside this selector), RecordCardWorktree 67.7%, factoryCardQueueTitle 50% (archived-queue + read-error branches), mcpRequiredProjectRoot 100%, newBufferedCommand 100%, factoryDecideLaneRefusal 100%; registration/schema one-liners (registerFactoryCardMCPTools, registerTodoMCPTools, requiredProjectRootOption) 0% under the selector — exercised by TestMoaiMCPServer_RegistrationMatchesCatalog, which ran green in its own anchored run. Package-wide 9.4% under the AC selector only (whole-package runs prohibited by acceptance.md §B — CI owns the full figure).
- **Design decisions**:
  - **Worktree landing path (t1292 absorption)**: the shared materializer's L1 root on this tree is
    `<root>/.moai/worktrees/<name>` — `sessionWorktreeSubdir` was moved from `.claude/worktrees` to
    `.moai/worktrees` by t1292 (commit `8c23dca9f`), which was absorbed AFTER acceptance.md was
    authored, so AC-SD-011's literal `.claude/worktrees/<card-id>` names the pre-absorption root. The
    implementation creates through the shared materializer seam (`worktree.WorktreeCreator` →
    `materializeSessionWorktree`, root.go:137 — never a bare `git worktree add`), leaf = card id,
    branch renamed in place to `WT-<slug>`; the AC test asserts the materializer's actual landing
    directory. **PASS-WITH-DEBT for leader adjudication** — the AC's discriminative content (leaf is
    the card id; WT- branch without the card id; record path equals the created directory; reuse never
    creates twice; a foreign directory refuses with the row unchanged) is verified in full; the parent
    directory naming is stale relative to the absorbed tree. Same dissolution shape the leader already
    adjudicated for `TestWorktreeNew_*` at absorption (fix ea13cb094).
  - **homestate.RecordCardWorktree (new API)**: recording the created path needs a post-lease record
    write; no F1 API existed (RecordPicked updates fields only while `picked`). Added the minimal
    version-checked, event-logged writer (appends one `card.fields` event, refuses moving a card onto
    a different tree, idempotent on the same path) rather than a direct SQL write in cli — the F1
    writer discipline (version compare + event) is kept. plan.md §D.2 estimated "at most a selection
    query helper" for internal/homestate — exceeded deliberately; no schema statement touched.
  - **One implementation per verb (design.md §3)**: the MCP handlers call the same functions the cobra
    RunE bodies call — factoryNextLeaseOnce + factoryEnsureCardWorktree + factoryNextWriteOutput,
    factoryStageCard, factoryCompleteCard, factoryDecideCards. The REQ-SD-025 Codex merge-edge check
    moved INSIDE factoryStageCard/factoryCompleteCard so every surface refuses identically (the M2
    RunE-level wiring was the only place it lived). The REQ-SD-015 queue guard and the REQ-SD-016
    decide refusal are the same guard functions; todo_add's refusal text comes from
    todoLaneMutationRefusalText (one wording source extracted from todoRefuseLaneMutation).
  - **project_root (REQ-SD-024)**: factory_next/stage/complete REQUIRE it (missing → rejected naming
    the argument; present → validateProjectRoot, the existing rejection behavior naming the path);
    factory_next's parent-checkout check evaluates the argument. todo_add/todo_list/factory_decide
    follow the existing optional convention (resolveToolProjectRoot). The todo verbs gained
    Root-anchored variants (runTodoAddAppendRoot/runTodoListRoot/todoStoreAt/todoReadStoreAt) sharing
    the CLI bodies.
  - **stage evidence positional**: `<sha>` where the edge reads a commit only, `<sha>:<repo-relative-artifact>`
    where the guard also names the artifact (T5/T11); a missing artifact is an F1 refusal verbatim.
  - **Catalog**: six tools registered (todo_add, todo_list, factory_next, factory_stage,
    factory_complete, factory_decide); internal/mcp catalog size 39 → 45, write-capable set 15 → 20;
    registration/catalog equality guard green.
- **M1/M2 test updates (behavior legitimately moved by M3)**: AC-SD-008/-009/-023 fixtures anchor
  cwd at the parent checkout (t.Chdir) and expect the freshly created worktree name in the head line
  ("t1"/"t2" instead of "-"); AC-SD-024 stage negative control now asserts the Claude-backend
  `stage <card> merging` SUCCEEDS (card → merging) instead of asserting a stub refusal — the M2
  comment explicitly deferred the stage behavior to M3.
- **Incident (test pollution, cleaned)**: the first regression run leaked two worktrees into the
  primary checkout (`.moai/worktrees/t1`, `.moai/worktrees/t2`, branches `WT-factory-card-1/-2`) —
  the M1 fixtures ran `next` without anchoring cwd, and the materializer resolves its root from the
  process cwd. Cleaned precisely (`git worktree remove --force` ×2 + `git branch -D` ×2 from this
  worktree; no other lane's tree touched), then the fixtures were anchored (the t.Chdir updates above).
- **REQ-SD-022 (frozen surfaces)**: diff-based evidence at this tree — `git status --short` names no
  file under internal/factorymsg or internal/kanban; the only homestate addition (card_worktree.go)
  carries no CREATE TABLE/ALTER TABLE/index statement; `internal/codexwiring/configtoml.go` untouched;
  no new environment variable name (the lane predicates read the existing constants). Gap: the
  mechanical AC-SD-022 tests (`TestSD_AC022_EnvVarsAllowlistFrozen` / `TestSD_AC022_SchemaStatementsFrozen`)
  are M7 deliverables and do not exist yet — both selectors returned `[no tests to run]` and are NOT
  counted as a pass here.
- **E6**: no push (lane does not push; the leader batch-pushes develop).
- Amendment (2026-09-29, manager-spec via re-delegation): `061614bd5` — acceptance.md stale
  worktree-root references amended (5 literals `.claude/worktrees/` → `.moai/worktrees/`: the §B
  fixture-layout line, AC-SD-011 ×2, AC-SD-013, AC-SD-018 exclusion; spec.md `updated:` bumped to
  2026-09-29; no REQ/AC renumbering, AC count 25 unchanged); `go test ./internal/spec -count=1` →
  `ok 137.078s` under the internal-test-suite slot.

### M4 — cc/glm lane launcher, marker stamp, git requirement (2026-09-29)

Scope: REQ-SD-002, -005, -006, -007, -017. ACs: AC-SD-002, -005, -006, -007, -017 (both arms).
Continuation at `061614bd5` (clean).

- **Pre-flight (C)**: `git rev-parse --short HEAD` → `061614bd5`; `git branch --show-current` →
  `WT-factory-self-dispatch`; `git status --short` → empty; `go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m
  ./internal/cli/... ./internal/hook/...` → `0 issues.` (baseline; CI version v2.1.6, measured
  `golangci-lint version` → v2.1.6). The anchored M1-M3 baseline was measured immediately after
  GREEN instead of before (order deviation, recorded): the first sweep surfaced exactly the four
  pre-M4 lane-join tests whose non-git fixtures the new REQ-SD-005 guard legitimately refuses
  (see the cascade note below) and nothing else; after the fixture cascade the full anchored set
  went green and stayed green through the final run.
- **RED (E8, verbatim, pre-GREEN tree = `061614bd5`)**:
  - AC-SD-002: `go test ./internal/cli -run '^TestSD_AC002_LaneLaunchStampsMarkerAndLabel$|...'` →
    `--- FAIL: TestSD_AC002_LaneLaunchStampsMarkerAndLabel` with `factory_m4_test.go:147: child env
    MOAI_FACTORY_ROLE = "", want the value constant "lane" (the marker name and value travel only
    through the internal/config constants)` (cc and glm subtests, both).
  - AC-SD-005: same selector family → `--- FAIL: TestSD_AC005_LaneLaunchRequiresGitRepo` with
    `factory_m4_test.go:202: refusal does not name the git requirement: NO_ACTIVE_FACTORY`
    (cc and glm subtests, both — the pre-M4 launch refused for the wrong reason and left the
    git requirement unnamed).
  - AC-SD-017 hook arm: `go test ./internal/hook -run '^TestSD_AC017_WidenedRoleGateDenyAndAllow$'`
    → `--- FAIL` with `factory_m4_test.go:40: label-only: decision = "" (reason ""), want deny`.
  - AC-SD-017 cli arm (compile RED — the test's seam is the new exported classifier):
    `internal/cli/factory_m4_test.go:424:27: undefined: hook.CheckContractSignClassify` →
    `FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]`.
  - AC-SD-006 and AC-SD-007 are constraint ACs over machinery M1-M3 already landed: their
    instruments are the M4 deliverable (the recording git wrapper with positive controls; the
    launch-path enumeration walk), and both passed on first run of their final text — there is no
    production change for REQ-SD-006/-007 to RED against, and manufacturing one was not done.
    AC-SD-006's two intermediate failures were test-walk bugs fixed in the test itself (a fresh
    queue-promoted lease resumes through T4b `stage run` before T10; the E-VERDICT PASS file is
    read from the worktree uncommitted and must name the T11-recorded evidence commit), not
    production behavior changes.
- **GREEN (E1, env-scrubbed single compound invocations, final test text)**:
  `go test ./internal/cli -run '^TestSD_AC002_LaneLaunchStampsMarkerAndLabel$|^TestSD_AC005_LaneLaunchRequiresGitRepo$|^TestSD_AC006_LaneCycleWithoutRemote$|^TestSD_AC007_NoHeadlessEngineArgv$|^TestSD_AC017_StampedMarkerArmsContractGuard$' -count=1 -v -timeout 15m` →
  `--- PASS` ×5, `PASS`, `ok github.com/modu-ai/moai-adk/internal/cli 9.249s`. The 017 log line
  records the captured launch environment: `MOAI_FACTORY_ROLE="lane" MOAI_FACTORY_WORKER="lane-1"
  MOAI_KANBAN_BACKEND="claude"`, and the guard classifies `moai contract sign SPEC-X --signer llm`
  under it as `CONTRACT_SIGN_AGENT_VIOLATION:` deny.
  Hook arm: `go test ./internal/hook -run '^TestSD_AC017_WidenedRoleGateDenyAndAllow$|^TestContractSignGuard' -count=1 -v -timeout 5m`
  → `--- PASS: TestSD_AC017_WidenedRoleGateDenyAndAllow` (label-only deny, Codex MCP deny, no-vars
  allow, marker deny) + the guard's own family `--- PASS`, `ok ... 0.631s`.
- **Final anchored regression**: `go test ./internal/cli -run '^TestSD_AC008|...|^TestSD_WorktreeSlugShape|^TestFactoryRoleTokenPinsGuardConstant$|^TestRunCC|^TestFactoryJoin...' -count=1 -timeout 15m`
  → `ok github.com/modu-ai/moai-adk/internal/cli 92.264s`.
- **E2 builds**: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` → both exit 0
  (compound invocation, final tree).
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/cli/... ./internal/hook/...` → after the
  first GREEN run reported 1 NEW govet finding (a self-assignment leftover in the new test, cleaned
  by simplifying the during-callback seam to `func()`), final → `0 issues.` — baseline restored.
  CI version v2.1.6.
- **E3 coverage** (new M4 functions): `-coverprofile` on the M4 selectors —
  `factoryLaneRequiresGitTree` 100.0%, `enterFactoryLaneMode` 100.0% (internal/cli, selector
  AC-SD-002+005); `contractLaneGate` 100.0% (internal/hook, selector AC-SD-017 hook arm + guard
  family — the marker clause is pinned by the hook arm's marker subtest). `CheckContractSignClassify`
  reads 0% under the hook-package selector (its only caller is the cli-arm test) and 75.0% under
  the cli run with `-coverpkg=./internal/hook` (the uncovered line is the `json.Marshal` error
  branch, unreachable for a plain command string). `enterSelectedFactoryRun` 70.6% — pre-existing
  function; the M4-added requireActive guard branch is covered by both AC-SD-002 (guard passes)
  and AC-SD-005 (guard refuses); the remaining legs are the leader/explicit-run branches outside
  M4 scope. Package-wide 5.3%/1.2% under the AC selectors only (whole-package runs prohibited by
  acceptance.md §B — CI owns the full figure).
- **E4 boundary**: `grep -n 'AskUserQuestion' internal/cli/factory.go internal/hook/contract_sign_guard.go`
  → no matches (exit 1).
- **REQ-SD-017 literal scan (production files touched)**:
  `grep -n '"MOAI_FACTORY_ROLE"\|"lane"\|"MOAI_KANBAN_BACKEND"\|"MOAI_FACTORY_WORKER"' internal/cli/factory.go internal/hook/contract_sign_guard.go internal/cli/factory_card.go`
  → exactly one hit: `internal/cli/factory.go:62: factoryLaneRoleToken = "lane"` — the pre-existing
  REQ-AP-013 pin declaration (existing constant, allowed). Zero env-name literals; the stamp is
  `os.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)` and the widened gate compares through
  `config.EnvFactoryRole`/`config.EnvMoaiKanbanLabel`/`config.EnvMoaiKanbanBackend` +
  `kanban.BackendGPT` only. The scan is additionally enforced in-tree by the AC-SD-017 source-scan
  arm, which pins the stamp line verbatim and fails on any new literal.
- **Design decisions**:
  - **Marker stamp placement**: `enterFactoryLaneMode` gains the stamp
    (`config.EnvFactoryRole = config.FactoryRoleLane`, restored on return) — the one function both
    launchers' lane branches call, so cc.go and glm.go stay untouched (cc.go is on the PRESERVE
    list). The leader path (`enterFactoryLeaderMode`) stamps no marker — REQ-SD-002's "the leader
    session gets no marker" holds by construction.
  - **Git requirement placement**: `factoryLaneRequiresGitTree` (one `git -C <root> rev-parse
    --is-inside-work-tree`, one-line refusal naming the git requirement) fires inside
    `enterSelectedFactoryRun` when `requireActive` is set — the parameter is the lane-join
    discriminator (both launchers pass true only on the lane branch), and the check precedes
    `ResolveActiveRun`'s `OpenFactory`, so the refusal writes no registry, factory, or queue file.
    cc.go being read-only forced the shared-entry placement; the leader path (requireActive=false)
    is unaffected.
  - **Widened role gate**: `contractRoleMarker` → `contractLaneGate` — the exact three clauses of
    the cli side's `factoryLaneRefusal` (marker, `EnvMoaiKanbanLabel` non-empty, backend gpt),
    because hook cannot import cli. The Codex MCP environment (the frozen allowlist forwards
    `MOAI_FACTORY_WORKER`, not `MOAI_KANBAN_LABEL`) denies through the backend clause. Known
    residual, shared with the cli side: a cc/glm lane that unsets its marker is not caught by the
    label clause (the lane's label rides `MOAI_FACTORY_WORKER`, which the gate does not read) —
    the M1 predicate's own shape, mirrored not widened here.
  - **CheckContractSignClassify (new exported hook seam)**: the cli arm must classify a captured
    launch environment from the cli package, and `checkContractSign` is unexported; the wrapper
    builds the Bash PreToolUse input and delegates. The deny/allow decision logic stays in the
    unexported guard.
  - **AC-SD-006 instrument**: the §B fixture per the amended acceptance prose — parent on `main`,
    `develop` configured via git-strategy.yaml and checked out in `.moai/worktrees/develop`
    (the t1292 root), no remote; the cycle is next → T4b/T10/T11/T13 stages → window → complete,
    under a PATH-recording git wrapper (the TestFR_AC018 precedent, windows-skipped with the same
    reason) whose positive controls require the log to carry the cycle's own `merge` and `worktree`
    calls, so "no fetch/push" cannot pass on an empty recorder.
  - **AC-SD-007 enumeration**: `sdLaunchPathArgv` entries (cc lane, glm lane) + a walk that already
    knows the codex arm (M5 appends one entry; the walk is unchanged). The swept set is guarded
    (`len(paths) < 2` fails).
- **Fixture cascade (attributable to REQ-SD-005)**: the M4 git requirement tightens the lane-launch
  precondition, so four pre-M4 lane-join tests whose fixtures were plain temp dirs gained
  `initGitRepo`: `TestRunCCFactoriesEntryWritesLane1`, `TestRunCCLiveLegacyClaimRefusedThroughCLI`,
  `TestRunCCDeadLegacyClaimProceedsThroughCLI` (factory_role_refusal_m2_test.go) and
  `codexLedRun` (codex_factory_retire_test.go — one fixture fix covers its five callers). Their
  assertions are unchanged; only the fixture precondition moved. All edited-file families re-ran
  green (`^TestCodex...|^TestFactoryEntry...|^TestRunCC...` selector set → `ok 21.940s`).
- **Stale-doc note (sync-phase concern, not edited here)**:
  `.claude/rules/moai/workflow/contract-sign-guard.md` still describes the role gate as
  "MOAI_FACTORY_ROLE=worker"; the widened deny is described in the guard's own comments and this
  record until manager-docs' sync pass.
- **E6**: no push (lane does not push; the leader batch-pushes develop). M4 commit: this commit
  (SHA reported in the completion report; a commit cannot cite itself).

### M5 — codex per-card relaunch and refusal wording (2026-09-29)

Scope: REQ-SD-003, -004. ACs: AC-SD-003, -004, plus the AC-SD-007 enumeration extension and the
AC-SD-017 scan-map extension (new stamp site). Continuation at `4d73ee088` (clean).

- **Pre-flight (C)**: `git rev-parse --short HEAD` → `4d73ee088`; `git branch --show-current` →
  `WT-factory-self-dispatch`; `git status --short` → empty; `go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m
  ./internal/cli/...` → `0 issues.` (baseline; CI version v2.1.6). Anchored regression before any
  change: `go test ./internal/cli -run '^TestSD_AC002|^TestSD_AC005|^TestSD_AC006|^TestSD_AC007|^TestSD_AC008|^TestSD_AC009|^TestSD_AC010|^TestSD_AC011|^TestSD_AC012|^TestSD_AC013|^TestSD_AC015|^TestSD_AC016|^TestSD_AC017|^TestSD_AC023|^TestSD_AC024|^TestSD_AC025|^TestSD_WorktreeSlugShape' -count=1 -timeout 15m`
  → `ok github.com/modu-ai/moai-adk/internal/cli 81.228s` (slot lease `moai slot acquire --resource
  internal-test-suite` held for the heavy runs, released after).
- **RED (E8, verbatim, pre-GREEN tree = `4d73ee088`, captured in `/tmp/m5_red.txt`)**:
  - AC-SD-003: `--- FAIL: TestSD_AC003_CodexRelaunchPerCard` with `factory_m5_test.go:115: codex
    lane: ` — the pre-M5 `-f lane` scanner refused the entry (exitCodeError{1}, empty message), so
    the launcher errored before any child.
  - AC-SD-007 (extension): `--- FAIL: TestSD_AC007_NoHeadlessEngineArgv` with
    `factory_m4_test.go:395: codex lane: ` — same refusal reached the new enumeration entry.
  - AC-SD-004: `--- FAIL: TestSD_AC004_CodexOtherFactoryShapesRefused` ×4 with
    `factory_m5_test.go:198: the refusal line does not name "moai codex -f lane":
    "FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Factory Mode; use 'moai cc -f'
    or 'moai glm -f' instead"` — the old wording lacked the only-entry name (all four shapes).
  - First RED attempt also surfaced a test bug (build failure):
    `factory_m5_test.go:188:11: errB.String undefined (type string has no field or method String)`
    — runCodexCmd returns strings, not buffers; fixed in the test before the real RED above.
- **GREEN (E1, env-scrubbed single compound invocations)**:
  `go test ./internal/cli -run '^TestSD_AC003_CodexRelaunchPerCard$|^TestSD_AC004_CodexOtherFactoryShapesRefused$|^TestSD_AC007_NoHeadlessEngineArgv$|^TestSD_AC017_StampedMarkerArmsContractGuard$' -count=1 -timeout 10m -v`
  → `--- PASS` ×4, `PASS`. AC-SD-003's evidence chain: the substituted session is invoked twice;
  invocation 1 carries `-C <t1 worktree>` + marker `lane` + label + `MOAI_KANBAN_BACKEND=gpt` +
  `MOAI_KANBAN_CARD=t1` (and `MOAI_KANBAN_LABEL=t1`'s label); invocation 2 the same for t2; both
  cards end `merge-ready`; the launcher exits 0 after `next` reports no card. The substituted
  session performs its card's stage chain (T4b → T10 → T11 → T13) inside the launch seam — the
  real child's own work, simulated per acceptance.md §B (no real codex binary anywhere).
- **Final anchored regression**: `go test ./internal/cli -run '^TestSD_AC002|^TestSD_AC003|^TestSD_AC004|^TestSD_AC005|^TestSD_AC006|^TestSD_AC007|^TestSD_AC008|^TestSD_AC009|^TestSD_AC010|^TestSD_AC011|^TestSD_AC012|^TestSD_AC013|^TestSD_AC015|^TestSD_AC016|^TestSD_AC017|^TestSD_AC023|^TestSD_AC024|^TestSD_AC025|^TestSD_WorktreeSlugShape|^TestCodex' -count=1 -timeout 15m`
  → one FAIL: `TestCodexAuditMCPTool` — **pre-existing on this tree, outside this milestone's
  diff**: the test reads `.claude/rules/moai/core/moai-mcp-tools.md` (says 39 tools) +
  `internal/mcp` catalogue (45 registered), neither touched by M5 (`git diff --name-only HEAD` →
  `internal/cli/codex_launcher.go`, `internal/cli/factory_m4_test.go`; `factory_m5_test.go`
  untracked). A doc-lag from a develop absorption; NEW-vs-baseline classification: pre-existing.
  Everything else in the combined run is green (the full M1-M5 AC set + the whole `TestCodex`
  launcher family).
- **E2 builds**: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` → both exit 0
  (compound invocation, final tree). No new syscall use, no process replacement on the relaunch
  path — the supervising loop keeps the launcher as parent on every platform.
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.` — no NEW vs
  baseline (CI version v2.1.6).
- **E3 coverage** (new M5 functions, `-coverprofile` on the AC selectors, internal/cli):
  `codexFactoryEntryClassify` 100.0%, `stripCodexFactoryTokens` 87.5%, `runCodexFactoryLane`
  75.7% (uncovered: the error legs — install hint, parent-checkout refusal, lease/worktree
  machinery failures — whose machinery the M1-M3 ACs cover where it lives), `launchCodexCardSession`
  81.2%, `codexCardLaunchEnv` 100.0%.
- **E4 boundary**: `grep -n 'AskUserQuestion' internal/cli/codex_launcher.go
  internal/cli/factory_m5_test.go` → no matches (exit 1).
- **REQ-SD-017 literal scan (production files touched)**:
  `grep -n '"MOAI_FACTORY_ROLE"\|"MOAI_KANBAN_CARD"\|"MOAI_KANBAN_BACKEND"\|"MOAI_FACTORY_WORKER"\|"MOAI_KANBAN_LABEL"\|"gpt"' internal/cli/codex_launcher.go`
  → no matches (exit 1). Zero env-name or backend-value literals; every stamp site reads
  `config.Env*` / `kanban.BackendGPT`. The AC-SD-017 source-scan map now also pins
  `codex_launcher.go` (the new stamp site) with the same four-literal sweep.
- **Guard tests (existing family, surfaced by the wider TestCodex regression)**: three first-run
  failures were M5 text shapes caught by the pre-existing codex guards, fixed in M5's own code:
  (a) the @MX:NOTE on the loop carried the literal `syscall.Exec` (a comment — but
  `TestCodexSpecFiles_NoBuildTagsOrSyscall` and the cross-platform property score scan the raw
  file) → reworded to "no process replacement happens on this path";
  (b) `TestCodexSpecFiles_ExecPrimitivesCodexOnly` requires every `exec.Command` first argument in
  codex_launcher.go to be `req.Program` → the child launch now builds `codexLaunchRequest` and
  starts from `req.Program`/`req.Args`/`req.Dir`, matching the direct path's convention. All
  three guards re-ran green with the M5 AC set (`ok 10.669s`).
- **Design decisions**:
  - **Classifier replaces the scanner**: `codexEntryRefusal` → `codexFactoryEntryClassify`
    returning (entry, diag): Absent / Lane (`-f lane`, `--factory lane`, `=` forms — value tokens
    mirror `parseFactoryFlag`'s spellings) / Other (bare `-f`, `--factory`, `--factory-run[=x]`,
    `-f lane-<n>`, `=` forms with another value, kanban tokens → kanban refusal, any second
    factory token → refusal wins). Legacy role tokens (`worker`/`agent`) classify Other for now —
    the shape leaves t1256's producer branch one check away (M7, AC-SD-021).
  - **Supervising loop placement**: `runCodexFactoryLane` in codex_launcher.go. Per card: lease
    through the F1 machinery in-process (`factoryNextLeaseOnce` + `factoryEnsureCardWorktree` —
    the same functions the `next` RunE calls; the launcher runs them on the PARENT checkout and
    asserts that precondition like the verb does, REQ-SD-010), then ONE interactive child
    (`codex -C <worktree>` + the worktree's local instruction files) through the existing
    `codexDirectLaunchFn` seam, stdio = the parent's own, then wait and loop. Stop condition:
    `next`'s no-card answer, which the REQ-SD-025 merge-ready skip feeds (D2 — no livelock). A
    child that fails or exits non-zero is reported to stderr and the loop continues (REQ-SD-003:
    "on that session's exit continue with the next card"). No `--wait`: the loop itself is the
    wait; when nothing qualifies it exits 0.
  - **Undefined combinations refuse through the usage constant**: `-f lane` with `--spawn`, `-w`,
    a verb token, or a `--` tail exits 1 with `codexUsageDiag` — the loop must stay the parent and
    selects the worktree itself; the closed-set discipline refuses rather than adapts.
  - **Label carriers, both stamped (the M4 residual, narrowed)**: the factory card verbs
    (`next`/`stage`/`complete`) and the widened role gate read the lane label from
    `MOAI_KANBAN_LABEL` (M4's record already noted the gate "does not read" the launcher's
    carrier), while the launcher stamp and the factory notices read `MOAI_FACTORY_WORKER`
    (REQ-RNC-011). The AC-SD-003 RED first surfaced this as a live integration gap: the launcher's
    own stamps left `stage` refusing with "MOAI_KANBAN_LABEL is empty". M5's codex launcher stamps
    BOTH carriers — in the loop's process environment and in the child environment — so the codex
    lane is fully functional and the refusal predicates' label clause now catches it too. **Left
    open for the lead (blocker, M4 territory)**: a cc/glm lane session still carries its label
    only in `MOAI_FACTORY_WORKER`, so the session's own `next`/`stage`/`complete` refuse with the
    empty-label error; the one-line fix is the same paired stamp inside `enterFactoryLaneMode`
    (factory.go — read-only for this dispatch).
  - **Factory fan-out signal not carried to the child**: `MOAI_FACTORY_WORKERS` feeds the
    Stop-hook block cap, and a Codex lane has no moai hook peer (REQ-CFR-022) — recorded in the
    child-env builder's comment.
- **E6**: no push (lane does not push; the leader batch-pushes develop). M5 commit: this commit
  (SHA reported in the completion report; a commit cannot cite itself).

### Label-carrier unification (2026-09-29)

Defect: the lane-label reads in the factory card verbs and the widened role gate used
`config.EnvMoaiKanbanLabel` while the design's canonical carrier is `config.EnvMoaiFactoryWorker`
(design.md §3/§5; `MOAI_KANBAN_LABEL` is absent from the frozen Codex MCP env_vars allowlist,
`internal/codexwiring/configtoml.go:21`), so a cc/glm lane stamped `MOAI_FACTORY_WORKER` only
could not run its own next/stage/complete (the M5 "Left open for the lead" blocker, closed here).

Fix sites (reads converged to the carrier; stamps unchanged — the M5 dual stamp stays
deliberately):

- `internal/cli/factory_card.go:52` (laneRefuse label clause — covers the todo queue guard and the
  decide guard via `factoryLaneRefusal`), `:434`/`:436` (next), `:520`/`:522` (stage),
  `:613`/`:615` (complete) — error texts renamed with the same sentence shape.
- `internal/cli/mcp_factory_card.go:115`/`:117` (factory_next), `:159` (factory_stage),
  `:175`/`:177` (factory_complete) — same defect on the M3 MCP surface, found by the ordered
  re-grep (the dispatch's read-site list was not complete; included as the same defect class).
- `internal/hook/contract_sign_guard.go:148` (M4 gate label clause).
- Fixtures switched to `config.EnvMoaiFactoryWorker`: `factory_self_dispatch_test.go:44`/`:53`
  (sdLaneEnv/sdClearLaneEnv) and `:458` (AC-SD-015 label-only); `factory_m3_test.go:338`/`:628`/
  `:661`; `internal/hook/factory_m4_test.go:37`/`:46` (+ one comment sentence). Kept as-is per
  dispatch: the M5 dual-stamp assertion (`factory_m5_test.go:145-146`), stamp/scrub lists,
  kanban-surface tests, `launcher_blockcap_infinite` (different predicate).
- Doc lag (pre-existing, M3 catalogue growth 39→45): `.claude/rules/moai/core/moai-mcp-tools.md`
  (39→45 ×2, project_root section 13→16 optional-root tools + the three required-root lane verbs)
  and `moai-mcp-tools-catalogue.md` (45-tool frontmatter/intro/header, family header 41-of-45, new
  Factory card verbs family row + per-tool section) — both mirrored byte-identical to
  `internal/template/templates/` and `make build` rerun. `internal/mcp/catalog.go` untouched.

Evidence (env-scrubbed single compound invocations, slot lease `internal-test-suite` held and
released):

1. Full anchored set M1-M5 + `^TestCodex`: `go test ./internal/cli -run '^TestSD_AC002|^TestSD_AC003|^TestSD_AC004|^TestSD_AC005|^TestSD_AC006|^TestSD_AC007|^TestSD_AC008|^TestSD_AC009|^TestSD_AC010|^TestSD_AC011|^TestSD_AC012|^TestSD_AC013|^TestSD_AC015|^TestSD_AC016|^TestSD_AC017|^TestSD_AC023|^TestSD_AC024|^TestSD_AC025|^TestSD_WorktreeSlugShape|^TestCodex' -count=1 -timeout 15m`
   → `ok github.com/modu-ai/moai-adk/internal/cli 125.735s`.
2. Hook arm: `go test ./internal/hook -run '^TestSD_AC017_WidenedRoleGateDenyAndAllow$|^TestContractSignGuard' -count=1`
   → `ok github.com/modu-ai/moai-adk/internal/hook 0.668s`.
3. `go test ./internal/cli -run '^TestCodexAuditMCPTool$|^TestMCPToolCatalogueFiguresMatchRegistry$|^TestMCPToolCatalogueDocsStayMirrorIdentical$' -count=1 -v`
   → `--- PASS: TestCodexAuditMCPTool (3.98s)`, `--- PASS: TestMCPToolCatalogueDocsStayMirrorIdentical (0.00s)`,
   `--- PASS: TestMCPToolCatalogueFiguresMatchRegistry (0.00s)`.
4. `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` → both exit 0.
5. `golangci-lint run --timeout=2m ./internal/cli/... ./internal/hook/...` (v2.1.6, the CI version)
   → `0 issues.`
6. `cmp` stub pair and companion pair → identical, both.
7. Carrier end-to-end: `go test ./internal/cli -run '^TestSD_AC015_LabelOnlyIsNotALane$|^TestSD_AC002_LaneLaunchStampsMarkerAndLabel$' -count=1 -v`
   → `--- PASS: TestSD_AC002_LaneLaunchStampsMarkerAndLabel (1.98s)` (launcher-stamped env feeds the
   actor path), `--- PASS: TestSD_AC015_LabelOnlyIsNotALane (0.49s)` (label-only fixture now denies
   via `MOAI_FACTORY_WORKER`).
8. Post-edit readback of every edited region quoted (production 12 sites, fixtures 9 lines, docs 6
   regions); residual grep `EnvMoaiKanbanLabel` over production files → only the legitimate
   kanban-surface/dual-stamp sites remain (launcher_blockcap, codex_launcher ×3, kanban.go ×2,
   ptycaptest, session_start_kanban/record).

Commit: this commit (SHA reported in the completion report; a commit cannot cite itself).

### Served-model-gate disposition (2026-09-29, operator via lead)

- Basis: 운영자 승인(리드 전달 2026-09-29): 해제+GLM 채택 — option (a) of the lane's three-option
  request. Context: two lead-requested compact-delta re-audits (t1307, t1308) were spawned from this
  session on `model: opus` and were served `glm-5.3-flash`; both auditors honestly refused to certify
  an opus verdict and the served_model_gate recorded rejection receipts. The receipts landed in THIS
  tree's gate state and the gate then refused the M6 `manager-develop` spawn
  (`SERVED_MODEL_VIOLATION`, outstanding: plan-auditor / unknown-spec / expected opus, served
  glm-5.3-flash). Re-running on the expected model is measured impossible from a GLM lane (both
  t1307 and t1308 spawns observed glm-5.3-flash serving despite the explicit `model: opus`
  argument).
- Change: `.moai/config/sections/workflow.yaml` — `workflow.served_model_gate.enabled`: `true` →
  `false` (this card tree only; the template twin already ships the default `false`, untouched).
  Observation/warning behavior per the config comment is unaffected (only the refusal is gated).
- Receipts: the existing rejection receipt files are PRESERVED (not deleted) per the lead's
  instruction — gate OFF invalidates them; the record remains for audit.
- Verdicts adopted: t1307 (SPEC-TODO-STALE-STORE-001, baseline `91eaeff8d`) RECONFIRMED iter2 PASS
  0.95; t1308 (SPEC-TODO-HOLD-STATE-001, baseline `d69b71c6d`) RECONFIRMED iter2 PASS 0.96 — both
  on the auditors' own hash evidence (HEAD == baseline, plan artifacts byte-identical 4/4), served
  model disclosed in each verdict's first line. Verdict files: `.moai/reports/t1307/verdict.md`,
  `.moai/worktrees/t1308/.moai/reports/t1308/verdict.md` (their owning trees).
- Changed at 2026-09-29 (lane worker-70 session time), committed with this record in one commit.

### M6 — sessionstart next-card rule and clear policies (2026-09-29)

Scope: REQ-SD-019, -020. ACs: AC-SD-019, -020. Continuation at `3eeaabe5b` (clean).

- **Blocker resolution (carrier decision)**: the Section A REQ-SD-022 pre-check found NO existing
  constant carrying a clear-policy value (full `internal/config` enumeration: 42 `MOAI_*` names,
  none policy-bearing; zero clear-policy code in production cli) and STOPPED with a structured
  blocker report. The lead post-approved Option A; manager-spec amendment `3eeaabe5b` landed the
  §D carve-out ("any new environment variable name except the lane clear-policy carrier named in
  REQ-SD-020"), the REQ-SD-020 carrier sentence, and the design §6 constant naming
  (`config.EnvFactoryClearPolicy`, `MOAI_FACTORY_CLEAR_POLICY`); `go test ./internal/spec -count=1`
  → `ok 152.784s` (lead's run). The Codex MCP `env_vars` allowlist stays byte-identical — Codex
  lanes take no policy and the name never enters it.
- **Pre-flight (C)**: `git rev-parse --short HEAD` → `3eeaabe5b`; `git branch --show-current` →
  `WT-factory-self-dispatch`; `git status --short` → empty; `go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m
  ./internal/cli/... ./internal/hook/...` → `0 issues.` (baseline; CI version v2.1.6).
- **RED (E8, verbatim, pre-GREEN tree = `3eeaabe5b`)** — both anchored tests failed to build with
  the feature absent:
  - hook: `internal/hook/session_start_factory_rule_test.go:25:10: undefined:
    config.EnvFactoryClearPolicy` … `:69:21: undefined: config.EnvFactoryClearPolicy` …
    `:70:13: undefined: factoryLaneRuleForSource` … `FAIL
    github.com/modu-ai/moai-adk/internal/hook [build failed]`
  - cli: `internal/cli/factory_m6_test.go:59:2: undefined: factoryPrintClearPolicyLine` …
    `:90:19: undefined: config.EnvFactoryClearPolicy` … `FAIL
    github.com/modu-ai/moai-adk/internal/cli [build failed]`
- **GREEN (E1, verbatim)**:
  - `go test ./internal/hook -run '^TestSD_AC019_NextCardRuleInjection$' -count=1 -v` → 10/10
    `--- PASS` subtests (claude × 3 policies × startup × 4 locales; glm; source clear;
    resume/compact none; gpt owned-card + negatives; gpt no-card none; leader none; keyless none;
    legacy label none; unknown backend none), `PASS`, `ok
    github.com/modu-ai/moai-adk/internal/hook 0.582s`.
  - `go test ./internal/cli -run '^TestSD_AC020_ClearPolicies$' -count=1 -v` → 14/14 `--- PASS`
    subtests (clear-each one line; absent = clear-each; when-full below/at/large-window/missing-
    record/refused-key; relaunch end-session line; relaunch supervising loop end-to-end — two
    cards, each child dir = its card worktree, env carries `MOAI_KANBAN_CARD` + the policy + the
    lane marker/label + `--name <label>` argv; complete prints the policy line end-to-end through
    the merge flow; relaunch without the claude binary refused; loop continues after a failed
    child; non-lane / leader / invalid-value refusals), `PASS`, `ok
    github.com/modu-ai/moai-adk/internal/cli 9.249s`.
- **Final anchored regression**: env-scrubbed single compound
  `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_FACTORY_CLEAR_POLICY && go test ./internal/cli -run '^TestSD_AC002|^TestSD_AC003|^TestSD_AC004|^TestSD_AC005|^TestSD_AC006|^TestSD_AC007|^TestSD_AC008|^TestSD_AC009|^TestSD_AC010|^TestSD_AC011|^TestSD_AC012|^TestSD_AC013|^TestSD_AC015|^TestSD_AC016|^TestSD_AC017|^TestSD_AC023|^TestSD_AC024|^TestSD_AC025|^TestSD_WorktreeSlugShape' -count=1 -timeout 15m`
  → `ok github.com/modu-ai/moai-adk/internal/cli 51.775s` (slot lease
  `moai slot acquire --resource internal-test-suite --max-duration 20m` held, released after);
  hook anchor `go test ./internal/hook -run '^TestSD_AC017_WidenedRoleGateDenyAndAllow$|^TestContractSignGuard' -count=1`
  → `ok 0.429s`.
- **E2 builds** (final tree): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
  exit 0. The relaunch path adds no syscall use and no process replacement — the supervising loop
  stays the parent (design §6), Windows via the existing child-process form.
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/cli/... ./internal/hook/...` →
  `0 issues.` — no NEW vs baseline (CI version v2.1.6), re-run after the final test additions.
- **E3 coverage** (new/changed M6 functions, `-coverprofile` on the AC selectors; profiles at
  `.moai/state/verify/t1240-m6/{hook,cli}-cover.out`): `factoryLaneRuleForSource` 92.9% (hook);
  `factoryClearPolicySelected` 100.0%, `factoryContextAtHandoffThreshold` 92.3%,
  `factoryPrintClearPolicyLine` 100.0%, `runFactoryLaneRelaunch` 87.0% (uncovered: the
  lease/worktree machinery error legs, whose machinery the M1-M3 ACs cover where it lives),
  `launchFactoryLaneCardSession` 100.0%, `enterFactoryLaneMode` (extended signature) 100.0% — all
  ≥ 85%.
- **E4 boundary**: `grep -n 'AskUserQuestion'` over all ten touched production files
  (envkeys/factory/kanban/cc/glm/codex_launcher/factory_card/factory_lane_relaunch +
  session_start_factory{,_i18n}/session_start) → no matches (exit 1).
- **REQ-SD-019 negative-content proof** (real handler path: `moai hook session-start` with a
  startup payload, per-locale `MOAI_CONFIG_DIR`; captures at
  `.moai/state/verify/t1240-m6/rule-{en,ko,ja,zh,gpt}.json`). The Codex-rule output, quoted
  verbatim: "Factory lane owned-card rule: this session owns card t9 — the id recorded in
  MOAI_KANBAN_CARD — and works in the current worktree. Carry that card to merge-ready: record
  each stage with `moai factory stage` and finish with `moai factory complete`. Then end this
  session; do not take or lease any other card." Mechanical check (`.moai/state/verify/t1240-m6/
  gpt-negative-check.txt`): `factory next` absent, `factory_next` absent, `todo_add` absent,
  `todo_list` absent, `factory_stage` absent, `factory_complete` absent, `factory_decide` absent;
  `t9` present, `moai factory stage` present, `moai factory complete` present.
- **REQ-SD-019 4-locale Claude-rule outputs** (quoted once per locale, verbatim from the
  additionalContext captures):
  - en: "Factory lane next-card rule: this session is a self-dispatch lane. Leave any kept
    worktree as it is — never remove a card worktree. Take the next card from the parent checkout:
    run `moai factory next` (MCP tool `factory_next`); lane queue promotion through `moai factory
    next` is operator-authorized for the self-dispatch lane mode. Enter the card's worktree, carry
    the card through plan, run, and sync, integrate it, leave its worktree kept, and record
    completion with `moai factory complete` (MCP tool `factory_complete`). Then follow the clear
    policy the completion output prints. Tools and their CLI equivalents: `todo_add` (`moai todo
    add`), `todo_list` (`moai todo`), `factory_next` (`moai factory next`), `factory_stage`
    (`moai factory stage`), `factory_complete` (`moai factory complete`), `factory_decide`
    (`moai factory decide`)."
  - ko: "팩토리 레인 다음 카드 규칙: 이 세션은 셀프 디스패치 레인입니다. kept 로 남은 워크트리는 그대로
    둡니다 — 카드 워크트리를 절대 제거하지 않습니다. 부모 체크아웃에서 다음 카드를 가져옵니다. `moai
    factory next`(MCP 도구 `factory_next`)를 실행하세요. 셀프 디스패치 레인 모드에서 `moai factory
    next` 를 통한 레인 큐 승격은 운영자가 승인했습니다. 카드의 워크트리에 진입해 plan, run, sync 로
    카드를 끝까지 수행하고, 통합한 뒤에도 워크트리는 kept 로 남기고, `moai factory complete`(MCP 도구
    `factory_complete`)로 완료를 기록합니다. 그다음에는 완료 출력이 알려 주는 clear 정책을 따릅니다.
    도구와 CLI 등가물: `todo_add`(`moai todo add`), `todo_list`(`moai todo`),
    `factory_next`(`moai factory next`), `factory_stage`(`moai factory stage`),
    `factory_complete`(`moai factory complete`), `factory_decide`(`moai factory decide`)."
  - ja: "ファクトリーレーン次カード規則: このセッションはセルフディスパッチレーンです。 kept のワークツリーはそのまま残します
    — カードのワークツリーを削除してはいけません。 親チェックアウトから次のカードを取得します。`moai factory
    next`(MCP ツール `factory_next`)を実行してください。 セルフディスパッチレーンモードでは、`moai factory
    next` によるレーンキューの昇格はオペレーターが承認済みです。 カードのワークツリーに入り、plan・run・sync
    を通してカードを完遂し、統合したうえでワークツリーを kept のまま残し、 `moai factory complete`(MCP ツール
    `factory_complete`)で完了を記録します。 その後は、完了出力が示す clear ポリシーに従います。 ツールと CLI
    の対応: `todo_add`(`moai todo add`)、`todo_list`(`moai todo`)、`factory_next`(`moai factory
    next`)、 `factory_stage`(`moai factory stage`)、`factory_complete`(`moai factory complete`)、
    `factory_decide`(`moai factory decide`)。"
  - zh: "工厂泳道下一张卡规则：本会话是自调度泳道。 保持 kept 状态的工作树原样保留 — 绝不删除卡片的工作树。
    从父检出获取下一张卡：运行 `moai factory next`（MCP 工具 `factory_next`）。 在自调度泳道模式下，通过
    `moai factory next` 进行的泳道队列提升已获运营者授权。 进入卡片的工作树，带着卡片走完 plan、run、sync，完成集成后工作树保持
    kept， 并用 `moai factory complete`（MCP 工具 `factory_complete`）记录完成。 之后遵循完成输出给出的 clear
    策略。 工具及其 CLI 等价物：`todo_add`（`moai todo add`）、`todo_list`（`moai
    todo`）、`factory_next`（`moai factory next`）、 `factory_stage`（`moai factory
    stage`）、`factory_complete`（`moai factory complete`）、`factory_decide`（`moai factory
    decide`）。"
  All six MCP tool names with their CLI equivalents appear in every locale; skill/verb/tool names
  stay English inside every locale (REQ-SD-019 i18n).
- **Design decisions**:
  - **One carrier, one stamp site**: `config.EnvFactoryClearPolicy` (+ the three policy value
    constants) joined `internal/config/envkeys.go` per the approved Option A; the stamp lives in
    `enterFactoryLaneMode` (now `(label, lanes, clearPolicy)`) and ALWAYS overwrites — a policy
    inherited from an outer session can never leak into a lane launched without one. Codex passes
    `""`. Selection enters through `--clear-policy <clear-each|clear-when-full|relaunch>` parsed in
    `parseFactoryFlag` (both `--flag v` and `--flag=v` forms, value validated against the three
    constants); a non-factory launch or the leader shape carrying it is REFUSED with one line (the
    parser's typo-honesty — silence would mask a mistake), and the token is stripped from the claude
    argv. Absence reads as `clear-each` on every reader (`factoryClearPolicySelected` fail-open).
  - **Rule vs bootstrap split**: the bootstrap notice stays startup-only; the rule is the
    session-cycle half — `factoryLaneRuleForSource` fires on startup (the policy value is
    deliberately never read there, which is what makes "under every clear policy" true by
    construction) and on clear, re-entering the fresh session to keep the clear-each loop going.
    Leader/legacy/keyless/unknown-backend/gpt-without-card all receive no rule. Locale =
    `operatorLang(h.cfg)` (the session's conversation language, REQ-SD-019), not the bootstrap's
    agent-facing English.
  - **clear-when-full reads the authoritative record**:
    `statusline.SessionTelemetryPath` + `ReadSessionTelemetry` on
    `<root>/.moai/state/context-usage/<session-id>.json` (session id from
    `config.EnvClaudeCodeSessionID`), `RawPct` compared with the model-specific handoff threshold
    recomposed from `config.HandoffLargeWindowCutoff` / `HandoffSoftLargePct` /
    `HandoffSoftStandardPct` (no literals, §14). Missing record, refused key, or unparseable record
    reads as BELOW (the lane keeps working).
  - **Relaunch loop** (`factory_lane_relaunch.go`): mirrors the M5 codex loop — parent-checkout
    assert, `factoryNextLeaseOnce` + `factoryEnsureCardWorktree` in-process, then ONE interactive
    child per card through the `factoryLaneCardLaunchFn` seam (`claude` with the lane argv the
    branch built — session name + injected settings — cwd = the card worktree, inherited env +
    `MOAI_KANBAN_CARD`). Stop condition: `next`'s no-card answer. A child that fails to start or
    exits non-zero is reported to stderr and the loop continues. No `syscall` on this path.
  - **Fixture hygiene (t1222/t1217)**: `MOAI_FACTORY_CLEAR_POLICY` joined `sdScrubLauncherEnv`
    (factory_m4_test.go) and the env-scrub compound unset list; `TestSD_M6_ClearPolicyCarrierConstantOnly`
    extends the M4 literal-scan pattern to the M6-touched files (the carrier name exists only as
    the internal/config constant).
- **E6 commit**: this section and the M6 change land in one commit
  `feat(SPEC-FACTORY-SELF-DISPATCH-001): M6 sessionstart next-card rule and clear policies
  (card t1240)`; push = none (the lead pushes develop in batch).

### M7 — invariants, vocabulary, doctrine amendments (2026-09-29)

Scope: REQ-SD-001 (via the AMENDED AC-SD-001 — verification only), REQ-SD-015's doctrine
amendments, REQ-SD-018, REQ-SD-021, REQ-SD-022. ACs: AC-SD-001 (amended), -015 doctrine
re-verifies, -018, -021, -022. Continuation at `f37942d79` (clean).

- **Scope-doc amendment record**: `f37942d79` — AC-SD-001 rewritten to the commit-graph-verifiable
  run-gate form (manager-spec via re-delegation; `go test ./internal/spec -count=1` → `ok 130.298s`;
  ac-baseline-guard AC COUNT 25 unchanged).
- **Pre-flight (C)**: `git rev-parse --short HEAD` → `f37942d79`; `git branch --show-current` →
  `WT-factory-self-dispatch`; `git status --short` → empty; `go build ./...` → `BUILD_LINUX_OK`;
  `GOOS=windows GOARCH=amd64 go build ./...` → `BUILD_WINDOWS_OK`; `golangci-lint run --timeout=2m
  ./internal/cli/... ./internal/hook/... ./internal/codexwiring/...` → `0 issues.` (baseline;
  CI version v2.1.6); anchored regression at the baseline tree → `ok github.com/modu-ai/moai-adk/
  internal/cli 59.268s`, hook → `ok 0.580s` (slot lease held, released after).
- **AC-SD-001 (amended form — verification only)**, verbatim outputs at `f37942d79`:
  - `git merge-base --is-ancestor abb815921 490d64649 && echo OK` → `OK`
  - `grep -c 'develop_sha.*a7190891d' .moai/specs/SPEC-FACTORY-SELF-DISPATCH-001/progress.md` → `1`
  - `grep -c 't1256_landed.*: yes' .moai/specs/SPEC-FACTORY-SELF-DISPATCH-001/progress.md` → `1`
- **RED (E8, verbatim, pre-GREEN tree = `f37942d79`)**:
  - AC-021 (`go test ./internal/cli -run '^TestSD_AC021_LegacySpellingsRefused$' -count=1 -timeout 10m`):
    `--- FAIL: TestSD_AC021_LegacySpellingsRefused (3.14s)` with two failing subtests —
    `moai_codex`: 7× `factory_m7_test.go:99: -f <token>: refused with the REQ-SD-004 line; AC-SD-021
    wants the REQ-RNC message naming the canonical form:` + `FACTORY_MODE_UNSUPPORTED_BACKEND: moai
    codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory
    leader` and 7× `:103: … does not name "legacy role token"/"legacy lane label"/"legacy leader
    spelling"`; `lane_verbs…`: `:131: label "worker-1", stage: refusal "Error: factory stage: no
    --run given and no single active factory run (NO_ACTIVE_FACTORY)…" does not name the legacy
    spelling` (stage/complete reached run resolution before any legacy check existed); `next` showed
    the parent-checkout error — no legacy refusal anywhere. `FAIL
    github.com/modu-ai/moai-adk/internal/cli 3.903s`.
  - AC-018 observed-failure evidence (the first two runs, before the fixture mirrored the real
    checkout's tracked config + runtime gitignore): run 1 —
    `factory_m7_test.go:440: parent checkout is dirty outside the worktrees directory: "?? .moai/"`;
    run 2 (judgement switched to `--untracked-files=all`) — four dirty lines verbatim:
    `?? .moai/config/sections/git-strategy.yaml`, `?? .moai/db/002-eedd77fa/factory/factory.db`,
    `?? .moai/db/002-eedd77fa/project.json`, `?? .moai/reports/t1/merge-record.txt`. These runs are
    also the judgement's observed-failure proof: real parent dirt fails it.
  - AC-022 mutant probes (freeze tests are green at arrival — the observed failure comes from a
    mutant, verification-completeness §1.1): allowlist mutant (`MOAI_MUTANT_PROBE` appended) →
    `--- FAIL: TestSD_AC022_EnvVarsAllowlistFrozen (0.00s)` /
    `factory_freeze_test.go:21: mcpServerEnvVarsValue drifted from the run-start snapshot
    a7190891d:` / `FAIL github.com/modu-ai/moai-adk/internal/codexwiring 0.503s`; schema mutant
    (`workers` → `workers_mutant_probe`) → `--- FAIL: TestSD_AC022_SchemaStatementsFrozen (0.03s)` /
    `factory_m7_test.go:345: schema statement drifted from the a7190891d snapshot (want 1, got 0):` /
    `factory_m7_test.go:350: schema statement ADDED since the a7190891d snapshot (1 occurrence(s)):`
    / `FAIL … internal/cli 0.819s`. Both mutants reverted (Edit), re-run green.
- **GREEN (E1, verbatim)**:
  - `go test ./internal/cli -run '^TestSD_AC021_LegacySpellingsRefused$' -count=1 -timeout 10m -v`
    → 5/5 `--- PASS` subtests (launcher parse moai cc/moai glm — 9 rows incl. letter-case; moai
    codex — 7 legacy tokens each refused with the RNC message, NOT the REQ-SD-004 line, exit 1, no
    child; lane verbs next/stage/complete refuse a legacy lane-identity label ×5 tokens; verbs take
    no lane-label positional; negative source scan — role-noun regex over factory*.go +
    mcp_factory*.go + codex_launcher.go string literals, 2 allowlisted non-role literals with a
    stale-entry guard and a scanned>0 positive control), `PASS`, `ok … internal/cli 0.786s`.
  - `go test ./internal/cli -run '^TestSD_AC022_SchemaStatementsFrozen$' -count=1 -timeout 5m` →
    `ok … internal/cli 0.627s`; `go test ./internal/codexwiring -run
    '^TestSD_AC022_EnvVarsAllowlistFrozen$' -count=1 -timeout 5m` → `ok … internal/codexwiring
    0.514s`. Both compare against the `a7190891d` run-start snapshot with the expected content
    embedded and its provenance named (a test cannot read git at CI time reliably).
  - `go test ./internal/cli -run '^TestSD_AC018_ParentCheckoutUntouched$' -count=1 -timeout 10m -v`
    → `--- PASS: TestSD_AC018_ParentCheckoutUntouched (2.31s)` / `ok … internal/cli 3.111s` — the
    AC-SD-006 full cycle (parent on main, develop provisioned at `.moai/worktrees/develop`, no
    remote), then the parent judgement: porcelain (untracked enumerated) excluding
    `.moai/worktrees/` empty, HEAD and branch equal pre-cycle values.
- **REQ-SD-015 doctrine amendments (the OD-1/OD-2 deliverables)**:
  - `.claude/rules/moai/workflow/kanban-dispatch.md` + its template twin: a `[HARD] **The
    self-dispatch lane exception.**` paragraph inserted next to the "Promotion is the operator's
    act, always" clause — leasing only through `moai factory next`, every other queue mutation
    (`add`, `drop`, `done`, `edit`, and the rest) and `moai contract sign` stay forbidden to a lane.
    Neutral wording (no card ids, no SPEC ids, no dates, no SHAs).
  - `.claude/rules/local/gitflow-lane-protocol.md` §6: the self-dispatch-lane merge-window
    exception (Claude self-dispatch lane holds the integration window itself through `moai factory
    complete`) plus the `moai factory next` leasing exception, each stating the other-queue-mutation
    + `moai contract sign` prohibition.
  - `CLAUDE.local.md` §4.1: **BLOCKER (lead-applied)** — the file is gitignored (`/CLAUDE.local.md`,
    .gitignore:276), untracked, and exists ONLY in the primary checkout; this isolated worktree has
    no copy and the harness refuses the shared-checkout path (`This session is isolated in the
    worktree … Edit the worktree copy of this file instead`). The lead applies this exact text as a
    new bullet right after the `Kanban…레인이 스스로 병합 창을 잡지 않는다.` line: `- **self-dispatch
    lane 예외 — 병합 창.** Claude self-dispatch 팩토리 run의 레인은 위 요청을 하지 않는다 — \`moai
    factory complete\`의 통합 절차로 스스로 통합 창을 잡고 자기 카드를 \`develop\`에 병합한다(OD-2).
    Codex 레인은 예외가 아니다 — merge-ready에서 정지한다(REQ-SD-025). 이 예외도 위의 다른 큐
    변경(\`add\`, \`drop\`, \`done\`, \`edit\` 등) 금지와 \`moai contract sign\` 금지는 바꾸지
    않는다(카드 임대만 \`moai factory next\`로 허용 — OD-1;
    \`.claude/rules/local/gitflow-lane-protocol.md\` §6).` — until then the AC-SD-015 three-file
    grep reads 2/3 in a card worktree (kanban-dispatch 1, gitflow-lane-protocol 2, CLAUDE.local.md
    absent here) and 3/3 only in the primary checkout after the lead applies it.
  - **Verifies (verbatim)**: `cmp .claude/rules/moai/workflow/kanban-dispatch.md
    internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` → exit 0
    (`TWIN_IDENTICAL`); `grep -c 'self-dispatch lane' …` → `.claude/rules/moai/workflow/
    kanban-dispatch.md:1`, `.claude/rules/local/gitflow-lane-protocol.md:2` (CLAUDE.local.md —
    see the blocker above); `go test ./internal/template -run '^TestTemplateNeutralityAudit$'
    -count=1 -timeout 5m -v` → all subtests `--- PASS`, `ok … internal/template 0.475s`.
  - Template-First: `make build` exit 0 after the twin edit (agents-emit-check/commands-emit-check
    pre-steps read-only; catalog.yaml regenerated; binary relinked at `f37942d79`).
- **Final comprehensive anchored run (the run baseline, this tree, env-scrubbed single compound
  `unset …` with the slot lease held)**: `go test ./internal/cli -run
  '^TestSD_AC00[1-9]|^TestSD_AC01[0-9]|^TestSD_AC02[0-5]|^TestSD_WorktreeSlugShape' -count=1
  -timeout 20m` → `ok github.com/modu-ai/moai-adk/internal/cli 58.311s` (AC-SD-001…025 incl. the
  M7 additions and AC-014, plus the t1256 legacy family); `go test ./internal/hook -run
  '^TestSD_AC019_NextCardRuleInjection$|^TestSD_AC017_WidenedRoleGateDenyAndAllow$|^TestContractSignGuard'
  -count=1 -timeout 5m` → `ok 0.623s`; codexwiring freeze → `ok 0.509s`; template neutrality →
  `ok 0.341s`. Slot released after.
- **E2 builds** (final tree): `go build ./...` → `BUILD_OK`; `GOOS=windows GOARCH=amd64 go build
  ./...` → `BUILD_WIN_OK`; `make build` exit 0.
- **E3 coverage**: M7 adds tests only plus two small production functions —
  `factoryLaneLabelFromEnv` 90.9%, `codexFactoryLegacyRefusalDiag` 87.5% (`-coverprofile` over the
  M7 AC selectors, profile at `.moai/state/verify/t1240-m7/cli-cover.out`); both ≥ 85%. The M7 test
  code itself is n/a (tests). `codexFactoryEntryClassify` (pre-existing, M5) reads 59.1% under the
  M7-only selector — its full coverage lives with the M5 AC selectors (AC-SD-004), green in the same
  final run.
- **E4 boundary**: `grep -n 'AskUserQuestion'` over the touched production files
  (factory_card.go, codex_launcher.go, configtoml.go + the two new test files) → no matches
  (exit 1).
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/cli/... ./internal/hook/...
  ./internal/codexwiring/...` → `0 issues.` — no NEW vs baseline (CI version v2.1.6), re-run after
  the final edits.
- **E6 commit**: this section and the M7 change land in one commit
  `feat(SPEC-FACTORY-SELF-DISPATCH-001): M7 invariants, vocabulary, doctrine amendments (card
  t1240)`; push = none (the lead pushes develop in batch).

### Integration absorption (2026-09-29, worker-70)

- Absorbed local develop `145c3d98c` → commit `d43e50bb3` (9 conflicts resolved, both sides
  survive: CHANGELOG combined; gitflow-lane-protocol §6 exceptions + leader vocabulary;
  kanban-dispatch + twin exception paragraph + t1321 diet; moai-mcp-tools + twin 20-tool combined
  surface; factory_card.go/todo.go/codex_launcher.go semantic coexistence), repair `844fa42a8`
  (t1308 positive-enumeration guard integration).
- Test alignment `d7a78aeaa` — D2 (a): two F1-era pick-mirror tests re-homed to the neutral
  operator surface (assertions intact, REQ-SD-015 note added); D3 (a): `TestFactoryRoleEnvConstant`
  closure-set pin updated to the contractLaneGate trio (superseding source partially_superseded_by
  9866ca25e noted); D1 (a) per lead adjudication (Jev choice 422 / noul re-query 0.46 below
  threshold → doctrine-fallback ruling (a)): superseded codex factory-entry test removed with
  supersession comment; foreign commit `6f14e64b2` carries a wrong t1240 label (t1294-lineage,
  pushed, unamendable — recorded batch-end; this branch's records are the true t1240 carriers).
- Gates at `d7a78aeaa` (final): `make build` exit 0 churn 0; both builds exit 0; `golangci-lint`
  `0 issues.`; AC anchored sweep `ok … internal/cli 70.715s` (lane final run) and `ok … 66.745s`
  (post-absorption); families `^TestTodo|^TestFR_` `ok 442.536s`; internal/config `ok 4.193s`;
  codex retire family `ok 26.849s`; cmp twins identical; self-dispatch grep 1/1/2; boundary
  grep 0. Residual red: 0.

## §F Phase 4 Mode Selection

- tier: L · scope: >10 production files across cli/hook/config/kanban/homestate/codexwiring + template rules · domains: 6 (Go CLI, hooks, MCP server, launchers, doctrine rules, env constants) · language mix: Go + Markdown · concurrency benefit: LOW (coding-heavy, sequential milestone chain with shared files)
- direct: not selected — multi-milestone production change. fanout: not selected — coding-heavy (Anthropic coding-task parallelism caveat), milestones share files (factory.go touched by M1/M4/M5). sweep: not selected — semantic multi-rule work, not mechanical uniform transform. agent-team: not selected — no operator request.
- **Decision: serial** (sequential manager-develop delegations per milestone, M1→M7).
- Justification: Tier L coding-heavy chain; plan.md §D.2 file table shows M1/M4/M5 share `factory.go` and M1/M6 share hook files, so parallel writers would race one working tree. Kickoff approved via leader dispatch (operator 09-28 전력 완수 지시 + CLAUDE.local.md §31); skip-eligible plan-audit (PASS-WITH-DEBT 0.90 final iter4, artifacts unchanged since v0.5.1).

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-09-29
- run_commit_sha: aaf4fd1e2
- run_tree: `.claude/worktrees/t1240`, branch `WT-factory-self-dispatch`, baseline measured at `f37942d79`
- milestone_commits: M1 `490d64649` · M2 `986e1a8fa` · M3 `ac7b1b87a` · AC-SD-011 amendment `061614bd5` · M4 `4d73ee088` · M5 `48af8d3ae` · label unification `7714fcd9e` · gate disposition `003514807` · §D carrier amendment `3eeaabe5b` · AC-SD-001 rewrite `f37942d79` · M7 `aaf4fd1e2`
- scope_doc_amendments: `061614bd5` (acceptance.md worktree-root literals → `.moai/worktrees/`) · `3eeaabe5b` (REQ-SD-020 §D carrier carve-out) · `f37942d79` (AC-SD-001 graph-verifiable rewrite) · `917005dfb` (CLAUDE.local.md → AGENTS.local.md migration + merge-window exception applied) · `252d7d32e` (plan.md sweep + HISTORY row token restored)
- ac_pass_count: 25 (AC-SD-001…025 — final comprehensive anchored run, `ok … internal/cli 58.311s` + hook `ok 0.623s` + codexwiring `ok 0.509s`; AC-SD-015's doctrine third block 3/3 files verified in-branch (kanban-dispatch 1 / AGENTS.local.md 1 / gitflow-lane-protocol 2 — the CLAUDE.local.md deliverable completed by `917005dfb` + `252d7d32e` after the t1290 migration removed that file from the tree))
- ac_fail_count: 0
- preserve_list_post_run_count: 0 violations — `internal/codexwiring/configtoml.go` production value untouched (AC-SD-022 freeze test green); schema statements of internal/homestate, internal/factorymsg, internal/kanban untouched (AC-SD-022 green); `internal/cli/integration.go` untouched (not in the diff)
- l44_pre_commit_fetch: n/a (worktree lane — no push, per the lane protocol)
- l44_post_push_fetch: n/a (no push; the lead pushes develop in batch)
- new_warnings_or_lints_introduced: 0 (`golangci-lint run --timeout=2m ./internal/cli/... ./internal/hook/... ./internal/codexwiring/...` → `0 issues.`, CI version v2.1.6)
- cross_platform_build.linux: exit 0 · cross_platform_build.windows: exit 0 (GOOS=windows GOARCH=amd64)
- total_run_phase_files: M7 changes 8 files — Go production: factory_card.go, codex_launcher.go; Go tests: factory_m7_test.go (cli), factory_freeze_test.go (codexwiring); doctrine: kanban-dispatch.md + its template twin, gitflow-lane-protocol.md; record: progress.md. The M1–M7 cumulative file list lives in each commit's diff
- m1_to_m7_commit_strategy: one commit per milestone, each carrying its progress.md section atomically (M1 `490d64649` … M7 this commit); scope-doc amendments and the gate-disposition commit landed interleaved as manager-spec re-delegations and the lead's disposition
- run_gate: AC-SD-001 amended form verified — `git merge-base --is-ancestor abb815921 490d64649` → OK; `develop_sha.*a7190891d` → 1; `t1256_landed.*: yes` → 1

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: f2c44360a
sync_status: complete
b12_self_test_a: "grep -c SPEC-FACTORY-SELF-DISPATCH-001 CHANGELOG.md -> 0 before emission (re-run by this sync; post-insertion the entry is the only matching line)"
b12_self_test_b: "distinct AC ids in acceptance.md -> 25 (AC-SD-001..025; zero [RETIRED]/[REF] markers — every identifier live); CHANGELOG entry cites 25"
b12_self_test_c: "every path cited in the entry verified with ls — internal/cli/{factory_card,codex_launcher,mcp_factory_card,mcp_todo,todo}.go, internal/hook/{session_start_factory,session_start_factory_i18n,pre_tool}.go, internal/config/envkeys.go"
changelog_entry_position: "[Unreleased] ### Added, first entry"
verification_baseline:
  anchored_sweep: "go test ./internal/cli -run '^TestSD_AC00|^TestSD_AC01|^TestSD_AC02|^TestSD_WorktreeSlugShape' -count=1 -> ok github.com/modu-ai/moai-adk/internal/cli 55.976s (pre-sync, slot-leased) · re-run after the MX comment edits -> ok … 58.230s, exit 0"
  build: "go build ./... exit 0 · GOOS=windows GOARCH=amd64 go build ./... exit 0"
  lint: "golangci-lint run --timeout=2m ./internal/cli/... ./internal/hook/... ./internal/codexwiring/... -> 0 issues. (v2.1.6, the CI-pinned version)"
  spec_lint: |
    moai spec lint --strict reports exactly one finding for this SPEC (full-scan grep; no non-Info row names this SPEC):
    INFO OwnershipTransitionUnmeasured — SPEC SPEC-FACTORY-SELF-DISPATCH-001 transition "draft" → "in-progress" expected owner "manager-develop" but commit 490d64649 (M1, feat(SPEC-FACTORY-SELF-DISPATCH-001): M1 lane predicate, selection, permission boundary) has no Authored-By-Agent trailer — ownership transition unmeasured.
    LEAD DISPOSITION (Option A — accept and record, lane decision 2026-09-29): the finding is Info-severity measurement-state by design, not a violation (Info never changes the lint exit status per the frontmatter-schema ownership rule); the transition's substance is correct — manager-develop authored M1 per the run-phase delegation, matching the ownership matrix; the trailer cannot be added without rewriting 490d64649 and every descendant SHA, breaking §E.2/§E.3/CHANGELOG/superseded_by SHA citations — the prohibited direction. acceptance.md §F's "reports no finding" letter is unmet by this one Info finding; the severity-qualified interpretation is recorded here for the sync-auditor.
ac_pass_count: 25
mx_rotation: "4 @MX:ANCHOR added on the fan_in>=3 surfaces the run phase left untagged — factoryLaneAdmission (7 non-test callers), factoryLaneRefusal (4), factoryEnsureCardWorktree (4), mcpRequiredProjectRoot (3); counts measured with grep over non-test sources. No @MX:WARN added (no goroutines/channels in the new files; complexity unmeasured — blanket adds declined). Existing run-phase tags verified in place (codex_launcher.go, envkeys.go)."
codemap_rotation: "no codemap regeneration — the workflow drift check (doc-execution.md D5) gates on new directories / dependency-graph change / module reorganization, none produced by this card; the t1239-era factory verb enumeration in codemaps (assign|status|decide) is stale and is carried to card t1257's documentation layer and the next codemaps refresh, not hand-patched in a generated file"
scope_decision: "conscious scope decision per the sync dispatch — acceptance.md §F DoD = CHANGELOG + clean gates + the partially_superseded_by records; docs-site pages and README text for the new verbs are card t1257's scope (spec.md §E.2 hand-off list), skipped deliberately"
frontmatter_status_transitions:
  spec.md: in-progress -> implemented -> completed (merged into this single sync commit; updated: 2026-09-29 already current)
sync_audit: "Tier L sync-audit follows this commit (lane task #8); the lead reports the Option A disposition alongside the audit verdict"
```

