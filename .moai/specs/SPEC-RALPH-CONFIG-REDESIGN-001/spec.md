---
id: SPEC-RALPH-CONFIG-REDESIGN-001
title: "Redesign ralph.yaml so every retained key has a live Go consumer (shrink inert YAML surface, wire the engine)"
version: "0.3.0"
status: in-progress
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P1
phase: "v3.x target"
module: ralph-config
lifecycle: spec-anchored
tags: "ralph-yaml, dead-config, verification-claim-integrity, ralph-engine, deps-wiring, two-senses-of-dead"
tier: M
related_specs: [SPEC-CONFIG-DEAD-SWEEP-001, SPEC-CONFIG-KEY-HONESTY-001]
---

# SPEC-RALPH-CONFIG-REDESIGN-001 — Redesign ralph.yaml (honest surface + wired engine)

## HISTORY

- 2026-08-04 v0.1.0 — Initial draft. FAILED plan-audit (0.52, BLOCKING): the core claim conflated two distinct senses of "dead" and proposed removing live struct fields, which would not compile.
- 2026-08-04 v0.2.0 — Rewrite. Corrects the two-senses-of-dead conflation, grounds every claim in freshly verified code reads, re-scopes the plan to (M1) remove the genuinely inert YAML-surface keys, (M2) wire `cfg.Ralph` into the engine as the real root-cause fix, (M3) defer `Session.StaleSeconds` to the sibling sweep (zero runtime consumers confirmed, YAML already absent). Owns the ralph half of the dead-config sweep; the non-ralph half (`cache.yaml` / `research.yaml` / `state.state_dir`) is owned by `SPEC-CONFIG-DEAD-SWEEP-001`.
- 2026-08-04 v0.3.0 — Audit-debt fixes after re-audit PASS-WITH-DEBT (0.82). (D1) `lifecycle: spec-first` → `spec-anchored` (the prior value sat outside the canonical enum `{spec-anchored, spec-lite, exploratory}` per the frontmatter schema SSOT — CI lint only checks missing/empty, but the value was a strict-read type mismatch). (D2) Fold the `Session.StaleSeconds` removal into THIS SPEC's M3, resolving the owner gap: the sibling `SPEC-CONFIG-DEAD-SWEEP-001` explicitly disowns `ralph.yaml` keys, and `stale_seconds` is a `ralph.yaml` key, so this SPEC owns its removal. M3 changes from "defer" to "OWN the removal" with a concrete 4-site deletion list (types.go SessionConfig field + ralphFileWrapper field/tag + loader.go injection block + defaults.go default entries). The sibling SPEC still owns the cache/research/state.state_dir partition.

## §A. User Story

**As a** MoAI user editing `.moai/config/sections/ralph.yaml` expecting my edits to take effect,
**I want** every key in ralph.yaml to be read by the Go runtime, and the inert keys silently ignored by the loader to be removed,
**so that** editing ralph.yaml is not a verification-claim-integrity hazard — the file is an honest contract between user intent and engine behavior.

**Outcome hypotheses (verified 2026-08-04):**
- 23 inert leaf keys across 5 top-level blocks (`enabled`, `lsp`, `ast_grep`, `loop`, `hooks`) are YAML-surface-only: the loader is non-strict (`internal/config/CLAUDE.md` — unknown keys silently ignored), they map to NO `RalphConfig` struct field, and editing them has zero effect.
- The 5 `RalphConfig` struct fields (`MaxIterations`, `AutoConverge`, `HumanReview`, `LintAsInstruction`, `WarnAsInstruction`) are ALL live in code. The prior SPEC's claim that they are "dead" was the blocking factual error.
- Only 2 of the 5 live struct fields are currently exposed in `ralph.yaml` (`lint_as_instruction`, `warn_as_instruction`). The other 3 (`max_iterations`, `auto_converge`, `human_review`) are MISSING from the yaml, and even if present would not reach the engine because `internal/cli/deps.go:99-100` builds the engine from `NewDefaultRalphConfig()` rather than `cfg.Ralph`.

## §B. Context and Background — two senses of "dead" (DO NOT conflate)

The prior SPEC failed plan-audit because it conflated two distinct senses of "dead". This section separates them.

### §B.1 Sense 1 — YAML-surface-dead (TRUE: the 23 inert leaf keys)

`RalphConfig` at `internal/config/types.go:330-343` has EXACTLY 5 fields:

```go
type RalphConfig struct {
    MaxIterations     int  `yaml:"max_iterations"`
    AutoConverge      bool `yaml:"auto_converge"`
    HumanReview       bool `yaml:"human_review"`
    LintAsInstruction bool `yaml:"lint_as_instruction"`
    WarnAsInstruction bool `yaml:"warn_as_instruction"`
}
```

There is NO `Enabled`, `Lsp`, `AstGrep`, `Loop`, `Completion`, or `Hooks` field. The loader is non-strict (`internal/config/CLAUDE.md`: unknown keys silently ignored), so the yaml's `enabled`, `lsp.*`, `ast_grep.*`, `loop.*`, `completion.*`, and `hooks.*` blocks map to nothing. Editing them has zero effect. Removing them from ralph.yaml is safe — no code references them, no struct field is named for them.

**Exact inert-leaf-key count (verified by YAML parse, 2026-08-04):** 23 leaf keys across 5 top-level blocks under `ralph:`:
- `enabled` (1 leaf)
- `lsp`: `auto_start`, `timeout_seconds`, `poll_interval_ms`, `graceful_degradation` (4 leaves)
- `ast_grep`: `enabled`, `config_path`, `security_scan`, `quality_scan`, `auto_fix` (5 leaves)
- `loop`: `max_iterations`, `auto_fix`, `require_confirmation`, `cooldown_seconds` (4 leaves) — NOTE: this `loop.max_iterations` is NOT `RalphConfig.MaxIterations`; the struct field binds to top-level `ralph.max_iterations`, not `ralph.loop.max_iterations`
- `loop.completion`: `zero_errors`, `zero_warnings`, `tests_pass`, `coverage_threshold` (4 leaves)
- `hooks.post_tool_lsp`: `enabled`, `trigger_on`, `severity_threshold` (3 leaves)
- `hooks.stop_loop_controller`: `enabled`, `check_completion` (2 leaves)

This is a documentation/honesty hazard (verification-claim-integrity §1.1 surface 3): a user who tightens `ralph.lsp.enabled` or `ralph.loop.max_iterations` expects the engine's behavior to change; nothing happens.

### §B.2 Sense 2 — struct-field-dead (FALSE: all 5 RalphConfig fields are live)

The prior SPEC's blocking error was claiming the struct fields are dead. They are not. All 5 are read by non-test Go code:

| Field | Read site | Mechanism |
|-------|-----------|-----------|
| `AutoConverge` | `internal/ralph/engine.go:62` | `if e.cfg.AutoConverge {` inside `Decide()` — stagnation detection |
| `HumanReview` | `internal/ralph/engine.go:74` | `if e.cfg.HumanReview && state.Phase == loop.PhaseReview {` inside `Decide()` — human review breakpoint |
| `MaxIterations` | `internal/cli/deps.go:135` | `loop.NewLoopController(..., ralphCfg.MaxIterations)` — iteration ceiling (via the default-config path; see §B.3) |
| `LintAsInstruction` | `internal/hook/post_tool.go:426` | `return cfg.Ralph.LintAsInstruction` — gates LSP-diagnostic-as-instruction injection |
| `WarnAsInstruction` | `internal/hook/post_tool.go:439` | `return cfg.Ralph.WarnAsInstruction` — gates warning-severity inclusion |

Removing ANY of these 5 struct fields WOULD NOT COMPILE (engine.go:62/74 reference `e.cfg.AutoConverge`/`HumanReview` directly). The prior plan's "remove every field except LintAsInstruction/WarnAsInstruction" was wrong and is corrected to: the `RalphConfig` struct is UNCHANGED by this SPEC.

### §B.3 The real root cause — deps.go builds the engine from defaults, not from cfg.Ralph

`internal/cli/deps.go:99-100`:

```go
ralphCfg := config.NewDefaultRalphConfig()           // compiled defaults, NOT user yaml
ralphEngine := ralph.NewRalphEngine(ralphCfg)
```

The user-edited `cfg.Ralph` (loaded faithfully from ralph.yaml at `loader.go:259-266`) is never passed to the engine. So even for the 3 struct fields that ARE in the yaml's schema (`max_iterations`, `auto_converge`, `human_review`), user edits do not reach `Decide()`. Only the two post_tool.go readers (`LintAsInstruction`, `WarnAsInstruction`) consume `cfg.Ralph` directly from the config manager — bypassing the engine.

This is the actual defect. M2 addresses it.

### §B.4 Session.StaleSeconds — dead ralph.yaml pipeline, OWNED by this SPEC (M3)

`ralph.stale_seconds` → `cfg.Session.StaleSeconds` injection pipeline (`types.go:1274` wrapper field + yaml tag → `types.go:571` SessionConfig field → `loader.go:267-269` injection write → `defaults.go:~154` + `defaults.go:~628` defaults). Verified 2026-08-04: `grep -rn "\.StaleSeconds" --include="*.go" internal/ cmd/ pkg/` returns ZERO runtime read sites — only the producer-side injection pipeline (field decl, wrapper, loader write, defaults). The field is a dead injection pipeline at the consumer side.

Ownership resolution (auditor's option a, the cheaper one): THIS SPEC owns the `StaleSeconds` removal because `stale_seconds` is a `ralph.yaml` key (the wrapper at `types.go:1274` is `ralphFileWrapper.StaleSeconds`, bind target `ralph.stale_seconds`). The sibling `SPEC-CONFIG-DEAD-SWEEP-001` explicitly disowns ralph.yaml keys. The `stale_seconds` yaml key is ALREADY ABSENT from the shipped `ralph.yaml`, so no yaml edit is needed — only the dead Go pipeline is removed. The sibling SPEC continues to own the non-ralph partition (`cache.yaml` / `research.yaml` / `state.state_dir`).

Concrete removal list (M3): delete `SessionConfig.StaleSeconds` field (`types.go:569-571`), delete `ralphFileWrapper.StaleSeconds` field + its yaml tag (`types.go:~1274`), delete the injection block (`loader.go:267-269` — the `if wrapper.Ralph.StaleSeconds > 0 { cfg.Session.StaleSeconds = ... }` block), delete the two default assignments (`defaults.go:~154` and `defaults.go:~628`). No `ralph.yaml` edit (key already absent).

## §C. Design Decision (resolved in plan.md §F)

**Recommendation: Option A (wire cfg.Ralph into the engine) — see plan.md §F M2 for the full tradeoff.**

### Option A (wire) — CHOSEN

Change `deps.go:99-100` to load the project config before constructing the engine and pass `cfg.Ralph` to `NewRalphEngine`. Pair with M1's yaml fix (add the 3 missing live keys with defaults matching `NewDefaultRalphConfig`) so users who never edit ralph.yaml get byte-identical behavior. The 5 struct fields become genuinely user-tunable.

### Option B-doc (document advisory-only) — FALLBACK

Leave `deps.go` as-is. Document in ralph.yaml comments that `max_iterations`/`auto_converge`/`human_review` are "advisory — engine uses compiled defaults; only `lint_as_instruction`/`warn_as_instruction` take effect at runtime." Smaller diff, zero behavior change, but leaves the root cause in place.

### Why Option A

1. It is the genuine root-cause fix (makes ralph.yaml effective for all 5 struct fields).
2. Backward-compatible: the 3 newly-added yaml keys carry defaults identical to `NewDefaultRalphConfig()` (`max_iterations: 5`, `auto_converge: true`, `human_review: true`), so users who never edited ralph.yaml observe no change.
3. The `RalphConfig` struct is UNCHANGED under both options — the prior SPEC's field-removal defect is corrected regardless.
4. Option B-doc perpetuates the verification-claim hazard this SPEC exists to fix.

Fallback: if run-phase finds the `InitDependencies` reorder genuinely circular (config load depends on engine, engine depends on config), fall back to Option B-doc for THIS SPEC and open a follow-up SPEC for the wiring. The reorder decision is lead-loaded into plan.md §F M2 as the most reversal-likely step.

## §D. Requirements (GEARS)

### REQ-RCR-001 (Ubiquitous)
The MoAI `ralph.yaml` config file shall expose only keys that map to a real `RalphConfig` struct field consumed by non-test Go code in `internal/`, `cmd/`, or `pkg/`.

### REQ-RCR-002 (Event-detected)
**When** the `RalphConfig` struct at `internal/config/types.go:330-343` is read at `engine.go:62` (`AutoConverge`), `engine.go:74` (`HumanReview`), `deps.go:135` (`MaxIterations`), and `post_tool.go:426,439` (`LintAsInstruction`/`WarnAsInstruction`), the maintainer shall preserve all 5 fields unchanged — removing any of them would break compilation.

### REQ-RCR-003 (Ubiquitous)
The maintainer shall remove the 23 inert leaf keys (the `enabled`, `lsp`, `ast_grep`, `loop`, `hooks` top-level blocks) from `.moai/config/sections/ralph.yaml` and `internal/template/templates/.moai/config/sections/ralph.yaml` because they map to no struct field and editing them has zero effect.

### REQ-RCR-004 (Event-detected)
**When** the 3 live-but-missing keys (`max_iterations`, `auto_converge`, `human_review`) are absent from `ralph.yaml`, the maintainer shall add them with defaults matching `NewDefaultRalphConfig()` (`max_iterations: 5`, `auto_converge: true`, `human_review: true`) so the file honestly documents all 5 struct fields.

### REQ-RCR-005 (Capability gate)
**Where** Option A is chosen, the maintainer shall wire `cfg.Ralph` into `ralph.NewRalphEngine` at `internal/cli/deps.go` (replacing the `NewDefaultRalphConfig()` construction) so user edits to all 5 keys reach the engine; under the Option B-doc fallback, the maintainer shall instead document the advisory-only status in `ralph.yaml` header comments.

### REQ-RCR-006 (Capability gate)
**Where** `lint_as_instruction` and `warn_as_instruction` are read at `internal/hook/post_tool.go:426,439`, the maintainer shall preserve these two keys in `ralph.yaml` and their fields in `RalphConfig`, and their read paths shall remain unchanged.

### REQ-RCR-007 (Event-detected)
**When** the dead `ralph.stale_seconds` → `cfg.Session.StaleSeconds` injection pipeline is verified to have zero runtime consumers (`grep -rn "\.StaleSeconds" --include="*.go" internal/ cmd/ pkg/` returns only the producer-side pipeline), the maintainer shall remove the pipeline in full — delete the `SessionConfig.StaleSeconds` field (`types.go:569-571`), the `ralphFileWrapper.StaleSeconds` field + yaml tag (`types.go:~1274`), the loader injection block (`loader.go:267-269`), and the default entries (`defaults.go:~154` and `defaults.go:~628`). No `ralph.yaml` edit is required because the `stale_seconds` key is already absent from the shipped yaml. This ownership replaces the prior v0.2.0 deferral — `stale_seconds` is a ralph.yaml key, so the removal belongs here, not in the sibling non-ralph sweep.

### REQ-RCR-008 (Capability gate)
**Where** the project is rebuilt after the redesign, `go build ./...` and `go test ./internal/ralph/... ./internal/config/... ./internal/hook/... ./internal/cli/...` shall all exit 0.

### REQ-RCR-009 (Event-detected)
**When** a user runs `/moai loop`, the lint-as-instruction / warn-as-instruction injection behavior shall remain governed by the user-edited `ralph.lint_as_instruction` and `ralph.warn_as_instruction` values — no regression from current behavior.

### REQ-RCR-010 (Capability gate)
**Where** the distributed template is regenerated, `internal/template/templates/.moai/config/sections/ralph.yaml` shall match the local `.moai/config/sections/ralph.yaml` on the retained key set.

## §E. Out of Scope

### Out of Scope — non-ralph dead config

- `cache.yaml`, `research.yaml`, and `state.state_dir` code removal — owned by `SPEC-CONFIG-DEAD-SWEEP-001` (the non-ralph partition). NOTE: `Session.StaleSeconds` is OWNED by THIS SPEC (it is a `ralph.yaml` key — see §B.4 / REQ-RCR-007 / M3), NOT by the sibling sweep.

### Out of Scope — ralph.yaml UI redesign in moai web

- The web settings dropdown for ralph.yaml MAY need to drop the removed inert keys from its schema (`internal/settings/schema_sections.go`), but a deeper UI redesign is out of scope. Only the schema-row deletion (mechanical follow-on) is in scope.

### Out of Scope — RalphConfig struct field removal

- Removing fields from `RalphConfig` is OUT OF SCOPE. All 5 fields are live (engine.go:62/74, deps.go:135, post_tool.go:426/439). Removing any would break compilation. The prior SPEC's field-removal plan is REJECTED.

### Out of Scope — engine behavior changes beyond wiring

- Wiring the 5 struct fields is in scope (Option A). Extending `RalphEngine.Decide()` to consult NEW knobs not present in the struct today (e.g. reading `loop.completion.*` as separate decision inputs) is a behavior-changing feature implementation, NOT part of this honesty fix. A separate follow-up SPEC may propose it.

## §F. Acceptance Criteria Summary

See `acceptance.md` for the full AC matrix (AC-RCR-001 through AC-RCR-009). Every AC maps to a REQ above and is Given-When-Then binary-testable. The prior SPEC's vacuous AC-RCR-002 (which grepped for non-existent fields `RalphConfig.Lsp|AstGrep|...` and matched zero by construction) is replaced by ACs that test REAL behavior: inert-key absence, live-key presence, struct-field preservation, build green, and `/moai loop` regression.

## §G. Risks

- **InitDependencies reorder complexity (Option A):** the engine is currently built at `deps.go:100` before the project config is loaded at `deps.go:144`. Wiring `cfg.Ralph` requires loading config first or lazy-building the engine. Mitigation: plan.md §F M2 lead-loads this as the most reversal-likely decision; if the reorder proves circular, fall back to Option B-doc and open a follow-up SPEC.
- **Behavior shift if defaults diverge:** if the 3 newly-added yaml keys carry defaults DIFFERENT from `NewDefaultRalphConfig()`, users who never edited ralph.yaml would observe a behavior change. Mitigation: REQ-RCR-004 mandates the defaults match (`5` / `true` / `true`), and AC-RCR-008 (`/moai loop` regression guard) catches divergence.
- **Settings-schema drift:** `internal/settings/schema_sections.go` and its testdata carry the full ralph.yaml today; the removed inert keys must be dropped from the schema rows or the web settings test fails. Mitigation: plan.md §F M4 covers the lockstep schema sync.
- **Template-local divergence:** the local and template `ralph.yaml` are byte-identical today (verified); both must shrink and grow identically. Mitigation: AC-RCR-007 enforces key-set parity.

## §H. Cross-References

- **`SPEC-CONFIG-DEAD-SWEEP-001`** — owns the non-ralph half of this sweep (`cache.yaml` / `research.yaml` / `state.state_dir`). Partition: this SPEC touches `ralph.yaml`, the `deps.go` engine-construction site, AND the `ralph.stale_seconds` → `Session.StaleSeconds` pipeline (because `stale_seconds` is a ralph.yaml key — see §B.4); that SPEC touches only the cache/research/state.state_dir surfaces. Run them in separate worktrees or sequence them.
- **`SPEC-CONFIG-KEY-HONESTY-001`** — predecessor; established the "every key must have a consumer" invariant.
- **`SPEC-CONFIG-AUDIT-REPAIR-001`** — landed `loader_gate.go` reading `AstGrepGateConfig` as the live SSOT for ast-grep gate; confirms `ralph.ast_grep.*` was always inert (the gate config lives in `gate.yaml`, not `ralph.yaml`).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — the defect-claim hazard this SPEC addresses (users edit inert keys expecting effect).
- `internal/config/CLAUDE.md` — documents the non-strict loader (unknown YAML keys silently ignored), the mechanism by which the 23 inert leaves survive.
