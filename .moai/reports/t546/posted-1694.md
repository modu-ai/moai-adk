
Thanks for the exceptionally detailed report — the file-count table made the diagnosis
direct. This is confirmed as a defect, not intentional behavior, and the root cause is
now fixed in development.

**Root cause.** The statusline (and a small family of state writers) resolved its state
anchor from the session's current working directory — the statusline receives the
session's `current_dir` — instead of resolving up to the project root. The resolver
never walked upward, so every directory a session touched gained its own `.moai/state/`
on the next render. Your counts map to the defect exactly:

| Your count | Component |
|---|---|
| `context-usage` 213 | statusline session-telemetry write (fires on every status render — one record per visited directory) |
| `config-cache` 150 | config loader cache write (it only lands where a `.moai` already exists, so these followed the telemetry strays) |
| `github/counts` 133 | statusline GitHub counts refresh (shares one board root with the landed-counts cache) |

Your observation that only `state/` artifacts appear — never `config/` — was the key
confirmation: cache writes were anchored to the working directory, while genuine project
initialization always lands at the project root.

One entry in your table is *not* part of this defect: `session-memo` (7). Those are
per-project records written once per project at compaction time through the hook chain,
which resolves its own project directory correctly — they don't grow per visited
directory, and we left that writer untouched.

**The fix.** All of the affected writers now resolve their state anchor through one
shared resolver (`internal/stateanchor`) with a fixed precedence: the project directory
from the runtime payload → the worktree's original directory → a git root resolution
(one root shared by every checkout and worktree of the repository). The families routed
through it: the session-telemetry write, the board root feeding both the landed counts
and the GitHub counts cache, the armed-goal state read, and the config cache. When no
project can be resolved at all, the state write is skipped instead of landing in the
current directory. Two visible improvements ride along: a session that has cd'd
elsewhere now still sees the project's armed goal, and no new stray `.moai` directories
appear. The statusline display itself is unchanged. New strays stop appearing with the
release that carries this fix.

**Your questions:**

1. CWD-relative state was unintentional — a resolution bug, not per-directory caching.
   Answered by the fix above.
2. Cleanup: your instinct is right that wholesale deletion needs care. A safe recipe:
   delete any `.moai` directory whose contents are ONLY `state/` (no `config/` inside);
   keep any `.moai` that contains `config/` — that is a real initialized root. A
   dedicated cleanup command is being considered separately.
3. Template gitignore hardening: your observation is correct — the shipped pattern is
   `.moai/state/`, which (containing an inner slash) is anchored to the repository root
   and does not cover nested strays. With the fix, new strays stop appearing at the
   source, so we would rather not ship a broad `**/.moai/state/` pattern that could mask
   real state in legitimate nested roots. If strays from pre-fix versions are a concern,
   the cleanup recipe in (2) handles them explicitly.

We'll follow up here once the fix ships in a release.
