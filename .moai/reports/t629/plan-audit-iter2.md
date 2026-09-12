# SPEC Review Report: SPEC-REVIEW-SECRET-SCAN-REFS-001
Iteration: 2/2 (Tier M ceiling 2 — final iteration)
Verdict: FAIL
Overall Score: 0.79 harmonic (arithmetic mean 0.80; Tier M PASS threshold 0.80)

card: t629 · auditor: plan-auditor (independent, read-only) · date: 2026-09-10
Reasoning context ignored per M1 Context Isolation — the lead's brief and the author's disclosed gap
were used only as a list of places to look; every reading below was re-measured in this run.

The verdict rests on two grounds, each sufficient on its own: D1 is only partially resolved (a prior
blocking defect not fully resolved is an automatic FAIL under the retry contract), and the revision
introduced blocking defects (D13, D15) in the criteria it rewrote. The harmonic aggregate is 0.79.
No score regression against iteration 1 (0.79 → 0.79 harmonic), so no STOP signal.

## Baseline-attribution

Tree measured: worktree `.claude/worktrees/t629`, branch `WT-secret-scan-refs`.

```
$ git rev-parse --show-toplevel   → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t629
$ git rev-parse --short HEAD      → d31a6973b
$ git status --porcelain          → (empty)
$ git branch --show-current       → WT-secret-scan-refs
```

Judging build for `moai spec lint`: the installed `moai`; the MCP server banner in this session names
`v3.2.0-rc.5`, commit `84fa4ece4`. `git merge-base --is-ancestor 84fa4ece4 HEAD` → exit 0, so the
build is behind HEAD; `git diff --stat 84fa4ece4 HEAD -- internal/spec internal/cli/spec_lint.go`
printed nothing. Gap: the installed binary's own version string was not read back in this run
(`moai version` output was cut by the terminal box), and lint may depend on packages outside those
two paths.

Artifacts read in full: `spec.md` (421 lines), `plan.md` (190), `acceptance.md` (483),
`progress.md` (90), iteration 1 report (278), the secret-scan section of the distributed
`review.md`; iteration 1's `acceptance.md` §D.1 and §D.13-§D.14 at `496fe6153` for the regression
check. Tier M; no `design.md` or `research.md` required.

## Re-measurements this run

| Reading | Command | Observed |
|---|---|---|
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | exit 0; `0 error(s), 0 warning(s)`; one INFO `OwnershipTransitionUnmeasured` on commit `78e29987f` |
| fix commit file set | `git show --stat d31a6973b` | 4 files, all under the SPEC dir; `242 insertions(+), 70 deletions(-)` |
| card commits | `git log --oneline feeecc980..HEAD` | 7 commits, linear, all `docs(t629)` |
| copies unchanged | `git diff --stat feeecc980 HEAD -- <both review.md copies>` | empty, exit 0 |
| copies identical | `diff -q LOC TPL` | exit 0 |
| `same coverage` baseline | `/usr/bin/grep -ci 'same coverage' LOC TPL` | `1`, `1` |
| AC-004 | `git diff feeecc980 HEAD --output=<scratch>`; `wc -l`; added-line credential regex `grep -cE` | `1693` lines; count `0`, exit 1 |
| regex over SPEC + iter-1 report | `/usr/bin/grep -cE -- '<REGEX>' <4 SPEC files> plan-audit-iter1.md` | `0` each, exit 1 |
| positive control | same grep over a PEM-header line assembled from two fragments in the scratchpad | `1`, exit 0 |
| REQ numbering | `grep -oE '^- \*\*REQ-[0-9]{3}' spec.md` | REQ-001 … REQ-013, count 13, no gap, no duplicate |
| AC numbering | `grep -oE '^\| AC-[0-9]{3}' acceptance.md`; `^## §D\.` headings | AC-001 … AC-016; §D.1-§D.16 carry AC-001-AC-016 in order; §D.17 candidate, §D.18 edge cases, §D.19 DoD |
| `§D.N` references | `grep -noE '§D\.[0-9]+'` over spec/plan/progress | spec §D.9, §D.5; plan §D.17, §D.1, §D.13, §D.14, §D.17, §D.18; progress §D.1, §D.16 — all point at the right section |
| gate substrings in `PROG` | `grep -nE 'Gate evidence\|Pinned procedure\|Gate cell\|Lead approval\|verdict:' progress.md` | no lines, exit 1 |
| decision line | `grep -c '^\*\*Decision:\*\* Option 2' spec.md` | `1` |
| cross-SPEC refs | `grep -ohE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` over the 4 files | only the SPEC's own ID (15×) |
| syscall | `grep -c syscall` over the 4 files | `0` each |
| clarification markers | `grep -rn 'NEEDS CLARIFICATION' plan.md`; `ls research.md` | exit 1; `research.md` absent |
| strict leak test | `grep -nE 'func Test\|MOAI_TEMPLATE_LEAK_STRICT' internal_content_leak_test.go`; `grep -rn` over `.github/workflows/` | `func TestTemplateNoInternalContentLeak` at line 1535; strict switch at line 1545; `template-neutrality-check.yaml:85-88` runs it with `MOAI_TEMPLATE_LEAK_STRICT: '1'` |
| AC-016 mechanics + D14 mutant | mock `progress.md` block and mock mutant section in the scratchpad; the AC-016 `sed` extraction, `sed -n 's/^Scan command: //p'`, `sed -n 's/^Missing-tip handling: //p'`, `grep -cF -f` | extraction printed heading + 4 body lines, kept `#### sub`, stopped before `### Lead approval…`; mutant section (tips recorded **after** the scan): scan-command count `1` exit 0, missing-tip count `1` exit 0; control for the pinned timing clause `0` exit 1 |

No full-history scan was run, no matched commit was opened, no listed allowlist value was written
(whole or fragmented), no `go test` and no AC-005 leak run was taken, nothing was committed.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** — REQ-001..REQ-013 sequential, 3-digit padding, no
  duplicates (`spec.md:100-162`).
- [PASS] **MP-2 EARS/GEARS format** — judged on the requirement layer (`spec.md` §2) only.
  Ubiquitous: REQ-001, 003, 004 (`spec.md:118`, now Ubiquitous), 006, 011; When: REQ-002, 009, 010,
  013 (`spec.md:158`, "**When** either copy … is edited, … shall be …"); Unwanted `shall not`:
  REQ-005, 007, 008, 012. ACs were not graded here.
- [PASS] **MP-3 YAML frontmatter** — `spec.md:2-14`: all 12 canonical fields, `version: "0.2.1"`
  quoted, ISO dates, `priority: P1`, `lifecycle: spec-anchored`, comma-separated `tags`, plus
  `tier: M`. No rejected alias.
- [PASS] **MP-4 language neutrality** — no language-specific tool named; REQ-006 and AC-005 forbid
  naming a language in the distributed section.
- [PASS] **MP-5 D7 cross-SPEC** — only the SPEC's own ID referenced; no BLOCKING finding.
- [PASS] **MP-6 D8 cross-platform** — `syscall` count 0 in all four files; auto-pass.
- [PASS] **MP-7 clarification gate** — no `[NEEDS CLARIFICATION` in `plan.md`; `research.md` absent.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | D2 and D4 fixed (`acceptance.md:102-107`, `spec.md:264-268`); AC-013/AC-014 leave "reported findings" undefined in granularity, which matters now that the negatives share the listed line's hunk (D13); AC-001's "exactly when" leaves its construction checks with no consequence (D15) |
| Completeness | 0.95 | 1.0 minus minor | All sections present; HISTORY carries the 0.2.1 entry (`spec.md:37-46`) and marks superseded notes (`spec.md:26-31`); REQ-013 and AC-016 added with matrix row (`acceptance.md:57`) and DoD entry (`acceptance.md:473`) |
| Testability | 0.75 | 0.75 | AC-002, AC-009, AC-014 construction checks tightened; but AC-013 can fail a correct procedure under a hunk-granular findings reading (D13), and AC-001 has an un-consequenced construction precondition (D15) |
| Traceability | 0.75 | 0.75 | REQ-013 → AC-016 exists, but AC-016 verifies two of the three pinned components the SPEC names; the tip-recording component has no check (D14, measured mutant) |

