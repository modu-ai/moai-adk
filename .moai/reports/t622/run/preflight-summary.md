# Run-phase pre-flight (plan.md §C steps 1-8) — card t622

Tree: worktree `.claude/worktrees/t622`, branch `WT-git-procedure-fixes`, HEAD
`fa13c27b624cb20a34c81e45f7e1b4e60165d90f` (`pre1-head.txt`). All files below live in this
directory. Every command was run with literal values (BASE `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089`,
CARD_BASE from `card-base.txt`). This file and the files it cites were committed BEFORE the first
change commit, so the commit graph orders the baseline ahead of the edits (verification-claim
integrity §2.3). Note: the M1 working-tree edits already existed on disk when this commit was made;
the ordering claim is the commit order, not authoring order.

## Step 1 — tree and range left end

| Check | Command | Result | Expected | Verdict |
|---|---|---|---|---|
| toplevel | `git rev-parse --show-toplevel` | `.../worktrees/t622` (`pre1-toplevel.txt`) | worktree root | OK |
| branch | `git branch --show-current` | `WT-git-procedure-fixes`, exit 0 | same | OK |
| status | `git status --porcelain` | only `?? run-stale-b412/`, `?? run/` (`pre1-status.txt`) | clean scope files | OK |
| CARD_BASE | `git merge-base --all develop HEAD > card-base.txt` | `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf`, 1 line, exit 0 | exactly 1 line | OK |
| range control | `git diff --name-only develop...HEAD > card-range-names.txt` | exit 0, `test -e` 0, `test -s` 0, 356 lines | ≥1 line | OK |
| snapshot freshness | `git diff --name-only <BASE> <CARD_BASE> -- <21 snapshot paths> > snapshot-stale.txt` | exit 0, `test -e` 0, `test -s` 1 (`pre1-staleness-results.txt`) | empty | OK — no re-export needed |
| mirror-baseline freshness | `git diff --name-only <BASE> <CARD_BASE> -- internal/template/ .claude/rules/moai/ > snapshot-stale-mirror.txt` | exit 0, `test -e` 0, `test -s` 1 | record | empty — mirror baseline current |

## Step 2 — base exports

`git show 255f88eb0…:<path> > base-<name>[-template].md` for the 10 local and 10 template paths plus
`base-manager-git.toml` and `b412-agent-common-protocol.md` (from `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0`);
every call exit 0. Line counts: `pre2-export-linecounts.txt`. Scope roots in the worktree equal
BASE: `git diff --stat <BASE> -- internal/template/templates/ .claude/ .agents/ Makefile` → 0 bytes
(`pre2-worktree-vs-base-scope.txt`).

## Step 3 — positive controls (RED cells) on base copies

| AC | Control | Observed | Expected | Hit? |
|---|---|---|---|---|
| 001 | Sync section paragraphs with fetch+rev-list / autofail (`ac001-base-*`) | 1 / `test -s` 0 | 1 / exit 0 | yes |
| 002 | order judge on base block (`ac002-base-order.txt`) | `fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS`; pass-count 1; session 1; section 52 lines; L/T base section diff exit 0; matrix 8 rows | same | yes |
| 003 | b412 vs BASE Pre-Spawn section diff (`ac003-control.diff`) | exit 1 | exit 1 | yes |
| 004 | `gh pr merge --squash --delete-branch` count / `merge_method` grep (`ac004-control-*`) | 2 and 2 / exit 1 and 1 | 2 / exit 1 | yes |
| 005 | squash lines (`ac005-control.txt`) / zero set (`ac005-control-zero.txt`) | 6 lines (mg 34·116, dl 368·380, toml 26·108) / all 0 | 6 / 0 | yes |
| 006 | dl default hits / de hit / source / optin (`ac006-base-*`) | 8·20 / exit 0 (line 8) / 0·0 / 1·1 | same | yes |
| 014 | base toml sections vs template md (`ac014-base-*.diff`) | both diff exit 0 (11 and 9 lines); toml example count 0 | exit 0; 0 | yes |
| 015 | spec controls / fixture (`ac015-control-*`, `ac015-fixture-*`) | specid 8, req 61, date 23, sha-980 1 (105 tokens) / letter 5, digits 2 | ≥1 / 5, 2 | yes |
| 016 | `<BASE> --not b412…` acp commits (`ac016-control.txt`) | 2 lines (97ef8e302…, 6896eef37…) | ≥1 | yes |
| 026 | fragment lines / (i) / (ii) / (iii) (`ac026-base-{local,template}-*`) | 1,3,4,6,1,52,1,1,24 / all 0 / 10 lines / 1,1,1,2; fragments L=T | same | yes |
| 027 | `--no-merge` dl lines (`ac026-base-*-027-*`) | (i) 8·19, (ii) 8·19 both copies | same | yes |
| 028/029 | (`ac028-029-base-controls.txt`) | 028a 0·0, 028b 0, 029a 0·0, 029b 0 both copies; mg section 9 lines, L/T diff 0 | same | yes |

## Step 4 — generator checks

`make agents-emit-check` exit 0 (`pre4-agents-emit-check.*`); `make commands-emit-check` exit 0
(`pre4-commands-emit-check.*`).

## Step 5 — AC-GDP-030 publish-check positive control (c)

Backup to the session scratchpad (not `/tmp`; the backup path is not load-bearing), `printf '\n'`
appended to the template published skill, then:

| Step | Result | Expected |
|---|---|---|
| `make commands-emit-check > ac030-red.txt` | make exit **2**; test output `--- FAIL: TestGoldenCommittedArtifactsMatchEmission … sha256 mismatch`, recipe line `make: *** [commands-emit-check] Error 1`, drift message printed | "exit 1" |
| restore + `cmp` | exit 0 (`ac030-restore-cmp.exit`) | 0 |
| `make commands-emit-check > ac030-control-green.txt` | exit 0 | 0 |
| `git status --porcelain -- <T>/.agents/skills/ .agents/skills/` | `test -e` 0, `test -s` 1 | empty |
| `git ls-files` published pair | 2 lines (`ac030-tracked-count.txt`) | 2 |

Deviation recorded, not a missed control: acceptance.md writes the red expectation as "exit 1", but
`make` itself returns 2 whenever a recipe fails; the recipe's own status is the `Error 1` on the make
line. The drift was detected (FAIL line + drift message), which is what the control exists to show.
The same `make` convention will apply to AC-GDP-014's "exit 1" RED step in M4.

## Step 6 — pair-diff baseline (BASE copies)

`pre6-pair-baseline.txt`: acp and published exit 0; delivery, doc-execution, qgc, skill, reference,
sync, command-sync, manager-git exit 1 with hunk heads identical to
`.moai/reports/t622/reanchor/pair-hunks.txt` (checked line by line: delivery 9, doc-execution 9,
qgc 7, skill 20, reference `229d228`, sync `29,31c29,31 65,74d64 81c71`, command-sync `2c2`,
manager-git `5,7c5`).

## Step 7 — Frozen baseline

`pre7-frozen-acp.txt`: CONST-V3R2-006/036/037/038 counts 1·1·1·1 in both acp copies, `[ZONE:Frozen]`
at line 17 in both. `[ZONE:` lines in the other scope files: all 0 (`pre7-zone-others.txt`,
`pre7-zone-mg.txt`). Registry control 13·13 (`pre7-registry-control.txt`), others 0·0
(`pre7-registry-others.txt`), registry detector mutant fixture → 2, Frozen-line judge mutant → 2.

## Step 8 — mirror-test baseline

`go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v`
→ exit 1 (`mirror-pre.txt`, `mirror-pre.exit`). Sets sorted and compared with
`.moai/reports/t622/reanchor/mirror-baseline-sets.txt`: `diff` exit 0
(`mirror-pre-vs-baseline.diff` empty). PASS 17, FAIL {`TestRuleTemplateMirrorDrift`,
`TestRuleTemplateMirrorDrift/spec-workflow.md`}.

## Cf counter proof

`cf-control-2.txt` carries exactly two planted Cf characters (U+200B, U+FEFF);
`perl -CSD -ne '$c+=()=/\p{Cf}/g; END{print $c+0,"\n"}'` → 2 (`cf-control-2.count`); the multi-file
form used for the edited files → 2 on the control and 0 on a base copy (`cf-counter-control-multi.txt`).

Pre-flight verdict: every positive control showed its expected hit (the `make` exit-code nuance above
included), the range control is non-empty, both freshness checks are empty, and the sorted mirror
sets match. Editing was permitted.
