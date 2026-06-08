# Acceptance Criteria — SPEC-PREPUSH-LOADER-WIRING-001

> All scenarios are testable. Grep ACs are authored byte-precise and idiom-tolerant: they
> match only real post-implementation symbols, contain no sibling-section enumeration, and
> reference no nonexistent path or test name (per the sibling lessons L_ac_grep_read_r_flag
> and L_plan_audit_grep_idiom_recurrence).

---

## AC-PLW-001 — git-strategy.yaml present → values loaded + section flagged (REQ-PLW-001, REQ-PLW-002)

**Given** a `.moai/config/sections/` directory (created via `t.TempDir()`) containing a
`git-strategy.yaml` file whose top-level `git_strategy:` block sets a non-default value
(e.g. `mode: team` with `team.hooks.pre_push: enforce`),
**When** `NewLoader().Load(configDir)` is called,
**Then** the returned `cfg.GitStrategy` reflects the file (the non-default value is present,
not the compiled default),
**And** `loader.LoadedSections()["git_strategy"]` returns `true`.

Verification (unit test):
```go
loaded := loader.LoadedSections()
require.True(t, loaded["git_strategy"])
// the file's non-default mode-profile hook value is observed (see AC-PLW-007 for the
// ActiveModeProfile assertion)
```

---

## AC-PLW-002 — git-strategy.yaml absent → defaults kept + flag unset (REQ-PLW-003)

**Given** a `.moai/config/sections/` directory that contains **no** `git-strategy.yaml` file
(but is otherwise a valid sections dir),
**When** `Load()` is called,
**Then** `cfg.GitStrategy` equals the compiled defaults from `NewDefaultGitStrategyConfig()`
(e.g. `cfg.GitStrategy.Mode == "team"` default, mode-profile `Hooks.PrePush == "warn"`),
**And** `loader.LoadedSections()["git_strategy"]` is `false` (the key is unset), mirroring
the absent-file behavior of the sibling loaders.

Verification (unit test):
```go
loaded := loader.LoadedSections()
require.False(t, loaded["git_strategy"])
```

---

## AC-PLW-003 — partial git-strategy.yaml → specified keys override, unspecified keep defaults (REQ-PLW-004)

**Given** a `git-strategy.yaml` that sets only a subset of keys (e.g. only
`git_strategy.mode: personal`, omitting the mode-profile hook blocks),
**When** `Load()` is called,
**Then** the specified key is overridden (`cfg.GitStrategy.Mode == "personal"`),
**And** every key absent from the file retains its compiled default (e.g. the `personal`
profile's `Hooks.PrePush` stays the compiled default `"warn"` because the file did not
override it).

---

## AC-PLW-004 — malformed / unknown-key git-strategy.yaml → Load() does not fail (REQ-PLW-005)

**Given** a `git-strategy.yaml` that contains a valid `git_strategy:` block plus an unknown /
extraneous top-level or nested key (e.g. `git_strategy.bogus_key: 123`),
**When** `Load()` is called,
**Then** `Load()` returns no error (the loader uses non-strict `yaml.Unmarshal`, so unknown
keys are silently ignored),
**And** the recognized keys are loaded normally.

Verification (unit test):
```go
cfg, err := loader.Load(configDir)
require.NoError(t, err)
require.NotNil(t, cfg)
```

---

## AC-PLW-005 — dead-config eliminated: wrapper + loader symbol exist (REQ-PLW-006, REQ-PLW-007)

**Given** the implemented change,
**When** the codebase is grepped for the new symbols,
**Then** the `gitStrategyFileWrapper` struct exists in `types.go` AND the
`loadGitStrategySection` symbol appears at least twice in `loader.go` (the method definition
AND the wired call inside `Load()`).

Verification (read-only grep):
```bash
# Wrapper struct — exactly the new symbol, no sibling enumeration
grep -n 'gitStrategyFileWrapper' internal/config/types.go
# Expected: ≥1 match (the struct declaration)

# Loader method definition + wired call inside Load()
grep -n 'loadGitStrategySection' internal/config/loader.go
# Expected: ≥2 matches (func definition + the l.loadGitStrategySection(...) call)
```

---

