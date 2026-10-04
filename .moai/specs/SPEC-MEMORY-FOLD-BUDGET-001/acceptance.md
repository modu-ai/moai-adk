# acceptance.md — SPEC-MEMORY-FOLD-BUDGET-001

Verification layer. Each criterion is an `AC-MFB-NNN` Given/When/Then, binary-testable, mapped to its requirement. The requirements themselves are the GEARS entries in `spec.md` §2; nothing here restates them as requirements.

## 0. Pins, provenance, conventions

- **Document-level tree pin: `2f492df19`.** It binds every RED-now cell below that carries no pin of its own. The plan commit descends from `2f492df19` and touches only `.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/**`, so no Go source differs between the pinned tree and the plan commit; the `fixtures/store-A/` files the commands read arrive in that commit (they were present untracked when the commands ran).
- **Binary provenance.** Every `./bin/moai` command was run with a binary built from this tree and invoked by path: `go build -buildvcs=false -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=2f492df19 -X github.com/modu-ai/moai-adk/pkg/version.Date=2026-10-04" -o bin/moai ./cmd/moai`; `./bin/moai version` prints commit `2f492df19`, the tree HEAD. A plain `go build` in this worktree stamped a different checkout's revision (`c8f245c2c9a5`, not an ancestor of HEAD), so the stamp was set explicitly; the commit is therefore declared, and the tracked tree was clean at build time (only this SPEC directory untracked).
- **Real store off limits.** No command in this document, and no green-path command, reads or writes the operator's real memory directory (`spec.md` C-1). Green-path tests copy the fixture or generate stores under `t.TempDir()`, set `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR` and `MOAI_HOME` to temporary directories, and fail if a resolved store lies outside the temporary root.
- **Commands are single invocations** run from the repository (tree) root, read-only, with the exit code recorded as its own field. Stdout is verbatim. A table cell never carries a command; the ledger below does, by id.
- **Green-path rule.** Every `go test -run` selector is anchored `^…$` on each alternation branch, and every green cell is read together with the swept count: `go test -list '<same pattern>' <package>` must list exactly as many names as the pattern has branches. A green with an empty sweep is not a pass. Test names below are the run-phase's to finalize; the binding is the behavior and the swept-count rule.
- **Release-blocking** = carries RED-now (ledger id) and a green path. **Regression-guard** = guards behavior that must not change, or has a RED that cannot be re-executed on the pinned tree; never recorded as an adopted gate.
- **Run-phase RED record.** For every Go-test criterion the run phase also records, before GREEN, the verbatim failing output of the new test itself (`tdd-result-contract.md` `EXPECTED_RED`) in `progress.md` §E.2. The plan-phase RED-now below proves the feature is absent; it does not replace that record.

## 1. Fixture and plan-phase baseline

`fixtures/store-A/` is a synthetic store (one `MEMORY.md`, one archive index `project_card_archive_2026_10.md` carrying three links, 11 further topic files and one orphan; 13 topic files besides `MEMORY.md`). Measured on the committed fixture (`wc -c`, `wc -m`, and a reference oracle in the scratch directory — not part of the tree):

| Quantity | Value |
|---|---|
| `MEMORY.md` bytes / characters / lines (doctor counting) | 1,155 / 1,095 / 17 |
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
note:    stderr, which is not the decision channel, reads: Unknown flag: --card
```

```
E2
command: ./bin/moai memory doctor --json --dir .moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A
stdout:
[{"store":{"dir":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A","origin":"--dir"},"exists":true,"topic_files":13,"cap":50,"index_lines":17,"findings":[{"Code":"MEMORY_ORPHAN_NOT_INDEXED","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/feedback_orphan.md","Detail":"feedback_orphan.md is linked from no index — a session loads an index, not the directory, so this memory is never recalled"},{"Code":"MEMORY_ORPHAN_NOT_INDEXED","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/feedback_queue_jump_queue.md","Detail":"feedback_queue_jump_queue.md is linked from no index — a session loads an index, not the directory, so this memory is never recalled"},{"Code":"MEMORY_DANGLING_INDEX_LINK","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/MEMORY.md","Detail":"index links verdict.md but no such file exists"},{"Code":"MEMORY_DANGLING_INDEX_LINK","Path":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1502/.moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A/MEMORY.md","Detail":"index links feedback_queue_jump_window.md but no such file exists"}]}]
exit:    0
tree:    2f492df19
red-because: the report has no index_bytes, index_chars or index_loaded_chars key; the two distinct repo-relative targets collapse into ONE dangling finding naming only the base verdict.md; the dangling finding carries no nearest-name suggestion; no MEMORY_REPO_RELATIVE_LINK finding exists
also-serves: positive control for E1 (the same binary and fixture, an existing verb, exit 0); regression evidence that "cap":50 and "topic_files":13 are today's values
```

```
E3
command: ./bin/moai memory doctor --byte-cap 1300 --dir .moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A
stdout:  (empty — 0 bytes)
exit:    1
tree:    2f492df19
red-because: the --byte-cap flag does not exist; 1,155 bytes is 88.8 % of 1,300, so a working budget check would warn here
note:    stderr reads: Unknown flag: --byte-cap
```

```
E4
command: ./bin/moai memory relink --dir .moai/specs/SPEC-MEMORY-FOLD-BUDGET-001/fixtures/store-A
stdout:  (empty — 0 bytes)
exit:    1
tree:    2f492df19
red-because: the relink verb does not exist
note:    stderr reads: Unknown flag: --dir
```

```
E5a
command: grep -c "memoryBudgetAdvisory" internal/hook/session_start.go
stdout:  0
exit:    1
tree:    2f492df19
red-because: no budget advisory is wired into the SessionStart handler
```

```
E5b  (positive control for E5a)
command: grep -c "guardLivenessAdvisory" internal/hook/session_start.go
stdout:  1
exit:    0
tree:    2f492df19
```

```
E6a
command: grep -c "foldClosedCardMemory" internal/cli/todo.go internal/cli/todo_autodone.go internal/cli/todo_auto.go
stdout:
internal/cli/todo_autodone.go:0
internal/cli/todo_auto.go:0
internal/cli/todo.go:0
exit:    1
tree:    2f492df19
red-because: none of the three card-close paths calls a memory fold
```

```
E6b  (positive control for E6a — the three close paths exist)
command: grep -c "ArchiveCard(" internal/cli/todo.go internal/cli/todo_autodone.go internal/cli/todo_auto.go
stdout:
internal/cli/todo_auto.go:1
internal/cli/todo_autodone.go:1
internal/cli/todo.go:1
exit:    0
tree:    2f492df19
```

### 2.1 Green-path commands (run-phase; selectors anchored, each paired with its swept count)

Each id below is a pair: the `-run` command, and the `-list` command with the same pattern. Green means exit 0, every listed name prints a PASS line, and the `-list` output names exactly as many tests as the pattern has branches. All of them run only against temporary directories and `fixtures/store-A/` read-only.

```
G-CORE      package ./internal/hook/memo/taxonomy — 6 branches (AC-MFB-015; M1)
go test -run '^TestReach_CorrectFoldPasses$|^TestReach_DeleteMutantFails$|^TestReach_DropLinkMutantFails$|^TestReach_UnlinkedFileMutantFails$|^TestReach_TruncateMutantFails$|^TestReach_SizeAloneNeverPasses$' -count=1 -v ./internal/hook/memo/taxonomy
go test -list '^TestReach_CorrectFoldPasses$|^TestReach_DeleteMutantFails$|^TestReach_DropLinkMutantFails$|^TestReach_UnlinkedFileMutantFails$|^TestReach_TruncateMutantFails$|^TestReach_SizeAloneNeverPasses$' ./internal/hook/memo/taxonomy

G-FOLD      package ./internal/cli — 7 branches (AC-MFB-001 … AC-MFB-007; M3)
go test -run '^TestMemoryFold_DryRunWritesNothing$|^TestMemoryFold_Classification$|^TestMemoryFold_ReachabilityPreserved$|^TestMemoryFold_LineForm$|^TestMemoryFold_Idempotent$|^TestMemoryFold_EdgeInputs$|^TestMemoryFold_ApplyOrderAndAbort$' -count=1 -v ./internal/cli
go test -list '^TestMemoryFold_DryRunWritesNothing$|^TestMemoryFold_Classification$|^TestMemoryFold_ReachabilityPreserved$|^TestMemoryFold_LineForm$|^TestMemoryFold_Idempotent$|^TestMemoryFold_EdgeInputs$|^TestMemoryFold_ApplyOrderAndAbort$' ./internal/cli

G-DOCTOR    package ./internal/cli — 5 branches (AC-MFB-009, -010, -011, -013; M2)
go test -run '^TestMemoryDoctor_Measures$|^TestMemoryDoctor_TopicCapUnchanged$|^TestMemoryDoctor_BudgetBoundaries$|^TestMemoryDoctor_BytesProxyWarns$|^TestMemoryDoctor_LinkClasses$' -count=1 -v ./internal/cli
go test -list '^TestMemoryDoctor_Measures$|^TestMemoryDoctor_TopicCapUnchanged$|^TestMemoryDoctor_BudgetBoundaries$|^TestMemoryDoctor_BytesProxyWarns$|^TestMemoryDoctor_LinkClasses$' ./internal/cli

G-WIRE      package ./internal/cli — 5 branches (AC-MFB-008; M4)
go test -run '^TestMemoryFoldOnDone_DisabledIsByteIdentical$|^TestMemoryFoldOnDone_EnabledFolds$|^TestMemoryFoldOnDone_FailOpen$|^TestMemoryFoldOnDone_RunsAfterQueueWrite$|^TestMemoryFoldOnDone_ThreeClosePaths$' -count=1 -v ./internal/cli
go test -list '^TestMemoryFoldOnDone_DisabledIsByteIdentical$|^TestMemoryFoldOnDone_EnabledFolds$|^TestMemoryFoldOnDone_FailOpen$|^TestMemoryFoldOnDone_RunsAfterQueueWrite$|^TestMemoryFoldOnDone_ThreeClosePaths$' ./internal/cli

G-HOOK      package ./internal/hook — 5 branches (AC-MFB-012; M5)
go test -run '^TestSessionStartMemoryBudget_FourSources$|^TestSessionStartMemoryBudget_BelowAbsentUnreadableKillSwitch$|^TestSessionStartMemoryBudget_JoinBound$|^TestSessionStartMemoryBudget_StaysUnderTempHome$|^TestHomeJoinSiteCountIsPinned$' -count=1 -v ./internal/hook
go test -list '^TestSessionStartMemoryBudget_FourSources$|^TestSessionStartMemoryBudget_BelowAbsentUnreadableKillSwitch$|^TestSessionStartMemoryBudget_JoinBound$|^TestSessionStartMemoryBudget_StaysUnderTempHome$|^TestHomeJoinSiteCountIsPinned$' ./internal/hook

G-RELINK    package ./internal/cli — 4 branches (AC-MFB-014; M6)
go test -run '^TestMemoryRelink_PreviewWritesNothing$|^TestMemoryRelink_RepairsUnambiguous$|^TestMemoryRelink_TieIsNotRewritten$|^TestMemoryRelink_NeverDeletes$' -count=1 -v ./internal/cli
go test -list '^TestMemoryRelink_PreviewWritesNothing$|^TestMemoryRelink_RepairsUnambiguous$|^TestMemoryRelink_TieIsNotRewritten$|^TestMemoryRelink_NeverDeletes$' ./internal/cli
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

