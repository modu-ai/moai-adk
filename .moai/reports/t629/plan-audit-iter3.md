# SPEC Review Report: SPEC-REVIEW-SECRET-SCAN-REFS-001
Iteration: 3/2 — **ceiling override**: the Tier M ceiling is 2 (`.moai/config/sections/harness.yaml`
`plan_audit_tier_ceilings` M: 2). Iteration 2 returned FAIL at that ceiling. The operator approved
exactly one further audit, exceeding the ceiling by one (answered in the lead session, relayed by the
lead, 2026-09-10). The operator also directed that if this iteration FAILs, no further fixes are made
and the question returns to the operator.
Verdict: PASS
Overall Score: 0.89 harmonic (arithmetic mean 0.89; Tier M PASS threshold 0.80)

card: t629 · auditor: plan-auditor (independent, read-only) · date: 2026-09-11
Reasoning context ignored per M1 Context Isolation — the lead's brief was used only as the scope
statement and a list of places to look; every reading relied on below was re-measured in this run.

**Scope (binding, from the brief).** This iteration audits only (1) whether iteration 2's blocking
findings D13, D14, D15, D16 are resolved on the current text, and (2) regressions or new defects
created by the edits that resolved them. D17-D20 were optional and deliberately not taken; they are
not counted against the verdict. Unchanged text outside the delta is not re-audited; one serious
observation about it is listed under "Out of scope, not scored".

## Baseline-attribution

Tree measured: worktree `.claude/worktrees/t629`, branch `WT-secret-scan-refs`.

```
$ git rev-parse --show-toplevel   → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t629
$ git rev-parse --short HEAD      → 121790a06
$ git status --porcelain          → (empty)
```

Delta under audit: `git diff 5fa97ecd7 121790a06` (267 lines, 4 files: acceptance.md +75/−, plan.md
7, progress.md +58, spec.md 9). Two commits:

```
121790a06 2026-09-11 03:09:01 +0900 docs(t629): resolve plan audit iteration 2 findings D13-D16 in the SPEC (card t629)
3bb7f2423 2026-09-10 23:32:46 +0900 docs(t629): measure secret-scan output granularity on a fixture for D13 (card t629)
```

`git show --name-only 3bb7f2423` → `progress.md` only; `git show --stat 121790a06` → the four SPEC
files only, `76 insertions(+), 30 deletions(-)`. The measurement commit is a strict predecessor of the
definition that rests on it, so the commit graph witnesses the ordering
(`verification-claim-integrity.md` §2.3).

Judging build for `moai spec lint`: the installed `moai`; the MCP server banner in this session names
`v3.2.0-rc.5`, commit `84fa4ece4`. Iteration 2 established that no change under `internal/spec` or
`internal/cli/spec_lint.go` lies between that commit and the card branch; the delta under audit
touches only SPEC markdown, so that attribution carries unchanged. Gap: not re-derived in this run.

Artifacts read: iteration 2 report in full; the full delta; current `acceptance.md` in full (506
lines); `plan.md` in full (191 lines); `spec.md` lines 1-60, 95-170, 255-345 (frontmatter, HISTORY,
all REQs, §3.4 gate); `progress.md` section anchors.

## Re-measurements this run

