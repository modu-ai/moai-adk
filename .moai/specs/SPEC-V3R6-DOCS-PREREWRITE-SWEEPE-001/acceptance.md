# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 (revision 0.2.0)

## A. Pin and conventions

- **Pin.** Tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (short `74ad197380a0`), branch WT-getting-started-faq. Measured 2026-10-19. This document-level pin binds every cell that names no other tree. Every ledger cell that reads a committed object names the commit in its command. The plan HEAD at the time of the revision-1 measurements was `bb7d928f682d34279d5d3584537fec0c3e68dd25`; cell S-00 shows the docs-site, `.claude`, `internal`, `cmd`, and `scripts` trees are identical between the pin and that HEAD, so the pin and the HEAD agree on every page file and every Go source.
- **Cell form** (verification-completeness §2.1). One read-only invocation, no `;`, `&&`, `|`, subshell, or redirection inside the cell. Stdout is verbatim and the exit code is its own field. The exit code was captured by appending `; echo "exit=$?"` outside the cell in the plain compound form; that suffix is part of the measurement, not of the cell. A zero-match `git grep` exits 1 and prints nothing; an empty stdout with exit 1 is a complete observation. Every path is written out in the command; no placeholder is used.
- **Run-time tree argument.** `git grep <commit>` reads committed objects only. Every flip check in the run phase therefore runs after the milestone's commit, with `HEAD` in place of `74ad197380a0`; an uncommitted edit is invisible to it.
- **Scope fence base.** The fence cells (S-13a, S-13b) compare `74ad197380a0` with `HEAD`. The run phase confirms S-00 at pre-flight; a non-empty S-00 means the docs-site tree moved and is a blocker report.
- **Tool provenance** (verification-claim-integrity §2.2). The verb-surface cell S-06d used a binary built by `go build -o <scratchpad>/moai-t1725 ./cmd/moai` from this worktree. `go version -m` on it reads `vcs.revision=478285ab4c1fc893b020bb464d14cb69c8422523`, `vcs.modified=true`, which does not equal the tree commit; the stamp is unreliable in this linked worktree (gap G-2). The last commit that touched `cmd` or `internal` is `74ad197380a0` itself (`git log -1 -- cmd internal`), and S-00 shows no Go source changed since, so the build procedure (sources read from this worktree) is the attribution, not the stamp.
- **Class.** Release-blocking: AC-PRC-001a, b, d, f, h, i, j, k — each has a RED-now cell with the command, verbatim stdout, exit code, and pin. Regression-guard (green at the pin, must stay green): AC-PRC-001c, e, g.

## B. Traceability

| AC | REQ | Cell(s) | RED-now? | Class |
|----|-----|---------|----------|-------|
| AC-PRC-001a | REQ-PRC-001, REQ-SWEEPE-002 | S-01 | red (10 rows, 18 matching lines, exit 0) | release-blocking |
| AC-PRC-001b | REQ-PRC-001, REQ-SWEEPE-002 | S-02 | red (10 rows, exit 0) | release-blocking |
| AC-PRC-001c | REQ-PRC-001 | S-05a, S-05b, S-05c | green (empty, exit 1) | regression-guard |
| AC-PRC-001d | REQ-SWEEPE-002, REQ-SWEEPE-003 | S-11 | red (empty, exit 1; four rows expected) | release-blocking |
| AC-PRC-001e | REQ-SWEEPE-005 | S-10 | green (exit 0) | regression-guard |
| AC-PRC-001f | REQ-SWEEPE-004 | S-13a, S-13b | red (empty, exit 0; twelve rows expected) | release-blocking |
| AC-PRC-001g | REQ-SWEEPE-001 | S-03 | green (empty, exit 1) | regression-guard |
| AC-PRC-001h | REQ-PRC-001 | S-06a, S-06b, S-06c | red (S-06a and S-06b print one row each, exit 0) | release-blocking |
| AC-PRC-001i | REQ-SWEEPE-006, REQ-SWEEPE-003 | S-12a, S-12b | red (S-12a empty, exit 1; S-12b four rows, exit 0) | release-blocking |
| AC-PRC-001j | REQ-SWEEPE-003 | S-14 | red (empty, exit 1; three rows expected) | release-blocking |
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

### AC-PRC-001d — feedback pages: replacement present in all four locales (Q13 and Q17 rulings)

Covers REQ-SWEEPE-002 and REQ-SWEEPE-003.

- Given operator ruling d-20261019T015908Z-220a (replace where an equivalent exists, otherwise delete; all four locales in the same change),
- When cell S-11 is run with `HEAD` in place of the pin,
- Then it prints four rows, each `...moai-feedback.md:1`, and exits 0; no locale is missing.
- RED now: S-11 prints nothing and exits 1. Why red: no feedback page carries a `> /moai:feedback` line yet.
- Mutant probe: fixing the English page only leaves three rows missing. Keeping the three alias lines and adding the new line passes this AC and fails AC-PRC-001a (control C-3).

### AC-PRC-001e — the strict i18n check still passes

- Given the edited tree,
- When `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` runs from the worktree root (cell S-10),
- Then it exits 0 with `Errors: 0` and `Warnings: 0`.
- Green now at the pin (153 files per locale).

### AC-PRC-001f — whole-tree scope fence and bounded diff

Covers REQ-SWEEPE-004.

- Given the base `74ad197380a0` (S-00 confirmed at pre-flight),
- When cells S-13a and S-13b are run after the final run-phase commit,
- Then S-13a prints exactly these twelve paths, in this order, and exits 0, and S-13b prints exactly the twelve numstat rows below, in this order, and exits 0:

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

- RED now: S-13a and S-13b print nothing and exit 0. Why red: no page file has been edited, so the expected twelve rows are absent. The fence does not hide a page file (the pathspec excludes only this SPEC directory and `.moai/reports/t1725`), so the page edits flip it.
- Mutant probe: an edit to any file outside the page files (a rule, README, SKILL.md, a docs-site page of another child) adds a path to S-13a and fails it; control C-5 shows the same command lists committed changes outside the excluded directories. Deleting a section, a file, or the whole verification-claim paragraph changes a count and fails S-13b. Editing the sibling FAQ bullet adds a second insertion and deletion to a FAQ row and fails S-13b.

### AC-PRC-001g — marker discipline

Covers REQ-SWEEPE-001.

