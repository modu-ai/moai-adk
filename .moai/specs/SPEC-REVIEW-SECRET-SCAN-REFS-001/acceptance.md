# Acceptance — SPEC-REVIEW-SECRET-SCAN-REFS-001

> Harness: **standard**. Card base `feeecc980`; plan-time HEAD `21e5837dc`. The review workflow
> copies are unchanged between the two (`git diff --stat feeecc980 21e5837dc -- <both copies>`
> printed nothing), so every plan-time baseline below applies to both.

Names used below:

- `LOC` = `.claude/skills/moai/workflows/review.md`
- `TPL` = `internal/template/templates/.claude/skills/moai/workflows/review.md`
- `REGEX` = the scan regex as written in `review.md`:
  `(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})`
- `SP` = the session scratchpad, outside this repository
- `SECTION` = the secret-scan section extracted to a file:
  `sed -n '/^#### Secrets Scan/,/^#### Data Isolation Check/p' TPL > SP/section.txt`
  (the heading text is adjusted to the final heading if the edit renames it)

**Command shape.** The worktree guard refuses a `git` command piped into another command. Every
criterion below therefore writes `git` output to a file first (`--output=` or a plain redirect) and
counts from that file in a separate command.

## §D AC Matrix

| AC | REQ | Verification | Plan-time baseline → required after |
|---|---|---|---|
| AC-001 | REQ-001·002 | fixture review sequence with HEAD-line control, side-branch cell, over-time cell | SIDECELL never reported by the per-review step → reported by the adopted procedure |
| AC-002 | REQ-005 | exact-phrase grep in both copies | 1 and 1 → 0 and 0 |
| AC-003 | REQ-006 | `diff -q LOC TPL`, and both copies changed since `feeecc980` | exit 0, neither changed → exit 0, both changed |
| AC-004 | REQ-007 | added-line grep for `REGEX` over the card diff | 0 matches (128-line diff) → 0 matches |
| AC-005 | REQ-006 | neutrality greps on `TPL` | all 0 → all 0 |
| AC-006 | REQ-003·004 | per-option coverage-statement checks on `SECTION` | 1 of 2 scan commands lacks `--all` → per option, §D.6 |
| AC-007 | REQ-008 | decision commit is a strict ancestor of the first commit touching either copy | `**Decision:** pending` → ancestor, exit 0 |

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
2. **R1** — run the adopted procedure's first-run step (for Options 2 and 3 this records the
   checkpoint).
3. Commit `S1` carrying `SIDECELL` on branch `side` from `C0`. Commit `C1` carrying `HEADCELL` on
   `main`. Record `git merge-base --is-ancestor side HEAD` → exit 1.
4. **R2** — run the adopted per-review step, with the checkpoint as R1 left it.
   **Cell H (scanner alive):** `HEADCELL` count in R2's output. **Cell S:** `SIDECELL` count in
   R2's output.
5. Commit `C2`, clean, on `main`. Commit `S2` carrying `LATECELL` on a new branch `side2` from `C2`.
   Record `git merge-base --is-ancestor side2 HEAD` → exit 1.
6. **R3** — run the adopted per-review step, with the checkpoint as R2 advanced it.
   **Cell T (over time):** `LATECELL` count in R3's output. Also record the `SIDECELL` count in R3
   (informational — see the note below).
7. **Option 3 only — P:** run the prescribed periodic full-scan step once, after R3.
8. **Option 2 only — R4:** with no ref moved since R3, run the per-review step again.
   **Cell N:** output is 0 bytes.
9. Immediately before reading cells S and T, re-run both `git merge-base --is-ancestor` checks and
   record exit 1 for each. No merge of `side` or `side2` happens at any point in the sequence.

**Required outcome per option.**

| Cell | Option 1 | Option 2 | Option 3 |
|---|---|---|---|
| B | 0 bytes | 0 bytes | 0 bytes |
| H | R2 `HEADCELL` ≥ 1 | R2 `HEADCELL` ≥ 1 | R2 `HEADCELL` ≥ 1 |
| S | R2 `SIDECELL` ≥ 1 | R2 `SIDECELL` ≥ 1 | R2 `SIDECELL` 0 expected and recorded; P `SIDECELL` ≥ 1 |
| T | R3 `LATECELL` ≥ 1 | R3 `LATECELL` ≥ 1 | R3 `LATECELL` 0 expected and recorded; P `LATECELL` ≥ 1 |
| N | n/a | 0 bytes | n/a |