| Reading | Command | Observed |
|---|---|---|
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | exit 0; `0 error(s), 0 warning(s)`; one INFO `OwnershipTransitionUnmeasured` on commit `78e29987f` |
| REQ numbering | `grep -oE '^- \*\*REQ-[0-9]{3}' spec.md`; `sort \| uniq -d \| wc -l` | REQ-001 … REQ-013 in order; duplicates `0` |
| AC numbering | `grep -oE '^\| AC-[0-9]{3}' acceptance.md`; `grep -nE '^## §D\.'` | AC-001 … AC-016; §D.1-§D.16 carry AC-001-AC-016 in order; §D.17 candidate, §D.18 edge cases, §D.19 DoD |
| decision / placeholders | `grep -c '^\*\*Decision:\*\* Option 2' spec.md`; `grep -c 'PENDING-'` over the 4 files | `1`; `0` each |
| credential regex | `/usr/bin/grep -cE -- '<REGEX>'` over the 4 SPEC files and `plan-audit-iter2.md` | `0` each, exit 1 |
| AC-004 at `121790a06` | `git diff feeecc980 121790a06 --output=SP/card-diff-121790a06.txt`; `wc -l`; `/usr/bin/grep -cE -- '^\+.*(<REGEX>)'` | diff exit 0; `2101` lines; count `0`, exit 1 |
| AC-004 positive control | same grep over an `SP` line `+x <PEM header>` assembled by `printf` from three fragments | `1`, exit 0 |
| review.md copies | `git diff --stat feeecc980 121790a06 -- <LOC> <TPL>`; `diff -q LOC TPL` | empty, exit 0; exit 0 |
| clarification markers | `grep -rn 'NEEDS CLARIFICATION' plan.md`; `ls` SPEC dir | exit 1; no `research.md` |
| cross-SPEC refs / syscall | `grep -ohE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` / `grep -c syscall` over the 4 files | only the SPEC's own ID (15×) / `0` each |
| progress.md anchors | `grep -nE '^## §E\.\|Scan output granularity measurement' progress.md` | `§E.1` at line 6; the measurement bullet at line 79; `§E.2` at line 138 — the measurement sits in §E.1 as HISTORY and AC-013 cite |
| pickaxe granularity (docs) | `man git-log \| col -b \| grep -A4 -- '--pickaxe-all'` | "When -S or -G finds a change, show all the changes in that changeset, not just the files that contain the change" — the default is file-granular, matching the recorded measurement |
| D14 mechanics + mutant | mock pin block (heading + three pinned lines + `####` sub + next `###`), a correct section, and a section recording tips **after** the scan; AC-016's `sed -nE` extraction, the three `sed -n 's/^… //p'` lines, `wc -l`, `grep -c .`, and the three `grep -cF -f` | block `5` lines; each pattern file `1` line, `grep -c .` `1`; correct section `1`/`1`/`1`; mutant section tip-recording `0` (exit 1), scan `1`, missing-tip `1` → AC-016 (a) fails the mutant |
| D13 counting rule + mutants | mock patch-shaped scan output (20 lines: commit/diff/`@@` headers, two context lines, the four same-hunk labels, `HEADCELL` and `SIDECELL` in other commits), PEM-header stand-ins assembled by `printf` from two fragments each; raw REGEX extract; one findings file per mutant; each findings file REGEX-extracted per AC-013; `grep -c` per label | table below |

D13 mock, per-label counts over the REGEX extracts (`raw` is the extract of the raw scan; context
lines: raw file `2`, raw extract `0`):

| Findings file | LISTCELL | NEARCELL | OTHERCELL | MIXCELL | HEADCELL | SIDECELL | AC-013 | AC-014 |
|---|---|---|---|---|---|---|---|---|
| raw extract | 1 | 1 | 1 | 1 | 1 | 1 | (raw ≥ 1 holds) | — |
| correct exact-value | 0 | 1 | 1 | 1 | 1 | 1 | pass | pass |
| no suppression | 1 | 1 | 1 | 1 | 1 | 1 | **fail** | pass |
| suppress everything | 0 | 0 | 0 | 0 | 0 | 0 | pass | **fail** |
| path / whole commit / whole file / whole hunk | 0 | 0 | 0 | 0 | 1 | 1 | pass | **fail** |
| whole line | 0 | 1 | 1 | 0 | 1 | 1 | pass | **fail** (MIXCELL) |
| prefix match | 0 | 0 | 1 | 1 | 1 | 1 | pass | **fail** (NEARCELL) |
| patch output as findings | 1 | 1 | 1 | 1 | 1 | 1 | **fail** | pass |

No history scan was run on this repository, no matched commit was opened, no access-key-shaped or
listed example value was written (whole or fragmented), no `go test` was run, nothing was committed.
The only file written inside the worktree is this report.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** — REQ-001..REQ-013 sequential, 3-digit padding, duplicate
  count `0` (`spec.md:107-169`). The delta does not touch any REQ.
- [PASS] **MP-2 EARS/GEARS format** — judged on the requirement layer (`spec.md` §2) only. The delta
  changes `spec.md` only at `version` (`spec.md:4`) and one HISTORY bullet (`spec.md:47-53`); every
  REQ keeps the pattern iteration 2 recorded (Ubiquitous REQ-001, 003, 004, 006, 011; When REQ-002,
  009, 010, 013; Unwanted REQ-005, 007, 008, 012). ACs were not graded here.
- [PASS] **MP-3 YAML frontmatter** — `spec.md:2-14`: all 12 canonical fields; `version: "0.2.2"`
  quoted semver; ISO dates; `priority: P1`; `lifecycle: spec-anchored`; comma-separated `tags`; plus
  `tier: M`. No rejected alias.
