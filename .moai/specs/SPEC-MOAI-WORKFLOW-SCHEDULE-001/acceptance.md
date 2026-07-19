# Acceptance Criteria — SPEC-MOAI-WORKFLOW-SCHEDULE-001

Given-When-Then acceptance matrix for MoAI-Workflow. Each AC-MWS-xxx maps to one or more REQ-MWS-xxx in spec.md §B.

## §D Acceptance Matrix

### §D.1 File format and storage

**AC-MWS-001** (REQ-MWS-001, 002, 005)
- **Given** a user creates a workflow named `drift-watch`,
- **When** the creation flow completes,
- **Then** a file exists at `.moai/workflows/drift-watch.md` with a YAML frontmatter block declaring at minimum `name`, `description`, `schedule`, and `safety`, and a Markdown body containing natural-language step instructions.

**AC-MWS-002** (REQ-MWS-003)
- **Given** a workflow frontmatter with a `schedule` field,
- **When** the frontmatter is parsed,
- **Then** the `schedule` carries both an interval-or-cron expression and a `mechanism` value that is exactly `cron` or `loop`.

**AC-MWS-003** (REQ-MWS-004, 006)
- **Given** a workflow frontmatter,
- **When** the `safety` field is present, **Then** its value is `read-only` or `write`;
- **And** when the `safety` field is absent, the workflow is treated as `read-only`.

### §D.2 Creation entry surface

**AC-MWS-004** (REQ-MWS-007)
- **Given** a user invokes `/moai workflow` with creation intent,
- **When** the orchestrator runs the guided flow,
- **Then** it captures `name`, `description`, `schedule` (expression + mechanism), `safety`, and the step body via AskUserQuestion rounds, and writes the file to `.moai/workflows/<name>.md`.

**AC-MWS-005** (REQ-MWS-008)
- **Given** a candidate workflow with an invalid `mechanism` value (e.g. `mechanism: daemon`) or a malformed schedule expression,
- **When** the creation flow validates before writing,
- **Then** the file is NOT written and the orchestrator surfaces the validation failure for correction.

**AC-MWS-006** (REQ-MWS-009)
- **Given** a user issues a natural-language capture request ("save this as a recurring nightly workflow") without typing `/moai workflow`,
- **When** the orchestrator routes it,
- **Then** it reaches the same guided-capture path and produces an equivalent workflow file (single creation code-path).

### §D.3 Schedule registration and unregistration

**AC-MWS-007** (REQ-MWS-010)
- **Given** a new workflow with `mechanism: cron`,
- **When** creation completes,
- **Then** the orchestrator invokes CronCreate with the declared expression and reports the registration using interval + mechanism vocabulary.

**AC-MWS-008** (REQ-MWS-011)
- **Given** a new workflow with `mechanism: loop`,
- **When** creation completes,
- **Then** the orchestrator guides the user to arm the session-scoped `/loop` and discloses that the `loop` schedule is session-scoped and must be re-armed each session (no persistent registration);
- **And** the workflow file records the declared loop schedule (record-only), and at the next session start the orchestrator surfaces a re-arm reminder while the user re-arms `/loop` manually (no orchestrator auto-arm) — per DECISION-MWS-D2 (plan.md §D).

**AC-MWS-009** (REQ-MWS-012, 013)
- **Given** an existing `mechanism: cron` workflow,
- **When** the user removes it,
- **Then** the orchestrator invokes CronDelete (unregister) BEFORE deleting the file, and the unregister notice names the mechanism; for a `mechanism: loop` workflow the notice names session-scoped loop cancellation instead.

**AC-MWS-010** (REQ-MWS-013, Cron-unavailable fallback)
- **Given** a runtime where Cron tools are unavailable,
- **When** a `mechanism: cron` workflow is created,
- **Then** the orchestrator degrades to guiding the user toward the session-scoped `/loop` form rather than failing, consistent with the cadence-bridge fallback clause.

### §D.4 Safety boundary (cadence invariant) — must-pass

**AC-MWS-011** (REQ-MWS-014, HARD)
- **Given** any scheduled workflow run (any `safety` tier),
- **When** the scheduled turn executes,
- **Then** it performs no commit, no push, and no run-phase entry.

**AC-MWS-012** (REQ-MWS-015)
- **Given** a scheduled workflow run whose steps would touch the working tree,
- **When** it executes,
- **Then** at most Level-1 uncommitted working-tree edits are made and they are left uncommitted and unpushed.

