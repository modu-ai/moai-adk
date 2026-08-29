# SPEC-TEMPDIR-CLEANUP-RACE-001 — acceptance criteria

Every criterion below is binary and names the command that decides it. Commands run from the
repository root of the working tree (the lane worktree during the run phase). The HEAD SHA each
command was measured on is recorded alongside its output in `progress.md` §E.2. Where a criterion
judges a **diff**, the base SHA is recorded as well — see AC-TCR-002b, whose base is
`origin/develop` @ `77b2bcae6` (the fork point; see § Base-SHA attribution there).

> **Tier S budget overrun, deliberate and recorded.** Nine criteria against the Tier S ceiling of 8
> (`spec-workflow.md:148`), in a separate file against the Tier S 2-file set
> (`spec-workflow.md:140`). Both overruns are stated rather than absorbed; `plan.md` §D.4 carries the
> per-criterion justification. Nothing here was folded or deleted to make a count fit.

---

## §D AC matrix

| AC | Requirement | Deciding command |
|----|-------------|------------------|
| AC-TCR-001 | REQ-TCR-001 | `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=5 -timeout 600s` |
| AC-TCR-002a | REQ-TCR-006 | `grep -n 'var deferredScansAsync = true' internal/hook/session_start.go` |
| AC-TCR-002b | REQ-TCR-002, REQ-TCR-006 | `git diff origin/develop...HEAD -- internal/hook/session_start.go` + `git diff --name-only origin/develop...HEAD -- internal/cli/deps.go` (empty) + the same over `internal/cli/binary_lag_test.go` (non-empty positive control) |
| AC-TCR-003 | REQ-TCR-003 | `go test ./internal/cli/ -run TestBinaryLag_OneSeamServesBothSurfaces -race -count=20 -timeout 900s` |
| AC-TCR-004a | REQ-TCR-004 (RED) | mutation, then `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=1 -timeout 600s` |
| AC-TCR-004b | REQ-TCR-004 (GREEN) | same command, mutation reverted |
| AC-TCR-005 | REQ-TCR-005 | `go test ./internal/hook/ -race -count=1 -timeout 900s` |
| AC-TCR-006 | §D constraints | `GOOS=windows GOARCH=amd64 go vet ./internal/hook/... ./internal/cli/...` and `GOOS=linux GOARCH=amd64 go vet ./internal/hook/... ./internal/cli/...` |
| AC-TCR-007 | §D constraints | `go test ./internal/cli/ -count=1 -timeout 900s` (whole package, incl. the `TestMain` residue guard) |

---

## §D.1 Criteria in Given-When-Then form

### AC-TCR-001 — no durable write outlives `Handle` under the option

- **Given** a directory the test itself creates and owns, padded with enough files that
  `runMXColdStartScan`'s `ScanDir` reliably finishes **after `Handle` returns** — that, not the
  250 ms join bound, is the operative condition: the reproduction measured `Handle` returning at
  ~60 ms in every row (`Handle=59.59ms` at `nFiles=8000`) because the advisory arrives long before
  the bound, so the bound is never reached in these runs. The index appeared 43 ms after return at
  2000 files and 223 ms after return at 8000 files, so the fixture pads to at least 8000,
- **When** the test calls `hook.NewSessionStartHandler(nil, <synchronous-deferred-scans option>).Handle`
  with `ProjectDir` set to that directory, records the directory's entry set at the instant `Handle`
  returns, waits a bounded settle period of at least 1 s, and records it again,
- **Then** the two entry sets are identical, and in particular
  `<dir>/.moai/state/mx-index.json` is either present in both snapshots or absent from both — never
  absent in the first and present in the second.
- **Command:** `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=5 -timeout 600s`
- **Passing output:** `ok  github.com/modu-ai/moai-adk/internal/cli` with exit status 0.
- **Scope note:** the comparison is over the directory's whole entry set, not over the MX index path
  alone. That is what makes this criterion match REQ-TCR-001's "every durable write" scope rather
  than only the single writer the reproduction established (`spec.md` §A.2).

### AC-TCR-002a — the production default is untouched

- **Given** the post-change tree,
- **When** the execution-mode seam's declaration is read,
- **Then** it still declares the production default `true`.
- **Command:** `grep -n 'var deferredScansAsync = true' internal/hook/session_start.go`
- **Passing output:** exactly one matching line; exit status 0.

### AC-TCR-002b — the async branch, the join bound, and the production call site are unchanged

- **Given** the diff of this card against its base,
- **When** `internal/hook/session_start.go` is inspected and `internal/cli/deps.go` is listed,
- **Then** (i) the `session_start.go` diff adds an option and a branch condition only: the body of
  the async branch (`spawnDeferredAdvisoryScans` + the `select` on `joinTimer`) and the value of
  `deferredScanJoinBound` are unmodified; and (ii) `internal/cli/deps.go` is **absent** from the
  change's file list, proving the variadic seam left the production call site
  (`internal/cli/deps.go:221`) compiling untouched (`spec.md` §D) — an emptiness result that counts
  only when the positive control below is non-empty.
