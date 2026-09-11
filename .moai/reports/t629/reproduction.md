# t629 — reproduction of SX-R01 on develop (secret scan misses unmerged refs)

card: t629 (Class C)
worktree: .claude/worktrees/t629
branch: WT-secret-scan-refs
card base: feeecc980 (no-ff absorb; HEAD^2 = local develop 80e9e0039)
source finding: instruction audit 20260910, SX-R01 (P1), measured against main 2213871af
measured: 2026-09-10, lane-5

## Claim

1. **The defect is live on develop.** The incremental secret scan in
   `.claude/skills/moai/workflows/review.md` scans `<last-sha>..HEAD`. A credential-shaped line
   committed on a branch that is not reachable from HEAD is never reported by it — not on the
   next review, and not on any later review while that branch stays unmerged.
2. **The scanner itself works.** The same command reports a credential-shaped line committed on
   the HEAD line. The side-branch miss is therefore a coverage defect, not a dead scanner. The
   audit's own evidence recorded only the side-branch cell; this control is added here.
3. **The document's equivalence claim is false.** `review.md` line 110 states the incremental range
   plus checkpoint "produces the same coverage over time as the former every-review full scan — no
   finding class is dropped". Measurement 2 below is the direct counterexample.
4. **The blind spot is exactly: refs unreachable from HEAD at scan time that are never merged.**
   Once the side branch is merged, the next incremental scan reports it (measurement 3).

## Evidence — the text on develop

The two copies are byte-identical and the secret-scan section is unchanged since the audit's tree:

```
$ wc -l .claude/skills/moai/workflows/review.md internal/template/templates/.claude/skills/moai/workflows/review.md
     508 .claude/skills/moai/workflows/review.md
     508 internal/template/templates/.claude/skills/moai/workflows/review.md
$ diff -q <local copy> <template copy> ; echo REVIEW_COPIES_DIFF_EXIT=$?
REVIEW_COPIES_DIFF_EXIT=0
$ git log --oneline 2213871af..HEAD -- <both copies>
918d65366 docs(SPEC-AUDIT-PARTICIPANT-COUNT-001): sync-phase artifacts + 3-phase close (t284)
$ git diff -U0 2213871af HEAD -- .claude/skills/moai/workflows/review.md   (hunk headers)
@@ -186,0 +187 @@ Folding the convergence result into the review verdict:
```

The only change since main is one line at 187, in the convergence section — not the secret-scan
section at 93-110. The card noted develop "may already be fixed"; it is not.

The relevant lines as they stand:

- 93: `#### Secrets Scan (Incremental with Checkpoint)`
- 97-100: where a checkpoint SHA exists, run `git log -p <last-sha>..HEAD -G '<regex>'`, then move
  the checkpoint to the current HEAD
- 103-106: only on first run or an explicit full-scan flag, run `git log -p --all -G '<regex>'`
- 110: the equivalence claim quoted above

The checkpoint exists only as prose: `secrets-scan-checkpoint` has zero matches in Go source under
`internal/`, `pkg/`, `cmd/` (grep exit 1) and appears only in these two `review.md` copies. The fix
surface is therefore the two document copies, not code.

## Evidence — fixture

A throwaway git repository in the session scratchpad, outside this repository (so no
credential-shaped string enters this repository's history). Configured with a fixture identity,
`commit.gpgsign=false`, and `core.hooksPath=/dev/null` so no hook ran. Every scan uses the exact
regex from `review.md`: `(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})`.

Markers are synthetic PEM-style private-key header lines carrying two distinct labels, `HEADCELL`
and `SIDECELL`, so each cell can be counted independently in one scan. Both labels match
`[A-Z]+`. The literal lines are deliberately not reproduced in this file.

Construction:

| step | commit | branch | content |
|---|---|---|---|
| C0 | `89ccdd332eb1f41d679ac534d0794d0cf008f137` | main | clean file — the checkpoint |
| S1 | `dfa7a0582a0d5b88ce142b867e14818e4c4d3e6f` | side (from C0) | SIDECELL marker |
| C1 | `396ec722aca37bcb93ee0007deba812087d4aae4` | main | HEADCELL marker |
| C2 | `37c1352` | main | clean file |
| C3 | merge of side into main | main | — |

`git merge-base --is-ancestor side HEAD` after C1 → exit 1: S1 is not reachable from HEAD.

Every scan's output was redirected to a file, its exit code read with no pipe, and the file then
measured (`wc -c`) and counted (`grep -c` per label).

| measurement | range | incremental scan | `--all` scan |
|---|---|---|---|
| baseline control, before any marker | — | — | exit 0, **0 bytes** |
| **1** — checkpoint C0 | `89ccdd332..HEAD` (HEAD = C1) | exit 0, 395 bytes, **HEADCELL 1**, **SIDECELL 0** | exit 0, 796 bytes, HEADCELL 1, SIDECELL 1 |
| **2** — next review, checkpoint moved to C1 | `HEAD~1..HEAD` (C1..C2) | exit 0, **0 bytes**, **SIDECELL 0** | exit 0, 796 bytes, SIDECELL 1 |
| **3** — after merging side | `HEAD~1..HEAD` (C2..C3) | exit 0, 400 bytes, **SIDECELL 1** | — |

Reading the rows:

- Baseline 0 bytes: the regex matches nothing in the clean baseline, so every later match is a marker.
- Measurement 1, HEADCELL 1 in the incremental scan: the scanner is alive. This is the control the
  card requires; without it, SIDECELL 0 could not be told apart from a scanner that reports nothing.
- Measurement 1, SIDECELL 0 incremental versus 1 under `--all`: the defect, in the audit's shape
  (audit: 0 bytes versus 365 bytes).
- Measurement 2, SIDECELL 0 again: advancing the checkpoint does not recover the miss. Over time
  the incremental scan never reaches the unmerged side branch — refuting line 110.
- Measurement 3, SIDECELL 1: merging brings S1 into the range, so the blind spot is specifically
  refs that are unreachable from HEAD and stay unmerged.

## Baseline-attribution

The document readings were taken on this worktree at HEAD feeecc980, whose tree equals local
develop 80e9e0039. The fixture was built and measured in this session. The audit's figures
(0 / 365 bytes, from its own fixture against main) are cited as the finding being reproduced, not
reused as measurements here.

## Gaps

- No real credential leak was looked for or found; the fixture is synthetic by design.
- Only one kind of unreachable ref was exercised: a local branch. Remote-only branches, tags, stash
  entries, and branches deleted after being pushed were not built; they are expected to behave the
  same because `--all` follows refs and `<last-sha>..HEAD` follows only HEAD's ancestry, but that is
  inferred, not measured.
- The cost of widening the scan on this repository is not in this file; it is measured separately.
- Whether any agent actually runs this section, and how often, was not observed.

## Residual-risk

The regex has three alternatives; only the PEM header alternative was exercised. The coverage
defect is about the commit range, not the pattern, so the other two are expected to miss in the
same way — not measured.
