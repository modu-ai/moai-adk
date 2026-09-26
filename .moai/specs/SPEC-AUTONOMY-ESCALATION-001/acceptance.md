# acceptance.md — SPEC-AUTONOMY-ESCALATION-001

25 acceptance criteria (AC-AE-001 … AC-AE-025), Given-When-Then, each binary. Requirements
stay in spec.md §D; this file is the verification layer only. Each criterion is one line so
the lint parser reads its REQ mapping; its test, RED-now, and green-path cells follow as
sub-bullets.

## §A — Tree pin, evidence ledger, test naming, and renumbering

Document-level pin: every RED-now cell below was measured on commit
`ca1d5dc435ae1fa28aed58add241f73c5d8b2ae0` (branch `WT-escalation-detector`, base `develop`)
unless a row carries its own pin. Commits after it on this branch touch only this SPEC
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

**Contract fixtures and the `card` field.** Criteria whose fixtures carry a `card:` field use the
field A1 defines at `65e0a9167` (`design.md:22`, `:124`, § Card Field `:148-163`; spec.md §F.2 R9,
closed). M1 pre-flight re-checks the field name and pattern against the A1 that lands.

**Test naming (binding).** Each criterion names the Go test that verifies it and its package;
a criterion naming two tests is GREEN only when both pass. Detector unit tests live in
`internal/escalation` (proposed package, design.md §F); hook-integration tests live in
`internal/hook`. If run-phase renames the package, the new path is recorded in progress.md §E.2
and the test names stay unchanged. A criterion is GREEN only when
`go test <package> -run '^<TestName>$' -v` prints `--- PASS: <TestName>`; a run printing
`[no tests to run]` is an empty sweep and counts as RED, never as PASS. `<contract store>` in a
criterion means `$MOAI_HOME/db/<project-key>/contract/` with `$MOAI_HOME` pointed at a
per-test temporary directory, so no fixture touches the operator's real store — which places the
store under the OS temporary directory, the case REQ-AE-013 judges before any exemption. `<card
log>` means `<contract store>/escalation/<card-id>.log.jsonl` and `<card state>` means
`<contract store>/escalation/<card-id>.json`.

