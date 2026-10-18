# SPEC-V3R6-DOCS-PREREWRITE-GATE-001 — Implementation Plan (revision 0.1.0)

Methodology: TDD (`.moai/config/sections/quality.yaml` `constitution.development_mode: tdd`). The checker is test-first: the test and fixtures land red before the checker exists. The workflow edits use the RED-now cells of acceptance.md §C as their red state and the recipes of acceptance.md §B.0 as their green check. Milestones are ordered by decision reversibility: the interface other work consumes first, the decisions most likely to change next, mechanical steps last. No time estimates are given.

## A. Context

- **Card and branch.** Factory card t1717 (Class C), branch `WT-docs-ci-gate`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1717`, base `ae34f268319ce91f028ee692f25c1bd476f9e6e5` (local `main`). Child 1 of 11 of SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001.
- **Artifacts.** spec.md, plan.md, acceptance.md (Tier M), plus decision-index.md and progress.md. The measurements behind every statement here are the ledger cells of acceptance.md §C (EG-nn, EP-nn).
- **File set (11, from spec.md § Scope).** New: `scripts/docs-cited-verbs.py`, `scripts/test-docs-cited-verbs.sh`, three fixtures under `scripts/docs-cited-verbs-testdata/`, `.github/workflows/docs-cited-verbs.yml`, `.github/workflows/docs-hugo-build.yml`, `.github/workflows/docs-link-check.yml`, and (if Q11 names htmltest) `docs-site/.htmltest.yml`. Modified: `.github/workflows/docs-i18n-check.yml`, `.moai/guards/liveness.yaml`.
- **Deviation from the card text, stated plainly.** The card said nine files, ten with a tool configuration. This plan has ten, eleven with the configuration, because `.moai/guards/liveness.yaml` must name every workflow file (`internal/guardstate/census_test.go`, EP-11). The addition is the registration of the three new workflows, and it is also the continued-firing answer that verification-completeness §1.3 requires of each new check. The leader confirms the count against the card before dispatch.
- **Decision rows.** Open and holding work: Q9 (M3), Q11 (M2), Q25 (trigger shape of M2, M4, M5). Applied by the Default rule: Q10 (M4), Q26 (M2). Inputs holding nothing: Q5, Q18, Q22.

## B. Known issues and pitfalls found while measuring

- **B-1 — the liveness census.** Adding a workflow file without a manifest entry turns `TestCensus_SetDifferenceEmptyBothDirections` red (EP-11). Register each new workflow in the same commit that adds it. Entry shape: five fields (`kind: github-workflow`, `locator`, `events`, `window`, `measure`); a path-gated pull-request check follows the parity workflow's precedent (`fired-with-effect`, a long window).
- **B-2 — a link checker is red on arrival.** The default htmltest run reports 73,692 findings; with the trailing-slash and empty-alt classes suppressed, 668 unresolved targets remain, and for most of them the live site resolves through Vercel redirects that a file-reading checker cannot see (EP-06, EP-07). The configuration therefore carries an explicit, commented ignore list (Q26), and AC-PRC-014 requires every entry to be necessary (R-14e).
- **B-3 — scratch Hugo builds.** A copy of `docs-site` outside the repository has no `.git`, and the site sets `enableGitInfo = true`; builds of copies use `HUGO_ENABLEGITINFO=false`. The CI jobs build inside the checkout and use the real configuration. The shallow-clone behavior on a runner is unmeasured (gap G-5), so the Hugo and link-check jobs check out with full history, as the Vercel deployment does (`VERCEL_DEEP_CLONE`).
- **B-4 — tool output location.** htmltest writes `tmp/` in its working directory, and neither `tmp/` nor `docs-site/tmp/` is ignored; the link-check job and R-14 run it from a directory outside the tree.
- **B-5 — sub-verb resolution.** `moai codex status` is valid but is not in the `codex` COMMANDS block; it appears among the usage alternatives (EP-10). The checker reads both. An exit-code probe cannot separate a real sub-verb from an absent one.
- **B-6 — the `auxiliary` entry.** `.github/required-checks.yml` lists `docs-i18n-check` as auxiliary, a name that matches no published check. This child does not edit that file (spec.md § Out of Scope); the parent reconciles it when it closes REQ-PRC-015.
- **B-7 — the Q24 marker.** SWEEPE defines the explicit example marker. The checker holds the marker in one named constant; M1 records the agreed token in `progress.md` §E.2 once SWEEPE's SPEC is committed. A marker-bearing line is reported until then.
- **B-8 — frontmatter and headings.** Canonical 12 fields in spec.md (`.claude/rules/moai/development/spec-frontmatter-schema.md`); `### Out of Scope — <topic>` H3 headings with bullets. The lint is run with `moai spec lint --strict` on this directory (recorded in progress.md).
- **B-9 — worktree guard.** From inside the card worktree, git is invoked as plain `git` or `git -C <own absolute path>`; compound forms and relative `-C` paths are refused.
- **B-10 — duplicated install steps.** The Hugo job and the link-check job both install Hugo and build the site. They stay as two independent workflow files with the same steps; a shared composite action would add a file outside the card's scope and is not warranted by two users.

