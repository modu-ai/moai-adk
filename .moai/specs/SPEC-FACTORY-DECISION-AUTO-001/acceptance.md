---
id: SPEC-FACTORY-DECISION-AUTO-001
title: "Acceptance criteria — AC matrix, scenarios, quality gates"
version: "0.2.0"
created: 2026-10-03
---

# acceptance.md — SPEC-FACTORY-DECISION-AUTO-001

Verification discipline: every release-blocking AC names a RED-now cell — a probe in research.md §2
recorded with its command, verbatim stdout, and exit code, measured at tree
`ba2033d22abee6cf37e02fdee1241038d6cc7356` before any implementation commit — and the milestone
that flips it green. Each probe shows the absence (or the to-be-removed presence) the AC asserts.
Regression-guard (RG) ACs assert a property that holds at their baseline and must still hold at
close; the two M0-dependent ACs are RG with the M0 measurement commit as their baseline (leader
decision D1/D9). AC-FDA-NNN verifies REQ-FDA-NNN one-to-one.

## §D AC Matrix

| AC | Asserts | REQ | Class | RED-now | Green |
|---|---|---|---|---|---|
| AC-FDA-001 | `moai decision record` writes to the HOME-surface board for the primary checkout's project key; the same record is read from two linked worktrees; no board file appears inside any tree; under lane refusal `record` exits non-zero with the board mtime unchanged while `read` exits 0 | REQ-FDA-001 | release-blocking | P1 (exit=1; control P1c exit=0) | M1 |
| AC-FDA-002 | A record missing a required field, carrying a kind outside the closed enum, a `standing` record without a predicate, or a `supersedes`/`resolves` reference to an unknown id is refused with a non-zero exit and nothing appended; a valid record round-trips every field including `resolves` | REQ-FDA-002 | release-blocking | P2 (exit=1) | M1 |
| AC-FDA-003 | `read --scope card:X` returns X's and standing records, hides superseded ones (shown with `--all`), and prints `board=absent`, `board=empty`, and `unparseable=<n>` for the three fixtures | REQ-FDA-003 | release-blocking | P3 (exit=1) | M1 |
| AC-FDA-004 | The watchdog skill and auto-semantics (local + template) instruct `moai decision read --scope card:<id>` at step ② before step ③, and state record-before-message, id-only nudges, and no re-send of a ruling a standing record governs | REQ-FDA-004 | release-blocking | P4 (all four counts 0, exit=1) | M6 |
| AC-FDA-005 | manager-spec (C1, C2, emitted C3) admits pinned `board:`/`mission:` anchors; a fixture row whose pinned line digest mismatches routes to FOUNDER | REQ-FDA-005 | release-blocking | P5 (exit=1) | M5 |
| AC-FDA-006 | plan-auditor (C1, C2, C3) verdict block carries `must_pass_failed`, `blocking_findings`, `audited_sha`, `fix_scope`, and `debts` (id, description, phase), and states the four-part PASS-WITH-DEBT emission rule | REQ-FDA-006 | release-blocking | P6, P11 (exit=1 each) | M2 |
| AC-FDA-007 | §9.1 admits a plan verdict per the plan-phase predicate (PASS or PASS-WITH-DEBT); the §9.2 blocked list (local + template) no longer lists bare PASS-WITH-DEBT and lists "PASS-WITH-DEBT not admitted by the predicate" | REQ-FDA-007 | release-blocking | P7 (line to be removed present in both trees, exit=0) | M2 |
| AC-FDA-008 | Run-entry instructions copy `debts` into progress under `Binding run conditions`; `sync-audit-4dim.js` and sync-auditor (C1, C2, C3) each re-read them; a fixture with one undisposed condition yields sync verdict `FAIL` from both owners | REQ-FDA-008 | release-blocking | P8 (exit=1) | M6 |
| AC-FDA-009 | One table-driven test drives the contract rule, the kickoff evaluator, and the card-transition guard. Plan phase (T7 + Kickoff): admitted = PASS meeting all checks, PASS-WITH-DEBT with debts; refused = PASS-WITH-DEBT without debts, blocking=1, score below the tier threshold with a PASS label, `must_pass_failed`=1 with a PASS label, hash mismatch, FAIL, INCONCLUSIVE, BYPASSED, absent. Sync phase (T13): PASS and PASS-WITH-DEBT admitted as today, FAIL refused. Outcomes identical across sites per fixture; existing T7/T13 tests characterized before the change | REQ-FDA-009 | release-blocking | P9 (label-only checks, exit=0), P9b (T7 and T13 share the guard, exit=0) | M2 |
| AC-FDA-010 | `harness.plan_audit_ceiling_policy` exists with `auto_delta_rounds: 1`, `on_final_hit: hold-and-split` in both trees; plan-auditor and spec-workflow state no numeric cap other than citing the tier map, and state that the policy binds every session | REQ-FDA-010 | release-blocking | P10 (exit=1), P12 (cap text present, exit=0) | M4 |
| AC-FDA-011 | Fixtures over two SHAs: diff inside `fix_scope` + progress/decision-index/reports with identical REQ/AC id sets → eligible, delta runs, one decision record cites the policy key, no question emitted; diff touching an Out-of-Scope bullet outside `fix_scope` → ineligible; an AC id added → ineligible; `fix_scope` absent → final-hit path | REQ-FDA-011 | release-blocking | P11 (exit=1), P12 (exit=0) | M4 |
| AC-FDA-012 | Iteration = ceiling + `auto_delta_rounds` without admission, an ineligible delta, or a STOP each produce a hold wait record and a split proposal (card evidence path, or SPEC progress outside a card) and no card; a non-lane fixture additionally emits a user-facing notice and at most an override question | REQ-FDA-012 | release-blocking | P12 (user routing present, exit=0) | M4 |
| AC-FDA-013 | Exception fixtures: all three conditions + leader board record → hunk-limited fix and re-read confirmation, no full re-audit; each of {not on a release dependency path, `blocking_findings`=2, `defect_class` not `ac-wording`, `reread_hunks` absent, no board record} alone → hold stands | REQ-FDA-013 | release-blocking | P13 (exit=1) | M4 |
| AC-FDA-014 | Decider `audit` approve succeeds on an admitted verdict with matching `audited_sha` and hash; refused on each of: hash mismatch, `audited_sha` mismatch, FAIL verdict, audit-ready absent, open blocker, operator hold, open product-level FOUNDER row; decider `foo` rejected; after each refusal a `human` approval still works | REQ-FDA-014 | release-blocking | P14 (exit=1) | M3 |
| AC-FDA-015 | After an `audit` approval the card is in the run stage with lease id and owner unchanged; a `human` approval still lands in `assigned` (T8 preserved) | REQ-FDA-015 | release-blocking | P15 (only T8, exit=0) | M3 |
| AC-FDA-016 | Under lane refusal `factory decide <own-leased-card> --gate kickoff --choice approve --decider audit` succeeds; the same call for an unleased card, any `human` call, and any push/abandon/resume call are refused; SPEC-FACTORY-SELF-DISPATCH-001 carries the Amendments row | REQ-FDA-016 | release-blocking | P16 (human-only, exit=0) | M3 |
| AC-FDA-017 | manager-spec (C1, C2, C3) defines `Class:`, `Default:`, `Alternate:`, the published Default rule, "no Class → product-level", and the clause narrowed to judgment calls; the C2 commit precedes or equals the C1 commit | REQ-FDA-017 | release-blocking | P17 (exit=1) | M5 |
| AC-FDA-018 | The kickoff step (local + template) states the three-item product-level definition, fills `DEFAULT-APPLIED <UTC> <runner+role>` on implementation-level rows with a Default, and blocks on a product-level row with an empty verdict | REQ-FDA-018 | release-blocking | P18 (exit=1) | M5 |
| AC-FDA-019 | auto-semantics §5.1/§14 and the watchdog skill (local + template) define wait ids, `resolves`, the one-shot recheck re-armed per open wait, delay default 5 with a 5-minute floor, and the codex named gap; a fixture where an unrelated same-card board record leaves the wait open; the shipped default equals the M0(b) value (≥ 5) | REQ-FDA-019 | RG (baseline: M0(b) commit) | P19 (exit=1) | M6 |
| AC-FDA-020 | Hook test: a matching cache on a live run performs the run-state probe and zero `factorymsg.Open` and peer queries, and emits no degraded notice even when Open would exceed its budget; M0(a) records the under-load degraded-notice rate without the cache, and the post-M7 re-measurement shows none on already-bound sessions | REQ-FDA-020 | RG (baseline: M0(a) commit) | P20 (exit=1), P21 (probe precedes Open, exit=0) | M7 |
| AC-FDA-021 | Hook test: a cached session whose run is retired in the fixture DB rebinds on the next prompt (rebind notice, cache rewritten, no stale binding); a different run likewise; a probe failure on a hit surfaces the degraded state | REQ-FDA-021 | release-blocking | P20 (exit=1), P21 (exit=0) | M7 |
| AC-FDA-022 | Hook test: a degraded inbox state produces one warn log line per occurrence and at most one session notice per configured interval | REQ-FDA-022 | release-blocking | P22 (only warn is broker close, exit=0) | M7 |
| AC-FDA-023 | Lane intake doctrine (kanban-dispatch, local + template) requires the MCP build vs `moai version` comparison and the `fallback=CLI` progress line | REQ-FDA-023 | release-blocking | P23 (exit=1) | M6 |
| AC-FDA-024 | Keep-set guard: no new code path records or approves push, release/main integration, queue admission, contract signing, abandon of an unintegrated branch, product-level verdicts, or final PASS/FAIL; `factory decide --gate push` still requires `human` | REQ-FDA-024 | RG (baseline: `ba2033d22`) | P24 (push branch present, exit=0), P16 | preserve-through-close |
| AC-FDA-025 | Every changed shipped file has its template mirror change (token presence in both trees, e.g. `plan_audit_ceiling_policy` in both `harness.yaml`); `make agents-emit-check` clean; no SPEC ID, card id, or date in template diffs | REQ-FDA-025 | release-blocking | P10 (absent in both trees, exit=1) | M8 |

