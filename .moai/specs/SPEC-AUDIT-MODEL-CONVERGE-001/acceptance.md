# SPEC-AUDIT-MODEL-CONVERGE-001 — acceptance

> Verification layer. Each criterion is Given-When-Then, binary, and names the
> `**Covers**` requirement(s) of spec.md §C. Judgement is by the real output of a
> real command. Go test names are the tests the run phase writes (plan.md §E);
> document items are judged by grep and by the template doc-surface test. 21
> criteria (Tier L ceiling 25; amended at 0.1.1 for decisions D5 and D6). This file has no frontmatter `status:`.
>
> **Command convention.** A RED-now cell is a plain single-invocation read-only
> command recorded in the evidence ledger below with its verbatim stdout and exit
> code (`verification-completeness.md` §2.1) — no pipes, redirection, `&&`, `;`,
> or subshells. The green-path Go commands carry `<SCRUB>` = the one compound
> prefix `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && `
> (plan.md §C); a prefix that carries `&&` is outside the single-invocation
> form, so a ledger row that needs it is informational and no criterion's RED-now
> rests on it alone. `-run` patterns are anchored (`^…$`): an unanchored pattern
> also selects longer names, and a pattern that selects nothing still prints `ok`.
>
> **Classification.** **release-blocking** = a single-invocation RED-now ledger
> row exists and reproduces on tree `c50da9c2f`. **regression-guard** = no
> RED-now; GREEN on arrival and must stay GREEN. Release-blocking: 001-007,
> 009-016, 020, 021. Regression-guard: 008, 017, 018, 019.
>
> **Baseline-first ordering (verification-claim-integrity §2.3).** AC-ACV-005's
> "zero non-test readers before, several after" and AC-ACV-008's "byte-identical"
> are claims about ordering. The baselines are: the ledger below, committed in
> the plan-phase commit that carries this file (it precedes every run-phase
> commit), and the golden file of plan.md M1, which lands in its own commit
> before any behaviour change.
>
> **RED-now / green-path pairing.** Each criterion states why it is red on
> `c50da9c2f` and which milestone flips it. A new behaviour's RED-now is the
> absence of the code that would provide it; each named Go test's own RED
> (verbatim failing output) is captured by manager-develop in M2-M6 before
> GREEN, per manager-develop-prompt-template §E8 — those captures are run-phase
> evidence, not ledger rows here.

## Evidence ledger (RED-now observations, tree `c50da9c2f`)

Every row was run in this plan-phase run against tree `c50da9c2f`; the pin binds
every criterion that carries none of its own. Exit codes were read through a
`sh -c '…; echo exit=$?'` wrapper; the command column is the plain command.

