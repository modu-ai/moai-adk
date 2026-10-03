---
id: SPEC-FACTORY-DECISION-AUTO-001
title: "Acceptance criteria — AC matrix, scenarios, quality gates"
version: "0.1.0"
created: 2026-10-03
---

# acceptance.md — SPEC-FACTORY-DECISION-AUTO-001

Verification discipline: every release-blocking AC names its RED-now cell (a probe from
research.md §2, measured at `d7112d005` before any implementation commit) and the milestone that
flips it green. Regression-guard ACs assert a property that holds at base and must still hold at
close.

## §D AC Matrix

| AC | Asserts | REQ | Class | RED-now | Green |
|---|---|---|---|---|---|
| AC-FDA-001 | `moai decision record` writes to the HOME-surface board for the primary checkout's project key; the same record is read from two different linked worktrees; no board file is created inside any tree | REQ-FDA-001 | release-blocking | P1 (no command) | M1 |
| AC-FDA-002 | A record lacking any required field, or carrying a kind outside the closed enum, or a `standing` record without a predicate, is refused with a non-zero exit and nothing is appended | REQ-FDA-002 | release-blocking | P1 | M1 |
| AC-FDA-003 | Under lane refusal `record` exits non-zero before any file I/O (board mtime unchanged); `read` succeeds | REQ-FDA-003 | release-blocking | P1 | M1 |
| AC-FDA-004 | `read --scope card:X` returns X's and standing records, hides superseded ones (shown with `--all`), and prints `board=absent`, `board=empty`, and `unparseable=<n>` statuses for the three fixtures | REQ-FDA-004 | release-blocking | P1 | M1 |
| AC-FDA-005 | The watchdog skill (local + template) instructs reading `moai decision read --scope card:<id>` at step ②, before step ③ | REQ-FDA-005 | release-blocking | P2 (no carrier) | M6 |
| AC-FDA-006 | auto-semantics (local + template) states record-before-message, message carries only the record id, and no re-send of a ruling a standing record governs | REQ-FDA-006 | release-blocking | P20 context | M6 |
| AC-FDA-007 | manager-spec (C1, C2, emitted C3) admits pinned `board:`/`mission:` anchors; a fixture row whose pinned line digest mismatches routes to FOUNDER | REQ-FDA-007 | release-blocking | P12 | M5 |
| AC-FDA-008 | plan-auditor (C1, C2, C3) states the four-part PASS-WITH-DEBT emission rule and the verdict block carries `blocking_findings` and `debts` | REQ-FDA-008 | release-blocking | P5 | M2 |
| AC-FDA-009 | §9.1 admits PASS-family per the predicate; §9.2 blocked list no longer lists bare PASS-WITH-DEBT and lists "PASS-WITH-DEBT without enumerated debts" | REQ-FDA-009 | release-blocking | P4 | M2 |
| AC-FDA-010 | A table-driven test runs one fixture set through all three sites (contract rule, kickoff decide, card-transition guard): well-formed PASS-WITH-DEBT admitted at all three; missing-debts and blocking>0 rejected at all three; FAIL/INCONCLUSIVE/BYPASSED/absent rejected at all three | REQ-FDA-011 | release-blocking | P3 (label-only admission) | M2 |
| AC-FDA-011 | `harness.plan_audit_ceiling_policy` exists with defaults `auto_delta_rounds: 1`, `on_second_hit: hold-and-split` (local + template); plan-auditor and spec-workflow no longer state a numeric cap other than citing the tier map | REQ-FDA-012 | release-blocking | P9, P10, P11 | M4 |
| AC-FDA-012 | Given a ceiling-hit verdict with `delta_eligible: true`, the documented lane procedure runs exactly one delta round and writes one decision record citing the policy key, with no question emitted | REQ-FDA-013 | release-blocking | P10 (routes to user) | M4 |
| AC-FDA-013 | Given a second ceiling hit, `delta_eligible: false`/absent, or a STOP, the procedure writes a hold wait record and a split proposal to the card evidence path and creates no card | REQ-FDA-014 | release-blocking | P10 | M4 |
| AC-FDA-014 | Kickoff approve with decider `audit` succeeds on a passing verdict with matching hash; fails on hash mismatch; fails on a FAIL verdict; decider `foo` rejected | REQ-FDA-015 | release-blocking | P6, P7 | M3 |
| AC-FDA-015 | After an `audit` approval the card is in the run stage with lease id and owner unchanged; a `human` approval still lands in `assigned` (T8 behavior preserved) | REQ-FDA-016 | release-blocking | P7 | M3 |
| AC-FDA-016 | Under lane refusal, `factory decide <own-leased-card> --gate kickoff --choice approve --decider audit` succeeds; the same call for a card the lane does not lease, any `human` call, and any push/abandon/resume call are refused | REQ-FDA-017 | release-blocking | P8 | M3 |
| AC-FDA-017 | manager-spec (C1, C2, C3) defines `Class:`, `Default:`, `Alternate:`, the reversibility rule, and "no Class → product-level" | REQ-FDA-018 | release-blocking | P12 | M5 |
| AC-FDA-018 | The kickoff step (plan workflow, local + template) fills `DEFAULT-APPLIED <UTC> <runner+role>` on implementation-level rows with a Default and blocks on a product-level row with an empty verdict | REQ-FDA-019 | release-blocking | P13 (gate on, no default path) | M5 |
| AC-FDA-019 | Run entry instructions copy `debts` into progress §E.2 `### Binding run conditions`; sync-auditor (C1, C2, C3) reports each as disposed/undisposed and an undisposed one as a finding | REQ-FDA-010 | release-blocking | P5 | M6 |
| AC-FDA-020 | Hook test: a second prompt with a matching bind cache performs zero `factorymsg.Open` calls (counted via the existing measurement seams); any field mismatch performs the full bind | REQ-FDA-021 | release-blocking | P14 | M7 |
| AC-FDA-021 | Hook test: a degraded inbox state produces one warn log line per occurrence and at most one session notice per configured interval | REQ-FDA-022 | release-blocking | P15, P16 | M7 |
| AC-FDA-022 | auto-semantics §5.1 and the watchdog skill (local + template) define the short recheck carrier (arm on wait-on-leader, delete on resolution, configurable cadence, codex named gap) | REQ-FDA-020 | release-blocking | P18 | M6 |
| AC-FDA-023 | Lane intake doctrine (kanban-dispatch, local + template) requires the MCP build vs `moai version` comparison and the `fallback=CLI` progress line | REQ-FDA-023 | release-blocking | P19 | M6 |
| AC-FDA-024 | Keep-set guard: no new code path records or approves push, release/main integration, queue admission, contract signing, abandon of an unintegrated branch, product-level verdicts, or final PASS/FAIL; `factory decide --gate push` still requires `human` | REQ-FDA-024 | regression-guard | holds at base (P8) | preserve-through-close |
| AC-FDA-025 | Every changed shipped file has its template mirror change (token presence in both trees); `make agents-emit-check` clean; no SPEC ID/card id/date in template diffs | REQ-FDA-025 | release-blocking | — | M8 |

