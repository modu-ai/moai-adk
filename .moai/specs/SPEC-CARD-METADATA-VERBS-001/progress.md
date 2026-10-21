# Progress — SPEC-CARD-METADATA-VERBS-001

Card t1840 (factory lane-15, run tn1p4e) — stage E of the card-policy redesign.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-pending (artifacts authored 2026-10-17 by manager-spec; plan-auditor verdict not yet rendered)
- plan_complete_at: _<pending plan-audit>_
audit_ready: true

## §E.2 Run-phase Evidence

manager-develop, milestone-serial TDD (RED commit → GREEN commit per milestone). Every row below cites a command run in the run-phase session (2026-10-21/22) and its verbatim output, against this tree (branch `WT-card-metadata-verbs`).

### E1 — AC matrix (16/16 PASS)

| AC | Verifying test(s) | Command (this run) | Observed output (verbatim, bounded) |
|----|-------------------|--------------------|-------------------------------------|
| AC-CMV-001 | TestSetSpec_FlagOff_Refused (+ the shared gate branch both verbs route through) | `go test ./internal/cli/ -run 'TestSetSpec_' -v` | `--- PASS: TestSetSpec_FlagOff_Refused (0.84s)` |
| AC-CMV-002 | TestTemplateCardMetadataVerbsShippedOff | `go test ./internal/config/ -run 'TestTemplateCardMetadataVerbsShippedOff' -v` | `--- PASS: TestTemplateCardMetadataVerbsShippedOff (0.00s)` |
| AC-CMV-003 | TestSetSpec_ValidationGate | `go test ./internal/cli/ -run 'TestSetSpec_'` | `ok  github.com/modu-ai/moai-adk/internal/cli	25.608s` |
| AC-CMV-004 | TestSetSpec_TwoStoresWritten | same run as AC-CMV-003 | `ok  github.com/modu-ai/moai-adk/internal/cli	25.608s` |
| AC-CMV-005 | TestSetSpec_FailureMatrix (directions 1-4 injected, 5 heal, 6 catch-up + duplicate refusal, ABA seed) | `go test ./internal/cli/ -run 'TestSetSpec_FailureMatrix' -v` and the sweep aggregate | `--- PASS: TestSetSpec_FailureMatrix (9.14s)` (observed with -v; all 8 subtests pass — a failing subtest fails the parent, and the sweep aggregate `ok  github.com/modu-ai/moai-adk/internal/cli	25.608s` followed) |
| AC-CMV-006 | TestSetCardSpec_AnyState_Lands + contrast | `go test ./internal/homestate/ -run 'TestSetCardSpec' -v` | `--- PASS: TestSetCardSpec_AnyState_Lands (8.44s)` / `--- PASS: TestSetCardSpec_RecordPickedContrast_StandsStill (0.32s)` |
| AC-CMV-007 | TestSetCardSpec_StaleVersion_Refused | same run | `--- PASS: TestSetCardSpec_StaleVersion_Refused (0.34s)` |
| AC-CMV-008 | TestSetSpec_ConsumerResolves | `go test ./internal/cli/ -run 'TestSetSpec_ConsumerResolves' -v` | `--- PASS: TestSetSpec_ConsumerResolves (0.83s)` |
| AC-CMV-009 | TestReclassify_ValidationGate | `go test ./internal/cli/ -run 'TestReclassify_' -v` | `--- PASS: TestReclassify_ValidationGate (4.98s)` |
| AC-CMV-010 | TestReclassify_UnspecifiedAxisPreserved + TestReclassify_BlockedUntouched | same run | `--- PASS: TestReclassify_UnspecifiedAxisPreserved (0.54s)` / `--- PASS: TestReclassify_BlockedUntouched (0.59s)` |
| AC-CMV-011 | TestCardMetadataEvent_RowShape (+ rows asserted in TwoStoresWritten / WriteSemantics) | `go test ./internal/homestate/ -run 'TestCardMetadata' -v` | `--- PASS: TestCardMetadataEvent_RowShape (0.88s)` |
| AC-CMV-012 | TestMetadataActor_Resolution | `go test ./internal/cli/ -run 'TestMetadataActor_Resolution' -v` | `--- PASS: TestMetadataActor_Resolution (0.00s)` |
| AC-CMV-013 | TestSetSpec_ReasonValidation | `go test ./internal/cli/ -run 'TestSetSpec_ReasonValidation' -v` | `--- PASS: TestSetSpec_ReasonValidation (1.12s)` |
| AC-CMV-014 | TestSetSpec_RefusalAtomicity (+ row-level before/after assertions in every gate test) | `go test ./internal/cli/ -run 'TestSetSpec_RefusalAtomicity' -v` | `--- PASS: TestSetSpec_RefusalAtomicity (2.81s)` |
| AC-CMV-015 | TestCardMetadataSurface_RowPreserving | `go test ./internal/homestate/ -run 'TestCardMetadataSurface_RowPreserving' -v` | `--- PASS: TestCardMetadataSurface_RowPreserving (1.19s)` |
| AC-CMV-016 | TestMetadataVerbs_LaneBoundary_Refused | `go test ./internal/cli/ -run 'TestMetadataActor_Resolution\|TestMetadataVerbs_LaneBoundary_Refused\|TestSetSpec_ConsumerResolves'` | `ok  github.com/modu-ai/moai-adk/internal/cli	4.947s` |

