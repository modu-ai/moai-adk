# Design — SPEC-STOP-EVIDENCE-WRITER-001

> Resolves the **record-time evidence-sourcing architecture**: exactly which PostToolUse events get recorded, how test-pass/fail is detected from a Bash command, how `PathKind` is classified from a path, and how it threads into the existing session-ledger aggregation + the Stop gate.

## §0 The central design tension (and its resolution)

**Tension.** The gate's evidence (`IsTestPass`/`IsTestFail`/`PathKind`) cannot come from the Skill-invocation record, because `logSkillUsage` fires at Skill *invocation* time — before any test runs (research.md §1.1, §1.3). So **evidence must be sourced from a different event** than the one that carries the success claim.

**Resolution — record evidence from the tool events that produce it, on the same session key.** The session ledger (`buildSessionLedger`) aggregates ALL records for a session, so the success claim and the binary evidence do NOT need to be on the same record — they only need to be in the same session. Therefore:

- The writer adds NEW `UsageRecord`s on the **Bash** (test result) and **Edit/Write** (path-kind) PostToolUse branches.
- Those records share `SessionID` with the rest of the session.
- `buildSessionLedger` OR-reduces `BinaryPass`/`BinaryFail`/`SuccessClaims` and explicit-`PathKind`-wins across all of them.

This is why no gate-side change is needed: the gate already reads the union ledger.

## §1 Where the writer hooks in (the production caller)

`internal/hook/post_tool.go` — `postToolHandler.Handle`. A single additive call is inserted alongside the existing observers (after `runMemoryAudit`, before `writeHookMetric`), guarded by tool name:

```go
// Record evidence-bearing tool events for the Stop evidence gate
// (SPEC-STOP-EVIDENCE-WRITER-001). Best-effort, additive — never blocks.
if input.ToolName == "Bash" || input.ToolName == "Edit" || input.ToolName == "Write" {
    logEvidence(input)
}
```

`logEvidence` lives in the NEW file `internal/hook/evidence_writer.go`. This is the only edit to `post_tool.go`. All existing branches (Skill→logSkillUsage, Agent/Task→logTaskMetrics, LSP/AST/MX/memory-audit, writeHookMetric) are untouched (REQ-SEW-012). The production caller is the live PostToolUse path — exercised on every tool call — so the writer is NOT behind an opt-in flag (this is what makes activation falsifiable, §A.4 / AC-SEW-009).

## §2 The two classifiers (pure, side-effect-free)

### §2.1 `classifyTestCommand(command string, result []byte) (isTest, isPass, isFail bool)`

Pure function in `evidence_writer.go`. Two stages:

1. **Recognize** (REQ-SEW-006): is this a test command? Match the command text against a fixed prefix/substring taxonomy:

   ```go
   var testCommandSignatures = []string{
       "go test", "pytest", "cargo test",
       "npm test", "npm run test", "pnpm test", "yarn test",
       "jest", "vitest",
   }
   ```

   `isTest = true` iff the trimmed command contains one of these signatures as a command token. Non-test commands (`go build`, `ls`, `git status`) → `isTest=false`, both flags false (REQ-SEW-008).

2. **Derive pass/fail** (REQ-SEW-007) from the observed result, only when `isTest`:
   - Parse the result (`input.ToolResponse` preferred, `input.ToolOutput` fallback). Look for an exit-code signal first: a structured `{"exit_code": 0}` / `{"interrupted": false}` → pass; non-zero exit → fail.
   - Fallback to output-text heuristics (only when no exit code): presence of go's `--- FAIL` / `FAIL\t` / `ok \t` / pytest `passed` / `failed` / cargo `test result: ok` / `test result: FAILED`. Pass markers → `isPass`; fail markers → `isFail`.
   - **Graceful degradation (REQ-SEW-013, R1):** when the result is absent, unparseable, or carries no recognizable pass/fail signal → set NEITHER flag. Absence ≠ fail. This keeps the gate conservative (no false finding).

Returns `(isTest, isPass, isFail)`. The function does no I/O — caller passes the already-available bytes.

### §2.2 `classifyPathKind(filePath string) string`

Pure function returning a `telemetry.PathKind*` constant.

```go
var codeExtensions = []string{
    ".go", ".py", ".ts", ".js", ".rs", ".java", ".kt", ".cs",
    ".rb", ".php", ".ex", ".cpp", ".scala", ".r", ".dart", ".swift",
}
var docsExtensions = []string{".md", ".mdx", ".txt", ".rst"}
```

Logic (first match wins):
1. Extension ∈ `codeExtensions` → `PathKindCodeChange` (REQ-SEW-009).
2. Extension ∈ `docsExtensions`, OR base name is `CHANGELOG*`/`README*`, OR path is under `.moai/specs/` or `docs/` with a prose extension → `PathKindDocsOnly` (REQ-SEW-010).
3. Otherwise → `PathKindUnknown` (REQ-SEW-011) — the gate treats unknown conservatively (no code-change claim, no finding).

### §2.3 Non-evidence events produce no record (write-volume discipline, R3)

`logEvidence` records a `UsageRecord` ONLY when the event carries genuine evidence:
- Bash: only when `isTest == true` (a recognized test command). A non-test Bash → early return, no record.
- Edit/Write: only when `classifyPathKind != PathKindUnknown`. An unknown-extension file → early return, no record.

This keeps the JSONL proportional to real evidence and avoids polluting the report aggregation.

## §3 The Outcome assignment (the activation hinge) — RESOLVED

`evaluateEvidence` fires only when `SuccessClaims > 0` (Outcome ∈ {success, partial}) AND a binary signal observed AND `BinaryPass == false` (research.md §2.2). So the writer's `Outcome` assignment per record is the hinge. Resolution:

| Event | `Outcome` set | `IsTestPass` | `IsTestFail` | `PathKind` | Rationale |
|-------|---------------|--------------|--------------|------------|-----------|
| Bash test, observed PASS | `OutcomeSuccess` | `true` | `false` | — | Backed success — `BinaryPass=true` → gate returns nil (correct: a real pass observed). |
| Bash test, observed FAIL | `OutcomeError` | `false` | `true` | — | NOT a success claim itself; contributes `BinaryFail` so a binary signal IS observed. |
| Bash test, ambiguous | `OutcomeUnknown` | `false` | `false` | — | No signal; graceful degradation. |
| Edit/Write code-change | `OutcomeSuccess` | `false` | `false` | `code-change` | **This is the success claim on a code-change session.** It carries no binary evidence itself. |
| Edit/Write docs-only | `OutcomeUnknown` | `false` | `false` | `docs-only` | docs-only is exempt; no success claim needed. |

**Why a code-change Edit/Write carries `Outcome=success`:** the targeted defect shape is "a code-change session that *claims success* but has no observed test-pass". A successful Edit/Write of a `.go` file IS the implicit success claim ("I changed the code and consider it done"). Pairing that `Outcome=success` + `PathKind=code-change` record with a session that ALSO recorded a test-fail (`BinaryFail=true`, `BinaryPass=false`) — or recorded NO test at all — produces the exact `evaluateEvidence` finding condition.

Trace of the **finding** case (AC-SEW-009 positive):
1. Edit `.go` → record {Outcome=success, PathKind=code-change}.
2. Bash `go test` FAIL → record {Outcome=error, IsTestFail=true}.
3. At Stop: `buildSessionLedger` → `PathKind=code-change` (explicit wins), `SuccessClaims=1` (from #1), `BinaryFail=true`, `BinaryPass=false`.
4. `evaluateEvidence`: not docs/unknown ✓, SuccessClaims>0 ✓, binaryObservable (BinaryFail) ✓, BinaryPass=false ✓ → **non-nil Finding** → advisory emitted. ✓

Trace of the **no-finding** case (AC-SEW-009 negative):
1. Edit `.go` → record {Outcome=success, PathKind=code-change}.
2. Bash `go test` PASS → record {Outcome=success, IsTestPass=true}.
3. At Stop: `SuccessClaims=2`, `BinaryPass=true`.
4. `evaluateEvidence`: `BinaryPass=true` → returns nil → NO advisory. ✓ (correct: the success IS backed by an observed pass.)

The behavioral delta between the two cases is exactly `IsTestPass` (false→true) — same as GATE-001's AC-SEG-002 binary-flip discipline. This is the falsifiable activation.

## §4 The writer function shape (mirrors logSkillUsage)

```go
// logEvidence records evidence-bearing PostToolUse events (Bash test results,
// Edit/Write path-kind) to the session telemetry store so the Stop evidence
// gate (SPEC-STOP-EVIDENCE-GATE-001) can fire. Best-effort: errors logged with
// slog.Warn, never returned. (REQ-SEW-005, REQ-SEW-012, REQ-SEW-013)
func logEvidence(input *HookInput) {
    projectRoot := resolveProjectRoot(input)   // REUSE — no-project-root skip (REQ-SEW-013)
    if projectRoot == "" {
        return
    }
    rec, ok := buildEvidenceRecord(input)       // pure: classify + assemble (§2, §3)
    if !ok {
        return                                   // non-evidence event → no record (§2.3)
    }
    if err := telemetry.RecordSkillUsage(projectRoot, rec); err != nil {  // REUSE store (REQ-SEW-005)
        slog.Warn("evidence writer: failed to record", "tool", input.ToolName,
            "session_id", input.SessionID, "error", err)
    }
}
```

`buildEvidenceRecord(input) (telemetry.UsageRecord, bool)` is the pure assembler: it parses `ToolInput` for `command` (Bash) or `file_path` (Edit/Write), calls the §2 classifiers, applies the §3 Outcome table, sets `SessionID`/`Timestamp`, and returns `ok=false` for non-evidence events. Returning a value + bool keeps it unit-testable without I/O.

## §5 Behavior-preservation & boundary guards

- **Additive only (REQ-SEW-012):** one new `if` block in `Handle` + one new file. No existing observer touched. The `HookOutput` returned by `Handle` is byte-identical (the writer mutates no metrics map, sets no systemMessage).
- **Fail-open (REQ-SEW-013):** every path in `logEvidence` either returns early or swallows the error via `slog.Warn`. `Handle` never sees an error from it.
- **≤5s budget (REQ-SEW-014):** classify is in-memory regex/substring; one append-write per evidence event (same as `logSkillUsage`). No test re-execution (the result is read from `input`, not re-run), no network, no repo scan.
- **C-HRA-008 (REQ-SEW-015):** `evidence_writer.go` contains no `AskUserQuestion` / `mcp__askuser` reference. CI guard: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v comment` = 0.
- **Session correlation (REQ-SEW-016):** `rec.SessionID = input.SessionID` so `LoadBySession` returns it.

## §6 @MX tags planned

- `@MX:ANCHOR` on `logEvidence` (fan_in via Handle + tests; the production write path) with `@MX:REASON` (gate-activation hinge — changes affect whether the gate fires in production).
- `@MX:NOTE` on the §3 Outcome table rationale (why a code-change Edit carries Outcome=success — non-obvious business rule).

## §7 Cross-references

- PRESERVE (read-side, unchanged): `internal/hook/session_ledger.go` (`runEvidenceGate`/`buildSessionLedger`/`inferPathKind`/`evaluateEvidence`), `internal/hook/stop.go`, `internal/telemetry/types.go` (omitempty fields already present).
- REUSE: `telemetry.RecordSkillUsage` / `LoadBySession` (`internal/telemetry/recorder.go`), `resolveProjectRoot` (`internal/hook/post_tool_metrics.go:98`).
- Doctrine: `.claude/rules/moai/core/verification-claim-integrity.md` (the invariant this line serves; codification owned by IMP-06, NOT this SPEC).
