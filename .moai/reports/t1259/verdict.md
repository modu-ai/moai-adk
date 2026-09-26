> **run 선행: SPEC-INSTRUCTION-FILES-UNIFY-001 M2, SPEC-ALWAYS-LOADED-DIET-002**

# t1259 — plan-phase close verdict

Card **t1259** · SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 · Tier L (threshold 0.85)
Branch `WT-local-instructions`, worktree `.claude/worktrees/t1259`, final HEAD **`5271f009f`**, unpushed.
Written by the lane. **The close call is the lead's, read from this file.**

---

## 1. Why this file exists instead of a fourth audit

The plan-audit ceiling of three iterations was reached:

| Iter | Verdict | Commit |
|---|---|---|
| 1 | PASS-WITH-DEBT 0.85 — 4 blocking | `b8fb023e8` |
| 2 | PASS-WITH-DEBT 0.92 — 5/6 repaired, D3 partial, new N1 | `28476f1a9` |
| 3 (scoped) | **FAIL 0.88** — N1 repaired, D3-residual not | `8d73a2a88` |

iter3's FAIL was not a score failure: 0.88 sits above the 0.85 threshold. The auditor failed it
because a Definition of Done that cannot be followed is not compensable by scores elsewhere, and
recorded that this was a single late finding rather than a deteriorating SPEC, so no STOP applied.

**The lead ruled (2026-09-26) that the close is decided by a mechanical condition rather than a
fourth audit.** That substitution is recorded here because it is a departure from the normal close
path: the audit ceiling was not waived and no fourth opinion was taken — the remaining defect was
propagation, not judgment, so the lead made the predicate mechanical and reserved the call.

## 2. The close predicate, and one property of it

The predicate is `grep -rn "PR head"` over the SPEC directory. Its passing condition is **"no live
normative hit", judged per hit** — never "no output", and never a count.

The distinction is load-bearing, not pedantry. The grep **cannot** return empty by construction:
the sentence stating the condition contains the search string, and every repair record of this
defect has to quote the wording it removed or the record says nothing. A reader expecting zero
output would read a clean SPEC as failing.

**Positive control.** A predicate that finds nothing looks identical to a predicate that is broken,
so the same grep was run against the pre-repair commit first. Run by the lane, this run:

```
$ git grep -n "PR head" 8d73a2a88 -- .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/
```

Nine hits, of which **five are live normative sentences** — the instrument fires:

```
acceptance.md:514  from this document. `AC-IFU-031` is read from the PR head's own CI run.
plan.md:140        `AC-IFU-031` (whole-change CI on the PR head) is read after all four milestones land.
plan.md:145        `go test ./...` locally. The full-suite verdict is CI's, on the PR head, in a clean environment.
plan.md:151        the PR head's CI run. The parent's `AC-IFU-025` stayed with the parent because it asserts that
progress.md:105    **(6) Whole-change CI criterion — authored as `AC-IFU-031`.** Read from the PR head's own CI run,
```

The audit had found four of these; `plan.md:145` was the fifth, missed because it is a general
statement about where the full-suite verdict comes from rather than a sentence about `AC-IFU-031`.

## 3. Post-repair measurement

```
$ git grep -n "PR head" 5271f009f -- .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/
```

**16 hits. Zero live normative sentences.** Every hit classified:

| Line | Class | Why it is not normative |
|---|---|---|
| `acceptance.md:436` | repair-record | v0.2.1 note — "the evidence **was** sited on a PR head" |
| `acceptance.md:437` | repair-record | quotes the superseded wording: "As authored it **read**…" |
| `acceptance.md:515` | **negation** | amended §D.3 — "**not** from a PR head" |
| `plan.md:162` | repair-record | "the original 'PR head' wording **proved to name** something this regime never produces" |
| `progress.md:106` | repair-record | v0.2.0 authoring record, marked **both superseded** |
| `progress.md:249` | repair-record | quotes the old text inside the v0.2.1 record |
| `progress.md:457` | repair-record | narrates the iter3 FAIL, quoting the grep string |
| `progress.md:469-472` | repair-record | the was/now table — the "was" column must quote what was removed |
| `progress.md:508` | condition-statement | states the close condition, so contains its own search string |
| `progress.md:517` | repair-record | class-table row describing the `:515` negation |
| `spec.md:24`, `:26` | HISTORY | v0.2.1 and v0.2.3 rows |
| `spec.md:38` | condition-statement | states the close condition |

All five control sites are repaired, and **re-pointed rather than paraphrased** — each instruction
now names the `origin/develop` head carrying the lane's merge SHA. Stripping the phrase while
leaving the instruction pointing at a PR head would have passed the predicate and failed the SPEC.

## 4. A defect the predicate caught that it was not designed to catch

Running the predicate as written — reading its output rather than a summary of it — surfaced that
`da0df2113` had shipped `progress.md` with the iter3 block **duplicated**, 113 lines including a
surviving copy of the brittle-count table that block had just replaced. The file simultaneously
argued "the count is a trap, use the class table" and carried the counted table underneath.

Cause, as reported by the repairing agent: a splice whose end anchor (`### Verification after these
repairs`) was not unique, so the end index resolved before the start index and the span was
re-emitted instead of removed.

**None of the five green checks caught it.** The predicate grep was clean, the AC counter read
`10`, the traceability diff was empty, and `moai spec lint` passed — none of them reads prose
structure. It surfaced only because the grep returned line numbers in the 500s and 600s for text
that should exist once.

Repaired at `5271f009f`. Verified by the lane, this run:

```
$ wc -l < progress.md                                  → 580   (was 668)
$ sort -u progress.md | wc -l                           → 440
$ grep -c '^### plan-audit iter3 (scoped)' progress.md  → 1     (was 2)
$ grep -c '^### Verification after these repairs' progress.md → 2  (was 3 — iter2 and iter3, one each)
```

