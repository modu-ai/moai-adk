---
id: SPEC-CON-AMEND-APPLY-001
title: "Constitution amendment apply step: exact-once source replacement, line-scoped registry update, readable evolution log, and three-file atomic apply"
version: "0.1.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/constitution, internal/cli"
lifecycle: spec-anchored
tags: "constitution, amendment, atomic-apply, evolution-log, rate-limiter, dry-run, t659"
tier: M
related_specs: [SPEC-V3R2-CON-002]
---

# SPEC-CON-AMEND-APPLY-001 — Constitution amendment apply step

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-11 | manager-spec | Initial plan-phase draft for card t659. Encodes the lead rulings in `.moai/reports/t659/verdict.md` §7 (Q1-Q5) without reopening them. Delivers SPEC-V3R2-CON-002 REQ-CON-002-011 (three-file atomic apply), which that SPEC's frontmatter reports as `implemented` but which the repro in verdict §2.1 shows is not. |

Card: **t659** (lane-6). Worktree `.claude/worktrees/t659`, branch `WT-amend-apply`. Every code coordinate below was read at HEAD `034d55c56`.

## §A Problem — measured shape

The evidence below is from `.moai/reports/t659/verdict.md` §2 (run on tree `5a066994b`, go1.26.8 darwin/arm64) and was re-read at `034d55c56` for this draft. It is cited, not re-measured, except where marked.

1. **The real apply path never applies.** `Pipeline.Execute(dryRun=false)` passes all five gates and approval, then `applyAmendment` stops at `updateSourceFile`, which returns `not yet implemented` (`internal/constitution/pipeline.go:256-260`). `updateRegistryClause` is the same stub (`pipeline.go:264-267`). The registry update and the evolution-log append are unreachable. Five characterization tests pin this stub behaviour and pass today (verdict §2.1).
2. **Dry-run reports success without validating anything.** The dry-run branch (`pipeline.go:133-137`) returns `createLogEntry` and never calls an apply function, so a dry-run cannot surface the failures a real apply would hit.
3. **The evolution-log schema is mismatched on both sides.** `AmendmentLog` (`internal/constitution/amendment.go:192-219`) carries no yaml tags, so the writer emits concatenated-lowercase keys (`ruleid`, `approvedat`) and integer zones (`zonebefore: 0`). A log written in snake_case (the SPEC-V3R2-CON-002 REQ-CON-002-004 notation, and the existing test fixture) reads back with empty `RuleID` and zero `ApprovedAt` (verdict §2.2). `TestLoadEvolutionLogs` asserts only `ID`, so the mismatch passes.
4. **The real log parses to zero entries, which blinds the rate limiter.** `LoadEvolutionLogs` splits on the substring `---`. The tracked `.moai/research/evolution-log.md` is human-authored: a HISTORY table with `|---|` separator rows, `---` horizontal rules, and one `## EVO-HRN-002` heading followed by a fenced yaml block. The split yields zero segments with a top-level `id:` (verdict §2.3). `rateLimiter.Admit` (`internal/constitution/rate_limiter.go:41-121`) therefore counts nothing — no 7-day window, no cooldown, no active cap.
5. **The exact-once rule is attainable on the real corpus.** Of 101 registry entries, the 97 live ones have their clause occur exactly once in their source file; the 4 with zero occurrences are the `[SUPERSEDED …]` retired entries (CONST-V3R2-021..024) (verdict §2.3). All 101 clauses are one-line double-quoted scalars inside the single yaml fence of `zone-registry.md`.

## §B Goal

When all five gates pass and the user approves, the amendment lands in the source rule file, the registry, and the evolution log together or not at all; a dry-run fails exactly where a real apply would fail; and the rate limiter reads every entry the log actually contains, including the human-authored one.

## §C Lead rulings encoded (verdict §7 — authoritative, not reopened)

| Ruling | Encoded as |
|---|---|
| Q1 Source replacement: exact clause string, exactly once in the whole file; 0 or ≥2 fails before any write; no whitespace normalization; no anchor narrowing | REQ-CAA-001, REQ-CAA-002 |
| Q2 Registry: replace only the target entry's `clause:` line inside the yaml fence as a string; re-parse immediately and verify entry count and target clause; re-serialization rejected | REQ-CAA-003, REQ-CAA-004 |
| Q3 Log: explicit snake_case yaml tags for writing; reading also accepts legacy concatenated keys; zone serialized as the registry writes it; parser reads the real file including human-format entries | REQ-CAA-005 … REQ-CAA-009 |
| Q4 Atomicity: back up all three; temp-write each; rename source → registry → log; any failure restores all three from backups; fault-injection tests for 2nd and 3rd rename; dry-run runs the apply validation | REQ-CAA-010, REQ-CAA-011, REQ-CAA-012 |
| Q5 Scope: REQ-CON-002-012 and the SPEC-V3R2-CON-002 status correction are separate cards | §F Out of Scope |
| CLI slot not granted; CLI behaviour verified in run through `internal/cli` tests | REQ-CAA-013, §E.3 |

## §D Requirements (GEARS)

Terms used below. The **apply step** is the part of `Execute` that runs after Layer 5 approval (and, per REQ-CAA-012, its validation half under dry-run). The **three files** are the target rule's source file, the zone registry at `<projectDir>/.claude/rules/moai/core/zone-registry.md`, and the evolution log at `<projectDir>/.moai/research/evolution-log.md`. The **current clause** is the target entry's `clause` value as decoded by the registry loader.

### §D.1 Source rule file

- **REQ-CAA-001** (event-driven) — **When** the apply step updates the source rule file, the apply step shall replace the current clause with the new clause only where the current clause occurs as an exact byte sequence exactly once in the whole file, and every byte outside that one occurrence shall remain unchanged. The match shall use no whitespace normalization and no narrowing to the section named by the entry's `anchor`.

- **REQ-CAA-002** (event-detected) — **When** the current clause occurs zero times or two or more times in the source rule file, the apply step shall return an error that names the file path and the occurrence count, and shall not create, modify, rename, or remove any of the three files.

### §D.2 Zone registry

- **REQ-CAA-003** (event-driven) — **When** the apply step updates the zone registry, the apply step shall rewrite only the `clause:` line belonging to the target entry inside the registry's yaml fence, encoding the new clause as a single-line yaml scalar that decodes to exactly the new clause; every other line of the registry, including comments, entry order, and the quoting of other entries, shall remain byte-identical. The registry shall not be produced by re-serializing parsed entries.

- **REQ-CAA-004** (event-detected) — **When** the rewritten registry content is produced, the apply step shall parse it with the same fence extraction and yaml decoding the registry loader uses, before any file is renamed into place; **when** that parse fails, the entry count differs from the pre-apply count, the target entry's decoded clause differs from the new clause, or the target entry has no single `clause:` line to rewrite, the apply step shall return an error and shall leave all three files byte-identical to their pre-apply contents.

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

- **REQ-CAA-009** (event-detected) — **When** a block the reader recognizes as an entry (REQ-CAA-007 or REQ-CAA-008) has no approval timestamp that parses under REQ-CAA-008, the reader shall return an error naming that entry's `id`, so that the rate limiter fails closed rather than silently leaving the entry out of the 7-day window and the cooldown.

### §D.4 Atomic apply

- **REQ-CAA-010** (event-driven) — **When** the apply step writes the three files, it shall first complete every validation in REQ-CAA-001 … REQ-CAA-004, then back up the pre-apply bytes of all three files (recording a file that does not yet exist as absent), then write each file's new content to a temporary file in that file's own directory, then rename the temporary files into place in the order source rule file, zone registry, evolution log. **When** any step after the first backup fails — a backup, a temporary write, or any of the three renames — the apply step shall restore all three files to their pre-apply bytes (removing a file recorded as absent) and return an error naming the failed step. After a successful apply, and after a completed restore, no temporary or backup file created by the apply step shall remain. The evolution-log content shall be the pre-apply bytes followed by the new entry, so every pre-existing byte of the log, human-authored entries included, is preserved.

