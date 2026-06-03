# Acceptance Criteria — SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001

## §A. Definition of Done

All Blocking AC (AC-MCC-001..009) PASS. Gap A removed across all binding sites; the const+env audit subsystem intact and behaviorally unchanged; Gap C documented with zero logic change; Gap B recorded with zero code change. Template neutrality + mirror-drift + struct-yaml symmetry guards green.

## §B. Given-When-Then Scenarios

### Scenario 1 — Inert MemoryConfig schema is gone, config still loads (Gap A core)
- **Given** the dev project on this branch after implementation, with strict-mode yaml.v3 in `internal/config`,
- **When** `config.Load()` runs against the local `.moai/config/sections/workflow.yaml` (no `memory:` block),
- **Then** the load succeeds with no "unknown key" error, and the symbol `MemoryConfig` no longer appears in non-test Go source.

### Scenario 2 — The real audit subsystem is untouched (Gap A preservation)
- **Given** the package-level const block in `internal/config/defaults.go` (`DefaultMemoryStalenessHours`, `DefaultMemoryIndexLineCap`, `DefaultMemoryStaleAggregateThreshold`) and the `MOAI_MEMORY_AUDIT` env-var path,
- **When** the memory-taxonomy audit subsystem runs (`internal/hook/memo/taxonomy`),
- **Then** its behavior is identical to before this SPEC — its tests pass unchanged, and the const block is still present.

### Scenario 3 — Gap C is documented, not changed
- **Given** a SessionEnd hook resolving the memory directory from session CWD,
- **When** the run-phase agent adds the `resolveMemoryDir` comment + reworded `persist.go` warn + doctrine note,
- **Then** `resolveMemoryDir` / `projectSlug` resolution logic is byte-identical (only comments added) and `resolve_memory_dir_test.go` assertions still pass unchanged.

## §C. Blocking Acceptance Criteria (testable)

### AC-MCC-001 — MemoryConfig fully removed (Gap A) [Blocking]
`grep -rn 'MemoryConfig' internal/ --include='*.go' | grep -v _test` → **0 matches**.
(Covers REQ-MCC-001.)

### AC-MCC-002 — Build + config tests green after removal [Blocking]
`go build ./... && go test ./internal/config/...` → **exit 0, all green**.
(Covers REQ-MCC-002, REQ-MCC-005. Proves strict-yaml load still works and the 3 updated test files pass.)

### AC-MCC-003 — Both YAML copies + embedded have no memory block [Blocking]
After `make build`:
- `grep -c '^\s*memory:' .moai/config/sections/workflow.yaml` → **0**
- `grep -c '^\s*memory:' internal/template/templates/.moai/config/sections/workflow.yaml` → **0**
- `go test ./internal/template/... -run 'TestRuleTemplateMirror|MirrorDrift|StructYAMLSymmetry'` → green (embedded agrees with template).
(Covers REQ-MCC-003.)

### AC-MCC-004 — Const block present AND subsystem intact [Blocking]
- `grep -n 'DefaultMemoryStalenessHours\|DefaultMemoryIndexLineCap\|DefaultMemoryStaleAggregateThreshold' internal/config/defaults.go` → **3 matches present** (const block at 73-76 survives).
- `go test ./internal/hook/memo/taxonomy/...` → **green**.
(Covers REQ-MCC-004.)

### AC-MCC-005 — Template neutrality + mirror-drift green [Blocking]
`go test ./internal/template/... -run 'TestTemplateNeutralityAudit'` and the full `go test ./internal/template/...` → **green**; no internal SPEC-ID / date / commit-SHA leaked into any template-bound file touched by this SPEC.
(Covers REQ-MCC-006.)

### AC-MCC-006a — Gap C comment + warn wording present [Blocking]
- `internal/hook/session_end.go` `resolveMemoryDir` carries a comment stating per-cwd resolution is intentional / aligned with Claude Code per-cwd memory model.
- `internal/hook/handoff/persist.go` warn message (the `os.Stat` miss branch) wording hints at worktree divergence.
(Covers REQ-MCC-007 part 1/2.)

