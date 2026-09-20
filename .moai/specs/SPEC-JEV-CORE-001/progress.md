# SPEC-JEV-CORE-001 — Progress

Card: t1020 · Tier L · plan-phase artifacts authored 2026-09-20. Split from `SPEC-JEV-INTEGRATION-001` on the M1 seam.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 22 / ceiling 25; AC 16 / ceiling 25 |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 22 (REQ-JEVC-001 … REQ-JEVC-022) |
| Acceptance criteria | 16 (AC-JEVC-001 … AC-JEVC-016) |
| Predecessor | none — first of four |
| Successor | `SPEC-JEV-OPTIN-MEASURE-001` |
| Status transition | (none) → draft |

Open questions carried to the Implementation Kickoff Approval gate: Q2 (model-id pin: compiled vs config), Q4 (credential-reveal route).

## §E.2 Run-phase Evidence

Card t1020 · cycle_type=tdd · branch `WT-jev-init-optin` · base local `develop` at `fd75cf692`.
All evidence below was measured in this run, against this tree. Baseline attribution for every
row: HEAD `1b5dd2e3b` at measurement start (the two SPEC-document commits `51ababad4` and
`1b5dd2e3b` precede this work; no implementation commit existed when the tests were first run RED).

### Resolved open questions

| # | Question | Resolution | Where it landed |
|---|---|---|---|
| Q2 | Is the pinned model id compiled, or operator-visible configuration? | **COMPILED.** `jev.ModelID = "jev-1.13"` and `jev.EndpointURL` are Go constants; `workflow.jev` is a bare `enabled` flag, not a block. Grounds: the repository's [HARD] hardcoding rule places model names, endpoint URLs, and API headers in Go constants — so the operator-visible alternative was never available here. The consequence is deliberate: no configuration path can move the pin without a release, and none can aim a request at a different endpoint. | `internal/jev/jev.go` (constants) · `internal/config/types.go` `WorkflowJevConfig` (bare flag) |
| Q4 | Is a credential-reveal route wanted? | **NO.** REQ-JEVC-020 is satisfied by the `Configured` + final-four disclosure alone, and nothing in this SPEC or the chain requests plaintext retrieval. The GLM precedent's `glmKeyRevealPath` was read and deliberately not copied: `internal/jevcred` ships no path by which the stored credential leaves the process. | `internal/jevcred/jevcred.go` § `View` doc comment |

These resolutions are recorded here rather than in `spec.md` §E: run-phase may not edit SPEC body
content, and the §E Open Questions table is body content. A sync-phase or manager-spec pass owns
closing that table.

### AC PASS/FAIL matrix

