# SPEC Review Report: SPEC-REVIEW-SECRET-SCAN-REFS-001
Iteration: 1/2 (Tier M ceiling 2)
Verdict: FAIL
Overall Score: 0.79 (Tier M PASS threshold 0.80)

card: t629 · auditor: plan-auditor (independent, read-only) · date: 2026-09-10
Reasoning context ignored per M1 Context Isolation — the lead's summary of the author's concerns was
used only as a list of places to look; every reading below was re-measured in this run.

## Baseline-attribution

Tree measured: worktree `.claude/worktrees/t629`, branch `WT-secret-scan-refs`.

```
$ git rev-parse --show-toplevel   → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t629
$ git rev-parse --short HEAD      → b9a436f2d
$ git status --porcelain          → (empty)
```

Judging build for `moai spec lint`: installed `~/go/bin/moai`, `v3.2.0-rc.5 list-974-g84fa4ece4`.
`git merge-base --is-ancestor 84fa4ece4 HEAD` exit 0, so the build is behind HEAD
(`git diff --shortstat 84fa4ece4 HEAD` → `326 files changed`). The lint code paths did not change in
that span: `git diff --stat 84fa4ece4 HEAD -- internal/spec internal/cli/spec_lint.go` printed
nothing, and both paths exist (`ls -d` listed both). Gap: lint could depend on packages outside those
two paths.

Artifacts read in full: `spec.md` (387 lines), `plan.md` (175), `acceptance.md` (382),
`progress.md` (68), `.moai/reports/t629/reproduction.md` (122), `.moai/reports/t629/cost-baseline.md`
(67), and the secret-scan section of the distributed `review.md`. Tier M; no `design.md` or
`research.md` is required.

## Re-measurements this run

| Reading | Command | Observed |
|---|---|---|
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | exit 0; `0 error(s), 0 warning(s)`; one INFO `OwnershipTransitionUnmeasured` on commit `78e29987f` |
| review.md copies unchanged | `git diff --stat feeecc980 HEAD -- <both copies>` | empty, exit 0 |
| copies identical | `diff -q LOC TPL` | exit 0 |
| card diff file set | `git diff --stat feeecc980 HEAD` | 6 files, all under the SPEC dir and `.moai/reports/t629/`; `1201 insertions(+)` |
| AC-004 | `git diff feeecc980 HEAD --output=SP/card-diff.txt`; `wc -l`; `grep -cE '^\+.*(REGEX)'` | `1237` lines; count `0`, exit 1 |
| regex over SPEC + evidence | `/usr/bin/grep -cE -- '(REGEX)' <4 SPEC files> <2 report files>` | `0` for each of 6 files, exit 1 |
| positive control | same grep over a PEM-header line assembled from two fragments in SP | `1`, exit 0 |
| decision commit alone | `git show --stat af7eb142b` | 1 file, `spec.md` only |
| revision commit | `git show --stat b9a436f2d` | 4 SPEC files only |
| REQ numbering | `grep -oE '^- \*\*REQ-[0-9]{3}' spec.md` | REQ-001 … REQ-012, no gap, no duplicate |
| AC numbering | `grep -oE '^\| AC-[0-9]{3}' acceptance.md`; `## §D.N` headings | AC-001 … AC-015 in matrix and in §D.1-§D.15 |
| clarification markers | `grep -rn 'NEEDS CLARIFICATION' plan.md` | exit 1; `research.md` absent |
| cross-SPEC refs | `grep -ohE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' <SPEC files>` | only the SPEC's own ID (12×) |
| syscall | `grep -c syscall <SPEC files>` | `0` each |
| looser equivalence wording | `grep -ci 'same coverage' TPL LOC` | `1`, `1` |
| develop drift on the copies | `git log --oneline feeecc980..refs/heads/develop -- <both copies>` | 0 lines (develop is 119 commits past `feeecc980`) |
| checkpoint path ignored | `git check-ignore -v .moai/state/secrets-scan-checkpoint.txt` | `.gitignore:354:.moai/state/` |
| CI strict leak tier | `internal/template/internal_content_leak_test.go:373`; `template-neutrality-check.yaml` step with `MOAI_TEMPLATE_LEAK_STRICT: '1'` | S2 class `\b[0-9a-f]{7,8}([\s\.,;:!?]|$)` is CI-enforced on template paths |

