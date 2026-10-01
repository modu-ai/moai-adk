# SPEC-STALE-RUN-LABEL-001 — Implementation Plan

## A. Context

Card t1373. Three measured cases (2026-09-30): a retired factory run (tlwgk9) whose lane sessions (worker-69/72) still carry its identity env; the hook prescription "end this session, retire the run with 'moai factory runs --retire tlwgk9', then relaunch" fired on every turn although the retire had already been performed; 2h40m of session work lost, 3 lanes locked out. Worker-70, with the same stale env, kept receiving dispatches — the message channel and the env-label judgment are separate code paths, and that separation is a preserved invariant (REQ-SRL-007).

Root cause and code map are in `spec.md` §C. Key files:

| File | Role |
|---|---|
| `internal/hook/factory_messages.go:54-71` | `registerFactoryHookPeer` — legacy-label branch short-circuits to the prescription before any run-state measurement |
| `internal/hook/session_stale_run.go` | `staleRunNoticeFor` / `legacyFactoryHookNotice` / `staleRunNotice` — the unconditional prescription text (4 locales); `"stale run:"` prefix is verbatim per locale (assertion anchor) |
| `internal/hook/session_start_factory.go:50-64` | SessionStart factory bootstrap — same legacy short-circuit |
| `internal/hook/session_start.go:466` | SessionStart → `registerFactorySessionStartPeer` wiring |
| `internal/hook/user_prompt_submit.go:154` | UserPromptSubmit → `registerFactoryUserPromptPeer` wiring (the every-turn surface) |
| `internal/hook/factory_messages.go:130-188` | `factoryHookBatch` — inbound Claim path, peer-record-keyed, MUST NOT change (REQ-SRL-007) |
| `internal/factorymsg/store.go:150-172` | `ValidateActiveRun` — the existing run-state accessor (`runs.status`), shared-accessor anchor for REQ-SRL-003 |
| `internal/homestate/paths.go:238` | `FactoryDBPath` — the factory state DB |
| `internal/cli/factory.go` / `internal/homestate/factory_run_retire.go:321-328` | retire path — writes `status='retired'` (runs DB only) |
| `internal/cli/launcher.go:1062` | `readSettingsLocalForLaunch` — audited at iteration 1: this carrier holds NO factory keys (no writer exists; live carrier measured 0 `MOAI_*` keys mid-run, plan-audit-r1.md §2); NOT a change target of this SPEC |
| `internal/config/envkeys.go` | env-name constants (REQ-SRL-009 carrier) |

## B. Known Issues

- The legacy-label branch predates the run-state gate: `ValidateActiveRun` exists and is reachable in the same function (used for current-vocabulary labels at `factory_messages.go:88`), but the legacy branch at lines 64-68 returns before it.
- The prescription fires on both SessionStart and UserPromptSubmit; the every-turn repetition is driven by the UserPromptSubmit path (`user_prompt_submit.go:154`).
- A "degraded" answer currently also exists on this path ("factory messaging degraded: NO_ACTIVE_FACTORY") for current-vocabulary labels against a dead run — the repair must give the legacy branch a coherent answer in the same dead-run case, not a second inconsistent one.
- Session identity for prescription dedup (REQ-SRL-002): `HookInput.SessionID` is stable across a session identity; a `resume`/`compact` keeps the identity while `clear` starts a new one. The dedup carrier must key on session identity, not process identity.
- The persisted settings carrier (`.claude/settings.local.json`) is NOT part of the residue (audited, plan-audit-r1.md §2): the retire path writes only the runs DB, no code path in this tree writes factory identity into that carrier, and the live carrier carries zero `MOAI_*` keys mid-run. The identity-persistence axis is child-process env stamped at launch plus tmux command-scoped assignments scrubbed-then-restamped per pane launch — the surface REQ-SRL-004's judgment layer covers.

## C. Pre-flight

1. Re-read `runs.status` schema and confirm the retired value is exactly `retired` (`internal/homestate/factory_run_retire.go:328`).
2. Confirm `factorymsg.ValidateActiveRun` is safe to call on the SessionStart/UserPromptSubmit hot path (200ms inspection budget, `factoryHookInspectionDeadline`) — it is a single indexed `SELECT` with a 100ms busy timeout; if it cannot complete inside the budget, the hook fails open to the existing degraded answer, never to a prescription.
3. Inventory every call site that renders the retire prescription (grep `runs --retire` under `internal/`) so the gate lands on all of them, not only the UserPromptSubmit one.

## D. Constraints

- Error wrapping `fmt.Errorf("operation: %w", err)`; English comments; no hardcoded env names or thresholds (envkeys.go / defaults.go single source — REQ-SRL-009).
- Template neutrality: this SPEC touches Go hook/CLI code only. No file under `internal/template/templates/` changes; no card provenance or SPEC-id strings enter the template tree.
- Tests run env-scrubbed in one compound invocation (lane env falsifies env-reading guard tests locally): `unset MOAI_FACTORY_WORKERS MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL && go test ./internal/hook/...` — affected packages only, no full suite.
- All acceptance tests use `t.TempDir()` factory DBs; no live factory run is required.
- Hooks fail open: every new measurement failure degrades to the existing degraded answer or silence, never to a prescription and never to a hook error.

## E. Self-Verification

- E1: AC matrix PASS/FAIL with verbatim test output (`acceptance.md`).
- E2: `go vet` + `golangci-lint run` on changed packages, verbatim output.
- E3: Coverage on `internal/hook` and `internal/factorymsg` for the touched paths (target: package gate per `quality.yaml`).
- E4: Separation grep — the diff adds no reference to env-label reading inside `factoryHookBatch`'s call tree (mechanical diff review + dedicated AC-SRL-006 test).
- E5: Diff-scoped env-literal instrument (AC-SRL-009, `TestNoNewEnvNameLiteralsInDiff`): added-line extraction over `git diff <card-base>...HEAD` asserting 0 literals, added-line count logged; baseline 28/12 pinned in acceptance.md as the RED cell.

