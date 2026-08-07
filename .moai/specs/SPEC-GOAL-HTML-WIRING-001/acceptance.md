# acceptance.md — SPEC-GOAL-HTML-WIRING-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-WIRE-NNN); this file does NOT restate them as requirements. **HARD AC requirement (AUTONOMY-TIERS lesson from predecessor)**: every AC verifies END-TO-END production wiring (production-code path → file-on-disk → DOM-verified OR grep-verified non-test construction site), NOT renderer-unit scope alone. The named anti-pattern is the predecessor's failure mode — its ACs verified renderer scope and left the zero-production-caller gap invisible.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-WIRE-001 | REQ-WIRE-001+002+003 | Surface 1 end-to-end: `runGoalRender` passes non-nil Verdict when session carries one; DOM shows 5 CeilingVerdict sections (not placeholder) | MUST | plan M1 |
| AC-WIRE-002 | REQ-WIRE-003 | Surface 1 regression: verdict absent → "no verdict yet" placeholder preserved byte-identical with AC-GHF-011 | MUST | plan M1 |
| AC-WIRE-003 | REQ-WIRE-004 | `moai plan render-html` is a real Cobra subcommand registered on root (`moai plan --help`, `moai plan render-html --help` exit 0) | MUST | plan M3 |
| AC-WIRE-004 | REQ-WIRE-005+006 | Surface 2 end-to-end: CLI invoked on fixture SPEC dir → `.moai/reports/plan-html/<SPEC-ID>-plan.html` written; DOM shows goal + 8-field contract + verdict score (when review present) | MUST | plan M3 |
| AC-WIRE-005 | REQ-WIRE-006+007 | Surface 2 fail-open: review file absent → report still written with placeholder + exit 0 (006); SPEC dir absent → non-zero exit + stderr + no file (007) | MUST | plan M3 |
| AC-WIRE-006 | REQ-WIRE-008 | Surface 2 template rewrite: spec-assembly.md Step 2.3.3a no longer references the dead `RenderPlanHTML(...)` Go-function instruction; replaced with `moai plan render-html <SPEC-ID>` executable CLI path | MUST | plan M4 |
| AC-WIRE-007 | REQ-WIRE-009 | Surface 3 end-to-end: `runGoalRender` constructs a `goal.ReArmContext` from `pending.json` `EmbeddedGoal` + new-session state (grep-verified non-test construction site) AND passes it to `RenderDashboardReArm` in production | MUST | plan M2 |
| AC-WIRE-008 | REQ-WIRE-009 | Surface 3 DOM: 3 AC-GHF-007 UI states render when corresponding state present — (a) re-arm indicator, (b) "re-armed under <id>" view, (c) D8 unbounded-rejection banner | MUST | plan M2 |
| AC-WIRE-009 | REQ-WIRE-010 | Surface 3 regression: no re-arm state → base view byte-identical to `RenderDashboard(g, v, nil)` (re-arm path purely additive) | MUST | plan M2 |
| AC-WIRE-010 | REQ-WIRE-011 | C-HRA-008: grep for `AskUserQuestion`/`mcp__askuser` in `internal/cli/plan*.go` AND `internal/cli/goal.go` non-test non-comment source yields 0; enforced by a Go test mirroring `TestWeb_NoAskUserQuestion` | MUST | plan M4 |
| AC-WIRE-011 | REQ-WIRE-012+013 | PRESERVE + Template-First: renderer signatures unchanged; `make build` regenerates embedded FS; §25 neutrality (no NEW internal SPEC IDs/REQ tokens/commit SHAs added to the distributed template) | MUST | plan M4+M5 |
| AC-WIRE-012 | REQ-WIRE-001..010 | Predecessor regression: GOAL-HTML-FLOW-001 11/11 ACs AND SPEC-INFINITE-GOAL-001 re-arm tests remain green (no regression introduced by wiring) | MUST | plan M5 |
| AC-WIRE-013 | REQ-WIRE-002 | Write-frequency: `SaveVerdict` called exactly 0 times during N non-exiting turns and exactly 1 time on the ceiling-exit turn (at-ceiling-only, NOT per-turn) | MUST | plan M1 |

### §D.1 Severity model

