# SPEC-MANAGER-TODO-001 — acceptance.md

Canonical AC enumeration. Every AC is mechanically checkable — a command plus an observable output. Every REQ in spec.md §B maps to ≥1 AC (traceability table in §D.8). Evidence per the evidence-bearing report format lands in `.moai/reports/t1306/verdict.md`.

## §D.1 — Agent rename/repurpose (M1)

**AC-MT-001** — REQ-MT-001

**Given** the worktree after M1 **When** both agent files are read **Then** `.claude/agents/moai/manager-todo.md` and `internal/template/templates/.claude/agents/moai/manager-todo.md` exist; each carries `name: manager-todo`, a primary mission covering todo-queue management, the Jev boundary, and dispatch/management ownership, plus a preserved sealed-snapshot judgment sub-role section; frontmatter has `tools:` as a CSV string, no `model:` and no `effort:` keys, and a `permissionMode` value in the official enum.

Check: `grep -c "^name: manager-todo" <both files>` = 1 each; `grep -E "^(model|effort):" <files>` exits 1 (no match); `awk` CSV check on `tools:` (no `YAML array` form).

**AC-MT-002** — REQ-MT-002

**Given** the C2 edit is in place **When** `make agents-emit && make agents-emit-check` runs **Then** both exit 0; `internal/template/templates/.codex/agents/moai/manager-todo.toml` exists; `mission-governor.toml` is absent from that directory; `git diff --stat` before/after shows no hand edit inside `internal/template/templates/.codex/agents/moai/` other than the emitter's own output.

**AC-MT-003** — REQ-MT-003

**Given** M1 is complete **When** `ls .claude/agents/moai/mission-governor.md internal/template/templates/.claude/agents/moai/mission-governor.md internal/template/templates/.codex/agents/moai/mission-governor.toml` runs **Then** every path reports "No such file or directory" (all three exits non-zero).

**AC-MT-004** — REQ-MT-004

**Given** the rename is complete **When** the catalog enumeration check runs **Then** (a) `go test ./internal/template/... ./internal/harness/rosterguard/... ./internal/harness/delegationmap/...` exits 0 (the retained-agents, catalog, and roster tests encode the 13/12 invariant); (b) `grep -rn "mission-governor" .claude/agents/ CLAUDE.md internal/template/retained_agents.go internal/template/catalog.yaml` returns zero hits; (c) the same grep for `manager-todo` returns ≥1 hit in each of those files.

## §D.2 — Reference sweep refresh (M2)

**AC-MT-005** — REQ-MT-005

**Given** M2 is complete **When** `git grep -in "mission.governor" -- . ':(exclude).moai/reports' ':(exclude).moai/specs'` runs **Then** every returned `file:line` appears in the research.md §B baseline (as updated at M2 start) with disposition FROZEN or HIST, and the count of hits outside those dispositions is exactly 0. The command output and the disposition reconciliation are pasted into the verdict.

**AC-MT-006** — REQ-MT-005

**Given** the M2-start re-sweep **When** its output is diffed against the plan-phase baseline table **Then** any hit absent from the plan-phase table was added to the checklist with a disposition before M2 closed (the reconciliation is recorded in the verdict; zero unreviewed new hits).

**AC-MT-007** — REQ-MT-006

**Given** the D-3 disposition is applied **When** the four roster surfaces are inspected **Then** `internal/cli/codex_audit_mcp.go`, `internal/cli/codex_role_fingerprint.go`, `internal/harness/rosterguard/`, and `internal/template/agentemit/agents-codex.yaml` agree on one post-rename role set (no `mission-governor` entry, `manager-todo` only if the disposition retained it), and `go test ./internal/cli/... ./internal/template/agentemit/...` exits 0.

**AC-MT-008** — REQ-MT-005 (frozen fixtures untouched)

**Given** M2 is complete **When** `git log --oneline -- internal/cli/testdata/codex-rollouts-t1171/` is read **Then** no commit in this card's range touched that directory (the fixture set is byte-identical to base `b7ff456b7`).

## §D.3 — `/moai:todo --auto` serial mode (M3)

**AC-MT-009** — REQ-MT-007 (serial, one at a time)

**Given** a fixture queue with 3 pickup targets **When** `moai todo --auto` runs against the fixture store **Then** the output shows the accept → process → accept-next cycle with exactly one card in flight at any time — no two card ids interleaved within one processing segment — and the cards are taken in the REQ-MT-008 order. Check: a table-driven test asserts the observed processing sequence equals the expected serial sequence, and a concurrency probe (two in-flight markers never co-occur) holds.

