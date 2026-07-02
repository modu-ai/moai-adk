# Implementation Plan — SPEC-DEAD-CONFIG-001

## §A. Context

Minimal dead-code removal: delete the dead on-disk `runtime.yaml` (both trees) and three dead rows in the
loader-completeness CI-guard allowlist. No production Go code changes. Tier **S**.

> Scope narrowed to `runtime.yaml` only (v0.2.0). `github-actions.yaml` is excluded — it is the live config
> of `SPEC-CI-MULTI-LLM-001` (implemented) and self-removes via `DeprecatedPaths` at v3.0.0. See spec.md §D.

## §B. Headline finding — LoadRuntime test dependency: SAFE-TO-REMOVE (no test fix)

The one risk that could have grown this SPEC beyond Tier S was a test reading the real
`.moai/config/sections/runtime.yaml`. Investigation resolved it as **safe**:

- `TestLoadRuntimeFromFile` (`internal/runtime/budget_test.go:368`) constructs its own YAML string and
  writes it to `filepath.Join(t.TempDir(), "runtime.yaml")`, then calls `LoadRuntime(cfgPath)` on that
  temp path. It never touches the repo's `runtime.yaml`.
- `TestLoadRuntimeMissingFile` (`budget_test.go:410`) calls `LoadRuntime("/nonexistent/path/runtime.yaml")`
  to assert the error path.
- `grep` confirms **no production caller** of `LoadRuntime`, and **no production importer** of
  `internal/runtime`. Production `NewTracker` paths use `DefaultRuntimeConfig()`.

Conclusion: removing `runtime.yaml` from both trees breaks NO test. Tier S holds. The Go surface
(`RuntimeConfig`, `DefaultRuntimeConfig`, `LoadRuntime`) stays untouched per REQ-DC-002.

## §C. Files affected (3 files + 1 build step)

| Action | Path | REQ |
|--------|------|-----|
| Delete | `.moai/config/sections/runtime.yaml` (local) | REQ-DC-001 |
| Delete | `internal/template/templates/.moai/config/sections/runtime.yaml` (template) | REQ-DC-001 |
| Edit | `internal/config/audit_loader_completeness_test.go` (remove 3 rows) | REQ-DC-003, REQ-DC-004 |
| Build | `make build` (regenerate embedded FS after template change) | REQ-DC-006 |

NOT touched (out of scope): `internal/runtime/config.go`, `internal/defs/dirs.go`, `dirs_test.go`,
`.moai/config/sections/github-actions.yaml` (live, excluded).

## §D. Allowlist edit — row-by-row plan

Target: `internal/config/audit_loader_completeness_test.go`.

**`acknowledgedUnloadedSections` (currently lines 15–29) — remove 2 rows, retain 1:**

| Row | Line (approx) | Action | Reason |
|-----|---------------|--------|--------|
| `"gate",` | 17 | REMOVE | No `gate.yaml` in either tree — never iterated by the template-dir guard. |
| `"github-actions",` | 18 | **RETAIN** | `github-actions.yaml` is a live, out-of-scope file (spec.md §D). The row stays. |
| `"memo",` | 20 | REMOVE | No `memo.yaml` in either tree — never iterated. |

Rows kept unchanged: `db`, `github-actions`, `lsp`, `mx`, `observability`, `project`, `security`, `sunset`,
`system` (out of scope; `lsp` deadness noted as a follow-up, NOT removed here).

**`acknowledgedDedicatedLoaders` (currently lines 35–40) — remove 1 row:**

| Row | Line (approx) | Reason |
|-----|---------------|--------|
| `"runtime",` | 38 | After the template `runtime.yaml` is deleted, `runtime` is never iterated → the acknowledgment is dead. |

Rows kept unchanged: `cache`, `harness`, `tool-policy`.

## §E. Milestones (priority-ordered; the ordering is a correctness constraint, not a schedule)

### M1 — Delete dead `runtime.yaml` (both trees) + regenerate embedded FS (REQ-DC-001, REQ-DC-006)

