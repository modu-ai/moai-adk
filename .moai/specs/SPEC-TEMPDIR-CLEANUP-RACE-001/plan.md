# SPEC-TEMPDIR-CLEANUP-RACE-001 — implementation plan

Tier S. One milestone. Ordering below is by decision-reversibility: the API-shape decision leads,
the mechanical edits follow.

---

## §A Context

Read `spec.md` §A and the evidence base `.moai/reports/t352/reproduction.md` before starting. The
plan does not restate the mechanism.

---

## §B Known issues entering the run

- The lane worktree carries an untracked measurement probe, `internal/cli/zz_t352_probe_test.go`.
  It was retained for the run phase. Decide its fate explicitly (keep as a skipped benchmark, or
  delete) rather than letting it land by accident.
- `internal/cli`'s full package was never run during reproduction. The run phase must run it, since
  the guard lands there.

---

## §C Pre-flight

- Confirm `git rev-parse --short HEAD` and `git branch --show-current` immediately before any commit
  (branch `WT-tempdir-cleanup-race`).
- Confirm `internal/hook/session_start.go:41` (the non-variadic constructor), `:240`, `:608-611`,
  `:1606-1613`, `internal/cli/binary_lag_test.go:57`, and `internal/cli/deps.go:221` are unchanged
  from the SHA the evidence base cites (`77b2bcae6`); if the branch has moved, re-read them before
  editing.
- Record `git merge-base origin/develop HEAD` at run-phase entry — this is AC-TCR-002b's base and it
  is unrecoverable later (`acceptance.md` § Base-SHA attribution).

---

## §D Mechanism ladder — decision (highest reversibility cost; decided first)

