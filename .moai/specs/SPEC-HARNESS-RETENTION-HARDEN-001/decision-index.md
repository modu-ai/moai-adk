# SPEC-HARNESS-RETENTION-HARDEN-001 — Decision Index

Carrier for decisions surfaced during plan-phase assembly that the operator has not settled in an interview (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`; `recommendation_mode: pull`, `:9`). Rows state what is unresolved and why; no row carries a recommendation or a preferred answer. The options with quantified costs live in `spec.md` §B, which also states the default the plan applies to each. `Operator verdict:` lines record what the leader session relayed on 2026-10-03 (a cross-session message, not a direct operator answer) and stay empty where nothing was relayed. This artifact carries no `status:` field.

Relayed verdicts, in one place: (1) the defaults stand with no change to the shared `internal/lockfile` contract, so D1-C, D2-B and D2-C remain operator-held and unselected; (2) event loss follows D3-B (prune-side tail-carry) plus documenting the residual window in source; (3) F5 and F6 stay "not reproduced, not measured" disclosure only; (4) Q4 is restricted: the pruner may delete only a state-path entry owned by the current user, an entry owned by another user is not deleted, the pruner warns only and the prune is skipped as today, and the restriction is pinned by a test; (5) one unexported test-only field on the pruner is allowed; (6) after a plan-audit PASS the Kickoff proceeds autonomously with a decision record.

### Q1: Does the Windows single-pruner limitation (F5) get a code repair, or only a corrected disclosure?

Label: EVIDENCE-NEEDED
Authority anchor: (none — no committed policy decides it. Context: `.moai/specs/SPEC-AGENT-TEAM-RETIRE-001/spec.md` REQ-ATR-001 requires the Windows in-process-mutex limitation preserved verbatim for the migration it governed; it does not decide whether a later retention concern may change that package.)
Why unresolved: the burst of concurrent prunes on Windows was neither reproduced nor measured (no Windows runtime exists in this environment), so the size of the loss a repair would remove is unknown. Options: a corrected disclosure only (no code, loss unchanged); a retention-local Windows guard (a new sentinel file, no shared-package change); an upgrade of `internal/lockfile/lockfile_windows.go` to a cross-process lock (changes Windows behaviour for all seven importers and goes through the amendment path of the completed SPEC-AGENT-TEAM-RETIRE-001). Costs are in `spec.md` §B D1. The lockfile option is operator-held.
Operator verdict: relayed by the leader session 2026-10-03: the disclosure-only default stands; no change to `internal/lockfile`; the lockfile option stays operator-held and unselected; F5 stays "not reproduced, not measured".

### Q2: Does the lock-wait behaviour (F6) change, or is the 5 s hook-timeout fact only recorded?

Label: EVIDENCE-NEEDED
Authority anchor: (none — `.moai/reports/t1425/decision-records.md` records the earlier choice (`lock_behavior=block_then_recheck`) and that the question omitted the 5 s hook timeout; it is a record of the earlier decision's inputs, not authority to keep or change it under the missing fact.)
Why unresolved: the waiter's actual delay and kill behaviour under real hook load was not reproduced or measured; what exists is the code shape (blocking lock only, event appended before the wait, hook `timeout: 5`, `async: true`) and one prune-duration range reported by the t1425 lane (0.53-1.79 s on 65.8 MB). Options: keep blocking waiters and record the facts (no code); add a non-blocking try-lock to `internal/lockfile` (new exported symbol, both platforms); a bounded wait built on that try-lock. The last two touch a shared package with seven importers and are operator-held. Costs in `spec.md` §B D2. The history point (the earlier question omitted the hook-timeout fact) is recorded here and in `progress.md`, not in source.
Operator verdict: relayed by the leader session 2026-10-03: the record-only default stands; no change to `internal/lockfile`; the try-lock options stay operator-held and unselected; F6 stays "not reproduced, not measured".

### Q3: How is the observed late-event loss during a prune addressed?

Label: FOUNDER
Authority anchor: (none — `.moai/reports/t1425/decision-records.md` states the loss was a known limit that card reduced but did not remove; no committed policy says how far it must go.)
Why unresolved: the loss is observed (a late event appended during the prune was absent afterwards). The options trade completeness against blast radius. Appenders take the state-file lock: closes the window on Unix; adds a lock round-trip to the append path of every observe hook; an append queued behind a prune longer than the 5 s hook timeout is not written. A retention-local tail-carry: no change to the append path; leaves a short residual window that is disclosed in source. Documenting the whole-prune window as residual risk: no code; closes nothing. Costs in `spec.md` §B D3. Appender locking is operator-held.
Operator verdict: relayed by the leader session 2026-10-03: D3-B (prune-side tail-carry) plus documenting the residual window in source; appender locking stays operator-held and unselected.

### Q4: May the pruner replace a symbolic-link or permission-denied state file, and whose entries may it delete?

Label: FOUNDER
Authority anchor: (none — no committed policy covers how a derived state file in the project's own directory is handled when it is hostile or foreign-owned.)
Why unresolved: replacing the entry closes F4 and F7 with one mechanism but removes a file the current user did not necessarily create; refusing and reporting leaves retention off while the fault persists and needs a way to surface the error that today's observer discards. The earlier draft allowed deletion of any such entry; the relayed verdict narrows it (below). Remaining design choices that follow from the verdict: where the refusal warning goes (Q7), how a foreign-owned entry is simulated in a test (Q8), and how concurrent healers are kept from removing each other's file (`spec.md` §B D4.c). Costs in `spec.md` §B D4.
Operator verdict: relayed by the leader session 2026-10-03: restricted. The pruner may delete only a state-path entry owned by the current user; an entry owned by another user is not deleted, the pruner warns only and the prune is skipped as today; the restriction must be pinned by a test.

### Q5: Does the stamp-write failure branch (mutant mC) get a production test seam, or is the surviving mutant accepted?

Label: FOUNDER
Authority anchor: (none — no committed policy decides whether a test-only seam may be added to a production struct.)
Why unresolved: a FIFO stops at the lock step on darwin (observed), `/dev/full` is Linux-only, and a user-owned file opened read-only makes the stamp write fail on darwin (observed) without any struct field; so pinning the branch can use a parameter seam (the locked phase takes the opened file), one unexported field, or no test. Costs in `spec.md` §B D5.
Operator verdict: relayed by the leader session 2026-10-03: one unexported test-only field on the pruner is allowed.

### Q6: Is Tier M the correct classification?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (classification table: M = 5-15 files, 3-file artifact set, 16 / 16 REQ and AC ceilings; S ceilings are 8 / 8).
Why unresolved: (none — the criteria are committed policy; this SPEC carries 15 REQ and 14 AC, above the Tier S ceilings and inside Tier M's, with 9 files affected (3 production, 6 test). Recorded because the line count alone would suggest Tier S.)
Operator verdict:

### Q7: Where does the warning for a refused foreign-owned state entry go?

Label: FOUNDER
Authority anchor: (none — no committed policy names a warning surface for the retention pruner; `internal/harness/safety/frozen_guard.go:124` is a precedent for a one-line standard-error warning in the same package family, not a rule.)
Why unresolved: the observer discards the pruner's error by design (`observer.go:93`, `:141`, unchanged by this SPEC), so a returned error alone reaches no operator. Options: one line on standard error from the pruner (no new field, file or exported symbol; repeats once per prune attempt while the condition persists, which is the hook-event rate; whether hook standard error is shown to a user was not measured); a typed error only (no noise; invisible in production). Costs in `spec.md` §B D4.b.
Operator verdict:

### Q8: Which seam receives the one allowed test-only field?

Label: FOUNDER
Authority anchor: (none — the allowance is the relayed Q5 verdict; no committed policy ranks the two pins.)
Why unresolved: two pins want a seam. The stamp-write failure (mutant mC) can be pinned by a parameter seam or the field. The foreign-owned refusal (Q4) cannot be constructed by a non-root user (`chown 0` fails with `Operation not permitted`, observed), so it needs the field, or a root-only test that pins it only where CI runs as root (not measured), or the real owner check pinned alone with the wiring unpinned. Costs in `spec.md` §B D5.
Operator verdict:
