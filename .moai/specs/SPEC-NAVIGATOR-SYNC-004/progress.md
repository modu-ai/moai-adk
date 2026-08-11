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

_(pending sync-phase — manager-docs populates sync_commit_sha on the single sync commit carrying the implemented → completed transition)_
