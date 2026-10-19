# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 (revision 0.1.2)

## A. Pin and conventions

- **Document-level pin.** Tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-workflow-cmd-index`, equal to local `main`). Every `git` cell names the commit in its command (`git grep … 74ad197380`, `git ls-tree … 74ad197380`), so the pin is in the cell. Measured on 2026-10-19; every cell of revision 0.1.0 was re-run in revision 0.1.1 and printed the same output.
- **Working-tree cells.** C-3, C-12, R-8, and R-9 read working-tree files, not a git object. The working tree is at `5d16c25af8e250491104f25b6e3f73e6268d4001`, the plan commit, which adds only files under `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001/`. Evidence that `docs-site`, `.claude`, the READMEs, `scripts`, and `.github` equal `74ad197380`: `git diff --stat 74ad197380 HEAD -- docs-site .claude README.md README.ko.md README.ja.md README.zh.md scripts .github` printed nothing, and `git status --short` printed nothing, both before and after the measurements.
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no pipe, redirection, chaining, or subshell. Its verbatim stdout, its exit code, and the tree are separate fields. The shell tool displays no status for a command with no output; a zero-match cell therefore says `no output` and gives the exit code git defines for it (1), marked `by git grep semantics, not displayed` (gap G-1). Cells that print output state exit code 0 because the tool reported no failure.
- **Pathspec.** Every cell addressing the twelve files uses three glob pathspecs: `':(glob)docs-site/content/*/workflow-commands/_index.md'`, `…/_meta.yaml'`, and `…/moai-design.md'`. C-3 prints the twelve paths these select, so an empty or short sweep is visible.
- **Classes.** `release-blocking (gated)`: re-executable on the tree; counts toward the completion verdict once its gate rows carry a verdict. `guard`: green today; its failure is shown by a named mutant. `unadopted`: no RED-now cell can be measured until a decision row has a verdict.
- **Verdict basis.** Q13 and Q17 carry operator verdicts (board record `d-20261019T015908Z-220a`, 2026-10-19). Q14 carries the operator verdict `d-20261019T021508Z-258e` (remove; relayed by the dispatching coordinator, G-10). The four items O-1 to O-4 carry the **leader** ruling `d-20261019T020706Z-1c58` (decided by the leader under operator delegation; not an operator verdict). No decision row is open; the rewrite and marker branches are retired (AC-SWF-006 is superseded).
- **Tool provenance (verification-claim-integrity §2.2).** Only system `git`, the system `awk`, and `yq` v4.53.2 were used for the cells; no `moai` binary measurement is cited in this file (the lint result is recorded in `progress.md` with its build provenance). Pattern cells use `git grep -E` and, for C-2, `git grep -P`.

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Gate rows | Class | RED-now cells | Green-path milestone |
|----|-----|-----------|-------|---------------|----------------------|
| AC-PRC-001 | REQ-PRC-001 | none (Q13, Q14, Q17 decided) | release-blocking | C-1 (C-3 sweep control) | M2 (`_meta.yaml`, page), M3 (`_index.md`) |
| AC-SWF-002 | REQ-PRC-001, REQ-SWF-002 | none | release-blocking | C-2 (C-8 list, C-3 sweep) | M2, M3 |
| AC-SWF-003 | REQ-SWF-002, REQ-SWF-003 | none (Q13, Q17 decided) | release-blocking | C-4 | M3 |
| AC-SWF-004 | REQ-SWF-007 | none | guard | none (green today; mutants in recipe R-4) | M4 |
| AC-SWF-005 | REQ-SWF-004 | none (Q14 decided: remove) | release-blocking | C-5 | M2 |
| AC-SWF-006 | REQ-SWF-005, REQ-SWF-006 (both superseded) | — | SUPERSEDED by the Q14 verdict; not applicable | none | none |
| AC-SWF-007 | REQ-SWF-008 | none (leader ruling O-4; Q13 decided) | release-blocking | C-11 (controls C-10, C-12) | M3 |
| AC-SWF-008 | REQ-SWF-009 | none (Q14 decided: remove) | release-blocking | C-13 (C-5, C-18) | M2 |
| AC-SWF-009 | REQ-SWF-010 | none (Q14 decided: remove) | release-blocking | C-14 (C-15, C-16) | M2 |

Every live REQ has an AC and every live AC maps to a REQ (8 of 8; REQ-SWF-005, -006 and AC-SWF-006 are superseded stubs, so 10 and 9 are numbered).

## B. Acceptance criteria

### AC-PRC-001 — no unresolved design-family citation remains in the twelve files (page part of REQ-PRC-001)

- **Given** 27 lines in the twelve files match the design family (C-1), spread over all twelve files, and the sweep selects exactly twelve files (C-3).
- **When** M2 deletes the four pages and the four `_meta.yaml` entries (Q14 verdict), and M3 deletes the design table rows of the four `_index.md` files (Q13, Q17 verdicts),
- **Then**
  1. The C-1 command run on the closing commit prints nothing and exits 1.
  2. The C-3 command run on the closing commit prints eight paths (the four `_index.md` and the four `_meta.yaml`). A closing run that prints fewer paths than that is a failure of the sweep, not a pass.
  3. The C-1 command run on `74ad197380` still prints the twelve lines of C-1 (positive control: the pattern matches where the work has not been done).
- **RED-now cells:** C-1. Red because 27 lines in twelve files carry the design family; it flips only when this work edits those files.
- **Green path:** M3 flips the five `_index.md` lines (en, ja, zh: 1 each; ko: 2). M2 flips the 4 `_meta.yaml` lines and the 18 page lines (5, 5, 4, 4) by deletion.
- **Mutant probe:** deleting the sentence around a citation instead of the row leaves the C-4 counts unchanged (see AC-SWF-003 Then 1); leaving the design family in one locale only leaves that locale's line in the C-1 output; replacing `/moai design` with another unregistered form (for example `/moai ui`) passes C-1 and fails AC-SWF-002; a spelling that C-1 does not match (two spaces between the words) passes C-1, and C-2 still matches it only if the verb token follows the single separator, so the closing review reads the diff of the twelve files, not the grep alone (gap G-9).
- **Gate:** none (Q14 decided: remove). `main.yaml` has its own criterion (AC-SWF-008).
- **Done when:** Then 1 to 3 hold on the committed tree and the commit SHA is recorded in `progress.md` §E.2 for the parent's M6.

### AC-SWF-002 — every remaining `/moai` slash form in the twelve files resolves (REQ-PRC-001, REQ-SWF-002)

- **Given** the registered commands are the seventeen files of C-8, and the negative-list search over the twelve files returns the same 27 lines as C-1 (C-2).
- **When** the Q13 verdict is applied (replace where an equivalent exists, delete otherwise) and the Q14 verdict is applied (page removed),
- **Then**
  1. The C-2 command run on the closing commit prints nothing and exits 1.
  2. The C-8 command run on the closing commit still prints seventeen paths with the same names (the list in the C-2 pattern is the list on the tree; a registered-command change invalidates the cell and the cell is re-run, not tuned).
  3. No replacement introduced by M2 or M3 is itself an unregistered `/moai` form (this is what Then 1 measures).
- **RED-now cells:** C-2. Red for the stated reason: the only unregistered verb in the twelve files is `design`, on 27 lines.
- **Green path:** the same edits as AC-PRC-001; this criterion additionally rejects a replacement that names another unregistered command.
- **Mutant probe:** replacing the design row with `/moai ui` or `/moai design-phase` is matched by the pattern (the token is not in the list and is not followed by a listed verb boundary); a pattern edited to include `design` in its list would print nothing and is detected by the C-2 positive control at `74ad197380` (the base prints twelve lines).
- **Gate:** none (Q14 decided).
- **Done when:** Then 1 to 3 hold on the committed tree.

### AC-SWF-003 — the four locales receive the same deletions in the same change (REQ-SWF-002, REQ-SWF-003)

- **Given** the per-locale count of `/moai` table rows in `_index.md` is en 21, ja 21, ko 23, zh 21 (C-4). The Korean file holds two design rows (a summary-table row and a subcommand-table row) and one `/moai goal` summary row that the other locales lack.
- **When** M3 deletes the design rows in all four locales (ko canonical first, then en, then ja and zh) in one commit,
- **Then**
  1. The C-4 command run on the closing commit prints exactly: en 20, ja 20, ko 21, zh 20. The ko figure stays one above the others because of the `/moai goal` summary row, which resolves and is not touched; it is not an error.
  2. The C-1 count of `_index.md` lines is zero in all four locales (AC-PRC-001 Then 1).
  3. The commit that makes these edits changes all four `_index.md` files (`git show --stat` lists four paths among the twelve files; recipe R-3).
- **RED-now cells:** C-4. Red because the counts today are 21, 21, 23, 21, not 20, 20, 21, 20; the figures differ from the closing figures by exactly the design rows (1, 1, 2, 1).
- **Green path:** M3 flips C-4 to 20, 20, 21, 20.
- **Mutant probe:** a Korean edit that removes only line 66 leaves ko at 22 and fails Then 1; an edit that deletes the wrong row (for example the `/moai goal` summary row) leaves the design lines and fails AC-PRC-001; an edit that changes three locales and leaves one fails Then 1 and Then 3; replacing a row with a prose sentence keeps the row count but leaves the design family, and is caught by AC-PRC-001; deleting more than the design rows (for example the whole table) drops a count below the closing figure.
- **Gate:** none (Q13 and Q17 decided). The count statements are not table rows and do not change C-4 (they have AC-SWF-007).
- **Done when:** Then 1 to 3 hold on the committed tree.

### AC-SWF-004 — nothing outside the change set changes (REQ-SWF-007) — guard

- **Given** the tree is clean at `74ad197380` and the other six workflow-commands pages per locale hold no design-family line (C-1 lists only the twelve files).
- **When** the whole change is committed,
- **Then**
  1. The R-4 command (`git diff --name-only 74ad197380 <closing-sha> -- docs-site README.md README.ko.md README.ja.md README.zh.md`) prints exactly the fourteen paths of § Scope of spec.md (eight modified content files, `docs-site/data/menu/main.yaml`, `docs-site/.locale-parity-baseline`, and four deleted pages). It prints none of: `docs-site/vercel.json`, `docs-site/layouts/…`, `docs-site/static/…`, any `workflow-commands/moai-<goal|harness|plan|project|run|sync>.md`.
  2. The same command with the pathspec `.claude .moai/docs .moai/config scripts internal cmd` prints nothing (`.moai/config` covers `.moai/config/sections/design.yaml`, which REQ-SWF-007 and plan.md §D keep unchanged).
  3. The edit to the baseline file is exactly one deleted line (AC-SWF-009 Then 2).
- **RED-now cells:** none. Green today by construction (no edit exists); the criterion is a guard. Its failure is shown by the mutants in R-4.
- **Mutant probe (recipe R-4):** touching `docs-site/vercel.json`, any of the other six pages, or `docs-site/layouts/_default/baseof.html` makes Then 1 print an extra path; leaving the baseline file untouched makes Then 1 print thirteen paths (a short sweep is a failure); editing a comment line of `.moai/config/sections/design.yaml` makes Then 2 print that path. Cells on `74ad197380` against HEAD, verbatim: `git diff --name-only 74ad197380 HEAD -- .claude .moai/docs .moai/config scripts internal cmd` printed nothing (exit 0, green today); the control `git diff --name-only 74ad197380 HEAD -- .moai/config .moai/specs` printed the five SPEC files, so the form prints a path when a path in its pathspec changes.
- **Gate:** none (O-2 and O-3 are closed by the leader ruling; the larger set is stated here).
- **Done when:** Then 1 to 3 hold on the committed tree.

### AC-SWF-005 — no link to a removed page remains (REQ-SWF-004)

- **Given** the page is linked from the four `_index.md` files, registered in the four `_meta.yaml` files, and listed in `docs-site/data/menu/main.yaml` (C-5: ten lines).
- **When** the Q14 verdict (remove) is applied and M2 deletes the four pages, the four `_meta.yaml` entries, and the `main.yaml` entry,
- **Then**
  1. The C-5 command run on the closing commit (after M3) prints nothing and exits 1: no line under `docs-site/content` and none in `docs-site/data`. After M2 alone the whole-set C-5 still prints the five `_index.md` lines (M3 deletes them); the M2 check is C-5 narrowed to the `_meta.yaml`, page, and `docs-site/data` paths (plan.md §B M2 step 4), which prints nothing.
  2. Each `_meta.yaml` still parses and lists only keys that name an existing page, `index`, or `new_items` (the exempt non-page key; recipe R-5: `yq 'keys'` on each file, compared with the page file names of the same directory).
- **RED-now cells:** C-5. Red because 10 lines name the page today (4 `_index.md` rows, with ko holding 2, 4 `_meta.yaml` keys, and `main.yaml:208`).
- **Green path:** M3 flips the five `_index.md` lines; M2 flips the four `_meta.yaml` keys and the `main.yaml` line.
- **Mutant probe:** deleting the pages and leaving the `_meta.yaml` key fails Then 1 and Then 2; deleting the key and leaving the page fails AC-SWF-004 only if the page is not in the deleted set, so Then 2 compares the two sets directly; deleting the `main.yaml` name-map lines and leaving the `ref` line fails Then 1 (see AC-SWF-008).
- **Gate:** none (Q14 decided: remove).
- **Done when:** Then 1 and 2 hold.

### AC-SWF-006 — [SUPERSEDED by the Q14 verdict `d-20261019T021508Z-258e`] (REQ-SWF-005, REQ-SWF-006)

The rewrite and marker branches (a kept page with a slash-free title, or a marked citation) are not taken. This criterion is retired, not applicable, and not imported by the parent's M6; the number is kept so references do not move. C-6 stays in the ledger as a measurement of the four slash-form titles (they are deleted with the pages).

### AC-SWF-007 — the count statements equal the subcommand row count (REQ-SWF-008; leader ruling O-4)

- **Given** the heading "All subcommands" and the sentence below it state 16 in all four locales (C-10), the section holds 16 subcommand rows in each locale (C-12: `16 16 16 16`), and 17 commands are registered (C-8). The stated number equals the row count today; deleting the design row (Q13 verdict) is what makes it false.
- **When** M3 deletes the design row of the "All subcommands" table and corrects the heading and the sentence in the same pass, in all four locales,
- **Then**
  1. The C-12 command run on the closing commit prints one number four times, the same in every locale (`15 15 15 15`, unless a row is added to that section).
  2. The C-11 probe, with that number substituted for 15, prints `:2` for each of the four files (the heading and the sentence both state it). The same probe with 16 prints nothing and exits 1.
  3. The number is a row count, not the registered count: the C-8 command still prints seventeen paths, and a closing state that states 17 while C-12 prints 15 fails Then 1 and 2.
- **Counting method (decided here, with evidence).** Count the table rows, inside the "All subcommands" section of each `_index.md`, whose first cell starts a link (`| [`): this is exactly the C-12 command. It does not count the registered slash commands. Why: (1) the leader ruling O-4 says "actual row count"; (2) the heading introduces the two tables beneath it, so the number a reader can verify there is the row count; (3) the registered count, 17, includes `/moai gtd` and `/moai todo`, which have utility-commands pages (C-19) but no row on this page, so stating 17 would disagree with the rows beneath it. The alternative reading, that "routes N subcommands" means the registered commands (17), was weighed and not taken: the ruling's wording and the visible rows decide against it, and the word "All" stays inexact until O-6 is answered.
- **RED-now cells:** C-11. Red because the heading and the sentence state 16 today, while the closing row count is 15: the probe for 15 prints nothing. Controls: C-11's probe for 16 prints `:2` four times (the pattern matches where the work has not been done), and C-12 prints `16 16 16 16` (the count method reads the section, and today the two agree, which is why the statement and the row must change together). Red for the stated reason: only this work edits the heading, the sentence, and the design row.
- **Green path:** M3 flips C-11 and moves C-12 from 16 to 15.
- **Mutant probe:** fixing three locales and leaving one at 16 prints three lines instead of four at Then 2; correcting the heading and not the sentence prints `:1` for that locale; deleting the design row and leaving 16 prints `16` at Then 1 for the stated number mismatch (the probe for 15 prints nothing); correcting the number to 15 and leaving the design row makes C-12 print 16 and fails Then 1; stating 17 fails Then 1 and 2; deleting a different row than the design row keeps C-12 at 15 but fails AC-PRC-001.
- **Gate:** none (Q13 decided; the ruling O-4 closed). A row added to the section would change the number, and the criterion is then met by agreement, not by the literal 15.
- **Done when:** Then 1 to 3 hold on the committed tree.

### AC-SWF-008 — the sidebar data carries no design-family form (REQ-SWF-009; leader ruling O-2)

- **Given** `docs-site/data/menu/main.yaml` lines 204 to 207 carry `/moai design` in four locales and line 208 carries `ref: /workflow-commands/moai-design` (C-13, C-5); the previous/next chain of `baseof.html` is built from this file without a page-exists guard (C-17).
- **When** the structure curator deletes the entry in M2 (one pass, with the four `_meta.yaml` files),
- **Then**
  1. The C-13 command run on the closing commit prints nothing and exits 1.
  2. The R-8 command prints nothing, the R-9 command prints `6`, and the C-5 command prints nothing (no `ref` line is left).
- **RED-now cells:** C-13. Red because four lines carry the design family today; it flips only when this work deletes lines 203 to 208 of this file. Controls: R-8 prints the four-line name map and R-9 prints `7` on the base.
- **Green path:** M2 flips C-13, R-8, R-9, and C-5's `main.yaml` line.
- **Mutant probe:** deleting only the four name-map lines and leaving the `ref` makes C-13 print nothing (a pass) while the file still parses with an entry whose `name` is null and whose `ref` is the removed page, so Then 2 (C-5 prints the `ref` line; R-9 prints `7` instead of `6`) catches it; deleting the four lines and the `- name:` line but leaving the `ref` attaches a second `ref` key to the `/moai sync` entry and C-5 prints the line; retitling the entry instead of deleting it leaves R-8 printing a name map and R-9 `7`; deleting a different entry loses a sidebar item and R-9 prints `5`.
- **Gate:** none (Q14 decided: remove).
- **Done when:** Then 1 and 2 hold on the committed tree.

### AC-SWF-009 — the parity baseline line follows the page (REQ-SWF-010; leader ruling O-3)

- **Given** line 51 of `docs-site/.locale-parity-baseline` reads `workflow-commands/moai-design.md` (C-14). The CI step "Validate 4-locale parity" runs `scripts/docs-i18n-check.sh` (C-16) and does not read this file; its reader is the oss-docs harness verify recipe (C-15).
- **When** the content author deletes the line in the same change as the pages (M2),
- **Then**
  1. The C-14 command run on the closing commit prints nothing and exits 1.
  2. The R-10 command prints `0	1	docs-site/.locale-parity-baseline` (no line added, one deleted; the file's own header calls adding a line "admitting new debt").
- **RED-now cells:** C-14. Red because the line is present today; it flips only when this work deletes the line. Control: the R-10 form prints a numstat triple on a changed file (C-14's note).
- **Green path:** M2, step 1 (content author).
- **Mutant probe:** replacing the line with a new path prints `1	1` at Then 2; deleting a neighbouring line leaves C-14 printing the design line (Then 1 fails); leaving the file untouched leaves C-14 printing it.
- **Gate:** none (Q14 decided: remove).
- **Done when:** Then 1 and 2 hold.

## C. Evidence ledger (verbatim)

Pin for every `git` cell: tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea`. Measured 2026-10-19 on the working tree (clean).

### C-1 — the design family in the twelve files (RED-now; AC-PRC-001)

- Command: `git grep -c -E "/moai[: ]design([^[:alnum:]-]|$)" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 12 lines; derived total 27):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:1
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/ja/workflow-commands/_index.md:1
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:4
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/zh/workflow-commands/_index.md:1
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:4
```

- Reading: re-measures the premise check (27 lines in 12 files). Red because the design family is cited; the work edits exactly these files.

### C-2 — any `/moai` verb outside the registered list (RED-now; AC-SWF-002)

- Command: `git grep -c -P "/moai[: ](?!(clean|codemaps|e2e|feedback|fix|gate|goal|gtd|harness|loop|mx|plan|project|review|run|sync|todo)(?![a-z0-9-]))[a-z][a-z0-9-]*" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 12 lines, identical to C-1):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:1
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/ja/workflow-commands/_index.md:1
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:4
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/zh/workflow-commands/_index.md:1
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:4
```

- Reading: the list in the pattern is the seventeen files of C-8. The output equals C-1 line for line, so `design` is the only unregistered verb in the twelve files. The cell is its own positive control: it matches where the work has not been done.

### C-3 — the swept set (control; AC-PRC-001, AC-SWF-002)

- Command: `git ls-files -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: working tree equal to `74ad197380` for `docs-site` (section A).
- Stdout (verbatim; 12 lines):

```
docs-site/content/en/workflow-commands/_index.md
docs-site/content/en/workflow-commands/_meta.yaml
docs-site/content/en/workflow-commands/moai-design.md
docs-site/content/ja/workflow-commands/_index.md
docs-site/content/ja/workflow-commands/_meta.yaml
docs-site/content/ja/workflow-commands/moai-design.md
docs-site/content/ko/workflow-commands/_index.md
docs-site/content/ko/workflow-commands/_meta.yaml
docs-site/content/ko/workflow-commands/moai-design.md
docs-site/content/zh/workflow-commands/_index.md
docs-site/content/zh/workflow-commands/_meta.yaml
docs-site/content/zh/workflow-commands/moai-design.md
```

### C-4 — `/moai` table rows per locale in `_index.md` (RED-now; AC-SWF-003)

- Command: `git grep -c -E "^\| \[.?/moai " 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 4 lines):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:21
74ad197380:docs-site/content/ja/workflow-commands/_index.md:21
74ad197380:docs-site/content/ko/workflow-commands/_index.md:23
74ad197380:docs-site/content/zh/workflow-commands/_index.md:21
```

- Reading: the pattern is a table row whose first cell links a `/moai` command (the optional character after `[` is the opening backtick). Korean holds two design rows (lines 42 and 66) and a `/moai goal` summary row (line 41) that the other locales lack. Closing figures: 20, 20, 21, 20 (en, ja, ko, zh). In en, ja, and zh the 21 rows are the 5 rows of the command summary table plus the 16 rows of the "All subcommands" tables (C-12); in ko, 7 summary rows plus 16.

### C-5 — references to the page (RED-now; AC-SWF-005)

- Command: `git grep -n -E "moai-design" 74ad197380 -- docs-site/content docs-site/data README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 10 lines):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:62:| [`/moai design`](./moai-design) | Design-phase collaboration (manager-design D1-D5) |
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/content/ja/workflow-commands/_index.md:62:| [`/moai design`](./moai-design) | デザイン段階のコラボレーション (manager-design D1-D5) |
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/content/ko/workflow-commands/_index.md:42:| [`/moai design`](./moai-design) | 조건부 | manager-design | - | UI SPEC을 위한 디자인 단계 협업 |
74ad197380:docs-site/content/ko/workflow-commands/_index.md:66:| [`/moai design`](./moai-design) | 디자인 단계 협업 (manager-design D1-D5) |
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/content/zh/workflow-commands/_index.md:62:| [`/moai design`](./moai-design) | 设计阶段协作(manager-design D1-D5) |
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/data/menu/main.yaml:208:        ref: /workflow-commands/moai-design
```

- Reading: nine lines are under `docs-site/content`, one is in `docs-site/data/menu/main.yaml`, which the leader ruling brought into the change set (REQ-SWF-009). The parity baseline line is outside this pathspec (C-14).

### C-6 — page titles are slash forms (supporting; AC-SWF-006 is superseded, the pages are deleted)

- Command: `git grep -n -E "^title: /moai" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 4 lines):