No full-history scan was run, no matched commit was opened, and no listed allowlist value was written
anywhere.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency** — REQ-001..REQ-012 sequential, 3-digit padding, no
  duplicates (`spec.md:89-145`).
- [PASS] **MP-2 EARS/GEARS format** — judged on the requirement layer (`spec.md` §2) only. All 12
  REQs follow a GEARS shape: Ubiquitous (REQ-001, 003, 006, 011), When (REQ-002, 009, 010),
  Where (REQ-004), Unwanted `shall not` (REQ-005, 007, 008, 012). ACs are Given-When-Then or
  command-sequence verification-layer entries and were not graded here. See D9 for a minor note on
  REQ-004's use of `Where`.
- [PASS] **MP-3 YAML frontmatter** — `spec.md:2-13` carries all 12 canonical fields with correct
  types (`version: "0.2.0"` quoted, `created`/`updated` ISO dates, `priority: P1`,
  `lifecycle: spec-anchored`, `tags` comma-separated string) plus `tier: M`. No rejected alias.
- [PASS] **MP-4 language neutrality** — the SPEC names no language-specific tool; REQ-006
  (`spec.md:117-119`) and AC-005 forbid naming a programming language in the distributed section.
- [PASS] **MP-5 D7 cross-SPEC** — no external SPEC IDs referenced; no BLOCKING finding.
- [PASS] **MP-6 D8 cross-platform** — `syscall` count 0 in all four files; auto-pass.
- [PASS] **MP-7 clarification gate** — no `[NEEDS CLARIFICATION` in `plan.md`; `research.md`
  absent.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | AC-001's required table and its fail conditions state different predicates (D2); cell ② has no exit-code requirement while cell ① does (D5); predicates are attributed to the operator (D4) |
| Completeness | 0.90 | 1.0 minus minor | All sections present (HISTORY `spec.md:19`, WHY §1, WHAT §2/§3, constraints §4, Out of Scope §5 with five `### Out of Scope —` H3s and bullets); one stale HISTORY bullet (D10) |
| Testability | 0.75 | 0.75 | AC-009's detection signal has no pinned command (D5); AC-014 does not fix where the negatives sit relative to the listed value (D3); AC-002 checks one exact phrase only (D6) |
| Traceability | 0.75 | 0.75 | Every REQ has an AC and every AC maps to an existing REQ, but the `spec.md:253-254` constraint that the edit prescribe only the gate-measured procedure has neither a REQ nor an AC (D1) |

Aggregate (mean) 0.79; harmonic mean 0.79. Below the Tier M threshold of 0.80.

## Answers to the audit focus

**Gate predicates — decidable, and which mutants pass?**

- Cell ① (`spec.md:262-274`, AC-008): decidable (`wc -c`, `grep -c`, exit codes, `is-ancestor`).
  A HEAD-only procedure gives `SIDECELL` 0 → untrustworthy. A full-rescan procedure passes ①; the
  SPEC says so at `acceptance.md:208-209` and relies on ③.
- Cell ③ (`spec.md:292-315`, AC-010): the bound is correct. `B − A + L` = |all| − |tips ∩ all| =
  |commits reachable from a ref but not from a recorded tip|, which is exactly the newly reachable
  set. A full rescan has scope `B`, which exceeds the bound whenever `A > L`; on this repository the
  recorded tips include live refs, so `A > L` holds. A too-narrow procedure (HEAD-only, or empty)
  passes ③ but fails ①, so the pair closes both directions. Having no seconds threshold matches the
  operator asking to see wall time; that choice is labelled a judgement at `spec.md:313-315`.
- Cell ② (`spec.md:276-290`, AC-009): a silent skip (exit 0, `GONECELL` 0) is untrustworthy; a
  silent fallback with no detection is untrustworthy. Two holes remain (D5): no exit-code
  requirement on the final or fallback scan, and no pinned command for reading the detection signal.
- **The main hole is in how the cells are combined, not in any one cell (D1).** Each cell measures
  the pinned procedure, but nothing checks that the shipped wording is that procedure. AC-001 re-runs
  only the ①-shaped sequence on the worded text. A run can gate a correct procedure, then word one
  that drops missing-tip handling or rescans all history, and pass every AC.

