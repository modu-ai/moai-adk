# SPEC Review Report: SPEC-PROJECT-CONTINUATION-KEY-001 (card t191)

Iteration: 1/2 (Tier M ceiling — `.moai/config/sections/harness.yaml:77` `plan_audit_tier_ceilings.M: 2`)
Verdict: **FAIL**
Overall Score: **0.78** (harmonic mean of the four rubric dimensions; Tier M PASS threshold **0.80** per `spec-workflow.md:141`)

Reasoning context ignored per M1 Context Isolation. Audited from `spec.md`, `plan.md`,
`acceptance.md` (Tier M input contract) plus `progress.md` and the source tree, at worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t191`, branch `WT-project-continuation`,
HEAD `8febfb8af`, base `2660bcd09` (`git rev-list --count --left-right 2660bcd09...HEAD` → `0	1`).

**FAIL is a near miss, not a rejection of the approach.** The design work in §3 is unusually
strong: D3, D4 and D6 are measured rather than asserted, and every figure in §1.1 and §3 D6 that
I re-ran reproduced exactly (§ Baseline re-measurement below). Three blocking defects stop it,
and two of them — D1 and D2 — would produce a wrong implementation rather than a merely
incomplete one, because they misdirect the implementer at the two places the plan says to
decide.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-PCK-001` … `REQ-PCK-011`, sequential, no gap, no
  duplicate, uniform 3-digit padding. Definitions at `spec.md:78-88`, one per line, verified by
  `grep -o "REQ-PCK-[0-9]\{3\}" spec.md | sort -u` → exactly 11 tokens `001`…`011`.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §2),
  never against the ACs. All 11 carry an explicit pattern label and match it:
  Ubiquitous (`001`, `002`, `007`, `008`, `011` — "The … shall"); Event-driven (`003`
  "When the configured value matches no token …, the resolver shall …", `010` "When the wizard
  presents …, it shall …"); State-driven (`004`, `005`, `006` — "While the resolved value is
  `<token>`, the Phase 14 workflow shall …"); Where/static-config-gate (`009` — "Where the
  distributed template carries `workflow.project.continuation`, its shipped value shall be
  `card`"). No informal "should"/"may" in normative text; no Given-When-Then presented as a REQ.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-14`): `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created`/`updated` ISO `2026-09-02`, `author`, `priority: P2`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at` / `updated_at` / `labels` / `spec_id`) present. `tier: M` and `related_specs`
  are additive and do not violate the schema.
- **[N/A] MP-4 Section 22 language neutrality** — the SPEC is scoped to this Go repository's own
  `internal/**` and its distributed template; it names no multi-language tooling matrix.
  N/A auto-passes per the MP-4 clause.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — three referenced SPEC-IDs, all present, none in
  `{retired, superseded, archived}`:

  ```
  $ for SID in SPEC-SYNC-STRATEGY-KEY-001 SPEC-TODO-ENABLE-FLAG-001 SPEC-V3R3-PROJECT-HARNESS-001; do ... done
  SPEC-SYNC-STRATEGY-KEY-001 EXISTS status=status: completed
  SPEC-TODO-ENABLE-FLAG-001 EXISTS status=status: completed
  SPEC-V3R3-PROJECT-HARNESS-001 EXISTS status=status: completed
  ```

  No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. D8 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-PROJECT-CONTINUATION-KEY-001/`
  → no match, `rc=1`. No open marker in `plan.md` (there is no `research.md`; Tier M does not
  require one).

All seven must-pass criteria clear. The FAIL is produced by the rubric score falling below the
Tier M threshold, not by the firewall.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.65 | between 0.50 and 0.75 | `REQ-PCK-006` (`spec.md:83`) does not state any behaviour distinguishing `pipeline` from `card` — D1 below. `REQ-PCK-011` (`spec.md:88`) "labels and descriptions" is resolved one way by `AC-PCK-011` and another by `plan.md:93` — D5. The remaining nine requirements are unambiguous. |
| Completeness | 0.85 | 0.75-1.0 | Every required section present: HISTORY (`spec.md:23`), WHY (§1 `spec.md:31`), WHAT (§2 `spec.md:74`), HOW (§3 `spec.md:92` + `plan.md` §F), REQUIREMENTS (§2), ACCEPTANCE CRITERIA (`acceptance.md`), Out of Scope — three `### Out of Scope — <topic>` H3s at `spec.md:160,166,172`, each with specific `-` bullets. §5 Gaps is unusually disciplined; it loses points because one of its five entries is factually wrong (D2). |
| Testability | 0.70 | 0.50-0.75 | 13 ACs, each naming a deciding command. `AC-PCK-008`'s first conjunct has no reachable red (D3); `AC-PCK-006`'s negative conjunct passes trivially; `AC-PCK-010` asserts less than `REQ-PCK-010` requires. The other ten have reachable reds, established individually below. |
| Traceability | 1.00 | 1.0 | `acceptance.md:10-24` matrix maps all 13 ACs to REQs. Every `REQ-PCK-001..011` appears in ≥1 AC row; every AC names a REQ that exists. No orphan, no uncovered REQ. `§D.2`/`§D.3` state the negative-verification and indirect-verification cases explicitly rather than hiding them. |

Aggregate: harmonic mean 4 / (1/0.65 + 1/0.85 + 1/0.70 + 1/1.00) = **0.78**.

---

## Hunt 1 — the D1 narrowing, both directions

**Both failure modes are present, and they share one root cause.** D1 narrowed `pipeline` until
its only remaining content is a sentence, and then constrained neither what that sentence must
say nor what it must not imply.

### (b) `pipeline` and `card` are behaviourally indistinguishable — this is the stronger finding

Read `REQ-PCK-005` and `REQ-PCK-006` against each other (`spec.md:82-83`):

- `card`: "shall issue exactly one derived `[PROJECT] ` card under the five standing-source
  properties and shall present `Create the SPEC and start now` as the recommended option."
- `pipeline`: "shall issue the card **exactly as under `card`**, and shall present as its
  recommended option the branch that names continuation through run-phase implementation and
  tests."

Card issuance: identical by the SPEC's own words. Gate: identical (`REQ-PCK-007`, `REQ-PCK-008`).
Recommended slot: occupied in both cases — and occupied by a *continuation* branch in both cases,
because the shipped `card` recommended option already is one. Measured verbatim at
`doc-generation.md:350`:

```
- Create the SPEC and start now (Recommended): Pick the card issued in Step 4.1.5 and begin
  immediately. … record it with `moai todo next <n>` and continue in this same session with
  `/moai plan "<card text>"`. … Run-phase entry still passes the Implementation Kickoff Approval
  gate … this branch never bypasses that gate.
```

That option already means "begin immediately and continue in this session". The entire delta
`pipeline` delivers over `card` is that its option text additionally *names* run-phase
implementation and tests. `AC-PCK-006` (`acceptance.md:58-62`) confirms this is all that is
verified — its Then clause is wholly about text: "names a recommended option whose text
references continuation through run-phase implementation and tests".

So the SPEC ships a third enum value whose complete observable effect is one string, and the AC
that claims to verify it is testing wording alone. That may still be a legitimate product
decision — but the SPEC never says so. D1's "What is given up, stated honestly" paragraph
(`spec.md:107`) measures the delta from the *card brief's literal reading*, not from `card`:
"it makes the continuation branch the pre-selected, recommended one" is already true under
`card`. Nowhere does the document state what `pipeline` does that `card` does not. A reader
approving this SPEC cannot tell whether they are approving a third value or a synonym.

### (a) The narrowing also leaks, at exactly the place D1 was written to protect

`REQ-PCK-006` authorizes a **new** recommended-option text. The `card` option it replaces carries
the kickoff-gate sentence quoted above — the only occurrence of "Implementation Kickoff Approval"
in the whole file:

```
$ grep -c "Implementation Kickoff Approval" .claude/skills/moai/workflows/project/doc-generation.md
1
```

Nothing in `REQ-PCK-006`, `REQ-PCK-008`, or `AC-PCK-006` obliges the `pipeline` option text to
carry that sentence. `AC-PCK-006`'s negative conjunct — "the row contains no instruction to
select, answer, skip, or default that option" — polices the presentation *mechanism* and says
nothing about what the option *promises the operator*.

The consequence is concrete: under `pipeline` the recommended option tells the operator (and the
orchestrator reading the prose) that this branch continues through implementation and tests,
with no gate reminder attached, at precisely the position where the `card` branch has one. An
orchestrator that reads "this branch means continuation through run-phase" and finds no
countervailing clause has been handed a reason to treat kickoff as answered by the branch
selection. That is `pipeline` reaching a decision the operator was supposed to make — the exact
failure D1 exists to prevent — and the AC set does not close it.

Note the interaction with (b): `AC-PCK-008`'s regression conjunct (`grep -c … ≥ 1`) currently
survives only because the `card` option retains the sentence. It would not detect the
`pipeline` option lacking one.

---

## Hunt 2 — per-AC reachable-red probe

`AC-PCK-008` is confirmed as the vacuity case, and it is the only one. Every other AC has a
reachable red, established individually.

**AC-PCK-008 — first conjunct has no reachable red.** I ran its command verbatim, unmodified,
at plan time before any implementation exists:

```
$ git diff --stat origin/develop...HEAD -- .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/goal-directive.md .claude/skills/moai/workflows/goal.md
rc=0
```

Empty. The criterion is **already satisfied**, and it was satisfied before the SPEC was written.
Its red state requires the implementer to affirmatively edit one of three files that no milestone
touches — `plan.md` §F M1-M6 name `doc-generation.md`, `project.md`, `types.go`, `defaults.go`,
`closed_sets.go`, `project_continuation.go`, the template and local `workflow.yaml`,
`shipped_key_inventory.yaml`, `questions.go`, `translations.go`, `types.go`, `wizard.go`,
`init.go`, `initializer{,_expansion}.go`, `schema_sections.go`, `i18n.js`, and `CHANGELOG.md`.
None of the three gate-owning files appears. This is the "AC that cannot fail through any
plausible implementation" class that `.moai/reports/t331/plan-audit-iter2.md` flagged twice.

The second conjunct is **not** vacuous: pre-change value measured as `1` (above), and M1 rewrites
the Step 4.2 option block that contains the sole occurrence, so an implementer replacing that
option text drops it to `0` and the criterion goes red. `AC-PCK-008` is therefore half-vacuous,
not wholly so — but its half-vacuous conjunct is the one that carries `REQ-PCK-008`'s weight,
and the surviving conjunct only guards the `card` branch (Hunt 1a).

Per-AC results:

| AC | Reachable red? | How established |
|---|---|---|
| 001 | **Yes** | `grep -rn "ProjectContinuation" internal/ --include='*.go' \| wc -l` → `0`. No accessor exists; a slice of the wrong three tokens fails the assert. |
| 002 | **Yes** | Same measurement — no resolver exists. A resolver returning `""` on absent, or panicking on a nil `*Config`, fails. |
| 003 | **Yes** | A resolver that returns `unmatched == ""` on `pipelien` fails the second conjunct; both conjuncts are positive assertions on values the implementer chooses. |
| 004 | **Yes** | `grep -ci "continuation" doc-generation.md` → `0` today. The `none` row does not exist; M1 must author it, and a row that fails to skip issuance fails the read. |
| 005 | **Yes** | Genuine regression guard on a file M1 **does** edit. The five standing-source steps are at `doc-generation.md:333-341`; M1 rewrites this region, so damaging them is a live risk. |
| 006 | **Partial** | Positive conjunct (text names continuation-through-run) has a reachable red. Negative conjunct ("no instruction to select, answer, skip, or default") passes trivially for any row that simply says nothing — it cannot distinguish a compliant row from a silent one. |
| 007 | **Yes** | `grep -n "No branch is taken on the operator's behalf"` matches `doc-generation.md:360` today; the new value-independence sentence does not exist. M1 rewrites the region, so both clause presence and scope are live. |
| 008 | **No (first conjunct)** / Yes (second) | Measured above. |
| 009 | **Yes — mutant verified in source** | `shipped_key_reader_test.go:99-110`: `for path := range shippedKeys { if _, ok := allowlist[path]; !ok { untriaged = append(...) } }` then `t.Errorf("REQ-CKH-008 anti-rot: %d shipped config key(s) are NOT in the triage inventory …")`. `loadAllowlistWithCount` (`:296-314`) keys the allowlist on `e.Path` alone. Removing the inventory row while the template ships the key therefore fails the test. The claimed mutant is real. |
| 010 | **Yes**, and stronger than the SPEC knows | See Hunt 3 — `TestWizardQuestionTranslationCompleteness` fails automatically. But `AC-PCK-010` itself asserts only per-locale Title non-emptiness and distinctness, which under-verifies `REQ-PCK-010`'s option-description clause. |
| 011 | **Yes** | `internal/web/i18n_governance_test.go` carries `TestI18nKeyCoverageForward` (`:408`) and `TestI18nKeyCoverageReverse` (`:423`) plus `TestI18nUntranslatedValues` (`:217`). A `FieldDef` whose i18n keys are absent from a locale map fails coverage; `grep -c "f.workflow.project" internal/web/assets/i18n.js` → `0` today. |
| 012 | **Yes** | Measured today: `cmp` rc `0/0/0/1` across the four pairs. A template-only edit without `make build` + mirror flips one of the three zeros. |
| 013 | **Yes** | The report line does not exist (`grep -ci "continuation" doc-generation.md` → `0`). Both branches of the conditional are positive assertions. |

**Proposed replacement for `AC-PCK-008`.** Delete the diff-stat conjunct and replace it with a
criterion whose red is reachable by the implementation the plan actually describes:

> **Given** `doc-generation.md` on the implementation branch,
> **When** the Step 4.2 recommended-option table is read for all three values,
> **Then** each of the three recommended-option texts (`none`, `card`, `pipeline`) that offers
> run-phase entry carries the Implementation Kickoff Approval sentence verbatim, and
> `grep -c "Implementation Kickoff Approval" doc-generation.md` returns a count **no lower than
> the number of such options** (pre-change value `1`, measured on `2660bcd09`).

That criterion goes red on exactly the leak Hunt 1(a) identifies, whereas the deleted conjunct
goes red on nothing the plan contemplates. Keeping the diff-stat check as a *non-blocking*
hygiene assertion is defensible; presenting it as the blocking verification of `REQ-PCK-008`
is not.

---

## Hunt 3 — the wizard localization gap is not a gap; the SPEC's claim is false

`spec.md:186` and `plan.md:17` both assert that no localized-option-description precedent exists
and that `GetLocalizedQuestion` may need extending. **Both statements are contradicted by
source.**

`GetLocalizedQuestion` already translates option labels and descriptions
(`internal/cli/wizard/translations.go:571-586`):

```go
	// Translate options if available
	if len(trans.Options) > 0 && len(q.Options) == len(trans.Options) {
		localized.Options = make([]Option, len(q.Options))
		for i, opt := range q.Options {
			localized.Options[i] = Option{
				Label: trans.Options[i].Label,
				Value: opt.Value, // Keep original value
				Desc:  trans.Options[i].Desc,
			}
```

The SPEC's premise — "the only `QuestionTypeSelect` located (`questions.go:65`,
`conversation_language`)" — is a sampling error. There are **twelve**:

```
$ grep -rn "QuestionTypeSelect" internal/cli/wizard/questions.go
internal/cli/wizard/questions.go:65:   internal/cli/wizard/questions.go:102:  internal/cli/wizard/questions.go:132:
internal/cli/wizard/questions.go:168:  internal/cli/wizard/questions.go:183:  internal/cli/wizard/questions.go:355:
internal/cli/wizard/questions.go:403:  internal/cli/wizard/questions.go:418:  internal/cli/wizard/questions.go:431:
internal/cli/wizard/questions.go:444:  internal/cli/wizard/questions.go:474:  internal/cli/wizard/questions.go:497:
```

The SPEC picked the one question that is **deliberately exempt** from option translation and
generalized the exemption into an absence. `internal/cli/wizard/translations_completeness_test.go:9-15`
names it as such:

```go
// optionTranslationExemptIDs are Select questions whose options are
// intentionally NOT carried in the fixed-length translation table:
//   - conversation_language: option labels are native language names, never
//     translated (GetLocalizedQuestion leaves them untouched).
var optionTranslationExemptIDs = map[string]bool{
	"conversation_language": true,
}
```

The precedent the SPEC says does not exist is `audit_model` (`translations.go:137-146`, ko;
`:296-305`, ja; `:455`, zh) — a closed-set select with a per-option `Label` and `Desc` in every
locale, structurally identical to what `project_continuation` needs.