- Given the marker definition (REQ-SWEEPE-001),
- When S-03's command is run with `HEAD` in place of the pin,
- Then it prints nothing and exits 1: no marker was added. (QE-2's alternate has its own criteria, § H.)
- Green now (precondition); regression-guard. The requirement's deliverable is the definition in spec.md § 2; the run phase adds zero markers, so this criterion can only guard that none is added to exempt a flagged line. Its residual is stated in § G.

### AC-PRC-001h — every inline `moai <verb>` token resolves

Covers REQ-PRC-001.

- Given the twelve page files at the run-phase HEAD,
- When cells S-06a (the retired form `moai brain`), S-06b (a first-level `moai <word>` outside the verified set), and S-06c (a sub-verb outside the verified set) are run with `HEAD` in place of the pin,
- Then each prints nothing and exits 1. The verified sets in S-06b and S-06c come from the tree-build verb surface (S-06d): first level `init`, `update`, `profile`, `version`, `plan`, `feedback`, `spec`; sub-verbs `profile setup`, `feedback scrub`, `feedback queue`, `spec audit`. A replacement that cites a verb outside those sets updates S-06b, S-06c, and S-06d together (blocker report), never the sets alone.
- RED now: S-06a and S-06b each print one row (`en/core-concepts/verification-claim-integrity.md:1`) and exit 0. Why red: line 84 cites the retired CLI form `moai brain`.
- Mutant probe: a new nonexistent verb in a replacement line, for example `moai does-not-exist-t1725`, passes AC-PRC-001a to g and fails S-06b (control C-6).

### AC-PRC-001i — FAQ replacement says the right thing in every locale

Covers REQ-SWEEPE-006 and REQ-SWEEPE-003.

- Given the pinned FAQ lines of spec.md § 4,
- When cells S-12a and S-12b are run with `HEAD` in place of the pin,
- Then S-12a prints four rows, each `...faq.md:1`, and exits 0 (each locale's whole line equals its pinned line), and S-12b prints nothing and exits 1 (no FAQ line pairs `/moai init` or `/moai project` with a "does nothing" phrase in any of the four languages).
- RED now: S-12a prints nothing and exits 1; S-12b prints four rows and exits 0. Why red: the FAQ pages still carry the false claim and none carries the pinned line.
- Mutant probe, all three observed on fixtures (controls C-1, C-2): keeping the false claim with the other command (`Typing /moai project in Claude Code does nothing`) fails S-12a and S-12b; swapping the locations (`inside the terminal, generate project documentation with /moai project`) fails S-12a; appending a trailing false clause to the pinned line fails S-12a and S-12b.

### AC-PRC-001j — foundations index rows: the three locales that carry the clause agree

Covers REQ-SWEEPE-003 (Q17 counterpart lines).

- Given the pinned wording of spec.md § 4,
- When cell S-14 is run with `HEAD` in place of the pin,
- Then it prints three rows (en, ja, zh foundations `_index.md`), each `:1`, and exits 0.
- RED now: S-14 prints nothing and exits 1. Why red: the en row still reads "the /moai relationship" and the ja and zh rows still read "の関係" and "的关系".

### AC-PRC-001k — the verification-claim paragraph keeps its remaining producers

Covers REQ-SWEEPE-002 (delete only the two retired items).

- Given the pinned retained fragment of spec.md § 4,
- When cell S-15 is run with `HEAD` in place of the pin,
- Then it prints one row (`en/core-concepts/verification-claim-integrity.md:1`) and exits 0.
- RED now: S-15 prints nothing and exits 1. Why red: the producer list still contains the two retired items between the pinned names.
- Mutant probe: deleting the whole paragraph or rewriting it fails S-15 and the S-13b count; leaving either retired item fails AC-PRC-001a or AC-PRC-001h.

## D. Evidence ledger (verbatim; pin 74ad197380a0 unless a cell names another tree)

### S-00 — tree identity between the pin and the plan HEAD

- Command: `git diff --stat 74ad197380a0 HEAD -- docs-site .claude internal cmd scripts`
- Tree: HEAD `bb7d928f682d34279d5d3584537fec0c3e68dd25`. Exit code: 0. Stdout: empty.

### S-01 — closing pattern on the twelve page files (RED-now)

- Command: `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 0.
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

Per-file lines (from `git grep -n` of the same pattern at HEAD, content-identical to the pin): en faq 24; ja faq 24; ko faq 28; zh faq 24; every moai-feedback page 32, 33, 34; en foundations 39; en verification-claim-integrity 84. The ja and zh foundations files are in the sweep list and print no row.

### S-02 — complement: any slash word outside the registered and router set (RED-now)

- Command: `git grep -c -P "/moai[: ](?!(feedback|loop|run|mx|fix|plan|project)(?![A-Za-z0-9_-])|-(\s|$))[A-Za-z0-9_-]+" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 0.
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
- Exit code: 1. Stdout: empty.

### S-04 — no HTML comment at all in the two narrative lines' files

- Command: `git grep -c -E "<!--" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 1. Stdout: empty.
- Reading: answers the Q24 start-of-work check. FAQ line 24 and verification-claim-integrity line 84 carry no marker of any HTML-comment form. Line 84 sits under the heading "Worked Example — Retention-Claim Hazard" (line 80), which the marker definition does not count (scope: the flagged line or the line above).

### S-05 — CLI families on the page files (precondition, green)

- S-05a: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.
- S-05b: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.
- S-05c: `git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.

### S-06 — CLI verb resolution (AC-PRC-001h)

- S-06a (the retired form; RED now): `git grep -c -E "(^|[^/])moai brain" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 0 — stdout: `74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`
- S-06b (any first-level `moai <word>` outside the verified set; RED now): `git grep -c -P "(^|[^/\w-])moai (?!(init|update|profile|version|plan|feedback|spec)(?![\w-]))[a-z][a-z-]*" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 0 — stdout: `74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`
- S-06c (any sub-verb outside the verified set; green): `git grep -c -P "(^|[^/\w-])moai (profile|feedback|spec) (?!(setup|scrub|queue|audit)(?![\w-]))[a-z][a-z-]*" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md` — exit 1, stdout empty.
- S-06d (the verb surface the sets come from), tree build `<scratchpad>/moai-t1725`, see § A for provenance:
  - `<scratchpad>/moai-t1725 brain` — exit 1 — output: `ERROR`, `Unknown command "brain" for "moai"`, `Did you mean this?`, `chain`.
  - `<scratchpad>/moai-t1725 --help` — exit 0 — lines read verbatim from its COMMANDS and PROJECT COMMANDS and TOOLS blocks (the full output exceeds the 50-line bound): `plan [command] [--flags]  Plan-Phase CLI utilities (render-html, …)`; `feedback [command] [--flags]  Feedback submission helpers (scrub, queue, …)`; `init [project-name] [--flags]  Initialize a new MoAI project`; `update [--flags]  Sync MoAI-ADK project templates to the latest version`; `spec [command]  Manage SPEC documents`; `profile [command] [--flags]  Manage Claude configuration profiles`; `version  Show version information`.
  - `<scratchpad>/moai-t1725 spec --help` — exit 0 — includes `audit [--flags]  Audit SPEC era classification and detect modern-era status drift`.
  - `<scratchpad>/moai-t1725 profile --help` — exit 0 — includes `setup [name]   Interactive setup wizard for profile preferences`.
  - `<scratchpad>/moai-t1725 feedback --help` — exit 0 — includes `scrub [--flags]  Mask and classify a feedback report read from stdin` and `queue [command]  Retry queue for a scrubbed report whose submission failed`.
  - Census of every `moai <word>` token in the page files (`git grep -o -h -P "(^|[^/\w-])moai [a-z][a-z-]*( [a-z][a-z-]*)?"`, derived): `moai init` (also `moai init my-project`), `moai update`, `moai version`, `moai profile setup`, `moai plan`, `moai feedback scrub`, `moai feedback queue`, `moai spec audit`, `moai brain`. Every one but `brain` is in the verified sets.

### S-07 — the other locales carry no flagged form for the two English-only pages

- Command: `git grep -n -E "/moai (brain|relationship)|moai brain|/moai(:| )(design|issue|bug|fb|init)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/ja/claude-code docs-site/content/ko/claude-code docs-site/content/zh/claude-code docs-site/content/ja/core-concepts docs-site/content/ko/core-concepts docs-site/content/zh/core-concepts`
- Exit code: 1. Stdout: empty.
- Reading: the ja and zh foundations rows read `組み込み・カスタムコマンドと /moai の関係` (ja line 40) and `内置·自定义命令与 /moai 的关系` (zh line 40); they translate the English clause but are not closing-pattern matches. They are the Q17 counterpart lines and are edited under decision QE-4 (AC-PRC-001j). The ko row (line 42) has no `/moai` clause. The ja, ko, and zh verification-claim-integrity pages carry no retired-feature text.

### S-08 — router Priority 1 (supporting; resolution evidence)

- Command: `git grep -n -E "^- \*\*(project|feedback)\*\*" 74ad197380a0 -- .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md`
- Exit code: 0. Stdout (verbatim):

```
74ad197380a0:.claude/skills/moai/SKILL.md:69:- **project** (aliases: init): Project documentation generation
74ad197380a0:.claude/skills/moai/SKILL.md:70:- **feedback** (aliases: fb): GitHub issue creation
74ad197380a0:internal/template/templates/.claude/skills/moai/SKILL.md:69:- **project** (aliases: init): Project documentation generation
74ad197380a0:internal/template/templates/.claude/skills/moai/SKILL.md:70:- **feedback** (aliases: fb): GitHub issue creation
```

- Reading: a text read of a prompt-level router; no executable resolver exists (gap G-1). No Priority 1 name or alias is `bug`, `issue`, `relationship`, `brain`, or `design`. The umbrella's word "unregistered" for `/moai fb` is looser than this resolution (decision QE-3).

### S-09 — registered slash commands (supporting)

- Command: `git ls-tree -r --name-only 74ad197380a0 -- .claude/commands/moai`
- Exit code: 0. Stdout: 17 files (`clean`, `codemaps`, `e2e`, `feedback`, `fix`, `gate`, `goal`, `gtd`, `harness`, `loop`, `mx`, `plan`, `project`, `review`, `run`, `sync`, `todo`, each `.claude/commands/moai/<name>.md`); no `init`, no `fb`, no `moai.md`.

### S-10 — strict i18n check (regression-guard, green)

- Command: `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` (worktree root as working directory; the script writes only to a `mktemp` directory).
- Tree: HEAD `bb7d928f682d34279d5d3584537fec0c3e68dd25` (content-identical to the pin for docs-site, S-00). Exit code: 0.
- Stdout (verbatim tail): `=== Summary ===`, `Errors:   0`, `Warnings: 0`, `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.` (153 `.md` files in each of ko, en, ja, zh).

### S-11 — replacement presence, feedback pages (RED-now; green path prints four rows)

- Command: `git grep -c -E "^> /moai:feedback$" 74ad197380a0 -- docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md`
- Exit code: 1. Stdout: empty.
- Green path: milestone M1 adds the line in each locale; the same command with `HEAD` after the M1 commit prints four rows, each ending `:1`.

### S-12a — FAQ pinned lines, whole-line match (RED-now; green path prints four rows)

- Command (written as a fenced block, because the pattern carries backticks and quotes that a table cell or an inline span would mangle; single-quoted so the shell does not expand the backticks):

```text
git grep -c -P '^\Q- `moai init` is a terminal command that sets up a project; inside Claude Code, generate project documentation with `/moai project`.\E$|^\Q- `moai init` はプロジェクトを設定するターミナルコマンドで、Claude Code の中では `/moai project` でプロジェクト文書を生成します。\E$|^\Q- `moai init`은 프로젝트를 설정하는 터미널 명령이고, Claude Code 안에서는 `/moai project`로 프로젝트 문서를 생성합니다.\E$|^\Q- `moai init` 是用于设置项目的终端命令；在 Claude Code 中，用 `/moai project` 生成项目文档。\E$' 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md
```

- Exit code: 1. Stdout: empty.
- Green path: milestone M1 writes the pinned line in each locale; the same command with `HEAD` after the M1 commit prints four rows, each ending `:1`.

### S-12b — FAQ false claim banned (RED-now; green path prints nothing)

- Command: `git grep -c -E "/moai (init|project).*(does nothing|動作しません|동작하지 않습니다|不会生效)|(does nothing|動作しません|동작하지 않습니다|不会生效).*/moai (init|project)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md`
- Exit code: 0. Stdout (verbatim):

```
74ad197380a0:docs-site/content/en/getting-started/faq.md:1
74ad197380a0:docs-site/content/ja/getting-started/faq.md:1
74ad197380a0:docs-site/content/ko/getting-started/faq.md:1
74ad197380a0:docs-site/content/zh/getting-started/faq.md:1
```

- Green path: after the M1 commit the same command with `HEAD` prints nothing and exits 1. The sibling `moai plan` bullet is not matched (it cites neither `/moai init` nor `/moai project`).

### S-13a — whole-tree scope fence, names (RED-now)

- Command: `git diff --name-only 74ad197380a0 HEAD -- . ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001' ':!.moai/reports/t1725'`
- Tree: HEAD `bb7d928f682d34279d5d3584537fec0c3e68dd25`. Exit code: 0. Stdout: empty.
- Green path: after the last run-phase commit the same command prints the twelve paths of AC-PRC-001f, in that order.

### S-13b — whole-tree scope fence, counts (RED-now)

- Command: `git diff --numstat 74ad197380a0 HEAD -- . ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001' ':!.moai/reports/t1725'`
- Tree: HEAD `bb7d928f682d34279d5d3584537fec0c3e68dd25`. Exit code: 0. Stdout: empty.
- Green path: the twelve rows of AC-PRC-001f, tab-separated.

### S-14 — foundations rows, three locales (RED-now; green path prints three rows)

- Command (fenced, single-quoted, for the same reason as S-12a):

```text
git grep -c -F -e 'and where `/moai` sits' -e 'カスタムコマンドと /moai の位置づけ' -e '内置·自定义命令与 /moai 的定位' 74ad197380a0 -- docs-site/content/en/claude-code/foundations/_index.md docs-site/content/ja/claude-code/foundations/_index.md docs-site/content/zh/claude-code/foundations/_index.md
```

- Exit code: 1. Stdout: empty.
- Green path: milestone M2 edits the three rows; the same command with `HEAD` after the M2 commit prints three rows, each ending `:1`.

### S-15 — verification-claim paragraph, pinned retained fragment (RED-now; green path prints one row)

- Command (fenced, single-quoted, for the same reason as S-12a):

```text
git grep -c -F -e 'every one was already gone: `workflows/brain.md`, the `manager-brain` agent, the `/moai project --from-brain` flag, the `templates/.moai/brain/` scaffold, and the docs-site pages in all four locales.' 74ad197380a0 -- docs-site/content/en/core-concepts/verification-claim-integrity.md
```

- Exit code: 1. Stdout: empty.
- Green path: milestone M2 deletes the two retired items; the same command with `HEAD` after the M2 commit prints `HEAD:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`.

### Controls (fixtures, not the page files)

The green paths and the mutants cannot be observed on the real pages in the plan phase, because the docs-site pages are run-phase work. Each control runs the same pattern as the named cell with `--no-index` against hand-written fixtures in `.moai/state/t1725-ctl/` (untracked, removed before the commit). A fixture proves what the pattern accepts and rejects; it does not prove the run-phase edit.

- C-1 (S-12a, `git grep --no-index -c -P <S-12a pattern> -- .moai/state/t1725-ctl`; exit 0): prints exactly `faq-en-ok.md:1`, `faq-ja-ok.md:1`, `faq-ko-ok.md:1`, `faq-zh-ok.md:1`. The three mutants `faq-mut-keep-claim.md` (`Typing /moai project in Claude Code does nothing — moai init is a terminal command.`), `faq-mut-swap.md` (`... inside the terminal, generate project documentation with /moai project.`), and `faq-mut-extra.md` (the pinned line plus `It does nothing otherwise.`) print no row.
- C-2 (S-12b pattern, `--no-index -c -E`; exit 0): prints `faq-mut-extra.md:1` and `faq-mut-keep-claim.md:1` and no `ok` fixture; the swap mutant is caught by C-1 only.
- C-3 (S-01 pattern and S-11 pattern): `fb-mut-aliases.md` (the old Usage fence) prints `fb-mut-aliases.md:3` for S-01 and no row for S-11; `fb-ok.md` (the replacement fence) prints `fb-ok.md:1` for S-11 and no row for S-01.
- C-4 (S-14 pattern, `-F`; exit 0): prints `fnd-en-ok.md:1`, `fnd-ja-ok.md:1`, `fnd-zh-ok.md:1`. (S-15 pattern, `-F`; exit 0): prints `vci-ok.md:1`, the post-edit paragraph.
- C-5 (fence sensitivity): `git diff --numstat 74ad197380a0 HEAD -- . ':!.moai/reports/t1725'` (exit 0) prints the five SPEC-directory rows (`197 0 acceptance.md`, `56 0 decision-index.md`, `63 0 plan.md`, `40 0 progress.md`, `136 0 spec.md` under `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001/`), so omitting the SPEC-directory exclusion makes S-13b list committed changes; the exclusion is what empties S-13b now. The `.moai/reports/t1725` exclusion cannot be exercised yet: no file under it is committed.
- C-6 (S-06b pattern, `-P`; exit 0): prints `vci-mut-newverb.md:1` for the paragraph that cites `moai does-not-exist-t1725`, and no row for `vci-ok.md` or any FAQ fixture.

## E. Definition of done

- AC-PRC-001a, b, d, f, h, i, j, k pass at the run-phase HEAD with verbatim output recorded; AC-PRC-001c, e, g are re-observed green.
- The umbrella's closing check on these page files: E-04 prints nothing for them; no marked line exists at the defaults.
- Sync-phase and CI results for the pull request are recorded apart; they are not counted as passes here.

## F. Gaps (not observed)

- G-1 — Resolution through the router was read from `SKILL.md` text (S-08). The router is a prompt-level specification and no executable resolver exists, so "`/moai fb` resolves" is a read of the documented contract, not a run. A text read is a hypothesis about runtime behaviour.
- G-2 — The tree build's VCS stamp (`478285ab4c1f`, modified) does not equal the tree commit in this linked worktree. The stamp was not used; provenance rests on the build procedure and on S-00 (no Go source changed since the build's sources were read).
- G-3 — No existing slash-resolution or stale-citation checker exists in the tree: `scripts/` and `.github/workflows/` hold none, and the umbrella's E-28 records no cited-verbs producer (exit 1); the GATE producer is not yet written. The cells use `git grep` with the umbrella's families plus the complement and verb-set patterns (S-02, S-06b, S-06c). The tool-based verdict is deferred to the GATE producer at the umbrella's M6.
- G-4 — The S-02 census and the S-06d token census used a `sort | uniq -c` pipeline or `-o` extraction (derived, not cells). The verified verb sets in S-06b and S-06c are hard-coded from S-06d; a new verb needs both updated together.
- G-5 — `decision-index.md` and the SPEC-directory `progress.md` with the `§E.1` to `§E.4` skeleton were added after the first plan commit (6d335cd483). The decision rows in `decision-index.md` mirror spec.md § 6 field for field. The `spec_audit` MCP tool (`filter_spec` this SPEC, `project_root` this worktree), re-run on the revision-0.2.0 working tree, classifies the SPEC as V3R6 (heuristic H-5) with one INFO finding (`EraAutoDetected`), `modern_era_clean: 1`, and no MUST-FIX finding.
- G-6 — The card's source file `.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md` is absent; the file and line list is verified only against the current tree.
- G-7 — Hugo was not built; the rendered output of the replacement lines was not observed.
- G-8 — The ja, ko, and zh pinned wording was written by the planner; its register and the glossary check of the new terms (`位置づけ`, `定位`, `프로젝트 문서`) were not reviewed by a native reviewer or run (S-10 was run on the unedited tree). The run phase runs S-10 after the edits.
- G-9 — The operator record d-20261019T015908Z-220a was read with `moai decision read --scope card:t1725` (tree build) during revision 0.1.0; its body matches the coordinator's addendum. The revision-1 rulings (QE-4, QE-5) are the leader's lane judgment of this revision round, anchored on that record and on d-20261014T060405Z-0fcd; they are not new operator verdicts.
- G-10 — The traceability verb is blind to the `REQ-PRC-001-00N` form of revision 0.1.0 (it read them as `REQ-PRC-001`); the renamed IDs are read by it (plan.md § F records the output). The AC IDs carry a lowercase suffix (`AC-PRC-001a`), which the verb reads as the base ID; mapping rows are therefore read by REQ cell, not by AC letter.
- G-11 — The green paths of S-11, S-12a, S-14, and S-15 and the mutants of AC-PRC-001h and i were observed on fixtures (controls C-1 to C-6), not on the real pages. The fence cells S-13a and S-13b have no green-path observation in the plan phase.
- G-12 — The ja and zh verification-claim-integrity, ko foundations, and other locales' pages were not re-read in full; S-07 is a pattern read.

## G. Residual risk

- The operator may read "explicit example marker" more widely than the marker defined here (for example, the "Worked Example" heading above line 84). Line 84 would then be exempt and its edit unneeded; the edit is harmless (decision QE-2, alternate criteria in § H).
- The umbrella closing pattern flags router aliases that resolve (QE-3). If the umbrella later carves them out, replacing `/moai fb` is no longer required for the check, though the replacement stays correct.
- The GATE checker may not honour the marker grammar; the interface is dormant at the defaults.
- REQ-SWEEPE-001 (the marker definition) is mapped only to the regression-guard AC-PRC-001g: its deliverable is a definition, authored at plan time, and no run-phase act can turn a RED cell for it.
- The pinned replacement lines make the run phase's output exact, so a native reviewer who dislikes a ja, ko, or zh sentence needs an amendment (a blocker report), not a local rewording.
- The ja and zh foundations rows now read "位置づけ" and "定位" while their linked pages' section headings still read "関係" and "关系" (decision QE-4); the English page heading reads "Where the /moai Command Sits".

## H. QE-2 alternate — its own criteria

Applies only if the operator records the marker alternate of decision QE-2 (keep line 84, add `<!-- slash-example: retired-feature incident narrative -->` on the line above). At the planner default none of this applies, and AC-PRC-001a to k above are the whole set.

- AC-PRC-001a, b, h run with `docs-site/content/en/core-concepts/verification-claim-integrity.md` removed from their path lists (eleven page files), and S-01, S-02, S-06a, S-06b exit as before on the remaining files.
- AC-PRC-001g becomes release-blocking with cell S-H1: `git grep -c -E "<!--[[:space:]]*slash-example:[[:space:]]*[^[:space:]-]" HEAD -- docs-site/content/en/core-concepts/verification-claim-integrity.md` prints one row ending `:1` and exits 0 (a non-empty reason). RED now: at the pin it prints nothing and exits 1 (measured: `74ad197380a0`, exit 1, empty stdout).
- AC-PRC-001f's row for that page becomes `1	0` and S-13a is unchanged; AC-PRC-001k does not apply.
- The alternate does not change AC-PRC-001c, d, e, i, or j.
