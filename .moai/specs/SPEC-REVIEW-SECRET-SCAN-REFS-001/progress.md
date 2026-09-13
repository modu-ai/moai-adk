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
- Tool provenance (`verification-claim-integrity.md` §2.2): `moai spec lint` for this round was judged
  by the installed `v3.2.0-rc.7` build, which is neither an ancestor nor a descendant of this tree; it
  is not a build made from this tree.

#### Gate cell 3

Gate round 2, cell ③ — the timing record on this repository (`spec.md` §3.4 cell ③, AC-010). Taken
2026-09-11 (11:51-11:53 +0900) from this worktree on branch `WT-secret-scan-refs`, HEAD `7e06766c8`
(the lead-approval commit; parent `f167a9cd8`), clean tree. `git version 2.50.1 (Apple Git-155)`.
`REGEX` is the scan regex as `review.md` writes it. `SP` is the session scratchpad; every file named
`SP/g3-*` below was written by a command of this attempt (plain `>` overwrite). The files of the two
earlier, stopped attempts were not read as evidence. Scan output was read only with `wc -c`, `wc -l`,
and `/usr/bin/grep -c`; no SHA, path, or matched value of any matching commit is written here.

Lead's start notice, quoted verbatim as received on 2026-09-11:

> [리드 — t629 셀 ③ 시작] lane-10 의 internal/cli 재측정이 끝났고 창도 반납됐습니다(11:31). 셀 ③ 을 시작하세요 — 새 에이전트로 ps 점검(git log -p·go test 0건)·끝점 기록·uptime 부터 새로. lane-3 의 무거운 대조군은 당신 스캔이 끝난 뒤로 잡아 두었습니다. 끝나면 real(total)·exit·범위 수치와 판정 파일 경로를 보고해 주세요. M2 금지 유지.

Lead's approval of the measurement tool, quoted verbatim as received on 2026-09-11:

> [리드 — t629 셀 ③ 측정 도구] A 승인합니다. 판독 기준이 비율이라 `real` → 셸 예약어 `time` 의 `total` 형식 차이는 판정을 바꾸지 않습니다. 조건: (1) 증거에 "승인된 측정 도구 차이: /usr/bin/time -p 가 가드에 거부돼 셸 예약어 time 사용, 리드 승인"과 레인이 재현한 거부/통과 두 명령을 함께 기록 (2) 스캔 명령 본체(`git log -p --all -G 'REGEX' --stdin < 저장소`)는 고정 문자열 그대로, 바뀌는 건 감싸는 래퍼뿐 (3) 낡은 준비물(끝점 729줄, 이전 load)은 쓰지 말고 ps 점검·끝점 기록·uptime 을 새로. AC-010 문구와의 차이는 sync 단계에서 문서에 반영할지 판정 때 정합니다. M2 금지 유지.

Approved measurement-tool difference: `/usr/bin/time -p` is refused by the worktree guard, so the shell
reserved word `time` is used; wall time is its `total` field rather than `real`; lead-approved.
Reproduction taken by the orchestrator on 2026-09-11 in worktree t629 at `7e06766c8`, on a one-commit
range, without `-p` and without `--all`:

| Form | Command | Result |
|---|---|---|
| refused | `/usr/bin/time -p git log --format=%h -G '(zz qq\|yy)' HEAD~1..HEAD > SP/lane-probe-time.txt 2> SP/lane-probe-time.err` | guard message begins "This session is isolated in the worktree … but this command hands time the text (zz qq\|yy) …" |
| passed | `time git log --format=%h -G '(zz qq\|yy)' HEAD~1..HEAD > SP/lane-probe-zsh.txt 2> SP/lane-probe-zsh.err` | tool output `git log --format=%h -G '(zz qq\|yy)' HEAD~1..HEAD >  2>   0.04s user 0.02s system 42% cpu 0.152 total` |

Scan command bodies below are byte-for-byte the pinned strings; only the leading `time` keyword, the
redirects, and the trailing `; echo "exit=$?"` are added. The `time` line lands in the tool output, not
in a file; it is copied verbatim.

Pre-start check, immediately before the first run:

| Step | Command | Exit | Reading |
|---|---|---|---|
| processes | `ps -axo pid,etime,command > SP/g3-ps-before.txt` | 0 | — |
| `git log -p` count | `/usr/bin/grep -c '[g]it log -p' SP/g3-ps-before.txt` | 1 | `0` |
| `go test` count | `/usr/bin/grep -c '[g]o test' SP/g3-ps-before.txt` | 1 | `0` |
| load | `uptime > SP/g3-load-prestart.txt` | 0 | `load averages: 11.48 7.18 8.10` |

First run (no recorded tips):

| Step | Command | Exit | Reading |
|---|---|---|---|
| load before | `uptime > SP/g3-first-load-before.txt` | 0 | `10.93 7.21 8.09` |
| HEAD | `git rev-parse HEAD > SP/g3-first-head.txt` | 0 | `7e06766c8` |
| tip recording | `git for-each-ref --format='^%(objectname)' > SP/g3-tips.next` | 0 | `wc -l` `729`; `/usr/bin/grep -vc '^\^'` → `0` (exit 1) |
| which step runs | `test -e SP/g3-tips.txt` | 1 | no store, so the full-history scan runs |
| scan | `time git log -p --all -G 'REGEX' > SP/g3-first.txt 2> SP/g3-first.time; echo "exit=$?"` | 0 | `git log -p --all -G  >  2>   56.35s user 0.76s system 95% cpu 1:00.01 total` |
| load after | `uptime > SP/g3-first-load-after.txt` | 0 | `6.89 6.81 7.86` |
| store replaced | `mv SP/g3-tips.next SP/g3-tips.txt` | 0 | only after the scan exited 0 |
| scope | `git log --format=%H --all > SP/g3-first-scope.txt` | 0 | `wc -l` `12490` |
| matching commits | `/usr/bin/grep -c '^commit ' SP/g3-first.txt` | 0 | `15` |
| bytes | `wc -c SP/g3-first.txt`; `wc -c SP/g3-first.time` | 0 | `560344`; `0` |

Incremental run (immediately after the first):

| Step | Command | Exit | Reading |
|---|---|---|---|
| store snapshot (D26) | `cp SP/g3-tips.txt SP/g3-inc-tips-read.txt`; `cmp SP/g3-tips.txt SP/g3-inc-tips-read.txt` | 0; 0 | both `wc -l` `729` |
| plain tips | `sed 's/^\^//' SP/g3-inc-tips-read.txt > SP/g3-tips-plain.txt` | 0 | `wc -l` `729` (equal); `/usr/bin/grep -c '^\^'` → `0` (exit 1) |
| `A` | `git rev-list --count --stdin < SP/g3-tips-plain.txt > SP/g3-A.txt` | 0 | `12488` |
| `B` | `git rev-list --count --all > SP/g3-B.txt` | 0 | `12490` |
| `L` | `git rev-list --count --not --all --stdin < SP/g3-tips-plain.txt > SP/g3-L.txt` | 0 | `0` |
| scope | `git log --format=%H --all --stdin < SP/g3-inc-tips-read.txt > SP/g3-inc-scope.txt` | 0 | `wc -l` `2` |
| load before | `uptime > SP/g3-inc-load-before.txt` | 0 | `6.66 6.77 7.79` |
| HEAD | `git rev-parse HEAD > SP/g3-inc-head.txt` | 0 | `7e06766c8` |
| tip recording | `git for-each-ref --format='^%(objectname)' > SP/g3-tips.next` | 0 | `wc -l` `729` |
| scan | `time git log -p --all -G 'REGEX' --stdin < SP/g3-tips.txt > SP/g3-inc.txt 2> SP/g3-inc.time; echo "exit=$?"` | 0 | `git log -p --all -G  --stdin <  >  2>   0.09s user 0.07s system 67% cpu 0.229 total` |
| load after | `uptime > SP/g3-inc-load-after.txt` | 0 | `8.49 7.15 7.91` |
| store read unchanged | `cmp SP/g3-tips.txt SP/g3-inc-tips-read.txt` | 0 | the store the scan read equals the snapshot |
| `B2` | `git rev-list --count --all > SP/g3-B2.txt` | 0 | `12490` |
| store replaced | `mv SP/g3-tips.next SP/g3-tips.txt` | 0 | only after the scan exited 0 and `B2` = `B` |
| matching commits | `/usr/bin/grep -c '^commit ' SP/g3-inc.txt` | 1 | `0` |
| bytes | `wc -c SP/g3-inc.txt`; `wc -c SP/g3-inc.time` | 0 | `0`; `0` |

Record per run:

| Field | First run | Incremental run |
|---|---|---|
| HEAD | `7e06766c8` | `7e06766c8` |
| load before / after | `10.93 7.21 8.09` / `6.89 6.81 7.86` | `6.66 6.77 7.79` / `8.49 7.15 7.91` |
| wall time (`total`) | `1:00.01` (60.01 s) | `0.229` s |
| exit | `0` | `0` |
| tips excluded | `0` (no store) | `729` |
| commits in scope | `12490` | `2` |
| matching commits (count only) | `15` | `0` |
| output / error bytes | `560344` / `0` | `0` / `0` |

Bound: `B − A + L` = `12490 − 12488 + 0` = `2`. `B2` = `12490` = `B`, so no ref moved during the
incremental run. Incremental commits in scope `2` ≤ bound `2` — holds.

Ratios: wall time, first ÷ incremental = `60.01 / 0.229` ≈ `262`; commits in scope, first ÷ incremental
= `12490 / 2` = `6245`. Measured under contention on a shared machine with other lanes active — the
ratio is the reading; the absolute seconds are not a clean benchmark.

Predicate (`spec.md` §3.4 cell ③, AC-010): every recorded field is present for both runs — holds; both
runs exited 0 — holds; `A` and `L` were computed (no recorded tip was absent) — holds; the incremental
commits-in-scope count `2` is at most `B − A + L` = `2` — holds. None of the untrustworthy shapes
occurred: no field is missing, no exit is non-zero, and the incremental scope does not exceed the bound
(a re-scan of all history would have shown a scope of `B` = `12490`).

Differences from AC-010's wording, named here and left for the sync-phase decision:

- Timing tool: the approved difference above. The `time` line is in the tool output, so
  `SP/g3-<run>.time` carries the scan's error stream only.
- Store snapshot (D26, lead-approved "회차마다 저장소 사본을 먼저 뜨고 그 사본으로 계산"): the
  incremental run's plain tips, scope listing, and tips-excluded count come from the snapshot
  `SP/g3-inc-tips-read.txt`, not from `SP/g3-tips.txt` directly. The scan itself reads `SP/g3-tips.txt`,
  per the pinned form; `cmp` exited 0 against the snapshot both before and after the scan.
- File names follow AC-010 where the dispatch named others: the error stream in `SP/g3-<run>.time`
  (dispatch: `.err`), the plain file `SP/g3-tips-plain.txt` (dispatch: `SP/g3-inc-tips-plain.txt`), and
  the approval-ordering log `SP/g3-evidence.txt` (dispatch: `SP/g3-e3.txt`).
- The first run's tips-excluded field is `0` from the store's absence (`test -e` exit 1), since AC-010's
  `wc -l SP/g3-tips.txt` has no file to count before the first run.

verdict: trustworthy

#### Gaps in cell 3

- AC-010 names `/usr/bin/time -p` and its `real` field; this record uses the shell reserved word `time`
  and its `total` field (approved difference above). Whether AC-010's wording changes is left for the
  sync phase.
- One machine, one git build (`git version 2.50.1 (Apple Git-155)`), one pair of runs, under contention.
  No clean-machine timing.
- The first run's scope was listed after its scan finished, while its tips were recorded before the scan
  started; the two commits counted by `B − A` (`2`) became reachable in that window and fall inside the
  first run's scope count but not inside its recorded tips. Whether the first scan itself covered them
  was not observed.
- A non-pinned side note: one non-git command of this attempt (a compound `cp`/`cmp`/`sed` line written
  with a shell variable for the scratchpad path) was refused by the worktree guard before anything ran;
  it was re-issued as plain commands with literal paths. No scan was refused or re-taken.

### Lead approval for gate cell 3

Recorded by the orchestrator (lane-5) in a commit of its own, after the gate round 2 evidence commit
`f167a9cd8` and before any gate cell 3 command (`plan.md` §C item 4, AC-010). The lead's message,
quoted verbatim as received on 2026-09-11:

> [리드 — t629 셀 ③ 승인]
> 셀 ①② 판정 수용합니다. 제가 읽은 근거는 보고 본문이고, 증거 커밋 f167a9cd8 에 대해서는 판정 전에 한 번 더 읽겠습니다. Gap 3건(빈 저장소, 그 밖의 non-zero 분기, 스캔 도중 착지한 커밋)과 lint 판정 빌드 귀속 Gap(VCI §2.2)은 그대로 Gap 으로 남기세요.
> 셀 ③ 은 보고한 계획대로 승인합니다: 실행 방식 (ii), 고정 명령 `/usr/bin/time -p git log -p --all -G 'REGEX' --stdin < SP/g3-tips.txt`, 회차마다 저장소 사본을 먼저 뜨고 그 사본으로 계산, uptime 전후·real·exit·B/B2 기록, 경합 속 측정이라 비율로 판독, 매칭 커밋은 개수만 기록. M2 금지는 유지합니다.
> 실행 순서 조건: 지금 lane-3 이 internal/cli 테스트 슬롯을 쓰고 있습니다(load 12.83). 그 실행이 끝났다고 제가 알릴 때까지 셀 ③ 은 시작하지 마세요. 그 전에 승인 인용 단독 커밋은 해 두셔도 됩니다. 시작하기 직전에 git log -p / go test 프로세스가 0건인지 확인하세요.

Summary in English: the lead accepts the round 2 cell 1 and cell 2 verdicts (and will re-read
`f167a9cd8` before judging), keeps the four gaps above as gaps, and approves gate cell 3 on the
reported plan — execution form (ii), per-run store snapshots, load averages before and after, wall
time, exit codes, `B`/`B2` re-checks, ratios as the reading under contention, and matching commits as
counts only. Two conditions bind the start: wait for the lead's notice that lane-3's `internal/cli` test
run has finished, and confirm immediately before starting that no `git log -p` or `go test` process is
running. M2 stays forbidden.

### Develop drift note

Recorded at the lead's instruction in the M2 commit (card t629, 2026-09-11). The facts below were
measured by the orchestrator in its own session; this segment did not re-measure them.

- Card-id collision: local develop's `91aca1011` ("merge: workflow audit F23 into develop (card
  t629)") is another session's commit that carries this card id. It is not part of card t629's work
  (lead note).
- Local develop moved `4c99d973e` → `ee99507fb`. `git diff --stat 4c99d973e ee99507fb -- <both
  review.md paths>` and `git diff --stat feeecc980 ee99507fb -- <both review.md paths>` printed
  nothing; `git cat-file -e ee99507fb:<each path>` exited 0 for both paths; the control
  `git diff --stat 4c99d973e ee99507fb -- <both workflows directories>` reported 19 files changed.
  Neither copy changed on develop, so this drift does not trigger the pinned-block re-comparison
  after the absorb. The absorb itself happens later, in the integration window.

### Draft section wording (M2)

Drafted 2026-09-11 on branch `WT-secret-scan-refs` on top of `023a25e7a`, after the lead confirmed
the M1 gate and approved M2. The block below, without its outer four-backtick fence, is the text M4
copies verbatim into both copies of the review workflow document, in place of the current
`#### Secrets Scan (Incremental with Checkpoint)` section. It carries the pinned `Tip recording:`
and `Missing-tip handling:` lines word for word, the pinned full-history scan and scan command
inside its command lines, and the pinned `Uncovered commits:` sentence. The pinned procedure itself
is not changed. The allowlist representation chosen for it is recorded in the next subsection.

````markdown
#### Secrets Scan (Incremental with Checkpoint)

Scan git history for credential leaks incrementally. The checkpoint is a tip store, `.moai/state/secrets-scan-tips.txt`: the tip of every ref at the last completed scan, one caret-prefixed object name per line. Each line reaches git as a negated revision, so a scan that reads the store leaves out every commit a recorded tip reaches.

Tip recording: before the scan starts, record the tip of every ref with `git for-each-ref --format='^%(objectname)' > .moai/state/secrets-scan-tips.next`, and replace `.moai/state/secrets-scan-tips.txt` with that file only after the scan that carries the final result exits 0.

Where the tip store does not exist (first run), or an explicit full-scan flag is passed, run the full-history scan. It scans every commit reachable from any ref or from HEAD:

```bash
git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' > .moai/state/secrets-scan-output.txt
```

Otherwise run the scan command. It scans every commit reachable from any ref or from HEAD that no recorded tip reaches — the commits that became reachable since the last completed scan — and does not scan a commit a recorded tip reaches:

```bash
git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' --stdin < .moai/state/secrets-scan-tips.txt > .moai/state/secrets-scan-output.txt 2> .moai/state/secrets-scan-error.txt
```

Missing-tip handling: when the scan exits non-zero and its error output reports `bad object` for a tip whose line in `.moai/state/secrets-scan-tips.txt`, without its leading caret, names that object, report that tip as missing and run the full-history scan in its place; any other non-zero exit is a scan failure that is reported and leaves `.moai/state/secrets-scan-tips.txt` unchanged.

To find the tip the error output names, strip the caret from every store line and search the error output for those object names; a printed line that contains `bad object` names a missing tip:

```bash
sed 's/^\^//' .moai/state/secrets-scan-tips.txt > .moai/state/secrets-scan-tips-plain.txt
grep -F -f .moai/state/secrets-scan-tips-plain.txt .moai/state/secrets-scan-error.txt
```

Because the tips are recorded before the scan starts, a commit that lands while the scan runs is outside the recorded tips and falls to the next review instead of being skipped; a commit that lands between the recording and the scan's own read of the refs is scanned twice.

Merge commits show no patch in the scan output, so a line that only a merge commit's own changes introduce, such as a conflict resolution, is reported by neither scan.

Uncovered commits: Commits that no ref and no HEAD reaches, such as commits reachable only through a reflog, are outside every scan step in this procedure, and no step scans them.

Each review also scans the working tree; the checkpoint does not govern that step, and the example-value rule below applies to its matches as well.

**Known example values.** A match is suppressed only when the text the regex matched equals a listed value exactly, character for character. The list holds only publicly published example values; it has one entry, the example access key ID that AWS publishes in its documentation. No path is excluded from any scan step, and a value that differs from every listed value, even by one character, is still reported wherever it sits. Each listed value is written here as a digest rather than as the value, so this document holds no line the scan's regex matches. The digest is the full-length git blob object name of the value followed by one newline, as `git hash-object --no-filters` computes it in a repository that uses the default SHA-1 object format; in a repository that uses the SHA-256 object format no digest matches, so nothing is suppressed.

Run the following over the output of the scan that carried the final result. Each digest is paired with the value on the same line of `.moai/state/secrets-scan-distinct.txt`, so `mkdir` must create a new, empty directory; if it fails, a previous run left the directory behind: remove it and start again. A `grep` exit status of 1 means no line matched and leaves an empty file; it is not an error.

```bash
printf '%s\n' 05c61e935c693e4244743a05a1ba4ae33bd71d64 > .moai/state/secrets-scan-allowlist.txt
grep -oE -- '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' .moai/state/secrets-scan-output.txt > .moai/state/secrets-scan-matches.txt
sort -u .moai/state/secrets-scan-matches.txt > .moai/state/secrets-scan-distinct.txt
mkdir .moai/state/secrets-scan-split
split -a 6 -l 1 .moai/state/secrets-scan-distinct.txt .moai/state/secrets-scan-split/v.
find .moai/state/secrets-scan-split -type f > .moai/state/secrets-scan-found.txt
sort .moai/state/secrets-scan-found.txt > .moai/state/secrets-scan-paths.txt
git hash-object --no-filters --stdin-paths < .moai/state/secrets-scan-paths.txt > .moai/state/secrets-scan-digests.txt
paste .moai/state/secrets-scan-digests.txt .moai/state/secrets-scan-distinct.txt > .moai/state/secrets-scan-table.txt
grep -vwF -f .moai/state/secrets-scan-allowlist.txt .moai/state/secrets-scan-table.txt > .moai/state/secrets-scan-kept.txt
cut -f2 .moai/state/secrets-scan-kept.txt > .moai/state/secrets-scan-unsuppressed.txt
grep -F -f .moai/state/secrets-scan-unsuppressed.txt .moai/state/secrets-scan-output.txt > .moai/state/secrets-scan-findings.txt
rm -r .moai/state/secrets-scan-split
```

Each line of `.moai/state/secrets-scan-findings.txt` is a finding: a source line of the scan output that carries a matched text no listed value equals, written as the scan printed it. The scan output also prints context lines and every hunk of a matching file; a line printed beside a finding is not a finding, and a line whose only matches are listed values is not reported.

Cross-reference findings against `.gitignore` to distinguish historical leaks from working-tree exposure.
This scan is separate from working-tree-only scanners.
````

Wording notes, for the reviewer of this draft:

- The equivalence sentence and the dropped-class sentence of the current section are removed, and
  nothing replaces them with another coverage claim (REQ-005). Each step now states the commits it
  scans and the commits it does not (REQ-003, REQ-004).
- The merge-commit sentence is taken from git's documented behaviour for `log -p` without a
  `--diff-merges` option; this draft does not measure it. M3 measures it as supplementary evidence,
  and a contrary reading is a wording-only fix that returns to M2.
- The working-tree sentence keeps the working-tree step the current section already names. The
  section still gives no command for it; that is unchanged from the current text.
- The missing-tip detection command pair is new wording that operationalises the pinned
  `Missing-tip handling:` line; gate cell 2 measured the same two operations with the missing tip's
  name taken from a separate file. M3 exercises the worded pair as supplementary evidence.

### Allowlist representation (M2)

Choice: **a full-length digest of each listed value** — the git blob object name of the value
followed by one newline, as `git hash-object --no-filters` computes it. Rejected: **assembly of
each listed value from fragments at scan time.**

Reasons, in order of weight:

1. The fragment candidate cannot be drafted. `spec.md` §3.4 and `plan.md` §D forbid any SPEC
   artifact or evidence file to write a listed value whole or in fragments, and the lead's
   instruction for this segment forbids writing the value to the repository even as fragments.
   Under the fragment candidate the section itself carries the fragments, so the draft above could
   not be recorded here verbatim, and M4 could not copy what was drafted. The digest candidate
   writes no part of the value anywhere.
2. Exact-value matching is kept. A match is suppressed only when the blob object name of its exact
   bytes equals a listed digest, and the `-wF` match over the table rows compares whole
   40-character names, never a prefix. NEARCELL, which differs from the listed value in its last
   character, and MIXCELL, which carries a listed and an unlisted value on one line, stay findings
   by construction (probe below; M3 measures both).
3. No committed text matches the scan regex: the digest is 40 lowercase hexadecimal characters.
   The CI strict leak tier flags a 7-8 character run with a word boundary before it
   (`internal/template/internal_content_leak_test.go` line 373, pattern
   `\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`); a 40-character run offers no boundary inside it. AC-005 runs
   the strict tier itself at M4.
4. Tool availability: `git hash-object` ships with git, which the scan already requires, so the
   digest adds no platform-specific tool. The alternatives differ by platform — `shasum` (macOS),
   `sha256sum` (Linux, Git for Windows' shell), `certutil` or `Get-FileHash` (native Windows). The
   other commands (`printf`, `grep`, `sort`, `mkdir`, `split`, `find`, `paste`, `cut`, `rm`) are
   POSIX-shaped utilities; their presence on Linux and in Git for Windows' shell is inferred, not
   measured — every reading below was taken on macOS.
5. The form runs in a worktree-isolated session: every step is a single plain command, the git
   step reads its paths from standard input, and no step uses a command substitution, a loop, or a
   shell variable. The glob form of the git step was refused (probe below), which is why the draft
   lists the paths with `find` and `sort` first.

Digest derivation, taken 2026-09-11 in the session scratchpad `SP`, outside this repository. The
listed value is written below only as `<listed value>` and its fragments as `<fragment N>`; neither
appears in any file of this repository.

| Step | Command | Exit | Reading |
|---|---|---|---|
| assemble | `printf '%s%s%s\n' '<fragment 1>' '<fragment 2>' '<fragment 3>' > SP/m2probe/listed.txt` | 0 | `wc -c` 21; `/usr/bin/grep -cE -- 'REGEX'` → `1` |
| object format | `git rev-parse --show-object-format` (this worktree) | 0 | `sha1` |
| digest | `git hash-object --no-filters SP/m2probe/listed.txt > SP/m2probe/listed-digest.txt` | 0 | `05c61e935c693e4244743a05a1ba4ae33bd71d64` |
| independent recomputation | `printf 'blob 21\000' > SP/m2probe/hdr.bin`; `cat SP/m2probe/hdr.bin SP/m2probe/listed.txt > SP/m2probe/blob.bin`; `shasum -a 1 SP/m2probe/blob.bin` | 0 | the same 40 characters |

Mechanics probe of the suppression steps, taken before the draft was fixed, on a probe file of six
lines — five carrying a match (`LISTCELL <listed value>`, `NEARCELL <near value>`,
`OTHERCELL <unlisted value>`, `MIXCELL <listed value> <unlisted value>`, a `HEADCELL` PEM-style
header) and one context line — placed as `.moai/state/secrets-scan-output.txt` under a throwaway
repository `SP/m2probe/fx`. Every path under `.moai/state/` was written with the probe root in
front; `grep` ran as `/usr/bin/grep` (BSD grep 2.6.0-FreeBSD) and `find` as `/usr/bin/find`,
because both names resolve to shell functions in this session.

| Step | Exit | Reading |
|---|---|---|
| allowlist `printf` | 0 | one line |
| `grep -oE` | 0 | 6 lines |
| `sort -u` | 0 | 5 lines |
| `mkdir` | 0 | — |
| `split -a 6 -l 1` | 0 | 5 files, `v.aaaaaa` to `v.aaaaae` |
| glob form `git -C SP/m2probe/fx hash-object --no-filters <split dir>/v.* > …` | refused | guard: "this command redirects git through a glob pattern that expands at runtime. Refusing to run it" |
| `find` then `sort` | 0; 0 | 5 paths |
| `git -C SP/m2probe/fx hash-object --no-filters --stdin-paths < <paths file> > <digests file>` | 0 | 5 lines; the listed digest counted `1` by `/usr/bin/grep -cxF -f <allowlist file>` |
| `paste` | 0 | 5 rows |
| `grep -vwF -f <allowlist file>` | 0 | 4 rows |
| `cut -f2` | 0 | 4 values; the listed value counted `0` by `/usr/bin/grep -cxF -f SP/m2probe/listed.txt` |
| findings `grep -F -f` | 0 | `LISTCELL` 0, `NEARCELL` 1, `OTHERCELL` 1, `MIXCELL` 1, `HEADCELL` 1 |
| empty pattern file: `/usr/bin/grep -F -f <empty file> <probe file>` | 1 | 0 lines — an empty unsuppressed list reports nothing |
| `rm -r` | 0 | the split directory is gone (`test -e` exit 1) |

Gaps and residual risk of the representation:

- Only macOS was exercised (git 2.50.1, BSD grep 2.6.0-FreeBSD). Linux and Windows readings are
  inferred.
- A repository using the SHA-256 object format computes other digests, so nothing is suppressed
  there and the listed value is reported (inferred from git's object-format rule, not measured).
  The failure direction is over-reporting, never a silent suppression.
- Suppressing an unlisted value would need a credential-shaped value whose blob object name equals
  the listed digest — a second preimage of SHA-1. No practical second-preimage attack on SHA-1 is
  known; the known attacks are collisions between two chosen inputs.
- The steps rely on `find` and `sort` ordering the split files in the order `split` wrote them.
  The names differ only in their last six lowercase letters, so the order is the same in any
  locale; M3 measures it on the fixture.

### M2 pre-commit readings

Taken 2026-09-11 on the working tree on top of `023a25e7a`, with every edit above in place and this
subsection not yet written. `SP/draft-section.txt` is the draft extracted by a `sed -n` range from
the `#### Secrets Scan` line through the closing four-backtick fence of PROG (60 lines, the fence
included);
`SP/draft-cmds.txt` is its `log -p` lines (`/usr/bin/grep -e 'log -p'`, exit 0, 2 lines). The pinned
lines were extracted with the AC-016 `sed` form from `git show 68c56be0d:PROG` and from the working
tree. `REGEX` is the scan regex as `review.md` writes it.

| Check | Command (outline) | Reading |
|---|---|---|
| pinned block unchanged | AC-016 `sed` extraction of the pin commit and of the working tree, then `cmp` | 28 and 28 lines; `cmp` exit 0 |
| AC-016 (a) | `/usr/bin/grep -cF -f <pinned line file> SP/draft-section.txt` for tip recording, scan command, missing-tip handling | `1`, `1`, `1` (each pattern file 1 non-empty line) |
| pinned full-history scan and uncovered sentence | the same form for `Full-history scan:` and `Uncovered commits:` | `2` (both command lines carry it), `1` |
| AC-006 | `grep -vc -e '--all'` and `grep -c -e '--stdin'` over the commands; `grep -cF '^%(objectname)'`, `grep -ci 'every ref'`, `grep -c 'the HEAD SHA of the last completed scan'` over the draft | `0`; `1`; `1`; `2`; `0` |
| AC-015 | `/usr/bin/grep -cE -- 'REGEX'` over the draft; control `SP/ctl-pem.txt` | `0`; `1` |
| AC-002 | the exact equivalence phrase, `no finding class is dropped`, `grep -ci 'same coverage'` | `0`, `0`, `0` |
| AC-005 | `SPEC-`, `t629`, the four cost figures, the 16-name language grep (`-ciwE`), `R language` | `0`, `0`, `0`, `0`, `0` |
| 7-8 hex run | `/usr/bin/grep -cE '(^\|[^0-9A-Za-z_])[0-9a-f]{7,8}([[:space:].,;:!?]\|$)'` over the draft; control line `see abcdef1 here` | `0`; `1` |
| language control | the 16-name grep over a control line `written in go today` | `1` |
| full-length digest | `/usr/bin/grep -cE '(^\|[^0-9A-Za-z_])[0-9a-f]{40}([^0-9a-f]\|$)'` over the draft | `1` |
| date shape | `/usr/bin/grep -cE '202[5-9]-[0-1][0-9]-[0-3][0-9]'` over the draft | `0` |
| credential regex, SPEC files | `/usr/bin/grep -cE -- 'REGEX'` over `spec.md`, `plan.md`, `acceptance.md`, `progress.md` | `0` each, exit 1 |
| gate verdict lines | `/usr/bin/grep -c '^verdict: trustworthy$' PROG` | `3` |
| ordering-anchor strings | `/usr/bin/grep -c` for the five pickaxe strings AC-007, AC-010, AC-011, AC-012 read, over PROG and over `git show HEAD:PROG` | `6` and `6` — this commit adds none |
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | exit 0, `✓ No findings — all SPEC documents are valid` |
| scope | `git diff --stat`; `git diff --stat feeecc980 -- <both review.md paths>` | `progress.md` only; nothing |

Tool provenance (`verification-claim-integrity.md` §2.2): the lint ran on the installed
`v3.2.0-rc.7` build, commit `ed71054d3` with a dirty tree; `git merge-base --is-ancestor ed71054d3
HEAD` exited 1. It is not a build made from this tree.

### Fixture proof (M3)

Taken 2026-09-11 from this worktree on branch `WT-secret-scan-refs`, HEAD `d22b4466b` (the M2
commit), on two fresh throwaway fixture repositories outside this repository, `SP/m3/fx1` and
`SP/m3/fx2`, where `SP` is the session scratchpad. `git version 2.50.1 (Apple Git-155)`. Each
fixture: `git init -b main`, a fixture `user.name` and `user.email`, `commit.gpgsign=false`,
`core.hooksPath=/dev/null`, read back with `git config --list --local`. `REGEX` is the scan regex
as `review.md` writes it.

How the worded procedure was run. Every command is the draft's command (§ Draft section wording
(M2)) with three mechanical substitutions and no other change: `git` runs as `git -C <fixture
root>`; a `.moai/state/` path in a git command's redirect is written with the fixture root in front,
because the shell resolves a redirect; and the draft's `grep` and `find` run as `/usr/bin/grep` (BSD
grep 2.6.0-FreeBSD) and `/usr/bin/find`, because both names resolve to shell functions in this
session. The non-git steps ran from the fixture root with the draft's relative paths. Each review
ran the whole draft sequence: tip recording, the which-step check (`test -e` on the store), the
scan, the store replacement with `mv` after the scan exited 0, then all thirteen suppression
commands. A `grep` exit of 1 is read as an empty result, per the draft; no command exited 2.

Markers and values: PEM-style private-key header lines with distinct uppercase labels, and
access-key-shaped values, each written by `printf` from fragments held only in `SP`. This record
writes a label as `<LABEL>`, the listed value as `<listed value>`, and other values as
`<near value>` and `<unlisted value>`; no value, fragment, fixture object name, or scan output line
appears here. Every marker file was checked with `/usr/bin/grep -cE -- 'REGEX'` → `1` per marker
line.

#### AC-001 review sequence

Fixture `SP/m3/fx1`; outputs in `SP/m3/`.

| Step | Command (outline) | Exit | Reading |
|---|---|---|---|
| C0 | clean `keys.txt` committed on `main` | 0 | — |
| cell B | `git log -p --all -G 'REGEX' > b.txt 2> b.err` | 0 | `wc -c` 0 and 0 |
| R1 tip recording | `git for-each-ref --format='^%(objectname)' > <fx>/.moai/state/secrets-scan-tips.next` | 0 | 1 line |
| R1 which step | `test -e <store>` | 1 | no store, so the full-history scan runs |
| R1 scan | the draft's full-history scan line | 0 | 0 bytes; `mv` to the store, exit 0 |
| R1 suppression | the thirteen draft commands | `grep -oE` 1, `grep -vwF` 1, findings `grep -F` 1, others 0 | findings 0 bytes |
| S1 | branch `side` from C0 (`switch -c`); `SIDECELL` marker in `side.txt`, committed | 0 | `side.txt` absent on `main` (`test -e` exit 1) |
| C1 | `HEADCELL` marker appended to `keys.txt` on `main`, committed | 0 | — |
| construction (step 3) | `git merge-base --is-ancestor side HEAD` | 1 | — |
| R2 tip recording, which step | as R1; `test -e <store>` | 0; 0 | the store read copied to `r2-tips-read.txt` (1 line) |
| R2 scan | the draft's scan command line, with its `2>` redirect | 0 | output 624 bytes, error 0 bytes; `mv` exit 0 (store 2 lines) |
| R2 suppression | the thirteen draft commands | all 0 | matches 2, distinct 2, paths 2, digests 2, kept 2, findings 2 lines |
| C2 | clean `notes.txt` committed on `main` | 0 | — |
| S2 | branch `side2` from C2; `LATECELL` marker in `late.txt`, committed | 0 | — |
| construction (step 5) | `git merge-base --is-ancestor side2 HEAD` | 1 | — |
| R3 | as R2 | scan 0; suppression all 0 | output 301 bytes, error 0 bytes; the store read 2 lines; distinct 1; findings 1 line; store 3 lines after `mv` |
| R4 tip recording | as R1 | 0 | `cmp <store> <store>.next` exit 0 — no ref moved since R3 |
| R4 | as R2 | scan 0; `grep -oE` 1, `grep -vwF` 1, `grep -F` 1, others 0 | output 0 bytes, error 0 bytes; findings 0 bytes |
| construction (step 8), immediately before the counts | `git merge-base --is-ancestor side HEAD`; the same for `side2` | 1; 1 | no merge of `side` or `side2` at any point |
| counts | `/usr/bin/grep -c '<LABEL>'` over `r2.txt` (HEADCELL, SIDECELL) and `r3.txt` (LATECELL, SIDECELL); `wc -c` over `b.txt` and `r4.txt` | 0 each | see below |

| Cell | Required | Reading | Holds |
|---|---|---|---|
| B | 0 bytes | 0 bytes | yes |
| H | R2 `HEADCELL` ≥ 1 | `1` | yes |
| S | R2 `SIDECELL` ≥ 1 | `1` | yes |
| T | R3 `LATECELL` ≥ 1 | `1` | yes |
| N | R4 0 bytes | 0 bytes | yes |

Informational: R3 `SIDECELL` `0` — the side branch's marker was reported once, in R2, and not again
(REQ-002 does not require re-reporting). Context: R2 carries 2 lines matching `REGEX` and 2
`^commit ` lines; R3 carries 1 and 1. Findings files: R2 `HEADCELL` 1 and `SIDECELL` 1; R3
`LATECELL` 1.

Predicate (`acceptance.md` §D.1): every cell of the required table holds, so none of the fail
conditions — their exact complement — occurs. Each marker was reported in the scan where its commit
first became reachable. Construction: every `is-ancestor` reading is exit 1, and R1-R4 each exited
0, so no construction gap arose and no procedure-caused non-zero exit occurred (run caution N1).

Result: **PASS**.

#### AC-013 and AC-014 allowlist controls

Fixture `SP/m3/fx2`; outputs in `SP/m3/`. The listed value is the value whose digest the draft
lists.

| Step | Command (outline) | Exit | Reading |
|---|---|---|---|
| C0 | clean `keys.txt` committed on `main`; `git log -p --all -G 'REGEX' > al-base.txt` | 0; 0 | 0 bytes |
| R1 | the whole draft sequence (store absent, so the full-history scan) | tip recording 0, `test -e` 1, scan 0, `mv` 0; suppression: `grep -oE` 1, `grep -vwF` 1, `grep -F` 1, others 0 | output 0 bytes |
| S1 | branch `side` from C0; `SIDECELL` marker in `side.txt`, committed | 0 | — |
| C1 | new file `list.txt` on `main`, five adjacent lines: `LISTCELL <listed value>`, `NEARCELL <near value>` (the listed value with its last character changed), `OTHERCELL <unlisted value>`, `MIXCELL <listed value> <unlisted value>`, the `HEADCELL` marker; committed | 0 | `/usr/bin/grep -cE -- 'REGEX'` over the file → `5` |
| construction | `git show HEAD --output=al-commit.txt`; `/usr/bin/grep -c '^@@'`; `/usr/bin/grep -c` for `LISTCELL`, `NEARCELL`, `OTHERCELL`, `MIXCELL` | 0 | `^@@` `1`; each label `1` (`HEADCELL` `1`; `^diff ` `1`) |
| construction | `git merge-base --is-ancestor side HEAD` | 1 | — |
| R2 tip recording, which step | as R1; `test -e <store>` | 0; 0 | — |
| R2 scan | the draft's scan command line | 0 | output 750 bytes, error 0 bytes; `mv` exit 0 |
| R2 suppression | the thirteen draft commands | all 0 | matches 7, distinct 6, paths 6, digests 6, table 6, kept 5, unsuppressed 5, findings 5 lines |
| raw and findings | `cp <fx>/.moai/state/secrets-scan-output.txt al-raw.txt`; `cp <fx>/.moai/state/secrets-scan-findings.txt al-findings.txt` | 0 | 31 lines; 5 lines |
| REGEX extracts | `/usr/bin/grep -E -- 'REGEX' al-raw.txt > al-raw-match.txt`; the same over `al-findings.txt` into `al-match-lines.txt` | 0; 0 | 6 lines; 5 lines |

Supplementary, taken before the draft's `rm -r` step: `xargs cat` over the sorted path list
reproduced `secrets-scan-distinct.txt` byte for byte (`cmp` exit 0), so `find` and `sort` ordered
the split files as `split` wrote them; `cut -f2` of the table equals the distinct list (`cmp` exit
0); exactly one table row carries the listed digest (`/usr/bin/grep -cwF` → `1`), and that row's value
equals the listed value (`cmp` exit 0); the listed value appears `0` times in the unsuppressed list
(`/usr/bin/grep -cxF`).

| Label | `al-raw-match.txt` | `al-match-lines.txt` |
|---|---|---|
| `LISTCELL` | `1` | `0` |
| `NEARCELL` | `1` | `1` |
| `OTHERCELL` | `1` | `1` |
| `MIXCELL` | `1` | `1` |
| `HEADCELL` | `1` | `1` |
| `SIDECELL` | `1` | `1` |

AC-013 predicate (`acceptance.md` §D.13): the raw count of `LISTCELL` is ≥ 1 (`1`), so the listed
value matched the regex and suppression was exercised; its findings count is `0`. Result: **PASS**.

AC-014 predicate (`acceptance.md` §D.14): construction holds — one hunk and each of the four labels
counted `1`; every negative label — `NEARCELL`, `OTHERCELL`, `MIXCELL`, `HEADCELL`, `SIDECELL` —
counts ≥ 1 in both extracts. Result: **PASS**.

#### Supplementary checks of new wording

Both checks exercise draft wording outside the three criteria above; neither is a criterion.

Missing-tip detection pair, on `SP/m3/fx2` after the allowlist run:

| Step | Command (outline) | Exit | Reading |
|---|---|---|---|
| branch `gone` | a plain `gone.txt` committed on a new branch `gone`; its tip written by `git rev-parse refs/heads/gone > mt-gone.txt` | 0 | — |
| R3 | tip recording; the scan command; `mv` | 0; 0; 0 | output 0 bytes; the stripped store holds the gone tip (`/usr/bin/grep -cxF -f mt-gone.txt` → `1`) |
| remove | `git branch -D gone`; `git reflog expire --expire=now --all`; `git gc --prune=now --quiet` | 0; 0; 0 | — |
| construction | `git cat-file --batch-check < mt-gone.txt` | 0 | one line ending ` missing` (`/usr/bin/grep -c` → `1`) |
| L1 | `GONECELL` marker in `after.txt`, committed on a new branch `after` | 0 | — |
| R4 | tip recording; the scan command | 0; 128 | output 0 bytes; error 59 bytes |
| worded detection | the draft's `sed 's/^\^//'` line, then its `grep -F -f` line with output to `mt-detect.txt` | 0; 0 | 1 line; `bad object` `1`; the gone tip's name `1` (`/usr/bin/grep -cF -f mt-gone.txt`) |
| handling | the draft's full-history scan in place of R4's scan; `mv` | 0; 0 | output 1055 bytes; `GONECELL` `1` |

Merge-commit sentence, on the same fixture: a plain `mx.txt` committed on a new branch `mx`; on
`main`, `git merge --no-ff --no-commit mx` (exit 0), the `MERGECELL` marker appended to `mx.txt`,
staged, and committed as the merge; `git rev-list --parents -n 1 HEAD` → three fields, so the commit
has two parents and only the merge's own change carries the marker.

| Scan | Exit | `MERGECELL` |
|---|---|---|
| the draft's scan command, reading the store R4 recorded | 0 | `0` (0 bytes) |
| the draft's full-history scan | 0 | `0` (1055 bytes) |
| control: the full-history scan with `--diff-merges=first-parent` added | 0 | `1` |
| control: the working-tree file `mx.txt` | — | `1` |

The two zero counts are not an empty-command artefact: the same history read with first-parent
merge diffs shows the marker. The sentence's example, a conflict resolution, was not built; the
merge here carried an added line without a conflict.

#### Gaps and residual risk in M3

- One machine, one git build, macOS only; the draft's `grep` and `find` ran as the system binaries,
  not as the session's shell functions. Linux and Windows are not exercised.
- Only the PEM-header and access-key alternatives of `REGEX` were exercised; the token alternative
  was not.
- Only local branches were exercised; tags, remote-tracking refs, and stash entries were not. The
  explicit full-scan flag, the working-tree step (the draft gives no command for it), a commit that
  lands between the tip recording and the scan, and a repository using the SHA-256 object format
  were not exercised.
- AC-016's gap D19a stands: the pinned lines appearing in the section does not show that the
  section prescribes no other timing for tip recording.
- The fixture outputs stay in `SP`, a machine-local scratchpad; they carry credential-shaped lines
  and are not exported. The counts, exit codes, and sizes above are the record; the raw files are a
  known loss once the scratchpad is cleared.

### M4 edit

Taken 2026-09-11 from this worktree on branch `WT-secret-scan-refs`. The edit commit `R` is
`cc40925131937a78cb8573bace47349ddb594bcf`, parent `e8128c708` (the M3 commit); it touches the two
review workflow copies and `internal/template/catalog.yaml` only. `SP` is the session scratchpad,
outside this repository. `REGEX` is the scan regex as `review.md` writes it. `SECTION` is
`SP/section.txt`, extracted from the edited template copy with the `acceptance.md` `sed -n` range
(61 lines: the 59 section lines, a blank line, and the next heading).

Operator decisions relayed by the lead, quoted verbatim as received on 2026-09-11:

> [리드] t629 운영자 결정: ① 허용 목록 표현 = A(SHA-1 git 블롭 이름 그대로 M4 진행). 누출 검사 오인(3)은 M4의 AC-005 strict 결과로 판정하고, 걸리면 그 자리에서 멈춰 보고하세요. ② 13단계 억제 절차 = 측정된 형태 유지(rm -r 고정 경로 확인 인정). 복잡도는 후속 개선 후보로 verdict에 남겨 주세요. M4 진행하세요. push 금지 유지.

Summary in English: the operator keeps allowlist representation A (the full-length SHA-1 git blob
object name), judges the leak-check misfire concern by the AC-005 strict run below (stop on a hit),
and keeps the thirteen-step suppression procedure in its measured form, with the `rm -r` fixed-path
step accepted.

Follow-up candidate: the thirteen-step suppression procedure could be simplified — for example, a
single-digest comparison while the list holds one value. It is kept in its measured form by operator
decision.

How the edit was made. The template copy was edited first: its secrets-scan section was replaced
with the text of § Draft section wording (M2) inside the four-backtick fence; nothing outside the
section changed. The local copy then received the same edit, not a copy of the file.

| Check | Command (outline) | Exit | Reading |
|---|---|---|---|
| pre-edit identity | `diff -q LOC TPL` | 0 | no output |
| load before | `ps -axo pid,etime,command > SP/m4-ps.txt`; `/usr/bin/grep -c` for running `go test` and `git log -p` processes; `uptime` | 0 | `0`, `0`; load averages 9.17 11.78 14.02 |
| section equals the draft | draft extracted with `sed -n '514,572p' PROG` before the edit (59 lines); the first 59 lines of `SECTION` compared with `cmp` | 0 | byte-identical; lines 60-61 are the blank line and `#### Data Isolation Check` |
| AC-003 (a) | `cmp LOC TPL`; `diff -q LOC TPL` | 0; 0 | no output |
| AC-003 (b) | `git diff --stat feeecc980 cc4092513 -- LOC TPL` | 0 | both files named, 53 lines changed each (2 files, 94 insertions, 12 deletions) |
| AC-002 | the exact equivalence phrase, `no finding class is dropped`, and `grep -ci 'same coverage'`, each over `LOC TPL` | 1 each | `0` and `0` for all three |
| AC-005 greps | `SPEC-`, `t629`, the four cost figures over `TPL`; the 16-name language grep (`-ciwE`) and `R language` over `SECTION` | 1 each | `0`, `0`, `0`, `0`, `0` |
| AC-005 strict leak | `MOAI_TEMPLATE_LEAK_STRICT=1 go test -v ./internal/template/ -run '^TestTemplateNoInternalContentLeak$' -count=1 > SP/leak.txt 2>&1` | 0 | `SP/leak.txt` 4 lines; `--- PASS: TestTemplateNoInternalContentLeak` count `1`; `--- FAIL` count `0` |
| AC-006 | `/usr/bin/grep -e 'log -p' SECTION > SP/cmds.txt` (2 lines); `grep -vc -e '--all'` and `grep -c -e '--stdin'` over the commands; `grep -cF '^%(objectname)'`, `grep -ci 'every ref'`, `grep -c 'the HEAD SHA of the last completed scan'` over `SECTION` | 0; 1; 0; 0; 0; 1 | `0`; `1`; `1`; `2`; `0` |
| AC-006 pinned sentence | the `Uncovered commits: ` line of the pinned block, prefix removed, as a `grep -cF -f` pattern file over `SECTION` | 0 | `1` |
| AC-015 | `/usr/bin/grep -cE -- 'REGEX' LOC TPL`; control: the same command over `SP/m4-ctl-pem.txt`, one PEM-style header line assembled by `printf` from fragments | 1; 0 | `0` and `0`; control `1` |
| AC-016 (a) | the `Tip recording: `, `Scan command: `, and `Missing-tip handling: ` lines of the pinned block, each prefix removed with `sed -n 's/^<prefix>//p'` into its own pattern file, then `/usr/bin/grep -cF -f <pattern file> SECTION` | 0 each | each pattern file 1 line, 1 non-empty; counts `1`, `1`, `1` |
| AC-016 (b), `G` | `git log --format=%H -S` for the gate-evidence heading, `feeecc980..cc4092513~1 -- PROG` | 0 | 3 lines; newest `f167a9cd8`; the heading count at `R~1` is `1` |
| AC-016 (b), block | the AC-016 `sed -nE` extraction from `git show f167a9cd8:PROG` and from `git show cc4092513~1:PROG`, then `cmp` | 0 | 28 and 28 lines; `cmp` exit 0 |
| pin unchanged since the pin commit | the same extraction from `git show 68c56be0d:PROG` and from the working tree before this subsection was written, then `cmp` | 0 | 28 and 28 lines; `cmp` exit 0 |
| AC-007 dry | `git merge-base --is-ancestor af7eb142b cc4092513` | 0 | `D` ≠ `R` |
| AC-011 dry | `git log --reverse --format=%H feeecc980..cc4092513 -- TPL` and the same for `LOC`; `git merge-base --is-ancestor f167a9cd8 cc4092513`; `git show cc4092513~1:PROG` then `/usr/bin/grep -c` for trustworthy verdict lines | 0; 0; 0 | first line `cc4092513` for both copies, so `R` is the first copy-touching commit; `G` ≠ `R`; count `3` |
| AC-004 reading at `R` | `git diff feeecc980 cc4092513 --output=SP/m4-card-diff.txt`; `wc -l`; the AC-004 added-line grep | 0; 0; 1 | 3882 lines; count `0` |
| credential regex, SPEC files | `/usr/bin/grep -cE -- 'REGEX'` over `spec.md`, `plan.md`, `acceptance.md`, `progress.md` | 1 | `0` each |
| lint | `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` | 0 | `✓ No findings — all SPEC documents are valid` |

