# acceptance.md — SPEC-MEMORY-FOLD-BUDGET-001

Verification layer. Each criterion is an `AC-MFB-NNN` Given/When/Then, binary-testable, mapped to its requirement. The requirements themselves are the GEARS entries in `spec.md` §2; nothing here restates them as requirements. Revision 0.3.0 (plan delta for audit iteration 2; the SessionStart half is split out, so 12 requirements and 13 criteria remain).

## 0. Pins, provenance, conventions

- **Document-level tree pin: `2f492df19`.** It binds every RED-now cell below that carries no pin of its own. The plan commits (`73c4ab646`, then `0dfae6d4a`) descend from `2f492df19` and touch only `.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/**`; no Go source differs between the pinned tree and the plan commits. The cells were re-executed for revision 0.3.0 on tree HEAD `0dfae6d4a` with the stamped binary below; observed output equals what the pinned tree produces (each ledger cell carries a `re-run` line).
- **Binary provenance.** Every `./bin/moai` command was run with a binary built from this tree and invoked by path: `go build -buildvcs=false -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=2f492df19 -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-10-04" -o bin/moai ./cmd/moai`; `./bin/moai version` prints commit `2f492df19`. A plain `go build` in this worktree stamped a different checkout's revision (`c8f245c2c9a5`, not an ancestor of HEAD), so the stamp was set explicitly; the commit is therefore declared, not derived.
- **Real store off limits.** No command in this document, and no green-path command, reads or writes the operator's real memory directory (`spec.md` C-1). Green-path tests copy the fixture or generate stores under `t.TempDir()`, set `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR` and `MOAI_HOME` to temporary directories, and fail if a resolved store lies outside the temporary root. Scratch stores used for design validation (§5) live under the session scratch directory and are always passed with an explicit `--dir`.
- **Commands are single invocations** run from the repository (tree) root, read-only, with the exit code recorded as its own field. Stdout is verbatim. A table cell never carries a command; the ledger below does, by id. One file per `grep -c` command: the shell used here replaces `grep` with a function whose multi-file output order is not stable.
- **Green-path rule.** Every `go test -run` selector is anchored `^…$` on each alternation branch, and every green cell is read together with the swept count: `go test -list '<same pattern>' <package>` must list exactly as many names as the pattern has branches. A green with an empty sweep is not a pass. Test names below are the run-phase's to finalize; the binding is the behavior and the swept-count rule.
- **Release-blocking** = carries RED-now (ledger id) and a green path. **Regression-guard** = guards behavior that must not change, or has a RED that cannot be re-executed on the pinned tree; never recorded as an adopted gate.
- **Run-phase RED record.** For every Go-test criterion the run phase also records, before GREEN, the verbatim failing output of the new test itself (`tdd-result-contract.md` `EXPECTED_RED`) in `progress.md` §E.2. The plan-phase RED-now below proves the feature is absent; it does not replace that record.

## 1. Fixture and plan-phase baseline

`fixtures/store-A/` is a synthetic store (one `MEMORY.md`, one archive index `project_card_archive_2026_10.md` carrying three links, 11 further topic files and one orphan; 13 topic files besides `MEMORY.md`). The fixture is **frozen**: this revision changes nothing under it, and the tests assert its file list, sizes and the SHA-256 of the two index files before using it, so a later edit cannot silently invalidate the numbers below (the fixture directory is outside the plan-artifact hash subject set).

| File | Bytes |
|---|---|
| `MEMORY.md` | 1155 |
| `project_card_archive_2026_10.md` | 373 |
| `feedback_alpha.md` | 160 |
| `feedback_beta.md` | 157 |
| `feedback_delta_note.md` | 143 |
| `feedback_hangul.md` | 166 |
| `feedback_orphan.md` | 154 |
| `feedback_queue_jump_queue.md` | 198 |
| `feedback_verify.md` | 186 |
| `project_card_t8998_older.md` | 163 |
| `project_card_t8999_prior.md` | 163 |
| `project_card_t9000_zero.md` | 161 |
| `project_card_t9001_alpha.md` | 173 |
| `project_card_t9005_epsilon.md` | 188 |

SHA-256 (measured with `shasum -a 256`, this run, unchanged from 0.2.0): `MEMORY.md` = `252722b91fb9b4f151bc30434656d0f59b29458490fb591ea16fa6ee1592c3d9`; `project_card_archive_2026_10.md` = `c353c3c9adc3255c4a15064deb017ff59ad0052edea6df4232ae336418891105`.

Measured on the committed fixture (`wc -c`, `wc -m`, and the reference oracle in the scratch directory — not part of the tree):

| Quantity | Value |
|---|---|
| `MEMORY.md` bytes / characters / lines (doctor counting) | 1,155 / 1,095 / 17 |
| index set I(S) (doctor rule, threshold 3) | `MEMORY.md` and `project_card_archive_2026_10.md` (3 resolved links) |
| \|R\| reachable store files (§1.5 of the spec) | 11 |
| \|T\| distinct full link targets over the index set | 14 (11 in `MEMORY.md`, 3 in the archive index) |
| distinct targets of `MEMORY.md` by full text / by base name | 11 / 10 (two repo-relative paths share the base `verdict.md`) |
| doctor findings at the pinned tree | 2 `MEMORY_ORPHAN_NOT_INDEXED`, 2 `MEMORY_DANGLING_INDEX_LINK` (one names only `verdict.md`) |

Fixture classification per card id (the expected plan of `fold --card <id>`): `t9001` STRONG (title-leading, `project_card_t9001_alpha.md`) plus one MENTION (the discipline line "first recorded on t9001"); `t9002` STRONG (target `.moai/reports/t9002/verdict.md`, a path segment); `t9006` STRONG (same shape); `t9004` AMBIGUOUS (title only); `t9005` AMBIGUOUS (target only); `t90` and `t9999` match nothing.

## 2. Evidence ledger (RED-now observations, tree `2f492df19`)

```
E1
command: ./bin/moai memory fold --card t9001 --dir .moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A
stdout:  (empty — 0 bytes)
exit:    1
tree:    2f492df19
red-because: the fold verb does not exist; the flag --card is rejected before any store is read
note:    stderr, which is not the decision channel, reads: Unknown flag: --card.
re-run:  0.3.0, tree 0dfae6d4a, binary commit 2f492df19 — stdout 0 bytes (redirected to a file, `wc -c` = 0), stderr 178 bytes, exit 1
```

```
E2
command: ./bin/moai memory doctor --json --dir .moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A
stdout:
[{"store":{"dir":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A","origin":"--dir"},"exists":true,"topic_files":13,"cap":50,"index_lines":17,"findings":[{"Code":"MEMORY_ORPHAN_NOT_INDEXED","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/feedback_orphan.md","Detail":"feedback_orphan.md is linked from no index — a session loads an index, not the directory, so this memory is never recalled"},{"Code":"MEMORY_ORPHAN_NOT_INDEXED","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/feedback_queue_jump_queue.md","Detail":"feedback_queue_jump_queue.md is linked from no index — a session loads an index, not the directory, so this memory is never recalled"},{"Code":"MEMORY_DANGLING_INDEX_LINK","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/MEMORY.md","Detail":"index links verdict.md but no such file exists"},{"Code":"MEMORY_DANGLING_INDEX_LINK","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/MEMORY.md","Detail":"index links feedback_queue_jump_window.md but no such file exists"}]}]
exit:    0
tree:    2f492df19
red-because: the report has no index_bytes, index_chars, index_loaded_chars or index_link_targets key; the two distinct repo-relative targets collapse into ONE dangling finding naming only the base verdict.md, and no MEMORY_REPO_RELATIVE_LINK finding exists
also-serves: positive control for E1 (the same binary and fixture, an existing verb, exit 0); regression evidence that "cap":50 and "topic_files":13 are today's values
re-run:  0.3.0, tree 0dfae6d4a, binary commit 2f492df19 — stdout (1330 bytes, redirected to a file) byte-identical to the stdout line above (`cmp` exit 0), exit 0
```

```
E3
command: ./bin/moai memory doctor --byte-cap 1300 --dir .moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A
stdout:  (empty — 0 bytes)
exit:    1
tree:    2f492df19
red-because: the --byte-cap flag does not exist; 1,155 bytes is 88.8 % of 1,300, so a working budget check would warn here
note:    stderr reads: Unknown flag: --byte-cap.
re-run:  0.3.0, tree 0dfae6d4a, binary commit 2f492df19 — stdout 0 bytes (redirected to a file, `wc -c` = 0), exit 1
```

```
E6a-todo
command: grep -c "foldClosedCardMemory" internal/cli/todo.go
stdout:  0
exit:    1
tree:    2f492df19
red-because: `todo done` / `gtd done` does not call a memory fold (identifier-keyed; the G-WIRE tests decide the pass)
re-run:  0.3.0, tree 0dfae6d4a — stdout 0, exit 1
```

```
E6a-autodone
command: grep -c "foldClosedCardMemory" internal/cli/todo_autodone.go
stdout:  0
exit:    1
tree:    2f492df19
red-because: `todo auto-done` does not call a memory fold
re-run:  0.3.0, tree 0dfae6d4a — stdout 0, exit 1
```

```
E6a-auto
command: grep -c "foldClosedCardMemory" internal/cli/todo_auto.go
stdout:  0
exit:    1
tree:    2f492df19
red-because: the `todo --auto` cycle does not call a memory fold
re-run:  0.3.0, tree 0dfae6d4a — stdout 0, exit 1
```

```
E6b-todo  (positive control for E6a-todo — the close path exists)
command: grep -c "ArchiveCard(" internal/cli/todo.go
stdout:  1
exit:    0
tree:    2f492df19
re-run:  0.3.0, tree 0dfae6d4a — stdout 1, exit 0
```

```
E6b-autodone  (positive control for E6a-autodone)
command: grep -c "ArchiveCard(" internal/cli/todo_autodone.go
stdout:  1
exit:    0
tree:    2f492df19
re-run:  0.3.0, tree 0dfae6d4a — stdout 1, exit 0
```

```
E6b-auto  (positive control for E6a-auto)
command: grep -c "ArchiveCard(" internal/cli/todo_auto.go
stdout:  1
exit:    0
tree:    2f492df19
re-run:  0.3.0, tree 0dfae6d4a — stdout 1, exit 0
```

```
E8a
command: grep -c "DefaultMemoryFoldOnDoneBound" internal/config/defaults.go
stdout:  0
exit:    1
tree:    2f492df19
red-because: the execution-bound constant of REQ-MFB-007 does not exist, so no value or ceiling can be asserted on it (identifier-keyed: it proves absence of the constant, not of the behavior; the G-WIRE ceiling and production-bound cells decide the pass)
re-run:  0.3.0, tree 0dfae6d4a — stdout 0, exit 1
```

```
E8b  (positive control for E8a — the file and its neighboring one-shot-process bound exist)
command: grep -c "DefaultHookAsyncJoinTimeout" internal/config/defaults.go
stdout:  2
exit:    0
tree:    2f492df19
re-run:  0.3.0, tree 0dfae6d4a — stdout 2, exit 0
```

Removed cells. The v0.2.0 cells E5a/E5b (SessionStart wiring) and E7a/E7b (hook `TestMain` home isolation) left with the SessionStart half to the follow-up card (`spec.md` §4); they were not re-run in this revision because nothing in this SPEC cites them. The v0.1.0 cell E4 for `./bin/moai memory relink --dir <fixture>` moved to the other follow-up card with the verb; its recorded reason was imprecise (the stderr `Unknown flag: --dir.` comes from the parent `memory` command lacking the flag; without `--dir` the message is `Unknown command "relink" for "moai memory".`), which the follow-up should record.

### 2.1 Green-path commands (run-phase; selectors anchored, each paired with its swept count)

Each id below is a pair: the `-run` command, and the `-list` command with the same pattern. Green means exit 0, every listed name prints a PASS line, and the `-list` output names exactly as many tests as the pattern has branches. All of them run only against temporary directories and `fixtures/store-A/` read-only.

