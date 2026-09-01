# SPEC Review Report: SPEC-GITIGNORE-ROOT-GUARD-001

**Iteration**: 2/3 — delta scope (operator-authorized above the Tier S ceiling of 1)
**Auditor**: plan-auditor (independent; M1 context isolation applied — reasoning context ignored)
**Verdict**: **PASS**
**Overall Score**: **0.8125** (Tier S threshold 0.75)
**Score movement**: 0.69 (iter 1) → 0.8125 (iter 2), monotonic upward; no STOP escalation condition

> Export note: this file is an export of the iteration-2 verdict as judged, not a re-audit. No
> criterion was re-derived while writing it, and no verdict was revised. Anything not actually run
> is recorded in the Gaps section rather than reconstructed.

---

## Baseline-attribution

| Coordinate | Value |
|---|---|
| Worktree (audit tree) | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t377` |
| Tree HEAD audited against | `3f03d9c36` (`3f03d9c36faf49bdcb155d98a7009fc9d8dd9659`) |
| Branch | `WT-gitignore-parity` |
| SPEC under audit | `.moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md` (294 lines, untracked) |
| Tier | S (ACs inline in spec.md; no `acceptance.md`, no `plan.md`, no `research.md`) |
| Tooling note | The moai MCP server was unavailable this session; every check ran via the Bash CLI. |
| Read-only proof | `git status --porcelain` over the SPEC dir, `.gitignore`, and `internal/template/` returned only `?? .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/` |

Every command quoted below was run in this run, in this worktree, against this HEAD.

---

## Verdict summary

- All eight reported defects (D1–D8): **CLOSED**.
- Five new defects found (N1–N5): all **minor**, none critical, **none applied** (this audit is read-only; they are recommendations for the run phase).
- No must-pass criterion failed.
- The two critical closures (D1, D2) rest on measured evidence, not on readable prose.

---

## Category scores

| Dimension | Score | Rubric band | Basis |
|-----------|-------|-------------|-------|
| Clarity | 0.75 | 0.75 | One interpretation per requirement; §A's restated risk and §B's numbers are precise. Deductions: §E:272 "One test file under `internal/template/`" leaves new-vs-existing unresolved, which matters because AC-GRG-007(b) pins a specific file (N2); REQ ordering 007-before-006 (N3); modality against project convention (N5). |
| Completeness | 0.75 | 0.75 | Frontmatter complete and valid; §A/§B/§C/§D/§E present; Out of Scope carries three specific bullets and passes the repo's own `OutOfScopeRule` (`internal/spec/lint.go:1017`, which requires the phrase plus a `-` bullet). One non-critical section missing: no `HISTORY` section (heading inventory shows §A–§E only) — pre-existing, not part of the delta. |
| Testability | 0.85 | above 0.75, short of 1.0 | The dimension that drove the 0.69, and the lift is real. Every AC now carries a command. AC-001's discriminator is measured to work (D1). AC-002/003 are mutations with restore verification. AC-004 names its own uncovered mutant and the criterion that closes it. AC-005 verified to return `1`. AC-006 is two literal string checks. AC-007 states RED-now per half. No weasel words anywhere. Short of 1.0 on two vacuity-adjacent structures: N2 and N4. |
| Traceability | 0.90 | above 0.75 | The REQ↔AC map (spec.md:254-266) covers all 7 REQs; no orphaned AC, no uncovered REQ. Deduction: traceability is reader-legible only — the ACs carry no `maps REQ-XXX` token, so `internal/spec/ears.go:128` extracts no mapping and `CoverageIncomplete` reports all 7 REQs uncovered once the bullets are made parseable (see N5). |

**Aggregate (arithmetic mean)** = (0.75 + 0.75 + 0.85 + 0.90) / 4 = **0.8125**.

---

## Must-pass results

- **[PASS] MP-1 REQ number consistency** — REQ-GRG-001…007, complete set, no gaps, no duplicates, consistent 3-digit padding. Presentation order places 007 before 006 (N3); none of MP-1's three enumerated failure conditions is met, so this is a clarity defect rather than a must-pass failure.
- **[PASS] MP-2 EARS/GEARS format compliance** — all six REQ entries are ubiquitous-form (`The <subject> <modality> <response>`); REQ-GRG-004 is the canonical negative form. Judgment made against the **requirement layer** (`REQ-XXX` in spec.md); the verification layer was graded under Group 4. **Stated so the PASS is not misread**: the modality keyword is `MUST`, not `shall` — a measured deviation from a 374:4 repo convention that the repo's own checker rejects (N5). Not escalated because none of MP-2's enumerated failure conditions applies: the statements are not informal language, not Given/When/Then presented as REQs, and not mixed informal/formal within one requirement.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`id`, `title`, `version: "0.1.0"` quoted, `status: draft`, `created`/`updated` ISO dates, `author`, `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags`), plus `tier: S`. No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) present.
- **[N/A] MP-4 language neutrality** — single-repository SPEC about this repository's own `.gitignore` and Go test tree; no multi-language tooling surface. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the only SPEC reference in the body is its own ID, so there is no reconciliation obligation. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` returns `0`; auto-PASS per D8-4.
- **[N/A] MP-7 clarification gate** — neither `plan.md` nor `research.md` exists (Tier S; the SPEC directory contains `spec.md` only). `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC directory returned nothing (exit 1).

---

## D1–D8 disposition table

| # | Sev | Disposition | Evidence that decided it |
|---|-----|-------------|--------------------------|
| D1 | critical | **CLOSED** | Zero-selection run emits no `--- PASS:` line; AC-GRG-001 requires that literal (measured — commands C1/C2) |
| D2 | critical | **CLOSED** | All `enforce*` / `neutrality-check` occurrences confined to the retraction paragraph (C7) |
| D3 | major | **CLOSED** | `EntryMerge` dispatch verified (C4); "not undone by the later merge" measured in a throwaway repo (C8/C9) |
| D4 | major | **CLOSED** | Anchored grep returns `1`, unanchored returns `3` — both cited figures reproduce (C3) |
| D5 | minor | **CLOSED** | AC-GRG-004 carries its own three-step command and names the mutant it does not close |
| D6 | minor | **CLOSED** | All six REQs active-voice; one legitimate passive residue in the unwanted-behaviour form |
| D7 | minor | **CLOSED** | AC-GRG-003 declares RED unobserved for this guard and marks observing it a run-phase obligation |
| D8 | minor | **CLOSED** | REQ↔AC map table present at spec.md:254-266, all 7 REQs covered |

### D1 (critical) — CLOSED. Judgement, not restatement.

The vacuity is **eliminated, not relocated**. The guard does not exist yet, so the anchored selector
necessarily selects nothing today — but AC-GRG-001's `--- PASS:` requirement makes that state
**distinguishable from a real pass**: the zero-selection run emits `PASS` and
`ok … [no tests to run]` and **no `--- PASS:` line at all**, under either flag combination measured.
AC-GRG-001 requires the literal `--- PASS: TestGitignoreDeclaredRulesOnBothSurfaces`, which an empty
run cannot produce.

The instrument also over-delivers relative to its own stated rationale — the ` [no tests to run]`
suffix and the `testing: warning:` line are two further discriminators the SPEC did not notice. That
gap sits in the reasoning, not in the instrument; it is recorded as N1.

REQ-GRG-006 pinning the test name closes the remaining path: a rename now breaks the selector loudly
instead of silently emptying it.

### D2 (critical) — CLOSED.

The enforcement claim survives only as a recorded retraction. Every occurrence of
`enforce*` / `neutrality-check` in the document (lines 99, 101-102, 109) sits inside the retraction
paragraph, which ends: *"Nothing in this SPEC depends on the enforcement claim, and the
byte-equality exclusion rests on the doctrine, not on a guard."* No requirement, criterion, or scope
clause cites it.

### D3 (major) — CLOSED, and the restated risk is supported.

Both discarded rationales are recorded rather than deleted (§A:53-66). The merge **direction** is
verified in code: `internal/cli/update/plan/plan.go:47` dispatches `.gitignore` to
`merge.EntryMerge`, and `MergeGitignoreFile` (`internal/cli/update/merge/merge.go:57+`) builds the
template line-set and re-appends user-only lines — a union. A root-only loss of a template-carried
rule is therefore healed by the next `moai update`, so §A's inversion is corrected in the right
direction.

The load-bearing clause — *"the commit is not undone by the later merge"* — was **measured**, not
reasoned to (C8/C9): in a throwaway repo, commit the artifact first, then add the ignore rule. The
artifact stays tracked and clean. An ignore rule added later does not untrack an already-tracked path
and does not touch the commit.

### D4 (major) — CLOSED.

Both figures cited in AC-GRG-005 reproduce exactly on this tree: anchored `1`, unanchored `3` (C3).

### D5 (minor) — CLOSED.

AC-GRG-004 now carries its own three-step command (re-measure divergence → confirm non-zero → run the
anchored test) and **names** the mutant it does not close (an implementation that skips its
assertions whenever divergence is non-zero), together with the criterion that does close it
(AC-GRG-002). Naming an uncovered mutant rather than implying coverage is the right move.

### D6 (minor) — CLOSED.

All six REQs are active-voice. REQ-GRG-007's second clause (*"that rule MUST NOT be added"*) reads
passively, but that is the GEARS unwanted-behaviour form prohibiting an act by an unnamed actor —
legitimate, not a residue to fix. The modality keyword is a separate matter (N5).

### D7 (minor) — CLOSED.

AC-GRG-003 states RED is unobserved for **this** guard, cites the equivalent RED observed in card
t373 for the single-surface guard, and marks observing it a run-phase obligation. Correct
disposition: a disclosure, not a claim.

### D8 (minor) — CLOSED.

The REQ↔AC map (spec.md:254-266) covers all seven REQs. No orphaned AC; no uncovered REQ.

---

## New findings N1–N5 (introduced by the fixes / fresh material)

**Applied status: NONE APPLIED.** This audit is read-only and did not modify the SPEC. All five are
recommendations for the run phase. None is critical; none reopens D1 or D2; none outranks the
closures.

| # | Severity | Class | Applied? | Summary |
|---|----------|-------|----------|---------|
| N1 | minor | blocking | NOT APPLIED | spec.md:212 states an unmeasured (and false) claim about `go test` output |
| N2 | minor | optional | NOT APPLIED | AC-GRG-007(b) is pinned to a file path, not to the declared set |
| N3 | minor | optional | NOT APPLIED | REQ-GRG-007 listed before REQ-GRG-006 |
| N4 | minor | optional | NOT APPLIED | Preservation-check discipline applied unevenly (AC-GRG-005 unlabelled) |
| N5 | minor | blocking (modality half only) | NOT APPLIED | REQ modality `MUST` against a measured 374:4 project convention |

### N1 — the D1 fix introduced an unmeasured claim about `go test` output, inside the section whose subject is not making unmeasured claims about `go test` output.

spec.md:212 states: *"a bare `ok` is also what a zero-selection run prints."* Measured, it is not:
the `ok` line carries a ` [no tests to run]` suffix in **both** the `-v` and the plain form, and
under `-v` a `testing: warning: no tests to run` line precedes it (C1/C2).

Nothing depends on the inaccuracy — the AC's actual instrument is strictly stronger than its stated
rationale — but the sentence teaches a false fact to the next reader, and the §D preamble pinned the
**prefix**-selector output verbatim while reasoning about the **zero**-selection output without
running it. One sentence to correct.

### N2 — AC-GRG-007(b) is pinned to a file path, not to the declared set.

It greps `internal/template/embed_gitignore_generated_test.go` for `mink`. That file holds
`generatedArtifactIgnoreRules` today (verified, C3), so the check is live now. But REQ-GRG-002
constrains the **count** of declaration sites, not their **location**: if the run phase moves the var
into the new guard's file, (b) becomes a grep of a file that no longer carries the declared set — a
vacuous pass, the same class as D1. Likelihood is low (§E's "extending the existing declared-rule
mechanism" points at reuse in place), which is why this is classed optional rather than blocking.
Fix: point (b) at whichever file AC-GRG-005 locates.

### N3 — REQ-GRG-007 is listed before REQ-GRG-006.

spec.md:180 vs :184 — a revision artifact of appending 006 after 007. No gap, no duplicate,
consistent padding, so MP-1 holds; but out-of-order enumeration is the drift this SPEC otherwise
polices.

### N4 — the preservation-check discipline is applied unevenly.

AC-GRG-007(b) is explicitly labelled *"a preservation check rather than a flip"*. AC-GRG-005 is
**also** satisfied on the pre-implementation tree (measured `1` today, C3) and is not labelled. Same
status, one disclosure.

### N5 — REQ modality is `MUST` against a measured 374:4 project convention, and the repo's own SPEC linter rejects it.

Two separable facts; only one is charged to this SPEC.

The linter reports clean on this document (C5) — and **that green is vacuous**.
`reqLineWidePattern` (`internal/spec/lint_req_widen.go:36`) requires `- **REQ-XXX**: <text>`; this
SPEC uses `- **REQ-GRG-001** — <text>` (em-dash), so **zero** requirements are extracted and every
REQ-level rule runs on an empty set. Mutating only the seven bullet separators from `—` to `:` in a
scratch copy surfaces 13 warnings the authored form hides (C6).

- **The em-dash half is NOT this SPEC's defect.** The corpus carries 599 em-dash REQ bullets against
  636 colon ones across 716 SPECs — a project-wide parser/convention gap belonging on someone else's
  card. The SPEC never claims to lint clean, so nothing here rests on the vacuous green.
- **The modality half IS this SPEC's own deviation.** Repo-wide, REQ statements of the form
  `The <subject> <modality>` use `shall`/`SHALL` 374 times and `MUST` 4 times; `isValidEARSFormat`
  (`internal/spec/lint.go:731`) rejects `THE … MUST` without ` SHALL`. One substitution to fix, and
  it aligns with 374 precedents.

---

## Fresh material judged on its own

### REQ-GRG-007 / AC-GRG-007 — the `**/.mink/auth/` in/out split

**The split is coherent and correctly recorded.** All measured premises reproduce (C10). The shape —
add the line, do not declare it — is right, and §B:164 states the general principle it protects:
*"what keeps 'add a missing rule' and 'extend the guard's contract' from becoming the same act by
default."* That is a genuinely useful boundary; it was not questioned.

**AC-GRG-007's two-part form is sound in structure.** Part (a) is a clean flip: `git check-ignore`
exits 1 today, must become 0 naming the root `.gitignore`. Part (b) is a preservation check, correctly
self-labelled as one. The vacuity risk in (b) is **not** that it is a preservation check (a
preservation check passing is the intended outcome) — it is the file-path coupling, recorded as N2.

### §E Scope (rewritten) — no contradiction

In Scope names the test file plus the one `.gitignore` line and explicitly excludes it from the
declared set; Out of Scope's first bullet carries the corresponding open judgment (whether credential
rules belong in the guard's contract). Both halves agree with each other and with §B:156-162.

### §B re-verified in full

Justified because AC-GRG-004's premise is that divergence is non-zero. Every §B number reproduces on
HEAD `3f03d9c36` (C11–C13): root 177, template 135, root-only 44, template-only 2 — and the two
template-only rules are exactly the two §B names. AC-GRG-004's premise holds today with margin.

---

## Commands run, with observed output

**C1 — anchored selector, `-v` (the D1 discriminator test):**

```
$ go test ./internal/template/ -run '^TestGitignoreDeclaredRulesOnBothSurfaces$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.472s [no tests to run]
exit=0
```

**C2 — anchored selector, no `-v`:**

```
$ go test ./internal/template/ -run '^TestGitignoreDeclaredRulesOnBothSurfaces$' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.259s [no tests to run]
```

No `--- PASS:` line in either. This is what closes D1 and what produces N1.

**C3 — declaration-site counts (D4, N2, N4):**

```
$ grep -rn 'generatedArtifactIgnoreRules' internal/template/ --include='*.go'
internal/template/embed_gitignore_generated_test.go:23:// generatedArtifactIgnoreRules are the .gitignore lines that must reach users.
internal/template/embed_gitignore_generated_test.go:26:var generatedArtifactIgnoreRules = []string{
internal/template/embed_gitignore_generated_test.go:48:	for _, rule := range generatedArtifactIgnoreRules {

$ grep -rn '^var generatedArtifactIgnoreRules' internal/template/ --include='*.go' | wc -l
       1
$ grep -rn 'generatedArtifactIgnoreRules' internal/template/ --include='*.go' | wc -l
       3
```

**C4 — merge dispatch (D3):**

```
$ sed -n '40,55p' internal/cli/update/plan/plan.go
func DetermineStrategy(filename string) merge.MergeStrategy {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)

	switch {
	case base == "CLAUDE.md":
		return merge.SectionMerge
	case base == ".gitignore":
		return merge.EntryMerge
	...
```

`MergeGitignoreFile` (`internal/cli/update/merge/merge.go:57+`) reads the deployed template content,
builds a set of its non-blank non-comment lines, and appends user-only lines — a union.

**C5 — repo SPEC linter, as authored (N5):**

```
$ go run ./cmd/moai spec lint .moai/specs/SPEC-GITIGNORE-ROOT-GUARD-001/spec.md
✓ No findings — all SPEC documents are valid
```

**C6 — same document, seven REQ bullet separators changed from em-dash to colon in a scratch copy (N5):**

```
$ go run ./cmd/moai spec lint <scratch>/spec.md
WARNING  ModalityMalformed   155  REQ REQ-GRG-001: EARS modality violation — SHALL missing or format mismatch: "The guard MUST assert the generated-artifact rule set against the repository's"
WARNING  ModalityMalformed   157  REQ REQ-GRG-002: EARS modality violation — SHALL missing or format mismatch: "The source MUST hold exactly one definition of the declared rule set. Two"
WARNING  ModalityMalformed   159  REQ REQ-GRG-003: EARS modality violation — SHALL missing or format mismatch: "The guard MUST compare discrete rule lines — never bytes, and never full"
WARNING  ModalityMalformed   161  REQ REQ-GRG-004: EARS modality violation — SHALL missing or format mismatch: "The guard MUST NOT fire on the two files' intended differences: divergent"
WARNING  ModalityMalformed   166  REQ REQ-GRG-007: EARS modality violation — SHALL missing or format mismatch: "The root `.gitignore` MUST carry `**/.mink/auth/`, and that rule MUST NOT be"
WARNING  ModalityMalformed   170  REQ REQ-GRG-006: EARS modality violation — SHALL missing or format mismatch: "The guard's test function MUST be named"
WARNING  CoverageIncomplete  155  REQ REQ-GRG-001 is not referenced by any AC
   ... (7 CoverageIncomplete rows, one per REQ)

0 error(s), 13 warning(s)
```

Diff between the two copies: 7 lines (the bullet separators only).

**C7 — residual enforcement claims (D2):**

```
$ grep -n "neutrality-check\|enforced\|enforcement\|template-neutrality" <spec.md>
99:**The doctrinal ground holds; a mechanical-enforcement claim does not, and is not made.** §25
101:earlier draft additionally asserted the prohibition was "enforced by
102:`template-neutrality-check.yaml`". That was **written without observation and is false for this
109:enforcement claim, and the byte-equality exclusion rests on the doctrine, not on a guard.
```

All four occurrences are inside the retraction paragraph.

**C8/C9 — throwaway-repo demonstration that a later ignore rule does not undo a prior commit (D3):**

```
# sequence: git init; create gen/out.txt; git add; git commit "commit inside the window";
#           then echo "gen/" > .gitignore   (simulating the later moai update merge)
$ git -C <tmp> ls-files gen/out.txt
gen/out.txt
$ git -C <tmp> status --porcelain --ignored gen/out.txt
(empty output — tracked, clean, not ignored)
```

**C10 — mink premises (REQ/AC-GRG-007):**

```
$ grep -c mink .gitignore                                          → 0
$ grep -n mink internal/template/templates/.gitignore              → 169:**/.mink/auth/
$ grep -c mink internal/template/embed_gitignore_generated_test.go → 0
$ git check-ignore -v .mink/auth/token                             → (no output) exit=1
$ ls -d .mink                                                      → ls: .mink: No such file or directory
```

**C11/C12/C13 — §B divergence re-measured:**

```
$ grep -vE '^\s*#|^\s*$' .gitignore | sort -u | wc -l                              → 177
$ grep -vE '^\s*#|^\s*$' internal/template/templates/.gitignore | sort -u | wc -l  → 135
$ comm -23 <root> <template> | wc -l    # root-only                                → 44
$ comm -13 <root> <template>            # template-only
.agents/skills/moai*
**/.mink/auth/
```

**C14 — heading inventory (Completeness):**

```
$ grep -n "^## \|^### " <spec.md>
19:## §A Context and Problem
47:### Why this is a live risk — stated as measured, after two wrong rationales
78:## §B Measurement — what the guard can and cannot be
138:### The `**/.mink/auth/` rule — in scope by operator judgment, deliberately outside the guard
167:## §C Requirements
189:## §D Acceptance Criteria
254:### Requirement ↔ criterion map
268:## §E Scope
270:### In Scope
278:### Out of Scope
```

No `HISTORY` section.

**C15 — repo-wide convention measurement (N5):**

```
$ (over 716 SPEC spec.md files) REQ statements matching "The <subject> (MUST|shall|SHALL)"
   4 MUST
 314 shall
  60 SHALL
$ REQ bullet form:  colon 636   em-dash 599
```

---

## Gaps — what was NOT observed

Recorded rather than reconstructed. Each is something this verdict does **not** rest on, or rests on
by reading rather than by running.

- The §B neutrality-leak experiment (D2's 9-char / 8-char SHA measurement) was **not** re-run. D2's
  disposition rests on the enforcement claim having been **withdrawn** — a textual verification —
  not on re-confirming leak-guard behaviour.
- `update_template_sync.go:459` / `:510` were **not** verified line-for-line. The healing claim rests
  on the `EntryMerge` dispatch at `plan.go:47` (C4) plus reading `MergeGitignoreFile`'s body.
- `moai update` was **not** executed against this repository. The healing behaviour is verified by
  code reading plus the dispatch, not by execution.
- AC-GRG-003's RED is unobserved by this auditor as well as by the SPEC: no template mutant was
  planted and no `make build` was run. It remains a run-phase obligation, as the SPEC now says.
- The 374:4 modality figure (C15) comes from one regex over `spec.md` REQ lines
  (`The <subject> (MUST|shall|SHALL)`); statements with other subjects or phrasings fall outside it.
  Directionally decisive, not exhaustive.
- No AC was executed against a working guard, because no guard exists. Every AC except AC-GRG-005 and
  AC-GRG-007(b) is unexercisable on this tree.

## Residual risk

- The guard does not exist yet. This PASS is a judgment about the criteria's **soundness as
  instruments**, not evidence that any of them will pass once the guard is written.
- **N2 is the live one**: if the run phase relocates `generatedArtifactIgnoreRules`, AC-GRG-007(b)
  silently stops observing anything — and does so with a green. The run-phase implementer should be
  told to keep the declared set in `embed_gitignore_generated_test.go`, or to re-point (b).
- `moai update` run against this repository would restore a missing declared rule, but
  `MergeGitignoreFile` collects only non-comment lines into the user-additions block: the root file's
  SHA-citing comment block would not survive the same way, and preserved user lines are **relocated**
  under a marker (order is semantic in gitignore, per the code's own comment at `merge.go:53-56`).
  Outside this SPEC's scope; worth knowing before anyone runs `moai update` in this repo.

---

## Recommendation

**PASS.** The eight defects are closed, the two critical ones on measured evidence rather than on
readable prose, and the new §B / §E / REQ-GRG-007 material holds up. The design remains right and was
not revisited (the iteration-1 standing recommendation is carried forward unchanged).

Four cheap edits before run-phase, in priority order — none blocks the verdict, none applied here:

1. Correct spec.md:212 to the measured zero-selection output (`ok … [no tests to run]`, plus the
   `testing: warning:` line under `-v`) rather than "a bare `ok`" (N1).
2. Substitute `shall` for `MUST` in the six REQ statements, matching 374 repo precedents and
   satisfying `isValidEARSFormat` (N5).
3. Re-point AC-GRG-007(b) at whichever file AC-GRG-005 locates, instead of the hard-coded path (N2).
4. Reorder REQ-GRG-006 before REQ-GRG-007, and label AC-GRG-005 a preservation check the way
   AC-GRG-007(b) already is (N3, N4).

**One item for a different card, not this one**: 599 of 716 SPECs use the em-dash REQ bullet form
that `reqLineWidePattern` cannot parse, so `moai spec lint`'s REQ-level rules are a no-op across most
of the corpus. That is a parser gap in `internal/spec`, and it is the same empty-sweep shape this
SPEC was written about — worth a card of its own.