## Scenarios (Given-When-Then)

### AC-FDA-001 — lane cannot write the board
- **Given** a session with lane refusal holding and an existing board file with mtime T
- **When** it runs `moai decision record --scope standing --kind standing-rule --body "x" --evidence y`
- **Then** the exit code is non-zero, the board mtime is still T, and `moai decision read` from the same session exits 0

### AC-FDA-009 — one phase-scoped predicate, three sites
- **Given** the plan-phase and sync-phase fixture sets listed in the matrix
- **When** each fixture is evaluated by the contract rule, the kickoff evaluator, and the card-transition guard
- **Then** the admit/refuse outcome is identical across sites for every fixture, and a PASS-labelled verdict with score below the tier threshold or `must_pass_failed: 1` is refused at T7

### AC-FDA-011 — mechanical delta eligibility
- **Given** a Tier M SPEC at its ceiling (iteration 2) with verdict FAIL and `fix_scope: [spec.md#REQ-X, acceptance.md#AC-Y]`
- **When** the next commit changes only those anchors plus `progress.md`, and the REQ/AC id sets are unchanged
- **Then** exactly one delta audit runs and one decision record citing `harness.plan_audit_ceiling_policy` is written, with no question emitted; the same fixture with an edited Out-of-Scope bullet outside `fix_scope` takes the final-hit path

