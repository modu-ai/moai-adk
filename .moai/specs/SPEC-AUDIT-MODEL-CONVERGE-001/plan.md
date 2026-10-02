# SPEC-AUDIT-MODEL-CONVERGE-001 — plan

> Tier L plan artifact (0.1.2). Requirements: spec.md §C. Reasoning and shapes:
> design.md. Measurements: research.md §R.1 and §R.5. Acceptance: acceptance.md.
> This file has no frontmatter `status:`.

## §A Context

- **Card** t1423, Class C (design change). Worktree `.moai/worktrees/t1423`,
  branch `WT-audit-model-convergence`; authored at `c50da9c2f`, amended at
  `53a42f013` (the SPEC commit; the code is identical, research.md R-28).
- **Methodology**: `development_mode: tdd` (research.md R-18) — RED first on
  every behaviour milestone, with the verbatim failing output captured before
  GREEN (manager-develop-prompt-template §E8).
- **Tier L** — about 38 files, 7 milestones (spec.md §E, re-derived at 0.1.2).
  The orchestrator's `manager-lead` entry predicate (>= 3 milestones AND >= 10
  files) is met; the run-phase mode is the orchestrator's decision and is logged
  in `progress.md §F`. Default if not routed to `manager-lead`: `serial`, one
  `manager-develop` per milestone (coding-heavy work).
- **Self-application.** Plan-audit and sync-audit of this SPEC should themselves
  call `audit_multi` with `gates: {claude: required, codex: required, glm:
  advisory}` and `project_root` set to this worktree's toplevel, and cite the
  receipts. The config toggle does not exist yet; the call argument does.

## §B Decisions and open items

### B.1 Decisions taken here

- **D-1 Surface = a CLI verb in the existing `moai verify` group**, spelled
  `moai verify audit-plan` (leader ruling D10; design.md §D.4 gives the
  spelling reasoning: it names a plan, follows `sync-gate` / `codex-review`
  hyphenation, and does not imply that it runs an audit; `moai spec audit`
  already owns the word "audit" for SPEC-era classification).
- **D-2 Symbols the acceptance criteria name.** `config.ResolveAuditPlan`,
  `config.AuditPlan`, method `AuditPlan.ExplicitGates()`, file
  `internal/config/audit_plan.go`; the verb in `internal/cli/audit_plan_cmd.go`
  (constructor `newVerifyAuditPlanCmd`, registered through `verifyExtraCommands`);
  the codex-leg limit `config.DefaultCodexAuditLegTimeout` in
  `internal/config/defaults.go`. A rename is a cheap edit in acceptance.md
  §AC-ACV-004/005/011/020 and nothing else.
- **D-3 `plan_source`** is the one new `ConvergenceResult` member (`omitempty`);
  its consumer is the verb's result check (REQ-ACV-022).
- **D-4 Test hermeticity.** Every test that reaches `resolveProjectDir()`
  without passing `project_root` sets `CLAUDE_PROJECT_DIR` to its own
  `t.TempDir()` (M1), so the committed yaml of M6 cannot reach it.
- **D-5 Order.** M1 first (baseline-first, §2.3); then the decision-bearing
  milestones by how likely they are to change (resolver interface, consumers and
  deadline, the verb's shape); the mechanical cleanup and the config flip; and
  the activation text — the local auditors, the skill and `sync.md` — LAST, so
  nothing instructs the new path before every mechanism it relies on is in the
  tree (design.md §D.13).
- **D-6 The result check is mechanical.** The auditors are model-driven; the
  rule "verify the result against the plan" is implemented once in the verb
  (`--check-session`) and the auditors run it, rather than re-deriving the rule
  in prose each time. It reads the result the server already persists.

### B.2 Open items

- **OQ-1 — closed (decision D7', CONFIRMED by the factory leader; no longer
  provisional).** An explicit `audit.model: claude` is the Claude gate `required`
  and explicit, with codex and glm left at the non-explicit default; it keeps
  `audit_multi` behaving as today and `cross_model_active` false. It supersedes
  the earlier D4 reading (Claude alone, confidence 0.57). The Jev confidence was
  0.22, below the 0.5 gate; the leader's ruling confirmed it. Evidence kept for
  the plan-audit: `defaults.go:1272-1304` pairs `Model: claude` with the default
  gates; `audit_models.go` calls `claude` both "the default single-backend audit
  model" and part of the default profile; wizard persistence of an untouched
  `claude` is unobserved. Error direction: if "Claude alone" were adopted and
  wrong, a project that persisted `claude` would silently lose codex and GLM in
  `audit_multi`; if "Claude explicit, rest default" is adopted and an operator
  wanted Claude alone, the cost is one visible edit (`gates.codex: off`,
  `gates.glm: off`). The row takes the safe direction (design.md §D.1).
- **OQ-2 — closed (decision D5; value and home revised by plan-audit PA1-D4).**
  The codex leg of `audit_multi` gets a deadline held in
  `config.DefaultCodexAuditLegTimeout` beside `DefaultCodexAuditTimeout` /
  `DefaultCodexTaskTimeout`, derived from `DefaultCodexAuditTimeout` (20 minutes,
  design.md §D.10), shortenable in tests like its siblings; any end of the leg
  context is `inconclusive` (REQ-ACV-020, AC-ACV-020, M3).
- **OQ-3** Assumption, flagged: a tool-argument `required` is not explicit
  (design.md §D.2). Correct it before Kickoff if the operator wants arguments to
  fail closed too; the change is one source label and one test.
- **OQ-4 — closed (plan-audit PA1-D2).** An unreadable `workflow.yaml` is a
  distinct non-passing state of the verb and blocks PASS for the auditors and
  the sync step (REQ-ACV-023). `audit_multi` called outside the auditors keeps
  today's reading (spec.md R-4).
- **OQ-5 — closed (leader ruling D10).** The verb is `moai verify audit-plan`.
  The original top-level `moai audit` group is not created.
- **OQ-6** Not observed: whether a Codex-hosted read-only auditor role can
  execute `moai verify audit-plan`. If it cannot, the codex roles take the
  legacy path with the named Gap (REQ-ACV-014).
- **OQ-7 — closed (decision D6).** The 4-dimension verdict stays the Claude
  input; the orchestrator additionally calls `audit_multi` after a clean PASS
  when codex is required; binding only when both pass (REQ-ACV-015, -021, -024).
  Decision D9 keeps this in prose plus the machine-readable result and drops the
  Go `IsBinding` change.
- **OQ-8** Assumption, flagged: 20 minutes (`DefaultCodexAuditTimeout`) is the
  right limit for the codex leg. It is derived from the repository's own bound
  for the same kind of audit, not measured; no codex leg duration was observed
  (research.md §R.4). Trade-off: a hung codex delays an audit by up to 20 minutes
  before failing closed, by name; a shorter value would turn a slow legitimate
  review into an unmet gate for the whole pipeline (spec.md R-2). Measure a few
  real codex audit durations before shortening it.
- **OQ-9** Assumption, flagged: where the orchestrator's statement of the
  binding sync verdict is persisted for the sync-audit to re-read is not
  specified in `sync.md` (a grep of `sync.md` and `sync/*.md` found no record
  location, research.md R-24). The cross-model lines are added to that statement
  (REQ-ACV-021) without pinning a file.
- **OQ-10** Assumption, flagged for the plan-audit: the result check is a mode of
  the verb (`--check-session <id>`, reading the persisted result) rather than
  prose-only instructions or a second verb. Reason: prose-literal criteria cannot
  prove a model follows them (plan-audit PA1-D8), and `plan_source` needs a
  production consumer (PA1-D1). It adds a read of one persisted file to a verb
  that otherwise reads only `workflow.yaml`.

## §C Pre-flight (the run inherits)

```bash
git rev-parse --short HEAD        # separate calls — the worktree guard refuses compounds
git branch --show-current
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m    # CI pins v2.1.6; distinguish NEW from baseline
```

Environment scrub — every Go verification below is ONE compound invocation
(each Bash call is a fresh process; a separate `unset` does nothing):

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test …
```

`<SCRUB>` below stands for that whole prefix. AC-ACV-018 is the one place the
opposite is wanted: it sets `CLAUDE_PROJECT_DIR` on purpose.

Scope: run only the packages and `-run` names a milestone names; push and let CI
run the full suite. Do not run the `internal/cli` or `internal/hook` full suites
locally (heavy-run resources; the slot lease applies if a wider run is ever
needed).

## §D Constraints

- **PRESERVE (do not edit):** `internal/cli/mcp_codex.go`, `mcp_glm.go`,
  `mcp_claude*.go` (backend legs); `internal/runtime/sync_4dim_binding.go` and
  its test (decision D9); `internal/hook/**` (including the Stop-hook wrappers
  and their 900 s budgets); `internal/mcp/catalog.go`;
  `.claude/skills/moai/workflows/review.md` and its template copy (decision
  D11); the distributed template `workflow.yaml`; the Go default `Audit.Model`;
  `.claude/workflows/sync-audit-4dim.js` except its header `VERDICT SCOPING`
  comment and `meta.description` text (no phase, schema, prompt or
  verdict-computation change); every pre-existing local-vs-template difference
  of research.md R-15 (edit the passages this plan names, nothing else in those
  files).
- **Structural guards the new code must pass:**
  `TestConvergence_NoNewAuditModelEnum`, `TestConvergence_NoDirectFrontmatterRead`
  (no `frontmatter` / `agent_overrides` / `AuditModel* =` in
  `mcp_convergence.go`), `TestAuditMulti_NoHardErrorPath_AC_AMM_024`,
  `TestClaudeAuditTemplateSurfacesAndCatalogHash`, the subagent-boundary rule
  (no `AskUserQuestion` / `mcp__askuser__` in any new CLI file — add the
  `TestNew_NoAskUserQuestion`-style static guard for the verb).
- **No inline env names or thresholds** — `internal/config/envkeys.go` constants
  only; the one new timeout is the derived variable of REQ-ACV-020
  (`DefaultCodexAuditLegTimeout = DefaultCodexAuditTimeout`), not a literal.
- **Raw reads only.** No new code path may build the resolver's input from
  `NewDefaultWorkflowConfig()` / `NewDefaultConfig()` or any merged loader
  (REQ-ACV-004; AC-ACV-004's grep and AC-ACV-008's pins-only fixture).
- **Template copies carry no internal identifiers.** Text added to any file
  under `internal/template/templates/` (agents, skill, `sync.md`, the script)
  must contain no SPEC id, card id or incident narrative —
  `TestTemplateNoInternalContentLeak` fails the build otherwise; only the local
  copies may cite them.
- **Template-First.** Edit `internal/template/templates/…` first, then mirror
  the local copy, then `make agents-emit` (regenerates the two template `.codex`
  role TOMLs), then `make build` (agents-emit-check → … →
  `gen-catalog-hashes --all` → build; refreshes the skill's catalogue hash).
  Local-vs-template parity is required only where a test demands it: the
  `## MCP Audit Tools (cross-model second opinion)` section of each auditor, and
  the skill directory hash. Installing a release-candidate binary is the
  leader's step (plan.md §J) and is not part of any milestone.
- **Language:** code comments English (`code_comments: en`); instruction
  documents English; commit subjects English with the card id `t1423`.
- **Cross-platform:** the new CLI file uses `filepath` only; verify with the
  Windows build above.
- **Never** `git add -A` / `git add .` (shared checkout rule) — stage by
  pathspec.

## §E Milestones

Ordering rule (D-5): M1 first by the commit-graph rule; M2-M4 carry the
decisions likeliest to change; M5-M6 are mechanical and activate the committed
config; M7 — the activation text — is the card's last commit. Each milestone is
one or more commits whose subject names the SPEC and card.

### M1 — Baseline-first characterization (own commit; precedes every behaviour commit)

Why first: AC-ACV-008 asserts "byte-identical with no config and no args", a
claim whose ordering the commit graph alone can witness (verification-claim-
integrity §2.3). The golden must be committed BEFORE the code that could change
it.

- New `internal/cli/mcp_audit_multi_baseline_test.go` and
  `internal/cli/testdata/audit_multi_default.golden.json`: drive
  `handleAuditMulti` with stubbed `backendCall` and no `gates`, over **two**
  temp project roots — one with NO `workflow.yaml`, one with the distributed
  shape (a `workflow.yaml` whose `audit` block carries only the three backend
  pins, copied in shape from the template yaml); zero `BuildCommit`/`BuildLag`;
  compare the marshalled result to the golden. Cover an all-pass case and a
  codex-inconclusive fail-open case per root. All must pass on the pre-change
  tree.
- Hermetic roots (D-4): add a helper that sets `CLAUDE_PROJECT_DIR` to
  `t.TempDir()` and apply it to every audit test in `internal/cli` that reaches
  `resolveProjectDir()` without `project_root` — the families measured in
  research.md R-29 (`TestAuditMulti_`, `TestRunMultiAudit_`, `TestConverge_`,
  `TestCodexAudit…`, `TestPerform…Audit_`, `TestPersist…`, the review-gate
  tests, `TestConfigOrphanedWorktree…`, `TestWSR…`) — and whatever AC-ACV-018
  shows is still exposed.
- Commit subject: `test(SPEC-AUDIT-MODEL-CONVERGE-001): M1 baseline golden + hermetic audit test roots (card t1423)`.
- Flips: AC-ACV-008 is GREEN here and must stay GREEN.

### M2 — The resolver (the interface most likely to change)

- New `internal/config/audit_plan.go` (+ `audit_plan_test.go`): the pure
  resolver per design.md §D.1-D.3 (the `claude` row of decision D7'), error text
  per REQ-ACV-003, `ExplicitGates()`, source labels. Table tests over every
  token × every gate source, an invalid token, an invalid gate value, and the
  precedence ladder. Update the `ValidAuditModels` / `AuditModel*` doc comments
  minimally (the deferral wording goes in M5).
- RED first: capture the failing output before the resolver exists.
- Flips: AC-ACV-001, -002, -003, -004.

### M3 — Consumers and the codex-leg deadline

- `internal/cli/mcp_audit_multi.go`: supplied-only gate reader; resolver call fed
  by `workflowAuditPins`; tool error on invalid configuration; `PlanSource`
  threading.
- `internal/cli/mcp_convergence.go`: add `PlanSource` (`omitempty`) to
  `ConvergenceResult` and `MultiAuditConfig`; wrap `performCodexAudit`'s context
  in `context.WithTimeout(…, config.DefaultCodexAuditLegTimeout)` and, after the
  RPC returns, return `inconclusiveReview` for ANY end of the leg context,
  discarding the session's text, with the summary naming a timeout or a
  cancellation (design.md §D.10 item 6). Nothing else in the engine changes
  (mind the two string guards of §D).
- `internal/config/defaults.go`: add `var DefaultCodexAuditLegTimeout =
  DefaultCodexAuditTimeout` beside its siblings with the same "distinct, not a
  const" comment.
- `internal/cli/mcp_worktree_root.go`: `resolveAuditGates` returns the plan's
  explicit gates (signature unchanged; the orphaned-worktree rules untouched).
  `internal/cli/audit_pin.go` only if a raw-config accessor is needed.
- `internal/auditreceipt/store.go`: `rawCodexGate` reads `audit.model` too and
  asks the resolver; keeps the fail-open reading for an invalid value.
- Tests: new `internal/cli/mcp_audit_multi_config_plan_test.go` (fallback per
  token including the `claude` row, args-win, fail-closed named, receipt on unmet
  gate, GLM origin under `multi`, parallel fan-out under `multi`, and the
  deadline tests: a `codexReviewRPC` stub that blocks until its context ends with
  `config.DefaultCodexAuditLegTimeout` shortened to milliseconds — an
  `inconclusive` entry naming a timeout, `overall_verdict: fail`, `gate_unmet:
  codex`; a stub returning a `pass` only after the deadline; a caller
  cancellation while the stub holds partial pass text; and the unconfigured
  fail-open case), extend `internal/auditreceipt/store_test.go` and
  `internal/cli/mcp_audit_receipt_test.go`.
- Flips: AC-ACV-006, -007, -009, -010, -017, -019, -020 (AC-008 stays GREEN;
  AC-005 completes in M4).

### M4 — The read-only verb and its result check

- New `internal/cli/audit_plan_cmd.go`: `moai verify audit-plan` registered
  through `verifyExtraCommands`, sharing `verify`'s `--project-root` (resolved
  through `verifyResolveRoot`), plus `--check-session <id>`; JSON per design.md
  §D.5, including the `unreadable` state (read through
  `loadWorkflowAuditSection`, with the config-orphaned-worktree routing of
  `resolveAuditGates`) and the `convergence_check` object; `audit-plan:`-prefixed
  errors, exit codes 0 / 1 / 2. New `audit_plan_cmd_test.go`: plan printing for
  every token (an explicit `model: claude` fixture asserts the D7' row), the
  pins-only and absent shapes, invalid configuration, the corrupt-yaml fixture,
  project-root handling, the output contract (what an older binary cannot
  produce), the result-check fixtures (an older-server-shaped result with overall
  pass, empty `gate_unmet` and no `plan_source`; an `inconclusive` codex entry;
  a wrong gate; a good result), the writes-nothing test (state directory
  byte-identical, no receipt, the `backendCall` seam never invoked) and the
  `AskUserQuestion` static guard.
- Flips: AC-ACV-005 (complete), -011, -021, -022.

### M5 — Stale-deferral cleanup (mechanical)

- `internal/cli/mcp_audit.go`: delete `multiConvergenceImplemented` and
  `activeAuditBackend`, fix the header comment; keep `buildAuditEnvBlock`.
  `internal/cli/mcp_audit_test.go`: delete the three `TestActiveAuditBackend_*`
  tests. `internal/config/audit_models.go`, `closed_sets.go`,
  `internal/settings/schema_sections.go`: comment-only edits.
  `internal/web/mcp_audit_surface_test.go`: sentinel → `ResolveAuditPlan` + a
  positive control that the symbol exists in non-test source of
  `internal/config`.
- Flips: AC-ACV-015.

### M6 — Config surface (it activates the behaviour in this repository)

- `.moai/config/sections/workflow.yaml`: `model: multi`; claude pin
  `effort: high`; header comment corrected. New test (in `internal/config/`)
  loading the committed file and asserting `Audit.Model == "multi"` and
  `Audit.Claude.Effort == DefaultClaudeAuditEffort`.
- Verification: AC-ACV-018 over the measured packages (existing audit tests
  under `CLAUDE_PROJECT_DIR` = the tree root), then `make build` is deferred to
  M7 (it needs the M7 template and catalogue edits to be in place).
- Flips: AC-ACV-016, -018.

### M7 — Activation text (the card's LAST commit)

- Template sources first, then local mirrors, in one commit:
  `plan-auditor.md`, `sync-auditor.md` (the `## MCP Audit Tools (cross-model
  second opinion)` section identical on both sides; **the
  `<!-- moai:closure-second-review:start -->` … `:end -->` markers in
  `sync-auditor.md` are preserved** when the token-keyed block is removed),
  `skills/moai-ref-cross-model-audit/SKILL.md` (plan-driven table; both copies
  identical), `skills/moai/workflows/sync.md` (the `Binding promotion` paragraph
  and a pointer sentence in `FO-SYNC-1`: the added post-PASS `audit_multi` call,
  the result check, the "both must pass" rule, the statement lines, the
  unmet/unreachable record). The old token-keyed block becomes the labelled
  **legacy path** paragraph; the main path is the verb. Required literals:
  `moai verify audit-plan` and `plan surface unreachable, legacy path used` in
  the two agents, the skill and `sync.md`; `cross_model_required`,
  `--check-session` and `audit_multi unreachable` in `sync.md`; the old
  "Single-backend audit mode (per the project's `audit_model`)" block heading gone.
- `.claude/workflows/sync-audit-4dim.js` (template source first, then local):
  the `VERDICT SCOPING` header comment and `meta.description` text only
  (design.md §D.8 table). Do not "fix" the pre-existing differences between the
  two copies (research.md R-15).
- A doc-surface test in `internal/template/` (new file, `TestAuditPlanDocSurface`)
  asserting the literals in the ten document copies and the absence of the old
  heading.
- `make agents-emit`, then `make build` (catalogue hash; all emit/drift checks).
- Flips: AC-ACV-012, -013, -014.

## §F Verification commands

All Go commands carry `<SCRUB>` except AC-ACV-018's. Per-milestone, scoped:

```bash
<SCRUB> go test -count=1 ./internal/config/ ./internal/auditreceipt/
<SCRUB> go test -count=1 -run '^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestConverge_.*|TestCodexAudit.*|TestAuditPlanCmd_.*)$' ./internal/cli/
<SCRUB> go test -count=1 ./internal/template/ ./internal/web/
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m
```

Read each run's swept count before its verdict: a `-run` pattern that selects
nothing prints `ok` (verification-completeness §1.1). Quote the `--- PASS:`
lines of the named tests.

## §G Anti-patterns (named for the run to refuse)

- **AP-1 Resolver in the wrong package** — putting it in `mcp_convergence.go`
  trips two structural guards; putting it in `internal/cli` makes
  `internal/auditreceipt` unable to reach it.
- **AP-2 Token without enforcement** — making `audit_multi` follow `multi` while
  leaving the three raw-gate readers alone reproduces fail-open (research.md
  §R.3).
- **AP-3 A new MCP tool or a top-level `moai audit` group** — rejected in
  design.md §D.4; the verb is `moai verify audit-plan`.
- **AP-4 Silent default on a bad token** — the resolver's error must reach the
  caller (REQ-ACV-003).
- **AP-5 Fixing pre-existing local/template drift** — research.md R-15; not
  this SPEC's change.
- **AP-6 Sweep-staging** — `git add -A` is forbidden in the shared checkout.
- **AP-7 Moving the sync call into the script** — `audit_multi` is write-capable
  and the script's agents are read-only with a pure-JS verdict; the call belongs
  to the orchestrator in `sync.md` (design.md §D.8). Edit the script's header and
  description text only.
- **AP-8 Demoting the 4-dimension verdict, or touching `IsBinding`** — superseded
  by decisions D6 and D9; the verdict stays the Claude input and binding together
  with a passing `audit_multi` and a good result check.
- **AP-9 A silent codex timeout** — a leg that ends by deadline OR cancellation
  must return `inconclusive` naming the cause, never a verdict parsed from
  partial output.
- **AP-10 A default-merged read** — feeding the resolver from
  `NewDefaultWorkflowConfig()` turns every unconfigured project into `claude`.
- **AP-11 Treating the unknown verb as an error** — `moai verify <unknown>` prints
  help and exits 0 on an older binary; detect the verb by its output contract.
- **AP-12 Calling the unreachable verb fail-closed** — an absent verb is the
  legacy path with a named Gap, not an outage; a configured required backend
  that does not answer is still fail-closed (REQ-ACV-014).

## §H Cross-references

`spec.md` · `design.md` · `research.md` · `acceptance.md` ·
`.claude/rules/moai/core/verification-claim-integrity.md` §1-§3.1 ·
`.claude/rules/moai/development/verification-completeness.md` §1-§3 ·
`.claude/rules/moai/core/moai-mcp-tools.md` · SPEC-AUDIT-MULTI-MODEL-001 ·
SPEC-CODEX-AUDIT-GATE-AXES-001 · SPEC-AUDIT-SNAPSHOT-001 ·
`.moai/reports/t1423/progress.md` (operator decisions D1-D11) ·
`.moai/reports/t1423/plan-audit-iter1.md` (the PA1 defects this version answers).

## §I Gaps (unobserved in this plan phase)

- No Go test of the SPEC's own behaviour was run (this phase edits no Go); every
  GREEN claim is a prediction until M1-M7 run. One throwaway behavioural test
  was run through `go test -overlay` for the RED of AC-ACV-009 (acceptance.md
  E22), with the test source outside the tree.
- Live codex / GLM behaviour; wizard persistence of `claude`; Codex-hosted
  execution of the verb; `moai update` preserving `model: multi` (research.md
  §R.4).
- The exact set of existing tests that read `resolveProjectDir()` unprotected is
  found by AC-ACV-018, not enumerated here.
- Real codex leg durations (OQ-8) and the persistence location of the sync
  binding statement (OQ-9) are unobserved.
- Whether an `Explore` workflow agent could carry MCP tools was not observed;
  the design does not depend on it (the call stays with the orchestrator).
- The `moai verify` group's behaviour for an unknown verb was measured on the
  installed binary only; a future cobra or group change would alter the trap of
  AP-11 (the output contract does not depend on it).

## §J Rollout, activation, and the Definition-of-Done install item

The verb's source does not exist on develop until this card merges, so
installing before the merge is impossible (leader correction). The safety comes
from behaviour — the legacy path of REQ-ACV-014 — with ordering as a second
layer: the activation text is the card's last commit (M7).

Sequence after the card's commits are done:

1. The leader merges the card into develop (the lane protocol's integration
   window).
2. A release-candidate build is made from develop and installed
   (`.claude/rules/local/gitflow-lane-protocol.md` §9; operator-requested).
3. The MCP server is reconnected (a new `audit_multi` result carries
   `plan_source`; an older server's does not).
4. Smoke, on the installed build, from the worktree or primary checkout that will
   run the audit: `moai verify audit-plan --project-root <abs toplevel>` prints a
   JSON plan with `config_status: "ok"`, `model: "multi"` in this repository; one
   `audit_multi` call on a trivial target with a `session_id`; then
   `moai verify audit-plan --project-root <abs toplevel> --check-session <id>`
   prints `convergence_check.ok: true`.
5. Only then does the new audit path apply to later audits. Until step 4 passes,
   the auditors and the sync step run the legacy path with the named Gap, which
   is the pre-change behaviour (no outage, nothing fail-closed on the missing
   verb).

Steps 2 and 3 are taken together: a new CLI with the old MCP server still
running is the one window in which the result check fails closed by name
(spec.md R-8). The Definition of Done (acceptance.md) carries this sequence as
its install-and-verify item.
