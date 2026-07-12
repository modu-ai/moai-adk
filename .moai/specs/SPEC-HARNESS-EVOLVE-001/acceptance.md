# SPEC-HARNESS-EVOLVE-001 — Acceptance Criteria (SSOT)

> AC discipline: every AC pins ONE discriminating pattern with a measured
> baseline (2026-07-12, main working tree) and a post-implementation expected
> state. NO compound `grep "A|B|C|D" >= N` checks. Cross-file registrations
> (router, workflow bodies, hook handler, CLI verb, template mirrors) are
> pinned as SEPARATE ACs — a feature landing in file A without its
> registration in file B is inert (two prior incidents in this repo).
> Absence assertions are used ONLY for preservation/no-write invariants.

## §D. AC Matrix

### Group 1 — Package existence + Go quality (REQ-HEV-007/021)

**AC-HEV-001 — routing package tests pass**
- REQ: REQ-HEV-007, REQ-HEV-021
- Baseline: `internal/harness/routing/` absent (package does not build/test)
- Verify: `go test ./internal/harness/routing/...`
- Expected: exit 0, `ok  github.com/modu-ai/moai-adk/internal/harness/routing`

**AC-HEV-002 — coverage ≥ 90% (hook-adjacent threshold)**
- REQ: REQ-HEV-021
- Baseline: n/a (package absent)
- Verify: `go test -cover ./internal/harness/routing/`
- Expected: `coverage: ≥90.0% of statements` (verbatim figure cited in §E.2)

**AC-HEV-003 — race safety of concurrent append**
- REQ: REQ-HEV-007
- Baseline: n/a
- Verify: `go test -race -run TestConcurrentAppend ./internal/harness/routing/`
- Expected: PASS (test exists AND passes under `-race`)

### Group 2 — Schema v1 fields (REQ-HEV-001..004)

**AC-HEV-004 — ledger path + schema_version literal**
- REQ: REQ-HEV-001
- Baseline: `grep -rc "routing-ledger.jsonl" internal/harness/routing/ 2>/dev/null` → no files (dir absent)
- Verify: `grep -rl "routing-ledger.jsonl" internal/harness/routing/*.go | wc -l`
- Expected: ≥ 1 (non-test source carries the canonical path constant)

**AC-HEV-005 — core routing field: matched_subcommand**
- REQ: REQ-HEV-002
- Baseline: repo-wide `matched_subcommand` in internal/ → 0
- Verify: `grep -rc "matched_subcommand" internal/harness/routing/types.go`
- Expected: ≥ 1 (JSON tag present in schema types)

**AC-HEV-006 — A2 convergence field: convergence_class**
- REQ: REQ-HEV-003
- Baseline: repo-wide `convergence_class` in internal/ → 0
- Verify: `grep -rc "convergence_class" internal/harness/routing/types.go`
- Expected: ≥ 1

**AC-HEV-007 — A4 delegation trajectory: delegations JSON tag**
- REQ: REQ-HEV-004
- Baseline: `grep -rn "\"delegations\"" internal/` → 0
- Verify: `grep -c "delegations" internal/harness/routing/types.go`
- Expected: ≥ 1 (array field + entry struct `{agent, cycle_type?, outcome, blocker?}`)

**AC-HEV-008 — nullable convergence semantics (evidence-or-null)**
- REQ: REQ-HEV-003, REQ-HEV-006
- Baseline: n/a
- Verify: `go test -run TestConvergenceNullWhenNoSignal ./internal/harness/routing/`
- Expected: PASS (absent machine signal ⇒ `goal_converged: null` /
  `convergence_class: null` in the serialized row — never an inferred value)

### Group 3 — Privacy + machine-signal-only (REQ-HEV-005/006/013)

**AC-HEV-009 — no verbatim request text persisted**
- REQ: REQ-HEV-005
- Baseline: n/a
- Verify: `go test -run TestRequestDigestNoVerbatim ./internal/harness/routing/`
- Expected: PASS (writer given a sentinel request string; ledger + pending
  files greped for the sentinel → 0 occurrences; digest field matches
  `sha256:[0-9a-f]{12}`)

**AC-HEV-010 — DeriveOutcome deterministic precedence, no override**
- REQ: REQ-HEV-013, REQ-HEV-006
- Baseline: n/a
- Verify: `go test -run TestDeriveOutcome ./internal/harness/routing/`
- Expected: PASS (table-driven: `abort`-kind evidence marker > non-zero
  gate_exit > terminal passing signal > non-terminal; no code path accepts a
  free-text outcome; sweep/reroute paths bypass DeriveOutcome per REQ-HEV-013)

**AC-HEV-011 — no `--outcome` flag on the WRITE surfaces (record / evidence)**
- REQ: REQ-HEV-006
- Baseline: `internal/cli/harness_ledger.go` absent (`moai harness ledger` →
  unknown command)
- Verify (two write-surface probes, each independently):
  `go run ./cmd/moai harness ledger record --help | grep -c -- '--outcome'`
  AND `go run ./cmd/moai harness ledger evidence --help | grep -c -- '--outcome'`
- Expected: 0 AND 0 (absence assertion on the WRITE surfaces only — the
  un-fakeable contract binds outcome INPUT; the `ledger list --outcome` READ
  filter is legitimate row selection and is NOT covered by this AC. The ONLY
  outcome writers are `DeriveOutcome`, reroute, and stale-sweep abort)

### Group 4 — Separation from usage-log (REQ-HEV-009)

**AC-HEV-012 — routing package never touches usage-log**
- REQ: REQ-HEV-009
- Baseline: n/a (dir absent)
- Verify: `grep -rn "usage-log" internal/harness/routing/ | wc -l`
- Expected: 0 (absence assertion — separation invariant)

**AC-HEV-013 — no usage-log Event type reuse**
- REQ: REQ-HEV-009
- Baseline: n/a
- Verify: `grep -rn "harness\.Event\b" internal/harness/routing/ | wc -l`
- Expected: 0 (routing schema types are package-local)

### Group 5 — Pending-row lifecycle (REQ-HEV-010/011/012/014/015)

**AC-HEV-014 — reroute on same-session re-record**
- REQ: REQ-HEV-010
- Baseline: n/a
- Verify: `go test -run TestReroute ./internal/harness/routing/`
- Expected: PASS (second record ⇒ prior row appended with `"outcome":"reroute"`)

**AC-HEV-015 — Stop finalize self-gate: no pending ⇒ no-op**
- REQ: REQ-HEV-012
- Baseline: n/a
- Verify: `go test -run TestFinalize_SelfGate_NoPendingNoOp ./internal/harness/routing/`
- Expected: PASS (no pending file ⇒ no ledger append, no error, no side effect)

**AC-HEV-016 — non-terminal evidence stays pending (multi-turn)**
- REQ: REQ-HEV-012
- Baseline: n/a
- Verify: `go test -run TestFinalize_NonTerminalStaysPending ./internal/harness/routing/`
- Expected: PASS (pending file survives Stop finalize when DeriveOutcome is
  non-terminal)

