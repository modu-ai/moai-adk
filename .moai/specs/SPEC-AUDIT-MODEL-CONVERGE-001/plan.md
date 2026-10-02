# SPEC-AUDIT-MODEL-CONVERGE-001 — plan

> Tier L plan artifact (0.1.3). Requirements: spec.md §C. Reasoning and shapes:
> design.md. Measurements: research.md §R.1, §R.5, §R.6. Acceptance:
> acceptance.md. This file has no frontmatter `status:`.

## §A Context

- **Card** t1423, Class C (design change). Worktree `.moai/worktrees/t1423`,
  branch `WT-audit-model-convergence`; authored at `c50da9c2f`, amended at
  `53a42f013` and `739556db5` (SPEC commits; the code is identical,
  research.md R-28).
- **Methodology**: `development_mode: tdd` (research.md R-18) — RED first on every
  behaviour milestone, with the verbatim failing output captured before GREEN
  (manager-develop-prompt-template §E8).
- **Tier L** — about 36 files, 7 milestones (spec.md §E, re-derived at 0.1.3). The
  orchestrator's `manager-lead` entry predicate (>= 3 milestones AND >= 10 files)
  is met; the run-phase mode is the orchestrator's decision, logged in
  `progress.md §F`. Default if not routed to `manager-lead`: `serial`, one
  `manager-develop` per milestone (coding-heavy work).
- **Self-application.** Plan-audit and sync-audit of this SPEC should themselves
  call `audit_multi` with `gates: {claude: required, codex: required, glm:
  advisory}` and `project_root` set to this worktree's toplevel, and cite the
  receipts. The config toggle does not exist yet; the call argument does.
- **Scope reduction at 0.1.3 (decision D13).** Removed: REQ-ACV-018/019 and their
  criteria (regression guards the existing tests already hold), the `IsBinding`
  path, the verb's session-store mode, the compound sync requirements, the
  per-test hermetic helper, the `sync-audit-4dim.js` text edits, the separate
  receipt test file. Added: the codex turn reader fix (`mcp_codex.go`). Net: 24 ->
  20 requirements, 22 -> 20 criteria, 38 -> 36 files.

## §B Decisions and open items

### B.1 Decisions taken here

- **D-1 Surface = a CLI verb in the existing `moai verify` group**, spelled
  `moai verify audit-plan` (leader ruling D10; design.md §D.4).
- **D-2 Symbols the acceptance criteria name.** `config.ResolveAuditPlan`,
  `config.AuditPlan`, method `AuditPlan.ExplicitGates()`, file
  `internal/config/audit_plan.go`; the verb in `internal/cli/audit_plan_cmd.go`
  (constructor `newVerifyAuditPlanCmd`, registered through `verifyExtraCommands`);
  the codex-leg limit `config.DefaultCodexAuditLegTimeout` in
  `internal/config/defaults.go`. A rename is a cheap edit in acceptance.md and
  nothing else.
- **D-3 `plan_source`** is the one new `ConvergenceResult` member (`omitempty`);
  its consumer is the verb's checker (REQ-ACV-016).
- **D-4 Test hermeticity.** One `os.Unsetenv(config.EnvClaudeProjectDir)` in
  `internal/cli/main_test.go`'s `TestMain` plus `TestMain_ScrubsClaudeProjectDir`
  (design.md §D.14); no per-test helper.
