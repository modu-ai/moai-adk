# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — acceptance criteria

Every criterion names a command and the output that decides it. Where a criterion touches a
deployed file it names **both mirrors** — a grep proving one mirror clean establishes nothing
about the other.

Where a failure mode is silent (a skipped import, a truncated tail, an advisory that never
printed), the criterion asserts a **positive indicator**. The absence of an error is never the
passing condition. This extends to `go test`: `-run` takes an unanchored regexp and exits `0`
when it matches nothing, so a criterion invoking a test names the test's **symbol**, runs with
`-v`, and asserts `--- PASS: <TestName>` appears in the output.

> **[HARD] Anchor both ends, on the pattern and on the asserted line.** Naming the symbol is
> necessary and not sufficient: a head-only anchor (`'^TestFoo'`) matches every sibling starting
> with `TestFoo`, and the asserted `--- PASS: TestFoo` is satisfied as a **substring** of
> `--- PASS: TestFoo_Bar (0.00s)`. So every `-run` pattern here is anchored at both ends
> (`'^TestFoo$'`, or an alternation of both-end-anchored symbols) and every asserted line carries
> the single space Go prints after the test name (`` `--- PASS: TestFoo ` ``).
>
> **[HARD] The class is judged by a `moai spec lint` rule, not by a command written here.** Two
> successive attempts to record the class check as a prose `grep` pipeline were themselves defective
> — the third plan-audit of the parent SPEC identified the cause as structural rather than as two
> authoring mistakes: a command living as text inside a document cannot record that it ran, cannot
> carry a positive control, keeps its scope inside the command rather than in its declaration, and
> sits in the same file it inspects. Pattern judgment therefore moves to **card t1269** —
> `VacuousAssertionRule` in `internal/spec/lint_vacuous_assertion.go`, registered in the rule slice
> in `internal/spec/lint.go`, whose `Check` receives one document at a time.
>
> [HARD] **The claim is forward-looking and t1269 has not landed.** What is claimed is that a SPEC's
> criteria **will** pass that rule once t1269 lands — not that any check runs against this file
> today. Until then the anchoring rule above is an authoring obligation enforced by review, not
> mechanically, and this block must not be read as citing a live check. The parent SPEC's
> `acceptance.md` head block is the canonical statement and its §D.3.1 carries the accounting;
> the prose commands this block used to carry are transferred to t1269, not deleted.
> `AC-IFU-011` arrived carrying the head-only defect (see its note) — a criterion transferred
> verbatim carries its defects verbatim too.
> **The general defect is wider than `go test`**: any assertion satisfiable by something other
> than the thing under test is vacuous. When writing a criterion, ask not "does this pass when
> the work is done" but "can this pass when it is **not**".

> **Carve note.** The seven criteria below are transferred **verbatim** from
> `SPEC-INSTRUCTION-FILES-UNIFY-001` at commit `1140bcd1d`, retaining their `AC-IFU-*` ids so
> existing cross-references and audit citations still resolve. `AC-IFU-007` additionally carries
> the D3 repair (which copy, and in which unit). The two **recorded debts** below are preserved
> exactly as authored — including their own statements of what they would split into. The Tier L
> ceiling that forced each fold no longer binds here (Tier M, 16/16, currently 7 criteria), so
> the headroom to unfold exists; **spending it is card t1259's plan-phase decision and is
> deliberately not taken by this carve.**

## §D AC Matrix

### D.1 Codex read order and fallback

**AC-IFU-011** — Given a fixture project carrying only `CLAUDE.local.md`, When the launcher
runs, Then that file's content reaches the launch AND a deprecation advisory naming
`moai migrate local-instructions` is emitted on the launcher's diagnostic stream — **not** inside
the `developer_instructions` payload (design.md §B). Verified by

```
go test ./internal/cli/ -run '^TestCodexLocalInstructions_FallbackAdvisory$' -v
```

whose output must contain `--- PASS: TestCodexLocalInstructions_FallbackAdvisory ` and must not
contain `no tests to run`. (REQ-IFU-007)

**AC-IFU-029** — Given the same fixture, When the launcher assembles `developer_instructions`,
Then the provenance preamble contains the literal string `CLAUDE.local.md` rather than a
substituted or normalized name. Verified by

```
go test ./internal/cli/ -run '^TestCodexLocalInstructions_DualFileMatrix$' -v
```

whose output must contain `--- PASS: TestCodexLocalInstructions_DualFileMatrix ` and must not
contain `no tests to run`. (REQ-IFU-008)

