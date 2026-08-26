# SPEC Review Report: SPEC-PRECOMMIT-PRESERVE-001

Iteration: 3/3 (Tier M ceiling of 2 consumed by iterations 1-2; this iteration runs under explicit user override)
Audited tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t230`, branch `WT-precommit-preserve`, HEAD `21db95293`
SPEC version audited: `0.4.0` (the trimmed scope; no prior audit has seen it)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.8875** (Tier M PASS threshold 0.80)
Score movement: 0.875 (iter1) → 0.8875 (iter2) → **0.8875 (iter3)** — **flat, not monotonic-up**

Reasoning context ignored per M1 Context Isolation. Every figure below was re-measured in this run
against this tree; nothing is inherited from `plan-audit.md`, `plan-audit-2.md`, the SPEC's own
HISTORY claims, or the dispatch.

---

## Preliminary: tree confirmation

```
$ pwd                        → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t230
$ git rev-parse --short HEAD → 21db95293
$ git branch --show-current  → WT-precommit-preserve
```

Matches the dispatch. Proceeded.

One dispatch figure did not reproduce: the dispatch states "12 REQ / 13 AC claimed". Measured
**12 REQ / 12 AC** (`grep -c '^- \*\*REQ-PCP-' spec.md` → `12`; `grep -c '^### AC-PCP-'
acceptance.md` → `12`), which is what spec.md §E and acceptance.md §D.3 both claim. The dispatch's
"13" is a dispatch error, not a SPEC defect, and the "verify the extra AC is a genuine sub-case
split" instruction is moot — there is no extra AC. AC-PCP-005 carries three sub-cases inside one
criterion, which is why the count did not move when N4 was addressed.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 12 requirements, no duplicates. The sequence reads
  001-010, 013, 014 with gaps at 011, 012, 015.
  Evidence: `grep -n '\*\*REQ-PCP-[0-9]*\*\*' spec.md` → 12 lines, IDs
  006/014/001/002/005/003/004/007/009/008/010/013, each appearing once.
  **Judgment, stated so it can be disagreed with**: a literal reading of "no gaps" would FAIL this.
  I record PASS because the criterion exists to catch an *unaccounted* discontinuity — a requirement
  that vanished without trace. Here every gap is accounted for in-document and verifiably: spec.md
  §B:362-367 names all three retirements and their destination, §D:415-436 states what moved and
  why, and §E:486-489 states the pairing invariant (a requirement and its criterion always leave
  together). The accountability is complete, so nothing is lost. See D6 below, which keeps the
  literal-reading tension visible rather than silently resolved.

- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-PCP-*` in
  `spec.md` §B) only, per the two-layer rule; the `AC-PCP-*` Given-When-Then entries in
  `acceptance.md` are verification-layer and are graded under Group 4, not here.
  All 12 match a GEARS pattern:
  - Ubiquitous: REQ-PCP-001 (`spec.md:290`), REQ-PCP-013 (`:357`, generalized artifact subject —
    "`preCommitHookContent` and `…/pre-commit` shall remain byte-identical", permitted since GEARS
    `<subject>` may be any noun).
  - Event-driven (`When …, the installer shall …`): REQ-PCP-002 (`:292`), REQ-PCP-003 (`:302`),
    REQ-PCP-004 (`:304`), REQ-PCP-005 (`:296`), REQ-PCP-008 (`:335`), REQ-PCP-010 (`:338`).
  - Unwanted (`shall not`, the GEARS canonical negative): REQ-PCP-006 (`:279`), REQ-PCP-007
    (`:328`), REQ-PCP-009 (`:330`), and the second clause of REQ-PCP-014 (`:286`) and REQ-PCP-004
    (`:309`).
  No legacy `If … then …` form and no deprecation marker is required anywhere. No Given-When-Then
  appears in the requirement layer.

- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical names
  and correct types (`spec.md:1-15`): `id`, `title` (quoted), `version: "0.4.0"` (quoted semver),
  `status: draft` (enum), `created: 2026-08-24`, `updated: 2026-08-24` (ISO dates), `author`,
  `priority: P2` (enum), `phase: "v3.1.4 target"` (release target, not a prohibited lifecycle
  token), `module: "internal/cli"`, `lifecycle: spec-anchored` (enum), `tags` (comma-separated
  string). Optional `tier: M` present. No rejected snake_case alias.
  Mechanically confirmed rather than inferred:
  ```
  $ ~/go/bin/moai spec lint .moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md
  ✓ No findings — all SPEC documents are valid
  ```
  `FrontmatterSchemaRule` therefore ran and emitted nothing, which also settles the multi-segment
  `id` (`SPEC-PRECOMMIT-PRESERVE-001`): the tool accepts it.

