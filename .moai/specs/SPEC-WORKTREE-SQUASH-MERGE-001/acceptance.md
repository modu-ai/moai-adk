# Acceptance Criteria — SPEC-WORKTREE-SQUASH-MERGE-001

---

## §A Evidence Provenance

Two classes of command appear below, and they are labelled distinctly so neither is
mistaken for the other:

- **Authoring-time evidence (executed).** The scenario-construction recipes and the raw git
  probes in §B and §C were run during plan-phase authoring against throwaway repositories,
  and their observed output is recorded verbatim. These establish the *expected verdict*
  for every scenario — the oracle the implementation is measured against.
- **Run-phase judge (to execute).** The `go test` invocations are named per criterion and
  are run during the run phase, once the code exists. They were necessarily not run at
  authoring time against a *passing* implementation; that gap is stated here rather than
  concealed. Every named selector WAS, however, executed against the pre-implementation
  tree so its baseline behaviour is on record — see the non-vacuity discipline below.

**Non-vacuity discipline (applies to every `go test -run` judge in this document).** A
`-run` selector matching nothing exits 0 and prints `testing: warning: no tests to run`,
which would make any criterion judged on exit status vacuously true. Every criterion below
that names a `go test -run` judge is therefore judged by **reading the `--- PASS:` lines**,
not the exit code: at least one `--- PASS:` line must be present, and the observed count is
recorded alongside the verdict. Zero is a FAIL regardless of exit status.

All **thirteen** selectors named in §D were executed against the pre-implementation tree
(`f5129374b`); the observed counts are:

```
IsBranchMerged.*Squash                                            PASS-lines=0   [no tests to run]
IsBranchMerged.*Rebase                                            PASS-lines=0   [no tests to run]
IsBranchMerged.*EmptyDiff                                         PASS-lines=0   [no tests to run]
IsBranchMerged.*Reachab                                           PASS-lines=0   [no tests to run]
IsBranchMerged.*(Partial|Unmerged)                                PASS-lines=0   [no tests to run]
IsBranchMerged.*(Rename|NonASCII|ColonPath|Textconv|Submodule)    PASS-lines=0   [no tests to run]
IsBranchMerged.*(ModeOnly|SymlinkFile)                            PASS-lines=0   [no tests to run]
IsBranchMerged.*ProbeExit                                         PASS-lines=0   [no tests to run]
SyntheticCommit.*Determinis                                       PASS-lines=0   [no tests to run]
IsBranchMerged.*Revert                                            PASS-lines=0   [no tests to run]
TestWorktreeIsBranchMerged                                        PASS-lines=3   ok
TestCleanStale_KeepsDirtyWorktree                                 PASS-lines=1   ok
TestCleanStale_PreviewsByDefault|TestCleanStale_RemovesWithYes    PASS-lines=2   ok
```

The ten zeros name scenario tests this SPEC has yet to add, and each exits **0** despite
matching nothing — which is precisely why the guard is load-bearing rather than decorative.
The three non-zero rows are the pre-existing tests AC-WSM-013, AC-WSM-010, and AC-WSM-016
check 2 pin at 3, 1, and 2 respectively; they are recorded here as the regression floor, and
a later count below those values means tests were lost rather than passing. Criteria whose
baseline output is individually interesting reproduce it verbatim in place.

Two selectors changed at v0.6.0 and both changes are substantive rather than cosmetic.
AC-WSM-006's was widened from `(Rename|NonASCII|ColonPath)` to
`(Rename|NonASCII|ColonPath|Textconv|Submodule)`, because the earlier form would not have
matched a textconv or submodule test — adding the scenarios without widening the selector
would have left them unjudged. AC-WSM-017's is new. Regex-matched against candidate test
names to confirm the widening is real:

```
TestIsBranchMerged_TextconvDriverCollapsesChange   old=no   widened=YES  mode-selector=no
TestIsBranchMerged_SubmoduleIgnoreDirective        old=no   widened=YES  mode-selector=no
TestIsBranchMerged_ModeOnlyDivergence              old=no   widened=no   mode-selector=YES
TestIsBranchMerged_SymlinkFileOIDCollision         old=no   widened=no   mode-selector=YES
TestIsBranchMerged_ColonPathRemovedFromBase        old=YES  widened=YES  mode-selector=no
```

Baseline for the authoring-time evidence: worktree
`.claude/worktrees/clean-stale-squash`, `git rev-parse --short HEAD` → `f5129374b`;
`git --version` → `git version 2.50.1 (Apple Git-155)`; `git config --get diff.renames`
exits 1 (unset), so git's default rename detection is in effect. Each scenario repository
sets `core.quotePath true` explicitly — that is git's **documented default**, and it is set
explicitly because this machine's `~/.gitconfig` overrides it to `false`, which would
otherwise mask the SC-12 defect on this machine only.

The §C oracle was re-executed in full at iteration 6 and extended from thirteen scenarios to
seventeen, and the §C.1 mutation grid was re-executed at 11 × 17 = 187 cells, one freshly
constructed repository per cell. The re-run is a fresh construction, not a transcription of
the iteration-5 tables: every SC-1..SC-13 cell below was observed again rather than carried
over, and every one is unchanged.

One harness error was caught and corrected before reporting, and it is recorded because it
would have produced a false regression report. The first run implemented the `both` mutation
as `weak-guard` + `no-state`, which made SC-7 read `keep` and appeared to contradict the
iteration-5 grid. The grid was right and the harness was wrong: §C.1 defines `weak-guard` as
"S3 accepts any `-` line" but `both` as "drop S5 *and* accept any **non-empty** cherry
output" — two different relaxations of the cherry guard, not the composition of the first two
mutations. SC-7's cherry output is a single `+` line, which is non-empty but contains no `-`,
so the two definitions genuinely diverge on that one cell. Corrected and re-run, `both`
reproduces the recorded grid exactly. The name `both` invites the misreading; the row
definitions in §C.1 are authoritative over the name.

---

## §B Scenario Construction Recipes (executed at authoring time)

Each recipe builds an isolated repository with a `main` branch and a `feat` branch in the
named merge state. All seventeen were executed; §C records the observed signal output.

Every commit is made with an incrementing `GIT_AUTHOR_DATE`/`GIT_COMMITTER_DATE`. This is
load-bearing rather than cosmetic: with identical metadata, commits created within the same
second collide on SHA and the merge-base moves, which silently corrupts the matrix. SC-9 is
the row most exposed to this — see its note below.

| Id | Scenario | Construction |
|---|---|---|
| SC-1 | squash-merged | `feat` adds `x.txt` then `y.txt` in two commits; `main` adds both files in one commit |
| SC-2 | rebase-merged | `feat` adds `p.txt` then `q.txt`; `main` drifts, then cherry-picks both commits individually |
| SC-3 | true merge commit | `feat` adds `m.txt`; `main` runs `git merge --no-ff feat` |
| SC-4 | strictly behind | `feat` created at base and left alone; `main` advances |
| SC-5 | empty-diff branch | `feat` carries one `--allow-empty` commit; `main` advances |
| SC-6 | partially applied | `feat` adds `a.txt` then `b.txt`; `main` adds only `a.txt` |
| SC-7 | fully unmerged (control) | `feat` adds `z.txt`; `main` drifts independently |
| SC-8 | **partially reverted** | `feat` adds `a.txt` then `b.txt`; `main` cherry-picks BOTH, then `git revert --no-edit HEAD` removes the `b` commit |
| SC-9 | base is a strict superset | `feat` adds `s.txt`; `main` **independently adds `s.txt` as its own commit**, then additionally adds `extra.txt` |
| SC-10 | **rename + re-add (squash)** | base commit holds a 10-line `old.txt`; `feat` runs `git mv old.txt new.txt` and commits, then appends a line to `new.txt` and commits — **in two commits**; `main` runs `git merge --squash feat` and commits, then re-adds `old.txt` with its original content |
| SC-11 | **rename + re-add (cherry-pick)** | base commit holds a 10-line `old.txt`; `feat` runs `git mv old.txt new.txt`; `main` cherry-picks that commit, then re-adds `old.txt` with its original content |
| SC-12 | **non-ASCII path removed from base** | `feat` adds `plain.txt` and `문서.txt` in one commit; `main` runs `git merge --squash feat` and commits, then `git rm 문서.txt` and commits |
| SC-13 | **leading-colon path removed from base** | `feat` adds `plain.txt` and `:note.txt` in one commit; `main` runs `git merge --squash feat` and commits, then `git rm :note.txt` and commits. Every builder `add`/`rm` touching the hazard path runs under `git --literal-pathspecs`, so the construction itself is not subject to the defect under test |
| SC-14 | **textconv driver collapses a changed file** | the seed commit adds `.gitattributes` containing `*.pdf diff=pdf`, and the repository sets `diff.pdf.textconv` to a script emitting a constant; `feat` adds `plain.txt` and `report.pdf` in one commit; `main` runs `git merge --squash feat` and commits, then overwrites `report.pdf` with different content and commits |
| SC-15 | **submodule pointer moved under an ignore directive** | a second throwaway repository is built with three commits `s0`/`s1`/`s2`; the parent adds it as a submodule at `sub` pinned to `s0` (with `.gitmodules`) and commits; `feat` moves the pointer to `s1` and adds `plain.txt` in one commit — **`.gitmodules` is not touched**; `main` runs `git merge --squash feat` and commits, then moves the pointer to `s2` and commits. The ignore directive is then set, in either of two locations (see notes) |
| P3c | **mode-only divergence** | the seed commit adds `x.txt`; `feat` chmods it to 755 and adds `plain.txt` in one commit; `main` runs `git merge --squash feat` and commits, then chmods `x.txt` back to 644 and commits. Content is byte-identical on both sides throughout |
| P4c | **symlink/file blob-OID collision** | the seed commit adds a symlink `link.txt -> target` and a regular file `target`; `feat` replaces `link.txt` with a **regular file whose content is exactly `target`** and adds `plain.txt` in one commit; `main` runs `git merge --squash feat` and commits, then restores `link.txt` as a symlink and commits |

Ten construction notes, each recording a way the recipe was got wrong before it was got
right:

- **SC-9 must NOT be built by cherry-picking `feat`'s commit.** A cherry-pick performed
  with no elapsed time reproduces the tree, parent, message, and identity, so the two
  commits collide on SHA, `feat` becomes a literal ancestor of `main`, the merge-base
  moves, and the scenario degenerates into SC-4 with S1 firing. `main` must create its own
  commit adding the same content. `spec.md` §2 finding 3 carries the same warning.
- **SC-10's rename and its edit must be two separate commits.** The recipe's earlier wording
  ("runs `git mv` **then** appends a line") was ambiguous between one commit and two, and the
  two readings do not produce the same matrix row. Built as **one** commit, `feat` has a
  single commit whose patch-id equals `main`'s squash commit, so plain `git cherry` matches
  and **S3 fires** — contradicting §C, which records `S3=keep` for this row. Built as two,
  no single `feat` commit equals the squash and S3 correctly does not fire. The composed
  verdict is `notmerged` either way, so this is a matrix-fidelity issue rather than a
  correctness one; it is recorded because a reader reproducing §C from the one-commit reading
  will see a cell mismatch and reasonably conclude the oracle is wrong.
- **SC-13's hazard path must be created and removed under `--literal-pathspecs`.** A builder
  that runs a bare `git add :note.txt` or `git rm :note.txt` hits the very defect the
  scenario exists to exercise: the pathspec is parsed as magic, the file is never staged or
  never removed, and the scenario silently stops testing anything. The builder therefore uses
  `git --literal-pathspecs add` / `rm`, which keeps the *construction* independent of the
  *mechanism under test*.
- **SC-10 and SC-11 need `old.txt` to exist at the merge-base and to be large enough for
  rename detection to fire.** A one-line file renamed and then extended falls below git's
  50% similarity threshold, no rename is detected, the enumeration is not folded, and the
  scenario silently stops exercising the defect. The recipe uses ten lines; the observed
  detection is `R086` (86% similarity).
- **`git cherry-pick` and `git revert` take no `-q` flag.** Passing one makes the command
  fail; with output redirected the failure is invisible and the scenario is built wrong.
  During the iteration-3 re-run this initially produced a silently un-reverted SC-8 and
  un-cherry-picked SC-2, both of which reported the wrong signals until the flag was
  removed.
- **SC-12 needs `core.quotePath` set to `true` explicitly, and needs `plain.txt` alongside
  the non-ASCII file.** `true` is git's documented default, but a developer machine may
  override it in `~/.gitconfig` (this one does, to `false`), in which case the non-ASCII
  path is *not* quoted and the scenario silently stops exercising the defect. Setting it
  per-repository restores the shipped default rather than contriving one. The second file
  matters for a different reason: with only `문서.txt` in the list, the newline-split
  pathspec would be a single unmatched name and `git diff --quiet` would compare nothing at
  all — the same rc=0, but reached without demonstrating that a *partially* valid list is
  equally unsafe. `plain.txt` is present, matches, and is identical in base, so the probe
  genuinely runs and still returns "no difference".
- **SC-14's `.gitattributes` must be committed at the seed, and must remain in the
  worktree.** The attribute is read from the **checked-out** file, not from the tree under
  judgement, so a fixture that commits `.gitattributes` and then removes it from the working
  directory silently stops exercising the defect. Measured on the built fixture: with the
  worktree `.gitattributes` present the unpatched comparison returns `merged`; delete it from
  the working directory — leaving it committed and still reported by `git ls-files` — and the
  same comparison returns `notmerged`. The `textconv` script must also be executable, and
  `diff.pdf.textconv` must point at an absolute path, or git reports `fatal: cannot run` and
  the driver never fires. The `.pdf` extension is not special; any extension bound by a
  `diff=<driver>` attribute behaves identically.
- **SC-15's `feat` commit must NOT touch `.gitmodules`.** This is the whole point of the
  scenario and the reason the earlier study missed the exposure: when `.gitmodules` also
  differs, that file carries the comparison and the verdict comes out correct for the wrong
  reason. The pointer must move alone. Adding the submodule needs
  `-c protocol.file.allow=always` on modern git, or `git submodule add` refuses a local path.
- **SC-15's ignore directive can live in either of two places, and both were built.** Setting
  `submodule.sub.ignore=all` in `.git/config` models an operator-set exposure; committing the
  same key inside `.gitmodules` models one the repository ships to every clone. Both were
  constructed and both reproduce identically — enumeration `[plain.txt]` without the flag
  versus `[plain.txt, sub]` with it, and the composed verdict flipping accordingly. The
  `.gitmodules` form is the one worth pinning in the run-phase fixture, because it needs no
  local configuration and is therefore the wider exposure.
- **P3c must chmod the file in the working tree, not only in the index.** The first build used
  `git update-index --chmod=+x`, which changes the index without touching the file on disk;
  the next `git checkout` then aborted with `Your local changes to the following files would
  be overwritten by checkout: x.txt` and the scenario was never built. The recipe uses an
  ordinary `chmod` followed by `git add`. `core.fileMode` must be left at its default
  (`true`); a fixture that disables it makes git ignore the mode entirely and the scenario
  stops testing anything.
- **P4c's regular file must contain the symlink's target with no trailing newline.** A symlink
  blob is exactly the target path, unterminated. Writing `target\n` into the regular file
  produces a *different* blob, the two sides no longer share an OID, and the scenario silently
  degenerates into an ordinary content difference — which every mechanism detects, so it would
  pin nothing. Verified on the built fixture: both sides show blob
  `1de565933b05f74c75ff9a6520af5f9f8a5a2f1d`, differing only in mode (`100644` vs `120000`).

---

## §C Signal Oracle (observed at authoring time)

Probe commands, run once per scenario:

```bash
mb=$(git merge-base main feat)
git branch --merged main | sed 's/^[*+ ]*//' | grep -qx feat        # S1 reachability
git diff --quiet "$mb" feat                                         # S2 empty diff
git cherry main feat                                                # S3 plain per-commit
git cherry main "$(git commit-tree "$(git rev-parse feat^{tree})" -p "$mb" -m _)"   # S4 synthetic
names=( )                                                           # S5 state check
while IFS= read -r -d '' l; do names+=("$l"); done \
  < <(git diff --ignore-submodules=none --no-renames --name-only -z "$mb" feat)
[ "${#names[@]}" -eq 0 ] || git --literal-pathspecs diff --quiet --no-textconv \
                                --ignore-submodules=none feat main -- "${names[@]}"
```

Six details of this block are load-bearing and were each got wrong before they were got
right:

- The `sed 's/^[*+ ]*//'` on S1 strips the leading indent as well as the `*`/`+` markers. A
  pattern that strips only the markers fails to match any branch that is not the current
  one, which reads as a false "keep" on SC-3 and SC-4.
- `names` is built as an **array** and expanded as `"${names[@]}"`, one argument per path.
  The earlier `names=$(...)` / `-- $names` string form is not equivalent: see §C.2.
- The enumeration carries `-z` and the read carries `-d ''`, so the list is split on NUL.
  The earlier newline-split form (`read -r l` over `--name-only` without `-z`) is not
  equivalent: git C-quotes paths containing a backslash, tab, double quote, or control
  character, and — under the default `core.quotePath=true` — any non-ASCII path, so a
  newline split yields quoted renderings that match nothing. See §C.2 route 3.
- The comparison carries `--literal-pathspecs`, **before** the `diff` subcommand. It is a
  git-level option; written as `git diff --literal-pathspecs ...` git rejects the invocation
  with a usage error. Without it a byte-perfect path whose first byte is `:` is parsed as
  pathspec magic and matches nothing, and `:!`/`:^` forms parse as EXCLUDE and invert what
  the probe compares. See §C.2 route 4.
- The comparison carries `--no-textconv`, **after** the subcommand — it is a diff-level
  option, and placed before it git exits 129 with `unknown option`. Without it a
  `.gitattributes` `diff=<driver>` rule with a `textconv` driver makes git compare the
  driver's *rendering* of each blob rather than the blobs, so two genuinely different files
  report no difference on a path list that is complete and byte-perfect. See §C.2 route 5.
- **`--ignore-submodules=none` appears twice — once on the enumeration and once on the
  comparison — and neither occurrence is redundant.** `diff.ignoreSubmodules=all` and
  `submodule.<name>.ignore=all` suppress a changed gitlink at *both* stages: without the flag
  on the enumeration the submodule path never enters `names`, and without it on the comparison
  a correct `names` is still ignored. Measured, removing it from either stage alone reproduces
  the SC-15 false positive. It is diff-level in both positions. See §C.2 route 5.

Observed verdicts. `MERGED` on S3/S4 is the **raw** history signal, before the S5
conjunction; `OK`/`NO` is the conjunct. The composed verdict is
`S1 OR S2 OR (S3 AND S5) OR (S4 AND S5)`.

| Id | S1 | S2 | S3 | S4 | S5 | **Required predicate verdict** |
|---|---|---|---|---|---|---|
| SC-1 squash | keep | keep | keep | **MERGED** | OK | **merged** |
| SC-2 rebase | keep | keep | **MERGED** | keep | OK | **merged** |
| SC-3 merge commit | **MERGED** | **EMPTY** | keep | keep | OK | **merged** |
| SC-4 behind | **MERGED** | **EMPTY** | keep | keep | OK | **merged** |
| SC-5 empty diff | keep | **EMPTY** | keep | keep | OK | **merged** |
| SC-6 partial | keep | keep | keep | keep | **NO** | **not merged** |
| SC-7 unmerged | keep | keep | keep | keep | **NO** | **not merged** |
| SC-8 **revert** | keep | keep | **MERGED** | keep | **NO** | **not merged** |
| SC-9 superset | keep | keep | **MERGED** | **MERGED** | OK | **merged** |
| SC-10 **rename+re-add (squash)** | keep | keep | keep | **MERGED** | **NO** | **not merged** |
| SC-11 **rename+re-add (pick)** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-12 **non-ASCII removed** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-13 **leading-colon removed** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-14 **textconv driver** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| SC-15 **submodule ignore=all** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| P3c **mode-only divergence** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |
| P4c **symlink/file OID collision** | keep | keep | **MERGED** | **MERGED** | **NO** | **not merged** |

Verbatim harness output from the **iteration-6 re-run**, under the composed predicate as now
specified (`--ignore-submodules=none --no-renames --name-only -z`, NUL split;
`--literal-pathspecs` git-level and `--no-textconv --ignore-submodules=none` diff-level on
the comparison). Every scenario was rebuilt from the §B recipes rather than transcribed, and
SC-1..SC-13 are cell-for-cell identical to the iteration-5 record. Column order is
`S1 S2 S3 S4 S5`; `Y` = fires, `.` = does not.

```
sc1    [. . . Y Y] merged     required=merged     OK  names=['x.txt', 'y.txt']
sc2    [. . Y . Y] merged     required=merged     OK  names=['p.txt', 'q.txt']
sc3    [Y Y . . Y] merged     required=merged     OK  names=[]
sc4    [Y Y . . Y] merged     required=merged     OK  names=[]
sc5    [. Y . . Y] merged     required=merged     OK  names=[]
sc6    [. . . . .] notmerged  required=notmerged  OK  names=['a.txt', 'b.txt']
sc7    [. . . . .] notmerged  required=notmerged  OK  names=['z.txt']
sc8    [. . Y . .] notmerged  required=notmerged  OK  names=['a.txt', 'b.txt']
sc9    [. . Y Y Y] merged     required=merged     OK  names=['s.txt']
sc10   [. . . Y .] notmerged  required=notmerged  OK  names=['new.txt', 'old.txt']
sc11   [. . Y Y .] notmerged  required=notmerged  OK  names=['new.txt', 'old.txt']
sc12   [. . Y Y .] notmerged  required=notmerged  OK  names=['plain.txt', '문서.txt']
sc13   [. . Y Y .] notmerged  required=notmerged  OK  names=[':note.txt', 'plain.txt']
sc14   [. . Y Y .] notmerged  required=notmerged  OK  names=['plain.txt', 'report.pdf']
sc15   [. . Y Y .] notmerged  required=notmerged  OK  names=['plain.txt', 'sub']
p3c    [. . Y Y .] notmerged  required=notmerged  OK  names=['plain.txt', 'x.txt']
p4c    [. . Y Y .] notmerged  required=notmerged  OK  names=['link.txt', 'plain.txt']
```

The four new rows all carry `S3=Y S4=Y`, which puts them in the sharpest shape the matrix
contains: two independent history signals agree the branch is merged, and S5 is the only
thing that can withhold the verdict. Their `names` lists are worth reading alongside the
verdicts — every one is complete and byte-perfect, so none of the four round-trip repairs is
what saves them:

- **SC-14** — `['plain.txt', 'report.pdf']`. The changed file *is* in the list. Only the
  comparison's diff semantics decide the verdict.
- **SC-15** — `['plain.txt', 'sub']`. The gitlink is present **because** the enumeration
  carries `--ignore-submodules=none`; without it the list is `['plain.txt']` and the moved
  pointer is never mentioned. This row is the one that fails at both stages.
- **P3c / P4c** — the diverging path is present and the comparison detects it, because
  `git diff` compares modes. These two rows pass as specified and exist to fail a future
  mode-blind comparison.

Ground truth for the four, recorded so the verdicts are independently checkable rather than
asserted:

```
sc14  feat:report.pdf = dcc7c925c1f3fb918a192d6f40835189490d4e4c
      main:report.pdf = 43d844c81ab8ccc32d09beb9d70f5831d1474402      <- blobs differ
      git diff --name-only feat main -- report.pdf   ->  report.pdf   <- still reported
      cmp bare                                        ->  rc=0        (OK,  vacuous)
      cmp --no-textconv                               ->  rc=1        (NO,  correct)

sc15  feat:sub = 1a7d9a656cfc3f54f58da6d184dcbc0c33a29cb3
      main:sub = 034452180de5fd3d9b2bba76a8bc15273cad6386             <- pointers differ
      enum without --ignore-submodules=none  ->  ['plain.txt']        <- gitlink DROPPED
      enum with    --ignore-submodules=none  ->  ['plain.txt', 'sub']
      cmp without the flag, correct list      ->  rc=0                (OK,  vacuous)
      cmp with    the flag, correct list      ->  rc=1                (NO,  correct)

p3c   feat 100755 blob 587be6b4c3f93f93c489c0111bba5596147a26cb  x.txt
      main 100644 blob 587be6b4c3f93f93c489c0111bba5596147a26cb  x.txt   <- ONE blob, two modes
      git diff comparison   ->  rc=1   detects
      blob-OID comparison   ->  EQUAL  MISSES

p4c   feat 100644 blob 1de565933b05f74c75ff9a6520af5f9f8a5a2f1d  link.txt
      main 120000 blob 1de565933b05f74c75ff9a6520af5f9f8a5a2f1d  link.txt <- ONE blob, two modes
      git diff comparison   ->  rc=1   detects
      blob-OID comparison   ->  EQUAL  MISSES
```

The SC-15 block also confirms the enumeration-side flag is load-bearing rather than a
duplicate of the comparison-side one: the two `enum` lines differ, and the shorter list
cannot be repaired by any comparison flag, because the path it omits is never passed.

Two further properties of SC-14 and SC-15, measured, because they make these members harder
to reason about than their predecessors:

```
sc14  worktree .gitattributes present  ->  unpatched cmp: merged
      worktree .gitattributes removed  ->  unpatched cmp: notmerged
      ...though git ls-files still reports .gitattributes as tracked

sc15  directive in .git/config                    ->  enum without flag: ['plain.txt']
      directive committed inside .gitmodules      ->  enum without flag: ['plain.txt']
      (.git/config key unset in the second case; .gitmodules is a TRACKED file)
```

The first says the verdict depends on the state of the judging **worktree**, not only on the
trees under comparison — a sensitivity none of the first four axes had. The second says this
exposure does not require the operator's own configuration: a repository can commit the
directive and ship it to every clone.

SC-13's `names` list is the row worth pausing on: the enumeration emits `:note.txt`
byte-perfectly, so nothing in the *list* is wrong. The defect lived entirely in how that
correct list was read back, which is why neither `--no-renames` nor `-z` touched it. The
isolated pair, in the same repository:

```
git diff --quiet feat main -- ':note.txt'                       rc=0   (OK,  vacuous)
git --literal-pathspecs diff --quiet feat main -- ':note.txt'   rc=1   (NO,  correct)
git diff --literal-pathspecs --quiet feat main -- ':note.txt'   usage error — git-level option
```

Retained verbatim from the iteration-4 re-run — the **encoding** axis (newline split vs
`-z`), which iteration 5 did not re-run because it changed nothing on that axis. Each
scenario is probed twice, both
times with `--no-renames`: once with the output split on **newline** (the v0.3.0 form) and
once with `-z` and the output split on **NUL** (the current form). The `names` line records
the two path lists. This is the axis iteration 4 tests; the rename-fold axis tested at
iteration 3 is preserved separately in the §C.1 mutation grid, where `folded-names` is a
row.

```
sc1   nl-split   keep keep keep MERGED OK merged
sc1   -z         keep keep keep MERGED OK merged
sc1     names(nl-split)=[x.txt y.txt]  names(-z)=[x.txt y.txt]

sc2   nl-split   keep keep MERGED keep OK merged
sc2   -z         keep keep MERGED keep OK merged
sc2     names(nl-split)=[p.txt q.txt]  names(-z)=[p.txt q.txt]

sc3   nl-split   MERGED EMPTY keep keep OK merged
sc3   -z         MERGED EMPTY keep keep OK merged
sc3     names(nl-split)=[]  names(-z)=[]

sc4   nl-split   MERGED EMPTY keep keep OK merged
sc4   -z         MERGED EMPTY keep keep OK merged
sc4     names(nl-split)=[]  names(-z)=[]

sc5   nl-split   keep EMPTY keep keep OK merged
sc5   -z         keep EMPTY keep keep OK merged
sc5     names(nl-split)=[]  names(-z)=[]

sc6   nl-split   keep keep keep keep NO notmerged
sc6   -z         keep keep keep keep NO notmerged
sc6     names(nl-split)=[a.txt b.txt]  names(-z)=[a.txt b.txt]

sc7   nl-split   keep keep keep keep NO notmerged
sc7   -z         keep keep keep keep NO notmerged
sc7     names(nl-split)=[z.txt]  names(-z)=[z.txt]

sc8   nl-split   keep keep MERGED keep NO notmerged
sc8   -z         keep keep MERGED keep NO notmerged
sc8     names(nl-split)=[a.txt b.txt]  names(-z)=[a.txt b.txt]

sc9   nl-split   keep keep MERGED MERGED OK merged
sc9   -z         keep keep MERGED MERGED OK merged
sc9     names(nl-split)=[s.txt]  names(-z)=[s.txt]

sc10  nl-split   keep keep keep MERGED NO notmerged
sc10  -z         keep keep keep MERGED NO notmerged
sc10    names(nl-split)=[new.txt old.txt]  names(-z)=[new.txt old.txt]

sc11  nl-split   keep keep MERGED MERGED NO notmerged
sc11  -z         keep keep MERGED MERGED NO notmerged
sc11    names(nl-split)=[new.txt old.txt]  names(-z)=[new.txt old.txt]

sc12  nl-split   keep keep MERGED MERGED OK merged
sc12  -z         keep keep MERGED MERGED NO notmerged
sc12    names(nl-split)=[plain.txt "\353\254\270\354\204\234.txt"]  names(-z)=[plain.txt 문서.txt]
```

(Column order per row: `S1 S2 S3 S4 S5 verdict`.)

Four readings of this output:

- **SC-1 through SC-11 are cell-for-cell identical in the two columns, and identical to the
  values recorded at v0.3.0.** `-z` regresses nothing. Their `names` lists are also
  identical, which is the reason: none of those eleven scenarios contains a path git would
  quote, so there is nothing for the encoding to corrupt.
- **SC-12 differs in exactly one cell — S5 — and that one cell decides the verdict.** Under
  the newline split the path list's second element is the *quoted rendering*
  `"\353\254\270\354\204\234.txt"`, which names no file. `git diff --quiet` therefore
  compares only `plain.txt`, which is identical in base, and exits 0. S5 reads `OK` and both
  history signals are granted unopposed: **merged**, a false positive on a branch whose file
  is genuinely absent from base.
- **SC-12 fires both S3 and S4**, which makes it the same strong shape as SC-11: two
  independent history signals agree the branch is merged, and the enumeration is the only
  thing preventing a wrong removal.
- **SC-2 remains the row the state conjunct was most likely to regress**, because
  rebase-merge detection rests on exactly the history semantic S5 narrows. It does not
  regress under either setting: the individually replayed patches survive in base HEAD, so
  S5 holds and the S3 verdict stands.

Retained verbatim from the iteration-3 re-run — the **rename** axis (`--name-only` with
rename detection on, "folded", versus `--no-renames`), which iteration 4 did not re-run
because it changed nothing on that axis. It is preserved rather than replaced because it is
the recorded evidence that `--no-renames` flips SC-10 and SC-11 and regresses SC-1..SC-9:

```
sc1   folded       S1=keep    S2=keep   S3=keep    S4=MERGED  S5=OK  => merged
sc1   --no-renames S1=keep    S2=keep   S3=keep    S4=MERGED  S5=OK  => merged
sc1     names(folded)=[x.txt y.txt ] names(--no-renames)=[x.txt y.txt ]

sc2   folded       S1=keep    S2=keep   S3=MERGED  S4=keep    S5=OK  => merged
sc2   --no-renames S1=keep    S2=keep   S3=MERGED  S4=keep    S5=OK  => merged
sc2     names(folded)=[p.txt q.txt ] names(--no-renames)=[p.txt q.txt ]

sc3   folded       S1=MERGED  S2=EMPTY  S3=keep    S4=keep    S5=OK  => merged
sc3   --no-renames S1=MERGED  S2=EMPTY  S3=keep    S4=keep    S5=OK  => merged
sc3     names(folded)=[] names(--no-renames)=[]

sc4   folded       S1=MERGED  S2=EMPTY  S3=keep    S4=keep    S5=OK  => merged
sc4   --no-renames S1=MERGED  S2=EMPTY  S3=keep    S4=keep    S5=OK  => merged
sc4     names(folded)=[] names(--no-renames)=[]

sc5   folded       S1=keep    S2=EMPTY  S3=keep    S4=keep    S5=OK  => merged
sc5   --no-renames S1=keep    S2=EMPTY  S3=keep    S4=keep    S5=OK  => merged
sc5     names(folded)=[] names(--no-renames)=[]

sc6   folded       S1=keep    S2=keep   S3=keep    S4=keep    S5=NO  => not merged
sc6   --no-renames S1=keep    S2=keep   S3=keep    S4=keep    S5=NO  => not merged
sc6     names(folded)=[a.txt b.txt ] names(--no-renames)=[a.txt b.txt ]

sc7   folded       S1=keep    S2=keep   S3=keep    S4=keep    S5=NO  => not merged
sc7   --no-renames S1=keep    S2=keep   S3=keep    S4=keep    S5=NO  => not merged
sc7     names(folded)=[z.txt ] names(--no-renames)=[z.txt ]

sc8   folded       S1=keep    S2=keep   S3=MERGED  S4=keep    S5=NO  => not merged
sc8   --no-renames S1=keep    S2=keep   S3=MERGED  S4=keep    S5=NO  => not merged
sc8     names(folded)=[a.txt b.txt ] names(--no-renames)=[a.txt b.txt ]

sc9   folded       S1=keep    S2=keep   S3=MERGED  S4=MERGED  S5=OK  => merged
sc9   --no-renames S1=keep    S2=keep   S3=MERGED  S4=MERGED  S5=OK  => merged
sc9     names(folded)=[s.txt ] names(--no-renames)=[s.txt ]

sc10  folded       S1=keep    S2=keep   S3=keep    S4=MERGED  S5=OK  => merged
sc10  --no-renames S1=keep    S2=keep   S3=keep    S4=MERGED  S5=NO  => not merged
sc10    names(folded)=[new.txt ] names(--no-renames)=[new.txt old.txt ]

sc11  folded       S1=keep    S2=keep   S3=MERGED  S4=MERGED  S5=OK  => merged
sc11  --no-renames S1=keep    S2=keep   S3=MERGED  S4=MERGED  S5=NO  => not merged
sc11    names(folded)=[new.txt ] names(--no-renames)=[new.txt old.txt ]
```

Under the folded enumeration SC-10 and SC-11's path list omits `old.txt`, S5 never examines
the path the branch deleted, S5 reads `OK`, and the history signal is granted unopposed.
The two axes are independent: `--no-renames` decides *which* paths are enumerated, `-z`
decides *what form they arrive in*, and the §C.1 grid confirms each mutation flips a
disjoint set of rows.

Raw output retained for the five must-keep rows (iteration-3 re-run; hashes are from
freshly constructed repositories, so they differ from the iteration-2 record while the
`-`/`+` shape — which is what the guards read — is what matters):

```
### sc6
  raw plain cherry : - e910b712441b5c3fa46b6ef4e038397b16fcd524|+ c066fd394b94381250d5ea20cdf2c23787a8eaa8|
  raw synth cherry : + 83aa061e1be0161e040fc0c782a59411c269b6f9|
  names            : a.txt b.txt
  state conjunct   : NO
### sc7
  raw plain cherry : + a46e853312ec914c3cc9dd76d69be879582dcc79|
  raw synth cherry : + 31b81a50c8d70917e4af5d78327dfa028028e464|
  names            : z.txt
  state conjunct   : NO
### sc8
  raw plain cherry : - f4ac9d1a06963c1589805cd8a02cd9b40da3cfcf|- 7d5687827616e801cf7c6322cc9ceb00ed82052e|
  raw synth cherry : + a7707d8ad4f5b5314a4551f3a5e3d4c79b94ea01|
  names            : a.txt b.txt
  state conjunct   : NO
### sc10
  raw plain cherry : + 6b2ed19d91c612ad0751c8e6faaf7e4a86add2dc|+ 592764111bfb217e8ea91ffc65f8e4bb297a8ffc|
  raw synth cherry : - cec712537b4f91f29dfb1d0ad5d78320d7e9e083|
  names            : new.txt old.txt
  state conjunct   : NO
### sc11
  raw plain cherry : - 1fa5d3634dcd1bfa1d38c21b96fbf5f37d7ea3c2|
  raw synth cherry : - 025a5e8f75eb8c67cf41a5b1a26f257322dc818c|
  names            : new.txt old.txt
  state conjunct   : NO
```

Raw output for SC-12 (iteration-4 construction), recorded in the same shape:

```
### sc12
  core.quotePath   : true                          (git's documented default)
  main tree        : plain.txt seed.txt
  feat tree        : plain.txt seed.txt "\353\254\270\354\204\234.txt"
  raw plain cherry : - 17f8ea59bade9eba9d08ba7cfc9c4518b31cde0b|
  raw synth cherry : - 35252ca72076c4ee755b469c4c74ed720393d41a|
  names(nl-split)  : plain.txt "\353\254\270\354\204\234.txt"
  names(-z)        : plain.txt 문서.txt
  S5 nl-split      : git diff --quiet feat main -- plain.txt '"\353\254\270\354\204\234.txt"'  -> rc=0   (OK,  vacuous)
  S5 -z            : git diff --quiet feat main -- plain.txt 문서.txt                          -> rc=1   (NO,  correct)
```

Note `feat`'s tree listing: `git ls-tree --name-only` quotes the path exactly as
`--name-only` does, which is the same rendering that fails to round-trip. The two S5 lines
are the whole defect in two commands — identical predicate, identical repository, opposite
verdicts, differing only in how the path list was read.

Four rows are load-bearing, each for a different guard:

- **SC-8** — both plain-cherry lines are `-`, so the every-line-`-` guard is satisfied and a
  history-only predicate returns merged while `b.txt` is absent from base HEAD. S5 is what
  withholds it.
- **SC-10** — the synth cherry is a single `-` line, so S4 fires. Only S5 withholds, and
  only because `old.txt` is in the `names` list. Note the plain cherry is `+ | +`, so S3
  does not fire here: SC-10 is an S4-only false positive.
- **SC-11** — *both* cherry probes report `-`, so S3 and S4 both fire. This is the strongest
  form of the rename defect: two independent history signals agree the branch is merged,
  and the enumeration is the only thing standing between them and a wrong removal.
- **SC-12** — both cherry probes again report `-`, so S3 and S4 both fire, and here the
  enumeration is not merely incomplete but *mis-encoded*. The `names(-z)` and
  `names(nl-split)` lines differ in one element and that element is the whole verdict.

### §C.2 Why S5 fails open — the unmatched-pathspec property

`git diff --quiet` treats a pathspec that matches nothing as "no difference" and exits **0**.
It does not warn, and it does not fail. Measured:

```
git diff --quiet feat main -- no-such-path.txt   ->  rc=0    <- matches nothing
git diff --quiet feat main -- b.txt              ->  rc=1    <- genuinely differs
git diff        feat main -- no-such-path.txt    ->  rc=0, empty stdout (no error at all)
```

Every defect that shrinks, corrupts, or re-interprets the path list therefore lands in the
**unsafe** direction: S5 holds, the conjunct is satisfied, and the branch is reported merged.
A fifth route reaches the same outcome without touching the list at all — the comparison
itself answers wrongly — and a sixth would open if the comparison mechanism were ever swapped
for a mode-blind one. Six distinct routes were reproduced:

**Route 1 — rename detection folds the deleted source path away** (the SC-10/SC-11 defect).
`git diff --name-only` reports only a rename's destination:

```
git diff --name-status <mb> feat          ->  R086  old.txt  new.txt
git diff --name-only <mb> feat            ->  [new.txt]              <- old.txt ABSENT
git diff --no-renames --name-only <mb> feat -> [new.txt old.txt]     <- old.txt PRESENT
```

**Route 2 — the path list is joined into one shell string.** Passing the list as a single
argument produces one pathspec that matches nothing, which by the property above reads as
merged. On a repository where the correct answer is `rc=1`:

```
argv: <feat> <main> <--> <old.txt >    (list joined into one argument)   rc=0   <- WRONG
argv: <feat> <main> <--> <old.txt>     (one argument per path)           rc=1   <- right
```

The joined form is shell-dependent, which is what makes it treacherous rather than merely
wrong — the same `-- $names` snippet behaves differently depending on the interpreter:

```
under bash :  unquoted $n rc=1   literal rc=1     <- splits, survives by luck
under zsh  :  unquoted $n rc=0   literal rc=1     <- does not split, silently wrong
```

**Route 3 — the path list is split on newline, so quoted paths never round-trip** (the
SC-12 defect). `git diff --name-only` does not emit paths verbatim; it emits a C-quoted
rendering whenever the path contains a backslash, tab, double quote, or control character,
and — under git's documented default `core.quotePath=true` — whenever it contains any
non-ASCII byte. Measured, by round-tripping each emitted line straight back as a pathspec
in a repository where the correct answer is `rc=1` for every path:

```
core.quotePath=true   (git's documented default)
  "a\\b.txt"                      rc=0   <- matched NOTHING
  "caf\303\251.txt"               rc=0   <- matched NOTHING
  plainname.txt                   rc=1   (matched)
  "tab\tname.txt"                 rc=0   <- matched NOTHING
  "\355\225\234\352\270\200.txt"  rc=0   <- matched NOTHING

core.quotePath=false  (which names are STILL quoted)
  "a\\b.txt"                      rc=0   <- backslash quoted regardless of the setting
  café.txt                        rc=1   (unquoted under quotePath=false)
  plainname.txt                   rc=1
  "tab\tname.txt"                 rc=0   <- control char quoted regardless of the setting
  한글.txt                         rc=1   (unquoted under quotePath=false)

-z (NUL-delimited), quotePath=true — never quoted, round-trip correct for all five
  a\b.txt   café.txt   plainname.txt   tab<TAB>name.txt   한글.txt      all rc=1
```

**Route 4 — the byte-perfect path is parsed as a pathspec expression, not a filename** (the
SC-13 defect). This route is not reached by corrupting the list; it is reached with a list
that is already correct. Git parses a pathspec's leading characters as *magic* before any
matching occurs, so a repo-root path whose first byte is `:` names no file. Measured in a
repository where the correct answer is `rc=1` in every row:

```
bare pathspec
  :root.txt        rc=0   <- parsed as magic, matched NOTHING
  :(glob)g.txt     rc=0   <- parsed as long-form :(glob) magic + path g.txt
  :!bang.txt       rc=1   but as EXCLUDE magic -- the comparison is inverted, not matched
  :^caret.txt      rc=1   but as EXCLUDE magic -- likewise
  dir/:nested.txt  rc=1   safe (the colon is not the pathspec's first byte)
  mid:colon.txt    rc=1   safe

--literal-pathspecs -- all six matched as literal filenames, all rc=1
```

The two EXCLUDE rows are the subtle ones: `:!x` and `:^x` do not merely fail to match, they
turn that element into a *negative* pathspec, so the probe ends up comparing everything
*except* the file it needed to check. In a composed scenario with an identical sibling file
present, that inversion yields `rc=0` and a merged false positive.

The interpretation axis has exactly one open symbol class. Glob metacharacters and
backslashes were measured to round-trip correctly even under a bare pathspec, because git's
matcher attempts a literal comparison in addition to wildmatch — so a file literally named
`*star.txt` is matched by the pathspec `*star.txt`. Measured in a repository where every
glob-matchable sibling is held identical on both branches (so a spurious match cannot mask
the result) and the correct answer is `rc=1` in every row:

```
path           bare       literal
*star.txt      rc=1       rc=1
?q.txt         rc=1       rc=1
[a]x.txt       rc=1       rc=1
a\b.txt        rc=1       rc=1
:lead.txt      rc=0       rc=1     <- the only class the bare form leaves open
```

A leading `:` was therefore the only symbol class still open, and `--literal-pathspecs`
closes it by disabling *all* pathspec magic rather than by enumerating hazardous symbols.
That distinction is what makes this an axis closure rather than a fifth point patch.

**Route 5 — the path list is correct and the comparison is configured not to notice** (the
SC-14 and SC-15 defects). This route is not reached by damaging the list in any way. The list
is complete, byte-perfect, and matched literally, and `git diff --quiet` still reports
equality — because equality is decided under the repository's diff configuration, and that
configuration can be narrowed.

**Twenty-two configuration keys, attribute forms, and environment settings were measured**
against an ordinary-file sharp scenario: `feat` adds `f.dat` plus `plain.txt`; `main`
squash-merges so S3 and S4 both fire, then overwrites `f.dat` with genuinely different
content. Required `notmerged` in every row; `MERGED` marks a configuration that makes the
unpatched comparison report equality on a correct path list.

```
probe                                verdict   affects?
diff.algorithm=histogram             keep      no
diff.algorithm=patience              keep      no
diff.algorithm=minimal               keep      no
diff.autoRefreshIndex=false          keep      no
core.fileMode=false                  keep      no      (affects the index, not tree-to-tree)
core.symlinks=false                  keep      no
diff.noprefix=true                   keep      no
diff.renames=copies                  keep      no
core.autocrlf=true                   keep      no      (filters do not apply tree-to-tree)
diff.suppressBlankEmpty=true         keep      no
diff.wsErrorHighlight=all            keep      no
diff.context=0                       keep      no
core.quotePath=false                 keep      no
gitattr  f.dat -diff (binary marker) keep      no
gitattr  f.dat binary  macro         keep      no
gitattr  f.dat filter= clean/smudge  keep      no      (blobs are already clean)
gitattr  f.dat text eol=lf +autocrlf keep      no
gitattr  f.dat diff=x + diff.x.textconv   MERGED  YES  <- member 1
diff.x.cachetextconv=true            MERGED    YES     (same member, not a distinct one)
diff.ignoreSubmodules=all            keep      no      <- see the scope note below
external diff driver diff.x.command  keep      no      (see below)
GIT_EXTERNAL_DIFF env                keep      no      (see below)
```

**Scope note, and it matters — this sweep cannot find the submodule members.** The fixture
above contains only ordinary files, and `diff.ignoreSubmodules` and `submodule.<n>.ignore`
act exclusively on **gitlinks**. Their `keep` rows here mean "not applicable to this
scenario", not "harmless": measured against a fixture that *does* contain a submodule
(SC-15), both suppress the change and flip the verdict. Reading this table alone would
therefore undercount the axis at one member. The honest total is **three members** — one
found by this sweep and two by SC-15 — and the reason the earlier study's equivalent table
recorded `diff.ignoreSubmodules` as a member is that it evaluated that row against a
submodule fixture rather than a file one. A future sweep of this axis must vary the *object
kind* (file, symlink, gitlink) as well as the configuration, or it will keep missing members
of exactly this shape.

The external-diff result is worth stating precisely, because it looked like a likely fourth
member and is not one. The driver **does** run on the ordinary diff path, but `--quiet`
short-circuits before consulting it:

```
git diff feat main -- f.dat                            # driver runs, prints its output
git --literal-pathspecs diff --quiet feat main -- f.dat   rc=1   <- correct; driver never consulted
```

This is the same asymmetry as textconv in the opposite direction, and together the two bound
what was probed: of everything measured — the twenty-two-probe file sweep above plus the
submodule fixture — **exactly one member is comparison-side only** (textconv) and **two are
dual-stage** (the submodule keys). The three members and their stage classification:

```
member                                    enumeration   comparison   closed by
.gitattributes diff=X + diff.X.textconv    unaffected    SUPPRESSED   --no-textconv
diff.ignoreSubmodules=all                  SUPPRESSED    SUPPRESSED   --ignore-submodules=none
submodule.<name>.ignore=all                SUPPRESSED    SUPPRESSED   --ignore-submodules=none
```

Measured for member 1 (textconv), on two genuinely different blobs — note the name-listing
path is **not** fooled, which is what makes it comparison-side only:

```
git diff --name-only            feat main -- report.pdf  ->  report.pdf              (still reported)
git diff --name-status          feat main -- report.pdf  ->  M  report.pdf           (still reported)
git diff --quiet                feat main -- report.pdf  ->  rc=0   <- ONLY this is fooled
git diff --quiet --no-textconv  feat main -- report.pdf  ->  rc=1
```

Measured for members 2 and 3 (submodule ignore), which suppress one stage earlier as well:

```
git diff --no-renames --name-only -z <mb> feat                        ->  ['plain.txt']
git diff --ignore-submodules=none --no-renames --name-only -z <mb> feat -> ['plain.txt','sub']
comparison without --ignore-submodules=none, given the correct list   ->  rc=0
comparison with    --ignore-submodules=none, given the correct list   ->  rc=1
```

Two properties make this route nastier than its four predecessors. Both values are read from
the **checked-out worktree** rather than from the trees under comparison, so the verdict
depends on checkout state — deleting the worktree's `.gitattributes` makes SC-14's false
positive vanish while the file remains committed and tracked. And `submodule.<name>.ignore`
is readable from `.gitmodules`, a **tracked** file, so unlike every other member of every
axis this exposure can be shipped inside a repository with no operator configuration at all.