### AC-MFB-003 — PRIMARY: fold preserves reachability, never size (maps REQ-MFB-003, REQ-MFB-004)

**Given** a copy of `fixtures/store-A` and the reference snapshot of its R, T and index set **When** the operator folds `t9001`, `t9002` and `t9006` with `--yes`, each on its own copy and then sequentially on one copy, and again on a generated variant in which the STRONG line is a grouped line carrying two link targets **Then** for every resulting store: (a) R(after) ⊇ R(before), (b) T(after) ⊇ T(before) counted by full target text across the whole index set, (c) every line removed from `MEMORY.md` has an equal-target line in the archive index; the grouped variant keeps both of its targets; the checker reports the byte and line decrease only as information and no assertion is satisfied by that decrease.

- RED-now: E1 (verb absent). Green path: M3 — G-FOLD with the M1 checker; expected: invariants hold, exit 0.
- Mutant probe (executed on the fixture with the reference oracle, §5): delete-lines, drop-link-on-move, move-to-an-unlinked-file and truncate-tail each FAIL the checker while shrinking `MEMORY.md` exactly as much as the correct fold; the correct fold PASSes.
- Class: release-blocking.

### AC-MFB-004 — The archived line has the specified form; the archive target is chosen as specified (maps REQ-MFB-003)

**Given** a copy of `fixtures/store-A` and a variant with two archive indexes (`…_2026_09.md` and `…_2026_10.md`), both linked from `MEMORY.md` **When** `t9001` is folded and `t9002` is folded **Then** the `t9001` archived line keeps its title and link bytes and carries the description `Alpha card closed and merged; fixture topic file` taken from `project_card_t9001_alpha.md`; the `t9002` line, whose target is not a store file, is filed byte-for-byte verbatim; lines are appended at the end of the archive index in their original order; in the variant the greater name (`…_2026_10.md`) receives them.

- RED-now: E1. Green path: M3 — G-FOLD.
- Mutant probe: a rewrite that replaces the title by the description, or strips the link, is killed by the byte-equality assertions on title and target; a target chooser that takes the first name instead of the greatest is killed by the variant. Paper mutants.
- Class: release-blocking.

### AC-MFB-005 — Fold is idempotent and recovers from a partial apply (maps REQ-MFB-005)

**Given** a copy of `fixtures/store-A` **When** `fold --card t9001 --yes` runs twice, and separately when a test seam stops an apply after the archive append and before the `MEMORY.md` rewrite and the command is then re-run **Then** the second run exits 0, reports nothing remains, and every file of the store is byte-identical to the state after the first run; after the interrupted apply the line exists in both `MEMORY.md` and the archive index, and the re-run removes it from `MEMORY.md` while the archive index carries exactly one line with that link-target set.

- RED-now: E1. Green path: M3 — G-FOLD.
- Mutant probe: an append-always implementation leaves two archive lines after the recovery re-run — killed by the exactly-one count. Paper mutant.
- Class: release-blocking.

### AC-MFB-006 — Edge inputs end as specified (maps REQ-MFB-006)

**Given** copies and generated variants of `fixtures/store-A` **When** the operator applies a fold with `--yes` for a card with no line (`t9999`); for a card with only AMBIGUOUS lines (`t9004`); in a variant with no archive index; in a variant whose greatest archive index is not linked from `MEMORY.md`; and with an invalid id **Then** the first two exit 0 and write nothing (store byte-identical); the last three exit non-zero, write nothing, and the message names the file to create or link or the invalid value; no archive index is ever created by the command.

- RED-now: E1. Green path: M3 — G-FOLD; the exit code is asserted as its own field in every cell.
- Mutant probe: an implementation that auto-creates the missing archive index is killed by the byte-identity assertion and the non-zero exit; one that prints the error and exits 0 (report without verdict) is killed by the exit-code field. Paper mutants.
- Class: release-blocking.

### AC-MFB-007 — Apply order, abort-on-change and failure safety (maps REQ-MFB-004)

