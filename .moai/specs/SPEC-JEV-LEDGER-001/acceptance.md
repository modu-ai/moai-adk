---
id: SPEC-JEV-LEDGER-001
title: "Acceptance — Jev decision ledger, shadow-mode hook adapter, cost report"
version: "0.1.0"
created: 2026-10-26
updated: 2026-10-26
---

# Acceptance — SPEC-JEV-LEDGER-001

**Document-level tree pin:** every RED-now cell below was measured on `cb30e05b3e5f64774b8dabd87e80c24332f4551d` unless the cell carries its own pin. RED-now cells cite an evidence-ledger entry (§L) that holds the command, the verbatim stdout, and the exit code.

Each criterion has the four parts the rule requires: the Given-When-Then, the RED-now cell (§L), the green path (milestone + passing output), and the three-part check spec: (a) WHEN, (b) failing INPUT, (c) who sees the red. The continued-firing answer (§1.3) is given once per check family in §C.

---

## §A. Criteria

### AC-JEVL-001 — gate off: zero requests, zero rows (REQ-JEVL-001) [release-blocking]

- **Given** `workflow.jev.enabled` is false and a fake consumer is registered through the test seam with a counting transport.
- **When** each of `pre-tool`, `stop`, and `subagent-stop` is dispatched once under `claude` and once under `codex`.
- **Then** the transport call count is 0, no `.moai/state/jev/` directory exists under the test project root, and the existing `TestDisabled_ConstructsNoRequest` (extended in M2) also asserts that no ledger file was created.
- RED-now: §L E1 (`ls internal/cli/hook_jev.go` → absent, exit 1). Red because the adapter does not exist, so there is nothing yet to keep at zero.
- Green path: M3 (adapter) + M2 (extension). Passing output: `--- PASS: TestHookJev_GateOff_ZeroRequests ` and `--- PASS: TestDisabled_ConstructsNoRequest `, each with transport calls = 0.
- Check spec: (a) runs in `go test ./internal/cli/ ./internal/jev/` on every CI run, after a consumer is registered. With no consumer registered it would pass vacuously, so the test registers one first. (b) failing input: a mutant adapter that skips the gate check → calls = 1. (c) `go test` exit 1 in the required Test job.

### AC-JEVL-002 — shipped registry is empty (REQ-JEVL-002) [release-blocking]

- **Given** the production build (no test seam).
- **When** the gate is on and `workflow.jev.consumers` names `ghost`.
- **Then** zero requests and zero rows. A test asserts that the compiled registry length is 0.
- RED-now: §L E1 (adapter absent).
- Green path: M3. `--- PASS: TestHookJev_ShippedRegistryEmpty ` and `--- PASS: TestHookJev_UnregisteredNameIgnored `.
- Check spec: (a) every CI run. (b) mutant: a consumer added to the registry → length 1 → red. (c) Test job exit 1.

### AC-JEVL-003 — both harnesses, three events, one batched request (REQ-JEVL-003)

- **Given** gate on and a fake consumer asking two questions, accepting all three events.
- **When** each event is dispatched under `claude` and `codex`.
- **Then** each dispatch produces exactly one transport call carrying both questions, and two ledger rows sharing one `call_id`, with `event` and `harness` set correctly.
- RED-now: §L E1.
- Green path: M3. A table test with 6 rows, all PASS.
- Check spec: (a) CI. (b) mutant: one request per question → calls = 2 → red. (c) Test job.

### AC-JEVL-004 — shadow mode is byte-identical (REQ-JEVL-004, REQ-JEVL-005) [release-blocking]

- **Given** a fixed hook stdin payload per event, and the fake consumer returning a valid answer (`model: jev-1.13.0`, p = 0.91).
- **When** the dispatch runs once with the gate off (baseline) and once with the gate on in `shadow`, under each harness.
- **Then** the captured stdout bytes, stderr bytes, and exit code of the two runs are identical (`bytes.Equal`), for all 6 event×harness pairs.
- RED-now: §L E1.
- Green path: M3. `--- PASS: TestHookJev_ShadowByteIdentical/* ` (6 subtests).
- Mutant probe: an adapter that prints `[model signal] ...` to stderr satisfies "an answer is recorded" but fails this criterion. That is why the comparison is over full stderr, not stdout only.
- Check spec: (a) CI, with an answered call (an unanswered call would make byte identity trivially true). (b) failing input: the stderr-printing mutant above. (c) Test job exit 1.

### AC-JEVL-005 — adapter runs after the output is fixed (REQ-JEVL-004)

- **Given** a fake consumer whose request builder records the moment it runs, and a test writer that records the moment the hook output is flushed.
- **When** a gated-on dispatch runs.
- **Then** the builder's moment is after the output flush and after the exit decision.
- RED-now: §L E1.
- Green path: M3. `--- PASS: TestHookJev_RunsAfterOutput `.
- Check spec: (a) CI. (b) mutant: adapter invoked before `writeHookOutput` → order inverted → red. (c) Test job.

### AC-JEVL-006 — mode enum and clamp (REQ-JEVL-006)

- **Given** consumer configs with `mode` = `shadow`, `advise`, `steer`, `confirm`, absent, and `bogus`.
- **When** the resolution function runs.
- **Then** the effective mode is `shadow` in all six cases, the configured value is preserved verbatim, and only `bogus` yields a warning.
- RED-now: §L E2 (`grep -n Consumers internal/config/types.go` → no output, exit 1). Red because no consumer config type exists.
- Green path: M1. `--- PASS: TestResolveJevConsumer_Mode/* ` (6 subtests).
- Check spec: (a) CI. (b) mutant: `advise` passed through as effective → red. (c) Test job.

### AC-JEVL-007 — threshold bounds (REQ-JEVL-007)

- **Given** thresholds `0`, `1`, `0.5`, `-0.1`, `1.5`, a string, and absent.
- **When** resolved.
- **Then** `0`, `1`, and `0.5` are kept. `-0.1`, `1.5`, and the string resolve to absent with a warning. Absent resolves to absent without a warning. `0` stays distinguishable from absent.
- RED-now: §L E2.
- Green path: M1. `--- PASS: TestResolveJevConsumer_Threshold/* `.
- Check spec: (a) CI. (b) mutant: open-interval check rejecting `0` → red. (c) Test job.

### AC-JEVL-008 — invalid config is recorded, never printed (REQ-JEVL-008)

- **Given** gate on, consumer `fake` with `mode: bogus` and `threshold: 2`.
- **When** a dispatch runs.
- **Then** each row carries a non-empty `config_warning`, and the captured stdout and stderr equal the gate-off baseline.
- RED-now: §L E1.
- Green path: M3. `--- PASS: TestHookJev_InvalidConfigSilent `.
- Check spec: (a) CI. (b) mutant: warning printed to stderr → red. (c) Test job.

### AC-JEVL-009 — ledger row fields (REQ-JEVL-009, REQ-JEVL-010) [release-blocking]

- **Given** a fake answered call (`model: jev-1.13.0`, choice question p = 0.72, noul question p = 0.31), `rubric_version: r1`, mode shadow.
- **When** the recorder writes.
- **Then** each line decodes to an object holding `model_id: "jev-1.13.0"`, `request_model: "jev-latest"`, `rubric_version: "r1"`, the matching `question_id`, `probability` equal to the raw answer value, `mode: "shadow"`, `action: "none"`, and `outcome_label: "unlabeled"`, plus `ts`, `consumer`, `event`, `harness`, `availability`, `input_tokens`, and `call_id`. For an unanswered call, `model_id` is `""` and `probability` is `null`. `model_id` never equals `jev.ModelID`.
- RED-now: §L E3 (`grep -rn outcome_label internal/jev` → no output, exit 1). Red because no row type exists.
- Green path: M2. `--- PASS: TestLedgerRow_Fields ` and `--- PASS: TestLedgerRow_UnansweredNullProbability `.
- Mutant probe: a recorder that writes `model_id: jev.ModelID` satisfies "model_id present" but fails the never-the-alias clause.
- Check spec: (a) CI. (b) the alias mutant. (c) Test job.

### AC-JEVL-010 — append, permissions, no interleaving (REQ-JEVL-011)

- **Given** a temp project root.
- **When** 50 goroutines append one row each.
- **Then** the file has 50 lines, each parses as JSON. On POSIX the directory mode is 0700 and the file mode is 0600.
- RED-now: §L E4 (`grep -rn decisions.jsonl internal` → no output, exit 1).
- Green path: M2. `--- PASS: TestLedger_ConcurrentAppend `, `--- PASS: TestLedger_Modes ` (skipped on Windows, with the skip named).
- Check spec: (a) CI on linux/darwin. (b) mutant: buffered writer flushing partial lines → a parse failure. (c) Test job.

### AC-JEVL-011 — ledger path stays ignored (REQ-JEVL-011) [regression-guard]

- **Given** the repository `.gitignore`.
- **When** `git check-ignore -v .moai/state/jev/decisions.jsonl` runs.
- **Then** exit 0 with a matching rule.
- RED-now: none. §L E5 shows the path is **already** ignored (`.gitignore:401:.moai/state/`). By the two-cell discipline this criterion would be vacuous as a release gate, so it is classified as a **regression-guard** and is not counted as a pass for this work.
- Green path: unchanged. It stays green as long as `.moai/state/` stays ignored in both the repository and the template `.gitignore`.
- Check spec: (a) runs at plan time and in the M2 self-verification. (b) failing input: removing `.moai/state/` from `.gitignore` → exit 1. (c) the self-verification table.

### AC-JEVL-012 — write failure is silent (REQ-JEVL-012)

- **Given** a project root where `.moai/state/jev` is a regular file, so `mkdir` fails.
- **When** a gated-on dispatch with an answered fake call runs.
- **Then** stdout, stderr, and exit code equal the gate-off baseline, and no panic occurs.
- RED-now: §L E3.
- Green path: M2 (recorder drops) + M3 (adapter ignores). `--- PASS: TestLedger_WriteFailureDropped `, `--- PASS: TestHookJev_LedgerFailureSilent `.
- Check spec: (a) CI. (b) mutant: the error returned and printed → red. (c) Test job.

### AC-JEVL-013 — time budget exceeded → no answer, exit 0 (REQ-JEVL-013) [release-blocking]

- **Given** a fake transport that blocks until its context is cancelled, the budget injected through the test seam (scaled down, e.g. 200 ms, with the same derivation the 5 s / 10 s constants use), and a separate case where the elapsed time already exceeds the budget.
- **When** a gated-on dispatch runs under each harness.
- **Then** the dispatch returns within budget + a small tolerance, the exit code is 0, the stdout and stderr equal the gate-off baseline, and the row carries `availability: "unreachable"`, `deadline_exceeded: true`, `probability: null`, `model_id: ""`. In the exhausted-budget case, no transport call and no row occur. A separate assertion checks that the compiled constants equal 5 s (claude) and 10 s (codex).
- RED-now: §L E6 (`grep -n JevHook internal/config/defaults.go` → no output, exit 1). Red because no budget constant exists.
- Green path: M1 (constants) + M3 (deadline). `--- PASS: TestHookJev_BudgetExceeded/* `, `--- PASS: TestHookJev_BudgetExhaustedSkips `, `--- PASS: TestJevHookBudgetConstants `.
- Check spec: (a) CI, with a transport that would otherwise block forever. (b) mutant: `context.Background()` instead of the deadline → test timeout → red. (c) Test job (the `go test` timeout surfaces as a FAIL).

### AC-JEVL-014 — boundary: declared importers, no new endpoint (REQ-JEVL-014) [release-blocking]

- **Given** the tree after M3/M4.
- **When** `TestJevCallPath_HasExactlyTheDeclaredConsumers` runs, together with a new assertion that `hook_jev.go` and `jev_cost.go` are in its allow-list with this SPEC's citation, and that neither file contains a URL literal or reads a credential directly.
- **Then** PASS. Any undeclared importer fails as before.
- RED-now: §L E7 (`grep -n hook_jev.go internal/cli/doctor_jev_test.go` → no output, exit 1).
- Green path: M3 + M4. `--- PASS: TestJevCallPath_HasExactlyTheDeclaredConsumers `.
- Mutant probe: routing the import through `internal/jevhook` keeps the original test green while evading it. The new by-name assertion catches that, because `hook_jev.go` must itself be an importer.
- Check spec: (a) CI. (b) the routing mutant. (c) Test job.

### AC-JEVL-015 — cost report (REQ-JEVL-015)

- **Given** a fixture ledger with 3 calls (`call_id` a, b, c; input_tokens 100, 200, 300, repeated across 2 rows each), of which 2 rows are labelled `correct`.
- **When** `moai jev cost` runs against the fixture.
- **Then** the output reports calls = 3, input tokens = 600, correct = 2, tokens per correct decision = 300, a model-signal label, exit 0, and the counting transport records 0 calls. With zero correct rows, the per-correct field reads `undefined`.
- RED-now: §L E8 (`ls internal/cli/jev_cost.go` → absent, exit 1).
- Green path: M4. `--- PASS: TestJevCost_Aggregates `, `--- PASS: TestJevCost_ZeroCorrectUndefined `.
- Mutant probe: summing tokens per row instead of per call gives 1200, which the test rejects.
- Check spec: (a) CI. (b) the per-row mutant. (c) Test job.

### AC-JEVL-016 — absent / malformed ledger, and template default (REQ-JEVL-016, REQ-JEVL-002)

