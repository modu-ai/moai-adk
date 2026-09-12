# Run-phase report — card t622, SPEC-GIT-DELIVERY-PROCEDURE-001 (Tier M, 0.2.6)

Author: manager-develop (run-phase part 2, cycle_type ddd). Part 1 (pre-flight, M1, X2, M2, M3)
is summarised in `preflight-summary.md`, `ac030-outcome.md` and progress.md §E.2; this report
covers part 2 (M4, M5, M6) and re-runs every judged criterion on the final tree.

All paths below are relative to `.moai/reports/t622/run/` unless they start with a repository
root directory. `T` = `internal/template/templates`.

## 0. Tree and baseline attribution

| Item | Value | Source |
|---|---|---|
| worktree / branch | `.claude/worktrees/t622`, `WT-git-procedure-fixes` | `git branch --show-current` re-read before every commit |
| part-2 start HEAD | `7a02b90e25936ddef7fe38a02496fc80faf964cd` | `git rev-parse HEAD` |
| M6 judging HEAD | `35c0e30dfcc945fa9e43405d81dd570f9dde7653` | `git rev-parse HEAD` before the M6 batch |
| evidence HEAD after M6 | `8d1b8c9206b6c95350302f090374f822101221f2` | `git rev-parse HEAD` |
| BASE (fixed snapshot) | `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089` | acceptance.md |
| CARD_BASE (part 2) | `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf`, 1 line — unchanged from part 1 | `git merge-base --all develop HEAD > card-base.txt` |
| local develop | `85868148c30f39d654694f61d1424b2bb5a4f456`, 25 commits past CARD_BASE (`git rev-list --count`) — moved since part 1, **not absorbed, not merged** (recorded only) | `git rev-parse develop` |
| range control | `git diff --name-only develop...HEAD > card-range-names.txt` → 760 lines, `test -e` 0, `test -s` 0 | |
| snapshot freshness | `snapshot-stale.txt` `test -e` 0 / `test -s` 1; `snapshot-stale-mirror.txt` `test -e` 0 / `test -s` 1 | both empty → BASE exports valid, no re-export (Caution N1 not triggered) |
| scope content identity | `git diff --name-only 5708e04d2 HEAD -- . ':!.moai/reports/t622/run/'` → empty (exit 0) | every judge run after the `.toml` commit read the same scope content (`p2-evidence-only-since-toml.txt` lists the evidence-only changes) |

The judging tool is the repository's own Makefile/`go test` targets, built from this tree by
`go test` at invocation (no installed `moai` binary was used as a judge).

## 1. E1 — AC matrix (16 judged criteria)

"Re-run" marks rows first judged in part 1 and re-run here on the final tree. Rows judged first in
part 2 ran either at M4 (`5708e04d2`/`8a115e0f5` content) and again at M6, or at M5 (`8a115e0f5`);
the scope content is identical across all of these HEADs (§0 last row). All git commands below
exited 0 unless a different exit is written; `test` / `grep` / `diff` exits are written explicitly.

| AC | Verdict | Command(s) (literal, from the worktree root; `E` = `.moai/reports/t622/run`) | Output file(s) | Exit / observed |
|---|---|---|---|---|
| 001 | PASS (re-run L·T; `.toml` new) | `sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' <mg copy>`; paragraph awk; autofail awk; listgroup awk (acceptance.md verbatim) | `p2-ac001-{local,template}-*`, `ac001-toml-{sync.md,paras.txt,paras.md,autofail.md,listgroup.md}`, `ac001-reading.md` (+ M4 `.toml` addendum) | paras 1·1·1; autofail `test -e` 0 / `test -s` 1 for L, T, `.toml`; listgroup `test -s` 1 ×3; L·T paragraphs `cmp` 0 vs the read targets; `.toml` paragraph `cmp` 0 vs template paragraph; control on BASE autofail `test -s` 0 |
| 002 | PASS | block awk + order-judge awk; `sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p'`; `diff` vs BASE sections; matrix `grep -E '^[|] '` + `diff` | `ac002-{local,template}-{block.md,order.txt,order-pass.txt,session.txt,section.md,section.diff}`, `ac002-matrix.diff`, `ac002-template-matrix.diff`, `p2-ac002-mut-i-*` | (a) both `fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS`, order-pass 1·1, session 1·1; (b) section diff exit 0·0 (52 lines each), matrix diff exit 0·0 (8 rows); control b412 block `order=FAIL` |
| 003 | PASS | `sed -n '/^### Pre-Edit Sync Check/,/^#### The sweep prohibition/p'` on BASE and both copies; `diff` | `ac003-base.md`, `ac003-base-template.md`, `ac003-{local,template}.md`, `ac003.diff`, `ac003-template.diff`, `p2-ac003-control.diff` | diff exit 0·0 (35 lines each); control (b412 vs BASE Pre-Spawn section) exit 1 |
| 004 | PASS (re-run) | `grep -c 'gh pr merge --squash --delete-branch'`; `grep -c 'gh pr merge --<merge_method> --delete-branch'`; `grep -n -E 'merge_method.*(squash\|default)'` | `p2-ac004-{local,template}-{squash,resolved,source}.txt` | squash 0·0; resolved 2·2; source grep exit 0·0 (local 377, template 352: the resolution sentence); BASE control 2 |
| 005 | PASS (re-run; `.toml` new) | `grep -c -E 'gh pr merge[^\|]*--squash'` per file; default-sentence and example `grep -c -F` | `p2-ac005-{local,template}-mg*.txt`, `p2-ac005-{local,template}-others.txt`, `ac005-toml.txt`, `ac005-toml-lines.txt`, `p2-ac005-toml.txt` | mg 1·1 (default sentence), default 1·1, example 1·1; other eight files 0 in both copies; `.toml` 1 = line 26, the default-explanation sentence |
| 006 | PASS (re-run) | Step 3.4 / Worktree Context Detection awk extraction; extended detector awk; `grep -n -i -E` de detector; `grep -c 'manager-[g]it[.]md'`; `grep -c -e '--auto-merge'` | `p2-{local,template}-{dl,de}.md`, `p2-{local,template}-006-*`, `p2-ac006-optin.txt`, `p2-ac006-dl-m1-vs-final*.diff`, `ac006-reading.md` (+ final-tree addendum) | dl-default `test -e` 0 / `test -s` 1 ×2; de-default grep exit 1 ×2; source dl 2 / de 1 ×2; opt-in 4·4; BASE control lines 8·20 |
| 013 | PASS | (a) `diff` acp pair / published pair; (b) `diff base-X base-X-template`, `diff <L> <T>`, header-strip `grep -v -E`, body `diff` ×8; (c) `grep -E '^argument-hint:'` + `diff`; (d) `go test ./internal/template/ -run '^(TestSanitizedPairParity\|TestTemplateNoInternalContentLeak)$' -v -count=1`; (e) mirror run + set comparison | (a) `ac013-acp.diff`, `ac013-published.diff`; (b) `ac013-*-base.{diff,body}`, `ac013-*-post.{diff,body}`, `ac013-*-body.diff` (M4: `m4-ac013-*`); (c) `ac013-hint*.txt/.diff`; (d) `ac013-gotest.txt`, `.exit`, `ac013-pass-count.txt`, `ac013-acp-subtest.txt`; (e) `mirror-m6*`, `mirror-new-fail.txt`, `mirror-lost-pass.txt` (M4: `mirror-m4*`, `mirror-*-m4.txt`) | (a) exit 0·0; (b) all eight base bodies non-empty, body diff exit 0 ×8 (M4 and M6); body-comparator control exit 1; (c) hint diff exit 0; (d) exit 0, top-level PASS 2, acp trace 6, empty-sweep tokens 0; (e) see §3 |
| 014 | PASS | `make agents-emit-check` → `make agents-emit` → `make agents-emit-check`; `git diff --name-only develop...HEAD -- $T/.codex/agents/moai/`; example `grep -c -F`; section `sed`/`awk` + `diff` | `ac014-red.{txt,exit}`, `ac014-emit.{txt,exit}`, `ac014-green.{txt,exit}`, `ac014-changed.txt`, `ac014-toml-example.txt`, `ac014-{md,toml}-{sync,pram}.md`, `ac014-sync.diff`, `ac014-pram.diff`, `ac014-toml-regen.diff`, `m6-agents-emit-check.*` | RED make exit 2 (recipe `Error 1`, `manager-git.toml` sha256 mismatch); emit exit 0; GREEN exit 0; changed = only `manager-git.toml`; example 1; sync diff exit 0 (11 lines), pram diff exit 0 (13 lines); M6 re-check exit 0 |
| 015 | PASS | `git diff develop...HEAD -- $T/`; added-line count; SPEC/REQ/date/`CLAUDE.local` greps; hex-word perl | `ac015-template.diff`, `ac015-files.txt`, `ac015-added-count.txt`, `ac015-{specid,req,date,local-ref}.txt`, `ac015-hex-tokens.txt`, `ac015-sha-{letter,digits}.txt`, `ac015-reading.md`, `p2-ac015-fixture-tokens.txt` | diff `test -s` 0; 38 added lines over 9 template files; four greps exit 1; sha-letter `test -e` 0 / `test -s` 1; hex tokens 0; fixture control 5 letter / 2 digit |
| 016 | PASS (SHOULD) | `git log --no-merges --format=%H HEAD --not develop -- <both acp paths>`; control `git log --no-merges --format=%H 255f88eb0… --not b412f8a33… -- <both acp paths>` | `ac016-acp-commits.txt`, `p2-ac016-control.txt`, `p2-ac016-control-count.txt`, `p2-ac016-card-commits-all.txt` | judge `test -e` 0 / `test -s` 1 over 34 card non-merge commits; control 2 commits (`97ef8e302…`, `6896eef37…`), `cmp` 0 vs pre-flight control |
| 025 | PASS | scope `git diff develop...HEAD -- <18 scope paths>`; `grep -F '[ZONE:Frozen]'` then `grep -n -E '^[-+].*\[ZONE:Frozen\]'`; four clause `grep -c -F`; registry `grep -c -E` ×2 | `ac025-scope.diff` (M4 `m4-ac025-*`, `cmp` 0), `ac025-frozen-{any,lines}.txt`, `ac025-const0{06,36,37,38}.txt`, `ac025-registry-{control,others}.txt`, `m4-ac025-zone-*.txt`, `p2-ac025-*` | scope `test -s` 0; Frozen-any grep exit 1, Frozen-lines exit 1 (`test -e` 0); clauses 1·1 for all four IDs in both copies; `[ZONE:Frozen]` at acp line 17 in both, 0 in the other scope files; registry 13·13 / 0·0; controls: Frozen-line fixture 2, registry mutant 2 |
| 026 | PASS (re-run) | nine fragment extractions per copy; (i) `grep -c -e '--auto-merge'`; (ii) direction awk; (iii) presence `grep -c -E`; merge-lines awk | `p2-{local,template}-{skill,ref,qgc-args,qgc-flags,sync,dl,sync-usage,hint,dl-next}.md`, `p2-{local,template}-026-*`, `ac026-reading.md`, `p2-ac026-base-control-ii.txt` | 18 fragments `test -s` 0, all `cmp` 0 vs part-1 final fragments; (i) every fragment ≥1 (qgc 1+2); (ii) `test -e` 0 / `test -s` 1 ×2; (iii) 1,1,1,2 ×2; merge lines 5 per copy, bodies `cmp` 0 vs read targets; BASE control 10 violations |
| 027 | PASS (re-run) | `grep -c -e '--no-merge'` on dl; (i)/(ii) awk over the nine fragments | `p2-{local,template}-027-{dl-count,i,ii,lines}.txt`, `p2-ac027-base-control-i.txt` | dl `--no-merge` lines 1 (dl:29) ×2, no other fragment; (i) and (ii) `test -e` 0 / `test -s` 1 ×2; BASE control lines 8·19 |
| 028 | PASS (re-run) | PR Auto-Merge awk extraction; (a)/(b)/mode-lines awk | `p2-{local,template}-mg.md`, `p2-{local,template}-028-*`, `ac028-reading.md` | (a) mg 1 / dl 1 ×2; (b) `test -e` 0 / `test -s` 1 ×2; mode lines 14 ×2, bodies `cmp` 0 vs read targets; BASE control (a) 0 |
| 029 | PASS (re-run) | (a)/(b) awk | `p2-{local,template}-029-*`, `ac028-reading.md` | (a) mg 1 / dl 1 ×2; (b) `test -e` 0 / `test -s` 1 ×2; BASE control (a) 0 |
| 030 | PASS — case (A) (re-run post-commit judges) | `git diff --name-only develop...HEAD -- $T/.agents/skills/ .agents/skills/`; two `git log --format=%H HEAD --not develop -- …`; path-form control `git diff --name-only …`; `git ls-files`; orphan `grep -v -x -F -f`; published L·T `diff`; flag awk; `make commands-emit-check` (M4, M6) | `p2-ac030-*`, `ac030-outcome.md` (part 1 control (c) + emit/check), `m4-commands-emit-check.*`, `m6-commands-emit-check.*` | changed `test -e` 0 / `test -s` 1; src commits `2f4dfd803b1b7dace279fea7583dad04e6637a23`; artifact commits `test -s` 1; path control 2; tracked 2; orphan `test -s` 1; published diff exit 0; flags `test -s` 1; emit checks exit 0·0 |

Reading records present and non-empty: `ac001-reading.md` (M4 addendum for `.toml`),
`ac006-reading.md` (final-tree addendum for the three M2 lines), `ac026-reading.md`,
`ac028-reading.md` — their read targets are byte-identical to the final-tree targets (`cmp` 0
rows above). `ac015-reading.md` is new.

## 2. Generator checks

| Check | Command | Exit | Output |
|---|---|---|---|
| RED (before regeneration) | `make agents-emit-check > ac014-red.txt 2>&1` | 2 (recipe `Error 1`; `golden_test.go:109: .codex/agents/moai/manager-git.toml: committed artifact differs from emission (sha256 mismatch)`) | `ac014-red.txt`, `.exit` — committed alone in `cc51d8479`, ahead of the artifact |
| regenerate | `make agents-emit > ac014-emit.txt 2>&1` | 0 | `ac014-emit.txt` |
| GREEN | `make agents-emit-check > ac014-green.txt 2>&1` | 0 | `ac014-green.txt` |
| M4 commands check | `make commands-emit-check > m4-commands-emit-check.txt 2>&1` | 0 | `m4-commands-emit-check.txt` |
| M6 agents check | `make agents-emit-check > m6-agents-emit-check.txt 2>&1` | 0 | `m6-agents-emit-check.txt` |
| M6 commands check | `make commands-emit-check > m6-commands-emit-check.txt 2>&1` | 0 | `m6-commands-emit-check.txt` |

`make` returns 2 for a failed recipe; the recipe's own status is the `Error 1` on the make line
(the same convention part 1 recorded for AC-GDP-030's control). acceptance.md writes the RED
expectation as "exit 1"; the drift was detected, which is what the RED step exists to show.

`.toml` diff (`ac014-toml-regen.diff`, 4 hunks): line 108 example `--squash` → `--<merge_method>`;
line 142 `per § PR Auto-Merge (Team Mode)` → `per § PR Auto-Merge`; the Synchronization paragraph
rewritten to fetch-first-then-rev-list; `## PR Auto-Merge (Team Mode)` → `## PR Auto-Merge` with
the opt-in sentence (`--merge` deprecated alias), the team-mode all-approvals line, the
personal/manual no-approval line, and the "Steps (all modes…)" lead-in. It is the machine emission
of the template `manager-git.md`; it was not hand-edited.

## 3. Mirror non-regression (AC-GDP-013 (e))

Command (both runs): `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v`.
Baseline: `.moai/reports/t622/reanchor/mirror-baseline-{fail,pass}.txt` (PASS 17, FAIL
{`TestRuleTemplateMirrorDrift`, `TestRuleTemplateMirrorDrift/spec-workflow.md`}).

| Run | go test exit | PASS count | FAIL set | new-FAIL file | lost-PASS file | sorted sets vs baseline |
|---|---|---|---|---|---|---|
| M4 (`mirror-m4.txt`) | 1 (recorded only — baseline is exit 1) | 17 | baseline pair only | `mirror-new-fail-m4.txt` `test -e` 0 / `test -s` 1 | `mirror-lost-pass-m4.txt` `test -e` 0 / `test -s` 1 | `diff` exit 0 |
| M6 (`mirror-m6.txt`) | 1 | 17 | baseline pair only | `mirror-new-fail.txt` `test -e` 0 / `test -s` 1 | `mirror-lost-pass.txt` `test -e` 0 / `test -s` 1 | `diff` exit 0 |

Comparator control: `p2-mirror-mutant-*` (M4 set with one PASS flipped to FAIL) → new-FAIL 1 line,
lost-PASS 1 line. The remaining FAIL (`spec-workflow.md`) is out of scope and pre-existing.

## 4. AC-GDP-016 judge and positive control

```
$ git log --no-merges --format=%H 255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089 --not b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0 -- .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md > p2-ac016-control.txt
97ef8e3023e9e7a29e7478289b69d28796dddbd7
6896eef3766a5265ac32b21154e95907e1173b54          (count 2; cmp 0 vs pre-flight ac016-control.txt)

$ git log --no-merges --format=%H HEAD --not develop -- .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md > ac016-acp-commits.txt
(empty)   test -e 0, test -s 1 — over 34 card non-merge commits (p2-ac016-card-commits-all.txt), range control 760 names
```

## 5. Commits, Cf counts, gaps, residual risks

Card commits since `fa13c27b6` (oldest first, `git log --no-merges fa13c27b6..HEAD`):

| SHA | Subject |
|---|---|
| `0da3bebf0` | pre-flight baseline before edits |
| `3f6c5163f` | M1 single opt-in with mode approval rules |
| `2f4dfd803` | X2 command-source `argument-hint` (AC-GDP-030 case A) |
| `af54bf1ff` | M2 resolve the merge method |
| `7a02b90e2` | M3 fetch before rev-list in manager-git sync |
| `cc51d8479` | M4 agents-emit-check RED evidence (before regeneration) |
| `5708e04d2` | M4 regenerated `manager-git.toml` |
| `8a115e0f5` | M4 evidence |
| `35c0e30df` | M5 evidence (no edits) |
| `8d1b8c920` | M6 evidence (no edits) |
| (this commit) | run report + progress.md §E.2/§E.3 |

Cf (`perl -CSD -ne '$c+=()=/\p{Cf}/g; END{print $c+0,"\n"}'`): regenerated `.toml` 0
(`p2-cf-toml.count`), BASE `.toml` 0, edited reading records (ac001/ac006/ac015) 0
(`p2-cf-reading-records.count`), counter proven on the 2-planted control = 2
(`p2-cf-control-2.count`). progress.md and this report: see the final-commit Cf file
`p2-cf-final-docs.count`.

Gaps (not observed in this run):
- CI (`origin/develop` full suite, darwin/windows matrix, `template-neutrality-check`): the branch is
  not pushed; lanes do not push. The verdict surface is the lead's develop push.
- Local develop moved 25 commits past CARD_BASE; this run did not absorb it. All range-based judges
  (014, 015, 016, 025, 030) are pre-merge judgements against the current merge-base; they must be
  re-measured after the integration-window absorb.
- `go test ./...` was not run (verification-load rule); only the targeted `internal/template` tests
  and the two make checks ran.
- `make build` / `make embed-check` were not run; no claim is made about an installed binary's
  embedded `.toml`.
- Part-1 control (c) for AC-GDP-030 (planted published drift) was not re-planted; its part-1
  record stands, and the M4/M6 `commands-emit-check` exits are the final-tree observation.

Residual risks:
- AC-GDP-002 is a regression guard; mutant (ix) (an order-reversing sentence outside the Pre-Spawn
  section) is invisible to it by design — AC-GDP-016 and the freshness check cover that shape.
- The line-level detectors of AC-GDP-026-029 can miss phrasings they were not written for; the
  reading records carry those criteria, and their targets were confirmed byte-identical.
- The residual noted by part 1 stands: doc-execution.md line 34 still says `is_worktree_context` is
  stored "for use in Phase 13" (unchanged; it does not tie the flag to merging).
- `make` exits 2 where acceptance.md writes "exit 1" for the RED steps (AC-GDP-014, AC-GDP-030
  control) — a documentation nuance of the acceptance text, not a detector failure.
