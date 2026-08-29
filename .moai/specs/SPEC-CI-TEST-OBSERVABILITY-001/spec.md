---
id: SPEC-CI-TEST-OBSERVABILITY-001
title: "CI per-test evidence: make a skipped test identifiable from CI artifacts alone"
version: "0.1.0"
status: in-progress
created: 2026-08-28
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: ".github/workflows"
lifecycle: spec-anchored
tier: M
tags: "ci, observability, go-test, skip-census, evidence"
---

## HISTORY

### v0.1.0 (2026-08-28)
- Authored from card **t358** (branch `WT-ci-test-observability`, base `origin/develop c6aa61346`).
- Originating observation (lane-2, a different card): CI job `98707120392` log had **0 hits** for the string `TestSpawnLaunch` — a test believed to have run left no trace in the CI artifacts.
- Approach was already decided from measurement before this SPEC was authored; this document records the decision, it does not re-open it.

---

## A. Problem

CI runs `go test` without `-v` or `-json`, so CI artifacts carry **no per-test evidence**. `rc=0` therefore cannot distinguish three different states:

1. the test ran and passed,
2. the test was skipped (`t.Skip`),
3. the selector matched nothing / the package had no tests.

### A.1 Measured evidence (this tree, base `c6aa61346`)

| Measurement | Command | Observed |
|---|---|---|
| Plain run hides skips | `go test -count=1 ./internal/statusline/...` | rc=0; output is exactly `ok  github.com/modu-ai/moai-adk/internal/statusline  14.394s` — **61 bytes, 1 line**. Two tests actually SKIPPED in that run (`TestProfilePhaseDistributions`, `TestReadOAuthToken_Keychain`) left **zero trace**. |
| JSON run names them | `go test -count=1 -json ./internal/statusline/...` piped to `jq -r 'select(.Action=="skip" and .Test!=null) \| "SKIP \(.Package)  \(.Test)"' \| sort -u` | 785,790 bytes of stream; census names **both** skipped tests exactly. |
| Console cost of `-v` | `internal/config` | `-v` → 173,376 bytes / 3,210 lines = **913× plain bytes**. |
| Console cost of `-json` | `internal/config` | `-json` → 1,178,655 bytes / 5,756 lines. |
| CPU cost of `-json` | `internal/config`, consecutive runs | plain wall 1.73s & 1.87s, user CPU 1.82–1.83s; `-json` wall 2.19s & 1.94s, user CPU 1.83–1.85s. **Identical user CPU** ⇒ test-body execution cost unchanged; wall ranges overlap ⇒ the delta is indistinguishable from noise. |
| Compressibility | gzip on the two JSON streams | 785,790→50,589 (**15.5×**) and 1,178,655→74,651 (**15.8×**). |
| Scale | this tree | 135 packages, 11,124 top-level test funcs, 631 `t.Skip` call sites. |
| Coverage compatibility | `go test -count=1 -json -coverprofile=… -covermode=atomic` on `./internal/hook/mx/complexity/` and on `./pkg/...` | rc=0; `coverage.out` written correctly in both (`mode: atomic` header + per-statement rows; 47 lines and 17 lines respectively) while JSON went to stdout. `-json` and coverage **coexist**. |
| Both forms of "did not run" are ONE query | `go test -count=1 -json ./cmd/moai/...` and `./internal/template/scripts/` (packages with 0 test files) | Emits an `output` event carrying `[no test files]`, then a **package-level** `{"Action":"skip"}` with **no `Test` field**. See §C.1 — this is the single-predicate property. |
| Zero-test-file packages in this tree | `go list -f …` | Exactly **three**: `cmd/moai`, `internal/template/scripts`, `scripts/convert-nextra-to-hextra`. |
| **Build diagnostics travel in-stream, NOT on stderr** | throwaway module outside the repo, one package with a deliberate syntax error, `go test -json ./...` | rc=1; **stdout 666 bytes, stderr 0 bytes**; action sequence `build-output build-output build-fail start output fail`, the compiler diagnostic carried inside the `build-output` events. Measured twice independently (coordinator, then this agent on darwin/arm64 · go 1.26.4) with the same shape. **Falsifies the original brief's premise that stderr carries build errors.** |
| Failure bodies survive in the stream | a path typo — `./internal/version/...`, not a package here | rc=1; the stream carried `Action=="fail"` plus the verbatim body `FAIL\t./internal/version/... [setup failed]`. Observed, not assumed. |

