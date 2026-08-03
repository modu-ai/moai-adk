---
id: SPEC-GOAL-HTML-FLOW-001
title: "HTML-first /moai goal flow — on-demand dashboard, plan-phase HTML report, resume re-arm UI"
version: "0.1.0"
status: draft
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P1
phase: "v3.1 target"
module: "internal/goal,internal/cli,workflows/plan,workflows/goal"
lifecycle: spec-anchored
tags: "goal-engine, html-dashboard, plan-report, autonomy-epic, rearm-ui, xss-escape"
tier: M
related_specs: [SPEC-INFINITE-GOAL-001, SPEC-AUTONOMY-TIERS-001]
---

# SPEC-GOAL-HTML-FLOW-001 — HTML-first /moai goal flow

## HISTORY

- 2026-08-04 — Initial draft. Codifies §3.2 of the autonomy-workflow redesign report (`.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.2, "5-node HTML-first flow", lines ~356-380) — the P1 successor to `SPEC-INFINITE-GOAL-001` (loop-continuation mechanics, completed) and `SPEC-AUTONOMY-TIERS-001` (tier-aware hooks, completed). User decision 2026-08-04 (this SPEC's scope meeting): dashboard per-turn / WebSocket live board / orchestrator AskUserQuestion simplification are DEFERRED to v3.2; this SPEC covers ONLY (A) on-demand static HTML dashboard via `moai goal render`, (B) plan HTML report at the plan→run boundary, and (C) render-only resume re-arm UI consuming the already-landed mechanical re-arm pipeline (`SPEC-INFINITE-GOAL-001` REQ-6).

## §A. User Story

**As a** MoAI maintainer running a long-horizon Tier-M/L SPEC under an armed `/moai goal`,
**I want** the goal evaluator's per-turn `Verdict` data model rendered as a self-contained HTML dashboard I can open in a browser, the plan→run boundary review surfaced as a rich HTML report instead of inline prose, and the `/clear` re-arm path exposed as a visible re-arm indicator —
**so that** "chat-monitoring" becomes "browser-monitoring", the plan-approval review is enriched without weakening the Implementation Kickoff Approval human gate, and the re-arm mechanic (already landed) is observable instead of silent.

**5-node HTML-first flow contract (§3.2, inlined verbatim from the design authority for clone-reproducibility — `moai-autonomy-workflow-redesign-20260803.html` §3.2 lines ~356-380):**

```
① /moai plan   (manager-spec author → plan-auditor)
   ▶
② HTML report  (spec + plan + acceptance + audit verdict + 8-field contract → single HTML artifact)
   ▶
③ async review (user reviews HTML; approve or return edit; NO inline blocking)
   ▶
④ autonomous execution on confirm (/moai goal arm, tier-selected; milestone folding under 1M limit; read-only fan-out · workflow · worktree-parallel write)
   ▶
⑤ per-turn-end HTML dashboard (evaluate.go Verdict rendered; AC matrix · turn/ceiling; failed-condition cmd/exit/tail)
```

This SPEC delivers nodes ② (plan HTML report — Deliverable B), the ON-DEMAND slice of ⑤ (`moai goal render` — Deliverable A, NOT the per-turn Stop-hook auto-refresh which is v3.2), and the render-only surface of the `/clear` re-arm pipeline (Deliverable C). Nodes ① and ④ are owned by pre-existing workflows; ③ is the existing Implementation Kickoff Approval gate, ENRICHED not replaced.

**Outcome hypotheses:**

- `evaluate.go`'s `Verdict` (`evaluate.go:57-79`), `FailedCond` (`:37-41`), and `CeilingVerdict` (`:46-52`) data models already exist; the per-turn Stop hook already produces them. Rendering them as self-contained HTML (via stdlib `html/template`, no JS/CSS framework dependency) is mechanical and does not require touching the evaluator. The 5 `CeilingVerdict` section names (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) are load-bearing — `AC-GLE-013` greps for them, and the dashboard MUST preserve them verbatim (comment at `internal/goal/evaluate.go:46`).
- The plan-auditor already writes its verdict to `.moai/reports/plan-audit/{SPEC-ID}-review-{N}.md` in a deterministic markdown schema (`plan-auditor.md:341-384`). Parsing that markdown into a struct and rendering it into the plan HTML report is mechanical; a JSON sidecar is deferred to a follow-up SPEC so the plan-auditor's shared contract is left untouched.
- The 8-field autonomy contract (goal / scope / non-goals / tools-permissions / stopping-condition / evidence / escalation / budget — the Osmani set cited in §3.2) can be derived deterministically from SPEC artifacts (frontmatter + §A/§B/§D + `acceptance.md` + `settings.json` `permissions.allow`), removing free-form synthesis from the path.
- The mechanical `/clear` re-arm pipeline already landed (`SPEC-INFINITE-GOAL-001` REQ-6: `internal/cli/handoff.go:76-92` save-embed + `internal/hook/handoff_inject.go:196-249` `rearmEmbeddedGoal` + `internal/hook/handoff/pending.go:36-65` `EmbeddedGoal` schema with `IsUnbounded` at `:47-49`). This SPEC owes ONLY the render/UI surface on top of that pipeline; it does NOT re-specify re-arm mechanics.

## §B. Scope

**In scope — three deliverables (on-demand dashboard, plan HTML report, re-arm UI):**

- **REQ-GHF-001** — Dashboard renderer substrate at `internal/goal/dashboard.go`: `RenderDashboard(g *Goal, v *Verdict) ([]byte, error)` produces a self-contained HTML document via stdlib `html/template` (no JS/CSS framework dependency; zero external CDN). The renderer is a pure function — it accepts the data models that already exist (`Goal` at `schema.go`, `Verdict` at `evaluate.go:57-79`) and returns bytes; no I/O.
- **REQ-GHF-002** — XSS auto-escape: untrusted fields (`Goal.Goal`, `Condition.Cmd`, `FailedCond.Cmd`/`FailedCond.Tail`, `CeilingVerdict.Claim`/`Evidence`/`BaselineAttribution`/`Gaps`/`ResidualRisk`) SHALL be rendered through `html/template`'s context-aware auto-escaping (the `{{.Field}}` default action in `html/template` HTML context, NOT `template.HTML`). No field reaches the DOM as raw markup unless it is a trusted static literal.
- **REQ-GHF-003** — CeilingVerdict section preservation: the rendered dashboard SHALL surface the 5-section verdict carrying the literal section headings Claim / Evidence / Baseline-attribution / Gaps / Residual-risk verbatim. These are load-bearing — `AC-GLE-013` greps for them and the comment at `evaluate.go:46` marks them so.
- **REQ-GHF-004** — `moai goal render` verb at `internal/cli/goal.go`: a new `renderCmd` registered alongside `armCmd`/`statusCmd`/`clearCmd` in the `cmd.AddCommand(...)` line, following the `statusCmd` pattern (~`:113-119`). It SHALL reuse `statusSessionID` (~`:202-210`) and `goal.LoadGoal` to resolve the live goal, call `RenderDashboard`, and write the self-contained HTML to `.moai/state/goal/<session>.html` (derive from `StatePath` at `state.go:25-31` + extension swap `.json` → `.html`). The command is idempotent: re-running overwrites the file.
- **REQ-GHF-005** — Sibling `.html` lifecycle: `ClearGoal` (`state.go:102-111`) and `PruneOrphans` (`prune.go:27`) SHALL remove / prune the `.html` sibling alongside the `.json` state file. `ClearGoal` SHALL use `os.Remove` on the derived `.html` path with the same `fs.ErrNotExist`-is-idempotent contract as the `.json` delete.
- **REQ-GHF-006** — Subagent-boundary guard: the `render` verb SHALL NOT invoke `AskUserQuestion` (every outcome is decided from observable state and reported via exit code + stderr). A static guard test mirroring `TestWeb_NoAskUserQuestion` (`internal/cli/CLAUDE.md` C-HRA-008, `internal/cli/web_test.go` pattern) SHALL scan the `internal/cli/goal.go` source for `AskUserQuestion` / `mcp__askuser` references and fail on any match outside test files and comments.
- **REQ-GHF-007** — Plan HTML report renderer: the plan-phase pipeline SHALL emit a single self-contained plan HTML report. The renderer parses the plan-auditor verdict markdown at `.moai/reports/plan-audit/{SPEC-ID}-review-{N}.md` (the plan-phase review stream; schema at `plan-auditor.md:341-384`) into a struct, derives the 8-field autonomy contract from SPEC artifacts, and renders both into the `{{audit_html}}` slot of `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache`. The JSON sidecar (structured plan-auditor output) is DEFERRED to a follow-up SPEC — the parser reads the markdown review file only.
- **REQ-GHF-008** — 8-field autonomy contract deterministic derivation: the renderer SHALL derive the 8 Osmani fields from SPEC artifacts deterministically (no free-form synthesis): (1) goal ← SPEC `§A` user-story goal clause; (2) scope ← SPEC `§B` in-scope bullets; (3) non-goals ← SPEC `§B` out-of-scope `### Out of Scope —` H3 sub-headings; (4) tools-permissions ← `settings.json` `permissions.allow` list; (5) stopping-condition ← `acceptance.md` §D AC matrix; (6) evidence ← `acceptance.md` Given-When-Then evidence rows + plan.md §E self-verification; (7) escalation ← plan.md open-questions / blocker-report cross-refs; (8) budget ← SPEC frontmatter `tier` + `phase` + plan.md milestone count. The derivation rules are enumerated in `plan.md §D` of this SPEC.
- **REQ-GHF-009** — Implementation Kickoff Approval ENRICHED, NOT replaced: the plan HTML report emission step added to `.claude/skills/moai/workflows/plan/spec-assembly.md` SHALL land AFTER plan-auditor PASS and BEFORE Implementation Kickoff Approval. The Implementation Kickoff Approval `AskUserQuestion` gate (`run.md:135-141`, `orchestration-mode-selection.md` §E, `CLAUDE.local.md` §19.1) stays MANDATORY and score-independent. The report replaces the REVIEW SURFACE (inline prose → rich HTML), NOT the gate. The orchestrator presents the report path in the same turn the AskUserQuestion round fires; the gate's three options (run-phase entry / further review / abort) and the `(권장)` first-option label are unchanged.
- **REQ-GHF-010** — Resume re-arm UI (render-only; mechanical layer ALREADY landed by `SPEC-INFINITE-GOAL-001` REQ-6): the dashboard renderer SHALL consume the existing `embedded_goal` field in `.moai/state/handoff/pending.json` and render a "this goal will re-arm on `/clear`" indicator in the dashboard when the field is non-nil. A re-arm verification view (rendered when the dashboard is invoked with the post-`/clear` new-session id) SHALL read `.moai/state/goal/<new-session>.json` and surface "re-armed under `<new-session-id>`" when the new goal file exists. The D8 unbounded-rejection (`EmbeddedGoal.IsUnbounded`, `pending.go:47-49`) SHALL be surfaced in the dashboard as a visible banner when the pending record carries an unbounded embedded goal (the existing `rearmEmbeddedGoal` path rejects it at rearm; this REQ surfaces that rejection in the UI).

**Out of scope — deferred to v3.2 or follow-up SPECs:**

### Out of Scope — Dashboard per-turn auto-refresh

- The per-turn Stop-hook auto-refresh of `.moai/state/goal/<session>.html` (§3.2 node ⑤ "per-turn-end HTML dashboard" as a LIVE board) — DEFERRED to v3.2. This SPEC delivers ONLY the on-demand `moai goal render` slice (user/maintainer invokes the verb; file is written).
- The WebSocket / push-based live dashboard board — DEFERRED to v3.2.

### Out of Scope — Orchestrator AskUserQuestion simplification

- The §3.2 simplification that collapses the orchestrator's "continue/clear/switch" decision into a single dashboard-driven AskUserQuestion (the report's "오케스트레이터는 사용자에게 '대시보드 갱신했습니다 — continue/clear/switch?'만 물을 뿐" line) — DEFERRED to v3.2. This SPEC does NOT modify the orchestrator's confirm-round surface; the report ENRICHES the Implementation Kickoff Approval gate only.

