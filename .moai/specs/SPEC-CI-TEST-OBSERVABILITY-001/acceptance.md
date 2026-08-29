# SPEC-CI-TEST-OBSERVABILITY-001 — Acceptance Criteria

Each AC is binary-testable and stated as Given / When / Then. **AC-CTO-007 is the positive control and the terminal AC** — no other AC closes this SPEC.

## D. AC matrix

| AC | Covers | Verified by | Closes |
|---|---|---|---|
| AC-CTO-001 | REQ-CTO-001, REQ-CTO-003 | CI console log byte/line count | pre-close |
| AC-CTO-002 | REQ-CTO-002, REQ-CTO-006 | census output against a fixture stream | pre-close |
| AC-CTO-003 | REQ-CTO-004, REQ-CTO-005 | **local** execution of the exact step body, failing-test tree | pre-close (CI-level red path = DEBT) |
| AC-CTO-003b | REQ-CTO-011 | **local** execution of the same body, broken-build tree | pre-close |
| AC-CTO-004 | REQ-CTO-007 | Codecov step + `coverage.out` presence | pre-close |
| AC-CTO-005a | REQ-CTO-008 | `if: always()` declared + artifact re-parses, on the green dispatch | pre-close |
| AC-CTO-005b | REQ-CTO-008 | artifact present on a FAILED run | **DEBT** — unobtainable from one run |
| AC-CTO-006 | REQ-CTO-009 | diff of the job `name:` value set | pre-close |
| **AC-CTO-007** | **REQ-CTO-010** | **a real CI run naming a real skip** | pre-close (post-merge if fallback taken = DEBT) |
| AC-CTO-008 | REQ-CTO-001, REQ-CTO-005, REQ-CTO-008 (all four sites) | grep across the four call sites | pre-close |

The two DEBT rows exist because the lead's ONE-run constraint is [HARD] and both need a *second*, deliberately-red run. They are recorded as debt rather than asserted, and discharge on the first genuinely red CI run after this lands (`spec.md` §I).

---

### AC-CTO-001 — the stream goes to a file, not the console

**Given** a CI job whose test step has been converted to `-json`,
**When** the job completes on a green run,
**Then** the raw JSON stream appears in **no** console line of the job log (zero lines matching `^{"Time":`), and the job's console log length is within the same order of magnitude as the pre-change log — not the ~913× inflation `-v` was measured to cause.

Evidence: the job log fetched and counted (bytes and lines), quoted with the command that fetched it.

### AC-CTO-002 — one predicate, two labels

**Given** a fixture `-json` stream containing (a) a test that called `t.Skip` and (b) a package with no test files,
**When** the summary implementation is run against it,
**Then** a **single** `Action=="skip"` pass catches both, and the output labels them **distinctly** by the presence or absence of the `Test` field — a single undifferentiated "skipped" bucket fails this AC, and so does an implementation that runs two independent detections.

Mutant probe (must fail the criterion): a census that filters `Action=="skip" and .Test != null` only. It passes a naive "names the skipped test" check while reporting nothing for the no-test-files package — so the criterion asserts BOTH labels, not just the first.

Evidence: fixture file path, the command, and the verbatim census output showing both labels.

### AC-CTO-003 — a red run stays red and stays readable (verified LOCALLY, by design)

What this AC tests is **shell semantics**, not CI infrastructure: whether `-e` kills the step before the census, whether the exit code survives, and whether non-JSON text stayed out of the stream. All three are reproducible on a developer machine, and none of them need a CI runner. The ONE-run constraint on the dispatch therefore does not weaken this AC — it relocates it.

Baseline already observed on this tree: a path typo (`./internal/version/...`, not a package here) produced rc=1 with `Action=="fail"` in the stream and the verbatim body `FAIL\t./internal/version/... [setup failed]`.

**Given** the **exact `run:` body** of a converted step, extracted verbatim from the workflow file, and a tree containing one deliberately failing test,
**When** that body is executed locally under the same shell string CI uses — `bash --noprofile --norc -e -o pipefail` (observed at run `33173944485`, job `98857764037`) —
**Then** all three hold:

1. the census **still prints** — i.e. `-e` did not kill the step at the failing `go test` (the `; rc=$?` form fails here; `|| rc=$?` passes);
2. the final exit status is **non-zero**;
3. the printed output contains the failing test's name and its captured output.

Evidence: the extracted step body, the exact invocation used, and its verbatim output. The deliberately failing test is reverted before close, and the revert is shown.

**Named debt (not a pass):** the CI-level confirmation of the red path is **not** obtained pre-merge, because it would require a second dispatch. Recorded in `progress.md` per `plan.md` M4.2; discharges on the first genuinely red CI run after this lands.

### AC-CTO-003b — the stream stays clean and the build failure stays visible

A **compiling** tree carrying a failing test (AC-CTO-003) and a **non-compiling** tree are mutually exclusive states, so this clause gets its own tree and its own execution rather than riding on AC-CTO-003's `Given`. Both are local, and they run in sequence.

**Given** the same extracted `run:` body, and a tree with a deliberately broken build (e.g. a syntax error in one package),
**When** that body is executed locally under the same shell string,
**Then** the build error appears **on the console** — via the census `BUILD FAILED` row, since stderr is empty and carries nothing (`spec.md` §A.1); the stream file contains **no raw non-JSON text**; and the stream file remains parseable by the AC-CTO-002 census.

