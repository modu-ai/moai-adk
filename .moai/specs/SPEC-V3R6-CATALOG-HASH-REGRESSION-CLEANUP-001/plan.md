---
id: SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001
title: "Catalog Hash Regression Cleanup — Plan"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.7.0"
module: "internal/template/catalog.yaml + internal/template/scripts/gen-catalog-hashes.go + internal/spec"
lifecycle: spec-anchored
tags: "catalog-hash, regression, cleanup, sprint-10-lane-b, tier-s-minimal, drift-prevention"
tier: S
depends_on: [SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001, SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001]
---

# Implementation Plan: Catalog Hash Regression Cleanup

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial plan-phase artifact. Tier S minimal LEAN, single M1 milestone, 1-pass run-phase expected. Phase 0.5 SKIP-eligible candidate per CONST-V3R5-026 (manager-spec self-projection ≥ 0.90). |

---

## §A — Milestone Breakdown

### M1: Defensive Verification + Cosmetic Refresh (single 1-pass milestone)

**Scope**: 3 deliverables in one atomic commit (per Tier S 1-pass discipline).

| Deliverable | File | Operation | LOC estimate |
|-------------|------|-----------|--------------|
| D1 | `internal/spec/catalog_hash_test.go` (NEW) OR `internal/template/catalog_hash_test.go` (NEW) | CREATE — Go test that loads catalog.yaml, recomputes 38 entry hashes, asserts parity | ~100-150 LOC (test fixture + walker + assertion + 1 sub-test for missing-templates if Variant A per AC-CHR-006) |
| D2 | `internal/template/catalog.yaml` | MODIFY — single-line `generated_at:` field update to current ISO-8601 UTC | 1 line changed |
| D3 (optional) | `internal/template/normalize_hash.go` (NEW, if §E.1 Variant A chosen) | CREATE — extract `normalizeForHash` as exported helper; update `gen-catalog-hashes.go` to import it | ~30-50 LOC |

**Decision**: D3 (extraction) is manager-develop's call at run-phase. If extraction is chosen, the generator's existing inline `normalizeForHash` is deleted and replaced with the imported helper — verify no behavior change via the existing dry-run output unchanged.

**Acceptance gating**: M1 completes when all 7 AC-CHR (AC-CHR-001 through AC-CHR-007) PASS independently via their §3 verification commands.

**No M2, M3, etc.** — Tier S minimal scope. If unexpected complexity emerges (e.g., generator coupling, test fixture explosion), manager-develop returns a blocker report with proposed M2 split rather than expanding M1 silently.

---

## §B — File-by-File Diff Plan

### B.1 NEW file: `internal/spec/catalog_hash_test.go` (or `internal/template/catalog_hash_test.go`)

**Package decision**: `internal/spec/` is the SHOULD per §D.2 of spec.md (alongside `drift_*.go` siblings). manager-develop MAY override if the import boundary (test needs to read `internal/template/catalog.yaml` + `internal/template/templates/**`) makes `internal/template/` cleaner. Either is acceptable.

**Skeleton structure** (illustrative, not prescriptive — manager-develop owns implementation choices):

```go
package spec_test  // or internal/template

import (
    "crypto/sha256"
    "encoding/hex"
    "os"
    "path/filepath"
    "testing"

    "gopkg.in/yaml.v3"
)

func TestCatalogHashParity(t *testing.T) {
    // Load catalog.yaml from a relative path stable across test contexts
    catalogPath := findRepoRelativePath(t, "internal/template/catalog.yaml")
    raw, err := os.ReadFile(catalogPath)
    if err != nil {
        t.Fatalf("read catalog.yaml: %v", err)
    }
    var cat catalogFile // mirror struct of gen-catalog-hashes.go
    if err := yaml.Unmarshal(raw, &cat); err != nil {
        t.Fatalf("parse catalog.yaml: %v", err)
    }
    entries := allEntries(&cat)
    templatesDir := findRepoRelativePath(t, "internal/template/templates")
    drift := 0
    for _, e := range entries {
        sourcePath, err := resolveHashSourcePath(templatesDir, e.Path)
        if err != nil {
            t.Errorf("entry %q: %v", e.Name, err)
            continue
        }
        body, err := os.ReadFile(sourcePath)
        if err != nil {
            t.Errorf("entry %q read source: %v", e.Name, err)
            continue
        }
        normalized := normalizeForHash(body)
        h := sha256.Sum256(normalized)
        computed := hex.EncodeToString(h[:])
        if computed != e.Hash {
            t.Errorf("entry %q hash drift: stored=%s computed=%s", e.Name, e.Hash, computed)
            drift++
        }
    }
    if drift == 0 {
        t.Logf("verified %d catalog entries against normalized source bodies — 0 drift", len(entries))
    }
}
```

**Risk**: The above duplicates `catalogFile`, `allEntries`, `resolveHashSourcePath`, `normalizeForHash` from the `main` package of `gen-catalog-hashes.go`. Since `gen-catalog-hashes.go` is a standalone script (`package main`), its types cannot be imported. manager-develop has two options:

- **Option α**: Duplicate the types and helpers in the test file with a doc comment `// MIRROR of gen-catalog-hashes.go — keep in sync; SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001 §E.1`
- **Option β**: Extract the shared logic to a new `internal/template/catalog_hash.go` exported package, import from both `gen-catalog-hashes.go` and `catalog_hash_test.go`

Option β is structurally superior but expands scope (touches `gen-catalog-hashes.go` to switch from inline to imported helpers). Option α is simpler and keeps the SPEC scope tight. Tier S preference: **Option α**. If manager-develop discovers Option β is necessary (e.g., compile failures, package boundary violations), it's an in-scope sub-decision recorded in the run-phase commit message.

### B.2 MODIFIED file: `internal/template/catalog.yaml`

**Change**: line 2 only

```diff
 version: 1.0.0
-generated_at: "2026-05-12T03:00:00Z"
+generated_at: "2026-05-25THH:MM:SSZ"
 catalog:
```

Where `HH:MM:SS` is the actual UTC time of the run-phase commit. Use `date -u +"%Y-%m-%dT%H:%M:%SZ"` to populate at commit time.

**Verification**: AC-CHR-002 grep regex `\"2026-05-2[5-9]T[0-9]{2}:[0-9]{2}:[0-9]{2}Z\"` matches.

### B.3 OPTIONAL NEW file: `internal/template/catalog_hash.go` (Option β only)

Only if manager-develop chooses §B.1 Option β. Contains exported `NormalizeForHash`, `ComputeHash`, `ResolveHashSourcePath`, `AllEntries`, `CatalogFile` types. `gen-catalog-hashes.go` then imports `internal/template` and replaces inline helpers with `template.NormalizeForHash(...)` etc.

**If Option β chosen**: 1 additional NEW file (~80-100 LOC) + ~5-line diff in `gen-catalog-hashes.go` (deletion of inline helpers + addition of imports). Total scope still Tier S.

---

## §C — Verification Batch (Orchestrator-side, parallel multi-Bash, 7 commands)

Per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution, the orchestrator MUST issue these 7 verification commands as a single-turn multi-Bash batch after manager-develop reports run-phase completion.

```bash
# 1. Full Go test suite (regression sweep)
go test ./...

# 2. New catalog hash parity test specifically (AC-CHR-004, AC-CHR-005)
go test ./internal/spec/... -run TestCatalogHashParity -v -count=1

# 3. Hash drift dry-run comparison (AC-CHR-001)
go run internal/template/scripts/gen-catalog-hashes.go --all --dry-run 2>&1 | tail -3

# 4. generated_at refresh verification (AC-CHR-002)
grep -E "^generated_at:" internal/template/catalog.yaml

# 5. Test file existence + build (AC-CHR-003)
ls internal/spec/catalog_hash_test.go internal/template/catalog_hash_test.go 2>/dev/null; go vet ./internal/spec/... ./internal/template/...

# 6. Test mutation guard (AC-CHR-007)
git status --porcelain > /tmp/pre.txt; go test ./internal/spec/... -run TestCatalogHashParity -count=3; git status --porcelain > /tmp/post.txt; diff /tmp/pre.txt /tmp/post.txt

# 7. Lint baseline (no regression from sibling lint rules)
golangci-lint run --timeout=2m ./internal/spec/... ./internal/template/...
```

**Expected outcome**: All 7 commands exit 0 (or produce expected positive matches per AC verification commands in spec.md §3). Any non-zero exit triggers run-phase re-iteration per the standard run-phase verification protocol.

