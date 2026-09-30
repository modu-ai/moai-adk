# SPEC-STALE-RUN-LABEL-001 — Progress

## Plan Phase (2026-09-30, card t1373)

- Status: draft — spec.md / plan.md / acceptance.md authored by manager-spec, then repaired at audit iteration 1.
- Research: read-only; mechanism verified in this worktree (`internal/hook/factory_messages.go`, `internal/hook/session_stale_run.go`, `internal/hook/session_start_factory.go`, `internal/factorymsg/store.go` `ValidateActiveRun`, `internal/homestate/factory_run_retire.go` `retireRun`, `internal/cli/codex_factory.go` env stamp, `internal/kanban/bootstrap.go` legacy vocabulary).
- Root cause summary: legacy-label branch short-circuits to the retire prescription before any run-state measurement; env residue survives /clear because the identity lives in the session process env (plus tmux pane env); no unbind path exists. Worker-70 separation: inbound Claim path is peer-record-keyed and never reads the env label.
- SPEC ID self-check: `SPEC-STALE-RUN-LABEL-001` regex PASS (executed), uniqueness confirmed against `.moai/specs/`.
- Next: plan-audit iteration 2 (delta scope D1-D5), then M1 (run-state-gated prescription + unbind state).

## Audit record

- **Iteration 1 — FAIL, score 0.81** (`.moai/reports/t1373/plan-audit-r1.md`, audited_sha d194083fb, 2026-09-30). 5 blocking defects (D1 carrier-premise false, D2 sweep instrument report-not-verdict, D3 unmapped release-blocking test, D4 key-set unenumerated, D5 tier field missing) + 3 optional (D6, D7, D8 — D8 no fix required).
- **Iteration 2 fix pass (spec/plan/acceptance 0.2.0, this update)**:
  - D1: §C.1 carrier sentence rewritten to the measured fact (settings.local.json carries NO factory keys — auditor measured 0 `MOAI_*` keys live, no writer in the tree; residue is process env + tmux pane env). Took the auditor's preferred **DROP branch** for the persisted-carrier scrub requirement (old REQ-SRL-005): no writer exists to guard against (Enforce Simplicity). Cascaded: acceptance AC-SRL-005 shrinks to unbind-notice + re-bind line; plan M2 scrub item removed; spec Out of Scope gains the dropped-scrub bullet.
  - D2: AC-SRL-009 / E5 instrument replaced with diff-scoped added-line extraction (`TestNoNewEnvNameLiteralsInDiff`, added-line count logged); measured baseline 28 total (hook+factorymsg 12, cli 16) pinned as the RED cell with its false-positive sources.
  - D3: `TestPrescriptionGateUnavailableFailsOpen` added to plan M1 (M1.1 + M1 test list); AC matrix Test column now carries owning milestones for all 9 ACs.
  - D4: dissolved by the D1 drop (no key-set to enumerate).
  - D5: `tier: M` added to spec.md frontmatter.
  - D6 (optional): REQ-SRL-003 clarifying parenthetical — the accessor's own DB-file probe is part of the shared measurement; hook-side re-derivation from broker-file absence is the prohibited act.
  - D7 (optional): plan M1.2 clause — the once-per-session-identity dedup carrier is shared across ALL prescription surfaces (SessionStart bootstrap + UserPromptSubmit peer path), so turn 1 cannot emit twice.
  - Renumbering: old REQ-SRL-006..010 → REQ-SRL-005..009 (sequential, no gaps); AC IDs unchanged AC-SRL-001..009 with mappings updated (005→REQ-005/006, 006→REQ-007, 007→REQ-008, 008→REQ-003, 009→REQ-009).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M set) under `.moai/specs/SPEC-STALE-RUN-LABEL-001/`.
- Frontmatter: 12 canonical fields present + `tier: M`; `status: draft`; version 0.2.0.
- Out of Scope: four H3 topics including the explicit card t1345 exclusion and the audited dropped-scrub note.
- Open items for audit iteration 2: RED-now cells for the Go-test release-blocking ACs are scheduled at M1 RED (plan-phase author cannot execute unexported-function tests without writing test files); AC-SRL-009's RED cell is already measured and pinned (auditor baseline 28/12 @ d194083fb); unbind-state carrier (session-record vs factorymsg peer marker) remains the M1 review decision, now with the cross-surface sharing clause (D7) attached.

## §E.2 Run-phase Evidence

### Run-phase environment attribution (read first)

This run executed in the Claude-isolated agent worktree `.claude/worktrees/agent-a703d55100343647b` on branch `WT-stale-run-gate` (renamed in place from the runtime agent branch), because the worktree-isolation guard refuses every cross-tree operation from this session — measured at session start: `git -C` toward the card tree was refused ("a worktree-isolated agent's git operations must target its own worktree") and a Write to the card tree was refused ("Edit the worktree copy"). The card tree `.moai/worktrees/t1373` still holds the plan-phase SPEC copies; the authoritative run-phase copies are the ones committed on `WT-stale-run-gate` — byte-identical clones (`cmp`-verified all four) that then received only the sanctioned M1 frontmatter transition (spec.md `status:`/`updated:`; all bodies untouched, audit r2 hash subjects preserved). Every test invocation ran env-scrubbed as ONE compound `unset <MOAI factory/kanban vars> && go test ...` invocation — this session carries the live factory run tm3yoq lane env (measured: `MOAI_KANBAN_ID=tm3yoq`, `MOAI_FACTORY_WORKER=lane-1`), which falsifies env-reading guard tests. Note: the dispatch's `env -u` prefix form was refused by this environment's worktree guard ("cannot be shown not to be git"); the repo-canonical `unset <VARS> && <command>` single-invocation form (kanban-dispatch.md § Verification load is lane-local) was used instead — same scrub semantics, one process.

