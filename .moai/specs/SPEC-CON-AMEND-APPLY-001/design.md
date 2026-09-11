# design.md — SPEC-CON-AMEND-APPLY-001

Tier L design artifact, added in revision 0.1.4 when the SPEC was raised from Tier M to Tier L (`.moai/reports/t659/verdict.md` §13.3). It records the architecture that the lead rulings already fixed: verdict §7 (Q1–Q5), §8 (G1–G5 and scope addition (a)), §9 (G6), §11 (G7), and the two agent extensions the lead accepted in §12.5. **No decision is introduced here.** Where the SPEC's encoding of a ruling adds a detail the verdict does not state word for word, the section says so and names the requirement that carries it.

Requirement wording lives in `spec.md` §D, verification in `acceptance.md`, milestone order in `plan.md` §F, and the codebase findings these decisions rest on in `research.md`. Function names, the exact shape of the two seams, the resolver's package home, and backup file naming are run-phase decisions (`spec.md` §E.1) and are not fixed here.

## §A Decision map

| § | Decision | Ruling | Requirements | Acceptance criteria | Mutants |
|---|---|---|---|---|---|
| §C | One containment check, applied at three sites | G6 option (ii), G7 option A | REQ-CAA-020, REQ-CAA-021 | AC-CAA-023, AC-CAA-024, AC-CAA-025 | M-20, M-21, M-22, M-23, M-24 |
| §D | One registry path resolver shared by the CLI and `Execute` | scope addition (a) | REQ-CAA-019 | AC-CAA-022 | M-19 |
| §E | Exact-once source replacement, with the `Before` and new-clause preconditions | Q1, G2, G3 | REQ-CAA-001, REQ-CAA-002, REQ-CAA-016, REQ-CAA-017 | AC-CAA-001, AC-CAA-002, AC-CAA-003, AC-CAA-019, AC-CAA-020 | M-1, M-2, M-16, M-17 |
| §F | One-line registry clause replacement with re-parse check | Q2 | REQ-CAA-003, REQ-CAA-004 | AC-CAA-004, AC-CAA-005 | M-3a, M-3b, M-4 |
| §G | Evolution-log tags, legacy-key reading, human-format parser, fail-closed error | Q3, G1 | REQ-CAA-005 … REQ-CAA-009 | AC-CAA-006 … AC-CAA-011, AC-CAA-018 | M-6, M-7, M-8a, M-8b, M-9, M-12, M-15 |
| §H | Backup, temporary write, ordered rename, restore; fault-injection seams | Q4, G4 | REQ-CAA-010, REQ-CAA-011, REQ-CAA-018 | AC-CAA-012, AC-CAA-013, AC-CAA-021 | M-5a, M-5b, M-5c, M-11a, M-18 |
| §I | Dry-run runs the validation steps and writes nothing | Q4 (last sentence) | REQ-CAA-012, REQ-CAA-013 | AC-CAA-014, AC-CAA-015 | M-10, M-11b, M-13 |
| §J | Verification seams and the approved CLI gap | §7 CLI slot, G5 | REQ-CAA-014, REQ-CAA-015 | AC-CAA-015, AC-CAA-016, AC-CAA-017 | M-14 |

Scope ruling Q5 (rejected-amendment logging and the SPEC-V3R2-CON-002 status correction are separate cards) shapes no design here; it is encoded in `spec.md` §F.

## §B Order of checks inside one `Execute` call

The sequence below restates what REQ-CAA-012, REQ-CAA-017, REQ-CAA-019, REQ-CAA-020, and REQ-CAA-021 already require, so the placement of each decision below can be read against it.

```
Execute(proposal, projectDir, dryRun)
  0  acquire lock (real mode only; behaviour unchanged, spec.md §E.2)
  1  resolve the registry path with the shared resolver            §D   scope addition (a)
  2  containment check on the registry path and on every
     entry's joined file: (at registry load), and on the
     evolution-log path (before Layer 1)                           §C   G6, G7
       refused -> load error naming the path; nothing written; lock released
  3  look up the rule; Before must equal the current clause        §E.1 G3
       mismatch -> error naming the rule ID; no gate runs; lock released
  4  Layers 1-5 (order and semantics unchanged, spec.md §E.2)
  5  pre-apply validations on the real bytes                       §E, §F  Q1, Q2, G2
       dry-run -> return the log entry or the validation error here; no write
  6  backups -> temporary writes -> rename source, registry, log   §H   Q4
       failure -> restore all three; a failed restore keeps the backups (G4)
  7  release lock
```

