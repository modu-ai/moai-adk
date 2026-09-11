# plan.md — SPEC-CON-AMEND-APPLY-001

Card t659 · Tier L (5 artifacts + progress.md; raised from Tier M in revision 0.1.4, verdict §13.3) · development mode per `.moai/config/sections/quality.yaml` (TDD: every milestone opens with a RED measurement). Code coordinates read at `034d55c56`; revision 0.1.1 authored on `ff11e752f` (verdict §8 added, no code change); revision 0.1.2 authored on `92c8c3f36` (verdict §9 added, no code change); revision 0.1.3 authored on `578afca87` (verdict §10–§11 and lint evidence added, no code change).

## §A Context

The five safety gates of SPEC-V3R2-CON-002 run, but the apply step behind them is a stub (`internal/constitution/pipeline.go:256-267`), dry-run skips it entirely (`pipeline.go:133-137`), the evolution log is unreadable in both directions (`evolution_log.go:19-50`, `amendment.go:192-219`), and the CLI validates a registry `Execute` may not be the one writing (`internal/cli/constitution.go:144-155` vs `pipeline.go:66`). The lead has ruled on every design question (`.moai/reports/t659/verdict.md` §7, §8, §9, §11); none is open. This plan orders the work so the decisions most likely to change are reviewed first.

## §B Known Issues (measured — do not re-litigate)

| Issue | Source |
|---|---|
| Real apply stops at `updateSourceFile` stub; five characterization tests pin it | verdict §2.1 |
| Dry-run never calls an apply function | `pipeline.go:133-137` |
| Writer emits `ruleid`/`approvedat` and integer zones; snake_case logs read back empty | verdict §2.2 |
| Real log parses to 0 entries (`---` split meets `\|---\|` rows) | verdict §2.3 |
| 97/97 live registry clauses occur exactly once in their file; 4 retired entries occur 0 times | verdict §2.3 |
| CLI resolves the registry by env precedence; `Execute` joins `projectDir` | `constitution.go:144-155`, `pipeline.go:66` |
| `LoadRegistry` rejects only an absolute registry path escaping `projectDir`; a relative path is not checked, and `applyAmendment` joins `file:` with no check — both closed by REQ-CAA-021 (verdict §11) | `loader.go:80-88`, `pipeline.go:192-195` |
| Real registry `file:` shape: 101 lines, 17 distinct, 87 under `.claude/`, 14 `CLAUDE.md`, 0 absolute, 0 containing `..`, 101 existing | read-only Python scan at `578afca87`, same figures re-measured at `4e9273d0b` (AC-CAA-025) |
| Every existing `Execute` test passes a `Before` equal to its fixture clause (REQ-CAA-017 breaks none) | `pipeline_test.go` lines 108-262 read at `ff11e752f` |
| No `t.Parallel` in the affected test files (so `t.Setenv` is usable) | `grep -c 't.Parallel()'` = 0 per file at `ff11e752f` |
| `MarkRolledBack` has 0 production callers | grep for `MarkRolledBack(` over `internal cmd pkg` Go files excluding `_test.go` → definition line only (tree `034d55c56`) |
| Default lock path is cwd-relative | `pipeline.go:227` |

Gaps carried: the non-dry-run CLI path (spec.md §E.4, approved reduction G5). G6 and G7 are resolved (REQ-CAA-020, REQ-CAA-021); no open question remains.

## §C Pre-flight (run at run-phase entry; stop and report on any mismatch)

### §C.1 Re-measure

```bash
git rev-parse --short HEAD
git branch --show-current
/usr/bin/grep -n 'not yet implemented' internal/constitution/pipeline.go     # expect 2 lines
/usr/bin/grep -c 'yaml:"' internal/constitution/amendment.go                  # expect 0 (no tags on AmendmentLog)
/usr/bin/grep -c 'yaml:"' internal/constitution/loader.go                     # control: expect >0 (rawEntry tags)
/usr/bin/grep -n 'func resolveRegistryPath' internal/cli/constitution.go      # expect 1 line
/usr/bin/grep -c 'EvalSymlinks' internal/constitution/loader.go internal/constitution/pipeline.go   # expect 0 each (no symlink resolution yet)
/usr/bin/grep -c 'filepath.IsAbs(cleanPath)' internal/constitution/loader.go  # expect 1 (absolute-only containment)
shasum -a 256 .claude/rules/moai/core/zone-registry.md .moai/research/evolution-log.md   # record; AC-CAA-017 compares
```

Tool provenance (verification-claim-integrity §2.2): any `moai spec lint` result cited as evidence names the judging build's commit next to the tree HEAD.

### §C.2 Disposition of existing tests (REQ-CAA-014)

These tests must not be deleted silently. Each is replaced in the milestone named; the replacement AC owns the property.

| Existing test (`internal/constitution/pipeline_test.go`) | What it pins today | Replaced by | Milestone |
|---|---|---|---|
| `TestPipeline_Execute_NonDryRun_AmendmentStubError` (:145) | non-dry-run returns `amendment application error` | success path: three files updated (AC-CAA-001) **and** failure path: two-occurrence fixture returns error, files byte-identical, lock released (AC-CAA-002) | M5 |
| `TestPipeline_applyAmendment_StubError` (:335) | only reachable error is `source file update error` | `applyAmendment` success + validation failures (AC-CAA-002, 005, 019) + fault injection (AC-CAA-012, 021) | M5 |
| `TestUpdateSourceFile_StubError` (:405) | `not yet implemented` | exactly-once replacement, 0/2 occurrence rejection, no normalization, new-clause pre-occurrence (AC-CAA-001, 002, 003, 019) | M3 |
| `TestUpdateRegistryClause_StubError` (:415) | `not yet implemented` | line-scoped rewrite + re-parse verification (AC-CAA-004, 005) | M3 |
| `TestPipeline_Execute_DryRun_Success` (:81) — assumption only | dry-run succeeds on a fixture whose rule file (`dummy.md`) does not exist | success assertion kept on a fixture whose rule file exists with the current clause once and the new clause absent; dry-run failure cases added (AC-CAA-014) | M6 |

Fixture consequences:
- `writeTestRegistry` points every entry at a non-existent `dummy.md`. Under REQ-CAA-012 dry-run must now fail on that fixture, so the fixture changes — the tests are updated, not weakened. Check `TestPipeline_Execute_CanaryUnavailable_Continues`, `TestPipeline_Execute_UserRejected`, `TestPipeline_Execute_HumanOversightError`, `TestPipeline_Execute_FrozenGuard_Rejects`, `TestPipeline_Execute_RuleNotFound`, and `TestPipeline_Execute_LockBusy` against the new fixture at M6 and record which ones changed.
- Every test that calls `Execute` or `runConstitutionAmend` sets `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` with `t.Setenv` (REQ-CAA-015); a shared fixture helper is the natural place.

## §D Constraints

- [HARD] No test writes the real registry, rule files, or evolution log (REQ-CAA-015). Tests that read the real log do so read-only.
- [HARD] Lane-local verification is scoped: `go test ./internal/constitution/ -count=1` plus the named `internal/cli` selectors, never the full suite. `internal/cli` runs need the compile slot granted at run-phase and `-timeout 600s`.
- Environment-scrubbed runs use one compound `unset … && go test …` invocation. The tests must not depend on that scrub — they set the variables themselves (REQ-CAA-015); AC-CAA-017 deliberately runs once without it.
- A leftover temporary or backup file must never land as `*.md` inside `.claude/rules/**`, where Claude Code loads markdown as rules. Name them with a non-`.md` suffix (for example `.<base>.amend-tmp-<random>` and `.<base>.amend-bak-<random>`), created in the target's own directory so the rename stays on one filesystem.
- Tests that use `t.Chdir` (AC-CAA-024 `relative_env_escape`, `in_root_control`, and its CLI case) change the process working directory; they stay non-parallel and pin the lock path, so the cwd-relative default lock path is never created.
- No new dependency. Reuse `extractYAMLFence` and the `rawEntry` decoding for REQ-CAA-004 rather than a second parser.
- The resolver reads env names from constants, not string literals (CLAUDE.local.md §14): reuse `config.EnvClaudeProjectDir` and move or re-export the existing `MOAI_CONSTITUTION_REGISTRY` constant rather than duplicating it.

## §E Self-Verification (run-phase fills `progress.md` §E.2)

- E1 AC matrix PASS/FAIL, each with its command and verbatim tail.
- E2 Mutant table: every mutant in `acceptance.md` §D.2 injected, the named AC observed RED, mutant reverted, AC GREEN again. An AC whose mutant could not be run is a Gap, not a PASS.
- E3 `go vet ./internal/constitution/ ./internal/cli/` and `golangci-lint run ./internal/constitution/... ./internal/cli/...`.
- E4 REQ-CAA-015 witness: real-file sha256 before and after the runs of AC-CAA-017; `internal/constitution/.moai` and `internal/cli/.moai` absent.

## §F Milestones (ordered by decision reversibility — most likely to change first)

### M1 — Evolution-log data model and reader (highest change likelihood)

Covers REQ-CAA-005 … REQ-CAA-009. The field mapping of REQ-CAA-008 and the error shape of REQ-CAA-009 are data-model decisions every later milestone reads.

- **Baseline first (separate commit, verification-claim-integrity §2.3):** add the `RuleID` / `ApprovedAt` assertions to `TestLoadEvolutionLogs` and the human-format, legacy-key, horizontal-rule, zone-name, and error-location tests; run them against unchanged production code; commit the RED output under `.moai/reports/t659/` before any production change.
- Approach: yaml tags on `AmendmentLog`; a `Zone` yaml marshal to the zone name and an unmarshal accepting the name or the legacy integer; the reader decodes each candidate block into a `yaml.Node` (which carries line numbers) and resolves aliases per REQ-CAA-006/008. Candidate blocks: every fenced yaml block, plus `---`-delimited segments scanned line-anchored with a one-line advance whenever a segment is not an entry (never pairwise). Line numbers in REQ-CAA-009 errors are file lines: node line + the block's starting line offset.
- Exit: AC-CAA-006 … AC-CAA-011 and AC-CAA-018 GREEN; mutants M-6, M-7, M-8a, M-8b, M-9, M-12, M-15 observed RED.

### M2 — Shared registry path resolver (new cross-package interface)

Covers REQ-CAA-019, REQ-CAA-020, and REQ-CAA-021. A new function both packages call is an interface decision; it lands before the apply wiring that depends on it.

- Baseline first: AC-CAA-022 RED against the current `Execute` (it joins `projectDir` and cannot load the registry at the env path).
- Approach: move the precedence of `resolveRegistryPath` into `internal/constitution` (for example `ResolveRegistryPath(projectDir string) string`); `internal/cli.resolveRegistryPath` becomes a thin call to it or is replaced; `Execute` calls it instead of its own join.
- Containment (REQ-CAA-020, REQ-CAA-021): one check function — clean, make absolute, resolve symbolic links (nearest existing ancestor for a path not yet created), compare with the resolved root at a separator boundary — called for the registry path and for every entry's joined `file:` at registry load, and for the evolution-log path before Layer 1. It replaces the loader's absolute-only refusal. The source rule file and the log keep their `projectDir` joins; the check only admits or refuses them. Baseline-first: AC-CAA-023 RED against current code (`Execute` ignores `CLAUDE_PROJECT_DIR`, loads the registry inside `projectDir`, and returns dry-run success or the stub error instead of a registry load error). AC-CAA-024's five escape rows are RED the same way, since no check exists for them.
- Exit: AC-CAA-022 GREEN; M-19 RED; AC-CAA-023 subtests `divergent_root_real` and `divergent_root_dry_run`, AC-CAA-024's five escape rows (real and dry-run), and AC-CAA-024's CLI case (compile slot) GREEN; M-20, M-21 (both variants), M-22 (both variants), and M-23 RED. AC-CAA-023 `same_root_control`, the real cases of AC-CAA-024 `in_root_control`, and AC-CAA-025 `real` turn GREEN at M5, once the apply is wired. Existing `internal/cli` constitution tests re-run with the compile slot.

### M3 — Source and registry transforms, in memory (Q1/Q2, G2)

Covers REQ-CAA-001 … REQ-CAA-004, REQ-CAA-016. Pure functions from (bytes, current clause, new clause) to new bytes or an error; no file I/O yet, so dry-run (M6) and real apply (M5) share one validation path.

- Baseline first: replacement tests for `updateSourceFile` / `updateRegistryClause` RED against the stubs.
- Order of checks on the source: new-clause occurrence count must be 0 (REQ-CAA-016), then current-clause count must be 1 (REQ-CAA-001/002). Both errors name path and count.
- Registry rewrite: locate the fence as the loader does; locate the target entry by its `- id: <RuleID>` line and that entry's own `clause:` line before the next `- id:`; require exactly one such line; emit a single-line double-quoted yaml scalar for the new clause (escape `\` and `"`; a new clause containing a newline cannot be one line and is rejected through the REQ-CAA-004 error path); re-parse the whole candidate content through the loader's fence + `[]rawEntry` decode; compare entry count and target clause.
- Exit: AC-CAA-001 … AC-CAA-005 and AC-CAA-019 GREEN at function level; mutants M-1, M-2, M-3a, M-3b, M-4, M-16 RED.

### M4 — Seams, backups, and restore (Q4, G4)

Covers REQ-CAA-010, REQ-CAA-011, REQ-CAA-018.

- Seams: two `Pipeline` fields — forward rename (default `os.Rename`) and restore write (default: write backup bytes to the target, or remove it when recorded absent). Struct fields rather than package variables, so parallel tests cannot race on them.
- Sequence: validate (M3) → write backup files → write three temps → rename source, registry, log through the rename seam → on failure restore all three through the restore seam → on complete restore remove temps and backups → on any restore failure remove temps but keep every backup and return an error listing all backup paths and the failed restore step.
- Log content = pre-apply bytes + one appended entry (REQ-CAA-010), which replaces the current `O_APPEND` writer on this path.
- Exit: AC-CAA-012, AC-CAA-013, AC-CAA-021 GREEN; M-5a, M-5b, M-5c, M-11a, M-18 RED.

### M5 — Wire into `Execute` / `applyAmendment` (G3)

Replace the stub calls; add the REQ-CAA-017 `Before` check after registry lookup and before Layer 1 (both modes); retire the stub tests per §C.2. Exit: AC-CAA-001, AC-CAA-002, AC-CAA-020 GREEN through `Execute(dryRun=false)` with `fakeOversight` and a `t.TempDir()` lock path; AC-CAA-016, AC-CAA-023, AC-CAA-024, and AC-CAA-025 GREEN; M-17, M-20, and M-24 RED.

### M6 — Dry-run validation and CLI

Covers REQ-CAA-012, REQ-CAA-013. Dry-run calls the M3 validation against real file bytes at the resolved paths and returns its error. Update fixtures (§C.2). Add `internal/cli` dry-run tests (compile slot required). Exit: AC-CAA-014, AC-CAA-015 GREEN; M-10, M-11b, M-13 RED.

### M7 — Isolation witness and cleanup (mechanical, last)

AC-CAA-017 (with mutant M-14); `go vet`, lint; `progress.md` §E.2 evidence.

## §G Risks

| ID | Risk | Mitigation |
|---|---|---|
| R-1 | Existing fixtures point at a missing `dummy.md`; dry-run tests flip RED once validation runs | §C.2 disposition; fix fixtures, do not relax REQ-CAA-012 |
| R-2 | REQ-CAA-009 fail-closed blocks every amendment if the human log gains a malformed entry | confirmed by the lead (G1); the error names path, line, and key so a human can fix it |
| R-3 | A leftover temp or backup file inside `.claude/rules/**` is loaded as a rule | non-`.md` suffixes; cleanup asserted in AC-CAA-012; a retained backup after a failed restore (REQ-CAA-018) is still non-`.md` |
| R-4 | After REQ-CAA-019, a session-exported `CLAUDE_PROJECT_DIR` steers any test that reaches the resolver toward the real registry | REQ-CAA-015 obliges tests to set both variables; AC-CAA-017 runs once with `CLAUDE_PROJECT_DIR` pointed at the repository root; the containment check is a second line of defence and is required behaviour (REQ-CAA-020, REQ-CAA-021, AC-CAA-023, AC-CAA-024) |
| R-5 | REQ-CAA-016 rejects a new clause that is a substring of the current clause (e.g. shortening a sentence by removing its tail) | literal consequence of the G2 ruling, stated in REQ-CAA-016; the user rewrites the proposal |
| R-6 | Once human entries are readable, the unused `MarkRolledBack` becomes lossy | 0 production callers; follow-up candidate (spec.md §F) |
| R-7 | An over-strict containment check refuses legitimate paths — a root reached through a symbolic link (the macOS temp directory is one), or a log not yet created | REQ-CAA-021 resolves both sides and judges a missing path by its nearest existing ancestor; AC-CAA-024 `in_root_control`, AC-CAA-025, M-24 |
| R-8 | Refusal at registry load also reaches commands that load the registry without amending (`moai constitution list`, `guard`) and entries the amendment does not target | intended by verdict §11 (load error); the real registry has 0 absolute and 0 `..` `file:` values of 101, so no current registry is refused (AC-CAA-025) |
| R-9 | A symbolic link swapped between the check and the write | not a requirement of this SPEC — whoever can rewrite links inside the project root already controls its files; recorded for plan-audit as residual risk |

## §H Anti-Patterns

- Deleting a stub test instead of replacing it (§C.2).
- A fault-injection test that never reaches the injected call. Always assert each seam's call count alongside the byte-identity or file-presence check — an unreached injector and a correct restore look identical otherwise.
- Comparing file bytes by re-reading after the operation without a pre-captured snapshot.
- Unanchored `-run` selectors: every AC command anchors `^…$` and states the expected top-level `=== RUN` count.
- Relying on the shell's `unset` for isolation instead of setting the variables in the test.
- Verifying against the repository's real registry or log through any write-capable path.

## §I Cross-References

- `spec.md` §C (ruling map), §D (requirements), §E.4 (approved gap), §F (exclusions), §G (no open question)
- `acceptance.md` §D (AC matrix), §D.2 (mutants)
- `design.md` (design decisions and rejected alternatives), `research.md` (codebase findings and provenance)
- `progress.md` (run-phase evidence skeleton)
- `.moai/reports/t659/verdict.md` §7, §8, §9, §11