- [PASS] **MP-4 language neutrality** — no language-specific tool named in the delta.
- [PASS] **MP-5 D7 cross-SPEC** — only the SPEC's own ID referenced; no BLOCKING finding.
- [PASS] **MP-6 D8 cross-platform** — `syscall` count `0` in all four files; auto-pass.
- [PASS] **MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' plan.md` exit 1;
  `research.md` absent.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | D13's findings file is now defined (`acceptance.md:363-368`) and D15's construction readings carry a consequence (`acceptance.md:109-110`). Residual minor: AC-001 classifies a procedure-caused non-zero exit as a construction error (N1); the counting rule assumes a finding line carries the matched text (N3) |
| Completeness | 0.95 | 1.0 minus minor | All sections present; HISTORY records 0.2.2 and the ceiling override (`spec.md:47-53`); the measurement basis is recorded with commands, controls, counts, and gaps (`progress.md:79-117`) |
| Testability | 0.85 | 0.75-1.0 | AC-013/AC-014 counting rule measured against every mutant in the pairing (table above); AC-016's third pinned line measured against the record-after-scan mutant. Residual minor: a procedure that errors under AC-001 stays a rebuild gap rather than an explicit fail (N1) |
| Traceability | 0.90 | 0.75-1.0 | REQ-013's three pinned components each have a check (`acceptance.md:444-450`); AC-004 now agrees with plan M5 (`acceptance.md:26`, `145`; `plan.md:131-133`). Residual minor: AC-004's "all of the card's commits" overstates a range ending at `K` (N2) |

Harmonic mean 0.886 → 0.89; arithmetic mean 0.8875 → 0.89. Tier M threshold 0.80. Score movement
0.79 (iteration 2) → 0.89: no regression, no STOP signal.

## Resolution of iteration 2 blocking findings

| ID | Status | Evidence on current text |
|---|---|---|
| D13 AC013-014-FINDINGS-GRANULARITY | **Resolved** | Definition `acceptance.md:363-368`: one line per finding, the line carrying the unsuppressed match; not the patch output; "a line printed with a finding is not a finding"; labels counted only over `REGEX`-matching lines. Commands `acceptance.md:371-374` (raw and findings each REGEX-extracted, then `grep -c 'LISTCELL'`); AC-014 reuses the same files and rule (`acceptance.md:397-400`, `402`). Under this definition a correct exact-value procedure reads `LISTCELL 0` and every negative ≥ 1 (mock table). Every mutant in the pairing rejects: no suppression and patch-output-as-findings fail AC-013; suppress everything, path, whole commit/file/hunk, whole line, and prefix fail AC-014 (`acceptance.md:405-411`, mock table). The raw extract's `LISTCELL ≥ 1` is a positive control on the same extraction command, so a wrongly substituted `REGEX` cannot yield a vacuous `0`. |
| D13 — measurement basis | **Sound as a basis** | `progress.md:79-117`: fixture outside the repository, markers from fragments, git version recorded, every output to a file, exit codes read without a pipe. Controls: pre-marker `--all` scan 0 bytes; untouched file `OTHERPLAIN 0`; F2 separates file from commit (non-matching file `FOXPLAIN` patch `1` / scan `0`); F3 separates hunk from file (non-matching hunk `HOTELPLAIN` `1` / `1`). The F1 line count (21) is consistent with a header block plus one hunk of 6 context and 4 added lines. Conclusion "file-granular within a matching commit" follows from F2/F3 and agrees with git's own `--pickaxe-all` documentation. Counts only; gaps stated (one regex alternative, one git build, `--pickaxe-all`, context width, merges, stash). |
| D14 AC016-TIP-RECORDING-UNCHECKED | **Resolved** | Pin format `plan.md:40-44` and AC-016 Given `acceptance.md:428-432` add a `Tip recording: ` line carrying how **and when** the tip set is recorded; extraction `acceptance.md:444`; the one-non-empty-line rule now binds all three pattern files (`acceptance.md:454-455`); (a) adds `grep -cF -f SP/pin-tips.txt` (`acceptance.md:448`); all three counts ≥ 1 (`acceptance.md:456-457`); the record-after-scan mutant is listed (`acceptance.md:460-461`). Measured on mock files: mutant tip-recording count `0` while the scan and missing-tip counts stay `1`, so the mutant now fails AC-016. Matrix row updated (`acceptance.md:57`). |
| D15 AC001-CONSTRUCTION-CONSEQUENCE | **Resolved** | `acceptance.md:109-110`: an `is-ancestor` reading other than exit 1 (steps 3, 5, 8) or a non-zero exit from any of R1-R4 → rebuilt, gap, not a pass. D2's closure is not re-opened: the fail conditions (`acceptance.md:102-107`) are unchanged, and the added sentence can turn a reading into a gap but never into a pass, so a one-review-late procedure (every exit 0, `is-ancestor` exit 1) still fails on cell S or T. A gap cannot satisfy the DoD (`acceptance.md:496`). See N1 for a residual attribution nuance (optional). |
| D16 AC004-ANCHOR-INCONSISTENCY | **Resolved** | Anchor list now names AC-004 (`acceptance.md:26`); command `git -C <worktree> diff feeecc980 K` (`acceptance.md:145`); plan-time reading relabelled as taken at `21e5837dc` in place of `K` (`acceptance.md:151-152`); plan M5 runs AC-004 at `K` (`plan.md:131-133`). The two artifacts agree. See N2 for the Given's scope wording (optional). |

