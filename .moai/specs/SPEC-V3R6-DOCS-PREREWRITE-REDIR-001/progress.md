# Progress — SPEC-V3R6-DOCS-PREREWRITE-REDIR-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
audit_verdict: plan audit iter2 PASS 0.91, audited_sha 49cb21de9, verdict file .moai/reports/t1718/plan-audit-iter2.md; history: iter1 FAIL 0.72 (.moai/reports/t1718/plan-audit-iter1.md).
plan_revision: 0.3.0
plan_complete_at: 2026-10-18T17:06:37Z
plan_revision_2_at: 2026-10-18T17:13:14Z
plan_revision_3_at: 2026-10-18T20:16:39Z
plan_base_commit: 0ddf5ca48bed7aaeeb68ed4fc686bcfa098a82b1
revision_note: r3: plan-audit iter1 FAIL 0.72 (.moai/reports/t1718/plan-audit-iter1.md) D1-D4 fixed and D5-D10 done; the check script compares source and destination locale, scopes the admitted URL to its two sources, rejects unsupported source characters; AC-REDIR-002 uses numstat 0 20; REQ-REDIR-003 maps to AC-REDIR-003 D0; suffixes renumbered. Mutants observed in acceptance.md E-R14. r2: leader verdicts a207 applied (board d-20261018T170818Z-a207, decided by the leader under operator delegation with Jev 0.88 and 0.92). Q23 = (c) remove only line 163; Q21 = (b) keep the two external changelog rules and amend REQ-PRC-010; Q16 = delete the four pages; Q15 and Q17 stay DEFAULT-APPLIED. Footer file not changed; the five-file scope is final.
open_decisions: none. Every decision row carries a verdict or a DEFAULT-APPLIED line (decision-index.md).
lint: `moai spec lint <this SPEC directory> --strict` reports 0 errors and 0 warnings after the r3 edits (one INFO: OwnershipTransitionUnmeasured, the lane commit cde13ce99 carries no Authored-By-Agent trailer) (installed build moai_cp/20261007_195927-1888-g7705b365e2, a strict ancestor of the tree HEAD; internal/spec unchanged between the two, acceptance.md G11).
scratch_proof: the exact a207 change set was applied on scratch copies under /private/tmp only; shadow turns green, resolve turns green only with the Q15 Default included (acceptance.md E-R14, G14).
ordering_note: the Q23 row (decision-index.md) and the AC that covers the moai-mx rule (AC-PRC-011, acceptance.md § B) are the first row and the first criterion in their files. That ordering is positional, not a commit-order witness.

The plan is ready for the independent plan audit. The Kickoff gate has no open decision row.

## §E.2 Run-phase Evidence

### M0 Baseline (no file edits)

- BASE: `git rev-parse HEAD` printed `56873b261f900fa363b681cbe41d3e37e807f78f`; `git status --short` printed nothing (clean tree), exit 0, branch `WT-id-spec-v3r6`.
- Instrument: the E-R0 block of acceptance.md (lines 109 to 218) was extracted verbatim to `/private/tmp/t1718-run/redirect-check.py` (outside the repository). `shasum -a 256` printed `d9074c46307802d5db8cffc7931fb2083440066c57a81d4fe6caeea9f4263ebf` (equals the value in acceptance.md E-R0), 110 lines.
- Baselines on BASE (run from the repository root with `python3 -I`):
  - E-R5 `shadow docs-site`: exit 1; first line `swept: redirects=39 pages=612 locales=en,ja,ko,zh`; last line `shadowed: 8` (10 lines, equals E-R5).
  - E-R4 `resolve docs-site`: exit 1; first line `swept: redirects=39 pages=612 locales=en,ja,ko,zh`; last line `unresolved: 14` (16 lines, equals E-R4).
  - E-R14a `resolve docs-site https://github.com/modu-ai/moai-adk/releases '/:locale(ko|en|ja|zh)/changelog' /changelog`: exit 1; `admitted external: 5`, `unresolved: 9` (equals E-R14a).
  - AC-REDIR-003 D1 baseline `bash scripts/docs-i18n-check.sh`: exit 0; tail `Errors:   0`, `Warnings: 0`, `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.`
  - AC-REDIR-003 D2 baseline `hugo --minify --gc --source docs-site --destination /private/tmp/t1718-run/public-base`: exit 0, 18 output lines, `grep -c -E 'WARN|ERROR'` over them printed 0 (no warning or error line), `Total in 5930 ms`. The gitignored `docs-site/.hugo_build.lock` the build left was deleted afterwards (`rm -f`); `git status --short` printed nothing.
