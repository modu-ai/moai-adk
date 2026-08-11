# progress.md — SPEC-FACTORY-BOOTSTRAP-001

> Phase evidence and audit-ready signals. §E.1 is populated at plan-phase; §E.2–§E.4 are placeholder headings to be populated by manager-develop (run-phase) and manager-docs (sync-phase) per the Forbidden-modifications matrix.

---

## §A. Plan-Phase Artifact Set

| Artifact | Path | Status |
|---|---|---|
| spec.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/spec.md` | emitted |
| plan.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/plan.md` | emitted |
| acceptance.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/acceptance.md` | emitted |
| design.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/design.md` | emitted (Tier L) |
| research.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/research.md` | emitted (Tier L) |
| progress.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/progress.md` | emitted (this file) |

---

## §B. Branch State (at plan-phase emission)

- **Branch**: `feat/factory-bootstrap-guidance`
- **HEAD**: `94025ce0a` (prior-art commit "feat(factory): announce companion session bootstrap from the SessionStart hook")
- **Base**: `chore/revert-kanban-rename` (`24c4674b5`) ← `origin/main`
- **Local ahead by**: 2 (no race)
- **Tree status at emission**: clean working tree, plan-phase artifacts are the only new files

---

## §C. Requirements Budget

- **REQ count**: 18 (REQ-FB-001..018) against the Tier L ceiling of 25.
- **AC count**: 27 against 18 REQ (9-criterion surplus reflects the multi-surface span — env / source / notice / help / docs-site / template-neutrality / sibling-boundary / worktree-isolation / breaking-change-pin). AC-FB-016a is a fail-open sub-clause paired with AC-FB-016, not a separate criterion.
- **MUST criteria**: 25; **SHOULD**: 1 (AC-FB-020); **meta**: 1 (AC-FB-026 worktree isolation).

---

## §D. Pre-Plan-Audit Self-Check

- [x] SPEC ID regex: `SPEC-FACTORY-BOOTSTRAP-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (Bash output: `PASS`).
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + tier L.
- [x] ID uniqueness: no prior `SPEC-FACTORY-BOOTSTRAP-001` directory exists at `.moai/specs/`.
- [x] Requirements in GEARS notation (Where/When/While + shall).
- [x] Out of Scope: §C carries ≥ 1 `### Out of Scope — <topic>` H3 heading with `-` bullets (satisfies `OutOfScopeRule`).
- [x] Artifact set matches Tier L (spec + plan + acceptance + design + research + progress).
- [x] spec.md carries no implementation detail (no function signatures, no API schemas — only observable behaviors, env-var names, file paths as evidence anchors).
- [x] Prior-art commit `94025ce0a` recorded in HISTORY and §A.1 as revised-not-reverted.
- [x] Sibling boundary one-sided (no edits to `.moai/specs/SPEC-KANBAN-*`).
- [x] AC003 preserve-tests named in plan.md §A.5 with package path.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_version: 0.2.0
spec_id: SPEC-FACTORY-BOOTSTRAP-001
tier: L
target_auditor_threshold: 0.85
requirements_count: 18
requirements_ceiling: 25
acceptance_criteria_count: 27
must_criteria: 25
should_criteria: 1
meta_criteria: 1
prior_art_commit: 94025ce0a
prior_art_relation: revised-not-reverted
predecessor_spec: SPEC-FACTORY-MODE-001
predecessor_status: completed
sibling_spec: SPEC-KANBAN-BOOTSTRAP-001
sibling_boundary: one-sided-from-this-side
out_of_scope_sections:
  - "Out of Scope — Topology-config-gated quorum and dispatch"
  - "Out of Scope — Forward reference to the sibling's supersedence"
  - "Out of Scope — Run-phase decisions"
  - "Out of Scope — Single-session factory mode"
preserve_tests:
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal"
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_BlockCapDoctrineClauseSpecific"
baseline_attribution:
  worktree: "/Users/goos/.moai/worktrees/kanban"
  branch: "feat/factory-bootstrap-guidance"
  head: "94025ce0a"
  base: "24c4674b5"
  tree_status: "clean"
audit_revision:
  iteration: 2
  prior_verdict: FAIL
  prior_score: 0.81
  findings_closed: [D1, D2, D3, D4, D5]
  breaking_change_94025ce0a: "companion-shape --name alone reclassified from companion entry to no-op (spec.md §A.2.1, REQ-FB-001 no--f clause, AC-FB-027)"