---

## §D — PRESERVE List (HARD — DO NOT modify)

The following files MUST be untouched by run-phase activity for this SPEC. Any modification triggers a HARD violation report.

### D.1 Predecessor SPEC bodies

- `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001/{spec.md,plan.md,acceptance.md,progress.md}` — body PRESERVED per L48 SSOT discipline
- `.moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/{spec.md,plan.md,acceptance.md,progress.md}` — body PRESERVED per L48 SSOT discipline

### D.2 Catalog hash field values (38 entries)

- `internal/template/catalog.yaml` `hash:` fields for all 38 entries MUST NOT be edited. Only the top-level `generated_at:` line is modified per §B.2.

### D.3 Catalog schema and structure

- `internal/template/catalog.yaml` `version:` field, top-level YAML structure (`catalog.core.{skills,agents,rules}` + `catalog.optional.*`), per-entry field set (`name:`, `tier:`, `path:`, `hash:`, `version:`) MUST NOT be altered.

### D.4 Generator behavior

- `internal/template/scripts/gen-catalog-hashes.go` MUST NOT be refactored, except possibly:
  - **Allowed under Option β only** (per §B.3): replace inline `normalizeForHash`, `computeHash`, `resolveHashSourcePath`, `allEntries`, `catalogFile` types with imports from new `internal/template/catalog_hash.go`. Behavior MUST be byte-identical (verified via dry-run output unchanged before/after).
  - **Forbidden**: adding auto-`generated_at` update, adding `--check` mode, adding YAML comment preservation logic.

### D.5 Sibling lint rules

- `internal/spec/lint.go`, `lint_ownership.go`, `drift.go`, `parser.go`, `status.go`, `transitions.go`, `dag.go`, `ears.go`, `errors.go`, `sarif.go` — all PRESERVED. This SPEC does NOT extend the lint engine.

### D.6 Unrelated configs (working-tree race-induced `M` entries)