1. `git rm .moai/config/sections/runtime.yaml`
2. `git rm internal/template/templates/.moai/config/sections/runtime.yaml`
3. `make build` — regenerate the embedded binary so the embedded FS drops the deleted template file.

**Intermediate-state greenness:** after M1, the template dir no longer contains `runtime.yaml`, so
`TestAuditLoaderCompleteness` no longer iterates a `runtime` section; the still-present `runtime`
allowlist row is harmlessly unused (dead rows never fail the guard). `gate`/`memo` were already dead. The
tree is green.

### M2 — Remove dead allowlist rows; retain github-actions (REQ-DC-003, REQ-DC-004, REQ-DC-005)

1. Edit `internal/config/audit_loader_completeness_test.go`:
   - Remove `"gate",` and `"memo",` from `acknowledgedUnloadedSections`. **Keep `"github-actions",`.**
   - Remove `"runtime",` from `acknowledgedDedicatedLoaders`.
2. Preserve the surrounding comments/formatting and the sorted-`[]string` convention (REQ-MIG003-013).

**Ordering constraint (REQ-DC-005):** M2's `runtime` row removal MUST NOT precede M1's deletion of the
template `runtime.yaml`. If the `runtime` row were removed while the template file still existed,
`TestAuditLoaderCompleteness` would report `YAML_SECTION_NO_LOADER: runtime`. Doing M1 before M2 (or
bundling both in one commit) keeps every intermediate commit green.

### M3 — Verify (REQ-DC-002, REQ-DC-007)

1. `go test ./internal/config/... ./internal/runtime/... ./internal/defs/...` → all green.
2. `go build ./...` → clean (embedded FS regenerated).
3. Confirm `internal/runtime/config.go`, `internal/defs/dirs.go`, `dirs_test.go`,
   `.moai/config/sections/github-actions.yaml` are untouched.

## §F. Technical approach

- Use `git rm` for the two deletions so the removals are staged atomically with the allowlist edit.
- Template-First (CLAUDE.local.md §2): the template `runtime.yaml` deletion requires `make build` to
  refresh the embedded FS; the local-tree deletion does not.
- Keep the commit surface minimal — one logical unit (deletions + allowlist edit + rebuild).

## §G. Risks and mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| A hidden consumer of `runtime.yaml` | Low | Plan-phase Go grep found no production caller of `LoadRuntime` and no importer of `internal/runtime`; the only callers are `budget_test.go` self-contained fixtures. Non-Go references (release notes) are historical narrative, not runtime consumers. |
| `runtime` row removed before template file → guard fails | Low | REQ-DC-005 ordering constraint; M1 before M2. |
| Accidental removal of the live `github-actions` row | Low | REQ-DC-003 explicitly retains it; AC-DC-003 asserts the row is present post-change. |
| Accidental edit to the `DeprecatedPaths` migration manifest | Low | REQ-DC-007 marks it out of scope; M3 confirms it is untouched; `dirs_test.go` would fail loudly if edited. |
| `make build` skipped → embedded FS stale | Low | M1 step 3 + M3 `go build ./...`. |

## §H. Anti-patterns to avoid

- Do NOT remove or rename the Go `LoadRuntime` / `RuntimeConfig` / `DefaultRuntimeConfig` surface — YAML only.
- Do NOT remove the `github-actions` allowlist row — the file is live and out of scope (spec.md §D).
- Do NOT delete `.moai/config/sections/github-actions.yaml` — excluded; it self-removes via `DeprecatedPaths` at v3.0.0.
- Do NOT edit `internal/defs/dirs.go` `DeprecatedPaths` or `dirs_test.go` — different mechanism, pinned by test.
- Do NOT remove the `lsp`, `db`, `sunset`, or other out-of-scope allowlist rows — scope creep.
- Do NOT introduce a RED commit; every intermediate state stays green (M1 → M2 ordering guarantees it).
