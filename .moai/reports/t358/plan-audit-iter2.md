# SPEC Review Report: SPEC-CI-TEST-OBSERVABILITY-001 (iteration 2)

Card: **t358** · Branch: `WT-ci-test-observability` · Base: `origin/develop c6aa61346`
Iteration: 2/3 (Tier M ⇒ `plan_audit_tier_ceilings` M=2; this is the ceiling iteration)
Auditor tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t358`

**Reasoning context ignored per M1 Context Isolation.** The dispatch supplied the iter-1 defect
list and the author's disagreement; both were treated as claims to re-verify, not findings to
adopt. Every figure below was re-measured in this tree, in this run.

## Audited version (pinned, re-hashed before and after)

| File | sha256 prefix | bytes | matches dispatch table |
|---|---|---|---|
| `spec.md` | `05a5ecae` | 16239 | ✓ |
| `plan.md` | `dc4a17da` | 11267 | ✓ |
| `acceptance.md` | `6390210f` | 11606 | ✓ |
| `progress.md` | `cfde51d8` | 2904 | ✓ |

All four match the dispatch table exactly. Re-hashed at the end of the audit: **unchanged**. No
mid-audit revision occurred, so this report is pinned to one version throughout — unlike iter-1.

---

## Verdict: **PASS-WITH-DEBT**

**Overall Score: 0.895** — delta from iter-1 **+0.085** (0.81 → 0.895). Monotonic.
**Tier M threshold 0.80** (`tier: M` now present at `spec.md:13`, verified mechanically) — clears
by 0.095. Under the Tier L threshold 0.85 it would also clear; the tier question is no longer
verdict-determining.

No must-pass criterion failed. Nine iter-1 defects were repaired or substantially repaired; none
regressed. Nine new findings, all **minor** — five blocking-class (internal-consistency nicks in
newly-added text), four optional. None is critical, none blocks kickoff on its own, and all five
blocking ones are text-only edits.

The "PASS-WITH-DEBT" label is chosen deliberately over plain PASS for two distinct reasons that
must not be conflated:

1. **The SPEC's own two named debts** (AC-CTO-003's CI-level red path, AC-CTO-005b) are
   structural, correctly recorded as debt, and **never asserted as passes**. That is a virtue of
   this revision, not a defect against it.
2. **Five blocking-class audit findings remain open** (N1-N5 below). They are the debt this
   *report* carries.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CTO-001` … `REQ-CTO-011` at `spec.md:89-99`.
  Set is complete: 11 REQs, **no gaps, no duplicates, uniform 3-digit padding** — the three tests
  MP-1 names. PASS. *However* the listing **order** is `001…009, 011, 010` (`spec.md:98` carries
  011, `:99` carries 010) — REQ-CTO-011 was inserted before REQ-CTO-010 by the D6 repair.
  Ordering is not one of MP-1's three tests, so this is recorded as finding **N2**, not a
  must-pass failure.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer**
  (`REQ-XXX` in `spec.md`) only; the Given-When-Then entries in `acceptance.md` are the correct
  verification-layer format and were NOT penalized here. All eleven REQ **texts** match a GEARS
  pattern. New since iter-1: `REQ-CTO-011` (`spec.md:98`) — "The test invocation shall not
  redirect stderr into the event-stream file" — canonical negative form, valid. Zero legacy
  `IF/THEN` forms. See § Ruling on the author's D11 disagreement for the label-vocabulary
  question, which does not bear on MP-2 (MP-2 binds text, not annotations).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  at `spec.md:2-14`, plus the optional `tier: M` at `:13`. No rejected snake_case alias.
  Corroborated mechanically: `moai spec lint .moai/specs/SPEC-CI-TEST-OBSERVABILITY-001/spec.md`
  → `✓ No findings — all SPEC documents are valid`.
- **[N/A] MP-4 Section 22 language neutrality** — `module: ".github/workflows"`; the SPEC touches
  no path under `internal/template/templates/`. Single-language scoped; auto-passes.
- **[N/A] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'
  spec.md | sort -u` returns exactly one token, the SPEC's own ID. The D7 verb has no subject.
  N/A per the MP-4 precedent.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. D8-4 auto-PASS.
- **[N/A] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION'` over the SPEC directory →
  rc=1, zero matches. `research.md` does not exist (Tier M artifact set: spec/plan/acceptance/
  progress only, confirmed by `ls`). N/A per the MP-4 precedent.

