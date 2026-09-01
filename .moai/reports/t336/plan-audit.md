# SPEC Review Report: SPEC-INTEGRATION-LOCK-ATOMIC-001 (card t336)

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.80** (Tier M PASS threshold = 0.80 — the aggregate sits exactly on the line; the
verdict is decided by the six blocking findings below, not by the aggregate)

Tree under audit: `15453140a`, branch `WT-integration-lock-atomic`, worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t336`.
Reasoning context ignored per M1 Context Isolation. Artifacts read per the Tier M input contract
(`spec.md` + `plan.md` + `acceptance.md`), plus `progress.md` and the named supporting context.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-ILA-[0-9]*' spec.md | sort -u` returns
  REQ-ILA-001..011, sequential, no gaps, no duplicates, consistent 3-digit padding. 11 REQs ≤ the
  Tier M ceiling of 16.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-ILA-*`
  in `spec.md` §C). All 11 match a GEARS pattern: Ubiquitous (001, 005, 007), Event-driven (002,
  003, 010, 011), State-driven (004), Capability-gate/`Where` (006), Unwanted (008, 009 — both in
  the canonical `shall not` form). Verification-layer Given-When-Then in `acceptance.md` §D.2 is
  the correct AC format and is graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types in
  all four artifacts: `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created`/`updated` (ISO `2026-08-29`), `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string). Plus `tier: M`. No rejected
  snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) appears.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/kanban`), not
  template-bound. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — two SPEC-ID references extracted:
  `SPEC-INTEGRATION-LOCK-LIVENESS-001` and `SPEC-KANBAN-BOARD-001`. Both directories exist under
  `.moai/specs/`; `grep -n '^status:'` returns `5:status: completed` for each (the
  `172:status: draft` in KANBAN-BOARD-001 is body text, not frontmatter). Neither is retired,
  superseded, or archived. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -n 'syscall' spec.md` returns nothing
  (rc=1), so D8 auto-passes on its literal trigger. Substantively the SPEC exceeds the bar: it
  mandates build-tagged files and `GOOS=windows go vet ./internal/kanban/...` (spec.md §D, plan.md
  M3, AC-ILA-007).
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/`
  returns only two lines, both of which are negative assertions that no marker is carried
  (`plan.md:104`, `progress.md:30`). No unresolved marker.

**No must-pass failure.** The FAIL verdict below is driven by the blocking defect set, not by the
firewall.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Prose is unusually precise and every boundary is named. Two load-bearing ambiguities remain: the PID the M1 helper records is never fixed (D1), and `plan.md` M1 contradicts `acceptance.md` on how many tests exist at closure (D4). |
| Completeness | 0.90 | 1.0-band, docked | All sections present; six `### Out of Scope — <topic>` H3 sub-headings each with specific `-` bullets (spec.md:255-283). Docked for the absent stop rule on the RED tuning loop (D2). |
| Testability | 0.65 | 0.50-0.75 | Two MUST-PASS criteria (AC-ILA-001, AC-ILA-006) turn on a probabilistic event with no stated ceiling; AC-ILA-001's deciding command is not runnable as written (D3); AC-ILA-009's command can match its own invocation (D7). AC-ILA-004/005/008 are exemplary — binary, with a named zero-baseline. |
| Traceability | 1.00 | 1.0 | `acceptance.md` §D.3 covers all 11 REQs; every AC names its REQ(s); §D.4 states honestly that REQ-ILA-007 and REQ-ILA-011 are verified by review rather than dressing them as measurable ACs. No orphan, no uncovered REQ. |

Harmonic mean = 4 / (1/0.75 + 1/0.90 + 1/0.65 + 1/1.00) = **0.803**.

---

## Answers to the six named hazards

### 1. Vacuous-criterion hunt — is AC-ILA-001 falsifiable and reachable?

**Falsifiable: yes. Reliably reachable: no. Stated plainly: the round count and the
zero-observation handling are NOT sound as written, and a deterministic interleaving mechanism is
required.**

