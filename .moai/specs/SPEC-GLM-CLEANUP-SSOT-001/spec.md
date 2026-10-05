---
id: SPEC-GLM-CLEANUP-SSOT-001
title: "GLM settings-axis cleanup — collapse five hand-maintained key lists onto one source"
version: "0.3.0"
status: in-progress
created: 2026-09-19
updated: 2026-09-19
author: manager-spec
priority: P1
phase: "v3.0.4 target"
module: "internal/config, internal/cli, internal/hook"
lifecycle: spec-anchored
tags: "glm, settings-local, cleanup, ssot, envkeys, session-end"
era: V3R6
tier: M
related_specs: [SPEC-CLIFIX-HYGIENE-001]
---

# SPEC-GLM-CLEANUP-SSOT-001

## §A Problem / Motivation

### A.1 The axis under discussion

The **settings axis** is the set of code paths that delete keys from the `env` object of
`.claude/settings.local.json`. Exactly three such functions exist in the tree:

| | Function | Location | Live caller | Keys deleted |
|---|---|---|---|---|
| **A** | `removeGLMEnv` | `internal/cli/launcher.go:330` | `applyCCMode` (`moai cc`) | 14 |
| **B** | `stripGLMCredsAndSetTeammateMode` | `internal/cli/settings.go:179` | **none** — `applyCGMode` returns `errCGRetired`; only tests call it | 13 |
| **C** | `cleanupGLMSettingsLocal` | `internal/hook/session_end.go:797` | SessionEnd hook | 9 |

`buildTmuxClearVars` is the **TMUX axis** and is explicitly out of scope: card t889 established
that comparing list lengths across the two axes is an axis-confusion error, not a measurement.

A's 14 keys: `MOAI_BACKUP_AUTH_TOKEN`, `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`,
`ANTHROPIC_DEFAULT_{HAIKU,SONNET,OPUS,FABLE}_MODEL`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`,
`API_TIMEOUT_MS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
`MOAI_STATUSLINE_CONTEXT_SIZE`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`.

Set relations, measured: **B = A − {`ANTHROPIC_DEFAULT_FABLE_MODEL`}**, and
**C = A − {`ANTHROPIC_DEFAULT_FABLE_MODEL`, `API_TIMEOUT_MS`,
`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
`MOAI_STATUSLINE_CONTEXT_SIZE`}** — C is a strict subset of A.

### A.2 Superseded premise (card t802 already closed it)

The originating card carried the premise *"`CLAUDE_CODE_MAX_CONTEXT_TOKENS` is cleared nowhere."*
**That premise is false in this tree and is recorded here as superseded, not as an open defect.**
All three of A / B / C delete that key today; commit `03d1904a7` (card t802) closed it. Verified:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER && go test -count=1 -run 'TestRemoveGLMEnvClearsEveryLiveKey|TestStripGLMCredsClearsEveryLiveKey|TestGLMCleanupRestoresBackedUpAuthToken' ./internal/cli/ -timeout 30m
→ PASS, exit 0
```

(The invocation above carries the §E environment scrub in one compound call and the `internal/cli`
`-timeout 30m`, as §E requires of every command this SPEC quotes.)

A future reader finding the premise in the card text should stop here: it is closed. What remains
is the list-divergence defect below.

### A.3 The remaining defect — five hand-maintained lists that disagree

Five lists spell the settings axis by hand today:

1. A's `delete()` block (`internal/cli/launcher.go`)
2. B's `delete()` block (`internal/cli/settings.go`)
3. C's `delete()` block (`internal/hook/session_end.go`)
4. `liveSettingsAxisKeys()` in `internal/cli/glm_settings_cleanup_test.go`
5. `liveHookWrittenKeys()` in `internal/hook/glm_settings_cleanup_test.go` (an identical 9-entry
   duplicate of #4's hook-side subset)

Nothing mechanically ties them together, so they have already drifted. Card t802 deliberately
landed four reproducers for the residue under the `residue_probe` build tag. Each asserts the
DESIRED state, so a **failure is the defect**, not a broken test:

| Probe | File | Currently | What it shows |
|---|---|---|---|
| `TestProbeCleanupLeavesLegacyAndOtherAxisKeys` | `internal/hook/glm_context_residue_probe_test.go` | FAILS | C leaves `ANTHROPIC_DEFAULT_FABLE_MODEL`, `MOAI_STATUSLINE_CONTEXT_SIZE`, `API_TIMEOUT_MS` |
| `TestProbeStrandedResidueSurvivesIndicatorLoss` | same file | FAILS | `CLAUDE_CODE_MAX_CONTEXT_TOKENS` survives in a file with no `ANTHROPIC_BASE_URL` indicator |
| `TestProbeStripGLMCredsLeavesLegacyKeys` | `internal/cli/glm_context_residue_probe_test.go` | FAILS | B leaves `ANTHROPIC_DEFAULT_FABLE_MODEL` |
| `TestProbeSettingsAxisListsAgree` | same file | FAILS | B keeps `FABLE_MODEL` that A clears |

### A.4 Who actually writes these keys

The only live writer into `settings.local.json`'s `env` is `ensureGLMCredentials`
(`internal/hook/session_start.go`). It writes `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`,
`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, and — via `maybeSet1MAutoCompactWindow` /
`maybeDeclareGLMContextWindow` — the pair `CLAUDE_CODE_AUTO_COMPACT_WINDOW` +
`CLAUDE_CODE_MAX_CONTEXT_TOKENS`. `MOAI_STATUSLINE_CONTEXT_SIZE` at `internal/cli/glm.go:526` sits
inside `buildTmuxInjectVars` (lines 510-548) — tmux axis; it never reaches the settings file from a
current binary.

Consequence: the five A-only keys reach `settings.local.json` **only from a file an older binary
wrote** (the `injectGLMEnv` path). The live route was closed by t802. Cleanup of those keys is
therefore legacy-residue cleanup, which is exactly why it is worth doing once, from one list,
rather than five times by hand.

### A.5 `injectGLMEnv` — what is measured, and in which tree

Every sentence below names the tree it describes. "Deleted" without a tree name is not a claim this
SPEC makes.

