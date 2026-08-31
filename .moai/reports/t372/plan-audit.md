# SPEC Review Report: SPEC-STRESS-INVARIANT-VERDICT-001 (card t372)

Iteration: 1/1 (Tier S ceiling — `harness.yaml:76`)
Verdict: **FAIL**
Overall Score: **0.69** (harmonic mean of the four dimensions)
Tier threshold applied: **0.75** (Tier S, as declared in `spec.md:L14`; `spec-workflow.md:L140`)

Reasoning context ignored per M1 Context Isolation. Audit tree: `.claude/worktrees/t372`,
HEAD `9d4f79281`, branch `WT-stress-invariant-guard`, base `b9149857c`.
t370's measurements are consumed as ground truth and not re-measured; no CI was run and no
load was generated.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-SIV-[0-9]*' spec.md | sort -u` yields exactly
  `REQ-SIV-001 … REQ-SIV-017`: 17 entries, sequential, no gap, no duplicate, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judgment made against the **requirement layer**
  (`spec.md` §B–§D `REQ-SIV-*`). All 17 match a GEARS pattern: event-driven (001, 002, 007),
  ubiquitous (003, 005, 008–012, 014–017), unwanted `shall not` (006, 013), Where-compound (004).
  The `AC-SIV-*` entries in `acceptance.md` are Given-When-Then, the correct verification-layer
  format; they were graded under Group 4, **not** here. Label defects exist (see D10) but no
  requirement is informal or a test scenario.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:L2-L13`): `id`, `title`, `version: "0.1.0"` (quoted), `status: draft` (valid enum),
  `created`/`updated` ISO, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`,
  `tags` comma-separated. No rejected snake_case alias (`created_at`/`updated_at`/`labels`/`spec_id`).
  Optional `tier`/`depends_on` are additive.
