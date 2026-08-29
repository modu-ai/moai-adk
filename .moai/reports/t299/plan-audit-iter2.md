# SPEC Review Report: SPEC-SYNC-SHA-SLOT-FORMAT-001 (v0.2.0)

Iteration: 2 (lead-directed; the Tier S ceiling is 1, and this pass runs on the lead's explicit
re-audit instruction after the iter-1 FAIL)
Verdict: **PASS**
Overall Score: **0.925** (iter-1: 0.80 — monotone improvement, no regression)
Audited in: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t299`, HEAD `a6bbbf82b` (unchanged from
iter-1: `git rev-parse --short HEAD` → `a6bbbf82b`; `git status --porcelain .moai/specs` → the SPEC's
own directory is still the only untracked path, so the corpus did not move between the two passes)

Scope: narrowed per instruction to closure of the iter-1 defects plus new defects introduced by the
edits. Figures already reproduced in iter-1 were re-derived only where an edit touched them.

Reasoning context ignored per M1 Context Isolation.

---

## Must-Pass Results (re-checked against v0.2.0)

- **[PASS] MP-1** — `grep -oE '^\*\*REQ-SSF-[0-9]+\*\*' spec.md` → REQ-SSF-001…008, unchanged,
  sequential, no gaps or duplicates.
- **[PASS] MP-2** — no requirement text changed except REQ-SSF-007's appended justification block,
  which is commentary below the `shall not` clause; the GEARS form of all eight is intact.
- **[PASS] MP-3** — frontmatter unchanged except `version: "0.2.0"`; all 12 canonical fields present
  with correct types.
- **[N/A] MP-4** — single-language scope, unchanged.
- **[PASS] MP-5** — no new SPEC reference introduced except `SPEC-CLIFIX-CONCURRENCY-001` and
  `SPEC-V3R6-BASH-RISK-GOVERNANCE-001` in the new mx table; both resolve and both are `completed`
  (`grep -m1 -H '^status:'`). None retired/superseded/archived.
- **[PASS] MP-6** — `grep -c 'syscall'` → 0 on all three artifacts.
- **[PASS] MP-7** — `grep -c 'NEEDS CLARIFICATION' plan.md` → 0.

---

## Iter-1 defect disposition

### D1 (critical) — **CLOSED**

§B.6 now reads **"Predicted: 5 total findings, 0 non-advisory"** (`spec.md:229`), with a new §B.5.1
that names the five per file and line and lists the seven exempt placeholders separately. Re-derived:

```
$ python3 .moai/reports/t299/grammar_check.py
total 346 | SHA 334 | PLACEHOLDER 7 | FLAGGED 5
```

The five in the SPEC's table match the classifier row-for-row —
`SPEC-V3R6-SESSION-LEGACY-COVERAGE-001:47` (`null`) and `:259` (`<pending>`),
`SPEC-V3R6-SPEC-ID-VALIDATION-001:103` (`TBD`), `SPEC-V3R6-SPEC-LINT-CLEANUP-001:45` (`null`),
`SPEC-V3R6-TEST-REFACTOR-001:149` (`pending`) — and the seven named exempt SPECs match the
classifier's PLACEHOLDER list exactly, so the seven are **accounted for, not dropped**: `plan.md` §F
M4 now requires **two distinct inventories** in `progress.md` §E.2 (the five findings; the seven
still-owed slots, "useful to the lead for scheduling, never a work item for this lane").

AC-SSF-006 (`acceptance.md:128-158`) matches on every element:
- expected values **5** total / **0** non-advisory, plus "contributes nothing to the `--strict` exit
  status";
- **"Fails when"** corrected — a wrong-attachment implementation yields **5** non-advisory, not 12;
- mutation corrected — promoting to `error` moves the non-advisory count **0 → 5**;
- and it goes beyond the fix I asked for: the five are now **asserted individually by file and
  line**, with the stated reason that "a count-only assertion would stay green if the rule found five
  *different* slots." That closes a hole iter-1 did not raise.

### D2 (major) — **CLOSED**

The exempt pair is gone from every location, not only the paragraph flagged.

```
$ grep -n 'non-advisory' spec.md plan.md acceptance.md
spec.md:211, 229, 385 | plan.md:85, 187, 188 | acceptance.md:34, 133, 155, 157, 232
```

Every one of those sites carries `5 / 0`. §D.3 (`spec.md:385-398`) now rests on "0 non-advisory
findings, 5 advisory, all five on closed history", and adds an explicit retraction: "it does not rest
on the two `implemented` SPECs. Both hold `pending-backfill-sync` and are exempt under REQ-SSF-005."
The only surviving mention of the old pair is `spec.md:211`, which is the retraction narrative itself.

### D3 (major) — **CLOSED**, both halves

*Mechanical decision.* AC-SSF-010 (`acceptance.md:203-224`) decides REQ-SSF-007 with a binary
command, and the command produces a binary answer today:

```
$ grep -A6 'var eraDemotableCodes' internal/spec/lint.go
var eraDemotableCodes = map[string]bool{
	"MissingExclusions":  true,
	"FrontmatterInvalid": true,
}
```

Two entries, no `SyncSHASlotFormat`. The stated mutation (add the entry → grep shows a third) turns
it red.

*Conditional justification.* REQ-SSF-007 (`spec.md:266-297`) now carries a `[HARD]` block stating the
justification "holds **only while the rule's severity is `warning` at the `Finding` level**", with a
two-row table separating `Report.HasErrors` (entry stays inert, requirement holds) from
`Finding.Severity` (premise void), and a named lane response: "**stop and report to the lead** rather
than adding the map entry on its own authority", with the reason — the choice is a policy decision the
operator owns.

*Honesty check — the SPEC does not claim t357 was verified.* `spec.md:282-284` reads: "`plan.md` §C.4
records, from the lead's dispatch and **not** independently verified (no t357 SPEC directory exists in
this tree to read)". Confirmed against the tree: `ls .moai/specs | grep -iE 'lint|strict'` returns no
t357 SPEC. The contingency is recorded, not resolved — which is the correct disposition for a claim
that cannot be verified here.

### D4 (major) — **CLOSED**, all four figures re-derived

| SPEC claim | Re-derived | Command |
|---|---|---|
| 85 mx slots | **85** | `grep -h '^mx_commit_sha:' .moai/specs/*/progress.md \| wc -l` |
| 9 non-SHA | **9** | same `\| sed … \| grep -vcE '^"?[0-9a-fA-F]{7,40}"?…'` |
| 6 already repaired (`null` ×3, `(this commit)` ×3) | **exact** | same `\| sort \| uniq -c` → `3 null`, `3 (this commit)`, plus the three declarations |
| 3 deliberate declarations, owners all `completed` | **exact** — `SPEC-CLIFIX-CONCURRENCY-001`, `SPEC-V3R6-BASH-RISK-GOVERNANCE-001`, `SPEC-V3R6-LIFECYCLE-REDESIGN-001`, all `status: completed` | `grep -l '^mx_commit_sha:…'` per value + `grep -m1 -H '^status:'` on each owner |

`spec.md` §D.2 measures the blast radius rather than asserting it, and reaches a conclusion iter-1 did
not: the exposure is **prospective, not live** — all three owners are already `completed`, so the
close path has no reason to run against them, and the Mx phase is itself retired by the third SPEC in
the list, so the class shrinks. §E carries the acceptance as an explicit Out-of-Scope bullet naming
all three values and routing a future durable "not applicable" need to its own card. An Out-of-Scope
bullet on a measured basis was the stated sufficient condition; this exceeds it.

### D5 (minor) — **CLOSED**

§B.3's command now carries `| grep -v 'SPEC-SYNC-SHA-SLOT-FORMAT-001'`, and the SPEC explains why the
exclusion is load-bearing. Re-derived with the command as now written:

```
  29 pending-backfill
  24 pending-backfill-sync
$ … | sort -u | wc -l  → 28
```

29 / 24 / 28 reproduce exactly as recorded.

### D6 (minor) — **CLOSED**

§B.2 row-4 commentary now reads "105 of the 346 values … and 99 of the 334 conforming ones", names
the six-value difference, and adds the deriving command. Re-derived:

```
$ grep -h '^sync_commit_sha:' .moai/specs/*/progress.md | sed 's/^sync_commit_sha:[[:space:]]*//' \
    | grep -E '^"?[0-9a-fA-F]{7,40}"?([[:space:]]|$)' | grep -c '#'   → 99