| Claim | Tree | Command | Observation |
|---|---|---|---|
| `injectGLMEnv` is absent | THIS tree, HEAD `0cca34439` (develop-based) | `grep -c "func injectGLMEnv" internal/cli/glm.go` | `0` |
| no caller exists | THIS tree | `grep -rn "injectGLMEnv(" --include='*.go' internal/` | 0 rows |
| `injectGLMEnv` is present | `main` | `git show main:internal/cli/glm.go \| grep -n "func injectGLMEnv"` | `989:func injectGLMEnv(...)`; writes `env["MOAI_BACKUP_AUTH_TOKEN"] = existing` at `:1005` |
| `main` is an ancestor of HEAD | both | `git merge-base --is-ancestor main HEAD` | exit 0 — HEAD is ahead of `main` |
| deletion history | THIS tree | `git log --oneline -S 'func injectGLMEnv' -- internal/cli/` | `03d1904a7` (card t802), `5026dabcb` (#1539) |

**Gap — not measured**: whether `main` still has production callers of its `injectGLMEnv`. This SPEC
asserts nothing in either direction about that, and no requirement below depends on it.

## §B Scope

**In scope**: introduce ONE canonical settings-axis declaration in `internal/config/envkeys.go`
exposing TWO derived views (cleanup / live); route all three production delete-lists onto the
cleanup view and both test lists onto the view each of them actually means; prove the resulting
guard is non-vacuous by mutation; widen C's indicator set conservatively; repair the one test
fixture that used a GLM-owned key as its stand-in for a user's own key; record the one deliberately
non-closed probe in prose.

### Out of Scope — the TMUX axis
- `buildTmuxClearVars` and `buildTmuxInjectVars` (`internal/cli/glm.go`) are the tmux
  injection/clear axis, not the settings axis. They are not touched, not merged, and their list
  lengths are never compared against the settings-axis set (card t889).
- **The tmux axis remains out of scope after the two-view split, and that is what the split is
  for.** The existing cross-axis guard `TestTmuxClearVarsCoversEveryLiveKeyItOwns`
  (`internal/cli/glm_settings_cleanup_test.go`) iterates the *live* view. Routing it onto the
  14-key cleanup view instead would require `buildTmuxClearVars` to carry
  `CLAUDE_CODE_TEAMMATE_DISPLAY` — measured absent from that list (`grep` over the function body:
  0 hits) — which is a tmux-axis edit this SPEC forbids. The guard therefore keeps the live view,
  unmodified, and no key is added to any tmux list. See REQ-8.

### Out of Scope — the producer
- `ensureGLMCredentials` (`internal/hook/session_start.go`) and its helpers
  `maybeSet1MAutoCompactWindow` / `maybeDeclareGLMContextWindow` are NOT modified. A marker-key
  approach (writing an explicit "MoAI wrote this" sentinel so cleanup can key on it instead of on
  an indicator heuristic) is a separate card and is explicitly excluded here.

### Out of Scope — removing the dead function B
- `stripGLMCredsAndSetTeammateMode` has no production caller. This SPEC does NOT delete it: tests
  depend on it and deletion is a different decision on a different axis. The fact is recorded here
  so a future card can decide its fate deliberately.

### Out of Scope — widening the indicator beyond the two admitted keys
- `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`, and
  `MOAI_STATUSLINE_CONTEXT_SIZE` are deliberately NOT added to the indicator set. Each can be set
  legitimately by a user who never touches GLM; the last is documented in the code as an explicit
  user override. See REQ-5 and §E.

### Out of Scope — repurposing `GLMEnvVarSet`
- `config.GLMEnvVarSet()` serves the inject↔clear parity obligation of SPEC-CLIFIX-HYGIENE-001
  (REQ-HYG-001-003) with a 3-key set. It is NOT widened and NOT repurposed; the new canonical set
  is a sibling in the same file.

### Out of Scope — GLM integration testing
- No GLM integration test is run in this repository, and no real settings file or GLM credential is
  read or written by any test added here (CLAUDE.local.md §13).

## §C Requirements (GEARS)

### REQ-1 — One canonical declaration, two derived views

The `config` package shall expose a single canonical, ordered, documented settings-axis GLM key
declaration as a sibling of `GLMEnvVarSet()` in `internal/config/envkeys.go`, together with exactly
two views derived from it:

| View | Membership | Consumers |
|---|---|---|
| `SettingsAxisCleanupKeys()` | the full 14 keys of §A.1 — the live set plus the 5-key legacy tail `{ANTHROPIC_DEFAULT_FABLE_MODEL, API_TIMEOUT_MS, CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC, CLAUDE_CODE_TEAMMATE_DISPLAY, MOAI_STATUSLINE_CONTEXT_SIZE}` | the three production cleanup functions (REQ-2/3/4) |
| `SettingsAxisLiveKeys()` | the 9 keys a live producer can put in the file — `ANTHROPIC_AUTH_TOKEN`, `MOAI_BACKUP_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS` | the existing live-key guards and the cross-axis tmux-parity guard (REQ-8) |

Both views shall derive from ONE ordered declaration carrying a per-key live/legacy classification;
neither view shall be a second literal list. `SettingsAxisCleanupKeys()` is the live view plus the
legacy tail, so the set relation is a property of the declaration rather than of two enumerations
that must be kept in step by hand.

The declaration shall carry a doc comment stating (a) that it is the settings axis (keys deleted
from `settings.local.json`'s `env`), (b) that it is distinct from the tmux axis, and (c) that five
former hand-maintained lists now derive from it.

### REQ-2 — `removeGLMEnv` derives from the cleanup view

The `removeGLMEnv` function shall delete exactly `SettingsAxisCleanupKeys()` — **except
`ANTHROPIC_AUTH_TOKEN`, whose disposition REQ-7 owns** — and shall not carry its own hand-written
key literals. The OAuth restore (REQ-7) shall remain ordered before the deletion loop, and that
loop shall skip the token pair per the REQ-7 carve-out.

### REQ-3 — `stripGLMCredsAndSetTeammateMode` derives from the cleanup view

The `stripGLMCredsAndSetTeammateMode` function shall delete exactly `SettingsAxisCleanupKeys()` —
**except `ANTHROPIC_AUTH_TOKEN`, whose disposition REQ-7 owns** — and shall not carry its own
hand-written key literals. Its existing teammate-mode side effect is unchanged, the OAuth restore
(REQ-7) shall remain ordered before the deletion loop, and that loop shall skip the token pair per
the REQ-7 carve-out.

Consequence: `ANTHROPIC_DEFAULT_FABLE_MODEL`, which it keeps today, is newly cleared.

### REQ-4 — `cleanupGLMSettingsLocal` derives from the cleanup view

**While** the indicator gate has admitted the file, `cleanupGLMSettingsLocal` shall delete exactly
`SettingsAxisCleanupKeys()` — **except `ANTHROPIC_AUTH_TOKEN`, whose disposition REQ-7 owns** — and
shall not carry its own hand-written key literals. The OAuth restore (REQ-7) shall remain ordered
before the deletion loop, and that loop shall skip the token pair per the REQ-7 carve-out.

This is a deliberate widening: SessionEnd will newly delete `API_TIMEOUT_MS`,
`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
`MOAI_STATUSLINE_CONTEXT_SIZE`, and `ANTHROPIC_DEFAULT_FABLE_MODEL` — but only inside the indicator
gate, matching what `moai cc` already does today.

### REQ-5 — Conservative indicator widening (option 가' — two keys)

**Where** a settings file carries either `ANTHROPIC_BASE_URL` or `MOAI_BACKUP_AUTH_TOKEN`,
`cleanupGLMSettingsLocal` shall treat the file as MoAI-written and proceed to cleanup.

The indicator set shall contain **exactly those two keys** and shall NOT include
`MOAI_STATUSLINE_CONTEXT_SIZE`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, or
`CLAUDE_CODE_MAX_CONTEXT_TOKENS`.

#### Why `MOAI_BACKUP_AUTH_TOKEN` is admitted

**Primary ground (tree-independent): the only code path that writes that key is the GLM injection
path, so a file carrying it is by definition a GLM user's file.** No non-MoAI flow produces it, and
no user sets it by hand — it exists solely as the place the GLM injection path parks a pre-existing
OAuth token so cleanup can put it back. That argument survives the key's writer being renamed,
moved, or re-implemented, because it is about what the key MEANS, not about where it currently
lives.

Secondary support, tree-specific (measured in this tree, HEAD `0cca34439`): `grep -rn
'MOAI_BACKUP_AUTH_TOKEN\|EnvMoaiBackupAuthToken' --include='*.go' internal/` outside tests returns
readers and deleters only — `internal/cli/launcher.go:356,358`, `internal/cli/settings.go:182,184`,
`internal/hook/session_end.go:852,858`, plus a membership entry at
`internal/cli/mcp_claude.go:230`. The writer lives on `main` at `internal/cli/glm.go:1005`, inside
`injectGLMEnv` (§A.5). This is support, not the ground: if the writer moved tomorrow the primary
argument would be unchanged.

#### Why the three excluded keys are excluded

| Excluded key | Reason |
|---|---|
| `MOAI_STATUSLINE_CONTEXT_SIZE` | The code documents it as a **documented explicit user override**, not a MoAI-exclusive marker. `internal/statusline/memory.go:143` lists it as resolution priority #1, labelled "explicit user override"; `internal/config/envkeys.go:59-63` declares it as the override for a context-size mismatch **in general** — "Useful when the upstream provider reports a context size that does not match the actual API limit" — with GLM given only as an `e.g.` example. A user who never touches GLM can legitimately set it. |
| `CLAUDE_CODE_AUTO_COMPACT_WINDOW` | A non-GLM user can legitimately set it themselves. |
| `CLAUDE_CODE_MAX_CONTEXT_TOKENS` | A non-GLM user can legitimately set it themselves. |

**The harm the exclusion prevents is concrete, not theoretical.** `internal/hook/session_end.go:854`
reads, verbatim:

```go
	} else {
		delete(env, config.EnvAnthropicAuthToken)
	}
```

If the gate opened on `MOAI_STATUSLINE_CONTEXT_SIZE` alone, a file carrying that override and the
user's own `ANTHROPIC_AUTH_TOKEN` but no backup would take the `else` branch and **SessionEnd would
delete that user's own API key** — the exact quiet regression the conservative option was chosen to
avoid, arriving through an *admitted* key rather than an excluded one.

**Provenance**: option (가) was the lead's 2026-09-19 verdict; option (가') is the same lead's
2026-09-19 revision of the admitted set from three keys to two, on the ground above. Both are
settled and are not re-opened by this SPEC.

#### Consequence — a file that was stranded is now cleaned

Under (가') a settings file carrying `MOAI_BACKUP_AUTH_TOKEN` and a GLM `ANTHROPIC_AUTH_TOKEN` but
**no** `ANTHROPIC_BASE_URL` is admitted, and the backed-up OAuth token is restored as
`ANTHROPIC_AUTH_TOKEN`. Before this change that file was declined by the gate and the user's
backed-up token stayed stranded. This is the net benefit of the widening and is pinned by a new
named test (AC-014) rather than argued in prose.

### REQ-6 — The early-return gate is preserved as intentional design

`cleanupGLMSettingsLocal` shall retain its early return for a settings file carrying no indicator.
This gate is intentional design, not a defect: its doc comment states the intent ("Claude Code's
OAuth flow never sets this variable"), and the landed default-suite negative control
`TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone`
(`internal/hook/glm_settings_cleanup_test.go`) pins that a file which was never in GLM mode keeps
its own `ANTHROPIC_AUTH_TOKEN`. Removing the gate would destroy a non-GLM user's own API key.

That negative control shall continue to pass unmodified.

### REQ-7 — OAuth restore behaviour preserved in all three functions

`removeGLMEnv`, `stripGLMCredsAndSetTeammateMode`, and `cleanupGLMSettingsLocal` shall each continue
to restore a backed-up OAuth token from `MOAI_BACKUP_AUTH_TOKEN` into `ANTHROPIC_AUTH_TOKEN`
**before** deleting keys, and the deletion of the canonical set shall remain ordered **after** that
restore.

All three carry the same restore-before-delete shape today (`internal/cli/launcher.go:356-358`,
`internal/cli/settings.go:182-184`, `internal/hook/session_end.go:852-858`), and the canonical set
contains BOTH `MOAI_BACKUP_AUTH_TOKEN` and `ANTHROPIC_AUTH_TOKEN`. A naive implementation that
replaces each delete-block with a single loop over the canonical set, placed where the old block
began, therefore **deletes the backup before anything reads it** — silently, in all three
functions at once. The ordering is the requirement; the loop is only the mechanism.

#### The carve-out — ordering alone is NOT sufficient

**Restore-before-delete does not, by itself, preserve the token.** Even when the restore runs first,
a subsequent loop over the full cleanup view reaches `ANTHROPIC_AUTH_TOKEN` — a member of that
view — and deletes the value the restore just wrote. The naive two-step "restore, then loop over
everything" therefore destroys the user's OAuth token in all three functions, and does so while
satisfying the ordering clause above.

The requirement is therefore stated in two parts, and both bind:

1. **Ordering** — the restore is ordered before the deletion of the canonical set (above).
2. **Carve-out** — `ANTHROPIC_AUTH_TOKEN`'s disposition is decided **exclusively by the restore
   branch**, and the loop that iterates `SettingsAxisCleanupKeys()` shall **skip both token keys**
   (`ANTHROPIC_AUTH_TOKEN`, `MOAI_BACKUP_AUTH_TOKEN`). `MOAI_BACKUP_AUTH_TOKEN` is consumed by the
   restore branch's own straight-line delete rather than by the loop; skipping it in the loop keeps
   the pair's handling in one place instead of splitting it across two mechanisms.

**Exactly one key is carved out of "exactly", and the ground is that it is a restore TARGET, not
residue.** No other member of the cleanup view has a conditional disposition; every other key is
deleted unconditionally. A future reader asking "if not `exactly`, then what else is excepted?"
has the complete answer here: `ANTHROPIC_AUTH_TOKEN`, because the restore branch owns it.

**This is preservation, not a behaviour change.** The shape the carve-out describes is the shape the
code already has today: an `if`/`else` handling the token pair, plus straight-line deletes for the
rest (`internal/hook/session_end.go:852-858` and its two CLI siblings). The requirement exists so
the run phase PRESERVES that shape rather than "simplifying" the `if`/`else` and the deletes into
one uniform loop. The landed guards encode the same carve-out on the test side —
`internal/cli/glm_settings_cleanup_test.go` skips exactly this pair in its all-keys-gone traversal
(`switch key { case config.EnvAnthropicAuthToken, "MOAI_BACKUP_AUTH_TOKEN": continue }`).

Pinned today by `TestGLMCleanupRestoresBackedUpAuthToken` (`internal/cli`, covering both CLI
functions) and `TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken` (`internal/hook`).

### REQ-8 — Each test list derives from the view it actually means

`liveSettingsAxisKeys()` (`internal/cli/glm_settings_cleanup_test.go`) and `liveHookWrittenKeys()`
(`internal/hook/glm_settings_cleanup_test.go`) shall each derive from one of the two REQ-1 views
rather than re-spelling a list, and shall derive it directly or by one explicit, named filter —
never by a second literal.

Both name the **live** view today, and both shall continue to: their doc comments say "every env key
`ensureGLMCredentials` can write", which is the live view's definition and not the cleanup view's.
Widening them to the cleanup view would silently change what their callers assert and would break
the out-of-scope tmux-parity guard (§B).

The doc comments of both helpers shall be updated to name the view they now derive from, so that a
later reader is not left with a comment describing a literal list that no longer exists.

### REQ-9 — The collapse shall be demonstrated non-vacuous by mutation

**When** the five lists have been folded onto the canonical declaration, the run-phase evidence
shall record a **named** test observed red under a deliberate one-key divergence of one consumer,
that test's verbatim failure output, and the same test observed green after the consumer is
restored.

A pure list-merge whose guard was never observed failing is not an accepted outcome: a guard that
cannot be shown to turn red has not been shown to guard anything.

### REQ-10 — The non-closed probe carries its rationale in prose

`TestProbeStrandedResidueSurvivesIndicatorLoss` shall still fail after this SPEC lands. That is the
accepted outcome, not an open defect, and the probe file shall record in prose:

1. that option (가') — conservative indicator widening to exactly `{ANTHROPIC_BASE_URL,
   MOAI_BACKUP_AUTH_TOKEN}` — was chosen, by whom, and when (the lead, 2026-09-19, revising the
   earlier three-key option 가 on the same day);
2. why the stranded case is deliberately not closed: widening the indicator to the context-window
   pair — or to `MOAI_STATUSLINE_CONTEXT_SIZE`, a documented explicit user override — would delete
   a non-GLM user's own settings;
3. that the only route into the stranded state is a file written by an older binary, the live route
   having been closed by card t802.

Without that record the next reader reopens a settled decision as a defect.

### REQ-11 — Producer and dead function untouched

The `ensureGLMCredentials` producer shall not be modified, and
`stripGLMCredsAndSetTeammateMode` shall not be deleted.

### REQ-12 — A user-owned-key assertion shall assert on a user-owned key

`internal/hook/session_end_test.go`'s `TestCleanupGLMSettingsLocal` asserts, under the heading
"Check non-GLM var preservation", that a key seeded in the fixture **survives** cleanup. The key it
uses as its stand-in for "a user's own key" is `CLAUDE_CODE_TEAMMATE_DISPLAY` — which
`internal/config/envkeys.go` declares as "the legacy GLM activation indicator env var … still
cleared to deactivate legacy GLM mode", and which `removeGLMEnv` already deletes. The fixture picked
a GLM-owned key, and REQ-4 makes SessionEnd delete it too.

The test's **intent shall be preserved and its stand-in key replaced**: the surviving-key assertion
shall be made on a key that appears in **no** cleanup list — `CUSTOM_VAR`, the stand-in the t802
guards already use (`internal/hook/glm_settings_cleanup_test.go:167`,
`internal/cli/cg_mode_hardening_test.go:174`).

The assertion shall NOT be weakened. Changing `wantOtherPresent` to `false`, deleting the assertion,
or narrowing it to cases the new behaviour happens to satisfy is the control being removed, not the
control being repaired. Only the fixture value changes.

## §D Evidence basis

Every measurement in §A was taken by the lane in this tree at base `0cca34439`, by reading the
three functions and the two test lists and by running the named `go test` invocation quoted in
§A.2. The four probe failures are the landed reproducers' own output under
`-tags residue_probe`.

## §E Constraints / Non-Goals

- **No production code in plan phase.** This SPEC records intent only.
- **Test isolation**: `t.TempDir()` only. No test may read or write a real settings file or a real
  GLM credential.
- **No GLM integration tests** in this repository (CLAUDE.local.md §13).
- **Package-scoped verification only**: `go test ./internal/cli/... ./internal/hook/...
  ./internal/config/...`; `internal/cli` needs `-timeout 30m`. Never `go test ./...` locally.
- **Environment scrub in one compound invocation** — a separate `unset` does not carry into the
  next command.
- **Three deliberate behaviour changes are to be stated as such** in the run-phase report and in
  CHANGELOG at sync. Each is intended; none may be reported as incidental:
  1. **REQ-4** — SessionEnd newly deletes five more keys (`ANTHROPIC_DEFAULT_FABLE_MODEL`,
     `API_TIMEOUT_MS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
     `MOAI_STATUSLINE_CONTEXT_SIZE`), inside the gate only.
  2. **REQ-5** — a settings file carrying `MOAI_BACKUP_AUTH_TOKEN` but no `ANTHROPIC_BASE_URL` is
     newly admitted, and the user's backed-up OAuth token is newly restored as
     `ANTHROPIC_AUTH_TOKEN` where it was previously stranded.
  3. **REQ-12** — `TestCleanupGLMSettingsLocal`'s user-owned-key fixture changes from
     `CLAUDE_CODE_TEAMMATE_DISPLAY` to `CUSTOM_VAR`; the assertion's intent is unchanged and its
     strength is unchanged.
