# SPEC-FACTORY-RETIRE-GUARD-001 — Decision Index

Decision gate: `interview.decision_gate: on` (`.moai/config/sections/interview.yaml` line 6). Each row below was surfaced while assembling the SPEC. No row carries a recommendation for a judgment call; a `Default:` line is the published Default rule applied (preserves current behavior; else single-revert undo; else smaller user-visible surface). The artifact is stateless on the status axis; the SPEC lifecycle is in `spec.md`.

Authority-register search performed at tree `d2d9ac08f`: `grep -rln RetireRunIfDead .moai/specs --include='spec.md'` returned, besides this SPEC, `SPEC-FACTORY-RUN-RETIRE-001` and `SPEC-FACTORY-LANE-JOIN-SOCKET-001` (both `completed`). `SPEC-FACTORY-RUN-RETIRE-001` REQ-008 says the operator command retires a run "only when the operator names that run id explicitly" and REQ-005 makes `retirable` the single gate; neither decides any row below, and this SPEC does not contradict them (it adds a refusal on top of the gate). `.moai/config/sections/*.yaml` carries no setting covering these questions. The audit decisions inherited from the twice-failed card line (SPEC-FACTORY-RUN-SELECT-001 / -CORE-001) live in sibling card worktrees that are not committed authority, so they are cited in `spec.md` §F as history, and the rows below route `FOUNDER` rather than `DECIDED`; the scope the card itself fixes (refuse-only retirement, no force path for a live-leader run, part C only, Korean operator prose) is held in the backlog record and the dispatch, not in the committed register, so it is stated in `spec.md` §E as dispatch-fixed scope and is not a row here.

### Q1: Should the retirement compare-and-set reach only the operator path (`retireRun`), or also the automatic reconciler's write (`retireRunUnlessLive`)?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (no committed artifact decides this)
Why unresolved: the card forbids any new force-cleanup path for a live-leader run. The claim that holds by construction is narrower than "every retirement entry point": `retireRun` is the only retirement writer this SPEC changes. The automatic writer has the same gap (no owner-stamp condition, `RowsAffected` unread, unconditional event) and is reached by every multi-run lane join today; extending the check to it needs a three-way outcome (retired / preserved-for-live-records / owner-changed), a signature change, a `ReconcileActiveRuns` routing edit and changes to the tests that pin the `Preserved` list — a larger, separately auditable change.
Default: operator path only; the reconciler is unchanged and its residual risk is recorded in `spec.md` §G (rule: preserves current behavior)
Alternate: extend the compare-and-set to `retireRunUnlessLive` with a three-way outcome in this SPEC
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:45:32Z manager-spec+plan-author

### Q2: Which card states count as "in-progress" for the retirement guard?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (the shared terminal set `{done, failed, abandoned}` exists in `internal/homestate/card_record.go:74` and is used by the t1707 live-records guard; no completed SPEC records it as a decision)
Why unresolved: an alternative reading counts only lease-holding states (`leased` through `merging`), which would ignore `picked`, `needs-decision`, `blocked`, `merged-local` and `pushed` cards that a retirement would still strand.
Default: every non-terminal state, the same set the automatic live-records guard uses (rule: preserves current behavior — one definition shared with the automatic guard)
Alternate: lease-holding states only
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:45:32Z manager-spec+plan-author

### Q3: What should `RetireRunIfDead` and `moai factory runs --retire` do for a run that is already retired when it is classified dead?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: today the call succeeds, prints `retired <run>` and appends a second `run.retired` event (probe P2, `acceptance.md`). The incident run itself (`tmvgvn`) is a retired run that still holds 41 non-terminal cards, so a bare idempotent success would be a silent report of stranded cards, and an error would change the visible result of an idempotent retry. Classification note: the exit code and the first output line are unchanged; the command adds an informational block and the library stops writing a duplicate event, so the row is classed implementation-level. If the operator reads the added block as a change of a shipped command's default user-visible behavior, the row becomes product-level and the verdict below must be replaced by an operator answer.
Default: idempotent no-op — no write, no card read, no second event, nil error, exit zero — plus an informational block listing the stranded non-terminal cards (id, state, owner label) with `moai factory decide <card> --choice abandon --run <run>` as an available exit, with its caveats (rule: preserves current behavior — the success exit and message)
Alternate: return an error wrapping `ErrRunOwnerNotDead` stating the run is no longer active
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:45:32Z manager-spec+plan-author

### Q4: How is the "no new force flag" criterion, which is green on the unmodified tree, classified?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the flag set of `moai factory runs` is already `{retire}` and the retirement-gate guards already pass, so a release-blocking label with a RED-now cell would be the vacuous direction of `verification-completeness.md` §2.
Default: label it a regression guard, green by design and not release-blocking (rule: the smaller change — no invented red clause; the refusal itself is pinned by the red AC-RRG-005)
Alternate: add a clause that is red today (for example that a refused retirement of a run holding cards offers no override flag, which AC-RRG-005 already partly covers)
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:45:32Z manager-spec+plan-author

### Q5: Should the five existing hint strings that tell an operator to run `moai factory runs --retire` be reworded to mention the in-progress-card refusal?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: `internal/cli/factory.go:500`, `:545`, `:687`, `internal/cli/factory_run_adopt.go:103` and `internal/factory/factory_slots.go:142` name the command as an exit. For a run that holds cards the command now refuses and prints the real exits, so the hints are not wrong, only one step shorter than the truth. Rewording them touches five strings and the tests that may pin them (not measured).
Default: leave the hint strings unchanged (rule: preserves current behavior)
Alternate: append "(refused while the run holds in-progress cards)" to each hint in this SPEC
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:45:32Z manager-spec+plan-author