The window is `integration_lock.go:186..205` — one `os.ReadFile`, a branch, one
`os.WriteFile` + `os.Rename`. On a warm filesystem that is tens of microseconds. Two children
released on a barrier flag diverge by the barrier's poll granularity, which the SPEC never fixes;
with a 1 ms polling barrier the per-round hit rate is a few percent, with a busy-spin barrier it is
much higher. The SPEC's response to a zero-observation run (`spec.md` §G) is "widen the window
(more rounds, or a slower filesystem path) and re-measure" — an unbounded loop with no ceiling, no
falsification point, and "a slower filesystem path" left undefined. A run phase that keeps widening
until RED appears is tuning until luck arrives; one that stops early proceeds to M2 with no
observation at all.

The minimal deterministic fix is a test-only hook invoked once between the decision and the write —
a package-level `var integrationLockMutationTestHook func()`, nil by default, called at
`integration_lock.go:204`, and set only by the cross-process child under an env flag. That converts
both the RED and the mutation guard from "wait for an interleaving" to "observe by construction".
Two costs the SPEC must then own explicitly: the hook is an edit to the *unrepaired* path, which
breaks AC-ILA-001's "on tree `15453140a`" wording (see D3); and REQ-ILA-005's "single-threaded
semantics unchanged" must be amended to permit a nil-by-default hook.

If the author declines the hook, the SPEC must at minimum carry a stated ceiling — "N rounds; zero
observations after N is a blocker, escalate rather than proceed to M2" — so the run phase can
neither loop indefinitely nor proceed silently.

### 2. Does AC-ILA-006 actually defend against vacuous green?

**In mechanism, yes. In reliability, no.**

The one-line revert is genuinely expressible given plan.md M2's chosen shape: inserting
`return fn()` as the first statement of `withIntegrationLockMutation(projectRoot, fn)` is one line,
leaves `fn` used so the package still compiles, and subtracts *exactly* mutual exclusion — the
in-section re-read required by REQ-ILA-003 lives inside `fn` and survives the revert. So disabling
it does not merely change an error message; it removes the property under test. That is a real
guard.

What it does not survive is D2. The FAIL direction AC-ILA-006 must observe is the identical
probabilistic event as AC-ILA-001, and `spec.md` §G's widen-and-remeasure escape is written for
AC-ILA-001 only and is not carried to AC-ILA-006. A MUST-PASS criterion whose failure direction may
simply not fire is a criterion that can itself go unestablished. Fixing D2 fixes this too.

### 3. The (b) design choice — is the deciding argument accurate, and is M3 scope creep?

**(b) is the right choice. The stated deciding argument is overstated.** Verified at this tree:
`AcquireBoardLock` is at `board_lock.go:75` and does return `ErrBoardLockHeld` without retry; its
`os.MkdirAll(BoardDir(root))` is at `board_lock.go:77`; `acquireBoardLockSerialized` is at
`board_store.go:156`. Every cited fact resolves. The *inference* does not: the board's actual
mutation path is `acquireBoardLockSerialized`, which retries `AcquireBoardLock` under
`boardLockWaitBudget` (`board_store.go:117` — 10 writers × 33 ms × 5 headroom = 1.65 s). An
option-(a) implementation would naturally use that same serialized entry point, so "a `moai todo`
write in flight would make `moai integration acquire` fail" does not follow from the cited lines —
it follows only if the implementer picked the raw non-retrying call, which nothing forces. What
survives is weaker and remediable by wrapping: the *sentinel* would be `ErrBoardLockHeld` on an
integration verb. Argument 2 (board-dir side effect) is accurate but minor. Argument 3
(scope-correctness — answering a scope-mismatch bug by borrowing a wider-scoped lock) is the
argument that actually decides, and the plan states it last and unquantified while asserting
"items 1 and 3 are [deciding]". See D5.

**M3 is genuinely necessary, not scope creep.** On Windows `acquireBoardLockImpl`
(`board_lock_windows.go:55`) returns `ErrBoardLockHeld` immediately on `os.IsExist` with no stale
handling, so a killed short-lived mutation-lock holder wedges every later mutation permanently. The
only alternative to parameterizing `ClearStaleBoardLock`'s path-keyed core is duplicating roughly
sixty lines of grace / pre-removal-re-read / liveness-probe logic, which is strictly worse. And
`board_lock_clear_windows.go` is `//go:build windows` (line 1), so the change cannot affect Unix
behavior at all — the risk is smaller than `spec.md` §G states, though §G's *mitigation* is wrong
in the opposite direction (D6).

