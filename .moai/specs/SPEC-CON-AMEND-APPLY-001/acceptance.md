# acceptance.md — SPEC-CON-AMEND-APPLY-001

Verification layer. Each AC is a Given-When-Then scenario with its verifying command and, where the property can pass vacuously, a mutant that must turn it RED. GEARS obligations live in `spec.md` §D. Test names are proposals; run-phase fixes the final names and records them in `progress.md` §E.2.

Revision 0.1.1 (same as spec.md HISTORY): AC-CAA-018 … AC-CAA-022 and mutants M-15 … M-19 appended for the verdict §8 rulings; AC-CAA-012 amended for on-disk backups; AC-CAA-014 and AC-CAA-017 amended. Existing IDs unchanged.

Revision 0.1.2 (same as spec.md HISTORY): AC-CAA-023 and mutant M-20 appended for the verdict §9 ruling on G6; AC-CAA-022 gains a note on how it fits with AC-CAA-023. Existing IDs unchanged.

Revision 0.1.3 (same as spec.md HISTORY): AC-CAA-024, AC-CAA-025 and mutants M-21 … M-24 appended for the verdict §11 ruling on G7; AC-CAA-023 now asserts the offending path instead of the loader's present wording; M-20 now names the registry-path site of the one containment check and also turns AC-CAA-024 `relative_env_escape` and its CLI case RED. Existing IDs unchanged.

Revision 0.1.5 (same as spec.md HISTORY): plan-audit iteration 1 defects fixed without new IDs. AC-CAA-012 gains a rename-then-fail injector and subtests `first_rename_applied` and `third_rename_applied`; AC-CAA-024 gains rows `symlinked_registry` and `symlinked_file` and a discriminating CLI case; AC-CAA-003, AC-CAA-005 (b), AC-CAA-007, AC-CAA-014, and AC-CAA-025 amended; M-5b and M-23 split into variants, and the registry-path mutant gains a CLI-only variant; every RED statement is labelled a code-reading prediction. Existing IDs unchanged.

Revision 0.1.6 (same as spec.md HISTORY): operator decision D4 — the containment check runs on the amend path only, and `LoadRegistry` keeps its present behaviour. AC-CAA-024 gains the preservation case `loader_unchanged` and AC-CAA-025 the preservation run of `TestLinter_AC08_DanglingRuleReference` (no new AC ID); M-20 gains variant (iii), which moves the check into `LoadRegistry` and must turn both preservation rows RED; M-20 variant (i) no longer claims AC-CAA-023's divergent rows, which the loader's retained absolute-only refusal also stops, and M-19 now names them; M-24 moves its AC-CAA-025 kill from `load` to `dry_run`; AC-CAA-022 and AC-CAA-023 reworded where they said the loader decides admission. Existing IDs unchanged.

Revision 0.1.7 (same as spec.md HISTORY): plan-audit iteration 2 defects N1–N6 fixed without new IDs. The common rules gain a numeric-assertion rule — every count and line-number assertion is checked on the error string with every fixture path removed and matched as a whole number — and AC-CAA-002, AC-CAA-003, AC-CAA-014, AC-CAA-018, and AC-CAA-019 state their numbers under it; AC-CAA-003 and AC-CAA-018 also assert that the number a wrong implementation would print is absent. AC-CAA-024 and AC-CAA-025 build their base with `filepath.EvalSymlinks(t.TempDir())`, M-22 (ii) compares against the unresolved `projectDir`, and the CLI case of AC-CAA-024 admits both path forms. AC-CAA-025's preservation assertion ranges from a pinned `BASELINE_SHA`. M-20 gains variant (iv), the RED cell of `loader_unchanged` case `absolute_escape`; the M-2, M-10, M-15, M-21, and M-22 rows are corrected. Existing IDs unchanged.

Common rules for every AC:
- Fixtures live under `t.TempDir()`: a project dir with `.claude/rules/moai/core/zone-registry.md`, a real rule file the target entry points at (current clause once, new clause absent unless the AC says otherwise), and `.moai/research/evolution-log.md` where needed. The lock path is pinned inside the temp dir; `fakeOversight` approves non-dry-run runs; the proposal's `Before` equals the current clause unless the AC says otherwise.
- Every test that calls `Execute` or `runConstitutionAmend` sets `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` with `t.Setenv` (empty unless the AC says otherwise) — REQ-CAA-015.
- "Byte-identical" means equal to a sha256 captured **before** the operation. "No leftover files" means the set of paths under the fixture dir equals the set captured before (plus exactly the intended changes on success).
- Selectors are anchored `^…$`. A run whose top-level `=== RUN` count is below the stated number is not a PASS.
- **Numeric-assertion rule.** Every assertion that a count or a line number appears in an error, or does not, is checked on the error string with every fixture path removed: the test first asserts the paths the AC names, then deletes from the error every path under a `t.TempDir()` it created, in its cleaned absolute form and its symbolic-link-resolved form, and only then looks for numbers. A number is matched as a whole number — not preceded or followed by another digit (for example `(^|[^0-9])2([^0-9]|$)`) — or read from a structured field where the error exposes one. The fixture keeps every other digit run left in that string (an entry `id`, clause text) different from each number the AC asserts present or absent. Without this rule a temporary-directory path such as `…/001/…` already carries the digits a count assertion looks for, and a mutant that prints the wrong number passes.
- ACs marked **baseline-first** are run against unchanged production code and their RED output is committed before the production change (plan.md §F).
- Every RED statement in this file, in the matrix and in §D.1, is a prediction from code reading at `54ca2e3b6`; production code is unchanged since `5a066994b` (a diff of `internal/constitution` and `internal/cli/constitution.go` between those two commits prints nothing). Where a verdict observation is cited, it covers the underlying behaviour, not the proposed test. A RED becomes an observation only when the run-phase baseline commit records the failing output.

## §D AC Matrix