**AC-HEV-017 — stale/foreign pending row swept as abort**
- REQ: REQ-HEV-014
- Baseline: n/a
- Verify: `go test -run TestStaleSweepAbort ./internal/harness/routing/`
- Expected: PASS (cross-session pending row — any age — and
  unresolvable-session pending row older than 24h are finalized
  `"outcome":"abort"` DIRECTLY (bypassing DeriveOutcome) before the new
  pending row is created; a same-session row is NEVER swept — it reroutes
  per REQ-HEV-010 precedence, covered by AC-HEV-014's table)

**AC-HEV-018 — fail-open finalizer**
- REQ: REQ-HEV-015
- Baseline: n/a
- Verify: `go test -run TestFinalize_FailOpen ./internal/harness/routing/`
- Expected: PASS (injected write failure ⇒ error surfaced to stderr sink,
  returned error nil-equivalent at the hook boundary; session end never blocked)

### Group 6 — Cross-file registrations (reachability — each pinned separately)

**AC-HEV-019 — Stop-hook handler registration (internal/cli/hook.go)**
- REQ: REQ-HEV-012
- Baseline (measured): `grep -c "harness/routing" internal/cli/hook.go` → **0**
- Verify: `grep -c "harness/routing" internal/cli/hook.go`
- Expected: ≥ 1 (the `runHarnessObserveStop` body calls the routing finalizer —
  without this registration the entire capture path is inert)

**AC-HEV-020 — CLI verb registration + reachability**
- REQ: REQ-HEV-011
- Baseline (measured): `moai harness ledger --help` → unknown-command error;
  `grep -rn "routing-ledger\|routingledger" internal/cli/` → **0**
- Verify: `go run ./cmd/moai harness ledger --help; echo "exit=$?"`
- Expected: exit 0 AND help text lists `record`, `evidence`, `list` sub-verbs

**AC-HEV-021 — HOI dual-gate inheritance (gate 0 + gate 1)**
- REQ: REQ-HEV-016
- Baseline: n/a
- Verify: `go test -run TestHarnessObserveStop_RoutingLedgerGated ./internal/cli/`
- Expected: PASS — table-driven over BOTH gates with explicit fixtures:
  (a) HOI OFF (`hook.opt_in.enabled` absent or `false` in the fixture
  `system.yaml`) ⇒ Stop-path finalizer NEVER reached, pending row survives
  (default-config dormancy per spec.md §D.3);
  (b) HOI ON (`hook.opt_in.enabled: true`) + `learning.enabled: false` ⇒
  no finalize, no append (gate 1);
  (c) HOI ON + `learning.enabled` absent/true ⇒ capture active, terminal
  evidence finalizes.
  Every active-path fixture MUST set `hook.opt_in.enabled: true` explicitly —
  relying on the default would silently test the dormant path.

### Group 7 — Workflow skill wiring (REQ-HEV-017/018) — per-file baseline-0 pins

**AC-HEV-022a — SKILL.md router recording obligation**
- REQ: REQ-HEV-017
- Baseline (measured): `grep -c "routing-ledger" .claude/skills/moai/SKILL.md` → **0**
- Verify: `grep -c "routing-ledger" .claude/skills/moai/SKILL.md`
- Expected: ≥ 1

**AC-HEV-022b — workflows/plan.md recording obligation**
- REQ: REQ-HEV-018
- Baseline (measured): `grep -c "routing-ledger" .claude/skills/moai/workflows/plan.md` → **0**
- Verify: `grep -c "routing-ledger" .claude/skills/moai/workflows/plan.md`
- Expected: ≥ 1

**AC-HEV-022c — workflows/run.md recording obligation**
- REQ: REQ-HEV-018
- Baseline (measured): `grep -c "routing-ledger" .claude/skills/moai/workflows/run.md` → **0**
- Verify: `grep -c "routing-ledger" .claude/skills/moai/workflows/run.md`
- Expected: ≥ 1

**AC-HEV-022d — workflows/sync.md recording obligation**
- REQ: REQ-HEV-018
- Baseline (measured): `grep -c "routing-ledger" .claude/skills/moai/workflows/sync.md` → **0**
- Verify: `grep -c "routing-ledger" .claude/skills/moai/workflows/sync.md`
- Expected: ≥ 1

### Group 8 — Template-First + neutrality (REQ-HEV-019/020)

**AC-HEV-023a — template mirror carries the router obligation (SKILL.md)**
- REQ: REQ-HEV-019
- Baseline (measured): `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/SKILL.md` → **0**
- Verify: `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/SKILL.md`
- Expected: ≥ 1

**AC-HEV-023b — template neutrality preserved on edited mirrors**
- REQ: REQ-HEV-019
- Baseline: neutrality CI guards green
- Verify: `grep -c "SPEC-HARNESS-EVOLVE" internal/template/templates/.claude/skills/moai/SKILL.md`
- Expected: 0 (absence assertion — no internal SPEC ID leaks into the
  template copy) AND both named guard tests stay green:
  `go test -run TestTemplateNeutralityAudit ./internal/template/...` +
  `go test -run TestTemplateNoInternalContentLeak ./internal/template/...`

**AC-HEV-023c — template mirror: workflows/plan.md**
- REQ: REQ-HEV-019
- Baseline (measured): `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/workflows/plan.md` → **0**
- Verify: `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/workflows/plan.md`
- Expected: ≥ 1

**AC-HEV-023d — template mirror: workflows/run.md**
- REQ: REQ-HEV-019
- Baseline (measured): `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/workflows/run.md` → **0**
- Verify: `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/workflows/run.md`
- Expected: ≥ 1

**AC-HEV-023e — template mirror: workflows/sync.md**
- REQ: REQ-HEV-019
- Baseline (measured): `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/workflows/sync.md` → **0**
- Verify: `grep -c "routing-ledger" internal/template/templates/.claude/skills/moai/workflows/sync.md`
- Expected: ≥ 1

**AC-HEV-024 — no ledger DATA in templates**
- REQ: REQ-HEV-020
- Baseline: `find internal/template/templates -name "routing-ledger*" | wc -l` → 0
- Verify: `find internal/template/templates -name "routing-ledger*" -o -name "routing-pending*" | wc -l`
- Expected: 0 (absence assertion — mechanism ships, data never does)

### Group 9 — Repo-level gates

**AC-HEV-025 — full suite + cross-platform build green**
- REQ: REQ-HEV-021 (+ all)
- Baseline: `go test ./...` green pre-change (verify at pre-flight)
- Verify: `go test ./... ; go build ./... ; GOOS=windows GOARCH=amd64 go build ./...`
- Expected: all exit 0; no existing test modified to pass (usage-log / hook
  tests untouched-green)

**AC-HEV-027 — verb-surface CI guard registration (`v3r5RequiredHarnessVerbs`)**
- REQ: REQ-HEV-011
- Baseline (measured): `grep -c '"ledger"' internal/cli/harness_retirement_test.go` → **0**
- Verify: `grep -c '"ledger"' internal/cli/harness_retirement_test.go`
- Expected: ≥ 1 (the `ledger` verb is added to `v3r5RequiredHarnessVerbs`
  (lines 31-50) so `TestHarnessV3R5VerbSurface` pins the live-tree
  registration — the guard that prevents the twice-shipped inert-verb failure
  mode) AND `go test -run TestHarnessV3R5VerbSurface ./internal/cli/` PASS

**AC-HEV-026 — gitignore coverage of the ledger path (preservation)**
- REQ: REQ-HEV-001
- Baseline (measured): `.moai/state/` present in .gitignore (lines 198, 265) — already green
- Verify: `git check-ignore -v .moai/state/routing-ledger.jsonl; echo "exit=$?"`
- Expected: exit 0 with the matching `.gitignore` pattern printed
  (`git check-ignore` evaluates ignore patterns without the file existing —
  NO `touch`/`rm` of the live ledger path; a post-implementation re-run must
  never delete real observation data)

## §D.1 REQ → AC Traceability

| REQ | AC(s) |
|-----|-------|
| REQ-HEV-001 | AC-HEV-004, AC-HEV-026 |
| REQ-HEV-002 | AC-HEV-005 |
| REQ-HEV-003 | AC-HEV-006, AC-HEV-008 |
| REQ-HEV-004 | AC-HEV-007 |
| REQ-HEV-005 | AC-HEV-009 |
| REQ-HEV-006 | AC-HEV-008, AC-HEV-010, AC-HEV-011 |
| REQ-HEV-007 | AC-HEV-001, AC-HEV-003 |
| REQ-HEV-008 | AC-HEV-001 (reader filter subtests: `TestReaderFilters`, `TestReaderSkipsMalformed`) |
| REQ-HEV-009 | AC-HEV-012, AC-HEV-013 |
| REQ-HEV-010 | AC-HEV-014 |
| REQ-HEV-011 | AC-HEV-020, AC-HEV-027 |
| REQ-HEV-012 | AC-HEV-015, AC-HEV-016, AC-HEV-019 |
| REQ-HEV-013 | AC-HEV-010 |
| REQ-HEV-014 | AC-HEV-017 |
| REQ-HEV-015 | AC-HEV-018 |
| REQ-HEV-016 | AC-HEV-021 |
| REQ-HEV-017 | AC-HEV-022a |
| REQ-HEV-018 | AC-HEV-022b, AC-HEV-022c, AC-HEV-022d |
| REQ-HEV-019 | AC-HEV-023a, AC-HEV-023b, AC-HEV-023c, AC-HEV-023d, AC-HEV-023e |
| REQ-HEV-020 | AC-HEV-024 |
| REQ-HEV-021 | AC-HEV-002, AC-HEV-025 |

## §D.2 Given-When-Then scenarios (representative)

**Scenario 1 — happy-path routing observation (10 routings → 10 rows)**
- Given a project with `hook.opt_in.enabled: true` set explicitly in
  `system.yaml` (HOI gate 0 open — required for the Stop path; spec.md §D.3)
  and `learning.enabled` unset (default enabled)
- When the orchestrator dispatches 10 `/moai` routings, each recorded via
  `ledger record`, each supplied terminal machine evidence, and Stop fires
- Then `routing-ledger.jsonl` carries exactly 10 rows, each with
  `schema_version: 1`, a `sha256:`-prefixed digest, and an `outcome` derived
  only from the supplied machine evidence (design §7 M1 verification clause:
  "라우팅 10회 → 10행, outcome 100%")

**Scenario 2 — multi-turn pipeline stays open, converges later**
- Given a `/moai run` dispatch recorded with no terminal evidence yet
- When Stop fires at intermediate turn-ends
- Then the pending row survives each Stop (no premature finalize); and when a
  terminal `gate_exit: 0` evidence lands and the next Stop fires, the row is
  finalized `success` with `loop_iterations`/`convergence_class` populated
  from machine artifacts (or null when no such artifact exists)

**Scenario 3 — session death → abort (edge)**
- Given a session that recorded a pending row and terminated without terminal
  evidence
- When a later session runs `ledger record`
- Then the foreign stale row is finalized `outcome: abort` first, then the new
  pending row is created (no orphaned pending files accumulate)

**Scenario 4 — delegation trajectory (edge, A4)**
- Given a run-phase dispatch that delegates to manager-develop (cycle_type=tdd)
  which returns a blocker report
- When the orchestrator appends the delegation evidence and the row finalizes
- Then `delegations[]` carries
  `{"agent":"manager-develop","cycle_type":"tdd","outcome":"fail","blocker":"<category>"}`
  (blocker presence is a structural artifact, not a prose judgment)

**Scenario 5 — learning disabled (gate 1, isolated)**
- Given `hook.opt_in.enabled: true` (gate 0 open — isolating gate 1) AND
  `.moai/config/sections/harness.yaml` with `learning.enabled: false`
- When `ledger record` / Stop finalize run
- Then no pending file and no ledger row are created (silent no-op, exit 0)

**Scenario 6 — HOI default dormancy (gate 0, edge — audit D1)**
- Given DEFAULT shipped config (`hook.opt_in.enabled` absent or `false`) and a
  pending row with terminal evidence
- When the Stop hook fires
- Then the Stop-path finalizer is never reached and the pending row survives
  (EXPECTED dormancy per spec.md §D.3); the row is later finalized only by
  record-time reroute or the lazy abort sweep — never as `success`/`fail`

## §D.3 Quality Gate / Definition of Done

- [ ] All 27 AC IDs PASS — 34 verification line items including sub-letters
  022a-d and 023a-e — with verbatim command outputs (E1 matrix)
- [ ] `internal/harness/routing` coverage ≥ 90%; no `t.Skip` without issue link
- [ ] `go test -race ./internal/harness/routing/...` green
- [ ] Cross-platform build green (`GOOS=windows GOARCH=amd64`)
- [ ] golangci-lint: 0 NEW issues vs pre-flight baseline
- [ ] `moai spec lint`: 0 errors for SPEC-HARNESS-EVOLVE-001
- [ ] Template neutrality guard green; live↔template pair diffs reviewed
- [ ] plan.md §H clarifications RESOLVED (3 pinned user decisions, markers
  struck — recorded in progress.md §E.1 addendum); MP-7 gate clear
- [ ] M4 local HOI opt-in (`hook.opt_in.enabled: true` in this dev repo's
  system.yaml) applied for live Stop-path verification, template default
  untouched
- [ ] No modifications outside plan.md §A.5 PRESERVE boundary
