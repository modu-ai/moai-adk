# Acceptance — SPEC-REVIEW-SECRET-SCAN-REFS-001

> Harness: **standard**. Card base `feeecc980`; plan-time HEAD `21e5837dc`. The review workflow
> copies are unchanged between the two (`git diff --stat feeecc980 21e5837dc -- <both copies>`
> printed nothing), so every plan-time baseline below applies to both.
>
> Revised after the operator decision, on top of decision commit `af7eb142b`. The copies are still
> unchanged: `git diff --stat feeecc980 af7eb142b -- <both copies>` wrote 0 bytes, and
> `diff -q LOC TPL` exited 0.

Names used below:

- `LOC` = `.claude/skills/moai/workflows/review.md`
- `TPL` = `internal/template/templates/.claude/skills/moai/workflows/review.md`
- `REGEX` = the scan regex as written in `review.md`:
  `(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})`
- `SP` = the session scratchpad, outside this repository
- `FX` = a throwaway fixture repository under `SP`, configured as in §D.1
- `PROG` = `.moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/progress.md`
- `SECTION` = the secret-scan section extracted to a file:
  `sed -n '/^#### Secrets Scan/,/^#### Data Isolation Check/p' TPL > SP/section.txt`
  (the heading text is adjusted to the final heading if the edit renames it)
- `K` = the card branch tip at which the closure checks run, recorded in `PROG` §E.2 from
  `git -C <worktree> rev-parse HEAD` **before** the develop absorb

**Closure-check anchor.** AC-003, AC-004, AC-007, AC-010, AC-011, AC-012, and AC-016 run at `K`, before the
develop absorb, and every range or revision in their commands ends at `K`, never at `HEAD`. A re-run
after the absorb uses the recorded `K`, so a develop commit touching either copy cannot enter those
ranges.

**Command shape.** The worktree guard refuses a `git` command piped into another command. Every
criterion below therefore writes `git` output to a file first (`--output=` or a plain redirect) and
counts from that file in a separate command.

**Adopted option.** The operator adopted Option 2 (`spec.md` §3). Criteria written per option in
the earlier draft now state the Option 2 outcome only; Options 1 and 3 were not adopted.

## §D AC Matrix

| AC | REQ | Verification | Plan-time baseline → required after |
|---|---|---|---|
| AC-001 | REQ-001·002 | fixture review sequence for the worded Option 2 procedure | SIDECELL never reported by the per-review step → reported in the scan where `side` first becomes reachable |
| AC-002 | REQ-005 | exact-phrase and `same coverage` greps in both copies | 1 and 1 for each → 0 and 0 for each |
| AC-003 | REQ-006 | `diff -q LOC TPL`, and both copies changed since `feeecc980` | exit 0, neither changed → exit 0, both changed |
| AC-004 | REQ-007 | added-line grep for `REGEX` over the card diff | 0 matches (readings in §D.4) → 0 matches |
| AC-005 | REQ-006 | neutrality greps on `TPL` | all 0 → all 0 |
| AC-006 | REQ-003·004 | Option 2 coverage-statement checks on `SECTION` | no `--stdin`, no `^%(objectname)`, no `every ref`, HEAD-SHA checkpoint phrase 1 → §D.6 |
| AC-007 | REQ-008 | decision commit is a strict ancestor of the first commit touching either copy | `D` = `af7eb142b`, no copy-touching commit yet → ancestor, exit 0 |
| AC-008 | REQ-001·009 | gate cell ①: side ref reported where it first becomes reachable | not measurable at plan time → trustworthy verdict recorded |
| AC-009 | REQ-009 | gate cell ②: recorded tip that no longer exists | not measurable at plan time → trustworthy verdict recorded |
| AC-010 | REQ-009 | gate cell ③: timing record on this repository, and the lead's approval committed before its evidence | not measurable at plan time → trustworthy verdict recorded, approval commit a strict ancestor |
| AC-011 | REQ-009 | gate evidence commit is a strict ancestor of the first commit touching either copy | no gate commit, no copy-touching commit → ancestor, exit 0, three trustworthy verdicts |
| AC-012 | REQ-010 | stop behaviour after an untrustworthy cell | no gate record → no copy-touching commit after the stop, stop recorded |
| AC-013 | REQ-011 | allowlist positive control | no allowlist exists → listed value not reported |
| AC-014 | REQ-011 | allowlist negative controls in the listed line's commit and hunk | no allowlist exists → near-miss, unlisted, mixed, and PEM markers all reported |
| AC-015 | REQ-007·012 | `REGEX` count over both copies after the edit | 0 and 0 → 0 and 0 |
| AC-016 | REQ-013 | pinned procedure appears verbatim in `SECTION` and is unchanged from gate evidence to edit | no pin, no gate record, no edit → all three pinned lines found, `cmp` exit 0 |

## §D.1 AC-001 — fixture review sequence

The fixture is a throwaway git repository under `SP`, **outside** this repository, configured with a
fixture identity, `commit.gpgsign=false`, and `core.hooksPath=/dev/null`.

**Markers.** Synthetic PEM-style private-key header lines, each with a distinct uppercase label:
`HEADCELL`, `SIDECELL`, `LATECELL`. Each marker line is assembled at runtime from fragments, so the
full line never appears in any file of this repository — including a fixture script persisted as
evidence (REQ-007). Every label matches `[A-Z]+`.

**Recording.** Each step's output is redirected to its own file under `SP`; its exit code is read
without a pipe; the file is measured with `wc -c` and counted per label with `grep -c`.

**Sequence.**

1. Commit `C0`, a clean file, on `main`. **Cell B (baseline):** `git log -p --all -G 'REGEX'`
   exits 0 with **0 bytes** of output.
2. **R1** — run the worded procedure's first completed scan, which records the tip set.
3. Commit `S1` carrying `SIDECELL` on branch `side` from `C0`. Commit `C1` carrying `HEADCELL` on
   `main`. Record `git merge-base --is-ancestor side HEAD` → exit 1.
4. **R2** — run the per-review step, with the tip set as R1 recorded it.
   **Cell H (scanner alive):** `HEADCELL` count in R2's output. **Cell S:** `SIDECELL` count in
   R2's output.
5. Commit `C2`, clean, on `main`. Commit `S2` carrying `LATECELL` on a new branch `side2` from `C2`.
   Record `git merge-base --is-ancestor side2 HEAD` → exit 1.
6. **R3** — run the per-review step, with the tip set as R2 recorded it.
   **Cell T (over time):** `LATECELL` count in R3's output. Also record the `SIDECELL` count in R3
   (informational — see the note below).
7. **R4** — with no ref moved since R3, run the per-review step again. **Cell N:** output is
   0 bytes.
8. Immediately before reading cells S and T, re-run both `git merge-base --is-ancestor` checks and
   record exit 1 for each. No merge of `side` or `side2` happens at any point in the sequence.

**Required outcome.**

| Cell | Required |
|---|---|
| B | 0 bytes |
| H | R2 `HEADCELL` ≥ 1 |
| S | R2 `SIDECELL` ≥ 1 |
| T | R3 `LATECELL` ≥ 1 |
| N | R4 0 bytes |

**Fail conditions.** AC-001 fails exactly when the required-outcome table does not hold: cell B is
not 0 bytes, or cell H (R2 `HEADCELL`) is 0, or cell S (R2 `SIDECELL`) is 0, or cell T (R3
`LATECELL`) is 0, or cell N (R4) is not 0 bytes. A marker first reported one review late — `SIDECELL`
in R3 or `LATECELL` in R4 — does not satisfy the criterion; each marker must be reported in the scan
where its commit first becomes reachable (`spec.md` REQ-001). A non-zero cell B also makes every
later count unattributable.

An `is-ancestor` reading other than exit 1, or a non-zero exit from any of R1-R4, means the sequence
was not constructed as specified; it is rebuilt, and that reading is a gap, not a pass.

**Note on re-reporting.** The `SIDECELL` count in R3 is informational. Option 2 scans only what
became reachable since R2 and is expected not to report it again. Requiring re-reporting would
contradict `spec.md` REQ-002.

**Relation to AC-008.** AC-008 measures the procedure pinned before the wording (the M1 gate). This
criterion measures the procedure as worded in the edited section (M3). Both must pass, and AC-016
checks that the worded procedure is the pinned one.

## §D.2 AC-002 — the equivalence claim is gone from both copies

- **Given** both copies, where the claim currently stands at line 110.
- **When** the following run after the edit:
  `/usr/bin/grep -c 'produces the same coverage over time as the former every-review full scan' LOC TPL`
  and `/usr/bin/grep -c 'no finding class is dropped' LOC TPL`
  and `/usr/bin/grep -ci 'same coverage' LOC TPL`
- **Then** every count is `0`. The third check catches a reworded equivalence claim that the two
  exact phrases would miss.
- **Plan-time baseline:** `1` and `1` for the first phrase; `1` and `1` for the second; `1` and `1`
  for `same coverage` (measured at plan revision 0.2.1).

## §D.3 AC-003 — the two copies are byte-identical and both were edited

- **Given** the edit is complete.
- **When** `diff -q LOC TPL; echo "exit=$?"` runs, and
  `git -C <worktree> diff --stat feeecc980 K -- LOC TPL` runs.
- **Then** `diff` prints nothing with `exit=0`, and the `--stat` output names **both** files.
- **Why both parts:** two unchanged, identical copies would also pass `diff -q`; the `--stat` part
  keeps the check from passing vacuously.
- **Plan-time baseline:** `diff` exit 0 (`reproduction.md`); `--stat` printed nothing.

## §D.4 AC-004 — no credential-shaped line is committed

- **Given** all of the card's commits, including evidence under `.moai/reports/t629/`.
- **When** `git -C <worktree> diff feeecc980 K --output=SP/card-diff.txt` runs, then
  `wc -l SP/card-diff.txt`, then
  `/usr/bin/grep -cE -- '^\+.*(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' SP/card-diff.txt`
- **Then** `wc -l` is non-zero, so the operand is not empty, and the grep count is `0`.
- The pattern requires an uppercase label, so a placeholder label such as `<LABEL>` and the regex text
  itself do not match it.
- **Plan-time reading (taken at commit `21e5837dc` in place of `K`, which does not exist until
  `plan.md` §F M5):** `git diff feeecc980 21e5837dc`, 128-line diff, count `0`, grep exit 1.
- **Plan-time reading (working tree on top of `af7eb142b`, every plan-revision edit in place):**
  `git diff feeecc980 --output=SP/card-diff-wt.txt`, then `wc -l` → `1236` lines, then the grep
  above over `SP/card-diff-wt.txt` → count `0`, grep exit `1`. Gap: the lines recording this reading
  are not in it.
- **Plan-time reading (working tree on top of `496fe6153`, every 0.2.1 revision edit in place):**
  `git diff feeecc980 --output=SP/card-diff-wt2.txt`, then `wc -l` → `1667` lines, then the grep
  above over `SP/card-diff-wt2.txt` → count `0`, grep exit `1` (read separately). Gap: the lines
  recording this reading are not in it.

## §D.5 AC-005 — template neutrality of the distributed copy

- **Given** the edited `TPL` and `SECTION` extracted from it.
- **When** the following run:
  - `/usr/bin/grep -c -e 'SPEC-' TPL`
  - `/usr/bin/grep -c -e 't629' TPL`
  - `/usr/bin/grep -cE -e '12,157|12157|76\.084|0\.207' TPL`
  - `/usr/bin/grep -ciwE 'go|golang|python|typescript|javascript|rust|java|kotlin|csharp|ruby|php|elixir|cpp|scala|flutter|swift' SP/section.txt`
  - `/usr/bin/grep -c 'R language' SP/section.txt`
  - the CI strict template leak tier, scoped:
    `MOAI_TEMPLATE_LEAK_STRICT=1 go test -v ./internal/template/ -run '^TestTemplateNoInternalContentLeak$'`
    with output redirected to `SP/leak.txt` and its exit code read without a pipe, then
    `/usr/bin/grep -c -e '--- PASS: TestTemplateNoInternalContentLeak' SP/leak.txt` and
    `/usr/bin/grep -c -e '--- FAIL' SP/leak.txt`
- **Then** every grep count on `TPL` and `SP/section.txt` is `0` — no new occurrence of an internal
  ID, no repository cost figure, and no programming language named in the section — and the leak
  run exits 0 with the `PASS` count ≥ 1 (so the selector ran the test) and the `FAIL` count `0`. The
  result is recorded in `PROG` §E.2.
- **Why the strict tier:** CI enforces it on template paths, and it flags a 7-8 character lowercase
  hexadecimal run; a digest cut short in the distributed copy would pass the greps above and fail CI
  (`spec.md` §3.4 requires full length).
- **`go` in prose:** with `-i` and `-w`, `go` also matches the English verb. When the language count
  is non-zero, the hits are listed with `/usr/bin/grep -niwE` on the same pattern and each is
  recorded as the verb or a language name; only a language-name hit fails.
- **Why `r` is absent from the list:** as a case-insensitive whole word it matches single-letter
  flags and placeholders in shell examples, so its count would not read as a language name. The
  `R language` phrase check covers it instead.
- **Plan-time baseline:** `0`, `0`, `0`, `0`; `R language` `0`, exit 1 (measured at plan revision
  0.2.1); the leak run was not taken at plan time. The file's four existing `primary` / `PRIMARY`
  hits (lines 121, 340, 344, 444) concern review focus and workflow execution paths, not a
  programming language, and are outside this section.

## §D.6 AC-006 — the coverage statement matches Option 2

Common set-up: `/usr/bin/grep -e 'log -p' SP/section.txt > SP/cmds.txt`, then `wc -l SP/cmds.txt`
must be ≥ 1.

- `/usr/bin/grep -vc -e '--all' SP/cmds.txt` → `0`, so every history-scan command reaches all refs.
- `/usr/bin/grep -c -e '--stdin' SP/cmds.txt` → ≥ 1, so a history-scan command reads the recorded
  tips from standard input.
- `/usr/bin/grep -cF '^%(objectname)' SP/section.txt` → ≥ 1, so the section records the tip store as
  caret-prefixed object names. `-F` reads the pattern as fixed text, so `^` is a literal caret, not a
  line anchor.
- `/usr/bin/grep -ci 'every ref' SP/section.txt` → ≥ 1, so the checkpoint is described as the tips
  of every ref.
- `/usr/bin/grep -c 'the HEAD SHA of the last completed scan' SP/section.txt` → `0`.
- The sentence naming the commits the per-review step does not cover (REQ-004) is pinned verbatim
  in `PROG` §E.2 **before** the edit (`plan.md` §C item 3).
  `/usr/bin/grep -cF '<pinned sentence>' SP/section.txt` → `1`.