```
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:2:title: /moai design
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:2:title: /moai design
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:2:title: /moai design
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:2:title: /moai design
```

### C-7 — no routing workflow names the design workflow as an entry (supporting; Q13 determination)

- Command: `git grep -n -i -E "workflows/design|design phase|manager-design|plan.{1,3}design" 74ad197380 -- .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/sync.md .claude/skills/moai/SKILL.md`
- Exit code: 1 by git grep semantics (no match; not displayed by the tool, gap G-1). Stdout: no output.
- Reading: none of the four routing files names the design workflow. The design route is described in `.claude/rules/moai/workflow/spec-workflow.md` § Conditional Design Route and `.claude/skills/moai/workflows/design.md`, and `.claude/rules/moai/design/constitution.md` records `/moai design` as no longer a routed subcommand. With C-8 and C-9 this is the basis of "no current equivalent exists". The positive control is C-9 (same form, matching file).

### C-8 — registered slash commands (supporting)

- Command: `git ls-tree -r --name-only 74ad197380 -- .claude/commands/moai`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 17 lines; no `design`):

```
.claude/commands/moai/clean.md
.claude/commands/moai/codemaps.md
.claude/commands/moai/e2e.md
.claude/commands/moai/feedback.md
.claude/commands/moai/fix.md
.claude/commands/moai/gate.md
.claude/commands/moai/goal.md
.claude/commands/moai/gtd.md
.claude/commands/moai/harness.md
.claude/commands/moai/loop.md
.claude/commands/moai/mx.md
.claude/commands/moai/plan.md
.claude/commands/moai/project.md
.claude/commands/moai/review.md
.claude/commands/moai/run.md
.claude/commands/moai/sync.md
.claude/commands/moai/todo.md
```

