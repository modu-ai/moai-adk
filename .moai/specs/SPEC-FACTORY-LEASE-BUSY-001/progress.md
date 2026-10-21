# Progress — SPEC-FACTORY-LEASE-BUSY-001

status: draft (SSOT is the `spec.md` frontmatter)

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-10-21 (card t1916, Tier M): spec.md, plan.md, acceptance.md,
  progress.md, decision-index.md (decision gate on). Plan-start HEAD `8249bff8c3`.
- Plan audit iter1: FAIL 0.69 — repairs D1-D10 applied (verified closed by iter2). Plan audit
  iter2: FAIL 0.90 (2 blocking D12/D13 + minor D11/D14/D15; `.moai/reports/t1916/plan-audit-iter2.md`)
  — its repair batch D11-D15 applied across spec.md / plan.md / acceptance.md / decision-index.md
  (all five verified RESOLVED by iter3). Plan audit iter3: FAIL 0.97 (2 blocking D16/D17, both
  confined to plan.md §F M4 clause (3); `.moai/reports/t1916/plan-audit-iter3.md`) — its
  single-clause repair applied in 8a2f6683eb. Ceiling release + confirming-audit authorization
  recorded in §G. Plan audit iter4 (scoped-delta confirming round): **PASS 1.00** at
  017bd3f3158ae2ccb361968b3862d47cae156af7 (`.moai/reports/t1916/plan-audit-iter4.md`, receipt
  `rcpt-050c5f674d104c5287e12f3d`, convergence ok; D16/D17 RESOLVED, D1-D15 standing) —
  `plan_status: audit-ready`.
- Open gates disposition at kickoff: decision-index Q1 default (exit 4 for every class) is
  encoded in the requirement layer (REQ-FLB-002/003/004/007) per the iter1 routing assessment —
  operator disposition remains open to close (DoD #3), non-blocking. Q2 resolves in-card (M4
  baseline, AC-FLB-009 baseline-relative), dispositioned by close. M0 hold LIFTED: PR #1859
  merged (17a3b98de2 in main history); the leader's card dispatch (2026-10-21, run tn8rpj) is
  the resume signal.

## §F Phase 4 Mode Selection

Decision: serial (the evaluation table and justification follow).

Input parameters: Tier M; run-phase scope ~8 files (plan.md §C: factory_card.go + test files,
todo store read path); 1 domain (Go CLI/factory); language mix Go-only; concurrency benefit LOW
(coding-heavy, single-domain); Agent Teams prereqs not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | semantic multi-file change |
| serial | **yes** | coding-heavy single-domain work per the Anthropic caveat; per-milestone manager-develop delegation (M1-M5 in order, M4 baseline before M5) |
| fanout | no | no research fan-out need; intra-card write serialization |
| sweep | no | not mechanical-uniform; < 30 files |