(The pass criteria are unchanged from the version this SPEC was audited on — only the stated *mechanism* of console visibility is corrected. The implementer's `BUILD FAILED` row is what satisfies clause 1; it was never stderr.)

A `2>&1 > f` implementation fails this criterion — that is the reason it is stated separately rather than folded into AC-CTO-003.

Evidence: the invocation, the console output showing the build error, and the stream file's contents. The broken build is reverted before close, and the revert is shown.

### AC-CTO-004 — coverage is unaffected

**Given** the required `test` job with `-json` added,
**When** the job runs,
**Then** `coverage.out` is produced with `mode: atomic` and the Codecov upload step reports the same outcome class as before the change.

Baseline already measured on this tree: `go test -count=1 -json -coverprofile=… -covermode=atomic ./internal/hook/mx/complexity/` → rc=0, `coverage.out` written, `mode: atomic`, 47 lines.

### AC-CTO-005a — the artifact exists, re-parses, and is declared unconditional

**Given** the single green dispatch run,
**When** the run finishes,
**Then** all three hold: the compressed stream artifact is present in that run's artifact list; it is uploaded via `actions/upload-artifact@v7` with `retention-days: 7` (the house convention at `ci.yml:433-438`) and an `if: always()` condition **present in the workflow file**; and downloading it yields a stream the AC-CTO-002 census parses without error.

Evidence: artifact name + size from the run, the quoted `if: always()` line, and the census re-run against the downloaded stream. The measured size is also the first real full-suite figure — record it and mark the `spec.md` §A.1 extrapolation superseded at that point.

### AC-CTO-005b — the artifact is uploaded on a FAILED run (DEBT, not asserted)

**Given** a CI run whose conclusion is failure,
**When** the run finishes,
**Then** the artifact is nonetheless present in its artifact list.

**This is NOT satisfiable from the single approved dispatch**, and this SPEC does not claim it. `if: always()` being declared (AC-CTO-005a) is evidence of *intent*, not of behaviour — an unobserved claim if reported as a pass. Recorded as named debt in `progress.md`; discharges on the first genuinely red CI run after this lands. Manufacturing a second dispatch to close it is prohibited.

### AC-CTO-006 — required check names are unchanged (verified by diff, no PR)

Under `.claude/rules/local/gitflow-lane-protocol.md` §1 / CLAUDE.local.md §4.1 this repo opens **no card PRs** — cards merge into `develop` locally. An AC gated on "opened as a PR" would have no green path, so this AC is stated mechanically instead: required check names are determined by the `name:` fields of the jobs, which are readable from the file.

**Given** the base and head versions of `ci.yml` and `release-pr-multi-os.yml`,
**When** the set of job `name:` values is extracted from each and compared,
**Then** the two sets are **identical** — in particular the four names carried by jobs this SPEC edits (`Test (${{ matrix.os }})`, `Race Test`, `Integration Tests (${{ matrix.os }})`, `Release Verify (${{ matrix.os }})`) plus the untouched required gate `Release PR Multi-OS Gate` are unchanged in spelling —
**And** the `test` ↔ `test-skip-marker` mutual-exclusivity invariant is intact: their `if:` conditions remain strict complements, so exactly one always reports `Test (ubuntu-latest)`.

Evidence: the two extracted name sets and the diff between them (expected: empty), plus the two quoted `if:` conditions.

### AC-CTO-007 — POSITIVE CONTROL (terminal AC)

[HARD] **This AC is satisfiable from ONE run.** The lead-approved dispatch is a single run: if it fails, the cause is established before any re-run. Re-executing until a pass appears is prohibited, and a pass obtained that way does not satisfy this AC.

**Given** a **single** real CI run of the converted workflow, executing a test that **genuinely skips** — preferring an already-existing skip measured on this tree (`TestProfilePhaseDistributions` or `TestReadOAuthToken_Keychain`, both in `internal/statusline`) over any planted mutant,
**When** that run completes and its artifacts and console log are read,
**Then** the census **names that test** as skipped, and the following are recorded in `progress.md`:

1. the run URL **and** job id,
2. the **verbatim census line(s)** naming the skipping test,
3. the artifact name from which (or the console section in which) the identification was made,
4. which of the two observation paths was used — dispatched `release-pr-multi-os.yml` against the pushed card branch (pre-merge), or the `ci.yml` run on the `develop` push (post-merge) — stated plainly, never conflated.

**Failing forms of this AC, named so they cannot be argued into a pass:**

- asserting only that the workflow YAML contains `-json`;
- citing a local `go test -json` run instead of a CI run;
- citing a run in which no test actually skipped;
- reporting a skip **count** without the test name;
- if a mutant was planted despite the preference against it, closing without showing the revert;
- obtaining the pass by re-running the dispatch after a failure, instead of establishing the cause first.

### AC-CTO-008 — replication is complete

**Given** the **four** call sites named in `spec.md` §F — `ci.yml:183`, `ci.yml:238`, `ci.yml:329`, `release-pr-multi-os.yml:189` —
**When** each is inspected after the change,
**Then** all four write a `-json` stream to a file, capture the exit status **in a form that survives `-e`** (`|| rc=$?` or an explicit `set +e` bracket — never `; rc=$?`), re-raise it, run the census, and upload an artifact under a name unique to the job **and** to the matrix OS for the two 3-OS jobs (`ci.yml:329`, `release-pr-multi-os.yml:189`), which produce three legs each.

**And** the four already-verbose `-run`-scoped invocations (`lsel-leak-guard.yaml:37` plus the three steps in `template-neutrality-check.yaml`) are **unchanged** — they carry `-v` today, so they have no observability gap and converting them is out of scope (`spec.md` §H).

Evidence: the four step bodies quoted, the artifact-name list showing no collision across all eight legs, and a diff confirming the four excluded `-v`-carrying steps are untouched.

---

## Definition of Done

- AC-CTO-001, -002, -003, -003b, -004, -005a, -006, -008 all PASS pre-close, each with its command and verbatim output cited.
- AC-CTO-007 PASSES pre-close **on the dispatch path**; on the fallback path it closes post-merge as debt (see below, and `spec.md` §I). It is not unconditionally pre-close.
- AC-CTO-005b and the CI-level red-path clause of AC-CTO-003 are recorded as **named debt** in `progress.md` — never reported as passes, and never closed by manufacturing a second dispatch.
- If the fallback observation path was taken, AC-CTO-007 closed **post-merge** and that is recorded as debt, stated as post-merge rather than presented as a pre-merge observation.
- AC-CTO-007 recorded in `progress.md` with the four items above.
- Any deliberately failing test (AC-CTO-003) and any planted skip are reverted, with the revert shown.
- No branch-protection required check name changed.
- The positive control was obtained from ONE dispatch run.
- The out-of-scope items in `spec.md` §H remain untouched — in particular no change to `ci.yml:30-32` concurrency.
