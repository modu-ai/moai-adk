# Internal: check-rules

When to read: in RUN step 5, before you decide which checks you run yourself and whether a remote run may stand in for one, and again from the failure-table row for a red check.

The receipt is RUN step 7 and the wide set belongs to the lander; neither is restated here. The tier line of a head and the floor list of paths stand only in the audit contract's section "Audit tiers".

## Who runs which check

| Role and change | Run locally | May cite a remote run |
| --- | --- | --- |
| Author: new or changed code | every check of each language and suite the change touches, on the exact tree | no for what the change touches; a remote run there is extra evidence, never a replacement |
| Author: a language or suite the change does not touch (for example the Rust suite of a change that touches only documents) | nothing for that suite; record its local run as `not_run` | yes, under the citation conditions |
| Repair lane: the repair changes code | every check on the repair head | no |
| Auditor: the audited head changes code | what the tier line of the head says: the lane's receipt is cited, the head is not run again, no mutant | no |
| Delta auditor: merge-only delta (the merge-tree result equals the pushed tree and the lane changed nothing itself) | only the tests that target the files the merge brought in; the local full suite is `not_run` | yes, under the citation conditions |
| Author: a skill package file changes | the file's sha256 entry refreshed in the package's source manifest in the same commit (`shasum -a 256 <file>`), then the package's manifest and route tests | no for the package tests; the rest follows the rows above |

## Citing a remote run

Only a release pull request has a `pull_request` run to cite; a lane opens no release pull request (`flows/next-flow`), so a lane's head has none. Cite a run only when all three hold:

1. The run's `head_sha` equals the release pull request's head exactly (`gh api "repos/<owner>/<repo>/actions/runs?head_sha=<head>"`). A `pull_request` run reports that head in `head_sha` and builds the merge ref, whose SHA is the run's `GITHUB_SHA`.
2. Every required job of the run is `completed` with conclusion `success`; take the job names from the workflow file, not from memory.
3. The auditor reads the run back itself and writes one line in the verdict, `ci: ci run=<run id> head=<40-hex head SHA>`, with the run's conclusion. An author's line in that form is also accepted; no verdict waits for one. `moai pr-self-audit-check` passes a release pull request body with or without it and cannot see GitHub.

Where each target runs: a release pull request runs `ci / linux-x86_64` only. A push to the branch it targets runs `ci / linux-x86_64`, `ci / macos-arm64` and `windows-build` (build only) with the same check set, the three supported platforms. The scope moves to after the push; it does not shrink. A failed run after a push never blocks a landing; it becomes one line of `docs/later.md` (`flows/next-flow`).

Never cite a remote run for a language or suite the change touches, when the repair changed code in it, when no run exists for the head (a `pull_request` run does not start while the release pull request has a merge conflict: write `ci: not_run` with the reason), or for checks it does not perform (mutations, behavioural probes). Citing never relaxes a gate: the independent audit, `ci` success with no failed `windows-build` on the head, and exact head and base equality stay required.

A docs-only release pull request gets light CI: when every changed path is under a verified-safe doc root, the heavy steps (installs, toolchain setup, build, test) are skipped and only `git diff --check` runs. The aggregate `ci` check, the verdict's CI citation and every condition above stay unchanged. The project's own CI decides its scope; the safe roots of the MoAI repository are `docs/` files ending `.md`, `.txt`, `.png` or `.svg`, minus files tests read. A failure or an unrecognized path fails safe to the full scope.

## Tests

- No wall-clock assertions: inject the clock or the time budget (a clock parameter or a budget flag) and fix it in the test. An auditor treats a direct wall-clock budget in a test assertion as a finding.
- Bound a child process in a test by progress (an invocation count or a log signal that resets the deadline) plus a hard hang ceiling and an invocation cap, never a bare wall-clock timeout.
- A test that fails on one run and passes when the same head is run again is flaky, not fixed. Classify from evidence only: the same `head_sha`, the same job, one failure and one later success, with the test's name and message quoted from the log. `moai flake-classify --failed <log> --later <log>` (saved `gh run view <id> --log` texts) sorts each test that failed in the first log into `flaky`, `still_failing` or `not_in_later`; it reads each job's head from the log, so a pass on another head is `head_mismatch`, never flaky.
- A failure that a re-run on the same SHA turned green does not stop the merge and is not forgotten. Open one card per test: `moai card add --kind task --title "flaky: <test>" --ac "the flake is fixed, or the test is retired with its reason"`; when one is open, add a `moai card note <card> --kind report` note instead (text on stdin). Record the test, the failed run id, the passing run id and the SHA. If the card cannot be written, put those four facts in the merge comment and say the card is missing.
- A re-run's pass is never cited as the verification of a change that touches that test or its fixtures. A repair of a flaky test never adds a sleep, a retry or a skip to make it pass: find its root cause and fix that. A test whose root cause is not found is deleted, with the reason written in the card and the SPEC.
- Tests run in parallel, one process each, so a test never shares a fixed path with another process: no `std::env::temp_dir().join("<fixed name>")` or `CARGO_TARGET_TMPDIR` plus a fixed name used as a folder, HOME or file the test or the binary may write. Take a fresh temporary folder per test. The only fixed name allowed is one that must stay absent, and it says so with `no-such` in the name.
- A test that runs `cargo` on the shared build folder re-copies the binary under test while other tests run it; such a test runs alone under cargo-nextest, with no other test running beside it.
- Add a test to a file of its own under `tests/` (or to the file that already tests the same module), not to a shared catch-all such as `tests/product_cli.rs`: two changes that both append there overlap, and the later one needs a merge at landing and a delta re-audit.