frontmatter_repo_conventions:
  - "related_specs: [SPEC-FACTORY-MODE-001, SPEC-KANBAN-BOOTSTRAP-001] is a repo convention NOT codified in the canonical 12-field or optional schema; spec-lint passes (moai spec lint → 0 findings) and the sibling SPEC-KANBAN-BOOTSTRAP-001 carries the same field. Schema codification is a separate SPEC."
open_clarifications: 0
blocker_report: none
```

---

## §E.2 Run-phase Evidence

| AC | Status | Milestone | Evidence |
|---|---|---|---|
| AC-FB-010 | PASS | M1 | `go test ./internal/cli/ -run TestPrepareFactorySettingsWritesTransientFile` → ok; writes `moai-factory-<pid>-<ns>.json` under `os.TempDir()` containing `{"crossSessionInbound":"accept"}` |
| AC-FB-011 | PASS | M1 | `go test ./internal/cli/ -run TestPrepareFactorySettingsHonorsOperatorSupplied` → ok; operator `--settings` suppresses injection in long, equals, and pre-marker forms |
| AC-FB-012 | PASS | M1 | `go test ./internal/hook/ -run TestFactoryLeadNoticeOperatorSettingsAdvisory` → ok; notice contains `verify`/`crossSessionInbound`/`accept` when `MOAI_FACTORY_SETTINGS_INJECTED` unset |
| AC-FB-013(d) | PASS | M1 | `go test ./internal/hook/ -run TestFactoryLeadNoticeInjectedSettingsAutoAccept` → ok; notice contains `auto-accept` when `MOAI_FACTORY_SETTINGS_INJECTED=1` |
| (others) | pending | M2-M6 | — |

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_

---

## §F Phase 4 Mode Selection

### Input parameters

- **tier**: L (18 REQ / 27 AC, >15 files affected)
- **scope (file count)**: ~14-15 — Go source (`internal/cli/factory.go`, `cc.go`, `glm.go`, `internal/hook/session_start_factory.go`, new settings-injection helper; `launcher_blockcap_infinite.go` read-only) + Go tests (`factory_bootstrap_test.go`, `session_start_factory_test.go`, `launcher_blockcap_infinite_test.go` extended, +companion-combo test) + docs-site 4-locale `factory-mode.md` + `data/menu/main.yaml`
- **domain count**: 3 (`internal/cli` dispatch + env, `internal/hook` SessionStart notice, docs-site 4-locale publishing)
- **file language mix**: Go source (~6) + Go tests (~4) + markdown docs-site (4) + YAML (1) — mixed, Go-majority
- **concurrency benefit**: LOW — the milestones repeatedly edit the SAME Go files (`factory.go`, `cc.go`, `glm.go` are touched by both M2 and M4; `session_start_factory.go` by M3), so parallel fan-out would race on shared write targets
- **Agent Teams prereqs**: n/a (Mode 3 retired)

### Mode evaluation table

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Tier L, 6 milestones, multi-file semantic change |
| 2 background | no | not a single async read-only task; multi-milestone write work |
| 3 agent-team | no | RETIRED (tombstone) |
| 4 parallel | no | coding-heavy + shared-file milestones → Anthropic coding-task parallelism caveat; concurrent edits to `cc.go`/`glm.go`/`factory.go` would race |
| 5 sub-agent | **yes** | sequential manager-develop delegation; plan.md provides per-milestone implementation guidance; same-Go-file recurrence demands serialization |
| 6 workflow | no | not a ≥30-file single-uniform-transform; milestones carry distinct semantic changes (dispatch rewrite, notice revision, settings injection, help text, docs pages) |

### Decision

`sub-agent`

### Justification

Mode 5 (sequential sub-agent) is selected per Anthropic's coding-task parallelism caveat: most coding tasks involve fewer truly parallelizable tasks than research, and the six milestones here repeatedly modify the same Go source files (`internal/cli/factory.go`, `cc.go`, `glm.go` by both M2 and M4; `internal/hook/session_start_factory.go` by M3). Parallel fan-out (Mode 4) would race on those shared write targets. plan.md §B already orders the milestones by decision-reversibility (M0 decision → M1 new surface → M2 core semantic → M3 notice → M4 help → M5 docs → M6 verify) and provides concrete implementation guidance per milestone, so a single sequential manager-develop delegation with the Tier L Section A-E template carries the full scope. Implementation Kickoff Approval was obtained in this session; plan-audit review-2 verdict is PASS 1.00 (Tier L threshold 0.85), D1-D5 closed, D6 non-blocking.
