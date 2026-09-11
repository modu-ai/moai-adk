# Gate round 1 — superseded

Card: t629 · SPEC: SPEC-REVIEW-SECRET-SCAN-REFS-001

This file holds gate round 1 of the measure-first gate, moved verbatim out of `progress.md` §E.2 on
the `plan.md` M3 re-gate path. Round 1 pinned the command-substitution form of the scan command (pin
commit `b8c0ef74b`) and recorded gate cells 1 and 2 on it (evidence commit `6e56840d5`). SPEC 0.3.0
(`9dab82a4a`) replaced that form with a caret-prefixed tip store read through standard input, so this
round is superseded and its verdicts do not stand for the document edit. Cell 3 was never taken in
this round, and no lead approval for it was recorded.

Everything below the rule is the moved text, byte for byte, as it stood in `progress.md` at
`2b1808efe`.

---

### Pinned procedure

Pinned on branch `WT-secret-scan-refs` on top of `f16f7c095`, in a commit of its own that precedes the
first gate command. The regex below is written exactly as `review.md` writes it. Paths are relative to
the project root.

Tip recording: before the scan starts, record the tip of every ref with `git for-each-ref --format='%(objectname)' > .moai/state/secrets-scan-tips.next`, and replace `.moai/state/secrets-scan-tips.txt` with that file only after the scan that carries the final result exits 0.
Full-history scan: git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})'
Scan command: git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' --not $(cat .moai/state/secrets-scan-tips.txt)
Missing-tip handling: when the scan exits non-zero and its error output reports `bad object` for a tip listed in `.moai/state/secrets-scan-tips.txt`, report that tip as missing and run the full-history scan in its place; any other non-zero exit is a scan failure that is reported and leaves `.moai/state/secrets-scan-tips.txt` unchanged.
Uncovered commits: Commits that no ref and no HEAD reaches, such as commits reachable only through a reflog, are outside every scan step in this procedure, and no step scans them.

Which step runs: the full-history scan runs when `.moai/state/secrets-scan-tips.txt` does not exist,
and the scan command runs otherwise. Because the tips are recorded before the scan starts, a commit
that lands while the scan runs is outside the recorded tips and falls to the next review instead of
being skipped; a commit that lands between the recording and the scan's own read of the refs is
scanned twice. `Uncovered commits:` carries the REQ-004 sentence that AC-006 checks, fixed here before
any document edit.

Gate execution method, fixed before the gate. This session's worktree guard refuses a `git` command
that contains a command substitution, so the gate runs the scan command with
`$(cat .moai/state/secrets-scan-tips.txt)` replaced by the store's lines typed as separate arguments in
store order, read with `cat` in the command immediately before, and checks the argument count against
`wc -l` of the store. For a store of full-length object names, one per line, shell word splitting of
the substitution yields exactly those arguments. Not exercised by this method: the substitution itself,
an empty store, and argument-list limits. On the fixture the tip store sits at `.moai/state/` under the
fixture's own root, the pinned path.

Design probes taken before this pin, on a separate throwaway repository in the session scratchpad with
no markers and no `-G`; they are syntax checks, not gate evidence. `git --version` →
`git version 2.50.1 (Apple Git-155)`.

- `git for-each-ref --format='^%(objectname)'` to a file → exit 0, one caret-prefixed line.
- `git log -p --all --stdin` reading a caret-prefixed line for an object that does not exist → exit
  128, stdout 0 bytes, stderr `fatal: bad object <name>` naming the object without the caret. A
  caret-prefixed store would therefore not hold the tip the way the error names it, so the store holds
  plain object names.
- The same form over a store of existing tips → exit 0.
- `man git-log`: a `--not` given on the command line does not affect revisions read through `--stdin`.
  A `--not` inside the store is honoured by this git build; it was not adopted, because acceptance
  AC-006 needs `--not` on the `git log -p` line itself.
