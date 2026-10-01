---
id: SPEC-AGENT-TIER-001
title: "Acceptance criteria — three-tier subagent model-effort configuration"
version: "0.1.0"
created: 2026-10-01
---

# SPEC-AGENT-TIER-001 — Acceptance Criteria

15 criteria; 13 release-blocking, 2 regression-guard (§D.1). Every criterion carries two cells
per verification-completeness.md §2: a RED-now observation pinned to the pre-implementation tree
(`f130aa041bb90c81b235029afffa806a65e88480`, observed 2026-10-01 — Evidence Ledger below) and a
green-path cell naming the milestone that flips it.

## §D AC Matrix

| AC | Requirement | Criterion | Severity | RED cell | Green path |
|---|---|---|---|---|---|
| AC-TIER-001 | REQ-TIER-002/012 | Tier-constants test RED at run start, GREEN after M1 | Blocking | R1 (run start) | M1 |
| AC-TIER-002 | REQ-TIER-004 | Go default audit pins byte-exact vs operator table | Blocking | EVID-TIER-A | M1 |
| AC-TIER-003 | REQ-TIER-004/014 | Template workflow.yaml pins byte-exact + mirror parity | Blocking | EVID-TIER-C | M1 |
| AC-TIER-004 | REQ-TIER-004 | Resolver terminal fallbacks consistent | Blocking | R2 (run start) | M1 |
| AC-TIER-005 | REQ-TIER-013 | No-hardcode sweep green with positive control | Blocking | R3 (run start) | M1 |
| AC-TIER-006 | REQ-TIER-009 | Closed tier set rejects unknown token | Blocking | R4 (run start) | M2 |
| AC-TIER-007 | REQ-TIER-007/008 | Default class→tier map + audit exclusion | Blocking | R4 (run start) | M2 |
| AC-TIER-008 | REQ-TIER-003 | Max-tier fallback pair recorded | Blocking | R1 (run start) | M1 |
| AC-TIER-009 | REQ-TIER-001 | Effort vocabulary unchanged; tier/effort separation documented | Blocking | EVID-TIER-C (partial) | M1/M2 |
| AC-TIER-010 | REQ-TIER-011 | Web widget renders three tiers with chart figures | Blocking | EVID-TIER-D | M3 |
| AC-TIER-011 | REQ-TIER-004 | User pin precedence preserved | Regression-guard | existing suite | M1 |
| AC-TIER-012 | REQ-TIER-010 | Doctrine non-regression (no model:/effort: frontmatter; observer inert) | Regression-guard | existing suite | M1-M3 |
| AC-TIER-013 | gates | Affected-package verification green | Blocking | baseline §C | M1-M3 |
| AC-TIER-014 | phase gate | plan-auditor PASS ≥ 0.80 (Tier M) | Blocking | — | plan→run gate |
| AC-TIER-015 | REQ-TIER-006 | GLM audit surface resolves from the shared pin constants | Blocking | EVID-TIER-F | M1 |

### AC-TIER-001 — the tier-constants test is RED at run start and GREEN after M1

**Given** the run-start tree before any M1 edit, **when** the tier-constants test is authored
first (`internal/config`, e.g. `TestClaudeTierConstants`) and executed via
`go test ./internal/config/ -run '^TestClaudeTierConstants$'`, **then** it fails with a compilation
error naming the undefined tier symbols (exit 1 — RED for the right reason: the symbols do not
exist anywhere in the package; EVID-TIER-B corroborates the absence by grep); and **after M1**
the same command exits 0 with assertions pinning max→{sonnet-5-5, max}, medium→{sonnet-5-5,
high}, low→{sonnet-5-5, medium} byte-exact.

### AC-TIER-002 — the Go default audit pins match the operator table byte-exact

**Given** `NewDefaultWorkflowConfig` in `internal/config/defaults.go`, **when** the `Audit:`
block is read (`grep -n -A 3 "Claude: ModelEffort{" internal/config/defaults.go` plus the Codex
and GLM cells), **then** Claude = {claude-opus-5-5, high}, Codex = {gpt-6.1-sol, high}, GLM =
{glm-5.3, max} — the exact strings, no other value. RED-now: EVID-TIER-A (claude effort
observed `medium` at `:1222`); the GLM pin is absent from the block today (EVID-TIER-C shows
the mirror side). Flipped by M1.

### AC-TIER-003 — the template workflow.yaml mirror carries the same pins and survives make build

