# Internal: lenses

When to read: only when a question or a plan needs one of the four ways of looking below; a routine change does not.

Each lens produces a report and changes no code: report first, and change code only in a separate, asked-for step.

1. **State machine.** Use it when behaviour depends on what happened before (a status field, a retry, a lock, a lifecycle). List every state, every event and the code that moves between them; mark a state with no way out and a transition no code performs. Report as a table: state, event, next state, where.
2. **CI in three stretches.** Use it when a pipeline feels slow. Measure waiting (queued before any runner), preparing (checkout, restore, install, build) and running (the checks themselves) from the run's own timestamps, separately. Report the three durations per run and the longest one first; say which runs you did not read.
3. **What would we keep if we built it again today.** Use it before extending something old. For each part, say whether you would build it again, build it smaller or drop it, and what depends on it. Report the keep list with the reason for each and the cost of removing the rest.
4. **Refactor risk.** Use it before a refactor that moves behaviour. List what the current code does that no test pins, and the callers that rely on it. Report the unpinned behaviours and the first test to add for each; add no code.