### AC-MCC-006b — Gap C doctrine note present, NO logic change [Blocking]
- A doctrine note documents per-cwd-by-design + L3 `--worktree` Block 0 re-anchoring (placement per plan §F M2 — prefer `.moai/docs/`; if a mirrored `.claude/rules/` file is used, its template mirror is synced in the same commit).
- `go test ./internal/hook/... -run 'ResolveMemoryDir|ProjectSlug'` → **green with unchanged assertions** (no behavioral/logic change; `resolve_memory_dir_test.go` not modified).
(Covers REQ-MCC-007 part 2/2 + the no-logic-change unwanted-behavior clause.)

### AC-MCC-007 — Gap B deferral recorded, no code change [Blocking]
- `spec.md` §B.2 + §E record the Gap B decision and reference the memory doctrine's "Available checks, NOT yet wired" disclosure (`.claude/rules/moai/workflow/moai-memory.md`).
- `git diff --name-only` for the implementation commits shows **no change** under `internal/hook/memo/taxonomy/audit.go` and no change to the PostToolUse audit path (post_tool.go memory-audit invocation).
(Covers REQ-MCC-008.)

### AC-MCC-008 — No byte-cap, no latent-check wiring [Blocking]
- `grep -rn '25.*KB\|25000\|25600\|byteCap\|ByteCap' internal/hook/memo/taxonomy/` → **0 new matches** (no byte-cap added).
- `grep -rn 'AuditIndex\|AuditDuplicates' internal/ --include='*.go' | grep -v _test.go` → still **only the definition sites** in `audit.go` (no new production caller wired).
(Covers REQ-MCC-009, EXCL-MCC-001/002.)

### AC-MCC-009 — Full suite green (regression gate) [Blocking]
`go test ./...` → **exit 0**. No cascading failure introduced by the removal.
(Covers REQ-MCC-002/004/005 cross-package regression.)

## §D. Edge Cases

- **Stale `embedded.go`**: if `make build` is skipped after editing the template YAML, AC-MCC-003 mirror-drift test fails. Run-phase MUST `make build` after the template edit.
- **Symmetry guard for sibling structs**: removal must not accidentally orphan a sibling YAML key — `audit_struct_yaml_symmetry_test.go` is the guard (AC-MCC-002 covers it).
- **Doc-note placement neutrality**: if the doctrine note is placed in a mirrored `.claude/rules/` file without dual-edit, AC-MCC-005 mirror-drift fails. Prefer `.moai/docs/` (plan §F M2).

## §E. Quality Gate Summary

| Gate | Command | Pass condition |
|------|---------|----------------|
| Gap A removal | `grep -rn 'MemoryConfig' internal/ --include='*.go' \| grep -v _test` | 0 matches |
| Build + config | `go build ./... && go test ./internal/config/...` | exit 0 |
| YAML/embedded agreement | `make build` + `go test ./internal/template/...` | green |
| Subsystem intact | `go test ./internal/hook/memo/taxonomy/...` | green |
| Gap C no-logic-change | `go test ./internal/hook/... -run ResolveMemoryDir` | green, assertions unchanged |
| Gap B no code change | `git diff --name-only` excludes audit.go / hook audit path | confirmed |
| Full regression | `go test ./...` | exit 0 |

## §F. AC ↔ REQ Traceability

| AC | REQ |
|----|-----|
| AC-MCC-001 | REQ-MCC-001 |
| AC-MCC-002 | REQ-MCC-002, REQ-MCC-005 |
| AC-MCC-003 | REQ-MCC-003 |
| AC-MCC-004 | REQ-MCC-004 |
| AC-MCC-005 | REQ-MCC-006 |
| AC-MCC-006a | REQ-MCC-007 |
| AC-MCC-006b | REQ-MCC-007 |
| AC-MCC-007 | REQ-MCC-008 |
| AC-MCC-008 | REQ-MCC-009 |
| AC-MCC-009 | REQ-MCC-002/004/005 (cross-package) |
