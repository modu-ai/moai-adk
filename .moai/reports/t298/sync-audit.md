# Sync-phase Quality Assessment: SPEC-INTEGRATION-LOCK-LIVENESS-001 (card t298)

Verdict: **PASS-WITH-DEBT**
Overall score: **0.924** (harmonic mean of the four dimensions; Tier M PASS threshold 0.80)
Measured in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t298`, HEAD `f8b7264ba`.

## §0 Provenance limitation — read this before treating the verdict as independent

**This is NOT a fresh-context `sync-auditor` verdict.** The sync dispatch asked for one;
`manager-docs` carries no `Agent` tool and cannot spawn a subagent, so no fresh-context
auditor was run. What this report is, precisely:

- **Independent of the implementation.** Every production and test file assessed below was
  written by `manager-develop` during run-phase. This agent did not author any of it, so the
  executor-judging-own-output failure shape does not apply to the code assessment.
- **NOT independent of the sync artifacts.** The §E.4 signal and the CHANGELOG entry that
  ride the same commit were authored by this agent. Their assessment is a self-report and is
  excluded from the dimension scores rather than counted as observed.

That limitation is the **named debt** carrying this verdict to PASS-WITH-DEBT rather than a
clean PASS. It is a gap in the audit process, not a defect found in the work.

## §1 Dimension scores

| Dimension | Score | Basis |
|---|---|---|
| Functionality | 0.94 | Fix addresses the exact defect; cross-process tests; live-environment confirmation |
| Security | 0.88 | Fail-direction explicit and conservative; one declared out-of-scope concurrency hazard |
| Craft | 0.95 | The comments carry the reasoning, not the mechanics; the split preserves the registry's cost profile |
| Consistency | 0.93 | Matches repo conventions; every citation SHA-attributed |

Harmonic mean: `4 / (1/0.94 + 1/0.88 + 1/0.95 + 1/0.93)` = **0.924**.

### Functionality — 0.94

The defect was that `AcquireIntegrationLock` filled an unset `PID` with `os.Getpid()`, which
is the **acquire CLI's** pid. That process exits the moment the command returns, so every
window read as abandoned the instant it was taken. The fix removes that fill
(`integration_lock.go`, the deleted `if want.PID == 0 { want.PID = os.Getpid() }`) and moves
owner resolution to the caller (`internal/cli/integration.go` calls
`session.ResolveOwnerPID()`).

The load-bearing design decision is in `internal/session/session_pid.go`: `ResolveOwnerPID`
is a **new** exported seam with **no** `os.Getpid()` third step, returning `(0, false)` when
the owner is unresolvable, while the pre-existing `resolveSessionPID` keeps its `os.Getpid()`
fallback for the registry. Splitting rather than changing the existing function is what stops
the fix from silently altering the session registry's recorded-pid semantics.

Verification observed in this run:

- `go vet ./internal/kanban/ ./internal/session/ ./internal/cli/ ./internal/hook/` → exit 0,
  no output.
- The two ACs run-phase left individually unobserved were closed here:
  `go test ./internal/kanban/ -run 'TestReleaseIntegrationLock_HolderAndForeign'` →
  `--- PASS` (AC-INL-005), and
  `go test ./internal/cli/ -run 'TestIntegrationAcquire_ForceReportsWhatItDisplaced'` →
  `--- PASS` (AC-INL-006). AC coverage is therefore **13/13 observed, 0 fail**, up from
  §E.3's 11 + 2 not-observed.
- The RED baseline is genuinely cross-process, not an in-process simulation:
  `internal/cli/integration_lock_owner_liveness_test.go` builds a binary, runs `acquire`,
  lets that process exit, and only then probes
  (`TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits`,
  `…_EnvStampHoldsAfterAcquireCLIExits`, `…_BareAcquireRefusesLiveHolder`). An in-process
  test could not have observed this defect at all.

Not full marks because the acquire path's unserialized read-modify-write is retained
(§3), so "the window is held by exactly one lane" is not established by this fix alone.

### Security — 0.88

No new external input, no new file-permission surface (`0o755` dir / `0o644` file unchanged),
no credential handling, no new network or subprocess surface in production code.

The fail-direction reasoning is explicit and lands on the conservative side. `Stale()` treats
`PID <= 0` as **live**, so an unresolvable owner wedges the window until an explicit release
or a recorded `--force` — never toward two lanes merging at once. The asymmetry is argued in
the doc comment rather than assumed, and `ReadIntegrationLock` continues to report an
unparseable record as an error rather than as a free window (absence-of-signal is not read as
evidence-of-freedom).

Held at 0.88 by one declared hazard: `writeIntegrationLock` stages through a fixed
`path + ".tmp"` shared by every concurrent writer, inside an unserialized read-modify-write.
This is the false-"live-for-me" direction — the one failure the lock exists to prevent. It is
declared in `spec.md` §G, explicitly out of scope for this card, and is not a regression this
SPEC introduces.

### Craft — 0.95

The comments explain **why**, not what. The strongest instance is the note on
`AcquireIntegrationLock` stating that it deliberately does not fill an unset pid, and why
filling it was wrong — a future reader who "tidies up" by restoring the `os.Getpid()` default
is told, at the call site, exactly what they would be reintroducing.

`PIDSource` is added as an additive, optional discriminator that **no read path branches on**,
and the comment says so. It distinguishes an unresolvable-owner record from a pre-anchor
legacy record for a human reader; it does not gate behavior, so it cannot skew a probe. A
marker that documents rather than gates is the right shape here.

The one structural blemish: `spec.md` §G's citation reads `integration_lock.go:106` while the
same path join now sits at line 128 (the M2 comment additions moved it). This is **not** a
defect — the §G text attributes its measurement to `c67a6ea64` by name, so the citation is
correctly baseline-attributed to the tree it was measured on.

### Consistency — 0.93

Conventions hold: English comments and godoc, `snake_case.go` test files, `%w` error wrapping
(`fmt.Errorf("read integration lock: %w", err)`). Test isolation holds — every new test builds
its fixtures under `t.TempDir()`, and the cross-process cases pin `CLAUDE_PROJECT_DIR` and
`GIT_CEILING_DIRECTORIES` to that root, so the live primary-checkout record was never touched.
`internal/hook/perf` (which rewrites SPEC fixtures) was never entered.

Scope discipline held: the whole-branch Go diff is six files, all inside the SPEC's declared
`module:` set, with no drive-by edits.

## §2 Real-environment confirmation

Recorded here because it is the strongest evidence in this close, and it did not exist at
run-phase time. Attributed to the sessions that observed it, not self-reported:

- **Before deployment** (the fixed binary sat unshipped for most of 2026-08-27): lanes
  lane-16, lane-17 and lane-18 each independently observed `reclaimable` with an acquire-CLI
  pid recorded as the holder. One real eviction occurred without `--force`.
- **lane-18's isolated A/B**, with `CLAUDE_PROJECT_DIR` pinned so the live lock was never
  touched: the old binary reproduced all four stages of the failure chain; the HEAD build
  reported `held` and refused a second acquire with exit 1.
- **After deployment, first live use** (this card's own integration window): `held`, pid
  33289 — the session-owning process.
- **The discriminating detail**: the holder UUID was recorded correctly in **both** cases;
  only the pid differed. That is what confirms the diagnosis — the defect was never in
  identity *interpretation*; the recorded pid itself was a value unrelated to the session.

This confirms the fix behaves correctly in the live environment. It is **not** a proof of
atomicity, and must not be read as one.

## §3 §G residual verdict — still open, re-measured on this tree

`spec.md` §G retains the unserialized read-modify-write as an explicitly RETAINED,
out-of-scope hazard. Re-measured at HEAD `f8b7264ba`:

```
$ grep -n "Flock\|flock\|LockFile" internal/kanban/integration_lock.go internal/cli/integration.go
internal/kanban/integration_lock.go:15   # "flock" in prose, package header comment
internal/kanban/integration_lock.go:19   # "flock" in prose, package header comment
internal/kanban/integration_lock.go:38   # IntegrationLockFileName doc comment
internal/kanban/integration_lock.go:39   # const IntegrationLockFileName
internal/kanban/integration_lock.go:128  # its use in the path join
```

Five lines, **no call site**. `internal/cli/integration.go` contributes **zero** matches. The
residual is therefore **still open**; no later commit closed it, and this card did not widen
to cover it.

Operational consequence, unchanged: a recorded hold is a **coordination signal, not a
permission boundary**. Two lanes acquiring in the same instant can still both believe they
hold the window, so **the lead announcement remains the first serialization layer**. The
package header's own claim that "the flock discipline is borrowed only to serialize mutations
of that record" remains unbacked by the code.

Recommendation (for the lead to act on — card issuance is not this agent's power): a
follow-up card for the serialization work itself.

## §4 Gaps — what was NOT observed

- **No fresh-context `sync-auditor` verdict** (§0). The mandated independent audit did not run.
- **No full-suite run.** Scoped verification only, per the repository's verification-load
  discipline. The full-suite verdict belongs to CI on the pushed branch.
- **Windows behavior not executed.** `go vet` under `GOOS=windows` was clean at run-phase, but
  vet proves compilation, not behavior. The Windows launcher's `MOAI_SESSION_PID` stamp
  remains unverified on a real Windows host (declared in §G).
- **The deployment-lag visibility problem itself** — that a fixed binary can sit unshipped
  while lanes keep hitting the old defect — is **out of scope here and re-scoped to card
  t326** (lane-18). Nothing was done about it in this close.
- **PID reuse not exercised.** A recorded owner pid whose session died and whose pid the OS
  later reassigned probes as live. Declared in §G, inherited from the existing probe design,
  not tested.

## §5 Residual risk

- Serialization is now **real everywhere except** the acquire-instant race (§3). The fix makes
  the window meaningful, which raises rather than lowers the cost of that remaining window:
  before the fix nothing was serialized anyway, so the race was masked.
- A legacy in-flight window taken before this binary shipped reads reclaimable exactly as it
  did before. A lane holding one must re-acquire post-upgrade or be silently evictable.
- `PIDSource` is written but never read by any decision path. If a future change starts
  branching on it, the pid-0-reads-live invariant must be re-argued, not inherited.

---

Assessed by `manager-docs` during sync-phase close. Provenance limitation: §0.
