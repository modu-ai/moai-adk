auditor-model: glm-5.3-flash

# SPEC Review Report: SPEC-FACTORY-LEASE-BUSY-001
Iteration: 2/3 (delta re-audit)
Verdict: FAIL
Overall Score: 0.90
Plan Artifact Hash: 268df9905d9216007ae7bfed0467f0bbcefcf0128e31fd33e26b245cc20b49d7 (SHA-256 over spec.md + plan.md + acceptance.md concatenated in that order, Tier M set, at HEAD ad8603214c)
Auditor Version: plan-auditor/v3.2 (served by glm-5.3-flash; independent of manager-spec)

verdict: FAIL
audited_sha: ad8603214cadccf48e4d0cc39140b626b07e53ed
overall_score: 0.90
must_pass_failed: 0
blocking_count: 2
plan_artifact_hash: 268df9905d9216007ae7bfed0467f0bbcefcf0128e31fd33e26b245cc20b49d7
scope: delta
fix_scope: acceptance.md#AC-FLB-009, plan.md#§F-M4
defect_class: ac-wording (D12); design (D13)
reread_hunks: acceptance.md#AC-FLB-009, acceptance.md#§C, plan.md#§F-M4
convergence_overall: fail
required_backend: codex fail

Delta basis: iter1 (FAIL 0.69, D1-D10) and its reread_hunks; repairs applied in commit
ad8603214c ("apply plan-audit iter1 repairs D1-D10"), whose diff against 61b5af4b1 was read in
full (spec.md 37 lines, acceptance.md 66, plan.md 41, progress.md 7 changed). Reasoning context
from the SPEC author was not present in the spawn input; M1 Context Isolation holds. Tier M
input contract: spec.md + plan.md + acceptance.md read in full (+ progress.md, decision-index.md).

## Must-Pass Results (re-measured this run where the delta could touch them)
- [PASS] MP-1 REQ number consistency: `grep -oE 'REQ-FLB-[0-9]+' spec.md | sort -u` → exactly
  REQ-FLB-001..012, 12 unique, sequential, zero-padding consistent. The repairs renumbered nothing.
- [PASS] MP-2 EARS/GEARS format compliance (requirement layer, spec.md § Requirements; ACs are the
  verification layer, graded under Group 4): the four repaired REQs — REQ-FLB-004 is Event-driven
  ("When a lease claim fails …, the lease path shall refuse", spec.md:94); REQ-FLB-005 is the
  legacy EARS negative form ("… shall not surface"), counted as a legacy equivalent inside the
  backward-compatibility window (through 2026-11-22; today 2026-10-21), spec.md:100; REQ-FLB-008
  is Event-driven (spec.md:109); REQ-FLB-009 is the canonical GEARS Where-pattern ("Where the
  lease path decides on a serial-slot outcome …, the lease path shall attribute no race loss",
  spec.md:117). No IF/THEN anywhere; no informal "should/may" in normative text.
- [PASS] MP-3 YAML frontmatter validity: untouched by the repair diff; all 12 canonical fields
  re-read at spec.md:2-13 with correct types, no snake_case aliases.
