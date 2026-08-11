# progress.md — SPEC-NAVIGATOR-SYNC-004 (BAS Epic M2 — Falconer Route)

> Plan-phase skeleton. Populated by manager-spec (§E.1) at creation; §E.2-§E.4 remain placeholder
> headings until run-phase / sync-phase owners populate them. The literal §E.2/§E.3/§E.4 heading
> tokens are load-bearing for era classification (internal/spec/era.go hasAnyProgressMarker).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-11
- tier: M
- artifact_set: spec.md + plan.md + acceptance.md + progress.md
- req_count: 12 (REQ-NS4-001 … REQ-NS4-012)
- ac_count: 12 MUST ACs (AC-NS4-001a/b … AC-NS4-012)
- trigger_decision: on-demand CLI (navigator-route Hidden subcommand); no PostToolUse real-time path
- owner_binding: three resolution paths (orphan→implementation_path, missing→symbol-via-graph-else-doc, detect→changed_path); owner = code/doc path, never person
- success_metric: Route accuracy ≥ 70% (fixture corpus + go test ratio)
- depends_on: SPEC-NAVIGATOR-SYNC-001 (M0 graph — hard dependency)
- related_inputs: SPEC-NAVIGATOR-SYNC-002 (M1 detect — soft input, fail-open), SPEC-PROJECT-NAVIGATOR-002 (002 audit — primary input)

## §E.2 Run-phase Evidence

_(pending run-phase — manager-develop populates this section with M2.1-M2.5 commit SHAs, test output, coverage, and the §E attribution triple per SPEC-SYNC-PARALLEL-DOCS-001 A9)_

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — manager-develop populates this section when all MUST ACs show PASS with cited commands + observed output)_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-12
sync_commit_sha: pending-backfill-after-merge  # D3 self-referential-hazard workaround — backfilled in a follow-up commit after the sync PR merges
run_commit_sha: 73650aa44                       # feat(SPEC-NAVIGATOR-SYNC-004) run commit on worktree-bas-m2-route
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed"  # merged into the single sync commit (3-phase close)
  updated_field_refreshed: "2026-08-12"
changelog_entry_position: "[Unreleased] / Added"   # SPEC-NAVIGATOR-SYNC-004 entry appended
readme_decision: skip                              # M2 is a Hidden CLI + runtime Go; no distributed .claude/ template surface (AC-NS4-012 confirmed git diff template = 0)
docs_site_decision: skip                           # internal Navigator subsystem; navigator-route is Hidden, no user-facing command surface (consistent with NS1/NS2/NS3 predecessor precedent)
mx_tag_validation: sub-step-complete               # MX tag validation is a sync sub-step (no separate Mx-phase commit)
ac_pass_count_final: 15                            # AC-NS4-001a..012 — all PASS (15 ACs including sub-variants)
ac_fail_count_final: 0
route_coverage_pct: 85.8                           # internal/navigator/route/... coverage >= 85% target MET
route_accuracy_pct: 83.3                           # 25/30 actionable work items >= 70% target MET (happy path)
open_blockers: 0
```
