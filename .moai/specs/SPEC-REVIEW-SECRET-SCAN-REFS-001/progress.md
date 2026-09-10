# Progress — SPEC-REVIEW-SECRET-SCAN-REFS-001

Card: t629 · Card base: `feeecc980` · Plan-time HEAD: `21e5837dc` (tree `38dc028c5`) · Plan revised
after the operator decision, on top of `af7eb142b`

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-10: `spec.md`, `plan.md`, `acceptance.md`, and this file.
  Tier M; no `design.md` or `research.md`.
- Defect and cost claims trace to `.moai/reports/t629/reproduction.md` and
  `.moai/reports/t629/cost-baseline.md` (both measured on tree `feeecc980`). The review workflow
  copies are unchanged between `feeecc980` and `21e5837dc`:
  `git diff --stat feeecc980 21e5837dc -- <both copies>` printed nothing.
- SPEC ID regex check executed:
  `[[ "SPEC-REVIEW-SECRET-SCAN-REFS-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- **Decision recorded:** `spec.md` §3 carries the operator's decision — Option 2 (per-ref tip-set
  checkpoint), conditional on the measure-first gate in `spec.md` §3.4, with an exact example-value
  allowlist and no path exclusions — recorded alone in commit `af7eb142b`. `spec.md`, `plan.md`, and
  `acceptance.md` changed after the earlier plan audit, so plan-audit must re-run before the run
  phase.
- Plan-time baselines, measured in this session on the worktree at `21e5837dc`:

| Reading | Command | Output |
|---|---|---|
| equivalence phrase | `/usr/bin/grep -c 'produces the same coverage over time as the former every-review full scan' LOC TPL` | `1`, `1` |
| dropped-class phrase | `/usr/bin/grep -c 'no finding class is dropped' LOC TPL` | `1`, `1` |
| checkpoint-content phrase | `/usr/bin/grep -c 'the HEAD SHA of the last completed scan' LOC TPL` | `1`, `1` |
| regex self-match in copies | `/usr/bin/grep -cE -- 'REGEX' LOC TPL` | `0`, `0`, exit 1 |
| internal IDs in `TPL` | `grep -c -e 'SPEC-'` and `grep -c -e 't629'` | `0`, `0` |
| repo cost figures in `TPL` | `/usr/bin/grep -cE -e '12,157\|12157\|76\.084\|0\.207' TPL` | `0`, exit 1 |
| language names in section | `sed` section extract, then `grep -ciwE '<16 names>'` | `0`, exit 1 |
| scan commands in section | `/usr/bin/grep -n -e 'log -p' SP/section.txt` | 2 lines; line 8 `<last-sha>..HEAD`, line 14 `--all` |
| credential-shaped added lines | `git diff feeecc980 HEAD --output=SP/…`, then `grep -cE '^\+.*(REGEX)'` | 128-line diff, `0`, exit 1 |

- `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001`, run after all four artifacts existed:
  `✓ No findings — all SPEC documents are valid`.
- Regex check over the four SPEC artifacts: `/usr/bin/grep -rcE -- 'REGEX' <SPEC dir>` → `0` for
  each file, exit 1.
- Gap: the credential-shaped added-line reading covers **committed** changes only. At plan time
  `.moai/reports/t629/cost-baseline.md` was untracked (`git status --short`), so that reading did
  not include it. AC-004 at close runs after the lane's evidence commit.
- Revision readings (plan 0.2.0, after the decision commit `af7eb142b`):
  - AC-004 over the working tree with every revision edit in place, on top of `af7eb142b`:
    `git diff feeecc980 --output=SP/card-diff-wt.txt`, then `wc -l` → `1236` lines, then
    `/usr/bin/grep -cE -- '^\+.*(REGEX)' SP/card-diff-wt.txt` → `0`, grep exit `1`.
    Gap: the lines recording this reading are not in it.
  - AC-006 on `TPL` at `af7eb142b`: `SP/cmds.txt` 2 lines; `grep -c -e '--not'` → `0` (exit 1);
    `grep -ci 'every ref'` on the section → `0` (exit 1); the HEAD-SHA checkpoint phrase → `1`.
  - AC-007: `git log --format=%H -S 'Decision:** Option' -- spec.md` → one line, `af7eb142b`;
    `git log --reverse --format=%H feeecc980..HEAD -- TPL` and `-- LOC` → 0 lines each.
  - AC-015: `/usr/bin/grep -cE -- 'REGEX' LOC TPL` → `0`, `0`, exit 1; a fragment-assembled
    PEM-header control → `1`, exit 0.
  - Requirement and criterion counts: `grep -cE '^- \*\*REQ-[0-9]{3} ' spec.md` → `12`;
    `grep -cE '^\| AC-[0-9]{3} ' acceptance.md` → `15` (Tier M ceilings: 16 and 16).
  - `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` → `0 error(s), 0 warning(s)`, with one INFO
    `OwnershipTransitionUnmeasured` on commit `78e29987f` (no `Authored-By-Agent` trailer).
- Revision readings (plan 0.2.1, resolving plan audit iteration 1 findings D1-D12, on top of
  `496fe6153`; the audit report is `.moai/reports/t629/plan-audit-iter1.md`):
  - AC-004 over the working tree with every 0.2.1 revision edit in place:
    `git diff feeecc980 --output=SP/card-diff-wt2.txt`, then `wc -l` → `1667` lines, then
    `/usr/bin/grep -cE -- '^\+.*(REGEX)' SP/card-diff-wt2.txt` → `0`, grep exit `1`, read
    separately. Gap: the lines recording this reading are not in it.
  - AC-002 reworded-claim check: `/usr/bin/grep -ci 'same coverage' LOC TPL` → `1`, `1`.
  - AC-005 on the section extracted from `TPL`: `/usr/bin/grep -c 'R language'` → `0`, exit 1; the
    16-name language grep with `-n` → no lines, exit 1. The strict leak test named in AC-005 exists:
    `grep -n 'func TestTemplateNoInternalContentLeak' internal/template/internal_content_leak_test.go`
    → line 1535. The leak run itself was not taken at plan time.
  - AC-016 mechanics on a scratch file in `SP`: the pinned-block `sed` extraction printed the heading
    and body, kept a level-4 sub-heading, and stopped before the next level-3 heading;
    `grep -cF -f <pattern file>` → `1` (exit 0) on a matching line and `0` (exit 1) on a control line.
  - Requirement and criterion counts: `grep -cE '^- \*\*REQ-[0-9]{3} ' spec.md` → `13`;
    `grep -cE '^\| AC-[0-9]{3} ' acceptance.md` → `16` (Tier M ceilings: 16 and 16); the §D.1-§D.16
    headings carry AC-001 to AC-016 in order.
  - Credential regex over the four SPEC artifacts: `/usr/bin/grep -cE -- 'REGEX' <4 files>` → `0`
    each, exit 1; a fragment-assembled PEM-header control in `SP` → `1`, exit 0.
  - `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001`, run with every other 0.2.1 edit in place →
    exit 0, `0 error(s), 0 warning(s)`, with the same one INFO `OwnershipTransitionUnmeasured` on
    commit `78e29987f`.
- Scan output granularity measurement, the basis for resolving plan audit iteration 2 finding D13
  (`.moai/reports/t629/plan-audit-iter2.md`). Taken 2026-09-10 on a throwaway fixture repository in
  the session scratchpad, outside this repository, with the worktree at `5fa97ecd7` (clean) and
  `git --version` → `git version 2.50.1 (Apple Git-155)`. No history scan ran on this repository.
  The record sits in this file because the subagent file-write guard refused a separate report file
  under `.moai/reports/t629/`.
  - Fixture: fixture identity, `commit.gpgsign=false`, `core.hooksPath=/dev/null`. Markers are
    PEM-style private-key header lines with distinct uppercase labels, each written by `printf` from
    four fragments and never reproduced here. Fixture commit F0: `keys.txt` with six context lines
    carrying `CTXLINE`, `other.txt` with two lines carrying `OTHERPLAIN`. F1: four adjacent added
    lines in one hunk of `keys.txt` — the `ALPHACELL` marker, a plain line carrying `DELTAPLAIN`,
    the `BRAVOCELL` marker, the `CHARLIECELL` marker. Supplementary F2, one commit: the `ECHOCELL`
    marker appended to `other.txt` and a plain line carrying `FOXPLAIN` appended to `keys.txt`.
    Supplementary F3, one commit: the `GOLFCELL` marker as the first line of `keys.txt` and a plain
    line carrying `HOTELPLAIN` at its end, two hunks. F2 and F3 were added because in F1 the hunk,
    the file, and the commit coincide.
  - Commands (`REGEX` = the scan regex as written in `review.md`; every output redirected to a file;
    every exit code read without a pipe; every `git` command exited 0): after F0
    `git -C FX log -p --all -G 'REGEX' > SP/gran/base.txt`, `wc -c` → `0` bytes; after each commit
    `git -C FX show HEAD --output=SP/gran/c<N>-show.txt` and the prescribed incremental form
    `git -C FX log -p HEAD~1..HEAD -G 'REGEX'` redirected to `SP/gran/inc.txt` (F1),
    `SP/gran/inc2.txt` (F2), and `SP/gran/inc3.txt` (F3); after F1 also
    `git -C FX log -p --all -G 'REGEX' > SP/gran/all.txt`; per file `wc -l`,
    `/usr/bin/grep -c '<LABEL>'`, `/usr/bin/grep -cE -- 'REGEX'`, `/usr/bin/grep -c '^@@'`,
    `/usr/bin/grep -c '^diff '`.
  - F1, patch / incremental scan / `--all` scan: `wc -l` 21 / 21 / 21; `ALPHACELL`, `BRAVOCELL`,
    `CHARLIECELL` each 1 in every file; lines matching `REGEX` 3 / 3 / 3; the non-matching added
    line `DELTAPLAIN` 1 / 1 / 1; context lines `CTXLINE` 6 / 6 / 6; `^@@` 1 / 1 / 1; the untouched
    file's `OTHERPLAIN` 0 / 0 / 0 (grep exit 1).
  - F2, patch / scan: `wc -l` 23 / 14; `^diff ` 2 / 1; `^@@` 2 / 1; `ECHOCELL` 1 / 1; lines
    matching `REGEX` 1 / 1; the non-matching file's `FOXPLAIN` 1 / 0.
  - F3, patch / scan: `wc -l` 20 / 20; `^diff ` 1 / 1; `^@@` 2 / 2; `GOLFCELL` 1 / 1; lines
    matching `REGEX` 1 / 1; the non-matching hunk's `HOTELPLAIN` 1 / 1.
  - Conclusion: the scan's native output is file-granular within a matching commit. Every hunk of a
    matching file is printed, with its non-matching added lines and its context lines; a
    non-matching file in the same commit is not printed. A label count over the raw scan output, or
    over any patch-shaped output, therefore also counts lines printed next to a finding. A listed
    line itself matches `REGEX`, so filtering patch-shaped output by `REGEX` does not remove a listed
    line printed next to a finding.
  - Gaps: only the PEM-header alternative; one git build; `--pickaxe-all`, a changed context width,
    merge commits, and stash entries were not exercised. The credential regex reading over this file
    is taken before this record's commit and recorded with the next revision readings; it does not
    include the lines that record it.
- Revision readings (plan 0.2.2, resolving plan audit iteration 2 findings D13-D16, on top of
  `3bb7f2423`; the audit report is `.moai/reports/t629/plan-audit-iter2.md`):
  - Ceiling override: plan audit iteration 2 was the last iteration within the Tier M ceiling of 2
    and returned FAIL. The operator approved one further audit — a third, exceeding the ceiling by
    one — scoped to D13-D16 and the regressions their fixes create (answered in the lead session,
    relayed by the lead, 2026-09-10). D17-D20 are not taken in this revision.
  - D13 basis: the scan output granularity measurement above, committed alone in `3bb7f2423` before
    the AC-013 and AC-014 definition that rests on it.
  - Credential regex over this file before `3bb7f2423`:
    `/usr/bin/grep -cE -- 'REGEX' progress.md` → `0`, exit 1; a fragment-assembled PEM-header
    control in `SP` → `1`, exit 0.
  - AC-004 over the working tree with every 0.2.2 revision edit in place, taken right before
    staging: `git diff feeecc980 --output=SP/card-diff-wt3.txt`, then `wc -l` → `2101` lines,
    then `/usr/bin/grep -cE -- '^\+.*(REGEX)' SP/card-diff-wt3.txt` → `0`, grep exit
    `1`, read separately. Gap: the lines recording this reading are not in it.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