- **Explicitly NOT a fourth behaviour change — the `ANTHROPIC_AUTH_TOKEN` carve-out (REQ-7).**
  `ANTHROPIC_AUTH_TOKEN` is a member of the 14-key cleanup view (§A.1) *and* the target of the
  OAuth restore, so composing "delete exactly the view" with "restore first" literally would destroy
  the restored token. REQ-7's carve-out stops that. Observable behaviour is **unchanged** from
  today — the carve-out preserves the existing `if`/`else` shape — so it is recorded here as a
  preservation obligation rather than as a behaviour change, and the list above stays at three.
- **Non-goal**: reducing the number of cleanup functions from three to fewer. Only their key lists
  are unified.

## §F Cross-References

- `internal/config/envkeys.go` — `GLMEnvVarSet()` (SPEC-CLIFIX-HYGIENE-001 REQ-HYG-001-003), the
  sibling the new set sits beside.
- `internal/cli/launcher.go:330` — A.
- `internal/cli/settings.go:179` — B.
- `internal/hook/session_end.go:797` — C.
- `internal/hook/session_start.go` — `ensureGLMCredentials`, the live producer (out of scope).
- `internal/cli/glm.go:510-548` — `buildTmuxInjectVars` (tmux axis, out of scope).
- `internal/cli/glm_context_residue_probe_test.go`, `internal/hook/glm_context_residue_probe_test.go`
  — the four landed reproducers.
