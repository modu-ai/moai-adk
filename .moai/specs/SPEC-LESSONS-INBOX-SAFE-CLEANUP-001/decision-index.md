# decision-index.md — SPEC-LESSONS-INBOX-SAFE-CLEANUP-001

Authored at plan phase per the decision gate. One row per decision surfaced during assembly that the operator has not settled. Labels: `DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER`. No row carries a recommendation; a `Default:` is the output of the published Default rule (preserve current behavior; else single-revert undo; else smaller user-visible surface), not a preference. No operator verdict is invented: the only verdicts filled are `DEFAULT-APPLIED` on implementation-level rows that carry a Default, as the plan-close procedure prescribes.

## Keep-set notes (not decisions of this SPEC)

- Permanent deletion of the backed-up original, the archive, or any frozen copy is out of scope and stays an operator act (irreversible). The tooling has no verb for it (REQ-LIS-022).
- Any push, pull request, release, or settings change stays with the leader or operator.
- Planner scope decisions already taken (lane instruction: decide in-scope versus follow-up card): the PermissionDenied advice rewrite (`post_tool_failure.go:219`, EL-P) is a follow-up card, not part of this SPEC; the classifier lists it as a disposition. Recorded in `spec.md` §4; reversible by re-scoping before kickoff.

### Q1: Is a legacy offset-only checkpoint that lies within the file adopted once (flagged unverified) or refused until an owner binds it?

Label: FOUNDER
Authority anchor: (none — no committed operator setting or completed SPEC decides adoption of unverifiable legacy checkpoints; the drain-stall SPEC predates the lineage idea)
Why unresolved: Adopting keeps every existing machine draining but may bind a wrong lineage once, exactly as today; refusing is stricter but stalls draining on every machine until someone binds, which is the silent-stall shape the drain-stall SPEC was written against. Changes REQ-LIS-004 and AC-LIS-006.
Class: implementation-level
Default: adopt once, flagged `checkpoint_unverified` (rule: preserves current behavior)
Alternate: refuse with `LINEAGE_MISMATCH` until IC-2 binds
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:34:08Z manager-spec+plan-owner

### Q2: Does the consumer stay shell with the store verbs in Go, or does the consumer move to Go?

Label: FOUNDER
Authority anchor: (none — the curator skill's shell form is its current state, not an operator setting)
Why unresolved: A shell consumer never depends on the installed `moai` build (which lags the tree; EL-B) but cannot hold the OS lock or call `fsync`; a Go consumer gains both but couples session start to the installed build. Changes `design.md` §1-§3 and milestones M1/M2.
Class: implementation-level
Default: consumer stays shell; store verbs in Go (rule: preserves current behavior)
Alternate: consumer moves to Go behind a capability probe
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:34:08Z manager-spec+plan-owner

### Q3: Does adding the stable `id` field bump the row schema version `v`?

Label: FOUNDER
Authority anchor: (none — the in-code comment says `v` bumps "only when the stub schema itself changes shape", which reads either way for an additive field, and is not an operator setting)
Why unresolved: Not bumping keeps `v` at 1 and readers distinguish by the presence of `id`; bumping to 2 marks the shape change but the existing readers' behavior on `v` is unmeasured beyond `InboxStubVersion`. Changes REQ-LIS-008 and AC-LIS-007.
Class: implementation-level
Default: no bump; `v` stays 1 (rule: preserves current behavior)
Alternate: bump to 2
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:34:08Z manager-spec+plan-owner

### Q4: How is the PR1841 premerge-review case recorded — as an items/ledger entry that references the case report hashes, or as an inbox row appended through the producer mechanism?

Label: FOUNDER
Authority anchor: (none — the operator follow-up relayed by the leader asks that the case be preserved and kept unprocessed, and leaves the recording form open; the relayed message is not a committed artifact)
Why unresolved: A ledger entry leaves the inbox byte-untouched and carries the pin and dependency list directly; an inbox row makes the case visible to the same consumer path as every other observation but adds a row of a kind the hook producer does not emit today. In both options the lane (sole maintenance writer) does not hand-write inbox rows outside the mechanism, and the pin rule (REQ-LIS-016) holds. Changes `design.md` §6 and where AC-LIS-024's fixture places the item.
Class: implementation-level
Default: items/ledger entry referencing the two input hashes (rule: preserves current behavior — the inbox gains no new row kind)
Alternate: inbox row written by the mechanism's append path
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:34:08Z manager-spec+plan-owner

### Q5: Is a second copy of the frozen backup kept outside the repository tree?

Label: FOUNDER
Authority anchor: (none — no committed setting covers backup placement for local observation logs)
Why unresolved: The first copy lives under the project root's `.moai/state/lessons-cleanup/` (owner-only, gitignored; it survives the lane worktree, which a gitignored report directory would not). A second copy under the user's home guards against loss of the checkout but duplicates up to 7 MB of raw failure text (summaries up to 200 runes of tool output) into another location. Changes REQ-LIS-013 and M7 step R1.
Class: implementation-level
Default: single copy under the project root's state directory (rule: smaller user-visible surface)
Alternate: add a second owner-only copy under `~/.moai/`
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:34:08Z manager-spec+plan-owner

### Q6: Does the real cutover (M7 step R4, only if the removal set is non-empty) wait for a recorded confirmation of the removal-set preview?

Label: FOUNDER
Authority anchor: (none — the card instruction authorizes removal of proven items; it does not say whether a preview is confirmed first, and the card text is not a committed artifact)
Why unresolved: The cutover is reversible by rollback and by the retained original, and the proof gate already requires resolvable evidence; a recorded confirmation adds one pause on a file every local session appends to. If the removal set is empty, R4 does not run and this row is moot. Changes `plan.md` §C step R4.
Class: implementation-level
Default: R4 waits for a leader confirmation recorded on the preview (rule: smaller user-visible surface — nothing changes in the shared file before the preview is read)
Alternate: R4 proceeds on the gate result alone
Operator verdict: DEFAULT-APPLIED 2026-10-19T05:34:08Z manager-spec+plan-owner
