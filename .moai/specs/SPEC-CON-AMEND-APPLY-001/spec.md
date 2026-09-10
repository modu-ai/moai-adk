---
id: SPEC-CON-AMEND-APPLY-001
title: "Constitution amendment apply step: exact-once source replacement, line-scoped registry update, readable evolution log, and three-file atomic apply"
version: "0.1.2"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/constitution, internal/cli"
lifecycle: spec-anchored
tags: "constitution, amendment, atomic-apply, evolution-log, rate-limiter, dry-run, registry-resolver, path-containment, t659"
tier: M
related_specs: [SPEC-V3R2-CON-002]
---

# SPEC-CON-AMEND-APPLY-001 — Constitution amendment apply step

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.2 | 2026-09-11 | manager-spec | Lead ruling G6 (`.moai/reports/t659/verdict.md` §9, option (ii)) applied. New REQ-CAA-020 — the registry, the source rule file, and the evolution log all come from inside `projectDir`; the registry loader's refusal of a registry path outside `projectDir` is the intended boundary, so a `CLAUDE_PROJECT_DIR` naming another tree stops `Execute` with a registry load error before any write (new AC-CAA-023, new mutant M-20). Option (i) — making the source and log paths follow the resolved root — recorded as rejected in §C, §D.2, and §F. AC-CAA-022 and AC-CAA-023 reconciled: the resolver chooses the path, the loader admits or refuses it. G6 removed from §G; one question surfaced while encoding it (G7) is recorded there. IDs kept stable; new IDs appended. |
| 0.1.1 | 2026-09-11 | manager-spec | Lead rulings on the plan design gaps (`.moai/reports/t659/verdict.md` §8) applied. G1: REQ-CAA-009 amended — the fail-closed error names file path, line number, and key (new AC-CAA-018). G2: new REQ-CAA-016 — the new clause must occur 0 times in the source file before apply (AC-CAA-019). G3: new REQ-CAA-017 — `Execute` rejects `Before` ≠ current clause (AC-CAA-020). G4: REQ-CAA-010 amended to on-disk backups, REQ-CAA-011 amended with a restore seam, new REQ-CAA-018 — a failed restore keeps the backups and names them (AC-CAA-021). G5: non-dry-run CLI path recorded as Gap §E.4. Scope (a): new REQ-CAA-019 — one registry path resolver shared by CLI and `Execute` (AC-CAA-022); removed from exclusions. REQ-CAA-012 and REQ-CAA-015 amended for the new validations and the environment the resolver reads. Exclusions (b)(c)(d) recorded as follow-up card candidates. §G now holds no G1–G5 or (a) item. IDs kept stable; new IDs appended. |
| 0.1.0 | 2026-09-11 | manager-spec | Initial plan-phase draft for card t659. Encodes the lead rulings in `.moai/reports/t659/verdict.md` §7 (Q1-Q5) without reopening them. Delivers SPEC-V3R2-CON-002 REQ-CON-002-011 (three-file atomic apply), which that SPEC's frontmatter reports as `implemented` but which the repro in verdict §2.1 shows is not. |

Card: **t659** (lane-6). Worktree `.claude/worktrees/t659`, branch `WT-amend-apply`. Code coordinates were read at `034d55c56`; the 0.1.1 revision was authored on `ff11e752f`, whose only change after `7b4d1ac89` is verdict §8 (no code change). The 0.1.2 revision was authored on `92c8c3f36`, whose only change after `01af243ae` is verdict §9 (`git diff --stat 01af243ae 92c8c3f36` lists `.moai/reports/t659/verdict.md` only).

## §A Problem — measured shape

The evidence below is from `.moai/reports/t659/verdict.md` §2 (run on tree `5a066994b`, go1.26.8 darwin/arm64) and was re-read at `034d55c56` for this draft. It is cited, not re-measured, except where marked.