Steps 2 and 3 run in dry-run and real mode alike (REQ-CAA-017, REQ-CAA-020, REQ-CAA-021). Step 5 is the "validation half" of the apply step that dry-run executes (REQ-CAA-012).

## §C One containment check (G6 option (ii), G7 option A)

### C.1 What the check decides

A path is inside the project root only when all of the following hold, in this order (verdict §11 implementation rule, encoded in REQ-CAA-021):

1. The path is cleaned (`filepath.Clean`).
2. It is made absolute against the process working directory (`filepath.Abs`), so a relative environment value such as `../other` is judged by where it actually points.
3. Its symbolic links are resolved.
4. The result equals the likewise resolved `projectDir`, or continues from it past a path separator.

Two details come from the SPEC's encoding of the ruling, not from the verdict's words:

- **The separator boundary** (step 4). The verdict's second mutant describes `/root-evil` passing a plain `/root` prefix test (§11). REQ-CAA-021 states the boundary that rejects that shape, and M-22 variant (i) keeps it honest.
- **Resolving both sides, and the nearest existing ancestor.** The root is resolved as well as the candidate, so a root reached through a symbolic link (the macOS temporary directory is one) is not refused (M-24, plan.md R-7). A path that does not exist yet, such as an evolution log not yet created, is judged by resolving its nearest existing ancestor and appending the remaining components (REQ-CAA-021, AC-CAA-024 `in_root_control`, plan.md R-7).

### C.2 Where the check is applied

| Site | When | Value shapes covered | Ruling | Acceptance rows | Mutant |
|---|---|---|---|---|---|
| Registry path returned by the resolver (§D) | at registry load | absolute; relative value from `MOAI_CONSTITUTION_REGISTRY` or `CLAUDE_PROJECT_DIR` | G6 (ii); G7 item 1 | AC-CAA-023 `divergent_root_real`, `divergent_root_dry_run`; AC-CAA-024 `relative_env_escape` and its CLI case | M-20 |
| Every registry entry's `file:`, joined with `projectDir` when relative | at registry load, for every entry, not only the target | absolute; relative containing `..`; absolute sibling whose name begins with the root's name | G7 item 2 | AC-CAA-024 `absolute_file` (target and `non_target`), `dotdot_file`, `sibling_prefix_file` | M-21 (`file:` variant), M-22 (both variants) |
| Evolution-log path `<projectDir>/.moai/research/evolution-log.md` | before Layer 1 | a directory on the path that is a symbolic link out of the root | G7 item 3 | AC-CAA-024 `symlinked_log` | M-21 (log variant), M-23 |

The in-root controls — AC-CAA-023 `same_root_control`, AC-CAA-024 `in_root_control` (`plain`, `symlinked_root`), and AC-CAA-025 — show the refusals come from the escaping shapes, not from a check that refuses everything. M-24 kills the case where only the candidate side is resolved.

### C.3 One check, two callers

The CLI's registry validation and `Execute` call the same check (REQ-CAA-021). It replaces the loader's present refusal, which checks containment only for an absolute path (`internal/constitution/loader.go:80-88`; `research.md` §G). The CLI dry-run case of AC-CAA-024 was added by the plan-phase agent to verify the "same check" clause and was accepted by the lead as within the requirement (verdict §12.4 item 1, §12.5).

### C.4 Refusal shape

A refused path yields a load error that names the offending path, in dry-run and real mode alike, before Layer 1 and before any backup, temporary file, or write; the lock is released (REQ-CAA-020, REQ-CAA-021; verdict §11 "쓰기 전 적재 오류"). Because the refusal happens at registry load, it also reaches an escaping `file:` in an entry the amendment does not target and any command that loads the registry (`moai constitution list`, `guard`). This is intended (plan.md R-8), and it refuses nothing on the real registry: 0 of 101 `file:` values are absolute and 0 contain `..` (`research.md` §F, AC-CAA-025).