### 4. Lifetime conflation

**The SPEC holds this line well** — this is the dimension I expected to find broken and did not.
REQ-ILA-007 states it; `spec.md` §D carries it as a binding constraint; `plan.md` §B closes the
design section with it; `plan.md` §G names "persisting the mutation artifact as the window" as an
anti-pattern; `acceptance.md` §D.4 states honestly that it is verified negatively (absence of a
call site in the diff) rather than dressing review as measurement. I found no sentence anywhere in
the four artifacts that lets the mutation lock's lifetime stand in for the window's.

One residual, classed optional: the mutation artifact is planned at
`.moai/state/integration-lock.mutation.lock`, in the same directory as `integration-lock.json` and
adjacent to the `.tmp` staging file M4 removes. REQ-ILA-007's negative verification is a one-time
diff review that does not survive into a lint or a test, so nothing stops a future reader from
globbing `.moai/state/integration-lock*`. I verified no current consumer is at risk — the four
non-test consumers (`internal/cli/integration.go:131` / `:185` / `:234`,
`internal/hook/integration_lock_guard.go:74`) all call named functions and none globs.

### 5. Scope boundaries — confirmed clean

`grep -n "go test \./\.\.\.\|internal/hook\|internal/config\|days\|hours\|week\|ASAP"` across all
four artifacts returns 12 lines, every one of which is an exclusion, a boundary statement, or a
prohibition:

- `internal/hook/**` / `internal/config/**` — appear only at `spec.md:185`, `:258-259`, `:108-109`
  (premise citation), `plan.md:97`, `:191`, `acceptance.md:154` (an emptiness gate). No requirement
  or AC modifies either tree.
- `moai integration release` error classification / message text — `spec.md` §D card boundary,
  `spec.md` § Out of Scope, `plan.md` §G, `acceptance.md` §D.5.3. Byte-identity is a closure gate.
- `go test ./...` — appears only as a prohibition (`spec.md:195`, `plan.md:98`, `:185`,
  `acceptance.md:20`). Every deciding command is package-scoped.
- Time estimates — none. Milestones carry `Priority: High/Medium` only.

### 6. Honesty — §A framing and citation spot-check

**§A framing holds.** `spec.md` §A carries an explicit "**Hypothesis, not yet a measurement**"
paragraph, attributes the gap to `preflight.md` §6, and states "nothing in this document may be
cited as a measured defect before [run-phase RED]". The consequence sentence is correctly
conditional ("The consequence, if the hypothesis holds"). What §A *does* assert as measured is
structural and verified: `grep -n 'flock\|Flock' internal/kanban/integration_lock.go` returns lines
15 and 19 only — both prose, exactly as claimed.

**One sentence upgrades the hypothesis** — `acceptance.md` AC-ILA-001's Claim column: "two
concurrent acquires from two SEPARATE OS processes **are** BOTH told they hold the window",
indicative mood. Filed as D8 (minor). It is the only such sentence in four artifacts.

**Citation spot-check — 14 references checked, 14 resolve.**