### AC-FDA-013 — release-blocking AC-wording exception
- **Given** a card on the dependency path of a card in an operator-approved release scope, a final verdict with `blocking_findings: 1`, `defect_class: ac-wording`, `reread_hunks: [acceptance.md#AC-Z]`, and a leader `card:` board record naming the exception
- **When** the ceiling procedure runs
- **Then** only the listed hunk changes and the auditor's re-read confirmation is recorded; removing any one of the three conditions or the board record leaves the hold in place

### AC-FDA-014 — audit decider refuses open keep-set conditions
- **Given** a card in `kickoff` with an admitted verdict and matching hashes, and a decision-index holding a product-level row with an empty verdict
- **When** `factory decide <card> --gate kickoff --choice approve --decider audit` runs
- **Then** the call is refused, the card stays in `kickoff`, and a subsequent `--decider human` approval succeeds

### AC-FDA-015 — lease kept
- **Given** a card in `kickoff` leased by worker-3 meeting every REQ-FDA-014 condition
- **When** the `audit` approval runs from worker-3
- **Then** the card's stage is `run`, the lease id and owner equal their prior values, and no re-lease is required

### AC-FDA-018 — FOUNDER default
- **Given** `decision_gate: on` and a decision-index with one implementation-level row with a Default and one product-level row, both with empty verdicts
- **When** the Kickoff step runs
- **Then** the first row reads `Operator verdict: DEFAULT-APPLIED <UTC> <runner+role>` and the autonomous Kickoff is blocked by the second row, which is routed to the operator

### AC-FDA-019 — unrelated record leaves the wait open
- **Given** a progress wait line `wait record: id=w1 waiting_on=leader ...` and a board record for the same card with no `resolves` field
- **When** the one-shot recheck fires
- **Then** the wait stays open and one further one-shot is armed; after a board record with `resolves: w1`, the next fire arms nothing

### AC-FDA-021 — retired run rebinds
- **Given** a bound session with a matching bind cache and the cached run marked retired in the fixture DB
- **When** the next prompt-submit hook runs
- **Then** the probe reports retired, the cache is invalidated, the rebind notice is emitted, and the hook context carries no binding to the retired run

## Edge cases

- A configured recheck delay below 5 minutes → clamped to 5 with a warning.
- Board file present but every line unparseable → `board=ok unparseable=<n>` with zero records, exit 0, warn on stderr.
- Verdict hash present but plan artifact deleted → audit decider refused (hash cannot be recomputed).
- Diff between two audited SHAs cannot be computed (SHA unreachable) → delta ineligible.
- Codex runner → no one-shot recheck; progress records the named gap.

## Quality gates

- Changed Go packages: `go test -timeout 30m ./internal/<pkg>/...`, `go test -race` on `internal/hook` and the board package, `go vet`, `golangci-lint run` (CI version).
- Coverage ≥ 85% on new packages (`quality.yaml` target).
- `make build`, `make agents-emit`, `make agents-emit-check` clean.
- `moai spec lint` clean on this SPEC.

## Definition of Done

The M0 baseline committed before the first implementation commit; all 22 release-blocking ACs green
with verbatim evidence in progress §E.2; the three RG ACs re-asserted against their baselines at close;
sync-audit admitted; template mirrors and emitted agents committed; no keep-set behavior changed.
