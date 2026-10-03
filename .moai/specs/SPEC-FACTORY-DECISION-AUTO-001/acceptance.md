---
id: SPEC-FACTORY-DECISION-AUTO-001
title: "Acceptance criteria — AC matrix, scenarios, quality gates"
version: "0.3.0"
created: 2026-10-03
---

# acceptance.md — SPEC-FACTORY-DECISION-AUTO-001

## Verification discipline

Every release-blocking AC names a RED-now cell. The cell is a probe in research.md §2 recorded with
its command, verbatim stdout, and exit code, measured at tree
`5d094991fb586b9c4aadee33b663fe92b686ff18` before any implementation commit. It also names the
milestone that flips it green. Each probe shows today's behavior: the absence, or the presence to be
removed, that the AC asserts will change. The single regression-guard (RG) AC asserts a property that
holds at its baseline and must still hold at close. M0 measurements are evidence obligations
(§ Measurement notes), not AC cells. AC-FDA-NNN verifies REQ-FDA-NNN one-to-one.

## §D AC Matrix

| AC | Asserts | REQ | Class | RED-now | Green |
|---|---|---|---|---|---|
| AC-FDA-001 | `moai decision record` writes to the HOME-surface board for the primary checkout's project key; the same record is read from two linked worktrees; no board file appears inside any tree; under lane refusal `record` exits non-zero with the board mtime unchanged while `read` exits 0 | REQ-FDA-001 | release-blocking | P1 (exit=1; control P1c exit=0) | M1 |
| AC-FDA-002 | A record missing a required field, carrying a kind outside the closed enum, a `standing` record without a predicate, a `release-scope` record without a release id or card-id list, or a `supersedes`/`resolves` reference to an unknown id is refused with a non-zero exit and nothing appended; a valid record round-trips every field | REQ-FDA-002 | release-blocking | P2 (exit=1), P27 (exit=1) | M1 |
| AC-FDA-003 | `read --scope card:X` returns X's and standing records, hides superseded ones (shown with `--all`), and prints `board=absent`, `board=empty`, and `unparseable=<n>` for the three fixtures | REQ-FDA-003 | release-blocking | P3 (exit=1) | M1 |
| AC-FDA-004 | The watchdog skill and auto-semantics (local + template) instruct `moai decision read --scope card:<id>` at step ② before step ③, and state record-before-message, id-only nudges, and no re-send of a ruling a standing record governs | REQ-FDA-004 | release-blocking | P4 (all four counts 0, exit=1) | M6 |
| AC-FDA-005 | manager-spec (C1, C2, emitted C3) admits pinned `board:`/`mission:` anchors; a fixture row whose pinned line digest mismatches routes to FOUNDER | REQ-FDA-005 | release-blocking | P5 (exit=1) | M5 |
| AC-FDA-006 | plan-auditor (C1, C2, C3) verdict block carries `must_pass_failed`, `blocking_count`, `audited_sha`, `scope` (`full`/`delta`/`reread`), `fix_scope`, `defect_class` (closed enum incl. `ac-wording`, `design`), `reread_hunks`, and `debts` (id, description, `dispose_in`), and states the four-part PASS-WITH-DEBT emission rule | REQ-FDA-006 | release-blocking | P6, P11, P13 (exit=1 each) | M2 |
| AC-FDA-007 | §9.1 admits a plan verdict per the plan-phase predicate (PASS or PASS-WITH-DEBT); the §9.2 blocked list (local + template) no longer lists bare PASS-WITH-DEBT and lists "PASS-WITH-DEBT not admitted by the predicate" | REQ-FDA-007 | release-blocking | P7 (line to be removed present in both trees, exit=0) | M2 |
| AC-FDA-008 | Run-entry instructions copy `debts` into progress under `Binding run conditions`; `sync-audit-4dim.js` and sync-auditor (C1, C2, C3) each re-read them; a fixture with one undisposed condition yields sync verdict `FAIL` from both owners | REQ-FDA-008 | release-blocking | P8 (exit=1) | M6 |
| AC-FDA-009 | One table-driven test drives the contract rule, the kickoff evaluator, and the card-transition guard. Plan phase (T7 + Kickoff): admitted = PASS meeting all checks, PASS-WITH-DEBT with debts; refused = PASS-WITH-DEBT without debts, `blocking_count`=1, score below the tier threshold with a PASS label, `must_pass_failed`=1 with a PASS label, hash mismatch, FAIL, INCONCLUSIVE, BYPASSED, absent. Sync phase (T13): exactly today's label-only outcomes (PASS and PASS-WITH-DEBT admitted, FAIL/INCONCLUSIVE/BYPASSED/absent refused), asserted unchanged by characterization tests run before and after the change. Outcomes identical across sites per fixture | REQ-FDA-009 | release-blocking | P9 (label-only checks, exit=0), P9b (T7 and T13 share the guard, exit=0) | M2 |
| AC-FDA-010 | `harness.plan_audit_ceiling_policy` exists with `auto_delta_rounds: 1`, `on_final_hit: hold-and-split` in both trees; plan-auditor and spec-workflow state no numeric cap other than citing the tier map, and state that the policy binds every session | REQ-FDA-010 | release-blocking | P10 (exit=1), P12 (cap text present, exit=0) | M4 |
| AC-FDA-011 | Fixtures over two SHAs: diff inside `fix_scope` anchor ranges + progress/reports with identical REQ/AC id sets → eligible, delta runs, one decision record cites the policy key, no question emitted; diff touching an Out-of-Scope bullet outside `fix_scope` → ineligible; a `decision-index.md` change → ineligible; an AC id added → ineligible; `fix_scope` absent → final-hit path | REQ-FDA-011 | release-blocking | P11 (exit=1), P12 (exit=0) | M4 |
| AC-FDA-012 | Iteration = ceiling + `auto_delta_rounds` without admission, an ineligible delta, or a STOP each produce a hold wait record and a split proposal (card evidence path, or SPEC progress outside a card) and no card; a non-lane fixture additionally emits a user-facing notice, at most an override question, and labels the split proposal informational | REQ-FDA-012 | release-blocking | P12 (user routing present, exit=0) | M4 |
| AC-FDA-013 | Exception fixtures, positive: card listed in a `release-scope` record; card reached via `depends` from a listed card; card reached via `blocks` to a listed card — each with `blocking_count: 1`, `defect_class: ac-wording`, `reread_hunks`, and a leader `card:` record → only listed hunks change (diff check), the auditor emits a `scope: reread` verdict admitted by the plan-phase predicate, and a board record with `resolves: <hold-id>` releases the hold. Negative, each alone → hold stands: not reachable (only `contains`/`conflicts` edges), `blocking_count: 2`, `defect_class: design`, `reread_hunks` absent, no leader record, a diff outside `reread_hunks`, a `scope: reread` verdict the predicate refuses | REQ-FDA-013 | release-blocking | P13 (exit=1), P27 (exit=1) | M4 |
| AC-FDA-014 | Decider `audit` approve succeeds on an admitted verdict with matching `audited_sha` and hash; refused on each of: hash mismatch, `audited_sha` mismatch, FAIL verdict, audit-ready absent, open blocker, operator hold, a `decision-index.md` edited after the audited SHA (a product-level row reclassified to implementation-level + DEFAULT-APPLIED), a product-level FOUNDER row with an empty verdict, an implementation-level FOUNDER row with no Default and an empty verdict; decider `foo` rejected; after each refusal a `human` approval still works | REQ-FDA-014 | release-blocking | P14 (exit=1), P25 (decision-index outside the digest inputs, exit=1; control P25c exit=0) | M3 |
| AC-FDA-015 | Real chain fixture: plan-audit → T7 kickoff (lease empty, as P26 asserts today) → `audit` approval → card in the run stage with `LeaseHolder` = the record owner label, a fresh expiry, and the worker heartbeat updated, all in one transition; a `human` approval still lands in `assigned` (T8 preserved) | REQ-FDA-015 | release-blocking | P15 (only T8, exit=0), P26 (exit=0) | M3 |
| AC-FDA-016 | Under lane refusal, `factory decide <card> --gate kickoff --choice approve --decider audit` succeeds when the lane label equals the card's record owner; the same call from another registered lane, any `human` call from a lane, and any lane push/abandon/resume call are refused; SPEC-FACTORY-SELF-DISPATCH-001 carries the Amendments row | REQ-FDA-016 | release-blocking | P16 (human-only, exit=0), P31 (exit=0) | M3 |
| AC-FDA-017 | manager-spec (C1, C2, C3) defines `Class:`, `Default:`, `Alternate:`, the published Default rule, the closed product-level list as the Class rule with "no Class → product-level" as a fail-closed fallback, and the clause narrowed to judgment calls; the C2 commit precedes or equals the C1 commit | REQ-FDA-017 | release-blocking | P17 (exit=1) | M5 |
| AC-FDA-018 | The plan-close step (local + template) fills `DEFAULT-APPLIED <UTC> <runner+role>` on implementation-level rows with a Default before the plan audit; fixtures: the filled row passes the decider, a product-level empty row blocks, an implementation-level row without a Default stays empty and blocks; the three-item product-level definition is stated | REQ-FDA-018 | release-blocking | P18 (exit=1) | M5 |
| AC-FDA-019 | auto-semantics §5.1/§14 and the watchdog skill (local + template) define wait ids, `resolves`, the one-shot recheck (`recurring: false`) re-armed per open wait, delay default 5 with a 5-minute floor, and the codex named gap; fixture: an unrelated same-card board record leaves the wait open and arms one more one-shot; a record with `resolves: <id>` ends it | REQ-FDA-019 | release-blocking | P28 (wait records carry no id today, exit=0), P29 (only the recurring carrier exists, exit=0), P19 (exit=1) | M6 |
| AC-FDA-020 | Hook test: a matching cache on a live run performs the run-state probe and zero `factorymsg.Open` and peer queries, and emits no degraded notice even when Open would exceed its budget; a non-matching cache performs the full bind | REQ-FDA-020 | release-blocking | P30 (Open on every bind today, exit=0), P20 (exit=1), P21 (exit=0) | M7 |
| AC-FDA-021 | Hook test: a cached session whose run is retired in the fixture DB rebinds on the next prompt (rebind notice, cache rewritten, no stale binding); a different run likewise; a probe failure on a hit surfaces the degraded state | REQ-FDA-021 | release-blocking | P20 (exit=1), P21 (exit=0) | M7 |
| AC-FDA-022 | Hook test: a degraded inbox state produces one warn log line per occurrence and at most one session notice per configured interval | REQ-FDA-022 | release-blocking | P22 (only warn is broker close, exit=0) | M7 |
| AC-FDA-023 | Lane intake doctrine (kanban-dispatch, local + template) requires the MCP build vs `moai version` comparison and the `fallback=CLI` progress line | REQ-FDA-023 | release-blocking | P23 (exit=1) | M6 |
| AC-FDA-024 | Keep-set guard: no new code path records or approves push, release/main integration, queue admission, contract signing, abandon of an unintegrated branch, product-level verdicts, or final PASS/FAIL; `factory decide --gate push` still requires `human` | REQ-FDA-024 | RG (baseline: `5d094991f`) | P24 (push branch present, exit=0), P16 | preserve-through-close |
| AC-FDA-025 | Every changed shipped file has its template mirror change (token presence in both trees, e.g. `plan_audit_ceiling_policy` in both `harness.yaml`); `make agents-emit-check` clean; no SPEC ID, card id, or date in template diffs | REQ-FDA-025 | release-blocking | P10 (absent in both trees, exit=1) | M8 |

