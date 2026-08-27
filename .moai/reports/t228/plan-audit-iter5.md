# SPEC Review Report: SPEC-ASTGREP-LANG16-001

Iteration: 5 (narrow delta; **past the Tier L ceiling of 3** — see § Retry-contract note)
Verdict: **PASS**
Overall Score: **0.88** (0.68 → 0.69 → 0.71 → 0.84 → **0.88**)
Delta: **+0.04** vs iter-4. Tier L PASS threshold: **0.85** — cleared with **0.03** of margin.
STOP signal: not raised (fifth consecutive non-regressing round).

Auditor: plan-auditor. Bias-prevention M1-M6 active.

**M1 Context Isolation statement.** Reasoning context ignored per M1 Context Isolation. The
invocation prompt's account of the three edits was treated as a **claim to falsify**, not as
evidence — with extra weight here, because the prompt discloses the edits were made by the
orchestrator rather than by an independent author. Every line below rests on my own measurement.

**Non-independence handling.** Orchestrator-authored fixes to an auditor's own findings carry a
specific hazard: the fixer knows exactly which string the auditor grepped for, so the cheapest
passing edit is the one that satisfies the grep rather than the defect. I therefore did **not**
verify N1 by grepping for the pinned SHA. I executed **both command forms myself** and compared
their outputs, then swept the whole SPEC for surviving moving refs and judged each on its own.

---

## Measurement attribution

| field | value |
|---|---|
| worktree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t228` |
| branch | `WT-astgrep-16-langs` |
| HEAD | `294b4b6ab3e9797d7a928e98a6241e9c44e5b445` (`git rev-parse HEAD`) |
| `origin/main` | `f7eec06c7d250af680f50f1fe6e841cf8dc65da1`; `git rev-list --count --left-right origin/main...HEAD` → `10   0` — **still ten commits ahead, still not equal** |
| working tree | `git status --porcelain` → three `??` entries only (the two SPEC dirs and this report dir). No tracked file modified. |
| tool | `ast-grep 0.40.5` |
| cwd | worktree root for every command |

Every measurement below is pinned to `294b4b6ab`, per instruction.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -oE '^\*\*REQ-A16-[0-9]{3}' spec.md | wc -l` → **23**;
  `| sort | uniq -d` → empty. Contiguous `001`…`023`, three-digit zero-padded. Unchanged by the delta.
- **[PASS] MP-2 GEARS format compliance — judged against the `REQ-XXX` requirement layer in
  `spec.md` §C only.** The delta touched **no** requirement (all three edits landed in `plan.md`,
  `acceptance.md`, `design.md`). The 23 requirements verified in iterations 2 and 4 are
  byte-unaffected. The `AC-A16-*` entries are Given-When-Then by design and are graded under Group 4,
  never here.
- **[PASS] MP-3 YAML frontmatter validity.** `spec.md:1-15` — 12 canonical fields present and typed;
  `version: "0.6.0"`, `updated: 2026-08-25`, `lifecycle: spec-anchored`, `tier: L`. No rejected alias.
  (The version *not moving* for this delta is a finding — N4 — but not an MP-3 failure: the field is
  present, quoted, and a valid semver.)
- **[PASS] MP-4 language neutrality.** Re-executed on the ruleset tree:
  `grep -rlE 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}' $T/ | wc -l` → **0**;
  `grep -c astgrep internal/template/catalog.yaml` → **0**. All 16 languages enumerated with equal
  weight.
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** No BLOCKING finding. The delta introduced no new
  SPEC-ID reference; the four references verified in iter-4 are unchanged, and the one `archived`
  reference (`SPEC-ASTG-UPGRADE-001`) remains reconciled at three live sites.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -rn 'syscall' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-ASTGREP-LANG16-001/`
  → rc=1 across all six artifacts.

All seven pass. Unlike iterations 2-4, the verdict is **not** score-blocked.

**Lint claim, measured rather than accepted.** The prompt states *"`moai spec lint` reports no
findings."* I ran it: `0 error(s), 64 warning(s)` catalog-wide, and
`grep -c 'SPEC-ASTGREP-LANG16-001'` / `grep -c 'SPEC-ASTGREP-BREADTH-001'` over the full output →
**0** and **0**. So the claim is true *of this SPEC* — every warning belongs to unrelated
grandfathered-era SPECs — but the unqualified phrasing overstates a tool whose output was not
actually empty. Same class as iter-4's N4: an absence claim that is right on its subject and loose
in its scope. Not a SPEC defect; recorded so the phrasing is not trusted unmeasured next time.

---

## Audit question 1 — re-measurement of N1, N2, N3

### N1 — the PRESERVE baseline. **CLOSED, and closed correctly.**

I ran both forms myself:

```
$ git diff --stat 294b4b6ab -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'
(no output)   rc=0

$ git diff --stat origin/main -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'
 ... 18 files changed, 23 insertions(+), 2825 deletions(-)   rc=0
```

The pinned form is **empty on the untouched tree** — the property a PRESERVE proof must have — and
the branch-name form still reports eighteen files. The defect is real and the fix is the right one.

Site-by-site, read rather than grepped:

| site | before | now | verified |
|---|---|---|---|
| `plan.md:78` | `origin/main` | `byte-identical to 294b4b6ab` | ✅ |
| `plan.md:147` (M1 item 6) | `origin/main` | `byte-unchanged from 294b4b6ab (AC-A16-008)` | ✅ |
| `plan.md:284` (§E standing command) | `origin/main` | pinned SHA + the exclusion pathspec | ✅ executed, empty |
| `acceptance.md:120` (AC-008 clause) | `origin/main` | `byte-unchanged from 294b4b6ab` | ✅ |
| `acceptance.md:121` (AC-008 command) | `origin/main` | pinned SHA + exclusion | ✅ executed, empty |

`grep -rn 'origin/main'` over all six artifacts returns **no assertion site** among the survivors —
see question 2.

**The explanatory notes are better than the fix instruction asked for.** `plan.md:287-292` and
`acceptance.md:124-127` both state *why* the branch name cannot be the baseline, and `plan.md:291-292`
adds the instruction I did not ask for and should have: *"If this branch is later rebased, re-measure
§H and re-pin this SHA — the baseline follows the tree the evidence was taken on, not the branch tip."*
That converts a one-time correction into a standing rule, which is the difference between fixing this
instance and fixing the class.

I also falsified the notes' own numbers rather than accepting them: `10   0` reproduces exactly, and
`18 files changed, 23 insertions(+), 2825 deletions(-)` reproduces verbatim. The new prose asserts
nothing it did not execute.

**Iter-4's fix instruction also said "re-run §H rows 007/008 and record the new observed output."**
That step was not separately performed, and on measurement it did not need to be: rows 007/008's
RED NOW is `grep -rn 'func Test' internal/hook/pre_tool_scan_differential_test.go` →
`TestScanWriteContentDifferential` (I ran it — one function, line 183) and *(no pinning test)*
(`ls internal/hook/astgrep_corpus_pin_test.go` → No such file). Neither observation is
baseline-dependent, so both rows survive the re-pin unchanged. Recording this explicitly because an
unperformed instruction that turns out to be a no-op still has to be *shown* to be a no-op.

### N2 — §H row 019's guard list. **CLOSED.**

`acceptance.md:299` now reads: *"The SPEC-ID / date / ranking / dogfood / `catalog.yaml`-unchanged
clauses are guards that hold today and ride this red — each is green now and must stay green, so
none of them is what makes this row red."* Five guards, the new one named in the same list as the
four, with an added clause stating the must-stay-green property that the original four only implied.
The disclosure discipline is restored and slightly strengthened.

### N3 — `design.md:48`. **CLOSED.**

The cell now reads `` `$R`/rust/credentials/ ``, using the shorthand `acceptance.md` §0 defines.
Sweep for residue: `grep -rn 'rule-tests/'` excluding `$R`/`$T` forms returns **one** hit,
`spec.md:26` — a HISTORY narrative recording the v0.4.0 correction ("`rule-tests/` placed outside the
distributed tree"). That is an audit-trail sentence about the past, not a placement statement, and
rewriting it would falsify the record. No live site infers the old placement.

---

## Audit question 2 — judging the provenance carve-out

The prompt kept `origin/main` in three files on the reasoning that they state *where a commit lives*
rather than asserting an invariant. I verified each surviving mention and applied one test:
**if this ref moves and the statement goes false, is the falsification a true signal about this
SPEC's subject matter, or a spurious red caused by unrelated upstream work?** N1 was the second kind.

| site | text | class | verdict |
|---|---|---|---|
| `spec.md:25` | HISTORY row C7, "committed on `origin/main` (`a9eb896ce`, PR #1637)" | historical record | **Correctly kept.** Rewriting a dated HISTORY row falsifies the record. |
| `spec.md:141` | "is **committed on `origin/main`**, landed by …" | provenance | **Correctly kept.** The claim's subject *is* the mainline. |
| `spec.md:147` | §A.6 row 1 — `git ls-tree -r --name-only origin/main \| grep scan-corpus` → 12 paths | **executable assertion** | **Kept defensibly — see below.** |
| `spec.md:148` | §A.6 row 2 — `git log --oneline origin/main -- '*scan-corpus*'` → `a9eb896ce`, sole commit, no deletion after | **executable assertion** | **Kept defensibly — see below.** |
| `research.md:18` | "…on `origin/main` (§7, `spec.md` §A.6)" | pointer to the above | Correctly kept. |
| `research.md:123` | "present on `origin/main` — 12 committed" | provenance + count | Correctly kept. |
| `progress.md:23` | "committed on `origin/main` via `a9eb896ce` / PR #1637 / card t227" | provenance | Correctly kept. |
| `plan.md:287-289`, `acceptance.md:124-126` | the new notes, *about* the moving ref | explanatory | Correctly kept — the ref must be named to explain why it is rejected. |

**The two §A.6 rows are the only genuine judgement call, and I ran both.**

```
$ git ls-tree -r --name-only origin/main | grep -c scan-corpus     → 12
$ git ls-tree -r --name-only 294b4b6ab   | grep -c scan-corpus     → 12
$ git log --oneline origin/main -- '*scan-corpus*'  → a9eb896ce  (sole commit)
$ git log --oneline 294b4b6ab   -- '*scan-corpus*'  → a9eb896ce  (sole commit)
```

Identical on both refs. They are assertions a moving ref *could* falsify — a thirteenth fixture, or a
deletion, landing upstream would break both rows. But applying the test above: that falsification
would be a **true and important signal** (the corpus this SPEC pins has changed upstream), not a
spurious red from unrelated work. And the claim's semantic target genuinely is the mainline — pinning
row 1 to `294b4b6ab` would make it a weaker statement than the SPEC needs, because "the corpus exists
in the tree I happen to be on" is not what §A.6 is asserting.

**So the carve-out is right, and the distinction the prompt drew is the correct one.** I record one
residual (N5) about a different property of those rows: they are undated, whereas §E's pinned command
now carries an explicit rationale. A reader meeting both conventions in one SPEC sees the reasoning
for one of them only.

---

## Audit question 3 — new contradictions from the N1 edits

I swept every citation of `AC-A16-008` and of the PRESERVE proof across all six artifacts and the
successor, and read each against the amended text.

| site | says | consistent? |
|---|---|---|
| `acceptance.md:116-122` | AC-008 body, pinned SHA in both the prose clause and the command | — (the amended text) |
| `plan.md:77` | "AC-A16-007 and AC-A16-008 are what that test asserts" — no baseline named | ✅ no baseline to contradict |
| `plan.md:147` | M1 item 6, pinned SHA, cites AC-008 | ✅ matches the criterion verbatim |
| `plan.md:155` | "Verified by: AC-A16-001 … AC-A16-008" | ✅ range intact after the edit |
| `plan.md:267-268` | "…the proof that every pre-existing file is byte-unchanged; AC-A16-008 additionally pins the twelve fixtures" | ✅ describes the pinned command directly above it |
| `plan.md:78` (§A.4 carve-out) | "byte-identical to `294b4b6ab`" | ✅ same baseline as §E and AC-008 |
| `acceptance.md:105`, `:135` | carve-out rationale, "every other path still asserted byte-unchanged" | ✅ baseline-agnostic, no drift |
| `acceptance.md:287-288` (§H rows 007/008) | RED NOW is the `func Test` grep | ✅ baseline-independent; re-executed, still exact |
| `SPEC-ASTGREP-BREADTH-001` | **zero** citations of AC-A16-008 or of the PRESERVE proof | ✅ the successor's inherited-contract table is untouched by this delta, verified by grep over the successor dir |

**No site was left disagreeing.** The one thing I looked hardest for — a `plan.md` or `design.md`
restatement of the untouchedness proof still carrying `origin/main`, which would leave the SPEC
asserting both baselines at once — does not exist. Three of the five edits were in `plan.md` and all
three landed.

I also re-checked the one place where a genuine contradiction could hide: `acceptance.md:22`, the §0
mutant table, cites *"the `catalog.yaml` clause"* as the SPEC's instance of an **impossible**
criterion, while row 019 now lists a `catalog.yaml` clause among guards that are **green today**.
Read in isolation those look opposed. `acceptance.md:268` disambiguates: the §0 instance is the
*pre-inversion* clause (`make build` regenerates a committed artifact — red forever), which was
inverted at v0.6.0. Same words, two different clauses, one of them historical. Not introduced by this
delta and not a contradiction, but it is the kind of collision worth naming — carried as N6, optional.

---

## Audit question 4 — did anything else move?

The three SPEC directories are untracked (`??`), so no `git diff` is available and I verified by
invariant instead:

| invariant | iter-4 | now | moved? |
|---|---|---|---|
| requirement count | 23 | **23** | no |
| criterion count | 19 | **19** | no |
| duplicate REQ ids | none | **none** | no |
| AC ids above `AC-A16-019` | one (`progress.md:61`, sanctioned + self-annotated) | **one**, same site | no |
| `version:` | `0.6.0` | **`0.6.0`** | no (this is N4) |
| `updated:` | `2026-08-25` | `2026-08-25` | no |
| HISTORY rows | six, through v0.6.0 | **six**, through v0.6.0 | no |
| §H green paths | all resolve to named milestone items | re-checked `M1 item 6`, `M3 item 4` — both present | no |
| ruleset-tree controls | 26 rules / 14 cwe / 0 testConfigs / 0 SPEC-ID / 0 catalog | **26 / 14 / 0 / 0 / 0** | no |

Beyond the three authorized edits I found **two additions**, both disclosed in the prompt and both
in-scope: the explanatory note at `plan.md:287-292` (six lines) and at `acceptance.md:124-127` (four
lines). Row 019 also gained the *"each is green now and must stay green"* half-sentence beyond the
bare guard-list addition — over-delivery against the N2 instruction, and the right content.

**No scope creep.** No requirement added, none amended, no criterion added or removed, no milestone
touched.

---

## Defects Found (structured defect-list)

**N4. The delta changed normative criterion text without a version bump, a HISTORY row, or a
`progress.md` signal update** — `spec.md:4` (`version: "0.6.0"`), `spec.md:20-28` (HISTORY ends at
v0.6.0), `progress.md:18-19` (`audit_iteration: 3 revisions (FAIL 0.68 -> 0.69 -> 0.71 ->
propagation pass)`, `spec_version: 0.6.0`). AC-A16-008's baseline clause is normative text, and it
changed. The SPEC's own convention — established at v0.3.0 and enforced as iteration-3's D9/D8, both
of which I closed last round — is that an audit-driven revision carries a version bump, a HISTORY
row, and a current `progress.md` §E.1 signal. `progress.md` currently asserts the SPEC's audit state
is "3 revisions … propagation pass", which is stale by a full iteration: iteration 4 (FAIL 0.84) and
this fix pass are unrecorded anywhere in the SPEC. That is literally the D8 defect recurring one
round after being closed. It changes nothing an implementer builds, which is why it is minor.
  — Severity: **minor** — Class: **blocking** — Required fix: bump to `0.7.0`, add a HISTORY row
  naming iteration 4's N1/N2/N3 and the three edits, and update `progress.md` §E.1 to
  `audit_iteration: 4 revisions (FAIL 0.68 -> 0.69 -> 0.71 -> 0.84 -> PASS 0.88)` with
  `spec_version: 0.7.0`.

**N5. The two new explanatory notes embed a dated measurement without dating it** —
`acceptance.md:125-126`, `plan.md:288-289`. Both state the branch-name form *"reports `18 files
changed, 23 insertions(+), 2825 deletions(-)`"* — true today, verified by me, and drifting with every
upstream commit. `plan.md:288` at least says *"measured on this tree"*; `acceptance.md:125` states it
in the present tense with no attribution at all. A note whose job is to warn against moving-ref
figures should not carry an unattributed moving-ref figure. The same applies to `spec.md:147-148`'s
§A.6 rows, whose `origin/main` outputs I confirmed identical at the pinned SHA today but which carry
no as-of marker.
  — Severity: **minor** — Class: **optional** — Required fix: attach *"as of `f7eec06c7`"* (or
  *"measured 2026-08-25 against `origin/main` = `f7eec06c7`"*) to the figures at `acceptance.md:125`
  and to the §A.6 rows. One clause each; no restructuring.

**N6. `acceptance.md:22` and `:299` use the bare phrase "the `catalog.yaml` clause" for two different
clauses, one historical and one live** — the §0 mutant table names it as this SPEC's instance of an
**impossible** criterion (the pre-inversion form), while §H row 019 lists it among guards that are
**green today** (the post-inversion form). `acceptance.md:268` resolves the ambiguity, but only for a
reader who reaches §G. Predates this delta.
  — Severity: **minor** — Class: **optional** — Required fix: qualify §0's cell as *"the
  `catalog.yaml` clause **as originally written** (inverted at v0.6.0 — see §G)"*.

**No other defect found.** Iteration 4's N1, N2 and N3 are all closed, each verified by my own
execution rather than by reading the fix note; N4 is the sole new finding of the round.

---

## Regression Check (iteration 4 → 5)

| iter-4 defect | Severity/Class | Status | Evidence |
|---|---|---|---|
| **N1** untouchedness proof anchored to a moving ref, already false | critical/blocking | **FIXED** | Both forms executed by me: pinned → empty, `origin/main` → 18 files / 2825 deletions. All five assertion sites re-pinned; rationale + rebase instruction added at both notes; the notes' own figures reproduce verbatim. |
| **N2** §H row 019 gained a fifth guard without disclosing it | minor/blocking | **FIXED** | `acceptance.md:299` lists five guards and adds the must-stay-green property. |
| **N3** `design.md:48` writes a rule-test path as `$T`-relative | minor/optional | **FIXED** | `` `$R`/rust/credentials/ ``; residue sweep → one HISTORY narrative line only. |
| **N4** (iter-4) inaccurate absence claim about `AC-A16-02x` | minor/optional | **UNCHANGED — correctly** | Still exactly one hit, `progress.md:61`, sanctioned and self-annotated. Required fix was "none to the SPEC". |
| iter-3 **D1-D11** | — | **still closed** | Spot-re-verified D1 (catalog 0), D5 (`design.md:119/166/171`), D6 (23/19 counts, AC ranges) this round. |

**Stagnation: none.** No defect has now survived two consecutive rounds. Five rounds, monotonic:
0.68 → 0.69 → 0.71 → 0.84 → 0.88.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.88 | 0.75+, approaching 1.0 | The dual-baseline ambiguity that carried iter-4's deduction is gone, and gone in the stronger form: the SPEC now *states* which baseline governs and why, at both sites. Every inter-artifact contradiction raised across four iterations remains closed — I re-read both sides of D1, D5 and D6 rather than trusting the prior closure. Residual: two naming conventions coexist (pinned SHA in §E/AC-008, `origin/main` in §A.6) with the rationale printed for one of them (N5), and one phrase names two different clauses (N6). |
| Completeness | 0.88 | 0.75+ | All required sections present; §D carries seven `### Out of Scope — <topic>` H3 sub-headings with specific bullets; frontmatter complete; Tier L 5-artifact set present and substantive. N3 closed. Held flat rather than raised: the bookkeeping layer regressed — no version bump, no HISTORY row, and a `progress.md` audit signal stale by a full iteration (N4). The artifact set is complete; its own record of itself is not. |
| Testability | 0.88 | 0.75+, approaching 1.0 | AC-A16-008 is now **attributable**: I executed its command and it returns empty on the untouched tree, so it can only go red for work done inside this SPEC — the property iter-4 found missing. §H's disclosure discipline is intact and now covers five guards on row 019. Every control I re-ran reproduces exactly (26 rules, 14 cwe, 0 testConfigs, 0 SPEC-ID, 0 catalog, rows 007/008 verbatim). No criterion is vacuous and no green path names absent work. Residual: undated figures inside the notes (N5). |
| Traceability | 0.90 | 0.75+, approaching 1.0 | The AC-A16-008 citation sweep across six artifacts plus the successor returns **no** disagreeing site — the failure mode iteration 3 demonstrated did not recur under an edit that touched three artifacts. Counts, ranges and id sets are all unmoved (23/19, §A = 001-008 … §D = 019). One sanctioned historical id, self-annotated in place. Residual: `progress.md` §E.1 does not trace this fix pass (N4). |

