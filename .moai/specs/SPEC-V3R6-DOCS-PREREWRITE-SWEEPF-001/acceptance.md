# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 (revision 0.1.0)

## A. Pin and conventions

- **Document-level pin.** Tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-workflow-cmd-index`, equal to local `main`). Every cell names the commit in its command (`git grep … 74ad197380`, `git ls-tree … 74ad197380`), so the pin is in the cell. `git status --short` printed nothing before and after the measurements. Measured on 2026-10-19.
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no pipe, redirection, chaining, or subshell. Its verbatim stdout, its exit code, and the tree are separate fields. The shell tool displays no status for a command with no output; a zero-match cell therefore says `no output` and gives the exit code git defines for it (1), marked `by git grep semantics, not displayed` (gap G-1).
- **Pathspec.** Every cell addressing the twelve files uses three glob pathspecs: `':(glob)docs-site/content/*/workflow-commands/_index.md'`, `…/_meta.yaml'`, and `…/moai-design.md'`. C-3 prints the twelve paths these select, so an empty or short sweep is visible.
- **Classes.** `release-blocking (gated)`: re-executable on the tree; counts toward the completion verdict once its gate rows carry a verdict. `guard`: green today; its failure is shown by a named mutant. `unadopted`: no RED-now cell can be measured until a decision row has a verdict.
- **Verdict basis.** Q13 and Q17 carry operator verdicts (board record `d-20261019T015908Z-220a`, 2026-10-19); Q14 is open. A criterion that depends on Q14 states each branch's pass condition.
- **Tool provenance (verification-claim-integrity §2.2).** Only `git` was used (system build); no `moai` binary measurement is cited. Pattern cells use `git grep -E` and, for C-2, `git grep -P`.

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Gate rows | Class | RED-now cells | Green-path milestone |
|----|-----|-----------|-------|---------------|----------------------|
| AC-PRC-001 | REQ-PRC-001 (+ REQ-SWF-006 on the marker branch) | Q14 (page and `_meta.yaml` parts); Q13, Q17 decided | release-blocking (gated) | C-1 (C-3 sweep control) | M2 (`_index.md`), M3 (`_meta.yaml`, page) |
| AC-SWF-002 | REQ-PRC-001, REQ-SWF-002 | Q14 (page part) | release-blocking (gated) | C-2 (C-8 list, C-3 sweep) | M2, M3 |
| AC-SWF-003 | REQ-SWF-002, REQ-SWF-003 | none (Q13, Q17 decided) | release-blocking | C-4 | M2 |
| AC-SWF-004 | REQ-SWF-007 | none | guard | none (green today; mutants in recipe R-4) | M4 |
| AC-SWF-005 | REQ-SWF-004 | Q14 = remove | release-blocking if Q14 = remove; otherwise not applicable | C-5 | M3 |
| AC-SWF-006 | REQ-SWF-005, REQ-SWF-006 | Q14 = rewrite or marker | unadopted until Q14 has a verdict (rewrite part has a RED cell: C-6) | C-6 | M3 |

Every REQ has an AC and every AC maps to a REQ (7 of 7, 6 of 6).

## B. Acceptance criteria

### AC-PRC-001 — no unresolved design-family citation remains in the twelve files (page part of REQ-PRC-001)

- **Given** 27 lines in the twelve files match the design family (C-1), spread over all twelve files, and the sweep selects exactly twelve files (C-3).
- **When** M2 deletes the design table rows of the four `_index.md` files (Q13, Q17 verdicts), and M3 applies the Q14 answer to the four `_meta.yaml` entries and the four pages,
- **Then**
  1. The C-1 command run on the closing commit prints nothing and exits 1 (Q14 = remove or rewrite). On the marker branch (Q14 = marker), the C-1 command run on the closing commit prints matches only in `moai-design.md` files, and the C-1 pattern with the marker negated (recipe R-1) prints nothing.
  2. The C-3 command run on the closing commit prints twelve paths (Q14 = rewrite or marker) or eight paths (Q14 = remove). A closing run that prints fewer paths than that is a failure of the sweep, not a pass.
  3. The C-1 command run on `74ad197380` still prints the twelve lines of C-1 (positive control: the pattern matches where the work has not been done).
- **RED-now cells:** C-1. Red because 27 lines in twelve files carry the design family; it flips only when this work edits those files.
- **Green path:** M2 flips the five `_index.md` lines (en, ja, zh: 1 each; ko: 2). M3 flips the 4 `_meta.yaml` lines and the 18 page lines (5, 5, 4, 4) according to Q14.
- **Mutant probe:** deleting the sentence around a citation instead of the row leaves the C-4 counts unchanged (see AC-SWF-003 Then 1); leaving the design family in one locale only leaves that locale's line in the C-1 output; replacing `/moai design` with another unregistered form (for example `/moai ui`) passes C-1 and fails AC-SWF-002; a spelling that C-1 does not match (two spaces between the words) passes C-1, and C-2 still matches it only if the verb token follows the single separator, so the closing review reads the diff of the twelve files, not the grep alone (gap G-9).
- **Gate:** Q14 for the page and `_meta.yaml` parts. The `_index.md` part has no open gate.
- **Done when:** Then 1 to 3 hold on the committed tree and the commit SHA is recorded in `progress.md` §E.2 for the parent's M6.

### AC-SWF-002 — every remaining `/moai` slash form in the twelve files resolves (REQ-PRC-001, REQ-SWF-002)

- **Given** the registered commands are the seventeen files of C-8, and the negative-list search over the twelve files returns the same 27 lines as C-1 (C-2).
- **When** the Q13 verdict is applied (replace where an equivalent exists, delete otherwise) and the Q14 answer is applied,
- **Then**
  1. The C-2 command run on the closing commit prints nothing and exits 1 (Q14 = remove or rewrite); on the marker branch it prints only lines that carry the marker.
  2. The C-8 command run on the closing commit still prints seventeen paths with the same names (the list in the C-2 pattern is the list on the tree; a registered-command change invalidates the cell and the cell is re-run, not tuned).
  3. No replacement introduced by M2 or M3 is itself an unregistered `/moai` form (this is what Then 1 measures).
- **RED-now cells:** C-2. Red for the stated reason: the only unregistered verb in the twelve files is `design`, on 27 lines.
- **Green path:** the same edits as AC-PRC-001; this criterion additionally rejects a replacement that names another unregistered command.
- **Mutant probe:** replacing the design row with `/moai ui` or `/moai design-phase` is matched by the pattern (the token is not in the list and is not followed by a listed verb boundary); a pattern edited to include `design` in its list would print nothing and is detected by the C-2 positive control at `74ad197380` (the base prints twelve lines).
- **Gate:** Q14 for the page part.
- **Done when:** Then 1 to 3 hold on the committed tree.

### AC-SWF-003 — the four locales receive the same deletions in the same change (REQ-SWF-002, REQ-SWF-003)

- **Given** the per-locale count of `/moai` table rows in `_index.md` is en 21, ja 21, ko 23, zh 21 (C-4). The Korean file holds two design rows (a summary-table row and a subcommand-table row) and one `/moai goal` summary row that the other locales lack.
- **When** M2 deletes the design rows in all four locales (ko canonical first, then en, then ja and zh) in one commit,
- **Then**
  1. The C-4 command run on the closing commit prints exactly: en 20, ja 20, ko 21, zh 20. The ko figure stays one above the others because of the `/moai goal` summary row, which resolves and is not touched; it is not an error.
  2. The C-1 count of `_index.md` lines is zero in all four locales (AC-PRC-001 Then 1).
  3. The commit that makes these edits changes all four `_index.md` files (`git show --stat` lists four paths among the twelve files; recipe R-3).
- **RED-now cells:** C-4. Red because the counts today are 21, 21, 23, 21, not 20, 20, 21, 20; the figures differ from the closing figures by exactly the design rows (1, 1, 2, 1).
- **Green path:** M2 flips C-4 to 20, 20, 21, 20.
- **Mutant probe:** a Korean edit that removes only line 66 leaves ko at 22 and fails Then 1; an edit that deletes the wrong row (for example the `/moai goal` summary row) leaves the design lines and fails AC-PRC-001; an edit that changes three locales and leaves one fails Then 1 and Then 3; replacing a row with a prose sentence keeps the row count but leaves the design family, and is caught by AC-PRC-001; deleting more than the design rows (for example the whole table) drops a count below the closing figure.
- **Gate:** none (Q13 and Q17 decided). If the operator later answers O-4 by editing the count sentences, they are not table rows and do not change C-4.
- **Done when:** Then 1 to 3 hold on the committed tree.

### AC-SWF-004 — nothing outside the twelve files changes (REQ-SWF-007) — guard

- **Given** the tree is clean at `74ad197380` and the other six workflow-commands pages per locale hold no design-family line (C-1 lists only the twelve files).
- **When** the whole change is committed,
- **Then**
  1. The R-4 command (`git diff --name-only 74ad197380 <closing-sha> -- docs-site README.md README.ko.md README.ja.md README.zh.md`) prints only paths from § Scope of spec.md (twelve paths, or eight modified and four deleted); it prints none of: `docs-site/data/menu/main.yaml`, `docs-site/.locale-parity-baseline`, or any `workflow-commands/moai-<goal|harness|plan|project|run|sync>.md`.
  2. The same command with the pathspec `.claude .moai/docs scripts internal cmd` prints nothing.
- **RED-now cells:** none. Green today by construction (no edit exists); the criterion is a guard. Its failure is shown by the mutants in R-4.
- **Mutant probe (recipe R-4):** touching `docs-site/data/menu/main.yaml`, or any of the other six pages, or `.locale-parity-baseline`, makes Then 1 print an extra path.
- **Gate:** O-2 and O-3 (if the leader adds `main.yaml` or the baseline file to the change, Then 1 is amended to name them).
- **Done when:** Then 1 and 2 hold on the committed tree.

### AC-SWF-005 — no link to a removed page remains (REQ-SWF-004) — applies only if Q14 = remove

- **Given** the page is linked from the four `_index.md` files and registered in the four `_meta.yaml` files (C-5), and two paths outside the twelve files also name it (C-5: `docs-site/data/menu/main.yaml:208`, and the parity baseline is not in the C-5 scope but is named in O-3).
- **When** Q14 records removal and M3 deletes the four pages and the four `_meta.yaml` entries,
- **Then**
  1. The C-5 command run on the closing commit prints nothing under `docs-site/content` and exits 1; its only remaining lines (if any) are in `docs-site/data/menu/main.yaml`, which is outside the change set (O-2).
  2. Each `_meta.yaml` still parses and lists only keys that name an existing page or `index` (recipe R-5: `yq 'keys'` on each file, compared with the page file names of the same directory).
- **RED-now cells:** C-5. Red because 9 lines under `docs-site/content` name the page today (4 `_index.md` rows, with ko holding 2, and 4 `_meta.yaml` keys); the C-5 output also shows the line in `main.yaml`.
- **Green path:** M2 flips the five `_index.md` lines; M3 flips the four `_meta.yaml` keys.
- **Mutant probe:** deleting the pages and leaving the `_meta.yaml` key fails Then 1 and Then 2; deleting the key and leaving the page fails AC-SWF-004 only if the page is not in the deleted set, so Then 2 compares the two sets directly.
- **Gate:** Q14. If Q14 is not "remove", this criterion is not applicable and the parent's M6 imports it as such.
- **Done when:** Then 1 and 2 hold.

### AC-SWF-006 — a kept page carries no slash title (rewrite) or a marked citation (marker) (REQ-SWF-005, REQ-SWF-006)

- **Given** the four page titles read `title: /moai design` today (C-6) and the page opens with a callout that tells the reader to type `/moai:design` (C-1).
- **When** Q14 records rewrite or marker and M3 edits the four pages and the four `_meta.yaml` entries,
- **Then**
  1. (rewrite) The C-6 command run on the closing commit prints nothing and exits 1, the C-1 command prints nothing (AC-PRC-001 Then 1), and the four page files still exist (C-3 prints twelve paths).
  2. (marker) Every line that the C-1 command still matches in a page carries the marker token of progress.md §E.2, and the per-locale count of marker-bearing lines equals the per-locale count of C-1 matches in that page.
  3. In both branches the four pages describe the same route: each page file has the same number of top-level headings (`^## `) across the four locales (recipe R-6), so the ja and zh pages are no longer shorter than the en and ko pages for lack of a section.
- **RED-now cells:** C-6 (rewrite part). Red because four titles are slash forms. The marker part has no RED-now cell: the token is not defined until Q14 is answered, so Then 2 is unadopted until then.
- **Green path:** M3 flips C-6.
- **Mutant probe:** retitling in three locales fails Then 1; a marker applied to a subset of lines fails Then 2; copying the en page structure over ja and zh without translation fails nothing here and is caught in review (gap G-5).
- **Gate:** Q14 = rewrite or marker.
- **Done when:** Then 1 to 3 hold for the chosen branch.

## C. Evidence ledger (verbatim)

Pin for every cell: tree commit `74ad197380a0711cdb9816d4e37c94d29af730ea`. Measured 2026-10-19 on the working tree at that commit (clean).

### C-1 — the design family in the twelve files (RED-now; AC-PRC-001)

- Command: `git grep -c -E "/moai[: ]design([^[:alnum:]-]|$)" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 12 lines; derived total 27):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:1
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/ja/workflow-commands/_index.md:1
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:4
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/zh/workflow-commands/_index.md:1
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:4
```

- Reading: re-measures the premise check (27 lines in 12 files). Red because the design family is cited; the work edits exactly these files.

### C-2 — any `/moai` verb outside the registered list (RED-now; AC-SWF-002)

- Command: `git grep -c -P "/moai[: ](?!(clean|codemaps|e2e|feedback|fix|gate|goal|gtd|harness|loop|mx|plan|project|review|run|sync|todo)(?![a-z0-9-]))[a-z][a-z0-9-]*" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 12 lines, identical to C-1):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:1
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/ja/workflow-commands/_index.md:1
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:4
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/zh/workflow-commands/_index.md:1
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:4
```

- Reading: the list in the pattern is the seventeen files of C-8. The output equals C-1 line for line, so `design` is the only unregistered verb in the twelve files. The cell is its own positive control: it matches where the work has not been done.

### C-3 — the swept set (control; AC-PRC-001, AC-SWF-002)

- Command: `git ls-files -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: working tree at `74ad197380`.
- Stdout (verbatim; 12 lines):

```
docs-site/content/en/workflow-commands/_index.md
docs-site/content/en/workflow-commands/_meta.yaml
docs-site/content/en/workflow-commands/moai-design.md
docs-site/content/ja/workflow-commands/_index.md
docs-site/content/ja/workflow-commands/_meta.yaml
docs-site/content/ja/workflow-commands/moai-design.md
docs-site/content/ko/workflow-commands/_index.md
docs-site/content/ko/workflow-commands/_meta.yaml
docs-site/content/ko/workflow-commands/moai-design.md
docs-site/content/zh/workflow-commands/_index.md
docs-site/content/zh/workflow-commands/_meta.yaml
docs-site/content/zh/workflow-commands/moai-design.md
```

### C-4 — `/moai` table rows per locale in `_index.md` (RED-now; AC-SWF-003)

- Command: `git grep -c -E "^\| \[.?/moai " 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 4 lines):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:21
74ad197380:docs-site/content/ja/workflow-commands/_index.md:21
74ad197380:docs-site/content/ko/workflow-commands/_index.md:23
74ad197380:docs-site/content/zh/workflow-commands/_index.md:21
```

- Reading: the pattern is a table row whose first cell links a `/moai` command (the optional character after `[` is the opening backtick). Korean holds two design rows (lines 42 and 66) and a `/moai goal` summary row (line 41) that the other locales lack. Closing figures: 20, 20, 21, 20 (en, ja, ko, zh).

### C-5 — references to the page (RED-now; AC-SWF-005)

- Command: `git grep -n -E "moai-design" 74ad197380 -- docs-site/content docs-site/data README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 9 lines):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:62:| [`/moai design`](./moai-design) | Design-phase collaboration (manager-design D1-D5) |
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/content/ja/workflow-commands/_index.md:62:| [`/moai design`](./moai-design) | デザイン段階のコラボレーション (manager-design D1-D5) |
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/content/ko/workflow-commands/_index.md:42:| [`/moai design`](./moai-design) | 조건부 | manager-design | - | UI SPEC을 위한 디자인 단계 협업 |
74ad197380:docs-site/content/ko/workflow-commands/_index.md:66:| [`/moai design`](./moai-design) | 디자인 단계 협업 (manager-design D1-D5) |
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/content/zh/workflow-commands/_index.md:62:| [`/moai design`](./moai-design) | 设计阶段协作(manager-design D1-D5) |
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:12:"moai-design":
74ad197380:docs-site/data/menu/main.yaml:208:        ref: /workflow-commands/moai-design
```

- Reading: nine lines are under `docs-site/content`, one is in `docs-site/data/menu/main.yaml` (outside the twelve files; O-2). The premise check (content and READMEs only) did not search `docs-site/data`.

### C-6 — page titles are slash forms (RED-now; AC-SWF-006 rewrite part)

- Command: `git grep -n -E "^title: /moai" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/moai-design.md'`
- Exit code: 0. Tree: `74ad197380`.
- Stdout (verbatim; 4 lines):

```
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:2:title: /moai design
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:2:title: /moai design
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:2:title: /moai design
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:2:title: /moai design
```

### C-7 — no routing workflow names the design workflow as an entry (supporting; Q13 determination)

- Command: `git grep -n -i -E "workflows/design|design phase|manager-design|plan.{1,3}design" 74ad197380 -- .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/sync.md .claude/skills/moai/SKILL.md`
- Exit code: 1 by git grep semantics (no match; not displayed by the tool, gap G-1). Stdout: no output.
- Reading: none of the four routing files names the design workflow. The design route is described in `.claude/rules/moai/workflow/spec-workflow.md` § Conditional Design Route and `.claude/skills/moai/workflows/design.md`, and `.claude/rules/moai/design/constitution.md` records `/moai design` as no longer a routed subcommand. With C-8 and C-9 this is the basis of "no current equivalent exists". The positive control is C-9 (same form, matching file).

### C-8 — registered slash commands (supporting)

- Command: `git ls-tree -r --name-only 74ad197380 -- .claude/commands/moai`
- Exit code: 0. Tree: `74ad197380`.
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

### C-9 — the design workflow is not user-invocable (supporting)

- Command: `git grep -n -E "user-invocable" 74ad197380 -- .claude/skills/moai/workflows/design.md`
- Exit code: 0. Stdout (verbatim): `74ad197380:.claude/skills/moai/workflows/design.md:11:user-invocable: false`

### C-10 — the count sentences on `_index.md` (supporting; open item O-4)

- Command: `git grep -n -E "subcommands|서브커맨드|サブコマンド|子命令" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'`
- Exit code: 0. Stdout (verbatim; 17 lines):

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:50:## All subcommands (16)
74ad197380:docs-site/content/en/workflow-commands/_index.md:52:The `/moai` orchestrator routes 16 subcommands. This (workflow) section covers the SPEC 3-Phase lifecycle commands, and the [utility commands](/en/utility-commands/) section covers automation, fix-loop, code-management, and feedback commands.
74ad197380:docs-site/content/ja/workflow-commands/_index.md:50:## 全サブコマンド (16 個)
74ad197380:docs-site/content/ja/workflow-commands/_index.md:52:`/moai` オーケストレーターは 16 個のサブコマンドをルーティングします。このセクション (ワークフロー) は SPEC 3-Phase ライフサイクルのコマンドを、[ユーティリティコマンド](/ja/utility-commands/) セクションは自動化・修正ループ・コード管理・フィードバックのコマンドを扱います。
74ad197380:docs-site/content/ja/workflow-commands/_index.md:56:| サブコマンド | 目的 |
74ad197380:docs-site/content/ja/workflow-commands/_index.md:67:| サブコマンド | 目的 |
74ad197380:docs-site/content/ja/workflow-commands/_index.md:102:自然言語でそのままリクエストしてもかまいません。`/moai "ログインのバグを直して"` のようにサブコマンドなしで入力すると、**Analyze-First ルーティング** が意図を分析して適切なワークフローに自動接続します。
74ad197380:docs-site/content/ko/workflow-commands/_index.md:53:## 전체 서브커맨드 (16개)
74ad197380:docs-site/content/ko/workflow-commands/_index.md:55:`/moai` 오케스트레이터는 16개 서브커맨드를 라우팅합니다. 이 섹션(워크플로우)에서는 SPEC 3-Phase 라이프사이클 명령어를 다루고, 자동화·수정 루프·코드 관리·피드백 명령어는 [유틸리티 명령어](/ko/utility-commands/) 섹션에 있습니다.
74ad197380:docs-site/content/ko/workflow-commands/_index.md:59:| 서브커맨드 | 목적 |
74ad197380:docs-site/content/ko/workflow-commands/_index.md:71:| 서브커맨드 | 목적 |
74ad197380:docs-site/content/ko/workflow-commands/_index.md:105:자연어로 바로 요청해도 됩니다. `/moai "로그인 버그 고쳐줘"`처럼 서브커맨드 없이 입력하면 **Analyze-First 라우팅**이 의도를 읽고 알맞은 워크플로우로 넘겨줍니다.
74ad197380:docs-site/content/zh/workflow-commands/_index.md:50:## 全部子命令(16 个)
74ad197380:docs-site/content/zh/workflow-commands/_index.md:52:`/moai` 编排器路由 16 个子命令。本节(工作流)讲解 SPEC 3-Phase 生命周期命令,[实用命令](/zh/utility-commands/)一节讲解自动化·修复循环·代码管理·反馈命令。
74ad197380:docs-site/content/zh/workflow-commands/_index.md:56:| 子命令 | 目的 |
74ad197380:docs-site/content/zh/workflow-commands/_index.md:67:| 子命令 | 目的 |
74ad197380:docs-site/content/zh/workflow-commands/_index.md:102:也可以直接用自然语言发出请求。像 `/moai "帮我修复登录 bug"` 这样不带子命令输入时,**Analyze-First 路由** 会分析意图并自动连接到合适的工作流。
```

- Reading: the stated count is 16 and the registered count is 17 (C-8); the en subcommand tables hold 6 workflow rows and 10 utility rows. Deleting the design row leaves 15 table rows beside the number 16. This is a consistency effect, not a slash citation (O-4).

## B.0 Local verification recipes

- **R-1 (marker branch).** C-1 with the marker negated: `git grep -c -P "/moai[: ]design(?![a-zA-Z0-9-])(?!.*<MARKER>)" <closing-sha> -- <the three pathspecs>`. `<MARKER>` is recorded in progress.md §E.2 after Q14. Expected: no output, exit 1.
- **R-3 (same change).** `git show --stat --format= <closing-sha>` lists the four `_index.md` paths in one commit.
- **R-4 (preservation guard).** `git diff --name-only 74ad197380 <closing-sha> -- docs-site README.md README.ko.md README.ja.md README.zh.md` expected to print only the paths of spec.md § Scope. Mutants: append a line to `docs-site/data/menu/main.yaml`, or to `docs-site/content/en/workflow-commands/moai-run.md`, in a scratch copy of the commit; the command prints the extra path.
- **R-5 (`_meta.yaml` parse and key set).** For each locale: `yq 'keys' docs-site/content/<L>/workflow-commands/_meta.yaml`, compared with `git ls-files docs-site/content/<L>/workflow-commands`; every key except `index` and `new_items` must name an `<key>.md` that exists.
- **R-6 (page structure parity).** `git grep -c -E "^## " <closing-sha> -- ':(glob)docs-site/content/*/workflow-commands/moai-design.md'` prints one equal count for the four locales.

## D. Definition of done

This child closes when AC-PRC-001, AC-SWF-002, and AC-SWF-003 hold on its committed tree, AC-SWF-004 holds as a guard, and the Q14 branch criteria (AC-SWF-005 or AC-SWF-006) hold for the answered branch. The parent imports the child's criteria by reference at its M6 (parent plan §B.3).

| AC | REQ | Class | Imported at the parent's M6 as |
|----|-----|-------|--------------------------------|
| AC-PRC-001 | REQ-PRC-001 | release-blocking (gated on Q14) | this child's AC-PRC-001 at its commit SHA |
| AC-SWF-002 | REQ-PRC-001, REQ-SWF-002 | release-blocking (gated on Q14) | this child's AC-SWF-002 |
| AC-SWF-003 | REQ-SWF-002, REQ-SWF-003 | release-blocking | this child's AC-SWF-003 |
| AC-SWF-004 | REQ-SWF-007 | guard | re-run as an observation |
| AC-SWF-005 | REQ-SWF-004 | release-blocking if Q14 = remove, else not applicable | this child's AC-SWF-005 or "not applicable" |
| AC-SWF-006 | REQ-SWF-005, REQ-SWF-006 | unadopted until Q14 has a verdict | this child's AC-SWF-006 for the answered branch |

Common gates: `moai spec lint --strict` on this directory reports no finding (recorded in `progress.md`); `git status --short` prints nothing after the verification runs; the deciding command and output lines of every claim are carried into the lane's verdict file.

## E. Edge cases

- The ko `_index.md` holds a design row in the summary table (line 42) that the other locales do not have; deleting only the subcommand-table row leaves a stale ko row (AC-SWF-003 Then 1).
- The ja and zh pages are 68 non-blank lines against 82 in en and ko; the usage example block (`> /moai design`) is absent from ja and zh. A rewrite or marker answer must bring the four to one structure (AC-SWF-006 Then 3).
- A page that survives Q14 is no longer linked from `_index.md` after M2 (the row is deleted by the Q13 verdict). Reachability then rests on the sidebar data file; whether to add a non-slash pointer is part of the Q14 answer.
- The Hugo menu template hides an entry whose `ref` page does not exist (`site.GetPage` guard in `layouts/partials/menu.html` lines 31 and 77), so a removed page leaves `main.yaml` stale but does not break the build; this was read from the template, not run (G-4).

## F. Quality gates (TRUST 5 mapping)

- Tested: each criterion has a RED-now cell or an explicit reason it has none (AC-SWF-004 guard, AC-SWF-006 marker part), a stated green output, and a mutant.
- Readable: the criteria name the row, file, and locale they count.
- Unified: edits follow the ko→en→ja/zh chain of `.moai/docs/docs-site-i18n-rules.md` §17.3 and the four-locale same-change obligation (Q17 verdict).
- Secured: documentation only; no secret, URL, or executable content changes. The URL blacklist and Mermaid TD-only rules are untouched because no URL or diagram is added.
- Trackable: commits carry the card id `t1727` and the SPEC id; the ledger is carried into the verdict file.

## G. Gaps (not observed)

- G-1 — The shell tool does not display the exit status of a no-output command. C-7 and every green-path "prints nothing" cell take exit 1 from git grep semantics; they were not captured with a wrapper (the worktree guard refuses compound forms).
- G-2 — No Hugo build, link check, or i18n parity script was run on a tree with the edits; the change does not exist yet. The GATE jobs will measure it later.
- G-3 — The structure curator's behavior on the four `_meta.yaml` edits was not run; plan.md §A cites its definition only.
- G-4 — The `menu.html` guard behavior (hidden entry for a missing page) was read from the template source, not rendered.
- G-5 — Translation fidelity of any rewritten text (Q14 = rewrite) is not machine-checked; it is a review item.
- G-6 — The marker token for Q14 = marker is undefined; whether the GATE checker exempts it depends on SWEEPE's Q24 token or a named exemption in GATE's contract C-7 (neither committed at the time of writing).
- G-7 — The parity-baseline file's consumer was not identified: a repository search for its name found no script or workflow that reads it among the paths searched in this pass (O-3).
- G-9 — The grep patterns do not cover every possible spelling of the design family (for example extra spaces or a line break between `/moai` and `design`); no such spelling was found by reading the 27 matching lines and the page headings, but the sweep is a pattern, not a parse.
- G-8 — Sibling children's own file lists are not in this tree; non-overlap was read from the parent table only.

## H. Residual risk

- A pattern that names the registered verbs (C-2) is stale the day a command is added or removed; the cell is pinned to its tree and re-run, not trusted later.
- The Q13 determination "no current equivalent exists" is measured on four routing files and the command list; a non-slash entry to the design route exists in rules and a skill, so an operator who reads "equivalent" more widely could ask for a non-slash pointer; that is covered by the Q14 answer.
- Deleting the `_index.md` rows while the page survives leaves a page that no index links; reachable through the sidebar only.
- Korean remains one `/moai` row above the other locales after the work (the `/moai goal` summary row); a reader comparing counts may misread this as drift.