| Cited | Verified at `15453140a` |
|---|---|
| `integration_lock.go:186` READ | `current, err := ReadIntegrationLock(projectRoot)` ✓ |
| `integration_lock.go:190-200` decide | `if current.Held() && ...` switch ✓ |
| `integration_lock.go:205` WRITE | `if err := writeIntegrationLock(path, &want)` ✓ |
| `integration_lock.go:216/220-225/226` release | read / `ErrIntegrationLockNotHeld` + foreign check / `os.Remove` ✓ |
| `integration_lock.go:257` fixed tmp | `tmp := path + ".tmp"` ✓ (AC-ILA-008's zero-baseline confirmed, rc=0) |
| `integration_lock.go:13-26` header | lifetime paragraph + coordination-signal paragraph ✓; L18-20 is the false claim ✓ |
| `integration_lock.go:99-112` fail-direction | `Stale()` doc comment, asymmetry paragraph ✓ |
| `integration_lock.go:156-159` re-acquire | "a lane that re-enters after a `/clear`" ✓ |
| `board_lock.go:75-86` / `:76-79` / `:124` | `AcquireBoardLock` at 75, `os.MkdirAll` at 77, `newLockOwnerRecord` at 124 ✓ |
| `board_lock_unix.go:36` / `:4-7` | `acquireBoardLockImpl` at 36; kernel-releases-flock note at 4-7 ✓ |
| `board_lock_windows.go:55` / `:3-8` | `acquireBoardLockImpl` at 55; "the artifact IS the lock" at 3-8 ✓ |
| `board_lock_clear_windows.go:75` / `:66-74` | `ClearStaleBoardLock` at 75; TOCTOU-residual paragraph immediately above ✓ |
| `board_store.go:114-151` / `:156` | retry-wait policy block / `acquireBoardLockSerialized` ✓ |
| `kanban_helper_test.go:22-113`, `board_lock_cross_test.go:43-104` / `:84-104` | `TestKanbanHelperProcess` at 22 (file is 114 lines); `TestBoardMutation_SerializedAcrossProcesses` at 43, `..._ConcurrencyPositiveControl` at 84 ✓ |
| `internal/cli/integration.go:131/185/234`, `internal/hook/integration_lock_guard.go:74` | status read / acquire / release / guard read ✓ |

No citation failed to resolve. This is the strongest part of the submission.

---

## Defects Found

**D1.** ILA-PID-UNSPECIFIED — `plan.md` M1 (helper op spec) / `acceptance.md` AC-ILA-001, AC-ILA-002
— The PID the M1 helper records into the lock is never specified, and both the RED and the GREEN
criteria hinge on it. `AcquireIntegrationLock` records `want.PID` verbatim (`integration_lock.go`
doc at :170-176, write at :205) and `Stale()` (`:112-119`) returns `false` for `PID <= 0` but
`!FactoryProcessAlive(pid)` otherwise. The helper op is specified as reading only `HELPER_ROOT`,
`HELPER_SESSION`, and a barrier flag. If the child records its own `os.Getpid()`, child A exits
immediately after writing, its record then reads STALE, and child B's acquire legitimately takes it
over — producing `successes=2` **with a correct mutation lock in place**. That breaks AC-ILA-002
for a reason unrelated to atomicity, and makes AC-ILA-001's RED a *misattributed* observation
(stale takeover, not the read-write race) — worse than no observation, because it looks like
success. AC-ILA-003 cannot catch it: the same-session-id path skips the `current.SessionID !=
want.SessionID` branch entirely. — Severity: **critical** — Class: **blocking** — Required fix:
(a) fix the recorded PID explicitly in `plan.md` M1 — a pid guaranteed live for the round (the
parent test process) or `0` (which `Stale()` treats as live by design); and (b) add a discriminator
to AC-ILA-001 so a `successes=2` round is attributed to the read-write window rather than to a
stale takeover — e.g. assert the winner's record is NOT `Stale()` at the moment of the second
acquire, or assert the loser's outcome is a refusal rather than a `replaced` takeover.

**D2.** ILA-RED-PROBABILISTIC — `spec.md` §G / `acceptance.md` AC-ILA-001, AC-ILA-006 — Two
MUST-PASS criteria turn on a probabilistic interleaving with no stop rule. §G's zero-observation
handling ("widen the window — more rounds, or a slower filesystem path — and re-measure") names no
ceiling, no falsification point, and no defined mechanism for "a slower filesystem path"; and it is
written for AC-ILA-001 only, so AC-ILA-006 — whose required FAIL direction is the same event —
carries no escape at all. The `~200 rounds` starting point in `plan.md` M1 is marked "tuned
in-run", which makes the criterion's own decidability a run-phase discretion. — Severity:
**critical** — Class: **blocking** — Required fix: add a deterministic interleaving mechanism (a
nil-by-default, test-only hook invoked between the decision and the write, set only by the
cross-process child under an env flag), amend REQ-ILA-005 to permit it explicitly, and extend the
zero-observation clause to AC-ILA-006. If the hook is declined, state a hard ceiling and make a
zero-observation run a blocker that escalates rather than proceeding to M2.

**D3.** ILA-AC001-UNRUNNABLE — `acceptance.md` AC-ILA-001 (§D matrix row 1) — The deciding command
is `go test ./internal/kanban/... -run TestIntegrationLockAcquire_DoubleHoldObserved -count=1 -v`
"on tree `15453140a` (repair reverted)". That test does not exist at `15453140a` — it is created in
M1. What is meant is "at HEAD with the critical section disabled". The looseness is easy to miss
because AC-ILA-005 uses `15453140a` correctly, as a `git diff` base. Compounding it, the observable
column reads "the test FAILS on the unrepaired tree by design" while the row is MUST-PASS, so the
criterion's polarity is decidable only by reading the prose. — Severity: **major** — Class:
**blocking** — Required fix: restate the tree as "HEAD with M2's critical section disabled (the
AC-ILA-006 revert)" and state the pass condition positively — "the test reports at least one round
with `successes=2`" — rather than as "the test fails".