### C-9 — the design workflow is not user-invocable (supporting)

- Command: `git grep -n -E "user-invocable" 74ad197380 -- .claude/skills/moai/workflows/design.md`
- Exit code: 0. Stdout (verbatim): `74ad197380:.claude/skills/moai/workflows/design.md:11:user-invocable: false`

### C-10 — the count statements on `_index.md` (supporting; AC-SWF-007)

- Command: `git grep -n -E "^## (All subcommands|全サブコマンド|전체 서브커맨드|全部子命令)|routes [0-9]+ subcommands|[0-9]+ 個のサブコマンド|[0-9]+개 서브커맨드|[0-9]+ 个子命令" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 8 lines):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:50:## All subcommands (16)
74ad197380:docs-site/content/en/workflow-commands/_index.md:52:The `/moai` orchestrator routes 16 subcommands. This (workflow) section covers the SPEC 3-Phase lifecycle commands, and the [utility commands](/en/utility-commands/) section covers automation, fix-loop, code-management, and feedback commands.
74ad197380:docs-site/content/ja/workflow-commands/_index.md:50:## 全サブコマンド (16 個)
74ad197380:docs-site/content/ja/workflow-commands/_index.md:52:`/moai` オーケストレーターは 16 個のサブコマンドをルーティングします。このセクション (ワークフロー) は SPEC 3-Phase ライフサイクルのコマンドを、[ユーティリティコマンド](/ja/utility-commands/) セクションは自動化・修正ループ・コード管理・フィードバックのコマンドを扱います。
74ad197380:docs-site/content/ko/workflow-commands/_index.md:53:## 전체 서브커맨드 (16개)
74ad197380:docs-site/content/ko/workflow-commands/_index.md:55:`/moai` 오케스트레이터는 16개 서브커맨드를 라우팅합니다. 이 섹션(워크플로우)에서는 SPEC 3-Phase 라이프사이클 명령어를 다루고, 자동화·수정 루프·코드 관리·피드백 명령어는 [유틸리티 명령어](/ko/utility-commands/) 섹션에 있습니다.
74ad197380:docs-site/content/zh/workflow-commands/_index.md:50:## 全部子命令(16 个)
74ad197380:docs-site/content/zh/workflow-commands/_index.md:52:`/moai` 编排器路由 16 个子命令。本节(工作流)讲解 SPEC 3-Phase 生命周期命令,[实用命令](/zh/utility-commands/)一节讲解自动化·修复循环·代码管理·反馈命令。
```

- Reading: this narrows the 17-line output of revision 0.1.0 to the two statements per locale. The stated count is 16 and the registered count is 17 (C-8).

### C-11 — the corrected count statements (RED-now; AC-SWF-007)

- Command (probe for 15): `git grep -c -E "\(15\)|\(15 個\)|\(15개\)|\(15 个\)|routes 15 subcommands|15 個のサブコマンド|15개 서브커맨드|15 个子命令" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'`
- Exit code: 1 by git grep semantics (no match; not displayed by the tool, gap G-1). Tree: `74ad197380`. Stdout: no output.
- Red because the heading and the sentence state 16 today and the closing row count is 15 (C-12 after M3).
- Control (same form for 16), exit code 0, stdout verbatim:

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:2
74ad197380:docs-site/content/ja/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/zh/workflow-commands/_index.md:2
```

- Command (control for 16): `git grep -c -E "\(16\)|\(16 個\)|\(16개\)|\(16 个\)|routes 16 subcommands|16 個のサブコマンド|16개 서브커맨드|16 个子命令" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'`
- Closing: the probe with the number N of C-12 in place of 15 prints the four lines above with `:2`; the probes for any other number print nothing.

### C-12 — the subcommand row count per locale (control; AC-SWF-007)

- Command: `awk '/^## /{s=($0 ~ /^## (All subcommands|全サブコマンド|전체 서브커맨드|全部子命令)/)} s && /^\| \[/{n[FILENAME]++} END{print n[ARGV[1]]+0, n[ARGV[2]]+0, n[ARGV[3]]+0, n[ARGV[4]]+0}' docs-site/content/en/workflow-commands/_index.md docs-site/content/ja/workflow-commands/_index.md docs-site/content/ko/workflow-commands/_index.md docs-site/content/zh/workflow-commands/_index.md`
- Exit code: 0. Tree: working tree equal to `74ad197380` for `docs-site` (section A). The semicolons are awk statement separators inside one quoted argument; the command is one invocation.
- Stdout (verbatim; order en, ja, ko, zh): `16 16 16 16`
- Reading: rows between the "All subcommands" heading and the next `## ` heading whose first cell starts a link. The en tables hold 6 workflow rows and 10 utility rows. Closing: `15 15 15 15`.

### C-13 — the design family in the sidebar data (RED-now; AC-SWF-008)

- Command: `git grep -n -E "/moai[: ]design([^[:alnum:]-]|$)" 74ad197380 -- docs-site/data/menu/main.yaml`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 4 lines):

```
74ad197380:docs-site/data/menu/main.yaml:204:          ko: /moai design
74ad197380:docs-site/data/menu/main.yaml:205:          en: /moai design
74ad197380:docs-site/data/menu/main.yaml:206:          ja: /moai design
74ad197380:docs-site/data/menu/main.yaml:207:          zh: /moai design
```

- Reading: the entry is lines 203 (`- name:`) to 208 (`ref: /workflow-commands/moai-design`, C-5). Red because the family is present; it flips only when this work edits those lines.

### C-14 — the page line in the parity baseline (RED-now; AC-SWF-009)

- Command: `git grep -n -E "moai-design" 74ad197380 -- docs-site/.locale-parity-baseline`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 1 line): `74ad197380:docs-site/.locale-parity-baseline:51:workflow-commands/moai-design.md`
- Control for the numstat form of R-10 (a changed file prints a triple): `git diff --numstat 74ad197380 HEAD -- .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001/progress.md` printed `32	0	.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001/progress.md` (exit 0). The same form on the baseline file printed nothing (the file is unmodified).

