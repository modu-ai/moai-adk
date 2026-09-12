# t576 (E3) — the 2026-09-12 `permissions.ask` deletion: writer identification

Card t576, second occurrence. The 2026-09-08 event (E1) and the 2026-09-10
restoration (E2) were established in `.moai/reports/t576/verdict.md`, which
lives on the branch `WT-permissions-ask-empty` in the worktree
`.claude/worktrees/t576`. That worktree is locked by a live session
(`lane-3`, pid 13354), so this continuation runs in a second worktree.

Worktree: `.claude/worktrees/t576-e3` · branch `WT-settings-ask-deletion` ·
base `1d150a27d` (= `origin/develop` at dispatch, `0 0` divergence).

Nothing was restored, reverted, or deleted by this investigation. The lead
had already restored the file at 2026-09-12T09:28:25Z before dispatch.

## Claim

1. **E3 is not the same kind of event as E1.** E1 was a whole-file JSON
   re-serialization: every top-level key reordered into the writer's
   insertion order, `ask` retained as a literal `[]`. E3 reorders nothing,
   re-serializes nothing, and removes the `ask` **key itself**.
2. **E3 is a contiguous line-range deletion.** The drifted file equals the
   committed `main` blob with lines 486–493 removed and **every other byte
   identical**. 21846 − 21693 = 153 bytes, one hunk, zero other changes.
3. **Therefore every program in this repository is excluded**, not by
   enumeration but by shape: each Go writer of this file re-serializes the
   `permissions` object through `renderPermissionsObject`, and a
   re-serializer cannot leave the remaining 21693 bytes byte-identical to a
   file it did not write. The launcher family the dispatch named as the
   leading suspect (#1676) is excluded on this ground **and**
   independently — it writes only `settings.local.json`.
4. **No Claude session performed it.** All 352 session transcripts across
   every profile, modified since 2026-09-10, were scanned for `Edit` /
   `Write` on this path and for `Bash` commands capable of mutating it. The
   only mutation found is the lead's restoration.
5. **The drifted content is not any committed state.** No blob in the
   history of `.claude/settings.json` matches its sha256, so a checkout,
   `git restore`, or merge artifact is excluded.
6. **The actor is not identified.** What remains consistent with the
   evidence is an edit made outside every surface that leaves a record here:
   an external editor, or a Claude Code UI surface that writes without
   emitting a transcript tool call. This is stated as a Gap, not narrowed
   further.

## Evidence

All commands run in this worktree, in this run, against the preserved
snapshot and the committed blob.

### The shape of the change

```
$ git show main:.claude/settings.json > <scratch>/main.json
$ shasum -a 256 <scratch>/main.json <preserved E3 snapshot>
9f4818a617c9ab2e3b4102efcd1aacc62d9ebcf845b2bf42cb05a1b5b3c62a0e  …/main.json
bddb867650c7b96adab29e0d53d6e3e6d02e66137c402fc1e96df4707bfaf420  …/settings.json.main.20260912T091322.926Z.bddb8676

$ diff <scratch>/main.json <preserved E3 snapshot>
486,493d485
<     ],
<     "ask": [
<       "Bash(rm:*)",
<       "Bash(sudo:*)",
<       "Bash(chmod:*)",
<       "Bash(chown:*)",
<       "Read(./.env)",
<       "Read(./.env.*)"
(exit 1 — one hunk, no others)
```

The deleted span in context (`main.json`, lines 482–497):

```
485	      "Bash(moai:*)"
486	    ],              ← deleted: the bracket that closed "allow"
487	    "ask": [        ← deleted
488-493	  the six entries ← deleted
494	    ],              ← kept: now closes "allow"
495	    "deny": [
```

This offset — take the *preceding* `],` and leave the *trailing* one — is
what keeps the result valid JSON after a line-range removal. It is the
signature of a line-based text deletion, and it is not a shape a serializer
produces: a serializer emits the whole object and would rewrite the
surrounding lines too.

### Structural comparison, E1 versus E3

```
$ jq -c '.permissions|keys_unsorted' <E3 snapshot>
["defaultMode","allow","deny"]
$ jq -c '.permissions|keys_unsorted' <current file>
["defaultMode","allow","ask","deny"]

$ jq -c 'keys_unsorted' <E3 snapshot>
["$schema","hooks","statusLine","skillListingBudgetFraction","showThinkingSummaries","cleanupPeriodDays","outputStyle","env","permissions","attribution","respectGitignore"]
$ jq -c 'keys_unsorted' <current file>
["$schema","hooks","statusLine","skillListingBudgetFraction","showThinkingSummaries","cleanupPeriodDays","outputStyle","env","permissions","attribution","respectGitignore"]
```

| | E1 (2026-09-08) | E3 (2026-09-12) |
|---|---|---|
| top-level key order | reordered to writer insertion order | committed order, unchanged |
| `ask` | present, `[]` | key absent |
| bytes outside the change | re-serialized | byte-identical |
| `allow` / `deny` | 114 / 48 | 114 / 48 |

Same six entries lost, two different mechanisms. Reading E3 as a recurrence
of E1 would attribute it to a writer that provably did not act.

### Why no Go writer can produce this

`renderPermissionsObject` (`internal/config/toolpolicy/settings_region.go:177`)
is the single serializer behind every toolpolicy entry point. It emits
`defaultMode, allow, ask, deny` in fixed order and **omits `ask` when the
list is empty** — which is why the E3 *key set* superficially resembles its
output. But `RenderSettingsJSON` (`codegen.go:111`) splices a freshly
serialized object over the whole `permissions` region, so every line of that
region is rewritten. E3's `permissions` region is byte-identical to `main`
except for the removed span, including indentation and entry order across
all 162 surviving specifiers. The serializer is excluded by the bytes it
would have had to rewrite and did not.

Independently, the doc it would have read is not empty:

```
$ grep -o "decision: [a-z]*" .moai/config/sections/tool-policy.yaml | sort | uniq -c
 109 decision: allow
   6 decision: ask
  64 decision: deny
```

Six `ask` rules — so even if a toolpolicy path had run, it would have
emitted the six entries, not dropped them.

The launcher family is excluded twice over:

```
$ grep -n 'settings\.local\.json\|settings\.json' internal/cli/launcher.go
… 1036: filepath.Join(".claude", "settings.local.json")
… 1039: filepath.Join(root, ".claude", "settings.local.json")
… 1059-1084: permissions.defaultMode → settings.local.json
```

Every permission mutation in `launcher.go` targets `settings.local.json`.
The project file is never opened for writing.

### Transcript enumeration

```
$ find ~/.moai/claude-profiles/moai-adk/projects -name '*.jsonl' -newermt '2026-09-11 12:00:00' | wc -l
191
$ find ~/.claude/projects ~/.moai/claude-profiles -name '*.jsonl' -newermt '2026-09-10 00:00:00' \
    | grep -v '/claude-profiles/moai-adk/projects/' | wc -l
161
```

352 transcripts, every profile directory on this machine, filtered to
`tool_use` records whose `file_path` or `command` names this path:

```
Edit/Write on */.claude/settings.json since 2026-09-10:
2026-09-10T06:12:58Z Edit  …/.claude/worktrees/t609/.claude/settings.json   (a different file)
2026-09-11T02:08:00Z Write …/scratchpad/proj-new/.claude/settings.json      (a test fixture)
2026-09-11T02:08:03Z Write …/scratchpad/proj-old/.claude/settings.json      (a test fixture)
2026-09-12T09:28:25Z Edit  /Users/goos/MoAI/moai-adk-go/.claude/settings.json ← the lead's restoration
```

The restoration timestamp matches the file's mtime (`2026-09-12 18:28:26`
local = `09:28:26Z`) to the second. Every `Bash` record naming the path in
the window is a read (`git diff`, `grep`, `jq -r`, `shasum`); the four hits
outside the `moai-adk` profile belong to other repositories.