- `internal/cli/glm_settings_cleanup_test.go`, `internal/hook/glm_settings_cleanup_test.go` — the
  two test key lists and the negative control; the former also holds the out-of-scope cross-axis
  guard `TestTmuxClearVarsCoversEveryLiveKeyItOwns`.
- `internal/hook/session_end_test.go` — `TestCleanupGLMSettingsLocal`, whose user-owned-key stand-in
  fixture REQ-12 repairs.
- `internal/statusline/memory.go:143`, `internal/config/envkeys.go:59-63` — the two places the code
  documents `MOAI_STATUSLINE_CONTEXT_SIZE` as an explicit user override (REQ-5 exclusion ground).
- Card t802 (origin, R3+R4), commit `03d1904a7`. Card t889 (axis separation). Card t888 (this).

## §G HISTORY

- **2026-09-19** v0.3.0 — plan-audit iteration 2 findings closed (FAIL 0.76; iteration ceiling
  reached, so the five fixes are verified individually rather than by a third full audit). Author:
  manager-spec. Changes, one per finding: **N1** — AC-002's fixture rule now excludes
  `MOAI_BACKUP_AUTH_TOKEN`, following the convention already landed in
  `internal/cli/glm_settings_cleanup_test.go` (`glmLiveDirtyEnv` / `assertFixtureIsDirty`,
  `absent by design`) and `internal/hook/glm_settings_cleanup_test.go`, with those locations cited
  in the criterion; the restore path stays covered by AC-010. **N2** — REQ-7 gains the
  `ANTHROPIC_AUTH_TOKEN` carve-out (the deletion loop skips both token keys; the restore branch owns
  their disposition), REQ-2/3/4 carry the matching `except` clause, and §E records it as preservation
  rather than a fourth behaviour change; plan.md §B / §E D8 / M2 / M4 / M5 restate it. **N3** —
  `CARD_BASE` is re-resolved at each use site instead of pinned at run-phase entry, per
  `.claude/rules/local/gitflow-lane-protocol.md` §8, with §8's post-merge limitation recorded in
  §D.0. **N4** — M3's `CUSTOM_VAR` seeding is narrowed to the two cases that already carry a
  stand-in key (measured: case 2 never seeds one), and AC-015's control is restated as a
  no-weakening condition matching REQ-12. **N5** — §C's progress-dependent positive control is
  replaced by an instrument-sanity control that is non-zero at dispatch. REQ count unchanged (12);
  AC count unchanged (16) — no criterion was added.