Arithmetic mean 0.80; harmonic mean 0.79 (reported as the aggregate, per the skeptical-evaluation
stance). Tier M threshold 0.80.

## Resolution of iteration 1 findings

| ID | Class (iter 1) | Status | Evidence on current text |
|---|---|---|---|
| D1 GATE-WORDING-LINK | blocking | **Partially resolved** | REQ-013 `spec.md:158-162`; §3.4 re-gate rule `spec.md:276-278`; AC-016 `acceptance.md:408-446`; plan M3 `plan.md:116-122`; pin format `plan.md:38-46`. The iteration-1 mutant (word a procedure that drops missing-tip handling) now fails AC-016 (a). Residual: the third pinned component — "how and when the tip set is recorded" (`spec.md:274`, `plan.md:38-39`) — has no pinned line and no check (D14, measured). See the G-selection analysis below |
| D2 AC001-PREDICATE-SPLIT | blocking | Resolved | `acceptance.md:102-107`: fail conditions are B ≠ 0 bytes, H = 0, S = 0, T = 0, N ≠ 0 bytes — the exact complement of the table at `acceptance.md:94-100`; "one review late" explicitly fails. The wording "fails exactly when" introduced D15 |
| D3 AC014-GRANULARITY | blocking | Resolved as prescribed | `acceptance.md:370-393`: same commit and same hunk, adjacent lines; `MIXCELL` added; construction check (`^@@` count `1`, each of four labels `1`) with gap-not-pass rule; whole-commit, whole-hunk, and whole-line mutants listed. The same-hunk placement introduced D13 |
| D4 PREDICATE-ATTRIBUTION | blocking | Resolved | `spec.md:264-268` separates the operator's measurements and stop rule from the lane's predicates; `spec.md:317-318` marks cell ②'s detection requirement lane-authored |
| D5 CELL2-EXIT-AND-DETECTION | blocking | Resolved | (a) `spec.md:311` "the scan carrying the final result exits 0", `spec.md:314-316` non-zero final is untrustworthy, `acceptance.md:264-267`; (b) `acceptance.md:261-263` pins `/usr/bin/grep -cF '<G1>' SP/g2.err SP/g2.txt` plus the fallback file, summed |
| D6 AC002-PHRASE-ONLY | optional | Resolved | `acceptance.md:123-127`; baseline `1`, `1` re-measured this run |
| D7 APPROVAL-ORDERING | optional | Resolved (commit order) | `plan.md:47-51`, `spec.md:323-324`, `acceptance.md:275-277` and `293-305` (`P` newest approval, `E3` newest cell-3 evidence, `P ≠ E3`, `is-ancestor P E3`). Residual: the graph witnesses approval-commit before evidence-commit, not approval before the run itself (D18) |
| D8 AC005-CI-TIER | optional | Resolved | `spec.md:372-375` full-length digest; `acceptance.md:167-184` runs the strict tier with PASS ≥ 1 / FAIL 0 and explains `go` and `r`. Test and strict switch confirmed at `internal_content_leak_test.go:1535,1545`; CI at `template-neutrality-check.yaml:85-88` |
| D9 REQ-SHAPE | optional | Resolved (REQ-004); REQ-011 not split | `spec.md:118` Ubiquitous; REQ-011 `spec.md:148-151` unchanged — accepted, the split was optional |
| D10 STALE-HISTORY | optional | Resolved | `spec.md:26-28` "(0.1.0) … Superseded by 0.2.0."; `spec.md:29-31` "(superseded by 0.2.0)" |
| D11 ABSORB-RANGE | optional | Resolved for the listed ACs; residual inconsistency | Anchor `K` `acceptance.md:23-29`; ranges end at `K` in AC-003 (`:133`), AC-007 (`:216`), AC-010 (`:294-296`), AC-011 (`:320-322`), AC-012 (`:336-343`), AC-016 (`:415`). AC-004 still ends at `HEAD` (`:142`) while `plan.md:130-133` says AC-004 runs at `K` (D16) |
| D12 CELL3-STATE-LOCATION | optional | Resolved | `spec.md:326-328`, `plan.md:65-67`, `acceptance.md:278` (`SP/g3-tips.txt`) |

