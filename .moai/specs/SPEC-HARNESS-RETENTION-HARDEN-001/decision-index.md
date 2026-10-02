# SPEC-HARNESS-RETENTION-HARDEN-001 — Decision Index

Carrier for decisions surfaced during plan-phase assembly that the operator has not settled in an interview (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`; `recommendation_mode: pull`, `:9`). Rows state what is unresolved and why; no row carries a recommendation or a preferred answer. The options with quantified costs and the defaults the plan applies live in `spec.md` §B. `Operator verdict:` lines are empty at authoring. This artifact carries no `status:` field.

### Q1: Does the Windows single-pruner limitation (F5) get a code repair, or only a corrected disclosure?

Label: EVIDENCE-NEEDED
Authority anchor: (none — no committed policy decides it. Context: `.moai/specs/SPEC-AGENT-TEAM-RETIRE-001/spec.md` REQ-ATR-001 requires the Windows in-process-mutex limitation preserved verbatim for the migration it governed; it does not decide whether a later retention concern may change that package.)
Why unresolved: the burst of concurrent prunes on Windows was neither reproduced nor measured (no Windows runtime exists in this environment), so the size of the loss a repair would remove is unknown. Options that repair it are a retention-local Windows guard (no shared-package change) and an upgrade of `internal/lockfile/lockfile_windows.go` to a cross-process lock (changes Windows behaviour for all seven importers and goes through the amendment path of the completed SPEC-AGENT-TEAM-RETIRE-001). Costs are in `spec.md` §B D1. Operator-held if the lockfile option is wanted.
Operator verdict:

### Q2: Does the lock-wait behaviour (F6) change, or is the 5 s hook-timeout fact only recorded?

Label: EVIDENCE-NEEDED
Authority anchor: (none — `.moai/reports/t1425/decision-records.md` records the earlier choice (`lock_behavior=block_then_recheck`) and that the question omitted the 5 s hook timeout; it is a record of the earlier decision's inputs, not authority to keep or change it under the missing fact.)
Why unresolved: the waiter's actual delay and kill behaviour under real hook load was not reproduced or measured; what exists is the code shape (blocking lock only, event appended before the wait, hook `timeout: 5`, `async: true`) and one prune-duration range reported by the t1425 lane (0.53-1.79 s on 65.8 MB). Options that change behaviour add a non-blocking try-lock to `internal/lockfile` (new exported symbol, both platforms) or a bounded wait built on it; both touch a shared package with seven importers. Costs in `spec.md` §B D2. Operator-held if either is wanted.
Operator verdict:

### Q3: How is the observed late-event loss during a prune addressed?

Label: FOUNDER
Authority anchor: (none — `.moai/reports/t1425/decision-records.md` states the loss was a known limit that card reduced but did not remove; no committed policy says how far it must go.)
Why unresolved: the loss is observed (a late event appended during the prune was absent afterwards). The options trade completeness against blast radius: appenders taking the state-file lock (closes the window on Unix, adds a lock round-trip and a wait for the whole prune to the hot append path of every observe hook, and can get an append killed at the 5 s hook timeout before it writes), a retention-local tail-carry (no change to the append path, leaves a short residual window that must be disclosed), or documenting the whole-prune window as residual risk (closes nothing). Costs in `spec.md` §B D3. Operator-held if appender locking is wanted.
Operator verdict:

### Q4: May the pruner replace a symbolic-link or permission-denied state file, or must it refuse and report?

Label: FOUNDER
Authority anchor: (none — no committed policy covers how a derived state file in the project's own directory is handled when it is hostile or foreign-owned.)
Why unresolved: replacing it closes F4 and F7 with one mechanism but deletes a file the current user did not necessarily create (a 35-byte derived stamp; for a symbolic link only the link is removed); refusing and reporting leaves retention off while the fault persists and needs a way to surface the error that today's observer discards. Costs in `spec.md` §B D4.
Operator verdict:

### Q5: Does the stamp-write failure branch (mutant mC) get a production test seam, or is the surviving mutant accepted?

Label: FOUNDER
Authority anchor: (none — no committed policy decides whether a test-only seam may be added to a production struct.)
Why unresolved: no file-system fault available on both CI platforms makes the write fail after open and lock succeed (a FIFO stops at the lock step on darwin, observed; the Linux-only `/dev/full` device is not portable), so pinning the branch needs a one-field unexported seam on the pruner; the alternative leaves N2 half-closed. Costs in `spec.md` §B D5.
Operator verdict:

### Q6: Is Tier M the correct classification?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (classification table: M = 5-15 files, 3-file artifact set, 16 / 16 REQ and AC ceilings; S ceilings are 8 / 8).
Why unresolved: (none — the criteria are committed policy; this SPEC carries 12 REQ and 15 AC, above the Tier S ceilings and inside Tier M's, with one production file. Recorded because the file-count guidance alone would suggest Tier S.)
Operator verdict:
