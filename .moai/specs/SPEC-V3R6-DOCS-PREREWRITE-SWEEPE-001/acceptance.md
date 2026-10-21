# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 (revision 0.3.0)

## A. Pin and conventions

- **Pin.** Tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (short `74ad197380a0`), branch WT-getting-started-faq. Measured 2026-10-19. This document-level pin binds every cell that names no other tree. Every ledger cell that reads a committed object names the commit in its command. The revision-2 measurements were taken at HEAD `e85847fd331b14320d4825338d2923830f4d1bb0` (the revision-1 commit); cell S-00 shows the docs-site, `.claude`, `internal`, `cmd`, and `scripts` trees are identical between the pin and that HEAD, so the pin and the HEAD agree on every page file and every Go source.
- **Cell form** (verification-completeness §2.1). One read-only invocation, no `;`, `&&`, `|`, subshell, or redirection inside the cell. Stdout is verbatim and the exit code is its own field. The exit code was captured by appending `; echo "exit=$?"` outside the cell in the plain compound form; that suffix is part of the measurement, not of the cell. A zero-match `git grep` exits 1 and prints nothing; an empty stdout with exit 1 is a complete observation. Every path is written out in the command; no placeholder is used. A `git grep -n` row carries the line number between the path and the matched text, and that number is the **position anchor** of the cell.
- **Run-time tree argument.** `git grep <commit>` reads committed objects only. Every flip check in the run phase therefore runs after the milestone's commit, with `HEAD` in place of `74ad197380a0`; an uncommitted edit is invisible to it.
- **Scope fence base.** The fence cells (S-13a, S-13b, S-16, S-17) compare `74ad197380a0` with `HEAD`. The run phase confirms S-00 at pre-flight; a non-empty S-00 means the docs-site tree moved and is a blocker report.
- **Tool provenance** (verification-claim-integrity §2.2). The verb-surface cells S-06d-1 to S-06d-5 use a binary built with `go build -o /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/84345352-92b4-4559-a09f-c8ce19d8e034/scratchpad/moai-t1725 ./cmd/moai` run in this worktree at HEAD e85847fd33 (exit 0). Any reader reproduces the build with the same `go build` command and substitutes the output path in the cell; the path above is the one measured. `go version -m` on the binary reads `vcs.revision=478285ab4c1fc893b020bb464d14cb69c8422523`, `vcs.modified=true`, which does not equal the tree commit; the stamp is unreliable in this linked worktree (gap G-2). `git log -1 --format=%H -- cmd internal` prints `74ad197380a0711cdb9816d4e37c94d29af730ea`, and S-00 shows no Go source changed since, so the build procedure (sources read from this worktree) is the attribution, not the stamp. The git cells use the installed `git`, whose output does not depend on the moai build.
- **Fixture reads.** The controls (§ D, C-1 to C-9, and C-10 in § H) run the same cell text against throwaway fixtures with the pin or `HEAD` replaced by a fixture branch name, run from the fixture repository directory. A fixture proves what a cell accepts and rejects; it does not prove the run-phase edit.
- **Class.** Release-blocking: AC-PRC-001a, b, d, f, h, i, j, k — each has a RED-now cell with the command, verbatim stdout, exit code, and pin. Regression-guard (green at the pin, must stay green): AC-PRC-001c, e, g.

## B. Traceability

| AC | REQ | Cell(s) | RED-now? | Class |
|----|-----|---------|----------|-------|
| AC-PRC-001a | REQ-PRC-001, REQ-SWEEPE-002 | S-01 | red (10 rows, 18 matching lines, exit 0) | release-blocking |
| AC-PRC-001b | REQ-PRC-001, REQ-SWEEPE-002 | S-02 | red (10 rows, exit 0) | release-blocking |
| AC-PRC-001c | REQ-PRC-001 | S-05a, S-05b, S-05c | green (empty, exit 1) | regression-guard |
| AC-PRC-001d | REQ-SWEEPE-002, REQ-SWEEPE-003 | S-11, S-18-en, S-18-ja, S-18-ko, S-18-zh | red (all five empty, exit 1) | release-blocking |
| AC-PRC-001e | REQ-SWEEPE-005 | S-10 | green (exit 0) | regression-guard |
| AC-PRC-001f | REQ-SWEEPE-004 | S-13a, S-13b, S-16-en, S-16-ja, S-16-ko, S-16-zh, S-17 | red (S-13a, S-13b, S-16 empty, exit 0; S-17 prints the twelve unedited blob ids, exit 0) | release-blocking |
| AC-PRC-001g | REQ-SWEEPE-001 | S-03 | green (empty, exit 1) | regression-guard |
| AC-PRC-001h | REQ-PRC-001 | S-06a, S-06b, S-06c | red (S-06a and S-06b print one row each, exit 0) | release-blocking |
| AC-PRC-001i | REQ-SWEEPE-006, REQ-SWEEPE-003 | S-12a-en, S-12a-ja, S-12a-ko, S-12a-zh, S-12b | red (S-12a-xx each empty, exit 1; S-12b four rows, exit 0) | release-blocking |
| AC-PRC-001j | REQ-SWEEPE-003 | S-14-en, S-14-ja, S-14-zh | red (each empty, exit 1) | release-blocking |
| AC-PRC-001k | REQ-SWEEPE-002 | S-15 | red (empty, exit 1; one row expected) | release-blocking |

Eleven criteria against the Tier M ceiling of 16; seven requirements against the same ceiling. Mapping note: `REQ-SWEEPE-001` to `006` refine the umbrella's REQ-PRC-001 page part; the umbrella imports this child's criteria by SPEC ID, criterion ID, and commit SHA.

## C. Acceptance criteria (Given-When-Then)

### AC-PRC-001a — no closing-pattern line remains in the page files

Covers REQ-PRC-001 and REQ-SWEEPE-002 (a flagged citation is replaced or deleted).