**AC-MT-010** — REQ-MT-008 + REQ-MT-010 (pickup predicate, positive state vocabulary)

**Given** the pickup-selection source and a fixture queue containing cards in every current state (queued, picked-with-dead-owner, picked-with-live-owner, done) **When** (a) the source is inspected and (b) the selection function is run table-driven **Then** (a) the predicate reads as `state == "queued"` plus the dead-owner `picked` carve-out — no negation-of-terminal-state form appears (`grep -n 'state != ' <source>` returns no pickup-predicate hit); (b) exactly the queued cards and the dead-owner picked cards are selected, in that order (dead-owner picked first, then queue order); and (c) the table includes a `hold`-shaped unknown-state row which is never selected — proving the forward-compatibility property with no SPEC revision.

**AC-MT-011** — REQ-MT-008 + REQ-MT-009 (liveness measurement)

**Given** three fixture scenarios — (1) registry-dead owner + clean lsof, (2) registry-live owner, (3) registry-dead owner but a live lsof cwd hit in the owning tree **When** the selection runs per scenario **Then** scenario 1 selects the picked card; scenarios 2 and 3 reject it (untouchable) and the cycle proceeds to queue-order cards or ends. The negative scenarios are the positive control inverted — both arms must reproduce, per the "a control that reproduces in neither arm cannot separate them" lesson.

**AC-MT-012** — REQ-MT-009 (absolute no-takeover)

**Given** a queue whose only card is picked by a measured-live owner **When** `--auto` runs to exhaustion **Then** that card's state is unchanged (still picked, same owner recorded) and the cycle reports zero eligible targets without mutating it. Check: fixture-store snapshot before/after is byte-identical for that card's row.

**AC-MT-013** — REQ-MT-011 (/clear guidance per completed card)

**Given** a fixture queue with 2 pickup targets **When** `--auto` processes both to completion **Then** the output contains a /clear guidance block after EACH completed card (2 blocks total), each naming the completed card and the next step, positioned before the next accept. Check: the emission-path source carries the guidance construction, and the fixture run's captured output matches the 2-block pattern.

**AC-MT-014** — REQ-MT-012 (batch approval = the invocation)

**Given** the command source and its help/skill text **When** inspected **Then** promotion authority is derived only from the `--auto` invocation event — the source contains no reordering, admission, or drop logic beyond queue order (`grep` for card-creating or reordering calls in the `--auto` path returns none), and the workflow text states the invocation-is-batch-approval reconciliation verbatim.

**AC-MT-015** — REQ-MT-013 (no factory lease/slot)

**Given** M3 is complete **When** (a) the `--auto` code path is inspected and (b) a fixture `--auto` run executes **Then** (a) `grep -n "lease\|factory_next\|factory" <auto-path files>` returns no factory-lease or factory-slot acquisition call, and (b) the fixture run creates no lease file and claims no slot (fixture-state assertion).

**AC-MT-016** — REQ-MT-007 (no parallel-card regression guard)

**Given** the `--auto` test suite **When** the suite runs `go test -timeout 30m ./internal/cli/...` **Then** the serial-cycle tests (AC-MT-009/010/011 fixtures) pass, providing the regression guard that the one-card-at-a-time invariant survives future edits.

## §D.4 — Jev boundary + GTD absorption (M4)

**AC-MT-017** — REQ-MT-014 (consultation is display-only)

**Given** the agent body and the `--auto` consultation path **When** inspected **Then** Jev invocation appears only as judgment-order/priority input (`triage`/`route` output read as a labelled signal), and no Jev output value is passed to any queue-mutating function (source-path check: no data flow from a Jev result into `add`/`pick`/`done`/`drop` call arguments).

**AC-MT-018** — REQ-MT-015 (never third-grade authority)

**Given** the agent body **When** a body-content test runs **Then** the test asserts the grade-3 authority list (queue mutation, completion verdict, merge approval, operator gates) is present as prohibited uses, and a degraded-mode scenario (Jev absent) proceeds with a labelled non-finding and completes the cycle on lead judgment (fixture: scripts absent from PATH → no error exit, guidance line emitted, exit 0).

**AC-MT-019** — REQ-MT-016 (GTD absorption, receipt intact)