| AC | Requirement | RED predicted from code reading (see common rules; first read at `ff11e752f`) | Mutant(s) |
|---|---|---|---|
| AC-CAA-001 | REQ-CAA-001, REQ-CAA-014 | stub error (verdict §2.1) — baseline-first | M-1 |
| AC-CAA-002 | REQ-CAA-002, REQ-CAA-014 | stub error, not an occurrence error — baseline-first | M-1 |
| AC-CAA-003 | REQ-CAA-001, REQ-CAA-002 | stub error — baseline-first | M-2 |
| AC-CAA-004 | REQ-CAA-003 | stub error — baseline-first | M-3a, M-3b |
| AC-CAA-005 | REQ-CAA-004 | stub error — baseline-first | M-4 |
| AC-CAA-006 | REQ-CAA-005 | writer emits `ruleid:` / `zonebefore: 0` (verdict §2.2) — baseline-first | M-9 |
| AC-CAA-007 | REQ-CAA-006 | legacy-only assertions: none expected — the untagged decoder reads legacy keys today, so their RED cell is the mutant; both-forms assertion (`rule_id: A` with `ruleid: B`): RED — the untagged decoder ignores `rule_id` and reads `RuleID = B` — baseline-first | M-6 |
| AC-CAA-008 | REQ-CAA-006, REQ-CAA-014 | `RuleID` empty on the snake_case fixture (verdict §2.2) — baseline-first | M-7 |
| AC-CAA-009 | REQ-CAA-007 | `---` split fabricates or drops entries (verdict §2.3) — baseline-first | M-12 |
| AC-CAA-010 | REQ-CAA-008, REQ-CAA-009 | 0 entries from the real log (verdict §2.3) — baseline-first | M-8a, M-8b |
| AC-CAA-011 | REQ-CAA-008 | limiter admits (sees 0 entries) — baseline-first | M-8a |
| AC-CAA-012 | REQ-CAA-010, REQ-CAA-011 | seams do not exist (compile failure, not discriminating) — the discriminating RED is the mutants | M-5a, M-5b, M-11a |
| AC-CAA-013 | REQ-CAA-010, REQ-CAA-011 | seam does not exist — the mutant is the RED | M-5c |
| AC-CAA-014 | REQ-CAA-012, REQ-CAA-014 | dry-run returns success on a two-occurrence fixture — baseline-first | M-10, M-11b |
| AC-CAA-015 | REQ-CAA-013 | CLI dry-run prints success on a two-occurrence fixture — baseline-first (needs compile slot) | M-13 |
| AC-CAA-016 | REQ-CAA-014 | `not yet implemented` present in pipeline.go and pipeline_test.go | — (grep with control) |
| AC-CAA-017 | REQ-CAA-015 | none expected — invariant guard | M-14 |
| AC-CAA-018 | REQ-CAA-009 | reader ignores the malformed block and returns no error — baseline-first | M-15 |
| AC-CAA-019 | REQ-CAA-016 | stub error, not a new-clause occurrence error — baseline-first | M-16 |
| AC-CAA-020 | REQ-CAA-017 | dry-run `Execute` with a stale `Before` succeeds — baseline-first | M-17 |
| AC-CAA-021 | REQ-CAA-011, REQ-CAA-018 | seams do not exist — the discriminating RED is the mutant | M-18 |
| AC-CAA-022 | REQ-CAA-019 | `Execute` joins `projectDir`, fails to load the registry at the env path — baseline-first | M-19 |
| AC-CAA-023 | REQ-CAA-020 | `Execute` ignores `CLAUDE_PROJECT_DIR`, loads the registry inside `projectDir`, and returns dry-run success or the stub error instead of a registry load error (code unchanged at `92c8c3f36`) — baseline-first | M-19 (not M-20 (i): under D4 the loader's retained absolute-only refusal also stops these rows — see the AC note) |
| AC-CAA-024 | REQ-CAA-020, REQ-CAA-021 | escape rows: `Execute` ignores both environment variables and joins `file:` with no check, so each returns dry-run success or the stub error; the CLI case returns `clause mismatch`, an error without the offending path; `in_root_control` real cases return the stub error (code unchanged at `578afca87`) — baseline-first; `loader_unchanged`: none expected — it states today's loader behaviour; preservation guard whose RED cells are M-20 (iii) and, for case `absolute_escape`, M-20 (iv) | M-20, M-21, M-22, M-23, M-24 |
| AC-CAA-025 | REQ-CAA-021 | `load`: none expected — a direct `LoadRegistry` call, unchanged under D4; no declared mutant turns it RED; `dry_run`: none expected — invariant guard whose RED cell is M-24; `internal/spec` preservation run: none expected — the existing test passes today, preservation guard whose RED cell is M-20 (iii). `real`: stub error — baseline-first | M-20, M-24 |

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
- AC-CAA-018 maps REQ-CAA-009
- AC-CAA-019 maps REQ-CAA-016
- AC-CAA-020 maps REQ-CAA-017
- AC-CAA-021 maps REQ-CAA-011, REQ-CAA-018
- AC-CAA-022 maps REQ-CAA-019
- AC-CAA-023 maps REQ-CAA-020
- AC-CAA-024 maps REQ-CAA-020, REQ-CAA-021
- AC-CAA-025 maps REQ-CAA-021

The union is REQ-CAA-001 … REQ-CAA-021; no requirement lacks an AC.

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
- **Then** each returns an error whose text names the rule file path and the count — `0` for (a), `2` for (b) — the count matched under the numeric-assertion rule of the common rules,
- **And** the rule file, registry, and log are byte-identical to their pre-apply snapshots and no path was added under the fixture dir,
- **And** the lock file is released.

Command: `go test ./internal/constitution/ -run '^TestApply_OccurrenceCount_Rejected$' -count=1 -v` — expect 1 top-level RUN with subtests `zero` and `two` both reported.

### AC-CAA-003 — no whitespace normalization

- **Given** a rule file where the clause text appears once with a doubled internal space, and once more elsewhere wrapped across a newline, while the registry clause uses single spaces on one line,
- **When** `Execute(dryRun=false)` runs with an approving oversight double,
- **Then** it returns an error naming the rule file path; under the numeric-assertion rule of the common rules, the exact-match count `0` is present and the whitespace-normalized count `2` is absent; and all three files are byte-identical to their pre-apply snapshots.

Command: `go test ./internal/constitution/ -run '^TestApply_NoWhitespaceNormalization$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-004 — registry rewrite touches one line and round-trips

- **Given** a registry fence with header comments, blank lines between entries, at least three entries, and a new clause containing `"`, `\`, `: `, and `#`,
- **When** the registry update runs,
- **Then** a line-by-line comparison with the pre-apply registry differs in exactly one line — the target entry's `clause:` line — and the line count is unchanged,
- **And** `LoadRegistry` decodes the target clause as exactly the new clause.

Command: `go test ./internal/constitution/ -run '^TestUpdateRegistryClause_SingleLineRoundTrip$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-005 — re-parse verification blocks a corrupt registry

- **Given** three fixtures: (a) the target entry's clause is a double-quoted scalar continued onto a second line, so rewriting only the `clause:` line leaves an orphaned continuation; (b) the target entry is a one-line yaml flow mapping, `- {id: <RuleID>, zone: Evolvable, file: <rule file>, clause: "<current clause>"}` plus any further field the fixture's other entries carry, which the loader decodes with the current clause while no `- id:` line and no `clause:` line of its own exist to rewrite — its rule file holds the current clause once and the new clause not at all, so every check before the registry rewrite passes; (c) the new clause contains a newline,
- **When** `Execute(dryRun=false)` runs on each,
- **Then** each returns an error whose text contains the registry path (REQ-CAA-004), and all three files are byte-identical to their pre-apply snapshots.

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

Baseline-first (predicted from code reading at `54ca2e3b6`): with production code unchanged, the both-forms assertion must fail — `AmendmentLog` carries no yaml tags, so the decoder ignores `rule_id` and reads `RuleID = B`. The legacy-only assertions pass today and are guarded by M-6.

### AC-CAA-008 — TestLoadEvolutionLogs asserts RuleID and ApprovedAt

- **Given** the existing snake_case fixtures in `TestLoadEvolutionLogs`,
- **When** the test runs,
- **Then** it asserts `ID`, `RuleID = CONST-V3R2-008`, and `ApprovedAt = 2026-04-28T10:00:00Z` for the single-entry case, and `RuleID` for both entries of the multi-entry case.

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs$' -count=1 -v` — expect 1 top-level RUN.

Baseline-first (predicted from code reading at `54ca2e3b6`; verdict §2.2 observed the underlying `ruleID=""` read, not this test): with only the assertions added and production code unchanged, this test must FAIL on `RuleID`. A PASS at that point means the assertion is not reaching the field.

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
- **And given** (c) a human-format block with `id: EVO-X-001`, `timestamp: "not-a-date"`, and no `approved_at`, **then** `LoadEvolutionLogs` returns an error whose text contains `EVO-X-001`. (The file, line, and key content of that error is AC-CAA-018.)

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs_HumanFormat$' -count=1 -v` — expect 1 top-level RUN with subtests `verbatim_block`, `real_file_readonly`, `malformed_timestamp_fails_closed`.

### AC-CAA-011 — the rate limiter now sees the human entry

- **Given** a temp log containing only the verbatim EVO-HRN-002 block and a `rateLimiter` whose `now` is `2026-05-13T01:00:00Z`,
- **When** `Admit` runs for any proposal,
- **Then** it returns `*ErrRateLimitExceeded` (24-hour cooldown from the entry's `ApprovedAt`),
- **And** with `now = 2026-06-13T00:00:00Z` on the same log it returns nil — so the rejection is caused by the entry, not by an always-reject limiter.

Command: `go test ./internal/constitution/ -run '^TestRateLimiter_SeesHumanFormatEntry$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-012 — a failed rename restores all three files

- **Given** a valid fixture, the default restore seam, and a rename seam that, on call N, fails in one of two modes and otherwise delegates to `os.Rename`: **fail-before** returns an error without renaming; **rename-then-fail** renames through `os.Rename`, records whether the destination path exists afterwards, and then returns an error,
- **When** `Execute(dryRun=false)` runs in each subtest below, on a fresh fixture,
- **Then** it returns an error naming the failed rename,
- **And** the rename seam was called exactly N times (reachability — without this, an injector that never fires is indistinguishable from a correct restore), and in rename-then-fail mode the recorded destination existed after the delegated rename, so the failing step had taken effect before the restore ran,
- **And** the rule file, registry, and log are byte-identical to their pre-apply sha256 snapshots, and the path set under the fixture dir equals its pre-apply snapshot, so no temporary or backup path remains:

  | Subtest | Mode | N | Log before apply | What only a correct restore passes |
  |---|---|---|---|---|
  | `first_rename_applied` | rename-then-fail | 1 | exists | the rule file, already replaced, is rewritten from its backup |
  | `second_rename` | fail-before | 2 | exists | the rule file, renamed by call 1, is restored |
  | `third_rename` | fail-before | 3 | exists | the rule file and the registry, renamed by calls 1 and 2, are restored |
  | `third_rename_applied` | rename-then-fail | 3 | exists | the log, already replaced, is rewritten from its backup |
  | `third_rename_log_absent` | rename-then-fail | 3 | absent | the log the failing call created is removed, so the log path does not exist |

- **And given** no injected failure (`no_fault_clean`), **then** the apply succeeds and no temporary or backup path remains.

Command: `go test ./internal/constitution/ -run '^TestApply_RenameFault_RestoresAll$' -count=1 -v` — expect 1 top-level RUN with subtests `first_rename_applied`, `second_rename`, `third_rename`, `third_rename_applied`, `third_rename_log_absent`, `no_fault_clean`.

Kill map: M-5a (no restore) fails every subtest except `no_fault_clean`. M-5b variant (i) (restore only the files whose rename reported success) fails `first_rename_applied`, `third_rename_applied`, and `third_rename_log_absent`; it passes `second_rename` and `third_rename`, where restoring only renamed files is the correct outcome. M-5b variant (ii) (skip removing a file recorded absent) fails `third_rename_log_absent`. M-11a fails `no_fault_clean`. M-5c is killed by AC-CAA-013. A fail-before injector alone could not kill either variant of M-5b: the failing rename never takes effect, so there is nothing a partial restore leaves behind.

### AC-CAA-013 — rename order is source, registry, log

- **Given** a recording rename seam that delegates to `os.Rename`,
- **When** a successful apply runs,
- **Then** the recorded destinations are, in order, the rule file path, the registry path, and the log path, and there are exactly three.

Command: `go test ./internal/constitution/ -run '^TestApply_RenameOrder$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-014 — dry-run runs the validation and writes nothing

- **Given** six fixtures: (a) valid; (b) current clause occurs twice; (c) registry continuation line (AC-CAA-005 a); (d) rule file missing; (e) new clause already present in the rule file; (f) proposal `Before` differs from the current clause,
- **When** `Execute(dryRun=true)` runs on each,
- **Then** (a) returns a log entry and no error; (b)–(f) each return an error containing the substrings below, and `Execute(dryRun=false)` on a fresh copy of the same fixture returns an error containing the same substrings: (b) the rule file path and `2`; (c) the registry path; (d) the rule file path; (e) the rule file path and `1`; (f) the rule ID — the numbers of (b) and (e) matched under the numeric-assertion rule of the common rules,
- **And** for all six, the snapshot of every path and sha256 under the fixture dir is identical before and after, and no lock file was created.

Command: `go test ./internal/constitution/ -run '^TestPipeline_Execute_DryRun_Validates$' -count=1 -v` — expect 1 top-level RUN with subtests `valid`, `two_occurrences`, `registry_continuation`, `missing_rule_file`, `new_clause_present`, `stale_before`. The former `TestPipeline_Execute_DryRun_Success` is folded into subtest `valid` or kept with an existing rule file (plan.md §C.2).

### AC-CAA-015 — CLI dry-run surfaces the validation failure

- **Given** an `internal/cli` test that sets `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` to empty with `t.Setenv`, and a `t.TempDir()` project fixture whose rule file contains the current clause twice,
- **When** it calls `runConstitutionAmend(stdout, stderr, projectDir, ruleID, before, after, "", true)`,
- **Then** the call returns a non-nil error and stdout does not contain `Dry-run success`,
- **And** on the valid-fixture control the call returns nil, stdout contains `Dry-run success`, and the fixture snapshot is unchanged.

Command (compile slot required at run-phase): `unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_DRY_RUN && go test ./internal/cli/ -run '^TestConstitutionAmend_DryRun_SurfacesValidation$' -count=1 -v -timeout 600s` — expect 1 top-level RUN with subtests `two_occurrences` and `valid`.

This AC is the whole CLI-level coverage; the non-dry-run CLI path is the approved Gap in spec.md §E.4.

### AC-CAA-016 — no stub characterization remains

- **Given** the tree after M5,
- **When** `/usr/bin/grep -rn 'not yet implemented' internal/constitution/pipeline.go internal/constitution/pipeline_test.go` runs,
- **Then** it prints 0 lines,
- **And** the control `/usr/bin/grep -c 'func (p \*Pipeline) Execute' internal/constitution/pipeline.go` prints `1` (the grep reached the file),
- **And** `/usr/bin/grep -nE 'func (TestPipeline_Execute_NonDryRun_AmendmentStubError|TestPipeline_applyAmendment_StubError|TestUpdateSourceFile_StubError|TestUpdateRegistryClause_StubError)\(' internal/constitution/pipeline_test.go` prints 0 lines, and each replacement named in plan.md §C.2 exists (`/usr/bin/grep -c` per name ≥ 1).

### AC-CAA-017 — real files untouched by the package runs, even with a session environment

- **Given** sha256 of `.claude/rules/moai/core/zone-registry.md` and `.moai/research/evolution-log.md` recorded before the runs,
- **When** `go test ./internal/constitution/ -count=1` runs twice — once as `unset MOAI_CONSTITUTION_REGISTRY CLAUDE_PROJECT_DIR && go test ./internal/constitution/ -count=1`, and once with `CLAUDE_PROJECT_DIR` exported as the repository root (`CLAUDE_PROJECT_DIR="$(git rev-parse --show-toplevel)" go test ./internal/constitution/ -count=1`) — and the AC-CAA-015 command completes,
- **Then** both sha256 values are unchanged after every run, `git status --porcelain -- .claude/rules .moai/research` shows nothing attributable to the runs, and neither `internal/constitution/.moai` nor `internal/cli/.moai` exists,
- **And** both package runs report the same PASS/FAIL result, which shows the tests set the variables themselves rather than relying on the shell.

### AC-CAA-018 — the fail-closed error names file, line, and key

- **Given** a temp log whose first 5 lines are prose, followed by a fenced yaml block for `id: EVO-X-001` in which `timestamp: "not-a-date"` sits on a known file line L1 (and `approved_at` is absent), and a second log whose first 3 lines are prose, followed by a machine `---` entry `id: LEARN-20260911-009` with no timestamp key at all, its `id:` key on known file line L2,
- **When** `LoadEvolutionLogs` reads each file,
- **Then** the first error contains the log file path, the key `timestamp`, and `EVO-X-001`, and — under the numeric-assertion rule of the common rules — the decimal line number L1 is present and the block-relative line number of `timestamp` (its line counted from the first line inside the yaml fence) is absent,
- **And** the second error contains its file path, the key `approved_at`, and `LEARN-20260911-009`, and under the same rule L2 is present and the segment-relative line number of `id:` (counted from the first line after its opening `---`) is absent,
- **And** the fixtures fix L1, L2, and both relative numbers in advance, no two equal and none occurring as a whole number elsewhere in the path-stripped error (the digits of `EVO-X-001` and `LEARN-20260911-009` included); the prose offsets make each file line differ from its relative line, so an error that reports the relative line fails.

Command: `go test ./internal/constitution/ -run '^TestLoadEvolutionLogs_FailClosedErrorLocation$' -count=1 -v` — expect 1 top-level RUN with subtests `unparseable_timestamp` and `missing_timestamp`.

### AC-CAA-019 — a new clause already present in the source is rejected

- **Given** two fixtures: (a) the rule file contains the current clause once and, elsewhere, the new clause once; (b) the new clause is a prefix of the current clause (so it occurs once, inside the current clause),
- **When** `Execute(dryRun=false)` runs on each,
- **Then** each returns an error whose text names the rule file path and the new-clause occurrence count `1`, matched under the numeric-assertion rule of the common rules,
- **And** the rule file, registry, and log are byte-identical to their pre-apply snapshots, no path was added under the fixture dir, and the lock is released.

Command: `go test ./internal/constitution/ -run '^TestApply_NewClausePresent_Rejected$' -count=1 -v` — expect 1 top-level RUN with subtests `elsewhere` and `inside_current_clause`.

### AC-CAA-020 — Execute rejects a stale Before

- **Given** a valid fixture and a proposal whose `Before` differs from the current clause by one trailing character, with oversight, canary, and contradiction doubles that record whether they were called,
- **When** `Execute` runs in dry-run and, separately, in real mode, called directly (not through the CLI),
- **Then** both return an error containing the rule ID,
- **And** no gate double was called (the check runs before Layer 1),
- **And** all three files are byte-identical and the lock is released in real mode.

Command: `go test ./internal/constitution/ -run '^TestPipeline_Execute_StaleBefore_Rejected$' -count=1 -v` — expect 1 top-level RUN with subtests `dry_run` and `real`.

Baseline-first (predicted from code reading at `54ca2e3b6`): against current code, the `dry_run` subtest must FAIL (dry-run `Execute` returns success today). The mutant M-17 keeps the CLI `--before` check and removes only the `Execute` check; the AC stays RED because it calls `Execute` directly.

### AC-CAA-021 — a failed restore keeps the backups and names them

- **Given** a valid fixture whose three files all exist, a rename seam that fails on call 2, and a restore seam that fails on its first call and otherwise delegates to the default restore,
- **When** `Execute(dryRun=false)` runs,
- **Then** it returns an error that contains the path of every backup file the apply wrote (three paths) and names the failed restore step,
- **And** every path named in the error exists and holds the pre-apply bytes of its file (sha256 equal to the pre-apply snapshot),
- **And** the rename seam was called exactly 2 times and the restore seam at least 1 time (reachability),
- **And** no temporary file remains under the fixture dir.

Command: `go test ./internal/constitution/ -run '^TestApply_RestoreFault_KeepsBackups$' -count=1 -v` — expect 1 top-level RUN.

### AC-CAA-022 — the CLI and Execute resolve the same registry

- **Given** a `t.TempDir()` project `P` holding two registries: the default location `P/.claude/rules/moai/core/zone-registry.md` with the target clause `A`, and `P/alt/zone-registry.md` with the target clause `B` (each pointing at its own rule file containing its clause once); a second temp dir `Q` holding a third registry at the default location; `MOAI_CONSTITUTION_REGISTRY = P/alt/zone-registry.md` and `CLAUDE_PROJECT_DIR = Q` set with `t.Setenv`,
- **When** the shared resolver is called for `P`, and `Execute(dryRun=false)` runs for `P` with `Before = B`,
- **Then** the resolver returns `P/alt/zone-registry.md` (the env override outranks `CLAUDE_PROJECT_DIR`),
- **And** `Execute` succeeds and `LoadRegistry(P/alt/zone-registry.md)` shows the new clause,
- **And** the default-location registry in `P` and the registry in `Q` are byte-identical to their pre-apply snapshots,
- **And** in `internal/cli`, `resolveRegistryPath(P)` under the same environment returns the same path as the shared resolver.

Command: `go test ./internal/constitution/ -run '^TestExecute_UsesSharedRegistryResolver$' -count=1 -v` (expect 1 top-level RUN) and, with the compile slot, `go test ./internal/cli/ -run '^TestResolveRegistryPath_MatchesExecute$' -count=1 -v -timeout 600s` (expect 1 top-level RUN).

Baseline-first (predicted from code reading at `54ca2e3b6`): against current code, `Execute` loads the default location in `P`, finds clause `A`, and — once REQ-CAA-017 lands — rejects `Before = B`; today it passes the gates and fails at the stub. Either way the success assertion is RED. The mutant M-19 keeps `Execute`'s own `projectDir` join and turns the AC RED the same way.

Relation to AC-CAA-023: this AC fixes *which* path the one resolver chooses; AC-CAA-023 fixes *whether* the chosen path is admitted (REQ-CAA-020). They do not contradict each other. Here `MOAI_CONSTITUTION_REGISTRY` names a path inside `P`, outranks `CLAUDE_PROJECT_DIR = Q`, and the amend path admits it — the containment check and the loader alike. In AC-CAA-023 `MOAI_CONSTITUTION_REGISTRY` is empty, so `CLAUDE_PROJECT_DIR = Q` supplies a path inside `Q`, and the amend path refuses it — the containment check and, because that path is absolute, the loader's retained refusal as well (D4). In both, nothing under `Q` is written.

### AC-CAA-023 — a registry in another tree stops Execute before any write

- **Given** two sibling `t.TempDir()` trees `P` and `Q` (neither inside the other) and a third `t.TempDir()` `L` holding the lock path; `P` is a valid fixture (registry at the default location, a rule file whose relative `file:` path resolves inside `P` and contains the current clause once and the new clause not at all, and an existing evolution log); `Q` holds a byte copy of `P`'s registry at its default location `Q/.claude/rules/moai/core/zone-registry.md`; the proposal's `Before` equals the current clause; `MOAI_CONSTITUTION_REGISTRY` is empty and `CLAUDE_PROJECT_DIR = Q`, both set with `t.Setenv`; oversight, canary, and contradiction doubles record whether they were called; the path set and sha256 of every file under `P` and `Q` are captured before the call,
- **When** `Execute` runs with `projectDir = P` in real mode and, on a fresh copy of the same fixture, in dry-run mode,
- **Then** each returns the registry load error, whose text contains the offending registry path `Q/.claude/rules/moai/core/zone-registry.md` in its cleaned absolute form or its symbolic-link-resolved form (the test computes both; the wording around the path is not asserted, because under D4 two layers refuse this absolute path on the amend path — the containment check of REQ-CAA-021 and `LoadRegistry`'s retained absolute-only refusal — and which fires first is a run-phase choice),
- **And** no gate double was called,
- **And** the path set and every sha256 under `P` and under `Q` equal their pre-call snapshots,
- **And** in real mode no lock file remains under `L` (the lock was released);
- **And given** the control `CLAUDE_PROJECT_DIR = P` on a fresh copy of the same fixture, **when** `Execute` runs with `projectDir = P` in real mode, **then** it returns a log entry and no error, `P`'s rule file, registry, and log carry the amendment, and every sha256 under `Q` is unchanged — so the refusal comes from the tree mismatch, not from a check that refuses everything.

Command: `go test ./internal/constitution/ -run '^TestExecute_RegistryOutsideProjectDir_Refused$' -count=1 -v` — expect 1 top-level RUN with subtests `divergent_root_real`, `divergent_root_dry_run`, `same_root_control`.

Why `Q` holds a copy of `P`'s registry: if nothing refused `Q`'s registry — neither the containment check nor the loader's absolute-path refusal — `Execute` would find the same entry with a clause equal to `Before` and resolve its `file:` against `P`, so the apply would succeed and write `P`'s rule file, `Q`'s registry, and `P`'s log — a cross-tree write. The fixture makes a mutant that removes both layers fail the error, byte-identity, and path-set assertions together, not only the error text.

Which mutant kills these rows under D4: M-19 (predicted from code reading). With `Execute` keeping its own `projectDir` join, it never looks at `Q`, loads `P`'s registry, and returns dry-run success or, in real mode, applies to `P` — RED on the error assertion, and in real mode on `P`'s byte-identity. M-20 variant (i) is not claimed here: it removes only the check's registry-path call, and the loader's retained refusal still stops this absolute path, so the rows stay GREEN under it. The check's registry-path site is killed through AC-CAA-024 instead (`relative_env_escape`, `symlinked_registry`, CLI case).

Baseline-first (predicted from code reading at `54ca2e3b6`): against current code, `Execute` ignores `CLAUDE_PROJECT_DIR` and loads `P`'s registry; `divergent_root_dry_run` returns success and `divergent_root_real` returns the stub error, so both are RED on the error assertion. `same_root_control` stays RED on the stub until M5.

### AC-CAA-024 — the containment check refuses every escaping path shape before any write

- **Given** a base `B` equal to `filepath.EvalSymlinks(t.TempDir())` — resolved, so no directory on the way to `B` is itself a symbolic link (on darwin `t.TempDir()` lies under `/var`, a link to `/private/var`) and the only symbolic links in the fixture are the ones a row creates on purpose, `P/linkdir`, `P/linked`, `P/.moai/research`, and `B/link` — holding the project root `P = B/root`, a sibling `B/root-evil`, and outside trees `B/other`, `B/outside`, and `B/outside-research`; a lock path pinned under a separate `t.TempDir()` `L`; `P` a valid fixture (registry at the default location, a target Evolvable entry whose rule file holds the current clause once and the new clause not at all, and an existing evolution log) except for the one escaping shape its row names; wherever a row's escaping path names a rule file, that file also holds the current clause once and the new clause not at all, so the containment check is the only thing that can stop the apply; `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` empty unless the row says otherwise, set with `t.Setenv`; oversight, canary, and contradiction doubles that record whether they were called; the path set and the sha256 of every file under `B` captured before the call; and these rows, each on a fresh fixture:

  | Row (subtest) | Escaping shape | Offending path the error names |
  |---|---|---|
  | `relative_env_escape` | working directory set to `P` with `t.Chdir`; `B/other` holds a byte copy of `P`'s registry at its default location. Case `registry_var`: `MOAI_CONSTITUTION_REGISTRY = ../other/.claude/rules/moai/core/zone-registry.md`, `CLAUDE_PROJECT_DIR` empty. Case `project_dir_var`: `MOAI_CONSTITUTION_REGISTRY` empty, `CLAUDE_PROJECT_DIR = ../other` | `B/other/.claude/rules/moai/core/zone-registry.md` |
  | `absolute_file` | the target entry's `file:` is the absolute path `B/outside/rule.md`. Case `non_target`: the target entry's `file:` stays relative and inside `P`, and a different entry carries that absolute `file:` | `B/outside/rule.md` |
  | `dotdot_file` | the target entry's `file:` is the relative value `../outside/rule.md` | `B/outside/rule.md` |
  | `sibling_prefix_file` | the target entry's `file:` is the absolute path `B/root-evil/rule.md`, a sibling whose name begins with the root's name | `B/root-evil/rule.md` |
  | `symlinked_registry` | `P/linkdir` is a symbolic link to `B/other`, which holds a byte copy of `P`'s registry as `B/other/zone-registry.md`; `MOAI_CONSTITUTION_REGISTRY = P/linkdir/zone-registry.md` (absolute), `CLAUDE_PROJECT_DIR` empty | `P/linkdir/zone-registry.md` |
  | `symlinked_file` | `P/linked` is a symbolic link to `B/outside`; the target entry's `file:` is the relative value `linked/rule.md`, so the file it reaches is `B/outside/rule.md` | `P/linked/rule.md` |
  | `symlinked_log` | `P/.moai/research` is a symbolic link to `B/outside-research`, which holds the existing log | `P/.moai/research/evolution-log.md` |

- **When** `Execute` runs with `projectDir = P` in real mode and, on a fresh copy of the same fixture, in dry-run mode,
- **Then** each returns an error whose text contains the row's offending path in its cleaned absolute form or its symbolic-link-resolved form (the test computes both),
- **And** no gate double was called,
- **And** the path set and every sha256 under `B` equal their pre-call snapshots — inside `P` and in every tree outside it,
- **And** in real mode no lock file remains under `L`;
- **And given** the control row `in_root_control` on a fresh fixture with every path inside the root — case `plain`: `projectDir = P`, working directory `P`, `MOAI_CONSTITUTION_REGISTRY = .claude/rules/moai/core/zone-registry.md` (a relative value inside the root), and `P/.moai/research/` present with no log file yet; case `symlinked_root`: the same, except that `P` is reached through a symbolic link `B/link → B/root`, so `projectDir = B/link` and the working directory is `B/link` — **when** `Execute` runs in real mode, **then** it returns a log entry and no error, the target rule file and the registry carry the amendment, the log now exists and holds exactly the one new entry, and every sha256 under `B` outside `P` is unchanged; **and when** `Execute` runs in dry-run mode on a fresh copy, **then** it returns a log entry, no error, and an unchanged snapshot of `B` — so the refusals come from the escaping shapes, not from a check that refuses relative values, a root reached through a link, or a log not yet created;
- **And given** the preservation case `loader_unchanged` (operator decision D4: `LoadRegistry` performs no containment check and keeps its present behaviour), each on a fresh fixture, with no call to `Execute` — case `relative_registry`: built as `relative_env_escape` (working directory `P`, `B/other` holding a byte copy of `P`'s registry), `LoadRegistry("../other/.claude/rules/moai/core/zone-registry.md", P)`; case `absolute_file`: built as the `absolute_file` row, `LoadRegistry` on `P`'s registry with `projectDir = P`; case `symlinked_registry`: built as the `symlinked_registry` row, `LoadRegistry("P/linkdir/zone-registry.md", P)`; case `absolute_escape`: `LoadRegistry("B/other/.claude/rules/moai/core/zone-registry.md", P)` on the `relative_registry` fixture — **when** each call runs, **then** `relative_registry`, `absolute_file`, and `symlinked_registry` each return no error and as many entries as the test counts `- id:` lines inside that registry's yaml fence, with the `absolute_file` case's target entry carrying `File` equal to `B/outside/rule.md` exactly as written; `absolute_escape` returns the loader's present error, which contains `escapes project dir` (`internal/constitution/loader.go:86`); and the snapshot of `B` is unchanged — so the loader admits and refuses exactly what it does before this SPEC, and the refusals of the rows above come from the amend path's check, not from the loader;
- **And given** the CLI case in `internal/cli`, on a fixture built as `relative_env_escape` case `project_dir_var` except that the registry copy under `B/other` gives the target entry a different clause, **when** `runConstitutionAmend(stdout, stderr, P, ruleID, before, after, "", true)` runs with `before` equal to `P`'s current clause, **then** it returns a non-nil error containing the offending path `B/other/.claude/rules/moai/core/zone-registry.md` in its cleaned absolute form or its symbolic-link-resolved form (the test computes both; `B` is the resolved base of the Given above), the error text contains neither `amendment failed` (the CLI's wrapper around a pipeline error, `internal/cli/constitution.go:544`) nor `clause mismatch` (the CLI's own `--before` error, `:529`), stdout does not contain `Dry-run success`, and the snapshot of `B` is unchanged — so the refusal came from the CLI's own registry validation using the same check (REQ-CAA-021), before the pipeline was reached. Why the copy differs: with a byte copy, a CLI that read `B/other` without the check would pass its `--before` comparison, call `Execute`, and return `Execute`'s refusal — the same path, output, and snapshot as a CLI that ran the check. With a different clause, that CLI stops at `clause mismatch`, an error naming no path, and the case fails.

Commands: `go test ./internal/constitution/ -run '^TestExecute_ContainmentCheck_RefusesEscapes$' -count=1 -v` — expect 1 top-level RUN with subtests `relative_env_escape`, `absolute_file`, `dotdot_file`, `sibling_prefix_file`, `symlinked_registry`, `symlinked_file`, `symlinked_log`, `in_root_control`, and `loader_unchanged`, each escape row and `in_root_control` reporting its `real` and `dry_run` cases, `loader_unchanged` reporting its four cases, and the nested cases named above; and, with the compile slot, `go test ./internal/cli/ -run '^TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape$' -count=1 -v -timeout 600s` — expect 1 top-level RUN.

Symbolic links: where the platform refuses to create a symbolic link, `symlinked_registry`, `symlinked_file`, `symlinked_log`, `in_root_control/symlinked_root`, and `loader_unchanged/symlinked_registry` call `t.Skip` with the error. A skip is recorded in `progress.md` §E.2 as a Gap, never as a PASS, and every variant of M-23, and M-24, are then unobserved on that platform.

Baseline-first (predicted from code reading at `54ca2e3b6`; see the common rules for the unchanged-code check): against current code, `Execute` ignores both environment variables and joins `file:` with no check, so every escape row returns dry-run success or, in real mode, the stub error — RED on the error assertion. `in_root_control` real cases stay RED on the stub until M5. `loader_unchanged` is predicted GREEN against current code and must stay GREEN: it states the loader's present behaviour (`loader.go:80-88` checks an absolute registry path only and never checks `file:`), so it is a preservation guard whose RED cells are M-20 variant (iii) and, for case `absolute_escape`, variant (iv). The CLI case is predicted RED as well: the CLI resolver returns the relative path, `LoadRegistry` checks containment only for an absolute path (`internal/constitution/loader.go:82`) and reads `B/other`'s registry, and the CLI returns `clause mismatch` (`internal/cli/constitution.go:529`), an error without the offending path.

Mutant coverage: M-20 variant (i) turns `relative_env_escape` (both cases), `symlinked_registry`, and the CLI case RED, variant (ii) turns the CLI case RED, and variant (iii) turns `loader_unchanged` cases `relative_registry`, `absolute_file`, and `symlinked_registry` RED (the check inside the loader refuses all three; `absolute_escape` is refused either way under (iii)), and variant (iv) turns `loader_unchanged` case `absolute_escape` RED (the opposite direction — a loader that dropped its present refusal reads `B/other`'s registry and returns no error); M-21 at the `file:` site turns `absolute_file` (both cases), `dotdot_file`, `sibling_prefix_file`, and `symlinked_file` RED, and at the log site turns `symlinked_log` RED; M-22 variant (i) turns `sibling_prefix_file` RED and variant (ii) turns `dotdot_file` RED; M-23 variant (i) turns `symlinked_registry` RED, variant (ii) `symlinked_file`, and variant (iii) `symlinked_log`; M-24 turns `in_root_control/symlinked_root` RED.

### AC-CAA-025 — the real registry's `file:` shape is still admitted

- **Given** a base `B` equal to `filepath.EvalSymlinks(t.TempDir())` (resolved as in AC-CAA-024, so `B/link` is the fixture's only symbolic link) with the project root `P = B/root` reached through a symbolic link `B/link → B/root`; the repository's real `.claude/rules/moai/core/zone-registry.md` opened read-only and copied byte for byte to `P/.claude/rules/moai/core/zone-registry.md`; a drift witness on that copy asserting its shape — 0 `file:` values that are absolute and 0 that contain `..` (the entry-count pin is not repeated here: it lives in `wantRegistryEntries`, `internal/constitution/registry_sync_test.go:49`, whose own rule updates it with any deliberate registry change, and whose package `constitution_test` this package-internal test cannot import); every distinct `file:` path of the copy created under `P`; a target entry chosen by rule rather than by ID — the first entry in registry order whose zone is `Evolvable` and whose clause does not begin with `[SUPERSEDED` — whose file holds its current clause once and the new clause not at all, and the other files holding placeholder text that contains neither clause of the proposal; `P/.moai/research/` present with no log file yet; `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` empty, set with `t.Setenv`; a lock path under a separate `t.TempDir()` `L`; the path set and sha256 of every file under `B`, and the sha256 of the real registry and the real log, captured before the calls,
- **When** `LoadRegistry("B/link/.claude/rules/moai/core/zone-registry.md", "B/link")` runs — the registry path given through the link, not as `B/root/…` — then `Execute` runs with `projectDir = B/link` in dry-run mode, and then, on a fresh copy of the fixture, in real mode,
- **Then** `LoadRegistry` returns no error and as many entries as the test counts `- id:` lines inside the copy's yaml fence with a line scan,
- **And** dry-run returns a log entry, no error, and an unchanged snapshot of `B`,
- **And** real mode returns a log entry and no error, the target rule file, the registry copy, and the new log carry the amendment, and every other file under `B` is byte-identical,
- **And** the real registry and the real log keep the sha256 captured before the calls (REQ-CAA-015);
- **And given** the preservation run for the non-amend callers of `LoadRegistry` (operator decision D4) — the existing `internal/spec` test `TestLinter_AC08_DanglingRuleReference`, which pairs the relative registry path `../../.claude/rules/moai/core/zone-registry.md` with `BaseDir: "testdata"` (`internal/spec/lint_test.go:18-22`, `:218-222`) and reads the real registry read-only — **when** it runs on the run-phase tree, **then** it passes, and this card's commits did not touch its source — `git log --first-parent --no-merges --format=%H $BASELINE_SHA..HEAD -- internal/spec/lint.go internal/spec/lint_test.go` prints nothing, where `BASELINE_SHA` is the HEAD recorded in plan.md §C.1 before the first run-phase commit — so the spec linter still loads the registry and still reports `DanglingRuleReference`. The range starts at a pinned commit rather than a moving ref, and `--first-parent` keeps out the commits a develop absorb brings in, which are reachable only through a merge's second parent (`--no-merges` alone drops the merge commit but still lists them), so another card's change to `internal/spec/lint.go` cannot turn this assertion RED.

Commands: `go test ./internal/constitution/ -run '^TestExecute_RealRegistryShape_Admitted$' -count=1 -v` — expect 1 top-level RUN with subtests `load`, `dry_run`, and `real`; and `go test ./internal/spec/ -run '^TestLinter_AC08_DanglingRuleReference$' -count=1 -v` — expect 1 top-level RUN and `--- PASS: TestLinter_AC08_DanglingRuleReference`.

Shape measured for the witness: a read-only scan of the real registry at `4e9273d0b` prints 101 `file:` lines, 17 distinct values, 0 absolute, 0 containing `..`, 87 beginning with `.claude/`, 14 equal to `CLAUDE.md`, 101 existing (the same figures plan.md §B records at `578afca87`; `git diff --stat 578afca87 4e9273d0b -- .claude/rules/moai/core/zone-registry.md` prints nothing). If the real registry later gains an absolute or `..` `file:` value, this AC fails on the witness rather than on the check: such a registry is refused on the amend path under REQ-CAA-021 (plan.md R-8), while the non-amend callers of `LoadRegistry` keep loading it (D4), and the registry change is reviewed rather than the witness relaxed.

Where the platform refuses the symbolic link, the test calls `t.Skip` with the error; the skip is recorded as a Gap and M-24 is unobserved on that platform.

Baseline-first (predicted from code reading at `54ca2e3b6`): against current code `load` and `dry_run` are expected GREEN — the loader admits relative `file:` values, `load` passes its registry path through `B/link` so the absolute-only check sees it inside `projectDir = B/link` (the `B/root/…` form would be refused today, `internal/constitution/loader.go:84-86`), and `Execute` joins `projectDir`. `dry_run` is a regression guard whose RED cell is M-24: with the candidate resolved to `B/root/…` and the root left as `B/link`, the amend path refuses the registry. `load` calls `LoadRegistry` directly, which runs no containment check under D4, so no declared mutant turns it RED (M-20 variant (iii) reaches it, but the check it moves into the loader resolves `B/link` on both sides and admits the path); it stays as the loader's witness on the real registry shape. `real` is RED on the stub until M5. The `internal/spec` preservation run is expected GREEN against current code (the test exists and this SPEC does not edit it) and must stay GREEN; its RED cell is M-20 variant (iii), predicted from code reading: with the check inside `LoadRegistry`, the relative registry path resolves against the package directory to the repository's `.claude/…`, outside `internal/spec/testdata`, the load is refused, `NewLinter` swallows the error (`internal/spec/lint.go:114-118`), and the expected `DanglingRuleReference` finding is missing.

## §D.2 Mutant list (each must turn its AC RED; record in progress.md §E.2)

| ID | Mutation | Must turn RED |
|---|---|---|
| M-1 | Replace all occurrences (or the first) instead of requiring exactly one | AC-CAA-002 (b) |
| M-2 | Normalize whitespace before counting | AC-CAA-003 (the normalized count `2` found and the exact-match count `0` not found in the path-stripped error — numeric-assertion rule) |
| M-3a | Re-serialize the parsed registry instead of rewriting one line | AC-CAA-004 (line diff ≠ 1) |
| M-3b | Interpolate the new clause into `"…"` without escaping `"` and `\` | AC-CAA-004 (round-trip) |
| M-4 | Skip the re-parse check | AC-CAA-005 (a) |
| M-5a | Remove the restore call on rename failure | AC-CAA-012 `first_rename_applied`, `second_rename`, `third_rename`, `third_rename_applied`, `third_rename_log_absent` |
| M-5b | (i) Restore only the files whose forward rename reported success; separately, (ii) restore bytes but skip removing a file recorded absent | (i) AC-CAA-012 `first_rename_applied`, `third_rename_applied`, `third_rename_log_absent`; (ii) AC-CAA-012 `third_rename_log_absent` |
| M-5c | Rename the log before the registry | AC-CAA-013 |
| M-6 | Drop the legacy concatenated-key aliases | AC-CAA-007 |
| M-7 | Drop the snake_case tag on `rule_id` (read side) | AC-CAA-008 |
| M-8a | Ignore fenced yaml blocks (human format) | AC-CAA-010 (a), AC-CAA-011 |
| M-8b | Treat an unparseable human timestamp as zero time instead of an error | AC-CAA-010 (c), AC-CAA-018 |
| M-9 | Serialize zone as an integer | AC-CAA-006 |
| M-10 | Dry-run skips apply validation (today's behaviour) | AC-CAA-014 (b)–(e) (case (f) is stopped by the REQ-CAA-017 check in `Execute`, which M-10 does not touch) |
| M-11a | Leave temp or backup files after success | AC-CAA-012 `no_fault_clean` |
| M-11b | Dry-run performs the real apply (writes) | AC-CAA-014 snapshot |
| M-12 | Pairwise `---` split (today's parser) | AC-CAA-009 |
| M-13 | CLI dry-run ignores the pipeline error and prints success | AC-CAA-015 `two_occurrences` |
| M-14 | A test writes a fixture path resolved against the repository root, or reads the registry path from the shell environment instead of setting it | AC-CAA-017 |
| M-15 | Drop the line number from the fail-closed error; separately, drop the key; separately, report the block-relative line instead of the file line | AC-CAA-018, each variant on its own: (i) L1 and L2 not found in the path-stripped error; (ii) the key not found; (iii) the block- or segment-relative line found and L1 / L2 not found — numbers under the numeric-assertion rule |
| M-16 | Allow a pre-existing occurrence of the new clause (skip the 0-occurrence precondition) | AC-CAA-019 |
| M-17 | Remove the `Execute` `Before` check while the CLI `--before` check stays | AC-CAA-020 |
| M-18 | On restore failure, delete the backups; separately, omit the backup paths from the error | AC-CAA-021 |
| M-19 | `Execute` keeps its own `projectDir` join instead of the shared resolver | AC-CAA-022; AC-CAA-023 `divergent_root_real` and `divergent_root_dry_run` |
| M-20 | (i) Remove the amend path's containment check at the registry-path site — the check's call on the registry path in `Execute` and in the CLI's registry validation alike — while `LoadRegistry` keeps its present absolute-only refusal (`internal/constitution/loader.go:80-88`, unchanged under D4); separately, (ii) the CLI's registry validation reads the registry without the check while `Execute` keeps it; separately, (iii) move the check into `LoadRegistry` — its calls on the registry path and on every entry's joined `file:` made inside the loader, so every caller of `LoadRegistry` runs them (the placement operator decision D4 excludes); separately, (iv) remove `LoadRegistry`'s present refusal of an absolute registry path outside `projectDir` (`internal/constitution/loader.go:82-87`), leaving the amend path's check in place | (i) AC-CAA-024 `relative_env_escape` (both cases), `symlinked_registry`, and its CLI case; (ii) AC-CAA-024 CLI case; (iii) AC-CAA-024 `loader_unchanged` cases `relative_registry`, `absolute_file`, and `symlinked_registry`, and the AC-CAA-025 `internal/spec` preservation run (`TestLinter_AC08_DanglingRuleReference`); (iv) AC-CAA-024 `loader_unchanged` case `absolute_escape` |
| M-21 | Remove the containment check's call on each entry's joined `file:`; separately, remove its call on the evolution-log path | `file:` variant: AC-CAA-024 `absolute_file` (both cases), `dotdot_file`, `sibling_prefix_file`, `symlinked_file`; log variant: AC-CAA-024 `symlinked_log` |
| M-22 | (i) Replace the separator-boundary comparison with a plain string prefix test on the cleaned, resolved paths, so `B/root-evil/rule.md` passes as inside `B/root`; separately, (ii) compare the uncleaned candidate — the root and the `file:` value concatenated with a separator, with no `filepath.Clean`, no absolutization, and no symbolic-link resolution — against the unresolved `projectDir` as passed, with a plain string prefix test | (i) AC-CAA-024 `sibling_prefix_file`; (ii) AC-CAA-024 `dotdot_file` |
| M-23 | Resolve no symbolic links on the candidate path at one site, leaving the other sites and the root resolved: (i) the registry path; separately, (ii) each entry's joined `file:`; separately, (iii) the evolution-log path | (i) AC-CAA-024 `symlinked_registry`; (ii) AC-CAA-024 `symlinked_file`; (iii) AC-CAA-024 `symlinked_log` |
| M-24 | Resolve symbolic links on the candidate path but not on `projectDir` | AC-CAA-024 `in_root_control/symlinked_root`; AC-CAA-025 `dry_run` |

M-20 variant (i) is not claimed for AC-CAA-023's divergent rows: under D4 the loader's retained absolute-only refusal also stops that absolute registry path, so those rows stay GREEN under it, and M-19 names them instead. M-24 is not claimed for AC-CAA-025 `load`: `load` calls `LoadRegistry` directly, which runs no containment check under D4.

Every mutant whose row lists variants — (i), (ii), (iii) — carries separate variants; each variant is injected and observed on its own, so one kill cannot hide the survival of another. Every mutant and every variant names at least one subtest that only a correct implementation passes, on darwin as on Linux. The base `B` of AC-CAA-024 and AC-CAA-025 is `filepath.EvalSymlinks(t.TempDir())`, so a platform temporary directory under a symbolic link (darwin's `/var`) cannot make a mutant that leaves a candidate unresolved (M-23) refuse every candidate and so pass the `symlinked_*` row it is named for, and M-22 (ii) compares against the unresolved `projectDir`. Every count or line-number assertion follows the numeric-assertion rule of the common rules, so no fixture path can supply the number a mutant fails to print (M-2, M-15). A platform that refuses to create a symbolic link skips the rows that need one, and their mutants are Gaps (AC-CAA-024, AC-CAA-025).

A mutant that cannot be injected, or whose AC run shows fewer top-level RUN lines than stated, is recorded as a Gap, not a kill.

## §D.3 Definition of Done

- All 25 ACs GREEN with commands and verbatim tails in `progress.md` §E.2; baseline-first REDs committed before their production change.
- All 29 mutants, each variant of a variant-carrying mutant on its own, observed RED and reverted; a mutant left unobserved because a platform refused a symbolic link (AC-CAA-024, AC-CAA-025) is a Gap, not a kill.
- `go vet` and `golangci-lint` clean on `internal/constitution` and `internal/cli`.
- The five tests in plan.md §C.2 replaced, none silently deleted.
- plan.md §C.1 real-file sha256 equals the post-run sha256 (AC-CAA-017).