**Route 6 — the comparison is replaced by something that cannot see a file's mode** (the P3c
and P4c defects). This route is **not open** in the specified mechanism: `git diff` compares
modes. It is recorded because the natural-looking repair for route 5 — comparing blob object
IDs, which no diff configuration can influence — opens it. A symlink and a regular file whose
content is exactly the symlink's target share **one blob OID**, and a `chmod` changes no
content at all:

```
P3c   feat 100755 blob 587be6b4…  x.txt      main 100644 blob 587be6b4…  x.txt
P4c   feat 100644 blob 1de56593…  link.txt   main 120000 blob 1de56593…  link.txt

git diff comparison   ->  rc=1   detects both
blob-OID comparison   ->  EQUAL  misses both
```

Unlike every member of route 5, this failure needs **no configuration whatsoever** — no
`.gitattributes`, no config key, no tracked directive. It is therefore a strictly worse
exposure than the one it would be adopted to fix, which is why `spec.md` §2 records the
OID-comparison hybrid as weighed and rejected, and why AC-WSM-017 pins these two rows against
a future mechanism swap.

The second block of route 3 is the one that matters most for a future reader:
**`core.quotePath=false` is not a fix.** It silences the non-ASCII half of the hazard and leaves the backslash and
control-character halves intact, so a repair phrased as "set `core.quotePath=false`" would
close part of the hole while reading as though it closed all of it. Only `-z` removes the
round-trip hazard for every byte, under every configuration.

The implementation consequence is stated in `plan.md` §D: the Go code builds `names` by
splitting `git diff --ignore-submodules=none --no-renames --name-only -z` output on `\x00`
(never on `\n`), passes it as a `[]string` argument list to the exec helper without ever
reconstructing a shell string, and issues the comparison as
`git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none …`. The shell
snippets in this document use `-z` with `read -r -d ''` into an array, `--literal-pathspecs`
before the subcommand, and `--no-textconv --ignore-submodules=none` after it, for all six
reasons.

Raw output retained for the three must-keep rows:

```
### SC-6 partial
  raw plain cherry : - cbac4fc9510a788dcbee90aab3bcbd987e3c4ebe | + c24465c5613dc4ce1f6bf9593bf29b4a17929385
  raw synth cherry : + 8c7d8354af073122424d401e43b42ec26b62a1cb
  state conjunct   : NO

### SC-7 unmerged
  raw plain cherry : + f23f4eb6e1a07d17defd9a302665b496fbed845d
  raw synth cherry : + ab9244e7372d4f57f996a13a59d8ed2b4d555a81
  state conjunct   : NO

### SC-8 revert
  raw plain cherry : - 616bf026a585f183d8c49ec58de025fb68fc4d15 | - fb4444864adbdc8bed6c4f61bd35d172551c4054
  raw synth cherry : + 4636d524ebeb559c3b41ae48cb48804d2a5fa17a
  state conjunct   : NO
```

SC-8 is the load-bearing row: both plain-cherry lines are `-`, so the S3 guard is satisfied
and a history-only predicate returns merged while `b.txt` is absent from base HEAD.

### §C.1 Mutation table (which mutation actually falsifies which row)

Executed by simulating the predicate under each mutation, one freshly constructed
repository per cell. This table is the authority for the falsification procedures in §D —
an untested falsification is how an earlier iteration shipped a guard that could not be
falsified by its own stated procedure.

Verbatim output (iteration-6 re-run at **11 mutations × 17 scenarios**, every one of the 187
cells freshly executed; the seven prior mutations reproduce the iteration-5 record on the
columns it carried, and `no-textconv`, `no-ignoresub-cmp`, `no-ignoresub-enum`, and
`oid-only-cmp` are the new rows). The grid runs across the full scenario set, so a mutation's
blast radius is visible on every row rather than only on the rows it was expected to touch:

```
mutation           sc1    sc2    sc3    sc4    sc5    sc6    sc7    sc8    sc9    sc10   sc11   sc12   sc13   sc14   sc15   p3c    p4c
baseline           MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   keep   keep   keep   keep   keep
weak-guard         MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   keep   keep   keep   keep   keep
no-state           MERGED MERGED MERGED MERGED MERGED keep   keep   MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED
both               MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED MERGED
folded-names       MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED MERGED MERGED keep   keep   keep   keep   keep   keep
newline-split      MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   MERGED keep   keep   keep   keep   keep
no-literal         MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   MERGED keep   keep   keep   keep
no-textconv        MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   keep   MERGED keep   keep   keep
no-ignoresub-cmp   MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   keep   keep   MERGED keep   keep
no-ignoresub-enum  MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   keep   keep   MERGED keep   keep
oid-only-cmp       MERGED MERGED MERGED MERGED MERGED keep   keep   keep   MERGED keep   keep   keep   keep   keep   keep   MERGED MERGED
```

Blast radius per mutation, restricted to the must-keep rows (computed from the grid above
rather than asserted):

```
weak-guard         -> (none)
no-state           -> sc8 sc10 sc11 sc12 sc13 sc14 sc15 p3c p4c
both               -> sc6 sc8 sc10 sc11 sc12 sc13 sc14 sc15 p3c p4c
folded-names       -> sc10 sc11
newline-split      -> sc12
no-literal         -> sc13
no-textconv        -> sc14
no-ignoresub-cmp   -> sc15
no-ignoresub-enum  -> sc15
oid-only-cmp       -> p3c p4c
```

**The SC-1..SC-13 columns are cell-for-cell identical to the iteration-5 grid** across every
mutation that grid carried, so the two new flags and the four new scenarios regress nothing.
The four new mutation rows leave all thirteen original columns unchanged, which is the
evidence the flags are additive rather than behaviour-altering.

