# t383 — plan-phase measurements

tree: 9328a5242 (= origin/develop at measurement time)
tool: moai v3.1.2, build commit 343399d2f

tool-provenance: 343399d2f IS an ancestor of 9328a5242 (`git merge-base --is-ancestor` rc=0),
so the installed build is BEHIND the tree. However
`git log 343399d2f..HEAD -- internal/cli/memory.go internal/hook/memo` is EMPTY, so the
`memory doctor` code path is identical between the judging build and the tree under
measurement. The lag therefore does not affect M4/M5 below.

## M1 — active index size

    wc -c  26280
    wc -m  18463      <-- characters, not bytes
    wc -l  163
    grep -c '^- \['  123

    ascii bytes      14493
    non-ascii bytes  11787   (45% of the file's bytes carry 3,970 of its 18,463 characters)

Doctrine (`.claude/rules/moai/workflow/moai-memory.md:117`) and the Go guard
(`internal/config/token_budget_guard.go` memoryHeadByteCap = 25*1024) both assert a
"200 lines OR 25KB, whichever comes first" cut.

## M2 — the byte-cap premise is NOT confirmed at this size

The file is 26,280 bytes, i.e. 680 bytes past 25*1024 = 25,600.
`tail -c +25601` yields 6 lines, 2 of which are `^- [` index entries.

OBSERVED: the session that took this measurement had the file's FINAL line
("배차 메시지는 큐 이동이 아니다 … 구별 불가로 보고") present in its injected context.
=> no cut at byte 25,600 occurred in that session.

Alternatives this single observation does NOT distinguish:
  - the cap is character-based (18,463 < 25,600 -> no cut)
  - the cap is line-based only (163 < 200 -> no cut)
  - the cap is byte-based but larger than 25*1024

This is evidence AGAINST a strict 25,600-BYTE cut. It is not proof of the cap's true shape.
The Claude Code loader is not in this repository, so its behaviour cannot be read from source
here; the observation above is the strongest available evidence.

## M3 — the Go guard never measures the real file

`alwaysLoadedSurface` measures `filepath.Join(repoRoot, "MEMORY.md")`.

    $ ls MEMORY.md
    ls: MEMORY.md: No such file or directory

The repository carries no root MEMORY.md, so the guard's MEMORY.md slot contributes 0 tokens
and its head cap never fires on the real index. The guard is vacuous with respect to the
auto-memory index it names. (The guard's own unit test supplies a fixture, so it stays green.)

## M4 — `moai memory doctor` on the ACTIVE store (the domain's dedicated tool)

    $ moai memory doctor --dir ~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory
      topic files : 177 (cap 50)
      index lines : 163
      MEMORY_ORPHAN_NOT_INDEXED      46
      MEMORY_DANGLING_INDEX_LINK     58
      MEMORY_TOPIC_COUNT_OVER_CAP    1

## M5 — TWO STORES. All 58 dangling links resolve in the other one.

  active  (loaded): ~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory
                    178 *.md; its MEMORY.md = 26,280 B / 163 lines  <-- the file in session context
  legacy (unread): ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory
                    1,098 *.md; its own MEMORY.md = 38,304 B / 217 lines

    dangling in active store: 58
      ...of which resolve in the LEGACY store: 58        <-- 58/58

### M5a — CORRECTION (revision 2). The 58/44 gap is a LINE POPULATION, not dedup.

This section has now been wrong twice. Both errors are recorded, because the second one was
introduced while fixing the first and that is the more instructive failure.

  rev 0 wrote "47% of index entries (58 of 123)" — read an occurrence count as a line count.
  rev 1 wrote "58 occurrences resolving to 44 unique files" — attributed the 58->44 gap to
        DEDUPLICATION. Measured, dedup accounts for 2, not 14.

Measured on this tree (all four figures from one script over the same file):

    scope: index entry lines only (`^- \[`)
      link occurrences                 136
      unique targets                   134      <-- dedup is worth 2
      unique MISSING targets            44
      entry lines carrying >=1 missing   40     -> 32.5% of 123 entry lines
      partially-dangling lines            0     -> every affected line is fully dead

    scope: whole file (every markdown link, any line shape)
      link occurrences                 191
      unique targets                   189      <-- dedup is worth 2 here too
      unique MISSING targets            58      <-- what `moai memory doctor` reports

    missing targets referenced ONLY outside `^- \[` lines: 14

The 14 sit on grouped lines whose shape the `^- \[` anchor does not match — bold-lead lines
and `- <text>: [link] …` lines. Enumerated:

    feedback_cli_existence_claim_full_output.md   feedback_no_local_full_suite.md
    feedback_number_without_its_unit.md           feedback_pkill_f_kills_own_shell.md
    feedback_prompt_premise_verification.md       feedback_verification_load_discipline.md
    project_branch_guard_discrim_complete.md      project_design_docs_v31_closed.md
    project_factory_mode_closed.md                project_harness_learning_evo_closed.md
    project_kanban_sweep_closed_2026_08_15.md     project_moai_mcp_integration_handoff.md
    project_navigator_epic.md                     project_spec_catalog_cleanup.md

The correct sentence: **58 unique missing targets file-wide; 44 of them reachable from
`^- \[` entry lines and 14 only from other line shapes; 40 of 123 entry lines affected
(32.5%).** Never restate any of these numbers without naming its scope AND its unit.

Two consequences that follow from this and were got wrong downstream:
  - restoring `moai memory doctor`'s dangling count to 0 requires all 58, so the active
    store goes 177 -> 235, not 177 -> 221.
  - `grep -c '^- \['` is blind to the 14. It cannot be the only entry-count metric.

### M5b — `moai memory doctor` WITHOUT `--dir` measures nothing here

    $ printenv CLAUDE_PROJECT_DIR      # rc=1, unset
    $ moai memory doctor --json        # run from this worktree
    [{"store":{"dir":".../projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t383/memory",
       "origin":"CLAUDE_CONFIG_DIR"},"exists":false,"topic_files":0,"index_lines":0,"findings":null},
     {"store":{"dir":"/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t383/memory",
       "origin":"default ~/.claude"},"exists":false,"topic_files":0,"index_lines":0,"findings":null}]

The store is derived from `CLAUDE_PROJECT_DIR` else `os.Getwd()`, so inside a worktree the
slug gains the worktree path and both candidate stores are absent. A finding count read from
this invocation is 0 because nothing was measured, not because nothing is wrong.
Every citation of this tool MUST carry `--dir <active store>` AND assert `exists: true`
with `index_lines > 0` before any count is read.

## Consequence for the card's three candidates

The card assumed the binding constraint is index LENGTH. Measured, it is not:

  - 40 of 123 index entry lines (32.5%) point only at files the session cannot open,
    and a further 14 missing targets are reachable only from grouped lines (M5a).
  - 46 topic files in the active store carry no index entry at all.
  - The 25,600-byte figure the previous compression optimized against is unconfirmed,
    and the one direct observation available contradicts it.

An index diet that shortens entries without addressing M5 shrinks a file that is
a third dead links by entry line (32.5%), and leaves both unreachability defects in place.
Whether a third is still enough to decline an index diet is a judgement the SPEC must
re-justify at this figure, not inherit from the "approximately half" wording of rev 0.
