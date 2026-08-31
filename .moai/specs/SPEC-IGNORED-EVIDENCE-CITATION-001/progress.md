# Progress — SPEC-IGNORED-EVIDENCE-CITATION-001

Card t381 · worktree `.claude/worktrees/t381` · branch `WT-ignored-evidence-cite` · base
`origin/develop` `3f03d9c36`.

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: M (spec.md + plan.md + acceptance.md + progress.md). Reasoning in the completion report;
  the diff is Tier-S-sized (5 lines) but the card carries a genuine per-file design decision, a
  cross-card canon coupling, and an AC set whose majority are measured negatives — those belong in a
  verification layer, not an inline block.
- **Artifacts**: spec.md, plan.md, acceptance.md, progress.md — created 2026-08-31, `status: draft`.
- **SPEC ID check**: `SPEC-IGNORED-EVIDENCE-CITATION-001` — Bash regex `PASS`; no existing SPEC
  directory or grep hit for the id.
- **Scope source**: `.moai/reports/t381/census.md` class C3, measured tree `3f03d9c36`; probe
  reproduced in this tree at 25 lines.
- **Requirements**: REQ-IEC-001..010 (written at iter1 as `001..009`; the iter4 renumber closed the
  gap the iter3 demotion opened — spec.md HISTORY 0.4.0 P1). **Acceptance**: ids `AC-IEC-001..007`,
  `010..012` — **ten MUST criteria, not a contiguous range**: former `AC-IEC-008` and `AC-IEC-009`
  were demoted to the §D structural checks S-1 and S-2 at iter2 (D3), so the ids are absent from §C
  by design. Each MUST criterion is decided by a named command.
- **Open question carried into run-phase**: does `SPEC-EVIDENCE-CITATION-CANON-001` REQ-ECC-004 reach
  an output-location statement (`internal/cli/audit_pin_live_test.go`)? Recorded at spec.md §A.4;
  plan.md §F M1 resolves or escalates it, with a conservative default if unresolved.
- ~~**Known limitation**: every REQ-ECC citation is to an unmerged t375 draft (v0.3.0). The
  requirement text was read directly from `.claude/worktrees/t375`; the relayed HEAD SHA `b64043481`
  was **not** independently verified here (the worktree-isolation guard refuses cross-tree
  `git -C`).~~ **Superseded at iter3** — t375 landed; every REQ-ECC citation now names tracked
  history in `origin/develop` `9328a5242`, read via `git show`. See the iter3 revision entry below.
  *(Swept in the run-phase F1-F3 pass: the same half-repair shape as F1 — the fact was updated in the
  iter3 section and left standing in the bullet that depends on it. Not raised by any audit
  iteration; found by reading each hit for meaning.)*

- **spec-lint: PASS, and the check is shown to be live.** Targeted run naming the file on argv —
  `moai spec lint .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md` → `✓ No findings`, rc=0.
  A mutation (renaming the five `### Out of Scope — ` headings in a scratch copy) made
  `MissingExclusions` fire, so the clean result is not a vacuous green. An earlier
  `moai spec lint | grep …` returning `rc=1` was **discarded** as evidence: the exit code came from
  `echo`, and the linter's output was never seen. Evidence:
  `.moai/reports/t381/verify/spec-lint-verification.md`.
- **Finding surfaced by the mutation, OUT OF SCOPE, reported to the lead — era misclassification of
  every plan-phase SPEC.** The mutant message carried `[grandfathered era — downgraded to warning]`.
  `moai spec audit` reports **714 SPECs / 286 grandfathered / 24 in V3R5**, and classifies this SPEC
  as **V3R5** via `H-3 (§E.2 present, sync_commit_sha missing)`. In `internal/spec/era.go` the
  heuristics return on first match and H-3 precedes H-5, so a plan-phase SPEC — which has §E.2 by
  construction and no `sync_commit_sha` until sync runs — can never reach H-5's modern-date
  tie-breaker, even when created today.

  **The consequence is narrower than iter1 stated.** iter1 wrote "lint errors are demoted to
  warnings", which is too broad. Read at `internal/spec/lint.go:272-275`, `eraDemotableCodes` contains
  exactly **two** codes — `MissingExclusions` and `FrontmatterInvalid`; a planted `status:` mutant
  kept `StatusValueInvalid` at ERROR. The broader consequence is at `lint.go:296-311`: on a protected
  SPEC every remaining warning is marked `Advisory`, so **`--strict` cannot escalate any of them**.
  Correction relayed by the lead and verified here by reading the source.

  Not acted on — it is a lint/era-engine concern requiring an `internal/spec/era.go` edit, unrelated
  to this card's 5 citation lines.

### iter2 revision (plan-audit iter1 = FAIL 0.74, Tier M threshold 0.80)

All seven must-pass criteria passed at iter1; the failure was Testability 0.55. Changes:

- **D1 (critical, fixed)** — all seven non-executable acceptance commands rewritten as plain single
  invocations. Every command in `acceptance.md` §C was **executed in this worktree** during iter2 and
  its observed output recorded beside it. No `for`, no `$(…)`, no subshell, no script file.
- **D2 (major, fixed)** — AC-IEC-001 no longer decides by reading a `grep -C2` window. It is now a
  binary `grep -L` (files lacking a marker; PASS = empty output), which is **RED today** on four
  files.
- **D6 (major, fixed)** — added two requirements, numbered **REQ-IEC-010** (stale evidence
  coordinates, Unwanted form) and **REQ-IEC-011** (behavior preservation) **at the time**; the iter4
  renumber moved them to **REQ-IEC-009** and **REQ-IEC-010**. AC-IEC-012 and AC-IEC-011 now cite
  requirements instead of document sections. Former AC-IEC-009 demoted to a §D structural check.