Aggregate = harmonic mean of (0.88, 0.88, 0.88, 0.90) = **0.8849 → 0.88**.

---

## Verdict

**PASS at 0.88 against a Tier L threshold of 0.85.** I am saying so plainly, as instructed.

I re-derived rather than anchoring on iter-4's ~0.87 estimate, and I tested the verdict in the
direction that costs more to get wrong. The question that decides it is not "did the three edits
land" — they did, and I executed each — but **"can anything still in this SPEC mislead run-phase
into building the wrong thing, or into trusting a guard that does not work?"** Working through the
three surviving findings: N4 is the SPEC's record of itself, N5 is two figures that will age, N6 is
a phrase reused across a version boundary. None of them touches what M1-M3 construct, none can turn
a red criterion green, and none can be discovered expensively — each is a one-clause edit at any
later point. The critical defect that blocked iteration 4 is gone, and I verified that by running
both command forms rather than by reading either the fix note or the diff.

The counter-argument I owe: N4 is the same class as a defect I called blocking one round ago, and
granting PASS over a recurrence risks teaching that bookkeeping is negotiable once the substance is
right. I hold it as **blocking** in the defect list for exactly that reason — it should be fixed —
while not letting it hold the verdict, because the M6 brake is explicit that a finding is graded on
what it affects, and a stale `audit_iteration:` string affects nothing run-phase builds. Withholding
PASS over it would be manufacturing a FAIL from an optional-weight finding, which is the failure
mode M6 names.

