# Addendum — iter-2, round 2 (version freeze + systematic-drift judgment)

## Version under judgment

Re-read performed **2026-09-01 00:48:19 KST**. Judgment artifacts, hashes verified against the
lead's frozen set at that moment:

```
685e18a7acc253bdf4047bce203a6d82  spec.md         (mtime 00:18:30)
7c1ae339a7005540da0609fc17230b0c  plan.md         (mtime 00:22:03)
a07c9eddc3435ce220e0e54612315149  acceptance.md   (mtime 00:19:57)
```

All three match the lead's values exactly — **no third write occurred**, and the main report's
judgment stands on the same bytes it was written against. Two cited-evidence artifacts moved after
my spawn and are re-read here:

```
84b7830dc9332b66af5488bcb9463fba  progress.md              (mtime 00:26:17, re-read 00:48)
962c4b7db16e9c39bfd3701e2d0d775e  premise-remeasure.md     (mtime 00:25:49, re-read 00:48)
```

Structural counts re-confirmed independently at re-read: **REQ distinct 15 / AC distinct 13**, one
definition per id in both layers.

### Which version each finding rests on — divergence disclosure

- **`progress.md`** — the version I read during the main audit was **already the post-00:26:17
  version**: my transcript of it contains both the `[iter-2 정정] 리드의 대조 트리 … 리드가 값을
  철회했다` block and `편집 대상 파일 수 12 → **14**`. So the "12파일" stale value the lead reports
  was already corrected before I read it. **I never saw the stale version**, and nothing in the main
  report rests on it. No divergence.
- **`premise-remeasure.md`** — I had **not** read it during the main audit; the main report cites it
  only as a path (in the input-contract line and in §F cross-references), never quoting a figure
  from it. Every §A figure the main report accepted was **re-measured by me directly against the
  tree**, not taken from that report. So there is no old-version dependency to disclose, and the
  new-version read below adds a finding rather than revising one.

The main report's verdict therefore needs no reconstruction — only extension.

---

## Verified this round (all re-run by me at HEAD `297a21ea7`)

### Lead item (1) — `-c` vs `-co`: the control figure is corrupt under **both** units

```
$ grep -cE  '<AC-CSN-012 regex>' spec.md   →  45     (matching LINES)
$ grep -coE '<AC-CSN-012 regex>' spec.md   →  56     (OCCURRENCES)
$ grep -oE  '<AC-CSN-012 regex>' spec.md | wc -l  →  56   (cross-check of -co semantics)
```

`34` reproduces under neither. It does reproduce as **`grep -cE` on the REQ/AC arm alone** (34), and
that arm's `-co` value is 42 — so `34` is a *line* count of *one arm*, while the AC's own command is
a *line* count of *all four*. The author reports taking it with `-coE`; under `-coE` the whole-regex
value is 56 and the arm value is 42, so that account does not reproduce either. Whatever produced
`34`, it was not the command the AC prints.

**Judgment: this is worse than the staleness I graded in the main report, and it is a defect of the
control clause itself, not of a decoration.** The [HARD] clause exists to establish that the regex
is not dead before the two zeros are believed — it is the SPEC's own anti-vacuity mechanism at
exactly the place a vacuous green would be most costly. A run-phase executor who follows the clause,
runs the AC's command, gets 45, and finds 34 printed beside it meets an unexplained mismatch at the
moment the clause is telling them to trust the control. Both likely resolutions are bad: discount
the clause, or overwrite the number without diagnosing the unit. **D13 is upgraded: minor → major,
optional → blocking.**

The criterion itself still stands — the pass condition is written as "0이 아닐 것", and I verified
the control is live on **all four arms** (SPEC-ID 7 / REQ-AC 34 / hex 2 / date 4), which is stronger
than the clause asks for. So the contamination degrades the mechanism's credibility, not its
decidability.

### Lead item (2) — self-inclusion: cited command no longer reproduces cited output

```
$ grep -m1 '^status:' .moai/specs/SPEC-CODEX-*/spec.md | wc -l   →  10
  … 9 × "status: completed"
  … SPEC-CODEX-SKILL-NEUTRAL-001/spec.md:status: draft   ← this SPEC itself
```

