# Plan — SPEC-GLM-CLEANUP-SSOT-001

> Milestones are ordered by decision-reversibility: the data-model decision (what the canonical
> declaration IS, and which two views derive from it) comes first, then the behaviour-visible
> consumer that changes what a user observes, then the mechanical routing, then the demonstration
> and the prose record.

## §A Context

Five hand-maintained lists spell one axis and have already drifted (spec.md §A.3). Card t802 landed
four reproducers under the `residue_probe` build tag which assert the desired state, so their
current failure IS the defect statement. This card folds the five lists onto one declaration with
two derived views, widens the SessionEnd indicator conservatively, repairs the one fixture that used
a GLM-owned key as its stand-in for a user's own key, proves the resulting guard turns red under
mutation, and records in prose the one probe that deliberately stays red.

## §B Known Issues / Traps

- **Axis confusion.** `buildTmuxClearVars` is the tmux axis. Do not compare its length against the
  settings-axis set; do not fold it in (card t889).
- **One flat 14-key set breaks the tmux-parity guard.** `TestTmuxClearVarsCoversEveryLiveKeyItOwns`
  iterates `liveSettingsAxisKeys()`; widening that to 14 requires `CLAUDE_CODE_TEAMMATE_DISPLAY` in
  `buildTmuxClearVars()`, where it is measurably absent. Two views, not one set (D3).
- **A grep for env-var names cannot see this defect.** The lists spell most keys as Go identifiers
  (`config.Env…`), not string literals, so a fully hand-written list yields zero hits. Verify at
  runtime (D7).
- **The canonical set contains the backup key.** A delete-loop placed where an old delete-block
  began destroys `MOAI_BACKUP_AUTH_TOKEN` before the restore reads it — in all three functions (D8).
- **Ordering alone does not save the token — the loop must skip it.** The canonical set also
  contains `ANTHROPIC_AUTH_TOKEN`, the restore's TARGET. Restoring first and then looping over the
  whole view deletes the value the restore just wrote, in all three functions, while satisfying the
  ordering rule. The loop skips both token keys; the restore branch owns their disposition (REQ-7
  carve-out, D8).
- **`MOAI_STATUSLINE_CONTEXT_SIZE` is a user override, not a MoAI marker.** It is cleaned (it is in
  the cleanup view) but it never admits a file (it is not an indicator). Those are two different
  roles for one key; do not collapse them (D4).
- **`cleanupGLMSettingsLocal` returns early under `MOAI_LAUNCH_PROVIDER`.** Any new test in
  `internal/hook` must call `scrubGatewayEnv(t)` first, or an exclusion assertion passes on the
  gateway early return instead of on the code under test.
- **`GLMEnvVarSet()` is not the place.** It answers a different question (inject↔clear parity for
  3 keys, SPEC-CLIFIX-HYGIENE-001). Add a sibling; do not widen it.
- **The early-return gate is design, not a bug.** Removing it destroys a non-GLM user's own API
  key; a landed negative control already pins that.
- **A stale `-run` filter reports the filter's exit code.** Run the named tests without a shell
  pipeline that can swallow the verdict.
- **A cached result is not a measurement of this tree.** Use `-count=1`.

## §C Pre-Flight (run-phase entry checks)

1. `git rev-parse --short HEAD` and `git branch --show-current` — confirm the worktree and branch.
2. Reproduce the four probe failures BEFORE any edit, and keep the verbatim output as the RED
   baseline:
   ```bash
   unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER && go test -tags residue_probe -count=1 ./internal/cli/... ./internal/hook/... -timeout 30m
   ```
3. **Establish** — do not assume — the default-suite baseline, so a later red is attributable:
   ```bash
   unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER && go test -count=1 ./internal/config/... ./internal/hook/... ./internal/cli/... -timeout 30m 2>&1 | tee .moai/state/verify/t888/baseline-default-suite.txt; echo "EXIT:${PIPESTATUS[0]}"
   ```
   **This SPEC makes no claim about the current state of that suite.** Plan-phase did not run it
   (the plan-audit recorded the same gap), so "the default suite is green today" is an *unverified
   pre-flight step the run phase must actually execute and record*, never a premise. If the
   baseline comes back red, stop and report which packages were already red BEFORE any edit —
   otherwise AC-012 and the M6 tmux-guard analysis become impossible to tell apart.
   Per the lane discipline, this is a heavy run: take the resource slot first
   (`moai slot acquire`) rather than racing another lane.

