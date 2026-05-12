# acceptance.md — SPEC-V3R4-STATUS-LIFECYCLE-001

Hierarchical Acceptance Criteria per SPEC-V3R2-SPC-001 schema (`.claude/rules/moai/workflow/spec-workflow.md` § Hierarchical Acceptance Criteria Schema). Depth >= 2; each leaf maps to a REQ.

---

## §1. Acceptance Structure

Three top-level AC nodes, one per Wave. Each node has 2-3 depth-1 children with depth-2 grandchildren where multi-variant testing is required.

```
AC-STATUS-LIFECYCLE-001-01  (Wave 1 — Policy + Lint)
AC-STATUS-LIFECYCLE-001-02  (Wave 2 — Hook + Transitions)
AC-STATUS-LIFECYCLE-001-03  (Wave 3 — Automation + Visibility)
```

---

## §2. Hierarchical Acceptance Criteria

### AC-STATUS-LIFECYCLE-001-01: Given Wave 1 (Policy + Lint) is merged

Given the canonical 8-value enum is documented and lint rules are active:

- **AC-STATUS-LIFECYCLE-001-01.a**: Given a SPEC file `.moai/specs/SPEC-FAKE-001/spec.md` with frontmatter `status: Planned` (capitalized)
  - **AC-STATUS-LIFECYCLE-001-01.a.i**: When `moai spec lint --strict` runs on the file, Then the command exits with code 1 AND the output contains a finding with code `StatusCaseInvalid` AND the message recommends `planned` (lowercase). (maps REQ-3.2)
  - **AC-STATUS-LIFECYCLE-001-01.a.ii**: When `moai spec lint` runs (non-strict), Then the command exits with code 0 AND the output contains the same `StatusCaseInvalid` finding at warning severity. (maps REQ-3.2)

- **AC-STATUS-LIFECYCLE-001-01.b**: Given the 21 SPECs in REQ-6 backfill list have been converted to YAML frontmatter
  - **AC-STATUS-LIFECYCLE-001-01.b.i**: When `moai spec lint --strict` runs on `.moai/specs/`, Then exit code is 0 AND zero findings of code `StatusValueInvalid`, `StatusCaseInvalid`, or `FrontmatterInvalid` are reported. (maps REQ-6)
  - **AC-STATUS-LIFECYCLE-001-01.b.ii**: When the diff is inspected, Then no SPEC has lost its title, requirements section, or any content outside frontmatter. (maps REQ-6, C-2)
  - **AC-STATUS-LIFECYCLE-001-01.b.iii**: When SPEC-MX-001 is examined, Then its frontmatter contains `status: completed` (downgraded from the legacy `Planned`, since main carries implementation) OR `status: archived` (if user judged it archive-worthy during W1-T3 review). (maps REQ-6, R-3 mitigation)

- **AC-STATUS-LIFECYCLE-001-01.c**: Given `.github/workflows/spec-lint.yml` is committed
  - **AC-STATUS-LIFECYCLE-001-01.c.i**: When a PR is opened that introduces a SPEC with `status: Foobar` (invalid enum value), Then the `spec-lint` workflow run completes with a failed status. (maps REQ-4, REQ-3.1)
  - **AC-STATUS-LIFECYCLE-001-01.c.ii**: When a PR is opened with only valid SPEC frontmatter changes, Then the `spec-lint` workflow completes green. (maps REQ-4)

- **AC-STATUS-LIFECYCLE-001-01.d**: Given `.claude/rules/moai/workflow/spec-workflow.md` § Status Lifecycle section is committed
  - **AC-STATUS-LIFECYCLE-001-01.d.i**: When the file is read, Then exactly 8 values are listed (`draft, planned, in-progress, implemented, completed, superseded, archived, rejected`) AND the hyphen form `in-progress` is used (not `in_progress`). (maps REQ-1, C-1)

---

### AC-STATUS-LIFECYCLE-001-02: Given Wave 2 (Hook + Transitions) is merged

Given the PR-merge transition table is implemented and the hook is extended:

- **AC-STATUS-LIFECYCLE-001-02.a**: Given `internal/spec/transitions.go` defines `ClassifyPRTitle(title) -> (category, targetStatus, error)`
  - **AC-STATUS-LIFECYCLE-001-02.a.i**: When called with `"plan(spec): SPEC-FOO-001 — initial draft"`, Then returns `(plan-merge, "planned", nil)`. (maps REQ-2)
  - **AC-STATUS-LIFECYCLE-001-02.a.ii**: When called with `"feat(SPEC-FOO-001): implement REQ-1"`, Then returns `(run-partial OR run-complete, "in-progress" OR "implemented", nil)`. The complete/partial distinction is resolved by querying AC completion status in the SPEC file. (maps REQ-2)
  - **AC-STATUS-LIFECYCLE-001-02.a.iii**: When called with `"docs(sync): SPEC-FOO-001 status=completed"`, Then returns `(sync-merge, "completed", nil)`. (maps REQ-2)
  - **AC-STATUS-LIFECYCLE-001-02.a.iv**: When called with `"chore(spec): auto-sync status for #999"` (own follow-up), Then returns `(skip-meta, "", nil)` AND the auto-sync workflow recognizes the skip category. (maps REQ-2, REQ-5 loop prevention)
  - **AC-STATUS-LIFECYCLE-001-02.a.v**: When called with `"revert: feat(SPEC-FOO-001): ..."`, Then returns `(no-op, "", nil)` (revert does not auto-roll-back status). (maps REQ-2, R-1 mitigation)