**Renumbering at v0.4.2.** Requirements keep their numbers. Criteria (v0.4.1 → v0.4.2):
001-010 keep their numbers (010 gains the store-under-temp case); 011 and 012 merge into 011;
013 → 012; 014 → 013; 015 → 014; 016 and 017 merge into 015; 018 → 016; 019 → 017 (its
state-file case moves to the new 018); new 018 (tamper: arming value removed while the log says
armed, and a naive rewrite); new 019 (another card's log appends leave this card untouched); new
020 (missing card log); 020 → 021; 021 → 022; 022 → 023; 023 → 024; 024 → 025. Earlier
renumbering is recorded in the commits that made it (`8c9ee29b7`, `fe118c4d1`).

## §B — Activation and contract resolution

- **AC-AE-001** (maps REQ-AE-001) — **Given** `workflow.autonomy.mode` absent, and separately set to `guided`, `""`, and `bogus`, **When** the PreToolUse and PostToolUse hooks process a fixed fixture set of Write, Edit, and Bash inputs, **Then** every hook output equals the pre-change golden output byte for byte and no file appears under any `escalation/` directory. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestEscalationGuidedGolden` in `internal/hook`.
  - RED-now: L1, L2 — no mode key and no detector exist; the golden-comparison test does not exist.
  - Green path: M1 — golden captured in its own commit before the detector lands (verification-claim-integrity §2.3).
- **AC-AE-002** (maps REQ-AE-002) — **Given** mode `contract`, a tree holding signed-valid contracts for in-progress `SPEC-A-001` (`card: t9001`) and in-progress `SPEC-B-001` (`card: t9002`), every worktree on a branch named `WT-other`, and a queue fixture in which no card carries a `spec_id`, **When** a tool call is processed from a subdirectory of worktree directories named `t9001` and `t9002`, **Then** `t9001` arms against SPEC-A-001's contract and `t9002` against SPEC-B-001's, neither resolution changes when the queue fixture is given `spec_id: SPEC-B-001` for `t9001` or when the branch is renamed `WT-t9002`, and no resolution reads the queue file. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestResolveContractByCardField` in `internal/escalation`.
  - RED-now: L2. Green path: M1.
- **AC-AE-003** (maps REQ-AE-002) — **Given** mode `contract` and worktrees named `t9003` (no contract carries `card: t9003`), `t9004` (two in-progress SPECs both carry `card: t9004`), `t9005` (its only contract belongs to a SPEC with `status: completed`), `t9006` (`status: "completed"`, quoted), `t9007` (`status: archived`), and `t9008` (its only contract is unsigned), **When** a tool call is processed from each, **Then** none arms, each writes exactly one `not-armed` line naming its cause, `t9004` additionally writes one warning line naming both SPEC IDs, and no escalation record is written for any of them. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestResolverNotArmedCases` in `internal/escalation`.
  - RED-now: L2. Green path: M1.
- **AC-AE-004** (maps REQ-AE-003, REQ-AE-021) — **Given** mode `contract` and a fixture that trips each of the ten classes, **When** the hooks run, **Then** no hook output carries a deny or ask decision that the same input did not carry under guided mode, no hook call waits on acknowledgement, the only files the detector writes are escalation records, the card audit log, and the card state file, and no detector code path references the user question channel.
  - Test: `TestDetectorNeverAltersToolCall` in `internal/hook`.
  - RED-now: L2. Green path: M5 (all classes wired).
- **AC-AE-005** (maps REQ-AE-004) — **Given** mode `contract` and the resolved `contract.yaml` made unreadable (permission-denied fixture) on a card never armed, **When** a write tool call is processed, **Then** the tool call proceeds, the card log gains a `not-checked` line naming the fault, and no escalation record is written.
  - Test: `TestFaultIsNotChecked` in `internal/escalation`.
  - RED-now: L2. Green path: M2.

## §C — Contract classes

- **AC-AE-006** (maps REQ-AE-005) — **Given** card `t9001` armed against a signed-valid contract recording `acceptance.sha256` and `acceptance.ac_count`, **When** in four separate fixtures a commit checkpoint runs after (a) one byte of `acceptance.md` changes, (b) the AC count changes, (c) an `ac_count_ambiguous` fixture, and (d) `acceptance.md` is deleted, all with `contract.yaml` bytes unchanged, **Then** each fixture writes one `acceptance-change` record citing recorded and measured values and naming every acceptance reason verify reported — `signature_acceptance_mismatch` among them — plus exactly one `detection-disarmed` record with `contract_ref: disarm:signature-invalid`, while an unchanged file writes neither; and **Given** the same edits made to an unsigned contract's SPEC during plan phase, **Then** no record of any class is written. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestAcceptanceChangeTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M2.
- **AC-AE-007** (maps REQ-AE-006, REQ-AE-022) — **Given** a contract whose `invariants` holds the command-kind entry `go test ./internal/spec/...`, a `constitution:` entry, and a second command-kind entry that no tool call executes, **When** a Bash call with the first command exits 1 and a checkpoint then runs, **Then** one `invariant-violation` (`command`) record is written, the same command exiting 0 or a different command exiting 1 writes none, and the checkpoint lists both the `constitution:` entry and the unexecuted command entry under `not_observed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestInvariantCommandFailureTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-008** (maps REQ-AE-007) — **Given** a contract whose `invariants` contains `frozen-files`, a caller identity other than the harness learner, `ownership.never: [internal/foo/secret/**]`, and a registry fixture naming one Frozen-zone target file, **When** Writes target `docs/CLAUDE.md` (a `**/CLAUDE.md` match), that Frozen-zone target file, and `internal/foo/secret/y.go`, **Then** each writes one `invariant-violation` (`frozen-file`) record whose glob comes from the A1-derived `frozen_files` cached at arming, a contract without `frozen-files` writes none of these for the same Writes, and the existing harness-learner deny tests in `internal/hook` still pass unchanged. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestFrozenFileUnionTrips` in `internal/hook`.
  - RED-now: `checkHarnessFrozenZone` returns early for any non-harness-learner identity (`internal/hook/pre_tool.go:1237-1239`, research.md P2). Green path: M3.
- **AC-AE-009** (maps REQ-AE-008, REQ-AE-012) — **Given** a signed contract with `ownership.write: [internal/foo/**, .moai/specs/<ID>/**]` and `ownership.never: [internal/foo/secret/**]`, **When** Writes target `internal/bar/x.go`, `internal/foo/secret/y.go`, `.moai/specs/<ID>/contract.yaml`, and `.moai/specs/<ID>/acceptance.md`, **Then** four `ownership-move` records are written, the second naming the `never` line and the last two naming the post-signing immutability from `effective_never`, while a Write to `internal/foo/z.go` writes none. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestOwnershipMoveTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-010** (maps REQ-AE-013, REQ-AE-022) — **Given** the AC-AE-009 contract plus `ownership.scratch: [tmp-build/**]`, a worktree directory named `t9001`, every exemption root determined, and the contract store placed under the OS temporary directory, **When** Writes target `.moai/reports/t9001/verdict.md`, `.moai/state/x.json`, a file under the OS temporary directory outside the store, a file under the session scratchpad, a file under the auto-memory store, `tmp-build/out.txt`, `<card state>` for `t9001`, `<card log>` for `t9001`, and an absolute path outside the worktree root covered by none of these, **Then** the first six write no record and the last three each write one `ownership-move` record, the two contract-store writes being judged before the OS-temporary-directory exemption; and with the scratchpad root undeterminable, the same outside-root write writes no record and is listed under `not_observed` while a write to `<card state>` still writes one `ownership-move` record and is not listed under `not_observed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestOwnershipExemptionsAndOutsideRoot` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-011** (maps REQ-AE-009, REQ-AE-022) — **Given** `new_api_detector: graph` and five Go fixture HEADs each adding exactly one of a new exported function in an existing file, a new package directory, a new CLI command registration, a new MCP tool name, or a new config yaml key, a sixth adding only an unexported function, and the negative fixtures (a) `new_api_detector: off` on the exported-function fixture, (b) `graph` on a Go fixture with no resolvable integration branch, and (c) `graph` on a Python fixture whose HEAD adds a new top-level module and a new command-line entry point, **When** the on-demand checkpoint compares the card base to HEAD, **Then** each of the five writes a record whose addition kind and name match its fixture and the sixth writes none, (a) writes no `new-architecture-or-api` record, (b) writes none and lists the unresolvable card base under `not_observed`, and (c) reports each addition under a sub-kind the extractor supports for that language or lists it under `not_observed`, with the CLI-verb, MCP-tool, and config-key sub-kinds each listed under `not_observed`.
  - Tests: `TestNewAPIAdditionsTrip` and `TestNewAPINotObservedCases` in `internal/escalation`.
  - RED-now: L1, L2; `graph.FileAPI` has no commit parameter (research.md P5). Green path: M4.
- **AC-AE-012** (maps REQ-AE-010, REQ-AE-022) — **Given** recorded inputs (a) an `audit_multi` result with `disagreement_flag: true`, (b) a first verdict PASS and a FAIL verdict from the contract's `review.second_model`, (c) a local verification pass and a recorded CI failure on the same head, (d) an `audit_multi` result with `disagreement_flag: null`, and (e) a local verification pass with no CI verdict recorded, **When** a checkpoint runs, **Then** (a)-(c) each write a `contradictory-evidence` record, and (d) and (e) write none while listing the flag and the CI verdict under `not_observed` rather than as agreement. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestContradictoryEvidenceTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-013** (maps REQ-AE-011) — **Given** a contract whose `actions` contains `push-develop`, **When** Bash calls `git push origin main`, `git push origin v1.0.0`, `git tag v1.0.0`, `gh release create v1.0.0`, and one command the existing destructive-command denylist denies are processed at PreToolUse, **Then** five `irreversible-action` records are written before execution and the denylisted command's deny output is byte-identical to its output with the detector off, while `git push origin develop` writes none; and **Given** a contract whose `actions` lacks `push-develop`, `git push origin develop` writes one record. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestIrreversibleActionTrips` in `internal/hook`.
  - RED-now: L2; `Bash(git tag:*)` sits in the settings allow block (research.md P8), so nothing observes it today. Green path: M5.

## §D — Operational trips and disarming

- **AC-AE-014** (maps REQ-AE-014) — **Given** a signed contract with `budget.operations: 3`, **When** a fourth write-capable or Bash tool call is observed at PostToolUse, **Then** a `budget-exceeded` record names `operations` with observed 4 and limit 3; after that card is disarmed a further over-budget operation still counts against the budget recorded in the card state file at arming; and a worktree with no resolved contract uses `workflow.autonomy.escalation.budget_default`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestBudgetExceededTrips` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-015** (maps REQ-AE-015, REQ-AE-016) — **Given** one failing command fingerprint, a contract with `budget.audit_retries: 0` and one plan-audit verdict file recording FAIL, and a contract with `budget.audit_retries: 2` and sync-audit verdict files recording FAIL at iterations 1 and then 3, **When** the command fails three consecutive times and a checkpoint runs after each verdict file, **Then** one `same-diagnostic-repeat` record is written with `contract_ref: rule:same-diagnostic-3` while the sequence fail, fail, pass, fail writes none, the `audit_retries: 0` case writes one `audit-fail-at-retry-cap` record, and the `audit_retries: 2` case writes none after iteration 1 and one after iteration 3. 「A1 plan-audit 통과본으로 재확인」
  - Tests: `TestSameDiagnosticRepeatTrips` and `TestAuditFailAtRetryCapTrips` in `internal/escalation`.
  - RED-now: super-advisor E1 is a doctrine row only (research.md P9); L2. Green path: M5.
- **AC-AE-016** (maps REQ-AE-017) — **Given** card `t9001` armed, its `<card log>` ending in an `armed` entry, **When** in three separate fixtures (a) the contract's `signature` block is removed and a PreToolUse Write follows, (b) the contract file is deleted and a PreToolUse Write follows, and (c) a byte outside `signature` changes so verify reports `contract_digest_mismatch` and a commit checkpoint runs, **Then** each fixture writes exactly one `detection-disarmed` record whose `contract_ref` is `disarm:signature-invalid`, `disarm:contract-absent`, and `disarm:signature-invalid` respectively, `<card log>` gains exactly one `disarmed` entry written before any `not-armed` line, a later out-of-scope Write writes no `ownership-move` record, and an over-budget operation still writes a `budget-exceeded` record. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestDisarmContractLossWritesOneRecord` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-017** (maps REQ-AE-017) — **Given** card `t9001` armed as in AC-AE-016, **When** in three separate fixtures (a) the SPEC's `status` becomes `completed` and a checkpoint runs, (b) the contract's `card` field is changed to `t9002` (re-signed) and a Write follows, and (c) a second in-progress SPEC's contract gains `card: t9001` and a Write follows, **Then** each writes exactly one `detection-disarmed` record with `contract_ref` `disarm:terminal-status`, `disarm:card-mismatch`, and `disarm:card-mismatch` respectively, each followed by one `disarmed` entry in `<card log>`; and **When** in fixture (a) the contract file is then also deleted and a further Write follows, **Then** no new record and no new `disarmed` entry is written and that record's `occurrences` becomes 2. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestDisarmTransitions` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-018** (maps REQ-AE-017) — **Given** card `t9001` armed, `<card log>` for `t9001` showing it armed with a valid chain, **When** in three separate fixtures (a) the `armed` value is deleted from `<card state>` by a Bash command and a Write follows, (b) `<card state>` bytes are rewritten by a Bash command without a matching `state` entry and a Write follows, and (c) one line in the middle of `<card log>` is altered, **Then** each writes exactly one `detection-disarmed` record with `contract_ref: disarm:state-tamper` and one `disarmed` entry, and in (a) no path reads the card as disarmed before that record is written — the mutant control for a trigger read from the state file, which would write no record at all. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestStateTamperJudgedFromCardLog` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-019** (maps REQ-AE-017) — **Given** cards `t9001` and `t9002` both armed, each with its own `<card log>` and a valid chain, **When** hooks for `t9002` append fifty entries to `t9002`'s log — arming, state, not-armed, and one full disarm-and-rearm cycle — interleaved with hooks for `t9001`, **Then** `t9001`'s log chain verifies, `t9001` writes no `detection-disarmed` record and no `state-tamper` of any kind, `t9001`'s log contains no entry written on behalf of `t9002`, and the two log files have the names `t9001.log.jsonl` and `t9002.log.jsonl` exactly. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestOtherCardLogDoesNotAffectThisCard` in `internal/escalation`.
  - RED-now: L2. Green path: M5.
- **AC-AE-020** (maps REQ-AE-017) — **Given** a worktree `t9001` whose `<card log>` is absent, **When** in three separate fixtures a Write is processed with (a) no `<card state>` and no escalation record for `t9001`, (b) a `<card state>` carrying an `armed` value, and (c) no `<card state>` but a `detection-disarmed` record under `.moai/reports/t9001/escalation/`, **Then** (a) writes no `detection-disarmed` record and proceeds as a never-armed card (resolution and a `not-armed` or arming entry in a new log), while (b) and (c) each write exactly one `detection-disarmed` record with `contract_ref: disarm:state-tamper` and start a new `<card log>` whose first entry is `disarmed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestMissingCardLogReading` in `internal/escalation`.
  - RED-now: L2. Green path: M5.

## §E — Reporting and marking

- **AC-AE-021** (maps REQ-AE-018) — **Given** any tripped class for card `t9001`, **When** the record is written, **Then** it lands at `<worktree root>/.moai/reports/t9001/escalation/<class>-<fingerprint>.md`, the queue store file's bytes are unchanged, `internal/kanban/backlog_schema_freeze_test.go` still passes, and the card lists as needs-decision exactly while a `contract` or `operational` record has `status: open`.
  - Test: `TestRecordPathAndQueueUntouched` in `internal/escalation`.
  - RED-now: L4 — no needs-decision surface exists. Green path: M1 (path) and M2 (writer).
- **AC-AE-022** (maps REQ-AE-018, REQ-AE-019) — **Given** one written `contract` record, written `operational` records of classes 7, 8, 9, and 10, and one hand-authored `revoke` record with class `revoke-operator`, **When** a YAML parser reads each file's frontmatter, **Then** every file parses, carries every spec.md §I.1 field with its stated type, has a `fingerprint` equal to the one in its file name and a `status` in {`open`, `resolved`}, each `contract_ref` has the §I.1 form for its class (`contract.yaml:<line>` pointing at the tripped line; `config:` or the budget line; `rule:same-diagnostic-3`; the `budget.audit_retries` line; `disarm:<reason>`), the body holds `## Observation`, `## Options` with at least two items, and `## Not observed` in that order, and the needs-decision computation counts the revoke record as not open.
  - Test: `TestRecordFrontmatterParses` in `internal/escalation`.
  - RED-now: L2. Green path: M2.
- **AC-AE-023** (maps REQ-AE-019, REQ-AE-020) — **Given** a tripped contract class, and separately a `budget-exceeded` trip, **When** the same observation trips a second time while the record is open, a third time after the record has been set to `status: resolved` with a fixture `decider` string, and the budget trip recurs with a higher observed count, **Then** the second trip leaves exactly one file whose `occurrences` is 2, the third writes `<class>-<fingerprint>-2.md` as a new open record, the resolved record's `status` and `decider` bytes are unchanged, and the higher budget count increments the existing budget record rather than creating a second file. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestRecordDedupAndRetripAfterResolve` in `internal/escalation`.
  - RED-now: L2. Green path: M2.

## §F — Not-observed coverage and first observation

- **AC-AE-024** (maps REQ-AE-022) — **Given** a signed contract whose `ownership` block cannot be decoded into globs while the rest verifies, **When** a Write is processed, **Then** no `ownership-move` record is written and the card log and next checkpoint output list `ownership` under `not_observed`. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestUnreadableContractFieldIsNotObserved` in `internal/escalation`.
  - RED-now: L2. Green path: M3.
- **AC-AE-025** (maps REQ-AE-023) — **Given** a card never armed and no checkpoint run since the worktree was created, **When** the first hook event is a PreToolUse Write outside `ownership.write`, with (a) the only candidate contract `signed-invalid` for `plan_audit_not_passing` and (b) the only candidate `signed-valid`, **Then** in (a) no `ownership-move` and no `detection-disarmed` record is written, the card log gains a `not-armed` line and a warning line carrying `plan_audit_not_passing`, and a later over-budget operation writes a `budget-exceeded` record against `budget_default`; and in (b) that same PreToolUse call arms the card, appends an `armed` entry to the card log, writes the card state file, and writes one `ownership-move` record, with no subprocess started by the hook process. 「A1 plan-audit 통과본으로 재확인」
  - Test: `TestFirstObservationVerifiedAtPreToolUse` in `internal/hook`.
  - RED-now: L2. Green path: M2.

## §G — Quality gates and Definition of Done

- Change-scoped tests pass: `go test ./internal/hook/... ./internal/config/... ./internal/kanban/... ./internal/escalation/...`, each with a non-empty swept count and every test named above present as `--- PASS`.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `golangci-lint run` reports no new issue against the pre-change baseline.
- Subagent boundary: no `AskUserQuestion` reference in the detector package outside tests.
- Template neutrality: any added template text contains no card id, SPEC id, internal date, or SHA.
- Coverage of the detector package is at least 85%.
- Every row of spec.md §F.1 and every criterion tagged 「A1 plan-audit 통과본으로 재확인」 is re-checked against the A1 that lands (pinned here at `65e0a9167`, v0.5.1) before M2; R8, R9, and R10 are closed at that pin.
