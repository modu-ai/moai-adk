# SPEC Review Report: SPEC-CI-TEST-OBSERVABILITY-001

Card: **t358** · Branch: `WT-ci-test-observability` · Base: `origin/develop c6aa61346`
Iteration: 1/3 (Tier-resolved ceiling = 3; see D8 — `tier:` absent ⇒ Tier L)
Auditor tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t358`

**Reasoning context ignored per M1 Context Isolation.** The delegating prompt supplied measured
figures; those were treated as claims to re-verify, not as findings to adopt. Every figure cited
below was re-measured in this tree, in this run.

## Audited version (pinned)

The author revised artifacts twice during this audit (plan.md between reads, then acceptance.md +
progress.md). The audit is pinned to these hashes, measured `2026-08-28T13:32:06Z`:

```
8a3ebe9a9433f3fac9c9462a4274606b363810d2b1b74b73c0a18fcb4a753557  spec.md
e6aa8477d1c05364667ef5d2fb712075f28ce0262ef5112f417a0d50eb8d0140  plan.md
10953e3bfdfdae29ddb11b78398b488c26130cbb7afb7c5f825500c697706ae5  acceptance.md
15dbdcc7690da528a19d487c5d738c983babc8053d3b53a7a06ea958a4cbb3dd  progress.md
```

Findings against any later revision are not asserted. MCP: `mcp__moai__spec_*` was not exercised;
verification ran through the Bash CLI (`moai spec lint`) and native Grep/Read. Not a blocker.

---

## Verdict: **FAIL**

**Overall Score: 0.81** (Tier L threshold 0.85 — see D8; at the claimed Tier M threshold 0.80 it
would clear by 0.01, which is itself a reason D8 must be resolved before any re-audit).

No must-pass criterion failed. The FAIL is driven by the aggregate score plus five **blocking**
findings, three of which are acceptance criteria whose `Given` cannot be realized under this
SPEC's own stated constraints. On a SPEC whose entire purpose is to stop a check that observes
nothing, shipping three criteria that cannot be discharged is the failure mode reproducing itself
one layer up.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-CTO-001` … `REQ-CTO-010` at `spec.md:88-97`,
  sequential, no gaps, no duplicates, uniform 3-digit padding. Verified by reading all ten lines.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer**
  (`REQ-XXX` in `spec.md`), never against the `AC-XXX` verification layer. All ten sentences match
  a GEARS pattern: Ubiquitous (001 `spec.md:88`, 005 `:92`, 006 `:93`), Event-driven (002 `:89`,
  004 `:91`, 008 `:95`, 010 `:97`), Unwanted (003 `:90` "shall not write", 009 `:96` "shall not
  alter"), Where/capability-gate (007 `:94` "Where a job collects coverage, the test invocation
  shall retain…"). Zero legacy `IF/THEN` forms. Label defect noted separately (D11).
  The Given-When-Then entries in `acceptance.md` are the correct verification-layer format and
  were NOT penalized here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types at
  `spec.md:2-13`; no rejected snake_case alias (`created:` / `updated:` / `tags:` / `id:` all
  canonical). `phase: "v3.1.4 target"` is a release target, not a prohibited lifecycle stage.
  Corroborated mechanically: `moai spec lint .moai/specs/SPEC-CI-TEST-OBSERVABILITY-001/spec.md`
  → `✓ No findings — all SPEC documents are valid`.
- **[N/A] MP-4 Section 22 language neutrality** — single-language scoped. `module:
  ".github/workflows"`; the SPEC touches no path under `internal/template/templates/`, so the
  16-programming-language neutrality obligation does not attach. Auto-passes.
- **[N/A] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md
  | sort -u` returns exactly one token, the SPEC's own ID. No external SPEC is referenced, so the
  D7 verification verb has no subject. N/A per the MP-4 precedent.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. D8-4 auto-PASS
  (the literal substring does not appear; no build-tag obligation arises).
- **[N/A] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md spec.md
  acceptance.md progress.md` → rc=1, zero matches. `research.md` does not exist (Tier M artifact
  set). No open clarification marker.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | Requirements are unambiguous and measurement-backed; `spec.md:35-47` is an exemplary measured-evidence table that separates observation from extrapolation (`:49`). Deductions: the one-run-vs-red-run ambiguity (D3) is a genuine reading ambiguity, and `spec.md:91`'s pattern label is wrong (D11). |
| Completeness | 0.75 | 0.75 | One machine-readable field missing (`tier:`, D8); a fourth in-file `go test` call site unaccounted for in either scope or exclusions (D7, `ci.yml:329`); the stderr-only discipline is stated in `plan.md:14` / `spec.md:110` but carried by no REQ and verified by no AC (D6); the git-flow branch-push deviation is unnamed (D5). |
| Testability | 0.75 | 0.50–0.75 | The AC set is above average — `acceptance.md:34` carries an explicit mutant probe, and `acceptance.md:85-92` names six failing forms so AC-CTO-007 cannot be argued into a pass. But three criteria have no obtainable `Given`: AC-CTO-003 and AC-CTO-005 (D3) and AC-CTO-006 (D2). A criterion with no green path is untestable regardless of how sharply it is written. |
| Traceability | 0.90 | 0.75–1.0 | All ten REQs are covered by at least one AC (`acceptance.md:9-16`), and every mapping resolves. One orphan: AC-CTO-008 traces to `plan §M3 replication`, not to any REQ — no requirement states the change must reach all three call sites (D10). |

Aggregate = (0.85 + 0.75 + 0.75 + 0.90) / 4 = **0.8125 → 0.81**.

---

## The central question: can AC-CTO-007 be argued into a pass?

This was the audit's primary charge. Answer: **AC-CTO-007 itself is sound — it is the strongest
part of the SPEC.** The five named failing forms at `acceptance.md:87-91` close each escape route
the brief asked about: YAML-content assertion, local run substituted for CI, a run with no skip, a
count without a name, and an unreverted planted mutant. The sixth form (`:92`, re-running until a
pass appears) was added during this audit and closes the remaining one.

Two premises AC-CTO-007 rests on were unverified by the SPEC. **Both hold — I measured them:**

- `TestReadOAuthToken_Keychain` (`internal/statusline/usage_test.go:180`) calls
  `t.Skip("TODO: mock security command for unit testing")` **unconditionally** (`:186`) — it skips
  on every OS, every run, with no env dependence.
- `TestProfilePhaseDistributions` (`internal/statusline/profile_bench_test.go:304`) skips unless
  `MOAI_PROFILE_PHASES == "1"` (`:305-307`); neither `release-pr-multi-os.yml` nor `ci.yml` sets
  that variable.

The dispatch target `release-pr-multi-os.yml:189` runs `go test -race -timeout 25m ./...`, which
includes `internal/statusline`. So the ONE approved dispatch run will contain at least two
genuine, named skips on all three matrix legs, and the brief's worry — "what if the single run
happens to contain no skips" — does not materialize. AC-CTO-007's satisfiability from one run is
therefore real, not merely asserted at `spec.md:122`.

**But the SPEC never established this.** `spec.md:37` measured both tests skipping on a local
macOS `go test` run and inferred the CI behaviour. One of the two is env-gated, so the inference
was not safe — it happens to hold. Recorded as a Gap, not a defect, because the premise is true;
the SPEC should nonetheless cite the skip conditions rather than the local observation.

**The defect is not in AC-CTO-007. It is that two of its sibling ACs cannot be reached at all
under the same ONE-run constraint** (D3), which is where the pressure to argue something into a
pass will actually land.

---

## Defects Found

**D1 — the fallback path discharges the terminal AC after the SPEC is already closed**
`spec.md:123` + `plan.md:67` + `acceptance.md:108-109` — Severity: **major** — Class: **blocking**
The fallback observation is a post-merge `ci.yml`-on-`develop` run. The Definition of Done
(`acceptance.md:108`) requires AC-CTO-001 … AC-CTO-008 all PASS at close, and `:109` requires
AC-CTO-007 recorded in `progress.md`. On the fallback path AC-CTO-007's evidence does not exist
until after the merge, so DoD is unsatisfiable at the moment it is evaluated. The SPEC gets the
honesty requirement right — `spec.md:123` and `plan.md:67` both forbid presenting the post-merge
observation as pre-merge, and `acceptance.md:83` item 4 makes the path a recorded field — but it
never resolves the ordering, and it never marks taking the fallback as **debt**.
*Required fix:* state explicitly that the fallback path closes the SPEC as PASS-WITH-DEBT with
AC-CTO-007 open, name the debt record surface, and name who discharges it after the develop push
— or state that the card does not close until the post-merge run is read.

**D2 — AC-CTO-006's `Given` never occurs in this repository**
`acceptance.md:66` — Severity: **major** — Class: **blocking**
"**Given** the change opened as a PR … Evidence: the check-name listing for the PR head." Under
this repo's git-flow lane model there are no card PRs: `.claude/rules/local/gitflow-lane-protocol.md`
§1 and `.claude/rules/local/repo-local-pr-policy.md` both state card worktrees branch from
`develop` and integrate by `git merge --no-ff` into local `develop` with **no card-level PRs**;
`main` advances only through release PRs. No PR head will ever exist for t358, so the criterion has
no green path — the failure shape `verification-completeness.md` §2 names as an unadopted criterion.
*Required fix:* re-anchor AC-CTO-006 to an artefact that exists on this path — e.g. the check-name
set reported on the `origin/develop` push run after integration, compared against the pre-change
name set — or state the criterion is discharged at the release PR and record it as deferred.

**D3 — AC-CTO-003 and AC-CTO-005 require a second, deliberately-red CI run that the ONE-run
constraint forbids**
`acceptance.md:42-46`, `:58-62`, `:74`, `:112` vs `spec.md:122` / `plan.md:63` — Severity:
**critical** — Class: **blocking**
AC-CTO-003 requires "the converted job runs it", "the job exit status", "the run's conclusion" —
GitHub Actions vocabulary, i.e. a CI run of a deliberately-failing tree. AC-CTO-005 is explicitly
chained to it: "**Given** the run from AC-CTO-003 (non-zero conclusion)" and requires "that run's
artifact list", which has no local analogue at all. Meanwhile `acceptance.md:74` now states
`[HARD] This AC is satisfiable from ONE run`, `spec.md:122` states "the dispatch is ONE run", and
DoD `:112` states "The positive control was obtained from ONE dispatch run". That is a minimum of
two CI runs demanded by the AC set and one permitted by the constraint. The SPEC never reconciles
them: `plan.md:51` says "M2's terminal check is a deliberately failing run" and the §F risk table
(`plan.md:77`) repeats it, but neither says whether that run is local or CI, nor how it fits inside
the single approved dispatch. The recent revision sharpened the ONE-run language without touching
AC-CTO-003/005, widening the contradiction.
*Required fix:* decide and state it. Either (a) the ONE-run constraint governs only the
positive-control dispatch and a separate red dispatch is separately approved — then say so in
`spec.md` §G and get that approval named; or (b) AC-CTO-003 is discharged locally (rewrite its
`Given/Then` in local vocabulary: process exit status + captured stdout, not "the run's
conclusion") and AC-CTO-005 is re-anchored to the same single positive-control run using
`if: always()` upload rather than to a red run.

**D4 — the plan's prescribed exit-code-capture form is broken under the shell it targets**
`plan.md:13` — Severity: **major** — Class: **blocking**
§B.1 prescribes `go test -json … > f; rc=$?; … ; exit $rc` as the mitigation for the
pipefail/`tee` hazard, and the §F risk table (`plan.md:78`) cites it as the mitigation. GitHub
Actions runs `shell: bash` steps as `bash --noprofile --norc -eo pipefail {0}` — the same flag
string the SPEC already cites for `-o pipefail` at `spec.md:109` and `plan.md:13`. It also carries
**`-e`**. Under `set -e`, `cmd > f; rc=$?` aborts at `cmd` when `cmd` fails: `rc=$?` never
executes, the census never prints, and the step dies before `exit $rc`. The step still ends red,
so no false green results — but REQ-CTO-004 (print failing test names) is silently defeated on
exactly the run where it matters, which is the regression AC-CTO-003 exists to catch. Compounding
with D3: the AC designed to catch this is the one that cannot be run.
*Required fix:* prescribe a form that survives `-e` — `rc=0; go test -json … > f || rc=$?` or an
explicit `set +e` / `set -e` bracket — and state the full flag string (`-eo pipefail`) rather than
`pipefail` alone, since citing half the flag set is what let this through.

**D5 — pushing the card branch to `origin` is a git-flow lane deviation the SPEC does not name**
`spec.md:120` + `plan.md:61` — Severity: **major** — Class: **blocking**
M4 depends on the card branch existing on `origin` so `workflow_dispatch` can resolve the ref, and
the step is present and correctly ordered first (`plan.md:61` "Push the card branch, dispatch…").
But under `.claude/rules/local/gitflow-lane-protocol.md` §1/§4 and `CLAUDE.local.md` §4.1, card
branches are **never** pushed — lanes push `origin/develop` only. The SPEC records the lead's
approval for **dispatching** (`spec.md:122`) but approving a dispatch is not approving a remote
card-branch push; they are separate acts with separate consequences (a stray remote branch that
nothing in the lane protocol disposes of). One mitigating fact the SPEC could have cited and does
not: `ci.yml` push triggers are `[main, develop]` (verified: `grep -n workflow_dispatch
.github/workflows/ci.yml` → rc=1; triggers confirmed at `ci.yml:20`ff), so pushing
`WT-ci-test-observability` triggers no workflow — the push is CI-inert.
*Required fix:* name the deviation in `spec.md` §G, state that it is CI-inert with the trigger
evidence, obtain and record the lead's approval for the push specifically (not only the dispatch),
and state who deletes the remote branch after M4.

**D6 — the stdout-only stderr discipline is carried by no REQ and verified by no AC**
`plan.md:14` + `spec.md:110` — Severity: **minor** — Class: **blocking**
Both surfaces get the substance right: `plan.md:14` mandates redirecting **stdout only** and states
that `2>&1` would both hide toolchain errors and corrupt the stream; `spec.md:110` says the summary
step "must not swallow stderr". The claim in the brief is verified — it is in the text. But it lives
only in design-constraint prose. REQ-CTO-001 (`spec.md:88`) says "emit … to a file, not to the
console" and REQ-CTO-003 (`:90`) says "shall not write the raw per-test event stream to the console
log"; neither constrains stderr, and no AC in `acceptance.md` tests it. A `2>&1 > f` implementation
would satisfy every stated requirement and every acceptance criterion while destroying the property
`plan.md:14` calls load-bearing.
*Required fix:* add a requirement (Unwanted form: "The CI job shall not redirect the test
invocation's standard error into the event-stream file") and an AC that observes a build or
toolchain error remaining console-visible.

**D7 — a fourth `go test ./...`-class call site in `ci.yml` is in neither scope nor exclusions**
`spec.md:115` — Severity: **major** — Class: **blocking**
"The scope is confirmed at **exactly these three sites**, not more." `grep -n "go test"
.github/workflows/ci.yml` returns four executable invocations: `:183`, `:238`, `:329`, and the
excluded `-run`-scoped ones. `ci.yml:329` is `go test -tags=integration -race -timeout 180s
./test/integration/harness/...` in job `test-integration` (`name: Integration Tests (${{ matrix.os
}})`, 3-OS matrix, `shell: bash`) — not `-run`-scoped, carries no `-v`, and has precisely the
observability gap §A describes. It appears in no §F row and in no §H exclusion, so the
"exactly three, not more" claim is a scope census that missed a site rather than a decision.
*Required fix:* either add `ci.yml:329` to §F, or add an `### Out of Scope — …` entry naming it
with a stated reason. Do not leave it unmentioned.

**D8 — the SPEC claims Tier M but carries no `tier:` field, so every mechanical consumer reads
Tier L**
`spec.md:2-13` vs `plan.md:3` and `progress.md:7` — Severity: **major** — Class: **blocking**
`plan.md:3` opens "Tier **M**" and `progress.md:7` repeats "(Tier M)", but `grep -n '^tier:'
spec.md` → rc=1. Per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields
and `spec-workflow.md` § SPEC Complexity Tier, an absent `tier:` means the SPEC is treated as
**Tier L**. That changes three mechanical things: the plan-auditor PASS threshold (0.85 not 0.80),
the retry ceiling, and the plan-artifact hash subject set for skip-eligibility. The claimed tier
exists only in prose. This is the single finding most likely to change this report's verdict —
at 0.81, the SPEC clears Tier M and fails Tier L.
*Required fix:* add `tier: M` to the frontmatter, or accept Tier L and its 0.85 threshold. Do not
leave the classification prose-only.

**D9 — Tier M is asserted without a scope-to-threshold justification, and the measured scope points
to Tier S**
`plan.md:3` + `plan.md:9` — Severity: **minor** — Class: **optional**
`plan.md:9` states the surface: "2 workflow files (3 invocation sites) + 1 summary implementation.
No Go production code, no data model, no user-facing UX." With a fixture and a unit-level check
that is roughly 5 files and well under 300 LOC of YAML plus a jq script — Tier S territory (<300
LOC, <5 files), sitting exactly on the 5-file boundary where the `orchestration-mode-selection.md`
§B.2 tie-breaker resolves toward the simpler classification. Tier M is defensible on ceremony
grounds (the CI-observation milestone is real work), but `plan.md:3` asserts it with no mapping to
the thresholds at all. Not blocking: the artifact set authored (3 files + progress.md) is
internally consistent with M, and tiering up is the safe error.
*Required fix (optional):* one sentence in `plan.md` §A mapping the estimated file count and LOC to
the tier table, so the classification is auditable rather than asserted.

**D10 — AC-CTO-008 traces to a plan milestone, not to any requirement**
`acceptance.md:16` — Severity: **minor** — Class: **optional**
The AC matrix column reads "plan §M3 replication". Every other AC names a REQ. No requirement in
`spec.md` §E states that the change must reach all three call sites — §F is a table of call sites,
not a normative statement — so AC-CTO-008 verifies an obligation no requirement imposes. The
substance is right and the AC is well-written (it now also asserts the four excluded steps are
untouched, `acceptance.md:100`); only the traceability link is missing.
*Required fix (optional):* add a Ubiquitous requirement ("The census, file-redirect, exit-code
capture, and artifact upload shall be applied at every in-scope call site") and retarget
AC-CTO-008 to it.

**D11 — `REQ-CTO-004` is labelled with a non-existent GEARS pattern**
`spec.md:91` — Severity: **minor** — Class: **optional**
Labelled "(Event-detected)". The five GEARS patterns are Ubiquitous, Event-driven, State-driven,
Where (capability gate), and Unwanted. The sentence itself is correct event-driven form ("When a
test failure is detected in the event stream, the summary step shall print…"), so MP-2 passes on
the requirement text; only the annotation is wrong.
*Required fix (optional):* change the label to "(Event-driven)".

---

## Gaps (explicitly not observed)

- The GitHub Actions `shell: bash` flag string `bash --noprofile --norc -eo pipefail {0}` is cited
  from documented runner behaviour, not measured on a runner in this session. D4's mechanism
  depends on the `-e` in that string. The SPEC's own `spec.md:109` cites `-o pipefail` from the
  same string, which is corroborating but not independent.
- No CI run was executed or read during this audit. Every claim about what a dispatch would
  produce is derived from the workflow files and test sources in this tree at
  `WT-ci-test-observability`, not from a run.
- The `test` / `test-skip-marker` mutual-exclusivity invariant was read from the gating conditions
  (`ci.yml:108` `if: needs.detect.outputs.go_code == 'true'` vs `ci.yml:266` `if:
  needs.detect.outputs.go_code != 'true'` — strict negation, both emitting `Test (${{ matrix.os
  }})`) but was not exercised. The proposed change adds steps inside `test` only and renames no
  job, so required-check parity is not at risk from anything the SPEC proposes. Adjacent
  observation, not a defect: on the skip-marker path the required check `Test (ubuntu-latest)`
  still reports green with nothing run and no census — the exact `rc=0`-says-nothing shape §A
  describes. That path already carries its own `$GITHUB_STEP_SUMMARY` explanation (`ci.yml:272-280`),
  so the gap is mitigated by an existing mechanism the SPEC neither checked nor mentioned. It
  would belong in §H.
- The full-suite `-json` artifact size remains unmeasured; `spec.md:49` states this correctly as an
  extrapolation and `acceptance.md:62` schedules the real figure. No defect — this is the SPEC
  handling an unknown honestly, and it is the strongest example of that in the document.

## Minor citation drift (not scored)

- `spec.md:129` and `acceptance.md:113` cite the concurrency policy at `ci.yml:29-31`. Measured:
  line 29 is the comment, `concurrency:` is line 30, `group:` 31, and `cancel-in-progress: true`
  is line **32** — the cited range excludes the line the claim is about. Correct range: `ci.yml:30-32`.
- `spec.md:109` says `shell: bash` is load-bearing "for the windows leg (see the comment at
  `ci.yml:157`)". The comment is verbatim at `:157` ✓, but ci.yml's `test` job matrix is
  ubuntu-latest only (confirmed by the `test-skip-marker` comment at `ci.yml:246-249`, "matrix
  narrowed to ubuntu-latest to mirror the real `test` job's matrix exactly"). The windows rationale
  applies to `release-pr-multi-os.yml`, not to "these jobs" collectively.
- Verified correct and worth noting since they were checked: `ci.yml:183` = `go test
  -coverprofile=coverage.out -covermode=atomic ./...` ✓; `ci.yml:238` = `go test -race -count=1
  ./...` ✓; `release-pr-multi-os.yml:189` = `go test -race -timeout 25m ./...` ✓;
  `release-pr-multi-os.yml` `workflow_dispatch` at line 16 ✓ with gates at 35 and 86 ✓;
  `ci.yml:433-438` is the `upload-artifact@v7` + `retention-days: 7` block ✓;
  `lsel-leak-guard.yaml:37` carries `-v` ✓ and all three `template-neutrality-check.yaml` steps
  (`:60`, `:70`, `:87`) carry `-v` ✓.

---

## Recommendation

Return to manager-spec. Fix in this order — the first three are the ones that decide whether this
SPEC can be discharged at all:

1. **Resolve D3.** Decide whether AC-CTO-003 / AC-CTO-005 are local-vocabulary criteria or require
   a separately-approved second dispatch, and rewrite their `Given/Then` accordingly. This is the
   single largest defect: two criteria currently have no reachable starting state, and the ONE-run
   language added in the last revision made the collision worse rather than better.
2. **Fix D4** — the `-e` hazard in `plan.md:13`. Prescribe `rc=0; go test -json … > f || rc=$?`
   and cite the full `-eo pipefail` flag string. Note that this defect and D3 compound: the
   criterion designed to catch it (AC-CTO-003) is one of the two that cannot be run.
3. **Re-anchor AC-CTO-006 (D2)** to the `origin/develop` push run, since no card PR will exist.
4. **Add `tier:` to the frontmatter (D8)** — this determines the threshold the re-audit applies.
5. **Name the branch-push deviation (D5)** in `spec.md` §G with the CI-inert trigger evidence, and
   record lead approval for the push specifically.
6. **Account for `ci.yml:329` (D7)** — §F row or §H exclusion, either is fine; silence is not.
7. **Give the stderr discipline a REQ and an AC (D6).**
8. **Resolve D1** — state the fallback path as PASS-WITH-DEBT with a named debt surface.
9. Optional, at the author's discretion: D9 (tier justification), D10 (AC-CTO-008 traceability),
   D11 (label).

**What is genuinely strong and should not be touched in revision:** §A.1's measured-evidence table
with its explicit extrapolation disclaimer (`spec.md:35-49`); §C.1's single-predicate rationale,
which records the *deciding* reason rather than a generic "JSON is machine-readable"
(`spec.md:59-71`); AC-CTO-002's mutant probe (`acceptance.md:34`); and AC-CTO-007's six named
failing forms (`acceptance.md:87-92`). That last one does what the card asked for — it is not
arguable into a pass, and I could not find a way to satisfy it with a local run, a grep, a count,
or a skipless run. The SPEC's problem is not its terminal control; it is the three criteria
standing beside it that cannot be reached.

The re-audit (iteration 2) will be scoped to this enumerated defect delta plus a regression check
over D1-D11, not a from-scratch review.