## Measurement notes (evidence obligations, not ACs)

- **M0(a):** the degraded-notice rate under load on the current bind path, committed before the
  first implementation commit, then re-measured after M7 with the cache. The post-M7 figure for
  already-bound sessions on a live run is expected to be zero; a non-zero figure is reported as a
  finding against AC-FDA-020.
- **M0(b):** the cache-write cost of a 5-minute one-shot recheck against longer delays. The shipped
  default equals the M0(b) recommendation, and it is never below 5 (REQ-FDA-019).

## Scenarios (Given-When-Then)

### AC-FDA-001 — lane cannot write the board
- **Given** a session with lane refusal holding and an existing board file with mtime T
- **When** it runs `moai decision record --scope standing --kind standing-rule --body "x" --evidence y`
- **Then** the exit code is non-zero, the board mtime is still T, and `moai decision read` from the same session exits 0

### AC-FDA-009 — one phase-scoped predicate, three sites
- **Given** the plan-phase and sync-phase fixture sets listed in the matrix
- **When** each fixture is evaluated by the contract rule, the kickoff evaluator, and the card-transition guard
- **Then** the admit/refuse outcome is identical across sites for every fixture, a PASS-labelled verdict with score below the tier threshold or `must_pass_failed: 1` is refused at T7, and every T13 outcome equals its pre-change characterization