Catalog. The template tree changed, so the hash of catalog entry `moai` (path
`templates/.claude/skills/moai/`, a whole-tree hash) went stale; only that entry was regenerated.

| Step | Command | Exit | Reading |
|---|---|---|---|
| before the edit | `go test ./internal/template/ -run 'TestCatalogHashCoversSkillSubfiles\|TestManifestHashFormat' -count=1 -v` | 0 | 2 PASS lines |
| after the edit, before regeneration | the same command | 1 | both FAIL: `CATALOG_HASH_UNSTABLE` and `CATALOG_HASH_SKINNY` for `moai`, stored `1d23838d…`, computed `21cf422c…` |
| dry run | `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai --dry-run` | 0 | `moai: 21cf422cb335b651a2d646049d4bc4b1e6d5d13270a7cf1c6261c37727c48983`, equal to the value the failing tests computed |
| regeneration | the same command without `--dry-run` | 0 | `catalog.yaml updated successfully (12899 bytes)` |
| scope | `git diff --numstat -- internal/template/catalog.yaml`; changed lines of the diff | 0 | `1 1`; one removed and one added `hash:` line of entry `moai`, nothing else |
| after regeneration | the two catalog tests by name, `-v` | 0 | exactly 2 PASS lines |
| package tree | `go test ./internal/template/... -count=1` | 0 | `ok` for `internal/template`, `agentemit`, `commandemit`; `scripts` has no test files; no `FAIL` |
| emitters | `go test ./internal/template/commandemit/... ./internal/template/agentemit/... -count=1` | 0 | `ok` for both; no emitter publishes the edited file, and nothing was regenerated |

Load during the package-tree run: 8.97 10.50 12.96 before, 8.44 10.22 12.77 after (`uptime`).

Tool provenance (`verification-claim-integrity.md` §2.2): the lint ran on the installed
`v3.2.0-rc.7` build, commit `ed71054d3` with a dirty tree; `git merge-base --is-ancestor ed71054d3
HEAD` exited 1. It is not a build made from this tree. The `go test` and `go run` readings were
compiled from this tree.

Gaps and residual risk in M4:

- The strict leak run passed with no failing class, but this run carries no positive control of its
  own; the 7-8 hex control of § M2 pre-commit readings is the evidence that the pattern detects a
  short hex run.
- AC-004, AC-007, AC-011, and AC-016 above were read at `R`, not at `K`; M5 records `K` and re-runs
  them there. Commits after `R`, including this one, are outside the AC-004 reading.
- `make build` and the embed check were not run; they belong to the lead at batch close.
- The fixture evidence of M3 and the scan outputs stay in `SP` and are not exported.

### Closure checks (M5)

Taken 2026-09-11 from this worktree on branch `WT-secret-scan-refs`, clean tree, before the develop
absorb. `SP` is the session scratchpad, outside this repository; every `git` output below went to a
file under `SP` and every exit code was read with `; echo "exit=$?"`, never through a pipe.

**K.** `git rev-parse HEAD`, read at the start of this segment →
`c10626ab72901b20e7ff89d4f0b0a3461df51727` (the M4 evidence commit; parent `cc4092513`, the edit
commit `R`). Every range and revision below ends at `K`, never at `HEAD` (`acceptance.md`
§ Closure-check anchor). `git diff --stat cc4092513 K` names `progress.md` only (80 insertions), so
the copies at `K` are the copies the M4 readings measured at `R`.

| AC | Command (outline, `K` as above) | Exit | Reading |
|---|---|---|---|
| AC-003 (a) | `diff -q LOC TPL` | 0 | no output |
| AC-003 (b) | `git diff --stat feeecc980 K -- LOC TPL > SP/m5-ac003-stat.txt` | 0 | both files named, 53 lines changed each; `2 files changed, 94 insertions(+), 12 deletions(-)` |
| AC-004 | `git diff feeecc980 K --output=SP/card-diff-k.txt`; `wc -l`; the AC-004 added-line grep over that file | 0; 0; 1 | `3962` lines; count `0`. Control: the same grep over `SP/m5-ctl-diffline.txt`, one `+`-prefixed PEM-style header line assembled by `printf` from fragments → `1`, exit 0 |
| AC-007, `D` | `git log --format=%H -S 'Decision:** Option' -- spec.md > SP/m5-ac007-d.txt` | 0 | 1 line, `af7eb142b` |
| AC-007, `R` | `git log --reverse --format=%H feeecc980..K -- TPL > SP/r-tpl.txt`; the same for `LOC` into `SP/r-loc.txt` | 0; 0 | first line `cc4092513` in both (1 line each) |
| AC-007 | `git merge-base --is-ancestor af7eb142b cc4092513` | 0 | `D` ≠ `R`; `LOC` resolves to the same `R`, so the same check covers it |
| AC-010, ordering | `git log --format=%H -S '### Lead approval for gate cell 3' feeecc980..K -- PROG > SP/g3-approval.txt`; `git log --format=%H -S '#### Gate cell 3' feeecc980..K -- PROG > SP/g3-evidence.txt` | 0; 0 | 1 line each: `P` = `7e06766c8`, `E3` = `023a25e7a` |
| AC-010, ordering | `git merge-base --is-ancestor 7e06766c8 023a25e7a` | 0 | `P` ≠ `E3` |
| AC-011, `G` | `git log --reverse --format=%H -S '### Gate evidence' feeecc980..K -- PROG > SP/g-commit.txt` | 0 | 3 lines; first `6e56840d5` (the round 1 evidence commit), then `e298f7336` (round 1 moved out), then `f167a9cd8` (round 2 evidence) |
| AC-011 | `git merge-base --is-ancestor 6e56840d5 cc4092513` | 0 | `G` ≠ `R`; `LOC` resolves to the same `R`. Supplementary: `git merge-base --is-ancestor f167a9cd8 cc4092513` → exit 0, so the standing round 2 evidence also precedes the edit |
| AC-011, verdicts | `git show cc4092513~1:PROG > SP/pre-edit-progress.txt`; `/usr/bin/grep -c '^verdict: trustworthy$'` | 0; 0 | `3` |
| AC-016, `G` | `git log --format=%H -S '### Gate evidence' feeecc980..cc4092513~1 -- PROG > SP/g-rounds.txt` | 0 | 3 lines; newest `f167a9cd8`. It differs from AC-011's `G` because a second gate round ran (`plan.md` M3); the heading count at `R~1` is `1` (`/usr/bin/grep -c '^### Gate evidence$' SP/prog-r1.txt`, exit 0) |
| AC-016 (b) | `git show f167a9cd8:PROG > SP/prog-g.txt`; `git show cc4092513~1:PROG > SP/prog-r1.txt`; the AC-016 `sed -nE` extraction into `SP/pin-g.txt` and `SP/pin-r1.txt`; `cmp` | 0 each | 28 and 28 lines; `cmp` exit 0 |
| pin unchanged since the pin commit | the same extraction from `git show 68c56be0d:PROG` into `SP/pin-68c.txt`, then `cmp` against `SP/pin-r1.txt` | 0 | 28 lines; `cmp` exit 0 |
| AC-016 (a) | `git show K:TPL > SP/tpl-k.txt`; `SECTION` extracted from it with the `acceptance.md` `sed -n` range into `SP/section-k.txt`; the three prefixed pin lines of `SP/pin-r1.txt` into `SP/pin-tips.txt`, `SP/pin-scan.txt`, `SP/pin-gone.txt`; `/usr/bin/grep -cF -f <pattern file> SP/section-k.txt` | 0 each | `SECTION` 61 lines; each pattern file 1 line, 1 non-empty; counts `1`, `1`, `1` |

AC-012 is not applicable: `git show K:PROG > SP/prog-k.txt`, then `/usr/bin/grep -c
'^verdict: trustworthy$'` → `3` (exit 0), at lines 270 (`#### Gate cell 1`), 312 (`#### Gate cell 2`),
and 452 (`#### Gate cell 3`); `/usr/bin/grep -c 'verdict: untrustworthy'` → `0` (exit 1); and
`/usr/bin/grep -c 'Run phase stopped'` → `0` (exit 1). No gate cell was untrustworthy, so no stop
occurred.

Run caution N2 applies to AC-004: its range ends at `K`, so this commit and any later commit are
outside it.

Consolidated AC matrix (`acceptance.md` §D.19):

| AC | Status | Evidence in this file | Measured at |
|---|---|---|---|
| AC-001 | PASS | § Fixture proof (M3) › AC-001 review sequence | tree `d22b4466b` (draft wording), fixture `SP/m3/fx1`; the draft equals the first 59 lines of the edited section (§ M4 edit, `cmp` exit 0) |
| AC-002 | PASS | § M4 edit | edited copies at `R` (`cc4092513`), unchanged to `K` |
| AC-003 | PASS | § Closure checks (M5) | `K` |
| AC-004 | PASS | § Closure checks (M5) | `feeecc980..K` |
| AC-005 | PASS | § M4 edit (greps and the strict leak run) | `R`, unchanged to `K` |
| AC-006 | PASS | § M4 edit | `R`, unchanged to `K` |
| AC-007 | PASS | § Closure checks (M5) | `K` |
| AC-008 | PASS | § Gate evidence › Gate cell 1 (`verdict: trustworthy`) | tree `68c56be0d`, fixture `SP/m1gate3/fx` |
| AC-009 | PASS | § Gate evidence › Gate cell 2 (`verdict: trustworthy`) | tree `68c56be0d`, fixture `SP/m1gate3/fx` |
| AC-010 | PASS | § Gate evidence › Gate cell 3 (`verdict: trustworthy`); approval ordering in § Closure checks (M5) | timing on this repository at `7e06766c8`; ordering at `K` |
| AC-011 | PASS | § Closure checks (M5) | `K` |
| AC-012 | not applicable | § Closure checks (M5), three `verdict: trustworthy` lines | `K` |
| AC-013 | PASS | § Fixture proof (M3) › AC-013 and AC-014 allowlist controls | tree `d22b4466b`, fixture `SP/m3/fx2` |
| AC-014 | PASS | § Fixture proof (M3) › AC-013 and AC-014 allowlist controls | tree `d22b4466b`, fixture `SP/m3/fx2` |
| AC-015 | PASS | § M4 edit | `R`, unchanged to `K` |
| AC-016 | PASS | § Closure checks (M5) | `K` |

Summary: 15 PASS, AC-012 not applicable, 0 FAIL.

Standing gaps:

- AC-010 names `/usr/bin/time -p` and its `real` field; gate cell 3 used the shell reserved word `time`
  and its `total` field, an operator-approved measurement-tool difference. Whether AC-010's wording
  changes is left for the sync phase.
- Gate cell 3 rests on one machine, one git build (`git version 2.50.1 (Apple Git-155)`), and one pair
  of runs under contention.
- In gate cell 3, the two commits counted by `B − A` became reachable between the first run's tip
  recording and its scope listing; whether the first scan itself covered them was not observed.
- `moai spec lint` is judged by the installed `v3.2.0-rc.7` build (`moai version` → `v3.2.0-rc.7
  moai_cp/20260910_130400-275-ged71054d3-dirty`); `git merge-base --is-ancestor ed71054d3 K` exited
  1, so it is not a build made from this tree (`verification-claim-integrity.md` §2.2).
- AC-004 covers commits up to `K` only (run caution N2).
- The thirteen-step suppression procedure is kept in its measured form by operator decision; its
  simplification is a follow-up candidate.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-09-11
k_sha: c10626ab72901b20e7ff89d4f0b0a3461df51727
run_commit_sha: pending-backfill   # the M5 commit, after K, cannot cite its own hash
ac_pass_count: 15
ac_fail_count: 0
ac_not_applicable: [AC-012]
cross_platform_build: not measured   # make build and the embed check belong to the lead at batch close
m1_to_mN_commit_strategy: one commit per step, stacked on the card branch, unpushed
evidence: "§E.2 › Closure checks (M5)"
```

Run-phase commits, `f16f7c095` through `K`:

- `f16f7c095` docs(t629): record the Phase 4 mode selection and kickoff approval before run (card t629)
- `b8c0ef74b` docs(t629): pin the Option 2 procedure before the measure-first gate (card t629)
- `6e56840d5` docs(t629): record measure-first gate cells 1 and 2 (card t629)
- `9dab82a4a` docs(t629): amend the SPEC to a caret-prefixed tip store read through stdin (card t629)
- `2b1808efe` docs(t629): add scoped amendment audit report, iteration 4 PASS 0.87 (card t629)
- `e298f7336` docs(t629): move superseded gate round 1 out of the progress record (card t629)
- `68c56be0d` docs(t629): pin the caret-store stdin procedure for gate round 2 (card t629)
- `f167a9cd8` docs(t629): record gate round 2 cells 1 and 2 on the stdin pin (card t629)
- `7e06766c8` docs(t629): quote the lead approval for gate cell 3 before any cell 3 command (card t629)
- `023a25e7a` docs(t629): record gate round 2 cell 3 timing on this repository (card t629)
- `d22b4466b` docs(t629): draft the secrets-scan section wording and allowlist representation (card t629)
- `e8128c708` docs(t629): record the fixture proof of the worded secrets-scan procedure (card t629)
- `cc4092513` docs(t629): replace the secrets-scan section with the tip-store procedure and exact allowlist (card t629)
- `c10626ab7` docs(t629): record the M4 edit readings (card t629)

The M5 commit that records this signal follows `K` and is outside the list and the AC-004 range.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-09-11
sync_base_sha: 9a9cedc3d   # the M5 closure-check commit; parent K = c10626ab7
sync_commit_sha: pending-backfill   # the sync commit cannot cite its own hash
changelog_entry_position: "CHANGELOG.md [Unreleased] › ### Fixed, first entry"
b12_self_test_a: "grep -c 'SPEC-REVIEW-SECRET-SCAN-REFS-001' CHANGELOG.md before emission → 0 (exit 1)"
b12_self_test_b: "distinct AC IDs in acceptance.md → 16; the entry states 16 (15 PASS, AC-012 not applicable)"
b12_self_test_c: "every path the entry cites resolved with ls before the commit"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (3-phase close; implemented merged into this commit)"
  updated: "2026-09-11 (already that date; unchanged)"
  plan.md / acceptance.md / progress.md: "no frontmatter block; nothing to transition"
mx_tags: not applicable   # no code change; the change surface is two markdown copies and a catalog hash line
docs_site: stale, not edited in this segment   # see follow-ups below
```

**docs-site finding.** The `/moai review` page describes the former single-SHA checkpoint
(`.moai/state/secrets-scan-checkpoint.txt`, range up to the current HEAD) in all four locales, at
line 77 of each:

- `docs-site/content/ko/utility-commands/moai-review.md`
- `docs-site/content/en/utility-commands/moai-review.md`
- `docs-site/content/ja/utility-commands/moai-review.md`
- `docs-site/content/zh/utility-commands/moai-review.md`

The 4-locale same-change obligation makes the update a separate step; it is not part of this
commit.

**Handed to the lead** (not decided in the sync phase):

- (a) AC-010's wording names `real`, while the operator-approved measurement for gate cell 3 used
  the shell `time` keyword's `total` field. Whether to amend AC-010 is a lead decision; an amendment
  would be a `manager-spec` edit.
- (b) AC-011's "oldest `G`" resolves to the superseded round 1 evidence commit `6e56840d5`. The AC
  passes; the standing round's precedence is shown by AC-016 and by the supplementary
  `git merge-base --is-ancestor f167a9cd8 cc4092513` check in § Closure checks (M5). Whether the
  wording changes is a lead decision.
- (c) §E.3 `run_commit_sha: pending-backfill` belongs to the run-phase owner and is not backfilled
  here.
- (d) Follow-up candidate: simplify the thirteen-step suppression procedure, kept in its measured
  form by operator decision.

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