The current `git status --porcelain` shows 4 `M` files plus 6 `??` files unrelated to this SPEC. Per L46 path-specific staging discipline, run-phase staging MUST list only the 2-3 files this SPEC produces:
- `.moai/config/sections/git-convention.yaml` — UNRELATED, PRESERVE
- `.moai/config/sections/language.yaml` — UNRELATED, PRESERVE
- `.moai/config/sections/quality.yaml` — UNRELATED, PRESERVE
- `.moai/harness/usage-log.jsonl` — UNRELATED, PRESERVE
- `.moai/harness/learning-history/` (untracked dir) — UNRELATED, PRESERVE
- `.moai/harness/observations.yaml` (untracked) — UNRELATED, PRESERVE
- `.moai/research/anthropic-best-practices-2026-05-24.md` (untracked) — UNRELATED, PRESERVE
- `.moai/research/v3.0-redesign-2026-05-23.md` (untracked) — UNRELATED, PRESERVE
- `.moai/specs/.moai/` (untracked) — UNRELATED, PRESERVE (likely accidental nested dir from prior session)
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/research.md` (untracked) — UNRELATED, PRESERVE (different SPEC)
- `i18n-validator` (untracked) — UNRELATED, PRESERVE
- `CLAUDE.md`, `CLAUDE.local.md`, `NOTICE.md`, ATR-001 progress.md (race-induced `M`) — UNRELATED, PRESERVE

---

## §E — TRUST 5 Mapping

| TRUST Dim | This SPEC's Approach | Evidence Path |
|-----------|----------------------|---------------|
| **T**ested | M1 D1 NEW test file (`catalog_hash_test.go`) directly enforces hash parity invariant; AC-CHR-004 + AC-CHR-005 + AC-CHR-007 ensure test runs, passes, and is non-destructive. CI auto-runs via existing `go test ./...` workflow per spec.md §E.4. | AC-CHR-003, AC-CHR-004, AC-CHR-005 |
| **R**eadable | Test code follows table-driven convention per `internal/spec/CLAUDE.md` Table-driven test convention. `t.Logf` summary line uses plain English ("verified N catalog entries — 0 drift") per REQ-CHR-005. | REQ-CHR-005, spec.md §F.4 |
| **U**nified | Test reuses generator's `normalizeForHash` semantics exactly (Option α mirror with doc cross-ref, OR Option β extracted shared helper). No divergent hashing logic. | spec.md §D.1 HARD #1, §E.1 mitigation |
| **S**ecured | Test is observation-only — `t.Errorf` on drift, NEVER mutates `catalog.yaml` or source bodies. Per REQ-CHR-008 + AC-CHR-007 git-status guard. Aligns with `internal/spec/CLAUDE.md` "Rules are observation-only" discipline. | REQ-CHR-008, AC-CHR-007, D.1 HARD #2 |
| **T**rackable | This SPEC's run-phase commit ties test failure (when it happens in the future) directly to the SPEC ID via SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001 commit-subject convention. Failure mode is debuggable: `t.Errorf` includes entry name + stored + computed hashes per REQ-CHR-004. | REQ-CHR-004, manager-develop run-phase commit subject |

---

## §F — Run-phase Self-Verification Checklist (manager-develop reference)

Before run-phase commit, manager-develop SHOULD self-verify the following 7 checks (mirror of orchestrator §C batch):

- [ ] Pre-write: SPEC ID `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001` regex decomposition PASS (already done in spec.md §A.1)
- [ ] D1: `catalog_hash_test.go` file authored, passes `go vet`, passes `go build ./...`
- [ ] D2: `internal/template/catalog.yaml` `generated_at:` line updated to current ISO-8601 UTC
- [ ] D3 (if Option β): `internal/template/catalog_hash.go` extracted; `gen-catalog-hashes.go` dry-run output byte-identical pre vs post extraction
- [ ] Full `go test ./...` PASS (existing tests + new TestCatalogHashParity)
- [ ] `git status --porcelain` after test run shows zero file changes (AC-CHR-007)
- [ ] Commit staging includes ONLY this SPEC's 2-3 files (per §D.6 PRESERVE list)

---

## §G — Anti-Patterns (avoid during run-phase)

### G.1 Anti-pattern: regenerating all 38 hashes "to be safe"

Running `go run internal/template/scripts/gen-catalog-hashes.go --all` (without `--dry-run`) during run-phase would rewrite all 38 hash fields. Since current drift is 0, this is a no-op for hash values BUT it would invoke `yaml.Marshal(&cat)` which:
- **Loses YAML comments** (per generator source code comment: "This does NOT preserve comments")
- May reorder YAML keys or change indentation in subtle ways
- Produces a noisy diff that obscures the actual SPEC intent

**Correct approach**: Do NOT invoke `gen-catalog-hashes.go --all` during this SPEC's run-phase. The only catalog.yaml edit is the single-line `generated_at:` update per §B.2.

### G.2 Anti-pattern: extending scope to fix the generator

Once manager-develop is in `gen-catalog-hashes.go`, the temptation is to "just add the auto-`generated_at` update while I'm here". This is scope creep. Per §B.2 SPEC §B.2 of spec.md, generator refactoring is explicitly OUT of scope. Return blocker if the work pulls toward generator changes beyond Option β extraction.

### G.3 Anti-pattern: adding a lint rule

While in `internal/spec/lint.go`-adjacent code, adding a `CatalogHashDriftRule` (so `moai spec lint` flags drift in SPEC-document context) is tempting. Per spec.md §E.6, this is explicitly OUT of scope — the Go test achieves equivalent enforcement. Defer to future `SPEC-V3R6-NAMESPACE-LINT-ENFORCE-001` or similar.

### G.4 Anti-pattern: hard-coding entry count

Hard-coding `if len(entries) != 38` in the test is brittle — the next legitimate catalog entry addition breaks the test for the wrong reason. Either compute `len(entries)` dynamically (preferred) or omit the count assertion entirely (the per-entry hash check is the substantive assertion).

---

## §H — Cross-References

- `spec.md` §A.4 ground-truth measurement table
- `spec.md` §3 7 AC enumeration with verification commands
- `spec.md` §F.1 SSOT cross-references
- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (7-command verification batch obligation)
- `.claude/rules/moai/workflow/verification-batch-pattern.md` § Grouping Heuristic (test + lint + sentinel + smoke pattern)
- `internal/spec/CLAUDE.md` § Catalog hash discipline (the doctrine this SPEC enforces at test-layer)