```
G-CORE      package ./internal/hook/memo/taxonomy — 11 branches (AC-MFB-013, AC-MFB-012 unit side; M1)
go test -run '^TestReach_CorrectFoldPasses$|^TestReach_DeleteMutantFails$|^TestReach_DropLinkMutantFails$|^TestReach_UnlinkedFileMutantFails$|^TestReach_TruncateMutantFails$|^TestReach_NonIndexTopicMutantFails$|^TestReach_LinkFreeLineMutantFails$|^TestReach_UnlinkedHubMutantFails$|^TestReach_StrayEditMutantFails$|^TestReach_ThresholdBoundary$|^TestReach_SizeAloneNeverPasses$' -count=1 -v ./internal/hook/memo/taxonomy
go test -list '^TestReach_CorrectFoldPasses$|^TestReach_DeleteMutantFails$|^TestReach_DropLinkMutantFails$|^TestReach_UnlinkedFileMutantFails$|^TestReach_TruncateMutantFails$|^TestReach_NonIndexTopicMutantFails$|^TestReach_LinkFreeLineMutantFails$|^TestReach_UnlinkedHubMutantFails$|^TestReach_StrayEditMutantFails$|^TestReach_ThresholdBoundary$|^TestReach_SizeAloneNeverPasses$' ./internal/hook/memo/taxonomy

G-FOLD      package ./internal/cli — 8 branches (AC-MFB-001 … AC-MFB-007; M3)
go test -run '^TestMemoryFold_DryRunWritesNothing$|^TestMemoryFold_Classification$|^TestMemoryFold_ReachabilityPreserved$|^TestMemoryFold_VerbatimFiling$|^TestMemoryFold_ArchiveSelection$|^TestMemoryFold_Idempotent$|^TestMemoryFold_EdgeInputs$|^TestMemoryFold_ApplyOrderAndAbort$' -count=1 -v ./internal/cli
go test -list '^TestMemoryFold_DryRunWritesNothing$|^TestMemoryFold_Classification$|^TestMemoryFold_ReachabilityPreserved$|^TestMemoryFold_VerbatimFiling$|^TestMemoryFold_ArchiveSelection$|^TestMemoryFold_Idempotent$|^TestMemoryFold_EdgeInputs$|^TestMemoryFold_ApplyOrderAndAbort$' ./internal/cli

G-DOCTOR    package ./internal/cli — 5 branches (AC-MFB-009, -010, -011, -012; M2)
go test -run '^TestMemoryDoctor_Measures$|^TestMemoryDoctor_TopicCapUnchanged$|^TestMemoryDoctor_BudgetBoundaries$|^TestMemoryDoctor_BytesProxyWarns$|^TestMemoryDoctor_LinkClasses$' -count=1 -v ./internal/cli
go test -list '^TestMemoryDoctor_Measures$|^TestMemoryDoctor_TopicCapUnchanged$|^TestMemoryDoctor_BudgetBoundaries$|^TestMemoryDoctor_BytesProxyWarns$|^TestMemoryDoctor_LinkClasses$' ./internal/cli

G-WIRE      package ./internal/cli — 10 branches (AC-MFB-008; M4)
go test -run '^TestMemoryFoldOnDone_DisabledDifferential$|^TestMemoryFoldOnDone_EnabledFolds$|^TestMemoryFoldOnDone_FailOpen$|^TestMemoryFoldOnDone_SeededPanic$|^TestMemoryFoldOnDone_BlockedRead$|^TestMemoryFoldOnDone_RunsAfterQueueWrite$|^TestMemoryFoldOnDone_ThreeClosePaths$|^TestMemoryFoldOnDone_DisabledNeverOpensStore$|^TestMemoryFoldOnDone_ProductionBoundEffective$|^TestMemoryFoldOnDone_BoundConstantCeiling$' -count=1 -v ./internal/cli
go test -list '^TestMemoryFoldOnDone_DisabledDifferential$|^TestMemoryFoldOnDone_EnabledFolds$|^TestMemoryFoldOnDone_FailOpen$|^TestMemoryFoldOnDone_SeededPanic$|^TestMemoryFoldOnDone_BlockedRead$|^TestMemoryFoldOnDone_RunsAfterQueueWrite$|^TestMemoryFoldOnDone_ThreeClosePaths$|^TestMemoryFoldOnDone_DisabledNeverOpensStore$|^TestMemoryFoldOnDone_ProductionBoundEffective$|^TestMemoryFoldOnDone_BoundConstantCeiling$' ./internal/cli
```

The `internal/cli` package suite is a heavy run (`.claude/rules/local/gitflow-lane-protocol.md` §8); these selectors are the scoped form, and the full package waits for the CI run after the push by the leader.

## 3. Acceptance criteria

### AC-MFB-001 — Fold previews by default, writes nothing, names its store (maps REQ-MFB-001)

**Given** a copy of `fixtures/store-A` under a temporary directory **When** the operator runs `moai memory fold --card t9001 --dir <copy>` without `--yes`, and again with `--json` **Then** the first stdout line is `store: <copy> (--dir)`, the output lists the one STRONG line it would file and the archive index it would file it into, `--json` is one valid JSON object carrying the same plan, the SHA-256 of every file of the copy is identical before and after, no temporary file remains, and the exit code is 0; **and** `--card 9001` is accepted as `t9001`, while `--card "t9001;x"` and `--card t` are refused with a non-zero exit before the store is read, stdout empty.

- RED-now: E1 (verb absent). Green path: M3 — G-FOLD; stdout as above, exit 0, hash-equality assertion.
- Mutant probe: a fold that applies without `--yes` is killed by the SHA-256 equality assertion; a fold that ignores `--dir` and resolves another store is killed by the `store:` line assertion plus the temporary-root guard. Paper mutants (code absent).
- Class: release-blocking.

### AC-MFB-002 — Only STRONG lines are planned; AMBIGUOUS and MENTION lines are kept with a reason (maps REQ-MFB-002)

**Given** a copy of `fixtures/store-A` **When** the preview runs for each of `t9001`, `t9002`, `t9006`, `t9004`, `t9005`, `t90` and `t9999` **Then** `t9001`, `t9002` and `t9006` plan exactly one STRONG line each (`t9001` additionally lists the discipline line as a MENTION, kept); `t9004` plans nothing and lists its line as AMBIGUOUS (title only); `t9005` plans nothing and lists its line as AMBIGUOUS (target only); `t90` and `t9999` plan nothing and report `no line`; a generated variant whose link target names two different card ids is AMBIGUOUS.