## C. Pre-flight (before any edit)

Record these baselines in `progress.md` §E.2 so the red states are witnessed by the commit graph before the changes (verification-claim-integrity §2.3): the tree HEAD equals `ae34f2683` or a descendant that leaves the cells of acceptance.md §C unchanged (re-run EG-05 to EG-07 and EG-08 on the current HEAD and compare); `go test ./internal/guardstate/ -run '^(TestCensus_SetDifferenceEmptyBothDirections|TestCensus_EveryEntryCarriesFiveFields|TestCensus_ReleaseOnlySubjectsDeclareTheirCondition)$' -count=1 -v` exits 0 with three `--- PASS` lines; `actionlint` exits 0 on the existing `docs-i18n-check.yml`; `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` exits 0. If any differs, stop and report the drift instead of continuing.

## D. Constraints (do not violate)

- Touch only the eleven paths of spec.md § Scope. No path under `internal/`, `cmd/`, `docs-site/content/`, `docs-site/vercel.json`, `.github/required-checks.yml`, or any other child's list.
- No push, no pull request, no branch-protection change; the operator batches the push.
- The checker uses the Python standard library only, has no third-party import, and writes nothing in the repository (C-8). Scripts are committed executable (`100755`) like the other `scripts/*.sh`.
- The test reads the checker path from the environment variable `CHECKER` (default `scripts/docs-cited-verbs.py`) so the mutant probe R-02m can substitute stubs.
- Workflow files: `actions/checkout@v7` and `actions/setup-go@v7` (the versions the other workflows use), `permissions: contents: read`, `timeout-minutes` set, no `continue-on-error: true`, no `|| true`, no interpolation of event data into a script body.
- Job names are exactly `Docs cited verbs`, `Docs Hugo build`, `Docs link check`; the parity job keeps `Validate 4-locale parity`.
- The Hugo job reads the version from `docs-site/vercel.json` and verifies the downloaded archive against the release's published SHA-256 file.
- Match the existing style of each file touched; do not reformat unrelated lines of `docs-i18n-check.yml` or `liveness.yaml`.

## E. Self-verification deliverables (reported verbatim, each with command, output, and HEAD SHA)

1. AC-PRC-002 Then 1 to 7: the test output, the stub-mutant runs (R-02m), `yq` and `actionlint` output for `docs-cited-verbs.yml`, the census test, `git status --short`.
2. AC-PRC-012 Then 1 to 6: the R-12 matrix, the `has("if")` output, the `ADVISORY ONLY` grep, `actionlint`, and the strict-script red/green pair.
3. AC-PRC-013 Then 1 to 7: the `yq` and grep outputs, the R-13 pair, `actionlint`, the census test.
4. AC-PRC-014 Then 1 to 7: R-14 baseline and injected runs, R-14e per entry, `git status --short`, `actionlint`, the census test.
5. `moai spec lint --strict` on `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-GATE-001` stays clean after any edit to this directory.
6. The real-tree `cited-verbs: <N>` figure with the per-family comparison against the parent's E-01 to E-04 (recorded, not a criterion).
7. A blocker report, never a question to the user, for any needed decision this plan did not settle.

## F. Milestones (priority order; reversibility first)

### M1 — Priority High: the citation checker, its test, and its fixtures (REQ-PRC-002, AC-PRC-002 Then 1 to 4 and 7)

The interface is fixed here because the parent's closing run consumes it: output lines, final line, exit codes (contract C-1 to C-8).

