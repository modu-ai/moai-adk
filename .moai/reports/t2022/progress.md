# t2022 progress record (lane-4, run tnwnit)

- 2026-11-03 intake: leased via `moai factory next --card t2022 --run tnwnit`; tree /Users/goos/.moai/worktrees/projects/moai-adk-go/t2022, branch WT-docs-site-readme.
- base: merged release/main-batch 5a3ccd1a4d into the card tree -> HEAD 742626a572 (merge commit, card t2022).
- class: A-shaped docs card (no SPEC, plan skipped). Stage order: run (edit via manager-docs) -> card-review -> merge-ready; STOP at merge-ready, no push/PR/paid audit (card text).
- scope found by grep (candidate-based): 26 files = README x4 + docs-site/content/{en,ko,ja,zh} (_index, advanced/factory-mode, advanced/manager-lead, cli-reference/launchers, utility-commands/moai-gtd, workflow-commands/moai-run). Card says about 22 launcher-guidance files; moai-gtd `--lane` is an unrelated `gtd engage` flag (expected excluded).
- semantics of `-a` read from internal/cli/factory.go parseFactoryFlag: takes no argument, lane-role entry, long `--agent` is deliberately NOT an entry token, `-a` conflicts with `-f`.
- decision record (auto-semantics section 10): replace `moai cc|glm -l` with `-a`; drop "`--lane` long form" claims for the launcher (no long form); keep `moai codex -l`, `lane-<n>` labels, `moai gtd engage --lane`, and the `moai factory relaunch --lane` option wording unless it names the launcher entry token.
- watchdog: recurring recheck cron 2eca4083 armed (every 5 minutes).

## Gaps (refusals and reshaped reruns, recorded per leader instruction)
Worktree-guard refusals ("too complex to verify that it stays inside the worktree ... Split it into plain, separate commands"; R2/R3 said perl input was computed at runtime). Commands are the writer's own elided quotes:
- Lane L1: Bash heredoc writing this progress.md chained with a git rev-parse -> refused; redone as a Write-tool write of the same file (changed-shape rerun).
- R1: `FILES=$(grep -rlE ... | grep -v moai-gtd.md) && ... perl -CSD -i -pe '...' $FILES && git diff --stat | tail -3 && grep ...`
- R2: `cat <scratchpad>/files.txt | tr ...; echo; xargs perl -CSD -i -pe '...' < <scratchpad>/files.txt`
- R3: `perl -CSD -i -pe "$(cat $S/p2.pl)" <files>; grep ...`
- R4: `{ echo 'while (<>) {'; cat p2.pl; echo 'print; }'; } > p3.pl && perl -CSD -i p3.pl <files>; grep ...`
- R5: `{ echo ...; cat ...; echo ...; } > p3.pl; wc -l < p3.pl`
- R6: `hugo ... ; git diff --name-only | sed ...`
Writer reruns in changed shape: after R1 a plain `perl -CSD -i -pe 's/moai codex -l/.../; s/(?<![\w-])-l(?![\w-])/-a/g; ...' <explicit files>` (this edited the files); p2.pl run (no-op); p3.pl via Write tool run twice (no-op, then with `use utf8` = second pass removing `--lane` clauses, ko particles); `hugo --source docs-site --quiet --destination <scratchpad>/hugo-out` alone and `git diff --name-only` alone. Scratch files stay in the session scratchpad.
Evidence handling: the hugo result is a rerun result and is NOT cited as evidence; the lane does not rerun hugo (leader's final-SHA job). Hugo build b3bf2magq was not killed. Unverified in-lane: locale parity, leftover `-l` grep, over/under-reach of the blanket substitution; the leader's independent review (.moai/reports/rc40/refusal-results-review.md) judged the 26 files fit.

## Task state
- hold released by leader (independent review: fit). Edit of progress.md Gaps done with the Edit tool.
- run DONE: committed the existing 26-file diff as fb6ed75ce4 (`[t2022] docs(docs-site): ...`), staged by explicit pathspec, HEAD 742626a572 and branch WT-docs-site-readme re-read before commit; `git diff --cached --shortstat` = 26 files, 161 insertions, 161 deletions. No perl/script edits rerun.
- leader update (user instruction renewed): single plain commands in the owned tree are normal; lease renewed with `moai factory renew t2022` -> until 2026-11-03T03:55:01Z.
- fresh verification at fb6ed75ce4 (one plain command each, this run): `hugo --source <tree>/docs-site --destination <scratchpad>/hugo-out-fb6ed75ce4` exit 0, no WARN/ERROR lines in the output, pages KO 186 / EN 184 / JA 184 / ZH 184. Leftover grep `moai (cc|glm)... -l|`-l`|--lane` over README*.md + docs-site/content: remaining hits are only (a) `moai codex -l` and the `moai factory relaunch --lane` option in README*.md line 74 (kept on purpose), (b) `moai gtd engage ... --lane lane-10` in the 4 moai-gtd.md pages (unrelated gtd flag). No `moai cc -l` / `moai glm -l` hit remains.
- card-review DONE: .moai/reports/t2022/card-review.md. Tool verdict `fail` (advisory, static): 3 findings (P1 internal/hook/branch_guard.go:463, P2 internal/cli/web.go:152, P2 internal/cli/factory_release_route.go:280), all in Go code from the merged release/main-batch, none in t2022's 26-file docs diff; dispositioned not adopted, carried to the leader as information. Not recorded as a pass.
- state: merge-ready stop reached at HEAD fb6ed75ce4 (evidence files committed in the next commit); no push, no PR, no paid audit; `moai factory complete` NOT run (leader: stop at merge-ready).

## Follow-up: verification at 650b5d27f2 and M1-M3 repair
- verify-only pass: .moai/reports/t2022/verify-650b5d27f2.md (found M1 ja/zh moai-run.md section gap, M2 usage line lacks `[-f | -a]`, M3 flag rows missing).
- leader ordered the repair in this card: lease renewed with `moai factory renew t2022 --run tnwnit`; release/main-batch 650b5d27f2 merged (9d76aaf975); manager-docs edited 6 docs-site files; lane read the diff; commit e79daf76a2.
- re-verification and the second card-review (verdict `fail`, advisory, same 3 batch-code findings, not adopted) are recorded in verify-650b5d27f2.md and card-review.md.
- state: merge-ready stop at the evidence commit on top of e79daf76a2; no push, no PR, no paid audit; `moai factory complete` NOT run.

## R437 flag flip
Operator decision R437 (supersedes the card text): launcher entry flags are `-f` (leader) and `-l` / `--lane` (lane); `-a` / `--agent` is removed. Docs only; code is card t2021.

- Files changed (28, 4 locales in one commit): README.md, README.ko.md, README.ja.md, README.zh.md; docs-site/content/{en,ko,ja,zh}/ for `_index.md`, `advanced/factory-mode.md`, `advanced/manager-lead.md`, `cli-reference/launchers.md`, `workflow-commands/moai-goal.md`, `workflow-commands/moai-run.md` (ja/zh moai-run Factory Mode section included).
- Edits: `moai cc -a` / `moai glm -a` -> `-l`; backticked `-a` -> `-l`; usage line `[-f | -a]` -> `[-f | -l]`; launchers.md flag row `-a` -> `-l, --lane` (the only `--lane` row, 4 locales; no `-a/--agent` row existed); ko particles adjusted after the change (`-l`이/을/과/은). `-a lane-2` examples in launchers callouts -> `-l lane-2`.
- No "join as an agent" / "agent session" wording for the joining session was found in the 28 files (grep of EN/KO/JA/ZH forms), so no wording edits beyond the flag.
- Kept hits: `docs-site/content/*/claude-code/agentic/agent-view.md` (`claude --agent`, Claude Code's own flag); `cli-reference/worktree.md` and `worktree/guide.md` (`--agent-name`, `--agent-response`, `moai worktree` options); `advanced/hooks-guide.md` (`docker system prune -a`); `docs-site/static/install.sh` (`shasum -a 256`); `moai codex -l` and `lane-<n>` label text untouched.
- Verification (this run, tree WT-docs-site-readme): (1) `git grep -n -e 'moai cc -a' -e 'moai glm -a' -e '\[-f | -a\]' -- docs-site README*.md` printed nothing, exit 1. (2) per-file heading counts (`grep -c '^#'`) identical before/after for all 28 files; locale parity en/ko/ja/zh: _index 5/5/5/5, factory-mode 13x4, manager-lead 20x4, launchers 7x4, moai-run 47x4, README 81x4, moai-goal 20/21/20/20 (ko has 21 at HEAD too; pre-existing, unchanged). (3) `hugo --quiet` from docs-site: exit 0, zero output lines, zero WARN/ERROR (no HUGO_ENABLEGITINFO needed). (4) `git diff --stat`: 28 files changed, 169 insertions(+), 169 deletions(-).
- Gaps: no mermaid/emoji/URL-blacklist re-scan (no such content added); not read in rendered form; ko particle choice reviewed by grep only.
- Residual risk: the ja/zh/ko prose may still say "-a" in a form not matching the greps (e.g. unbackticked), checked by a boundary regex over the touched files only. Code side (t2021) must actually remove `-a`, else docs lead the binary.
- Tooling note: the worktree guard refused compound git/perl commands with computed args; a perl script file was used instead of an inline loop.

## Codex lane guidance removal
Fact (leader, measured against a build of release/main-batch): `moai codex -l` and `moai codex status` only print help and exit 0; `codex_cmd.go` lost its RunE, so `moai codex [cli|status|app]` and `moai codex -l` are not working commands. Docs must not present them as runnable. Scope: docs only; Go code and help strings belong to t2029.

Hit table (hits classified from the `git grep -n 'moai codex'` listing before the edit; script `scratchpad/codex_rm.pl`, literal replacements, applied to 32 files):

| Where (all 4 locales unless noted) | Action | Reason |
|---|---|---|
| README x4 getting-started code line `moai codex -l` | removed line | runnable Codex lane join |
| README x4 "moai codex has no leader entry ..." sentence | removed sentence | Codex lane guidance |
| README x4 "GPT models run through moai codex" | removed sentence | runnable command presented |
| README x4 "Which backend" `moai codex -l` clause | dropped clause, sentence kept | Codex lane |
| README x4 relaunch paragraph "Codex only moai codex -l" | dropped clause | Codex lane |
| README x4 "GPT gateway withdrawn" note | `moai codex` -> "Codex CLI" | command not runnable; fact (GPT via native harness) kept |
| README x4 command table row `moai codex [cli|status|app]` | removed row | non-working command |
| `_index.md` code line + "cc, glm, codex" backend list | line removed; list now cc, glm | Codex lane |
| `advanced/factory-mode.md` entry-forms row clause, "no leader entry" bullet, `$ moai codex -l` line, refused-forms `-f on moai codex` row, `moai gpt` row tail, "Codex lanes take no policy" | removed / trimmed | Codex lane or non-working command |
| `cli-reference/launchers.md` `-l, --lane` row tail and callout clause | trimmed | Codex lane / `-f` on `moai codex` |
| `advanced/claude-md-guide.md` table cell and Mermaid node (TD kept) | `moai codex` -> `Codex` / `Codex launcher` | descriptive, command name not runnable |
| `advanced/harness-learning.md`, `getting-started/quickstart.md`, `claude-code/context-memory/memory.md` | `moai codex` -> `Codex` | descriptive, command name not runnable |

Kept hits (8): `advanced/codex-dual-harness.md` lines 33 and 37 in en/ko/ja/zh. Reason: the conceptual dual-harness page, explicitly excluded by the task; the lines describe which files the Codex launcher reads (`AGENTS.local.md` / `CLAUDE.local.md` fallback, `developer_instructions`) and do not tell the reader to run anything. Line 33 does list route names (bare, cli, app, --spawn, -w) as launcher behavior, so it still names `moai codex` as if it works.

Verification (this run, tree WT-docs-site-readme):
1. `git grep -n 'moai codex' -- docs-site README.md README.ko.md README.ja.md README.zh.md` after the edit: only the 8 codex-dual-harness.md lines above (en/ko/ja/zh x lines 33, 37).
2. `git grep -n -e 'moai codex -l' -e 'moai codex status' -e 'moai codex cli' -e 'moai codex app' -- docs-site README*.md`: no output, exit 1. A looser grep for Codex near lane words (lane/レーン/레인/泳道) over README*.md and docs-site/content: no output.
3. Heading counts (`git grep -c '^#'`) over the 4 READMEs and the 7 changed page types x 4 locales: before/after files diff exit 0 (identical). README 81x4; `_index` 5x4; factory-mode 13x4; launchers 7x4; harness-learning 7x4; memory 18x4; claude-md-guide 46/21/46/46 (en/ko/ja/zh) and quickstart 32/31/32/32 are pre-existing mismatches, unchanged.
4. `hugo --quiet` from docs-site: exit 0, no output; a separate non-quiet `hugo` run piped to `grep -ci 'WARN\|ERROR'` counted 0.
5. `git diff --stat` before commit: 32 files changed, 61 insertions(+), 85 deletions(-) (code diff, before this section).

Gaps: ko/ja/zh phrasing reviewed from the diff text only, no humanize skill pass and no rendered-page read; no mermaid/emoji/URL-blacklist rescan (only the Mermaid node label text changed, direction still TD); `codex_review`/`codex_audit` descriptions were not searched separately beyond the `moai codex` grep; hits in Go help text (t2029) untouched.
Residual risk: codex-dual-harness.md still names `moai codex` routes (8 lines) and can read as runnable until the command is restored or the page is revised by the leader's decision; mentions of Codex without the `moai codex` string (for example `--llm gpt`, harness choice) were not changed and may still imply a launcher; a first scripted pass left a stray "and" in README.md line 62, found in the diff and repaired by hand before commit.
