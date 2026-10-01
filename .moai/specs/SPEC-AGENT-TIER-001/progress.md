# SPEC-AGENT-TIER-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).

- 14 GEARS requirements (`REQ-TIER-001..014`); 15 acceptance criteria (`AC-TIER-001..015`),
  13 release-blocking + 2 regression-guard.
- Methodology: tdd (tier constants are new code with a RED-first test,
  `TestClaudeTierConstants`).
- Milestones: M1 tier constants + audit pin defaults (High) → M2 profile matrix wiring (High)
  → M3 moai web tier widget (Medium). Ordered by decision-reversibility; M2 carries the one
  open in-constraint design decision (spawn-time injection channel; candidates + recommended
  default recorded in plan.md §F M2, to be decided and logged at M2 start).
- Operator decisions carried verbatim: audit pins are operator-fixed defaults (claude
  {claude-opus-5-5, high} · codex {gpt-6.1-sol, high} · glm {glm-5.3, max}); audit surfaces
  excluded from the tier matrix; tier pairs max {sonnet-5-5, max} / medium {sonnet-5-5, high} /
  low {sonnet-5-5, medium}; fallback {claude-opus-5-5, xhigh} recorded, not automated.
- Q2 terminology: adopted as recommended (tier tokens are configuration-key names, not effort
  values). One naming deviation recorded at spec.md REQ-TIER-002: constants keyed by TIER token
  (DefaultClaudeTierMax/Medium/Low), not by effort-slot suffix — the suggested
  Max/High/Medium suffixes would alias tier-medium to a `...TierHigh` constant and recreate the
  exact ambiguity Q2 prevents. Values are byte-exact to the operator table; no value deviation.
- Tree findings vs the delegation brief: pin block measured at `defaults.go:1218-1238` (brief
  said `:1202`); claude pin literals at `:1221-1222`; GLM pin currently absent from the Go
  default (ships empty in the template); the per-agent profile matrix is retired by
  SPEC-AGENT-MODEL-INHERIT-001, so the tier axis is a bounded partial supersession (spec.md
  §A.1).
- RED-now observations: 5 cells observed on the pre-implementation tree at
  `f130aa041bb90c81b235029afffa806a65e88480` (acceptance.md Evidence Ledger, EVID-TIER-A..E);
  4 planned run-start cells (R1..R4) to be executed and pinned before the first M1 edit.
- Grounding: `.moai/reports/agent-tier-design-20261001.md` (untracked, primary checkout);
  SPEC-critical figures carried in spec.md §E.

plan_status: audit-ready
plan_complete_at: 2026-10-01

## §E.2 Run-phase Evidence

Run window 2026-10-01 → 2026-10-02, card worktree `WT-agent-tier-max`. Run-start SHA
`01022d8cd`; M1 `a4d24cf89`; M2 `44116a924`; M3 `a0bd604da` (final HEAD for the verification
batch). Every verification below ran in this run, in this tree.

### M2 design decisions (recorded at M2 start, per plan.md §F)

1. **Config surface: option (a) — `workflow.agent_tiers` block in workflow.yaml** (the
   recommended default). Reason: the tier axis is workflow-scoped delegation policy; the audit
   pins it must stay disjoint from already live in `workflow.audit` of the same section file;
   the loader `Validate()` convention and the shipped-key anti-rot inventory
   (`internal/config/testdata/shipped_key_inventory.yaml`) key on this section, so rejection
   semantics (REQ-TIER-009) and key-honesty guards come from the existing machinery. The
   alternative (an llm.yaml key) was rejected — llm.yaml is gitignored and wiped by
   `moai update`, so a tier table there would be non-durable.