All thirteen ACs are MUST-pass. The load-bearing trio is AC-WIRE-001 + AC-WIRE-004 + AC-WIRE-007: each proves END-TO-END wiring for one of the three inert surfaces (verdict auto-fill, plan-HTML CLI, re-arm UI construction) via file-on-disk + DOM-parse + grep-verified-non-test-construction-site. These three ACs are the direct antidote to the predecessor's "AC pass ≠ feature works" failure — they fail unless a production caller actually invokes the renderer with non-nil state.

The secondary trio (AC-WIRE-006 + AC-WIRE-010 + AC-WIRE-012) guards the cross-cutting invariants: the template rewrite replaces the dead LLM-instruction with an executable path; the subagent boundary holds across new AND modified CLI code; the predecessor + INFINITE-GOAL mechanisms regress nothing.

The remaining ACs (002, 003, 005, 008, 009, 011, 013) cover regression, CLI shape, fail-open semantics, the 3 re-arm UI DOM states, Template-First neutrality, and at-ceiling-only verdict-persistence write-frequency. AC-WIRE-013 is the write-frequency backstop: without it, a per-turn `SaveVerdict` implementation could pass AC-WIRE-001 (read-side) while silently violating the c1 scope (persistence must be at-ceiling-only, NOT the per-turn Stop-hook write the user declined under c2).

### §D.2 AC definitions (Given-When-Then)

#### AC-WIRE-001 — Surface 1 end-to-end: non-nil Verdict → 5 CeilingVerdict sections render

**Given** a session with an armed goal (`.moai/state/goal/<session>.json` exists) AND a produced Verdict persisted at `.moai/state/goal/<session>.verdict.json` carrying a `*CeilingVerdict` with non-empty `Claim`, `Evidence`, `BaselineAttribution`, `Gaps`, `ResidualRisk`.

**When** `moai goal render` is invoked (or `runGoalRender` is exercised directly in a Go test).

**Then** the file `.moai/state/goal/<session>.html` is written, AND a DOM parse of the file (via `golang.org/x/net/html`) shows the 5 `CeilingVerdict` section headings verbatim — `Claim`, `Evidence`, `Baseline-attribution`, `Gaps`, `Residual-risk` — with their content from the persisted verdict. The "no verdict yet" placeholder does NOT appear. A concurrent source-level check confirms `runGoalRender` passed a non-nil `*Verdict` to `goal.RenderDashboard` (or `RenderDashboardReArm`): grep the production path for the `LoadVerdict` call that feeds the renderer.

**Test shape**: Go test in `t.TempDir()` — construct goal-state JSON + verdict sidecar via `goal.SaveVerdict`; invoke `runGoalRender`; read back the `.html`; DOM-parse; assert 5 sections. The DOM parse is load-bearing — a test asserting only "`<session>.html` exists" fails this AC.

#### AC-WIRE-002 — Surface 1 regression: verdict absent → placeholder preserved

**Given** a session with an armed goal AND NO `.verdict.json` sidecar (or an unparseable one).

**When** `moai goal render` is invoked.

**Then** the rendered `.html` carries the "no verdict yet" placeholder byte-identical to the AC-GHF-011 baseline (predecessor). The 5 CeilingVerdict sections do NOT appear. The command exits 0 (graceful degradation, not error).

**Test shape**: Go test — `t.TempDir()` goal-state with no sidecar; invoke `runGoalRender`; DOM-parse; assert placeholder present, 5 sections absent. Variant: write a corrupt sidecar; same outcome.

#### AC-WIRE-003 — `moai plan render-html` is a registered Cobra subcommand

**Given** the `moai` binary built with this SPEC's changes.

**When** `moai plan --help` AND `moai plan render-html --help` are invoked.

**Then** both exit 0 AND their stdout names `render-html` as a subcommand (for the parent) AND names the `<SPEC-ID>` positional argument (for the child). A grep of `internal/cli/root.go` confirms `newPlanCmd()` is registered on `rootCmd`.

**Test shape**: Go test using `newPlanCmd().Execute()` (or `rootCmd.Find([]string{"plan", "render-html"})`) — assert the command tree resolves. Help-text variant via `--help` flag capture.

#### AC-WIRE-004 — Surface 2 end-to-end: CLI → file-on-disk + DOM-verified

