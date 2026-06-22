# Design — SPEC-V3R6-CG-MODE-HARDENING-001

This design doc covers the two non-mechanical decisions: the detector source-of-truth realignment (REQ-CGH-006, the headline) and the settings atomicity/ordering refactor (REQ-CGH-002/003/005). The remaining requirements (001 launch-safety, 004 doc, 007 URL validation, 008 precondition, 009 coverage) are mechanical and follow existing in-repo patterns cited in plan.md §H.

## §A. The CG-mode invariant (what the design must protect)

CG mode's contract: the **leader pane runs Claude with a CLEAN env** (no GLM credentials), while **teammate panes inherit GLM credentials from the tmux SESSION env**. The cost optimization holds only when teammate spawning correctly detects "we are in CG mode" and routes to `moai glm`.

Two facts make the current detector wrong:
1. The leader's process env is deliberately clean of GLM creds (that is the whole point of CG mode).
2. `IsCGMode` (cg_detect.go:63) gates on `hasGLMEnv()`, which reads exactly that clean process env.

Consequence: a `moai worktree new <SPEC> --team` issued from the clean leader pane can see no GLM env (tmux `update-environment` import is non-deterministic across shells) → `IsCGMode` false → teammates spawn as `moai cc` (Claude) → silent cost-optimization defeat. **The detection signal contradicts the design it is meant to detect.**

## §B. Detector SSOT decision (REQ-CGH-006)

### Options considered

- **Option A — process env (status quo)**: REJECTED. Contradicts the clean-leader design; non-deterministic.
- **Option B — `llm.yaml team_mode` as authoritative**: `persistTeamMode(root, "cg")` writes `team_mode: cg` deterministically at `moai cg` time (launcher.go:189). This is a tracked, deterministic file that is set exactly when CG mode is established. PREFERRED primary signal.
- **Option C — tmux SESSION env via `tmux show-environment`**: reads the session-level env (where CG actually injects GLM creds), not the pane's imported process env. Deterministic within a tmux session. PREFERRED corroborating signal.
- **Option D — layered OR (B as additive primary, C as new clean-env path, process-env retained as a sufficient fallback)**: SELECTED. `team_mode == "cg"` is the NEW authoritative/additive signal that makes CG detectable from a clean leader process env; the tmux SESSION env corroborates GLM-cred presence (and is where the `REQ-WTL-009` drift warning relocates). Critically, the EXISTING `teammateMode == "tmux"` + GLM-env-present path is PRESERVED as a sufficient fallback (a layered OR, NOT an exclusive gate) so that the sibling SPEC's existing detection behavior is unchanged and its tests stay green.

### Why a layered OR, not an exclusive gate (D1 — sibling-test preservation)

An exclusive gate (`if team_mode != "cg": return false`) would BREAK the sibling test `TestIsCGMode_InTmux_TeammateMode_GLMToken_True` (`cg_detect_test.go:42`), which sets ONLY `teammateMode=tmux` + `ANTHROPIC_AUTH_TOKEN` (NO `llm.yaml team_mode=cg`) and asserts `true`. Under an exclusive gate that case returns false → semantic test failure → C-7 / plan.md R1 violation. The NEW capability (detect CG from a clean process env via `team_mode`/session-env) is therefore purely ADDITIVE: it adds a detection path, it does not remove the existing one.

### Selected design

```
# 2-arg signature PRESERVED — IsCGMode(settingsPath, stderrSink) — to avoid
# breaking all 10 call sites (1 prod new.go:246 + 9 cg_detect_test.go).
# The session-env reader is a package-level injectable var (NOT a new param).
IsCGMode(settingsPath, stderrSink):
    if not InTmuxSession():            return false        # CG requires tmux (unchanged)

    # NEW additive path: detect CG from a clean leader process env.
    teamMode = readLLMTeamMode()       # llm.yaml team_mode (NEW authoritative signal)
    if teamMode == "cg" and sessionEnvHasGLM():   # sessionEnvReaderFn — tmux show-environment, NOT process env
        return true
    if teamMode == "cg" and not sessionEnvHasGLM():
        emit REQ-WTL-009 drift warning             # reconciled, not deleted (relocated to session-env layer)
        return false

    # PRESERVED existing path (sufficient fallback — keeps sibling tests green):
    settings = readSettingsLocal(settingsPath)
    if settings.teammateMode != "tmux":   return false
    if not hasGLMEnv():                                # process-env check (existing behavior)
        emit REQ-WTL-009 drift warning                 # existing drift warning (unchanged)
        return false
    return true
```

The two `team_mode == "cg"` branches are the NEW capability; the trailing `teammateMode == "tmux"` + `hasGLMEnv()` block is the EXISTING behavior preserved verbatim. The OR structure means: CG is detected if (team_mode=cg AND session-env has GLM) OR (teammateMode=tmux AND process-env has GLM). The sibling test (teammateMode=tmux + token, no team_mode) takes the second disjunct and returns true — unchanged.

### Why this reconciles REQ-WTL-009 rather than deleting it