### C-15 — who reads the parity baseline (supporting; AC-SWF-009)

- Command: `git grep -l -F "locale-parity-baseline" 74ad197380 -- . ':(exclude).moai/specs' ':(exclude).moai/reports'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 2 lines):

```
74ad197380:.claude/skills/hns-oss-docs-verify/SKILL.md
74ad197380:CHANGELOG.md
```

- Command (the reader's lines): `git grep -n -F "locale-parity-baseline" 74ad197380 -- .claude/skills`
- Exit code: 0. Stdout (verbatim; 2 lines):

```
74ad197380:.claude/skills/hns-oss-docs-verify/SKILL.md:76:The gate is a **ratchet**, not an absolute check. `docs-site/.locale-parity-baseline`
74ad197380:.claude/skills/hns-oss-docs-verify/SKILL.md:98:grep -v '^#' ../.locale-parity-baseline | grep -v '^[[:space:]]*$' | sort > /tmp/parity-base.txt
```

- Reading: no workflow, script, Makefile, Go source, or agent definition under the searched tree names the file; the one reader is the verify recipe's section 4 (a ratchet: a divergent page not in the baseline fails; a baseline page that has converged is reported "prune", informational). `CHANGELOG.md` only records earlier prunes. A reader outside this repository tree was not searched (gap G-7).

### C-16 — what the CI step runs (supporting; AC-SWF-009)

- Command: `git grep -n -F "Validate 4-locale parity" 74ad197380 -- .github/workflows`
- Exit code: 0. Stdout (verbatim): `74ad197380:.github/workflows/docs-i18n-check.yml:53:    name: Validate 4-locale parity`
- Command: `git grep -n -F "scripts/docs-i18n-check.sh" 74ad197380 -- .github/workflows/docs-i18n-check.yml`
- Exit code: 0. Stdout (verbatim; 8 lines):

```
74ad197380:.github/workflows/docs-i18n-check.yml:13:#   Phase 2 (flip trigger: scripts/docs-i18n-check.sh reports 0 ERROR drifts
74ad197380:.github/workflows/docs-i18n-check.yml:25:      - 'scripts/docs-i18n-check.sh'
74ad197380:.github/workflows/docs-i18n-check.yml:35:      - 'scripts/docs-i18n-check.sh'
74ad197380:.github/workflows/docs-i18n-check.yml:63:        run: chmod +x scripts/docs-i18n-check.sh
74ad197380:.github/workflows/docs-i18n-check.yml:89:            DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh 2>&1 | tee check-output.txt
74ad197380:.github/workflows/docs-i18n-check.yml:91:            DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh 2>&1 | tee check-output.txt
74ad197380:.github/workflows/docs-i18n-check.yml:150:              'DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh',
74ad197380:.github/workflows/docs-i18n-check.yml:153:              'scripts/docs-i18n-check.sh',
```

- Command: `git grep -n -E "^# Check [0-9]" 74ad197380 -- scripts/docs-i18n-check.sh`
- Exit code: 0. Stdout (verbatim; 4 lines):

```
74ad197380:scripts/docs-i18n-check.sh:91:# Check 1: File path parity
74ad197380:scripts/docs-i18n-check.sh:130:# Check 2: YAML frontmatter `title` presence
74ad197380:scripts/docs-i18n-check.sh:175:# Check 3: H1 heading existence
74ad197380:scripts/docs-i18n-check.sh:220:# Check 4: Glossary term preservation
```

- Reading: the step checks file-path parity (every ko page needs an en, ja, and zh counterpart), a non-empty `title`, an H1, and glossary terms. It does not read the baseline. Under Q14 = remove, deleting all four pages together keeps check 1 clean; a rewrite must keep a non-empty `title` (check 2). The script was read, not run (gap G-2).

### C-17 — what reads `main.yaml` (supporting; AC-SWF-008)

- Command: `git grep -n -E "hugo.Data.menu.main|GetPage" 74ad197380 -- docs-site/layouts/_default/baseof.html docs-site/layouts/partials/menu.html`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 6 lines):

```
74ad197380:docs-site/layouts/_default/baseof.html:58:          {{- range hugo.Data.menu.main.main -}}
74ad197380:docs-site/layouts/partials/menu.html:17:  {{- range hugo.Data.menu.main.main -}}
74ad197380:docs-site/layouts/partials/menu.html:28:  {{- range hugo.Data.menu.main.main }}
74ad197380:docs-site/layouts/partials/menu.html:31:    {{- if and (or .external (site.GetPage .ref)) (or (not .ko_only) (eq $.Site.Language.Lang "ko")) -}}
74ad197380:docs-site/layouts/partials/menu.html:77:          {{- if or .external (site.GetPage .ref) -}}
74ad197380:docs-site/layouts/partials/menu.html:89:            {{- $subPage := site.GetPage .ref -}}
```

- Reading: `menu.html` guards the sidebar with `site.GetPage`; `baseof.html` flattens the same sub entries into the previous/next chain (lines 58 to 86, read in full) and carries no `GetPage` line. `docs-site/layouts/index.html` and `docs-site/layouts/_default/list.html` also read the file (`git grep -l -F "hugo.Data.menu.main" 74ad197380 -- docs-site/layouts` listed four layout files). Read from source, not rendered (gap G-4).

### C-18 — bounded sweep for other design-family citations (supporting; premise check for O-2)

- Command: `git grep -n -E "/moai[: ]design([^[:alnum:]-]|$)" 74ad197380 -- docs-site ':(exclude)docs-site/content' .moai/config .github scripts Makefile ':(glob)*.json' ':(glob)*.yaml' ':(glob)*.yml' ':(glob)*.toml'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 6 lines):