1. Write the three fixtures and the test first, with the five cases and the `RESULT pass=N fail=M` line. Run the test and keep the red output (EG-02's successor: the test now exists and every case fails because the checker does not).
2. Write `scripts/docs-cited-verbs.py`. Approach: a line scanner that tracks fenced blocks (backtick and tilde fences) and extracts inline code spans; ANSI sequences stripped from help output; the verb set read from `<binary> --help` under every heading that ends in `COMMANDS`, `COMMANDS:`, or `TOOLS:`; the sub-verb set read from `<binary> <verb> --help` COMMANDS blocks plus the bracketed usage alternatives; the slash set from the file names in the commands directory; an unreadable binary, an unreadable scan path, or an empty scan set exits 2. Command-line options: `--moai`, `--commands-dir`, `--root`, and positional paths.
3. Make the five cases pass; run R-02m and confirm each stub fails the test.
4. Run the finished checker once on the real tree with a binary built from the tree; record the figure (deliverable 6).

### M2 — Priority High: the link-check job and its configuration (REQ-PRC-014, AC-PRC-014; gate Q11, Q26, Q25)

Decision-heavy and most likely to change. Do not start until Q11 carries an operator verdict.

1. Write the tool configuration (for htmltest: the file measured in EP-06 configuration B, `DirectoryPath: public`, comments on every ignore entry naming the baseline links and counts of EP-07). If Q11 names another tool, re-measure the three observations with that tool first.
2. Write `docs-link-check.yml`: checkout with full history, install Hugo as in M4, build, install the tool at a pinned version, run it from `$RUNNER_TEMP` against `docs-site/public`.
3. Run R-14 and R-14e; register the workflow in the liveness manifest in the same commit.

### M3 — Priority High: the i18n parity job blocks pull requests (REQ-PRC-012, AC-PRC-012; gate Q9)

Do not start until Q9 carries an operator verdict.

1. Keep the R-12 baseline output (EG-08) as the red record.
2. Edit `docs-i18n-check.yml`: the pull-request branch of the strictness step (lines 76 to 77) writes `strict=true`; the `continue-on-error` expression at line 85 no longer depends on the strictness output for a pull request; the header comment (lines 3 to 16) describes pull requests as blocking and push as still advisory, and the "Phase 2" paragraph is rewritten to match. The push branch (line 74) and the `workflow_dispatch` branch stay unchanged.
3. Run R-12 and compare with AC-PRC-012 Then 1 line by line; run the other Then checks.

### M4 — Priority Medium: the Hugo build job (REQ-PRC-013, AC-PRC-013; Q10 applied, gate Q25)

1. Write `docs-hugo-build.yml`: checkout with full history; read `HUGO_VERSION` from `docs-site/vercel.json`; download `hugo_extended_<version>_linux-amd64.tar.gz` and `hugo_<version>_checksums.txt` from the release, verify with `sha256sum`, unpack; run `hugo --minify --gc` in `docs-site`.
2. Run R-13 and the `yq` and grep checks; register the workflow in the liveness manifest in the same commit.

### M5 — Priority Medium: the cited-verbs job, the registration sweep, and the full verification (REQ-PRC-002, AC-PRC-002 Then 5 and 6; gate Q25)

1. Write `docs-cited-verbs.yml`: checkout, `actions/setup-go@v7` with `go-version-file: go.mod`, `go build -o "$RUNNER_TEMP/moai" ./cmd/moai`, the test, then the checker with `--moai "$RUNNER_TEMP/moai"`. Triggers per Q25.
2. Confirm all three manifest entries and run the census test.
3. Run the whole deliverable list of §E on the final tree and write the evidence into `progress.md` §E.2 and the lane's verdict file.

## G. Anti-patterns

- Declaring a criterion green from a file-existence cell alone; the behavioral Then clauses are the criterion.
- Tuning the ignore list until the baseline is green without R-14e: an entry that no baseline link needs is a hole in the gate.
- Treating the `FLAG`, `README`, `SWEEP`, or `REDIR` work as done or doing it here (spec.md § Out of Scope).
- Citing a scratch file as evidence; the deciding lines are copied into the verdict.
- Starting M2 or M3 on a Default or a card sentence instead of the operator's verdict.

## H. Cross-references

- Parent: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (index spec.md, plan.md §B.2 and §B.3 for M6 and the import rule, acceptance.md §C for the parent ledger).
- `.claude/rules/moai/development/verification-completeness.md` (§1.1 observed failure, §1.3 continued firing, §2 two-cell adoption, §2.1 RED-now cell form).
- `.claude/rules/moai/core/verification-claim-integrity.md` (§2.2 tool provenance, §2.3 ordering).
- `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- Sibling children that share a boundary: FLAG (Q18), SWEEPE (Q24 marker), REDIR (Q21 and the `/changelog` rule).

## I. Open questions (none blocks authoring; each has a stated default or an owner)

- **OQ-1.** The checker's trigger paths if Q25 chooses path-filtered triggers: README files, `docs-site/content/**`, the checker, test and fixtures, `.claude/commands/moai/**`, and the CLI sources (`cmd/**`, `internal/cli/**`) whose change can alter the verb set. A verb registered outside those directories would not retrigger the check (unmeasured).
- **OQ-2.** The concrete ignore patterns are the three of EP-06 configuration B; if REDIR changes the `/changelog` rule, the first pattern changes with it (gap G-9).
- **OQ-3.** `window` and `measure` values of the three manifest entries follow the parity workflow's precedent (`fired-with-effect`, a long window); the run phase confirms them against the entry's own trigger.
