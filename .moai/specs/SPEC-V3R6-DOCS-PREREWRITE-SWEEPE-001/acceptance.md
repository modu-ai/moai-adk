# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001 (revision 0.1.0)

## A. Pin and conventions

- **Pin.** Tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (short `74ad197380a0`), branch WT-getting-started-faq, tracked tree clean (`git status --short` printed nothing). Measured 2026-10-19. Every ledger cell that reads a committed object names the commit in its command. The cells that run the tree build or a script (S-06c, S-06d, S-10) are attributed to the working tree at that commit.
- **Cell form** (verification-completeness §2.1). One read-only invocation, no `;`, `&&`, `|`, subshell, or redirection. Stdout is verbatim and the exit code is its own field. A zero-match `git grep` exits 1 and prints nothing; the status was confirmed by a measurement wrapper outside the cell and is marked `(wrapper-confirmed)`.
- **Tool provenance** (verification-claim-integrity §2.2). The verb-surface cells S-06c and S-06d used a binary built with `go build -o <scratchpad>/moai-t1725 ./cmd/moai` from this worktree at the pin. Its embedded VCS stamp reads `vcs.revision=478285ab4c1fc893b020bb464d14cb69c8422523`, `vcs.modified=true`, which does not equal the tree HEAD; the stamp is unreliable in this linked worktree (gap G-2). The attribution therefore rests on the build procedure (sources read from this clean tree at the pin), not on the stamp.
- **`<P10>`** stands for the ten pathspecs, in this order: `docs-site/content/{en,ja,ko,zh}/getting-started/faq.md`, `docs-site/content/{en,ja,ko,zh}/utility-commands/moai-feedback.md`, `docs-site/content/en/claude-code/foundations/_index.md`, `docs-site/content/en/core-concepts/verification-claim-integrity.md`. Cells S-01, S-11, and S-12 spell the paths out; the others use `<P10>` as shorthand for the same list, expanded when run.
- **Run-time base.** AC-PRC-001f uses `BASELINE_SHA`, captured with `git rev-parse HEAD` at run-phase pre-flight, before the first run-phase commit (umbrella remedy R2), and recorded in the progress record.
- **Class.** AC-PRC-001a, b, d, f, g: release-blocking (the Q13 and Q17 rulings carry an operator verdict). AC-PRC-001c and e: regression-guard (green at the pin, must stay green).

## B. Traceability

| AC | REQ | Cell(s) | RED-now? | Class |
|----|-----|---------|----------|-------|
| AC-PRC-001a | REQ-PRC-001, -002 | S-01 | red (18 matching lines in 10 files) | release-blocking |
| AC-PRC-001b | REQ-PRC-001, -002 | S-02, S-06a | red (18 lines in 10 files; 1 `moai brain`) | release-blocking |
| AC-PRC-001c | REQ-PRC-001 | S-05 | green (0) | regression-guard |
| AC-PRC-001d | REQ-PRC-001-003, -002 | S-11, S-12 | red (no replacement present) | release-blocking |
| AC-PRC-001e | REQ-PRC-001-005 | S-10 | green | regression-guard |
| AC-PRC-001f | REQ-PRC-001-004 | S-13 (run time) | n/a before edits | release-blocking |
| AC-PRC-001g | REQ-PRC-001-001 | S-03 | green for markers (0 at pin) | release-blocking |

## C. Acceptance criteria (Given-When-Then)

### AC-PRC-001a — no closing-pattern line remains in the ten files

- Given the ten page files at the run-phase HEAD,
- When the closing pattern (cell S-01's command, with the run-phase HEAD in place of the pin) is run,
- Then it prints nothing and exits 1, except for any line that carries a valid example marker; at the planner defaults no marker is added, so the output is empty.
- RED now: S-01 prints 10 file rows (18 matching lines) and exits 0. Why red: the 18 lines are the stale citations; this work edits exactly those lines, so editing flips it.
- Mutant probe: replacing `/moai fb` with a new unresolved word such as `/moai fbk` passes this AC; AC-PRC-001b catches it.

### AC-PRC-001b — no unresolved slash word and no `moai brain` remain

- Given the ten page files at the run-phase HEAD,
- When the complement command (cell S-02) and the `moai brain` command (cell S-06a) are run with the run-phase HEAD in place of the pin,
- Then both print nothing and exit 1: every `/moai[: ]<word>` in the ten files names a registered or router word (`feedback`, `loop`, `run`, `mx`, `fix`, `plan`, `project`) or is the link-text separator `/moai -`, and no line cites the CLI form `moai brain`.
- RED now: S-02 prints ten rows (18 lines); S-06a prints one row.

### AC-PRC-001c — the CLI families stay at zero on the ten files

- Given the ten page files,
- When the umbrella's E-01, E-02, and E-03 family commands are run restricted to the ten files (cells S-05a, S-05b, S-05c),
- Then each prints nothing and exits 1 before and after the edits.
- Green now (precondition). Guards against a replacement that introduces an absent CLI verb.

### AC-PRC-001d — four-locale parity of the replacement (Q13 and Q17 rulings)

- Given operator ruling d-20261019T015908Z-220a (replace where an equivalent exists, otherwise delete; all four locales in the same change),
- When the replacement-presence cells run with the run-phase HEAD (S-11 for the four feedback pages, S-12 for the four FAQ pages),
- Then S-11 prints four rows, each count 1, and S-12 prints four rows, each count 1; no locale is missing or differs.
- RED now: S-11 prints nothing and exits 1; S-12 prints nothing and exits 1.
- Mutant probe: fixing the English page only flips the English rows; the four-row requirement catches it.

### AC-PRC-001e — the strict i18n check still passes

- Given the edited tree,
- When `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` runs from the worktree root (cell S-10),
- Then it exits 0 with `Errors: 0` and `Warnings: 0`.
- Green now at the pin (153 files per locale).

### AC-PRC-001f — scope fence and bounded diff

- Given `BASELINE_SHA` captured at pre-flight,
- When `git diff --numstat BASELINE_SHA HEAD -- docs-site` is run after the edits (cell S-13),
- Then it lists exactly the ten page files, and per file: each FAQ page at most 1 insertion and 1 deletion; each feedback page at most 2 insertions and 4 deletions; the foundations page at most 1 insertion and 1 deletion; the verification-claim page at most 1 insertion and 1 deletion.
- Mutant probe: deleting a whole section or file satisfies AC-PRC-001a and fails these counts.

### AC-PRC-001g — marker discipline

- Given the marker definition (REQ-PRC-001-001),
- When S-03's command is run with the run-phase HEAD,
- Then at the planner defaults it prints nothing and exits 1 (zero markers added); if the operator chooses the marker alternate of QE-2, it prints exactly one row (the verification-claim page, count 1) and that line carries a non-empty reason.

## D. Evidence ledger (verbatim, pin 74ad197380a0)

### S-01 — closing pattern on the ten files (RED-now)

- Command: `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md docs-site/content/en/claude-code/foundations/_index.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 0.
- Stdout (verbatim; 10 rows; derived 18 matching lines):

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

Per-file lines (from `git grep -n` of the same pattern): en faq 24; ja faq 24; ko faq 28; zh faq 24; every moai-feedback page 32, 33, 34; en foundations 39; en verification-claim-integrity 84.

### S-02 — complement: any slash word outside the registered and router set (RED-now)

- Command: `git grep -c -P "/moai[: ](?!(feedback|loop|run|mx|fix|plan|project)(?![A-Za-z0-9_-])|-(\s|$))[A-Za-z0-9_-]+" 74ad197380a0 -- <P10>`
- Exit code: 0.
- Stdout (verbatim): the same ten rows and counts as S-01.
- Reading: the `(?![A-Za-z0-9_-])` boundary is required; a `\b` boundary mis-flags Korean particles (`/moai run을`) and produced three false rows in the ko feedback page during authoring. Derived census of every `/moai[: ]word` token in the ten files (`sort | uniq -c`): feedback 68, loop 18, run 12, mx 12, fix 10, plan 8, `:feedback` 4, issue 4, init 4, fb 4, bug 4, `-` 4, relationship 1, project 1, brain 1. The six unresolved-or-alias words are exactly the six words in the closing pattern, so the pattern is complete for these files.

### S-03 — no example marker in the ten files

- Command: `git grep -c -E "<!--[[:space:]]*slash-example:" 74ad197380a0 -- <P10>`
- Exit code: 1 (wrapper-confirmed). Stdout: no output.

### S-04 — no HTML comment at all in the two narrative lines' files

- Command: `git grep -c -E "<!--" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/en/core-concepts/verification-claim-integrity.md`
- Exit code: 1 (wrapper-confirmed). Stdout: no output.
- Reading: answers the Q24 start-of-work check. FAQ line 24 and verification-claim-integrity line 84 carry no marker of any HTML-comment form. Line 84 sits under the heading "Worked Example — Retention-Claim Hazard" (line 80), which the marker definition does not count (scope: the flagged line or the line above).

### S-05 — CLI families on the ten files (precondition, green)

- S-05a: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 74ad197380a0 -- <P10>` — exit 1 (wrapper-confirmed), no output.
- S-05b: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)" 74ad197380a0 -- <P10>` — exit 1 (wrapper-confirmed), no output.
- S-05c: `git grep -c -E "(^|[^/])moai (run|sync) SPEC-ID" 74ad197380a0 -- <P10>` — exit 1 (wrapper-confirmed), no output.

