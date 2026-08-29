# SPEC-CI-TEST-OBSERVABILITY-001 — Implementation Plan

> Tier **M** (`tier: M` in `spec.md` frontmatter — prose and frontmatter agree). Justified against the actual scope: **four** `go test` call sites across two workflow files, a new census implementation with its own fixture-based check, and a positive-control CI observation carrying named debt. Tier S is a single-surface change with ACs inline in `spec.md §3`; this has four call sites, a new executable artifact, and an 8-criterion acceptance set that must carry external run evidence — so S is too small. It is not L: no data model, no new Go production path, no design space left open. Milestones below are ordered by **decision-reversibility**: the census output contract (§M1) is the decision most likely to change under review and leads; mechanical replication across the remaining call sites (§M3) is deferred; the positive-control CI observation (§M4) is the terminal milestone by construction — it cannot run until everything before it has landed.

## A. Context

Card **t358**, branch `WT-ci-test-observability`, base `origin/develop c6aa61346`. Problem statement, measured figures, and the chosen approach are in `spec.md` §A–§C and are **not re-derived here**.

Surface: 2 workflow files (**4** invocation sites — `ci.yml:183`, `ci.yml:238`, `ci.yml:329`, `release-pr-multi-os.yml:189`) + 1 census implementation. No Go production code, no data model, no user-facing UX.

## B. Known issues / constraints carried into implementation

1. **Exit-code preservation under `set -e` — read this before writing the step.** The `shell: bash` key in GitHub Actions resolves to a verbatim shell string, observed on a real run of this repo's `Race Test` job (run `33173944485`, job `98857764037`, step header line 3):

   ```
   shell: /usr/bin/bash --noprofile --norc -e -o pipefail {0}
   ```

   The path shown is the Linux runner's; the Windows runner resolves a different `bash` path with the **identical flag set**, so nothing below is Linux-specific.

   **`-e` is present, not only `-o pipefail`.** Therefore:

   - **WRONG — `go test ... > f; rc=$?; ...; exit $rc`.** Under `-e` the shell exits AT the failing `go test`. The `rc=$?` never runs, the census never prints, and the job dies with no per-test evidence — *precisely the failure this SPEC exists to close, reproduced inside its own implementation.* An earlier draft of this plan prescribed exactly this form.
   - **WRONG — piping into `tee` to dodge `-e`.** That makes the step status the last pipeline element's, which is the same masking failure in a different costume.
   - **CORRECT — suppress `-e` for the one command whose failure is expected:**

     ```bash
     rc=0
     go test -json ... > "$STREAM" || rc=$?
     <census over "$STREAM">
     exit $rc
     ```

     The `||` places the command in a context where `-e` does not fire, so `rc` is captured and the census still runs. An explicit `set +e` / `set -e` bracket around the invocation is equally acceptable.

   Do not "simplify" the `|| rc=$?` back to `; rc=$?`: it looks equivalent and is not, and the difference is invisible on a green run — it appears only on the red run, where the evidence matters most.
2. **Build diagnostics are INSIDE the stream, and stderr is empty — do not restore the old rationale.** Measured (`spec.md` §A.1): for a package with a syntax error, `go test -json ./...` produced rc=1 with **stdout 666 bytes and stderr 0 bytes**, the compiler diagnostic carried in `build-output` events terminated by `build-fail`. An earlier draft of this plan said the opposite — that stderr carries build errors and a stdout-only redirect keeps them visible. **That was false.**

   What follows:

   - Redirect **stdout only** — still correct, but for stream integrity (no non-JSON text in the file), NOT for console visibility. `2>&1` would corrupt the stream.
   - A bare `> file` redirect **swallows build errors from the console**. The census's `BUILD FAILED` row is what puts them back, and it is **load-bearing**. Do not delete it as redundant: a green run cannot contradict its removal, so the loss would surface only on the red run where it matters — the same hazard shape as `; rc=$?` in §B.1 above.