The class table referenced by `spec.md:38` survives the de-duplication (2 matching rows present).

## 5. Amended §D.3 Definition of Done — verbatim at `5271f009f`

```
## §D.3 Definition of Done

All ten criteria pass. Every deployed-file criterion is verified separately against both mirrors
with separate exit codes. The §D.2 verification command is run at close and its empty output
recorded. `AC-IFU-007`'s before-and-after character counts are both recorded with the commands that
produced them, the before-value measured at the milestone rather than read from this document.

**`AC-IFU-031` is read from the CI runs for the `origin/develop` head carrying this lane's merge
SHA** — not from a PR head. This lane opens no PR: under the git-flow lane protocol only
`release/vX.Y.Z` PRs to `main`, and that head carries many cards. See the criterion body for what
is read there and at what weight.

Four close-time duties, each established by a repair and each needing a home outside the criterion
that produced it — a duty recorded only inside a criterion's own note is a duty the close will not
perform:

1. **Every test-invoking criterion is read on two signals, not one.** Its `--- PASS: <TestName> `
   line is captured, **and** its output is confirmed NOT to contain `no tests to run`. The exit code
   alone is insufficient and so is the PASS line alone: `go test` exits `0` and prints `PASS` on a
   selector matching nothing, so the marker is what distinguishes a passing run from one that tested
   nothing (measured — `AC-IFU-029` v0.2.2 note).
2. **The `docs i18n parity check` log is read and its reading recorded** in `progress.md` §E.4 at
   close — the run's conclusion, its drift counts, and the fact that it is advisory. `AC-IFU-031`
   requires that the check *ran and was read*, never that it passed; a reading with no record is
   indistinguishable from no reading.
3. **The two decaying external readings are re-read at close, not cited.** `develop`'s protection
   state and Vercel's docs-site build configuration are both mutable outside this repository and
   nothing in the tree changes when they move. The values pinned in `AC-IFU-031`'s notes are pinned
   to `0d7c7e44e` and are evidence of what was true then. Re-measure both; where either has moved,
   the criterion's wording is re-read before the close proceeds.
4. **The docs-parity residual is restated at close, not quietly inherited.** M4 lands 24 locale
   files whose content no *blocking* check inspects (`AC-IFU-023` is the gate, and it runs in the
   lane). The general condition has an owner already — **`SPEC-V3R3-DOCS-PARITY-001`**, named in
   `docs-i18n-check.yml:74` as the Phase 2 strict-flip route. This SPEC does not own that flip and
   does not take it on; the close names the pointer so the residual is handed over rather than
   dropped.

**The operator gate on the repository's own migration (M3) is discharged in the lane, by the
operator, before that milestone starts** — not inferred from the earlier milestones having gone
well, and not from this document. The lane records the confirmation with the turn it arrived in.
```

## 6. No-regression measurements, re-run by the lane at `5271f009f`

```
$ AC counter (extracted from .claude/agents/moai/manager-docs.md)
live=10 excluded=3 ambiguous=0 ; 10 ; exit=0
$ grep -c '^\*\*AC-IFU-' acceptance.md
10
$ go test ./internal/spec/ -run TestACCounterFullCorpusMatchesBaseline -v -count=1
--- PASS   (this SPEC listed "absent-from-snapshot … COUNT 10" — report-only, not a
            cascade trigger, per .moai/docs/ac-count-baseline-refresh.md §2)
$ git status --short
(clean)
```

Counter and declared count agree. No AC-baseline cascade is owed: the file is new to the corpus, so
its row is absent rather than moved.

## 7. Gaps — not closed by assertion

- **Whether the Vercel project deploys `develop` at all** is unread. It is project-side
  configuration, absent from this tree. It is why the docs-site build clause was dropped rather
  than re-sited: re-siting would have rested on this unverified premise.
- **Two external readings decay silently** — `develop` being unprotected, and Vercel's conditional
  build. Both are mutable outside the repository and nothing in the tree changes when they move.
  Pinned to `0d7c7e44e`; §D.3 duty 3 requires re-measuring rather than citing them at close.
- **Both run-phase blocking dependencies are unmet** (`.moai/reports/t1259/blocking-dependencies.md`):
  the parent SPEC is `status: draft` with `codex_launcher.go:125` still iterating Claude-first, and
  `SPEC-ALWAYS-LOADED-DIET-002` is absent from `origin/develop`. This read decays and is re-run at
  dispatch, not carried from here.
- **`internal/cli` was not run in full**, and the no-regression trio does not read prose structure —
  §4 is the demonstration that it does not.
- **`moai spec lint`** passes scoped to this SPEC id (`✓ No findings`, exit 0, reported by the
  repairing agent and by the auditor independently at iter3). The lane did not re-run it at
  `5271f009f`; the two prior runs are at earlier commits.

## 8. Residual risk

- The predicate is a **string** detector. It establishes that no live sentence carries the phrase
  "PR head"; it does not establish that no instruction anywhere points at a PR head by another
  name. The audit's own reading of §D.3 is what covers that, and §5 reproduces the text so the lead
  can read it directly rather than trust the grep.
- `AC-IFU-007` carries two blockquotes dated the same day citing **44,381** (the carve record) and
  **44,740** (the v0.2.0 re-measurement). The v0.2.0 note explains the supersession and the
  criterion itself asserts no absolute figure, so no pass condition depends on either — but a
  reader meeting the older quote first could take 44,381 as current.
- Three audit iterations found one defect class recurring in a new location each round. The
  propagation sites are now measured clean, but the class has not been shown absent — only absent
  from the sites the predicate reaches.