### Focus checks on the D1 resolution

**Can a mutant word a procedure different from the gated one and pass?** Yes, in two ways.
(1) Tip-recording timing: the pin format (`plan.md:41-43`) fixes only a `Scan command: ` line and a
`Missing-tip handling: ` line; AC-016 extracts only those two. A mock run in this session — a section
that prescribes recording tips **after** the scan while the pin says **before** — gave scan-command
count `1`, missing-tip count `1`, and `0` for the pinned timing clause. Gate cells ①-③ cannot tell
the two timings apart either (no commit lands during a fixture scan), and AC-001 on the worded
procedure passes the same way. `plan.md:144-146` names this as a real coverage risk. (2) Containment,
not prescription: `grep -cF` ≥ 1 shows the pinned lines appear, not that they are what the section
prescribes; a section quoting them and prescribing something else in addition passes AC-016. AC-001
(cell N) and AC-006 close the full-rescan variant of (2), so (2) is recorded as optional (D19).

**Does the "newest gate round" G selection hold?** Yes. `-S '### Gate evidence'` over
`feeecc980..R~1 -- PROG` lists every commit that changes the substring's count. Round 1 adds (0→1),
the M3 move-out removes (1→0), round 2 adds (0→1); the newest line is round 2's add, and the
`### Gate evidence` count `1` at `R~1` (`acceptance.md:420, 430-431`) rejects a newest line that is a
removal. If a lane folded move-out and the new round into one commit, the count would not change,
`G` would stay round 1, and `cmp` would fail on the changed pin — a conservative false FAIL, not a
false pass. A doubled `### Pinned procedure` block fails the "exactly one non-empty line" rule
(`acceptance.md:433-434`). Plan-time `PROG` holds none of the pickaxe substrings (grep exit 1), so no
plan-phase commit can be selected.

**Does the M3 move-out keep AC-011, AC-016, AC-010 sound?** AC-011 keeps the **oldest** `G`
(round 1), which is still an ancestor of `R`; its `verdict: trustworthy` count of `3` at `R~1` reads
the standing round, because round 1's lines were moved to `.moai/reports/t629/`, outside the
`-- PROG` path. AC-016 uses the newest `G` as above. AC-010 uses the newest `P` and newest `E3`: after
move-out, round 2's fresh approval and round 2's cell-3 evidence are the newest adds, so the ordering
check reads the standing round; a round 2 with no fresh approval makes `P` the move-out commit and
`is-ancestor P E3` still holds only if evidence follows — see D17 for the missing count guard.
AC-012 is unaffected (oldest `U`). All three readings stay sound on a linear card branch; the one
weakness is AC-010's lack of a count-at-`K` guard (D17, optional).

**Author-disclosed gap (AC-004 ends at `HEAD`).** Classified as D16. After the absorb,
`feeecc980..HEAD` includes develop's own added lines, so the failure direction is a false FAIL (a
develop fixture line matching the regex), not a false pass. But `plan.md:130-133` lists AC-004 among
the closure checks run at `K`, which contradicts `acceptance.md:26` (anchor list omits AC-004) and
`acceptance.md:142` (`HEAD`). Internal inconsistency → blocking, one-token fix.

## Defects Found

Iteration 1 carry-overs are listed by resolution above. New or residual defects:

D13. AC013-014-FINDINGS-GRANULARITY — `acceptance.md:359-363`, `acceptance.md:370-386` — AC-013
requires `grep -c 'LISTCELL' SP/al-findings.txt` = `0`; AC-014 now places `NEARCELL`, `OTHERCELL`, and
`MIXCELL` in the same hunk as `LISTCELL`. Neither criterion, nor REQ-011 (`spec.md:148-151`), defines
what `SP/al-findings.txt` holds: one line per unsuppressed match, or the scan's patch output after
suppression. The document's current reporting shape is patch output (the distributed section
prescribes `git log -p … -G …`). Under a patch- or hunk-granular reading, the reported hunk for
`NEARCELL` carries the `LISTCELL` line as a neighbouring added line, so the `LISTCELL` count is ≥ 1
and a correct exact-value procedure fails AC-013. Under a match-line reading both pass. Two readings
of a DoD criterion pair, one of which is unsatisfiable. The hunk-content consequence is reasoned from
the unified-diff format, not measured with a fixture in this run. — Severity: major — Class:
blocking — Required fix: define the findings file in AC-013 (and reference it in AC-014) as one line
per reported regex match, the matched line only; or count every label only over lines that match
`REGEX` (e.g. `/usr/bin/grep -E -- 'REGEX' SP/al-findings.txt > SP/al-match-lines.txt`, then
`grep -c` per label on that file). State in AC-013 that context lines printed with a finding are not
findings.

D14. AC016-TIP-RECORDING-UNCHECKED — `acceptance.md:410-436`, `plan.md:38-43`, `spec.md:158-162`,
`spec.md:273-275` — REQ-013 requires the edited copies to prescribe "the procedure pinned", and the
SPEC names three pinned components: how and when the tip set is recorded, the scan command, and
missing-tip handling. The pin format and AC-016 cover only the last two. Measured mutant (mock files
in the scratchpad): a section recording tips after the scan while the pin records them before gives
scan-command count `1`, missing-tip count `1`, pinned-timing clause `0` — AC-016 passes. No gate cell
and no other AC distinguishes the timing, and `plan.md:144-146` identifies record-after-scan as a
coverage hole. This is the D1 defect surviving for one component. — Severity: major — Class:
blocking — Required fix: add a third pinned line `Tip recording: ` to the pin format (`plan.md` §C
item 3, `acceptance.md:410-413`), extract it with `sed -n 's/^Tip recording: //p'` into
`SP/pin-tips.txt` with the same one-non-empty-line rule, and require
`/usr/bin/grep -cF -f SP/pin-tips.txt SP/section.txt` ≥ 1; add the record-after-scan mutant to the
mutant pairing.

D15. AC001-CONSTRUCTION-CONSEQUENCE — `acceptance.md:77-78`, `82-83`, `89-90`, `102-107` — steps 3, 5,
and 8 record `git merge-base --is-ancestor` exit 1 and forbid any merge of `side` or `side2`, and each
step's scan exit code is read, but the fail conditions now say AC-001 "fails exactly when the
required-outcome table does not hold", which excludes those readings from any consequence. Mutant: if
`side` and `side2` are merged by construction error, a HEAD-anchored procedure reports `SIDECELL` in
R2 and `LATECELL` in R3 and passes the table. The siblings treat the same situation as a gap
(AC-009 `:264-268`, AC-014 `:385-388`) or as part of the predicate (AC-008 `:237-238`). Iteration 1's
fail-condition wording was not exhaustive, so this is introduced by the D2 rewrite. — Severity:
minor — Class: blocking — Required fix: add one sentence after `acceptance.md:107`: an
`is-ancestor` reading other than exit 1, or a non-zero exit from any of R1-R4, means the sequence was
not constructed as specified; it is rebuilt, and that reading is a gap, not a pass.