- **Commands:**
  - `git diff origin/develop...HEAD -- internal/hook/session_start.go`
  - `git diff --name-only origin/develop...HEAD -- internal/cli/deps.go` (the assertion)
  - `git diff --name-only origin/develop...HEAD -- internal/cli/binary_lag_test.go` (the positive
    control — same diff form, same base, over a path this change is known to touch per `plan.md`
    §E step 4)
- **Passing output:** for (i), no hunk touching the `select { case advisory := <-advisoryCh: ... }`
  block or the `deferredScanJoinBound` declaration. Judged by reading the diff, and the reading is
  recorded. For (ii), the assertion command prints **nothing** AND the control command prints
  `internal/cli/binary_lag_test.go`. Both outputs are recorded.
- **Why the control is mandatory — the emptiness is otherwise vacuous.** `git diff --name-only`
  over a pathspec prints nothing in two entirely different situations: the file is genuinely
  unchanged (the result this AC wants), and the pathspec matched nothing at all (a typo, a moved or
  renamed file, a wrong base, a command run from the wrong directory). An emptiness check cannot
  distinguish them by itself, and the situation it silently accepts is the failure mode — a
  mistyped path passes this AC forever while `deps.go` is being edited. The control removes the
  ambiguity: it exercises the identical command form against a path the change certainly touches.
  **If the control also comes back empty, the pathspec form itself is broken and the assertion's
  empty result means nothing** — do not read it as a pass. Record both, fix the command, and re-run
  the pair.

#### Base-SHA attribution (mandatory)

The three-dot form is deliberate and load-bearing: `origin/develop...HEAD` diffs from
`merge-base(origin/develop, HEAD)` — the branch's **fork point** — not from `origin/develop`'s
moving tip, so an advancing develop does not change what this criterion compares. That was measured,
not assumed: `origin/develop` was already at `947f5cffb` when the plan audit ran and at `c6aa61346`
when these repairs were made, while `git merge-base origin/develop HEAD` stayed `77b2bcae6`
throughout — the SHA the evidence base cites.

The base is nevertheless a **derived** value and cannot be reconstructed once refs move, so:

- **The base SHA in force at plan time is `77b2bcae6`** (`origin/develop` fork point), recorded here
  so the attribution survives regardless of where the refs go.
- The run phase MUST record the output of `git merge-base origin/develop HEAD` next to its reading
  in `progress.md` §E.2, alongside the HEAD SHA. A reading of a diff whose base is unattributed is
  an unattributed claim (`verification-claim-integrity.md` §2).
- **A mid-card merge of `develop` into this branch moves the merge-base forward.** That is the
  semantically correct baseline for "this card's diff", but it silently changes what this criterion
  compared, so any such merge MUST be noted in `progress.md` §E.2 together with the re-recorded base.

### AC-TCR-003 — the original failing test uses the seam and stays green under `-race`

- **Given** `internal/cli/binary_lag_test.go` constructing its handler with the option,
- **When** the previously-flaking test runs repeatedly under the race detector,
- **Then** every iteration passes and no `TempDir RemoveAll cleanup: unlinkat ... directory not empty`
  line appears.
- **Command:** `go test ./internal/cli/ -run TestBinaryLag_OneSeamServesBothSurfaces -race -count=20 -timeout 900s`
- **Passing output:** `ok`; exit status 0; the string `directory not empty` absent from the output.
- **Note (carried from `spec.md` §F):** 50 local iterations passed on the *pre-fix* tree too. This
  AC is a non-regression check, **not** the evidence that the defect is fixed — AC-TCR-004 is.

### AC-TCR-004a — the guard is observed RED before it is trusted (mutation)

- **Given** the guard from AC-TCR-001 present on the tree,
- **When** the seam is mutated back to the defective behaviour — the guard's own `Handle` call is
  changed to omit the synchronous option, so the async path with its 250 ms join runs — and the
  guard is executed once,
- **Then** the guard **fails**, and its failure message names the entry that appeared after `Handle`
  returned.
- **Command:** apply the mutation, then
  `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=1 -timeout 600s`
- **Passing output for this AC:** a `FAIL` line and exit status 1, with the verbatim failure text
  recorded in `progress.md` §E.2 together with the mutation's diff.
- **Why this AC exists:** a check whose red has never been seen has proven nothing, whatever its
  green says (`verification-completeness.md` §1.1). It also excludes the vacuous direction: a guard
  that is green on the defective tree is not measuring the defect. This is the only criterion in the
  set that establishes the guard has been seen to fail; it is not foldable into AC-TCR-004b.

