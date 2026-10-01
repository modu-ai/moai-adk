# plan.md — SPEC-SYNC-GATE-SKIP-SUBSHELL-001

Tier: M (artifact set: spec.md + plan.md + acceptance.md; + progress.md; decision-index.md per `interview.decision_gate: on`). Rationale: the mechanical diff is 4 lines, but the surface is multi-layer (template original + deployed local copy + embedded binary + Go test file) and the file carries two neighboring regression families (t1392, t1385/t602) that the ACs must pin. LOC well under the M band; file count 3 (template gate script, local gate copy, one test file) + evidence artifacts.

## A. Context

- Defect: four language checks (csharp `dotnet` 807, elixir `mix` 823, flutter `dart` 882, swift `swift` 886 — line numbers at `89e3164af`) wrap `run_step` in a pipeline; the absent-tool branch's `SKIPPED_TOOLS` update (line 693) lands in a subshell and is lost.
- Consequence chain: notice (999-1001) never emitted → stdout empty → audit `skipped_tools=` empty → `pass` record written (1011) → next Stop turn silently short-circuits (545-547).
- Observational RED already executed at plan phase, from the committed harness `red-now-t1395.sh` (this SPEC directory): all three defect cells confirmed, all three positive controls PASS, twice — first run from `.moai/reports/t1395/` (historical lane evidence) and re-execution from the committed path with identical observations (spec.md §B P4, acceptance.md §D.1).

## B. Known Issues

- The existing C# fixture (`hook_gate_exit_status_test.go`) stubs tools as PRESENT; it exercises only the t602 exit-status shapes and cannot see the absent-tool branch. The new test is a sibling harness with the tool genuinely absent from PATH.
- `.moai/reports/*` is gitignored — the RED harness therefore lives as a committed plan-phase artifact in this SPEC directory (`red-now-t1395.sh`, byte-identical to the first-run copy kept as historical lane evidence under `.moai/reports/t1395/`), and the RED/GREEN ledger contents are quoted verbatim into `acceptance.md` §D.1 so the citation survives post-merge (t1378 precedent: a RED citation to a gitignored artifact is non-re-executable).

## C. Pre-flight (run-phase entry checks)

1. `git rev-parse --short HEAD` → confirm tree; re-run the static discriminator on the template copy: `grep -cE "run_step [a-zA-Z]+ c[12] [^']*\|" internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` → expect 4 (pre-fix). If not 4, stop and re-anchor (t1392-style merge may have moved shapes).
2. `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` → identical before editing (if not, a foreign surface drifted — resolve before proceeding).
3. Re-run the RED harness once to confirm the defect is still present on this tree: `bash .moai/specs/SPEC-SYNC-GATE-SKIP-SUBSHELL-001/red-now-t1395.sh "$PWD/.claude/hooks/moai/sync-phase-quality-gate.sh" <writable-evidence-dir>` → three CONFIRMED cells (evidence dir may be any writable scratch path; the harness is committed in this SPEC directory so the pre-flight is executable from a fresh checkout).

## D. Constraints

- Template-First (HARD): edit `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` FIRST; mirror to `.claude/hooks/moai/sync-phase-quality-gate.sh`; `make build` regenerates the embedded copy (`internal/template/embed.go` `//go:embed all:templates`).
- Shell conventions per `.moai/docs/hook-development.md`: this gate is self-contained bash (no jq, no moai-binary dependency); keep the `run_step` function's contract untouched (signature, journal lines, slot semantics).
- Tests: extend `internal/template/hook_gate_exit_status_test.go` (or a sibling file in the same package) with the absent-tool harness; scrubbed env with explicit PATH construction; `runtime.GOOS == "windows"` skip like `gateExitRequire`; no reliance on installed toolchains.
- Lane discipline: no git operations; no agent spawning; blocker reports instead of questions.

## E. Self-Verification (run-phase exit checks)

- E1: AC matrix of `acceptance.md` all PASS with commands + verbatim outputs.
- E2: `go test ./internal/template/ -run '^TestSyncGate'` and `go test ./internal/hook/ -run '^(TestSyncGateFailState|TestHookWrapperCopies)'` green (scoped packages; no local full-suite run).
- E3: static discriminator = 0 hits on BOTH surfaces; `cmp` byte-identity holds; `make build` exit 0.
- E4: RED harness re-run post-fix: CELL1 flips to "stdout carries systemMessage notice", CELL2 flips to `skipped_tools= dotnet` (non-empty), CELL3 remains "pass record written" (t1385 contract preserved). Record as the GREEN half of the two-cell ledger.

## F. Milestones (priority-ordered, no time estimates)

- M1 (High) — Template repair: remove the external pipeline (`| head -N`) from the four language checks, keeping each check's `|| true` suffix decision minimal (see Implementation Note IN-2). Anchor by quoted content, not line numbers. Mirror to the local copy (byte-identical).
- M2 (High) — GREEN + sweep evidence: re-run RED harness (cells flip); static discriminator 0 on both surfaces; positive control re-run against `git show 89e3164af:<template>` stays 4.
- M3 (High) — Behavioral test: add `TestSyncGateSkipNotice_ToolAbsentNotifies_AllFourLanguages` (subtests csharp/elixir/flutter/swift) to the template test package: scrubbed PATH without the target tool, C# fixture markers per `gateExitCases`, assert audit `skipped_tools=<tool>` non-empty AND stdout carries `systemMessage` AND decision=allow. Same file: extend with the cache-preservation assertion (`pass` record still written on allow-with-skip) or pin it via the existing failstate family if coverage already exists.
- M4 (Medium) — Regression sweep: full gate test families green (`TestSyncGateExitStatus_*`, `TestSyncGateMultiLanguage_*`, `TestSyncGateCpp_*`, `TestHookWrapperCopiesStayIdentical`, `TestSyncGateFailState_*`); `make build` exit 0.
- M5 (Medium) — Evidence ledger: RED/GREEN two-cell ledger + all command outputs quoted into `.moai/reports/t1395/verdict.md` inputs; progress.md §E.2/§E.3 populated.

## G. Anti-Patterns (forbidden in run phase)

- Do NOT "fix" the subshell by keeping the pipe and adding `run_step ... | head -30 || SKIPPED_TOOLS=...` recover hacks — the repair is the removal of the external pipeline.
- Do NOT touch the seven internally-piped `sh -c` checker scripts (javac/kotlinc/scalac/ruby/php/g++/R).
- Do NOT change RECORD_FILE / PAYLOAD_FILE / retry semantics, `resolve_gate_mode`, or the exit-0-with-stdout-JSON blocking contract (t1396 territory).
- Do NOT run the local full Go test suite (lane rule); scoped packages only.
- Do NOT let the new test inherit ambient PATH — construct it explicitly; assert the tool is unresolvable inside the test env before running the gate (positive control inside the test, mirroring PC1).

## H. Implementation Notes

- IN-1: The four repaired lines keep their comments' intent — if a line carries an explanatory comment about the pipe (none of the four does today), update the comment with the change.
- IN-2: The trailing `|| true` on the four lines: `run_step` returns the status of its last command (a `printf` to the journal in both branches → 0), so the suffix is redundant armor. Keep or drop is an implementation detail with no observable difference today; if dropped, the four lines join the shape of the javac/kotlinc/scalac lines (`... || true` retained there). Default: keep `|| true` to minimize diff surface.
- IN-3: The `head -N` truncation governed what little crossed the pipe — nothing, since line 681 already redirects all checker output to the per-check log. The logs grow by the truncation delta (full checker output instead of first N lines); acceptable and arguably more useful; note it in the run report.

## I. Cross-References

- spec.md §B (verified premises), §E (out of scope), decision-index.md (Q1 repair-shape row, FOUNDER).
- `.moai/reports/t1395/` — RED harness + captured outputs.
