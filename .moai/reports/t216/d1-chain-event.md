# t216 / D-1 — `chain-event.sh` wiring drift

> Investigation record, card t216, axis D-1. Read-only; HEAD `a9eb896ce`,
> worktree `.claude/worktrees/t216`, branch `WT-hook-wiring-drift`.
> Persisted by the orchestrator — the investigating agent is read-only and
> holds no Write tool.

## Claim

> `chain-event.sh` is wired in the distributed template's settings.json but NOT
> in this project's `.claude/settings.json` … the local settings.json is
> unmodified, so this is a stale rendered output. Every project initialized
> before the template gained the new hook entry has the same gap.

**Verdict: the wiring fact is true. Three of the surrounding premises are
wrong**, and the stated consequence ("the chain ledger's completion edge has
never been recorded") is true but materially understates the situation.

| Premise | Status |
|---|---|
| `chain-event.sh` wired in template, absent locally | **TRUE** |
| It is the only drift | **FALSE** — there are two, and the second landed today |
| `.claude/settings.json` is a stale *rendered output* | **FALSE** — it is a tracked, hand-maintained, committed file |
| Fixing the wiring records the completion edge | **FALSE** — a no-op; nothing creates the node the edge attaches to |
| Every project initialized before the change has the gap | **TRUE, and worse** — no project can ever receive it via `moai update` |

## Evidence

### E1 — the hook-entry diff, both directions

Rendered the template (all `{{if}}` branches taken, `{{else}}` dropped) and
set-diffed every hook entry keyed on `(event, matcher, script, if, timeout,
async)`:

```
### TEMPLATE ONLY
   ('PostToolUse', 'Write|Edit|MultiEdit', 'status-transition-ownership.sh', 'Edit(**/.moai/specs/**)', 5, True)
   ('PostToolUse', 'Write|Edit|MultiEdit', 'status-transition-ownership.sh', 'MultiEdit(**/.moai/specs/**)', 5, True)
   ('PostToolUse', 'Write|Edit|MultiEdit', 'status-transition-ownership.sh', 'Write(**/.moai/specs/**)', 5, True)
   ('SubagentStop', '<none>', 'chain-event.sh', '<none>', 5, None)
### LOCAL ONLY
   ('PostToolUse', 'Write|Edit|MultiEdit', 'status-transition-ownership.sh', '<none>', 5, None)
total template entries 36 local 33
```

Both files carry the identical 20 hook events; the drift is entirely inside
`SubagentStop` and `PostToolUse`.

**The card names one drift. There are two.**

1. **`chain-event.sh`** — third entry under `SubagentStop` in the template
   (`settings.json.tmpl:198`), absent locally. Local `SubagentStop` carries only
   `handle-subagent-stop.sh` and `handle-harness-observe-subagent-stop.sh`.
2. **`status-transition-ownership.sh`** — the template has **three** `if`-scoped
   async entries (`Write(**/.moai/specs/**)`, `Edit(…)`, `MultiEdit(…)`); the
   local file has **one** unscoped, **synchronous** entry. Two independent
   divergences are folded together here: the `if`-scoping split (2026-08-01,
   `796181198`) and the `async: true` addition (2026-08-24, `ea4c6736f` — card
   t214, merged today). Operationally the local hook fires on *every*
   Write/Edit/MultiEdit anywhere in the tree, synchronously, inside the
   PostToolUse budget; the template's fires only on spec-file edits, async.

No local-only hook entry exists that the template lacks, other than that one
degenerate form.

### E2 — the file is tracked in git, not a rendered artifact

```
$ git ls-files --error-unmatch .claude/settings.json
.claude/settings.json

$ diff <worktree>/.claude/settings.json /Users/goos/MoAI/moai-adk-go/.claude/settings.json
IDENTICAL
$ git status --porcelain .claude/settings.json
(empty)
```

`.claude/settings.json` is a **committed, version-controlled file**, clean in
both the worktree and the primary checkout. Not gitignored, not a rendered
output. Its history shows it is hand-maintained in parallel with the template:

```
34a94cc80 2026-08-19 feat(goal): disarm an armed goal when a turn dies on an unrecoverable API error
e7acc30a7 2026-08-16 chore(release-update): align with Claude Code 2.1.228..2.1.233
3a2dfbaf0 2026-08-13 fix(template): harden hook commands against missing script (#1505)
1bc4ba998 2026-08-12 feat(hook): widen PreToolUse timeout 5→10 (project+template) (#1473)
e8e46396f 2026-08-10 fix(hooks): make the two review gates reachable (#1421)
```

Some commits touch both files; the three that produced the current drift touched
**only** the template:

```
$ git log --format='%h %ad %s' --date=short -S 'chain-event.sh' -- internal/template/templates/.claude/settings.json.tmpl
435bc2bbd 2026-08-13 feat(SPEC-CHAIN-CORE-001): Origin-Trail Chain — Phase 1 worktree lineage tree (#1485)

$ git show --stat 435bc2bbd | grep settings
 .../template/templates/.claude/settings.json.tmpl  |   8 +-

$ git show --stat ea4c6736f | grep settings
 .../template/templates/.claude/settings.json.tmpl  |   9 +-
```

So the mechanism is not "stale render" — it is **an author-discipline gap in this
repo's own dogfood config**, one that no test or check enforces.

### E3 — what `chain-event.sh` does, and the ledger's actual state

`internal/template/templates/.claude/hooks/moai/chain-event.sh` resolves a `moai`
binary (project-local → `$HOME/go/bin` → PATH), pipes the SubagentStop payload to
`moai hook chain-event`, and exits 0 unconditionally. `internal/hook/chain_event.go:70-105`
opens `.moai/state/chain/events.jsonl`, calls `pop.ResolveCurrentNode(cwd,
sessionID)`, and appends a `completion-edge` via `AppendCompletionEdge`.

Ledger state:

```
$ ls -la /Users/goos/MoAI/moai-adk-go/.moai/state/chain/
-rw-r--r--@ 1 goos staff 0 Aug 19 21:53 .gitkeep
```

**There is no `events.jsonl`, anywhere.** Not in the primary checkout, not in
this worktree. The ledger is not "degraded" or "missing one edge class" — it has
never existed.

The reason is upstream of the hook wiring:

```
$ grep -rn "CreateNodeAtSpawn\|BackfillSessionID\|AppendCompletionEdge" --include="*.go" internal | grep -v _test
internal/chain/populate.go:53:  func (p *Populator) CreateNodeAtSpawn(...)      <- definition only
internal/chain/populate.go:107: func (p *Populator) BackfillSessionID(...)
internal/chain/populate.go:134: func (p *Populator) AppendCompletionEdge(...)
internal/hook/chain_event.go:98:   pop.AppendCompletionEdge(...)
internal/hook/chain_banner.go:86:  pop.BackfillSessionID(...)

$ grep -rn "EnvChainNodeID" --include="*.go" internal | grep -v _test
internal/config/envkeys.go:238:  EnvChainNodeID = "MOAI_CHAIN_NODE_ID"
internal/chain/populate.go:54:   parentID := os.Getenv(config.EnvChainNodeID)
internal/chain/populate.go:155:  if envID := os.Getenv(config.EnvChainNodeID); envID != ""
internal/hook/chain_banner.go:78: envNodeID := os.Getenv(config.EnvChainNodeID)
```

`CreateNodeAtSpawn` — the *only* writer of `node-enter` events, i.e. the only
thing that can create the ledger — **has no production caller**. Nothing ever
*sets* `MOAI_CHAIN_NODE_ID`; every reference only reads it. There is no
`moai chain enter` subcommand (`internal/cli/chain.go` exposes only `status`,
`lineage`, `back`, `list`, `prune` — all read/prune paths).

Consequence: even if `chain-event.sh` were wired today, `ResolveCurrentNode`
(`populate.go:153-166`) would find no node, `chain_event.go:86` would log
`no matching chain node (non-blocking)` and return nil. **Wiring the hook alone
changes nothing observable.** The card's fix, taken literally, is a no-op.

### E4 — how `moai update` treats hook entries: the decisive code path

`.claude/settings.json` *is* a merge target — `internal/cli/update_template_sync.go:381`
lists it in `collectMergeableFiles`. After deploy, `merge.MergeUserFiles`
(`internal/cli/update/merge/merge.go:173`) runs a 3-way merge with `JSONMerge`
(`internal/merge/strategies.go:81-82`, extension-based).

The base side is *derived*, because no real base is stored
(`internal/cli/update/merge/base.go`). `pruneToShared` narrows the deployed
template to the keys the user's file already has:

```go
// internal/cli/update/merge/base.go:118-128
		updatedChild, updatedIsMap := updatedVal.(map[string]any)
		currentChild, currentIsMap := currentVal.(map[string]any)
		if updatedIsMap && currentIsMap {
			pruned[key] = pruneToShared(updatedChild, currentChild)
			continue
		}
		pruned[key] = updatedVal
```

This recurses **only through maps**. It fixes the add-a-new-*key* case — and
`base.go`'s own header comment says so ("a new MCP server entry, a new settings
key … silently failed to arrive on every existing install"). It does **not** fix
arrays, and a hook entry is an array element.

Simulated the real files through this exact algorithm:

```
base['hooks'] == template['hooks'] ?  True
base['hooks'] == local['hooks']    ?  False
baseChanged= True  updChanged= False  -> result = curVal (user's hooks)
```

Because every one of the 20 event keys is present on both sides, `pruneToShared`
copies the template's `hooks` map through **verbatim**, so `base == updated` at
the `hooks` key. Then in `deepMergeMap`:

```go
// internal/merge/strategies.go:427-429
			case baseChanged && !updChanged:
				// Only user changed.
				result[key] = curVal
```

**`internal/merge/strategies.go:427-429` is the line that decides.** The merge
concludes the *user* edited `hooks` and the template did not, and preserves the
user's entire `hooks` block wholesale. Recursion into `SubagentStop` never
happens.

**Answer to the card's central question: NO.** A hook entry added to the template
after a project was initialized can never reach that project via `moai update`.
The only exception is a project whose `settings.json` lacks an entire top-level
key or an entire event key that the template has — then `!inBase && !inCurrent &&
inUpdated` (`strategies.go:389-391`) adds it. Adding an *element* to an existing
event's array is structurally unreachable.

### E5 — blast radius and detection surfaces

- **Affected population:** every project whose `.claude/settings.json` predates
  any template hook-array change — every project initialized before 2026-08-13
  for `chain-event.sh`, before 2026-08-01 for the `status-transition`
  `if`-scoping, and *every existing project* for the 2026-08-24 `async: true`.
  This repo's own dogfood config is in all three cohorts.
- **`moai doctor`:** `checkHooksConfig` (`internal/cli/doctor.go:677-692`) does
  exactly one thing — `os.Stat(".claude/hooks")` and reports "hook handlers
  directory found". It never opens `settings.json` and never compares against the
  template. No other doctor check does either.
- **`moai update --preview`:** `AnalyzeMergeChanges` (`merge.go:306`) classifies
  `settings.json` as "High risk" *by filename*, not by content delta. It says the
  file is risky, not that an entry is missing.
- **Tests:** `internal/template/settings_test.go` has 20 tests including
  `TestSettingsTemplateRequiredHooks`, `TestSettingsTemplateAllHookEvents`,
  `TestSettingsTemplateHookIfConditions`, `TestSettingsTemplateHookEventCount` —
  **all assert against the template only**. No test compares a project's
  `.claude/settings.json` to the template.
- **The one runtime signal that exists** is the wrapper's fail-open logger: each
  hook command is `bash -c '[ -f "$0" ] && exec bash "$0"; … >> "$d/.moai/logs/hook-missing.log"'`.
  That catches a *missing script for a registered entry*. It cannot catch an
  *unregistered entry* — the inverse case, which is this bug.
  `/Users/goos/MoAI/moai-adk-go/.moai/logs/hook-missing.log` does not exist.

**How would anyone notice? They wouldn't.** The drift is silent on every surface.
It was found here only by hand-diffing two files.

### E6 — fix options, and the hard case

The hard case is uniform across all of these: **there is no way to distinguish
"the user deliberately deleted this hook entry" from "this entry was added to the
template after the project was created."** Both look identical — an array element
the template has and the project does not. No deploy-time template snapshot exists
(`base.go` says so: closing this "needs the deployed template content to be
snapshotted at deploy time so a genuine base exists on the next update"). Every
option below is a bet on which reading is more often correct.

| Option | Cost | Behaviour on a deliberately-removed entry |
|---|---|---|
| **A. Extend `pruneToShared` to arrays** — key hook entries by `(event, matcher, if, script)` and omit template-only entries from the base, so `deepMergeMap`'s `!inBase && !inCurrent && inUpdated` branch adds them | Smallest code change, but `deepMergeMap` compares array values whole (`valuesEqual` marshals to JSON) — no element-level merge, so this needs a new hook-aware strategy alongside the base change. Two packages touched | **Silently re-adds it on every update.** The removal is permanently un-honourable with no opt-out short of editing after each update. Worst behaviour of the five |
| **B. Snapshot the deployed template at deploy time** (write the rendered `settings.json` to `.moai/state/` on deploy), giving a genuine base next update | The correct fix; also closes the "template changed a value the user never touched" limitation `base.go` documents. Costs a new on-disk artifact, migration for projects with no snapshot (they fall back to today's derived base), and array-aware merge is *still* needed on top | **Correctly preserved** — base has the entry, current doesn't, updated does → `inBase && !inCurrent && inUpdated` → "respect user deletion" (`strategies.go:409-411`). The only option that gets the hard case right by construction |
| **C. Doctor diagnostic that reports drift, changes nothing** — render the template in-memory, diff hook entries against `.claude/settings.json`, print both directions | Cheapest, lowest-risk. Reuses the existing `renderSettings` path (`update.go:1198`). Purely additive; no merge semantics touched | **Reports it as drift forever** — noisy for a deliberate removal unless paired with an ignore list. But it never destroys intent, and it turns a silent gap into a visible one |
| **D. `moai update --hooks`** — explicit opt-in that re-renders only the `hooks` block | Needs a partial-render path that leaves `permissions`/`env`/`statusLine` alone. User must know to run it, so discovery is unsolved unless C ships too | **Clobbers it, but only when explicitly asked.** Consent is explicit, which makes this defensible where A is not |
| **E. Full re-render of `settings.json`** | Simplest to implement | **Destroys every customization in the file** — permissions, env, model, statusLine. Not viable |

**The `CleanMoaiManagedPaths` interaction** (`internal/cli/update/deploy/deploy.go`):
`moai update` wipes `.claude/hooks/moai` wholesale before redeploying, so the
*scripts* always land fresh — `chain-event.sh` is already present at
`/Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/chain-event.sh` despite never
being registered. The asymmetry is: **scripts are force-synced, registrations are
never synced.** Any recommendation that says "just run `moai update`" is wrong for
this class of bug — it would refresh a script that already exists, leave the
missing registration untouched, and destroy any hand-added script under
`.claude/hooks/moai`.

**Recommended sequencing:** C first (cheap, safe, makes the class of bug visible
and lets you measure the affected population), then B (the only structurally
correct fix), with D as the interim escape hatch. A is the tempting one-liner and
it is the one that silently overrides user intent.

## Baseline-attribution

- `chain-event.sh` in template, not in project — `435bc2bbd` (2026-08-13,
  SPEC-CHAIN-CORE-001 #1485). Template-only commit. **11 days old.**
- `status-transition-ownership.sh` `if`-scoping split (1 entry → 3) —
  `796181198` (2026-08-01, #1271). Template-only commit. **23 days old.**
- `status-transition-ownership.sh` `async: true` — `ea4c6736f` (2026-08-24, card
  t214, #1625). Template-only commit. **Landed today**; this drift is a direct
  sibling of t214, and t214 is its baseline.
- The merge-base derivation (`internal/cli/update/merge/base.go`) is itself a
  *fix* for the map-key version of this exact bug — the array-element version is
  a known-shape gap closed for keys and left open for arrays. Not a regression, a
  partial fix.
- The chain population gap (`CreateNodeAtSpawn` uncalled) is baseline to
  SPEC-CHAIN-CORE-001 Phase 1 as merged; nothing regressed it.

## Gaps

1. **`moai update` was not executed** (destructive, prohibited), so the merge
   outcome is traced in code and reproduced by simulating `pruneToShared` +
   `deepMergeMap` on the real file pair — not observed end-to-end. A
   characterization test in `internal/cli/update/merge/` would settle it; none
   exists for the array case.
2. **`.claude/settings.local.json` was not checked** for a `chain-event`
   registration in the primary checkout. It is Claude Code's higher-precedence
   overlay; if it carries one, E3's conclusion changes. `rest-and-wiring.md:158`
   discusses that file, so it exists.
3. **Whether SPEC-CHAIN-CORE-001 *intended* Phase 1 to leave the spawn-side
   population unimplemented** (deferred to a later phase) is not established.
   `rest-and-wiring.md:546` raises the same open question. It determines whether
   E3 is a bug or a documented partial delivery.
4. **The non-hook portions of the two settings files were not diffed**
   (permissions, env, statusLine). There may be further committed drift outside
   the hooks block.

## Residual-risk

- **Acting on the card as written produces a no-op.** Adding `chain-event.sh` to
  `.claude/settings.json` will not cause a single ledger event to be written,
  because no `node-enter` event ever gets created. Anyone who "fixes" t216 by
  editing settings.json and stopping there closes the card without changing
  observable behaviour — and with no ledger to inspect, that will not be caught.
- **The card's "stale rendered output" framing misdirects the fix.** Because
  `.claude/settings.json` is tracked and hand-maintained, the *local* symptom is
  fixed by a commit, not by any update-path change. The update-path bug (E4) is
  real and independent and would survive fixing the local file. Conflating them
  risks fixing the visible one and shipping the invisible one.
- **If option A is chosen without a snapshot base**, every user who ever removed a
  hook entry on purpose gets it silently reinstated on their next update, with no
  diagnostic and no way to make the removal stick. That failure is worse than the
  current silence because it is *active*.
- **The `async: true` drift means this repo's own PostToolUse budget differs from
  what every user gets** — so hook-timing work done in this checkout is being
  measured against a configuration no user runs.