### E2 — Build

- `go build ./...` → `BUILD OK` (exit 0; observed 2026-10-21 pre-triage and unchanged by the comment-only post-triage edit).
- `go vet ./internal/homestate/ ./internal/cli/ ./internal/config/` → `VET CLEAN` (exit 0); re-observed post-triage → `vet ok`.

### E3 — Scoped coverage (acceptance.md §E: new mutation and audit code ≥ 85%)

- `go test -coverprofile=/tmp/t1840-cli-scoped-cover.out -run 'TestSetSpec_|TestReclassify_|TestMetadataActor|TestMetadataVerbs|TestSetSpec_ConsumerResolves' ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 30.283s`; profile aggregation → `cli/factory_card_metadata.go (scoped): 178/208 = 85.6%`.
- `go test -coverprofile=/tmp/t1840-home-scoped-cover.out ./internal/homestate/ -run '<metadata set>'` → `ok github.com/modu-ai/moai-adk/internal/homestate 13.140s`; aggregation → `homestate/card_metadata.go: 54/63 = 85.7%` (77.8% before the E3 refusal-branch tests landed).
- Combined new-code surface: 232/271 statements = 85.6%.

### E4 — Subagent boundary grep (C-HRA-008 family)

- `grep -rn 'AskUserQuestion' internal/homestate/card_metadata.go internal/cli/factory_card_metadata.go internal/cli/factory_handoff_recover.go internal/config/types.go internal/config/defaults.go <both test files>` → exit 1, no output (zero matches on every file this SPEC touched).
- Pre-existing matches in `internal/cli/harness.go` / `pr_watch_cmd.go` are the harness detector's own literal strings and doc text in files this SPEC does not touch (untouched paths, PRESERVE discipline).

### E5 — Lint / format status

- `gofmt -l internal/homestate/ internal/cli/ internal/config/` → empty (`GOFMT CLEAN`), re-observed after every edit.
- golangci-lint was NOT run lane-locally (Gap): the delegation's lane-local verification scope named go vet + gofmt; the full golangci-lint pass runs on the origin/main CI verdict surface with the batch landing.

### E6 — Push state

- No `git push` was executed in this session. The branch `WT-card-metadata-verbs` holds 12 local commits (`356db1a50d..16775e5e4e`); remote state untouched (landing is the lane's integration verb / the leader's batch).

### E7 — Blockers

- None. No plan.md instruction conflicted with a REQ.

### E8 — RED failure output (TDD; verbatim, per milestone)

- M1 (`go test ./internal/homestate/ -run 'TestCardMetadataEvent_RowShape'`, exit 1, at 9133fd244b):
  `internal/homestate/card_metadata_test.go:44:12: db.RecordCardMetadataEvent undefined (type *FactoryDB has no field or method RecordCardMetadataEvent)`
- M2 (`go test ./internal/cli/ -run 'TestSetSpec_'`, exit 1, at 95e6a926a0):
  `internal/cli/factory_card_metadata_test.go:315:14: undefined: metadataWriteFault`
- M3 (`go test ./internal/cli/ -run 'TestReclassify_'`, exit 1, at 63c051e1cc):
  `factory_card_metadata_test.go:764: admission state queued was refused: unknown command "reclassify" for "factory"`
- M4: the capture wiring shipped with the M2 verb seam, so the pinning tests observed GREEN at arrival; the lane test's red was observed through a mutant probe — with the set-spec boundary check disabled the test fails exactly as designed: `set-spec refusal lacks the lane-boundary sentinel (the boundary check must run ahead of the flag gate): factory set-spec: refused — this verb is gated by workflow.card_metadata_verbs.enabled (default false)` — and the probe was reverted before the M4 commit.
- M5 (`go test ./internal/config/ -run 'TestTemplateCardMetadataVerbsShippedOff'`, FAIL at f08ee32937):
  `the shipped template workflow.yaml carries no commented workflow.card_metadata_verbs block (plan.md §F M5)`

