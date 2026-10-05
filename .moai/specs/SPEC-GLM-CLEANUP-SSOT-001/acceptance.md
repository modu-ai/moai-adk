# Acceptance — SPEC-GLM-CLEANUP-SSOT-001

> Layer note: `spec.md` §C carries the GEARS requirement layer. This file is the **verification
> layer**, so each criterion is written `Given … When … Then …` and is binary-testable by the named
> command beside it. A Given-When-Then scenario is never presented here as a GEARS requirement.
>
> Every command below is prefixed by the environment scrub, in ONE compound invocation — a
> standalone `unset` does not carry into the next command. Written once here as `$SCRUB`:
>
> ```
> $SCRUB = unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER &&
> ```
>
> `internal/cli` runs need `-timeout 30m`. Every run uses `-count=1` — a cached result is not a
> measurement of this tree.
>
> **The range edge is re-derived at every check, never kept.** Each range check below runs
> `git merge-base develop HEAD` as its own plain line immediately before the range command and
> records the printed value in the evidence; the range itself uses the three-dot form
> `git diff … develop...HEAD`, which is by definition the merge-base..HEAD range. The governing
> rule is `.claude/rules/local/gitflow-lane-protocol.md` §8 [HARD]: the left
> edge of a range predicate is re-derived **at the moment of reading**, and the value is **not
> pinned**. Holding one resolved value in a shell variable is pinning exactly as much as writing a
> literal SHA — the hazard is the retention, not the spelling. This repository's lane procedure
> *requires* absorbing `develop` mid-card (`CLAUDE.local.md` §4.1), which moves the merge-base
> forward; an edge resolved before that absorption stays at the old fork point and silently pulls
> **another card's** commits into every range below it. (The earlier compound-assignment form
> `CARD_BASE=$(git merge-base …)` is refused outright by the worktree-isolation guard and has
> been restated throughout this file.)

## §D.0 Instrument discipline (binds every criterion below)

Three rules, each closing a defect the iteration-1 audit found in this file:

1. **An absence observation carries a control.** Every criterion whose expected observation is
   "no output", "no key", "no diff", or "no hit" names, in the same block, a control that WOULD
   produce output if the instrument were broken. An uncontrolled empty result is reported as
   *not measurable*, never as a pass.
2. **A control must discriminate.** A control that already produces its expected signal on the
   unchanged tree separates nothing. Where a control is textual, its required behaviour is stated
   as a transition (hits → no hits, or absent → present), not as a standing fact.
