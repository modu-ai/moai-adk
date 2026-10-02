# Decision Index — SPEC-MOAI-BOARD-MOD-001

Decisions surfaced while assembling this SPEC that the operator has not settled. Each row states what is unresolved and why; none carries a preferred answer (decision gate on, recommendation mode `pull`: `.moai/config/sections/interview.yaml`). Labels: `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. No row qualifies as `DECIDED` or `POLICY-COVERED`: none has an authority anchor verifiable in the committed tree, so none is relabeled to look settled. Measurements cited are `M-n` / `G-n` of `spec.md`, taken at tree `802a72235` (M-10, M-11, M-13, M-14 and M-15 re-measured at `5f6c7d343`; M-16 and M-17 at `3de688996`).

### Q1: Where does the mod load from once the prototype stage ends — the user's mods folder, a project scope, or a repository-distributed path?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: card t1434 (plugin load-scope measurement) is not done, so which scopes load a mod, and with what trust prompts, is unmeasured. `spec.md` D-1 fixes only the prototype location (`mods/moai-board/`, loaded with `claude --plugin-dir`), and says it moves with this answer.
Operator verdict:

### Q2: Should the text list (`moai gtd list --limit 0`) be the primary queue source, with the JSON read kept only as a detail path, instead of JSON-primary with a text fallback?

Label: FOUNDER
Authority anchor: —
Why unresolved: the JSON read is 1,938,159 bytes, of which `items` is 194,613 (M-2), and sits at 46% of the 4,194,304-byte read cap with an unmeasured growth rate (G-4); the text read is 39,534 bytes and 1.14 s (M-3), and its rows matched the JSON's non-dropped items in set and order at measurement time. The text rows carry no `spec_id` or `added_at`, and the text form is not documented as a parseable contract. The leader's dispatch named the JSON read as the data source; this SPEC follows it and adds the fallback of REQ-MBM-008.
Operator verdict:

### Q3: What polling interval, at or above the 15,000 ms floor of REQ-MBM-006, does the board use?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: the cost of parsing about 1.9 MB inside the mod's isolate against the 10 s hook budget is unmeasured (G-3), and `moai`'s own wall time varied from 1.46 s to 2.68 s across runs (M-2). The floor is derived from the slowest measured call; the value above it depends on how fresh the operator needs the view and on the in-engine parse cost.
Operator verdict:

### Q4: Does the SPEC tab read the session root only, or also the primary checkout's `.moai/specs`?

Label: FOUNDER
Authority anchor: —
Why unresolved: `moai spec status --list` follows the process's working directory: this worktree printed 1,010 SPEC-id rows at `5f6c7d343` (M-14) and the primary checkout 678 lines (M-6), so the tab shows whichever tree the session sits in. A SPEC that exists only on a card's branch is absent from the primary until merged. Reading both needs a way to locate the primary from a worktree and a rule for a SPEC present in both; none is specified here.
Operator verdict:

### Q5: Is the pick button offered in a lane session, where the CLI refuses queue mutation?

Label: FOUNDER
Authority anchor: —
Why unresolved: from a lane, `moai gtd next` exits 1 with `refused — lane boundary: a lane session cannot mutate the queue` (M-10), and even bare `moai gtd next` is refused. Offering the button there yields a guaranteed refusal shown in the pane; hiding it needs the mod to read the session's environment (`MOAI_FACTORY_ROLE=lane` here), which REQ-MBM-013 currently forbids. A successful pick was not observable from a lane (G-2).
Operator verdict:

### Q6: Which statuses count as "active" for the SPEC tab's default filter?

Label: FOUNDER
Authority anchor: —
Why unresolved: REQ-MBM-011 currently names draft and in-progress (25 + 15 = 40 rows, M-6). `implemented` has 143 rows and covers SPECs whose code landed but which are not closed; whether those belong in the default view is a product reading of "active" that no committed artifact states.
Operator verdict:

### Q7: Must the mod also load under the operator's own Claude Code profile, whose rollout switch refuses hooks modules, before the card counts as delivered?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: under the operator profile `claude plugin test` exited 1 with `hooks modules are turned off in this process: the rollout switch served off` in every run after the first four of this session, `--help` included (M-13), and the same switch gates loading a hooks module in a session (G-11). Under an empty temp config dir the same runner executes tests and separates pass from fail (M-16), so the engine criteria no longer wait on it and acceptance.md §A.2 / §F grant no waiver. What is open: why the operator profile is off, whether it flips back, and whether the operator needs `/moai-board` to load there (AC-MBM-014 is the check) or accepts the mod being exercised only under the temp profile until the switch changes. The leader reports the profile refusal to the operator; it is not a SPEC blocker.
Operator verdict:
