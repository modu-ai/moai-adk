# Plan — SPEC-RALPH-CONFIG-REDESIGN-001

## §A. Context

`ralph.yaml` ships 25 leaf keys under `ralph:`. 23 are YAML-surface-only (the loader is non-strict, so they silently map to nothing); 2 are live (`lint_as_instruction`, `warn_as_instruction`). Separately, 3 of the 5 `RalphConfig` struct fields (`max_iterations`, `auto_converge`, `human_review`) are MISSING from the yaml, and even if present would not reach the engine because `deps.go:99-100` builds the engine from `NewDefaultRalphConfig()` rather than `cfg.Ralph`. The `RalphConfig` struct itself is ALL-live (5/5 fields consumed) and is UNCHANGED by this SPEC.

Two reversal-likely decisions are lead-loaded into §F: (1) the Option A vs Option B-doc engine-wiring decision (M2), and (2) the yaml shrink-and-regrow shape (M1). The mechanical follow-ons (settings schema sync) trail at the bottom.

### §A.1 Verified fact anchors (re-verify at run-phase start)

- `internal/config/types.go:330-343` — `RalphConfig` has exactly 5 fields (no Lsp/AstGrep/Loop/Completion/Hooks).
- `internal/ralph/engine.go:62` — `if e.cfg.AutoConverge {` (live read).
- `internal/ralph/engine.go:74` — `if e.cfg.HumanReview && state.Phase == loop.PhaseReview {` (live read).
- `internal/cli/deps.go:99-100` — `NewDefaultRalphConfig()` → `NewRalphEngine` (the root-cause wiring gap).
- `internal/cli/deps.go:135` — `ralphCfg.MaxIterations` consumed via the default-config path.
- `internal/hook/post_tool.go:426,439` — `LintAsInstruction` / `WarnAsInstruction` live reads.
- `internal/config/defaults.go:551-558` — `NewDefaultRalphConfig()` returns `MaxIterations:5, AutoConverge:true, HumanReview:true, LintAsInstruction:true, WarnAsInstruction:false`.
- Local + template `ralph.yaml` are byte-identical (verified via `diff` 2026-08-04).

## §B. Known Issues

- **InitDependencies ordering.** The engine is constructed at `deps.go:100` BEFORE the project config is loaded at `deps.go:144+`. Wiring `cfg.Ralph` (Option A) requires either (i) moving the config-load above the engine-construction, or (ii) lazy-building the engine after config load. Run-phase must verify there is no circular dependency (the config loader should not depend on the engine).
- **`loop.max_iterations` ≠ `RalphConfig.MaxIterations`.** The yaml nests `max_iterations` under `loop:`, but the struct field binds to top-level `ralph.max_iterations`. Removing the inert `loop:` block and adding top-level `ralph.max_iterations` are BOTH required for honesty; they are NOT the same key.
- **Settings-schema testdata.** `internal/settings/schema_sections.go` and `internal/settings/testdata/sections/ralph.yaml` carry the full 25-key ralph.yaml today. They must shrink to the 5-key surface in lockstep with the template edit, or the web settings test fails.
- **`NewDefaultRalphConfig()` is also used at `defaults.go:270`.** `NewDefaultConfig()` calls `NewDefaultRalphConfig()` to populate `cfg.Ralph` defaults before the yaml load merges over them. This is the DEFAULT layer — it is correct and stays. Option A changes ONLY the `deps.go` engine-construction site, not `defaults.go`.

## §C. Pre-flight (before M1)

- Re-run `grep -rn '\.StaleSeconds' --include='*.go' internal/ cmd/ pkg/` to confirm zero runtime consumers are still present (only the producer-side pipeline M3 removes) — re-verify before the M3 deletion.
- Re-run `grep -rn 'RalphConfig\.' internal/ cmd/ pkg/ | grep -v _test.go` to confirm the 5 field-read sites are unchanged.
- Confirm `make build` baseline is green.
- Confirm `go test ./internal/ralph/... ./internal/cli/... ./internal/hook/... ./internal/config/...` baseline is green.
- Enumerate the exact 23 inert leaf keys via YAML parse (the §A.1 anchor list).

## §D. Constraints

- Template-First: edit template, `make build`, sync local. Both files byte-identical before and after.
- `RalphConfig` struct is UNCHANGED — no field added, none removed, none renamed.
- No regression in `/moai loop` behavior — lint/warn-as-instruction injection MUST stay governed by user edits.
- No SSOT ambiguity introduced: `lsp.yaml` (gopls bridge) and `gate.yaml` (ast-grep gate) remain the uncontested SSOTs for their concerns; removing the inert `ralph.lsp.*` / `ralph.ast_grep.*` blocks just makes this honest.

## §E. Self-Verification (plan-phase)

- [x] SPEC ID regex PASS (executed at authoring: `SPEC-RALPH-CONFIG-REDESIGN-001` matched `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`).
- [x] Frontmatter 12 canonical fields present (+ `tier: M`, `related_specs`).
- [x] Out of Scope section carries ≥1 `### Out of Scope — <topic>` H3 with `-` bullets.
- [x] Every REQ uses a GEARS pattern (Ubiquitous / Event-detected / Capability gate / State-driven).
- [x] Every AC in acceptance.md is Given-When-Then and binary-testable.
- [x] The two-senses-of-dead distinction (§B.1 vs §B.2) is grounded in line-cited code reads.
- [x] The prior SPEC's vacuous AC-RCR-002 (grepped non-existent fields) is replaced with behavior-testing ACs.

## §F. Milestones

### M2 — Engine-wiring design decision: Option A (wire cfg.Ralph)

**Lead-loaded first because this is the most reversal-likely decision.**

**Decision: Option A (wire `cfg.Ralph` into `ralph.NewRalphEngine`).**

Rationale:

- **Option A (wire):** Change `deps.go:99-100` to use the loaded `cfg.Ralph` instead of `NewDefaultRalphConfig()`. Paired with M1's addition of the 3 missing yaml keys (with defaults matching `NewDefaultRalphConfig`), all 5 struct fields become genuinely user-tunable. Users who never edit ralph.yaml observe zero behavior change because the yaml defaults equal the compiled defaults. The `RalphConfig` struct is unchanged.
  - Risk: `InitDependencies` currently builds the engine (`deps.go:100`) BEFORE loading the project config (`deps.go:144+`). Wiring requires either reordering (load config first) or lazy engine construction. Run-phase must verify no circular dependency (config loader must not depend on the engine).
  - Evidence the reorder is bounded: `config.NewConfigManager()` is created at `deps.go:138` (after the engine), but `Load()` is called later. The config-load does not reference the engine, so moving `Load()` earlier should not introduce a cycle. Verify at run-phase start.

- **Option B-doc (fallback):** Leave `deps.go` as-is. Document in `ralph.yaml` header that `max_iterations`/`auto_converge`/`human_review` are "advisory — engine uses compiled defaults; only `lint_as_instruction`/`warn_as_instruction` take effect at runtime." Zero code change, zero behavior change. Leaves the root cause in place.
  - Choose this ONLY if M2 run-phase finds the reorder genuinely circular.

**Recommendation stands: Option A.** The wiring is the root-cause fix; B-doc perpetuates the hazard. If a future user discovers the wiring is too entangled, the fallback to B-doc is a single-commit retreat — no structural damage.

The decision is documented here, not deferred — run-phase proceeds under Option A unless the pre-flight finds the circular dependency.

### M1 — Template-First ralph.yaml shrink + regrow (5-key honest surface)

- Edit `internal/template/templates/.moai/config/sections/ralph.yaml`:
  - REMOVE the 5 inert top-level blocks (`enabled`, `lsp`, `ast_grep`, `loop`, `hooks`) — 23 inert leaf keys total.
  - ADD the 3 missing live keys with defaults matching `NewDefaultRalphConfig()`: `max_iterations: 5`, `auto_converge: true`, `human_review: true`.
  - PRESERVE `lint_as_instruction: true` and `warn_as_instruction: false`.
  - Result: exactly 5 keys under `ralph:`, all live. Header comment documents that `lsp.yaml` / `gate.yaml` govern the other concerns and that all 5 keys are read by Go code (cite engine.go / deps.go / post_tool.go line anchors).
- Sync local `.moai/config/sections/ralph.yaml` to match (or run the deploy step).
- Run `make build` to regenerate embedded templates.
- Do NOT touch `internal/config/types.go:330-343` (`RalphConfig`) — the struct is unchanged.

### M3 — OWN the Session.StaleSeconds removal (dead ralph.stale_seconds pipeline)

Verified 2026-08-04: `grep -rn '\.StaleSeconds' --include='*.go' internal/ cmd/ pkg/` returns ZERO runtime read sites — only the producer-side injection pipeline. The field is dead at the consumer side.

Ownership resolution (auditor's option a): `stale_seconds` is a `ralph.yaml` key (the wrapper at `types.go:1274` is `ralphFileWrapper.StaleSeconds`, bind target `ralph.stale_seconds`), and the sibling `SPEC-CONFIG-DEAD-SWEEP-001` explicitly disowns ralph.yaml keys. So this SPEC owns the removal — the v0.2.0 deferral is retracted.

Concrete removal list (run-phase executes all 4 deletions in one milestone):
- Delete the `SessionConfig.StaleSeconds` field declaration at `internal/config/types.go:569-571`.
- Delete the `ralphFileWrapper.StaleSeconds` field + its `yaml:"stale_seconds"` tag at `internal/config/types.go:~1274`.
- Delete the injection block at `internal/config/loader.go:267-269` (the `if wrapper.Ralph.StaleSeconds > 0 { cfg.Session.StaleSeconds = ... }` block).
- Delete the two default assignments at `internal/config/defaults.go:~154` and `internal/config/defaults.go:~628`.

No `ralph.yaml` edit is needed (the `stale_seconds` key is already absent from the shipped yaml — verified 2026-08-04). Template-local parity is therefore a no-op for this milestone.

Partition note: this removal is the `ralph.stale_seconds` pipeline, DISTINCT from the sibling SPEC's `cache.yaml` / `research.yaml` / `state.state_dir` scope. The two SPECs remain non-overlapping; coordinate via separate worktrees or sequential execution.

### M4 — Settings schema + docs sync

- `internal/settings/schema_sections.go` — shrink the ralph schema rows to the 5 retained keys; remove rows for the 23 inert leaves.
- `internal/settings/testdata/sections/ralph.yaml` — shrink to match the new 5-key template (byte-identical to the deployed `ralph.yaml`).
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` — update the `ralph.yaml` row to reflect the 5-key contract and that all 5 are live.
- `internal/web/assets/i18n.js` — keep i18n entries for the 5 retained keys; remove entries for the 23 removed leaves (run-phase grep `ralph\.` to enumerate the exact set).

### M5 — Verify

- `go build ./...` exits 0.
- `go test ./internal/ralph/... ./internal/cli/... ./internal/hook/... ./internal/config/... ./internal/settings/...` exits 0.
- `golangci-lint run` baseline unchanged (no NEW findings beyond the repo baseline).
- Re-run `grep -rn '\.StaleSeconds' --include='*.go' internal/ cmd/ pkg/` — confirm ZERO matches total (M3 removed the producer-side pipeline too, so this grep is now empty end-to-end).
- Re-run `grep -rn 'RalphConfig\.\(Lsp\|AstGrep\|Loop\|Completion\|Hooks\|Enabled\)' internal/ cmd/ pkg/` — confirm zero matches (these fields never existed; the grep is a guard against a future mis-edit).
- `/moai loop` smoke test on a `/tmp/test-project` with a deliberate lint error — confirm the loop runs, converges or aborts per the engine's normal flow, and that `lint_as_instruction: true` injects the finding (regression guard for REQ-RCR-009).
- `moai spec lint --strict .moai/specs/SPEC-RALPH-CONFIG-REDESIGN-001/` exits 0.

## §G. Anti-Patterns

- **AP-1: Remove a RalphConfig struct field.** All 5 fields are live. Removing any breaks compilation (engine.go:62/74). The struct is UNCHANGED.
- **AP-2: Cross into cache/research/state.state_dir surfaces.** Collision with `SPEC-CONFIG-DEAD-SWEEP-001`. Stay in `ralph.yaml` + `deps.go` + `schema_sections.go` + the `ralph.stale_seconds` pipeline (M3, owned here). Do NOT touch `cache.yaml` / `research.yaml` / `state.state_dir` — those are the sibling SPEC's partition.
- **AP-3: Edit local before template.** Template-First binds — edit template, `make build`, then sync local.
- **AP-4: Add new yaml keys without defaults matching `NewDefaultRalphConfig`.** A divergent default under Option A silently changes behavior for users who never edited ralph.yaml.
- **AP-5: Conflate `loop.max_iterations` with `ralph.max_iterations`.** They are different keys. The struct binds to the top-level path. Removing `loop:` AND adding top-level `ralph.max_iterations` are both required.
- **AP-6: Reuse the prior SPEC's vacuous AC greps.** The prior AC-RCR-002 grepped for `RalphConfig.Lsp|AstGrep|...` (fields that never existed) and passed vacuously. The rewritten ACs grep for REAL behavior (inert-key absence, live-key presence, build green).

## §H. Cross-References

- `acceptance.md` — AC-RCR-001..009.
- `SPEC-CONFIG-DEAD-SWEEP-001/plan.md` — the non-ralph half (`cache.yaml` / `research.yaml` / `state.state_dir`); sequence or isolate. NOTE: `Session.StaleSeconds` is OWNED by THIS SPEC (M3) — it is a `ralph.yaml` key, not part of the sibling's partition.
- `internal/ralph/engine.go` — evidence that `AutoConverge` / `HumanReview` are live `Decide()` inputs.
- `internal/cli/deps.go` — the engine-construction site M2 rewires.
- `internal/config/CLAUDE.md` — documents the non-strict loader mechanism by which the 23 inert leaves survived.
- `feedback_defect_claim_verification.md` — re-run every grep at run-phase start; the prior SPEC's failure is the named hazard.