Every row's command was run in this tree. `go test ./internal/jev/... ./internal/jevcred/...
-count=1` reported `ok` for both packages with 49 top-level tests passing and 0 skipped.

| AC | Status | Verification command | Observed |
|---|---|---|---|
| AC-JEVC-001 | PASS | `go test -run TestQueueFileHash_UnchangedAcrossAFullEnabledCall ./internal/jev/` | `ok` — queue-file SHA-256 identical before/after a full enabled call with a credential present; the same test's positive control confirms the comparison detects a deliberate one-byte change |
| AC-JEVC-002 | PASS | `go test -run 'TestPackageImports_AreStandardLibraryOnly\|TestQueueFileHash_UnchangedAcrossEveryUnavailablePath' ./internal/jev/` | `ok` — every non-test import of `internal/jev` is standard library (scanned 1 file, non-vacuity asserted); no `BacklogItem`-bearing type is reachable because no such dependency exists. Unavailable paths (disabled / no-credential / unreachable) also leave the file unchanged |
| AC-JEVC-003 | PASS | `go test -run 'TestJevCallPath_HasExactlyTheDeclaredConsumers\|TestJevCallPath_UnreachableFromDecisionSurfaces' ./internal/cli/` | `ok` — exactly one `internal/cli` file imports the client (`doctor_jev.go`, which doubles as the scanner's positive control); `gtd.go`, `todo_analysis.go`, `todo_autodone.go`, `integration.go`, `integration_settings_drift.go` import it zero times |
| AC-JEVC-004 | PASS | `go test -run TestAnswerLabel_IsDistinguishableFromMeasurementAndJudgement ./internal/jev/` | `ok` — `Answer.Label()` carries the `model signal` marker and the pinned model id, and contains none of `measured` / `verified` / `confirmed` |
| AC-JEVC-005 | PASS | `go test -run 'TestDefaults_JevDisabled\|TestTemplateWorkflowYAML_JevShipsOff' ./internal/config/` | `ok` — compiled default `false`; the shipped template decodes to `false`, with a sibling-key positive control proving the decode read the file |
| AC-JEVC-006 | PASS-WITH-DEBT | `go test -run 'TestDisabled_ConstructsNoRequest\|TestCheckJev_DisabledProbesNothing' ./internal/jev/ ./internal/cli/` | `ok` — zero transport calls and zero reachability probes while disabled. **Debt**: the criterion asks for byte-identical stdout/stderr on every affected command, and `moai doctor` is NOT byte-identical — it gains exactly one row. That row is the SPEC's single sanctioned surface change (REQ-JEVC-022), so the divergence is intended, not a regression; it is recorded as debt rather than PASS because the criterion's wording admits no such carve-out. Diff measured: `git diff -- internal/cli/testdata` shows one added check line plus the summary counters in each of the three golden files, and nothing else. No other command's output changes, because the capability ships with zero other consumers |
| AC-JEVC-007 | PASS | `go test -run 'TestNoCredential_TypedUnavailableNotAnError\|TestCheckJev_EnabledWithoutCredentialIsAdvisoryNotAFailure\|TestNoticeLine_NamesTheConditionAndStaysOneLine' ./internal/jev/ ./internal/cli/` | `ok` — an absent credential returns a typed `no-credential` value (never an error), `NoticeLine()` is a single line for every unavailable condition and empty when available, and the doctor check reports advisory rather than fail |
| AC-JEVC-008 | PASS | `go test -run TestTransportConditions_MapToTypedUnavailable ./internal/jev/ -v` | `ok` — six sub-cases: `401 unauthorized` → `unauthorized`, `429 rate limited` → `rate-limited`, `529 overloaded` → `overloaded`, `network unreachable` → `unreachable`, plus `unexpected 500` and `undecodable body` → `unreachable`. Each carries a non-empty observed condition and a one-line notice |
| AC-JEVC-009 | PASS | `go test -run 'TestRetry_NotIssuedForANonRepeatableRequest\|TestRetry_BoundedForARepeatableRequestThatFailedBeforeDelivery\|TestRetry_NotIssuedForAnAmbiguousFailure' ./internal/jev/` | `ok` — a request not declared repeatable: 1 transport call with `MaxRetries=3`; a repeatable request failing pre-delivery (dial): 3 calls with `MaxRetries=2`; a repeatable request failing **ambiguously** (read error, delivery unknown): 1 call. The third is the case the requirement is written for |
| AC-JEVC-010 | PASS | `go test -run TestSave_FileMode0600 ./internal/jevcred/` | `ok` — a fresh write lands at mode 0600 |
| AC-JEVC-011 | PASS | `go test -run TestSave_NarrowsExisting0644to0600 ./internal/jevcred/` | `ok` — a pre-existing 0644 credential file is tightened to 0600 on the next write, and reads back correctly |
| AC-JEVC-012 | PASS | `go test -run 'TestTypeSafeCredential_AbsentFromSchema\|TestSchemaScanPositiveControl' ./internal/jevcred/` | `ok` — no `settings.AllFields()` entry matches any of six credential-shaped needles; the scan asserts a non-empty field set (non-vacuity) and a paired positive control proves the substring predicate fires on a synthetic leak |
| AC-JEVC-013 | PASS | `go test -run TestView_DisclosureFloor ./internal/jevcred/ -v` | `ok` — sub-cases `five chars` → hint `bcde`, `long fixture` → hint `wxyz`; the hint is exactly four characters and never equals the stored credential |
| AC-JEVC-014 | PASS | `go test -run TestView_DisclosureFloor ./internal/jevcred/ -v` | `ok` — sub-cases `one char` and `exactly four` → `Configured: true` with an **empty** hint; `absent` → `Configured: false` |
| AC-JEVC-015 | PASS | `go test -run 'TestScreening_\|TestScreenPayload_BothDirectionsDirectly' ./internal/jev/` | `ok` — both directions. Detecting: four credential shapes (`sk-`, `ghp_`, `AKIA`, a secret-named dotenv assignment) plus the stored credential appearing verbatim are each refused with zero transport calls; non-detecting: an ordinary payload carrying paths, a card id and a commit prefix sends normally, and a two-character stored credential does not turn a clean payload into a hit |
| AC-JEVC-016 | PASS | `go test -run TestCheckJev_ ./internal/cli/` | `ok` — the check reports enabled-state, credential presence (via the four-character hint, never the credential), and endpoint reachability; the probe counter reads exactly 1 when enabled and exactly 0 when disabled, so no judgment request is sent on either path |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| The capability ships with ZERO consumers | PASS | `TestJevCallPath_HasExactlyTheDeclaredConsumers` states the consumer set as an exact one-element set (`doctor_jev.go`), not as an absence, and fails if that one file stops importing the client — the positive control that makes the zero-result attributable |
| No test contacts the real endpoint | PASS | The only occurrence of the host literal in any test file is a constant-equality assertion (`jev_test.go:579`) which dials nothing; every client that reaches transport in a test sets either a fake `Doer` or an `httptest` endpoint. No test requires a credential to be present to pass |
| No test writes into the project tree | PASS | Every temp directory is `t.TempDir()`; `jevcred.HomeDirFn` is swapped per test and restored on cleanup. `t.Setenv("HOME", …)` is used nowhere |
| Template neutrality of the shipped block | PASS | `grep -nE 'SPEC-[A-Z]\|REQ-[A-Z]\|card t\|[0-9]{4}-[0-9]{2}-[0-9]{2}\|jev-1' <jev block>` → no output; the block names no SPEC id, card id, date, price, model id, or endpoint URL |

### Files changed

| File | Role |
|---|---|
| `internal/jev/jev.go` | NEW — the single call path: pinned model id + endpoint constants, `Availability` enum, request/answer types, gate, size bounds, secret screening, transport, fail-open mapping, retry policy, usage accounting |
| `internal/jev/jev_test.go` | NEW — 30 tests over the pin, the gate, fail-open, retry, bounds, batching, usage, screening (both directions), labelling, and the local-server transport shape |
| `internal/jev/display_only_test.go` | NEW — the display-only invariant: standard-library-only import assertion (with classifier positive control) and the queue-file SHA-256 before/after comparison (with change-detection positive control) |
| `internal/jevcred/jevcred.go` | NEW — the credential SSOT: `~/.moai/.env.typesafe` at 0600 with wider-mode tightening, `HomeDirFn` seam, dotenv escape/unescape, the four-character disclosure floor, no reveal route |
| `internal/jevcred/jevcred_test.go` | NEW — 16 tests over path resolution, round-trip, mode 0600, the 0644→0600 tightening, the test-env seam, unreadable/empty/absent degradation, and the disclosure floor |
| `internal/jevcred/schema_absence_test.go` | NEW — the AC-JEVC-012 anti-leak regression guard plus its positive control and the envkeys alias-parity check |
| `internal/defs/dirs.go` | EXTEND — `TypeSafeEnvFileName = ".env.typesafe"` beside the GLM sibling |
| `internal/config/envkeys.go` | EXTEND — `EnvTestTypeSafeKey` as the canonical owner of the test-seam literal |
| `internal/config/types.go` | EXTEND — `WorkflowJevConfig` + the `Jev` field on `WorkflowConfig` |
| `internal/config/defaults.go` | EXTEND — the compiled `false` default, in the opt-in switch family |
| `internal/config/cache.go` | EXTEND — cache schema version 4 → 5, because a Workflow field addition otherwise lets an old-binary cache serve `enabled=false` over an `enabled: true` workflow.yaml |
| `internal/config/workflow_jev_test.go` | NEW — default-off, key shape, absent-key, template-ships-off (with decode positive control), cache-version bump |
| `internal/config/testdata/shipped_key_inventory.yaml` | EXTEND — `workflow.jev.enabled` triaged as class W (live reader), required by the shipped-key anti-rot guard |
| `internal/cli/doctor_jev.go` | NEW — the `Jev` readiness check: gate read first and short-circuiting, credential presence via the bounded hint, TCP-dial reachability probe behind a seam, never a failing status |
| `internal/cli/doctor_jev_test.go` | NEW — the doctor check's six cases plus the AC-JEVC-003 call-path reachability guards |
| `internal/cli/doctor.go` | EXTEND — one registry row for the check |
| `internal/cli/binary_lag_test.go` | EXTEND — `jevCheckName` added to the after-baseline allow-list, which is how that guard admits a later SPEC's legitimate check |
| `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` | REGEN — one added row and the summary counters, per AC-JEVC-006's recorded debt |
| `internal/template/templates/.moai/config/sections/workflow.yaml` | EXTEND — the shipped `jev.enabled: false` block with neutral prose |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-20
run_commit_sha: c032cd15a
run_status: audit-ready
ac_pass_count: 15
ac_fail_count: 0
ac_pass_with_debt_count: 1     # AC-JEVC-006 — see the matrix row
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable   # card worktree on WT-jev-init-optin; no push in this phase
l44_post_push_fetch: not-applicable    # run-phase does not push (lane protocol: lead batch-pushes develop)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass      # go build ./... → exit 0
  windows_amd64: pass     # GOOS=windows GOARCH=amd64 go build ./... → exit 0
  linux_amd64: pass       # GOOS=linux GOARCH=amd64 go build ./... → exit 0
coverage:
  internal/jev: 90.6       # target 85
  internal/jevcred: 85.0   # target 85
lint:
  changed_packages: 0 issues   # golangci-lint run over jev, jevcred, cli, config, defs
template_first:
  make_build: pass             # catalog hashes unchanged; binary rebuilt
total_run_phase_files: 23   # counted from `git status --porcelain --untracked-files=all`, not estimated
m1_to_mN_commit_strategy: single-commit   # M1a-M1e landed together; no intermediate commit
```

**Gaps (explicitly not observed).**

- The full suite (`go test ./...`) was NOT run locally, by repository rule. Package-scoped runs
  covered `internal/jev`, `internal/jevcred`, `internal/config`, `internal/defs`,
  `internal/template`, and `internal/cli` (the last in full, post-regeneration, under a
  `cli-suite` slot lease: `go test ./internal/cli/ -count=1 -timeout 22m` → `ok … 968.376s`).
  Every other package is unmeasured here; CI on the integration branch is the full-suite verdict.
- **A `FAIL` on `internal/cli` can be the clock rather than the code, and the two are not
  distinguishable from the summary line.** Two runs of that package at the default timeout each
  reported `FAIL … 600.8s`: the first carried four real `--- FAIL` lines (the golden files and the
  check-name allow-list, both since addressed), the second carried none at all. The second was the
  10-minute default `go test` timeout expiring on a loaded machine — the package needs ~968s here.
  A `FAIL` with no `--- FAIL` line above it is therefore a measurement that did not finish, and
  reading it as a code failure is an unobserved defect claim. Raise `-timeout` and re-measure.
- No real network call was made to the TypeSafe endpoint, by design. Nothing here establishes that
  the constructed request shape is accepted by the live API — only that it is constructed,
  screened, bounded, and decoded as specified.