### S-06 — the retired CLI form `moai brain` (RED-now) and the verb surface

- S-06a: `git grep -c -E "(^|[^/])moai brain" 74ad197380a0 -- <P10>` — exit 0 — stdout (verbatim): `74ad197380a0:docs-site/content/en/core-concepts/verification-claim-integrity.md:1`
- S-06c: `<scratchpad>/moai-t1725 brain` (tree build, see §A) — exit 1 — output: `ERROR`, `Unknown command "brain" for "moai"`, `Did you mean this?`, `chain`.
- S-06d: `<scratchpad>/moai-t1725 --help` lists `plan`, `feedback`, `init`, `version`, `update`, `profile`, `spec`; `profile --help` lists `setup`; `spec --help` lists `audit`; `feedback --help` lists `scrub` and `queue`. Every inline-code CLI verb in the ten files other than `brain` therefore resolves (derived census: `moai init` 12, `version` 8, `update` 8, `profile setup` 8, `plan` 4, `feedback scrub` 4, `feedback queue` 4, `spec audit` 2, `brain` 1).

### S-07 — the other locales carry no stale form for the two English-only pages

- Command: `git grep -n -E "/moai (brain|relationship)|moai brain|/moai(:| )(design|issue|bug|fb|init)([^[:alnum:]-]|$)" 74ad197380a0 -- docs-site/content/ja/claude-code docs-site/content/ko/claude-code docs-site/content/zh/claude-code docs-site/content/ja/core-concepts docs-site/content/ko/core-concepts docs-site/content/zh/core-concepts`
- Exit code: 1 (wrapper-confirmed). Stdout: no output.
- Reading: the ja and zh foundations rows say "と /moai の関係" and "与 /moai 的关系"; they translate the same clause but are not closing-pattern matches and name no unresolved command. The ko row has no `/moai` clause. No edit is needed in those locales; Q17 is met because every locale that carries a stale form is edited.

### S-08 — router Priority 1 (supporting; resolution evidence)

- Command: `git grep -n -E "^- \*\*(plan|run|sync|project|feedback|fix|loop|mx|review|clean|codemaps|gate|e2e|harness|goal|todo|gtd)\*\*" 74ad197380a0 -- .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md`
- Exit code: 0. Stdout: 17 lines per file, identical in both files. The two relevant: `.claude/skills/moai/SKILL.md:69:- **project** (aliases: init): Project documentation generation` and `.claude/skills/moai/SKILL.md:70:- **feedback** (aliases: fb): GitHub issue creation`. No name or alias is `bug`, `issue`, `relationship`, `brain`, or `design`.
- Reading: a text read of a prompt-level router; no executable resolver exists (gap G-1).

### S-09 — registered slash commands (supporting)

- Command: `git ls-tree -r --name-only 74ad197380a0 -- .claude/commands/moai`
- Exit code: 0. Stdout: 17 files (`clean`, `codemaps`, `e2e`, `feedback`, `fix`, `gate`, `goal`, `gtd`, `harness`, `loop`, `mx`, `plan`, `project`, `review`, `run`, `sync`, `todo`); no `init`, no `fb`, no `moai.md`.

### S-10 — strict i18n check (regression-guard, green)