1. **The real apply path never applies.** `Pipeline.Execute(dryRun=false)` passes all five gates and approval, then `applyAmendment` stops at `updateSourceFile`, which returns `not yet implemented` (`internal/constitution/pipeline.go:256-260`). `updateRegistryClause` is the same stub (`pipeline.go:264-267`). The registry update and the evolution-log append are unreachable. Five characterization tests pin this stub behaviour and pass today (verdict §2.1).
2. **Dry-run reports success without validating anything.** The dry-run branch (`pipeline.go:133-137`) returns `createLogEntry` and never calls an apply function, so a dry-run cannot surface the failures a real apply would hit.
3. **The evolution-log schema is mismatched on both sides.** `AmendmentLog` (`internal/constitution/amendment.go:192-219`) carries no yaml tags, so the writer emits concatenated-lowercase keys (`ruleid`, `approvedat`) and integer zones (`zonebefore: 0`). A log written in snake_case (the SPEC-V3R2-CON-002 REQ-CON-002-004 notation, and the existing test fixture) reads back with empty `RuleID` and zero `ApprovedAt` (verdict §2.2). `TestLoadEvolutionLogs` asserts only `ID`, so the mismatch passes.
4. **The real log parses to zero entries, which blinds the rate limiter.** `LoadEvolutionLogs` splits on the substring `---`. The tracked `.moai/research/evolution-log.md` is human-authored: a HISTORY table with `|---|` separator rows, `---` horizontal rules, and one `## EVO-HRN-002` heading followed by a fenced yaml block. The split yields zero segments with a top-level `id:` (verdict §2.3). `rateLimiter.Admit` (`internal/constitution/rate_limiter.go:41-121`) therefore counts nothing — no 7-day window, no cooldown, no active cap.
5. **The exact-once rule is attainable on the real corpus.** Of 101 registry entries, the 97 live ones have their clause occur exactly once in their source file; the 4 with zero occurrences are the `[SUPERSEDED …]` retired entries (CONST-V3R2-021..024) (verdict §2.3). All 101 clauses are one-line double-quoted scalars inside the single yaml fence of `zone-registry.md`.
6. **The CLI and `Execute` can read different registries.** `runConstitutionAmend` validates `--before` against the registry `resolveRegistryPath` returns (`MOAI_CONSTITUTION_REGISTRY`, then `CLAUDE_PROJECT_DIR`, then cwd — `internal/cli/constitution.go:144-155`), while `Execute` loads `<projectDir>/.claude/rules/moai/core/zone-registry.md` (`pipeline.go:66`). Once this card turns writing on, that split is a write to a file the user never validated (verdict §8, scope addition (a)). When `CLAUDE_PROJECT_DIR` names a tree other than `projectDir`, the resolver returns a registry path in that other tree, and the registry loader already refuses it: `LoadRegistry` rejects an absolute path that escapes `projectDir` (`internal/constitution/loader.go:80-88`, read at `92c8c3f36`). Verdict §9 makes that refusal the boundary.

## §B Goal

When all five gates pass and the user approves, the amendment lands in the source rule file, the registry, and the evolution log together or not at all; the registry written is the registry the CLI validated; a dry-run fails exactly where a real apply would fail; and the rate limiter reads every entry the log actually contains, including the human-authored one.

## §C Lead rulings encoded (verdict §7, §8, and §9 — authoritative, not reopened)

| Ruling | Encoded as |
|---|---|
| Q1 Source replacement: exact clause string, exactly once in the whole file; 0 or ≥2 fails before any write; no whitespace normalization; no anchor narrowing | REQ-CAA-001, REQ-CAA-002 |
| Q2 Registry: replace only the target entry's `clause:` line inside the yaml fence as a string; re-parse immediately and verify entry count and target clause; re-serialization rejected | REQ-CAA-003, REQ-CAA-004 |
| Q3 Log: explicit snake_case yaml tags for writing; reading also accepts legacy concatenated keys; zone serialized as the registry writes it; parser reads the real file including human-format entries | REQ-CAA-005 … REQ-CAA-009 |
| Q4 Atomicity: back up all three; temp-write each; rename source → registry → log; any failure restores all three from backups; fault-injection tests for 2nd and 3rd rename; dry-run runs the apply validation | REQ-CAA-010, REQ-CAA-011, REQ-CAA-012 |
| Q5 Scope: REQ-CON-002-012 and the SPEC-V3R2-CON-002 status correction are separate cards | §F Out of Scope |
| §7 CLI slot not granted; CLI behaviour verified in run through `internal/cli` tests | REQ-CAA-013, §E.3 |
| §8 G1 Fail-closed confirmed; the error names the unreadable entry by file, line, and key | REQ-CAA-009 (amended) |
| §8 G2 The new clause must occur 0 times in the source file before apply | REQ-CAA-016 |
| §8 G3 `Execute` rejects `Before` ≠ current clause; the CLI check stays as early guidance | REQ-CAA-017 |
| §8 G4 A failed restore keeps the backups and returns their paths; one restore-failure fault-injection test | REQ-CAA-010, REQ-CAA-011 (amended), REQ-CAA-018 |
| §8 G5 CLI tests cover dry-run only; the non-dry-run CLI path is a Gap covered by `Execute`-level tests | §E.4 |
| §8 (a) One registry path resolver shared by the CLI and `Execute` | REQ-CAA-019 |
| §8.1 items 3–5 follow-up card candidates | §F |
| §9 G6 option (ii): the registry loader's refusal of a registry path outside `projectDir` is the intended boundary; the registry, the source rule file, and the evolution log all come from the one project root | REQ-CAA-020 |
| §9 G6 option (i) rejected: making the source rule file and evolution-log paths follow the resolved root would widen the set of directory trees an amendment can write | §D.2 rationale under REQ-CAA-020, §F |

## §D Requirements (GEARS)

Terms used below. The **apply step** is the part of `Execute` that runs after Layer 5 approval (and, per REQ-CAA-012, its validation half under dry-run). The **registry path** is the path returned by the resolver of REQ-CAA-019. The **three files** are the target rule's source file, the registry at the registry path, and the evolution log at `<projectDir>/.moai/research/evolution-log.md`. The **current clause** is the target entry's `clause` value as decoded by the registry loader. The **pre-apply validations** are REQ-CAA-001 … REQ-CAA-004, REQ-CAA-016, and REQ-CAA-017.

### §D.1 Source rule file

- **REQ-CAA-001** (event-driven) — **When** the apply step updates the source rule file, the apply step shall replace the current clause with the new clause only where the current clause occurs as an exact byte sequence exactly once in the whole file, and every byte outside that one occurrence shall remain unchanged. The match shall use no whitespace normalization and no narrowing to the section named by the entry's `anchor`.

- **REQ-CAA-002** (event-detected) — **When** the current clause occurs zero times or two or more times in the source rule file, the apply step shall return an error that names the file path and the occurrence count, and shall not create, modify, rename, or remove any of the three files.

- **REQ-CAA-016** (event-detected) — **When** the new clause already occurs as an exact byte sequence one or more times in the source rule file before the apply, the apply step shall return an error that names the file path and that occurrence count, and shall not create, modify, rename, or remove any of the three files. The check uses the same exact matching as REQ-CAA-001; a new clause contained inside the current clause therefore also counts as an occurrence and is rejected.

### §D.2 Zone registry

- **REQ-CAA-003** (event-driven) — **When** the apply step updates the zone registry, the apply step shall rewrite only the `clause:` line belonging to the target entry inside the registry's yaml fence, encoding the new clause as a single-line yaml scalar that decodes to exactly the new clause; every other line of the registry, including comments, entry order, and the quoting of other entries, shall remain byte-identical. The registry shall not be produced by re-serializing parsed entries.

- **REQ-CAA-004** (event-detected) — **When** the rewritten registry content is produced, the apply step shall parse it with the same fence extraction and yaml decoding the registry loader uses, before any file is renamed into place; **when** that parse fails, the entry count differs from the pre-apply count, the target entry's decoded clause differs from the new clause, or the target entry has no single `clause:` line to rewrite, the apply step shall return an error and shall leave all three files byte-identical to their pre-apply contents.

- **REQ-CAA-017** (event-detected) — **When** `Execute` receives a proposal whose `Before` differs from the current clause as an exact byte sequence, `Execute` shall return an error naming the rule ID before Layer 1 runs, in both dry-run and real mode, and shall not create, modify, rename, or remove any of the three files; a lock it acquired shall be released. The CLI's own `--before` comparison stays as early guidance and does not replace this check.

- **REQ-CAA-019** (ubiquitous) — The CLI's registry validation and `Execute` shall obtain the registry path from one shared resolver whose precedence is: a non-empty `MOAI_CONSTITUTION_REGISTRY`, then `<CLAUDE_PROJECT_DIR>/.claude/rules/moai/core/zone-registry.md` when `CLAUDE_PROJECT_DIR` is non-empty, then `<projectDir>/.claude/rules/moai/core/zone-registry.md`; `Execute` shall not construct the registry path by any other join, so the file the CLI validates is the file `Execute` reads and writes.