### Out of Scope — Plan-auditor JSON sidecar

- A JSON sidecar from plan-auditor (structured `{verdict, score, must_pass, defects}` payload) that the renderer would consume instead of parsing the markdown review file — DEFERRED to a follow-up SPEC. This SPEC keeps the plan-auditor shared contract unchanged and parses the markdown review file only.

### Out of Scope — Mechanical re-arm pipeline

- The mechanical `/clear` re-arm pipeline (save-side embed, inject-side `rearmEmbeddedGoal`, schema, D8 `IsUnbounded` defense-in-depth) — ALREADY LANDED by `SPEC-INFINITE-GOAL-001` REQ-6 and its tests. This SPEC renders the existing pipeline's state; it does NOT re-specify, modify, or extend the mechanics. Any change to the mechanics goes back to `SPEC-INFINITE-GOAL-001` as an amendment.

### Out of Scope — Web live dashboard (v3.1 target)

- The `moai web` WebSocket live board (separate initiative, v3.1 deployment target per project memory `project_web_live_dashboard_v31.md`) — unrelated to this SPEC's on-demand static dashboard.

## §C. Requirements (GEARS)

### REQ-GHF-001 — Dashboard renderer substrate

**Where** the `internal/goal` package owns the goal data model, a new file `internal/goal/dashboard.go` SHALL expose `RenderDashboard(g *Goal, v *Verdict) ([]byte, error)`. The function SHALL produce a self-contained HTML document (inline CSS; zero external JS/CSS framework dependency; zero CDN) via the standard library's `html/template` package, consuming the existing `Goal` (`schema.go`) and `Verdict` (`evaluate.go:57-79`) data models. The function SHALL be a pure function — it accepts the data models and returns the rendered bytes; it performs NO file I/O and NO subprocess invocation.

