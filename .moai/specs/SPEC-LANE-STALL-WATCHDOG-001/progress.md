# progress.md — SPEC-LANE-STALL-WATCHDOG-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-30
plan_phase_tree: 3dd5adf2f (0.1.0) → revision commit (0.2.0), branch
WT-lane-stall-watchdog
note: iter-2 artifact set complete (spec/plan/acceptance/progress, Tier M),
frontmatter schema-validated, SPEC-ID regex Bash check PASS, spec lint clean
at each plan-phase commit; plan-auditor (--deep) verdict is the orchestrator's
next step and has not yet run.

## §A Overlap Adjudication (plan-phase record)

t1343 SPEC-RELATION-PICKUP-FILTER-001 (status: completed; filter code present
in this tree at `internal/cli/todo_auto.go:162-187`) owns the QUEUE-level
pickup decision. Iter-2 EXCLUDES the queue surface from this card's coverage
entirely (operator-unused), so the two SPECs now share zero coverage: t1343
owns pre-pickup exclusion, this card owns the in-flight lane remedy. The
blocked-by remedy reads the same data surfaces (relation findings,
`.moai/reports/<card-id>/` evidence) directly from disk and performs zero
queue mutation — it never picks, drops, edits, or unrelates. REQ-LSW-004
carries the boundary; AC-LSW-006 verifies the skill states it.

## §B Doc-vs-Code Surface Enumeration (plan-phase record, iter-2)

| Kind | Path | Mirror |
|---|---|---|
| NEW rule | `.claude/rules/moai/workflow/auto-semantics.md` | template-first + `make build` |
| NEW skill | `.claude/skills/moai-lane-watchdog/SKILL.md` | template-first + `make build` |
| EDIT rules | `.claude/rules/moai/workflow/kanban-dispatch.md` + the §A.7 amendment-list files (each mirrored) | template mirror same change |
| EDIT local-only | `.claude/rules/local/gitflow-lane-protocol.md` §6 | NEVER mirrored |
| unchanged | Go (`internal/`, `pkg/`, `cmd/`), `.moai/config/**`, `.claude/loop.md` (+ mirror), `factory decide` code | — |

The iter-2 content growth (ladder, neutrality table, gate inventory, decision
records) lives INSIDE the same two new files — the file set is unchanged from
0.1.0; the zero-Go boundary survives the autonomous transition because
decision records go to the disk board, not to `factory decide` (which stays
human-only, spec.md §A.7 row E). `.claude/loop.md` is template-managed
(mirror measured, 1254 bytes) and deliberately untouched — the lane-side
awaken path is doctrine + skill, not a second loop driver.

## §C --auto Surface Inventory (plan-phase record, iter-2 scoping)

- lane (factory/kanban worker; `moai cc` / `moai glm` / `moai codex`) — the
  COVERED surface: watchdog + ladder + doctrine (spec.md §A.3).
- `todo --auto` — OUT OF SCOPE (operator-unused): queue serial cycle,
  `--auto-wait` 30 min evidence deadline + unpick + dead-owner rescue
  (`internal/cli/todo.go:313`, `todo_auto.go:17-27, 142-161, 276-328`) —
  cited as design precedent only.
- `goal --auto` — OUT OF SCOPE (already bounded): mission lifecycle + turn
  ceiling + stagnation + wall-clock (`goal.go:148-154, 194`;
  `evaluate.go:325, 337-347, 358-363`) — boundary inventory only.
- Jev — the single judgment surface is the `jev_ask` MCP tool
  (`mcp_server.go:650-658`, wraps `jev.Client.Ask`) behind
  `workflow.jev.enabled` (default false, `types.go:522, 783`); `scripts/jev/`
  measured ABSENT — the earlier ask.sh premise was stale and is stricken.

## §D Scope Revision Record (plan-phase iter-2 — operator feedback via lead)

1. **Queue surface excluded** (directive 1): `todo --auto` batch cycle leaves
   coverage; 30m unpick = design precedent only. Applied in spec.md §A.3,
   §D, REQ-LSW-007; plan.md §B2/§G.
2. **Lead-wait elimination via the ladder** (directive 2 — primary goal):
   t1311-style reply waiting is removed as a stall cause; the five-step
   ladder (disk evidence → decision board → MCP audit cross → jev_ask → lead
   chat last resort) is REQ-LSW-003. The lead's ⑤-carrier note (SendMessage
   absent on codex; broker polling + queue-on-disk substitute) verified
   against `mcp_server.go:550, 566` and SPEC-CODEX-SESSION-MSG-001 (landed).
3. **Harness neutrality REQ** (directive 2 addendum): REQ-LSW-010 + the
   §B.6 per-runner table; hook-independence chosen over dual-parity hook
   work (t1099 binds future hook surfaces only).
4. **Gate inventory + default-autonomous transition** (directive 3):
   REQ-LSW-008/012/013; the inventory + dispositions + three-category
   keep-set in spec.md §B.5; the amendment list §A.7 (rows A-H). Premise
   verified: `factory decide` surface at `factory_card.go:1469-1485` (verbs
   + `DeciderHuman`-only + lane refusal) — disposition: code UNCHANGED,
   autonomous records go to the decision board.
5. **Jev surface correction** (directive 3 principle 2): `scripts/jev/`
   measured absent (`ls`: No such file or directory) — the ask.sh premise
   from the earlier brief was STALE; the canonical surface is `jev_ask`
   (measured `mcp_server.go:650-658`) behind `workflow.jev.enabled` (measured
   default false). All ask.sh mentions stricken; REQ-LSW-009 rewritten;
   repeatable-parameter + stall-judgment-only scope recorded.
6. **update_plan discovery** (directive 4): codex's native TODO tool adopted
   as the codex-side rendering view (REQ-LSW-011); SSOT stays disk on both
   runners; view–SSOT rule codified; ExecPlan/PLANS.md marked reference-only.
   NOTE: `update_plan` schema and TaskList-absence-on-codex are LEAD-REPORTED
   external measurements (OpenAI Codex Prompting Guide) — not locally
   verifiable from this tree; registered in acceptance.md §E.2 and §G R-4.
7. **CLI verb hygiene** (directive 1 note): "advance" measured absent from
   the todo verb set (add / list / done / next / unpick — `todo.go:685-1315`);
   "lane self-advance" is used as a concept name only.
8. **plan-audit iter-3 repair** (verdict `.moai/reports/t1370/plan-audit-verdict.md`,
   FAIL 0.80 at `dc3710d95`, blocking D1-D7 + optional D8-D11 — all applied;
   delta re-audit is the LAST round):
   - **D1** — E-6 re-targeted to concept tokens (`pickup filter|queue-readonly`);
     the skill carries NO internal SPEC ID; the composition citation moved to
     the local-only surface (new probe E-6b on `gitflow-lane-protocol.md`,
     RED `0`/exit 1 observed) — mirror-parity × neutrality × citation now
     jointly satisfiable.
   - **D2** — §A.7 extended: 8 measured surfaces added as rows I-N (all
     re-verified this tree: AGENTS.local.md:428 §19/§19.1, kanban-dispatch
     §Boundaries, cadence-bridge:26, skills goal.md:263, CLAUDE.md §2,
     cache-aware-execution:31, goal-directive-detail ×5, archived-agent-
     rejection) + the auditor's 5 reviewed-no-change surfaces recorded with
     dispositions; plan E9 sweep re-scoped to the FULL list.
   - **D3** — three-way traceability alignment: §D.2's REQ-LSW-003 → AC-LSW-004a
     only; §C.1 rewritten with explicit a/b ids; matrix/§D.2/§C.1 now name
     the same relations.
   - **D4** — probes re-targeted to what their criteria claim: E-2 split into
     E-2a ("autonomous adjudication") + E-2b ("display-only"); E-12 split
     into E-12a (kickoff approve) + E-12b (keep-set token alternation); E-13
     counts the §B.8 record-shape line (`decided_by=.*evidence_refs=.*ladder_path=`).
   - **D5** — contract-signing gate added to the §B.5 inventory (PRESERVED
     EQUIVALENT FORM + justification); §A.7 row H names the reserved
     `## Autonomous Kickoff` section (measured contract-autonomy.md line 142)
     as an M3 amendment target.
   - **D6** — §B.2 outcome→action table added; authority-gate fail-closed
     invariant stated (negative/inconclusive audit never proceeds);
     availability failures fail-open downward.
   - **D7** — §B.1 observation-snapshot rule added (path
     `.moai/state/watchdog/<card-id>.json` lane-local, fields
     observed_at/head_sha/evidence_mtime_max/window_state, two-point
     comparison + N-minute window, single-writer, fail-open first
     observation); REQ-LSW-001 references it.
   - **D8** — ghost-db wording re-measured: absent in the current tree
     (0-byte at `3dd5adf2f`); §A.6/§B.7 state both measurements.
   - **D9** — acceptance.md §A states the family-count rationale (25 rows =
     16 families = Tier M ceiling exactly met).
   - **D10** — §B.6 names the awaken carrier as scheduler-mediated out of
     scope; the rule binds the awaken turn's first action, not the scheduler.
   - **D11** — REQ-LSW-003's non-action tail clause removed; REQ-LSW-009
     rewritten as a proper (Unwanted) shall-not form.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
