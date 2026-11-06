# t2022 verification-only pass at release/main-batch 650b5d27f2

Scope: docs vs. actual CLI at `650b5d27f2`. No product or doc edit. Source read through `git archive` into the session scratchpad (`.../scratchpad/v650`), never from or into `main-batch`. Lease: `moai factory next --card t2022 --run tnwnit` once, accepted (stage=-, worktree=t2022). Tree toplevel: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2022, `git status --short` empty before the pass.

## Claim / Evidence / Baseline

### 1. Hugo builds without warnings
- First attempt: `hugo --source <v650>/docs-site --destination <scratch>` exit 1: `failed to load Git data: fatal: not a git repository` (the archive has no .git; the site config reads git info). Not a docs defect; recorded as Gap G1.
- Rerun, plain: `env HUGO_ENABLEGITINFO=false hugo --source <v650>/docs-site --destination <scratch>/v650-hugo-out` -> built, no WARN/ERROR line in the output; `Total in 16507 ms`; pages KO 186 / EN 184 / JA 184 / ZH 184; Aliases 6/5/5/5; Static files 264 each.

### 2. Notation counts (README x4 + docs-site/content, 4 locales)
- Strict leftover `moai (cc|glm)( [^` ]+)* -l`: `grep -rnE` -> no output (0 hits).
- `moai cc/glm -a` per README: README.md 7, README.ko.md 7, README.ja.md 7, README.zh.md 7 (`grep -rncE`).
- `--lane` hits (`grep -rnE -- '--lane'`): only (a) README x4 line 74, the `moai factory relaunch` option list plus "`-a` takes no argument, a `--lane` you pass is not carried" and (b) `moai gtd engage ... --lane lane-10` and its sentence in the 4 `utility-commands/moai-gtd.md` pages. No `--lane` as a launcher entry token remains.
- Remaining `moai codex -l`: README x4 lines 62/74, `advanced/factory-mode.md` x4 line 26, `cli-reference/launchers.md` x4 line 34 — intended, matches `moai codex --help` (`-l, --lane  join the active factory as the next lane (CLI only)`).

### 3. 4-locale file existence and section counts (`grep -rcE '^## '`)
- Present in all four locales: `_index.md` (5/5/5/5), `advanced/factory-mode.md` (9/9/9/9), `advanced/manager-lead.md` (12/12/12/12), `cli-reference/launchers.md` (7/7/7/7), `workflow-commands/moai-goal.md` (11/11/11/11); READMEs 12/12/12/12.
- MISMATCH M1: `workflow-commands/moai-run.md` en 12, ko 12, ja 11, zh 11. en has `## Factory Mode — carry a card all the way through (v3.2)` (line 536) that ja (headings at 13..531) does not carry; zh count is 11 too (heading list not printed for zh). `git diff --stat 742626a572 fb6ed75ce4 -- ja/zh moai-run.md` is empty: this card did not touch ja/zh `moai-run.md`; the gap predates it.

### 4. Docs vs binary help
Binary: `go -C <v650> build -o <scratch>/moai-650 ./cmd/moai` (exit 0), then `moai-650 cc --help`, `glm --help`, `factory relaunch --help`, `codex --help`.
- Matches: `-f, --factory` takes no argument; `-a` takes no argument and claims the next free `lane-<n>`; `--leader <name>` only with `-a` (legacy spelling refused); `-f` beside `-a` is an error; the former `-l/--lane` is removed and refused with a line naming `-a`; `-k` retired, refused with a line naming `-f` and `-a`; `moai glm -a`/`-f` same; `moai codex -l, --lane` kept; `factory relaunch` options `--provider --lane --run --from-run --dry-run`, `--lane` recorded for the notice only because `-a` takes no argument — README line 74 and launchers/factory-mode pages say the same.
- Documented flags confirmed in source (grep): `--factory-run` (factory.go:164), `--no-auto-dispatch` (factory.go:87), `clear-when-full` in internal/config/envkeys.go and internal/cli/factory_card.go.
- MISMATCH M2: docs `cli-reference/launchers.md` usage line is `moai cc [-p profile] [-w [name]] [-- claude-args...]` (en line 19); the binary's USAGE is `moai cc [-p profile] [-f | -a] [-- claude-args...] [--flags]`.
- MISMATCH M3: the docs `moai cc` flag table (en lines 24-38) has no row for `--spawn` and `--branch <existing>`, which `moai cc --help` lists; `moai glm` help also lists them.
- Not verified, not a confirmed mismatch: docs say `-a` "is refused when no factory is running"; help says that when the run's record is missing/retired and a live leader exists, the join restores the run. Source shows a refusal path (factory.go:577) but the exact no-factory case was not executed (new session forbidden).

