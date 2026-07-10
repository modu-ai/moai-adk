# Progress — SPEC-HARNESS-MCP-PROVISION-001

> Lifecycle progress ledger. §E is parser-load-bearing (era.go string-matches the
> literal `§E.2` / `§E.3` / `§E.4` heading tokens + `sync_commit_sha`). Do NOT rename
> the §E.N headings. Plan-phase populates §E.1 only; §E.2-§E.4 are placeholder
> headings owned by manager-develop (run) and manager-docs (sync).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-11
plan_version: 0.1.1
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 11
ac_count: 15
depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]
open_clarifications: []
resolved_clarifications:
  - marker: "mcp-matrix config surface"
    decision: "standalone template DATA RESOURCE (internal/template/templates/.moai/config/sections/mcp-matrix.yaml, read as prose-context by project/doc-generation.md) — NOT a new Go config section; no typed loader / struct field. Keeps the SPEC doc/config-only."
  - marker: "doctor manifest-mcp validate-vs-tolerate"
    decision: "TOLERATE-ONLY, zero Go change. Verified: doctor.go + applier.go use plain json.Unmarshal with no DisallowUnknownFields; v4manifest/validate.go checks only required fields. AC-HMP-010 encodes documented-tolerance grep + regression guard grep -c DisallowUnknownFields internal/harness/v4manifest/*.go == 0. Active mcp-block validation deferred to a follow-up Go SPEC."
notes: >
  SPEC 2 of the 3-SPEC Project-Harness Pipeline Epic. Doc/config-only (markdown+yaml);
  no Go code. Phase 0.5 Depends_on pre-flight WILL block run-phase entry until
  SPEC-PROJECT-HARNESS-BRIDGE-001 reaches status: completed (currently draft) — expected.
  Plan-audit fixes applied at v0.1.1: both clarifications resolved (markers struck from
  plan.md); AC-HMP-014 added for REQ-HMP-003 (write-on-approval coverage gap);
  AC-HMP-010 de-vacuumed (grep+guard, dropped repo-wide doctor smoke); Epic artifact
  numbering corrected (MCP fragment = artifact 7 OPTIONAL; verify skill = artifact 6
  mandatory, SPEC-HARNESS-VERIFY-PROMOTE-001); AC-HMP-015 added for the harness-builder
  "exactly 5" prose reconciliation.
```

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs; carries sync_commit_sha>_