**When** `v` is nil (no verdict yet produced — e.g., an armed-but-not-yet-evaluated goal), the renderer SHALL produce a dashboard carrying the goal metadata (condition text, ceiling, turns-used, status) without the verdict section, with a placeholder reading "no verdict yet".

### REQ-GHF-002 — XSS auto-escape (security)

**Where** the dashboard renderer handles untrusted fields (`Goal.Goal`, each `Condition.Cmd`, each `FailedCond.Cmd` and `FailedCond.Tail`, each `CeilingVerdict` string field `Claim`/`Evidence`/`BaselineAttribution`/`Gaps`/`ResidualRisk`), the renderer SHALL emit each field through `html/template`'s default auto-escaping (the `{{.Field}}` action in HTML context), NOT through `template.HTML` or any raw-markup injection.

**When** the dashboard's tests embed a `<script>alert(1)</script>` payload in each untrusted field, a DOM parse of the rendered HTML SHALL classify the payload as inert text content (the `<script>` substring MUST NOT appear as a child script element in the parsed DOM).

### REQ-GHF-003 — CeilingVerdict section names preserved verbatim

**While** a `CeilingVerdict` is present on the rendered `Verdict`, the dashboard SHALL surface the 5-section evidence-bearing report carrying the literal section headings verbatim: `Claim`, `Evidence`, `Baseline-attribution`, `Gaps`, `Residual-risk`. The headings are load-bearing — `AC-GLE-013` greps for them (per the comment at `evaluate.go:46`) — and MUST NOT be paraphrased, re-cased, or localized.

