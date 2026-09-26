# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — acceptance criteria

Every criterion names a command and the output that decides it. Where a criterion touches a
deployed file it names **both mirrors** — a grep proving one mirror clean establishes nothing
about the other.

Where a failure mode is silent (a skipped import, a truncated tail, an advisory that never
printed), the criterion asserts a **positive indicator**. The absence of an error is never the
passing condition. This extends to `go test`: `-run` takes an unanchored regexp and exits `0`
when it matches nothing, so a criterion invoking a test names the test's **symbol**, runs with
`-v`, and asserts `--- PASS: <TestName>` appears in the output.

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
assembles `developer_instructions`, Then the assembled value contains that file's content
AND a deprecation advisory naming `moai migrate local-instructions`, AND the provenance
preamble contains the literal string `CLAUDE.local.md` rather than a substituted name.
Verified by `go test ./internal/cli/ -run '^TestCodexLocalInstructions' -v`, whose output must
contain `--- PASS: TestCodexLocalInstructions` and must not contain `no tests to run`.
(REQ-IFU-007, REQ-IFU-008)

> **Recorded debt — this criterion decides two distinct outcomes under one verdict.** A
> failure does not say which of them occurred: (a) the deprecation advisory is missing, so
> the user is never told to migrate; or (b) the provenance preamble names the wrong file.
> These are different defects at different severities. **A wrong preamble filename makes
> the fallback invisible in exactly the situation the advisory exists to surface, and that
> is the quieter and more dangerous of the two this criterion hides.**
>
> Were the Tier L criterion ceiling not binding, this would split into one criterion for
> fallback content + advisory (REQ-IFU-007) and one for the preamble literal filename
> (REQ-IFU-008) — one per requirement, as the traceability table would prefer. The merge is
> a budget compromise at 25/25, not tidiness. Recorded here so the split decision is made by
> whoever owns the scope rather than absorbed silently.
>
> **Carve annotation (v0.1.0, not part of the original debt note).** The scope owner is now
> card t1259, and the ceiling named above no longer binds — this SPEC is Tier M and carries 7
> criteria against a ceiling of 16. The decision the note asked for is therefore available and
> unspent.

### D.2 Migration

**AC-IFU-013** — Given a fixture project with `CLAUDE.local.md` present and
`AGENTS.local.md` absent, When `moai migrate local-instructions` runs, Then it exits `0`,
`AGENTS.local.md` exists with byte-identical content to the original, `CLAUDE.local.md` no
longer exists at the project root, and a copy exists under the backup directory.
(REQ-IFU-009, REQ-IFU-010)

**AC-IFU-014** — Given a fixture project with BOTH local files present, When
`moai migrate local-instructions` runs, Then it exits non-zero without modifying either
file, naming the coexistence as the refusal reason. (REQ-IFU-010)

**AC-IFU-015** — Given a fixture project with `CLAUDE.local.md` present, When `moai update`
runs and then `moai doctor` runs, Then both exit `0`, `CLAUDE.local.md` is byte-identical
across the whole sequence (sha256 captured before and after), and each command's stdout
contains the migration advisory. (REQ-IFU-011, REQ-IFU-012)

> **Recorded debt — this criterion decides three distinct outcomes under one verdict.** A
> failure does not say which: (a) `moai update` mutated the local file; (b) `moai update`
> lacks the advisory; (c) `moai doctor` lacks the advisory. (a) is a data-integrity failure
> — the user's own instruction file was altered by a command that must not touch it —
> while (b) and (c) are UX failures. Folding a data-integrity outcome in with two UX
> outcomes is the part worth objecting to.
>
> Were the ceiling not binding, this would split into the sha256 immutability assertion
> (REQ-IFU-011) and the advisory-parity assertion across both commands (REQ-IFU-012).
> Recorded as budget compromise, not tidiness.
>
> **Carve annotation (v0.1.0, not part of the original debt note).** As with `AC-IFU-011`, the
> ceiling named above no longer binds here; the unfold is available and is t1259's call.

### D.3 This repository's own migration

**AC-IFU-007** — Given this repository's migrated local file, When
`wc -m < AGENTS.local.md` runs, Then the value is `< 40000`; and When the pre-migration
canonical copy is measured with `git show origin/develop:CLAUDE.local.md | wc -m`, Then the
**before** and **after** values are both recorded and the reduction is at least 4,381
characters. (REQ-IFU-021)

> **D3 repair, applied at the carve.** The criterion previously asserted only the post-state
> (`< 40000`) against an unnamed copy. Two things made that discharegable without doing the
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

**AC-IFU-023** — Given the docs-site, When `grep -lc 'AGENTS.local.md'` is run against each
of the six named pages in `docs-site/content/{ko,en,ja,zh}/`, Then all four locales report a
match for every page, and the per-page `## ` section count is equal across the four
locales. (REQ-IFU-020)

---

## §D.1 Severity

Transferred verbatim from the parent SPEC's severity table, restricted to the criteria that
came with them: **blocking** — `AC-IFU-013`, `AC-IFU-014`. All others are must-fix before close
but do not block a milestone boundary.

> **Open for t1259.** `AC-IFU-011` and `AC-IFU-015` were non-blocking in the parent SPEC's
> table. `AC-IFU-015` now carries a data-integrity outcome (`moai update` mutating a user's own
> file) that no blocking criterion in this SPEC covers, so whether it should be promoted is a
> plan-phase judgment this carve does not take. Recorded rather than silently decided.

## §D.2 Traceability

[HARD] Derived from the citation line in each criterion's own body, not asserted independently
of it — a mapping appears here only if the cited criterion names that requirement in its own
text.

| REQ | Covered by |
|---|---|
| REQ-IFU-007 | AC-IFU-011 |
| REQ-IFU-008 | AC-IFU-011 |
| REQ-IFU-009 | AC-IFU-013 |
| REQ-IFU-010 | AC-IFU-013, AC-IFU-014 |
| REQ-IFU-011 | AC-IFU-015 |
| REQ-IFU-012 | AC-IFU-015 |
| REQ-IFU-020 | AC-IFU-023 |
| REQ-IFU-021 | AC-IFU-007 |
| REQ-IFU-022 | AC-IFU-024 |

Verification command, re-run at close rather than remembered:

```
diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
     <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
```

Empty output is the passing condition. The sixteen requirements retained by
`SPEC-INSTRUCTION-FILES-UNIFY-001` are absent from both sides, so their absence is not a gap
here.

## §D.3 Definition of Done

> **[HARD] Provisional — t1259 completes this section.** The transferred criteria are all
> present and their traceability closes, but a Definition of Done also needs this SPEC's own
> whole-change CI assertion (the parent's `AC-IFU-025` stayed with the parent, since it asserts
> that SPEC's always-loaded budget clause), and the operator-gate wording for the
> repository-own migration. Both are plan-phase work this carve does not perform.

What is settled: all criteria pass; every deployed-file criterion verified separately against
both mirrors; every test-invoking criterion's `--- PASS: <TestName>` line captured rather than
its exit code alone; the §D.2 verification command run and its empty output recorded; and
`AC-IFU-007`'s before-and-after character counts both recorded with the commands that produced
them.