> **[HARD] v0.1.1 repair — the transferred pattern was vacuous.** The criterion arrived with
> `-run '^TestCodexLocalInstructions'`, head-anchored only, asserting `--- PASS:
> TestCodexLocalInstructions`. No symbol of that exact name exists. Measured 2026-09-26 against
> `653e53572`: that command runs 22 pre-existing sibling tests and the asserted string matches
> **74** of their PASS lines, none of which asserts a deprecation advisory — the criterion could
> not fail before the work or after it. This is the same class the plan-audit of `653e53572`
> found in the parent SPEC's `AC-IFU-010 [REF]` and `AC-IFU-012 [REF]`; it reached this file because the
> carve transferred the criterion verbatim, which is the correct transfer policy and is why the
> anchoring rule now binds at the top of this file rather than per criterion.
>
> **Both ends anchored, and the asserted line carries Go's trailing space**, so no longer
> sibling name can satisfy it. `TestCodexLocalInstructions_FallbackAdvisory` **does not exist
> yet and this criterion requires its creation** — it is the test that asserts the advisory
> naming `moai migrate local-instructions`, which is new behaviour this SPEC adds.
> `TestCodexLocalInstructions_DualFileMatrix` **does** exist and already asserts the provenance
> preamble's literal filename (`<!-- source: CLAUDE.local.md -->`), which is the `REQ-IFU-008`
> half; naming it keeps that half attributed to the test that actually carries it rather than to
> the new one.
>
> The recorded debt below is unchanged by this repair: the criterion still decides two outcomes
> under one verdict. Anchoring makes it capable of failing; it does not make it capable of
> saying which half failed. Naming one test per half is a partial mitigation and is noted as an
> input to the unfold decision, not as the unfold.

> **Debt discharged at v0.2.0 — the fold is undone.** The note this replaces recorded that the
> criterion decided two outcomes under one verdict: (a) the advisory is missing, so the user is
> never told to migrate; (b) the provenance preamble names the wrong file. It stated the merge
> was a **budget compromise at the Tier L 25/25 ceiling, not tidiness**, and asked that the split
> be decided by whoever owned the scope.
>
> That owner is card t1259 and the stated reason is gone: this SPEC carries 10 criteria against a
> Tier L ceiling of 25. The criteria are split one per requirement — `AC-IFU-011` for
> `REQ-IFU-007`, `AC-IFU-029` for `REQ-IFU-008` — which is what the traceability table prefers
> and which makes a failure say which of the two defects occurred. The quieter one the note
> singled out (a wrong preamble filename makes the fallback invisible in exactly the situation
> the advisory exists to surface) now has a verdict of its own.
>
> **The v0.1.1 anchoring repair is preserved through the split.** Each criterion keeps a
> both-end-anchored `-run` pattern and asserts the PASS line with Go's trailing space.
> `TestCodexLocalInstructions_FallbackAdvisory` still does not exist and this SPEC still requires
> its creation; `TestCodexLocalInstructions_DualFileMatrix` exists and already asserts
> `<!-- source: CLAUDE.local.md -->`.

### D.2 Migration

**AC-IFU-013** — Given a fixture project with `CLAUDE.local.md` present and
`AGENTS.local.md` absent, When `moai migrate local-instructions` runs, Then it exits `0`,
`AGENTS.local.md` exists with byte-identical content to the original, `CLAUDE.local.md` no
longer exists at the project root, and a copy exists under the backup directory.
(REQ-IFU-009, REQ-IFU-010)

**AC-IFU-014** — Given a fixture project with BOTH local files present, When
`moai migrate local-instructions` runs, Then it exits non-zero without modifying either
file (sha256 of each captured before and after and compared), naming the coexistence as the
refusal reason. (REQ-IFU-010b)

**AC-IFU-015** — [BLOCKING] Given a fixture project with `CLAUDE.local.md` present, When
`moai update` runs and then `moai doctor` runs, Then both exit `0` and `CLAUDE.local.md` is
byte-identical across the whole sequence — sha256 captured before the first command and after
the last, and the two digests compared. (REQ-IFU-011)

**AC-IFU-030** — Given the same fixture and the same sequence, When each command's stdout is
read, Then **each** of the two contains the migration advisory naming
`moai migrate local-instructions`. A criterion satisfied by one of the two is a fail: the
parity across both commands is the assertion. (REQ-IFU-012)

> **Debt discharged at v0.2.0 — the fold is undone, and the data-integrity half is promoted.**
> The note this replaces recorded three outcomes under one verdict: (a) `moai update` mutated the
> user's own instruction file; (b) `moai update` lacks the advisory; (c) `moai doctor` lacks the
> advisory. It objected specifically to **folding a data-integrity outcome in with two UX
> outcomes**, and recorded the merge as a budget compromise.
>
> The ceiling no longer binds, so the split is taken along exactly the line the note drew:
> `AC-IFU-015` keeps the sha256 immutability assertion alone (`REQ-IFU-011`), and `AC-IFU-030`
> carries the advisory parity across both commands (`REQ-IFU-012`). Outcome (a) is now separable
> from (b) and (c) at the verdict, which is what the objection asked for.
>
> **`AC-IFU-015` is promoted to blocking** (§D.1). The carve left this open and it resolves with
> the unfold rather than against it: while the criterion was a bundle, promoting it would have
> made two UX outcomes block a milestone boundary, which is why the parent's table left it
> non-blocking. Unfolded, the blocking half is exactly the data-integrity one — a command that
> must not touch a user-authored file altering it — and no other blocking criterion in this SPEC
> covers that outcome. `AC-IFU-030` stays non-blocking; a missing advisory is recoverable, a
> mutated file is not.

