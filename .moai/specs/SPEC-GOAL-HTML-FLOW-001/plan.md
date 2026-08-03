# plan.md — SPEC-GOAL-HTML-FLOW-001

> Implementation plan. Order is decision-reversibility-first: the dashboard renderer substrate + escaping (the most central + security-binding decision) leads, then the verb + sibling lifecycle, then the plan HTML report renderer, then the workflow emission step, then the re-arm UI (render-only), then end-to-end AC verification.

## §A. Context

This SPEC is the **P1 successor** to `SPEC-INFINITE-GOAL-001` (loop mechanics + `/clear` re-arm pipeline, completed) and `SPEC-AUTONOMY-TIERS-001` (tier-aware hooks, completed) in the autonomy-workflow epic. It codifies §3.2 of the design report (`moai-autonomy-workflow-redesign-20260803.html` lines ~356-380) — the 5-node HTML-first flow — and is scoped by a 2026-08-04 user decision to deliver ONLY: (A) on-demand static HTML dashboard, (B) plan HTML report at the plan→run boundary, (C) render-only re-arm UI. The per-turn auto-refresh, WebSocket live board, and AskUserQuestion simplification are deferred to v3.2 (§B Out of Scope in spec.md).

The re-arm mechanical pipeline is ALREADY LANDED (`SPEC-INFINITE-GOAL-001` REQ-6) — this SPEC renders its state; it does NOT re-implement it.

### A.1 Work location

- Branch: `plan/SPEC-GOAL-HTML-FLOW-001` (this worktree, off origin/main).
- Repo-local PR policy (`repo-local-pr-policy.md`): ALL tiers use Route B (PR); plan-phase artifacts land via a `plan/` PR (self-merge allowed; CI must pass). The orchestrator/manager-git handles PR creation; manager-spec authors files only.

### A.2 SPEC artifact paths