**Plan-time baseline, re-measured for 0.3.0 (`TPL` at `6e56840d5`; neither copy changed since
`feeecc980`):** `SP/section.txt` 20 lines; `SP/cmds.txt` 2 lines; `grep -vc -e '--all'` → `1`, exit 0
(the `<last-sha>..HEAD` command); `grep -c -e '--stdin'` → `0`, exit 1;
`grep -cF '^%(objectname)'` → `0`, exit 1; `grep -ci 'every ref'` → `0`, exit 1;
`the HEAD SHA of the last completed scan` → `1`, exit 0. Positive controls, over a scratch file in `SP`
holding one line of each form: the `--stdin` count → `1`, exit 0; the `^%(objectname)` count → `1`,
exit 0. The pinned-sentence check is not measured here; it reads the edited section. The earlier
check for `--not` (reading `0`, exit 1, at `af7eb142b`) is retired with the command-substitution form.

## §D.7 AC-007 — the operator decision precedes the document edit

- **Given** the operator's choice recorded in `spec.md` §3 as a line beginning
  `**Decision:** Option`.
- **When** the following run:
  - `git -C <worktree> log --format=%H -S 'Decision:** Option' -- .moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/spec.md`
    — take the oldest SHA as `D`.
  - `git -C <worktree> log --reverse --format=%H feeecc980..K -- TPL` — take the first SHA as `R`.
  - `git -C <worktree> merge-base --is-ancestor D R; echo "exit=$?"`
- **Then** `exit=0` and `D` ≠ `R`. Repeat with `LOC` in place of `TPL`.
- **Plan-time reading (HEAD `af7eb142b`):** the `-S` log printed one SHA, `af7eb142b` — the decision
  commit, committed alone. The `-- TPL` and `-- LOC` log forms each wrote 0 lines, so `R` does not
  exist yet and the ancestor check cannot run until a commit touches a copy. Control: the same log
  form over `spec.md` wrote 3 lines.

## §D.8 AC-008 — gate cell ①: the side ref is reported when it first becomes reachable

- **Given** the Option 2 procedure pinned verbatim in `PROG` §E.2 (`plan.md` §C item 3), and `FX`
  configured as in §D.1, with markers assembled from fragments at runtime.
- **When** the cell runs, in order:
  - commit `C0`, clean, on `main`; the pre-marker scan `git -C FX log -p --all -G 'REGEX'`
    redirected to `SP/g1-base.txt`, its exit code read without a pipe, then `wc -c SP/g1-base.txt`;
  - the procedure's first completed scan `R1`, which records the tip set;
  - commit `S1` carrying `SIDECELL` on `side` from `C0`, and `C1` carrying `HEADCELL` on `main`;
  - `git -C FX merge-base --is-ancestor side HEAD; echo "exit=$?"`;
  - the procedure's next scan `R2` redirected to `SP/g1-r2.txt`, its exit code read without a pipe;
  - `wc -c SP/g1-r2.txt`, `/usr/bin/grep -c 'HEADCELL' SP/g1-r2.txt`,
    `/usr/bin/grep -c 'SIDECELL' SP/g1-r2.txt`.
- **Then** the verdict is `trustworthy` exactly when `spec.md` §3.4 cell ① holds: base 0 bytes,
  every scan exit 0, is-ancestor exit 1, and in R2 `HEADCELL` ≥ 1 and `SIDECELL` ≥ 1. Otherwise it is
  `untrustworthy`. The verdict is recorded as a line `verdict: <value>` under a `#### Gate cell 1`
  heading in `PROG` §E.2.
- **Plan-time reading:** not red-measurable at plan time — no Option 2 procedure is pinned and none
  has run (`spec.md` §3.2). The failure this cell guards against is measured for the current
  HEAD-anchored procedure: `reproduction.md` measurement 1, `HEADCELL 1`, `SIDECELL 0`.
- **Mutant note:** a procedure that re-scans all history on every review passes this cell. AC-010's
  scope bound rejects it.

## §D.9 AC-009 — gate cell ②: a recorded tip that no longer exists

- **Given** the pinned procedure, and `FX` after a completed scan whose recorded tip set includes the
  tip `G1` of branch `gone`, where `G1` is reachable from no other ref.
- **When** the cell runs, in order:
  - `git -C FX branch -D gone`; `git -C FX reflog expire --expire=now --all`;
    `git -C FX gc --prune=now --quiet`;
  - construction check: `git -C FX cat-file -e <G1>; echo "exit=$?"`;
  - commit `L1` carrying `GONECELL` on a new branch `after`, not reachable from HEAD;
  - the procedure's next scan redirected to `SP/g2.txt`, its error stream to `SP/g2.err`, its exit
    code read without a pipe; any fallback scan the procedure runs redirected to
    `SP/g2-fallback.txt` the same way;
  - `wc -c` on each file, `/usr/bin/grep -c 'GONECELL'` on the file carrying the final scan result,
    and that final scan's exit code;
  - the detection reading: `/usr/bin/grep -cF '<G1>' SP/g2.err SP/g2.txt`, with
    `SP/g2-fallback.txt` added as a third operand when that file exists, where `<G1>` is the recorded
    tip's object name **without the leading caret** the tip store carries — git names a missing object
    that way; the per-file counts are summed.
  - how `<G1>` is derived: before `gone` is deleted, `git -C FX rev-parse refs/heads/gone > SP/g2-gone.txt`
    records the plain name; `sed 's/^\^//' <tip store> > SP/g2-store-plain.txt` strips the caret from
    every store line; `/usr/bin/grep -cxF -f SP/g2-gone.txt SP/g2-store-plain.txt` → `1` shows the
    store recorded that tip; `<G1>` is the one line of `SP/g2-gone.txt`.
- **Then** the construction check exits non-zero, and the verdict is `trustworthy` exactly when
  `spec.md` §3.4 cell ② holds: the detection reading totals ≥ 1, the scan carrying the final result
  exits 0, and it reports `GONECELL` ≥ 1 through behaviour (a) or (b). An unhandled error, a final
  scan with a non-zero exit, a silent skip, or a detection total of 0 gives `untrustworthy`. If the construction check exits 0, the cell was not constructed; it is rebuilt,
  and that reading is a gap, not a pass. The verdict is recorded under `#### Gate cell 2` in
  `PROG` §E.2.
- **Plan-time reading:** not red-measurable at plan time — no procedure is pinned. The failure it
  guards against is shown by a design probe (`progress.md` §E.2, git 2.50.1): `git log -p --all --stdin`
  reading a caret-prefixed line for a missing object exited 128 with 0 bytes on stdout and
  `fatal: bad object <name>` on stderr, naming the object without the caret.
- **Mechanics check (0.3.0, scratch files in `SP`, a one-line caret store of this worktree's branch
  tip):** the `sed` form above turned the 42-byte store into a 41-byte plain line —
  `/usr/bin/grep -c '^\^'` printed `0`, exit 1, on the plain file and `1`, exit 0, on the store; and
  `/usr/bin/grep -cxF -f` with the plain name as the pattern file printed `1`, exit 0, over the
  stripped store and `0`, exit 1, over the caret store, so the membership check needs the stripped
  store.

## §D.10 AC-010 — gate cell ③: the timing record on this repository

- **Given** the pinned procedure; the lead's approval of the first run recorded in `PROG` §E.2
  before that run, under a `### Lead approval for gate cell 3` heading quoting the lead's message,
  in its own commit ahead of the commit that records this cell's evidence; the tip set both runs
  record and read stored at `SP/g3-tips.txt`, outside this repository; this repository at a
  recorded HEAD commit.
- **When**, for the first run and then the incremental run:
  - `uptime > SP/g3-<run>-load-before.txt` before the scan, and `uptime > SP/g3-<run>-load-after.txt`
    after it;
  - the scan under `/usr/bin/time -p`, output to `SP/g3-<run>.txt`, timing and errors to
    `SP/g3-<run>.time`, exit code read without a pipe;
  - scope: the same revision arguments with `--format=%H` and without `-p` or `-G`, reading
    `SP/g3-tips.txt` on standard input where the run reads the store, to `SP/g3-<run>-scope.txt`, then
    `wc -l`;
  - matching commits: `/usr/bin/grep -c '^commit ' SP/g3-<run>.txt` — a count only;
  - tips excluded: `wc -l SP/g3-tips.txt`;
  - incremental run only, immediately before it, each to its own file: first the plain tip names,
    `sed 's/^\^//' SP/g3-tips.txt > SP/g3-tips-plain.txt`, then `wc -l` on both files (equal) and
    `/usr/bin/grep -c '^\^' SP/g3-tips-plain.txt` → `0`; then `A` =
    `git rev-list --count --stdin < SP/g3-tips-plain.txt`, `B` = `git rev-list --count --all`, and
    `L` = `git rev-list --count --not --all --stdin < SP/g3-tips-plain.txt`. The plain file is
    required: `A` and `L` count commits reachable **from** the recorded tips, and the store's
    caret lines read directly would negate them. A command-line `--not` does not negate revisions
    read through `--stdin`, so in `L` it negates `--all` only;
  - mechanics check (0.3.0, this worktree at `6e56840d5`, a one-line plain file holding the branch
    tip, in `SP`): the `A` form printed `7081`, equal to `git rev-list --count HEAD` → `7081`; and
    `git rev-list --count --not HEAD~1 --stdin < <that file>` printed `1` — the command-line `--not`
    left the standard-input tip positive (negated, the count would be `0`);
  - approval ordering, at `K`:
    `git -C <worktree> log --format=%H -S '### Lead approval for gate cell 3' feeecc980..K -- PROG > SP/g3-approval.txt`
    — the first line (newest) is `P`;
    `git -C <worktree> log --format=%H -S '#### Gate cell 3' feeecc980..K -- PROG > SP/g3-evidence.txt`
    — the first line (newest) is `E3`; then
    `git -C <worktree> merge-base --is-ancestor P E3; echo "exit=$?"`.
- **Then** the verdict is `trustworthy` exactly when `spec.md` §3.4 cell ③ holds: every field present
  for both runs, both runs exit 0, and the incremental scope line count ≤ `B − A + L`. If `A` or `L`
  cannot be computed because a recorded tip is absent, the pair is re-taken and is a gap until then.
  The verdict is recorded under `#### Gate cell 3` in `PROG` §E.2. No SHA, path, or matched value from
  the `SP/g3-*` files is copied into this repository. The approval ordering holds when both log
  files are non-empty, `P` ≠ `E3`, and the ancestor check prints `exit=0`; the newest line is used so
  that, after a new gate round (`plan.md` M3), the check reads the standing round's approval.
- **Judgement:** no threshold is set on seconds; the predicate decides only whether the record is
  complete and consistent (`spec.md` §3.4).
- **Plan-time reading:** not red-measurable at plan time — it needs the pinned procedure and the
  lead's approval. The only cost evidence on this repository, `cost-baseline.md`, measures the
  HEAD-anchored and full `--all` scans under contention, not Option 2.
- **Mutant note:** a procedure that re-scans all history on every review has an incremental scope of
  `B`, which exceeds `B − A + L` whenever any commit reachable from the recorded tips is still
  reachable (`A` > `L`).

## §D.11 AC-011 — the gate evidence precedes the document edit

- **Given** gate evidence recorded in `PROG` §E.2 under a `### Gate evidence` heading, with one
  `#### Gate cell N` sub-heading and one `verdict:` line per cell.
- **When** the following run:
  - `git -C <worktree> log --reverse --format=%H -S '### Gate evidence' feeecc980..K -- PROG > SP/g-commit.txt`
    — the first line is `G`.
  - `git -C <worktree> log --reverse --format=%H feeecc980..K -- TPL > SP/r-tpl.txt` — the first
    line is `R`.
  - `git -C <worktree> merge-base --is-ancestor G R; echo "exit=$?"`
  - `git -C <worktree> show R~1:.moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/progress.md > SP/pre-edit-progress.txt`,
    then `/usr/bin/grep -c '^verdict: trustworthy$' SP/pre-edit-progress.txt`
- **Then** `exit=0`, `G` ≠ `R`, and the count is `3`. Repeat with `LOC` in place of `TPL`.
- **Plan-time reading (HEAD `af7eb142b`):** not red-measurable — no gate evidence and no
  copy-touching commit exist (the `-- TPL` and `-- LOC` log forms each wrote 0 lines; the same form
  over `spec.md` wrote 3). The gate-evidence heading is deliberately absent from `PROG` at plan time,
  so `G` cannot resolve to a plan-phase commit.

## §D.12 AC-012 — the run phase stops on an untrustworthy cell

- **Given** a gate cell whose `verdict: untrustworthy` line first appears in `PROG` in commit `U`:
  `git -C <worktree> log --reverse --format=%H -S 'verdict: untrustworthy' feeecc980..K -- PROG > SP/u-commit.txt`,
  first line.
- **When** the following run:
  - `git -C <worktree> log --format=%H U..K -- TPL LOC > SP/after-stop.txt`, then
    `wc -l SP/after-stop.txt`
  - `git -C <worktree> log --format=%H feeecc980..U -- TPL LOC > SP/before-stop.txt`, then
    `wc -l SP/before-stop.txt`
  - control: `git -C <worktree> log --format=%H feeecc980..K -- .moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/spec.md > SP/ctl-log.txt`,
    then `wc -l SP/ctl-log.txt`
  - `/usr/bin/grep -c '^Run phase stopped: gate cell [123]' PROG`
- **Then** both copy-touching line counts are `0`, the control is ≥ 1 (the log form prints commits
  when they exist, so the zeros are not an empty-command artefact), and the stop-line count is ≥ 1.
- **Applicability:** only when a gate cell is untrustworthy. When all three are trustworthy, AC-012
  is recorded as not applicable, citing the three verdict lines.
- **Plan-time reading:** not measurable — no gate record exists. Neither an untrustworthy verdict
  line nor a stop line is present in `PROG` at plan time.

## §D.13 AC-013 — allowlist positive control

- **Given** the procedure as worded after M2, its allowlist in the chosen representation, and `FX`
  whose HEAD line carries a commit adding a line `LISTCELL <value>`, where `<value>` equals one listed
  public example value. The value is assembled from fragments held only in `SP`; no fixture script
  carrying those fragments is committed to this repository.
- **Findings file.** `SP/al-findings.txt` holds one line per finding the procedure reports: the line
  carrying the unsuppressed match. It is not the scan's patch output. The scan prints the whole diff
  of every matching file, including context lines and neighbouring added lines (measured on a
  fixture: `progress.md` §E.1, scan output granularity measurement). A line printed with a finding is
  not a finding, and a listed line printed next to a finding is not reported by being printed.
  Labels are counted only over lines that match `REGEX`.
