---
id: SPEC-CODEX-LANE-SLOTS-001
title: "Codex factory lane slot-claim repair: capacity-aware auto-growth, default capacity policy, shared lane pool, and legacy-path parity"
version: "0.1.0"
status: in-progress
created: 2026-09-30
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "factory,lane,slots,codex,claim,legacy-label,performance,reconcile"
tier: M
---

# SPEC-CODEX-LANE-SLOTS-001

## §A — History

- 2026-09-30: Card t1378 opened from an operator-measured defect. With factory run `tm3yoq` active (created without an explicit `-f N` size, so its join bound reads as 1..1) and a Claude lane occupying `lane-1`, `moai codex -f lane` fails with `factory run tm3yoq has no free lane slots in 1..1` instead of attaching as `lane-2`.
- 2026-09-30: Leader scope addition — after the rc.23 update (binary built from `d194083fb`), the same command no longer fails immediately; the launch takes a long time and eventually fails. Two observed behaviors must be explained: the immediate upper-bound refusal (pre-update build) and the long-launch failure (rc.23). Diagnosis items ①–④ folded into this SPEC; `moai factory runs` cost (item ④, measured 3.0s user / 1.86s sys) is in repair scope as one card with the original axes.
- 2026-09-30: Plan-phase premise verification executed against this tree (see `progress.md` § Premise Verification). Root cause CONFIRMED at `internal/kanban/factory_slots.go:246-256`.

## §B — Requirements

### §B.1 Root defect — capacity-blind automatic slot claim

- REQ-001 (Event-driven): **When** a factory lane joins a run in automatic label selection **While** the run's join bound is smaller than the highest live claimed lane number, the claim engine shall grow the automatic scan bound to one past the highest live claimed lane number and claim the first free slot in the grown range.
- REQ-002 (Ubiquitous): The automatic growth in REQ-001 shall never reuse a slot already held by a live claim of any backend.
- REQ-003 (Event-driven): **When** an explicit `lane-<n>` request names a slot outside the join's allowed range, the claim engine shall refuse the request with the out-of-range refusal, and shall not grow the bound for explicit requests.

### §B.2 Default capacity policy for runs created without `-f N`

- REQ-004 (Ubiquitous): A factory run shall record its declared lane capacity at leader start — the explicit `-f N` count when the operator supplied one, and an explicit derived-capacity marker when the operator did not.
- REQ-005 (Event-driven): **When** a lane joins a run whose recorded capacity is a derived-capacity marker, the join shall treat the run as capacity-open and apply the growth rule of REQ-001 without an upper refusal.
- REQ-006 (Event-driven): **When** a lane joins a run whose recorded capacity is an explicit operator-declared count and all declared slots hold live claims, the join shall refuse the automatic claim with the current no-free-slot refusal naming the declared range — the full explicit run never grows (t1294 bound contract preserved for operator-declared capacity).

### §B.3 Shared lane pool across backends

- REQ-007 (Ubiquitous): Lane numbers shall belong to a single run-scoped pool shared by every backend; a codex lane and a Claude lane in the same run shall draw from the same numbering and shall never hold the same number.
- REQ-008 (Ubiquitous): The claim engine shall remain backend-blind: the join of a codex lane into a run holding live claims recorded by another backend shall follow the identical selection, growth, and refusal rules as any other join.

### §B.4 Legacy-label interaction (t1109 parity)

- REQ-009 (Event-driven): **When** a claim scan encounters a live record carrying a legacy label (`worker-<n>` / `agent-<n>`) stamped with a run id, the claim engine shall refuse the whole claim with the legacy-run refusal naming the legacy value, the run id, and the retire step — on the bounded join path exactly as on the unbounded path.
- REQ-010 (Event-driven): **When** an explicit request uses a legacy label spelling, the claim engine shall refuse it naming the canonical replacement — unchanged from current behavior.
- REQ-011 (Ubiquitous): A legacy record whose run id is empty shall be neither refused nor counted as a lane — unchanged from current behavior.

### §B.5 Launch-latency diagnosis and `moai factory runs` cost (leader items ①–④)

- REQ-012 (Event-driven): **When** a codex lane launch's pre-exec phase exceeds the operator-configurable slow-launch threshold, the launcher shall print one timing line per pre-exec step, naming each step and its wall-time. The threshold's default value lives in `internal/config/defaults.go` — never an inline literal at the call site. (Trigger definition is operator-configurable by design; the report capability itself is unconditionally testable.)
- REQ-013 (Ubiquitous): The `moai factory runs` owner-classification shall probe process identity through the platform process-fingerprint seam — all of its build-tagged platform variants — with per-listing probe work bounded, not once per run row.
- REQ-014 (Ubiquitous): Every latency or probe-count measurement this SPEC requires shall be recorded as command-plus-verbatim-output in the SPEC's progress evidence, attributed to the tree measured.

## §C — Acceptance Criteria (summary)

Full Given-When-Then scenarios live in `acceptance.md`. Binary-testable summary:

| AC | Verifies | Class | Mechanical check |
|----|----------|-------|------------------|
| AC-001 | Repro sequence: full 1-slot run + live lane-1 → codex-shape auto join claims lane-2 | release-blocking | `go test ./internal/kanban/...` new test; RED-now: acceptance.md ledger RED-1 |
| AC-002 | Explicit out-of-range request still refused | regression-guard | same package, new test |
| AC-003 | Full explicitly-sized run still refuses (t1294 contract) | regression-guard | same package, new test |
| AC-004 | Live legacy record refuses with legacy-run refusal on the bounded path | regression-guard | existing `role_naming_m1_test.go` tests keep passing + new bounded-path test |
| AC-005 | Legacy-label explicit request refused naming canonical; empty-run-id legacy ignored | regression-guard | existing tests keep passing |
| AC-006 | Run capacity recorded at leader start (explicit vs derived) | release-blocking | targeted `internal/cli` test; RED-now: ledger RED-2 |
| AC-007 | Owner probe bounded per listing invocation, platform variants pinned | release-blocking | instrumented test in `internal/homestate`; RED-now: ledger RED-4 |
| AC-008 | Whole-package re-measurement green | regression-guard | `go test ./internal/kanban/...` exit 0 |
| AC-009 | Shared run-scoped lane pool: a codex-shape bounded join draws a distinct number while a claude-shape claim holds lane-1 | release-blocking | `go test ./internal/kanban/...` new test; RED-now: ledger RED-1 (shared) |
| AC-010 | Per-step pre-exec timing report names every pre-exec step | release-blocking | targeted launcher test; RED-now: ledger RED-3 |

Classification discipline: release-blocking criteria carry an executed RED-now observation (acceptance.md § Evidence Ledger) and name the milestone that flips them; regression-guard criteria preserve already-correct behavior (no red is reproducible today) and are never recorded as passes before the green lands.

## §D — Constraints

- Go conventions of this repo: error wrapping `fmt.Errorf("operation: %w", err)`; no hardcoded URLs, models, or thresholds — defaults live in `internal/config/defaults.go`, env names in `internal/config/envkeys.go`; all code, comments, and godoc in English.
- Tests: `t.TempDir()` for every temp directory; never `t.Setenv` with OTEL variables; tests that read lane env vars must pin every axis they read (lane env leaks falsify such tests locally); the legacy-run error path tests must keep passing unmodified in intent.
- Re-measurement unit: the whole `internal/kanban` package (`go test -timeout 30m ./internal/kanban/...`) plus targeted builds/tests of touched packages; no local full-suite runs.
- The default lane-count constant (`config.DefaultFactoryLeaderLanes`) is a config-layer default — any policy change to its value or role goes through `internal/config/defaults.go`, never an inline literal at the call site.

## §E — Design Notes (from plan-phase diagnosis)

- The join-side bound is not run state today: the codex twin passes the constant `config.DefaultFactoryLeaderLanes = 1` (`internal/cli/codex_factory.go:140-141`, `internal/config/defaults.go:648`) while the Claude twin passes an unbounded claim (`internal/cli/factory.go:761`). The capacity record (REQ-004) is what turns the join bound from a launcher-side guess into run state. Milestone ordering and the option analysis live in `plan.md`.
- Diagnosis findings ①–④ with file:line evidence are recorded in `progress.md` § Premise Verification and `plan.md` §B.

## §F — Out of Scope

### Out of Scope — t1294 boundary (launch_pending visibility, argv/state binding)

- Queue card t1294 (near-duplicate signal p=0.84) owns the broader "Codex F2 session/slot misassignment regression prevention" theme: `launch_pending` endpoint visibility and the argv/state binding checks (the `registerFactoryLaunchPending` surface at `internal/cli/codex_direct_posix.go:47`). This SPEC touches the slot-claim selection rules and capacity policy only; it does not alter launch-pending registration, rollback, or binding verification.

### Out of Scope — worktree ghost-directory hygiene

- Unregistered ghost directories under `.claude/worktrees/` (e.g. `t1363`) are state hygiene owned by the worktree sweep tooling lineage (SPEC-WORKTREE-SWEEP-001). No directory scan was found in the codex launch pre-exec path (see `progress.md` finding ②); this SPEC adds no worktree-scanning behavior.

### Out of Scope — Codex hook re-approval UX

- The "Codex stops changed hooks until they are re-approved" behavior (`internal/codexwiring/codexwiring.go:74`) acts inside the codex process after the launcher replaces itself — on the POSIX path via `syscall.Exec` (`internal/cli/codex_direct_posix.go:53`, `//go:build !windows`); Windows uses its own exec mechanism (`internal/cli/codex_direct_windows.go`). The moai launcher cannot gate on or accelerate codex's own trust prompt. Surfacing the untrusted-hooks state in the launcher readout is a candidate follow-up card, not this SPEC.

### Out of Scope — backend-prefixed lane namespaces

- Splitting lane numbers into per-backend namespaces (e.g. `codex-lane-N`) would break leader routing that addresses lanes by label and contradict REQ-007. Explicitly rejected; see `plan.md` §E.

## §G — Cross-References

- `SPEC-FACTORY-LANE-JOIN-SOCKET-001` — the shared lane-join gate (`enterFactoryLaneRun`) this SPEC's join path flows through.
- `SPEC-CODEX-FACTORY-RETIRE-001` — the retire surface the legacy-run refusal names.
- Card t1294 — the bound that introduced the observed 1..1 refusal (`5ee35b2f5`); its regression-prevention remainder stays with the queue card, not this SPEC.
