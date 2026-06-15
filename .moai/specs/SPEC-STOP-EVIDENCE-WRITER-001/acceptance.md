# Acceptance Criteria — SPEC-STOP-EVIDENCE-WRITER-001

> GEARS-format AC. Each criterion is verifiable by a concrete shell / grep / go-test command an auditor can run from the repo root.
> Status: draft. cycle_type=tdd — each AC is RED before its milestone and GREEN after.
> AC-SEW-009 is the **falsifiable gate-activation proof** (avoids the GATE-001 dormant-scaffold trap).

## D. AC Matrix

| AC | Maps to REQ | Severity | Summary |
|----|-------------|----------|---------|
| AC-SEW-001 | REQ-SEW-017, 003, 004 | MUST | evidence record assembled with correct Outcome/binary/PathKind per the §3 table |
| AC-SEW-002 | REQ-SEW-001, 002, 005 | MUST | Bash/Edit/Write PostToolUse events write a UsageRecord via the existing store (was a no-op before) |
| AC-SEW-003 | REQ-SEW-003, 007 | MUST | test pass/fail derived from command + observed result by a pure classifier |
| AC-SEW-004 | REQ-SEW-004, 009, 010 | MUST | path classified code-change vs docs-only by a pure classifier |
| AC-SEW-005 | REQ-SEW-005 | MUST | persisted through `telemetry.RecordSkillUsage`; NO new on-disk store created |
| AC-SEW-006 | REQ-SEW-006 | MUST | canonical multi-language test commands recognized (go/pytest/cargo/npm/jest/vitest...) |
| AC-SEW-007 | REQ-SEW-015 | MUST | C-HRA-008 — no AskUserQuestion/mcp__askuser in non-test, non-comment hook code |
| AC-SEW-008 | REQ-SEW-008, 011 | MUST | non-test Bash + unknown-extension path produce NO binary flag / NO false code-change claim |
| AC-SEW-009 | REQ-SEW-017, 001, 002 | MUST | **FALSIFIABLE ACTIVATION** — given writer records, the GATE-001 gate emits a finding (code-change + test-fail) and emits NONE (test-pass); IsTestPass flip is the only delta |
| AC-SEW-010 | REQ-SEW-012, 013, 014, 016 | MUST | behavior preservation + fail-open + ≤5s budget + SessionID correlation |

REQ↔AC bidirectional coverage: REQ-SEW-001→AC-002/009; 002→AC-002/009; 003→AC-003; 004→AC-004; 005→AC-005; 006→AC-006; 007→AC-003; 008→AC-008; 009→AC-004; 010→AC-004; 011→AC-008; 012→AC-010; 013→AC-010; 014→AC-010; 015→AC-007; 016→AC-010; 017→AC-001/009. Every REQ-SEW-001..017 maps to ≥1 AC; every AC maps to ≥1 REQ.

---

## AC-SEW-001 — evidence record assembled per the §3 Outcome table

| Field | Value |
|-------|-------|
| **Given** | M2 is complete; `buildEvidenceRecord(input) (telemetry.UsageRecord, bool)` exists. |
| **When** | The assembler receives each event class (Bash test-pass, Bash test-fail, Bash ambiguous, Edit `.go`, Write `.md`, non-test Bash, unknown-ext Edit). |
| **Then** | It returns the record per design.md §3: test-pass→{Outcome=success, IsTestPass=true}; test-fail→{Outcome=error, IsTestFail=true}; ambiguous→{Outcome=unknown, neither}; Edit `.go`→{Outcome=success, PathKind=code-change}; Write `.md`→{Outcome=unknown, PathKind=docs-only}; non-test Bash→ok=false; unknown-ext→ok=false. |
| **Evidence** | `go test ./internal/hook/ -run 'TestBuildEvidenceRecord' -v -count=1` shows all subtests passing (one per row of the §3 table). |
| **Pass criterion** | Every §3-table row's subtest passes; the code-change Edit subtest yields `Outcome=success` + `PathKind=code-change` (the success-claim hinge per REQ-SEW-017). |

---

## AC-SEW-002 — Bash/Edit/Write events now write a UsageRecord (was a no-op)

| Field | Value |
|-------|-------|
| **Given** | M3 is complete; `logEvidence` is wired into `postToolHandler.Handle`. |
| **When** | A PostToolUse event with `ToolName=Bash` (recognized test command) is dispatched through `Handle`, in a temp project root. |
| **Then** | A `UsageRecord` is appended to `<root>/.moai/evolution/telemetry/usage-YYYY-MM-DD.jsonl` and is returned by `telemetry.LoadBySession(root, sessionID)` — whereas before this SPEC the same event wrote nothing. |
| **Evidence** | `go test ./internal/hook/ -run 'TestHandle_RecordsEvidenceOnBashEditWrite' -v -count=1` passes: it dispatches a Bash test event through `Handle`, then asserts `LoadBySession` returns ≥1 record with the expected evidence fields. AND `grep -n 'logEvidence' internal/hook/post_tool.go` shows the call wired in `Handle`. |
| **Pass criterion** | The integration test proves a Bash/Edit/Write event now produces a loadable session record (closing the GATE-001 no-op gap), AND `logEvidence` is invoked from the production `Handle` path (not only from tests). |

---

## AC-SEW-003 — test pass/fail derived by a pure classifier

| Field | Value |
|-------|-------|
| **Given** | M1 is complete; `classifyTestCommand(command, result) (isTest, isPass, isFail bool)` exists. |
| **When** | The classifier is given a recognized test command with a PASS result, a FAIL result, and an ambiguous/absent result. |
| **Then** | PASS→`(true,true,false)`; FAIL→`(true,false,true)`; ambiguous→`(true,false,false)` (graceful degradation — absence ≠ fail). |
| **Evidence** | `go test ./internal/hook/ -run 'TestClassifyTestCommand' -v -count=1` shows passing subtests `go_test_pass→(true,true,false)`, `go_test_fail→(true,false,true)`, `go_test_ambiguous→(true,false,false)`. |
| **Pass criterion** | The three subtests pass; the ambiguous case sets NEITHER binary flag (REQ-SEW-007 / REQ-SEW-013 graceful degradation). The function performs no I/O (verify: signature takes `command string, result []byte`, returns 3 bools — no `*HookInput`, no file/network access). |

---

## AC-SEW-004 — path classified code-change vs docs-only by a pure classifier

| Field | Value |
|-------|-------|
| **Given** | M1 is complete; `classifyPathKind(filePath string) string` exists. |
| **When** | The classifier is given a `.go` path, a `.md`/`README`/`.moai/specs/*.md` path, and an unknown-extension path. |
| **Then** | `.go`→`PathKindCodeChange`; docs path→`PathKindDocsOnly`; unknown→`PathKindUnknown`. |
| **Evidence** | `go test ./internal/hook/ -run 'TestClassifyPathKind' -v -count=1` shows passing subtests `go_ext→code-change`, `md_ext→docs-only`, `readme_base→docs-only`, `unknown_ext→unknown`. |
| **Pass criterion** | All path-class subtests pass; the returned values are the `telemetry.PathKind*` constants (verify the classifier returns `telemetry.PathKindCodeChange`/`PathKindDocsOnly`/`PathKindUnknown`, not bare strings). |

---

## AC-SEW-005 — persisted through RecordSkillUsage; no new store

| Field | Value |
|-------|-------|
| **Given** | M3 is complete; `logEvidence` writes via the telemetry store. |
| **When** | An auditor inspects the writer's persistence path and checks for any new on-disk store. |
| **Then** | `logEvidence` calls `telemetry.RecordSkillUsage` (the existing daily JSONL path) and opens no new store file directly. |
| **Evidence** | `grep -c 'telemetry.RecordSkillUsage' internal/hook/evidence_writer.go` returns ≥ 1. AND `grep -cE 'os\.Create|os\.OpenFile|os\.WriteFile|os\.MkdirAll' internal/hook/evidence_writer.go` returns `0` (writer opens no file directly — bare `-E` alternation `|`, dots escaped `os\.`). |
| **Pass criterion** | `RecordSkillUsage` referenced ≥ 1 in `evidence_writer.go` AND zero direct file-write primitives in `evidence_writer.go` (REQ-SEW-005 — reuse store, no new store). |

---

## AC-SEW-006 — canonical multi-language test commands recognized

| Field | Value |
|-------|-------|
| **Given** | M1 is complete; `classifyTestCommand` is implemented. |
| **When** | The recognizer is given `go test ./...`, `pytest`, `cargo test`, `npm test`, `npm run test`, `jest`, `vitest`, `pnpm test`, `yarn test`. |
| **Then** | Each returns `isTest=true`. |
| **Evidence** | `go test ./internal/hook/ -run 'TestClassifyTestCommand' -v -count=1` shows a passing subtest per canonical command (each asserting `isTest=true`). |
| **Pass criterion** | All canonical multi-language test-command subtests return `isTest=true` per REQ-SEW-006. |

---

## AC-SEW-007 — C-HRA-008 subagent boundary holds in hook code

| Field | Value |
|-------|-------|
| **Given** | M4 is complete; `evidence_writer.go` is the new hook-domain file. |
| **When** | An auditor greps the hook package for the user-interaction channel. |
| **Then** | No `AskUserQuestion` / `mcp__askuser` appears in non-test, non-comment hook code (including `evidence_writer.go`). |
| **Evidence** | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v "_test.go" \| grep -v "^[^:]*:[0-9]*:[ \t]*//"` returns no matches (exit 1). |
| **Pass criterion** | Zero matches (basic-grep `\|` alternation is intentional here, mirroring GATE-001 AC-SEG-007). The writer adds no user-interaction channel per REQ-SEW-015. |

---

## AC-SEW-008 — non-evidence events produce no false claim

| Field | Value |
|-------|-------|
| **Given** | M1+M2 complete; classifiers + assembler implemented. |
| **When** | The assembler is given a non-test Bash command (`go build`, `ls`, `git status`) and an Edit of an unknown-extension file (`config.xyz`). |
| **Then** | Non-test Bash → `buildEvidenceRecord` returns `ok=false` (no record) and `classifyTestCommand` returns `isTest=false` with both binary flags false; unknown-extension Edit → `classifyPathKind` returns `PathKindUnknown` and `buildEvidenceRecord` returns `ok=false`. |
| **Evidence** | `go test ./internal/hook/ -run 'TestClassifyTestCommand\|TestClassifyPathKind\|TestBuildEvidenceRecord' -v -count=1` shows passing subtests `go_build→not_test`, `ls→not_test`, `unknown_ext→unknown`, `non_test_bash→ok_false`, `unknown_ext_edit→ok_false`. |
| **Pass criterion** | Non-test commands set no binary flag (REQ-SEW-008) AND unknown paths classify as unknown with no record (REQ-SEW-011) — so no false code-change claim enters the ledger. |

---

## AC-SEW-009 — FALSIFIABLE GATE ACTIVATION (the value proof)

| Field | Value |
|-------|-------|
| **Given** | M1-M3 + M5 complete; the writer produces records and the GATE-001 read path (`telemetry.LoadBySession` → `buildSessionLedger` → `evaluateEvidence`) is unchanged (PRESERVE). |
| **When** | **(Positive)** A session is built where a code-change-bearing event ran (Edit `.go` → record {Outcome=success, PathKind=code-change}) AND a test event was observed as a FAIL (Bash `go test` FAIL → record {Outcome=error, IsTestFail=true}), with NO observed test-pass; then `evaluateEvidence(buildSessionLedger(LoadBySession(...)))` is evaluated. **(Negative)** The same session but the test event was observed as a PASS (record {Outcome=success, IsTestPass=true}). |
| **Then** | **(Positive)** the gate returns a **non-nil `Finding`** (the advisory fires — code-change session claimed success with no observed pass). **(Negative)** the gate returns **nil** (the success is backed by an observed pass — no advisory). The ONLY delta between positive and negative is `IsTestPass` (false→true). |
| **Evidence** | `go test ./internal/hook/ -run 'TestEvidenceGate_ActivatesInProduction' -v -count=1` shows TWO passing subtests: (1) `code_change_test_fail_no_pass → finding` returning a non-nil Finding with `Finding.HumanReadable()` containing `path-kind=code-change`; (2) `code_change_test_pass → nil` returning nil. The records in both subtests are produced by the SPEC's own `buildEvidenceRecord` + persisted via `telemetry.RecordSkillUsage` into a `t.TempDir()` store, then read back by the unmodified GATE-001 `LoadBySession`/`buildSessionLedger`/`evaluateEvidence` chain. |
| **Pass criterion** | Both subtests pass; flipping ONLY `IsTestPass` (false→true) flips the gate verdict (finding→nil) while the code-change success claim is held constant — proving (a) the writer's records reach the gate's ledger end-to-end, and (b) the gate, dormant under GATE-001's stream, now FIRES on genuine observed evidence produced by this SPEC's writer. This is the falsifiable activation: if the writer were a no-op (records never written, or evidence fields never set), subtest (1) would not produce a finding and this AC would FAIL. There is no separate opt-in flag — the production caller is `postToolHandler.Handle` (AC-SEW-002), exercised on every live PostToolUse event. |

---

## AC-SEW-010 — behavior preservation + fail-open + ≤5s + SessionID

| Field | Value |
|-------|-------|
| **Given** | M3+M4 complete; `logEvidence` wired additively into `Handle`. |
| **When** | (a) A Skill event is dispatched through `Handle`; (b) a Bash event with a read-only telemetry parent dir is dispatched; (c) the writer records an evidence event. |
| **Then** | (a) `logSkillUsage` still writes its Skill record and the `HookOutput` JSON is unchanged (no new systemMessage from the writer); (b) the write error is swallowed (`slog.Warn`) and `Handle` returns allow + nil error (fail-open); (c) the record's `SessionID == input.SessionID` and the writer performs no test re-execution / no network (structural — it reads the result from `input`, calls only `RecordSkillUsage`). |
| **Evidence** | `go test ./internal/hook/ -run 'TestHandle_PreservesExistingObservers\|TestLogEvidence_NeverBlocks\|TestLogEvidence_SessionID\|TestLogEvidence_FailOpen' -v -count=1` all pass. AND `grep -cE 'http\.|net\.Dial|exec\.Command|os/exec' internal/hook/evidence_writer.go` returns `0` (no network / no subprocess test re-execution — ≤5s budget structural proof; bare `-E` alternation `|`, dots escaped). |
| **Pass criterion** | Existing observers preserved (Skill record still written, HookOutput unchanged) per REQ-SEW-012; write errors swallowed and Handle returns allow+nil per REQ-SEW-013; `SessionID` propagated per REQ-SEW-016; zero network/subprocess primitives in the writer proving the ≤5s in-memory-classify + single-append budget per REQ-SEW-014. |

---

## E. Quality Gate Criteria (Definition of Done)

- All 10 AC PASS (binary, verified by the listed commands).
- AC-SEW-009 (falsifiable activation) PASS — the gate provably fires in production via the writer's records.
- `go test ./...` GREEN (zero regressions); `go vet ./...` clean; `golangci-lint run` no NEW issues.
- Cross-platform build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- New-code coverage ≥ 85% (`internal/hook` evidence_writer.go functions); no telemetry-package regression.
- C-HRA-008 grep = 0 matches.
- PRESERVE verified: `git diff --stat` shows NO change to `internal/hook/session_ledger.go`, `internal/hook/stop.go`, `internal/telemetry/types.go`, `internal/telemetry/recorder.go`.
- progress.md §E.3 run-phase evidence populated; §E.8 commit SHAs recorded.
