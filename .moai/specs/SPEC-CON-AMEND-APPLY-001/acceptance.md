# acceptance.md — SPEC-CON-AMEND-APPLY-001

Verification layer. Each AC is a Given-When-Then scenario with its verifying command and, where the property can pass vacuously, a mutant that must turn it RED. GEARS obligations live in `spec.md` §D. Test names are proposals; run-phase fixes the final names and records them in `progress.md` §E.2.

Common rules for every AC:
- Fixtures live under `t.TempDir()`: a project dir with `.claude/rules/moai/core/zone-registry.md`, a real rule file the target entry points at, and `.moai/research/evolution-log.md` where needed. The lock path is pinned inside the temp dir; `fakeOversight` approves non-dry-run runs.
- "Byte-identical" means equal to a sha256 captured **before** the operation. "No leftover files" means the set of paths under the fixture dir equals the set captured before (plus exactly the intended changes on success).
- Selectors are anchored `^…$`. A run whose top-level `=== RUN` count is below the stated number is not a PASS.
- ACs marked **baseline-first** are run against unchanged production code and their RED output is committed before the production change (plan.md §F).

## §D AC Matrix

| AC | Requirement | RED at `034d55c56` | Mutant(s) |
|---|---|---|---|
| AC-CAA-001 | REQ-CAA-001, REQ-CAA-014 | stub error (verdict §2.1) — baseline-first | M-1 |
| AC-CAA-002 | REQ-CAA-002, REQ-CAA-014 | stub error, not an occurrence error — baseline-first | M-1 |
| AC-CAA-003 | REQ-CAA-001, REQ-CAA-002 | stub error — baseline-first | M-2 |
| AC-CAA-004 | REQ-CAA-003 | stub error — baseline-first | M-3a, M-3b |
| AC-CAA-005 | REQ-CAA-004 | stub error — baseline-first | M-4 |
| AC-CAA-006 | REQ-CAA-005 | writer emits `ruleid:` / `zonebefore: 0` (verdict §2.2) — baseline-first | M-9 |
| AC-CAA-007 | REQ-CAA-006 | none expected — the untagged decoder reads legacy keys today; invariant guard whose RED cell is the mutant | M-6 |
| AC-CAA-008 | REQ-CAA-006, REQ-CAA-014 | `RuleID` empty on the snake_case fixture (verdict §2.2) — baseline-first | M-7 |
| AC-CAA-009 | REQ-CAA-007 | `---` split fabricates or drops entries (verdict §2.3) — baseline-first | M-12 |
| AC-CAA-010 | REQ-CAA-008, REQ-CAA-009 | 0 entries from the real log (verdict §2.3) — baseline-first | M-8a, M-8b |
| AC-CAA-011 | REQ-CAA-008 | limiter admits (sees 0 entries) — baseline-first | M-8a |
| AC-CAA-012 | REQ-CAA-010, REQ-CAA-011 | seam does not exist (compile failure, not discriminating) — the discriminating RED is the mutants | M-5a, M-5b, M-11a |
| AC-CAA-013 | REQ-CAA-010, REQ-CAA-011 | seam does not exist — the mutant is the RED | M-5c |
| AC-CAA-014 | REQ-CAA-012, REQ-CAA-014 | dry-run returns success on a two-occurrence fixture — baseline-first | M-10, M-11b |
| AC-CAA-015 | REQ-CAA-013 | CLI dry-run prints success on a two-occurrence fixture — baseline-first (needs compile slot) | M-13 |
| AC-CAA-016 | REQ-CAA-014 | `not yet implemented` present in pipeline.go and pipeline_test.go | — (grep with control) |
| AC-CAA-017 | REQ-CAA-015 | none expected — invariant guard | M-14 |

### §D.0 REQ coverage (machine-readable)