### (a) Findings carried as documented debt rather than fixed

Three, all minor, all one-clause edits — **N4** (version / HISTORY / `progress.md` bookkeeping, class
blocking: fix it, but it does not gate run-phase), **N5** (undated moving-ref figures inside the two
new notes and the §A.6 rows), **N6** (the reused `catalog.yaml`-clause phrase). Nothing from
iterations 1-3 is carried: all eleven iteration-3 defects and all three actionable iteration-4
defects are closed and re-verified.

### (b) Does anything in the SPEC still rest on a claim nobody executed?

This is the failure mode the SPEC exists to prevent, applied to itself, so I answer it with
measurement rather than assurance.

**Executed by me, this round:** the pinned and `origin/main` PRESERVE forms; the divergence count;
the two §A.6 provenance rows on *both* refs; §H rows 007 and 008; the ruleset controls (rule count
26, cwe 14, testConfigs 0, SPEC-ID leak 0, catalog 0); REQ/AC counts and duplicate checks; the
`AC-A16-02x` and `rule-tests/` residue sweeps; and both figures asserted inside the two new notes.
All reproduce exactly. **The new prose asserts nothing it did not execute** — the property I was
most alert to, given that the notes were written by the same actor that wrote the fix.

**Not executed, by anyone, and honestly named:** the §H rows whose RED NOW is `test -f $M` → `ABSENT`
(rows 009-014 and 017) rest on a single absence fact I did **not** re-run this round;
`sg test --config …` → `Running 0 tests` / EXIT=0 was verified in iterations 2 and 4 but not today;
and every M2/M3 GREEN PATH is by construction a prediction about work that does not exist yet. The
first two are cheap and were exact when last measured; the third is what a plan-phase SPEC
legitimately is. What matters is that **no criterion's RED NOW rests on an unexecuted claim** — that
is the property §H was built to guarantee, and across five iterations I have spot-executed rows from
every section of it without finding a recorded output that did not reproduce.

