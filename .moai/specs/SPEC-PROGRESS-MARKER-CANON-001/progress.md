# SPEC-PROGRESS-MARKER-CANON-001 — Progress

> This progress.md dogfoods the canonical §E section skeleton (convention B) that
> SPEC-PROGRESS-MARKER-CANON-001 instructs manager-spec to emit at plan-phase.
> The `§E.2`-`§E.5` headings are what make `internal/spec/era.go` `ClassifyEra` see a
> recognized `§E.*` marker and skip H-2 (V3R2-R4) misclassification — its
> `hasAnyProgressMarker` (era.go L165-169) greps for `§E.2`/`§E.3`/`§E.4`/`§E.5`, NOT
> `§E.1`. The `§E.1` heading is included for human/audit readability, not for H-2-avoidance.
> §E.2-§E.5 *content* is populated by the downstream owners (manager-develop / manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-06-16
plan_commit_sha: <pending — set on plan-phase artifact commit>
plan_status: draft
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
tier: M
era: V3R6
plan_auditor_verdict: <pending — plan audit gate>
```

## §E.2 Run-phase Evidence

> Owner: manager-develop. Populated during run-phase with the AC PASS/FAIL matrix
> (Actual Output + Status columns per AC row).

### AC Matrix (run-phase)

| AC ID | Verification (summary) | Actual Output | Status |
|-------|------------------------|---------------|--------|
| AC-PMC-001 | conv-A Sync heading absent / conv-B §E.4 heading present | conv-A `0`; §E.4 `2` | PASS |
| AC-PMC-002 | region §E.2 + §E.5 survive (H-4 non-regression) | §E.2 `5`; §E.5 `3` | PASS |
| AC-PMC-003 | heuristic table + JSON excerpt unchanged | (empty — no `+` line touches table/JSON) | PASS |
| AC-PMC-004 | `run-evidence start marker` count in era.go (3 comment sites) | `4` (≥3) | PASS |
| AC-PMC-004b | region §E.4 Sync + §E.2 Run headings co-exist, §E.2 Sync absent | §E.4 `1`; §E.2 Run `1`; §E.2 Sync `0` | PASS |
| AC-PMC-005 | era.go comment-only diff (zero non-comment added lines) | (empty) | PASS |
| AC-PMC-006 | `go test ./internal/spec/...` + `go build ./...` | `ok …/internal/spec 6.198s`; build exit `0` | PASS |
| AC-PMC-007 | all 5 §E headings present in manager-spec.md | §E.1 `1`, §E.2 `2`, §E.3 `2`, §E.4 `2`, §E.5 `2` | PASS |
| AC-PMC-008 | skeleton instruction reference present | `3` hits | PASS |
| AC-PMC-009 | Forbidden-modifications boundary preserved | matched (`does NOT authorize`, `belongs to manager-develop/docs`) | PASS |
| AC-PMC-010 | mirror parity + build regen | `PARITY OK`; make build exit `0`, catalog.yaml regen (manager-spec hash `9b154a14…`) | PASS |
| AC-PMC-011 | lifecycle-sync-gate.md no template mirror | `NO MIRROR (correct)` | PASS |
| AC-PMC-012 | era.go no template mirror | `NO MIRROR (correct)` | PASS |

> AC-PMC-010 embedded.go note: this project uses `//go:embed all:templates` compile-time embedding
> (no generated `embedded.go` file in the tree). The build-system artifact invalidated by the
> manager-spec.md edit is `internal/template/catalog.yaml` (catalog hash regen), which `make build`
> regenerated. The substance of AC-PMC-010 (build regenerated its embedded representation of the changed
> template) is satisfied via catalog.yaml.

### Quality gates (run-phase)

- spec-lint: `✓ No findings — all SPEC documents are valid` (exit 0)
- Template neutrality CI guard (`TestTemplateNeutralityAudit`): `ok` (exit 0)
- Full `internal/template` test suite (mirror parity + embed invariants): `ok` (exit 0)
- EC-2 grandfather scope: `git status --porcelain .moai/specs/` → only `SPEC-PROGRESS-MARKER-CANON-001/` touched
- EC-4 template neutrality: `grep -c 'PROGRESS-MARKER-CANON' <mirror>` → `0` (no SPEC ID embedded in agent body)

## §E.3 Run-phase Audit-Ready Signal

> Owner: manager-develop. Populated during run-phase.

```yaml
run_complete_at: 2026-06-16
run_commit_sha: 195b61e9b   # M3 (final run-phase milestone); M1 0ce8f1f77 / M2 41ff7209a
run_status: implemented
ac_pass_count: 13           # AC-PMC-001..012 + AC-PMC-004b, all PASS
ac_fail_count: 0
preserve_list_post_run_count: 0   # era.go grep logic + heuristic table + JSON excerpt byte-identical
new_warnings_or_lints_introduced: false
total_run_phase_files: 5    # era.go, lifecycle-sync-gate.md, manager-spec.md (+mirror), catalog.yaml (build regen)
m1_to_mN_commit_strategy: "M1 era.go comment-only (draft→in-progress) / M2 lifecycle-sync-gate worked-example / M3 manager-spec skeleton + mirror + catalog.yaml regen"
cross_platform_build:
  note: "documentation/comment-only SPEC; no syscall or platform-specific code introduced. go build ./... exit 0."
zero_new_production_code: true   # CON-PMC-002: only era.go comment text changed; no new functions/lint-rules/tests
```

## §E.4 Sync-phase Audit-Ready Signal

> Owner: manager-docs. Placeholder at plan-phase — populated during sync-phase
> (sync_complete_at, sync_commit_sha, sync_status, changelog_entry_position, frontmatter_status_transitions, etc.).

_<pending sync-phase>_

## §E.5 Mx-phase Audit-Ready Signal

> Owner: manager-docs / orchestrator. Placeholder at plan-phase — populated at Mx-phase
> close (mx_commit_sha, 4-phase close signal).

_<pending Mx-phase>_