- Controls re-run with the extracted instrument on synthetic fixtures built by `/private/tmp/t1718-run/mkctl.py` (E-R6 equivalents): `pos` shadow exit 1 `shadowed: 2`; `pos` resolve exit 1 `unresolved: 2`; `neg` shadow and resolve exit 0; `empty` shadow and resolve exit 2 (`redirect-check: empty sweep: redirects=0 pages=4`); `tok` shadow exit 2 (`redirect-check: unsupported source token: /ko/:slug`). All equal E-R6.
- E-R14 equivalents on scratch copies built by `/private/tmp/t1718-run/mkexact.py` from the BASE `docs-site` (whole-line deletion; `_meta.yaml` omitted): `before` shadow exit 1 `shadowed: 8`, resolve-with-admitted exit 1 `admitted external: 5` / `unresolved: 9`; `good` (lines 87-101 and 162-166 deleted, four changelog pages deleted) shadow exit 0 `swept: redirects=35 pages=608` / `shadowed: 0`, resolve-with-admitted exit 0 `admitted external: 5` / `unresolved: 0`, resolve-without-admitted exit 1 `unresolved: 5`, numstat `0	20`; `leader-only` shadow exit 0, resolve-with-admitted exit 1 `unresolved: 9`; `mut-replace` both exit 0, numstat `1	21`. All equal E-R14.
- E-R8 (external; observation, not a pass) taken 2026-10-19T00:09:51Z to 00:09:52Z with `curl -s -o /dev/null -m 15 -w 'http=%{http_code} location=%{redirect_url}\n' https://adk.mo.ai.kr<path>`: the 19 lines (17 URLs of the E-R8 list plus the two controls `/ko/utility-commands/moai-fix` http=200 and `/en/book` http=308) equal the 2026-10-18 observation (moai-mx URLs 308 to the section index, changelog URLs 307 to the GitHub Releases URL, `/ko/agency` 308 to `/ko/design/` then 308 to `/ko/design` then 404). The deployed commit is unknown (G1).
- In-flight branch overlap (plan M0 item): not re-measured in this run (the loop over branches is a compound form the worktree guard refused for the plan author, G5; no enumerated branch list was supplied). The leader's relayed observation stands as relayed.

### M1 and M2 commits

- M1 `15b596315b0088fa422c4b4c645f8952e9326854`: `docs-site/vercel.json` only, 20 deletions (`git show --stat`: `1 file changed, 20 deletions(-)`). Removed sources: `/:locale(ko|en|ja|zh)/agency/:path*`, `/:locale(ko|en|ja|zh)/agency`, `/agency/:path*`, `/:locale(ko|en|ja|zh)/utility-commands/moai-mx`. Before commit: `git diff --numstat 56873b261 -- docs-site/vercel.json` printed `0	20	docs-site/vercel.json`; `python3 -I -c "import json; ..."` loaded the file (`valid json, redirects=35`).
- M2 `dc5b1f2f42f1ffbcb56bf961f9d0af5c38e6abab` (closing change commit): `git rm` of the four `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` files (`4 files changed, 58 deletions(-)`). HEAD and branch were re-read before each commit (`56873b261f` then `15b596315b`, branch `WT-id-spec-v3r6` both times). Staged by explicit pathspec; messages by `git commit -F`; both end with `🗿 MoAI`.

### M3 Done checks at the closing commit

All run from the repository root against HEAD `dc5b1f2f42f1ffbcb56bf961f9d0af5c38e6abab` (working tree equals HEAD except this file, which is modified and not yet committed), on 2026-10-19 around 00:10Z to 00:12Z. `<R>` is `/private/tmp/t1718-run/redirect-check.py` (SHA-256 `d9074c46...263ebf`, re-confirmed after the run). `exit` is the status the shell reported for the command itself. BASE is `56873b261f`.

AC-PRC-011 (gating)

