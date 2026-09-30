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

Run tree: 04a646a27 (run baseline) → 147c25d77 (M1) → 7e612e0f1 (M2) →
d2b45d1ca (M3) → the M4 progress commit; branch WT-lane-stall-watchdog,
worktree .moai/worktrees/t1370. All commands run in this run, this tree.

### RED-first evidence (E8; observed at 04a646a27 BEFORE each flip)

Verbatim observation set at the run baseline (24 probes; the deciding
outputs): E-1 `test -f .claude/rules/moai/workflow/auto-semantics.md` →
(empty) · exit 1. E-2a/E-2b/E-11/E-12a/E-12b/E-13/E-14 doc probes →
`grep: .claude/rules/moai/workflow/auto-semantics.md: No such file or
directory` · exit 2 each. E-3 `test -f .claude/skills/moai-lane-watchdog/SKILL.md`
→ (empty) · exit 1. E-4's four label probes (blocked-by / awaited-judgment /
shell-error / accidental-stop), E-5, E-6, E-14-skill, E-15 both →
`grep: .claude/skills/moai-lane-watchdog/SKILL.md: No such file or directory`
· exit 2 each. E-6b/E-7/E-8/E-9 measured zeros → `0` · exit 1 each.
E-10 → `diff: .claude/skills/moai-lane-watchdog: No such file or directory`
· exit 2. E-1b already GREEN at baseline (2 · exit 0 — flipped by iter-2
authoring, guarded through close). The full verbatim transcript was captured
in this session's tool output; each green below cites its RED root.

### AC PASS matrix (E1; closing tree d2b45d1ca, this run)

| AC | Probe (command form) | Observed | Verdict |
|---|---|---|---|
| AC-LSW-001a | test -f auto-semantics.md | exit 0 | PASS |
| AC-LSW-001b | grep -c orchestration-mode-selection spec.md | 2 · exit 0 | PASS (guard) |
| AC-LSW-002a | grep -c "autonomous adjudication" doc | 1 · exit 0 | PASS |
| AC-LSW-002b | grep -c "display-only" doc | 1 · exit 0 | PASS |
| AC-LSW-003 | test -f SKILL.md | exit 0 | PASS |
| AC-LSW-004a-d | label counts in SKILL.md | 2/2/2/2 · exit 0 | PASS |
| AC-LSW-005 | grep -cE integration[- ]window SKILL.md | 1 · exit 0 | PASS |
| AC-LSW-006 | concept tokens in SKILL.md (2) + SPEC-ID in local-only (E-6b: 1) | all exit 0 | PASS |
| AC-LSW-007a | grep -c "explicit wait" kanban-dispatch.md | 1 · exit 0 | PASS |
| AC-LSW-007b | grep -c moai-lane-watchdog kanban-dispatch.md | 1 · exit 0 | PASS |
| AC-LSW-009 | grep -c "explicit wait" gitflow-lane-protocol.md | 1 · exit 0 | PASS |
| AC-LSW-010a/b | diff -r skill pair / diff doc pair | exit 0 both | PASS |
| AC-LSW-011 | git diff --name-only <merge-base 3dd5adf2f>..HEAD -- '*.go' | 0 files | PASS (guard) |
| AC-LSW-012 | spec lint | No findings · exit 0 | PASS (guard) |
| AC-LSW-013a | grep -c update_plan doc | 2 · exit 0 | PASS |
| AC-LSW-013b | grep -c session_msg doc | 1 · exit 0 | PASS |
| AC-LSW-014 | E-12a kickoff approve (1) + E-12b keep-set alternation (4) | both exit 0 | PASS |
| AC-LSW-015 | record-shape regex | 1 · exit 0 | PASS |
| AC-LSW-016a | jev_ask in doc | 4 · exit 0 | PASS |
| AC-LSW-016b | jev_ask in skill | 1 · exit 0 | PASS |
| AC-LSW-017a | update_plan in skill | 1 · exit 0 | PASS |
| AC-LSW-017b | TaskCreate/TaskList in skill | 1 · exit 0 | PASS |

26/26 rows green. Substitutes per plan §E3: zero Go changes exist, so
coverage is N/A and E4+E5 carry the substitute gates — E4 boundary grep
`grep -rn AskUserQuestion` over the two NEW artifacts → zero matches (exit 1);
E5 make build exit 0 (final, d2b45d1ca build), neutrality scoped test ok,
spec lint exit 0. Stale-surface guard: no scripts/jev or ask.sh reference in
either new artifact (grep exit 1).

### E9 amendment sweep (spec.md §A.7 rows A-N + reviewed-no-change)

Applied (both sides for mirrored files; byte-identical parity verified by
diff -q per pair): A orchestration-mode-selection header + progression note +
§E sweep anti-patterns + §C.3 precondition rows + §D.1 #5; A
askuser-protocol Recommendation-mode sentence; B spec-workflow Plan-to-Run
trigger + skip-policy tail; C run.md §1 ordering + two HARD paragraphs, C
goal-directive (3 anchors), C goal-directive-detail (T1 + 5 note hits); D
session-handoff invariant bullet + Block 1 SEED note + Block 5; F
kanban-dispatch batch-authorization sentence + foreman SKILL boundary 1; G
AGENTS.local kickoff policy line + kickoff-autonomy.md §1; H
contract-autonomy reserved ## Autonomous Kickoff section FILLED; I
AGENTS.local §19/§19.1 reference line; J kanban-dispatch Boundaries bullet;
K cadence-bridge governing paragraph + cross-reference line; L skills
goal.md heading + intro + Safety Invariant 1; M CLAUDE.md §2 stage 4
(template first); N cache-aware-execution Non-goals bullet. Residual scan:
no "mandatory-restoration / mandatory and score-independent" wording remains
in any amended file; a broader Kickoff-mention sweep found only form-neutral
mentions, dispositioned below.

Reviewed-no-change dispositions (re-read at M3/M4): dynamic-workflows.md —
"the gate governs launching the batch" / "separate from the plan-to-implement
gate" are form-neutral (the gate applies in both forms; the batch is still
not the approval) — no change. .claude/skills/moai/workflows/plan.md —
describes the gate as executed; both forms require clarifications resolved —
no change. session-handoff-examples.md — zero hits for the amended invariant
text; examples carry no conflicting wording — no change. .claude/loop.md —
zero Kickoff hits — no change. archived-agent-rejection.md §E — the
anti-pattern targets SILENT substitution losing the gate signal; under the
new default the decision record replaces that signal, so the semantics hold;
the stale policy-name parenthetical is historical — no change. Form-neutral
label-stale mentions kept with disposition: orchestration-mode-selection
contract-mode block (preserved equivalent form), §B.1b "does NOT relax" +
dynamic-workflows "decided by the orchestrator" (true under both forms — the
orchestrator adjudicates the autonomous form before launch).

factory decide help text unchanged (E9): `moai factory decide --help` →
"Record an operator decision: --gate kickoff --choice approve|reject, --gate
push, or --choice resume|block|unblock|abandon" · exit 0 — human-only
wording preserved, zero Go delta.

Cascade follow-up (render-surface sentinel): the session-handoff drift-mitigation
sentinel requires the render surface (.claude/output-styles/moai/moai.md §8)
to match the amended SSOT. Three compact Kickoff clauses there carried the
old mandatory wording; all three were synced to the default-autonomous
standard (lines ~685/714/716), mirror applied, byte-identical verified, and
make build re-run exit 0. Attributable to this SPEC's scope envelope as
row D's render-surface pair.

### §D.3 review pass (behavioral verification of the model-mediated procedure)

Skill walked against spec.md §B.2 step by step: ladder carriers ①-⑤ present
in the skill §3.1 in order; outcome→action table matches the doctrine §7
row-for-row (① negative → never proceed; ② empty board → ③, never wait;
③ negative/inconclusive at an authority gate → fail-closed + escalate ⑤;
④ gated off → ⑤ fail-open; ⑤ explicit wait record); availability failures
fail-open downward stated in both. Per-runner dry read-through: Claude runner
path uses TaskCreate/TaskUpdate views + session-messaging ⑤; codex path uses
update_plan view + broker (session_msg_register/session_msg_send + poll) +
queue-on-disk ⑤ substitute — both walk the same disk SSOT. §D.6 edge cases:
all nine covered (docs-only progress not a stall §1; dropped predecessor
names the operator escape without running it §3.2; determinate error zero
retries §3.3; empty board → ③; authority-gate fail-closed §0+step 3;
gate-off → ⑤; KEEP-gate awaken rechecks and yields §0; codex no-messaging
§4; fail-open first observation §1).

### Gaps and residual

- R-4 codex-premise re-check: the DOCUMENT content (the §B.6/§8 neutrality
  table rows and their stated premises) verified against the SPEC text; the
  EXTERNAL re-check against current OpenAI Codex documentation could NOT be
  executed in this session — no web tool is available to this run-phase
  agent. The update_plan shape and TaskList-absence remain lead-reported
  premises (acceptance §A reported-premise criteria verify document content
  only). Carried as an explicit gap for the sync audit or the leader.
- The decision board's concrete HOME-surface path is fixed by the doctrine
  (§11, moai-home state directory keyed by project key) as designed; no
  runtime exercise of a real ladder resolution was performed (the procedure
  is model-mediated; the dry read-through is its verification instrument per
  acceptance §D.3).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-30
run_commit_sha: pending-backfill-run (implementation complete at d2b45d1ca;
the M4 progress commit carries this section and cannot cite its own SHA)
run_status: complete
ac_pass_count: 26
ac_fail_count: 0
preserve_list_post_run_count: 0 (violations; .claude/loop.md + mirror,
.moai/config/**, internal/*.go, pkg/, cmd/ all untouched — measured
04a646a27..HEAD)
l44_pre_commit_fetch: n/a (isolated worktree lane; no shared-checkout commit)
l44_post_push_fetch: n/a (no push from this lane — integration is the
serial develop window per the lane protocol)
new_warnings_or_lints_introduced: 0 (no Go changes; golangci baseline clean;
spec lint exit 0 at every commit; template neutrality + leak tests ok)
cross_platform_build.darwin: pass (go build ./... exit 0)
cross_platform_build.windows: pass (GOOS=windows GOARCH=amd64 go build ./...
exit 0)
total_run_phase_files: 38 (04a646a27..d2b45d1ca) + this progress commit
m1_to_mN_commit_strategy: per-milestone commits M1/M2/M3 + M4 progress
commit; card t1370 marker + Authored-By-Agent trailer on every commit;
explicit pathspec staging only; no push

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
