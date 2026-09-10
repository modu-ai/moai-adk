# plan.md — SPEC-CON-AMEND-APPLY-001

Card t659 · Tier M (3 artifacts) · development mode per `.moai/config/sections/quality.yaml` (TDD: every milestone opens with a RED measurement). Measurement tree for this plan: `.claude/worktrees/t659` @ `034d55c56`.

## §A Context

The five safety gates of SPEC-V3R2-CON-002 run, but the apply step behind them is a stub (`internal/constitution/pipeline.go:256-267`), dry-run skips it entirely (`pipeline.go:133-137`), and the evolution log is unreadable in both directions (`evolution_log.go:19-50`, `amendment.go:192-219`). The lead has ruled on every design question (`.moai/reports/t659/verdict.md` §7). This plan orders the work so the decisions most likely to change are reviewed first.

## §B Known Issues (measured — do not re-litigate)

| Issue | Source |
|---|---|
| Real apply stops at `updateSourceFile` stub; five characterization tests pin it | verdict §2.1 |
| Dry-run never calls an apply function | `pipeline.go:133-137` |
| Writer emits `ruleid`/`approvedat` and integer zones; snake_case logs read back empty | verdict §2.2 |
| Real log parses to 0 entries (`---` split meets `\|---\|` rows) | verdict §2.3 |
| 97/97 live registry clauses occur exactly once in their file; 4 retired entries occur 0 times | verdict §2.3 |
| `MarkRolledBack` has 0 production callers | grep for `MarkRolledBack(` over `internal cmd pkg` Go files excluding `_test.go` → definition line only (tree `034d55c56`) |
| Default lock path is cwd-relative | `pipeline.go:227` |

Gaps carried from the verdict: CLI-level execution was never observed (verdict §4). Restore-failure behaviour (a restore that itself fails) is not specified by the rulings — see §G risk R-4.

## §C Pre-flight (run at run-phase entry; stop and report on any mismatch)

### §C.1 Re-measure

```bash
git rev-parse --short HEAD
git branch --show-current
/usr/bin/grep -n 'not yet implemented' internal/constitution/pipeline.go     # expect 2 lines
/usr/bin/grep -c 'yaml:"' internal/constitution/amendment.go                  # expect 0 (no tags on AmendmentLog)
/usr/bin/grep -c 'yaml:"' internal/constitution/loader.go                     # control: expect >0 (rawEntry tags)
shasum -a 256 .claude/rules/moai/core/zone-registry.md .moai/research/evolution-log.md   # record; AC-CAA-017 compares
```

Tool provenance (verification-claim-integrity §2.2): any `moai spec lint` result cited as evidence names the judging build's commit next to the tree HEAD.

### §C.2 Disposition of the stub-characterization tests (REQ-CAA-014)

These tests must not be deleted silently. Each is replaced in the milestone named; the replacement AC owns the property.

| Existing test (`internal/constitution/pipeline_test.go`) | What it pins today | Replaced by | Milestone |
|---|---|---|---|
| `TestPipeline_Execute_NonDryRun_AmendmentStubError` (:145) | non-dry-run returns `amendment application error` | success path: three files updated (AC-CAA-001) **and** failure path: two-occurrence fixture returns error, files byte-identical, lock released (AC-CAA-002) | M4 |
| `TestPipeline_applyAmendment_StubError` (:335) | only reachable error is `source file update error` | `applyAmendment` success + validation failures (AC-CAA-002, AC-CAA-005) + fault injection (AC-CAA-012) | M4 |
| `TestUpdateSourceFile_StubError` (:405) | `not yet implemented` | exactly-once replacement, 0/2 occurrence rejection, no normalization (AC-CAA-001, 002, 003) | M2 |
| `TestUpdateRegistryClause_StubError` (:415) | `not yet implemented` | line-scoped rewrite + re-parse verification (AC-CAA-004, 005) | M2 |
| `TestPipeline_Execute_DryRun_Success` (:81) — assumption only | dry-run succeeds on a fixture whose rule file (`dummy.md`) does not exist | success assertion kept on a fixture whose rule file exists with the clause once; dry-run failure cases added (AC-CAA-014) | M5 |

Note on the last row: `writeTestRegistry` points every entry at a non-existent `dummy.md`. Under REQ-CAA-012 that fixture makes dry-run fail (the source file cannot be read), so the fixture changes — the test is updated, not weakened. `TestPipeline_Execute_CanaryUnavailable_Continues`, `TestPipeline_Execute_UserRejected`, `TestPipeline_Execute_HumanOversightError`, and the lock tests either stop before the apply step or run in dry-run; check each against the new fixture at M5 and record which ones needed the fixture change.

## §D Constraints

- [HARD] No test writes the real registry, rule files, or evolution log (REQ-CAA-015). Tests that read the real log do so read-only.
- [HARD] Lane-local verification is scoped: `go test ./internal/constitution/ -count=1` plus the named `internal/cli` selector, never the full suite. `internal/cli` runs need the compile slot granted at run-phase and `-timeout 600s`.
- Environment-scrubbed runs use one compound `unset … && go test …` invocation.
- A leftover temporary or backup file must never land as `*.md` inside `.claude/rules/**`, where Claude Code loads markdown as rules. Name them with a non-`.md` suffix (for example `.<base>.amend-tmp-<random>`), created in the target's own directory so the rename stays on one filesystem.
- No new dependency. Reuse `extractYAMLFence` and the `rawEntry` decoding for REQ-CAA-004 rather than a second parser.

## §E Self-Verification (run-phase fills `progress.md` §E.2)

- E1 AC matrix PASS/FAIL, each with its command and verbatim tail.
- E2 Mutant table: every mutant in `acceptance.md` §D.2 injected, the named AC observed RED, mutant reverted, AC GREEN again. An AC whose mutant could not be run is a Gap, not a PASS.
- E3 `go vet ./internal/constitution/ ./internal/cli/` and `golangci-lint run ./internal/constitution/... ./internal/cli/...`.
- E4 REQ-CAA-015 witness: real-file sha256 before and after the `internal/constitution` run; `internal/constitution/.moai` and `internal/cli/.moai` absent.

## §F Milestones (ordered by decision reversibility — most likely to change first)

### M1 — Evolution-log data model and reader (highest change likelihood)

Covers REQ-CAA-005 … REQ-CAA-009. The field mapping of REQ-CAA-008 and the fail-closed choice of REQ-CAA-009 are the decisions most likely to be revised (spec.md §G item 3), so they land and are reviewed first.

- **Baseline first (separate commit, verification-claim-integrity §2.3):** add the `RuleID` / `ApprovedAt` assertions to `TestLoadEvolutionLogs` and the human-format, legacy-key, horizontal-rule, and zone-name tests; run them against unchanged production code; commit the RED output under `.moai/reports/t659/` before any production change.
- Approach: yaml tags on `AmendmentLog`; a `Zone` yaml marshal to the zone name and an unmarshal accepting the name or the legacy integer; the reader decodes each candidate block into a generic mapping and resolves aliases per REQ-CAA-006/008 (a custom `UnmarshalYAML` or a map-then-assign step — run-phase choice). Candidate blocks: every fenced yaml block, plus `---`-delimited segments scanned line-anchored with a one-line advance whenever a segment is not an entry (never pairwise).
- Exit: AC-CAA-006 … AC-CAA-011 GREEN; mutants M-6, M-7, M-8a, M-8b, M-9, M-12 observed RED.

### M2 — Source and registry transforms, in memory (Q1/Q2 behaviour)

Covers REQ-CAA-001 … REQ-CAA-004. Pure functions from (bytes, current clause, new clause) to new bytes or an error; no file I/O yet. This keeps dry-run (M5) and real apply (M4) on one validation path.

- Baseline first: replacement tests for `updateSourceFile` / `updateRegistryClause` RED against the stubs.
- Registry rewrite: locate the fence as the loader does; locate the target entry by its `- id: <RuleID>` line and that entry's own `clause:` line before the next `- id:`; require exactly one such line; emit a single-line double-quoted yaml scalar for the new clause (escape `\` and `"`; a new clause containing a newline cannot be one line and is rejected through the REQ-CAA-004 error path); re-parse the whole candidate content through the loader's fence + `[]rawEntry` decode; compare entry count and target clause.
- Exit: AC-CAA-001 … AC-CAA-005 GREEN at function level; mutants M-1, M-2, M-3a, M-3b, M-4 RED.

### M3 — Rename seam and restore (Q4 mechanism)

Covers REQ-CAA-010, REQ-CAA-011.

- Seam: a `Pipeline` field holding the forward-rename function, defaulting to `os.Rename` when nil. A struct field rather than a package variable, so parallel tests cannot race on it. Restore writes backup bytes back directly (and removes files recorded absent) without calling the field.
- Sequence: validate (M2) → back up three files in memory (bytes + existed flag) → write three temps → rename source, registry, log through the seam → on any failure restore all three → remove temps.
- Log content = pre-apply bytes + one appended entry (REQ-CAA-010), which replaces the current `O_APPEND` writer on this path.
- Exit: AC-CAA-012, AC-CAA-013 GREEN; M-5a, M-5b, M-5c, M-11a RED.

### M4 — Wire into `Execute` / `applyAmendment`

Replace the stub calls; retire the stub tests per §C.2. Exit: AC-CAA-001, AC-CAA-002 GREEN through `Execute(dryRun=false)` with `fakeOversight` and a `t.TempDir()` lock path; AC-CAA-016 GREEN.

### M5 — Dry-run validation and CLI

Covers REQ-CAA-012, REQ-CAA-013. Dry-run calls the M2 validation against real file bytes and returns its error. Update `writeTestRegistry` fixtures (§C.2 last row). Add `internal/cli` dry-run tests (compile slot required). Exit: AC-CAA-014, AC-CAA-015 GREEN; M-10, M-11b, M-13 RED.

### M6 — Isolation witness and cleanup (mechanical, last)

AC-CAA-017 (with mutant M-14); `go vet`, lint; `progress.md` §E.2 evidence.

## §G Risks

| ID | Risk | Mitigation |
|---|---|---|
| R-1 | Existing fixtures point at a missing `dummy.md`; dry-run tests flip RED once validation runs | §C.2 disposition; fix fixtures, do not relax REQ-CAA-012 |
| R-2 | REQ-CAA-009 fail-closed blocks every amendment if the human log gains a malformed entry | the error names the entry id; spec.md §G item 3 |
| R-3 | A leftover temp file inside `.claude/rules/**` is loaded as a rule | non-`.md` suffix + cleanup asserted in AC-CAA-012 |
| R-4 | Restore itself fails (disk full, permission) — behaviour not covered by the rulings | run-phase must not delete backups when a restore fails and must return an error naming them; record as Residual-risk |
| R-5 | Once human entries are readable, the unused `MarkRolledBack` becomes lossy | 0 production callers; out of scope, recorded in spec.md §F |

## §H Anti-Patterns

- Deleting a stub test instead of replacing it (§C.2).
- A fault-injection test that never reaches the injected call. Always assert the seam's call count alongside the byte-identity check — an unreached injector and a correct restore look identical otherwise.
- Comparing file bytes by re-reading after the operation without a pre-captured snapshot.
- Unanchored `-run` selectors: every AC command anchors `^…$` and states the expected top-level `=== RUN` count.
- Verifying against the repository's real registry or log through any write-capable path.

## §I Cross-References

- `spec.md` §C (ruling map), §D (requirements), §F (exclusions), §G (open questions)
- `acceptance.md` §D (AC matrix), §D.2 (mutants)
- `.moai/reports/t659/verdict.md`