- **REQ-CAA-011** (ubiquitous) — The apply step shall perform its forward renames through a replaceable per-pipeline operation, so that a test can fail the Nth forward rename against `t.TempDir()` fixtures without touching real files and without mutating process-global state; the restore path shall not route through that replaceable operation.

### §D.5 Dry-run and CLI

- **REQ-CAA-012** (state-driven) — **While** `Execute` runs in dry-run mode, the apply step shall perform the validations of REQ-CAA-001 … REQ-CAA-004 against the real files and return the same error a real apply would return, and shall not create, modify, rename, or remove any file — no backup, no temporary file, no lock file.

- **REQ-CAA-013** (event-driven) — **When** `moai constitution amend` runs with `--dry-run` (or `MOAI_CONSTITUTION_DRY_RUN=true`) and the apply-step validation fails, the command shall return a non-nil error and shall not print its dry-run success line.

### §D.6 Test suite

- **REQ-CAA-014** (ubiquitous) — The `internal/constitution` test suite shall assert success-path and failure-path outcomes of the apply step and shall contain no assertion that an apply function returns `not yet implemented`: the five stub-characterization tests `TestPipeline_Execute_NonDryRun_AmendmentStubError`, `TestPipeline_applyAmendment_StubError`, `TestUpdateSourceFile_StubError`, `TestUpdateRegistryClause_StubError`, and the success-without-validation assumption of `TestPipeline_Execute_DryRun_Success` shall each be replaced by the assertions named in `plan.md` §C.2, not deleted without replacement. `TestLoadEvolutionLogs` shall assert `RuleID` and `ApprovedAt` in addition to `ID`.

- **REQ-CAA-015** (unwanted) — No test or verification command for this SPEC shall write the repository's real `zone-registry.md`, any real rule file under `.claude/rules/`, or the real `.moai/research/evolution-log.md`; every write shall target a `t.TempDir()` fixture.

## §E Constraints

### §E.1 Scope boundary (What, not How)

Function names, the shape of the rename seam, and backup naming are run-phase decisions guided by `plan.md`; this document fixes only observable behaviour.

### §E.2 Stability

- Five-layer gate order and every gate's semantics are unchanged.
- `acquireLock` / `releaseLock` behaviour is unchanged (see §F on the default lock path).
- The registry loader's parse rules (`internal/constitution/loader.go`) are unchanged; REQ-CAA-004 reuses them rather than adding a second parser.
- No new third-party dependency (`gopkg.in/yaml.v3` is already used).

### §E.3 Verification route

- Package-level behaviour is verified in `internal/constitution` tests with `t.TempDir()` fixtures and in-package test doubles (`fakeOversight`, the injectable `rateLimiter.now`, the rename seam of REQ-CAA-011).
- CLI-level behaviour is verified in `internal/cli` tests that call `runConstitutionAmend` directly. No built binary and no `HOME` override are used. **Correction to the dispatch premise:** the dispatch and verdict §7 call this a "home seam", but `runConstitutionAmend` and `Pipeline.Execute` read no home directory. The seam the test uses is the `projectDir` parameter pointed at `t.TempDir()`, with `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` cleared through `t.Setenv(…, "")`, because `resolveRegistryPath` (`internal/cli/constitution.go:144-155`) reads both before falling back to `projectDir`, and a Claude session exports `CLAUDE_PROJECT_DIR`.
- CLI-level coverage is dry-run only. A non-dry-run CLI test would need a stdin seam (`NewHumanOversight` hardcodes `os.Stdin`, `internal/constitution/human_oversight.go:20-24`) and would create the cwd-relative lock file inside the package directory (§F); the real apply path is verified at package level instead.

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

- The default lock path `.moai/research/.amendment.lock` (`pipeline.go:227`) is relative to the process working directory, not to `projectDir`. **Observed issue, recorded, not fixed here.** The atomicity design of REQ-CAA-010 does not need it: the lock serializes concurrent amenders, while atomicity concerns one amender's three writes. Its only effect on this SPEC is that it keeps the CLI-level test dry-run-only (§E.3), because a non-dry-run CLI test would drop a lock file into `internal/cli/.moai/research/`.

### Out of Scope — adjacent defects observed while planning

- `MarkRolledBack` / `rewriteEvolutionLog` rewrite the whole log in machine format. Once REQ-CAA-008 makes human entries readable, a rollback would re-serialize them lossily. It has **zero production callers** (`grep -rn 'MarkRolledBack(' --include='*.go' internal cmd pkg | grep -v _test.go` prints only its definition at `evolution_log.go:88`, tree `034d55c56`), so the hazard is latent. Not changed here.
- `runConstitutionAmend` loads the registry through `resolveRegistryPath` (environment first), while `Execute` loads `<projectDir>/…/zone-registry.md`. When the environment points elsewhere, the `--before` pre-check and the apply read different registries. Not changed here.
- The CLI reads `MOAI_CONSTITUTION_DRY_RUN == "true"` while SPEC-V3R2-CON-002 REQ-CON-002-031 says `=1`. Not changed here.
- `Execute` does not itself compare `proposal.Before` with the current clause; the CLI does (`constitution.go:527-530`). The apply step searches for the current clause (REQ-CAA-001), so a direct `Execute` caller with a stale `Before` would still record that stale value as `clause_before`. Not changed here — open question §G item 2.
- A new clause that already occurs in the source file leaves two occurrences after apply, which blocks every later amendment of that entry under REQ-CAA-002. Not ruled on — open question §G item 1.
- The human entry EVO-HRN-002 names `const_registry_entry: CONST-V3R2-153`, whose registry `file:` is `session-handoff.md`, while the entry's own `target_file` is `design/constitution.md`. REQ-CAA-008 maps the value as written; correcting the log is not this card's business. The rate limiter uses `RuleID` only in its rollback check, which this entry (no `rolled_back`) never triggers.
- Crash recovery across process death between renames (backups are in-process only).
- The template mirror of `zone-registry.md` under `internal/template/templates/`: an apply writes only the project registry.

## §G Open questions for the lead (not decided here)

1. Should the apply step reject a proposal whose new clause already occurs in the source file (it would leave two occurrences)?
2. Should `Execute` itself reject `proposal.Before` ≠ current clause, or is the CLI pre-check sufficient?
3. REQ-CAA-009 fail-closed: the Q3 ruling asked this SPEC to define how human-format entries map and what happens to missing fields. Failing closed on a recognized entry with no parseable timestamp is this draft's choice (an undercounting rate limiter is the defect Q3 targets). Confirm or override.

## §H Cross-References

- `.moai/reports/t659/verdict.md` — repro evidence (§2) and lead rulings (§7)
- `.moai/specs/SPEC-V3R2-CON-002/spec.md` — REQ-CON-002-004 (log fields), REQ-CON-002-011 (atomic apply), REQ-CON-002-031 (dry-run)
- `internal/constitution/pipeline.go`, `evolution_log.go`, `amendment.go`, `loader.go`, `rate_limiter.go`, `human_oversight.go`
- `internal/cli/constitution.go` — `newConstitutionAmendCmd`, `runConstitutionAmend`, `resolveRegistryPath`
- `internal/constitution/registry_sync_test.go` — existing literal-clause (no normalization) guard on the real registry, the read-only precedent for REQ-CAA-001's exact matching