The simplicity ladder is applied cheapest-existing-capability-first
(`moai-constitution.md` § Agent Core Behaviors #4). Four rungs were evaluated.

### D.1 Rung evaluation

| # | Rung | Verdict | Why |
|---|------|---------|-----|
| 1 | **Export a deliberate seam from `internal/hook`** — a functional option on `NewSessionStartHandler` (e.g. `hook.WithSynchronousDeferredScans()`) that any caller, test binary or in-process embedder, can pass to request the inline path | **TAKEN** | Ladder step 2 (reuse what already exists here): the package **already** encodes exactly this concept in `deferredScansAsync` / `deferredScansAsyncEnabled()`, and already has a fully-exercised inline branch at `session_start.go:269-283`. The only defect is that the switch is package-private and so cannot cross a test-binary boundary. Exporting it adds an option, not a behaviour. It is also the only rung that actually removes the durable write from outliving `Handle` at the caller boundary rather than hiding the consequence |
| 2 | **Change the call site only** — replace `t.TempDir()` with `os.MkdirTemp` plus a cleanup that tolerates the late write (retry `RemoveAll`, or ignore the error) | REJECTED | Cheapest rung, and it would silence this CI failure. But it leaves the leak in place and merely stops one test noticing it: the next cross-package caller inherits the same race, un-signposted, and the guard REQ-TCR-004 asks for would have nothing to guard. Cost of taking it: the defect survives with its only detector removed |
| 3 | **Make the deferred goroutine refuse the write when the target directory is gone** — stat `projectDir` before `runMXColdStartScan` writes | REJECTED | Does not close the race. The failing window is precisely the one where the directory still exists while `RemoveAll` is walking it; a stat then returns success and the write proceeds into a tree being deleted. It converts a reliable race into a narrower race that is harder to reproduce. Cost of taking it: a fix that measures as an improvement while the defect remains |
| 4 | **Give `Handle` a join handle** — return an additional waitable the caller can block on | REJECTED | Changes the production `Handle` signature and contract for every caller, to serve one test's need. Strictly more API surface than rung 1 for the same effect, and it invites callers to wait on the production path — the exact regression `spec.md` §C forbids |

### D.2 Tension acknowledged, not waved away

Rung 1 adds production API surface for what is presently a test-only need. That is a real cost and
the argument for accepting it is narrow: the concept is already in the package as a private
variable, the option makes an existing branch reachable rather than adding one, and the alternative
(rung 2) deletes the only signal that the defect exists. If the option were introducing a new code
path, the balance would go the other way.

### D.3 Design constraints on the chosen rung

- **The option parameter MUST be variadic.** The exported constructor is
  `func NewSessionStartHandler(cfg ConfigProvider) Handler` (`internal/hook/session_start.go:41`) —
  **non-variadic**, one parameter. There is a cross-package **production** caller,
  `internal/cli/deps.go:221` (`deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))`),
  so adding the seam as a second positional argument would break that line. Introducing it as a
  variadic option parameter (`NewSessionStartHandler(cfg ConfigProvider, opts ...Option) Handler`)
  keeps `deps.go:221` compiling **unchanged**, which is the required outcome — not merely the
  convenient one. Mechanically verified twice: AC-TCR-002b asserts `internal/cli/deps.go` is absent
  from the change's file list, and AC-TCR-006 / AC-TCR-007 compile and run the package that contains
  it. A diff touching `deps.go` means the seam was added in the wrong shape.
- **Off by default.** No option ⇒ current production behaviour, byte-for-byte. REQ-TCR-002.
- **Per-handler, not a package-level exported setter.** A `hook.SetDeferredScansSync(bool)` would be
  process-global mutable state and would race across parallel tests — the exact class of defect
  `internal/hook/main_test.go` exists to prevent. The option belongs on the handler value.
- **Interaction with the existing private seam.** `deferredScansAsyncEnabled()` (package-global) and
  the new per-handler option must compose without either overriding the other's intent: the inline
  path runs when **either** the private test-binary seam is inline **or** the handler was
  constructed with the option. `internal/hook`'s own tests, and
  `session_start_parallel_test.go`'s deliberate opt-back-into-async, must be unaffected.
- **`binaryLagAdvisory` is a separate decision, and its exemption is cited rather than asserted.**
  `session_start_binary_lag.go:51` reads the same private seam, and it is called from `Handle` at
  `session_start.go:479`. Its background goroutine performs a **read-only** git comparison — only
  `git rev-parse HEAD` and `git merge-base --is-ancestor` (`internal/binlag/binlag.go:101,111`), no
  write call exists anywhere in the package — so it is outside REQ-TCR-001. The exemption is also
  belt-and-braces: AC-TCR-001 compares the caller directory's **whole entry set**, so a writer this
  plan wrongly exempted would still be caught (`spec.md` §A.2). Threading the option to
  `binaryLagAdvisory` is optional: do it only if it costs one argument, and say so in the commit; do
  not expand scope to chase it.

### D.4 Tier classification and the recorded budget overrun

**Tier stays S, and the overruns are recorded rather than absorbed.** This SPEC carries nine
acceptance criteria against the Tier S ceiling of 8 (`spec-workflow.md:148`), in a separate
`acceptance.md` against the Tier S 2-file set (`spec-workflow.md:140`). Both are stated here so the
next reader meets a declared overrun rather than an unnoticed one.

*Why not reclassify to Tier M.* Tier is classified by implementation scope, and this SPEC's scope is
squarely Tier S: one variadic option plus a branch condition in `internal/hook`, one call-site edit
in `internal/cli/binary_lag_test.go`, one new guard test, and a disposition decision on the probe —
well under 300 LOC and under 5 files. Tier M describes 300-1000 LOC across 5-15 files, which is
false of this work. Tiering up so that a third artifact and a ninth criterion become "in budget"
would classify the SPEC to fit its paperwork rather than its scope — the inversion the tier taxonomy
exists to prevent — and would silently raise the plan-auditor threshold from 0.75 to 0.80 on the
basis of a scope claim that is not true.

*Why no criterion was folded or deleted.* The obvious fold, AC-TCR-004a into AC-TCR-004b, is the
worst available: 004a is the RED observation — the only criterion in the set that establishes the
guard has been seen to fail — and merging it into the GREEN cell would leave a guard whose red has
never been observed, exactly the defect `verification-completeness.md` §1.1 names. Removing a check
to make a count fit shrinks the coverage the criteria provide; that is the wrong direction, and a
budget is not a reason to take it.

*Why each retained criterion earns its place.*

| AC | Why it cannot be dropped |
|----|--------------------------|
| AC-TCR-001 | The guard itself — the only criterion asserting REQ-TCR-001's invariant. |
| AC-TCR-002a | Cheapest possible check that the production default is untouched; a one-line grep with no overlap elsewhere. |
| AC-TCR-002b | The only criterion covering diff *shape*: the async branch body, the join bound, and the untouched production call site (D1's constraint). |
| AC-TCR-003 | Non-regression on the originally-flaking test under `-race`; disclaims itself as non-evidence of the fix, which is precisely why it is not a substitute for 004. |
| AC-TCR-004a | The RED observation. Load-bearing; without it the guard is unproven. |
| AC-TCR-004b | The GREEN cell of the two-cell pair; without it, RED does not show this work can flip the guard. |
| AC-TCR-005 | The blast-radius check on `internal/hook` itself, incl. goleak and the parallel-test opt-back-into-async. |
| AC-TCR-006 | Cross-platform compile — the only criterion covering windows/linux, and the compile proof for `deps.go`. |
| AC-TCR-007 | Whole-package run, incl. the sibling residue guard and the CI-headroom wall-clock the §D constraint requires. |

*Why `acceptance.md` stays a separate file.* Inlining nine criteria plus their commands, passing
outputs, base-SHA attribution, severity, and traceability into `spec.md` §3 would make the SPEC body
the longest artifact in the set and bury the requirements and exclusions the run phase actually reads
first. The 2-file Tier S prescription assumes an AC set small enough to inline; this one is not, and
the honest response is to record the deviation, not to hide it by compressing the criteria.

---

## §E Milestone M1 (single milestone)

Priority: High. Ordered within the milestone by reversibility.

1. **Seam (`internal/hook`).** Add the functional option as a **variadic** parameter on
   `NewSessionStartHandler` (per §D.3 — `internal/cli/deps.go:221` must keep compiling untouched)
   plus the per-handler field; make the `session_start.go:240` branch consult it. No change to
   `deferredScansAsync`'s default, to `deferredScanJoinBound`, or to the async branch's body.
2. **Guard (`internal/cli`).** Add the regression guard described in `acceptance.md` AC-TCR-004: a
   test owning its own directory, padded so the scan is measurably slow, that asserts no new entry
   appears under that directory after `Handle` returns.
3. **RED observation.** Mutate — revert step 1's option at the call site (or pass the async path
   explicitly) — run the guard, record the verbatim failure, restore. A guard never seen red is not
   a guard (`verification-completeness.md` §1.1).
4. **Call site.** Change `internal/cli/binary_lag_test.go:57` to construct the handler with the
   option. REQ-TCR-003.
5. **Probe disposition.** Delete or deliberately retain `internal/cli/zz_t352_probe_test.go`; state
   which in the commit message.
6. **Verification.** Run the `acceptance.md` command set; push and read CI for the full-suite
   verdict per `CLAUDE.local.md` §4.

---

## §F Self-verification

Every AC in `acceptance.md` names its deciding command. The run phase records each command's
verbatim output in `progress.md` §E.2 with the HEAD SHA it was measured on. Two additional
recordings are mandatory:

- **AC-TCR-002b's base SHA.** Record `git merge-base origin/develop HEAD` next to the diff reading.
  The base is a derived value (`77b2bcae6` at plan time) and is unrecoverable once refs move; a
  mid-card merge of `develop` moves it forward and must be noted with the re-recorded base.
- **AC-TCR-007's wall-clock.** Record the `internal/cli` package duration so the guard's added cost
  against CI's 10-minute default per-package timeout (`.github/workflows/ci.yml:238`) is measured
  rather than assumed (`spec.md` §D).

---

## §G Anti-patterns for this card

- **Making the production path wait.** Any diff that causes the default `Handle` to block on
  `runMXColdStartScan` fails REQ-TCR-002 regardless of how green the tests are.
- **Guarding with an empty fixture.** The reproduction measured the empty-directory case flipping
  between two bases. A guard whose fixture is an unpadded directory is itself a flake.
- **Grepping for the fixed form as evidence.** Verify by observing the mutant red and the fixed
  green on the same tree, not by grepping for the option's name.
- **Editing the production call site to accommodate the seam.** If `internal/cli/deps.go` appears in
  the diff, the option was added as a positional argument instead of a variadic one (§D.3). The fix
  is to change the seam's shape, not to update `deps.go:221`.
- **Widening to observation 1.** `spec.md` §C excludes it. A diff touching `internal/graph` is out
  of scope for this card.
- **Reporting observation 2 as "low-frequency".** One CI appearance is not a rate (`spec.md` §A.4).
  Its cost is unquantified, not small; the 1-in-5 figure in the t322 verdict belongs to a different
  test that this card excludes.

---

## §H Cross-references

- `spec.md` — requirements and exclusions.
- `acceptance.md` — the AC matrix and its commands.
- `.moai/reports/t352/reproduction.md` — evidence base.
- `.claude/rules/moai/development/verification-completeness.md` — the two-cell (RED-now + green
  path) adoption discipline and the mutant probe.