## Mismatch list (report only, nothing fixed)
- M1 ja/zh `moai-run.md` miss the `Factory Mode` section that en/ko carry (11 vs 12 `##`).
- M2 launchers usage line omits `[-f | -a]`.
- M3 launchers flag table omits `--spawn` and `--branch`.

## Gaps
- G1 Hugo ran with `HUGO_ENABLEGITINFO=false` because the archive has no git history; page lastmod from git is not exercised. A build in a real git checkout at 650b5d27f2 was not run. The first failing hugo run is not evidence of a docs warning.
- G2 The binary was built with a plain `go build` (no Makefile ldflags): it carries no version/commit stamp, so its provenance is the source tree `650b5d27f2` only; no `moai version` comparison was made.
- G3 No role or session was launched; the `-a` no-factory refusal text and the `relaunch --dry-run` output were not executed.
- G4 `git merge-base --is-ancestor fb6ed75ce4 650b5d27f2` printed nothing and the tool reported no error; the exit code itself was not displayed.
- G5 Locale parity beyond section (`##`) counts (paragraph/table content) was not compared; zh `moai-run.md` heading list was not printed.
- G6 No guard refusal occurred in this pass. `internal/cli` full run not performed.

## Residual risk
- Docs and help drift again whenever the launcher help changes; this pass is one snapshot at 650b5d27f2.
- M1-M3 were unrepaired at the time of this pass; see the repair section below.

## Repair pass (leader instruction: repair M1-M3 in this card)
- Base: release/main-batch 650b5d27f2 merged into the card tree (merge commit 9d76aaf975), then manager-docs edits, lane read the diff, commit e79daf76a2 (6 files under docs-site/content, +40/-4): launchers.md x4 (usage line `[-f | -a]`; rows `--branch <existing>` and `--spawn`), ja/zh moai-run.md (Factory Mode section added before the related-documents heading).
- M3 flag rows rest on `moai cc --help` of a binary built from this tree (writer's `go build`, plain, unstamped); `moai glm` has no flag table in launchers.md, so none was added. Not added: `--factory-run`, `--clear-policy`, `--no-auto-dispatch` are already in the docs table but are absent from `moai cc --help`; source defines them (factory.go:87, :164; clear-when-full in internal/config/envkeys.go, internal/cli/factory_card.go). Left unchanged.
- Re-verification in this git-backed tree (plain commands, lane-observed): `hugo --source <tree>/docs-site --destination <scratch>` -> no WARN/ERROR, pages KO 186 / EN 184 / JA 184 / ZH 184, Total 6119 ms (this closes G1 for the repaired tree; the first failing hugo run on the archive in this file stays as recorded); `grep -c '^## '`: moai-run.md en/ko/ja/zh 12/12/12/12, launchers.md 7/7/7/7; URL blacklist grep 0 hits; Mermaid LR/RL grep 0 hits; body-emoji scan on the 6 edited files 0 hits.
- Gaps: not run: the per-page `.locale-parity-baseline` ratchet over the whole tree and the README parity re-count; only the touched pages were re-counted. Not changed (outside scope): the related-documents list in ja/zh moai-run.md lacks the `/moai goal` and Factory Mode bullets that en/ko carry (writer's finding, lane did not re-verify it).
