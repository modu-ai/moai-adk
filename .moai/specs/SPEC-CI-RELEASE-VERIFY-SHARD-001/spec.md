---
id: SPEC-CI-RELEASE-VERIFY-SHARD-001
title: "Release Verify sharding — split the 3-OS race run into ci.yml's six shards so no leg reaches the job timeout"
version: "0.1.6"
status: in-progress
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, github-actions, release-verify, sharding, race-detector, multi-os, required-gate, timeout"
tier: M
related_specs: [SPEC-CI-VERDICT-INTEGRITY-001, SPEC-CANDIDATE-CI-001, SPEC-CI-FLAKY-STABILIZE-001]
---

# SPEC-CI-RELEASE-VERIFY-SHARD-001

## A. History

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-19 | manager-spec | Initial draft. Card t1892 (lane-40, run tn1p4e, Class C, P0). Every premise measured on tree `74ad197380` (PR #1830 head); evidence ledger in research.md, adoption cells in acceptance.md. |
| 0.1.1 | 2026-10-19 | manager-spec | Plan-audit iter1 fixes (FAIL 0.78, `.moai/reports/t1892/plan-audit-iter1.md`): REQ-RVS-011 pinned to the exact six shards per OS; REQ-RVS-003 parser edges and Example/Fuzz guard; REQ-RVS-014 added (remainder exclusion asserted); AC-RVS-016 added (shard weight, post-landing); WHY under its own heading; F.4 reworded. |
| 0.1.2 | 2026-10-19 | manager-spec | Plan-audit iter2 fixes (FAIL 0.84, codex gate fail, `.moai/reports/t1892/plan-audit-iter2.md`): REQ-RVS-001 pins the final leg set (no `exclude`, no duplicate `include`); REQ-RVS-003 split into 003 (divergence), 015 (zero/two-match cover) and 016 (Go-discovery guard replacing the over-broad name-prefix ban); section C wording; ledger exit codes. |
| 0.1.3 | 2026-10-19 | manager-spec | Plan-audit iter3 fixes (FAIL 0.84, codex fail, `.moai/reports/t1892/plan-audit-iter3.md`): the cover and discovery checks run over the names `go test -list` prints instead of a re-implemented Go rule; duplicate-include mutant prediction corrected (caught structurally, not by leg count); REQ-RVS-015 defines its own term; constraint D.3 names the guard's REQs. |
| 0.1.4 | 2026-10-19 | manager-spec | Plan-audit iter4 fixes (FAIL 0.88, codex fail, `.moai/reports/t1892/plan-audit-iter4.md`): the name listing is taken under the shards' build conditions (`go test -race -list`), with a race-only-name mutant; stale "parsed from source" wording aligned; nested-list compile weight recorded as residual risk; listing fixture preserved under repro/. |
| 0.1.5 | 2026-10-19 | manager-spec | Plan-audit iter5 fixes (label-only FAIL 0.90, codex fail, `.moai/reports/t1892/plan-audit-iter5.md`): mutant (x) requires a paired plain/`-race` run on one clean-base fixture (race-only name `FuzzRaceOnly`); the "none exists today" aside moved to research.md; `TestMain` listing cost noted. |
| 0.1.6 | 2026-10-19 | manager-spec | Plan-audit iter6 fix (FAIL 0.91, codex fail, `.moai/reports/t1892/plan-audit-iter6.md`): race-pair base no longer contains `ExampleAlpha` (vet `tests` analyzer would fail the build); no `-vet=off` on the race-pair commands; run 1 must exit 0. |

## A.1 Why this SPEC exists

Problem statement: on PR #1830 (head `74ad197380`, run 39180942620) all three `Release Verify` legs (ubuntu, macos, windows) ended `cancelled` at the job-level `timeout-minutes: 40` while inside the single step `Run tests with race detector` (`go test -json -race -timeout 35m ./...`), 00:29Z to 01:09Z. A cancelled leg makes `Release PR Multi-OS Gate` — a required check on `main` — fail, so the release PR cannot merge. The same shape already failed on two earlier `workflow_dispatch` runs (38363720321 on 2026-10-13, 37861821414 on 2026-10-09), and a ubuntu leg that passed on 2026-10-09 did so in 38.2 of its 40 minutes: the margin was never real. `ci.yml` already runs the identical race suite as six independent shards (internal/cli 1-4 of 4, rosterguard, remainder), none of which approaches its ceiling. Release Verify never adopted that split.

## B. Requirements (GEARS)

- **REQ-RVS-001** — **When** the Release Verify job runs, the workflow shall execute the race-detector suite on each of `ubuntu-latest`, `macos-latest`, and `windows-latest` as the same six shards `ci.yml` defines — internal/cli 1 of 4, 2 of 4, 3 of 4, 4 of 4, the rosterguard package, and the remainder packages — so that every test `ci.yml`'s race run covers also runs on every OS. The final leg set shall be exactly the 18 pairs of the three OS values and the six shards, each once: the matrix shall carry no `exclude` entry and no duplicate `include` entry.

- **REQ-RVS-002** — The shard selectors shall partition the top-level test names of `internal/cli` exactly: every name matches exactly one of the four selectors, and the remainder shard shall run every package other than `internal/cli` and the rosterguard package exactly once. No test is excluded and none runs twice on one OS.

- **REQ-RVS-003** — **When** a shard selector or the remainder exclusion list in `ci.yml` differs from the one in the release workflow, a repository test shall fail and name the divergence.

- **REQ-RVS-004** — **When** a shard's test stream contains no test event, that shard shall fail; a selector that matches nothing shall never read as a pass.

- **REQ-RVS-005** — The `Release PR Multi-OS Gate` job shall keep its name, its aggregation logic, and its dependency on the detection job and on the job that carries the OS legs, so that any cancelled, timed-out, failed, or skipped-without-named-exclusion leg still fails the gate.

- **REQ-RVS-006** — The workflow change shall not alter any required status-check context of `main` branch protection. The legs introduced by this SPEC shall not become required contexts, and no run-phase agent shall write branch protection.

- **REQ-RVS-007** — The workflow change shall not raise any timeout above its current value (`-timeout 35m` per test binary, `timeout-minutes: 40` per job), shall not exclude or skip any test, shall not add `continue-on-error`, and shall not add a retry of a failed test.

- **REQ-RVS-008** — The workflow shall use exactly six shards per OS — the same count `ci.yml` uses — shall run each test once per OS per workflow run, and shall run `go vet ./...` once per OS rather than once per shard.

- **REQ-RVS-009** — **When** a shard runs the race detector, it shall run with test-result caching disabled, so a result restored from a prior run's build cache cannot stand in for an observation made in this run.

- **REQ-RVS-010** — The workflow shall preserve the `sg` (ast-grep) install and identity check, `go vet`, the windows-only ancestor-`.moai` diagnostic steps before and after the tests, the per-test census, and the compressed test-event-stream upload; the uploaded artifact name shall be unique for every shard and OS combination.

- **REQ-RVS-011** — **When** `scripts/release.sh --require-matrix-run` evaluates a dispatched Release Verify run, it shall accept the sharded leg names: each of the three OS values must carry exactly the six declared shard ids (`cli-1`, `cli-2`, `cli-3`, `cli-4`, `rosterguard`, `remainder`), once each, every one concluded `success`; a missing shard (on one OS or on all three), an unknown shard name, a duplicated shard job, or a non-success shard shall withhold the release gate with a reason naming the OS and shard.

- **REQ-RVS-012** — The shard command shall run under `bash` on all three OS runners, shall receive its test selector through an environment variable rather than by interpolating it into shell text, and shall keep the existing rule that a missing test-stream artifact on windows is an error.

- **REQ-RVS-013** — **When** the release PR (#1830) is re-run with the change on its head, every Release Verify shard on every OS shall conclude `success` and `Release PR Multi-OS Gate` shall conclude `success` — a post-landing observation made by the leader, not by the lane.

- **REQ-RVS-014** — **When** the remainder shard builds its package list, it shall fail if the list still contains `internal/cli` or the rosterguard package, in addition to failing on an empty list, so a listing that did not remove the two excluded packages cannot run them a second time unnoticed.

- **REQ-RVS-015** — **When** a top-level `internal/cli` test name matches zero or more than one of the four internal/cli selectors, a repository test shall fail and name the test and the selectors involved. A "top-level `internal/cli` test name" is a name beginning `Test` that the go tool prints under the shards' own build conditions, `go test -race -list '.*' ./internal/cli/` (the trailing `ok`/`FAIL` line is not a name); `Benchmark...` names are ignored (benchmarks do not run without `-bench`) and `Fuzz...`/`Example...` names are handled by REQ-RVS-016. The bare name `Test` is such a name and matches none of the four selectors. The listing is taken with `-race` because the shards run with `-race`: a name that exists only in a `//go:build race` file runs in the shards and is listed only under that flag.
- **REQ-RVS-016** — **When** the go tool lists, for `internal/cli` or the rosterguard package, a fuzz target or an example (a name the tool enumerates that is neither a `Test...` nor a `Benchmark...` name), a repository test shall fail and name it, because no `^Test` selector would run it. The set of names is whatever `go test -race -list '.*'` prints for the package (the shards' own build conditions, so a fuzz target or example that exists only in a `//go:build race` file is seen); the SPEC does not restate Go's rules for what counts as a runnable test, fuzz target or example, so a function Go does not list (a lowercase-suffixed `Fuzz...` helper, an example without an output comment) cannot fail the test, and one Go does list (a bare `Fuzz`, a bare `Example`, an example whose output marker is lower-case or in a block comment) cannot be missed.

## C. Acceptance Criteria

Tier M — the Given-When-Then set lives in `acceptance.md` (16 criteria, AC-RVS-001..AC-RVS-016). Each release-blocking criterion carries a RED-now cell pinned to tree `74ad197380` and a green-path cell naming the milestone that flips it; regression-guard criteria carry a baseline (green now) and post-landing criteria a reference observation. The evidence ledger the RED cells cite is research.md §B.

## D. Constraints

1. **No timeout increase, no test exclusion, no gate relaxation, no required-check change** (operator directive for card t1892). Each is a testable requirement above (REQ-RVS-005/006/007) and an acceptance criterion.
2. **Cost.** Shard count equals `ci.yml`'s (six). The organization stated a 10 USD/month Actions budget; the repository is public (research.md E3), so hosted standard-runner minutes are not billed — recorded as a discrepancy for the leader, not as a licence to add shards or duplicate runs.
3. **Implementation scope (run phase).** `.github/workflows/release-pr-multi-os.yml`; `scripts/release.sh` and `internal/template/release_script_test.go` (leg lookup, REQ-RVS-011); one new guard test under `internal/template/` (REQ-RVS-003, 015, 016 and the matrix-shape, artifact-name and vet-leg assertions of AC-RVS-001, 007, 013). `.github/workflows/ci.yml` is read, never edited. Repository-local only; no template-tree mirror is involved (none of these files has one).
4. **Windows and macOS legs are unmeasured for shard weight.** Per-shard proxies exist for ubuntu only (`ci.yml` comment, run 37861953651). The first full run of the new structure is the measurement (REQ-RVS-013); the SPEC does not pre-assert a per-OS duration.
5. **Plan-phase boundary.** This phase writes only `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/`. No workflow, script, or test file is touched here.

## E. Out of Scope

### Out of Scope — editing ci.yml

- `.github/workflows/ci.yml` is read as the source of the shard configuration and is not modified; deduplicating selectors into a shared source consumed by both workflows is a follow-up (decision-index Q3).
- The `Test (ubuntu-latest)` job and the guard-family run in `ci.yml` are unrelated to the race shards and are untouched.
- The comment at `ci.yml` ~L265-267 says release PRs run `go test -race ./...` through the release workflow's ubuntu leg; it goes stale after this split (the leg runs the race suite as shards). Recorded as a follow-up for the card that edits `ci.yml`; not fixed here.

### Out of Scope — required checks and the gate

- Any change to the required contexts of `main` protection (`Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)`, `Release PR Multi-OS Gate`) and any change to the gate job's text.
- The `detect-release` job, the shared go-code paths filter, and the docs-only skip behaviour.

### Out of Scope — runtime budget and test health

- Raising any timeout, rebalancing shards by measured per-OS weight, and repairing slow or flaky tests. A shard that exceeds its ceiling on windows or macos after the split is reported by the first live run and handled by a follow-up card.
- GitHub plan concurrency (macos is capped at 5 concurrent jobs on the Team plan), billing, and runner-class changes.

### Out of Scope — retry

- Adopting `scripts/ci/retry-flaky.sh` on release legs (decision-index Q5).

## F. Success Criteria

1. The Release Verify structure runs 18 shard legs (3 OS x 6 shards) and every test `ci.yml` covers runs once on each OS, shown by the exact-cover listing, the parity test, and the empty-sweep refusal flipping from their RED cells.
2. The gate job, the required-check names, and every timeout are byte-for-byte or value-for-value unchanged against base `74ad197380`.
3. `actionlint` and the required-checks validator stay clean; `scripts/release.sh --require-matrix-run` passes its extended scenario table.
4. After landing, the leader's re-run of PR #1830 shows every shard green and the gate green (REQ-RVS-013), and the per-shard durations are recorded (AC-RVS-016). These post-landing observations (AC-RVS-014, AC-RVS-016, and the windows half of AC-RVS-012) cannot be closed inside the lane.

## G. Residual risk (explicit)

The parity test (REQ-RVS-003/015/016) shells out to `go test -race -list '.*'` for `internal/cli` and for the rosterguard package, which compiles two race test binaries inside whichever job runs `internal/template` tests — on the release workflow that is the `remainder` shard on all three OS values, and `Test (ubuntu-latest)` on ordinary PRs. That compile weight is unmeasured — and `internal/cli`'s `TestMain` setup probably also runs on every nested listing, since `go test -list` executes the test binary (an inference from how the tool works, not measured here) — and adds to the unmeasured macOS/Windows shard weight below; it is a run-phase measurement item (time the test on this machine under the resource-entry rules, then read the remainder-shard duration on the first live run, AC-RVS-016), not done at plan time. If it pushes the remainder shard over its ceiling, the response is the same: a rebalance card, never a timeout raise.

macOS and Windows shard weight is unmeasured. By this SPEC's own arithmetic (research.md §C.3) the heaviest shard (rosterguard, 782 s on ubuntu) lands at roughly 26-39 minutes on a 2x-3x slower OS against the unchanged 35-minute test ceiling, so the first live run can still fail for the same reason as the original incident. The SPEC does not change shard count or ceilings in response. An overrun is handled by a rebalance card (split the overrunning shard by selector), never a timeout raise (decision-index Q8); the first live run is the measurement (AC-RVS-016). Flagged for the leader.
