# Progress: SPEC-TODO-LAND-AUTO-DONE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
grounding:
  tree: WT-auto-done-on-land
  head: 74d872aaf
  surfaces_verified:
    - internal/cli/todo.go (newTodoDoneCmd, todoRequireLanded)
    - internal/cli/todo_landed.go (runTodoLanded, validateSuppliedSHA)
    - internal/cli/todo_undone.go (newTodoUndoneCmd, RestoreCard)
    - internal/kanban/prlink_landed.go (GitLandedQuerier, subjectAttribution)
    - internal/kanban/landing_evidence.go (LandingEvidence, closed operator provenance)
    - internal/kanban/integration_lock.go + internal/cli/integration.go (window)
    - internal/kanban/backlog_store.go (BacklogItem, states, ArchiveCard)
trigger_decision: option-c-scan-at-remote-landing-confirmation
req_count: 14
ac_count: 17
clarification_markers: 0
delta_history:
  - 0.2.0 — D1-D3 plan-audit delta fix (FAIL 0.75): AC-AD-015 (--fetch fetch count),
    AC-AD-016 (close-surface exclusivity + todo-landed control run), four-token skip
    vocabulary in REQ-AD-010, cite fix (prlink_landedref.go:63), pinned AC-AD-008,
    lane decisions recorded below
  - 0.3.0 — re-audit PASS 0.875 final delta: E1 AC-AD-016 control-run clause re-worded
    to the observable contract (State/position/text/spec-id unchanged, no archive
    entry; only the landing field may change — runTodoLanded writes the Landing column,
    internal/cli/todo_landed.go:127-140); E2 doc-step AC folded back into DoD §D.3
    (REQ-AD-014 traced at DoD); new AC-AD-017 added for the lead's field-observed
    FALSE-NEGATIVE fixtures — t603 (landing commit d8b7836aa, comma-form subject
    'fix(hooks): sync-phase 게이트의 C++ 검사 복구 (t603, H08)') must attribute, and t681
    (merge ae980ef2d carries the id non-attributing, repair 4fb28a5c6 subject has no id)
    must close on the recorded Landing.SHA — stated in contrast to the still-binding
    reissued-id collision gates (t654/t656/t657, AC-AD-004/005); provenance: lead
    observation 2026-09-13, both cards hand-closed via 'landed --sha' after
    '--require-landed' false rejections
lane_decisions:
  sync_gate: strict 'status: completed' only — cards close after sync; matches the
    run-landed/sync-pending guard (REQ-AD-007)
  body_negation: stays OUT of scope as residual risk — SPEC-TODO-LANDING-ATTRIBUTION-001's
    domain; only subject-level negation is guarded (REQ-AD-008)
  audit_log_location: RuntimeStateDirForRoot — machine-local, matches queue locality
    (REQ-AD-011)
  scheduler: none ships in this SPEC — cron/loop wiring is a sync-phase doc-surface
    decision (REQ-AD-002 keeps the scan invoked, never ambient)
```

## §E.2 Run-phase Evidence

### M1 — Guard predicates in internal/kanban (2026-09-13)

- `internal/kanban/autodone_scan.go` (new): closed four-token skip vocabulary
  (`AutoDoneSkipReasons`), evidence-form constants (`sha-recorded` /
  `subject-attribution`), `AutoDoneTri` three-valued gate answers,
  `AutoDoneFacts`/`AutoDoneDecision`/`AutoDoneDecide` (the pure scan policy:
  M2 sync gate → form 1 recorded SHA → M1 collision gate on form 2 → form 2 /
  inconclusive / not-landed), `AutoDoneDistinctTexts` (the reissued-id
  collision counter), `LandedCommit` + `LandedScanArgs` +
  `ScanLandedSubjects` + `LandedAttributions` (the one-query SHA-keyed
  subject stream).
- `internal/kanban/prlink_landed.go`: form 2c comma-form trailing
  parenthetical (AC-AD-017 Shape A) + the negation exclusion (REQ-AD-008,
  `not merged` / `not landed`, case-insensitive, subject-stream only) added
  to `subjectAttribution`; the id must OPEN the group and a group carrying
  two distinct card tokens still attributes nothing (pinned boundaries kept:
  t80 branch-name control unchanged).
- `internal/kanban/prlink_landed_corpus_test.go`: corpus partition re-pinned
  309/38 → 310/37 — exactly one id (t68) reclassified by form 2c; its corpus
  subject `merge: Factory Mode -f N worker fan-out (t68,
  SPEC-FACTORY-WORKER-FANOUT-001)` is the comma-form shape AC-AD-017 makes
  first-class. Re-measured against the SAME pinned commit 7835148d3.
- `internal/kanban/autodone_scan_test.go` (new): M1 tests — negation both
  directions + body-negation-out-of-scope edge, comma-form + the two negative
  boundaries, collision counting (reissue ≠ same text), decision table
  covering every skip token and precedence pair, closed-vocabulary test,
  scan-argv shape tripwire.
- RED evidence: initial run of the new tests failed with
  `undefined: AutoDoneDistinctTexts / ScanLandedSubjects / LandedAttributions
  / AutoDoneTri / AutoDoneFacts ... [build failed]` (captured verbatim before
  GREEN).
- GREEN evidence: `go test ./internal/kanban/ -count=1` →
  `ok github.com/modu-ai/moai-adk/internal/kanban 172.818s` (whole package,
  this run, this tree).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
