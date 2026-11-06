auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-CONTAINMENT-MAPPING-DOC-001 (plan-audit round 2)

Reasoning context ignored per M1 Context Isolation (no author reasoning was passed; spec.md, plan.md, acceptance.md, progress.md and the round-1 verdict were read).

verdict: FAIL
audited_sha: 1b2319571070c4d3b60d270037d2f90f7d7ffa8d
overall_score: 0.875
must_pass_failed: 0
blocking_count: 2
plan_artifact_hash: cc8ac755507f762b58d6e3a03460647b87c4c7fe8fa6bdfd91707e78b172b5c5 (auditor-computed `cat .moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/*.md | shasum -a 256`, same method as round 1; NOT the runtime ComputeHash, see Gaps G5)
scope: delta
fix_scope: acceptance.md#E10a acceptance.md#E10b acceptance.md#AC-CMM-010 acceptance.md#Evidence-ledger acceptance.md#Definition-of-Done acceptance.md#E08 acceptance.md#AC-CMM-008 acceptance.md#Instrument-validation
defect_class: N1=ac-wording N2=ac-wording
reread_hunks: acceptance.md#E10a acceptance.md#E10b acceptance.md#E08 acceptance.md#Definition-of-Done
convergence_overall: fail
required_backend: codex fail

Iteration: 2/3. Tier M (threshold 0.80). Previous score 0.66, now 0.875 (no regression, no STOP). The aggregate clears the Tier M threshold, but the verdict is FAIL because the required codex gate returned `fail` and two blocking findings are open (admission needs blocking_count 0 and a passing required backend). Auditor Version: plan-auditor (sonnet-5-5 in-session) plus the codex required gate through `mcp__moai__audit_multi`.

## Round-1 blocker status

