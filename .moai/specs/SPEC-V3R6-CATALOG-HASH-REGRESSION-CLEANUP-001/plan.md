---
id: SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001
title: "Catalog Hash Regression Cleanup — Plan"
version: "0.1.1"
status: completed
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
| 0.1.1 | 2026-05-25 | manager-spec | Scope amendment mirror of spec.md v0.1.1 — post-ATR-001 M8 baseline drift discovery (12 ORPHAN + 8 HASH DRIFT at HEAD `f91acb3a3`, see spec.md §A.4-amended). §A M1 D2 deliverable expanded from cosmetic `generated_at` 1-line refresh to full cleanup (3 ordered steps: ORPHAN purge → hash refresh → generated_at). §B.2 catalog.yaml diff plan replaced with 3-step expansion. §A acceptance gating updated 7 AC → 8 AC (AC-CHR-008 added per spec.md). Phase 0.5 plan-auditor iter-1 PASS skip-eligible 0.915 verdict remains valid for expanded scope — amendment stays within Tier S envelope (mechanical YAML edit + same M1 atomic commit pattern + same test design). Predecessor SPEC bodies PRESERVED. |
| 0.1.1 (run-close) | 2026-05-25 | manager-develop | Frontmatter-only update: `status: in-progress → implemented` per Status Transition Ownership Matrix (mirrors spec.md). Body content NOT modified (manager-develop forbidden from §A-H body touch per ownership matrix). Run-phase M1 close: 5 deliverables in single atomic commit (D1 test + D2 catalog full cleanup + D3 progress.md + D4 spec.md frontmatter + D5 plan.md frontmatter). 8/8 AC-CHR PASS. v0.1.1 §A.6 OPTIONAL D3 (`internal/template/normalize_hash.go` extraction) NOT performed — exported helpers already present in pre-existing `internal/template/catalog_hash_norm.go` (per progress.md §B Option α confirmation). |
| 0.1.1 (sync-close) | 2026-05-25 | manager-docs | Frontmatter-only update: `status: implemented → synced` per Status Transition Ownership Matrix. Body content NOT modified. |
| 0.1.1 (mx-close) | 2026-05-25 | orchestrator | Frontmatter-only update: `status: synced → completed` per Status Transition Ownership Matrix (Mx chore orchestrator-direct path). Mx Step C: **EVALUATE-SKIP** (test file fan_in 0 + markdown/YAML 비대상). 4-phase fully closed on main. Mirrors spec.md v0.1.1 (mx-close). |

---

## §A — Milestone Breakdown

### M1: Defensive Verification + Cosmetic Refresh (single 1-pass milestone)

**Scope**: 3 deliverables in one atomic commit (per Tier S 1-pass discipline).

