# SPEC-V3R2-WF-003 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.2.0, `research.md` v0.1.0, `plan.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-03 | MoAI Plan Workflow (Phase 1B)     | Initial G/W/T conversion of 15 ACs (AC-WF003-01 through AC-WF003-15)   |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow (iter 2)       | Added AC-WF003-16 (REQ-007 matrix publication) + AC-WF003-17 (REQ-015 future-extension) G/W/T scenarios. Added Traceability Matrix (REQ → AC) and reverse matrix (AC → REQ) sections per plan-auditor iteration 1 D3 fix. |

---

## Scope

This document converts each of the 15 ACs from `spec.md` §6 into Given/When/Then format with happy-path + edge-case + test-mapping notation. Test mapping uses the extended `internal/template/agentless_audit_test.go` (3 new test functions added in plan §M1) and direct skill-content audit. Runtime assertions are anchored at the skill-body sentinel level since the slash-command path has no Go handler (per research.md §2.1).

Notation:
- **Test mapping** identifies which Go test function (or manual verification step) covers the AC.
- AC-WF003-06, 07, 11 have the highest test importance (sentinel enforcement; cross-spec consistency).

---

## AC-WF003-01 — `/moai run SPEC-001 --mode autopilot` runs single-lead orchestration

Maps to: REQ-WF003-001, REQ-WF003-002.

### Happy path

- **Given** SPEC-V3R2-WF-001 (or any valid SPEC) exists in `.moai/specs/`
- **When** the user invokes `/moai run SPEC-001 --mode autopilot`
- **Then** the orchestrator parses `--mode autopilot` and routes to the autopilot dispatch branch in `run.md` Mode Dispatch section
- **And** Phase 0.95 Scale-Based Mode Selection auto-picks Fix/Focused/Standard/Full Pipeline based on SPEC scope (no team mode)
- **And** Phase 2A or 2B executes per `quality.yaml development_mode` (single manager-ddd or manager-tdd lead, not parallel teammates)
- **And** the workflow completes without invoking `Skill("moai-workflow-loop")` or `TeamCreate`

### Edge case — autopilot is the harness-default for minimal/standard

- **Given** harness level = `minimal` AND user invokes `/moai run SPEC-001` WITHOUT `--mode` flag
- **When** the orchestrator runs the mode resolver
- **Then** the resolver determines: no CLI flag → no `workflow.default_mode` set → harness == minimal → default to `autopilot`
- **And** the same single-lead dispatch executes
- **And** no info log about mode selection is emitted (autopilot is the silent default)

### Test mapping

- Manual verification: run `/moai run SPEC-001 --mode autopilot` in a test fixture; trace orchestrator log to confirm autopilot branch was taken.
- Static content audit: extended `internal/template/agentless_audit_test.go` `TestRunDesignSkillsContainModeUnknownSentinel/run.md` indirectly verifies the dispatch section is present.
- Skill body content check: `grep -n "autopilot" .claude/skills/moai/workflows/run.md` returns the Mode Dispatch section enumerating the 4 mode values.

---

## AC-WF003-02 — `/moai run SPEC-001 --mode loop` invokes Ralph engine

Maps to: REQ-WF003-008.

### Happy path

- **Given** SPEC-001 exists and the user wants iterative auto-fixing
- **When** the user invokes `/moai run SPEC-001 --mode loop`
- **Then** the orchestrator parses `--mode loop` and the Mode Dispatch branch delegates execution to `Skill("moai-workflow-loop")` with the SPEC-ID and remaining args
- **And** Phase 2A/2B is BYPASSED (no manager-ddd or manager-tdd direct delegation)
- **And** `loop.md` per-iteration cycle (`loop.md:46-138`) executes: Step 1 Completion Check → Step 2 Memory Check → ... → Step 9 Repeat or Exit
- **And** the loop terminates per `loop.md:140-152` exit conditions (completion marker, max iterations, memory pressure)

### Edge case — loop with --max flag combined

- **Given** `/moai run SPEC-001 --mode loop --max 50` is invoked
- **When** orchestrator dispatches to `moai-workflow-loop` skill
- **Then** the `--max 50` flag is forwarded to the loop skill (parsed per `loop.md:35-44` Supported Flags)
- **And** the loop runs at most 50 iterations
- **And** behavior is identical to invoking `/moai loop SPEC-001 --max 50` directly (REQ-WF003-004 alias contract)

### Edge case — loop mode never auto-selected

- **Given** harness level = `thorough` (would auto-select team if prereqs met)
- **When** user invokes `/moai run SPEC-001` WITHOUT `--mode` flag
- **Then** the harness-based default selects `team` (or `autopilot` fallback per REQ-WF003-012); NEVER `loop`
- **And** `loop` mode is opt-in only (via explicit `--mode loop` or `workflow.default_mode: loop` config)

### Test mapping

- Static content audit: `TestLoopAliasCrossReference` in extended `internal/template/agentless_audit_test.go` asserts `loop.md` contains the literal phrase `/moai run --mode loop`, documenting the cross-route equivalence.
- Manual verification: run `/moai run SPEC-001 --mode loop` and observe orchestrator delegating to `moai-workflow-loop` skill (via skill load message).

---

## AC-WF003-03 — `/moai loop SPEC-001` is identical to `/moai run --mode loop SPEC-001`

Maps to: REQ-WF003-004.

### Happy path

