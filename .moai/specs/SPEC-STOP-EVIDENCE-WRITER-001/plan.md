# Implementation Plan — SPEC-STOP-EVIDENCE-WRITER-001

> cycle_type=tdd (RED→GREEN→REFACTOR per milestone). Tier M. Priority-ordered milestones M1-M6 (no time estimates).

## §A Context

- **Location**: project root `/Users/goos/MoAI/moai-adk-go`.
- **Branch**: `feat/SPEC-STOP-EVIDENCE-WRITER-001` (run-phase; main checkout, Hybrid Trunk 1-person OSS per git-workflow-doctrine — Tier M main 직진 allowed).
- **SPEC artifacts**: `.moai/specs/SPEC-STOP-EVIDENCE-WRITER-001/{spec,plan,acceptance,design,research}.md` + `progress.md`.
- **Predecessor**: SPEC-STOP-EVIDENCE-GATE-001 (completed, origin/main `59561c92d`) — built the dormant gate this SPEC activates.
- **PRESERVE**: `internal/hook/session_ledger.go`, `internal/hook/stop.go`, `internal/telemetry/types.go` (omitempty fields already present), `internal/telemetry/recorder.go`, all existing `post_tool.go` observers (Skill/Agent/LSP/AST/MX/memory-audit/writeHookMetric).
- **EXTEND**: `internal/hook/post_tool.go` (one additive `if` block in `Handle`).
- **NEW**: `internal/hook/evidence_writer.go` + `internal/hook/evidence_writer_test.go`.

### §A.1 Tier M justification

2-4 files (post_tool.go EXTEND + evidence_writer.go NEW + evidence_writer_test.go NEW + optional post_tool_test.go EXTEND). Estimated 300-600 LOC (2 pure classifiers + 1 assembler + 1 writer + behavior-preservation wiring + comprehensive table-driven tests + the end-to-end gate-activation test). Single coding domain (`internal/hook`). Above Tier S (<5 files but >300 LOC and a falsifiable activation requirement warrants design.md + research.md). Below Tier L (no constitutional change, <15 files). → **Tier M**.

## §B Known Issues (Section B auto-injection, filtered to relevant categories)

- **B3 (C-HRA-008)**: `evidence_writer.go` is hook-domain code — MUST NOT call `AskUserQuestion`/`mcp__askuser`. CI guard grep = 0 (AC-SEW-007).
- **B1 (Cross-platform)**: the writer uses no `syscall` — pure Go string/json/regexp + `telemetry.RecordSkillUsage`. `GOOS=windows GOARCH=amd64 go build ./...` MUST pass (no build tag split needed).
- **B7 (project-root resolution)**: REUSE `resolveProjectRoot(input)` (handles `$CLAUDE_PROJECT_DIR` + `.moai/` guard) — do NOT inline `os.Getenv`.
- **B8 (working-tree hygiene)**: do NOT touch runtime-managed `.moai/evolution/telemetry/*.jsonl` or `.moai/state/`. Tests use `t.TempDir()`.
- **B4 (frontmatter)**: spec.md uses canonical 12 fields + `era: V3R6` + `tier: M` (no snake_case aliases).
- **B9 (commit + push)**: manager-develop self-commits per milestone (`feat(SPEC-STOP-EVIDENCE-WRITER-001): M{N} <subject>`), pushes at end. No `--no-verify`.
- **B10 (PRESERVE scope)**: do NOT modify the GATE-001 read-side files (C.1) — they are correct and complete.

## §C Pre-flight (run-phase entry, before code)

```bash
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5          # baseline (NEW vs pre-existing)
go test ./internal/hook/... ./internal/telemetry/... 2>&1 | tail -5   # green baseline
grep -n 'input.ToolName == "Skill"' internal/hook/post_tool.go        # confirm insertion locus
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v '//'  # 0 baseline
```

## §D Constraints (DO NOT VIOLATE)

- PRESERVE the GATE-001 read-side (C.1): no edit to `session_ledger.go` / `stop.go` / `types.go`.
- No new on-disk store (REQ-SEW-005): reuse `telemetry.RecordSkillUsage`.
- No `logSkillUsage` retrofit (C.6): add NEW records on Bash/Edit/Write, do not change the Skill record.
- No advisory→blocking promotion (C.3): the gate stays advisory.
- No config file / regex registry (C.4): fixed constant taxonomies only.
- Forbidden: `--no-verify`, `--amend`, force-push to main.
- Required: Conventional Commits + `🗿 MoAI` trailer.
- C-HRA-008 binary: grep 0 matches in non-test, non-comment hook code.

## §E Self-Verification (manager-develop completion report)

E1 AC binary PASS/FAIL matrix (AC-SEW-001..010) · E2 cross-platform build (linux + windows exit 0) · E3 coverage (`internal/hook` ≥85% on new code; no telemetry regression) · E4 C-HRA-008 grep 0 · E5 lint NEW vs baseline · E6 branch HEAD + push state · E7 blocker report if any.

## §F Milestones (RED→GREEN→REFACTOR)

### M1 — Test-command + path-kind classifiers (pure functions)

- **RED**: `evidence_writer_test.go` — `TestClassifyTestCommand` table (go test PASS/FAIL/ambiguous, pytest, cargo test, npm test, jest, non-test `go build`/`ls` → not-test) + `TestClassifyPathKind` table (`.go`→code-change, `.md`/`README`/`.moai/specs/*.md`→docs-only, `.xyz`→unknown). All RED.
- **GREEN**: implement `classifyTestCommand` (§2.1) + `classifyPathKind` (§2.2) in `evidence_writer.go`.
- **REFACTOR**: extract `testCommandSignatures` / `codeExtensions` / `docsExtensions` constants; godoc.
- **AC coverage**: AC-SEW-003, AC-SEW-004, AC-SEW-006, AC-SEW-008, AC-SEW-009(classifiers), AC-SEW-010, AC-SEW-011.

### M2 — Evidence record assembler (`buildEvidenceRecord`)

- **RED**: `TestBuildEvidenceRecord` table — Bash test PASS→{Outcome=success, IsTestPass=true}; Bash test FAIL→{Outcome=error, IsTestFail=true}; Bash ambiguous→{Outcome=unknown, neither flag}; Edit `.go`→{Outcome=success, PathKind=code-change}; Write `.md`→{Outcome=unknown, PathKind=docs-only}; non-test Bash→ok=false; unknown-ext Edit→ok=false; SessionID propagated. All RED.
- **GREEN**: implement `buildEvidenceRecord(input) (UsageRecord, bool)` per §3 Outcome table + §2.3 non-evidence skip.
- **REFACTOR**: tidy the Outcome-table switch; `@MX:NOTE` on the code-change=success rationale.
- **AC coverage**: AC-SEW-001 (Outcome table), AC-SEW-005 (record shape), AC-SEW-016 (SessionID).

### M3 — Writer + Handle wiring (production caller)

- **RED**: `TestLogEvidence_*` — no-project-root skip; non-evidence event no-write; evidence event writes via temp store (assert JSONL line via `LoadBySession`); error path swallowed (read-only logPath parent → slog.Warn, no panic). `TestHandle_RecordsEvidenceOnBashEditWrite` — Handle with a Bash test event produces a loadable record. All RED.
- **GREEN**: implement `logEvidence` (§4); add the `if input.ToolName == "Bash"||"Edit"||"Write" { logEvidence(input) }` block in `post_tool.go Handle` (additive, after `runMemoryAudit`, before `writeHookMetric`).
- **REFACTOR**: `@MX:ANCHOR` on `logEvidence`; ensure no metrics-map mutation.
- **AC coverage**: AC-SEW-002, AC-SEW-005, AC-SEW-013(skip/fail-open), AC-SEW-016.

### M4 — Behavior preservation + boundary guards

- **RED**: `TestHandle_PreservesExistingObservers` — Skill→logSkillUsage still writes its record; HookOutput JSON unchanged for a Bash event (no systemMessage added); `TestLogEvidence_NeverBlocks` (returns nothing; Handle still returns allow + nil err). C-HRA-008 guard test (`subagent_boundary` grep assertion) if not already covered.
- **GREEN**: confirm additive insertion preserves all branches (no code change expected if M3 done right; fix if regression found).
- **REFACTOR**: none beyond comments.
- **AC coverage**: AC-SEW-007 (C-HRA-008), AC-SEW-012 (behavior preservation), AC-SEW-013, AC-SEW-014 (no test re-exec / no network — structural assertion).

### M5 — Falsifiable gate-activation end-to-end test (the value proof)

- **RED**: `TestEvidenceGate_ActivatesInProduction` — build a session: write evidence records via `buildEvidenceRecord`+`RecordSkillUsage` (Edit `.go`→code-change/success + Bash `go test` FAIL→IsTestFail) into a temp store, then call the GATE-001 read path (`telemetry.LoadBySession` → `buildSessionLedger` → `evaluateEvidence`) and assert a **non-nil Finding**. Second case: same code-change record + Bash `go test` PASS→IsTestPass → assert **nil Finding**. The IsTestPass false→true flip is the only delta. RED until M1-M3 land.
- **GREEN**: passes once the writer produces the records the gate consumes (no gate change — PRESERVE).
- **REFACTOR**: assert the advisory `Finding.HumanReadable()` names `path-kind=code-change`.
- **AC coverage**: AC-SEW-009 (THE falsifiable activation proof).

### M6 — Integration verification + progress.md §E run-phase evidence

- **RED**: n/a (verification milestone).
- **GREEN**: full `go test ./...` GREEN; `go vet`; `golangci-lint run`; cross-platform build; C-HRA-008 grep 0; coverage report; populate `progress.md §E.3` run-phase evidence + §E.8 commit SHAs.
- **AC coverage**: all AC re-verified; E1-E7 completion report.

## §G Anti-Patterns to avoid

- Retrofitting `logSkillUsage` to guess test results (impossible — fires before tests; C.6).
- Recording a `UsageRecord` for every Bash/Edit/Write (write bloat; record only evidence-bearing events per §2.3).
- Editing the GATE-001 gate to "make it fire" (the gate is correct; only the writer was missing — C.1).
- Adding a config flag for the test-command list (over-engineering; C.4 / §A.5).
- Claiming activation without the M5 end-to-end test (would re-create the dormant-scaffold unfalsifiability — `L_plan_auditor_value_realization_unfalsifiable`).

## §H Cross-References

- spec.md §A.4 (falsifiable activation) ↔ acceptance.md AC-SEW-009 ↔ M5.
- design.md §3 (Outcome table) ↔ M2.
- design.md §1 (Handle insertion) ↔ M3.
- research.md §2.2 (gate finding condition) — the activation hinge.
