# Plan — SPEC-CONFIG-DEAD-SWEEP-001

## §A. Context

Three dead non-ralph config surfaces are removed in a behavior-preserving sweep. Evidence verified 2026-08-04 via grep across `internal/`, `cmd/`, `pkg/` (non-test). The ralph half of the sweep is owned by `SPEC-RALPH-CONFIG-REDESIGN-001` (§H of spec.md).

**One reversal-likely design decision is lead-loaded into M2** (state_dir option a vs b). Mechanical removals are deferred to M3-M4.

## §B. Known Issues

- **ValidSessionTTLs is live.** The task brief said "remove LoadCacheConfig/ValidSessionTTLs" but `internal/settings/schema_sections.go:328` consumes `config.ValidSessionTTLs()` for the web settings dropdown. **Resolution:** extract `ValidSessionTTLs` to a standalone constant slice (or a smaller stub file) so the seam stays live; delete only `LoadCacheConfig`, the `CacheConfig` struct, and the file-loading registration. This is a deviation from the brief's literal "remove ValidSessionTTLs" wording, taken on the evidence.
- **`loader.go:345` "learning is legacy" comment is wrong.** Verified: `learning` is LIVE at `internal/cli/hook.go:551-1106`. Fix the comment, do not remove the sub-system.
- **`audit_registry.go:75` observability comment is wrong.** Verified: `observability_master.go:85` reads `enabled` live. Fix to "partial direct-read".

## §C. Pre-flight (before M1)

- Confirm no new consumers of `cfg.Research`, `cfg.State.StateDir`, or `LoadCacheConfig` have landed since 2026-08-04 (re-run the grep at run-phase start — `feedback_defect_claim_verification.md`).
- Confirm `make build` baseline is green.
- Confirm `go test ./...` baseline is green.

## §D. Constraints

- Template-First: `internal/template/templates/.moai/config/sections/` edited first, `make build`, local sync.
- `ValidSessionTTLs` extraction MUST land in the same commit as the cache_config.go slim-down (no intermediate broken state).
- No behavior change to `moai update`, `moai web`, statusline, or `findStateDir()`.

## §E. Self-Verification (plan-phase)

- [ ] SPEC ID regex PASS (executed at authoring: `SPEC-CONFIG-DEAD-SWEEP-001` matched `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`).
- [ ] Frontmatter 12 canonical fields present.
- [ ] Out of Scope section carries ≥1 `### Out of Scope — <topic>` H3 with `-` bullets.
- [ ] Every REQ uses a GEARS pattern (Ubiquitous / When / Where).
- [ ] Every AC in acceptance.md is Given-When-Then and binary-testable.
- [ ] No implementation detail (function names are cited as evidence, not prescribed as design).

## §F. Milestones

### M1 — Template-First deletion of cache.yaml + research.yaml

- Delete `internal/template/templates/.moai/config/sections/cache.yaml`.
- Delete `internal/template/templates/.moai/config/sections/research.yaml`.
- Run `make build` to regenerate the embedded FS.
- Verify the catalog/manifest (if section files are registered in a catalog) no longer references them; if it does, remove the registration.

### M2 — state_dir design decision (option a) + removal

**Decision: option (a) — remove the dead field, keep the hardcoded literal as SSOT.**

Tradeoff documented:
- **Option (a) (chosen):** smallest diff; `findStateDir()` and `state_guard.go` already hardcode `.moai/state/` and that is the de-facto SSOT. Cost: the YAML key disappears, so any user who set `state.state_dir` to a custom path silently loses that knob (but it never worked anyway — zero readers).
- **Option (b) (rejected):** wire `findStateDir()` and `state_guard.go` to read from `cfg.State.StateDir`. Higher risk (changes runtime path resolution), more code, and only useful if there is evidence users want a custom state dir — none found.

Concrete changes:
- `internal/config/types.go:562` — remove `StateDir string yaml:"state_dir"` from the `StateConfig` struct.
- `internal/config/defaults.go:620` — remove the `StateDir: DefaultStateDir,` line.
- `internal/config/defaults.go:150` — keep `DefaultStateDir` constant IF other code reads it; otherwise remove. (Run-phase grep for `DefaultStateDir` consumers decides.)
- `internal/template/templates/.moai/config/sections/state.yaml` — remove the `state_dir:` key line.
- `.moai/config/sections/state.yaml` (local) — same.
- `internal/cli/state.go:210-211` and `internal/worktree/state_guard.go:25` — UNCHANGED (hardcoded literal stays as SSOT).

### M3 — Slim cache_config.go (extract ValidSessionTTLs)

- Create or inline a standalone constant slice for the TTL options (e.g., in `internal/config/cache_config.go` as a package-level `var validSessionTTLs = [...]string{...}` exposed via `ValidSessionTTLs()`). The accessor MUST survive.
- Remove `LoadCacheConfig` (cache_config.go:136), the `CacheConfig` struct type if orphaned after ValidSessionTTLs extraction, and any file-loader registration.
- Remove `internal/template/templates/.moai/config/sections/cache.yaml` (already in M1).
- Remove `.moai/config/sections/cache.yaml` (local).
- Remove the `cache` section entry from any loader map (e.g., `loader.go` section-file registration if present).

### M4 — Remove research.yaml loader + types

- `internal/config/loader.go:277-284` — remove the `researchFileWrapper` load + writeback.
- `internal/config/types.go` — remove the `Research` field from the root `Config` struct (and the `ResearchConfig` type if orphaned).
- `internal/config/defaults.go` — remove the `Research:` default initialization if present.
- Remove `.moai/config/sections/research.yaml` (local).
- Remove the `research` row from `internal/template/templates/.claude/rules/moai/core/settings-management.md` (the settings-management doc enumerates section files).
- Remove the `cache` and `research` rows from the same doc if they were listed.

### M5 — Stale comment fixes (fold-in)

- `internal/config/loader.go:345` — change "Legacy sub-system (out-of-scope)" to "Live sub-system — consumed at internal/cli/hook.go:551-1106".
- `internal/config/audit_registry.go:75` — change "no Go loader yet" to "partial direct-read (observability_master.go:85 reads `enabled` live)".

### M6 — Verify

- `go build ./...` exits 0.
- `go test ./...` exits 0.
- `golangci-lint run` exits 0 (or baseline warnings unchanged).
- Re-run the grep trio to confirm zero remaining readers of removed symbols.
- `moai update` smoke test on a `/tmp/test-project` — confirm remaining sections merge cleanly.

## §G. Anti-Patterns

- **AP-1: blanket delete ValidSessionTTLs** — would silently empty the web settings dropdown. Forbidden.
- **AP-2: edit local before template** — violates Template-First (CLAUDE.local.md §2).
- **AP-3: remove `learning` because the stale comment said "legacy"** — the comment is wrong; `learning` is LIVE. Only the comment is fixed.
- **AP-4: cross into ralph.yaml** — collision with `SPEC-RALPH-CONFIG-REDESIGN-001`. Stay in the three named files.
- **AP-5: skip `make build` after template edits** — embedded FS goes stale; the deleted files would still ship in the binary.

## §H. Cross-References

- `acceptance.md` — AC-CDS-001..010.
- `SPEC-RALPH-CONFIG-REDESIGN-001/plan.md` — the ralph half; sequence or isolate.
- `feedback_defect_claim_verification.md` — re-run the grep at run-phase start.
- `CLAUDE.local.md` §2 (Template-First), §25 (Template Internal-Content Isolation).