- AC-CAA-001 maps REQ-CAA-001, REQ-CAA-014
- AC-CAA-002 maps REQ-CAA-002, REQ-CAA-014
- AC-CAA-003 maps REQ-CAA-001, REQ-CAA-002
- AC-CAA-004 maps REQ-CAA-003
- AC-CAA-005 maps REQ-CAA-004
- AC-CAA-006 maps REQ-CAA-005
- AC-CAA-007 maps REQ-CAA-006
- AC-CAA-008 maps REQ-CAA-006, REQ-CAA-014
- AC-CAA-009 maps REQ-CAA-007
- AC-CAA-010 maps REQ-CAA-008, REQ-CAA-009
- AC-CAA-011 maps REQ-CAA-008
- AC-CAA-012 maps REQ-CAA-010, REQ-CAA-011
- AC-CAA-013 maps REQ-CAA-010, REQ-CAA-011
- AC-CAA-014 maps REQ-CAA-012, REQ-CAA-014
- AC-CAA-015 maps REQ-CAA-013
- AC-CAA-016 maps REQ-CAA-014
- AC-CAA-017 maps REQ-CAA-015

The union is REQ-CAA-001 … REQ-CAA-015; no requirement lacks an AC.

## §D.1 AC Details

### AC-CAA-001 — exactly-once replacement lands in all three files

- **Given** a fixture whose Evolvable rule file contains the current clause exactly once, surrounded by other text, and a registry with at least three entries,
- **When** `Execute(dryRun=false)` runs with an approving oversight double,
- **Then** it returns a log entry with no error, and the rule file equals its pre-apply bytes with that one span replaced by the new clause,
- **And** `LoadRegistry` on the registry returns the new clause for the target entry and the unchanged clause for every other entry,
- **And** `LoadEvolutionLogs` returns exactly one more entry than before, whose `RuleID`, `ClauseBefore`, `ClauseAfter`, and non-zero `ApprovedAt` match the proposal,
- **And** the lock file is released.

Command: `go test ./internal/constitution/ -run '^TestApply_ExactlyOnce_Success$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-002 — zero or multiple occurrences fail before any write

- **Given** two fixtures: (a) the rule file lacks the current clause; (b) it contains the current clause twice,
- **When** `Execute(dryRun=false)` runs on each,
- **Then** each returns an error whose text names the rule file path and the count (`0` / `2`),
- **And** the rule file, registry, and log are byte-identical to their pre-apply snapshots and no path was added under the fixture dir,
- **And** the lock file is released.

Command: `go test ./internal/constitution/ -run '^TestApply_OccurrenceCount_Rejected$' -count=1 -v` — expect 1 top-level RUN with subtests `zero` and `two` both reported.

### AC-CAA-003 — no whitespace normalization

- **Given** a rule file where the clause text appears once with a doubled internal space, and once more elsewhere wrapped across a newline, while the registry clause uses single spaces on one line,
- **When** the apply step validates,
- **Then** the exact-match count is `0`, the apply fails, and all three files are byte-identical.

Command: `go test ./internal/constitution/ -run '^TestApply_NoWhitespaceNormalization$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-004 — registry rewrite touches one line and round-trips

- **Given** a registry fence with header comments, blank lines between entries, at least three entries, and a new clause containing `"`, `\`, `: `, and `#`,
- **When** the registry update runs,
- **Then** a line-by-line comparison with the pre-apply registry differs in exactly one line — the target entry's `clause:` line — and the line count is unchanged,
- **And** `LoadRegistry` decodes the target clause as exactly the new clause.

Command: `go test ./internal/constitution/ -run '^TestUpdateRegistryClause_SingleLineRoundTrip$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-005 — re-parse verification blocks a corrupt registry

- **Given** three fixtures: (a) the target entry's clause is a double-quoted scalar continued onto a second line, so rewriting only the `clause:` line leaves an orphaned continuation; (b) the target entry has no `clause:` line; (c) the new clause contains a newline,
- **When** `Execute(dryRun=false)` runs on each,
- **Then** each returns an error and all three files are byte-identical to their pre-apply snapshots.

Command: `go test ./internal/constitution/ -run '^TestApply_RegistryReparse_Rejects$' -count=1 -v` — expect 1 top-level RUN with subtests `continuation`, `no_clause_line`, `newline_clause`.

### AC-CAA-006 — writer emits snake_case keys and zone names

- **Given** an `AmendmentLog` with `ZoneBefore = ZoneAfter = ZoneEvolvable` appended to an empty temp log,
- **When** the raw file bytes are read,
- **Then** they contain `rule_id:`, `approved_at:`, `zone_before: Evolvable`, and `zone_after: Evolvable`,
- **And** they contain none of `ruleid:`, `approvedat:`, `zonebefore:`, or a zone line whose value is an integer,
- **And** a control appending `ZoneFrozen` yields `zone_before: Frozen` (a writer that always prints one name cannot pass both).

Command: `go test ./internal/constitution/ -run '^TestAppendEvolutionLog_SnakeCaseAndZoneNames$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-007 — legacy concatenated keys and integer zones still read

- **Given** a log written in the legacy shape observed in verdict §2.2 (`ruleid: CONST-V3R2-003`, `zonebefore: 1`, `approvedat: 2026-09-11T00:00:00Z`, …),
- **When** `LoadEvolutionLogs` reads it,
- **Then** the entry has `RuleID = CONST-V3R2-003`, `ZoneBefore = ZoneEvolvable`, and `ApprovedAt = 2026-09-11T00:00:00Z`,
- **And** a second entry carrying both `rule_id: A` and `ruleid: B` reads `RuleID = A`.

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs_LegacyKeys$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-008 — TestLoadEvolutionLogs asserts RuleID and ApprovedAt

- **Given** the existing snake_case fixtures in `TestLoadEvolutionLogs`,
- **When** the test runs,
- **Then** it asserts `ID`, `RuleID = CONST-V3R2-008`, and `ApprovedAt = 2026-04-28T10:00:00Z` for the single-entry case, and `RuleID` for both entries of the multi-entry case.

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs$' -count=1 -v` — expect 1 top-level RUN.

Baseline-first: with only the assertions added and production code unchanged, this test must FAIL on `RuleID` (verdict §2.2 observed `ruleID=""`). A PASS at that point means the assertion is not reaching the field.

### AC-CAA-009 — markdown rules and table separators do not disturb parsing

- **Given** a log with a HISTORY table containing `|---|` separator rows, two standalone `---` horizontal rules, prose between them, and then two machine `---`-delimited entries,
- **When** `LoadEvolutionLogs` reads it,
- **Then** it returns exactly those two entries with their correct `ID` and `RuleID`,
- **And** the same fixture with a third horizontal rule inserted before the first entry still returns exactly two (the count must not depend on parity).

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs_MarkdownNoise$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-010 — human-format entries map as specified; malformed ones fail closed

- **Given** (a) a temp log holding the `## EVO-HRN-002` section copied verbatim from `.moai/research/evolution-log.md` (heading, fenced yaml block, the surrounding `---` rules and the HISTORY table),
- **When** `LoadEvolutionLogs` reads it,
- **Then** it returns one entry with `ID = EVO-HRN-002`, `RuleID = CONST-V3R2-153`, `ApprovedAt = 2026-05-13T00:00:00Z`, `ZoneBefore = ZoneAfter = ZoneFrozen`, `RolledBack = false`, `RollbackAt = nil`;
- **And given** (b) the repository's real `.moai/research/evolution-log.md` opened read-only, **then** the returned entries include one with `ID = EVO-HRN-002` (a drift witness for the real format; this sub-case must not write);
- **And given** (c) a human-format block with `id: EVO-X-001`, `timestamp: "not-a-date"`, and no `approved_at`, **then** `LoadEvolutionLogs` returns an error whose text contains `EVO-X-001`.

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs_HumanFormat$' -count=1 -v` — expect 1 top-level RUN with subtests `verbatim_block`, `real_file_readonly`, `malformed_timestamp_fails_closed`.

### AC-CAA-011 — the rate limiter now sees the human entry

- **Given** a temp log containing only the verbatim EVO-HRN-002 block and a `rateLimiter` whose `now` is `2026-05-13T01:00:00Z`,
- **When** `Admit` runs for any proposal,
- **Then** it returns `*ErrRateLimitExceeded` (24-hour cooldown from the entry's `ApprovedAt`),
- **And** with `now = 2026-06-13T00:00:00Z` on the same log it returns nil — so the rejection is caused by the entry, not by an always-reject limiter.

Command: `go test ./internal/constitution/ -run '^TestRateLimiter_SeesHumanFormatEntry$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-012 — failed 2nd or 3rd rename restores all three files

- **Given** a valid fixture whose log file exists, and a pipeline whose rename seam fails on call N (N = 2, then separately N = 3) and otherwise delegates to `os.Rename`,
- **When** `Execute(dryRun=false)` runs,
- **Then** it returns an error naming the failed rename,
- **And** the seam was called exactly N times (reachability — without this, an injector that never fires is indistinguishable from a correct restore),
- **And** the rule file, registry, and log are byte-identical to their pre-apply sha256 snapshots,
- **And** no temporary or backup path remains under the fixture dir;
- **And given** the same N = 3 case on a fixture whose log file does not exist yet, **then** after the failure the log path does not exist;
- **And given** no injected failure, **then** the apply succeeds and no temporary or backup path remains.

Command: `go test ./internal/constitution/ -run '^TestApply_RenameFault_RestoresAll$' -count=1 -v` — expect 1 top-level RUN with subtests `second_rename`, `third_rename`, `third_rename_log_absent`, `no_fault_clean`.

### AC-CAA-013 — rename order is source, registry, log

- **Given** a recording rename seam that delegates to `os.Rename`,
- **When** a successful apply runs,
- **Then** the recorded destinations are, in order, the rule file path, the registry path, and the log path, and there are exactly three.

Command: `go test ./internal/constitution/ -run '^TestApply_RenameOrder$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-014 — dry-run runs the validation and writes nothing

- **Given** four fixtures: (a) valid; (b) clause occurs twice; (c) registry continuation line (AC-CAA-005 a); (d) rule file missing,
- **When** `Execute(dryRun=true)` runs on each,
- **Then** (a) returns a log entry and no error; (b), (c), (d) each return an error of the same kind a real apply returns on that fixture,
- **And** for all four, the snapshot of every path and sha256 under the fixture dir is identical before and after, and no lock file was created.

Command: `go test ./internal/constitution/ -run '^TestPipeline_Execute_DryRun_Validates$' -count=1 -v` — expect 1 top-level RUN with subtests `valid`, `two_occurrences`, `registry_continuation`, `missing_rule_file`. The former `TestPipeline_Execute_DryRun_Success` is folded into subtest `valid` or kept with an existing rule file (plan.md §C.2).

### AC-CAA-015 — CLI dry-run surfaces the validation failure

- **Given** an `internal/cli` test that sets `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` to empty with `t.Setenv`, and a `t.TempDir()` project fixture whose rule file contains the current clause twice,
- **When** it calls `runConstitutionAmend(stdout, stderr, projectDir, ruleID, before, after, "", true)`,
- **Then** the call returns a non-nil error and stdout does not contain `Dry-run success`,
- **And** on the valid-fixture control the call returns nil, stdout contains `Dry-run success`, and the fixture snapshot is unchanged.

Command (compile slot required at run-phase): `unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_DRY_RUN && go test ./internal/cli/ -run '^TestConstitutionAmend_DryRun_SurfacesValidation$' -count=1 -v -timeout 600s` — expect 1 top-level RUN with subtests `two_occurrences` and `valid`.

### AC-CAA-016 — no stub characterization remains

- **Given** the tree after M4,
- **When** `/usr/bin/grep -rn 'not yet implemented' internal/constitution/pipeline.go internal/constitution/pipeline_test.go` runs,
- **Then** it prints 0 lines,
- **And** the control `/usr/bin/grep -c 'func (p \*Pipeline) Execute' internal/constitution/pipeline.go` prints `1` (the grep reached the file),
- **And** `/usr/bin/grep -nE 'func (TestPipeline_Execute_NonDryRun_AmendmentStubError|TestPipeline_applyAmendment_StubError|TestUpdateSourceFile_StubError|TestUpdateRegistryClause_StubError)\(' internal/constitution/pipeline_test.go` prints 0 lines, and each replacement named in plan.md §C.2 exists (`/usr/bin/grep -c` per name ≥ 1).

### AC-CAA-017 — real files untouched by the package run

- **Given** sha256 of `.claude/rules/moai/core/zone-registry.md` and `.moai/research/evolution-log.md` recorded before the run,
- **When** `go test ./internal/constitution/ -count=1` and the AC-CAA-015 command complete,
- **Then** both sha256 values are unchanged, `git status --porcelain -- .claude/rules .moai/research` shows nothing attributable to the run, and neither `internal/constitution/.moai` nor `internal/cli/.moai` exists.

## §D.2 Mutant list (each must turn its AC RED; record in progress.md §E.2)

| ID | Mutation | Must turn RED |
|---|---|---|
| M-1 | Replace all occurrences (or the first) instead of requiring exactly one | AC-CAA-002 (b) |
| M-2 | Normalize whitespace before counting | AC-CAA-003 |
| M-3a | Re-serialize the parsed registry instead of rewriting one line | AC-CAA-004 (line diff ≠ 1) |
| M-3b | Interpolate the new clause into `"…"` without escaping `"` and `\` | AC-CAA-004 (round-trip) |
| M-4 | Skip the re-parse check | AC-CAA-005 (a) |
| M-5a | Remove the restore call on rename failure | AC-CAA-012 `second_rename`, `third_rename` |
| M-5b | Restore only files already renamed, or skip removing a log recorded absent | AC-CAA-012 `third_rename_log_absent` |
| M-5c | Rename the log before the registry | AC-CAA-013 |
| M-6 | Drop the legacy concatenated-key aliases | AC-CAA-007 |
| M-7 | Drop the snake_case tag on `rule_id` (read side) | AC-CAA-008 |
| M-8a | Ignore fenced yaml blocks (human format) | AC-CAA-010 (a), AC-CAA-011 |
| M-8b | Treat an unparseable human timestamp as zero time instead of an error | AC-CAA-010 (c) |
| M-9 | Serialize zone as an integer | AC-CAA-006 |
| M-10 | Dry-run skips apply validation (today's behaviour) | AC-CAA-014 (b), (c), (d) |
| M-11a | Leave temp or backup files after success | AC-CAA-012 `no_fault_clean` |
| M-11b | Dry-run performs the real apply (writes) | AC-CAA-014 snapshot |
| M-12 | Pairwise `---` split (today's parser) | AC-CAA-009 |
| M-13 | CLI dry-run ignores the pipeline error and prints success | AC-CAA-015 `two_occurrences` |
| M-14 | A test writes a fixture path resolved against the repository root | AC-CAA-017 |

A mutant that cannot be injected, or whose AC run shows fewer top-level RUN lines than stated, is recorded as a Gap, not a kill.

## §D.3 Definition of Done

- All 17 ACs GREEN with commands and verbatim tails in `progress.md` §E.2; baseline-first REDs committed before their production change.
- All 19 mutants observed RED and reverted.
- `go vet` and `golangci-lint` clean on `internal/constitution` and `internal/cli`.
- The five tests in plan.md §C.2 replaced, none silently deleted.
- plan.md §C.1 real-file sha256 equals the post-run sha256 (AC-CAA-017).
