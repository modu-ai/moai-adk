# SPEC-MANAGER-TODO-001 — plan.md

Implementation plan. Card: t1306 (Class C, Tier L). Branch: `WT-manager-todo-agent` (base: local develop `b7ff456b7`).

## A. Context

Operator directive 2026-09-29: rename and repurpose the `mission-governor` agent into `manager-todo` — dedicated todo-queue management + Jev decision wiring + card dispatch/management ownership + a new `/moai:todo --auto` serial processing mode. The card's linked design artifact URL was unreadable from this lane; the card text is the requirements canon. Full measured context: research.md; design decisions: design.md.

Method: DDD (ANALYZE-PRESERVE-IMPROVE) — the rename preserves the judgment sub-role's contract textually; the `--auto` mode is new behavior with RED-first table-driven tests on the pickup predicate.

## B. Known Issues

- `[NEEDS CLARIFICATION: codex read-only role roster disposition]` — design D-3 removes manager-todo from the Codex read-only role roster, which removes the mission-judgment role from the Codex path. The lead must confirm this removal (or elect the larger sub-role-sandbox redesign) before run-phase entry. Tracked here and in research.md §C; **not** in spec.md/acceptance.md.
- Parallel factory card (`WT-factory-self-dispatch`, M4-M7 pending) overlaps `internal/cli/todo.go` — absorbed by the serial merge window; a semantic clash is a run-phase blocker report, never a forced merge (research.md §D).
- rosterguard fixture strings and delegationmap staleness comments enumerate the old name in prose that also encodes test expectations — renaming them is test-editing, not prose editing; run-phase must re-run the affected package tests, not trust the sweep.

## C. Pre-flight

1. Re-run the reference sweep at M2 start (`git grep -n -i "mission.governor" -- . ':(exclude).moai/reports' ':(exclude).moai/specs'`) and reconcile against research.md §B — line numbers decay as parallel cards land.
2. Confirm the C3 emission is clean before starting: `make agents-emit-check` exit 0 on the base.
3. Confirm the affected-package test baseline before any edit: `go test -timeout 30m ./internal/cli/... ./internal/harness/... ./internal/mission/... ./internal/template/...` (serialized, one package group at a time; no full-suite local run).
4. Confirm the queue store path in code (`internal/cli/todo.go` help text) before writing any `--auto` prose — the store is `~/.moai/db/<project-key>/todo/backlog.db`.

## D. Constraints