## Regression Check

Defects from the previous iteration (blocking):

- D13: RESOLVED — `acceptance.md:363-374`, `397-411`; measured on a mock (table above).
- D14: RESOLVED — `plan.md:40-44`; `acceptance.md:428-461`; measured on a mock.
- D15: RESOLVED — `acceptance.md:109-110`.
- D16: RESOLVED — `acceptance.md:26`, `145`, `151-152`; `plan.md:131-133`.

Optional D17-D20: not taken, by decision; not scored.

Regression readings on the edited surface:

- **Summary matrix.** AC-016 row now says "all three pinned lines found" (`acceptance.md:57`); AC-014
  row unchanged and still accurate; AC-004 row "added-line grep for `REGEX` over the card diff"
  (`acceptance.md:45`) remains accurate for a diff ending at `K`.
- **DoD.** Unchanged (`acceptance.md:496-506`); still requires AC-001-AC-011 and AC-013-AC-016.
- **§D numbering and cross-references.** §D.1-§D.19 unchanged; `plan.md` M3 still points to §D.1,
  §D.13, §D.14 (`plan.md:116`); §D.17/§D.18 references at `plan.md:75,159` unchanged. The new
  `progress.md` §E.1 citations (`spec.md:50`, `acceptance.md:366`) resolve: the measurement bullet
  sits at `progress.md:79`, between `§E.1` (line 6) and `§E.2` (line 138).
- **REQ/AC counts.** 13 REQs, 16 ACs — within the Tier M ceiling of 16; no REQ or AC added or removed.
- **HISTORY / version.** `version: "0.2.2"` (`spec.md:4`); the 0.2.2 bullet lists exactly D13-D16
  and the ceiling override (`spec.md:47-53`), matching the delta. `updated: 2026-09-10` while
  `121790a06` is dated 2026-09-11 03:09 +0900; the spec.md changes themselves are consistent with a
  2026-09-10 edit, so this is recorded as an observation, not a defect.
- **review.md copies.** Unchanged since `feeecc980` (diff-stat empty, exit 0) and byte-identical
  (`diff -q` exit 0).
- **Lint / credential regex / AC-004.** Lint 0/0 with the known INFO; regex `0` in all four SPEC files
  and the iteration 2 report; AC-004 over `feeecc980..121790a06` `2101` lines, count `0`, with a
  positive control `1`.

## Defects Found

No blocking defect. New optional findings created by the D13-D16 edits:

N1. AC001-EXIT-ATTRIBUTION — `acceptance.md:109-110` vs `acceptance.md:241-243` (AC-008) and
`acceptance.md:268-271` (AC-009) — the added sentence treats a non-zero exit from R1-R4 as a
construction failure (rebuild, gap). A non-zero exit can equally be the worded procedure's own
defect; in the sibling gate cells the same reading is a predicate failure (`untrustworthy`). A
procedure that errors under AC-001 therefore never produces an explicit fail — it stays a gap, so
`plan.md` M3's "If a cell fails" routing (`plan.md:117-118`) does not fire. No false pass: a gap is
not a PASS and the DoD blocks. The wording was the fix iteration 2 prescribed. — Severity: minor —
Class: optional — Suggested fix: keep the rebuild rule for the `is-ancestor` readings, and state that
a non-zero exit from R1-R4 on a correctly constructed sequence is a fail of AC-001.

