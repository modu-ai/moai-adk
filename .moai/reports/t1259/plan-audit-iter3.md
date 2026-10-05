# SPEC Review Report: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001

Card: t1259 · Iteration: 3/3 · **Scoped** — D3-residual and N1 only · Tier: L (threshold 0.85)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1259` · Branch `WT-local-instructions`
Subject: `3a040026f` (clean tree) · Prior: `28476f1a9` (iter2), `bac73d358` (iter1 repairs)
Date: 2026-09-26

**Verdict: FAIL**
**Overall Score: 0.88** (down from 0.92)
**Plan phase is NOT closeable as it stands.** One blocking defect, in four lines, all mechanical.

Scope per the dispatch: D3-residual and N1 only, plus the narrow no-regression trio. Iter2's 0.92
stands on everything else. The lead's evidence file `.moai/reports/t1259/d3-ci-surface.md`
(`0d7c7e44e`) was read as input and its two readings were re-derived here where the tree permits.

The score falls despite one item being cleanly repaired because the other is repaired **in the
criterion body only**, and the stale text it left behind sits in the two places a close actually
reads.

---

## N1 — REPAIRED. The positive control is live, and the author's self-correction was the right one.

The split is real and each command discriminates on its own. The absent symbol, in isolation:

```
$ unset MOAI_KANBAN … && go test ./internal/cli/ -run '^TestCodexLocalInstructions_FallbackAdvisory$' -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	(cached) [no tests to run]
```

The marker appears twice — the `testing:` warning line and the `[no tests to run]` suffix — so the
criterion's read fires. The present symbol, in isolation:

```
$ unset MOAI_KANBAN … && go test ./internal/cli/ -count=1 -run '^TestCodexLocalInstructions_DualFileMatrix$' -v
    --- PASS: TestCodexLocalInstructions_DualFileMatrix/body/body (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.656s
```

Its own `--- PASS: ` line present, no marker. The two commands now separate cleanly, which is what
N1 asked for.

**The author's correction of its own draft is the substantive part and it is correct.** I had
written "two exit codes" in the N1 fix instruction; that was wrong, and the measurement above shows
why — `go test` exits `0` and prints `PASS` on a selector matching nothing, so an exit-code read
passes a run that tested nothing. The marker is the discriminator. Recording that transcript inside
the criterion is better than what I asked for, and the generalization now stated there — *an
alternation is the wrong shape for a conjunction; `-run` takes a disjunction, so N symbols need N
commands* — is the right lesson at the right altitude.

One cosmetic note, not a defect: the recorded transcript shows `0.813s` where a re-run shows
`(cached)`. Both are real runs of the same command; nothing turns on it.

---

## D3-residual — NOT REPAIRED. The criterion body is fixed; the Definition of Done and the plan are not.

### What is repaired, and verified

Every factual claim the new `AC-IFU-031` text makes about the workflow surface holds:

```
$ grep -rn 'hugo\|vercel' .github/workflows/
(no output; exit 1)

$ sed -n '229p' .github/workflows/ci.yml
          go test -json -coverprofile=coverage.out -covermode=atomic ./... > test-stream.json || rc=$?

$ grep -n -A8 'push:' .github/workflows/spec-lint.yml
16:  push:
17-    branches:
18-      - main
19-      - develop
20-    paths:
21-      - '.moai/specs/**'

$ grep -n '^name:' .github/workflows/{docs-i18n-check,spec-lint,ci}.yml
docs-i18n-check.yml:1:name: docs i18n parity check
spec-lint.yml:1:name: SPEC Lint
ci.yml:1:name: CI
```

The five sibling jobs named (`lint`, `build`, `test-race`, `test-integration`,
`constitution-check`) all exist in `ci.yml`. The advisory reading of `docs-i18n-check` is
confirmed independently — header line 3 `# ADVISORY ONLY — NOT BLOCKING`, line 8
`Phase 1 (current): warn-only mode (DOCS_I18N_STRICT=0)`, line 11 naming the 35 existing drifts,
and the push branch of the strictness switch setting `strict=false`.

**Dropping the build clause rather than re-siting it was the right call**, and for the reason
given: re-siting onto a Vercel deployment needs the premise that the Vercel project deploys
`develop`, which is project-side configuration absent from this tree. Spending the same
defect shape — an evidence source asserted without being read — a third time to repair its second
occurrence would have been the worst available outcome. The three properties recorded in place of
"unmeasured" (different head, conditional via `ignoreCommand`, `github.silent: true`) are each
sufficient on their own, and none of them depends on the unread premise.

**No stale "unmeasured" claim about the build survives in the shipped text.** `grep -rn 'unmeasured'`
over the SPEC directory returns five hits: `acceptance.md:404` uses the word to say the reason
*improved past* "unmeasured", `acceptance.md:344` is about the v0.2.1 parity half, `spec.md:25` is
the HISTORY row, and `progress.md:178` / `:349-352` are the dependency note and the author's own
record of the correction. None asserts the build is unmeasured. That check passes.

### What is not repaired — the same claim, still live, in the two places a close reads

`AC-IFU-031`'s body has now been re-sited twice. The **Definition of Done that governs whether the
SPEC may close** was not touched either time:

```
$ grep -rn "PR head" .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/
acceptance.md:514:from this document. `AC-IFU-031` is read from the PR head's own CI run.
plan.md:140:`AC-IFU-031` (whole-change CI on the PR head) is read after all four milestones land.
plan.md:145:`go test ./...` locally. The full-suite verdict is CI's, on the PR head, in a clean environment.
plan.md:151:the PR head's CI run. The parent's `AC-IFU-025` stayed with the parent because it asserts that
acceptance.md:436-437, 445 — historical v0.2.1 note, correctly past-tense
progress.md:105, 247 — historical narrative, correctly past-tense
```

Four of these are live normative text, not history:

| Site | Text | Section |
|---|---|---|
| `acceptance.md:514` | "`AC-IFU-031` is read from the PR head's own CI run." | **§D.3 Definition of Done** |
| `plan.md:140` | "`AC-IFU-031` (whole-change CI on the PR head) is read after all four milestones land." | §D Close |
| `plan.md:145` | "The full-suite verdict is CI's, on the PR head, in a clean environment." | §E Self-verification |
| `plan.md:151` | "…the PR head's CI run." | §E |

This is the original D3 defect, unchanged, in its third location. The iter-1 finding said in as many
words that *the disclosure lived in `progress.md` while the unsatisfiable obligation lived in
`acceptance.md`, and the run phase reads `acceptance.md`* — the repair moved the criterion and left
the obligation standing sixty lines below it, in the same file, in the section whose entire job is
to say when the SPEC may close.

The consequence is concrete and is not a wording nit. A run phase that reaches close and reads §D.3
is instructed to obtain a verdict from a PR head that this lane, by `plan.md` §C's own rule, never
produces. The criterion body it would have to consult instead is not what §D.3 points at. Two
sections of one file now give different close instructions for the same criterion.

**Severity: major. Class: blocking.** It is four lines of mechanical edit, and it is the third
appearance of one defect, which is what makes it blocking rather than optional: the pattern is that
this claim is repaired wherever the audit points and left everywhere the audit did not.

**Required fix.** Replace all four with the criterion's own current wording — the workflow runs for
the `develop` head carrying the lane's merge SHA — and, while in §D.3, add the two obligations the
repairs established but did not propagate there: the `no tests to run` marker read (N1's finding —
§D.3 currently says only "`--- PASS:` captured rather than its exit code alone", which N1 showed is
necessary but not sufficient), and where the `docs i18n parity check` log reading is recorded.
Then `grep -rn "PR head"` over the SPEC directory should return only past-tense historical notes.

---

## New findings

**N2. Two figures in the new `AC-IFU-031` text do not reproduce.** —
`acceptance.md:377` and `spec.md:25` — Severity: **minor** — Class: **optional**

- The criterion cites `docs-i18n-check.yml:71-74` for `strict=false` on push. Measured, the
  assignment is at **line 75**; `71-74` is the `elif` branch head plus its three comment lines:

  ```
  $ grep -n 'strict=false\|event_name.*push' .github/workflows/docs-i18n-check.yml
  71:          elif [[ "${{ github.event_name }}" == "push" ]]; then
  75:            echo "strict=false" >> "$GITHUB_OUTPUT"
  ```

  The range identifies the right branch and stops one line short of the line carrying the claim.
  `plan.md` §B's own rule — *"Source locations are cited by symbol, not by line"* — applies.

- The v0.2.2 HISTORY row says the `hugo\|vercel` grep found nothing "across 20 files". Measured:

  ```
  $ ls .github/workflows/*.yml .github/workflows/*.yaml | wc -l
  19
  ```

  The load-bearing claim is unaffected — the grep is recursive over the directory and covered all
  19, and its empty result reproduces exactly. Only the decorative count is wrong.

Neither changes a pass condition. They are reported because an unverified figure inside a criterion
is the class this SPEC has now repaired three times, and a small one written during the repair of a
large one is worth seeing.

---

## Judgments referred to this audit

**Is "has RUN and its log is read" mechanically evaluable?** Split verdict, and on balance
acceptable.

*"Has run"* is fully mechanical — a run for a workflow at a given head either exists or does not,
and the criterion names both the workflow file and the head. *"Its log is read"* is not decidable by
a command, and it would be a judgment term smuggled into a criterion **if it stood alone**. It does
not: it is paired with "recorded, not gated" and with the instruction to state the path-filter
coupling when recording. That makes the discharge artifact a written record, which is the same shape
as `AC-IFU-007`'s before/after recording obligation — a shape I accepted at iter1 and should accept
here for the same reason. Naming a check at its true weight, with a recording duty instead of a
false gate, is more honest than either promoting it to a pass condition it cannot bear or omitting
the one docs-side signal that demonstrably fires.

The gap is small and named in the required fix above: the criterion does not say **where** the
reading is recorded, whereas `AC-IFU-007`'s analogue does and §D.3 names the close for it. Fix that
in the same pass as the §D.3 edit.

**Is writing the uncovered-docs residual into the criterion sufficient, or does it need an owner?**
Sufficient as written, for this SPEC. The paragraph is accurate — M4's 24 locale files sit behind no
blocking content check, because the Go suite says nothing about docs and the parity check cannot
fail the run — and it correctly points at `AC-IFU-023` as the criterion that actually decides M4's
docs work. `AC-IFU-023` is not weak: it asserts 24 non-zero greps with prior existence checks plus
the equal-delta clause, all mechanically decidable, and that is a real gate even though it runs in
the lane rather than in CI.

An owner is warranted for the *general* condition, not for this SPEC's slice of it: flipping the
parity check to Phase 2 strict once the 35 baseline drifts clear is a repository-wide change with
its own blast radius, already tracked (`SPEC-V3R3-DOCS-PARITY-001` is named in the workflow's own
comment at `docs-i18n-check.yml:74`). Requiring this card to own it would be the scope creep the SPEC
has correctly refused twice already. **Recommend: no new owner from this SPEC; note the existing
tracker in the paragraph so the reader can follow it.** That is a one-clause addition, optional.

---

## No-regression trio — re-run, not accepted

```
$ grep -c '^\*\*AC-IFU-[0-9]\{3\}\*\*' acceptance.md        → 10
$ grep -c '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md               → 9   (10 clauses)
$ diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
       <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
(empty; exit 0)
$ unset MOAI_KANBAN … && ~/go/bin/moai spec lint SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001
✓ No findings — all SPEC documents are valid
```

All three hold. The repairs added no criterion, moved no id, and broke no lint rule. Note that lint
passing is not evidence against the D3-residual finding: no rule in the engine compares §D.3's prose
to a criterion body.

---

## Score

| Dimension | iter2 | iter3 | Why |
|---|---|---|---|
| Clarity | 0.95 | 0.90 | Two sections of `acceptance.md` now give different close instructions for `AC-IFU-031`; `plan.md` agrees with the superseded one. |
| Completeness | 0.95 | 0.95 | Unchanged. |
| Testability | 0.80 | 0.85 | N1 closed and the control verified live; `AC-IFU-031`'s body is now grounded against measured workflow files. Held back by §D.3 pointing elsewhere. |
| Traceability | 1.00 | 0.85 | The §D.2 REQ↔AC mapping is intact, but `AC-IFU-031`'s evidence source is stated three ways across two binding artifacts. |

Harmonic mean: `4 / (1/0.90 + 1/0.95 + 1/0.85 + 1/0.85)` = `4 / 4.5487` = **0.879** → **0.88**.

Above the 0.85 threshold on aggregate, and the verdict is still **FAIL**, because the M5 firewall is
not an average: a blocking defect that makes the Definition of Done unfollowable is not compensable
by scores elsewhere. There is no score regression across iterations in the sense the LEAN clause
means (0.85 → 0.92 → 0.88 is not a deteriorating SPEC; it is one finding surfacing late), so no STOP
escalation is proposed.

---

## Gaps carried forward — carrying is acceptable, with one condition

- **Whether Vercel deploys `develop`** — genuinely unreadable from this tree, and the repair was
  deliberately built not to need it. Carrying is correct.
- **Both external readings decay silently** (unprotected `develop`; conditional Vercel build).
  Pinning them to `0d7c7e44e` and marking "re-read at close" is the right handling — it is what
  `verification-claim-integrity.md` §2 asks for, and the alternative (citing them as current) would
  be the unattributed-claim defect. **The condition:** the re-read duty currently lives in a note
  inside `AC-IFU-031`; §D.3 does not carry it. Fold it into the §D.3 edit, or the close will not
  perform it.
- **Full `internal/cli` state unrun; parent-SPEC M2 and t1175 unmet** — unchanged, out of scope for
  a scoped iteration, and the latter two are the lead's dispatch read.

## Gaps in this audit

- I did not query GitHub. Both of the lead's `gh` readings are taken as reported; what I re-derived
  is only what the tree holds (workflow files, `vercel.json` is quoted in the evidence file, not
  independently read here).
- Nothing outside the two scoped items was re-audited. Iter2's findings on D1, D2, D4, D5, D6 stand
  unretested, which is what a scoped iteration means.

## Recommendation

**FAIL, and it is one edit away from closeable.** N1 is fully repaired and its self-correction
improved on the instruction it was given. D3-residual's criterion body is now the best-grounded text
in the file — every claim in it reproduces. The failure is that the repair stopped at the criterion.

Before close, in one pass:

1. **Replace the four live "PR head" sites** (`acceptance.md:514`, `plan.md:140,145,151`) with the
   criterion's current wording. Then `grep -rn "PR head"` should return only past-tense notes.
2. **While in §D.3**, add the three obligations the repairs established there: the `no tests to run`
   marker read, where the `docs i18n parity check` log reading is recorded, and the close-time
   re-read of the two decaying external readings.
3. Optional: N2's two figures, and the `SPEC-V3R3-DOCS-PARITY-001` pointer in the uncovered-docs
   paragraph.

Item 1 is the blocker; items 2 and 3 are cheap enough to travel with it. A confirming pass should be
scoped to a single `grep -rn "PR head"` plus a read of §D.3 — the iteration ceiling is reached, so
if the lead prefers, that grep returning only historical hits is a sufficient close condition
without a fourth full audit.

🗿 MoAI