- RED-now: E1. Green path: M3 — G-FOLD.
- Mutant probe: a substring or prefix matcher makes `t90` select five lines — killed by the `t90` cell; a disjunctive matcher moves the `t9004` and `t9005` lines — killed by the AMBIGUOUS-kept assertions; a title-only matcher moves the `t9004` line — killed likewise. Paper mutants.
- Class: release-blocking.

### AC-MFB-003 — PRIMARY: fold preserves reachability and equals the plan exactly, never judged by size (maps REQ-MFB-003, REQ-MFB-004, REQ-MFB-012)

**Given** a copy of `fixtures/store-A` and the reference snapshot of its R, T and index set **When** the operator folds `t9001`, `t9002` and `t9006` with `--yes`, each on its own copy and then sequentially on one copy, and again on a generated variant in which the STRONG line is a grouped line carrying two link targets **Then** for every resulting store: (a) R(after) ⊇ R(before), (b) T(after) ⊇ T(before) counted by full target text across the whole index set, (c) every line removed from `MEMORY.md` has an equal-target line in the archive index, **and (d) the store equals the plan exactly**: (d1) `MEMORY.md` equals its prior bytes minus exactly the planned lines — the blockquote line, both headings and every other line of the fixture still present, byte for byte, in order; (d2) the archive index equals its prior bytes plus exactly the planned lines verbatim at the end; (d3) the SHA-256 of every other file of the copy is unchanged and the file list is identical; the grouped variant keeps both of its targets; the doctor's linkage audit run in-process on every resulting store reports no `MEMORY_ORPHAN_NOT_INDEXED` finding that the store did not carry before the fold (the checker and the doctor agree); and the checker reports the byte and line decrease only as information and no assertion is satisfied by that decrease.

- RED-now: E1 (verb absent). Green path: M3 — G-FOLD with the M1 checker; expected: invariants (a)-(d) hold, doctor orphan set unchanged, exit 0.
- Mutant probe (executed, §5): delete-lines, drop-link-on-move, move-to-an-unlinked-file, truncate-tail and move-into-a-linked-non-index-topic-file each FAIL (a)-(c); **three further mutants PASS (a), (b) and (c) and FAIL (d)**: the correct fold plus deletion of a link-free blockquote line and the `## Open work` heading (`MEMORY.md` 730 bytes); a fold that files the lines into an unlinked topic file carrying three resolved links; a fold that also appends a byte to a non-index topic file. The correct fold PASSes all four invariants. An inclusion-only checker (a)-(c) therefore cannot adopt this criterion; the exact-equality (d) is what kills the three.
- Class: release-blocking.

### AC-MFB-004 — The archived line is verbatim; the archive index is chosen as §1.5 specifies (maps REQ-MFB-003)

**Given** a copy of `fixtures/store-A` and a variant with two archive indexes (`…_2026_09.md` and `…_2026_10.md`), both present and both linked from `MEMORY.md` **When** `t9001` is folded and `t9002` is folded **Then** each archived line is byte-for-byte equal to the line it was removed from (title, every target and trailing text; the `t9001` line keeps its trailing text `— merged; detail lives in the topic file`); lines are appended at the end of the archive index in their original order and no other byte of the archive index changes; in the variant the greater name (`…_2026_10.md`) receives them and `…_2026_09.md` is byte-identical; and the archive index of a copy with Windows line endings receives lines that follow its existing line ending.

- RED-now: E1. Green path: M3 — G-FOLD (`ArchiveSelection` for the selection cells, `VerbatimFiling` for the byte equality).
- Mutant probe: a rewrite that replaces the title or the trailing text by the target's description, or strips the link, is killed by the byte-equality assertions; a target chooser that takes the first name instead of the greatest is killed by the variant. Paper mutants. The month-rollover selection cell (older archive linked, newer present but unlinked) lives in AC-MFB-006, which states the one selection criterion again from the same §1.5 definition.
- Class: release-blocking.

### AC-MFB-005 — Fold is idempotent, recovers from a partial apply, and keeps a differing line (maps REQ-MFB-005)

**Given** a copy of `fixtures/store-A` **When** `fold --card t9001 --yes` runs twice, and separately when a test seam stops an apply after the archive append and before the `MEMORY.md` rewrite and the command is then re-run, and separately in a variant whose archive index already carries a line with the same link-target set as the `t9001` line but different trailing text **Then** the second run exits 0, reports nothing remains, and every file of the store is byte-identical to the state after the first run; after the interrupted apply the line exists in both `MEMORY.md` and the archive index, and the re-run removes it from `MEMORY.md` while the archive index carries exactly one line with that link-target set; in the differing-text variant the `t9001` line stays in `MEMORY.md`, the command reports `archive carries a different line for the same targets`, the archive index is byte-identical to before, and the exit code is 0.

- RED-now: E1. Green path: M3 — G-FOLD.
- Mutant probe: an append-always implementation leaves two archive lines after the recovery re-run — killed by the exactly-one count; a dedupe keyed on the target set alone removes the `t9001` line without filing its new text — killed by the differing-text variant. Paper mutants.
- Class: release-blocking.

### AC-MFB-006 — Edge inputs end as specified, including a month rollover and an archive index too small to carry the fold (maps REQ-MFB-006, REQ-MFB-003)

**Given** copies and generated variants of `fixtures/store-A` **When** the operator applies a fold with `--yes` for a card with no line (`t9999`); for a card with only AMBIGUOUS lines (`t9004`); in a variant whose only archive-pattern file is present but not linked from `MEMORY.md`; in a variant with no archive-pattern file at all; in a **month-rollover variant** in which `project_card_archive_2026_09.md` (three resolved links) is linked from `MEMORY.md` and `project_card_archive_2026_10.md` is present but not linked; in a variant whose archive index carries one resolved link (so the fold would leave it with two, under the threshold of three); in a variant whose archive index carries two resolved links (so the fold leaves it with exactly three); and with an invalid id **Then**:

- the first two exit 0 and write nothing (store byte-identical);
- the unlinked-archive variant and the no-archive variant exit non-zero, write nothing, and the message names the file to link (the unlinked file) or to create;
- **the month-rollover variant folds into `…_2026_09.md` (the greatest name among the files `MEMORY.md` links and that exist, §1.5), exits 0, leaves `…_2026_10.md` byte-identical, and lists `…_2026_10.md` as an unlinked archive-pattern file skipped as information**; the same single criterion that AC-MFB-004 asserts for two linked archives;
- the one-link variant exits non-zero, writes nothing, and the message names the archive index, its resolved-link count after the fold (2) and the threshold (3);
- the two-link variant folds and passes the checker with the doctor reporting no new orphan;
- the invalid id exits non-zero before any read;
- no archive index is ever created by the command.

- RED-now: E1. Green path: M3 — G-FOLD (`EdgeInputs` and `ArchiveSelection`); the exit code is asserted as its own field in every cell. The boundary pair (one link refused, two links accepted) is read against the accessor value of the doctor's threshold, not a literal.
- Mutant probe: an implementation that auto-creates the missing archive index is killed by the byte-identity assertion and the non-zero exit; one that prints the error and exits 0 (report without verdict) is killed by the exit-code field; one that files into the one-link archive index and exits 0 is killed by the refusal cell (the doctor would then report new orphans); a chooser that takes the greatest-named archive-pattern file overall and then tests the link refuses (or worse, files into the unlinked file) in the rollover variant — killed by its expected outcome. Paper mutants.
- Class: release-blocking.

### AC-MFB-007 — Apply order, abort-on-change, pre-write exactness and failure safety (maps REQ-MFB-004)

**Given** a copy of `fixtures/store-A` **When** a test seam injects a failure (i) at the archive append, (ii) at the `MEMORY.md` rewrite, (iii) mutates `MEMORY.md` between plan and apply, and (iv) makes the in-memory store that the plan would produce differ from the plan (one extra line removed from `MEMORY.md`, or one stray byte in a non-index file) **Then** (i) leaves the store byte-identical and exits non-zero; (ii) leaves the line present in both files, invariants (a) and (b) holding, and exits non-zero; (iii) aborts with a non-zero exit, writes no archive line, and leaves `MEMORY.md` byte-equal to the mutated version; (iv) aborts **before any write**, exits non-zero, names invariant (d), and leaves the store byte-identical; in every case no temporary file remains.

- RED-now: E1. Green path: M3 — G-FOLD.
- Mutant probe: remove-then-append ordering loses the line in (ii) — killed by invariant (a); skipping the content re-check overwrites the concurrent edit in (iii) — killed by the byte-equality assertion; a pre-write check limited to (a)-(c) writes the divergent store in (iv) — killed by the byte-identity assertion. Paper mutants.
- Class: release-blocking.

### AC-MFB-008 — Card close folds only when enabled, opens nothing when disabled, fails open and bounded, names its store (maps REQ-MFB-007)

**Given** an isolated environment (temporary `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR`, `MOAI_HOME`; a store seeded from the fixture at the resolved profile path; a queue seeded with the fixture's card ids; a path-recording seam on the helper's file opens) **When** a card is closed through `todo done`, through `todo auto-done`, and through the `todo --auto` cycle **Then**:

- (i) **differential, gate unset:** within one test, a close with the gate unset and a close with the fold disabled by construction (the fold helper replaced by a no-op) produce equal stdout, stderr, exit code and queue record, and the store's SHA-256 is unchanged — no pre-existing artifact is needed; the values `0`, `false`, `yes` and the empty string each behave as unset, and `1`, `TRUE` and ` true ` each enable;
- (ii) with the gate enabled the store is folded as AC-MFB-003 requires, exactly one extra stderr line names the store, the archive file and the line count, and stdout equals the disabled run's byte for byte;
- (iii) with the gate enabled and the archive index absent, the close exits 0 with the identical stdout and the queue record archived, plus one stderr line;
- (iv) the fold starts only after the queue already holds the archived card (outside the lock);
- (v) no more than one store is touched, and it lies beneath the temporary root;
- (vi) **seeded panic:** with the gate enabled and the fold helper made to panic, each of the three close paths returns with the same stdout, the same exit status and the same queue record as in (iii), plus at most one stderr line, in under 400 ms (the test-overridden bound of 200 ms plus 200 ms of slack);
- (vii) **blocked read:** with the gate enabled, the store's `MEMORY.md` replaced by a FIFO with no writer, and the bound overridden to 200 ms through the test seam, each of the three close paths returns in under 400 ms (bound plus 200 ms slack) with the same stdout, exit status and queue record as in (iii), at most one stderr line, and no write to the store afterwards (the FIFO is released at teardown and a leak check passes);
- (viii) **gate off means no open, proven on a blocking FIFO:** with the gate unset, the store's `MEMORY.md` replaced by a FIFO with no writer, and the bound **left at its production value**, each of the three close paths returns in under 1 second with the same stdout, exit status and queue record as the fold-disabled-by-construction run, and the path-recording seam logs no open of any file of the store. A FIFO with no writer blocks every open for reading until a writer appears, so only a code path that never opens the file returns early; a mutant that resolves the store and reads `MEMORY.md` before checking the gate would block until the 2-second bound (or forever without one) and cannot return in under 1 second, nor leave the recorder empty. This cell proves no open for reading or writing; a bare existence check is neither and is not asserted against;
- (ix) **the production bound is the one in force:** with the gate enabled, `MEMORY.md` a FIFO with no writer and the bound **not** overridden, `todo done` returns in under 3 seconds (the production bound of 2 seconds plus 1 second of slack) with the identical stdout, exit status and queue record, and at most one stderr line;
- (x) **value and ceiling:** `config.DefaultMemoryFoldOnDoneBound` equals 2 seconds and is at or below the ceiling of 5 seconds stated in REQ-MFB-007 (`0 < bound <= 5 s` is asserted by name, so a production constant of minutes fails here even though every other cell overrides the bound).

- RED-now: E6a-todo, E6a-autodone, E6a-auto, each with its control E6b-* (no close path calls a fold, and the three paths exist), and E8a with control E8b (the bound constant does not exist; its neighbors do). Green path: M4 — G-WIRE (10 branches). The FIFO cells (vii), (viii) and (ix) are Unix-only: on Windows they are skipped with a recorded reason (plan B1), never passed silently.
- Mutant probe: wiring only `todo done` is killed by the auto-done and `--auto` cells; running the fold inside the `Mutate` callback is killed by (iv); default-on is killed by (i); an unrecovered panic is killed by (vi); a synchronous unbounded read is killed by (vii); a store read before the gate is checked is killed by (viii); a production bound of minutes is killed by (ix) and (x) (premise P11: a doctor-style read of a FIFO `MEMORY.md` blocks). Paper mutants.
- Class: release-blocking.