```
74ad197380:.moai/config/sections/design.yaml:2:# Controls /moai design workflow, GAN loop, Claude Design integration, and brand context.
74ad197380:.moai/config/sections/design.yaml:36:  # Claude Design integration (path A of /moai design)
74ad197380:docs-site/data/menu/main.yaml:204:          ko: /moai design
74ad197380:docs-site/data/menu/main.yaml:205:          en: /moai design
74ad197380:docs-site/data/menu/main.yaml:206:          ja: /moai design
74ad197380:docs-site/data/menu/main.yaml:207:          zh: /moai design
```

- Command: `git grep -n -E "moai-design([^.]|$)" 74ad197380 -- docs-site ':(exclude)docs-site/content' .moai/config .github scripts Makefile ':(glob)*.json' ':(glob)*.yaml' ':(glob)*.yml' ':(glob)*.toml'`
- Exit code: 0. Stdout (verbatim; 4 lines):

```
74ad197380:docs-site/data/menu/main.yaml:208:        ref: /workflow-commands/moai-design
74ad197380:docs-site/static/moai-components.css:10: * Loading order (head/custom.html): moai-brand → moai-design → moai-docs-tokens
74ad197380:docs-site/static/moai-docs-layout.css:674: * here beat moai-brand / moai-design / moai-docs-theme. Where those files use
74ad197380:docs-site/static/moai-docs-layout.css:1274: * moai-brand / moai-design / moai-docs-theme.
```

- Reading: outside `docs-site/content` and the baseline, the page's name map and `ref` are in `main.yaml` only. The other hits are not the page: `moai-design` in the CSS comments and in `docs-site/layouts/partials/head/custom.html` (`moai-design.css`, found by `git grep -n -E "moai-design" … docs-site/layouts`) names a stylesheet. The two `.moai/config/sections/design.yaml` comment lines are open item O-5; `docs-site/vercel.json` carries only `/design` redirects (REDIR, Q15), found by `git grep -n -E "design" 74ad197380 -- docs-site/vercel.json docs-site/hugo.toml docs-site/layouts/partials/menu.html`. The sweep covers `docs-site/` outside `content/`, `.moai/config`, `.github`, `scripts`, `Makefile`, and root-level `*.json|yaml|yml|toml` (none matched); it is a pattern sweep, not a parse (gap G-9). The `.claude/` rule and skill files that mention the workflow are out of scope.

### C-19 — the two utility pages exist (supporting; AC-SWF-007 counting method, O-6)

- Command: `git ls-files -- ':(glob)docs-site/content/en/utility-commands/moai-todo.md' ':(glob)docs-site/content/en/utility-commands/moai-gtd.md'`
- Exit code: 0. Tree: working tree equal to `74ad197380` for `docs-site`.
- Stdout (verbatim; 2 lines):

```
docs-site/content/en/utility-commands/moai-gtd.md
docs-site/content/en/utility-commands/moai-todo.md
```

- Reading: the registered commands `/moai gtd` and `/moai todo` have pages, and neither has a row in the "All subcommands" tables (C-10 shows the en rows: plan, run, sync, project, design, harness, fix, loop, mx, feedback, review, clean, codemaps, gate, e2e, goal).

## B.0 Local verification recipes

- **R-1, R-6, R-7** are retired with the rewrite and marker branches (Q14 verdict); the numbers are kept so references do not move.
- **R-3 (same change).** `git show --stat --format= <closing-sha>` lists the four `_index.md` paths in one commit.
- **R-4 (preservation guard).** `git diff --name-only 74ad197380 <closing-sha> -- docs-site README.md README.ko.md README.ja.md README.zh.md` expected to print only the paths of spec.md § Scope; the companion pathspec `.claude .moai/docs .moai/config scripts internal cmd` expected to print nothing. Mutants: append a line to `docs-site/vercel.json`, to `docs-site/layouts/_default/baseof.html`, or to `docs-site/content/en/workflow-commands/moai-run.md`, in a scratch copy of the commit; the command prints the extra path.
- **R-5 (`_meta.yaml` parse and key set).** For each locale: `yq 'keys' docs-site/content/<L>/workflow-commands/_meta.yaml`, compared with `git ls-files docs-site/content/<L>/workflow-commands`; every key except `index` and `new_items` must name an `<key>.md` that exists. (Base: `yq 'keys'` on the ko file prints the eight page keys, the two comment lines, and `new_items`.)
- **R-8 (sidebar name map).** `yq '.main[] | select(.ref == "/workflow-commands") | .sub[] | select(.ref == "/workflow-commands/moai-design") | .name' docs-site/data/menu/main.yaml`. Base (exit 0), verbatim: `ko: /moai design`, `en: /moai design`, `ja: /moai design`, `zh: /moai design`. Closing: no output.
- **R-9 (sidebar entry count).** `yq '.main[] | select(.ref == "/workflow-commands") | .sub | length' docs-site/data/menu/main.yaml`. Base (exit 0): `7`. Closing: `6`.
- **R-10 (baseline edit shape).** `git diff --numstat 74ad197380 <closing-sha> -- docs-site/.locale-parity-baseline`. Closing: `0	1	docs-site/.locale-parity-baseline`. (Base against `HEAD`: no output; form control in C-14.)

## D. Definition of done

This child closes when AC-PRC-001, AC-SWF-002, AC-SWF-003, and AC-SWF-005, AC-SWF-007, AC-SWF-008, and AC-SWF-009 hold on its committed tree, and AC-SWF-004 holds as a guard (AC-SWF-006 is superseded). The parent imports the child's criteria by reference at its M6 (parent plan §B.3).

| AC | REQ | Class | Imported at the parent's M6 as |
|----|-----|-------|--------------------------------|
| AC-PRC-001 | REQ-PRC-001 | release-blocking | this child's AC-PRC-001 at its commit SHA |
| AC-SWF-002 | REQ-PRC-001, REQ-SWF-002 | release-blocking | this child's AC-SWF-002 |
| AC-SWF-003 | REQ-SWF-002, REQ-SWF-003 | release-blocking | this child's AC-SWF-003 |
| AC-SWF-004 | REQ-SWF-007 | guard | re-run as an observation |
| AC-SWF-005 | REQ-SWF-004 | release-blocking | this child's AC-SWF-005 |
| AC-SWF-006 | REQ-SWF-005, REQ-SWF-006 | superseded | "not applicable" (superseded by the Q14 verdict) |
| AC-SWF-007 | REQ-SWF-008 | release-blocking | this child's AC-SWF-007 |
| AC-SWF-008 | REQ-SWF-009 | release-blocking | this child's AC-SWF-008 |
| AC-SWF-009 | REQ-SWF-010 | release-blocking | this child's AC-SWF-009 |

