# Progress — SPEC-CODEX-HOOK-ADAPTER-001

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-08-22
plan_audit_iterations: 2
plan_audit_iter1: FAIL 0.63 (Clarity 0.50 / Completeness 0.75 / Testability 0.50 / Traceability 1.00)
plan_audit_iter2: PASS 0.92 (Clarity 0.75 / Completeness 1.00 / Testability 1.00 / Traceability 1.00)
tier: M
tier_pass_threshold: 0.80
```

Iteration 1 returned five blocking defects, two of which were factual errors in
the SPEC's own findings — both verified against the source before being accepted
(the dispatcher registers a counterpart for all eleven Codex events; the
`version` key failure is reported in the `--json` stream rather than silently).
Reports: `.moai/reports/t83/plan-audit.md`, `plan-audit-iter2.md`.

## §F Phase 4 Mode Selection

**Input parameters** — tier M; scope 7 new files in one new package; 1 domain
(Go source); language mix 100% Go; concurrency benefit LOW (coding-heavy).

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | Not trivial — seven requirements with behavioral tests |
| `serial` | **yes** | Coding-heavy single-domain work; the default fallback |
| `fanout` | no | Single domain, and Anthropic's coding-task parallelism caveat applies |
| `sweep` | no | One package, not a ≥30-file uniform mechanical transform |

**Decision: serial**

Justification: the work is coding-heavy and confined to one new package, which
is the case the coding-task parallelism caveat points at sequential execution.
There is no independent-file fan-out to exploit — the milestones build on each
other (the discard branches are only knowable once the output mapping exists).

## §E.2 Run-phase Evidence

Implemented as `internal/codexadapter` — a new package, outside `internal/hook`
per REQ-7. TDD throughout: each milestone's test file was written and observed
failing before the implementation existed.

| Milestone | Deliverable | Files |
|---|---|---|
| M1 | Event table + refusal paths | `events.go`, `events_test.go`, `dispatcher_registration_test.go` |
| M2 | Output mapping for the three inert keys | `output.go`, `output_test.go` |
| M3 | Discard diagnostics + sink | `diagnostics.go`, `diagnostics_test.go` |
| M4 | Per-event stderr classification | `stderr.go`, `stderr_test.go` |
| M5 | Config constraint validator | `config.go`, `config_test.go` |
| M6 | Golden-file tests | `golden_test.go` |
| M7 | Verification | (this section) |

### RED evidence (test-first, verbatim pre-GREEN)

```
$ go test ./internal/codexadapter/          # M1, before events.go existed
internal/codexadapter/events_test.go:84:14: undefined: Resolve
internal/codexadapter/events_test.go:104:7: undefined: IsUnadapted
internal/codexadapter/events_test.go:107:6: undefined: IsUnknownEvent
FAIL	github.com/modu-ai/moai-adk/internal/codexadapter [build failed]

$ go test ./internal/codexadapter/          # M2/M3, before output.go existed
internal/codexadapter/output_test.go:198:20: undefined: DiscardBranchCount
internal/codexadapter/output_test.go:204:7: undefined: isDiscardableKey
FAIL	github.com/modu-ai/moai-adk/internal/codexadapter [build failed]
```

### Design decisions taken during implementation

- **Pass-through returns the original bytes**, not a re-marshalled equivalent.
  AC-REQ-2d demands byte-identical output for keys that measured working, and
  round-tripping a map through `encoding/json` reorders keys.
- **`continue: true` is left alone.** Only the blocking form carries intent the
  rewrite must preserve; rewriting the non-blocking form would invent a decision
  the hook never made.
- **The sink is explicitly not stderr**, and the stderr mirror is suppressed
  when the hook exited 2 (AC-REQ-3c). stderr on that path carries the blocking
  reason or continuation prompt REQ-4 requires passing through unmodified.
- **`ValidateConfig` collects every violation** rather than returning the first.
  An operator fixing a generated config wants the whole list, since each bad key
  disables the file the same silent way.

## §E.3 Run-phase Audit-Ready Signal

```
run_status: audit-ready
run_complete_at: 2026-08-22
base_commit: 3556ca1de
```

### AC PASS/FAIL matrix

Every row is a test that was run and observed in this tree.

| AC | Status | Verification | Observed |
|---|---|---|---|
| AC-REQ-1a | PASS | `TestEventTableRowCount`, `TestEventTableMapping`, `TestAdaptedRowCount`, `TestResolveAdapted`, `TestResolveRecognizedButUnadapted`, `TestDispatcherArgsExist` | all PASS; 11 rows, 6 adapted |
| AC-REQ-1b | PASS | `TestResolveUnknownEvent`, `TestResolveErrorNamesReceivedValue` | PASS — unknown refused, distinguishable from unadapted |
| AC-REQ-2a | PASS | `TestContinueFalseRewritesToDecisionBlock` | PASS |
| AC-REQ-2b | PASS | `TestContinueFalseWithoutStopReasonGetsDefaultReason` | PASS — reason non-empty |
| AC-REQ-2c | PASS | `TestSystemMessageRoutedToAdditionalContext` | PASS |
| AC-REQ-2d | PASS | `TestWorkingKeysPassThroughByteIdentical` | PASS — 3 shapes byte-identical |
| AC-REQ-3a | PASS | `TestSystemMessageDiscardedWhereNoChannel`, `TestDiscardRecordCarriesNoContent`, `TestRecordDiscardsWritesSinkRecord` | PASS — event+key+length recorded, content absent |
| AC-REQ-3b | PASS | `TestDiscardBranchCountMatchesTested` | PASS — 3 tested == `DiscardBranchCount` |
| AC-REQ-3c | PASS | `TestRecordDiscardsSilentOnStderrWhenHookBlocked` | PASS — stderr empty, sink record still written |
| AC-REQ-4a | PASS | `TestPreToolUseStderrIsBlockingReason` | PASS |
| AC-REQ-4b | PASS | `TestStopStderrIsContinuationPrompt`, `TestTwoClassesAreDistinct` | PASS |
| AC-REQ-4c | PASS | `TestUnmeasuredClassIsAnnotated`, `TestExcludedEventsHaveNoClass` | PASS |
| AC-REQ-5a | PASS | `TestValidatorRejectsVersionKey` | PASS |
| AC-REQ-5b | PASS | `TestValidatorAcceptsMeasuredGoodShape`, `TestValidatorRejectsPerLevelUnknownKeys` | PASS — 3 levels |
| AC-REQ-6 | PASS | `TestGoldenPayloadsParse`, `TestGoldensAreTracked` | PASS — 6 events from tracked `testdata/` |
| AC-REQ-7 | PASS | `git diff --name-only 3556ca1de..HEAD -- internal/hook/` | `0` files |

38 tests, 38 PASS, 0 FAIL.

### Verification commands and observed output

```
$ go build ./...                                   -> exit 0
$ go vet ./internal/codexadapter/...               -> exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/codexadapter/...  -> exit 0
$ GOOS=linux GOARCH=amd64 go build ./...           -> exit 0
$ gofmt -l internal/codexadapter/                  -> (no output)
$ golangci-lint run --timeout=3m ./internal/codexadapter/...
    0 issues.