## F. Milestones

Priority order; the first two milestones carry the decisions most likely to change in review (state semantics, dedup carrier), the last is mechanical wiring.

### M1 (High) — Run-state-gated prescription + unbind state (data-model-first)

The highest-change-likelihood decisions live here: what "unbound" means on disk, and where the one-time-notice dedup state is carried.

1. Extend the factorymsg/kanban state layer with a shared run-state query the hooks use for the prescription gate (REQ-SRL-003) — either widen `ValidateActiveRun` with a tri-state result (active / not-active / unavailable) or add a sibling accessor; the tri-state is the recommended shape because "unavailable" (inspection budget spent, DB busy) must fail open differently from a measured `retired`. Test: `TestPrescriptionGateUnavailableFailsOpen` (AC-SRL-008, release-blocking) belongs to this step.
2. Define the unbind state carrier (REQ-SRL-005): one row/record keyed by session identity recording "unbind notice emitted at <turn>" — recommended carrier is the existing session-record store (`kanban.Read` family) or a factorymsg peers-side marker, decided at M1 review; the carrier is the reviewable decision. The carrier is SHARED across ALL prescription surfaces of the session identity — SessionStart bootstrap AND UserPromptSubmit peer path — so turn 1 cannot emit the notice twice (startup + first prompt). The same carrier serves the REQ-SRL-002 once-per-session prescription dedup.
3. Gate the legacy branch in `registerFactoryHookPeer` (`factory_messages.go:64-68`) and the SessionStart siblings on the measured state: not-active → unbind path (one notice, then silence); active → prescription once (dedup per REQ-SRL-002); unavailable → current degraded answer, never a prescription.
4. RED-first tests: `TestStaleRunNoticeSilentWhenRunRetired`, `TestStaleRunNoticeFiresWhenRunActive` (positive control), `TestStaleRunNoticeOncePerSession`, `TestUnbindNoticeThenSilence`, `TestPrescriptionGateUnavailableFailsOpen` — each on a `t.TempDir()` factory DB, env-scrubbed compound invocation. RED-now cells recorded and pinned here per `verification-completeness.md` §2.1.

### M2 (High) — /clear boundary + unbind notice re-bind line

1. SessionStart source=clear path (REQ-SRL-004): the fresh session's factory identity evaluation consults run state before any bind or notice; a dead-run label yields the unbound semantics, not a carried-over binding. Test: `TestClearSourceDeadRunEnvYieldsUnbound`.
2. Unbind notice re-bind line (REQ-SRL-006): when an active run exists in the same root, the notice names the `moai cc -f lane-<n>` join; otherwise it omits the line. Test: `TestUnbindNoticeRebindLinePresence`. (The persisted-carrier scrub of the former M2.2 was DROPPED at audit iteration 1 — no writer produces factory keys in that carrier; plan-audit-r1.md §2.)

### M3 (Medium) — Separation guard + sweeps (mechanical)

1. AC-SRL-006 separation test: a session with a live broker peer record receives claims while its env carries a legacy label for a dead run — `factoryHookBatch` behavior unchanged (`TestInboundClaimIndependentOfEnvLabel`).
2. Diff-scoped env-literal instrument (REQ-SRL-009 / AC-SRL-009): `TestNoNewEnvNameLiteralsInDiff` per the instrument restatement in acceptance.md — baseline 28/12 pinned as RED cell, mutate probe executed before adoption.
3. REQ-SRL-008 behavior-preservation test for the current-vocabulary bind path (`TestCurrentVocabularyBindPathUnchanged`).
4. Full AC matrix re-run, E1-E5 self-verification, lint, vet.

## G. Anti-Patterns

- Do NOT gate `factoryHookBatch` (the Claim path) on env labels — that is the worker-70 regression (REQ-SRL-007).
- Do NOT infer run state from broker-file absence (`messages/<run>/broker.db` missing) — a not-yet-flushed active run looks identical; measure `runs.status` through the shared accessor (the accessor's own DB-file probe is part of the shared measurement).
- Do NOT emit the unbind notice once per PROCESS — after `/clear` the session identity is new and one notice is correct again; key dedup on session identity, shared across all prescription surfaces.
- Do NOT fail the hook (exit != 0) on run-state measurement failure — fail open to the existing degraded answer.
- Do NOT scrub keys from a live session's process environment (impossible from a hook; see spec §F Out of Scope).
- Do NOT add a persisted-carrier scrub for `.claude/settings.local.json` — audited (plan-audit-r1.md §2): no writer in the tree puts factory keys there; a scrub guards a path nothing produces.
- Do NOT hardcode `retired` as a literal at the gate — use the shared accessor's tri-state so a future status vocabulary change lands in one place.

## H. Cross-References

- `spec.md` §C root cause, §D requirements
- `acceptance.md` AC matrix and two-cell discipline
- `.moai/reports/t1373/plan-audit-r1.md` — audit iteration 1 (FAIL 0.81; D1-D5 blocking resolved in spec/plan/acceptance 0.2.0)
- SPEC-ROLE-NAMING-CODE-001 (vocabulary, REQ-RNC-022 refusal), SPEC-FACTORY-RUN-RETIRE-001 (retire semantics, runs DB), SPEC-FACTORY-SELF-DISPATCH-001 (lane env stamping, clear policy), card t1345 (out-of-scope self-healing card)
