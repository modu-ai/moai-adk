---
id: SPEC-HARNESS-CADENCE-BUILD-001
title: "Harness Builder Cadence Integration — build-time recurrence, discovery-queue scheduling, ANALYZE research, retroactive schedules"
version: "0.1.2"
status: in-progress
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: High
phase: "v3.0.0 target"
module: "internal/harness/v4manifest, internal/cli/harness, .claude/skills/moai/workflows, .claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "harness, cadence, scheduling, discovery-queue, builder, manifest, research"
tier: M
related_specs: [SPEC-CADENCE-BRIDGE-001, SPEC-V3R6-HARNESS-V4-001, SPEC-HARNESS-EVO-PIPE-REPAIR-001]
---

# SPEC-HARNESS-CADENCE-BUILD-001 — Harness Builder Cadence Integration

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-13 | manager-spec | Initial draft — 4 capabilities (build-time recurrence question, discovery-queue scheduled execution, ANALYZE research sub-step, retroactive schedule path) under the preserved cadence-bridge HARD invariant |
| 0.1.1 | 2026-07-13 | manager-spec | Plan-audit iter-1 revision — 3 clarifications RESOLVED via orchestrator AskUserQuestion (registration-only thin-harness / gate-round question placement / minimal 3-field schedule; recorded in plan.md §B); D2 AC-034 invocation-surface scoping, D3 AC-023 queue-path delta-frame, D4 new AC-HCB-002, D5 REQ-044 formal modal, D6 AC-025 severity align, D8 awk-bounded windowed greps + pinned anchors, D9 Retrofit-precedence pin |
| 0.1.2 | 2026-07-13 | manager-spec | Plan-audit iter-2 revision (PASS-WITH-DEBT 0.90) — N1 (major): AC-034 clause-1 boundary grep delta-framed against the re-measured baseline of 5 pre-existing help-text matches (requirement = no NEW matches) + plan §E E4 mirror + exemption extended to documentation strings; minors swept: N2 AC-030 smoke-gate→CronCreate ordering mechanized (awk ACTIVATE window), N3 AC-051 ordering tokens pinned (recurrence/CronCreate), N4 awk flat-window + ordering-token authoring constraints recorded in plan M2 |

## §A Context and Problem

The v4 harness Builder (`.claude/skills/moai/workflows/harness-build-entry.md` + `harness-builder.md`, orchestrator-direct ANALYZE / PLAN / GENERATE / ACTIVATE) builds one-shot harnesses: every `/harness:<name>` run is user-initiated. Separately, `.claude/rules/moai/workflow/cadence-bridge.md` (SPEC-CADENCE-BRIDGE-001) established a sanctioned recipe catalog composing native `/loop` and Cron tools with read-only `/moai` entry points, bound by a catalog-level HARD invariant: **scheduled runs never commit, never push, never enter run-phase**. Nothing today connects the two — a built harness cannot declare, register, surface, or unregister a recurring schedule, the Builder never asks the recurrence question, and the ANALYZE phase scans only the local codebase (no official-docs / best-practice / library-docs research feeds PLAN).

This SPEC integrates cadence into the Builder under the **user-confirmed discovery-queue model**: scheduled harness runs execute in discovery-only mode (read-only analysis; findings persisted to a queue surface; surfaced at the next interactive session; any write work requires human approval in an interactive session). The cadence-bridge catalog invariant is explicitly PRESERVED, not weakened.

Schedule declaration SSOT: the optional `schedule` object in the harness `manifest.json` (`.claude/commands/harness/<name>/manifest.json`), decoded by `internal/harness/v4manifest` and consumed by the `moai harness` v4 lifecycle verbs (`internal/cli/harness/v4lifecycle.go`, `doctor.go`).

### §A.1 Verified baseline anchors (measured 2026-07-13)

