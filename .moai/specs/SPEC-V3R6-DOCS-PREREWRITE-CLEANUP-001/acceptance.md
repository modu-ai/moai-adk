# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 (revision 0.2.2, umbrella)

## A. Pin and conventions

- **Document-level pin.** Tree commit `f35103480f03ad5a8418652bb03223dd67efa3dd`, tree object `ca8fc594f5f23c4691cf41c876a4b12191cab47d`. Every cell that reads the tree names the commit in its command (`git grep … f35103480f`, `git ls-tree … f35103480f`, `git cat-file -s f35103480f:<path>`, `git log … f35103480f`), so the pin is in the cell, not only in this header. Measured in revision 0.1.1 on 2026-10-14. The SPEC directory was untracked (as at f35103480f, revision 0.1.x); `git status --short` showed only that directory at every measurement of that revision. It has been tracked since d2932a0555.
- **Binary pin (verb-surface cells only).** `<scratchpad>/moai-head-make` is the tree build at f35103480f, built with the LDFLAGS of the `build` recipe (plan.md §A.2). Its version output carries the BuildID `moai_cp/20261007_195927-847-gf35103480f`. `<scratchpad>` means `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/b3ac8dde-8e32-43db-90ae-6e6be9148620/scratchpad`, a session-scoped, untracked directory.
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no `;`, `&&`, `|`, subshell, redirection, or `echo $?` wrapper. Its verbatim stdout and its exit code are separate fields of its ledger entry. The earlier wrapper exception (the `; echo "exit=$?"` form) is removed.
- **Exit codes.** The Bash tool displays a non-zero exit code. For a command with no output it displays no status, so a zero-match exit code is shown as `exit 1 (wrapper-confirmed)` only where the status was confirmed with a measurement wrapper run outside the cell. Otherwise the entry says `no output`.
- **Derived totals** (line sums with awk, or a Python parse of a pinned blob) are measurements outside the cell. They are labelled `derived`, and they are never the RED observation itself.
- **External cells** (E-20c, E-21) are labelled `external`, are not tree-attributed, and are never recorded as a pass (verification-completeness §2.1, undecidable disposition). They were observed in this revision before `2026-10-14T03:19:53Z`.
- **RED-now.** A cell is red when it reports a nonzero match count, a present forbidden path, or an absent required job.
- **Classes.** `release-blocking (gated)`: the closing check is re-executable on the tree and counts toward the completion verdict once its gate rows carry operator verdicts (§D). `regression-guard`: not counted toward the verdict; the reason is stated in §D.

Revisions 0.1.0 to 0.2.1 measured at f35103480f. Revision 0.2.2 re-measured the RED-now cells at 8816de4dae (the card branch after absorbing main 763193861b) and recorded their outputs verbatim. The supporting, verification, and protocol cells stay at f35103480f and are labelled; the drift table in acceptance section A lists the four items that differ at 8816de4dae. The whole-tree cells E-09b2 and E-28 exclude this SPEC's own directory by definition (leader ruling d-20261014T072226Z-9b67).

### Drift at 8816de4dae (revision 0.2.2)

| Item | At f35103480f | At 8816de4dae | Disposition |
|------|---------------|---------------|-------------|
| E-05b (flags in Go source) | 24 lines, 101 matching lines | 25 lines, 102 matching lines (one more file under internal/cli) | Cell stays at f35103480f. |
| E-27b (SPEC ID not tracked) | no output | 5 files, because the SPEC directory is tracked since d2932a0555 | Cell stays at f35103480f and means "before the SPEC was committed". |
| E-00 (HEAD and tree status) | HEAD `f35103480f03ad5a8418652bb03223dd67efa3dd`; tree `ca8fc594f5f23c4691cf41c876a4b12191cab47d`; `git status --short` shows `?? .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` | HEAD `8816de4dae`; the tracked tree is clean | Cell stays at f35103480f. |
| E-06 BuildID (binary cells E-06, E-07, E-07b, E-25a) | BuildID `moai_cp/20261007_195927-847-gf35103480f` carries the f35103480f build | not rebuilt at 8816de4dae; the binary still carries the f35103480f build | These cells stay at f35103480f until a rebuild at 8816de4dae. |

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Milestone | Gate (decision rows) | Class | Ledger entries | Owner child |
|----|-----|-----------|----------------------|-------|----------------|-------------|
| AC-PRC-001 | REQ-PRC-001 | M4; M1 producer | Q13, Q14 | release-blocking (gated) | E-01, E-02, E-03, E-04 (E-06, E-07 verification) | README (t1720); SWEEPA (t1721), SWEEPB (t1722), SWEEPC (t1723), SWEEPD (t1724), SWEEPE (t1725), SWEEPF (t1727), through partial criteria; closing: parent |
| AC-PRC-002 | REQ-PRC-002 | M1, M2, M6 | M1 inputs (Q5, Q18); Q12, Q22 if required | release-blocking (gated) | E-28 | GATE (t1717) |
| AC-PRC-003 | REQ-PRC-003 | M4 | Q8 | regression-guard | E-20a, E-20b, E-20c | README (t1720) |
| AC-PRC-004 | REQ-PRC-004 | M4 | Q1 | release-blocking (gated) | E-08, E-08b | README (t1720) |
| AC-PRC-005 | REQ-PRC-005 | M5 | Q2 | release-blocking (gated) | E-09a, E-09b, E-09b2, E-09c | RESID (t1719) |
| AC-PRC-006 | REQ-PRC-006 | M5 | Q3 | release-blocking (gated) | E-10 | RESID (t1719) |
| AC-PRC-007 | REQ-PRC-007 | M5 | Q4 | release-blocking (gated) | E-11a, E-11b, E-12, E-16b-skill, E-16b-hextra; precondition E-16c | RESID (t1719) |
| AC-PRC-008 | REQ-PRC-008 | M5 | Q6 | release-blocking (gated) | E-13 | RESID (t1719) |
| AC-PRC-009 | REQ-PRC-009 | M5 | Q7 | release-blocking (gated) | E-14a, E-14b | RESID (t1719) |
| AC-PRC-010 | REQ-PRC-010 | M3 | Q15, Q21 | release-blocking (gated) | E-15a (with its derived redirect count), E-15b, E-15d-design, E-15e | REDIR (t1718) |
| AC-PRC-011 | REQ-PRC-011 | M3 | Q16; Q23 (moai-mx pair) | release-blocking (gated) | E-15c, E-15d-changelog | REDIR (t1718) |
| AC-PRC-012 | REQ-PRC-012 | M2 | Q9 | release-blocking (gated) | E-16a, E-16d, E-21 | GATE (t1717) |
| AC-PRC-013 | REQ-PRC-013 | M2 | Q10 | release-blocking (gated) | E-17a, E-17b, E-19a, E-19b | GATE (t1717) |
| AC-PRC-014 | REQ-PRC-014 | M2 | Q11 | release-blocking (gated) | E-18a, E-18b | GATE (t1717) |
| AC-PRC-015 | REQ-PRC-015 | M6 (operator-held) | Q12, Q22 | regression-guard | E-21 | parent (operator-held) |

## B. Acceptance criteria

### AC-PRC-001 — stale CLI and slash references are gone (REQ-PRC-001)

Owner: README (SPEC-V3R6-DOCS-PREREWRITE-README-001) for the README files; SWEEPA (SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001), SWEEPB (SPEC-V3R6-DOCS-PREREWRITE-SWEEPB-001), SWEEPC (SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001), SWEEPD (SPEC-V3R6-DOCS-PREREWRITE-SWEEPD-001), SWEEPE (SPEC-V3R6-DOCS-PREREWRITE-SWEEPE-001), and SWEEPF (SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001) for their page files. Closing run: parent (M6).

- Given the tree at 8816de4dae, four family cells report matches for absent or unregistered forms: E-01 (CLI forms of the absent verbs `cg`, `gpt`, `audit`, `codemaps`, `model`; 60 lines in 48 files, with the retained retirement notice excluded), E-02 (absent sub-verbs `codex job|task|setup` and `todo rename|pick` and `spec list`; 24 lines in 4 files), E-03 (Mermaid labels citing `moai run SPEC-ID` and `moai sync SPEC-ID` as CLI verbs; 8 lines), and E-04 (slash forms the commands surface does not register; 45 matching lines in 22 files (one stdout line of the E-04 block in section C per file)). E-06 and E-07 are not measured at 8816de4dae: they show the absent forms are absent from the tree build at f35103480f, whose binary (BuildID carrying f35103480f) was not rebuilt at 8816de4dae; see the drift table in section A.
- When the closing tree is measured with the same four cells and the same exclusions, then each cell exits 1 with empty stdout.
- RED-now cells: E-01, E-02, E-03, E-04. Verification cells: E-06, E-07. The flags `--factory-run` and `--clear-policy` (E-05) are not cells of this criterion (verdict D3; decision Q18).
- Green path: M4 (content fixes per Q13 and Q14), with the M1 producer as the independent check.
- Gate: Q13, Q14; the M1 producer (its inputs Q5, Q18).
- Closure through partial criteria, one per child. A partial criterion means zero matches of that child's family pattern (E-01 to E-04) on that child's own file list, and no change outside that list. This parent AC closes when all partial criteria close and the closing run prints zero. The partial criteria are: README (t1720), on the four README files; SWEEPA (t1721), SWEEPB (t1722), SWEEPC (t1723), SWEEPD (t1724), SWEEPE (t1725), and SWEEPF (t1727), on their page files. FLAG (t1726) has no partial criterion here, because its flag clause is outside this AC and its own REQ and AC are new.
- Done when: E-01 to E-04 exit 1 with empty stdout on the closing tree, and the producer exits 0 on that tree.

### AC-PRC-002 — the cited-verbs producer exists, blocks pull requests, and was seen red (REQ-PRC-002)

Owner: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001)

- Given no producer exists in the tracked tree (E-28: no match for any cited-verbs name, and no workflow names the check).
- When M1 adds the producer and its fixture cases, then case (a) (a fixture citing `moai codex job`) fails, case (b) (a fixture citing `moai spec status`) passes, and case (c) (a fixture citing `moai run SPEC-ID` in a Mermaid label) fails.
- When M2 adds the pull-request step (path filter: README files and `docs-site/content/**`), then an injected stale token on a pull request fails the run.
- RED-now cell: E-28. The fixture red and the pull-request red are recorded at M1 and M2; they are not observable today.
- Green path: M1 (fixture red), M2 (pull-request red), M6 (exit 0 on the closing tree with `cited-verbs: 0`).
- Gate: the M1 producer's inputs (Q5, Q18); Q12 and Q22 if the step becomes a required context.
- Done when: the producer is tracked; the three fixture outcomes match plan.md §B M1; the pull-request run is red on the injected token; the closing run exits 0 and prints `cited-verbs: 0`.