| Round-1 defect | Status | One-line basis |
|---|---|---|
| D1 E08 `grep -w` misses Hangul-attached tokens | RESOLVED | New E08 uses explicit `[^A-Za-z0-9_]` boundaries. Run on a 24-line must-hit fixture including `t1760은`, `2026년`, `develop를`, `카드t1760`, `한글2026년`: 24 of 24 counted. Run on a 25-line must-stay-zero fixture (`developer를`, `30,000건`, `v3.2.0은`, `spec-anchored`, ...): 0. Run on a Korean passing fixture: 0. |
| D2 E07 refused by guard; counts occurrences; suffix after `.md` passes | RESOLVED | New E07 is one plain awk call with no `system()` and no `|` in the program text; it was admitted by the guard and executed here. It prints `TOO-FEW-DISTINCT n=1` for eight citations of one path, `BAD-SUFFIX` for `.md.NONEXISTENT`, `MISSING` for a nonexistent `.md`, `TOO-FEW-DISTINCT n=7` for seven distinct paths, and exits 0 on nine distinct existing paths. |
| New: E10a/E10b scope guard | OPEN, see N1 | E10a exits 1 on the current tree for a reason the work can never fix. |
| New: E08 date shapes | OPEN, see N2 | Lowercase month names and day-first order are not counted. |

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: spec.md:L64-L74 carries REQ-CMM-001..011, sequential, no gap, no duplicate, uniform zero-padding. REQ count 11 and AC count 14 are inside the Tier M ceiling of 16 each. The traceability verb printed `COLLECTED: 11 REQ definitions (acceptance input: read)`.
- [PASS] MP-2 GEARS, judged on the requirement layer (`REQ-XXX` in spec.md only): all eleven carry a labelled pattern at spec.md:L64-L74 (Ubiquitous x7, Event-driven L67/L71, State-driven L69, Unwanted legacy L73). The Given/When/Then entries in acceptance.md are ACs and were not graded here.
- [PASS] MP-3 frontmatter: all 12 canonical fields at spec.md:L2-L14, correct types, `version: "0.1.0"` quoted, `tags` a comma string, no rejected snake_case alias. `mcp__moai__spec_audit` (project_root = this worktree) returned `total_specs: 1`, `modern_era_clean: 1`, only an INFO `EraAutoDetected` (H-5).
- [N/A] MP-4 language neutrality: single-repo documentation SPEC, no multi-language tooling.
- [PASS] MP-5 D7: `grep -Eo 'SPEC-...'` over spec/plan/acceptance printed only `SPEC-CONTAINMENT-MAPPING-DOC-001` (self) and `SPEC-AUTH-001` (acceptance.md, a must-hit test string in the E08 fixture list, not a dependency). No referenced SPEC has a status to reconcile; no `REVIEW:` line.
- [PASS] MP-6 D8: `grep -c syscall` over spec.md, plan.md, acceptance.md printed `0`, `0`, `0`; the D8 awk verb printed no `BLOCKING:` line.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` printed nothing for plan.md; `research.md` does not exist (Tier M), grep exit 2 for that reason only.
- [PASS] MP-8 RED-now re-execution. Every release-blocking criterion (AC-CMM-001..009) was re-run on the current tree and the RED reproduced: E01 exit 1 empty stdout; E02, E04a, E04b, E06, E08, E09 exit 2 with `ugrep: warning: ... No such file or directory`; E03, E05, E07 exit 2 with `awk: can't open file .moai/docs/multi-agent-containment-mapping.md`. These match the ledger entries (command, stdout, stderr, exit) one for one; the document-level pin `cf5a52f55...` names a tree where the document equally does not exist, and the document is still absent at HEAD `1b2319571`. No cited command was refused by the guard. No cited command is a test runner. Each reproducing RED is a missing-file red, which is the stated reason in each criterion's "RED reason", so none is vacuous or wrong-reason. AC-CMM-010 is explicitly a regression-guard (no RED-now cell) and is excluded from MP-8.
- [PASS, hand-read] MP-9 cross-artifact ordering. The CN-4 verb printed `COLLECTED: 0 milestones in plan order (none), 0 exit bindings, 7 ordering candidates` and `GAP: 0 milestone headings collected from the plan`, because plan.md lists M1-M4 as numbered list items (plan.md:L73-L76), not headings, and has no `Exit:` lines. The verb therefore did not observe the order mechanically (recorded as Gap G3, not read as a pass). Read by hand: plan order is M1 re-verify, M2 skeleton, M3 rows, M4 acceptance battery (plan.md:L73-L76). The seven candidates were each read: acceptance.md:L7, L24, L136 (E05 ledger text), L188 (E10b ledger), L205, L212 order nothing across milestones; acceptance.md:L236 ("evaluated at the run-phase tip ... before the sync phase begins") names the sync phase, which the plan does not schedule as a milestone, and M4 (plan.md:L76) runs E10a/E10b last, so the clause is satisfiable. Green paths named in acceptance.md ("M2", "M3", "M2 to M4", "M1's flags fill it") agree with the plan order. No ordering conflict found.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity | AC-CMM-008 prose says "an English month name and a day number" (acceptance.md:L44) while the command covers only capitalized month-first forms (N2); the acceptance.md:L7 sentence about the file difference is stale at HEAD (O1). |
| Completeness | 1.0 | all sections | HISTORY spec.md:L19-L22, WHY L26, deliverable L30, REQ L60, Constraints L78, `### Out of Scope — ...` x4 spec.md:L87-L104, frontmatter complete. |
| Testability | 0.75 | one non-discriminating command, one unsatisfiable guard | E08 admits lowercase and day-first dates (N2); E10a cannot reach exit 0 at the run-phase tip (N1). All other commands were observed red on mutants (below). |
| Traceability | 1.0 | complete | Verb result: `COLLECTED: 11 REQ definitions (acceptance input: read)`, no `UNCOVERED:` line, no `ORPHAN:` line. Table at acceptance.md:L77-L89 maps each REQ to at least one AC; the incidental-looking mappings (REQ-CMM-004 to AC-CMM-011, REQ-CMM-009 to AC-CMM-013) are real reviewer procedures with a named evidence record (progress.md section E.2). |

Aggregate: (0.75 + 1.0 + 0.75 + 1.0) / 4 = 0.875.

## Defects Found (structured defect-list)