- **REQ-CAA-020** (ubiquitous) — The registry, the source rule file, and the evolution log that one `Execute` call reads and writes shall all lie inside the same project root, `projectDir`; no read or write of the apply step shall reach another directory tree. **When** the registry path of REQ-CAA-019 lies outside `projectDir` — as it does when `MOAI_CONSTITUTION_REGISTRY` is empty and `CLAUDE_PROJECT_DIR` names a directory outside `projectDir` — `Execute` shall return the registry loader's load error, in dry-run and real mode alike, before Layer 1 and before the check of REQ-CAA-017, and shall leave every file in both trees byte-identical, adding or removing no path in either; a lock it acquired shall be released. The source rule file and the evolution log shall keep resolving against `projectDir`, never against the root the environment names.

  Rationale (verdict §9, G6): the loader's refusal (`internal/constitution/loader.go:80-88`) is the boundary, not a gap to close. The alternative, option (i) — letting the source rule file and evolution-log paths follow the resolved root, so that a divergent environment moves all three files to the other tree — was rejected because it widens the set of directory trees an amendment can write. How this fits with REQ-CAA-019: REQ-CAA-019 decides *which* registry path is chosen; this requirement decides *whether* the chosen path is admitted. Together they mean a `CLAUDE_PROJECT_DIR` naming a tree outside `projectDir` can only yield a refused path, while an override inside `projectDir` (AC-CAA-022) is admitted. Two path shapes this boundary does not yet reach are §G item 1 (G7).

### §D.3 Evolution log

- **REQ-CAA-005** (ubiquitous) — The evolution-log writer shall emit each entry with the snake_case keys `id`, `rule_id`, `zone_before`, `zone_after`, `clause_before`, `clause_after`, `canary_verdict`, `contradictions`, `approved_by`, `approved_at`, `rolled_back`, `rollback_reason`, and `rollback_at`, and shall write `zone_before` and `zone_after` as the zone names the registry uses (`Frozen`, `Evolvable`), never as integers.

- **REQ-CAA-006** (ubiquitous) — The evolution-log reader shall accept, for every field in REQ-CAA-005, both the snake_case key and the legacy concatenated-lowercase key the untagged writer produced (`ruleid`, `zonebefore`, `zoneafter`, `clausebefore`, `clauseafter`, `canaryverdict`, `approvedby`, `approvedat`, `rolledback`, `rollbackreason`, `rollbackat`), and shall accept a zone written either as a zone name or as the legacy integer (`0` = Frozen, `1` = Evolvable). **When** one entry carries both key forms for the same field, the reader shall use the snake_case value.

- **REQ-CAA-007** (unwanted) — The evolution-log reader shall not drop, merge, or fabricate entries because of markdown horizontal rules (`---`), table separator rows (`|---|`), or other `---` substrings elsewhere in the file; a `---`-delimited entry shall be recognized whenever its content decodes to a mapping with a non-empty top-level `id`, regardless of how many `---` lines precede it.

- **REQ-CAA-008** (capability gate) — **Where** the evolution log contains a fenced yaml code block whose decoded top-level mapping has a non-empty `id` (the human-authored format of `## EVO-HRN-002`), the reader shall return that block as an entry, mapping its fields onto the entry as follows, so that the rate limiter counts it:

  | Entry field | Taken from (first present wins) | When none is present |
  |---|---|---|
  | `ID` | `id` | block is not an entry |
  | `RuleID` | `rule_id`, then `const_registry_entry` | empty |
  | `ApprovedAt` | `approved_at`, then `timestamp`; a date-only `YYYY-MM-DD` value means 00:00:00 UTC of that date; RFC 3339 is also accepted | REQ-CAA-009 |
  | `ZoneBefore`, `ZoneAfter` | `zone_before` / `zone_after`, then `zone` for both | Frozen (zero value) |
  | `ApprovedBy` | `approved_by`, then `approver` | empty |
  | `CanaryVerdict` | `canary_verdict` | empty |
  | `RolledBack` | `rolled_back` | `false` (counted as active) |
  | `RollbackAt`, `RollbackReason` | `rollback_at`, `rollback_reason` | nil / empty |
  | `ClauseBefore`, `ClauseAfter`, `Contradictions` | the snake_case keys only | empty (the rate limiter does not read them) |

  Human-format keys with no entry field (`spec_id`, `target_file`, `before_snippet`, `after_snippet`, `con_002_layers`, `status`, and the rest) shall be ignored by the reader, not rejected.