**Given** a fixture SPEC directory at `<tmp>/.moai/specs/SPEC-FIXTURE-001/` containing `spec.md` (frontmatter + §A/§B/§D), `plan.md` (§F milestones), `acceptance.md` (§D AC matrix), AND a plan-audit review file at `<tmp>/.moai/reports/plan-audit/SPEC-FIXTURE-001-review-1.md` carrying `Verdict: PASS`, `Overall Score: 0.85`, at least one must-pass row + one defect row.

**When** `moai plan render-html SPEC-FIXTURE-001` is invoked (via `newPlanCmd()` execution in a Go test, with project root redirected to `<tmp>`).

**Then** the file `<tmp>/.moai/reports/plan-html/SPEC-FIXTURE-001-plan.html` is written (exists on disk), AND a DOM parse shows: (a) the SPEC goal text (from §A), (b) all 8 autonomy-contract fields with non-empty derived content, (c) the verdict score `0.85`, (d) the must-pass row(s), (e) the milestone list from `plan.md §F`. The command exits 0.

**Test shape**: Go test — fixture SPEC dir in `t.TempDir()`; invoke the CLI; read back the `.html`; DOM-parse; assert each element. End-to-end CLI execution (NOT a direct `RenderPlanHTML` call — that's the predecessor's unit-tested scope; this AC verifies the CLI WRAPPER actually calls it and writes the file).

#### AC-WIRE-005 — Surface 2 fail-open + missing-SPEC-dir

**Given** (a) a fixture SPEC dir with NO review file in `.moai/reports/plan-audit/`; (b) a non-existent SPEC-ID `SPEC-NO-SUCH-999`.

**When** (a) `moai plan render-html SPEC-FIXTURE-001` (no review); (b) `moai plan render-html SPEC-NO-SUCH-999`.

**Then** (a) the `.html` is still written with the "audit verdict unavailable" placeholder (REQ-GHF-007 fail-open preserved) AND exit 0; (b) the command exits NON-ZERO, emits a diagnostic to stderr naming the missing SPEC-ID, AND no `.html` file is written.

**Test shape**: Go test — two sub-tests. (a) assert file exists + DOM shows placeholder + exit 0. (b) assert exit code ≠ 0, stderr contains SPEC-ID, `os.Stat` on the expected `.html` path returns `fs.ErrNotExist`.

#### AC-WIRE-006 — spec-assembly.md template rewrite (dead LLM-instruction removed)

**Given** the edited template source `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` after M4.

**When** the file is grepped.

**Then** the literal string `RenderPlanHTML(specDir=` (the dead Go-function call instruction at the old line 204) does NOT appear in Step 2.3.3a, AND the literal `moai plan render-html` (the executable CLI path) DOES appear in Step 2.3.3a. The `[HARD]` Implementation Kickoff Approval paragraph and the fail-open clause are preserved verbatim.

**Test shape**: Go test (or CI grep) reading the template source file — assert old instruction absent, new instruction present, HARD paragraph present. Run AFTER `make build` to also assert the embedded FS carries the rewrite (parse the embedded template, assert same).

#### AC-WIRE-007 — Surface 3 end-to-end: ReArmContext constructed in production + passed to RenderDashboardReArm

**Given** the modified `internal/cli/goal.go` after M2.

**When** the production source is grepped for `ReArmContext{` construction sites outside `_test.go` files.

**Then** at least ONE non-test construction site exists in `internal/cli/goal.go` (the `runGoalRender` path), AND the constructed `*ReArmContext` is passed to `goal.RenderDashboardReArm(g, v, reArm)` (NOT to `RenderDashboard` — the re-arm-aware variant is the production path now). The construction is gated on `pending.json` `EmbeddedGoal` presence AND/OR a post-`/clear` new-session goal file per REQ-WIRE-009.

**Test shape**: Go test — (1) grep-based assertion: count `ReArmContext{` constructions in `internal/cli/goal.go` excluding `_test.go` and `//` lines, assert ≥ 1. (2) Behavioral: fixture with `pending.json` carrying an `EmbeddedGoal`; invoke `runGoalRender`; DOM-parse the output; assert re-arm indicator present. Both sub-checks are load-bearing.