## Scenarios (Given-When-Then)

### AC-FDA-003 — lane cannot write the board
- **Given** a session with `MOAI_FACTORY_ROLE=lane` and an existing board file with mtime T
- **When** it runs `moai decision record --scope standing --kind standing-rule --body "x" --evidence y`
- **Then** the exit code is non-zero, the board mtime is still T, and `moai decision read` from the same session exits 0

### AC-FDA-010 — one predicate, three sites
- **Given** verdict fixtures {PASS ok, PASS-WITH-DEBT with 2 debts, PASS-WITH-DEBT with no `debts`, PASS-WITH-DEBT blocking=1, FAIL, INCONCLUSIVE, BYPASSED, absent}
- **When** each fixture is evaluated by the contract rule, the kickoff evaluator, and the card-transition guard
- **Then** the admit/reject outcome is identical across the three sites for every fixture, and only the first two admit

### AC-FDA-012 — automatic delta round
- **Given** a Tier M SPEC at iteration 2 (the tier ceiling) with verdict FAIL and `delta_eligible: true`
- **When** the lane follows the ceiling procedure
- **Then** exactly one further audit runs, one decision record citing `harness.plan_audit_ceiling_policy` is written to the card progress record, and no question is emitted to anyone

### AC-FDA-013 — second hit
- **Given** the delta round of AC-FDA-012 also ends without a PASS-family verdict
- **When** the lane follows the ceiling procedure
- **Then** a `wait record: waiting_on=leader reason=ceiling-second-hit` line and a split proposal are written to the card evidence path, plan iteration stops, and the queue is unchanged

### AC-FDA-015 — lease kept
- **Given** a card in `kickoff` leased by worker-3 with a passing verdict whose hash matches
- **When** `factory decide <card> --gate kickoff --choice approve --decider audit` runs from worker-3
- **Then** the card's stage is `run`, the lease id and owner equal their prior values, and no re-lease is required

### AC-FDA-018 — FOUNDER default
- **Given** `decision_gate: on` and a decision-index with one implementation-level row with a Default and one product-level row, both with empty verdicts
- **When** the Kickoff step runs
- **Then** the first row reads `Operator verdict: DEFAULT-APPLIED <UTC> <runner+role>` and the autonomous Kickoff is blocked by the second row, which is routed to the operator

### AC-FDA-020 — bind cache hit
- **Given** a bound lane session with a cache file matching session, run, PID, and process start
- **When** the next prompt-submit hook runs
- **Then** the open-call counter seam records zero opens for the bind

## Edge cases

- Board file present but every line unparseable → `board=ok unparseable=<n>` with zero records, exit 0, warn on stderr.
- A `supersedes` reference to an unknown id → refused at record time.
- Verdict hash present but plan artifact deleted → audit decider refused (hash cannot be recomputed).
- Wait-on-leader resolved while the short carrier is mid-fire → the fire's watchdog pass sees the resolving record and deletes the carrier.
- Codex runner → no short carrier; progress records the named gap.

## Quality gates

- Changed Go packages: `go test -timeout 30m ./internal/<pkg>/...`, `go test -race` on `internal/hook` and the board package, `go vet`, `golangci-lint run` (CI version).
- Coverage ≥ 85% on new packages (`quality.yaml` target).
- `make build`, `make agents-emit`, `make agents-emit-check` clean.
- `moai spec lint` clean on this SPEC.

## Definition of Done

All 24 release-blocking ACs green with verbatim evidence in progress §E.2; AC-FDA-024 re-asserted at
close; sync-audit PASS-family; template mirrors and emitted agents committed; no keep-set behavior
changed.