**Extrapolation, stated as such**: the full-suite artifact size is **not measured**. It is an extrapolation from two packages (785,790 B and 1,178,655 B for one package each, over 135 packages) and is expected to be in the tens of megabytes uncompressed, ~15× smaller gzipped. No requirement below depends on a specific full-suite byte figure.

## B. Why this matters

Every downstream verification claim that cites CI rests on `rc=0`. An `rc=0` that cannot say *which* cells executed makes "the test passed" an unobserved claim (`verification-claim-integrity.md` §1.1). The absence of a failure signal is not evidence a check ran.

## C. Chosen approach (recorded, not re-opened)

`go test -json` redirected to a **file**, plus a **summary step** printing a census and any failures to the console, plus a **gzipped uploaded artifact**.

### C.1 The single-predicate property (the actual reason this option was chosen)

The `-json` stream expresses **both** forms of "did not run" through one event type, distinguished only by the presence of a field:

```
{"Action":"output","Package":".../cmd/moai","Output":"?   \t.../cmd/moai\t[no test files]\n"}
{"Action":"skip","Package":".../cmd/moai","Elapsed":0}
```

- `Action=="skip"` **with** a `Test` field → a test that skipped itself.
- `Action=="skip"` **without** a `Test` field → a package where nothing ran.

So a single query catches both, and the census classifies them by reading one field rather than running two independent detections. This is the design gain that decided the option — not merely that `-json` is machine-readable. Measured on `go test -count=1 -json ./cmd/moai/...` and `./internal/template/scripts/`.

Rejected alternatives and the measured reason:

- **`-v`** — inflates the console log ~913× (measured) and destroys readability at 11,124 tests.
- **A bare skip-COUNT gate** — a count never answers *which* cell did not run, so it cannot satisfy the completion condition in §D. A count MAY later be surfaced alongside the census as a gate hook; it is not the mechanism.

## D. Completion condition (the one that binds)

Not "the configuration changed" — but:

> **A test that actually skips is identifiable from the CI artifacts alone, on a real CI run, and that identification was OBSERVED and recorded.**

An AC asserting only that the workflow YAML contains `-json` is **insufficient** and must not be the terminal AC. See `acceptance.md` AC-CTO-007 (positive control).

## E. Requirements (GEARS)

- **REQ-CTO-001** (Ubiquitous) — The CI test invocation shall emit a machine-readable per-test event stream to a file, not to the console.
- **REQ-CTO-002** (Event-driven) — When a test suite invocation completes, the CI job shall print to the console a census naming every test or package that did not run.
- **REQ-CTO-003** (Ubiquitous) — The CI job shall not write the raw per-test event stream to the console log.
- **REQ-CTO-004** (Event-detected) — When a test failure is detected in the event stream, the summary step shall print the failing test names together with their captured output.
- **REQ-CTO-005** (Ubiquitous) — The CI job shall preserve the non-zero exit status of the test invocation, so redirecting the stream never converts a red run into a green one.
- **REQ-CTO-006** (Ubiquitous) — The census shall detect both forms of "did not run" with a single `Action=="skip"` query, and shall label them distinctly by the presence or absence of the `Test` field (§C.1).
- **REQ-CTO-007** (Capability gate) — Where a job collects coverage, the test invocation shall retain its `-coverprofile` and `-covermode` behaviour unchanged.
- **REQ-CTO-008** (Event-driven) — When a test invocation finishes, the job shall upload the compressed event stream as a build artifact, whether the invocation passed or failed.
- **REQ-CTO-009** (Ubiquitous) — The change shall not alter any branch-protection required check name, and shall not alter the console-visible failure text a reader relies on to diagnose a red run.
- **REQ-CTO-010** (Event-driven) — When the change has landed in a workflow that a real CI run executes, the SPEC owner shall record the run URL / job id and the verbatim census lines naming a genuinely skipping test.
- **REQ-CTO-011** (Ubiquitous) — The test invocation shall not redirect stderr into the event-stream file, so that no non-JSON text can corrupt the stream. **The rationale is stream integrity, NOT console visibility**: under `-json`, `go test` routes compiler diagnostics into **stdout** as `build-output` events terminated by `build-fail`, and stderr is empty (§A.1), so a bare stdout redirect *removes build errors from the console* rather than preserving them. Consequently the census **shall** surface build failures explicitly — stderr will not do it.