#### AC-WIRE-008 — Surface 3 DOM: 3 AC-GHF-007 UI states render

**Given** three fixture configurations: (a) `pending.json` with a bounded `EmbeddedGoal` (non-empty condition, `IsUnbounded() == false`); (b) `pending.json` with `EmbeddedGoal` AND a post-`/clear` new-session goal file exists; (c) `pending.json` with an `EmbeddedGoal` where `IsUnbounded() == true`.

**When** `runGoalRender` is invoked for each.

**Then** a DOM parse of each rendered `.html` shows: (a) the "re-arm on /clear" indicator mentioning the embedded condition; (b) the "re-armed under <new-session-id>" view mentioning the new session id; (c) the D8-rejection banner mentioning the unbounded embedded condition. Each state's presence is conditioned on the corresponding fixture (cross-fixture: indicator absent in (c), banner absent in (a)/(b) — D8 takes precedence per `applyReArm`).

**Test shape**: Go test — three sub-tests, each a fixture + render + DOM-parse + state-specific assertion. The D8-precedence rule is asserted by checking (a) does NOT show the banner AND (c) does NOT show the indicator.

#### AC-WIRE-009 — Surface 3 regression: no re-arm state → base view byte-identical

**Given** a session with NO `pending.json` (or `pending.json` with `EmbeddedGoal == nil`) AND NO post-`/clear` new-session goal file.

**When** `runGoalRender` is invoked.

**Then** the rendered `.html` is byte-identical to the output of `RenderDashboard(g, v)` for the same `(g, v)` — i.e., the re-arm path is purely additive (AC-GHF-007 contract, AC-GHF-009 determinism). The 3 re-arm UI elements (indicator / re-armed-view / banner) are all absent.

**Test shape**: Go test — fixture without re-arm state; render via `runGoalRender`; render via direct `RenderDashboard(g, v)`; assert byte-equality (no field stripping). The 8-field derived content is deterministic per AC-GHF-009, and the `(g, v)` inputs are identical between the two renders, so any byte difference indicates a real regression rather than a non-deterministic field.

#### AC-WIRE-010 — C-HRA-008 subagent boundary (new + modified CLI code)

**Given** the modified `internal/cli/goal.go` and the new `internal/cli/plan.go` after M3+M4.

**When** a Go test (`TestPlan_NoAskUserQuestion` + `TestGoal_NoAskUserQuestion`) reads its own source file and scans for `AskUserQuestion` / `mcp__askuser`.

**Then** zero matches in non-test non-comment source. The test mirrors `internal/cli/web_test.go::TestWeb_NoAskUserQuestion` verbatim in shape.

**Test shape**: Go test — read `internal/cli/plan.go` and `internal/cli/goal.go` via `os.ReadFile`; assert `!strings.Contains(src, "AskUserQuestion")` AND `!strings.Contains(src, "mcp__askuser")`. (Comments and test files are out of scope — the assertion is on production source.)

#### AC-WIRE-011 — PRESERVE + Template-First + §25 neutrality

**Given** the diff produced by this SPEC.

**When** the renderer signatures are inspected AND the embedded template FS is rebuilt AND the template source is neutrality-scanned.

**Then** (a) `RenderDashboard`, `RenderDashboardReArm`, `RenderPlanHTML` signatures are byte-identical to the GOAL-HTML-FLOW-001 baseline (diff the function declarations); (b) `make build` regenerates the embedded FS and the rewritten `spec-assembly.md` is retrievable from the embedded FS; (c) a neutrality scan of the rewritten `spec-assembly.md` shows NO NEW internal SPEC IDs (`SPEC-GOAL-HTML-WIRING-001`, `SPEC-GOAL-HTML-FLOW-001`), NO REQ tokens (`REQ-WIRE-*`, `REQ-GHF-*`), NO AC tokens (`AC-WIRE-*`, `AC-GHF-*`), NO commit SHAs beyond what was already in the file pre-rewrite (the existing `RenderPlanHTML` renderer-cross-reference is permitted and preserved; the template file carries no other GHF tokens — verified via `grep 'GHF' internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md`, which returns 0 hits).