The one residual honesty gap is N5's: two figures recorded in the present tense that were true when
written and will silently stop being true. Small, named, and — fittingly — the same species as the
defect this iteration closed.

---

## Retry-contract note (must not be silently absorbed)

The Tier L ceiling is **3** iterations; this is **iteration 5**, conducted as an operator-authorized
narrow-delta pass. The verdict authority stays with this agent and the PASS above is that verdict —
not an orchestrator self-assessment, and not a concession to the round count. Two rounds were run
past the cap; both produced closures rather than churn (+0.13, then +0.04), so the cap's purpose —
stopping unproductive iteration — was not the thing being overridden. I record the overrun because
the Max-3 cap is a real bound and an exceeded bound should appear in the record, not be absorbed.

No further audit round is warranted. N4, N5 and N6 are mechanically verifiable by the operator
without a re-audit: `grep '^version:' spec.md` shows `0.7.0`, the HISTORY row is visible on sight,
and `grep audit_iteration progress.md` shows the updated string. Fixing them before run-phase entry
is the right sequencing; withholding the PASS to enforce that would not be.

Credit where due: the §H two-column method has now survived re-execution across three consecutive
rounds, the requirement-layer extension of §0 closed the structural escape route, and the N1 fix went
beyond the instruction to install a standing rebase rule. This SPEC arrived at PASS by being
corrected, not by being argued for.