### AC-PRC-003 — the badge and hugo.toml name the Q8 release (REQ-PRC-003) — regression-guard

Owner: README (SPEC-V3R6-DOCS-PREREWRITE-README-001)

- Given the four README Release badges and the two `hugo.toml` lines name v3.1.3 (E-20a, E-20b), and the public release list marks v3.1.2 Latest and v3.1.3 Draft (E-20c, external).
- When Q8 is applied, then the tree cells match the Q8 verdict, and `version` and `releaseDate` change in the same commit.
- RED-now cells: E-20a, E-20b (tree, pinned). E-20c is external and is not tree-attributed.
- Class: regression-guard. Reason: the closing comparison needs the external Latest release, which the repository cannot observe (verification-completeness §2.1). The external observation is recorded with its time, not counted as a pass.
- Gate: Q8.
- Done when (guard): at closing, the tree cells match the Q8 verdict, and the external observation is recorded with its UTC time.

### AC-PRC-004 — the CG retirement passage lives on one Korean page (REQ-PRC-004)

Owner: README (SPEC-V3R6-DOCS-PREREWRITE-README-001)

- Given nine Korean pages carry the exact retirement sentence (E-08), and seven of the eight pages other than the canonical page already link to the CG page (E-08b). `what-is-moai-adk.md` carries the sentence and no link.
- When the duplicates are replaced by link sentences, then the sentence appears only on the page that Q1 names.
- RED-now cells: E-08 (9 files; 8 outside the canonical page) and E-08b (the link baseline).
- Proxy caveat: the card's "쪽" (page) measure has no repository definition (gap G7). This criterion counts copies of the exact sentence as the proxy.
- Gate: Q1.
- Done when: `git grep -c -F "설정 이전 안내와 함께 종료합니다" <closing> -- docs-site/content/ko` lists exactly one file (the Q1 page), and each of the eight other pages carries a link to it (the E-08b pattern, count ≥ 1 for each).

### AC-PRC-005 — unreferenced images are removed (REQ-PRC-005)

Owner: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001)

- Given five tracked image files (E-09a) with no reference in `README*.md`, `docs-site`, `assets`, `.moai/docs`, `.claude/skills`, or `internal` (E-09b, scoped). Their sizes total 5,751,702 bytes (E-09c, derived).
- Given one tracked historical report names them (E-09b2: `.moai/reports/historical/gpt-final-design-20260914-b40ff9f4/inventory.json`). It is a historical artifact, not a display reference, and it is outside the scoped paths above. This is the correction of the verdict's D16 false universal.
- When Q2 is applied as deletion, then `git ls-tree` on the closing tree returns none of the five paths.
- RED-now cells: E-09a (five paths tracked), E-09b (zero scoped references), E-09b2 (the named historical reference), E-09c (sizes).
- Gate: Q2.
- Done when: `git ls-tree -r --name-only <closing> -- <the five paths>` is empty, and the E-09b pattern exits 1 with no output on the closing tree.

### AC-PRC-006 — the OG image is re-encoded under the Q3 budget (REQ-PRC-006)

Owner: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001)

- Given `assets/images/moai-adk-og.png` is 7,837,648 bytes (E-10) and is referenced by the four README files (E-10 refs).
- When Q3 records a budget and the file is re-encoded to it, then its blob size is at or below that budget.
- RED-now cell: E-10. The budget is not set in this SPEC; it is recorded here when Q3 has an operator verdict.
- Gate: Q3 (EVIDENCE-NEEDED).
- Done when: `git cat-file -s <closing>:assets/images/moai-adk-og.png` is at or below the Q3 verdict value, and the four README references still resolve.

### AC-PRC-007 — Nextra-era scripts and their narrative are removed or corrected (REQ-PRC-007)

Owner: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001)

- Given two Nextra-named scripts are tracked (E-11a, E-11b: `docs-site/scripts/fix-mdx-formatting.js`, `scripts/convert-nextra-to-hextra/main.go`) and no workflow, Makefile, lefthook entry, or build configuration invokes them (E-12).
- Given the oss-docs skills describe `docs-i18n-check.sh` as absent (E-16b-skill: 3 lines, including the sentence split across `hns-oss-docs-i18n-rules/SKILL.md` lines 115–116), and the i18n rule document names Hextra as the current toolchain (E-16b-hextra: `Markdown/Hextra` on line 35, `Hextra/Goldmark` on line 37).
- When Q4 is applied as deletion and the statements are corrected, then the two scripts are untracked, and the statements say that `scripts/docs-i18n-check.sh` exists (tracked, mode 100755, E-16c) and that gen_menu.py does not.
- RED-now cells: E-11a, E-11b, E-12, E-16b-skill, E-16b-hextra. E-16c is a precondition (green), not a RED cell.
- Gate: Q4. Q5 does not gate this criterion: `translate.mjs` and `migrate-frontmatter.sh` are not among its paths.
- Done when: `git ls-tree -r --name-only <closing> -- docs-site/scripts/fix-mdx-formatting.js scripts/convert-nextra-to-hextra` is empty; the E-16b-skill pattern exits 1 with no output on the closing tree; and `git grep -c -F "Markdown/Hextra" <closing> -- .moai/docs/docs-site-i18n-rules.md` and `git grep -c -F "Hextra/Goldmark" <closing> -- .moai/docs/docs-site-i18n-rules.md` each exit 1 with no output.

### AC-PRC-008 — the Python-era docs-site README is removed or replaced (REQ-PRC-008)

Owner: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001)

- Given `docs-site/README.md` is tracked at 42,512 bytes (E-13).
- When Q6 is applied, then the path is absent (deletion), or the replacement carries the Q6 alternate with a current description.
- RED-now cell: E-13.
- Gate: Q6.
- Done when: `git cat-file -s <closing>:docs-site/README.md` fails (path absent), or the replacement passes the closing check that it names no Python-era toolchain.

### AC-PRC-009 — llms.txt is removed or regenerated without EARS wording (REQ-PRC-009)

Owner: RESID (SPEC-V3R6-DOCS-PREREWRITE-RESID-001)

- Given `docs-site/static/llms.txt` was last changed 2026-04-20T06:15:43Z (E-14a) and carries one EARS-style line (E-14b).
- When Q7 is applied, then the file is absent, or its content contains no EARS wording.
- RED-now cells: E-14a, E-14b.
- Gate: Q7 (product-level, operator-held).
- Done when: the file is absent, or `git grep -c -i -E "ears|shall|THEN" <closing> -- docs-site/static/llms.txt` exits 1 with no output.

### AC-PRC-010 — every redirect destination resolves (REQ-PRC-010)

Owner: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001)

- Given `docs-site/vercel.json` has 42 `"source"` keys (E-15a; derived in E-15a: 39 redirect objects, one rewrite, two header rules). Three redirects point at `design` destinations (E-15b), and no `design` content path exists in any locale (E-15d-design).
- Given the two changelog rules point at an external destination (`https://github.com/modu-ai/moai-adk/releases`, lines 113–121). REQ-PRC-010 is not satisfied by them until decision Q21 is answered; this SPEC does not amend REQ-PRC-010.
- When Q15 and Q21 are applied, then no destination names `design`, and each remaining destination maps to an existing same-locale page in all four locales, except external destinations that Q21 admits.
- RED-now cells: E-15b, E-15d-design. Verification: E-15e (the ko destinations of the non-design rules resolve).
- Gate: Q15, Q21.
- Done when: `git grep -n -E "\"destination\": \"[^\"]*design" <closing> -- docs-site/vercel.json` exits 1 with no output, and each remaining destination has a content file in each locale named by its rule (one cell per destination path; a ko-only check is not sufficient, gap G11).

### AC-PRC-011 — no redirect shadows an existing page (REQ-PRC-011)

Owner: REDIR (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001)

- Given the changelog redirect sources at `vercel.json` lines 113 and 118 (E-15c), and four `changelog/_index.md` pages under the same URLs (E-15d-changelog).
- When Q16 is applied, then no redirect source matches an existing content page.
- RED-now cells: E-15c, E-15d-changelog.
- Gate: Q16 (its interaction with Q21 is recorded in decision-index.md).
- Extension under Q23 (leader ruling B04, `d-20261014T060405Z-0fcd`): the REDIR child (t1718) extends this criterion to the moai-mx redirect pair at `docs-site/vercel.json` lines 163 and 168. Those rules send `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` to `/:locale/utility-commands`, and `/utility-commands/moai-mx` to `/ko/utility-commands`, so they shadow the four existing `utility-commands/moai-mx.md` pages. Q23 decides between removing the two rules and amending REQ-PRC-011 to exempt the pair. The REDIR child writes the pair's done check once Q23 has a verdict.
- Done when: `git ls-tree -r --name-only <closing> -- docs-site/content/en/changelog docs-site/content/ja/changelog docs-site/content/ko/changelog docs-site/content/zh/changelog` is empty, or the redirect sources are removed under Q16's alternate.

### AC-PRC-012 — the i18n job blocks pull requests (REQ-PRC-012)

Owner: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001)

- Given the strict run passes on this tree (E-16d: exit 0, 0 errors, 0 warnings). The pull-request path of the job is advisory: its strictness step writes `strict=false` (`docs-i18n-check.yml:77`), the `continue-on-error` expression at line 85 evaluates to true for it (E-16a), and the job is not a required context (E-21).
- When Q9 is applied and M2 edits the pull-request strictness step (lines 76–77) and the expression at line 85, as plan.md §B M2 states, then an injected drift fails the job on a pull request run.
- RED-now cells: E-16a (line 85), E-16d (strict exit 0, so the drift must be injected to observe red), E-21 (not a required context).
- Gate: Q9.
- Done when: a pull-request run with an injected drift is red, and the closing-tree run is green. The criterion closes only after the pull-request observation, and the closing record names that run.

### AC-PRC-013 — a Hugo build job runs on pull requests (REQ-PRC-013)

Owner: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001)

- Given no workflow contains a Hugo build step (E-19a); the Vercel pin is 0.160.1 (E-17b); the local Hugo v0.160.1 builds the site with exit 0 (E-17a, E-19b).
- When M2 adds the job at the Q10 version, then it runs Hugo on pull requests and fails on an injected template error.
- RED-now cells: E-19a (absent), E-17b (the pin), E-19b (green local build, so the content does not block the job).
- Gate: Q10.
- Done when: a `pull_request` job runs the Q10 Hugo version and fails on the named fixture (plan.md §E).

