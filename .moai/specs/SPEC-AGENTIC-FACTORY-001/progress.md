# SPEC-AGENTIC-FACTORY-001 — progress

status: draft

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: revised-unaudited
plan_iteration: 6
plan_iteration_kind: delta (iteration-5 defects D1-D6 + leader-relayed canon alignment)
spec_version: "0.6.5"
plan_complete_at: 2026-10-29
prior_verdicts:
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter1.md (FAIL 0.69)"
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter2.md (FAIL 0.69, blocking R-D5 N1-N8)"
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter3.md (FAIL 0.69, blocking D1-D7)"
  - ".moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter4.md (FAIL 0.76, blocking N1-N4, split proposal)"
  - "iteration-5 findings D1-D6 as relayed in the leader brief (verdict file not re-read in this run)"
operator_decisions: "DEC-AF-01, DEC-AF-02 recorded (spec.md §C.7); open operator decisions OD-1..OD-12 (spec.md §G.1)"
child_spec: SPEC-AGENTIC-FACTORY-IDENTITY-001 (0.2.1, draft)
tree_head: 5b975dd1c0 (0.6.4 commit; 0.6.5 edits uncommitted on top; 0.6.3 committed as d4685469cd, 0.6.2 as 09142133bc, 0.6.1 as 5b7035403e, 0.6.0 as 89f4e812a3)
revision_0_6_1: "follow-up corrections — reserved decision row removed (no renumbering), unset wip_limit / checkpoint-cap behaviour moved to OD-7 / OD-8, byte check of the eight reason strings against lane-18 Go constants (8/8 match), lane-18/lane-20 relayed facts and gaps added"
revision_0_6_2: "intake.max_concurrent unset behaviour moved from AC-AF-012 (d) to OD-9 (AC-AF-012 still 10 + 5 PASS lines); code-vs-SPEC differences (a)-(e) kept as observation-only gaps; op-not-defined location stated per leader correction; OD-10 added (collector baseline over the claim-document limit; no recommendation, no new REQ/AC); PARKED reason spelling aligned to canon S46 (stop_unknown, start_unknown); OD-11 (checkpoint push retry count) and OD-12 (operator-hold as a PARKED reason) added without recommendation; design §2.3 old-state mapping replaced by the relayed leader decision; observation-only fact records §H (f)-(j) and §B.2 rows added"
revision_0_6_3: "leader correction applied — design §2.3 old-state mapping is evidence-based per canon E03 (writer handoff receipt + leader handoff acceptance => REVIEWING; merged=true => MERGING; otherwise BUILDING/IMPLEMENT + check evidence; blocked => current stage + PARKED; on-hold => OD-12); §H (i)/(j) rewritten as observations closed by the rule; §H (k) added (precedence unstated, relation to E03 guard list); §B.2 leader report on lane-18 uncommitted core/legacy.go recorded unverified; no code change demanded; REQ/AC counts unchanged; lane-18 af80e0bdec provisional values (items 1,2,3,5,7,6; item 4 not supplied) and the leader decision on stage-changing edges while PARKED recorded as observations in §H/§B.2, checked against canon S46 E05/E15/E16/E18/E19/E20/E24 (conflict observations for 3, 5, 7; circularity noted for 6)"
revision_0_6_4: "leader correction recorded as observations in §H/§B.2 (items 3, 5, 7 were misclassified as canon-silent; S46 L70/L61/L65 are not silent); leader's code-side instruction to lane-18 recorded as reported, unverified facts; item 4 'not supplied' confirmed; leader decision on stage-changing edges while PARKED kept verbatim; no code change demanded; REQ/AC counts unchanged"
revision_0_6_5: "observations added to spec §H/§B.2 — lane-18 52363f7ac7 (E03 guard covers all nine S46 L49 items, E15/E16 outcome-specific guards, E19 reason refused outside core.ClassifiedReasons; the 11 values read by this run from that commit) and two leader decisions (old state kept as progress record with FromLegacy-derived stage/sub/park; merge-gate reason G moves from overlay to a stage-holding gate record, behaviour change 'Resume after blocked' -> 'revalidate with new evidence at the same stage'); landing rule recorded in §K; canon basis not judged; no code change demanded; REQ/AC counts unchanged"
judging_build: "moai-tree5 — reports v3.2.0-tree 8dc835756 built unknown; the Commit ldflag was set by hand to 8dc835756 when the binary was built, while this tree's HEAD is 3275706cc9 (the build stamp does not name the tree it judged); working tree carried uncommitted SPEC-doc edits only"
req_count: 25
ac_count: 25
id_set_changed: true (ids kept 001-025; meanings re-mapped — crosswalk in research.md R12)
note: "revised, unaudited — no audit verdict exists for 0.6.0, 0.6.1, 0.6.2, 0.6.3, 0.6.4 or 0.6.5; a new independent audit was not authorized"
```

Lint for 0.6.5 (judging build stamp `8dc835756`, hand-set; tree HEAD `5b975dd1c0` + uncommitted 0.6.5 SPEC edits):

```
$ moai-tree5 spec lint --strict .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md
WARNING   REQTableRowsRejected           .../SPEC-AGENTIC-FACTORY-001/spec.md  217   table NOT read as a requirement-definition table (rejected_rows=25) ...
INFO      OwnershipTransitionUnmeasured  .../SPEC-AGENTIC-FACTORY-001/spec.md  1     ... commit e14b5bbac2 ... has no Authored-By-Agent trailer — ownership transition unmeasured