> **Pattern labels.** The canonical GEARS five are Ubiquitous, Event-driven, State-driven, Capability gate, and Event-detected. `Unwanted` belongs to the legacy EARS table, not to GEARS, so the three always-active negative constraints here (REQ-CTO-003, -009, -011) carry **Ubiquitous** — a prohibition with no trigger is an always-active requirement stated negatively. REQ-CTO-004 carries **Event-detected**: a detected failure producing a response is exactly that form.

## F. Call sites in scope

| File:line | Job | Current invocation | Check status |
|---|---|---|---|
| `.github/workflows/ci.yml:183` | `test` | `go test -coverprofile=coverage.out -covermode=atomic ./...` | required — `Test (ubuntu-latest)` |
| `.github/workflows/ci.yml:238` | `test-race` | `go test -race -count=1 ./...` | advisory (`Race Test`, deliberately not required) |
| `.github/workflows/ci.yml:329` | `test-integration` | `go test -tags=integration -race -timeout 180s ./test/integration/harness/...` | advisory — `Integration Tests (<os>)` is not in branch protection; **3-OS matrix** |
| `.github/workflows/release-pr-multi-os.yml:189` | `full-matrix-test` | `go test -race -timeout 25m ./...` | required via `Release PR Multi-OS Gate`, **3-OS matrix** |

Design constraints these sites impose:

1. `shell: bash` resolves to a verbatim shell string carrying **`-e` as well as `-o pipefail`** — on the Linux runner, `/usr/bin/bash --noprofile --norc -e -o pipefail {0}`. The Windows runner resolves a different `bash` path with the identical flag set, so `-e` binds on all four sites and nothing here is Linux-specific. Under `-e` the step dies AT a failing `go test`, so a following `rc=$?` never executes and the census never prints. Any exit-code capture must survive `-e`; `plan.md` §B.1 carries the prescribed form and the reason. (`shell: bash` is separately load-bearing on the two 3-OS jobs, where the windows runner would otherwise parse POSIX syntax as PowerShell — see `ci.yml:157`. That windows rationale does **not** apply to the ubuntu-only `test` job; the `-e -o pipefail` consequence applies to all four sites.)
2. **Build diagnostics travel inside the stream, not on stderr — and a bare stdout redirect therefore hides them.** Under `-json`, `go test` emits compiler diagnostics as `build-output` events terminated by `build-fail`, with **stderr empty (0 bytes)** — measured, §A.1. An earlier draft of this SPEC asserted the opposite (that stderr keeps build errors console-visible); that assertion was **falsified by measurement** and is corrected here.

   Two consequences follow, and the second is the load-bearing one:

   - stderr is still not redirected into the file (REQ-CTO-011), but the reason is **stream integrity**, not console visibility.
   - Console visibility of a build failure is restored **only** by the census emitting an explicit `BUILD FAILED` row carrying the `build-output` text. That row is **load-bearing, not cosmetic**: delete it and a build failure produces a red job whose console says nothing about why. No green run will ever contradict its removal — the same hazard shape as the `; rc=$?` form in `plan.md` §B.1.

   In-suite failures are separately covered: a path typo produced rc=1 with `Action=="fail"` and the verbatim body `FAIL\t./internal/version/... [setup failed]` (§A.1).
