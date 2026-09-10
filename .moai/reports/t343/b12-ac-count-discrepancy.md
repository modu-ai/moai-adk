# t343 — B12 AC-count discriminator: a non-zero but wrong count

Recorded for card **t338** (AC-count discriminator, still open) at the lead's
request. **No card issued from here** — this is evidence for t338's closing
verdict, not a new finding to act on.

tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t343`
branch: `WT-red-now-threshold`
measured at: `42dd07a0d` (t343 sync close, before absorbing `ee50984ab`)
subject file: `.moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md`

## The claim

The B12 counter returns **17**. The SPEC has **16** acceptance criteria. The
count is non-zero, so it is not the vacuous-pass shape the discriminator was
built to catch — it is a different failure: a plausible number that is wrong in
**both directions at once**, and whose two errors partially cancel.

## Evidence

```
$ grep -oE 'AC-[A-Z0-9-]*[0-9]' \
    .moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md | sort -u
AC-5
AC-6
AC-RNT-001
AC-RNT-002
AC-RNT-003
AC-RNT-004
AC-RNT-005
AC-RNT-006
AC-RNT-007
AC-RNT-008
AC-RNT-009
AC-RNT-010
AC-RNT-011
AC-RNT-012
AC-RNT-013
AC-RNT-014
AC-RNT-015
```

17 lines. The authoritative count, measured against the table rows themselves:

```
$ grep -c "^| \*\*AC-RNT-" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md
16
```

16 matches `acceptance.md` §D.1 (13 release-blocking + 3 regression-guard) and
the 16-row PASS matrix in `progress.md` §E.2.

## Why the two numbers differ — two independent errors

**Over-count (+2): quoted foreign identifiers.** `AC-5` and `AC-6` are
`plan-auditor.md` Group 4 checklist identifiers, quoted *inside* AC-RNT-007's
cells because that criterion is about the Group 4 checklist gaining an item.
They are the subject matter of a criterion, not criteria of this SPEC. Any SPEC
whose subject is the audit machinery itself will quote identifiers shaped like
its own.

**Under-count (−1): lettered criteria collapse.** `AC-RNT-009` is carried as two
distinct criteria, `AC-RNT-009a` and `AC-RNT-009b` (the both-direction fixture
pair — a planted violating fixture that must be reported, and a legitimate one
that must not). A pattern terminating at a digit truncates both to `AC-RNT-009`,
and `sort -u` folds them into one row.

17 − 2 + 1 = 16.

## Why this shape matters to t338

The discriminator distinguishes a real count from a vacuous zero. This case
passes that test — 17 is not 0 — while still being wrong. The two errors are
independent, so they cancel only by coincidence at this size; a SPEC with three
lettered pairs and one quoted identifier would land further off, still non-zero,
still plausible.

Both error sources are structural rather than incidental:

- Quoting foreign `AC-N` identifiers is normal in any SPEC whose subject is the
  acceptance-criteria machinery — which is exactly the class of SPEC that will
  be counted most often.
- Lettered criteria (`009a` / `009b`) are the standard way to express a
  both-direction pair, and the both-direction pair is itself required doctrine
  (`verification-completeness.md` §1.1 — confirming only the pass direction is
  indistinguishable from not having built the check).

So the counter is least reliable precisely where the doctrine is best followed.

## Disposition in t343

Not adopted. The CHANGELOG entry states **16**, derived from the table-row
count. The B12 self-test was recorded as returning 17 with the discrepancy
explained rather than reported as a pass.

## Gaps

- Only this SPEC's `acceptance.md` was measured. No sweep was run to find how
  many other SPECs the pattern miscounts, or in which direction.
- The B12 clause text itself was not re-read at the tree where it currently
  lives; the pattern used above is the one that produced the 17 in this run.
- Whether t338's landed discriminator already accounts for either error source
  was not checked — that is t338's own surface to judge.