- The worktree guard refused a `printf` carrying `--not`, and refused
  `git log -p --all --not $(cat <store>)` as a form too complex to verify; hence the execution method
  above.

### Gate evidence

Taken 2026-09-11 from this worktree on branch `WT-secret-scan-refs`, HEAD `b8c0ef74b` (the pin
commit), on a throwaway fixture repository outside this repository at `SP/m1gate/fx`, where `SP` is the
session scratchpad. `git version 2.50.1 (Apple Git-155)`. Fixture set-up: `git init -b main`, a fixture
`user.name` and `user.email`, `commit.gpgsign=false`, and `core.hooksPath=/dev/null`, read back with
`git config --list --local`. `REGEX` is the scan regex as `review.md` writes it.

- Markers: PEM-style private-key header lines with distinct uppercase labels, each written by `printf`
  from a format string split around two fragments and not reproduced here. Each marker file was checked
  with `/usr/bin/grep -cE -- 'REGEX' <file>` → `1`.
- Every command's output went to a file under `SP/m1gate/`, and every exit code was read with
  `; echo "exit=$?"`, never through a pipe.
- Tip store: `SP/m1gate/fx/.moai/state/secrets-scan-tips.txt`, recorded through `.next` and replaced
  with `mv` only after the scan carrying the final result exited 0.
- Execution method as pinned: each run of the scan command replaced
  `$(cat .moai/state/secrets-scan-tips.txt)` with the store's lines, read with `cat` and `wc -l` in the
  command before. Arguments against store lines: R2 1/1, R3 2/2, R4 3/3.
- Fixture object names are kept in `SP` only. `<G1>` below stands for the recorded tip of branch
  `gone`, a full 40-character object name.

#### Gate cell 1

| Step | Command (in `SP/m1gate/fx`; outputs in `SP/m1gate/`) | Exit | Reading |
|---|---|---|---|
| C0 | clean `keys.txt` committed on `main` | 0 | — |
| clean-history scan | `git log -p --all -G 'REGEX' > g1-base.txt 2> g1-base.err` | 0 | `wc -c` 0 and 0 bytes |
| R1 tip recording | `git for-each-ref --format='%(objectname)' > .moai/state/secrets-scan-tips.next` | 0 | 1 line |
| R1 scan (no store yet, so the full-history scan) | `git log -p --all -G 'REGEX' > g1-r1.txt 2> g1-r1.err` | 0 | 0 and 0 bytes; `mv` to the store, exit 0 |
| S1 | `SIDECELL` marker in `side.txt`, committed on `side` created from C0 | 0 | `show --stat side`: 1 file, 1 insertion, parent C0 |
| C1 | `HEADCELL` marker appended to `keys.txt`, committed on `main` | 0 | `side.txt` absent on `main` (`test -e` exit 1) |
| construction | `git merge-base --is-ancestor side HEAD` | 1 | — |
| R2 tip recording | as R1 | 0 | — |
| R2 scan command | `git log -p --all -G 'REGEX' --not <1 store line> > g1-r2.txt 2> g1-r2.err` | 0 | 634 and 0 bytes; `mv` to the store (2 lines), exit 0 |
| construction, re-read before the counts | `git merge-base --is-ancestor side HEAD` | 1 | — |
| counts | `/usr/bin/grep -c 'HEADCELL' g1-r2.txt`; `/usr/bin/grep -c 'SIDECELL' g1-r2.txt` | 0; 0 | `1`; `1` |
| context | `/usr/bin/grep -cE -- 'REGEX' g1-r2.txt`; `/usr/bin/grep -c '^commit ' g1-r2.txt` | 0; 0 | `2`; `2` |

Predicate (`spec.md` §3.4 cell ①, AC-008): the clean-history scan printed 0 bytes — holds; every scan
exited 0 (clean-history scan, R1, R2) — holds; `is-ancestor` exited 1 immediately before the counts —
holds; in R2, `HEADCELL` ≥ 1 (`1`) and `SIDECELL` ≥ 1 (`1`) — both hold. The side branch's marker was
reported in the scan where it first became reachable, next to the HEAD-line control.

