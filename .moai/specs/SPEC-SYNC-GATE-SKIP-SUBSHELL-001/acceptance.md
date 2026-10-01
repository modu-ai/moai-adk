# acceptance.md — SPEC-SYNC-GATE-SKIP-SUBSHELL-001

Canonical AC enumeration. Every AC is mechanically verifiable; commands are plain-command form; moving coordinates (line numbers) are anchored by quoted content patterns. Evidence home: `.moai/reports/t1395/` (gitignored — on-disk lane evidence; verdict quotes decide).

## D. AC Matrix

| AC | Statement (one line) | Severity | Verification |
|----|----------------------|----------|--------------|
| AC-001 | Two-cell observational ledger: RED-now confirmed the defect on the pre-implementation tree; GREEN re-run post-fix shows all three cells flipped | Blocker | RED: already executed (ledger §D.1); GREEN: re-run `red-now.sh` post-fix |
| AC-002 | Static sweep: zero pipeline-wrapping-`run_step` shapes in BOTH gate surfaces, positive control 4 on the pre-fix commit | Blocker | `grep -cE "run_step [a-zA-Z]+ c[12] [^']*\|"` |
| AC-003 | Behavioral sweep test: all four languages, tool absent → audit `skipped_tools=<tool>` non-empty + stdout carries the systemMessage notice + decision=allow | Blocker | `go test ./internal/template/ -run '^TestSyncGateSkipNotice_ToolAbsentNotifies_AllFourLanguages$'` |
| AC-004 | Dual-surface byte-identity + embedded regeneration | Blocker | `cmp` + `TestHookWrapperCopiesStayIdentical` + `make build` |
| AC-005 | t602 exit-status family unregressed: failing checker still blocks in all four repaired languages | Blocker | `go test ./internal/template/ -run '^TestSyncGateExitStatus'` |
| AC-006 | t1385 gate-cache contract unregressed: pass record written on allow-with-skip; failstate family green | Blocker | `go test ./internal/hook/ -run '^TestSyncGateFailState'` + AC-003's record assertion |
| AC-007 | t1392 behaviors unregressed: multi-language gate family green | High | `go test ./internal/template/ -run '^TestSyncGateMultiLanguage'` |
| AC-008 | cpp gate behavior family unregressed (shares the case table) | High | `go test ./internal/template/ -run '^TestSyncGateCpp'` |

## D.1 AC-001 — Two-cell observational ledger (RED-now / GREEN-later)