2. **Spawn-time injection channel: candidate (a) — launcher/session-env injection via the
   `CLAUDE_CODE_SUBAGENT_MODEL` chain.** Implementation evidence: the codebase already treats
   `CLAUDE_CODE_SUBAGENT_MODEL` as a session-launch env token (it sits in the claude audit
   scrub key set, `internal/cli/mcp_claude.go` `claudeAuditScrubKeySet`); the declaration-only
   spawn observer (`internal/hook/agent_model_guard.go`) classifies env-declared spawns as
   declared, so the observer's declared/inherit semantics survive unchanged; and the settings
   field `effortLevel` rejects `max`, so a resolved max must ride the launcher launch-argument
   path (coding-standards compatibility table) — launcher-side, consistent with (a). (b)
   delegation-layer spawn-argument injection re-introduces hand-passing doctrine pressure;
   (c) frontmatter `tier:` is the per-agent-file channel SPEC-AGENT-MODEL-INHERIT-001 removed.
   **Boundary recorded:** this run lands the config surface + resolver + validation + web
   surface (the ACs' subject); the launcher-side consumption of the recorded channel is
   downstream wiring no AC in this SPEC gates.

### Run-start RED cells (tree `01022d8cd`, observed 2026-10-01, before any M1 edit)

**R1 — AC-TIER-001/008** (`go test ./internal/config/ -run '^TestClaudeTierConstants$'`, after
the test was authored; the compile failure names exactly the symbols M1 creates):

```
# github.com/modu-ai/moai-adk/internal/config [github.com/modu-ai/moai-adk/internal/config.test]
internal/config/claude_tier_constants_test.go:17:5: undefined: DefaultClaudeTierMax
internal/config/claude_tier_constants_test.go:18:65: undefined: DefaultClaudeTierMax
internal/config/claude_tier_constants_test.go:20:5: undefined: DefaultClaudeTierMedium
internal/config/claude_tier_constants_test.go:21:69: undefined: DefaultClaudeTierMedium
internal/config/claude_tier_constants_test.go:23:5: undefined: DefaultClaudeTierLow
internal/config/claude_tier_constants_test.go:24:68: undefined: DefaultClaudeTierLow
internal/config/claude_tier_constants_test.go:35:5: undefined: DefaultClaudeTierMaxFallback
internal/config/claude_tier_constants_test.go:36:80: undefined: DefaultClaudeTierMaxFallback
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
FAIL
```
exit 1. RED for the right reason: the symbols do not exist anywhere in the package (EVID-TIER-B).

**AC-TIER-002 Go side** (`go test ./internal/config/ -run
'^TestDefaultWorkflowConfig_AuditPinsOperatorTable$'`):

```
--- FAIL: TestDefaultWorkflowConfig_AuditPinsOperatorTable (0.00s)
    audit_pin_defaults_test.go:18: default Claude pin = {claude-opus-5-5 medium}, want {claude-opus-5-5 high}
    audit_pin_defaults_test.go:24: default GLM pin = { }, want {glm-5.3 max} (non-empty per REQ-TIER-006)
FAIL
```
exit 1. Corroborates EVID-TIER-A at the Go level.

**R2 — AC-TIER-004/015** (`go test ./internal/cli/ -run
'^TestAuditResolverTerminalFallbacks_OperatorTable$'`):

```
--- FAIL: TestAuditResolverTerminalFallbacks_OperatorTable (1.14s)
    audit_pin_resolver_defaults_test.go:27: claude terminal fallback = {claude-opus-5-5 medium}, want {claude-opus-5-5 high}
    audit_pin_resolver_defaults_test.go:32: glm terminal fallback = {glm-5.3-flash }, want {glm-5.3 max}
FAIL
```
exit 1. RED for the right reason: the fallbacks bound the t1368 values EVID-TIER-A/EVID-TIER-F
observed.

**R3 — AC-TIER-005** (sweep test, `go test ./internal/config/ -run '^TestPinLiteralSweep'
-count=1 -v`): the positive control PASSED on the first authored version after exposing a real
defect in the sweep's own first cut (a root-relative suffix bug that matched nothing — fixed
before any green was read; the control proved the detector fires). The real-tree run was RED on
exactly the two pin-literal restatements M1 removes:

```
--- PASS: TestPinLiteralSweep_PositiveControl (0.01s)
=== RUN   TestPinLiteralSweep_RealTreeClean
    pin_literal_sweep_test.go:147: pin literal sweep: swept 1554 files under internal/ (excl. tests, template mirror, documented non-pin axes)
    pin_literal_sweep_test.go:152: pin literals found outside the declared single-source locations (14):
    pin_literal_sweep_test.go:154:   cli/mcp_claude.go carries pin literal claude-opus-5-5
    pin_literal_sweep_test.go:154:   cli/mcp_codex.go carries pin literal gpt-6.1-sol
    ... (12 further hits: internal/config package literals + comments on non-pin axes, re-classified into the exclusion set with documented reasons)
--- FAIL: TestPinLiteralSweep_RealTreeClean
```
exit 1. After the exclusion-set classification (`internal/config/` package = the declared
single-source home per REQ-TIER-013's own wording; `template/model_policy.go`,
`statusline/memory.go`, `cli/glm.go` = documented non-pin axes), the final cut measured RED on
exactly the two real restatements, 1339 files swept, positive control green.

