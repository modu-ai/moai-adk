# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.2.4 — pre-kickoff application of optional findings N1-N3 from reduced plan-audit iteration 3 (`.moai/reports/t622/plan-audit-reduced-iter3.md`) per operator decision via lead: Implementation Kickoff approved on condition these are applied first; no re-audit follows; requirements, decisions, scope, judged REQ 12 and judged AC 16 unchanged. Revision 0.2.3 — targeted fix of reduced plan-audit iteration 2 findings D1-D8 (`.moai/reports/t622/plan-audit-reduced-iter2.md`, FAIL 0.75) per operator decision via lead: detector strings and wording only; requirements, decisions, scope, judged REQ 12 and judged AC 16 unchanged. Revision 0.2.2 — applies the lead's consumer-scope ruling X1-X4 (2026-09-11). 0.2.1 was committed as `caa601d7c`. An earlier 0.2.2 attempt stopped on a session limit and its partial edits were preserved as WIP commit `ea09ca650`; this revision completes it from that commit.
- **Revision note (0.2.6) — limited re-anchor audit FAIL 0.75 repaired** (`.moai/reports/t622/plan-audit-reanchor.md`, D1-D6). Requirement count, scope, judged REQ 12 / AC 16, and the kickoff approval are unchanged. Evidence: `.moai/reports/t622/reanchor-fix/` (`index.md` lists commands and observed values).
  - D1 (blocking): every "what did this card change / commit" judge now takes its left end at read time — diffs `git diff develop...HEAD -- <paths>` (= `git diff $CARD_BASE HEAD`, `--name-only` outputs `cmp` exit 0), commit lists `git log --no-merges --format=%H HEAD --not develop -- <paths>`, range control `git diff --name-only develop...HEAD` ≥1 line else "unmeasurable". Rewritten: AC-GDP-014, 015, 016, 030, and AC-GDP-025 (same literal form, found by sweep); plan §C step 1 and M5. Pre-merge only (post-merge evidence = merge-tree identity). `$BASE` kept only for fixed snapshots (exports, `b412f8a33`-range positive controls, parity hunks, mirror baseline, AC-GDP-002/003 reference sections); a pre-judge snapshot-staleness check (`git diff --name-only $BASE $CARD_BASE -- <21 snapshot paths>` empty) was added. The 0.2.5 "no re-absorb after BASE" premise and the "move BASE to the absorb merge" remedy were removed. Reproduced on the current tree (HEAD `f2fa64e08`, merge-base `f1f034bb4`) and on an emulated re-absorb — dangling object `7fae4764e` from `git merge-tree --write-tree HEAD develop` + `git commit-tree` (no ref moved; develop and HEAD re-read unchanged): card commit list 23 = 23 (`diff` exit 0); AC-016 judge `test -e` 0 / `test -s` 1 at both; literal `$BASE..` range grew 3 → 12 with 9 develop commits; literal template diff on the emulated merge carried develop's 6 template files / 146 added lines while `develop...` stayed empty; moving BASE to the absorb merge empties the range (0 lines). Positive controls: `255f88eb0 --not b412f8a33` on the acp copies → 2 commits; `ee99507fb...97ef8e302` on `.codex/agents/moai/` → 2 files; card t645 (`0db675bed...c7874923a`) → its own 8 files vs 19 with its pre-absorb pinned base. Snapshot staleness: empty at `f1f034bb4` and at emulated `00ae57ad7`; `b412f8a33..BASE` control → 8 files; mirror-baseline staleness empty now, 8 files after the emulated absorb (includes the baseline FAIL pair `spec-workflow.md`).
  - D2 (blocking): AC-GDP-002 split into (a) ordering (fetch line anchored to end of line, `exit([[:space:]]|$)`, ordered one-line `;`/`&&` join accepted) and (b) whole Pre-Spawn section `diff` against the fixed reference (local + template). Current local/template blocks: `order=PASS shape=PASS`, section `diff` exit 0. Mutants i-viii: all AC-GDP-002 FAIL — (a) catches i, iii, iv, v, vi; (b) catches all eight. Retry for a passing violator: heading rename, duplicate section, `sh` fence, template-only edit all FAIL (b); an override sentence outside the section (ix) PASSES AC-GDP-002 — recorded as a limit, covered by AC-GDP-016 (card commit) and the staleness check (absorb). The judge text in acceptance.md is byte-identical to the tested program (`cmp` exit 0).
  - D3 (blocking): REQ-GDP-002 restored to its original substance (fetch completes before rev-list starts; not listed as independent parallel items; third-command concurrency and both tables unchanged); status observation + abort moved to a non-normative note carried by AC-GDP-002 (b). spec §A.2 Pre-Edit `;` row stays "정상" — consistent with REQ-GDP-002 and with AC-GDP-002 (a), which accepts the ordered join. The 0.2.5 HISTORY row carries a bracketed correction.
  - D4: AC-GDP-002 MUST-PASS justified as a `verification-completeness.md` §4 tree-pinned preservation assertion (like AC-GDP-003). D5: plan §C step 8 sorts both sets before `diff` (reproduced: unsorted `diff` exit 1, sorted exit 0). D6: AC-GDP-016 GIVEN now names all card commits (plan commits included).
- **Revision note (0.2.5) — re-anchored after the develop absorbs.** The card absorbed local develop twice (merge `7ac8b8491` for develop `ee99507fb`, merge `255f88eb0` for develop `f1f034bb4`). Develop card t635 rewrote the `agent-common-protocol.md` Pre-Spawn block into the Lane A (ordered) / Lane B (independent) shape and card dr0911 mirrored it to the template (copies byte-identical). Lead decision (a), operator-routed, final: that shape satisfies REQ-GDP-002; this card makes no edit to either copy. Requirement substance, scope, judged REQ 12 / AC 16 and the kickoff approval are unchanged. Applied: BASE → `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089`; cited line numbers declared template-based with a local-line table (spec §A.1); local/template parity baseline re-measured (byte-identical: acp + published skill; intended differences: eight pairs); REQ-GDP-002 reworded to the landed property; AC-GDP-002 rewritten for Lane A/B as a regression guard (mutants i-iv FAIL, current blocks PASS); M5 verification-only; AC-GDP-016 rewritten to "no card commit touches either acp copy" (control `b412f8a33..BASE` → 2 commits); mirror-test non-regression check added to M4/M6/AC-GDP-013 (e) (baseline exit 1, 17 PASS lines, FAIL {`TestRuleTemplateMirrorDrift`, `TestRuleTemplateMirrorDrift/spec-workflow.md`}); local-line citations fixed. Files this card changes: 17 (19 in case B). The re-anchor evidence lives in `.moai/reports/t622/reanchor/` (`index.md` lists every command and observed value). A plan-auditor pass on the REQ-002/AC-002 delta and its knock-on items follows. The §F mode record below predates this revision and is left as the lane wrote it (its file count 19/21 was measured before acp dropped out of the edit set).
- **Revision note (0.2.4)** — acceptance.md only (plus this note and the spec.md HISTORY row):
  - N1: the command convention now prescribes literal values for `BASE`/`E`/`T` (`$BASE`/`$E`/`$T` in judge commands are notation to substitute) and forbids the `BASE=…; E=…; T=…; <command>` prefix form for git commands and commands naming `manager-git.md` (guard refusal observed in audit iteration 3: T2, T7, T4, L2). No judge command was rewritten — the audit names no judge unrunnable when its values are substituted literally (T3, T8, L1, V2-V5 ran).
  - N2: `test -e <file>` (expected exit 0; exit 1 = judge never ran = not PASS) placed before every expected-empty judge — 13 `test -s … exit 1` judges (AC-001 ×2, AC-002, AC-006, AC-015, AC-026, AC-027 ×2, AC-028, AC-029, AC-030 ×3), the AC-016 after-list ("빈 파일" → `test -s` exit 1), the AC-030 (A) expectations for `ac030-changed.txt` and `ac030-artifact-commits.txt`, and the AC-030 case selector `ac030-emit-status.txt`. Mutant (scratch): missing `ac028-b.txt` → `test -e` exit=1, `test -s` exit=1 — the old emptiness-only check would have read PASS.
  - N3: AC-028/029 reading-list selector gained `|| /--auto-merge/`. Mutant (scratch, 3 lines): old selector selects line 1 only; new selector selects lines 1 and 2 ("`--auto-merge` merges as soon as CI checks pass."); line 3 selected by neither.
- **Revision note (0.2.3)** — every fixture line, command and expected result is recorded in acceptance.md next to its criterion; the lines below are the observed outputs from this authoring run (scratch fixtures; base fragments extracted from the template copies on HEAD `0af445525`):
  - D1 AC-GDP-026: added (iii) `--merge` presence ≥ 1 in the SKILL Flags line, the `workflows/sync.md` `**Flags**:` line, the Supported Flags section and the delivery Step 3.4 section; (ii) made directional (the `--merge` line must carry "deprecated alias of/for `--auto-merge`" after the token, and must not say "not/no longer/never deprecated" or "un-deprecated"); reading record `$E/ac026-reading.md` added as a PASS precondition. Observed: 10-line fixture → (ii) hits 1, 2, 3, 5, 8, 9, 10 (auditor's reversed line "`--merge`: auto-merge the PR after sync (the older `--auto-merge` spelling is deprecated)" = line 2, hit), correct lines 4 and 6 not hit, `--merged-only` line 7 not hit; X1-X3 fixture → hits 1, 3, 5 (unchanged). Removal fixture (auditor's "Modes: auto, force, status, project. Flags: --auto-merge, --skip-mx" plus three fragments without `--merge`) → presence `0 0 0 0`. Positive control → presence `1 1 1 1`, (ii) file empty (`test -s` exit 1), (i) `1 1 1 1`. Base fragments → (ii) the same 10 lines as 0.2.2, presence skill 1 / sync 1 / qgc-flags 1 / dl 2.
  - D2 AC-GDP-028: (a) now requires "all" within two words before "approv"; (b) adds optional / not needed / unnecessary / at least one / a single / after·on·once·upon·with + any·one·an·a·some·the first approval / majority; reading record `$E/ac028-reading.md` (covers AC-028 and AC-029). Observed on the 15-line fixture: (a) hits 3, 15; (b) hits 2, 7, 8, 12, 13 — auditor lines 7 ("approvals are optional") and 8 ("at least one approval") hit (b) and miss (a). Base mg/dl sections: (a) 0, (b) 0.
  - D3 AC-GDP-029 (b): trigger widened to approval or review; exemptions widened (without requiring, no approval/review, not needed); a gating phrase (after/once/until/pending/require + approval/review) is a violation regardless of exemptions. Observed: (a) hits 4, 6, 9, 10, 11, 14; (b) hits 4, 11, 14 — auditor's correct lines 9 and 10 no longer hit, "after a code review" (11) hits. Base: (a) 0, (b) 0.
  - D4 AC-GDP-027 (ii): added disable / override / turns off / suppress / bypass / cancel. Observed on the 7-line fixture: (i) hits 1, 3; (ii) hits 1, 2, 3, 5, 6, 7; correct line 4 not hit. Base dl: (i) 8, 19; (ii) 8, 19.
  - D5: scope set recounted as nine path pairs (manager-git, agent-common-protocol, delivery, doc-execution, moai SKILL, reference, quality-gates-context, workflows/sync.md, command source) — "열 개" → "아홉 개", "나머지 아홉" (after excluding agent-common-protocol) → "나머지 여덟", §A.5 "위 아홉" → "위 여덟", §E.1 "뺀 아홉" → "뺀 여덟"; §C.4 and plan §A affected files 21 → 19 (21 in case B). Counting rule stated in spec §C.4 and acceptance conventions.
  - D6: spec §E.2 citation `caa601d7c` → `0af445525`; `git diff --stat b412f8a33 HEAD -- <scope L·T, published SKILL local, zone-registry local, internal/template, Makefile, docs-site/content>` at `0af445525` → no output (`test -s` exit 1).
  - D7: acceptance conventions now require `BASE`/`E`/`T` in the same invocation as each judge command.
  - D8 AC-GDP-016: anchor is the oldest commit touching either copy of `agent-common-protocol.md` (`git log --format=%H --reverse … -- <local> <template>`, first line), with an explicit literal SHA substitution step. Observed at `0af445525`: list empty, `test -s` exit 1 (expected at base). The ordering mutants are logic-only (no commits made in this run).
- **Revision note (0.2.2)**:
  - In scope, local and template, line-level: X1 `workflows/sync.md` usage line (local 95 / template 85), X2 `argument-hint` in `.claude/commands/moai/sync.md:3` and template `sync.md.tmpl:3`, X3 `delivery.md:404` "Auto-Merge PR (/moai sync --merge)" (same line number in both copies). All three verified by reading the lines on this tree.
  - X2 is a command source, so REQ-GDP-014 now also binds `make commands-emit` then `make commands-emit-check` (exit 0) with any regenerated published skill in the same commit as the source edit; new AC-GDP-030 judges both outcomes (A no change / B change) and carries a control showing the check can fail.
  - X4 (docs-site four locales) is out of scope; the follow-up docs card and its measured inputs are recorded in spec.md §D.
  - Kept from the WIP commit after re-measurement: scope set (nine path pairs; the 0.2.2 text called it "ten", corrected in 0.2.3), nine AC-GDP-026/027 fragments and their base controls, AC-GDP-013 parity rows for the command pair and published skill, AC-GDP-015/016/025 file lists, AC-GDP-030 body.
  - Corrected: AC-GDP-025 registry detector `manager-git\.md` → `manager-[g]it\.md` (the worktree guard refuses the literal word inside a pattern; the uncorrected command could not have run at run-phase). spec.md §C.5 content moved into §D (the ruling asks for it there). Base measurement HEAD updated to `ea09ca650`.
  - Added: empty-set guards on AC-GDP-015 (template diff must be non-empty), AC-GDP-016 (the always-loaded-rule commit must exist), AC-GDP-025 (scope diff must be non-empty; mutant for the "others" registry detector), AC-GDP-030 (tracked-path control, source-commit existence, mechanical same-commit subset check, published file non-empty).
- **Tier**: M kept per lead ruling (judged counts decide; file-count guidance is reference). Judged REQ 12 and AC 16 are within 16/16; affected files 19 (nine scope path pairs × local·template = 18, plus the generated `.toml`), +2 if AC-GDP-030 outcome B (21). 0.2.2 stated 21 with +2 by miscounting the scope list as ten. **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree re-measured for 0.2.3**: HEAD `0af445525` (parent `3512b9f75`; both change only SPEC/report files) — `git diff --stat b412f8a33 HEAD -- <scope files L·T, published moai-sync SKILL.md local, zone-registry.md local, internal/template, Makefile, docs-site/content>` → no output (`test -s` exit 1). **Tree measured (0.2.2)**: HEAD `ea09ca650` (parent `caa601d7c`; both change only SPEC files). `git diff --stat b412f8a33 ea09ca650 -- <scope files L·T, manager-git.toml, published moai-sync SKILL.md L·T, internal/template/commandemit, internal/template/agentemit, zone-registry.md L·T, docs-site/content, Makefile>` → no output; same for the eight guarding test files; `git diff --stat b412f8a33 -- internal/template/templates/` → no output. Plan-time counts on these copies are base counts.
- **Audit count**: reduced iteration 2 FAIL 0.75 (`.moai/reports/t622/plan-audit-reduced-iter2.md`, measured on `3512b9f75` = 0.2.2); this 0.2.3 revision is for a limited iteration 3 restricted to D1-D8 and regressions, declared the last one. Reduced iteration 1 FAIL 0.71 (`.moai/reports/t622/plan-audit-reduced-iter1.md`). No audit file for 0.2.1 exists in `.moai/reports/t622/` (listing at 0.2.2 authoring). This revision is for reduced iteration 2. Full-scope iterations before the split: FAIL 0.67 and FAIL 0.75.
- **Pre-write self-checks**: SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` on `SPEC-GIT-DELIVERY-PROCEDURE-001` → `PASS` (Bash, 0.2.2). Frontmatter 12 canonical fields present; `phase: "v3.2.0 target"`.
- **Counts**: REQ 26 numbers — judged 12 (001-006, 013-015, 024-026), moved 12, withdrawn 2. AC 30 numbers — judged 16 (001-006, 013-016, 025-030), moved 12, withdrawn 2.
- **commands-emit grounding (read, not executed — no make/go test at plan time)**: `Makefile:51-52` `commands-emit` = `COMMAND_EMIT_UPDATE=1 go test ./internal/template/commandemit/... -run TestGoldenCommittedArtifactsMatchEmission`; `Makefile:58-60` `commands-emit-check` = same test with the env var scrubbed, read-only. `golden_test.go:28` `templatesDir = "../templates"`; `:57-74` update branch writes only template-tree published paths; `:94-106` `TestCommandSourcesUnmodified` compares hashes before and after emission within one run, so a source edit alone does not trip it. `commandemit.go:49-53` roots `.claude/commands/moai` → `.agents/skills`. `loader.go:4-6` `argument-hint` and `allowed-tools` are "NOT carried into the published skill"; `emit.go:122-130` `renderSkill` writes header, `name`, `description`, body. Published copies: local and template both 7 lines, `diff` exit 0, case-insensitive `merge` count 0 and 0, both tracked (`git ls-files` lists all four of the two published copies and two command sources). Expected outcome A; judged by command in run-phase.
- **Re-measured controls for new or changed criteria (this tree = base)**:
  - AC-GDP-026 fragments (7)(8)(9), template and local copies each: lines 1 / 1 / 24; `--auto-merge` count 0 / 0 / 0; (ii) violations at fragment lines 1 (usage), 1 (hint), 9 (next-steps = file line 404). Local and template next-steps fragments `diff` exit 0. X1-X3 mutant fixture (6 lines) → (ii) hits lines 1, 3, 5; `--auto-merge` lines 1, 2, 4, 6.
  - AC-GDP-027 on fragments (7)(8)(9): `--no-merge` count 0 / 0 / 0 (both copies).
  - AC-GDP-013 command source `argument-hint` local vs template `diff` exit 0; command source copies `diff` → `2c2` only.
  - AC-GDP-005 command sources: `gh pr merge[^|]*--squash` count 0 and 0.
  - AC-GDP-025: `[ZONE:` lines in command sources 0 and 0; registry entries naming agent-common-protocol.md 13 and 13 (positive control); registry "others" detector (with `manager-[g]it`) 0 and 0; mutant fixture (3 lines: commands path, workflows/sync.md path, agent-common-protocol path) → 2.
  - AC-GDP-030 same-commit subset check (`/usr/bin/grep -v -x -F -f <src> <artifact>`): artifact SHA absent from src → exit 0 (FAIL detected); artifact SHA present → exit 1; empty src file → exit 0 with 1 line (FAIL — an empty source list cannot pass).
  - AC-GDP-015/016/025 new empty-set guards: `git diff --stat b412f8a33 -- internal/template/templates/` has no output on this tree, so `test -s` on the template or scope diff is exit 1 at base — the guard blocks a pass over an empty diff.
- **Docs-site X4 inputs (re-measured)**: `git grep -c -e --merge -e --no-merge b412f8a33 -- <moai-sync.md en/ko/ja/zh, what-is-moai-adk.md en/ko/ja/zh>` → moai-sync.md en 5, ko 4, ja 6, zh 9; what-is-moai-adk.md 1 each. Line numbers from `git grep -n` with the same paths plus `-e --auto-merge` (no `--auto-merge` hits) are recorded in spec.md §D.
- **Frozen check for the command files**: `.claude/commands/moai/sync.md` and `internal/template/templates/.claude/commands/moai/sync.md.tmpl` carry 0 `[ZONE:` lines and 0 zone-registry entries (local and template registry); edit site line 3 is outside any Frozen block.
- **Carried from 0.2.1 (unchanged base)**: AC-GDP-001 paragraphs 1 / autofail 1 / listgroup 0; AC-GDP-002 block 9 lines, fetch-without-rev-list 1, standalone rev-list at block line 5, joined 0, rev-list 1, session command 1, table rows 8; AC-GDP-004 squash 2, merge_method 0; AC-GDP-005 six `--squash` lines (manager-git 32/114, delivery 343/355, toml 26/108); AC-GDP-006 delivery Step 3.4 hits at section lines 8 and 20, doc-execution subsection hit at line 8; AC-GDP-014 `.toml` sections 11 and 9 lines diff exit 0; AC-GDP-026 fragments (1)-(6) lines 1/3/4/6/1/52 and violations 7; AC-GDP-027 delivery `--no-merge` at 8 and 19; AC-GDP-028/029 base (a) 0 and (b) 0 with their mutant fixtures; AC-GDP-015 fixture 5 letter / 2 digit tokens; AC-GDP-025 four clause counts 1 per copy.
- **plan_status**: audit-ready

## §E.2 Run-phase Evidence

Run-phase part 1 (pre-flight + M1, X2, M2, M3) by manager-develop, cycle_type ddd. Part 2 (M4-M6) is a
separate delegation. Evidence directory: `.moai/reports/t622/run/`. Tree: worktree
`.claude/worktrees/t622`, branch `WT-git-procedure-fixes`, start HEAD `fa13c27b6`, CARD_BASE
`f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf` (`card-base.txt`, 1 line), range control 356 names
(`card-range-names.txt`), snapshot freshness and mirror-baseline freshness both empty.

**Pre-flight (plan.md §C 1-8)**: all passed — `preflight-summary.md`. Every RED cell showed its expected
hit on the BASE exports. One recorded deviation: the AC-GDP-030 control (c) red step returned make exit 2
(`make` reports a failed recipe as 2; the recipe line reads `Error 1` and the drift FAIL is printed),
where acceptance.md writes "exit 1". Sorted mirror sets equal the 0.2.5 baseline (`diff` exit 0).

**Commits (in order)**: `0da3bebf0` pre-flight baseline + status flip (baseline committed before any
change — verification-claim-integrity §2.3); `3f6c5163f` M1; `2f4dfd803` X2 (command source
argument-hint); `af54bf1ff` M2 (+ AC-GDP-030 outcome); M3 is the commit carrying this section.

| AC | Scope judged in part 1 | Command (literal, run from the worktree root or `.moai/reports/t622/run`) | Output file(s) | Result | Status |
|---|---|---|---|---|---|
| AC-GDP-001 | local + template sections (`.toml` clause waits for M4 regeneration) | `sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' <mg>` → paragraph / autofail / listgroup awk (acceptance.md verbatim) | `ac001-{local,template}-sync.md`, `-paras.txt`, `-autofail.md`, `-listgroup.md`, `ac001-judge-summary.txt`, `ac001-reading.md` | paras 1·1; autofail `test -e` 0, `test -s` 1 (both); listgroup empty; reading both answers Yes | PASS for L·T; `.toml` pending M4 |
| AC-GDP-004 | both copies | `grep -c 'gh pr merge --squash --delete-branch'` / `grep -c 'gh pr merge --<merge_method> --delete-branch'` / `grep -n -E 'merge_method.*(squash\|default)'` | `ac004-{local,template}-squash.txt`, `-resolved.txt`, `-source.txt`, `ac004-judge-summary.txt` | 0 / 2 / exit 0 (local 377, template 352: the resolution-source sentence) | PASS |
| AC-GDP-005 | eight scope pairs + manager-git (`.toml` clause waits for M4) | `grep -c -E 'gh pr merge[^\|]*--squash'` per file; default-sentence and example `grep -c -F` | `ac005-{local,template}-mg.txt`, `-mg-default.txt`, `-mg-example.txt`, `-others.txt` | mg 1·1 (the default explanation), default 1·1, example 1·1, others all 0 | PASS for scope files; `.toml` (expected 1) pending M4 |
| AC-GDP-006 | both copies | Step 3.4 / Worktree Context Detection extraction + extended detector (acceptance.md verbatim) | `ac006-{local,template}-*`, `final-{local,template}-006-*`, `ac006-{local,template}-optin.txt`, `ac006-reading.md` | dl-default `test -e` 0 `test -s` 1; de-default grep exit 1; source dl 2 / de 1; `--auto-merge` in manager-git 4·4; reading all No / `--auto-merge` | PASS |
| AC-GDP-026 | nine fragments × both copies | fragment extraction + (i)/(ii)/(iii) (acceptance.md verbatim) | `ac026-{local,template}-*`, `final-{local,template}-026-*`, `ac026-frag-lt.txt`, `ac026-reading.md` | (i) every fragment ≥1 (qgc 1+2); (ii) `test -e` 0 `test -s` 1; (iii) 1,1,1,2; 5 merge lines per copy; fragments L=T; reading 10/10 Yes | PASS |
| AC-GDP-027 | nine fragments × both copies | (i)/(ii) awk (acceptance.md verbatim) | `ac027-{local,template}-*`, `final-{local,template}-027-*` | dl `--no-merge` lines 1; (i) and (ii) `test -e` 0 `test -s` 1; other fragments 0 `--no-merge` lines (`ac027-*-per-fragment.txt`) | PASS |
| AC-GDP-028 | mg + dl sections × both copies | (a)/(b)/mode-lines awk (acceptance.md verbatim) | `ac028-{local,template}-*`, `final-{local,template}-028-*`, `ac028-reading.md` | (a) mg 1, dl 1; (b) `test -e` 0 `test -s` 1; 14 mode lines, reading 14/14 Yes | PASS |
| AC-GDP-029 | same sections | (a)/(b) awk | `ac029-{local,template}-*`, `final-{local,template}-029-*`, `ac028-reading.md` | (a) mg 1, dl 1; (b) `test -e` 0 `test -s` 1 | PASS |
| AC-GDP-030 | command source pair + published pair | control (c), `make commands-emit`, `make commands-emit-check`, post-commit range/commit-list judges | `ac030-*`, `ac030-outcome.md` | control detected drift; emit exit 0; check exit 0; changed list empty; src commits `2f4dfd803…`; artifact commits empty; path control 2; orphan empty; published L/T diff exit 0; flag detector empty | PASS — case (A) |

Not judged in part 1 (M4-M6 scope): AC-GDP-002, 003, 013, 014, 015, 016, 025, and the `.toml` clauses of
AC-GDP-001/005. Pre-commit parity sanity checks run in part 1 (not the AC-GDP-013 verdict): every edited
pair's diff body equals its BASE body (`m1-pair-*`, `m2-pair-*`, `m3-pair-*`: body diff exit 0) and the
`argument-hint` lines match (`ac013-hint.diff` exit 0). Cf: every edited file and its BASE copy count 0
(`cf-m1-*`, `cf-m2-*`, `cf-m3-*`); counter proven on a 2-character control (`cf-control-2.count` = 2).

Residual noted for part 2 / sync: doc-execution.md line 34 still says `is_worktree_context` is stored
"for use in Phase 13", although no Phase 13 step reads it after M1 (left unchanged — it does not tie the
flag to merging, and the next sentence states it is not a merge condition).

### Part 2 (M4, M5, M6) — manager-develop, cycle_type ddd

Full report: `.moai/reports/t622/run/run-report.md`. Part-2 start HEAD `7a02b90e2`; M6 judging HEAD
`35c0e30df`; CARD_BASE `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf` (unchanged, 1 line); range control 760
names; snapshot and mirror freshness both empty (BASE exports valid, no CARD_BASE re-export). Local
develop moved to `85868148c` (25 commits past CARD_BASE) — recorded, not absorbed. Scope content is
identical from the `.toml` commit `5708e04d2` to the final HEAD (`git diff --name-only 5708e04d2 HEAD --
. ':!.moai/reports/t622/run/'` empty).

**Commits (in order)**: `cc51d8479` M4 agents-emit-check RED evidence (committed before the artifact);
`5708e04d2` M4 regenerated `manager-git.toml` (`make agents-emit`, never hand-edited); `8a115e0f5` M4
evidence; `35c0e30df` M5 evidence (no edits); `8d1b8c920` M6 evidence (no edits); the commit carrying
this section adds `run-report.md`.

| AC | Scope judged in part 2 | Output file(s) (under `.moai/reports/t622/run/`) | Result | Status |
|---|---|---|---|---|
| AC-GDP-001 | `.toml` clause (new) + L·T re-run | `ac001-toml-*`, `p2-ac001-*`, `ac001-reading.md` (M4 addendum) | paras 1·1·1; autofail `test -e` 0 / `test -s` 1 ×3; listgroup empty ×3; `.toml` paragraph `cmp` 0 vs template | PASS |
| AC-GDP-002 | both copies (M5) | `ac002-{local,template}-*`, `ac002-*matrix.diff`, `p2-ac002-mut-i-*` | (a) `order=PASS shape=PASS` ×2; (b) section diff exit 0 ×2, matrix diff exit 0 ×2; control b412 `order=FAIL` | PASS |
| AC-GDP-003 | both copies (M5) | `ac003-*.md`, `ac003.diff`, `ac003-template.diff` | Pre-Edit section diff exit 0 ×2 | PASS |
| AC-GDP-004 | re-run | `p2-ac004-*` | squash 0·0, resolved 2·2, source exit 0·0 | PASS |
| AC-GDP-005 | `.toml` clause (new) + re-run | `ac005-toml*.txt`, `p2-ac005-*` | `.toml` 1 line = default sentence; mg 1·1; others 0 | PASS |
| AC-GDP-006 | re-run | `p2-{local,template}-006-*`, `p2-ac006-*`, `ac006-reading.md` (final-tree addendum) | dl-default `test -s` 1 ×2, de-default exit 1 ×2, source 2/1, opt-in 4·4 | PASS |
| AC-GDP-013 | (a)-(e), M4 and M6 | `ac013-*`, `m4-ac013-*`, `ac013-gotest.*`, `mirror-m4*`, `mirror-m6*`, `mirror-new-fail*.txt`, `mirror-lost-pass*.txt` | (a) exit 0·0; (b) 8 body diffs exit 0; (c) hint diff 0; (d) go test exit 0, top-level PASS 2, acp trace 6; (e) new-FAIL and lost-PASS empty both runs, sorted sets equal baseline | PASS |
| AC-GDP-014 | full | `ac014-*`, `m6-agents-emit-check.*` | RED make exit 2 (recipe `Error 1`, sha256 mismatch) → emit exit 0 → GREEN exit 0; changed = `manager-git.toml` only; example 1; sync/pram diff exit 0 | PASS |
| AC-GDP-015 | full | `ac015-*`, `p2-ac015-fixture-tokens.txt` | 38 added lines in 9 template files; SPEC/REQ/date/`CLAUDE.local` exit 1; sha-letter `test -s` 1; hex tokens 0; control 5/2 | PASS |
| AC-GDP-016 | full | `ac016-acp-commits.txt`, `p2-ac016-*` | `test -e` 0 / `test -s` 1 over 34 card commits; control 2 commits | PASS |
| AC-GDP-025 | M4 and M6 | `ac025-*`, `m4-ac025-*`, `p2-ac025-*` | no Frozen +/- lines; clauses 1·1 ×4; registry 13·13 / 0·0; controls 2 / 2 | PASS |
| AC-GDP-026 | re-run | `p2-{local,template}-*.md`, `p2-{local,template}-026-*` | (i) all ≥1; (ii) `test -s` 1 ×2; (iii) 1,1,1,2; merge-line bodies `cmp` 0 vs read targets | PASS |
| AC-GDP-027 | re-run | `p2-{local,template}-027-*` | dl 1 line; (i)/(ii) `test -s` 1 ×2 | PASS |
| AC-GDP-028 | re-run | `p2-{local,template}-028-*`, `p2-{local,template}-mg.md` | (a) 1/1 ×2; (b) `test -s` 1 ×2; mode-line bodies `cmp` 0 vs read targets | PASS |
| AC-GDP-029 | re-run | `p2-{local,template}-029-*` | (a) 1/1 ×2; (b) `test -s` 1 ×2 | PASS |
| AC-GDP-030 | re-run (post-commit judges) | `p2-ac030-*`, `m4-/m6-commands-emit-check.*` | case (A): changed empty, artifact commits empty, src `2f4dfd803…`, path control 2, orphan empty, published diff 0, flags empty; checks exit 0·0 | PASS |

Mirror non-regression, generator RED→GREEN, AC-GDP-016 control and all gaps: `run-report.md` §2-§5.
Cf: regenerated `.toml` 0, edited reading records 0, counter 2 on the planted control.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-11
run_commit_sha: pending-backfill-run   # the commit carrying this block cannot cite itself; evidence HEAD before it: 8d1b8c9206b6c95350302f090374f822101221f2
run_status: audit-ready
ac_pass_count: 16          # AC-GDP-001..006, 013..016, 025..030 (016 is SHOULD-PASS)
ac_fail_count: 0
ac_na_count: 14            # placeholders AC-GDP-007..012, 017..024
preserve_list_post_run_count: 0   # agent-common-protocol.md (both copies) unchanged: AC-GDP-016 empty, AC-GDP-002 (b)/003 diff exit 0
l44_pre_commit_fetch: not-run     # lane does not fetch/push; local develop read directly (85868148c), not absorbed
l44_post_push_fetch: not-applicable   # no push (lead batch-pushes develop)
new_warnings_or_lints_introduced: none observed   # make agents-emit-check / commands-emit-check exit 0; no Go source changed; golangci-lint not run (no Go edits)
cross_platform_build:
  applicable: false        # Markdown / .tmpl / generated .toml only; no Go source edits
  go_build: not-run
  goos_windows_build: not-run
targeted_tests:
  - "go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1 -> exit 0, top-level PASS 2"
  - "mirror non-regression (M4, M6) -> new-FAIL 0, lost-PASS 0"
total_run_phase_files: 17  # 8 scope pairs (16) + regenerated manager-git.toml; agent-common-protocol.md not edited
m1_to_mN_commit_strategy: per-milestone commits on WT-git-procedure-fixes (0da3bebf0 baseline, 3f6c5163f M1, 2f4dfd803 X2, af54bf1ff M2, 7a02b90e2 M3, cc51d8479 M4 RED, 5708e04d2 M4 toml, 8a115e0f5 M4 evidence, 35c0e30df M5, 8d1b8c920 M6); no push
gaps: [ci-not-observed (branch unpushed), develop-not-absorbed (25 commits; range judges are pre-merge), go-test-all-not-run, embed-check-not-run]
```

## Sync step 1 record (manager-docs, pre-audit — not the §E.4 close)

- **Edited** (commit `0dc007201`, on top of run HEAD `28f9cc6c9`): `doc-execution.md:34` in both copies (`.claude/skills/moai/workflows/sync/doc-execution.md`, template mirror), identical edit — "Store result as `is_worktree_context` boolean for use in Phase 13" → "Store result as `is_worktree_context` boolean as informational context only; no later phase reads it". Line 36 unchanged (`cksum` equal to HEAD). CHANGELOG `[Unreleased]` → `### Changed` top entry (B12: pre-emission `grep -c` 0; 16 judged ACs per acceptance.md; 12 cited paths `ls` exit 0 — `changelog-path-check.txt`). spec.md frontmatter `status: implemented`, `updated: 2026-09-12`; no body edit. README: no merge-flag text (`grep -c -E 'auto-merge|no-merge|moai sync'` → 3 per locale, all plain `/moai sync` usage), not edited.
- **Judges re-run** (`.moai/reports/t622/sync/`): AC-GDP-006 doc-execution L·T — section `test -s` 0·0, default detector grep exit 1·1, extended detector `test -e` 0 / `test -s` 1 ×2, `manager-[g]it[.]md` 1·1, `--auto-merge` in manager-git.md 4·4; BASE control grep exit 0 (section line 8). AC-GDP-013 (b) doc-execution — BASE L·T diff exit 1, post diff exit 1 (same 9 hunk headers), body diff exit 0, base body 6008 B `cmp` 0 vs `run/ac013-doc-execution-base.body`; one-copy mutant body diff exit 1. AC-GDP-015 on `git diff develop...HEAD -- internal/template/templates/` (post-commit) — `test -s` 0, 39 added lines over 9 files, SPEC/REQ/date/`CLAUDE.local` greps exit 1 ×4, hex tokens 0, sha-letter `test -e` 0 / `test -s` 1; controls spec.md 8/61/23, fixture 5 letter / 2 digit. `make commands-emit-check` exit 0, `make agents-emit-check` exit 0. Cf: local 0, template 0, CHANGELOG entry 0, planted control 2.
- **Sync-audit F1 correction** (`.moai/reports/t622/sync-audit.md` F1; evidence `.moai/reports/t622/sync-fix/`): the `0dc007201` wording of `doc-execution.md:34` was false — `delivery.md` Phase 13 Step 3.2 "**Worktree context**" route (local 284-287 / template 259-262) and Phase 14 "**If worktree context:**" options (local 447-450 / template 422-425) do consume it. Both copies now read "Store result as `is_worktree_context` boolean; it selects the worktree delivery route (Phase 13 Step 3.2) and the worktree next-step options (Phase 14), but it never decides auto-merge"; line 36 unchanged. The step-1 record's citation of `0dc007201` for line 34 is superseded by the sync-audit F1 fix commit `ae4861945`. Re-judged: AC-GDP-006 de L·T `test -s` 0·0, default grep exit 1·1, extended `test -e` 0 / `test -s` 1 ×2, `manager-[g]it[.]md` 1·1, BASE control grep exit 0 ×2 (section line 8), reading `ac006-reading.md`; AC-GDP-013 (b) doc-execution post diff exit 1, body `cmp` 0 vs `run/ac013-doc-execution-base.body` (6008 B each); `make commands-emit-check` 0, `make agents-emit-check` 0; Cf local 0 / template 0 / planted control 2. AC-GDP-015 re-run after fix commit `ae4861945` (`sync-fix/ac015-*`): `git diff develop...HEAD -- internal/template/templates/` `test -s` 0, 39 added lines (new line 34 present, count 1), SPEC/REQ/date/`CLAUDE.local` greps exit 1 ×4, hex tokens 0, sha-letter `test -e` 0 / `test -s` 1; controls spec.md 8/61/23, fixture 5 letter / 2 digit.
- **Pending**: sync-auditor, lane verdict (`.moai/reports/t622/verdict.md`), then the separate close step (`status: completed`, §E.4, `sync_commit_sha`).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-12
sync_status: closed        # audit-ready -> closed; spec.md status implemented -> completed in this close commit
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own SHA; manager-docs backfills it in a follow-up commit (spec-frontmatter-schema.md § SHA placeholder backfill exemption)
sync_commits:
  - 0dc007201   # sync step 1 — doc-execution.md:34 wording (later superseded by ae4861945), CHANGELOG [Unreleased] entry, status: implemented
  - 50cebf3c2   # sync step 1 — AC-GDP-015 post-commit judge and progress record
  - ae4861945   # sync-audit F1 fix — is_worktree_context wording in doc-execution.md:34 (both copies)
  - c61405c8b   # AC-GDP-015 re-run after the F1 fix
sync_audit:
  iteration_1: FAIL 84   # .moai/reports/t622/sync-audit.md (F1 blocking; F2-F7 accepted as follow-ups/notes)
  iteration_2: PASS 91   # .moai/reports/t622/sync-audit-iter2.md (delta audit: F1 resolved, no regression)
lane_verdict: .moai/reports/t622/verdict.md   # lane recommendation PASS; final PASS/FAIL is the lead's
changelog_entry_position: "[Unreleased] ### Changed, top entry (0dc007201)"
b12_self_test_a: "pre-emission grep -c SPEC-GIT-DELIVERY-PROCEDURE-001 CHANGELOG.md -> 0 (sync step 1)"
b12_self_test_b: "16 judged ACs per acceptance.md (sync step 1)"
b12_self_test_c: "12 cited paths ls exit 0 (.moai/reports/t622/sync/changelog-path-check.txt)"
frontmatter_status_transitions:
  in_progress_to_implemented: 0dc007201
  implemented_to_completed: this close commit (subject names SPEC-GIT-DELIVERY-PROCEDURE-001)
known_placeholders:
  - "§E.3 run_commit_sha: pending-backfill-run — owned by run-phase (manager-develop); left untouched by this close"
  - "§E.4 sync_commit_sha: pending-backfill-sync — owned by manager-docs; backfilled after this close lands"
token_accounting: not-measured   # moai tokens record writes a machine-local ledger from a session transcript; no CLI path writes progress.md §I (WriteSectionI has no non-test caller), so no §I section is written
```

**Attribution — what this close describes.** The close commit's tree carries the card's full content as of this commit on `WT-git-procedure-fixes` (parent `41067054d`). The instruction files (`.claude/`, `internal/`, `CHANGELOG.md`) last changed in `ae4861945`; the commits after it (`c61405c8b`, `947cc8439`, `41067054d`, and this close) touch only SPEC records and `.moai/reports/t622/` evidence. The measurements cited above were taken before the card absorbed local develop (card base `f1f034bb4`, the merge-base with local develop `ac6c42c2d` at close time). The range-based judges AC-GDP-014, AC-GDP-015, AC-GDP-016, AC-GDP-025, AC-GDP-030 and M5 are pre-merge judges: the lane re-measures them on the merge tree after absorbing local develop in the integration window, per `verdict.md` §4. This close therefore does **not** claim a post-absorb measurement.

## §F Phase 4 Mode Selection

Decision: serial

Recorded by the lane (card orchestrator) before the first run-phase `Agent()` spawn, after Implementation Kickoff Approval (operator via lead, 2026-09-11, conditioned on N1-N3, applied in `ccfe3005e`).

**Input parameters**

| Parameter | Value |
|---|---|
| tier | M (judged REQ 12, AC 16) |
| scope (files) | 19 affected (nine scope files as local/template pairs = 18, plus generated `manager-git.toml`); 21 if the published `moai-sync/SKILL.md` pair changes (AC-GDP-030 case B) |
| domain count | 2 — agent/rule/skill/command instruction text, and generated artifacts via `make agents-emit` / `make commands-emit` |
| file language mix | Markdown instruction files, one Go-template command source (`.md.tmpl`), generated `.toml` and published `SKILL.md`; no Go source edits |
| concurrency benefit | LOW — milestones share files (delivery.md in M1 and M2, manager-git.md in M1-M3) and carry ordering obligations: command-source regeneration in the same commit (M1), agent regeneration after all manager-git.md edits (M4), always-loaded rule edit last (M5, AC-GDP-016) |
| Agent Teams prerequisites | not requested (no `--team`) |
| tree at selection | HEAD `7f5a4420f` (SPEC 0.2.4 `ccfe3005e` + evidence), clean; `moai session list --json --filter-spec=SPEC-GIT-DELIVERY-PROCEDURE-001` → `[]` |

**Mode evaluation**

| Mode | Result | Rationale |
|---|---|---|
| direct | not selected | multi-file, generator runs and acceptance judges exceed a direct orchestrator edit |
| serial | **selected** | one write-capable `manager-develop` runs M1→M6 in order; shared files and commit-ordering obligations make sequential execution the correct shape |
| fanout | not selected | below the multi-domain threshold, and parallel writers would race on delivery.md / manager-git.md and break the same-commit and last-edit ordering |
| sweep | not selected | not ≥ ~30 files and not one uniform mechanical transform |
| agent-team | not selected | not requested by the operator |

Decision: serial

**Justification.** The run phase edits a small set of instruction files line by line where later milestones depend on earlier commits (M4 regenerates the `.toml` only after M1-M3 finish editing `manager-git.md`; M5 must be the final instruction-file commit). A single sequential `manager-develop` preserves those orderings and keeps one writer on the worktree. Compile scope needs no `internal/cli` slot: `go list -deps -test` of `./internal/template/`, `./internal/template/commandemit/` and `./internal/template/agentemit/` lists no `internal/cli` package (same-form control on `internal/template` hits).