| id | command | verbatim stdout | exit | meaning |
|---|---|---|---|---|
| E1 | `grep -rln --exclude="*_test.go" "ResolveAuditPlan" internal cmd` | (no output) | 1 | no non-test file names the resolver |
| E1c | `grep -rln --exclude="*_test.go" "ValidAuditModels" internal cmd` | `internal/settings/schema_sections.go` · `internal/config/closed_sets.go` | 0 | positive control: the same form hits a symbol that exists, so E1's empty output is not an empty scan |
| E2 | `grep -rn --exclude="*_test.go" "Audit\.Model" internal cmd` | (no output) | 1 | the baseline: ZERO non-test readers of `audit.model` |
| E2c | `grep -rn --exclude="*_test.go" "activeAuditBackend" internal cmd` | `internal/settings/schema_sections.go:367` (comment) · `internal/config/closed_sets.go:87` (comment) · `internal/cli/mcp_audit.go:6` (comment) · `:46` (comment) · `:52` (`func activeAuditBackend(model string) (string, error) {`) | 0 | one definition, four comments, no call — "defined, never read" |
| E3 | `moai audit plan --help` | ` ERROR Unknown command "audit" for "moai". Try --help for usage.` | 1 | the verb does not exist |
| E4 | `grep -c "moai audit plan" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md` | `…plan-auditor.md:0` · `…sync-auditor.md:0` · `…SKILL.md:0` | 1 | no agent or skill instructs the plan call |
| E4c | `grep -c "audit_multi" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md` | `…plan-auditor.md:4` · `…sync-auditor.md:5` · `…SKILL.md:6` | 0 | positive control: the same files do carry the tool name, so E4's zeros are not an empty scan |
| E4b | `grep -c "plan surface unreachable" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md` | `:0` · `:0` · `:0` | 1 | the unreachable-surface Gap wording is absent |
| E4d | `grep -cF "Single-backend audit mode (per the project" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` | `:1` · `:1` · `:1` · `:1` | 0 | the prose-interpretation block exists in all four agent copies |
| E4t | `grep -c "moai audit plan" internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md` | `:0` · `:0` · `:0` | 1 | the template copies carry no plan call either |
| E4x | `grep -c "moai audit plan" internal/template/templates/.codex/agents/moai/plan-auditor.toml internal/template/templates/.codex/agents/moai/sync-auditor.toml` | `:0` · `:0` | 1 | the emitted `.codex` roles carry none (they must follow the `.md` after `make agents-emit`) |
| E5 | `grep -rn "multiConvergenceImplemented" --include="*.go" internal` | `internal/cli/mcp_audit_test.go:40` · `:41` · `internal/cli/mcp_audit.go:27` · `:31` · `:49` | 0 | the stale sentinel and a test pinning it |
| E6 | `grep -rn --exclude="*_test.go" "AP-8" internal` | `internal/config/audit_models.go:28` · `internal/cli/mcp_audit.go:8` · `:27` | 0 | the deferral markers in non-test source |
| E7 | `grep -n -A1 "model: claude-opus-5-5" .moai/config/sections/workflow.yaml` | `24:            model: claude-opus-5-5` · `25-            effort: medium` | 0 | the committed claude pin is `medium` (a bare `effort: medium` grep is non-specific — the file has sixteen other hits) |
| E8 | `grep -nE "^        model: multi$" .moai/config/sections/workflow.yaml` | (no output) | 1 | the committed yaml has no `audit.model` |
| E9 | `grep -n "claude-opus-5-5/medium" .moai/config/sections/workflow.yaml` | `20:    # z.ai states low\|high\|max. Claude defaults to claude-opus-5-5/medium and` | 0 | the header comment states `medium` |
| E10 | `grep -n "effort: high" internal/template/templates/.moai/config/sections/workflow.yaml` | `136:            effort: high` · `139:            effort: high` | 0 | the distributed template's claude pin is already `high`; no template change is needed |
| E11 | `grep -n "CrossModelRequired" internal/runtime/sync_4dim_binding.go` | (no output) | 1 | the binding predicate has no cross-model input |
| E12 | `grep -n "const sentinel" internal/web/mcp_audit_surface_test.go` | `58:	const sentinel = "activeAuditBackend" // the M3 resolver symbol in internal/cli` | 0 | the web guard's sentinel names the symbol M6 deletes |
| E13 | `grep -c "internal/config" internal/auditreceipt/store.go` | `0` | 1 | the receipt store does not import the config package |
| E14 | `grep -c "cross_model_required" .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` | 1 | neither `sync.md` copy mentions the cross-model binding condition |
| E15 | `grep -c "^package config" internal/config/audit_plan.go` | (stderr) `grep: internal/config/audit_plan.go: No such file or directory` | 2 | the resolver file does not exist |
| E15c | `grep -c "^package config" internal/config/audit_models.go` | `1` | 0 | positive control: the same form hits an existing file of the package |
| E16 | `grep -c "model: multi" internal/template/templates/.moai/config/sections/workflow.yaml` | `0` | 1 | baseline for the "template unchanged" guard |
| E17 | `grep -rn "codexAuditTimeout" internal` | (no output) | 1 | no codex-leg limit exists |
| E17c | `grep -n "claudeAuditTimeout" internal/cli/mcp_claude.go` | `27:	claudeAuditTimeout       = 5 * time.Minute` · `110:	auditCtx, cancel := context.WithTimeout(ctx, claudeAuditTimeout)` | 0 | positive control: the Claude leg has its limit and applies it; the same form finds nothing for codex |
| E18 | `grep -n "context.WithTimeout" internal/cli/mcp_convergence.go` | (no output) | 1 | `performCodexAudit` and the fan-out apply no deadline to the codex leg |
| E18c | `grep -n "context.WithTimeout" internal/cli/mcp_claude.go` | `110:	auditCtx, cancel := context.WithTimeout(ctx, claudeAuditTimeout)` | 0 | positive control: the same form hits where a deadline exists |
| E19 | `grep -c "audit_multi unreachable" .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md` | `:0` · `:0` | 1 | neither `sync.md` copy records the unavailable-`audit_multi` outcome |
| E20 | `grep -c "audit_multi" .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` | `:0` · `:0` | 1 | neither script copy says the verdict is the Claude input to an `audit_multi` call |
| E21 | `grep -c "mcp__moai__" .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` | `:0` · `:0` | 1 | baseline for "the script never calls an MCP tool" — it must stay 0 after M5 |

## AC-ACV-001 — every token maps to a total plan

**Covers**: maps REQ-ACV-001

- **Classification**: release-blocking (RED-now: E1, positive control E1c; E15).
- **Given** the closed set `claude`, `codex`, `glm`, `multi` and the empty token,
  **When** each is resolved with no configured gates and no caller gates,
  **Then** each plan assigns each of `claude`, `codex`, `glm` exactly one of
  `off`, `advisory`, `required`, and the five rows equal design.md §D.1
  (empty and `multi`: required / required / advisory; `claude`: required / off /
  off; `codex`: off / required / off; `glm`: off / off / required); the empty
  token's entries carry source `default`, every other token's carry
  `config.model`.
- **RED-now**: E1/E15 — there is no resolver.
- **green** (M2): `<SCRUB>go test -count=1 -v -run '^TestResolveAuditPlan_TokenTable$' ./internal/config/` →
  `--- PASS: TestResolveAuditPlan_TokenTable ` (name followed by a space) and the
  five sub-test names, exit 0.
- **Mutant probe**: a resolver that returns only the `multi` row for every token
  fails the table; one that leaves a backend unset in any row fails the
  exactly-one-gate assertion.

## AC-ACV-002 — precedence, with the deciding source recorded

**Covers**: maps REQ-ACV-002

- **Classification**: release-blocking (RED-now: E1, E15).
- **Given** `audit.model: multi`, `audit.gates.glm: off`, and a caller gate
  `codex: advisory`, **When** the plan is resolved, **Then** codex is `advisory`
  (source `argument`), glm is `off` (source `config.gates`), claude is `required`
  (source `config.model`); and with the caller gate removed codex returns to
  `required` (source `config.model`); and with an empty token and no gates every
  entry has source `default`.
- **RED-now**: E1/E15.
- **green** (M2): `<SCRUB>go test -count=1 -v -run '^TestResolveAuditPlan_Precedence$' ./internal/config/` →
  `--- PASS: TestResolveAuditPlan_Precedence `, exit 0.
- **Mutant probe**: swapping argument and config, or config.gates and
  config.model, changes at least one asserted gate; dropping the source field
  fails the source assertions.

## AC-ACV-003 — an invalid token or gate names itself and is never defaulted

**Covers**: maps REQ-ACV-003

- **Classification**: release-blocking (RED-now: E1, E2).
- **Given** a tree whose `audit.model` is `grok` (and, separately, whose
  `audit.gates.codex` is `requird`), **When** the resolver, `audit_multi`, or
  `moai audit plan` reads it, **Then** the resolver returns an error whose text
  is `audit_model "grok" unknown (want one of claude|codex|glm|multi)` (for the
  gate: the key, the value, and `off|advisory|required`); `audit_multi` returns a
  tool error (`isError: true`) carrying that text and invokes no backend; the
  verb exits 1 with the text on stderr and nothing on stdout. Token matching is
  exact and case-sensitive, with surrounding whitespace trimmed (`Multi` is
  rejected; ` multi ` is accepted).
- **RED-now**: E1/E2 — nothing reads the token, so nothing can reject it.
- **green** (M2, M3, M4): `<SCRUB>go test -count=1 -v -run '^(TestResolveAuditPlan_RejectsUnknown|TestAuditMulti_UnknownConfiguredTokenIsToolError|TestAuditPlanCmd_InvalidConfigExitsOne)$' ./internal/config/ ./internal/cli/` →
  three `--- PASS:` lines, exit 0.
- **Mutant probe**: a resolver that maps an unknown token to the default plan
  fails all three; a handler that swallows the error and fans out fails the
  `audit_multi` test's no-backend assertion.

## AC-ACV-004 — the resolver is pure and lives in `internal/config`

**Covers**: maps REQ-ACV-004

- **Classification**: release-blocking (RED-now: E15, positive control E15c).
- **Given** the resolver source file, **When** its imports are scanned, **Then**
  `internal/config/audit_plan.go` exists, declares `package config`, and imports
  none of `os`, `time`, `net`, `io`, `path/filepath`, `os/exec`.
- **RED-now**: E15 — the file does not exist (exit 2); E15c shows the control
  form works on an existing file of the package.
- **green** (M2): `grep -c "^package config" internal/config/audit_plan.go` →
  `1`, exit 0; and `grep -nE "\"(os|time|net|io|path/filepath|os/exec)\"" internal/config/audit_plan.go` →
  (no output), exit 1.
- **Mutant probe**: a resolver that reads `os.Getenv` or the yaml file itself
  fails the second command.

## AC-ACV-005 — the resolver is consumed by non-test callers

**Covers**: maps REQ-ACV-005

- **Classification**: release-blocking (RED-now: E1, E2, E2c; baseline-first —
  the ledger commit precedes M2).
- **Given** the run-phase tree, **When** the non-test files that reference the
  resolver are listed, **Then** the list names the definition file and at least
  four consumers: `internal/config/audit_plan.go`,
  `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_worktree_root.go`,
  `internal/cli/audit_plan_cmd.go`, `internal/auditreceipt/store.go` — files from
  all three packages. Before the change (E1, E2) there were zero such files and
  zero non-test readers of `audit.model`.
- **RED-now**: E1 (no reference), E2 (no reader of the token), E2c (the old
  validator was defined and never called).
- **green** (M2-M4): `grep -rl --exclude="*_test.go" "ResolveAuditPlan(" internal` →
  stdout containing those five paths, exit 0; and AC-ACV-006's handler-level
  test passes (a reference without behaviour does not satisfy this criterion).
- **Mutant probe**: calling the resolver only from the verb leaves the
  `mcp_audit_multi.go` and `store.go` paths out of the listing and fails
  AC-ACV-006 and AC-ACV-010.

## AC-ACV-006 — `audit_multi` follows the configured plan when no gates are passed

**Covers**: maps REQ-ACV-006