**Given** `internal/template/templates/.moai/config/sections/workflow.yaml`, **when** the
`audit:` block is read, **then** `claude.effort` is `high`, `glm` is `{model: glm-5.3,
effort: max}`, `codex` stays `{gpt-6.1-sol, high}`; and after `make build` the regenerated
embedded copy is byte-identical to the committed template (no drift). RED-now: EVID-TIER-C
(`medium` at `:115`; glm empty at `:120-121`). Flipped by M1.

### AC-TIER-004 — the audit resolver terminal fallbacks agree with the defaults

**Given** `resolveClaudeAuditModelEffort` (internal/cli/mcp_claude.go:181),
`resolveCodexAuditModelEffort` (internal/cli/mcp_codex.go:213), and the GLM audit pin path
(internal/cli/mcp_glm.go), **when** a unit test exercises each with no user pin and no explicit
caller override, **then** claude resolves {claude-opus-5-5, high}, codex resolves
{gpt-6.1-sol, high} (unchanged), and glm resolves {glm-5.3, max} with the effort forwarded
verbatim (REQ-AMP-006 semantics). RED cell R2: the tests are authored RED-first at run start
(the claude fallback returns medium today; the GLM pin is empty). Flipped by M1.

### AC-TIER-005 — the no-hardcode sweep is green only after a demonstrated red

**Given** M1 complete, **when** the sweep greps the pin literals (`gpt-6.1-sol`, and the
{claude-opus-5-5, high} / {glm-5.3, max} pairings) across `internal/` excluding the declared
single-source locations and the template mirror, **then** zero third locations appear; and the
sweep is first demonstrated red on a planted violation (a scratch file carrying a pin literal,
then removed) with the swept file count reported, so an empty sweep can never read as a pass.
RED cell R3 (run start: the sweep script + positive control are authored and demonstrated
before M1's green is read). Flipped by M1.

### AC-TIER-006 — an unknown tier token fails the load

**Given** a config fixture whose tier assignment carries `extreme`, **when** the loader runs,
**then** the load fails with an error naming `extreme` and the fixture path; and fixtures
carrying each of `max` / `medium` / `low` load cleanly. RED cell R4: the validation test is
authored RED-first in M2 (no tier surface exists at run start — EVID-TIER-B). Flipped by M2.

### AC-TIER-007 — the default class→tier map assigns exactly the operator table

**Given** the default assignment surface from M2, **when** a unit test resolves each named
class, **then** super-advisor→max, manager-spec→max, manager-develop→medium, manager-docs→medium,
e2e-tester→medium, Explore→low, and the general implementation lane→medium; and plan-auditor,
sync-auditor, and the claude/codex/glm audit backends resolve to NO tier (audit surfaces are
excluded and ride the audit pins). RED cell R4 (same RED-first test authoring). Flipped by M2.

### AC-TIER-008 — the max-tier fallback pair is recorded as a named declaration

**Given** `internal/config`, **when** a unit test reads the max-tier fallback declaration, then
it equals {claude-opus-5-5, xhigh} and its doc comment cites the chart grounding (65% @ ~$5).
RED cell R1 (authored with the tier-constants test). Flipped by M1.

### AC-TIER-009 — the effort vocabulary is unchanged and the Q2 separation is documented

**Given** the existing effort-vocabulary surfaces (template `llm.yaml` / `workflow.yaml` effort
keys, `resolveLaunchEffort`, the GLM effort overlay), **when** M1/M2 land, **then** those
surfaces' effort vocabularies are byte-unchanged (targeted grep diff empty), and the tier
constants' doc comment states the Q2 definition (tier tokens are configuration-key names, not
effort values). Partial RED-now: EVID-TIER-C shows the pin-side state as it must NOT remain.
Flipped by M1 (doc comment) and M2 (vocabulary diff).

### AC-TIER-010 — the web console renders the three tiers with the chart figures

**Given** the moai web console build, **when** a render test executes the tier UI surface,
**then** the rendered output contains the three tier keys (`max`, `medium`, `low`) and the
figures `70.6%`, `$11`, `45%`, `$2.3`, `29%`, `$0.8`, and the selection accepts only the closed
set. RED-now: EVID-TIER-D (no chart figure anywhere under `internal/web/`). Flipped by M3.

### AC-TIER-011 — user-set pin precedence survives the default flip (regression-guard)

**Given** the existing precedence tests (user pin > SSOT cell > default; empty model = no pin;
explicit caller model outranks the pin), **when** the defaults change, **then** those tests pass
unmodified except where a test asserts the DEFAULT VALUE itself (those are updated in the same
commit as the default, and the update is listed in progress.md §E.2). Verified against the §C
pre-flight baseline counts.

