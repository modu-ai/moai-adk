# Implementation Plan — SPEC-GITSTRATEGY-SAVE-ISOLATION-001

> Tier S (minimal): a bounded isolation fix in `internal/config`. 1 production file (`manager.go`) + the two existing tests as the verification harness. The reproduction test is already RED.

## §A. Context summary

`ConfigManager.Save()` (`internal/config/manager.go` ~line 155-201) unconditionally re-serializes all six owned sections. The web writer `writeProjectConfig` (`internal/web/projectconfig.go`) modifies only quality + git_convention but cannot prevent `Save()` from also rewriting `git-strategy.yaml` from the in-memory defaults — clobbering the sentinel the loader dropped on read. The fix must make `git-strategy.yaml` writes conditional on the git_strategy section actually having been modified, while preserving the two SAVE-WIRING tests.

## §B. Known constraint tension (read before choosing mechanism)

Two existing SAVE-WIRING tests pin opposite ends:

1. **`TestConfigManagerSaveGitStrategyRoundTrip`** — `Load()` → `SetSection("git_strategy", probe)` → `Save()` → fresh `Load()` recovers the mutated values. ⇒ git_strategy MUST be written when explicitly `SetSection`'d.
2. **`TestConfigManagerSaveCreatesGitStrategyFile`** — fresh `t.TempDir()` → `Load()` (no SetSection) → `Save()` → `git-strategy.yaml` exists with the `git_strategy:` key. ⇒ a bare `Save()` on a project with NO `git-strategy.yaml` MUST still create it.

3. **`TestWriteProjectConfigSectionIsolation`** (the regression / reproduction) — fixture HAS `git-strategy.yaml` (sentinel) on disk → `LoadRaw` → `SetSection(quality)` + `SetSection(git_convention)` → `Save()` → `git-strategy.yaml` MUST be byte-unchanged. ⇒ a `Save()` that did NOT `SetSection("git_strategy")` MUST NOT rewrite an EXISTING `git-strategy.yaml`.

The reconciling invariant: **`Save()` rewrites `git-strategy.yaml` only when the git_strategy section was modified via `SetSection`, OR when the file does not yet exist (greenfield create).** This satisfies all three tests simultaneously.

## §C. Pre-flight (read-only, before editing)

1. Re-run `go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation` → confirm RED (reproduction baseline).
2. Re-run `go test ./internal/config/ -run 'TestConfigManagerSaveGitStrategyRoundTrip|TestConfigManagerSaveCreatesGitStrategyFile'` → confirm GREEN (no-regression baseline).
3. Read `Save()` + `SetSection` + the `ConfigManager` struct fields in `manager.go` to confirm there is no existing dirty/modified tracking surface.

## §D. Fix mechanism (the chosen approach)

**Dirty-tracking on the git_strategy section, gating its conditional Save write.**

The minimal mechanism that satisfies §B's reconciling invariant:

- Add a private boolean field to `ConfigManager` (e.g. `gitStrategyDirty bool`) — set to `true` inside `SetSection` when `name == "git_strategy"`. Reset to `false` after a successful `Save()` (and on `Load()` / `Reload()` which replace the whole in-memory config).
- In `Save()`, replace the unconditional `saveSection("git-strategy.yaml", ...)` call with a guarded write:
  - Write `git-strategy.yaml` **if** `gitStrategyDirty` is true (the section was explicitly set this session), **or** the file does not yet exist on disk (greenfield create, satisfies `TestConfigManagerSaveCreatesGitStrategyFile`).
  - Otherwise skip the write, leaving the existing `git-strategy.yaml` (sentinel or user content) byte-unchanged (satisfies `TestWriteProjectConfigSectionIsolation`).

### D.1 Why this mechanism (vs alternatives)

- **vs "remove git_strategy from Save() entirely"** — would break `TestConfigManagerSaveGitStrategyRoundTrip` (set→save→reload would no longer persist). Rejected.
- **vs "make ALL six sections dirty-tracked"** — broader than needed; risks regressing the other four callers and the other sections' existing behavior. Out of scope per EX-5. The defect is git_strategy-specific (it is the only owned section whose on-disk content the loader does not fully model). Keep the fix surgical to git_strategy.
- **vs "round-trip-preserve unmodeled keys on load"** — would change the loader partial-override contract (EX-2 forbids) and is a much larger change. Rejected.

