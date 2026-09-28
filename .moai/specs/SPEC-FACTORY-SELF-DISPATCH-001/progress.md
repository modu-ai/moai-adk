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

## §F Phase 4 Mode Selection

- tier: L · scope: >10 production files across cli/hook/config/kanban/homestate/codexwiring + template rules · domains: 6 (Go CLI, hooks, MCP server, launchers, doctrine rules, env constants) · language mix: Go + Markdown · concurrency benefit: LOW (coding-heavy, sequential milestone chain with shared files)
- direct: not selected — multi-milestone production change. fanout: not selected — coding-heavy (Anthropic coding-task parallelism caveat), milestones share files (factory.go touched by M1/M4/M5). sweep: not selected — semantic multi-rule work, not mechanical uniform transform. agent-team: not selected — no operator request.
- **Decision: serial** (sequential manager-develop delegations per milestone, M1→M7).
- Justification: Tier L coding-heavy chain; plan.md §D.2 file table shows M1/M4/M5 share `factory.go` and M1/M6 share hook files, so parallel writers would race one working tree. Kickoff approved via leader dispatch (operator 09-28 전력 완수 지시 + CLAUDE.local.md §31); skip-eligible plan-audit (PASS-WITH-DEBT 0.90 final iter4, artifacts unchanged since v0.5.1).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