- **D5 (blocking-minor, fixed)** — the phantom `.codex/agents/moai/manager-lead.toml` corrected to
  `internal/template/templates/.codex/agents/moai/manager-lead.toml` in spec.md §C.5 and AC-IEC-007.
  Verified: `git ls-files '*manager-lead.toml'` returns that one path.
- **D4 (blocking-minor, fixed)** — t375's `status:` corrected to `in-progress`, pinned to the read
  date 2026-08-31, in both places.
- **D3 (optional, accepted)** — former AC-IEC-008/009 demoted to §D structural checks; AC-IEC-010
  re-scoped to assert the evidence **M5 actually writes**, making it RED until M5 runs.
- **D7 / D8 / D9 (optional, accepted)** — `492 − 25 = 467` shown with its command; the lane-8
  convergence claim relabelled **relayed** (per the lead's 2026-08-31 message) with the separately
  verified part named; the brace-glob exemption re-argued against the REQ-ECC-004+005 pair.
- **New finding, this session** — the in-scope pattern is narrowed to `.moai/state/verify`. The html
  file carries **two** `.moai/state/` occurrences and line 529 (`loop-verdict`, census B2) is out of
  scope; iter1's broader pattern would have swept it in. Recorded at spec.md §C.1 and acceptance.md §A.

### iter3 revision (plan-audit iter2 = FAIL 0.78, operator-approved delta-scoped iter3)

Monotonic 0.74 → 0.78, no dimension regressed. Scope limited to N1 + N2 by operator override; the
sibling-lane hazard (an iter2 that converted an omission into an assertion and scored 0.85 → 0.83)
is the reason nothing else was widened.

- **N1 (critical, closed)** — AC-IEC-001 could never go green. `grep -L` cannot see whether a citation
  is present, so `mcp_glm.go` stayed in its output after a *correct* treatment-(a) repair (measured:
  `grep -ciE '<markers>' internal/cli/mcp_glm.go` → `0`), and `plan.md` gated four of five repairs on
  it. Replaced with `grep -l '\.moai/state/verify' <5 files> | xargs grep -LiE '<markers>'` — the
  obligation is now conditional on a citation still being present. Observed **RED** today, listing
  exactly the four files that carry a citation with no marker. Both edge cases were run here before
  adoption: all-pass → empty output `exit=0`; empty input → no hang, no `(standard input)` false
  positive (and unreachable in practice, since `extract.txt` keeps its citation).
- **N2 (major, closed — both axes)** — the criterion was green today, and its deciding command could
  not discriminate: `git status --short` collapses an untracked directory to one `??` line regardless
  of contents. Replaced with an `ls` over the ten named per-criterion evidence files
  (`ac-iec-001.txt` … `ac-iec-012.txt`); observed `exit=1` today — genuinely RED, green only once M5
  writes them. `plan.md` M5 now fixes those filenames so the criterion and the milestone agree. The
  matrix RED-now cell is corrected from "(dir not yet populated)" — the directory in fact held three
  lint-verification files.
- **N3 (checked, survived, closed as instructed)** — it did **not** resolve as a by-product: N1 touches
  AC-IEC-001 (REQ-IEC-001) and N2 touches AC-IEC-010 (REQ-IEC-007); neither creates coverage for
  the probe-boundary requirement (**numbered `REQ-IEC-008` at the time**; that id was reassigned to
  Collision avoidance by the iter4 renumber), which then appeared only in `spec.md`. Closed by
  **demoting the requirement** to a §D constraint — not by re-promoting the demoted check, which
  would have restored the vacuity iter2 removed. ~~REQ ids 009/010/011 keep their numbers so every AC
  cross-reference stays valid.~~ **Superseded at iter4 (P1)**: that ground did not survive
  measurement — `git grep 'REQ-IEC-' origin/develop -- . ':!.moai/specs'` returns nothing, so no
  external consumer existed to protect, and the gap cost the MP-1 firewall. 009/010/011 were
  renumbered to 008/009/010; `REQ-IEC-011` no longer exists.
- **Baseline moved — t375 landed.** `SPEC-EVIDENCE-CITATION-CANON-001` is `status: completed` in
  `origin/develop` `9328a5242` (read 2026-08-31 via `git show origin/develop:…`; `develop` was NOT
  merged into this branch — absorption belongs to the integration window). Every citation updated from
  unmerged-draft framing to a landed citation. Two upgrades fell out of the re-read: the rule-body
  sentence is now **verified** at `manager-lead.md:150` rather than anticipated, and the lane-8
  convergence claim is now **verified** at `manager-lead.md:152` rather than relayed — the landed text
  keeps "MUST resolve at audit time" and re-points it at the tracked path, which is the same split
  this card selected as treatment (d). `status:` has now held three values across this card
  (draft → in-progress → completed), which is why the citation pins a read date.
- **Three iter2 figures corrected to their measured values** — `…html:2` → **1** (`grep -c` counts
  lines; both matches are on line 684); `ls … rc=2` → **`exit=1`** (BSD `ls`); and AC-IEC-010's
  "RED until M5" → green-then, RED-now under the new command.
- **N4 (done, flagged as a judgment call)** — `spec.md` and `plan.md` bumped `0.1.0` → `0.3.0`,
  `acceptance.md` `0.2.0` → `0.3.0`, with iter2 and iter3 HISTORY rows. Strictly this is a Completeness
  fix rather than N1/N2, but iter3 makes substantive edits to all three files and leaving them
  unrecorded would have introduced a *new* provenance inaccuracy on a card about provenance. Flagged
  so the coordinator can reject it.
- **Kept open by lead decision** — the REQ-ECC-004 boundary question (does 004 reach a write-location
  statement?). Now that t375 has landed it has an **owner to ask** rather than a moving document to
  wait on; `plan.md` §F M1 keeps the conservative default (correct only the false assertion).

~~_Awaiting plan-audit iter3 and Implementation Kickoff Approval._~~ **Superseded**: plan-audit iter3
returned FAIL-on-firewall (score 0.83 cleared) and iter4 returned **PASS-WITH-DEBT 0.89** (Tier M
threshold 0.80), monotonic 0.74 → 0.78 → 0.83 → 0.89 with no dimension regressed
(`.moai/reports/t381/plan-audit-iter4.md`). Implementation Kickoff Approval granted by the operator;
run-phase entered. *(Also swept in the F1-F3 pass — same half-repair shape.)*

### M1 — settled repair direction (run-phase, recorded before any target file was edited)

Each of the five treatments was confirmed against the file **as it actually reads in this tree**
(`3f03d9c36`), not against the dispatch's framing. Two of the five confirmations changed something.

| File | Treatment | Confirmed against the file by |
|---|---|---|
| `internal/cli/mcp_glm.go` | **(a)** delete the path, keep the `t225` token | read `:98-115` — all five figures (3667, 3480, 3072, 1024, 1.02) are in the comment body, so the path carries nothing the reader loses |
| `internal/cli/audit_pin_live_test.go` | **(d)** correct the assertion, keep the output-location statement | read `:31-33` — the sentence is two clauses with different truth values, exactly as §C.3 describes |
| `internal/hook/evidence_writer_zeroexec_test.go` | **(b)** demote to an explicitly-non-resolving origin note | read `:1-11` — the six runner versions are in the body; the pointer is provenance for the **raw** captures only |
| `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt` | correct the stale coordinate only | read `:1-11` — the header already declares its own non-resolution; only `.gitignore:284` is live |
| `.moai/reports/template-skill-improvement-plan-20260710.html` | **(b)** with a recorded inability — **NOT (c)** | see the discriminator correction below |

**Open question — carried forward unresolved, conservative default applied.** Does REQ-ECC-004 (a
citation shall name a single file, not a directory) reach an **output-location statement** such as
`audit_pin_live_test.go:32`? This lane cannot ask the lead (subagent boundary — `AskUserQuestion` is
the orchestrator's channel), so the default of `plan.md` §F M1 step 2 stands: treat the sentence as
an output-location statement outside REQ-ECC-004's reach and repair **only** the false assertion
(treatment (d)) — the narrower change, and the one lane-8 independently converged on
(`manager-lead.md:152`). Recorded as an unresolved carry-forward per `acceptance.md` §F.

**Discriminator correction — `plan.md` M3 branches on the wrong fact.** M3 says: *if the `eb01063e`
scratch tree exists on this machine → (c), export the one JSON that decided the claim; if not →
(b)*. Measured in the primary checkout:

```
$ ls -d /Users/goos/MoAI/moai-adk-go/.moai/state/verify/eb01063e   → exists (Jul 10 02:22)
$ find …/eb01063e/skill-audit -type f | wc -l                      → 8   (244K total)
```

The tree **exists**, so M3's stated branch selects (c) — but (c) is unavailable, because its own
precondition fails for a different reason. The citation is the report's **report-wide raw-data
footer**, and the eight JSONs are eight per-category batch audits (`domain.json`,
`workflow-testing.json`, `ref-harness.json`, …). No single one decides the footer's claim: the two
D-grade findings alone span two files, and the "B 이상 77%" figure aggregates all eight. Exporting
all eight instead is prohibited outright — REQ-ECC-004, "whole-scratch-directory export is
prohibited".

`spec.md` §C.3 states the discriminator correctly — *"(c) **if the file can be identified**,
otherwise (b) with a recorded inability"* — so the SPEC is right and **`plan.md` M3 is the artifact
that is wrong**: it restated an identifiability test as a tree-existence test, and the two disagree
in exactly this case. Treatment follows §C.3: **(b)**, with the inability recorded at the citation
site per REQ-IEC-005.

One consequence for the requirement text, reported rather than edited (body edits are manager-spec's
to make): **REQ-IEC-005's** parenthetical ties the inability to a single cause — *"because the
scratch tree no longer exists on this machine"*. The inability found here has a **different** cause
(the claim is genuinely multi-source), and writing the requirement's stated cause into the file
would have been a false statement. The requirement's normative core — record the inability, never
silently retain the glob — applies unchanged and is what the repair follows.

## §E.2 Run-phase Evidence

Every command below was executed **in this worktree**, at HEAD `3f03d9c36`, merge base
`origin/develop...HEAD` = `3f03d9c369` (re-read immediately before the second pass). Each row's
verbatim output is persisted at `.moai/reports/t381/verify/ac-iec-<NNN>.txt` — a **tracked** path,
never `.moai/state/` (REQ-IEC-007; the card would be self-refuting otherwise).

### AC matrix

| AC | Command | Actual output | Status |
|---|---|---|---|
| AC-IEC-001 | `grep -l '\.moai/state/verify' <5 files> \| xargs grep -LiE '<markers>'` | *(empty)*, `exit=0` | **PASS** |
| AC-IEC-001 companion | `git grep -c '\.moai/state/verify' -- <5 files>` | `extract.txt:1`, `…html:1`, `audit_pin_live_test.go:1`, `evidence_writer_zeroexec_test.go:1` — `mcp_glm.go` absent (treatment (a) dropped it from the listing) | **PASS** (each ≤ 1) |
| AC-IEC-002 | `grep -c 'still resolve at audit time' internal/cli/audit_pin_live_test.go` / `grep -c 'Evidence lands in' …` | `0` (`exit=1`) / `1` (`exit=0`) | **PASS** (assertion gone, location statement survives) |
| AC-IEC-003 | `grep -oE '3667\|3480\|3072\|1024\|1\.02' internal/cli/mcp_glm.go \| sort -u \| wc -l` | `5` | **PASS** (regression guard held — no figure lost with the path) |
| AC-IEC-004 | `grep -cE '\.moai/reports/\|\.moai/state/verify' <2 files>` | `evidence_writer_zeroexec_test.go:1`, `…html:1` | **PASS** (each ≥ 1 — no origin deleted outright) |
| AC-IEC-005 | `grep -c 'skill-audit/\*' …html` / `grep -ciE 'not exported\|반출되지\|cannot be named\|식별 불가' …html` | `1` / `1` | **PASS** via the second branch (glob retained **with** an explicit inability marker — never silently) |
| AC-IEC-006 | `git diff --exit-code --stat origin/develop...HEAD -- <8 carve-out files>` <!-- moving-ref-ok: the moving ref is the SUBJECT of the criterion, not its anchor. The three-dot form re-resolves merge-base(origin/develop, HEAD) at verification time; that is the iter4 P3 fix (acceptance.md §A). Pinning a SHA here reinstates the exact defect P3 removed — after absorbing origin/develop the frozen base reports t375's landed edits as this card's. Anchor: HEAD 3f03d9c36, merge base 3f03d9c369, run 2026-08-31. --> | *(empty)*, `exit=0` | **PASS** |
| AC-IEC-006 positive control | `git grep -c '\.moai/state/verify' -- <carve-out paths>` | `1+1+3+1+1+1+3+1` = **12** | **PASS** (unchanged is not satisfied by deletion) |
| AC-IEC-007 | `git diff --exit-code --stat origin/develop...HEAD -- <**8** t375-owned paths — the guard file folded in by the iter5 correction>` <!-- moving-ref-ok: same as AC-IEC-006 — the three-dot merge-base resolution is the criterion's subject and must stay moving (iter4 P3). Anchor: HEAD 0e903d464, merge base 3f03d9c369, re-run 2026-08-31 against origin/develop 59e898b31. --> | *(empty)*, `exit=0` | **PASS** |
| AC-IEC-007 (post-absorption, verified WITHOUT absorbing) | `git merge-tree --write-tree origin/develop HEAD` → `2acdfbf88`, then `git diff --exit-code --stat origin/develop 2acdfbf88 -- <same 8 paths>` | *(empty)*, `exit=0` | **PASS** — the case the retired `ls` form would have FAILED |
| AC-IEC-007 (non-vacuity) | `git ls-tree 2acdfbf88 -- internal/template/evidence_citation_guard_test.go` | `100644 blob 9b1970fe32f0…` | **PASS is real** — present in the merged tree, absent from this card's diff |
| AC-IEC-007 (RED-capability) | same three-dot form over a path this card **did** change (`spec.md`) | `1 file changed, 594 insertions(+)`, `exit=1` | **can fail** — not a vacuously-empty command |
| ~~AC-IEC-007 (guard file, RETIRED iter5)~~ | ~~`ls internal/template/evidence_citation_guard_test.go`~~ | ~~`No such file or directory`, `exit=1`~~ | **SUPERSEDED** — asserted tree state, not this card's diff; see §E.2 record 1 |
| AC-IEC-010 | `ls <ten named evidence files>` / `git check-ignore -v .moai/reports/t381/verify` / `ls .moai/state/verify` | all ten listed, `exit=0` / `exit=1` (**not** ignored) / `No such file or directory` | **PASS** (all three sub-checks) |
| AC-IEC-011 | `go build ./internal/cli/... ./internal/hook/...` | `exit=0` | **PASS** |
| AC-IEC-011 | `go test ./internal/cli/... ./internal/hook/...` | 28 packages `ok` (`internal/cli` 473.9s, `internal/hook` 83.7s), `exit=0`, `grep -c FAIL` = `0` | **PASS** |
| AC-IEC-012 | `grep -c 'gitignore:284' extract.txt` / `grep -ci 'gitignore' extract.txt` | `0` (`exit=1`) / `2` | **PASS** (stale coordinate gone, non-resolution statement retained) |

**Ten MUST criteria, ten PASS, zero FAIL, zero PASS-WITH-DEBT.**

### Invariant rows

| Invariant | Command | Actual output | Status |
|---|---|---|---|
| Carve-outs untouched (REQ-IEC-006) | AC-IEC-006 both halves | `exit=0` + sum 12 | **HELD** |
| t375-owned files untouched (REQ-IEC-008) | AC-IEC-007 both halves | `exit=0` + file absent | **HELD** |
| No behavior change (REQ-IEC-010) | AC-IEC-011 + `git diff` read | every hunk is a comment / header / footer; no Go statement, no assertion, no runtime path | **HELD** |
| Diff confined to the declared scope | `git diff --stat` | exactly the 5 in-scope files, `18 insertions(+), 7 deletions(-)` | **HELD** |
| Nothing written under `.moai/state/` | `ls .moai/state/verify` | `No such file or directory` | **HELD** |
| Probe unchanged since census | `git grep -n '\.moai/state/verify' -- . ':!*.md' \| wc -l` | `25` (census figure, re-measured at run-phase entry) | **HELD** |
| Ignore rule at the census coordinate | `git check-ignore -v .moai/state/verify/t225` | `.gitignore:298:.moai/state/` | **HELD** |
| §D S-1 (probe + blind spots pinned) | `grep -c "<probe>" spec.md` / `grep -cE '^\| B[1-6] \|' spec.md` | `1` / `6` | **HELD** (recorded, non-gating) |
| §D S-2 (C4 exclusion stated with reason) | `grep -c '^### Out of Scope' / 'class C4' / 're-produced' spec.md` | `5` / `1` / `1` | **HELD** (recorded, non-gating) |
| spec-lint (this SPEC) | `./bin/moai spec lint .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md` | `✓ No findings`, `exit=0` | **HELD** |
| spec-lint (repo-wide baseline) | `./bin/moai spec lint` | `0 error(s), 1096 warning(s)`, `exit=0` | **HELD** (matches the expected baseline exactly) |

The lint binary was **built from this tree** (`make build` → `bin/moai`, `BuildID
v3.1.2-943-g3f03d9c36-dirty`) and invoked as `./bin/moai`. A PATH-resolved binary may predate the
rules in this tree, and a green from it would mean nothing.

### Post-run corrections (iter5) — four records

manager-spec corrected `spec.md` / `plan.md` / `acceptance.md` after the run-phase commits landed.
These four records are what run-phase owes the record; they are defects this card produced or
judgments it made, not commentary.

#### Record 1 — the unjudged-criterion distinction (P3's uncovered half)

AC-IEC-007's second half used to require `internal/template/evidence_citation_guard_test.go` to be
**absent from the tree**, as evidence that this card did not create t375's guard file. t375 landed
that file into `origin/develop` (`46d27faa9` / `2b00bf0b4`), so absorbing would have failed the
criterion for a file this card never touched.

I first raised it as an open hazard rather than fixing it, on the ground that AC-IEC-007 was passing
and passing criteria should not be rewritten. **The operator overturned that premise**, and the
reasoning is the record:

> The stop rule protects a criterion passing **on its judgment**. This half passed because this
> worktree happens not to contain that file, while develop does. **A criterion whose answer flips
> when you change trees is not passing — it is unjudged.**

So the fix is not a rewrite of a passing criterion; it repairs an unjudged one to measure what it
was written to measure. The guard file is now simply the eighth path of the same three-dot diff,
changing the question from *is it absent from the tree* to *is it absent from this card's own diff*
— a property this card controls.

**This is the half of the P3 hazard that P3 could not reach.** P3 replaced a frozen SHA with the
three-dot form, which fixed every criterion decided by a **diff**. File existence is not a diff, so
the `ls` half sat outside P3's reach and kept the original defect in a different shape. A repair
that fixes a hazard on one axis and leaves the same hazard standing on another is the same
stop-at-the-boundary form recorded in rule 4 below.

#### Record 2 — this card's own milestone branched on the defect it repairs

`plan.md` M3 keyed the html treatment on whether the `.moai/state/verify/eb01063e` scratch tree
**exists**. Measured, both directions:

```
$ ls -d /Users/goos/MoAI/moai-adk-go/.moai/state/verify/eb01063e   → exists (8 files, 244K)
$ ls .moai/state/verify                                            → No such file or directory
```

The same test answers **(c) in the primary checkout and (b) in this worktree**. `.moai/state/` is
gitignored and therefore tree-local, so a milestone gated on its existence is a milestone gated on
a non-resolving path — precisely the defect class this SPEC exists to repair, committed by the SPEC's
own plan. Identifiability — the discriminator `spec.md` §C.3 states correctly — is a property of the
**evidence**, not of the tree asking, so it does not vary. It also selects the right branch here
independently: the tree exists, yet no single file decides a report-wide footer drawn from eight
per-category batch audits.

**Provenance, stated honestly rather than as a clean catch.** I relayed that tree's existence to the
lead **without carrying which tree measured it** — the same defect the lead had just flagged in my
own reporting, committed in the act of reporting the correction. The correction is stronger for
having been found that way: it was caught because the missing tree-attribution made the reading
ambiguous, which is the mechanism working, not the author being careful.

#### Record 3 — why the first commit carries the whole lifecycle

`ee77a6c88` carries `(none) → draft → in-progress`, so **no `draft → in-progress` edge appears in
git history**. The reason, stated next to the fact so a later auditor does not read the missing edge
as a lost transition: the plan phase never committed — both `.moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/`
and `.moai/reports/t381/` were untracked at run-phase entry (`git status` showed two `??` lines), so
the run-phase commit is also the SPEC's first commit.

A retroactive split was rejected. Reconstructing a plan-phase commit that never happened would
fabricate the record — the same hazard as rule 3 below, which forbids rewriting the record of a past
judgment. **The operator approved this explicitly.**

#### Record 4 — a moving base validated the three-dot choice, twice, by accident

`origin/develop` advanced **during** the session through the shared object store, with this lane
never fetching — and it did so twice:

```
9328a5242  (iter3/iter4)  →  297a21ea7  (+5, card t377)  →  59e898b31  (+9 more)
```

Every three-dot measurement survived both moves **with no edit to any criterion**. A frozen SHA
would have needed re-pinning twice in a single card. This is measured evidence for the P3 decision,
not an anecdote: the base moved on its own and the criteria that re-resolve their merge base kept
answering, while the one criterion that asserted fixed tree state (record 1) is exactly the one that
broke.

**One figure did move, and the record says so.** manager-spec's iter5 note recorded "the merged-tree
SHA unchanged → `1a137719…`" — true across `9328a5242 → 297a21ea7`, false across
`297a21ea7 → 59e898b31`, where the merged tree is now `2acdfbf88`. A merged-tree SHA is a function
of **both** parents, so it is itself a moving coordinate and cannot be quoted as a stable one. The
`ac-iec-007.txt` entry is **appended to, never rewritten**: the earlier reading stands as the record
of what was measured then, and the appended entry supersedes only its SHA, not its verdict. Note the
useful split the re-measurement exposes — the guard file's **blob** hash is identical across both
bases (`9b1970fe32f0…`) while the **tree** hash moved. When the claim is about a file, cite the blob.

### The four rules this card earned

Recorded as deliverables, not commentary. Each came from a defect this card actually produced.

1. **An id resolving does not make the sentence containing it true.** F1's lesson. `grep` answers
   "does this token exist", never "does this sentence still say something true" — so a reference that
   *rebinds* to a different referent is invisible to the sweep that a *dangling* reference would trip.
   Every hit is read for meaning.
2. **When an identifier is reassigned rather than removed, grep before reassigning.** Afterwards a
   stale reference and a correct one are indistinguishable. (Already carried at `spec.md` §E.2 —
   confirmed present and unchanged this round.)
3. **Renumber the live artifact; never renumber the record of a past judgment.** Why
   `plan-audit-iter1/2/3.md` were deliberately left carrying the pre-renumber ids: an audit report
   records what was true when it ran, and rewriting it falsifies the trail to buy cosmetic
   consistency. The reports are dated and iteration-labelled, which is the frame a reader needs.
4. **Understating a fix's extent is how a fix gets applied partially.** The auditor estimated six
   edits for the renumber; the measured figure was **18 occurrence-lines across four artifacts** —
   and F1 is that gap realized: the sweep ran on `spec.md` and stopped there.

**The shape recurs outside this card.** A sibling lane hit the same thing the same day: one artifact
retracted an instruction while another still issued it. **Repairs that stop at a file boundary** is
this batch's recurring form — the fix is correct, and it is applied to one of the N places the fact
lives. Rules 1 and 4 are the two halves of that failure (you cannot see the remaining places, and
you under-estimate how many there are).

### F1-F3 sweep — closed, plus two more of the same shape found by applying the rule

`grep -n 'REQ-IEC\|REQ ids' progress.md` returned six hits; each was read for **meaning**.

| Site | Was | Now |
|---|---|---|
| `:101` (**F1**, major) | "neither creates coverage for **REQ-IEC-008**, which appears only in `spec.md`" — both clauses false after the renumber: `REQ-IEC-008` is now Collision avoidance, which **is** covered (AC-IEC-007) and **does** appear in acceptance.md | names the **probe-boundary requirement** with the id time-marked as "numbered `REQ-IEC-008` at the time", and records the reassignment |
| `:61-62` (**F1**, minor half) | iter2's additions restated in present numbering (`009`/`010`) | time-marked: numbered `010`/`011` **at the time**, renumbered to `009`/`010` at iter4 |
| `:103` (**F2**) | "REQ ids 009/010/011 keep their numbers so every AC cross-reference stays valid" — the iter3 ground iter4 overturned; `REQ-IEC-011` no longer exists | struck through and marked superseded, with the measurement that overturned it: `git grep 'REQ-IEC-' origin/develop -- . ':!.moai/specs'` → *(empty)*, `exit=1`, measured 2026-08-31 against `origin/develop` = `9328a5242` (verified this run — no external consumer existed to protect) |
| `:17` (**F3**) | "Requirements: REQ-IEC-001..009" (stale) and "Acceptance: AC-IEC-001..012" (reads as contiguous) | `001..010` with the iter1 value marked historical; the AC ids restated as `001..007, 010..012` — **ten MUST criteria, not a range** — since 008/009 were demoted to §D at iter2 |
| `:22-24` (**found by the sweep, not raised by any audit**) | "**Known limitation**: every REQ-ECC citation is to an unmerged t375 draft… HEAD SHA `b64043481` not independently verified" — superseded at iter3 when t375 landed, and left standing in the bullet that depends on it | struck through and superseded, pointing at the iter3 entry that corrected it |
| `:125` (**found by the sweep**) | "_Awaiting plan-audit iter3 and Implementation Kickoff Approval._" — iter3 and iter4 both ran; kickoff granted | struck through and superseded with iter4's verdict and the approval |

The last two are the same half-repair shape as F1, at lower severity, and neither was raised by any
of the four audit iterations. They were found only by applying rule 1 — reading each hit for meaning
rather than checking that it resolves. Flagged explicitly so the lead can reject either: both are
edits to `progress.md`, which this agent owns, and neither touches a requirement, a criterion, or
scope.

### Declined, with reason

- **P7 (adjacency enforcement)** — not implemented. Excluded by operator decision; `spec.md` §E.1
  carries the analysis a follow-up card needs. See §E.3 residual risk.
- **P6 (t375 framed as a live parallel lane at `spec.md` §C.5)** — not touched. It is `spec.md`
  **body** content, owned by manager-spec; the exclusion itself remains correct as scope discipline,
  so nothing false is being shipped, only a stale rationale.
- **REQ-IEC-005's parenthetical cause** — not edited, reported instead (see §E.1 M1). Requirement
  text is manager-spec's to change.
- **`origin/develop` absorption** — not performed. It is 11 commits ahead; absorption belongs to the
  integration window. Verified the five repair targets are unchanged across that range, so this base
  is sound: `git diff --stat 3f03d9c36 origin/develop -- <5 in-scope files>` → *(empty)*.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: ee77a6c88   # the M1-M5 commit; backfilled in 343bb4bd4 because a commit cannot name its own hash. Later run-phase commits on this branch: 343bb4bd4 (backfill), 0e903d464 (header reflow), plus the iter5 post-run correction commit that carries this line.
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 20   # 12 carve-out lines (AC-IEC-006 control sums to 12) + 8 t375-owned paths (AC-IEC-007), all verified untouched
l44_pre_commit_fetch: "git fetch origin develop; git rev-list --count --left-right origin/develop...HEAD -> 11 0, measured 2026-08-31 with origin/develop resolved to 9328a5242 and HEAD 3f03d9c36 (a dated point-in-time divergence reading, not a standing invariant)"
l44_post_push_fetch: not-applicable   # lane does not push; integration belongs to the lead-assigned window
new_warnings_or_lints_introduced: 0  # final state. NOT zero on the first pass: the §E.2/§E.3 text I authored raised 4 NEW MovingRefUnpinned warnings (progress.md:213,215,277,311), caught by re-linting after the edit. Closed before commit — two by R3 (`<!-- moving-ref-ok: -->` with the reason, because the three-dot merge-base form MUST stay moving per iter4 P3) and two by R1/R4 (dated point-in-time readings pinned to origin/develop 9328a5242). Re-measured: ./bin/moai spec lint -> 0 error(s), 1096 warning(s), identical to the pre-edit baseline; targeted lint on this SPEC -> "No findings". go build + go test exit 0.
cross_platform_build:
  darwin_amd64_or_native: "go build ./internal/cli/... ./internal/hook/... → exit=0"
  windows_amd64: not-run             # comment-only diff, no syscall surface touched; CI owns the matrix
  status: scoped-to-change
total_run_phase_files: 5             # the 5 in-scope citation files; + this SPEC directory and .moai/reports/t381/ (both first-commit)
m1_to_mN_commit_strategy: "M1-M5 landed as ONE commit (ee77a6c88), followed by two mechanical commits (343bb4bd4 SHA backfill, 0e903d464 header reflow) and one iter5 post-run correction commit. The M1-M5 commit is also the SPEC's FIRST commit: the plan phase never committed — the SPEC directory and .moai/reports/t381/ were both untracked at run-phase entry (git status: two '??' lines) — so it carries (none) -> draft -> in-progress and lands at in-progress, and NO draft->in-progress edge appears in git history because no draft commit ever existed. A retroactive split was rejected: reconstructing a plan-phase commit that never happened would fabricate the record, the same hazard as rule 3 (never rewrite the record of a past judgment). Operator-approved explicitly; see §E.2 record 3."
```

### Residual risk

- **The C4 instruction is untouched** (`spec.md` §C.8). The same citations can be re-produced by the
  agents still instructed to write into `.moai/state/verify/$MOAI_SESSION_ID/`. This card cleans a
  corpus; it does not close the axis that fills it. t375 owns that axis.
- **Adjacency is not enforced** (P7, `spec.md` §E.1). AC-IEC-001 is file-granular — `grep -L` accepts
  a marker anywhere in the file, so a marker five hundred lines from the citation would pass. All
  four surviving repairs place the marker **at** the citation site, so the gap is latent today, not
  live. It goes live the moment someone "satisfies" AC-IEC-001 from a distance.
- **B2-B6 blind spots** (`spec.md` §C.7) are disclosed, not investigated. B2 alone — `.moai/state/`
  outside `/verify` — is 467 lines, 18× this card's scope.
- **`origin/develop` is 11 ahead and unabsorbed.** The three-dot baseline re-resolves the merge base,
  so absorbing it does not invalidate AC-IEC-006 or AC-IEC-007; but the criteria have not been run
  against the absorbed tree, because that tree does not exist yet.
- **The eight `eb01063e` JSONs remain unexported**, by decision (§E.1 M1). They exist on the
  authoring machine today and will not survive it. Their loss is now recorded at the citation site
  rather than disguised, which is what the landed `manager-lead.md:150` sentence prescribes — but it
  is a recorded loss, not a repaired one.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: 4f2eee332   # `docs(SPEC-IGNORED-EVIDENCE-CITATION-001): sync-phase artifacts — 3-phase close (t381)`. Left as an empty slot in that commit — a commit cannot cite its own hash — and backfilled here, in the immediately following commit. No `pending-backfill` placeholder was ever written to the branch.
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-IGNORED-EVIDENCE-CITATION-001' CHANGELOG.md -> 0 (exit=1), measured before emission at HEAD 2c566eaf3. No duplicate entry; emission proceeds."
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 12 DISTINCT identifiers (AC-IEC-001..012), NOT 10. The CHANGELOG entry states TEN and names the range explicitly as 001..007 + 010..012. The gap is real and explained, not softened: AC-IEC-008 and AC-IEC-009 were demoted to non-gating §D structural checks S-1/S-2 at plan-audit iter2, and acceptance.md carries NO reserved [RETIRED]/[REF] adjacency marker on either id, so the mechanical counter reads both as live. Under the marker convention this is not the 'ambiguous' case (no occurrence of either id is marked), so the counter emits 12 rather than halting. REPORTED to the lead rather than reconciled here: adding the markers would be a body edit to acceptance.md, forbidden to manager-docs."
b12_self_test_c: "ls over every path claimed in the CHANGELOG entry -> exit=0. Verified present: internal/cli/mcp_glm.go, internal/cli/audit_pin_live_test.go, internal/hook/evidence_writer_zeroexec_test.go, .moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt, .moai/reports/template-skill-improvement-plan-20260710.html, .moai/reports/t381/census.md, .moai/reports/t381/verify/ (all ten ac-iec-<nnn>.txt present)."
changelog_entry_position: "CHANGELOG.md [Unreleased] -> ### Fixed, first entry (inserted immediately after the heading, ahead of the card t373 entry, matching the section's newest-first ordering)."
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (merged close; the single sync commit carries implemented and completed together). updated: 2026-08-31 — already the correct value, unchanged."
  plan_md: "NOT touched. Carries no `status:` and no `updated:` field. Per spec-frontmatter-schema.md § Artifact Statelessness, non-spec.md artifacts are stateless on the status axis; adding either field would create the exact defect card t369 removed from SPEC-INTEGRATION-LOCK-ATOMIC-001."
  acceptance_md: "NOT touched. Carries no `status:` (correct, same rule). `updated: 2026-08-31` already holds the sync-commit date, so no refresh was needed."
  progress_md: "no frontmatter block; §E.4 (this section) is the sync-phase record."
canary_compliance_check:
  applicable: false
  reason: "This SPEC defines no forward-looking policy that its own sync would test. It repairs a five-line corpus under an already-landed sibling canon (SPEC-EVIDENCE-CITATION-CANON-001) and explicitly builds no guard (spec.md §C.8 'Out of Scope — guard construction')."
mx_tag_validation:
  status: "no-op, verified rather than assumed"
  reason: "Every hunk in the run-phase diff is a comment, a header, or a footer. No exported function was created or changed, no fan_in changed, no goroutine or dangerous pattern introduced — so no @MX:NOTE / @MX:ANCHOR / @MX:WARN obligation is triggered. No @MX:TODO was left outstanding by run-phase."
docs_sync:
  readme_4_locale: "not applicable — measured, not assumed. grep -rln over README.md / README.ko.md / README.ja.md / README.zh.md / docs-site/ for the five in-scope filenames returns ZERO hits (exit=1)."
  template_mirror: "not applicable — `ls internal/template/templates/internal` -> No such file or directory. The template tree carries no mirror for `internal/`, and the two report files live under .moai/reports/ (internal evidence, never shipped)."
sync_verification:
  binary: "built from THIS tree — `make build` at HEAD 2c566eaf3 -> bin/moai, BuildID v3.1.2-948-g2c566eaf3. Invoked as ./bin/moai, never the PATH-resolved binary (which may predate this tree's rules and return a meaningless green)."
  spec_lint_repo_wide: "TWO figures, and the second is the one this commit ships. Pre-edit, at HEAD 2c566eaf3: `0 error(s), 1096 warning(s)`, exit=0. Post-edit, with this §E.4 block in place: `0 error(s), 1097 warning(s)`, exit=0. Both measured IN THIS TREE with the tree-built binary and reported as this tree's own figures, not as agreement with a count measured elsewhere. The +1 is named below and self-heals at the backfill commit."
  new_warning_introduced: "ONE, deliberate and transient — reported rather than absorbed. `SyncSHASlotFormat` at progress.md:440: the rule landed by SPEC-SYNC-SHA-SLOT-FORMAT-001 (card t299) admits a commit SHA or the canonical `pending-backfill` family, and does NOT admit the empty string. The dispatch for this card explicitly instructs the empty-slot form and explicitly forbids leaving a `pending-backfill` placeholder on the branch, so the two rules genuinely conflict on this one line. The dispatch was followed; the cost is one WARNING (never an error) living for exactly one commit. Self-healing OBSERVED, not predicted: after the backfill wrote the real SHA, targeted lint on this SPEC -> `✓ No findings` and repo-wide -> `0 error(s), 1096 warning(s)`, back to the pre-edit figure. CONFLICT REPORTED to the lead — reconciling it is a doctrine decision (which of the two close conventions is canonical), not a manager-docs edit."
  ac_rerun: "All ten MUST criteria re-executed at the sync commit's parent state: 001 (empty output, exit=0) + companion (four files at 1, mcp_glm.go absent by treatment (a)); 002 (0 / 1); 003 (5); 004 (1 / 1); 005 (1 / 1, PASS via the recorded-inability branch); 006 (diff exit=0, positive control sums to 12); 007 (eight-path diff exit=0); 010 (all ten evidence files listed exit=0, check-ignore exit=1 i.e. NOT ignored, `ls .moai/state/verify` -> No such file or directory); 011 build half (`go build ./internal/cli/... ./internal/hook/...` exit=0); 012 (0 / 2)."
  ac_011_test_half: "NOT re-run at sync, and the reason is stated rather than elided: the sync diff touches CHANGELOG.md, spec.md frontmatter, and this progress.md section — zero Go files, so no Go behaviour can have changed. The build half WAS re-run (exit=0). The test half stands on the run-phase measurement (28 packages ok, internal/cli 473.9s, internal/hook 83.7s, exit=0). Full-suite and cross-platform judgment belongs to CI, per CLAUDE.local.md §4."
  moving_base_note: "origin/develop moved AGAIN during this card, unfetched, through the shared object store: 9328a5242 -> 297a21ea7 -> 59e898b31 -> e79272713 (four readings, three moves). The merge base is unchanged at 3f03d9c36. Every three-dot criterion absorbed all three moves with no edit — the fourth independent confirmation of the iter4 P3 decision, and a fourth re-pin a frozen SHA would have needed."
integration:
  pushed: false
  reason: "Lane does not push. Integration into develop and the repository-wide CI verdict belong to the lead-assigned window; origin/develop is NOT absorbed here."
```

### Sync-phase residual risk

- **`sync_commit_sha` is empty until the backfill commit lands.** Between the sync commit and its
  successor, a reader of this block sees an empty slot. That is deliberate — the alternative, a
  `pending-backfill` placeholder, is a value that can survive un-backfilled and be mistaken for one.
- **The AC-count self-test disagrees with the SPEC's own prose (B12-b above), and the disagreement is
  shipped rather than resolved.** `acceptance.md` says ten MUST criteria; the mechanical counter says
  twelve identifiers. Both are correct about different questions, and the CHANGELOG names the range
  so a reader can tell which. Reconciling it needs `[RETIRED]` markers on the two demoted ids in
  `acceptance.md` — a body edit outside this agent's ownership, so it is reported to the lead.
- **Everything §E.3 records as residual is unchanged by this close** — the C4 instruction axis, the
  unenforced adjacency, the B2-B6 probe blind spots, and the eight unexported `eb01063e` JSONs. The
  sync phase repaired no residual and closed no axis; it recorded them.