**The consequence is worse than a wrong gap entry.** `plan.md:17` authorizes a fallback — "the
option descriptions collapse into the question `Description`" — and `acceptance.md:129` blesses
it: "if it resolves by folding descriptions into the question body, `AC-PCK-010` stands as
written". That path **fails an existing test**. `TestWizardQuestionTranslationCompleteness`
(`translations_completeness_test.go:95-134`) iterates every question from `InitQuestions(root)`
and, for any non-exempt `QuestionTypeSelect`, requires:

```go
			if len(trans.Options) != len(q.Options) {
				t.Errorf("locale %q: question %q has %d option translations, want %d", ...)
			}
			for j, opt := range trans.Options {
				if opt.Label == "" { t.Errorf("locale %q: question %q option %d has empty label", ...) }
				if opt.Desc == "" { t.Errorf("locale %q: question %q option %d has empty desc", ...) }
```

for each of `ko`, `ja`, `zh` (`:7`). A new `project_continuation` select with no option
translations produces nine errors. The plan directs the implementer toward a known red state and
the acceptance criteria pre-approve the outcome.

**Judgment: `REQ-PCK-010` is deliverable exactly as written** — better than the SPEC believes.
It needs no new helper and no restating. What must change is `spec.md` §5, `plan.md` §B item 1,
`acceptance.md` §D.5, and `AC-PCK-010`, which must assert the option Label/Desc obligation
`REQ-PCK-010` actually states and name the existing test as its enforcer.

---

## Hunt 4 — D5 and AC-PCK-009 are correct; verified in source, not accepted from description

`spec.md:139`'s description of the mechanism holds. The guard's anti-rot gate is
`shipped_key_reader_test.go:99-110` (quoted in the Hunt 2 table); the allowlist loader is
`:296-314` and keys on `e.Path`. The cited precedent row exists:

```
$ sed -n '2918,2926p' internal/config/testdata/shipped_key_inventory.yaml
- path: "workflow.worktree.auto_cleanup"
  class: P
  evidence: .claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md
```

Class `P` with a skill-path evidence field — the shape `spec.md:139` claims, at lines 2921-2923
(the SPEC cites 2922-2924; off by one, D6 below). The inventory size reproduces exactly:
`grep -c "^- path:" …` → `975`, matching `spec.md:60`.

`AC-PCK-009`'s mutant is genuinely non-vacuous: the shipped-key set is enumerated from
git-tracked template YAMLs, so shipping `continuation: card` in the template `workflow.yaml`
without the inventory row lands the key in `untriaged` and calls `t.Errorf`. The claim is
sound. **No defect here.**

One diagnostic nuance, non-blocking: `plan.md` M2 builds a `ProjectContinuation()` accessor, but
the production consumer is orchestrator prose, so the key may classify `unresolved`/`unbound`
rather than prose-consumed. `shipped_key_reader_test.go:112-118` states these "do not fail the
guard (the triage decision stands)", so class `P` remains correct and the milestone is safe.