### AC-PRC-014 — a link-check job blocks broken internal links (REQ-PRC-014)

Owner: GATE (SPEC-V3R6-DOCS-PREREWRITE-GATE-001)

- Given no workflow contains a link checker or a link-check job (E-18a), and none of lychee, linkinator, htmltest, or markdown-link-check is installed locally (E-18b; `hugo` is installed).
- When Q11 is applied and M2 adds the job, then it fails on an injected broken internal link and passes on the closing tree.
- RED-now cells: E-18a, E-18b.
- Gate: Q11.
- Done when: a `pull_request` job runs the chosen tool, fails on the named fixture, and passes on the closing tree.

### AC-PRC-015 — the three jobs are required status checks (REQ-PRC-015) — regression-guard

Owner: parent (SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001), operator-held (Q12, Q22)

- Given the five required contexts of `main` do not include the three jobs (E-21, external), and enforce_admins is enabled (E-21).
- When the operator adds the job names as required contexts (Q12, with Q22 deciding the requirement's place), then the external observation lists them.
- RED-now cell (external): E-21.
- Class: regression-guard (operator-held, Q22). Reason: branch protection is external and operator-held (keep-set); the repository cannot observe it (verification-completeness §2.1). The observation is recorded with its time, not counted as a pass, and the Out of Scope exclusion conflicts with the requirement until Q22 is answered.
- Gate: Q12, Q22.
- Done when (guard): the external observation at closing lists the job names exactly as the workflow job names, recorded with its UTC time.

## C. Evidence ledger (verbatim)

Each entry gives the command (one invocation, pinned), the exit code as its own field, and the stdout verbatim in a fenced block. Entries marked `derived` are measurements outside the cell.

### E-00 — pins and tree status

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: changed; see the drift table in section A.

- `git -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655 rev-parse --show-toplevel` — exit 0 — stdout: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655`
- `git -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655 rev-parse HEAD` — exit 0 — stdout: `f35103480f03ad5a8418652bb03223dd67efa3dd`
- `git -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655 rev-parse HEAD^{tree}` — exit 0 — stdout: `ca8fc594f5f23c4691cf41c876a4b12191cab47d`
- `git -C /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655 status --short` — exit 0 — stdout: `?? .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (the untracked SPEC directory; no tracked file is modified)

### E-01 — family F1: CLI forms of the absent verbs (RED-now; notice excluded)

- Command: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 8816de4dae -- README.md README.ko.md README.ja.md README.zh.md docs-site/content ':!docs-site/content/en/multi-llm/cg-mode.md' ':!docs-site/content/ja/multi-llm/cg-mode.md' ':!docs-site/content/ko/multi-llm/cg-mode.md' ':!docs-site/content/zh/multi-llm/cg-mode.md'`
- Exit code: 0.
- Stdout (verbatim; 48 lines; derived: 48 files, 60 matching lines):

```
8816de4dae:README.ja.md:4
8816de4dae:README.ko.md:4
8816de4dae:README.md:4
8816de4dae:README.zh.md:4
8816de4dae:docs-site/content/en/advanced/bas-navigator.md:1
8816de4dae:docs-site/content/en/advanced/factory-mode.md:1
8816de4dae:docs-site/content/en/advanced/profile-matrix.md:1
8816de4dae:docs-site/content/en/advanced/tokenomics-overview.md:1
8816de4dae:docs-site/content/en/core-concepts/what-is-moai-adk.md:1
8816de4dae:docs-site/content/en/cost-optimization/prompt-caching.md:1
8816de4dae:docs-site/content/en/getting-started/cli.md:1
8816de4dae:docs-site/content/en/getting-started/windows-guide.md:1
8816de4dae:docs-site/content/en/guides/mcp-server.md:1
8816de4dae:docs-site/content/en/multi-llm/_index.md:1
8816de4dae:docs-site/content/en/multi-llm/model-policy.md:1
8816de4dae:docs-site/content/ja/advanced/bas-navigator.md:1
8816de4dae:docs-site/content/ja/advanced/factory-mode.md:1
8816de4dae:docs-site/content/ja/advanced/profile-matrix.md:1
8816de4dae:docs-site/content/ja/advanced/tokenomics-overview.md:1
8816de4dae:docs-site/content/ja/core-concepts/what-is-moai-adk.md:1
8816de4dae:docs-site/content/ja/cost-optimization/prompt-caching.md:1
8816de4dae:docs-site/content/ja/getting-started/cli.md:1
8816de4dae:docs-site/content/ja/getting-started/windows-guide.md:1
8816de4dae:docs-site/content/ja/guides/mcp-server.md:1
8816de4dae:docs-site/content/ja/multi-llm/_index.md:1
8816de4dae:docs-site/content/ja/multi-llm/model-policy.md:1
8816de4dae:docs-site/content/ko/advanced/bas-navigator.md:1
8816de4dae:docs-site/content/ko/advanced/factory-mode.md:1
8816de4dae:docs-site/content/ko/advanced/profile-matrix.md:1
8816de4dae:docs-site/content/ko/advanced/tokenomics-overview.md:1
8816de4dae:docs-site/content/ko/core-concepts/what-is-moai-adk.md:1
8816de4dae:docs-site/content/ko/cost-optimization/prompt-caching.md:1
8816de4dae:docs-site/content/ko/getting-started/cli.md:1
8816de4dae:docs-site/content/ko/getting-started/windows-guide.md:1
8816de4dae:docs-site/content/ko/guides/mcp-server.md:1
8816de4dae:docs-site/content/ko/multi-llm/_index.md:1
8816de4dae:docs-site/content/ko/multi-llm/model-policy.md:1
8816de4dae:docs-site/content/zh/advanced/bas-navigator.md:1
8816de4dae:docs-site/content/zh/advanced/factory-mode.md:1
8816de4dae:docs-site/content/zh/advanced/profile-matrix.md:1
8816de4dae:docs-site/content/zh/advanced/tokenomics-overview.md:1
8816de4dae:docs-site/content/zh/core-concepts/what-is-moai-adk.md:1
8816de4dae:docs-site/content/zh/cost-optimization/prompt-caching.md:1
8816de4dae:docs-site/content/zh/getting-started/cli.md:1
8816de4dae:docs-site/content/zh/getting-started/windows-guide.md:1
8816de4dae:docs-site/content/zh/guides/mcp-server.md:1
8816de4dae:docs-site/content/zh/multi-llm/_index.md:1
8816de4dae:docs-site/content/zh/multi-llm/model-policy.md:1
```

- Reading: the `moai-codemaps` pages contribute no line here, because their 80 matches are the registered slash command (E-01s). Before the notice exclusion the same command gives 52 files and 64 lines; the four excluded lines are E-01a.

### E-01a — the retained retirement notice (exclusion evidence)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -n -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" f35103480f -- docs-site/content/en/multi-llm/cg-mode.md docs-site/content/ja/multi-llm/cg-mode.md docs-site/content/ko/multi-llm/cg-mode.md docs-site/content/zh/multi-llm/cg-mode.md`
- Exit code: 0.
- Stdout (verbatim; one line per page, all at line 8):

```
f35103480f:docs-site/content/en/multi-llm/cg-mode.md:8:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.
f35103480f:docs-site/content/ja/multi-llm/cg-mode.md:8:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。
f35103480f:docs-site/content/ko/multi-llm/cg-mode.md:8:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다.
f35103480f:docs-site/content/zh/multi-llm/cg-mode.md:8:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。
```

### E-01s — slash forms of the family (verification; derived)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -c -E "/moai codemaps([^[:alnum:]-]|$)" f35103480f -- README.md README.ko.md README.ja.md README.zh.md docs-site/content` — derived: 44 files, 152 matching lines. Every slash match of the E-01 family is this registered command (`.claude/commands/moai/codemaps.md`, E-26b). The E-01 family's slash lines therefore resolve.
- Command: `git grep -n -E "/moai (cg|gpt|audit|model)([^[:alnum:]-]|$)" f35103480f -- README.md README.ko.md README.ja.md README.zh.md docs-site/content` — exit 1 (wrapper-confirmed) — stdout: no output. No slash form of an absent name exists.

### E-02 — family F2: CLI forms of absent sub-verbs (RED-now; `web console` and todo examples excluded)

- Command: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)" 8816de4dae -- README.md README.ko.md README.ja.md README.zh.md docs-site/content ':!docs-site/content/en/utility-commands/moai-todo.md' ':!docs-site/content/ja/utility-commands/moai-todo.md' ':!docs-site/content/ko/utility-commands/moai-todo.md' ':!docs-site/content/zh/utility-commands/moai-todo.md'`
- Exit code: 0.
- Stdout (verbatim; 4 lines; derived: 24 matching lines):

```
8816de4dae:docs-site/content/en/guides/mcp-server.md:6
8816de4dae:docs-site/content/ja/guides/mcp-server.md:6
8816de4dae:docs-site/content/ko/guides/mcp-server.md:6
8816de4dae:docs-site/content/zh/guides/mcp-server.md:6
```

- Correction (N1, found in re-measurement). The 0.1.0 family `codex (job|task|setup)|web console|todo (rename|pick)|spec list` gave 42 lines, which are 24 CLI lines (above), 8 lines in the moai-todo pages (E-02t), and 10 `web console` lines that name a UI, not a command: `README.md:116`, `README.md:428`, `README.md:431`, `docs-site/content/en/_index.md:34`, `docs-site/content/en/_index.md:36`, `docs-site/content/en/advanced/config-sections.md:215`, `docs-site/content/en/advanced/config-sections.md:269`, `docs-site/content/en/cli-reference/epic.md:107`, `docs-site/content/en/guides/mcp-server.md:293`, `docs-site/content/en/utility-commands/moai-feedback.md:103`. No CLI form of `web console` exists: `moai web --help` lists flags only (E-07). The alternative is dropped.

### E-02t — the moai-todo examples (exemption evidence)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)" f35103480f -- docs-site/content/en/utility-commands/moai-todo.md docs-site/content/ja/utility-commands/moai-todo.md docs-site/content/ko/utility-commands/moai-todo.md docs-site/content/zh/utility-commands/moai-todo.md`
- Exit code: 0.
- Stdout (verbatim):

```
f35103480f:docs-site/content/en/utility-commands/moai-todo.md:2
f35103480f:docs-site/content/ja/utility-commands/moai-todo.md:2
f35103480f:docs-site/content/ko/utility-commands/moai-todo.md:2
f35103480f:docs-site/content/zh/utility-commands/moai-todo.md:2
```