**Fail conditions (any option).** AC-001 fails when cell H is ≥ 1 but `SIDECELL` totals 0 across
every step the adopted procedure prescribes (R2, R3, and P where it exists), or when `LATECELL`
totals 0 across the same steps. A procedure that reports the HEAD line but not the side branch fails.
It also fails when cell B is non-zero, because every later count would then be unattributable.

**Note on re-reporting.** The `SIDECELL` count in R3 is informational. Option 1 re-scans all history
and is expected to report it again; Option 2 scans only what became reachable since R2 and is
expected not to. Requiring re-reporting would silently select Option 1 (`spec.md` REQ-002).

**Note on Option 3.** The zero counts in R2 and R3 are the per-review narrowing that Option 3 accepts
by design. They pass AC-001 only together with AC-006's disclosure requirement for Option 3.

## §D.2 AC-002 — the equivalence claim is gone from both copies

- **Given** both copies, where the claim currently stands at line 110.
- **When** the following run after the edit:
  `/usr/bin/grep -c 'produces the same coverage over time as the former every-review full scan' LOC TPL`
  and `/usr/bin/grep -c 'no finding class is dropped' LOC TPL`
- **Then** every count is `0`.
- **Plan-time baseline:** `1` and `1` for the first phrase; `1` and `1` for the second.

## §D.3 AC-003 — the two copies are byte-identical and both were edited

- **Given** the edit is complete.
- **When** `diff -q LOC TPL; echo "exit=$?"` runs, and
  `git -C <worktree> diff --stat feeecc980 HEAD -- LOC TPL` runs.
- **Then** `diff` prints nothing with `exit=0`, and the `--stat` output names **both** files.
- **Why both parts:** two unchanged, identical copies would also pass `diff -q`; the `--stat` part
  keeps the check from passing vacuously.
- **Plan-time baseline:** `diff` exit 0 (`reproduction.md`); `--stat` printed nothing.

## §D.4 AC-004 — no credential-shaped line is committed

- **Given** all of the card's commits, including evidence under `.moai/reports/t629/`.
- **When** `git -C <worktree> diff feeecc980 HEAD --output=SP/card-diff.txt` runs, then
  `wc -l SP/card-diff.txt`, then
  `/usr/bin/grep -cE -- '^\+.*(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})' SP/card-diff.txt`
- **Then** `wc -l` is non-zero, so the operand is not empty, and the grep count is `0`.
- The pattern requires an uppercase label, so a placeholder label such as `<LABEL>` and the regex text
  itself do not match it.
- **Plan-time reading (HEAD `21e5837dc`):** 128-line diff, count `0`, grep exit 1.

## §D.5 AC-005 — template neutrality of the distributed copy

- **Given** the edited `TPL` and `SECTION` extracted from it.
- **When** the following run:
  - `/usr/bin/grep -c -e 'SPEC-' TPL`
  - `/usr/bin/grep -c -e 't629' TPL`
  - `/usr/bin/grep -cE -e '12,157|12157|76\.084|0\.207' TPL`
  - `/usr/bin/grep -ciwE 'go|golang|python|typescript|javascript|rust|java|kotlin|csharp|ruby|php|elixir|cpp|scala|flutter|swift' SP/section.txt`
- **Then** every count is `0` — no new occurrence of an internal ID, no repository cost figure, and no
  programming language named in the section.
- **Plan-time baseline:** `0`, `0`, `0`, `0`. The file's four existing `primary` / `PRIMARY` hits
  (lines 121, 340, 344, 444) concern review focus and workflow execution paths, not a programming
  language, and are outside this section.

## §D.6 AC-006 — the coverage statement matches the adopted procedure

Common set-up: `/usr/bin/grep -e 'log -p' SP/section.txt > SP/cmds.txt`, then `wc -l SP/cmds.txt`
must be ≥ 1.

- **Option 1**
  - `/usr/bin/grep -vc -e '--all' SP/cmds.txt` → `0`, so every scan command covers all refs.
  - `/usr/bin/grep -c 'the HEAD SHA of the last completed scan' SP/section.txt` → `0`.
- **Option 2**
  - `/usr/bin/grep -vc -e '--all' SP/cmds.txt` → `0`.
  - `/usr/bin/grep -c -e '--not' SP/cmds.txt` → ≥ 1, so the command excludes previously recorded tips.
  - `/usr/bin/grep -ci 'every ref' SP/section.txt` → ≥ 1, so the checkpoint is described as the tips of every ref.
  - `/usr/bin/grep -c 'the HEAD SHA of the last completed scan' SP/section.txt` → `0`.
- **Option 3**
  - The sentence naming the uncovered refs, and the sentence stating the full-scan period, are pinned
    verbatim in `progress.md` §E.2 **before** the edit. `/usr/bin/grep -cF '<pinned sentence>' SP/section.txt`
    → `1` for each.
  - `/usr/bin/grep -c -e '--all' SP/cmds.txt` → ≥ 1, so the periodic full-scan step is present.

**Plan-time baseline:** `SP/cmds.txt` holds 2 lines, and `grep -vc -e '--all'` gives `1` (the
`<last-sha>..HEAD` command). `the HEAD SHA of the last completed scan` gives `1` in both copies.

## §D.7 AC-007 — the operator decision precedes the document edit

- **Given** the operator's choice is recorded in `spec.md` §3 as a line beginning
  `**Decision:** Option`.
- **When** the following run:
  - `git -C <worktree> log --format=%H -S 'Decision:** Option' -- .moai/specs/SPEC-REVIEW-SECRET-SCAN-REFS-001/spec.md`
    — take the oldest SHA as `D`.
  - `git -C <worktree> log --reverse --format=%H feeecc980..HEAD -- TPL` — take the first SHA as `R`.
  - `git -C <worktree> merge-base --is-ancestor D R; echo "exit=$?"`
- **Then** `exit=0` and `D` ≠ `R`. Repeat with `LOC` in place of `TPL`.
- **Plan-time baseline:** `spec.md` §3 reads `**Decision:** pending`; no commit since `feeecc980`
  touches either copy.

## §D.8 Candidate criterion C-1 — other unreachable ref kinds (not in the Definition of Done)

Only a local branch was exercised. Extending the §D.1 fixture is cheap — a few commands — so this
is recorded as a candidate that the run phase may promote to an AC with operator agreement:

- a lightweight tag on an unmerged commit carrying `TAGCELL`;
- a branch present only as a remote-tracking ref in a clone, carrying `REMOTECELL`;
- a stash entry whose stashed change carries `STASHCELL`.

The expected outcome is **uncertain, not merely unmeasured**, for the stash case. Git's `log`
documentation says merge commits show no patch — and do not match `-S`/`-G`-style searches — unless a
`--diff-merges` variant is given. A stash's working-tree change lives in a merge commit, so `--all`
alone may not report `STASHCELL`. Measuring it would settle the question.

## §D.9 Edge cases (recorded when encountered; not criteria)

- **Unresolvable checkpoint.** Option 2 with a previous tip that no longer exists, or Option 3 with a
  checkpoint SHA rewritten by a force-push. Record the observed behaviour; the fallback is unspecified
  (`plan.md` §G).
- **Credentials introduced in a merge resolution.** For the reason in §D.8, these may not surface
  under `git log -p` in any option. Inferred, not measured.

## §D.10 Definition of Done

- AC-001 through AC-007 PASS, each recorded in `progress.md` §E.2 with the command, its verbatim
  output, and the tree it was measured on.
- The operator's decision is recorded in `spec.md` §3 (AC-007).
- `moai spec lint SPEC-REVIEW-SECRET-SCAN-REFS-001` reports no findings.
- The lane does not run `make build`. The build and embed check belong to the lead at batch close
  and are not part of this Definition of Done.
- Any cost figure cited in the run-phase evidence states its load condition and tree.