### C.5 Rejected alternatives

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| Make the source rule file and evolution-log paths follow the root the environment names (G6 option (i)) | verdict §9 | widens the set of directory trees an amendment can write |
| Narrow REQ-CAA-020 to an absolute registry path and record relative registry paths and escaping `file:` values as observed only (G7 option B) | verdict §11 | a declared invariant would be half enforced; once writing is on, an out-of-root write is a path-traversal defect |
| Keep the loader's absolute-only check | verdict §10.3, §11 | leaves the relative registry path and the `file:` join unchecked (`research.md` §G) |
| Plain string prefix test without a separator boundary | verdict §11 (mutant (2)); split into two variants at §12.4 item 2, accepted §12.5 | `/root-evil` passes as inside `/root`; M-22 (i) |
| Prefix test on the concatenated, uncleaned candidate | same | `..` survives into the comparison; M-22 (ii). `filepath.Join` already cleans, so this variant is the one that exposes a missing `Clean` on the join path |
| No symbolic-link resolution, or resolution on the candidate only | verdict §11 ("심볼릭 링크 해석"), encoded in REQ-CAA-021 | a symlinked directory escapes (M-23); a symlinked root is refused (M-24) |

### C.6 Residual risk and gaps

- A symbolic link swapped between the check and the write is not a requirement of this SPEC; whoever can rewrite links inside the root already controls its files (plan.md R-9).
- Where the platform refuses to create a symbolic link, the symlink rows call `t.Skip`; the skip is a Gap and M-23 and M-24 are unobserved on that platform (verdict §12.5).

## §D One registry path resolver (scope addition (a))

### D.1 Decision

The CLI's registry validation and `Execute` obtain the registry path from one resolver whose precedence is the one `resolveRegistryPath` has today: a non-empty `MOAI_CONSTITUTION_REGISTRY`, then `<CLAUDE_PROJECT_DIR>/.claude/rules/moai/core/zone-registry.md`, then `<projectDir>/.claude/rules/moai/core/zone-registry.md` (REQ-CAA-019; current code `internal/cli/constitution.go:144-154`, `research.md` §H). `Execute` stops constructing its own join (`internal/constitution/pipeline.go:66`).

The resolver reads the environment names from constants rather than literals: `config.EnvClaudeProjectDir` and the existing `MOAI_CONSTITUTION_REGISTRY` constant, moved or re-exported (plan.md §D). Its package home is a run-phase choice; plan.md M2 gives `internal/constitution` as an example.

### D.2 How it fits with the containment check

The resolver decides which path is chosen; the containment check (§C) decides whether that path is admitted (REQ-CAA-020 rationale). An override inside `projectDir` is chosen and admitted (AC-CAA-022). A `CLAUDE_PROJECT_DIR` naming another tree can only yield a refused path (AC-CAA-023). In both, nothing under the other tree is written.

### D.3 Rejected alternative

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| `Execute` keeps its own `projectDir` join while the CLI keeps the environment precedence | verdict §8 scope addition (a) | once this card turns writing on, the CLI validating one registry while `Execute` writes another is not a latent defect but a write accident; M-19 |

## §E Exact-once source replacement (Q1, G2, G3)

### E.1 `Before` must equal the current clause, checked inside `Execute` (G3)

`Execute` compares the proposal's `Before` with the target entry's current clause as an exact byte sequence, before Layer 1, in both modes, and returns an error naming the rule ID on a mismatch (REQ-CAA-017). The CLI's own `--before` comparison stays as early guidance and does not replace it.

- Rejected: relying on the CLI check alone (verdict §8 G3, defence in depth). M-17 keeps the CLI check and removes only the `Execute` check; AC-CAA-020 calls `Execute` directly and stays RED.

### E.2 The new clause must be absent from the source file (G2)

Before any write, the new clause must occur 0 times in the source rule file, using the same exact matching as E.3 (REQ-CAA-016). plan.md M3 orders this check before the current-clause count; both errors name the path and the count.

- Consequence stated in the requirement: a new clause contained inside the current clause (for example, a sentence shortened by removing its tail) also counts as an occurrence and is rejected (plan.md R-5).
- Rejected: allowing a pre-existing occurrence (verdict §8 G2); M-16.

### E.3 Current clause exactly once in the whole file (Q1)

The current clause is replaced only when it occurs as an exact byte sequence exactly once in the whole file; 0 or 2 or more occurrences fail before any write, with an error naming the path and the count (REQ-CAA-001, REQ-CAA-002). The rule is attainable on the real corpus: 97 of 97 live registry clauses occur exactly once in their file (`research.md` §E).

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| Whitespace-normalized matching (what the existing validator does) | verdict §7 Q1 | silent mis-match risk; M-2 |
| Narrowing the search to the section named by `anchor` | verdict §7 Q1 | unnecessary: the 97 live clauses already occur once in the whole file. There is also no reusable anchor-range rule: `SentinelAnchorNotFound` is defined and unused (`research.md` §I) |
| Replace every occurrence, or the first | Q1 exactly-once rule | M-1 |

### E.4 Shape

The source and registry transforms are pure functions from (bytes, current clause, new clause) to new bytes or an error, with no file I/O, so dry-run (§I) and the real apply (§H) share one validation path (plan.md M3).

## §F One-line registry clause replacement with re-parse check (Q2)

### F.1 Decision

The apply step rewrites only the target entry's `clause:` line inside the registry's yaml fence, as a string, and leaves every other line byte-identical, including comments, entry order, and the quoting of other entries (REQ-CAA-003). Immediately afterwards, before any rename, it parses the whole candidate content with the loader's own fence extraction and yaml decoding and compares the entry count and the target entry's decoded clause (REQ-CAA-004).

Encoding as plan.md M3 fixes it:

- Locate the fence the way the loader does (`extractYAMLFence`).
- Locate the target entry by its `- id: <RuleID>` line, and that entry's own `clause:` line before the next `- id:`. Exactly one such line is required.
- Emit the new clause as a single-line double-quoted yaml scalar, escaping `\` and `"`.
- A new clause containing a newline cannot be one line and is rejected through the REQ-CAA-004 error path.
- Reuse the loader's fence and `rawEntry` decoding; no second parser (plan.md §D).

Basis: the registry has one yaml fence holding 101 entries, and every clause is a one-line double-quoted string (`research.md` §E).

### F.2 Rejected alternatives

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| Re-serialize the parsed registry | verdict §7 Q2 | would change the formatting of the whole 101-entry file; M-3a (line diff ≠ 1) |
| Interpolate the new clause into quotes without escaping | Q2 round-trip check | M-3b |
| Write without re-parsing | verdict §7 Q2 ("쓰기 직후 재파싱 검증") | a continuation line, a missing `clause:` line, or a newline clause corrupts the registry; M-4, AC-CAA-005 |

## §G Evolution log: tags, legacy keys, human format (Q3, G1)

### G.1 Writer

`AmendmentLog` carries explicit snake_case yaml tags (`id`, `rule_id`, `zone_before`, …), and zones are written as the names the registry uses (`Frozen`, `Evolvable`), never as integers (REQ-CAA-005; verdict §7 Q3 "zone 표현은 등록부 형식을 따른다"). plan.md M1 places the zone rendering on the `Zone` type: marshal to the name; unmarshal from the name or the legacy integer.

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| Keep the untagged concatenated-lowercase keys (verdict §6 Q3 option (b)) | verdict §7 Q3 | a snake_case log, the notation of SPEC-V3R2-CON-002 and the test fixture, reads back with empty `RuleID` and zero `ApprovedAt` (`research.md` §C) |
| Integer zones | verdict §7 Q3 | M-9 |

### G.2 Reader compatibility

The reader accepts both the snake_case keys and the legacy concatenated keys the untagged writer produced, and a zone as a name or as the legacy integer (`0` = Frozen, `1` = Evolvable). When one entry carries both key forms for a field, the snake_case value wins (REQ-CAA-006; verdict §7 Q3 "읽기는 기존 소문자 연결형 키도 호환").

- Rejected: dropping the legacy aliases (M-6); dropping the snake_case tag on the read side (M-7).

### G.3 Candidate blocks

The reader considers two kinds of candidate block (plan.md M1):

- every fenced yaml block;
- `---`-delimited segments, scanned line-anchored, advancing one line whenever a segment is not an entry, never pairwise.

A `---`-delimited entry is recognized whenever its content decodes to a mapping with a non-empty top-level `id`, however many `---` lines precede it (REQ-CAA-007).

