# decision-index.md — SPEC-MEMORY-FOLD-BUDGET-001

Questions the interview did not settle, one row each. Detect → Explain → Ask: each row says what is unresolved and why, and carries no preferred answer. The recommended defaults that the leader asked for live in `plan.md` §A.1, not here. Labels use the fixed vocabulary `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`.

### Q1: Is fold-on-done on by default at first release, or opt-in? (plan.md OD-1)

Label: FOUNDER
Authority anchor: none — no committed setting or completed SPEC decides this question as written.
Why unresolved: the card says to wire the fold into card close and also says the feature is applied to the real store only after the leader confirms; the two do not fix whether the compiled default is enabled. It is a choice about who bears the risk of an unattended write to a store the host also writes.
Operator verdict:

### Q2: Is the gate an environment variable with a constant, or a key in a configuration section? (OD-2)

Label: FOUNDER
Authority anchor: none.
Why unresolved: both mechanisms exist in this subsystem (an environment kill switch for the audit; typed configuration sections elsewhere), and the cost and discoverability differ; no setting names the intended one.
Operator verdict:

### Q3: Which measure and which byte cap should the 80 % warning be tied to? (OD-3, OD-4)

Label: EVIDENCE-NEEDED
Authority anchor: none — the doctrine records the host announcement (200 lines or 25KB) and a completed SPEC records a non-truncation observation, neither fixes the unit.
Why unresolved: whether the loader cuts by raw bytes, characters, loaded content or lines, and whether "25KB" is 25,000 or 25,600 bytes, needs a deliberate truncation experiment that does not exist yet; this SPEC reports all four measures and treats bytes as an advisory proxy.
Operator verdict:

### Q4: Which archive index does fold file into when the month rolls over or none exists? (OD-5)

Label: FOUNDER
Authority anchor: none.
Why unresolved: creating a new month index automatically conflicts with the orphan audit's three-link threshold (measured at the pinned tree), while never creating one leaves month rollover as a manual step; which cost the operator prefers is not recorded anywhere.
Operator verdict:

### Q5: Does a card line need both a title-leading card id and a card-identifying link target? (OD-7)

Label: FOUNDER
Authority anchor: none.
Why unresolved: the card names both signals ("link target AND text naming the card id") without saying whether either alone is enough; a looser rule would move general-discipline lines that merely cite a card, a stricter one leaves unusual lines behind as reported-but-kept.
Operator verdict:

### Q6: Is link repair part of this card, and what similarity counts as unambiguous? (OD-8)

Label: EVIDENCE-NEEDED
Authority anchor: none.
Why unresolved: one name pair is measured (0.778 by file name); a threshold from a single data point is not evidence of how the real dangling set will split, and the real store may not be read by this card. Whether to split link repair into its own SPEC is a scope choice with no recorded precedent.
Operator verdict:

### Q7: Should SessionStart reach the operator as well as the orchestrator, and how is the store derived there? (OD-9)

Label: FOUNDER
Authority anchor: none.
Why unresolved: the existing advisories use one channel, and the repository's own doctrine and code disagree on which store key the host loads inside a linked worktree; adding a second channel or a shared resolver are both open.
Operator verdict:

### Q8: What should a repo-relative link become? (OD-10)

Label: FOUNDER
Authority anchor: none.
Why unresolved: the card says "rewrite or move, never delete" without choosing; an absolute path can itself be dead on another machine, and moving the line is the fold's job rather than a link repair's.
Operator verdict:
