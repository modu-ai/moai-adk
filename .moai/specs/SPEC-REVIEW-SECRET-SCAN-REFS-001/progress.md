# Progress — SPEC-REVIEW-SECRET-SCAN-REFS-001

Card: t629 · Card base: `feeecc980` · Plan-time HEAD: `21e5837dc` (tree `38dc028c5`)

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-10: `spec.md`, `plan.md`, `acceptance.md`, and this file.
  Tier M; no `design.md` or `research.md`.
- Defect and cost claims trace to `.moai/reports/t629/reproduction.md` and
  `.moai/reports/t629/cost-baseline.md` (both measured on tree `feeecc980`). The review workflow
  copies are unchanged between `feeecc980` and `21e5837dc`:
  `git diff --stat feeecc980 21e5837dc -- <both copies>` printed nothing.
- SPEC ID regex check executed:
  `[[ "SPEC-REVIEW-SECRET-SCAN-REFS-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- **Open decision:** `spec.md` §3 "Open decision — coverage versus cost" is unresolved and reserved
  to the operator. The run phase is gated on it (REQ-008, AC-007).
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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