- **Classification**: release-blocking (RED-now: E1, E2).
- **Given** a temp project root whose `workflow.yaml` sets `audit.model` to each
  of `codex`, `glm`, `claude`, `multi` in turn, and a recording `backendCall`
  stub, **When** `audit_multi` is called with no `gates` argument (and, in one
  row, with a `gates` object omitting the codex key), **Then** the backends the
  stub records are exactly the non-`off` entries of that token's §D.1 row
  (`codex` → [codex]; `glm` → [glm]; `claude` → [claude]; `multi` → [claude,
  codex, glm]), and the omitted-key row takes codex from the plan.
- **RED-now**: E1/E2 — with no reader of the token, every row would fan out to
  the default three.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_ConfigPlanFallback$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_ConfigPlanFallback `, exit 0.
- **Mutant probe**: a handler that ignores config invokes three backends for the
  `codex` row and fails it.

## AC-ACV-007 — explicit call arguments win over configuration

**Covers**: maps REQ-ACV-006

- **Classification**: release-blocking (RED-now: E1, E2).
- **Given** `audit.model: multi` and a call whose `gates` is
  `{codex: off, glm: off, claude: required}`, **When** `audit_multi` runs,
  **Then** codex and glm are not invoked and the result carries no codex or glm
  entry in `per_backend_verdicts`; and given `audit.gates.codex: off` with
  `model: multi` and a call that supplies `codex: required`, codex is invoked.
- **RED-now**: E1/E2.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_ArgsWinOverConfig$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_ArgsWinOverConfig `, exit 0.
- **Mutant probe**: config-over-argument precedence fails the first sub-case.

## AC-ACV-008 — no configuration and no arguments: byte-identical to before

**Covers**: maps REQ-ACV-007

- **Classification**: regression-guard (GREEN on the pre-change tree; the golden
  lands in M1's own commit before any behaviour change).
- **Given** a temp project root with no `workflow.yaml`, stubbed backends, and a
  call with no `gates`, **When** the marshalled `ConvergenceResult` (with
  `BuildCommit` and `BuildLag` zeroed) is compared with
  `internal/cli/testdata/audit_multi_default.golden.json`, **Then** they are
  byte-equal for an all-pass case and for a codex-inconclusive case (the result
  stays a pass — fail-open — and carries no `plan_source`, no `gate_unmet`).
- **RED-now**: none by design (a guard). It is red against any change that adds
  a member, flips the codex-inconclusive case to a fail, or changes the default
  gates.
- **green** (M1 on arrival; M3-M7 keep it): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_NoConfigNoArgs_ByteIdentical$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_NoConfigNoArgs_ByteIdentical `, exit 0.
- **Mutant probe**: always emitting `plan_source`, or treating the default codex
  `required` as explicit, fails the golden.

## AC-ACV-009 — a required backend that cannot answer fails the gate by name

**Covers**: maps REQ-ACV-008

- **Classification**: release-blocking (RED-now: E2 — nothing reads the token, so
  a `multi` configuration cannot make an unanswered codex gate matter).
- **Given** a temp project root whose `workflow.yaml` sets ONLY
  `audit.model: multi` (no `audit.gates` key), claude and glm stubs returning
  `pass`, and a codex stub returning `inconclusive` (one row each for binary
  missing, key missing, quota refusal, timeout, unauthenticated, malformed — all
  `inconclusive` at the seam), **When** `audit_multi` runs, **Then** each row's
  result has `overall_verdict: "fail"`, `gate_unmet: "codex"`, a
  `residual_risk_note` containing `codex`, the codex `per_backend_verdicts` entry
  still `inconclusive` and listed in `fail_open_backends`, and no pass built from
  claude and glm alone; and with the same stubs and NO configuration the result
  stays a pass (AC-ACV-008's case).
- **RED-now**: E2.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed$' ./internal/cli/` →
  `--- PASS: TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed ` and its
  six sub-test names, exit 0.
- **Mutant probe**: enforcement keyed only on a written `audit.gates` block
  (today's behaviour) leaves the pass in place and fails the test; a fix that
  downgrades to a claude-only pass fails it.

## AC-ACV-010 — one reading: receipts and the single-backend gate follow the token

**Covers**: maps REQ-ACV-005, REQ-ACV-009

- **Classification**: release-blocking (RED-now: E13 — the receipt store does not
  import the resolver; E2).
- **Given** trees whose `workflow.yaml` sets `audit.model` to `multi`, `codex`,
  `claude`, `glm`, empty, and `multi` with `audit.gates.codex: advisory`,
  **When** the receipt predicate `CodexGateRequired` is evaluated, **Then** it is
  true for `multi` and `codex`, false for `claude`, `glm`, empty, and the
  `advisory` override; and given `model: multi` with codex `inconclusive`, the
  `audit_multi` result carries a non-empty `audit_receipt` even though
  `gate_unmet` is `codex`, and `codex_audit` with the same missing codex returns
  `verdict: fail` with a non-empty `gate_unmet`.
- **RED-now**: E13 (store imports no config), E2.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestCodexGateRequired_FromModelToken|TestAuditMulti_ModelMultiRecordsReceiptOnUnmetGate|TestCodexAudit_ModelCodexUnmetGateFails)$' ./internal/auditreceipt/ ./internal/cli/` →
  three `--- PASS:` lines, exit 0; and `grep -c "internal/config" internal/auditreceipt/store.go` →
  a count of at least `1`, exit 0.
- **Mutant probe**: leaving `rawCodexGate` on the raw `audit.gates` key makes the
  `multi` row false and fails the first test.

## AC-ACV-011 — `moai audit plan` prints the plan and does nothing else

**Covers**: maps REQ-ACV-010, REQ-ACV-011, REQ-ACV-012

- **Classification**: release-blocking (RED-now: E3).
- **Given** fixture trees (no yaml; `model: multi`; `model: codex` with
  `gates.glm: advisory`; `model: grok`; a config-orphaned worktree), **When**
  `moai audit plan` runs in each (and once with `--project-root`), **Then** it
  prints JSON with the members of design.md §D.5 and exits 0 (no yaml:
  `config_status: "absent"`, `model_source: "default"`, every `explicit: false`);
  invalid configuration prints nothing on stdout, the REQ-ACV-003 text on
  stderr, exit 1; and across every run the tree's `.moai/state` directory is
  byte-identical before and after, no receipt file exists, and the `backendCall`
  seam recorded zero calls.
- **RED-now**: E3 — `Unknown command "audit"` (exit 1).
- **green** (M4): `<SCRUB>go test -count=1 -v -run '^(TestAuditPlanCmd_PrintsPlan|TestAuditPlanCmd_InvalidConfigExitsOne|TestAuditPlanCmd_WritesNothing|TestAuditPlanCmd_ProjectRoot|TestAuditPlanCmd_NoAskUserQuestion)$' ./internal/cli/` →
  five `--- PASS:` lines, exit 0; and `moai audit plan --help` on the built binary →
  usage text, exit 0.
- **Mutant probe**: a verb that calls `audit_multi` or writes a receipt fails
  `WritesNothing`; one that prints a plan for an invalid token fails the exit-1
  assertion.

## AC-ACV-012 — the auditors follow the measured plan, not prose

**Covers**: maps REQ-ACV-013

- **Classification**: release-blocking (RED-now: E4, E4t, E4x, E4d; positive
  control E4c).
- **Given** the six agent copies (`plan-auditor.md`, `sync-auditor.md`, local and
  template — four files — plus the emitted `.codex` pair) and the cross-model
  skill (local and template), **When** they are scanned after M5, **Then** each
  `.md` agent and each skill copy contains `moai audit plan`; the old block
  `Single-backend audit mode (per the project` is absent from all four agent
  copies; the two template `.codex` role TOMLs, regenerated by `make agents-emit`,
  contain `moai audit plan`; and `TestClaudeAuditTemplateSurfacesAndCatalogHash`
  and the new doc-surface test pass (identical audit section between local and
  template agents; catalogue hash current).
- **RED-now**: E4, E4t, E4x (zero), E4d (one each).
- **green** (M5): `grep -c "moai audit plan" .claude/agents/moai/plan-auditor.md …` →
  every count at least `1`, exit 0; `grep -cF "Single-backend audit mode (per the project" …` →
  every count `0`, exit 1; and `<SCRUB>go test -count=1 -v -run '^(TestClaudeAuditTemplateSurfacesAndCatalogHash|TestAuditPlanDocSurface|TestTemplateNoInternalContentLeak)$' ./internal/template/` →
  three `--- PASS:` lines, exit 0 (the leak test proves the added template text
  carries no internal SPEC or card identifier).
- **Mutant probe**: editing only the local copies fails the parity and
  doc-surface tests; editing the `.md` without `make agents-emit` leaves E4x at
  zero and fails the TOML check.

## AC-ACV-013 — an unmet gate or an unreachable surface is a named Gap, not a PASS

**Covers**: maps REQ-ACV-014

- **Classification**: release-blocking (RED-now: E4b; positive control E4c).
- **Given** the two auditor agents and the cross-model skill after M5, **When**
  scanned, **Then** each contains the literal `plan surface unreachable` and the
  instruction to cite `audit_receipt` in the `AUDIT-VERDICT` line on an unmet
  gate, and states that an unmet gate is reported in Gaps and Residual-risk and
  does not yield PASS; the existing `AUDIT-VERDICT` parser still accepts a
  `FAIL` line carrying a receipt id.
- **RED-now**: E4b — the literal is absent in all three files.
- **green** (M5): `grep -c "plan surface unreachable" .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .claude/skills/moai-ref-cross-model-audit/SKILL.md` →
  every count at least `1`, exit 0; and the `TestAuditPlanDocSurface` run of
  AC-ACV-012 asserts the same literal in the template copies.
- **Mutant probe**: wording the Gap without the fixed literal fails the grep
  (the literal is the test hook).

## AC-ACV-014 — the 4-dimension verdict is binding only together with a passing `audit_multi`

**Covers**: maps REQ-ACV-015, REQ-ACV-021

- **Classification**: release-blocking (RED-now: E11).
- **Given** a clean `FourDimVerdict` (verdict PASS, no zero scores, no missing,
  no findings), **When** `IsBinding()` is evaluated with (a) no cross-model
  requirement, (b) the requirement and an `audit_multi` outcome `pass` with no
  unmet backend, (c) the requirement and outcome `pass` with `gate_unmet`
  `codex`, (d) the requirement and outcome `fail`, (e) the requirement and no
  outcome (the call unavailable), **Then** (a) and (b) return `(true, "")`; (c)
  returns `(false, reason)` with `reason` containing `cross-model` and `codex`;
  (d) returns `(false, reason)` containing `cross-model` and `fail`; (e) returns
  `(false, reason)` containing `cross-model` and `audit_multi unreachable`; and
  with the requirement set a four-dimension verdict that already trips an
  existing trigger (INCOMPLETE, a zero score, a critical finding) still returns
  the existing reason, not a `cross-model` one — the existing triggers are
  unchanged and the cold-auditor fallback keeps its meaning.
- **RED-now**: E11 — the predicate has no cross-model input.
- **green** (M5): `<SCRUB>go test -count=1 -v -run '^(TestIsBinding_CrossModelBothMustPass|TestIsBinding_CleanPassStillBinding|TestIsBinding_ExistingTriggersUnchanged)$' ./internal/runtime/` →
  three `--- PASS:` lines (the first with its five sub-test names), exit 0; the
  existing `TestIsBinding*` tests in `sync_4dim_binding_test.go` still pass.
- **Mutant probe**: a predicate that ignores the cross-model fields fails (c)-(e);
  one that returns not-binding whenever the requirement is set, even with a
  passing outcome, fails (b); one that lets a `cross-model` reason replace an
  existing trigger's reason fails the last clause.

## AC-ACV-015 — the stale deferral is gone and its guard is not vacuous

**Covers**: maps REQ-ACV-016

- **Classification**: release-blocking (RED-now: E5, E6, E12).
- **Given** the run-phase tree, **When** it is scanned, **Then**
  `multiConvergenceImplemented` appears nowhere in Go source or tests; the string
  `AP-8` appears in no non-test Go file; the web guard's sentinel is
  `ResolveAuditPlan` and the guard fails when the sentinel is absent from the
  non-test source of `internal/config`.
- **RED-now**: E5 (five hits), E6 (three hits), E12 (sentinel is
  `activeAuditBackend`).
- **green** (M6): `grep -rn "multiConvergenceImplemented" --include="*.go" internal` →
  (no output), exit 1; `grep -rn --exclude="*_test.go" "AP-8" internal` → (no
  output), exit 1; `grep -n "const sentinel" internal/web/mcp_audit_surface_test.go` →
  a line naming `ResolveAuditPlan`, exit 0; and `<SCRUB>go test -count=1 -v -run '^(TestWebConsole_AuditNoForkedInterpreter|TestWebConsole_AuditSentinelExists)$' ./internal/web/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: deleting `activeAuditBackend` without retargeting the
  sentinel leaves the guard green and vacuous — the positive-control test
  `TestWebConsole_AuditSentinelExists` is the part that fails when the sentinel
  names no live symbol.

## AC-ACV-016 — the committed yaml opts in and the pin is aligned; the template is untouched

**Covers**: maps REQ-ACV-017

- **Classification**: release-blocking (RED-now: E7, E8, E9; template guard E10,
  E16).
- **Given** the run-phase tree, **When** the committed
  `.moai/config/sections/workflow.yaml` is read, **Then** the line after
  `model: claude-opus-5-5` is `effort: high`, the key `model: multi` sits under
  `audit:`, the header comment no longer says `claude-opus-5-5/medium`; the
  committed file loads through the config loader to `Audit.Model == "multi"` and
  `Audit.Claude.Effort == DefaultClaudeAuditEffort`; the template yaml still has
  no `model: multi` and its claude pin is still `high`; the Go default
  `Audit.Model` is still `claude`.
- **RED-now**: E7 (`medium`), E8 (no key), E9 (comment says medium).
- **green** (M7): `grep -n -A1 "model: claude-opus-5-5" .moai/config/sections/workflow.yaml` →
  `…effort: high`, exit 0; `grep -nE "^        model: multi$" .moai/config/sections/workflow.yaml` →
  one line, exit 0; `grep -n "claude-opus-5-5/medium" .moai/config/sections/workflow.yaml` →
  (no output), exit 1; `grep -c "model: multi" internal/template/templates/.moai/config/sections/workflow.yaml` →
  `0`, exit 1 (E16 unchanged); and
  `<SCRUB>go test -count=1 -v -run '^(TestCommittedWorkflowYamlAuditAlignment|TestAuditConfig_DefaultProfile)$' ./internal/config/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: setting only the pin and not the token fails the
  `grep -nE` and the loader test; adding `model: multi` to the template fails
  the `grep -c` guard.

## AC-ACV-017 — the fan-out stays concurrent under a config plan

**Covers**: maps REQ-ACV-018

- **Classification**: regression-guard (the default plan already activates all
  three legs concurrently, so the test is GREEN against the pre-change tree and
  must stay GREEN).
- **Given** `audit.model: multi` and three stub backends that each sleep a fixed
  interval, **When** `audit_multi` runs, **Then** all three are invoked, the
  elapsed time is below the sum of the three intervals, and the resolver
  activated no backend outside `claude`, `codex`, `glm`.
- **RED-now**: none by design.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_ConfigPlanKeepsParallelFanOut|TestRunMultiAudit_ParallelFanOut_AC_AMM_002)$' ./internal/cli/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: a serialized fan-out (legs run one after another) exceeds the
  bound and fails the first test.

## AC-ACV-018 — the committed `multi` does not leak into existing audit tests

**Covers**: maps REQ-ACV-007, REQ-ACV-017

- **Classification**: regression-guard (GREEN before M7; its job is to stay GREEN
  after the yaml flip, in a session whose environment points at the real tree).
- **Given** M7 has landed (the committed yaml says `model: multi`), **When** the
  existing audit tests run with `CLAUDE_PROJECT_DIR` set to this tree's root,
  **Then** they all pass — no test reads the committed yaml unintentionally.
- **RED-now**: none by design; a non-hermetic test turns RED only after M7.
- **green** (M7): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && CLAUDE_PROJECT_DIR=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1423 go test -count=1 -v -run '^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestConverge_.*|TestCodexAudit_.*|TestPerformCodexAudit_.*|TestPerformGLMAudit_.*|TestDefaultBackendCaller_.*|TestPersistConvergenceResult_.*)$' ./internal/cli/` →
  ends `ok  	github.com/modu-ai/moai-adk/internal/cli`, no `--- FAIL`, and the output
  lists the `--- PASS:` lines of at least `TestAuditMulti_RespectsCodexGateOff_AC_AMM_014`,
  `TestAuditMulti_NoHardErrorPath_AC_AMM_024` and
  `TestRunMultiAudit_ParallelFanOut_AC_AMM_002` (an empty swept set prints `ok`
  too — read the names).
- **Mutant probe**: removing M1's hermetic root from one audit test makes that
  test red here after M7 (the intended detection).

## AC-ACV-019 — a GLM- or GPT-launched lane still gets an independent Claude verdict

**Covers**: maps REQ-ACV-019

- **Classification**: regression-guard (the existing behaviour, held under the new
  plan).
- **Given** launch provider `glm` (then `gpt`), `audit.model: multi`, a supplied
  `claude_verdict` of `pass`, and stubs where the Claude backend returns `fail`,
  **When** `audit_multi` runs, **Then** the supplied verdict is ignored, the
  Claude backend is invoked, the result's claude entry has
  `source: "mcp_claude_audit"` and verdict `fail`, and the overall verdict is
  `fail`; and with the Claude backend `inconclusive` under `multi` the result
  fails by name (`gate_unmet` contains `claude`).
- **RED-now**: none by design.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_GPTAndGLMOriginsRunActualClaude_AC_CLA_009_010|TestAuditMulti_GLMOriginUnderMultiPlan)$' ./internal/cli/` →
  the existing test's sub-tests and `--- PASS: TestAuditMulti_GLMOriginUnderMultiPlan `,
  exit 0.