### AC-TIER-012 — doctrine and observer non-regression (regression-guard)

**Given** the agent definition trees (C1/C2/C3) and the spawn observer, **when** M1-M3 land,
**then** no agent file gains `model:`/`effort:` (or `tier:`) frontmatter, `make agents-emit-check`
exits 0, and the existing `agent_model_guard` declaration-only tests pass unchanged.

### AC-TIER-013 — affected-package verification is green (phase evidence)

**Given** M1-M3 complete, **when** the batch `go vet ./internal/config/... ./internal/cli/...
./internal/web/...`, `golangci-lint run`, `go test` on the same packages, `make build`, and
`make embed-check` runs, **then** every command exits 0 with the output recorded verbatim in
progress.md §E.2.

### AC-TIER-014 — plan-auditor PASS is the phase gate

**Given** the completed plan-phase artifact set, **when** plan-auditor runs, **then** the
verdict is PASS at score ≥ 0.80 (Tier M threshold) with no unaddressed blocking finding; that
verdict is the plan→run entry evidence.

### AC-TIER-015 — the GLM audit surface resolves from the shared pin constants

**Given** the GLM audit default path in `internal/cli/mcp_glm.go` (the audit default constants
and the audit pin resolution), **when** a unit test reads the resolved audit model/effort with
no user pin and no explicit caller override, **then** the model is `DefaultGLM53` (`glm-5.3` —
NOT the flash slot default `DefaultGLMHigh`) and the effort is `max` forwarded verbatim
(REQ-AMP-006 semantics), with no inline model/effort literals anywhere in that audit path.
RED-now: EVID-TIER-F (the current default binds `DefaultGLMHigh`, the flash variant). Flipped
by M1.

## §D.1 Severity

- **Release-blocking (13):** AC-TIER-001..010, AC-TIER-013, AC-TIER-014, AC-TIER-015.
- **Regression-guard (2):** AC-TIER-011, AC-TIER-012 (already-green invariants preserved;
  verified by the existing suite, not by a new RED flip).

## §D.2 Traceability