`premise-remeasure.md` §4 prints this command with an output of 9 lines, all `completed`. **The
command as printed now returns 10, and the tenth is the SPEC doing the counting.** Reproduced
exactly as the lead reports; the mechanism is that the glob matches a working-tree directory that
did not exist when the output was captured.

The contrast the lead flags is real and I verified it:

```
$ git ls-tree -d --name-only origin/develop -- .moai/specs/ | grep -c CODEX   →  9
$ git ls-tree -d --name-only 297a21ea7     -- .moai/specs/ | grep -c CODEX   →  9
$ git ls-tree -d --name-only 48239c7dc     -- .moai/specs/ | grep -c CODEX   →  7
```

The ref-read ground is **structurally immune** to this failure — the SPEC's own directory is
untracked, so no ref can contain it — and it returns 9 from both the deployment line and this
SPEC's base. So the conclusion value survives on a ground I verified, while the working-tree ground
printed beside it has rotted. Two grounds for one fact; one decayed, one cannot.

Derived-claim sites: `premise-remeasure.md`:243, `progress.md`:24, `plan.md`:11. Only the first
prints the broken command; the other two carry the (correct) value.

### Lead item (3) — the hex arm's false red

```
$ printf 'defaced accede feedbed cabbage\n' | grep -oE '\b[0-9a-f]{7,40}\b'
defaced
feedbed
```

Reproduced (independently found in the main report as D14, with `defaced` / `deadbeef`).

**On whether this repository's own standard applies** — the lead is right that the standard exists
and right to ask, and it applies, but not at the magnitude that killed the dogfood guard. The
predecessor's ruling rejected a guard whose red would fire **routinely and structurally** (every
time the installed binary lagged the tree — a documented recurring state), so its red would have
meant "your binary is stale" more often than "the token came back", and the next reader would learn
to discount it. Here the red fires only if a word drawn from a small class enters one of two files;
`AGENTS.md` measures 0 today and its subject matter makes such a word unlikely. So the correct
disposition is not rejection of the arm but **making its red self-diagnosing**: the AC currently
uses `grep -cnE`, which prints a count and no match, so a false red is unattributable from the AC's
own output. Print the matches on the failure path and a false red is separable from a real one in
one step. **D14 stays optional; its required fix is tightened from "consider" to "record the match
text on the failure path."**

---

## D16 — the drift is systematic, and the discriminant is self-reference, not "command + output"

Severity: **major** · Class: **blocking** (one-line fixes; the discipline clause is the real
deliverable)

I tested the author's own diagnosis rather than accepting it, because the prescription follows from
the discriminant and a wrong discriminant produces a prescription that does not bind.

**The author's second ground — "corrupt sites are all command+output; clean sites are all structural
counts" — is true as an inclusion but false as a discriminant.** I re-measured every command+output
figure I could reach in the three frozen artifacts. Fifteen of them reproduce exactly:

| Figure | Cited | Re-measured |
|---|---|---|
| §A.1 ① skill dirs / `SKILL.md` | 34 / 34 | 34 / 34 |
| §A.1 ② tool-name skills | 14 | 14 |
| §A.1 ③ `SKILL.md` unit / skills / files / lines | 3 / 4 / 9 / 46 | 3 / 4 / 9 / 46 |
| §A.1 ④ agent `.md` / `.toml` | 11 / 11 | 11 / 11 |
| §A.4 HARD / SOFT / `moai/SKILL.md` | 6 / 40 / 19 | 6 / 40 / 19 |
| §A.6 both `AGENTS.md` / ceiling / free | 14229 / 24576 / 10347 | 14229 / 24576 / 10347 |
| §A.8 template tree lines / files / rules lines | 50 / 11 / 4 | 50 / 11 / 4 |
| §E.1 edit-target sum | 14 | 2+9+2+1 = 14 |

All fifteen are command+output citations, and none is corrupt. So "command+output" does not separate
the two populations.

**What does separate them: whether the measurement's SUBJECT includes the SPEC's own artifacts.**

| Corrupt figure | Subject | Why it moved |
|---|---|---|
| AC-012 control `34` | this SPEC's own `spec.md` | the SPEC's iter-2 edits grew the document it measures |
| `SPEC-CODEX-*` = 9 (working-tree glob) | a set this SPEC **joined by being created** | the glob acquired a tenth member: the counter |
| `progress.md` `12파일` | this SPEC's own edit-target count | the SPEC's iter-2 scope change moved its own figure |