### AC-MFB-009 — Doctor reports bytes, characters, loaded-content characters and lines; the topic-file cap is unchanged (maps REQ-MFB-008, REQ-MFB-010)

**Given** the fixture, a generated store whose `MEMORY.md` contains 3-byte characters, and a generated store whose `MEMORY.md` opens with a YAML frontmatter block and contains an HTML comment **When** `moai memory doctor --json --dir <store>` runs **Then** `index_bytes` equals the file size, `index_chars` equals the Unicode code-point count, `index_loaded_chars` equals the code-point count after removing the leading frontmatter block and the HTML comment, `index_lines` is unchanged; on the fixture the values are 1,155, 1,095, 1,095 and 17; in the 3-byte store bytes exceed characters; the text render shows the same four figures; **and** `topic_files`, `cap` (50) and `MEMORY_TOPIC_COUNT_OVER_CAP` behave exactly as at the pinned tree, including with `--cap 3`.

- RED-now: E2 (the new keys are absent). Green path: M2 — G-DOCTOR. The unchanged-cap half is a **regression-guard** (E2 records `"cap":50` and `"topic_files":13` today); the measurement half is release-blocking.
- Mutant probe: computing characters as the byte length is killed by the 3-byte store; measuring bytes after trimming the trailing newline is killed by the size equality; counting loaded characters without removing the comment is killed by the third store. Paper mutants.
- Class: release-blocking (measurement), regression-guard (cap).

### AC-MFB-010 — Budget findings follow configuration values at the boundaries (maps REQ-MFB-009)

**Given** the fixture (1,155 bytes, 17 lines) copied to a temporary directory and passed as `--dir <temp store>` in every invocation **When** doctor runs with `--byte-cap` 1444, 1443 and 1155, with `--line-cap` 22, 21, 17 and 16, and, with no cap flag, on generated stores sized one byte below and exactly at the warn percentage of `config.DefaultMemoryIndexByteCap` **Then** at byte cap 1444 (79.98 %) no budget finding is emitted; at 1443 (80.04 %) `MEMORY_INDEX_BUDGET_WARN` is emitted for bytes; at 1155 (100 %) `MEMORY_INDEX_BUDGET_AT_CAP` is emitted for bytes in place of the warning; at line cap 22 (77.3 %) none, at 21 (81.0 %) a lines-axis `MEMORY_INDEX_BUDGET_WARN`, at 17 (100 %, lines equal to the cap, not above it) the lines-axis warning and no `MEMORY_INDEX_OVERFLOW`, and at 16 (lines above the cap) `MEMORY_INDEX_OVERFLOW` and no lines-axis warning — `--line-cap` governs the existing overflow check as well as the budget axis, so one line cap is in force; the flagless generated stores follow the constants by reference (below: none, at: warning), and their line counts stay under the lines-axis warn point so no lines finding mixes into the bytes cell; every budget finding names the axis, the value, the cap, the percentage and the basis text.

- RED-now: E3 (the flag is absent; a working check would warn at cap 1300). Green path: M2 — G-DOCTOR; boundaries are integer comparisons (`value*100 >= warnPercent*cap`).
- Mutant probe: a literal threshold in the check is killed by the flag cells (cap 1443 versus 1444 cannot both hold) and by the flagless constant-tied cells; keying on percentage rounding rather than the integer test is killed by the 1443/1444 pair; a `--line-cap` that does not reach the overflow check is killed by the 16/17 pair. Paper mutants.
- Class: release-blocking.

### AC-MFB-011 — The warning keys on bytes and states that bytes is a proxy (maps REQ-MFB-009)

**Given** a generated store whose `MEMORY.md` is made of 3-byte characters so that its bytes are at least 80 % of `--byte-cap` while its characters are below 80 % of that cap **When** doctor runs with that cap and `--dir <temp store>` **Then** `MEMORY_INDEX_BUDGET_WARN` is emitted although the character count alone would not trigger it; the finding text contains `raw bytes`, `conservative proxy` and `unconfirmed`; the JSON carries both `index_bytes` and `index_chars` so the divergence is visible.

- RED-now: E3. Green path: M2 — G-DOCTOR.
- Mutant probe: keying on characters yields no finding — killed; omitting the proxy statement — killed by the text assertions. Paper mutants.
- Class: release-blocking.

### AC-MFB-012 — Doctor classifies links, never collapses distinct targets, names the count by key (maps REQ-MFB-011, REQ-MFB-012)

**Given** the fixture copied to a temporary directory **When** `moai memory doctor --json --dir <copy>` and the text form run **Then** exactly two `MEMORY_REPO_RELATIVE_LINK` findings name `.moai/reports/t9002/verdict.md` and `.moai/reports/t9006/verdict.md` in full; no dangling finding names `verdict.md`; exactly one `MEMORY_DANGLING_INDEX_LINK` names `feedback_queue_jump_window.md` and carries no replacement suggestion; the key `index_link_targets` equals 11 in `--json` and appears with 11 in the text render; and the SHA-256 of every file of the copy is identical before and after the run (the doctor rewrites nothing, which is the doctor half of REQ-MFB-012).

- RED-now: E2 (one dangling finding for both repo-relative targets, naming only `verdict.md`; no repo-relative finding; no `index_link_targets` key). Green path: M2 — G-DOCTOR.
- Mutant probe: keeping base-name keying yields one `verdict.md` finding and a count of 10 — killed; omitting the key leaves the count with no carrier — killed by the key assertion; a doctor that rewrites a link — killed by the hash equality. Paper mutants.
- Class: release-blocking.

### AC-MFB-013 — The reachability checker is itself falsifiable; size alone never passes (maps REQ-MFB-012)

