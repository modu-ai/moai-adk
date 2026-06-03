# Implementation Plan — SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001

## §A. Context

Tier S (minimal) cleanup SPEC. Three work classes derived from the audit re-scope:

- **(A) code removal lockstep** — Gap A: remove the inert `config.MemoryConfig` schema across all binding sites + update mirror/assertion tests.
- **(C) doc/comment additions** — Gap C: explanatory comment + warn-wording + doctrine note, NO logic change.
- **(B) decision-record only** — Gap B: already recorded in spec.md §B.2 + §E (REQ-MCC-008/009). No code touched.

Only work class (A) modifies production Go code. (C) is comment/wording/doc. (B) is zero code.

## §B. Known Issues / Hazards

- **[PRIMARY HAZARD] Strict-YAML atomic-removal coupling.** `internal/config` uses `gopkg.in/yaml.v3` strict mode — unknown keys fail loud. If the `memory:` block is removed from the YAML but the `Memory MemoryConfig` struct field remains (or vice versa), `config.Load()` fails OR a struct↔YAML symmetry guard fails. The struct field, both YAML copies, the default block, and `embedded.go` MUST be brought into agreement in the **same milestone** (M1), not spread across commits.
- **Dual-YAML-copy + `make build` coupling.** There are two `workflow.yaml` copies: the template source (`internal/template/templates/.moai/config/sections/workflow.yaml`) and the local deployed copy (`.moai/config/sections/workflow.yaml`). The embedded `internal/template/embedded.go` is regenerated from the template by `make build`. All three must drop the `memory:` block together. Forgetting `make build` leaves `embedded.go` stale → mirror-drift test fails.
- **Struct↔YAML symmetry guard.** `internal/config/audit_struct_yaml_symmetry_test.go` (`TestStructYAMLSymmetry_*`) fails if a Go struct field lacks a matching YAML key or vice versa. After removal, neither side may reference `memory`.
- **Three test files reference `Memory.*`** — `workflow_nested_test.go` (`Memory.IndexLineCap` assertion), `types_test.go` (4 reflection assertions on `Memory.*`), `defaults_test.go` (4 default-value assertions on `cfg.Memory.*`). These must be removed/updated in lockstep with the struct removal or the package won't compile/pass.
- **Template neutrality (Gap C doc placement).** Editing the template-mirrored doctrine files (`session-handoff.md`, `moai-memory.md`) requires byte-parity dual-edit AND no internal SPEC-ID/date/commit-SHA leakage. The Gap C doctrine note SHOULD therefore go in a non-mirrored `.moai/docs/` file (dev-local, neutrality-exempt) to avoid the dual-edit + neutrality coupling. The run-phase agent verifies whether any touched `.claude/rules/` file has a template counterpart before editing it.

## §C. Pre-flight (run-phase entry checks)

1. `grep -rn 'MemoryConfig' internal/ --include='*.go' | grep -v _test.go` → confirm baseline binding sites (types.go:315/364-370, defaults.go:365-370).
2. `grep -rn 'AuditIndex\|AuditDuplicates' internal/ --include='*.go' | grep -v _test.go` → confirm zero non-definition callers (Gap B no-touch baseline).
3. Confirm const block present: `internal/config/defaults.go:73-76` (Gap A: must survive).
4. `go build ./... && go test ./internal/config/... ./internal/hook/...` → green baseline before any edit.

## §D. Constraints

- Behavior-preserving for the audit subsystem (const + env path unchanged) — REQ-MCC-004.
- No new byte-cap, no latent-check wiring — EXCL-MCC-001/002, REQ-MCC-009.
- No `resolveMemoryDir`/`projectSlug` logic change — EXCL-MCC-003, REQ-MCC-007.
- Template neutrality + mirror-drift stay green — REQ-MCC-006.
- Prefer `Edit` over `Write` for existing Go/YAML files; `make build` regenerates `embedded.go` (never hand-edit).

## §E. Self-Verification (run-phase exit, read-only batch)

- `grep -rn 'MemoryConfig' internal/ --include='*.go' | grep -v _test` → 0 matches.
- `grep -rn 'memory:' .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` → 0 `memory:` block matches.
- `go build ./... && go test ./internal/config/...` → green.
- `go test ./internal/hook/memo/taxonomy/...` → green (subsystem intact).
- `go test ./internal/template/...` (neutrality + mirror-drift + struct-yaml symmetry) → green.
- `go test ./internal/hook/... -run ResolveMemoryDir` → unchanged assertions still pass (Gap C no-logic-change proof).

## §F. Milestones

