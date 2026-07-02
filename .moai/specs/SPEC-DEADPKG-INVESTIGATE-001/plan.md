# Plan — SPEC-DEADPKG-INVESTIGATE-001

## §A. Context

Per-package investigation-then-disposition of 5 packages flagged as candidate dead code.
Deletion is **conditional** on an evidence-backed verdict, never prescribed up front. The
governing doctrine is verification-claim-integrity §1.1 surface 3: a "dead package" is a
defect claim, hence a hypothesis until the run-phase reachability/intent/retention/owning-SPEC
checks confirm it.

Development mode: DDD (characterization-first for existing code) — before removing or relocating
any symbol, capture the behavior the flagged package currently provides so a RELOCATE preserves
it and a removal is provably safe (no live consumer to break).

## §B. Known Issues / Traps (surfaced at plan-phase)

- **B1 — Stale `@MX:REASON` (research).** `internal/research/safety/limiter.go` claims
  `NewRateLimiter` is "used by research, CLI, and auto-update"; grep shows **zero** external
  importers. The retention claim is a hypothesis to disconfirm, not evidence of life.
- **B2 — `gobin` is LIVE (runtime).** `internal/runtime/gobin` must not be touched. Every
  `runtime` command in this plan is scoped to top-level `internal/runtime/*.go`, never `gobin/`.
- **B3 — Owning SPEC = implemented (runtime).** `SPEC-WF-AUDIT-GATE-001` is `implemented`; the
  top-level `runtime` package is its implementation, currently only test-exercised. Default to
  INTENTIONALLY RETAINED unless disconfirming evidence surfaces.
- **B4 — Skill retirement ≠ package retirement (design).** `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001`
  retired the *skill*; whether the Go DTCG validator was meant to survive is a distinct question.
- **B5 — Orphaned referent (migrate).** If `internal/migrate` is removed, the `@MX:REASON`
  retention line in `internal/hook/retired_events.go` becomes dangling and must be updated in
  the same milestone (REQ-DPI-010).
- **B6 — Atomic-write safety (migrate).** `hook_cleanup.go` carries a P0-4 `@MX:WARN`
  atomic-write regression guard. Any disposition must preserve it (REQ-DPI-013).

## §C. Pre-flight (run-phase, before M1)

1. `git fetch origin main` + divergence check (multi-session race mitigation).
2. Rebuild the binary so audit reasoning is against a fresh tree (avoid stale-binary trap).
3. Confirm the plan-phase reachability baseline (spec.md §A.1) still holds.

## §D. Constraints

- Cross-platform build green after every milestone (REQ-DPI-012): `go build ./...` + the
  project cross-platform matrix; `go test ./...` passes.
- No removal without recorded verdict evidence in `progress.md` §E.2 (REQ-DPI-002).
- `internal/runtime/gobin` untouched (REQ-DPI-011).
- Removed packages: `_test.go` files removed in the same milestone (REQ-DPI-010).
- No new production wiring (spec.md §D Out of Scope).

## §E. Self-Verification (run-phase deliverables)

Each milestone MUST record in `progress.md` §E.2:
- E-a: the 4 investigation outputs (reachability grep from `cmd/moai`, `git log`, `@MX` scan,
  owning-SPEC status) — verbatim command + observed output.
- E-b: the verdict (GENUINELY DEAD / INTENTIONALLY RETAINED / RELOCATE) with the decisive
  evidence line.
- E-c: the disposition action taken (marker added / symbol relocated + call sites updated /
  package+tests removed + orphan referent cleaned).
- E-d: `go build ./...` + `go test ./...` result after the milestone.

## §F. Milestones (priority-based, one investigation+disposition unit per package)

> Order chosen so the highest-signal / lowest-risk investigations run first and the
> relocation-heavy packages last. No time estimates (priority ordering only).

