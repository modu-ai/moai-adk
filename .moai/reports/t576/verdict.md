# t576 — `.claude/settings.json` `permissions.ask` emptied: cause investigation

Card t576 (GH-less, lane-4/t518 window detection, security axis). The card's
first deliverable is **establishing what happened**, not repairing it. No
value was restored, reverted, or deleted by this investigation.

Worktree: `.claude/worktrees/t576` · branch `WT-permissions-ask-empty` · base
`d60b2956b`.

## Claim

1. **The card's premise is stale.** `permissions.ask` in the primary
   checkout's working copy is **no longer empty** — it carries the 6 entries
   and the file is byte-identical to `main`'s committed version. There are
   therefore **two** events to explain, not one.
2. **E1 (the emptying)** was a whole-file JSON rewrite by a writer that
   preserves its own key insertion order. It is **not** any code path in this
   Go repository, **not** a hand edit of the file, and **not** a git
   operation. The writer is Claude Code's own permission-rule editing surface
   — first narrowed to it by a key-order signature, then **witnessed
   reproducing that exact signature live** during this card (see § E1's
   mechanism, witnessed live). A person instructing that surface is the
   actor; for the live event the operator states the deletion was deliberate.
   The 09-08 actor was never observed and stays an inference.
3. **E2 (the restoration)** was a `moai update` redeploy followed by the
   `git restore` step that `CLAUDE.local.md` §2.3 documents as post-update
   procedure. It silently overtook the operator's 2026-09-08 decision to
   leave the drift in place until the batch ended.
4. **The 6 entries were lost, not migrated.** `settings.local.json` carries
   no `ask` key at all.
5. **The empty `ask` was not the load-bearing exposure.**
   `settings.local.json` sets `permissions.defaultMode:
   "bypassPermissions"`, which disables the prompt layer the `ask` list feeds.
   This is worse news than an empty list, not better.
6. **A latent second finding, in two parts**: the `moai update` merge's
   template-versus-user conflict branch is **unreachable for every shared
   key**, and there is no path for the template to re-assert a value — so an
   already-empty `ask` would have survived every future update with no signal
   to the user. The user's value winning is *designed*; the unreachable
   conflict detector and the missing signal are the defects. Static-derived,
   **not executed**.

## Evidence

