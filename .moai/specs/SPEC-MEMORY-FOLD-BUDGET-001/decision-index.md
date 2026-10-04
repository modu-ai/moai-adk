# decision-index.md — SPEC-MEMORY-FOLD-BUDGET-001

Questions the interview did not settle, one row each. Detect → Explain → Ask: each row says what is unresolved and why. Labels use the fixed vocabulary `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. Revision 0.2.0 (plan delta for audit iteration 1): every row carries `Class:`; rows classed implementation-level carry the `Default:` picked, the `Alternate:` and a dated `DEFAULT-APPLIED` verdict; the two product-level rows carry the leader's ruling as the recorded verdict; the two evidence rows state what was and was not gathered. A ninth row (Q9) is added for plan.md OD-6, which had none.

Source note for the dated rulings below: the leader's instruction file `.moai/reports/t1502/leader-disposition.md` is gitignored (`.moai/reports/*`, `.gitignore:235`), so it is **not a committed authority anchor** in the sense of the authority register; no row cites it as an `Authority anchor`. It is named only as the source of a recorded instruction, and those rows keep the label `FOUNDER`.

### Q1: Is fold-on-done on by default at first release, or opt-in? (plan.md OD-1)

Label: FOUNDER
Class: product-level
Authority anchor: none — no committed setting or completed SPEC decides this question as written.
Why unresolved: the card says to wire the fold into card close and also says the feature is applied to the real store only after the leader confirms; the two do not fix whether the compiled default is enabled. It is a choice about who bears the risk of an unattended write to a store the host also writes.
Operator verdict: 2026-10-04 — per leader disposition `.moai/reports/t1502/leader-disposition.md` (item 2): Q1 = default OFF + environment-variable gate. The operator is informed by report.

### Q2: Is the gate an environment variable with a constant, or a key in a configuration section? (OD-2)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: both mechanisms exist in this subsystem (an environment kill switch for the audit; typed configuration sections elsewhere), and the cost and discoverability differ; no setting names the intended one.
Default: environment variable read at the call site, with a compiled default constant (rule: the `MOAI_MEMORY_AUDIT` kill-switch precedent in `internal/hook/post_tool.go` and `session_start.go`; consistent with the leader's environment-variable gate for Q1).
Alternate: a key in a configuration section (needs a typed struct, a loader, a template mirror and the loader-completeness test).
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q3: Which measure and which byte cap should the 80 % warning be tied to? (OD-3, OD-4)

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: none — the doctrine records the host announcement (200 lines or 25KB) and a completed SPEC records a non-truncation observation, neither fixes the unit.
Why unresolved: whether the loader cuts by raw bytes, characters, loaded content or lines, and whether "25KB" is 25,000 or 25,600 bytes, needs a deliberate truncation experiment that does not exist yet.
Evidence on hand (gathered in this delta, without a truncation experiment and without the real store): (1) `SPEC-MEMORY-STORE-RECONCILE-001` `spec.md:60-61` — a measured index of 26,280 bytes, 18,463 characters and 163 lines whose final line was present in the measuring session's injected context, so a raw-byte cut at 25,600 did not occur; three explanations stay undistinguished (a character cap, a line-only cap, a larger byte cap). (2) The changelog entries quoted in `.claude/rules/moai/workflow/moai-memory.md` § MEMORY.md Index Budget — 2.1.83 (200 lines or 25KB), 2.1.210 (over-limit write is an explicit error), 2.1.211 (the over-limit warning measures loaded content, excluding frontmatter and HTML comments), 2.1.268 (the truncation warning states the lines cut); read here from the doctrine file, not re-fetched from upstream. (3) The doctor measures on the fixture (1,155 bytes, 1,095 characters, 17 lines) — a measurement of the doctor's own reconstruction, not of the loader. The loader's cut unit cannot be established from this evidence, and no truncation experiment was run.
Effect on this SPEC while unresolved: none of the requirements depends on the answer — the thresholds are configuration constants, the check is advisory, and every finding states that the byte basis is a conservative proxy.
Operator verdict:

### Q4: Which archive index does fold file into when the month rolls over or none exists? (OD-5)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: creating a new month index automatically conflicts with the orphan audit's three-link threshold (measured at the pinned tree), while never creating one leaves month rollover as a manual step; which cost the operator prefers is not recorded anywhere.
Default: file into the greatest-named existing archive index that `MEMORY.md` links; never create one; refuse, naming the file, when none qualifies or when the index would carry fewer resolved links than the doctor's threshold after the fold (rule: `.claude/rules/moai/workflow/moai-memory.md` § Compressing the index means making entries shorter — never fewer and § Admission — acceptance is reachability; measured premise P7).
Alternate: create the month's archive index when missing (adds a pointer line to `MEMORY.md` and still needs the threshold's links).
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q5: Does a card line need both a title-leading card id and a card-identifying link target? (OD-7)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: the card names both signals ("link target AND text naming the card id") without saying whether either alone is enough; a looser rule would move general-discipline lines that merely cite a card, a stricter one leaves unusual lines behind as reported-but-kept.
Default: conjunctive — both signals; one signal is AMBIGUOUS, reported and never moved (rule: `moai-memory.md` § Admission keeps general discipline in the always-loaded index; and the fail-safe direction of `.claude/rules/moai/core/askuser-protocol.md` § The three adopted conditions, "when uncertain, escalate, never downgrade").
Alternate: disjunctive — either signal moves the line.
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q6: Is link repair part of this card, and what similarity counts as unambiguous? (OD-8)

Label: EVIDENCE-NEEDED
Class: implementation-level
Authority anchor: none.
Why unresolved: the scope half is settled by the leader's ruling of 2026-10-04 (`.moai/reports/t1502/leader-disposition.md`, item 1: option A, link repair splits to a follow-up card; this card keeps classification and report only). The threshold half moves with it: the similarity measure and its threshold belong to the follow-up card, and no requirement of this SPEC uses them. The only evidence on hand is one measured name pair (0.778 by file name, computed from two strings; the real store was not read) — a single data point, not evidence of how a real dangling set splits — and the plan-audit found a counter-example shape in the same naming scheme (two names differing only in the card id score above the threshold; audit D11).
Operator verdict:

### Q7: Should SessionStart reach the operator as well as the orchestrator, and how is the store derived there? (OD-9)

Label: FOUNDER
Class: product-level
Authority anchor: none.
Why unresolved: the existing advisories use one channel, and the repository's own doctrine and code disagree on which store key the host loads inside a linked worktree; adding a second channel or a shared resolver are both open.
Operator verdict: 2026-10-04 — per leader disposition `.moai/reports/t1502/leader-disposition.md` (item 2): Q7 = additionalContext only (no per-session screen output). The operator is informed by report. (The store derivation stays as plan.md OD-9 records: same two keys in the same order as the doctor, no shared resolver in this card.)

### Q8: What should a repo-relative link become? (OD-10)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: the card says "rewrite or move, never delete" without choosing; an absolute path can itself be dead on another machine, and moving the line is the fold's job rather than a link repair's. This question belongs to the split-out half: the repo-relative-link rewrite moves to the follow-up card (leader ruling item 1).
Default: within this card a repo-relative link is left exactly as written and only reported as `MEMORY_REPO_RELATIVE_LINK` (rule: `moai-memory.md` § Compressing the index — never fewer, which forbids removing or rewriting an entry's reach; and the leader's reduction of card item 3 to "report").
Alternate: rewrite to the absolute path when the file exists under the project root (the follow-up card's candidate), or move the line (the fold's job).
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q9: Does the fold preview by default or apply by default? (OD-6)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: the card asks for consistency with the existing memory verbs, but they differ: measured at the pinned tree, `moai memory drain` previews by default and takes `--yes`, while `moai memory archive` applies immediately.
Default: preview by default, `--yes` applies (rule: the `moai memory drain` precedent, `.claude/rules/moai/workflow/moai-memory.md` § Worktree Mirror and Drain — "Preview by default (writes nothing)").
Alternate: apply by default, as `moai memory archive` does.
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec
