# SPEC-AUDIT-MODEL-CONVERGE-001 — plan

> Tier L plan artifact. Requirements: spec.md §C. Reasoning and shapes:
> design.md. Measurements: research.md §R.1. Acceptance: acceptance.md. This
> file has no frontmatter `status:`.

## §A Context

- **Card** t1423, Class C (design change). Worktree `.moai/worktrees/t1423`,
  branch `WT-audit-model-convergence`, base HEAD `c50da9c2f`.
- **Methodology**: `development_mode: tdd` (research.md R-18) — RED first on
  every behaviour milestone, with the verbatim failing output captured before
  GREEN (manager-develop-prompt-template §E8).
- **Tier L** — about 40 files, 7 milestones (spec.md §E, re-derived at 0.1.1). The orchestrator's
  `manager-lead` entry predicate (>= 3 milestones AND >= 10 files) is met; the
  run-phase mode is the orchestrator's decision and is logged in
  `progress.md §F`. Default if not routed to `manager-lead`: `serial`, one
  `manager-develop` per milestone (coding-heavy work).
- **Self-application.** Plan-audit of this SPEC should itself call
  `audit_multi` with `gates: {claude: required, codex: required, glm: advisory}`
  and `project_root` set to this worktree's toplevel, and cite the receipts. The
  config toggle does not exist yet; the call argument does.

## §B Decisions and open items

### B.1 Decisions taken here

- **D-1 Surface = CLI verb** `moai audit plan` (design.md §D.4).
- **D-2 Symbols the acceptance criteria name.** `config.ResolveAuditPlan`,
  `config.AuditPlan`, method `AuditPlan.ExplicitGates()`, file
  `internal/config/audit_plan.go`. The CLI verb lives in
  `internal/cli/audit_plan_cmd.go`; the codex leg limit `codexAuditTimeout` and its
  test seam `codexLegTimeout` in `internal/cli/mcp_codex.go`. A rename is a cheap
  edit in acceptance.md §AC-ACV-004/005/020 and nothing else.
- **D-3 `plan_source`** is the one new `ConvergenceResult` member
  (`omitempty`, design.md §D.6).
- **D-4 Test hermeticity.** Every test that reaches `resolveProjectDir()` without
  passing `project_root` sets `CLAUDE_PROJECT_DIR` to its own `t.TempDir()`
  (M1), so the committed yaml of M7 cannot reach it.
- **D-5 Order.** M1 first (baseline-first, §2.3); then decision-bearing
  milestones by how likely they are to change (resolver interface, consumers,
  the verb's shape, the agent/skill text); the mechanical cleanup and the config
  flip that activates everything last.

### B.2 Open items

- **OQ-1 — closed as a flagged assumption (decision D4).** Adopted at
  confidence 0.57; **plan-audit please re-weigh.** Question: is the `claude`
  token "Claude alone" (codex and glm off — design.md §D.1) or "the default
  profile" (identical to the empty token)? Evidence:
  `internal/config/audit_models.go` calls `claude` both "the default
  single-backend audit model" and part of "the locked distributed-default
  profile … claude+codex required, glm advisory"; the skill table says
  `audit_model: claude (default)` runs Claude only. Reading adopted: Claude
  alone. Consequence if the other reading is right: a project that persisted
  `claude` (wizard persistence unobserved) would have codex turned off in
  `audit_multi` instead of required-but-fail-open — a behaviour change only for
  a caller of `audit_multi` in such a project, and one the operator could undo
  with `audit.gates.codex: required`.
- **OQ-2 — closed (decision D5, confidence 0.92).** The codex leg of
  `audit_multi` gets a named deadline, `codexAuditTimeout = claudeAuditTimeout`
  (5 minutes), derived in design.md §D.10; expiry is an `inconclusive` entry
  naming the codex timeout, hence an unmet gate under a `required` codex
  (REQ-ACV-020, AC-ACV-020, M3). This reverses the earlier default of "no new
  deadline".
- **OQ-3** Assumption, flagged: a tool-argument `required` is not explicit
  (design.md §D.2). Correct it before Kickoff if the operator wants arguments to
  fail closed too; the change is one source label and one test.
- **OQ-4** Assumption, flagged: an unreadable `workflow.yaml` keeps today's
  fail-open reading on the `audit_multi` path (spec.md R-4); the verb reports it.
- **OQ-5** Assumption, flagged: the spelling `moai audit plan` (new top-level
  `audit` group). Alternatives: `moai doctor audit`, `moai config audit-plan`
  (the `config` group's help says it holds mutating operations).
- **OQ-6** Not observed: whether a Codex-hosted read-only auditor role can
  execute `moai audit plan`. If it cannot, the codex roles keep an `audit_multi`
  call without the plan step and REQ-ACV-014's unreachable-surface Gap applies.
- **OQ-7 — closed (decision D6, confidence 0.89).** The 4-dimension verdict
  stays the Claude input and keeps its A3 fast path; after a clean PASS, when
  codex is required, the orchestrator additionally calls `audit_multi`, and the
  sync verdict is binding only when both pass (REQ-ACV-015, REQ-ACV-021,
  design.md §D.8). This reverses the earlier default of demoting the verdict to
  non-binding and spawning the cold `sync-auditor`.
- **OQ-8** Assumption, flagged (new): 5 minutes is enough for a real codex
  adversarial review. No codex review duration was measured (research.md §R.4);
  the value is derived from the Claude stage limit, not observed. If a
  legitimate review outlasts it, a `required` codex gate fails closed, by name.
  Measure a few real codex audit durations before the deadline is relied on.
- **OQ-9** Assumption, flagged (new): where the orchestrator's statement of the
  binding sync verdict is persisted for the sync-audit to re-read is not
  specified in `sync.md` (grep of `sync.md` and `sync/quality-gates-quality.md`
  found no record location). The cross-model line is added to that statement
  (REQ-ACV-021) without pinning a file; the sync-audit's re-read of it is
  unverified.

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
run the full suite. Do not run the `internal/cli` full suite locally (it is a
heavy-run resource; the slot lease applies if a wider run is ever needed).

## §D Constraints

- **PRESERVE (do not edit):** `internal/cli/mcp_codex.go` (except adding the
  `codexAuditTimeout` constant beside its other codex constants), `mcp_glm.go`,
  `mcp_claude*.go` (backend legs); `.claude/workflows/sync-audit-4dim.js`
  (except the header `VERDICT SCOPING` comment and `meta.description` text — no
  phase, schema, prompt or verdict-computation change); `internal/hook/**` (including the Stop-hook
  wrappers and their 900 s budgets); `internal/mcp/catalog.go`;
  `.claude/skills/moai/workflows/review.md`;
  the distributed template `workflow.yaml`; the Go default `Audit.Model`; every
  pre-existing local-vs-template difference of research.md R-15 (edit the
  passages this plan names, nothing else in those files).
- **Structural guards the new code must pass:**
  `TestConvergence_NoNewAuditModelEnum`, `TestConvergence_NoDirectFrontmatterRead`
  (no `frontmatter` / `agent_overrides` / `AuditModel* =` in
  `mcp_convergence.go`), `TestAuditMulti_NoHardErrorPath_AC_AMM_024`,
  `TestClaudeAuditTemplateSurfacesAndCatalogHash`, the subagent-boundary rule
  (no `AskUserQuestion` / `mcp__askuser__` in any new CLI file — add the
  `TestNew_NoAskUserQuestion`-style static guard for the verb).
- **No inline env names or thresholds** — `internal/config/envkeys.go` constants
  only; the one new timeout is the derived constant of REQ-ACV-020
  (`codexAuditTimeout = claudeAuditTimeout`), not a literal.
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
  operator's step (gitflow lane protocol §9) and is not part of this plan.
- **Language:** code comments English (`code_comments: en`); instruction
  documents English; commit subjects English with the card id `t1423`.
- **Cross-platform:** the new CLI file uses `filepath` only; verify with the
  Windows build above.
- **Never** `git add -A` / `git add .` (shared checkout rule) — stage by
  pathspec.

## §E Milestones

Ordering rule (D-5): M1 first by the commit-graph rule; M2-M5 carry the
decisions likeliest to change; M6-M7 are mechanical and activate last. Each
milestone is one or more commits whose subject names the SPEC and card.

### M1 — Baseline-first characterization (own commit; precedes every behaviour commit)

Why first: AC-ACV-008 asserts "byte-identical with no config and no args", a
claim whose ordering the commit graph alone can witness (verification-claim-
integrity §2.3). The golden must be committed BEFORE the code that could change
it.

- New `internal/cli/mcp_audit_multi_baseline_test.go` and
  `internal/cli/testdata/audit_multi_default.golden.json`: drive
  `handleAuditMulti` with stubbed `backendCall`, a temp project root with NO
  `workflow.yaml`, and no `gates`; zero `BuildCommit`/`BuildLag`; compare the
  marshalled result to the golden. Cover an all-pass case and a
  codex-inconclusive fail-open case. Both must pass on the pre-change tree.
- Hermetic roots (D-4): add a helper that sets `CLAUDE_PROJECT_DIR` to
  `t.TempDir()` and apply it to the audit tests that reach
  `resolveProjectDir()` without `project_root` — in `mcp_audit_multi_test.go`,
  `mcp_convergence_test.go`, `mcp_audit_receipt_test.go`, and any
  `mcp_codex*_test.go` / `mcp_glm*_test.go` / `audit_pin*_test.go` test that
  `AC-ACV-018` shows is exposed.
- Commit subject: `test(SPEC-AUDIT-MODEL-CONVERGE-001): M1 baseline golden + hermetic audit test roots (card t1423)`.
- Flips: AC-ACV-008 is GREEN here and must stay GREEN.

### M2 — The resolver (the interface most likely to change)

- New `internal/config/audit_plan.go` (+ `audit_plan_test.go`): the pure
  resolver per design.md §D.1-D.3, error text per REQ-ACV-003, `ExplicitGates()`,
  source labels. Table tests over every token × every gate source, an invalid
  token, an invalid gate value, and the precedence ladder. Update the
  `ValidAuditModels` / `AuditModel*` doc comments minimally (the deferral
  wording goes in M6).
- RED first: capture the failing output before the resolver exists.
- Flips: AC-ACV-001, -002, -003, -004.

### M3 — Consumers (audit_multi, enforcement, receipts)

- `internal/cli/mcp_audit_multi.go`: supplied-only gate reader; resolver call;
  tool error on invalid configuration; `PlanSource` threading.
- `internal/cli/mcp_convergence.go`: add `PlanSource` (`omitempty`) to
  `ConvergenceResult` and `MultiAuditConfig`; wrap `performCodexAudit`'s context
  in the codex-leg deadline and return `inconclusiveReview` naming the codex
  timeout when the leg context ended by deadline and the caller's did not,
  discarding any partial review text (design.md §D.10). Nothing else in the
  engine changes (mind the two string guards of §D).
- `internal/cli/mcp_codex.go`: add `codexAuditTimeout = claudeAuditTimeout`
  beside the other codex constants, and the package variable `codexLegTimeout`
  initialised from it (the test seam).
- `internal/cli/mcp_worktree_root.go`: `resolveAuditGates` returns the plan's
  explicit gates (signature unchanged; the orphaned-worktree rules untouched).
  `internal/cli/audit_pin.go` only if a raw-config accessor is needed.
- `internal/auditreceipt/store.go`: `rawCodexGate` reads `audit.model` too and
  asks the resolver; keeps the fail-open reading for an invalid value.
- Tests: new `internal/cli/mcp_audit_multi_config_plan_test.go` (fallback per
  token, args-win, fail-closed named, receipt on unmet gate, GLM origin under
  `multi`, parallel fan-out under `multi`, and the codex-leg deadline test: a
  `codexReviewRPC` stub that blocks until its context ends, `codexLegTimeout`
  shortened to milliseconds, `model: multi` configured — assert an `inconclusive`
  codex entry whose summary names the timeout, `overall_verdict: fail`,
  `gate_unmet: codex`; a second sub-case where the stub returns a `pass` only
  after the deadline and the entry is still `inconclusive`), extend
  `internal/auditreceipt/store_test.go` and `internal/cli/mcp_audit_receipt_test.go`.
- Flips: AC-ACV-005, -006, -007, -009, -010, -017, -019, -020 (AC-008 stays
  GREEN).

### M4 — The read-only surface

- New `internal/cli/audit_plan_cmd.go`: `moai audit plan [--project-root <abs>]`,
  JSON per design.md §D.5, exit codes 0 / 1 / 2, registered in the root command's
  existing registration point; new `audit_plan_cmd_test.go` including the
  writes-nothing test (state directory byte-identical, no receipt, the
  `backendCall` seam never invoked) and the `AskUserQuestion` static guard.
- Flips: AC-ACV-011.

### M5 — Auditors, skill, sync binding

- Template sources first, then local mirrors: `plan-auditor.md`,
  `sync-auditor.md` (the `## MCP Audit Tools (cross-model second opinion)`
  section identical on both sides), `skills/moai-ref-cross-model-audit/SKILL.md`
  (plan-driven table; both copies identical), `skills/moai/workflows/sync.md`
  (the `Binding promotion` paragraph and a pointer sentence in `FO-SYNC-1`: the
  added post-PASS `audit_multi` call, the "both must pass" rule, the
  `cross_model:` lines of the binding statement, the unmet/unreachable record).
  Required literals: `moai audit plan` and `plan surface unreachable` in the two
  agents and the skill; `cross_model_required` and `audit_multi unreachable` in
  `sync.md`; the old "Single-backend audit mode (per the project's
  `audit_model`)" block and the token-keyed table removed.
- `.claude/workflows/sync-audit-4dim.js` (template source first, then local):
  the `VERDICT SCOPING` header comment and `meta.description` text only
  (design.md §D.8 table). Do not "fix" the pre-existing differences between the
  two copies (research.md R-15).
- `internal/runtime/sync_4dim_binding.go` (+ test): the cross-model requirement,
  outcome and unmet-backend inputs and the `cross-model` reasons (design.md
  §D.8).
- A doc-surface test in `internal/template/` (new file, `TestAuditPlanDocSurface`)
  asserting the literals in the ten document copies and the absence of the old
  prose.
- `make agents-emit`, then `make build` (catalogue hash).
- Flips: AC-ACV-012, -013, -014, -021.

### M6 — Stale-deferral cleanup (mechanical)

- `internal/cli/mcp_audit.go`: delete `multiConvergenceImplemented` and
  `activeAuditBackend`, fix the header comment; keep `buildAuditEnvBlock`.
  `internal/cli/mcp_audit_test.go`: delete the three `TestActiveAuditBackend_*`
  tests. `internal/config/audit_models.go`, `closed_sets.go`,
  `internal/settings/schema_sections.go`: comment-only edits.
  `internal/web/mcp_audit_surface_test.go`: sentinel → `ResolveAuditPlan` + a
  positive control that the symbol exists in non-test source of
  `internal/config`.
- Flips: AC-ACV-015.

### M7 — Config surface (last: it activates the behaviour in this repository)

- `.moai/config/sections/workflow.yaml`: `model: multi`; claude pin
  `effort: high`; header comment corrected. New test (in `internal/config/`)
  loading the committed file and asserting `Audit.Model == "multi"` and
  `Audit.Claude.Effort == DefaultClaudeAuditEffort`.
- Verification: AC-ACV-018 (existing audit tests under `CLAUDE_PROJECT_DIR` =
  the tree root), then `make build` (all three emit/drift checks).
- Flips: AC-ACV-016, -018.

## §F Verification commands

All Go commands carry `<SCRUB>` except AC-ACV-018's. Per-milestone, scoped:

```bash
<SCRUB> go test -count=1 ./internal/config/ ./internal/auditreceipt/ ./internal/runtime/
<SCRUB> go test -count=1 -run '^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestConverge_.*|TestCodexAudit_.*|TestAuditPlanCmd_.*)$' ./internal/cli/
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
- **AP-3 A new MCP tool** — rejected in design.md §D.4; do not add `audit_plan`
  to the catalogue.
- **AP-4 Silent default on a bad token** — the resolver's error must reach the
  caller (REQ-ACV-003).
- **AP-5 Fixing pre-existing local/template drift** — research.md R-15; not
  this SPEC's change.
- **AP-6 Sweep-staging** — `git add -A` is forbidden in the shared checkout.
- **AP-7 Moving the call into the script** — `audit_multi` is write-capable and
  the script's agents are read-only with a pure-JS verdict; the call belongs to
  the orchestrator in `sync.md` (design.md §D.8). Edit the script's header and
  description text only.
- **AP-8 Demoting the 4-dimension verdict** — superseded by decision D6; the
  verdict stays the Claude input and binding together with a passing
  `audit_multi`.
- **AP-9 A silent codex timeout** — a leg that outlasts the deadline must return
  `inconclusive` naming the timeout, never a verdict parsed from partial output.

## §H Cross-references

`spec.md` · `design.md` · `research.md` · `acceptance.md` ·
`.claude/rules/moai/core/verification-claim-integrity.md` §1-§3.1 ·
`.claude/rules/moai/development/verification-completeness.md` §1-§2.1 ·
`.claude/rules/moai/core/moai-mcp-tools.md` · SPEC-AUDIT-MULTI-MODEL-001 ·
SPEC-CODEX-AUDIT-GATE-AXES-001 · SPEC-AUDIT-SNAPSHOT-001 ·
`.moai/reports/t1423/progress.md` (the operator decisions D1-D6).

## §I Gaps (unobserved in this plan phase)

- No Go test was run (this phase edits no Go); every GREEN claim is a
  prediction until M1-M7 run.
- Live codex / GLM behaviour; wizard persistence of `claude`; Codex-hosted
  execution of the verb; `moai update` preserving `model: multi` (research.md
  §R.4).
- The exact set of existing tests that read `resolveProjectDir()` unprotected is
  found by AC-ACV-018, not enumerated here.
- `moai audit plan` spelling and group placement (OQ-5) are unreviewed by the
  operator.
- Real codex review durations (OQ-8) and the persistence location of the sync
  binding statement (OQ-9) are unobserved.
- Whether an `Explore`-agent workflow could carry MCP tools was not observed;
  the design does not depend on it (the call stays with the orchestrator).
