# progress — SPEC-JEV-LEDGER-001

status: completed

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-26T10:21:34Z
tier: M
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
tree_measured: 954adab25e7d6d83b63f8fb1c8f1aafa859f50c1
card: t1762 (split "a")
note: scope reduced at 0.2.0 by factory-leader ruling (operator-delegated); 11 files, ~770 lines est.; REQ 12/16, AC 11/16 (plan.md §E)

iter2_repair:
  trigger: plan-audit FAIL 0.6875 at audited_sha d1aa4ed656af948c95c2567718c4f9a7be54e191 (.moai/reports/t1762/jev-ledger-plan-audit.md)
  repaired_at: 2026-10-26T10:41:41Z
  defects: D1-D8 (see git history of this file at 0c2d2aa4c2)

iter3_repair:
  trigger: plan-audit FAIL 0.81 at audited_sha 82aff00a6d3b9d1b3dd6f9c2a6fc68ebf5b8e0d0 (.moai/reports/t1762/plan-audit-verdict.md; that path now holds the final verdict at 0c2d2aa4c2)
  version: 0.1.2
  defects: N1-N6 (see git history of this file at 0c2d2aa4c2)

scope_reduction_0_2_0:
  trigger: Tier M audit ceiling reached with FAIL 0.88 at audited_sha 0c2d2aa4c230d2a6671037cefea260ede18e295d (.moai/reports/t1762/plan-audit-verdict.md R1-R4; .moai/reports/t1762/iter-3-adversarial-verdict.md); hold record .moai/reports/t1762/hold-and-split.md
  ruling: factory leader, operator-delegated — split; t1762 keeps M1 (config map, budget constants, template, make build) and M2 (ledger recorder with the regular-file-checked write open)
  migrated: hook adapter, call site, shadow byte identity, deadline anchor and T, M3 measurement (R3), credential wiring (R4) -> SPEC-JEV-LEDGER-002 (card t1970); moai jev cost and its unit label -> SPEC-JEV-LEDGER-003 (card t1971); full id table in spec.md §D.1
  repaired_in_scope: R1 (FIFO fixture and Unix-only flags in a build-tagged test file with a complement-tagged counterpart; DoD adds GOOS=windows go vet); R2 (AC-012 case (iv), FIFO with an open reader, and the handle-check-dropped mutant); O-b (open_other.go complement tag); O-e (ledger pointers re-collected)
  new_ids: REQ-JEVL-017, REQ-JEVL-018, AC-JEVL-017, AC-JEVL-018
  package_decision: recorder in new package internal/jevledger; no file in internal/jev edited (decision-index Q7; card t1968 owns the internal/jev client)
  red_recollected: E1-E12, B1 on 0c2d2aa4c2 (acceptance §L)
  open_gap: the 5 s / 1 s constants remain decisions until -002 measures

reduced_iter1_repair_0_2_1:
  trigger: plan-audit FAIL 0.81 at audited_sha 954adab25e7d6d83b63f8fb1c8f1aafa859f50c1 (.moai/reports/t1762/plan-audit-verdict.md, blocking D1-D4)
  version: 0.2.1
  repaired: D1 (AC-014 forbidden-import list for jev_consumers.go; jevledger stays stdlib-only; plan §B/§D file placement), D2 (threshold in REQ-009 and AC-009), D3 (§D.1 REQ-009 part row), D4 (§D.1 REQ-008 part warning passthrough)
  optional_repaired: D5 (AC-013 binding JevHook constraint), D6 (test-deletion residual risk), D7 (REQ-014 Ubiquitous positive form), D8 (E11/E12 platform-split pointer), D9 (AC-012 case (v) directory)
  ids_unchanged: REQ 12/16, AC 11/16
  red_recollected: E1-E12, B1 on 954adab25e, identical to the 0c2d2aa4c2 record (acceptance §L)

## Binding run conditions

Copied verbatim from .moai/reports/t1762/plan-audit-verdict.md (PASS-WITH-DEBT, audited_sha d6dc623eb803d061225f022598435fce61c1651e) at the plan-to-run Kickoff; the sync auditors re-read them (.claude/rules/moai/workflow/auto-semantics.md §9.1).

- debt: N1 dispose_in=run AC-JEVL-014의 import 검사는 금지 목록 방식이라 `net/rpc`, `net/smtp`, `os/exec`(두 대상 모두), 제3자 HTTP 클라이언트(`jev_consumers.go`)가 통과한다. run 단계 경계 테스트는 AC 문구를 바꾸지 않고 더 엄격하게 짤 수 있다: (a)는 `net`·`net/*`·`os/exec`도 거부, (b)는 "표준 라이브러리 + `gopkg.in/yaml.v3`만 허용"으로 판정하고 그 거부 사례를 변이 탐침으로 기록한다.
- debt: N4 dispose_in=run `JevConsumerConfig`(원시 `yaml.Node` 필드)를 어느 파일에 선언할지 run 단계에서 정하고 progress.md에 기록한다. `types.go`에 두면 `types.go`가 yaml을 import하게 되어 plan §B의 "types.go imports no yaml library" 진술과 어긋난다. 어느 AC도 깨지지 않는다.

## §E.2 Run-phase Evidence

Run-phase owner: manager-develop, cycle_type=tdd. Start tree: `ff7a0a087584de12d4fd9d8d0c167a201b11506f` (clean, 0 behind origin/main). One commit per milestone; the M1 and M2 commit SHAs are recorded in §E.3 (a commit cannot cite its own hash).

### M1 — config surface, budget constants, template block

RED (tdd-result-contract EXPECTED_RED). The RED-now cells of acceptance.md §L were re-run on the start tree first: E1, E2, E3, E4, E5, E6, E7, E8, E9 each printed nothing (or the `ls` error) and exited 1; E10 exited 0 with `.gitignore:401:.moai/state/	.moai/state/jev/decisions.jsonl` (regression-guard); B1 `GOOS=windows go vet ./internal/config/` exited 0. Then a signatures-only skeleton (`internal/config/jev_consumers.go` with empty bodies, `DefaultJevHook*` = 0) was added so the tests compile, and the new tests were run before any behaviour existed:

```
$ go test ./internal/config/ -count=1 -run 'TestResolveJevConsumer|TestJevConsumers|TestLoadWorkflow_|TestJevHookBudgetConstants|TestTemplateJevConsumersEmpty|TestConfigCacheSchemaVersion_BumpedForJevConsumers'   -> exit=1, 34 FAIL lines
jev_consumers_test.go:59: effective mode = "", want "shadow" for every value in this SPEC
jev_consumers_test.go:125: threshold absent, want 0
jev_consumers_test.go:299: loaded 0 consumers, want all 8: []
jev_consumers_test.go:363: section warnings = [], want exactly one returned as data
jev_consumers_test.go:456: configCacheSchemaVersion = 13, want >= 14 — WorkflowJevConfig gained Consumers without a cache schema bump
jev_consumers_test.go:463: DefaultJevHookBudget = 0s, want 5s
jev_consumers_test.go:469: offset at the defaults = 0s, want 4s
jev_consumers_test.go:506: workflow.jev.consumers kind = 0, want an (empty) mapping; zero means the key is absent
```

Every failure is an assertion at the intended behaviour; none is a compile, discovery or fixture failure.

GREEN (PASS). `go test ./internal/config/ -count=1 -run '<same selector>' -v` -> exit 0. Named PASS lines: `--- PASS: TestResolveJevConsumer_Mode` (+ 6 subtests shadow, advise, steer, confirm, absent, bogus), `--- PASS: TestResolveJevConsumer_Threshold` (+ 12 subtests incl. zero, one, half, negative, above-one, string, mapping), `--- PASS: TestLoadWorkflow_MalformedConsumerQuarantined`, `--- PASS: TestLoadWorkflow_NonMappingConsumersQuarantined`, `--- PASS: TestLoadWorkflow_JevCreatesNothing` (+ enabled-false, enabled-true; AC-JEVL-001 config half, partial until M2), `--- PASS: TestJevHookBudgetConstants`, `--- PASS: TestTemplateJevConsumersEmpty`, plus supporting `TestResolveJevConsumer_ModeShapes`, `TestResolveJevConsumer_EntryShape`, `TestJevConsumers_AliasEntry`, `TestJevConsumers_NonScalarKey`, `TestJevConsumers_SurviveYAMLAndJSONRoundTrips`, `TestConfigCacheSchemaVersion_BumpedForJevConsumers`. Regression set: `go test ./internal/config/ -count=1` exit 0 (`ok ... 10.586s`); `go test ./internal/template/ -count=1` exit 0 (`ok ... 360.499s`); `go test ./internal/settings/ -count=1` exit 0.

Mutants run against the implementation, each reverted afterwards (`diff` against the saved copy printed nothing):

| AC | Mutant | Result |
|---|---|---|
| 006 | `advise` passed through as the effective mode | RED: `TestResolveJevConsumer_Mode/advise` |
| 007-1 | open-interval check (`f > 0`) rejects 0 | RED: `.../Threshold/zero`, `TestLoadWorkflow_MalformedConsumerQuarantined` |
| 007-2 | a decode error from the threshold propagates to the section (what a typed `*float64` field does) | RED: enabled and the other setting reverted, 0 of 8 consumers loaded |
| 007-3 | non-mapping `consumers` fails the decode (typed map) | RED: section reset, no section warning |
| 007-4 | resolver logs through `slog.Warn` | RED: `slog records=2` |
| 013-1 | margin 500 ms | RED: `DefaultJevHookMargin = 500ms, want 1s` |
| 013-2 | derivation returns a stored 4 s | RED: `offset for 250ms/50ms = 4s, want 200ms` |
| 017-1 | template ships a sample consumer | RED: `has 2 nodes, want an empty mapping` |
| 017-2 | template drops the `consumers` key | RED: `kind = 0 ... key is absent` |

One probe first stayed green and was discarded as not being the mutant: making only `JevConsumerConfig.UnmarshalYAML` return an error left AC-007 green, because `JevConsumers.UnmarshalYAML` deliberately discards that error. The faithful reset-the-section mutant is 007-2 above, where the error propagates.

Other M1 verification: `make build` exit 0 and `git status --short` afterwards lists only the intended files; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `GOOS=windows go vet ./internal/config/` exit 0; `golangci-lint run ./internal/config/` -> `0 issues.`; `go test -cover` per function in `jev_consumers.go` is 100.0% for every function.

Dispositions and decisions recorded at M1:

- **N4 (binding run condition)** — `JevConsumerConfig` (raw `yaml.Node` fields) is declared in `internal/config/jev_consumers.go`, next to the two `UnmarshalYAML` methods, following `slot_lease_config.go` and `loader_card_age_alarm.go`. `types.go` gains only the `Consumers JevConsumers` field and a doc comment and imports no yaml library, so plan §B's "types.go imports no yaml library" stays true. Check: `grep -n 'yaml.v3\|"gopkg' internal/config/types.go` prints two comment lines (`:346`, `:2142`) and no import line; the import block at `:3-8` is `slices`, `time`, `pkg/models`.
- **Consumers carrier is a struct, not a bare map** — plan §B says "a map type". A bare map has nowhere to carry the section-level warning for a non-mapping `consumers` value as data (AC-JEVL-007 requires the map to be empty AND exactly one warning returned). `JevConsumers` is therefore `struct{ Entries map[string]JevConsumerConfig; Warnings []string }` with its own `UnmarshalYAML`/`MarshalYAML`. The two custom decode methods, the resolver and the tolerance semantics are as planned.
- **Marshal methods added** — `internal/config/manager.go` `saveSection` yaml-marshals the whole `WorkflowConfig`, and the config cache JSON-marshals `Config`; without `MarshalYAML` a configured consumer would be rewritten as `entries:/warnings:` and lose its shape. `TestJevConsumers_SurviveYAMLAndJSONRoundTrips` pins both.
- **Cache schema bump** — `configCacheSchemaVersion` 13 -> 14 in `internal/config/cache.go`, per that file's own rule for a field added to `Workflow`. This is a sixth M1 file beyond plan §E's five (cascade follow-up within the SPEC's scope).
- **Local dogfood `.moai/config/sections/workflow.yaml`** — carries the same `consumers: {}` block as the template, with the local `enabled: true` left untouched. Plan §D said this was not required; the run brief asked for the two to be kept consistent.
- **`spec.md` / `progress.md` status** — left at `draft`: the run brief forbids editing `spec.md` (a hashed audit subject), and the `draft -> in-progress` frontmatter transition is a `spec.md` edit. Reported as a Gap for the leader.

### M2 — ledger recorder (`internal/jevledger`)

RED (EXPECTED_RED, with one stated limit). Before M2 the cells E3, E4, E5, E6, E7 exited 1 (start tree, listed under M1). A signatures-only skeleton of `ledger.go` (a recorder whose `Append` returns `Result{Reason: "not implemented"}`) was added so the tests compile, then:

```
$ go test ./internal/jevledger/ -count=1 -timeout 60s -v   -> exit=1, 8 FAIL lines
ledger_test.go:149: positive control failed: the append did not land (not implemented)
ledger_test.go:188: append did not land: not implemented
ledger_test.go:288: append 0 dropped: not implemented
FAIL: TestLedger_NothingUntilAppend, TestLedgerRow_Fields, TestLedgerRow_UnansweredNullProbability, TestLedgerRow_RecorderOwnsActionAndOutcome, TestLedgerRow_Threshold, TestLedger_ConcurrentAppend, TestLedger_Modes, TestLedger_AppendAfterMalformedLines
```

The limit: the fail-closed cases (`TestLedger_WriteFailureDropped`, `TestLedger_NonRegularPathDropped/*`, `TestLedger_FIFODropped/*`) and the import-boundary tests PASSED against that skeleton, because a recorder that never writes satisfies every "nothing was written" assertion. That is the never-writes mutant of acceptance.md AC-JEVL-001, and it is why each of those tests is only meaningful together with the mutant table below, where each one went red against a real defect. Their RED-now evidence is the §L cells (E6, E7: the files did not exist) plus those mutants.

GREEN (PASS). `go test ./internal/jevledger/ -count=1 -timeout 120s -v` -> exit 0. Named PASS lines: `--- PASS: TestLedger_NothingUntilAppend`, `--- PASS: TestLedgerRow_Fields`, `--- PASS: TestLedgerRow_UnansweredNullProbability`, `--- PASS: TestLedgerRow_RecorderOwnsActionAndOutcome`, `--- PASS: TestLedgerRow_Threshold`, `--- PASS: TestLedger_ConcurrentAppend`, `--- PASS: TestLedger_Modes`, `--- PASS: TestLedger_AppendAfterMalformedLines`, `--- PASS: TestLedger_WriteFailureDropped`, `--- PASS: TestLedger_NonRegularPathDropped` (+ `/symlink`, `/directory`), `--- PASS: TestLedger_FIFODropped` (+ `/no-reader`, `/with-reader`), `--- PASS: TestLedger_UnencodableRowDropped`, `--- PASS: TestLedgerPackage_ImportBoundary` (+ `/recorder-package`, `/config-resolver-file`, `/classifier-controls`, `/mutants` with 13 subtests), `--- PASS: TestLedgerPackage_NeverReadsLedger`. Existing guard run unchanged: `go test ./internal/cli/ -count=1 -run TestJevCallPath_HasExactlyTheDeclaredConsumers` -> `--- PASS: TestJevCallPath_HasExactlyTheDeclaredConsumers`, exit 0. On windows the FIFO cases are absent by construction: `fifo_other_test.go` carries `TestLedger_FIFOCasesUnsupportedHere` (a named SKIP), which has not been executed on windows here (Gap).

Mutants run against the implementation, each reverted afterwards (`diff` against the saved copies printed nothing):

| AC | Mutant | Result |
|---|---|---|
| 009-1 | empty `model_id` filled from `request_model` | RED: `TestLedgerRow_UnansweredNullProbability`, `TestLedgerRow_Fields` |
| 009-2 | caller's `action` / `outcome_label` pass through | RED: `TestLedgerRow_RecorderOwnsActionAndOutcome` |
| 009-3 | absent probability written as 0 | RED: `unanswered probability = 0 (present=true), want null` |
| 009-4 | `threshold` tagged `omitempty` / removed from the output | RED: `row 2 lacks the key "threshold"` / all three rows lack it |
| 009-5 | absent threshold written as 0 | RED: `unanswered row threshold = 0 (present=true), want null` |
| 009-6 | supplied threshold 0 written as null | RED: `choice row threshold = <nil> (present=true), want the number 0` |
| 012-1 | drop printed to stderr | RED: every drop test (`a drop must be silent`) |
| 012-2 | open without `O_NONBLOCK` | RED: `FIFODropped/no-reader`: `Append did not return within 2s`; the pre-registered cleanup released the goroutine and the binary exited |
| 012-3 | no `O_NOFOLLOW`, no handle check | RED: `FIFODropped/with-reader` (359 bytes reached the pipe) and `NonRegularPathDropped/symlink` (link target changed) |
| 012-4 | keeps `O_NOFOLLOW`/`O_NONBLOCK`, drops the handle check | RED in `FIFODropped/with-reader` ONLY (359 bytes reached the pipe); (i), (ii), (iii), (v) stayed green, as the audit predicted |
| 012-5 | no `O_NOFOLLOW`, keeps the handle check | RED in `NonRegularPathDropped/symlink` only |
| 012-6 | FIFO fixture in an untagged `_test.go` file | RED: `GOOS=windows go vet ./internal/jevledger/` -> `undefined: syscall.Mkfifo`, exit=1 |
| 012-7 | a directory at the ledger path used as the ledger directory | RED in `NonRegularPathDropped/directory` only (the row landed inside it) |
| 010 | row written with two `Write` calls | RED: `line 24 is not JSON` (3 of 3 runs) |
| 018-1 | open with `O_TRUNC` | RED: `TestLedger_AppendAfterMalformedLines` and 5 more |
| 018-2 | row dropped when the existing ledger is not valid JSON | RED: `the row was dropped on account of the malformed lines` |
| 001-1 | `MkdirAll` in `New` | RED: `constructing a recorder created something under the root` |
| 001-2 | `Append` reports success without writing a row | RED: `line 0 is not JSON: ... ""` |
| 014 (a) | recorder imports `net/rpc` / `os/exec` / `gopkg.in/yaml.v3` | RED each: `(a) ledger.go imports ...` |
| 014 (a) | recorder calls `os.Open(` | RED: `ledger.go contains [os.Open(]` |
| 014 (b) | `jev_consumers.go` imports `net/smtp` | RED: `(b) jev_consumers.go imports net/smtp` |

Not run as real source edits: a recorder importing `internal/jev`, and a third-party HTTP client in either file. The run environment's worktree guard refused the shell loop that carried those module paths (a refused command, reported as a Gap). Those cases are covered by file-based fixtures inside `TestLedgerPackage_ImportBoundary/mutants` (`internal-jev`, `third-party-http-client`, `resolver-internal-jev`, `resolver-third-party-http-client`), which write each import into a real `.go` file and run the scanner on it.

Other M2 verification: `gofmt -l` empty; `go vet` exit 0 for `./internal/jevledger/ ./internal/config/`; `GOOS=windows go vet ./internal/jevledger/ ./internal/config/` exit 0; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run ./internal/jevledger/ ./internal/config/` -> `0 issues.`; `go test ./internal/jevledger/ -race -count=5` exit 0; `go test -cover ./internal/jevledger/` -> `coverage: 88.2% of statements` (darwin build set; `open_other.go` is not compiled here, so it is unmeasured); `grep -rn AskUserQuestion internal/jevledger internal/config/jev_consumers.go` -> no output, exit 1.

Disposition of the binding run condition **N1** — `internal/jevledger/boundary_test.go` is stricter than AC-JEVL-014's wording without changing it. (a) The recorder package is checked by an allowlist (standard library only: the first path element has no dot) plus a denylist of `net`, every `net/*`, and `os/exec`. (b) `internal/config/jev_consumers.go` is checked by the same denylist plus the allowlist "standard library or `gopkg.in/yaml.v3`". The rejection cases are recorded as an executable mutant probe, `TestLedgerPackage_ImportBoundary/mutants`: net/http, net/rpc, net/smtp, os/exec, internal/jev, a third-party HTTP client and (for the recorder only) the yaml library, each written into a real file and required to produce exactly one finding; the one allowed non-standard import (yaml in the resolver file) is required to produce none. `classifier-controls` pins the classifier itself. The read-back check is lexical, as the AC says.

Other decisions recorded at M2:

- `Append` returns `Result{Landed bool; Reason string}`; `Reason` is empty when the row landed. `Row` carries `Action` and `OutcomeLabel` as inputs that the recorder overwrites (`OutcomeLabel` always `unlabeled`; `Action` is `none` under `mode == "shadow"` or when empty, otherwise the caller's value passes through).
- `ts` is written as the caller supplies it; the recorder does not stamp a time.
- The boundary test sits in `internal/jevledger/boundary_test.go`, a seventh test-side file beyond plan §E's six for M2 (it also reads `../config/jev_consumers.go`).
- `@MX:ANCHOR` + `@MX:REASON` on `Recorder.Append`; `@MX:WARN` + `@MX:REASON` on the Unix `openLedger`.

### Run-phase repair after the first sync audit (FAIL 0.73, audited_sha 5a3579e310)

The first sync audit (`.moai/reports/t1762/sync-audit-iter1.md`: audited_sha 5a3579e31009fc2e7b22570b6af59645912afbe4, FAIL 0.73, receipt rcpt-d80490b6aba4a5df1617e807) failed the Security must-pass dimension on three reproduced defects (F1-F3). The second, delta verdict (`.moai/reports/t1762/sync-audit-iter2.md`: audited_sha 6df6a76d7e2cb57d3bde0692faea99fd7648f88e, PASS-WITH-DEBT 0.88, receipt rcpt-73f00376213086df4f2c9936) is the one that closed F1-F3. This repair changes only what they need; the SPEC artifacts and the plan are untouched.

- **F1 (High) and F2 (Medium), fixed in `internal/config/jev_consumers.go`.** `JevConsumerConfig.UnmarshalYAML` no longer stores the original node of a `mode` or `threshold` value. New helper `jevRetain` keeps a scalar in full (kind, style, tag, value) and any other kind as its kind alone, with no content, anchor, or alias link. The resolver reads only the resolved tag, the value, and the kind, so its observable behaviour is unchanged: warnings are returned as data, every mode runs as shadow, and one malformed consumer is quarantined without resetting the workflow section. Removing the alias graph closes F1 (the config cache's JSON encoding followed it: 479 bytes of YAML became 122,690,481 bytes) and F2 (`MarshalYAML` wrote `mode` before `threshold`, so an anchor defined in `threshold` was emitted after an alias to it, and the saved file failed to reload with `unknown anchor`). Tests: `TestJevConsumers_AliasBombStaysBoundedInCache` (through `LoadWithCache`, bounded against a benign malformed config of the same shape, then a cache-HIT reload) and `TestJevConsumers_SaveThenReloadWithAnchors` (the `saveSection` round trip; zero log records, settings and both consumers kept, no anchor or alias in the saved text).
- **Deviation from plan.md (not edited here).** Plan section B describes `JevConsumerConfig` as holding "raw `yaml.Node`" fields, and section D says the resolver interprets the raw nodes. The fields are still `yaml.Node` values, but they now hold the retained form above: a scalar copy, or a kind-only placeholder for a mapping, sequence, or alias. The raw original is what F1 and F2 exposed, so the wording in the plan is superseded for the non-scalar case. The earlier M1 bullets in this section that say "raw nodes" describe the pre-repair state.
- **F3 (Medium), fixed in `internal/jevledger/boundary_test.go`.** The import classifier trimmed only double quotes, so an import spelled with backticks (for example ``import _ `net/http` ``) produced no finding in either the recorder-package check or the resolver-file check. The path is now resolved by `strconv.Unquote` in `importFinding`, which handles both quote styles, and a literal that cannot be resolved is itself a finding. `TestLedgerPackage_ImportBoundary/mutants` gains seven backtick cases (`raw-*`, each required to yield exactly one finding), the allowed yaml import is checked in both spellings, and `unparsable-literal-is-a-finding` pins the unresolvable-literal rule. This corrects the N1 disposition above: the statement that `net/rpc`, `net/smtp` and `os/exec` are refused held for the double-quoted spelling only until this fix.
- **Cache schema.** `configCacheSchemaVersion` stays 14: the retained node is a subset of the previous JSON shape, so a cache written by the earlier build still unmarshals, and the schema bump this card already made has not shipped.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete-pending-audit
run_complete_at: 2026-10-27
run_commit_sha: fe227a31c1497c770b9c3157048ba24a1bfa730f   # M2; M1 is 07fc4b93f3a444f24ccd28c9789216402569c7f8
ac_pass_count: 10 (AC-001, 006, 007, 009, 010, 012, 013, 014, 017, 018) by the named PASS lines in section E.2; AC-011 is a regression-guard (cell E10 exit 0), not counted as a pass
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE list; internal/jev and internal/cli have no changed file
l44_pre_commit_fetch: "git fetch origin main exit 0; git rev-list --count --left-right origin/main...HEAD -> 1 8 (origin/main 84ea2ffa35 is one docs-only commit touching factory-dispatch rules and AGENTS.md; no overlap with this card's files; not absorbed here)"
l44_post_push_fetch: not-applicable   # this run does not push
new_warnings_or_lints_introduced: 0   # golangci-lint on both packages: 0 issues
cross_platform_build:
  go_build_all: exit 0
  goos_windows_goarch_amd64_go_build_all: exit 0
  goos_windows_go_vet_jevledger_and_config: exit 0
  fifo_cases_on_windows: not executed here (named SKIP present in fifo_other_test.go)
repair_commit_shas:   # run-phase repair after the first sync audit (FAIL 0.73); not milestone commits
  f1_f2: 57c4aa64dfc33f172d5309629250218bf2902a22   # retain no alias graph in a consumer's mode and threshold
  f3: c56087c2d0939f84ac5fd0806f313d270e1235f5      # resolve import paths with strconv.Unquote in the boundary test
total_run_phase_files: 15 (git diff --stat ff7a0a0875 HEAD at the M2 commit; +1917 -9 lines; plan section E estimated 11 distinct files, ~770 lines; the extra files are progress.md, the local workflow.yaml, cache.go and boundary_test.go). After the two repair commits the file count is unchanged at 15 (the union of files in `git show --name-only --format= 07fc4b93f3 fe227a31c1 57c4aa64df c56087c2d0` is the same 15; the repairs touched only jev_consumers.go, jev_consumers_test.go, boundary_test.go and progress.md) and the line total is +2162 -37 (sum of `git show --shortstat` for the four commits: +845 -8, +1072 -1, +175 -5, +70 -23). The brief's range `git diff --name-only 55b20438fd..HEAD -- internal .moai/specs` lists 26 files because it also spans the plan-phase commits and commits absorbed from origin/main, so it is not the run-phase set.
m1_to_mN_commit_strategy: one commit per milestone (M1 07fc4b93f3, M2 fe227a31c1), then one docs commit that backfills this SHA; the two repair commits above followed the first sync audit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete   # 3-phase close; sync audit iteration 3 PASS-WITH-DEBT 0.89 @5418ffa31ab3 (receipt rcpt-dc38b1a966aa78915098cf7d)
sync_complete_at: 2026-10-27
sync_commit_sha: d4eff5d8a31619abd4f83fe1d22d05be9f828905   # the single sync commit; backfilled by the lane in the next commit (a commit cannot cite its own hash)
sync_scope: auto                    # sync-coverage-scope-contract.md
coverage_scope: SPEC-JEV-LEDGER-001 only (CHANGELOG.md [Unreleased] entry, spec.md frontmatter status/updated, this section); no code, no docs-site, no README change (no user-facing documentation surface is touched by this SPEC)
tier: M
ac_source: .moai/specs/SPEC-JEV-LEDGER-001/acceptance.md (resolved, non-empty)
ac_headings_declared: 11   # grep -n '^### AC-JEVL' acceptance.md; no [RETIRED]/[REF] token present (grep -c = 0); the B12 awk counter was not run (worktree guard forbids awk in this brief) - Gap
b12_self_test_a: pass   # grep -c 'SPEC-JEV-LEDGER-001' CHANGELOG.md printed 0 before emission (exit 1 = no match)
b12_self_test_b: pass-by-headings   # CHANGELOG states 11 AC headings = the 11 '### AC-JEVL' headings; counter itself not run (see Gap)
b12_self_test_c: pass   # every path named in the CHANGELOG entry listed by ls (internal/jevledger/{ledger.go,open_unix.go,open_other.go,ledger_test.go,fifo_unix_test.go,fifo_other_test.go,boundary_test.go}, internal/config/{jev_consumers.go,jev_consumers_test.go,cache.go,defaults.go}); configCacheSchemaVersion = 14 at cache.go:57
changelog_entry_position: "CHANGELOG.md ## [Unreleased] -> ### Added, first entry"
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed on this single sync commit (status and updated only)
  progress_md_status_line: draft -> completed
  plan_md_acceptance_md: not touched (no status field; updated: left as authored)
canary_compliance_check: not-applicable   # this SPEC defines no forward-looking policy that its own sync tests
mx_validation: not-run   # sync sub-step not exercised here; the run phase recorded @MX:ANCHOR on Recorder.Append and @MX:WARN on the Unix openLedger (E.2)
lane_run_evidence_not_reobserved_by_sync: "at ef620b8f8d the lane re-ran go test ./internal/jevledger/ ./internal/config/ -count=1 (ok), go vet and GOOS=windows go vet on both (exit 0), gofmt -l (empty)"
sync_audit_iterations: "iteration 1 .moai/reports/t1762/sync-audit-iter1.md (5a3579e310, FAIL 0.73); iteration 2 .moai/reports/t1762/sync-audit-iter2.md (6df6a76d7e2cb57d3bde0692faea99fd7648f88e, PASS-WITH-DEBT 0.88, receipt rcpt-73f00376213086df4f2c9936); iteration 3 .moai/reports/t1762/sync-audit-verdict-iter3.md (5418ffa31ab344a4c8cd3ce8cd461b0496708f24, PASS-WITH-DEBT 0.89, receipt rcpt-dc38b1a966aa78915098cf7d)"
sync_audit_debts_remaining: "run-phase D1-D7 accepted by the leader as run debts for a follow-up card; sync-phase D8-D10 disposed by this edit (D8 line anchors and wording re-measured below, D10 trailing attribution removed from the CHANGELOG entry)"
```

### Disposition of the binding run conditions

Both conditions in `## Binding run conditions` were disposed in the run phase; the sync phase re-read the cited artifacts and confirmed each path and name below exists by `grep` / `ls` in this tree (existence only; the tests were not re-run by the sync phase).

- **N1 (dispose_in=run) — DISPOSED.** Run-phase disposition: §E.2 "Disposition of the binding run condition **N1**". `internal/jevledger/boundary_test.go` is stricter than AC-JEVL-014's wording without changing it: the recorder package is checked by a standard-library allowlist plus a `net`, `net/*`, `os/exec` denylist; `internal/config/jev_consumers.go` by the same denylist plus the allowlist "standard library or `gopkg.in/yaml.v3`". The rejection cases are the executable mutant probe `TestLedgerPackage_ImportBoundary/mutants` (net/http, net/rpc, net/smtp, os/exec, internal/jev, a third-party HTTP client, and the yaml library for the recorder only). The import spellings are resolved with `strconv.Unquote`, so a double-quoted and a backtick-quoted import path are both refused (commit `c56087c2d0`; the earlier N1 statement held for the double-quoted spelling only). Verified present by `grep -n` after the F1-F3 repair: `func TestLedgerPackage_ImportBoundary` at `boundary_test.go:104`, subtests `classifier-controls` (`:131`) and `mutants` (`:178`), the `net-rpc` case (`:185`), and `func TestLedgerPackage_NeverReadsLedger` (`:252`). Line numbers are as measured at this sync; the symbol names are the stable anchors.
- **N4 (dispose_in=run) — DISPOSED.** Run-phase disposition: §E.2 "Dispositions and decisions recorded at M1", first bullet (**N4**). `JevConsumerConfig` (its `Mode` and `Threshold` are `yaml.Node` fields that hold a retained form: a scalar in full, any other node kind as a kind-only node with no content, anchor or alias link; helper `jevRetain`, `:228`) is declared in `internal/config/jev_consumers.go` (`type JevConsumerConfig struct` at `:43`, beside the `UnmarshalYAML` methods of `JevConsumerConfig` at `:73` and of `JevConsumers` at `:96`), so `internal/config/types.go` gains only the `Consumers JevConsumers` field and imports no yaml library, and plan section B's "types.go imports no yaml library" stays true. The run phase's own check was `grep -n 'yaml.v3\|"gopkg' internal/config/types.go` (two comment lines, no import line); the sync phase did not re-run it.

Gaps (sync phase): the B12 awk counter was not run (brief forbade awk), so the AC count rests on the heading grep; the repository-wide suite, `moai spec audit` and the lane evidence at `ef620b8f8d` were not re-observed here; `mx_validation` was not exercised; Windows execution of the FIFO cases remains unobserved (carried from §E.2).