**Correction to the iteration-4 grid.** That grid's `folded-names` row recorded `sc12` as
`MERGED`, which contradicted its own summary table and the prose immediately below it (both
of which said `keep`, and the criterion's rationale depends on `keep`). Re-measured here, the
correct value is **`keep`**: `folded-names` retains `-z`, so SC-12's path arrives unquoted
and S5 still withholds. The `MERGED` cell was a transcription of the *`folded-B`* variant
(rename detection on **and** `-z` dropped) into a row labelled for the `-z`-retaining one —
the same conflation N9 identifies in AC-WSM-006's falsification wording, surfacing a second
time in the grid itself. Both are corrected in this revision, and the row label now states
`-z` **retained** explicitly.

The must-keep rows are SC-6, SC-7, SC-8, SC-10 through SC-15, P3c, and P4c; the remaining six
are must-merge and read `MERGED` under every mutation, which is the sanity check on the grid
— the mutations narrow the predicate rather than break it outright. Restricted to the
must-keep columns plus SC-2 (the rebase row most at risk of regression); `M` = MERGED,
`k` = keep, and a **bold** `M` marks a flip from baseline:

| Mutation | SC-2 | SC-6 | SC-7 | SC-8 | SC-10 | SC-11 | SC-12 | SC-13 | SC-14 | SC-15 | P3c | P4c |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| baseline (as designed) | M | k | k | k | k | k | k | k | k | k | k | k |
| `weak-guard` — S3 accepts *any* `-` line instead of *every* | M | k | k | k | k | k | k | k | k | k | k | k |
| `no-state` — drop the S5 conjunct | M | k | k | **M** | **M** | **M** | **M** | **M** | **M** | **M** | **M** | **M** |
| `both` — drop S5 *and* accept any **non-empty** cherry output | M | **M** | **M** | **M** | **M** | **M** | **M** | **M** | **M** | **M** | **M** | **M** |
| `folded-names` — enumerate with `--name-only -z` (rename detection on, **`-z` retained**) | M | k | k | k | **M** | **M** | k | k | k | k | k | k |
| `newline-split` — enumerate with `--no-renames --name-only` (no `-z`) and split on `\n` | M | k | k | k | k | k | **M** | k | k | k | k | k |
| `no-literal` — enumeration unchanged; drop `--literal-pathspecs` from the comparison | M | k | k | k | k | k | k | **M** | k | k | k | k |
| `no-textconv` — enumeration unchanged; drop `--no-textconv` from the comparison | M | k | k | k | k | k | k | k | **M** | k | k | k |
| `no-ignoresub-cmp` — drop `--ignore-submodules=none` from the **comparison** only | M | k | k | k | k | k | k | k | k | **M** | k | k |
| `no-ignoresub-enum` — drop `--ignore-submodules=none` from the **enumeration** only | M | k | k | k | k | k | k | k | k | **M** | k | k |
| `oid-only-cmp` — replace the comparison with a mode-blind blob-OID comparison | M | k | k | k | k | k | k | k | k | k | **M** | **M** |

Note `weak-guard` and `both` relax the cherry guard **differently** — the first accepts any
`-` line, the second accepts any non-empty output — so `both` is not the composition of
`weak-guard` and `no-state`. SC-7 is the cell where they diverge: its cherry output is a
single `+` line, which is non-empty but contains no `-`. A harness that treated `both` as
`weak-guard` + `no-state` reported SC-7 as `keep` and appeared to contradict this grid; the
grid is correct and the row definitions here are authoritative over the row *names*.

Eight consequences, which together fix the falsification procedures in §D:

- Dropping S5 alone flips **only SC-8** among the original four. That makes `no-state` the
  precise, minimal falsification for the revert criterion.
- Under the conjunct design, weakening the cherry guard alone no longer flips SC-6 or
  SC-7 — S5 still withholds. Their falsification therefore requires **both** mutations.
  A single-mutation procedure for those two would pass without falsifying anything.
- `folded-names` flips **exactly SC-10 and SC-11** and nothing else — notably **not** SC-12,
  which is what establishes that the rename axis and the encoding axis are independent. It
  is therefore the precise, minimal falsification for the rename half of AC-WSM-006, and it
  is a *different* mutation from `no-state` even though both flip those rows: `no-state`
  removes the conjunct, `folded-names` leaves the conjunct intact and starves it of the one
  path it needed to look at. A guard test that only exercised `no-state` would pass with the
  rename fold still present.
- `newline-split` flips **exactly SC-12** and nothing else — notably **not** SC-10, SC-11,
  or SC-13. It is therefore the precise, minimal falsification for the encoding axis of
  AC-WSM-006.
- `no-literal` flips **exactly SC-13** and nothing else — notably **not** SC-10, SC-11, or
  SC-12. It is the precise, minimal falsification for the interpretation axis. Its row is
  the starkest in the grid: the enumeration is untouched, so every path list is complete and
  byte-perfect in all thirteen scenarios, and one verdict still flips. That is the whole
  argument for why the interpretation axis is a separate guard rather than a consequence of
  the other two.
- `no-textconv` flips **exactly SC-14** and nothing else — notably **not** SC-15, which
  establishes that the two comparison-semantics members are independent rather than one
  effect with two names. It is the precise, minimal falsification for the textconv member.
  Its row is as stark as `no-literal`'s: every path list in all seventeen scenarios is
  complete, byte-perfect, and literally matched, and one verdict still flips — this time
  because git's *answer* about a correct list was wrong rather than the list itself.
- `no-ignoresub-cmp` and `no-ignoresub-enum` **both flip exactly SC-15, and only SC-15**.
  This is the one place in the grid where two mutations share a row, and the sharing is the
  finding rather than a defect in it: `--ignore-submodules=none` is a single guard applied at
  two stages, and dropping it from either stage alone reproduces the false positive. If the
  two occurrences were redundant, removing one would leave SC-15 at `keep`. They do not, so
  both are required. This is why AC-WSM-006's fifth falsification is stated as "remove it
  from either stage" rather than as two separate procedures.
- `oid-only-cmp` flips **exactly P3c and P4c** and nothing else. It is the precise, minimal
  falsification for AC-WSM-017, and it is the only mutation in the grid that does not remove
  a flag — it swaps the comparison mechanism. It is also the only mutation whose exposure
  needs no configuration at all: P3c and P4c differ solely in file mode under **default** git
  settings, so a mode-blind comparison misses them in any repository.
- **The five targeted repair mutations flip pairwise-disjoint row sets** — `{SC-10, SC-11}`,
  `{SC-12}`, `{SC-13}`, `{SC-14}`, `{SC-15}` — and the mechanism-swap mutation flips a sixth
  disjoint set, `{P3c, P4c}`. That disjointness is the evidence that `--no-renames`, `-z`,
  `--literal-pathspecs`, `--no-textconv`, `--ignore-submodules=none`, and mode sensitivity
  are six separate guards rather than six spellings of one: applying any five leaves the
  sixth's scenario failing. It is also why AC-WSM-006 requires all five of its falsifications
  to be run separately, and why AC-WSM-017 is a distinct criterion rather than a sixth
  falsification on AC-WSM-006 — its falsification is a mechanism substitution, not a flag
  removal, and its scenarios pass today.
- **`no-state` remains a forbidden falsification for both criteria**, and its blast radius is
  now wider than ever: it flips SC-8 and every one of SC-10 through SC-15, P3c, and P4c. A
  procedure using it would appear to work while exercising neither the enumeration, nor the
  comparison's diff semantics, nor its mode sensitivity — it would pass with the rename fold,
  the newline split, the bare pathspec, the textconv collapse, the submodule suppression, and
  a mode-blind comparison all still present.
- SC-2 reads MERGED under every mutation, which is the sanity check on the table: it
  confirms the mutations are narrowing the predicate rather than breaking it outright.

---

## §D Acceptance Criteria

### Detection correctness

**AC-WSM-001 — squash-merged branch is detected**
*Given* a repository in scenario SC-1,
*When* `IsBranchMerged("feat", "main")` is called,
*Then* it returns `(true, nil)`.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*Squash' -count=1 -v`
**Non-vacuity guard (required).** A `-run` selector matching nothing exits 0 with
`testing: warning: no tests to run`, which would make this criterion vacuously true.
Measured against the pre-implementation tree, where the selector matches nothing:
```
$ go test ./internal/core/git/ -run 'IsBranchMerged.*Squash' -count=1 -v
testing: warning: no tests to run
PASS
ok  github.com/modu-ai/moai-adk/internal/core/git  0.481s [no tests to run]
```
The criterion is judged by reading the `--- PASS:` lines, not the exit code. At least one
`--- PASS:` line naming a squash-scenario test MUST be present; zero `--- PASS:` lines is a
FAIL regardless of exit status. Record the observed count alongside the verdict.
Falsification: delete signal S4 from the predicate; this criterion must then FAIL.
Binds REQ-WSM-002.

**AC-WSM-002 — rebase-merged branch is detected**
*Given* a repository in scenario SC-2,
*When* `IsBranchMerged("feat", "main")` is called,
*Then* it returns `(true, nil)`.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*Rebase' -count=1 -v`
(non-vacuity discipline per §A. Baseline against the pre-implementation tree: `--- PASS:`
count 0, `[no tests to run]`.)
Falsification: delete signal S3; this criterion must then FAIL.
Binds REQ-WSM-003.

**AC-WSM-003 — empty-diff branch is detected**
*Given* a repository in scenario SC-5,
*When* `IsBranchMerged("feat", "main")` is called,
*Then* it returns `(true, nil)`.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*EmptyDiff' -count=1 -v`
(non-vacuity discipline per §A. Baseline against the pre-implementation tree: `--- PASS:`
count 0, `[no tests to run]`.)
Falsification: delete signal S2; this criterion must then FAIL. This is the only scenario
S2 uniquely covers, so its removal is otherwise silent.
Binds REQ-WSM-004.

**AC-WSM-004 — reachability cases are not regressed**
*Given* repositories in scenarios SC-3 and SC-4,
*When* `IsBranchMerged("feat", "main")` is called for each,
*Then* both return `(true, nil)`.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*Reachab' -count=1 -v`
(non-vacuity discipline per §A; at least two `--- PASS:` lines expected, one per scenario.
Baseline against the pre-implementation tree: `--- PASS:` count 0, `[no tests to run]`.)
Falsification: delete signal S1; SC-3 must then FAIL (SC-4 is also covered by S2, so S1
removal alone does not falsify it — SC-3 is the load-bearing case here).
Binds REQ-WSM-005.

**AC-WSM-005 — partially applied and fully unmerged branches are NOT reported merged**
*Given* repositories in scenarios SC-6 and SC-7,
*When* `IsBranchMerged("feat", "main")` is called for each,
*Then* both return `(false, nil)`.
A false positive on either deletes unmerged work.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*(Partial|Unmerged)' -count=1 -v`
(non-vacuity discipline per §A; at least two `--- PASS:` lines expected, one per scenario.
Baseline against the pre-implementation tree: `--- PASS:` count 0, `[no tests to run]`.)
Falsification (both mutations required, per §C.1): drop the S5 conjunct **and** relax S3/S4
to accept any non-empty `git cherry` output; this criterion must then FAIL for both
scenarios. Neither mutation alone falsifies it — verified in §C.1, where `weak-guard` and
`no-state` each leave SC-6 and SC-7 at `keep`. Applying only one mutation and observing a
still-passing test is not evidence the guard works. Note in particular that "treat empty
`git cherry` output as merged" does **not** falsify the SC-7 half, because SC-7's cherry
output is not empty — it is a single `+` line (§C raw output).
Binds REQ-WSM-006.

> SC-6 and SC-7 were separate criteria through v0.2.0. They are consolidated here because
> they bind the same requirement and share one falsification procedure verbatim — the same
> grounds on which AC-WSM-004 already covers SC-3 and SC-4 together. The freed slot is
> taken by AC-WSM-006 below, which covers a scenario class that had no criterion at all.

**AC-WSM-006 — a branch whose S5 conjunct cannot observe a real difference is NOT reported merged**
*Given* repositories in scenarios SC-10, SC-11, SC-12, SC-13, SC-14, and SC-15 — in each, a
path whose state the branch changed is absent from base HEAD, and in each the *only* thing
that can detect this is S5 actually observing that difference: its path list being complete,
faithfully encoded, and matched literally, **and its comparison not being narrowed by the
repository's diff configuration**,
*When* `IsBranchMerged("feat", "main")` is called for each,
*Then* all six return `(false, nil)`.
Run-phase judge:
`go test ./internal/core/git/ -run 'IsBranchMerged.*(Rename|NonASCII|ColonPath|Textconv|Submodule)' -count=1 -v`
(non-vacuity discipline per §A; at least six `--- PASS:` lines expected, one per scenario.)
**The selector was widened at v0.5.0 and again at v0.6.0.** The v0.4.0 form
`IsBranchMerged.*(Rename|NonASCII)` would not have matched a colon-scenario test, and the
v0.5.0 form `(Rename|NonASCII|ColonPath)` would not have matched a textconv or submodule
test — so adding scenarios required widening the selector, not only adding tests. Re-measured
against the pre-implementation tree (`f5129374b`) in its current form:
```
$ go test ./internal/core/git/ -run 'IsBranchMerged.*(Rename|NonASCII|ColonPath|Textconv|Submodule)' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/core/git	0.478s [no tests to run]
(--- PASS: line count: 0)
```
The widened selector is therefore non-vacuous at baseline exactly as the narrower ones were:
it exits 0 while producing zero `--- PASS:` lines, so it genuinely fails today. §A records
the regex check confirming the widening is real rather than cosmetic — the v0.5.0 form does
**not** match `TestIsBranchMerged_TextconvDriverCollapsesChange` or
`TestIsBranchMerged_SubmoduleIgnoreDirective`.

This is the criterion that pins S5's ability to see what it is asked about, on **all four**
of its axes. The four scenario groups fail through different mechanisms and no repair fixes
another:

- **SC-10 / SC-11 (completeness).** The branch renames a path that exists at the
  merge-base, base takes the branch's work, and base then re-adds the original path, so the
  branch's *deletion* of that path is absent from base HEAD. An implementation computing
  S5's list with a plain `git diff --name-only` folds the rename's deleted source path away
  (§C.2 route 1); S5 then never examines that path, holds vacuously, and the history signal
  is granted unopposed.
- **SC-12 (encoding).** The branch adds a non-ASCII path and base removes it after
  squash-merging, so the branch's content is absent from base HEAD. An implementation
  splitting the enumeration's output on newline receives that path C-quoted (§C.2 route 3);
  the quoted rendering matches no file, so S5 compares only the paths that happen to be
  ASCII, holds vacuously, and the history signal is again granted unopposed.
- **SC-13 (interpretation).** The branch adds a repo-root path whose first byte is `:` and
  base removes it after squash-merging. Here the path list is **already correct** — complete
  and byte-perfect — and the predicate still fails, because git parses a pathspec's leading
  characters as magic before matching (§C.2 route 4). An implementation that passes the list
  as a bare pathspec has S5 compare only the paths that happen not to begin with `:`; it
  holds vacuously and the history signal is granted unopposed a third time.
- **SC-14 / SC-15 (comparison semantics).** Here the path list is **not** at fault in the way
  the first three groups are, and the failure is one layer further down. In SC-14 the branch
  adds a file bound by a `.gitattributes` `diff=<driver>` rule to a `textconv` driver, and
  base overwrites it after squash-merging; the list is complete and byte-perfect, and
  `git diff --quiet` still reports equality because it compares the driver's *rendering* of
  each blob rather than the blobs (§C.2 route 5). In SC-15 the branch moves a submodule
  pointer with no `.gitmodules` change, and base moves it elsewhere after squash-merging;
  under a submodule-ignore directive the gitlink is suppressed at **both** stages — dropped
  from the enumeration so it never reaches the comparison, and ignored by the comparison even
  when the list is correct. An implementation issuing the comparison under the repository's
  own diff configuration has S5 hold vacuously and the history signal granted unopposed a
  fourth and fifth time.

SC-11 through SC-15 are the sharper cases: in all five, **S3 and S4 both fire**, so S5 is the
only thing preventing a wrong removal. SC-15 is sharper still, being the only scenario in the
matrix whose exposure a repository can ship to every clone with no operator configuration —
`submodule.<name>.ignore` is readable from a tracked `.gitmodules`.

Falsification — **five single mutations, all required**, per §C.1:

1. `folded-names` — compute the S5 path list with `git diff --name-only -z`, **retaining
   `-z`** and changing nothing else, so that only rename detection differs. SC-10 and SC-11
   must then report merged and this criterion must FAIL.
2. `newline-split` — keep `--no-renames` but drop `-z` and split the output on `\n` instead
   of `\x00`, changing nothing else. SC-12 must then report merged and this criterion must
   FAIL.
3. `no-literal` — leave the enumeration exactly as specified and drop `--literal-pathspecs`
   from the S5 comparison, changing nothing else. SC-13 must then report merged and this
   criterion must FAIL.
4. `no-textconv` — leave the enumeration exactly as specified and drop `--no-textconv` from
   the S5 comparison, changing nothing else. SC-14 must then report merged and this criterion
   must FAIL.
5. `no-ignoresub` — drop `--ignore-submodules=none` from **either** the comparison **or** the
   enumeration, changing nothing else. SC-15 must then report merged and this criterion must
   FAIL. Measured, **each stage alone is sufficient** to reproduce the false positive
   (§C.1 rows `no-ignoresub-cmp` and `no-ignoresub-enum`), which is the evidence both
   occurrences are load-bearing rather than one being a copy of the other. Running the
   comparison-side removal alone is an acceptable discharge of this falsification, but the
   enumeration-side removal is the more informative of the two, because it is the occurrence
   a future edit is likelier to delete as duplication.

> The `-z`-retaining clause in falsification 1 is load-bearing and was added at v0.5.0. The
> earlier wording ("compute the S5 path list with `git diff --name-only` instead of
> `git diff --no-renames --name-only -z`") drops **both** flags, which is a different
> mutation: measured, renames-on-with-`-z`-retained flips SC-10 and SC-11 only, while
> renames-on-with-`-z`-dropped flips SC-10, SC-11 **and** SC-12. An implementer following the
> earlier wording literally would apply the second and destroy the very disjointness this
> criterion's rationale rests on. The criterion's discriminating power was never affected —
> both forms make it FAIL — so this is a precision fix, not a correctness one.

§C.1 confirms each mutation flips **exactly** its own rows and nothing else: `folded-names`
flips SC-10 and SC-11; `newline-split` flips SC-12; `no-literal` flips SC-13; `no-textconv`
flips SC-14; and either submodule mutation flips SC-15. Each leaves every other scenario in
the criterion at `keep`, and all five leave SC-2, SC-6, SC-7, and SC-8 unchanged. That
disjointness is the reason all five mutations are required: applying only one would leave
three axes unexercised, and the criterion would pass with a live defect present.

The single deliberate exception is that the two submodule mutations flip the **same** row.
That is not a disjointness failure — it is the evidence the flag is needed at two stages
rather than one. If the occurrences were redundant, removing one would leave SC-15 passing.

None of the five falsifications may be `no-state`. Dropping the S5 conjunct entirely flips
SC-10 through SC-15 (§C.1), so a procedure that used it here would appear to work while never
exercising the enumeration or the comparison at all — it would pass with the rename fold, the
newline split, the bare pathspec, the textconv collapse, and the submodule suppression all
still present. That is the same failure shape as the iteration-1 defect in which a stated
falsification exercised the wrong half of a guard.

> This criterion absorbed SC-12 at v0.4.0, SC-13 at v0.5.0, and SC-14 and SC-15 at v0.6.0,
> rather than new criteria being added for each: all bind the same requirement
> (REQ-WSM-014) through the same conjunct as SC-10 and SC-11, and the criterion's scope has
> widened correspondingly — from "rename fold" to "path-list integrity" to "path-list
> round-trip integrity" and now to **"S5 actually observes the difference"**, which is what
> all six scenarios have in common. SC-15 needed no widening at all: a suppressed gitlink is
> an under-inclusive path list, the criterion's original scope. SC-14 widened it by one
> clause, because there the list is perfect and only the comparison is wrong.
>
> The cost is that one criterion now carries five falsifications, and §F's Definition of Done
> requires each to be run separately for exactly that reason. P3c and P4c were **not** folded
> in as a sixth and seventh scenario, even though the AC budget made that tempting: their
> claim is different (the comparison must be mode-sensitive), their falsification is a
> mechanism substitution rather than a flag removal, and they pass under the current design
> rather than motivating a repair. They are AC-WSM-017. `plan.md` §B records why that
> decision takes the SPEC one criterion over the Tier M ceiling and why the alternatives
> were judged worse.
Binds REQ-WSM-006, REQ-WSM-014.

**AC-WSM-007 — a probe failure surfaces as an error; a verdict-carrying exit does not**
*Given* the four cases below,
*When* `IsBranchMerged` is called for each,
*Then* each yields the stated outcome:

| Case | Call | Expected |
|---|---|---|
| unknown base ref | `IsBranchMerged("feat", "no-such-base")` | non-nil error; the boolean is not used as a verdict |
| unrelated histories (D7) | `IsBranchMerged("orphan", "main")` in a repo whose `orphan` branch shares no ancestor with `main` | `(false, nil)` — kept, no error |
| branch with a non-empty diff vs merge-base | any of SC-1, SC-2, SC-6, SC-7 | nil error (the `git diff --quiet` rc=1 is a verdict, not a failure) |
| branch with an empty diff vs merge-base | SC-5 | `(true, nil)` |

Rows 3 and 4 are what make this criterion bind REQ-WSM-007's narrowing rather than its
earlier over-broad form: a predicate that errors on `git diff --quiet` rc=1 fails row 3,
and one that errors on `git merge-base` rc=1 fails row 2.
Authoring-time evidence for the unrelated-histories row (executed):
```
git merge-base main orphanb   ->  rc=1, stdout=''
```
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*ProbeExit' -count=1 -v`
(non-vacuity discipline per §A; at least four `--- PASS:` lines expected, one per row of the
table above. Baseline against the pre-implementation tree: `--- PASS:` count 0,
`[no tests to run]`.)
Falsification: swallow the unknown-base probe error and return `false, nil` — row 1 must
then FAIL; treat any non-zero exit as a failure — rows 2 and 3 must then FAIL.
Binds REQ-WSM-007.

### Synthetic-object hygiene

**AC-WSM-008 — the synthetic commit is deterministic**
*Given* a repository in scenario SC-1 whose `feat` branch is unmodified between calls,
*When* the predicate's synthetic-commit construction runs twice,
*Then* both runs produce the same object hash and the dangling-commit count increases by
at most one across both.
Authoring-time evidence (executed, with identity pinned). The inputs are recorded so the
value is independently reproducible rather than asserted — the fixture is a two-commit
repository whose commit dates are pinned to `@1700000001` and `@1700000002`:
```
merge-base (parent) = 652f98d4a2806a26a2ad47397c40ed492beec65c
feat^{tree}         = a822cc2d8c3bf904ac9faf19c2a90fc5fa119727
message             = moai-squash-probe
identity            = moai <moai@localhost>, GIT_AUTHOR_DATE=GIT_COMMITTER_DATE=@0 +0000

run1=d4b2e2e11f29a6bdad158f627aad5509ddcaca8f
run2=d4b2e2e11f29a6bdad158f627aad5509ddcaca8f
dangling commits after both runs: 1
```
The specific hash is a property of those four inputs plus git's commit-object format; an
implementation choosing a different fixed message or identity will produce a different but
equally stable value. What the criterion asserts is the **determinism**, not the literal
digest — a run-phase implementation is judged on `run1 == run2` and on the dangling count
increasing by at most one, not on matching `d4b2e2e1…`.
Run-phase judge: `go test ./internal/core/git/ -run 'SyntheticCommit.*Determinis' -count=1 -v`
(non-vacuity discipline per §A. Baseline against the pre-implementation tree: `--- PASS:`
count 0, `[no tests to run]`.)
Falsification: unpin `GIT_COMMITTER_DATE`; the two hashes must then differ and this
criterion must FAIL.
Binds REQ-WSM-008.

**AC-WSM-009 — no repository-wide object reclamation**
*Given* the post-change tree,
*When* the following runs,
*Then* it prints `0`:
```bash
grep -nE '"(prune|gc)"' internal/core/git/worktree.go | grep -v '"worktree", "prune"' | wc -l
```
Two repairs versus the previous iteration, both verified by injection:
- `-r` is dropped. With `-r`, grep prefixes each line with the path
  `internal/core/git/worktree.go:`, which itself contains `worktree` — so the old exclusion
  pattern `worktree.*prune` matched *every* line in the file containing `prune`, swallowing
  the very calls the guard exists to catch.
- The exclusion anchors on the literal argument pair `"worktree", "prune"` (the legitimate
  call at worktree.go:132) rather than a loose `.*`.

Falsification (executed against the baseline tree, then reverted): append
`func probeInjected() { _, _ = execGit(nil, "", "prune") }` to `worktree.go`, then run both
judges.
```
=== injected bare git prune ===
--- BROKEN judge (previous iteration) ---   0     <- misses it
--- REPAIRED judge ---                      1     <- catches it
--- repaired judge, surviving line ---
295:func probeInjected() { _, _ = execGit(nil, "", "prune") }
=== restored ===
git status --porcelain internal/core/git/worktree.go   -> (empty)
repaired judge on restored tree                        -> 0
```
The falsification uses `git prune`, not `git gc`. `git prune` is the primary prohibition in
REQ-WSM-009 and the one §5 decision 2 argues at length is hazardous; the previous
iteration's `git gc` falsification passed while never exercising the broken half of the
filter.
Note this is a presence check, not a reachability check; AC-WSM-011's full-suite run is
what establishes the predicate itself is reached.
Binds REQ-WSM-009.

### Preserved safety contract

**AC-WSM-010 — the `--stale` two-condition gate is intact**
*Given* a worktree whose branch is merged but whose working tree has uncommitted or
untracked changes,
*When* the `--stale` sweep evaluates it,
*Then* it is kept and reported, and never removed.
Run-phase judge: the pre-existing `TestCleanStale_KeepsDirtyWorktree` must pass unchanged.
```bash
go test ./internal/cli/worktree/ -run 'TestCleanStale_KeepsDirtyWorktree' -count=1 -v
```
**Non-vacuity guard (required).** Read the `--- PASS:` lines, not the exit code: a `-run`
selector matching nothing exits 0 and would make this criterion vacuously true. Exactly one
`--- PASS:` line is expected. Observed at baseline (the test exists at
`internal/cli/worktree/clean_stale_test.go:79`):
```
--- PASS: TestCleanStale_KeepsDirtyWorktree (0.00s)
ok  github.com/modu-ai/moai-adk/internal/cli/worktree  0.372s
(--- PASS: line count: 1)
```
A count of 0 means the test was renamed or removed rather than passing, which is itself the
regression this criterion exists to catch.
Falsification: the test's `removeFunc` already fails the test on `force=true`; deleting the
cleanliness check in `staleKeepReason` must make this criterion FAIL.
Binds REQ-WSM-010.

**AC-WSM-011 — the interface signature did not change**
*Given* the post-change tree,
*When* the following runs,
*Then* both commands exit 0 with no output:
```bash
go build ./... && go vet ./...
git diff --exit-code -- internal/core/git/types.go
```
The second command is the load-bearing one: an unchanged `types.go` is the evidence that
no test double or call site needed editing, which is the premise of the shared-predicate
decision (`spec.md` §5 decision 3).
Falsification: widen the signature; the `git diff --exit-code` must then exit non-zero.
Binds the §C constraint in `plan.md`.

**AC-WSM-012 — both removal call sites stay non-forced**
*Given* the post-change tree,
*When* the following runs,
*Then* it prints `2`:
```bash
grep -cE 'Remove\([A-Za-z_.]+, false\)' internal/cli/worktree/clean.go
```
Observed at baseline (executed):
```
89:			if err := WorktreeProvider.Remove(wt.Path, false); err != nil
194:		if err := WorktreeProvider.Remove(c.path, false); err != nil
count: 2      (the previous iteration's 'Remove(wt.Path, false)' literal counted only 1)
```
The pattern matches on the `false` argument independent of the receiver spelling, so a
rename of the loop variable does not break the criterion, and it pins **both** non-forced
sites — clean.go:89 (`--merged-only`) and clean.go:194 (`--stale`) — rather than only the
first.
Authoring-time evidence that non-forced removal is a real safety net (executed; see also
AC-WSM-016 for the state it does *not* refuse):
```
[modified]        rc=128 present=YES  fatal: '...' contains modified or untracked files, use --force to delete it
[untracked-only]  rc=128 present=YES  fatal: '...' contains modified or untracked files, use --force to delete it
```
Falsification: change either argument to `true`; the count must then drop to 1.
Binds REQ-WSM-013.

### Regression

**AC-WSM-013 — the pre-existing predicate tests still pass**
*Given* the post-change tree,
*When* `go test ./internal/core/git/ -run TestWorktreeIsBranchMerged -count=1 -v` runs,
*Then* it reports PASS, and the number of test cases run is non-zero.
The non-zero assertion is required: `go test -run` with a selector matching nothing exits 0,
which would make this criterion vacuously true. Confirm by reading the `--- PASS:` lines,
not the exit code alone. Three tests currently match this selector
(`TestWorktreeIsBranchMerged`, `TestWorktreeIsBranchMerged_NotMerged`,
`TestWorktreeIsBranchMerged_BranchCheckedOutInWorktree`), so a count below three means
tests were lost rather than passing.
Binds REQ-WSM-005.

**AC-WSM-014 — full suite green**
*Given* the post-change tree,
*When* `go test ./... -count=1` runs,
*Then* it exits 0.
This is a backstop, not a substitute for a per-requirement criterion: a green suite does not
demonstrate any specific requirement holds unless a test asserts it.
Binds all requirements.

### State semantic

**AC-WSM-015 — a partially reverted branch is NOT reported merged**
*Given* a repository in scenario SC-8 — `main` cherry-picked every `feat` commit and then
reverted one of them, so both plain-`git cherry` lines are prefixed `-` while `b.txt` is
absent from `main`'s tree,
*When* `IsBranchMerged("feat", "main")` is called,
*Then* it returns `(false, nil)`.
This criterion is what pins the **state** semantic of REQ-WSM-001 against the history
semantic the patch-id probes implement on their own. Without it, REQ-WSM-001's wording
("already present in the base branch") admits an implementation that only ever asks whether
a patch-id once existed in base's history.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*Revert' -count=1 -v`
**Non-vacuity guard (required).** This is the criterion that pins the entire state
semantic, so a vacuous pass here is the most costly of any in this document. The selector
matches nothing against the pre-implementation tree and exits 0:
```
$ go test ./internal/core/git/ -run 'IsBranchMerged.*Revert' -count=1 -v
testing: warning: no tests to run
PASS
ok  github.com/modu-ai/moai-adk/internal/core/git  0.204s [no tests to run]
```
The criterion is judged by reading the `--- PASS:` lines, not the exit code. At least one
`--- PASS:` line naming a revert-scenario test MUST be present; zero `--- PASS:` lines is a
FAIL regardless of exit status. Record the observed count alongside the verdict.
Authoring-time evidence (executed):
```
main tree      : a.txt seed.txt
feat tree      : a.txt b.txt seed.txt
plain cherry   : - 616bf026a585f183d8c49ec58de025fb68fc4d15 | - fb4444864adbdc8bed6c4f61bd35d172551c4054
names(mb..feat): a.txt b.txt
state probe    : git diff --quiet feat main -- a.txt b.txt  ->  rc=1
```
Falsification (single mutation, per §C.1): drop the S5 conjunct from S3 and S4. SC-8 must
then report merged and this criterion must FAIL. §C.1 confirms this `no-state` mutation is
the **only** one of the five that flips SC-8: `weak-guard` and `folded-names` both leave it
at `keep`. So a passing AC-WSM-005 alongside a failing AC-WSM-015 is the expected signature
of exactly this defect.
Note that `no-state` also flips SC-10 and SC-11 (§C.1), so AC-WSM-006 fails under it too.
That is expected — removing the conjunct defeats every scenario the conjunct protects — and
it is why AC-WSM-006 uses `folded-names` rather than `no-state` as *its* falsification:
only `folded-names` isolates the enumeration from the conjunct.
Binds REQ-WSM-001, REQ-WSM-014.

### `--stale` sweep constraints

**AC-WSM-016 — the sweep deletes no branch and performs nothing without `--yes`**
*Given* the post-change tree,
*When* the two checks below run,
*Then* both hold:

1. No branch-deletion call exists on the sweep path:
```bash
grep -cE 'DeleteBranch|"branch", "-[dD]"' internal/cli/worktree/clean.go
```
   must print `0`. Observed at baseline: `0`.
2. The preview/apply split is asserted by tests that actually run:
```bash
go test ./internal/cli/worktree/ -run 'TestCleanStale_PreviewsByDefault|TestCleanStale_RemovesWithYes' -count=1 -v
```
   must report PASS for **both** named tests. Read the `--- PASS:` lines, not the exit code:
   a `-run` selector matching nothing exits 0 and would make this vacuous. Observed at
   baseline:
```
--- PASS: TestCleanStale_PreviewsByDefault (0.00s)
--- PASS: TestCleanStale_RemovesWithYes (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	3.487s
```
Falsification: add a branch-deletion call to `clean.go` — check 1 must then exceed `0`;
make the sweep remove without `--yes` — `TestCleanStale_PreviewsByDefault` must then FAIL.
Known gap this criterion does **not** cover: a clean worktree whose branch carries unpushed
commits is removed without objection (`rc=0`, directory gone), and `git status --porcelain`
reports it clean. That case is accepted with the branch-retention mitigation and recorded in
`spec.md` §5 decision 3 and §4 Out of Scope — unpushed-commit detection. Check 1 above is
what keeps the mitigation true.
Binds REQ-WSM-011, REQ-WSM-012.

### Comparison mode-sensitivity

**AC-WSM-017 — a branch whose only divergence is a file mode is NOT reported merged**
*Given* repositories in scenarios P3c and P4c — in P3c the branch chmods a file to 755 and
base chmods it back; in P4c the branch replaces a symlink with a regular file whose content
is exactly the symlink's target, and base restores the symlink. In both, the two sides share
**one blob OID** and differ only in mode,
*When* `IsBranchMerged("feat", "main")` is called for each,
*Then* both return `(false, nil)`.
Run-phase judge: `go test ./internal/core/git/ -run 'IsBranchMerged.*(ModeOnly|SymlinkFile)' -count=1 -v`
(non-vacuity discipline per §A; at least two `--- PASS:` lines expected, one per scenario.)
Measured against the pre-implementation tree (`f5129374b`):
```
$ go test ./internal/core/git/ -run 'IsBranchMerged.*(ModeOnly|SymlinkFile)' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/core/git	0.313s [no tests to run]
(--- PASS: line count: 0)
```

**This criterion pins a property the specified mechanism already has.** `git diff` compares
file modes, so both scenarios pass as designed — no flag is added for them and no repair is
required. It exists because the property is invisible in the mechanism's spelling and is
exactly what the most plausible future "improvement" would discard. Authoring-time ground
truth (executed):
```
P3c   feat 100755 blob 587be6b4c3f93f93c489c0111bba5596147a26cb  x.txt
      main 100644 blob 587be6b4c3f93f93c489c0111bba5596147a26cb  x.txt
P4c   feat 100644 blob 1de565933b05f74c75ff9a6520af5f9f8a5a2f1d  link.txt
      main 120000 blob 1de565933b05f74c75ff9a6520af5f9f8a5a2f1d  link.txt

git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none …  rc=1  detects
blob-OID-only comparison over the same paths                                   EQUAL misses
```

Three properties make this worth a criterion of its own rather than a note:

- **The exposure it guards against needs no configuration.** Every member of the
  comparison-semantics axis requires non-default state — a committed `.gitattributes`, a
  committed `.gitmodules` directive, or an operator config key. A mode-blind comparison
  misses P3c and P4c in **any** repository under **default** git settings. It would be the
  worst exposure in the SPEC.
- **The swap it guards against is attractive, not careless.** Comparing blob object IDs is
  the natural answer to the comparison-semantics axis — content hashes cannot be narrowed by
  diff configuration — and it was formally proposed and studied. `spec.md` §2 records why it
  was rejected; this criterion is what makes the rejection enforceable rather than advisory.
- **Two of the three candidate mechanisms cannot satisfy it even in principle.**
  `git rev-parse <rev>:<path>` returns an OID and nothing else, and
  `git cat-file --batch-check` rejects the `%(objectmode)` atom outright, so neither can
  report a mode. `cat-file` additionally reports `missing` for a gitlink on both sides at
  exit 0, making two different submodule pointers compare equal.

Falsification (single mutation, per §C.1): replace the S5 comparison with a mode-blind
blob-OID comparison — two `git ls-tree -r -z --full-tree` maps keyed by path, storing only
the OID and discarding the mode, compared over the enumerated names. P3c and P4c must then
report merged and this criterion must FAIL. §C.1's `oid-only-cmp` row confirms this mutation
flips **exactly** P3c and P4c and leaves all fifteen other scenarios unchanged.

This falsification may **not** be `no-state`. Dropping the S5 conjunct also flips P3c and P4c
(§C.1), so a procedure using it would pass with a mode-blind comparison still present — the
same substitution error §D forbids for AC-WSM-006, one layer further down.

> This criterion takes the SPEC to 17 acceptance criteria against a Tier M ceiling of 16.
> The overage is declared rather than absorbed: folding P3c and P4c into AC-WSM-006 would
> leave that criterion carrying eight scenarios and six falsifications, and merging two
> unrelated criteria elsewhere to free a slot would break the commitment recorded at v0.4.0
> that no consolidation would be manufactured to create one. `plan.md` §B carries the full
> reasoning and the recommendation to tier up to L.
Binds REQ-WSM-006, REQ-WSM-014.

---

## §E Quality Gates

| Gate | Threshold |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0, no findings |
| `golangci-lint run` | no new findings versus the pre-change baseline |
| `go test ./... -count=1` | exit 0 |
| `internal/core/git` coverage | not below the pre-change measured baseline |

The coverage gate is expressed as a delta against a baseline measured at run-phase start,
not against a remembered number. Measure it, record the command and its output, then
compare.

---

## §F Definition of Done

- [ ] AC-WSM-001 through AC-WSM-017 each PASS, with the judging command and its observed
      output recorded. Every criterion names one: AC-WSM-001 through AC-WSM-008 plus
      AC-WSM-010, AC-WSM-013, AC-WSM-015, AC-WSM-016 check 2 and AC-WSM-017 are judged by a
      `go test -run` selector (thirteen distinct selectors, enumerated with their baseline
      counts in §A); AC-WSM-009, AC-WSM-011, AC-WSM-012 and AC-WSM-016 check 1 are judged
      by a shell command; AC-WSM-014 is the full-suite backstop.
- [ ] For every criterion judged by a `go test -run` selector, the observed `--- PASS:`
      line count is recorded and is non-zero (§A non-vacuity discipline). An exit code
      alone is not an accepted verdict for any of them.
- [ ] Each falsification procedure was performed for the guard criteria (AC-WSM-005,
      AC-WSM-006, AC-WSM-008, AC-WSM-009, AC-WSM-015, AC-WSM-017), and each made its
      criterion FAIL as stated. For AC-WSM-005 this means applying **both** mutations named
      in §C.1 — a single mutation leaves it passing and proves nothing. For AC-WSM-006 it
      means all **five** of `folded-names` (with `-z` retained), `newline-split`,
      `no-literal`, `no-textconv`, and `no-ignoresub` (from either stage), each run
      separately, and **not** `no-state`: the mutations flip disjoint row sets, so any one
      alone leaves three axes unexercised. For AC-WSM-017 it means `oid-only-cmp`, and
      **not** `no-state`.
- [ ] The seventeen-scenario oracle in §C is reproduced by the test suite, one case per row,
      including SC-8 (partially reverted), SC-9 (superset), SC-10 / SC-11 (rename +
      re-add), SC-12 (non-ASCII path removed from base), SC-13 (leading-colon path removed
      from base), SC-14 (textconv driver), SC-15 (submodule pointer under an ignore
      directive), P3c (mode-only divergence), and P4c (symlink/file blob-OID collision). Per
      the §B construction notes: the SC-12 fixture sets `core.quotePath true` explicitly; the
      SC-13 fixture stages and removes its hazard path under `git --literal-pathspecs`; the
      SC-14 fixture keeps `.gitattributes` in the **worktree** as well as committed, with an
      executable textconv script at an absolute path; the SC-15 fixture moves the submodule
      pointer **without touching `.gitmodules`** and places the ignore directive inside the
      tracked `.gitmodules`; and the P4c fixture writes the symlink target with **no trailing
      newline**, so both sides genuinely share one blob OID.
- [ ] The S5 path list is computed with `--ignore-submodules=none --no-renames --name-only -z`,
      split on `\x00` (never on `\n`), and passed to the exec helper as a `[]string`, never
      as a joined shell string (§C.2 routes 1-3, 5). `core.quotePath=false` is not accepted
      as a substitute for `-z`, and the enumeration-side `--ignore-submodules=none` is not
      accepted as redundant with the comparison-side one.
- [ ] The S5 comparison is issued as
      `git --literal-pathspecs diff --quiet --no-textconv --ignore-submodules=none …`, with
      `--literal-pathspecs` **before** the subcommand and the other two **after** it (§C.2
      routes 4-5). Every misplacement exits 129 rather than producing a verdict, so a
      position error is loud. `GIT_LITERAL_PATHSPECS=1` via the per-call environment hook is
      an accepted equivalent for the first; a per-element `:(literal)` prefix is **not**
      accepted, because it re-introduces the per-path string transformation that has failed
      on three of the four round-trip axes.
- [ ] The S5 comparison remains **mode-sensitive**. `git diff` satisfies this; a blob-OID
      comparison built on `git rev-parse <rev>:<path>`, `git cat-file --batch-check`, or an
      `ls-tree` map storing only the OID does **not**, and turns P3c and P4c into silent
      false positives under default git configuration (§C.2 route 6). If a future change ever
      requires an OID-based comparison, it must carry the file **mode** alongside the OID.
- [ ] Every git invocation in the predicate runs with its **working directory at the
      repository root** — `execGit`'s `dir` argument is `w.root`, i.e. `rev-parse
      --show-toplevel`. The S5 enumeration emits root-relative paths while the comparison's
      pathspecs resolve against the process cwd, so a nested cwd makes every path an
      unmatched pathspec and S5 fails open (finding 5). `--literal-pathspecs` disables the
      `:(top)` remedy, so this must hold at the invocation level; the argument vector alone
      does not carry it.
- [ ] No git invocation on the predicate's path is routed through a shell. Verified by the
      `plan.md` §D grep, which returns no match; a shell would re-open the argv axis and
      fail silently in the merged direction.
- [ ] `types.go` is unmodified.
- [ ] No `git prune` or `git gc` invocation was introduced.
- [ ] The `@MX:ANCHOR` recording why reachability is retained alongside the patch-id probes
      is present on the predicate.
