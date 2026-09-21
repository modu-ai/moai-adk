# SPEC-AUDIT-EXPORT-CLAUSE-001 — acceptance criteria

> Every AC names a command and the output that closes it. Commands run from the
> repository root of the tree under test. A zero-count result is admissible only
> when the AC's paired positive control returned non-zero in the same run
> (§D.3).

---

## §A Given-When-Then scenarios

### AC-AEC-001 — the false sentence is gone from every hand-edited copy

**Given** the four agent definitions enumerated in SPEC §A.1,
**When** the repository is searched for the removed sentence,
**Then** the search reports zero matches across the `.md` layer.

```bash
grep -rc 'Never write the verdict to a gitignored location' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:0`.

Maps REQ-AEC-006

### AC-AEC-002 — the positive control fires

**Given** the same four files,
**When** they are searched for a string known to remain present,
**Then** every file reports a non-zero count, proving the instrument of
AC-AEC-001 can find text in these files at all.

```bash
grep -rc 'Export mandate' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher.

**This AC gates AC-AEC-001.** A zero here means the grep is broken or the
paths are wrong; AC-AEC-001's zeros are then meaningless and MUST NOT be read
as a pass. `Export mandate` is chosen because the pattern itself spans no
emphasis marker: it sits immediately after the clause's opening `**`, and a
literal search for it therefore matches whatever emphasis surrounds it. The
hazard the control rules out is a pattern that *crosses* a `**` boundary — such
a pattern silently matches nothing, and its zero is indistinguishable from
absence.

Maps REQ-AEC-006

> **Mapping corrected in v0.2.0.** This AC previously declared `Maps
> REQ-AEC-012`, which it does not test: it is AC-AEC-001's positive control and
> says nothing about whether introduced sentences are decidable or free of
> tree-state assertions. REQ-AEC-012's real test is AC-AEC-016. It now maps
> REQ-AEC-006 — the all-four-copies requirement whose instrument it actually
> validates.

### AC-AEC-003 — the forced stage is present in every copy

**Given** the repaired clause,
**When** the four files are searched for the verb the repair adds,
**Then** every file reports at least one match.

```bash
grep -rc 'git add -f' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher.

Maps REQ-AEC-001, REQ-AEC-006

### AC-AEC-004 — the check is present, with its flag

**Given** the repaired clause,
**When** the four files are searched for the check and its load-bearing flag,
**Then** every file reports at least one match of the flagged form.

```bash
grep -rc 'check-ignore -v --no-index' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher.

A match of `check-ignore` **without** `--no-index` in these files is a FAIL, not
a partial pass: the unflagged form answers a different question (SPEC §A.5).

Maps REQ-AEC-002

### AC-AEC-005 — the FORBIDDEN prohibition survives, per copy

**Given** that each copy carries its own FORBIDDEN wording,
**When** each file is inspected,
**Then** the plan-auditor copies still reference the convention's § Where
FORBIDDEN directory and the sync-auditor copies still name
`.moai/reports/plan-audit/`.

```bash
grep -c 'FORBIDDEN' .claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md
grep -c 'plan-audit/` is FORBIDDEN' .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: every line non-zero.

Maps REQ-AEC-005

### AC-AEC-006 — no copy asserts the ignore state

**Given** the two-state constraint (plan §B),
**When** **all six** scope files are searched for an assertion about whether the
destination is ignored,
**Then** no file contains a declarative form.

```bash
grep -rniE 'destination (is|is not) (git)?ignore' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — the same pattern relaxed to the hedged form the repair uses
MUST match, proving the regex reaches these files:

```bash
grep -rc 'may be ignore-matched' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
```

Expected: six non-zero lines.

> **Widened in v0.2.0** from four files to six, tracking REQ-AEC-004's widening
> from the export-mandate clause to every specified sentence. The two convention
> copies were outside this AC's file list in v0.1.0, which is why the tree-state
> assertion in the specified § Committing wording was not caught by any
> criterion.

Maps REQ-AEC-004

### AC-AEC-007 — the convention § Committing sentence is repaired

**Given** `.moai/docs/audit-artifact-convention.md`,
**When** the § Committing section is read,
**Then** the forced stage and the check both appear before the sentence
"They reach the integration branch with the card's evidence commit".

```bash
awk '/^## Committing/,/^## What makes/' .moai/docs/audit-artifact-convention.md
```

Expected output contains, in this order: `git check-ignore -v --no-index`,
`git add -f`, and only then `reach the integration branch`.

Maps REQ-AEC-003, REQ-AEC-007

### AC-AEC-008 — the two convention copies are byte-identical

**Given** the local convention doc and its template mirror,
**When** they are compared,
**Then** they are byte-identical.

```bash
cmp internal/template/templates/.moai/docs/audit-artifact-convention.md \
    .moai/docs/audit-artifact-convention.md && echo BYTE-IDENTICAL