`REQ-WTL-009` originally warned "teammateMode=tmux but GLM env absent → fall back to P2". Under the layered design the drift warning fires in BOTH the new path (team_mode=cg but session env lacks GLM creds) AND the preserved existing path (teammateMode=tmux but process env lacks GLM creds). The warning's PURPOSE (surface a credential-rotation drift) and its existing trigger are both preserved; the new path ADDS a second, session-env-based trigger. C-7 + AC-CGH-006 Scenario 6b enforce that the sibling SPEC's `IsCGMode` tests stay green.

### Testability seam

`IsCGMode` does NOT change its 2-arg signature `IsCGMode(settingsPath, stderrSink)` (all 10 call sites — 1 prod `new.go:246` + 9 `cg_detect_test.go` — are 2-arg; a signature change would break them all). The session-env reader is introduced as a NEW package-level injectable var `var sessionEnvReaderFn = <tmux show-environment reader>` (no existing equivalent; the worktree layer calls `tmux.NewSessionManager()` directly without a recording-fake var). Tests override `sessionEnvReaderFn` with a recording fake so AC-CGH-006 can assert detection from a clean process env without a live tmux.

### Import-cycle note

`cg_detect.go` deliberately avoids importing `internal/cli/*` (R4 mitigation, documented at cg_detect.go:19). Reading `llm.yaml team_mode` must use a minimal local reader (like the existing `settingsLocalMin` view) or a shared low-level config reader — NOT an `internal/cli` import.

## §C. Settings atomicity + ordering decision (REQ-CGH-002/003/005)

### Current failure modes (confirmed)

1. **Ordering (REQ-CGH-002)**: `applyCGMode` injects tmux env (returns on failure at launcher.go:175) BEFORE `removeGLMEnv` (193). A tmux-inject failure leaves stale GLM creds in the leader config.
2. **Double write (REQ-CGH-003)**: `removeGLMEnv` sets `TeammateMode=""` (launcher.go:229); `ensureSettingsLocalJSON` then sets it to `"tmux"` (glm.go:509). Two non-atomic writes; a crash between them leaves `teammateMode` absent → teammates lose GLM inheritance (#468) AND `IsCGMode` false.
3. **Non-atomic RMW (REQ-CGH-005)**: all mutators use unlocked `ReadFile→Unmarshal→MarshalIndent→os.WriteFile`. Concurrent launches race → truncation / last-writer-wins.

### Selected design

Introduce one helper in `internal/cli/settings.go`:

```
mutateSettingsLocal(path string, mutate func(*SettingsLocal)) error:
    acquire flock(path)                         # serialize concurrent writers (internal/cli/team_spawn_lock_unix.go infra: lockFile/unlockFile)
    defer release
    s := readOrEmpty(path)                       # tolerate absent (EC-1) + empty (EC-2)
    mutate(&s)                                   # caller mutates struct in memory
    writeAtomic(path, s)                         # temp-file + os.Rename (saveLLMSection pattern, glm.go:680-698)
```

CG leader cleanup becomes a SINGLE `mutateSettingsLocal` call at the TOP of `applyCGMode`:

```
applyCGMode:
    require InTmuxSession() AND Detector.IsAvailable()   # REQ-CGH-008
    mutateSettingsLocal(path, func(s){                    # REQ-CGH-002 (top) + REQ-CGH-003 (single RMW)
        stripGLMCreds(s.Env)                              # leader stays clean
        s.TeammateMode = "tmux"                           # set once, atomically
    })
    persistTeamMode(root, "cg")                           # llm.yaml SSOT (unchanged; feeds REQ-CGH-006)
    injectTmuxSessionEnv(...)                             # failure now cannot leave stale leader creds
```

`removeGLMEnv`, `syncPermissionModeToSettingsLocal`, and `ensureSettingsLocalJSON` are re-expressed as `mutate` closures routed through `mutateSettingsLocal`, eliminating the three independent unlocked writers.

### User-key preservation (R3 mitigation)

The `mutate` closure only touches the GLM credential keys + `teammateMode`. `defaultMode`, `env.PATH`, and unrecognized keys are preserved by struct round-trip (per `internal/cli/CLAUDE.md` settings-helper convention + `CLAUDE.local.md §22`). DDD characterization tests on the three existing mutators capture this BEFORE the refactor (plan.md M2 step 1).

### Lock platform note

`team_spawn_lock_unix.go` is unix-tagged. The locking helper needs a Windows-safe fallback (no-op or `LockFileEx`) so the cross-platform build (C-4 / REQ-CGH-001) is preserved. CG runtime is POSIX-only (tmux), but the helper compiles on all platforms.

## §D. Security: URL validation (REQ-CGH-007)

Validation lives at the config boundary (`internal/config/validation.go` LLM section, alongside the existing model-name checks at :349-352). Rule: `base_url` must parse as a well-formed `https://` URL AND its host must be either the canonical `api.z.ai` family OR an explicit user-configured value present in `llm.yaml` (the constraint guards against silent token redirect, not legitimate self-host overrides — R5). `DefaultGLMBaseURL` always passes. On failure, return a `ValidationError` with the field path + offending value (matching the existing `checkStringField` error shape).

## §E. Non-goals (design boundary)

Per spec.md §H: no leader-process-env scrubbing (disproven), no `--team` pattern/swarm redefinition (sibling SPEC), no new subcommands, no model-selection changes, no `SettingsLocal` schema redesign.