## AC-PLW-006 — wired call invoked from Load() (REQ-PLW-007)

**Given** the implemented change,
**When** `loader.go` is inspected for the call site,
**Then** `Load()` contains an `l.loadGitStrategySection(sectionsDir, cfg)` invocation
(method-receiver call form), placed adjacent to the `loadGitConventionSection` call.

Verification (read-only grep — call-syntax precise, tolerant of leading whitespace):
```bash
grep -nE 'l\.loadGitStrategySection\(' internal/config/loader.go
# Expected: ≥1 match (the wired invocation; the func-definition line lacks the 'l.' receiver-call prefix)
```

---

## AC-PLW-007 — end-to-end chain completion: ActiveModeProfile reads YAML hook value (REQ-PLW-008)

**Given** a `git-strategy.yaml` fixture (in a `t.TempDir()` sections dir) declaring
`git_strategy.mode: team` AND `git_strategy.team.hooks.pre_push: enforce`,
**When** `Load()` is called and `cfg.GitStrategy.ActiveModeProfile()` is invoked,
**Then** the accessor returns `(profile, true)` (mode is valid),
**And** `profile.Hooks.PrePush == "enforce"` — proving the YAML value flowed end-to-end and
SPEC-PREPUSH-MODE-WIRING-001's `resolvePrePushAction()` now reads the real file value.

**Contrast (the dead-config it fixes):** without `loadGitStrategySection` wired in, the same
fixture would yield `profile.Hooks.PrePush == "warn"` (the compiled default), because the
file would never be read. The test SHOULD assert the post-wiring value is `"enforce"`.

Verification (unit test):
```go
cfg, err := loader.Load(configDir)
require.NoError(t, err)
profile, ok := cfg.GitStrategy.ActiveModeProfile()
require.True(t, ok)
require.Equal(t, "enforce", profile.Hooks.PrePush)
```

---

## AC-PLW-008 — scope boundary: Save() unchanged (WRITE path not added) (REQ-PLW-009, Exclusions)

**Given** the implemented change,
**When** `manager.go` is grepped for git-strategy persistence,
**Then** `Save()` still does NOT persist `git-strategy.yaml` — the WRITE path was
intentionally not added (READ-only scope).

Verification (read-only grep — count stays 0):
```bash
grep -c 'git-strategy.yaml' internal/config/manager.go
# Expected: 0 (Save() persists only user/language/quality/git-convention/llm)
```

---

## Edge Cases

- **Empty git_strategy block** (`git_strategy:` present but with no children) → loader seeds
  defaults, `yaml.Unmarshal` overlays nothing → result equals defaults; `loadedSections`
  flag is still `true` because the file existed and parsed (consistent with sibling loaders,
  which key the flag off `loaded` from `loadYAMLFile`, not off whether values changed).
- **Invalid YAML syntax** (genuinely unparseable, not just unknown keys) → `loadYAMLFile`
  returns an error; `loadGitStrategySection` logs `slog.Warn` and keeps defaults; `Load()`
  does NOT propagate the error (mirror sibling resilience). `loadedSections["git_strategy"]`
  stays unset.

---

## Definition of Done

- [ ] AC-PLW-001 through AC-PLW-008 all pass.
- [ ] Edge cases (empty block, invalid YAML) covered by tests.
- [ ] `go test ./internal/config/...` (FULL suite) green; any fixture-driven expectation
      shift reconciled per plan.md §E.1 (not reverted).
- [ ] `golangci-lint run ./internal/config/...` clean.
- [ ] No change to `Save()`, `SetSection`/`GetSection`, `validation.go`, `defaults.go`,
      or templates (footprint bound per REQ-PLW-009).
- [ ] Coverage for `internal/config` does not regress.

---

## Quality Gate Criteria (TRUST 5)

- **Tested**: 8 ACs + 2 edge cases with table-driven `t.TempDir()` fixtures.
- **Readable**: loader method is a 1:1 structural mirror of the sibling — no new idiom.
- **Unified**: identical error/bookkeeping semantics as `loadGitConventionSection`.
- **Secured**: non-strict unmarshal cannot fail on hostile keys; no new attack surface.
- **Trackable**: REQ ↔ AC traceability in spec.md §E; dead-config chain referenced.
