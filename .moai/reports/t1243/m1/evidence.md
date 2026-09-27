# t1243 M1 — run-phase measurements (AC-IFU-021, AC-IFU-022)

Measured 2026-09-27 from worktree `.claude/worktrees/t1243`, branch `WT-instruction-files`, HEAD
`7fe658815`. Tools: `claude 2.1.283 (Claude Code)`, `codex-cli 0.157.0`. Every fixture was created
in this run under the session scratchpad
(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/<session>/scratchpad/`), never inside this
repository. Every probe ran as one compound invocation with
`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && ...`
and was bounded by `timeout`. No figure below is carried over from M0 or any document.

## M1a — real-worktree ancestor discovery (AC-IFU-021)

### Fixture

- `m1a/primary`: `git init -b main`; committed `CLAUDE.md` (`WT_CONTROL = OSCAR4`, positive control
  present in every tree) and a `.gitignore` listing `AGENTS.local.md`, `CLAUDE.local.md`,
  `.claude/worktrees/`.
- Untracked, gitignored, primary root only: `AGENTS.local.md` = `LOCAL_AGENTS_TOKEN = PAPA6`,
  `CLAUDE.local.md` = `LOCAL_CLAUDE_TOKEN = QUEBEC2`.
- Two REAL linked worktrees, created by `git worktree add`:
  - nested (this repository's geometry): `git -C <primary> worktree add -q -b wt-nested .claude/worktrees/wt`
  - sibling (contrast): `git -C <primary> worktree add -q -b wt-sibling ../sibling-wt`
- `.git` in both worktrees is a FILE, not a directory:

```
-rw-r--r--@ 1 goos  wheel  139 Sep 27 19:42 primary/.claude/worktrees/wt/.git
-rw-r--r--@ 1 goos  wheel  147 Sep 27 19:42 sibling-wt/.git
gitdir: .../scratchpad/m1a/primary/.git/worktrees/wt
```

- Neither worktree contains `AGENTS.local.md` or `CLAUDE.local.md` (`ls -A` of each: `.git`,
  `.gitignore`, `CLAUDE.md` only). No `CLAUDE.md` / `CLAUDE.local.md` / `AGENTS.md` exists in any
  of the five ancestors above `m1a/` (`ls` scan printed nothing).

### Probe prompt (identical for all three runs)

`claude -p "Answer only from your loaded instructions, run no tools. Three lines: (1) exact value of WT_CONTROL (write NOT_PRESENT if absent), (2) exact value of LOCAL_AGENTS_TOKEN (write NOT_PRESENT if absent), (3) exact value of LOCAL_CLAUDE_TOKEN (write NOT_PRESENT if absent)." --model claude-haiku-4-5-20251001`
— abbreviated `<probe>` below; each wrapped in `timeout 180`.

### Results (verbatim; raw files `m1a-*.out` / `m1a-*.err` beside this file)

```
$ cd .../m1a/primary/.claude/worktrees/wt && timeout 180 claude -p <probe>
exit=0
1. OSCAR4
2. NOT_PRESENT
3. QUEBEC2
stderr: (empty)

$ cd .../m1a/sibling-wt && timeout 180 claude -p <probe>
exit=0
OSCAR4
NOT_PRESENT
NOT_PRESENT
stderr: (empty)

$ cd .../m1a/primary && timeout 180 claude -p <probe>
exit=0
1. WT_CONTROL = OSCAR4
2. LOCAL_AGENTS_TOKEN = NOT_PRESENT
3. LOCAL_CLAUDE_TOKEN = QUEBEC2
stderr: (empty)
```

### Reading

- Positive control `OSCAR4` present in all three runs — every session loaded its instructions.
- `CLAUDE.local.md` IS discovered from a real nested linked worktree (`QUEBEC2`, run 1). The
  sibling worktree does NOT get it (run 2), so the mechanism is a directory-ancestor walk that
  happens to cross the primary root because `.claude/worktrees/<name>` lies inside it; it is not
  git-worktree-aware.
- `AGENTS.local.md` is NOT discovered anywhere — not from the nested worktree (run 1), and not even
  at the primary root itself (run 3). Claude Code does not discover that filename at all.
- design.md §A.5 branch: **row 2 — "Discovery confirmed for `CLAUDE.local.md` but NOT for
  `AGENTS.local.md`"** → Option 3 degrades to Option 1; recorded as a known limitation, not retried.

## M1b — Codex discovery of `AGENTS.local.md` (AC-IFU-022)

### Fixture (`m1b/fixture`, `git init -b main`, nothing committed)

| File | First line | Final line |
|---|---|---|
| `AGENTS.md` | `CONTRACT_HEAD` | `CONTRACT_TAIL` |
| `AGENTS.local.md` (listed in `.gitignore`) | `LOCAL_HEAD` | `LOCAL_TAIL` |
| `CLAUDE.md` | `# Claude layer` (body: `@AGENTS.md`, `@AGENTS.local.md`) | — |

`~/.codex/config.toml` carries no `fallback` / `project_doc` key
(`grep -n -i "fallback\|project_doc" ~/.codex/config.toml` printed nothing); `CODEX_HOME` unset.
Verb re-confirmed against this codex-cli version: `codex debug --help` lists
`prompt-input  Render the model-visible prompt input list as JSON`.

### Result (verbatim)

```
$ cd .../m1b/fixture && codex --version && timeout 120 codex debug prompt-input > ../prompt-input.json 2> ../prompt-input.err
codex-cli 0.157.0
exit=0
   40358 ../prompt-input.json
       0 ../prompt-input.err
CONTRACT_HEAD count=1
CONTRACT_TAIL count=1
LOCAL_HEAD count=0
LOCAL_TAIL count=0
AGENTS.local mentions: 0
Fixture contract body: 1
Maintainer-only: 0
```

Counts are `grep -o <token> prompt-input.json | wc -l` over the whole render. Full render sha256:
`eea9b78fc070dc2b1a1d0a5cdcc1767685d519879aba847c90837a0d0c3f2a09` (40,358 B). The full JSON is NOT
committed: its `<INSTRUCTIONS>` block also carries the operator's personal global
`~/.codex/AGENTS.md`. The project-doc segment alone is committed as `m1b-project-doc-excerpt.txt`:

```
--- project-doc ---\n\nCONTRACT_HEAD\n\n# Fixture contract\n\nThis is the fixture standing contract body.\n\nCONTRACT_TAIL\n\n</INSTRUCTIONS>
```

### Reading (positive control first)

- `CONTRACT_HEAD` present → the run is conclusive.
- `LOCAL_HEAD` absent (and `LOCAL_TAIL`, and any `AGENTS.local` mention) → **Codex does not
  discover `AGENTS.local.md`; the design's budget arithmetic holds.** Not a blocker.
- `CONTRACT_TAIL` present → the fixture `AGENTS.md` was not truncated.

## Gaps

- M1a reads what a Haiku model repeated, not a rendered context dump; Claude Code has no
  `prompt-input` equivalent used here. The sentinels are unique random-word pairs and the positive
  control held, but a model declining to echo a loaded token would read as `NOT_PRESENT`.
- M1a ran headless only; interactive sessions were not measured. One run per geometry, no replication.
- M1b ran with no `developer_instructions` injection (plain `codex`, not the moai Codex launcher),
  so it answers filename discovery only — which is the question AC-IFU-022 asks.
- Only codex-cli 0.157.0 and claude 2.1.283 were measured.