3. **Windows legs (two of the four sites).** `ci.yml:329` (`test-integration`) and `release-pr-multi-os.yml:189` (`full-matrix-test`) both run a 3-OS matrix including `windows-latest`, so the summary implementation must be portable — no GNU-specific text tooling. `jq` is already used in four workflows in this repo, so it is an **established** dependency and is the default choice rather than a new one to justify.
4. **One predicate, two labels.** `Action=="skip"` catches both forms of "did not run"; the presence or absence of the `Test` field labels which (`spec.md` §C.1). The census must not be built as two independent detections — that was an earlier framing and is superseded.
5. **Required check names are name-matched by branch protection.** `Test (ubuntu-latest)` and its skip-marker twin (`ci.yml` `test-skip-marker`) must keep emitting the same names. Adding steps inside the job is safe; renaming the job is not.

## C. Pre-flight

- Read `ci.yml:150–250` and `release-pr-multi-os.yml:170–200` before editing; the surrounding comments are load-bearing (`shell: bash` rationale at `ci.yml:157`, the advisory-by-name rationale for `Race Test`).
- Confirm the card branch has no other in-flight writer.

## D. Milestones

### M1 — Census contract and summary implementation (highest change-likelihood — review this first)

Define, and implement once, what the summary step prints. Proposed contract:

```
=== test census ===
SKIPPED TEST  <package>  <TestName>     # Action==skip, .Test present
NOTHING RAN   <package>                 # Action==skip, .Test absent
FAILED        <package>  <TestName>     # Action==fail
  <captured output for that test>
=== totals: N packages, N tests run, N skipped, N packages with nothing run, N failed ===
```

Both non-run rows come from **one** `Action=="skip"` pass, split on `.Test` (§B.4). Expected baseline for `NOTHING RAN` on a full-suite run of this tree: the three zero-test-file packages `cmd/moai`, `internal/template/scripts`, `scripts/convert-nextra-to-hextra` — a census reporting fewer than three has a bug, and one reporting more has found something worth reading.

Deliverables: the summary implementation (script or inline step), plus its own unit-level check against a **fixture JSON stream** so the census logic is testable without a CI run.

Decisions open to review here, and only here: the exact line format, whether totals are emitted, and the row labels.

### M2 — Wire the required `test` job (`ci.yml:183`)

Change the invocation to write `-json` to a file, keep `-coverprofile`/`-covermode` unchanged (measured compatible on two package sets, `spec.md` §A.1), capture `rc`, run the M1 summary, `gzip` the stream, and upload with `actions/upload-artifact@v7` and `retention-days: 7` — the house convention already at `ci.yml:433-438` — **plus `if: always()`, which is this SPEC's own addition and is not part of that convention** (REQ-CTO-008 needs the artifact on a red run; the existing convention does not carry it).

Codecov reads the `coverage.out` **file**, not stdout, so moving the `coverage: NN% of statements` lines into JSON `output` events does not affect the upload. Confirm this by observation, not by reading the YAML.

Terminal check for this milestone: the job's console log does **not** grow by the raw stream, and a deliberately failing test still prints its failure text.

### M3 — Replicate across the remaining three call sites (mechanical)

`ci.yml:238` (`test-race`), `ci.yml:329` (`test-integration`, 3-OS matrix), and `release-pr-multi-os.yml:189` (`full-matrix-test`, 3-OS matrix). Same shape as M2, minus coverage.

`ci.yml:329` was missed in the first draft of this SPEC and is in scope by lead ruling — identical defect, identical one-line change. It runs `-tags=integration` against `./test/integration/harness/...` only, so its census covers that one package tree; that is expected, not a bug.

Artifact names must be distinct per job **and** per matrix OS — the two matrix jobs produce three legs each, and a shared name collides on upload.

Optionally in this milestone: add `workflow_dispatch` to `ci.yml` as an enabling change for **future** cards, recording in the commit message that it does not become dispatchable until it lands on the default branch.

### M4 — Positive-control observation (terminal)

Push the card branch — a deliberate, minimal exception to the lane protocol, whose basis is the lead's dispatch approval, since a dispatch resolves the workflow at a **remote** ref. The exception is **CI-inert**: `ci.yml` carries no dispatch trigger and fires only on push to `[main, develop]` and PR to `[main]`, so pushing this branch starts no run and consumes no runner. No PR is opened. Then dispatch `release-pr-multi-os.yml` against it (**lead-approved**), and read the resulting run's artifacts and console log. The control is one of the two **already-existing** real skips measured in `spec.md` §A.1 — `TestProfilePhaseDistributions` or `TestReadOAuthToken_Keychain` in `internal/statusline`. Prefer an existing skip; plant nothing.