**Test shape**: Go test — (a) `go/ast` or regex signature diff; (b) `make build` + embedded-FS readback; (c) regex scan for `SPEC-[A-Z]+-[0-9]{3}` and `REQ-[A-Z]+-[0-9]{3}` patterns in the rewritten file, compared to a baseline scan of the pre-rewrite file (the DIFF must be empty — no new tokens added).

#### AC-WIRE-012 — Predecessor + INFINITE-GOAL regression

**Given** the full diff of this SPEC.

**When** the GOAL-HTML-FLOW-001 AC suite (`go test -run TestGHF ./internal/goal/... ./internal/report/planhtml/...` — or the predecessor's named test functions) AND the SPEC-INFINITE-GOAL-001 re-arm tests (`go test -run TestReArm ./internal/hook/...`) are executed.

**Then** both suites remain green (all tests pass, no regressions). The SPEC-INFINITE-GOAL-001 `rearmEmbeddedGoal` / `IsUnbounded` / `pending.json` shape are untouched (diff confirms read-only consumption).

**Test shape**: direct `go test` invocations with the named test filters. A non-green result fails this AC.

#### AC-WIRE-013 — Verdict persistence write-frequency: at-ceiling-only, NOT per-turn

**Given** a mock evaluator sequence of N (≥3) non-exiting turns — each returning a `(Verdict, bool)` whose `Verdict.Verdict *CeilingVerdict` field is `nil` — followed by 1 ceiling-exit turn whose returned `Verdict` carries a non-nil `*CeilingVerdict`.

**When** the stop-goal evaluator path runs across the full N+1-turn sequence with `SaveVerdict` invocation counting instrumented on the goal package (a test spy / mock counter wrapping the sidecar writer).

**Then** `goal.SaveVerdict` is called exactly **0 times** during the N non-exiting turns AND exactly **1 time** on the ceiling-exit turn. Persistence is at-ceiling-only, one-time per session — NOT the per-turn Stop-hook write the user declined under Surface 1 c2 (spec.md §A.3 / §G out-of-scope). This AC is the write-frequency backstop: without it, a per-turn `SaveVerdict` implementation could pass AC-WIRE-001 (read-side: `LoadVerdict` returns the persisted verdict) while silently violating the c1 scope the user approved.

**Test shape**: Go test — drive a mock evaluator through the N+1-turn sequence; count `SaveVerdict` invocations via a wrapper / function-variable hook (the evaluator's `*CeilingVerdict`-nil / non-nil branching at `internal/goal/evaluate.go:307-362` is the load-bearing discriminant); assert `count(non-exiting) == 0` AND `count(ceiling-exit) == 1`. The test MUST fail if the implementation writes on every turn.

### §D.3 Closure gates

**Definition of Done** = all 13 MUST ACs PASS + E1-E7 §E deliverables populated in `progress.md` + `gofmt`/`go vet`/`golangci-lint` clean + cross-platform build green + `make build` embedded-FS readback verified.

### §D.4 Forward-looking checks (deferred to follow-up SPECs)

- **Surface 1 c2 (per-turn LIVE board auto-refresh)**: explicitly out-of-scope per spec.md §G. A follow-up SPEC will wire the Stop-hook `.html` auto-refresh with advisory-check-discipline guards.
- **`moai plan` CLI namespace expansion** (e.g., `moai plan lint`, `moai plan audit`): this SPEC introduces the namespace with a single `render-html` subcommand. Additional subcommands are follow-up SPECs.
- **Verdict sidecar pruning beyond the `.json`+`.html` pair**: AC-GHF-003/004 precedent extended to `.verdict.json` in M1; a future audit may consolidate pruning logic.

### §D.5 Traceability summary

Each REQ maps to ≥1 AC; each AC verifies ≥1 REQ. The three inert surfaces each have ≥1 END-TO-END AC (not renderer-unit only):

| Surface | End-to-end AC |
|---------|---------------|
| 1 (verdict c1) | AC-WIRE-001 (file-on-disk + DOM + non-nil verdict grep) |
| 2 (plan-HTML) | AC-WIRE-004 (CLI invoked → file-on-disk + DOM 8-field) |
| 3 (re-arm UI) | AC-WIRE-007 (grep-verified non-test ReArmContext construction + DOM) + AC-WIRE-008 (3 UI states DOM-verified) |