**Given** a copy of `fixtures/store-A` **When** a test seam injects a failure (i) at the archive append, (ii) at the `MEMORY.md` rewrite, and (iii) mutates `MEMORY.md` between plan and apply **Then** (i) leaves the store byte-identical and exits non-zero; (ii) leaves the line present in both files, invariants (a) and (b) holding, and exits non-zero; (iii) aborts with a non-zero exit, writes no archive line, and leaves `MEMORY.md` byte-equal to the mutated version; in every case no temporary file remains.

- RED-now: E1. Green path: M3 — G-FOLD.
- Mutant probe: remove-then-append ordering loses the line in (ii) — killed by invariant (a); skipping the SHA re-check overwrites the concurrent edit in (iii) — killed by the byte-equality assertion. Paper mutants.
- Class: release-blocking.

### AC-MFB-008 — Card close folds only when enabled, fails open, names its store (maps REQ-MFB-007)

**Given** an isolated environment (temporary `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR`, `MOAI_HOME`; a store seeded from the fixture at the resolved profile path; a queue seeded with the fixture's card ids) **When** a card is closed through `todo done`, through `todo auto-done`, and through the `todo --auto` cycle **Then** (i) with the gate unset the stdout, stderr, exit code and queue record equal a golden recording of the pre-change behavior and the store's SHA-256 is unchanged; (ii) with the gate enabled the store is folded as AC-MFB-003 requires, exactly one extra stderr line names the store, the archive file and the line count, and stdout equals the disabled run's byte for byte; (iii) with the gate enabled and the archive index absent, the close exits 0 with the identical stdout and the queue record archived, plus one stderr line; (iv) the fold starts only after the queue already holds the archived card (outside the lock); (v) no more than one store is touched, and it lies beneath the temporary root.

- RED-now: E6a with control E6b (no close path calls a fold, and the three paths exist). Green path: M4 — G-WIRE.
- Mutant probe: wiring only `todo done` is killed by the auto-done and `--auto` cells; running the fold inside the `Mutate` callback is killed by (iv); default-on is killed by (i). Paper mutants.
- Class: release-blocking.

### AC-MFB-009 — Doctor reports bytes, characters, loaded-content characters and lines; the topic-file cap is unchanged (maps REQ-MFB-008, REQ-MFB-010)

**Given** the fixture, a generated store whose `MEMORY.md` contains 3-byte characters, and a generated store whose `MEMORY.md` opens with a YAML frontmatter block and contains an HTML comment **When** `moai memory doctor --json --dir <store>` runs **Then** `index_bytes` equals the file size, `index_chars` equals the Unicode code-point count, `index_loaded_chars` equals the code-point count after removing the leading frontmatter block and the HTML comment, `index_lines` is unchanged; on the fixture the values are 1,155, 1,095, 1,095 and 17; in the 3-byte store bytes exceed characters; the text render shows the same four figures; **and** `topic_files`, `cap` (50) and `MEMORY_TOPIC_COUNT_OVER_CAP` behave exactly as at the pinned tree, including with `--cap 3`.

- RED-now: E2 (the new keys are absent). Green path: M2 — G-DOCTOR. The unchanged-cap half is a **regression-guard** (E2 records `"cap":50` and `"topic_files":13` today); the measurement half is release-blocking.
- Mutant probe: computing characters as the byte length is killed by the 3-byte store; measuring bytes after trimming the trailing newline is killed by the size equality; counting loaded characters without removing the comment is killed by the third store. Paper mutants.
- Class: release-blocking (measurement), regression-guard (cap).

### AC-MFB-010 — Budget findings follow configuration values at the boundaries (maps REQ-MFB-009)

**Given** the fixture (1,155 bytes, 17 lines) **When** doctor runs with `--byte-cap` 1444, 1443 and 1155, and with `--line-cap` 22 and 21, and once with no flag on a generated store sized one byte below and exactly at the warn percentage of `config.DefaultMemoryIndexByteCap` **Then** at cap 1444 (79.98 %) no budget finding is emitted; at 1443 (80.04 %) `MEMORY_INDEX_BUDGET_WARN` is emitted for bytes; at 1155 (100 %) `MEMORY_INDEX_BUDGET_OVER` is emitted for bytes in place of the warning; at line cap 22 (77.3 %) none, at 21 (81.0 %) a lines-axis warning; the flagless generated stores follow the constants by reference (below: none, at: warning); every finding names the axis, the value, the cap, the percentage and the basis text.

- RED-now: E3 (the flag is absent; a working check would warn at cap 1300). Green path: M2 — G-DOCTOR; boundaries are integer comparisons (`value*100 >= warnPercent*cap`).
- Mutant probe: a literal threshold in the check is killed by the flag cells (cap 1443 versus 1444 cannot both hold) and by the flagless constant-tied cells; keying on percentage rounding rather than the integer test is killed by the 1443/1444 pair. Paper mutants.
- Class: release-blocking.

### AC-MFB-011 — The warning keys on bytes and states that bytes is a proxy (maps REQ-MFB-009)

**Given** a generated store whose `MEMORY.md` is made of 3-byte characters so that its bytes are at least 80 % of `--byte-cap` while its characters are below 80 % of that cap **When** doctor runs with that cap **Then** `MEMORY_INDEX_BUDGET_WARN` is emitted although the character count alone would not trigger it; the finding text contains `raw bytes`, `conservative proxy` and `unconfirmed`; the JSON carries both `index_bytes` and `index_chars` so the divergence is visible.

- RED-now: E3. Green path: M2 — G-DOCTOR.
- Mutant probe: keying on characters yields no finding — killed; omitting the proxy statement — killed by the text assertions. Paper mutants.
- Class: release-blocking.

### AC-MFB-012 — SessionStart adds exactly one budget line, always, bounded, fail-open (maps REQ-MFB-011, REQ-MFB-012)

**Given** an isolated profile root (temporary `HOME`, `USERPROFILE`, `CLAUDE_CONFIG_DIR`) with a store whose `MEMORY.md` is at least 80 % of `config.DefaultMemoryIndexByteCap` **When** the SessionStart handler runs for the sources startup, resume, clear and compact, and the handler's serialized stdout is parsed **Then** each run's `additionalContext` contains exactly one `[moai:memory-budget]` line naming the store path; a store one byte below the threshold, an absent store, an unreadable `MEMORY.md`, and `MOAI_MEMORY_AUDIT=0` each add none and the handler returns no error; a read seam that exceeds the join bound adds none and the handler returns within the bound plus slack; the derived store lies beneath the temporary root and the pinned home-join allowlist test passes with its new row.

- RED-now: E5a with control E5b (no wiring exists; the neighboring advisory does). The grep proves absence of the wiring, not of the behavior; the behavior's RED is the run-phase record. Green path: M5 — G-HOOK; the assertion reads the serialized output, never the `Data` map.
- Mutant probe: conditioning on `source == startup` is killed by the four-source loop; an unbounded read is killed by the bound cell; placing the line in the `json:"-"` `Data` map is killed by asserting on serialized output. Paper mutants.
- Class: release-blocking.

### AC-MFB-013 — Doctor classifies links, never collapses distinct targets, suggests only when unambiguous (maps REQ-MFB-013)

**Given** the fixture, and a variant with two store files both scoring at or above the similarity threshold against one dangling name **When** doctor runs **Then** on the fixture: exactly two `MEMORY_REPO_RELATIVE_LINK` findings naming `.moai/reports/t9002/verdict.md` and `.moai/reports/t9006/verdict.md` in full; no dangling finding names `verdict.md`; exactly one `MEMORY_DANGLING_INDEX_LINK` naming `feedback_queue_jump_window.md` with the nearest name `feedback_queue_jump_queue.md`; the distinct-target count of `MEMORY.md` is 11 by full text; in the variant the dangling finding carries no suggestion.

- RED-now: E2 (one dangling finding for both repo-relative targets, naming only `verdict.md`; no suggestion; no repo-relative finding). Green path: M2 — G-DOCTOR.
- Mutant probe: keeping base-name keying yields one `verdict.md` finding and a count of 10 — killed; suggesting on a tie — killed by the variant. Paper mutants.
- Class: release-blocking.

### AC-MFB-014 — Relink repairs only unambiguous cases and never deletes (maps REQ-MFB-014)

**Given** a copy of the fixture and a temporary project root (`CLAUDE_PROJECT_DIR`) containing `.moai/reports/t9002/verdict.md` but not the `t9006` path **When** `moai memory relink` runs in preview, then with `--yes` **Then** the preview and the result retarget `feedback_queue_jump_window.md` to `feedback_queue_jump_queue.md`, rewrite the `t9002` link to its absolute path, and leave the `t9006` link unchanged and reported; after apply |R| grows from 11 to 12, the resolvable-target set grows or is equal, the count of link occurrences and the line count of `MEMORY.md` are unchanged; a variant with two equally near files is not rewritten; preview leaves every file byte-identical.

- RED-now: E4 (verb absent). Green path: M6 — G-RELINK.
- Mutant probe: rewriting the dangling name to a guessed sibling in the tie variant is killed by the variant; deleting the unresolvable `t9006` line is killed by the line-count and occurrence-count equality. Paper mutants.
- Class: release-blocking.

### AC-MFB-015 — The reachability checker is itself falsifiable; size alone never passes (maps REQ-MFB-015)

**Given** the checker of M1 and the five store transformations of §5 (the correct fold and four lossy mutants) applied to the fixture **When** the checker evaluates each **Then** the correct fold passes and each lossy mutant fails on the invariant it breaks, while every transformation, correct or not, shrinks `MEMORY.md` identically; a transformation that only reduces bytes or lines never passes; the checker is applied by every apply-path test of this SPEC; and no test or command of this SPEC reads or writes a store outside a temporary directory or `fixtures/store-A/` read-only.

- RED-now: not executable on the pinned tree — the Go checker does not exist and a test selector for it would sweep nothing and exit 0, the vacuous direction. Disposition per the undecidable rule: **regression-guard**, not recorded as a pass at plan phase. The design-validation observation is the reference oracle run in §5 (scratch, reproducible by the stated transformations, never counted as RED-now). Run phase records the Go checker's own RED and mutant outputs in `progress.md` §E.2.
- Green path: M1 — G-CORE.
- Class: regression-guard.

## 4. Check specifications (three parts) and continued firing

### 4.1 The byte-budget check (doctor)

- **(a) WHEN it must run to be meaningful:** on demand (`moai memory doctor`) and, unasked, at every session start through the SessionStart line (4.2). It is meaningful only on a store near or past the threshold, so the boundary cells (79.98 % / 80.04 % / 100 %) are the proof it can tell the two sides apart.
- **(b) the INPUT that turns it red:** the fixture with `--byte-cap 1300` (88.8 %), observed red-by-absence at E3 today and required to produce `MEMORY_INDEX_BUDGET_WARN` after M2.
- **(c) who sees the red:** the operator who runs doctor reads the finding on stdout; a machine consumer reads `findings` in `--json`. The doctor exit code stays 0 (advisory, as today), so **no exit-code consumer and no CI job sees this red** — CI never reads the real store. The unasked surface is the SessionStart line, whose reach is the orchestrator model (the operator sees it only if relayed).
- **Continued-firing answer.** If the doctor check stopped working tomorrow nothing would change in what a healthy-looking session shows, because on-demand liveness is not liveness. The answer is the SessionStart line, which fires unasked and unconditionally (REQ-MFB-012), backed by three independent signals: (1) the SessionStart test of AC-MFB-012 fails in CI if the wiring is removed or conditioned; (2) the binary-lag advisory reports an installed binary that predates the feature; (3) `MOAI_MEMORY_AUDIT=0`, an explicit operator off-switch, **is itself unobservable** — setting it silences the line with no trace. That last residual is named, not closed.

### 4.2 The SessionStart warning

- **(a) WHEN:** every session start — startup, resume, clear and compact — with no path, branch or source condition.
- **(b) the INPUT:** a store whose `MEMORY.md` is at least 80 % of `config.DefaultMemoryIndexByteCap`; the cell asserts the line appears for all four sources and not for a store one byte below.
- **(c) who sees the red:** the orchestrator model, through `additionalContext`; a line placed in the hook's `Data` map would reach nobody (`json:"-"`), which AC-MFB-012 closes by asserting on the serialized output. Whether the operator sees the line depends on the orchestrator relaying it.
- **Continued-firing answer.** As 4.1: unconditional by construction, CI test on the wiring, binary-lag advisory for a stale binary, with the kill-switch residual named. A further residual: a host change that stops delivering `additionalContext` would silence the line and nothing in this SPEC would notice.

### 4.3 The card-close wiring

Default-off means its non-execution is the designed state, so absence of the stderr line is **not** evidence the fold ran or failed. The check that it works is AC-MFB-008 in CI. A fold failure inside an unattended `todo --auto` cycle is visible only as one stderr line in that cycle's output; no durable record exists (named residual).

## 5. Mutant probe record (executed against the committed fixture)

A reference oracle written in the scratch directory (not committed) implements §1.5's R, T and (a)(b)(c) over in-memory snapshots, and five transformations of the fixture's `MEMORY.md`: the correct fold of the three STRONG cards (lines moved verbatim to the archive index) and four mutants. Observed, verbatim from the run:

```
card lines selected: 3
baseline: reach=11 targets=14 lines=14 bytes=1155
correct        size_after= 821  reach=11 targets=14  verdict=PASS
delete         size_after= 821  reach=10 targets=11  verdict=FAIL
drop-link      size_after= 821  reach=10 targets=11  verdict=FAIL
unlinked-file  size_after= 821  reach=10 targets=11  verdict=FAIL
truncate       size_after= 441  reach= 1 targets= 3  verdict=FAIL
```

(`lines=14` in that output counts non-empty lines of `MEMORY.md`; the doctor's `index_lines` of 17 counts every line.) The first four rows share an identical post-size of 821 bytes: a size criterion cannot tell the correct fold from the three that destroy data, which is why no AC here is a size criterion. The truncate mutant is the pure size reducer; it shrinks furthest and fails (a), (b) and (c).

## 6. Edge cases

An empty `MEMORY.md`; a `MEMORY.md` without a trailing newline; a STRONG line whose first target is a store file lacking frontmatter (filed verbatim); a card id that is a prefix of another (`t90` versus `t9001`); a line carrying two card ids in its targets (AMBIGUOUS); an archive index with Windows line endings (the appended lines follow the file's existing line ending); a store directory with a `_archive/` subdirectory (never read as an index); `MEMORY.md` being a symbolic link (refused, nothing written).

## 7. Quality gates and Definition of Done

All of: AC-MFB-001 … AC-MFB-014 green with their swept counts and AC-MFB-015 recorded; `go build ./...` and the Windows cross-build exit 0; `golangci-lint run` reports nothing new; coverage of `internal/hook/memo/taxonomy` and the new `internal/cli` files at least 85 %; `moai spec lint` on this SPEC reports no error from a binary built from the tree; the subagent-boundary grep empty; the doctrine pair identical under `cmp`; every threshold and the gate's variable name referenced as a constant; no test, command or step touched the operator's real store — evidenced by the temporary-root assertions in the tests, since the real store itself may not be read to prove it. Quantitative decreases in bytes or lines appear only in reports as information.