### D.3 This repository's own migration

**AC-IFU-007** — Given this repository's migrated local file, When `wc -m < AGENTS.local.md`
runs, Then the value is **at most 39,999**; and When the pre-migration canonical copy is
measured **at the milestone** with `git show origin/develop:CLAUDE.local.md | wc -m`, Then both
the **before** and the **after** value are recorded with the command that produced each, and the
reduction reported is their difference. No absolute before-value or reduction floor is asserted
here — see the v0.2.0 note. (REQ-IFU-021)

> **[HARD] v0.2.0 — the fixed floor is withdrawn, because the number it was fixed to moved.**
> v0.1.1 repaired an off-by-one in a reduction floor ("at least 4,381" → "4,382") derived from a
> measured before-value of 44,381. Re-measured in this worktree, 2026-09-26:
> `git show origin/develop:CLAUDE.local.md | wc -m` → **44,740**, and
> `git show develop:CLAUDE.local.md | wc -m` agrees. The before-value had moved by 359
> characters, which makes the repaired floor wrong in the other direction — it would now be
> satisfiable by a migration that lands at 40,358 and fails the cap.
>
> The class of defect is not arithmetic. `CLAUDE.local.md` is a live maintainer document that
> sibling cards keep editing, so **any** absolute figure written into a criterion about it is
> stale by construction, and an arithmetic repair to a stale constant makes it look authoritative.
> The criterion therefore asserts only the bound that does not drift — `after <= 39,999` — and
> requires the before-value to be **measured at the milestone** rather than read from here.
>
> The recording obligation is unchanged and is the load-bearing part: both values, each with its
> command. That is what stops the criterion being discharged by measuring an already-compliant
> copy (§F anti-pattern). What is dropped is only the predicted floor, which added no failure
> mode the `<= 39,999` bound does not already catch.
>
> Re-checked across every other numeric clause in this SPEC: `AC-IFU-023`'s locale-count equality
> and `AC-IFU-013`'s exit `0` carry no drifting constant, so this was the only instance.

> **D3 repair, applied at the carve.** The criterion previously asserted only the post-state
> (`< 40000`) against an unnamed copy. Two things made that dischargeable without doing the
> work: the SPEC named neither which `CLAUDE.local.md` copy migrates — and that file's §0
> exists precisely because its copies diverge — nor which unit the 40,000 cap measures, while
> the surrounding prose mixed 61,908 bytes with 44,381 characters.
>
> Both are now fixed. The copy is the **`develop`-committed** one, per that file's §0.1
> discriminant (§0.2 forbids citing an uncommitted working copy; §0.3 records `main`'s copy as
> a retired model). The unit is **characters** (`wc -m`) throughout. Measured 2026-09-26:
> `git show origin/develop:CLAUDE.local.md | wc -m` → **44,381**, so the canonical copy does
> not already satisfy the cap and the requirement is a real ~10% reduction. Requiring the
> before/after pair is what stops the criterion being discharged by measuring an
> already-compliant copy.

**AC-IFU-024** — Given the migrated repository-local file, When its §0 is read, Then it
names `AGENTS.local.md` as the file whose canonical copy it discriminates, and
`grep -c 'CLAUDE.local.md' AGENTS.local.md` reports only historical-reference occurrences,
each within a sentence marking it as a retired filename. (REQ-IFU-022)

### D.4 Documentation

**AC-IFU-023** — Given the docs-site, When `grep -c 'AGENTS.local.md'` is run against each of
the 24 files `docs-site/content/{ko,en,ja,zh}/<page>` for the six `<page>` **paths** named in
`REQ-IFU-020`, Then every one of the 24 reports a non-zero count, and the per-page `^## `
section count is equal across the four locales. All 24 paths are asserted to exist before the
grep runs — a missing file must fail the criterion rather than be skipped. (REQ-IFU-020)

