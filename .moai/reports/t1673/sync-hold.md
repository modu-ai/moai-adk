# t1673 sync hold (rule C final round FAILed)

- Card: t1673 (Class C, run tn1p4e), SPEC-PRECOMMIT-AFFECTED-SCOPE-001 (Tier M)
- Tree: `.moai/worktrees/t1673`, branch `WT-pre-commit-gate`, HEAD `52302620768b2880a9fcc6d5f91dd6888514a9fc` (clean, unpushed), base local main `3642a3a9b`
- Stage at hold: `sync` (v9). Lease: not renewed past 2026-10-18T04:32:07Z.

## Why held

Standing rule C (board 9e76; applied to sync chains by board b5b1 / 8ada / 8b4f): one correction commit plus one delta audit after the ceiling; a FAIL on that final delta means hold and shrink, no more rounds. Sync audit history on this card:

| Round | Scope | Verdict | Required codex receipt | Blocking |
|---|---|---|---|---|
| 1 | full card diff, `be0bb13dd` | FAIL 0.74 | rcpt-8b5dc83de1625ec3b29b2e61 | F1, F2, F3 |
| 2 | delta `183c4b18d..4b7da88ea` | FAIL 0.84 | rcpt-61e4f2bdaaa1e03d1d4a8342 | R2-1 |
| 3 | delta `d9f5b3e6e..788e1d894` | FAIL 0.88 | rcpt-004e21f1d6a55ad84226f3ff | R3-1 |

Verdict files: `sync-audit.md` (183c4b18d), `sync-audit-iter2.md` (d9f5b3e6e), `sync-audit-iter3.md` (523026207). Every round the codex required backend failed on one demonstrated under-selection edge of the same query family; the auditor reproduced each one.

## State of the work (observed, this run)

- All 15 live ACs pass; the RED-to-GREEN flip reproduces (round 2 and 3 audits re-ran probe016.sh: RED on 00e7843ab, GREEN on HEAD; probe009.sh remains the immutable RED witness).
- Fixed and covered by regression tests: F1 (`diff.relative`), F2 (index vs working-tree Go files), F3 (shared embed owners), R2-1 (untracked go.mod / go.work / go.work.sum).
- Open blocker R3-1 [P2]: `git --literal-pathspecs commit` exports `GIT_LITERAL_PATHSPECS=1` into the hook environment; the `ls-files` query for untracked module files uses `:(top,glob)` patterns, which are then matched literally, so the query succeeds with no output and the full-module fallback never fires (auditor repro: staged `a` plus untracked `b/go.mod` ran `go test ./a`; plain `git commit` ran `./...`). Reach: only that non-default git option, or the variable set in the environment.
- Open spec-body debt (manager-spec, not edited after kickoff): S1-a, S1-b, S2, AC-PAS-003 mutant wording, B12 counter artifact (marks `[RETIRED]` on AC-PAS-012 plus a prefix declaration), F6 (`mode: ""` accepted).
- Known gaps: Q4 unmeasured; no full-package `-race` outside `internal/hook/quality`; no full `internal/template` / `internal/cli` run (ruling 35e9); baseline failure `internal/config` `TestNoHardcodedContexts` (scripts/release.sh, untouched by the card).
- `sync_commit_sha` in progress.md E.4 is still `pending-backfill`.

## Shrink candidates (recorded, not issued)

1. Close R3-1 with one bounded correction (outside rule C, needs the leader or operator): scrub `GIT_LITERAL_PATHSPECS` from the untracked-file query environment, or list untracked files without a pathspec and filter by base name in Go; add a regression test through `git --literal-pathspecs commit`. Then a delta audit limited to that commit.
2. Accept R3-1 as a documented limit (a plan.md risk row plus one docs sentence): this needs a SPEC-body edit, so it goes through manager-spec first, and a required-backend FAIL is not lowered by a note, so the next audit round is still needed.
3. Carry the repaired-and-passing parts forward by landing them without the untracked-module clause: not recommended by the lane, because it would undo a fix the audits asked for.

Lane observation (inference, not a measurement): five consecutive findings (F1, F2, F3, R2-1, R3-1) come from the query family that reads git state under the user's git configuration and environment; a single hardening that fixes the environment for all these queries may close the class, but this was not tried.

Decision record: decided_by=lane-27 (rule C applied; the leader records the board entry) evidence_refs=sync-audit.md;sync-audit-iter2.md;sync-audit-iter3.md ladder_path=standing 9e76 rule C, precedent board 8b4f (t1690)