**Allowlist controls.** AC-013 (positive, raw ≥ 1 and findings 0) paired with AC-014 (negatives
≥ 1 in findings) rejects both an all-suppress and a no-suppress mutant, and `NEARCELL` rejects a
prefix match (`acceptance.md:330-332`). But AC-014 puts the negatives only in "the same file"
(`acceptance.md:320-321`). A procedure that suppresses a whole commit or hunk once it contains a
listed value passes if the negatives were committed separately (D3).

**Cross-layer consistency.** No stale Option 1/3 requirement remains in §2, acceptance, or plan; the
Option 1/3 text in `spec.md` §3.2 is labelled as the decision record (`spec.md:201`). One stale
HISTORY bullet remains (D10). `plan.md` M3 ("If a cell fails, return to M2 before editing any
document", `plan.md:108-109`) lets the procedure change without re-gating (part of D1). No
requirement reaches a file outside the two copies and the SPEC/evidence tree.

**Evidence and privacy.** Satisfied. `cost-baseline.md:53-57` keeps matched SHAs out of the
repository. Cell ③ records counts only (`spec.md:295-297`, `acceptance.md:254-255`,
`plan.md:59-60`). The one SHA in `progress.md:56` is a card commit, not a matched commit. The first
full-history run is gated on lead approval (`plan.md:43-45`, `spec.md:294-296`,
`acceptance.md:236-237`); the approval's timing is not checkable from the commit graph (D7).

**Template neutrality with the allowlist.** Achievable. A full-length hex digest does not match the
CI strict tier's short-SHA class (a 7-8 hex run must be followed by whitespace, punctuation, or end
of line), and fragment assembly adds no SPEC ID, date, or language name. AC-005 does not cover that
CI tier, though, so a digest cut to 7-8 lowercase hex characters would pass AC-005 and turn CI red
(D8).

**Author flags.** (1) Cell ③ predicate: sound, see above. (2) Cell ② strictness: defensible, but it
is stricter than the condition the operator stated, and the text credits it to the operator (D4).
(3) AC-011/AC-012 pickaxe checks: `-S` is a literal string, scoped to `PROG`, with `G ≠ R` and a
count at `R~1`; AC-012's `U..HEAD` / `feeecc980..U` split catches an edit landed in the same commit
as the stop. Sound on a linear card branch; one weakness appears once develop is absorbed (D11).
(4) No plan-time RED for AC-008..AC-014: acceptable — each states why, AC-008 cites the measured
HEAD-anchored failure (`acceptance.md:205-207`), and AC-015 is marked as a regression guard with a
detecting control (`acceptance.md:343-345`).

## Defects Found

D1. GATE-WORDING-LINK — `spec.md:253-254`, `plan.md:108-109`, `acceptance.md:103-104` — §3.4 states
"The document edit may prescribe only a procedure the gate measured", but no REQ and no AC checks it.
AC-001 re-runs only the ①-shaped sequence on the worded text; cells ② (vanished tip) and ③ (scope
bound) are never checked against the shipped wording. `plan.md` M3 sends a failing worded procedure
back to M2 for rewording without re-taking the gate, and nothing checks that the pinned procedure
text in `progress.md` §E.2 is unchanged between the first gate command and the edit commit.
Mutant that passes every AC: gate a correct procedure, then word one that drops missing-tip handling.
This contradicts the operator's measure-first condition. — Severity: major — Class: blocking —
Required fix: add a requirement (e.g. REQ-013, When the copies are edited, the prescribed procedure
shall be the one pinned before the gate) and an AC that (a) checks the pinned scan command and the
pinned missing-tip handling text each appear verbatim in `SECTION` (`grep -cF` → ≥ 1 each), and
(b) checks the pinned-procedure block in `PROG` §E.2 is byte-identical at `G` and at `R~1`
(`git show G:PROG` vs `git show R~1:PROG`, block extracted by heading). Change `plan.md` M3 so any
change to the procedure re-takes gate cells ①-③ (③ with fresh lead approval) before editing.