- **Given** (i) no ledger file, (ii) an empty one, (iii) one with a non-JSON line between two valid rows. Separately, the template `workflow.yaml`.
- **When** `moai jev cost` runs, and the template is parsed.
- **Then** (i)/(ii) report zero calls, exit 0. (iii) reports the valid rows plus `skipped: 1`, exit 0. The template carries `jev.enabled: false` and an empty `consumers` mapping.
- RED-now: §L E9 (`grep -n consumers internal/template/templates/.moai/config/sections/workflow.yaml` → no output, exit 1).
- Green path: M1 (template) + M4 (reader). `--- PASS: TestJevCost_AbsentAndMalformed/* ` and `--- PASS: TestTemplateJevConsumersEmpty `.
- Check spec: (a) CI. (b) mutant: the reader aborts on the first malformed line → exit 1. (c) Test job.

---

## §B. Edge cases covered above

- Unanswered calls: AC-009, AC-013.
- Refused-before-send calls (oversize, secret) record `input_tokens: 0` (plan §F). Covered by an AC-009 subtest.
- Windows: AC-010 permissions skip, named. The append path runs on all OSes in CI.
- A configured but uncompiled consumer: AC-002.

## §C. Continued-firing answers (§1.3)

| Check family | If it stopped firing tomorrow, what would be different? |
|---|---|
| Go tests in `internal/cli`, `internal/jev`, `internal/config` (AC-001..010, 012..016) | Each test name is unique and listed in `progress.md` §E.2 at run time. A selector change that drops one shows up as a lower test count in `go test -v` output, and the self-verification matrix requires each named test's `--- PASS` line. A missing line is a gap, not a pass. The required CI Test job runs the whole package, so a dropped selector has no effect there. |
| Byte-identity tests (AC-004, -008, -012, -013) | Each test first asserts that the answered path was taken (transport calls ≥ 1, ledger rows ≥ 1) before comparing bytes. A test that stopped exercising the answered path fails on that precondition instead of passing vacuously. |
| Allow-list guard (AC-014) | The guard has a positive control (`doctor_jev.go` must be seen), so a scan that silently swept zero files fails. The by-name assertion on the new files fails if they are renamed or deleted. |
| Regression guard (AC-011) | Not a CI test. If it stopped being checked, nothing would differ. This is stated as a residual risk: the template `.gitignore` copy (`internal/template/templates/.gitignore:243: .moai/state/`) is the only other carrier. |

## §D. Quality gates and Definition of Done

- All release-blocking ACs (001, 002, 004, 009, 013, 014) PASS with verbatim `--- PASS` lines in `progress.md` §E.2.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test -cover` ≥ 85% for `internal/jev` and the new `internal/config` file. The new `internal/cli` files are covered by their tests (package-wide coverage is not re-baselined).
- `golangci-lint run` shows no NEW findings against the baseline.
- `make build` run after the template edit, with no tracked-file drift beyond the intended files.
- No user-facing question channel used anywhere in the new code. The C-HRA-008 grep returns 0.

---

## §L. Evidence ledger (RED-now cells)

All entries were measured on tree `cb30e05b3e5f64774b8dabd87e80c24332f4551d`, card worktree `t1762`, 2026-10-26. Exit codes were observed in the same run. For the grep entries the exit code was read by a follow-up `echo "exit=$?"` in a separate measuring invocation, which is not part of the cited command.

| Id | Command | Verbatim stdout | Exit |
|---|---|---|---|
| E1 | `ls internal/cli/hook_jev.go` | *(stdout empty; stderr: `ls: internal/cli/hook_jev.go: No such file or directory`)* | 1 |
| E2 | `grep -n Consumers internal/config/types.go` | *(empty)* | 1 |
| E3 | `grep -rn outcome_label internal/jev` | *(empty)* | 1 |
| E4 | `grep -rn decisions.jsonl internal` | *(empty)* | 1 |
| E5 | `git check-ignore -v .moai/state/jev/decisions.jsonl` | `.gitignore:401:.moai/state/	.moai/state/jev/decisions.jsonl` | 0 |
| E6 | `grep -n JevHook internal/config/defaults.go` | *(empty)* | 1 |
| E7 | `grep -n hook_jev.go internal/cli/doctor_jev_test.go` | *(empty)* | 1 |
| E8 | `ls internal/cli/jev_cost.go` | *(stdout empty; stderr: `ls: internal/cli/jev_cost.go: No such file or directory`)* | 1 |
| E9 | `grep -n consumers internal/template/templates/.moai/config/sections/workflow.yaml` | *(empty)* | 1 |

Baseline of the existing test that AC-JEVL-001 extends (same tree): `go test ./internal/jev/ -run '^TestDisabled_ConstructsNoRequest$' -count=1 -v` → `--- PASS: TestDisabled_ConstructsNoRequest (0.00s)` … `ok  	github.com/modu-ai/moai-adk/internal/jev	0.143s`. This is green today on purpose: AC-JEVL-001's RED comes from the missing adapter (E1), not from this test.