> Tier S — priority-ordered, no time estimates. M1 is the only code-change milestone; M2 is doc-only; M3 is verification + decision-record confirmation.

### M1 — Gap A lockstep removal (the only production code change)
Priority: High. Atomic across all binding sites in one logical unit (strict-yaml hazard).
1. `internal/config/types.go` — remove `MemoryConfig` struct (364-370) AND the `Memory MemoryConfig` field of `WorkflowConfig` (315).
2. `internal/config/defaults.go` — remove the `Memory: MemoryConfig{...}` default block (365-370). **KEEP** the const block (73-76).
3. `internal/template/templates/.moai/config/sections/workflow.yaml` — remove the `memory:` block (142-150). Then `make build` to regenerate `embedded.go`.
4. `.moai/config/sections/workflow.yaml` — remove the local `memory:` block (18-22).
5. Update tests: `workflow_nested_test.go` (~89-90), `types_test.go` (~329-332), `defaults_test.go` (~396/419-421). Verify `audit_loader_completeness_test.go` / `audit_struct_yaml_symmetry_test.go` do not enumerate `Memory`.
6. Add `@MX:NOTE` at the removal site (defaults.go const block) per §D MX target — explain const+env is the real wiring.
Exit: `go build ./... && go test ./internal/config/... ./internal/template/...` green; `MemoryConfig` grep = 0.

### M2 — Gap C document-only additions (no logic change)
Priority: Medium. Comment + wording + doctrine note.
1. `internal/hook/session_end.go` — add comment at `resolveMemoryDir` (184-196): per-cwd intentional, aligned with Claude Code per-cwd memory model.
2. `internal/hook/handoff/persist.go` — reword the `os.Stat` miss warn (123-127) to hint at worktree divergence.
3. Add a short doctrine note documenting per-cwd-by-design + L3 `--worktree` Block 0 re-anchoring. **Placement decision**: prefer a non-mirrored `.moai/docs/` file (dev-local, neutrality-exempt) to avoid template byte-parity + neutrality coupling. If a `.claude/rules/` file is chosen instead, verify+sync its `internal/template/templates/` mirror in the same commit.
Exit: `go test ./internal/hook/... -run ResolveMemoryDir` unchanged-pass; `go build ./...` green.

### M3 — Verification + Gap B decision-record confirmation
Priority: Medium. Read-only batch (§E) + confirm Gap B requires zero code.
1. Run the §E self-verification batch.
2. Confirm Gap B is fully captured in spec.md §B.2 + §E (REQ-MCC-008/009) and that `audit.go` / the hook audit path are untouched.
Exit: all §E checks green; no diff under `internal/hook/memo/taxonomy/`.

## §G. Anti-Patterns to Avoid

- Removing the YAML block but leaving the struct field (or vice versa) → strict-yaml / symmetry-guard failure. Lockstep in M1.
- Hand-editing `embedded.go` → always `make build`.
- Removing the const block (73-76) thinking it is part of the dead schema → it is the REAL wiring (Gap A confusion case). Keep it.
- Editing template-mirrored doctrine for the Gap C note without dual-edit → mirror-drift failure. Prefer `.moai/docs/`.
- Adding a byte-cap or wiring latent checks "while here" → scope-discipline violation (EXCL-MCC-001/002).
- Changing `resolveMemoryDir` logic → regression vs Claude Code per-cwd model (EXCL-MCC-003).

## §H. Recommended cycle_type for run-phase: **ddd**

Rationale: This is a **behavior-preserving removal**, not new functionality. The DDD ANALYZE-PRESERVE-IMPROVE cycle fits precisely:
- ANALYZE: the binding-site inventory + caller analysis is already done (spec §B.1) and re-verified in §C pre-flight.
- PRESERVE: the existing config + hook tests are the characterization safety net. The audit subsystem (const+env) must behave identically (REQ-MCC-004) — the existing `internal/hook/memo/taxonomy` tests + `resolve_memory_dir_test.go` are the unchanged-behavior contract.
- IMPROVE: remove the dead struct; the test green-ness before/after proves no behavior changed.

A failing-test-first (tdd) approach is NOT warranted: there is no new behavior to specify with a RED test. The "test" for a removal is that the existing suite stays green after the dead code is gone (characterization), which is exactly DDD's PRESERVE phase. The only test edits (M1 step 5) are deletions/updates of assertions on the removed symbol, not new RED specs.

## §I. Cross-References
- spec.md §B (verdicts), §C (REQ-MCC-001..009), §E (exclusions).
- acceptance.md (AC-MCC-001..009).
