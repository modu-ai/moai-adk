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

### M2 — CLI verb `moai todo auto-done` (2026-09-13)

- `internal/cli/todo_autodone.go` (new): the scan verb (`--fetch`, `--dry-run`,
  `--json`) wiring the M1 decision function; facts gathered OUTSIDE the queue
  lock, closes applied inside ONE locked `Mutate` whose callback re-checks
  each planned close (C3 byte-identity inherited); canonical close-line
  contract `done <id> landing=landed source=auto-land ref=<ref> form=<form>`
  (C4) + `skip <id> reason=<reason>` + summary last; exit 0 for skip
  outcomes, exit 1 only when the store is unreadable; `--fetch` runs exactly
  one `git fetch <remote> <branch>`, absent it zero network fetches;
  `--dry-run` writes nothing (no Mutate, no log rows); `recordFactoryCardState`
  preserved per closed card; append-only JSONL execution log under
  `RuntimeStateDirForRoot` (`auto-done-log.jsonl`) carrying card id, RFC 3339
  UTC instant, form, subject+SHA / recorded SHA, ref + ref head, skip reason,
  source `auto-land`; NO Landing-column writes (REQ-AD-009); help documents
  the two evidence forms, the closed four-token skip set, the exit-code
  policy, and the dry-run contract.
- `internal/cli/todo_undone.go`: `undone` appends a reversal row naming the
  original closure row when it restores a scan-closed card (REQ-AD-012);
  fail-open, manual closures log nothing.
- `internal/cli/todo.go`: verb registered; `internal/kanban/prlink_landed.go`:
  `LandedBranchFromRef` exported (single ref→branch derivation for the scan's
  one-pass attribution); `internal/cli/todo_surface_test.go`:
  `auto-done` declared as a permitted verb addition with the SPEC citation.
- `internal/cli/todo_autodone_test.go` (new): 21 tests covering
  AC-AD-001..017, the §D.1 edges (recorded SHA unreachable after a ref
  rollback → skip not-landed; unreadable SPEC → skip; no SpecID → gate not
  applied), the `--json` surface, and the help DoD. Coverage of
  todo_autodone.go: every decision path >= 89%, remaining sub-100% lines are
  IO-failure/stderr branches.
- RED evidence: first M2 run → `unknown command "auto-done" for "todo"`
  across the suite (verbatim captured), plus the fixture-layer discovery
  below.
- GREEN evidence: `go test ./internal/cli/ -run 'TestTodoAutoDone' -count=1`
  → `ok github.com/modu-ai/moai-adk/internal/cli 28.398s`, 19/19 PASS.

### M2 finding — the reissued-id fixture premise vs the identity invariant

`todo_identity.go` `ensureRecordIdentities` refuses, on EVERY whole-record
write, a record holding the same card id live AND archived. The AC-AD-004/005
"given" (predecessor archived + reissued live card in one record) is
therefore unwritable through the store — and AC-AD-005's close-through-
collision would be refused at the write even if SQL-injected. Resolutions
taken in run-phase (no SPEC body change required):

- AC-AD-004: the fixture injects the predecessor via direct SQL
  (`injectArchivedPredecessor`); the collision card skips, the scan issues no
  mutation, and the criterion's observable (skip `ambiguous-id`, card stays
  live) passes end-to-end.
- AC-AD-005: the DECISION is proven twice —
  `TestAutoDoneDecide/"recorded SHA closes through a collision"` (unit) and
  the CLI-level `--dry-run` run over the injected reissue state, which prints
  the `form=sha-recorded` close end-to-end; the non-dry write is then
  asserted to be REFUSED loudly by the identity invariant
  (`duplicate card identity`), queue record byte-identical. In the field the
  reissue predecessor sits in another store (the allocator's blind spot), so
  the live record holds no duplicate and the close applies.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-13
run_commit_sha: pending-backfill-run
run_status: implemented-run-phase
ac_pass_count: 17
ac_fail_count: 0
ac_notes:
  - AC-AD-004/005 fixtures inject the reissued-id predecessor via direct SQL;
    the store's identity invariant (todo_identity.go) refuses that state on
    any whole-record write — see §E.2 M2 finding for the resolution and the
    two-level (unit + CLI dry-run) proof of AC-AD-005's observable
ac_edge_cases:
  - recorded-SHA-on-different-ref → skip not-landed (decision table row)
  - SPEC unreadable → skip spec-not-completed (TestTodoAutoDone_SpecUnreadableIsNotCompleted)
  - no SpecID → gate not applied (AC-AD-001..003 fixtures carry no spec id)
  - body negation with clean subject → attributes normally
    (TestLandedPredicate_NegationIsSubjectStreamOnly)
l44_pre_commit_fetch: not-run (lane does not push; worktree-scoped)
l44_post_push_fetch: not-run (push is the lead's batch act)
new_warnings_or_lints_introduced: 0 (golangci-lint over internal/kanban +
  internal/cli: zero findings in touched files; 37 pre-existing baseline
  findings in untouched files)
cross_platform_build:
  darwin_arm64: go build ./... exit 0
  windows: not-run (lane-local; CI matrix owns it)
total_run_phase_files: 8
m1_to_m2_commit_strategy: one commit per milestone (M1 a5c0c0eca; M2 this)
milestones:
  - M1 guard predicates + scan substrate (kanban) — done
  - M2 CLI verb + log + reversibility — done
  - M3 lead-procedure docs (todo.md skill + template mirror) — deferred to
    sync phase per run-phase delegation instruction
doc_step_owner: manager-docs (sync phase), per delegation instruction
residual_risks:
  - the negation guard is subject-stream-only; a body negation with a clean
    attributing subject still attributes (recorded §D.1 edge, SPEC-TODO-
    LANDING-ATTRIBUTION-001's domain)
  - the collision gate can never fire on a writable in-store record while
    the identity invariant holds; it guards cross-store/divergent-queue
    reissue shapes and legacy records
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