3. Two of the four sites (`ci.yml:329`, `release-pr-multi-os.yml:189`) run a 3-OS matrix including windows, where the summary implementation must not assume GNU-only tooling. `jq` is already used in four workflows in this repo, so it is an established dependency rather than a new one.
4. Artifact steps follow the house convention already in this repo: `actions/upload-artifact@v7` with `retention-days: 7`, matching `ci.yml:433-438`. The two matrix jobs need the OS in the artifact name or their three legs collide.
5. With `-json`, the human-readable `coverage: NN% of statements` lines move into JSON `output` events. Codecov reads the `coverage.out` **file**, not stdout, so the upload step is unaffected.

The scope is **exactly these four sites**. `ci.yml:329` was missed in the first draft of this SPEC and is brought in by lead ruling: it is the identical defect and the identical one-line change, and leaving a known-identical gap open while claiming completeness is the weaker choice. The remaining `-run`-scoped invocations — `lsel-leak-guard.yaml:37` and the three steps in `template-neutrality-check.yaml` — already carry `-v`, so a zero-match selector is visible in their logs today; they are excluded (§H).

## G. Observation vehicle (measured constraint)

- `ci.yml` has **no** `workflow_dispatch`; its triggers are push to `[main, develop]` and PR to `[main]`. A push to this card branch triggers **nothing**.
- `release-pr-multi-os.yml` **does** carry `workflow_dispatch` (line 16 on `origin/main`), and both gates admit dispatch events (`detect-release` at line 35, `full-matrix-test` at line 86). It runs the same full-suite `go test`. A `workflow_dispatch` run executes the workflow file **as it exists at the dispatched ref**, so the edited file is what runs — but the ref must exist on `origin`, so the card branch must be pushed for this path.
- Adding `workflow_dispatch` to `ci.yml` is **in scope as an enabling change**, with an honest limitation: a `workflow_dispatch` trigger only becomes dispatchable once the workflow file carrying it exists on the default branch. Adding it in this card does **not** make `ci.yml` dispatchable for this card's own observation.
- **Pushing the card branch to `origin` is a deliberate, minimal exception** to the lane protocol (`.claude/rules/local/gitflow-lane-protocol.md` §1 / CLAUDE.local.md §4.1 — cards merge into `develop`, they do not get PRs). Its basis is the lead's dispatch approval: `workflow_dispatch` resolves the workflow at a **remote** ref, so no push means no observation. Its cost is bounded and measurable: `ci.yml` carries no `workflow_dispatch` and triggers only on push to `[main, develop]` and PR to `[main]`, so **pushing this branch is CI-inert** — it starts no run and consumes no runner. The exception buys a remote ref and nothing else, and no PR is opened.
- The lead has **approved** dispatching `release-pr-multi-os.yml` against this branch ref, with one HARD constraint: **the dispatch is ONE run.** If it fails, the cause is established before any re-run; waiting for a pass by repeated re-execution is prohibited. The positive-control criterion is therefore written to be satisfiable from a single run's artifacts (`acceptance.md` AC-CTO-007).
- The post-merge `ci.yml`-on-`develop` run remains the fallback if dispatch proves unavailable, and choosing it must be stated as post-merge rather than presented as a pre-merge observation.

## H. Exclusions

### Out of Scope — the `develop` concurrency policy

- `concurrency.cancel-in-progress: true` at `ci.yml:30-32` cancels the prior run on every push, which left 8 consecutive heads unadjudicated on 2026-08-28. That is a **different axis**: this card asks "is what ran recorded in the log", that one asks "does the run finish". Pointer only; no requirement here addresses it.