3. **Range checks use the three-dot form `develop...HEAD` with the merge-base recorded on its own
   line in the same block, never a bare `git diff <path>` and never a pinned edge.** A bare `git
   diff` measures the working tree against `HEAD`, so under per-milestone commit discipline it
   returns empty unconditionally — which makes an absence assertion vacuously true the moment the
   change is committed. Existence assertions are made by reading file content, not by reading a
   diff.

   The left edge is re-derived at the moment of the check — `git merge-base develop HEAD` on its
   own line (value recorded in the evidence), then the three-dot range — per
   `.claude/rules/local/gitflow-lane-protocol.md` §8 [HARD] ("읽는 시점에 … 다시 구하고, 값을 핀하지
   않는다"). A `CARD_BASE` resolved once at run-phase entry and carried in the shell is a pinned
   edge; after the procedurally-required `develop` absorption it contaminates every range with other
   cards' commits, turning AC-011's "producer unmodified" red on somebody else's `session_start.go`.

   **Limitation, from the same §8: a range predicate is valid only BEFORE the card is merged.**
   Once the card lands on `develop`, `git merge-base develop HEAD` returns the card's own tip, the
   range is empty, and every absence assertion in this file passes **vacuously**. The non-empty
   controls attached to each range check are what surface that state — a control of `0` after
   integration is "not measurable", not a pass. Post-merge, evidence is made by tree-identity (the
   merge commit's tree equals the re-measured tree), never by these ranges.

## §D AC Matrix

| AC ID | REQ | Severity | Summary |
|---|---|---|---|
| AC-001 | REQ-1 | MUST-PASS | One canonical declaration + two derived views exist in `internal/config/envkeys.go`, documented |
| AC-002 | REQ-1, REQ-8 | MUST-PASS | Runtime equivalence: each production consumer deletes exactly the cleanup view (fixture proven dirty first; `MOAI_BACKUP_AUTH_TOKEN` absent by design — restore path owned by AC-010) |
| AC-003 | REQ-4 | MUST-PASS | `TestProbeCleanupLeavesLegacyAndOtherAxisKeys` flips FAIL → PASS |
| AC-004 | REQ-3 | MUST-PASS | `TestProbeStripGLMCredsLeavesLegacyKeys` flips FAIL → PASS **and** the teammate-mode side effect is asserted |
| AC-005 | REQ-2, REQ-3 | MUST-PASS | `TestProbeSettingsAxisListsAgree` flips FAIL → PASS |
| AC-006 | REQ-10 | MUST-PASS | `TestProbeStrandedResidueSurvivesIndicatorLoss` still FAILS **and** its rationale is in the file |
| AC-007 | REQ-5 | MUST-PASS | Indicator set = exactly two keys; `MOAI_STATUSLINE_CONTEXT_SIZE` and the context-window pair all excluded |
| AC-008 | REQ-9 | MUST-PASS | Mutation demonstration: a named test turns red on divergence, green on restore |
| AC-009 | REQ-6 | MUST-PASS | Negative control `TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone` passes, unmodified across the card |
| AC-010 | REQ-7 | MUST-PASS | Restore-before-delete preserved in **all three** cleanup functions |
| AC-011 | REQ-11 | MUST-PASS | Producer unmodified across the card; `stripGLMCredsAndSetTeammateMode` still present |
| AC-012 | §E constraints | MUST-PASS | Whole-scope verification green: three packages, vet clean, gofmt empty |
| AC-013 | REQ-8 | MUST-PASS | Both test lists derive from the LIVE view; the out-of-scope tmux-parity guard stays green, unmodified |
| AC-014 | REQ-5 | MUST-PASS | Net benefit: a backup-only file is newly admitted and the user's OAuth token is restored |
| AC-015 | REQ-12 | MUST-PASS | The user-owned-key assertion survives, asserted on a key no cleanup list contains |
| AC-016 | REQ-2, REQ-3, REQ-4 | MUST-PASS | Structural: each production consumer reads the canonical view exactly once, spelling no key itself |

Sixteen criteria — the Tier M ceiling. Adding a seventeenth is a signal to split the SPEC, not to
relax the budget.

## §D.1 Severity and traceability

All sixteen are MUST-PASS. This is a Tier M behaviour-affecting change in a path that mutates a
user's `settings.local.json`; there is no nice-to-have tier here.

Each AC names **one or more** REQs or constraints from `spec.md` — several criteria legitimately
verify a property two requirements share, and the matrix column records every one of them.

Traceability, both directions:

| REQ | Covered by |
|---|---|
| REQ-1 | AC-001, AC-002 |
| REQ-2 | AC-005, AC-016 |
| REQ-3 | AC-004, AC-005, AC-016 |
| REQ-4 | AC-003, AC-016 |
| REQ-5 | AC-007, AC-014 |
| REQ-6 | AC-009 |
| REQ-7 | AC-010 |
| REQ-8 | AC-002, AC-013 |
| REQ-9 | AC-008 |
| REQ-10 | AC-006 |
| REQ-11 | AC-011 |
| REQ-12 | AC-015 |

No orphan REQ (every requirement has at least one criterion) and no orphan AC (every criterion's
REQ column resolves to a requirement that exists, or to the §E constraints block). The
"no hand-written key literals" clause shared by REQ-2/3/4/8 is carried by AC-002 (runtime) and
AC-016 (structural) — two instruments, not one, because the iteration-1 version rested the whole
clause on a single check that could not see the defect.

## §D.2 Given-When-Then scenarios

### AC-001 — Canonical declaration and its two views, documented

```
GIVEN internal/config/envkeys.go
WHEN the canonical settings-axis declaration is read
THEN SettingsAxisCleanupKeys() returns exactly the 14 keys of spec.md §A.1, in a stable order
  AND SettingsAxisLiveKeys() returns exactly the 9 live keys of spec.md REQ-1
  AND the cleanup view equals the live view plus the 5-key legacy tail
  AND both views derive from ONE declaration — neither is a second literal list
  AND the doc comment states that it is the settings axis, that it is distinct from the tmux
      axis, and that five former hand-maintained lists derive from it
  AND config.GLMEnvVarSet() is unchanged (still its original 3-key set)
```

**Command**
```
$SCRUB go test -count=1 ./internal/config/...
```
**Expected**: PASS, exit 0 — including the M1 membership / cardinality / set-relation test.

**Second check — the doc comment is prose, so it is READ, not inferred.** It is read from the file,
not from a diff (§D.0 rule 3):
```
grep -ci 'settings axis' internal/config/envkeys.go
grep -ci 'tmux axis' internal/config/envkeys.go
grep -ci 'five former hand-maintained lists' internal/config/envkeys.go
```
**Expected**: each ≥ 1.
**Control**: the same three patterns against `internal/config/envkeys.go` at the card base, with the
base re-resolved in the same invocation —
`git merge-base develop HEAD` (record as CARD_BASE) then `git show <CARD_BASE>:internal/config/envkeys.go | grep -ci 'settings axis'`
etc. — MUST return
`0` for each. That absent → present transition is what distinguishes "the comment was added" from
"the pattern was always matching something".

---

### AC-002 — Runtime equivalence: each consumer deletes exactly the cleanup view

This is the **primary** instrument for the five→one collapse. A textual scan is not, and cannot be:
measured on the unchanged tree, the five lists spell most keys as Go identifiers
(`config.EnvAnthropicDefaultFableModel`), so a consumer retaining its entire hand-written list —
the exact forbidden state — produces zero hits.

```
GIVEN a t.TempDir() fixture whose env carries EVERY key of config.SettingsAxisCleanupKeys()
      EXCEPT MOAI_BACKUP_AUTH_TOKEN, which is absent by design
  AND the fixture is asserted dirty before the consumer runs (emptiness guard), that guard
      skipping the same one key
WHEN each of the three production cleanup functions runs against it
THEN every key of config.SettingsAxisCleanupKeys() is absent afterwards
  AND the assertion iterates config.SettingsAxisCleanupKeys() rather than a list written here
```

**Why `MOAI_BACKUP_AUTH_TOKEN` is excluded from the fixture — this is the convention already landed
in this tree, not a new exception.** Its presence is what tells the cleanup paths to **restore** the
backed-up value as `ANTHROPIC_AUTH_TOKEN`, so a fixture seeding it with a value makes
`ANTHROPIC_AUTH_TOKEN` legitimately SURVIVE — and `ANTHROPIC_AUTH_TOKEN` is itself a member of the
cleanup view, so a whole-view "every key is absent" traversal would then be false against a
*correct* implementation. Omitting the backup key makes the restore branch's `else` fire,
`ANTHROPIC_AUTH_TOKEN` is deleted, and the whole-view traversal holds.

The landed guards do exactly this, verbatim:

- `internal/cli/glm_settings_cleanup_test.go:38` — `glmLiveDirtyEnv()` skips the key with
  `continue`, under the doc comment "MOAI_BACKUP_AUTH_TOKEN is deliberately left out: its presence
  tells the cleanup paths to RESTORE that value as ANTHROPIC_AUTH_TOKEN, which is the documented
  OAuth-preservation behaviour, not residue. The restore path has its own case below."
- `internal/cli/glm_settings_cleanup_test.go:63` — `assertFixtureIsDirty()` skips the same key with
  `continue // absent by design — see glmLiveDirtyEnv`.
- `internal/hook/glm_settings_cleanup_test.go:81` — the hook-side fixture carries the same skip and
  the same comment.
- `internal/hook/glm_settings_cleanup_test.go:141` — the separate named restore case asserts the
  survival this fixture deliberately avoids: `if env[config.EnvAnthropicAuthToken] !=
  "user-oauth-token" { t.Errorf("backed-up token must be restored, …") }`. That guard passes today.

The new tests follow that shape. **Restore-path coverage is not lost**: it is owned by AC-010 (and
AC-014 for the backup-only file), exactly as the landed guards split it into a separate named case.

The three tests live in their consumers' own packages: two in `internal/cli` (for `removeGLMEnv`
and `stripGLMCredsAndSetTeammateMode`) and one in `internal/hook` (for `cleanupGLMSettingsLocal`,
seeded so the indicator gate admits it — through `ANTHROPIC_BASE_URL`, since the backup key is
absent — and calling `scrubGatewayEnv(t)` first).

**Commands**
```
$SCRUB go test -count=1 -run 'CleanupViewEquivalence' ./internal/cli/ -v -timeout 30m
$SCRUB go test -count=1 -run 'CleanupViewEquivalence' ./internal/hook/ -v
```
**Expected**: PASS, exit 0, with a non-zero count of subtests run in each package.

**Control — a filtered run that selects nothing also prints `ok`.** The `-v` output MUST name each
expected test; a run reporting `no tests to run`, or `ok` with no `=== RUN` lines, is *not
measurable* and is reported as a gap, never as a pass.

**Emptiness guard — stated as a property of the test, not of this document.** Each test asserts the
fixture is dirty BEFORE running the consumer, skipping `MOAI_BACKUP_AUTH_TOKEN` exactly as
`assertFixtureIsDirty` does today (`internal/cli/glm_settings_cleanup_test.go:63`). Without the
guard, a fixture that silently stopped seeding would make every "key is absent" assertion pass
vacuously. The guard's own presence is checkable:
```
grep -c 'assertFixtureIsDirty\|fixture is dirty' internal/cli/glm_settings_cleanup_test.go internal/hook/glm_settings_cleanup_test.go
```
**Expected**: ≥ 1 in each file.

**Secondary (textual residual-literal scan).** Retained only in a discriminating form — matching Go
identifiers as well as string literals:
```
sed -n '/func removeGLMEnv/,/^}/p' internal/cli/launcher.go | grep -cE 'config\.Env[A-Za-z]+|"ANTHROPIC_|"CLAUDE_CODE_|"API_TIMEOUT_MS"|"MOAI_'
```
…and the same extraction for `stripGLMCredsAndSetTeammateMode` (`internal/cli/settings.go`) and
`cleanupGLMSettingsLocal` (`internal/hook/session_end.go`).
**Expected after the change**: a small number per function — the canonical-view reference plus the
`MOAI_BACKUP_AUTH_TOKEN` / `ANTHROPIC_AUTH_TOKEN` pair the restore path legitimately names by hand
(REQ-7). Each remaining hit is quoted in the evidence with its role named; none may be a member of
the cleanup view other than those two restore keys.
**Control (discriminating)**: the same three extractions at the card base, base re-resolved in the
same invocation
(`git merge-base develop HEAD` recorded as CARD_BASE, then `git show <CARD_BASE>:internal/cli/launcher.go | sed -n … | grep -c …`)
MUST return a
substantially higher count — today each function names 8-12 keys itself. The required signal is the
**transition** from many to few; a standing count proves nothing.

This secondary check never overrides the runtime result. Where the two disagree, the runtime tests
are the verdict and the discrepancy is reported.

---

### AC-003 — Hook-side legacy/other-axis residue closed

```
GIVEN a settings.local.json in t.TempDir() carrying ANTHROPIC_BASE_URL plus
      ANTHROPIC_DEFAULT_FABLE_MODEL, MOAI_STATUSLINE_CONTEXT_SIZE and API_TIMEOUT_MS
WHEN cleanupGLMSettingsLocal runs
THEN none of those three keys remains in the file's env object
```

**Command**
```
$SCRUB go test -tags residue_probe -count=1 -run TestProbeCleanupLeavesLegacyAndOtherAxisKeys ./internal/hook/
```
**Expected**: `PASS`, exit 0.
**Baseline**: FAIL at `0cca34439` — `.moai/state/verify/t888/baseline-probe-hook.txt`, `EXIT:1`.

**Option (가') regression note, proven by running not by reasoning.** Narrowing the indicator set
from three keys to two does not re-open this probe: its fixture seeds `ANTHROPIC_BASE_URL`, so the
gate admits the file through that key and `MOAI_STATUSLINE_CONTEXT_SIZE`'s demotion from indicator
to merely-cleaned is irrelevant to admission. **That reasoning is not the evidence.** The evidence
is this command's PASS after M2 lands; if it instead FAILs, the (가') encoding is wrong and the run
phase stops rather than adjusting the probe.

---

### AC-004 — `stripGLMCredsAndSetTeammateMode` legacy key closed, side effect asserted

```
GIVEN an env map carrying ANTHROPIC_DEFAULT_FABLE_MODEL
WHEN stripGLMCredsAndSetTeammateMode runs
THEN that key is absent afterwards
  AND the teammate-mode side effect is unchanged
```

**Commands** — two, because the probe does not assert the second conjunct. The probe
(`internal/cli/glm_context_residue_probe_test.go`) checks legacy-key absence only; the
teammate-mode assertion lives in `TestStripGLMCredsClearsEveryLiveKey`, which asserts on
`got.TeammateMode`.
```
$SCRUB go test -tags residue_probe -count=1 -run TestProbeStripGLMCredsLeavesLegacyKeys ./internal/cli/ -timeout 30m
$SCRUB go test -count=1 -run TestStripGLMCredsClearsEveryLiveKey ./internal/cli/ -v -timeout 30m
```
**Expected**: both PASS, exit 0.
**Baseline (first command)**: FAIL — `.moai/state/verify/t888/baseline-probe-cli.txt`, `EXIT:1`.
**Control**: the `-v` output of the second command MUST show a `=== RUN` line for the named test; a
selector matching nothing prints `ok` and would otherwise read as a pass.

---

### AC-005 — The two CLI-side lists agree

```
GIVEN removeGLMEnv and stripGLMCredsAndSetTeammateMode
WHEN their effective delete sets are compared
THEN the two sets are equal
```

**Command**
```
$SCRUB go test -tags residue_probe -count=1 -run TestProbeSettingsAxisListsAgree ./internal/cli/ -timeout 30m
```
**Expected**: `PASS`, exit 0.
**Baseline**: FAIL — B kept `FABLE_MODEL` that A cleared
(`.moai/state/verify/t888/baseline-probe-cli.txt`, `EXIT:1`).

---

### AC-006 — The stranded probe stays red, with its rationale recorded

```
GIVEN a settings file carrying CLAUDE_CODE_MAX_CONTEXT_TOKENS and no indicator key
WHEN cleanupGLMSettingsLocal runs
THEN the key survives (the gate declines the file)
  AND internal/hook/glm_context_residue_probe_test.go records in prose:
      (a) that option (가') — conservative widening to exactly
          {ANTHROPIC_BASE_URL, MOAI_BACKUP_AUTH_TOKEN} — was chosen, by the lead, on 2026-09-19,
          revising the earlier three-key option (가) the same day;
      (b) that the stranded case is deliberately not closed because widening the indicator to the
          context-window pair — or to MOAI_STATUSLINE_CONTEXT_SIZE, a documented explicit user
          override — would delete a non-GLM user's own settings;
      (c) that the only route into the stranded state is a file written by an older binary,
          the live route having been closed by card t802
```

**Command (failure is the expected observation)**
```
$SCRUB go test -tags residue_probe -count=1 -run TestProbeStrandedResidueSurvivesIndicatorLoss ./internal/hook/
```
**Expected**: `FAIL` — the accepted outcome, cited as such in the run-phase report so a later
reader does not read it as an unfinished milestone.
**Control**: the failure message MUST name `CLAUDE_CODE_MAX_CONTEXT_TOKENS`. A FAIL for any other
reason (compile error, fixture panic) is a different failure and does not satisfy this AC.

**Rationale-presence check** — content search, not a fixed line window. M8 lengthens the header
comment, so a `sed -n '1,60p'` window would push items out of view and read them as absent:
```
grep -c "가'" internal/hook/glm_context_residue_probe_test.go
grep -c 'MOAI_STATUSLINE_CONTEXT_SIZE' internal/hook/glm_context_residue_probe_test.go
grep -c 't802' internal/hook/glm_context_residue_probe_test.go
```
**Expected**: each ≥ 1, and the three items are then read in prose to confirm each says what (a),
(b), (c) require. A grep hit is the locator; the reading is the verdict. A missing item fails this
AC even when the test behaves as expected.

---

### AC-007 — Indicator set is exactly the two MoAI-exclusive keys

```
GIVEN cleanupGLMSettingsLocal's indicator check
WHEN a settings file carries ONLY ANTHROPIC_BASE_URL
  OR ONLY MOAI_BACKUP_AUTH_TOKEN
THEN the file is admitted and cleaned
WHEN a settings file carries ONLY MOAI_STATUSLINE_CONTEXT_SIZE
  OR ONLY CLAUDE_CODE_AUTO_COMPACT_WINDOW
  OR ONLY CLAUDE_CODE_MAX_CONTEXT_TOKENS
THEN the file is NOT admitted and is left untouched
```

**Command** (the M2 table-driven test, both directions in one table)
```
$SCRUB go test -count=1 -run 'TestCleanupGLMSettingsLocalIndicatorSet' ./internal/hook/ -v
```
**Expected**: PASS with all five cases asserted — two admitted, three excluded. A test asserting
only the admitted direction does not satisfy this AC; the exclusion is the half that protects a
non-GLM user, and `MOAI_STATUSLINE_CONTEXT_SIZE` is the case this SPEC's revision exists for.

**Control 1 — the exclusion half must not pass for the wrong reason.**
```
grep -A2 'func TestCleanupGLMSettingsLocalIndicatorSet' internal/hook/session_end_test.go internal/hook/glm_settings_cleanup_test.go | grep -c 'scrubGatewayEnv'
```
**Expected**: `1`. Without the scrub, `isGatewaySession()` makes `cleanupGLMSettingsLocal` return
immediately under any launcher-started session, so "NOT admitted, left untouched" would pass on the
gateway early return rather than on the indicator logic.

**Control 2 — the run must have swept the cases.** The `-v` output MUST list a subtest line per
table case; a run reporting fewer than five is *not measurable*.

---

### AC-008 — Mutation demonstration: the guard turns red

```
GIVEN the five lists folded onto the canonical declaration and the suite green
WHEN one consumer is deliberately diverged from the canonical view
     (for example: one key dropped from removeGLMEnv's effective set)
THEN a NAMED test fails, and its name and verbatim output are recorded
WHEN the consumer is restored
THEN that same test passes again
```

**Runnable recipe**
```
# 1. green baseline
$SCRUB go test -count=1 ./internal/cli/... ./internal/hook/... -timeout 30m

# 2. introduce the divergence by hand in internal/cli/launcher.go:
#    replace the canonical-view iteration with an explicit slice omitting ONE key
#    (e.g. ANTHROPIC_DEFAULT_FABLE_MODEL).

# 3. observe RED — capture the failing test NAME and its verbatim output
$SCRUB go test -count=1 ./internal/cli/... -timeout 30m > .moai/state/verify/t888/mutation-red.txt 2>&1; echo "EXIT:$?"
tail -30 .moai/state/verify/t888/mutation-red.txt

# 4. revert step 2 (git checkout -- internal/cli/launcher.go), then re-run
$SCRUB go test -count=1 ./internal/cli/... ./internal/hook/... -timeout 30m
```
**Expected**: step 3 exits non-zero and shows `--- FAIL:` with a named test; step 4 returns to `ok`
for every package. Both the red output and the restored green are cited in the run-phase evidence.

**Failure condition**: if step 3 produces no named failure, the guard is vacuous — M1-M6 are
incomplete and an assertion is missing. Do not proceed.

---

### AC-009 — The negative control is preserved, unmodified

```
GIVEN a settings.local.json that was never in GLM mode, carrying the user's own
      ANTHROPIC_AUTH_TOKEN and no indicator key
WHEN cleanupGLMSettingsLocal runs
THEN the file is left alone and the user's own token survives
```

**Command**
```
$SCRUB go test -count=1 -run TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone ./internal/hook/ -v
```
**Expected**: `PASS`, with a `=== RUN` line naming the test.
**Baseline**: PASSing at `0cca34439` — `.moai/state/verify/t888/baseline-guards-hook.txt`, `EXIT:0`.

**Unmodified check — over the card's whole range, not the working tree.** A bare
`git diff -- <path>` returns empty the moment the milestone is committed, so it would pass even if
the control had been gutted and committed:
```
git merge-base develop HEAD
git diff develop...HEAD -- internal/hook/glm_settings_cleanup_test.go
```
**Expected**: no hunk touching `TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone`. Hunks elsewhere
in the file are expected (M6 routes `liveHookWrittenKeys()` and AC-002 adds a test there) and are
quoted in the evidence with their milestone named.
**Control**: `git merge-base develop HEAD` recorded, then `git diff --name-only develop...HEAD | wc -l`
MUST be ≥ 1. A control of `0` means the range is empty and the check measured nothing — report it as
a gap, not a pass. (At run-phase *entry* this control is necessarily `0`, which is why it lives here
at the use site and not in `plan.md` §C pre-flight.)

A change to the control that makes it agree with the new behaviour is not evidence — it is the
control being removed.

---

### AC-010 — Restore-before-delete preserved in all three functions

```
GIVEN a settings file (or env map) carrying a backed-up OAuth token in MOAI_BACKUP_AUTH_TOKEN
WHEN removeGLMEnv, stripGLMCredsAndSetTeammateMode, or cleanupGLMSettingsLocal runs
THEN ANTHROPIC_AUTH_TOKEN holds the restored OAuth token
  AND MOAI_BACKUP_AUTH_TOKEN is gone
```

**Commands** — one per side, because the canonical view contains BOTH keys and a delete loop placed
ahead of the restore destroys the backup in all three functions at once:
```
$SCRUB go test -count=1 -run TestGLMCleanupRestoresBackedUpAuthToken ./internal/cli/ -v -timeout 30m
$SCRUB go test -count=1 -run TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken ./internal/hook/ -v
```
**Expected**: both PASS. The first covers `removeGLMEnv` and the
`stripGLMCredsAndSetTeammateMode` subtest; the second covers `cleanupGLMSettingsLocal`.
**Control**: each `-v` output MUST name its test (and, for the first, its
`stripGLMCredsAndSetTeammateMode` subtest). A selector matching nothing prints `ok`.

---

### AC-011 — Producer untouched across the card; dead function retained

```
GIVEN the card's whole range of commits
WHEN internal/hook/session_start.go and internal/cli/settings.go are inspected
THEN ensureGLMCredentials (and maybeSet1MAutoCompactWindow / maybeDeclareGLMContextWindow)
     carry no change anywhere in the range
  AND stripGLMCredsAndSetTeammateMode still exists
```

**Commands**
```
git merge-base develop HEAD
git diff --stat develop...HEAD -- internal/hook/session_start.go
grep -n 'func stripGLMCredsAndSetTeammateMode' internal/cli/settings.go
```
**Expected**: the first prints nothing; the second prints one match.

**Control for the first (an absence observation).**
```
git merge-base develop HEAD
git diff --name-only develop...HEAD | wc -l
```
**Expected**: ≥ 1, and the listed paths MUST include the files this card does change
(`internal/config/envkeys.go`, `internal/cli/launcher.go`, `internal/cli/settings.go`,
`internal/hook/session_end.go`). An empty or 1-line list means the range itself is wrong and the
producer check measured nothing — report as a gap. The working-tree form
(`git diff -- internal/hook/session_start.go`) is explicitly NOT used: it returns empty even when
the producer was modified and committed.

---

### AC-012 — Whole-scope verification green

> **Observation method amended 2026-09-19 (lead-approved). The criterion's guarantee is
> unchanged; only how it is observed for `internal/cli` changed.**
>
> **(1) Original wording, quoted verbatim:** *"WHEN the three affected packages are verified in a
> scrubbed environment / THEN **every package reports `ok`**, vet is silent, and gofmt lists no
> file"*, with the command `$SCRUB go test -count=1 ./internal/config/... ./internal/hook/...
> ./internal/cli/... -timeout 30m`.
>
> **(2) Corrected observation method:** `internal/config` and `internal/hook` are verified by a
> full-package run as before. **`internal/cli` is verified as "all 17 subpackages `ok` **plus** the
> delta-targeted set `ok`"**; the root package's full run is recorded as an explicit Gap.
>
> **(3) Why it changed — the original is unsatisfiable by any correct implementation.** The
> `internal/cli` ROOT package does not complete within the run budget. Measured twice, both cut at
> the boundary with zero output rows: first a background run exiting 144, then an outer
> `timeout 1500` cutting at 25 minutes with `EXIT:124`. The machine was **not** the cause — load at
> that measurement was `7.68` 1-min with **81% idle**. A cut run is *unmeasured*, not FAIL, and an
> unmeasured run cannot serve as a comparison point in either direction: re-running the root after
> the change would be cut identically, so neither before nor after yields a value. This is the same
> class as N1 and N5 — a criterion no correct work can satisfy — differing only in that the cause
> here is environmental rather than textual. The defect is in the criterion's wording, not in the
> implementation, so the wording is what changes. Root-cause diagnosis (which test consumes the
> budget) is card **t962**, cross-referenced in the Gap below; a useful coordinate for it: in this
> same package the delta-targeted set completes in **9.5 seconds** while the root full run exceeds
> 25 minutes.

**Commands**
```
$SCRUB go test -count=1 ./internal/config/... ./internal/hook/... -timeout 30m
$SCRUB go test -count=1 ./internal/cli/... -timeout 30m   # 17 subpackages; root is listed separately below
$SCRUB go test -count=1 -run GLM   ./internal/cli/ -timeout 30m
$SCRUB go test -count=1 -run Tmux  ./internal/cli/ -timeout 30m
$SCRUB go test -count=1 -run OAuth ./internal/cli/ -timeout 30m
go vet ./internal/config/... ./internal/cli/... ./internal/hook/...
gofmt -l internal/config internal/cli internal/hook
```
**Expected**: every command `ok`; vet silent; `gofmt -l` prints **nothing** — judged by empty
output, never by exit code, because `gofmt -l` reports findings while exiting 0.

**What the 17 subpackages do and do not establish.** Measured: **0** of the 17 reference any symbol
this card changes (`removeGLMEnv`, `stripGLMCredsAndSetTeammateMode`, `liveSettingsAxisKeys`,
`buildTmuxClearVars`), while **12** root test files do. Their green therefore carries **no
behavioural attribution** for this change and MUST NOT be cited as evidence that the change is
safe. What it does carry is compile/import-level assurance: 5 of the 17 import `internal/config`,
which M1 modifies. State it that way in the run-phase report — "17 subpackages pass" read as a
safety argument is exactly the substitution this note exists to prevent.

**Control**: compared against the §C.3 pre-flight baselines recorded at entry — `internal/config`
(`baseline-config.txt`), `internal/hook` (measured after the window holder's run), and the
`internal/cli` delta-targeted set (`baseline-cli-delta.txt`, all three `EXIT:0` at load 8.45 /
59.61% idle). Where a baseline was not recorded, this AC is *not measurable* for any package that
is red — a package red both before and after is a pre-existing condition, and without the baseline
the two cases cannot be told apart.

**Gap (explicit, not a silent omission)**: the `internal/cli` ROOT package full run is NOT observed
by this card. Evidence for why: `.moai/state/verify/t888/baseline-default-suite.txt` (this lane's
aborted attempt) plus the window holder's two boundary-cut runs. Owning card for the diagnosis:
**t962**.

---

### AC-013 — Both test lists derive from the LIVE view; the tmux axis is untouched

```
GIVEN liveSettingsAxisKeys() and liveHookWrittenKeys()
WHEN each is read after the change
THEN each returns config.SettingsAxisLiveKeys(), directly or via one explicitly named filter
  AND neither is a second literal list
  AND neither is widened to the 14-key cleanup view
  AND each helper's doc comment names the view it derives from
  AND TestTmuxClearVarsCoversEveryLiveKeyItOwns still passes with its exclusion switch unchanged
  AND buildTmuxClearVars carries no newly-added key
```

**Commands**
```
$SCRUB go test -count=1 -run TestTmuxClearVarsCoversEveryLiveKeyItOwns ./internal/cli/ -v -timeout 30m
git merge-base develop HEAD
git diff develop...HEAD -- internal/cli/glm.go
sed -n '/func liveSettingsAxisKeys/,/^}/p' internal/cli/glm_settings_cleanup_test.go
sed -n '/func liveHookWrittenKeys/,/^}/p' internal/hook/glm_settings_cleanup_test.go
```
**Expected**: the test PASSes with a `=== RUN` line; the `git diff` shows **no hunk inside
`buildTmuxClearVars`**; both helper bodies reference `config.SettingsAxisLiveKeys()` and spell no
key of their own.

**Control for the `git diff` absence**: the same `$CARD_BASE` range control as AC-011 — the range
MUST be non-empty, or the tmux-untouched claim measured nothing.

**Why this AC exists**: routing these two helpers onto the cleanup view instead would pull
`CLAUDE_CODE_TEAMMATE_DISPLAY` into the tmux-parity loop, where it is measurably absent from
`buildTmuxClearVars()` — turning that guard red and putting the change in direct conflict with
AC-012. This criterion pins the resolution rather than leaving it to be rediscovered.

---

### AC-014 — Net benefit: a backup-only file is newly admitted and its token restored

```
GIVEN a settings.local.json in t.TempDir() carrying MOAI_BACKUP_AUTH_TOKEN with the user's
      OAuth token and ANTHROPIC_AUTH_TOKEN holding a GLM key, and NO ANTHROPIC_BASE_URL
WHEN cleanupGLMSettingsLocal runs
THEN the file is admitted (the indicator gate does not decline it)
  AND ANTHROPIC_AUTH_TOKEN holds the user's restored OAuth token
  AND MOAI_BACKUP_AUTH_TOKEN is gone
```

This is the widening's net benefit, stated as an observation rather than as an argument: before
this change that file was declined by the gate and the user's backed-up token stayed stranded.

**Command** (a new named test added in M2, calling `scrubGatewayEnv(t)` first)
```
$SCRUB go test -count=1 -run TestCleanupGLMSettingsLocalAdmitsBackupOnlyFileAndRestoresToken ./internal/hook/ -v
```
**Expected**: PASS, with a `=== RUN` line naming the test.

**RED-now cell — the test must be shown red before M2.** Written against the unchanged tree, this
test fails: the gate keyed on `ANTHROPIC_BASE_URL` alone declines the file, so the GLM key survives
as `ANTHROPIC_AUTH_TOKEN` and the backup is never read. That pre-implementation FAIL is captured
verbatim and cited; a test first observed green proves nothing about what the change accomplished.

---

### AC-015 — The user-owned-key assertion survives, on a genuinely user-owned key

```
GIVEN TestCleanupGLMSettingsLocal in internal/hook/session_end_test.go
WHEN the table is read after the change
THEN its "non-GLM var preservation" assertion still asserts that a seeded key SURVIVES cleanup
  AND the key it asserts on appears in NO cleanup list (CUSTOM_VAR)
  AND the assertion's wantOtherPresent expectation is not flipped, deleted, or narrowed
```

**Commands**
```
$SCRUB go test -count=1 -run TestCleanupGLMSettingsLocal ./internal/hook/ -v
grep -n 'CUSTOM_VAR' internal/hook/session_end_test.go
$SCRUB go test -count=1 -run 'CleanupViewMembership' ./internal/config/ -v
```
**Expected**: the table test PASSes with every case listed; `CUSTOM_VAR` appears in the fixture and
in the preservation assertion; the config membership test confirms `CUSTOM_VAR` is not a member of
either view (asserted there, not assumed here).

**Intent-preservation control — stated as NO WEAKENING, which is what REQ-12 requires.** The range
diff MUST show the assertion still present and no case's expectation weakened:
```
git merge-base develop HEAD
git diff develop...HEAD -- internal/hook/session_end_test.go | grep -E '^[+-].*wantOtherPresent'
```
**Expected**, all three conditions:

1. No case's expectation moves `true → false`. (`false → true` is a strengthening, not a weakening,
   and is not rejected by this control — REQ-12's wording is "shall NOT be **weakened**".)
2. No `wantOtherPresent` line is deleted without a replacement for that case, and the field itself
   is not removed from the struct.
3. The assertion's scope is not narrowed — no case is dropped from the table and no `t.Skip` is
   introduced.

A `-  wantOtherPresent: true` with no `+` line for that case is the control being removed, and
fails this AC regardless of whether the suite is green.

**Fixture agreement with M3 (reconciled).** M3 seeds `CUSTOM_VAR` only where a stand-in key is
seeded today — cases 1 (`:331`) and 3 (`:367`) — and **not** in case 2 (`:349`), which carries no
stand-in and whose `wantOtherPresent: false` (`:364`) therefore means "never seeded", not "deleted".
Under that seeding, every case's expectation value is in fact unchanged (`true` / `false` / `true`),
so this control and M3 describe the same fixture for every case. The condition is nonetheless
written as no-weakening rather than as value-equality, because value-equality would also reject a
legitimate future strengthening.

---

### AC-016 — Structural: each consumer reads the canonical view exactly once

```
GIVEN the bodies of removeGLMEnv, stripGLMCredsAndSetTeammateMode and cleanupGLMSettingsLocal
WHEN each body is extracted and read after the change
THEN each references config.SettingsAxisCleanupKeys() exactly once
  AND each deletes its keys by iterating that reference
  AND none enumerates settings-axis keys itself, except the MOAI_BACKUP_AUTH_TOKEN /
      ANTHROPIC_AUTH_TOKEN pair the REQ-7 restore path legitimately names
```

**Commands**
```
sed -n '/func removeGLMEnv/,/^}/p' internal/cli/launcher.go | grep -c 'SettingsAxisCleanupKeys()'
sed -n '/func stripGLMCredsAndSetTeammateMode/,/^}/p' internal/cli/settings.go | grep -c 'SettingsAxisCleanupKeys()'
sed -n '/func cleanupGLMSettingsLocal/,/^}/p' internal/hook/session_end.go | grep -c 'SettingsAxisCleanupKeys()'
```
**Expected**: `1` from each.

**Control — the extraction itself can silently miss.** `sed` range extraction loses its scope if a
function is split, renamed, or inlined, and an extraction that captured nothing yields `0`, which
reads like a violation rather than like a broken instrument. So each extraction is first shown to
have captured the right body:
```
sed -n '/func removeGLMEnv/,/^}/p' internal/cli/launcher.go | wc -l
```
**Expected**: a plausible non-zero body length for each of the three. A length of `0` or `1` means
the extraction failed and the result is *not measurable*, not a failure of the code.

This criterion is the structural half of the "no hand-written literals" clause; AC-002 is the
runtime half. Neither alone was enough: the iteration-1 version rested the clause on a single
textual check that could not see the defect, and a mutant keeping all five lists while manually
agreeing with the canonical set satisfied every other criterion.

## §D.3 Probe state after this SPEC (the closure table)

| Probe | Before | After | Status |
|---|---|---|---|
| `TestProbeCleanupLeavesLegacyAndOtherAxisKeys` | FAIL | PASS | closed (AC-003) |
| `TestProbeStripGLMCredsLeavesLegacyKeys` | FAIL | PASS | closed (AC-004) |
| `TestProbeSettingsAxisListsAgree` | FAIL | PASS | closed (AC-005) |
| `TestProbeStrandedResidueSurvivesIndicatorLoss` | FAIL | FAIL | **accepted non-closure**, rationale recorded in the file (AC-006) |

The three "Before" FAILs are cited, not re-derived: `.moai/state/verify/t888/baseline-probe-hook.txt`
and `baseline-probe-cli.txt`, both `EXIT:1`, measured by the lane at base `0cca34439`.

## §D.4 Definition of Done

- All sixteen ACs observed, each with its command's verbatim output quoted in the run-phase
  evidence, and each absence observation accompanied by its control's output (§D.0 rule 1).
- The §C.3 default-suite pre-flight baseline was actually RUN and recorded before the first edit.
  It is a step, not a premise: an unrecorded baseline makes AC-012 unresolvable for any red package.
- The mutation demonstration's RED output is cited by test name (AC-008) — the card is not done
  without it.
- AC-014's RED-now output is cited verbatim: the net-benefit test was observed failing before M2.
- The accepted non-closure is stated as accepted, with its recorded rationale, in the run-phase
  report as well as in the probe file.
- The three deliberate behaviour changes of `spec.md` §E are each stated as deliberate in the
  run-phase report and carried to CHANGELOG at sync-phase: the five newly-deleted SessionEnd keys
  (REQ-4), the newly-admitted backup-only file and its token restore (REQ-5), and the fixture
  stand-in-key change (REQ-12).
- No real settings file or GLM credential was read or written by any test.
