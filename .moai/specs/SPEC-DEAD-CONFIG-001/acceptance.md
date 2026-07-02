# Acceptance Criteria — SPEC-DEAD-CONFIG-001

All criteria are mechanical (filesystem state, grep, `go test`, `go build`). Each maps to a REQ.

## §A. Mechanical acceptance criteria

### AC-DC-001 — `runtime.yaml` absent in BOTH trees (REQ-DC-001)

```bash
test ! -e .moai/config/sections/runtime.yaml \
  && test ! -e internal/template/templates/.moai/config/sections/runtime.yaml \
  && echo BOTH-ABSENT || echo STILL-PRESENT
```
Expected: `BOTH-ABSENT`.

### AC-DC-002 — runtime Go surface preserved (REQ-DC-002)

```bash
grep -c 'func LoadRuntime\|func DefaultRuntimeConfig\|type RuntimeConfig' internal/runtime/config.go
```
Expected: `3` (all three symbols still present; `internal/runtime/config.go` unchanged).

### AC-DC-003 — dead `gate`/`memo` rows removed AND `github-actions` retained (REQ-DC-003)

```bash
# gate + memo removed from the allowlist:
grep -E '^\s*"(gate|memo)",' internal/config/audit_loader_completeness_test.go | wc -l   # expect 0
# github-actions RETAINED (live, out-of-scope file):
grep -E '^\s*"github-actions",' internal/config/audit_loader_completeness_test.go | wc -l # expect 1
```
Expected: `0` for gate/memo; `1` for github-actions.

### AC-DC-004 — dead `runtime` dedicated-loader row removed (REQ-DC-004)

```bash
grep -E '^\s*"runtime",' internal/config/audit_loader_completeness_test.go | wc -l
```
Expected: `0`.

### AC-DC-005 — loader-completeness guard still green (REQ-DC-003, REQ-DC-004, REQ-DC-005)

```bash
go test ./internal/config/ -run TestAuditLoaderCompleteness -count=1
```
Expected: `ok` — no `YAML_SECTION_NO_LOADER: runtime` (or any) failure.

### AC-DC-006 — runtime tests still green after YAML removal (REQ-DC-002)

```bash
go test ./internal/runtime/ -run 'TestLoadRuntime' -count=1
```
Expected: `ok` — both `TestLoadRuntimeFromFile` and `TestLoadRuntimeMissingFile` pass (they use
self-contained `t.TempDir()` fixtures, not the removed on-disk file).

### AC-DC-007 — full affected-package suite green (REQ-DC-002, REQ-DC-003, REQ-DC-004)

```bash
go test ./internal/config/... ./internal/runtime/... ./internal/defs/... -count=1
```
Expected: `ok` for every package.

### AC-DC-008 — embedded FS regenerated, build clean (REQ-DC-006)

```bash
make build && go build ./...
```
Expected: build succeeds; the embedded template FS no longer contains `runtime.yaml`.

### AC-DC-009 — migration manifest + live github-actions untouched (REQ-DC-007)

```bash
git status --porcelain | grep -E 'internal/defs/dirs\.go|internal/defs/dirs_test\.go|config/sections/github-actions\.yaml' | wc -l
```
Expected: `0` (the manifest, its test, and the live github-actions.yaml are all unmodified/undeleted).

```bash
go test ./internal/defs/ -run TestDeprecatedPaths -count=1
```
Expected: `ok` (Category B enumeration — still lists `github-actions.yaml`/`gate.yaml`/`memo.yaml` for
user-project cleanup — remains intact).

## §B. Given-When-Then scenarios

### Scenario 1 — Loader-completeness guard after runtime.yaml removal

- **Given** `internal/template/templates/.moai/config/sections/runtime.yaml` has been deleted and the
  `runtime` row removed from `acknowledgedDedicatedLoaders`,
- **When** `TestAuditLoaderCompleteness` iterates the template sections directory,
- **Then** no `runtime` section is encountered, no `YAML_SECTION_NO_LOADER` error is emitted, and the
  test passes.

### Scenario 2 — LoadRuntime remains functional without the on-disk file

- **Given** both `runtime.yaml` files are removed but `internal/runtime/config.go` is unchanged,
- **When** `TestLoadRuntimeFromFile` writes its own YAML to a `t.TempDir()` path and calls
  `LoadRuntime` on it,
- **Then** parsing succeeds and the assertions pass — proving the Go surface is decoupled from the
  removed on-disk YAML.

### Scenario 3 — Live github-actions.yaml is preserved, its allowlist row retained

- **Given** `github-actions.yaml` is the live config of `SPEC-CI-MULTI-LLM-001` and is excluded from
  this SPEC,
- **When** the allowlist edit removes only `gate` / `memo` / `runtime`,
- **Then** `.moai/config/sections/github-actions.yaml` still exists on disk AND the `github-actions` row
  remains in `acknowledgedUnloadedSections` (it self-removes later via `DeprecatedPaths` at v3.0.0).

### Scenario 4 — Intermediate commit greenness (ordering)

- **Given** M1 (runtime.yaml deletions + `make build`) is committed before M2 (allowlist edit),
- **When** the loader-completeness guard runs on the M1 commit,
- **Then** it is green — the still-present `runtime` allowlist row is harmlessly unused, and no section
  is uncovered.

## §C. Edge cases

- The `runtime` allowlist row must NOT be removed while the template `runtime.yaml` still exists — doing
  so mid-change produces `YAML_SECTION_NO_LOADER: runtime`. Enforced by M1→M2 ordering (REQ-DC-005).
- `gate` / `memo` were already dead (no file in either tree); removing their rows changes no guard result.
- `github-actions.yaml` is deliberately NOT deleted and its row is deliberately KEPT — removing either
  would either break the live CI-MULTI-LLM config or require reconciling docs-site ×4 references.

## §D. Definition of Done

- [ ] AC-DC-001 … AC-DC-009 all pass.
- [ ] All four Given-When-Then scenarios verified.
- [ ] `go test ./internal/config/... ./internal/runtime/... ./internal/defs/...` green.
- [ ] `make build && go build ./...` clean.
- [ ] `internal/runtime/config.go`, `internal/defs/dirs.go`, `dirs_test.go`,
      `.moai/config/sections/github-actions.yaml` confirmed unmodified/undeleted.
- [ ] `github-actions` allowlist row confirmed retained.
- [ ] No RED intermediate commit (M1 → M2 ordering preserved).
- [ ] No out-of-scope YAML or allowlist row removed (scope discipline held).