- **Given** the user invokes `/moai loop SPEC-001`
- **When** the thin command wrapper at `.claude/commands/moai/loop.md:7` executes (`Use Skill("moai") with arguments: loop $ARGUMENTS`)
- **Then** the orchestrator loads `.claude/skills/moai/workflows/loop.md` directly (today's behavior preserved)
- **And** the per-iteration cycle executes identically to `/moai run SPEC-001 --mode loop`
- **And** `loop.md` § `## Invocation Routes (SPEC-V3R2-WF-003)` (added in M4a) documents this equivalence verbatim

### Edge case — argument forwarding identity

- **Given** `/moai loop SPEC-001 --max 50 --auto-fix --memory-check` is invoked
- **When** orchestrator parses arguments
- **Then** all flags are forwarded to `loop.md` per its Supported Flags section
- **And** `/moai run SPEC-001 --mode loop --max 50 --auto-fix --memory-check` produces identical iteration behavior

### Edge case — alias documentation in CHANGELOG

- **Given** the SPEC-V3R2-WF-003 implementation is merged
- **When** a user reads CHANGELOG `## [Unreleased]` (post M5)
- **Then** the alias relationship is explicitly noted: "`/moai loop` becomes an alias for `/moai run --mode loop` (REQ-WF003-004)"

### Test mapping

- Static content audit: `TestLoopAliasCrossReference` in `internal/template/agentless_audit_test.go` asserts the literal string `/moai run --mode loop` appears in `loop.md`. Pseudocode for the Go test (do not implement here — outline only):
  - Read `loop.md` from embedded FS.
  - Assert `bytes.Contains(body, []byte("/moai run --mode loop"))` is true.
  - On failure: emit `t.Errorf("loop.md missing /moai run --mode loop alias documentation per REQ-WF003-004")`.

---

## AC-WF003-04 — Harness `thorough` + `team.enabled: true` auto-selects team mode

Maps to: REQ-WF003-003.

### Happy path

- **Given** `.moai/config/sections/harness.yaml` triggers harness level = `thorough` for the SPEC (e.g., security_keywords present per `harness.yaml:29`) AND `.moai/config/sections/workflow.yaml: team.enabled: true` AND env var `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is set
- **When** the user invokes `/moai run SPEC-001` WITHOUT `--mode` flag
- **Then** the orchestrator's mode resolver: no CLI flag → no `workflow.default_mode` → harness == thorough AND team enabled AND env set → resolve to `team`
- **And** Team Mode Routing (`run.md:927-943`) executes: TeamCreate spawns backend-dev + frontend-dev + tester + quality teammates
- **And** the workflow completes via team coordination

### Edge case — `default_mode` config override

- **Given** the same conditions as happy path BUT `.moai/config/sections/workflow.yaml: default_mode: autopilot`
- **When** user invokes `/moai run SPEC-001` WITHOUT `--mode` flag
- **Then** the resolver: no CLI flag → `workflow.default_mode: autopilot` → use `autopilot` (config beats harness auto)
- **And** Team Mode is NOT activated
- **And** single-lead autopilot dispatch executes (REQ-WF003-018 precedence rule)

### Test mapping

- Manual verification: set up a test fixture with thorough-triggering SPEC + team enabled + env var; run `/moai run SPEC-XXX`; trace mode resolver logs to confirm `team` was selected.
- Skill body content check: `run.md` Mode Dispatch section enumerates the auto-selection rules per REQ-WF003-002/003.

---

## AC-WF003-05 — `/moai design --mode team` spawns parallel copywriting + brand-design teammates

Maps to: REQ-WF003-009.

### Happy path

- **Given** team prerequisites are satisfied (`workflow.team.enabled: true` AND `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`) AND brand context exists in `.moai/project/brand/`
- **When** the user invokes `/moai design --mode team`
- **Then** the orchestrator skips Phase 1 path selection AskUserQuestion (since explicit `--mode team` is supplied)
- **And** the orchestrator goes directly to a team-coordinated Phase B-Common variant
- **And** TeamCreate spawns two teammates in parallel:
  - `moai-domain-copywriting` teammate (role_profile: designer per `workflow.yaml:52-56`)
  - `moai-domain-brand-design` teammate (role_profile: designer)
- **And** both teammates feed into `moai-workflow-gan-loop` for evaluation per Phase C
- **And** the existing GAN Loop contract (max 5 iterations, pass_threshold 0.75) governs convergence

### Edge case — team prereqs missing → MODE_TEAM_UNAVAILABLE

- **Given** user invokes `/moai design --mode team` BUT `workflow.team.enabled: false` OR `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` unset
- **When** the orchestrator checks prerequisites
- **Then** the orchestrator emits `MODE_TEAM_UNAVAILABLE` error and suggests `--mode autopilot` fallback (per AC-WF003-07)
- **And** TeamCreate is NOT invoked
- **And** the user must explicitly retry with `--mode autopilot` or fix the prerequisites

### Test mapping

- Static content audit: `TestRunDesignSkillsContainModeUnknownSentinel/design.md` indirectly verifies the Mode Dispatch section is present.
- Manual verification: with team prereqs satisfied, invoke `/moai design --mode team` and observe TeamCreate calls in trace.

---

## AC-WF003-06 — `/moai run --mode banana` emits MODE_UNKNOWN

Maps to: REQ-WF003-010.

### Happy path

- **Given** the user invokes `/moai run SPEC-001 --mode banana` (any value not in the 4 valid set: `autopilot`, `loop`, `team`, `pipeline`)
- **When** the orchestrator parses arguments and the mode validator runs
- **Then** the orchestrator emits an error message containing the literal string `MODE_UNKNOWN`
- **And** the error message lists the 4 valid `--mode` values: `autopilot`, `loop`, `team`, `pipeline`
- **And** the workflow does NOT execute Phase 0.95 or any phase past argument parsing
- **And** the user receives guidance to retry with a valid mode value

### Edge case — case sensitivity

- **Given** the user invokes `/moai run SPEC-001 --mode AUTOPILOT` (uppercase)
- **When** the validator runs
- **Then** `MODE_UNKNOWN` is emitted (mode values are case-sensitive lowercase per spec.md §5.1 REQ-WF003-001)
- **OR** the orchestrator MAY normalize to lowercase before validation (implementation choice; document in `run.md` Mode Dispatch section)

### Edge case — empty value

- **Given** the user invokes `/moai run SPEC-001 --mode` (no value)
- **When** the orchestrator parses the flag
- **Then** the standard CLI argument parser rejects with a missing-value error (orchestrator-level, not `MODE_UNKNOWN`)

### Edge case — same rejection on `/moai design`

- **Given** the user invokes `/moai design --mode banana`
- **When** validator runs
- **Then** `MODE_UNKNOWN` is emitted with the 4 valid values listed for design (`autopilot`, `import`, `team`, `pipeline`)
- **And** the design-specific valid set differs from run's set (per spec.md §2.1 — design supports `import` instead of `loop`); the error message MUST reflect the per-skill valid set

### Test mapping

- Static content audit: `TestRunDesignSkillsContainModeUnknownSentinel` in `internal/template/agentless_audit_test.go` (M1) walks `run.md` and `design.md` and asserts each contains the literal string `MODE_UNKNOWN`. Pseudocode (do not implement — outline only):
  - For each of {`run.md`, `design.md`} under `.claude/skills/moai/workflows/` in embedded FS:
    - Read body bytes.
    - Assert `bytes.Contains(body, []byte("MODE_UNKNOWN"))` is true.
    - On failure: emit `t.Errorf("%s missing MODE_UNKNOWN sentinel per SPEC-V3R2-WF-003 REQ-WF003-010", path)`.
- Manual verification: invoke `/moai run --mode banana` in test fixture; observe error containing `MODE_UNKNOWN` and 4 valid values.

---

## AC-WF003-07 — `/moai run --mode team` without prereqs emits MODE_TEAM_UNAVAILABLE

Maps to: REQ-WF003-011.

### Happy path

- **Given** the user invokes `/moai run SPEC-001 --mode team` BUT `.moai/config/sections/workflow.yaml: team.enabled: false`
- **When** the orchestrator parses `--mode team` and runs the team prerequisite check (per `run.md:927-943`)
- **Then** the orchestrator emits an error message containing the literal string `MODE_TEAM_UNAVAILABLE`
- **And** the error message suggests `--mode autopilot` as fallback
- **And** the error message identifies which prerequisite is missing (e.g., `workflow.team.enabled=false` or `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS unset`)
- **And** TeamCreate is NOT invoked
- **And** the user must explicitly retry with a different mode

### Edge case — env var missing but config enabled

- **Given** `workflow.team.enabled: true` BUT `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var is unset
- **When** user invokes `/moai run SPEC-001 --mode team`
- **Then** `MODE_TEAM_UNAVAILABLE` is emitted (BOTH conditions must be true; either missing triggers the sentinel)
- **And** the error message specifically calls out the env var requirement

### Edge case — same behavior on `/moai design --mode team`

- **Given** user invokes `/moai design --mode team` without prereqs
- **When** orchestrator parses
- **Then** `MODE_TEAM_UNAVAILABLE` is emitted (same sentinel, applied to design)
- **And** the suggested fallback is `--mode autopilot` (design's autopilot = code-based path B per spec.md §2.1)

### Edge case — distinction from MODE_AUTO_DOWNGRADE

- **Given** harness auto-selects team but prereqs missing (no `--mode` flag from user)
- **When** the orchestrator runs auto-resolution
- **Then** the orchestrator emits a `[mode-auto-downgrade]` info log (REQ-WF003-012, per AC-WF003-08), NOT `MODE_TEAM_UNAVAILABLE`
- **And** the workflow proceeds with autopilot (silent fallback)
- **And** the distinction is: explicit user `--mode team` request → error sentinel; implicit auto-resolution → info log only

### Test mapping

- Static content audit: `TestRunSkillContainsModeTeamUnavailableSentinel` in `internal/template/agentless_audit_test.go` (M1) asserts `run.md` contains the literal string `MODE_TEAM_UNAVAILABLE`. Pseudocode:
  - Read `.claude/skills/moai/workflows/run.md` from embedded FS.
  - Assert `bytes.Contains(body, []byte("MODE_TEAM_UNAVAILABLE"))` is true.
  - On failure: emit `t.Errorf("run.md missing MODE_TEAM_UNAVAILABLE sentinel per SPEC-V3R2-WF-003 REQ-WF003-011")`.
- Manual verification: with `team.enabled: false`, invoke `/moai run --mode team`; observe error.

---

## AC-WF003-08 — Harness auto-select team but prereqs missing → autopilot with info log

Maps to: REQ-WF003-012.

### Happy path

- **Given** harness level = `thorough` (e.g., security_keywords present) BUT team prereqs not satisfied (`team.enabled: false` OR env var unset)
- **When** the user invokes `/moai run SPEC-001` WITHOUT `--mode` flag
- **Then** the mode resolver: no CLI flag → no `workflow.default_mode` → harness wants team but prereqs missing → fallback to `autopilot`
- **And** the orchestrator emits an info-level log message (NOT an error): `[mode-auto-downgrade] Harness level=thorough requested team mode, but prerequisites are not satisfied (workflow.team.enabled=<bool>, CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=<set/unset>). Falling back to autopilot mode.`
- **And** the workflow proceeds normally with autopilot dispatch (single-lead orchestration)
- **And** the user may continue without intervention (this is a UX-soft fallback, not a blocking error)

### Edge case — info log contains remediation hint

- **Given** auto-downgrade triggered
- **When** info log is rendered
- **Then** the message includes a hint: "To use team mode, set workflow.team.enabled=true and export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1"
- **And** the user can fix prerequisites between runs to enable team mode in the future

### Edge case — explicit override prevents downgrade

- **Given** harness wants team but prereqs missing AND user explicitly supplies `--mode autopilot`
- **When** orchestrator runs the resolver
- **Then** the resolver short-circuits at "CLI flag wins" (REQ-WF003-018) without invoking the harness path
- **And** NO info log is emitted (downgrade is harness-resolved, not user-requested; user explicitly chose autopilot)

### Test mapping

- Skill body content check: `grep -n "mode-auto-downgrade\|REQ-WF003-012" .claude/skills/moai/workflows/run.md` returns the Mode Dispatch section documenting the fallback behavior.
- Manual verification: configure thorough harness + `team.enabled: false`; invoke `/moai run` without flag; observe info log AND autopilot execution.

---

## AC-WF003-09 — `/moai plan --mode loop` ignores `--mode` and runs plan workflow

Maps to: REQ-WF003-005.

### Happy path

- **Given** the user invokes `/moai plan --mode loop "feature description"`
- **When** the orchestrator parses `--mode loop` on a non-mode-axis subcommand
- **Then** the `--mode` flag is silently ignored
- **And** the plan workflow proceeds normally (per `plan.md` Phase 0 / 0.5 / 1A / 1B / 2 / 3 / 4)
- **And** no info log or error related to mode is emitted (silent ignore for non-pipeline mode values)

### Edge case — same behavior for `/moai sync --mode team`

- **Given** user invokes `/moai sync SPEC-001 --mode team`
- **When** orchestrator parses
- **Then** `--mode team` is silently ignored
- **And** the sync workflow proceeds normally
- **And** no team activation occurs (sync is mode-NA per REQ-WF003-005)

### Edge case — `/moai plan --mode pipeline` triggers MODE_PIPELINE_ONLY_UTILITY (separate AC)

- **Given** user invokes `/moai plan --mode pipeline`
- **When** orchestrator parses
- **Then** MODE_PIPELINE_ONLY_UTILITY is emitted (per AC-WF003-11) — NOT silently ignored
- **And** this is the ONE special case where plan/sync DO act on `--mode` value

### Test mapping

- Skill body content check: `plan.md` and `sync.md` `## Mode Flag Compatibility` section (refined in M4d) explicitly states "any `--mode` value supplied to `/moai plan` is silently ignored" with exception for `pipeline`.
- Manual verification: invoke `/moai plan --mode loop "test"`; observe normal plan execution with no mode-related log.

---

## AC-WF003-10 — `/moai fix --mode autopilot` operates in pipeline mode (not autopilot)

Maps to: REQ-WF003-006, REQ-WF003-016.

### Happy path

- **Given** the user invokes `/moai fix --mode autopilot`
- **When** the orchestrator parses arguments
- **Then** the `--mode` flag is silently ignored (per WF-004 REQ-WF004-011 — utility subcommands ignore mode flag)
- **And** the orchestrator emits info log `MODE_FLAG_IGNORED_FOR_UTILITY` (per WF-004 contract)
- **And** the fix workflow proceeds with the standard 3-phase Agentless pipeline (localize → repair → validate per `fix.md` flow declaration)
- **And** the `autopilot` value is irrelevant — fix is ALWAYS pipeline-classified per WF-004 REQ-WF004-001

### Edge case — distinction from /moai run --mode autopilot

- **Given** the user knows `/moai run --mode autopilot` runs single-lead orchestration AND assumes `/moai fix --mode autopilot` should also do single-lead
- **When** they discover `--mode` is ignored on `/moai fix`
- **Then** the info log + Pipeline Contract section in `fix.md` (added by WF-004 M2) clarifies: "fix is Agentless-classified; --mode flag is ignored"
- **And** the user understands the cross-spec contract: WF-003 governs run/design mode dispatch; WF-004 governs utility classification

### Test mapping

- Static content audit: `TestUtilitySkillsContainModeFlagIgnoredSentinel/fix.md` in `internal/template/agentless_audit_test.go` (added by WF-004 M1) asserts `fix.md` contains `MODE_FLAG_IGNORED_FOR_UTILITY`.
- Static content audit: WF-004's `TestImplementationSkillsContainPipelineRejectionSentinel` is OUT-OF-SCOPE for this AC (fix is utility, not implementation).
- Manual verification: invoke `/moai fix --mode autopilot`; observe `MODE_FLAG_IGNORED_FOR_UTILITY` info log and standard 3-phase fix pipeline execution.

---

## AC-WF003-11 — `/moai run --mode pipeline` rejects with MODE_PIPELINE_ONLY_UTILITY

Maps to: REQ-WF003-016.

### Happy path

- **Given** the user invokes `/moai run SPEC-001 --mode pipeline`
- **When** the orchestrator parses `--mode pipeline` on a multi-agent subcommand
- **Then** the orchestrator emits an error message containing the literal string `MODE_PIPELINE_ONLY_UTILITY`
- **And** the error message points to the utility subcommand set: `fix`, `coverage`, `mx`, `codemaps`, `clean` (5 utility subcommands per WF-004)
- **And** the workflow does NOT execute Phase 0.95 or any phase past argument parsing
- **And** the user must retry with a valid run mode (`autopilot`, `loop`, or `team`)

### Edge case — same rejection on all 4 implementation subcommands

- **Given** any of `/moai {plan,run,sync,design} --mode pipeline` is invoked
- **When** argument parsing runs
- **Then** all 4 produce `MODE_PIPELINE_ONLY_UTILITY` error
- **And** none of the 4 begin their respective workflows
- **And** this matches WF-004 AC-WF004-11 (cross-SPEC consistency)

### Edge case — sentinel string identity across SPECs

- **Given** the same error key `MODE_PIPELINE_ONLY_UTILITY` is documented in WF-003 spec.md (REQ-WF003-016) AND WF-004 spec.md (REQ-WF004-014)
- **When** both SPECs are implemented
- **Then** the error key is byte-identical in both code paths
- **And** the static audit `TestImplementationSkillsContainPipelineRejectionSentinel` (added by WF-004 M1) enforces this in `plan.md`, `run.md`, `sync.md`, `design.md` skill bodies
- **And** WF-003's run phase does NOT introduce a parallel sentinel; it preserves the WF-004-added string

### Test mapping

- Static content audit: `TestImplementationSkillsContainPipelineRejectionSentinel` (added by WF-004 M1) walks `plan.md`, `run.md`, `sync.md`, `design.md` and asserts each contains `MODE_PIPELINE_ONLY_UTILITY`. Pseudocode (do not implement — outline only):
  - For each of {`plan.md`, `run.md`, `sync.md`, `design.md`} under `.claude/skills/moai/workflows/` in embedded FS:
    - Read body bytes.
    - Assert `bytes.Contains(body, []byte("MODE_PIPELINE_ONLY_UTILITY"))` is true.
    - On failure: emit `t.Errorf("%s missing MODE_PIPELINE_ONLY_UTILITY sentinel (shared SPEC-V3R2-WF-003 REQ-WF003-016 + SPEC-V3R2-WF-004 REQ-WF004-014)", path)`.
- Manual verification: invoke `/moai run SPEC-001 --mode pipeline`; observe error.

---

## AC-WF003-12 — `workflow.yaml: default_mode: loop` activates loop mode without `--mode` flag

Maps to: REQ-WF003-014.

### Happy path

- **Given** `.moai/config/sections/workflow.yaml: default_mode: loop` is configured
- **When** the user invokes `/moai run SPEC-001` WITHOUT `--mode` flag
- **Then** the mode resolver: no CLI flag → `workflow.default_mode: loop` → use `loop` (config beats harness auto)
- **And** the orchestrator delegates to `Skill("moai-workflow-loop")` per AC-WF003-02
- **And** Phase 2A/2B is bypassed
- **And** the loop runs to convergence

### Edge case — invalid default_mode value

- **Given** `workflow.yaml: default_mode: banana` is configured
- **When** the orchestrator reads the config
- **Then** the validator emits `MODE_UNKNOWN` (per AC-WF003-06) at config-load time, NOT at command-invoke time
- **And** the orchestrator either fails fast OR falls back to harness-auto (implementation choice; document in M4c yaml comment)

### Edge case — empty default_mode

- **Given** `workflow.yaml: default_mode: ""` (the default per M4c)
- **When** mode resolver runs without `--mode` flag
- **Then** the resolver: no CLI → empty string treated as "not set" → fall through to harness-auto
- **And** standard auto-selection per REQ-WF003-002/003 applies

### Test mapping

- Manual verification: set `default_mode: loop` in fixture; invoke `/moai run SPEC-001`; observe Ralph engine invocation.
- Static content audit: M4c yaml comment must document the valid values (`autopilot`, `loop`, `team`, empty); audit can verify the comment exists via grep on the workflow.yaml file.

---

## AC-WF003-13 — CLI `--mode autopilot` overrides `default_mode: team` config

Maps to: REQ-WF003-018.

### Happy path

- **Given** `.moai/config/sections/workflow.yaml: default_mode: team` is configured AND team prereqs are satisfied
- **When** the user invokes `/moai run SPEC-001 --mode autopilot`
- **Then** the mode resolver: CLI flag `--mode autopilot` present → use `autopilot` (CLI beats config per REQ-WF003-018)
- **And** team mode is NOT activated
- **And** single-lead autopilot dispatch executes
- **And** no info log about the override is emitted (CLI override is the documented behavior)

### Edge case — full precedence chain example

- **Given** harness level = `thorough` (would auto-select team) AND `workflow.default_mode: loop` (config) AND user supplies `--mode autopilot` (CLI)
- **When** mode resolver runs
- **Then** precedence applied: CLI wins → resolved mode = `autopilot`
- **And** neither config (`loop`) nor harness (`team`) is used
- **And** the resolution is documented in the trace log

### Edge case — sentinel checks still apply

- **Given** user supplies `--mode banana` AND `default_mode: autopilot` is set
- **When** validator runs
- **Then** `MODE_UNKNOWN` is emitted (CLI is invalid; precedence rule applies BEFORE config fallback)
- **And** the workflow does NOT silently fall back to config's autopilot value

### Test mapping

- Skill body content check: `run.md` Mode Dispatch section explicitly documents the precedence rule (CLI > config > harness auto) with example.
- Manual verification: configure `default_mode: team`; invoke `/moai run --mode autopilot`; observe autopilot execution (no team).

---

## AC-WF003-14 — Ralph loop terminates on convergence (not infinitely)

Maps to: REQ-WF003-017.

### Happy path

- **Given** `/moai run SPEC-001 --mode loop` (or `/moai loop SPEC-001`) is invoked AND the SPEC has fixable issues
- **When** the per-iteration cycle (`loop.md:46-138`) executes
- **Then** Step 1 Completion Check (`loop.md:49-52`) detects the completion marker `<moai>DONE</moai>` or `<moai>COMPLETE</moai>`, OR Step 4 Completion Condition Check (`loop.md:84-87`) detects "zero errors AND all tests passing AND coverage meets threshold"
- **And** the loop exits with success
- **And** Step 9 (Repeat or Exit) reports the final iteration's output
- **And** the loop does NOT continue past convergence

### Edge case — max iterations reached without convergence

- **Given** the loop runs 100 iterations (default `--max 100`) without convergence
- **When** Step 9 evaluates the iteration counter
- **Then** the loop exits with "max iterations reached" status (`loop.md:140-152` exit conditions)
- **And** remaining issues are displayed to the user
- **And** options are presented (continue, abort, etc.)

### Edge case — memory pressure exit (early termination)

- **Given** loop is running and memory pressure threshold is exceeded (`loop.md:55-71`)
- **When** Step 2 Memory Pressure Check fires
- **Then** the loop saves a memory-emergency snapshot AND exits early with checkpoint
- **And** the exit message suggests `/moai loop --resume memory-emergency`

### Edge case — completion marker auto-detected from previous iteration

- **Given** the previous iteration's output contained `<moai>COMPLETE</moai>`
- **When** Step 1 of the next iteration evaluates
- **Then** the marker is detected immediately (before any work this iteration)
- **And** the loop exits with success (`loop.md:140`)

### Test mapping

- Manual verification: invoke `/moai loop` (or `/moai run --mode loop`) on a fixture with achievable convergence; observe exit at convergence point.
- Reference verification: read `.claude/skills/moai/workflows/loop.md:140-152` to confirm exit conditions are documented and unchanged by WF-003.

---

## AC-WF003-15 — `/moai design --mode import` skips path B, only runs design-import

Maps to: REQ-WF003-013.

### Happy path

- **Given** the user has a Claude Design handoff bundle (.zip or .html) ready
- **When** the user invokes `/moai design --mode import`
- **Then** the orchestrator parses `--mode import` and the Mode Dispatch branch routes to Path A (Claude Design)
- **And** Phase 1 path selection AskUserQuestion (`design.md:64-94`) is SKIPPED (since explicit `--mode import` is supplied)
- **And** Phase B-Common (`design.md:167-192`) is SKIPPED (no copywriting + brand-design loaded)
- **And** the orchestrator goes directly to Step A2 (collect bundle path) → Step A3 (invoke `moai-workflow-design-import`)
- **And** Phase C quality gate (GAN loop) still runs on the imported artifacts (per `design.md:196-217`)

### Edge case — bundle path validation failure

- **Given** user invokes `/moai design --mode import` AND provides an invalid bundle path
- **When** Step A3 attempts to invoke `moai-workflow-design-import`
- **Then** the import skill reports an error (per `design.md:120-123` Step A5)
- **And** the orchestrator presents AskUserQuestion options: switch to Path B1/B2, or correct the bundle path
- **And** the user can recover without restarting `--mode import`

### Edge case — `--mode import` without bundle path

- **Given** user invokes `/moai design --mode import` WITHOUT providing bundle path argument
- **When** orchestrator parses
- **Then** Step A2 AskUserQuestion still fires (collecting bundle path is unavoidable)
- **And** the user provides the path interactively
- **And** Step A3 proceeds normally

### Test mapping

- Skill body content check: `design.md` Mode Dispatch section (added in M3) explicitly documents that `--mode import` skips Phase 1 selection AND Phase B-Common.
- Manual verification: invoke `/moai design --mode import` with a valid bundle; observe trace showing Phase 1 + Phase B-Common skipped, Phase A executed.

---

## AC-WF003-16 — Subcommand × Mode matrix is published in spec-workflow.md

Maps to: REQ-WF003-007.

### Happy path

- **Given** the implementation of WF-003 M4b (matrix extension) and M5 (cross-link footers) is complete
- **When** a reviewer reads `.claude/rules/moai/workflow/spec-workflow.md` and locates the `## Subcommand Classification` section
- **Then** the section contains a matrix table that explicitly enumerates **all 10 known subcommands** (5 utility: `fix`, `coverage`, `mx`, `codemaps`, `clean`; 4 implementation: `plan`, `run`, `sync`, `design`; plus the `loop` alias row added by WF-003 M4b)
- **And** the matrix has the 3 WF-003-added columns: "Default mode", "Valid `--mode` values", "Sentinel on invalid mode"
- **And** the matrix correctly classifies:
  - **Utility 5** (`fix`, `coverage`, `mx`, `codemaps`, `clean`): pipeline-only; `--mode` value is silently ignored with `MODE_FLAG_IGNORED_FOR_UTILITY` info log (per WF-004 REQ-WF004-011)
  - **Implementation 2 mode-axis** (`run`, `design`): `--mode` axis valid values = `autopilot|loop|team` (run) or `autopilot|import|team` (design); invalid value triggers `MODE_UNKNOWN`
  - **Implementation 4 mode-NA** (`plan`, `sync`, `project`, `db`): `--mode` value is silently ignored except `--mode pipeline` which triggers `MODE_PIPELINE_ONLY_UTILITY` (per REQ-WF003-005, REQ-WF003-016)
- **And** the matrix is the single source of truth — back-referenced by all 5 mode-NA + mode-axis skill bodies (`run.md`, `design.md`, `loop.md`, `plan.md`, `sync.md`) per M5 cross-link footers

### Edge case — matrix is byte-stable across embedded template parity

- **Given** WF-003 M5 has executed `make build`
- **When** the matrix is read from the source-of-truth file `.claude/rules/moai/workflow/spec-workflow.md` AND from the template mirror `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
- **Then** both copies are byte-identical (embedded-template parity per CLAUDE.local.md §2 HARD)
- **And** `internal/template/embedded.go` regeneration reflects the matrix content

### Edge case — new mode value addition path

- **Given** a future v3.x SPEC adds a new mode value (e.g., `ultrawork`)
- **When** the matrix is updated to include the new value in the "Valid `--mode` values" column
- **Then** all skill body cross-references continue to function (back-link footer pattern preserved)
- **And** REQ-WF003-007's contract — that the matrix exists and is unified — is preserved through the addition

### Test mapping

- Static content audit: the matrix presence is verifiable via grep on `.claude/rules/moai/workflow/spec-workflow.md` for the literal section header `## Subcommand Classification` AND the literal column headers `Default mode`, `Valid \`--mode\` values`, `Sentinel on invalid mode`.
- Manual verification: open `spec-workflow.md` post-M5; confirm 10 rows are present, 3 WF-003-added columns are populated for each row, classification is internally consistent.
- Cross-spec consistency check: WF-004's `## Subcommand Classification` section (added by WF-004 M4) is the same section; WF-003 EXTENDS it (does not replace). Verify by reading the section's history comment and column count (5 base columns + 3 WF-003 = 8 total columns).

---

## AC-WF003-17 — Mode schema accepts additive future extension

Maps to: REQ-WF003-015.

### Happy path

- **Given** a future v3.x version introduces a new mode value (e.g., `ultrawork`) by editing the relevant skill bodies and the matrix in `spec-workflow.md`
- **When** the existing `--mode` consumers (orchestrator parsers in `run.md` and `design.md` Mode Dispatch sections) encounter the new value
- **Then** the parsing schema accepts the addition without breaking REQ-WF003-001's valid-set enforcement
- **And** the validation check is **documentation-driven** (skill body enumeration in the Mode Dispatch section), NOT a Go enum freeze
- **And** the new value can be added by editing 4 places coherently: (a) `run.md` and/or `design.md` Mode Dispatch enumeration; (b) `spec-workflow.md` matrix row update; (c) audit test enumeration update if a sentinel is added; (d) CHANGELOG entry
- **And** users not supplying the new mode value continue to receive existing behavior (additive guarantee preserved)

### Edge case — extension breaks valid-set check (anti-case)

- **Given** a future PR adds a new mode value to `run.md` enumeration BUT forgets to update the matrix in `spec-workflow.md`
- **When** `TestRunDesignSkillsContainModeUnknownSentinel` runs
- **Then** the audit test does NOT fail (it only checks for the literal `MODE_UNKNOWN` string presence, not the enumeration completeness)
- **But** the matrix and skill enumeration drift — caught by future plan-auditor reviews on the next SPEC, not by WF-003's audit
- **Note**: REQ-WF003-015 guarantees schema extensibility, NOT enforcement of cross-file consistency on addition. Cross-file consistency is the responsibility of the SPEC adding the new mode value.

### Edge case — extension via config (not skill body)

- **Given** a future SPEC adds support for `workflow.default_mode: ultrawork` (extending REQ-WF003-014 via M4c schema)
- **When** the orchestrator reads the config
- **Then** the validator emits `MODE_UNKNOWN` at config-load time UNLESS the new value has also been added to the skill body Mode Dispatch enumeration
- **And** the precedence rule (REQ-WF003-018 — CLI > config > harness auto) continues to apply unchanged

### Edge case — no Go-side breakage on additive extension

- **Given** the future extension only adds documentation (skill body enumeration + matrix row), no Go code changes
- **When** the existing `agentless_audit_test.go` audit suite runs
- **Then** all tests pass (audit checks for sentinel string presence, not exhaustive enumeration)
- **And** the additive guarantee (REQ-WF003-015) is preserved at the Go runtime layer

### Test mapping

- Documentation-contract verification: REQ-WF003-015 is satisfied by the **architectural choice** to use skill body enumeration (declarative) instead of a Go `enum` type (which would require code changes for additions). This choice is documented in `plan.md` §1.2 (Implementation Methodology — "no new Go runtime parser needed") and `plan.md` §4 (Technology Stack Constraints — "no new Go modules").
- Manual verification: simulate adding a hypothetical 5th mode value `ultrawork` to `run.md` Mode Dispatch enumeration + matrix row; verify that no audit test fails AND no Go code modification is required.
- Cross-reference: this AC validates the architectural-contract layer; AC-WF003-06 validates the runtime sentinel layer for invalid values.

---

## Traceability Matrix (REQ → AC)

This matrix is the canonical single-source-of-truth for REQ → AC mapping. All 18 REQs have at least one AC mapping (100% coverage). Each AC maps to one or more REQs as declared in `spec.md` §6.

| REQ ID | REQ Type | Description (short) | AC IDs |
|--------|----------|---------------------|--------|
| REQ-WF003-001 | Ubiquitous | `--mode` flag on run/design | AC-WF003-01 |
| REQ-WF003-002 | Ubiquitous | autopilot default for minimal/standard | AC-WF003-01 |
| REQ-WF003-003 | Ubiquitous | team default for thorough+enabled | AC-WF003-04 |
| REQ-WF003-004 | Ubiquitous | `/moai loop` = `/moai run --mode loop` alias | AC-WF003-03 |
| REQ-WF003-005 | Ubiquitous | plan/sync/project/db ignore `--mode` | AC-WF003-09 |
| REQ-WF003-006 | Ubiquitous | fix/coverage/mx/codemaps/clean = pipeline | AC-WF003-10 |
| REQ-WF003-007 | Ubiquitous | Subcommand × mode matrix publication | AC-WF003-16 |
| REQ-WF003-008 | Event-Driven | `--mode loop` invokes Ralph engine | AC-WF003-02 |
| REQ-WF003-009 | Event-Driven | `--mode team` design spawns parallel teammates | AC-WF003-05 |
| REQ-WF003-010 | Event-Driven | Invalid `--mode` → `MODE_UNKNOWN` | AC-WF003-06 |
| REQ-WF003-011 | Event-Driven | `--mode team` no prereqs → `MODE_TEAM_UNAVAILABLE` | AC-WF003-07 |
| REQ-WF003-012 | State-Driven | Auto-downgrade team → autopilot with info log | AC-WF003-08 |
| REQ-WF003-013 | State-Driven | `--mode import` skips Path B | AC-WF003-15 |
| REQ-WF003-014 | Optional | `workflow.default_mode` config | AC-WF003-12 |
| REQ-WF003-015 | Optional | Future mode-value extension | AC-WF003-17 |
| REQ-WF003-016 | Complex (Unwanted) | `--mode pipeline` on multi-agent → `MODE_PIPELINE_ONLY_UTILITY` | AC-WF003-10, AC-WF003-11 |
| REQ-WF003-017 | Complex (State+Event) | Ralph loop convergence terminates | AC-WF003-14 |
| REQ-WF003-018 | Complex (Unwanted) | CLI `--mode` > config > harness precedence | AC-WF003-13 |

**Coverage summary**: 18 REQs, all mapped (100%). Orphan check: zero orphans (REQ-WF003-007 → AC-16; REQ-WF003-015 → AC-17 added in iter 2).

---

## Traceability Matrix (AC → REQ)

Reverse matrix for completeness — verifies every AC traces back to at least one REQ.

| AC ID | Mapped REQs | Notes |
|-------|-------------|-------|
| AC-WF003-01 | REQ-WF003-001, REQ-WF003-002 | Autopilot dispatch + harness-default |
| AC-WF003-02 | REQ-WF003-008 | Ralph engine invocation |
| AC-WF003-03 | REQ-WF003-004 | Alias contract |
| AC-WF003-04 | REQ-WF003-003 | Team auto-select |
| AC-WF003-05 | REQ-WF003-009 | Design team mode parallel teammates |
| AC-WF003-06 | REQ-WF003-010 | MODE_UNKNOWN sentinel |
| AC-WF003-07 | REQ-WF003-011 | MODE_TEAM_UNAVAILABLE sentinel |
| AC-WF003-08 | REQ-WF003-012 | Auto-downgrade info log |
| AC-WF003-09 | REQ-WF003-005 | plan/sync mode ignore (covers plan/sync; project/db are skill-absent per research.md §2.2.4) |
| AC-WF003-10 | REQ-WF003-006, REQ-WF003-016 | Utility pipeline + pipeline rejection cross-ref |
| AC-WF003-11 | REQ-WF003-016 | MODE_PIPELINE_ONLY_UTILITY sentinel |
| AC-WF003-12 | REQ-WF003-014 | Config default_mode |
| AC-WF003-13 | REQ-WF003-018 | CLI > config precedence |
| AC-WF003-14 | REQ-WF003-017 | Ralph convergence termination |
| AC-WF003-15 | REQ-WF003-013 | Design import path A skip |
| AC-WF003-16 | REQ-WF003-007 | Matrix publication contract (added iter 2) |
| AC-WF003-17 | REQ-WF003-015 | Future-extension schema contract (added iter 2) |

**Coverage summary**: 17 ACs, all reverse-mapped. AC-WF003-09 has a documented scope footnote: REQ-WF003-005 enumerates 4 subcommands (plan, sync, project, db); AC-09 happy path covers plan, edge case covers sync. project/db skill bodies do not exist (per research.md §2.2.4) — REQ-005's enumeration is satisfied by the matrix row classification in `spec-workflow.md` (validated by AC-16) rather than a per-skill audit.

---

## Quality Gate Hooks (cross-reference)

### TRUST 5 framework alignment

- **Tested**: AC-WF003-06, 07, 03 enforce sentinel + alias presence via static audit. AC-WF003-11 inherits WF-004's existing audit. All 15 ACs have explicit test mapping.
- **Readable**: Mode Dispatch sections (M2, M3, M4a, M4b, M4d) use consistent template wording for clarity. Cross-references unify the dispatch contract.
- **Unified**: Embedded-template parity (`make build` after every skill/rule/yaml edit per `CLAUDE.local.md` §2 Template-First Rule) ensures source and embedded copies remain in sync.
- **Secured**: No new attack surface — this SPEC only adds documentation sections, audit tests, and a yaml schema field. No runtime behavior change for users not supplying `--mode`.
- **Trackable**: CHANGELOG entry (M5) + commit messages per Conventional Commits + audit tests provide audit trail.

### LSP quality gates

Per `.moai/config/sections/quality.yaml lsp_quality_gates`:
- **plan phase**: `require_baseline: true` — captured at this SPEC's plan completion (current commit `5ab409292`).
- **run phase**: `max_errors: 0`, `max_type_errors: 0`, `max_lint_errors: 0`, `allow_regression: false` — extended `agentless_audit_test.go` must compile cleanly; no new lint warnings introduced.
- **sync phase**: `max_warnings: 10`, `require_clean_lsp: true` — verified post-M5 by running `golangci-lint run` and `go vet ./...`.

### Pre-submission self-review checklist

Per `.claude/rules/moai/workflow/spec-workflow.md:79-80`:

- [ ] Full diff reviewed against this acceptance.md before commit.
- [ ] Asked "Is there a simpler approach?" — answer documented in plan.md §1.3 (this is the simpler approach: declarative mode dispatch + static audit, not a Go runtime mode parser).
- [ ] Asked "Would removing any changes still satisfy the SPEC?" — minimum set is M1 (audit tests) + M2/M3 (run/design dispatch sections) + M4b (matrix extension); M4a (loop alias note), M4c (default_mode yaml), M4d (plan/sync mode-NA), M5 are necessary for spec completeness but could be split if needed.
- [ ] Verified stacked PR base transition hook is documented in plan.md §1.4 and §7.

---

## Definition of Done (DoD)

The implementation is "done" when ALL of the following are true:

1. ✅ Three new test functions in `internal/template/agentless_audit_test.go` (`TestRunDesignSkillsContainModeUnknownSentinel`, `TestRunSkillContainsModeTeamUnavailableSentinel`, `TestLoopAliasCrossReference`) all PASS.
2. ✅ `.claude/skills/moai/workflows/run.md` contains a `## Mode Dispatch (Multi-Mode Router)` section with all 4 mode values documented + `MODE_UNKNOWN` + `MODE_TEAM_UNAVAILABLE` + `MODE_PIPELINE_ONLY_UTILITY` (preserved from WF-004) sentinels.
3. ✅ `.claude/skills/moai/workflows/design.md` contains a `## Mode Dispatch (Multi-Mode Router)` section with `autopilot|import|team|pipeline` values + `MODE_UNKNOWN` + `MODE_PIPELINE_ONLY_UTILITY` (preserved) sentinels.
4. ✅ `.claude/skills/moai/workflows/loop.md` contains a `## Invocation Routes (SPEC-V3R2-WF-003)` section documenting the alias relationship with literal phrase `/moai run --mode loop`.
5. ✅ `.claude/rules/moai/workflow/spec-workflow.md` `## Subcommand Classification` section is extended with 3 new columns + 1 new `/moai loop` row + a `### Mode Dispatch Cross-Reference` sub-section documenting the 4 sentinels and precedence rule.
6. ✅ `.moai/config/sections/workflow.yaml` contains a new optional `default_mode: ""` field with comment.
7. ✅ `.claude/skills/moai/workflows/plan.md` and `.claude/skills/moai/workflows/sync.md` `## Mode Flag Compatibility` sections are refined to clarify REQ-WF003-005 (silent ignore except `pipeline`).
8. ✅ `internal/template/templates/.claude/...` and `.../.moai/config/sections/workflow.yaml` mirror all 8 source-of-truth edits (embedded template parity per CLAUDE.local.md §2 HARD).
9. ✅ `make build` regenerates `internal/template/embedded.go` cleanly.
10. ✅ Full repository test suite passes: `go test ./...` returns 0 (per `CLAUDE.local.md` §6 HARD rule).
11. ✅ CHANGELOG `## [Unreleased]` section has the SPEC-V3R2-WF-003 entry.
12. ✅ MX tags per plan.md §6 inserted in all 6 target locations (8 tags total).
13. ✅ `progress.md` updated with `run_complete_at` and `run_status: implementation-complete`.
14. ✅ Stacked PR base transitioned from `feature/SPEC-V3R2-WF-004` → `main` BEFORE PR #765 merges (orthogonal to milestones; documented in plan.md §1.4).

---

End of acceptance.md.