- **When** the raw scan — the procedure's scan command without suppression — is redirected to
  `SP/al-raw.txt`, and the procedure's reported findings after suppression to `SP/al-findings.txt`,
  each exit code read without a pipe; then the lines matching `REGEX` are extracted from each file,
  `/usr/bin/grep -E -- 'REGEX' SP/al-raw.txt > SP/al-raw-match.txt` and
  `/usr/bin/grep -E -- 'REGEX' SP/al-findings.txt > SP/al-match-lines.txt`, and
  `/usr/bin/grep -c 'LISTCELL'` runs on each extracted file.
- **Then** the raw count is ≥ 1, so the listed value matches the regex and suppression is actually
  exercised, and the findings count is `0`. Counting over `SP/al-raw.txt` itself would not show
  this: `LISTCELL` shares its file with the negatives of AC-014 and is printed whenever they are.
- **Plan-time reading:** not measurable — no allowlist and no suppression step exist. Under today's
  procedure every raw match is a finding, so a listed value would be reported (inferred from the
  absence of any suppression text in `SECTION`).

## §D.14 AC-014 — allowlist negative controls

- **Given** the same fixture run as AC-013, with these lines added in the **same commit and the same
  hunk** as the `LISTCELL` line — on the lines immediately next to it — each value assembled from
  fragments held only in `SP`:
  - `NEARCELL <value>` — the listed value with only its last character changed, still matching the
    regex;
  - `OTHERCELL <value>` — a credential-shaped value on no list;
  - `MIXCELL <listed value> <unlisted value>` — one line carrying both the listed value and a
    credential-shaped value on no list;
  - plus the `HEADCELL` and `SIDECELL` PEM-header markers of §D.1, `SIDECELL` on the unmerged side
    branch.
- **When** construction is checked first: the commit carrying `LISTCELL` is written with
  `git -C FX show <that commit> --output=SP/al-commit.txt`, then `/usr/bin/grep -c '^@@' SP/al-commit.txt`
  and `/usr/bin/grep -c` for `LISTCELL`, `NEARCELL`, `OTHERCELL`, and `MIXCELL` on that file; then the
  two files of AC-013 and their `REGEX`-matching extracts `SP/al-raw-match.txt` and
  `SP/al-match-lines.txt` are produced exactly as AC-013 defines them, and `/usr/bin/grep -c` runs for
  `NEARCELL`, `OTHERCELL`, `MIXCELL`, `HEADCELL`, and `SIDECELL` on each extract. The findings file
  and the counting rule are AC-013's: a line printed with a finding is not a finding.
- **Then** the construction reading shows one hunk (`^@@` count `1`) and each of the four labels
  counted `1`; every label counts ≥ 1 in `SP/al-raw-match.txt` and ≥ 1 in `SP/al-match-lines.txt`.
  A construction reading other than that means the cell was not built; it is rebuilt, and that
  reading is a gap, not a pass.
- **Mutant pairing:** suppressing everything fails this criterion; suppressing nothing fails AC-013;
  a path-based suppression that passes AC-013 fails here, because the negative lines share the
  listed line's file; a suppression of the whole commit, the whole file, or the whole hunk once it
  contains a listed value fails on `NEARCELL` and `OTHERCELL`, which sit in that commit, file, and
  hunk; a suppression of the whole line fails on `MIXCELL`; a prefix match on the listed value fails
  on `NEARCELL`; reporting the scan's patch output as the findings file fails AC-013, because the
  listed line is printed with its neighbours and itself matches `REGEX`.
- **Plan-time reading:** not measurable — no allowlist exists.

## §D.15 AC-015 — no regex-matching text in either copy

- **Given** the edited copies.
- **When** `/usr/bin/grep -cE -- 'REGEX' LOC TPL` runs, and, as a positive control, the same command
  runs over a scratch file in `SP` holding a PEM-header line assembled from fragments.
- **Then** `0` and `0` for the copies, and `1` for the control.
- **Relation to AC-004:** AC-004 covers every added line in the card diff; this criterion pins the
  two copies after the edit, where a literal allowlist value would land (REQ-012).
- **Plan-time reading (HEAD `af7eb142b`):** `0`, `0`, grep exit 1; control `1`, exit 0. This is a
  regression guard: green at plan time because no allowlist exists yet; the control shows the command
  detects a regex-shaped line.

## §D.16 AC-016 — the edited copies prescribe the procedure the gate measured

- **Given** `PROG` §E.2 carrying the pinned procedure under a `### Pinned procedure` heading, with
  how and when the tip set is recorded on one line beginning `Tip recording: `, the scan command on
  one line beginning `Scan command: `, and the handling of a recorded tip that no longer exists on one
  line beginning `Missing-tip handling: `, each written in the form the document will carry
  (`plan.md` §C item 3); `R` as in AC-011; and `SECTION` extracted from the edited `TPL`.
- **When** the following run, at `K`:
  - `git -C <worktree> log --format=%H -S '### Gate evidence' feeecc980..R~1 -- PROG > SP/g-rounds.txt`
    — the first line (newest) is `G`, the gate evidence commit of the round standing at the edit.
    With no new gate round (`plan.md` M3) the file has one line and `G` equals AC-011's `G`.
  - `git -C <worktree> show G:.moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/progress.md > SP/prog-g.txt`
    and `git -C <worktree> show R~1:.moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/progress.md > SP/prog-r1.txt`
  - `/usr/bin/grep -c '^### Gate evidence$' SP/prog-r1.txt`
  - the pinned block, extracted by heading from each file:
    `sed -nE '/^### Pinned procedure$/,/^###? /{/^### Pinned procedure$/p;/^###? /!p;}' SP/prog-g.txt > SP/pin-g.txt`,
    the same over `SP/prog-r1.txt` into `SP/pin-r1.txt`, then `wc -l` on each
  - (b) `cmp SP/pin-g.txt SP/pin-r1.txt; echo "exit=$?"`
  - `sed -n 's/^Tip recording: //p' SP/pin-r1.txt > SP/pin-tips.txt`,
    `sed -n 's/^Scan command: //p' SP/pin-r1.txt > SP/pin-scan.txt`, and
    `sed -n 's/^Missing-tip handling: //p' SP/pin-r1.txt > SP/pin-gone.txt`, then `wc -l` and
    `/usr/bin/grep -c .` on each
  - (a) `/usr/bin/grep -cF -f SP/pin-tips.txt SP/section.txt`,
    `/usr/bin/grep -cF -f SP/pin-scan.txt SP/section.txt`, and
    `/usr/bin/grep -cF -f SP/pin-gone.txt SP/section.txt`
- **Then** all of these hold: the `### Gate evidence` count at `R~1` is `1`, so the newest listed
  commit added the heading rather than removed it; each extracted block is at least 2 lines, so the
  comparison is not between two empty files; `cmp` prints `exit=0`, so the pinned procedure is
  byte-identical at `G` and at `R~1`; `SP/pin-tips.txt`, `SP/pin-scan.txt`, and `SP/pin-gone.txt`
  each hold exactly one non-empty line, so no pattern file is empty or a match-everything blank line;
  and all three fixed-string counts are ≥ 1, so the pinned tip recording, the pinned scan command,
  and the pinned missing-tip handling each appear verbatim in `SECTION`. `LOC` is covered by
  AC-003's byte identity.
- **Mutant pairing:** gating one procedure and then wording another that drops the missing-tip
  handling fails (a); wording a section that records the tip set after the scan while the pin records
  it before the scan fails (a) on the tip-recording line; changing the pin after the gate without a
  new gate round fails (b), because `G` stays the round that measured the old pin; a pin that exists
  only after the gate fails the 2-line block check at `G`.
- **Mechanics check (plan revision 0.2.1, a scratch file in `SP`, not this repository):** the `sed`
  extraction above printed the heading and its body lines, kept a `####` sub-heading, and stopped
  before the next `###` heading; the `grep -cF -f` form printed `1` (exit 0) on a line holding the
  pattern and `0` (exit 1) on a control line without it.
- **Plan-time reading:** not measurable — no pin, no gate record, and no copy-touching commit exist,
  and `PROG` carries no line beginning `### Pinned procedure` or `### Gate evidence`.

## §D.17 Candidate criterion C-1 — other unreachable ref kinds (not in the Definition of Done)

Only a local branch was exercised. Extending the §D.1 fixture is cheap — a few commands — so this
is recorded as a candidate that the run phase may promote to an AC with operator agreement:

- a lightweight tag on an unmerged commit carrying `TAGCELL`;
- a branch present only as a remote-tracking ref in a clone, carrying `REMOTECELL`;
- a stash entry whose stashed change carries `STASHCELL`.

The expected outcome is **uncertain, not merely unmeasured**, for the stash case. Git's `log`
documentation says merge commits show no patch — and do not match `-S`/`-G`-style searches — unless a
`--diff-merges` variant is given. A stash's working-tree change lives in a merge commit, so `--all`
alone may not report `STASHCELL`. Measuring it would settle the question.

## §D.18 Edge cases (recorded when encountered; not criteria)

- **Vanished recorded tip.** Now a gate cell (AC-009), not an edge case.
- **Rewritten tip.** A recorded tip that a force-push replaced still exists until garbage collection,
  so its `^<tip>` store line keeps excluding commits that are no longer on any ref. Record the observed
  behaviour; it is not a gate cell.
- **Credentials introduced in a merge resolution.** For the reason in §D.17, these may not surface
  under `git log -p`. Inferred, not measured.

## §D.19 Definition of Done

- AC-001 through AC-011 and AC-013 through AC-016 PASS, each recorded in `progress.md` §E.2 with the
  command, its verbatim output, and the tree it was measured on.
- AC-012 applies only when a gate cell is untrustworthy. The run phase then ends at the stop, and
  AC-012 PASS with the stop report is the phase's outcome in place of the criteria after the gate.
  When all three cells are trustworthy, AC-012 is recorded as not applicable, citing the three
  verdict lines.
- The operator's decision is recorded in `spec.md` §3 (AC-007).
- `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` reports 0 errors and 0 warnings.
- The lane does not run `make build`. The build and embed check belong to the lead at batch close
  and are not part of this Definition of Done.
- Any cost figure cited in the run-phase evidence states its load condition and tree.
