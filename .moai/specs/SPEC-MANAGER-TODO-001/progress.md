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

Live-surface baseline: 262 hits / 95 files, enumerated per file:line in research.md §B. The 111-hit remainder under `.moai/reports/**` + `.moai/specs/**` is historical and dispositioned out of scope.

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

Cross-reference: plan.md §I DP1. The plan-phase discovery round (Socratic/clarify, workflow phase 1) is satisfied by the approval proxy — the operator-approved card plus the lead's explicit plan-commencement directive dated 2026-09-29 — recorded verbatim in plan.md §I; no AskUserQuestion channel exists inside a lane, so no clarification round was run. The one genuinely open decision is not guessed: it is carried as `[NEEDS CLARIFICATION: codex read-only role roster disposition]` in plan.md §B/§I and research.md §C (never in spec.md/acceptance.md), and gates run-phase M2 step 3.

### Decision Point 1 disposition

DP1 (plan-phase entry): approval proxy = operator-approved card t1306 + lead's explicit "plan 착수" directive dated 2026-09-29. Recorded as such; run-phase entry still requires the standing Implementation Kickoff gate.

### Audit-ready signal

```yaml
plan_complete_at: 2026-09-29T00:00:00Z
plan_status: audit-ready
```

(`plan_complete_at` timestamp is set to the plan-phase authoring date; the exact commit SHA lands with the artifact commit — the sync-phase audit reads the SPEC directory's git history.)

### Plan-phase artifact set (Tier L)

- `spec.md` — 12-field frontmatter, 18 REQs (GEARS), Out of Scope (5 H3 topics)
- `plan.md` — milestones M1-M5, pre-flight, self-verification, decision points DP1-DP3, 1 clarification marker
- `acceptance.md` — 22 ACs (AC-MT-001..022), all mechanical; edge cases; quality gates; REQ→AC traceability map (complete, 18/18)
- `research.md` — measured current state, 262-hit baseline enumeration, contract-conflict report, t1240 seam
- `design.md` — decisions D-1..D-10

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