- **AC-STATUS-LIFECYCLE-001-02.b**: Given `internal/hook/spec_status.go` is extended with `isGhPrMergeCommand` and transition table lookup
  - **AC-STATUS-LIFECYCLE-001-02.b.i**: When a PostToolUse event arrives with `command: "git commit -m 'feat(SPEC-FOO-001): impl'"`, Then the handler updates `SPEC-FOO-001` status to `implemented` (preserves existing REQ-3 of STATUS-AUTO-001). (maps REQ-2, regression)
  - **AC-STATUS-LIFECYCLE-001-02.b.ii**: When a PostToolUse event arrives with `command: "gh pr merge 871 --squash --admin"` AND the PR title is `"plan(spec): SPEC-FOO-001 — initial"`, Then the handler updates `SPEC-FOO-001` status to `planned`. (maps REQ-2)
  - **AC-STATUS-LIFECYCLE-001-02.b.iii**: When the same `gh pr merge` event arrives a second time (idempotency check), Then the handler is a no-op (status already at target). No error logged, no spurious commit. (maps REQ-2, C-5)

- **AC-STATUS-LIFECYCLE-001-02.c**: Given `.claude/skills/moai/workflows/plan.md` and `sync.md` are updated
  - **AC-STATUS-LIFECYCLE-001-02.c.i**: When `plan.md:454` is read, Then the enum statement matches the canonical 8 values from W1-T1 (`draft | planned | in-progress | implemented | completed | superseded | archived | rejected`). (maps REQ-1)
  - **AC-STATUS-LIFECYCLE-001-02.c.ii**: When `sync.md` is read, Then no line uses an enum value outside the canonical 8 (in particular, no `approved` appears). (maps REQ-1)
  - **AC-STATUS-LIFECYCLE-001-02.c.iii**: When `make build` is run, Then `internal/template/embedded.go` is regenerated with the updated content AND `go test ./internal/template/...` passes. (maps REQ-1, template-first discipline)

---

### AC-STATUS-LIFECYCLE-001-03: Given Wave 3 (Automation + Visibility) is merged

Given CI auto-sync, drift CLI, and SessionStart integration are deployed:

- **AC-STATUS-LIFECYCLE-001-03.a**: Given `.github/workflows/spec-status-auto-sync.yml` is committed
  - **AC-STATUS-LIFECYCLE-001-03.a.i**: When a PR with title `"plan(spec): SPEC-FAKE-001 — initial"` is merged via squash, Then within 90 seconds a follow-up commit `chore(spec): auto-sync status for #<PR>` appears on main AND `SPEC-FAKE-001`'s frontmatter has `status: planned`. (maps REQ-5)
  - **AC-STATUS-LIFECYCLE-001-03.a.ii**: When the follow-up commit from a.i is merged (it shouldn't be a PR, it's a direct commit), Then the workflow does NOT re-trigger an auto-sync (loop prevention via title skip). (maps REQ-5)
  - **AC-STATUS-LIFECYCLE-001-03.a.iii**: When a PR with title `"chore(deps): bump go 1.26"` (no SPEC-XXX) is merged, Then the workflow runs but produces no commit (zero SPEC-IDs extracted). (maps REQ-5)

- **AC-STATUS-LIFECYCLE-001-03.b**: Given `moai spec drift` subcommand is registered
  - **AC-STATUS-LIFECYCLE-001-03.b.i**: When `moai spec drift` is invoked on a repo with zero drift, Then exit code is 0 AND the output contains "0 drift, N aligned". (maps REQ-7)
  - **AC-STATUS-LIFECYCLE-001-03.b.ii**: When `moai spec drift --exit-code-on-drift` is invoked on a repo with at least one drift, Then exit code is 1 AND each drift line names the SPEC-ID + drift category. (maps REQ-7)
  - **AC-STATUS-LIFECYCLE-001-03.b.iii**: When `moai spec drift --json` is invoked, Then stdout is valid JSON parseable as `{aligned: [], drift: [{spec_id, category, evidence_commit}], unknown: []}`. (maps REQ-7)
  - **AC-STATUS-LIFECYCLE-001-03.b.iv**: When the command runs on the current repo (180+ SPECs), Then execution completes within 2 seconds. (maps REQ-7, A-5)