---

## Hunt 5 — baseline re-measurement

Every ref-dependent figure re-run in this worktree. **All reproduce**, with one scoping note.

| Figure | SPEC claim | Re-measured | Verdict |
|---|---|---|---|
| §1.1 key grep | 6, all under `.moai/reports/t332/` | `git grep -n … 2660bcd09` → `6`, in `.moai/reports/t332/cards/batch-1.md` (3) and `.moai/reports/t332/input/all-cards.md` (3) | **Reproduces at the declared tree.** At `HEAD` it reads `17` because the SPEC's own artifacts and `plan-baselines.md` now match. Scoping note below. |
| §1.1 Go binding | 0 | `grep -rn "ProjectContinuation" internal/ --include='*.go' \| wc -l` → `0` | Reproduces |
| §1.1 i18n rows | 0 | `grep -c "f.workflow.project" internal/web/assets/i18n.js` → `0` | Reproduces |
| §1.1 wizard question | 0 | `grep -rn "project_continuation" internal/cli/wizard/ \| wc -l` → `0` | Reproduces |
| §1.1 prose consumer | 0 | `grep -ci "continuation" doc-generation.md` → `0` | Reproduces |
| §1.1 inventory size | 975 | `grep -c "^- path:" …` → `975` | Reproduces |
| §1 P1 diffstat | 9 files, 165 insertions, 14 deletions | `git show --stat --format="" e91def4ca` → identical block, `9 files changed, 165 insertions(+), 14 deletions(-)` | Reproduces byte-for-byte |
| §3 D4 pre-P1 wording | "Create SPEC (Recommended): Run `/moai plan` …" | `git show e91def4ca --format="" -- doc-generation.md` shows that exact line as a `-` deletion; the two [HARD] clauses appear as `+` additions | Reproduces exactly |
| §3 D6 `cmp` matrix | 0 / 0 / 0 / 1 (`differ: char 17, line 2`) | `tab_schema rc=0`, `doc-generation rc=0`, `todo rc=0`, `workflow.yaml differ: char 17, line 2` rc=1 | Reproduces exactly |
| §3 D6 no i18n.js twin | `find` returned nothing | `find internal/template/templates -name 'i18n.js' \| wc -l` → `0` | Reproduces |
| §3 D5 inventory row cite | lines 2922-2924 | actual 2921-2923 | **Off by one** — D6 below |