- Reading: the two matched lines per page are a natural-language input (`$ moai todo rename the stale hint`, which the page and the `moai todo --help` text both describe as an add, line 175 of the en page) and a documented error example (`moai todo pick t151` in the grammar table, line 231). Neither is a CLI citation. `moai todo --help` states that `moai todo <two or more words>` adds the card, and that `moai todo pick t151` is refused as a mistyped verb.

### E-03 — family F3: Mermaid labels citing `moai run` and `moai sync` (RED-now; CLI form)

- Command: `git grep -n -E "(^|[^/])moai (run|sync) SPEC-ID" 8816de4dae -- docs-site/content`
- Exit code: 0.
- Stdout (verbatim; 8 lines; the value is 8, not 12):

```
8816de4dae:docs-site/content/en/worktree/_index.md:91:        B2 --> B3["moai run SPEC-ID"]
8816de4dae:docs-site/content/en/worktree/_index.md:92:        B3 --> B4["moai sync SPEC-ID"]
8816de4dae:docs-site/content/ja/worktree/_index.md:87:        B2 --> B3["moai run SPEC-ID"]
8816de4dae:docs-site/content/ja/worktree/_index.md:88:        B3 --> B4["moai sync SPEC-ID"]
8816de4dae:docs-site/content/ko/worktree/_index.md:119:        B2 --> B3["moai run SPEC-ID"]
8816de4dae:docs-site/content/ko/worktree/_index.md:120:        B3 --> B4["moai sync SPEC-ID"]
8816de4dae:docs-site/content/zh/worktree/_index.md:96:        B2 --> B3["moai run SPEC-ID"]
8816de4dae:docs-site/content/zh/worktree/_index.md:97:        B3 --> B4["moai sync SPEC-ID"]
```

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command (slash form; not in the family; verification): `git grep -n -E "/moai (run|sync) SPEC-ID" f35103480f -- docs-site/content` — exit 0 — stdout (verbatim; 4 lines):

```
f35103480f:docs-site/content/en/worktree/faq.md:123:    E --> F["/moai run SPEC-ID"]
f35103480f:docs-site/content/ja/worktree/faq.md:121:    E --> F["/moai run SPEC-ID"]
f35103480f:docs-site/content/ko/worktree/faq.md:130:    E --> F["/moai run SPEC-ID"]
f35103480f:docs-site/content/zh/worktree/faq.md:121:    E --> F["/moai run SPEC-ID"]
```

### E-04 — family F4: slash forms the commands surface does not register (RED-now)

- Command: `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 8816de4dae -- README.md README.ko.md README.ja.md README.zh.md docs-site/content`
- Exit code: 0.
- Stdout (verbatim; 22 lines; derived: 45 matching lines):

```
8816de4dae:docs-site/content/en/claude-code/foundations/_index.md:1
8816de4dae:docs-site/content/en/core-concepts/verification-claim-integrity.md:1
8816de4dae:docs-site/content/en/getting-started/faq.md:1
8816de4dae:docs-site/content/en/utility-commands/moai-feedback.md:3
8816de4dae:docs-site/content/en/workflow-commands/_index.md:1
8816de4dae:docs-site/content/en/workflow-commands/_meta.yaml:1
8816de4dae:docs-site/content/en/workflow-commands/moai-design.md:5
8816de4dae:docs-site/content/ja/getting-started/faq.md:1
8816de4dae:docs-site/content/ja/utility-commands/moai-feedback.md:3
8816de4dae:docs-site/content/ja/workflow-commands/_index.md:1
8816de4dae:docs-site/content/ja/workflow-commands/_meta.yaml:1
8816de4dae:docs-site/content/ja/workflow-commands/moai-design.md:4
8816de4dae:docs-site/content/ko/getting-started/faq.md:1
8816de4dae:docs-site/content/ko/utility-commands/moai-feedback.md:3
8816de4dae:docs-site/content/ko/workflow-commands/_index.md:2
8816de4dae:docs-site/content/ko/workflow-commands/_meta.yaml:1
8816de4dae:docs-site/content/ko/workflow-commands/moai-design.md:5
8816de4dae:docs-site/content/zh/getting-started/faq.md:1
8816de4dae:docs-site/content/zh/utility-commands/moai-feedback.md:3
8816de4dae:docs-site/content/zh/workflow-commands/_index.md:1
8816de4dae:docs-site/content/zh/workflow-commands/_meta.yaml:1
8816de4dae:docs-site/content/zh/workflow-commands/moai-design.md:4
```

- Reading (N2, open). The family also counts narrative and error examples: `docs-site/content/en/getting-started/faq.md:24` (`/moai init` described as a misconception), `docs-site/content/en/core-concepts/verification-claim-integrity.md:84` (a retired `/moai brain` narrative; one line, text not reproduced here), and the invocation list at `docs-site/content/en/utility-commands/moai-feedback.md:32–34` (`/moai fb`, `/moai bug`, `/moai issue`, which are unregistered). An exemption decision is needed before the producer counts the first two. The E-04 pattern is unchanged in this revision.

### E-05 — family F5: the two flags in docs (informational; not a cell of AC-PRC-001)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -c -F -e "--factory-run" -e "--clear-policy" f35103480f -- README.md README.ko.md README.ja.md README.zh.md docs-site/content`
- Exit code: 0.
- Stdout (verbatim; 12 lines; derived: 36 matching lines):

```
f35103480f:README.ja.md:3
f35103480f:README.ko.md:3
f35103480f:README.md:3
f35103480f:README.zh.md:3
f35103480f:docs-site/content/en/advanced/factory-mode.md:4
f35103480f:docs-site/content/en/cli-reference/launchers.md:2
f35103480f:docs-site/content/ja/advanced/factory-mode.md:4
f35103480f:docs-site/content/ja/cli-reference/launchers.md:2
f35103480f:docs-site/content/ko/advanced/factory-mode.md:4
f35103480f:docs-site/content/ko/cli-reference/launchers.md:2
f35103480f:docs-site/content/zh/advanced/factory-mode.md:4
f35103480f:docs-site/content/zh/cli-reference/launchers.md:2
```

- Reading: the flags are real parsed flags, not stale citations (verdict D3). They are not in AC-PRC-001. Their disposition is decision Q18.

### E-05b (= E-25b) — the flags are defined and parsed in the Go source

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: changed; see the drift table in section A.

- Command: `git grep -c -F -e "factory-run" -e "clear-policy" f35103480f -- internal cmd`
- Exit code: 0.
- Stdout (verbatim; 24 lines; derived: 101 matching lines):

```
f35103480f:internal/cli/codex_factory.go:9
f35103480f:internal/cli/codex_factory_retire_test.go:2
f35103480f:internal/cli/codex_launcher.go:2
f35103480f:internal/cli/factory.go:16
f35103480f:internal/cli/factory_ambiguous_join_test.go:13
f35103480f:internal/cli/factory_autodispatch_test.go:2
f35103480f:internal/cli/factory_card.go:1
f35103480f:internal/cli/factory_join_discovery_test.go:9
f35103480f:internal/cli/factory_launch_helpers.go:2
f35103480f:internal/cli/factory_m5_test.go:2
f35103480f:internal/cli/factory_m6_test.go:9
f35103480f:internal/cli/factory_mixed_test.go:2
f35103480f:internal/cli/factory_relaunch.go:3
f35103480f:internal/cli/factory_relaunch_test.go:10
f35103480f:internal/cli/lane_entry_m2_test.go:3
f35103480f:internal/cli/lane_resume.go:2
f35103480f:internal/config/envkeys.go:1
f35103480f:internal/factory/factory_relaunch_cmd.go:2
f35103480f:internal/factory/factory_relaunch_cmd_test.go:3
f35103480f:internal/homestate/factory_run_retire.go:1
f35103480f:internal/template/embed_manifest_gen.go:1
f35103480f:internal/template/rule_template_mirror_test.go:3
f35103480f:internal/template/templates/.claude/rules/moai/workflow/cache-aware-execution.md:1
f35103480f:internal/template/testdata/binding_ledger.json:2
```

### E-06 — verb surface of the tree build (verification; not a RED cell)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: binary output, not rebuilt at 8816de4dae; see the drift table in section A.

- Version output. Command: `<scratchpad>/moai-head-make version` — exit 0 — stdout (verbatim):

