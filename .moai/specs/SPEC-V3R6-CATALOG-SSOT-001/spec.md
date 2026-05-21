---
id: SPEC-V3R6-CATALOG-SSOT-001
title: "Catalog Single Source of Truth — hash regen + Makefile + CI guard"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: GOOS Kim
priority: P0
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tier: S
tags: "catalog, manifest, hash, ci-gate, ssot, v3, wave1"
depends_on: ["SPEC-V3R6-HARNESS-LEARNER-FIX-001"]
---

# SPEC-V3R6-CATALOG-SSOT-001 — Catalog Single Source of Truth

## 1. Problem Statement

`internal/template/catalog.yaml` is the authoritative manifest for 52 embedded skills/agents shipped with the `moai` binary. Each entry carries a stored `hash` field that must match the sha256 (after CRLF→LF + trailing-whitespace + single-newline normalization) of the source `SKILL.md` / agent `.md` file. The audit test `TestManifestHashFormat` (`catalog_tier_audit_test.go:332`) enforces this equality and currently FAILS on `main` (HEAD `b1b1ec8be`) with two `CATALOG_HASH_UNSTABLE` diagnostics:

- `moai-harness-learner`: stored `14a2df169edbf99418b051196fc8a4d6da468bfbba14d4a0a43e7a6e883a147c`, computed `53fa7251d9068594801e4dced3b8de3e12ea569aeb764c48d82d63b4fefe0b3c`
- `moai-meta-harness`: stored `00ea090d6bcaabe3a85d50e055865ec34d02d53e9563ed55a3fd11ba757813cf`, computed `e3bf9e8ecb93781c5a0a3b244f0c584f9f8810eeaf9c47c2a42b2a39f1a33a7f`

Root cause: SPEC-V3R6-HARNESS-LEARNER-FIX-001 (predecessor SPEC, fully landed in commits `c8f42153f` + `b92e89675` + `b1b1ec8be`) modified `moai-harness-learner/SKILL.md` in both local and template paths to satisfy subagent boundary discipline. The manifest hash for `moai-harness-learner` was therefore intentionally left stale by the predecessor SPEC to motivate this work. The drift on `moai-meta-harness` is parallel: its template `SKILL.md` was also touched in the same template-first corrective.

Two structural problems compound the drift:

1. **No build-step refresh**: `make build` invokes only `go build` and never calls `internal/template/scripts/gen-catalog-hashes.go`. Hash regeneration is an out-of-band manual step (`go run internal/template/scripts/gen-catalog-hashes.go --all`). Future SKILL.md edits will silently produce a divergent manifest until a developer remembers to run the script or until CI fails late.
2. **No documented CI guard contract**: `TestManifestHashFormat` is BLOCKING (a Go test `t.Errorf` produces a non-zero exit), but no top-level rule names it as the catalog SSoT guard, leaving its status discoverable only by code reading.

The goal of this SPEC is to (a) restore manifest integrity by regenerating the two stale entries, (b) make `make build` self-healing so drift cannot be merged unnoticed, and (c) document `TestManifestHashFormat` as the binding CI gate. The SPEC explicitly does NOT extend the gating mechanism to hooks, schema, or template-mirror coverage (see § 5).

## 2. Goals

- G1 — Replace the two stale hash entries with the computed values produced by `internal/template/scripts/gen-catalog-hashes.go --all` against the current template tree.
- G2 — Wire `make build` so that the hash refresh script runs before the Go build, eliminating manual-step drift.
- G3 — Document `TestManifestHashFormat` as the binding catalog SSoT CI guard inside the existing `catalog_doc.md` artifact, so the gate is discoverable without code-reading.
- G4 — Keep the existing `gen-catalog-hashes.go` script API and the `catalog.yaml` schema unchanged (zero refactor surface).

## 3. EARS Requirements

- **REQ-CSS-001 (Ubiquitous)** — The `internal/template/catalog.yaml` file SHALL contain, for every entry, a `hash` field whose value equals the normalized sha256 hex of the entry's source file as computed by `catalog_hash_norm.NormalizeForHash` followed by `sha256.Sum256`.

- **REQ-CSS-002 (State-Driven)** — WHILE `moai-harness-learner` and `moai-meta-harness` carry stale stored hashes (`14a2df…` and `00ea090d…`), the implementation SHALL replace them with `53fa7251d9068594801e4dced3b8de3e12ea569aeb764c48d82d63b4fefe0b3c` and `e3bf9e8ecb93781c5a0a3b244f0c584f9f8810eeaf9c47c2a42b2a39f1a33a7f` respectively, leaving all other 50 entries byte-identical.

