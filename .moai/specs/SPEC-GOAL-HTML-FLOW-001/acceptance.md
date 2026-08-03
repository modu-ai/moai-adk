# acceptance.md — SPEC-GOAL-HTML-FLOW-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-GHF-NNN); this file does NOT restate them as requirements. **HARD AC requirement (AUTONOMY-TIERS run-phase lesson)**: every AC verifies end-to-end wiring (render → `.html` file actually written → browser-visible DOM parsed), NOT unit-level grep alone. The named anti-pattern is claiming AC PASS on written criteria while deferring the init→bundle wiring.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-GHF-001 | REQ-GHF-001+002+003+004 | `moai goal render` writes `.html`; DOM parse shows goal + failed-cond cmd/exit/tail + turn/ceiling + 5 verdict sections verbatim | MUST | plan M1+M2 |
| AC-GHF-002 | REQ-GHF-002 | XSS auto-escape: `<script>` payload in every untrusted field is inert under DOM parse | MUST | plan M1 |
| AC-GHF-003 | REQ-GHF-005 | `moai goal clear` removes BOTH `<session>.json` AND `<session>.html` | MUST | plan M2 |
| AC-GHF-004 | REQ-GHF-005 | `PruneOrphans` moves the `.html` sibling alongside `.json` (best-effort) | MUST | plan M2 |
| AC-GHF-005 | REQ-GHF-007+008 | Plan HTML report written; DOM shows goal + 8-field contract + verdict score + milestones | MUST | plan M3 |
| AC-GHF-006 | REQ-GHF-009 | Implementation Kickoff Approval `AskUserQuestion` STILL fires in the same turn the report exists (gate not replaced) | MUST | plan M4 |
| AC-GHF-007 | REQ-GHF-010 | Re-arm UI: "re-arm on `/clear`" indicator + post-`/clear` "re-armed under `<new-id>`" view + D8 unbounded-rejection banner | MUST | plan M5 |
| AC-GHF-008 | REQ-GHF-006 | Subagent boundary: `grep` for `AskUserQuestion` / `mcp__askuser` in `internal/cli/goal.go` non-test non-comment source yields 0 matches | MUST | plan M2 |
| AC-GHF-009 | REQ-GHF-008 | 8-field derivation determinism: two runs over the same artifact set produce byte-identical output | MUST | plan M3 |
| AC-GHF-010 | REQ-GHF-004 | `moai goal render` with no armed goal → non-zero exit + stderr names session id + NO HTML file written | MUST | plan M2 |
| AC-GHF-011 | REQ-GHF-001 | `RenderDashboard(g, nil)` produces goal metadata + "no verdict yet" placeholder (no panic) | MUST | plan M1 |

### §D.1 Severity model

All eleven ACs are MUST-pass. The load-bearing trio is AC-GHF-001 + AC-GHF-002 + AC-GHF-006: AC-001 proves the end-to-end render→file→DOM wiring (the AUTONOMY-TIERS lesson — AC pass ≠ feature works unless the DOM parse is verified); AC-002 proves the XSS auto-escape is real (not a grep for `html/template`); AC-006 proves the Implementation Kickoff Approval gate is preserved (not silently replaced by the report). AC-GHF-005 + AC-GHF-009 together prove the plan HTML report is real (parsed markdown + deterministic 8-field derivation). AC-GHF-007 proves the re-arm UI consumes the already-landed mechanical state. The remaining ACs (003, 004, 008, 010, 011) cover lifecycle, boundary, and edge cases.

### §D.2 AC definitions (Given-When-Then)

#### AC-GHF-001 — `moai goal render` writes `.html`; DOM parse shows goal + failed-cond + turn/ceiling + 5 verdict sections

**Given** a session with an armed goal (`.moai/state/goal/<session>.json` exists with `Status == "armed"`) AND a `Verdict` carrying at least one `FailedCondition` (cmd / exit / tail) AND a non-nil `CeilingVerdict`.

**When** `moai goal render` is invoked (or `goal.RenderDashboard(g, v)` is called directly in a unit test).

**Then** the file `.moai/state/goal/<session>.html` is written (exists on disk), AND a DOM parse of the file (via `golang.org/x/net/html` — already a project dependency) shows: (a) the goal condition text from `g.Goal`, (b) each failed condition's `Cmd`, `Exit`, and `Tail` from `v.FailedConditions`, (c) the `Turn` and `Ceiling` values from `v`, AND (d) the 5 `CeilingVerdict` section headings verbatim: `Claim`, `Evidence`, `Baseline-attribution`, `Gaps`, `Residual-risk`.