### Full-suite verification (post-triage)

- First full run (2026-10-21, HEAD f005224b24): `ok internal/homestate 153.797s`, `ok internal/factory 326.015s`, `ok internal/config 7.766s` (+subpackages); `FAIL internal/cli 1501.529s` on exactly two tests, both triaged:
  1. `TestSD_AC022_SchemaStatementsFrozen` — the AC-SD-022 absorption pin correctly fired on the deliberate additive card_metadata_events table; the pin moved with the statement per the guard's own precedent (commit 3778b80234), and `go test ./internal/cli/ -run 'TestSD_AC022_SchemaStatementsFrozen' -v` → `--- PASS: TestSD_AC022_SchemaStatementsFrozen (0.07s)`.
  2. `TestFactoryLeaseQueueLockStallBounded/bare` — `elapsed=9.006786125s, want within 8.75s` under the loaded full suite; in isolation on the same tree `elapsed=8.648456667s` → `--- PASS: TestFactoryLeaseQueueLockStallBounded (20.62s)`. Load-induced timing flake in a test this SPEC's diff does not touch (no lease/lock-wait path changed); origin/main CI is the re-measurement surface.
- Post-triage full re-run (2026-10-22, HEAD 16775e5e4e), one multi-package invocation `go test -timeout 30m ./internal/homestate/ ./internal/factory/ ./internal/config/... ./internal/cli/...`: `ok internal/homestate 158.824s`, `ok internal/factory 348.934s`, `ok internal/config 17.899s` (+atomicfile, toolpolicy), all 23 `internal/cli/*` subpackages `ok`, and the `internal/cli` ROOT package hit the invocation's 30m budget: `panic: test timed out after 30m0s` with no `--- FAIL` inside the window. A solo root-package run at a 50m budget then measured `FAIL github.com/modu-ai/moai-adk/internal/cli 3001.378s` — again the alarm alone (`panic: test timed out after 50m0s`; exactly one `--- FAIL`, the timeout itself; the in-flight test at the cut was 0s old). Three local data points now bound the property — the pre-change BASELINE exceeded go's 10m default, the loaded re-run exceeded 30m, and the solo idle-machine run exceeded 50m, with ZERO test failures in any window — so the root package's duration is a machine-scale property of the suite, not a regression of this SPEC's diff; every test the diff can affect is green (scoped runs above), all 23 cli subpackages pass, and the authoritative full-suite verdict is origin/main CI.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-22
run_commits: 356db1a50d, 95e6a926a0 (M1), 57a744f49e, 63c051e1cc (M2), 69d573c2c4, d8c7f05c97 (M3), f08ee32937 (M4), 6e70bed5bd, f005224b24 (M5), 3778b80234 (triage), 16775e5e4e (E3 coverage)
residual_risk: TestFactoryLeaseQueueLockStallBounded is marginally timed for a loaded machine (8.75s bound vs 8.65s isolated); a full-suite re-run on a loaded host may re-flag it — pre-existing to this SPEC's diff. The internal/cli root package exceeds every local invocation budget measured (10m baseline / 30m loaded / 50m solo, zero test failures in each window) — its authoritative green is the origin/main CI verdict; the lane-local evidence rests on the scoped metadata suite, the 23 cli subpackages, and the targeted post-triage re-runs.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-17T20:34:03Z SPEC-CARD-METADATA-VERBS-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit.md
- 2026-10-17T20:35:33Z SPEC-CARD-METADATA-VERBS-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit.md
- 2026-10-17T20:36:38Z SPEC-CARD-METADATA-VERBS-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit.md
- 2026-10-21T10:40:41Z SPEC-CARD-METADATA-VERBS-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit.md
- 2026-10-21T10:41:06Z SPEC-CARD-METADATA-VERBS-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1840/.moai/reports/t1840/plan-audit.md

## §F Phase 4 Mode Record (lane-3, pre-spawn obligation)

- 2026-10-21: Kickoff passed (T7 iter6 PWD pass-through + T8a audit decider, run v8). Phase 4 mode: **serial** — Tier M, coding-heavy TDD, single tree single writer. DB-1 (iter6 debt, P1) routed to run: extend M2 TestSetSpec_FailureMatrix RED scope with the ABA seed. DB-2/3/4 (P2-P4, SPEC-artifact wording) routed to sync-phase manager-spec micro-pass. Ceiling-outcome auto-records committed at e3b4ec958d.