- **[N/A] MP-4 language neutrality** — single-language SPEC. Its entire surface is one Go package
  (`internal/cli`) plus one shell template; it makes no multi-language tooling claim. Auto-passes.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the verb executed:
  `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md plan.md acceptance.md progress.md | sort -u`
  → exactly one ID, `SPEC-PRECOMMIT-PRESERVE-001` (itself). No cross-SPEC reference exists, so no
  retired/superseded/archived referent can be unreconciled. No BLOCKING finding.
  Note: the SPEC's forward dependencies are expressed as **cards** (t237/#1641, t235/#1639, "a
  successor card"), not SPEC IDs, so D7's mechanical sweep cannot see them. D7 is clean; the card
  dependency is nonetheless where this audit's blocking finding lives — see D1.

- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rc syscall` over all four artifacts → `0` in
  each. D8-4 auto-PASS: no `syscall` mention, so no build-tag concern exists.

- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC directory
  → rc 1, no matches. `plan.md` exists and is clean; `research.md` does not exist, which is correct
  for Tier M (3 artifacts). No open clarification marker.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | Unusually explicit throughout: §A.0 pins the citation convention, §A.3 states the design crux as a table, §A.5 states rejected alternatives per decision. One requirement-adjacent ambiguity remains: §C.2 (`spec.md:386-396`) says the t237 collision is "moot" and that the diffs "may land first in any order" without naming which axis that holds on — true for textual conflict, false for release composition (D1). |
| Completeness | 0.85 | 0.75-1.0 | All required sections present. Seven `### Out of Scope — <topic>` H3 sub-headings with specific bullets (`spec.md:415,438,443,450,455,460,465`). Frontmatter complete. Missing: an owner for the release-composition constraint where the body-changing card is t237 rather than the successor card (D1), and a stated disposition for AC-PCP-005(c) when the hook body legitimately changes (D3). |
| Testability | 0.85 | 0.75-1.0 | Every criterion carries Decides / Baseline / Mutant / Failing-input, and the mutant discipline is genuinely adversarial (the "correct but silent" mutant is named once and jointly defeated by AC-004/006/007). One criterion's deciding command cannot observe the gap its own prose declares (D2, AC-PCP-005). No weasel words found in any criterion. |
| Traceability | 1.00 | 1.0 | 12 REQ ↔ 12 AC, exactly 1:1 (`spec.md:471-489` matrix re-verified against both files). Every AC heading names a valid, existing REQ; every REQ has a criterion. No orphan AC, no uncovered REQ. All three retired IDs are accounted for in three places and appear nowhere as a live obligation (measured: every `PCP-011|012|015` hit is a HISTORY entry, a retirement note, or a progress.md record — never a plan step or a trace row). |

Aggregate: (0.85 + 0.85 + 0.85 + 1.00) / 4 = **0.8875**. That it lands on iteration 2's figure
exactly is a coincidence of the bands, not an anchor: I computed the four dimensions before
comparing. What it means is stated in the verdict section — the trim closed real defects and opened
new ones, so quality moved sideways rather than up.

---

## Part 1 — did the trim actually dissolve N1-N4?