- **RED-now (executed at plan phase — 2026-10-02; harness committed as a plan-phase artifact at `red-now-t1395.sh` in this SPEC directory, re-executed FROM the committed path):**
  - Given the pre-implementation tree — `git rev-parse --short HEAD` = `89e3164af` observed immediately before AND after the run (full SHA `89e3164aff20345618c5db053107b066adc98316`; gate script sha256 prefix `19180598f67db114`, recomputed post-run and unchanged),
  - When `bash .moai/specs/SPEC-SYNC-GATE-SKIP-SUBSHELL-001/red-now-t1395.sh "$PWD/.claude/hooks/moai/sync-phase-quality-gate.sh" "$PWD/.moai/reports/t1395/reexec"` runs (C# fixture: `f.csproj` + `a.cs`, single git commit; scrubbed PATH `/usr/bin:/bin`; `MOAI_SYNC_GATE_BLOCKING=1`; stdin closed),
  - Then (observed, verbatim, harness exit 0):
    ```
    pc1 PASS: dotnet unresolvable in scrubbed env
    === gate exit code: 0 ===
    CELL1 CONFIRMED: stdout EMPTY
    CELL2 CONFIRMED: audit skipped_tools empty:
    skipped_tools=
    CELL3 CONFIRMED: pass record written: d5dac44b6ecdc01e37c41ad2c06faa6740e99b01 pass cf131089db9b7300b2acb5cfc18618bc5e08e1204abc26be7da7e38878afc8ff
    PC2 PASS: gate reached the csharp case
    PC3 PASS: audit line carries the skipped_tools= token
    harness-exit=0
    ```
    Positive controls: `pc1 PASS` (dotnet unresolvable), `PC2 PASS` (`language=csharp` in audit), `PC3 PASS` (`skipped_tools=` token present). The temp fixture's git commit SHA (`d5dac44b…`) differs per run by design; the worktree content id (`cf131089…`) is identical across the first run and the re-execution.
  - Zero-hit grep positive control: the audit's `skipped_tools=` token IS present (PC3) while its value is empty — the empty value is a real observation, not a missing line.
  - Citation note: the harness lives in this committed SPEC directory so the RED cell is re-executable from a fresh checkout post-merge (t1378 precedent — a RED citation to a gitignored `.moai/reports/` artifact is non-re-executable). The first-run copies under `.moai/reports/t1395/` are historical lane evidence.
- **GREEN-later (run phase, same harness, same recipe):**
  - Given the repaired tree (both surfaces byte-identical, fix applied),
  - When the same harness runs,
  - Then CELL1 reads stdout carrying `systemMessage` (the skip notice), CELL2 reads `skipped_tools=` followed by the absent tool name (non-empty), CELL3 STILL reads a `pass` record (t1385 contract preserved — flipping CELL3 to "no record" is a FAILURE of AC-006, not a success).
- Both cells live in one fenced evidence ledger in the run-phase verdict inputs; the RED cell quotes this plan-phase execution verbatim.

## D.2 AC-002 — Static four-check sweep with positive control

- Given the repaired gate script,
  - When `grep -nE "run_step [a-zA-Z]+ c[12] [^']*\|" internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` runs,
  - Then it exits 1 with zero output (no pipeline wraps `run_step` in any of the four checks), and the identical pattern against `git show 89e3164af:internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` returns exactly 4 hits (positive control — the pattern detects the pre-fix shapes; proven at plan phase: count 4 on both working copy and pre-fix commit).
  - And the same grep on the local deployed copy returns zero.
- The discriminator's precision is proven: it matches ONLY the four defective lines (807/823/882/886 at `89e3164af`) and none of the seven internally-piped `sh -c` checkers.

## D.3 AC-003 — Behavioral four-language sweep test

- Given a temp project carrying each language's markers in turn (csharp: `f.csproj`+`a.cs`; elixir: `mix.exs`+`a.ex`; flutter: `pubspec.yaml`+`a.dart`; swift: `Package.swift`+`a.swift` — mirroring `gateExitCases()` line 53-56),
  - When the gate runs with the language's tool genuinely ABSENT from a constructed PATH (the test asserts `exec.LookPath`-equivalent unresolvability inside its env before invoking the gate — the in-test analog of positive control PC1) and `MOAI_SYNC_GATE_BLOCKING=1`,
  - Then for EACH of the four subtests (`csharp`, `elixir`, `flutter`, `swift`): the audit log's final line carries `skipped_tools=` followed by the tool name (`dotnet`/`mix`/`dart`/`swift`), stdout is non-empty and carries `"systemMessage"` with the checker-absent notice, `decision=allow` appears in the audit line, and the `steps.journal`-derived audit trail shows `tool=<tool> skipped`.
- Exact test name: `TestSyncGateSkipNotice_ToolAbsentNotifies_AllFourLanguages` (subtests named per language). Test env discipline: explicit PATH construction, `runtime.GOOS == "windows"` skip, `t.TempDir()` isolation, no ambient lane env leakage.
- Negative control inside the suite: the existing `TestSyncGateExitStatus_PassingCheckerAllows` continues to pass (a PRESENT, passing tool must not block) — guards the repair from flipping into over-blocking.

## D.4 AC-004 — Dual-surface byte-identity + embedded regeneration

- Given the repair landed,
  - When `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` runs, it exits 0 (byte-identical); `go test ./internal/hook/ -run '^TestHookWrapperCopiesStayIdentical$'` passes (the existing copy contract test at `wrapper_copies_contract_test.go:73`); and `make build` exits 0 (embedded `//go:embed all:templates` copy regenerated — Template-First closure).

## D.5 AC-005 — t602 exit-status family unregressed

- Given the four repaired lines no longer carry the external pipe,
  - When `go test ./internal/template/ -run '^TestSyncGateExitStatus'` runs,
  - Then all subtests pass, including `TestSyncGateExitStatus_FailingCheckerBlocks` for csharp/elixir/flutter/swift (a failing checker — stub present, exiting 7 — still blocks) and `TestSyncGateExitStatus_PassingCheckerAllows` (passing checker allows).

## D.6 AC-006 — t1385 gate-cache contract unregressed

- Given the repair touched only the four check lines,
  - When `go test ./internal/hook/ -run '^TestSyncGateFailState'` runs, all failstate subtests pass (`AC003_PassThenSameHeadStaysSilent`, `AC005_TornWriteNeverSilentPass`, `AC007_NoPathLooserThanToday`, the full family),
  - And AC-003's GREEN run demonstrates the `pass` record is still written on allow-with-skip (CELL3 of AC-001 GREEN).
  - Checker-tool-presence-in-key (gate lines "The checker-tool presence rides the key…" block) is not modified — verified by `git diff` showing no hunks outside the four check lines (and any comment lines directly attached to them).

## D.7 AC-007 / AC-008 — t1392 and cpp-family regressions

- `go test ./internal/template/ -run '^TestSyncGateMultiLanguage'` green (t1392's multi-language routing and later-language aggregation surface) — AC-007, High.
- `go test ./internal/template/ -run '^TestSyncGateCpp'` green (the case table's most heavily-annotated neighbor; guards against incidental table damage) — AC-008, High.

## Edge cases

- Tool present but the repair's env lacks it transiently — out of scope (no behavior change for the present-tool path).
- Multiple tools absent in one run (e.g. csharp AND swift markers changed): `SKIPPED_TOOLS` accumulates both names — the audit line carries both; AC-003's subtests each exercise one; the accumulation is exercised implicitly by the multi-language family.
- macOS `stat -f` vs GNU `stat -c` in `record_age_seconds` — untouched by this SPEC (failstate family owns it).

## Quality gates

- Scoped-package tests only (lane rule): `./internal/template/` and `./internal/hook/` gate families; no local full-suite run; CI owns the full verdict.
- `make build` MUST succeed (embed regeneration + agents-emit-check prerequisites).

## Definition of Done

- All eight ACs PASS with commands + verbatim outputs recorded in the run-phase evidence record; RED/GREEN two-cell ledger complete; both surfaces byte-identical; progress.md §E.2/§E.3 populated; no hunks outside the four check lines (+ their directly-attached comments) in the gate script.