D16. AC004-ANCHOR-INCONSISTENCY — `plan.md:130-133` vs `acceptance.md:26`, `acceptance.md:141-145` —
plan M5 says AC-004 runs at `K`; the acceptance anchor list omits AC-004 and its command diffs
`feeecc980 HEAD`. After the develop absorb, the diff also carries develop's added lines, so the
"Given all of the card's commits" premise no longer holds (failure direction: false FAIL). Disclosed
by the author. — Severity: minor — Class: blocking (internal consistency between two artifacts) —
Required fix: change the AC-004 command to `git -C <worktree> diff feeecc980 K --output=…` and add
AC-004 to the anchor list at `acceptance.md:26`; or remove AC-004 from `plan.md:131-133`'s "run at
`K`" sentence.

D17. AC010-NEWEST-WITHOUT-COUNT-GUARD — `acceptance.md:293-305` — `P` and `E3` are the newest
pickaxe hits, but, unlike AC-016's count-at-`R~1` guard, AC-010 does not require the
`#### Gate cell 3` and `### Lead approval for gate cell 3` substring counts at `K` to be `1`. A later
commit before `K` that changes either count (a removal, or an incidental verbatim mention in `PROG`)
becomes the selected commit; if it changes the `E3` substring after an approval-and-evidence
same-commit violation, the ordering check passes. Exposure is low today (plan-time `PROG` holds
neither substring). — Severity: minor — Class: optional — Suggested fix: add
`git show K:<PROG path> > SP/prog-k.txt`, then require `/usr/bin/grep -c '^#### Gate cell 3$'` and
`/usr/bin/grep -c '^### Lead approval for gate cell 3$'` each `1`.

D18. ORDERING-WITNESS-LIMITS — `acceptance.md:437-440`, `spec.md:273-275`, `spec.md:323-324` — two
before-claims rest on nothing the commit graph witnesses: the pin written before the first gate
command (REQ-013), and the approval obtained before cell ③'s first run (only approval-commit before
evidence-commit is checked). AC-016's mutant note "a pin that exists only after the gate fails the
2-line block check at G" holds only if "after the gate" means after the evidence commit; a pin
written after the gate commands and committed with the evidence passes. — Severity: minor — Class:
optional (no commit-based check can witness an unrecorded gate attempt) — Suggested fix: commit the
pin alone before the first gate evidence commit and add `is-ancestor <pin commit> G`; or reword the
mutant note to "a pin absent at `G`". For the approval, cite the lead message's timestamp next to
the `SP/g3-first-load-before.txt` `uptime` wall-clock time.

D19. AC016-CONTAINMENT-AND-SCOPE — `acceptance.md:413-436` — (a) `grep -cF` ≥ 1 proves the pinned
lines appear in `SECTION`, not that they are what it prescribes; (b) `SECTION` is read from the
working-tree `TPL`, not from `git show K:TPL`, so a re-run after the absorb mixes `K`-anchored history
with merged content; (c) pin invariance is checked from `G` to `R~1` only, where `R` is the first
`TPL` commit — a pin change between `R` and a later edit, or a `LOC` edit landing before `R`, is not
checked. AC-001 (cell N) and AC-006 already reject the full-rescan form of (a). — Severity: minor —
Class: optional — Suggested fix: extract `SECTION` from `git show K:<TPL path>`; extend the `cmp` to
`K`'s pinned block; take `R` as the first commit touching either copy.

D20. CELL3-PROCEDURE-VARIANT — `spec.md:320-328`, `acceptance.md:278` — cell ③ is described as "two
runs of the pinned procedure", but its tip store is relocated to `SP/g3-tips.txt`, so the run is a
parameter variant of the pin whenever the pinned scan command or tip-recording text names the
worktree checkpoint path. — Severity: minor — Class: optional — Suggested fix: state that the tip
store path is the one parameter cell ③ overrides, and require the pin to write it as a placeholder.

## Regression Check

Defects from the previous iteration:

- D1: PARTIALLY RESOLVED — AC-016 (a)/(b) close the iteration-1 mutant; the tip-recording component
  remains unchecked (D14, measured).
