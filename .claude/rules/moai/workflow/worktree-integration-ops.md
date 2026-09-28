---
paths: "**/.claude/worktrees/**"
---

# Worktree Integration — Session Operations

> Loading scope: detail companion of `worktree-integration.md`, loading only on the `**/.claude/worktrees/**` trigger. The parent keeps the glossary, the worktree selection rules, the launcher catalogue, and the disposal contract; this file owns the operational procedures a session working *inside* a worktree needs — manual disposal beyond the automatic sweep, and the command-refusal catalogue of the worktree-session guard. Sections here are needed when worktree files are the subject; they are not needed when only agent definitions or team state are touched.

## Disposing a Worktree the Automatic Sweep Does Not Reach

Automatic disposal covers one shape only: the PR-merge auto-cleanup sweep enumerates `git worktree list` and treats a tree as a candidate solely when its branch carries the launcher's `WT-` prefix. Every other registered worktree — one named after the change it makes, one entered by hand, one whose branch was renamed — falls outside that sweep and stays on disk until someone disposes of it. That is the safe direction (nothing is removed unasked), but it is not a disposal plan.

**Worktree-ness is a property of the checkout, not of the branch name.** A branch-name glob finds only the trees named a particular way; `git worktree list` finds all of them. Any inventory of what is actually on disk therefore starts from the listing, never from a name pattern.

The shipped inventory is the `--stale` sweep's own evaluation, rendered as data:

```bash
moai worktree clean --stale --json
```

It emits one object per non-protected registered worktree, carrying the path, the branch, the keep-reason, and the four predicates behind that reason — dirty state, merge state, anchor state, and ignored-content state. It removes nothing: `--json` is a report, and it overrides `--yes` rather than combining with it. A predicate the sweep short-circuited before asking reads `not-checked`, which is deliberately distinct from `undetermined` — the latter means it asked and could not tell. Neither is a negative.

Read the report, then dispose of what it shows as removable:

```bash
moai worktree clean --stale        # preview: names the trees it would remove
moai worktree clean --stale --yes  # perform the removals
```

Both paths honour the same guards: a dirty tree, an unmerged branch, a tree anchoring a live session, a tree holding gitignored content that nothing regenerates, and a tree whose state could not be read are each kept and reported with the reason. The ignored-content guard matters because `git status --porcelain` and a non-forced `git worktree remove` both disregard gitignored files: without it a tree whose only remaining content is agent memory reads as clean and is destroyed silently. It shares one allowlist with the automatic sweep, so both agree on what is regenerable — runtime state, runtime-managed config, build output, test residue — and anything unclassified keeps the tree. Branches are never deleted — the commits stay reachable by branch name after the directory is gone. The merge comparison is against `origin/main` by default, the same ref the automatic sweep uses, so the two cannot reach opposite conclusions about the same tree; `--base` overrides it.

For a tree outside the sweep entirely, the manual path is unchanged: `git worktree unlock <path>` when a dead session's lock is still on it, then `git worktree remove <path>`. The unpushed-branch rule in `worktree-integration.md` § Terminology Glossary governs the timing in every case.

## Refused Commands in a Worktree-Isolated Session

### Which guard refused — read the message, it is the only discriminator

Two different guards refuse commands in the same worktree-isolated session, and nothing but the wording of the refusal tells them apart. Read the message first; everything else follows from it.

| The refusal says | Whose guard | Can it be changed here? |
|---|---|---|
| `Dangerous command blocked:` … | **This project's own** pre-tool hook, in `internal/hook/pre_tool.go` | **Yes** — it is this repository's code |
| … `too complex to verify that it stays inside the worktree` | **The Claude Code binary** | **No** — there is no source here that implements or configures it |

A reader who confuses the two searches the wrong place. Hunting through this repository's source for the worktree guard finds only prose quoting the message and never the guard; treating the dangerous-command refusal as an untouchable upstream behaviour leaves a local hook unexamined when it is in fact the thing refusing.