### Out of Scope — a skip-count threshold gate

- Failing CI when the skip count exceeds a number. A count cannot name the cell that did not run, so it cannot satisfy §D. It may be added later on top of the census emitted here.

### Out of Scope — reducing the 631 `t.Skip` call sites

- This SPEC makes skips *visible*. Deciding which skips are illegitimate, and removing them, is separate work.

### Out of Scope — test-result reporting UI

- Third-party test-report actions, PR annotations, check-run summaries, and dashboards. The artifact plus console census is the deliverable surface.

### Out of Scope — the already-verbose `-run`-scoped invocations

- `lsel-leak-guard.yaml:37` and the three `go test` steps in `template-neutrality-check.yaml` already carry `-v`, so a zero-match selector is visible in their logs today. They have no observability gap and are not converted.

### Out of Scope — non-Go test surfaces

- Shell-script hook tests, docs-site builds, and any non-`go test` verification step.

## I. Closure ordering — what may close post-merge, and what is debt

Two ACs cannot be satisfied from the single approved dispatch, and this SPEC says so rather than asserting them:

| Item | Status at close | Why |
|---|---|---|
| AC-CTO-001, -002, -003, -003b, -004, -005a, -006, -007 (dispatch path), -008 | MUST PASS pre-close | obtainable locally or from the one green dispatch |
| AC-CTO-003 CI-level red-path confirmation | **DEBT** | needs a second, deliberately-red CI run; the ONE-run constraint forbids it. The shell semantics it guards are verified locally instead (AC-CTO-003). |
| AC-CTO-005b (artifact present on a FAILED run) | **DEBT** | same reason; the `if: always()` declaration and re-parseability are verified on the green run |

Both debts discharge on the first genuinely red CI run after this lands — an ordinary event, not a scheduled task.

If the fallback observation path is taken (post-merge `ci.yml` on the `develop` push), AC-CTO-007 closes **post-merge** and taking the fallback is recorded as **debt**, never as a pre-merge pass. The rule at §G — never present the fallback as a pre-merge observation — is unchanged and binding.

---

## J. Residual risks and gaps

Recorded per `verification-claim-integrity.md` §3 — these are **not** defects and **not** passes; they are what was left unobserved.

### Residual risk — the stream file sits in the repo root during the run

`test-stream.json` is written to the repository root *while* `go test ./...` is executing. It is gitignored, so every git-based check stays clean; but a test that walks the repo root **through the filesystem** rather than through git would see a file that did not exist when its expectations were written. **Unmeasured** — no such test is known to exist, and none was searched for. Recorded as residual risk rather than as a defect. Mitigation if it ever bites: write the stream to a path outside the repo root (e.g. `$RUNNER_TEMP`) and upload from there.

### Gap — the census has run on darwin/arm64 only

Every local verification of the census (AC-CTO-002, -003, -003b) ran on **darwin/arm64**. The census depends on `jq`, and on `gzip` / `sed` / `sort` in the surrounding step. Measured precedent in this repo: **all four** workflows that use `jq` (`auto-merge.yml`, `graph-freshness.yml`, `release-drafter-cleanup.yml`, `review-quality-gate.yml`) are `runs-on: ubuntu-latest` — so there is **no** existing evidence of `jq` on the windows or macOS runners, and none for `gzip`/`sed`/`sort` under git-bash.

Two of the four call sites are 3-OS matrices (`ci.yml:329`, `release-pr-multi-os.yml:189`), so **the approved dispatch is the first exposure of the census to windows and macOS**. This is a named Gap, not a predicted failure and not a pass: if a non-ubuntu leg fails on tooling, that is the ONE-run dispatch doing its job, and the cause is established before any re-run (`acceptance.md` AC-CTO-007).

---

## Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — why `rc=0` without per-test evidence is an unobserved claim.
- `.github/workflows/ci.yml`, `.github/workflows/release-pr-multi-os.yml` — the call sites in §F.