**Test shape:** Go unit test — construct a `Goal` + `Verdict` fixture with known values + a `<script>`-free failed-condition tail; call `RenderDashboard`; parse the bytes via `golang.org/x/net/html`; walk the DOM asserting each element is present. End-to-end CLI variant: run `moai goal render` (via a test that exercises `runGoalRender`), assert the file exists on disk, read it back, DOM-parse, assert the same elements. The DOM parse is load-bearing — a unit test that asserts only "bytes contain substring X" fails this AC.

#### AC-GHF-002 — XSS auto-escape verified by DOM parse (script tag inert)

**Given** a `Goal` whose `Goal` field is `"<script>alert(1)</script>"`, a `Condition` whose `Cmd` is `"<script>alert(2)</script>"`, a `Verdict` whose `FailedConditions[0].Tail` is `"<script>alert(3)</script>"`, AND a `CeilingVerdict` whose `Claim` is `"<script>alert(4)</script>"` (and similarly for `Evidence`, `BaselineAttribution`, `Gaps`, `ResidualRisk`).

**When** `RenderDashboard(g, v)` is called.

**Then** a DOM parse of the rendered HTML classifies EVERY `<script>` payload as inert text content — the parsed DOM contains ZERO `<script>` child elements whose text content is `alert(1)` / `alert(2)` / `alert(3)` / `alert(4)`. The literal substring `<script>` MAY appear in the rendered HTML (as escaped text `&lt;script&gt;`), but the DOM tree MUST NOT contain an executable script node.

**Test shape:** Go unit test — fixture with `<script>` in every untrusted field; render; DOM-parse; assert `len(domScriptNodes) == 0` by walking the parsed tree. Negative case: a fixture WITHOUT script payloads renders identically (no false positives from the escaping machinery).

#### AC-GHF-003 — `moai goal clear` removes BOTH `<session>.json` AND `<session>.html`

**Given** a session with an armed goal AND both `.moai/state/goal/<session>.json` AND `.moai/state/goal/<session>.html` exist on disk (the `.html` written by a prior `moai goal render`).

**When** `moai goal clear` is invoked (or `goal.ClearGoal(root, session)` is called directly).

**Then** BOTH files are absent on disk after the call returns, AND the call exits 0 / returns nil. Idempotency: a second `ClearGoal` call on the same session (both files already absent) ALSO exits 0 / returns nil (the `fs.ErrNotExist`-is-idempotent contract extends to the `.html` sibling).

**Test shape:** Go unit test — `t.TempDir()` root; write both `<session>.json` and `<session>.html`; call `ClearGoal`; assert `os.Stat` returns `fs.ErrNotExist` for both. Negative case: only `.json` exists (no prior render) — `ClearGoal` still succeeds, `.html` absence is not an error.

#### AC-GHF-004 — `PruneOrphans` moves the `.html` sibling alongside `.json`

**Given** a goal state directory (`.moai/state/goal/`) carrying an orphaned session's `<session>.json` AND its `<session>.html` sibling (orphaned = session absent from the active-sessions list OR mtime older than `OrphanTTL`).

**When** `goal.PruneOrphans(root, activeSessions, now)` is called.

**Then** the `<session>.json` is moved to `.moai/state/goal/consumed/<session>.json` AND the `<session>.html` is moved to `.moai/state/goal/consumed/<session>.html`. Best-effort: a simulated failure to move the `.html` sibling (e.g., a permission error on the `.html` source) does NOT abort the `.json` move or the rest of the sweep.

**Test shape:** Go unit test — `t.TempDir()` root; write both files with an old mtime; call `PruneOrphans`; assert both files moved to `consumed/`. Best-effort variant: make the `.html` source unreadable (chmod); assert `.json` still moved + no error returned.

#### AC-GHF-005 — Plan HTML report written; DOM shows goal + 8-field contract + verdict score + milestones

**Given** a SPEC directory with `spec.md` (frontmatter + §A/§B/§D), `plan.md` (§F milestones), `acceptance.md` (§D AC matrix), AND a plan-auditor review file at `.moai/reports/plan-audit/{SPEC-ID}-review-{N}.md` carrying `Verdict: PASS`, `Overall Score: 0.85`, and at least one must-pass row + one defect row.

**When** the plan-HTML-report renderer is invoked (`RenderPlanHTML(specDir, reviewFile)` or via the M4 emission step).

**Then** the file `.moai/reports/plan-html/{SPEC-ID}-plan.html` is written, AND a DOM parse of the file shows: (a) the SPEC's goal text (from §A), (b) all 8 autonomy-contract fields (goal / scope / non-goals / tools-permissions / stopping-condition / evidence / escalation / budget) each with non-empty derived content, (c) the verdict score `0.85`, (d) the must-pass row(s), AND (e) the milestone list (from `plan.md §F`).

**Test shape:** Go unit test — fixture SPEC directory + fixture review-file markdown; render; read the output file; DOM-parse; assert each element. End-to-end variant: invoke the M4 emission step on the fixture directory; assert the file lands at the expected path + DOM contents.

