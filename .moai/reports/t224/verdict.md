# t224 — Lane Standing Spawn Authority (2026-09-02, lane-7)

## Claim

Factory lanes and kanban companions now carry a STANDING agent-spawn authority in their own
bootstrap context — closing the tk8hce defect class where lanes refused to spawn the
phase-required specialist (runtime default guidance stood unoverridden) and fell back to
editing SPEC bodies directly, routing artifact writes around the Status Transition Ownership
Matrix. The card's three design questions are decided and encoded; the card's regression
requirement (execution-based, not grep) is discharged by firing the real binary.

## The three design decisions (the card's judgment items)

1. **Scope — ownership-matrix-bounded, not arbitrary.** The lane spawns the specialist the
   matrix names for the work at hand: plan → `manager-spec`, run → `manager-develop`,
   sync → `manager-docs`, plus the workflow chain's prescribed auditors. Principle-phrased
   rather than enumerated, so the bound tracks the matrix instead of drifting from it.
2. **Depth — depth-1 seal.** Agents a lane spawns are leaf workers and never spawn further
   agents — the same flat-hierarchy shape `manager_lead_depth_test.go` enforces for the
   lead's own fan-out.
3. **Placement — the bootstrap context is the operative layer.** The tk8hce lesson is that a
   lead's approval CANNOT lift a lane's session instruction (the lead is not the lane's
   user), so the authority rides what the lane reads at startup: the SessionStart join
   notice. The doctrine files carry the normative text; no runtime permission change was
   needed (measured: `Agent` already ships in the template's `permissions.allow`, and the
   launcher seeds the per-lane concurrent-subagent cap — `seedLaneAgentCap`, the t118 axis).

## Evidence

All measurements 2026-09-02, worktree `.claude/worktrees/t224`, branch `WT-lane-spawn-authority`
(develop `45bb6dacd` + these commits).

| Check | Command | Result |
|---|---|---|
| New regression tests | `go test ./internal/hook -run 'LaneSpawnAuthority\|Notice\|Bootstrap' -count=1` | ok |
| Full touched package | `go test ./internal/hook -count=1` | `ok ... 39.687s` |
| Template emit golden | `go test ./internal/template -run 'AgentEmit\|Catalog' -count=1` (via `make agents-emit`) | ok |
| Lint | `golangci-lint run ./internal/hook/...` | `0 issues.` |
| Cross-platform build | `GOOS=windows` + `GOOS=linux go build ./internal/hook/...` | both exit 0 |
| Mirror parity | `diff -q` template vs local (kanban-dispatch.md, manager-lead.md sampled) | identical |

**Execution-based regression (the card's [HARD] item) — the real binary, not a grep:**

```
MOAI_FACTORY_WORKER=lane-9 MOAI_FACTORY_WORKERS=8 CLAUDE_PROJECT_DIR=<tree> \
  ./bin/moai hook session-start < startup.json
→ grep -c "Standing spawn authority" out.json → 1

MOAI_KANBAN_LABEL=run CLAUDE_PROJECT_DIR=<tree> \
  ./bin/moai hook session-start < startup.json
→ grep -c "Standing spawn authority" out.json → 1
```

Both lane flavors answer with the authority sentence in their `additionalContext` output —
the rendered context the lane session actually reads, executed through the full hook
dispatcher path.

**Contract updates to existing tests (legitimate breaks, each justified):**
`TestFactoryWorkerNoticeNamesLabel` and `TestKanbanCompanionNoticeJoinOnly` pinned
whole-output equality / single-line on the join notice; the authority is now the one
deliberate second block, so the assertions became prefix-presence (join line intact) and
exactly-two-blocks (join + authority), keeping the launch-block prohibition (`--name`
absence) intact. `TestKanbanCompanionNoticeRoleless` caught a real wording hazard — the
draft sentence's "as the lane session" tripped AC-FB-016's role-clause check — and the
sentence was reworded rather than the guard weakened.

## Files

| File | Change |
|---|---|
| `internal/hook/lane_spawn_authority.go` | NEW — the authority sentence + the tk8hce provenance + the three design decisions |
| `internal/hook/session_start_factory.go` / `session_start_kanban.go` | both join notices append the sentence |
| `internal/hook/lane_spawn_authority_test.go` | NEW — rendered-output assertions + fail-open preservation |
| `internal/hook/session_start_factory_test.go` / `session_start_kanban_test.go` | contract updates (above) |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | § Lane spawn authority (standing) |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | Factory in-lane 3-stage: normative bullet |
| `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | § User Interaction Boundary: lane sessions are orchestrator-class |
| `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` | § MoAI Orchestrator: delegation duty + authority bind lanes |
| `internal/template/templates/.claude/agents/moai/manager-lead.md` | Role B: never micromanage lane authority; a "told not to spawn" report is a defect |
| `internal/template/templates/.codex/agents/moai/manager-lead.toml` | re-emitted by `make agents-emit` |
| local mirrors (`.claude/rules/moai/**`, `.claude/agents/moai/manager-lead.md`) | byte-identical copies |
| `CHANGELOG.md` | [Unreleased] Changed entry |

## Baseline-attribution

Measured in this run, this tree (`WT-lane-spawn-authority` @ develop `45bb6dacd` + these
commits). Package-scoped per dispatch rule 5; no full suite locally — CI on the develop push
is the full-suite verdict.

## Gaps

- **Runtime tool layer out of reach (reported, not fixed):** this lane's own session —
  lane-7 on a GLM backend — has NO Agent tool at all in its function list. The doctrine and
  bootstrap text cannot conjure a tool the backend channel withholds; that is a separate,
  runtime/backend-layer gap to verify separately (which session types get the Agent tool,
  and whether GLM-backend lanes can spawn at all). The tk8hce defect was the instruction
  layer; this observation is the tool layer.
- The authority sentence is English-only by the two-audience rule (agent-facing
  `additionalContext` is rendered `langEnglish` at both call sites); operator-facing
  `systemMessage` is untouched, so no 4-locale notice translation was added.
- `manager_lead_depth_test.go`'s depth seal was not extended to assert anything about lanes
  (lanes are sessions, not agent definitions — the seal binds agent definitions); the depth
  bound for lanes is doctrine text, enforced by review, not a unit-testable seal here.

## Residual-risk

- A lane whose backend withholds the Agent tool will now read an authority it cannot
  exercise — the sentence names the Agent tool explicitly, so the mismatch is visible (the
  lane can report it precisely) rather than silent.
- The scope bound is principle-phrased; a lane arguing "the matrix requires X" for an
  unusual spawn is a judgement the review lenses and the lead's evidence-reading already
  police.