- **REQ-CSS-003 (Event-Driven)** — WHEN a developer runs `make build`, the Makefile SHALL invoke `go run ./internal/template/scripts/gen-catalog-hashes.go --all` BEFORE the `go build` step, so any uncommitted hash drift is reflected in the working tree prior to compilation.

- **REQ-CSS-004 (Unwanted)** — The Makefile change SHALL NOT modify the `install`, `release-local`, `test`, `lint`, `fix`, `vet`, `fmt`, `generate`, `clean`, `tidy`, `constitution-check`, `run`, `help`, `ci-local`, `pr-merge`, `ci-disable`, `verify-required-checks`, `tui-snapshot`, `tui-snapshot-verify`, `preflight`, `lint-fast`, or `test-race-short` targets, nor SHALL it touch the `LDFLAGS`, `BINARY_NAME`, `MODULE`, `VERSION`, `COMMIT`, `DATE`, `LOCAL_RELEASE_DIR`, `PLATFORM`, or `RELEASE_BINARY` variables.

- **REQ-CSS-005 (Ubiquitous)** — The `internal/template/catalog_doc.md` artifact SHALL document `TestManifestHashFormat` as the binding catalog SSoT CI guard, naming the exact failure mode (`CATALOG_HASH_UNSTABLE`) and stating that the gate is BLOCKING (non-zero exit on drift).

- **REQ-CSS-006 (Optional)** — WHERE the developer wants to validate hashes without writing to `catalog.yaml`, the Makefile target SHOULD accept a `--dry-run` invocation path via documentation in `catalog_doc.md` (no new make target required; `go run … --dry-run` is documented).

## 4. Binary Acceptance Criteria

| AC ID | Verification Command (run from project root) | Pass Condition |
|-------|----------------------------------------------|----------------|
| AC-CSS-001 | `go test -count=1 -run TestManifestHashFormat ./internal/template/...` | Exit code 0; stdout contains `PASS`; stdout does NOT contain `CATALOG_HASH_UNSTABLE` |
| AC-CSS-002 | `grep -c "hash: 53fa7251d9068594801e4dced3b8de3e12ea569aeb764c48d82d63b4fefe0b3c" internal/template/catalog.yaml` | Output equals exactly `1` |
| AC-CSS-003 | `grep -c "hash: e3bf9e8ecb93781c5a0a3b244f0c584f9f8810eeaf9c47c2a42b2a39f1a33a7f" internal/template/catalog.yaml` | Output equals exactly `1` |
| AC-CSS-004 | `grep -c "hash:" internal/template/catalog.yaml` | Output equals exactly `52` (no entries lost or added) |
| AC-CSS-005 | `grep -E "gen-catalog-hashes\.go" Makefile` | Exit code 0; at least one matching line exists in the recipe for the `build` target |
| AC-CSS-006 | `go test -count=1 ./internal/template/...` | Exit code 0 (full template package suite passes, no regressions) |
| AC-CSS-007 | `grep -c "TestManifestHashFormat" internal/template/catalog_doc.md` AND `grep -c "CATALOG_HASH_UNSTABLE" internal/template/catalog_doc.md` | Both outputs ≥ `1` (test name AND failure-mode literal appear in body) |

## 5. Out of Scope

### 5.1 Out of Scope

The following items are explicitly deferred to separate future SPECs:

- **EXCL-CSS-001** — PostToolUse hook auto-updating `catalog.yaml` after every `SKILL.md` / agent `.md` write. Tracked as future SPEC `CATALOG-CI-HOOK-001` (not created in this SPEC). Rationale: hook design crosses subagent boundary (`.claude/hooks/`) and needs separate scope.
- **EXCL-CSS-002** — Template-mirror drift fix exemplified by `TestLateBranchTemplateMirror`. Different mechanism (file-identity check across `internal/template/templates/` ↔ `.claude/`), separate from hash audit, requires its own SPEC.
- **EXCL-CSS-003** — Refactor or feature addition to `gen-catalog-hashes.go` (e.g., comment-preserving YAML round-trip, parallel hashing, JSON output). Tool is functionally adequate as-is for the 52-entry catalog; any improvement is a follow-up.
- **EXCL-CSS-004** — Schema changes to `catalog.yaml` (e.g., adding marketplace tier reserved fields, version bumps, or new tier sections). The current schema is sufficient.
- **EXCL-CSS-005** — Modifying or pruning the predecessor SPEC's working-tree byproducts (5 modified + 7 untracked PRESERVE list in plan.md). Out of scope for catalog SSoT work.
- **EXCL-CSS-006** — Re-hashing entries OTHER than the two named stale entries. The current `go test` baseline confirms only two diagnostics; all 50 other entries are known-good and MUST NOT be touched.
- **EXCL-CSS-007** — Extending CI guard semantics beyond `TestManifestHashFormat` (e.g., adding a separate spec-lint guard, a GitHub Actions check, or a pre-commit hook). The Go test layer is the canonical gate; no new gates are added.
- **EXCL-CSS-008** — Cross-package catalog integrity (e.g., enforcing `catalog.yaml` ↔ `internal/manifest/types.go` parity). The audit test only checks file content sha256; structural parity is a separate concern.