```

Expected: `BYTE-IDENTICAL`, exit 0.

Maps REQ-AEC-010

### AC-AEC-009 — the FORBIDDEN bullet no longer rests on "is gitignored"

**Given** § Where of the convention doc,
**When** the FORBIDDEN bullet is read,
**Then** it states that nothing written there is force-staged, and does not
offer being gitignored as the distinguishing property.

```bash
awk '/^## Where/,/^## When/' .moai/docs/audit-artifact-convention.md | \
  grep -nE 'force-staged|deliberately gitignored'
```

Expected: a `force-staged` match present; `deliberately gitignored` absent.

Maps REQ-AEC-008

### AC-AEC-010 — the mechanical check covers branch reachability

**Given** § What makes the convention stick,
**When** the mechanical-check bullet is read,
**Then** it names a command that distinguishes present-on-disk from
present-on-branch.

```bash
grep -c 'git ls-files --error-unmatch' .moai/docs/audit-artifact-convention.md
```

Expected: non-zero.

Maps REQ-AEC-009

### AC-AEC-011 — the emitted codex layer matches its source

**Given** that the C2 agent copies were modified,
**When** the emission drift check runs,
**Then** it exits 0 without regenerating anything.

```bash
make agents-emit-check; echo "exit=$?"
```

Expected: `exit=0`.

Paired negative check — the emitted layer must no longer carry the removed
sentence either, which is the observable consequence of the regeneration
actually having happened:

```bash
grep -rc 'Never write the verdict to a gitignored location' \
  internal/template/templates/.codex/agents/moai/plan-auditor.toml \
  internal/template/templates/.codex/agents/moai/sync-auditor.toml
```

Expected: both lines ending `:0`. Control: `grep -rc 'Export mandate'` on the
same two files must be non-zero.

Maps REQ-AEC-013

### AC-AEC-012 — template neutrality on every template-tree edit

**Given** the three files changed under `internal/template/templates/`,
**When** **the lines this card adds** are swept for the forbidden content
classes,
**Then** no match is found in any class.

```bash
git diff -U0 -- \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md \
  | grep '^+' | grep -vE '^\+\+\+' \
  | grep -E 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|REQ-[A-Z]|CLAUDE\.local|/Users/|\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b|\b[0-9a-f]{9,40}\b'
echo "exit=$?"
```

Expected: no output, `exit=1`. **This is the sole closing command for this AC.**

Paired control — the same added-line extraction, filtered instead for a token the
repair is known to introduce, MUST match; otherwise the empty result above means
the diff pipeline reached nothing rather than that the classes are absent:

```bash
git diff -U0 -- \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md \
  | grep '^+' | grep -vE '^\+\+\+' | grep -c 'add -f'
```

Expected: non-zero.

> **The whole-file form was removed in v0.2.0, not demoted.** v0.1.0 stated the
> whole-file sweep first, with `Expected: no output, exit=1` attached to it, and
> called the diff-scoped form binding only in prose underneath. Measured against
> the pre-implementation tree, the whole-file form returns **19** matching lines
> in `internal/template/templates/.claude/agents/moai/plan-auditor.md` (the other
> two targets return 0) — rubric placeholders, documentation examples, and a
> date, all legitimate template content the regex over-matches. An implementer
> who stops at the first fenced block therefore reads a FAIL that is false by
> construction. Keeping the form as "informational" would leave the same trap one
> sentence further down, so it is deleted rather than annotated.

Maps REQ-AEC-011

### AC-AEC-013 — the specified wording's predicted consequence is itself true

**Given** the repaired clause's own instruction,
**When** a reader follows it against a throwaway path that does not yet exist,
**Then** the check runs its stated branch **and the plain `git add` behaves as
that branch predicts**.

```bash
P=.moai/reports/ZZAC013/plan-audit.md
mkdir -p "$(dirname "$P")" && printf 'ac-013 probe\n' > "$P"

git check-ignore -v --no-index "$P"; CHECK=$?; echo "check_exit=$CHECK"

git add "$P"; ADD=$?; echo "add_exit=$ADD"
git status --porcelain -- "$P"; echo "--- end staged ---"

git rm --cached -q "$P" 2>/dev/null; rm -rf "$(dirname "$P")"
```

The AC closes on the **pairing** of the two exit codes, not on either alone:

| `check_exit` | Required `add_exit` | Required staged state | Branch the clause describes |
|---|---|---|---|
| 0 (path is ignore-matched) | **non-zero** | nothing staged | the forced-stage branch — `git add -f` is the sanctioned remedy |
| 1 (path is not matched) | **0** | `A <path>` | the plain-stage branch |

Any other pairing is a FAIL: it means the clause predicts a consequence the tree
does not produce.

**Either row passes this AC**, and that is deliberate — asserting which row this
tree is in is exactly what REQ-AEC-004 forbids. What is *not* optional is that
the observed row's `add_exit` matches. Record both exit codes verbatim; a report
citing `check_exit` alone does not close this criterion.

> **Mutant probe (`verification-completeness.md` §2).** v0.1.0's form ran the
> check and read its exit code only, so a clause reading *"exit 0 means a plain
> `git add` will succeed and stage it"* satisfied it **byte-identically** while
> being false — and that mutant is precisely the v0.1.0 consequence clause, which
> shipped into the specified wording and was caught in review rather than here.
> Against the form above the mutant fails: on the `check_exit=0` row it requires
> `add_exit=0` and a staged file, and the measured tree gives `add_exit=1` with
> nothing staged. The AC now falsifies the defect it was named for.

Maps REQ-AEC-004, REQ-AEC-002, REQ-AEC-003

### AC-AEC-014 — the export obligation is not weakened

**Given** the repaired clause,
**When** the four copies are searched,
**Then** the incomplete-audit declaration still stands in each.

```bash
grep -rc 'incomplete audit' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four non-zero lines.

