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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
