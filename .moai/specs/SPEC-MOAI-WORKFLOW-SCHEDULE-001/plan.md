# Implementation Plan — SPEC-MOAI-WORKFLOW-SCHEDULE-001

> Milestones are ordered by **decision-reversibility**: the highest-change-likelihood decisions (data model, user-facing surface) lead; mechanical/scaffolding steps trail. Human review should focus on the top milestones.

## §A Context

MoAI-Workflow adds a user-facing simple layer for saving recurring-task workflows as Markdown files under `.moai/workflows/` and registering them on the Claude Code native scheduler. No new execution engine — the feature is markdown + orchestration + template scaffold. The three fixed design decisions (native-scheduler-only, markdown+frontmatter format, cadence read-only default) are settled in spec.md §A.1 and are NOT reopened here.

## §B Known Constraints and Precedents

- **cadence read-only invariant is HARD** (`cadence-bridge.md`): scheduled runs never commit / push / enter run-phase; Level-1 uncommitted edits are the sole write exception. This binds REQ-MWS-014..017.
- **Schedule vocabulary is fixed** (harness-v4 Branch A.1): reuse `interval + mechanism`, `CronCreate`/`CronList`/`CronDelete`, session-scoped loop cancellation. Do NOT fork a new vocabulary (REQ-MWS-013).
- **Cron tools may be unavailable** in older runtime versions — degrade `mechanism: cron` guidance to the session-scoped `/loop` form, consistent with the cadence-bridge fallback clause.
- **Template-First** (`CLAUDE.local.md` §2): template scaffold source lands in `internal/template/templates/.moai/workflows/` FIRST, then `make build`.
- **Template neutrality** (`§25.1`): scaffold must carry no internal SPEC IDs/dates/SHAs.

## §C Decisions Made (stated explicitly, not deferred to run-phase)

1. **Entry surface = `/moai workflow` guided creation** (orchestrator + AskUserQuestion), routed through a new skill workflow file `.claude/skills/moai/workflows/workflow.md`. A natural-language capture request routes to the same guided-capture code-path (REQ-MWS-009). Per §27.3 a `/moai:workflow` wrapper command is added alongside.
2. **Go CLI `moai workflow` subcommand is DEFERRED** to a follow-up SPEC (spec.md §E "Out of Scope — `moai workflow` Go CLI subcommand"). Lifecycle (list/edit/remove) is orchestrator-driven filesystem ops in this SPEC, because Cron registration is orchestrator-only anyway.
3. **`safety` tier governs interactive invocation only**; scheduled runs are ALWAYS cadence-bounded regardless of tier (REQ-MWS-016). This resolves the apparent "what does a scheduled `write` workflow do?" tension cleanly without a new enum.
4. **Frontmatter validation is orchestrator-side at creation** (REQ-MWS-008); no mechanical Go validator (spec.md §E).

## §D Decisions Resolved

Both plan-phase open questions were resolved via `AskUserQuestion` (final, user-confirmed 2026-07-17). They are recorded here as DECISION entries; no `[NEEDS CLARIFICATION]` marker remains.

- **DECISION-MWS-D1 (name-collision policy on workflow creation) — REJECT + re-prompt guidance.** _(2026-07-17)_ When a user creates a workflow whose `<name>.md` already exists, the orchestrator rejects the creation and guides the user to choose a different name. The existing file and its registered schedule are NEVER silently overwritten or destroyed. Auto-suffix (`<name>-2.md`) is NOT used. **Rationale**: overwrite silently destroys a registered schedule (data loss), and auto-suffix produces a name the user did not intend while leaving the collision unresolved; reject-and-re-prompt is the only option that preserves the existing workflow and keeps the user in control of naming. Affects REQ-MWS-007 / AC-MWS-005 / EC-5.
- **DECISION-MWS-D2 (session-scoped `loop` re-arm responsibility) — RECORD-ONLY + session-start reminder.** _(2026-07-17)_ For `mechanism: loop`, the workflow file records the declared loop schedule (record-only); no persistent registration is created (a `/loop` dies with its session). After `/clear` or on a new session, the orchestrator surfaces a "re-arm needed" reminder at session start for any `mechanism: loop` workflow present, and the USER re-arms `/loop` manually. The orchestrator does NOT auto-arm the loop. **Rationale**: auto-arming a loop each session would resume unbounded background scheduling the user did not re-consent to; a record-only file plus a session-start advisory reminder keeps re-arm an explicit user action while ensuring the intent is not forgotten. Affects REQ-MWS-011 / REQ-MWS-018.