---

## Category Scores (rubric-anchored)

| Dimension | Score | iter-1 | Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.88 | 0.85 | 0.75–1.0 | `plan.md:13-31` is now an exemplary WRONG/WRONG/CORRECT contrast that names the failure mechanism and explicitly forbids "simplifying" back to the broken form. The D3 one-run ambiguity is fully resolved. Deductions: four small internal-consistency nicks in newly-added text (N1-N4). |
| Completeness | 0.90 | 0.75 | 0.75–1.0 | `tier: M` added (D8); the fourth call site added to §F with the false "exactly three" claim replaced (D7); stderr discipline given REQ-CTO-011 + a concrete AC clause (D6); branch-push deviation named with verified CI-inertness (D5); §I closure-ordering table added (D1); tier justification added (D9). Residual: no owner named for remote-branch disposal (N5); no REQ normatively imposes universal application (D10 residual); §F row 3's check-status cell is blank (N7). |
| Testability | 0.90 | 0.75 | 0.75–1.0 | The three unreachable-`Given` criteria are fixed: AC-CTO-003 relocated to local shell-semantics verification (I executed its discriminator — see below), AC-CTO-005 split into 005a/005b with 005b explicitly not claimed, AC-CTO-006 re-anchored to a mechanical name-set diff I confirmed executable. AC-CTO-002's mutant probe and AC-CTO-007's six named failing forms intact. Deduction: AC-CTO-003 clause 4 introduces a precondition absent from — and incompatible with — its own `Given` (N3). |
| Traceability | 0.90 | 0.90 | 0.75–1.0 | All eleven REQs are covered by at least one AC (`acceptance.md:9-18`); every AC resolves to a real REQ. AC-CTO-008 retargeted from "plan §M3 replication" to REQ-CTO-001/005/008 (D10 addressed at the link level). Residual unchanged from iter-1: no requirement normatively states the change must reach every in-scope call site — "(all four sites)" lives in the matrix cell, not in any REQ text. |

Aggregate = (0.88 + 0.90 + 0.90 + 0.90) / 4 = **0.895**.

---

## Independent verification performed this run

**The D4 repair was tested in a shell, not reasoned about.** Three step bodies were executed under
the real CI shell flags:

```
$ /bin/bash --noprofile --norc -e -o pipefail  <body with `; rc=$?`>
step exit=1                                    ← census line NEVER printed
$ /bin/bash --noprofile --norc -e -o pipefail  <body with `|| rc=$?`>
CENSUS PRINTED (right form), rc=1
step exit=1
$ /bin/bash --noprofile --norc -e -o pipefail  <body with set +e / set -e bracket>
CENSUS PRINTED (bracket form), rc=1
step exit=1
```

Both forms `plan.md:24-30` prescribes survive `-e`, capture the non-zero rc, still run the census,
and re-raise. The form it forbids dies silently before the census — exactly as `plan.md:16` claims.
**AC-CTO-003 clause 1's discriminator ("the `; rc=$?` form fails here; `|| rc=$?` passes") is
mechanically real.** *Residual risk:* executed on GNU bash 3.2.57 (macOS); CI runs bash 5.x. The
`-e`/`||` interaction is unchanged between those versions, but I did not measure on 5.x.