```
╭────────────────────────────────────╮
│                                    │
│  moai-adk moai_cp/20261007_195927  │
│                                    │
│                                    │
╰────────────────────────────────────╯
 moai_cp/20261007_195927   moai_cp/20261007_195927-847-gf35103480f   built 2026-10-14T03:13:23Z
```

  The stamped commit is `f35103480f` (the BuildID's `gf35103480f`), which equals the tree HEAD.

- Absence probes. Each command is `<scratchpad>/moai-head-make <verb> --help`. Stdout is empty for all eight; the messages are on the error channel (whitespace trimmed here).

| Command | Exit | Error channel (verbatim words) |
|---------|------|--------------------------------|
| `cg --help` | 1 | `moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration` |
| `gpt --help` | 1 | `ERROR`; `Unknown command "gpt" for "moai"`; `Did you mean this?` glm, gate, gtd, pr |
| `audit --help` | 1 | `ERROR`; `Unknown command "audit" for "moai".` |
| `codemaps --help` | 1 | `ERROR`; `Unknown command "codemaps" for "moai".` |
| `design --help` | 1 | `ERROR`; `Unknown command "design" for "moai".` |
| `model --help` | 1 | `ERROR`; `Unknown command "model" for "moai"`; `Did you mean this?` codex |
| `run --help` | 1 | `ERROR`; `Unknown command "run" for "moai".` |
| `sync --help` | 1 | `ERROR`; `Unknown command "sync" for "moai"`; `Did you mean this?` spec |

### E-07 — COMMANDS blocks, command column (verification; not a RED cell)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: binary output, not rebuilt at 8816de4dae; see the drift table in section A.

- `<scratchpad>/moai-head-make spec --help` — exit 0 — COMMANDS, command column: `status <SPEC-ID> <new-status>`, `drift`, `view <SPEC-ID>`, `lint [SPEC-ID | path/to/spec.md | SPEC directory ...]`, `close SPEC-ID`, `audit`, `archive`, `ceiling <SPEC-ID>`. No `list`.
- `<scratchpad>/moai-head-make todo --help` — exit 0 — COMMANDS, command column: `add <text>`, `list`, `done <n>`, `undone <n>`, `next [<n>]`, `claim`, `unpick <n>`, `edit <n> <text>`, `move <n>`, `drop <n> <reason>`, `undrop <n>`, `merge <into> <from>`, `hold <n>`, `unhold <n>`, `analyze`, `relate <a> <b>`, `unrelate <index>`, `why <n>`, `trace <id>`, `pr [<id>]`, `landed <n>`, `auto-done`, `export-json`, `history [<id|n>]`, `show <id|n>`, `triage <id>...`. No `rename`, no `pick` (`next` picks; `unpick` reverts).
- `<scratchpad>/moai-head-make codex --help` — exit 0 — COMMANDS, command column: `audit <role>`, `role-audit <role>`. No `job`, `task`, or `setup`.
- `<scratchpad>/moai-head-make web --help` — exit 0 — no COMMANDS block; flags `--port <int>`, `--no-open`, `--no-reuse`.

### E-07b — absent sub-verbs exit 0 and print the parent help (observation O8)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: binary output, not rebuilt at 8816de4dae; see the drift table in section A.

- `<scratchpad>/moai-head-make spec list` — no error displayed (exit 0) — stdout: the `spec` parent help, beginning `Manage SPEC documents in .moai/specs/ directory.`
- `<scratchpad>/moai-head-make codex job --help` — no error displayed (exit 0) — stdout: the `codex` parent help, beginning `Launch Codex, or ask what is installed and wired.`
- `<scratchpad>/moai-head-make todo rename --help` — no error displayed (exit 0) — stdout: the `todo` parent help, beginning `Operate the backlog queue at ~/.moai/db/<project-key>/todo/backlog.db.`
- Consequence for the producer: exit codes cannot mark an absent sub-verb; the COMMANDS block is the test (plan §B M1).

### E-08 — CG retirement sentence copies (RED-now; Korean pages)

- Command: `git grep -c -F "설정 이전 안내와 함께 종료합니다" 8816de4dae -- docs-site/content/ko`
- Exit code: 0.
- Stdout (verbatim; 9 lines; 8 outside the canonical page):

```
8816de4dae:docs-site/content/ko/advanced/tokenomics-overview.md:1
8816de4dae:docs-site/content/ko/cli-reference/launchers.md:1
8816de4dae:docs-site/content/ko/core-concepts/what-is-moai-adk.md:1
8816de4dae:docs-site/content/ko/cost-optimization/prompt-caching.md:1
8816de4dae:docs-site/content/ko/getting-started/cli.md:1
8816de4dae:docs-site/content/ko/getting-started/introduction.md:1
8816de4dae:docs-site/content/ko/getting-started/windows-guide.md:1
8816de4dae:docs-site/content/ko/multi-llm/_index.md:1
8816de4dae:docs-site/content/ko/multi-llm/cg-mode.md:1
```

### E-08b — link baseline to the CG page on the eight other Korean pages (RED-now)

- Command: `git grep -c -F "multi-llm/cg-mode" 8816de4dae -- docs-site/content/ko/advanced/tokenomics-overview.md docs-site/content/ko/cli-reference/launchers.md docs-site/content/ko/core-concepts/what-is-moai-adk.md docs-site/content/ko/cost-optimization/prompt-caching.md docs-site/content/ko/getting-started/cli.md docs-site/content/ko/getting-started/introduction.md docs-site/content/ko/getting-started/windows-guide.md docs-site/content/ko/multi-llm/_index.md`
- Exit code: 0.
- Stdout (verbatim; 7 lines; `what-is-moai-adk.md` has no match):

```
8816de4dae:docs-site/content/ko/advanced/tokenomics-overview.md:1
8816de4dae:docs-site/content/ko/cli-reference/launchers.md:2
8816de4dae:docs-site/content/ko/cost-optimization/prompt-caching.md:2
8816de4dae:docs-site/content/ko/getting-started/cli.md:1
8816de4dae:docs-site/content/ko/getting-started/introduction.md:1
8816de4dae:docs-site/content/ko/getting-started/windows-guide.md:2
8816de4dae:docs-site/content/ko/multi-llm/_index.md:2
```

### E-09a — unreferenced image paths (RED-now)

- Command: `git ls-tree -r --name-only 8816de4dae -- assets/images/cg-mode-infographic-en.png assets/images/cg-mode-infographic-ja.png assets/images/cg-mode-infographic-ko.png assets/images/cg-mode-infographic-zh.png assets/images/moai-adk-social.png`
- Exit code: 0.
- Stdout (verbatim; 5 lines):

```
assets/images/cg-mode-infographic-en.png
assets/images/cg-mode-infographic-ja.png
assets/images/cg-mode-infographic-ko.png
assets/images/cg-mode-infographic-zh.png
assets/images/moai-adk-social.png
```

### E-09b — references in the scoped paths (RED-now; rescoped per verdict D16)

- Command: `git grep -c -F -e "cg-mode-infographic" -e "moai-adk-social" 8816de4dae -- README.md README.ko.md README.ja.md README.zh.md docs-site assets .moai/docs .claude/skills internal`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### E-09b2 — tracked references in the whole tree (RED-now; the historical report named; this SPEC's directory excluded)

- Command: `git grep -l -F -e "cg-mode-infographic" -e "moai-adk-social" 8816de4dae -- ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001' .`
- Exit code: 0.
- Stdout (verbatim; 1 line):

```
8816de4dae:.moai/reports/historical/gpt-final-design-20260914-b40ff9f4/inventory.json
```

  This tracked historical report names the five images. It is a historical artifact outside the scoped paths of E-09b; it does not make the images referenced for display. The deletion decision (Q2) is unaffected.

  The historical report inventories past state and is not a live reference (leader ruling 75ac); expected value 1.

### E-09c — sizes of the five files (RED-now; `derived` total)

- Command: `git ls-tree -r -l 8816de4dae -- assets/images/cg-mode-infographic-en.png assets/images/cg-mode-infographic-ja.png assets/images/cg-mode-infographic-ko.png assets/images/cg-mode-infographic-zh.png assets/images/moai-adk-social.png`
- Exit code: 0.
- Stdout (verbatim; 5 lines):

```
100644 blob 1b97cbb3bc2adf559ffbaefb3791a82ed0c91982 2821584	assets/images/cg-mode-infographic-en.png
100644 blob 869801b6a606cc21465817522962a709e55e0d69  886241	assets/images/cg-mode-infographic-ja.png
100644 blob 46eb8b3aa1807733b5cf91ded1bcd8c228d9d480  877974	assets/images/cg-mode-infographic-ko.png
100644 blob e9fe3bf828be1cf0f7cd1547df4d1f172571b9fd  883349	assets/images/cg-mode-infographic-zh.png
100644 blob 59102d80f8f8b31e504da0ea41ceddfb4e2b463a  282554	assets/images/moai-adk-social.png
```

- Derived total: 5,751,702 bytes (5.75 MB decimal; the card's "5.8MB" is this sum rounded).

### E-10 — the OG image (RED-now)

- Command: `git cat-file -s 8816de4dae:assets/images/moai-adk-og.png` — exit 0 — stdout: `7837648`
- Command: `git grep -n -F "moai-adk-og.png" 8816de4dae -- README.md README.ko.md README.ja.md README.zh.md` — exit 0 — stdout (verbatim; 4 lines):

```
8816de4dae:README.ja.md:2:  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
8816de4dae:README.ko.md:2:  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
8816de4dae:README.md:2:  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
8816de4dae:README.zh.md:2:  <img src="./assets/images/moai-adk-og.png" alt="MoAI-ADK" width="100%">
```

### E-11a — Nextra-name counts (RED-now)

- Command: `git grep -c -i -e "nextra" 8816de4dae -- docs-site README.md README.ko.md README.ja.md README.zh.md scripts .claude/skills .claude/rules .moai/docs AGENTS.md`
- Exit code: 0.
- Stdout (verbatim; 7 lines):

```
8816de4dae:.claude/skills/moai-workflow-project/SKILL.md:3
8816de4dae:.moai/docs/docs-site-i18n-rules.md:1
8816de4dae:docs-site/README.md:1
8816de4dae:docs-site/content/en/core-concepts/ddd.md:1
8816de4dae:docs-site/scripts/fix-mdx-formatting.js:2
8816de4dae:docs-site/themes/hugo-geekdoc/static/js/search-8a051f6b.bundle.min.js:1
8816de4dae:scripts/convert-nextra-to-hextra/main.go:5
```

### E-11b — tracked files in the two script folders (RED-now)

- Command: `git ls-tree -r --name-only 8816de4dae -- docs-site/scripts scripts/convert-nextra-to-hextra`
- Exit code: 0.
- Stdout (verbatim; 7 lines):

```
docs-site/scripts/README_TRANSLATE.md
docs-site/scripts/fix-mdx-formatting.js
docs-site/scripts/generate-favicon-icons.md
docs-site/scripts/generate-favicons.js
docs-site/scripts/migrate-frontmatter.sh
docs-site/scripts/translate.mjs
scripts/convert-nextra-to-hextra/main.go
```

### E-12 — callers of the Nextra-era and translation scripts (RED-now)

- Command: `git grep -n -E "fix-mdx-formatting|convert-nextra-to-hextra|translate\.mjs|migrate-frontmatter" 8816de4dae -- .github Makefile lefthook.yml docs-site/hugo.toml docs-site/vercel.json scripts docs-site/scripts`
- Exit code: 0.
- Stdout (verbatim; 18 lines; every match is inside `docs-site/scripts/` or `scripts/convert-nextra-to-hextra/`; the paths `.github`, `Makefile`, `lefthook.yml`, `hugo.toml`, and `vercel.json` have none):

```
8816de4dae:docs-site/scripts/README_TRANSLATE.md:5:The `translate.mjs` script translates Korean MDX content to English, Japanese, and Chinese while preserving code blocks, frontmatter structure, and technical terms.
8816de4dae:docs-site/scripts/README_TRANSLATE.md:9:`/Users/goos/MoAI/moai-docs/scripts/translate.mjs`
8816de4dae:docs-site/scripts/README_TRANSLATE.md:27:node scripts/translate.mjs
8816de4dae:docs-site/scripts/README_TRANSLATE.md:30:node scripts/translate.mjs --target-locale en
8816de4dae:docs-site/scripts/README_TRANSLATE.md:33:node scripts/translate.mjs --dry-run
8816de4dae:docs-site/scripts/README_TRANSLATE.md:36:node scripts/translate.mjs --source-dir content/core-concepts --target-locale en
8816de4dae:docs-site/scripts/README_TRANSLATE.md:39:node scripts/translate.mjs --help
8816de4dae:docs-site/scripts/fix-mdx-formatting.js:11: * Usage: node scripts/fix-mdx-formatting.js
8816de4dae:docs-site/scripts/migrate-frontmatter.sh:2:# migrate-frontmatter.sh — Hextra → hugo-geekdoc frontmatter 마이그레이션
8816de4dae:docs-site/scripts/migrate-frontmatter.sh:12:#   ./scripts/migrate-frontmatter.sh --dry-run   # 변경 사항 미리보기
8816de4dae:docs-site/scripts/migrate-frontmatter.sh:13:#   ./scripts/migrate-frontmatter.sh             # 실제 적용
8816de4dae:docs-site/scripts/translate.mjs:14: *   node scripts/translate.mjs [options]
8816de4dae:docs-site/scripts/translate.mjs:126:  node scripts/translate.mjs [options]
8816de4dae:docs-site/scripts/translate.mjs:138:  node scripts/translate.mjs
8816de4dae:docs-site/scripts/translate.mjs:141:  node scripts/translate.mjs --target-locale en
8816de4dae:docs-site/scripts/translate.mjs:144:  node scripts/translate.mjs --dry-run
8816de4dae:docs-site/scripts/translate.mjs:147:  node scripts/translate.mjs --source-dir content/core-concepts --target-locale en
8816de4dae:scripts/convert-nextra-to-hextra/main.go:1:// convert-nextra-to-hextra: Bulk Nextra MDX → Hextra Markdown converter
```

### E-13 — the Python-era docs-site README (RED-now)

- Command: `git cat-file -s 8816de4dae:docs-site/README.md` — exit 0 — stdout: `42512`

### E-14a — llms.txt last change (RED-now)

- Command: `git log -1 --format=%cI 8816de4dae -- docs-site/static/llms.txt` — exit 0 — stdout: `2026-04-20T06:15:43Z`

### E-14b — EARS-style wording in llms.txt (RED-now)

- Command: `git grep -n -i -E "ears|shall|THEN" 8816de4dae -- docs-site/static/llms.txt`
- Exit code: 0.
- Stdout (verbatim):

```
8816de4dae:docs-site/static/llms.txt:8:- **SPEC 기반 개발**: EARS 형식으로 요구사항을 명확화하고 DDD 방식으로 안전하게 구현
```

### E-15a — redirect source keys (supporting)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -c -E "\"source\"" f35103480f -- docs-site/vercel.json` — exit 0 — stdout: `f35103480f:docs-site/vercel.json:42`
- Derived (E-15f): `git show f35103480f:docs-site/vercel.json | python3 -I -c "…"` (a parse of the pinned blob, outside the cell) — `redirects=39 rewrites=1 headers=2`.

### E-15b — design destinations (RED-now)

- Command: `git grep -n -E "\"destination\": \"[^\"]*design" 8816de4dae -- docs-site/vercel.json`
- Exit code: 0.
- Stdout (verbatim; 3 lines; the three sources are at lines 88, 93, and 98):

```
8816de4dae:docs-site/vercel.json:89:      "destination": "/:locale/design/:path*",
8816de4dae:docs-site/vercel.json:94:      "destination": "/:locale/design",
8816de4dae:docs-site/vercel.json:99:      "destination": "/design/:path*",
```

### E-15c — changelog redirect lines (RED-now)

- Command: `git grep -n -E "changelog" 8816de4dae -- docs-site/vercel.json`
- Exit code: 0.
- Stdout (verbatim; 2 lines; the destinations at lines 114 and 119 are `https://github.com/modu-ai/moai-adk/releases`, `"permanent": false`):

```
8816de4dae:docs-site/vercel.json:113:      "source": "/:locale(ko|en|ja|zh)/changelog",
8816de4dae:docs-site/vercel.json:118:      "source": "/changelog",
```

### E-15d-design — design content paths in the four locales (RED-now)

- Command: `git ls-tree -r --name-only 8816de4dae -- docs-site/content/en/design docs-site/content/en/design.md docs-site/content/ja/design docs-site/content/ja/design.md docs-site/content/ko/design docs-site/content/ko/design.md docs-site/content/zh/design docs-site/content/zh/design.md`
- Exit code: 0.
- Stdout: no output (no `design` path in any locale).

### E-15d-changelog — changelog content paths in the four locales (RED-now)

- Command: `git ls-tree -r --name-only 8816de4dae -- docs-site/content/en/changelog docs-site/content/en/changelog.md docs-site/content/ja/changelog docs-site/content/ja/changelog.md docs-site/content/ko/changelog docs-site/content/ko/changelog.md docs-site/content/zh/changelog docs-site/content/zh/changelog.md`
- Exit code: 0.
- Stdout (verbatim; 4 lines):

```
docs-site/content/en/changelog/_index.md
docs-site/content/ja/changelog/_index.md
docs-site/content/ko/changelog/_index.md
docs-site/content/zh/changelog/_index.md
```

### E-15e — Korean destinations of the non-design rules (verification; ko only)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git ls-tree -r --name-only f35103480f -- docs-site/content/ko/design docs-site/content/ko/design.md docs-site/content/ko/advanced/manager-lead.md docs-site/content/ko/advanced/manager-lead docs-site/content/ko/advanced/profile-matrix.md docs-site/content/ko/cli-reference/inventory.md docs-site/content/ko/cli-reference/status.md docs-site/content/ko/cli-reference/update.md docs-site/content/ko/cli-reference/profile.md docs-site/content/ko/core-concepts/book.md docs-site/content/ko/core-concepts/book docs-site/content/ko/guides/github-integration.md docs-site/content/ko/utility-commands/moai-codemaps.md docs-site/content/ko/utility-commands/_index.md docs-site/content/ko/changelog docs-site/content/ko/changelog.md docs-site/content/ko/advanced/factory-mode.md docs-site/content/ko/guides/github-integration docs-site/content/ko/cli-reference/_index.md`
- Exit code: 0.
- Stdout (verbatim; 13 lines; no `design` path is printed):

```
docs-site/content/ko/advanced/factory-mode.md
docs-site/content/ko/advanced/manager-lead.md
docs-site/content/ko/advanced/profile-matrix.md
docs-site/content/ko/changelog/_index.md
docs-site/content/ko/cli-reference/_index.md
docs-site/content/ko/cli-reference/inventory.md
docs-site/content/ko/cli-reference/profile.md
docs-site/content/ko/cli-reference/status.md
docs-site/content/ko/cli-reference/update.md
docs-site/content/ko/core-concepts/book.md
docs-site/content/ko/guides/github-integration.md
docs-site/content/ko/utility-commands/_index.md
docs-site/content/ko/utility-commands/moai-codemaps.md
```

- Reading: `docs-site/content/ko/changelog/_index.md` is the page that the shadowing rule of Q16 concerns. en, ja, and zh were not checked path by path (gap G11).

### E-16a — the workflow matches for hugo, link-check, and continue-on-error (RED-now)

- Command: `git grep -n -i -e "hugo" -e "link-check" -e "lychee" -e "linkcheck" -e "markdown-link" -e "continue-on-error" 8816de4dae -- .github/workflows`
- Exit code: 0.
- Stdout (verbatim; 13 lines):

```
8816de4dae:.github/workflows/ci.yml:257:  # blocks the PR feedback loop. NOT continue-on-error: a race failure must
8816de4dae:.github/workflows/ci.yml:262:  # green-mergeable regardless of this job's result. (continue-on-error:true
8816de4dae:.github/workflows/ci.yml:886:  # SPEC-V3R2-CON-001 — continue-on-error so CI doesn't block on missing registry
8816de4dae:.github/workflows/ci.yml:893:    continue-on-error: true
8816de4dae:.github/workflows/ci.yml:922:        # SPEC-ZONE-REGISTRY-RESYNC-001 (M2): this job is continue-on-error: true,
8816de4dae:.github/workflows/docs-i18n-check.yml:85:        continue-on-error: ${{ steps.mode.outputs.strict != 'true' }}
8816de4dae:.github/workflows/pr-multi-os-gate.yml:11:# a pinned SHA, continue-on-error, and the matrix keyed on `!= 'false'` — is
8816de4dae:.github/workflows/pr-multi-os-gate.yml:58:        continue-on-error: true
8816de4dae:.github/workflows/pr-multi-os-gate.yml:70:  # blocks the gate: there is no continue-on-error and no exclusion — the
8816de4dae:.github/workflows/release-pr-multi-os.yml:57:      # continue-on-error plus full-matrix-test's `!= 'false'` comparison keep
8816de4dae:.github/workflows/release-pr-multi-os.yml:64:        continue-on-error: true
8816de4dae:.github/workflows/release-pr-multi-os.yml:101:    # pinned to continue-on-error while it carried 71 known failures across
8816de4dae:.github/workflows/test-install.yml:46:  # originals. adk.mo.ai.kr serves the docs-site copies (Hugo static/), while
```

- The strictness step (re-measured at 8816de4dae (block below)): lines 67–78 set `strict=true` for `workflow_dispatch` with the input, `strict=false` for `push` (line 74), and `strict=false` for every other event, which is pull_request (lines 75–77). The run step (lines 88–92) uses `DOCS_I18N_STRICT=1` only when `strict` is `true`.

### E-16b — the i18n script references in the workflow (verification)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -n -F "scripts/docs-i18n-check.sh" f35103480f -- .github/workflows`
- Exit code: 0.
- Stdout (verbatim; 8 lines):

```
f35103480f:.github/workflows/docs-i18n-check.yml:13:#   Phase 2 (flip trigger: scripts/docs-i18n-check.sh reports 0 ERROR drifts
f35103480f:.github/workflows/docs-i18n-check.yml:25:      - 'scripts/docs-i18n-check.sh'
f35103480f:.github/workflows/docs-i18n-check.yml:35:      - 'scripts/docs-i18n-check.sh'
f35103480f:.github/workflows/docs-i18n-check.yml:63:        run: chmod +x scripts/docs-i18n-check.sh
f35103480f:.github/workflows/docs-i18n-check.yml:89:            DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh 2>&1 | tee check-output.txt
f35103480f:.github/workflows/docs-i18n-check.yml:91:            DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh 2>&1 | tee check-output.txt
f35103480f:.github/workflows/docs-i18n-check.yml:150:              'DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh',
f35103480f:.github/workflows/docs-i18n-check.yml:153:              'scripts/docs-i18n-check.sh',
```

### E-16c — the i18n script is tracked and executable (precondition; green)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git ls-tree f35103480f -- scripts/docs-i18n-check.sh` — exit 0 — stdout:

```
100755 blob 970c58dac392d693476b72290be46d34a37951f5	scripts/docs-i18n-check.sh
```

### E-16b-skill — negative existence statements about docs-i18n-check (RED-now; done check)

- Command: `git grep -n -i -E "docs-i18n-check.*(not exist|DO$)" 8816de4dae -- '.claude/skills/hns-oss-docs-*' .moai/docs/docs-site-i18n-rules.md`
- Exit code: 0.
- Stdout (verbatim; 3 lines):

```
8816de4dae:.claude/skills/hns-oss-docs-i18n-rules/SKILL.md:115:Also: the scripts `docs-i18n-check.sh` and `gen_menu.py` referenced there DO
8816de4dae:.claude/skills/hns-oss-docs-verify/SKILL.md:9:  docs-i18n-check.sh and gen_menu.py do not exist.
8816de4dae:.claude/skills/hns-oss-docs-verify/SKILL.md:23:`docs-i18n-check.sh` and `gen_menu.py` DO NOT exist — never shell out to
```

- Reading: the sentence at `hns-oss-docs-i18n-rules/SKILL.md` lines 115–116 reads "…referenced there DO" / "NOT exist." The pattern is therefore written to match the split line. The unrelated line `.claude/skills/moai-workflow-project/SKILL.md:157` ("…features that do not exist in the codebase") matched the 0.1.0 broad pattern and is outside this pattern's file set.

### E-16b-hextra — Hextra as the current toolchain (RED-now; done check)

- Command: `git grep -n -F "Hextra" 8816de4dae -- .moai/docs/docs-site-i18n-rules.md`
- Exit code: 0.
- Stdout (verbatim; 2 lines):

```
8816de4dae:.moai/docs/docs-site-i18n-rules.md:35:### §17.2 Markdown/Hextra 작성 규칙
8816de4dae:.moai/docs/docs-site-i18n-rules.md:37:**[HARD] 강조 표기와 괄호 사이 공백 필수** (Nextra 시대 유래, Hextra/Goldmark 검증 완료 후 유지 여부 재평가):
```

### E-16d — docs i18n check, strict mode (green; AC-PRC-012 precondition)

- Command: `DOCS_I18N_STRICT=1 scripts/docs-i18n-check.sh` (run with the worktree root as the working directory; the measurement used the absolute path of the same tracked script, which resolves its root from its own location).
- Exit code: 0.
- Stdout (verbatim):

```

=== Preflight: locale directory presence ===

=== Check 1: File path parity (ko ↔ target locales) ===
ℹ️  INFO   ko: 153 .md files (canonical)
ℹ️  INFO   en: 153 .md files
ℹ️  INFO   ja: 153 .md files
ℹ️  INFO   zh: 153 .md files

=== Check 2: Frontmatter title presence ===

=== Check 3: H1 heading existence ===

=== Check 4: Glossary term preservation (canonical: ko) ===

=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

- Reading: the script writes only to a `mktemp` directory and removes it on exit (`scripts/docs-i18n-check.sh` line 96). The tracked tree is unchanged.

### E-17a — Hugo version (RED-now; pin check)

- Command: `hugo version` — exit 0 — stdout: `hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew`

### E-17b — the Vercel Hugo pin (RED-now; pin check)

- Command: `git grep -n -F "HUGO_VERSION" 8816de4dae -- docs-site/vercel.json`
- Exit code: 0.
- Stdout (verbatim): `8816de4dae:docs-site/vercel.json:11:      "HUGO_VERSION": "0.160.1",`

### E-18a — link-check tools in the workflows (RED-now)

- Command: `git grep -c -i -E "lychee|linkinator|htmltest|link-check|linkcheck|markdown-link-check" 8816de4dae -- .github/workflows`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### E-18b — link-check tools installed locally (RED-now)

- `command -v lychee` — exit 1 — stdout: none (not installed).
- `command -v linkinator` — exit 1 — stdout: none.
- `command -v htmltest` — exit 1 — stdout: none.
- `command -v markdown-link-check` — exit 1 — stdout: none.
- `command -v hugo` — exit 0 — stdout: `/opt/homebrew/bin/hugo` (control).

### E-19a — a Hugo build step in any workflow (RED-now)

- Command: `git grep -c -i -E "hugo +(--minify|build)|setup-hugo|hugo-action|peaceiris|cli/hugo" 8816de4dae -- .github/workflows`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### E-19b — local Hugo build into the scratchpad (green; AC-PRC-013 precondition)

- Command: `hugo -s /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655/docs-site --destination <scratchpad>/hugo-public --cacheDir <scratchpad>/hugo-cache --minify --gc`
- Exit code: 0.
- Stdout (verbatim):

```
Start building sites … 
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew


              │ KO  │ EN  │ JA  │ ZH  
──────────────┼─────┼─────┼─────┼─────
 Pages        │ 187 │ 185 │ 185 │ 185 
 Paginator    │   0 │   0 │   0 │   0 
 pages        │     │     │     │     
 Non-page     │  12 │  12 │  12 │  12 
 files        │     │     │     │     
 Static files │ 264 │ 264 │ 264 │ 264 
 Processed    │   0 │   0 │   0 │   0 
 images       │     │     │     │     
 Aliases      │   6 │   5 │   5 │   5 
 Cleaned      │   0 │   0 │   0 │   0 

Total in 16237 ms
```

- Reading: the build writes to the scratchpad only; `git status --short` after the build is empty. The Total in N ms line varies from run to run; it is included verbatim and is not a compared value. The page counts and the exit code are the checked values.

### E-20a — README Release badges (supporting; tree)

- Command: `git grep -n -F "img.shields.io/badge/Release-" 8816de4dae -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 0.
- Stdout (verbatim; 4 lines):

```
8816de4dae:README.ja.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.1.3-blue.svg" alt="Release"></a>
8816de4dae:README.ko.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.1.3-blue.svg" alt="Release"></a>
8816de4dae:README.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.1.3-blue.svg" alt="Release"></a>
8816de4dae:README.zh.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.1.3-blue.svg" alt="Release"></a>
```

### E-20b — hugo.toml version lines (supporting; tree)

- Command: `git grep -n -E "version|releaseDate" 8816de4dae -- docs-site/hugo.toml`
- Exit code: 0.
- Stdout (verbatim):

```
8816de4dae:docs-site/hugo.toml:55:  version = "v3.1.3"
8816de4dae:docs-site/hugo.toml:56:  releaseDate = "2026-08-24"
```

### E-20c — public release list (external; not tree-attributed)

- Command: `gh release list --repo modu-ai/moai-adk --limit 6` — observed in this revision, before `2026-10-14T03:19:53Z` — exit 0 — stdout (verbatim; tab-separated):

```
v3.1.3	Draft	v3.1.3	2026-08-21T19:28:08Z
v3.1.2	Latest	v3.1.2	2026-08-21T06:50:41Z
v3.1.1		v3.1.1	2026-08-20T09:15:39Z
v3.1.0		v3.1.0	2026-08-15T14:55:53Z
v3.0.1		v3.0.1	2026-07-24T03:27:57Z
v3.0.0		v3.0.0	2026-07-19T21:32:00Z
```

### E-21 — required status checks of `main` (external; not tree-attributed)

- Command: `gh api repos/modu-ai/moai-adk/branches/main/protection` — observed in this revision, before `2026-10-14T03:19:53Z` — exit 0 — stdout (verbatim; one JSON line):

```
{"url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection","required_status_checks":{"url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection/required_status_checks","strict":false,"contexts":["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS Gate"],"contexts_url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection/required_status_checks/contexts","checks":[{"context":"Test (ubuntu-latest)","app_id":15368},{"context":"Lint","app_id":15368},{"context":"Build (linux/amd64)","app_id":15368},{"context":"Analyze (Go) (go)","app_id":15368},{"context":"Release PR Multi-OS Gate","app_id":15368}]},"required_pull_request_reviews":{"url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection/required_pull_request_reviews","dismiss_stale_reviews":true,"require_code_owner_reviews":false,"require_last_push_approval":false,"required_approving_review_count":0},"required_signatures":{"url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection/required_signatures","enabled":false},"enforce_admins":{"url":"https://api.github.com/repos/modu-ai/moai-adk/branches/main/protection/enforce_admins","enabled":true},"required_linear_history":{"enabled":false},"allow_force_pushes":{"enabled":false},"allow_deletions":{"enabled":false},"block_creations":{"enabled":false},"required_conversation_resolution":{"enabled":false},"lock_branch":{"enabled":false},"allow_fork_syncing":{"enabled":false}}
```

- Reading: the five contexts are the same five that the verdict recorded. None of the three card jobs is among them, and the i18n job, although it runs on pull requests, is not a required context.

### E-22 — badge policy in committed docs and rules (supporting; Q8 anchor)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -n -i -E "badge" f35103480f -- .moai/docs .claude/rules AGENTS.md`
- Exit code: 0.
- Stdout (verbatim): `f35103480f:.claude/rules/moai/languages/rust.md:108:- `/launchbadge/sqlx` - SQLx async SQL` (an unrelated package name; no release-badge policy exists).

### E-23a — strict mode in the docs-site and CG-retirement SPECs (supporting)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -c -i -E "strict" f35103480f -- .moai/specs/SPEC-DOCS-SITE-001 .moai/specs/SPEC-MOAI-CG-RETIRE-001`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### E-23b — SPEC status and phase lines (supporting)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -n -E "^status:|^phase:" f35103480f -- .moai/specs/SPEC-MOAI-CG-RETIRE-001/spec.md .moai/specs/SPEC-DOCS-SITE-001/spec.md`
- Exit code: 0.
- Stdout (verbatim):

```
f35103480f:.moai/specs/SPEC-DOCS-SITE-001/spec.md:4:status: implemented
f35103480f:.moai/specs/SPEC-DOCS-SITE-001/spec.md:11:phase: "v2.x - Legacy"
f35103480f:.moai/specs/SPEC-MOAI-CG-RETIRE-001/spec.md:5:status: draft
f35103480f:.moai/specs/SPEC-MOAI-CG-RETIRE-001/spec.md:10:phase: "v3.3.0 target"
```

### E-24 — four-locale same-PR obligation, corrected pattern (supporting; Q17)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -n -E "동일 PR|4개 locale|4로케일|four locales|same PR|en/zh/ja" f35103480f -- .moai/docs/docs-site-i18n-rules.md`
- Exit code: 0.
- Stdout (verbatim; 4 lines):

```
f35103480f:.moai/docs/docs-site-i18n-rules.md:61:기능 추가/변경/제거로 ko 문서가 바뀌면 **동일 PR 내에서 en/zh/ja 모두 반영**. 예외: 5,000 단어 이상 대규모 콘텐츠는 ko 머지 후 48h 이내 en, 추가 72h 이내 zh/ja. 해당 파일 front matter에 `translation_status: pending` 필수.
f35103480f:.moai/docs/docs-site-i18n-rules.md:65:- [ ] ko/en/zh/ja 4개 locale 해당 파일 모두 수정
f35103480f:.moai/docs/docs-site-i18n-rules.md:78:- 4개 locale 간 파일 개수/경로 일치
f35103480f:.moai/docs/docs-site-i18n-rules.md:110:- [ ] Vercel Preview URL 수동 검증 — 4개 locale 홈 페이지
```

- Reading: the 0.1.0 pattern (`같은 PR|4로케일|four locales|same PR`) returned no match, which was a false negative (E-24 in 0.1.0). Lines 61 and 65 are the four-locale obligation. Lines 78 and 110 name the locale count only.

### E-25a — help text does not name the flags (supporting; Q18)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: binary output, not rebuilt at 8816de4dae; see the drift table in section A.

- Commands: `<scratchpad>/moai-head-make cc --help` (91 lines), `… factory --help` (30 lines), `… goal --help` (55 lines), and `… codex --help` (read in full). Exit 0 for each.
- Observation: none of the four outputs contains `factory-run` or `clear-policy`. The full stdout is not reproduced here (supporting cell, about 180 lines); each output was read in full in this revision. `cc` was not measured in 0.1.0, so gap G6 is closed for `cc`.

### E-26a — the design workflow is not user-invocable (supporting; Q14)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git grep -n -i -E "user-invocable|user_invocable" f35103480f -- .claude/skills/moai/workflows/design.md`
- Exit code: 0.
- Stdout (verbatim): `f35103480f:.claude/skills/moai/workflows/design.md:11:user-invocable: false`

### E-26b — registered slash commands (supporting)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `git ls-tree -r --name-only f35103480f -- .claude/commands/moai`
- Exit code: 0.
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

### E-27a — SPEC ID pre-write self-check (protocol; not a RED cell)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: unchanged.

- Command: `ID="SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` — exit 0 — stdout: `PASS`

### E-27b — the SPEC ID is not tracked (protocol; supporting)

> Measured at f35103480f in revision 0.1.1; compared at 8816de4dae in revision 0.2.2: changed; see the drift table in section A.

- Command: `git ls-tree -r --name-only f35103480f -- .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001 .moai/specs/_archive/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001` — exit 0 — stdout: no output.

### E-28 — no cited-verbs producer in the tracked tree (RED-now; this SPEC's directory excluded)

- Command: `git grep -n -i -E "cited-verbs|cited_verbs|citedverbs|docs-cited|cited verbs" 8816de4dae -- ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001' .`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

This SPEC's directory is excluded because the SPEC quotes the pattern it measures. Without the exclusion the query returns 22 lines, all inside this directory (measured at 8816de4dae). No producer exists outside this directory.

## D. Definition of done — umbrella closing table (every AC is listed; each is release-blocking (gated) or a regression-guard with a reason)

Completion verdict. The card's verdict counts the release-blocking rows at their gates, and it reports the regression-guard rows separately. A missing external observation is recorded as a gap, not a pass (gap G1). M6 re-runs the same set on the closing tree.

The umbrella closes only after every child SPEC has landed (owner column; spec.md § Child SPECs). M6 then re-runs this table on the closing tree, and the GATE child's citation checker must print `cited-verbs: 0` with exit 0.

| AC | REQ | Owner child | Class | Gate (decision rows) | Closing check (re-run at M6) |
|----|-----|-------------|-------|----------------------|------------------------------|
| AC-PRC-001 | REQ-PRC-001 | README (t1720); SWEEPA to SWEEPF (t1721 to t1725, t1727), through partial criteria | release-blocking (gated) | Q13, Q14; M1 producer | E-01, E-02, E-03, E-04 exit 1 with empty stdout; producer exits 0 |
| AC-PRC-002 | REQ-PRC-002 | GATE (t1717) | release-blocking (gated) | M1 inputs (Q5, Q18); M2 step; Q12, Q22 if required | fixtures (a) red, (b) green, (c) red; pull-request run red on injected token; closing run prints `cited-verbs: 0`, exit 0 |
| AC-PRC-003 | REQ-PRC-003 | README (t1720) | regression-guard | Q8 | E-20a, E-20b match the Q8 verdict; E-20c recorded as an external observation. Reason: the closing comparison needs the external Latest release, which the repository cannot observe (verification-completeness §2.1). |
| AC-PRC-004 | REQ-PRC-004 | README (t1720) | release-blocking (gated) | Q1 | E-08 lists exactly one file (the Q1 page); each of the eight other pages links to it (E-08b pattern, count ≥ 1) |
| AC-PRC-005 | REQ-PRC-005 | RESID (t1719) | release-blocking (gated) | Q2 | E-09a empty on the closing tree; E-09b exit 1; E-09b2 recorded |
| AC-PRC-006 | REQ-PRC-006 | RESID (t1719) | release-blocking (gated; open until Q3 has a verdict) | Q3 | blob size at or below the Q3 verdict value; E-10 refs resolve |
| AC-PRC-007 | REQ-PRC-007 | RESID (t1719) | release-blocking (gated) | Q4 | `ls-tree` of the two scripts empty; E-16b-skill exit 1; E-16b-hextra patterns exit 1 |
| AC-PRC-008 | REQ-PRC-008 | RESID (t1719) | release-blocking (gated) | Q6 | `cat-file` of `docs-site/README.md` fails, or the replacement passes its check |
| AC-PRC-009 | REQ-PRC-009 | RESID (t1719) | release-blocking (gated; operator-held) | Q7 | file absent, or the E-14b pattern exits 1 |
| AC-PRC-010 | REQ-PRC-010 | REDIR (t1718) | release-blocking (gated) | Q15, Q21 | E-15b pattern exits 1; each remaining destination maps to a same-locale page in four locales, one cell per path |
| AC-PRC-011 | REQ-PRC-011 | REDIR (t1718) | release-blocking (gated) | Q16; Q23 for the moai-mx pair | E-15d-changelog empty, or the redirect sources removed (Q16 alternate); the moai-mx pair at lines 163 and 168 resolved per Q23 |
| AC-PRC-012 | REQ-PRC-012 | GATE (t1717) | release-blocking (gated) | Q9 | pull-request run red on injected drift; closing run green; the closing record names the pull-request run |
| AC-PRC-013 | REQ-PRC-013 | GATE (t1717) | release-blocking (gated) | Q10 | pull-request job runs the Q10 Hugo version; injected template error red |
| AC-PRC-014 | REQ-PRC-014 | GATE (t1717) | release-blocking (gated) | Q11 | pull-request job red on an injected broken link; green on the closing tree |
| AC-PRC-015 | REQ-PRC-015 | parent (operator-held) | regression-guard (operator-held) | Q12, Q22 | external observation lists the job names, recorded with its time. Reason: branch protection is external and operator-held (keep-set), so the observation is recorded, not passed. |

## E. Test named, not run

Named in plan.md §E and not run here: `docs_cited_verbs_test` (three fixture cases), `hugo_build_gate_red_on_broken_shortcode`, `link_check_red_on_broken_internal_link`, and `redirect_destinations_resolve_all_locales`.

## F. Gaps (not observed)

- G1 — Branch protection effect. The blocking status of the three jobs is a GitHub branch-protection fact that the repository cannot prove. E-21 shows the three jobs are not required today, and AC-PRC-015 is a regression-guard.
- G2 — External observations were made in revision 0.1.1, before `2026-10-14T03:19:53Z`. They are not the closing-time state.
- G3 — The Vercel `installCommand` (`hugo mod get -u`) was not run. The local Hugo build (E-19b) is not a full CI replica.
- G4 — No link checker was run. None of the four probed tools is installed (E-18b).
- G5 — No live request was made to the deployed site. The current behavior of the redirect rules (Q15, Q16, Q21) is inferred from the content tree and the redirect precedence, and the inference is labelled in decision-index.md.
- G6 — Closed for `cc`, `factory`, `goal`, and `codex` help (E-25a): none names the two flags. The flags are still a decision (Q18).
- G7 — The card's "쪽" page measure has no repository definition. AC-PRC-004 uses the sentence-copy count as a proxy.
- G8 — The stale-reference counts (E-01 to E-04) come from a sweep, not from the producer. The producer's independent run is the M1 deliverable.
- G9 — The en, ja, and zh copies of the CG passage were not counted by sentence. Only the family cells cover them.
- G10 — The `docs-site/content/en/core-concepts/ddd.md` "nextra" line is not classified (E-11a).
- G11 — The redirect destinations were checked path by path for ko only (E-15e). en, ja, and zh are open.
- G12 — The strict i18n run (E-16d) is one observation in this revision and is not repeated on a closing tree.
- G13 — The installed `moai` (v3.2.0-rc.31, commit ec193e4777) was not used. The verb-surface observations come from the tree build only (E-06, E-07).
- G14 — Corrected in 0.1.1. The four-locale same-PR obligation is stated in `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 (E-24, corrected pattern). That document is outside the authority register, so the obligation's scope for this SPEC is an open decision (Q17). The 0.1.0 statement that no committed document carries it was false.
- G15 — E-02 and E-04 still count some non-command lines. E-02's `web console` lines and the moai-todo examples are excluded in 0.1.1. E-04's narrative examples are open (N2).
- G16 — Derived totals (awk line sums, the Python redirect parse) are not cells. They are reproducible from the commands recorded above.
- G17 — The M2 workflow edit (plan §B M2) is not performed in this revision, because this revision writes only inside the SPEC directory.
- G18 — The `build` recipe's prerequisites were not run (plan §A.2). The binary is built from the recipe's `go build` line, with the scratchpad as its output.
- G19 — Zero-match exit codes are confirmed with a measurement wrapper outside the cell only where marked `wrapper-confirmed`; the other zero-match entries show `no output`.
- G20 — The Q7 regenerate branch has no generator in the tree. The verdict's section 5, item 8 reports the same finding; it was not re-measured in revision 0.1.1.
- G21 — Child SPEC content is written by the lanes that hold the child cards, and it is not observed in revision 0.2.1; no child spec.md, plan.md, or acceptance.md exists yet.