N1. acceptance.md E10a/E10b + AC-CMM-010 + Definition of Done — acceptance.md:L52-L53, L178-L186, L236 — E10a exits 1 on the current tree, and will exit 1 at the run-phase tip whatever the run phase does. Cause: commit `b50a31b1a` force-added `.moai/reports/t1760/plan-audit.md` and `.moai/reports/t1760/progress.md` (`git ls-files .moai/reports/t1760` lists both; `git diff --name-only 629bb51... HEAD` lists them beside the four SPEC files). A tracked file is not hidden by the ignore rule at `.gitignore:235`, so the statement at acceptance.md:L53 ("Card evidence under `.moai/reports/` registers at neither") is false for any evidence that was force-added, and the Definition of Done clause "E10a exits 0 ... before the sync phase begins" (L236) cannot be met. This is the "impossible / wrong-reason red" direction of verification-completeness.md section 2: red at arrival, red forever, for files the work never touches. The ledger line "green (current tree): exit 0" (L182) was measured at `cf5a52f55`, before those files existed, and is stale at HEAD. The independent codex leg reported the same defect (P2, confidence high). — Severity: major — Class: blocking — Required fix: add `':!.moai/reports'` to the pathspec list of both E10a and E10b (and to the probe descriptions), re-observe and re-record the green exit and the two probe results on the current tree, and replace the "registers at neither" sentence with a statement that is true for tracked card evidence. Measured here: with that exclusion the same command exits 0 (`E10a with reports excluded exit=0`).

N2. acceptance.md E08 / AC-CMM-008 — acceptance.md:L44-L45, L164, L227 — The command misses dates the criterion says it forbids. Observed on single-line inputs: `october 16` -> 0, `OCTOBER 16` -> 0, `16 October` -> 0, while `Oct 16` -> 1. The month alternation is capitalized and month-first only, and the AC text promises "an English month name and a day number" without those qualifiers. A document line `16 October` satisfies the command and violates REQ-CMM-010; this is a writable mutant (verification-completeness.md section 2, mutant probe). The independent codex leg reported the same two shapes (P2, confidence high). — Severity: major — Class: blocking — Required fix: either extend E08 (case-insensitive month names, e.g. `[Oo][Cc][Tt]...` classes without making the `SPEC-` branch case-insensitive, plus a day-first branch `[0-9]{1,2}[ .]+(month)`) and add the new shapes to the must-hit list and the 25-line must-stay-zero list with re-observed counts, or narrow the AC-CMM-008 wording to exactly what the command checks and record the uncovered shapes under "Known limit" (acceptance.md:L231). Do not leave the prose broader than the command.

O1. (optional) acceptance.md:L7 and L193-L196 — Stale measurement statements at HEAD: "`git diff --name-only 629bb51... HEAD` listed `acceptance.md`, `plan.md`, `progress.md`, `spec.md`" is true for `cf5a52f55` (verified: that diff lists exactly those four) and false for the `HEAD` it names (it now lists six files, two under `.moai/reports/t1760`). Resolve together with N1 by saying `cf5a52f55` where that is the tree meant. — Severity: minor — Class: optional.

O2. (optional) acceptance.md:L231 — Known-limit paragraph lists full-width digits and HTML entities, but not other date spellings: `Oct16`, `10.16`, and `16 Oct` printed no match in a four-line probe (`16 Oct`, `Oct16`, `10.16`, `16 October`). If N2 is fixed by extending the command, extend these too or list them. — Severity: minor — Class: optional.

O3. (optional) plan.md:L73-L76 — Milestones are a numbered list, so the CN-4 verb returns GAP and the milestone order cannot be observed mechanically. Writing them as `### M1 ...` headings with an `Exit:` line naming the criteria they flip would let the verb observe the order and let the Definition of Done clause at acceptance.md:L236 be checked rather than hand-read. — Severity: minor — Class: optional.

O4. (optional) acceptance.md:L28-L29 (AC-CMM-004) — The verdict is "two numbers are equal and each is 4 or more", read across two separate commands (E04a, E04b), so no single exit code carries it. Acceptable for a reviewer-run comparison; a single awk call that prints a mismatch would make it a verdict. — Severity: minor — Class: optional.

## Regression Check (iteration 2)

