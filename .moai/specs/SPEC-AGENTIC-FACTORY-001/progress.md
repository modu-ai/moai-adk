# SPEC-AGENTIC-FACTORY-001 — progress

status: draft

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: revised-unaudited
plan_iteration: 6
plan_iteration_kind: delta (iteration-5 defects D1-D6 + leader-relayed canon alignment)
spec_version: "0.6.0"
plan_complete_at: 2026-10-29
prior_verdicts:
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter1.md (FAIL 0.69)"
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter2.md (FAIL 0.69, blocking R-D5 N1-N8)"
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter3.md (FAIL 0.69, blocking D1-D7)"
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter4.md (FAIL 0.76, blocking N1-N4, split proposal)"
  - "iteration-5 findings D1-D6 as relayed in the leader brief (verdict file not re-read in this run)"
operator_decisions: "DEC-AF-01, DEC-AF-02 recorded; DEC-AF-03 reserved — content not delivered (spec.md §C.7, §H)"
child_spec: SPEC-AGENTIC-FACTORY-IDENTITY-001 (0.2.0, draft)
tree_head: 3275706cc91a7453488bdaefcdfa73ce55bc0bb5
judging_build: "moai-tree5 — reports v3.2.0-tree 8dc835756 built unknown; the Commit ldflag was set by hand to 8dc835756 when the binary was built, while this tree's HEAD is 3275706cc9 (the build stamp does not name the tree it judged); working tree carried uncommitted SPEC-doc edits only"
req_count: 25
ac_count: 25
id_set_changed: true (ids kept 001-025; meanings re-mapped — crosswalk in research.md R12)
note: "revised, unaudited — no audit verdict exists for 0.6.0; a new independent audit was not authorized in this run"
```

Lint (judging build stamp `8dc835756`, hand-set; tree HEAD `3275706cc9` + uncommitted SPEC docs):

```
$ moai-tree5 spec lint --strict .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md
WARNING   REQTableRowsRejected           .../SPEC-AGENTIC-FACTORY-001/spec.md  195   table NOT read as a requirement-definition table (rejected_rows=25) ...
INFO      OwnershipTransitionUnmeasured  .../SPEC-AGENTIC-FACTORY-001/spec.md  1     ... commit e14b5bbac2 ... has no Authored-By-Agent trailer — ownership transition unmeasured

0 error(s), 1 warning(s)
exit=0
```

The remaining warning points at the §E REQ→AC trace table (spec.md line 195), which deliberately carries no SHALL text.

The first lint run of this revision exited 1 with 33 warnings: 25 `CoverageIncomplete` (acceptance.md wrote "주 REQ: REQ-…", which the sibling-coverage extractor does not read; now "maps REQ-…") and 7 `VacuousTestAssertion` (unanchored `-run` patterns and a `--- PASS:` prefix without a trailing delimiter). While fixing the latter, every `-run '^(TestAcceptanceManifest/…|…)$'` form was found to select no manifest subtest at all — Go splits `-run` levels only on a `/` outside parentheses (Go 1.26.8 `src/testing/match.go` `splitRegexp`, read in this run). All such commands now use `'^TestAcceptanceManifest$/^(…)$|^(…)$'` (acceptance §A).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