$ go test ./internal/codexadapter/... -cover
    ok  github.com/modu-ai/moai-adk/internal/codexadapter  0.605s  coverage: 87.1% of statements
$ go test -race ./internal/codexadapter/...
    ok  github.com/modu-ai/moai-adk/internal/codexadapter  1.809s
$ git diff --name-only 3556ca1de..HEAD -- internal/hook/ | wc -l
    0
```

Baseline attribution: all of the above ran in this tree, on branch
`WT-hook-adapter`, against base commit `3556ca1de`.

### Gaps (not verified)

- **Full suite not run locally** — per the repo's load discipline, the whole-tree
  verdict comes from CI on the pushed head. Only the affected package was run
  here, plus a whole-tree `go build`.
- **No live Codex run exercises this code.** Every behavioral assertion is a unit
  test against the transform, not an end-to-end run through codex-cli. The
  measurements the transforms encode were made in the plan phase; the adapter
  itself has not been wired into a hook and observed firing.
- **`internal/hook` was not re-run.** It is unmodified (0 files), and this package
  only imports its exported constants.
- **The five unadapted events have no behavioral coverage** beyond being refused.

### Residual risk

- The mapping encodes one Codex build (0.147.0). A build that makes an inert key
  live would make part of REQ-2 redundant — wasteful rather than wrong.
- `continue`/`stopReason`/`systemMessage` were measured inert on PreToolUse and
  PostToolUse only. If one is live on another event, the rewrite does unnecessary
  work there.
- Nothing calls this package yet. It is a library with tests; wiring it to the
  dispatcher is downstream work, so a defect in how it would be *invoked* is out
  of reach of this SPEC's evidence.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-08-22"
sync_commit_sha: "0264af589"
sync_status: "completed"
spec_id: SPEC-CODEX-HOOK-ADAPTER-001
changelog_entry_position: "[Unreleased] / ### Added / internal/codexadapter entry"
b12_self_test_a:
  pre_emission_grep_count: 0
  result: PASS
b12_self_test_b:
  acceptance_ac_count: 7
  changelog_cited_ac_count: 7
  result: PASS
b12_self_test_c:
  cited_paths_verified:
    - internal/codexadapter (package, 13 files)
    - .moai/reports/t83/precondition-measurement.md
    - .moai/reports/t83/precondition-measurement-round3.md
  result: PASS
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  applied_at: "2026-08-22 (single sync commit, 3-phase close)"
run_verification:
  source: "orchestrator independent re-execution at tree 95c4a253b (WT-hook-adapter)"
  builds_darwin_windows_linux: "exit 0"
  go_test_cover: "ok ./internal/codexadapter/... coverage: 87.1%"
  tests: "38 PASS / 0 FAIL"
  race: "ok"
  golangci_lint: "0 issues"
  internal_hook_diff_base_3556ca1de: "0 files"
scope_note: >-
  Library-only card. CHANGELOG entry frames the adapter as a tested library,
  not a live feature; the wiring is the pending follow-up card (M4 --agent
  generator). README and docs-site deliberately untouched.
```

