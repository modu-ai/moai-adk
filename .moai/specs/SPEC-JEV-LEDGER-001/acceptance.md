---
id: SPEC-JEV-LEDGER-001
title: "Acceptance — Jev decision ledger, shadow-mode hook adapter, cost report"
version: "0.1.2"
created: 2026-10-26
updated: 2026-10-26
---

# Acceptance — SPEC-JEV-LEDGER-001

**Document-level tree pin:** every RED-now cell below was measured on `d1aa4ed656af948c95c2567718c4f9a7be54e191` unless the cell carries its own pin. RED-now cells cite an evidence-ledger entry (§L) that holds the command, the verbatim stdout, and the exit code.

Each criterion has the four parts the rule requires: the Given-When-Then, the RED-now cell (§L), the green path (milestone + passing output), and the three-part check spec: (a) WHEN, (b) failing INPUT, (c) who sees the red. The continued-firing answer (§1.3) is given once per check family in §C.

**Partial vs final.** A green path that spans milestones names a **partial** milestone and a **final** milestone (plan.md §E, AC closure map). At the partial milestone, the named subset of tests must PASS and is recorded in `progress.md` §E.2, but the AC is **not** reported PASS. The AC is reported PASS only at its final milestone, with every test in its green path passing on the same tree.

---

## §A. Criteria

### AC-JEVL-001 — gate off: zero requests, zero rows (REQ-JEVL-001) [release-blocking]

- **Given** `workflow.jev.enabled` is false, a fake consumer registered through the test seam with a counting request builder, the adapter's `Ask` seam (plan.md §B) wrapped by a counter around the real `jev.Client`, and a counting transport.
- **When** each of `pre-tool`, `stop`, and `subagent-stop` is dispatched once under `claude` and once under `codex`.
- **Then** the fake consumer's request-builder call count is 0, the `Ask` call count on the adapter's `jev.Client` is 0, the transport call count is 0, and no `.moai/state/jev/` directory exists under the test project root. Positive control in the same test: the same fixture with the gate on shows builder calls ≥ 1 and `Ask` calls ≥ 1.
- RED-now: §L E1 (`ls internal/cli/hook_jev.go` → absent, exit 1; re-collected at iter-3 on `82aff00a6d`). Red because the adapter, the builder seam, and the `Ask` seam do not exist, so there is nothing yet to keep at zero.
- Green path: **final at M3** (`--- PASS: TestHookJev_GateOff_ZeroRequests `). No partial milestone: the iter-2 M2 partial extended `TestDisabled_ConstructsNoRequest`, which calls only `Ask`, and `Ask` never writes a ledger, so its no-ledger assertion could not fail. It was removed (plan.md §E).
- Check spec: (a) runs in `go test ./internal/cli/` on every CI run, after a consumer is registered. With no consumer registered it would pass vacuously, so the test registers one first. (b) failing input: a mutant adapter that skips the gate check → builder call count 1 → red. The transport count alone cannot catch this mutant, because `Ask` itself returns `Disabled` before sending (`internal/jev/jev.go:538-544`). That is why the builder and `Ask` counts are asserted. (c) `go test` exit 1 in the required Test job.

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
- **When** the dispatch runs once with the gate off (baseline), then once with the gate on for **each configured mode** `shadow`, `advise`, `steer`, and `confirm`, under each harness.
- **Then** for every one of the 24 cases (4 modes × 3 events × 2 harnesses), the captured stdout bytes, stderr bytes, and exit code equal the gate-off baseline of the same event and harness (`bytes.Equal` on stdout and stderr, `==` on the exit code). Each gated-on case first satisfies the answered-path precondition (§C): transport calls ≥ 1 and ledger rows ≥ 1.
- RED-now: §L E1.
- Green path: final at M3. `--- PASS: TestHookJev_ShadowByteIdentical/* ` (24 subtests, named `<mode>/<event>/<harness>`). The run must show 24 subtest PASS lines. A lower count is a gap, not a pass.
- Mutant probes: (1) an adapter that prints `[model signal] ...` to stderr satisfies "an answer is recorded" but fails this criterion, which is why the comparison covers full stderr. (2) A call site that prints an advisory, or requests confirmation, when the configured mode is `advise` or `confirm` stays green under a `shadow`-only test. It fails here because the non-shadow modes are run at the hook level, not only in the resolver.
- Check spec: (a) CI, with an answered call (an unanswered call would make byte identity trivially true). (b) failing input: either mutant above. (c) Test job exit 1.

### AC-JEVL-005 — adapter runs after the output is fixed (REQ-JEVL-004)

- **Given** a fake consumer whose request builder records the moment it runs, and a test writer that records the moment the hook output is flushed.
- **When** a gated-on dispatch runs.
- **Then** the builder's moment is after the output flush and after the exit decision.
- RED-now: §L E1.
- Green path: M3. `--- PASS: TestHookJev_RunsAfterOutput `.
- Check spec: (a) CI. (b) mutant: adapter invoked before `writeHookOutput` → order inverted → red. (c) Test job.

### AC-JEVL-006 — mode enum and clamp (REQ-JEVL-006)

- **Given** consumer configs with `mode` = `shadow`, `advise`, `steer`, `confirm`, absent, and `bogus`.
- **When** (i) the resolution function runs, and (ii) a gated-on dispatch with an answered fake call runs for each of the six configs, under each harness.
- **Then** (i) the effective mode is `shadow` in all six cases, the configured value is preserved verbatim, and only `bogus` yields a warning. (ii) Every ledger row carries `mode: "shadow"`, `action: "none"`, and `configured_mode` equal to the configured value (empty when absent), and the hook output equals the gate-off baseline (as in AC-JEVL-004).
- RED-now: §L E2 (`grep -n Consumers internal/config/types.go` → no output, exit 1). Red because no consumer config type exists.
- Green path: **partial at M1**: `--- PASS: TestResolveJevConsumer_Mode/* ` (6 subtests). **Final at M3**: `--- PASS: TestHookJev_CallSiteUsesEffectiveMode/* ` (12 subtests: 6 configs × 2 harnesses).
- Mutant probes: (1) `advise` passed through as effective by the resolver → (i) red. (2) A hook call site that reads the raw configured mode and branches on it (writing `action: "advise"`, or emitting anything for `advise` / `steer` / `confirm`) while the resolver stays correct → (ii) red. Part (i) alone cannot catch this, which is why the final closure is at M3.
- Check spec: (a) CI. (b) either mutant. (c) Test job.

### AC-JEVL-007 — threshold bounds (REQ-JEVL-007)

- **Given** a temp project whose `.moai/config/sections/workflow.yaml` sets `workflow.jev.enabled: true`, one other non-default `workflow` setting, and seven consumers whose `threshold` is `0`, `1`, `0.5`, `-0.1`, `1.5`, `"high"` (a YAML string), and absent. One further consumer has a mapping in place of `threshold`.
  Separately, a second temp project whose `workflow.jev.consumers` is the scalar `"x"` instead of a mapping, with the same `enabled: true` and the same other non-default setting.
- **When** the file is loaded through the real loader (`internal/config/loader.go`, `loadWorkflowSection`) and each consumer is resolved, with stdout, stderr, and the default `slog` handler captured.
- **Then** `0`, `1`, and `0.5` are kept. `-0.1`, `1.5`, the string, and the mapping each resolve to absent, with a warning returned as data. Absent resolves to absent without a warning. `0` stays distinguishable from absent. `workflow.jev.enabled` and the other non-default setting keep their file values, not the defaults. All eight consumers are present in the loaded map. Captured stdout and stderr are empty, and the `slog` capture holds no record. For the scalar `consumers: "x"` project, the loaded consumer map is empty, exactly one section-level warning is returned as data, `enabled` and the other setting keep their file values, and stdout, stderr, and the `slog` capture are empty.
- RED-now: §L E2. Red because no consumer config type exists.
- Green path: final at M1. `--- PASS: TestResolveJevConsumer_Threshold/* `, `--- PASS: TestLoadWorkflow_MalformedConsumerQuarantined `, and `--- PASS: TestLoadWorkflow_NonMappingConsumersQuarantined `.
- Mutant probes: (1) an open-interval check rejecting `0` → red. (2) A plain `Threshold *float64` field. Here `yaml.Unmarshal` fails on `"high"`, and `loadWorkflowSection` logs `failed to load workflow config, using defaults` and keeps the defaults (`loader.go:239-241`). So `enabled` and the other setting revert, and the `slog` capture is non-empty → red. A resolver-only unit test passes this mutant, which is why the test goes through the loader. (3) A typed `map[string]JevConsumerConfig` for `consumers` fails on the scalar `"x"` the same way and reverts the whole section → the second project goes red.
- Check spec: (a) CI. (b) either mutant. (c) Test job.

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

### AC-JEVL-012 — write failure and non-regular ledger path are silent (REQ-JEVL-011, REQ-JEVL-012)

- **Given** three project roots, each with the budget and margin injected through the test seam (250 ms budget, 50 ms margin, so the deadline is process start + 200 ms):
  - (i) `.moai/state/jev` is a regular file, so `mkdir` fails.
  - (ii) `.moai/state/jev/decisions.jsonl` is a FIFO (`syscall.Mkfifo`) with no reader.
  - (iii) `.moai/state/jev/decisions.jsonl` is a symbolic link to a file outside the project root whose bytes are recorded before the dispatch.
- **When** a gated-on dispatch with an answered fake call runs in each root.
- **Then** in every case the write-failure precondition holds (§C): transport calls ≥ 1 and ledger rows = 0. In (i), no `decisions.jsonl` exists anywhere under the root. In (ii), the FIFO is still a FIFO and no row was written. In (iii), the link target's bytes equal the recorded bytes (zero rows appended) and the link is still a link. In every case, stdout, stderr, and exit code equal the gate-off baseline, no panic occurs, and the dispatch returns at or before process start + 200 ms + T, with **T = 50 ms** (the same T as AC-JEVL-013).
- **Cleanup guarantee (FIFO).** The FIFO lives under `t.TempDir()`, which the test framework removes. The dispatch runs in a goroutine, and the test waits on it with `select` against a timer of 200 ms + T + 1 s. On expiry the test fails instead of hanging. A `t.Cleanup` registered before the dispatch opens the FIFO's read side with `O_RDONLY|O_NONBLOCK` and closes it, so a mutant blocked in a write open is released and the goroutine ends. The package `go test -timeout` is the outer bound. No load runs outside these guards. Case (ii) is skipped on Windows and case (iii) is skipped where symlink creation fails, each with `t.Skip` and a named reason.
- RED-now: §L E3 (no row type) and §L E10 (`grep -rn O_NOFOLLOW internal/jev` → no output, exit 1; red because no write-side regular-file open exists).
- Green path: **partial at M2**: `--- PASS: TestLedger_WriteFailureDropped ` and `--- PASS: TestLedger_NonRegularPathDropped/* ` (FIFO and symlink, each calling the recorder directly). **Final at M3**: `--- PASS: TestHookJev_LedgerFailureSilent/* ` (3 subtests: `mkdir-fails`, `fifo`, `symlink`).
- Mutant probes: (1) the error returned and printed → stderr differs → red. (2) The plain `O_APPEND|O_CREATE|O_WRONLY` open with no `O_NONBLOCK`: case (ii) blocks, the `select` timer fires → red, and the cleanup releases the goroutine. (3) An open without `O_NOFOLLOW` and without a handle check: case (iii) appends a row to the link target → target bytes differ → red.
- Check spec: (a) CI on linux/darwin, with all three cases. (b) any mutant above. (c) Test job exit 1.

### AC-JEVL-013 — time budget exceeded → no answer, hook output unchanged (REQ-JEVL-013) [release-blocking]

- **Given** a fake transport that blocks until its context is cancelled and records the deadline of the context it receives. The budget and margin are injected through the test seam as 250 ms and 50 ms, with the same derivation the 5 s / 1 s constants use, so the deadline is process start + 200 ms. The process start instant is injected too. **T = 50 ms** is the return tolerance throughout this AC. Four cases:
  - (a) **timeout**: elapsed time at adapter entry ≈ 0.
  - (b) **intermediate**: 150 ms already elapsed before the adapter is entered (injected start = adapter-entry instant − 150 ms).
  - (c) **exhausted**: elapsed time at adapter entry already exceeds budget − margin.
  - (d) **exit-2 baseline**: case (a) with a hook handler whose output carries exit code 2, using the existing fault-registry fixture (`faultRegistry{output: &hook.HookOutput{ExitCode: 2}}`, `internal/cli/hook_fault_injection_test.go:181-189`), under both harnesses.
- **When** a gated-on dispatch runs for each of the three events under each harness, and the same dispatch runs with the gate off as the baseline.
- **Then:**
  - In (a), (b), and (d), the timeout precondition holds (§C): transport calls ≥ 1, and ledger rows equal the consumer's question count, each a no-answer row with `availability: "unreachable"`, `deadline_exceeded: true`, `probability: null`, and `model_id: ""`.
  - In (a) and (b), the deadline the transport received equals injected process start + 200 ms exactly, and the dispatch returns at or before injected process start + 200 ms + T.
  - In (a), (b), and (d), stdout bytes, stderr bytes, **and the exit code** equal the gate-off baseline of the same event and harness. In (d), that baseline exit code is 2, and the gated-on dispatch must also exit 2.
  - In (c), the skipped precondition holds: transport calls = 0, ledger rows = 0.
  - A separate assertion checks that the compiled constants are a 5 s budget and a 1 s margin, and that the derived deadline is the same value for all three events under both harnesses (no per-harness or per-event branch).
- RED-now: §L E6 (`grep -n JevHook internal/config/defaults.go` → no output, exit 1; re-collected at iter-3 on `82aff00a6d`). Red because no budget constant exists. The constant name is fixed by M1. If M1 picks a name without the `JevHook` stem, M1 re-points this cell's command to that name in the same commit.
- Green path: **partial at M1**: `--- PASS: TestJevHookBudgetConstants `. **Final at M3**: `--- PASS: TestHookJev_BudgetExceeded/* ` (6 subtests), `--- PASS: TestHookJev_BudgetIntermediateElapsed/* ` (6 subtests), `--- PASS: TestHookJev_BudgetExceededExit2/* ` (6 subtests), and `--- PASS: TestHookJev_BudgetExhaustedSkips `.
- Mutant probes: (1) `context.Background()` instead of the deadline → (a) blocks until the test timeout → red. (2) The deadline anchored at adapter entry instead of process start (entry + budget − margin, with a skip only when entry elapsed exceeds budget − margin): it passes (a) and (c), but in (b) the transport sees entry + 200 ms = start + 350 ms ≠ start + 200 ms, and the return lands near start + 350 ms > start + 250 ms → red. (3) The timeout path returns exit 0 or nil regardless of the handler's decision → (d) exits 0 ≠ baseline 2 → red.
- Check spec: (a) CI, with a transport that would otherwise block forever. (b) any mutant above. (c) Test job (the `go test` timeout surfaces as a FAIL).

### AC-JEVL-014 — boundary: declared importers, no new endpoint (REQ-JEVL-014) [release-blocking]

- **Given** the tree after M3/M4.
- **When** `TestJevCallPath_HasExactlyTheDeclaredConsumers` runs, together with a new assertion that `hook_jev.go` and `jev_cost.go` are in its allow-list with this SPEC's citation, and that neither file contains a URL literal or reads a credential directly.
- **Then** PASS. Any undeclared importer fails as before.
- RED-now: §L E7 (`grep -n hook_jev.go internal/cli/doctor_jev_test.go` → no output, exit 1).
- Green path: **partial at M3**: `--- PASS: TestJevCallPath_HasExactlyTheDeclaredConsumers ` with `hook_jev.go` asserted by name. **Final at M4**: the same test with both `hook_jev.go` and `jev_cost.go` asserted by name. Before M4 `jev_cost.go` does not exist, so the by-name assertion for it cannot pass.
- Mutant probe: routing the import through `internal/jevhook` keeps the original test green while evading it. The new by-name assertion catches that, because `hook_jev.go` must itself be an importer.
- Check spec: (a) CI. (b) the routing mutant. (c) Test job.

### AC-JEVL-015 — cost report (REQ-JEVL-015)

- **Given** a fixture ledger with 3 calls (`call_id` a, b, c; input_tokens 100, 200, 300, repeated across 2 rows each), of which 2 rows are labelled `correct`.
- **When** `moai jev cost` runs against the fixture.
- **Then** the output reports calls = 3, estimated input tokens = 600, correct = 2, and estimated input tokens per correct decision = 300. It contains the literal label `estimated input tokens` on each token figure and the literal heading `model signal aggregate`. It exits 0, and the counting transport records 0 calls. With zero correct rows, the per-correct field reads `undefined`. No figure is labelled as charged or billed.
- RED-now: §L E8 (`ls internal/cli/jev_cost.go` → absent, exit 1).
- Green path: M4. `--- PASS: TestJevCost_Aggregates `, `--- PASS: TestJevCost_ZeroCorrectUndefined `.
- Mutant probes: (1) summing tokens per row instead of per call gives 1200, which the test rejects. (2) A report that prints the token figures without the `estimated input tokens` label → red.
- Check spec: (a) CI. (b) the per-row mutant. (c) Test job.

### AC-JEVL-016 — absent / malformed ledger, and template default (REQ-JEVL-016, REQ-JEVL-002)

- **Given** (i) no ledger file, (ii) an empty one, (iii) one with a non-JSON line between two valid rows of two different calls (`call_id` a with `input_tokens: 100`, `call_id` b with `input_tokens: 200`, one row labelled `correct`). Separately, the template `workflow.yaml`. A counting transport is installed in every case.
- **When** `moai jev cost` runs, and the template is parsed.
- **Then** (i)/(ii) report calls = 0, exit 0. (iii) reports calls = 2, estimated input tokens = 300, correct = 1, tokens per correct decision = 300, and `skipped: 1`, exit 0. The valid rows are aggregated, not discarded. The counting transport records 0 calls in all three cases. The template carries `jev.enabled: false` and an empty `consumers` mapping.
- RED-now: §L E9 (`grep -n consumers internal/template/templates/.moai/config/sections/workflow.yaml` → no output, exit 1).
- Green path: **partial at M1**: `--- PASS: TestTemplateJevConsumersEmpty `. **Final at M4**: `--- PASS: TestJevCost_AbsentAndMalformed/* ` (3 subtests).
- Mutant probes: (1) the reader aborts on the first malformed line → exit 1 → red. (2) The reader reports zero calls once any line is malformed → calls = 0 in (iii) → red. (3) The reader drops the malformed line without counting it → `skipped:` missing → red.
- Check spec: (a) CI. (b) any of the three mutants. (c) Test job.

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
| Byte-identity and path tests (AC-001, -002, -004, -006, -008, -012, -013) | Each test first asserts the precondition of **the path it claims to exercise**, then compares bytes. A test that silently fell onto a different path fails on that precondition instead of passing vacuously. There are four paths, each with its own precondition, and no blanket precondition applies across them. **Answered** (AC-004, -006 part ii, -008): transport calls ≥ 1 **and** ledger rows ≥ 1, at least one row with non-null `probability`. **Write-failure** (AC-012): transport calls ≥ 1 **and** ledger rows = 0. **Timeout** (AC-013, timeout case): transport calls ≥ 1 **and** ledger rows = the question count, every row a no-answer row (`deadline_exceeded: true`, `probability: null`). This SPEC writes a no-answer row on timeout and never writes zero rows there (REQ-JEVL-013). **Skipped** (AC-001 gate off, AC-002 unregistered name, AC-013 exhausted case): transport calls = 0 **and** ledger rows = 0, paired in the same test with a positive control: the same fixture with the gate on, a registered consumer, and time left must show calls ≥ 1. That way the zero is attributable to the skip and not to a dead fixture. |
| Allow-list guard (AC-014) | The guard has a positive control (`doctor_jev.go` must be seen), so a scan that silently swept zero files fails. The by-name assertion on the new files fails if they are renamed or deleted. |
| Regression guard (AC-011) | Not a CI test. If it stopped being checked, nothing would differ. This is stated as a residual risk: the template `.gitignore` copy (`internal/template/templates/.gitignore:243: .moai/state/`) is the only other carrier. |

## §D. Quality gates and Definition of Done

- All release-blocking ACs (001, 002, 004, 009, 013, 014) PASS **at their final milestone** (plan.md §E closure map) with verbatim `--- PASS` lines in `progress.md` §E.2. A partial-milestone PASS line is evidence of progress, not of AC closure.
- M3 records the worst-case wall-clock time from wrapper-script start (bash start, `stat`, `command -v moai`, `exec moai hook <event>`) to process exit, with Jev blocked until its deadline, as at least 20 samples per event plus their p99 (plan.md §B). It returns a blocker report when that p99 exceeds 5 s − S, with S = 0.5 s (decision-index Q2). The 1 s margin remains a decision until this measurement exists.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test -cover` ≥ 85% for `internal/jev` and the new `internal/config` file. The new `internal/cli` files are covered by their tests (package-wide coverage is not re-baselined).
- `golangci-lint run` shows no NEW findings against the baseline.
- `make build` run after the template edit, with no tracked-file drift beyond the intended files.
- No user-facing question channel used anywhere in the new code. The C-HRA-008 grep returns 0.

---

## §L. Evidence ledger (RED-now cells)

**Tree:** `d1aa4ed656af948c95c2567718c4f9a7be54e191` (`git rev-parse HEAD` in card worktree `t1762`, 2026-10-26, iter-2 re-collection). Every entry below supersedes the iter-1 entries measured on `cb30e05b3e`.

**Collection method.** Each command was run exactly once. Its exit status was captured **inside the same Bash invocation** that ran it, by a wrapper of the form `sh -c '<command>; echo "exit=$?"'`. The wrapper printed the cited command's own status, not the status of a later, separate call. The **Command** column records the bare command, which is the single-invocation form a reader re-runs. The wrapper is the measuring harness, not part of the cited command. The **Exit** column is the value the wrapper printed. The **Verbatim stdout** column is the command's own output with the wrapper's `exit=N` line removed. One exception: E5 names `git`, and the worktree-isolation guard refused it inside `sh -c`. E5 was therefore run as the plain command, and its exit 0 is the tool's own success report for that call (a non-zero exit surfaces as an error in that same call). E5 is a regression-guard, not a release-blocking cell.

| Id | Command | Verbatim stdout | Exit | Wrapper output observed |
|---|---|---|---|---|
| E1 | `ls internal/cli/hook_jev.go` | *(stdout empty; stderr: `ls: internal/cli/hook_jev.go: No such file or directory`)* | 1 | `exit=1` |
| E2 | `grep -n Consumers internal/config/types.go` | *(empty)* | 1 | `exit=1` |
| E3 | `grep -rn outcome_label internal/jev` | *(empty)* | 1 | `exit=1` |
| E4 | `grep -rn decisions.jsonl internal` | *(empty)* | 1 | `exit=1` |
| E5 | `git check-ignore -v .moai/state/jev/decisions.jsonl` | `.gitignore:401:.moai/state/	.moai/state/jev/decisions.jsonl` | 0 | *(plain run; tool reported success)* |
| E6 | `grep -n JevHook internal/config/defaults.go` | *(empty)* | 1 | `exit=1` |
| E7 | `grep -n hook_jev.go internal/cli/doctor_jev_test.go` | *(empty)* | 1 | `exit=1` |
| E8 | `ls internal/cli/jev_cost.go` | *(stdout empty; stderr: `ls: internal/cli/jev_cost.go: No such file or directory`)* | 1 | `exit=1` |
| E9 | `grep -n consumers internal/template/templates/.moai/config/sections/workflow.yaml` | *(empty)* | 1 | `exit=1` |

Why each RED is red for the right reason: E1/E8 because the adapter and report files do not exist. E2/E6/E9 because no consumer type, no budget constant, and no template block exist. E3/E4 because no row type and no ledger path exist. E7 because the allow-list does not name the adapter yet. Each flips only through a file this SPEC's milestones create or edit.

**Iter-3 re-collection (own pin: tree `82aff00a6d3b9d1b3dd6f9c2a6fc68ebf5b8e0d0`, `git rev-parse HEAD` printed in the same invocation as E1').** These cells cover the criteria whose content changed at iter-3: AC-001 (E1'), AC-012 (E3', E10), and AC-013 (E6'). The commands for E1', E3', and E6' are unchanged from E1, E3, and E6. They were re-run because the criteria changed. E10 is new. Collection method: the cited command followed by `; echo "exit=$?"` in one Bash invocation, with no `sh -c` wrapper. The printed `exit=N` is the cited command's own status.

| Id | Command | Verbatim stdout | Exit | Wrapper output observed |
|---|---|---|---|---|
| E1' | `ls internal/cli/hook_jev.go` | *(stdout empty; stderr: `ls: internal/cli/hook_jev.go: No such file or directory`)* | 1 | `exit=1` |
| E3' | `grep -rn outcome_label internal/jev` | *(empty)* | 1 | `exit=1` |
| E6' | `grep -n JevHook internal/config/defaults.go` | *(empty)* | 1 | `exit=1` |
| E10 | `grep -rn O_NOFOLLOW internal/jev` | *(empty)* | 1 | `exit=1` |

E10 is red because no write-side regular-file open exists in `internal/jev`. It flips at M2 through `ledger_open_unix.go` (plan.md §F).

The iter-1 baseline of `TestDisabled_ConstructsNoRequest` (tree `cb30e05b3e`, `--- PASS`) is no longer part of AC-JEVL-001. The iter-2 extension of that test was removed at iter-3, because it could not fail (AC-JEVL-001 Green path).