**Given** M4 is complete **When** (a) the goal workflow text + template mirror are read and (b) `go test ./internal/mission/... ./internal/cli/...` runs **Then** (a) both name `manager-todo` for the judgment role and preserve the read-only sub-role contract clauses (bounded decision, never applies, blocker-on-excess-scope), and (b) the receipt tests pass with the unchanged schema — `goal run --governor-receipt` accepts a receipt of the same shape (flag name unchanged; help text updated; `grep -c "governor-receipt" internal/cli/goal.go` ≥ 1).

## §D.5 — Docs and locales (M5)

**AC-MT-020** — REQ-MT-018 (docs-site parity)

**Given** M5 is complete **When** the docs-site verification recipe runs (hugo build warning-free; 4-locale file-existence and section parity for every touched page) **Then** all checks exit 0, and every locale of each page named RENAME in research.md §B carries the manager-todo naming (spot grep per locale).

**AC-MT-021** — REQ-MT-018 (README 4-locale parity)

**Given** M5 is complete **When** the README set is checked **Then** README.md / README.ko.md / README.ja.md / README.zh.md each have their 2 baseline hits updated with section parity across the 4 files (per-file hit-count equality on the updated sections), and no `mission-governor` remains outside HIST-dispositioned files.

**AC-MT-022** — REQ-MT-018 (template neutrality)

**Given** the full milestone diff **When** the neutrality checks run — `make build`, `make agents-emit-check`, `make commands-emit-check`, the template-neutrality forbidden-class spot grep over `git diff --name-only base...HEAD -- internal/template/templates/` content, and `go test ./internal/template/...` (including content-leak tests) **Then** all exit 0 and no forbidden class (other cards' ids, internal dates, commit SHAs, macOS-bias paths, local-instruction references) appears in any template-mirrored diff hunk.

## §D.6 — Edge cases

- Empty queue at `--auto` start: cycle ends after one zero-target report; no mutation, no guidance emission for a non-existent card.
- `--auto` interrupted mid-card: the in-flight card remains in its pre-cycle state for the current session (no phantom done/pick); the next invocation re-runs the pickup predicate from scratch (D-4 re-measurement rule).
- Queue order ties (same enqueue timestamp): deterministic order by the store's primary key — the table test pins the ordering function.
- `lsof` unavailable on the platform: the liveness channel degrades to registry-only with a labelled degraded-measurement notice, and the conservative default holds (no takeover on an unmeasurable channel).

## §D.7 — Quality gates

- TRUST 5: Tested (affected-package tests, ≥85% coverage on new `--auto` code per `quality.yaml` `test_coverage_target: 85`), Readable/Unified (gofmt, golangci-lint), Secured (no Jev authority, no takeover of live-owner work — both are safety properties under test), Trackable (conventional commits carrying the card id).
- Gates: `make agents-emit-check`, `make commands-emit-check`, `make build`, `golangci-lint run` on changed packages, docs-site verify recipe (M5).

## §D.8 — Definition of Done

1. All 22 ACs PASS with evidence in `.moai/reports/t1306/verdict.md`.
2. REQ→AC traceability complete (every REQ-MT-001..018 appears in the map below).
3. `[NEEDS CLARIFICATION: codex read-only role roster disposition]` resolved and recorded.
4. Sweep grep returns only dispositioned hits (AC-MT-005).
5. Working tree clean; artifacts and implementation committed on `WT-manager-todo-agent`.

### Traceability map

| REQ | ACs |
|-----|-----|
| REQ-MT-001 | AC-MT-001 |
| REQ-MT-002 | AC-MT-002 |
| REQ-MT-003 | AC-MT-003 |
| REQ-MT-004 | AC-MT-004 |
| REQ-MT-005 | AC-MT-005, AC-MT-006, AC-MT-008 |
| REQ-MT-006 | AC-MT-007 |
| REQ-MT-007 | AC-MT-009, AC-MT-016 |
| REQ-MT-008 | AC-MT-010, AC-MT-011 |
| REQ-MT-009 | AC-MT-011, AC-MT-012 |
| REQ-MT-010 | AC-MT-010 |
| REQ-MT-011 | AC-MT-013 |
| REQ-MT-012 | AC-MT-014 |
| REQ-MT-013 | AC-MT-015 |
| REQ-MT-014 | AC-MT-017 |
| REQ-MT-015 | AC-MT-018 |
| REQ-MT-016 | AC-MT-019 |
| REQ-MT-017 | AC-MT-007, AC-MT-019 |
| REQ-MT-018 | AC-MT-020, AC-MT-021, AC-MT-022 |
