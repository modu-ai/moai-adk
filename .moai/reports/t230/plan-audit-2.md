# SPEC Review Report: SPEC-PRECOMMIT-PRESERVE-001 (iteration 2)

Card: t230 · Tier M · Iteration: **2/2** (Tier M ceiling, `harness.plan_audit_tier_ceilings`)
Audited tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t230`, branch
`WT-precommit-preserve`, HEAD `e7fdd4a47`. Artifacts at v0.3.0.

**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.8875** (Tier M PASS threshold `0.80`, `spec-workflow.md` § SPEC Complexity Tier)
**Movement: 0.875 → 0.8875 — monotonic (+0.0125). No STOP escalation.**

Reasoning context ignored per M1 Context Isolation. Every figure below was re-measured in this tree
in this run. Nothing was inherited from iteration 1's report, from the SPEC's own HISTORY claims, or
from the dispatch — including the remediation claims, each of which was re-derived from the artifact
text and, where it named a measurement, from the tree.

One new BLOCKING defect (N1) sits on the material iteration 1 could not have audited: the
release-composition design added at v0.3.0. It is a mechanism gap, not a reasoning failure —
REQ-PCP-015 is the right constraint, checked at the one point where it cannot fail. Gate: **before
M3 lands**, not before run-phase. M1 and M2 are fit to proceed today.

---

## Part 1 — closure of the iteration-1 defects

| # | Severity (iter-1) | Closure verdict | Basis |
|---|---|---|---|
| D1 | critical / blocking | **genuinely closed — by strengthening** | REQ-PCP-010 split into two precedence sub-clauses; AC-PCP-010(a) Then now asserts post-state; destroying mutant named verbatim |
| D2 | critical / blocking | **genuinely closed — by fixing, not weakening** | The harder of the two offered repairs was taken; all three surfaces re-read and agree; every writer binding re-measured |
| D3 | major / blocking | **substantially closed; one asked-for item closed by argument rather than by a criterion** (→ N4) | One-entry corpus evaluated on its own terms; Decision 3 taken; magnitude table reproduces exactly. The new design carries N1-N3 |
| D4 | minor | **closed** (new staleness nit → N6) | §A.4 now states what the pattern means; both sweeps re-run |
| D5 | minor | **closed** | Edge row present in `acceptance.md §D.1` with the "no replacement notice" outcome |
| D6 | minor | **closed** | AC-PCP-002 carries an explicit falsification-class clause; `progress.md` restated 11/3 → 10/5; arithmetic checks |
| D7 | minor | **closed** | `Where` → `When` on 005, 011, 012; 010 restructured. Residual modality nit → N7 |

**Nothing was closed by weakening.** I checked specifically for the regression shape the dispatch
named — a criterion edited until it stops failing. The v0.2.0 → v0.3.0 requirement diff is purely
additive plus rewording: diffing the sorted REQ bullet first-lines at `7b2f42be0` against `e7fdd4a47`
shows 004, 005, 010, 011, 012 reworded, 015 added, **nothing removed and nothing merged**. No AC
lost a clause: AC-PCP-010 gained one, AC-PCP-004 and AC-PCP-007 each gained a clause and a mutant,
AC-PCP-002 gained a falsification-class statement.

### D1 — closed, and the mutant now fails

REQ-PCP-010 (`spec.md §B`, Preserved behaviour) now states the precedence explicitly:

> **When the backup write fails** — which is *before* the hook write — the installer shall report a
> warning, shall **not overwrite the hook**, and shall not fail the caller.

AC-PCP-010(a) Then: "…**and the installed hook's bytes are byte-identical to their pre-run value** —
no replacement occurred." The Mutant field names the destroying implementation first, in bold, and
states that sub-case (a)'s post-state assertion is the *only* thing that reaches it — with the
correct reason (AC-PCP-006's Given is the AC-PCP-003 scenario, in which the backup succeeds).

**Mutant constructed and re-run against v0.3.0**: backup fails → warn → overwrite anyway. Returns
normally ✓, emits a warning ✓, does not panic or abort ✓, hook bytes unchanged ✗ → **FAILS
AC-PCP-010(a)**. The hole is closed. `plan.md §F` M2 and `§G` both carry the same precedence, so an
implementer reading only the plan reaches the same place.

The SPEC additionally does something iteration 1 did not ask for, which I checked for soundness: it
argues the precedence is **not** a carve-out from REQ-PCP-006 ("no backup means no replacement, so
the invariant holds by never replacing it at all"). That reasoning is correct — the invariant is
about replacement, and this branch does not replace.

### D2 — closed on all three surfaces, each re-measured

The SPEC took repair option (ii) (add a writer, permit the caller change), the *more* expensive of
the two iteration 1 offered, and justified it with an argument iteration 1 did not make: a data-loss
notice on stdout is swallowed by a scripted `moai update >/dev/null`. That is a fix, not a
weakening.

Re-measured at HEAD `e7fdd4a47` (identical to `7b2f42be0` for all source files — the only diff
between those commits is this SPEC directory):

| Claim in the SPEC | Command | Result |
|---|---|---|
| single writer today | `sed -n '166p' internal/cli/hook_install_precommit.go` | `func installPreCommitHookOptional(projectRoot string, skip bool, out io.Writer)` ✓ |
| update binds stdout | `sed -n '69p' internal/cli/update_template_sync.go` | `out := cmd.OutOrStdout()` ✓ |
| second unrelated `out` at `:604` | `sed -n '604p'` | `out := cmd.OutOrStdout()` ✓ — the disambiguating note is correct |
| update call site | `grep -n installPreCommitHookOptional …` | `:575` passes `out` ✓ |
| `errOut` already held | `sed -n '72p'` | `errOut := cmd.ErrOrStderr()` ✓ |
| init binds stderr | `sed -n '898p' internal/cli/init.go` | `cmd.ErrOrStderr()` ✓ |

Surface agreement, verified by reading each: `spec.md` REQ-PCP-004 (warning writer distinct from
progress writer; both callers bind it to stderr) · `spec.md §C.4` (exactly one change per call site,
naming the writer each already holds) · `plan.md §A` ("Each caller changes by exactly one line, and
no more", with the v0.2.0 "neither caller changes" retracted in place) · `plan.md §D` (caller edits
limited to the one argument) · AC-PCP-004 (two captured writers; clause (i) names the **warning**
writer, clause (ii) the call sites) · AC-PCP-007 (both writers; (i) warning writer non-empty, (ii)
contains the backup path). Every AC clause names which writer it inspects, as the dispatch required.

One residual on the *check* rather than on the requirement — N5 below.

### D3 — the one-entry corpus is now rejected on its own terms; the release decision is taken

The cheap form is evaluated in its own block with three reasons, and I re-measured every figure that
block rests on:

| §A.5 claim | Re-measured | Result |
|---|---|---|
| body unchanged since `883d53852` (2026-07-28) | log of the template path, last commit | `883d53852 2026-07-28` ✓ |
| carried by `v3.1.0`…`v3.1.2` | tags containing `883d53852` | `v3.1.0`, `-rc.0/1/2`, `v3.1.1`, `v3.1.2` ✓ |
| `v3.1.2` body == HEAD body | tag blob piped to `cmp` | rc 0 ✓ |
| `v3.0.1` constant ≠ HEAD constant | constant block extracted, `shasum -a 256` on each | `f79adf7f…` vs `3442efc9…` ✓ **both digests reproduce exactly** |

The three rejection reasons hold on inspection. Reason 1 (no-op at the release it would first ship
in) follows from Decision 3 and the measured byte-equality. Reason 2 (its only coverage is
version-skippers at the M3 release) is correct: a user who took the classifier release carries a
record, so the one-entry digest can only matter to someone who skipped it. Reason 3 (it becomes a
third member of the constant/template paired edit) applies the same maintenance-coupling argument
that got the in-body stamp rejected. This is a rejection of the cheap form, not a restatement of the
expensive one — D3's core complaint is answered.

Decision 3 itself is sound and the option table is honest: option (c) is rejected for the right
reason (it ships a body change through the *unprotected* installer), a stronger argument than
iteration 1 anticipated.

What is **not** fully closed: D3's required fix also asked for "a criterion or a DoD line that
measures the first-upgrade path with an unmodified legacy hook that equals the *previous* shipped
content." AC-PCP-015's commentary argues AC-PCP-005 sub-case (b) covers that population once clause
(i) holds — but that is an argument, not a criterion. See N4.

---

## Part 2 — v0.3.0 audited as a whole

### The core mutant, re-run against the changed AC set

The card's headline mutant — overwrites correctly, backs up correctly, **prints nothing**:

- AC-PCP-004 clause (i): the **warning** writer must contain all three of the backup path, a
  replacement statement, and `pre-commit.local` → **FAILS**
- AC-PCP-006: one run must produce **both** a backup file and a notice naming it → **FAILS**
- AC-PCP-007 clause (i): the warning writer's output must be non-empty → **FAILS**

Still jointly defeated, and the two-writer rewrite strengthened rather than loosened the trap: under
the v0.2.0 single-writer wording AC-PCP-007 was a bare inequality against the success line, which a
whitespace-padded line satisfied; the new clause (i) catches that directly, and AC-PCP-007's own
Mutant field says so. Two additional mutants the split creates are both covered: notice routed to
the *progress* writer (clause (i) inspects the warning writer), and the warning parameter wired to
`out` at the update site (clause (ii)).

### Falsifiability

The v0.3.0 split — ten falsifiable against the untouched tree, five only against a named
implementation mutant (AC-PCP-008, -012, -013, -015, plus -002) — is arithmetically consistent
across `progress.md §E.1`, `acceptance.md` § How to read a criterion, and `§D.3`, and each of the
five carries the label in its own heading or a `Falsification class` clause. Spot-checked the four
behaviour-preserving baselines against the tree: AC-PCP-008 (`:145` returns `ErrUserHookExists` —
verified), AC-PCP-012 (hook exits 0 today), AC-PCP-013 (test present at `:38`, `t.Skipf` at `:45` —
both verified), AC-PCP-015 (all three clauses green at baseline — verified below). The labels are
honest; no criterion passes against the untouched tree while claiming to observe today's behaviour.

AC-PCP-015 is the one new criterion and is correctly labelled green-at-baseline. Its *stated failing
input* ("the tree with M3 applied") is constructible during run-phase. Its stated **check point** is
the problem — see N1.

### The 15th requirement — no ceiling pressure, nothing dropped

`spec-workflow.md:149` → Tier M ceilings are `16` requirements and `16` acceptance criteria, applied
independently. Measured: 15 REQ (`REQ-PCP-001`…`015`, no gap, no duplicate, consistent 3-digit
padding), 15 AC (`AC-PCP-001`…`015`), 15 traceability rows. Within budget on both axes with one slot
spare each. The v0.2.0→v0.3.0 requirement diff (above) shows nothing merged and nothing dropped to
make room.

### Citation integrity under the §A.0 convention

Scanned all four artifacts for `file:line` citations lacking an adjacent `@<sha>`. Eleven matches;
every one resolves its tree, either inline in the same sentence (`init.go:773` … "correct on the
primary checkout at `a1b1ca696`") or under an explicit group prefix (`progress.md §E.1`: "all
`@294b4b6ab`"; `spec.md:427`: "both `@294b4b6ab`"). **No unanchored citation exists.** The grouped
form deviates from the literal `path:line @<sha>` shape §A.0 declares, but it loses no information;
I do not raise it as a defect.

Line anchors re-measured at HEAD — all resolve:

| Citation | Observed |
|---|---|
| `hook_install_precommit.go:26` | `const preCommitHookContent = ` |
| `:126` / `:131` / `:139` / `:145` / `:151` / `:166` / `:172` | `func … InstallPreCommitHook` / `hookDir := …` / `if _, err := os.Stat(hookPath)` / `return ErrUserHookExists` / `os.WriteFile(hookPath, …)` / `func installPreCommitHookOptional(…)` / the success line |
| `hook_install_precommit_test.go:38` / `:45` | `func TestPreCommitTemplateMatchesConstant` / `t.Skipf("template not found …")` |
| `update_template_sync.go:572` / `:575` | prepush call / precommit call |
| `init.go:895` / `:898` | prepush call / precommit call |

§A.2 baseline table: all seven values reproduce (`0`, `0`, `0`, `3245`, `179`, `575`, `898`).

### Paired-edit constraint

`TestPreCommitTemplateMatchesConstant` exists and does what the SPEC says: it reads the template,
compares `string(templateBytes) != preCommitHookContent`, and `t.Fatalf`s on divergence — with a
`t.Skipf` escape at `:45` when the template file is absent. Every requirement touching the hook body
(REQ-PCP-011, -013, -015) is bound by an AC (AC-PCP-011, -013, -015 clause (iii)), not by prose.
AC-PCP-013 requires PASS and rejects SKIP, closing the live hazard. AC-PCP-015's Mutant field
correctly identifies that clause (i) alone (which compares the *template*) is defeated by a
constant-only edit, and routes the defence through clause (iii). That is a well-constructed
criterion — the flaw is when it runs, not what it asserts.

### Requirement C closure — re-run

The recorded sweep (`exec [0-9][0-9]*[<>]` over `*.sh`, `*.tmpl`, `pre-commit`, `pre-push`, `*.go`,
`*.md`, excluding `./.git/`) returns **0 hits** in this tree. The closure is correct.

Under the looser reading (`exec [0-9]*[<>]`, zero or more digits) the sweep now returns **9** hits,
not the 6 §A.4 records; all nine are `printf … | exec <bin>` command-replacement prose in
`handle-stop-goal.sh`, its `.tmpl` twin, `stop_goal_single_exec_test.go`, and report/spec text. None
is a file-descriptor redirection, so the closure stands under either reading. The count drift is N6.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-PCP-001`…`REQ-PCP-015`, 15 bullet definitions, no gap,
  no duplicate, uniform 3-digit padding. The §B presentation order is thematic, which is grouping,
  not a numbering defect.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-PCP-*` in
  `spec.md §B`); ACs graded under Group 4. All 15 match a GEARS pattern: Ubiquitous (001, 013, 014,
  015), Event-driven (002, 003, 004, 005, 011, 012), Unwanted / `shall not` (006, 007, 009, 010).
  Two non-agent subjects (N7) are precision, not compliance.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`id`, `title`, `version: "0.3.0"`, `status: draft`, `created`/`updated` ISO, `author`,
  `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` CSV string) plus optional
  `tier: M`. No rejected snake_case alias. Corroborated: `moai spec lint` on the SPEC →
  `✓ No findings — all SPEC documents are valid`.
- **[PASS] MP-4 language neutrality** — reaches template-distributed content
  (`internal/template/templates/.git_hooks/pre-commit`, REQ-PCP-011); passes because the
  `pre-commit.local` delegation is shell-level and language-agnostic and no requirement names a
  language-specific tool as primary.
- **[N/A] MP-5 D7 cross-SPEC reconciliation** — extracting `SPEC-([A-Z][A-Z0-9]+-)+[0-9]+` across all
  four artifacts returns exactly one id, the SPEC's own. The D7 verb has no input. N/A per the MP-4
  precedent.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — searching the SPEC directory for `NEEDS CLARIFICATION` returns
  rc 1, no match.

No must-pass failure. The M5 firewall does not force FAIL.

---

## Category Scores

| Dimension | Score | Δ | Rubric Band | Evidence |
|-----------|-------|---|-------------|----------|
| Clarity | 0.85 | +0.05 | 0.75-1.0 | The two iteration-1 ambiguities are gone: REQ-PCP-010 states the precedence in bold, REQ-PCP-004 states the writer split and both bindings. Deducted for N2 — the new REQ-PCP-015 and its AC name the same anchor three different ways, and the requirement's own wording is the reading its criterion says to reject. |
| Completeness | 0.85 | 0.00 | 0.75-1.0 | Decision 3 adds a genuine release-composition analysis with measured magnitude; §D.1 gained the D5 row; §A.4 gained the pattern semantics. Offset by N1 (no enforcement point for the requirement Decision 3 rests on) and N3 (§C.2's collision analysis not updated for the release-level constraint the same version introduced). |
| Testability | 0.85 | 0.00 | 0.75-1.0 | Real gains: AC-PCP-010(a)'s post-state clause, AC-PCP-004/007's two-writer clauses with named mutants, AC-PCP-002's honest falsification class. Offset by the new criterion being the weakest in the set — AC-PCP-015 is checked where it cannot fail (N1), covers only one of REQ-PCP-015's two obligations, and the first-upgrade population it argues about is never measured (N4). AC-PCP-004(ii)'s stated Decides command is not self-deciding (N5). |
| Traceability | 1.00 | 0.00 | 1.0 | 15 REQ ↔ 15 AC, 1:1, verified by count and by reading the §E matrix. No orphaned AC, no uncovered REQ. REQ-PCP-015's untested second obligation is a coverage gap inside a mapped pair, scored under Testability, not here. |

Aggregate = (0.85 + 0.85 + 0.85 + 1.00) / 4 = **0.8875** ≥ 0.80.

---

## Defects Found

**N1 — `spec.md` REQ-PCP-015 + `acceptance.md` AC-PCP-015 + `plan.md §F` / `acceptance.md §D.3` — the release constraint is checked at the one point where it cannot fail, and the plan schedules the violating commit into the same branch by default — Severity: critical — Class: blocking**

REQ-PCP-015 binds a **release**. Everything that enforces it binds something else:

1. `plan.md §F` places the check "at the end of M2" — before M3 exists on the branch. At that point
   clauses (i)/(ii)/(iii) are green **by construction**, which the AC's own baseline concedes
   ("behaviour-preserving by construction"). A criterion evaluated only where it cannot fail
   observes nothing about the thing it binds.
2. M3 is a milestone of **this** SPEC's run-phase (`plan.md §F` M3: "Land as its own commit, last";
   "M3's milestone content above is unchanged; only its release target moves"). So the default
   run-phase outcome is that the `pre-commit.local` body change merges to `main` alongside the
   classifier. The repo cuts a release by merging a `release/vX.Y.Z` PR and then tagging
   (`.moai/docs/version-management.md:87`), so a landed commit ships in the next tag unless someone
   deliberately cuts the release branch from a point before it. Nothing in the four artifacts names
   that mechanism — not a branch, not a hold, not a cherry-pick, not "defer M3 to a later SPEC
   run-phase". The SPEC distinguishes *land* from *ship* (`plan.md §D`) and then never says how the
   landed commit is kept out of the release.
3. The only release-time gate is a DoD sentence: "AC-PCP-015 is checked against the release
   candidate before the classifier release ships." No owner, no trigger, no automation. Measured:
   listing files under `.github/workflows/` and `scripts/` that reference
   `git_hooks/pre-commit` returns **no match** (rc 1). No CI job or release script touches the hook
   body, so nothing will re-run the check at tag time.
4. That same DoD line makes the Definition of Done unevaluable at SPEC close: the SPEC closes at
   sync-phase, before any release exists, yet DoD requires "REQ-PCP-011's paired edit lands … **in a
   later release than the rest of this SPEC**". Either the SPEC cannot close, or the item is marked
   done on a promise.

So the answer to "mechanically checkable, or resting on a human remembering" is: the *command* is
mechanical and well-built; its *invocation at the moment that matters* rests entirely on a human
remembering, and the plan makes the violation the default rather than the accident. This is the same
failure class the SPEC is elsewhere disciplined about — `plan.md §G` lists "Folding M3 into the
M1/M2 release" as an anti-pattern and asserts "REQ-PCP-015 / AC-PCP-015 exist to catch this", which
is an unobserved claim: as scheduled, they do not catch it.

**Required fix** — any one of these closes it, and the SPEC should pick one explicitly:
(a) remove M3 from this SPEC's run-phase entirely and record it as a successor card, so the branch
physically cannot carry the body change; or (b) state the release mechanism in `plan.md` — M3 lands
on a branch that is not merged until after the `v3.1.4` tag, named as such; or (c) add a re-check
point that fires *after* M3 lands (a release-time or tag-time gate, with a named owner). Whichever is
chosen, restate the AC-PCP-015 check point so it is run somewhere it can go red, and reconcile the
DoD line so SPEC close does not depend on an event that post-dates it.

**N2 — `spec.md` REQ-PCP-015 ⊥ `acceptance.md` AC-PCP-015 (Then vs Mutant) ⊥ `spec.md §A.5` Decision 3 / `plan.md §C`, §F — "the last released hook body" has three referents, and the requirement's own wording is the one its criterion says to reject — Severity: major — Class: blocking**

Four surfaces name the anchor four ways:

| Surface | Wording | Resolves to |
|---|---|---|
| REQ-PCP-015 | "byte-identical to the **most recently released** hook body" | the latest tag at release time |
| AC-PCP-015 Then (i) | `git show <last-released-tag>:…` | unresolved placeholder |
| AC-PCP-015 Mutant | "the last release **whose hook body predates this SPEC** (`v3.1.2` at authoring time)" | `v3.1.2`, fixed |
| §A.5 Decision 3, `plan.md §F`, `plan.md §C` pre-flight step 5, AC-PCP-015 Baseline | `v3.1.2` | `v3.1.2`, fixed |

AC-PCP-015's Mutant field explicitly rejects the "latest tag" reading as a defeatable mutant
("re-pointing `<last-released-tag>` at the current release after landing M3, making the comparison
trivially true"). REQ-PCP-015's wording **is** that reading. A test author who implements the
requirement as written builds the mutant the criterion exists to defeat.

Latent, not live, today — measured: the tag list `v3.*` sorted ends `v3.1.0-rc.2`, `v3.1.1`,
`v3.1.2`, so `v3.1.2` is the latest tag; and the in-flight `release/v3.1.3` branch (`b37e86b64`)
carries a hook body byte-identical to HEAD (`cmp` rc 0), so all four readings coincide right now.
They diverge the moment any release ships a changed body — which is exactly the event REQ-PCP-015
exists to reason about.

**Required fix**: name one anchor in REQ-PCP-015 and repeat it verbatim in AC-PCP-015's Then. The
Mutant field already contains the right formulation ("the last release whose hook body predates this
SPEC — `v3.1.2`"); promote it into both.

**N3 — `spec.md §C.2` ("Only REQ-PCP-011 collides") is stale against REQ-PCP-015, and no rule says which yields — Severity: major — Class: blocking**

REQ-PCP-015 does not constrain this SPEC's diff — it constrains the **whole release**: AC-PCP-015
clause (i) compares the shipped template against the tag, so *any* hook-body change in the
classifier release makes it red, whoever authored it. §C.2 was written before Decision 3 and still
says "Only REQ-PCP-011 collides", then argues the collision window narrows to a separate release
cycle. Under REQ-PCP-015 the opposite holds for the classifier release: card t237 / issue #1641
edits the same constant/template pair, and if it lands in `v3.1.4` this SPEC's own criterion declares
that release unshippable — while `§D.3` says "a release candidate that fails it is not shippable",
with no conflict-resolution rule and no owner.

Measured: t237 has **not** landed — searching the last 300 commits of `origin/main` for `1641` or
`t237` returns no match (rc 1). The collision is live, and `plan.md §C` step 4 handles only the
rebase axis, not the release axis.

**Required fix**: state in §C.2 (and in `plan.md §C`) that REQ-PCP-015 binds the release, not the
diff, and say what happens when an unrelated card needs a hook-body change in the classifier
release — most plausibly: the classifier release moves, or t237's body change moves, decided by
whoever holds the release. Silence here means the first person to hit it resolves it by weakening
AC-PCP-015, which is the erosion this SPEC is otherwise built to prevent.

**N4 — the first-upgrade population is covered by argument, not by a criterion — Severity: major — Class: blocking (D3's residual)**

AC-PCP-015's commentary claims it "covers the first-upgrade path with an unmodified legacy hook,
which nothing covered before", reasoning that when clause (i) holds, a `v3.1.0`-`v3.1.2` user's
installed hook equals the incoming content and therefore runs AC-PCP-005 sub-case (b). The reasoning
is correct, but it is prose in a commentary block, and no criterion executes it.

AC-PCP-005 sub-case (b)'s Given is "content identical to **the incoming content**" — a synthetic
fixture. It never says "content equal to the `v3.1.2` shipped body". So sub-case (b) passes with a
hand-built fixture regardless of whether the real installed base is covered, and the coupling
between the two criteria exists only through REQ-PCP-015 — the requirement N1 shows is unenforced at
the point it binds. The one population whose silence Decision 3 was chosen to buy is therefore never
measured against real bytes, in a SPEC whose whole discipline is "a check never seen red has not
been shown to observe anything".

**Required fix**: add a sub-case to AC-PCP-005 (or a fourth clause to AC-PCP-015) whose installed
fixture is the `v3.1.2` blob of the template rather than the incoming constant, with the expected
outcome "no backup, no notice, record written". It is a fixture change, not a new mechanism, and it
converts the argument into a check.

**N5 — `acceptance.md` AC-PCP-004 clause (ii)'s stated Decides command cannot decide it — Severity: minor — Class: optional**

Clause (ii) is decided by "`grep -n 'installPreCommitHookOptional(' internal/cli/update_template_sync.go internal/cli/init.go` showing an `ErrOrStderr`-derived writer in the warning-writer position at each site."

Measured, the update call site under the *correct* implementation reads
`installPreCommitHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out, errOut)` — the line
contains `errOut`, not `ErrOrStderr`; the derivation lives two lines away at `:72`. The grep output
alone therefore cannot distinguish the correct implementation from one passing an arbitrary
identically-named variable, and a checker grepping for `ErrOrStderr` finds nothing at the update site
even when the wiring is right. The criterion's own escape hatch ("a Go test asserting the same is
equally acceptable") is the mechanical route.

**Required fix**: make the Go test the primary Decides for clause (ii), or extend the grep to cover
the writer's derivation (`grep -n 'errOut :=' …` alongside the call-site line).

**N6 — `spec.md §A.4`'s "6 hits" for the loose glob is now 9 in this tree — Severity: minor — Class: optional**

The 0-hit closure under the correct pattern reproduces exactly. The parenthetical count for the
*looser* pattern does not: it is now 9, because §A.4's own explanatory text, `spec.md:421`, and the
iteration-1 audit report are themselves matches. A self-counting figure drifts on every edit.

**Required fix**: drop the numeral and keep the qualitative statement ("all matches are
`printf … | exec <bin>` command-replacement prose, none a redirection"), or anchor it with a `@<sha>`
per §A.0.

**N7 — `spec.md §B` — two requirements take a non-agent subject, and REQ-PCP-015 bundles two obligations of which its AC tests one — Severity: minor — Class: optional**

REQ-PCP-010's lead clause ("**Failure of a supporting write** shall never fail `moai init` or
`moai update`") and REQ-PCP-015 ("**The release** … shall leave …") place the `shall` on a condition
and on an artifact rather than on the installer. Both still match a GEARS pattern under the
generalized `<subject>` rule, so MP-2 passes; this is precision. REQ-PCP-015 additionally carries two
obligations in one requirement — body-identity in the classifier release, and REQ-PCP-011 shipping
later — and AC-PCP-015 tests only the first. That is the traceability shape under which N1 hides.

**Required fix**: optional — rephrase both onto the installer / the release process, and either
split REQ-PCP-015 or add a clause to AC-PCP-015 covering the second obligation (which N1's fix will
likely supply anyway).

---

## Recommendation

**Fit to enter run-phase: partially — M1 and M2 yes, M3 no.**

None of N1-N4 touches the classifier (M1) or the disclosure surface (M2). Both iteration-1 blocking
defects that gated M2 are genuinely closed, by strengthening rather than by weakening, and I
re-verified each against the tree rather than against the SPEC's word. The artifacts as they stand
are sufficient to implement REQ-PCP-001 through REQ-PCP-010 and REQ-PCP-014.

All four new blocking findings sit on the release-composition design added at v0.3.0, and all four
gate the same moment: **M3 must not land until they close.**

1. **N1 (critical)** — give REQ-PCP-015 an enforcement point that can go red, and stop scheduling the
   violating commit into the release-bound branch by default. Today `plan.md §G` claims
   "REQ-PCP-015 / AC-PCP-015 exist to catch this"; as scheduled they do not, and no CI or release
   script does either (measured: no workflow or script references the hook template).
2. **N2 (major)** — collapse the three referents for "the last released hook body" into one, using
   the formulation AC-PCP-015's own Mutant field already contains.
3. **N3 (major)** — update §C.2: REQ-PCP-015 binds the release, not the diff, so t237's body change
   collides at the release level. State which yields.
4. **N4 (major)** — convert AC-PCP-015's first-upgrade argument into a criterion with a `v3.1.2`
   fixture. This is the one item of D3's required fix that was closed by prose.

N5-N7 are optional; folding them into whichever edit closes N1-N4 costs nothing, and none affects
correctness.

**Iteration ceiling.** This is iteration 2 of the Tier M ceiling of 2
(`harness.plan_audit_tier_ceilings`, S=1 / M=2 / L=3). A further plan-audit iteration on this SPEC is
outside the Tier M budget, so the four blocking findings above are routed as a gated PASS-WITH-DEBT
rather than into a third audit round: the score clears the threshold, movement is monotonic
(0.875 → 0.8875), and no must-pass criterion failed. If a confirming re-audit of the N1-N4 delta is
wanted, it requires a Tier reclassification or an explicit user override of the ceiling — not an
auditor decision.