**AC-MWS-013** (REQ-MWS-016) — must-pass
- **Given** a workflow declaring `safety: write`,
- **When** it runs on the schedule (unattended),
- **Then** the `write` tier does NOT relax the cadence invariant — the scheduled run remains bounded by AC-MWS-011/012; the `write` permission applies only to an interactive (user-present) invocation.

**AC-MWS-014** (REQ-MWS-017)
- **Given** a workflow step that would require an Implementation Kickoff Approval-class human gate,
- **When** a scheduled run reaches that step,
- **Then** the scheduled run does not satisfy the gate; it queues the pending decision and surfaces it at the next interactive session.

### §D.5 Discovery and lifecycle

**AC-MWS-015** (REQ-MWS-018)
- **Given** two workflows exist (one with a schedule, one without),
- **When** the user requests the list,
- **Then** the orchestrator enumerates `.moai/workflows/*.md` and renders each name + description + schedule (interval + mechanism); the schedule-less workflow renders without schedule detail (no error).

**AC-MWS-016** (REQ-MWS-019)
- **Given** an existing workflow,
- **When** the user requests to edit it,
- **Then** the orchestrator opens the workflow Markdown file for direct modification, treating the frontmatter as SSOT.

**AC-MWS-017** (REQ-MWS-020)
- **Given** a removal request for a workflow that does not exist,
- **When** the orchestrator processes it,
- **Then** it reports a no-op rather than failing hard; **and** for an existing workflow, removal unregisters the schedule then deletes the file.

### §D.6 Template deployment

**AC-MWS-018** (REQ-MWS-021)
- **Given** the template source tree,
- **When** `internal/template/templates/.moai/workflows/` is inspected,
- **Then** it contains a `README.md` explaining the workflow file format and exactly one neutral example workflow file.

**AC-MWS-019** (REQ-MWS-022) — must-pass
- **Given** the template scaffold content,
- **When** the template-neutrality CI guard (`template-neutrality-check.yaml`) runs,
- **Then** the scaffold contains no internal SPEC IDs, no internal dates, and no commit SHAs, and the guard passes.

**AC-MWS-020** (REQ-MWS-023)
- **Given** a change to the template scaffold,
- **When** the change is committed,
- **Then** the source under `internal/template/templates/.moai/workflows/` was edited first and embedded via `make build` (no local-only drift).

### §D.7 Boundary / non-duplication

**AC-MWS-021** (REQ-MWS-024)
- **Given** the four sibling assets (harness-v4 schedule, `.claude/workflows/*.js`, native `/loop`, cadence-bridge recipes),
- **When** the boundary section (spec.md §C) is reviewed,
- **Then** MoAI-Workflow is documented as the distinct user-facing markdown layer, and no sibling asset's behavior is duplicated or forked (vocabulary is reused, not re-invented).

## §D.8 Edge Cases

- **EC-1**: Workflow name with path separators or `..` — rejected at creation (no directory traversal outside `.moai/workflows/`).
- **EC-2**: Empty body (frontmatter only) — rejected; a workflow must carry step instructions (REQ-MWS-005).
- **EC-3**: `mechanism: cron` with a malformed cron expression — validation failure (AC-MWS-005).
- **EC-4**: Removing a workflow whose Cron registration was already externally deleted — CronDelete no-op, file still deleted, reported cleanly.
- **EC-5**: Name collision on create — the orchestrator rejects the creation and re-prompts the user for a different name; the existing file and its registered schedule are never overwritten or auto-suffixed (DECISION-MWS-D1, plan.md §D).

## §D.9 Definition of Done

- [ ] All 24 REQ-MWS map to ≥1 passing AC-MWS.
- [ ] Must-pass ACs (AC-MWS-011, 013, 019) all pass.
- [ ] The two plan.md §D decisions (DECISION-MWS-D1 name-collision, DECISION-MWS-D2 loop re-arm) are resolved (user-confirmed 2026-07-17) and reflected in REQ/AC wording.
- [ ] `/moai workflow` guided creation works end-to-end for both `cron` and `loop` mechanisms.
- [ ] Template scaffold deployed, neutral, embedded via `make build`, CI-neutral.
- [ ] Boundary (spec.md §C) documented; no duplication of sibling assets.
- [ ] Cadence read-only invariant demonstrably enforced for scheduled runs (AC-MWS-011/012/013).
