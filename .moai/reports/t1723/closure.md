# t1723 — closure record (lane-31, 2026-10-19)

Card t1723, SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001, branch `WT-cli-windows-guide`, tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1723`.

## Ruling

Board record `d-20261019T020706Z-9558` (read by the lane with `moai decision read`, 821 records; `decided_by=leader (session d1aa9eb6) under operator delegation`, `scope=card:t1723`, `kind=ruling`, follows `c26e`): **Q-SWC-1 = (b)**. An accurate retirement notice ("`moai cg` is retired") is not a stale citation; the twelve lines stay. OI-1 is moot under (b). Consequence recorded by the ruling: zero page edits, so the card closes on the recorded basis (no run or sync work). The leader runs the factory `abandon` and drops the queue card; the lane only stops renewing the lease.

Earlier operator record `d-20261019T015908Z-220a` (decided_by=operator, AskUserQuestion 10-19): Q13 = replace with the current equivalent where one exists, delete where none exists; Q17 = all four locales in the same change. Both are cited in `decision-index.md`; under (b) neither produces an edit.

## Measured basis

- Command: `git grep -n -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 74ad197380 -- <the twelve paths of the three pages x ko, en, ja, zh>`.
- Observed in this run, against tree `74ad197380`: 12 lines, one per file, and every line is the sentence "`moai cg` has been retired…" in its locale (verbatim in `acceptance.md` E-SWC-2). Class C (a stale citation telling the reader to run an absent command): 0. Class N (retirement notice): 12.
- Plan commit: `6cb246dbf9` (Tier M, 5 requirements, 6 acceptance criteria), five files under `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001/`, `moai spec lint --strict` on spec.md printed "No findings" (reported by manager-spec; not re-run by the lane).

## Findings kept for the other cards

- The card's claim "shares ko files with child 4 (README) but different lines" is false for three Korean lines (`cli.md`, `windows-guide.md`, `what-is-moai-adk.md`): they are the Korean CG sentence the README child's REQ-PRC-004 targets (`acceptance.md` E-SWC-8; the lane reproduced the same-sentence match with the E-01 grep above on the ko lines 338, 108, 441). The leader records this as a note on the README card (t1720).
- The card's named sources (`child-cards.md`, `split-proposal.md` of the t1655 tree) do not exist (`acceptance.md` G1; lane check in `premise-check.md` P2).
- The umbrella id is `SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001`, not `...-PREREWRITE-001`.

## What was not done

- **The plan audit (plan-auditor, iteration 0) did not complete.** The lane stopped it after the ruling made it moot. At the stop, `.moai/reports/t1723/` held only `premise-check.md`, `git status --short` was empty, and `.moai/state/audit-receipts/rejections` did not exist. Whether the auditor had already called `audit_multi`, and therefore whether a receipt was issued, was NOT observed. Treat any receipt for this SPEC as unfiled and unread.
- The SPEC's own `progress.md` was not edited by the lane (it is a plan-specialist artifact); the SPEC stays as committed (`status: draft`, Q-SWC-1 recorded as open there). This record and the board id above are the closure basis.
- No lease renewal after 02:22:11Z (the last renewal was 2026-10-19T02:07:11Z → until 02:22:11Z); the renewal cron was removed.