Shell history holds four `vi …settings.json` entries, all stamped
`1782909052` = `2026-07-01T12:30:52Z` — ten weeks before the window.

### The content is not a committed state

```
$ for each commit touching .claude/settings.json: sha256 of its blob vs bddb8676…
scan done   (no MATCH line emitted)
```

## Baseline-attribution

- Preserved snapshot: `.moai/state/settings-drift/settings.json.main.20260912T091322.926Z.bddb8676`,
  sha256 `bddb8676…f420`, 21693 bytes, ledger row `measured_at
  2026-09-12T09:13:22.926974Z`, branch `main`, `bypassed: false`.
- Committed baseline: `main:.claude/settings.json`, sha256 `9f4818a6…2a0e`,
  21846 bytes — re-derived in this run, not carried from the E1 verdict.
- Code read at `1d150a27d` (this worktree's HEAD = `origin/develop`).
- Transcript corpus: every `*.jsonl` under `~/.claude/projects` and
  `~/.moai/claude-profiles/*/projects` with mtime ≥ 2026-09-10T00:00:00.

## Gaps

- **The actor is unidentified.** Nothing observed names who deleted the
  lines. The claim is about the *mechanism* (a line-range deletion) and the
  *exclusions*, not about a person or a process.
- **The window is bounded only loosely** — after 2026-09-10T02:37Z (the E2
  restoration, per the E1 verdict) and before 2026-09-12T09:13:22Z (the
  detection). `moai integration acquire` records only hits, so the clean
  acquires between those points left no row to narrow it with.
- **Intent is unknown.** The 2026-09-08 event was later stated by the
  operator to be deliberate; nobody has been asked about this one.
- **No repair was attempted or assessed.** Whether the `ask` list should be
  restored differently, or defended mechanically, is not addressed here.
- **The E1 verdict was read, not re-verified.** Its measurements are cited
  as prior art; this run re-derived only the `main` baseline it shares.

## Residual-risk

- **A recurrence would not be caught any sooner.** Detection still depends
  on someone running `moai integration acquire`; the gate records and
  preserves, but nothing watches. The nine-day blind spot that
  `CLAUDE.local.md` §4.1 records is unchanged by this investigation.
- **The exclusion of Go writers rests on the current tree.** A future writer
  that performs surgical line edits rather than re-serialization would
  defeat the shape argument used here.
- **The transcript scan proves absence only where transcripts exist.** A
  session whose profile directory is outside the two roots scanned, or whose
  transcript was pruned, would not appear. The scan covers what is on this
  machine now.
- **The `ask` list is not the load-bearing control.** As the E1 verdict
  established, `settings.local.json` sets `defaultMode:
  "bypassPermissions"`, which disables the prompt layer `ask` feeds. A
  restored `ask` list does not by itself restore the protection.
- **`moai todo pr` reads this card as `landed` while E3 is not.** Measured
  in this run: `moai todo pr t576` prints the `landed` column for t576,
  yet `git merge-base --is-ancestor 1b0cd7097 develop` exits non-zero and
  `git rev-list --count --left-right develop...HEAD` reports `30 1` — the
  E3 commit is outside `develop`. The `landed` predicate resolves on the
  card's earliest delivering branch (E1, which develop does contain) and
  carries no notion of a card whose work landed in parts, so a card with
  an outstanding branch reads clean. Anyone gating a `done` transition on
  that column alone would close this card with E3 unmerged; the
  ancestor check per branch is the predicate that separates the two.