**Baselines already measured (cite, do not re-run).** All four probes FAIL at base `0cca34439`:
`.moai/state/verify/t888/baseline-probe-hook.txt` and `baseline-probe-cli.txt`, both `EXIT:1`. The
13 existing hook cleanup guards — including the negative control
`TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone` — all PASS:
`.moai/state/verify/t888/baseline-guards-hook.txt`, `EXIT:0`.

**Card base, for every range-scoped check below — re-resolved at each use, never pinned.**
`.claude/rules/local/gitflow-lane-protocol.md` §8 [HARD] governs this and states the rule
literally: the left edge of a range predicate is re-derived **at the moment of reading** with
`CARD_BASE=$(git merge-base develop HEAD)`, and the value is **not pinned**. Pinning means keeping
one resolved value — in a shell variable exactly as much as in a literal SHA; the hazard is the
retention, not the spelling.

The consequence of pinning is not hypothetical in this repository: the lane procedure *requires*
absorbing `develop` (`CLAUDE.local.md` §4.1 — "창을 받으면 … `git merge develop` 흡수 → 병합 트리에서
재측정"), and absorption moves the merge-base forward. A left edge pinned before that absorption
stays at the old fork point, so every range below silently takes in the **other cards'** commits
that came with the absorbed `develop` — AC-011 ("producer unmodified") goes red on somebody else's
`session_start.go` edit, and AC-009 / AC-013 / AC-015 load foreign hunks. So:

```bash
# Correct — resolution and use travel in ONE invocation, at the moment of the check:
CARD_BASE=$(git merge-base develop HEAD) && git diff --stat "$CARD_BASE"..HEAD -- <path>
```

**§C resolves nothing and keeps nothing.** There is no `CARD_BASE` to carry out of pre-flight; each
range check in `acceptance.md` re-resolves its own.

**Limitation, carried from gitflow-lane-protocol.md §8: a range predicate is valid only BEFORE the
card is merged.** Once the card lands on `develop`, `git merge-base develop HEAD` returns the card's
own tip, the range is empty, and every absence assertion passes vacuously. Post-merge evidence is
made by tree-identity (the merge commit's tree equals the re-measured tree), never by these ranges.

**Instrument-sanity control (N5 — replaces the former progress-dependent one).** The earlier §C
control was `git diff --name-only "$CARD_BASE"..HEAD | wc -l  # MUST be ≥ 1`, which is **necessarily
`0` at run-phase entry** — measured in this tree at `0cca34439`, where HEAD = `develop` = merge-base:
exit 0, output `0`. A check placed where its condition cannot yet differ proves nothing by passing
and, worse, made pre-flight open by reporting a gap (`verification-completeness.md` §1.2(a)).

Of the two repairs the audit offered, the second is taken: **the control is moved onto an axis that
does not depend on the card having made progress.** Re-defining it as post-milestone was rejected
because §C is entered before any milestone exists, so the control would simply be absent from
pre-flight — and the progress-dependent form already lives, correctly, at each use site
(`acceptance.md` AC-009, AC-011, AC-013). What pre-flight actually needs to establish is that the
range *machinery* works — merge-base resolves, and `git diff --name-only` over a range built from it
produces output:

```bash
CARD_BASE=$(git merge-base develop HEAD) && git diff --name-only "$CARD_BASE~1"..HEAD | wc -l
```

**Expected: ≥ 1, at run-phase entry, before any edit.** The range deliberately reaches one commit
behind the base, so it is non-empty by construction whatever the card has done. Measured in this
tree at `0cca34439`: `1`. A `0` here means the instrument is broken (merge-base unresolvable,
`develop` missing, wrong tree) — stop and report, rather than trusting any range check below it.

## §D Constraints (HARD)

- `t.TempDir()` only; never a real settings file or GLM credential.
- No GLM integration tests in this repository (CLAUDE.local.md §13).
- Package-scoped verification; `internal/cli` needs `-timeout 30m`; never `go test ./...` locally.
- The environment scrub and the command travel in ONE compound invocation — each Bash call is a
  fresh process, so a standalone `unset` scrubs nothing.
- Do not modify `ensureGLMCredentials`. Do not delete `stripGLMCredsAndSetTeammateMode`.
- No time estimates anywhere in run-phase reporting.

## §E Self-Verification (design decisions)

### D1 — The canonical set lives in `internal/config/envkeys.go` as a sibling of `GLMEnvVarSet()`

Both `internal/cli` and `internal/hook` already import `internal/config`, so a set declared there is
reachable from every consumer without a new dependency edge. Declaring it in either consumer package
would force the other to import it, inverting the layering.

Rejected: widening `GLMEnvVarSet()` itself. It is the parity anchor for a different obligation
(3 keys, inject↔clear); widening it silently changes what that obligation asserts.

### D2 — The set is ordered and documented, not a bare map

Order makes the diff of any future addition readable, and the doc comment is where the axis
distinction (settings vs tmux) and the "five lists derive from this" fact live. Without that comment
the next reader re-derives the axis by guessing — which is how the drift started.

### D3 — One declaration, two derived views — and the tmux guard keeps the LIVE view

The naive shape (one 14-key set, every consumer on it) breaks an out-of-scope axis. Measured:
`TestTmuxClearVarsCoversEveryLiveKeyItOwns` (`internal/cli/glm_settings_cleanup_test.go:172`)
iterates `liveSettingsAxisKeys()` and requires each key to appear in `buildTmuxClearVars()`.
Widening that helper to 14 keys pulls `CLAUDE_CODE_TEAMMATE_DISPLAY` into the loop, and that key is
absent from `buildTmuxClearVars()` (grep over the function body: 0 hits) — so the guard goes red
and AC-012 (three packages `ok`) is unsatisfiable.

All three escape routes from the naive shape violate something: adding the key to the tmux list is
a forbidden tmux-axis edit (§B, §G); widening the guard's exclusion `switch` weakens an existing
cross-axis guard that nothing else replaces; not widening the helper at all violates REQ-8.

**The resolution is that the naive shape was wrong, not that one of its casualties is acceptable.**
ONE ordered declaration in `internal/config/envkeys.go` classifies each key live-or-legacy; two
views read it:

- `SettingsAxisCleanupKeys()` — all 14. The three production cleanup functions use it.
- `SettingsAxisLiveKeys()` — the 9 live keys. The existing live-key guards and the tmux-parity
  guard use it, **unmodified**.

The SSOT property holds (one declaration, no second literal), REQ-8 is satisfied by each test list
deriving from the view it actually means, and the tmux axis is not touched at all. The tmux-parity
guard's continued green is pinned by AC-013.

### D4 — Indicator widening is conservative (option 가') — lead verdict, not re-openable here

Admitted, exactly two: `ANTHROPIC_BASE_URL`, `MOAI_BACKUP_AUTH_TOKEN`.

`MOAI_BACKUP_AUTH_TOKEN` is admitted on a tree-independent ground: **the only code path that writes
it is the GLM injection path, so a file carrying it is by definition a GLM user's file.** That holds
however the writer is re-organised. The tree measurements in spec.md §REQ-5 are support beneath
that ground, not the ground.

Excluded: `MOAI_STATUSLINE_CONTEXT_SIZE` (the code documents it as an explicit user override —
`internal/statusline/memory.go:143` resolution priority #1, `internal/config/envkeys.go:59-63`
scoping it to provider context-size mismatch in general with GLM as an `e.g.`), plus
`CLAUDE_CODE_AUTO_COMPACT_WINDOW` and `CLAUDE_CODE_MAX_CONTEXT_TOKENS`. Admitting any of the three
would let `session_end.go:854`'s `else { delete(env, EnvAnthropicAuthToken) }` branch fire on a
non-GLM user's file and delete that user's own API key.

**This revises the earlier three-key option (가)** — same lead, same day, on the ground above. Both
the original verdict and the revision are settled; the run phase encodes them and does not re-open
them.

The marker-key alternative (option 다: have the producer write an explicit sentinel so cleanup keys
on a fact rather than a heuristic) is the correct long-term shape and is a separate card.

### D5 — One probe stays red, deliberately

`TestProbeStrandedResidueSurvivesIndicatorLoss` is not closable under D4. It is kept red **and
annotated**, because a red test with no explanation is indistinguishable from an unfinished one.
The annotation is a REQ (REQ-10), not a courtesy.

### D6 — The mutation demonstration is an acceptance criterion, not advice

A guard folded from five lists onto one is only a guard if divergence turns it red. The
demonstration is recorded with the named test and its verbatim output (AC-008).

### D7 — The collapse is verified at runtime, not by grepping for key names

The obvious check — grep the consumers for env-var names — cannot see the defect it is aimed at.
Measured on the unchanged tree: the five lists spell most keys as **Go identifiers**
(`config.EnvAnthropicDefaultFableModel`), not string literals, so a consumer that keeps its entire
hand-written list — the exact forbidden state — yields no hits. A consumer could pass the check
while violating every one of REQ-2/3/4/8.

So the primary instrument is a **runtime equivalence test in each consumer's own package**: seed a
`t.TempDir()` fixture carrying every key of `SettingsAxisCleanupKeys()`, assert the fixture is
dirty beforehand (the emptiness guard — without it the test passes vacuously when the fixture stops
seeding), run the consumer, assert every key is gone. That is falsifiable by the defect: a consumer
whose hand-written list omits one key fails on that key by name.

A textual residual-literal scan is retained as a **secondary** check only, and only in a form that
discriminates: it matches `config.Env…` identifiers as well as string literals, and its positive
control must go from hits to no-hits across the change. A control that already produces hits on the
unchanged tree (as the previous AC-002 control did — 22 hits before any edit) is not a control.

### D8 — Restore-before-delete is an ordering constraint on three functions — AND the loop must skip the token pair

All three cleanup functions read `MOAI_BACKUP_AUTH_TOKEN` and write `ANTHROPIC_AUTH_TOKEN` before
deleting; the canonical set contains both keys. Replacing each delete-block with a loop placed
where the block began deletes the backup before anything reads it — in all three at once, silently.

**Ordering is necessary and not sufficient.** `ANTHROPIC_AUTH_TOKEN` is the restore's TARGET and
also a member of the 14-key cleanup view, so a loop over the whole view placed *after* a correct
restore deletes the value the restore just wrote — again in all three functions, and this time while
obeying the ordering rule. The token is destroyed either way; ordering only moves which step does it.

**The resolution (REQ-7 carve-out): the deletion loop skips both token keys
(`ANTHROPIC_AUTH_TOKEN`, `MOAI_BACKUP_AUTH_TOKEN`); the restore branch owns their disposition.**
`ANTHROPIC_AUTH_TOKEN` is set by the restore branch or deleted by its `else`;
`MOAI_BACKUP_AUTH_TOKEN` is consumed by the restore branch's own straight-line delete.

**This preserves today's shape rather than changing it.** The code already looks like this — an
`if`/`else` on the token pair, plus straight-line deletes for the rest
(`internal/hook/session_end.go:852-858`, `internal/cli/launcher.go:356-358`,
`internal/cli/settings.go:182-184`). The trap is the plausible *simplification* that folds the
`if`/`else` and the deletes into one uniform loop over the view. Do not fold it. The landed guards
already encode the same carve-out on the test side (`internal/cli/glm_settings_cleanup_test.go`:
`switch key { case config.EnvAnthropicAuthToken, "MOAI_BACKUP_AUTH_TOKEN": continue }`).

Each of M2/M4/M5 — the three milestones that touch a cleanup function — therefore restates BOTH the
ordering and the carve-out, and names the restore test that catches a violation, rather than leaving
the constraint on one milestone and hoping the others infer it.

## §F Milestones (ordered by decision-reversibility)

Every milestone below that touches a cleanup function restates the restore-before-delete ordering
(D8). That is deliberate repetition, not redundancy: the constraint is violated by a plausible
mechanical implementation, and a constraint stated in one milestone is a constraint the other two
implementers do not read.

### M1 — Declare the canonical settings axis: one declaration, two views (data-model decision)

**File**: `internal/config/envkeys.go`

Add ONE ordered declaration of the 14 settings-axis keys carrying a per-key live/legacy
classification, as a sibling of `GLMEnvVarSet()`, plus the two views REQ-1 names:
`SettingsAxisCleanupKeys()` (all 14) and `SettingsAxisLiveKeys()` (the 9 live keys). Neither view
may be a second literal list.

Doc comment per REQ-1: settings axis; distinct from the tmux axis; five former hand-maintained
lists derive from it. Add an `internal/config` test asserting each view's exact membership and
cardinality AND the set relation `cleanup = live ∪ legacy-tail`, so a future addition is a
deliberate diff.

Verify: `$SCRUB go test -count=1 ./internal/config/...` (AC-001)

### M2 — Route `cleanupGLMSettingsLocal` (C) onto the cleanup view + revise the indicator

**File**: `internal/hook/session_end.go`

The behaviour-visible milestone, so it comes before the mechanical ones. Two changes:

1. The delete list becomes `SettingsAxisCleanupKeys()` (REQ-4). SessionEnd newly deletes
   `API_TIMEOUT_MS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
   `MOAI_STATUSLINE_CONTEXT_SIZE`, `ANTHROPIC_DEFAULT_FABLE_MODEL` — inside the gate only.
   **Ordering + carve-out (REQ-7)**: the `MOAI_BACKUP_AUTH_TOKEN` → `ANTHROPIC_AUTH_TOKEN` restore
   at `session_end.go:852-858` stays ordered BEFORE the deletion loop. A loop placed where the old
   delete-block began destroys the backup before it is read. **And ordering alone is not enough:**
   the loop MUST skip both token keys, because `ANTHROPIC_AUTH_TOKEN` is in the view and a loop over
   the whole view would delete the value the restore just wrote. Keep the existing `if`/`else` on
   the token pair; do not fold it into the loop. Violation caught by
   `TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken` (AC-010) and AC-014.
2. The indicator becomes exactly `{ANTHROPIC_BASE_URL, MOAI_BACKUP_AUTH_TOKEN}` (REQ-5, option 가').
   `MOAI_STATUSLINE_CONTEXT_SIZE` is NOT an indicator. The early return itself stays (REQ-6).

New tests in this milestone, both of which **must call `scrubGatewayEnv(t)` on their first line**.
`cleanupGLMSettingsLocal` returns immediately when `isGatewaySession()` is true
(`internal/hook/gateway_guard.go`, keyed on `MOAI_LAUNCH_PROVIDER`), which every launcher-started
session sets; without the scrub an exclusion assertion passes on the gateway early return rather
than on the indicator logic — green for the wrong reason. Every sibling test in that file already
calls it for this reason.

- `TestCleanupGLMSettingsLocalIndicatorSet` — table-driven, both directions (AC-007).
- `TestCleanupGLMSettingsLocalAdmitsBackupOnlyFileAndRestoresToken` — the net benefit (AC-014).
- The runtime equivalence test for C (AC-002).

Verify: `TestProbeCleanupLeavesLegacyAndOtherAxisKeys` flips to PASS (its fixture seeds
`ANTHROPIC_BASE_URL`, so it is admitted under 가' exactly as before — AC-003);
`TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone` still PASSes, unmodified (AC-009);
`TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken` still PASSes (AC-010).

### M3 — Repair the `CLAUDE_CODE_TEAMMATE_DISPLAY` stand-in fixture (REQ-12)

**File**: `internal/hook/session_end_test.go`

Lands with M2, because M2 is what breaks it. `TestCleanupGLMSettingsLocal` seeds
`CLAUDE_CODE_TEAMMATE_DISPLAY` and asserts it SURVIVES (`wantOtherPresent: true`, the "Check non-GLM
var preservation" block) — but that key is a GLM-owned legacy indicator that `removeGLMEnv` already
deletes and REQ-4 makes SessionEnd delete too.

Replace the **stand-in key**, keep the assertion: retarget the shared preservation assertion to
`CUSTOM_VAR`, and seed `CUSTOM_VAR` **exactly where a stand-in key is seeded today** — i.e. in the
two cases that already carry one.

**Which cases those are, measured in this tree** (`internal/hook/session_end_test.go`, the three
table cases at `:331`, `:349`, `:367`):

| # | Case name | Seeds `CLAUDE_CODE_TEAMMATE_DISPLAY` today | `wantOtherPresent` today | Seed `CUSTOM_VAR`? | `wantOtherPresent` after |
|---|---|---|---|---|---|
| 1 | "GLM active with backup OAuth token…" (`:331`) | yes (`:339`) | `true` (`:347`) | **yes** | `true` |
| 2 | "GLM active without backup OAuth token…" (`:349`) | **no** | `false` (`:364`) | **no** | `false` |
| 3 | "no GLM vars present: file unchanged" (`:367`) | yes (`:369`) | `true` (`:377`) | **yes** | `true` |

Case 2's `false` does **not** mean "the stand-in was deleted" — it means the case never seeded one.
Seeding `CUSTOM_VAR` there would make it survive (no cleanup list contains it), forcing
`wantOtherPresent: false → true` and putting this milestone in direct conflict with AC-015's
control. Narrowing the seeding to cases 1 and 3 keeps the milestone and the control describing the
**same fixture for every case**, with each case's expectation value unchanged.

Leave `CLAUDE_CODE_TEAMMATE_DISPLAY` seeded where it already is (cases 1 and 3), now expected to be
deleted in the gate-admitted case and to survive in the no-indicator case — which turns the repair
into extra coverage rather than lost coverage.

**Do not weaken.** Flipping `wantOtherPresent` to `false`, deleting the assertion, or scoping it to
cases the new behaviour happens to satisfy is the control being removed.

Verify: `$SCRUB go test -count=1 -run TestCleanupGLMSettingsLocal ./internal/hook/ -v` (AC-015)

### M4 — Route `stripGLMCredsAndSetTeammateMode` (B) onto the cleanup view

**File**: `internal/cli/settings.go`

Replace the 13 literals with `SettingsAxisCleanupKeys()`; leave the teammate-mode side effect
alone; do not delete the function (REQ-11).

**Ordering + carve-out (REQ-7)**: the restore at `settings.go:182-184` stays ordered BEFORE the
deletion loop, **and the loop skips both token keys** — `ANTHROPIC_AUTH_TOKEN` is a member of the
view and the restore's target, so a loop over the whole view deletes the restored token even when
the ordering is correct. Keep the existing `if`/`else` on the token pair; do not fold it into the
loop. Violation caught by `TestGLMCleanupRestoresBackedUpAuthToken/stripGLMCredsAndSetTeammateMode`
(AC-010).

Verify: `TestProbeStripGLMCredsLeavesLegacyKeys` and `TestProbeSettingsAxisListsAgree` flip to PASS
(AC-004, AC-005); `TestStripGLMCredsClearsEveryLiveKey` still PASSes — it is the test that actually
asserts the teammate-mode side effect (AC-004);
`TestGLMCleanupRestoresBackedUpAuthToken/stripGLMCredsAndSetTeammateMode` still PASSes (AC-010).
Add the runtime equivalence test for B (AC-002).

### M5 — Route `removeGLMEnv` (A) onto the cleanup view

**File**: `internal/cli/launcher.go`

Replace the 14 literals with `SettingsAxisCleanupKeys()`.

**Ordering + carve-out (REQ-7)**: the restore at `launcher.go:356-358` stays ordered BEFORE the
deletion loop, **and the loop skips both token keys** — `ANTHROPIC_AUTH_TOKEN` is a member of the
view and the restore's target, so a loop over the whole view deletes the restored token even when
the ordering is correct. Keep the existing `if`/`else` on the token pair; do not fold it into the
loop. Violation caught by `TestGLMCleanupRestoresBackedUpAuthToken` (AC-010).

Verify: `TestGLMCleanupRestoresBackedUpAuthToken` still PASSes (AC-010). Add the runtime equivalence
test for A (AC-002).

### M6 — Route both test lists onto the LIVE view; leave the tmux axis alone

**Files**: `internal/cli/glm_settings_cleanup_test.go`,
`internal/hook/glm_settings_cleanup_test.go`

`liveSettingsAxisKeys()` and `liveHookWrittenKeys()` both return `config.SettingsAxisLiveKeys()` —
directly, or via one explicitly named filter where a genuine narrowing is still needed after M2.
Neither is widened to the cleanup view (D3).

Update both helpers' doc comments to name the view they derive from (REQ-8): a comment describing a
literal list that no longer exists is a new trap.

**`buildTmuxClearVars` is not edited, and no key is added to it.**
`TestTmuxClearVarsCoversEveryLiveKeyItOwns` keeps its existing exclusion `switch` unchanged — the
live view it iterates is unchanged, so it neither goes red nor needs weakening.

Verify: `$SCRUB go test -count=1 ./internal/cli/... ./internal/hook/... -timeout 30m`, with
`TestTmuxClearVarsCoversEveryLiveKeyItOwns` explicitly named in the evidence (AC-013).

### M7 — Mutation demonstration (REQ-9)

Deliberately diverge ONE consumer from the canonical view — e.g. drop a single key from
`removeGLMEnv`'s effective set — run the package suite, record the **named** failing test and its
verbatim output, then restore the consumer and re-run to green. Both runs are evidence; the red one
is the load-bearing half.

If no named test turns red, the guard is vacuous and M1-M6 are not complete: add the missing
assertion before proceeding (AC-008).

### M8 — Annotate the deliberately-red probe (REQ-10)

**File**: `internal/hook/glm_context_residue_probe_test.go`

Extend the file's prose comment with the three items REQ-10 enumerates: option (가') chosen (by the
lead, 2026-09-19, revising the earlier three-key 가 the same day); why the stranded case is not
closed — widening the indicator to the context-window pair OR to `MOAI_STATUSLINE_CONTEXT_SIZE`
would delete a non-GLM user's own settings; and that the only route into the stranded state is a
file written by an older binary, the live route having been closed by card t802 (AC-006).

### M9 — Full verification sweep + commit

```bash
$SCRUB go test -count=1 ./internal/config/... ./internal/hook/... ./internal/cli/... -timeout 30m
$SCRUB go test -tags residue_probe -count=1 ./internal/cli/... ./internal/hook/... -timeout 30m
go vet ./internal/config/... ./internal/cli/... ./internal/hook/...
gofmt -l internal/config internal/cli internal/hook
```

`gofmt -l` reports findings on stdout while exiting 0 — judge it by empty output, never by exit
code.

Expected probe state after the sweep: three PASS, `TestProbeStrandedResidueSurvivesIndicatorLoss`
still FAIL with its rationale recorded in the file (AC-012).

## §G Anti-Patterns (avoid)

- Folding `buildTmuxClearVars` in, or citing its length as evidence about the settings axis.
- Adding `CLAUDE_CODE_TEAMMATE_DISPLAY` (or any key) to a tmux list to keep the parity guard green.
- Widening `TestTmuxClearVarsCoversEveryLiveKeyItOwns`'s exclusion `switch` to keep it green.
- Widening `GLMEnvVarSet()` instead of adding a sibling.
- Removing the `ANTHROPIC_BASE_URL` early return "to close the stranded probe".
- Adding the context-window pair — or `MOAI_STATUSLINE_CONTEXT_SIZE` — to the indicator set.
- Placing the canonical-set delete loop before the OAuth restore in any of the three functions.
- Looping over the whole cleanup view *after* a correct restore — the loop reaches
  `ANTHROPIC_AUTH_TOKEN` and destroys the restored token. Skip the token pair (REQ-7 carve-out).
- "Simplifying" the existing `if`/`else` on the token pair plus straight-line deletes into one
  uniform loop over the view. That shape is required, not incidental.
- Seeding `CUSTOM_VAR` into the table case that never carried a stand-in key (case 2), which flips
  its `wantOtherPresent` and trips AC-015's no-weakening control.
- Verifying the collapse with a grep for env-var name strings and calling an empty result a pass.
- Reporting a positive control that already produces hits on the unchanged tree as a control.
- Weakening `TestCleanupGLMSettingsLocal`'s preservation assertion instead of replacing its
  stand-in key.
- Adding a test in `internal/hook` without `scrubGatewayEnv(t)`.
- Deleting `stripGLMCredsAndSetTeammateMode` because it has no caller.
- Declaring the collapse done without the mutation demonstration.
- Leaving the stranded probe red with no prose rationale.
- Asserting the default suite was green before the change without having run and recorded it.
- Measuring "what this card changed" against a **pinned** left edge — a literal SHA *or* a
  `CARD_BASE` resolved once and kept in the shell. Re-resolve `$(git merge-base develop HEAD)` in
  the same invocation as each range check (gitflow-lane-protocol.md §8).
- Running a range predicate **after** the card is merged into `develop`: merge-base becomes the
  card's own tip, the range empties, and the check passes vacuously.
- Placing a positive control at a moment when its condition cannot yet differ — the run-phase-entry
  `"$CARD_BASE"..HEAD` count is necessarily `0`.
- `go test ./...` locally; a bare `unset` in its own Bash call; judging `gofmt -l` by exit code.

## §H Cross-References

- `spec.md` §A (measurements), §A.5 (`injectGLMEnv` per-tree claims), §C (REQ-1..REQ-12),
  §E (constraints, incl. the three deliberate behaviour changes).
- `.moai/reports/t888/plan-audit.md` — iteration 1 verdict (FAIL 0.69); this plan is the
  iteration-2 rework of the verification layer it faulted.
- `acceptance.md` (AC matrix, one command per criterion).
- Card t802 (origin; landed the four probes), card t889 (axis separation), card t888 (this).
- SPEC-CLIFIX-HYGIENE-001 — owns `GLMEnvVarSet()`.