**D4.** ILA-TEST-COUNT-CONTRADICTION — `plan.md` M1 (closing note) vs `acceptance.md` §D matrix
rows 1-2 — `plan.md` M1 says "After M2, M1's test **flips** to the GREEN assertion";
`acceptance.md` names two distinct functions, `TestIntegrationLockAcquire_DoubleHoldObserved` and
`TestIntegrationLockAcquire_SerializedAcrossProcesses`, as if both exist at the delivered tree. If
M1's test flips, AC-ILA-001's named test is absent from the deliverable and its command cannot be
run at closure — while closure gate §D.5.1 requires every MUST-PASS criterion's command *and its
verbatim output* to be recorded. The two artifacts describe different deliverables. — Severity:
**major** — Class: **blocking** — Required fix: decide explicitly. Either retain the RED test as a
distinct, revert-gated test (in which case `plan.md` M1's "flips" sentence is wrong and both
functions ship), or redefine AC-ILA-001 as a run-phase-only historical observation recorded in
`progress.md` §E.2 and remove it from the closure-gate command set.

**D5.** ILA-OPTION-A-OVERSTATED — `plan.md` §B, "Why not (a)", item 1 — The cited facts resolve but
the inference does not. `AcquireBoardLock` (`board_lock.go:75`) indeed does not retry, but the
board's mutation path is `acquireBoardLockSerialized` (`board_store.go:156`), which retries under
`boardLockWaitBudget` (`board_store.go:117` = 1.65 s). An option-(a) implementation would use that
same entry point, so "a `moai todo` write in flight would make `moai integration acquire` fail"
does not follow — it follows only from a choice nothing forces. What survives is the weaker,
wrapping-remediable point that the *sentinel* would be `ErrBoardLockHeld` on an integration verb.
The plan then asserts "items 1 and 3 are [the deciding factors]", resting half its decision on the
overstated half. Item 3 (scope-correctness) is the argument that actually decides and is stated
last and unquantified. **This does not change the verdict on (b), which remains correct** — it
changes the honesty of the reasoning a later reader will re-derive from. — Severity: **major** —
Class: **blocking** — Required fix: correct item 1 to name `acquireBoardLockSerialized` and its
budget, restate the residual cost as error-classification plus cross-scope serialization latency,
and promote item 3 to the stated deciding argument.

**D6.** ILA-M3-MITIGATION-CANNOT-FIRE — `spec.md` §G, third risk bullet — "the mitigation is that
the board's own behavior and its existing criteria must remain green, which AC-ILA-005 and the
scoped package run cover." `board_lock_clear_windows.go` is `//go:build windows` (line 1), so on
the darwin lane machine no test compiling that file runs at all — neither AC-ILA-005 nor
`go test ./internal/kanban/...` can observe a regression in the code M3 edits. The only local
signal is `GOOS=windows go vet`, which proves compilation and not behavior; the only behavioral
judge is CI's windows job. A mitigation that cannot fire against the risk it is offered for is a
false assurance, and it is the kind that survives review because it names real commands. —
Severity: **major** — Class: **blocking** — Required fix: restate the mitigation as "compilation
checked locally by `GOOS=windows go vet ./internal/kanban/...`; behavioral non-regression of
`ClearStaleBoardLock` judged by CI's windows job", and add the board's existing windows clear
criteria explicitly to AC-ILA-007(b)'s CI-judged surface.

