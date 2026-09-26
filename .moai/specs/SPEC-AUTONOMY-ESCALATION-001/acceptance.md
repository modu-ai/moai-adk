# acceptance.md — SPEC-AUTONOMY-ESCALATION-001

22 acceptance criteria (AC-AE-001 … AC-AE-022), Given-When-Then, each binary. Requirements
stay in spec.md §D; this file is the verification layer only. Each criterion is one line so
the lint parser reads its REQ mapping; its RED-now and green-path cells follow as sub-bullets.

## §A — Tree pin and evidence ledger

Document-level pin: every RED-now cell below was measured on commit
`ca1d5dc435ae1fa28aed58add241f73c5d8b2ae0` (branch `WT-escalation-detector`, base `develop`)
unless a criterion carries its own pin.

| Ledger id | Command (single invocation) | stdout | exit |
|---|---|---|---|
| L1 | `grep -rn autonomy internal/template/templates/.moai/config/sections/workflow.yaml` | (empty) | 1 |
| L2 | `grep -rln 'escalation-report\|EscalationRecord\|escalate_on' --include=*.go internal cmd pkg` | (empty) | 1 |
| L3 | `grep -rn escalate_on internal/template/templates` | (empty) | 1 |
| L4 | `grep -rn needs-decision --include=*.go internal` (positive control for the grep shape: `BacklogStatePicked` → 17 non-test lines) | (empty) | 1 |

The ledger rows establish that no detector, no config key, and no escalation record exists on
the pinned tree — which is why every criterion below is RED-now. A criterion verified by a new
Go test is GREEN only when the `-v` output contains its `--- PASS: <TestName>` line; a run that
prints `[no tests to run]` is an empty sweep and counts as RED, never as PASS.

## §B — Activation and inertness

- **AC-AE-001** (maps REQ-AE-001) — **Given** `workflow.autonomy.mode` absent from the config, **When** the PreToolUse and PostToolUse hooks process a fixed fixture set of Write, Edit, and Bash inputs, **Then** every hook output equals the pre-change golden output byte for byte and no file appears under the escalation record location.
  - RED-now: L1, L2 — no mode key and no detector exist; the golden-comparison test does not exist.
  - Green path: M1 — golden captured in its own commit before the detector lands (verification-claim-integrity §2.3).
- **AC-AE-002** (maps REQ-AE-001) — **Given** mode values `guided`, `""`, and `bogus`, **When** the AC-AE-001 fixture runs, **Then** all three produce outputs identical to the absent case.
  - RED-now: L1. Green path: M1.
- **AC-AE-003** (maps REQ-AE-002) — **Given** mode `contract` and two SPECs, one with no `contract.yaml` and one whose contract A1 verify reports `unsigned`, **When** a checkpoint fires for each, **Then** no escalation report is written and the audit log gains exactly one `not-armed` line per SPEC naming it and its reason. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M2.
- **AC-AE-004** (maps REQ-AE-003, REQ-AE-018) — **Given** mode `contract` and a fixture that trips each of the nine classes, **When** the hooks run, **Then** no hook output carries a deny or ask decision that the same input did not carry under guided mode, and no detector code path references the user question channel.
  - RED-now: L2. Green path: M5 (all classes wired).
- **AC-AE-005** (maps REQ-AE-004) — **Given** mode `contract` and an unreadable `contract.yaml` (permission-denied fixture), **When** a write tool call is processed, **Then** the tool call proceeds, the audit log gains a `not-checked` line naming the fault, and no escalation or needs-decision record is written.
  - RED-now: L2. Green path: M2.

## §C — Contract classes

- **AC-AE-006** (maps REQ-AE-005) — **Given** a signed-valid contract recording `acceptance.sha256` and `acceptance.ac_count`, **When** one byte of `acceptance.md` changes so that A1 verify reports `acceptance_hash_mismatch` and a checkpoint runs, **Then** one `acceptance-change` report is written citing recorded and measured hash, and with the file unchanged no report is written. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M2.
- **AC-AE-007** (maps REQ-AE-005) — **Given** an `acceptance.md` whose AC count differs from the contract's `acceptance.ac_count` so that A1 verify reports `ac_count_mismatch`, **When** a checkpoint runs, **Then** the report names both the recorded and the measured count, and an `ac_count_ambiguous` fixture likewise writes one `acceptance-change` report. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M2.
- **AC-AE-008** (maps REQ-AE-006) — **Given** a contract whose `invariants` holds the command-kind entry `go test ./internal/spec/...`, **When** a Bash call with that exact command exits 1, **Then** an `invariant-violation` (`command`) report is written, while the same command exiting 0, a different command exiting 1, and the `constitution:` and `frozen-files` entries each write none from this path. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M3.
- **AC-AE-009** (maps REQ-AE-007) — **Given** a contract whose `invariants` contains `frozen-files` and a caller identity other than the harness learner, **When** a Write targets `CLAUDE.md`, **Then** an `invariant-violation` (`frozen-file`) report is written, a contract without `frozen-files` writes none for the same Write, and the existing harness-learner deny tests in `internal/hook` still pass unchanged. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: `checkHarnessFrozenZone` returns early for any non-harness-learner identity (`internal/hook/pre_tool.go:1237-1239`, research.md P2). Green path: M3.
- **AC-AE-010** (maps REQ-AE-008) — **Given** `ownership.write: [internal/foo/**]` and `ownership.never: [internal/foo/secret/**]`, **When** Writes target `internal/bar/x.go` and `internal/foo/secret/y.go`, **Then** two `ownership-move` reports are written with the second naming the `never` line, while a Write to `internal/foo/z.go` writes none. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M3.
- **AC-AE-011** (maps REQ-AE-009) — **Given** `new_api_detector: graph` and a Go fixture repo whose HEAD adds one exported function to an existing file, **When** a checkpoint compares base to HEAD, **Then** the report lists that function by name and file, while adding only an unexported function writes none.
  - RED-now: L2; `graph.FileAPI` has no commit parameter (research.md P5). Green path: M4.
