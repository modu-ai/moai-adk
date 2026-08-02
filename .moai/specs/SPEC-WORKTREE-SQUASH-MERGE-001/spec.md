---
id: SPEC-WORKTREE-SQUASH-MERGE-001
title: "worktree clean: detect squash-merged branches so --stale reclaims them"
version: "0.6.1"
status: in-progress
created: 2026-08-02
updated: 2026-08-02
author: manager-spec
priority: P1
phase: "v3.0.x"
module: "internal/core/git, internal/cli/worktree"
lifecycle: spec-anchored
tags: "worktree, git, cleanup, squash-merge, patch-id, disk-reclaim"
tier: M
---

# SPEC-WORKTREE-SQUASH-MERGE-001 — Squash-merge detection for worktree cleanup

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-02 | Initial plan-phase authoring. Root cause, detector matrix, and all four scope decisions resolved from executed git experiments. | manager-spec |
| 0.2.0 | 2026-08-02 | Iteration-2 revision clearing plan-audit iteration 1 (FAIL, 0.62). Added the S5 state conjunct so S3/S4 are no longer sufficient alone, closing the partially-reverted false positive; re-ran the detector matrix at nine scenarios × five signals. Narrowed REQ-WSM-007 to name per-probe verdict codes. Added REQ-WSM-014. Recorded the clean-worktree-with-unpushed-commits removal case in §5 decision 3 and §4. | manager-spec |
| 0.3.0 | 2026-08-02 | Iteration-3 revision clearing plan-audit iteration 2 (FAIL, 0.75). Closed the rename-fold false positive: the S5 path enumeration now uses `git diff --no-renames --name-only`, and the silent-failure mechanism behind it (a pathspec matching nothing makes `git diff --quiet` exit 0) is recorded in REQ-WSM-007's table and REQ-WSM-014. Matrix re-run at eleven scenarios × five signals under both settings; SC-10 (rename + re-add, squash) and SC-11 (rename + re-add, cherry-pick) added. `acceptance.md`: AC-WSM-005 consolidated to cover SC-6 and SC-7 together, AC-WSM-006 repurposed as the rename criterion, and the non-vacuity guard extended to AC-WSM-001, AC-WSM-010, and AC-WSM-015. | manager-spec |
| 0.4.0 | 2026-08-02 | Iteration-4 revision, authorized by explicit user override of the three-iteration ceiling and **scoped to one finding**. Closed the path-encoding false positive: `git diff --name-only` C-quotes any path containing a backslash, tab, double quote, or control character — and, under git's documented default `core.quotePath=true`, any non-ASCII path — and the quoted string fed back as a pathspec matches nothing, which §2 finding 5 shows reads as merged. The S5 enumeration is now `git diff --no-renames --name-only -z` split on NUL. Matrix extended to twelve scenarios with SC-12 (non-ASCII path); SC-1..SC-11 re-run and cell-for-cell unchanged. REQ-WSM-007's table and REQ-WSM-014 record the encoding mechanism, including that `core.quotePath=false` is **not** a sufficient mitigation. `plan.md` §D's sentence asserting that a `strings.Split(out, "\n")` path list is safe — an affirmatively wrong claim — is corrected. `acceptance.md`: AC-WSM-006 extended with the `newline-split` falsification rather than adding a seventeenth criterion (Tier M AC ceiling), §C.1 grid re-run at 6 mutations × 12 scenarios, and run-phase judging commands added to AC-WSM-002..AC-WSM-008 so §F's per-criterion recording requirement holds as written. | manager-spec |
| 0.6.0 | 2026-08-02 | Iteration-6 revision, adopting the conclusions of a commissioned OID-comparison-hybrid study. Closed the **comparison-semantics** false positive (D1/N10): S5's own notion of "difference" is configurable, so `git diff --quiet` can report equality for genuinely different content even when the path list is complete, byte-perfect, and matched literally. The S5 comparison now carries `--no-textconv` and `--ignore-submodules=none`, and the S5 **enumeration** now carries `--ignore-submodules=none` as well — the enumeration-side flag is **not** redundant, because two of the axis's three members suppress at the enumeration stage. §2's axis map gains a fifth axis with all three measured members and their stage classification. **The closure sentence at v0.5.0 ("no open axis remains — there is no sixth patch queued behind this one") is deleted**: it was false by measurement, and an affirmative false claim in the unsafe direction was the iteration-3 and iteration-5 FAIL cause. It is replaced by a bounded statement naming what was probed and what the probe found. Matrix extended to seventeen scenarios with SC-14 (textconv), SC-15 (submodule pointer under `submodule.<n>.ignore=all`), P3c (mode-only divergence), and P4c (symlink/file blob-OID collision); SC-1..SC-13 re-run and cell-for-cell unchanged. The commissioned hybrid was built, attacked, and **rejected** on measured grounds; that outcome is recorded so the option is not re-proposed a third time. `acceptance.md`: AC-WSM-006 absorbs SC-14 and SC-15 with two further falsifications, a new AC-WSM-017 pins mode sensitivity via P3c and P4c, and §C.1's grid is re-run at 11 mutations × 17 scenarios = 187 cells. | manager-spec |
| 0.5.0 | 2026-08-02 | Iteration-5 revision, adopting the conclusions of a commissioned reformulation-viability study. Closed the pathspec-**interpretation** false positive (N8): a path whose first byte is `:` round-trips byte-perfectly under `-z` and is still parsed as pathspec magic, so S5 holds vacuously. The S5 comparison now carries `--literal-pathspecs` as a git-level option, which disables **all** pathspec magic and globbing by definition rather than enumerating hazardous symbols. Matrix extended to thirteen scenarios with SC-13 (leading-colon path); SC-1..SC-12 re-run and cell-for-cell unchanged. §2's honest-limits passage is rewritten from an instance count to the **four-axis** framing — completeness, argv, encoding, interpretation — recording that all four are now closed and that glob metacharacters and backslashes were measured to round-trip correctly, so `:` was the only open symbol class. The study built the named reformulation, attacked it, and **rejected** it on measured grounds (scenario A1b); that outcome is recorded so the option is not silently re-proposed. `plan.md` §D gains the no-shell requirement as a checked property. `acceptance.md`: AC-WSM-006 extended with the `no-literal` falsification and a widened selector rather than adding a seventeenth criterion (Tier M AC ceiling), §C.1 grid re-run at 7 mutations × 13 scenarios, SC-10's recipe disambiguated to "in two commits", and the shared `diff.ignoreSubmodules` exposure recorded as a stated limitation. | manager-spec |
| 0.6.1 | 2026-08-02 | Pre-M1 plan-artifact pin; no mechanism change. Plan-audit iteration 6 (PASS, 0.89) reported D1 — a **sixth axis, the invocation's frame of reference**: the S5 enumeration emits repo-root-relative paths while the comparison's pathspecs resolve against the process working directory, so a cwd below the repository root turns every path into an unmatched pathspec and finding 5 reads that as merged (measured under default git configuration — no `.gitattributes`, no `.gitmodules`, no config key: `rc=1 / S5=NO` at the root, `rc=0 / S5=OK` from a subdirectory). Judged **not a live defect**: `plan.md` §C commits the implementation to the existing `execGit(ctx, w.root, …)` helper, which sets `cmd.Dir` to `repo.Root()` (`rev-parse --show-toplevel`), so the mechanism as it will be built is correct — what was missing is the **pin**, since §F pinned every property of the argument vector and nothing about the working directory. Therefore pinned rather than fixed, in exactly two additions: an `acceptance.md` §F Definition-of-Done item binding each invocation's working directory to the repository root, and a sixth row in `spec.md` §2's axis table recording the failure, the closure, and that `--literal-pathspecs` forecloses the in-band `:(top)` pathspec remedy. §2's axis count and instance count are updated to six and seven respectively; the bounded "no *known* open axis" framing is unchanged and no completeness claim is introduced. No acceptance criterion was added or altered, the seventeen-scenario matrix and the 187-cell §C.1 grid are untouched, and no measurement changes — so nothing requires re-running. | manager-spec |

---

## §1 Context and Problem

`moai worktree clean --stale` sweeps abandoned worktrees that hold nothing to lose. Its
second safety condition — "the branch carries no commits of its own beyond base" — is
evaluated by `WorktreeProvider.IsBranchMerged`, whose sole implementation runs:

```
git branch --merged <base>
```

That command is a **reachability** test: a branch is listed only when its tip commit is an
ancestor of `base`. A squash merge collapses N branch commits into ONE NEW commit on
`base`; the original commits never become ancestors. `IsBranchMerged` therefore returns
`false` forever for every squash-merged branch, `staleKeepReason` returns
`"branch has commits not in <base>"`, and the worktree is silently kept. No error is
raised, so the failure is invisible.

This repository auto-merges every pull request with **squash**, so squash is the default
merge path rather than an exception. The blind spot covers essentially every normally
merged branch.

### Observed impact

Across 45 worktree branches on this repository: 4 recognized as merged by the current
predicate, 33 actually merged but missed (squash), 8 genuinely unmerged — a reclaim rate
near 11%. Disk held: `.claude/worktrees` 5.9G plus `~/.moai/worktrees` 1.3G.

Reproduction (executed against this repository):

```
$ git branch --merged origin/main | grep -c 'fix/worktree-redesign-m0m1'
0          # squash-merged via PR #1278, yet absent from the listing
```

### Affected call sites

| Location (content anchor) | Path |
|---|---|
| `IsBranchMerged(branch, base string) (bool, error)` interface method | `internal/core/git/types.go` |
| `func (w *worktreeManager) IsBranchMerged` implementation | `internal/core/git/worktree.go` |
| `cleanMergedWorktrees` — the `--merged-only` path | `internal/cli/worktree/clean.go` |
| `staleKeepReason` — the `--stale` path | `internal/cli/worktree/clean.go` |

---

## §2 Detector Evidence (executed)

Seventeen merge scenarios were constructed as isolated git repositories and each was probed
with five candidate signals. Every cell below is an observed result, not an inference.

The signals:

| Id | Signal | Probe |
|---|---|---|
| S1 | reachability | `git branch --merged <base>` lists `<branch>` |
| S2 | empty diff | `git diff --quiet <merge-base> <branch>` reports no difference |
| S3 | rebase-merge (history) | plain `git cherry <base> <branch>`, non-empty and every line prefixed `-` |
| S4 | squash-merge (history) | synthetic-commit `git cherry`, non-empty and every line prefixed `-` |
| S5 | **state reproduction** | every path the branch touched since its merge-base — enumerated with `git diff --ignore-submodules=none --no-renames --name-only -z`, so that a gitlink the branch moved is included (`--ignore-submodules=none`), a path the branch *deleted* or renamed away from is included (`--no-renames`), and every path is emitted verbatim rather than C-quoted (`-z`) — then fed back under `--literal-pathspecs` so the bytes are matched as a literal filename rather than parsed as pathspec magic, and compared under `--no-textconv --ignore-submodules=none` so the comparison cannot be configured to overlook a real difference — has identical content in base HEAD |

S3 and S4 are recorded below as **raw** history signals, before the S5 conjunction; the
composed verdict is `S1 OR S2 OR (S3 AND S5) OR (S4 AND S5)`.

| Id | Scenario | S1 | S2 | S3 | S4 | S5 | Required verdict |
|---|---|---|---|---|---|---|---|
| SC-1 | squash-merged | keep | keep | keep | **MERGED** | **OK** | **merged** |
| SC-2 | rebase-merged (individual replay) | keep | keep | **MERGED** | keep | **OK** | **merged** |
| SC-3 | true merge commit | **MERGED** | **EMPTY** | keep | keep | OK | **merged** |
| SC-4 | strictly behind base | **MERGED** | **EMPTY** | keep | keep | OK | **merged** |
| SC-5 | empty-diff branch (empty commit only) | keep | **EMPTY** | keep | keep | OK | **merged** |
| SC-6 | partially applied | keep | keep | keep | keep | **NO** | **not merged** |
| SC-7 | fully unmerged (control) | keep | keep | keep | keep | **NO** | **not merged** |
| SC-8 | **partially reverted** | keep | keep | **MERGED** | keep | **NO** | **not merged** |
| SC-9 | base is a strict superset | keep | keep | **MERGED** | **MERGED** | OK | **merged** |
| SC-10 | **rename + re-add, squash-merged** | keep | keep | keep | **MERGED** | **NO** | **not merged** |
| SC-11 | **rename + re-add, cherry-picked** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-12 | **non-ASCII path removed from base** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-13 | **leading-colon path removed from base** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-14 | **textconv driver collapses a changed file** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-15 | **submodule pointer moved, `submodule.<n>.ignore=all`** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| P3c | **mode-only divergence (chmod)** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| P4c | **symlink/file blob-OID collision** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |

SC-10 and SC-11 are the rename pair. In both, the branch renames a file that exists at the
merge-base, base takes the branch's work, and base *later re-adds the original path*. The
branch's deletion of that path therefore does not survive in base HEAD, so the required
verdict is not-merged. Their S5 column reads `NO` **only** because the enumeration uses
`--no-renames`; under git's default rename detection S5 reads `OK` in both rows and the
composed verdict flips to a false positive. See finding 5 and Falsified alternative 3.

SC-12 is the encoding case, and it is a *third* route to the same failure. The branch adds
a file whose name is non-ASCII (`문서.txt`); base squash-merges the branch and then removes
that file, so the branch's content does not survive in base HEAD and the required verdict
is not-merged. Its S5 column reads `NO` **only** because the enumeration is NUL-delimited
(`-z`); split on newline, the path arrives C-quoted as `"\353\254\270\354\204\234.txt"`,
that literal matches no file, and S5 reads `OK`. Both S3 and S4 fire here, so the
enumeration is again the only thing standing between two agreeing history signals and a
wrong removal. See finding 6 and Falsified alternative 4.

SC-13 is the *interpretation* case, and it is a fourth route to the same failure — the one
that survives every repair above. The branch adds a repo-root file named `:note.txt`; base
squash-merges the branch and then removes that file, so the required verdict is not-merged.
Its path list is byte-perfect under `-z` — the enumeration emits `:note.txt` exactly — and
the probe still fails, because git parses a leading `:` in a *pathspec* as magic before any
matching happens. Its S5 column reads `NO` **only** because the comparison carries
`--literal-pathspecs`; without it the pathspec names no file and S5 reads `OK`. Both S3 and
S4 fire here as well. See finding 7 and Falsified alternative 5.

SC-14 and SC-15 are the *comparison-semantics* pair, and they are a fifth route to the same
failure — the first one that does not touch the path list at all. In SC-14 the branch adds
`report.pdf` under a `.gitattributes` rule binding a `textconv` diff driver; base
squash-merges and then overwrites the file with different content. The path list is complete
and byte-perfect, and the comparison still reports equality, because the driver renders both
blobs to the same text. In SC-15 the branch moves a submodule pointer with **no**
`.gitmodules` change; base squash-merges and then moves the pointer elsewhere. Under
`submodule.<name>.ignore=all` the gitlink is dropped at *both* stages — it never enters the
enumeration, and the comparison would ignore it even if it did. Their S5 columns read `NO`
**only** because the comparison carries `--no-textconv --ignore-submodules=none` and the
enumeration carries `--ignore-submodules=none`. Both S3 and S4 fire in each. See finding 8
and Falsified alternative 6.

P3c and P4c are not repairs — they are **retention** rows. The design as specified already
reports both correctly, and they are recorded because they are the only two scenarios in the
matrix whose verdict depends on the comparison being sensitive to a file's **mode**. In P3c
the branch chmods a file to 755 and base chmods it back; in P4c the branch replaces a symlink
with a regular file whose content is exactly the symlink's target, and base reverts it. In
both, the two sides share **one blob OID** and differ only in mode:

```
P3c   feat 100755 blob 587be6b4c3f93f93c489c0111bba5596147a26cb  x.txt
      main 100644 blob 587be6b4c3f93f93c489c0111bba5596147a26cb  x.txt
P4c   feat 100644 blob 1de565933b05f74c75ff9a6520af5f9f8a5a2f1d  link.txt
      main 120000 blob 1de565933b05f74c75ff9a6520af5f9f8a5a2f1d  link.txt

git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none …   rc=1  (detects)
blob-OID-only comparison over the same two paths                                EQUAL (misses)
```

`git diff` compares modes, so the specified mechanism passes both. Any future edit that
swaps the comparison for something mode-blind — a blob-OID comparison being the obvious
candidate, and the one an earlier proposal reached for — turns both rows into silent
fail-open false positives under **default** git configuration, with no `.gitattributes` and
no config key involved. That is a strictly worse exposure than any axis-5 member, all of
which require non-default state. AC-WSM-017 exists to fail if that swap is made.

Eight findings govern the design:

1. **The signals are complementary, not competing.** No single signal covers all merged
   scenarios. `git branch --merged` is the only one that recognises a true merge commit and
   a strictly-behind branch; the synthetic-commit probe is the only one that recognises a
   squash merge. Replacing the existing reachability test rather than extending it would
   regress the two scenarios it already handles correctly.
2. **S3 and S4 are history signals and are not sufficient alone.** Both ask whether an
   equivalent patch-id existed at *some point* in base's history since the merge-base. They
   do not ask whether the content *survives* in base HEAD. SC-8 separates the two: a
   later revert leaves the patch-ids matched while the content is gone, and S3 fires on a
   branch whose work is genuinely absent. S3 and S4 are therefore each conjoined with S5,
   which asks the state question directly. S5 is never a merged verdict on its own — it is
   only ever a conjunct, so it can withhold a verdict but never grant one.
3. **SC-9 is pinned to one construction, and the row does not generalise.** The recorded
   row is produced by exactly the `acceptance.md` §B recipe: base re-creates the branch's
   change as its own commit (`main` adds `s.txt`, then additionally adds `extra.txt`).
   That is the construction the cells describe, and it is the only one they describe.
   The superficially similar phrasing "base cherry-picks the branch's commit and then adds
   more" is **not** an equivalent recipe and MUST NOT be substituted: a cherry-pick
   performed with no elapsed time reproduces the branch commit's tree, parent, message, and
   identity, so the two commits collide on SHA, the branch becomes a literal ancestor of
   base, the merge-base moves, and the scenario silently degenerates into SC-4 (strictly
   behind) with S1 firing. `acceptance.md` §B carries the same warning about identical
   commit metadata; it applies here. A superset built so that base's version of the change
   has a *different* patch-id yields no history signal and therefore a not-merged verdict —
   a harmless false negative, since the worktree is merely kept.
4. **Across this seventeen-scenario matrix, no false positive survives.** The eleven
   scenarios that must be kept (SC-6 partially applied, SC-7 fully unmerged, SC-8 partially
   reverted, SC-10 and SC-11 rename + re-add, SC-12 non-ASCII path, SC-13 leading-colon
   path, SC-14 textconv, SC-15 submodule-ignore, P3c mode-only, P4c symlink/file collision)
   are reported as "keep" by the composed predicate. This statement is scoped to the
   constructed matrix; it is not a claim of general correctness over all possible histories.
   **This claim has been read as general and falsified five times, and the record of that is
   the reason the scoping sentence above is load-bearing rather than decorative.** SC-8
   exists because the earlier seven-scenario version of this claim was falsified; SC-10 and
   SC-11 exist because the nine-scenario version was falsified the same way; SC-12 because
   the eleven-scenario version was falsified by attacking the *output encoding* of the
   enumeration the previous repair introduced; SC-13 because the twelve-scenario version was
   falsified by attacking how that byte-perfect output is *parsed* when handed back; SC-14
   and SC-15 because the thirteen-scenario version was falsified by attacking the
   *comparison* that received a list correct on every prior axis. The first four
   falsifications share one shape — S5's input did not name what it was supposed to name.
   The fifth has a different shape: S5's input was correct and S5 itself was configured not
   to notice. The axis map below records both shapes and which repair closed each.
5. **A pathspec that matches nothing is silently read as "no difference".** `git diff
   --quiet <a> <b> -- <pathspec-matching-nothing>` exits **0**, not non-zero (measured
   below and recorded in REQ-WSM-007's table). S5 therefore fails *open*: any defect that
   makes the enumerated path list under-inclusive does not surface as an error — it
   surfaces as a merged verdict. Omission becomes a pass, in the unsafe direction. This is
   the mechanism behind Falsified alternative 3, and it is why REQ-WSM-014 states the
   enumeration's completeness as a requirement rather than leaving it to the mechanism.
6. **`git diff --name-only` does not emit paths verbatim — it C-quotes them.** A path
   containing a backslash, a tab, a double quote, or a control character is emitted inside
   double quotes with C-style escapes, **unconditionally**; and under git's documented
   default `core.quotePath=true`, so is any path containing a non-ASCII byte. The quoted
   string is not the path: fed back as a pathspec it matches nothing, and by finding 5 that
   reads as "no difference". This is a distinct defect from the rename fold — the fold
   *omits* a path, the quoting *corrupts* one — but it lands in the same unsafe direction,
   and it survived the `--no-renames` repair untouched because the two are orthogonal.
   `core.quotePath=false` is **not** a sufficient mitigation: it suppresses quoting only for
   the non-ASCII case, while backslash and control characters stay quoted regardless
   (measured below). The mechanism that emits paths verbatim in every case is `-z`, which
   NUL-delimits and never quotes. This is the mechanism behind Falsified alternative 4, and
   it is why REQ-WSM-014 binds the enumeration's *encoding* alongside its completeness.
7. **A byte-perfect path is still not a filename — git parses a pathspec before matching
   it.** `-z` guarantees the bytes leaving the enumeration are the path's own bytes. It says
   nothing about how those bytes are read when handed back, and git reads a pathspec's
   leading characters as *magic* before any matching occurs. A repo-root entry whose first
   byte is `:` is therefore parsed as a magic prefix rather than as a filename: `:root.txt`
   and `:(glob)g.txt` match nothing, while `:!bang.txt` and `:^caret.txt` are worse than
   unmatched — they become EXCLUDE pathspecs and silently invert what the probe compares.
   This is a distinct defect from both the rename fold and the quoting: the fold *omits* a
   path, the quoting *corrupts* one, and this *re-interprets* one that arrived intact.
   Measured, in a repository where the correct answer is `rc=1` in every row:

   ```
   bare pathspec (as specified through v0.4.0)
     :root.txt        rc=0   <- parsed as magic, matched NOTHING
     :(glob)g.txt     rc=0   <- parsed as long-form :(glob) magic + path g.txt
     :!bang.txt       rc=1   but as EXCLUDE magic — the probe's comparison is inverted
     :^caret.txt      rc=1   but as EXCLUDE magic — likewise
     dir/:nested.txt  rc=1   safe (the colon is not the pathspec's first byte)
     mid:colon.txt    rc=1   safe

   --literal-pathspecs — all six round-trip as filenames, all rc=1
   ```

   The exposure is bounded: only repo-root entries whose **first byte** is `:`, and `:` is
   illegal in Windows filenames, so it is POSIX-only. The mechanism that closes it is
   `--literal-pathspecs`, which disables **all** pathspec magic and globbing by definition
   rather than enumerating hazardous symbols. This is the mechanism behind Falsified
   alternative 5, and it is why REQ-WSM-014 binds *round-trip* fidelity rather than encoding
   fidelity alone.
8. **A correct path list is still not a correct answer — git's notion of "difference" is
   itself configurable.** Findings 5 through 7 all concern the path list: whether it is
   complete, whether its bytes survive, whether those bytes are read as a filename. This
   finding concerns what happens *after* all three succeed. `git diff --quiet` decides
   equality under the repository's diff configuration, and that configuration can be
   narrowed until it reports equality for genuinely different content. Three members were
   measured, and they do **not** act at the same stage:

   ```
   member                                  enumeration     comparison    closed by
   .gitattributes diff=X + diff.X.textconv  unaffected      SUPPRESSED    --no-textconv
   diff.ignoreSubmodules=all                SUPPRESSED      SUPPRESSED    --ignore-submodules=none
   submodule.<name>.ignore=all              SUPPRESSED      SUPPRESSED    --ignore-submodules=none
   ```

   The stage classification is the load-bearing part, because it decides where the flags go.
   textconv is **comparison-side only** — measured, the name-listing path never runs the
   driver, so the enumeration still reports the changed file and only `--quiet` is fooled:

   ```
   git diff --name-only  feat main -- report.pdf   ->  report.pdf   (still reported)
   git diff --quiet      feat main -- report.pdf   ->  rc=0         <- ONLY this is fooled
   git diff --quiet --no-textconv feat main -- report.pdf -> rc=1
   ```

   The two submodule keys act at **both** stages: they drop the gitlink from
   `git diff --name-only` itself, so the path never reaches the comparison at all. Measured
   on SC-15, where the correct path list is `[plain.txt, sub]`:

   ```
   git diff --no-renames --name-only -z <mb> feat                        ->  [plain.txt]
   git diff --ignore-submodules=none --no-renames --name-only -z <mb> feat -> [plain.txt, sub]
   ```

   That is why `--ignore-submodules=none` is required on the **enumeration** as well as the
   comparison, and why removing it from either stage alone reproduces the false positive
   (both mutations flip SC-15; `acceptance.md` §C.1). A future edit that deletes the
   enumeration-side flag as redundant with the comparison-side one re-opens SC-15.

   `submodule.<name>.ignore` matters more than its sibling because it does not need the
   operator's own config: the key is readable from `.gitmodules`, which is a **tracked
   file**, so a repository can ship this exposure to every clone. Measured — with
   `.git/config` carrying no such key and the directive committed inside `.gitmodules`
   instead, SC-15 reproduces identically. It shares textconv's worktree-state dependence:
   both values are read from the checked-out file, not from the tree under judgement, so the
   verdict depends on the state of the judging worktree. Removing the worktree's
   `.gitattributes` makes SC-14's false positive vanish while the file is still committed —
   a sensitivity none of the first four axes had. This is the mechanism behind Falsified
   alternative 6, and it is why REQ-WSM-014 binds the *observed comparison* rather than the
   path list alone.

### Falsified alternative 1 — plain `git cherry <base> <branch>` as the squash detector

Plain per-commit `git cherry` is the most plausible-looking answer and is **wrong for the
squash case**. A squash merge collapses N commits into one, so the individual per-commit
patch-ids never match the single squashed commit's patch-id. On the two verification
branches it reported 8 and 2 commits respectively as "not applied" — a false negative in
both. It is retained in this design only as the **rebase-merge** signal, where it is
correct; it must never be used alone as the squash detector.

### Falsified alternative 2 — history-only patch-id equivalence (no state check)

A predicate composed of S1/S2/S3/S4 alone reports SC-8 as **merged** while `b.txt` is
genuinely absent from base. Observed:

```
main tree      : a.txt seed.txt
feat tree      : a.txt b.txt seed.txt
plain cherry   : - 616bf026a585f183d8c49ec58de025fb68fc4d15 | - fb4444864adbdc8bed6c4f61bd35d172551c4054
                 (both lines '-', so S3 fires)
names(mb..feat): a.txt b.txt
state probe    : git diff --quiet feat main -- a.txt b.txt  ->  rc=1  (content differs; S5 = NO)
```

Both cherry lines are prefixed `-`, so the S3 guard ("non-empty AND every line `-`") is
satisfied and a history-only predicate returns merged. S5 is the signal that withholds the
verdict. This is why the target semantic in REQ-WSM-001 is stated in **state** terms and
why REQ-WSM-014 makes the conjunction binding.

### Synthetic-commit mechanism (Candidate A)

```
mb=$(git merge-base <base> <branch>)
synth=$(git commit-tree "$(git rev-parse <branch>^{tree})" -p "$mb" -m <fixed-message>)
git cherry <base> "$synth"     # leading '-' means the patch is already present in base
```

Collapsing the branch's entire diff-versus-merge-base into one synthetic commit gives it
the same shape as the single squashed commit on `base`, so their patch-ids match.
Verified on this repository against a known squash-merged branch: `git cherry` emitted a
single line prefixed `-`.

### State-conjunct mechanism (S5)

```
mb=$(git merge-base <base> <branch>)
# paths the branch touched, NUL-delimited and split on NUL — never on newline
names=( ); while IFS= read -r -d '' p; do names+=("$p"); done \
  < <(git diff --ignore-submodules=none --no-renames --name-only -z "$mb" <branch>)
# when names is empty the branch touched nothing: S5 holds trivially
# --literal-pathspecs is a GIT-LEVEL option and MUST precede the subcommand;
# --no-textconv and --ignore-submodules=none are DIFF-level and MUST follow it
git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none \
    <branch> <base> -- <names as separate arguments>
                                                 # rc=0 -> S5 holds; rc=1 -> S5 withholds
```

`--ignore-submodules=none` appears **twice on purpose**, once per stage. It is the flag a
future reader is most likely to delete as duplication, and deleting either occurrence
re-opens SC-15: the two submodule-ignore keys suppress the gitlink at the enumeration stage
*and* at the comparison stage, so closing one stage leaves the other open (finding 8;
mutation-checked in `acceptance.md` §C.1, where `no-ignoresub-cmp` and `no-ignoresub-enum`
each flip SC-15 on their own).

Option position is not stylistic, and the three flags do not share a position.
`--literal-pathspecs` is a **git-level** option and belongs before `diff`; `--no-textconv`
and `--ignore-submodules=none` are **diff-level** and belong after it. Every misplacement is
rejected outright rather than silently misread. Measured, all five forms executed:

```
git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none feat main -- …
                                                                       rc=1    <- correct
git diff --ignore-submodules=none --no-renames --name-only -z <mb> feat rc=0    <- enumeration, correct
git --no-textconv diff --quiet feat main -- report.pdf                  rc=129  unknown option
git --ignore-submodules=none diff --quiet feat main -- report.pdf       rc=129  unknown option
git diff --literal-pathspecs --quiet feat main -- ':note.txt'           rc=129  usage error
```

`rc=129` is distinct from both verdict codes (0 and 1), so a misplacement announces itself
instead of degrading into a wrong verdict. This is the one property the round-trip repairs
do **not** share: `--no-renames`, `-z`, and a joined path list all fail silently in the
merged direction, while every flag-position error here fails loudly.

The probe asks a state question: for every path the branch changed relative to its
merge-base, is base HEAD's content identical to the branch's? A `rc=1` means at least one
of those paths differs in base HEAD, so part of the branch's work is not present there —
regardless of what the patch-id history says.

Seven properties of this block are load-bearing, and each has a failure mode that is
silent rather than loud when the property is *omitted*:

- **`--no-renames` is required, not decorative.** `git diff --name-only` applies rename
  detection by default (`diff.renames` has defaulted to true since git 2.9; verified unset
  in a fresh repository, so the default applies). For a renamed path it reports only the
  **destination**; the deleted source path is folded away and never enters the pathspec.
  The branch's deletion of that path is then never checked. Measured, with `old.txt`
  present at the merge-base and renamed to `new.txt` on the branch:

  ```
  git diff --name-status <mb> feat        ->  R086  old.txt  new.txt
  git diff --name-only            <mb> feat  ->  [new.txt]            <- old.txt ABSENT
  git diff --no-renames --name-only <mb> feat ->  [new.txt old.txt]   <- old.txt PRESENT
  ```

  A future edit that removes the flag as redundant re-opens SC-10 and SC-11 as false
  positives. AC-WSM-006 exists to fail if it is removed.

- **`-z` is required, not decorative.** Without it `git diff --name-only` C-quotes any path
  containing a backslash, a tab, a double quote, or a control character, and — under git's
  documented default `core.quotePath=true` — any non-ASCII path. The quoted form is a
  different string from the path, so it matches nothing as a pathspec. Measured, by
  round-tripping each `--name-only` line straight back as a pathspec where the correct
  answer is `rc=1` in every row:

  ```
  core.quotePath=true (git's documented default)
    "a\\b.txt"                      rc=0   <- matched NOTHING
    "caf\303\251.txt"               rc=0   <- matched NOTHING
    plainname.txt                   rc=1   (matched)
    "tab\tname.txt"                 rc=0   <- matched NOTHING
    "\355\225\234\352\270\200.txt"  rc=0   <- matched NOTHING

  core.quotePath=false — which names are STILL quoted
    "a\\b.txt"        <- backslash quoted regardless of quotePath
    "tab\tname.txt"   <- control char quoted regardless of quotePath
    café.txt          (unquoted under quotePath=false)
    한글.txt           (unquoted under quotePath=false)

  -z (NUL-delimited) — never quoted; round-trip correct for all five
    a\b.txt   café.txt   plainname.txt   tab<TAB>name.txt   한글.txt      all rc=1
  ```

  Two consequences follow, and the second is the one a future reader is most likely to get
  wrong. First, the defect is **not** confined to non-ASCII repositories: a backslash or a
  control character in a path quotes on any machine, whatever the config. Second,
  **`core.quotePath=false` is not a fix** — it silences only the non-ASCII half, leaving the
  backslash and control-character halves intact, so a repair that merely documents the
  config setting closes two-thirds of the hole and reads as if it closed all of it. `-z` is
  what removes the round-trip hazard entirely. A future edit that drops `-z` as redundant
  re-opens SC-12 as a false positive; AC-WSM-006's second falsification exists to fail if
  it is removed.

- **`--literal-pathspecs` is required, not decorative.** `-z` makes the enumeration's output
  byte-exact; it does nothing about how those bytes are parsed when handed back. Git reads a
  pathspec's leading characters as magic before matching, so a repo-root path beginning with
  `:` names no file (finding 7), and `:!`/`:^` forms silently invert the comparison. The flag
  disables all pathspec magic and all globbing, which closes the interpretation axis
  categorically rather than symbol-by-symbol. Losing glob expansion costs nothing on the
  safety axis: over-matching can only add differences, and an added difference withholds a
  verdict. A future edit that drops the flag as redundant re-opens SC-13 as a false positive;
  AC-WSM-006's third falsification exists to fail if it is removed.

- **`--no-textconv` is required, not decorative.** The three prior flags all govern the path
  list; this one governs what the comparison *does* with a correct list. A `.gitattributes`
  rule binding a `textconv` diff driver makes `git diff --quiet` compare the driver's
  rendering rather than the blobs, so two genuinely different files compare equal
  (finding 8). The trigger is a documented, ordinary recipe — `*.pdf diff=pdf` with a
  `pdftotext` driver is an example in git's own `gitattributes(5)`, and the same pattern is
  routine for `*.docx` and image formats — and the predicate runs in whatever repository the
  user has. A future edit that drops the flag re-opens SC-14 as a false positive;
  AC-WSM-006's fourth falsification exists to fail if it is removed.

- **`--ignore-submodules=none` is required on BOTH stages, not decorative and not
  duplication.** `diff.ignoreSubmodules=all` and `submodule.<name>.ignore=all` suppress a
  changed gitlink at the enumeration stage *and* at the comparison stage, so a single-stage
  repair leaves the other stage open (finding 8). This is the only guard in the block that
  must be applied twice, and the enumeration-side occurrence is the one most likely to be
  deleted as redundant — it is not: without it the gitlink never enters `names`, and by the
  fail-open property below an absent path is silently "no difference". `submodule.<name>.ignore`
  is readable from a **tracked** `.gitmodules`, so unlike every other member of every axis
  this exposure can arrive inside the repository rather than from operator configuration. A
  future edit that drops either occurrence re-opens SC-15; AC-WSM-006's fifth falsification
  exists to fail if either is removed.

- **An under-inclusive pathspec fails open (finding 5).** `git diff --quiet` exits 0 when
  the pathspec matches nothing, so the omission caused by the fold does not raise an error
  — it reads as "no difference", which the predicate reads as merged. That is why the
  rename fold produced a false *positive* rather than a crash.

- **`names` must be passed as separate arguments, never as one reconstructed string.**
  This is the same failure surface reached by a different route. Passing the list as a
  single joined string yields one pathspec that matches nothing, and by the property above
  that reads as merged. Measured on the same repository, where the correct answer is
  `rc=1`:

  ```
  argv: <feat> <main> <--> <old.txt >     (list joined into one argument)   rc=0   <- WRONG
  argv: <feat> <main> <--> <old.txt>      (one argument per path)           rc=1   <- right
  ```

  The joined form arises from shell word-splitting differences (`bash` splits an unquoted
  `$names` and happens to survive; `zsh` does not split and fails), which is exactly why
  the implementation must not route the list through a shell string at all — see
  `plan.md` §D.

The empty-`names` guard matters: with no pathspec, `git diff --quiet <branch> <base>`
compares the whole trees and would report a difference for any branch behind an advanced
base. Scoping to the branch's own touched paths is what makes the probe answer "is the
branch's work present" rather than "are the trees equal".

### Falsified alternative 3 — enumerating S5's paths with rename detection on

Computing `names` with a plain `git diff --name-only` — the form this SPEC specified at
v0.2.0 — is falsified by SC-10 and SC-11. Both were run under both settings:

```
scenario                       names(rename-folded)  names(--no-renames)     folded      --no-renames
SC-10 rename + re-add, squash  [new.txt]             [new.txt old.txt]       merged      not merged
SC-11 rename + re-add, pick    [new.txt]             [new.txt old.txt]       merged      not merged
```

Under the folded enumeration, S5 reads `OK` in both rows — it never examines `old.txt`,
which the branch deleted and which base still holds — so the S4 signal (SC-10) and the
S3/S4 signals (SC-11) are granted unopposed and the composed verdict is **merged** while
the branch's deletion is genuinely absent from base. This is the same defect class as
SC-8: a patch-id history signal fires while the content does not survive in base HEAD.
SC-8 reached it through a revert; SC-10 and SC-11 reach it through a path the conjunct
could not see.

The repair is the `--no-renames` flag on the enumeration only. The matrix was re-run under
both settings: SC-1 through SC-9 are **cell-for-cell identical** in the two columns, so the
flag regresses nothing, and SC-10 and SC-11 flip from merged to not-merged. Verbatim output
is in `acceptance.md` §C.

### Falsified alternative 4 — splitting the enumeration's output on newline

Splitting `git diff --no-renames --name-only` output on newline — the form this SPEC
specified at v0.3.0 — is falsified by SC-12. Run under both splits, using the enumeration
exactly as v0.3.0 prescribed it:

```
scenario                      names(newline-split)                       names(-z)              nl-split    -z
SC-12 non-ASCII path removed  [plain.txt "\353\254\270\354\204\234.txt"] [plain.txt 문서.txt]    merged      not merged
```

Under the newline split the second path arrives C-quoted. That literal matches no file, so
`git diff --quiet` compares only `plain.txt` — which *is* identical in base — and exits 0.
S5 reads `OK`, both history signals are granted unopposed, and the composed verdict is
**merged** while `문서.txt` is genuinely absent from base HEAD. This is the same defect
class as SC-8 and as SC-10/SC-11, reached by a fourth route: not a revert, not a rename
fold, not a joined shell string, but the *encoding* of the enumeration's own output.

The repair is `-z` on the enumeration plus a NUL split. The full twelve-scenario matrix was
re-run under both splits: SC-1 through SC-11 are **cell-for-cell identical**, so the flag
regresses nothing, and SC-12 flips from merged to not-merged. Verbatim output is in
`acceptance.md` §C.

### Falsified alternative 5 — feeding the byte-perfect path list back as a bare pathspec

Passing the `-z` path list to `git diff --quiet` without literal-pathspec handling — the
form this SPEC specified at v0.4.0 — is falsified by SC-13. The enumeration is not at fault
here and the path list is not corrupt; the same bytes are read differently on the way back:

```
scenario                        names(-z)                    bare pathspec   --literal-pathspecs
SC-13 leading-colon path removed [:note.txt plain.txt]        merged          not merged
```

Under the bare form git parses `:note.txt` as a magic prefix rather than a filename, so it
names no file. `git diff --quiet` compares only `plain.txt` — identical in base — and exits
0. S5 reads `OK`, both history signals are granted unopposed, and the verdict is **merged**
while `:note.txt` is genuinely absent from base HEAD. This is the same defect class again,
reached by a fifth route: not a revert, not a rename fold, not a joined shell string, not
the enumeration's output encoding, but git's *parsing* of a path list that arrived intact.

The repair is `--literal-pathspecs` as a git-level option on the comparison. The full
thirteen-scenario matrix was re-run under both forms: SC-1 through SC-12 are
**cell-for-cell identical**, so the flag regresses nothing, and SC-13 flips from merged to
not-merged. Verbatim output is in `acceptance.md` §C.

### Falsified alternative 6 — comparing a correct path list under the repository's own diff configuration

Issuing the comparison without `--no-textconv` and without `--ignore-submodules=none` — the
form this SPEC specified at v0.5.0 — is falsified by SC-14 and SC-15. Neither the
enumeration nor the round-trip is at fault here; on SC-14 the path list is complete and
byte-perfect, and the comparison still reports equality:

```
scenario                             names                        v0.5.0 form   with the two flags
SC-14 textconv driver                [plain.txt report.pdf]       merged        not merged
SC-15 submodule ignore=all           [plain.txt] / [plain.txt sub] merged        not merged
```

On SC-14 the `.gitattributes` rule `*.pdf diff=pdf` binds a `textconv` driver; `git diff
--quiet` compares the driver's rendering of each blob rather than the blobs themselves, both
render to the same text, and the comparison exits 0. Ground truth is that the blobs differ:

```
feat:report.pdf = dcc7c925c1f3fb918a192d6f40835189490d4e4c
main:report.pdf = 43d844c81ab8ccc32d09beb9d70f5831d1474402
git --literal-pathspecs diff --quiet feat main -- report.pdf                 rc=0   <- WRONG
git --literal-pathspecs diff --quiet --no-textconv feat main -- report.pdf   rc=1   <- right
```

On SC-15 the failure is one stage earlier as well: under `submodule.<name>.ignore=all` the
moved gitlink never enters the path list at all, so the comparison is never asked about it.
Ground truth is that the pointers differ:

```
feat:sub = 1a7d9a656cfc3f54f58da6d184dcbc0c33a29cb3
main:sub = 034452180de5fd3d9b2bba76a8bc15273cad6386
enumeration without --ignore-submodules=none  ->  [plain.txt]          <- gitlink DROPPED
enumeration with    --ignore-submodules=none  ->  [plain.txt, sub]     <- gitlink present
comparison without the flag, given the correct list                 rc=0   <- WRONG
comparison with    the flag, given the correct list                 rc=1   <- right
```

S5 reads `OK` in both, both history signals are granted unopposed, and the composed verdict
is **merged** while the branch's content is genuinely absent from base HEAD. This is the same
defect class again, reached by a sixth route: not a revert, not a rename fold, not a joined
shell string, not the enumeration's output encoding, not git's parsing of the path list —
but git's *judgement* of a path list that was correct on every prior axis.

The repair is `--no-textconv` and `--ignore-submodules=none` on the comparison, plus
`--ignore-submodules=none` on the enumeration. The full seventeen-scenario matrix was re-run
under both forms: SC-1 through SC-13 are **cell-for-cell identical**, so the flags regress
nothing, and SC-14 and SC-15 flip from merged to not-merged. Verbatim output is in
`acceptance.md` §C.

### Six honest limits on these repairs, taken together

The repairs are better read as an axis map than as a patch count, because the axis map is
what says where the mechanism can still be attacked. Four axes concern the **path list** —
whether it names what it should. The fifth concerns the **comparison** — whether a correct
list is judged honestly. The sixth concerns the **invocation** — whether the two stages
resolve paths against the same root at all:

| Axis | Failure | Closed by | Falsified by |
|---|---|---|---|
| completeness | the rename fold hides the branch's deleted source path | `--no-renames` | SC-10, SC-11 |
| argv | the path list is joined into one string | `[]string`, one argument per path | measured in §C.2 route 2 |
| encoding | git C-quotes the path, so the bytes are not the path's | `-z` + NUL split | SC-12 |
| interpretation | the bytes are correct and git parses them as magic | `--literal-pathspecs` | SC-13 |
| **comparison semantics** | the list is correct and the comparison is configured not to notice | **`--no-textconv`** (comparison) + **`--ignore-submodules=none`** (**both** stages) | SC-14, SC-15 |
| **frame of reference** | the list is correct and the comparison resolves it against a different root — the enumeration emits repo-root-relative paths, the pathspecs resolve against the process working directory, so a nested cwd turns every path into an unmatched pathspec and finding 5 reads that as merged | **invoking git with cwd = the repository root** (`execGit`'s `dir` argument is `w.root`, i.e. `rev-parse --show-toplevel`). The fix must be at the invocation level: `--literal-pathspecs` forecloses the in-band `:(top)` remedy by reading the magic prefix as a literal filename | measured — default configuration, cwd below the root: `rc=1 / S5=NO` at the root, `rc=0 / S5=OK` from a subdirectory. Not a live defect under the specified design, and pinned rather than patched (`acceptance.md` §F) |

Three things follow, and the third is the one a reader is most likely to get wrong:

- **`--literal-pathspecs` closes an axis, not a symbol.** It disables *all* pathspec magic
  and *all* globbing by definition, rather than enumerating hazardous characters. That
  distinction is load-bearing, because a symbol-by-symbol repair would leave the axis open
  to whichever form went unenumerated.
- **The interpretation axis had exactly one open symbol class.** Glob metacharacters and
  backslashes were *measured* to round-trip correctly under a bare pathspec — git's matcher
  attempts a literal comparison in addition to wildmatch, so a file literally named
  `*star.txt` is matched by the pathspec `*star.txt`. Measured, where the correct answer is
  `rc=1` in every row: `*star.txt`, `?q.txt`, `[a]x.txt`, and `a\b.txt` each return `rc=1`
  under both the bare and the literal forms. A leading `:` was therefore the only class the
  bare form left open, and the flag closes it along with every other magic form.
- **The comparison-semantics axis is not closed by a single flag, because its members do not
  share a stage.** textconv is comparison-side only; the two submodule-ignore keys suppress
  at both stages. `--no-textconv` on the comparison plus `--ignore-submodules=none` on the
  comparison *and* the enumeration is what covers all three. A repair that patched only the
  comparison would close one member of three and read as though it closed the axis.

**What is known about completeness, stated as a boundary rather than as a closure.** The
previous version of this passage asserted that no open axis remained and that no sixth patch
was queued. That sentence was **false when written** and is deleted rather than softened: a
sixth instance was already constructible, and an affirmative claim in the unsafe direction
is the specific error that failed this SPEC at iteration 3 and again at iteration 5. What
can honestly be said is bounded by what was probed:

- Twenty-two configuration keys, attribute forms, and environment settings were measured
  against an ordinary-file sharp scenario (`acceptance.md` §C.2 route 5). That sweep finds
  **one** member: the `textconv` diff driver (`diff.<d>.cachetextconv` is the same member,
  not a distinct one). Two further members — `diff.ignoreSubmodules=all` and
  `submodule.<name>.ignore=all` — are invisible to a file-only sweep because they act
  exclusively on gitlinks, and were found instead by a fixture containing a submodule
  (SC-15). **Three members total, all now closed.** The methodological lesson is recorded
  because it is why the sweep initially undercounted: an axis sweep must vary the *object
  kind* (file, symlink, gitlink) as well as the configuration.
- **External diff drivers were measured NOT to be a member.** `diff.<d>.command` and
  `GIT_EXTERNAL_DIFF` do run on the ordinary diff path, but `--quiet` short-circuits before
  consulting them, so the comparison is unaffected. This is the same asymmetry as textconv
  in the opposite direction, and it is recorded because it looked like a likely fourth
  member and is not one.
- **Whether a fourth member exists is unknown.** Territory not reached: git attribute macros
  beyond `binary`, `diff.<d>.binary`, partial-clone / promisor object behaviour,
  `core.hooksPath`-driven state, alternate object stores, and `GIT_ATTR_NOSYSTEM`
  interactions. No claim is made about them either way.

The honest framing is **"no *known* open axis"**, and every closure claim in this SPEC should
be read that way. Each of the seven instances so far was found by attacking the mechanism the
previous repair produced — the seventh, the frame-of-reference axis, at audit iteration 6 —
and the SPEC's confidence has now been wrong twice in exactly that manner. A future revision
that adds a seventh axis is the expected outcome of another adversarial pass, not evidence
that this one was written carelessly.

**The reformulation was built, attacked, and rejected — on measured grounds.** Earlier
versions of this passage named a structural alternative (comparing the branch's tree against
base's tree restricted to the branch-touched subtree, computing the intersection in-process
so no path is ever handed back to git) and directed that a fifth patch be weighed against it
rather than adopted reflexively. That weighing was performed as a commissioned study rather
than left as an open question, and the outcome is recorded here so the option is not
silently re-proposed:

- Both candidates were correct on everything constructible — 12/12 scenarios, 13/13 path
  hazards, 14/14 structural probes, and 1500/1500 differential-fuzz cases against an
  independent oracle implemented from the definition rather than from either mechanism.
  Invocation counts are identical (9 per branch evaluation, 8 when the touched set is
  empty). The choice was therefore never about measured correctness; it was about which
  residual surface is smaller.
- **The reformulation relocates the defect class rather than eliminating it.** The pathspec
  form needs `--no-renames` on *one* enumeration; the reformulation needs it on *two*, and
  dropping it from the second is a silent fail-open false positive of exactly the same
  character as the five before it. The study constructed the case (`A1b`: base squash-merges
  the branch, then renames the branch's file away): with rename detection left on for the
  divergence enumeration, the `f.txt → g.txt` pair folds to its destination, the
  intersection is spuriously empty, and the verdict is merged while the branch's file is
  absent from base HEAD. None of the twelve existing scenarios covers it. The pathspec form
  is immune to `A1b` without any flag, because restricting the diff to `f.txt` structurally
  prevents git from pairing it with `g.txt`.
- **Guard count favours the reformulation; guard *maturity* favours the patch.** The
  reformulation carries 2 silently-fatal guards against the patch's 4 — a genuine advantage,
  and the strongest argument for it. But all four of the patch's guards are specified,
  verified across the whole matrix, and pinned by acceptance criteria, whereas one of the
  reformulation's two is new, unpinned, and demonstrably breakable. Trading four mature
  guards for two of which one is immature is not obviously a gain, and it is not a gain at
  this point in the SPEC's lifecycle.

**The OID-comparison hybrid was weighed and rejected — also on measured grounds.** After the
comparison-semantics axis was found, a second structural alternative was commissioned and
studied: keep the enumeration exactly as it is and replace only the *comparison* with a
blob-OID comparison, on the reasoning that object IDs are content hashes and no diff
configuration can make two different contents hash alike. That reasoning is sound about the
comparison and irrelevant to the enumeration, which is the half the hybrid keeps by design.
The outcome is recorded here so the option is not proposed a third time:

- **It closes one member of the axis and inherits the other two.** textconv is
  comparison-side only, so the hybrid closes it completely and structurally. But
  `diff.ignoreSubmodules=all` and `submodule.<name>.ignore=all` suppress at the
  **enumeration** stage, which the hybrid retains verbatim — measured, the hybrid returns
  the same `merged` false positive on both, identically to the unpatched form. Adopting it
  would leave the axis open while requiring a mechanism rewrite. The two flags close all
  three members instead.
- **Its own defects are unconditional where the axis's are conditional.** A symlink and a
  regular file with identical content **share one blob OID** (P4c), and a mode change leaves
  the OID untouched (P3c), so an OID comparison that does not also carry the file mode
  reports equality for genuinely different trees — silently, fail-open, under **default** git
  configuration. Every comparison-semantics member requires non-default state: a committed
  `.gitattributes`, a committed `.gitmodules` directive, or an operator config key. Trading
  config-conditional exposure for unconditional exposure is a bad trade. Two of the three
  mechanisms considered (`git rev-parse <rev>:<path>` and `git cat-file --batch-check`)
  cannot report the mode at all, and `cat-file` additionally reports `missing` for a gitlink
  on both sides at exit 0 — reading an error as agreement.
- **Cost regresses in the case that actually runs.** The only viable mechanism (two full
  `git ls-tree` invocations) is flat in the touched-path count and measured at ~169 ms on
  this repository's 6,698-entry tree, against 43-87 ms for the current design at typical
  branch sizes, crossing over only past several hundred touched paths. `--stale` pays this
  once per registered worktree.
- **Guard maturity, again.** The hybrid trades four mature guards for three, one of which
  (include the mode) is new, unpinned, and demonstrably breakable. This is the same trade
  declined for the reformulation one iteration earlier, and the case for declining is
  stronger here because the reduction is smaller.

**What the hybrid genuinely offered, recorded so it is not lost.** It passes **no path to
git** — the two `ls-tree` invocations take only a revision, and the enumerated names are used
solely as keys into an in-process map — so the argv, encoding, and interpretation axes are
structurally *absent* from its comparison rather than closed by flags. That is a real
advantage and the strongest argument in its favour. It also admits a fail-loud upgrade the
pathspec form cannot express: assert that every enumerated name resolves in at least one of
the two trees, and error otherwise, which would convert the fail-open-on-unmatched-name
property into a loud failure. Neither outweighs the four points above. But if a future
revision ever finds a **comparison-side** axis-5 member that `--no-textconv` does not cover,
the hybrid's `ls-tree` form is the mechanism to reach for — with the mode included in the
comparison, which P3c and P4c exist to enforce.

The residual honest limit is unchanged in kind: S5 still asks git for a list of names, then
asks git a second question phrased in terms of those names, and now also depends on git's
answer to that second question being configured honestly. The round-trip and the comparison
are both structural hazards. What has changed is that **both** alternatives to the current
mechanism are no longer untested ideas — each was built, attacked, and measured to carry a
defect of the same class in a position this matrix would not have covered.
- The not-merged verdict on SC-10 through SC-15, P3c, and P4c is conservative by design. One
  could argue base "re-added the file deliberately", "removed the file deliberately", "moved
  the submodule pointer deliberately", or "reset the file mode deliberately", and that the
  branch's content work did land. The state semantic in REQ-WSM-001 does not make that
  distinction, and the direction it errs in is the safe one: the worktree is kept.

### Known limitations (accepted)

**Amended replay — a conservative false negative.** A branch merged by replaying its commits
individually **and** subsequently amended so that neither the per-commit patch-ids nor the
collapsed patch-id match will not be detected. The worktree is kept, which is the safe
outcome.

**`diff.ignoreSubmodules=all` — CLOSED at v0.6.0; the hedge below is retired.** Through
v0.5.0 this was recorded as an exposure that was *stated, not verified*: the reformulation
study's submodule probe returned the correct verdict under `diff.ignoreSubmodules=all`, but
only because `.gitmodules` also differed and carried the comparison, so the case of a pointer
bumped with **no** accompanying `.gitmodules` change was inferred from the enumeration's
mechanics rather than observed.

That scenario has since been constructed (SC-15) and the inference confirmed: the gitlink is
dropped from `git diff --name-only` itself, and the comparison would ignore it independently
even when handed a correct list. Two corrections to the v0.5.0 wording follow. First, the
exposure hits **both** stages, not only the enumeration the earlier text named. Second, a
**third** member exists that the earlier text did not know about — `submodule.<name>.ignore=all`,
which is readable from a tracked `.gitmodules` and therefore does **not** require a
non-default local config; a repository can ship it to every clone. The v0.5.0 claim that "the
exposure does not exist under git's shipped settings" was true of `diff.ignoreSubmodules` and
false of its sibling.

The exposure is now closed by `--ignore-submodules=none` on the enumeration and on the
comparison (finding 8), pinned by SC-15 and by AC-WSM-006's fifth falsification. What remains
true from the earlier wording is that the exposure was never specific to this design: it is a
property of `git diff` and applies identically to any formulation built on that enumeration,
including both structural alternatives weighed above — the OID hybrid inherits it in full.

---

## §3 Requirements (GEARS)

### Detection semantics

**REQ-WSM-001** (Ubiquitous) — The worktree merge predicate shall report a branch as
merged when the branch's changes relative to its merge-base with the base branch are
present **in the base branch's HEAD tree**, irrespective of the merge strategy that placed
them there. This is a state condition, not a history condition: an equivalent patch-id
having existed at some point in base's history since the merge-base does not satisfy it if
the content was subsequently removed (see §2 SC-8 and REQ-WSM-014).

**REQ-WSM-002** (event-driven) — **When** a branch's entire diff relative to its merge-base
has been applied to the base branch as a single squashed commit, the worktree merge
predicate shall report the branch as merged.

**REQ-WSM-003** (event-driven) — **When** every commit on a branch since its merge-base has
an equivalent patch present in the base branch, the worktree merge predicate shall report
the branch as merged.

**REQ-WSM-004** (event-driven) — **When** a branch's diff relative to its merge-base with
the base branch is empty, the worktree merge predicate shall report the branch as merged.

**REQ-WSM-005** (state-driven) — **While** a branch's tip commit is an ancestor of the base
branch, the worktree merge predicate shall report the branch as merged, preserving the
existing reachability behaviour for true merge commits and strictly-behind branches.

**REQ-WSM-006** (unwanted) — The worktree merge predicate shall not report a branch as
merged while any part of the branch's diff relative to its merge-base is absent from the
base branch's HEAD tree.

> The state conjunct that makes REQ-WSM-002 and REQ-WSM-003 safe against a later revert is
> stated as REQ-WSM-014 at the end of this section, where the sequential numbering puts it.

**REQ-WSM-007** (event-driven) — **When** a git probe the predicate issues exits with a
status that probe does not define as a verdict, the predicate shall return an error rather
than a boolean verdict, so that `staleKeepReason` keeps the worktree and reports the
failure. The verdict-carrying exit statuses are enumerated below and are the complete set;
any status outside a probe's listed verdict column is a failure.

Measured on `git version 2.50.1 (Apple Git-155)`:

| Probe | Verdict-carrying exits | Meaning | Failure exits |
|---|---|---|---|
| `git branch --merged <base>` | `0` | listing produced (branch present or absent in it) | `128` (e.g. unknown base ref) |
| `git merge-base <base> <branch>` | `0` merge-base found; `1` no common ancestor, empty stdout | see REQ-WSM-007 note below | `128` (e.g. unknown ref) |
| `git diff --quiet <a> <b>` | `0` no difference; `1` difference exists | both are verdicts, not errors | `128` (e.g. unknown ref) |
| `git diff --quiet <a> <b> -- <paths>` | `0` no difference; `1` difference exists | both are verdicts, not errors | `128` (e.g. unknown ref) |
| `git diff --quiet <a> <b> -- <pathspec matching nothing>` | `0` | **no error is raised** — an unmatched pathspec is reported as "no difference" | (none; an unmatched pathspec never fails) |
| `git diff --quiet <a> <b> -- <path whose first byte is ':'>` | `0` | the leading `:` is parsed as pathspec **magic**, not as a filename, so the pathspec matches nothing and reports "no difference"; `:!`/`:^` forms parse as EXCLUDE and invert what is compared | (none; the misparse never fails) |
| `git --literal-pathspecs diff --quiet <a> <b> -- <paths>` | `0` no difference; `1` difference exists | all pathspec magic and globbing disabled — every element is matched as a literal filename | `128` (e.g. unknown ref). Note the option is **git-level**: placed after the subcommand git rejects the invocation with `129` |
| `git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none <a> <b> -- <paths>` | `0` no difference; `1` difference exists | the specified comparison — a `textconv` driver cannot collapse a real difference, and a changed gitlink cannot be ignored | `128` (e.g. unknown ref); `129` if either diff-level flag is placed **before** the subcommand (measured: `unknown option`) |
| `git diff --quiet <a> <b> -- <path under a textconv diff driver>` | `0` | the driver's rendering is compared instead of the blobs, so two genuinely different files report "no difference" | (none; the collapse never fails) |
| `git diff --quiet <a> <b> -- <changed gitlink>` under `diff.ignoreSubmodules=all` or `submodule.<n>.ignore=all` | `0` | the gitlink is ignored, so a moved submodule pointer reports "no difference" | (none; the suppression never fails) |
| `git diff --ignore-submodules=none --no-renames --name-only -z <a> <b>` | `0` | listing produced (possibly empty), rename sources included, paths verbatim and NUL-delimited, **and a changed gitlink included** even under a repository- or operator-set submodule-ignore directive | `128` (e.g. unknown ref) |
| `git diff --no-renames --name-only <a> <b>` | `0` | listing produced (possibly empty), rename sources included — but paths are **C-quoted** when they contain a backslash, tab, double quote, or control character, and (under default `core.quotePath=true`) when they contain any non-ASCII byte | `128` (e.g. unknown ref) |
| `git diff --no-renames --name-only -z <a> <b>` | `0` | listing produced (possibly empty), rename sources included, paths emitted **verbatim** and NUL-delimited — never quoted, for any byte, under any `core.quotePath` value | `128` (e.g. unknown ref) |
| `git cherry <base> <head>` | `0` | listing produced (possibly empty) | `128` (e.g. unknown ref) |
| `git commit-tree <tree> -p <parent> -m <msg>` | `0` | object written | `128` (e.g. invalid tree) |

Note on `git merge-base` exit `1`: it is a verdict meaning the two refs have no common
ancestor (unrelated histories), and stdout is empty. It is not a failure, but it also does
not yield a merge-base, so signals S2, S3 (state conjunct), S4, and S5 cannot be computed.
The predicate shall treat this case as **not merged** with a nil error — the conservative
outcome that keeps the worktree — rather than returning an error or a merged verdict.

Read literally, an earlier form of this requirement ("any git probe exits non-zero →
return an error") made S3, S4, and S5 unreachable, because `git diff --quiet` uses exit `1`
as a verdict and every branch not caught by S2 produces it. The table above exists so the
implementer does not have to guess which non-zero statuses are verdicts.

Note on the unmatched-pathspec row: it is in the table because it is the one entry that is
**not** a safety net. Every other row lets a mistake surface as an error; this one converts
a mistake — an incomplete or malformed path list — into a silent merged verdict (§2
finding 5). It is recorded so that a future reader treats an empty S5 result as suspicious
rather than reassuring.

Note on the three enumeration rows: they are all listed, rather than only the one the design
uses, so that a future reader considering "`-z` is noise, drop it" can see from the table
alone what breaks. Without `-z` the enumeration's output is a *quoted rendering* of the
paths rather than the paths themselves, and by the unmatched-pathspec row above every
quoted name silently degrades into "no difference". Note also what the non-`-z` row does
**not** say: it does not say the hazard is scoped to non-ASCII paths, and it does not say
`core.quotePath=false` removes it. Backslash, tab, double quote, and control characters
quote regardless of that setting (§2 finding 6), so config alone cannot substitute for
`-z`. The third row — the one carrying `--ignore-submodules=none` — is the form the design
actually uses, and it is listed separately for the same reason: a reader considering "the
comparison already has that flag, so the enumeration's copy is redundant" can see from the
table alone that the two rows differ in whether a changed gitlink is listed at all.

### Synthetic-object hygiene

**REQ-WSM-008** (event-driven) — **When** the predicate creates a synthetic commit object, it
shall pin the author and committer name, email, and date to fixed values, so that repeated
evaluation of an unchanged branch yields a byte-identical object rather than a new one.

**REQ-WSM-009** (unwanted) — The predicate shall not invoke `git prune`, `git gc`, or any
other repository-wide object-reclamation command.

### Preserved safety contract

**REQ-WSM-010** (state-driven) — **While** the `--stale` sweep evaluates a worktree, both
existing safety conditions shall be required together: the working tree is clean of
uncommitted and untracked files, AND the branch holds no unique work beyond base. Neither
condition may be dropped or weakened.

**REQ-WSM-011** (unwanted) — The `--stale` sweep shall not delete any branch; it removes
only the worktree directory.

**REQ-WSM-012** (state-driven) — **While** the `--yes` flag is absent, the `--stale` sweep
shall preview the removals and perform none of them.

**REQ-WSM-013** (event-driven) — **When** the `--merged-only` sweep removes a worktree, it
shall continue to call the removal with force disabled, so that git refuses removal of a
worktree containing modified or untracked files.

### State conjunct

**REQ-WSM-014** (event-driven) — **When** the predicate evaluates a patch-id history signal
(plain `git cherry` or synthetic-commit `git cherry`), it shall additionally require that
every path the branch changed relative to its merge-base has identical content in the base
branch's HEAD tree, and shall report merged only when both the history signal and this
state condition hold. The state condition shall never be treated as a merged verdict on its
own.

The set of "every path the branch changed" shall be **complete**, **faithfully encoded**, and
**matched literally**, and the comparison applied to it shall be **observed under unnarrowed
diff semantics**. Each part is stated at the requirement layer for the same reason: the
conjunct fails open, so a path set that is incomplete, corrupted, *or* re-interpreted — and
equally a correct path set judged under a narrowed comparison — produces a merged verdict
with no error (§2 findings 5 and 8). None of the four can be left implicit in the mechanism.

- **Completeness.** The set shall include paths the branch **deleted**, paths the branch
  **renamed away from**, and **gitlinks whose commit pointer the branch moved** — not only
  the paths that exist as ordinary files on the branch after the change. A path the branch
  removed is part of the branch's diff, so base HEAD still holding it means the branch's work
  is not fully present; the same holds for a submodule pointer. The mechanisms satisfying
  this are `--no-renames` and `--ignore-submodules=none` on the enumeration; §2 SC-10 and
  SC-11 falsify any weaker rename handling, and §2 SC-15 falsifies an enumeration that omits
  the submodule flag. A repository can suppress the gitlink through a tracked `.gitmodules`
  directive, so this part is not satisfied by assuming default configuration.
- **Encoding fidelity.** Each element of the set shall be the path itself, byte for byte,
  such that passing it back as a pathspec matches the file it names. It shall **not** be a
  quoted or escaped rendering of the path. `git diff --name-only` emits a C-quoted
  rendering for any path containing a backslash, tab, double quote, or control character —
  and, under git's documented default `core.quotePath=true`, for any non-ASCII path — and
  such a rendering matches nothing (§2 finding 6). The mechanism satisfying this is `-z`
  (NUL-delimited output, split on NUL); §2 SC-12 is the scenario that falsifies a
  newline-split enumeration. Setting `core.quotePath=false` does **not** satisfy this
  requirement: it suppresses quoting for non-ASCII paths only, leaving backslash and
  control-character paths quoted, so it is not an accepted substitute for `-z`.
- **Literal matching.** Each element shall be matched as a literal filename when it is
  passed back, and shall **not** be parsed as a pathspec expression. Encoding fidelity alone
  does not give this: a path whose bytes are transmitted perfectly is still re-read on the
  way back, and git parses a pathspec's leading characters as magic before matching, so a
  repo-root path whose first byte is `:` names no file and `:!`/`:^` forms invert the
  comparison (§2 finding 7). The mechanism satisfying this is the git-level option
  `--literal-pathspecs` on the comparison — or, equivalently, `GIT_LITERAL_PATHSPECS=1` in
  the invocation's environment; §2 SC-13 is the scenario that falsifies a bare pathspec.
  Prefixing each element with `:(literal)` also satisfies the requirement mechanically but
  is **not** the accepted mechanism: it re-introduces a per-path string transformation,
  which is the precise operation that has failed on three of the four axes above.
- **Observed comparison.** The comparison applied to the path set shall report a difference
  whenever the branch's and base's content for a named path genuinely differ, and shall not
  be narrowed by repository or operator configuration into reporting equality. Correctness of
  the path list does not give this: a complete, byte-perfect, literally-matched list is still
  judged by `git diff --quiet` under the repository's diff configuration, and that
  configuration can collapse a real difference — a `textconv` diff driver renders two
  different blobs to the same text, and a submodule-ignore directive drops a moved gitlink
  (§2 finding 8). The mechanisms satisfying this are `--no-textconv` and
  `--ignore-submodules=none` on the comparison; §2 SC-14 and SC-15 are the scenarios that
  falsify a comparison issued under the repository's own settings. The comparison shall also
  remain sensitive to a path's **mode**: a symlink and a regular file with identical content
  share one blob OID, so a content-hash-only comparison reports equality for genuinely
  different trees under default configuration (§2 P3c, P4c). `git diff` satisfies this; a
  mode-blind substitute does not.

Taken together the four parts give the full mechanism: enumerate with
`git diff --ignore-submodules=none --no-renames --name-only -z` split on NUL, and compare
with `git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none`.

---

## §4 Exclusions

### Out of Scope — auxiliary merge signals

- Querying the forge for merged pull requests (`gh pr list --head <branch> --state merged`).
  It was verified accurate on the two probe branches, but it requires network access, the
  `gh` binary, and GitHub hosting, so it is unsuitable as the default path of a local CLI.
  It is not added as an optional flag either; that would be a separate feature.

### Out of Scope — `--merged-only` safety expansion

- Adding an explicit working-tree cleanliness check to the `--merged-only` path. That path
  currently relies on git refusing a non-forced removal of a dirty worktree, which was
  verified sufficient (see §5 decision 3). Changing its gate is a behaviour change beyond
  this defect.

### Out of Scope — unpushed-commit detection

- Detecting that a worktree's branch carries commits absent from its upstream remote. A
  clean worktree whose branch holds unpushed commits is removed by `git worktree remove`
  without `--force` and is reported clean by `git status --porcelain`, so neither existing
  safety condition sees it (measured in §5 decision 3). Adding an upstream-comparison gate
  is a behaviour change to both sweep paths and belongs to a separate SPEC. The exposure is
  bounded by REQ-WSM-011: the branch is never deleted, so the commits stay reachable by
  branch name after the directory is removed.

### Out of Scope — worktree lifecycle changes

- Automatic branch deletion after worktree removal.
- Changing the default of `--stale` from preview to apply.
- Any change to worktree creation, entry, or the `moai worktree done` disposal contract.
- Reclaiming the currently-held disk; this SPEC changes detection only. Running the sweep
  is a user action taken after the fix lands.

### Out of Scope — object-store maintenance

- Scheduling, configuring, or invoking garbage collection. Git's own auto-gc governs
  reclamation of the synthetic objects.

---

## §5 Resolved Design Decisions

### Decision 1 — False-positive boundary

The predicate treats "the branch's changes relative to its merge-base are present in base
HEAD" as merged — a **state** semantic, not a history semantic. The boundary cases were
resolved against executed results, not reasoning:

| Case | Verdict | Why |
|---|---|---|
| Empty-diff branch (zero changes vs merge-base) | **Reclaim** | There is nothing to lose by definition. The branch itself is never deleted, so any empty commits remain reachable by branch name. |
| Strictly behind base | **Reclaim** | Both the reachability signal and the empty-diff signal fire; this already worked before the change and must keep working. |
| Base is a strict superset of the branch | **Reclaim** | Observed SC-9: S3 and S4 both fire and S5 holds. The branch's work is present; base merely carries more. |
| Partially applied (some hunks landed, some did not) | **Keep** | No history signal fires and S5 withholds. Observed: plain `git cherry` returned one `-` and one `+`; the synthetic probe returned `+`; S5 = NO. |
| **Partially reverted** (all patches applied, then some reverted) | **Keep** | The case that forced the state semantic. Observed SC-8: both plain-`git cherry` lines are `-`, so S3 fires on history alone, but S5 = NO because `b.txt` is absent from base HEAD. The conjunction withholds the verdict. |
| **Rename + re-add** (branch renames a path; base takes the work, then re-adds the original path) | **Keep** | The case that forced the enumeration to be rename-blind. Observed SC-10 and SC-11: the branch's deletion of `old.txt` does not survive in base HEAD, so S5 = NO — but only once the path list is built with `--no-renames`. Conservative: base may have re-added the path deliberately, and the predicate does not attempt to tell the two apart. |
| **Non-ASCII (or otherwise quoting) path removed from base** | **Keep** | The case that forced the enumeration to be NUL-delimited. Observed SC-12: the branch's `문서.txt` does not survive in base HEAD, so S5 = NO — but only once the path list is read with `-z`. Split on newline the path arrives C-quoted, matches nothing, and S5 holds vacuously. |
| **Leading-colon path removed from base** | **Keep** | The case that forced the comparison to use literal pathspecs. Observed SC-13: the branch's `:note.txt` does not survive in base HEAD, so S5 = NO — but only once the comparison carries `--literal-pathspecs`. As a bare pathspec the byte-perfect path is parsed as magic, matches nothing, and S5 holds vacuously. |
| **Changed file under a `textconv` diff driver** | **Keep** | The case that forced the comparison to fix its own diff semantics. Observed SC-14: the branch's `report.pdf` content does not survive in base HEAD, so S5 = NO — but only once the comparison carries `--no-textconv`. Without it the driver renders both blobs to the same text and the comparison reports equality on a correct, byte-perfect path list. |
| **Submodule pointer moved under a submodule-ignore directive** | **Keep** | The case that forced both stages to carry `--ignore-submodules=none`. Observed SC-15: the branch's pointer move does not survive in base HEAD, so S5 = NO — but only once **both** the enumeration and the comparison carry the flag. Without it on the enumeration the gitlink never enters the path list; without it on the comparison a correct list is still ignored. |
| **Mode-only divergence, and symlink/file blob-OID collision** | **Keep** | Not a repair — a *retention* row. Observed P3c and P4c: `git diff` compares modes, so the specified mechanism already reports both correctly. They are pinned because both sides share one blob OID, so any future swap to a mode-blind comparison turns them into silent false positives under default configuration. |

Eight guards carry the safety argument, and each was checked by mutation rather than by
reasoning (observed verdicts under each mutation are recorded in `acceptance.md` §C.1):

- **Every `git cherry` line must be `-`.** A single `+` line means part of the branch's work
  never landed.
- **Empty `git cherry` output is not a merged verdict.** An absence of output means "no
  commits compared" and can never be read as merged.
- **S5 must hold alongside S3 or S4.** This is the guard SC-8 requires; the other two do not
  catch it, because a revert leaves the patch-id history intact.
- **S5's path list must be built rename-blind.** This is the guard SC-10 and SC-11 require.
  It is a property of the conjunct's *input*, not of the conjunct, which is why the
  previous three guards did not cover it: S5 was present and correct, and still returned
  `OK`, because the path it needed to examine was never handed to it.
- **S5's path list must be read NUL-delimited.** This is the guard SC-12 requires. It is
  likewise a property of the input, and it is orthogonal to the previous one — the fourth
  guard governs *which* paths are enumerated, this one governs *what form they arrive in*.
  A rename-blind enumeration split on newline still hands S5 a quoted string that matches
  nothing, so the fourth guard alone does not cover it.
- **S5's comparison must match its paths literally.** This is the guard SC-13 requires. It is
  the last of the guards that governs the path list's *round-trip*: a rename-blind,
  NUL-delimited, byte-perfect list is still re-parsed when handed to `git diff`, and a
  leading `:` is parsed as magic. The previous two guards cannot cover it because the list
  they produce is already correct.
- **S5's comparison must not be narrowed by diff configuration.** This is the guard SC-14
  requires, and it is the first that governs neither which paths are enumerated, nor what
  form they arrive in, nor how they are read back, but what git *concludes* about them. A
  `textconv` diff driver makes `git diff --quiet` compare renderings rather than blobs, so a
  correct list is judged wrongly. None of the previous guards can cover it, because there is
  nothing wrong with the list.
- **S5 must see a changed gitlink, at both stages.** This is the guard SC-15 requires, and it
  is the only guard that must be applied twice: a submodule-ignore directive — settable by
  the operator *or* shipped in a tracked `.gitmodules` — suppresses the gitlink both when the
  path list is built and when the comparison is made. Closing one stage leaves the other
  open.
- **S5's comparison must be sensitive to file mode.** This is the guard P3c and P4c require.
  It is the only guard that is already satisfied by the specified mechanism and is pinned
  against *future* edits rather than against a present defect: `git diff` compares modes,
  while a content-hash comparison does not, and a symlink and a regular file with identical
  content share one blob OID.

The eight guards are independent rather than redundant, and the independence was measured
rather than argued. Removing S5 alone flips SC-8 to merged, while SC-6 and SC-7 additionally
require the cherry guard to be relaxed before they flip — so no single mutation defeats all
the must-keep scenarios. The five enumeration-and-comparison guards are likewise minimal and
specific, and their blast radii are disjoint except where the mechanism itself is shared:
dropping `--no-renames` flips exactly SC-10 and SC-11; splitting on newline instead of NUL
flips exactly SC-12; dropping `--literal-pathspecs` flips exactly SC-13; dropping
`--no-textconv` flips exactly SC-14; dropping `--ignore-submodules=none` from **either**
stage flips exactly SC-15; and swapping the comparison for a mode-blind one flips exactly
P3c and P4c. Each leaves every other row unchanged (observed across the full 11 × 17 grid;
`acceptance.md` §C and §C.1).

The one deliberate non-disjointness is worth stating rather than smoothing over: the two
submodule mutations flip the *same* row, because they are two stages of one guard rather
than two guards. That is the evidence the flag is needed twice — if the stages were
redundant, removing one would leave SC-15 passing.

### Decision 2 — Synthetic commit objects

**Accept the dangling objects; pin the identity to bound them.** Each `git commit-tree`
call creates one unreferenced, gc-eligible commit object; this was measured directly
(dangling-commit count went 1 → 2 across one call). With the author and committer name,
email, and date pinned to fixed values, two successive calls over an unchanged branch
produced a byte-identical object hash, so the object count is bounded by the number of
distinct branch states rather than by the number of sweeps.

Explicit cleanup was rejected. Git offers no way to delete a single unreferenced object;
the available command, `git prune`, is repository-global and would also destroy
unreferenced objects that a concurrent session or tool is still holding. Git's own
auto-gc reclaims them without that hazard.

### Decision 3 — `--merged-only` co-change

**Intended, via a single shared predicate. The two paths do not diverge.** Both commands
ask the same question, and the `--merged-only` help text already promises to remove
worktrees "whose branches are merged into base" — a squash-merged branch is merged, so that
path is under-detecting today for the same reason.

`cleanMergedWorktrees` calls removal with force disabled. Three worktree states were
measured against `git worktree remove` without `--force`; the third is the one that does
**not** refuse:

```
[modified]        rc=128 present=YES  fatal: '...' contains modified or untracked files, use --force to delete it
[untracked-only]  rc=128 present=YES  fatal: '...' contains modified or untracked files, use --force to delete it
[clean-unpushed]  rc=0   present=NO   (removed, no objection)
```

So the no-data-loss argument holds for uncommitted and untracked content, and **does not
hold** for a clean worktree whose branch carries unpushed commits — precisely the state a
broadened predicate newly admits, since a squash-merged branch's worktree is typically
clean. `git status --porcelain` was measured to report an empty string in that state
(`porcelain output: ''`), so the `--stale` path's independent cleanliness guard does not
see it either. The gap exists on both paths, not just `--merged-only`.

This case is **accepted, not solved**, with one mitigation named: REQ-WSM-011 forbids
deleting the branch, so the unpushed commits remain reachable by branch name after the
directory is removed and can be recovered with a fresh `git worktree add`. What is lost is
the working directory, not the commits. Closing the gap properly requires an
upstream-comparison gate on both sweep paths, which is recorded in §4 as out of scope.

The residual asymmetry — `--merged-only` has no explicit porcelain check of its own — is
likewise recorded as an out-of-scope follow-up rather than silently expanded into this
change.

### Decision 4 — Preserved constraints

Recorded as binding requirements REQ-WSM-010 through REQ-WSM-013. The existing
`@MX:ANCHOR` block on `cleanStaleWorktrees` already encodes the two-condition gate and is
retained unchanged.

---

## §6 Traceability

| Requirement | Artifact |
|---|---|
| REQ-WSM-001 .. REQ-WSM-007, REQ-WSM-014 | `internal/core/git/worktree.go` — the merge predicate |
| REQ-WSM-008, REQ-WSM-009 | `internal/core/git/worktree.go` — synthetic commit construction |
| REQ-WSM-010 .. REQ-WSM-012 | `internal/cli/worktree/clean.go` — `cleanStaleWorktrees`, `staleKeepReason` |
| REQ-WSM-013 | `internal/cli/worktree/clean.go` — `cleanMergedWorktrees` |

Acceptance criteria: `acceptance.md`.