All commands run in this worktree (or read-only against the primary
checkout's files), in this run.

### The premise is stale

```
$ jq '{allow:(.permissions.allow|length), deny:(.permissions.deny|length),
       ask:(.permissions.ask|length), has_ask:(.permissions|has("ask"))}' \
     /Users/goos/MoAI/moai-adk-go/.claude/settings.json
{ "allow": 114, "deny": 48, "ask": 6, "has_ask": true }

$ jq -c '.permissions.ask' /Users/goos/MoAI/moai-adk-go/.claude/settings.json
["Bash(rm:*)","Bash(sudo:*)","Bash(chmod:*)","Bash(chown:*)","Read(./.env)","Read(./.env.*)"]

$ git show main:.claude/settings.json > <scratch>/settings.main.json
$ shasum -a 256 /Users/goos/MoAI/moai-adk-go/.claude/settings.json \
                <scratch>/settings.main.json
9f4818a617c9ab2e3b4102efcd1aacc62d9ebcf845b2bf42cb05a1b5b3c62a0e  …/.claude/settings.json
9f4818a617c9ab2e3b4102efcd1aacc62d9ebcf845b2bf42cb05a1b5b3c62a0e  …/settings.main.json

$ wc -c < …/.claude/settings.json ; wc -c < …/settings.main.json ; wc -c < …/settings.develop.json
21846                21846                24399
```

The preserved drift snapshot still carries the emptied state, so the evidence
of E1 survives the restoration:

```
$ jq -c '{ask_len:(.permissions.ask|length), allow:(.permissions.allow|length),
          deny:(.permissions.deny|length)}' \
     .moai/state/settings-drift/settings.json.main.20260908T031051.626Z.f509fffa
{"ask_len":0,"allow":114,"deny":48}
```

Three ledger rows, all the same sha256 `f509fffa…`, at
`2026-09-08T03:10:51Z`, `06:21:48Z`, `08:28:05Z` — the drift persisted across
three `moai integration acquire` calls unchanged.

### E1 — the writer's identity, by key order

The discriminator is the **top-level key order**, because each candidate
writer has a different, forced ordering behaviour.

```
$ jq -r 'keys_unsorted|join(",")' <drift snapshot>
$schema,respectGitignore,cleanupPeriodDays,skillListingBudgetFraction,env,attribution,permissions,hooks,statusLine,outputStyle,showThinkingSummaries

$ jq -r 'keys_unsorted|join(",")' <main HEAD>      # develop HEAD is identical
$schema,hooks,statusLine,skillListingBudgetFraction,showThinkingSummaries,cleanupPeriodDays,outputStyle,env,permissions,attribution,respectGitignore

$ /usr/bin/grep -n '^  "' internal/template/templates/.claude/settings.json.tmpl
$schema,hooks,statusLine,skillListingBudgetFraction,showThinkingSummaries,cleanupPeriodDays,model,outputStyle,env,permissions,attribution,respectGitignore,includeGitInstructions,plansDirectory
```

The key **sets** are identical between the drift snapshot and `main` HEAD, and
the drifted order is **not** alphabetical:

```
$ diff <(jq -r 'keys|.[]' <drift>) <(jq -r 'keys|.[]' <main HEAD>)
SETS-IDENTICAL=yes
$ jq -r 'keys|join(",")' <drift>        # alphabetical, for contrast
$schema,attribution,cleanupPeriodDays,env,hooks,outputStyle,permissions,respectGitignore,showThinkingSummaries,skillListingBudgetFraction,statusLine
```

What each observation eliminates:

| Candidate writer | Forced ordering | Verdict |
|---|---|---|
| Go `json.Marshal` of `map[string]any` | **alphabetical** | ELIMINATED — drifted order is not alphabetical |
| Go `json.Marshal` of a struct | struct field declaration order | ELIMINATED — a repo-wide sweep for `type .*Permissions.* struct` found exactly one such type (`toolpolicy.PermissionsBlock`, 4 fields), and no struct modelling the 11 top-level keys |
| Template deploy (`settings.json.tmpl`) | the template's literal order | ELIMINATED — different order, and the template carries three keys (`model`, `includeGitInstructions`, `plansDirectory`) the drifted file lacks |
| Human edit | none | ELIMINATED (practically) — a hand edit does not reorder all 11 top-level keys while leaving the key set identical and emptying exactly one array |
| A writer preserving its own insertion order (JS/TS `JSON.stringify`) | insertion order | **SURVIVES** |

The surviving shape is corroborated by the one file in this tree that Claude
Code demonstrably owns and rewrites — `.claude/settings.local.json`, which
carries the same non-alphabetical insertion-order signature:

```
$ jq -r 'keys_unsorted|join(",")' .claude/settings.local.json
permissions,enabledMcpjsonServers,disabledMcpjsonServers,hooks,statusLine,outputStyle,prefersReducedMotion,teammateMode
```

### E1 — the Go codebase is ruled out by enumeration, not by assumption

A read-only sweep of `internal/`, `pkg/`, `cmd/` enumerated **seven** writers
of the project-level `.claude/settings.json`. None can produce the observed
shape (114 allow / 48 deny intact, `ask` present as a literal `[]`):

| Writer | Location | Can it produce `ask: []`? |
|---|---|---|
| `template.Deployer.Deploy` | `internal/core/project/initializer.go:434` | No — the template ships the 6 entries verbatim |
| `stripRetiredV2DenyEntries` | `internal/cli/update_deny_migration.go:42` | No — only ever assigns `perms["deny"]` |
| `migrations.m002Apply` | `internal/migration/migrations/m002_settings_cleanup.go:46` | No — touches only the top-level `hooks` key |
| `updatemerge.MergeUserFiles` → `deepMergeMap` | `internal/cli/update/merge/merge.go:173`, `internal/merge/strategies.go:300-461` | No new emptying — but **preserves** an already-empty `ask` (see § Latent second defect) |
| `toolpolicy.BuildInto` / `BuildIntoAuto` | `internal/config/toolpolicy/codegen.go:207` | No — **omits the key entirely** when `len(Ask) == 0` (`settings_region.go:204`) |
| `toolpolicy.RenderTierPermissions` | `internal/config/toolpolicy/tier_render.go:55` | No — same key-omission behaviour |
| `toolpolicy.WriteUserDefaultMode` | `internal/config/toolpolicy/tier_render.go:106` | Out of scope — writes USER-scope `~/.claude/settings.json` |

Two paths that would have been the natural suspects are ruled out explicitly:

- **The launchers** (`moai cc` / `glm` / `cg`): every permission mutation
  targets `.claude/settings.local.json` only
  (`internal/cli/launcher.go:1120-1140`, `internal/cli/settings.go:179-207`);
  the project file is never opened.
- **The SessionStart hook / all of `internal/hook`**: of 135 non-test files,
  13 mention the string `settings.json`, all in comments; zero write sites.

`internal/settings` is a misleading name — 0 of its 10 non-test files mention
`settings.json`; it drives `.moai/config/sections/*.yaml`.

Commit `48239c7dc` (`fix(settings): stop GLM settings loss across console
saves, launcher rewrites, and inert effort keys`, #1676) was a natural
candidate and is **ruled out**: grepping its diff for `"ask"`,
`permissions.ask`, `PermissionsBlock`, and `.claude/settings.json` returns 0
matches across all 21 files it touched. It is an ancestor of the installed
binary (`git merge-base --is-ancestor 48239c7dc 84fa4ece4` → yes), so its
presence or absence does not discriminate either event.

### E2 — the restoration, by mtime clustering

```
$ find .claude .moai/config -type f \
       -newermt '2026-09-10 11:30:00' ! -newermt '2026-09-10 11:45:00' | wc -l
263
```

263 files under exactly the roots `CLAUDE.local.md` §2.3 documents as the
`moai update` wipe list moved inside one 15-minute window (local time; the
settings.json mtime in that window is `2026-09-10 11:37:28` local =
`02:37:28Z`). A targeted `git restore` of one file cannot produce that; a
`moai update` redeploy is the shape that does.

But the resulting bytes equal `main` HEAD, which the current template cannot
render (it carries three keys the committed file lacks, above), and after the
event only **two** tracked files remain modified in the primary checkout
(`.moai/config/sections/git-strategy.yaml`, `CLAUDE.local.md`) — both of
which `CLAUDE.local.md` §2.3 names as requiring manual re-application after
an update. The coherent reconstruction: **update redeployed, then the
documented `git restore` swept tracked files back to `main` HEAD.**

Consequences, both unnoticed at the time:

- The operator's 2026-09-08 decision — *"배치 진행 중에는 그대로 둔다 …
  배치 종료 후 복원한다"* — was overtaken by routine maintenance **during**
  the batch.
- That same update rewrote the mtimes of all 263 files, **destroying the
  mtime evidence for E1**. This is why E1 had to be pursued through key
  order rather than through timestamps.

### E1's mechanism, witnessed live — the writer is no longer an inference

**This section supersedes the "NOT established" entry for E1's trigger in
§ Gaps.** While this card was being worked, the operator ran Claude Code's
`/permissions` command in this session, whose working directory is the t576
worktree. Its output:

```
Deleted ask rule Bash(chmod:*)
Deleted ask rule Bash(chown:*)
Deleted ask rule Bash(rm:*)
Deleted ask rule Bash(sudo:*)
Deleted ask rule Read(./.env.*)
Deleted ask rule Read(./.env)
```

Those are the same six entries the 09-08 drift lost. Measured immediately
afterwards, in the worktree whose project root the session held:

```
$ jq -c '{has_ask:(.permissions|has("ask")), ask_len:((.permissions.ask//[])|length),
          allow:(.permissions.allow|length), deny:(.permissions.deny|length)}' .claude/settings.json
{"has_ask":true,"ask_len":0,"allow":114,"deny":48}
```

`ask` present but empty, `allow` and `deny` intact at 114 and 48 — the exact
shape of the 09-08 snapshot. And the discriminator this investigation used to
identify the writer reproduces:

```
$ diff <(jq -r 'keys_unsorted|.[]' .claude/settings.json) \
       <(jq -r 'keys_unsorted|.[]' <the 09-08 drift snapshot>)
KEY-ORDER-IDENTICAL=yes
```

The two files are **not** byte-identical — 26389 bytes against 23556, because
this worktree's committed content is develop-line and the 09-08 file was
main-line. The *content* differs while the *key order* matches exactly. That
is what a writer signature is: it survives a change of content, because it is
a property of the writer rather than of the data.

So the writer is established by observation: **Claude Code's permission-rule
editing surface rewrites the project `settings.json` of the session's own
project root, reordering the top-level keys into its insertion order and
leaving `ask` as a literal `[]`.** Every candidate the key-order table
eliminated stays eliminated; the surviving one is now witnessed rather than
narrowed to.

Three boundaries on that claim:

- **The primary checkout was not touched.** Its file kept mtime
  `2026-09-10 11:37:28` and sha256 `9f4818a6…`, still carrying the six
  entries. The write followed the session's working directory into the
  worktree. This is also why the earlier reading of `git status` inside the
  worktree showed ` M .claude/settings.json` while the primary was clean —
  two different files with the same relative path.
- **Intent, for this event, is the operator's stated intent**: asked directly,
  they answered that the deletion was deliberate. That closes the intent
  question for *this* occurrence only.
- **The 09-08 actor remains an inference, not an observation.** The signature
  match is strong evidence that the same surface produced it, and a
  deliberate operator action through that surface is the most economical
  explanation — but nobody observed 09-08, and this document does not record
  it as observed.

One consequence worth stating plainly: no change in this repository can
prevent this. The writer is the tool that owns the file, acting on a person's
instruction. What this repository can affect is whether such a state is
*noticed* and whether it can *heal* — which is why the latent finding below,
and cards t598 / t599, are where the repairable surface actually is.

### Axis 3 — is the empty `ask` dangerous?

```
$ jq -c '{has_ask:(.permissions|has("ask")), allow_len:(.permissions.allow|length),
          defaultMode:(.permissions.defaultMode // null)}' .claude/settings.local.json
{"has_ask":false,"allow_len":207,"defaultMode":"bypassPermissions"}
```

Two readings, and the pessimistic one is the supported one:

1. The 6 entries did **not** migrate to local settings — `has_ask: false`.
   E1 was a loss.
2. `defaultMode: "bypassPermissions"` is set in the local file, which takes
   precedence over project settings. Claude Code documents that mode as
   bypassing all permission checks (`claude --help`:
   `--allow-dangerously-skip-permissions … bypassing all permission checks`).
   So for any session reading this local file, the project `ask` list is
   **inert whether or not it is populated**.

This is why the operator's note that *"ask 복원의 즉시 효과가 불분명"* was
correct — and it reframes the exposure. Restoring `ask` buys nothing while
bypass mode is on. The load-bearing exposure is `bypassPermissions` itself,
which is a deliberate operating choice for this batch, not a drift.

### Latent second finding — an unreachable conflict branch, and no re-assert path

**Correction to an earlier framing in this investigation.** This was first
written up as "the merge is broken because the both-changed branch cannot
fire". That framing is wrong and is withdrawn: it contradicts the code's own
statement of intent. `pruneToShared` says so verbatim
(`internal/cli/update/merge/base.go:109-111`):

> Values always come from updated. A key present on both sides therefore
> enters the base carrying the template's value, **which is what makes a
> user's edit to that key read as their change during the merge.**

So `base == updated` for a shared key is **designed**, and a user's value
winning on a shared key is the intended behaviour. Calling that a defect
would collapse on one reading of the source.

The two real findings sit in the gap between that intent and its
instrumentation.

**(i) The conflict branch is unreachable for every shared key.**
`deepMergeMap` computes
(`internal/merge/strategies.go:419-420`):

```go
baseChanged := !valuesEqual(baseVal, curVal)
updChanged  := !valuesEqual(baseVal, updVal)
```

Because `pruneToShared` set `baseVal` **to** `updVal` for a shared key
(`base.go:128`, `pruned[key] = updatedVal`), `updChanged` is **always false**
there. The `default:` both-changed arm (`strategies.go:435-436`) — the one that
detects a template-versus-user conflict and would surface it — therefore
cannot execute for any shared key. Only `baseChanged && !updChanged` ("only
user changed", `strategies.go:427-429`) can.

A reader of this code believes `moai update` detects and reports
template-versus-user conflicts. It does not, and nothing says so. This is the
shape this repository already has a name for: **an unreachable check is an
instrument defect.** The scope is wider than this card's subject — it binds
**every shared key**, not `ask` alone.

**(ii) There is no path for the template to re-assert a value.** Whatever the
user's file holds for a shared key wins, whatever wrote it. For most keys that
is correct and desirable. For a **security-relevant key emptied by an
unidentified writer**, the emptied state is then preserved indefinitely and no
signal reaches the user.

One precise boundary, measurable in the same code: an **omitted** key heals,
an **empty array** does not. A key absent from the user's file is not shared,
so `pruneToShared` leaves it out of the base (`base.go:116-120`), the merge
sees the template introducing it, and the template's value lands. This is why
the in-repo toolpolicy writers — which omit `ask` entirely when empty
(`settings_region.go:204`) — would have self-healed on the next update, while
the literal `"ask": []` that the external writer actually produced would not.

Had E2's `git restore` not intervened, the empty `ask` would have survived
every subsequent `moai update`. Static-derived from reading `base.go` and
`strategies.go`; **NOT executed** — until it is reproduced in an isolated
sandbox with controls, (i) and (ii) are hypotheses.

### Secondary observation — a non-canonical evidence path in the ledger

Every ledger row records `preserved_path` under
`/Users/goos/moai/moai-adk-go/…` (lowercase `moai`), while the files actually
live under `/Users/goos/MoAI/moai-adk-go/…`. Both resolve on this
case-insensitive filesystem, so nothing is broken today; a later automated
read on a case-sensitive filesystem, or any exact-string comparison against
the canonical project root, would miss.

## Baseline-attribution

Every figure was measured in this run: the working-copy readings against
`/Users/goos/MoAI/moai-adk-go/.claude/settings.json` as it stands now, the
committed versions via `git show main:` / `git show HEAD:` from this worktree
(`HEAD = d60b2956b`), the drift snapshot from the preserved file named above,
and the mtime clustering from a single `find` over the two wipe-list roots.
The writer enumeration was produced by a read-only sweep in this worktree with
`/usr/bin/grep` (not the ugrep wrapper), and each row carries its own
`file:line`. No figure is carried from the card text: the card's `ask = 0`
premise was re-measured and found stale, which is finding 1.

**Coordinate attribution.** Every `file:line` in this document was re-measured
against this worktree at tree `715078afb` and each one was read back to confirm
it lands on the cited construct. An earlier revision of this document carried
coordinates from a first pass that had drifted by two to three lines
(`base.go:105-107` for the stated intent, `:126`, `strategies.go:426`, `:434`,
`base.go:117-121`); those are corrected above rather than left to be trusted. A
line citation decays like a `HEAD` reading — after a merge or rebase these
require re-measurement, which is why the SHA is pinned here.

## Gaps

- ~~**The exact triggering action for E1 is NOT established.**~~ **CLOSED by
  observation** — see § E1's mechanism, witnessed live. The triggering surface
  is Claude Code's `/permissions` rule editing, observed producing the exact
  shape and the exact key-order signature. The gap this bullet described no
  longer stands; it is kept struck-through rather than deleted so a reader can
  see which question the investigation opened with.
- ~~**Intent is NOT established.**~~ **Closed for the live event only**: asked
  directly, the operator stated the deletion was deliberate. It stays open for
  09-08, whose actor was never observed — the signature match makes the same
  surface the economical explanation, and this document does not upgrade that
  to an observation.
- **The tool version live at E1 is not recoverable.** The installed binary is
  `v3.2.0-rc.5 / 84fa4ece4`, built `2026-09-09T02:48:44Z` — *after* E1. What
  ran on 09-08 cannot be read off this tree, and the Claude Code version then
  in use is likewise unrecorded.
- **E2's actor is reconstructed, not witnessed.** The 263-file clustering and
  the two-remaining-modified-files state are consistent with
  update-then-`git restore`, and no other candidate explains byte-identity
  with `main` HEAD; but no log records the commands, and the drift ledger
  records only detections at `acquire`, never restorations. **A restoration
  leaves no trace in the ledger at all** — that is itself a gap in the
  instrumentation, not merely in this investigation.
- **The latent merge defect is unexecuted.** The "empty `ask` survives every
  update" conclusion is derived from static reading; no test was run against
  a live file.
- **`internal/web` was grepped, not fully read.** An indirect interaction
  between the settings console and the project file cannot be excluded with
  full confidence, though the commit that reworked that console shows no
  `permissions`/`ask` touches.

## Residual-risk

- Restoring the 6 entries without establishing E1's trigger does not prevent
  recurrence — the card says so, and the investigation supports it: the writer
  is outside this codebase, so no change in this repository can stop it.
- The instrumentation is asymmetric: drift is detected, preserved, and
  ledgered on every `acquire`, but restoration is invisible. A future drift
  could appear and be erased by routine maintenance again, and the ledger
  would show only the detection, making it look unresolved when it had been
  silently reverted.
- Any repair to `.claude/settings.json` in this repository must go through
  `internal/template/templates/.claude/settings.json.tmpl` as well, because
  `moai update` redeploys the file wholesale (`CLAUDE.local.md` §2.3). A
  local-only fix is reverted by the next update — and, per the latent merge
  defect, a template-only fix does not reach a user whose on-disk `ask` is
  already empty.
- `bypassPermissions` in the local file makes the entire project `ask` layer
  inert. Treating the restored 6 entries as a live defence would be an
  unobserved-claim: they are configured, not enforced, for as long as that
  mode is set.