- **AC-AE-012** (maps REQ-AE-009) — **Given** four fixture HEADs each adding exactly one of a new package directory, a new CLI command registration, a new MCP tool name, or a new config yaml key, **When** a checkpoint runs, **Then** each produces a report whose addition kind matches its fixture.
  - RED-now: L2. Green path: M4.
- **AC-AE-013** (maps REQ-AE-009) — **Given** `new_api_detector: off`, **When** the AC-AE-011 fixture runs, **Then** no `new-architecture-or-api` report is written.
  - RED-now: L1. Green path: M4.
- **AC-AE-014** (maps REQ-AE-010, REQ-AE-019) — **Given** recorded inputs (a) an `audit_multi` result with `disagreement_flag: true`, (b) a first verdict PASS and a FAIL verdict from the contract's `review.second_model`, (c) a local verification pass and a recorded CI failure on the same head, **When** a checkpoint runs, **Then** each writes a `contradictory-evidence` report, while an `audit_multi` result with `disagreement_flag: null` writes no report and lists the flag under not-observed. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M5.
- **AC-AE-015** (maps REQ-AE-011) — **Given** a contract whose `actions` contains `push-develop`, **When** Bash calls `git push origin main`, `git push origin v1.0.0`, `git tag v1.0.0`, and `gh release create v1.0.0` are processed at PreToolUse, **Then** four `irreversible-action` reports are written before execution, while `git push origin develop` writes none. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2; `Bash(git tag:*)` sits in the settings allow block (research.md P8), so nothing observes it today. Green path: M5.

## §D — Operational trips

- **AC-AE-016** (maps REQ-AE-012) — **Given** a signed-valid contract with `budget.operations: 3`, **When** a fourth write tool call is observed, **Then** a `budget-exceeded` report names `operations` with observed 4 and limit 3, and for a SPEC with no signed-valid contract the `workflow.autonomy.escalation.budget_default` value applies. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M5.
- **AC-AE-017** (maps REQ-AE-013) — **Given** one failing command fingerprint, **When** it fails three consecutive times, **Then** one `same-diagnostic-repeat` report is written, while the sequence fail, fail, pass, fail writes none.
  - RED-now: super-advisor E1 is a doctrine row only (research.md P9); L2. Green path: M5.
- **AC-AE-018** (maps REQ-AE-014) — **Given** a Tier S SPEC (ceiling 1) with one plan-audit verdict file recording FAIL, **When** a checkpoint runs, **Then** an `audit-fail-at-retry-cap` report is written, while a Tier L SPEC (ceiling 3) with one FAIL writes none.
  - RED-now: L2. Green path: M5.

## §E — Reporting and marking

- **AC-AE-019** (maps REQ-AE-015) — **Given** any tripped class, **When** the report is written, **Then** the queue store file's bytes are unchanged, `internal/kanban/backlog_schema_freeze_test.go` still passes, and the card is listed as needs-decision by reading the escalation records.
  - RED-now: L4 — no needs-decision surface exists. Green path: M2.
- **AC-AE-020** (maps REQ-AE-016) — **Given** a tripped contract class, **When** the report is read, **Then** it carries class, observation with verbatim evidence, at least two options, `contract.yaml:<line>` plus the `escalate_on` token, SPEC ID, card id or `unresolved`, and HEAD SHA, with the config key in place of the contract line for an operational class. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M2.
- **AC-AE-021** (maps REQ-AE-017) — **Given** the same observation tripping twice, **When** both are processed, **Then** exactly one report exists and its occurrence count is 2.
  - RED-now: L2. Green path: M2.
- **AC-AE-022** (maps REQ-AE-020) — **Given** a contract that A1 verify reports `signed-invalid` with reason `contract_digest_mismatch`, **When** an out-of-scope Write and an over-budget operation occur, **Then** no `ownership-move` report is written, the audit log gains a `not-armed` line carrying `contract_digest_mismatch`, and a `budget-exceeded` report is still written against `budget_default`. 「A1 plan-audit 통과본으로 재확인」
  - RED-now: L2. Green path: M5.

## §F — Quality gates and Definition of Done

- Change-scoped tests pass: `go test ./internal/hook/... ./internal/config/... ./internal/kanban/...` plus the detector's own package, each with a non-empty swept count.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `golangci-lint run` reports no new issue against the pre-change baseline.
- Subagent boundary: no `AskUserQuestion` reference in the detector package outside tests.
- Template neutrality: any added template text contains no card id, SPEC id, internal date, or SHA.
- Coverage of the detector package is at least 85%.
- Every criterion tagged 「A1 plan-audit 통과본으로 재확인」 is re-checked against the plan-audit-passed A1 schema (drafted at `8f77d9a33`) before M2.