A third, separate guard also lives here: `internal/hook/branch_guard.go` defends branch-state mutation in the primary checkout. It neither implements nor configures the worktree guard, and editing it has no effect on a `too complex to verify` refusal.

### What has been observed to trip the worktree guard

The full refusal reads:

> This session is isolated in the worktree `<worktree-path>`, but this command is too complex to verify that it stays inside the worktree.

**This table is a record of what has been seen, not a specification.** It is explicitly non-exhaustive, and a handful of observations do not establish a general rule — the trigger condition was never narrowed, so do not infer from it that "complex commands are refused", and do not infer a tokenization mechanism from the shape of these cases.

| Observed trigger | Provenance | Notes |
|---|---|---|
| A quoted-delimiter heredoc (`<<'EOF'`) whose **body contains braces wrapping quoted key/value pairs** — a JSON line, for example | **First-hand** — measured by paired probes in a worktree-isolated session | Body size, pipes, backticks, and command substitution are **not** the trigger: a body of roughly 6.7 KB of prose was accepted, a ten-character JSON line in the same position was refused, and a command substitution in the body was folded correctly and passed |
| **Several mutation steps bundled into one compound command** | **Second-hand** — reported by another working lane, not measured here | Recorded because it carried the same refusal sentence; it has not been reproduced by the session that wrote this section |
| A heredoc whose **body names a git subcommand** — prose such as `run git merge --no-ff <sha>` or `git checkout -b <branch>` fed to a non-git command | **First-hand** — measured by paired probes in a worktree-isolated session at 2.1.251 | Refused with the variant sentence `… this command names git in a form too complex to verify …`. The same heredoc without git text passed, and the same git text passed when carried as a double-quoted argument or read from a file (`--stdin < <file>`). **This row was contradicted at 2.1.275** — see § Two observations disagree on the heredoc-carrying-git-text shape below; do not cite it as current behaviour |

**Gap**: the boundary between an accepted and a refused command is unmeasured in every row. Anything outside these observations is unknown, not permitted.

### Why the two heredoc delimiter forms differ

The distinction the observation above turns on is the delimiter quoting, and it
is bash semantics, not guard behavior:

- **A quoted delimiter (`<<'EOF'`) makes the body inert text.** Bash performs no
  expansion of any kind inside such a body — no parameter expansion, no command
  substitution, no arithmetic expansion, and no brace expansion. A brace there
  is a literal character and cannot be brace expansion. This is the same fact a
  guard relies on when it folds a command substitution appearing in such a
  body, as observed above.
- **An unquoted delimiter (`<<EOF`) makes the body live text.** Parameter
  expansion, command substitution, arithmetic expansion, and brace expansion
  all apply. No guard — and no reader of this file — may treat an
  unquoted-delimiter body as inert, and a refusal that errs on the side of
  caution there is correct behavior, not a defect.

The observed asymmetry is therefore narrow: in a position provably free of
expansion (the quoted-delimiter body), braces alone are treated as live, while
command substitutions in the same position are already folded.

The refusal half of this asymmetry was re-measured by paired probes in a
worktree-isolated session on this repository: the brace form was refused with
the `too complex to verify` sentence above before the command executed (the
target file was confirmed absent afterward), and the paired probe — the same
shape with a command substitution in the body — executed and passed with the
body preserved as a literal. This remains a record of observations, not a
specification; the boundary of the guard's analyzer is still unmeasured.

### The refusal's message shapes

The refusal is not one sentence. Seven distinct clauses have been seen — two of them first met while writing this section — and the difference matters because a reader who greps for the wording they remember concludes the guard did not fire.

Four shapes are pinned in this repository as quoted fixtures (`internal/hook/worktree_guard_refusal_test.go`); three more have been observed since and are not pinned anywhere:

| Shape | The clause after `…but this command` | Provenance | Pinned |
|---|---|---|---|
| Cross-tree redirect via `-C` | `redirects git to the shared checkout via -C` | First-hand, main session (card t529); re-measured first-hand at 2.1.275 (card t852) — byte-identical but for the path | `sampleGuardRefusalDashC` |
| Cross-tree redirect via `--git-dir` | `redirects git to the shared checkout via --git-dir` | First-hand, **background subagent** (card t529) | `sampleGuardRefusalGitDir` |
| Unverifiable command | `is too complex to verify that it stays inside the worktree` | First-hand, main session (card t529); re-measured first-hand at 2.1.275 (card t852) | `sampleGuardRefusalComplex` |
| Working-directory resolution | `'s working directory resolved to the shared checkout (<path>)` | **Card-quoted, never measured.** The quote is truncated and the continuation is unobserved — the fixture reproduces that truncation deliberately | `sampleGuardRefusalCwdCardQuoted` |
| Runtime-computed target | `points git at a directory computed at runtime (-C <path>)` | Second-hand — another lane, 2.1.275 (card t880); not reproduced here | — |
| Unverifiable git form | `names git in a form too complex to verify` | First-hand at 2.1.251 (recorded in the trigger table above); second-hand at 2.1.275 (card t880) | — |
| Unverifiable non-git command | `runs <command> with <argument> in a plain command, so what it runs cannot be shown not to be git` | First-hand at 2.1.275 (card t852) — see the reproduction note below | — |

**A seventh shape exists and resisted narrowing.** The row above was met while writing this very section: a compound command assigning a shell variable and then running `printf` with a long multi-line argument was refused, with the guard naming `printf` and quoting its whole argument. Four paired probes in the same session failed to reproduce it — `printf 'gitignore'` alone passed, an argument carrying backticks passed, the two combined passed, and the same `printf` redirected into a runtime-computed `"$VAR/path"` passed. So neither the `git` substring, nor backticks, nor a computed redirect target is the trigger on its own.

This is worth more than the row itself: it is the clearest available demonstration that **the guard's analyzer refuses on a property none of the observations here has isolated**, and that a refusal can name a command that has nothing to do with git. Treat an unfamiliar refusal clause as a seventh, eighth, or ninth shape rather than as a misfire, and record its verbatim wording — the catalogue above grew twice while this section was being written.

**The classifier does not key on any of these clauses.** `internal/hook/post_tool_failure.go` matches the anchor `isolated in the worktree` alone, which is why the three unpinned shapes still classify as `WorktreeGuardRefusal` rather than falling into the catch-all. The anchor is an observed dependency on upstream wording, not a contract — if the runtime rewrites that opening clause, detection goes silently to zero.

### The axis is git — with one measured exception

Card t741 established that a cross-tree **non-git** argument is not refused: `ls -la <another worktree>/go.mod` passed from a session anchored elsewhere, while `git -C <another worktree> log` in the same session was refused and `git -C <own worktree> log` passed. The discriminating axis there is git, not the crossing of trees.

Card t880 measured the same axis from the other side at 2.1.275 and reported the refusal target as **a git call that cannot be statically bound to the worktree** — not nested shell expansion as such. Backticks, `$( )`, `for` loops, and heredocs were all reported to pass when no git call was involved.

**That generalization does not hold universally, and the counter-example was measured here.** At 2.1.275, in a worktree-isolated session with the guard demonstrably live (a `git -C` control refused in the same session), a quoted-delimiter heredoc whose body was the ten characters of a JSON object — no git anywhere in the command — was refused with the `too complex to verify` clause. So "git-free commands pass" is false as stated; what survives is the narrower claim that a statically unbindable git call is *sufficient* for refusal, not that it is *necessary*.

**Gap**: no measurement here narrows what else is sufficient. The brace observation in the trigger table above and this one are the same shape, and its boundary is still unmeasured.

### Two observations disagree on the heredoc-carrying-git-text shape

The trigger table above records, first-hand at **2.1.251**, that a quoted-delimiter heredoc whose body names a git subcommand was refused. Two later observations disagree with it and with each other:

| Session kind | Version | Result |
|---|---|---|
| Worktree-isolated | 2.1.251 | **Refused** — `names git in a form too complex to verify` (trigger table above) |
| Worktree-isolated | 2.1.275 | **Passed** — measured first-hand for card t852, with a `git -C` control refused in the same session, so the pass is a pass and not a dead guard |
| Primary checkout | 2.1.275 | **Refused** — reported by another lane (card t852 dispatch) |

Two explanations fit, and **neither is established**:

- **A version change.** The behaviour flipped between 2.1.251 and 2.1.275 for this shape. The same session that saw the flip also re-measured the JSON-brace shape as still refused at 2.1.275, so any such change was selective rather than a general relaxation.
- **A layer difference.** The worktree-isolated session meets the Claude Code runtime guard, while the primary checkout meets this repository's own branch guard (`internal/hook/branch_guard.go`) — two different refusers with overlapping-sounding wording, per the discriminator table at the top of this section.

**To settle it**, run the identical heredoc in both session kinds at one pinned version and compare the refusal wording as well as the outcome; the wording is what separates the two guards. Until that is done, do not cite either observation as the behaviour.

### A background subagent carries no anchor of its own

A background subagent is **not** pinned to the tree it was spawned in. Its working directory is re-resolved against the session's current anchor at each call, including while the session is moving between trees.

Card t741 measured this directly: twelve path-free calls from one background subagent, no `cd` anywhere in them, and the working directory changed twice across the run as the parent session moved — twelve passes, zero refusals.

Two consequences, and the second is the dangerous one:

- **A refused subagent command is not silent.** The `--git-dir` fixture above was captured from a background subagent, so a refusal reaches `PostToolUseFailure` from a subagent exactly as it does from a main session.
- **A re-anchored subagent is entirely silent.** A path-free command keeps succeeding while the tree underneath it changes. Nothing refuses, nothing is logged, and the subagent cannot tell that its later work landed in a different tree from its earlier work. This is the audit-degradation path that refusal records do not catch.

Operationally: do not move a session's worktree while a background auditor is live. Where that is unavoidable, the resulting verdict's Gaps section must carry the refusal record — `verification-claim-integrity.md` §3.1 (a refusal is a Gap, never a silent substitution) governs, and a lead reading the verdict is the only thing enforcing it.

**Not recommended**: per-agent anchor registration. It would require changing the Claude Code binary rather than this repository, and it inverts the guard's purpose — a subagent pinned to its spawn tree keeps writing to a tree the operator believes was disposed of.

### Workarounds — two situations, not two competing options

Which one applies is decided by what you were trying to do, so identify the situation before reaching for a form.

**Writing file content → use the `Write` tool.** This is the repository's convention. It reaches the filesystem without issuing a shell command, so the guard never parses the content and the body's shape stops mattering. Reach for it first rather than reshaping the heredoc.

Ad-hoc detours happen to work — routing the content through an interpreter's own file-write call, or splitting the file into brace-free pieces — but they are **not** the convention and are not equal options. Two paths going their own way is not a convention: a reader who meets one detour in one commit and a different one elsewhere learns nothing reusable.

**Running a compound command → split it into separate plain commands.** Issue the steps one at a time rather than chaining them. Note that this trades away one property worth keeping in mind: an environment scrub written as `unset … && <command>` is load-bearing as a single invocation, because each command runs in a fresh process, so that particular pairing is not one to split apart.

### Acceptance-criteria commands — the measured boundary and the authoring rule

A corpus-wide, block-level census of acceptance-criteria verification commands (111 files
carrying complex git forms across `**/acceptance.md`) probed each file with a read-only replica
of its block composition in a worktree-isolated session at Claude Code **2.1.278** (every
disposition traces to a recorded probe). Like every table in this section it is a record of
observations, not a specification of the parser.

Refused — the `names git in a form too complex to verify` refusal fires and nothing executes:

| Refused form | Notes |
|---|---|
| git inside `$()` whose result a later statement expands — `BASE=$(git merge-base A B)` then `git diff "$BASE"..HEAD` | refuses for `;`, newline, and `&&` separation alike; refuses with `head`, `tail`, and `sort` pipeline terminators too |
| git nested inside another git command's arguments — `git diff "$(git merge-base A B)"..HEAD` | quoted or unquoted |
| a tree-write bundle: variable assignment (`mktemp`, `mkdir`) + git + an `rm -rf` tail in one invocation | refuses even when every git verb in it is read-only — the bundle is what refuses, not the mutation |
| git piped into `while`/`for`, or `for f in $(git …)` | refuses with or without git inside the loop body |
| a `( … )` subshell compound containing git | |
| `test "$(<git $()>)"` | refuses even with a `wc -l` pipeline terminator |
| a non-git command (`printf`, `rg`) whose argument carries a git `$()`; an env-prefixed `VAR="$(git …)" go test …` | `echo` is the measured exception, below |

Measured executable: plain separately-invocable git verbs; `;` / `&&` / `||` compounds of them;
`$?`-capture lines; a `$()` assignment with no later expansion; `$()` embedded in the same
statement's `echo "…$(…)…"` (bare, `|wc`, `|head`, `|awk` terminators all pass); a `$()`
assignment expanded later when the `$()` pipeline terminates in a counter stage (`wc -l`,
`grep -c`, `awk '{print $1}'`); redirect-to-file capture (`git … > /tmp/f`); git text inside a
quoted grep pattern or inside a comment.

**Gap**: the counter-terminator exception and the echo-embedding exception are observed
boundaries, not a mechanism — the territory between the rows is unmeasured, and per the section
gap note above, unknown is not permitted.

**Authoring rule for AC verification commands** (normative — this is the repo-side fix; the
guard itself is the binary's):

1. One plain git verb per line, run from the worktree root.
2. Capture exit codes in separate lines — `git diff --quiet …; echo "exit=$?"` — so the exit
   code is the verification's own field (`verification-completeness.md` §2.1), never a
   `$()`-captured variable.
3. Never git inside `$()`. Derive ranges with the three-dot form (`git diff --name-only
   develop...HEAD`) and record the merge-base on its own line (`git merge-base develop HEAD`)
   when the base value itself is evidence.
4. No write/cleanup composition tail in a verification command. Scratch writes live under
   `/tmp`; cleanup is a separate plain step, never an `rm -rf` bundled into the same invocation.
5. Never relocate a refused command into a script file — the guard cannot read inside a script,
   so the relocation hides the risk instead of removing it. Reduce the verification instead.
6. Pin the tree SHA the measurement was taken on (`verification-completeness.md` §4).

Working example — a refused form and its executable restatement:

```bash
# REFUSED (assignment + later expansion of a git-bearing substitution):
B=$(git merge-base develop HEAD)
git diff --name-only "$B"..HEAD -- internal/pkg/ | wc -l

# EXECUTABLE (plain verbs; base recorded on its own line; three-dot range):
git merge-base develop HEAD                          # record the base value as evidence
git diff --name-only develop...HEAD -- internal/pkg/ | wc -l
git diff --quiet develop...HEAD -- internal/pkg/; echo "diff_exit=$?"
```

**Versions measured**: the trigger table and the delimiter asymmetry were measured at Claude Code **2.1.251**. The message-shape catalogue, the git-axis counter-example, and the heredoc disagreement were measured at **2.1.275** (card t852; `claude --version` read in the measuring session). The subagent-anchor observations were measured at the version current when card t741 was measured, which was not recorded there.

Guard behaviour is version-dependent — one shape has already been observed to flip between these two versions — so **state the version whenever you add a row here, and read the version before citing one.** No behaviour above is known to hold at any version other than the one its row names.

---

Version: 1.0.0 (split from worktree-integration.md; section content moved verbatim — see the parent file for the transfer note)