### D.2 Greenfield-create clause rationale

`TestConfigManagerSaveCreatesGitStrategyFile` runs on a fresh TempDir with no `git-strategy.yaml`. A pure dirty-only gate (write only when SetSection'd) would skip the write and fail that test. The "file does not yet exist" disjunct preserves the SAVE-WIRING file-creation contract for the greenfield path without re-introducing the isolation defect (the isolation test's fixture HAS the file, so the disjunct is false there).

> NOTE for run-phase: confirm the exact existence-check semantics against `m.root` + `defs.MoAIDir` + `defs.SectionsSubdir` + `"git-strategy.yaml"`. Verify the `loadedSections["git_strategy"]` map (already populated by the loader when the file was read) is NOT a sufficient substitute on its own — it tells you the file existed at load time but not whether it was modified. The dirty flag is the modification signal; the existence check is the create signal. Use both.

## §E. Self-verification (run-phase exit gate)

All read-only, batch in a single turn:

```bash
# 1. Primary reproduction — must flip RED → GREEN
go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation -v

# 2. SAVE-WIRING no-regression — must stay GREEN
go test ./internal/config/ -run 'TestConfigManagerSaveGitStrategyRoundTrip|TestConfigManagerSaveCreatesGitStrategyFile' -v

# 3. Full affected-package suite — no caller regression
go test ./internal/config/... ./internal/web/... ./internal/cli/... ./internal/profile/...

# 4. Race + vet hygiene
go test -race ./internal/config/...
go vet ./internal/config/...

# 5. Lint baseline
golangci-lint run ./internal/config/... ./internal/web/...
```

## §F. Milestones (priority-ordered, no time estimates)

- **M1 (Priority High)** — Add `gitStrategyDirty` tracking to `ConfigManager`: set in `SetSection` for `git_strategy`, reset in `Save()` (on success) and on `Load`/`Reload` (whole-config replacement).
- **M2 (Priority High)** — Replace the unconditional `git-strategy.yaml` `saveSection` call in `Save()` with the guarded write (dirty OR file-absent).
- **M3 (Priority High)** — Run §E batch 1+2: confirm reproduction flips GREEN and SAVE-WIRING stays GREEN.
- **M4 (Priority Medium)** — Run §E batch 3+4+5: confirm no caller regression, race-clean, lint-clean.
- **M5 (Priority Medium)** — Add a focused config-package unit test asserting the isolation invariant at the manager level (Save without SetSection(git_strategy) leaves an existing git-strategy.yaml byte-unchanged) — complements the web-level guard test by pinning the invariant in the package that owns the fix. (Optional if M3 coverage is judged sufficient; run-phase decides per coverage delta.)

## §G. Risks & Anti-Patterns

- **Risk R1 — over-generalization**: turning this into a full "dirty section map" Save redesign. Mitigation: EX-5 — keep the dirty flag git_strategy-specific.
- **Risk R2 — weakening the guard test**: making the reproduction pass by loosening its assertion instead of fixing Save. Mitigation: REQ-GSI-006 + DoD require the guard test body unchanged.
- **Risk R3 — breaking greenfield create**: a dirty-only gate skips the file on a fresh project. Mitigation: §D.2 file-absent disjunct, pinned by `TestConfigManagerSaveCreatesGitStrategyFile`.
- **Risk R4 — concurrency**: `gitStrategyDirty` is mutated under the existing `m.mu` lock (SetSection and Save both lock). Mitigation: place the flag mutation inside the already-held lock; verify with `go test -race`.
- **AP-1**: writing `git-strategy.yaml` from `m.config.GitStrategy` whenever the loader reports the section was loaded — this does not distinguish "loaded" from "modified" and re-introduces the bug. Use the dirty flag, not `loadedSections`.

## §H. Cross-References

- `internal/config/CLAUDE.md` — module conventions (Loader struct, atomic save, env-key SSOT)
- `internal/config/manager.go` — `Save()`, `SetSection`, `saveSection`, `atomicWrite`
- `internal/config/loader.go` — `loadGitStrategySection` (partial-override contract; do not change)
- Predecessor: `.moai/specs/SPEC-PREPUSH-SAVE-WIRING-001/plan.md` (the feature this isolates)
