# card-review — t1630 (SPEC-USER-SETTINGS-PROTECT-001)

## Round 1 — 2026-10-14

- Backend: codex (mcp codex_review, scope=card), base ec193e4777, HEAD 4d72317f4e, advisory=true.
- Verdict: **fail** — 1 finding (P1).
- Finding P1 (`internal/config/toolpolicy/tier_render.go:132`): the sidecar `managed-allow.json` is read but never written; a policy-generated allow entry later removed from the policy document is misclassified as user-added and persists (revoked-permission resurrection). Review basis: static comparison only.
- Lane verification: confirmed by grep — zero writer sites for `last_generated`/`managed-allow` in internal/ (non-test); the code comment at tier_render.go:106-113 documents the drop semantics that the missing writer prevents from ever firing.
- Disposition: **fixed in this card** — the finding is the missing half of the decided Q5 union rule (AC-005), not new behavior; revoked-permission persistence is the exact hazard class this SPEC exists to prevent; sync has not started, so repair now is cheaper than a follow-up card. Repair delegated to manager-develop (persist `last_generated` after a successful policy render/write + RED-first tests including the removal-drop deciding test). Re-review round 2 to follow within the 2-review ceiling.

## Round 2 (re-review #1 of 2) — 2026-10-14

- Repair landed: dd1285d066 (persistManagedAllowRecord + managedAllowRecordPath, called after a successful PROJECT-scope render inside RenderTierPermissions) + 534efa0e6e (§E.2 evidence).
- Repair evidence: RED pre-fix observed (record absent; removed generated entry survived — allow [a b] want [a]); GREEN post-fix (record written/refreshed; removed entry dropped; failed render leaves record untouched); AC-005 both cases PASS unmodified; families ok; builds darwin+windows exit 0; toolpolicy coverage 90.4%; lint 0; gofmt clean.
- Lane re-verification: commits present, tree clean, writer site present at tier_render.go:95/117, deciding drop test + AC-005 re-run ok.
- Re-review dispatched (codex_review scope=card, background); result appended below when it returns.

## Round 2 result — 2026-10-14

- Verdict: **fail** — the round-1 P1 is resolved (no longer reported); two new P2s on the repair's completeness:
  - P2-1 (`tier_render.go:95`): settings-first ordering — a failed sidecar save after the settings write leaves new allow entries outside the record → later removal misclassifies them as user-added (persist).
  - P2-2 (`tier_render.go:174`): the `BuildInto` path (`moai tool-policy build`) generates the same project allow list without refreshing the record — same hazard via that path.
- Disposition: **fixed in this card** — both are completeness gaps of the decided record-maintenance rule, in scope, small; one re-review remains within the 2-review ceiling. Repair re-delegated to the same manager-develop agent (record-before-settings ordering + BuildInto wiring, RED-first).

## Round 2 repair + Round 3 (final re-review #2 of 2) — 2026-10-14

- Round-2 repair landed: 026ba839a2 (record-ahead staging via stage/converge in RenderTierPermissions + BuildInto TargetJSON record maintenance in codegen.go) + 5d6d54092b (§E.2 evidence). Evidence: 4 RED observed pre-fix, 4 GREEN post-fix (record-stage failure leaves settings untouched; settings-write failure classifies correctly next render; BuildInto maintains record, removal drops; failed render leaves safe staged union), round-1 families + AC-005 PASS, builds exit 0 both platforms, coverage 91.0%, lint 0, gofmt clean.
- Lane re-verification: commits present, tree clean, stage/converge + BuildInto write sites present, two deciding tests re-run ok.
- Residual named for leader (§E.2): the two-file update is non-atomic — a settings-write failure coinciding with a policy revocation in the SAME render window can persist the revoked entry as user-added; closing it fully needs compensating record recovery on the settings-failure path (undecided surface).
- Round 3 (final re-review, ceiling 2/2) dispatched in the background; result appended when it returns. Ceiling consequence: PASS → sync stage; still-fail → open-finding disposition written and raised to the leader.

## Round 3 result (final re-review #2 of 2 — ceiling reached) — 2026-10-14

- Verdict: **fail** — 2 findings. No further re-review will run (ceiling); the sync-auditor reads the final state.
- Finding A (P1, `internal/cli/init_autonomy_wiring_test.go:157`): the M1 keep-rule breaks the pre-existing init wiring test. **Lane-measured RED** (`go test -run TestRunInit_WizardAutonomyTierAppliesBundle ./internal/cli/` → `USER permissions.defaultMode = "default", want "auto"`) — a real CI-blocking regression hidden by the t1542 cli-suite gap. Disposition: **fixed** — test expectation updated to the preserved contract + absent-key subtest + repo-wide class sweep; repair delegated (no re-review; verification by targeted runs).
- Finding B (P2, `tier_render.go:184`): staging records a user-added entry C as managed before the settings write; a settings-write failure plus a later policy exclusion of C deletes the user's own entry on the next render — the sharpened form of the repair's own named non-atomicity residual (failure-window only). Disposition: **carried to leader** — undecided failure-window surface; raised with the completion report.

## Round 3 repair closure — 2026-10-14

- Finding A repaired: db70a4266b (test-only: TestRunInit_WizardAutonomyTierAppliesBundle re-scoped to two subtests — existing-defaultMode-preserved, absent-defaultMode-applies-tier) + c4317ef4b3 (evidence). Class sweep: 1 stale fixed; flag/non-interactive/degraded tests confirmed correct (fresh-HOME absent-key paths); launcher/cg/doctor hits not-in-class with reasons; core/project clean.
- Lane verification: commits present, tree clean, `go test -run TestRunInit_WizardAutonomyTierAppliesBundle ./internal/cli/` → ok; TestRunInit prefix family ok (44.1s); toolpolicy+project ok; builds exit 0.
- Stage CLOSED: 3 rounds, re-review ceiling 2/2 consumed, final round dispositions written (A fixed, B carried to leader). Independent read continues at the sync-phase audit.

## Named debts transferred at close (operator decision d-20261014T085956Z-bee2) — 2026-10-14

The inter-process lock surface (F2' repair, 9b9fc68069) closes in this card in its CURRENT state: in-process serialization (recordProtocolMu) + first-generation inter-process lockfile. The following edge-case findings are NAMED DEBTS transferred to follow-up card **t1730** (after=t1630), sourced from the re-entry #2 audit round whose verdict file is NOT USED as this card's binding verdict (auditor-model served glm-5.3-flash despite an opus spawn — model-mapping repair routed to t1728; the operator elected transfer over a further repair round):

- **F9** (tier_render.go:236-244): the age branch bypasses the pid probe — a >30s lockfile is removed even with a live holder mid-render.
- **F10** (tier_render.go:231-244 + :203-210): Stat→Remove takeover TOCTOU and ownership-free release can delete a subsequent owner's fresh lock.
- **F11** (codegen.go): extractPermissions reads the settings body BEFORE the mutex+file lock — a concurrent `plan` write can be reverted to stale `bypassPermissions`.
- **F12**: USER-scope writes lack cross-process serialization (the USER render precedes the lock; the :471 preservation path is unlocked).

Primary scenario closed in-card and verified (two-live-render repro, -race). The in-process protocol (F2/F3) holds fixed. Landing-path disposition of the unused verdict file is leader-owned (pending); merge-ready HOLD unchanged (t1619 + t1591).