| # | Iteration-2 defect | Disposition | Evidence measured in this run |
|---|---|---|---|
| **N1** | REQ-PCP-015 bound a release but AC-PCP-015 ran at end-of-M2 where it was green by construction; no release-time enforcement existed | **DISSOLVED, with a residual (→ D1)** | REQ/AC-PCP-015 are gone (`grep` finds them only in HISTORY, retirement notes, and progress.md). Nothing remaining is green-by-construction. Body stability inside this SPEC is now enforced by three checks that **can** fail: AC-PCP-005 sub-case (c), AC-PCP-013's PASS-not-SKIP requirement, plan.md §D:62-69, and plan.md M2's end-of-milestone scope gate (`plan.md:115-118`). That is a genuine improvement on the removed criterion. **Residual**: the *release-scope* half of the old constraint — "a body change must not ship in the classifier's release" — is now assigned entirely to the successor card (`spec.md:201-205`, `:430-434`), while the card actually about to change the body is t237. See D1. |
| **N2** | Three conflicting referents for "the last released hook body" | **DISSOLVED** | The phrase is gone from every artifact. `grep -rn 'last released\|last shipped\|previous shipped\|previously-shipped\|historically-shipped'` over all four files returns: `spec.md:218` and `:223`/`:230` (inside §A.5's *rejected-alternative* discussion, where each occurrence states its referent in the same sentence — "sha256(previous shipped `preCommitHookContent`)"), plus one `progress.md:50` mention of N2 itself as history. **No live requirement or criterion depends on the phrase.** Where a released body is now named, it is named as a concrete tag (`v3.1.2`), which resolves. |
| **N3** | t237/#1641 edits the same constant/template pair; unresolved which card yields | **PARTIALLY DISSOLVED — conflict axis yes, composition axis NO (→ D1)** | Verified: no remaining requirement, milestone, or AC causes an edit to `preCommitHookContent` or the template twin. `grep -rn 'preCommitHookContent'` over the SPEC returns only prohibitions (`plan.md:62`, `spec.md:379-384`), assertions that it stays unchanged (`acceptance.md:313,327`), and test fixtures that *read* it. `plan.md:6-10` confirms the body-changing milestone is gone. So the **textual** collision is genuinely dissolved. The **composition** hazard is not — see D1. |
| **N4** | The first-upgrade legacy population was never measured against real shipped bytes; AC-PCP-005(b)'s fixture was "identical to the incoming content" | **PARTIALLY CLOSED (→ D2, D3)** | Sub-case (c) does use real shipped bytes, and I verified the referent resolves and the claim holds: `git rev-parse -q --verify refs/tags/v3.1.2` → `50c271d79…` (tag present); `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit \| cmp - internal/template/templates/.git_hooks/pre-commit` → **rc 0, identical**; `wc -c` → `3245`. No network call — a local tag lookup. The population claim also re-measures: `git tag --contains 883d53852` → `v3.1.0 v3.1.0-rc.0 v3.1.0-rc.1 v3.1.0-rc.2 v3.1.1 v3.1.2` (plus a non-release `sw-pre-merge-backup` tag the SPEC omits, correctly, since it is not a release). And the v3.0.x split re-measures exactly: v3.0.1 constant digest `f79adf7f…` vs HEAD `3442efc9…` — differ, as §A.5 states. So (c) covers the real no-record population, on real bytes. **But** its deciding command cannot distinguish the skip it declares a gap (D2), and it fuses two purposes with no stated exit (D3). |

### D1 in detail — the defect the trim moved rather than removed

This is the finding the "scope reduction can hide a defect" hazard predicts, so it is stated with
its evidence in full.

**What the SPEC now asserts.** `spec.md:386-396` (§C.2, "Card t237 — no collision remains"):

> "**The question is now moot**: with the extension point moved to a successor card (§D Out of
> Scope), this SPEC touches neither file (§C.1), so the two diffs are disjoint and either may land
> first in any order."

**What I measured.** `gh issue view 1641` → **OPEN**, titled "pre-commit hook: go vet runs from the
repository root's go.mod…". Its body states the fix requires editing, byte-identically:
`internal/cli/hook_install_precommit.go` (the `preCommitHookContent` constant) and
`internal/template/templates/.git_hooks/pre-commit` (the template twin) — and records a **verified
patch already on a branch** (`t312-precommit-vet @ b6f478b1a`). A hook-body change is therefore not
hypothetical and not far off.

**Why "moot" is only half true.** The claim holds on the axis of *textual merge conflict* — the two
diffs touch disjoint files, so either order merges cleanly. It does not hold on the axis the
retired REQ-PCP-015 existed to bind:

1. **Same-release alarm fatigue.** If t237's body change ships in the same release as the
   classifier, then for every existing installation `installed != incoming` and no provenance record
   exists, so REQ-PCP-005 (`spec.md:296-298`) classifies **100% of the installed base** as
   user-modified — a backup plus a notice for every user on first upgrade. That is verbatim the
   trap the SPEC itself names at `spec.md:430-434` ("the alarm fatigue §A.3 exists to prevent,
   delivered by the mechanism meant to prevent it") and at `spec.md:201-205` ("Any *later* change to
   the hook body inherits this arithmetic and must not be folded into the classifier's own
   release"). Both passages then assign that constraint **exclusively** to "the successor card that
   carries the extension point". t237 is not that card, and nothing binds it.
2. **AC-PCP-005 sub-case (c) inverts.** If t237 lands before the classifier's release, then
   `v3.1.2` bytes ≠ incoming bytes, and the *correct* behaviour for a v3.1.2-era hook with no record
   becomes backup + notice. Sub-case (c) asserts the opposite ("no backup is taken, no notice is
   emitted", `acceptance.md:126-128`). The criterion would go red for a legitimate sibling change,
   and the cheapest repair under pressure is to delete or silently re-pin it — losing the only
   criterion that measures the real population.

So §A.5's own constraint ("must not be folded into the classifier's own release") survives the trim
as prose, while the mechanism that bound it (REQ/AC-PCP-015) was removed and its ownership was
narrowed to a card that is not the one about to change the body. The premise "either may land first
in any order" is an **unverified premise** on the composition axis — reachability of a clean merge
is not justification for release-order indifference.

I am not arguing REQ-PCP-015 should return; iteration 2 was right that it was unenforceable where it
sat. The gap is that its *subject* was narrowed from "any hook-body change in this release" to "the
successor card", and t237 fell through.

---

## Part 2 — v0.4.0 on its own terms

### The core mutant (an implementation that overwrites and backs up correctly but prints nothing)

Constructed independently and walked against the 12 criteria as they now read. Such an
implementation:

- passes AC-PCP-001 (sidecar written), AC-PCP-002 (silence is what it does), AC-PCP-003 (backup file
  present with pre-run bytes), AC-PCP-005(a) file-state, AC-PCP-008, AC-PCP-009, AC-PCP-013,
  AC-PCP-014;
- **fails AC-PCP-004 clause (i)** — the warning writer's captured output must contain the backup
  path *and* a replacement statement (`acceptance.md:86-88`), and the mutant's warning writer is
  empty;
- **fails AC-PCP-006** — a single run must produce **both** artefacts, asserted in one case
  (`acceptance.md:158-161`), which is what stops the "each checked in a separate test" escape;
- **fails AC-PCP-007 clause (i)** — the warning writer's output must be non-empty
  (`acceptance.md:177-178`).

Three independent criteria catch it, and AC-PCP-004's failing-input line names it explicitly as the
input that must be observed red (`acceptance.md:116-118`). **The mutant is defeated.** The two-writer
rewording of AC-PCP-007 also closes the older escape (a whitespace-padded success line now leaves
the warning writer empty rather than merely differing).

### Orphan-reference hazard — clean

REQ-PCP-004 (`spec.md:309-313`) forbids the notice from naming `.git/hooks/pre-commit.local`, and
acceptance.md:90-93 mirrors the prohibition and explains that asserting the string would lock in a
false instruction. I checked every `pre-commit.local` occurrence in all four artifacts (15 hits):
every one is either a prohibition, an Out-of-Scope entry, a HISTORY/progress record, or the §A.5
Decision 2 disclosure that durable recovery is deferred. **No requirement, AC, plan step, or asserted
notice string points a user at a facility this SPEC does not ship.** The specific hazard the dispatch
named — an AC asserting a notice string that names a non-existent facility, which would force the
implementation to emit it — was present at v0.3.0 and is removed: the deleted line read
`…a statement that the hook was replaced, and the string `pre-commit.local`` (see the weakening
table below). Its removal is correct.

### Weakening check — `git diff e7fdd4a47..21db95293`

Diffstat: acceptance.md ±146, plan.md ±76, progress.md +48, spec.md ±201 (252 insertions, 219
deletions). Every removed or edited clause, classified:

| Change | Class | Evidence |
|---|---|---|
| AC-PCP-004 clause (i): "**all three** of: path, replacement statement, and the string `pre-commit.local`" → "**both** of: path, replacement statement" | **Legitimately out of scope** | The third element names a facility this SPEC no longer ships; asserting it would force a false instruction. Justified in place (`acceptance.md:90-93`) and in REQ-PCP-004. |
| AC-PCP-004 clause (ii) Decides: "a static call-site check — `grep …` (a Go test asserting the same is equally acceptable)" → "**a Go test is the deciding check**; a grep is not sufficient on its own" | **Strengthened** | `acceptance.md:95-105` now explains why the grep is defeated by the real wiring (`errOut` on the call line, `cmd.ErrOrStderr()` two lines away). I verified that reasoning: `update_template_sync.go:575` currently passes `out`; `:69` is `out := cmd.OutOrStdout()`; `:72` is `errOut := cmd.ErrOrStderr()`. A call-line grep for `ErrOrStderr` would indeed find nothing under a correct implementation. This closes iteration 2's N5. |
| AC-PCP-005: sub-cases (a)+(b) → (a)+(b)+**(c)** | **Strengthened** | New sub-case measures real `v3.1.2` bytes; new second mutant named. |
| AC-PCP-011, AC-PCP-012, AC-PCP-015 removed | **Legitimately out of scope** — each left *with* its requirement | Verified pairing: REQ-PCP-011/012/015 and AC-PCP-011/012/015 all absent; §E:486-489 states the invariant; no orphan on either side. |
| §D.2: "`make build` after any change to `preCommitHookContent` or the template twin" → "`make build` is **not** expected" | **Legitimately consequent** | Follows from §C.1 (no body change). Guarded: `acceptance.md:313-315` makes a run needing `make build` a stop-and-re-open condition rather than an absorbed edit. |
| §D.3: "All 15 AC … 4 mutant-only (008, 012, 013, 015)" → "All 12 AC … 2 mutant-only (008, 013)" | **Arithmetically consistent** | 012 and 015 left with the extension point; 002 remains separately classified as unconstructible-on-tree. Recount: 12 − 2 − 1 = 9 "fail against the untouched tree", which is what §D.3 claims. |
| Removed DoD line: "REQ-PCP-011's paired edit lands … **in a later release than the rest of this SPEC** … a release candidate that fails it is not shippable" | **Removed because it was failing — and its subject was narrowed** | This is the D1 finding. The clause was unenforceable where it sat (iteration 2's N1, correctly diagnosed), but its removal also removed the only text binding *any* body change in the classifier's release. §D:430-434 re-homes it onto the successor card only. |

**No AC lost an asserted clause while silently keeping its ID.** The one clause-losing edit
(AC-PCP-004) is disclosed in the criterion body, in REQ-PCP-004, in plan.md §G, and in HISTORY.

### Counts and parity

12 REQ / 12 AC, 1:1, against a Tier M ceiling of 16/16 — within budget on both axes independently.
The set **is** 1:1 (the dispatch's "no longer 1:1" premise does not reproduce); AC-PCP-005's three
sub-cases live inside one criterion, which is why closing N4 changed no count.

### Retired-identifier gaps

Every `PCP-011` / `PCP-012` / `PCP-015` occurrence (14 hits across four files) is a HISTORY row, a
retirement note (`spec.md:362-367`, `:419-420`, `:426`, `:486`), a §C.2 genealogy sentence, or a
progress.md record of the audit round. **No plan step, milestone, trace row, or DoD item references
a retired identifier as a live obligation.** Clean.

### Falsifiability of every criterion

Checked each for the "would pass against the untouched tree" failure. The three self-declared
non-falsifiable-on-tree criteria are honestly labelled:

- **AC-PCP-008** (behaviour-preserving) — baseline verified: `hook_install_precommit.go:145` at HEAD
  is `if !hasMarker {`, the marker-absent branch. Correctly labelled.
- **AC-PCP-013** (behaviour-preserving) — baseline verified: `hook_install_precommit_test.go:38` is
  `func TestPreCommitTemplateMatchesConstant(t *testing.T) {` and `:45` is the `t.Skipf` branch the
  criterion names as its mutant. Correctly labelled, and it is the one criterion that explicitly
  rejects SKIP.
- **AC-PCP-002** — labelled with an explicit "Falsification class" paragraph (`acceptance.md:62-66`)
  stating its Given is unconstructible on the untouched tree. Honest, and it says outright it must
  not be counted as evidence about the pre-implementation tree.

The other nine are falsifiable on the untouched tree, and their baselines re-measure:
`grep -c 'pre-commit\.bak' internal/cli/hook_install_precommit.go` → **0**;
`grep -c 'sha256' …` → **0**; `wc -l` → **179**; the sole output line at
`hook_install_precommit.go:172` is `_, _ = fmt.Fprintln(out, "  Pre-commit hook installed
(.git/hooks/pre-commit)")`; `installPreCommitHookOptional` at `:166` has the single-writer
signature `(projectRoot string, skip bool, out io.Writer)`, exactly as REQ-PCP-004 and §C.4 state.

### Citation integrity (§A.0 convention)

Every source citation re-resolves at HEAD `21db95293`. The SPEC's claim that its two cited SHAs
differ from HEAD only in this SPEC's own files is **verified**:

```
$ git diff --name-only 294b4b6ab..HEAD   → the 4 SPEC files, nothing else
$ git diff --name-only 7b2f42be0..HEAD   → the 4 SPEC files, nothing else
```

So `@294b4b6ab` and `@7b2f42be0` citations are equally valid at HEAD, as §A.2:85-89 claims. Spot-check
at HEAD: `:26` = `const preCommitHookContent = ` + backtick; `:126` = `func (p *PreCommitInstaller)
InstallPreCommitHook(skip bool) error {`; `:139` = `if _, err := os.Stat(hookPath); err == nil {`;
`:145` = `if !hasMarker {`; `:151` = `if err := os.WriteFile(hookPath, []byte(preCommitHookContent),
0o755); err != nil {`; `:166`, `:172` as above; test `:38`, `:45` as above. Call sites:
`update_template_sync.go:575` and `init.go:898`, both as cited. One imprecision — see D5.

### Requirement C closure (§A.4) — re-run independently

```
$ grep -rEn 'exec [0-9][0-9]*[<>]' --include='*.sh' --include='*.tmpl' --include='pre-commit' \
    --include='pre-push' --include='*.go' --include='*.md' . | grep -v '^\./\.git/' | wc -l
0
```

**0 hits.** The closure is correct under the strict pattern as written. §A.4's explanation of why no
count is recorded for the loose pattern (the paragraph is itself one of its matches, so any numeral
drifts on the next edit) is sound and closes iteration 2's N6. Not blocking.

---

## Defects Found (structured defect-list)

**D1. Release-composition constraint survives the trim but its subject was narrowed to the wrong
card** — `.moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md`:L386-396 (§C.2), with L201-205 (§A.5) and
L430-434 (§D) — Severity: **critical** — Class: **blocking**.
*Evidence*: §C.2 declares the t237 collision "moot" and the diffs free to "land first in any order";
`gh issue view 1641` shows the card OPEN, editing `preCommitHookContent` **and** the template twin,
with a verified patch on branch `t312-precommit-vet @ b6f478b1a`. §A.5:201-205 states the constraint
("a body change … must not be folded into the classifier's own release") and assigns it to "the
successor card that carries the extension point"; §D:430-434 repeats the assignment. t237 is not that
card. Consequence (a): a same-release t237 makes `installed != incoming` for 100% of the no-record
installed base, so REQ-PCP-005 fires a backup + notice for every user — the alarm fatigue §A.3 exists
to prevent. Consequence (b): a t237 landing before this SPEC's release makes AC-PCP-005 sub-case (c)
assert the wrong outcome, since backup + notice then becomes the *correct* behaviour for a v3.1.2
body.
*Required fix*: restate the §A.5/§D composition constraint so it binds **any** hook-body change
shipping in the classifier's release — naming t237/#1641 explicitly alongside the successor card —
and correct §C.2 to say the collision is dissolved **on the merge-conflict axis only**, with the
release-order question still open. Whether that constraint is re-bound by a requirement, by a
release-checklist item outside this SPEC, or by an explicit accept-the-noise decision is a choice
this SPEC should state; leaving it unstated is what makes this blocking.

**D2. AC-PCP-005's deciding command cannot observe the skip its own prose calls a gap** —
`acceptance.md`:L136-140 — Severity: **major** — Class: **blocking**.
*Evidence*: the criterion states "A test that cannot reach git (shallow clone, no tags) must **skip
loudly**, never pass quietly; a skipped (c) is a gap, not a pass." Its Decides is
`go test ./internal/cli/ -run TestPreCommitLegacyNoRecord -count=1`, which exits **0** on a `t.Skip`.
The deciding command therefore returns the same verdict for "(c) passed" and "(c) never ran". This is
the identical hazard AC-PCP-013 identifies and solves one criterion away (`acceptance.md`:L260-269:
"**Then** it reports PASS and **not** SKIP", Decides = "that command, with the run's skip status
inspected") — the fix pattern exists in the same document and was not applied here. Partial
mitigation measured: this repo's CI Test jobs use `fetch-depth: 0` (`.github/workflows/ci.yml`:L119,
L219, L316, L343, L398, L455), so tags resolve in CI; the exposure is forks, tarball environments,
and shallow clones. It matters because sub-case (c) is the **sole** evidence closing iteration 2's
N4.
*Required fix*: give AC-PCP-005 the AC-PCP-013 treatment — add `-v` and "with the run's skip status
inspected" to Decides, and state in the Then clause that a skipped (c) fails the criterion.

**D3. AC-PCP-005 sub-case (c) fuses a requirement check with a release guard, with no stated
disposition** — `acceptance.md`:L126-134, L148-154 — Severity: **major** — Class: **blocking**.
*Evidence*: (c) is asked to do two jobs at once — prove REQ-PCP-005 against the real installed
population, and (per L130-134) act as the tripwire that "keeps this SPEC honest about leaving the
hook body untouched … the day someone changes `preCommitHookContent`, (c) goes red and says so". The
two signals are entangled: when the body legitimately changes (the successor card, or t237), (c) goes
red for a reason that has nothing to do with REQ-PCP-005, and REQ-PCP-005's correct behaviour for a
`v3.1.2` hook flips to backup + notice — the opposite of what (c) asserts. Nothing in spec.md,
plan.md, or acceptance.md states what the successor card does with (c). A dual-purpose criterion that
is designed to go red on a legitimate future change, with no retirement or re-pin plan, is a test
that gets deleted under deadline pressure — and deleting it removes the only measurement of the real
population.
*Required fix*: state (c)'s disposition on a legitimate body change — either re-pin its tag to "the
last release whose hook body predates the change" (the phrasing the removed AC-PCP-015 already used
to defeat exactly this re-pointing mutant), or split the body-stability guard into its own criterion
so the two signals are separable.

**D4. The notice is correctly scoped but the recovery gap it leaves is disclosed nowhere the user
sees** — `spec.md`:L268-273 (§A.5 Decision 2), REQ-PCP-004 `spec.md`:L304-313 — Severity: **minor** —
Class: **optional**.
*Evidence*: Decision 2 admits "Until it lands, recovery is manual re-application from the backup.
That is a real limitation, and it is stated rather than papered over." REQ-PCP-004 then constrains
the notice to exactly two elements and forbids naming `pre-commit.local` — both correct, and D-free.
The residue is that the user is told their patch was replaced and where the copy is, but not that the
next `moai update` will do the same thing again. That is a presently-true, non-vaporware statement
that names no facility. REQ-PCP-006's disclosure floor holds without it, which is why this is
optional rather than blocking.
*Required fix (if taken)*: permit REQ-PCP-004's notice a third element stating the replacement will
recur on the next update, and add the assertion to AC-PCP-004 clause (i). Equally defensible to
decline — a smaller notice is a stronger notice.

**D5. §D.1 edge-case citation points at an assignment, not the creation it describes** —
`acceptance.md`:L298 — Severity: **minor** — Class: **optional**.
*Evidence*: the row reads "`.git/hooks/` absent | created, as today
(`internal/cli/hook_install_precommit.go:131 @294b4b6ab`)". Line 131 at HEAD is
`hookDir := filepath.Join(p.repoRoot, ".git", "hooks")` — the path assignment, not the `MkdirAll`
that creates it. §A.0's symbol-anchor convention makes this harmless in practice, but the row's claim
and the cited line do not match.
*Required fix*: cite the creating call, or name the symbol instead of the line.

**D6. MP-1 literal-reading tension recorded so it is not silently resolved** — `spec.md`:L360-367 —
Severity: **minor** — Class: **optional**.
*Evidence*: REQ numbering reads 001-010, 013, 014 with three gaps. I scored MP-1 PASS on the
reasoning given above (every gap accounted for in three places; renumbering would falsify the two
prior audit reports' citations, which is a real cost). A reader applying "even one gap = FAIL"
literally reaches the opposite verdict on the same facts. The disagreement is about the criterion,
not about the SPEC.
*Required fix*: none required. If a mechanical lint later enforces contiguity, the SPEC's §B:362-367
paragraph is the place to carry a documented exemption.

---

## Regression Check (defects from prior iterations)

Re-verified independently rather than accepted from the prior reports. Iteration 1's D1-D7 were
reported closed at v0.3.0 and I found no evidence of regression at v0.4.0: REQ-PCP-010's
backup-failure precedence is present and explicit (`spec.md`:L338-356) with AC-PCP-010(a)'s
post-state assertion as its check (`acceptance.md`:L228-231, L243-250); REQ-PCP-004's warning-writer
split is present with §C.4's one-line caller permission (`spec.md`:L402-411) mirrored verbatim in
`plan.md` §A:L27-31; §A.4's sweep-pattern semantics are stated; the AC-PCP-010 backup-succeeded /
write-failed edge row is present (`acceptance.md`:L305); AC-PCP-002 carries its falsification label;
no `Where`-misuse remains in the requirement layer.

Iteration 2's N1-N4 and N5-N7: N2, N5 (AC-PCP-004 clause (ii) Decides), N6 (§A.4 self-counting
numeral) and N7 are closed on measured evidence. N1 and N3 are closed on their stated axes with the
residual recorded as D1; N4 is closed on its fixture question with D2 and D3 recorded against its
deciding command and its dual purpose.

**No defect appears unchanged across all three iterations**, so no stagnation flag. The pattern
instead is that each round closes what it names and surfaces something the previous round's structure
hid — which is why the score is flat rather than climbing.

---

## Verdict and fitness for run-phase

**PASS-WITH-DEBT, 0.8875, against the Tier M threshold of 0.80.** No must-pass criterion fails; the
aggregate clears the threshold comfortably; the debt is D1, D2 and D3.

**Score movement is flat (0.8875 → 0.8875), not up.** The LEAN STOP-escalation clause fires on a
*lower* score, so it does not fire here. But flatness across a scope reduction whose whole purpose
was to remove four blocking defects is itself the signal worth reading: the trim genuinely dissolved
N1 and N2, genuinely improved two criteria, and genuinely removed an unenforceable constraint — and
in the same motion narrowed that constraint's subject onto the wrong card (D1) and leaned the whole
of N4's closure on a criterion whose deciding command cannot see its own gap (D2). Quality moved
sideways.

**Is this SPEC fit to enter run-phase as it stands? Not quite — but the distance is small, and it is
not in the implementation.**

- **M1 (the three-way classifier) is fit.** Its four requirements are unambiguous, its criteria are
  falsifiable on the untouched tree, and AC-PCP-014's mandatory first case defeats the two-way
  implementation. Nothing in D1-D3 touches M1's implementation content.
- **M2 (backup and notice) is fit on its implementation content.** The card's headline mutant is
  defeated three times over, the stream choice is checked at the call sites, and REQ-PCP-010's
  precedence is asserted by post-state rather than by return value.
- **What must close first** — all three are edits to the SPEC's prose and one criterion, not to the
  design:
  1. **D2** — one line: give AC-PCP-005 the AC-PCP-013 skip-rejecting Decides. Smallest fix, and it
     is what makes N4's closure hold outside this repo's CI.
  2. **D3** — state sub-case (c)'s disposition when the hook body legitimately changes.
  3. **D1** — correct §C.2's "moot / any order" to name the axis it holds on, and re-scope the
     §A.5/§D composition constraint to bind any body change in the classifier's release, naming
     t237/#1641. This one is a release-sequencing decision the SPEC should record, not a
     re-architecture.

D4, D5 and D6 are optional and should not be routed into a revision on this auditor's account.

**Iteration ceiling.** This is iteration 3, the hard cap. Whatever is decided next, it is an
escalation decision rather than another audit round: accept PASS-WITH-DEBT and enter run-phase with
D1-D3 recorded as debt; close D1-D3 first (they are small and none needs a re-audit of the design);
or reduce scope further. My recommendation is the second — close the three, then enter run-phase —
because D2 and D3 are one-line and one-paragraph edits to a criterion that is load-bearing for the
defect this trim exists to close, and D1 is a decision that costs less to record now than to discover
after t237 merges.