**D7.** ILA-PGREP-SELF-MATCH — `acceptance.md` AC-ILA-009 (§D matrix, §D.2) — `pgrep -fl
'kanban\.test' | wc -l` matches on full command line, so a measuring pipeline whose own command
line contains the literal `kanban.test` can be counted. "Both readings are `0`" may therefore be
undecidable as written, and a spurious `1`/`1` reads as a clean run rather than as a broken
measurement. — Severity: **minor** — Class: **optional** — Required fix: exclude the measuring
process (build the pattern from a shell variable, or filter `$$` and the pipeline's own pid) and
restate the observable as "no PID other than the measuring pipeline, before and after".

**D8.** ILA-CLAIM-MOOD — `acceptance.md` AC-ILA-001, Claim column — "two concurrent acquires from
two SEPARATE OS processes **are** BOTH told they hold the window" is indicative, asserting the race
as present fact. `spec.md` §A, `plan.md` §A, `plan.md` §G, and `progress.md` §E.1 all hold the
hypothesis framing scrupulously; this is the one surface that drops the hedge, and a Claim column
is exactly what a later reader quotes. — Severity: **minor** — Class: **optional** — Required fix:
"…**shall be** BOTH told…", or prefix "the criterion asserts that…".

**D9.** ILA-MUTATION-ARTIFACT-NEIGHBOURING — `plan.md` M2 (artifact path) — The mutation artifact
is planned at `.moai/state/integration-lock.mutation.lock`, sharing a directory and a filename stem
with the window record `integration-lock.json`. REQ-ILA-007's separation is verified negatively by
one-time diff review (`acceptance.md` §D.4) and does not survive into any test or lint, so nothing
mechanically stops a future reader from globbing `.moai/state/integration-lock*` and folding the
two lifetimes back together. No current consumer is at risk (verified: the four non-test consumers
all call named functions; none globs). — Severity: **minor** — Class: **optional** — Required fix
(discretionary): name the artifact without the record's stem (e.g. `integration-mutation.lock`), or
add a one-line note under REQ-ILA-007 that the stem-sharing is deliberate and glob-based discovery
of the record is forbidden.

---

## Recommendation

FAIL. Six blocking findings, concentrated in exactly the place the card's value lives: the
bidirectional regression evidence. Ordered by what unblocks the most:

1. **Fix D1 first** — it is the highest-consequence defect and the cheapest to close. Fix the PID
   the M1 helper records (`plan.md` M1) and add the stale-takeover discriminator to AC-ILA-001 and
   AC-ILA-002 (`acceptance.md` §D matrix rows 1-2 and §D.2). Without this, a run-phase RED can be
   observed for the wrong cause and reported as success — the failure mode with no later signal to
   correct it.
2. **Fix D2** — add the deterministic interleaving hook (nil-by-default, test-only, invoked at
   `integration_lock.go:204`), amend REQ-ILA-005 to permit it, and extend `spec.md` §G's
   zero-observation clause to AC-ILA-006. If the hook is declined, state a hard round ceiling and
   make a zero-observation run an escalating blocker.
3. **Fix D3 and D4 together** — they are the same underlying question (does the RED test survive to
   the delivered tree?). Decide it once, then correct AC-ILA-001's tree wording and pass polarity,
   and `plan.md` M1's "flips" sentence, to agree.
4. **Fix D5** — correct `plan.md` §B item 1 to name `acquireBoardLockSerialized` and its 1.65 s
   budget; promote item 3 (scope correctness) to the deciding argument. The decision for (b) stands
   and needs no re-litigation.
5. **Fix D6** — restate `spec.md` §G's third risk mitigation as `GOOS=windows go vet` for
   compilation plus CI's windows job for behavior, and add the board's existing windows clear
   criteria to AC-ILA-007(b).

D7-D9 are optional: surface them to the author, route them at the orchestrator's discretion, and do
not hold the card for them.

What is already sound and should not be touched in the repair: the §A hypothesis framing, the 14
verified citations, the traceability matrix and its honest §D.4 indirect-verification statement,
the six-heading Exclusions block, the cleanup obligation's rejection of a trailing `kill`, the
choice of option (b), and the necessity of M3. On the lead's hazard 4 specifically — the lifetime
separation the audit was warned would most likely fail silently — I found no conflation anywhere in
the four artifacts.

Re-audit on iteration 2 will be scoped to the D1-D6 defect delta plus a regression check over
this list.
