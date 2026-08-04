# Acceptance — SPEC-CONFIG-DEAD-SWEEP-001

## §D. AC Matrix

| AC ID | REQ | Severity | Summary |
|-------|-----|----------|---------|
| AC-CDS-001 | REQ-CDS-003 | MUST | cache.yaml removed from local + template |
| AC-CDS-002 | REQ-CDS-003 | MUST | LoadCacheConfig removed; CacheConfig struct removed if orphaned |
| AC-CDS-003 | REQ-CDS-004 | MUST | research.yaml removed from local + template |
| AC-CDS-004 | REQ-CDS-004 | MUST | cfg.Research loader + Research field removed |
| AC-CDS-005 | REQ-CDS-005 | MUST | state_dir key + State.StateDir field removed; hardcoded literal unchanged |
| AC-CDS-006 | REQ-CDS-003 | MUST | ValidSessionTTLs still resolves the TTL list for the web seam |
| AC-CDS-007 | REQ-CDS-007 | MUST | go build ./... && go test ./... exit 0 |
| AC-CDS-008 | REQ-CDS-007 | MUST | moai update still merges remaining sections (smoke test) |
| AC-CDS-009 | REQ-CDS-008 | MUST | template-neutrality-check CI green |
| AC-CDS-010 | REQ-CDS-006 | SHOULD | stale comments at loader.go:345 and audit_registry.go:75 corrected |

## §D.1 AC Definitions (Given-When-Then)

### AC-CDS-001 — cache.yaml removed
**Given** the template source tree and the local project,
**When** a search runs for `cache.yaml` under `internal/template/templates/.moai/config/sections/` and `.moai/config/sections/`,
**Then** neither path exists.

### AC-CDS-002 — LoadCacheConfig + CacheConfig removed
**Given** the post-removal source tree,
**When** `grep -rn "LoadCacheConfig\|CacheConfig\b" internal/ cmd/ pkg/ | grep -v _test.go` runs,
**Then** zero matches (excluding the ValidSessionTTLs stub, which MUST NOT reference the CacheConfig type).

### AC-CDS-003 — research.yaml removed
**Given** the template source tree and the local project,
**When** a search runs for `research.yaml` under both config/sections paths,
**Then** neither path exists.

### AC-CDS-004 — Research loader + field removed
**Given** the post-removal source tree,
**When** `grep -rEn "cfg\.Research|researchFileWrapper" internal/ cmd/ pkg/ | grep -v _test.go` runs,
**Then** zero matches. (The `ResearchConfig` type alternative is dropped from the pattern: the type itself MAY remain if referenced elsewhere — a residual type reference is not a death signal for the loader+field removal.)

### AC-CDS-005 — state_dir field removed, literal preserved
**Given** the post-removal source tree,
**When** `grep -n "StateDir" internal/config/types.go internal/config/defaults.go` runs,
**Then** the `State.StateDir` field and its default-population line are absent;
**And when** `grep -n "\.moai/state" internal/cli/state.go internal/worktree/state_guard.go` runs,
**Then** the hardcoded literal `.moai/state` is still present in both files (SSOT preserved);
**And when** `grep -rn "cfg\.State\.StateDir\|\.State\.StateDir" internal/ cmd/ pkg/ | grep -v _test.go` runs,
**Then** zero matches.

### AC-CDS-006 — ValidSessionTTLs still resolves the TTL list
**Given** the post-extraction source tree,
**When** the test that calls `config.ValidSessionTTLs()` runs — located at run-phase via `grep -rn 'ValidSessionTTLs' --include='*_test.go' internal/` (known callers today: `internal/config/cache_config_test.go::TestValidSessionTTLs` and `internal/settings/schema_sections_test.go` which asserts session_ttl option parity against `config.ValidSessionTTLs()`) —
**Then** the test still passes, confirming the TTL select options populate (non-empty list returned by `config.ValidSessionTTLs()`). The pass predicate is the green test run, not a fixed test name.

### AC-CDS-007 — build + test green
**Given** the post-removal source tree,
**When** `go build ./...` runs,
**Then** exit 0;
**And when** `go test ./...` runs,
**Then** exit 0.

### AC-CDS-008 — moai update smoke test
**Given** a `/tmp/test-project` initialized from the rebuilt template,
**When** `moai update` runs against the project,
**Then** exit 0;
**And** the remaining section files (e.g., `quality.yaml`, `language.yaml`, `system.yaml`, `ralph.yaml`) are present and merged correctly (no error about missing `cache.yaml` / `research.yaml`).

### AC-CDS-009 — template-neutrality CI green
**Given** the PR is open against main,
**When** the `template-neutrality-check.yaml` workflow runs,
**Then** exit 0 (deleted template files trivially pass; no new content-class violations introduced).

### AC-CDS-010 — stale comments corrected
**Given** the post-fix source tree,
**When** `sed -n '345p' internal/config/loader.go` runs,
**Then** the line no longer contains "Legacy sub-system" and instead describes `learning` as live (consumed at cli/hook.go:551-1106);
**And when** `sed -n '75p' internal/config/audit_registry.go` runs,
**Then** the line no longer contains "no Go loader yet" and instead reads "partial direct-read".

## §D.2 Severity Definitions

- **MUST** — blocks run-phase completion; failure aborts the run.
- **SHOULD** — fixed in the same PR if cheap; otherwise recorded as `@MX:DEBT` with a follow-up SPEC.

## §D.3 Indirect Verification

- `golangci-lint run` exit code is NOT a load-bearing AC (repo carries baseline warnings); recorded as a sanity check, not a gate.
- The `ValidSessionTTLs` web-settings test may not exist as a named `TestSchema*`; run-phase must locate the actual test that exercises the dropdown (grep `ValidSessionTTLs` call sites in `_test.go`).

## §D.4 Closure Gate

All MUST ACs MUST be PASS. The run-phase §E.2 evidence block MUST cite the verbatim command + output for each AC (verification-claim-integrity.md §2).

## §D.5 Forward-Looking Checks

- After merge, `moai spec audit` MUST classify this SPEC as `era_final: true` (V3R6, 3-phase close) — not grandfather.
- A follow-up SPEC may be opened if the grep at §D.1 AC-CDS-005 reveals a previously-hidden reader of `cfg.State.StateDir` post-merge (defensive; not expected).
