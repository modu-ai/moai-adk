# plan.md — SPEC-GOAL-HTML-WIRING-001

> Implementation plan. Milestones ordered by decision-reversibility (most-likely-to-change first). No time estimates (CLAUDE.local.md §agent-common-protocol Time Estimation).

## §A. Context

- SPEC artifacts: `.moai/specs/SPEC-GOAL-HTML-WIRING-001/{spec,plan,acceptance}.md`
- Tier M — 3 counted artifacts + progress.md §E skeleton
- Working tree: main checkout, feature branch `plan/SPEC-GOAL-HTML-WIRING-001` (per repo-local PR-mandatory policy — `enforce_admins: true` blocks main-direct push)
- Predecessor: SPEC-GOAL-HTML-FLOW-001 (status: completed, 11/11 AC green, NOT modified by this SPEC)
- Re-arm mechanism source: SPEC-INFINITE-GOAL-001 (status: completed, re-arm pipeline shipped — NOT modified by this SPEC)

### §A.1 Inert surfaces (file:line verified)

| Surface | Site | Current |
|---------|------|---------|
| 1 (verdict c1) | `internal/cli/goal.go:444` | `RenderDashboard(g, nil)` — nil hardcoded |
| 2 (plan-HTML) | `internal/report/planhtml/renderer.go:54` | 0 callers; spec-assembly.md:204 dead LLM-instruction |
| 3 (re-arm UI) | `internal/goal/dashboard.go:69` ReArmContext | 0 production constructions |

### §A.2 Verdict persistence (current state)

The `Verdict` struct lives at `internal/goal/evaluate.go:57` and carries `Verdict *CeilingVerdict` (5-section: Claim/Evidence/Baseline-attribution/Gaps/Residual-risk). It is produced by the stop-goal evaluator at turn-end and emitted as stdout JSON to the hook framework. **There is NO on-disk persistence today** — `internal/goal/state.go` carries `LoadGoal`/`SaveGoal` for the goal-state JSON only; no `LoadVerdict`/`SaveVerdict` exists. REQ-WIRE-002 closes this gap.

### §A.3 Re-arm data flow (current state)

`internal/hook/handoff_inject.go:208` `rearmEmbeddedGoal(projectDir, newSessionID, rec *handoff.PendingRecord)` writes a new goal state file under `newSessionID` when `rec.EmbeddedGoal != nil` and `!eg.IsUnbounded()`. The `PendingRecord` shape (with `EmbeddedGoal`) lives in `internal/hook/handoff.go`. The data is THERE; only the `runGoalRender` path does not read it to construct a `ReArmContext`.

## §B. Known Issues Auto-Injection (selected)

- **B3 / B11 — C-HRA-008 subagent boundary**: `internal/cli/` is subagent-domain code. No `AskUserQuestion` / `mcp__askuser`. CI guard test mirrors `internal/cli/web_test.go::TestWeb_NoAskUserQuestion`.
- **B10 — Scope discipline**: PRESERVE list (renderer signatures, predecessor artifacts, INFINITE-GOAL mechanism) is load-bearing. A parallel session on this shared checkout is possible — use explicit pathspecs.
- **Template-First (CLAUDE.local.md §2 + §25)**: `spec-assembly.md` edit → `make build` → §25 neutrality check. The CI guard `template-neutrality-check.yaml` is the safety net.
- **Advisory-Check Discipline (CLAUDE.local.md)**: REQ-WIRE-002 verdict persistence fires at the exit transition ONLY — the evaluator's `Verdict.Verdict *CeilingVerdict` field is non-nil only at a ceiling / wall-clock / stagnation exit (verified against `internal/goal/evaluate.go:307-362`), so the sidecar write is at most ONCE per session, NOT per-turn. The write is a single `os.WriteFile` at that exit transition, bounded cost, no scan, on the c1 path the user approved. It is NOT the per-turn Stop-hook `.html` write the user declined (c2, spec.md §A.3 / §G out-of-scope).

## §C. Pre-flight (run before M1)

```bash
git branch --show-current            # expect: plan/SPEC-GOAL-HTML-WIRING-001
git rev-parse HEAD                   # record baseline
go build ./...                        # baseline build
GOOS=windows GOARCH=amd64 go build ./...   # cross-platform baseline
golangci-lint run --timeout=2m 2>&1 | tail -5   # lint baseline (distinguish NEW vs pre-existing)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go internal/cli/plan*.go | grep -v _test | grep -v '// '   # expect 0
ls internal/report/planhtml/ internal/goal/    # confirm renderer sources exist
```

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE renderer signatures: `RenderDashboard(g *Goal, v *Verdict)`, `RenderDashboardReArm(g *Goal, v *Verdict, reArm *ReArmContext)`, `RenderPlanHTML(specDir, reviewFile string) ([]byte, error)`.
- PRESERVE predecessor artifacts: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/**` untouched.
- PRESERVE INFINITE-GOAL mechanism: `rearmEmbeddedGoal`, `IsUnbounded`, `pending.json` shape — consume, do NOT modify.
- PRESERVE AC-GHF-011: `RenderDashboard(g, nil)` placeholder path stays byte-identical when verdict is absent.
- Forbidden: `--no-verify`, `--amend`, force-push to main, `git add -A` on shared checkout.
- Forbidden: `AskUserQuestion` / `mcp__askuser` in `internal/cli/` subagent-domain code.

## §E. Self-Verification (manager-develop §E deliverables)

- E1 AC matrix — all 13 ACs MUST-PASS with verbatim command + observed output (see acceptance.md §D).
- E2 cross-platform build — `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- E3 coverage ≥ 85% per package on touched packages (`internal/cli`, `internal/goal`).
- E4 subagent-boundary grep — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go internal/cli/plan*.go | grep -v _test | grep -v '// '` yields 0.
- E5 lint — no NEW findings beyond baseline.
- E6 push state — list commit SHAs, push result.
- E7 blocker report on any need for a user decision (NEVER call AskUserQuestion).

## §F. Milestones

> Ordered by decision-reversibility (highest-change-likelihood first). The data-model decision (verdict persistence shape) leads; the mechanical refactor (spec-assembly.md rewrite) closes.

### M1 — Verdict persistence + Surface 1 wiring (data-model decision first)

**Why first**: REQ-WIRE-002 introduces a new on-disk artifact (`.verdict.json` sidecar). This is the decision most likely to change shape during review (storage location, JSON schema, who writes it). Landing it first surfaces the design risk before downstream wiring commits.

**Deliverables**:
1. `internal/goal/state.go` (or a new `verdict.go`): add `SaveVerdict(projectRoot, sessionID string, v *Verdict) error` and `LoadVerdict(projectRoot, sessionID string) (*Verdict, error)` operating on `.moai/state/goal/<session-id>.verdict.json`. Path helper `VerdictPath(projectRoot, sessionID) string` mirrors `HTMLPath`.
2. `internal/hook/` stop-goal evaluator path: after the evaluator produces a Verdict **whose `*CeilingVerdict` field is non-nil** (i.e. at a ceiling / wall-clock / stagnation exit transition — NOT on every turn), call `goal.SaveVerdict(...)`. The evaluator's per-turn `(Verdict, bool)` return shape (verified at `internal/goal/evaluate.go:307-362`) carries `Verdict.Verdict == nil` on non-exiting turns — those turns write NOTHING. Persistence is at-ceiling-only, one-time per session, on the c1 path the user approved; it is NOT the per-turn Stop-hook write the user declined (c2, §A.3 out-of-scope). Fail-open — a persistence error MUST NOT block the evaluator's primary stdout-JSON duty.
3. `internal/cli/goal.go:444`: replace `goal.RenderDashboard(g, nil)` with a load-then-render sequence — `v, _ := goal.LoadVerdict(root, sessionID)` (fail-open to nil), then `goal.RenderDashboard(g, v)`.
4. `internal/goal/prune.go`: extend `PruneOrphans` (and `ClearGoal`) to move/remove the `.verdict.json` sibling alongside `.json` and `.html` (per AC-GHF-003/004 precedent).

**Design decision recorded**: verdict persistence uses a sidecar file (NOT extending the existing `<session>.json` goal-state schema, NOT re-running the evaluator). The sidecar is written by the evaluator, read by the render command — single-writer / single-reader, no concurrency hazard.

**RED test first**: a characterization test that fails until `LoadVerdict` returns the verdict written by `SaveVerdict` on the same session (round-trip). Then a test that fails until `runGoalRender` passes a non-nil verdict to the renderer when a sidecar exists.

### M2 — Surface 3 re-arm UI construction (consumes existing data)

**Deliverables**:
1. `internal/cli/goal.go` `runGoalRender`: after loading `g` and `v`, load `.moai/state/handoff/pending.json` via the existing `handoff` package loader; check for a post-`/clear` new-session goal file (the session whose goal file was created by `rearmEmbeddedGoal`). Construct a `*goal.ReArmContext` populated per the AC-GHF-007 mapping:
   - `EmbeddedCondition` / `EmbeddedMaxTurns` / `EmbeddedMaxDuration` / `EmbeddedCostCap` / `EmbeddedUnbounded` ← from `pending.json` `EmbeddedGoal` (when non-nil).
   - `NewSessionID` ← the post-`/clear` session whose goal file exists (when one exists for this goal's lineage).
2. Pass `goal.RenderDashboardReArm(g, v, reArm)` (NOT `RenderDashboard` — escalate to the re-arm-aware variant; `nil` reArm produces byte-identical output per AC-GHF-007).
3. Fail-open: missing `pending.json`, missing `EmbeddedGoal`, missing new-session file → `reArm` is nil or partial; render still succeeds.

**RED test first**: a test that fails until a non-test code path constructs a `ReArmContext` and passes it to `RenderDashboardReArm`. The grep-verified non-test construction site is load-bearing (acceptance.md AC-WIRE-006).

### M3 — Surface 2 `moai plan render-html` Cobra verb (new CLI surface)

**Why separate from M1/M2**: this milestone introduces a NEW top-level Cobra command (`moai plan`). The decision to create a new `plan` CLI namespace (vs. folding under an existing verb) is the second-most-reversible decision in this SPEC — landing it after the data-model milestone but before the template rewrite lets the reviewer course-correct the CLI shape if the naming is wrong.

**Deliverables**:
1. `internal/cli/plan.go` (NEW file): `newPlanCmd() *cobra.Command` returning a `plan` parent command. Register `render-html` subcommand. Parent registered on `rootCmd` via `rootCmd.AddCommand(newPlanCmd())` in `internal/cli/root.go` (near the existing `newGoalCmd` registration).
2. `render-html` subcommand: positional `<SPEC-ID>`. Resolve `specDir` = `<root>/.moai/specs/<SPEC-ID>/`. Resolve `reviewFile` = most recent `.moai/reports/plan-audit/<SPEC-ID>-review-*.md` (highest `<N>`; absent → empty string). Call `planhtml.RenderPlanHTML(specDir, reviewFile)`. Write bytes to `<root>/.moai/reports/plan-html/<SPEC-ID>-plan.html` (mkdir if absent). Exit 0 on success; non-zero + stderr on missing SPEC dir (REQ-WIRE-007).
3. `--json` output flag mirroring `goal render` shape (`action`, `spec_id`, `path`, `bytes`).
4. `internal/cli/plan_test.go` (NEW): round-trip test against a fixture SPEC dir in `t.TempDir()`; DOM-parse the written `.html` via `golang.org/x/net/html` asserting goal + 8-field contract presence (AC-WIRE-004).

**RED test first**: `TestPlanRenderHTML_WritesFile` fails until the Cobra command exists, is registered, and writes the file.

### M4 — spec-assembly.md template rewrite + subagent boundary test

**Deliverables**:
1. Edit `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` Step 2.3.3a: replace step 1 ("Invoke the plan-HTML renderer: `RenderPlanHTML(...)`") with "Execute the CLI verb: `moai plan render-html <SPEC-ID>`" — the orchestrator can run shell commands. Preserve the fail-open clause and the `[HARD]` Implementation Kickoff Approval paragraph verbatim.
2. `make build` to regenerate the embedded template FS.
3. §25 neutrality self-check: the rewrite must NOT add new internal SPEC IDs / REQ tokens / commit SHAs. `RenderPlanHTML` already appears in the template file (verified via `grep 'GHF\|RenderPlanHTML' internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` — only `RenderPlanHTML` is present, NO `SPEC-GHF-*` / `REQ-GHF-*` / `AC-GHF-*` tokens); it is the sole permitted renderer-cross-reference, and the rewrite preserves it as-is.
4. `internal/cli/plan_subagent_boundary_test.go` (NEW): `TestPlan_NoAskUserQuestion` mirroring `internal/cli/web_test.go::TestWeb_NoAskUserQuestion`. Also extend `internal/cli/goal_test.go` (or a sibling file) to re-assert `TestGoal_NoAskUserQuestion` for the modified `goal.go` path (AC-WIRE-008 re-assertion).

### M5 — Cross-cutting guards + regression

**Deliverables**:
1. Run the full GOAL-HTML-FLOW-001 AC suite (`go test -run TestGHF ./internal/goal/... ./internal/report/planhtml/...`) — confirm 11/11 still green (predecessor regression guard).
2. Run the SPEC-INFINITE-GOAL-001 re-arm tests (`go test -run TestReArm ./internal/hook/...`) — confirm green (re-arm mechanism regression guard).
3. `gofmt -l .` clean; `go vet ./...` clean.
4. Lint + cross-platform build final pass.

## §G. Anti-Patterns

- **AP-1 — AC verifies renderer scope only** (the AUTONOMY-TIERS lesson): acceptance.md MUST verify end-to-end wiring (construct fixture → invoke CLI / production path → file-on-disk → DOM-verified). A unit test asserting only "bytes contain substring X" fails the AC. Cite AC-WIRE-001/004/006.
- **AP-2 — Recomputing the verdict at render time**: rejected. REQ-WIRE-002 persists; render loads. Do NOT call the evaluator from `runGoalRender`.
- **AP-3 — Extending the `<session>.json` goal-state schema to carry the verdict**: rejected. Sidecar file (`.verdict.json`) keeps the writer/reader pair single-purpose and avoids a schema migration on the existing goal-state file.
- **AP-4 — Folding `moai plan render-html` under an existing verb** (`moai render`, `moai report`): rejected per user-confirmed scope decision 3. The new `plan` CLI namespace matches the `goal render` parallel structure.
- **AP-5 — Touching SPEC-INFINITE-GOAL-001 mechanism**: forbidden. Consume `pending.json` + `EmbeddedGoal` + `rearmEmbeddedGoal` read-only.

## §H. Cross-References

- `.moai/specs/SPEC-GOAL-HTML-FLOW-001/` — predecessor (renderer source; stays completed)
- `.moai/specs/SPEC-INFINITE-GOAL-001/` — re-arm mechanism source (stays completed)
- `internal/cli/web_test.go::TestWeb_NoAskUserQuestion` — C-HRA-008 test precedent
- `internal/cli/goal.go:444` — Surface 1 site
- `internal/report/planhtml/renderer.go:54` — Surface 2 renderer
- `internal/goal/dashboard.go:69-140` — Surface 3 ReArmContext + applyReArm
- `internal/hook/handoff_inject.go:208` — rearmEmbeddedGoal (INFINITE-GOAL mechanism, consume read-only)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` §B3/B11 — C-HRA-008 boundary
- CLAUDE.local.md §2 + §25 — Template-First + neutrality