- **AC-STATUS-LIFECYCLE-001-03.c**: Given the SessionStart hook integrates `moai spec drift --count`
  - **AC-STATUS-LIFECYCLE-001-03.c.i**: When a Claude Code session starts on a repo with drift count >= 5, Then stderr emits exactly one line: `[drift] N SPECs out of sync — run 'moai spec drift' for details`. (maps REQ-8)
  - **AC-STATUS-LIFECYCLE-001-03.c.ii**: When a session starts on a repo with drift count < 5, Then stderr emits no drift-related output. (maps REQ-8)
  - **AC-STATUS-LIFECYCLE-001-03.c.iii**: When the SessionStart hook is invoked, Then the drift check completes within 500ms (sub-second to avoid session-start latency). (maps REQ-8)

- **AC-STATUS-LIFECYCLE-001-03.d**: Given `StatusGitConsistencyRule` is registered in `internal/spec/lint.go`
  - **AC-STATUS-LIFECYCLE-001-03.d.i**: When the rule runs on a SPEC with `status: draft` but `git log main` contains a `plan(SPEC-X)` commit, Then a finding with code `StatusGitDrift` is emitted at warning severity. (maps REQ-3.3)
  - **AC-STATUS-LIFECYCLE-001-03.d.ii**: When the same rule runs under `--strict`, Then the same finding is elevated to error severity AND `moai spec lint --strict` exits with code 1. (maps REQ-3.3)

- **AC-STATUS-LIFECYCLE-001-03.e**: Given `manager-spec.md`, `manager-docs.md`, `manager-develop.md` are updated
  - **AC-STATUS-LIFECYCLE-001-03.e.i**: When `manager-spec.md` is read, Then the body contains a "Responsibility Matrix" section listing `draft (initial)` and `draft → planned (after plan PR merge)`. (maps REQ-9)
  - **AC-STATUS-LIFECYCLE-001-03.e.ii**: When `manager-develop.md` is read, Then it lists `planned → in-progress → implemented`. (maps REQ-9)
  - **AC-STATUS-LIFECYCLE-001-03.e.iii**: When `manager-docs.md` is read, Then it lists `implemented → completed`. (maps REQ-9)

---

## §3. Edge Cases

| ID | Scenario | Expected Behaviour |
|----|----------|---------------------|
| E-1 | PR title contains TWO SPEC-IDs (`feat(SPEC-A-001, SPEC-B-002):`) | Both SPECs receive the status update independently. |
| E-2 | PR title has SPEC-ID but the SPEC directory does not exist | Auto-sync logs a warning, skips, exit code 0. |
| E-3 | SPEC has `status: archived` and a new `plan(SPEC-X)` PR is merged | Status is updated to `planned` (overwrites archived). Open question: should this be blocked? **Decision**: No block, but `StatusGitConsistencyRule` emits a warning on the next lint run. |
| E-4 | Auto-sync workflow runs on a fork's PR | Workflow respects `GITHUB_TOKEN` permissions; if fork has no write access, workflow exits gracefully without committing. |
| E-5 | Backfill (W1-T3) encounters a SPEC where the legacy status is already valid (lowercase YAML) | No-op for that file; not counted as a backfilled SPEC. |
| E-6 | Concurrent PR merges trigger two auto-sync workflow runs | Each run commits independently; second commit may need rebase. REQ-5 retry-with-rebase covers this. |
| E-7 | Drift CLI encounters a SPEC with no spec.md (only plan.md / acceptance.md) | Classified as `unknown`; not counted as aligned or drift. |

---

## §4. Quality Gate Criteria

All AC nodes (AC-STATUS-LIFECYCLE-001-01, -02, -03) must be GREEN before the SPEC moves to `completed`. Wave-level gates:

- **Wave 1 gate**: AC-...-01.* all GREEN.
- **Wave 2 gate**: AC-...-02.* all GREEN AND Wave 1 still GREEN (regression check).
- **Wave 3 gate**: AC-...-03.* all GREEN AND Waves 1 and 2 still GREEN.

LSP quality gates per project standard:
- Wave 1: zero errors, zero type errors, zero lint errors (Go side) + new lint rules pass on all 180+ SPECs.
- Wave 2: same + new unit tests >= 90% coverage on `internal/spec/transitions.go` and extended hook paths.
- Wave 3: same + smoke test for both new workflows pass on a test PR.

---

## §5. Definition of Done

This SPEC is `completed` when:

1. All three Wave gates are GREEN (see §4).
2. Sync PR for this SPEC is merged and frontmatter reflects `status: completed`.
3. 30-day monitoring period begins (per spec.md §7 Verification).
4. After 30 days, M-1 (zero retrofit PRs), M-2 (drift count <= 2), M-3 (lint --strict clean) are evaluated and recorded in a follow-up `docs(monitoring):` PR.
5. If M-1 fails, a follow-up SPEC is opened to address the gap (this SPEC does NOT auto-revert).

The 30-day monitoring period is the **success contract** for this SPEC. Implementation completion is necessary but not sufficient; the system must demonstrate drift suppression in practice.

---

## §6. Out of Acceptance Scope

Not acceptance-tested (deferred):

- Long-term drift trends beyond 30 days.
- Adoption of the standard by downstream projects using moai-adk as a template.
- Performance under repos with >1000 SPECs (current repo has ~190; scaling deferred).
- Web UI for status visualization.