> **[HARD] v0.2.0 — the criterion named a page that resolves to two files, and a directory that
> holds none.** Measured 2026-09-26 in this worktree: the six stems were cited against
> `docs-site/content/{ko,en,ja,zh}/`, which is **not** where any of them live — every page sits
> one or two directories deeper (`advanced/`, `getting-started/`, `cli-reference/`,
> `claude-code/context-memory/`), so a literal reading of the old glob matches nothing and the
> criterion passes vacuously on an empty set.
>
> Worse, `memory` resolved to **two** files per locale — `cli-reference/memory.md` and
> `claude-code/context-memory/memory.md` — and the criterion named neither. They are different
> pages: `grep -c 'CLAUDE.local.md\|CLAUDE.md'` reports **0** for `ko/cli-reference/memory.md`
> (the `moai memory` CLI reference) and **28** for `ko/claude-code/context-memory/memory.md`
> (the Claude Code memory-file concept page). The three-file instruction structure belongs to
> the latter, which is the one `REQ-IFU-020` now names by path.

**AC-IFU-031** — Given the whole change on its PR head, When CI completes, Then every required
check reports success, and the run includes both `go test ./internal/cli/...` and the docs-site
build. The verdict is read from the PR head's own run — a local pass, or a run against an
earlier head, does not discharge it. (REQ-IFU-007 … REQ-IFU-012, REQ-IFU-020 … REQ-IFU-022)

> **Authored at v0.2.0.** The carve recorded that this SPEC had no whole-change assertion of its
> own, the parent's `AC-IFU-025 [REF]` having stayed with the parent because it asserts that
> SPEC's always-loaded budget clause. This one asserts nothing about a budget: it is the
> cross-milestone integration check, and it is the only criterion here whose evidence comes from
> a clean environment rather than the lane's machine. Its requirement citation is deliberately
> the full set — it is a whole-change criterion, and §D.2 treats it as covering none of them
> individually.

---

## §D.1 Severity

**Blocking** — `AC-IFU-013`, `AC-IFU-014`, `AC-IFU-015`. Each carries a data-integrity outcome:
the migration losing content, the verb choosing between two user-authored files, or `moai update`
altering one. All others are must-fix before close but do not block a milestone boundary.

> **v0.2.0 — `AC-IFU-015` promoted, and the carve's open question closed.** The parent's table
> left it non-blocking, correctly, while it bundled one data-integrity outcome with two UX ones.
> Unfolded (§D.2 note under `AC-IFU-030`), the blocking half is exactly the data-integrity
> assertion and nothing else travels with it. `AC-IFU-029`, `AC-IFU-030`, and `AC-IFU-031`
> inherit the non-blocking severity of the criteria they were split from or, for `AC-IFU-031`,
> take it as a close-gate rather than a milestone gate.

## §D.2 Traceability

[HARD] Derived from the citation line in each criterion's own body, not asserted independently
of it — a mapping appears here only if the cited criterion names that requirement in its own
text.

| REQ | Covered by |
|---|---|
| REQ-IFU-007 | AC-IFU-011 |
| REQ-IFU-008 | AC-IFU-029 |
| REQ-IFU-009 | AC-IFU-013 |
| REQ-IFU-010a | AC-IFU-013 |
| REQ-IFU-010b | AC-IFU-014 |
| REQ-IFU-011 | AC-IFU-015 |
| REQ-IFU-012 | AC-IFU-030 |
| REQ-IFU-020 | AC-IFU-023 |
| REQ-IFU-021 | AC-IFU-007 |
| REQ-IFU-022 | AC-IFU-024 |

Every requirement is covered by exactly one criterion, and every criterion covers exactly one
requirement — the one-to-one the two unfolds at v0.2.0 produced. `AC-IFU-031` is deliberately
absent from this table: it is the whole-change CI criterion and cites the full requirement set,
so counting it as coverage would let it stand in for a missing per-requirement criterion.

`REQ-IFU-010`'s two sub-clauses appear as separate rows because they have different correct
outcomes and therefore different criteria (spec.md §C.2, design.md §A). The verification command
below matches on the three-digit id, so both rows fold to `REQ-IFU-010` on its left-hand side.

Verification command, re-run at close rather than remembered:

```
diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
     <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
```

Empty output is the passing condition. The sixteen requirements retained by
`SPEC-INSTRUCTION-FILES-UNIFY-001` are absent from both sides, so their absence is not a gap
here.

## §D.3 Definition of Done

All ten criteria pass. Every deployed-file criterion is verified separately against both mirrors
with separate exit codes. Every test-invoking criterion's `--- PASS: <TestName> ` line is
captured rather than its exit code alone. The §D.2 verification command is run at close and its
empty output recorded. `AC-IFU-007`'s before-and-after character counts are both recorded with
the commands that produced them, the before-value measured at the milestone rather than read
from this document. `AC-IFU-031` is read from the PR head's own CI run.

**The operator gate on the repository's own migration (M3) is discharged in the lane, by the
operator, before that milestone starts** — not inferred from the earlier milestones having gone
well, and not from this document. The lane records the confirmation with the turn it arrived in.