- Rejected: today's pairwise split on the substring `---` (`internal/constitution/evolution_log.go:29`). On the real log it meets `|---|` table rows and horizontal rules and yields 0 entries (`research.md` §D); M-12.

### G.4 Human-format entries

A fenced yaml block whose decoded top-level mapping has a non-empty `id` — the format of the human-authored `## EVO-HRN-002` — is returned as an entry. Its fields map through the table in REQ-CAA-008: `const_registry_entry` feeds `RuleID`, `timestamp` feeds `ApprovedAt` (date-only means 00:00:00 UTC), `zone` feeds both zone fields, and `approver` feeds `ApprovedBy`. Keys with no entry field are ignored, not rejected.

- Decision source: verdict §7 Q3 puts making the parser read the real file, human entries included, inside this card. A gate whose log is written but unreadable by the rate limiter is powerless.
- Rejected: ignoring fenced blocks (M-8a).

### G.5 Fail closed, and say where (G1)

When a recognized entry has no approval timestamp that parses, the reader returns an error rather than leaving the entry out of the 7-day window and the cooldown (REQ-CAA-009; verdict §8 G1 "Frozen 게이트는 편의보다 안전"). The error contains the log file path, the 1-based line number in the file, the key, and the entry's `id`. When a timestamp key is present but unparseable, the line and key are that key's; when none is present, the key is `approved_at` and the line is that of `id`.

plan.md M1 obtains the line by decoding each candidate into a `yaml.Node` and adding the block's starting line in the file, so a block-relative line is never reported.

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| Skip an unparseable entry, as today's reader does with an unmarshal failure | verdict §8 G1 | the rate limiter would silently under-count; M-8b |
| An error without line, without key, or with the block-relative line | verdict §8 G1 | a human cannot locate the entry to fix; M-15 (three variants) |

Risk carried: a malformed human entry blocks every amendment until fixed (plan.md R-2), accepted by the lead under G1.

### G.6 The log on the apply path

The new log content is the pre-apply bytes followed by the one new entry, written through the temporary-file-and-rename sequence of §H. This replaces the `O_APPEND` writer on this path (`evolution_log.go:67`) and preserves every pre-existing byte, human entries included (REQ-CAA-010).

`MarkRolledBack` and `rewriteEvolutionLog` are not changed. Their lossy rewrite of human entries is a follow-up candidate (`spec.md` §F; `research.md` §L).

## §H Three-file atomic apply (Q4, G4)

### H.1 Sequence

The apply step runs these steps in order (REQ-CAA-010; verdict §7 Q4):

1. Complete every pre-apply validation (§E, §F).
2. Write a backup of the pre-apply bytes of each of the three files, in that file's own directory. A file that does not yet exist is recorded as absent and gets no backup file.
3. Write each file's new content to a temporary file in that file's own directory.
4. Rename the temporary files into place in the order source rule file, registry, evolution log.
5. On success, remove the temporary files and the backups.

**On failure after the first backup** — a backup, a temporary write, or any rename — restore all three files from the backups, remove a file recorded as absent, remove the temporary files and backups once the restore completes, and return an error naming the failed step.

### H.2 Why backups as well as renames

A rename is atomic per file only; atomicity across three files needs further design (verdict §6 Q4). The ruling combines the two approaches the question offered: temporary write then rename for each file, plus backups and a full restore to cover the span between the first and the last rename (verdict §7 Q4).

- Order: source, registry, log (Q4; AC-CAA-013; M-5c).
- A log recorded absent must be removed on restore, not left half-created (AC-CAA-012 `third_rename_log_absent`; M-5b).
- Removing the restore call is M-5a.

### H.3 A failed restore keeps the evidence (G4)

When restoring any file fails, the apply step deletes no backup, leaves every backup holding its pre-apply bytes, removes the temporary files, and returns an error containing every backup path and the failed restore step (REQ-CAA-018; verdict §8 G4).

- Rejected: deleting the backups, or omitting their paths from the error (M-18, two variants).
- There is no automatic recovery on a later run; retained backups are left for a human (`spec.md` §F).

### H.4 Fault-injection seams

Q4 requires fault-injection tests for the second and third renames; G4 adds one restore-failure test. REQ-CAA-011 fixes two independent replaceable per-pipeline operations:

| Seam | Default | Used by |
|---|---|---|
| Forward rename | `os.Rename` | AC-CAA-012 (fail call N = 2, then N = 3), AC-CAA-013 (recording) |
| Restore write | write the backup bytes to the target, or remove the target when recorded absent | AC-CAA-021 (fail first call) |

- Both are struct fields on `Pipeline`, not package variables, so parallel tests cannot race on them (plan.md M4).
- The restore path never routes through the forward-rename seam, so a rename fault cannot also break the restore.
- Every fault-injection test asserts the seam's call count alongside byte identity. An injector that never fires and a correct restore otherwise look identical (plan.md §H).

### H.5 Placement and naming of temporary and backup files

Temporary and backup files are created in the target's own directory, so each rename stays on one filesystem. They carry a non-`.md` suffix, so a leftover inside `.claude/rules/**` is never loaded as a rule (plan.md §D, R-3). The exact names are run-phase; leftover files after success are M-11a.

### H.6 Boundaries

- Crash recovery across process death between renames is out of scope (`spec.md` §F).
- The single-writer lock is unchanged, and its cwd-relative default path is a recorded observation, not fixed. The lock serializes concurrent amenders; atomicity concerns one amender's three writes (`spec.md` §F).

## §I Dry-run runs the validation steps (Q4)

While `Execute` runs in dry-run mode it performs every pre-apply validation against the real bytes at the resolved paths and returns the same error a real apply would. It creates, modifies, renames, or removes no file — no backup, no temporary file, no lock file (REQ-CAA-012; verdict §7 Q4 "dry-run 은 적용 함수의 검증 단계…까지 실제로 실행"). The containment check and the `Before` check (steps 2 and 3 of §B) run in dry-run as well.

The CLI's `--dry-run` path returns the pipeline's error and does not print its success line (REQ-CAA-013).

| Alternative | Rejected by | Reason recorded |
|---|---|---|
| Today's dry-run, which returns a log entry without calling any apply function (`pipeline.go:133-137`) | verdict §7 Q4 | a dry-run cannot surface the failure a real apply would hit; M-10 |
| A dry-run that performs the real apply | REQ-CAA-012 | writes; M-11b |
| A CLI that ignores the pipeline error in dry-run | REQ-CAA-013 | prints success on a failing fixture; M-13 |

Consequence: fixtures pointing at a non-existent `dummy.md` must change, and the affected tests are updated rather than weakened (plan.md §C.2, R-1).

## §J Verification seams (§7 CLI slot, G5)

- **No binary, no home override.** CLI-level behaviour is verified by calling `runConstitutionAmend` directly in `internal/cli` tests (verdict §7). The seam is the `projectDir` parameter pointed at `t.TempDir()`, not a home directory; the verdict's "home seam" wording is corrected in `spec.md` §E.3.
- **Environment set by the test.** Every test reaching the resolver sets `MOAI_CONSTITUTION_REGISTRY` and `CLAUDE_PROJECT_DIR` itself (REQ-CAA-015). AC-CAA-017 runs the package once with `CLAUDE_PROJECT_DIR` exported as the repository root to show the tests do not depend on the shell's scrub; M-14.
- **Approved gap (G5).** CLI tests cover dry-run only. The non-dry-run CLI path would need a stdin seam and would create the cwd-relative lock inside the package directory; the apply itself is covered at `Execute` level (`spec.md` §E.4).
- **Stub tests are replaced, not deleted** (REQ-CAA-014; plan.md §C.2; AC-CAA-016).

## §K Left to run-phase

- Function and type names, including the containment check and the resolver.
- The resolver's package home.
- The concrete signatures of the two seams.
- Temporary and backup file names beyond "same directory, non-`.md` suffix".
- Final test names; `acceptance.md` gives proposals.

## §L Cross-references

- `spec.md` §C (ruling map), §D (requirements), §E (constraints and gaps), §F (exclusions)
- `plan.md` §C.2 (stub-test disposition), §D (constraints), §F (milestones), §G (risks), §H (anti-patterns)
- `acceptance.md` §D (AC matrix), §D.2 (mutants)
- `research.md` — the findings cited above
- `.moai/reports/t659/verdict.md` §6 (design questions), §7, §8, §9, §11 (rulings), §12.5 (accepted extensions), §13.3 (Tier L)
