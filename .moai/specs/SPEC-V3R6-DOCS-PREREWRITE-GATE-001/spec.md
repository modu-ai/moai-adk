---
id: SPEC-V3R6-DOCS-PREREWRITE-GATE-001
title: "Docs Pre-Rewrite Cleanup — CI gate child: cited-verbs check, blocking i18n parity, Hugo build job, link-check job (card t1717)"
version: "0.1.1"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".github/workflows, scripts, docs-site, .moai/guards"
lifecycle: spec-anchored
tags: "docs, docs-site, ci, gate, i18n, link-check, hugo, v3.2"
tier: M
related_specs: [SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001, SPEC-DOCS-SITE-001]
---

# SPEC-V3R6-DOCS-PREREWRITE-GATE-001 — Docs CI gate (child 1 of 11)

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-18 | 0.1.0 | manager-spec | Initial plan-phase authoring for factory card t1717 (Class C), the GATE child of the umbrella SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (split ruling `d-20261014T060405Z-0fcd`; routing `d-20261014T114834Z-e039`). It carries REQ-PRC-002, -012, -013, -014 and AC-PRC-002, -012, -013, -014 under their parent numbers, unrenumbered. REQ-PRC-012 is conditional on Q9 and REQ-PRC-013 on Q10, as the parent pointers state. Every measurement below was taken on tree `ae34f268319ce91f028ee692f25c1bd476f9e6e5` (card branch `WT-docs-ci-gate`, equal to local `main` at that moment) in this planning pass; nothing is carried over from the parent ledger except where a cell says so. **Audit defect B01 fixed:** the RED cell of AC-PRC-012 no longer cites E-16d (a green observation); it cites cells EG-05, EG-06, EG-07, and the supplemental simulation EG-08, which are red on this tree for the stated reason (acceptance §C). **Found while measuring, not in the card scope text:** (1) `internal/guardstate/census_test.go` fails when a workflow file has no `.moai/guards/liveness.yaml` entry, so the real file set is 10, or 11 when the link-check tool needs a config file (the card said 9 and 10); (2) a default-configured link checker is red on arrival (73,692 findings), so the link-check job needs a config and an explicit baseline policy (decision rows Q11, Q26); (3) the Q24 example-marker token is defined by the SWEEPE child and read by this child's checker (dependency, below). Status stays `draft`. |
| 2026-10-18 | 0.1.1 | manager-spec | Correction after plan-audit iteration 1 (FAIL 0.83 at `ec977eaf1`, `.moai/reports/t1717/plan-audit.md`; defect numbers follow it). **Fixed:** D1 (REQ-PRC-012 no longer promises a failure "whatever the strictness output says", which the planned edit could not produce; it now states what the job does: strict script on pull requests, failure on a non-zero exit, no continue-on-error; lane ruling LANE-RULED-1); D2 (AC-PRC-012 runs the real step bodies, new recipe R-12b, and the `echo skipped` mutant fails it); D3 (AC-PRC-002 grows from 5 to 13 test cases that discriminate C-1, C-2, C-6, and C-7, inside the existing three fixtures and one test); D4 (the ignore-list necessity check gets a per-entry sibling negative control, R-14e); D5 (Q25 binds the three new jobs only, and the parity job's trigger is stated as out of scope, LANE-RULED-5); D6 (exit codes recorded for EG-04, EG-11, EG-12, EG-13, EP-13; gap G-3 corrected). **Taken because the touched hunks allowed it:** D7 (Out of Scope names anchors and the 13 baseline links), D8 for AC-PRC-013 (parsed-step assertions), D9 (the Q9 negative branch), D10 (absent key counts as false in R-12), D11 (gap G-5). No requirement other than REQ-PRC-012 changed. Status stays `draft`. |

**Tier note.** Tier M: 10 to 11 files (inside the 5 to 15 band), an estimated 600 to 900 lines in total (an estimate, not a measurement), 4 requirements and 4 acceptance criteria (ceilings 16 and 16). No `design.md` or `research.md` is authored. The artifact set is spec.md, plan.md, acceptance.md, plus `decision-index.md` (decision gate is on in this repository) and `progress.md`.

## Overview (WHY)

The v3.2 documentation cleanup (parent card t1655, stage 5-0) only holds if CI stops it from regressing. Four gates are missing or advisory, and this child builds them. Measured state on tree `ae34f2683` (evidence ids refer to acceptance.md §C):

- The docs i18n parity check passes in strict mode on this tree and on `origin/main` (exit 0, 0 errors, 0 warnings; EP-01, EP-02), and the strict run does go red on a known failing input (a deleted `en` page: exit 1; EP-03). Its pull-request path is nevertheless advisory: the strictness step writes `strict=false` for `pull_request` (`docs-i18n-check.yml` lines 74 and 77; EG-05) and the `continue-on-error` expression at line 85 evaluates to true for it (EG-06; simulated, EG-08). The header comment still says "ADVISORY ONLY — NOT BLOCKING" (EG-07).
- No workflow builds the Hugo site and none checks links (EG-09, EG-10, EG-12, EG-13). The local Hugo build is green (EP-05) and the deployed pin `HUGO_VERSION` is `0.160.1` in `docs-site/vercel.json` line 11, equal to the local binary (EP-04). An unknown shortcode turns the local build red, exit 1 (EP-05).
- No producer for the cited-verbs figure exists in the tracked tree (EG-01, EG-03, EG-04). The verb surface it must read behaves as the parent recorded: an absent sub-verb prints the parent help and exits 0, while an absent top-level verb exits 1 (EP-10).
- A default-configured link checker is red on arrival: htmltest v0.17.0 reports 73,692 findings on the built site. With trailing-slash and empty-alt noise suppressed, 668 unresolved targets remain, of which 636 are the footer link `/changelog/` and 32 are relative links, and the site is only reachable through Vercel redirects for most of them (EP-06, EP-07). A job that is red before the work starts could never satisfy "passes on the closing tree" (verification-completeness §2, impossible direction), so the link-check job needs an explicit baseline policy (Q26).

## Scope (WHAT)

The files this child changes. Paths are repository-relative.

| # | Path | Change | Serves |
|---|------|--------|--------|
| 1 | `scripts/docs-cited-verbs.py` | new — the citation checker (one file) | REQ-PRC-002 |
| 2 | `scripts/test-docs-cited-verbs.sh` | new — the checker's test (one file) | REQ-PRC-002 |
| 3 | `scripts/docs-cited-verbs-testdata/red-absent-subverb.md` | new — fixture (a): cites an absent sub-verb | REQ-PRC-002 |
| 4 | `scripts/docs-cited-verbs-testdata/green-listed-subverb.md` | new — fixture (b): cites only resolvable forms | REQ-PRC-002 |
| 5 | `scripts/docs-cited-verbs-testdata/red-mermaid-label.md` | new — fixture (c): a Mermaid label citing `moai run SPEC-ID` | REQ-PRC-002 |
| 6 | `.github/workflows/docs-i18n-check.yml` | modified — pull-request strictness step (lines 76 to 77), `continue-on-error` expression (line 85), header comment (lines 3 to 16) | REQ-PRC-012 |
| 7 | `.github/workflows/docs-cited-verbs.yml` | new — job "Docs cited verbs" | REQ-PRC-002 |
| 8 | `.github/workflows/docs-hugo-build.yml` | new — job "Docs Hugo build" | REQ-PRC-013 |
| 9 | `.github/workflows/docs-link-check.yml` | new — job "Docs link check" | REQ-PRC-014 |
| 10 | `.moai/guards/liveness.yaml` | modified — one entry per new workflow (3 entries); required by `TestCensus_SetDifferenceEmptyBothDirections` | REQ-PRC-002, -013, -014 |
| 11 | `docs-site/.htmltest.yml` | new, conditional on decision Q11 naming htmltest — link-check tool configuration | REQ-PRC-014 |

Rows 1 to 9 are the card's nine files. Row 10 is added by this SPEC for the reason stated in the HISTORY row (a Go test, measured green on the baseline, EP-11, goes red the moment a workflow file lacks an entry). Row 11 is the card's "10 if the link-check tool needs a config" case; EP-06 shows the measured candidate needs one. No path under `internal/`, `cmd/`, `docs-site/content/`, `docs-site/vercel.json`, or `.github/required-checks.yml` is changed.

## Definitions

- **Blocking.** A job is blocking on a pull request when its conclusion is failure under the stated condition. Whether that failure also stops the merge depends on the job being a required status check of `main`, which is REQ-PRC-015, owned by the parent and operator-held (decisions Q12 and Q22). This child makes no branch-protection change.
- **Job names.** The check names this child publishes are exactly: "Validate 4-locale parity" (existing, unchanged), "Docs cited verbs", "Docs Hugo build", "Docs link check". The parent closes REQ-PRC-015 against these names, so a rename after landing breaks that closure (parent plan §D, job-name drift).
- **Pin.** Unless a cell states another commit, every tree-attributed observation is on `ae34f268319ce91f028ee692f25c1bd476f9e6e5`.

## Requirements (GEARS)

The GEARS text of each requirement keeps the parent's REQ number. The parent's pointer entries (spec.md § Requirements) name this child as the owner.

### CI gates (card item 7)

- **REQ-PRC-002** (Event-driven): **When** a pull request changes README or docs-site content and a cited verb, sub-verb, or slash command does not resolve under REQ-PRC-001, the docs cited-verbs check **shall** fail the run.
- **REQ-PRC-012** (Ubiquitous, conditional on decision Q9): **Where** decision Q9 records that pull requests are blocked on i18n drift, the docs i18n parity job **shall** run the parity script in strict mode on `pull_request` events, **shall** fail the run when that script exits non-zero, **shall not** continue on error for a `pull_request` run, and its header comment **shall not** describe the pull-request path as advisory.
- **REQ-PRC-013** (Ubiquitous, conditional on decision Q10): A Hugo build job **shall** run on pull requests, **shall** build the site with the Hugo version that decision Q10 records (default: the version named by `HUGO_VERSION` in `docs-site/vercel.json`), and **shall** fail the run when the build fails.
- **REQ-PRC-014** (Ubiquitous): A link-check job **shall** run on pull requests and **shall** fail the run when an internal link in the built docs-site is broken.

### Contract of the cited-verbs check (elaborates REQ-PRC-002; verified by AC-PRC-002)

These are normative statements about the check's observable interface. The parent's closing run (M6) consumes the final output line and the exit status, so they are fixed here.

- **C-1** (Ubiquitous): The cited-verbs check **shall** scan `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, and the Markdown files under `docs-site/content`, and **shall** read inline code spans and fenced code blocks only.
- **C-2** (Ubiquitous): The check **shall** derive the resolvable verb and sub-verb set at run time from the help output of a `moai` binary built from the tree under test, and the resolvable slash set from the file names under `.claude/commands/moai/`; it **shall** hold no hard-coded verb list.
- **C-3** (Ubiquitous): The check **shall** treat a sub-verb as resolved only when the parent verb's help lists it in a COMMANDS block or names it among the usage alternatives, because an absent sub-verb prints the parent help and exits 0.
- **C-4** (Ubiquitous): The check **shall** print one line per unresolved citation in the form `<path>:<line>:<cited form>`, followed by a final line `cited-verbs: <N>`.
- **C-5** (Ubiquitous): The check **shall** exit 0 when N is 0, exit 1 when N is above 0, and exit 2 when a required input cannot be read or when the scan set contains no file; an unreadable input and an empty scan set **shall not** produce exit 0, because a pass whose swept set is empty asserts nothing.
- **C-6** (Unwanted): The check **shall not** count command-line flags; the flag clause belongs to the FLAG child under Q18.
- **C-7** (Ubiquitous): The check **shall** carry each exemption as a named entry with a stated reason: the four `multi-llm/cg-mode.md` pages (the retained retirement notice) and the four `utility-commands/moai-todo.md` pages (natural-language input and documented error examples). A narrative-example exemption under Q24 **shall** be added only through the marker that the SWEEPE child defines.
- **C-8** (Unwanted): A run of the check **shall not** write any file inside the repository working tree.

## Acceptance

The Given-When-Then criteria AC-PRC-002, -012, -013, -014, their RED-now and green-path cells, and the evidence ledger are in `acceptance.md`.

## Out of Scope

Each item is named so that a later change cannot claim it silently. Child-level scope for the other ten children lives in their own SPECs.

### Out of Scope — branch protection and required checks

- Changing the GitHub branch protection of `main`, and any edit to `.github/required-checks.yml` (the repository's single source for required contexts). REQ-PRC-015 and decisions Q12 and Q22's closure stay with the parent; Q22 has owner GATE (leader-confirmed) and is carried as a decision row in this child, but this SPEC contains no step that depends on its answer.
- Reconciling the stale `docs-i18n-check` entry in the `auxiliary` list of `.github/required-checks.yml` with the job's real check name "Validate 4-locale parity". That entry is consumed by `internal/ciwatch` and belongs to the REQ-PRC-015 closure.

### Out of Scope — content and redirect fixes

- Editing any page under `docs-site/content/`, any README, or `docs-site/vercel.json`. The stale citations are fixed by the README and SWEEP children, the redirects by REDIR. This child only builds the gates that detect them.
- Fixing the 32 relative links and the footer link `/changelog/` that the link checker reports on the baseline; this child records them as baseline debts under Q26.
- Anchor (fragment) validity: REQ-PRC-014 gates internal page links only. The measured configuration switches anchor checking off, and the 20 anchor findings on the baseline are unclassified (acceptance.md G-8). The 13 baseline relative links that no redirect serves (EP-07) are ignored by the Q26 list and are not gated.

### Out of Scope — flags, Go code, and the push trigger

- Counting `--factory-run` and `--clear-policy` or any other flag (FLAG child, Q18), and any change under `internal/` or `cmd/`.
- Changing the `push` branch of the i18n strictness step (line 74): main-push behavior stays as it is; Q9 covers pull requests only.
- Changing the `on:` triggers of the parity workflow, including its `paths:` filter (`docs-i18n-check.yml` lines 22 to 26). AC-PRC-012 Then 5 freezes them. Decision Q25 binds the three new jobs only; whether the parity job should report on every pull request before it becomes a required check is a follow-up for the leader and the parent's REQ-PRC-015 closure (LANE-RULED-5 in decision-index.md).

### Out of Scope — pushing and remote observation

- Pushing, opening a pull request, or observing a remote CI run. New workflows cannot run on remote CI from this plan, so every acceptance criterion is verified locally; the first remote runs are recorded as regression-guard observations after the operator's batch.

## Dependencies and Parallel Work

- No `depends_on`. No file overlaps another child (rows 1 to 11 appear in no sibling file list). The citation checker is the input of the parent's closing run M6; it must exist, be tested, and print `cited-verbs: <N>` as the final line before M6 can record `cited-verbs: 0`.
- **Q5** (translate.mjs and migrate-frontmatter.sh classification) is an input row owned here; **Q18** (flag citations) is decided by FLAG and read here. Neither changes any output of the checker, because the checker scans neither `docs-site/scripts/` nor flags (C-1, C-6; decision-index.md).
- **Q24 marker.** The SWEEPE child defines the explicit example marker (operator verdict `d-20261014T060405Z-0fcd`). The checker reads that marker from one named constant; its value is recorded when SWEEPE's SPEC is committed. Until then the checker reports a marker-bearing line as unresolved (the conservative direction).
- **REDIR (Q21).** The link-check baseline policy (Q26) ignores the footer link `/changelog/`, which only resolves through the `/changelog` rule at `docs-site/vercel.json` line 118. If REDIR removes that rule, the footer link becomes a real 404 that the ignore entry would hide; REDIR and this child must agree on that entry (acceptance.md gap G-9).
- The parent's closing run M6 re-runs this child's committed criteria at this child's commit SHA and imports them by reference (parent plan §B.3).