| Deliverable | File | Operation | LOC estimate |
|-------------|------|-----------|--------------|
| D1 | `internal/spec/catalog_hash_test.go` (NEW) OR `internal/template/catalog_hash_test.go` (NEW) | CREATE — Go test that loads catalog.yaml, recomputes 38 entry hashes, asserts parity | ~100-150 LOC (test fixture + walker + assertion + 1 sub-test for missing-templates if Variant A per AC-CHR-006) |
| D2 | `internal/template/catalog.yaml` | MODIFY — **full cleanup (v0.1.1 amendment per spec.md §A.4-amended)**: (a) remove 12 ORPHAN entries (`claude-code-guide`, `manager-{brain,project,quality,strategy}`, `researcher`, 6 `expert-*`) under `catalog.core.agents` + `catalog.optional_packs.{backend,deployment,devops,frontend}.agents` arrays — manual YAML edit; (b) refresh 8 hash drift entries (`moai-foundation-core`, `evaluator-active`, `manager-{develop,docs,git,spec}`, `plan-auditor`, `builder-harness`) via `gen-catalog-hashes.go --all` (post-ORPHAN-purge — `--all` errors on ORPHAN entries' missing source files, hence the strict sequence); (c) update `generated_at:` to current ISO-8601 UTC (verify post-script-run — manual fallback if script does not auto-update per spec.md §A.4 generator "NO" cell) | ~30-40 lines changed (12 ORPHAN entries × 2-3 lines each + 8 hash field updates + generated_at) |
| D3 (optional) | `internal/template/normalize_hash.go` (NEW, if §E.1 Variant A chosen) | CREATE — extract `normalizeForHash` as exported helper; update `gen-catalog-hashes.go` to import it | ~30-50 LOC |

**Decision**: D3 (extraction) is manager-develop's call at run-phase. If extraction is chosen, the generator's existing inline `normalizeForHash` is deleted and replaced with the imported helper — verify no behavior change via the existing dry-run output unchanged.

**Acceptance gating**: M1 completes when all 8 AC-CHR (AC-CHR-001 through AC-CHR-008) PASS independently via their §3 verification commands. (v0.1.1: AC-CHR-008 added — ORPHAN purge post-condition verification.)

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

### B.2 MODIFIED file: `internal/template/catalog.yaml` (v0.1.1 amendment per spec.md §A.4-amended)

**Original v0.1.0 plan**: single-line `generated_at:` field update only (1 line changed).

**Revised v0.1.1 plan**: 3-step ordered cleanup covering the full post-ATR-001 M8 drift (12 ORPHAN + 8 HASH DRIFT). The ordering is enforced because `gen-catalog-hashes.go --all` errors on missing source files (`resolveHashSourcePath` returns `stat %q: %w` at line 123) — ORPHAN entries MUST be removed before script invocation.

**Step 1 — Manual ORPHAN entry removal** (12 entries):

Remove the following entries from `internal/template/catalog.yaml` (located under `catalog.core.agents` and `catalog.optional_packs.{backend,deployment,devops,frontend}.agents` arrays — exact YAML location varies per entry):

- `claude-code-guide` (under `catalog.core.agents`)
- `manager-brain` (under `catalog.core.agents`)
- `manager-project` (under `catalog.core.agents`)
- `manager-quality` (under `catalog.core.agents`)
- `manager-strategy` (under `catalog.core.agents`)
- `researcher` (under `catalog.core.agents` or `catalog.optional_packs.*.agents`)
- `expert-backend` (under `catalog.optional_packs.backend.agents`)
- `expert-frontend` (under `catalog.optional_packs.frontend.agents`)
- `expert-security` (under `catalog.optional_packs.backend.agents` or `catalog.optional_packs.devops.agents`)
- `expert-devops` (under `catalog.optional_packs.devops.agents`)
- `expert-refactoring` (under `catalog.optional_packs.backend.agents` or similar)
- `expert-performance` (under `catalog.optional_packs.backend.agents` or similar)

manager-develop MAY execute Step 1 via either: (α) manual YAML editing with text editor + `git diff` verification, OR (β) a one-off scripted removal (e.g., `yq` or sed pattern). Either approach is acceptable per Tier S envelope. Post-condition: `grep -cE "claude-code-guide|manager-brain|manager-project|manager-quality|manager-strategy|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-refactoring|expert-performance" internal/template/catalog.yaml` returns **0** (down from baseline 24 = 12 entries × 2 lines).

**Step 2 — Hash refresh via generator** (8 entries):

```bash
go run internal/template/scripts/gen-catalog-hashes.go --all
```

Updates the `hash:` field for the 8 retained entries whose source body was modified by ATR-001: `moai-foundation-core`, `evaluator-active`, `manager-develop`, `manager-docs`, `manager-git`, `manager-spec`, `plan-auditor`, `builder-harness`. The script reads each entry's `path:` field, normalizes the source body (CRLF→LF, trailing-whitespace strip, single trailing newline), recomputes sha256, and writes back via `yaml.Marshal(&cat)`.

Post-condition: `go test ./internal/spec/ -run TestCatalogHashParity` reports `0 drift` (or `PASS — verified <N> catalog entries — 0 drift` per REQ-CHR-005).

**Step 3 — `generated_at` refresh verification**:

```bash
grep -E "^generated_at:" internal/template/catalog.yaml
# Expected match: generated_at: "2026-05-25THH:MM:SSZ" (current UTC)
```

The Step 2 script may or may not auto-update `generated_at:` (per spec.md §A.4 original measurement "NO" cell — the generator serializes the field AS-IS via `yaml.Marshal(&cat)`). If grep returns the stale `"2026-05-12T03:00:00Z"`, perform manual one-line update:

```diff
 version: 1.0.0
-generated_at: "2026-05-12T03:00:00Z"
+generated_at: "2026-05-25THH:MM:SSZ"
 catalog:
```

Where `HH:MM:SS` is current UTC time (`date -u +"%Y-%m-%dT%H:%M:%SZ"`).

**Execution order rationale**: Steps 1 → 2 → 3 sequencing is critical. Step 2 (`--all`) on ORPHAN entries errors out and aborts before reaching the retained entries (script does not have `--skip-orphan` flag — see spec.md §B.2 OOS). manager-develop SHOULD verify Step 1 completion via the grep check BEFORE invoking Step 2 to avoid retry cycles.

**Verification batch (post-run, all 3 steps complete)**:
- AC-CHR-001: per-entry drift check (8 entries refresh confirmed)
- AC-CHR-002: `generated_at:` matches `2026-05-2[5-9]T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`
- AC-CHR-004: `TestCatalogHashParity` reports 0 drift (HASH DRIFT + ORPHAN both 0)
- AC-CHR-008: grep against 12 archived name pattern returns 0

**Scope discipline reminder** (per spec.md §B.2 OOS): manager-develop MUST NOT use the in-tree time to refactor `gen-catalog-hashes.go` to auto-update `generated_at:` or add `--skip-orphan` mode. Generator refactoring remains deferred to future `SPEC-V3R6-CATALOG-GENERATOR-MODE-001` per spec.md §F.3. The manual fallback in Step 3 is the v0.1.1 amendment's accepted approach.

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
