# progress.md — SPEC-LANE-STALL-WATCHDOG-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-30
plan_phase_tree: 3dd5adf2f (branch WT-lane-stall-watchdog)
note: artifact set complete (spec/plan/acceptance/progress, Tier M),
frontmatter schema-validated, SPEC-ID regex Bash check PASS, spec lint clean
at authoring commit; plan-auditor (--deep) verdict is the orchestrator's next
step and has not yet run.

## §A Overlap Adjudication (plan-phase record)

t1343 SPEC-RELATION-PICKUP-FILTER-001 (status: completed; landed — filter
code present in this tree at `internal/cli/todo_auto.go:162-187`) owns the
QUEUE-level pickup decision: relation-blocked `queued` cards are excluded
from `todo --auto` pickup while the finding lives in the record (REQ-RPF-001/
002/004; dead-owner rescue deliberately ungated per REQ-RPF-006). Card t1370
owns the LANE-level in-flight remedy: a dispatched lane stalled on a blocked
predecessor reads the predecessor's on-disk evidence directly and resumes on
landing evidence, else records an explicit wait. Composition, not
duplication: different lifecycle moment (pre-pickup vs in-flight), shared
data surfaces (relation findings + `.moai/reports/<card-id>/`), and the lane
remedy performs zero queue mutation — it never picks, drops, edits, or
unrelates. REQ-LSW-004 carries this boundary in the requirement text; AC-LSW-006
verifies the skill states it.

## §B Doc-vs-Code Surface Enumeration (plan-phase record)

| Kind | Path | Mirror |
|---|---|---|
| NEW rule | `.claude/rules/moai/workflow/auto-semantics.md` | template-first + `make build` |
| NEW skill | `.claude/skills/moai-lane-watchdog/SKILL.md` | template-first + `make build` |
| EDIT rule | `.claude/rules/moai/workflow/kanban-dispatch.md` | template mirror same change |
| EDIT local-only | `.claude/rules/local/gitflow-lane-protocol.md` §6 | NEVER mirrored |
| unchanged | Go (`internal/`, `pkg/`, `cmd/`), `.moai/config/**`, `.claude/loop.md` (+ mirror) | — |

`.claude/loop.md` is template-managed (mirror measured, 1254 bytes) and
deliberately untouched — the bare-`/loop` driver is the leader-side foreman
iteration; its queue-level 30-minute evidence deadline already watches worker
death. The lane-side awaken path is doctrine + skill (spec.md §B.3/§B.4).

## §C Three-Surface --auto Inventory (plan-phase record, code-verified)

- `todo --auto` — queue serial cycle; watchdog = `--auto-wait` default 30 min
  evidence deadline + unpick + dead-owner rescue (`internal/cli/todo.go:313`,
  `internal/cli/todo_auto.go:17-27, 142-161, 276-328`).
- lane (factory/kanban worker) — NO watchdog for the four stall classes
  (awaited-actor, blocked-by, shell-error, accidental stop); the open-ended
  wait posture is doctrine text (`gitflow-lane-protocol.md` §6). This is the
  gap the card closes.
- `goal --auto` — natural-language mission (draft → approve → run; blocked →
  resume), bounded by turn ceiling + stagnation guard + wall-clock; not
  infinite (`internal/cli/goal.go:148-154, 194`; `internal/goal/evaluate.go:325,
  337-347, 358-363`).
- Jev — display-only on every surface (`internal/cli/todo_auto.go:221-226,
  339-349`; `AGENTS.local.md` §29).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