- D1 (round 1): RESOLVED — see table above; 24/24 must-hit, 0/25 must-stay-zero, 0 on the Korean passing fixture.
- D2 (round 1): RESOLVED — guard admitted the command; five fixture runs gave the expected discriminating outputs.
- Round-1 D3 (E03): E03 now rejects duplicate ids and short fragments: four BP-1 rows only gave `COUNT BP-1=2 / COUNT BP-2=0`; a 10-character fragment gave `BAD-FRAGMENT`; four rows with flag `MAYBE` gave four `BAD-FLAG` lines. Passing fixture exit 0.
- Round-1 D4 (E05): rejects a swapped header (`HEADER-COUNT=0`), Fit `maybe` (`BAD-FIT` x3), a defense cell with no `§` (`NO-PATH-SECTION`), and a missing theme (`THEME-MISSING: merge-queue-alternative`). Passing fixture exit 0.
- Round-1 D5 (E10): see N1 (open, new reason).
- Round-1 D6/D7/D8/D9/D10/D11: acceptance.md AC-CMM-006 is conditional (L36), the DoD names the phase (L236), the ledger separates stdout/stderr/exit (L93), plan.md section D carries the citation note, section B.2 carries the measured Hangul line counts, E08 is case-sensitive on `SPEC-` and does not count `spec-anchored` (zero on that input). Not re-opened in depth beyond these spot checks.

## Recommendation

1. Fix N1: add `':!.moai/reports'` to E10a and E10b, re-observe the green state and both probes at the current HEAD, correct the "registers at neither" sentence and the stale `HEAD` sentence (O1).
2. Fix N2: close or honestly scope the E08 date shapes (lowercase month, day-first); keep the `SPEC-` branch case-sensitive; extend the must-hit and must-stay-zero lists with re-observed counts.
3. Optional: O2, O3, O4.
4. The re-audit can be scoped to the fix_scope anchors above, with the CN-4 ordering verb re-run in full (it is exempt from delta scope).

## Evidence (five-section format)

### Claim

C1. Round-1 D1 is fixed: E08 now counts Hangul-attached forbidden tokens.
C2. Round-1 D2 is fixed: E07 runs under the worktree guard and counts distinct paths.
C3. Every release-blocking RED-now cell reproduces on the current tree.
C4. Two new defects are open: E10a is unsatisfiable (N1) and E08 misses lowercase/day-first dates (N2).
C5. Must-pass criteria MP-1 to MP-9 hold; the verdict is FAIL because of the required codex gate and two blocking findings.

### Evidence

All commands were run from `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1760`, in the audit's own shell. Fixtures live in the session scratchpad `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/39eb45b2-9233-459e-a4d1-94f379f1058c/scratchpad/` (not committed).

RED-now reproduction (C3):
```
$ test -f .moai/docs/multi-agent-containment-mapping.md
exit=1
$ grep -c -i -F 'reference-only' .moai/docs/multi-agent-containment-mapping.md
ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
E02 exit=2        (E04a, E04b, E06, E08, E09: same warning, exit 2)
$ awk -F'|' -v rowre=... (E03 verbatim from acceptance.md:L113) .moai/docs/multi-agent-containment-mapping.md
awk: can't open file .moai/docs/multi-agent-containment-mapping.md
 source line number 1
exit 2            (E05, E07: same message, exit 2)
```

E08 must-hit (C1), the exact E08 command from acceptance.md:L164 against `e08_hit.txt` (24 lines: the 19 ledger examples plus `Sept. 9`, `Dec 5에`, `카드t1760`, `한글2026년`, `t100을`):
```
24
hit exit=0
```
E08 must-stay-zero, against `e08_zero.txt` (25 lines, includes `developer를`, `30,000건`, `2,500건`, `v3.2.0은`, `spec-anchored`, `Opus 5.5`, `prompting-claude-opus-5-5`, the two quoted source figures):
```
0
zero exit=1
```
E08 against the Korean passing fixture `f_pass.md`: `0`, exit 1.

E08 date probe (N2), single-line inputs:
```
october 16 -> 0
OCTOBER 16 -> 0
16 October -> 0
Oct 16 -> 1
```
and a four-line file (`16 Oct`, `Oct16`, `10.16`, `16 October`): `grep -n` printed nothing, exit 1.

