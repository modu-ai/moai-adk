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

### Run cautions

Recorded at the lead's instruction before the M1 gate (card t629, 2026-09-11). These notes read plan
audit iteration 3 (`.moai/reports/t629/plan-audit-iter3.md`); they change nothing in `spec.md`,
`plan.md`, or `acceptance.md`.

1. N1: AC-001 files a non-zero exit from R1-R4 under a construction gap, while AC-008 and AC-009 read
   the same exit as a predicate failure. A procedure that errors under AC-001 therefore stays a gap and
   never becomes an explicit fail, although the Definition of Done still blocks it. When AC-001 is
   read, a non-zero exit caused by the procedure itself counts as a failure signal.
2. N2: AC-004's Given names all of the card's commits, but its diff range ends at `K`; commits made
   after `K` are outside that range.
3. N3: the findings counting rule assumes each finding line carries the matched source line. Findings
   are written as the source line, whatever display form the procedure uses.
4. Exit codes (iteration 3, out-of-scope note): AC-008 and gate cell 1 require every scan to exit 0.
   The gate keeps "no match" apart from "tool error": the pinned scan commands end in no filter that
   exits non-zero on an empty result, a search tool's no-match exit (grep exit 1) is read as a zero
   count and never as a scan failure, and a tool error (grep exit 2, a non-zero git exit) is recorded
   and never swallowed.
5. Gap (D19a): AC-016 proves that each pinned line appears in the edited section, not that the
   section prescribes no other timing for tip recording.

### Iteration 4 handling

Lead-accepted handling of the optional findings of plan audit iteration 4
(`.moai/reports/t629/plan-audit-iter4.md`, PASS 0.87), recorded at the lead's instruction before gate
round 2 (card t629, 2026-09-11). Nothing in `spec.md`, `plan.md`, or `acceptance.md` changes.

- D21: not taken up in this segment; `spec.md` §3.4 § Tip store and scan command is the form the
  round 2 pin follows.
- D22: gate cells 1 and 2 of round 2 measure that `git log` accepts the caret store on standard input
  and reports a missing tip. That the store's lines exclude the recorded tips is measured only by gate
  cell 3's scope bound. Cell 1 also records, as supplementary evidence outside its predicate, the
  commits in scope of its R2 scan with the store fed and without it.
- D23: the re-gate follows `plan.md` M3 in order — the move of round 1 out of this file in a commit of
  its own, then the round 2 pin in a commit of its own, then cells 1 and 2 in a commit of their own;
  cell 3 follows a fresh lead approval committed alone.
- D24: see § Design probes below.
- D25: the round 2 pin writes the store redirect on the scan command's own line
  (`--stdin < .moai/state/secrets-scan-tips.txt`).
- D26: applies to gate cell 3, taken in a later segment. The audit's suggested fix — copy the store a
  run reads before that run, and derive the scope, the tips excluded, and the plain file for `A` and
  `L` from that copy — is for that segment to take up.

### Design probes

The design probes taken before the round 1 pin moved with that round and now sit in
`.moai/reports/t629/gate-round-1.md`, inside its pinned-procedure block. They were syntax checks on a
separate throwaway repository, not gate evidence. Gate round 2 adopts one observation from them only:
`git log -p --all --stdin` reading a caret-prefixed line for an object that does not exist exited 128,
with 0 bytes on stdout and an error line of the form `fatal: bad object <name>` that names the object
without the caret (`git version 2.50.1 (Apple Git-155)`). The conclusion drawn there — that the store
therefore holds plain object names — is superseded by SPEC 0.3.0, whose store holds caret-prefixed
names and whose missing-tip handling removes the caret before comparing (`spec.md` §3.4 § Tip store
and scan command).

### Pinned procedure

Gate round 2. Pinned on branch `WT-secret-scan-refs` on top of `e298f7336`, in a commit of its own that
precedes the first command of this round. It replaces the round 1 pin, now in
`.moai/reports/t629/gate-round-1.md`, and follows `spec.md` 0.3.0 § Tip store and scan command. The
regex below is written exactly as `review.md` writes it. Paths are relative to the project root.

Tip recording: before the scan starts, record the tip of every ref with `git for-each-ref --format='^%(objectname)' > .moai/state/secrets-scan-tips.next`, and replace `.moai/state/secrets-scan-tips.txt` with that file only after the scan that carries the final result exits 0.
Full-history scan: git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})'
Scan command: git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' --stdin < .moai/state/secrets-scan-tips.txt
Missing-tip handling: when the scan exits non-zero and its error output reports `bad object` for a tip whose line in `.moai/state/secrets-scan-tips.txt`, without its leading caret, names that object, report that tip as missing and run the full-history scan in its place; any other non-zero exit is a scan failure that is reported and leaves `.moai/state/secrets-scan-tips.txt` unchanged.
Uncovered commits: Commits that no ref and no HEAD reaches, such as commits reachable only through a reflog, are outside every scan step in this procedure, and no step scans them.

Which step runs: the full-history scan runs when `.moai/state/secrets-scan-tips.txt` does not exist,
and the scan command runs otherwise. Because the tips are recorded before the scan starts, a commit
that lands while the scan runs is outside the recorded tips and falls to the next review instead of
being skipped; a commit that lands between the recording and the scan's own read of the refs is
scanned twice. `Uncovered commits:` carries the REQ-004 sentence that AC-006 checks, fixed here before
any document edit. It is reused verbatim from round 1: the stdin form does not make it false, because
every store line is a negated revision and adds nothing to the scope `--all` sets.

Gate execution method, fixed before the gate. The gate runs these commands as written: the scan
command reads the store through standard input, so no tip is typed by hand and no command
substitution is involved. On the fixture, `git` runs with `-C <fixture root>`, and each path under
`.moai/state/` that a command names is written with the fixture root in front, because the shell, not
`git`, resolves a redirect; the arguments are otherwise unchanged. The tip store sits at
`.moai/state/` under the fixture's own root, the pinned path.

### Gate evidence

Gate round 2. Taken 2026-09-11 from this worktree on branch `WT-secret-scan-refs`, HEAD `68c56be0d`
(the round 2 pin commit), on a fresh throwaway fixture repository outside this repository at
`SP/m1gate3/fx`, where `SP` is the session scratchpad. `git version 2.50.1 (Apple Git-155)`. Fixture
set-up: `git init -b main`, a fixture `user.name` and `user.email`, `commit.gpgsign=false`, and
`core.hooksPath=/dev/null`, read back with `git config --list --local`. `REGEX` is the scan regex as
`review.md` writes it.

- Markers: PEM-style private-key header lines with distinct uppercase labels, each written by `printf`
  from a format string split around two fragments and not reproduced here. Each marker file was checked
  with `/usr/bin/grep -cE -- 'REGEX' <file>` → `1`; the plain `gone.txt` → `0`.
- Every command's output went to a file under `SP/m1gate3/`, and every exit code was read with
  `; echo "exit=$?"`, never through a pipe. A grep exit of 1 is read as a zero count; no grep exited 2,
  and every non-zero git exit is recorded in the tables.
- Execution method as pinned: `git -C <fixture root>`, and each `.moai/state/` path written with the
  fixture root in front. The scan command read the store through standard input; no tip was typed as a
  scan argument and no command substitution ran.
- Tip store: `SP/m1gate3/fx/.moai/state/secrets-scan-tips.txt`, recorded through `.next` and replaced
  with `mv` only after the scan carrying the final result exited 0.
- Fixture object names are kept in `SP` only. `<G1>` below stands for the recorded tip of branch
  `gone`: the one line of `SP/m1gate3/g2-gone.txt`, a full 40-character object name without a caret.

#### Gate cell 1

| Step | Command (in the fixture; outputs in `SP/m1gate3/`) | Exit | Reading |
|---|---|---|---|
| C0 | clean `keys.txt` committed on `main` | 0 | — |
| clean-history scan | `git log -p --all -G 'REGEX' > g1-base.txt 2> g1-base.err` | 0 | `wc -c` 0 and 0 bytes |
| R1 tip recording | `git for-each-ref --format='^%(objectname)' > .moai/state/secrets-scan-tips.next` | 0 | 1 line; `/usr/bin/grep -c '^\^'` → `1` |
| which step runs (R1) | `test -e .moai/state/secrets-scan-tips.txt` | 1 | no store, so the full-history scan runs |
| R1 full-history scan | `git log -p --all -G 'REGEX' > g1-r1.txt 2> g1-r1.err` | 0 | 0 and 0 bytes; `mv` to the store, exit 0 (1 line) |
| S1 | branch `side` at C0; `SIDECELL` marker in `side.txt`, committed on `side` | 0 | `show --stat side`: 1 file, 1 insertion; its parent line `cmp` against C0's name → exit 0 |
| C1 | `HEADCELL` marker appended to `keys.txt`, committed on `main` | 0 | `side.txt` absent on `main` (`test -e` exit 1) |
| construction | `git merge-base --is-ancestor side HEAD` | 1 | — |
| R2 tip recording | as R1, in the same step as the construction check | 0 | store present, so the scan command runs; the store copied to `r2-tips-read.txt` before the scan (`cmp` exit 0, 1 line) |
| R2 scan command | `git log -p --all -G 'REGEX' --stdin < .moai/state/secrets-scan-tips.txt > g1-r2.txt 2> g1-r2.err` | 0 | 666 and 0 bytes; `mv` to the store, exit 0 (2 lines) |
| construction, re-read before the counts | `git merge-base --is-ancestor side HEAD` | 1 | — |
| counts | `/usr/bin/grep -c 'HEADCELL' g1-r2.txt`; `/usr/bin/grep -c 'SIDECELL' g1-r2.txt` | 0; 0 | `1`; `1` |
| context | `/usr/bin/grep -cE -- 'REGEX' g1-r2.txt`; `/usr/bin/grep -c '^commit ' g1-r2.txt` | 0; 0 | `2`; `2` |

Predicate (`spec.md` §3.4 cell ①, AC-008): the clean-history scan printed 0 bytes — holds; every scan
exited 0 (clean-history scan, R1, R2) — holds; `is-ancestor` exited 1 immediately before the counts —
holds; in R2, `HEADCELL` ≥ 1 (`1`) and `SIDECELL` ≥ 1 (`1`) — both hold. The side branch's marker was
reported by the stdin-form scan in which it first became reachable, next to the HEAD-line control.

verdict: trustworthy

Supplementary evidence, outside the predicate (iteration 4 handling, D22): the commits in scope of R2's
revision arguments, listed without `-p` and `-G`.

| Listing | Command | Exit | `wc -l` |
|---|---|---|---|
| with the store R2 read | `git log --format=%H --all --stdin < r2-tips-read.txt > g1-d22-stdin.txt` | 0 | `2` |
| without a store | `git log --format=%H --all > g1-d22-all.txt` | 0 | `3` |

#### Gate cell 2

Taken on the same fixture after cell 1.

| Step | Command (in the fixture; outputs in `SP/m1gate3/`) | Exit | Reading |
|---|---|---|---|
| G1 | branch `gone` from `main`; plain `gone.txt` committed on `gone` | 0 | `for-each-ref --contains gone` → `refs/heads/gone` only |
| record `<G1>` | `git rev-parse refs/heads/gone > g2-gone.txt` | 0 | 1 line, 41 bytes |
| R3 tip recording | `git for-each-ref --format='^%(objectname)' > .moai/state/secrets-scan-tips.next` | 0 | 3 lines, 3 caret-prefixed; stripped with `sed 's/^\^//'`, `/usr/bin/grep -cxF -f g2-gone.txt` → `1` |
| R3 scan command | `git log -p --all -G 'REGEX' --stdin < .moai/state/secrets-scan-tips.txt > g2-r3.txt 2> g2-r3.err` | 0 | 0 and 0 bytes; `mv` to the store, exit 0 (3 lines) |
| membership | `sed 's/^\^//' .moai/state/secrets-scan-tips.txt > g2-store-plain.txt`; `/usr/bin/grep -cxF -f g2-gone.txt g2-store-plain.txt` | 0; 0 | caret lines left in the plain file `0` (grep exit 1); membership `1` |
| delete | `git branch -D gone` | 0 | — |
| expire | `git reflog expire --expire=now --all` | 0 | — |
| prune | `git gc --prune=now --quiet` | 0 | — |
| construction | `git cat-file -e <G1>` | 1 | second reading: `git cat-file --batch-check < g2-gone.txt` exit 0, one line ending ` missing` (`/usr/bin/grep -c` → `1`) |
| L1 | `GONECELL` marker in `after.txt`, committed on `after` created from `main` | 0 | — |
| construction | `git merge-base --is-ancestor after HEAD` | 1 | — |
| R4 tip recording | as R3 | 0 | 3 lines; the store copied to `r4-tips-read.txt` before the scan (`cmp` exit 0, 3 lines) |
| R4 scan command | `git log -p --all -G 'REGEX' --stdin < .moai/state/secrets-scan-tips.txt > g2.txt 2> g2.err` | 128 | `g2.txt` 0 bytes; `g2.err` 59 bytes, 1 line; `/usr/bin/grep -c '^fatal: bad object '` → `1` |
| handling check | `sed 's/^\^//' r4-tips-read.txt > g2-r4-store-plain.txt`; `/usr/bin/grep -cxF -f g2-gone.txt g2-r4-store-plain.txt`; `/usr/bin/grep -cF -f g2-gone.txt g2.err` | 0; 0 | `1`; `1`: the store line R4 read, without its caret, names the object the error reports, so the handling runs the full-history scan. Control: the same `-cxF` over the caret store `r4-tips-read.txt` → `0` (exit 1) |
| full-history scan in its place | `git log -p --all -G 'REGEX' > g2-fallback.txt 2> g2-fallback.err` | 0 | 984 and 0 bytes; `mv` R4's `.next` to the store, exit 0 (3 lines; stripped, `<G1>` counted `0`, exit 1) |
| detection reading | `/usr/bin/grep -cF -f g2-gone.txt g2.err g2.txt g2-fallback.txt` | 0 | `g2.err:1`, `g2.txt:0`, `g2-fallback.txt:0`; sum `1` |
| final-result count | `/usr/bin/grep -c 'GONECELL' g2-fallback.txt` | 0 | `1` |
| context | `/usr/bin/grep -c` for `HEADCELL` and `SIDECELL`; `/usr/bin/grep -cE -- 'REGEX'` and `/usr/bin/grep -c '^commit '` on `g2-fallback.txt` | 0 | `1`, `1`; `3`; `3` |

Predicate (`spec.md` §3.4 cell ②, AC-009 as amended): the construction check exited non-zero (`1`) —
holds; the detection reading, with the name taken from `g2-gone.txt` and carrying no caret, totals ≥ 1
(`1`) — holds; the scan carrying the final result, the full-history scan, exited 0 — holds; it reports
`GONECELL` ≥ 1 (`1`) — holds, through behaviour (a), falling back to a full `--all` scan. None of the
untrustworthy shapes occurred: the error was followed by a follow-on scan, the final scan exited 0,
`GONECELL` was not 0, and the detection total was not 0.

verdict: trustworthy

#### Gaps in cells 1 and 2

- Cells 1 and 2 measure that `git log` accepts the caret store on standard input and that the pinned
  handling detects and recovers from a missing tip. That the store's lines exclude the recorded tips is
  proven only by gate cell 3's scope bound; the D22 counts above are supplementary and sit outside every
  predicate.
- The D26 snapshot rule — copy the store a run reads before that run, and derive the scope, the tips
  excluded, and the plain file for `A` and `L` from that copy — applies to gate cell 3, not taken in
  this segment. Cell 3 needs the lead's approval, committed in its own commit first (`plan.md` §C
  item 4).
- Only local branches were exercised; tags, remote-tracking refs, stash entries, and merge commits were
  not.
- Only the PEM-header alternative of `REGEX` was exercised.
- Cell 2 exercised one missing tip, on line 1 of a 3-line store. A store with several missing tips was
  not exercised.
- Not exercised: an empty store, a non-zero scan exit other than a missing tip (the pinned "any other
  non-zero exit" branch), and a commit landing between the tip recording and the scan.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Decision: serial

Recorded by the orchestrator (lane-5) before the first run-phase spawn, 2026-09-11, on tree `7a711c549`.

**Gate record.** Plan audit iteration 3 PASS 0.89 (`.moai/reports/t629/plan-audit-iter3.md`, commit
`7a711c549`); the lead read the verdict file and judged PASS. Implementation Kickoff Approval was given
by the lead under the batch's delegated conditions (audit PASS, zero blocking findings, a one-sentence
change), with the design decisions (Option 2, exact allowlist) already taken by the operator.
Progression: autonomous, with two mandatory stops set by the lead: (a) after the M1 measure-first gate,
report cells 1-3 and do not start M2 before the lead confirms; (b) cell 3 (full ref-reachable history on
this repository) runs only after a separately committed lead approval, requested together with the
load-recording plan.

**Input parameters.** Tier M; scope two review workflow document copies plus SPEC/evidence files;
domain count 1 (documentation with a measurement gate); file language mix markdown only; concurrency
benefit low (ordered gate, single integration surface); Agent Teams not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-milestone gated work, not a one-line change |
| serial | **yes** | ordered measure-first gate with stop points; one writer on the worktree |
| fanout | no | not multi-domain research; cells share one fixture and one progress record |
| sweep | no | two files, no high-volume mechanical transform |

**Justification.** The run phase is an ordered sequence whose later milestones depend on the gate's
verdicts, and two of its stops are human confirmations, so a single sequential `manager-develop`
delegation per stop segment is the simplest mode that fits. Parallel spawns would add a second writer
to the progress record the ordering checks (AC-010, AC-011, AC-016) read.