- **2026-09-19** v0.2.0 — plan-audit iteration 2 amendment (FAIL 0.69 → rework). Author:
  manager-spec. Changes: indicator set revised to option (가') = `{ANTHROPIC_BASE_URL,
  MOAI_BACKUP_AUTH_TOKEN}`, `MOAI_STATUSLINE_CONTEXT_SIZE` excluded as a documented explicit user
  override (REQ-5, audit D3); canonical declaration split into two derived views so the
  out-of-scope tmux-parity guard is not broken (REQ-1/REQ-8, audit D2); REQ-7 bound to all three
  cleanup functions with the restore-before-delete ordering stated (audit D7); new REQ-12 repairing
  the `CLAUDE_CODE_TEAMMATE_DISPLAY` stand-in fixture; new §A.5 naming the tree for every
  `injectGLMEnv` claim; REQ-9 re-subjected to the run-phase evidence (audit D12); §A.2 quoted
  command given its §E scrub and `-timeout 30m` (audit D9). Problem statement (§A.1-§A.4 set
  relations, four RED probes) unchanged — the audit reproduced it with zero mismatches.
- **2026-09-19** v0.1.0 — initial draft (plan-phase). Author: manager-spec. Measurements taken by
  the lane in this tree at `0cca34439`. Card premise about `CLAUDE_CODE_MAX_CONTEXT_TOKENS`
  recorded as superseded (§A.2). Indicator-widening scope fixed to option (가) by lead verdict.
