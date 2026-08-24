# t232 — guard failure-observation scenario (AC-ZRR-007)

Prepared during the #1611 review-quota wait, so that M3 executes rather than designs.

The acceptance criterion is not "the guard passes". It is: **the guard was seen failing on a
known-bad input, and seen passing on the repaired tree.** A guard observed only in its passing
state demonstrates nothing about what it protects — which is the exact defect class this whole
card exists to remove (a check that reports green while observing nothing).

This file fixes, in advance: what gets mutated, what output counts as the observation, and what
must be true before and after. Deciding those at the moment of the run invites a scenario shaped
to whatever the guard happens to do.

---

## 0. Preconditions (assert before starting; do not proceed on a miss)

| # | Check | Expected |
|---|---|---|
| P1 | `git status --porcelain` in the worktree | empty — the mutation must be the only diff |
| P2 | the repair (M1) has landed | clause hit-count 101/101 exactly-once, per mirror |
| P3 | the guard (M2) has landed and passes | exit 0, and its output reports the evaluated-entry count |
| P4 | evaluated-entry count in that passing output | `101` for each mirror, reported separately |

P4 is not decoration. Without it a guard that iterates 3 entries and returns early also "passes",
and every mutation that lands outside its exclusion set still makes it go red — so the red
observation below would prove nothing.

---

## 1. Mutation targets — fixed in advance, two of them

`CONST-V3R2-004` is named by the SPEC, and one entry is drawn at random. Both are mutated in
separate runs, not together.

Fixing the named target matters because "mutate any one entry" is not a reproducible verdict —
two runs picking different entries would both read as satisfied while testing different things.
Adding the random one matters because a guard could special-case the named entry.

| Run | Target | Field | Mutation |
|---|---|---|---|
| R1 | `CONST-V3R2-004` | `clause:` | insert one character mid-span (e.g. `language` → `languagex`) |
| R2 | random entry, ID recorded at run time | `clause:` | same, one character |
| R3 | `CONST-V3R2-004` | `anchor:` | change the slug to one that resolves to no heading |
| R4 | `CONST-V3R2-004` in the **template** mirror only | `clause:` | one character |

R3 exists because a guard that checks clauses and silently ignores anchors passes R1 and R2. R4
exists because a guard that reads only the local mirror passes R1-R3 — and the template copy is
the one users actually receive.

One character, not a deleted line: a single-character edit is the smallest thing a real rules edit
could do, so it tests the boundary rather than an obvious break.

---

## 2. What counts as the observation

For each run, ALL of the following must be captured verbatim into `progress.md` §E.2. A summary
("the guard failed as expected") is not an observation.

1. The guard's **non-zero exit code**.
2. The failure message **naming the mutated entry's ID** — not just a count. A guard that reports
   "1 entry failed" without saying which one cannot be distinguished from a guard that failed for
   an unrelated reason.
3. For R4, evidence that the **template mirror** is the surface that failed, not the local one.
4. The **restored** state: mutation reverted, guard green again, `git status --porcelain` empty.

Point 4 is what makes the red meaningful — a guard that stays red after the revert was red for
some other reason.

---

## 3. The CI-conclusion observation (separate, and not substitutable)

A local non-zero exit does not establish that a pull request would be blocked. A step wrapped in
`|| true`, or a step-level `continue-on-error`, produces a locally-red guard inside a green job.

So R1 is additionally pushed on a scratch commit, and the observation is the **CI job's own
conclusion**:

```bash
gh pr checks <PR> --json name,state --jq '.[] | select(.name=="<guard job>")'
```

Expected: that job concludes `fail`. Quote the output. Then revert the scratch commit.

Reading the workflow file and asserting it contains no `|| true` is a weaker check — it verifies
the text, not the behavior, and misses suppression introduced anywhere else in the chain.

---

## 4. Failure modes of this scenario itself

Recorded so they are not discovered as surprises:

- **The mutation makes the guard fail for the wrong reason.** If the mutated clause line breaks
  YAML parsing, the guard fails at load time and the run proves nothing about drift detection.
  Mitigation: mutate inside the quoted value, never the quoting or the key.
- **The random target lands on an entry whose clause is unusually long or short**, changing what
  the one-character edit means. Mitigation: record the chosen ID and its clause length alongside
  the output.
- **The revert is incomplete.** Mitigation: P1's `git status --porcelain` is re-asserted after
  every run, not only at the start.
- **The CI observation is read from `gh pr checks` alone.** For CodeRabbit that list prints `pass`
  byte-identically whether a review ran or was rate-limited; the same "the list is not the state"
  hazard applies to any check read from a summary rather than from its own conclusion.

---

## 5. Order

R1 → R2 → R3 → R4 locally, each with its own revert and re-green, then the single CI push for R1.
Running them together would leave the guard's per-entry reporting untested — several simultaneous
failures cannot show that the guard names the right entry.