- **Mutant probe**: a plan path that lets a supplied anchor stand in for the
  backend under a non-Claude origin fails the second assertion.

## AC-ACV-020 — the codex leg has a deadline, and its expiry is an unmet gate naming codex

**Covers**: maps REQ-ACV-020, REQ-ACV-008, REQ-ACV-007

- **Classification**: release-blocking (RED-now: E17, E18; positive controls E17c,
  E18c).
- **Given** a temp project root whose `workflow.yaml` sets only
  `audit.model: multi`, claude and glm stubs returning `pass`, a `codexReviewRPC`
  stub that blocks until its context ends, and the codex-leg limit shortened
  through its test seam to a few milliseconds, **When** `audit_multi` runs,
  **Then** it returns within a bound well under one second; the codex
  `per_backend_verdicts` entry is `inconclusive` with a summary naming a codex
  timeout; `overall_verdict` is `fail`, `gate_unmet` is `codex`, and
  `residual_risk_note` names codex; and in a second sub-case where the stub
  returns a `pass` only after the deadline, the entry is still `inconclusive`
  (partial or late output is never read as a verdict); and with NO configuration
  the same hung stub yields the fail-open result (codex `inconclusive`, overall
  not failed) — the deadline changes the unconfigured case only by ending a
  hang that previously never ended.
- **RED-now**: E17 (no constant), E18 (no `context.WithTimeout` around the codex
  leg) — the leg currently has no deadline, so the hung stub never returns.
- **green** (M3): `<SCRUB>go test -count=1 -v -run '^(TestAuditMulti_CodexLegDeadline_FailsClosedNamed|TestAuditMulti_CodexLegDeadline_LateOutputIgnored|TestAuditMulti_CodexLegDeadline_UnconfiguredFailsOpen)$' ./internal/cli/` →
  three `--- PASS:` lines, exit 0; `grep -rn "codexAuditTimeout" internal` →
  the constant's definition and its use, exit 0; and
  `grep -n "claudeAuditTimeout" internal/cli/mcp_codex.go` → the line deriving the
  codex limit from the Claude limit, exit 0 (the derivation of design.md §D.10,
  not a restated number).