- [N/A] MP-4 language neutrality: single-language (Go) SPEC — auto-passes per LN-3.
- [PASS] MP-5 D7 cross-SPEC reconciliation: re-run — 4 grep hits = the SPEC's own id plus
  SPEC-FACTORY-ATOMIC-LEASE-001 (in-progress), SPEC-TODO-AUTO-PICK-001 (completed),
  SPEC-FACTORY-LEASE-RENEW-001 (completed); zero REVIEW lines (nothing retired/superseded/
  archived), zero SHOULD lines.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall` over spec/plan/acceptance → 0/0/0.
  Auto-pass per D8-4 (observed absence, re-measured).
- [PASS] MP-7 clarification gate: `grep -n 'NEEDS CLARIFICATION' plan.md` → exit 1 (no match) on
  the repaired plan.md; research.md absent, the correct Tier M artifact set (directory listing
  read: acceptance, decision-index, plan, progress, spec).
- [N/A] MP-8 RED-now cell completeness: `grep -c 'release-blocking' acceptance.md` → 0; no
  criterion is classified release-blocking, so the four-element RED-now obligation does not bind
  (reason stated per the MP-4 precedent). No command cited in any RED-now cell was executed at
  audit time.
- [PASS] MP-9 cross-artifact ordering consistency: the CN-4 awk verb was NOT re-attempted (the
  worktree isolation guard refused it in iter1; refused commands are not retried — the refusal is
  named per VCI §3.1). The plain-grep substitution was re-run fresh against the changed
  acceptance.md: `grep -inE 'before|after|first|prior to|pre-change'` → 11 candidates (L5, 36,
  45, 46, 50, 66, 87, 88, 92, 121, 126), each hand-read against plan §F order M0→M1→M2→M3→M4→M5→M6
  (no `Exit:` bindings exist in the plan, so no CONFLICT class is reachable). L66 (AC-FLB-008,
  "baseline … precedes the first commit modifying the read path") is jointly satisfiable with
  M4-before-M5; L121 ("RED output … before its GREEN") matches M1→M2; L126 ("M0 gate … before any
  implementation commit") matches M0-first; L45/46/50 are AC-FLB-006's intra-invocation failure
  arms and L87/88/92 AC-FLB-011's intra-invocation retry sequence — none orders work across
  milestones; L36's pre-change capture precedes the M2 flip via M1's RED phase. Zero conflicts.

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.00 | single-interpretation requirements | All four repaired REQs are single-interpretation and code-anchored: REQ-FLB-004 names the observed signals and asserts no unconditional cause (spec.md:94-99, D4 closed); REQ-FLB-005 scoped to the cause table with the queue-lock exception inline (spec.md:100-103, D2 closed); REQ-FLB-008's wait-retry clause anchored to the 1270-1276 predicate (spec.md:109-116, D3 closed); REQ-FLB-009 names both real decision sites (spec.md:117-122, D8 closed). AC-FLB-001's deferred token strings remain a deliberate M1 deliverable (plan §F M1), as iter1 accepted |
| Completeness | 0.85 | substantive gaps closed, narrow residue | D3 (--wait surface), D6 (recipe executability), D7 (text preservation ×3 classes), D9 (ownership routing) all closed — see Regression Check. Residue: D13 (the repaired M4 recipe carries an internal directive collision), D11 (stale AC-count annotations the repair sweep missed), D15 (M4 root/fixture clause tension) |
| Testability | 0.75 | one or two ACs need interpretation | D1/D5/D10 closed (AC-FLB-006 split both-ways testable; AC-FLB-009's ceiling blocks the 10x-regression mutant; AC-FLB-010 name-level). Residue: AC-FLB-009's "exactly one" quantifier fails the best case in the floor-escape world and is rescued only by §C's edge-case note (D12 — a judgment call the criterion should not require); AC-FLB-007's correlational quantifier can false-red under a competitor composition its Given admits (D14). The other nine ACs are binary-testable with no weasel words |
| Traceability | 1.00 | full coverage | acceptance.md §B (lines 96-103) re-read: all 12 REQs mapped, REQ-FLB-008 → AC-FLB-006 + AC-FLB-011 (the new row), AC-FLB-010 regression-declared, no orphan AC; `grep -oE 'AC-FLB-[0-9]+'` → 001..011 unique; DoD #1 updated to "All eleven ACs" (acceptance.md:125). The traceability awk verb was not re-attempted (iter1 guard refusal, not retried); the mapping was hand-verified from the full file read — the hand-read is the measurement |

Aggregate 0.90 ≥ Tier M threshold 0.80, but 2 blocking defects (D12, D13) force FAIL. Score
monotonicity holds: 0.69 → 0.90 (no STOP signal).

## Defects Found (structured defect-list)

New findings this iteration (D-numbers continue iter1's sequence; D1-D10 dispositions in the
Regression Check section):

D11. FLB-D11 — plan.md:82 ("all ten AC-FLB ids"), plan.md:150 ("AC-FLB-001..010"), spec.md:36
  ("criterion count 10") — stale AC-count annotations the AC-FLB-011 repair sweep missed:
  acceptance.md carries eleven criteria (AC-FLB-001..011) and DoD #1 says "All eleven", but plan
  §E still directs manager-develop to report ten rows and plan §H/spec HISTORY still say ten.
  An E1 matrix built from plan §E omits AC-FLB-011's evidence row, failing DoD #1 at close.
  Related rider: decision-index.md:44 still cites "tests 443/464" by line number — the same
  anchor-drift class D10 fixed elsewhere, outside iter1's fix_scope. — Severity: minor — Class:
  optional — Required fix: three token edits (ten→eleven, 001..010→001..011, 10→11) plus the
  decision-index anchor note at the next manager-spec touch.
D12. FLB-D12 — acceptance.md:70-74 (AC-FLB-009) vs acceptance.md:114-115 (§C "Latency floor
  escape") — the criterion's operative quantifier "exactly one recorded outcome holds" is false
  in the floor-escape world whenever the fix also improved: baseline p95 = 100 ms, post-change
  p95 = 40 ms makes outcome 1 true (40 ≤ 50) AND outcome 2 true (100 < 200 and 40 ≤ max(200,
  200)), so "exactly one" fails a 2.5x improvement — while §C says the same world "records
  'defect not reproduced' — a named pass exit". The criterion's letter and §C give opposite
  verdicts for the same measurement. (Surfaced by the codex convergence leg with a computed
  example; confirmed independently by arithmetic on the text — iter1 graded only the
  regression-mutant direction, D5, and missed this exclusivity direction; so did my first
  pass.) The overlap pre-dates the repair, but the repaired ceiling clause sharpened outcome 2
  without reconciling the quantifier. — Severity: major — Class: blocking — Required fix: fold
  the selection precedence into the criterion itself — e.g. "the recorded outcome is selected
  by the baseline first: when baseline p95 < 200 ms the outcome is 'defect not reproduced'
  (with the post-change p95 recorded at or below max(2× baseline p95, 200 ms)); otherwise the
  outcome is 'defect fixed' when post-change p95 ≤ 50% of baseline p95" — replacing "exactly
  one … holds", and align REQ-FLB-012 (spec.md:135-139) to the same conditional shape (it still
  carries the ceiling-free second arm; AC-stricter-than-REQ parity).
D13. FLB-D13 — plan.md:124 vs plan.md:128-130 (M4) — directive collision inside the repaired
  recipe: clause (1) requires building the binary per revision (`go build -o
  <scratch>/moai-<revision> ./cmd/moai`) "and invoked by that path", while clause (4) requires
  "the full invocation line recorded verbatim in the baseline artifact and reissued verbatim,
  unchanged, for the post-change measurement". The full invocation line contains the binary
  path — the one token that MUST change between revisions. Literal compliance reissues the
  baseline's command and measures the baseline binary twice, silently corrupting the
  AC-FLB-008/009 comparison (the exact carry-over/stale-instrument failure VCI §2 prohibits);
  updating the path "violates" clause (4)'s letter. (Surfaced by the codex convergence leg;
  confirmed from the text.) — Severity: major — Class: blocking — Required fix: name the binary
  path as the sole revision-varying measurement parameter ("every other argument, environment
  variable, and fixture input is byte-identical across the two runs"), and provide the fixture
  seed and root setup as executable commands in the recipe.
D14. FLB-D14 — acceptance.md:61 (AC-FLB-007) — the then-clause's correlational quantifier ("no
  rendered line of either outcome carries a race claim") is broader than REQ-FLB-009's causal
  verb ("a read failure … produces no race claim"): under a competitor composition the
  criterion's Given admits but does not mandate (peek read fails open → the attempt proceeds →
  a genuine race is lost downstream), REQ-FLB-002 itself mandates the `raced` line the
  criterion's letter forbids. Within the criterion's own defined scenario (forced peek
  failure, no mandated competitor) it is satisfiable and correctly gates the hygiene
  requirement, so this is a precision defect with conditional bite, not an operative
  contradiction. (Surfaced by the codex convergence leg with a Go-overlay reproduction I did
  not re-execute; confirmed from the code anatomy measured in iter1 R4 — fail-open peek,
  version-checked claim — and the two texts.) — Severity: major — Class: optional — Required
  fix: scope the prohibition causally — "the rendered line for the peek-failure outcome
  attributes no race claim to the read failure; a subsequent genuine race loss keeps its
  `raced` per REQ-FLB-002". Recommended as a rider in the same iter3 repair commit.
D15. FLB-D15 — plan.md:126-129 (M4 clauses (2)/(3)) — clause (2) sets "a fresh `mktemp -d`
  project directory as the factory root for every run" while clause (3) has the fixture
  "created once in that root … reused byte-identically by both runs": fresh-root-per-run vs
  created-once read in tension. Both readings yield comparable measurements for a read-only
  command and the recorded verbatim commands make the actual choice auditable, so this is a
  wording wrinkle, not a measurement hazard. — Severity: minor — Class: optional — Required
  fix: one clarifying clause ("the same seed step re-creates the identical fixture set in each
  fresh root" or "one root serves both runs; `factory status` is read-only").

## Regression Check (Iteration 2)
Defects from iteration 1 (D1-D10) — each verified against the artifact line now on disk:

- D1 (AC-FLB-006 unsatisfiable on the mid-claim arm) — [RESOLVED]: acceptance.md:44-54 rewrites
  the criterion split by failure point; the pre-record-write arm restores `queued` (citing
  TestFactoryLeaseRecordStallBounded's 155-157) and the mid/post-write arm preserves
  `assigned`/`picked` (citing TestFactoryLeaseMidClaimStallBounded's 302-307) — each arm now
  matches the assertion it cites, both directions testable.
- D2 (REQ-FLB-005 vs queue-lock Out of Scope) — [RESOLVED]: spec.md:100-103 scopes the
  prohibition to "the claim-failure classification of this SPEC's cause table" AND names the
  queue-lock refusal as sitting outside it (both halves of iter1's required fix).
- D3 (waitable() third surface unaddressed) — [RESOLVED]: spec.md:111-116 adds the
  wait-retryable disposition to REQ-FLB-008 anchored at the 1270-1276 predicate; plan.md:106-108
  adds the M2 waitable() clause citing AC-FLB-011; acceptance.md:87-92 adds AC-FLB-011
  (retry-then-exit-0); acceptance.md:99 adds the §B row "REQ-FLB-008 | AC-FLB-006, AC-FLB-011".
- D4 (REQ-FLB-004 mandated a causal narrative the evidence contradicts) — [RESOLVED]:
  spec.md:96-99 now mandates signal-naming ("the expired claim deadline and, as recorded, the
  presence or absence of a concurrent busy error … asserting neither a load-only nor a
  race-only cause"); AC-FLB-002 (acceptance.md:23-26) aligns verbatim.
- D5 (AC-FLB-009's not-reproduced arm unbounded) — [RESOLVED]: acceptance.md:70-74 adds "the
  post-change p95 is recorded at or below max(2× baseline p95, 200 ms)"; the 10x-regression
  mutant now fails (at baseline 50 ms: 500 > 200; at baseline 190 ms: 1900 > 380). (The arm's
  exclusivity defect is a separate axis — new D12.)
- D6 (M4 recipe not executable) — [RESOLVED]: plan.md:122-131 adds the four elements
  (per-revision binary build invoked by path, isolated mktemp root, fixed fixture, verbatim
  record-and-reissue exact command); AC-FLB-008 requires the artifact to carry the recipe.
  (The collision inside clauses (1)/(4) is new D13.)
- D7 (text preservation unmeasured for 3 of 4 classes) — [RESOLVED]: acceptance.md:31-39
  extends AC-FLB-004 with byte-identical recorded-prior-text comparison for quota-hold,
  backend-skip, and foreign-worktree, the pre-change capture pinned to a tree SHA in progress.md
  §E.2 (VCI §4-compliant pinning), and adds backend-skip to the class list.
- D8 (REQ-FLB-009 named no locatable event) — [RESOLVED]: spec.md:117-122 re-scopes to
  race-claim hygiene over the two real sites (in-memory slot check 330/1368; fail-open peek
  1857-1865); acceptance.md:55-62 gives the induction method (forced peek store-read failure);
  plan.md:111-118 re-scopes M3 and states the REQ-FLB-010 collision guard iter1's advisory
  flagged.
- D9 (M1 deliverable assigned a plan.md edit to manager-develop) — [RESOLVED]: plan.md:98-101
  moves the token/class table deliverable to progress.md §E.2 with the ownership reason stated
  and notes the classes are already normative in REQ-FLB-001.
- D10 (AC-FLB-010's 443/464 line-number anchors) — [RESOLVED]: acceptance.md:80-82 pins by
  function name and subtests ("TestFactoryLeaseQueueLockStallBounded with both its subtests
  `nominated` and `bare`"); plan.md:44-48 and spec.md:145-148 carry the same name-level form.

All ten iter1 defects RESOLVED; the two new blocking defects (D12, D13) are newly surfaced
instrument defects, not regressions of the prior list.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

| # | Claim | Evidence (command → verbatim output) | Baseline-attribution |
|---|-------|--------------------------------------|----------------------|
| R1 | Tree identity matches the audited_sha | `git rev-parse HEAD` → `ad8603214cadccf48e4d0cc39140b626b07e53ed`; `git status --porcelain` → empty | This run, this worktree |
| R2 | The repair commit's exact subject set | `git diff --stat 61b5af4b1..ad8603214c` → 6 files (spec 37, plan 41, acceptance 66, progress 7 lines, 2 report files); full diff read | This run |
| R3 | REQ/AC id sets | `grep -oE 'REQ-FLB-[0-9]+' spec.md \| sort -u` → REQ-FLB-001..012 (12); `grep -oE 'AC-FLB-[0-9]+' acceptance.md \| sort -u` → AC-FLB-001..011 (11) | This run |
| R4 | D7/MP-5/MP-6/MP-7 measurements | cross-SPEC grep → 3 refs, statuses in-progress/completed/completed; `grep -c syscall` → 0/0/0; `grep -n 'NEEDS CLARIFICATION' plan.md` → exit 1; `grep -c 'release-blocking' acceptance.md` → 0 | This run |
| R5 | CN-4 substitution | `grep -inE 'before\|after\|first\|prior to\|pre-change' acceptance.md` → 11 candidates (L5, 36, 45, 46, 50, 66, 87, 88, 92, 121, 126), hand-read, zero conflicts | This run |
| R6 | Convergence | `moai verify audit-plan` → config_status ok, codex gate required, cross_model_active true; `mcp__moai__audit_multi` (target = the repair diff 61b5af4b1..ad8603214c, no claude_verdict — GLM-served session) → codex verdict fail (3 findings, receipt rcpt-352a3910c3ca6010710839f5); digest written to `.moai/state/audit-plan-result.json` and `--result-file` re-check → `convergence_check: {ok: true, unmet: []}` — the required backend answered; the gate is met; its fail verdict is recorded truthfully and adjudicated above (F1→D14, F2→D12, F3→D13) | This run, tree ad8603214c |
| R7 | D12 arithmetic | read from the text: base=100/post=40 → outcome 1 (40 ≤ 50) true; outcome 2 (100 < 200 ∧ 40 ≤ max(200, 200)) true → "exactly one" false; §C:114-115 calls the same world a "named pass exit" | This run, text-arithmetic on acceptance.md:68-74 and 114-115 |
| R8 | Plan-artifact hash | `cat spec.md plan.md acceptance.md \| shasum -a 256` → 268df9905d9216007ae7bfed0467f0bbcefcf0128e31fd33e26b245cc20b49d7 | This run |

**Gaps**: (1) the CN-4 awk verb and the traceability awk verb were refused by the worktree
isolation guard in iter1 and were NOT retried (refused commands are not retried per the standing
instruction); the substitutions — plain grep + hand-read of all 11 ordering candidates, and the
§B mapping hand-verified from the full file reads plus the R3 id greps — are named here per VCI
§3.1. (2) codex's Go-overlay transient verification of its F1 scenario was not re-executed by me;
D14 rests on my own read of the code anatomy (fail-open peek measured in iter1 R4,
version-checked claim path) and the two artifact texts. (3) build_lag notice: the installed moai
binary was built from 8249bff8c, an ancestor of this tree's HEAD — no moai CLI measurement is
cited in this verdict, so no claim rests on the stale build. (4) No RED-now cell was re-executed
(MP-8 discipline; no criterion is release-blocking).

**Residual-risk**: iteration 3 of 3 is the last before the ceiling policy; both blocking fixes
are narrow (one AC clause + precedence fold + the REQ-FLB-012 alignment rider for D12; one M4
clause for D13), and D11/D14/D15 are one-pass riders. If iter3 leaves either blocking defect
unresolved, the ceiling policy (auto_delta_rounds, then the admission ladder) applies — the fixes
carry scoped fix anchors, so the split arm stays reachable. The M0 sequencing hold (PR #1859
merged as 17a3b98de2; leader resume) remains the run-phase entry precondition and was not
re-litigated here, per the delta basis.

## decision-index routing assessment (re-checked against the changed text)
Q1's published default (exit 4 for every class, token-only discrimination) remains encoded in
REQ-FLB-002/003/004/007 ("exit code 4" in each) and measured by AC-FLB-004 — does not block. Q2
remains baseline-relative by design (AC-FLB-009, M4/M5). Q4 stands DEFAULT-APPLIED and the
changed text now cites it twice (REQ-FLB-005's exception, Out of Scope). Routing unchanged from
iter1; no new decision surfaced by the repairs.

## Corroboration re-run (stop-hook receipt check)
A second `mcp__moai__audit_multi` call against the same delta target returned codex verdict
`fail` with exactly the two blocking findings of this report, independently: D12 (confidence
1.00, Python-computed: baseline=100 ms, post=40 ms → both outcomes true → "exactly one" false,
conflicting with §C's "defect not reproduced" pass) and D13 (synthetic-executable reproduction:
the verbatim-reuse directive measured the baseline binary twice and required a path change for
the post-change run). No new finding beyond D12/D13/D14. Fresh receipt: rcpt-af990f9bc29438e3ba9c5918
(supersedes rcpt-352a3910c3ca6010710839f5 as the citation of record). Verdict unchanged.

## Recommendation (numbered fixes for manager-spec, iter3 scope)
1. D12: rewrite AC-FLB-009's outcome selection as baseline-first precedence (acceptance.md:68-74)
   and align REQ-FLB-012's second arm (spec.md:135-139) to the same conditional shape.
2. D13: in plan §F M4, name the per-revision binary path as the sole revision-varying token of
   the reissued command; pin every other argument, env var, and fixture input byte-identically;
   add the fixture-seed and root-setup commands to the recipe (plan.md:122-131).
3. Riders (non-blocking, same commit): D11 three-token count sweep (plan.md:82, plan.md:150,
   spec.md:36; decision-index.md:44 anchor note); D14 causal scoping of AC-FLB-007's prohibition
   (acceptance.md:61); D15 one-clause M4 root/fixture clarification.

Fit to enter run phase after M0: NOT in current form — the two blocking instrument defects
(D12, D13) must land first; the D1-D10 repair set itself is fully closed and the contract's
sequencing, decision routing, and traceability remain correctly structured.