- Windows behaviour of the 0600 / 0644→0600 file-mode assertions is unmeasured: both tests skip on
  `windows`, because the platform does not enforce POSIX permission bits. The credential-tightening
  guarantee is observed on unix only.
- Two Bash invocations were **refused rather than executed** by the worktree-isolation guard (a
  heredoc file-write and a grep whose pattern contained an arithmetic-looking token). Both were
  authoring or convenience calls, not verification commands: the first was re-issued through the
  Write tool and the second through a simpler grep, and no measurement was substituted by reading.

**Residual risk.**

- The token estimate is a byte-count heuristic (`len(s)/4`), not a tokenizer. It over-estimates
  multi-byte scripts, so the 32k/64k bounds refuse earlier than the vendor would — safe in the
  refusal direction, but a request the vendor would have accepted can be refused.
- The secret screener recognizes six declared credential shapes plus the stored credential
  verbatim. A credential shape outside that set passes it. The set is deliberately narrow: an
  entropy heuristic would refuse ordinary payloads, and a screener that refuses everything is
  indistinguishable from one that refuses nothing.
- `preDelivery` classifies retry-safety from `net.OpError{Op: "dial"}`. A transport wrapping its
  dial failure in a shape that does not unwrap to that type is classified ambiguous and therefore
  never retried — the safe direction, but it means the retry path is narrower in practice than the
  `MaxRetries` field suggests.
- The doctor reachability probe is a TCP dial. It establishes that the host accepts a connection,
  not that the API is serving — which is what REQ-JEVC-022 asks for (readiness without a judgment
  request), and no more.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