Common gates: `moai spec lint --strict` on this directory reports no finding (recorded in `progress.md`); `git status --short` prints nothing after the verification runs; the deciding command and output lines of every claim are carried into the lane's verdict file.

## E. Edge cases

- The ko `_index.md` holds a design row in the summary table (line 42) that the other locales do not have; deleting only the subcommand-table row leaves a stale ko row (AC-SWF-003 Then 1). That summary row is outside the "All subcommands" section, so it does not enter the count of C-12.
- The ja and zh pages are 68 non-blank lines against 82 in en and ko (the usage example block is absent from ja and zh); with the pages deleted the asymmetry disappears with them.
- The pages, the `_meta.yaml` keys, and the `main.yaml` entry go in one pass (M2), so the sidebar and the previous/next chain never point at a missing page; the `_index.md` rows (M3) may land in the same or a following commit, and a row left behind would link `./moai-design` to a missing page, which C-5 and C-1 catch.
- The sidebar template `menu.html` hides an entry whose `ref` page does not exist (`site.GetPage` guard at lines 31 and 77), but the previous/next chain of `baseof.html` has no such guard (C-17): a removed page with its `main.yaml` entry left behind would leave `/moai sync` and `/moai harness` linking to a missing page. This corrects revision 0.1.0, which read only `menu.html`. Both were read from source, not rendered (G-4).
- A `main.yaml` edit that removes the four name-map lines and leaves `ref:` produces a YAML entry with a null `name` and the removed page as `ref` (the keys stay siblings), which parses; only the criteria's second and third clauses catch it (AC-SWF-008 mutants).
- Deleting the baseline line keeps the harness recipe clean; leaving it would make the recipe report the page as "converged, prune" (informational) (C-15).

## F. Quality gates (TRUST 5 mapping)

- Tested: each criterion has a RED-now cell or an explicit reason it has none (AC-SWF-004 guard; AC-SWF-006 is superseded), a stated green output, and a mutant.
- Readable: the criteria name the row, file, and locale they count.
- Unified: edits follow the ko→en→ja/zh chain of `.moai/docs/docs-site-i18n-rules.md` §17.3 and the four-locale same-change obligation (Q17 verdict).
- Secured: documentation only; no secret, URL, or executable content changes. The URL blacklist and Mermaid TD-only rules are untouched because no URL or diagram is added.
- Trackable: commits carry the card id `t1727` and the SPEC id; the ledger is carried into the verdict file; the rulings are cited by board record id.

## G. Gaps (not observed)

- G-1 — The shell tool does not display the exit status of a no-output command. C-7, C-11 (the probe for 15), and every green-path "prints nothing" cell take exit 1 from git grep semantics; they were not captured with a wrapper (the worktree guard refuses compound forms).
- G-2 — No Hugo build, link check, i18n parity script, or verify recipe was run on a tree with the edits; the change does not exist yet. `scripts/docs-i18n-check.sh` was read, not run. The GATE jobs will measure it later.
- G-3 — The behavior of the structure curator, the content author, and the locale translator on these edits was not run; plan.md §A.1 cites their definitions only. Neither the content-author nor the translator definition names page deletion; the reading in A.1 rule 6 is an inference from the translator's parity rule.
- G-4 — The `menu.html` and `baseof.html` behavior for a missing page (hidden sidebar entry; unguarded previous/next link) was read from the template source, not rendered.
- G-5, G-6 — retired with the rewrite and marker branches (Q14 verdict); no rewritten text and no marker token exist.
- G-7 — The reader of the parity baseline was identified by a repository search (C-15) of the tree at `74ad197380`, excluding `.moai/specs` and `.moai/reports`; a reader outside this tree (a personal script, another repository) cannot be excluded.
- G-8 — Sibling children's own file lists are not in this tree; non-overlap was read from the parent table only, and the parent table lists the workflow-commands content directories, not `main.yaml` or the baseline file.
- G-9 — The grep patterns do not cover every possible spelling of the design family (for example extra spaces or a line break between `/moai` and `design`); no such spelling was found by reading the 27 matching lines and the page headings, but the sweep is a pattern, not a parse. C-18 is the same kind of sweep over a bounded path set.
- G-10 — The text of the leader ruling `d-20261019T020706Z-1c58` and the Q14 operator verdict `d-20261019T021508Z-258e` are taken from the messages of the dispatching coordinator, who states it verified both on the board; neither board record was opened in this pass.
- G-11 — C-12 and R-7 to R-9 read working-tree files and run `awk` and `yq`; the equality of the working tree with `74ad197380` rests on the `git diff --stat` and `git status --short` observations of section A.
- G-12 — The pass probe of C-11 assumes ASCII parentheses in the zh heading (`全部子命令(16 个)`); the probe for 16 matched all four headings and sentences, which confirms the assumption on this tree.

## H. Residual risk

- A pattern that names the registered verbs (C-2) is stale the day a command is added or removed; the cell is pinned to its tree and re-run, not trusted later.
- The Q13 determination "no current equivalent exists" is measured on four routing files and the command list; a non-slash entry to the design route exists in rules and a skill, so an operator who reads "equivalent" more widely could ask for a non-slash pointer; the Q14 verdict (remove) leaves nothing to point at.
- The public URLs of the four pages stop serving and no redirect is in this change (REDIR's, Q15).
- Korean remains one `/moai` row above the other locales after the work (the `/moai goal` summary row); a reader comparing counts may misread this as drift.
- "All subcommands (15)" is a true row count and an inexact "All": two registered commands have no row (O-6). A reader who counts registered commands finds 17.
- The lane's standing spawn authority does not name the harness specialists; the plan records the leader ruling and grants nothing (plan.md §A.1 rule 5).
- The ruling's statement that CI reads the parity state is not what the tree shows (C-15, C-16); the baseline edit is applied as ruled, and a stale line would have been informational only.
