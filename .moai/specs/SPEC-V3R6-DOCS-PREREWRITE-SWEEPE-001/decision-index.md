# Decision Index — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001

Revision 0.1.0 (2026-10-19). Each row states one decision the operator does not settle in the interview. The rows mirror spec.md § 6 field for field and carry the same `Operator verdict:` values; this index adds no verdict of its own. No row carries a recommendation. A `Default:` is a policy application, not a recommendation. This index follows the umbrella's convention (`.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/decision-index.md`); no sibling child carries a decision index yet.

Operator records read through `moai decision read --scope card:t1725`: d-20261019T015908Z-220a (Q13, Q17) and d-20261014T060405Z-0fcd (Q24, from the umbrella).

### Q13: Is each stale citation deleted or replaced?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a (decision board, `moai decision read --scope card:t1725`)
Why unresolved: the umbrella row (copy). Replace where a current document or command exists, otherwise delete.
Class: implementation-level
Child: README and SWEEPA to SWEEPF (this child applies it to its ten files)
Operator verdict: replace the stale citation with the current equivalent where one exists, delete where none exists — operator, d-20261019T015908Z-220a (AskUserQuestion, 2026-10-19, scope=standing)

### Q17: Do the same edits land in en, ja, ko, and zh in the same change?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a
Why unresolved: the umbrella row (copy).
Class: implementation-level
Child: cross-cutting (this child applies it to its ten files)
Operator verdict: all four locales (ko, en, ja, zh) change in the same card and the same change — operator, d-20261019T015908Z-220a (AskUserQuestion, 2026-10-19, scope=standing)

### Q24: Which narrative examples are exempt from the stale-citation check (N2)?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261014T060405Z-0fcd
Why unresolved: the full row stays in the umbrella; this child records the marker definition (REQ-PRC-001-001) and the measured marker state (neither line carries one).
Class: implementation-level
Child: SWEEPE
Operator verdict: exempt only narrative examples that carry an explicit example marker; lines without a marker are fixed — operator, d-20261014T060405Z-0fcd (2026-10-14T06:04:05Z)

### QE-1: How are the three feedback alias lines handled?
Label: FOUNDER
Authority anchor: none in the committed tree; operator record d-20261019T015908Z-220a (standing ruling, Q13, decision board). The ruling sets the rule; the row records the residual wording choice
Why unresolved: `fb` resolves only as a router alias; `bug` and `issue` do not resolve. The ruling requires replace-where-an-equivalent-exists; `/moai:feedback` is the registered equivalent for all three. Choice left: delete instead of replace.
Class: implementation-level
Default: replace with `> /moai:feedback` (rule: the Q13 ruling)
Alternate: delete the lines and keep only the standard form
Operator verdict: DEFAULT-APPLIED by Q13 ruling d-20261019T015908Z-220a

### QE-2: Does verification-claim-integrity line 84 stay fixed, or gain an example marker?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the line sits under the heading "Worked Example — Retention-Claim Hazard" and its text names retired commands as history. The Q24 verdict exempts only lines that carry a marker; adding one in this change would exempt the line the verdict says to fix. Whether a marker added by the sweep counts is the operator's reading.
Class: implementation-level
Default: fix without a marker, deleting the `/moai brain` and `moai brain` list items (rule: the option the Q24 verdict and Q13 ruling both name)
Alternate: keep the paragraph and add `<!-- slash-example: retired-feature incident narrative -->`
Operator verdict:

### QE-3: Is the closing check stricter than resolution?
Label: FOUNDER
Authority anchor: none in the committed tree
Why unresolved: the umbrella pattern also matches the router aliases `fb` and `init`, which do resolve. Keeping the pattern forces removal of the `fb` and `init` citations. A carve-out for router aliases is an umbrella amendment, not this child's.
Class: implementation-level
Default: keep the umbrella pattern unchanged (rule: preserves the umbrella's closing check)
Alternate: amend the umbrella pattern to exclude router aliases, then keep `/moai fb`
Operator verdict:
