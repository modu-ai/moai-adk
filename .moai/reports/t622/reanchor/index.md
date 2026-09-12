# t622 — SPEC 0.2.5 re-anchor evidence

Tree: worktree `.claude/worktrees/t622`, branch `WT-git-procedure-fixes`, HEAD `b24f2e184` (measurements taken before the 0.2.5 SPEC commit).
BASE = `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089` (merge absorbing local develop `f1f034bb4`).

| # | Command (run from worktree root; literal values) | Observed | File |
|---|---|---|---|
| 1 | `git diff --stat 255f88eb0… HEAD` | exit 0; 2 files: `.moai/reports/t622/absorb-ee99507fb.md`, `absorb2-mirror-baseline.txt` | `base-vs-head-stat.txt` |
| 2 | `git diff --stat 255f88eb0… -- internal/template/templates/ .claude/ .agents/ Makefile docs-site/content` | empty (`test -e` 0, `test -s` 1) | `base-vs-worktree-scope-stat.txt` |
| 3 | same form on `b412f8a33..255f88eb0` (positive control) | `113 files changed, 3636 insertions(+), 635 deletions(-)` | `control-b412-to-base-scope-stat.txt` |
| 4 | `git diff --stat b412f8a33 255f88eb0 -- <8 test files, commandemit, agentemit, Makefile, defaults.go, shipped_key_inventory.yaml, zone-registry L/T, git-strategy.yaml.tmpl, toml, published L/T, docs-site/content, plan-auditor.md>` | only `plan-auditor.md`, `Makefile`, docs-site `moai-web-console.md` ×4 | `support-files-b412-to-base-stat.txt` |
| 5 | `/usr/bin/grep -n -E 'merge_method\|--squash\|--auto-merge\|^## \|…' <manager-git.md local / template>` | local = template + 2 from line 6 (32→34, 114→116, 148→150, 156→158, 164→166, 166→168) | `lines-mg-local.txt`, `lines-mg-template.txt` |
| 6 | same on delivery.md | local 355/362-363/368/373-374/380/381/421/429 vs template 330/337-338/343/348-349/355/356/396/404 | `lines-delivery-*.txt` |
| 7 | doc-execution, SKILL, reference, qgc, workflows/sync.md, command sources | only workflows/sync.md differs (95/114 vs 85/104) | `lines-doc-execution.txt`, `lines-flag-surfaces.txt` |
| 8 | acp both copies (`[ZONE:Frozen]`, headings, block lines) | identical: 13, 17, 52, 290, 296-297, 299-312 (fetch 301, status 302, rev-list 307, session 311), 341, 360, 375; `ZONE:Evolvable` 16 each | `lines-acp.txt` |
| 9 | `diff <local> <template>` for nine pairs + published | exit 0: acp, published; exit 1: the other eight; hunk headers | `pair-*.diff`, `pair-hunks.txt` |
| 10 | AC-013 (b) mutant: local-only edit of manager-git line 116 vs both-copies edit | one-copy body compare exit 1; both-copies exit 0 | `mut013/result.txt` |
| 11 | `git show b412f8a33…:<acp>`, `git show 255f88eb0…:<acp local/template>` then block extraction awk | old block 9 lines; BASE blocks 12 lines, identical | `acp-*.md`, `ac002-acp-*-block.md` |
| 12 | AC-002 order judge (acceptance.md §D.1) on current blocks and mutants i-iv | current local/template `verdict=PASS`; i, ii, iii, iv `verdict=FAIL` | `ac002-judge-results.txt`, `ac002-mut-*.md` |
| 13 | interpretation-table rows b412 vs BASE, BASE L vs T; session-list count | 8/8/8 rows, diff exit 0 twice; session count 1 in all three blocks | `ac002-acp-*-matrix.txt` |
| 14 | Pre-Edit section b412 vs BASE; Pre-Spawn section b412 vs BASE (AC-003 control) | exit 0 (35 lines each); exit 1 | `ac003-b412-vs-base.diff`, `ac003-control-section.diff` |
| 15 | `git log --no-merges --format=%H 255f88eb0…..HEAD -- <acp local> <acp template>` (AC-016 judge) | `test -e` 0, `test -s` 1 (empty) | `ac016-judge-at-head.txt` |
| 16 | same on `b412f8a33..255f88eb0` (AC-016 positive control) | 2 lines: `97ef8e302…` (dr0911), `6896eef37…` (t635) | `ac016-control-b412-to-base.txt` |
| 17 | `/usr/bin/grep -rn 'func TestHookWrapperCopiesStayIdentical' internal` | `internal/hook/wrapper_copies_contract_test.go:73` — reads six hook wrapper scripts only; excluded | `hookwrapper-locate.txt` |
| 18 | `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity\|TestRuleTemplateMirrorDrift' -v` | exit 1; sets identical to `../absorb2-mirror-baseline.txt` (diff exit 0); 19 lines = PASS 17 + FAIL 2 | `mirror-baseline-rerun.txt`, `.exit`, `mirror-baseline-{sets,fail,pass}.txt` |
| 19 | AC-013 (e) set comparison on fixtures | baseline: new-fail empty, lost-pass empty; PASS→FAIL fixture: both non-empty; dropped-PASS fixture: lost-pass non-empty. Judge self-test on the rerun output: PASS count 17, new-fail `test -s` 1, lost-pass `test -s` 1 | `mutmirror/result.txt`, `mutmirror/selftest-result.txt` |
| 20 | `go test ./internal/template/ -run '^(TestSanitizedPairParity\|TestTemplateNoInternalContentLeak)$' -v -count=1` | exit 0, top-level PASS lines 2 | `ac013-gotest-at-base.txt`, `.exit` |
| 21 | AC-001/004/005/006 controls on BASE local exports | AC-001 paras 1, autofail `test -s` 0; delivery squash 2; `--squash` lines mg 34/116, delivery 368/380, toml 26/108; dl hits 8, 20; de hit 8 | `base-*.md`, `ac001-base-*`, `ac006-base-*` |
| 22 | AC-026 fragments, both copies at BASE | nine fragments L vs T diff exit 0; line counts 1/3/4/6/1/52/1/1/24; (ii) template hits skill 1, ref 2, qgc-args 4, qgc-flags 4, sync 1, dl 9·20, usage 1, hint 1, next 9; `--no-merge` dl 8·19 | `frag/` |
| 23 | PR Auto-Merge section local vs template | 9 lines each, diff exit 0 | `mg-pram-*.md` |
| 24 | Frozen: `[ZONE:` counts, registry, four clause counts at BASE | other eight files 0 (L/T); registry acp 13/13, others 0/0; clauses 1 each copy | `zone-lines-others.txt`, `registry-others.txt`, `clause-counts.txt` |
| 25 | `git grep -c`/`-n` X4 docs-site at BASE | unchanged from 0.2.2 (en 5, ko 4, ja 6, zh 9; what-is 1 each) | `x4-docs-*.txt` |
| 26 | Cf counter on the four SPEC files before edit; control with 2 planted Cf (`U+200B`, `U+FEFF`) | 0 / 0 / 0 / 0; control 2 | `cf-before.txt`, `cf-control.txt` |