Decision: serial
Justification: the change is one cohesive behavior reclassification across a lease path and its
tests; milestones share state and ordering constraints (M4's baseline must precede M5), so a
single sequential implementer with per-milestone commits is both the safe default and the
cheapest envelope. Kickoff gate met in autonomous form: plan-audit iter4 PASS 1.00 (≥ Tier M
0.80, zero blocking, receipt rcpt-050c5f674d104c5287e12f3d), plan artifacts unchanged since the
verdict (audited_sha 017bd3f315... = current HEAD at audit; the only later commits are this
progress record and the verdict file), M0 lifted, no open blocker (§G ceiling authorization
recorded). Decision record: §G ceiling-outcome entry + this section; the sync audit re-reads
both.


## §E.2 Run-phase Evidence

### M0 — hold gate

LIFTED before this phase opened: PR #1859 merged (17a3b98de2 in main history); the leader's
card dispatch (2026-10-21, run tn8rpj) is the resume signal (recorded in §E.1). No code, skipped
per the dispatch.

### M1 — RED: cause-class contract (test commit)

- Pre-flight (2026-10-22, tree aaab995138, clean): `go build ./internal/cli/ ./internal/factory/`
  exit 0; `go vet ./internal/cli/` exit 0; `golangci-lint run --timeout=2m ./internal/cli/` —
  baseline 2 pre-existing staticcheck ST1008 findings in `internal/cli/paths.go` (untouched by
  this SPEC); `grep -rn 'Retired\|superseded' internal/cli/factory_card.go internal/cli/todo.go`
  — one unrelated `factoryRelaunchSupersededNote` constant (card t1554), no conflicts.
- **AC-FLB-004 pre-change capture (AC evidence pair, REQ-FLB-007 text preservation).** Captured
  2026-10-22 on the pre-change tree — production code identical to commit `aaab995138` (the
  capture test file is the only addition) — via
  `go test ./internal/cli/ -run TestFactoryLeaseRefusalTextCapture -count=1 -v` (PASS, 4/4
  subtests). Verbatim rendered stderr lines:
  - serial-slot: `factory next: refused serial-slot: another serial card is in flight and holds the serial slot`
  - quota-hold: `factory next: refused quota-hold: quota hold: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z`
  - backend-skip: `factory next: refused backend-skip: a Codex lane cannot advance a card recorded at merge-ready`
  - foreign-worktree: `factory next: refused foreign-worktree: the worktree directory <fixture-root>/.moai/worktrees/t1 already exists and no card record names it; a card never adopts another card's tree (remove or rename the directory first)`
    — the foreign-worktree line embeds the fixture root path (environment-derived, per-invocation
    temp dir); its byte-freeze comparison normalizes that path prefix only.
- **Class table (the M1 deliverable per plan.md §F M1; AC-FLB-001):**

| cause class (REQ-FLB-001) | reason token | marking | detail names | wait-retryable | exit |
|---|---|---|---|---|---|
| race lost to another claimant (REQ-FLB-002) | `raced` | `homestate.ErrStaleVersion` / `ErrLeaseHolder` surviving the retries (unmarked) | the claimant lost the version-checked edge | yes (unchanged) | 4 |
| record-write contention (REQ-FLB-003) | `record-busy` | `errFactoryRecordBusy` (existing sentinel, narrowed to live-deadline busy) | the record lock held past every retry | yes (new) | 4 |
| claim-deadline expiry (REQ-FLB-004) | `claim-deadline` | `errFactoryClaimDeadline` (new; takes precedence over a joined busy error — spec § Constraints) | the expired claim deadline + the presence/absence of a concurrent busy error; neither a host-load-only nor a race-only cause asserted | yes (new) | 4 |
| queue-lock-held (out of scope, Q4) | `raced` | — PRESERVED at factory_card.go 1527-1528 | the queue lock's wait budget | yes (unchanged) | 4 |

  Bare-form leases keep their current exit (1) and plain-error surface; the marked error text
  names the true class (AC-FLB-004 second clause).
- **Verbatim RED output** (command: `go test ./internal/cli/ -run
  'TestFactoryLeaseClaimClassesWaitRetryable|TestFactoryLeaseDeadlineClassRefusalToken|TestFactoryLeaseRecordContentionClassRefusalToken|TestFactoryLeaseClaimFailureMarksDistinctClasses'
  -count=1 -v`; FAIL exit 1, all four tests red for the stated reason — the misclassification
  this SPEC fixes, not an unrelated pre-existing failure):
  - `waitable("record-busy") = false, want true` / `waitable("claim-deadline") = false, want true`
  - deadline e2e: `stderr = "factory next: refused raced: t1 was not leased: the factory record
    stayed busy — the claim deadline (2.2s) passed ... (SQLITE_BUSY); retry when the host is
    quieter\n", want exactly one line "factory next: refused claim-deadline: "<detail>`
  - contention e2e: `stderr = "factory next: refused raced: t1 was not leased: the factory record
    stayed busy — another writer held the record past every retry: database is locked (5)
    (SQLITE_BUSY); retry when the host is quieter\n", want exactly one line "factory next:
    refused record-busy: "<detail>`
  - marking seam: `the joined marking = the factory record stayed busy — the claim deadline
    (2.2s) passed while a record write was still running: the host is under load, not a lock
    race: database is locked (5) (SQLITE_BUSY), want its detail to name the concurrent busy
    error's presence`
- Sanctioned pin flips (the misclassification pins REQ-FLB-003/004 legitimately flip; plan §B):
  `TestFactoryLeaseRecordStallBounded` (raced assertion + doc comment) and
  `TestFactoryLeaseMidClaimStallBounded` (raced assertion + doc comment) — both arms produce the
  joined deadline+busy error (the deadline is the outer bound), so they flip to `claim-deadline`.
  Additionally, REQ-FLB-004's signal-naming detail mandate rewords the claim-deadline detail, so
  two deadline-surface tests flip their phrase assertions:
  `TestFactoryClaimDeadlineRefusalNamesTheDeadline` (bare-form detail phrase) and
  `TestFactoryNominatedBusyRefusalCarriesTheClaimError` (token + detail phrase).
  `TestFactoryLeaseClaimRespectsWaitCap`'s marking assertion widens to accept both non-race
  markings (its invariant — the give-up is a hard busy refusal, not a race to retry — is
  unchanged). `TestFactoryLeaseQueueLockStallBounded` (subtests `nominated` and `bare`) stays
  untouched (AC-FLB-010's unchanged split).


## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §G Override and Refusal Record

- 2026-10-21T12:25:00Z SPEC-FACTORY-LEASE-BUSY-001 ceiling-outcome outcome=operator-authorized-confirming-audit reasons="the plan-audit ceiling is exhausted at iter3 of 3 (score trajectory 0.69→0.90→0.97, monotonic, no STOP; must-pass 0 failed; iter2 defects D11-D15 all verified RESOLVED by iter3). The residual blocking pair D16/D17 is confined to one clause (plan.md §F M4 (3) seed recipe) with the auditor-prescribed fix applied in commit 8a2f6683eb. The leader — under operator delegation (+Jev, p=0.74), citing the t1834 iter13 and t1831 iter6 precedents — authorized exactly ONE scoped-delta confirming audit (iter4) after the repair. Terms: iter4 PASS (≥0.80 Tier M, zero blocking) admits the M0/kickoff review; an iter4 miss records a hold with no further rounds and re-requests." evidence=".moai/reports/t1916/plan-audit-iter3.md (rcpt-c010efb4f9d4c78c7b1e6ce8), leader ceiling-grant cross-session message 2026-10-21, repair commit 8a2f6683eb" decided_by=leader under operator delegation relayed-by=lane-6 ladder_path=gate boundary (operator-held ceiling release)