- Given the twelve page files at the run-phase HEAD,
- When cell S-01's command is run with `HEAD` in place of the pin,
- Then it prints nothing and exits 1. At the planner defaults no marker is added, so no flagged line is exempt. (QE-2's alternate has its own criteria, § H.)
- RED now: S-01 prints 10 file rows (18 matching lines) and exits 0. Why red: the 18 lines are the flagged lines; this work edits exactly those lines, so editing flips it.
- Mutant probe: replacing `/moai fb` with a new unresolved word such as `/moai fbk` passes this AC; AC-PRC-001b catches it (its pattern flags any word outside the registered set).

### AC-PRC-001b — no unresolved slash word remains

Covers REQ-PRC-001 and REQ-SWEEPE-002.

- Given the twelve page files at the run-phase HEAD,
- When cell S-02's command is run with `HEAD` in place of the pin,
- Then it prints nothing and exits 1: every `/moai[: ]<word>` in the page files names a registered or router word (`feedback`, `loop`, `run`, `mx`, `fix`, `plan`, `project`) or is the link-text separator `/moai -`.
- RED now: S-02 prints ten rows (18 lines) and exits 0. Why red: the same 18 lines name `init`, `fb`, `bug`, `issue`, `relationship`, or `brain`.

### AC-PRC-001c — the CLI families stay at zero

- Given the twelve page files,
- When the umbrella's E-01, E-02, and E-03 family commands are run restricted to them (cells S-05a, S-05b, S-05c),
- Then each prints nothing and exits 1 before and after the edits.
- Green now (precondition). Guards against a replacement that introduces an absent CLI verb of those families; the general verb check is AC-PRC-001h.

### AC-PRC-001d — feedback pages: replacement present at its position in all four locales (Q13 and Q17 rulings)

Covers REQ-SWEEPE-002 and REQ-SWEEPE-003 (the flagged alias lines and their block label line).

- Given operator ruling d-20261019T015908Z-220a (replace where an equivalent exists, otherwise delete; all four locales in the same change) and the pinned block of spec.md § 4,
- When cell S-11 and cells S-18-en, S-18-ja, S-18-ko, S-18-zh are run with `HEAD` in place of the pin,
- Then S-11 prints four rows, each `HEAD:docs-site/content/<locale>/utility-commands/moai-feedback.md:32:> /moai:feedback`, and S-18-xx each print exactly one row, `HEAD:<its own file>:31:<its own pinned label>` (en `# Direct slash command`, ja `# 直接スラッシュコマンド`, ko `# 직접 슬래시 명령`, zh `# 直接斜杠命令`), and every cell exits 0. Each S-18 cell lists only its own file and its own label, so a label in the wrong language, at another line, or present twice fails it.
- RED now: S-11 and all four S-18 cells print nothing and exit 1. Why red: no feedback page carries a `> /moai:feedback` line yet and none carries a pinned label.
- Mutant probe (fixtures, controls C-1 to C-3): keeping the old comment `# Short aliases` at line 31 fails S-18-xx; writing the English label into the ja, ko, and zh pages fails S-18-ja, S-18-ko, and S-18-zh while S-18-en still passes; fixing the English page only leaves three S-11 rows missing. Keeping the three alias lines and adding the new line passes this AC and fails AC-PRC-001a.

### AC-PRC-001e — the strict i18n check still passes

- Given the edited tree,
- When `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` runs from the worktree root (cell S-10),
- Then it exits 0 with `Errors: 0` and `Warnings: 0`.
- Green now at the pin (153 files per locale).

### AC-PRC-001f — whole-tree scope fence, bounded diff, and content fence

Covers REQ-SWEEPE-004 (the edit set is the flagged lines, the Q17 counterpart lines, and the block label lines; every other line is byte-identical).

- Given the base `74ad197380a0` (S-00 confirmed at pre-flight),
- When cells S-13a, S-13b, S-16-en, S-16-ja, S-16-ko, S-16-zh, and S-17 are run after the final run-phase commit,
- Then S-13a prints exactly these twelve paths, in this order, and exits 0; S-13b prints exactly the twelve numstat rows below, in this order, and exits 0; each S-16-xx prints exactly the whole-line diff block recorded in the ledger for its locale and exits 0; and S-17 prints exactly the twelve blob ids recorded in the ledger, in the order of the paths below, and exits 0:

```
1	1	docs-site/content/en/claude-code/foundations/_index.md
1	1	docs-site/content/en/core-concepts/verification-claim-integrity.md
1	1	docs-site/content/en/getting-started/faq.md
2	4	docs-site/content/en/utility-commands/moai-feedback.md
1	1	docs-site/content/ja/claude-code/foundations/_index.md
1	1	docs-site/content/ja/getting-started/faq.md
2	4	docs-site/content/ja/utility-commands/moai-feedback.md
1	1	docs-site/content/ko/getting-started/faq.md
2	4	docs-site/content/ko/utility-commands/moai-feedback.md
1	1	docs-site/content/zh/claude-code/foundations/_index.md
1	1	docs-site/content/zh/getting-started/faq.md
2	4	docs-site/content/zh/utility-commands/moai-feedback.md
```

- The `2 4` rows are the four moai-feedback pages: two inserted lines (the pinned label and `> /moai:feedback`) against four deleted lines (`# Short aliases` or its translation, and the three alias lines). The label line is in the edit set of REQ-SWEEPE-004 (spec.md § 2, decision QE-6), so the numstat, the requirement, and the edit plan agree.
- RED now: S-13a, S-13b, and the four S-16 cells print nothing and exit 0; S-17 prints the twelve blob ids of the unedited files and exits 0 (they differ from the pinned ids). Why red: no page file has been edited, so the expected rows are absent. The fence does not hide a page file (the pathspec excludes only this SPEC directory and `.moai/reports/t1725`), so the page edits flip it.
- Mutant probe: an edit to any file outside the page files adds a path to S-13a and fails it (control C-9). Deleting a section, a file, or the whole verification-claim paragraph changes a count and fails S-13b. Editing the sibling FAQ bullet adds a second insertion and deletion to a FAQ row and fails S-13b. **The same-count overwrite mutant** — delete the three alias lines, add `> /moai:feedback`, keep the old comment `# Short aliases`, and overwrite one unrelated sentence of the page — keeps the numstat at `2 4` per feedback page and passes S-01, S-02, S-11, and S-10; it fails S-18-xx (the pinned label is absent from line 31), S-16-xx (the diff prints a second hunk and a different new blob id), and S-17 (a different blob id). Observed on fixtures, control C-2. A variant that edits the pinned lines correctly and also overwrites an unrelated sentence has numstat `3 5` and fails S-13b, S-16-xx, and S-17 (control C-4). A same-line edit inside the verification-claim paragraph beyond the two pinned deletions passes S-15 and the numstat and fails S-17 (control C-7).

### AC-PRC-001g — marker discipline

Covers REQ-SWEEPE-001.

- Given the marker definition (REQ-SWEEPE-001),
- When S-03's command is run with `HEAD` in place of the pin,
- Then it prints nothing and exits 1: no marker was added. (QE-2's alternate has its own criteria, § H.)
- Green now (precondition); regression-guard. The requirement's deliverable is the definition in spec.md § 2; the run phase adds zero markers, so this criterion can only guard that none is added to exempt a flagged line. Its residual is stated in § G.

### AC-PRC-001h — every inline `moai <verb>` token resolves

Covers REQ-PRC-001.

- Given the twelve page files at the run-phase HEAD,
- When cells S-06a (the retired form `moai brain`), S-06b (a first-level `moai <word>` outside the verified set), and S-06c (a sub-verb outside its own parent's verified set) are run with `HEAD` in place of the pin,
- Then each prints nothing and exits 1. The verified sets in S-06b and S-06c come from the tree-build verb surface (S-06d-1 to S-06d-5): first level `init`, `update`, `profile`, `version`, `plan`, `feedback`, `spec`; sub-verbs keyed by parent: `profile setup`, `feedback scrub`, `feedback queue`, `spec audit`. A replacement that cites a verb outside those sets updates S-06b, S-06c, and S-06d together (blocker report), never the sets alone.
- RED now: S-06a and S-06b each print one row (`en/core-concepts/verification-claim-integrity.md:1`) and exit 0. Why red: line 84 cites the retired CLI form `moai brain`.
- Mutant probe: a new nonexistent verb in a replacement line, for example `moai does-not-exist-t1725`, passes AC-PRC-001a to g and fails S-06b (control C-8). A real sub-verb under the wrong parent — `moai spec queue`, `moai feedback setup`, `moai profile audit` — passes the earlier unkeyed S-06c and fails the parent-keyed S-06c (control C-6).

### AC-PRC-001i — FAQ replacement says the right thing in every locale

Covers REQ-SWEEPE-006 and REQ-SWEEPE-003.

- Given the pinned FAQ lines of spec.md § 4,
- When cells S-12a-en, S-12a-ja, S-12a-ko, S-12a-zh, and S-12b are run with `HEAD` in place of the pin,
- Then each S-12a-xx prints exactly one row, `HEAD:<its own faq.md>:<N>:<its own pinned line>` (N = 24 for en, ja, and zh; 28 for ko), and exits 0; each cell lists only its own file and its own locale's pinned line, so a FAQ page written in another language fails its own cell. S-12b prints nothing and exits 1 (no FAQ line pairs `/moai init` or `/moai project` with a "does nothing" phrase in any of the four languages).
- RED now: the four S-12a-xx cells print nothing and exit 1; S-12b prints four rows and exits 0. Why red: the FAQ pages still carry the false claim and none carries the pinned line.
- Mutant probe, all observed on fixtures (controls C-5): keeping the false claim with the other command (`Typing /moai project in Claude Code does nothing`) fails S-12a and S-12b; swapping the locations (`inside the terminal, generate project documentation with /moai project`) fails S-12a; appending a trailing false clause to the pinned line fails S-12a and S-12b; **writing the English pinned line into the ja, ko, and zh pages** fails S-12a-ja, S-12a-ko, and S-12a-zh (the earlier single-alternation cell printed four rows for it); placing the pinned line at another line and keeping the false claim prints a wrong line number in S-12a-en and four rows in S-12b.

### AC-PRC-001j — foundations index rows: the three locales that carry the clause agree

Covers REQ-SWEEPE-003 (Q17 counterpart lines).

- Given the pinned foundations rows of spec.md § 4,
- When cells S-14-en, S-14-ja, and S-14-zh are run with `HEAD` in place of the pin,
- Then each prints exactly one row, `HEAD:<its own _index.md>:<N>:<its own pinned row>` (N = 39 for en; 40 for ja and zh), and exits 0. Each cell lists only its own file and its own whole pinned row.
- RED now: the three cells print nothing and exit 1. Why red: the en row still reads "the /moai relationship" and the ja and zh rows still read "の関係" and "的关系".
- Mutant probe (controls C-5): the English row written into the ja and zh pages fails S-14-ja and S-14-zh (the earlier three-string cell printed three rows for it).

### AC-PRC-001k — the verification-claim paragraph keeps its remaining producers

Covers REQ-SWEEPE-002 (delete only the two retired items).

- Given the pinned retained fragment of spec.md § 4,
- When cell S-15 is run with `HEAD` in place of the pin,
- Then it prints one row (`HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`) and exits 0.
- RED now: S-15 prints nothing and exits 1. Why red: the producer list still contains the two retired items between the pinned names.
- Mutant probe: deleting the whole paragraph or rewriting it fails S-15 and the S-13b count; leaving either retired item fails AC-PRC-001a or AC-PRC-001h; deleting the two items and also altering another word of the same line passes S-15 and fails S-17 (control C-7).

## D. Evidence ledger (verbatim; pin 74ad197380a0 unless a cell names another tree)

### S-00 — tree identity between the pin and the revision-1 HEAD

- Command: `git diff --stat 74ad197380a0 HEAD -- docs-site .claude internal cmd scripts`
- Tree: HEAD `e85847fd331b14320d4825338d2923830f4d1bb0`. Exit code: 0. Stdout: empty.

### S-01 — closing pattern on the twelve page files (RED-now)

- Command: `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 0. Re-executed in revision 0.3.0.
- Stdout (verbatim; 10 rows; 18 matching lines):

```
74ad197380a0:docs-site/content/en/claude-code/foundations/_index.md:1
74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1
74ad197380a0:docs-site/content/en/getting-started/faq.md:1
74ad197380a0:docs-site/content/en/utility-commands/moai-feedback.md:3
74ad197380a0:docs-site/content/ja/getting-started/faq.md:1
74ad197380a0:docs-site/content/ja/utility-commands/moai-feedback.md:3
74ad197380a0:docs-site/content/ko/getting-started/faq.md:1
74ad197380a0:docs-site/content/ko/utility-commands/moai-feedback.md:3
74ad197380a0:docs-site/content/zh/getting-started/faq.md:1
74ad197380a0:docs-site/content/zh/utility-commands/moai-feedback.md:3
```

Per-file lines (from `git grep -n` of the same pattern at HEAD, content-identical to the pin): en faq 24; ja faq 24; ko faq 28; zh faq 24; every moai-feedback page 32, 33, 34; en foundations 39; en verification-claim-integrity 84. The block label line 31 of every moai-feedback page (`# Short aliases` and its translations) is not a closing-pattern match: the page prints `:3`, not `:4` (spec.md § 2, block label line). The ja and zh foundations files are in the sweep list and print no row.

### S-02 — complement: any slash word outside the registered and router set (RED-now)

- Command: `git grep -c -P "/moai[: ](?!(feedback|loop|run|mx|fix|plan|project)(?![A-Za-z0-9_-])|-(\s|$))[A-Za-z0-9_-]+" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 0. Re-executed in revision 0.3.0.
- Stdout (verbatim; identical to S-01):

```
74ad197380a0:docs-site/content/en/claude-code/foundations/_index.md:1
74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1
74ad197380a0:docs-site/content/en/getting-started/faq.md:1
74ad197380a0:docs-site/content/en/utility-commands/moai-feedback.md:3
74ad197380a0:docs-site/content/ja/getting-started/faq.md:1
74ad197380a0:docs-site/content/ja/utility-commands/moai-feedback.md:3
74ad197380a0:docs-site/content/ko/getting-started/faq.md:1
74ad197380a0:docs-site/content/ko/utility-commands/moai-feedback.md:3
74ad197380a0:docs-site/content/zh/getting-started/faq.md:1
74ad197380a0:docs-site/content/zh/utility-commands/moai-feedback.md:3
```

- Reading: the `(?![A-Za-z0-9_-])` boundary is required; a `\b` boundary mis-flags Korean particles (`/moai run을`) and produced three false rows in the ko feedback page during authoring. Derived census of every `/moai[: ]word` token in the pre-revision ten files (`sort | uniq -c`): feedback 68, loop 18, run 12, mx 12, fix 10, plan 8, `:feedback` 4, issue 4, init 4, fb 4, bug 4, `-` 4, relationship 1, project 1, brain 1. The six unresolved-or-alias words are exactly the six words in the closing pattern, so the pattern is complete for these files.

### S-03 — no example marker in the page files

- Command: `git grep -c -E "<!--[[:space:]]*slash-example:" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 1. Stdout: empty. Re-executed in revision 0.3.0.

### S-04 — no HTML comment at all in the two narrative lines' files

- Command: `git grep -c -E "<!--" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 1. Stdout: empty. Re-executed in revision 0.3.0.
- Reading: answers the Q24 start-of-work check. FAQ line 24 and verification-claim-integrity line 84 carry no marker of any HTML-comment form. Line 84 sits under the heading "Worked Example — Retention-Claim Hazard" (line 80), which the marker definition does not count (scope: the flagged line or the line above).

### S-05 — CLI families on the page files (precondition, green)

All three re-executed in revision 0.3.0: exit 1, stdout empty.

- S-05a: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.
- S-05b: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.
- S-05c: `git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.

### S-06 — CLI verb resolution (AC-PRC-001h)

- S-06a (the retired form; RED now; re-executed in revision 0.3.0): `git grep -c -E "(^|[^/])moai brain" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 0 — stdout: `74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`
- S-06b (any first-level `moai <word>` outside the verified set; RED now; re-executed in revision 0.3.0): `git grep -c -P "(^|[^/\w-])moai (?!(init|update|profile|version|plan|feedback|spec)(?![\w-]))[a-z][a-z-]*" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 0 — stdout: `74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`
- S-06c (any sub-verb outside its own parent's verified set; green; **parent-keyed in revision 0.3.0**, re-executed in revision 0.3.0): `git grep -c -P "(^|[^/\w-])moai (profile (?!setup(?![\w-]))|feedback (?!(scrub|queue)(?![\w-]))|spec (?!audit(?![\w-])))[a-z][a-z-]*" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty. The earlier form accepted any of the four sub-verbs under any parent; each alternative now admits only its own parent's sub-verbs.
- S-06d (the verb surface the sets come from; derived from the tree build, see § A for the build command and provenance; re-executed in revision 0.3.0 at HEAD e85847fd33). The binary path in each command is `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/84345352-92b4-4559-a09f-c8ce19d8e034/scratchpad/moai-t1725`, written `<bin>` below for line length only; the run used the full path.
  - S-06d-1: `<bin> brain` — exit 1 — stdout: empty (0 bytes) — stderr (654 bytes, non-blank lines): `ERROR`, `Unknown command "brain" for "moai"`, `Did you mean this?`, `chain`, `.`, `Try --help for usage.`
  - S-06d-2: `<bin> --help` — exit 0 — stdout 101 lines (over the 50-line bound, so redirected to a file and read with `grep -n -E "^ +(plan|feedback|init|update|spec|profile|version) "`, a separate read of that file): line 37 `plan [command] [--flags]`, 38 `feedback [command] [--flags]`, 51 `init [project-name] [--flags]`, 54 `update [--flags]`, 61 `spec [command]`, 75 `profile [command] [--flags]`, 94 `version`, each followed by its description (whitespace between columns collapsed here).
  - S-06d-3: `<bin> spec --help` — exit 0 — the COMMANDS block lists `status`, `drift`, `view`, `lint`, `close`, `audit [--flags]  Audit SPEC era classification and detect modern-era status drift`, `archive`, `ceiling` (excerpt; whitespace collapsed).
  - S-06d-4: `<bin> profile --help` — exit 0 — the COMMANDS block lists `list`, `current`, `delete [name]`, `setup [name]  Interactive setup wizard for profile preferences` (excerpt).
  - S-06d-5: `<bin> feedback --help` — exit 0 — the COMMANDS block lists `scrub [--flags]  Mask and classify a feedback report read from stdin`, `queue [command]  Retry queue for a scrubbed report whose submission failed`, `participation [command]` (excerpt).
  - Reading: the verified sets of S-06b and S-06c are the verbs the page files use, a subset of the tree surface (`spec` also has `lint`, `drift`, and others; `feedback` also has `participation`). A new page citation of any of those is outside the sets until S-06b, S-06c, and S-06d are updated together. Stdout of S-06d-2 to S-06d-5 is an excerpt, not the full output (gap G-14). Census of every `moai <word>` token in the page files (`git grep -o -h -P "(^|[^/\w-])moai [a-z][a-z-]*( [a-z][a-z-]*)?"`, derived): `moai init` (also `moai init my-project`), `moai update`, `moai version`, `moai profile setup`, `moai plan`, `moai feedback scrub`, `moai feedback queue`, `moai spec audit`, `moai brain`. Every one but `brain` is in the verified sets.

### S-07 — the other locales carry no flagged form for the two English-only pages

- Command: `git grep -n -E "/moai (brain|relationship)|moai brain|/moai(:| )(design|issue|bug|fb|init)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/ja/claude-code docs-site/content/ko/claude-code docs-site/content/zh/claude-code docs-site/content/ja/core-concepts docs-site/content/ko/core-concepts docs-site/content/zh/core-concepts`
- Exit code: 1. Stdout: empty. Re-executed in revision 0.3.0.
- Reading: the ja and zh foundations rows read `組み込み・カスタムコマンドと /moai の関係` (ja line 40) and `内置·自定义命令与 /moai 的关系` (zh line 40); they translate the English clause but are not closing-pattern matches. They are the Q17 counterpart lines and are edited under decision QE-4 (AC-PRC-001j). The ko row (line 42) has no `/moai` clause. The ja, ko, and zh verification-claim-integrity pages carry no retired-feature text.

### S-08 — router Priority 1 (supporting; resolution evidence)

- Command: `git grep -n -E "^- \*\*(project|feedback)\*\*" 74ad197380a0 -- .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md`
- Exit code: 0. Stdout (verbatim; re-executed in revision 0.3.0):

```
74ad197380a0:.claude/skills/moai/SKILL.md:69:- **project** (aliases: init): Project documentation generation
74ad197380a0:.claude/skills/moai/SKILL.md:70:- **feedback** (aliases: fb): GitHub issue creation
74ad197380a0:internal/template/templates/.claude/skills/moai/SKILL.md:69:- **project** (aliases: init): Project documentation generation
74ad197380a0:internal/template/templates/.claude/skills/moai/SKILL.md:70:- **feedback** (aliases: fb): GitHub issue creation
```

- Reading: a text read of a prompt-level router; no executable resolver exists (gap G-1). No Priority 1 name or alias is `bug`, `issue`, `relationship`, `brain`, or `design`. The umbrella's word "unregistered" for `/moai fb` is looser than this resolution (decision QE-3).

### S-09 — registered slash commands (supporting)

- Command: `git ls-tree -r --name-only 74ad197380a0 -- .claude/commands/moai`
- Exit code: 0. Stdout: 17 files (`clean`, `codemaps`, `e2e`, `feedback`, `fix`, `gate`, `goal`, `gtd`, `harness`, `loop`, `mx`, `plan`, `project`, `review`, `run`, `sync`, `todo`, each `.claude/commands/moai/<name>.md`); no `init`, no `fb`, no `moai.md`. Re-executed in revision 0.3.0.

### S-10 — strict i18n check (regression-guard, green)

- Command: `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` (worktree root as working directory; the script writes only to a `mktemp` directory).
- Tree: HEAD `e85847fd331b14320d4825338d2923830f4d1bb0` (content-identical to the pin for docs-site, S-00). Exit code: 0. Re-executed in revision 0.3.0.
- Stdout (verbatim tail): `=== Summary ===`, `Errors:   0`, `Warnings: 0`, `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.`

### S-11 — replacement line, feedback pages, anchored at its position (RED-now; green path prints four rows)

- Command: `git grep -n -P "^> /moai:feedback$" 74ad197380a0 -- docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md`
- Exit code: 1. Stdout: empty.
- Green path: milestone M1 writes the line at line 32 of each page; the same command with `HEAD` after the M1 commit prints four rows, `HEAD:docs-site/content/<en|ja|ko|zh>/utility-commands/moai-feedback.md:32:> /moai:feedback`, and exits 0.

### S-12a — FAQ pinned lines, one cell per locale, whole line at its position (RED-now; green path prints one row each)

Four cells, each listing only its own file and only its own locale's pinned line. Each command is written as a fenced block, because the pattern carries backticks, quotes, and CJK that a table cell or an inline span would mangle; each is single-quoted so the shell does not expand the backticks. `\Q...\E` quotes the whole line; `^` and `$` anchor it.

S-12a-en:

```text
git grep -n -P '^\Q- `moai init` is a terminal command that sets up a project; inside Claude Code, generate project documentation with `/moai project`.\E$' 74ad197380a0 -- docs-site/content/en/getting-started/faq.md
```

S-12a-ja:

```text
git grep -n -P '^\Q- `moai init` はプロジェクトを設定するターミナルコマンドで、Claude Code の中では `/moai project` でプロジェクト文書を生成します。\E$' 74ad197380a0 -- docs-site/content/ja/getting-started/faq.md
```

S-12a-ko:

```text
git grep -n -P '^\Q- `moai init`은 프로젝트를 설정하는 터미널 명령이고, Claude Code 안에서는 `/moai project`로 프로젝트 문서를 생성합니다.\E$' 74ad197380a0 -- docs-site/content/ko/getting-started/faq.md
```

S-12a-zh:

```text
git grep -n -P '^\Q- `moai init` 是用于设置项目的终端命令；在 Claude Code 中，用 `/moai project` 生成项目文档。\E$' 74ad197380a0 -- docs-site/content/zh/getting-started/faq.md
```

- Exit code (each of the four, re-executed in revision 0.3.0): 1. Stdout: empty.
- Green path: milestone M1 writes the pinned line in each locale at the line it replaces (en 24, ja 24, ko 28, zh 24); the same command with `HEAD` after the M1 commit prints exactly one row, `HEAD:<file>:<N>:<the pinned line>`, and exits 0. Observed on the fixture branch `good` of control C-5: `good:docs-site/content/en/getting-started/faq.md:24:- ...`, ja `:24:`, ko `:28:`, zh `:24:`.

### S-12b — FAQ false claim banned (RED-now; green path prints nothing)

- Command: `git grep -c -E "/moai (init|project).*(does nothing|動作しません|동작하지 않습니다|不会生效)|(does nothing|動作しません|동작하지 않습니다|不会生效).*/moai (init|project)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md`
- Exit code: 0. Stdout (verbatim; re-executed in revision 0.3.0):

```
74ad197380a0:docs-site/content/en/getting-started/faq.md:1
74ad197380a0:docs-site/content/ja/getting-started/faq.md:1
74ad197380a0:docs-site/content/ko/getting-started/faq.md:1
74ad197380a0:docs-site/content/zh/getting-started/faq.md:1
```

- Green path: after the M1 commit the same command with `HEAD` prints nothing and exits 1. The sibling `moai plan` bullet is not matched (it cites neither `/moai init` nor `/moai project`).

### S-13a — whole-tree scope fence, names (RED-now)

- Command: `git diff --name-only 74ad197380a0 HEAD -- . ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001' ':!.moai/reports/t1725'`
- Tree: HEAD `e85847fd331b14320d4825338d2923830f4d1bb0`. Exit code: 0. Stdout: empty (re-executed in revision 0.3.0).
- Green path: after the last run-phase commit the same command prints the twelve paths of AC-PRC-001f, in that order.

### S-13b — whole-tree scope fence, counts (RED-now)

- Command: `git diff --numstat 74ad197380a0 HEAD -- . ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001' ':!.moai/reports/t1725'`
- Tree: HEAD `e85847fd331b14320d4825338d2923830f4d1bb0`. Exit code: 0. Stdout: empty (re-executed in revision 0.3.0).
- Green path: the twelve rows of AC-PRC-001f, tab-separated.

### S-14 — foundations rows, one cell per locale, whole row at its position (RED-now; green path prints one row each)

Three cells, each listing only its own file and its own whole pinned row (fenced and single-quoted for the same reason as S-12a).

S-14-en:

```text
git grep -n -P '^\Q| [Slash Commands](/en/claude-code/foundations/commands) | Built-in and custom commands, and where `/moai` sits |\E$' 74ad197380a0 -- docs-site/content/en/claude-code/foundations/_index.md
```

S-14-ja:

```text
git grep -n -P '^\Q| [スラッシュコマンド](/ja/claude-code/foundations/commands) | 組み込み・カスタムコマンドと /moai の位置づけ |\E$' 74ad197380a0 -- docs-site/content/ja/claude-code/foundations/_index.md
```

S-14-zh:

```text
git grep -n -P '^\Q| [斜杠命令](/zh/claude-code/foundations/commands) | 内置·自定义命令与 /moai 的定位 |\E$' 74ad197380a0 -- docs-site/content/zh/claude-code/foundations/_index.md
```

- Exit code (each of the three, re-executed in revision 0.3.0): 1. Stdout: empty.
- Green path: milestone M2 replaces the row at its line (en 39, ja 40, zh 40); the same command with `HEAD` after the M2 commit prints exactly one row, `HEAD:<file>:<N>:<the pinned row>`, and exits 0 (observed on fixture branch `good`, control C-5: `:39:`, `:40:`, `:40:`).

### S-15 — verification-claim paragraph, pinned retained fragment (RED-now; green path prints one row)

- Command (fenced, single-quoted, for the same reason as S-12a):

```text
git grep -c -F -e 'every one was already gone: `workflows/brain.md`, the `manager-brain` agent, the `/moai project --from-brain` flag, the `templates/.moai/brain/` scaffold, and the docs-site pages in all four locales.' 74ad197380a0 -- docs-site/content/en/core-concepts/verification-claim-integrity.md
```

- Exit code: 1. Stdout: empty (re-executed in revision 0.3.0).
- Green path: milestone M2 deletes the two retired items; the same command with `HEAD` after the M2 commit prints `HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`. The rest of the line is fenced by S-17.

### S-16 — feedback pages, whole-line diff of the changed lines against the pinned text (RED-now; green path prints the block)

Four cells, one per locale. The command compares the pin with `HEAD` for the one file and prints only the changed lines (zero context) plus the content-addressed `index` line, so the new blob id fences every byte of the file. The `--abbrev=12` flag fixes the id width, so the output does not depend on the repository's abbreviation setting.

- S-16-en: `git diff -U0 --no-color --abbrev=12 74ad197380a0 HEAD -- docs-site/content/en/utility-commands/moai-feedback.md`
- S-16-ja: `git diff -U0 --no-color --abbrev=12 74ad197380a0 HEAD -- docs-site/content/ja/utility-commands/moai-feedback.md`
- S-16-ko: `git diff -U0 --no-color --abbrev=12 74ad197380a0 HEAD -- docs-site/content/ko/utility-commands/moai-feedback.md`
- S-16-zh: `git diff -U0 --no-color --abbrev=12 74ad197380a0 HEAD -- docs-site/content/zh/utility-commands/moai-feedback.md`
- Tree: HEAD `e85847fd331b14320d4825338d2923830f4d1bb0`. Exit code (each of the four): 0. Stdout: empty (re-executed in revision 0.3.0). Why red: HEAD still equals the pin for these files.
- Green path (stdout verbatim from the fixture branch `good` of control C-1, the pinned edits applied to the pin blobs; the old blob ids equal the pin's blob ids). After the run phase's last commit the same command prints exactly:

```text
diff --git a/docs-site/content/en/utility-commands/moai-feedback.md b/docs-site/content/en/utility-commands/moai-feedback.md
index a74cfdaac50b..2bd283a9d1a3 100644
--- a/docs-site/content/en/utility-commands/moai-feedback.md
+++ b/docs-site/content/en/utility-commands/moai-feedback.md
@@ -31,4 +31,2 @@ Use this command when you find a bug while using MoAI-ADK, need a new feature, o
-# Short aliases
-> /moai fb
-> /moai bug
-> /moai issue
+# Direct slash command
+> /moai:feedback
```

```text
diff --git a/docs-site/content/ja/utility-commands/moai-feedback.md b/docs-site/content/ja/utility-commands/moai-feedback.md
index 87ac8379f286..15386d1976fe 100644
--- a/docs-site/content/ja/utility-commands/moai-feedback.md
+++ b/docs-site/content/ja/utility-commands/moai-feedback.md
@@ -31,4 +31,2 @@ MoAI-ADK を使っていてバグを発見したり、新しい機能が必要
-# 短いエイリアス
-> /moai fb
-> /moai bug
-> /moai issue
+# 直接スラッシュコマンド
+> /moai:feedback
```

```text
diff --git a/docs-site/content/ko/utility-commands/moai-feedback.md b/docs-site/content/ko/utility-commands/moai-feedback.md
index bc6086e4c65d..27c00aab4302 100644
--- a/docs-site/content/ko/utility-commands/moai-feedback.md
+++ b/docs-site/content/ko/utility-commands/moai-feedback.md
@@ -31,4 +31,2 @@ MoAI-ADK를 쓰다가 버그를 만났거나, 아쉬운 기능이 있거나, 개
-# 짧은 별칭
-> /moai fb
-> /moai bug
-> /moai issue
+# 직접 슬래시 명령
+> /moai:feedback
```

```text
diff --git a/docs-site/content/zh/utility-commands/moai-feedback.md b/docs-site/content/zh/utility-commands/moai-feedback.md
index e542d848c744..dbc9b9f89eb6 100644
--- a/docs-site/content/zh/utility-commands/moai-feedback.md
+++ b/docs-site/content/zh/utility-commands/moai-feedback.md
@@ -31,4 +31,2 @@ draft: false
-# 简短别名
-> /moai fb
-> /moai bug
-> /moai issue
+# 直接斜杠命令
+> /moai:feedback
```

- Reading: one hunk per file, `@@ -31,4 +31,2 @@`, and the changed lines are exactly the four deleted and two inserted lines. A second hunk, a different line, or a different new blob id changes the stdout and fails the cell.

### S-17 — content fence, the blob id of every page file (RED-now; green path prints the twelve pinned ids)

- Command: `git rev-parse HEAD:docs-site/content/en/claude-code/foundations/_index.md HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md HEAD:docs-site/content/en/getting-started/faq.md HEAD:docs-site/content/en/utility-commands/moai-feedback.md HEAD:docs-site/content/ja/claude-code/foundations/_index.md HEAD:docs-site/content/ja/getting-started/faq.md HEAD:docs-site/content/ja/utility-commands/moai-feedback.md HEAD:docs-site/content/ko/getting-started/faq.md HEAD:docs-site/content/ko/utility-commands/moai-feedback.md HEAD:docs-site/content/zh/claude-code/foundations/_index.md HEAD:docs-site/content/zh/getting-started/faq.md HEAD:docs-site/content/zh/utility-commands/moai-feedback.md`
- Tree: HEAD `e85847fd331b14320d4825338d2923830f4d1bb0`. Exit code: 0. Stdout (verbatim; the blob ids of the unedited files, re-executed in revision 0.3.0; they equal the pin's blob ids):

```text
d7d6bdc523c99b2c8adc490675be9cc1735db4c6
1223289b8580c53b8c14bc6bdc18b0449c7ddf82
62fd344662394b23841e040104c8e90a1240292a
a74cfdaac50b8570757c65ef3bb67f7738953767
94c5b7edf3511de235b9f557c4432f68ad2cf1d6
1e0a4130869c6eb1feac89179781269b7e38b01d
87ac8379f286dc6e11ea0a75f76b6ed3056fc915
e8df4954906e5444d3932021d5159fb616f09d92
bc6086e4c65d0e32c26c7b3813533c1e04f7f9dd
efd8920f97a93544fe9a216739b3ea6f446ef2e2
7293a1c812d1454dbb17c798a272430c944c13d2
e542d848c7444af04accc195e7f0e586ac0be0e1
```

- Green path (the ids a file has once exactly the pinned edits of spec.md § 4 are applied to its pin blob; stdout verbatim from the fixture branches `good` of controls C-1 and C-5). After the run phase's last commit the same command prints exactly these twelve lines, in the order of the paths above:

```text
99b8de197d71ff79c0a9f827e9bbdc1404f919cf
62c87efeaf51e300860e9c1ed7c1c3961dd5ccf3
e7de4ec667bf81235612c50acaa5f31d1eb54fe7
2bd283a9d1a3ab5dd64b170a7115f20bc03fede3
f6fe617f76157513b1df8ff9cf6e3de59d575d19
a137a94f2cbe66f7172efdcc66cc83a67b082af5
15386d1976fecb9509a3ae481ebd7620cd27d53f
42ef2ea422f35066220b7700cab58b6a347260d1
27c00aab43024fe9126e8c219600c327967a8de9
8d10d0cfef812cc65289d8b8d5d75c56d62b25c4
26a422462ea47b4fbbed38867fe03966433ac500
dbc9b9f89eb65728dadccee5417c0e200c156c5d
```

- Reading: the id is the hash of the committed bytes, so any difference from the pinned text — a reworded sentence, an extra space, an overwritten unrelated line — changes it. It is the complete "no other byte changed" check; S-16 is its readable form for the feedback pages. A pinned-text amendment (a blocker report) re-derives this block together with the S-12a, S-14, S-15, S-16, and S-18 cells. The ids depend on LF line endings, which the repository's `.gitattributes` fixes for these files (`text`, `eol=lf`).

### S-18 — feedback pages, the pinned label exactly once at its position (RED-now; green path prints one row each)

Four cells, each listing only its own file and its own locale's pinned label (fenced and single-quoted for the same reason as S-12a).

S-18-en:

```text
git grep -n -P '^\Q# Direct slash command\E$' 74ad197380a0 -- docs-site/content/en/utility-commands/moai-feedback.md
```

S-18-ja:

```text
git grep -n -P '^\Q# 直接スラッシュコマンド\E$' 74ad197380a0 -- docs-site/content/ja/utility-commands/moai-feedback.md
```

S-18-ko:

```text
git grep -n -P '^\Q# 직접 슬래시 명령\E$' 74ad197380a0 -- docs-site/content/ko/utility-commands/moai-feedback.md
```

S-18-zh:

```text
git grep -n -P '^\Q# 直接斜杠命令\E$' 74ad197380a0 -- docs-site/content/zh/utility-commands/moai-feedback.md
```

- Exit code (each of the four, re-executed in revision 0.3.0): 1. Stdout: empty. Why red: no page carries a pinned label yet; line 31 still holds the old comment.
- Green path: milestone M1 writes the label at line 31; the same command with `HEAD` after the M1 commit prints exactly one row, `HEAD:<file>:31:<the pinned label>`, and exits 0 (observed on fixture branch `good`, control C-1: `good:...:31:# Direct slash command` and the ja, ko, zh counterparts).

### Controls (fixtures, not the page files)

The green paths and the mutants cannot be observed on the real pages in the plan phase, because the docs-site pages are run-phase work. Every control runs the cell text of the named cell, with the pin or `HEAD` replaced by a fixture branch name, against hand-written fixtures under `.moai/state/t1725-ctl/` (gitignored, untracked, removed before the commit). Two throwaway git repositories hold the page files: `repo` (the four moai-feedback pages) and `repo2` (the four FAQ pages, the three foundations pages, and the verification-claim-integrity page). Each has a `base` commit whose blobs equal the pin's blobs (en feedback `a74cfdaac50b8570757c65ef3bb67f7738953767`, en verification-claim-integrity `1223289b8580c53b8c14bc6bdc18b0449c7ddf82`), a `good` branch applying exactly the pinned edits, and mutant branches.

- C-1 (`repo`, branch `good`; S-11, S-16, S-18, numstat): `git diff --numstat base good` prints `2	4` for each of the four feedback pages; S-18-en prints `good:docs-site/content/en/utility-commands/moai-feedback.md:31:# Direct slash command` and S-18-ja, S-18-ko, S-18-zh print the matching rows at line 31; S-11 prints four rows at line 32, each `:32:> /moai:feedback`; the S-16 stdout is the four blocks recorded in S-16. The S-01 pattern prints four rows `:3` on `base` and nothing, exit 1, on `good`.
- C-2 (`repo`, branch `mut-overwrite` — the iteration-2 mutant: three alias lines deleted, `> /moai:feedback` added, old comment kept at line 31, one unrelated sentence overwritten): `git diff --numstat base mut-overwrite` prints `2	4` for all four pages (the S-13b row passes); S-11 prints four rows at line 32 (passes); S-18-en prints nothing, exit 1 (**killed**); `git diff -U0 --no-color --abbrev=12 base mut-overwrite -- <en page>` prints `index a74cfdaac50b..1a185129fde7`, a first hunk `@@ -32,3 +32 @@` and a second hunk `@@ -37 +35 @@` with `-Running the command guides you through choosing the feedback type and entering the content.` and `+OVERWRITTEN UNRELATED SENTENCE.` (**differs from the S-16-en block**, killed).
- C-3 (`repo`, branch `mut-english-comment`, the English label written into all four pages): S-18-ja, S-18-ko, and S-18-zh each print nothing, exit 1 (**killed**); S-18-en prints `mut-english-comment:docs-site/content/en/utility-commands/moai-feedback.md:31:# Direct slash command` (correct for the English page).
- C-4 (`repo`, branch `mut-extra`, the pinned edits plus one overwritten unrelated sentence): `git diff --numstat base mut-extra` prints `3	5` for all four pages (**S-13b and S-16 fail**); the S-16-zh diff prints a second hunk `@@ -37 +35 @@ draft: false`.
- C-5 (`repo2`, branches `good`, `mut-english`, `mut-position`; S-12a, S-12b, S-14): on `good`, S-12a-en prints `good:docs-site/content/en/getting-started/faq.md:24:` followed by the pinned line, S-12a-ja `:24:`, S-12a-ko `:28:`, S-12a-zh `:24:`; S-14-en prints `:39:` and S-14-ja and S-14-zh `:40:` with the pinned rows; S-12b and the closing pattern (S-01) print nothing on `good`. On `mut-english` (the English pinned FAQ line and the English foundations row written into the ja, ko, and zh pages): S-12a-ja, S-12a-ko, S-12a-zh, S-14-ja, and S-14-zh each print nothing, exit 1 (**killed**); S-12a-en and S-14-en print their rows; the earlier four-pattern single-alternation S-12a, run with `-c` over the four FAQ files, printed four rows `mut-english:...:1` (the **defect the split removes**), and the earlier three-string `-F` S-14 printed three rows. On `mut-position` (the false claim kept at line 24 and the pinned English line written over line 25): S-12a-en prints `mut-position:docs-site/content/en/getting-started/faq.md:25:` followed by the pinned line (wrong position, **differs from the expected `:24:` row**), and S-12b prints four rows.
- C-6 (`.moai/state/t1725-ctl/fx3/`; S-06c, parent-keyed, run with `--no-index -c -P`): the allowed pairs file `ok.md` (`moai spec audit`, `moai profile setup`, `moai feedback scrub`, `moai feedback queue`) prints no row; `mut-spec-queue.md`, `mut-feedback-setup.md`, and `mut-profile-audit.md` each print `:1` (**killed**), and the earlier unkeyed S-06c pattern prints nothing (exit 1) on all of them (the **defect the keying removes**).
- C-7 (`repo2`, branch `mut-vci-extra`, the two items deleted and the word "reachable" upper-cased in the same line): S-15 prints `mut-vci-extra:docs-site/content/en/core-concepts/verification-claim-integrity.md:1` (**passes**, the pinned fragment is intact); `git rev-parse mut-vci-extra:docs-site/content/en/core-concepts/verification-claim-integrity.md` prints `ff78c27aee73a360e6d5c76303ee1791d08f4dda`, not the pinned `62c87efeaf51e300860e9c1ed7c1c3961dd5ccf3` (**S-17 kills it**).
- C-8 (`.moai/state/t1725-ctl/fx5/`; S-06b, `--no-index -c -P`): `mut-newverb.md` (`moai does-not-exist-t1725`) prints `:1`; `ok.md` (`moai init`, `moai update`, `moai version`) prints no row.
- C-9 (fence sensitivity): `git diff --numstat 74ad197380a0 HEAD -- . ':!.moai/reports/t1725'` (exit 0, at HEAD e85847fd33) prints the five SPEC-directory rows (`372 0 acceptance.md`, `74 0 decision-index.md`, `67 0 plan.md`, `44 0 progress.md`, `190 0 spec.md` under `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001/`), so omitting the SPEC-directory exclusion makes S-13b list committed changes; the exclusion is what empties S-13b now. The `.moai/reports/t1725` exclusion cannot be exercised yet: `git ls-tree -r --name-only HEAD -- .moai/reports/t1725` prints nothing (exit 0), so no file under it is committed.

## E. Definition of done

- AC-PRC-001a, b, d, f, h, i, j, k pass at the run-phase HEAD with verbatim output recorded; AC-PRC-001c, e, g are re-observed green.
- The umbrella's closing check on these page files: E-04 prints nothing for them; no marked line exists at the defaults.
- Sync-phase and CI results for the pull request are recorded apart; they are not counted as passes here.

## F. Gaps (not observed)

- G-1 — Resolution through the router was read from `SKILL.md` text (S-08). The router is a prompt-level specification and no executable resolver exists, so "`/moai fb` resolves" is a read of the documented contract, not a run. A text read is a hypothesis about runtime behaviour.
- G-2 — The tree build's VCS stamp (`478285ab4c1f`, modified) does not equal the tree commit in this linked worktree. The stamp was not used; provenance rests on the build procedure and on S-00 (no Go source changed since the build's sources were read).
- G-3 — No existing slash-resolution or stale-citation checker exists in the tree: `scripts/` and `.github/workflows/` hold none, and the umbrella's E-28 records no cited-verbs producer (exit 1); the GATE producer is not yet written. The cells use `git grep` with the umbrella's families plus the complement and verb-set patterns (S-02, S-06b, S-06c). The tool-based verdict is deferred to the GATE producer at the umbrella's M6.
- G-4 — The S-02 census and the S-06d token census used a `sort | uniq -c` pipeline or `-o` extraction (derived, not cells). The verified verb sets in S-06b and S-06c are hard-coded from S-06d; a new verb needs both updated together.
- G-5 — `decision-index.md` and the SPEC-directory `progress.md` with the `§E.1` to `§E.4` skeleton were added after the first plan commit (6d335cd483). The decision rows in `decision-index.md` mirror spec.md § 6 field for field (nine rows from revision 0.3.0, QE-6 added). The `spec_audit` MCP tool result for revision 0.3.0 is recorded in progress.md § E.1.
- G-6 — The card's source file `.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md` is absent; the file and line list is verified only against the current tree.
- G-7 — Hugo was not built; the rendered output of the replacement lines was not observed.
- G-8 — The ja, ko, and zh pinned wording was written by the planner; its register and the glossary check of the new terms (`位置づけ`, `定位`, `프로젝트 문서`, `直接スラッシュコマンド`, `직접 슬래시 명령`, `直接斜杠命令`) were not reviewed by a native reviewer. S-10 was re-run on the unedited tree (revision 0.3.0); the run phase runs S-10 after the edits.
- G-9 — The operator record d-20261019T015908Z-220a was read with `moai decision read --scope card:t1725` (tree build) during revision 0.1.0; its body matches the coordinator's addendum. The revision-1 and revision-2 rulings (QE-4, QE-5, QE-6) are the planner's judgments anchored on that record and on d-20261014T060405Z-0fcd; they are not new operator verdicts.
- G-10 — The traceability verb is blind to the `REQ-PRC-001-00N` form of revision 0.1.0 (it read them as `REQ-PRC-001`); the renamed IDs are read by it (plan.md § F records the output). The AC IDs carry a lowercase suffix (`AC-PRC-001a`), which the verb reads as the base ID; mapping rows are therefore read by REQ cell, not by AC letter.
- G-11 — The green paths of S-11, S-12a-xx, S-14-xx, S-16-xx, S-17, and S-18-xx, and the mutants of AC-PRC-001d, f, h, i, j, and k, were observed on fixtures (controls C-1 to C-9), not on the real pages. The fence cells S-13a and S-13b have no green-path observation in the plan phase.
- G-12 — The ja and zh verification-claim-integrity, ko foundations, and other locales' pages were not re-read in full; S-07 is a pattern read.
- G-13 — The S-16 and S-17 green paths were derived on throwaway git repositories (`repo`, `repo2`) with the installed git, from the pin blobs and the pinned edits applied by a script. Their old blob ids equal the pin's (`git rev-parse` of the same path at the pin); the repositories carry no `.gitattributes`, while the real tree fixes `text` and `eol=lf` for these files, which leaves LF content bytes unchanged. A real run-phase commit was not made, so the real-tree output of S-16 and S-17 after the edits is not observed.
- G-14 — S-06d-2 to S-06d-5 quote excerpts of larger outputs (101-line root help; the other three are under 30 lines but whitespace-collapsed here). The relevant lines were read from the saved output of the same invocation. The scratchpad binary path belongs to this session; a later reader rebuilds with the command in § A.
- G-15 — The Group 4 traceability verb result for revision 0.3.0 is recorded in plan.md § F; the CN-4 ordering verb (`Exit:` binding check) is recorded there too.

## G. Residual risk

- The operator may read "explicit example marker" more widely than the marker defined here (for example, the "Worked Example" heading above line 84). Line 84 would then be exempt and its edit unneeded; the edit is harmless (decision QE-2, alternate criteria in § H).
- The umbrella closing pattern flags router aliases that resolve (QE-3). If the umbrella later carves them out, replacing `/moai fb` is no longer required for the check, though the replacement stays correct.
- The GATE checker may not honour the marker grammar; the interface is dormant at the defaults.
- REQ-SWEEPE-001 (the marker definition) is mapped only to the regression-guard AC-PRC-001g: its deliverable is a definition, authored at plan time, and no run-phase act can turn a RED cell for it.
- The pinned replacement lines make the run phase's output exact, so a native reviewer who dislikes a ja, ko, or zh sentence needs an amendment (a blocker report), not a local rewording. S-17 makes this strict by design: any byte that differs from the pinned text fails it.
- **Accepted residual (defect D-N7, decision QE-4).** The ja and zh foundations rows now read "位置づけ" and "定位" while the section headings of their linked pages still read "関係" and "关系". Measured at 74ad197380a0 in `claude-code/foundations/commands.md`: en line 238 "Where the /moai Command Sits", ko line 238 "/moai 명령은 어디에 얹혀 있나", ja line 178 "MoAI-ADK の /moai コマンドとの関係", zh line 178 "与 MoAI-ADK 的 /moai 命令的关系". The four headings share no wording, and a four-locale-consistent row would need the three or four unflagged heading lines edited, which is outside the twelve page files, so the heading is not added to the edit set. The en and ko headings already mean "where it sits"; the ja and zh rows follow that meaning. Reverting the ja and zh rows is the QE-4 alternate.
- The label line is in the edit set by decision QE-6 (planner default); if the operator chooses the alternate, REQ-SWEEPE-002 and -004, the pinned block, the `2 4` rows, and the S-16, S-17, and S-18 cells are re-derived together (`1 3` rows).

## H. QE-2 alternate — its own criteria

Applies only if the operator records the marker alternate of decision QE-2 (keep line 84, add `<!-- slash-example: retired-feature incident narrative -->` on the line above). At the planner default none of this applies, and AC-PRC-001a to k above are the whole set.

- AC-PRC-001a, b, h run with `docs-site/content/en/core-concepts/verification-claim-integrity.md` removed from their path lists (eleven page files), and S-01, S-02, S-06a, S-06b exit as before on the remaining files.
- AC-PRC-001g becomes release-blocking with two cells that anchor the marker to the line it covers.
  - S-H1a: `git grep -n -E "<!--[[:space:]]*slash-example:[[:space:]]*[^[:space:]-]" HEAD -- docs-site/content/en/core-concepts/verification-claim-integrity.md` prints exactly one row, `HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:84:<!-- slash-example: retired-feature incident narrative -->`, and exits 0 (a non-empty reason, on line 84). RED now: at the pin it prints nothing and exits 1 (measured again in revision 0.3.0: `74ad197380a0`, exit 1, empty stdout).
  - S-H1b: `git grep -n -o -E "/?moai brain" HEAD -- docs-site/content/en/core-concepts/verification-claim-integrity.md` prints exactly two rows, `HEAD:...:85:/moai brain` and `HEAD:...:85:moai brain`, and exits 0 (the retired citations sit on line 85, immediately below the marker). RED now: at the pin it prints the same two rows at line 84 and exits 0 (measured in revision 0.3.0), so the cell is red for the stated reason: the marker has not pushed the paragraph down by one line.
  - Mutant control C-10 (`.moai/state/t1725-ctl/fx4/`, `--no-index`): `alt-ok.md` (marker at line 84, paragraph at line 85) prints the marker at `:84:` for S-H1a and both citations at `:85:` for S-H1b; `alt-mut-far.md` (marker at line 1, paragraph at line 85) prints the marker at `:1:` for S-H1a, which differs from the expected `:84:` row (**killed**: a marker anywhere else in the page no longer passes).
- AC-PRC-001f's row for that page becomes `1	0` and S-13a is unchanged; the blob id of that page in S-17 is re-derived; AC-PRC-001k does not apply.
- The alternate does not change AC-PRC-001c, d, e, i, or j.