Every clean figure has a subject **external** to the SPEC — the template tree, the launcher, the
byte guard, the agent set. None of those can move when the SPEC is edited. Every corrupt figure has
a subject that **includes the SPEC**. The split is **3/3 corrupt among self-subject figures, 0/15
among external-subject figures.**

That is systematic, and it is a defect of document form rather than of author attention, exactly as
the author argued — but one level sharper than they put it. The mechanism is not "measurements are
outside the edit's field of view"; it is that **a self-referential measurement is invalidated by the
very act of editing the document that cites it**, so the moment of writing is the moment of
breaking. No amount of care at authoring time survives the next edit, which is why all three got
through lint, MUST-PASS, and the author's own sweep.

### Does it contaminate the judgment layer?

**Yes, at exactly one site — and that site is the anti-vacuity mechanism.** I enumerated every
figure in `acceptance.md` by subject:

| AC | Figure | Subject | Status |
|---|---|---|---|
| AC-012 | control `34` | **own spec.md** | **corrupt** |
| AC-008 | pre-value 46 | template tree | clean (verified) |
| AC-011 | 4 lines / 2 files | template tree | clean (verified) |
| AC-005 | byte values | `AGENTS.md` + guard | clean (verified) |
| AC-010 | base SHA | git | clean, pinned |
| AC-009 | census 0 → 1 | template tree, run-phase | external by construction |

The judgment layer contains **exactly one** self-referential figure, and it is the corrupt one —
1/1, consistent with the discriminant. So the contamination reaches the judgment layer, but it lands
on a provenance parenthetical whose criterion I verified still decides correctly. **No AC's pass/fail
outcome is changed by any of the three corruptions.**

The honest summary for the lead's question: **the drift is systematic, it does reach the judgment
layer, and it is nonetheless run-phase-survivable** — because at the single judgment-layer site it
degrades a control's credibility rather than a criterion's decidability, and because the fix is one
line. What is **not** survivable-by-ignoring is the generative property: run-phase writes
`progress.md` and the §E.2 evidence section, both self-referential by construction, so run-phase will
manufacture more of these unless the discipline binds it.

### Is the author's proposed discipline sufficient?

The proposal — *"자기 산출물을 대상으로 잰 수치는 인용 시점 트리 상태를 함께 못박거나 아예 인용하지
말 것"* — is correct in both of its branches and **incomplete in two respects**.

Correct: for AC-012, **dropping the number is strictly better than restating 45**, because the value
is intrinsically non-stationary — the AC lives inside the document it measures, so any pinned value
re-arms the same decay on the next edit, and the pass condition ("0이 아닐 것") already carries the
whole criterion. For `SPEC-CODEX-*`, pinning the tree state is right, and the artifact **already
contains the stronger form** — the ref-read, which I verified returns 9 and is structurally immune
because the SPEC's own directory is untracked.

Incomplete in two respects:

1. **It addresses staleness but not the unit mismatch.** The `34` defect is two stacked failures, and
   only one is self-reference. Dropping the number removes the stale value; it does not stop the
   next author citing a `-co` value beside a `-c` criterion. The clause must add: **the control is
   run with the same command as the criterion it controls** — cite the command, not a number.
2. **Both branches are authoring-time rules, and authoring-time is precisely when these defects are
   invisible** (all three passed lint, MUST-PASS, and a self-sweep). The prescription needs a
   detection step, and the SPEC already owns the right one everywhere else: **a figure whose subject
   includes the SPEC's own artifacts is re-measured at judgment time and never carried from
   plan-phase** — the same pre/post discipline AC-006/008/011 already mandate for external figures.

I would also state the `SPEC-CODEX-*` branch in its stronger, already-proven form: **read a source
fact against a ref (`git ls-tree <ref>` / `git grep <pattern> <ref>`), not against the working
tree** — the ref is baked into the command, so the label cannot be lost. `premise-remeasure.md`'s
own preamble already states this discipline; the residual defect is only that the broken
working-tree command is still printed beside the ref-read that superseded it.

**Required fixes (D16):**

- (a) `acceptance.md` AC-CSN-012 [HARD] clause — drop `계획 단계 측정: 34`; state instead that the
  control is run with the AC's own command at judgment time and must be non-zero.