**R4 — AC-TIER-006/007** (`go test ./internal/config/ -run '^TestAgentTiers_RejectsUnknownToken$'`
at M2 start, before the tier surface existed):

```
internal/config/agent_tiers_test.go:48:24: undefined: ValidAgentTiers
internal/config/agent_tiers_test.go:61:22: undefined: AgentTierMax
... (ValidAgentTiers, AgentTierMax/Medium/Low, ResolveAgentClassTier — all undefined)
```
exit 1 (build failed). RED for the right reason: no tier surface exists at M2 start.

### GREEN evidence (per AC; final verification at HEAD `a0bd604da` unless noted)

- **AC-TIER-001** `go test ./internal/config/ -run '^TestClaudeTierConstants$|^TestClaudeTierMaxFallbackRecorded$' -count=1 -v`:
  `--- PASS: TestClaudeTierConstants` / `--- PASS: TestClaudeTierMaxFallbackRecorded` / `ok github.com/modu-ai/moai-adk/internal/config 0.292s`. exit 0.
- **AC-TIER-002** `TestDefaultWorkflowConfig_AuditPinsOperatorTable` PASS (same command, `-run '…PinsOperatorTable$'`): asserts {claude-opus-5-5, high} · {gpt-6.1-sol, high} · {glm-5.3, max} byte-exact. `grep -n -A 3 "Claude: ModelEffort{" internal/config/defaults.go` now shows the block at `defaults.go:1269-1272` deriving `DefaultClaudeAuditModel`/`DefaultClaudeAuditEffort`; the GLM cell sits at `:1290-1293`.
- **AC-TIER-003** template block post-edit: `effort: high` (:115), codex `gpt-6.1-sol/high` (:117-118), `glm-5.3/max` (:120-121); `make build` re-embedded (exit 0), `make embed-check` = Pass 1 Warn 0 Fail 0, and `TestClaudeAuditTemplateSurfacesAndCatalogHash` (reads the embedded copy) passes in the template suite.
- **AC-TIER-004** `go test ./internal/cli/ -run '^TestAuditResolverTerminalFallbacks_OperatorTable$' -count=1`: `ok github.com/modu-ai/moai-adk/internal/cli 3.318s`. exit 0.
- **AC-TIER-005** `go test ./internal/config/ -run '^TestPinLiteralSweep' -count=1 -v`: `--- PASS: TestPinLiteralSweep_PositiveControl` + `pin literal sweep: swept 1339 files under internal/` + `--- PASS: TestPinLiteralSweep_RealTreeClean`. The sweep is a permanent package test (fires on every `go test ./internal/config/`).
- **AC-TIER-006** `TestAgentTiers_RejectsUnknownToken` PASS: `extreme` fails the load with an error naming the token, and the loader wrap names `workflow.yaml` (`workflow.yaml: workflow.agent_tiers.classes[manager-develop] = "extreme" invalid: want one of max|medium|low`); `max`/`medium`/`low` fixtures load cleanly.
- **AC-TIER-007** `TestDefaultAgentTierAssignment` PASS: super-advisor→max, manager-spec→max, manager-develop→medium, manager-docs→medium, e2e-tester→medium, explore→low, lane→medium; plan-auditor, sync-auditor, audit-claude/codex/glm → "" (no tier). `TestAgentTiers_UserOverrideWins` PASS (project override per class).
- **AC-TIER-008** `TestClaudeTierMaxFallbackRecorded` PASS; the constant's doc comment carries the chart grounding (65% @ ~$5, availability/dispersion alternative) and the record-only clause.
- **AC-TIER-009** targeted diff empty: `git diff --stat -- internal/template/templates/.moai/config/sections/llm.yaml internal/cli/launch_effort_settings.go internal/template/glm_effort_overlay.go internal/config/envkeys.go` → empty (byte-unchanged). The workflow.yaml effort-line diff carries ONLY the sanctioned audit-pin flips (`medium→high`, `""→max`) and the new agent_tiers comment; no tier token occupies any effort slot anywhere. Q2 definition documented on the tier constants' comment block (defaults.go).
- **AC-TIER-010** `go test ./internal/web/ -run '^TestAgentTiersSection' -count=1 -v`: `--- PASS: TestAgentTiersSection_ChartGrounding` (tier keys max/medium/low rendered as code chips + figures 70.6% / ~$11 / 45% / ~$2.3 / 29% / ~$0.8) and `--- PASS: TestAgentTiersSection_SelectionClosedSet` (each class's radio group offers exactly {max, medium, low}, no value outside the closed set).
- **AC-TIER-012** `grep -rn "^model:\|^effort:\|^tier:"` over C1/C2 agent trees and the toml keys over C3: empty; `git diff --stat 01022d8cd..HEAD -- internal/hook/ .claude/agents/ internal/template/templates/.claude/agents/ internal/template/templates/.codex/agents/` → empty; `make agents-emit-check` exit 0; `go test ./internal/hook/ -run 'AgentModel|ModelGuard' -count=1` → `ok internal/hook 0.752s`.
- **AC-TIER-014** plan-auditor PASS 0.91 ≥ 0.80 (plan-phase record; §E.1 / `.moai/reports/t1391/plan-audit.md`).
- **AC-TIER-015** the resolver test GREEN (model `DefaultGLM53`, effort `max` forwarded verbatim); the sweep's green proves no inline `glm-5.3` literal remains anywhere in the audit path (mcp_glm.go carries zero pin literals — `glmAuditDefaultModel = config.DefaultGLMAuditModel`, `glmAuditDefaultEffort = config.DefaultGLMAuditEffort`). RED-now: EVID-TIER-F (the t1368 `DefaultGLMHigh` binding).

### AC-TIER-011 — user pin precedence preserved (regression-guard)

Baseline (pre-edit binaries compiled at launch, before any edit): `go test -timeout 30m
./internal/config/... ./internal/cli/... ./internal/web/...` → config ok ×3; cli root FAIL with
4 failures (`TestStopChainEffectParityGolden`, `TestStopChainMemberCostWithinBudget`,
`TestSyncGateLanguageDetectionMatchesScript`, `TestCodexTaskBackgroundHandshakeHonorsTaskBound`);
all cli subpackages ok; web ok 146.1s.

Final (post-M3): cli root FAIL with exactly 2 — `TestStopChainEffectParityGolden` (57.63s) and
`TestSyncGateLanguageDetectionMatchesScript` (1.79s). BOTH are in the pre-edit baseline and both
are known pre-existing CI red (card t1390's known-red set). The other two baseline failures did
NOT reproduce in isolation (flaky under the baseline run's parallel load; stop-chain/codex-task
domains this SPEC never touches). **NEW failures attributable to this change: 0.** User-pin
precedence tests (`TestResolveClaudeAuditModelEffort_FieldsOverrideIndependently`,
`audit_pin_test.go`, `mcp_codex_audit_pin_test.go`, `mcp_glm_audit_pin_test.go` pin cases) pass
unmodified — they reference the defaults symbolically. Tests updated in the same commit as the
default they pin: `mcp_audit_config_test.go` (`TestAuditConfig_DefaultProfile`: medium→high +
GLM additions), `mcp_glm_fallback_test.go` (3 tests re-pinned to the new audit/task default
split), `mcp_glm_test.go` (`TestResolveGLMAuditModel_BackendDefault`), `mcp_glm_audit_pin_test.go`
(absent-pin case + task-resolution assertion).

### AC-TIER-013 — verification batch (final; HEAD `a0bd604da`, env-scrubbed compound form)

| Command | Result | exit |
|---|---|---|
| `go vet ./internal/config/... ./internal/cli/... ./internal/web/...` | (no output) | 0 |
| `golangci-lint run --timeout=2m` (v2.1.6 — the CI-pinned version) | `0 issues.` | 0 |
| `go test -timeout 30m ./internal/config/... ./internal/cli/... ./internal/web/...` | config ok ×3 · cli root FAIL = 2 pre-existing (above) · cli subpackages ok ×21 · web ok 74.950s | 1 (pre-existing only) |
| `make build` | catalog regenerated + binary built | 0 |
| `make embed-check` | Pass 1 Warn 0 Fail 0 | 0 |
| `make agents-emit-check` | `ok internal/template/agentemit` | 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | (no output) | 0 |

Repo-specific guards absorbed during M2/M3 (both GREEN after the fix): the shipped-key anti-rot
inventory (`TestShippedConfigKeysHaveReaders` — 7 new W-class entries with reader evidence) and
the web schema parity guard (`TestSchemaParity_EditableFieldsHaveRenderHome` — the tier
sub-section registered as a render home, the jevSectionFields pattern).

Merge-tree pre-gate repairs (post-sync, card t1391; three roots caught by the integration-window
re-measurement that the run-phase suite verdict did not surface):

1. `c310cd4f5` + `c42d4a706` — the pin sweep tripped on an untracked worktree-local fixture
   (`internal/cli/.moai/config/sections/llm.yaml`, never tracked in any commit), which then
   exposed that `filepath.SkipAll` ENDS the whole walk: the run-phase sweep stopped at its first
   excluded directory, so `internal/web/` was never swept by the instrument. The sweep now skips
   `.moai/` directories entirely (the user-project config/state axis) and returns
   `filepath.SkipDir`; the swept-count plausibility guard caught the 30-file truncation live.
2. The glm audit default expectations in `internal/cli/model_backend_default_test.go` and
   `internal/cli/retained_model_surfaces_char_test.go` — AC-TIER-011 default-asserting updates
   the run phase missed: both asserted the former empty-effort backend default {glm-5.3-flash,
   ""} and surfaced as failures on the merge-tree suite re-measurement (the run-phase report
   listed only the 2 known pre-existing failures). Both now assert the {glm-5.3, max} default
   pin via `glmAuditDefaultModel`/`glmAuditDefaultEffort` (t1386 constant-reference precedent)
   — re-run green on the card branch (`ok internal/cli 1.121s`, the two repaired tests named
   exactly) and again on the merge tree. This record is the AC-TIER-011 update listing.

## §E.3 Run-phase Audit-Ready Signal

Milestones: M1 `a4d24cf89` (tier constants + operator audit pin defaults, RED-first R1/R2/R3
observed) → M2 `44116a924` (agent_tiers profile matrix wiring, RED-first R4 observed) → M3
`a0bd604da` (moai web tier widget). progress evidence commit follows this section.

- AC matrix: AC-TIER-001..010 PASS, AC-TIER-015 PASS (blocking); AC-TIER-013 PASS with the
  pre-existing-failure classification above (0 new); AC-TIER-014 PASS (plan-phase record);
  AC-TIER-011/012 regression-guards PASS (baseline comparison + unchanged trees).
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- Lint: v2.1.6 (CI-pinned), 0 issues — no NEW lint findings (baseline-equivalent).
- Injection channel recorded: candidate (a) — launcher/session-env via the
  `CLAUDE_CODE_SUBAGENT_MODEL` chain; launcher-side wiring is downstream (recorded above).
- Known boundaries (not defects): (i) the 2 pre-existing cli failures ride the branch into CI as
  they rode the baseline; (ii) tier chart figures live in `internal/config` (single source) and
  render via `AgentTierChartTable()`; (iii) `DefaultClaudeTierMaxFallback` is a record only —
  no failover machinery (REQ-TIER-003/§C).

run_status: audit-ready
run_complete_at: 2026-10-02

## §E.4 Sync-phase Audit-Ready Signal

sync window 2026-10-02, card worktree `WT-agent-tier-max`, sync base `f0fa3cf75` (the run-phase
final HEAD). Sync scope: CHANGELOG `[Unreleased]` `### Added` entry (top, most-recent-first
convention), this §E.4 signal, README judgment, and the `spec.md` frontmatter
`status: in-progress → implemented → completed` + `updated: 2026-10-02` transition. docs-site is
out of this SPEC's scope (follow-up suggestion recorded in the sync report).

- sync_commit_sha: 44f302a68
- sync_date: 2026-10-02
- sync_status: complete
- b12_self_test_a: PASS — pre-emission grep `grep -c 'SPEC-AGENT-TIER-001' CHANGELOG.md` = 0
  before the entry landed (no duplicate from a parallel BATCH-SYNC session).
- b12_self_test_b: PASS — acceptance.md distinct live AC identifiers = 15 (AC-TIER-001..015;
  zero `[RETIRED]`/`[REF]` markers anywhere in the file) = CHANGELOG entry count 15.
- b12_self_test_c: PASS — every file path named in the CHANGELOG entry verified present via ls
  (internal/config, internal/cli, internal/web, internal/settings, template workflow.yaml).
- README judgment: no update — no README statement describes the audit pin values or a per-agent
  tier surface; the "every agent inherits the session's model and reasoning effort" sentence
  stays true because launcher-side tier consumption is downstream wiring no AC in this SPEC
  gates, and the tier sub-section renders inside the existing Workflow tab so the settings-tab
  enumeration is unchanged.
- Verification (this run, this tree): `go test -timeout 30m ./internal/spec/... ./internal/config/...`
  env-scrubbed compound form — ok (spec/config ok, exit 0); `make build` exit 0;
  `./bin/moai spec lint` scoped to SPEC-AGENT-TIER-001 — clean (verbatim outputs carried in the
  sync report and the sync commit evidence). internal/cli is not re-run in sync (sync changes no
  Go code in it; the merge-tree re-measurement covers it).

## §F Phase 4 Mode Selection

### Mode Selection

Input parameters:

- tier: M · scope: ~8 files (internal/config, internal/cli, internal/web, template
  workflow.yaml + tests)
- domain count: 4 (Go config / Go CLI / Go web+templ / YAML template) — coding-heavy
- file language mix: Go + YAML + templ
- concurrency benefit: LOW (milestones are dependency-ordered M1→M2→M3; Anthropic
  coding-task parallelism caveat)
- Agent Teams prereqs: not requested

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file semantic change |
| serial | **YES** | coding-heavy, dependency-ordered milestones; one implementation spawn carries M1→M2→M3 |
| fanout | no | coding-heavy work — parallelism caveat; no independent research lanes |
| sweep | no | semantic new-code work, ~8 files, not mechanical-uniform |
| agent-team | no | not operator-requested |

Decision: serial

Justification: coding-heavy single-tree work with strict milestone dependencies (M1's
constants feed M2's wiring and M3's figures), so the Anthropic coding-task parallelism
caveat puts this on the sequential path. One implementation delegate keeps exactly one
writer on the tree. The delegate is dispatched as a general-purpose worker performing the
manager-develop role — the typed manager-develop spawn auto-isolates to its own L1 tree and
refuses card-tree writes (card t1318 lesson), and the card worktree is the required landing
site. Boundary Case: none hit — scope and domain counts sit below the fanout thresholds.

### Implementation Kickoff decision record

- Gate form: **operator direct answer** (keep-set class: operator-held decisions). The
  operator answered the Kickoff panel directly, relayed by the leader (2026-10-01).
- Audit cross: plan-auditor PASS 0.91 ≥ 0.80 (Tier M) — report
  `.moai/reports/t1391/plan-audit.md` (auditor model glm-5.3-flash; GLM lanes cannot emit
  Opus audits — lane lesson, checked against the report's first line);
  `plan_status: audit-ready` recorded in §E.1; no open blocker.
- Operator decisions recorded verbatim:
  1. Q2 approved: tier tokens are configuration-key names; the effort field keeps the
     chart-axis vocabulary (low/medium/high/xhigh/max).
  2. Constant-naming deviation approved: `DefaultClaudeTierMax` / `DefaultClaudeTierMedium`
     / `DefaultClaudeTierLow`.
  3. Audit pins fixed: claude {claude-opus-5-5, high} · codex {gpt-6.1-sol, high} · glm
     {glm-5.3, max}; the 3-tier matrix applies to the general profile surfaces only — audit
     surfaces excluded.
  4. Minor amendments D1 + D2 approved and ordered applied BEFORE run entry ("D1·D2 반영 후
     run 진입", Kickoff record first).
- D1 (selector anchoring): the unanchored `-run TestClaudeTierConstants` citations in
  acceptance.md (AC-TIER-001 body, Evidence Ledger R1) and plan.md §F M1 become
  `go test ./internal/config/ -run '^TestClaudeTierConstants$'` — unanchored selectors are
  substring matches and can sweep a superset while reporting ok (t1371 root cause);
  anchoring pins the swept set to the exact test.
- D2 (GLM audit surface single-sourcing): new AC-TIER-015 (REQ-TIER-006) —
  `internal/cli/mcp_glm.go`'s audit defaults must resolve from the shared config constants
  ({DefaultGLM53, max}) with no inline model/effort literals; RED-now cell EVID-TIER-F
  observed on the run-start tree; flipped by M1.
- Post-amendment artifact hashes (run-entry basis; sha256, measured after D1/D2 land,
  before the first run-phase spawn):

  - spec.md: 9d94ce733a4c10f6e51fa165a1f598f54b49605ee94bbfcf7825ade031636d88
  - plan.md: ff5920332a6a9a1168922b6aebebd4d598d610381c4d0011a2e81b7b19d22c4e
  - acceptance.md: f3c0728bdd576bc6b3d1b1852741b432b345adddbcc0cf4efbe145fccb57ddce

- Pre-amendment verdict state: the plan-audit verdict judged the pre-D1/D2 artifact set;
  the deltas are exactly D1+D2 (operator-ordered), nothing else moved.
- Progression mode: operator-driven lane cadence — no `/moai goal` armed (the lane's Stop
  evaluator is not the card's judge; the leader advances the card on evidence per
  kanban-dispatch).