- **[N/A → PASS] MP-4 language neutrality** — single-language SPEC (Go, `module: internal/kanban`).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference, `SPEC-BACKLOG-LOCK-BUDGET-001`;
  `.moai/specs/SPEC-BACKLOG-LOCK-BUDGET-001/spec.md` exists, `status: implemented` — not in
  {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` = `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-STRESS-INVARIANT-VERDICT-001/`
  → rc=1, no match. `research.md` absent (Tier S).

**No must-pass failure. The FAIL is score-driven**, from the defects below.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.65 | below 0.75 | REQ-SIV-008's verb "**covers**" (`spec.md:L121-123`) asserts sufficiency; `spec.md:L196-197` and `board_lock_wait_test.go:L7-8` both deny sufficiency — an internal contradiction on the SPEC's central new artifact (D5). AC-SIV-009's mutant menu (`acceptance.md:L94-96`) is ambiguous in two of three branches (D1, D3). |
| Completeness | 0.75 | 0.75 | All sections present (HISTORY L20, §A WHY L26, §B/§E scope, requirements L78-178, ACs in `acceptance.md`, four `### Out of Scope — …` H3s with bullets L184-216). Missing: no residual-risk clause for the over-broad Unix sentinel (D6), no attempt-conservation requirement (D2), tier budget breach unaddressed (D4). |
| Testability | 0.55 | 0.50 | AC-SIV-002/004/006/007/008/012 carry concrete verification verbs. AC-SIV-001 (`acceptance.md:L30-35`) and AC-SIV-005 (`L61-64`) have Givens the SPEC's own §E forbids producing (`spec.md:L200-201`, no load generation, no local reproduction) and name no verification verb — neither is dischargeable as written. The binding AC-SIV-009 admits a mutant that cannot reach the invariant block (D1). |
| Traceability | 0.90 | 0.75–1.0 | Every REQ-SIV-001..017 is covered by ≥1 AC (`acceptance.md:L12-24` matrix, verified against the REQ list). One orphan: AC-SIV-012 traces to "scope discipline (spec.md §E)", not to a REQ (D11). |

Harmonic mean: `4 / (1/0.65 + 1/0.75 + 1/0.55 + 1/0.90) = 0.6895` → **0.69 < 0.75**.

---

## Defects Found

### D1 — the binding invariant mutant can be discharged with non-discriminating evidence
`.moai/specs/.../acceptance.md:L94-96` — AC-SIV-009 offers three mutants and states "At minimum one
of (a)/(b)/(c) is planted". Branch **(a) "an `Add` path that mints a duplicate id"** never reaches
the invariant block: `backlog_sqlite.go:109` declares `id TEXT NOT NULL UNIQUE`, so the duplicating
mutation aborts inside `Mutate` and `Add` returns `ErrBacklogIDConflict`
(`backlog_sqlite.go:160,171`; the behaviour is already pinned by `TestDuplicateIDRejectedByStorage`,
`backlog_concurrency_test.go:L101-131`). Under the post-change design that error does **not** satisfy
`IsBoardLockHeld`, so it trips the REQ-SIV-002 hard-failure gate and `t.Fatalf`s **before** the
invariant assertions run. The test goes RED and the failure "names an invariant" only by coincidence
of wording — it proves the hard-failure gate fires, and proves nothing whatever about the invariant
block. An implementer picking (a) discharges the card's binding AC with evidence that establishes
nothing about the property under suspicion. This is precisely the "mutant that manufactures the
appearance of discrimination" hazard.
— Severity: **critical** — Class: **blocking**
— Required fix: delete branch (a), or respecify it as a mutant that bypasses storage rejection.
Add to AC-SIV-009: "The RED must be produced **by a named assertion inside the invariant block**;
the recorded output must cite that assertion's message and its source line. A RED produced by the
hard-failure gate, by the zero-progress floor, or by a storage-layer rejection does NOT discharge
this AC."

### D2 — no attempt-conservation assertion; the progress floor admits 1 success in 48
`.moai/specs/.../spec.md:L116-117` (REQ-SIV-007) and `plan.md:L79`, `L94-95` — M3 step 3 deletes
`len(issued) != wantTotal` (`backlog_concurrency_test.go:L65-67`), which today is the only thing tying
the invariant set to the work the test actually attempted. Nothing replaces it. The only remaining
floor is `len(issued) == 0`. After the change, a run in which 47 of 48 adds are starved satisfies every
requirement in §B: the four invariants are checked against a one-element issued set and pass, the
floor passes, and the test reports green. t370 measured actual starvation at **3–7 of 48**
(`.moai/reports/t370/verdict.md:L49-51,L87-90`), so a `> 0` floor is roughly 40× weaker than observed
behaviour and would not notice a total collapse of throughput. Worse, a future accounting bug that
silently drops successes (an `err == nil` branch that fails to record) leaves every invariant
self-consistent against the smaller set and green. That is the "deletion wearing a repair's clothes"
failure relocated from the invariants to the accounting.
— Severity: **critical** — Class: **blocking**
— Required fix: add a REQ (machine-independent, so it does not reintroduce the REQ-SIV-009 hazard):
"The stress test shall assert `successes + starved + hardFailures == stressWriters *
stressAddsPerWriter` — every attempted add is accounted for in exactly one class." Add the matching
AC. Do **not** convert the floor into a fractional success threshold; that would be a load sensor
again — instead state the `> 0` floor's weakness explicitly as a residual risk in §D.

### D3 — the `last_seq` mutant branch is direction-ambiguous, and one direction cannot fire
`.moai/specs/.../acceptance.md:L96` — branch (c) reads "a `last_seq` advance that does not match the
item count". Its downward direction is erased before the assertion sees it:
`normalizeBacklogRecord` (`backlog_store.go` → `backlog_store.go:L770-772`, "the mark must clear every
id the record HOLDS") raises `LastSeq` to the maximum present id, and it runs post-mutate inside
`Mutate`. A mutant that lowers `last_seq` is silently repaired and `TestConcurrencyStress` stays
GREEN — an implementer would then read the non-firing mutant as evidence of something. Only an
upward mutation fires.
— Severity: **major** — Class: **blocking**
— Required fix: restate (c) as "a `last_seq` advance **above** the item count", and add the note
that a downward mutation is neutralized by `normalizeBacklogRecord` and does not discharge this AC.

### D4 — declared Tier S breaches the [HARD] REQ/AC budget by 2.1× and 1.6×
`.moai/specs/.../spec.md:L14` declares `tier: S`. `spec-workflow.md:L146-150` caps Tier S at **8
requirements and 8 acceptance criteria**; this SPEC carries **17 REQ and 13 AC** (counted:
`grep -o 'REQ-SIV-[0-9]*' | sort -u` = 17, `AC-SIV-001..013` = 13). The rule states exceeding the
ceiling "is a signal to tier up or to split the SPEC, not to relax the budget". Two consequences are
not cosmetic: the Tier S plan-auditor threshold (0.75) is the most lenient available, and
`harness.yaml:76` caps Tier S at **one** audit iteration — so a misdeclared tier both lowers the bar
and removes the revision loop for a SPEC whose central hazard is concealment. `progress.md:L8-9`
acknowledges the artifact-set deviation but is silent on the budget.
— Severity: **major** — Class: **blocking**
— Required fix: either (i) set `tier: M` and fold requirements to ≤16 (natural merges: REQ-SIV-009
into REQ-SIV-008; REQ-SIV-015 into REQ-SIV-014; REQ-SIV-003's mechanism clause into REQ-SIV-002),
accepting the 0.80 threshold and the 2-iteration ceiling; or (ii) state explicitly in `spec.md` why
the ceiling is exceeded and which tier's threshold and iteration ceiling the operator intends to
apply. Silence is not an option — the tier field is read mechanically.

### D5 — REQ-SIV-008 asserts sufficiency; §E and the host file deny it
`.moai/specs/.../spec.md:L121-123` — "A dedicated budget guard shall assert that `boardLockWaitBudget`
**covers** the mutation count the stress test actually serializes". On the tree this ships against
that reading is empirically false, by the SPEC's own cited evidence: t370 back-derived the CI `-race`
per-mutation cost at **42–105 ms** (`verdict.md:L87-90`), so the wait actually required is
`48 × 42..105ms = 2.0..5.0 s` against a 1.65 s budget. The guard passes only at the **declared**
33 ms. `spec.md:L196-197` says the opposite in plain words ("asserts the derivation is *coherent*,
not that the budget is *large*"), and the host file's own header says
"NOT that any budget is sufficient" (`board_lock_wait_test.go:L7-8`). A run-phase implementer reading
REQ-SIV-008 alone will write a failure/success message that claims coverage, and a CI reader will
then believe a false thing about a machine that has never satisfied it.
— Severity: **major** — Class: **blocking**
— Required fix: reword REQ-SIV-008 to "shall assert the budget is **coherent with** the mutation
count the stress test serializes **at the declared per-mutation cost `boardLockCIMutationCost`**",
and require the guard's own message to state that this is a coherence relation at a declared cost,
never evidence the budget suffices on any real machine.

### D6 — the classification predicate is mechanical, but its sentinel is broader than "contention"
`internal/kanban/board_lock_unix.go:L41-43` — `acquireBoardLockImpl` maps **every** `unix.Flock`
error to a bare `ErrBoardLockHeld`, not only `EWOULDBLOCK`. `ENOLCK` (no locks available — reachable
on some filesystems), `EINTR`, and `EBADF` are therefore indistinguishable from contention at the
predicate. Today they fail the test; after this change they are silently counted as starvation and
tolerated. The Windows substrate is correctly narrow by contrast (`board_lock_windows.go:L67-77`
sentinels only `os.IsExist`), so the SPEC's tolerance is wider on exactly the platform the CI
`Race Test` job runs. Separately, `Mutate` returns `errors.Join(mutErr, relErr)`
(`backlog_store.go:L681`) and `IsBoardLockHeld` is `errors.Is`, which traverses joins — a future
joined error carrying a lock-held branch alongside a real defect would be tolerated wholesale.
REQ-SIV-003 (`spec.md:L90-92`) claims the classification is mechanical, which is true, and says
nothing about the sentinel's width, which is the thing that decides what gets swallowed.
— Severity: **major** — Class: **blocking**
— Required fix: add a §D non-claim: "The tolerated class is whatever `board_lock_unix.go` maps to
`ErrBoardLockHeld`, which on Unix is every `flock(2)` failure, not only `EWOULDBLOCK`. This SPEC
does not narrow it and cannot distinguish `ENOLCK`/`EINTR` from contention." Add a residual-risk line
covering `errors.Is` join traversal. Optionally raise a follow-up card to narrow
`board_lock_unix.go:L41-43`; narrowing it here is out of scope (`spec.md:L188`).

### D7 — AC-SIV-001 and AC-SIV-005 cannot be discharged under the SPEC's own exclusions
`.moai/specs/.../acceptance.md:L30-35` and `L61-64` — both Givens require a run in which adds are
starved (AC-SIV-005: *every* add starved). `spec.md:L200-201` forbids generating load, running CI, or
reproducing the failure locally, and §D.5 (`acceptance.md:L158`) demands each of AC-SIV-001..007 be
discharged "with cited command + verbatim output". On an unloaded machine with a 1.65 s budget the
Given is never satisfied, so AC-SIV-001 passes vacuously — the exact vacuous-green shape this card
exists to eliminate. Neither AC names a verification verb, unlike AC-SIV-002/004/007 which do.
— Severity: **major** — Class: **blocking**
— Required fix: give both a deterministic construction. The pattern already exists in the tree:
`TestBacklogLockStuckHolderSurfacesBoundedNamedError` (`board_lock_wait_test.go:L107-119`) seeds a
holder with `acquireBoardLockImpl` and releases it via `t.Cleanup` — no load, no background process,
no CI. Add a sub-test that holds the lock for the whole run (drives AC-SIV-005's total starvation)
and one that holds it for part of it (drives AC-SIV-001), or reclassify both ACs as source-reading
criteria with an explicit grep verb and drop them from the "cited command + verbatim output" list.

### D8 — the closure gate cannot discriminate the hazard it is named for
`.moai/specs/.../acceptance.md:L142-152` — AC-SIV-013 requires `TestConcurrencyStress` green on ≥5
post-landing develop heads. But t370's fact 1 (`spec.md:L39-43`) is that the invariants were red in
**0 of 14** runs before the change. A window of 5 greens under the new criterion is therefore fully
consistent with the invariants having been switched off, and confirms only that no *new* failure
mode was introduced. The SPEC presents it as the closure judgment ("Judgment requires a firing rate
across multiple runs") without stating that the pre-change invariant firing rate was already zero.
The gate that actually discriminates is the AC-SIV-008/009 mutant pair — which is present, and is
what keeps this a stated-limit defect rather than a structural one.
— Severity: **minor** — Class: **blocking** (it is a stated-limit gap on the SPEC's own honesty axis)
— Required fix: add to AC-SIV-013's rationale: "The invariant criterion was red in 0 of the 14
observed runs, so a green window evidences only that no new failure mode was introduced. It does not
evidence that the invariants still fire; that is AC-SIV-009's sole burden."

### D9 — one of the two cited "green `-race` runs" is recorded by t370 as a red job
`.moai/specs/.../spec.md:L172-173` — REQ-SIV-016 states "Two green `-race` runs already exist
post-repair (`51daada00`, `c6aa61346`)". `.moai/reports/t370/verdict.md:L63` records
`c6aa61346` as "잡은 붉지만 원인은 다른 테스트" — the job was **red**; only `TestConcurrencyStress`
was green in it. Calling it a green run overstates the cited source in a clause whose whole purpose
is citation honesty.
— Severity: **minor** — Class: **blocking** (it is a misstatement inside the non-claims clause)
— Required fix: rewrite as "one green `Race Test` job (`51daada00`) and one run in which
`TestConcurrencyStress` was green inside a job reddened by another test (`c6aa61346`)".

### D10 — two GEARS pattern labels do not match the sentence written
`.moai/specs/.../spec.md:L96` — REQ-SIV-004 is labelled `(Ubiquitous)` but is written as a
Where-conditional ("**Where** some adds are tolerated as starved, …"). `spec.md:L86` — REQ-SIV-002 is
labelled `(Event-detected, unwanted)` and carries two obligations in one requirement (an
event-driven `shall fail` and an unwanted `shall not tolerate`), which makes its AC coverage
non-atomic.
— Severity: **minor** — Class: **optional**
— Required fix: relabel REQ-SIV-004 `(Where)`; split REQ-SIV-002's second sentence into its own
unwanted requirement, or fold it into REQ-SIV-003 (which already owns the mechanism).

### D11 — AC-SIV-012 traces to a section, not to a requirement
`.moai/specs/.../acceptance.md:L23` — the matrix row reads "scope discipline (spec.md §E)". Every
other AC names a REQ. Scope discipline has no REQ-SIV number, so it is invisible to any
requirement-level traceability check.
— Severity: **minor** — Class: **optional**
— Required fix: add a REQ-SIV for the scope constraint (or fold it under REQ-SIV-016's
report-discipline umbrella) and point AC-SIV-012 at it.

### D12 — the new guard's independent catch-set is narrower than §B claims, at a 4% margin
`plan.md:L20-21` calls AC-SIV-008 "the only thing establishing the guard is not vacuous". The
arithmetic checks out — floor `48 × 33ms = 1.584 s`, mutated budget `10 × 33ms × 4 = 1.32 s < 1.584 s`
→ RED; unmutated `1.65 s ≥ 1.584 s` → GREEN — but the margin is 66 ms, **4.2%**, and the existing
`TestBoardLockWaitBudgetDerivedFromNamedInputs` already pins `boardLockSupportedWriters == 10`
(L45), `boardLockCIMutationCost >= 33ms` (L54) and `boardLockHeadroom >= 2` (L59). The new guard's
*independent* catch-set is therefore only `boardLockHeadroom ∈ {2,3,4}` plus increases to
`stressWriters`/`stressAddsPerWriter`. That is a real regression tripwire and worth having; it is
not the broad guarantee §B.4's framing implies.
— Severity: **minor** — Class: **optional**
— Required fix: state the catch-set and the 4.2% margin in `plan.md` §B, so nobody later reads the
guard as coverage evidence.

### D13 — AC-SIV-008 does not require the discriminating half of its own observation
`.moai/specs/.../acceptance.md:L85-91` — the command `go test -run 'TestBoardLockWaitBudget'` matches
**both** guards. The AC requires only that "the new budget guard FAILS". The evidence that the new
guard is the discriminator is that the **old** guard stays GREEN under the same mutant (it does:
budget is a derived const, so `budget == recomputed` still holds and `headroom 4 >= 2` passes).
— Severity: **minor** — Class: **optional**
— Required fix: add "and `TestBoardLockWaitBudgetDerivedFromNamedInputs` remains GREEN under the same
mutant — recorded verbatim" to AC-SIV-008.

---

## Answers to the six directed questions

1. **Mutant pair discriminating?** Half. AC-SIV-008's arithmetic is correct and the mutant fires
   (D12 bounds what it proves). AC-SIV-009 is **not** discriminating as written — branch (a) cannot
   reach the invariant block (D1) and branch (c)'s stated direction is ambiguous with one direction
   neutralized (D3). Only branch (b) reliably fires.
2. **Tolerance leaking into invariants?** The four assertions are correctly re-anchored to the issued
   set and REQ-SIV-006 forbids conditionalising them — that half is sound. The leak is one layer
   down, in the accounting: nothing asserts that attempts are conserved, and the `> 0` floor admits
   1 success in 48 (D2).
3. **Classification mechanical?** Yes — `IsBoardLockHeld` is `errors.Is(err, ErrBoardLockHeld)`
   (`board_lock.go:L28-30`), no string matching anywhere, and the budget-exhaustion error does wrap
   the sentinel (`backlog_store.go:L738-740`). But the sentinel is over-broad at its source on Unix,
   so tolerance is wider than contention (D6).
4. **Guard non-vacuous and machine-independent?** Machine-independent — yes, constants only, and
   AC-SIV-007 forbids the clock explicitly. Non-vacuous — yes, but narrowly (D12), and its
   `covers` framing asserts a sufficiency the evidence contradicts (D5).
5. **Non-claims honest and complete?** Both required limits are stated (REQ-SIV-016) and enforced in
   reports (AC-SIV-011) — this is the strongest part of the SPEC. Two gaps: one cited run is
   misdescribed (D9), and the closure gate's discriminating power is overstated by omission (D8).
6. **Scope discipline?** Clean. Branch A and branch B are both explicitly excluded with rationale
   (`spec.md:L184-197`), AC-SIV-012 pins the diff to three files and asserts the constants unchanged,
   and `plan.md` §G names "re-tuning a constant to make something pass" as an anti-pattern with the
   correct response (report a blocker). **No requirement drifts into A or B.**

---

## Recommendation

FAIL at 0.69 against the Tier S threshold of 0.75. Route the seven blocking defects (D1–D7) and the
two stated-limit defects (D8, D9) back to `manager-spec`; D10–D13 are optional and left to the
orchestrator's discretion.

Fix order, by how much each changes the artifacts downstream:

1. **D1 + D3** — respecify AC-SIV-009's mutant menu and require the RED to originate inside the
   invariant block. This is the card's binding condition; everything else is secondary to it.
2. **D2** — add the attempt-conservation requirement and its AC.
3. **D5** — reword REQ-SIV-008 from "covers" to "coherent with, at the declared cost".
4. **D6** — add the sentinel-width non-claim to §D.
5. **D7** — give AC-SIV-001/005 a deterministic seeded-holder construction, or reclassify them.
6. **D4** — resolve the tier declaration (this also decides whether a second audit iteration is
   available; under the declared Tier S it is not).
7. **D8 + D9** — repair the two honesty clauses.

Note for the orchestrator: under the declared `tier: S` the plan-audit ceiling is **1 iteration**
(`harness.yaml:76`), so a revision cannot be re-audited without first resolving D4.

---

**Gaps (not observed in this audit):** no test was executed and no CI was read — every code claim
above is a source-reading claim citing file and line on tree `9d4f79281`. The t370 figures are
consumed as cited ground truth and were not re-derived. Whether
`TestBoardLockWaitBudgetDerivedFromNamedInputs` actually passes today was not measured, only derived
from its source.

**Residual risk:** D6's `ENOLCK`/`EINTR` reachability is argued from the source's error mapping, not
from an observed occurrence; it may be unreachable in practice on the CI filesystem. D12's
catch-set enumeration assumes the existing guard's assertions stay as they are today.

---
---

# SPEC Review Report — Iteration 2: SPEC-STRESS-INVARIANT-VERDICT-001 (card t372)

Iteration: **2/2** (Tier M ceiling — `.moai/config/sections/harness.yaml:77` `plan_audit_tier_ceilings: M: 2`)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.83** (harmonic mean of the four dimensions)
Tier threshold applied: **0.80** (Tier M, `spec-workflow.md:141`; declared `spec.md:14`)
Score movement: **0.69 → 0.83, monotone increase.** No STOP escalation (the LEAN score-regression
clause does not fire).

Reasoning context ignored per M1 Context Isolation. Audit tree `.claude/worktrees/t372`, HEAD
`93ced5ca5`, branch `WT-stress-invariant-guard`. Artifacts re-read at this HEAD, not carried from
iteration 1. `progress.md`'s defect-response table was **not** taken as evidence — every claim below
was re-verified against the SPEC text or the Go source it cites. t370's measurements are consumed as
given ground truth; no test was run, no CI read, no load generated.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-SIV-[0-9]*' spec.md | sort -u` = exactly
  `REQ-SIV-001 … REQ-SIV-016`, 16 entries, sequential, no gap, no duplicate, uniform 3-digit padding.
  16 bold declarations at `spec.md:109,113,116,123,127,137,143,146,165,177,200,211,219,239,274,282`.
- **[PASS] MP-2 GEARS format compliance — judged against the REQUIREMENT layer (`spec.md` §B–§E
  `REQ-SIV-*`), not the verification layer.** All 16 match a canonical GEARS pattern: event-driven
  `When … shall` (001, 002, 007), ubiquitous `The … shall` (005, 008, 009, 010, 011, 012, 013, 014,
  015), unwanted `shall not` (003, 006, 016), Where-conditional (004). The `AC-SIV-*` entries in
  `acceptance.md` are Given-When-Then — the correct verification-layer format — and were graded under
  Group 4 only. Nit, not a failure: the label string is `(Event-detected)` where GEARS says
  *Event-driven*; the sentence shape is canonical, so MP-2 binds on the sentence and passes.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-13`): `id`, `title`, `version: "0.2.0"` (quoted), `status: draft` (valid enum),
  `created`/`updated` ISO `2026-08-31`, `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` comma-separated. No rejected snake_case alias
  (`created_at`/`updated_at`/`labels`/`spec_id`). `tier: M` and `depends_on` are additive optionals.
- **[N/A → PASS] MP-4 language neutrality** — single-language SPEC (Go, `module: internal/kanban`).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference,
  `SPEC-BACKLOG-LOCK-BUDGET-001`; `.moai/specs/SPEC-BACKLOG-LOCK-BUDGET-001/spec.md:5` reads
  `status: implemented` — not in {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` = `0`, auto-PASS per
  D8-4. Recorded honestly: the auto-PASS is a literal-substring result, not an affirmative
  cross-platform review. Independently, the SPEC *does* address the platform split on its own
  initiative (`spec.md:258-262`, REQ-SIV-014 clause 4, Unix `flock` vs Windows `os.IsExist`).
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-STRESS-INVARIANT-VERDICT-001/`
  → rc=1, no match, across `spec.md`/`plan.md`/`acceptance.md`/`progress.md`. `research.md` absent
  (Tier M does not require it).

**No must-pass failure.**

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Five iteration-1 clarity defects closed with source-cited exclusions (`acceptance.md:156-167` records why the duplicate-id and downward-`last_seq` mutants cannot discriminate). Residual: two load-bearing terms are still undefined — "the invariant block" (N1) and `successes` (N2) — and REQ-SIV-009's "at the declared cost" framing implies a cost-dependence its own §B.4 algebra denies (N3). |
| Completeness | 0.90 | 0.75–1.0 | All sections present: HISTORY `L20-25`, §A WHY `L27`, §B–§E scope, 16 requirements, 14 ACs in `acceptance.md`, six `### Out of Scope — <topic>` H3s with bullets (`L291,301,308,314,323,331`). Every iteration-1 gap filled: conservation REQ+AC, sentinel non-claim, closure-gate limit, tier justification, `>0`-floor residual risk, catch-set table. Docked for two placement gaps: the cost-independence limit lives in REQ-SIV-010/`plan.md` §B but not in REQ-SIV-009 where the message wording is mandated; AC-SIV-012's mutant carve-out names AC-SIV-008 only (N4). |
| Testability | 0.72 | 0.75, docked | AC-SIV-001/005 now carry a deterministic construction with a named verb (`acceptance.md:34-46,73-81`) — the iteration-1 vacuous-Given defect is gone. AC-SIV-008 requires the old guard's GREEN plus a non-zero `-run` selector match (`L115-119`). AC-SIV-014 is source-readable plus exercised. Docked because the **binding** AC-SIV-009 turns on an undefined boundary (N1) and AC-SIV-014's identity is not binary until `successes` is defined (N2). |
| Traceability | 1.00 | 1.0 | Matrix `acceptance.md:12-25` verified row-by-row against the 16 REQ ids: REQ-SIV-001..016 each covered by ≥1 AC; every AC names a real REQ. The iteration-1 orphan is gone — AC-SIV-012 now traces to REQ-SIV-016 (`L24`), which exists at `spec.md:282`. No uncovered REQ, no orphaned AC. |

Harmonic mean: `4 / (1/0.75 + 1/0.90 + 1/0.72 + 1/1.00) = 4 / 4.8333 = 0.8276` → **0.83 ≥ 0.80**.

---

## Directed-verification findings (fix claims checked against source, not accepted)

**1. AC-SIV-009's remaining branches reach the invariant block; the exclusion list is one short.**
Branch (c) upward fires: `normalizeBacklogRecord` (`internal/kanban/backlog_store.go:751`, called
post-mutate at `:663`) raises `LastSeq` only when `max > rec.LastSeq` (`:772-774`), so an upward
mutation survives to the `rec.LastSeq != issuedCount` assertion. Branch (b) fires on invariants (b)
and (c). Both verified reachable. The four enumerated non-qualifying sources are **not exhaustive**:
the `store.Load()` error path (`backlog_concurrency_test.go:72-75`, `t.Fatalf("Load after stress:
%v")`) is a fifth — it is a named assertion, with a message and a source line, sitting textually
between the collision assertions and the presence assertions, and it is not one of the four
REQ-SIV-005 invariants. Because "the invariant block" is nowhere bounded, this RED could be cited as
discharging the card's binding AC. See N1.

**2. Conservation is machine-independent and closes the hole iteration 1 named — but not the one it
does not claim to.** `successes + starved + hardFailures == stressWriters * stressAddsPerWriter`
(`spec.md:148`) counts outcomes only; AC-SIV-014 (`acceptance.md:219-220`) forbids any wall-clock or
fractional term. It closes the **silent-drop** hole (an `err == nil` branch failing to record breaks
the identity while every invariant stays self-consistent). It does **not** close the 1-success-in-48
hole — a run of 1 success + 47 starved satisfies the identity, the floor, and all four invariants —
and the SPEC states exactly that, in the place iteration 1 required (`spec.md:157-161` and the fifth
residual risk at `spec.md:269-272`). That is the honest outcome, not an evasion. Bucket exhaustiveness
holds under a dedicated counter; it does **not** hold if `successes` is read as `len(issued)`. See N2.

**3. The D7 seeded-holder construction is deterministic and matches the in-tree pattern.** Verified
against `board_lock_wait_test.go:107-119`: `acquireBoardLockImpl(store.LockPath())` +
`t.Cleanup(held.release)`. Determinism is structural — `flock(LOCK_EX|LOCK_NB)` binds to the open
file description, so a second `unix.Open` in the same process conflicts
(`board_lock_unix.go:37-44`), which the existing test already relies on by asserting the mutation
fails. No spawned process, no `go func` outliving the test, no load. The bounded wait the author
claims is accurate: `acquireLock` polls to `boardLockWaitBudget` (`backlog_store.go:730-742`) then
returns the wrapped sentinel, so 1–2 adds pay 1.65–3.3 s once or twice, never 48 times.

**4. Cost-independence — stated, but not in the place that authors the CI-visible message.** I
re-derived the cancellation independently and it holds: `boardLockWaitBudget =
supportedWriters × cost × headroom`, floor `= stressWriters × stressAddsPerWriter × cost`, so the
inequality reduces to `10 × 5 = 50 >= 8 × 6 = 48`, independent of `boardLockCIMutationCost`. The
SPEC states this at `spec.md:193-198` and `plan.md:32-50` (with the "unaffected — it cancels" row and
the 4.2 % margin). **No requirement or AC claims the guard protects the budget's adequacy** — D5's
`covers` is genuinely gone. The residual overclaim is narrower and lives in the message contract:
REQ-SIV-009 (`spec.md:170-171`) and `plan.md:115-119` mandate a CI-visible string saying the relation
holds "at the declared 33 ms per mutation", which implies a cost-conditioning the guard does not
have. See N3.

**5. REQ-SIV-010's vacuity observation is complete on both halves.** The unreachability is verified:
`board_lock_wait_test.go:26-33` `Fatalf`s unless `budget == recomputed`, and `:37-38` rebuilds
`floor` from the identical expression, so `:39` `budget < floor` cannot fire for any inputs. The SPEC
states the shape must not be reproduced (`spec.md:186-191`, verified by AC-SIV-007
`acceptance.md:100-104`) **and** that the new guard's independently-sourced floor is what replaces it
(`plan.md:64-66`). Repairing the old branch is explicitly out of scope (`spec.md:308-312`).

**6. Tier M reclassification is legitimate, and the binding pair survived intact.** Counts verified
mechanically: 16 REQ, 14 AC, against Tier M's 16/16 (`spec-workflow.md:146-150`). Adversarial check —
tiering up is **not** a lenience purchase: it raises the threshold 0.75 → 0.80 while granting the
second iteration, and `spec-workflow.md:152` names tiering up as the rule's own response to a budget
breach. The 17 → 16 consolidation loses no subject area (classification 3, invariants 3, accounting
2, guard 3, observability 1, mutant 1, non-claims 2, scope 1 = 16, with two genuinely new
requirements added). **AC-SIV-008 and AC-SIV-009 remain separate, both marked binding**
(`acceptance.md:20-21,110,131`); neither was merged, weakened, or diluted — AC-SIV-008 gained the old
guard's GREEN and the selector-match clause, AC-SIV-009 gained the positive origin requirement.

**7. D6 boundary is correct in both directions.** The non-claim is present and accurate: verified
`board_lock_unix.go:41-43` maps every `unix.Flock` error to a bare `ErrBoardLockHeld`, and
`joinBacklogReleaseErr` returns `errors.Join` (`backlog_store.go:681`) which `errors.Is` traverses —
both stated at `spec.md:258-267`. **No scope drift**: narrowing the sentinel is explicitly listed
out of scope (`spec.md:296-297`) and recorded as a follow-up card. Listing `EBADF` alongside `ENOLCK`
and `EINTR` slightly over-states reachability (the fd came from a successful `open`), but a non-claim
erring wide is the conservative direction and is not a defect.

**8. Both honesty clauses check out against the cited source.** REQ-SIV-014 clause 1 ("no
before/after comparison exists, in any quantity") is unchanged and correct. Clause 2's correction is
verified against t370: `measurements.md:211` records `51daada00` job `success`; `:206` and `:222`
record `c6aa61346` as job `failure` with `TestConcurrencyStress` 통과, reddened by
`TestGitDiffNameCount_Predicate`. `spec.md:246-250` and `acceptance.md:191-193` now describe both
runs exactly that way. Clause 3 (the 0-of-14 limit) appears in both the requirement (`spec.md:251-256`)
and the closure gate's own rationale (`acceptance.md:241-245`). No requirement or AC claims more than
t370 measured.

---

## Iteration-1 defect discharge table

| # | Iteration-1 defect | Status | Basis (verified at `93ced5ca5`) |
|---|---|---|---|
| D1 | AC-SIV-009 dischargeable by a non-discriminating mutant | **Discharged** (residual N1) | Branch (a) deleted; exclusion recorded with source evidence (`acceptance.md:156-162`, verified `backlog_sqlite.go:109` UNIQUE + `:162` `ErrBacklogIDConflict`). Positive origin requirement added (`L138-139`). Remaining branches verified reachable. |
| D2 | no attempt conservation | **Discharged as specified** (residual N2) | REQ-SIV-008 `spec.md:146-148` + AC-SIV-014 `acceptance.md:214-222`. Machine-independent; fractional floors forbidden; `>0`-floor weakness stated as residual risk — exactly the fix iteration 1 prescribed. |
| D3 | `last_seq` direction ambiguous | **Discharged** | `acceptance.md:134` reads "**above** the item count"; downward exclusion `L163-167` verified against `backlog_store.go:772-774` (raise-only) and `:663` (post-mutate). |
| D4 | Tier S breaches the REQ/AC budget | **Discharged** | `tier: M`; 16/14 verified inside 16/16; justification `spec.md:79-103` names both consequences (threshold 0.80, 2 iterations) and the split-rejection rationale. |
| D5 | REQ asserts budget sufficiency | **Partially discharged** (N3) | `covers` → "coherent with … at the declared per-mutation cost" (`spec.md:165-175`). Sufficiency claim gone; the mandated message still implies a cost-dependence the algebra denies. |
| D6 | over-broad Unix sentinel unstated | **Discharged** | REQ-SIV-014 clause 4 (`spec.md:257-267`); verified accurate; narrowing kept out of scope (`L296-297`) — no drift. |
| D7 | AC-SIV-001/005 undischargeable under §E | **Discharged** | Seeded-holder construction + named verb (`acceptance.md:34-46,73-81`); pattern verified at `board_lock_wait_test.go:107-119`; `spec.md:316-319` records why it is not load generation. |
| D8 | closure gate cannot discriminate | **Discharged** | REQ-SIV-014 clause 3 + AC-SIV-013 rationale (`acceptance.md:241-245`), both carrying the 0-of-14 limit and assigning the burden to AC-SIV-009. |
| D9 | `c6aa61346` miscalled a green run | **Discharged** | `spec.md:246-250` corrected; matches `t370/measurements.md:206,222` verbatim in substance. |
| D10 | GEARS labels vs sentences | **Discharged** | REQ-SIV-004 relabelled `(Where)` (`spec.md:123`); the two-obligation requirement split into REQ-SIV-002 (event) and REQ-SIV-003 (unwanted). |
| D11 | AC-SIV-012 traced to a section | **Discharged** | REQ-SIV-016 added (`spec.md:282`); matrix row `acceptance.md:24` points at it. |
| D12 | guard's catch-set overstated | **Discharged** | `plan.md:27-54` — 4.2 % margin, cost-independent reduction, and the four-row catch-set table. |
| D13 | AC-SIV-008 lacks its discriminating half | **Discharged, strengthened** | `acceptance.md:115-119` requires the old guard's GREEN under the same mutant *and* a non-zero `-v` selector match. |

**11 discharged, 1 discharged-as-specified with a new residual, 1 partial.** None of the thirteen is
relabelling: in every case the artifact text changed in a way I could check against Go source or
against t370, and the two source-backed exclusions (`UNIQUE(id)`, `normalizeBacklogRecord`) are
independently correct.

---

## Defects Found (iteration 2 — new)

**N1 — "the invariant block" is never bounded, and the binding AC turns on that boundary.**
`.moai/specs/SPEC-STRESS-INVARIANT-VERDICT-001/acceptance.md:138` (also `spec.md:228`,
`plan.md:185`) — AC-SIV-009 requires the RED to originate "at a named assertion inside the invariant
block", but no artifact defines where that block begins or ends. The consequence is concrete: the
`store.Load()` failure `t.Fatalf` (`internal/kanban/backlog_concurrency_test.go:72-75`) is a named
assertion with a message and a source line, sits textually inside the invariant region, is not one of
the four REQ-SIV-005 invariants, and is **not** on the four-item non-qualifying list at
`acceptance.md:148-151`. A mutant that corrupts persisted state could redden there and be cited as
discharging the card's binding evidence. — Severity: **major** — Class: **blocking**
— Required fix: define the invariant block as exactly the four assertions of REQ-SIV-005 (a)–(d), and
require the cited assertion to be one of them **by name**; add "a failure of `store.Load()` itself, a
`-race` DATA RACE report, or a panic" to the non-qualifying list at `acceptance.md:148-151`.

**N2 — `successes` is undefined, and the two available readings give the conservation identity
different meanings.** `spec.md:148` / `acceptance.md:217` state
`successes + starved + hardFailures == stressWriters * stressAddsPerWriter`, but `successes` is
defined nowhere; `plan.md:167` introduces `issuedCount := len(issued)` where `issued` is a
`map[string]int` of **counts**, so `len(issued)` is the *distinct-id* count, not the *successful-add*
count. If an implementer sets `successes := len(issued)`, the identity no longer says "every attempt
lands in exactly one class" (a duplicate issuance would be counted zero times), and it silently
becomes a second collision detector competing with REQ-SIV-005(a) for the RED. REQ-SIV-008's own
claim — "accounted for in exactly one class" — is therefore not established as written.
— Severity: **major** — Class: **blocking**
— Required fix: define `successes` in REQ-SIV-008 as a dedicated counter incremented once per `Add`
call returning a nil error — explicitly **not** `len(issued)` — and state in AC-SIV-014 that the
identity is checked against that counter, so the invariant assertions keep `issuedCount` and the
conservation assertion keeps `successes` as separate quantities.

**N3 — the guard's mandated CI-visible message implies a cost-dependence the guard does not have.**
`spec.md:170-171` (REQ-SIV-009) and `plan.md:115-119` require the guard's failure and success
messages to state the relation holds "at the declared `boardLockCIMutationCost`" / "at the declared
33 ms per mutation". By the SPEC's own algebra twenty lines later (`spec.md:195-198`, which I
re-derived and confirm), the cost cancels and the guard enforces
`boardLockSupportedWriters * boardLockHeadroom >= stressWriters * stressAddsPerWriter` — a
cost-**independent** relation. A CI reader given that message would reasonably conclude the guard
would notice a change to `boardLockCIMutationCost`; it would not (`plan.md:49` says so in a table the
message's reader never sees). This is a smaller misdescription than D5's `covers`, but it sits on
exactly the surface this card exists to keep honest. — Severity: **major** — Class: **blocking**
— Required fix: add one sentence to REQ-SIV-009: "the per-mutation cost appears on both sides and
cancels, so this relation is cost-independent; the message shall not imply that a change to
`boardLockCIMutationCost` would be caught here"; amend `plan.md:117-119`'s prescribed wording to say
the relation compares the lock policy's supported-lane budget against the stress test's serialized
mutation count.

**N4 — AC-SIV-012's mutant carve-out names AC-SIV-008 only, while AC-SIV-009's mutant must live in
production storage code.** `acceptance.md:208-212` — the scope AC exempts "a temporary mutant under
AC-SIV-008" and then forbids any change to `backlog_store.go`. AC-SIV-009's sanctioned branches — a
dropped item after issuance, or an upward `last_seq` advance — are produced in
`backlog_store.go`/`backlog_sqlite.go` by construction, so the implementer reads a contradiction
between two MUST criteria. — Severity: **minor** — Class: **blocking**
— Required fix: extend the parenthetical to "a temporary mutant under AC-SIV-008 **or AC-SIV-009** is
reverted before the diff is taken", and note that the post-revert `git diff --stat` is the evidence.

**N5 — AC-SIV-014 is filed under "§D.3 Reporting criteria".** `acceptance.md:176,214` — attempt
conservation is an accounting criterion, not a reporting one; it also breaks the section's
AC-numbering order (010, 011, 012, **014**). Cosmetic, but it makes the binding accounting criterion
easy to skim past in the section a reader reaches for report wording.
— Severity: **minor** — Class: **optional**
— Required fix: move AC-SIV-014 into §D.1 after AC-SIV-005, or give it its own §D.1b.

**N6 — `(Event-detected)` is not a GEARS label.** `spec.md:109,113,143` — the canonical pattern name
is *Event-driven*. The sentences are canonical `When … shall`, so MP-2 is unaffected; the label
string is simply non-standard. — Severity: **minor** — Class: **optional**
— Required fix: relabel the three requirements `(Event-driven)`.

---

## Regression Check (iteration-1 defects)

Covered in full by the discharge table above. No iteration-1 defect reappears unchanged, so the
stagnation clause does not fire. D5 is the only one carried forward in altered form (as N3), and its
severity is reduced: an unsupportable *sufficiency* claim became a misleading *conditioning* claim.

---

## Recommendation

**PASS-WITH-DEBT at 0.83 against the Tier M threshold of 0.80.** No must-pass criterion fails, the
score cleared the bar monotonically, and — the test that matters on this card — the repairs are
repairs: each of the thirteen was re-verified against Go source or against t370's measurements rather
than against the author's own summary, and the two load-bearing exclusions the fix round introduced
(`UNIQUE(id)`, `normalizeBacklogRecord`'s raise-only mark) are independently correct. Nothing was
concealed by wording; the one place where wording outran the evidence (N3) understates rather than
overstates the hazard, and the SPEC contradicts itself into honesty twenty lines later rather than
away from it.

The debt is four blocking findings (N1–N4). **This is iteration 2 of 2**, so they cannot be routed
through a third audit; per the retry-loop contract they are handed to the orchestrator with the
ceiling reached. Two of them (N1, N2) bear directly on whether the card's binding evidence can be
discharged honestly, so they should be closed **before** run-phase entry rather than during it:

1. **N1** — bound "the invariant block" to REQ-SIV-005 (a)–(d) and extend the non-qualifying list.
   Without it, AC-SIV-009 — the single criterion separating "the verdict moved" from "the rule was
   switched off" — can be discharged by a `Load` failure.
2. **N2** — define `successes` as a nil-error `Add` counter, distinct from `len(issued)`.
3. **N3** — put the cost-independence sentence inside REQ-SIV-009 and fix `plan.md`'s prescribed
   message wording.
4. **N4** — extend AC-SIV-012's mutant carve-out to AC-SIV-009.
5. N5, N6 are optional and left to the orchestrator's discretion.

All four fixes are single-sentence edits to `spec.md` / `acceptance.md` / `plan.md`. None requires
re-deriving a measurement, re-opening the tier decision, or touching the requirement count (16/14
stays inside 16/16).

---

**Gaps (explicitly not observed in this audit):** no test was executed, no CI was read, no load was
generated, and no binary was built. Every code claim above is a source-reading claim against tree
`93ced5ca5`, citing file and line. Whether `TestBoardLockWaitBudgetDerivedFromNamedInputs` or
`TestConcurrencyStress` actually passes today was not measured. The t370 figures were consumed as
cited ground truth and not re-derived. The iteration-1 report's own line citations were not re-checked
against tree `9d4f79281`; where iteration 2 cites a line it is a fresh read at `93ced5ca5`.

**Residual risk:** N1's `Load`-failure path is argued from the current test's structure — the
post-change test may relocate or remove that `t.Fatalf`, in which case the fifth source is different
rather than absent, and the undefined boundary remains the actual defect. N2 assumes the implementer
has a free choice between the two readings; a run-phase author who reads `plan.md` §F M3 in order may
land on the correct one by accident, which would leave the SPEC's text wrong but the code right. The
cost-independence algebra in finding 4 was re-derived by hand from the constant definitions at
`board_store.go:96-117` and not machine-checked.