### M1 — `internal/runtime` (top-level) investigation + disposition
- Priority: High (clearest owning-SPEC signal; sets the retention-marker pattern).
- Investigate (REQ-DPI-004..007): reachability from `cmd/moai`; confirm only test importers;
  read `SPEC-WF-AUDIT-GATE-001` status + intent; disambiguate the `GateConfig`/`AuditGate`
  textual matches in `config/`/`lsp/`/`hook/` (confirm they are a *different* gate, not
  importers of `internal/runtime`).
- Disposition per verdict (hypothesis: INTENTIONALLY RETAINED): add retention marker
  (`@MX:NOTE` naming the owning SPEC + the wiring-pending condition). Do NOT touch `gobin`.

### M2 — `internal/research` investigation + disposition
- Priority: High (stale `@MX:REASON` disconfirmation is high-value).
- Investigate: disconfirm the `safety/limiter.go` retention claim across the whole repo;
  reachability from `cmd/moai`; `git log` for a staged A/B rollout SPEC; `@MX` scan.
- Disposition per verdict (hypothesis: GENUINELY DEAD): if confirmed dead, remove the
  package + all `_test.go` in the same milestone. If a staged SPEC surfaces → RETAINED + marker.

### M3 — `internal/design` investigation + disposition
- Priority: Medium.
- Investigate: read `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` scope (did it intend to keep the Go
  DTCG validator?); check whether any live rule/config consumes DTCG validation; reachability;
  `@MX` scan (REQ-DPL-011 / design constitution references may themselves be retired).
- Disposition per verdict (hypothesis: GENUINELY DEAD): remove package + `_test.go` if confirmed
  dead; else RETAINED + marker.

### M4 — `internal/i18n` investigation + disposition
- Priority: Medium (relocation into github test helper).
- Investigate: confirm only `internal/github/*_test.go` consume `NewCommentGenerator`/`CommentData`;
  determine whether github *production* code should call it (wiring gap → out of scope, flag
  follow-up) vs test-only helper.
- Disposition per verdict (hypothesis: RELOCATE): move `errors.go`/`templates.go` into a
  `internal/github` test helper, update the github test call sites, remove the emptied
  `internal/i18n`. If a production wiring gap is confirmed → RETAINED + flag follow-up SPEC.

### M5 — `internal/migrate` investigation + disposition
- Priority: Medium (relocation into the LIVE `internal/migration` framework; highest coupling).
- Investigate: decide if "future observability tooling" is a genuine staged plan or aspirational;
  read the `internal/migration` framework shape to find the correct migration-step home.
- Disposition per verdict (hypothesis: RELOCATE): fold `CleanupUserSettings` into a migration
  step under `internal/migration` (preserving atomic-write safety, REQ-DPI-013), update the
  `@MX:REASON` referent in `internal/hook/retired_events.go`, remove the emptied
  `internal/migrate` + its `_test.go`. If genuine staged plan → RETAINED + marker.

### M6 — Cross-package verification + verdict catalog
- Priority: High (closure gate).
- Verify: full cross-platform build green; `go test ./...` passes; `internal/runtime/gobin`
  untouched (diff check); no dangling `@MX:REASON` referents; every package has a recorded
  verdict + evidence in `progress.md` §E.2.
- Produce a final verdict catalog table (package → verdict → disposition → evidence anchor).

## §G. Anti-Patterns (do NOT)

- AP-1: Deleting a package because grep shows no importer, without the `git log` + `@MX` +
  owning-SPEC checks (the era=NONE 29-SPEC trap).
- AP-2: Treating an `@MX:REASON` retention claim as proof of life without disconfirming it
  against the reachability grep.
- AP-3: Removing a package but leaving its `_test.go` files or a dangling `@MX:REASON` referent.
- AP-4: Touching `internal/runtime/gobin`.
- AP-5: Building new production wiring to "rescue" a package (out of scope — flag a follow-up).

## §H. Cross-References

- spec.md §B (REQ-DPI-001..013), §C (per-package hypotheses).
- acceptance.md §D (AC matrix).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §5.
