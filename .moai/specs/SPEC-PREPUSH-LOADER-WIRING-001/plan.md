# Implementation Plan — SPEC-PREPUSH-LOADER-WIRING-001

> READ-only loader wiring for the `git_strategy` config section. Tier S, `internal/config`.

## A. Context

This plan implements the READ half of the `git_strategy` config pipeline: a
`gitStrategyFileWrapper` struct, a `loadGitStrategySection` loader method (an exact
structural mirror of `loadGitConventionSection`), and the wired call inside `Load()`.

The change closes the 3rd dead-config gap in the PREPUSH chain — see spec.md §A. After this
SPEC, the user's `git-strategy.yaml` values become observable at runtime, completing the
end-to-end chain that SPEC-PREPUSH-MODE-WIRING-001's `resolvePrePushAction()` depends on.

## B. Known Issues / Ground-Truth Anchors

All anchors verified against live source during plan-phase (cite these in implementation):

- Sibling pattern to mirror: `loadGitConventionSection` at `loader.go:149-161`.
- Sibling wrapper to mirror: `gitConventionFileWrapper` at `types.go:1071-1073` — BUT note
  it wraps `models.GitConventionConfig`; the new wrapper wraps the **local-package**
  `GitStrategyConfig` (declared at `types.go:134`, field on `Config` at `types.go:20`).
- Wired-call insertion point: immediately after `l.loadGitConventionSection(...)`
  (`loader.go:56`) inside `Load()`.
- Non-strict unmarshal: `loadYAMLFile` at `loader.go:398-413` (plain `yaml.Unmarshal`).
- `loadedSections` map initialized in `Load()` at `loader.go:35`.
- Top-level YAML key: `git_strategy:`.
- `ActiveModeProfile()` at `types.go:165` returns `(*ModeProfile, bool)` — a TWO-value
  return; the `bool` MUST be checked before dereferencing the `*ModeProfile`.
- Compiled default `Hooks.PrePush` = `"warn"` for all three modes (`defaults.go:249,261,276`).

## C. Pre-flight Checklist

- [ ] Read `internal/config/loader.go` (sibling loaders + `Load()` + `loadYAMLFile`)
- [ ] Read `internal/config/types.go` (`gitConventionFileWrapper`, `GitStrategyConfig`,
      `Config.GitStrategy`, `ActiveModeProfile`, `HooksConfig`)
- [ ] Read `internal/config/manager.go` `Save()` (confirm git-strategy persistence absent —
      it MUST stay absent)
- [ ] Confirm `go test ./internal/config/...` is green at baseline BEFORE changes

## D. Constraints

- Tier S minimal: mirror the existing sibling loader pattern exactly. Do NOT refactor the
  loader, do NOT add a generic section-registry, do NOT touch the WRITE/Save path,
  `SetSection`/`GetSection`, `validation.go`, `defaults.go`, or templates.
- No new behavioral idiom: identical `slog.Warn` + keep-defaults-on-error and
  set-flag-on-success-only semantics as the siblings.
- Estimated footprint: ~3-4 files, < 150 LOC (1 struct + 1 method + 1 wired line + tests).

## E. Self-Verification & Regression Risk (READ CAREFULLY before run)

### E.1 Fixture-driven test-expectation shift (a potential, low-likelihood shift)

[Advisory — flag to implementer] IF any **existing** `internal/config` loader test routes a
`git-strategy.yaml` fixture through `Loader.Load()` and asserts `cfg.GitStrategy == <compiled
default>` (or a specific default-derived field) afterward, that test would **load real values**
once `loadGitStrategySection` is wired in. Independent plan-audit found **no such test
currently** — the only `git_strategy` fixture (`git_strategy_nested_test.go`) uses a direct
`yaml.Unmarshal`, not `Loader.Load()` — so this is a defensive, low-likelihood check rather
than a guaranteed regression. Run the full `go test ./internal/config/...` suite to confirm;
if a genuine shift surfaces, **reconcile the expectation (do not revert the wiring)**.

This is **expected, not a defect** — it is the dead-config being fixed. The implementer MUST:

1. Run the FULL `go test ./internal/config/...` suite (per CLAUDE.local.md §6 — after fixing
   ANY test, run the full suite to catch cascading failures).
2. Identify any test whose expectation shifted because its fixture now actually loads
   `git-strategy.yaml`.
3. Reconcile each such expectation to the post-load truth (the loaded value), NOT revert the
   wiring. Update the assertion to reflect that the section now loads.
4. Distinguish this expected shift from a genuine regression: a genuine regression would be a
   test in an UNRELATED loader (`user`/`language`/`quality`/etc.) breaking — that would
   signal an unintended side effect and MUST be investigated, not rubber-stamped.

Do not hit this blind: grep the existing test fixtures for `git-strategy.yaml` / `git_strategy`
before running, so the shift is anticipated.

### E.2 Verification commands (run-phase, read-only batch)

```bash
# Functional
go test ./internal/config/...
go test -coverprofile=cover.out ./internal/config/...

# Dead-config elimination (loader symbol + wired call ≥ 2 matches)
grep -n 'loadGitStrategySection' internal/config/loader.go

# Wrapper struct exists
grep -n 'gitStrategyFileWrapper' internal/config/types.go

# Scope boundary: Save() still does NOT persist git-strategy (stays 0)
grep -c 'git-strategy.yaml' internal/config/manager.go

# Lint baseline
golangci-lint run --timeout=2m ./internal/config/...
```

## F. Milestones (priority-ordered, no time estimates)

### M1 — Wrapper struct (`types.go`)
Add `gitStrategyFileWrapper` mirroring `gitConventionFileWrapper`:
- Top-level YAML key `git_strategy:`
- Single field wrapping the local-package `GitStrategyConfig`
- Godoc one-liner referencing the section file.
Priority: High (M2 depends on it).

### M2 — Loader method + wired call (`loader.go`)
- Add `loadGitStrategySection(dir string, cfg *Config)` as an exact mirror of
  `loadGitConventionSection`: seed `wrapper := &gitStrategyFileWrapper{GitStrategy: cfg.GitStrategy}`
  → `loadYAMLFile(dir, "git-strategy.yaml", wrapper)` → on `loaded`, set
  `cfg.GitStrategy = wrapper.GitStrategy` + `l.loadedSections["git_strategy"] = true`; on
  error `slog.Warn` + return (keep defaults).
- Wire `l.loadGitStrategySection(sectionsDir, cfg)` into `Load()` immediately after the
  `loadGitConventionSection` call.
Priority: High.

### M3 — Unit tests (`loader_test.go` or new `git_strategy_loader_test.go`)
Cover REQ-PLW-001..008:
- present → values loaded + `loadedSections["git_strategy"] == true`
- absent → defaults kept + flag unset
- partial → specified keys override, unspecified keep defaults
- unknown keys → `Load()` does not fail
- end-to-end: fixture `mode: team` + `team.hooks.pre_push: enforce` →
  `ActiveModeProfile()` returns `(profile, true)` with `profile.Hooks.PrePush == "enforce"`
  (contrast: without the fixture, the default is `warn`).
Use `t.TempDir()` for all fixtures (CLAUDE.local.md §6 test isolation). Table-driven.
Priority: High.

### M4 — Full-suite reconciliation
Run `go test ./internal/config/...` (full), reconcile any fixture-driven expectation shift
per §E.1, confirm no unrelated loader regressed. Priority: High (closing gate).

## G. Anti-Patterns to Avoid

- Wrapping `models.GitStrategyConfig` — wrong; `GitStrategyConfig` is local-package.
- Dereferencing `ActiveModeProfile()` without checking the returned `bool`.
- Adding a `saveSection("git-strategy.yaml", ...)` call — that is the OUT-OF-SCOPE WRITE path.
- Reverting the wiring when an existing fixture-driven test shifts — that is the bug being fixed.
- Introducing strict-mode unmarshal or an enum gate — out of scope, changes behavior.

## H. Cross-References

- spec.md §A (ground-truth facts), §C (Exclusions)
- acceptance.md (Given-When-Then + grep ACs)
- SPEC-PREPUSH-WIRING-001, SPEC-PREPUSH-MODE-WIRING-001 (dead-config chain predecessors)
- `internal/config/CLAUDE.md` (module conventions: section-file layout, `Loader`+`Load()` pattern)
- CLAUDE.local.md §6 (Go test execution rules — full-suite-after-fix, `t.TempDir()`)