## 6. Risks

| Risk ID | Severity | Description | Mitigation |
|---------|----------|-------------|------------|
| R-CSS-001 | Medium | `go run internal/template/scripts/gen-catalog-hashes.go --all` rewrites `catalog.yaml` via `yaml.Marshal`, which the script's own comment (line 254) warns "does NOT preserve comments." If catalog.yaml currently carries inline comments, they will be lost on script invocation, producing a noisy diff. | Pre-flight inspection of `catalog.yaml` confirms no inline comments are present (only data lines + structural whitespace). The script's `--entry NAME` mode hashes a single entry but still writes the full file via Marshal, so to minimize blast radius, the implementation MAY use targeted `Edit` operations on just the two stale `hash:` lines instead of invoking the script for the corrective. The script invocation in `make build` then becomes the long-term safeguard. |
| R-CSS-002 | Medium | Adding a `go run` step to `make build` introduces a transitive dependency on `gopkg.in/yaml.v3` resolution at build time. If the module cache is empty (fresh CI runner), `go run` will trigger a download. | This is acceptable: `go run` already participates in the module cache that `go build` uses immediately after. The download cost is amortized into the first build and is bounded by the existing `go.sum` integrity check. No new dependency is introduced (yaml.v3 already imported by other internal packages). |
| R-CSS-003 | Low | `make build` recipes commonly assume idempotency; running `go run … --all` will modify `catalog.yaml` on disk if the working tree is out-of-sync, causing `git status` to show diffs after a "clean" build. | Document this behavior explicitly in `catalog_doc.md` (REQ-CSS-005). Developers can use `--dry-run` (REQ-CSS-006) to validate without writes. The behavior is intentional — drift detection at build time is the point. |
| R-CSS-004 | Low | The two computed hashes baked into REQ-CSS-002 are tied to the current normalized content of the two SKILL.md files at HEAD `b1b1ec8be`. If the predecessor SPEC's templates change again between plan and run, REQ-CSS-002 values will be stale. | Mitigation: AC-CSS-001 (`TestManifestHashFormat` exit 0) is the binding gate, not the hard-coded hash values. Implementation MUST re-derive the computed hashes by running the script (or by reading `go test` diagnostic output) at run-phase start. The values in REQ-CSS-002 are diagnostic, not normative. |
| R-CSS-005 | Low | A future Makefile `.PHONY` target list maintenance pass could drop the catalog refresh from `build:`. | Document the dependency in `catalog_doc.md` (REQ-CSS-005). The AC-CSS-005 grep check, run as part of `go test ./internal/template/...`, would not catch a Makefile regression — but a future SPEC can add a dedicated `Makefile` lint test if drift recurs. Out of scope here. |

## 7. REQ ↔ AC Traceability

| Requirement | Acceptance Criterion | Notes |
|-------------|----------------------|-------|
| REQ-CSS-001 | AC-CSS-001, AC-CSS-004 | Test enforces the universal invariant + count |
| REQ-CSS-002 | AC-CSS-002, AC-CSS-003 | Two specific hash values appear in catalog.yaml |
| REQ-CSS-003 | AC-CSS-005 | Makefile recipe contains script reference |
| REQ-CSS-004 | AC-CSS-006 | Full template suite green = no collateral damage on other targets |
| REQ-CSS-005 | AC-CSS-007 | catalog_doc.md mentions the test name + failure-mode literal |
| REQ-CSS-006 | (documentation in catalog_doc.md via AC-CSS-007) | Optional; covered by the catalog_doc.md content requirement |

---

**SPEC tier**: S (LEAN minimal — 2 artifacts: spec.md + plan.md; AC inline in §4).
**Workflow**: Late-Branch per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005 — commits stay on main until PR creation.
**Predecessor evidence**: SPEC-V3R6-HARNESS-LEARNER-FIX-001 commits `c8f42153f` + `b92e89675` + `b1b1ec8be` produced the hash drift this SPEC addresses.
**Canonical catalog**: `internal/template/catalog.yaml` (52 entries) embedded via `go:embed` and loaded by `catalog_loader.LoadEmbeddedCatalog()`.