N2. AC004-GIVEN-SCOPE — `acceptance.md:144-145` — the Given says "all of the card's commits,
including evidence under `.moai/reports/t629/`", while the range now ends at `K`. Commits made after
`K` — at minimum the commit that records `K` and the closure-check readings in `PROG` §E.2
(`plan.md:131-135`), and any later sync-phase commit — are outside the range, and a re-run after the
absorb deliberately keeps `K`. A single-moment check cannot cover later commits whichever anchor it
uses, so this is mainly a wording overstatement; the failure direction is a false pass for a
credential-shaped line added after `K`. — Severity: minor — Class: optional — Suggested fix: reword
the Given to "all of the card's commits up to `K`", or add a reading over the card's own post-`K`
commits that excludes the absorb's develop parent (e.g. `git log -p --first-parent K..<last card commit>`
written to a file, then the same added-line grep).

N3. AC013-FINDING-TEXT-ASSUMPTION — `acceptance.md:363-368`, `402` — counting labels only over lines
matching `REGEX` assumes every finding line carries the matched text verbatim. A procedure that
displays findings with the value masked or truncated would read `0` for every AC-014 negative and
fail AC-014 although its suppression is correct. The definition "the line carrying the unsuppressed
match" already tells a careful tester to write the source line of each reported finding, so a
consistent reading exists. — Severity: minor — Class: optional — Suggested fix: add "the source line,
whatever display form the procedure uses" to the findings-file definition.

## Out of scope, not scored

- AC-008 (`acceptance.md:241-243`) and `spec.md` cell ① (`spec.md:299-305`) require "every scan
  exits 0", including the first completed scan on a clean history. If the pinned procedure's step
  ends in a filter that exits non-zero on no match, that cell reads `untrustworthy` for a correct
  procedure. Unchanged text; noted because N1 touches the same exit-code semantics. The wording of
  the scan command in the pin decides whether it bites.
- D19(a) (optional, not taken) still bounds the D14 mutant claim: `grep -cF` ≥ 1 proves the pinned
  tip-recording line appears in `SECTION`, not that nothing else in `SECTION` prescribes a different
  timing. AC-001 cell T does not distinguish tip-recording timing either.

## Recommendation

PASS at 0.89 against the Tier M threshold 0.80. All seven must-pass criteria PASS with the evidence
above; D13, D14, D15, and D16 are resolved on the committed text at `121790a06`, and the D13 and D14
resolutions were checked against their mutants on mock files in this run.

- MP-1: REQ-001..013, no gap, no duplicate (`spec.md:107-169`).
- MP-2: requirement layer untouched by the delta; GEARS patterns as recorded in iteration 2.
- MP-3: 12 canonical fields plus `tier`, version `"0.2.2"` (`spec.md:2-14`).
- MP-4-MP-7: no language-specific tool, no foreign SPEC reference, no `syscall`, no clarification
  marker, no `research.md`.

N1-N3 are optional; each is a one-sentence wording change the orchestrator may take or leave. This
verdict does not bypass the Implementation Kickoff Approval gate, which remains mandatory and
score-independent.

## Gaps

- The D13 and D14 mutant checks ran on mock files, not on a git fixture and not on a real pinned
  procedure (none exists yet). The worktree guard refused a compound mock-building command; the
  mock was rebuilt from plain commands. The mock output is patch-shaped by hand; its granularity is
  taken from the recorded measurement and git's documentation, not re-measured with git in this run.
- Only the PEM-header alternative of `REGEX` was used as a stand-in; the access-key alternative was
  not exercised (writing such a value is forbidden by the brief).
- The measurement in `progress.md` §E.1 was read and checked for internal consistency and against
  git's documentation; its fixture was not rebuilt.
- Lint-code invariance between the judging build `84fa4ece4` and the card branch was carried from
  iteration 2, not re-derived.
- `spec.md` was read by region (lines 1-60, 95-170, 255-345), not in full; the delta shows no change
  elsewhere in that file.
- The AC-005 strict leak tier was not run (not in scope). Cross-model audit backends were not
  invoked (no `audit_model` instruction in the brief).

## Residual-risk

With D13-D16 resolved, the gate still measures a documented procedure, not an enforced one. The
allowlist criteria are exercised on one fixture shape and one regex alternative. AC-016 proves
containment of the pinned lines, not exclusivity (D19, optional). Commits after `K` are not covered
by AC-004 (N2). Cell ③'s timing remains informational on a contended machine.