- **D-5 Order.** M1 first (baseline-first, §2.3); then the decision-bearing
  milestones by how likely they are to change (resolver interface, consumers and
  the codex turn, the verb's shape); the mechanical cleanup and the config flip;
  and the activation text — the local auditors, the skill and `sync.md` — LAST, so
  nothing instructs the new path before every mechanism it relies on is in the
  tree (design.md §D.13).
- **D-6 The checker is pure (decision D12).** `--result '<json>'` is the only input;
  there is no session id and no read of the `audit-multi` store. The argument form
  is chosen because the worktree guard refuses heredocs and compound commands and
  `sync-auditor` has no `Write` tool (design.md §D.5).
- **D-7 The codex turn reader is fixed at the source (PA2-D6).** The two
  non-completed returns of `awaitCodexTurnReview` become errors; `runTurn` already
  maps an error to `inconclusive`. The leg adds only a deadline and the timeout
  wording (design.md §D.10).

### B.2 Open items

- **OQ-1 — closed (decision D7', CONFIRMED by the factory leader).** An explicit
  `audit.model: claude` is the Claude gate `required` and explicit, with codex and
  glm left at the non-explicit default; it keeps `audit_multi` behaving as today and
  `cross_model_active` false. Jev confidence 0.22, below the 0.5 gate; the leader's
  ruling confirmed it. Evidence kept for the plan-audit: `defaults.go:1272-1304`
  pairs `Model: claude` with the default gates; `audit_models.go` calls `claude`
  both "the default single-backend audit model" and part of the default profile;
  wizard persistence of an untouched `claude` is unobserved. Error direction: if
  "Claude alone" were adopted and wrong, a project that persisted `claude` would
  silently lose codex and GLM in `audit_multi`; if "Claude explicit, rest default" is
  adopted and an operator wanted Claude alone, the cost is one visible edit.
- **OQ-2 — closed (decision D5; value and home per plan-audit PA1-D4).** The codex
  leg gets a deadline held in `config.DefaultCodexAuditLegTimeout`, derived from
  `DefaultCodexAuditTimeout` (20 minutes, design.md §D.10), shortenable in tests
  like its siblings (REQ-ACV-020, AC-ACV-020, M3).
- **OQ-3** Assumption, flagged: a tool-argument `required` is not explicit
  (design.md §D.2). Correct it before Kickoff if the operator wants arguments to
  fail closed too; the change is one source label and one test.
- **OQ-4 — closed (plan-audit PA1-D2).** An unreadable `workflow.yaml` is a
  distinct non-passing state of the verb and blocks PASS for the auditors and the
  sync step (REQ-ACV-012, -015). `audit_multi` called outside the auditors keeps
  today's reading (spec.md R-4).
- **OQ-5 — closed (leader ruling D10).** The verb is `moai verify audit-plan`.
- **OQ-6** Not observed: whether a Codex-hosted read-only auditor role can execute
  `moai verify audit-plan`. If it cannot, it takes the blocking-Gap branch of
  REQ-ACV-015 (a refused run is not the old-binary signature), not the legacy path.
- **OQ-7 — closed (decision D6).** The 4-dimension verdict stays the Claude input;
  the orchestrator additionally calls `audit_multi` after a clean PASS when codex is
  required; binding only when both pass (REQ-ACV-017). Decision D9 keeps this in
  prose plus the machine-readable result.
- **OQ-8** Assumption, flagged: 20 minutes (`DefaultCodexAuditTimeout`) is the right
  limit for the codex leg. It is derived from the repository's own bound for the same
  kind of audit, not measured (research.md §R.4). Trade-off: a hung codex delays an
  audit by up to 20 minutes before failing closed, by name; a shorter value would
  turn a slow legitimate review into an unmet gate for the whole pipeline (spec.md
  R-2). Measure a few real codex audit durations before shortening it.
- **OQ-9** Assumption, flagged: where the orchestrator's statement of the binding
  sync verdict is persisted for the sync-audit to re-read is not specified in
  `sync.md` (research.md R-24). The cross-model lines are added to that statement
  (REQ-ACV-017) without pinning a file.
- **OQ-10 — closed (decision D12).** The result check is a pure checker fed the
  result JSON the caller holds; the earlier session-store mode is gone.
- **OQ-11** Assumption, flagged (new): making a closed stream and an ended context
  `inconclusive` in `awaitCodexTurnReview` is safe for the other callers of
  `runTurn` — the codex review gate, `codex_audit` and `codex_task` (the last races
  its own context against the turn, so the context/timer race is unaffected, but its
  stream-closed arm changes from `completed` with partial output to `failed` with
  the cause, `codex_task.go:497-504`). It
  is checked by the existing codex test families staying green (AC-ACV-020), not
  proved; a test that pinned the old partial-text behaviour would have to change.
- **OQ-12** Assumption, flagged (new): `moai doctor` is the right per-session check
  that a running MCP server matches the installed binary. A check named "MCP Server
  Version" exists (research.md R-34); its comparison logic was not exercised against
  an old server in this phase.

## §C Pre-flight (the run inherits)

```bash
git rev-parse --short HEAD        # separate calls — the worktree guard refuses compounds
git branch --show-current
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m    # CI pins v2.1.6; distinguish NEW from baseline
```

Environment scrub — every Go verification below is ONE compound invocation (each
Bash call is a fresh process; a separate `unset` does nothing):

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test …
```

`<SCRUB>` stands for that whole prefix. AC-ACV-019 is the one place the opposite is
wanted: it sets `CLAUDE_PROJECT_DIR` on purpose.

Scope: run only the packages and `-run` names a milestone names; push and let CI
run the full suite. Do not run the `internal/cli` or `internal/hook` full suites
locally (heavy-run resources; the slot lease applies if a wider run is ever
needed).

## §D Constraints

- **PRESERVE (do not edit):** `internal/cli/mcp_glm.go`, `mcp_claude*.go` (backend
  legs); `internal/cli/mcp_codex.go` except the two non-completed returns of
  `awaitCodexTurnReview` (and nothing else in that file);
  `internal/runtime/sync_4dim_binding.go` and its test (decision D9);
  `.claude/workflows/sync-audit-4dim.js` and its template copy (not edited at all);
  `internal/hook/**` (including the Stop-hook wrappers and their 900 s budgets);
  `internal/mcp/catalog.go`; `.claude/skills/moai/workflows/review.md` and its
  template copy (decision D11); the distributed template `workflow.yaml`; the Go
  default `Audit.Model`; every pre-existing local-vs-template difference of
  research.md R-15 (edit the passages this plan names, nothing else in those files).
- **Structural guards the new code must pass:**
  `TestConvergence_NoNewAuditModelEnum`, `TestConvergence_NoDirectFrontmatterRead`
  (no `frontmatter` / `agent_overrides` / `AuditModel* =` in `mcp_convergence.go`),
  `TestAuditMulti_NoHardErrorPath_AC_AMM_024`,
  `TestClaudeAuditTemplateSurfacesAndCatalogHash`, the subagent-boundary rule (no
  `AskUserQuestion` / `mcp__askuser__` in any new CLI file — add the
  `TestNew_NoAskUserQuestion`-style static guard for the verb).
- **No inline env names or thresholds** — `internal/config/envkeys.go` constants
  only; the one new timeout is the derived variable of REQ-ACV-020
  (`DefaultCodexAuditLegTimeout = DefaultCodexAuditTimeout`), not a literal.
- **Raw reads only.** No new code path may build the resolver's input from
  `NewDefaultWorkflowConfig()` / `NewDefaultConfig()` or any merged loader
  (REQ-ACV-004; AC-ACV-004's grep and AC-ACV-008's pins-only fixture).
- **No store read in the verb.** `audit_plan_cmd.go` must not reference
  `audit-multi`, `loadConvergenceResult` or a session id (AC-ACV-015's grep).
- **Template copies carry no internal identifiers.** Text added to any file under
  `internal/template/templates/` (agents, skill, `sync.md`) must contain no SPEC id,
  card id or incident narrative — `TestTemplateNoInternalContentLeak` fails the
  build otherwise; only the local copies may cite them.
- **Template-First.** Edit `internal/template/templates/…` first, then mirror the
  local copy, then `make agents-emit` (regenerates the two template `.codex` role
  TOMLs), then `make build` (agents-emit-check → … → `gen-catalog-hashes --all` →
  build; refreshes the skill's catalogue hash). Local-vs-template parity is required
  only where a test demands it: the `## MCP Audit Tools (cross-model second
  opinion)` section of each auditor, and the skill directory hash. Installing a
  release-candidate binary is the leader's step (plan.md §J) and is not part of any
  milestone.
- **Language:** code comments English (`code_comments: en`); instruction documents
  English; commit subjects English with the card id `t1423`.
- **Cross-platform:** the new CLI file uses `filepath` only; verify with the Windows
  build above.
- **Never** `git add -A` / `git add .` (shared checkout rule) — stage by pathspec.

## §E Milestones

Ordering rule (D-5): M1 first by the commit-graph rule; M2-M4 carry the decisions
likeliest to change; M5-M6 are mechanical and activate the committed config; M7 —
the activation text — is the card's last commit. Each milestone is one or more
commits whose subject names the SPEC and card.

| Milestone | Files (source · tests) | Flips |
|---|---|---|
| M1 baseline-first + scrub | `internal/cli/main_test.go` (scrub, `TestMain_ScrubsClaudeProjectDir`), `internal/cli/mcp_audit_multi_baseline_test.go`, `internal/cli/testdata/audit_multi_default.golden.json` | AC-008 (GREEN on arrival), AC-019 |
| M2 resolver | `internal/config/audit_plan.go` · `audit_plan_test.go` | AC-001, -002, -003 (resolver part), -004 |
| M3 consumers + codex turn | `mcp_audit_multi.go`, `mcp_convergence.go`, `mcp_codex.go` (two returns), `defaults.go`, `mcp_worktree_root.go`, `auditreceipt/store.go` · `mcp_audit_multi_config_plan_test.go`, `codex_turn_incomplete_test.go`, `store_test.go` | AC-006, -007, -009, -010, -020 |
| M4 verb + checker | `internal/cli/audit_plan_cmd.go` · `audit_plan_cmd_test.go` | AC-005 (complete), -011, -012 (verb part), -015, -003 (verb part) |
| M5 cleanup | `mcp_audit.go`, `audit_models.go`, `closed_sets.go`, `schema_sections.go` · `mcp_audit_test.go`, `internal/web/mcp_audit_surface_test.go` | AC-017 |
| M6 config flip | `.moai/config/sections/workflow.yaml` · `internal/config/` yaml test | AC-018 |
| M7 activation text | 8 documents (local + template), 2 `.codex` TOMLs, `catalog.yaml` · `internal/template/` doc-surface test | AC-013, -014, -016, -012 (document part) |

### M1 — Baseline-first characterization and the `TestMain` scrub (own commit; precedes every behaviour commit)

Why first: AC-ACV-008 asserts "byte-identical with no config and no args", a claim
whose ordering the commit graph alone can witness (verification-claim-integrity
§2.3). The golden must be committed BEFORE the code that could change it.

- New `internal/cli/mcp_audit_multi_baseline_test.go` and
  `internal/cli/testdata/audit_multi_default.golden.json`: drive `handleAuditMulti`
  with stubbed `backendCall` and no `gates`, over **two** temp project roots — one
  with NO `workflow.yaml`, one with the distributed shape (a `workflow.yaml` whose
  `audit` block carries only the three backend pins, copied in shape from the
  template yaml); zero `BuildCommit`/`BuildLag`; compare the marshalled result to the
  golden. Cover an all-pass case and a codex-inconclusive fail-open case per root.
  All must pass on the pre-change tree.
- `internal/cli/main_test.go`: add `_ = os.Unsetenv(config.EnvClaudeProjectDir)` in
  `TestMain` beside the existing ambient-env scrubs, and
  `TestMain_ScrubsClaudeProjectDir`. `internal/hook` already scrubs the variable
  (research.md R-31); no other package needs it.
- Commit subject: `test(SPEC-AUDIT-MODEL-CONVERGE-001): M1 baseline golden + TestMain project-dir scrub (card t1423)`.

### M2 — The resolver (the interface most likely to change)

- New `internal/config/audit_plan.go` (+ `audit_plan_test.go`): the pure resolver per
  design.md §D.1-D.3 (the `claude` row of decision D7'), error text per REQ-ACV-003,
  `ExplicitGates()`, source labels. Table tests over every token × every gate
  source, an invalid token, an invalid gate value, and the precedence ladder. Update
  the `ValidAuditModels` / `AuditModel*` doc comments minimally (the deferral wording
  goes in M5).
- RED first: capture the failing output before the resolver exists.

### M3 — Consumers, the codex turn reader, and the codex-leg deadline

- `internal/cli/mcp_audit_multi.go`: supplied-only gate reader; resolver call fed by
  `workflowAuditPins`; tool error on invalid configuration; `PlanSource` threading.
- `internal/cli/mcp_convergence.go`: add `PlanSource` (`omitempty`) to
  `ConvergenceResult` and `MultiAuditConfig`; wrap `performCodexAudit`'s context in
  `context.WithTimeout(…, config.DefaultCodexAuditLegTimeout)` and, after the RPC
  returns, replace the summary with the timeout wording when the leg's own deadline
  ended it and the caller's did not (design.md §D.10 item 5). Mind the two string
  guards of §D.
- `internal/cli/mcp_codex.go`: in `awaitCodexTurnReview` make the two non-completed
  returns (`ctx.Err()` at 1264, stream closed at 1268) return an error naming the
  cause; leave the completed arm (1319) and every other line alone.
- `internal/config/defaults.go`: add `var DefaultCodexAuditLegTimeout =
  DefaultCodexAuditTimeout` beside its siblings with the same "distinct, not a
  const" comment.
- `internal/cli/mcp_worktree_root.go`: `resolveAuditGates` returns the plan's
  explicit gates (signature unchanged; the orphaned-worktree rules untouched).
  `internal/cli/audit_pin.go` only if a raw-config accessor is needed.
- `internal/auditreceipt/store.go`: `rawCodexGate` reads `audit.model` too and asks
  the resolver; keeps the fail-open reading for an invalid value.
- Tests: new `internal/cli/mcp_audit_multi_config_plan_test.go` (fallback per token
  including the `claude` row, args-win, fail-closed named, the receipt tests
  `TestAuditMulti_ModelMultiRecordsReceiptOnUnmetGate` and
  `TestCodexAudit_ModelCodexUnmetGateFails`, and the two `audit_multi` deadline
  tests with a fake `codexSession` whose `start(ctx)` closes its stream on `ctx.Done`
  and `config.DefaultCodexAuditLegTimeout` shortened and restored); new
  `internal/cli/codex_turn_incomplete_test.go` (the three `TestCodexTurnReview_*`
  tests over `runCodexTurnWithLines`); extend `internal/auditreceipt/store_test.go`
  (`TestCodexGateRequired_FromModelToken`).
- Run the existing codex families (AC-ACV-020's last command) before moving on; a
  test that pinned the partial-text return is the signal (OQ-11).

### M4 — The read-only verb and its pure checker

- New `internal/cli/audit_plan_cmd.go`: `moai verify audit-plan` registered through
  `verifyExtraCommands`, sharing `verify`'s `--project-root` (resolved through
  `verifyResolveRoot`), plus `--result '<json>'`; JSON per design.md §D.5, including
  the `unreadable` state (read through `loadWorkflowAuditSection`, with the
  config-orphaned-worktree routing of `resolveAuditGates`) and the
  `convergence_check` object; `audit-plan:`-prefixed errors, exit codes 0 / 1 / 2.
  No store read, no session id. New `audit_plan_cmd_test.go`: plan printing for every
  token (an explicit `model: claude` fixture asserts the D7' row), the pins-only and
  absent shapes, invalid configuration, the corrupt-yaml fixture, project-root
  handling, the output contract, the checker fixtures (a)-(h), the writes-no-files
  test (a recursive listing of regular files under the tree's `.moai`, before and
  after; the `backendCall` seam never invoked) and the `AskUserQuestion` static
  guard.

### M5 — Stale-deferral cleanup (mechanical)

- `internal/cli/mcp_audit.go`: delete `multiConvergenceImplemented` and
  `activeAuditBackend`, fix the header comment; keep `buildAuditEnvBlock`.
  `internal/cli/mcp_audit_test.go`: delete the three `TestActiveAuditBackend_*`
  tests. `internal/config/audit_models.go`, `closed_sets.go`,
  `internal/settings/schema_sections.go`: comment-only edits.
  `internal/web/mcp_audit_surface_test.go`: sentinel → `ResolveAuditPlan` + a
  positive control that the symbol exists in non-test source of `internal/config`.

### M6 — Config surface (it activates the behaviour in this repository)

- `.moai/config/sections/workflow.yaml`: `model: multi`; claude pin `effort: high`;
  header comment corrected. New test (in `internal/config/`) loading the committed
  file and asserting `Audit.Model == "multi"` and `Audit.Claude.Effort ==
  DefaultClaudeAuditEffort`. `TestMain_ScrubsClaudeProjectDir` (M1) is what keeps the
  `internal/cli` tests off this file; re-run it with the variable set (AC-ACV-019).

### M7 — Activation text (the card's LAST commit)

- Template sources first, then local mirrors, in one commit: `plan-auditor.md`,
  `sync-auditor.md` (the `## MCP Audit Tools (cross-model second opinion)` section
  identical on both sides; **the `<!-- moai:closure-second-review:start -->` …
  `:end -->` markers in `sync-auditor.md` are preserved** when the token-keyed block
  is removed), `skills/moai-ref-cross-model-audit/SKILL.md` (plan-driven table; both
  copies identical), `skills/moai/workflows/sync.md` (the `Binding promotion`
  paragraph and a pointer sentence in `FO-SYNC-1`: the added post-PASS `audit_multi`
  call, the checker call, the "both must pass" rule, the statement lines, the
  unmet/unreachable record). The old token-keyed block becomes the labelled
  **legacy path** paragraph; the main path is the verb. Required literals:
  `moai verify audit-plan`, `--result`, `plan surface unreachable, legacy path used`,
  `Shared diagnostic snapshot contract` and `config_status: unreadable` in the two
  agents, the skill and `sync.md`; `cross_model_required`, `audit-plan --result` and
  `audit_multi unreachable` in `sync.md`; the old "Single-backend audit mode (per the
  project's `audit_model`)" block heading gone.
- A doc-surface test in `internal/template/` (new file, `TestAuditPlanDocSurface`)
  asserting the literals in the eight document copies and the absence of the old
  heading.
- `make agents-emit`, then `make build` (catalogue hash; all emit/drift checks).
- `.claude/workflows/sync-audit-4dim.js` is NOT edited (spec.md R-11).

## §F Verification commands

All Go commands carry `<SCRUB>` except AC-ACV-019's. Per-milestone, scoped:

```bash
<SCRUB> go test -count=1 ./internal/config/ ./internal/auditreceipt/
<SCRUB> go test -count=1 -run '^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestConverge_.*|TestCodexAudit.*|TestCodexTurnReview_.*|TestAuditPlanCmd_.*|TestMain_.*)$' ./internal/cli/
<SCRUB> go test -count=1 ./internal/template/ ./internal/web/
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m
```

Read each run's swept count before its verdict: a `-run` pattern that selects
nothing prints `ok` (verification-completeness §1.1). Quote the `--- PASS:` lines of
the named tests.

## §G Anti-patterns (named for the run to refuse)

- **AP-1 Resolver in the wrong package** — putting it in `mcp_convergence.go` trips
  two structural guards; putting it in `internal/cli` makes `internal/auditreceipt`
  unable to reach it.
- **AP-2 Token without enforcement** — making `audit_multi` follow `multi` while
  leaving the three raw-gate readers alone reproduces fail-open (research.md §R.3).
- **AP-3 A new MCP tool or a top-level `moai audit` group** — rejected in design.md
  §D.4; the verb is `moai verify audit-plan`.
- **AP-4 Silent default on a bad token** — the resolver's error must reach the
  caller (REQ-ACV-003).
- **AP-5 Fixing pre-existing local/template drift** — research.md R-15; not this
  SPEC's change.
- **AP-6 Sweep-staging** — `git add -A` is forbidden in the shared checkout.
- **AP-7 Moving the sync call into the script, or editing it** — `audit_multi` is
  write-capable and the script's agents are read-only with a pure-JS verdict; the
  call belongs to the orchestrator in `sync.md` (design.md §D.8).
- **AP-8 Demoting the 4-dimension verdict, or touching `IsBinding`** — superseded by
  decisions D6 and D9.
- **AP-9 A silent codex turn** — a turn that ends by deadline, cancellation or a
  closed stream must be `inconclusive` naming the cause, never a verdict parsed from
  partial output.
- **AP-10 A default-merged read** — feeding the resolver from
  `NewDefaultWorkflowConfig()` turns every unconfigured project into `claude`.
- **AP-11 Treating the unknown verb as an error, or "no output" as the old binary** —
  `moai verify <unknown>` prints help and exits 0 on an older binary; the legacy path
  needs that positive signature, and every other failure to run is a blocking Gap.
- **AP-12 A session-store read in the checker** — the pure checker takes the result
  it is handed; reading `.moai/state/audit-multi/` brings back the wrong-reader,
  stale-file and path-escape problems (design.md §D.5).
- **AP-13 A per-test hermetic helper** — one `TestMain` scrub is the mechanism
  (design.md §D.14).

## §H Cross-references

`spec.md` · `design.md` · `research.md` · `acceptance.md` ·
`.claude/rules/moai/core/verification-claim-integrity.md` §1-§3.1 ·
`.claude/rules/moai/development/verification-completeness.md` §1-§3 ·
`.claude/rules/moai/core/moai-mcp-tools.md` · SPEC-AUDIT-MULTI-MODEL-001 ·
SPEC-CODEX-AUDIT-GATE-AXES-001 · SPEC-AUDIT-SNAPSHOT-001 ·
`.moai/reports/t1423/progress.md` (operator decisions D1-D13) ·
`.moai/reports/t1423/plan-audit-iter1.md` and `plan-audit-iter2.md` (the PA1 and PA2
defects this version answers).

## §I Gaps (unobserved in this plan phase)

- No Go test of the SPEC's own behaviour was run (this phase edits no Go); every
  GREEN claim is a prediction until M1-M7 run. Two throwaway behavioural tests were
  run through `go test -overlay` for the RED of AC-ACV-009 and AC-ACV-020
  (acceptance.md E22, E29), with the test sources outside the tree.
- Live codex / GLM behaviour; wizard persistence of `claude`; Codex-hosted execution
  of the verb; `moai update` preserving `model: multi` (research.md §R.4).
- Real codex leg durations (OQ-8) and the persistence location of the sync binding
  statement (OQ-9) are unobserved. The `moai doctor` MCP-server comparison was not
  exercised against an old server (OQ-12).
- Whether the existing codex test families stay green after the reader fix (OQ-11)
  is checked in M3, not here.
- Whether an `Explore` workflow agent could carry MCP tools was not observed; the
  design does not depend on it (the call stays with the orchestrator).
- The `moai verify` group's behaviour for an unknown verb was measured on the
  installed binary only; a future cobra or group change would alter the signature of
  REQ-ACV-014, and the outcome would then fall into the blocking-Gap branch.
- A `pass` from partial codex text was not reproduced (E29): only a partial `fail`
  was; the reader fix covers both.

## §J Rollout, activation, and the Definition-of-Done install item

The verb's source does not exist on develop until this card merges. The safety comes
from behaviour first — the legacy path of REQ-ACV-014 while the verb is absent —
with ordering as a second layer: the activation text is the card's last commit (M7).

Sequence after the card's commits are done:

1. The leader merges the card into develop (the lane protocol's integration window).
2. A release-candidate build is made from develop and installed
   (`.claude/rules/local/gitflow-lane-protocol.md` §9; operator-requested).
3. **Every live session that will run a plan-audit or sync-audit — the leader's and
   each lane's — reconnects its MCP server or restarts.** After the install the CLI
   is resolved per call, but each session keeps the MCP server it started with.
   `moai doctor` ("MCP Server Version") is the documented check that the running
   server matches the installed binary.
4. Smoke, on the installed build, from the tree that will run the audit:
   `moai verify audit-plan --project-root <abs toplevel>` prints a JSON plan with
   `config_status: "ok"`, `model: "multi"` in this repository; one `audit_multi`
   call on a trivial target returns a result carrying `plan_source: "config"`; and
   `moai verify audit-plan --project-root <abs toplevel> --result '<digest>'` of that
   result prints `convergence_check.ok: true`.
5. Only then does the new audit path apply to that session's later audits.

What is true meanwhile, per session: while the verb is absent the auditors and the
sync step run the legacy path with the named Gap, which blocks nothing. After the
install, a session whose MCP server has not been reconnected is **not** unaffected:
its `audit_multi` results lack `plan_source`, the checker reports an unmet gate by
name, and that session's audits fail closed until it reconnects — a visible, bounded,
per-session window. The leader's hand-off note after the install is therefore: list
the live sessions, reconnect or restart each, run `moai doctor` in each, and do not
start a plan-audit or sync-audit in a session that has not. The Definition of Done
(acceptance.md) carries this sequence as its install-and-verify item.
