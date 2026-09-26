# acceptance.md — SPEC-AUTONOMY-ESCALATION-001

23 acceptance criteria (AC-AE-001 … AC-AE-023), Given-When-Then, each binary. Requirements
stay in spec.md §D; this file is the verification layer only. Each criterion is one line so
the lint parser reads its REQ mapping; its test, RED-now, and green-path cells follow as
sub-bullets.

## §A — Tree pin, evidence ledger, test naming, and renumbering

Document-level pin: every RED-now cell below was measured on commit
`ca1d5dc435ae1fa28aed58add241f73c5d8b2ae0` (branch `WT-escalation-detector`, base `develop`)
unless a row carries its own pin. Commits after it up to `d8926ff9a` touched only this SPEC
directory, so no Go or template path differs between them.

| Ledger id | Command (single invocation) | stdout | exit | Pin |
|---|---|---|---|---|
| L1 | `grep -rn autonomy internal/template/templates/.moai/config/sections/workflow.yaml` | (empty) | 1 | `ca1d5dc43` |
| L2 | `grep -rln 'escalation-report\|EscalationRecord\|escalate_on' --include='*.go' internal cmd pkg` | (empty) | 1 | `ee9f57151` |
| L2+ | positive control, same form: `grep -rln 'BacklogStatePicked' --include='*.go' internal cmd pkg` (file count read with `wc -l`) | `33` files | 0 | `ee9f57151` |
| L3 | `grep -rn escalate_on internal/template/templates` | (empty) | 1 | `ca1d5dc43` |
| L4 | `grep -rn needs-decision --include='*.go' internal` (positive control for the grep shape: `BacklogStatePicked` → 17 non-test lines) | (empty) | 1 | `ca1d5dc43` |

The ledger establishes that no detector, no config key, and no escalation record exists on the
pinned trees — which is why every criterion below is RED-now. L2 is quoted
(`--include='*.go'`): unquoted, zsh aborts with `no matches found` and exit 1, indistinguishable
from the RED observation; L2+ shows the quoted form fires.

**Test naming (binding).** Each criterion names the Go test that verifies it and its package.
Detector unit tests live in `internal/escalation` (proposed package, design.md §F);
hook-integration tests live in `internal/hook`. If run-phase renames the package, the new path
is recorded in progress.md §E.2 and the test names stay unchanged. A criterion is GREEN only
when `go test <package> -run '^<TestName>$' -v` prints `--- PASS: <TestName>`; a run printing
`[no tests to run]` is an empty sweep and counts as RED, never as PASS.

**Renumbering at v0.3.0.** Requirements (v0.2.1 → v0.3.0, document order): 001-011 unchanged;
021 → 012; 022 → 013; 012 → 014; 013 → 015; 014 → 016; 023 → 017; 015 → 018; 016 → 019;
017 → 020; 018 → 021; 019 → 022; 020 → 023; 024 and 025 moved to card t1245 (spec.md §K).
Criteria (v0.2.1 → v0.3.0): 001 → 001; 002 → 002 (rewritten for the two-layer resolver); new
003; 003 → 004; 004 → 005; 005 → 006; 006 → 007; 007 → 008; 008 → 009; 009 → 010; 010 → 011;
011 → 012; 012 → 013; 013 → 014; 014 → 015; 015 → 016; 016 → 017; 019 → 018; 017 → 019; new
020; 018 → 021; 020 → 022; 021 → 023; 022-025 moved to card t1245.

## §B — Activation and contract resolution

- **AC-AE-001** (maps REQ-AE-001) — **Given** `workflow.autonomy.mode` absent, and separately set to `guided`, `""`, and `bogus`, **When** the PreToolUse and PostToolUse hooks process a fixed fixture set of Write, Edit, and Bash inputs, **Then** every hook output equals the pre-change golden output byte for byte and no file appears under any `escalation/` directory. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestEscalationGuidedGolden` in `internal/hook`.
  - RED-now: L1, L2 — no mode key and no detector exist; the golden-comparison test does not exist.
  - Green path: M1 — golden captured in its own commit before the detector lands (verification-claim-integrity §2.3).
- **AC-AE-002** (maps REQ-AE-002) — **Given** mode `contract`, a queue fixture recording card `t9001` with `spec_id: SPEC-A-001` and card `t9002` with no `spec_id`, and a tree holding signed contracts for in-progress SPEC-A-001 and in-progress SPEC-B-001, all on branches named `WT-other`, **When** a tool call is processed from a subdirectory of worktree directories named `t9001`, `t9002`, and `t9003` (no queue record), **Then** `t9001` arms against SPEC-A-001's contract despite SPEC-B-001's signed contract being present, while `t9002` and `t9003` each write one `not-armed` audit line naming the failing layer, and no case consults the branch name. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestResolveContractByCardAndSpec` in `internal/escalation`.
  - RED-now: L2. Green path: M1.