### AC-FDA-011 — mechanical delta eligibility
- **Given** a Tier M SPEC at its ceiling (iteration 2) with verdict FAIL and `fix_scope: [spec.md#REQ-X, acceptance.md#AC-Y]`
- **When** the next commit changes only those anchor ranges plus `progress.md`, and the REQ/AC id sets are unchanged
- **Then** exactly one delta audit runs and one decision record citing `harness.plan_audit_ceiling_policy` is written, with no question emitted; the same fixture with an edited Out-of-Scope bullet outside `fix_scope`, or any `decision-index.md` change, takes the final-hit path

### AC-FDA-013 — release-blocking AC-wording exception
- **Given** a standing `release-scope` record listing card R, a relation `R depends C`, a final verdict for C with `blocking_count: 1`, `defect_class: ac-wording`, `reread_hunks: [acceptance.md#AC-Z]`, a hold wait `w1`, and a leader `card:C` board record naming the exception
- **When** the session changes only AC-Z's range and the auditor emits a `scope: reread` verdict at the fix SHA
- **Then** the reread verdict is admitted by the plan-phase predicate, a board record with `resolves: w1` releases the hold, and removing any one condition or the leader record leaves the hold in place

### AC-FDA-014 — post-audit decision-index edit refused
- **Given** a card in `kickoff` with an admitted verdict whose hash covered a decision-index holding a product-level row with an empty verdict
- **When** the row is rewritten as implementation-level with `DEFAULT-APPLIED` and `factory decide <card> --gate kickoff --choice approve --decider audit` runs
- **Then** the call is refused on the hash mismatch, the card stays in `kickoff`, and a subsequent `--decider human` approval succeeds