- **Mutant probe**: a leg with no deadline never returns and the test times out
  red; a deadline that still reads the partial review text fails the
  late-output sub-case; a deadline that maps expiry to `pass` or drops the entry
  fails the named-gate assertions; a hard-coded `5 * time.Minute` literal fails
  the derivation grep.

## AC-ACV-021 — where the sync call lives, what the record shows, and that the script stays read-only

**Covers**: maps REQ-ACV-015, REQ-ACV-021

- **Classification**: release-blocking (RED-now: E14, E19, E20; baseline E21).
- **Given** both copies of `workflows/sync.md` and of `sync-audit-4dim.js` after
  M5, **When** they are scanned, **Then** each `sync.md` copy names
  `cross_model_required`, `moai audit plan`, the post-PASS `audit_multi` call,
  the "both must pass" rule, the `cross_model:` / `audit_multi:` / `gate_unmet:`
  / `audit_receipt:` lines of the binding statement, and the literal
  `audit_multi unreachable`; each script copy names `audit_multi` in its header
  comment and `meta.description` as the orchestrator's call, and still contains
  no `mcp__moai__` (E21 stays 0); the script's phases, schemas and verdict
  computation are unchanged (the existing sync-audit-4dim script tests, if any,
  and the `TestIsBinding*` tests pass); and the template copies carry no internal
  SPEC or card identifier.
- **RED-now**: E14 and E19 (zero in both `sync.md` copies), E20 (zero in both
  script copies).
- **green** (M5): `grep -c "cross_model_required" …` and
  `grep -c "audit_multi unreachable" …` over both `sync.md` copies → each at
  least `1`, exit 0; `grep -c "audit_multi" .claude/workflows/sync-audit-4dim.js internal/template/templates/.claude/workflows/sync-audit-4dim.js` →
  each at least `1`, exit 0; `grep -c "mcp__moai__" …` over both script copies →
  `0`, exit 1; and `<SCRUB>go test -count=1 -v -run '^(TestAuditPlanDocSurface|TestTemplateNoInternalContentLeak)$' ./internal/template/` →
  two `--- PASS:` lines, exit 0.
- **Mutant probe**: moving the call into the script adds `mcp__moai__` and fails
  the zero-count grep; editing only the local copies fails the doc-surface test;
  a bare four-dimension PASS recorded on an unmet gate lacks the
  `audit_multi unreachable` / `gate_unmet` literals and fails the grep.

## Edge cases

- **EC-1 Matching.** Token and gate values match exactly and case-sensitively
  after trimming surrounding whitespace (AC-ACV-003).
- **EC-2 Config-orphaned worktree with no identifiable primary.** The existing
  rule stands: codex is assumed `required` with its note, so the gate is
  fail-closed (`resolveAuditGates`, unchanged).
- **EC-3 All-off plan.** `audit.model: codex` with `audit.gates.codex: off`
  leaves every backend off; the verb prints `cross_model_active: false` and no
  `enforced_required` entry, and `audit_multi` keeps its existing result for a
  fan-out with no participants (this SPEC does not change it).
- **EC-4 Unreadable yaml.** The verb prints the default plan with
  `config_status: "unreadable"` and a note; `audit_multi` keeps today's fail-open
  reading (spec.md §D, R-4).
- **EC-6 Deadline versus caller cancellation.** A leg ended by the caller's own
  cancellation (not by the deadline) is not reported as a codex timeout; it is
  reported as the cancellation it is, still `inconclusive` (AC-ACV-020 covers
  the deadline arm only).
- **EC-5 Old build.** `moai audit plan` is an unknown command; the auditor's
  `plan surface unreachable` Gap applies (AC-ACV-013).

## Quality gates and Definition of Done

1. All 21 criteria pass; each release-blocking criterion's RED-now was observed
   red on `c50da9c2f` (ledger) and the named Go test's own RED output is in
   `progress.md §E.2`.
2. `moai spec lint SPEC-AUDIT-MODEL-CONVERGE-001` exits 0; `go build ./...` and
   `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run` shows
   no NEW finding against the pre-run baseline; `make build` exits 0 (agents-emit
   check, tool-policy drift check, catalogue hashes).
3. Coverage of the new files `internal/config/audit_plan.go` and
   `internal/cli/audit_plan_cmd.go` is at least 85%
   (a coverage profile of the criteria's named tests, read per file with
   `go tool cover -func`; a package-level percentage does not satisfy this item).
4. The distributed template `workflow.yaml`, the Go default `Audit.Model`, the
   MCP catalogue, and the pre-existing local-vs-template differences of
   research.md R-15 are unchanged.
5. TRUST 5: Tested (the criteria above); Readable (English comments, named
   constants, no inline env names); Unified (gofmt, golangci-lint); Secured (no
   secret in output — the verb prints gates, never keys); Trackable (commit
   subjects carry `SPEC-AUDIT-MODEL-CONVERGE-001` and `t1423`).
6. The sync report names the stale `sync-audit-4dim.js` header (plan.md AP-7) and
   the operator-owned rc install (the behaviour is inert in this repository
   until a build carrying it is installed and the MCP server reconnects).