### REQ-GHF-004 — `moai goal render` verb

**Where** the `internal/cli/goal.go` `cmd.AddCommand(...)` line today registers `armCmd`, `statusCmd`, and `clearCmd`, a new `renderCmd` SHALL be registered alongside them. The verb SHALL follow the `statusCmd` pattern (~`:113-119`): `Use: "render"`, `Args: cobra.NoArgs`, `SilenceUsage: true`, and a `RunE` delegating to a `runGoalRender` helper.

**Where** `runGoalRender` resolves the active session, it SHALL reuse `statusSessionID` (~`:202-210`) — the same session-id resolver the read/idempotent verbs (`status`, `clear`) use — and `goal.LoadGoal` to read the live goal. It SHALL call `goal.RenderDashboard`, write the bytes to the path derived from `StatePath` (`state.go:25-31`) with the extension swapped from `.json` to `.html`, and report the written path on stdout (or as JSON when `--json` is passed).

**When** no goal is armed for the resolved session, the verb SHALL exit non-zero with a stderr message naming the session id (mirroring `status`'s no-goal behavior) and SHALL NOT write an HTML file.

### REQ-GHF-005 — Sibling `.html` lifecycle (clear + prune)

**Where** `ClearGoal` (`state.go:102-111`) today removes the `.json` state file via `os.Remove` with an `fs.ErrNotExist`-is-idempotent contract, it SHALL ADDITIONALLY remove the `.html` sibling at the derived `.html` path with the SAME idempotent contract (`fs.ErrNotExist` → return nil). The sibling path is derived once and both removes are attempted; failure on either (other than not-exist) is returned.

**Where** `PruneOrphans` (`prune.go:27`) moves orphaned `.json` state files to `.moai/state/goal/consumed/`, it SHALL ADDITIONALLY move the `.html` sibling when it exists (best-effort: a failure to move the `.html` sibling does NOT abort the `.json` move or the rest of the sweep, mirroring the existing best-effort contract).

### REQ-GHF-006 — Subagent-boundary guard test

**Where** the `render` verb is CLI-surface code (orchestrator-invocable, never subagent-prompted), the verb SHALL NOT invoke `AskUserQuestion` or any `mcp__askuser*` tool. Every outcome (success, no-goal, write-failure) is decided from observable state and reported via exit code and stderr.

**Where** the test suite enforces the `TestWeb_NoAskUserQuestion` pattern (`internal/cli/CLAUDE.md` C-HRA-008, `internal/cli/web_test.go`), a sibling test file (e.g. `internal/cli/goal_render_boundary_test.go` OR an extension of an existing `goal_test.go`) SHALL scan the non-test, non-comment source of `internal/cli/goal.go` for `AskUserQuestion` / `mcp__askuser` references and FAIL on any match.

### REQ-GHF-007 — Plan HTML report renderer (Deliverable B core)

**Where** the plan-phase pipeline emits the plan-auditor verdict to `.moai/reports/plan-audit/{SPEC-ID}-review-{N}.md` (plan-phase review stream; schema at `plan-auditor.md:341-384`), a new plan-HTML-report renderer SHALL parse that markdown verdict into a struct (verdict, overall score, must-pass rows, category scores, defects) and render it into the `{{audit_html}}` slot of `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache`.

**Where** the renderer derives the 8-field autonomy contract (REQ-GHF-008), it SHALL render the 8 fields into the plan HTML report alongside the verdict and the SPEC's milestone list (from `plan.md §F`). The output is a single self-contained HTML artifact at `.moai/reports/plan-html/{SPEC-ID}-plan.html` (new directory; gitignored).

**When** the plan-auditor review file is absent or unparseable, the renderer SHALL emit the report with an "audit verdict unavailable" placeholder in the audit slot rather than fail the plan-phase pipeline (fail-open; the plan HTML report is enrichment, not a gate).

### REQ-GHF-008 — 8-field autonomy contract deterministic derivation

**Where** the plan HTML report carries the 8-field autonomy contract, the renderer SHALL derive each field deterministically from SPEC artifacts (no free-form LLM synthesis). The derivation rules:

| # | Field | Source |
|---|-------|--------|
| 1 | goal | SPEC `§A` user-story "so that" clause |
| 2 | scope | SPEC `§B` in-scope bullet list (REQ IDs + one-line summaries) |
| 3 | non-goals | SPEC `§B` `### Out of Scope — <topic>` H3 sub-headings + their `-` bullets |
| 4 | tools-permissions | `settings.json` `permissions.allow` list (project + user scope union) |
| 5 | stopping-condition | `acceptance.md` §D AC matrix (AC IDs + MUST/SHOULD severity) |
| 6 | evidence | `acceptance.md` Given-When-Then "Then" clauses + `plan.md §E` self-verification items |
| 7 | escalation | `plan.md` open-questions + blocker-report cross-references |
| 8 | budget | SPEC frontmatter `tier` + `phase` + `plan.md §F` milestone count |

**When** a source artifact is absent or a field cannot be derived (e.g., `settings.json` has no `permissions.allow`), the renderer SHALL render the field as "undetermined — <reason>" rather than omit it. Deterministic derivation means two runs over the same artifact set produce byte-identical 8-field output.

### REQ-GHF-009 — Implementation Kickoff Approval gate ENRICHED, not replaced

**Where** `.claude/skills/moai/workflows/plan/spec-assembly.md` orchestrates the plan-phase closeout, an emission step SHALL land AFTER plan-auditor PASS and BEFORE Implementation Kickoff Approval. The step invokes the plan-HTML-report renderer (REQ-GHF-007) and surfaces the resulting HTML path to the orchestrator.

**While** the plan HTML report exists in a given turn, the Implementation Kickoff Approval `AskUserQuestion` gate (`run.md:135-141`, `orchestration-mode-selection.md` §E, `CLAUDE.local.md` §19.1) SHALL remain MANDATORY and score-independent — a plan-auditor PASS or a high skip-eligible score does NOT auto-bypass it, and the existence of the report does NOT relax it. The report replaces the REVIEW SURFACE (the inline prose summary the orchestrator previously rendered) with a rich HTML artifact; it does NOT replace the gate. The orchestrator's three gate options (run-phase entry / further review / abort) and the `(권장)` first-option label are unchanged.

### REQ-GHF-010 — Resume re-arm UI (render-only; mechanical layer ALREADY landed)

**Where** `SPEC-INFINITE-GOAL-001` REQ-6 landed the mechanical `/clear` re-arm pipeline (`internal/cli/handoff.go:76-92` save-embed, `internal/hook/handoff_inject.go:196-249` `rearmEmbeddedGoal`, `internal/hook/handoff/pending.go:36-65` `EmbeddedGoal` schema), this SPEC owes ONLY the render/UI surface on top of that pipeline.

**Where** `.moai/state/handoff/pending.json` carries a non-nil `embedded_goal` field, the dashboard renderer (`RenderDashboard`) SHALL render a visible "this goal will re-arm on `/clear`" indicator carrying the embedded condition text and the embedded ceiling (`max_turns`, `max_duration`, `cost_cap`).

**Where** the dashboard is rendered against a post-`/clear` new session id (the new `.moai/state/goal/<new-session>.json` exists), a re-arm verification view SHALL surface "re-armed under `<new-session-id>`" with a link/pointer to the new goal file.

**Where** the pending record carries an unbounded embedded goal (`EmbeddedGoal.IsUnbounded()`, `pending.go:47-49`), the dashboard SHALL render a visible D8-rejection banner — the existing `rearmEmbeddedGoal` path rejects the rearm at injection time; this REQ surfaces that rejection in the UI so the user can see WHY the goal did NOT re-arm.

## §D. Constraints

1. **Stdlib-only HTML rendering**: `html/template` only. No external JS/CSS framework, no CDN, no `template.HTML` injection. The HTML artifact is openable offline (email-attachment-safe, print-safe).
2. **Existing data models unchanged**: this SPEC does NOT modify `Goal` (`schema.go`), `Verdict` / `FailedCond` / `CeilingVerdict` (`evaluate.go`). It consumes them. Any data-model change goes back to the owning SPEC.
3. **CeilingVerdict section names load-bearing**: the 5 literal headings (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) are grepped by `AC-GLE-013` and MUST be preserved verbatim in the rendered output.
4. **Implementation Kickoff Approval stays mandatory**: the plan HTML report ENRICHES the gate; it does NOT replace it. The gate's score-independence (`CLAUDE.local.md` §19.1) is unchanged.
5. **Re-arm mechanics are sibling-owned**: the `/clear` re-arm pipeline is owned by `SPEC-INFINITE-GOAL-001` REQ-6 and is already landed. This SPEC renders its state; it does NOT re-implement, extend, or modify the mechanics.
6. **Plan-auditor shared contract unchanged**: this SPEC parses the markdown review file only. A JSON sidecar is deferred to a follow-up SPEC (§B Out of Scope).
7. **Subagent boundary**: the `render` verb is orchestrator-invoked CLI; it never prompts. The `TestWeb_NoAskUserQuestion` pattern is mandatory (C-HRA-008).
8. **Sibling `.html` cleanup**: both `ClearGoal` and `PruneOrphans` MUST handle the `.html` sibling so the goal state directory does not accumulate orphan HTML after `/clear` or session-TTL expiry.
9. **Reversibility-first milestone order**: M1 (renderer substrate + escaping) lands before M2 (verb + lifecycle) because the renderer signature is the most central decision; M3 (plan report) lands before M4 (emission step) because the report structure is more reversible than the workflow edit; M5 (re-arm UI) lands last because it consumes M1's renderer.
10. **MP-3 lesson**: `tags` is a comma-separated quoted string (not a YAML array). Self-verified via `moai spec lint` before return.

## §E. Assumptions

1. `internal/goal/schema.go` `Goal` and `internal/goal/evaluate.go` `Verdict` (`:57-79`), `FailedCond` (`:37-41`), `CeilingVerdict` (`:46-52`) are stable exported types — verified by reading the source; the renderer imports and consumes them without modification.
2. `internal/cli/goal.go` `cmd.AddCommand(armCmd, statusCmd, clearCmd)` is the single registration site; the `renderCmd` slots in alongside. `statusSessionID` (`:202-210`) and `goal.LoadGoal` are reusable as-is for the read path.
3. `internal/goal/state.go` `StatePath` (`:25-31`) returns `<root>/.moai/state/goal/<session>.json`; swapping the `.json` extension to `.html` is a `strings.TrimSuffix(..., ".json") + ".html"` derivation — no new helper required.
4. `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache` already exists (verified, 15514 bytes) and accepts Mustache slots; adding `{{audit_html}}` is a slot extension, not a template rewrite. The existing `moai-domain-html-report` skill's rendering pipeline is the host.
5. `.claude/agents/moai/plan-auditor.md:341-384` documents the review-file markdown schema deterministically (verdict, overall score, must-pass rows, category scores, defects). The schema is stable; parsing it is mechanical.
6. `.moai/reports/plan-audit/{SPEC-ID}-review-{N}.md` is the plan-phase review stream's iteration file (per `spec-workflow.md` § Report Persistence). The renderer reads the final-iteration file (highest `N` present).
7. The mechanical re-arm pipeline is verified landed: `internal/cli/handoff.go:76-92` (save-embed), `internal/hook/handoff_inject.go:196-249` (`rearmEmbeddedGoal`), `internal/hook/handoff/pending.go:36-65` (`EmbeddedGoal`), `internal/hook/handoff_rearm_test.go` (tests). This SPEC's re-arm UI consumes `pending.json` and the post-`/clear` `<new-session>.json` — both already written by the landed pipeline.

## §F. Open Questions

- **OQ-1 (REQ-GHF-007) — plan-auditor markdown schema stability.** RESOLVED: the schema at `plan-auditor.md:341-384` is treated as stable for THIS SPEC. The renderer parses the markdown review file's verdict / overall-score / must-pass / category-score / defect sections. If a future SPEC amends the schema, the parser is updated in lockstep (no JSON sidecar needed for v3.1).
- **OQ-2 (REQ-GHF-009) — where the plan HTML report path is surfaced.** RESOLVED: the orchestrator surfaces the HTML path as a prose line in the SAME turn the Implementation Kickoff Approval `AskUserQuestion` fires (not as a preview field — the report path is a pointer, not the report content). The `AskUserQuestion` option text is unchanged; the path is additive context above the gate.
- **OQ-3 (REQ-GHF-010) — re-arm verification view trigger.** RESOLVED: the "re-armed under `<new-id>`" view fires when `RenderDashboard` is invoked against a session id whose `.moai/state/goal/<id>.json` exists AND whose `pending.json` archive (`.moai/state/handoff/consumed/`) carries a recently-consumed record naming the OLD session id with a non-nil `embedded_goal`. The "recently-consumed" window is 24 hours (best-effort; the consumed archive is the source of truth, not a heuristic).

## §G. References

- Design authority: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.2 ("5-node HTML-first flow", lines ~356-380). **Recommendation**: this report is currently gitignored (`.moai/reports/`); commit it as a citation so the inlined §3.2 contract is traceable. The §3.2 contract is inlined verbatim into §A above for clone-reproducibility.
- Goal data model: `internal/goal/schema.go` (`Goal`, `Ceiling`), `internal/goal/evaluate.go:37-41` (`FailedCond`), `:46-52` (`CeilingVerdict`), `:57-79` (`Verdict`).
- Goal CLI: `internal/cli/goal.go` (`cmd.AddCommand` line, `statusCmd` ~`:113-119`, `statusSessionID` ~`:202-210`, `runGoalStatus`).
- Goal state lifecycle: `internal/goal/state.go:25-31` (`StatePath`), `:102-111` (`ClearGoal`), `internal/goal/prune.go:27` (`PruneOrphans`).
- Re-arm mechanical pipeline (sibling-owned, ALREADY landed): `internal/cli/handoff.go:76-92` (save-embed), `internal/hook/handoff_inject.go:196-249` (`rearmEmbeddedGoal`), `internal/hook/handoff/pending.go:36-65` (`EmbeddedGoal` + `IsUnbounded` `:47-49`), `internal/hook/handoff_rearm_test.go` (tests).
- Plan-auditor review schema: `.claude/agents/moai/plan-auditor.md:341-384` (Output Format). Plan-audit review stream: `.moai/reports/plan-audit/{SPEC-ID}-review-{N}.md` (plan-phase, iteration-based).
- Plan HTML template: `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache` (existing; `{{audit_html}}` slot to be added).
- Plan-phase workflow: `.claude/skills/moai/workflows/plan/spec-assembly.md` (emission-step insertion site).
- Implementation Kickoff Approval gate: `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E, `.claude/skills/moai/workflows/run.md:135-141`, `CLAUDE.local.md` §19.1.
- Sibling SPECs: `SPEC-INFINITE-GOAL-001` (loop mechanics + re-arm pipeline, completed), `SPEC-AUTONOMY-TIERS-001` (tier-aware hooks, completed). `SPEC-INFINITE-GOAL-001/spec.md:53` defers the re-arm UI to "sibling P1" (= this SPEC); `SPEC-AUTONOMY-TIERS-001/spec.md:48,:57` names this SPEC as the §3.2 sibling.
- Subagent-boundary guard: `internal/cli/CLAUDE.md` C-HRA-008; `internal/cli/web_test.go` `TestWeb_NoAskUserQuestion` pattern.
- Integrity invariants: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (the dashboard's DOM-visible state is a claim surface — REQ-GHF-002's auto-escape is the security binding; REQ-GHF-003's section-name preservation is the traceability binding).

## §H. Acceptance Criteria (summary — full GWT in acceptance.md)

- AC-GHF-001 (REQ-1+2+3, E2E): `moai goal render` writes `.html`; DOM parse shows goal text + failed-condition cmd/exit/tail + turn/ceiling + 5 verdict sections verbatim.
- AC-GHF-002 (REQ-2, XSS): a `<script>` payload in every untrusted field is inert under DOM parse.
- AC-GHF-003 (REQ-5): `moai goal clear` removes BOTH `<session>.json` AND `<session>.html`.
- AC-GHF-004 (REQ-5): `PruneOrphans` moves the `.html` sibling alongside the `.json` (orphan-TTL path).
- AC-GHF-005 (REQ-7+8, E2E): plan HTML report written; DOM shows goal + 8-field contract + verdict score + milestones.
- AC-GHF-006 (REQ-9, gate-preservation): Implementation Kickoff Approval `AskUserQuestion` STILL fires in the same turn the report exists; the report does NOT replace the gate.
- AC-GHF-007 (REQ-10, re-arm UI): dashboard shows "re-arm on `/clear`" indicator when `embedded_goal` non-nil; post-`/clear` view shows "re-armed under `<new-id>`"; D8 unbounded-rejection banner renders when `IsUnbounded()`.
- AC-GHF-008 (REQ-6, subagent boundary): `grep` for `AskUserQuestion` / `mcp__askuser` in `internal/cli/goal.go` non-test non-comment source yields 0 matches (guard test green).
- AC-GHF-009 (REQ-8, determinism): two runs of the 8-field derivation over the same artifact set produce byte-identical output.
- AC-GHF-010 (REQ-4, no-goal path): `moai goal render` with no armed goal exits non-zero + stderr names the session id + writes NO HTML file.
- AC-GHF-011 (REQ-1, no-verdict path): `RenderDashboard(g, nil)` produces a dashboard with goal metadata + a "no verdict yet" placeholder (no panic).