- `internal/harness/v4manifest/types.go` — 8-field `Manifest` struct; JSON decoder tolerates unknown fields (the `mcp` block precedent relies on this).
- `internal/harness/v4manifest/validate.go` — `func Validate(m Manifest) error` (extension point).
- `internal/cli/harness/v4lifecycle.go` — `ListHarnesses` / `EditHarness` / `RemoveHarness`; `HarnessEntry` carries `Domain` + `EntryCommand`; cobra wrappers in `internal/cli/harness/v4lifecycle_cmd.go`.
- `internal/cli/harness/doctor.go` — 4-axis reference-integrity gate; axis 2 reuses `v4manifest.Validate`, so Validate extensions propagate to doctor with no doctor code change; severities are ERROR|INFO only; command-only thin harnesses (github/release) are INFO, never ERROR.
- `cadence-bridge.md` — Recipes 1-3; invariant stated ONCE at catalog level, deliberately not restated per-recipe; Discovery-to-Queue Contract (active TaskList when live, else `.moai/reports/cadence/<date>.md`); fallback clause "Cron tools unavailable → degrade to native `/loop`".
- All four target markdown surfaces (cadence-bridge.md, harness-build-entry.md, harness-builder.md, moai SKILL.md) are currently byte-identical to their `internal/template/templates/` mirrors (measured PARITY ×4).

## §B Requirements (GEARS)

### Group 1 — Build-time recurrence question (REQ-HCB-001..005)

- **REQ-HCB-001**: **When** the Builder PLAN phase presents the PLAN→GENERATE approval gate, the orchestrator shall include a recurrence question in the same orchestrator-issued AskUserQuestion round, asking whether this harness should run on a recurring schedule. (Subagents never prompt; the orchestrator holds the gate boundary — existing design.)
- **REQ-HCB-002**: **When** the user confirms recurrence, the orchestrator shall capture the interval and the mechanism preference (native `/loop` vs Cron tools CronCreate/CronList/CronDelete), with option descriptions stating the trade-off: `/loop` arming is session-scoped (dies with the session, re-armed per session); Cron registration is persistent across sessions.
- **REQ-HCB-003**: **When** recurrence is confirmed, the PLAN phase shall record a `schedule` object (`interval`, `mechanism`, `mode: "discovery-only"`) in the draft manifest before GENERATE writes Artifact 5 (manifest.json).
- **REQ-HCB-004**: **When** the user declines recurrence, the Builder shall omit the `schedule` field entirely, and the generated manifest shall be shape-identical to the pre-change baseline (no empty/null schedule key).
- **REQ-HCB-005**: The recurrence question's option descriptions shall state that scheduled runs execute in discovery-only mode: read-only analysis, findings persisted to a queue surface, no writes/commits/pushes, no run-phase entry.

### Group 2 — Manifest `schedule` schema (REQ-HCB-010..014)

- **REQ-HCB-010**: The v4 manifest schema shall gain an OPTIONAL top-level `schedule` object carrying exactly three required sub-fields when present: `interval` (non-empty string — e.g. `"30m"`, `"nightly"`, or a cron expression), `mechanism` (enum: `"loop"` | `"cron"`), `mode` (literal: `"discovery-only"`).
- **REQ-HCB-011**: **While** a manifest omits `schedule`, `v4manifest.Validate` shall accept it unchanged — every pre-existing valid manifest remains valid (backward compatibility; additive-only schema change).
- **REQ-HCB-012**: **When** a manifest declares `schedule` with `mode` not equal to `"discovery-only"`, `v4manifest.Validate` shall reject it with an error naming the discovery-only invariant.
- **REQ-HCB-013**: **When** a manifest declares `schedule` with an unknown `mechanism` or an empty `interval`, `v4manifest.Validate` shall reject it.
- **REQ-HCB-014**: **Where** a schedule is declared, the `mode` field shall be an explicit machine-checkable invariant marker — declared verbatim in the manifest, never defaulted or inferred by the decoder — so any future write-mode proposal fails validation loudly instead of passing silently.