- **REQ-CAA-009** (event-detected) — **When** a block the reader recognizes as an entry (REQ-CAA-007 or REQ-CAA-008) has no approval timestamp that parses under REQ-CAA-008, the reader shall return an error — so that the rate limiter fails closed rather than silently leaving the entry out of the 7-day window and the cooldown — and that error shall contain the log file path, the 1-based line number in that file, the key, and the entry's `id`. **Where** a timestamp key (`approved_at`, `approvedat`, or `timestamp`) is present but unparseable, the line and key shall be that key's; **where** no timestamp key is present, the key shall be named `approved_at` and the line shall be that of the entry's `id` key.

### §D.4 Atomic apply

- **REQ-CAA-010** (event-driven) — **When** the apply step writes the three files, it shall first complete every pre-apply validation, then write a backup file holding the pre-apply bytes of each of the three files in that file's own directory (recording a file that does not yet exist as absent, with no backup file for it), then write each file's new content to a temporary file in that file's own directory, then rename the temporary files into place in the order source rule file, registry, evolution log. **When** any step after the first backup fails — a backup, a temporary write, or any of the three renames — the apply step shall restore all three files to their pre-apply bytes from the backups (removing a file recorded as absent) and return an error naming the failed step. After a successful apply, and after a completed restore, no temporary or backup file created by the apply step shall remain. The evolution-log content shall be the pre-apply bytes followed by the new entry, so every pre-existing byte of the log, human-authored entries included, is preserved.

- **REQ-CAA-011** (ubiquitous) — The apply step shall perform its forward renames through one replaceable per-pipeline operation and its restore writes through a second, separate replaceable per-pipeline operation, so that a test can fail the Nth forward rename, and independently fail a restore, against `t.TempDir()` fixtures without touching real files and without mutating process-global state; the restore path shall not route through the forward-rename operation.

- **REQ-CAA-018** (event-detected) — **When** restoring any of the three files from its backup fails, the apply step shall not delete any backup file, shall leave every backup file holding its pre-apply bytes, and shall return an error that contains the path of every backup file it wrote together with the failed restore step.

### §D.5 Dry-run and CLI

- **REQ-CAA-012** (state-driven) — **While** `Execute` runs in dry-run mode, `Execute` shall perform every pre-apply validation against the real files at the registry path and the source and log paths, and return the same error a real apply would return, and shall not create, modify, rename, or remove any file — no backup, no temporary file, no lock file.

- **REQ-CAA-013** (event-driven) — **When** `moai constitution amend` runs with `--dry-run` (or `MOAI_CONSTITUTION_DRY_RUN=true`) and a pre-apply validation fails, the command shall return a non-nil error and shall not print its dry-run success line.

### §D.6 Test suite

- **REQ-CAA-014** (ubiquitous) — The `internal/constitution` test suite shall assert success-path and failure-path outcomes of the apply step and shall contain no assertion that an apply function returns `not yet implemented`: the five stub-characterization tests `TestPipeline_Execute_NonDryRun_AmendmentStubError`, `TestPipeline_applyAmendment_StubError`, `TestUpdateSourceFile_StubError`, `TestUpdateRegistryClause_StubError`, and the success-without-validation assumption of `TestPipeline_Execute_DryRun_Success` shall each be replaced by the assertions named in `plan.md` §C.2, not deleted without replacement. `TestLoadEvolutionLogs` shall assert `RuleID` and `ApprovedAt` in addition to `ID`.

- **REQ-CAA-015** (unwanted) — No test or verification command for this SPEC shall write the repository's real `zone-registry.md`, any real rule file under `.claude/rules/`, or the real `.moai/research/evolution-log.md`; every write shall target a `t.TempDir()` fixture, and every test that reaches the resolver of REQ-CAA-019 shall set `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` itself (empty, or a path inside its `t.TempDir()`), so that a value exported by the surrounding session cannot redirect it to the real registry.

## §E Constraints

### §E.1 Scope boundary (What, not How)

Function names, the shapes of the two seams, the resolver's package home, and backup naming are run-phase decisions guided by `plan.md`; this document fixes only observable behaviour.

### §E.2 Stability

- Five-layer gate order and every gate's semantics are unchanged. REQ-CAA-017 is a precondition checked before Layer 1, not a sixth layer.
- `acquireLock` / `releaseLock` behaviour is unchanged (see §F on the default lock path).
- The registry loader's parse rules (`internal/constitution/loader.go`) are unchanged; REQ-CAA-004 reuses them rather than adding a second parser. Its containment check (`loader.go:80-88`) is unchanged too, and is now required behaviour (REQ-CAA-020) rather than an incidental guard.
- The resolver's precedence is the one `resolveRegistryPath` has today; REQ-CAA-019 moves `Execute` onto it and does not change what the CLI resolves.
- No new third-party dependency (`gopkg.in/yaml.v3` is already used).

### §E.3 Verification route

- Package-level behaviour is verified in `internal/constitution` tests with `t.TempDir()` fixtures and in-package test doubles (`fakeOversight`, the injectable `rateLimiter.now`, the two seams of REQ-CAA-011). None of the affected test files uses `t.Parallel` today (`grep -c 't.Parallel()'` returns 0 for `internal/constitution/pipeline_test.go`, `evolution_log_test.go`, and `internal/cli/constitution_test.go`, `constitution_integration_test.go`, `constitution_guard_test.go` at `ff11e752f`), so `t.Setenv` is usable for REQ-CAA-015.
- CLI-level behaviour is verified in `internal/cli` tests that call `runConstitutionAmend` directly. No built binary and no `HOME` override are used. **Correction to the dispatch premise:** verdict §7 calls this a "home seam", but `runConstitutionAmend` and `Pipeline.Execute` read no home directory. The seam is the `projectDir` parameter pointed at `t.TempDir()`, with `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` set by the test (REQ-CAA-015).

### §E.4 Gaps (approved reductions)

- **G5 — non-dry-run CLI path not tested at CLI level.** Approved by the lead (verdict §8 G5). CLI tests cover dry-run only (AC-CAA-015). The real-apply path through `runConstitutionAmend` is not exercised by any test in this SPEC: it would need a stdin seam (`NewHumanOversight` hardcodes `os.Stdin`, `internal/constitution/human_oversight.go:20-24`) and would create the cwd-relative lock file inside the package directory (§F). The apply itself is covered at `Execute` level (AC-CAA-001, 002, 005, 012, 013, 019, 020, 021, 022, 023). What stays unobserved is only the CLI wrapper around a successful or failed real apply: its output lines and exit status.

## §F Out of Scope

### Out of Scope — rejected-amendment logging

- REQ-CON-002-012 (writing `.moai/research/rejected-amendments/LEARN-YYYYMMDD-NNN.md` when a gate fails). Separate card candidate per verdict §7.1 item 1; measured absent (`rejected-amendments` has 0 Go references, verdict §2.4).

### Out of Scope — SPEC-V3R2-CON-002 status

- Changing SPEC-V3R2-CON-002's frontmatter `status: implemented`. Separate card candidate per verdict §7.1 item 2. That SPEC is not edited by this card.

### Out of Scope — matching variants rejected by the lead

- Anchor-scoped matching (narrowing the clause search to the entry's `anchor` section).
- Whitespace-normalized matching. The existing validator compares normalized text; the apply step deliberately does not (Q1).
- Re-serializing the registry yaml.

### Out of Scope — default lock path

- The default lock path `.moai/research/.amendment.lock` (`pipeline.go:227`) is relative to the process working directory, not to `projectDir`. **Observed issue, recorded, not fixed here.** The atomicity design of REQ-CAA-010 does not need it: the lock serializes concurrent amenders, while atomicity concerns one amender's three writes. It contributes to the G5 reduction (§E.4).

### Out of Scope — source and log paths following the resolved root

- Resolving the source rule file and the evolution log against the root `CLAUDE_PROJECT_DIR` names (verdict §9, G6 option (i)). Rejected: it widens the set of directory trees an amendment can write. REQ-CAA-020 keeps all three files inside `projectDir` instead.

### Out of Scope — follow-up card candidates (verdict §8.1)

- **Dry-run environment value mismatch** (verdict §8.1 item 3). The CLI reads `MOAI_CONSTITUTION_DRY_RUN == "true"` (`internal/cli/constitution.go:470`) while SPEC-V3R2-CON-002 REQ-CON-002-031 and AC-CON-002-07 say `=1`. Not changed here.
- **EVO-HRN-002 entry mismatch** (verdict §8.1 item 4). The entry names `const_registry_entry: CONST-V3R2-153`, whose registry `file:` is `.claude/rules/moai/workflow/session-handoff.md` (`zone-registry.md:678-681`), while its own `target_file` is `.claude/rules/moai/design/constitution.md`. REQ-CAA-008 maps the value as written; correcting the log is not this card's business. The rate limiter uses `RuleID` only in its rollback check, which this entry (no `rolled_back`) never triggers.
- **`MarkRolledBack` lossy rewrite** (verdict §8.1 item 5). `MarkRolledBack` / `rewriteEvolutionLog` rewrite the whole log in machine format; once REQ-CAA-008 makes human entries readable, a rollback would re-serialize them and lose their other fields. It has zero production callers (`grep -rn 'MarkRolledBack(' --include='*.go' internal cmd pkg | grep -v _test.go` prints only its definition at `evolution_log.go:88`, tree `034d55c56`), so the hazard is latent. Not changed here.

### Out of Scope — other boundaries

- Crash recovery across process death between renames. Backups that survive a failed restore (REQ-CAA-018) are left for a human; no automatic recovery on the next run.
- The template mirror of `zone-registry.md` under `internal/template/templates/`: an apply writes only the registry at the registry path.

## §G Open questions for the lead

All §7, §8, and §9 items (Q1–Q5, G1–G6, scope addition (a)) are resolved and encoded in §C; none of them remains open. One question surfaced while encoding G6 and is not decided here:

1. **G7 — path shapes the containment boundary does not reach.** REQ-CAA-020 states that nothing the apply step reads or writes lies outside `projectDir`. Code reading at `92c8c3f36` (not executed) finds two shapes the current checks do not refuse:
   - **A relative registry path.** `LoadRegistry` checks containment only when the cleaned path is absolute (`loader.go:82`); a relative path goes straight to `os.ReadFile` (`loader.go:90`) and is read against the process working directory, not `projectDir`. The resolver returns a relative path when `MOAI_CONSTITUTION_REGISTRY` holds a relative value, or when it is empty and `CLAUDE_PROJECT_DIR` holds one (for example `../other`).
   - **A registry entry whose `file:` is absolute or climbs out with `..`.** `applyAmendment` uses an absolute `file:` as given and joins a relative one to `projectDir` with no containment check (`pipeline.go:192-195`), so the source rule file an apply writes can lie outside `projectDir`. On the real registry this shape is latent: of 101 `file:` lines, 0 start with `/` and 0 contain `..` (`/usr/bin/grep -cE` over `.claude/rules/moai/core/zone-registry.md` at `92c8c3f36`; control: 87 of the 101 start with `.claude/`).

   Should REQ-CAA-020 be enforced for both shapes (run-phase adds the refusals, each with an AC and a mutant), or be narrowed to an absolute registry path, with both shapes recorded in §F as observed and not fixed? Until this is decided, AC-CAA-023 covers only the shape verdict §9 names.

## §H Cross-References

- `.moai/reports/t659/verdict.md` — repro evidence (§2), lead rulings (§7, §8, §9), follow-up candidates (§7.1, §8.1)
- `.moai/specs/SPEC-V3R2-CON-002/spec.md` — REQ-CON-002-004 (log fields), REQ-CON-002-011 (atomic apply), REQ-CON-002-031 (dry-run)
- `internal/constitution/pipeline.go`, `evolution_log.go`, `amendment.go`, `loader.go`, `rate_limiter.go`, `human_oversight.go`
- `internal/cli/constitution.go` — `newConstitutionAmendCmd`, `runConstitutionAmend`, `resolveRegistryPath`
- `internal/config/envkeys.go:353` — `EnvClaudeProjectDir`
- `internal/constitution/registry_sync_test.go` — existing literal-clause (no normalization) guard on the real registry, the read-only precedent for REQ-CAA-001's exact matching