### AC-TCR-004b — the mutation is reverted and the guard returns green

- **Given** AC-TCR-004a recorded,
- **When** the mutation is reverted and the same command re-run,
- **Then** the guard passes.
- **Command:** `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=1 -timeout 600s`
- **Passing output:** `ok`; exit status 0.
- **Pairing note:** AC-TCR-004a and AC-TCR-004b are adopted as a pair. 004a alone does not show the
  work can flip the guard; 004b alone does not show the guard measures anything.

### AC-TCR-005 — `internal/hook` is unaffected, goleak still clean

- **Given** the new option in the package,
- **When** the whole `internal/hook` test binary runs under `-race`,
- **Then** every test passes and `goleak.VerifyTestMain` reports no leaked goroutine — in particular
  `session_start_parallel_test.go`'s deliberate opt-back-into-async still behaves as before.
- **Command:** `go test ./internal/hook/ -race -count=1 -timeout 900s`
- **Passing output:** `ok  github.com/modu-ai/moai-adk/internal/hook`; exit status 0; the string
  `found unexpected goroutines` absent.

### AC-TCR-006 — cross-platform compile

- **Given** the change,
- **When** the two affected packages are vetted for the non-host platforms,
- **Then** both vets succeed — which also proves the production call site `internal/cli/deps.go:221`
  still compiles against the new constructor signature (`spec.md` §D).
- **Command:** `GOOS=windows GOARCH=amd64 go vet ./internal/hook/... ./internal/cli/...` and
  `GOOS=linux GOARCH=amd64 go vet ./internal/hook/... ./internal/cli/...`
- **Passing output:** no output; exit status 0 for each.
- **Scope note:** `go vet` proves compilation only, not behaviour
  (`feedback_gate_vet_only_proves_compilation`). Behaviour on those platforms is CI's verdict.

### AC-TCR-007 — the whole `internal/cli` package, including its residue guard

- **Given** the guard landing in `internal/cli`,
- **When** the whole package runs,
- **Then** it passes and the `TestMain` residue guard does not fire — no `RESIDUE GUARD FAIL` line
  and no `internal/cli/.moai` directory created by the run.
- **Command:** `go test ./internal/cli/ -count=1 -timeout 900s`
- **Passing output:** `ok`; exit status 0; the string `RESIDUE GUARD FAIL` absent.
- **Wall-clock recording (`spec.md` §D, CI headroom):** the run's reported package duration is
  recorded in `progress.md` §E.2, so the guard's added cost against CI's 10-minute default
  per-package timeout (`.github/workflows/ci.yml:238`) is measured rather than assumed.
- **Why:** reproduction never ran this package whole (`spec.md` §F). The new guard writes into a
  directory it owns, but the assertion that it does not also write into the package working
  directory has to be made, not assumed.

---

## §D.2 Severity

| AC | Severity |
|----|----------|
| AC-TCR-001, AC-TCR-004a, AC-TCR-004b | MUST — the card's entire purpose |
| AC-TCR-002a, AC-TCR-002b | MUST — the rejected-regression boundary in `spec.md` §C and the §D seam-shape constraint |
| AC-TCR-003, AC-TCR-005, AC-TCR-006, AC-TCR-007 | MUST — non-regression |

## §D.3 Traceability

| REQ | AC |
|-----|-----|
| REQ-TCR-001 | AC-TCR-001 |
| REQ-TCR-002 | AC-TCR-002b |
| REQ-TCR-003 | AC-TCR-003 |
| REQ-TCR-004 | AC-TCR-004a, AC-TCR-004b |
| REQ-TCR-005 | AC-TCR-005 |
| REQ-TCR-006 | AC-TCR-002a, AC-TCR-002b |
| §D constraints | AC-TCR-006, AC-TCR-007 |

## §D.4 Definition of Done

- Every AC above recorded PASS with its verbatim command output and the HEAD SHA measured on;
  AC-TCR-002b additionally records its base SHA (`git merge-base origin/develop HEAD`).
- `internal/cli/zz_t352_probe_test.go` disposition stated in the commit message.
- CI green on the pushed head — CI, not the local run, is the full-suite verdict
  (`CLAUDE.local.md` §4).
- `spec.md` §F risks re-read at close and re-affirmed as still open; observation 1 explicitly
  reported as **not** closed by this card.

## §D.5 What this AC set does NOT establish

- It does not establish that the CI failure string is gone from CI: CI was never re-run during
  reproduction and the guard measures the write, not the collision (`spec.md` §F).
- It does not establish anything about `internal/graph` / observation 1.
- It does not establish that no other `t.TempDir` caller in the repository has a similar race; no
  such sweep was performed.
- It does not establish how often this race bites in CI. Observation 2 has a single CI appearance
  and no measured frequency — **one appearance is not a rate** — so its cost is unquantified, not
  low (`spec.md` §A.4).