E07 (C2), the exact command from acceptance.md:L155, run on fixtures:
```
== f_pass            exit=0                                   (nine distinct existing .md paths cited as `path` § section)
== e07_dup           TOO-FEW-DISTINCT n=1                     exit=1   (one path cited eight times)
== e07_suffix        BAD-SUFFIX: .claude/rules/moai/core/moai-constitution.md.NONEXISTENT
                     TOO-FEW-DISTINCT n=7                     exit=1
== e07_missing       MISSING: .claude/rules/moai/core/does-not-exist.md   exit=1
== e07_seven         TOO-FEW-DISTINCT n=7                     exit=1
```
E03 and E05 mutants: see Regression Check above (outputs `COUNT BP-1=2`, `COUNT BP-2=0`, `BAD-FRAGMENT`, `BAD-FLAG` x4, `BAD-FIT` x3, `NO-PATH-SECTION`, `THEME-MISSING: merge-queue-alternative`, `HEADER-COUNT=0`; passing fixture exit 0 for both). One E03 mutant (`0,/re/` sed form) produced no change under the BSD sed in this environment and exited 0; it was an auditor fixture error, redone with a plain `s///` that produced four `BAD-FLAG` lines and exit 1.

E02/E04/E06/E09 on the passing fixture: `1`, `4`, `4`, `1`, `0` respectively (reference-only count, URL lines, URL lines ending in a flag, gap rows, rule-marker lines).

E10a/E10b (N1):
```
$ git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001'
E10a exit=1
$ git status --porcelain --untracked-files=all -- . ':!.moai/docs/...' ':!.moai/specs/SPEC-...'
E10b exit=0   (stdout empty)
$ git diff --name-only 629bb51cda73fc70422769988b8c28182c9bd2d5 HEAD
.moai/reports/t1760/plan-audit.md
.moai/reports/t1760/progress.md
.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/acceptance.md
.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/plan.md
.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/progress.md
.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/spec.md
$ git ls-files .moai/reports/t1760
.moai/reports/t1760/plan-audit.md
.moai/reports/t1760/progress.md
$ git diff --quiet 629bb51... -- . ':!...mapping.md' ':!...SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
E10a with reports excluded exit=0
$ git diff --name-only 629bb51cda73fc70422769988b8c28182c9bd2d5 cf5a52f55992e59fe9593fa508ace5d4dba59fc7
(the four SPEC files only)
```
This round's own verdict file is gitignored (`git check-ignore -v` -> `.gitignore:235:.moai/reports/*`) and untracked; if the lane force-adds it as it did round 1, it falls under the same N1 exclusion.

Traceability verb (full output): `COLLECTED: 11 REQ definitions (acceptance input: read)`; no `UNCOVERED:` or `ORPHAN:` line.
CN-4 verb: `COLLECTED: 0 milestones in plan order (none), 0 exit bindings, 7 ordering candidates` / `GAP: 0 milestone headings collected from the plan — milestone order not observed`.
`mcp__moai__spec_audit` (project_root set): `total_specs: 1, modern_era_clean: 1`, one INFO `EraAutoDetected` (H-5).

Audit plan and convergence: `moai verify audit-plan --project-root <toplevel>` reported `model: codex`, gates claude=off, codex=required, glm=off, `cross_model_active: true`, `enforced_required: ["codex"]`. `mcp__moai__audit_multi` (no `gates` argument, `project_root` = this toplevel, `target: baseBranch`) returned `overall_verdict: fail`, `per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]`, `plan_source: config`, `audit_receipt: rcpt-f822b71fe8a50a2376562518`, with two P2 findings that match N1 (E10a, confidence high, exit 1 at HEAD `1b2319571`) and N2 (E08, `october 16`, `OCTOBER 16`, `16 October` all stdout 0). I had reached both findings independently before reading the codex output (N1 by `git diff` of the pinned base; N2 by the single-line probes above). Digest written to `.moai/state/audit-plan-result.json` and checked with `--result-file`: `convergence_check: {"ok": true, "unmet": []}` (the required backend answered; its verdict was `fail`, which is a reviewed defect, not an unmet gate).