[HARD] **The dispatch is ONE run.** If it fails, the cause is established before any re-run — re-executing until a pass appears is prohibited, and a pass obtained that way is not evidence. Everything AC-CTO-007 needs must be obtainable from a single run's artifacts, which is why M1-M3 carry their own local verification: the dispatch confirms, it does not debug.

Record: run URL / job id, the verbatim census lines naming the skipping test, and the artifact name. If a planted mutant proves unavoidable, the milestone is not complete until the mutant is reverted and that revert is shown.

Fallback path if dispatch is unavailable: read the `ci.yml` run triggered by the push of `develop` after this card integrates. This path is **post-merge**, and choosing it must be stated as such rather than presented as a pre-merge observation.

### M4.1 — Delete the remote card branch (ordering is load-bearing)

The push in M4 exists only to give the dispatch a remote ref, so the ref is removed once it has served that purpose: `git push origin --delete WT-ci-test-observability`.

[HARD] **This runs only AFTER** the dispatch run's artifacts and console log have been read **and** AC-CTO-007's four items are recorded in `progress.md`. Deleting the ref earlier would strand the observation — the run's artifacts are what the AC cites, and the branch is the only thing tying the run to this card's work. Delete-then-read is not recoverable by re-pushing, because the ONE-run constraint forbids a second dispatch.

The local branch and its worktree are NOT touched here: the card's work is unmerged, so the worktree still holds the only copy (`AGENTS.md` §3).

### M4.2 — Record the two debts explicitly

Two criteria cannot be had from one run, and the ONE-run constraint forbids manufacturing a second. Before close, `progress.md` records both as **named debt**, not as passes:

1. **CI-level red-path confirmation** (AC-CTO-003) — the shell semantics are verified locally against the exact converted step body; the CI-level red path is not exercised pre-merge.
2. **Artifact-on-failure** (AC-CTO-005b) — `if: always()` is declared and the artifact is verified re-parseable on the green run; "present on a FAILED run" is not observed.

Both discharge on the first genuinely red CI run after this lands. If the fallback observation path is taken, AC-CTO-007 closes post-merge and that is recorded as debt too (`spec.md` §I).

## E. Self-verification

Every claim in the closing report carries the command that produced it and that command's verbatim output. Specifically: the console-size claim is a byte/line count of the actual CI log, not an estimate; the exit-code-preservation claim is an observed red run, not a reading of the YAML.

## F. Risks

| Risk | Mitigation |
|---|---|
| Redirect hides a failure that used to be console-visible | M2's terminal check is a deliberately-failing run, not a passing one |
| `pipefail` / `tee` masks a red suite | explicit `rc` capture (§B.1); reviewed in M1 |
| Windows leg breaks on the summary implementation | portability constraint stated in §B.3; the 3-OS matrix is the observation vehicle, so a break surfaces in M4 |
| Artifact-name collision across matrix legs | per-OS artifact names in M3 |
| Census tooling (`jq`, `gzip`, `sed`, `sort`) unverified on windows/macOS runners | named Gap in `spec.md` §J — all four existing `jq` workflows are ubuntu-only, so the dispatch is first exposure; a tooling failure on a non-ubuntu leg is a diagnosis to establish, not a re-run to attempt |
| `test-stream.json` visible in the repo root mid-run to a filesystem-walking test | residual risk in `spec.md` §J — gitignored, so git-based checks are unaffected; unmeasured, mitigation is to write outside the repo root |
| Full-suite stream size is unknown (extrapolated) | gzip measured at ~15.5×; if the artifact proves oversized, the fallback is per-package artifacts, not dropping the census |

## G. Anti-patterns

- Asserting the workflow YAML contains `-json` and calling the SPEC done. That is the failure mode `spec.md` §D names explicitly.
- Emitting only a skip **count**.
- Reporting the fallback post-merge observation as if it were the pre-merge one.

## H. Cross-references

- `spec.md` §F (call sites), §G (observation vehicle), §H (exclusions).
- `acceptance.md` AC-CTO-007 — the positive-control criterion M4 discharges.
