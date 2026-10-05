# Acceptance Criteria — SPEC-GOBIN-GOTOOLCHAIN-001

Card: t969. Tier S (6 criteria, within the Tier S ceiling of 8).

## D.0 — Rules binding every criterion below

- **R-1 — The precompiled form is named, not implied.** Where a criterion names the
  precompiled-binary form, `go test` is NOT an acceptable substitute for it: `go test` prepends
  the resolved toolchain's `bin` to the test process's PATH, so the child `go` has nothing to
  resolve and the condition cannot arise (`spec.md` §1). A `go test` green where the precompiled
  form was required is a gap, not a pass.
- **R-2 — A compile-failure red is not a catch.** Where RED is claimed, the failure must be the
  criterion's own assertion failing, not the package failing to build.
- **R-3 — RED-now cells are owed by the run phase, not the plan phase.** The characterization
  test does not exist at plan-phase, so no RED has been measured and none is recorded here. Each
  criterion below carries its green path and the RED it requires; the four-element RED-now cell
  (command / verbatim stdout / exit code / tree SHA) is written into `progress.md §E.2` during
  M1, before the fix lands. **No criterion below is recorded as a pass at this time.**
- **R-4 — Every figure is attributed.** A figure quoted from card t964's verdict is cited as a
  prior measurement and never restated as a fresh one.

## D — AC Matrix

### AC-GGT-001 — Both subprocesses carry GOTOOLCHAIN=local

**Given** `internal/runtime/gobin/resolver.go` after the fix,
**When** its two `exec.Command("go", "env", …)` call sites are read,
**Then** each one sets `cmd.Env` to a slice that contains `GOTOOLCHAIN=local`.

- Verified by: source read of both `goEnvGOBIN` and `goEnvGOPATHBin`, plus
  `grep -c 'GOTOOLCHAIN=local' internal/runtime/gobin/resolver.go` returning `2`.
- Green path: M2. Pre-fix state: `0` (measured in this tree in this run — the package has zero
  occurrences of `GOTOOLCHAIN`, with a positive control confirming the grep form was live).
- Binary: the count is 2 or the criterion fails.

### AC-GGT-002 — The parent environment is preserved

**Given** the fixed resolver,
**When** the `cmd.Env` construction is read,
**Then** it is `append(os.Environ(), "GOTOOLCHAIN=local")` — the parent environment plus the one
override — and NOT a bare literal slice that would strip `HOME`, `PATH`, `GOPATH`, and `GOBIN`.

- Verified by: source read; `grep -c 'os.Environ()' internal/runtime/gobin/resolver.go` returning
  `2`, and the package compiling with `os` imported.
- Green path: M2.
- Why this is a separate criterion: a bare `[]string{"GOTOOLCHAIN=local"}` satisfies AC-GGT-001
  while silently changing what `gobin.Detect` returns. That mutant is exactly what this criterion
  exists to reject.

### AC-GGT-003 — No toolchain is downloaded into a temp HOME

**Given** a test that binds `HOME` to `t.TempDir()`,
**When** `gobin.Detect` is called,
**Then** the glob `<HOME>/go/pkg/mod/golang.org/toolchain@*` is empty — no match at all, or a
directory holding no entries; both are passes.

- Verified by: the new characterization test in `internal/runtime/gobin/`, run via the
  precompiled-binary form (R-1):
  `go test ./internal/runtime/gobin -c -o <path>` then executing `<path>`.
- Green path: M2 flips it. **RED required first** (AC-GGT-004).
- Binary: the glob is empty or the criterion fails.

### AC-GGT-004 — RED observed, through the precompiled form, before the fix

**Given** unfixed `resolver.go`,
**When** the AC-GGT-003 assertion is executed via the precompiled-binary form,
**Then** it FAILS, and the failure is the assertion's own (R-2), not a build error.

- Verified by: the four-element RED-now cell written into `progress.md §E.2` — the exact
  command, its verbatim stdout, its exit code, and the tree SHA it was measured on.
- Ordering: the commit graph is the only witness of "before" (`verification-claim-integrity.md`
  §2.3), so M1's test commit MUST precede M2's fix commit. A test and fix sharing one commit
  leaves this criterion permanently unverifiable.
- Prior measurement (card t964, cited not re-measured): the equivalent precompiled run was
  `--- FAIL (8.07s)`.
- Not observable ⇒ not passed: where the run machine's PATH `go` already satisfies the module's
  `go` directive, no RED can be produced. That is reported as a gap under R-3, never recorded as
  a pass and never substituted with a `go test` green.

### AC-GGT-005 — GREEN in BOTH invocation forms after the fix

**Given** the fixed resolver and the characterization test,
**When** the package is exercised twice — once as the precompiled binary run directly, once as
`go test ./internal/runtime/gobin` —
**Then** BOTH report PASS.

- Verified by: both runs' verbatim output recorded in `progress.md §E.2`.
- Green path: M3. This turns card t964's control pair (precompiled FAIL / `go test` PASS — prior
  measurement) into PASS / PASS.
- Binary: both forms pass, or the criterion fails. A single green form is a gap.

### AC-GGT-006 — vet and lint clean on the changed package

**Given** the fixed package,
**When** `go vet ./internal/runtime/gobin/...` and
`golangci-lint run ./internal/runtime/gobin/...` are run,
**Then** each exits 0 and reports no issue against the changed file.

- Verified by: both commands' verbatim output and exit codes recorded in `progress.md §E.2`.
- Green path: M4.
- Note: when judging formatting, read `gofmt`'s **output** — `gofmt -l` exits 0 while listing
  files, so its exit code is not the verdict.

## D.1 — Traceability

| AC | Requirement | Milestone |
|---|---|---|
| AC-GGT-001 | REQ-GGT-001 | M2 |
| AC-GGT-002 | REQ-GGT-002, REQ-GGT-004 | M2 |
| AC-GGT-003 | REQ-GGT-003 | M1 (test) + M2 (fix) |
| AC-GGT-004 | REQ-GGT-005, REQ-GGT-006 | M1 |
| AC-GGT-005 | REQ-GGT-003, REQ-GGT-006 | M3 |
| AC-GGT-006 | C-2 | M4 |

Every requirement in `spec.md` §2 is covered by at least one criterion. No criterion covers the
out-of-scope R2 (`spec.md` §4), by design.

## D.2 — Mutant probe (adoption check)

Two mutants were considered against this criterion set, per the mutant-probe rule:

- **Mutant A — `cmd.Env = []string{"GOTOOLCHAIN=local"}`** (satisfies the requirement's letter,
  violates its intent by stripping the environment). Caught by AC-GGT-002.
- **Mutant B — the glob asserted against a path that can never match** (a vacuous assertion that
  is green in both the fixed and unfixed tree). Caught by AC-GGT-004: an assertion that has been
  observed red is not vacuous.

## D.3 — Definition of Done

All six criteria pass, the M1-before-M2 commit ordering is visible in the git graph, and every
figure in `progress.md §E.2` is either measured in that run against that tree or explicitly
attributed to card t964's verdict.