0 error(s), 1 warning(s)
exit=0
```

The remaining warning points at the §E REQ→AC trace table (spec.md line 217 in 0.6.5; it moved with the added HISTORY and §B.2 rows), which deliberately carries no SHALL text.

The first lint run of this revision exited 1 with 33 warnings: 25 `CoverageIncomplete` (acceptance.md wrote "주 REQ: REQ-…", which the sibling-coverage extractor does not read; now "maps REQ-…") and 7 `VacuousTestAssertion` (unanchored `-run` patterns and a `--- PASS:` prefix without a trailing delimiter). While fixing the latter, every `-run '^(TestAcceptanceManifest/…|…)$'` form was found to select no manifest subtest at all — Go splits `-run` levels only on a `/` outside parentheses (Go 1.26.8 `src/testing/match.go` `splitRegexp`, read in this run). All such commands now use `'^TestAcceptanceManifest$/^(…)$|^(…)$'` (acceptance §A).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §K Landing rule (leader decision, progress record)

Source: 'leader decision, relayed via the leader session' (recorded 2026-10-29, SPEC 0.6.5). Measurements marked "(this run)" were taken in this worktree with read-only git commands; nothing was fetched, so "local only" is the leader's statement plus the absence of remote-tracking refs as of the last fetch.

- The AF work is held from `main`. No push and no PR now.
- Product template commit `d968a67784` on branch `WT-agentic-factory-templates` is local only. It changes 45 files, +4856 lines: 44 added and 1 modified, the modified one being `internal/template/embed_manifest_gen.go` regenerated by `make embed-manifest`. (This run: `git show --shortstat d968a67784` → `45 files changed, 4856 insertions(+)`; `git diff --name-only --diff-filter=A d968a67784^ d968a67784` → 44 lines; `git show --name-status d968a67784 -- internal/template/embed_manifest_gen.go` → `M`; branch tip `WT-agentic-factory-templates` = `d968a67784`; `git branch -r --contains d968a67784` → no output.)
- At AF landing time: absorb the then-current `main` FIRST, then regenerate `internal/template/embed_manifest_gen.go` with `make embed-manifest` — NEVER hand-merge that file — then run `make embed-manifest-check`.
- PR #1944 (t1985, lane-16) lands first and touches the same generated file.
- Settings reader commits are local only: `5c278d8b51` (collaboration keys; tip of `WT-agentic-factory-settings-keys`, this run), `6c833ee4b4` (provisional `pr_queue.wip_limit` key) and `8920c0ecf8` (browser and codex opt-in reader). This run: both of the latter are contained in `WT-agentic-factory-settings-keys` and `WT-agentic-factory-integration`; `git branch -r --contains` for `5c278d8b51`, `6c833ee4b4` and `8920c0ecf8` → no output each.