```

### D7, D8, D9 — **OPEN, unchanged, non-blocking as before**

- **D7** (implementation identifiers in REQ-SSF-002/007/008) — recorded by the author, no change.
  Correct call for a SPEC whose subject is an internal contract.
- **D8** (AC-SSF-008 orphan, partly non-mechanical decider) — still the only orphan AC. AC-SSF-010
  maps to REQ-SSF-007, so it does not add a second one. SHOULD-level; unchanged.
- **D9** (§E.2/§E.3/§E.4 naming drift) — pre-existing repository-wide, correctly left recorded.

### Regression check — **none**

`grep -n 'twelve\|\b12\b'` across all three artifacts returns 18 sites; every one refers to the twelve
**values** and every one sits in a context that distinguishes values from findings — including two new
guard rails: `acceptance.md:143-145` ("A run reporting 12 findings has broken AC-SSF-003, not exceeded
expectations") and `plan.md:210` (a new §G anti-pattern, "Reading '12 non-SHA values' as '12
findings'"). No stale `2 non-advisory` survives anywhere. `spec.md` §B.2's four counts
(346/334/12/105) were not touched by the edits and were verified in iter-1.

---

## New defects introduced by the remediation

**None of blocking class.** Three observations, all minor and optional.

**N1. The strongest available justification for the lint rule is still unnamed.** — `spec.md` §A —
**Severity: minor** — **Class: optional**.

§A justifies the lint rule as a backstop for "a value that is neither a SHA nor a placeholder, so no
close will ever be triggered to repair it". That is true for a terminal-status owner, but it is the
weaker of two available arguments, and it leaves a reviewer able to ask "then why have the rule at
all?" The stronger argument is mechanical and sits one function away:

```
$ grep -n 'func resolveRecentSpecCommitSHA' -A 8 internal/spec/closer.go
434: cmd := exec.Command("git", "log", "-1", "--format=%H", "--grep="+specID)
438: if err != nil { return "" }
$ sed -n '324,329p' internal/spec/closer.go
      if needsSHABackfill(state.SyncCommitSHA) {
              resolved := resolveRecentSpecCommitSHA(baseDir, specID)
              if resolved != "" { … }        ← empty resolution: no backfill, prose survives the close
      }
```

A close that runs but whose SHA resolution returns empty completes anyway, leaving the prose in place
while flipping the owner to `completed` — which permanently shelters the finding as advisory. That is
the mechanism by which the existing five froze, and it is a class the closer demonstrably *cannot*
help even when it does run. Naming it would convert §A's justification from "no close will run" (a
circumstance) into "the close path can complete without repairing" (a defect). Optional: the rule's
purpose is already sound without it.

**N2. §A says §B.6 "measures" a class §B.6 itself is careful to call a derivation.** — `spec.md` §A
("§B.6 measures that class at five slots") vs `spec.md:234` ("a prediction derived from code plus a
classifier run … not a measurement of a rule that does not yet exist") — **Severity: minor** —
**Class: optional**. A one-word drift against the document's own discipline; `derives` is the word.

**N3. §B.5.1's closing sentence conflates findings with owning SPECs.** — `spec.md:206-207` ("All five
owning `spec.md` files carry `status: completed` … over the four SPECs") — **Severity: minor** —
**Class: optional**. Five findings live in four SPECs (`SPEC-V3R6-SESSION-LEGACY-COVERAGE-001` holds
two, as §B.5 says). Both numbers are individually right; the sentence puts them in apposition.

---

## Reviewer judgement — does the lint rule still have a purpose? (lead's check 7)

**The new statement holds mechanically, and the rule's purpose survives it.**

The claim tests true against §D.1: `pending-backfill-sync` matches
`^pending-backfill(-[A-Za-z0-9-]+)?$`, so REQ-SSF-005 requires silence, and the classifier puts
`SPEC-BACKLOG-LOCK-BUDGET-001` in the PLACEHOLDER list. On the closer side
`!isCommitSHAToken("pending-backfill-sync")` → `true` → backfill. The two guards do split the way the
new §A table says they do.

The rule's purpose is sound, on three grounds the SPEC either states or has to hand:

1. **Timing.** The lint fires at any point in a SPEC's life; the closer fires only at close. On an
   *active* SPEC, prose in the slot produces a **non-advisory** finding that reddens `--strict` —
   the card's only forward-looking teeth.
2. **The closer can complete without repairing** (N1). The five existing findings are the evidence
   that this path is real, not hypothetical.
3. **The classes are genuinely different questions**, and the new §A table makes that legible in a way
   v0.1.0 did not.

The honest reading — which the SPEC now states rather than hides — is that on day one the rule
enforces **nothing**: all five findings are advisory and it contributes zero to the strict exit
status. Its day-one value is an inventory; its enforcement value is prospective. A lead may reasonably
ask whether that justifies a new rule, and the SPEC has given them the numbers to answer it rather
than a rationale that assumes the answer. That is the correct disposition for a plan-phase artifact,
and it is why this is not a defect.

One consequence worth carrying to the lead: **N1 plus the t357 contingency interact.** If t357 M2
promotes at `Finding.Severity`, the five advisory findings become five hard errors on closed history —
and REQ-SSF-007's block already routes that to "stop and report". The SPEC handles it; I note only
that the same five slots are the pivot for both the "does this rule do anything?" question and the
"does t357 break it?" question, so a lead reviewing one should review both together.

---

## Category Scores

| Dimension | Score | Δ from iter-1 | Evidence |
|---|---|---|---|
| Clarity | 0.95 | +0.10 | The §A two-guard table (`spec.md:52-70`) resolves the card's most confusable point — which guard handles the motivating case — before a reader can misread it. Deductions: N2, N3 wording only. |
| Completeness | 0.90 | +0.15 | mx blast radius measured and accepted (D4); the seven exempt placeholders given their own inventory with its own decision owner (`plan.md` §F M4); REQ-SSF-007's condition and contingency named. Deduction: N1. |
| Testability | 0.95 | +0.15 | AC-SSF-006 asserts five findings **per file and line**, not a count, with the reason stated; AC-SSF-010 is binary and verified against live output; every AC still carries a named falsifying input and a mutation. |
| Traceability | 0.90 | +0.10 | REQ-SSF-007 → AC-SSF-010 closes the only uncovered requirement. Deduction: AC-SSF-008 remains an orphan (D8). |

Arithmetic mean **0.925**, above the Tier S PASS threshold (0.75), monotone from 0.80.

---

## Recommendation

**PASS.** All six defects the lead scoped to this pass are closed, each against a command whose output
I ran in this tree rather than against the author's account of it. Two remediations exceed what iter-1
asked for — AC-SSF-006's per-file/line assertion and the two-inventory split in M4 — and neither
introduced a contradiction. The three residual observations (N1-N3) are optional and would be cheap to
fold into any later edit; none of them should hold up run-phase entry, and per M6 they must not be
routed into a revision cycle on their own.

Two things for the lead to carry forward rather than for the author to fix:

- **AC-SSF-006 is now a reporting instrument with real teeth.** Its expected `5 / 0` is a derivation
  plus a classifier run, not a measurement of a rule that exists. If the run phase measures anything
  else, that is information about the derivation — the SPEC says so in three places and the DoD
  repeats it. Hold the lane to reporting it, not to hitting it.
- **The five slots are the pivot for both open questions** — whether the rule enforces anything on day
  one (it does not; the value is prospective) and whether t357 M2 voids REQ-SSF-007's premise. Review
  those together when t357's promotion layer becomes readable.