#### AC-GHF-006 — Implementation Kickoff Approval AskUserQuestion STILL fires in the same turn (gate not replaced)

**Given** the plan-phase emission step (M4) has fired AND `.moai/reports/plan-html/{SPEC-ID}-plan.html` exists.

**When** the orchestrator reaches the plan→run boundary in the same turn.

**Then** the Implementation Kickoff Approval `AskUserQuestion` round STILL fires — the turn's tool-call trace includes an `AskUserQuestion` invocation carrying the three canonical options (run-phase entry / further review / abort) with the first option carrying the `(권장)` / `(Recommended)` label. The HTML report path is surfaced as additive prose context in the same turn; it is NOT substituted for the gate. The gate's score-independence is preserved: a plan-auditor PASS or a high skip-eligible score does NOT auto-bypass the AskUserQuestion.

**Test shape:** workflow/skill-layer fixture — drive the plan-phase closeout sequence (plan-auditor PASS → M4 emission step → Implementation Kickoff Approval) on a fixture SPEC; capture the orchestrator's tool-call trace for the turn; assert (a) the `AskUserQuestion` invocation is present, (b) its options match the canonical 3, (c) the first option carries `(권장)` / `(Recommended)`, AND (d) the report path appears in the turn's prose. Negative case: a workflow variant that skips the AskUserQuestion (gate substitution) fails this AC.

#### AC-GHF-007 — Re-arm UI: indicator + verification view + D8 banner

**Given** three fixture states: (a) `.moai/state/handoff/pending.json` carries a non-nil `embedded_goal` whose `IsUnbounded()` is false; (b) `pending.json` is absent/empty BUT `.moai/state/goal/<new-session>.json` exists for a post-`/clear` session; (c) `pending.json` carries an `embedded_goal` whose `IsUnbounded()` is true.

**When** `RenderDashboard` is invoked against each fixture (with the appropriate session id for state b).

**Then** for state (a): the rendered dashboard DOM contains a visible "this goal will re-arm on `/clear`" indicator carrying the embedded condition text + the embedded ceiling (`max_turns` / `max_duration` / `cost_cap`). For state (b): the rendered dashboard DOM contains a "re-armed under `<new-session-id>`" view with a pointer to the new goal file. For state (c): the rendered dashboard DOM contains a D8-rejection banner naming the unbounded embedded goal as the cause (the existing `rearmEmbeddedGoal` path at `handoff_inject.go:196-249` rejects the rearm; this UI surfaces that rejection).

**Test shape:** Go unit test — three fixture `pending.json` / `<session>.json` pairs; render each; DOM-parse; assert each indicator/view/banner is present with the expected content. The D8 banner test MUST call `EmbeddedGoal.IsUnbounded()` (already implemented at `pending.go:47-49`) — it does NOT re-implement the predicate.

#### AC-GHF-008 — Subagent-boundary grep guard

**Given** the `internal/cli/goal.go` source file (the file that gains the `render` verb in M2).

**When** the subagent-boundary guard test (mirroring `internal/cli/web_test.go` `TestWeb_NoAskUserQuestion`) scans the file.

**Then** the grep `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go | grep -v _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*#'` yields ZERO matches. The `render` verb (and every other verb in the file) reports every outcome via exit code + stderr; none prompt the user.

**Test shape:** Go unit test — read `internal/cli/goal.go` source; run the grep above; assert zero matches. The test fails on ANY `AskUserQuestion` / `mcp__askuser` reference outside test files and comments.

#### AC-GHF-009 — 8-field derivation determinism

**Given** a fixed SPEC artifact set (`spec.md`, `plan.md`, `acceptance.md`, `settings.json`) AND a fixed plan-auditor review file.

**When** `RenderPlanHTML` is invoked twice (or N times) over the same artifact set.

**Then** the 8-field contract portion of the rendered HTML is byte-identical across all runs. Determinism is at the FIELD level: each of the 8 fields (goal / scope / non-goals / tools-permissions / stopping-condition / evidence / escalation / budget) renders the same string content in the same order every time. (Whitespace differences in the surrounding HTML template are permissible; the 8-field DERIVED CONTENT is byte-identical.)

**Test shape:** Go unit test — fixture artifact set; call `RenderPlanHTML` twice; extract the 8-field block from each rendered output (e.g., via DOM parse of the contract section); assert byte-equality of each field's text content.

#### AC-GHF-010 — `moai goal render` with no armed goal exits non-zero + names session + writes NO HTML

**Given** a session with NO armed goal (`.moai/state/goal/<session>.json` absent) AND the `.html` sibling also absent.

**When** `moai goal render` is invoked for that session.