**Given** the checker of M1 and the nine store transformations of §5 (the correct fold and eight lossy mutants) applied to the fixture **When** the checker evaluates each **Then** the correct fold passes all four invariants (a)-(d) and each lossy mutant fails: the five link-losing mutants fail (a), (b) or (c), including the mutant that files a STRONG line into a linked topic file that is not an index (fewer than the doctor's threshold of resolved links), which fails (a), (b) and (c); the three further mutants — the link-free line and heading deletion, the unlinked-index-like destination, the stray edit to a non-index file — pass (a), (b) and (c) and **fail (d)**, naming (d1), (d2) and (d3) respectively; every transformation, correct or not, shrinks `MEMORY.md` by at least as much as the correct fold (a transformation that only reduces bytes or lines never passes); the checker takes the plan's `removed` and `appended` lists and compares full store snapshots; the index set is built from the doctor's threshold read through the accessor (a boundary test flips membership at threshold − 1 versus threshold); the checker is applied by every apply-path test of this SPEC; and no test or command of this SPEC reads or writes a store outside a temporary directory or `fixtures/store-A/` read-only.

- RED-now: not executable on the pinned tree — the Go checker does not exist and a test selector for it would sweep nothing and exit 0, the vacuous direction. Disposition per the undecidable rule: **regression-guard**, not recorded as a pass at plan phase. The design-validation observation is the reference oracle run in §5 (scratch, reproducible by the stated transformations, never counted as RED-now). Run phase records the Go checker's own RED and mutant outputs in `progress.md` §E.2.
- Green path: M1 — G-CORE (11 branches).
- Class: regression-guard.

### Requirement-to-criterion map (derived from the `maps` clauses above)

| Requirement | Criteria |
|---|---|
| REQ-MFB-001 | AC-001 |
| REQ-MFB-002 | AC-002 |
| REQ-MFB-003 | AC-003, AC-004, AC-006 |
| REQ-MFB-004 | AC-003, AC-007 |
| REQ-MFB-005 | AC-005 |
| REQ-MFB-006 | AC-006 |
| REQ-MFB-007 | AC-008 |
| REQ-MFB-008 | AC-009 |
| REQ-MFB-009 | AC-010, AC-011 |
| REQ-MFB-010 | AC-009 |
| REQ-MFB-011 | AC-012 |
| REQ-MFB-012 | AC-003, AC-012, AC-013 |

12 requirements, 13 criteria; every requirement has at least one criterion and every criterion maps at least one requirement.

## 4. Check specifications (three parts) and continued firing

### 4.1 The byte-budget check (doctor)

- **(a) WHEN it must run to be meaningful:** on demand (`moai memory doctor`). It is meaningful only on a store near or past the threshold, so the boundary cells (79.98 % / 80.04 % / 100 %) are the proof it can tell the two sides apart.
- **(b) the INPUT that turns it red:** the fixture with `--byte-cap 1300` (88.8 %), observed red-by-absence at E3 today and required to produce `MEMORY_INDEX_BUDGET_WARN` after M2.
- **(c) who sees the red:** the operator who runs doctor reads the finding on stdout; a machine consumer reads `findings` in `--json`. The doctor exit code stays 0 (advisory, as today), so **no exit-code consumer and no CI job sees this red** — CI never reads the real store.
- **Continued-firing answer (stated honestly).** Question: if the doctor check stopped working tomorrow, what would be different in what is seen? On the operator's machine, nothing: this SPEC ships no unasked surface for the budget, because the SessionStart warning moved to the follow-up card (`spec.md` §4). What this SPEC does provide: (1) the `G-DOCTOR` tests run on every push and fail in CI if the check is removed or its boundaries move; (2) inside a checkout that carries history, the binary-lag advisory reports an installed binary that predates the feature (this holds only where a comparison commit exists, `verification-claim-integrity.md` §2.2 — a released binary inside a user project has no such signal). **Named residual, not closed:** until the follow-up card lands, a stopped or never-run doctor produces no unasked signal for the operator.

### 4.2 The reachability and exactness checker (invariants (a)-(d))

- **(a) WHEN:** in every apply path as a pre-write assertion on the in-memory result, and in every test.
- **(b) the INPUT that turns it red:** each of the eight lossy mutants of §5; the three that satisfy (a)-(c) turn (d) red.
- **(c) who sees the red:** the operator running `fold --yes` gets a non-zero exit and a message naming the violated invariant, with the store untouched; CI sees the `G-CORE` and `G-FOLD` results.
- **Continued-firing answer:** the checker runs on every real apply, so a regression that stopped it from running would remove the abort in AC-MFB-007 (iv), which CI exercises on every push; its non-execution is not indistinguishable from success because that cell asserts the abort.

### 4.3 The card-close wiring

Default-off means its non-execution is the designed state, so absence of the stderr line is **not** evidence the fold ran or failed. The check that it works is AC-MFB-008 in CI. A fold failure inside an unattended `todo --auto` cycle is visible only as one stderr line in that cycle's output; no durable record exists (named residual). An `auto-done` or `--auto` command closing N cards can take up to N times the per-card bound in the worst case (the bound is per card, `spec.md` REQ-MFB-007); that worst case is a named residual, not tested beyond one blocked card per path.

## 5. Mutant probe record (executed against the committed fixture, design validation)

Two reference oracles live in the scratch directory, not committed. `oracle2.py` implements §1.5's R, T and (a)(b)(c) over in-memory snapshots under two index-set definitions — the doctor's rule (threshold 3) and the v0.1.0 one (files `MEMORY.md` links) — and six transformations. `oracle3.py` adds invariant (d) and three more transformations, and writes each transformed store to a scratch directory so the doctor can read it. Neither reads a real store.

**Part 1 — the link-losing mutants (oracle2, re-run in revision 0.3.0, exit 0; per-invariant failure detail lines omitted here, present in the run output).** Verbatim verdict lines:

```
card lines selected: 3
== I(S) new = doctor rule (threshold 3)
baseline: reach=11 targets=14 bytes=1155
correct          size_after= 821 reach=11 targets=14 verdict=PASS
delete           size_after= 821 reach=10 targets=11 verdict=FAIL
drop-link        size_after= 821 reach=10 targets=11 verdict=FAIL
unlinked-file    size_after= 821 reach=10 targets=11 verdict=FAIL
truncate         size_after= 441 reach= 4 targets= 6 verdict=FAIL
non-index-topic  size_after= 821 reach=10 targets=13 verdict=FAIL
== I(S) old = files MEMORY.md links
baseline: reach=11 targets=14 bytes=1155
correct          size_after= 821 reach=11 targets=14 verdict=PASS
delete           size_after= 821 reach=10 targets=11 verdict=FAIL
drop-link        size_after= 821 reach=10 targets=11 verdict=FAIL
unlinked-file    size_after= 821 reach=10 targets=11 verdict=FAIL
truncate         size_after= 441 reach= 1 targets= 3 verdict=FAIL
non-index-topic  size_after= 821 reach=11 targets=14 verdict=PASS
```

Under the old definition the fifth mutant (`non-index-topic`: the `t9001` STRONG line filed into `feedback_verify.md`, which `MEMORY.md` links but which carries one resolved link, the two repo-relative lines filed into the archive index) PASSES with `MEMORY.md` at the same 821 bytes as the correct fold; under the doctor's rule it FAILS (a), (b) and (c).

**Part 2 — the exact-plan mutants (oracle3, run in revision 0.3.0, exit 0).** Three transformations satisfy (a), (b) and (c) under the doctor's index-set rule and are rejected only by (d). Verbatim output:

```
correct        MEMORY.md bytes= 821  (a)(b)(c)=PASS  (d)=PASS
m6-linkfree    MEMORY.md bytes= 730  (a)(b)(c)=PASS  (d)=FAIL
      d1: MEMORY.md differs from before-minus-planned-lines; unplanned lines gone: ['> Synthetic fixture for SPEC-M', '## Open work']
m7-hub         MEMORY.md bytes= 821  (a)(b)(c)=PASS  (d)=FAIL
      d2: project_card_archive_2026_10.md differs from before-plus-planned-lines
      d3: feedback_hub.md changed
m8-stray       MEMORY.md bytes= 821  (a)(b)(c)=PASS  (d)=FAIL
      d3: feedback_alpha.md changed
```

- `m6-linkfree` (the audit's sixth mutant): the correct fold plus deletion of the blockquote line and the `## Open work` heading — lines that carry no `.md` link. `MEMORY.md` is 730 bytes against the correct fold's 821.
- `m7-hub` (the audit's seventh mutant): a store carrying a planted `feedback_hub.md` — three resolved links, linked from no index — and a fold that files the three STRONG lines into it instead of the archive index. The hub is a member of the doctor's index set because it carries three resolved links, so (a), (b) and (c) pass.
- `m8-stray`: the correct fold plus one appended byte in `feedback_alpha.md`, a non-index topic file.

The doctor cannot tell these from the correct fold either. `./bin/moai memory doctor --json --dir <scratch copy>` (binary commit `2f492df19`, tree HEAD `0dfae6d4a`; scratch stores under the session scratch directory, never the real store), every exit 0:

| Store | `index_lines` | Findings |
|---|---|---|
| correct fold (`d3-correct`) | 14 | `MEMORY_ORPHAN_NOT_INDEXED` for `feedback_orphan.md` and `feedback_queue_jump_queue.md`; `MEMORY_DANGLING_INDEX_LINK` for `feedback_queue_jump_window.md` (path `MEMORY.md`) and `verdict.md` (path the archive index). Four findings. |
| `m6-linkfree` (`d3-m6-linkfree`) | 12 | The **same four findings** as the correct fold — two lines and a heading are gone and the doctor reports nothing different. |
| `m8-stray` (`d3-m8-stray`) | 14 | The **same four findings** as the correct fold. |
| `m7-hub` (`d3-m7-hub`) | 14 | The planted hub brings its own baseline findings (three `MEMORY_INDEX_DUPLICATE_ENTRY`, one `MEMORY_ORPHAN_NOT_INDEXED` for the hub); against its own baseline (`d3-base-hub`, 17 lines, 8 findings) the transformed store differs only in the `verdict.md` dangling finding moving to `feedback_hub.md`, and `project_card_t9001_alpha.md` is reported as neither orphan nor duplicate. |

An exact-match invariant catches these three where (a), (b) and (c) did not: it compares `MEMORY.md`, the archive index and every other file against what the plan alone would produce, so a deleted link-free line, a wrong destination and a stray edit each leave a difference that cannot be explained by the plan. All eight lossy mutants leave `MEMORY.md` at or below the correct fold's 821 bytes (the truncation at 441 and the link-free deletion at 730 reduce further): a size criterion cannot tell the correct fold from the mutants that destroy or hide data at the same or a smaller size, which is why no AC here is a size criterion.

## 6. Edge cases

An empty `MEMORY.md`; a `MEMORY.md` without a trailing newline (invariant (d1) preserves the trailing-newline state); a STRONG line whose first target is a store file lacking frontmatter (filed verbatim, as every line is); a card id that is a prefix of another (`t90` versus `t9001`); a line carrying two card ids in its targets (AMBIGUOUS); an archive index with Windows line endings (the appended lines follow the file's existing line ending); an archive index lacking a final line terminator (one is supplied before the appended lines, invariant (d2)); a store directory with a `_archive/` subdirectory (never read as an index); `MEMORY.md` being a symbolic link (refused, nothing written); `MEMORY.md` being a FIFO (the bounded read abandons it, AC-MFB-008 (vii), (viii), (ix)).

## 7. Quality gates and Definition of Done

All of: AC-MFB-001 … AC-MFB-012 green with their swept counts and AC-MFB-013 recorded; `go build ./...` and the Windows cross-build exit 0; `golangci-lint run` reports nothing new; coverage of `internal/hook/memo/taxonomy` and the new `internal/cli` files at least 85 %; `moai spec lint` on this SPEC reports no error from a binary built from the tree; the subagent-boundary grep empty; no rule file and no template mirror modified by this card (`git diff --name-only` against the card's base lists neither the `.claude/rules/moai/workflow/moai-memory.md` pair nor anything under `internal/template/templates/`), and no file under `internal/hook/` modified (the hook side moved to the follow-up card); every threshold, the execution bound and the gate's variable name referenced as a constant; no test, command or step touched the operator's real store — evidenced by the temporary-root assertions in the tests, since the real store itself may not be read to prove it. Quantitative decreases in bytes or lines appear only in reports as information.