| Item | Command | Verbatim stdout | Exit | Result |
|------|---------|-----------------|------|--------|
| D1 | `python3 -I <R> shadow docs-site` | `swept: redirects=35 pages=608 locales=en,ja,ko,zh` / `shadowed: 0` | 0 | PASS |
| D2 | `git ls-tree -r --name-only HEAD -- docs-site/content/{ko,en,ja,zh}/utility-commands/moai-mx.md` (four explicit paths) | four lines: `docs-site/content/en/utility-commands/moai-mx.md`, `.../ja/...`, `.../ko/...`, `.../zh/...` | 0 | PASS |
| D3 | `git grep -n -E '^draft:' HEAD -- <the same four paths>` | four lines, each `HEAD:docs-site/content/<loc>/utility-commands/moai-mx.md:4:draft: false` | 0 | PASS |
| D4 | `git ls-tree -r --name-only HEAD -- docs-site/content/{ko,en,ja,zh}/changelog` (four explicit paths) | (no output) | 0 | PASS |
| D5a | `git grep -n -F '"source": "/:locale(ko\|en\|ja\|zh)/utility-commands/moai-mx"' HEAD -- docs-site/vercel.json` | (no output) | 1 | PASS |
| D5b | `git grep -n -F '"source": "/utility-commands/moai-mx"' HEAD -- docs-site/vercel.json` | `HEAD:docs-site/vercel.json:148:      "source": "/utility-commands/moai-mx",` | 0 | PASS |
| D6 | controls: E-R6 fixtures re-run (`mkctl.py`) after the closing commit; E-R14 copies at M0 | see M0 and the line below | n/a | PASS |

D6 detail: after the closing commit `pos` shadow exit 1 `shadowed: 2`, `pos` resolve exit 1 `unresolved: 2`, `neg` both exit 0, `empty` both exit 2, `tok` shadow exit 2. The E-R14 copies were built at M0 (before the edit); they were not rebuilt after the closing commit because `mkexact.py` reads the repository content tree, which no longer holds the BASE pages.

AC-PRC-010 (gating)

| Item | Command | Verbatim stdout | Exit | Result |
|------|---------|-----------------|------|--------|
| D1 | `python3 -I <R> resolve docs-site https://github.com/modu-ai/moai-adk/releases '/:locale(ko\|en\|ja\|zh)/changelog' /changelog` | `swept: redirects=35 pages=608 locales=en,ja,ko,zh` / `admitted external: 5` / `unresolved: 0` | 0 | PASS |
| D2 | `git grep -n -E '"source": "[^"]*agency' HEAD -- docs-site/vercel.json` | (no output) | 1 | PASS |
| D3a | `git grep -n -F -e '"source": "/changelog"' -e '"source": "/:locale(ko\|en\|ja\|zh)/changelog"' HEAD -- docs-site/vercel.json` | `HEAD:docs-site/vercel.json:98:      "source": "/:locale(ko\|en\|ja\|zh)/changelog",` / `HEAD:docs-site/vercel.json:103:      "source": "/changelog",` | 0 | PASS |
| D3b | `git grep -n -F '"destination": "https://github.com/modu-ai/moai-adk/releases"' HEAD -- docs-site/vercel.json` | `...:99:      "destination": "https://github.com/modu-ai/moai-adk/releases",` / `...:104:      "destination": "https://github.com/modu-ai/moai-adk/releases",` | 0 | PASS |
| D4 | controls of E-R6 and E-R14 | as D6 above | n/a | PASS |

AC-REDIR-001 (guard, observation)

| Item | Command | Verbatim stdout | Exit |
|------|---------|-----------------|------|
| D1a | `git grep -n -E 'href="/changelog' HEAD -- docs-site` | `HEAD:docs-site/layouts/partials/site-footer.html:9:        <a href="/changelog/">{{ i18n "footer_changelog" }}</a> ·` | 0 |
| D1b | the D3a command of AC-PRC-010 | the two lines shown above (lines 98 and 103) | 0 |
| D2 | `git diff --name-only 56873b261f..HEAD -- docs-site/layouts docs-site/i18n` | (no output) | 0 |
| D3 | `git grep -c -E '/utility-commands/moai-mx' HEAD -- docs-site ':!docs-site/vercel.json'` | 13 file entries: per locale `advanced/mx-scanner-internals.md:2`, `cli-reference/graph.md:1`, `workflow-commands/_index.md:1`, plus `docs-site/data/menu/main.yaml:1` (equals the E-R9c shape) | 0 |

AC-REDIR-002 (guard, observation)

| Item | Command | Verbatim stdout | Exit |
|------|---------|-----------------|------|
| D1 | `git diff --name-only 56873b261f..HEAD -- . ':!docs-site/vercel.json' ':!docs-site/content/*/changelog/_index.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-REDIR-001' ':!.moai/reports/t1718'` | (no output) | 0 |
| D1 control | `git diff --name-only 56873b261f..HEAD -- docs-site/vercel.json docs-site/content` | five lines: the four `docs-site/content/{en,ja,ko,zh}/changelog/_index.md` paths and `docs-site/vercel.json` | 0 |
| D2 | `git diff --numstat 56873b261f..HEAD -- docs-site/vercel.json` | `0	20	docs-site/vercel.json` | 0 |
| D3a | `git diff --name-only --diff-filter=D 56873b261f..HEAD -- docs-site/content` | exactly the four `docs-site/content/{en,ja,ko,zh}/changelog/_index.md` paths | 0 |
| D3b | `git diff --name-only --diff-filter=AMR 56873b261f..HEAD -- docs-site/content` | (no output) | 0 |
| D4 | the four removed sources (agency x3, locale-aware moai-mx) plus the 20 deleted lines | `git diff -U0 56873b261 -- docs-site/vercel.json` before M1 showed 20 `-` lines and exactly those four `"source"` lines; the `swept: redirects=35` count agrees (observation) | n/a |

AC-REDIR-003

| Item | Command | Verbatim output | Exit | Result |
|------|---------|-----------------|------|--------|
| D0 | same command as AC-PRC-011 D4 | (no output; pre-change it printed four lines in E-R3) | 0 | PASS (gating) |
| D1 | `bash scripts/docs-i18n-check.sh` (redirected to `/private/tmp/t1718-run/close-i18n.txt`) | tail: `=== Summary ===` / `Errors:   0` / `Warnings: 0` / `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.` | 0 | observation |
| D2 | `hugo --minify --gc --source docs-site --destination /private/tmp/t1718-run/public-close` (redirected to `/private/tmp/t1718-run/close-hugo.txt`) | 18 lines; `grep -c -E 'WARN\|ERROR'` over them printed `0`; table `Pages │ 185 │ 183 │ 183 │ 183` (BASE: `187 │ 185 │ 185 │ 185`); `Total in 8552 ms` | 0 | observation |

The gitignored `docs-site/.hugo_build.lock` left by the build was deleted afterwards; `git status --short` then listed only this progress file.

Observed end state against the expected one: `swept: redirects=35 pages=608` (as expected); `shadowed: 0` (as expected); `resolve` with the admitted URL `unresolved: 0` and `admitted external: 5` (as expected); without the admitted arguments the same tree prints `unresolved: 5` and exits 1 by design (observed at M0 on the `good` scratch copy, not re-run on the repository tree); i18n `Errors: 0` / `Warnings: 0`; Hugo exit 0 with no `WARN` or `ERROR` line. No figure differed from the expectation.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-19T00:12Z
run_commit_sha: dc5b1f2f42f1ffbcb56bf961f9d0af5c38e6abab (closing change commit M2; M1 is 15b596315b0088fa422c4b4c645f8952e9326854; the evidence-record commit that carries this section follows M2 and is named in the completion report)
base_commit: 56873b261f900fa363b681cbe41d3e37e807f78f
ac_pass_count: 3 gating (AC-PRC-011 D1-D6, AC-PRC-010 D1-D4, AC-REDIR-003 D0)
ac_fail_count: 0
guards_observed_not_counted: AC-REDIR-001 D1-D3, AC-REDIR-002 D1-D4, AC-REDIR-003 D1 and D2 (all held; recorded in §E.2)
preserve_list_post_run_count: not measured (no PRESERVE list in this SPEC; the AC-REDIR-002 D1 scope check printed nothing)
l44_pre_commit_fetch: not run (no `git fetch` was issued in this run)
l44_post_push_fetch: not run (nothing was pushed; the lead batch-pushes)
new_warnings_or_lints_introduced: not measured with a linter (no lint run; Go lint is outside this docs-only change); i18n `Warnings: 0`, Hugo `WARN|ERROR` count 0
cross_platform_build: not applicable (no Go code changed; no `go build` was run per the lane rule)
total_run_phase_files: 5 (docs-site/vercel.json and the four docs-site/content/{ko,en,ja,zh}/changelog/_index.md deletions) plus this SPEC's progress.md
m1_to_mN_commit_strategy: M1 and M2 as separate commits, then one evidence-record commit; M0 and M3 edit no repository file other than progress.md
spec_status_transition: not performed (the run instruction forbids editing spec.md; `status: draft -> in-progress` is left to the orchestrator, see the completion report Gaps)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-19T00:17Z
sync_commit_sha: c8e10fa2c40975f1b31cec9367a1827460c1fe3d
base_head_at_sync_start: b58a46ac9653d35c1876ab12ec2c30b62cdfa6b2 (branch WT-id-spec-v3r6, tree clean)
changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Fixed`, first bullet (SPEC-V3R6-DOCS-PREREWRITE-REDIR-001); the sync-phase decision was that a docs-site redirect and page change gets an entry because the project records each SPEC close in [Unreleased] and the change alters what adk.mo.ai.kr serves.
b12_self_test_a: pre-emission `grep -c 'SPEC-V3R6-DOCS-PREREWRITE-REDIR-001' CHANGELOG.md` printed 0 before the edit (1 after it, the entry itself).
b12_self_test_b: ac_source=.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-REDIR-001/acceptance.md, tier M, resolved (non-empty). The B12 counter printed 6 (stderr `live=6 excluded=0 ambiguous=0`, exit 0). One of the six is `AC-REDIR-00n`, a prose suffix placeholder on acceptance.md line 23 ("AC-REDIR-00n verifies REQ-REDIR-00n"), matched by the identifier pattern because of its trailing `n`; it declares no criterion. The five declared criteria are the `### AC-` headings at lines 29, 48, 64, 75, 87 (`grep -c '^### AC-'` printed 5): AC-PRC-011, AC-PRC-010, AC-REDIR-001, AC-REDIR-002, AC-REDIR-003. The CHANGELOG entry states 5. The counter was not softened and acceptance.md was not edited (body is manager-spec's); marking line 23 `[REF]` would make the counter print 5.
b12_self_test_c: every path named in the entry exists or is verified absent as intended: `ls docs-site/vercel.json` and `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-REDIR-001/spec.md` present; `ls docs-site/content/{ko,en,ja,zh}/changelog` reported "No such file or directory" for all four (the deleted pages); `git cat-file -t 15b596315b` and `dc5b1f2f42` each printed `commit`.
frontmatter_status_transitions: spec.md `status: in-progress -> completed` on this single sync commit (the implemented step is merged into the same commit per the 3-phase close); `updated:` already read 2026-10-19 (the sync date) and was left as is. plan.md, acceptance.md, and decision-index.md carry no frontmatter block, so no `updated:` refresh applies. No other spec.md line changed.
lint: `moai spec lint .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 --strict` printed `0 error(s), 0 warning(s)` and exited 0, with one INFO `OwnershipTransitionUnmeasured` (the draft to in-progress transition commit b58a46ac96 has no Authored-By-Agent trailer). Judging build: `moai version` printed `moai-adk moai_cp/20261007_195927`; commit 7705b365e2 (named by the moai MCP server banner) is an ancestor of the tree HEAD (`git merge-base --is-ancestor 7705b365e2 HEAD` exit 0). That the installed CLI build is exactly that commit was not read from `moai version` in this run.
canary_compliance_check: not applicable (this SPEC defines no forward-looking policy that its own sync tests).
scope_findings: docs-site README/content/layout/data scan (`git grep` for changelog-page and `/agency` references outside vercel.json and SPEC/report paths) found no statement made false by the deleted pages or removed rules; the one hit is `.claude/agents/harness/hns-release-specialist.md` line 195, a stale pointer to the four deleted pages that the SPEC already names under Out of Scope with an owner; it was not edited here.
not_run: Hugo build, `scripts/docs-i18n-check.sh`, and the redirect-check instrument were not re-run in the sync phase; their results are the run-phase observations in §E.2 (M3) against HEAD dc5b1f2f42. No Go build or test ran (no Go code changed).

## §G Override and Refusal Record

- 2026-10-18T20:31:31Z SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter2.md
- 2026-10-18T20:33:31Z SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter2.md
- 2026-10-18T20:33:34Z SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter2.md
- 2026-10-18T20:33:41Z SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1718/.moai/reports/t1718/plan-audit-iter2.md