D2. AC001-PREDICATE-SPLIT — `acceptance.md:86-97` — the required-outcome table requires R2
`SIDECELL` ≥ 1 and R3 `LATECELL` ≥ 1, but the fail conditions only fail when `SIDECELL` totals 0
"across R2 and R3" and `LATECELL` totals 0 "across R3 and R4". A procedure that reports a side ref one
review late satisfies the fail-condition reading and violates the table and REQ-001
(`spec.md:92-93`, "the first completed scan after its commit becomes reachable"). Two readings of one
AC. — Severity: major — Class: blocking — Required fix: make the fail conditions the exact complement
of the table (fail when R2 `SIDECELL` = 0, or R3 `LATECELL` = 0, or H = 0, or B ≠ 0, or N ≠ 0), or
delete the fail-condition paragraph and keep the table as the only predicate.

D3. AC014-GRANULARITY — `acceptance.md:320-332` — the negatives `NEARCELL` and `OTHERCELL` must sit
only "in the same file" as `LISTCELL`. If they are committed separately, a procedure that drops a
whole commit or hunk containing a listed value passes AC-013 and AC-014, yet violates REQ-011's
matched-text exactness (`spec.md:137-140`). — Severity: minor — Class: blocking — Required fix:
require `NEARCELL` and `OTHERCELL` in the **same commit and the same hunk** as `LISTCELL` (adjacent
lines), and add that granularity mutant to the mutant-pairing list. Optional extension: one line
carrying both the listed value and an unlisted credential-shaped value, required to appear in
findings.

D4. PREDICATE-ATTRIBUTION — `spec.md:243-245` vs `spec.md:262-315` — the Source paragraph credits
§3.4's conditions to the operator. The operator's stated conditions are the three measurements and
the stop rule; the cell predicates (the detection requirement and two recovery paths in ②, the
`B − A + L` bound in ③, the ≥ 1 thresholds in ①) are the lane's operationalization. Only cell ③
labels its predicate a judgement. The detection requirement in ② is stricter than anything the
operator stated. A decision record should keep who decided apart from who operationalized. —
Severity: minor — Class: blocking — Required fix: add one sentence after `spec.md:245`: the three
measurements and the stop rule are the operator's; each cell's predicate is the lane's
operationalization, submitted for plan audit and open to operator correction. Mark cell ②'s
detection requirement as lane-authored.

D5. CELL2-EXIT-AND-DETECTION — `spec.md:284-290`, `acceptance.md:219-230` — (a) cell ① requires
"every scan exits 0" (`spec.md:269`), but cell ② sets no exit-code condition on the final or
fallback scan. A fallback that prints `GONECELL` and then fails with a non-zero exit (partial output)
counts as trustworthy, and the line between that and "an error with no follow-on scan" is not
drawn. (b) The detection signal is "read from `SP/g2.err` or the procedure's captured output" with no
command, so "naming it" is left to judgement. — Severity: minor — Class: blocking — Required fix:
(a) add "the scan carrying the final result exits 0" to the trustworthy predicate;
(b) pin `/usr/bin/grep -cF '<G1>' SP/g2.err SP/g2.txt` with a total ≥ 1 as the detection reading
(plus `SP/g2-fallback.txt` if it exists).

D6. AC002-PHRASE-ONLY — `acceptance.md:106-113` — REQ-005 (`spec.md:113-115`) forbids stating that
the incremental scan gives the same coverage as a full-history scan, but AC-002 checks only the two
current exact phrases. A reworded claim ("gives the same coverage as a full scan") passes. Measured
now: `grep -ci 'same coverage' TPL LOC` → `1`, `1`. — Severity: minor — Class: optional —
Suggested fix: add `grep -ci 'same coverage' LOC TPL` → `0` to AC-002 (baseline `1`, `1`).

