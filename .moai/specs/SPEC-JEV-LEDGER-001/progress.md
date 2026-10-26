# progress — SPEC-JEV-LEDGER-001

status: draft

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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
