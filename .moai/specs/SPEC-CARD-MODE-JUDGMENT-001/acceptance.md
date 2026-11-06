# acceptance.md — SPEC-CARD-MODE-JUDGMENT-001

The verification layer. Requirements (GEARS) are in `spec.md` §4; this file holds Given-When-Then criteria, each with an evidence-ledger entry. A GEARS requirement is never restated here and a Given-When-Then scenario is never a requirement. Every criterion header names the full requirement ids it verifies, so the traceability verb of the plan-auditor (Group 4) can read the mapping.

## 0. Conventions

- **Document-level pin** (binds every ledger entry that carries no pin of its own, `verification-completeness.md` §2.1): tree SHA `d31fb64239a688c053e455000cc9fe37f540b790`, read with `git rev-parse HEAD` at the time of the measurement in correction round 2 (branch `WT-fail-safe-serial`, clean, unpushed). Every command ran with cwd `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894`; the judging `moai` build for CLI observations is `moai_cp/20261007_195927`. No Go command ran in this phase. The template and deployed document trees are identical between this pin, the earlier pins `b7b44433c312da73de9f329ab08c8fa6c5aa4de0`, `74ad197380a0711cdb9816d4e37c94d29af730ea` and `55d4f1486a454ff446be6f0f0645d4cea76bf0cc`, and the audited tree `cc109a75b6bf9c06ce86d811a8f0068551f02484` (`git diff --stat b7b44433c312da73de9f329ab08c8fa6c5aa4de0 d31fb64239a688c053e455000cc9fe37f540b790 -- internal .claude .agents .codex` printed nothing in round 2; the pair `74ad197380a0711cdb9816d4e37c94d29af730ea` / `b7b44433c312da73de9f329ab08c8fa6c5aa4de0` printed nothing in round 1). Every entry below was re-run on the document-level pin in correction round 2; none is carried forward from an earlier round.
- **Carrier**: fenced evidence-ledger entries `E-nnn`, cited by the criteria. Each entry carries `command`, `stdout` (verbatim as printed), and `exit` (its own field).
- **Exit status**: for a command that is not `git`, the run appended `echo "exit=$?"` outside the cell, so the exact status is the recorded one. The worktree guard refuses a trailing `$?` after a `git` command, so for the three `git` cells (E-002b, E-002c, E-017) the exit is stated from what git guarantees and what the run observed, in the entry. Where a command prints nothing, its exit is still recorded (a zero-hit grep exits 1).
- **Tools**: `sh`, `awk`, `grep` (only `-c -E -F -x -l -r -e`), `cmp`, `git` (`grep` with `--and`/`--not`, `ls-files`, `diff --no-index`), `python3` (`-I`, including the read-only `closed-world-check.py` of this SPEC directory, which calls `git show`, `git diff --name-only`, `git ls-files` and `git diff-tree` itself). **Awk and the worktree guard**: the guard refuses an awk program that carries a `|` after a `print`/`printf`, so no ledger awk program contains one — every alternation is passed as `-v re='a|b'` and matched with `$0 ~ re` (the cells E-002a, E-006 and E-009 were refused as written by the iteration 1 audit and were rewritten in this form; each was run through the guard in round 2 and printed the RED values recorded below). No cell uses PCRE (`grep -P` is not portable: the system `grep` printed `invalid option -- P` for the former E-008a/E-008d), no `\b` (git's regex on this machine matched nothing for `\bt1393\b` on a file holding two such lines), no `{n}` quantifier (the worktree guard refuses it). Cells were run through `sh`; the judging `grep` for every cell is `/usr/bin/grep` (BSD 2.6.0-FreeBSD) — the shell-function `grep` of the Claude Code Bash tool is ugrep 7.8.4 and it mishandles the five-way alternation of `E-013c` (it prints `0` where `/usr/bin/grep` prints `2` `2` `1`), so where the two engines disagree a cell is judged by `/usr/bin/grep` (stated in the ceiling-exception round, defect D24 of the iteration 2 audit). The `paths` mode of `closed-world-check.py` also judges an explicit path list given after `--` instead of querying git; the negative-control cells `E-020f` and `E-020g` use that form.
- **Classes**: `R` release-blocking (has a RED-now cell that is red now for the stated reason, a green-path cell naming the milestone that flips it, and a mutant probe); `G` regression-guard (green by design now; it records a *known failing input* on which the instrument was observed red, so the green is interpretable — `verification-completeness.md` §1.1); `O` operational (undecidable by a command; recorded, never counted as a pass by this SPEC).
- **Verdict by exit**: where a cell used to print a number and always exit 0, it now exits 1 unless the printed numbers equal the expected ones (`verification-completeness.md` §1.1, report-not-verdict).
- **Both trees**: every `R` criterion reads the template side and the deployed side in one invocation, so an edit to one side only leaves one operand red. `AC-MJ-011` additionally holds byte equality (`cmp`) for the seven fully-mirrored pairs and `AC-MJ-017` holds the one declared fork.
- **`grep -c` over several files** prints one `path:count` line per file; the order of those lines is not part of the expectation, so green outputs below are stated as sets.
- **Green paths and mutants** were observed on a scratch simulation of the post-run text (the planned wording applied to copies of the eight documents outside the repository; `progress.md` G-15). That simulation is not evidence about the real tree after the run; it is evidence that the criteria are satisfiable by the planned wording and that each mutant listed is red. The RED-now cells run on the real tree. The planned wording is pinned in full by `closed-world-manifest.txt` (AC-MJ-020): its rows were generated from the line difference between the simulation and the base commit, and the simulation prints `0 0 0 0` on all sixteen files against it, so the pinned cells and the manifest cannot disagree about a line. In the ceiling-exception round the copies of all sixteen files rebuilt from the manifest rows alone, listed rows at a non-planned position, print `0 0 0 0 0` against the five-number checker (observed; the fifth number is `0` on any arrangement of the listed rows by construction, because a fence's predecessor is read walking past the listed lines).
- Milestones M1-M5 are those of `plan.md`.

## 1. Criteria

### AC-MJ-001 — The mode-judgment step exists on both sides with its eight contract tokens (R; REQ-MJ-001, REQ-MJ-002, REQ-MJ-003, REQ-MJ-004)

**Given** the cards companion `factory-dispatch-cards.md` on the template side and the deployed side,
**When** each file is scanned for a `### Mode judgment` subsection and the tokens `expected file set`, `in-flight card`, `same files or shared state`, `slot-gated surface`, `conflict check`, `--classification-file`, `names the conflicting card`, `reissued with an explicit classification` are counted *inside that subsection only*,
**Then** the scan prints `8` for the template copy and `8` for the deployed copy and exits 0 (it exits 1 when any count differs from 8).

The last two tokens verify the two halves of the requirements that the first six do not see: the recorded reason names the conflicting card of a `serial` card (REQ-MJ-003), and a card that reached the queue on the default is reissued with an explicit classification (REQ-MJ-004).

- RED-now: `E-001` (`0`, `0`, exit 1). Red because neither copy has a `### Mode judgment` heading; the card-classes section ends at the parallelism paragraph (`research.md` S1).
- Green path: M1 adds the subsection to the template and the deployed copy in one commit → `8` / `8`, exit 0.
- Mutants tried (observed, round 2): heading renamed → `0`/`0`; the predicate narrowed from "same files or shared state" to "same files" (the board-narrowing defect) → `7`/`7`; the clause that names the conflicting card deleted → `7`/`7`; the reissue sentence deleted → `7`/`7`; subsection broken in the deployed copy only → `8`/`0`. Residual: all eight tokens present but the predicate semantically reversed (spec R-2, read by the audit); an extra sentence inside the subsection is not seen here and is bounded by AC-MJ-020.

### AC-MJ-002 — No issuance instruction in sweep set R relies on the default decider (R; REQ-MJ-004)

**Given** the six documents of sweep set R (`research.md` §7),
**When** the sweep for a violating sentence V is run, and a positive presence check runs on the todo skill,
**Then** (c1) the sweep over the working tree prints nothing and exits 0; (c2) the same predicate over the base commit still prints the `gtd.md:52` line (the control: the predicate is live); (c3) the swept set is exactly the six paths; (c4) each copy of `gtd.md` has at least one non-table line carrying both `moai todo add "` and `--classification-file`.

V is defined once: a non-table line (`^[^|]`) containing `moai todo add "` or `moai gtd add "` that does not contain `--classification-file`. The cell exits 1 when it prints a violating line.

- RED-now: `E-002a` (prints the `gtd.md:52` line, exit 1) and `E-002d` (`:0`, `:0`, exit 1); `E-002b` (control over the base commit, one line) and `E-002c` (swept set) hold their green values already. Red because `gtd.md:52` is the sole instruction to issue a card and it names no classification. The sweep engine of E-002a (`awk`) and of the control E-002b (`git grep` with boolean operators) differ; on the pre-edit tree both printed the identical single line.
- Green path: M1 rewrites that instruction → c1 empty/exit 0, c4 `:1`-or-more on both copies; c2 and c3 unchanged.
- Mutants tried (observed): the instruction written with single quotes (`moai todo add '<description>'`) → c4 `0` (red); the flag dropped from the instruction → c1 prints the line and c4 `0`. Not observed, by construction: delete the instruction (c1 passes, c4 fails); the flag on a table row only (c1 never swept it, c4 stays `0`); the flag on a later physical line of a wrapped instruction (c1 still fires — a deliberate false positive, strict). Residual: an instruction written as prose with no command text (AC-MJ-003 and the C home carry it).

### AC-MJ-003 — The leader's queue-entry clause requires the explicit classification and cites its rule (R; REQ-MJ-004)

**Given** `factory-dispatch.md` on both sides,
**When** lines are counted that start with `[HARD]` and contain both `--classification-file` and `factory-dispatch-cards.md`,
**Then** each copy reports `1`.

- RED-now: `E-003` (`:0`, `:0`, exit 1). Red because the leader clause at line 30 names only the verb `moai gtd add`.
- Green path: M1 adds one new [HARD] line directly after line 30 on both sides (line 30 itself is ledger-pinned and stays verbatim, `design.md` §D-3) → `:1` / `:1`.
- Mutants tried (observed): the new line without the cite → `:0`. By construction: new line in one tree only → `:1` / `:0`; line placed in C instead of F → `:0`.

### AC-MJ-004 — Class C: `--spec` is given at pick, and an empty `spec_id` is not called normal for it (R; REQ-MJ-005)

**Given** C and `gtd.md` on both sides,
**When** (a) C is searched for the literal `cannot be added after assign`, (b) `gtd.md` is searched for the unqualified fragment ``An empty `spec_id` on a picked card is``, (c) `gtd.md` is searched for the two whole lines that carry the qualified normality sentence, (d) the lines of `gtd.md` containing `normal state` are counted, and (e) `gtd.md` is searched for the two whole lines that carry the Class C clause of the pick step,
**Then** (a) reports `1` for each copy of C, (b) reports `0` for each copy of `gtd.md`, (c) reports `2` per copy, (d) reports `1` per copy, and (e) reports `2` per copy.

The pinned lines are: `  `next --spec` is explicitly given. A Class C card is picked with `--spec`.` and `  On a Class A or Class B card an empty `spec_id` is the normal state; it is` (c), and `   Attach the SPEC only when one exists and its identifier is known; a Class C` and `   card is picked with `--spec` (`factory-dispatch-cards.md` § Mode judgment).` (e), each as a whole physical line including its leading spaces.

- RED-now: `E-004a` (`:0`, `:0`, exit 1), `E-004b` (`:1`, `:1`, exit 0), `E-004c` (`:0`, `:0`, exit 1), `E-004e` (`:0`, `:0`, exit 1); `E-004d` (`:1`, `:1`) is green already and is the guard against a second normality claim. Red because C says nothing about Class C picks and `gtd.md` still calls an empty `spec_id` the normal state without qualification.
- Green path: M1 adds the sentence to `### Mode judgment` and rewrites the `gtd.md` spec_id bullet and the pick step → (a) `:1`/`:1`, (b) `:0`/`:0`, (c) `:2`/`:2`, (d) `:1`/`:1`, (e) `:2`/`:2`.
- Mutants tried (observed): the sentence reordered ("On a picked card an empty `spec_id` is the normal state") → (c) `:1` (red; the former fragment check (b) alone passes this mutant); a second normality claim added → (d) `:2` (red); the pick step left unqualified → (e) `:0` (red).

### AC-MJ-005 — A free lane is defined by "no card actually executing", the inclusion clause is pinned, and the old definition is gone, in the Routing bullet and in the transition bullet (R; REQ-MJ-006)

**Given** `factory-dispatch-detail.md` on both sides,
**When** (a) the literal `no card is actually executing` is counted, (b) the literal `free means the lane's previous card reached` is counted, (c) the literal `not a free lane` is counted, (d) the three whole lines of the Routing bullet (the parent line and its two sub-bullets, pinned in `E-005d`) are counted, and (e) per copy, the lines that tie leaving the card worktree or the next card to `done` alone (a line carrying `ExitWorktree(keep)` or `next card`, and `done`, and not `merge-ready`) are counted, together with the transition-boundary bullet that carries the merge-ready / held (operator-wait) / serial-waiting inclusion and the `next --card` re-entry,
**Then** (a) reports `1` per copy, (b) reports `0` per copy, (c) reports at least `1` per copy, (d) reports `3` per copy, and (e) prints `0 1` per copy and exits 0 (it exits 1 when the first number is not `0` or the second is not `1` in either copy).

The sub-bullet that carries the inclusion clause reads: a free lane is a lane on which no card is actually executing, and a lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is a free lane. Pinning it whole closes the polarity hole a token check leaves (a sentence saying such a lane is *not* free would still carry the tokens).

- RED-now: `E-005a` (`:0`, `:0`, exit 1), `E-005b` (`:1`, `:1`, exit 0), `E-005c` (`:0`, `:0`, exit 1), `E-005d` (`:0`, `:0`, exit 1). Red because D:133 still defines free as "the lane's previous card reached `done`" and says nothing about merge-ready, held or serial-waiting cards or about the live-session condition.
- Green path: M3 rewrites the Routing bullet on both sides in one commit → (a) `:1`, (b) `:0`, (c) `:1`, (d) `:3`.
- Mutants tried (observed): the inclusion clause reversed ("is not a free lane") → (c) and (d) change; a sentence added *to* (edited into) the Routing parent line with none of the promotion stems → (d) falls below `3` (the edited line no longer equals the pinned line); the old definition appended again → (b) `:1` and (d) change. Not observed: define free in M instead (a fails; AC-MJ-016 lists M).
- Does not see, stated (observed in round 2): a line *added* beside the three pinned lines leaves (d) at `3`, because (d) counts the pinned lines that are present and an extra line is not one of them — an extra stem-less sub-bullet in the Routing bullet (two wordings), a sentence in the card-review section and a sentence before the subsection heading all printed `3`. Bounding every added line is AC-MJ-020's work (all four mutants are red there).
- (e), added by the sync-audit amendment (0.1.8; sync audit `.moai/reports/t1894/sync-audit.md` S1): the M3 text left D's transition bullet unchanged — "A factory lane leaves its card worktree (`ExitWorktree(keep)`) once its card reaches `done`, before the next card is routed to it" — which restates, in other words, the rule (b) removes from the Routing bullet. (a)-(d) never read that line, and AC-MJ-020 allowed it because an unchanged line is not an added one. RED-now: `E-005e` (`1 0`, `1 0`, exit 1, pinned to `88bc5bcfb41a31afb9944ea76e1bd7b9748a1ee7`). Red because each copy still holds the `done`-only transition bullet (first number `1`) and neither holds the bullet with the inclusion and the re-entry (second number `0`). The base commit is red the same way: the bullet is older than this SPEC. Green path: the correction commit after this amendment replaces the bullet on both sides with the manifest's D row → `0 1`, `0 1`, exit 0 (observed on scratch copies of the two twins carrying that row).
- Mutants tried for (e) (observed, scratch, each file given as both operands): the old bullet kept and the new one added → `1 1`, exit 1; the new bullet in place and a paraphrase added, "A lane leaves its tree (`ExitWorktree(keep)`) only after its card is `done`." → `1 1`, exit 1; the new bullet deleted → `0 0`, exit 1; the new bullet in place and a paraphrase with no worktree verb added, "The leader routes the next card to a lane only after its previous card reached `done`." → `1 1`, exit 1. Does not see: a paraphrase that names neither `ExitWorktree(keep)` nor `next card` (for example "a lane is given new work only once its card is done") — the cell is lexical (`spec.md` R-8), and the audit reads for it.

### AC-MJ-006 — The delegated-dispatch promotion is stated once, with every limit of the operator ruling and no other (R; REQ-MJ-007)

**Given** the detail companion D on both sides,
**When** the subsection `### Delegated-dispatch promotion` is scanned — bounded by the next line that starts with `## `, so that sub-headings and bullets inside it are counted — and the command prints, per file, `<distinct subsection lines equal to a pinned line> <non-empty subsection lines> <1 if the next level-2 heading is exactly "## The card-review stage"> <lines outside the subsection containing the stems promot, queued or parallelizable>`,
**Then** each copy reports `8 8 1 1` and the command exits 0 (it exits 1 on any other numbers).

The fourth number is `1` today and stays `1`: the one line outside the subsection that carries a stem is the deputy-mode sentence "A deputy recommendation promoted to a verdict …", which this SPEC does not touch. A widening that carries a stem, written anywhere else in D, raises it.

- RED-now: `E-006` (`0 0 0 1` for the template copy, `0 0 0 1` for the deployed copy; exit 1). Red because D has no such subsection — its `##` headings run from `## Design intent` to `## Cross-references` and none is this one (`research.md` S12, S20). This work can flip it: M3 adds the subsection to the template copy and the deployed copy in one commit.
- Green path: M3 step 2 → `8 8 1 1` and `8 8 1 1`, exit 0. The eight pinned lines are the ruling's seven limits and one labelled consequence, in the ruling's order: the delegation condition with the statement of the case (line 1), where (line 2), which cards (line 3), order (line 4), what is never promoted (line 5), serial cards (line 6), what is unchanged (line 7), and the consequence for the assignment input (line 8). Positive control: the same command on a file that holds the planned text prints `8 8 1 1`.
- Why whole lines, not tokens: every limit of the ruling is a restriction or a negation, and a token check cannot see a limit reversed inside a sentence that still carries the token. Whole-line equality plus the line count makes a widening visible however it is written — a reworded limit changes a line, an added limit adds a line, a dropped limit removes one.
- Why the end marker and the stem scan: the former scan ended at any `^##` line, so a `####` sub-heading, a separate `###` heading or a Routing-bullet sentence placed after the nine lines printed `9 9` (observed before this correction). The scan now ends at the next `^## ` line, requires that line to be the card-review heading, and counts stems outside the subsection.
- Mutants tried (observed, each on a scratch copy, with the numbers and the exit the command printed): good copy `8 8 1 1` exit 0; "a held card," removed `7 8 1 1` exit 1; delegation condition removed `7 8 1 1` exit 1; serial line replaced by "promoted too" `7 8 1 1` exit 1; trailing space on the order line `7 8 1 1` exit 1; a pinned line duplicated `8 9 1 1` exit 1; a `####` sub-heading carrying a widening `8 10 1 1` exit 1; a separate `###` heading carrying a widening `8 10 1 1` exit 1; a widening with a stem in the Routing bullet `8 8 1 2` exit 1; a widening with a stem in a Routing sub-bullet `8 8 1 2` exit 1; a widening appended to line 1 `7 8 1 1` exit 1; an extra bullet `8 9 1 1` exit 1; an extra `##` heading before the card-review heading `8 8 0 2` exit 1; heading renamed `0 0 0 6` exit 1; heading deleted `0 0 0 5` exit 1; end marker renamed `8 8 0 1` exit 1; the consequence line relabelled a limit `7 8 1 1` exit 1. Not caught here, and listed under *Evades* below: the four stem-less added lines (`8 8 1 1`, exit 0).
- Evades, stated (observed in round 2 on the good copy plus one added line, each printing `8 8 1 1` and exit 0): a widening with none of the three stems, written as a line *added* anywhere in D outside the subsection — an extra sub-bullet in the Routing bullet (two wordings), a sentence in the card-review section, a sentence before the subsection heading. Neither this command nor AC-MJ-005 (d), which counts the pinned Routing lines that are present, sees an added line. AC-MJ-020 does: it lists the lines the run may add to and remove from D and reports any other line, with or without a stem (all four mutants observed red there). A paraphrase of the exception written as an added line in any of the eight documents is reported by AC-MJ-020, and one written in any other file is an unexpected path there. What no command sees is the manifest's own wording, which the audit reads (spec R-8). A wording improvement to a pinned line fails the command by design; the criterion and the manifest are amended with the wording.

### AC-MJ-007 — The renewal rule has one home, states its facts as pinned lines, and nothing else sits in it (R; REQ-MJ-008, REQ-MJ-009)

**Given** M on both sides,
**When** the subsection `### Lease renewal` is scanned (bounded by the next heading) and the command prints, per file, `<distinct subsection lines equal to a pinned line> <non-empty subsection lines> <1 if the next heading is exactly "## Isolation — launcher, tiers, branch naming, and the traceability carriers">`,
**Then** each copy reports `4 4 1` and the command exits 0 (it exits 1 on any other numbers).

The four pinned lines carry: the definition of *executing* (once, here); the rule that a lane renews the lease of the card it is executing and no other, containing `only the card the lane is actually executing`; the statement that a lease on a merge-ready, held (operator-wait) or serial-waiting card may expire and that `expiry is harmless`, with the landing-slot reassignment or `factory next --card` re-lease; and the statement that the standing recheck cron is kept while the lane executes the card, that a lane waiting on a slot or a trigger arms and keeps `no polling cron` for that wait, and that the leader sends a `notice` when a slot or a landing turn opens. The last line verifies the no-polling half of REQ-MJ-009, which the token list did not.

- RED-now: `E-007` (`0 0 0`, `0 0 0`, exit 1). Red because M:26 renews "at the start of each watchdog pass" with no executing-card scope and there is no `### Lease renewal` subsection.
- Green path: M2 adds the subsection (and rewrites M:26 to point at it) on both sides in one commit → `4 4 1` / `4 4 1`, exit 0.
- Mutants tried (observed): a contrary paragraph added inside the subsection → `4 5 1` (the old token list passed this); "expiry is harmful" → `3 4 1`; "arms a polling cron for that wait" → `3 4 1`. By construction: heading only → `0 0 1`; subsection in the template only → `4 4 1` / `0 0 0`.

### AC-MJ-008 — Every statement that arms or keeps the standing cron is scoped to the executing card, site by site, and none is deleted (R; REQ-MJ-010)

**Given** K, F, A, W and M on both sides,
**When** (a) lines are listed that carry a keep-word (`kept`, `keeps`, `keep it`, `stays armed`, `stay armed`, `armed`) and a cron-context word (`completion report`, `standing recheck`, `standing carrier`, `recheck cron`, `CronCreate`, `**standing**`, `standing wake`) and do **not** carry `while the lane executes the card` (version-history lines `^Version:` excluded); (b) lines are counted that carry one of the six site fragments — `is kept while the lane executes the card.` (K), `recheck armed before the first stage, kept while the lane executes the card;` (F), `keeps it while the lane executes the card:` (F), `and kept while the lane executes the card.` (A), `is kept only while the lane executes the card` (A), `armed at card intake and kept while the lane executes the card` (W); (c) the wording ``every renewal period of `workflow.factory_lease.renew_interval` `` is counted in K and F; (d) the literal `when the lane stops executing the card` is counted in A; (e) the two whole physical lines that carry the fire-and-renew sentence of A §5.1 (`  first (§0a of the watchdog skill), which renews the lease of the card the lane` and `  is executing, so the renewal period is also the cadence at which that lease`) are counted in A; (f) the opening ``While it executes the card, a lane renews its own card lease at the start of each watchdog pass with `moai factory renew <card>` `` of the renew paragraph is counted in M,
**Then** (a) prints nothing and exits 0; (b) reports K `1`, F `2`, A `2`, W `1` on both trees; (c) reports `1` for each of K and F on both trees; (d) reports `1` for A on both trees; (e) reports `2` for A on both trees; (f) reports `1` for M on both trees.

M's own site (the cron sentence of `### Lease renewal`) is verified by AC-MJ-007's pinned line. (e) and (f) cover the two sites of REQ-MJ-010 that carry no keep-word and so cannot appear in (a)'s list: the fire-and-renew sentence of A §5.1 (it scopes the renewal to the card being executed, and is not the `when the lane stops executing the card` delete sentence that (d) pins) and the renew paragraph of M (`factory-dispatch-mechanics.md` line 26, which AC-MJ-007 does not read because it scans the `### Lease renewal` subsection only). (e) and (f) exit 1 unless every printed count is the expected one.

The sites, measured (`research.md` §11): K:28; F:119 and F:122; A §5.1 intro, fire-and-renew sentence, delete sentence, and the §14 carrier sentence; W:244 (constraint bullet) and W §0a (AC-MJ-009); M:26 and M's `### Lease renewal`. Kept unedited because they only name the cron: A:103 and A:201.

- RED-now: `E-008a` (fourteen lines — K:28, F:119, F:122, A:117, A:118, A:396, W:244 on each tree — exit 1), `E-008b` (all eight `:0`, exit 1), `E-008d` (`:0`, `:0`, exit 1), `E-008e` (`:0`, `:0`, exit 1), `E-008f` (`:0`, `:0`, exit 1). `E-008c` is a regression-guard cell, `1` ×4 today, and must stay so. Red because those seven lines state "stays armed / kept / keeps … completion report" or "armed at card intake" with no scope, and because the fire-and-renew sentence and the renew paragraph still read as they did before the work.
- Green path: M2 → (a) empty/exit 0, (b) K `1`, F `2`, A `2`, W `1`, (c) unchanged, (d) `1`/`1`, (e) `2`/`2`, (f) `1`/`1`.
- Mutants tried (observed): K reverted to "stays armed to the completion report." → (a) fires and (b) K `0`; a marker phrase pasted in parentheses onto an otherwise unscoped sentence → (a) passes, (b) K `0` (red; the former marker-count check passed this); A intro, W:244, A §14 and F:122 each reverted → (a) fires and (b) drops; F:119's scope clause deleted → (a) fires and (b) F `1`; the fire-and-renew sentence of A §5.1 left in its present wording with everything else as planned → (e) `0` while AC-MJ-007 (`4 4 1`) and (a)-(d) stay green (the unchecked-site defect of the iteration 1 audit, observed with the old cells and again with the new ones); M:26 left unedited → (f) `0` with every other cell green (same observation); one of the two lines of the fire-and-renew sentence reworded → (e) `1`, exit 1; the "completion report" words deleted from the unscoped sentences while a marker phrase stays elsewhere in each file → (b) K `0`, F `0`, A `0` (the former checks — PCRE sweep, marker count, wording count — all passed this mutant, observed before this correction).
- Limit: a sentence that carries a site fragment and also contradicts it evades every lexical check (spec R-8).

### AC-MJ-009 — The watchdog's first-step renewal is scoped to the executing card by a pinned line (R; REQ-MJ-011)

**Given** `moai-lane-watchdog/SKILL.md` on both sides,
**When** the section `## 0a` is scanned and the command prints, per file, `<lines equal to the pinned scope line> <lines containing "moai factory renew"> <lines containing "every card" or "each card">`,
**Then** each copy reports `1 1 0` (the second number at least `1`) and the command exits 0.

The pinned scope line says that the step renews the card the lane is executing and no other card, that a lane whose cards are all merge-ready, held, or waiting for the serial slot does not renew in that pass, and that the pass carries the previous `renew_result` forward.

- RED-now: `E-009` (`0 1 0`, `0 1 0`, exit 1). Red because §0a renews "before anything else" with no state condition and the pinned line is absent.
- Green path: M2 adds the scope line, keeping the heading word `Renew`, the `moai factory renew` call and the closed set of `renew_result` values (`design.md` §D-3) → `1 1 0`, exit 0.
- Mutants tried (observed): the scope line removed → `0 1 0`; "Renew every card this lane holds." added beside it → `1 1 1`; the polarity flipped inside the line ("renews in that pass") → `0 1 0`. The former token check (`executing`, `merge-ready`) passed a §0a that renews every card (observed before this correction, printed `2`). By construction: the clause appended in §1 → the awk is section-scoped → `0`; reordering Renew after Measure (the Go guard `TestLaneWatchdogRenewsFirst` fails, run-phase).

### AC-MJ-010 — No reclassification verb is described as available; the factual default sentence stands (G; REQ-MJ-015)

**Given** the template and deployed trees,
**When** (a) the `.claude` trees (runtime directories `agent-memory` and `worktrees` excluded) are searched recursively for `moai (todo|gtd|factory) reclassif…` or `reclassify verb`, and (b) the `gtd.md` default-decider sentence is counted,
**Then** (a) lists no file and exits 1; (b) reports `1` per copy.

- Known failing input: `E-010c` — the same pattern over `.moai/reports/t1894/premise-check.md`, which states the verb, prints `1` (exit 0): the instrument goes red on text that describes the verb.
- Evidence now: `E-010a` (empty, exit 1), `E-010b` (`:1`, `:1`, exit 0).
- Mutants tried: a sentence "use `moai todo reclassify <id>`" (a lists the file); rewriting the default sentence of the `moai todo add` row (b fires); a sentence "reclassification verb" without the word `reclassify` ("re-classify") evades (a) (residual).

### AC-MJ-011 — Template and deployed copies stay byte-identical for the seven fully mirrored files, and each milestone commit carries both sides (G; REQ-MJ-013)

**Given** the pairs C, K, D, M, A, G, W,
**When** `cmp <template> <deployed>` runs for each,
**Then** each exits `0` with no output.

- Evidence now: `E-011` (seven exit-0 results). This is green before any edit (premise P1) and is **only** meaningful together with AC-MJ-001..009, which read both sides and are red now; it cannot be satisfied by "edit neither" because those criteria then stay red.
- Known failing input: `E-011f` — `cmp` on the `factory-dispatch.md` pair exits `1` (`differ: char 23127, line 222`); the instrument reports a difference when one exists. That pair is the declared fork and is excluded here and covered by AC-MJ-017.
- Run-phase obligation (commit shape; takes the undecidable disposition now, because no milestone commit exists yet, and is not recorded as a pass by this plan): for each of M1, M2 and M3, `python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py pairs .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt <milestone-sha>:<n>` — `<n>` is `3` for M1 (C, G, F), `5` for M2 (M, K, F, A, W) and `4` for M3 (D, F, G, M) — prints `<sha> pairs <n> unpaired 0 expected <n>` and exits 0; it exits 1 when any document of the manifest appears on one side only in that commit, or when the number of two-sided pairs differs from `<n>`. The commit graph shows that both sides land together; it does not show which was edited first, and this SPEC does not claim that (`verification-claim-integrity.md` §2.3). The earlier form of this obligation, a bare `git diff-tree --no-commit-id --name-only -r <milestone-sha>` listing, exited 0 whatever it listed and decided nothing.
- Instrument observed on known inputs (round 2, history of this repository): `E-011g` — a deployed-only commit prints `pairs 0 unpaired 7` and exits 1; `E-011h` — a commit that changed two pairs on both sides, given the expected count `2`, prints `pairs 2 unpaired 0 expected 2` and exits 0.
- Mutants tried (observed on the historical commits above and, by construction, for a milestone): edit the template only → one cmp exits 1 and the matching `R` criterion shows `:0` on the deployed operand; edit both sides to different wording → cmp exits 1; a milestone commit that carries a template document without its twin → `unpaired` is at least `1` and the command exits 1; a commit that carries the right pairs but a different number of them than planned → exit 1 on the `:<n>` comparison.

### AC-MJ-012 — The always-loaded core stub stays within 10,000 UTF-16 code units (G; REQ-MJ-014)

**Given** `factory-dispatch-core.md` (template side; the deployed copy is `cmp`-equal),
**When** its UTF-16 code-unit count is compared with 10,000,
**Then** the command prints `OK <n>` with `n ≤ 10000` and exits 0 (it prints `OVER <n>` and exits 1 above 10,000).

- Evidence now: `E-012a` prints `OK 9999` (headroom 1; exit 0). This is not vacuous: the work edits this file twice — the scope clause of K:28 (+4 units with the wording `is kept while the lane executes the card.`) and the opening clause of K:20 (−6 units) — both in the M2 commit, so the stub nets −2 and the first commit that touches it already lands within the budget (`design.md` §D-4, `research.md` §8).
- Known failing input: `E-012b` — the same command on `factory-dispatch.md` prints `OVER 25803` and exits 1 (the former form printed `OVER` and exited 0, which a gate reading the exit status could not see).
- Mutants tried (observed on the simulation): the planned wording gives `OK 9997`; a pointer-form K:28 (`; waiting cards are not renewed (…) …`) → `OVER 10162`; the K:20 clause edit omitted → `OVER 10003`; the marker phrase pasted onto an unscoped sentence → `OVER 10027`. A trim that removes a ledger-covered line passes this criterion and fails `TestBindingLedgerIntegrity` (run-phase).
- Limit: the measurement is a UTF-16 count of the file as stored; the Go gate renders the template first. The two agree for this file (no template directives); the Go gate is authoritative.

### AC-MJ-013 — The changed template text carries no new card id, board id, SPEC id, date, or long hex run (G; REQ-MJ-014)

**Given** the eight changed template files,
**When** a portable leak pattern is counted per file in the working tree — card ids (`t` + 3 or 4 digits) **except the five literals the card-id leak baseline already lists for these files** (`t133`, `t224`, `t696`, `t1339`, `t1393`), SPEC ids, board ids (`d-2026…`, `a672`, `0c51`, `02f4`), ISO dates, and hex runs of nine or more —
**Then** it reports `0` for every file and exits 1 (no line matched).

The exclusion is written into the pattern as digit classes, so a re-wrap of the paragraph that carries `t1393` and `t1339` in `auto-semantics.md` §5.1 (which the run edits) cannot trip it; a diff-based `-G` check would fire on the moved lines for text the work did not introduce (observed before this correction, and on this machine the former pattern's `\b` branches matched nothing at all). The five literals are the ones `internal/template/card_id_leak_test.go` lists for these eight files; the edit keeps those sentences (REQ-MJ-016).

- Evidence now: `E-013a` (eight `:0`, exit 1 — green by design, and not trivially so: `E-013c` shows the five baselined literals are present on five lines of three files, and the pattern does not count them).
- Known failing input: `E-013b` — the same pattern over `.moai/reports/t1894/premise-check.md`, which carries card ids, SPEC ids and SHAs, prints `8` (exit 0): the instrument goes red on text that carries them.
- Run-phase obligation (M4): re-run `E-013a` after M1-M3; a hit names the leaking file.
- Mutants tried (observed on the simulation): a card id `t1500` added to D → hit; a date added to G → hit; a SPEC id added to M → hit. Observed on scratch copies of `auto-semantics.md`: the paragraph re-wrapped so the two lines that hold `t1393` move → `0` (stays clean); `t1500` added next to `t1393` → `1` (fires). A card id spelled `t 1500` evades (residual).

### AC-MJ-014 — The pinned Go guards stay green after the change (G; REQ-MJ-013, REQ-MJ-014, REQ-MJ-016; run-phase only)

**Given** the run-phase tree after M1-M4,
**When** the scoped Go checks of `plan.md` §Verification run (one invocation per package; no `go test ./...`),
**Then** every test named below passes and each log is cited from the verdict file.

`internal/template`: `TestBindingLedgerIntegrity`, `TestRoleCoreStubGate`, `TestDeployedAlwaysLoadedCharBudget`, `TestRuleTemplateMirrorDrift`, `TestDeclaredRuleMirrorForks`, `TestReviewOwnership_CardReviewDoctrine`, `TestFactoryDispatchBodyCarriesRenewIntervalLinkage`, `TestLaneRecheckDoctrine`, `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures`, `TestWorkflowRulePathsPinned`, `TestTemplateNoInternalContentLeak`, `TestCardIDBaselineHasNoStaleEntries`, the catalog hash audit (`TestCatalogHashCoversSkillSubfiles` and its neighbours in `catalog_tier_audit_test.go`). `internal/cli`: `TestAutoPickMirrorParity`, `TestAutoPickDocDoctrine`, `TestTodoSkillDocumentsClassification`, `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestAutoRankMarkerDisclosure` (the last three are the guards that pin the old operator-only sentences and the passage the new pointer line sits in — `research.md` §6, §10).

- Disposition: not re-executable now (a Go run is not a single read-only invocation and none is allowed in this phase). It takes the undecidable disposition: it is a regression-guard, **not recorded as a pass** by this plan. Its RED is observable by the mutants above (for example the pointer-form K edit turns `TestRoleCoreStubGate` red).
- Per-commit coherence: these tests run once, on the final tree (M5). Each milestone commit is kept coherent by construction — both sides in the commit, the ledger rows and the catalog hash of what the commit changes in the same commit, and both stub edits in one commit — and is checked by the read-only cells of its own milestone, not by the Go run (`plan.md` §E).

### AC-MJ-015 — The lanes acknowledge receipt of the new guidance (O; REQ-MJ-010)

**Given** the landed change, **When** the leader sends the lane notice (the changed rules, with the SessionStart notice caveat of spec R-1), **Then** each live lane replies or records that it read the rules before its next card.

- Not re-executable by a command (a lane's reading is not observable on disk). The leader records the acknowledgements in the card's progress record; they are not counted in this SPEC's PASS and the criterion is never "satisfied" by this SPEC's own commands.

### AC-MJ-016 — Each rule has exactly one home (R; REQ-MJ-012)

**Given** the eight changed template files,
**When** the files are listed that contain any of the six home tokens `slot-gated surface`, `no card is actually executing`, `only the card the lane is actually executing`, `expiry is harmless`, `cannot be added after assign`, `while the operator has delegated dispatch`,
**Then** exactly `factory-dispatch-cards.md`, `factory-dispatch-detail.md` and `factory-dispatch-mechanics.md` are listed.

- RED-now: `E-016` (no file listed, exit 1). Red because none of the home sentences exists yet (the set is empty, not the right three).
- Green path: M1-M3 → the three homes.
- Mutants tried (observed): a home token (`expiry is harmless`) restated in F → F is listed (red). By construction: paraphrase a rule into a second file (evades — residual, read by the audit); put all rules in one file (only one listed → red).

### AC-MJ-017 — The declared fork of `factory-dispatch.md` stays exactly one differing line (G; REQ-MJ-013)

**Given** the two copies of `factory-dispatch.md`,
**When** `git diff --no-index --shortstat` compares them,
**Then** it prints ` 1 file changed, 1 insertion(+), 1 deletion(-)` (exit 1).

- Evidence now: `E-017`. The `1 insertion, 1 deletion` figure is a baseline measured on the pinned tree, not an absolute law: the criterion is "unchanged from the baseline measured at the start of the run phase". If another landing resolves the fork before this SPEC runs (the Go guard `TestAutoPickMirrorParity` would change with it), re-measure with the `E-017` command and re-pin before M1 (`verification-completeness.md` §4); never re-quote this number. Known failing input: any edit to one copy only changes the counts (by construction: the pair already differs by the worktree-sweep line).
- Mutants tried (by construction): apply the same edit to both → still `1/1` (correct); edit one side only → `2/2` or more (red).

### AC-MJ-018 — Each sentence that called promotion the operator's alone now carries the exception by pointer, and none restates it (R; REQ-MJ-017)

**Given** the leader's body F and the core stub K on both sides, the todo skill G on both sides, and M on both sides,
**When** (a) the exact pointer line of F is counted as a whole line, (b) the opening clause `The one reconciliation is named` is counted in F and K, (c) the replacement clause `A reconciliation is named, not excepted:` is counted in F and K, (d) the exact pointer line of G is counted as a whole line, and (e) the exact row of M's leader-duty enumeration is counted as a whole line,
**Then** (a) reports `1` per copy of F, (b) reports `0` for each of the four copies, (c) reports `1` for each of the four copies, (d) reports `1` per copy of G, and (e) reports `1` per copy of M.

The pointer lines say that the leader's promotion without a pick is the case stated in `factory-dispatch-detail.md` § Delegated-dispatch promotion, within the limits given there; they carry none of the six home tokens and no limit of their own, and they do not claim the case is the only one.

- RED-now: `E-018a` (`:0`, `:0`, exit 1), `E-018b` (`:1` ×4, exit 0), `E-018c` (`:0` ×4, exit 1), `E-018d` (`:0`, `:0`, exit 1), `E-018e` (`:0`, `:0`, exit 1). Red because neither pointer line nor the enumeration row exists and the paragraph at F:44 / K:20 still opens by claiming a single reconciliation (`research.md` §10, P2-P6, and §12 for the enumeration).
- Green path: M2 edits K and F's clause (K:20, F:44, with the ledger row ALB-0257 in the same commit) — the clause is the same in both, and F's copy of the paragraph also ends with one pointer sentence (``A further case is stated in `factory-dispatch-detail.md` § Delegated-dispatch promotion.``, the manifest's F row), so F's line is not the K line that ALB-0257 covers (mid-run amendment; see § 3 "A binding line copied into an on-demand file"); M3 edits F (the pointer line after F:40), G (the pointer paragraph after the [HARD] pick rule) and M (the table row), each side in the same commit → (a) `1`/`1`, (b) `0` ×4, (c) `1` ×4, (d) `1`/`1`, (e) `1`/`1`.
- Mutants tried (observed on the simulation): pointer line deleted from F → (a) `0`; clause reverted in the template stub only → (b) and (c) change (and the stub reads `OVER 10003`); enumeration row deleted → (e) `0`. By construction: pointer line in the template only → (a) `1`/`0`; a pointer line that restates the limits → (a) `0` because the line differs, and `E-016` lists F when the restatement carries a home token; the pointer moved into K → (a) `0` and AC-MJ-012 goes red on size.
- Does not establish: that no other sentence in the tree says the same thing in other words — `research.md` §10 and §12 are the measured lists, and the sync audit reads for a paraphrase.

### AC-MJ-019 — The sentences the doctrine guards pin are still present (G; REQ-MJ-016, REQ-MJ-017)

**Given** F on both sides,
**When** the pinned prohibition sentence ("Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.") and the pinned admission sentence ("queue ADMISSION (production) stays the operator's.") are counted,
**Then** each reports `1` per copy.

- Evidence now: `E-019a` (`:1`, `:1`, exit 0) and `E-019c` (`:1`, `:1`, exit 0). Green by design before any edit; it guards the run, because the natural wrong edit is to fold the exception into these sentences.
- Known failing input: `E-019b` — the prohibition pattern over the core stub, which does not carry that sentence, prints `0` (exit 1): the instrument goes red when the sentence is missing.
- Mutants tried (observed on the simulation): reword the sentence to carry the exception ("Outside an --auto authorization or while dispatch is delegated …") → `0` (red), and the Go doctrine tests of AC-MJ-014 go red too. By construction: delete it → `0`; move it to the detail companion → `0`.
- Does not establish: the two `absent` strings of `TestAutoPickDocDoctrine` ("Promotion is the operator's act, always.", "The leader never picks for the operator") stay absent. No command here carries a live control for an absence check, so that guard is left to AC-MJ-014's Go run (undecidable now).

### AC-MJ-020 — Every line the run adds to or removes from the eight changed document pairs is a line the SPEC lists, and no other path changes (R; REQ-MJ-006, REQ-MJ-007, REQ-MJ-012, REQ-MJ-017)

**Given** the eight changed template documents C, K, F, D, M, A, G, W, their eight deployed twins, and the manifest `closed-world-manifest.txt` of this SPEC directory (the complete list of lines to add and to remove, per file, pinned to the base commit `d31fb64239a688c053e455000cc9fe37f540b790`),
**When** (a) `closed-world-check.py lines` compares, for every file and on each side, the non-blank lines of the working copy with the non-blank lines of the same path at the base commit as multisets, and prints five numbers per side — added lines the manifest does not list, listed lines that were not added, removed lines the manifest does not list, listed removals that were not made, and code-fence placement changes against the base (a fence whose nearest preceding non-blank line, walking past the manifest's listed lines, differs; an unlisted line, fences included, present on both sides with a different number of fences before it; a listed line inside a code block beyond the base's count) — (b) `closed-world-check.py paths` lists every changed or new path that the manifest does not name, that is not the repository-root file `CHANGELOG.md` (exact path, admitted because the sync phase's mandatory `[Unreleased]` entry lands there), and that is not under one of the four paths a run of this SPEC legitimately writes (`.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/`, `.moai/reports/`, `.moai/state/`, `.moai/logs/`), tracked `.moai/` projections and any other `CHANGELOG` path included, and (c) the SHA-256 of the manifest and of the check are compared with the two values pinned in `E-020d`,
**Then** (a) prints `0 0 0 0 0` for the template side and `0 0 0 0 0` for the deployed side of every file and exits 0; (b) prints `unexpected-paths 0` and exits 0; (c) prints `same same` and exits 0. Each command exits 1 on any other number.

What this closes. AC-MJ-005 (d) and AC-MJ-006 are open-world: they look for known lines and known stems. Four stem-less additions passed both in the iteration 1 audit and again here (`E-005d` printed `3` and `E-006` printed `8 8 1 1`, exit 0): an extra sub-bullet in the Routing bullet (two wordings), a sentence in the card-review section, a sentence before the subsection heading. This criterion is closed-world: a line is allowed only if the manifest lists it, so one more line anywhere in any of the sixteen files — with or without a stem, in a heading, a table row or prose — turns the first number (or, for a deleted or rewritten line, the third) non-zero. A rewritten line is one listed removal and one listed addition, each by whole line; a non-fence line moved without change within its fence-delimited segment is not seen (the comparison ignores order) — a code fence moved across any line, alone or together with its preceding line or any run of lines, and a listed row written inside a code block, are reported by the fifth number — and a blank-line change is not seen. A line present once and written twice is an added line (the comparison counts).

The manifest is a SPEC artifact in the ownership class of this file: the plan phase wrote every row from the planned post-run text, the run phase does not edit it, and a wording change the run phase needs is a blocker report that manager-spec answers by amending the manifest together with the pinned cells that carry the same lines (E-004c/e, E-005d, E-005e, E-006, E-007, E-009, E-018a/d/e). It is frozen in the plan-phase commit, so the list of allowed lines is fixed by the party that does not write the documents and before the first document commit (`verification-claim-integrity.md` §2.3). `catalog.yaml` and `binding_ledger.json` are named by `also:` rows and are bounded by the Go guards of AC-MJ-014, not by line content. Neither the manifest nor the check belongs to the plan-artifact hash subject set (`spec-workflow-detail.md` § Report Persistence), so an edit to either would leave a cached audit verdict valid; their SHA-256 values are therefore pinned in `E-020d`, in this file, which is hashed: changing either file requires a matching change here.

- RED-now: `E-020a` (eight rows, exit 1). Red for the stated reason: every row's second number is non-zero (the listed additions have not been made: C 2, K 2, F 5, D 13, M 7, A 8, G 10, W 2 on each side) and the fourth is non-zero where the work rewrites a line (K 2, F 3, D 2, M 1, A 8, G 8, W 2); the first, third and fifth numbers are `0` today, so no unlisted line, no unlisted removal and no moved code fence exists on the unedited tree and the red cannot come from anything but the work not having been done. (D was 12 and 1 before the sync-audit amendment added the transition-bullet pair to the manifest; the D row of `E-020a` was re-measured on the base-commit documents for that amendment.) After the M1-M3 commits and before the correction commit, `E-020i` is red on D alone (`0 1 0 1 0` per side, exit 1): the transition-bullet pair is listed and not yet applied.
- Green path: M1-M3 apply the manifest rows to both sides. After every milestone commit the first, third and fifth numbers are `0` for every file (nothing outside the list was touched; no code fence moved) and the second and fourth numbers fall to `0` for the files whose rows that milestone has finished; after M3 all five are `0` on every row and the command exits 0. The simulation of the planned text printed `0 0 0 0` on all sixteen files (observed, round 2, four-number form); in the ceiling-exception round the copies rebuilt from the manifest rows alone, listed rows at a non-planned position, printed `0 0 0 0 0` on all sixteen files against the five-number checker (observed). D28 correction round, against the widened fifth number: copies rebuilt from the manifest rows with the listed rows appended last, prepended first, and written where the first removed row stood each printed `0 0 0 0 0` on all sixteen files, exit 0 (observed); `E-020h` carries the same good path for C in a re-executable cell (`good 0 0`).
- The path cell: `E-020b` (`unexpected-paths 0`, exit 0 — green by design before any edit, so it is a regression-guard cell; on the closing tree it stays green only because `CHANGELOG.md` is admitted, see below) and its known failing inputs `E-020c` (the same command against an old base commit: `unexpected-paths 199`, exit 1), `E-020f` (an edit inside the tracked projection `.moai/policies/workflow/factory-dispatch-detail.md`: `unexpected-paths 1`, exit 1) and `E-020g` (an edit inside the tracked projection `.moai/workflows/gtd.md`: `unexpected-paths 1`, exit 1) — the two negative controls were added with the narrowed exemption in the ceiling-exception round (defect D23 of the iteration 2 audit), and the green control (simulated paths under the four writable prefixes, one each in `.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/`, `.moai/reports/`, `.moai/state/`, `.moai/logs/`) printed `unexpected-paths 0`, exit 0. The `CHANGELOG.md` admission (sync-audit amendment 2, S2 of the sync audit): before it, the path half could never be green on the closing tree, because every Tier L SPEC's sync commit adds a `CHANGELOG.md` `[Unreleased]` entry — on `120f9eeb6e37fefc9b2e3673b3c4a967eba121ea` `E-020b` printed `unexpected-paths 1` / `CHANGELOG.md`, exit 1, a criterion no correct work could satisfy (`verification-completeness.md` §2, the impossible direction). The admission is one exact path, not a prefix: `E-020j` (green control, `-- CHANGELOG.md`: `unexpected-paths 0`, exit 0), `E-020k` (`-- CHANGELOG.md.bak`: `unexpected-paths 1`, exit 1) and `E-020l` (`-- docs/CHANGELOG.md`: `unexpected-paths 1`, exit 1). Mutants of the admission tried on scratch copies of the check, each killed by exactly one of the three cells: `CHANGELOG.md` as a prefix of the writable tuple (`E-020k` prints `unexpected-paths 0`, exit 0), admission by base name (`E-020l` prints `unexpected-paths 0`, exit 0), and no admission (`E-020j` prints `unexpected-paths 1`, exit 1); 3 mutants, 3 red, counted apart from the 80 document mutants below. The admission is a path admission only: an edit to `CHANGELOG.md` outside this SPEC's `[Unreleased]` entry is not seen by this criterion (the sync audit reads the CHANGELOG). The hash cell: `E-020d` (`same same`, exit 0, green by design) and its known failing input `E-020e` (one hex digit of the manifest value changed: `DIFFERENT same`, exit 1). The fence cell: `E-020h` (known failing input of the fifth number, re-executable on the current tree: C rebuilt in memory from its base and its manifest rows, then the closing code fence moved to the end together with its preceding line `moai graph build && moai graph query --milestones-no-card` — the pair-move of defect D28 of the iteration 3 audit; prints `C good 0 0 pair-move 44 44`, exit 1).
- Mutants tried (observed, round 2, each on a copy of the simulated post-run text with both sides changed unless noted; every one exits 1 and shows a non-zero first, second, third, fourth or fifth number): the 21 mutants of the AC-MJ-006 / AC-MJ-005 family of the iteration 1 audit (the author's 16, the stem-less edit of the Routing parent line, and the four stem-less additions above); the 37 mutants of the earlier cross-criterion list (one or more per criterion AC-MJ-001 to -019); and 16 further ones — one extra line in each of C, K, F, M, G, W and A, an extra stem-less line before the card-review heading of D, a reworded heading in D, a reworded sentence in C's new paragraph, a listed line deleted, an unrelated base line deleted, an unrelated base line written twice, a trailing space added to a base line, a template-only edit and a deployed-only edit (both flag the side that differs). 74 distinct mutants in all; 74 red. Ceiling-exception round: the checker gained the fifth number and the list gains the fence-move mutant — the existing closing code fence of C moved behind the new Mode judgment paragraph, the structural mutant the iteration 2 audit reported invisible (`0 0 0 0` on the four-number checker) — then `0 0 0 0 1` on the C template row, exit 1; one stem-less addition (an extra Routing sub-bullet in D) re-run through the five-number checker prints `1 0 0 0 0` on the D template row, exit 1, the four-number semantics unchanged. 75 distinct mutants in all; 75 red. D28 correction round: the fifth number compared only fence/predecessor pairs, and four fence mutants passed it with `0 0 0 0 0` on every row, exit 0 (observed before the widening): the closing fence of C moved to the end together with its preceding line (D28), the opening fence of C moved to the end together with its preceding line, the two listed rows of C written inside C's code block, and the closing fence of C moved with its preceding line past one following line. The widened fifth number (segment and inside marks added to the pair marks) prints on the C rows, both sides, `44`, `46`, `2` and `1` respectively, exit 1; a fifth, the closing fence of A's first block moved with its preceding line past one following line, prints `1` on the A rows, exit 1 (it, too, printed `0` before). Re-run through the widened checker with unchanged output: the fence-alone moves (closing fence of C now `45`, opening fence `46`, red as before), and the non-fence mutants — the D stem-less addition `1 0 0 0 0`, one extra line, an unrelated base line deleted (`0 0 1 0 0`), a base line written twice, a trailing space (`1 0 1 0 0`), a listed line deleted (`0 1 0 0 0`), a template-only and a deployed-only edit — each identical to its five-number output before the widening, exit 1; a base line moved across listed rows only stays `0 0 0 0 0`, exit 0 (the disclosed order-blindness). 80 distinct mutants in all; 80 red.
- Does not establish: that a listed line is the right wording (the audit reads the manifest); anything about a non-fence line that is moved without change within its fence-delimited segment (a moved code fence, alone or carried with any run of lines, and a listed row inside a code block are the fifth number's subject); where outside a code block a listed row lands (its position among the other lines is not compared); a code fence that is not three backticks at column 0 (none exists in the sixteen files; a new one is an added line, the first number); that the two non-document files are right (AC-MJ-014); what an edit to the admitted `CHANGELOG.md` says, or whether it stays inside this SPEC's `[Unreleased]` entry. The check compares the working tree, so it also reports an uncommitted edit.

## 2. Evidence ledger

```text
E-001   (AC-MJ-001: eight tokens inside the Mode judgment subsection, both trees)
command: awk 'FNR==1{ if(NR>1){ s=a+b+c+d+e+g+h+i; print s; if(s!=8) bad=1 } a=b=c=d=e=g=h=i=0; f=0 } /^### Mode judgment/{f=1;next} /^##/{f=0} f&&/expected file set/{a=1} f&&/in-flight card/{b=1} f&&/same files or shared state/{c=1} f&&/slot-gated surface/{d=1} f&&/conflict check/{e=1} f&&/classification-file/{g=1} f&&/names the conflicting card/{h=1} f&&/reissued with an explicit classification/{i=1} END{ s=a+b+c+d+e+g+h+i; print s; if(s!=8) bad=1; exit bad }' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
stdout:
0
0
exit: 1
```

```text
E-002a   (AC-MJ-002 c1: violating-sentence sweep V over sweep set R (exit 1 when it prints a line))
command: awk -v re='moai (todo|gtd) add "' '/^[^|]/ && $0 ~ re && !/--classification-file/ { print FILENAME ":" FNR ":" $0; bad=1 } END{ exit bad }' internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:52:`moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`.
exit: 1
```

```text
E-002b   (AC-MJ-002 c2: control — the same predicate over the base commit (git grep boolean form))
command: git grep -n -E -e '^[^|].*moai (todo|gtd) add "' --and --not -e '--classification-file' 74ad197380a0711cdb9816d4e37c94d29af730ea -- internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
stdout:
74ad197380a0711cdb9816d4e37c94d29af730ea:internal/template/templates/.claude/skills/moai/workflows/gtd.md:52:`moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`.
exit: 0   (git grep exits 0 exactly when it selected a line, and it printed one; the guard refuses a trailing exit probe on a git command, so this is stated, not probed)
```

```text
E-002c   (AC-MJ-002 c3: the swept set)
command: git ls-files -- internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
stdout:
internal/template/templates/.claude/agents/moai/manager-lead.md
internal/template/templates/.claude/agents/moai/manager-todo.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
internal/template/templates/.claude/skills/moai/workflows/gtd.md
exit: 0   (git ls-files exits 0 when it lists; the guard refuses a trailing exit probe on a git command)
```

```text
E-002d   (AC-MJ-002 c4: positive presence on the todo skill)
command: grep -c -E '^[^|].*moai todo add ".*--classification-file' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/skills/moai/workflows/gtd.md:0
exit: 1
```

```text
E-003   (AC-MJ-003: leader clause requires the flag and cites C)
command: grep -c -E '^\[HARD\].*--classification-file.*factory-dispatch-cards\.md' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
exit: 1
```

```text
E-004a   (AC-MJ-004 (a): C carries the Class C sentence)
command: grep -c -F 'cannot be added after assign' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md:0
.claude/rules/moai/workflow/factory-dispatch-cards.md:0
exit: 1
```

```text
E-004b   (AC-MJ-004 (b): the unqualified fragment must be gone)
command: grep -c -F 'An empty `spec_id` on a picked card is' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:1
.claude/skills/moai/workflows/gtd.md:1
exit: 0
```

```text
E-004c   (AC-MJ-004 (c): the two pinned lines of the spec_id bullet)
command: grep -c -x -F -e '  `next --spec` is explicitly given. A Class C card is picked with `--spec`.' -e '  On a Class A or Class B card an empty `spec_id` is the normal state; it is' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/skills/moai/workflows/gtd.md:0
exit: 1
```

```text
E-004d   (AC-MJ-004 (d): lines carrying "normal state" (guard against a second claim))
command: grep -c -F 'normal state' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:1
.claude/skills/moai/workflows/gtd.md:1
exit: 0
```

```text
E-004e   (AC-MJ-004 (e): the two pinned lines of the pick step)
command: grep -c -x -F -e '   Attach the SPEC only when one exists and its identifier is known; a Class C' -e '   card is picked with `--spec` (`factory-dispatch-cards.md` § Mode judgment).' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/skills/moai/workflows/gtd.md:0
exit: 1
```

```text
E-005a   (AC-MJ-005 (a))
command: grep -c -F 'no card is actually executing' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
.claude/rules/moai/workflow/factory-dispatch-detail.md:0
exit: 1
```

```text
E-005b   (AC-MJ-005 (b))
command: grep -c -F "free means the lane's previous card reached" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:1
.claude/rules/moai/workflow/factory-dispatch-detail.md:1
exit: 0
```

```text
E-005c   (AC-MJ-005 (c))
command: grep -c -F 'not a free lane' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
.claude/rules/moai/workflow/factory-dispatch-detail.md:0
exit: 1
```

```text
E-005d   (AC-MJ-005 (d): the three pinned lines of the Routing bullet)
command: grep -c -x -F -e '- **Routing.** The leader routes a card WHOLE to a free lane. A lane busy on a card is not addressed; with every lane busy, the card waits in the queue rather than being dispatched. The address block carries `cmd`, which names the entry stage the class prescribes (`/moai plan` for C, `/moai run` for B, the direct close for A), and the lane proceeds through the remaining stages without further dispatches.' -e '  - A free lane is a lane on which no card is actually executing; a lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is a free lane.' -e '  - Free is a different condition from live: a lane with no live session is not a free lane but a fault (`factory-dispatch-mechanics.md` § Boundaries), and a live lane with a card executing is not free.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
.claude/rules/moai/workflow/factory-dispatch-detail.md:0
exit: 1
```

```text
E-005e   (AC-MJ-005 (e): per copy, lines tying the worktree exit or the next card to `done` alone, and the transition bullet with the inclusion and the re-entry (exit 1 unless `0 1` per file; the alternation is passed with -v re=, see §0 Tools); pin: tree 88bc5bcfb41a31afb9944ea76e1bd7b9748a1ee7, measured in the sync-audit amendment round)
command: awk -v re='ExitWorktree[(]keep[)]|next card' 'FNR==1 && NR>1 {print n+0, p+0; bad += (n > 0) + (p != 1); n = 0; p = 0} $0 ~ re && /done/ && !/merge-ready/ {n++} /^- \*\*The transition boundary is between cards\.\*\*/ && /merge-ready, held \(operator-wait\), or serial-waiting/ && /next --card/ {p++} END {print n+0, p+0; bad += (n > 0) + (p != 1); exit (bad > 0)}' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
1 0
1 0
exit: 1
```

```text
E-006   (AC-MJ-006: the eight pinned lines of the delegated-dispatch subsection, its end marker, and the stem scan (exit 1 unless 8 8 1 1 per file; the alternation is passed with -v re=, see §0 Tools))
command: awk -v re='promot|queued|parallelizable' 'BEGIN{ g[1]="Promotion remains an act of the operator. One further case is stated here, besides the `--auto` batch approval and a lane taking a lease itself: while the operator has delegated dispatch (operator away, or a standing delegation), the leader itself promotes queued cards to `picked` and assigns them, with no per-card operator pick. This section states every limit of that case, once."; g[2]="- Where: only to a free live lane, as defined under Routing above."; g[3]="- Which cards: only queued cards explicitly classified `parallelizable` by the mode judgment recorded at issuance (`factory-dispatch-cards.md` § Mode judgment)."; g[4]="- Order: queue priority and readiness first."; g[5]="- Never promoted: a keep-set card, a held card, or a [보류-marked card (the keep-set is defined in `auto-semantics.md` §9.3)."; g[6]="- Serial cards: outside this case; a serial card follows the slot rules (`factory-dispatch-mechanics.md` § The leader\047s remaining role)."; g[7]="- Unchanged: lane self-dispatch through `moai factory next`."; g[8]="Consequence, not a further limit: the assignment input of a free live lane is the picked cards plus, only while dispatch is delegated, the queued `parallelizable` cards."; for(i=1;i<=8;i++) want[g[i]]=1 } FNR==1{ if(NR>1){ print m" "n" "e" "k; if(!(m==8&&n==8&&e==1&&k==1)) bad=1 } m=n=e=k=0; f=0; split("",seen) } /^### Delegated-dispatch promotion$/{f=1;next} f&&/^## /{ f=0; e=($0=="## The card-review stage") } f&&NF{ n++; if(($0 in want)&&!seen[$0]++) m++; next } !f&&tolower($0)~re{ k++ } END{ print m" "n" "e" "k; if(!(m==8&&n==8&&e==1&&k==1)) bad=1; exit bad }' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
stdout:
0 0 0 1
0 0 0 1
exit: 1
```

```text
E-007   (AC-MJ-007: the four pinned lines of the lease-renewal subsection and its end marker (exit 1 unless 4 4 1 per file))
command: awk 'BEGIN{ g[1]="Executing means the lane is carrying out a stage of the card — plan, run, sync, the card-review, or the landing sequence — and is not waiting on a slot, a trigger, or the operator."; g[2]="A lane renews the lease of the card it is executing, and of no other card: only the card the lane is actually executing is renewed."; g[3]="A lease on a merge-ready, held (operator-wait), or serial-waiting card may expire; expiry is harmless, because at the landing slot the leader reassigns the card or the lane re-leases it with `factory next --card`, the stage preserved."; g[4]="The lane\047s standing recheck cron is kept while the lane executes the card; a lane waiting on a slot or a trigger arms and keeps no polling cron for that wait, and the leader sends a notice when a slot or a landing turn opens."; for(i=1;i<=4;i++) want[g[i]]=1 } FNR==1{ if(NR>1){ print m" "n" "e; if(!(m==4&&n==4&&e==1)) bad=1 } m=n=e=0; f=0; split("",seen) } /^### Lease renewal/{f=1;next} f&&/^##/{ f=0; e=($0=="## Isolation — launcher, tiers, branch naming, and the traceability carriers") } f&&NF{ n++; if(($0 in want)&&!seen[$0]++) m++ } END{ print m" "n" "e; if(!(m==4&&n==4&&e==1)) bad=1; exit bad }' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
stdout:
0 0 0
0 0 0
exit: 1
```

```text
E-008a   (AC-MJ-008 (a): unscoped keep-statements (exit 1 when it prints a line))
command: awk '(/(kept|keeps|keep it|stays armed|stay armed|armed)/) && /(completion report|standing recheck|standing carrier|recheck cron|CronCreate|[*][*]standing[*][*]|standing wake)/ && !/while the lane executes the card/ && !/^Version:/ { print FILENAME ":" FNR ":" substr($0,1,90); bad=1 } END{ exit bad }' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md .claude/skills/moai-lane-watchdog/SKILL.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:28:[HARD] A stuck lane records a disk wait and ends its turn — never idles on a reply; on a
.claude/rules/moai/workflow/factory-dispatch-core.md:28:[HARD] A stuck lane records a disk wait and ends its turn — never idles on a reply; on a
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:119:[HARD] A stuck lane writes a disk wait (reason, whom, recheck point), ends its turn, never
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:122:[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, reche
.claude/rules/moai/workflow/factory-dispatch.md:119:[HARD] A stuck lane writes a disk wait (reason, whom, recheck point), ends its turn, never
.claude/rules/moai/workflow/factory-dispatch.md:122:[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, reche
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:117:therefore **standing**: armed once at card intake, before the first stage, and
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:118:kept until the completion report.
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:396:once it resolves; the standing carrier of §5.1 stays armed. The codex runner
.claude/rules/moai/workflow/auto-semantics.md:117:therefore **standing**: armed once at card intake, before the first stage, and
.claude/rules/moai/workflow/auto-semantics.md:118:kept until the completion report.
.claude/rules/moai/workflow/auto-semantics.md:396:once it resolves; the standing carrier of §5.1 stays armed. The codex runner
internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md:244:   fires the awaken is the lane's standing recheck cron (doctrine §5.1), armed
.claude/skills/moai-lane-watchdog/SKILL.md:244:   fires the awaken is the lane's standing recheck cron (doctrine §5.1), armed
exit: 1
```

```text
E-008b   (AC-MJ-008 (b): the six site fragments, counted per file)
command: grep -c -F -e 'is kept while the lane executes the card.' -e 'recheck armed before the first stage, kept while the lane executes the card;' -e 'keeps it while the lane executes the card:' -e 'and kept while the lane executes the card.' -e 'is kept only while the lane executes the card' -e 'armed at card intake and kept while the lane executes the card' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md .claude/skills/moai-lane-watchdog/SKILL.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:0
.claude/rules/moai/workflow/factory-dispatch-core.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:0
.claude/rules/moai/workflow/auto-semantics.md:0
internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md:0
.claude/skills/moai-lane-watchdog/SKILL.md:0
exit: 1
```

```text
E-008c   (AC-MJ-008 (c): renewal-period wording retained (regression-guard cell))
command: grep -c -F 'every renewal period of `workflow.factory_lease.renew_interval`' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:1
.claude/rules/moai/workflow/factory-dispatch-core.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
exit: 0
```

```text
E-008d   (AC-MJ-008 (d): the delete-on-stop sentence in A)
command: grep -c -F 'when the lane stops executing the card' internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:0
.claude/rules/moai/workflow/auto-semantics.md:0
exit: 1
```

```text
E-008e   (AC-MJ-008 (e): the two whole lines of the fire-and-renew sentence in A (exit 1 unless 2 per file))
command: awk -v a='  first (§0a of the watchdog skill), which renews the lease of the card the lane' -v b='  is executing, so the renewal period is also the cadence at which that lease' 'FNR==1{ if(NR>1){ print n; if(n!=2) bad=1 } n=0 } { n += ($0==a) + ($0==b) } END{ print n; if(n!=2) bad=1; exit bad }' internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md
stdout:
0
0
exit: 1
```

```text
E-008f   (AC-MJ-008 (f): the opening of the renew paragraph in M (exit 1 unless 1 per file))
command: awk -v s='While it executes the card, a lane renews its own card lease at the start of each watchdog pass with `moai factory renew <card>`' 'FNR==1{ if(NR>1){ print n; if(n!=1) bad=1 } n=0 } index($0,s){ n++ } END{ print n; if(n!=1) bad=1; exit bad }' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
stdout:
0
0
exit: 1
```

```text
E-009   (AC-MJ-009: pinned scope line, renew call, and "every card" count inside section 0a (exit 1 unless 1 >=1 0 per file; the alternation is passed with -v re=))
command: awk -v re='every card|each card' 'BEGIN{ w="This step renews the card the lane is executing and no other card: a lane whose cards are all merge-ready, held, or waiting for the serial slot does not renew in that pass, and the pass carries the previous `renew_result` forward." } FNR==1{ if(NR>1){ print p" "r" "q; if(!(p==1&&r>=1&&q==0)) bad=1 } p=r=q=0; f=0 } /^## 0a/{f=1;next} /^## /{f=0} f&&$0==w{p++} f&&/moai factory renew/{r++} f&&$0~re{q++} END{ print p" "r" "q; if(!(p==1&&r>=1&&q==0)) bad=1; exit bad }' internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md .claude/skills/moai-lane-watchdog/SKILL.md
stdout:
0 1 0
0 1 0
exit: 1
```

```text
E-010a   (AC-MJ-010 (a): recursive sweep for the reclassification verb)
command: grep -r -l -E --exclude-dir=agent-memory --exclude-dir=worktrees 'moai (todo|gtd|factory) reclassif|reclassify verb' internal/template/templates/.claude .claude
stdout: (empty)
exit: 1
```

```text
E-010b   (AC-MJ-010 (b): the factual default-decider sentence)
command: grep -c -F 'Without it the deterministic default decider records `priority=normal, blocked=false, mode=serial, decider=default`' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:1
.claude/skills/moai/workflows/gtd.md:1
exit: 0
```

```text
E-010c   (AC-MJ-010: known failing input for E-010a)
command: grep -c -E 'moai (todo|gtd|factory) reclassif|reclassify verb' .moai/reports/t1894/premise-check.md
stdout:
1
exit: 0
```

```text
E-011   (AC-MJ-011: the seven mirrored pairs, one cmp invocation each)
commands:
  cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
  cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
  cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
  cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
  cmp internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md
  cmp internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
  cmp internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md .claude/skills/moai-lane-watchdog/SKILL.md
stdout: (empty for each)
exit: 0, 0, 0, 0, 0, 0, 0   (in the order above)
```

```text
E-011f   (AC-MJ-011: known failing input — the declared fork)
command: cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md differ: char 23127, line 222
exit: 1
```

```text
E-011g   (AC-MJ-011: known failing input for the pair instrument — a historical commit that changed one side only)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py pairs .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt 957967601351fccb7723713625b41f4d6e05b386
stdout:
9579676013 pairs 0 unpaired 7
  .claude/rules/moai/workflow/factory-dispatch-cards.md
  .claude/rules/moai/workflow/factory-dispatch-core.md
  .claude/rules/moai/workflow/factory-dispatch.md
  .claude/rules/moai/workflow/factory-dispatch-detail.md
  .claude/rules/moai/workflow/factory-dispatch-mechanics.md
  .claude/rules/moai/workflow/auto-semantics.md
  .claude/skills/moai/workflows/gtd.md
exit: 1
```

```text
E-011h   (AC-MJ-011: known passing input for the pair instrument — a historical commit that changed two pairs on both sides, expected count 2)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py pairs .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt 1969709ce693dc4975754a3b1ad1447e56b4ccc5:2
stdout:
1969709ce6 pairs 2 unpaired 0 expected 2
exit: 0
```

```text
E-012a   (AC-MJ-012: UTF-16 code units of the core stub (exit 1 above 10000))
command: python3 -I -c "import sys; n=len(open(sys.argv[1],encoding='utf-8').read().encode('utf-16-le'))//2; print('OVER' if n>10000 else 'OK', n); sys.exit(1 if n>10000 else 0)" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
OK 9999
exit: 0
```

```text
E-012b   (AC-MJ-012: known failing input)
command: python3 -I -c "import sys; n=len(open(sys.argv[1],encoding='utf-8').read().encode('utf-16-le'))//2; print('OVER' if n>10000 else 'OK', n); sys.exit(1 if n>10000 else 0)" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
stdout:
OVER 25803
exit: 1
```

```text
E-013a   (AC-MJ-013: leak pattern per file, baselined card ids excluded)
command: grep -c -E -e '(^|[^A-Za-z0-9_])t([03-57-9][0-9][0-9]([0-9])?|1([0-24-9][0-9]([0-9])?|3([0-24-8]([0-9])?|3[0-8]|9([0-24-9])?))|2([013-9][0-9]([0-9])?|2([0-35-9]([0-9])?|4[0-9]))|6([0-8][0-9]([0-9])?|9([0-57-9]([0-9])?|6[0-9])))([^A-Za-z0-9_]|$)|SPEC-[A-Z][A-Z0-9-]*-[0-9][0-9][0-9]|d-2026[0-9]|20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]|(^|[^0-9a-f])[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+([^0-9a-f]|$)|(^|[^0-9a-z])(a672|0c51|02f4)([^0-9a-z]|$)' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:0
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:0
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md:0
exit: 1
```

```text
E-013b   (AC-MJ-013: known failing input for E-013a)
command: grep -c -E -e '(^|[^A-Za-z0-9_])t([03-57-9][0-9][0-9]([0-9])?|1([0-24-9][0-9]([0-9])?|3([0-24-8]([0-9])?|3[0-8]|9([0-24-9])?))|2([013-9][0-9]([0-9])?|2([0-35-9]([0-9])?|4[0-9]))|6([0-8][0-9]([0-9])?|9([0-57-9]([0-9])?|6[0-9])))([^A-Za-z0-9_]|$)|SPEC-[A-Z][A-Z0-9-]*-[0-9][0-9][0-9]|d-2026[0-9]|20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]|(^|[^0-9a-f])[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+([^0-9a-f]|$)|(^|[^0-9a-z])(a672|0c51|02f4)([^0-9a-z]|$)' .moai/reports/t1894/premise-check.md
stdout:
8
exit: 0
```

```text
E-013c   (AC-MJ-013: the five baselined literals are present (and not counted by E-013a))
command: grep -c -E -e '(^|[^A-Za-z0-9_])t(133|224|696|1339|1393)([^A-Za-z0-9_]|$)' internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md:2
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md:2
internal/template/templates/.claude/skills/moai/workflows/gtd.md:1
exit: 0
```

```text
E-016   (AC-MJ-016: files carrying a home token)
command: grep -l -F -e "slot-gated surface" -e "no card is actually executing" -e "only the card the lane is actually executing" -e "expiry is harmless" -e "cannot be added after assign" -e "while the operator has delegated dispatch" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md internal/template/templates/.claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md
stdout: (empty)
exit: 1
```

```text
E-017   (AC-MJ-017: the declared fork stays one differing line)
command: git diff --no-index --shortstat internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
 1 file changed, 1 insertion(+), 1 deletion(-)
exit: 1   (observed as non-zero by a wrapper outside the cell, `|| echo exit-nonzero`; git diff --no-index exits 1 when the files differ)
```

```text
E-018a   (AC-MJ-018 (a): the pointer line of F, whole line)
command: grep -c -x -F '[HARD] **The delegated-dispatch case.** Besides the `--auto` batch approval, the leader may promote queued cards in the case stated in `factory-dispatch-detail.md` § Delegated-dispatch promotion, with the limits given there; nothing in this section widens it.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
exit: 1
```

```text
E-018b   (AC-MJ-018 (b): the old opening clause, F and K, both trees)
command: grep -c -F 'The one reconciliation is named' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:1
.claude/rules/moai/workflow/factory-dispatch-core.md:1
exit: 0
```

```text
E-018c   (AC-MJ-018 (c): the replacement clause, F and K, both trees)
command: grep -c -F 'A reconciliation is named, not excepted:' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:0
.claude/rules/moai/workflow/factory-dispatch.md:0
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md:0
.claude/rules/moai/workflow/factory-dispatch-core.md:0
exit: 1
```

```text
E-018d   (AC-MJ-018 (d): the pointer line of G, whole line)
command: grep -c -x -F 'Promotion by the leader without a pick is limited to the case stated in `factory-dispatch-detail.md` § Delegated-dispatch promotion; every pick the operator makes follows the rule above.' internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md
stdout:
internal/template/templates/.claude/skills/moai/workflows/gtd.md:0
.claude/skills/moai/workflows/gtd.md:0
exit: 1
```

```text
E-018e   (AC-MJ-018 (e): the enumeration row of M, whole line)
command: grep -c -x -F '| Delegated-dispatch promotion | promoting queued cards to a free live lane without a per-card pick, in the case and within the limits stated once in `factory-dispatch-detail.md` § Delegated-dispatch promotion |' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:0
.claude/rules/moai/workflow/factory-dispatch-mechanics.md:0
exit: 1
```

```text
E-019a   (AC-MJ-019: pinned prohibition sentence in F)
command: grep -c -F 'Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
exit: 0
```

```text
E-019b   (AC-MJ-019: known failing input — the same sentence over the core stub)
command: grep -c -F 'Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md
stdout:
0
exit: 1
```

```text
E-019c   (AC-MJ-019: pinned admission sentence in F)
command: grep -c -F "queue ADMISSION (production) stays the operator's." internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md:1
.claude/rules/moai/workflow/factory-dispatch.md:1
exit: 0
```

```text
E-020a   (AC-MJ-020 (a): added and removed lines against the manifest, per file and per side (exit 1 unless every number is 0); sync-audit amendment round: the whole cell re-measured against the amended manifest on the sixteen documents of the base commit extracted by `git archive d31fb64239a688c053e455000cc9fe37f540b790` to a scratch root and read by the same check's optional root operand (`lines <manifest> <root>`), because the working tree now carries the run; every row but D is unchanged, D was `0 12 0 1 0` per side)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py lines .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt
stdout:
C template 0 2 0 0 0 deployed 0 2 0 0 0
K template 0 2 0 2 0 deployed 0 2 0 2 0
F template 0 5 0 3 0 deployed 0 5 0 3 0
D template 0 13 0 2 0 deployed 0 13 0 2 0
M template 0 7 0 1 0 deployed 0 7 0 1 0
A template 0 8 0 8 0 deployed 0 8 0 8 0
G template 0 10 0 8 0 deployed 0 10 0 8 0
W template 0 2 0 2 0 deployed 0 2 0 2 0
exit: 1
```

```text
E-020b   (AC-MJ-020 (b): changed or new paths that the manifest does not name, that are not CHANGELOG.md, and that are not under the four writable prefixes; sync-audit amendment 2: re-measured on the closing tree, which carries the sync commit's CHANGELOG.md entry — before the admission the same command printed `unexpected-paths 1` / `CHANGELOG.md`, exit 1; pin: tree 120f9eeb6e37fefc9b2e3673b3c4a967eba121ea with this round's check edit applied, the check's SHA-256 as in E-020d)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt
stdout:
unexpected-paths 0
exit: 0
```

```text
E-020c   (AC-MJ-020 (b): known failing input — the same command against an old base commit; re-measured in the ceiling-exception round, 175 under the former whole-`.moai/` exemption, 200 under the narrowed one; re-measured in sync-audit amendment 2, 200 → 199, because `CHANGELOG.md` differs from that old commit and is now admitted — the first eight printed paths are unchanged; pin as E-020b)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt 3642a3a9b9cf101863656153d2d99b69fcb66a05
stdout:
unexpected-paths 199
.claude/agents/moai/manager-git.md
.claude/agents/moai/manager-todo.md
.claude/rules/local/gitflow-lane-protocol.md
.claude/rules/moai/workflow/delivery-policy.md
.claude/rules/moai/workflow/spec-workflow.md
.claude/rules/moai/workflow/worktree-integration-ops.md
.claude/rules/moai/workflow/worktree-integration.md
.claude/skills/moai-factory-foreman/SKILL.md
exit: 1
```

```text
E-020d   (AC-MJ-020 (c): the manifest and the check are the files the plan phase froze (SHA-256 pinned here because neither is in the plan-artifact hash subject set); the check's value re-measured in the ceiling-exception round, when the fifth number and the narrowed path exemption changed the check and not the manifest, and again in the D28 correction round, when the fifth number gained the segment and inside marks; the manifest value re-pinned in the mid-run amendment round, when the F reconciliation row gained one trailing pointer sentence, from `c2b0ef170297e49d191712e6e433cd16fb856e2da5abbe2eb0b6f444c28a2c08`, and again in the sync-audit amendment round, when section D gained the transition-bullet pair, from `d0aceca49ec28f9bcad299776f84c23b1bf646542cd9b5412800a2575fd4d63c`; the check's value re-pinned in sync-audit amendment 2, when `paths` gained the exact `CHANGELOG.md` admission, from `d8df759fb641518e1838ee1553f5c5935b895d39eac95dd65abc7e1fa6a161e3`; the manifest's value unchanged in that round)
command: python3 -I -c "import hashlib,sys; w=[('.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt','7eb04601e1678958e74c69f7085bb6fc20c9726da6a473fc53f1a920afad099f'),('.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py','22a55ded775d72199937adf1f8568a28ba94802558a010da59e5688389058caa')]; r=[hashlib.sha256(open(p,'rb').read()).hexdigest()==h for p,h in w]; print(*['same' if x else 'DIFFERENT' for x in r]); sys.exit(0 if all(r) else 1)"
stdout:
same same
exit: 0
```

```text
E-020e   (AC-MJ-020 (c): known failing input — one hex digit of the manifest value changed)
command: python3 -I -c "import hashlib,sys; w=[('.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt','7eb04601e1678958e74c69f7085bb6fc20c9726da6a473fc53f1a920afad099e'),('.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py','22a55ded775d72199937adf1f8568a28ba94802558a010da59e5688389058caa')]; r=[hashlib.sha256(open(p,'rb').read()).hexdigest()==h for p,h in w]; print(*['same' if x else 'DIFFERENT' for x in r]); sys.exit(0 if all(r) else 1)"
stdout:
DIFFERENT same
exit: 1
```

```text
E-020f   (AC-MJ-020 (b): known failing input — an edit inside the tracked projection .moai/policies/workflow/factory-dispatch-detail.md, judged by the explicit path list after `--` (added with the narrowed exemption, ceiling-exception round))
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt -- .moai/policies/workflow/factory-dispatch-detail.md
stdout:
unexpected-paths 1
.moai/policies/workflow/factory-dispatch-detail.md
exit: 1
```

```text
E-020g   (AC-MJ-020 (b): known failing input — an edit inside the tracked projection .moai/workflows/gtd.md, judged by the explicit path list after `--` (added with the narrowed exemption, ceiling-exception round))
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt -- .moai/workflows/gtd.md
stdout:
unexpected-paths 1
.moai/workflows/gtd.md
exit: 1
```

```text
E-020h   (AC-MJ-020 (a): known failing input of the fifth number — C, both sides, rebuilt in memory from the base blob and the manifest rows (good), then its closing code fence moved to the end together with its preceding line (pair-move, defect D28 of the iteration 3 audit); the check's own functions, read-only (added in the D28 correction round; exit 1 when the pair-move is seen))
command: python3 -I -c "import importlib.util as u,sys; p='.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/'; sp=u.spec_from_file_location('c',p+'closed-world-check.py'); c=u.module_from_spec(sp); sp.loader.exec_module(c); b,st,fs,_=c.load_manifest(p+'closed-world-manifest.txt'); f=fs[0]; L=f['plus']|f['minus']; mv=lambda W:(lambda i:W[:i-1]+W[i+1:]+W[i-1:i+1])([k for k,l in enumerate(W) if l.startswith(chr(96)*3)][1]); side=lambda B:(lambda W:[c.fence_moves(W,B,L),c.fence_moves(mv(W),B,L)])([x for x in B if x not in f['minus']]+list(f['plus'].elements())); r=[side(c.nonblank(c.git('show',b+':'+q).decode('utf-8'))) for q in (f['path'],f['path'][len(st):])]; print(f['k'],'good',r[0][0],r[1][0],'pair-move',r[0][1],r[1][1]); sys.exit(1 if r[0][1] or r[1][1] else 0)"
stdout:
C good 0 0 pair-move 44 44
exit: 1
```

```text
E-020i   (AC-MJ-020 (a): the same command on the tree the run left behind, against the amended manifest — red on D alone because the transition-bullet pair is listed and not yet applied; flips to all `0 0 0 0 0`, exit 0, with the correction commit; pin: tree 88bc5bcfb41a31afb9944ea76e1bd7b9748a1ee7, measured in the sync-audit amendment round)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py lines .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt
stdout:
C template 0 0 0 0 0 deployed 0 0 0 0 0
K template 0 0 0 0 0 deployed 0 0 0 0 0
F template 0 0 0 0 0 deployed 0 0 0 0 0
D template 0 1 0 1 0 deployed 0 1 0 1 0
M template 0 0 0 0 0 deployed 0 0 0 0 0
A template 0 0 0 0 0 deployed 0 0 0 0 0
G template 0 0 0 0 0 deployed 0 0 0 0 0
W template 0 0 0 0 0 deployed 0 0 0 0 0
exit: 1
```

```text
E-020j   (AC-MJ-020 (b): green control of the CHANGELOG.md admission — the exact repository-root path, judged by the explicit path list after `--`; added in sync-audit amendment 2; the pre-amendment check printed `unexpected-paths 1` / `CHANGELOG.md`, exit 1; pin as E-020b)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt -- CHANGELOG.md
stdout:
unexpected-paths 0
exit: 0
```

```text
E-020k   (AC-MJ-020 (b): known failing input — a path that starts with CHANGELOG.md is not admitted (kills a prefix admission); added in sync-audit amendment 2; pin as E-020b)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt -- CHANGELOG.md.bak
stdout:
unexpected-paths 1
CHANGELOG.md.bak
exit: 1
```

```text
E-020l   (AC-MJ-020 (b): known failing input — a CHANGELOG.md outside the repository root is not admitted (kills a base-name admission); added in sync-audit amendment 2; pin as E-020b)
command: python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py paths .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-manifest.txt -- docs/CHANGELOG.md
stdout:
unexpected-paths 1
docs/CHANGELOG.md
exit: 1
```

## 3. Edge cases

- **A ledger row is stale after the rewrite.** If M2 rewrites F:119 and leaves ALB-0276 untouched, `TestBindingLedgerIntegrity` reports "after-text absent … obligation lost"; the row's `after_text` is updated in the same commit (REQ-MJ-016). The same holds for ALB-0257 and the K:20 / F:44 clause.
- **A binding line copied into an on-demand file.** `TestBindingLedgerIntegrity` (AC-ALB-021(2)) fails when a line of 40 or more UTF-16 units of a binding or normative row's `after_text` is also a whole line of an on-demand (`paths:`-scoped) rule file, unless that line already stood in the file's copy at the ledger anchor. The run observed it: with F's reconciliation paragraph byte-identical to the K line that ALB-0257 covers, the guard reported row ALB-0257 in `factory-dispatch.md` (`.moai/reports/t1894/verdict.md` §2.4). The manifest's F row therefore ends the paragraph with one pointer sentence that K does not carry; the opening clause E-018b/E-018c count is unchanged.
- **The core stub exceeds 10,000 units by a few characters.** AC-MJ-012 prints `OVER` and exits 1; the planned wording nets −2 when both K edits land in the same commit, so an overrun means a wording drift (`design.md` §D-4).
- **The exception is folded into a pinned sentence.** Rewording F:38, F:40 or the K:20 / F:44 paragraph's pinned clauses to carry the exception turns AC-MJ-019 red and the Go doctrine tests of AC-MJ-014 red; the exception is added beside them, by pointer.
- **A lane or leader treats an unstated delegation as given.** The ruling names no mechanism that establishes "the operator has delegated dispatch" (`spec.md` R-5, decision-index Q7); AC-MJ-006 pins the text and cannot see the practice.
- **`factory-dispatch.md` edited identically in both trees but the declared fork line also moves.** AC-MJ-017 shows a count other than `1/1`.
- **The per-line trap.** A scoped sentence whose physical tail line still reads "kept until the completion report" keeps AC-MJ-008(a) red; reflow the paragraph so a scoped sentence sits on one physical line. The same holds for every pinned whole line: write it on one physical line, with the leading spaces shown.
- **The card-id baseline.** `auto-semantics.md` §5.1 carries `t1393` and `t1339`; deleting either sentence fails `TestCardIDBaselineHasNoStaleEntries`, and re-wrapping them is safe for AC-MJ-013.
- **A line is added that no stem or token catches.** AC-MJ-005 (d) and AC-MJ-006 print their green numbers (`3`, `8 8 1 1`) with an extra sub-bullet, an extra sentence in the card-review section or an extra sentence before the subsection heading; AC-MJ-020's first number is `1` for the file and the command exits 1. Remove the line, or return a blocker report so that manager-spec decides whether the manifest gains the row.
- **A wording improvement during the run.** The pinned cells and the manifest carry the same lines; the run phase never edits either. The improvement is a blocker report, and manager-spec amends the manifest and the cells in one commit (mid-run authority, D-NEW-1).
- **A reflow that moves a listed line.** The comparison counts physical lines: a hard-wrapped paragraph of `auto-semantics.md` re-wrapped differently from the manifest shows a non-zero first and second number, even when the words are unchanged. Reproduce the manifest rows' line breaks exactly.

## 4. Quality gates and Definition of Done

- All `R` criteria (AC-MJ-001..009, -016, -018, -020) have observed RED cells (above) and pass on the run-phase tree with the command output pasted into the verdict file `.moai/reports/t1894/verdict.md` (decision-bearing lines only, per the evidence-export obligation).
- All `G` criteria (AC-MJ-010..014, -017, -019) pass; AC-MJ-014's Go logs are cited from the verdict file.
- AC-MJ-015 is recorded by the leader and is not a SPEC PASS.
- Decision-index rows Q1, Q2, Q4, Q5, Q6 are stamped `DEFAULT-APPLIED`; Q3 is met by the operator ruling d314 (its verdict is recorded in the row); Q7 was answered by the operator on 2026-10-23, relayed by the leader (its verdict is recorded in the row). No requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final because it carries the ruling's words and no mechanism. Per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition. Q7 was answered on 2026-10-23 (relayed by the leader) with exactly that statement, so it no longer blocks the autonomous Kickoff form.
- Commit shape (graph-verifiable, replaces the former ordering clause): each of M1, M2 and M3 lands a changed template document together with its deployed twin in one commit, with the ledger rows and the catalog hash entry of what it changes in the same commit (AC-MJ-011, run-phase obligation; the `pairs` command decides it and takes the expected pair count of each milestone). Template-First (edit the template side, then bring the deployed copy to the same bytes) is the authoring discipline the repository asks for; the commit graph cannot show the order of two edits inside one commit, so this SPEC does not make it a criterion.
- `moai spec lint --strict` on `spec.md` exits 0 (recorded in `progress.md` §E.1).