**Then** the command exits non-zero, the stderr message names the resolved session id (mirroring `status`'s no-goal behavior), AND NO `.html` file is written (the non-existent `.html` path remains non-existent).

**Test shape:** Go unit test / CLI smoke — invoke `runGoalRender` with a session id whose goal file is absent; assert non-zero exit + stderr contains the session id + `os.Stat` on the `.html` path returns `fs.ErrNotExist`.

#### AC-GHF-011 — `RenderDashboard(g, nil)` produces goal metadata + "no verdict yet" placeholder (no panic)

**Given** a `Goal` fixture AND a nil `Verdict` (the armed-but-not-yet-evaluated case).

**When** `RenderDashboard(g, nil)` is called.

**Then** the call does NOT panic, AND the rendered HTML contains: (a) the goal metadata (condition text, ceiling, turns-used, status), AND (b) a visible "no verdict yet" placeholder where the verdict section would otherwise render. The 5 `CeilingVerdict` section names are NOT rendered (the verdict is nil — there is nothing to surface).

**Test shape:** Go unit test — `RenderDashboard(g, nil)`; assert no panic; DOM-parse; assert the goal metadata elements are present AND the placeholder text is present AND the 5 section names are absent.

### §D.3 Indirect verification

- The `html/template` auto-escape behavior is verified indirectly: AC-GHF-001 asserts the happy-path render, AC-GHF-002 asserts the XSS-inertness property, and AC-GHF-011 asserts the nil-verdict edge case. The escape mechanism itself is NOT modified; it is the standard library's. Any regression in `html/template`'s escaping would surface as a failure in AC-GHF-002.
- The `CeilingVerdict` section-name preservation (REQ-GHF-003) is verified indirectly: AC-GHF-001 sub (d) greps the rendered DOM for all 5 literals; AC-GHF-011 asserts they are absent when the verdict is nil. The struct's JSON tags (`evaluate.go:46-52`) are unchanged.

### §D.4 Closure gates

- All eleven MUST ACs green with attributed evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per `verification-claim-integrity.md` §3).
- The load-bearing trio (AC-GHF-001 end-to-end DOM + AC-GHF-002 XSS inertness + AC-GHF-006 gate preservation) is green — these are the AUTONOMY-TIERS lesson + the security binding + the gate-preservation binding.
- The plan HTML report pair (AC-GHF-005 + AC-GHF-009) is green — the report is real and deterministic.
- The re-arm UI (AC-GHF-007) is green against the already-landed mechanical state.
- Subagent boundary (AC-GHF-008) green.
- LSP gate: zero errors, zero type errors, lint clean.
- Cross-platform build: `GOOS=windows GOARCH=amd64 go build ./...` exits 0.

### §D.5 Forward-looking checks (advisory, non-blocking for this SPEC)

- The per-turn Stop-hook auto-refresh of `.moai/state/goal/<session>.html` (§3.2 node ⑤ LIVE board) is deferred to v3.2. When that SPEC lands, it SHOULD reuse `RenderDashboard` (M1) as its renderer; the renderer signature is stable.
- The WebSocket live board (the `moai web` initiative, v3.1 deployment target per project memory) is a separate surface; it MAY consume `RenderDashboard`'s output as its initial paint, but the live-update transport is out of scope here.
- The plan-auditor JSON sidecar (structured plan-auditor output) is a documented FOLLOW-UP. When it lands, the M3 markdown parser is replaced by (or supplemented with) a JSON reader; the `{{audit_html}}` slot and the 8-field derivation are unchanged.
- The orchestrator "continue/clear/switch" AskUserQuestion simplification (§3.2's "오케스트레이터는 사용자에게 '대시보드 갱신했습니다 — continue/clear/switch?'만 물을 뿐" line) is deferred to v3.2. When it lands, it consumes the same dashboard renderer + the per-turn auto-refresh (also v3.2).

### §D.6 Definition of Done

This SPEC is DONE when:

1. All eleven MUST ACs pass with attributed evidence (5-section format per `verification-claim-integrity.md`).
2. The load-bearing trio (AC-GHF-001 + AC-GHF-002 + AC-GHF-006) is green — the end-to-end DOM wiring is verified, the XSS auto-escape is real, and the Implementation Kickoff Approval gate is preserved (not replaced).
3. The plan HTML report pair (AC-GHF-005 + AC-GHF-009) is green — the report renders the parsed verdict + the deterministic 8-field contract.
4. The re-arm UI (AC-GHF-007) is green against the already-landed mechanical state (no mechanical re-specification).
5. The subagent boundary (AC-GHF-008) is green.
6. The project's standard quality gate is green (lint + test + cross-platform build).
7. Frontmatter `status` transitions `draft → in-progress → implemented → completed` are owned by manager-develop and manager-docs; this plan-phase authoring only emits `draft`.