### Baseline-attribution

- Tree: HEAD `1b2319571070c4d3b60d270037d2f90f7d7ffa8d`, branch `WT-10-15-release-3`, `git status --porcelain` empty at the start of the audit; pinned comparison base `629bb51cda73fc70422769988b8c28182c9bd2d5`; ledger pin `cf5a52f55992e59fe9593fa508ace5d4dba59fc7` (document absent on all three).
- Judging build: `moai` MCP server and CLI `v3.2.0-rc.36-t1822c`, commit `ee981a932` (server banner); `git merge-base --is-ancestor ee981a932 HEAD` printed exit 1, so that commit is not an ancestor of the tree's HEAD, and the build is not a strict ancestor lagging the tree (verification-claim-integrity.md section 2.2). `grep` in this shell is ugrep 7.8.4; `awk` is `/usr/bin/awk` (BSD one-true-awk; `awk -W version` is not supported, so the build string was not observed).
- Measured in this run, against this tree; nothing carried over from the round-1 report except the round-1 defect text.

### Gaps (not observed)

G1. No tool was refused by the worktree guard in this audit: every acceptance command and verb was admitted and executed. (The round-1 refusals of the old E07 and of the CN-4 verb were not re-run.)
G2. The final document does not exist, so E01-E09 were observed RED on the absent document and on constructed fixtures only, never green on a real deliverable. The fixtures are auditor-written and live in session scratch (not committed); they approximate, not reproduce, the author's fixtures.
G3. The CN-4 verb could not observe the milestone order (`GAP:` line) because plan milestones are list items; MP-9 rests on a hand read of plan.md:L73-L76 and the seven candidates.
G4. Source-page claims (the four URLs, the quoted fragments, the VERIFIED flags in spec.md L46-L49 and plan.md section C) were not re-fetched; they are not plan-phase checkable without the network and belong to the run phase (AC-CMM-011).
G5. `plan_artifact_hash` is an auditor-computed sha256 of the four SPEC markdown files, not the runtime `ComputeHash` over the tier's artifact set; the admission rule may compute a different value.
G6. Round-1 defects D3-D11 were spot-checked only (see Regression Check), not re-measured in depth. The helper `spec-lint` verb (`moai spec lint`) was not re-run in this round.
G7. The `claude` and `glm` audit backends are `off` in the plan and were not called; only the required codex leg ran.
G8. awk and ugrep behaviour was observed on this machine only (macOS, ugrep 7.8.4, BSD awk); a CI or Linux runner with GNU awk/grep was not observed. Interval syntax `{20,}` (E03) and `{3,4}` (E08) worked here.

### Residual-risk

- Even with N1 and N2 fixed, E08 counts lines and enumerates date shapes; other spellings (`Oct16`, `10.16`, HTML entities, full-width digits) remain open (O2).
- AC-CMM-003/005/007 prove well-formedness and path existence, not that a quoted fragment is on the source page or that a cited section was read; those are reviewer-checked (AC-CMM-011, AC-CMM-014) and unobserved here.
- The E07 distinct-path floor is 8 against a seed set of 9 existing paths (plan.md section D), so one dropped row after M3 re-reading leaves a margin of one.
- Tracked card evidence (N1) will keep accumulating as the card proceeds; any other tracked path outside the allowed pair also turns E10a red, which is intended, but the `.moai/reports` exclusion deliberately stops E10a from seeing evidence tampering.

## Operational Notes (unverified)

- Measure the fix: re-run `git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'` after editing and record its exit code. Status: measured here (exit 0 on this HEAD).
- Measure the E08 fix: run the revised E08 over the single-line set `october 16`, `OCTOBER 16`, `16 October`, `Oct16`, `10.16` and expect a non-zero count for each, then over `e08_zero`-style lines and expect 0. Status: assumption (not yet run on a revised command).

AUDIT-VERDICT: FAIL spec=SPEC-CONTAINMENT-MAPPING-DOC-001 receipts=rcpt-f822b71fe8a50a2376562518