Maps REQ-AEC-015

### AC-AEC-015 — no ignore-policy change was made

**Given** the out-of-scope declaration,
**When** the change set is inspected,
**Then** no `.gitignore` appears in it.

```bash
git diff --name-only | grep -c 'gitignore'
```

Expected: `0`.

Maps REQ-AEC-014

### AC-AEC-016 — no introduced sentence asserts this repository's tree state

**Given** REQ-AEC-012, which until v0.2.0 had a mapping but no test,
**When** the lines this card adds across **all six** scope files are swept for
declarative tree-state forms,
**Then** none is found.

```bash
git diff -U0 -- \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md \
  | grep '^+' | grep -vE '^\+\+\+' \
  | grep -nE '\b(is|are|is not|are not) (ignore-matched|gitignored|tracked|untracked)\b'
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — the hedged forms the repair introduces MUST match, proving the
diff pipeline and the regex both reach the added lines:

```bash
git diff -U0 -- \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md \
  | grep '^+' | grep -vE '^\+\+\+' | grep -cE 'may be ignore-matched'
```

Expected: non-zero.

**Scope of the sweep is the added lines, not the files.** Every one of the six
files legitimately contains declarative tree-state sentences this card did not
write; REQ-AEC-012 binds sentences *this SPEC introduces*, so the diff is the
correct subject and a whole-file form would be false by construction (the same
defect AC-AEC-012 carried in v0.1.0).

> **New in v0.2.0.** REQ-AEC-012 previously had exactly one mapping — AC-AEC-002,
> which is AC-AEC-001's grep positive control and tests nothing the requirement
> asserts. The requirement was nominally covered and actually unverified, which is
> why a tree-state assertion reached the specified § Committing wording and
> survived to review. Had this criterion existed, it would have caught it before
> the wording was written.

Maps REQ-AEC-012

---

## §B Edge cases

| Case | Expected handling |
|---|---|
| A copy's FORBIDDEN wording differs from both variants in SPEC §C.1 | Blocker report, not an improvised third variant — the divergence is a finding about the copy |
| `cmp` reports the convention copies already divergent before the edit | Blocker report; the mirror is then edited section-by-section and the pre-existing divergence recorded, never flattened by copying |
| `make agents-emit-check` fails after `make agents-emit` | Emission is non-deterministic or the source layer is malformed; blocker, never resolved by editing a `.toml` |
| A sibling card lands mid-implementation and changes the ignore rules | No change to this SPEC. REQ-AEC-004 means the wording is unaffected; only AC-AEC-013's observed branch may flip, and both branches pass |
| A grep positive control returns zero | Every absence claim in the same run is void. Re-derive the instrument; do not report the zeros |

---

## §C Quality gates

- No `.gitignore` in the change set (AC-AEC-015).
- No file under `internal/template/templates/.codex/` in the change set as a
  hand edit; its presence is acceptable only as `make agents-emit` output.
- The card's own verdict is exported and force-staged per the repaired
  § Committing procedure, and `git ls-files --error-unmatch` on it exits 0
  before the card is reported complete.

---

## §D Definition of Done

### §D.1 Mandatory

1. AC-AEC-001 through AC-AEC-016 all pass, with each absence claim preceded in
   the same run by its positive control.
2. `make agents-emit-check` exits 0.
3. The two convention copies are byte-identical.
4. The verdict artifact for this card exists, is force-staged, and
   `git ls-files --error-unmatch` on it exits 0.

### §D.2 Evidence format

The card's verdict is written in the five-section evidence-bearing format
(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). Every AC above
appears with its command and its verbatim observed output — a summary line is
not evidence.

### §D.3 The zero-result rule

[HARD] No AC above may be closed on a zero-count result unless its paired
positive control returned non-zero **in the same run**. The clause under repair
carries `**bold**` runs, so a literal grep whose pattern crosses one of those
`**` delimiters matches nothing regardless of the file's content — which makes a
clean zero an ambiguous signal: it means either the target is absent or the
pattern never reached the file. The control separates those two readings, and without it
a passing verification and a broken instrument are indistinguishable.