- D2: RESOLVED — `acceptance.md:102-107` is the exact complement of the table.
- D3: RESOLVED — `acceptance.md:370-393`.
- D4: RESOLVED — `spec.md:264-268`, `317-318`.
- D5: RESOLVED — `spec.md:308-316`; `acceptance.md:259-267`.
- D6: RESOLVED — `acceptance.md:123-127`.
- D7: RESOLVED (commit order) — `acceptance.md:293-305`.
- D8: RESOLVED — `spec.md:372-375`; `acceptance.md:167-184`.
- D9: RESOLVED for REQ-004 — `spec.md:118`.
- D10: RESOLVED — `spec.md:26-31`.
- D11: RESOLVED for the six listed ACs; AC-004 left inconsistent (D16).
- D12: RESOLVED — `spec.md:326-328`.

Regressions introduced by the revision:

- D13 — the same-hunk placement (D3 fix) turns an undefined findings granularity into a possible
  unsatisfiable DoD criterion.
- D15 — "fails exactly when" (D2 fix) strips AC-001's construction checks of any consequence.

Cross-layer consistency checked and found sound: REQ ↔ AC matrix (every REQ-001..013 has an AC,
every AC-001..016 names an existing REQ; `acceptance.md:40-57`); plan milestones name AC-016 (M5,
`plan.md:133`) and the re-gate (M3, `plan.md:116-122`) consistently with `spec.md:276-278`; the
renumbered §D.17-§D.19 references in `plan.md:74,158` and `acceptance.md:469` resolve to the right
sections; DoD (`acceptance.md:473`) includes AC-016; AC-005's added leak run matches the real test
name and strict switch; §4 and §5 are unchanged and still consistent. `acceptance.md:267` has two
sentences run together on one long line (cosmetic, not a defect).

No STOP signal: harmonic score 0.79 → 0.79, not lower.

## Recommendation

FAIL at the Tier M ceiling (iteration 2 of 2). All seven must-pass criteria PASS. The orchestrator
escalates to the user with the three LEAN options (PASS-with-debt, scope reduction, explicit
override to iterate). For any of the three, the blocking fixes are small and local:

1. D14 — add a `Tip recording: ` pinned line and a third `grep -cF -f` check to AC-016
   (`plan.md:41-43`, `acceptance.md:410-436`).
2. D13 — define `SP/al-findings.txt` as matched lines only, or count labels over REGEX-matching lines
   only (`acceptance.md:359-363`, `383-386`).
3. D15 — one sentence making AC-001's `is-ancestor` and scan-exit readings a rebuild-and-gap
   condition (`acceptance.md:107`).
4. D16 — end AC-004's diff at `K` and list AC-004 in the anchor paragraph (`acceptance.md:26,142`).

If the user chooses PASS-with-debt, D13-D16 should be carried as run-phase preconditions, fixed in
the SPEC before M1 starts, because D14 and D13 bind the gate and the allowlist criteria. D17-D20 are
optional.

## Gaps

- The judging build is behind HEAD; lint-code invariance was checked for `internal/spec` and
  `internal/cli/spec_lint.go` only, and the installed binary's version string was not read back.
- D13's consequence (the `LISTCELL` line printed inside a reported hunk) is reasoned from the
  unified-diff output format; no fixture repository was built in this run. The ambiguity itself
  (no definition of the findings file) is read from the text.
- D14's mutant was measured on mock files, not on a real pinned procedure (none exists yet).
- AC-016's pickaxe behaviour under move-out was reasoned from `git log -S` count-change semantics,
  not run on a scratch repository.
- The AC-005 strict leak run was not taken (brief scope); whether it is green on the current tree
  independently of this card is unknown.
- Cross-model audit backends were not invoked (no `audit_model` instruction in the brief).

## Residual-risk

With D13-D16 fixed, the gate still measures a documented procedure, not an enforced one; evidence
for cells ① and ② could be re-used across rounds without re-measurement (only cell ③'s fresh lead
approval is a human witness); cell ③'s timing is recorded on a contended machine for information
only. The ref kinds not exercised (tags, remote-only branches, stash, merge resolutions,
`acceptance.md` §D.17-§D.18) remain outside the Definition of Done by design.