- Command: `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` (worktree root as working directory; the script writes only to a `mktemp` directory).
- Exit code: 0 (no failure displayed).
- Stdout (verbatim tail): `=== Summary ===`, `Errors:   0`, `Warnings: 0`, `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.` (153 `.md` files in each of ko, en, ja, zh).

### S-11 — replacement presence, feedback pages (RED-now; green path prints four rows)

- Command: `git grep -c -E "^> /moai:feedback$" 74ad197380a0 -- docs-site/content/en/utility-commands/moai-feedback.md docs-site/content/ja/utility-commands/moai-feedback.md docs-site/content/ko/utility-commands/moai-feedback.md docs-site/content/zh/utility-commands/moai-feedback.md`
- Exit code: 1 (wrapper-confirmed). Stdout: no output.
- Green path: milestone M1 adds the line in each locale; the same command with the run-phase HEAD prints four rows, each ending `:1`.

### S-12 — replacement presence, FAQ pages (RED-now; green path prints four rows)

- Command: `git grep -c -E "moai init.*/moai project|/moai project.*moai init" 74ad197380a0 -- docs-site/content/en/getting-started/faq.md docs-site/content/ja/getting-started/faq.md docs-site/content/ko/getting-started/faq.md docs-site/content/zh/getting-started/faq.md`
- Exit code: 1 (wrapper-confirmed). Stdout: no output.
- Green path: milestone M1 writes the replacement bullet in each locale; four rows, each ending `:1`.

### S-13 — bounded diff (run time)

- Command: `git diff --numstat BASELINE_SHA HEAD -- docs-site`
- Recorded at run time; limits are in AC-PRC-001f. Its state at pre-flight is empty output.

## E. Definition of done

- AC-PRC-001a, b, d, f, g pass at the run-phase HEAD with verbatim output recorded; AC-PRC-001c and e are re-observed green.
- The umbrella's closing check on these ten files: E-04 prints nothing for them; marked lines would be listed by file and line, and none exist at the defaults.
- Sync-phase and CI results for the pull request are recorded apart; they are not counted as passes here.

## F. Gaps (not observed)

- G-1 — Resolution through the router was read from `SKILL.md` text (S-08). The router is a prompt-level specification and no executable resolver exists, so "`/moai fb` resolves" is a read of the documented contract, not a run. A text read is a hypothesis about runtime behaviour.
- G-2 — The tree build's VCS stamp (`478285ab4c1f`, modified) does not equal the tree HEAD (`74ad197380a0`) in this linked worktree. The stamp was not used; provenance rests on the build procedure.
- G-3 — No existing slash-resolution or stale-citation checker exists in the tree: `scripts/` and `.github/workflows/` hold none, and the umbrella's E-28 records no cited-verbs producer (exit 1); the GATE producer is not yet written. The cells use `git grep` with the umbrella's families plus a complement pattern (S-02). The tool-based verdict is deferred to the GATE producer at the umbrella's M6.
- G-4 — The S-02 census used a `sort | uniq -c` pipeline (derived, not a cell).
- G-5 — The dispatch limited this card to three SPEC files, so `decision-index.md` (required when `interview.decision_gate` is `on`) and the SPEC-directory `progress.md` with the `§E.1` to `§E.4` skeleton were not authored. Decision rows are in spec.md § 6. Without a SPEC-directory progress.md, era classification applies rule H-1 (absent file: V2.x).
- G-6 — The card's source file `.moai/worktrees/t1655/.moai/reports/t1655/child-cards.md` is absent; the file and line list is verified only against the current tree.
- G-7 — Hugo was not built; the rendered output of the replacement lines was not observed.
- G-8 — The ja, ko, and zh replacement wording is native prose to be written in the run phase; its register was not reviewed here.
- G-9 — The operator record d-20261019T015908Z-220a was read with `moai decision read --scope card:t1725` (tree build); its body matches the coordinator's addendum.

## G. Residual risk

- The operator may read "explicit example marker" more widely than the marker defined here (for example, the "Worked Example" heading above line 84). Line 84 would then be exempt and its edit unneeded; the edit is harmless (decision QE-2).
- The umbrella closing pattern flags router aliases that resolve (QE-3). If the umbrella later carves them out, replacing `/moai fb` is no longer required for the check, though the replacement stays correct.
- The GATE checker may not honour the marker grammar; the interface is dormant at the defaults.