**D5's CI-inertness claim verified, not taken on trust.** `grep -n workflow_dispatch
.github/workflows/ci.yml` → rc=1 (no match). Trigger block at `ci.yml:16-19`: `push: branches:
[main, develop]`, `pull_request: branches: [main]`. Pushing `WT-ci-test-observability` matches no
trigger — the push is genuinely CI-inert, starts no run, consumes no runner. **The SPEC's claim at
`spec.md:120` is true.**

**D7's four-row scope verified exhaustively.** `grep -rn "go test" .github/workflows/` returns
exactly four executable non-`-run`-scoped invocations: `ci.yml:183`, `ci.yml:238`, `ci.yml:329`,
`release-pr-multi-os.yml:189` — matching §F's four rows one-for-one. `ci.yml:329` confirmed as job
`test-integration`, `name: Integration Tests (${{ matrix.os }})`, `os: [ubuntu-latest,
macos-latest, windows-latest]`, `shell: bash`, no `-v`. The four excluded `-run`-scoped steps
(`lsel-leak-guard.yaml:37` + `template-neutrality-check.yaml:60,70,87`) all confirmed carrying
`-v`.

**AC-CTO-008's "eight legs" arithmetic verified.** `test` = ubuntu-only (1), `test-race`
`runs-on: ubuntu-latest` (1), `test-integration` 3-OS (3), `full-matrix-test` 3-OS (3) = **8**. ✓

**AC-CTO-006's complement invariant verified.** `ci.yml:108` `if: needs.detect.outputs.go_code ==
'true'` vs `ci.yml:266` `if: needs.detect.outputs.go_code != 'true'` — strict complements. The
newly-added clause is factually checkable.

**AC-CTO-007's two premises re-confirmed intact** (dispatch says do not re-litigate; confirmed, not
weakened): `internal/statusline/usage_test.go:186` carries an unconditional
`t.Skip("TODO: mock security command for unit testing")`; `profile_bench_test.go:305-307` gates on
`MOAI_PROFILE_PHASES != "1"`. All six named failing forms are present at `acceptance.md:139-144`
and none is softened. **AC-CTO-007 is intact and remains the strongest part of the SPEC.**

**Branch protection measured** (new this run): `gh api repos/modu-ai/moai-adk/branches/main/
protection --jq '.required_status_checks.contexts'` →
`["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS
Gate"]`. `develop` returns `Branch not protected` (404).

---

## Repair status of every iter-1 defect

| # | Sev (iter-1) | Status | Evidence |
|---|---|---|---|
| D1 | major | **Partially repaired** | §I closure-ordering table added (`spec.md:154-165`) splitting pre-close vs debt; DoD gained the conditional fallback bullet. But DoD bullet 1 (`acceptance.md:130`) still lists AC-CTO-007 among ACs that "all PASS pre-close" unconditionally, contradicting §I and the AC matrix. No discharge owner named for the fallback case. → **N4** |
| D2 | major | **Repaired** | AC-CTO-006 rewritten as a mechanical job-`name:`-set diff, explicitly noting "this repo opens **no card PRs**", plus the `test` ↔ `test-skip-marker` complement invariant. Both halves verified executable. |
| D3 | **critical** | **Repaired** | AC-CTO-003 relocated to local execution ("verified LOCALLY, by design"), with the reason stated correctly — it tests shell semantics, not CI infrastructure. AC-CTO-005 split: 005a (green dispatch, pre-close) / 005b (**DEBT**, "this SPEC does not claim it"). The debt is recorded as debt in three places (`spec.md` §I, `acceptance.md` matrix + AC body, `progress.md:17-19`) and asserted as a pass in none. **No second dispatch was quietly authorized** — the prohibition is restated at `spec.md:122`, `plan.md:97-99`, `acceptance.md:80`, `:106`, and DoD. |
| D4 | **critical** (dispatch) / major (iter-1) | **Repaired, and mechanically confirmed** | `plan.md:13-31` gives the shell string verbatim with provenance (run `33173944485`, job `98857764037`, step header line 3), names TWO wrong forms (`; rc=$?` and the `tee` dodge) with mechanisms, prescribes `|| rc=$?` plus the `set +e` bracket, and adds an anti-simplification warning. Made mechanically checkable in AC-CTO-003 clause 1 and AC-CTO-008. **I executed all three forms — the repair is correct.** |
| D5 | major | **Partially repaired** | `spec.md:120` names the push as "a deliberate, minimal exception" to the lane protocol, cites the lead's dispatch approval as its basis, and states CI-inertness with the trigger evidence — **which I verified is true**. Missing: iter-1's required fix also asked who deletes the remote branch after M4. Nobody is named. → **N5** |
| D6 | minor/blocking | **Repaired** | `REQ-CTO-011` added (`spec.md:98`, Unwanted/negative form) + AC-CTO-003 clause 4, which states outright "A `2>&1 > f` implementation fails this clause". The implementation that would have satisfied every iter-1 criterion now fails a named one. |
| D7 | major | **Repaired in spec.md / acceptance.md; one stale line left in plan.md** | §F now has four rows; "exactly these three sites, not more" replaced by "The scope is **exactly these four sites**" (`spec.md:118`) with the lead ruling recorded. AC-CTO-008 enumerates four. But `plan.md:9` still reads "2 workflow files (**3 invocation sites**)" — six lines below `plan.md:3`'s "**four** `go test` call sites". → **N1** |
| D8 | major | **Repaired** | `tier: M` at `spec.md:13`, verified by grep and by `moai spec lint`. Threshold resolves to 0.80. |
| D9 | minor/optional | **Repaired** | `plan.md:3` now maps scope to tier explicitly and rules out both S ("this has four call sites, a new executable artifact, and an 8-criterion acceptance set") and L ("no data model, no new Go production path, no design space left open"). The justification matches the *current* four-site scope. |
| D10 | minor/optional | **Partially repaired** | AC-CTO-008's matrix cell retargeted from "plan §M3 replication" to `REQ-CTO-001, REQ-CTO-005, REQ-CTO-008 (all four sites)` — the orphan link is gone. But no REQ text normatively imposes universal application; "(all four sites)" is a matrix parenthetical. Optional, unchanged in substance. |
| D11 | minor/optional | **WITHDRAWN — the finding was wrong** | See the ruling below. The author complied anyway (`spec.md:92` now reads "(Event-driven)"), which is harmless: both labels are defensible for that sentence. |

**Papered over: none.** **Regressed: none.** Every repair changed the substance, not just the
wording, and the two debts D3 created are carried openly rather than converted into assertions.

---

## Ruling on the author's D11 disagreement

**The author is right. D11 was a false finding, and the defect is in the audit rule, not the SPEC.**

`"Event-detected"` is a named GEARS pattern in both documents the author cited, measured in this
tree:

- `.claude/skills/moai-workflow-spec/SKILL.md:59` — a row of the table headed "GEARS Five Patterns
  (current notation)" at `:51`: `| Event-detected (replaces IF/THEN) | "**When**
  <undesired-condition-detected>, the <subject> shall <response>" | ...`
- `.claude/skills/moai-foundation-core/SKILL.md:113` — "Five patterns — Ubiquitous …;
  Event-driven …; State-driven …; Where (capability gate) …; **Event-detected** (replaces the
  deprecated conditional modality) …"

Corroborated in three further project files: `moai-foundation-core/modules/commands-reference.md:98`,
`modules/spec-first-ddd.md:34`, and `.claude/skills/moai/workflows/plan/spec-assembly.md:74` (which
enumerates "the 5 GEARS patterns … Event-detected unwanted").

The canonical five-name set in this project is therefore: **Ubiquitous, Event-driven, State-driven,
Capability gate, Event-detected.** The name `"Unwanted"` — which my own rubric uses for the fifth
pattern and which iter-1 enforced as one of only five admissible names — **does not appear in the
canonical table at all.** The divergence is real and it runs in the opposite direction from D11's
claim: iter-1 pushed the author *away* from the canonical name toward a non-canonical one.

Consequences, stated plainly:

1. **D11 is withdrawn.** `spec.md`'s original "(Event-detected)" label for REQ-CTO-004 was correct
   and arguably more precise than the "(Event-driven)" it was changed to, since the sentence is
   "When a test failure **is detected** …" — the undesired-condition-detected form.
2. **This is a finding against the plan-auditor rubric**, which admits five pattern names not
   matching the project's own skill documents. A rule that contradicts the project's canonical
   authoring guide will misfire on any SPEC that labels correctly. Recommend the rubric's fifth
   pattern be reconciled with `moai-workflow-spec/SKILL.md:51-59`.
3. **No score effect either way.** MP-2 and rubric RQ-6 bind requirement *text*, not annotations.
   The SPEC's `(Unwanted)` labels on REQ-CTO-003/-009/-011 and `(Event-driven)` on REQ-CTO-004 all
   resolve to a documented pattern under one of the two authorities, and every REQ sentence is
   well-formed. Nothing was scored on labels.

## Ruling on the `[SUPERSEDED]` marking in progress.md

**Correct practice here — though the cited authority is the wrong one.**

`progress.md:12` preserves the falsified "scope stated at three call sites" clause with an inline
`[**SUPERSEDED** — a fourth, ci.yml:329, was found in plan audit round 1 and brought in scope by
lead ruling]`, paired with an affirmative corrective at `:14` ("**Scope corrected to FOUR call
sites.**"). Keep it.

The reason is not the Lessons Protocol. That protocol (`moai-constitution.md` § Lessons Protocol)
governs the auto-memory lesson store — `feedback_*.md` topic files under the project memory
directory — not a SPEC's `progress.md`. Citing it here is a mis-citation.

The correct basis is the evidence-trail principle in `verification-claim-integrity.md` §2:
`progress.md` is this SPEC's evidence record, and silently rewriting a claim that was **measured
false** would erase the record that the SPEC once asserted completeness it did not have and was
corrected. An auditor reading only a rewritten line could not tell a never-wrong SPEC from a
corrected one. Preserving the falsified claim with its correction visible is what makes the record
auditable. The pairing with an affirmative corrective line is what keeps it from being ambiguous.

---

## New defects introduced by the revision

Nine findings, all **minor**. The SPEC grew ~40%; these live in the newly-added text, as expected.

**N1 — `plan.md` contradicts itself on scope, six lines apart**
`plan.md:9` — Severity: **minor** — Class: **blocking**
`plan.md:3` says "**four** `go test` call sites across two workflow files". `plan.md:9` still says
"Surface: 2 workflow files (**3 invocation sites**) + 1 summary implementation." The D7 repair
updated §F, AC-CTO-008, `progress.md`, and `plan.md:3`, but missed the surface line. Measured
truth: four (verified by exhaustive grep). Note `plan.md:74` ("the remaining **three** call sites")
is **correct** — M2 takes `ci.yml:183`, so M3 takes the remaining three; that line is not stale.
*Required fix:* `plan.md:9` → "2 workflow files (4 invocation sites)".

**N2 — `REQ-CTO-011` is listed before `REQ-CTO-010`**
`spec.md:98-99` — Severity: **minor** — Class: **blocking**
The D6 repair appended REQ-CTO-011 to the end of the §E list, which already ended at REQ-CTO-010,
producing the order `001…009, 011, 010`. The *set* is complete and correctly padded (MP-1 PASS),
but a reader or reviewer scanning §E sequentially hits an inversion.
*Required fix:* swap the two lines so §E reads 001…011 in order.

**N3 — AC-CTO-003 clause 4 requires a tree state its own `Given` excludes**
`acceptance.md:48-58` — Severity: **minor** — Class: **blocking**
The `Given` is "a tree containing one deliberately failing test". Clause 4 then requires "with a
deliberately broken build (e.g. a syntax error), the build error appears on the console and **not**
inside the stream file". A broken build and a failing test are **mutually exclusive tree states** —
under a syntax error nothing compiles, so no test fails and clauses 1-3 cannot be evaluated in the
same execution. The AC presents as one execution and actually requires two.
*Required fix:* split into AC-CTO-003a (failing test → clauses 1-3) and AC-CTO-003b (broken build →
clause 4), each with its own `Given`. The substance of both is right; only the packaging is wrong.

**N4 — DoD bullet 1 asserts unconditionally what §I and the AC matrix make conditional**
`acceptance.md:130` vs `spec.md:163-165` + `acceptance.md:16` — Severity: **minor** — Class:
**blocking**
DoD bullet 1: "AC-CTO-001, -002, -003, -004, -005a, -006, -007, -008 all PASS pre-close." But
`spec.md:163` and the AC matrix both say AC-CTO-007 closes **post-merge as debt** if the fallback
path is taken, and DoD bullet 3 repeats that. Bullet 1 and bullet 3 disagree about AC-CTO-007.
This is the residue of D1: the honesty requirement is now stated in three places, but the DoD's
own top line was not made conditional.
*Required fix:* qualify bullet 1 — "…all PASS pre-close (AC-CTO-007 pre-close on the dispatch path;
post-merge as recorded debt on the fallback path)" — and name who reads the develop-push run and
discharges it.

**N5 — the remote card branch has no disposal owner**
`spec.md:120` + `plan.md:87` — Severity: **minor** — Class: **blocking**
The push exception is now well argued and its CI-inertness is true (I verified it). But iter-1's
required fix also asked who deletes the remote branch after M4, and nobody is named. `spec.md:120`
says "The exception buys a remote ref and nothing else, and no PR is opened" — which bounds the
*cost* without assigning the *cleanup*. Nothing in the lane protocol disposes of a pushed card
branch, because the protocol does not contemplate one existing.
*Required fix:* one sentence naming who deletes `origin/WT-ci-test-observability` and when
(natural point: immediately after AC-CTO-007's evidence is recorded, since the ref's only purpose
is resolving the dispatch).

**N6 — AC-CTO-006's illustrative name list omits the job M3 actually edits**
`acceptance.md:117` — Severity: **minor** — Class: **optional**
The "in particular" list names `Test (${{ matrix.os }})`, `Race Test`, `Integration Tests (${{
matrix.os }})`, and `Release PR Multi-OS Gate`. Measured: the job carrying `release-pr-multi-os.yml:189`
is key `full-matrix-test` with `name: Release Verify (${{ matrix.os }})` (`:83-84`).
`Release PR Multi-OS Gate` is a **different** job, key `release-pr-gate` (`:229-230`), which the
SPEC never modifies. So the list names an untouched gate and omits the one name M3 puts at risk.
The AC's *mechanism* (whole-set `name:` diff) catches this regardless, which is why this is
optional rather than blocking — but the list could mislead an implementer into checking four names
and missing the fifth.
*Required fix (optional):* add `Release Verify (${{ matrix.os }})` to the enumeration.

**N7 — §F row 3's "Check status" cell states no status**
`spec.md:107` — Severity: **minor** — Class: **optional**
Rows 1, 2, and 4 each state required-or-advisory ("required — `Test (ubuntu-latest)`", "advisory
(`Race Test`, deliberately not required)", "required via `Release PR Multi-OS Gate`"). Row 3 gives
only the name and matrix shape. Measured this run: `Integration Tests (<os>)` is **not** in
`main`'s required contexts (the required set is `Test (ubuntu-latest)`, `Lint`, `Build
(linux/amd64)`, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`), so the honest cell value is
*advisory*. Cheap to fill and it makes REQ-CTO-009's blast radius readable at a glance.
*Required fix (optional):* fill the cell — "advisory (not in `main`'s required contexts)".

**N8 — `plan.md` M2 attributes `if: always()` to a house convention that does not contain it**
`plan.md:70` — Severity: **minor** — Class: **optional**
"…upload with `actions/upload-artifact@v7`, `retention-days: 7`, `if: always()` — the house
convention already at `ci.yml:433-438`." Measured `ci.yml:433-438`: `- name: Upload artifact` /
`uses: actions/upload-artifact@v7` / `with:` / `name:` / `path:` / `retention-days: 7`. **No
`if:` line.** The action and retention *are* the house convention; `if: always()` is an addition
beyond it. `spec.md:115` states this correctly ("`actions/upload-artifact@v7` with
`retention-days: 7`, matching `ci.yml:433-438`") — only `plan.md` over-bundles. (`if: always()` is
an established pattern elsewhere in the repo, e.g. the comment at `release-pr-multi-os.yml:191`,
so the practice is fine; only the citation is.)
*Required fix (optional):* move `if: always()` outside the "house convention at ci.yml:433-438"
attribution.

**N9 — the shell string is given with a Linux-only path for four sites, two of which run Windows**
`spec.md:112` — Severity: **minor** — Class: **optional**
"`shell: bash` resolves to the verbatim shell string `/usr/bin/bash --noprofile --norc -e -o
pipefail {0}`" is asserted for all four sites, but `ci.yml:329` and `release-pr-multi-os.yml:189`
each run a `windows-latest` leg, where the interpreter path differs (Git-bash, not `/usr/bin/bash`).
The **flags** are identical, so every semantic consequence the SPEC draws from the string — the
`-e` hazard, the `|| rc=$?` remedy, AC-CTO-003's local reproduction — holds unchanged on all legs.
`acceptance.md:47` already states it flags-only ("`bash --noprofile --norc -e -o pipefail`"), which
is the accurate form.
*Required fix (optional):* qualify `spec.md:112` — "on the Linux legs; the interpreter path differs
on `windows-latest`, the flags do not".

---

## Gaps (explicitly not observed)

- **No CI run was executed or read.** Every claim about what a dispatch would produce is derived
  from workflow files and test sources in this tree, not from a run. AC-CTO-007 remains unclosed by
  construction — that is the SPEC's design, not an audit gap.
- **Shell semantics measured on bash 3.2.57 (macOS), not the runner's bash 5.x**, and not on a
  Windows Git-bash leg. The `-e` / `||` interaction is stable across those versions, but I did not
  measure it there.
- **`develop` branch protection returns 404 (not protected).** Only `main`'s required contexts were
  read. If the fallback path is taken, no required-check parity risk arises on the develop push,
  but I did not pursue that further — it is out of this SPEC's scope.
- **The census implementation does not exist yet** (M1 deliverable), so AC-CTO-002's fixture check
  could not be exercised — only its criterion text reviewed. Expected at this phase.
- **Codecov behaviour under `-json` was not observed**, only reasoned from "Codecov reads the file".
  `plan.md:72` correctly instructs "Confirm this by observation, not by reading the YAML".

## Residual risk

- N3 is the finding most likely to cause real trouble in run-phase: an implementer who reads
  AC-CTO-003 as one execution will produce evidence for clauses 1-3 and quietly hand-wave clause 4,
  which is the exact stderr property REQ-CTO-011 exists to protect.
- N1's stale "3 invocation sites" sits in the plan's *surface* line — the sentence an implementer
  scans to size the work. It is the likeliest source of a missed fourth site during M3.
- The SPEC's two named debts are correctly recorded, but nothing mechanically enforces that they
  are re-read on the first red CI run. "Discharges on an ordinary event" is honest, and it is also
  how debt goes unpaid.

---

## Recommendation

**PASS-WITH-DEBT — proceed to Implementation Kickoff Approval after a text-only correction pass.**

The five blocking-class findings (N1-N5) are all single-sentence edits to newly-added text. None
requires re-deciding anything, none touches AC-CTO-007, and none warrants a third iteration or a
re-audit: the Tier M ceiling is 2 iterations and this is iteration 2. Fix them in one pass and
record the pass in `progress.md`.

Order (cheapest first, all mechanical):

1. **N1** — `plan.md:9`: "3 invocation sites" → "4 invocation sites".
2. **N2** — `spec.md:98-99`: swap so §E reads REQ-CTO-001…011 in order.
3. **N4** — `acceptance.md:130`: make DoD bullet 1 conditional on the observation path, consistent
   with §I and bullet 3.
4. **N5** — `spec.md` §G: one sentence naming who deletes the remote card branch and when.
5. **N3** — `acceptance.md`: split AC-CTO-003 into 003a (failing test, clauses 1-3) and 003b
   (broken build, clause 4). Slightly larger than the others but still packaging, not substance.
6. Optional, at the author's discretion: N6, N7, N8, N9.

**Do not touch on this pass** — verified strong, and in two cases verified *by execution* rather
than by reading:

- **AC-CTO-007** (`acceptance.md:127-144`) — six named failing forms, all intact, none softened;
  both premises re-confirmed in source. It survived a second adversarial pass. Leave it alone.
- **`plan.md` §B.1** (`:13-31`) — the `-e` analysis. I ran all three forms; the prescription is
  correct and the anti-simplification warning is load-bearing. This is now the best-evidenced
  paragraph in the SPEC.
- **`spec.md` §A.1's measured-evidence table with its explicit extrapolation disclaimer** (`:35-49`)
  and **§C.1's single-predicate rationale** (`:59-71`) — unchanged from iter-1 and still the model
  for how to separate observation from inference.
- **`spec.md` §I** and the AC matrix's DEBT rows — the debt is recorded as debt in three surfaces
  and asserted as a pass in none. That is the behaviour the iter-1 D3 repair was meant to produce.

**One finding against the harness, not the SPEC:** the plan-auditor rubric's five GEARS pattern
names diverge from `moai-workflow-spec/SKILL.md:51-59` and `moai-foundation-core/SKILL.md:113`.
The rubric admits "Unwanted" (absent from the canonical table) and rejects "Event-detected"
(present in it). Iter-1's D11 was a direct consequence. This should be reconciled before the rubric
misfires on another SPEC.