verdict: trustworthy

#### Gate cell 2

Taken on the same fixture after cell 1.

| Step | Command (in `SP/m1gate/fx`; outputs in `SP/m1gate/`) | Exit | Reading |
|---|---|---|---|
| G1 | plain `gone.txt` with a line unique to G1, committed on `gone` created from `main` | 0 | `for-each-ref --contains gone` → `refs/heads/gone` only |
| R3 tip recording | `git for-each-ref --format='%(objectname)' > .moai/state/secrets-scan-tips.next` | 0 | 3 lines; `/usr/bin/grep -cxF '<G1>'` → `1` |
| R3 scan command | `git log -p --all -G 'REGEX' --not <2 store lines> > g2-r3.txt 2> g2-r3.err` | 0 | 0 and 0 bytes; `mv` to the store (3 lines, `<G1>` counted `1`), exit 0 |
| delete | `git branch -D gone` | 0 | — |
| expire | `git reflog expire --expire=now --all` | 0 | — |
| prune | `git gc --prune=now --quiet` | 0 | — |
| construction | `git cat-file -e <G1>` | 1 | the recorded tip is absent from the object store |
| L1 | `GONECELL` marker in `after.txt`, committed on `after` created from `main` | 0 | — |
| construction | `git merge-base --is-ancestor after HEAD` | 1 | — |
| R4 tip recording | as R3 | 0 | — |
| R4 scan command | `git log -p --all -G 'REGEX' --not <3 store lines> > g2.txt 2> g2.err` | 128 | `g2.txt` 0 bytes; `g2.err` 59 bytes, one line of the form `fatal: bad object <G1>` |
| handling check | `/usr/bin/grep -cxF '<G1>' .moai/state/secrets-scan-tips.txt` | 0 | `1`: the object named is a listed tip, so the handling runs the full-history scan |
| full-history scan in its place | `git log -p --all -G 'REGEX' > g2-fallback.txt 2> g2-fallback.err` | 0 | 945 and 0 bytes; `mv` to the store, exit 0 |
| detection reading | `/usr/bin/grep -cF '<G1>' g2.err g2.txt g2-fallback.txt` | 0 | `g2.err:1`, `g2.txt:0`, `g2-fallback.txt:0`; sum `1` |
| final-result count | `/usr/bin/grep -c 'GONECELL' g2-fallback.txt` | 0 | `1` |
| context | `/usr/bin/grep -c` for `HEADCELL` and `SIDECELL`; `/usr/bin/grep -cE -- 'REGEX'` on `g2-fallback.txt` | 0 | `1`, `1`; `3` |

Predicate (`spec.md` §3.4 cell ②, AC-009): the construction check exited non-zero (`1`) — holds; the
detection reading totals ≥ 1 (`1`) — holds; the scan carrying the final result, the full-history scan,
exited 0 — holds; it reports `GONECELL` ≥ 1 (`1`) — holds, through behaviour (a), falling back to a full
`--all` scan. None of the untrustworthy shapes occurred: the error was followed by a follow-on scan,
the final scan exited 0, `GONECELL` was not 0, and the detection total was not 0.

verdict: trustworthy

#### Gaps in cells 1 and 2

- The scan command's `$(cat …)` substitution was not executed; the pinned hand expansion stood in for
  it. An empty store and argument-list limits were not exercised.
- Only local branches were exercised; tags, remote-tracking refs, stash entries, and merge commits were
  not.
- Only the PEM-header alternative of `REGEX` was exercised.
- Cell 2 exercised one missing tip, listed first in its store. Git stops at the first bad object, so a
  store with several missing tips reports only one per attempt; not exercised.
- Cell 3 is not taken in this segment. It needs the lead's approval, committed in its own commit first
  (`plan.md` §C item 4).
