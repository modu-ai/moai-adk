# Acceptance Criteria — SPEC-DEADPKG-INVESTIGATE-001

## §A. Definition of Done

- All 5 flagged packages carry a recorded verdict (GENUINELY DEAD / INTENTIONALLY RETAINED /
  RELOCATE) with the 4 mandatory evidence outputs in `progress.md` §E.2.
- No package removed or relocated without its verdict evidence (REQ-DPI-002).
- `internal/runtime/gobin` unchanged (REQ-DPI-011).
- Cross-platform build green + `go test ./...` passing after every milestone and at closure
  (REQ-DPI-012).
- `internal/migrate` atomic-write safety contract preserved regardless of disposition
  (REQ-DPI-013).
- Final verdict catalog produced in M6.

## §B. Given-When-Then Scenarios

### Scenario 1 — Investigation precedes disposition (the core invariant)

- **Given** a flagged package with zero external importers (a "dead" hypothesis),
- **When** the run-phase agent begins the package's milestone,
- **Then** it runs all four checks (reachability from `cmd/moai`, `git log` intent, `@MX:REASON`
  scan, owning-SPEC status) and records their verbatim output in `progress.md` §E.2 **before**
  assigning a verdict, and **no removal occurs until the verdict is GENUINELY DEAD**.

### Scenario 2 — Stale `@MX:REASON` disconfirmation (internal/research)

- **Given** `internal/research/safety/limiter.go` carries `@MX:REASON` claiming `NewRateLimiter`
  is used by CLI + auto-update,
- **When** the run-phase agent greps the whole repo for external importers of
  `internal/research/safety`,
- **Then** it records the observed result; **if** the grep confirms zero external importers, the
  retention claim is marked **disconfirmed** and does not by itself block a GENUINELY DEAD verdict.

### Scenario 3 — Owning SPEC forces retention default (internal/runtime)

- **Given** `internal/runtime` (top-level) has owning `SPEC-WF-AUDIT-GATE-001` with
  `status: implemented` and is exercised only by test files,
- **When** the run-phase agent checks the owning SPEC status,
- **Then** the package defaults to **INTENTIONALLY RETAINED** and a retention marker is added,
  **unless** the agent records disconfirming evidence that the owning SPEC no longer intends the
  package to exist. `internal/runtime/gobin` is not modified.

### Scenario 4 — RELOCATE preserves behavior and cleans the original (internal/migrate)

- **Given** `internal/migrate.CleanupUserSettings` has zero production callers but carries a P0-4
  atomic-write safety contract,
- **When** the verdict is RELOCATE,
- **Then** the symbol is moved to `internal/migration` (as a migration step) with the atomic-write
  behavior + its regression test preserved, all call sites updated, the `@MX:REASON` referent in
  `internal/hook/retired_events.go` updated, and the emptied `internal/migrate` + its `_test.go`
  removed **in the same milestone**, with the build staying green.

### Scenario 5 — Genuinely dead removal is complete (edge: orphan cleanup)

- **Given** a package receives a GENUINELY DEAD verdict,
- **When** it is removed,
- **Then** its `_test.go` files are removed in the same milestone AND any now-orphaned
  `@MX:REASON` referent elsewhere is updated, so `grep` finds no dangling reference to the
  removed package, and `go build ./...` + `go test ./...` pass.

## §C. Edge Cases

- **EC-1 — Ambiguous reachability (runtime).** Textual `GateConfig`/`AuditGate` matches exist in
  `config/`/`lsp/`/`hook/`. The agent MUST disambiguate whether these import `internal/runtime`
  or are a different gate; a textual match is not an importer.
- **EC-2 — Skill retired but package alive (design).** `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001`
  retired the skill; the agent MUST NOT infer the Go package is dead solely from the skill
  retirement — it must check for live DTCG consumers.
- **EC-3 — Wiring gap vs dead (i18n).** If github production code *should* call
  `NewCommentGenerator` but does not, that is a wiring gap (out of scope) → RETAINED + flag a
  follow-up SPEC, not a removal.
- **EC-4 — Genuine staged plan (migrate).** If "future observability tooling" is a real staged
  plan (owning SPEC or roadmap entry), the verdict is RETAINED + marker, not RELOCATE.

## §D. AC Matrix (REQ → AC)

| AC ID | REQ | Criterion | Verification |
|-------|-----|-----------|--------------|
| AC-DPI-001 | REQ-DPI-001 | Each of 5 packages has exactly one verdict recorded | Read `progress.md` §E.2; 5 verdicts present |
| AC-DPI-002 | REQ-DPI-002/003 | No removal/relocation without recorded evidence; deletion not prescribed up front | `git log --stat` shows no package removal in a milestone whose §E.2 lacks the 4 evidence outputs |
| AC-DPI-003 | REQ-DPI-004 | Reachability grep from `cmd/moai` recorded per package | §E.2 contains the verbatim grep command + output for each package |
| AC-DPI-004 | REQ-DPI-005 | `git log` intent check recorded per package | §E.2 contains the most-recent commit + SPEC provenance per package |
| AC-DPI-005 | REQ-DPI-006 | `@MX:REASON` scan recorded; each retention claim confirmed/disconfirmed | §E.2 marks the research `safety/limiter.go` claim disconfirmed (or records live consumer) |
| AC-DPI-006 | REQ-DPI-007 | Owning-SPEC status recorded; retention default applied where implemented | §E.2 records `SPEC-WF-AUDIT-GATE-001: implemented` and runtime → RETAINED default |
| AC-DPI-007 | REQ-DPI-008 | Retained packages get a greppable retention marker | `grep` finds the `@MX:NOTE`/package-doc marker in each RETAINED package |
| AC-DPI-008 | REQ-DPI-009 | Relocated symbols moved + call sites updated + original removed same milestone | Build green; original package gone; new home imported by prior consumers |
| AC-DPI-009 | REQ-DPI-010 | Dead packages removed with `_test.go` + orphan referents cleaned | `grep` finds no dangling reference to any removed package |
| AC-DPI-010 | REQ-DPI-011 | `internal/runtime/gobin` unchanged | `git diff --stat` shows no change under `internal/runtime/gobin/` |
| AC-DPI-011 | REQ-DPI-012 | Cross-platform build + tests green after every milestone and at closure | `go build ./...` + `go test ./...` output in §E.2 per milestone |
| AC-DPI-012 | REQ-DPI-013 | `internal/migrate` atomic-write safety preserved | The atomic-write regression test still exists and passes at the symbol's disposition site |
| AC-DPI-013 | REQ-DPI-001..010 | M6 final verdict catalog produced | `progress.md` §E.2 contains the package → verdict → disposition → evidence table |

## §E. Quality Gate

- Zero build errors, zero test failures at closure.
- Lint clean on all touched files.
- No `@MX:REASON` / import reference to any removed package survives (grep-clean).
- Every verdict is evidence-attributed per verification-claim-integrity §2 (command + observed
  output, not an assumption).
