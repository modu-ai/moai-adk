# t352 — lane verdict

Card: t352 · Lane worktree: `.claude/worktrees/t352` · Branch: `WT-tempdir-cleanup-race`
SPEC: `SPEC-TEMPDIR-CLEANUP-RACE-001` (Tier S) · HEAD `cebebae8a`, 6 commits, **unpushed**
Merge-base: `origin/develop` @ `48d8ef4be` · Written: 2026-08-29

---

## Claim

The card was opened on two observations sharing a symptom string. **One was reproduced and fixed;
the other was not reproduced and is untouched.** That split is the verdict.

- **Fixed** — `TestBinaryLag_OneSeamServesBothSurfaces` (`internal/cli`, `.moai/state: directory
  not empty`). Mechanism established from source and measured, repair landed, guard observed RED
  twice for the stated reason and green after.
- **Not fixed, and not claimed** — `TestGitDiffNameCount_Predicate` (`internal/graph`,
  `.git/objects: directory not empty`). Ten iterations across two operating systems found no
  post-command writer. Cause unestablished.
- **Landing this does not turn `develop` green.** It closes one of `develop`'s current reds, not all
  of them (see the update below).

  **Update (2026-08-29, head `a6bbbf82b`, runs `33244792917` / `33244792905`).** The picture at the
  earlier head `c6aa61346` — where the *excluded* observation was `develop`'s sole failing test —
  no longer holds. On `a6bbbf82b` the failing set is:

  | Job | Failing test | Owner |
  |---|---|---|
  | `Test (ubuntu-latest)` | `TestBinaryLag_OneSeamServesBothSurfaces` | **this card** |
  | `Race Test` | `TestConcurrencyStress` | t354 |
  | `spec-lint` (separate run) | — | not this card |

  Two consequences, both against this lane's earlier framing: the observation this card **fixes** is
  currently red on `develop` and the one it **excludes** did not fire on this head, which is the
  reverse of the head the verdict was first written against. And the failure is again in the
  **non-race** job — a fourth observation of detector-independence, now on a head this lane did not
  choose.
- **No CI verdict is claimed.** Every figure here is local, darwin/arm64; the branch has never been
  pushed, so `acceptance.md` §D.4 clause 3 ("CI green on the pushed head") is **unmet**. The SPEC
  reads `completed` on local evidence alone — a judgement, not something the evidence forces.

## Evidence

### The mechanism, and why the fix is shaped as it is

`internal/hook`'s `TestMain` neutralises the deferred goroutine by flipping the package-private
`deferredScansAsync`; that cannot cross a test-binary boundary. `internal/cli` therefore sees the
production value, and `Handle` returns on the 250 ms advisory join while `runMXColdStartScan` is
still ahead of it, writing `<ProjectDir>/.moai/state/mx-index.json` into the directory `t.TempDir`
cleanup is removing. The ordering is deliberate and documented in the source
(`session_start.go:608-611`).

The repair exposes that same capability as an opt-in seam: `hook.Option`,
`hook.WithSynchronousDeferredScans()`, and a **variadic** `NewSessionStartHandler(cfg
ConfigProvider, opts ...Option)`. Variadic because the production call site `internal/cli/deps.go:221`
had to keep compiling untouched.

Full reproduction record, including the two corrections made to it mid-card:
`.moai/reports/t352/reproduction.md`.

### Independent verification — the lane's own batch, not the agent's report

Each command run by this session at HEAD `c27917886` (implementation) unless noted:

| # | Check | Command | Result |
|---|---|---|---|
| V1 | commits, tree | `git log --oneline -3`, `git status --short` | 2 commits, tree clean |
| V2 | seam shape | `grep -n 'func NewSessionStartHandler'` | `:85` variadic |
| V2 | default untouched | `grep -n 'var deferredScansAsync = true'` | `:1546`, one line |
| V2 | probe deleted | `ls internal/cli/zz_t352_probe_test.go` | `No such file or directory` |
| V3 | `deps.go` unchanged | `git diff --name-only origin/develop...HEAD -- internal/cli/deps.go` | empty |
| V3 | positive control | same form over `internal/cli/binary_lag_test.go` | `internal/cli/binary_lag_test.go` (non-empty) |
| V4 | build | `go build ./...` | rc=0 |
| V4 | vet (compiles tests) | `go vet ./internal/hook/... ./internal/cli/...` | rc=0 |
| V5 | the guard | `go test ./internal/cli/ -run TestSessionStartDeferredWrite… -count=3` | rc=0 |
| V6 | goleak | `go test ./internal/hook/ -race -count=1` | rc=0, 31.892s |
| V7 | whole package | `go test ./internal/cli/ -count=1` | rc=0, 239.814s; `RESIDUE GUARD FAIL` count 0; `internal/cli/.moai` absent |
| V8 | **mutation RED** | option dropped, `-count=1` | **rc=1** (below) |
| V8 | revert | option restored, `-count=1` | rc=0, tree clean |

V8 verbatim — the lane applied the mutation itself rather than accepting the run-phase report:

```
--- FAIL: TestSessionStartDeferredWriteDoesNotOutliveHandle (1.88s)
    session_start_deferred_write_test.go:131: a durable write outlived Handle: 1 entr(ies)
      appeared under the caller's directory during the 1s settle window after Handle returned:
      [.moai/state/mx-index.json]
        REQ-TCR-001 requires every durable write into the caller's ProjectDir to complete before
        Handle returns.
```

The guard is red **for the stated reason**, naming the entry — not merely red.

At HEAD `cebebae8a` (the `@MX:NOTE` addition, comment-only): `go build ./internal/hook/...` rc=0,
`go vet ./internal/hook/...` rc=0, `gofmt -l` empty.

### The excluded observation is currently `develop`'s only red

Read directly from `.moai/state/verify/lead-ci/c6aa61346-failed.log` (primary checkout), run
`33173944485`, head `c6aa61346`:

```
51: Race Test  Run race detector across all packages  --- FAIL: TestGitDiffNameCount_Predicate (0.14s)
52: Race Test  ...  testing.go:1464: TempDir RemoveAll cleanup: unlinkat
      /tmp/TestGitDiffNameCount_Predicate2957256880/001/.git/objects: directory not empty
```

One `FAIL:` line in the file. `Test (ubuntu-latest)` passed on that head — confirmed not from the
log (which carries only `Race Test` rows and so cannot establish it) but from
`gh run view 33173944485 --json jobs`, which the plan agent queried when it declined to record the
claim on insufficient evidence.

### The fixed observation is not race-detector-dependent

Read from `.moai/state/verify/aad73b87/ci-48d8ef4be-failed.log` (primary), run `33175578950`,
`attempt=1`, head `48d8ef4be` — **this branch's own merge-base**, i.e. the pre-fix tree the seam was
written against:

```
969: Test (ubuntu-latest)  Run tests with coverage (fast — no race detector)
       --- FAIL: TestBinaryLag_OneSeamServesBothSurfaces (0.01s)
970: ...  testing.go:1464: TempDir RemoveAll cleanup: unlinkat
       /tmp/TestBinaryLag_OneSeamServesBothSurfaces1836762982/001/.moai/state: directory not empty
```

Four hit lines across **two** jobs. The non-race job names itself, so the defect needs no detector
to bite — pure timing, exactly the established mechanism. This supersedes the single Race-only
observation carried from `t322/verdict.md`.

### Audit and phase record

Plan audit iter-1 PASS 0.88 → iter-2 **PASS 0.93** (Tier S threshold 0.75), monotonic, MUST-PASS all
green. Ten defects raised; D8 partially closed (a requirement created to absorb D1 re-paired two
subjects), the rest closed. Report: `.moai/reports/t352/plan-audit.md`.

Commits: `4d055c2e6` plan · `00e9723dc` excluded-observation CI record · `05678676e` develop merge ·
`410f6241d` M1 implementation · `c27917886` run evidence + `draft → in-progress` · `804402bdc` sync
(CHANGELOG, 3-phase close, §E.4) · `a5c65f28e` `sync_commit_sha` backfill + CHANGELOG restructure ·
`cebebae8a` `@MX:NOTE`.

## Baseline-attribution

- Every V-row above was run by this session in this worktree, at the HEAD stated in that row.
- CI figures are read from log files in the **primary checkout** and are attributed to their run id,
  head, and attempt. None was produced by this lane.
- The 15→14→1 baseline comparison against `5e194bba2` is a **carry** from the lead's reading; this
  lane did not re-measure the baseline run.
- `origin/develop` was 58 commits ahead at sync time and was **not** merged; the branch stands on
  merge-base `48d8ef4be`.

## Gaps

- **CI has never run this branch.** The full-suite and cross-platform verdict does not exist yet.
- **The guard measures the write, not the collision.** It proves the write lands after `Handle`
  returns; it never produces the `unlinkat ... directory not empty` string. Mechanism → CI-string is
  a well-supported inference, not an observation.
- **Observation 1 untouched**, cause unestablished. Ruled out: in-process goroutines, package
  parallelism, and any writer detectable within 750 ms on macOS or Linux aarch64. The untried lever
  is the race detector **under amd64** — this lane's Linux run was aarch64 and passed.
- **Frequency unmeasured for both observations.** Two jobs in one run is not a rate, and there are
  counter-observations (50 local pre-fix iterations passed; two further local runs passed).
- **`golangci-lint` was not run** — no AC named it and the lane did not add it unilaterally.
- **`gofmt` is dirty repo-wide** (45 files pre-existing); only the file this card made unclean was
  formatted.
- **The three non-`ProjectDir` seam readers have no AC.** Their coverage is a design judgement,
  evidenced only by `internal/hook`'s goleak run — which cannot see `internal/cli`, as that package
  has no goleak `TestMain`.
- **No sync-auditor pass.** This verdict is the lane's own reading.

## Residual risk

- **`develop` moved 58 commits and will move further before integration.** Everything above is
  measured on `48d8ef4be`; the integration merge can invalidate any of it, so the window's work is
  merge → re-verify → integrate, not integrate on these figures.
- **`completed` was recorded without CI.** If the integration CI reds this branch, the status is
  ahead of its evidence and must be walked back rather than explained.
- **The card's "one class, two instances" premise remains half-supported.** The two observations now
  differ measurably — one is detector-independent, the other detector-only — which is evidence
  *against* a shared cause rather than for it. A future card should not inherit the class reading.
- **Errors this lane made and corrected, recorded so they are not re-made**: a frequency figure
  belonging to observation 1 was applied to observation 2; a `_test.go`-scoped grep was reported as
  the blast radius while a production caller existed; an exit code was read from a pipe's tail
  rather than the command; and an ambiguous operator message ("병합 동결 완료 확인") was resolved in
  the permissive direction and relayed onward as a settled authorisation. The last was corrected by
  asking the operator directly; no action had followed from it.
