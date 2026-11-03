# SPEC-AGENTIC-FACTORY-IDENTITY-001 — progress

status: draft

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: revised-unaudited
plan_iteration: 2
plan_iteration_kind: delta (child iteration-1 defects D1-D11 + leader-relayed canon alignment)
spec_version: "0.2.1"
plan_complete_at: 2026-10-29
tier: M (+ design.md)
req_count: 7
ac_count: 9
id_set_changed: false
meanings_changed: true (reason codes, grant authority, per-Issue role exclusion, milestone order)
tree_head: 89f4e812a3395fd10cbd2890dc05b250dfaf4809 (0.2.1 edits uncommitted on top; 0.2.0 was measured at 3275706cc9)
parent_spec: SPEC-AGENTIC-FACTORY-001 (0.6.1)
completion_requires: "parent OD-4 operator risk acceptance (OQ-AFI-1); lock lift is parent M7 (lane-18), not this SPEC"
note: "revised, unaudited — no audit verdict exists for 0.2.0 or 0.2.1; a new independent audit was not authorized"
judging_build: "moai-tree5 — reports v3.2.0-tree 8dc835756 built unknown; the Commit ldflag was set by hand to 8dc835756 when the binary was built, while this tree's HEAD is 3275706cc9 (the build stamp does not name the tree it judged); working tree carried uncommitted SPEC-doc edits only"
```

Lint for 0.2.1 (judging build stamp `8dc835756`, hand-set; tree HEAD `89f4e812a3` + uncommitted 0.2.1 SPEC edits):

```
$ moai-tree5 spec lint --strict .moai/specs/SPEC-AGENTIC-FACTORY-IDENTITY-001/spec.md
WARNING   REQTableRowsRejected           .../SPEC-AGENTIC-FACTORY-IDENTITY-001/spec.md  65    table NOT read as a requirement-definition table (rejected_rows=7) ...
INFO      OwnershipTransitionUnmeasured  .../SPEC-AGENTIC-FACTORY-IDENTITY-001/spec.md  1     ... commit 3275706cc9 ... has no Authored-By-Agent trailer — ownership transition unmeasured

0 error(s), 1 warning(s)
exit=0
```

The warning points at the §E REQ→AC trace table (7 rows, deliberately no SHALL text); the seven requirement rows carry SHALL and are collected. The first lint run of this revision reported 7 `CoverageIncomplete` warnings because acceptance.md declared mappings as "주 REQ: REQ-…" — not a form the sibling-coverage extractor reads (`internal/spec/lint_coverage_sibling.go`). The lines now read "주 REQ: maps REQ-…" and the warnings are gone.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