### Group 3 — Discovery-queue scheduled execution (sanctioned recipe) (REQ-HCB-020..025)

- **REQ-HCB-020**: The cadence-bridge recipe catalog shall gain a new sanctioned recipe class, "Scheduled Harness Discovery" (Recipe 4), covering scheduled discovery runs for any built harness — additive to the existing catalog.
- **REQ-HCB-021**: Recipe 4's scheduled payload shall be a **self-contained discovery prompt** (naming the harness and its domain scope, and embedding the read-only / queue-persistence / no-commit-push / no-run-phase constraints inline), NOT the raw `/harness:<name>` entry command — so a scheduled turn carries its constraints even when the rule file is not loaded, and the write-capable Runner specialist dispatch is never invoked by a scheduled run.
- **REQ-HCB-022**: The catalog-level HARD invariant sentence shall remain unmodified, and Recipe 4 shall NOT restate it — preserving the catalog's single-statement invariant design.
- **REQ-HCB-023**: **When** a scheduled harness discovery run finds work, it shall persist findings per the existing Discovery-to-Queue Contract (active TaskList when a session ledger is live, otherwise `.moai/reports/cadence/<date>.md`) and surface them at the next interactive session — the queue mechanism is reused verbatim, no new queue surface is introduced.
- **REQ-HCB-024**: The cadence-bridge eligibility table shall gain a row classifying the harness discovery prompt as cadence-safe, with its read-only/advisory rationale.
- **REQ-HCB-025**: **Where** Cron tools are unavailable in the runtime version, Recipe 4 shall degrade to native `/loop` only, consistent with the catalog's existing fallback clause.

### Group 4 — ACTIVATE registration + lifecycle verb awareness (REQ-HCB-030..035)

- **REQ-HCB-030**: **While** the manifest declares a schedule, **When** the ACTIVATE phase's reference-integrity smoke gate (`moai harness doctor`) reports zero ERROR findings and the user approves activation, the orchestrator shall register the schedule: `mechanism: "cron"` → issue CronCreate carrying the Recipe 4 discovery prompt; `mechanism: "loop"` → emit the paste-ready `/loop <interval> <discovery prompt>` line for the user, with the session-scoped caveat stated.
- **REQ-HCB-031**: **When** `moai harness list` renders a harness whose manifest declares a schedule, it shall surface the schedule (interval + mechanism) in both the human-readable and `--json` outputs.
- **REQ-HCB-032**: **When** `moai harness remove <name>` removes a harness whose manifest declared a schedule, it shall emit an unregister notice naming the declared mechanism (CronDelete for `cron`; loop cancellation for `loop`); the notice content shall be computed from the manifest BEFORE deletion.
- **REQ-HCB-033**: **When** `moai harness doctor` scans a harness whose manifest declares a schema-invalid schedule (REQ-HCB-012/013), it shall report an ERROR-severity finding (achieved via the existing axis-2 `v4manifest.Validate` reuse) and exit non-zero.
- **REQ-HCB-034**: The `moai` Go CLI shall NOT invoke Cron tools, arm loops, or prompt the user; registration and unregistration are orchestrator-side actions — the CLI emits structured output only (existing subagent/CLI boundary preserved).
- **REQ-HCB-035**: **While** a harness manifest omits `schedule`, the `list` / `edit` / `remove` / `doctor` outputs shall remain identical to the pre-change baseline (no schedule-related noise for unscheduled harnesses; `--json` omits the field).

### Group 5 — Build-time research sub-step in ANALYZE (REQ-HCB-040..044)