D7. APPROVAL-ORDERING — `acceptance.md:236-237`, `spec.md:294-296`, `plan.md:43-45` — "the lead's
approval … recorded before the run" is a precondition with no check. If approval and timing evidence
land in one commit, the order cannot be verified afterwards (same shape as VCI §2.3). — Severity:
minor — Class: optional — Suggested fix: commit the approval record (quoting the lead's message) in
its own commit before the cell ③ evidence commit, and add an `is-ancestor` check to AC-010; or state
that the lead's own sent message is the ordering witness.

D8. AC005-CI-TIER — `acceptance.md:140-152` — AC-005 does not run the CI strict leak tier that
guards template paths (`internal_content_leak_test.go:373`, S2 short-hex class; enforced in
`template-neutrality-check.yaml` with `MOAI_TEMPLATE_LEAK_STRICT: '1'`). A digest representation cut
to 7-8 lowercase hex characters would pass AC-005 and fail CI. Separately, `-w go` with `-i` matches
the English verb "go" in prose, and `r` is left out of the 16-language list without explanation. —
Severity: minor — Class: optional — Suggested fix: state in `spec.md` §3.4 that a digest is written
at full length, or add the scoped `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run
TestTemplateNoInternalContentLeak` to AC-005; note the `go`/`r` handling.

D9. REQ-SHAPE — `spec.md:107-111`, `spec.md:137-140` — REQ-004 uses `Where` for a static property of
the adopted design, where GEARS reserves `Where` for a capability gate (acceptable as legacy EARS
Optional within the compatibility window). REQ-011 packs four obligations into one entry. —
Severity: minor — Class: optional — Suggested fix: restate REQ-004 as Ubiquitous; split REQ-011 if
the run phase wants separate verdicts.

D10. STALE-HISTORY — `spec.md:26-30` — the unversioned HISTORY sub-bullet "The choice between the
options in §3 is **deliberately left open** … does not choose" and the 0.1.1 note "The decision in §3
is still pending" read as current unless the reader reaches the 0.2.0 bullet. — Severity: minor —
Class: optional — Suggested fix: prefix the first with "(0.1.0)" and add "superseded by 0.2.0" to
both.

D11. ABSORB-RANGE — `acceptance.md:119`, `180`, `270-276`, `289-291` — AC-003, AC-007, AC-011, and
AC-012 take ranges from `feeecc980..HEAD` with a path filter. After the develop absorb that the
integration window requires, any develop commit touching either copy would enter those ranges: a
false "both edited" pass in AC-003, a false non-ancestor fail in AC-007/AC-011, a false count in
AC-012. Latent today: `git log --oneline feeecc980..refs/heads/develop -- <both copies>` → 0 lines.
— Severity: minor — Class: optional — Suggested fix: run the closure checks before the absorb and
record the SHA they ran at, or add `--first-parent` to the range commands.

D12. CELL3-STATE-LOCATION — `spec.md:292-297` — running the pinned procedure on this repository will
record a tip set. If the procedure writes to `.moai/state/` in the worktree (ignored by
`.gitignore:354`, so no commit risk), a later real review in that tree would start from the gate's
tips. — Severity: minor — Class: optional — Suggested fix: pin the cell ③ tip store under `SP`.

## Regression Check

Not applicable (iteration 1).

## Recommendation

FAIL on score (0.79 < 0.80) with five blocking findings; all seven must-pass criteria PASS. In
priority order:

1. D1 — tie the shipped wording to the gate-measured procedure (new REQ + AC; pinned-block
   invariance between `G` and `R~1`; M3 re-gates on any procedure change).
2. D2 — make AC-001's fail conditions the exact complement of its required table.
3. D5 — add an exit-0 requirement for cell ②'s final scan and pin the detection grep.
4. D3 — put AC-014's negatives in the same commit and hunk as `LISTCELL`.
5. D4 — separate the operator's conditions from the lane's predicates in §3.4.

D6-D12 are optional; D6 and D10 are one-line edits worth taking in the same revision. The re-audit
(iteration 2, the Tier M ceiling) will cover the D1-D12 delta plus a regression check.

## Gaps

- The judging build is behind HEAD; lint-code invariance was checked for `internal/spec` and
  `internal/cli/spec_lint.go` only.
- No fixture was built; gate behaviour (for example, git's error text for a pruned `--not` tip) is
  reasoned, not measured.
- The operator's exact wording is available only as relayed in `spec.md` §3 and the lead's brief; D4
  rests on that relay.
- Cross-model audit backends were not invoked (no `audit_model` instruction in the brief).

## Residual-risk

With D1-D5 fixed, the gate still measures on a contended machine, and cell ③'s timing is recorded
for information only. The ref kinds not exercised (tags, remote-only branches, stash, merge
resolutions, `acceptance.md` §D.16-§D.17) stay outside the Definition of Done by design.