## §E Milestones (decision-reversibility ordered)

### M1 — Workflow file data model + frontmatter contract (highest change-likelihood)

Define the `.moai/workflows/<name>.md` frontmatter schema (`name`, `description`, `schedule` = expression + `mechanism: cron|loop`, `safety` = `read-only|write`) and the body-as-natural-language-steps contract. Covers REQ-MWS-001..006. This is the most reversible decision (schema shape drives everything downstream) so it leads.

### M2 — Safety-boundary encoding (cadence invariant)

Encode the cadence read-only invariant into the workflow skill body: scheduled runs never commit/push/enter run-phase; `safety: write` governs interactive runs only; human gates are cadence-unsatisfiable and surface to the queue. Covers REQ-MWS-014..017. High-review-value because it is the safety-critical decision.

### M3 — Creation entry surface + guided capture (user-facing UX)

Author `.claude/skills/moai/workflows/workflow.md` + the `/moai:workflow` wrapper + router entry in the `/moai` catalog. AskUserQuestion-guided capture of name/description/schedule/safety/steps with creation-time validation; natural-language capture routes to the same path. Covers REQ-MWS-007..009. Resolve the §D name-collision clarification before finalizing.

### M4 — Schedule registration / unregistration wiring

Wire CronCreate (cron mechanism) / `/loop` guidance (loop mechanism) at creation, and CronDelete / session-scoped cancellation at removal, reusing the harness-v4 vocabulary. Include the Cron-unavailable fallback to `/loop`. Covers REQ-MWS-010..013. Resolve the §D loop-re-arm clarification.

### M5 — Discovery + lifecycle (list / edit / remove)

Orchestrator-driven list (`.moai/workflows/*.md` enumeration with name/description/schedule render), edit (open markdown file), remove (unregister then delete, no-op on absent). Covers REQ-MWS-018..020.

### M6 — Template scaffold deployment (mechanical, trails)

Add `internal/template/templates/.moai/workflows/README.md` (format explainer) + one neutral example workflow file. Honor template-neutrality (no internal IDs/dates/SHAs) and Template-First (`make build` after). Covers REQ-MWS-021..023. Lowest change-likelihood (mechanical scaffolding) so it trails.

### M7 — Boundary documentation (non-duplication)

Ensure the boundary section (spec.md §C) is reflected wherever the four sibling assets are documented, so future readers do not re-fork the layer. Covers REQ-MWS-024.

## §F Technical Approach

- The feature is **orchestration + skill-body + template**, with essentially no Go code (the deferred CLI aside). The primary deliverables are: one skill workflow file, one wrapper command, one router-catalog line, and a template scaffold directory.
- Registration is via native Cron tools (orchestrator-invoked); no persistence layer beyond the workflow files themselves and whatever the native scheduler stores.
- Reuse, do not fork: the schedule vocabulary (Branch A.1) and the safety invariant (cadence-bridge) are consumed verbatim.

## §G Risks and Anti-Patterns

- **R1 — Cron tool availability drift**: CronCreate/List/Delete availability varies by runtime version. Mitigation: degrade to `/loop` guidance; document the fallback (matches cadence-bridge precedent).
- **R2 — Safety-tier confusion**: a user may expect `safety: write` to let a scheduled run commit. Mitigation: REQ-MWS-016 explicitly scopes `write` to interactive runs; the skill body and template README must state this prominently.
- **R3 — Template neutrality leak**: the example workflow could accidentally reference an internal SPEC/date. Mitigation: REQ-MWS-022 + CI guard (`template-neutrality-check.yaml`) + pre-commit self-check.
- **R4 — Scope creep toward a scheduler daemon**: the temptation to "just add a small daemon" for reliability. Anti-pattern — explicitly out of scope (spec.md §E). Native scheduler only.
- **R5 — Duplication with cadence-bridge recipes**: a user workflow could re-implement a fixed cadence recipe. Not a defect (user autonomy), but the boundary doc (§C) should point users at the fixed recipes for the common cases.

## §H Cross-References

- spec.md §C — boundary definition vs harness-v4 / `.js` workflows / `/loop` / cadence-bridge.
- acceptance.md — AC-MWS matrix + DoD.
- `.claude/rules/moai/workflow/cadence-bridge.md`, `.claude/skills/moai/SKILL.md` § Branch A.1.