- **REQ-HCB-040**: The Builder ANALYZE phase shall gain an external-research sub-step alongside the existing codebase fan-out, covering: official Claude Code documentation (WebFetch/WebSearch), domain best practices (WebSearch with URL verification), and library documentation via context7 MCP (`resolve-library-id` then `query-docs`).
- **REQ-HCB-041**: The research sub-step's findings shall feed the PLAN phase's manifest draft — pattern selection, specialist role definitions, and companion-skill content — as part of the ANALYZE aggregate the PLAN sub-agent reasons over.
- **REQ-HCB-042**: **When** context7 MCP or web tools are unavailable or fail, the research sub-step shall degrade gracefully per the MCP Fallback Strategy (detect → inform → WebFetch fallback where possible → established best-practice patterns → continue) — a harness build never blocks on research availability.
- **REQ-HCB-043**: **While** the session is GLM-backed, the research sub-step shall route web search / web fetch through the z.ai MCP tools per the GLM web-tooling routing table instead of the built-in WebSearch/WebFetch.
- **REQ-HCB-044**: **Where** the confirmed domain is purely internal (no external libraries or external docs are relevant), the orchestrator MAY skip the research sub-step; **When** the skip is taken, the orchestrator shall record the skip rationale — consistent with the ANALYZE phase's existing load-bearing-minimum collapse policy.

### Group 6 — Retroactive schedule path for existing harnesses (REQ-HCB-050..053)

- **REQ-HCB-050**: **When** a natural-language `harness` request references an EXISTING harness (the name resolves to `.claude/commands/harness/<name>.md`) together with scheduling intent, the build-entry workflow shall route to a Schedule Retrofit branch instead of the Builder creation pipeline. Retrofit detection shall be evaluated BEFORE the build-entry name-collision handling: an existing-name + scheduling-intent request routes to the Retrofit branch, never to the collision re-derive path (`<name>-v2` / rename).
- **REQ-HCB-051**: The Schedule Retrofit branch shall run the same recurrence question round (REQ-HCB-002/005); **When** the target harness is manifest-bearing, the orchestrator shall apply an orchestrator-mediated edit adding the `schedule` object to its manifest.json, then register per REQ-HCB-030 (with `moai harness edit` as the path-discovery surface).
- **REQ-HCB-052**: **When** the target harness is command-only (no manifest.json — e.g. the github/release maintainer harnesses, whose Runner-less shape is a deliberate design), the Retrofit branch shall register the schedule via the Recipe 4 discovery prompt WITHOUT fabricating a manifest, and shall inform the user that `list`/`doctor` schedule surfacing is unavailable for manifest-less harnesses.
- **REQ-HCB-053**: The Retrofit branch shall not cause dev-only harness artifacts (release-update/github/release and their manifests) to appear under `internal/template/templates/` — the dev-only isolation contract is preserved.

### Group 7 — Template mirroring and neutrality (REQ-HCB-060..062)

- **REQ-HCB-060**: Every changed template-managed markdown surface (cadence-bridge.md, harness-build-entry.md, harness-builder.md, moai SKILL.md) shall be mirrored byte-identically to its `internal/template/templates/` counterpart (Template-First rule; all four pairs are byte-identical at baseline).
- **REQ-HCB-061**: The added doctrine text shall carry NO internal SPEC IDs, REQ tokens, or audit citations — in the template mirrors AND in the live files (byte-identity makes the live text template content; template content neutrality binds both).
- **REQ-HCB-062**: **When** template files change, `make build` shall be run so the embedded template FS matches the source tree, and the template CI guards (neutrality + leak tests) shall pass.

## §C Non-Functional Constraints

- **C-1 (Frozen boundaries)**: The Implementation Kickoff Approval human gate, the AskUserQuestion channel monopoly, and the cadence-bridge catalog HARD invariant are untouchable. No requirement in this SPEC weakens them; REQ-HCB-022 pins the invariant's textual preservation.
- **C-2 (Additive-only schema)**: The `schedule` field is optional and additive. No existing manifest, fixture, or generated artifact becomes invalid.
- **C-3 (CLI boundary)**: `internal/cli/harness/**` continues to never prompt the user (existing boundary test coverage extends to new code paths).
- **C-4 (No scheduler mechanics in Go)**: The moai binary gains no daemon, timer, or cron engine. Scheduling mechanics remain native `/loop` + Cron tools; the binary only declares/validates/surfaces schedule metadata.
- **C-5 (16-language neutrality)**: No template-mirrored change privileges a project language; the discovery prompt template and recipe text are language-neutral.
- **C-6 (No time estimates)**: Plan artifacts use priority labels only.

## §D Acceptance Criteria Structure

The full AC matrix (AC-HCB-001..073) with per-AC verification commands, severity, and traceability to REQ groups lives in `acceptance.md` §D. Structure: Go-verifiable ACs (schema/CLI behavior via `go test` fixtures), doctrine-text ACs (grep both trees — for prose deliverables the text IS the deliverable, but cross-file reachability ACs pin the routing seams: recipe → queue contract, builder gate → manifest field, retrofit branch → CLI verb), and repo-global gates (`make build`, `go test ./...`, `golangci-lint`, `moai harness doctor`, `moai spec lint`).

## Exclusions

### Out of Scope — Mechanical enforcement of the discovery-only invariant

- No hook, daemon, or runtime blocker mechanically prevents a scheduled turn from writing. Enforcement stays doctrine-level, consistent with cadence-bridge's existing "mechanical blocking … is out of scope for this doctrine-only bridge" stance. A future enforcement SPEC may add it.

### Out of Scope — Go-side scheduler or schedule-state registry

- No cron daemon, timer loop, or `.moai/state/` registry of armed loops/crons in the moai binary. Registration state lives where the mechanism lives (Claude Code Cron store; session-scoped /loop). `moai harness list` surfaces the DECLARED schedule, not live registration state.

### Out of Scope — Runner/command discovery-mode branch

- No `discover` argument branch in the generated entry command (`v4manifest.CommandTemplate`) or Runner JS. The scheduled payload is the self-contained Recipe 4 discovery prompt (REQ-HCB-021); the Runner's specialist dispatch is deliberately never invoked by a scheduled run. Adding an interactive `discover` verb to generated commands is deferred.

### Out of Scope — harness-spec.yaml recurrence axis

- Extending `.moai/project/harness-spec.yaml` (the `/moai project` Phase 3.2 bridge artifact) with a recurrence field is deferred to a follow-up of the project-harness bridge line. The Builder asks the recurrence question live; no pre-satisfaction from harness-spec.yaml in this SPEC.

### Out of Scope — Write-capable scheduled runs

- Any scheduled invocation that commits, pushes, or enters run-phase. Explicitly rejected by the user's confirmed discovery-queue model decision; the catalog invariant is preserved verbatim.

### Out of Scope — Schedule schema for the learning-subsystem manifest.jsonl

- The `internal/harness` learning-subsystem lineage (`manifest.jsonl`) and `internal/manifest` (template-deployment manifest) are untouched; only `internal/harness/v4manifest` changes.

## §E Cross-References

- `.claude/rules/moai/workflow/cadence-bridge.md` — catalog invariant, eligibility table, Discovery-to-Queue Contract, fallback clauses (Recipe 4 lands here).
- `.claude/skills/moai/workflows/harness-build-entry.md` — Phases 0-3 + Retrofit branch landing site.
- `.claude/skills/moai/workflows/harness-builder.md` — ANALYZE/PLAN/GENERATE/ACTIVATE + Artifact 5 content contract.
- `.claude/skills/moai/SKILL.md` § harness Branch A.1 — verb-description alignment (list/remove schedule awareness).
- `internal/harness/v4manifest/{types,validate}.go`, `internal/cli/harness/{v4lifecycle,v4lifecycle_cmd,doctor}.go` — Go extension points.
- `.claude/rules/moai/core/agent-common-protocol.md` § MCP Fallback Strategy; `.claude/rules/moai/core/glm-web-tooling.md` — research sub-step degradation/routing.
- `.claude/rules/moai/workflow/goal-directive.md` — native `/loop` vs `/moai loop` distinctness (consumed, not modified).
