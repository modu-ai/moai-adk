# Progress — SPEC-DOCTOR-STAT-SEAM-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t563 · worktree `.claude/worktrees/t563` · branch `WT-doctor-stat-shim` · base `ef10a2524`
- Tier: S · Class C (scope decision: both stat sites, settled by card t563 — encoded in spec.md §3.1)
- Artifact set: spec.md + plan.md + acceptance.md (+ this progress.md); `tier: S` in frontmatter;
  acceptance.md included per the orchestrator's explicit deliverable list
- Requirements: 7 (REQ-001..REQ-007, GEARS) · Acceptance criteria: 8 (AC-SEAM-001..008)
- SPEC ID regex check executed as Bash: `PASS` (verbatim output cited in the plan-phase session)
- ID uniqueness: no `SPEC-DOCTOR-STAT-SEAM-001` in `.moai/specs/` (checked via directory listing)
- Measured facts re-confirmed on this tree: `os.Stat` at `doctor_codex.go:459,857`; `os.Lstat` at
  `:441`; `osStatFn` 0 occurrences in `doctor_codex.go`; seam at `update_preserve_inventory.go:59`
- Status: `draft` — awaiting plan audit and Implementation Kickoff Approval

### Repair iteration 1 (spec.md 0.2.0 · plan.md revised)

Resolves the four blocking findings of `.moai/reports/t563/plan-audit.md` (PASS-WITH-CONDITIONS
0.94): D1 — site-B coverage narrative corrected in spec.md §1 and plan.md §B (indirect
`TestCodexSkillPath_*` / `TestCheckCodexWiring_*` suite acknowledged; the real gap re-scoped to
stat-argument observability + portable stat-failure injection; M1 reframed as struct-level direct
pinning, folding existing fixtures into the M2 identity baseline); D2 — plan.md §C.1 preflight
rewritten as an ancestry assertion (`merge-base --is-ancestor`) after HEAD advanced past the
recorded base; D3 — REQ-005 / M1 no-injection claim scoped to the STAT seam (the `userHomeDirFn`
home seam may be overridden to reach the unresolvable-home arm); D4 — M1 exit selector re-pinned
to the named `TestCodexStaleSkillFinding_|TestInspectSkillMirror_` convention with a required
swept-test count. D5 (optional) — t540 pending-ID note added to spec.md §5. D6 — recorded, no
action.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