| REQ | ACs |
|---|---|
| REQ-TIER-001 | AC-TIER-009 |
| REQ-TIER-002 | AC-TIER-001 |
| REQ-TIER-003 | AC-TIER-008 |
| REQ-TIER-004 | AC-TIER-002, AC-TIER-003, AC-TIER-004, AC-TIER-011 |
| REQ-TIER-005 | AC-TIER-003 (web parity via AC-TIER-010's closed set), AC-TIER-010 |
| REQ-TIER-006 | AC-TIER-002, AC-TIER-003, AC-TIER-004, AC-TIER-015 |
| REQ-TIER-007 | AC-TIER-007 |
| REQ-TIER-008 | AC-TIER-007 |
| REQ-TIER-009 | AC-TIER-006 |
| REQ-TIER-010 | AC-TIER-012 |
| REQ-TIER-011 | AC-TIER-010 |
| REQ-TIER-012 | AC-TIER-001 (+ every blocking AC's two-cell structure) |
| REQ-TIER-013 | AC-TIER-005 |
| REQ-TIER-014 | AC-TIER-003, AC-TIER-012 |

## §D.3 Indirect verification

- AC-TIER-011 and AC-TIER-012 are verified indirectly through the existing test suites and
  `make agents-emit-check` (no new RED instrument is authored for them); their evidence is the
  baseline-vs-after comparison of the §C pre-flight counts.
- Template mirror parity (AC-TIER-003) is doubly verified: content grep + `make build`
  byte-identity.

## §D.4 Closure gates

- Run→sync entry: AC-TIER-001..010, AC-TIER-015 and AC-TIER-013 green with verbatim evidence in
  progress.md §E.2; regression-guards green against baseline.
- Sync close: sync-audit verdict recorded in progress.md §E.4 per the standard flow.

## §D.5 Forward-looking checks

- Automatic failover activation for the max tier — revisit when an availability signal exists
  (deliberately out of scope; REQ-TIER-003 records only).
- Cost telemetry per tier (the grounding report's success metrics: audit PASS rate ≥95,
  implementation-lane processing time, per-card cost on max-tier surfaces) — measurable once
  the tier axis lands; not gates in this SPEC.
- Benchmark re-adjudication trigger (FrontierCode / CursorBench tab reversal) — a new card, not
  an amendment here.

## Evidence Ledger (RED-now observations — observed 2026-10-01)

Tree SHA for every cell below: `f130aa041bb90c81b235029afffa806a65e88480`.
Exit-code provenance: this harness surfaces non-zero exits as an explicit `Exit code N` marker
(calibrated in-run with `false` → `Exit code 1`) and suppresses the marker on output-bearing
runs; cells marked "exit 0 (observed)" showed command output with no error marker, which under
the calibration is the zero-exit rendering for plain grep.

**EVID-TIER-A** (AC-TIER-002 RED-now)
- Command: `grep -n -A 3 "Claude: ModelEffort{" internal/config/defaults.go`
- stdout (verbatim):
```
1220:			Claude: ModelEffort{
1221-				Model:  "claude-opus-5-5",
1222-				Effort: "medium",
1223-			},
```
- exit: 0 (observed)
- Red because: AC-TIER-002 requires `high`; the Go default pins `medium`. The file is exactly
  the file M1 edits, so the green path exists (no wrong-reason red).

**EVID-TIER-B** (AC-TIER-001/006/007/008 corroborating absence)
- Command: `grep -rn "DefaultClaudeTierMax\|DefaultAgentTier" internal/config/`
- stdout: (empty — no selection)
- exit: no-match per grep semantics; harness rendered empty output without an exit marker
  (stated per the calibration note — the red fact is carried by the empty stdout: the symbols
  the criteria require do not exist)
- Red because: the tier constants are absent from the package; RED for the right reason (M1
  creates exactly these symbols; nothing pre-existing can satisfy the criteria).

**EVID-TIER-C** (AC-TIER-003/009 RED-now)
- Command: `grep -n -B 1 "effort:" internal/template/templates/.moai/config/sections/workflow.yaml`
- stdout (verbatim):
```
114-            model: claude-opus-5-5
115:            effort: medium
--
117-            model: gpt-6.1-sol
118:            effort: high
--
120-            model: ""
121:            effort: ""
```
- exit: 0 (observed)
- Red because: the mirror pins `medium` (must be `high`) and ships the GLM pin empty (must be
  {glm-5.3, max}); codex `high` is already correct (no-change surface).

**EVID-TIER-D** (AC-TIER-010 RED-now)
- Command: `grep -rn "70\.6" internal/web/`
- stdout: (empty — no selection)
- exit: no-match per grep semantics; harness rendering as EVID-TIER-B
- Red because: no web surface carries the chart figures; M3's render test is the green path.

**EVID-TIER-E** (REQ-TIER-006 corroborating)
- Command: `grep -n -A 2 "        glm:" internal/template/templates/.moai/config/sections/workflow.yaml`
- stdout (verbatim):
```
119:        glm:
120-            model: ""
121-            effort: ""
```
- exit: 0 (observed)
- Red because: the GLM audit pin ships empty where {glm-5.3, max} is required.

**EVID-TIER-F** (AC-TIER-015 RED-now)
- Command: `grep -n "DefaultGLM" internal/cli/mcp_glm.go`
- stdout (verbatim):
```
6:// (loadGLMKey → ~/.moai/.env.glm) and endpoint (config.DefaultGLMBaseURL) the
54:	glmAuditDefaultModel = config.DefaultGLMHigh
56:	// glmMessagesPath is appended to config.DefaultGLMBaseURL to form the
303:	url := config.DefaultGLMBaseURL + glmMessagesPath
```
- exit: 0 (observed)
- tree: `f130aa041bb90c81b235029afffa806a65e88480`
- Red because: the GLM audit default binds `DefaultGLMHigh` (the flash slot default) where the
  operator pin targets full glm-5.3 with effort max; M1 edits exactly this file, so the green
  path exists (no wrong-reason red).

**Planned run-start RED cells** (executable only once the corresponding test files exist; each
is executed and pinned to the run-start SHA in progress.md §E.2 BEFORE the first M1 edit):

- **R1** (AC-TIER-001/008): `go test ./internal/config/ -run '^TestClaudeTierConstants$'` after the
  test is authored — expected compile failure naming the undefined tier symbols, exit 1.
- **R2** (AC-TIER-004): the resolver fallback tests authored RED-first — expected claude
  fallback asserting `high` fails against today's `medium`, exit 1.
- **R3** (AC-TIER-005): the sweep demonstrated red on a planted pin-literal scratch file —
  expected non-zero sweep verdict + reported swept count, then the file removed.
- **R4** (AC-TIER-006/007): the tier-surface validation and mapping tests authored RED-first in
  M2 — expected compile/absence failure, exit 1.
