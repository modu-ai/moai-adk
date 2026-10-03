# t1432 — residual risk: the check-then-remove window in the state-entry heal

Decision basis: sync-audit finding F1 (`.moai/reports/t1432/sync-audit.md`), the cross-model convergence FAIL (required backend codex, P1), and the leader's ruling of 2026-10-03 to repair if possible within the SPEC wording, otherwise to record the residual risk (relayed by the leader session, not an operator answer).

## What the window is

`removeStateEntryIfUnchanged` (`internal/harness/retention.go`) re-inspects the state-path entry with `os.Lstat`, compares identity, type, mode and modification time, then calls `os.Remove`. These are two system calls. A concurrent healer that puts a fresh state file at the path between them has that file removed, and the two processes then lock different inodes: two pruners in one interval, the condition REQ-HRH-005 exists to prevent.

## Measurement (this audit's, not the lane's)

- Probe: a single-swap test against the production helper, scratch test through `go test -overlay`, outside the tree. Result: `trials=5235 helper_removed_the_swapped_in_fresh_entry=5` (about 0.1 percent per trial). The swapper was made to race the window on purpose, so this is not a production rate.
- A second, independent probe by the delta audit measured 235 removals in 20000 trials (about 1.2 percent), roughly twelve times the first figure. The rate depends on how the probe races the window, so neither number is a production rate; both only show the window is real and reachable. This supersedes the SPEC's "not measured" for the existence of the window, not for its production frequency.
- Codex reported a 3 of 3 reproduction through its own overlay, which was not available to the audit.
- Platform: darwin, APFS, one machine. Linux not observed.
- Production trigger: a faulty entry owned by the current user plus a second hook process healing at the same instant. Not measured.

## Consequence, bounded

The state file itself is a 35-byte derived timestamp, so losing it destroys no user content directly. The consequence that matters is what its loss permits: a second pruner in the same interval, which is the condition card t1425 was written to stop. `spec.md` D4.c states that hazard as re-opening concurrent appends to one archive file, and the t1425 evidence (`.moai/reports/t1425/archive-damage.md`) measured the damage of concurrent pruners on the live archives: three monthly archives fail `gzip -t`, and the salvage found most recovered lines to be duplicates. So the window is not harmless: its effect is possible archive damage and duplicated archived events, once, in an interval where a faulty owned state entry is healed by two hook processes at the same instant. Whether a single extra pruner in that one interval produces the same damage as the sustained storm of t1425 was not measured here.

An earlier version of this record, and the lane's message to the leader, described the effect as only one extra prune and "not a loss of user data". That understated the consequence and is corrected here (delta audit D1). The leader's acceptance of 2026-10-03 cited that wording, so the acceptance should be read against this paragraph.

Who can trigger it: the heal only starts on an entry the current user owns, so another user cannot cause it through this path. The claim that directory permissions or the sticky bit refuse a foreign swapped-in entry is reasoned, not observed, and is not universal: a user who owns the directory can unlink any entry in it (delta audit D2).

## Why it is recorded rather than repaired

The repair the audit proposed is rename the entry to a unique temporary name, inspect the renamed entry, unlink it if it matches the inspected one, and otherwise put it back with `os.Link`. Reading it against the code and the SPEC:

1. It does not close the window. While the entry sits under the temporary name the state path is empty, so a third process can create a file there; the put-back then fails with "already exists", the winner's file stays orphaned under the temporary name, and two inodes are in use. The failing interleaving needs three parties instead of two, which lowers its probability and does not remove it.
2. It would move an entry away before verifying it, while REQ-HRH-005 says the pruner shall remove an entry only if it is still the entry inspected. In the unlucky case the winner's file is moved off the path, which is what the requirement forbids.
3. The put-back relies on `os.Link` semantics that differ across platforms (a link to a symbolic link, Windows). None of that can be observed here beyond darwin.

Closing the window needs a serialization point, which is SPEC D4.c option C (a separate heal lock file). That option was rejected in the plan as adding an artifact and a failure mode, and adopting it changes the SPEC body, which invalidates the audited plan-artifact hash. The leader ruled that an option needing a SPEC revision falls back to recording the residual risk. This is an analysis from reading the code, not a measurement of the alternative.

## What stays disclosed

`spec.md` D4.c and §F ("Heal burst"), `CHANGELOG.md` ("narrows but does not close"), and this record. If a production occurrence is ever observed, the follow-up is a new card that amends D4.c towards option C.