### M1 RED cells — observed 2026-10-01 on tree `a1a22b919`

Command (single compound invocation, verbatim; exit code 1):

```
unset MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_FACTORY_ROLE MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -v -run 'TestStaleRunNoticeSilentWhenRunRetired|TestStaleRunNoticeFiresWhenRunActive|TestStaleRunNoticeOncePerSession|TestUnbindNoticeThenSilence|TestUnbindNoticeRebindLinePresence|TestPrescriptionGateUnavailableFailsOpen|TestClearSourceDeadRunEnvYieldsUnbound|TestInboundClaimIndependentOfEnvLabel' ./internal/hook/
```

Evidence ledger (verification-completeness.md §2.1 four-element cells; carrier = this ledger, cited by id):

- **RED-AC-SRL-001** — test `TestStaleRunNoticeSilentWhenRunRetired`; expected red: the legacy branch prescribes a retire for a measured-retired run; observed stdout: `stale_run_gate_test.go:71: peer path prescribed a retire for a measured-retired run: "stale run: lane label \"worker-69\" is legacy vocabulary ... 'moai factory runs --retire srl-retired', then relaunch"`; exit 1; tree `a1a22b919`. Right-reason red: the branch returns at `factory_messages.go:66-68` before any run-state measurement.
- **RED-AC-SRL-003** — test `TestStaleRunNoticeOncePerSession`; expected red: the prescription repeats on the second turn of the same session identity; observed: `stale_run_gate_test.go:105: second turn repeated the prescription: "stale run: lane label \"worker-69\" ... 'moai factory runs --retire srl-once', then relaunch"`; exit 1; tree `a1a22b919`. Right-reason red: no dedup carrier exists.
- **RED-AC-SRL-005a** — test `TestUnbindNoticeThenSilence`; expected red: the unbind path does not exist, the first turn emits the prescription; observed: `stale_run_gate_test.go:127: unbind notice carries the retire prescription: "stale run: ... 'moai factory runs --retire srl-unbind', then relaunch"`; exit 1; tree `a1a22b919`.
- **RED-AC-SRL-005b** — test `TestUnbindNoticeRebindLinePresence`; expected red: the unbind path does not exist, the answer is the prescription; observed: `stale_run_gate_test.go:152: unbind path prescribed a retire: "stale run: ... 'moai factory runs --retire srl-dead', then relaunch"`; exit 1; tree `a1a22b919`.
- **RED-AC-SRL-008** — test `TestPrescriptionGateUnavailableFailsOpen`; expected red: with an unmeasurable run state (corrupt factory.db) the branch still prescribes; observed (re-measured — the first RED run of this test used a hand-built DB path that missed `homestate.FactoryDBPath`'s actual layout, so its premise measured not-active rather than unavailable; the corrected setup was re-measured RED on the same pre-implementation tree before any implementation existed, with the working-tree gate files moved aside and the three call-site files restored to `a1a22b919` for the measurement): `stale_run_gate_test.go:187: unmeasurable run state still prescribed a retire: "stale run: lane label \"worker-69\" is legacy vocabulary ... 'moai factory runs --retire srl-unavailable', then relaunch"`; exit 1; tree `a1a22b919`. Right-reason red: no measurement precedes the prescription.
- **RED-AC-SRL-004** — test `TestClearSourceDeadRunEnvYieldsUnbound`; expected red: the `/clear` boundary's SessionStart peer path prescribes for a dead-run label; observed: `stale_run_gate_test.go:197: the /clear boundary prescribed a retire for a dead-run label: "stale run: ... 'moai factory runs --retire srl-clear', then relaunch"`; exit 1; tree `a1a22b919`.
- **GREEN-AT-ARRIVAL-AC-SRL-002** — test `TestStaleRunNoticeFiresWhenRunActive`; observed: `--- PASS: TestStaleRunNoticeFiresWhenRunActive (0.65s)` on tree `a1a22b919`, pre-gate. This is the acceptance matrix's own positive control: the unconditional branch fires regardless of state, so no honest red exists on any pre-gate tree — a red cell here is fabricable only. Recorded per the acceptance contract's letter as regression-guard-class evidence (red not observed at M1), while noting its release-blocking FUNCTION: it is the anti-suppression control proving the M1 gate measures state rather than blanket-suppressing, and it is re-exercised in the final matrix.
- **GREEN-AT-ARRIVAL-AC-SRL-006** — test `TestInboundClaimIndependentOfEnvLabel`; observed: `--- PASS: TestInboundClaimIndependentOfEnvLabel (0.42s)` on tree `a1a22b919`, pre-gate. Same green-at-arrival classification: the Claim path never reads the env label today (worker-70 separation is the existing behavior being preserved); recorded as regression-guard-class evidence guarding REQ-SRL-007 against coupling, re-exercised in the final matrix.

AC-SRL-009's RED cell is the audit-pinned baseline (acceptance.md, whole-tree 28 literals / hook+factorymsg 12, measured by the auditor at `d194083fb`) — no M1 re-measurement is required or performed for it; its test lands at M3 with the mutate probe.


## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