Restated from spec.md §C where they shape milestone execution: C-2 (C3 emit-only — plan schedules `make agents-emit` + `make agents-emit-check`, never a hand edit), C-3 (template neutrality for everything under `internal/template/templates/**` — no other cards' ids, no internal dates, no commit SHAs in template-mirrored files), C-4 (no t1240 dependency), C-5 (store description accuracy).

## E. Self-Verification

Run-phase closes each milestone with the canonical verification batch (single-turn multi-Bash, file-redirect contract, evidence exported to `.moai/reports/t1306/verdict.md`):

- Affected-package tests: `go test -timeout 30m ./internal/cli/... ./internal/harness/... ./internal/mission/... ./internal/template/...`
- Emission checks: `make agents-emit-check`, `make commands-emit-check` (command sources touched in M3), `make build` after template edits
- Sweep verdict: the REQ-MT-005 grep returning only dispositioned hits, pasted verbatim into the verdict
- Lint: `golangci-lint run` on changed packages
- Template neutrality: the CI guard locally (`make build` succeeds; neutrality spot-grep for forbidden classes in the diff)

## F. Milestones (priority-ordered; no time estimates)

### M1 — Agent rename/repurpose (Priority High)

1. Author C1 `.claude/agents/moai/manager-todo.md`: primary mission = todo-queue management + Jev boundary + dispatch/management ownership; preserved judgment sub-role section (design D-1); frontmatter per agent-authoring rules.
2. Mirror to C2 `internal/template/templates/.claude/agents/moai/manager-todo.md` (hand-edited, content-neutral per C-3).
3. Delete C1+C2 `mission-governor.md` copies; run `make agents-emit` (regenerates C3: `manager-todo.toml` in, `mission-governor.toml` out); `make agents-emit-check`; `make build`.
4. Verify REQ-MT-001..004 / AC-MT-001..004.

Risk: rosterguard/delegationmap/retained-agents Go surfaces break at compile/test time the moment the name changes — they are updated in M2, so M1's exit is "definition layer consistent + emission clean", with Go-surface updates immediately following in M2. Sequenced this way so each milestone owns a coherent test-green state: M1 and M2 are committed as one reviewable unit if the interim state cannot go green (decision left to run-phase; the plan prefers separate commits).

### M2 — Reference sweep refresh (Priority High)

1. Re-run the baseline sweep; reconcile per-file against research.md §B (dispositions: RENAME / FROZEN / HIST).
2. Update every RENAME hit: Go role registries (`codex_audit_mcp.go`, `codex_role_fingerprint.go`, rosterguard, delegationmap, `retained_agents.go`, `catalog.yaml`, mission receipt prose, `goal.go` flag help), test expectations, rules files + template mirrors, CLAUDE.md/AGENTS.md.tmpl, goal workflow skill + mirror.
3. Apply the D-3 roster disposition consistently across the four roster surfaces (gated on the clarification being resolved).
4. Codemaps: record pending-regen (design D-6).
5. Verify REQ-MT-005..006 / AC-MT-005..008 — the sweep grep returns only dispositioned hits.

### M3 — `/moai:todo --auto` serial mode (Priority High)

1. CLI: `--auto` flag on the todo command tree (design D-2); serial cycle implementation against current verbs only; pickup selection per REQ-MT-008/010 (two-channel liveness per D-4; positive state vocabulary per D-5); /clear guidance emission per D-7; batch-approval semantics per D-8.
2. Skill/command text: document the serial contract and the batch-approval reconciliation in the todo/gtd workflow text (thin-command pattern — mechanics in CLI).
3. RED-first: table-driven tests for the pickup predicate over all known states; liveness fixtures for registry-dead/lsof-clean, registry-alive, lsof-hit; serial-order test; guidance-emission test; no-factory-lease assertion (grep + behavior test).
4. Verify REQ-MT-007..013 / AC-MT-009..016.

Risk: highest-change-likelihood milestone (new interface + user-visible behavior) — plan review should focus here and on D-4/D-5.

### M4 — Jev boundary + GTD absorption wiring (Priority Medium)

1. Agent body Jev boundary section (REQ-MT-014/015) — already drafted in M1's body; this milestone verifies it against the grade-3 list with a body-content test (forbidden-authority phrases absent).
2. Goal workflow text + mirror name manager-todo (REQ-MT-016); receipt contract prose updated; `--governor-receipt` flag help updated (schema untouched, D-9).
3. Codex roster disposition applied (from M2 step 3); conflict report finalized in research.md §C.
4. Verify REQ-MT-016..017 / AC-MT-017..019.

### M5 — docs-site 4-locale + README 4-locale sync (Priority Medium)

1. Update all RENAME-dispositioned docs-site hits across en/ko/ja/zh (per-locale line numbers in research.md §B) with section parity.
2. README.{md,ko.md,ja.md,zh}.md updates (2 hits each) with 4-file parity.
3. Template-neutrality checks over the full milestone-set diff; codemaps regen noted.
4. Verify REQ-MT-018 / AC-MT-020..022.

### Milestone dependency notes

M1 → M2 → {M3, M4} → M5. M3 and M4 are independent of each other. M5 depends only on the surfaces it documents (can start once M2 lands). The M1+M2 interim-state caveat is the only sequencing flexibility.

## G. Anti-Patterns

- Hand-editing C3 `.toml` — forbidden (C-2); the only verb is `make agents-emit`.
- Renaming FROZEN testdata (`internal/cli/testdata/codex-rollouts-t1171/**`) — falsifies a recorded historical fixture.
- Expressing pickup as `state != done` — breaks hold-forward-compatibility (D-5).
- Trusting the session registry alone for owner liveness — the project lesson mandates the lsof cwd channel.
- Describing the queue store as `.moai/state/kanban/backlog.json` — stale prose; the store is the home SQLite db (C-5).
- Taking a factory lease/slot or importing unlanded factory code (REQ-MT-013).
- Letting Jev output mutate the queue or answer a completion verdict (REQ-MT-015).

## H. Cross-References

- research.md §B — the 262-hit baseline (M2 checklist); §C — roster conflict report; §D — t1240 seam.
- design.md D-1..D-10 — all design decisions.
- `.claude/rules/moai/development/agent-authoring.md` (frontmatter + three-copy rule) · `internal/template/agentemit/` (emitter contract) · `.moai/docs/template-internal-isolation-doctrine.md` §25 (neutrality) · AGENTS.local.md §29 (Jev boundary, lane-local doctrine — not citable from templates).
- Parallel: SPEC-FACTORY-SELF-DISPATCH-001 (seam only).

## I. Decision Points

- **DP1 — plan-phase entry (approval proxy).** The card was operator-approved, and the lead issued the explicit plan-commencement directive dated 2026-09-29 ("plan 착수"). Recorded as the approval proxy; no AskUserQuestion is available or required inside a lane. Disposition: plan-phase proceeded on this proxy; run-phase entry still requires the standing kickoff gate.
- **DP2 — codex read-only role roster disposition.** Open; `[NEEDS CLARIFICATION: codex read-only role roster disposition]` (research.md §C, design D-3). Must be resolved by the lead before run-phase M2 step 3.
- **DP3 — M1/M2 commit granularity.** One unit or two commits depending on whether the interim state can go test-green; run-phase decides on measurement (M1 risk note).