- **AC-AE-003** (maps REQ-AE-002) — **Given** mode `contract`, card `t9004` whose queue record links `SPEC-C-001` (`status: completed`) with a signed `contract.yaml`, card `t9005` linking `SPEC-D-001` (`status: archived`) with a signed contract, and card `t9006` linking in-progress `SPEC-E-001` whose contract is unsigned, **When** a tool call is processed from each worktree, **Then** none arms, each writes one `not-armed` line naming the terminal status or the missing signature, and no escalation record is written for the completed or archived contracts. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestResolverIgnoresTerminalAndUnsignedContracts` in `internal/escalation`.
  - RED-now: L2. Green path: M1.
- **AC-AE-004** (maps REQ-AE-003, REQ-AE-021) — **Given** mode `contract` and a fixture that trips each of the ten classes, **When** the hooks run, **Then** no hook output carries a deny or ask decision that the same input did not carry under guided mode, no hook call waits on acknowledgement, the only files the detector writes are escalation records, the audit log, and the card state file, and no detector code path references the user question channel.
  - Test: `TestDetectorNeverAltersToolCall` in `internal/hook`.
  - RED-now: L2. Green path: M5 (all classes wired).
- **AC-AE-005** (maps REQ-AE-004) — **Given** mode `contract` and a resolved `contract.yaml` made unreadable (permission-denied fixture), **When** a write tool call is processed, **Then** the tool call proceeds, the audit log gains a `not-checked` line naming the fault, and no escalation record is written.
  - Test: `TestFaultIsNotChecked` in `internal/escalation`.
  - RED-now: L2. Green path: M2.

## §C — Contract classes

- **AC-AE-006** (maps REQ-AE-005) — **Given** a signed-valid contract recording `acceptance.sha256` and `acceptance.ac_count`, **When** a checkpoint runs after (a) one byte of `acceptance.md` changes so verify reports `acceptance_hash_mismatch`, (b) the AC count changes so verify reports `ac_count_mismatch`, and (c) an `ac_count_ambiguous` fixture, **Then** each writes one `acceptance-change` record citing recorded and measured values, and an unchanged file writes none. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestAcceptanceChangeTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M2.
- **AC-AE-007** (maps REQ-AE-006, REQ-AE-022) — **Given** a contract whose `invariants` holds the command-kind entry `go test ./internal/spec/...`, a `constitution:` entry, and a second command-kind entry that no tool call executes, **When** a Bash call with the first command exits 1 and a checkpoint then runs, **Then** one `invariant-violation` (`command`) record is written, the same command exiting 0 or a different command exiting 1 writes none, and the checkpoint lists both the `constitution:` entry and the unexecuted command entry under `not_observed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestInvariantCommandFailureTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-008** (maps REQ-AE-007) — **Given** a contract whose `invariants` contains `frozen-files`, a caller identity other than the harness learner, and `ownership.never: [internal/foo/secret/**]`, **When** Writes target `docs/CLAUDE.md` (base-name match), a Frozen-zone target file named by the zone registry fixture, and `internal/foo/secret/y.go`, **Then** each writes one `invariant-violation` (`frozen-file`) record, a contract without `frozen-files` writes none of these for the same Writes, and the existing harness-learner deny tests in `internal/hook` still pass unchanged. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestFrozenFileUnionTrips` in `internal/hook`.
  - RED-now: `checkHarnessFrozenZone` returns early for any non-harness-learner identity (`internal/hook/pre_tool.go:1237-1239`, research.md P2). Green path: M3.
- **AC-AE-009** (maps REQ-AE-008, REQ-AE-012) — **Given** a signed contract with `ownership.write: [internal/foo/**, .moai/specs/<ID>/**]` and `ownership.never: [internal/foo/secret/**]`, **When** Writes target `internal/bar/x.go`, `internal/foo/secret/y.go`, `.moai/specs/<ID>/contract.yaml`, and `.moai/specs/<ID>/acceptance.md`, **Then** four `ownership-move` records are written, the second naming the `never` line and the last two naming the post-signing immutability, while a Write to `internal/foo/z.go` writes none. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestOwnershipMoveTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-010** (maps REQ-AE-013, REQ-AE-022) — **Given** the AC-AE-009 contract plus `ownership.scratch: [tmp-build/**]`, a worktree directory named `t9001`, and every exemption root determined, **When** Writes target `.moai/reports/t9001/verdict.md`, `.moai/state/x.json`, a file under the OS temporary directory, a file under the session scratchpad, a file under the auto-memory store, `tmp-build/out.txt`, and an absolute path outside the worktree root covered by none of these, **Then** the first six write no record and the last writes one `ownership-move` record; and with the scratchpad root undeterminable, the same outside-root write writes no record and is listed under `not_observed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestOwnershipExemptionsAndOutsideRoot` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-011** (maps REQ-AE-009) — **Given** `new_api_detector: graph` and five Go fixture HEADs each adding exactly one of a new exported function in an existing file, a new package directory, a new CLI command registration, a new MCP tool name, or a new config yaml key, plus a sixth adding only an unexported function, **When** the on-demand checkpoint compares the card base to HEAD, **Then** each of the five writes a record whose addition kind and name match its fixture and the sixth writes none.
  - Test: `TestNewAPIAdditionsTrip` in `internal/escalation`.
  - RED-now: L2; `graph.FileAPI` has no commit parameter (research.md P5). Green path: M4.
- **AC-AE-012** (maps REQ-AE-009, REQ-AE-022) — **Given** (a) `new_api_detector: off` on the AC-AE-011 exported-function fixture, (b) `graph` on a Go fixture with no resolvable integration branch, and (c) `graph` on a Python fixture whose HEAD adds a new top-level module and a new command-line entry point, **When** the on-demand checkpoint runs, **Then** (a) writes no `new-architecture-or-api` record, (b) writes none and lists the unresolvable card base under `not_observed`, and (c) reports each addition under a sub-kind the extractor supports for that language or lists it under `not_observed`, with the CLI-verb, MCP-tool, and config-key sub-kinds each listed under `not_observed`.
  - Test: `TestNewAPINotObservedCases` in `internal/escalation`.
  - RED-now: L1, L2. Green path: M4.
- **AC-AE-013** (maps REQ-AE-010, REQ-AE-022) — **Given** recorded inputs (a) an `audit_multi` result with `disagreement_flag: true`, (b) a first verdict PASS and a FAIL verdict from the contract's `review.second_model`, (c) a local verification pass and a recorded CI failure on the same head, (d) an `audit_multi` result with `disagreement_flag: null`, and (e) a local verification pass with no CI verdict recorded, **When** a checkpoint runs, **Then** (a)-(c) each write a `contradictory-evidence` record, and (d) and (e) write none while listing the flag and the CI verdict under `not_observed` rather than as agreement. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestContradictoryEvidenceTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-014** (maps REQ-AE-011) — **Given** a contract whose `actions` contains `push-develop`, **When** Bash calls `git push origin main`, `git push origin v1.0.0`, `git tag v1.0.0`, `gh release create v1.0.0`, and one command the existing destructive-command denylist denies are processed at PreToolUse, **Then** five `irreversible-action` records are written before execution and the denylisted command's deny output is byte-identical to its output with the detector off, while `git push origin develop` writes none; and **Given** a contract whose `actions` lacks `push-develop`, `git push origin develop` writes one record. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestIrreversibleActionTrips` in `internal/hook`.
  - RED-now: L2; `Bash(git tag:*)` sits in the settings allow block (research.md P8), so nothing observes it today. Green path: M5.

## §D — Operational trips

- **AC-AE-015** (maps REQ-AE-014) — **Given** a signed contract with `budget.operations: 3`, **When** a fourth write-capable or Bash tool call is observed at PostToolUse, **Then** a `budget-exceeded` record names `operations` with observed 4 and limit 3; a contract that is `signed-invalid` only for `acceptance_hash_mismatch` still supplies its own budget; and a worktree with no resolved contract uses `workflow.autonomy.escalation.budget_default`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestBudgetExceededTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-016** (maps REQ-AE-015) — **Given** one failing command fingerprint, **When** it fails three consecutive times, **Then** one `same-diagnostic-repeat` record is written, while the sequence fail, fail, pass, fail writes none.
  - Test: `TestSameDiagnosticRepeatTrips` in `internal/escalation`.
  - RED-now: super-advisor E1 is a doctrine row only (research.md P9); L2. Green path: M5.
- **AC-AE-017** (maps REQ-AE-016) — **Given** a contract with `budget.audit_retries: 0` and one plan-audit verdict file recording FAIL, and a contract with `budget.audit_retries: 2` and sync-audit verdict files recording FAIL at iterations 1 and then 3, **When** a checkpoint runs after each file, **Then** the first case writes one `audit-fail-at-retry-cap` record, and the second writes none after iteration 1 and one after iteration 3. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestAuditFailAtRetryCapTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-018** (maps REQ-AE-017) — **Given** card `t9001` armed against a contract recorded as signed-valid in `.moai/state/escalation/t9001.json`, **When** in three separate fixtures (a) its `signature` block is removed and a PreToolUse Write follows, (b) the file is deleted and a PreToolUse Write follows, and (c) a byte outside `signature` changes so verify reports `contract_digest_mismatch` and a commit checkpoint runs, **Then** each fixture writes exactly one `contract-void` record naming the condition and stating that classes 2-6 are disarmed, that record is written before the resolver's `not-armed` line in (a) and (b), the audit log records the disarming, a later out-of-scope Write writes no `ownership-move` record, and an over-budget operation still writes a `budget-exceeded` record. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestContractVoidPrecedesNotArmed` in `internal/escalation`.
  - RED-now: L2. Green path: M5.

## §E — Reporting and marking

- **AC-AE-019** (maps REQ-AE-018) — **Given** any tripped class for card `t9001`, **When** the record is written, **Then** it lands at `.moai/reports/t9001/escalation/<class>-<fingerprint>.md`, the queue store file's bytes are unchanged, `internal/kanban/backlog_schema_freeze_test.go` still passes, and the card lists as needs-decision exactly while a `contract` or `operational` record has `status: open`.
  - Test: `TestRecordPathAndQueueUntouched` in `internal/escalation`.
  - RED-now: L4 — no needs-decision surface exists. Green path: M1 (path) and M2 (writer).
- **AC-AE-020** (maps REQ-AE-018, REQ-AE-019) — **Given** one written `contract` record, one written `operational` record, and one hand-authored `revoke` record with class `revoke-operator`, **When** a YAML parser reads each file's frontmatter, **Then** every file parses, carries every spec.md §I.1 field with its stated type, has a `fingerprint` equal to the one in its file name and a `status` in {`open`, `resolved`}, the contract record's `contract_ref` matches `contract.yaml:<line>` pointing at the tripped line, the operational record's matches `config:` or `verify:`, the body holds `## Observation`, `## Options` with at least two items, and `## Not observed` in that order, and the needs-decision computation counts the revoke record as not open.
  - Test: `TestRecordFrontmatterParses` in `internal/escalation`.
  - RED-now: L2. Green path: M2.
- **AC-AE-021** (maps REQ-AE-019, REQ-AE-020) — **Given** a tripped contract class, **When** the same observation trips a second time while the record is open, and a third time after the record has been set to `status: resolved` with a `decider`, **Then** the second trip leaves exactly one file whose `occurrences` is 2, the third writes `<class>-<fingerprint>-2.md` as a new open record, and the resolved record's `status` and `decider` are unchanged. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestRecordDedupAndRetripAfterResolve` in `internal/escalation`.
  - RED-now: L2. Green path: M2.

## §F — Not-observed coverage and first-resolution invalidity

- **AC-AE-022** (maps REQ-AE-022) — **Given** a signed contract whose `ownership` block cannot be decoded into globs while the rest verifies, **When** a Write is processed, **Then** no `ownership-move` record is written and the audit log and next checkpoint output list `ownership` under `not_observed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestUnreadableContractFieldIsNotObserved` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-023** (maps REQ-AE-023) — **Given** a card whose only surviving contract is `signed-invalid` with reason `plan_audit_not_passing` at its first observation (no signed-valid state in the card state file), **When** an out-of-scope Write and an over-budget operation occur, **Then** no `ownership-move` record and no `contract-void` record are written, the audit log gains a `not-armed` line and a warning line carrying `plan_audit_not_passing`, and a `budget-exceeded` record is still written against `budget_default`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestFirstResolutionInvalidNotArmed` in `internal/escalation`.
  - RED-now: L2. Green path: M2.

## §G — Quality gates and Definition of Done

- Change-scoped tests pass: `go test ./internal/hook/... ./internal/config/... ./internal/kanban/... ./internal/escalation/...`, each with a non-empty swept count and every test named above present as `--- PASS`.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `golangci-lint run` reports no new issue against the pre-change baseline.
- Subagent boundary: no `AskUserQuestion` reference in the detector package outside tests.
- Template neutrality: any added template text contains no card id, SPEC id, internal date, or SHA.
- Coverage of the detector package is at least 85%.
- Every row of spec.md §F.1 and every criterion tagged 「A1 plan-audit 통과본으로 재확인」 is re-checked against the plan-audit-passed A1 schema before M2.
