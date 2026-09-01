# SPEC Review Report: SPEC-INTEGRATION-LOCK-ATOMIC-001 (card t336) — iteration 2

Iteration: 2/2 (Tier M ceiling)
Verdict: **PASS**
Overall Score: **0.936** (Tier M PASS threshold = 0.80)
Movement: **0.803 → 0.936, monotonic increase (+0.133).** No dimension regressed; no STOP signal.

Tree under audit: `15453140a`, branch `WT-integration-lock-atomic`. Artifacts at v0.1.1 (all four).
Reasoning context ignored per M1 Context Isolation. Scope: the D1-D9 delta plus a regression check
over the iteration-1 "already sound" list, per the Retry Loop Contract's delta-scoped re-audit.

Author's change summary was **not accepted as evidence**; every claim below was re-derived from the
files and the tree.

---

## Must-Pass Results (re-verified, not carried forward)

- **[PASS] MP-1** — `grep -o 'REQ-ILA-[0-9]*' spec.md | sort -u` → REQ-ILA-001..011, sequential, no
  gap, no duplicate. AC-ILA-001..009 likewise. 11 REQ / 9 AC, both within the Tier M ceiling of 16.
- **[PASS] MP-2** — requirement layer re-checked after the REQ-ILA-005 amendment and the
  REQ-ILA-007 extension. Both keep their GEARS form: 005 remains Ubiquitous ("shall be unchanged")
  with the amendment as a scoped carve-out rather than a second requirement; 007 remains Ubiquitous
  across all three of its now-clauses ("shall be", "shall remain", "shall NOT share", "shall remain
  forbidden"). All 11 still match a pattern.
- **[PASS] MP-3** — 12 canonical fields present and correctly typed in all four artifacts;
  `version: "0.1.1"` quoted semver across all four (verified by reading each frontmatter block). No
  rejected snake_case alias.
- **[N/A] MP-4** — single-language SPEC. Auto-passes.
- **[PASS] MP-5 (D7)** — same two references; `SPEC-INTEGRATION-LOCK-LIVENESS-001` and
  `SPEC-KANBAN-BOARD-001` both exist and both read `status: completed`. No BLOCKING.
- **[PASS] MP-6 (D8)** — no `syscall` literal. Build-tag discipline is if anything stronger at
  iter-2 (plan.md M3 now states explicitly that no darwin command compiles the windows-tagged code).
- **[PASS] MP-7** — `[NEEDS CLARIFICATION]` appears only in the two negative assertions
  (`plan.md:104`, `progress.md:30`). No unresolved marker.

Supporting evidence consumed as given (not re-run per instruction): scoped lint at
`.moai/reports/t336/spec-lint-iter2.txt` → `✓ No findings — all SPEC documents are valid`.

---

## Category Scores

| Dimension | iter-1 | iter-2 | Movement | Evidence |
|-----------|--------|--------|----------|----------|
| Clarity | 0.75 | **0.90** | +0.15 | D1 is now a `[HARD]` block naming the env var, the marker, and the forbidden alternative; "One test, two tree states" (plan.md M1) removes the D4 ambiguity outright. Docked for N1 (hook nil-guard unstated) and N2 (distinct session id stated only as prose setup). |
| Completeness | 0.90 | **0.95** | +0.05 | 3-attempt ceiling with a named owner and action; closure gate 6 added; §G risk 3 rewritten; §H history row records the repair. Docked for N3 (stall-release timeout unquantified on one side of a stated ordering). |
| Testability | 0.65 | **0.90** | +0.25 | The largest movement, and it is earned: the criterion no longer waits for an interleaving, it constructs one. Three-part attributed observable; `busy` and `REPLACED=<session>` both explicitly FAIL; pgrep self-exclusion; one shipped test function. Docked for N2 and N3. |
| Traceability | 1.00 | **1.00** | — | §D.3 matrix and §D.4 indirect-verification statement unchanged verbatim; all 11 REQs still covered; the new closure gate 6 adds a check without disturbing the mapping. |

Harmonic mean = 4 / (1/0.90 + 1/0.95 + 1/0.90 + 1/1.00) = **0.936**.

**Monotonicity: satisfied.** Every dimension moved up or held. The score-regression STOP clause does
not fire.

---

## D1-D9 disposition — verified against the files

| # | iter-1 defect | Status | Verification |
|---|---|---|---|
| D1 | Helper PID unspecified | **CLOSED** | `plan.md` M1 carries a `[HARD]` block: `want.PID` from `HELPER_OWNER_PID` (parent pid) with `PIDSource: PIDSourceSessionOwner`, `os.Getpid()` forbidden, with the stale-takeover mechanism spelled out and `PID: 0` named as the acceptable alternative. Parity claim verified: `internal/cli/integration.go:184` is `ownerPID, _ := session.ResolveOwnerPID()` and `:188` sets `PIDSource: kanban.PIDSourceSessionOwner`. Discriminator added to AC-ILA-001 (three-part) and AC-ILA-002 (`REPLACED=<session>` ⇒ FAIL). `plan.md` §G gained both anti-patterns. |
| D2 | RED probabilistic, no stop rule | **CLOSED** | Hook adopted. REQ-ILA-005 amendment (spec.md:143-152); §G risk 1 fully rewritten to "deterministic, not lucky"; 3-attempt ceiling binding **both** AC-ILA-001 and AC-ILA-006 with "No widen-and-retry loop is authorized"; closure gate 6 added. |
| D3 | AC-ILA-001 command unrunnable | **CLOSED** | Tree restated as "HEAD with M2's critical section disabled … NOT at `15453140a`, where neither the test nor the interleaving hook exists" — the exclusion is explicit. Pass polarity now positive: "That positive report is the pass condition; the test's non-zero exit is a by-product, not the signal." |
| D4 | Two artifacts, different deliverables | **CLOSED** | `grep -c 'DoubleHoldObserved'` = **0 in all four artifacts** (verified). `plan.md` M1's "flips" sentence replaced by "One test, two tree states"; M2 now reads "M1's test is unchanged by M2". |
| D5 | Option-(a) argument overstated | **CLOSED** | §B item 1 is now scope correctness, explicitly labelled "this is the deciding argument". Item 2 states the retry fact correctly (`acquireBoardLockSerialized`, `board_store.go:117`, 10 × 33 ms × 5 = 1.65 s, "would WAIT, not fail") and carries an in-text retraction of the earlier claim. Arithmetic re-verified against `board_store.go:96-117`. |
| D6 | §G mitigation could not fire | **CLOSED** | §G risk 3 now says the darwin commands "would be a mitigation that cannot fire", names `//go:build windows` (line 1, verified), and splits compilation (`GOOS=windows go vet`) from behavior (CI windows job). AC-ILA-007(b) names both files; `internal/kanban/board_lock_clear_windows_test.go` exists (verified). |
| D7 | pgrep self-match | **APPLIED** | Bracketed first character plus `$$` filter, with the failure mode stated: "a `1`/`1` reading means the self-match exclusion is not working and the measurement is broken, NOT that the run was clean". |
| D8 | Claim mood | **APPLIED** | Now "shall both be told" / "shall be attributable". |
| D9 | Artifact stem | **APPLIED** | `grep -n 'integration-lock.mutation.lock'` → no match (rc=1); `.moai/state/integration-mutation.lock` at `plan.md:190`. REQ-ILA-007 gained the stem prohibition and the glob prohibition. |

---

## Answers to the six named iter-2 hazards

### 1. The D2 hook — test-only, unreachable in production, and does it weaken GREEN?

**Unreachable by assignment: yes. Closure gate 6 works: yes. GREEN is not weakened — it is
strengthened. One gap: the call site's nil guard is never stated (N1).**

- *Test-only by assignment.* The variable is package-level and unexported, so only `package kanban`
  can set it; `internal/` is fully covered by gate 6's recursive grep. The realistic evasion — an
  exported setter in a non-test file (`func SetHook(f func()) { integrationLockMutationTestHook = f }`)
  — still contains the assignment on a non-`_test.go` line and is caught. The only constructions
  that escape are contrived (taking the variable's address and assigning through the pointer).
- *Gate 6 executes here.* I tested the regex form rather than assuming it:
  `grep -n 'integrationLockMutationTestHook\s*=' <fixture>` matched with rc=0 on this machine, as
  did the POSIX-portable `grep -nE '…[[:space:]]*='`. The `\s` escape is a GNU extension and is not
  guaranteed on a stock BSD `grep`; the gate is executable in this environment, and switching to
  the bracket-expression form would make it environment-independent at zero cost (trivial, folded
  into N1's fix note).
- *GREEN is strengthened, not trivialized.* This was the question worth asking, and the answer is
  the opposite of the worry. With the hook, in the repaired state: A takes the mutation lock, stalls
  at the hook between decision and write; B blocks on the mutation lock; A is released, writes,
  releases; B then enters, **re-reads**, sees A's record, and is refused. That path exercises
  REQ-ILA-003 — "its decision is made against the state the previous mutation published, never
  against a read taken before the wait" — which iteration 1's AC-ILA-002 traced but had **no
  mechanism to force**. A lock-free implementation with the same hook still yields `successes=2`
  (that is precisely AC-ILA-001), so GREEN continues to discriminate. The hook did not weaken the
  criterion; it made the one requirement that was previously unobservable observable.

### 2. The timeout-vs-budget ordering — real, or aspiration?

**Real and checkable, but quantified on only one side, and stated without margin (N3).**

The budget side is fully quantified and derived in-text: `boardLockWaitBudget` =
`boardLockSupportedWriters × boardLockCIMutationCost × boardLockHeadroom` = 10 × 33 ms × 5 = 1.65 s
(`plan.md` §B item 2, re-verified against `board_store.go:96-117`). The timeout side is named
(`stall-release timeout`) and bounded ("bounded", "shorter than the mutation-lock wait budget") but
**carries no number**. That is still checkable — the run phase can assert
`stallReleaseTimeout < boardLockWaitBudget` mechanically against a named constant — and it has a
loud observable backstop: `RESULT=busy` FAILS AC-ILA-002 with the misconfiguration named in the AC
itself. So this is not an aspiration.

What it lacks is **margin**. Strict inequality admits 1.6 s against a 1.65 s budget, where B has
been retrying under jitter for nearly its whole budget and a scheduling hiccup tips it into `busy`
— reintroducing flakiness into a MUST-PASS criterion by a different door. The correct form is a
stated ratio or absolute headroom (e.g. "≤ ⅓ of `boardLockWaitBudget`"), not `<`.

### 3. D1's discriminator — does it hold on EVERY path?

**Yes on every path the harness can take, including `force` and the unreadable-record path. One
narrow hole survives, and it is not the force path (N2).**

Path-by-path through `AcquireIntegrationLock` (`integration_lock.go:177-209`):

| Path | `replaced` | Reported | Discriminated? |
|---|---|---|---|
| Free record | nil | `RESULT=acquired REPLACED=none` | — (the RED-eligible path) |
| Held, different session, not stale, no force | nil | `RESULT=held REPLACED=none` | — (the GREEN path) |
| Held, different session, **STALE** (`:194`) | `current` | `REPLACED=<session>` | ✓ caught |
| Held, different session, **`force`** (`:192`) | `current` | `REPLACED=<session>` | ✓ caught — and unreachable anyway: `plan.md` M1's helper op reads `HELPER_ROOT`, `HELPER_SESSION`, `HELPER_OWNER_PID`, `HELPER_STALL_FLAG` and has **no force input**, so a child cannot take this branch |
| Unreadable record, no force (`:187-189`) | — | `RESULT=error` | ✓ not counted as a success |
| Unreadable record + force | nil (`Held()` has a nil-receiver guard at `:96`, so no panic) | — | unreachable: no force input |
| **Same session id** (`:190` false) | nil | `RESULT=acquired REPLACED=none` | **✗ NOT caught** |

The last row is the surviving hole. Two children acquiring with the *same* id both succeed with
`REPLACED=none` — indistinguishable, under every stated discriminator, from the read-write race.
The only thing preventing it in AC-ILA-001 is prose in the setup clause (`acceptance.md:60`, "child
B runs its entire acquire with a DIFFERENT session id"); nothing in the *observable* column asserts
the two ids differ. A harness bug that passed one id to both children would manufacture a false RED
that passes all three parts of the three-part observable. Narrow — it requires a specific harness
bug — but it is exactly the "every path" question, and closing it costs one clause (N2).

### 4. The 3-attempt ceiling — terminates, and owned?

**Both.** `spec.md` §G: "Either criterion observing zero double-holds in **3 consecutive attempts**
is a **blocker**: the run phase stops and escalates to the orchestrator with the attempt outputs,
and MUST NOT proceed to M2 (for AC-ILA-001) or record AC-ILA-006 as satisfied. No widen-and-retry
loop is authorized, and absence of the observation is never read as evidence the race does not
exist." Three is finite; the loop is explicitly de-authorized; the owner is the orchestrator; the
action is stop + escalate + carry the outputs; and the two forbidden continuations are each named.
Mirrored in `plan.md` M1 (escalation paragraph), M5, `plan.md` §G (anti-pattern), and both
AC-ILA-001 and AC-ILA-006. This is a concrete obligation with an addressee, not an unowned one.

### 5. Regression over the iteration-1 "already sound" list

All eight items re-read in full this pass. **No damage.**

| Item | Status |
|---|---|
| §A hypothesis framing | Unchanged verbatim, including the "**Hypothesis, not yet a measurement**" paragraph and the conditional consequence sentence. |
| The 14 verified citations (§B) | §B unchanged verbatim; all 14 still resolve. |
| Traceability matrix + §D.4 | Unchanged verbatim; §D.4's honest "review, not measurement" statement intact. |
| Six-heading Exclusions block | `grep -c '^### Out of Scope — '` = **6**, 8 specific bullets. Unchanged. |
| Trailing-`kill` rejection | Present in all four places (`spec.md:212`, `acceptance.md:24`, `plan.md:177`, `plan.md:265`). |
| Option (b) | Retained; argument restructured, not the decision. |
| M3's necessity | Retained; M3 body unchanged except the §G-aligned evidence sentence. |
| Lifetime separation (**D9 touched this — checked specifically**) | **Strengthened, not damaged.** REQ-ILA-007 gained two clauses that both push the same direction: the artifact must not share the `integration-lock` stem, and glob discovery of the record stays forbidden. `plan.md` M2 restates the reason at the artifact path. Nothing anywhere reads the mutation artifact to decide who holds the window. `plan.md` §G still carries "Persisting the mutation artifact as the window." One trivial nick: §C's preamble promises requirements are "implementation-neutral … not the code shape", and REQ-ILA-007 now names a literal filename stem — a small internal tension between the section's stated principle and one of its members (N4). |

### 6. New iter-2 citations — all resolve

| New at iter-2 | Verified |
|---|---|
| `integration_lock.go:113-121` (`Stale()` body, D1's mechanism) | `func (l *IntegrationLock) Stale() bool` at 113 through `}` at 121 ✓ |
| `internal/cli/integration.go:184` (production records the owner pid, not its own) | `ownerPID, _ := session.ResolveOwnerPID()` ✓; `PIDSource: kanban.PIDSourceSessionOwner` at :188 confirms the marker-parity claim ✓ |
| `PIDSourceSessionOwner` (the marker M1 requires) | declared `integration_lock.go:71`, documented :65 ✓ |
| `board_lock_clear_windows_test.go` (AC-ILA-007(b)'s second file) | exists, 8556 bytes ✓ |
| `board_store.go:117` + the 10 × 33 ms × 5 = 1.65 s arithmetic | `boardLockWaitBudget` at :117; factors at :96-117 ✓ |
| `board_lock_clear_windows.go` line 1 = `//go:build windows` | ✓ |
| `integration_lock.go:204` (hook insertion point) | Line 204 currently holds the `}` closing the `AcquiredAt` block, with the write at :205 — so "after the decision, before the write" describes the position exactly, though the line number will shift by the insertion itself. Descriptive, unambiguous; no finding. |

---

## Defects Found

All six iteration-1 blocking findings are closed and none recurred. Three new findings, all
introduced by the repair, all **optional** class. Per M6, an optional list does not manufacture a
FAIL and none of these is routed as a gate.

**N1.** ILA-HOOK-NIL-GUARD — `spec.md:143-152` (REQ-ILA-005 amendment) / `plan.md` M1 (hook bullet)
/ `acceptance.md` §D.5 gate 6 — The hook is specified as "nil-by-default" and "invoked once at
`:204`", but **no artifact states that the invocation is guarded by a nil check**. `grep -n 'nil'
… | grep -i hook` returns five hits, every one describing the *variable's* default, none
describing the *call site*. Calling a nil `func()` in Go panics, so an implementer following the
text literally could write `integrationLockMutationTestHook()` unguarded and panic every production
acquire — which would make REQ-ILA-005's own sentence ("With the hook nil … behavior is
byte-for-byte what this requirement states") false. Why this is optional rather than blocking: the
failure is immediate and loud, and AC-ILA-005 (`-run IntegrationLock`, MUST-PASS) panics on the
first acquire, so it cannot reach production. Gate 6 does not catch it because gate 6 checks
assignment, not invocation. — Severity: **minor** — Class: **optional** — Suggested fix: one clause
on the REQ-ILA-005 amendment — "invoked only when non-nil" — and, at no extra cost, switch gate 6's
`\s` to `[[:space:]]` so the gate does not depend on GNU grep.

**N2.** ILA-SAME-SESSION-UNASSERTED — `acceptance.md` AC-ILA-001 / AC-ILA-002 (observable columns)
— The three-part discriminator catches the stale path and the force path, but the **same-session-id
path** produces `successes=2` with `REPLACED=none` on both children and a non-`Stale()` winner —
passing all three parts. The distinct-id requirement lives only in the §D.2 setup prose
(`acceptance.md:60`), not in what the criterion asserts, so a harness bug passing one id to both
children would yield a false RED that survives every stated check. This is the one residual in
answer 3's path table. — Severity: **minor** — Class: **optional** — Suggested fix: add a fourth
part to AC-ILA-001's observable — "and the two children's reported session ids differ" — so the
precondition is asserted rather than assumed.

**N3.** ILA-TIMEOUT-NO-MARGIN — `plan.md` §C (budget-vs-timeout bullet) / `plan.md` M1 / `spec.md`
§G risk 1 — The ordering constraint is stated as a strict inequality against a quantified budget
(1.65 s) with the timeout side unquantified and no headroom. `stallReleaseTimeout = 1.6 s` satisfies
the constraint as written while leaving B retrying under jitter for essentially its whole budget,
so a scheduling hiccup produces `busy` and fails MUST-PASS AC-ILA-002 — flakiness re-entering by a
different door than iteration 1's. Not blocking because the failure is loud, named in the AC as a
misconfiguration rather than a verdict, and correctable in-run. — Severity: **minor** — Class:
**optional** — Suggested fix: state headroom rather than order — e.g. "the stall-release timeout
MUST NOT exceed ⅓ of `boardLockWaitBudget`" — which quantifies both sides and is assertable in the
test.

**N4.** ILA-REQ007-NEUTRALITY-NICK — `spec.md:115` (§C preamble) vs REQ-ILA-007 — §C promises
"Requirements are implementation-neutral: they fix the observable mutation contract, not the code
shape"; REQ-ILA-007 now names a literal filename stem (`integration-lock`) and a literal glob
(`.moai/state/integration-lock*`). The D9 repair is right and I asked for it, but it landed a
code-shape constraint inside a section that declares itself shape-free. — Severity: **minor** —
Class: **optional** — Suggested fix: either soften to "shall not share the record's filename stem"
without the parenthetical literals, or add a half-sentence to the §C preamble acknowledging that
REQ-ILA-007's naming clause is a deliberate exception and why.

---

## Recommendation

**PASS.** The six blocking findings are closed, each verified against the files rather than against
the summary, and the closures are substantive rather than cosmetic. Two are worth naming as genuine
engineering improvements rather than compliance:

- **D2's hook did more than fix D2.** By forcing A to publish before B decides, it made REQ-ILA-003
  — the re-read-after-waiting requirement — observable for the first time. Iteration 1's AC-ILA-002
  claimed to cover REQ-ILA-003 and had no mechanism to produce the ordering it needed. That was a
  latent traceability hole I did not name at iteration 1, and the repair closed it as a side effect.
- **D3+D4's collapse onto one test function** removed an entire class of closure-gate failure: no
  criterion now names a function absent from the deliverable, and "One test, two tree states" is a
  cleaner articulation of RED/GREEN than the two-function shape it replaced.

The four new findings are all minor and all optional. My recommendation to the orchestrator: route
**N1** and **N3** into the run phase's opening checklist rather than a further plan iteration —
each is a single clause, both concern the code the run phase writes first, and neither warrants
re-opening the artifacts on their own. N2 and N4 are the author's discretion.

Nothing in this iteration triggers the STOP escalation: the score moved up, not down. The card is
cleared to proceed to the Implementation Kickoff Approval gate, which this verdict does not and
cannot bypass.

I modified no SPEC artifact, ran no `go test`, spawned no background process, and made no commit or
push.