### AC-FDA-015 — owner leased in the same transition
- **Given** a card owned by worker-3 that passed T7 into `kickoff` with an empty lease, meeting every REQ-FDA-014 condition
- **When** worker-3 runs the `audit` approval
- **Then** the card's stage is `run`, `LeaseHolder` is worker-3 with a fresh expiry, and no separate lease call is needed

### AC-FDA-018 — FOUNDER default at plan close
- **Given** `decision_gate: on`, one implementation-level row with a Default, one product-level row, and one implementation-level row without a Default, all with empty verdicts
- **When** the plan-close step runs and the plan audit then passes
- **Then** only the first row reads `Operator verdict: DEFAULT-APPLIED <UTC> <runner+role>`, and the `audit` decider refuses Kickoff because the other two rows remain empty

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
- Verdict hash present but a plan artifact deleted → audit decider refused (the hash cannot be recomputed).
- The diff between two audited SHAs cannot be computed (unreachable SHA) → delta and exception both ineligible.
- A relation cycle in the release-reachability walk → the walk visits each card once; the cycle is not an error.
- Codex runner → no one-shot recheck; progress records the named gap.

## Quality gates

- Changed Go packages: `go test -timeout 30m ./internal/<pkg>/...`, `go test -race` on `internal/hook` and the board package, `go vet`, `golangci-lint run` (CI version).
- Coverage ≥ 85% on new packages (`quality.yaml` target).
- `make build`, `make agents-emit`, `make agents-emit-check` clean.
- `moai spec lint` clean on this SPEC.

## Definition of Done

The M0 baseline is committed before the first implementation commit. All 24 release-blocking ACs are
green with verbatim evidence in progress §E.2. AC-FDA-024 is re-asserted against its baseline at
close. Both measurement notes carry recorded figures. Sync-audit is admitted. Template mirrors and
emitted agents are committed. No keep-set behavior changed.
