# decision-index.md — SPEC-MEMORY-FOLD-BUDGET-001

Questions the interview did not settle, one row each. Detect → Explain → Ask: each row says what is unresolved and why. Labels use the fixed vocabulary `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. Revision 0.3.0 (plan delta for audit iteration 2): the two product-level rows (Q1, Q7) now carry the operator's own confirmation; the two evidence rows (Q3, Q6) state that no requirement depends on them and why that leaves them non-blocking for the Kickoff; every `Default:` names the step of the published default rule that selects it; Q7 is kept because the SessionStart follow-up card inherits it; Q4's default text follows the single archive-index definition of `spec.md` §1.5. Revision 0.2.0 had added `Class:` to every row, dated `DEFAULT-APPLIED` verdicts to the implementation-level rows, and the ninth row (Q9).

Source note for the two operator verdicts below (Q1, Q7). Both were confirmed by the operator directly on 2026-10-04 and relayed through the leader's question channel. That relay is the only source recorded: no row cites a file as an `Authority anchor`, and no verdict rests on a gitignored report. The rows keep the label `FOUNDER` because no committed setting or completed SPEC decides them.

### Q1: Is fold-on-done on by default at first release, or opt-in? (plan.md OD-1)

Label: FOUNDER
Class: product-level
Authority anchor: none — no committed setting or completed SPEC decides this question as written.
Why unresolved: the card says to wire the fold into card close and also says the feature is applied to the real store only after the leader confirms; the two do not fix whether the compiled default is enabled. It is a choice about who bears the risk of an unattended write to a store the host also writes.
Operator verdict: operator confirmed (via the leader's question channel), 2026-10-04: the `gtd done` wiring ships default OFF and is enabled by an environment variable.

### Q2: Is the gate an environment variable with a constant, or a key in a configuration section? (OD-2)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: both mechanisms exist in this subsystem (an environment kill switch for the audit; typed configuration sections elsewhere), and the cost and discoverability differ; no setting names the intended one.
Default: environment variable read at the call site, with a compiled default constant (rule: step 3 of the published default rule — the option with the smaller user-visible surface, since a configuration key adds a typed struct, a loader and a template mirror; consistent with the `MOAI_MEMORY_AUDIT` kill-switch precedent in `internal/hook/post_tool.go` and `session_start.go` and with the operator's environment-variable gate for Q1).
Alternate: a key in a configuration section (needs a typed struct, a loader, a template mirror and the loader-completeness test).
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q3: Which measure and which byte cap should the 80 % warning be tied to? (OD-3, OD-4)

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: none — the doctrine records the host announcement (200 lines or 25KB) and a completed SPEC records a non-truncation observation, neither fixes the unit.
Why unresolved: whether the loader cuts by raw bytes, characters, loaded content or lines, and whether "25KB" is 25,000 or 25,600 bytes, needs a deliberate truncation experiment that does not exist yet.
Evidence on hand (gathered in revision 0.2.0, without a truncation experiment and without the real store): (1) `SPEC-MEMORY-STORE-RECONCILE-001` `spec.md:60-61` — a measured index of 26,280 bytes, 18,463 characters and 163 lines whose final line was present in the measuring session's injected context, so a raw-byte cut at 25,600 did not occur; three explanations stay undistinguished (a character cap, a line-only cap, a larger byte cap). (2) The changelog entries quoted in `.claude/rules/moai/workflow/moai-memory.md` § MEMORY.md Index Budget — 2.1.83 (200 lines or 25KB), 2.1.210 (over-limit write is an explicit error), 2.1.211 (the over-limit warning measures loaded content, excluding frontmatter and HTML comments), 2.1.268 (the truncation warning states the lines cut); read here from the doctrine file, not re-fetched from upstream. (3) The doctor measures on the fixture (1,155 bytes, 1,095 characters, 17 lines) — a measurement of the doctor's own reconstruction, not of the loader. The loader's cut unit cannot be established from this evidence, and no truncation experiment was run.
No requirement of this SPEC depends on it. Judged against the current text (revision 0.3.0): REQ-MFB-008 reports four figures; REQ-MFB-009 compares the doctor's own measured values with configured values and states the basis as an unconfirmed proxy; AC-MFB-009 checks `index_loaded_chars` against this SPEC's own stated reconstruction of the documented exclusion, never against the loader; AC-MFB-010 and AC-MFB-011 assert threshold arithmetic and the proxy wording. No requirement or criterion asserts what the loader does.
Why that makes it non-blocking for the Kickoff: a row blocks only when some requirement cannot be implemented or verified without its answer, and none here is in that position — the byte cap, line cap and warn percentage are configuration values with per-invocation flags, every finding is advisory (it warns, never enforces, and the doctor exit code stays 0), and the safe error direction is a false alarm (bytes is never smaller than characters). The answer, when a truncation experiment supplies it, changes constants in `internal/config/defaults.go` (plan OD-4) and possibly the axis REQ-MFB-009 keys on (OD-3) — a revision of REQ-MFB-009 with AC-MFB-010 and AC-MFB-011, not a precondition of the run.
Operator verdict:

### Q4: Which archive index does fold file into when the month rolls over or none exists? (OD-5)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: creating a new month index automatically conflicts with the orphan audit's three-link threshold (measured at the pinned tree), while never creating one leaves month rollover as a manual step; which cost the operator prefers is not recorded anywhere.
Default: file into the archive index defined once in `spec.md` §1.5 — among the `project_card_archive_<YYYY>_<MM>.md` files that `MEMORY.md` links and that exist, the greatest name; never create one; refuse, naming the file, when none qualifies (naming any unlinked archive-pattern file as the one to link) or when it would carry fewer resolved links than the doctor's threshold after the fold (rule: step 3 of the published default rule — the option with the smaller user-visible surface, since creating an index adds a file and a pointer line to `MEMORY.md`; grounded in measured premise P7 and consistent with `.claude/rules/moai/workflow/moai-memory.md` § Compressing the index means making entries shorter — never fewer and § Admission — acceptance is reachability).
Alternate: create the month's archive index when missing (adds a pointer line to `MEMORY.md` and still needs the threshold's links).
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q5: Does a card line need both a title-leading card id and a card-identifying link target? (OD-7)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: the card names both signals ("link target AND text naming the card id") without saying whether either alone is enough; a looser rule would move general-discipline lines that merely cite a card, a stricter one leaves unusual lines behind as reported-but-kept.
Default: conjunctive — both signals; one signal is AMBIGUOUS, reported and never moved (rule: step 3 of the published default rule — the option with the smaller user-visible surface, since the conjunctive rule moves fewer lines; `moai-memory.md` § Admission keeps general discipline in the always-loaded index, and the fail-safe direction of `.claude/rules/moai/core/askuser-protocol.md` § The three adopted conditions applies — "when uncertain, escalate, never downgrade").
Alternate: disjunctive — either signal moves the line.
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q6: Is link repair part of this card, and what similarity counts as unambiguous? (OD-8)

Label: EVIDENCE-NEEDED
Class: implementation-level
Authority anchor: none.
Why unresolved: the scope half was settled by the leader's ruling of 2026-10-04 (option A: link repair splits to a follow-up card; this card keeps classification and report only). The threshold half moved with it: the similarity measure and its threshold belong to the follow-up card. The only evidence on hand is one measured name pair (0.778 by file name, computed from two strings; the real store was not read) — a single data point, not evidence of how a real dangling set splits — and the plan-audit found a counter-example shape in the same naming scheme (two names differing only in the card id score above the threshold; audit D11).
No requirement of this SPEC depends on it. Its whole subject — the similarity measure, its threshold and the repair of a link — is Out of Scope here (`spec.md` §4, link repair): REQ-MFB-011 classifies and reports, and states that the doctor suggests no replacement for any link; no requirement, criterion or constant of this SPEC carries a similarity measure.
Why that makes it non-blocking for the Kickoff: the question left this SPEC with the split, so no requirement can be unimplementable or unverifiable for want of its answer. The follow-up card inherits the row, the one data point and the card-id-token guard.
Operator verdict:

### Q7: Should SessionStart reach the operator as well as the orchestrator, and how is the store derived there? (OD-9)

Label: FOUNDER
Class: product-level
Authority anchor: none.
Why unresolved: the existing advisories use one channel, and the repository's own doctrine and code disagree on which store key the host loads inside a linked worktree; adding a second channel or a shared resolver are both open. The SessionStart warning itself left this SPEC with the second split (`spec.md` §4); this row stays because it is the verdict the follow-up card inherits, and no requirement of this SPEC reads it.
Operator verdict: operator confirmed (via the leader's question channel), 2026-10-04: the warning goes through `additionalContext` only (no screen output).

### Q8: What should a repo-relative link become? (OD-10)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: the card says "rewrite or move, never delete" without choosing; an absolute path can itself be dead on another machine, and moving the line is the fold's job rather than a link repair's. This question belongs to the split-out half: the repo-relative-link rewrite moves to the follow-up card (leader ruling, option A, 2026-10-04).
Default: within this card a repo-relative link is left exactly as written and only reported as `MEMORY_REPO_RELATIVE_LINK` (rule: step 1 of the published default rule — the option that preserves current behavior, since the doctor today rewrites nothing; and `moai-memory.md` § Compressing the index — never fewer, which forbids removing or rewriting an entry's reach).
Alternate: rewrite to the absolute path when the file exists under the project root (the follow-up card's candidate), or move the line (the fold's job).
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec

### Q9: Does the fold preview by default or apply by default? (OD-6)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: the card asks for consistency with the existing memory verbs, but they differ: measured at the pinned tree, `moai memory drain` previews by default and takes `--yes`, while `moai memory archive` applies immediately.
Default: preview by default, `--yes` applies (rule: step 3 of the published default rule — the option with the smaller user-visible surface, since a preview writes nothing; the `moai memory drain` precedent, `.claude/rules/moai/workflow/moai-memory.md` § Worktree Mirror and Drain — "Preview by default (writes nothing)").
Alternate: apply by default, as `moai memory archive` does.
Operator verdict: DEFAULT-APPLIED 2026-10-04T10:59:44Z claude lane-18 via manager-spec
