# SPEC-MANAGER-TODO-001 — progress.md

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase status: complete (artifacts authored 2026-09-29 by manager-spec, worktree `WT-manager-todo-agent`, base develop `b7ff456b7`).

### Plan-phase evidence (commands run this session + verbatim output)

1. **SPEC ID pre-write self-check** (HARD protocol):

```
$ ID="SPEC-MANAGER-TODO-001" && [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

2. **Reference-sweep baseline** (research.md §B; REQ-MT-005 checklist):

```
$ git grep -n -i "mission.governor" -- . ':(exclude).moai/reports' ':(exclude).moai/specs' | awk -F: '{print $1":"$2}' | wc -l
     262
$ git grep -i "mission.governor" -- . | wc -l
     373
```

Live-surface baseline: 262 hits / 105 unique files (hit count: the sweep piped to `wc -l` = 262; unique-file aggregation: same command piped through `cut -d: -f1 | sort -u | wc -l` = 105 — the iter-1 draft's "95 files" was an under-count, corrected per audit finding D7), enumerated per file:line in research.md §B. The 111-hit remainder under `.moai/reports/**` + `.moai/specs/**` is historical and dispositioned out of scope.

3. **Three-copy confirmation**:

```
$ ls .claude/agents/moai/ | grep mission-governor
mission-governor.md
$ ls internal/template/templates/.codex/agents/moai/ | grep mission-governor
mission-governor.toml
```

(C2 mirror `internal/template/templates/.claude/agents/moai/mission-governor.md` confirmed present in the same sweep.)

4. **Todo state vocabulary + store**:

```
$ grep -n "backlog.db" internal/cli/todo.go | head -1
197:		Long: `Operate the kanban backlog queue at ~/.moai/db/<project-key>/todo/backlog.db.
```

Current states measured: picked / queued / done (`hold` absent) — the pickup predicate is written in positive state vocabulary (REQ-MT-010) so the future `hold` state is skipped without SPEC revision.

5. **Codex read-only role roster** (conflict-report source):

```
$ grep -n "mission-governor" internal/cli/codex_audit_mcp.go | head -1
199:			mcp.WithDescription("Start a role whose permission contract is read-only (plan-auditor, sync-auditor, mission-governor, super-advisor) ...
$ grep -n "mission-governor: read-only" internal/template/agentemit/agents-codex.yaml
83:      mission-governor: read-only
489:    mission-governor: read-only
```

6. **internal/web measured zero**:

```
$ git grep -n -i "mission.governor" -- internal/web/
(no output — zero hits)
```

### Phase-1 SKIP rationale

Cross-reference: plan.md §I DP1. The plan-phase discovery round (Socratic/clarify, workflow phase 1) is satisfied by the approval proxy — the operator-approved card plus the lead's explicit plan-commencement directive dated 2026-09-29 — recorded verbatim in plan.md §I; no AskUserQuestion channel exists inside a lane, so no clarification round was run. The one open decision (DP2, roster disposition) was escalated, not guessed — it was carried as a clarification marker through iteration 1 and is now RESOLVED by the lead (2026-09-29, rename-registered, zero removals; plan.md §I DP2, research.md §C, design D-3). Zero `[NEEDS CLARIFICATION]` markers remain in the artifact set (measured post-repair: grep returns nothing).

### Decision Point 1 disposition

DP1 (plan-phase entry): approval proxy = operator-approved card t1306 + lead's explicit "plan 착수" directive dated 2026-09-29. Recorded as such; run-phase entry still requires the standing Implementation Kickoff gate.

### Audit-ready signal

_<withheld — the lane appends the audit-ready signal lines after iteration 2 passes; the iter-1 draft carried a provisional signal that was removed at repair time (iter1 FAIL 0.80, so `plan_status: audit-ready` does not hold)_

### Plan-phase artifact set (Tier L)

- `spec.md` — 12-field frontmatter, 18 REQs (GEARS), Out of Scope (5 H3 topics)
- `plan.md` — milestones M1-M5, pre-flight, self-verification, decision points DP1-DP3 (DP2 resolved at repair round)

### Iteration 1 verdict + repair round (2026-09-29)

**Iter-1 verdict**: FAIL 0.80 (Tier L threshold 0.85) + 3 blocking defects; report `.moai/reports/t1306/plan-audit-iter1.md` (measured against `e11110c60`). Repair round authorized by the lane.

**DP2 resolution applied** (lead decision 2026-09-29, option a — rename-registered, never removed): plan.md §I DP2 carries the verbatim decision + the four surfaces (`codex_audit_mcp.go:199`, `agents-codex.yaml:83,489`, catalogue rule row, fingerprint prose); plan.md §B item 1 rewritten; M2 step 3 gate reads resolved; design D-3 rewritten (removal disposition superseded); research.md §C + §G updated; acceptance.md DoD item 3 updated; spec.md REQ-MT-006/017 reworded to zero-removals. Zero `[NEEDS CLARIFICATION]` markers remain:

```
$ grep -rn "NEEDS CLARIFICATION" .moai/specs/SPEC-MANAGER-TODO-001/
(no output — exit 1)
```

**Defect repairs**:

- D1 (blocking): REQ-MT-005 + AC-MT-005 now carry the third allowed disposition **RENAME-at-next-regen** with a mechanical codemaps arm (count equals baseline 8, disposition literal present in the §B table, D-6 conditional-flip arm); research.md §G's "either is consistent" claim corrected.
- D2 (blocking): REQ-MT-003's sweep-refresh clause removed (property owned by REQ-MT-005/AC-MT-005 at M2 closure); AC-MT-003 relabeled to verify the deletion clause it claims.
- D3 (blocking): design D-11 added — processing means = `moai-kanban-foreman` pattern (pick → isolated in-session Agent() worker → completion judged on disk evidence read → done; failure path unpicks to `queued` with labelled non-finding; explicit no-factory-run/lease/slot boundary); REQ-MT-007 reworded onto that contract; AC-MT-009 extended with the evidence-absent failure arm and the entry-point arm (D6).
- D4: REQ-MT-009 active prohibition form; REQ-MT-012 `Where`→`When`; REQ-MT-014 capability-gate phrasing.
- D5: AC-MT-001 body checks made mechanical (keyword counts + sub-role heading grep); AC-MT-014 pins the D-8 canonical string; AC-MT-015 converted to counted grep (0 hits, exit 1) + lease-dir snapshot; AC-MT-017 converted to call-site consumption grep + poisoned-value behavior test.
- D6: AC-MT-009 entry-point arm (alias routing check + `--auto` in help output).
- D7: "95 files" corrected to 105 unique files with the aggregation command, re-measured this session before the edit (262 hits / 105 files) — research.md §B + progress.md above.
- D8: plan.md M4 verify row corrected to REQ-MT-014..017 / AC-MT-017..019.
- D10: plan.md §B rewritten (marker removed; acceptance.md citation restated accurately).
- D11: AC-MT-011 non-cache arm added (revived-owner re-measurement between consecutive decisions).

**Post-repair lint**:

```
$ ~/go/bin/moai spec lint SPEC-MANAGER-TODO-001
✓ No findings — all SPEC documents are valid
```

**Repair commit**: <see git log — this section is finalized at commit time; the lane reads the SHA from the commit message trail.>

- `acceptance.md` — 22 ACs (AC-MT-001..022), all mechanical; edge cases; quality gates; REQ→AC traceability map (complete, 18/18)
- `research.md` — measured current state, 262-hit baseline enumeration, contract-conflict report (resolved), t1240 seam
- `design.md` — decisions D-1..D-11

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