- `.moai/specs/SPEC-GOAL-HTML-FLOW-001/spec.md` (this SPEC's requirements)
- `.moai/specs/SPEC-GOAL-HTML-FLOW-001/plan.md` (this file)
- `.moai/specs/SPEC-GOAL-HTML-FLOW-001/acceptance.md` (AC matrix + GWT)
- `.moai/specs/SPEC-GOAL-HTML-FLOW-001/progress.md` (§E lifecycle skeleton — plan-phase §E.1 populated; §E.2-§E.4 placeholders)

### A.3 plan-auditor verdict

_Not yet run._ plan-auditor is invoked at plan-phase closeout; its verdict lands in `.moai/reports/plan-audit/SPEC-GOAL-HTML-FLOW-001-review-{N}.md`. The skip-eligible threshold for this SPEC is `0.80` (Tier M).

### A.4 PRESERVE list (do NOT modify)

- `internal/goal/schema.go` (`Goal`, `Ceiling`, `Condition` types) — consumed unchanged.
- `internal/goal/evaluate.go` (`Verdict`, `FailedCond`, `CeilingVerdict` types; the evaluator itself) — consumed unchanged.
- `internal/cli/handoff.go:76-92` (save-embed of `EmbeddedGoal`) — sibling-owned by `SPEC-INFINITE-GOAL-001` REQ-6.
- `internal/hook/handoff_inject.go:196-249` (`rearmEmbeddedGoal`) — sibling-owned.
- `internal/hook/handoff/pending.go:36-65` (`EmbeddedGoal`, `IsUnbounded`) — sibling-owned.
- `.claude/agents/moai/plan-auditor.md:341-384` (review-file markdown schema) — shared contract; this SPEC parses the markdown only.

### A.5 EXTEND targets

- `internal/goal/dashboard.go` (NEW — renderer substrate).
- `internal/cli/goal.go` (add `renderCmd` + `runGoalRender`).
- `internal/goal/state.go` (`ClearGoal` `.html` sibling remove).
- `internal/goal/prune.go` (`PruneOrphans` `.html` sibling move).
- `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache` (`{{audit_html}}` slot).
- `.claude/skills/moai/workflows/plan/spec-assembly.md` (emission step).
- `.claude/skills/moai/workflows/goal.md` (`### /moai goal render` subsection).
- NEW: plan-HTML-report renderer package or file (location TBD at M3 — candidate: `internal/report/planhtml/` or extension of `internal/goal/dashboard.go` into a shared html-report helper).

## §B. Known Issues

- **K-1 — `html/template` auto-escape is context-aware, NOT field-aware.** The default `{{.Field}}` action in HTML context escapes correctly; but if a field is rendered inside a `<script>` block or a JS string literal, the escaping rules differ. Mitigation: the dashboard template renders EVERY untrusted field in HTML context (between tags, never inside `<script>`). The XSS AC (AC-GHF-002) embeds a `<script>` payload and asserts inertness via DOM parse, not via grep for the literal substring.
- **K-2 — `template.HTML` escape hatch is a footgun.** Any `{{.Field}}` where the field is of Go type `template.HTML` bypasses auto-escaping. Mitigation: the renderer's struct fields are ALL `string` (or `[]FailedCond`, etc., whose inner fields are `string`); NO field is typed `template.HTML`. The M1 test asserts this at the struct-definition level.
- **K-3 — `CeilingVerdict` JSON tags differ from the rendered headings.** The struct uses `json:"Claim"`, `json:"Evidence"`, `json:"Baseline-attribution"`, `json:"Gaps"`, `json:"Residual-risk"`; the rendered headings MUST be the same 5 literals verbatim. The M1 test greps the rendered HTML for all 5.
- **K-4 — `.html` sibling path derivation.** `StatePath` returns `<root>/.moai/state/goal/<session>.json`; the renderer/clear path derives the `.html` path via `strings.TrimSuffix(path, ".json") + ".html"`. Edge case: if `StatePath` ever returns a path without the `.json` suffix (it does not today, but defensively), the derivation falls back to `path + ".html"` (producing `<session>.json.html` — ugly but safe). The M2 test covers both branches.
- **K-5 — Plan-auditor markdown parser fragility.** The review-file schema at `plan-auditor.md:341-384` is stable but markdown is whitespace-tolerant. The M3 parser is strict about the section headings (`## Must-Pass Results`, `## Category Scores`, `## Defects Found`) and falls back to an "audit verdict unavailable" placeholder (REQ-GHF-007 fail-open) on any parse miss. The parser does NOT fail the plan-phase pipeline.
- **K-6 — 8-field derivation rule for `settings.json` permissions.** `settings.json` may be absent, may have no `permissions.allow`, or may be user-scope vs project-scope. The derivation reads the UNION of project + user scope (best-effort; absence → "undetermined"). Two runs over the same artifact set produce byte-identical output (AC-GHF-009 determinism).
- **K-7 — Implementation Kickoff Approval gate is a human gate, NOT a code site.** The emission step in `spec-assembly.md` runs BEFORE the gate; the gate itself is an orchestrator-side `AskUserQuestion` round. The M4 workflow edit MUST NOT add a code-level bypass, a score-check, or an auto-proceed. The AC (AC-GHF-006) asserts the `AskUserQuestion` STILL fires in the same turn.
- **K-8 — Re-arm UI consumes already-written state.** The dashboard's "re-arm on `/clear`" indicator reads `.moai/state/handoff/pending.json` (`embedded_goal` field); the post-`/clear` view reads `.moai/state/goal/<new-session>.json`. Both are ALREADY written by the landed mechanical pipeline. The M5 UI does NOT write either file — it only reads + renders. The D8-rejection banner reads `pending.json`'s `embedded_goal` and calls `IsUnbounded()` (already implemented at `pending.go:47-49`).

## §C. Pre-flight (read-only reconnaissance — before M1)

1. Read `internal/goal/schema.go` — confirm `Goal`, `Ceiling`, `Condition` exported field names + JSON tags.
2. Read `internal/goal/evaluate.go:30-85` — confirm `FailedCond` (`:37-41`), `CeilingVerdict` (`:46-52`), `Verdict` (`:57-79`) struct shapes + the comment at `:46` naming the 5 sections load-bearing for `AC-GLE-013`.
3. Read `internal/cli/goal.go:100-140` — confirm `cmd.AddCommand(armCmd, statusCmd, clearCmd)` is the registration site + `statusCmd` pattern.
4. Read `internal/cli/goal.go:195-215` — confirm `statusSessionID` signature + the no-silent-pid-fallback discipline (arm-path property, not status/clear).
5. Read `internal/goal/state.go:20-35,95-115` — confirm `StatePath` derivation + `ClearGoal` `os.Remove` + `fs.ErrNotExist` idempotency.
6. Read `internal/goal/prune.go:20-40` — confirm `PruneOrphans` `os.ReadDir` + best-effort move contract.
7. Read `internal/hook/handoff/pending.go:30-70` — confirm `EmbeddedGoal` schema + `IsUnbounded()` + `PendingRecord.EmbeddedGoal` field.
8. Read `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache` — confirm existing slot structure + where `{{audit_html}}` slots in.
9. Read `.claude/agents/moai/plan-auditor.md:341-384` — confirm review-file markdown schema (verdict / score / must-pass / category / defects).
10. Read `.claude/skills/moai/workflows/plan/spec-assembly.md` — confirm plan-phase closeout sequence + where the emission step lands (after plan-auditor PASS, before Implementation Kickoff Approval).
11. Read `internal/cli/web_test.go` `TestWeb_NoAskUserQuestion` — confirm the subagent-boundary guard test pattern to mirror for `internal/cli/goal.go`.

## §D. Constraints (recap from spec.md §D — binding on the plan)

1. Stdlib `html/template` only; no JS/CSS framework; no `template.HTML` injection.
2. Existing data models unchanged — consume, don't modify.
3. 5 CeilingVerdict section names preserved verbatim.
4. Implementation Kickoff Approval stays mandatory + score-independent.
5. Re-arm mechanics sibling-owned (do NOT re-specify/extend).
6. Plan-auditor shared contract unchanged (parse markdown only).
7. Subagent boundary: `render` verb never prompts; `TestWeb_NoAskUserQuestion` pattern mandatory.
8. `.html` sibling cleanup in both `ClearGoal` and `PruneOrphans`.
9. 8-field derivation rules (spec.md REQ-GHF-008 table) are deterministic — byte-identical across runs.
10. `tags` comma-quoted-string; self-lint before return.

## §E. Self-Verification (run-phase — what manager-develop must demonstrate)

- Go unit test: `RenderDashboard(g, v)` returns non-empty bytes; a DOM parse (via `golang.org/x/net/html` — already a dep) shows goal text + each failed-condition's cmd/exit/tail + turn + ceiling (AC-GHF-001).
- Go unit test: a `<script>` payload in every untrusted field renders inert under DOM parse (AC-GHF-002).
- Go unit test: rendered HTML contains the 5 literal section names Claim / Evidence / Baseline-attribution / Gaps / Residual-risk when `v.Verdict != nil` (AC-GHF-001 sub, K-3).
- Go unit test: `ClearGoal` removes BOTH `.json` and `.html` siblings; idempotent on both (AC-GHF-003).
- Go unit test: `PruneOrphans` moves the `.html` sibling alongside `.json` (orphan-TTL path); best-effort on `.html` does not abort `.json` move (AC-GHF-004).
- CLI smoke: `moai goal render` writes `<session>.html`; stdout names the path; `--json` emits JSON (AC-GHF-001).
- CLI smoke: `moai goal render` with no armed goal → non-zero exit + stderr names session id + NO HTML file written (AC-GHF-010).
- Go unit test: `RenderDashboard(g, nil)` does not panic; produces goal metadata + "no verdict yet" placeholder (AC-GHF-011).
- Go unit test: plan-HTML-report renderer parses a fixture review-file markdown; DOM parse of the rendered HTML shows verdict + score + must-pass + 8-field contract + milestones (AC-GHF-005).
- Go unit test: two runs of the 8-field derivation over the same fixture artifact set produce byte-identical output (AC-GHF-009).
- Grep test: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go | grep -v _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*#'` yields 0 matches (AC-GHF-008, subagent boundary).
- Hook/workflow fixture: the plan-HTML-emission step fires AFTER plan-auditor PASS and BEFORE Implementation Kickoff Approval; in the same turn, the Implementation Kickoff Approval `AskUserQuestion` STILL fires (AC-GHF-006 — verified at the workflow/skill layer, not pure unit test).
- Go unit test: the dashboard's "re-arm on `/clear`" indicator renders when `pending.json` carries a non-nil `embedded_goal`; the post-`/clear` view renders "re-armed under `<new-id>`" when the new goal file exists; the D8 banner renders when `IsUnbounded()` returns true (AC-GHF-007).
- Lint / build / cross-platform build / test clean.

## §F. Milestones

### Milestone M1 — Dashboard renderer substrate + escaping (highest reversibility)

The renderer signature + escaping approach is the most central decision. The data models already exist; this milestone adds `internal/goal/dashboard.go` only.

**Files (expected):**
- `internal/goal/dashboard.go` — NEW. `RenderDashboard(g *Goal, v *Verdict) ([]byte, error)` via `html/template`. Inline CSS (no framework). Struct fields typed `string` (NOT `template.HTML`).
- `internal/goal/dashboard_test.go` — NEW. AC-GHF-001 (DOM parse via `golang.org/x/net/html`), AC-GHF-002 (`<script>` inertness), AC-GHF-011 (`v == nil` no-verdict path), AC-GHF-001 sub (5 section names verbatim).

**Exit:** `RenderDashboard` produces a self-contained HTML; escaping verified by DOM parse; section names preserved; nil-verdict path does not panic.

### Milestone M2 — `moai goal render` verb + sibling `.html` lifecycle

The CLI verb + the clear/prune sibling handling. Slots the renderer behind a user-invocable command.

**Files (expected):**
- `internal/cli/goal.go` — add `renderCmd` + `runGoalRender`; register in `cmd.AddCommand(...)`. Reuse `statusSessionID` + `goal.LoadGoal`. Derive `.html` path from `StatePath` via extension swap. Write bytes; report path on stdout (or JSON with `--json`).
- `internal/goal/state.go` — extend `ClearGoal` to remove the `.html` sibling with the same idempotent contract.
- `internal/goal/prune.go` — extend `PruneOrphans` to move the `.html` sibling (best-effort).
- Tests: AC-GHF-003 (`ClearGoal` both files), AC-GHF-004 (`PruneOrphans` sibling), AC-GHF-008 (subagent-boundary grep guard), AC-GHF-010 (no-armed-goal non-zero exit).
- `.claude/skills/moai/workflows/goal.md` — add `### /moai goal render` subsection.

**Exit:** `moai goal render` writes the dashboard; `clear` removes both siblings; pruner moves both; boundary guard green.

### Milestone M3 — Plan HTML report renderer (md-parse + 8-field + mustache slot)

The report structure is more reversible than the workflow edit (M4), so M3 lands first.

**Files (expected):**
- NEW package or file (candidate: `internal/report/planhtml/renderer.go`) — `RenderPlanHTML(specDir string, reviewFile string) ([]byte, error)`. Parses the plan-auditor markdown review (strict on section headings; fail-open placeholder on miss). Derives the 8-field contract per spec.md REQ-GHF-008 table. Renders into the `{{audit_html}}` slot of `plan.html.mustache`.
- `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache` — add `{{audit_html}}` slot.
- Tests: AC-GHF-005 (DOM parse shows goal + 8-field + score + milestones), AC-GHF-009 (determinism — two runs byte-identical), AC-GHF-007 sub (fail-open when review file absent).

**Exit:** renderer produces a single self-contained plan HTML report; 8-field derivation deterministic; fail-open on missing review.

### Milestone M4 — Plan-phase emission step (workflow edit)

The workflow edit (`spec-assembly.md`) is less reversible than the renderer (M3), so it lands after.

**Files (expected):**
- `.claude/skills/moai/workflows/plan/spec-assembly.md` — add the emission step AFTER plan-auditor PASS, BEFORE Implementation Kickoff Approval. The step invokes `RenderPlanHTML`, writes `.moai/reports/plan-html/{SPEC-ID}-plan.html`, and surfaces the path to the orchestrator. The Implementation Kickoff Approval `AskUserQuestion` gate is UNCHANGED — the path is additive prose context, NOT a gate substitution.
- Tests / fixture: AC-GHF-006 (the AskUserQuestion STILL fires in the same turn the report exists).

**Exit:** plan-phase closeout emits the HTML report; Implementation Kickoff Approval gate preserved verbatim.

### Milestone M5 — Resume re-arm UI (render-only)

Consumes M1's renderer + the already-landed re-arm mechanical pipeline. Lowest central-decision risk because the mechanics are sibling-owned.

**Files (expected):**
- `internal/goal/dashboard.go` — extend `RenderDashboard` (or add a sibling renderer) to consume `pending.json`'s `embedded_goal` field + the post-`/clear` `<new-session>.json` + `EmbeddedGoal.IsUnbounded()`. Render the "re-arm on `/clear`" indicator, the "re-armed under `<new-id>`" view, and the D8-rejection banner.
- Tests: AC-GHF-007 (all three UI states).

**Exit:** re-arm indicator + verification view + D8 banner render correctly against the landed mechanical state.

### Milestone M6 — End-to-end AC verification + sync

Lowest reversibility. Full `go test ./...` + race (`internal/goal/...`, `internal/cli/...`, `internal/report/...` if added). All AC matrix green with attributed evidence. Cross-platform build (`GOOS=windows GOARCH=amd64 go build ./...`). Lint clean. Sync-phase close (CHANGELOG, docs) per the 3-phase lifecycle.

## §G. Anti-Patterns (specific to this SPEC)

- **AP-1 — Using `template.HTML` to inject any field.** `template.HTML` bypasses auto-escaping. Every rendered field is `string`; the `html/template` default action in HTML context is the only escape path. (K-2.)
- **AP-2 — Rendering an untrusted field inside a `<script>` block.** `html/template` escapes differently in JS context; the dashboard renders untrusted fields in HTML context ONLY (between tags). (K-1.)
- **AP-3 — Paraphrasing the 5 CeilingVerdict section names.** "Claim" → "Summary", "Baseline-attribution" → "Baseline", etc. breaks `AC-GLE-013`'s grep. The 5 literals are load-bearing. (K-3.)
- **AP-4 — Adding a code-level Implementation Kickoff Approval bypass.** A score check, an auto-proceed, or a "skip if report exists" branch in the workflow violates REQ-GHF-009 + `CLAUDE.local.md` §19.1. The gate is a human gate; the report enriches the review surface only. (K-7.)
- **AP-5 — Re-implementing the re-arm mechanics.** The save-embed, `rearmEmbeddedGoal`, `EmbeddedGoal` schema, and `IsUnbounded` are ALREADY landed (`SPEC-INFINITE-GOAL-001` REQ-6). This SPEC renders their state; any mechanical change goes back to the sibling SPEC as an amendment. (K-8.)
- **AP-6 — Modifying the plan-auditor shared contract.** Adding a JSON sidecar to plan-auditor to make parsing easier is OUT OF SCOPE (deferred to a follow-up SPEC). This SPEC parses the markdown review file only.
- **AP-7 — Letting `PruneOrphans` abort on `.html` sibling failure.** The existing best-effort contract (`.json` move failure does not abort the sweep) extends to `.html`: a failure to move the `.html` sibling is logged + swallowed.
- **AP-8 — Free-form LLM synthesis for the 8-field contract.** The 8 fields are DERIVED from SPEC artifacts per REQ-GHF-008's table. A renderer that calls an LLM to "summarize the goal" violates the determinism AC (AC-GHF-009).
- **AP-9 — Treating `moai goal render` as subagent-invocable.** The verb is orchestrator-invoked CLI; it never prompts. Adding an `AskUserQuestion` call (even for "no goal armed") violates the subagent boundary (REQ-GHF-006).

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/spec.md`.
- acceptance.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/acceptance.md`.
- progress.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/progress.md` (§E lifecycle skeleton).
- Design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.2 (lines ~356-380; inlined verbatim into spec.md §A). Recommendation: commit this report as a citation (currently gitignored under `.moai/reports/`).
- Sibling SPECs: `SPEC-INFINITE-GOAL-001` (loop mechanics + re-arm pipeline; owns `internal/cli/handoff.go`, `internal/hook/handoff_inject.go`, `internal/hook/handoff/pending.go`), `SPEC-AUTONOMY-TIERS-001` (tier-aware hooks).
- Implementation Kickoff Approval gate: `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E + `.claude/skills/moai/workflows/run.md:135-141` + `CLAUDE.local.md` §19.1.
- Subagent-boundary guard: `internal/cli/CLAUDE.md` C-HRA-008; `internal/cli/web_test.go` `TestWeb_NoAskUserQuestion`.
- Plan-auditor review schema: `.claude/agents/moai/plan-auditor.md:341-384`.
- Plan HTML template host: `.claude/skills/moai-domain-html-report/references/templates/plan.html.mustache`.
- Integrity invariants: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (dashboard DOM-visible state is a claim surface).