**Divergence note the SPEC does not carry.** Local `develop` is at `2660bcd09` — the SPEC's
declared baseline — but `origin/develop` has advanced to `f7cabfc29`:

```
$ git rev-list --count --left-right origin/develop...HEAD
0	35
```

`HEAD` is 35 commits ahead of `origin/develop` and behind by none, so the two `origin/develop`-
based ACs (`AC-PCK-008`, and `AC-PCK-012`'s implicit base) resolve today. This is stated for the
record; it produces no defect at this HEAD, but a rebase onto a newer `origin/develop` would
change what `AC-PCK-008`'s diff-stat covers, which is one more reason that criterion is a poor
carrier for `REQ-PCK-008`.

---

## Hunt 6 — D4's `none` deviation is coherent

**Judgment: the deviation is coherent and adequately stated; this hunt does not produce a
blocking defect.** I verified D4's reconstruction against the diff rather than against the
SPEC's summary of it. `git show e91def4ca --format="" -- doc-generation.md` confirms all four of
D4's rows: Step 4.1.5 is wholly `+` (did not exist); the pre-P1 recommended option is the deleted
`-` line quoted verbatim; the remaining three options (Review and Edit · Generate harness · Done)
are untouched context; and both [HARD] clauses are `+` additions.

The hunt's concern was that `none` "reproduces neither the pre-P1 state nor P1's state exactly"
and that the SPEC states an approximation rather than an identity. Reading `spec.md:124-133`
against that charge: D4 does not approximate. It decomposes — three enumerated rows reproduced,
two named clauses deliberately retained, with the reason given ("those constrain the *question*
… not the *card behaviour* the key governs") and the deviation explicitly labelled a decision
rather than drift. `spec.md:175` carries it into Out of Scope. A reader can reconstruct exactly
what `none` is from the document. That is a plain statement, and I decline to manufacture a
defect from it.

One genuine under-specification, non-blocking (D7 below): P1 also added a **fifth** option,
"Create SPEC later: Leave the card queued and stop here" (`doc-generation.md:351`). Under `none`
no card is issued, so that option is incoherent — and `REQ-PCK-004`'s "pre-P1 next-steps option
set" correctly excludes it, since the pre-P1 set is four options. But `REQ-PCK-004` never says
so, and the retained four-option-cap clause carries a P1-specific instruction routing "Review and
Edit Documentation" to the `Other` path — an instruction that is inert under `none`, where all
four options fit. The requirement should say which of the two it means.

---

## Defects Found

**D1 — `pipeline` is behaviourally indistinguishable from `card`, and its new option text is
under-constrained at the gate** — `spec.md:83` (`REQ-PCK-006`), `spec.md:105-107` (D1
resolution), `acceptance.md:58-62` (`AC-PCK-006`) — Severity: **critical** — Class: **blocking**
— Both directions of the D1 hunt land. Nothing in the document states any behaviour under
`pipeline` that does not also hold under `card`: both issue the card identically by
`REQ-PCK-006`'s own words, both occupy the recommended slot with a continuation branch, both pass
the same gate. The entire delta is one sentence of option text, and `AC-PCK-006` verifies that
sentence alone. Separately, that new sentence is not obliged to carry the Implementation Kickoff
Approval clause that the `card` option it replaces carries (`doc-generation.md:350`, the file's
sole occurrence), so `pipeline` presents a continuation-through-run branch with the gate reminder
removed.
**Required fix**, both parts:
(a) Add to `spec.md` §3 D1 a subsection "What `pipeline` changes relative to `card`" stating the
delta in one sentence, and amend `REQ-PCK-006` so it names that delta positively rather than by
reference to `card`. If, on inspection, the honest answer is "only the recommended option's
wording", say exactly that in `REQ-PCK-006` and in D1 — a presentation-only third value is
shippable, an undeclared synonym is not.
(b) Extend `REQ-PCK-006` with: "…and the option text shall carry the Implementation Kickoff
Approval clause in the same terms as the `card` option." Extend `AC-PCK-006`'s Then clause with
the positive assertion `grep -c "Implementation Kickoff Approval"` over the `pipeline` row
returns ≥ 1, replacing the untestable negative conjunct.

**D2 — the wizard-localization gap is factually false, and the fallback it authorizes breaks an
existing test** — `spec.md:186` (§5 Gaps), `plan.md:17` (§B item 1), `acceptance.md:129` (§D.5) —
Severity: **critical** — Class: **blocking** — `GetLocalizedQuestion` already translates option
labels and descriptions (`translations.go:571-586`); twelve `QuestionTypeSelect` questions exist,
not one; `conversation_language` is the single member of `optionTranslationExemptIDs`
(`translations_completeness_test.go:13-15`) and is therefore the worst possible precedent to
generalize from; `audit_model` (`translations.go:137-146`) is the precedent the SPEC says is
absent. The authorized fallback — folding option descriptions into the question body — makes the
new question a non-exempt `QuestionTypeSelect` with zero option translations, which fails
`TestWizardQuestionTranslationCompleteness` (`:95-134`) with nine errors across ko/ja/zh.
**Required fix**: delete the §5 Gaps bullet at `spec.md:186` and replace it with a measured
statement that the precedent is `audit_model` at `translations.go:137`; delete `plan.md` §B item
1 entirely and replace M4's instruction with "populate `trans.Options` with three
`OptionTranslation{Label, Desc}` entries in each of ko/ja/zh, matching `audit_model`'s shape";
delete the second clause of `acceptance.md:129` (the folding branch); and extend `AC-PCK-010`'s
Then clause to assert non-empty, mutually distinct option `Label` and `Desc` for all three tokens
in each of ko/ja/zh, naming `TestWizardQuestionTranslationCompleteness` as the enforcing test.

**D3 — `AC-PCK-008`'s blocking conjunct has no reachable red** — `acceptance.md:70-74` — Severity:
**major** — Class: **blocking** — The `git diff --stat origin/develop...HEAD` over three
gate-owning files returns empty **today**, before any implementation exists (measured verbatim,
Hunt 2), and none of the three files appears in any of `plan.md` §F M1-M6's file lists. The
criterion cannot fail through any implementation the plan describes, which makes it the
"AC that cannot fail" class flagged twice in `.moai/reports/t331/plan-audit-iter2.md` (its D1
and D3). The second conjunct is reachable but guards only the `card` branch, so `REQ-PCK-008` —
the kickoff invariant, one of the two prohibitions the whole D1 narrowing rests on — has no
blocking criterion with a reachable red.
**Required fix**: replace `AC-PCK-008` with the criterion drafted in Hunt 2 above (per-branch
kickoff-clause presence, counted against the measured pre-change value of `1`). Retain the
diff-stat check if wanted, but re-file it as non-blocking hygiene in `acceptance.md` §D.1 and
strike it from `REQ-PCK-008`'s verification path in §D.2.

**D4 — `REQ-PCK-011`'s "descriptions" is resolved two ways** — `spec.md:88`, `acceptance.md:92`,
`plan.md:93` — Severity: **minor** — Class: **non-blocking** — `REQ-PCK-011` requires "labels and
descriptions present in all four locales". `AC-PCK-011` resolves this as five keys (`.title`,
`.desc`, three `.opt.<value>`) × four locales = 20 entries — i.e. field-level description only.
`plan.md:93`'s proposed row calls bare `closedSeam(...)` without the `withOptionDesc` wrapper
that `internal/settings/schema_sections.go:143` provides and that the neighbouring `audit_model`
rows (`:380-389`) use for exactly this purpose. If "descriptions" means per-option, both the AC
and the plan under-deliver.
**Required fix**: amend `REQ-PCK-011` to read "with option labels and a field-level description
present in all four locales" if `AC-PCK-011`'s reading is intended; otherwise wrap the M5 row in
`withOptionDesc` and raise `AC-PCK-011`'s count from 20 to 32 entries.

**D5 — inventory line citation is off by one** — `spec.md:139` — Severity: **minor** — Class:
**non-blocking** — The SPEC cites `workflow.worktree.auto_cleanup` at "inventory line 2922-2924";
`sed -n '2918,2926p'` places the row at 2921-2923. The row, its class `P`, and its skill-path
evidence field are all as described, so the substance is correct.
**Required fix**: change "2922-2924" to "2921-2923" at `spec.md:139`.

**D6 — `REQ-PCK-004` does not say what happens to the two P1-added options under `none`** —
`spec.md:81`, `spec.md:133` — Severity: **minor** — Class: **non-blocking** — Detailed in Hunt 6.
"Create SPEC later" is incoherent when no card is issued and is correctly absent from the pre-P1
set, but the requirement never says so; and the retained four-option-cap clause carries a
P1-specific `Other`-path routing instruction that is inert under `none`'s four-option set.
**Required fix**: extend `REQ-PCK-004` with "…and shall omit the `Create SPEC later` option,
which has no queued card to refer to; the four-option cap is satisfied without the `Other` path."

Blocking: **3** (D1, D2, D3). Non-blocking: **3** (D4, D5, D6).

---

## What is right, briefly

- §3 D3, D4 and D6 are measured rather than asserted, and every measurement I re-ran reproduced.
  D3's reasoning about *why* `SYNC-STRATEGY`'s stop does not transfer (push/PR irreversibility
  versus a display preference) is the correct discriminant, not a convenience.
- D2's rejection of `*string` is right and correctly reasoned: `todo.enabled`'s pointer buys a
  distinction a named-default enum does not have.
- D5 and `AC-PCK-009` are correct in mechanism and non-vacuous in the mutant, verified in source.
- Traceability is complete, and §D.2/§D.3 declare the negative and indirect verification cases
  rather than hiding them — the honesty that made D3 findable.
- §5 Gaps is the right instinct, disciplined and specific. Its failure (D2) is one wrong entry,
  not an absent section.

---

## Gaps — what this audit did not observe

- **No test was executed.** `TestShippedConfigKeysHaveReaders`,
  `TestWizardQuestionTranslationCompleteness`, and the `internal/web` i18n governance tests were
  read in source and their failure paths traced by reading the assertion bodies; none was run.
  The claim that the folding fallback produces nine errors is derived from the loop bounds
  (3 locales × 3 options) and the two `t.Errorf` calls at `:126-130`, not from an observed run.
- **`make build` was not run**, so no claim is made about the embedded template versus
  `internal/template/templates/`. `AC-PCK-012` was probed only by re-running its `cmp` matrix on
  the current tree.
- **`internal/settings/schema_sections.go` was read only around `seamSectionFields`**
  (`:317-390`) and the helper list. I did not verify that a `workflow → project → continuation`
  path resolves through the settings write layer, nor that `yamlpatch.KeyEdit` handles a
  three-segment path — `plan.md:87` asserts the latter and I did not check it.
- **`internal/core/project/initializer{,_expansion}.go` and `internal/cli/init.go` were not
  opened.** M4's persistence instruction is unverified.
- **The four `translations.go` locale maps were confirmed for ko/ja/zh** via the `audit_model`
  entries at `:137`, `:296`, `:455`. I did not enumerate the maps to confirm no fourth map
  exists; `translations_completeness_test.go:7` (`localizableLocales = []string{"ko","ja","zh"}`)
  is the basis for treating English as the source language.
- **PR #1601 / #1600 CHANGELOG collision** (`spec.md:188`) was not reproduced from git history;
  it remains as the SPEC carries it, taken from the card brief.
- **`git status --short` was not captured at report time**; the working tree contained this
  report file when written.

Nothing else was left unobserved.

---

## Recommendation

FAIL at 0.78 against a Tier M threshold of 0.80, with all seven must-pass criteria clear. The
score is two hundredths short and three blocking defects are enumerated above; the Tier M
ceiling permits one further iteration, so this routes as a scoped delta fix and a second audit
rather than an escalation.

Fix in this order — the first two change what gets built, the third changes what proves it:

1. **D2 first**, because it is purely factual and unblocks M4 immediately: correct `spec.md` §5,
   delete `plan.md` §B item 1, strike the folding branch from `acceptance.md` §D.5, and
   strengthen `AC-PCK-010`. Nothing here requires a design decision.
2. **D1 second**, because it is the one genuine design question left open: state what `pipeline`
   changes relative to `card`, in `REQ-PCK-006` and in D1. If the honest answer is "the
   recommended option's wording and nothing else", write that — the SPEC's own §3 D1 already
   demonstrates it can state an unflattering resolution plainly, and it should do so here too.
   Then add the kickoff-clause obligation to `REQ-PCK-006`/`AC-PCK-006`.
3. **D3 third**: replace `AC-PCK-008` with the per-branch kickoff-clause criterion. This depends
   on D1(b), since it counts the clause across whichever options D1 settles on.
4. D4, D5, D6 are one-line edits and can ride along.

D1 and D3 are the same defect seen from the requirement side and the verification side. Fixing
one without the other leaves `REQ-PCK-008` — the invariant the entire narrowing exists to
protect — with no blocking criterion that can fail.