- (b) `premise-remeasure.md` §4 — replace the working-tree glob with the ref-read that already sits
  beneath it, or annotate the printed output as captured-before-this-SPEC-existed. Do not leave a
  printed command beside an output it no longer produces.
- (c) Add one clause to `plan.md` §G (an anti-pattern) and to `acceptance.md` §D.4 (a
  done-condition): **any figure whose subject includes this SPEC's own artifacts is re-measured at
  the moment it is cited, or not cited as a number at all.** This is the clause that binds
  run-phase, which is where the next instances will be generated.

---

## Revised scores

| Dimension | main report | revised | reason |
|---|---|---|---|
| Clarity | 0.85 | **0.85** | unchanged — D12 remains the deduction |
| Completeness | 0.90 | **0.85** | D16(b): the evidence artifact `§A` rests on prints a command that no longer reproduces its output. The §A values themselves all reproduce (15/15 verified), so the baseline is sound and the deduction is one band-step, not a collapse |
| Testability | 0.75 | **0.70** | D13 upgraded to blocking — the defect sits inside the SPEC's own anti-vacuity mechanism, and it is a unit mismatch on top of staleness. Three of the layer's defects (D11, D13, D14) now land on the single newest criterion |
| Traceability | 0.80 | **0.80** | unchanged — D11's partial mapping remains the deduction; D16 is booked under Completeness as an evidence-integrity defect rather than a REQ↔AC mapping defect |

Aggregate = mean(0.85, 0.85, 0.70, 0.80) = **0.800**.

The result is unchanged if D16 is booked under Traceability instead (0.85 / 0.90 / 0.70 / 0.75 =
0.800), which is the reason I am willing to report a figure sitting exactly on the threshold: it is
robust to the dimension assignment I was least certain about.

**Verdict: PASS-WITH-DEBT at 0.800** against the Tier M threshold of 0.80 — unchanged in kind from
the main report, with the debt list grown by D16 and D13 upgraded to blocking. No score regression
across iterations (0.75 → 0.800), so the STOP clause does not fire.

---

## Consolidated debt run-phase inherits (supersedes the main report's list)

**Blocking — close before the milestone that creates the exposure:**

1. **D11** — extend AC-CSN-012's file list to the two rules files now in scope
   (`skill-authoring.md`, `worktree-integration.md`) with their measured pre-values (`2` / `0`), or
   narrow REQ-CSN-013's mapping and declare the remainder. **Before M3 lands.**
2. **D13 (upgraded)** — drop `34` from AC-CSN-012's [HARD] control clause and state that the control
   runs the criterion's own command. **At entry** — it is inside the mechanism that prevents a
   vacuous green, and M2 is the milestone that relies on it.
3. **D16(c)** — add the self-referential-figure clause to `plan.md` §G and `acceptance.md` §D.4.
   **At entry**, because run-phase generates `progress.md` and §E.2 evidence, both self-referential
   by construction; without this clause the defect class continues into run-phase.
4. **D12** — one clause in `plan.md`:L13 (`측정 결과` → the measured part, failure mode marked as the
   inference M1 will observe). **At entry** — first framing a run-phase reader meets.
5. **D16(b)** — `premise-remeasure.md` §4's printed command. Evidence artifact, not a judgment
   artifact; close it whenever §A is next touched.

**Optional:** D14 (print match text on AC-CSN-012's failure path — recommended, near-zero cost),
D15 (:386's expected classification; annotate M1's REQ-CSN-009 as closing-by-record), the
`skill-authoring.md`:219 counter-note.

**None of the five blocking items prevents run-phase entry.** Four are one-line edits to plan/
acceptance text that should be taken at entry; one (D11) blocks M3's close rather than entry. No
implementation decision in this SPEC is wrong because of any of them — what they endanger is the
accuracy of claims *about* the implementation, which is precisely why item 3 (the binding discipline
clause) matters more than the three instance fixes.

## Probe hygiene

This round created no artifacts under version control. Two throwaway files were written to `/tmp`
(`hexprobe.txt`, `hex2.txt`) and nothing else; `git status --short` shows only the two untracked
directories this card owns. No tracked file was modified at any point in either round of this audit,
and the three judgment artifacts were verified byte-identical to the frozen set at re-read.
