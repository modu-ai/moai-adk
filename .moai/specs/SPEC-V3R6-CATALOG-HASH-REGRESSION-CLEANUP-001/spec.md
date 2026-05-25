---
id: SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001
title: "Catalog Hash Regression Cleanup — Drift Verification + Generated-At Refresh + Lint Regression Guard"
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

# SPEC: Catalog Hash Regression Cleanup — Drift Verification + Generated-At Refresh + Lint Regression Guard

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial draft — plan-phase artifacts authored. Pre-flight ground-truth measurement: 0 hash drift across 38 catalog entries, stale `generated_at: "2026-05-12T03:00:00Z"` (13 days), no lint enforcement against future drift. Tier S minimal LEAN form (2 artifacts: spec.md + plan.md, AC inline in §3). |
| 0.1.1 | 2026-05-25 | manager-spec | Scope amendment per user AskUserQuestion Option A (post-ATR-001 M8 baseline drift discovery). §A.4 ground truth updated: pre-flight `0 drift / 38 entries` (against plan-phase commit `128947eb6`) invalidated by intervening ATR-001 M3-M8 cohort commits (`476b04ffb` M3 archive + `fdd4aa37a` M4 hooks + M5/M6 rules/doctrine + `f91acb3a3` M8 template parity). Actual post-M8 state at HEAD `f91acb3a3` is **12 ORPHAN + 8 HASH DRIFT** (20 total drift) — live `TestCatalogHashParity` failure confirms. M8 commit message claimed `catalog.yaml grep for archived names = 0` (false positive — actual `grep` returns 24 = 12 entries × 2 lines). D2 deliverable expanded from cosmetic `generated_at` 1-line refresh to full cleanup: (a) 12 ORPHAN entry removal (`claude-code-guide`, `manager-{brain,project,quality,strategy}`, `researcher`, 6 `expert-*`), (b) 8 hash refresh via `gen-catalog-hashes.go --all` post-ORPHAN-purge, (c) `generated_at` refresh. AC-CHR-004 wording expanded to include ORPHAN as drift type (alongside HASH DRIFT). NEW **AC-CHR-008** added (ORPHAN purge post-condition verification). NEW **REQ-CHR-009** added (Ubiquitous — catalog excludes ORPHAN entries). Total AC count 7 → 8. **L66 NEW pattern (proposed)**: pre-flight ground truth invalidated by intervening cohort commits — defer + re-baseline OR scope-expand. Side-effect benefit: discharges M8 misreported PROCEED-WITH-DEBT. Predecessor SPEC bodies still PRESERVED (HARNESS-NAMESPACE-CLEANUP-001 + LOCAL-NAMESPACE-CONSOLIDATION-001 untouched). |
| 0.1.1 (run-close) | 2026-05-25 | manager-develop | Frontmatter-only update: `status: in-progress → implemented` per Status Transition Ownership Matrix (`in-progress → implemented` owned by manager-develop on M1 commit close per `.claude/rules/moai/development/spec-frontmatter-schema.md`). Body content NOT modified by this update (manager-develop is forbidden from touching §A-H body content per ownership matrix). Run-phase 8/8 AC-CHR PASS per progress.md §E.2. M1 commit SHA: `eabb8db14` (atomic single commit covering 5 deliverables D1-D5). |
| 0.1.1 (sync-close) | 2026-05-25 | manager-docs | Frontmatter-only update: `status: implemented → synced` per Status Transition Ownership Matrix (sync-phase-owned transition). CHANGELOG.md (2 entries: Added `catalog_hash_test.go`, Changed `catalog.yaml`), progress.md §A (Sync phase row appended with sync_commit_sha placeholder per L60 chicken-and-egg pattern). Body content NOT modified by this update. |
| 0.1.1 (mx-close) | 2026-05-25 | orchestrator | Frontmatter-only update: `status: synced → completed` per Status Transition Ownership Matrix (Mx chore commit orchestrator-direct path). Mx Step C verdict: **EVALUATE-SKIP** — `internal/spec/catalog_hash_test.go` test file fan_in 0 + `resolveHashSourcePath` helper fan_in 2 (below @MX:ANCHOR ≥3 threshold), catalog.yaml/markdown 비대상 per mx-tag-protocol.md §a. progress.md §A Mx row pending → complete + §E.5 mx_step_c_decision SKIP-confirmed + mx_phase_close_marker 4-phase-closed. **4-phase fully closed on main** (plan `128947eb6` + amend `61016ad3b` + run `eabb8db14` + sync `5171da19e`+`a045b2506` + mx `014b0344d`+`7c8178972`). |

---

## §A — Context

### A.1 SPEC ID Decomposition (L51 Pre-Write Self-Check)

```
decomposition: SPEC ✓ | V3R6 ✓ | CATALOG ✓ | HASH ✓ | REGRESSION ✓ | CLEANUP ✓ | 001 ✓ → PASS
```

Canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` matches the literal `SPEC` prefix + 6 domain segments (`V3R6`, `CATALOG`, `HASH`, `REGRESSION`, `CLEANUP`, each `[A-Z][A-Z0-9]*` compliant) + `001` digit-only end anchor.

### A.2 Origin and Trigger

**Origin**: Sprint 10 lane B 3rd-SPEC entry. The §24 namespace separation policy (CLAUDE.local.md §24, established 2026-05-23 chore `4f1135684`) introduced `my-harness-*` vs `moai-*` distinction and a new `.claude/agents/local/` namespace (via SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 2026-05-25 4-phase close `c2a41e963`). After the namespace cascade, the user's pre-flight intuition was that `internal/template/catalog.yaml` hash entries had likely drifted because many skill/agent body edits landed since the catalog's `generated_at: "2026-05-12T03:00:00Z"` timestamp.

**Pre-flight ground-truth measurement (orchestrator-verified before plan-phase spawn)**:

```bash
$ go run internal/template/scripts/gen-catalog-hashes.go --all --dry-run
Computing hashes for all 38 entries...
[dry-run] catalog.yaml not modified

# Drift comparison (per-entry old vs new sha256):
total: 38 | drift: 0 | unchanged: 38

# Orphan / missing-source scan:
(no MISSING_FROM_CATALOG entries; no ORPHAN_ENTRY entries)
```

**Reality**: 0 actual hash drift. All 38 catalog entries' sha256 values match current normalized source bodies (post-CRLF, post-trim, single-trailing-newline). The user's intuition was a false alarm.

**True residual gap (3 cosmetic + structural items)**:

1. **Cosmetic stale `generated_at`**: top-of-file `generated_at: "2026-05-12T03:00:00Z"` is 13 days stale. Misleading anyone who reads it as "this catalog is current as of 2026-05-12 and may have drifted since". The generator `gen-catalog-hashes.go` writes back via `yaml.Marshal(&cat)` which serializes `GeneratedAt` AS-IS — so even invoking `--all` does NOT refresh the timestamp. This is a latent generator gap, not a content gap.

2. **No future-drift regression guard**: `internal/spec/lint.go` has 8+ rules (`EARSModalityRule`, `FrontmatterSchemaRule`, `OwnershipTransitionRule`, etc.) but NO rule that recomputes catalog entry hashes and compares against the stored values. Future SPECs that edit a SKILL.md body without invoking `gen-catalog-hashes.go --all` (per `internal/spec/CLAUDE.md` Catalog hash discipline) will produce real drift undetected until cosmetic discovery weeks later. The §24 namespace policy itself is doc-only — `grep HarnessNamespace internal/spec/lint.go` returns 0 matches.

3. **Defensive lock-in test absent**: `internal/spec/` test suite has `drift_chore_skip_test.go`, `drift_specid_grep_test.go`, but no `catalog_hash_test.go` that loads `catalog.yaml`, recomputes all 38 entries' hashes via `gen-catalog-hashes.go` parity logic, and asserts 0 drift. Without this test, the "0 drift today" finding could regress silently the next time a contributor edits a skill body without regen.

### A.3 Predecessor Context

This SPEC closes a residual gap left by:

- **SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001** (status: implemented, 2026-05-25 close `0b99c4943`) — §24 SSOT self-consistency cleanup. Did NOT include catalog.yaml hash regeneration in M1-M3 (per manager-spec scope). PRESERVED — body untouched by this SPEC.
- **SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001** (status: implemented, 2026-05-25 4-phase close `c2a41e963`) — `.claude/agents/local/` namespace introduction + 2 agent migration + 1184 LOC consolidation. Did NOT include catalog.yaml regeneration (out of scope per spec.md §B). PRESERVED — body untouched by this SPEC.

### A.4 Verified Ground Truth (Plan-Phase Pre-Flight)

| Fact | Value | Verification Command |
|------|-------|----------------------|
| catalog.yaml total entries | 38 | `grep -cE "^[[:space:]]+- name:" internal/template/catalog.yaml` |
| Actual hash drift | 0 | `gen-catalog-hashes.go --all --dry-run` + per-entry diff |
| Orphan catalog entries (no source file) | 0 | `awk` over catalog `path:` fields + `[ -e ]` check |
| Skills missing from catalog | 0 | `find templates/.claude/skills -name SKILL.md` + grep |
| Agents missing from catalog | 0 | `find templates/.claude/agents -name *.md` + grep |
| Stale `generated_at` age | 13 days | `2026-05-25` minus `2026-05-12` |
| `gen-catalog-hashes.go` updates `generated_at` on `--all` | NO | Source inspection: `yaml.Marshal(&cat)` serializes AS-IS |
| Lint rules referencing catalog hash | 0 | `grep -E "CatalogHash\|catalog\.hash" internal/spec/lint.go` |
| §24 namespace lint rules | 0 | `grep -E "HarnessNamespace\|harness\.namespace" internal/spec/lint.go` |

### A.4-amended (v0.1.1): Post-ATR-001 M8 actual ground truth

**Discovery context** (orchestrator pre-flight verification, 2026-05-25, run-phase resumption attempt):

The original §A.4 measurement above was performed against plan-phase commit `128947eb6` (catalog state at that point: 0 drift / 38 entries valid). Between commit `128947eb6` and current HEAD `f91acb3a3`, the SPEC-V3R6-AGENT-TEAM-REBUILD-001 (ATR-001) cohort committed M3 through M8 (5 ATR-001 commits + 1 LNCO reconcile interleaved):

- `476b04ffb` ATR-001 M3 — 11 agent archive (deleted 11 source files under `internal/template/templates/.claude/agents/{core,meta,expert}/`)
- `fdd4aa37a` ATR-001 M4 — 3 NEW hook scripts (orchestrator-direct)
- M5 `498ea18a2` — rule files (10 rules)
- M6 `3136733f6` — supersedence verify + AC-ATR-012 reinforcement
- M7 `4f37a7643` — CLAUDE.md catalog + CLAUDE.local.md doctrine + NOTICE.md attribution
- `f91acb3a3` ATR-001 M8 — template parity + catalog.yaml 12-archived purge + 7-item verification batch (PROCEED-WITH-DEBT)

ATR-001 M3 archive made the 12 catalog entries point to non-existent source files (ORPHAN). ATR-001 M1/M5/M6/M7 modified 6+ retained agent bodies (manager-develop, manager-docs, manager-git, manager-spec, plan-auditor, builder-harness, evaluator-active) + 1 SKILL.md (moai-foundation-core) without invoking `gen-catalog-hashes.go --all`, producing 8 HASH DRIFT.

**Actual post-M8 ground truth (verified via `go test ./internal/spec/ -run TestCatalogHashParity 2>&1`, 2026-05-25, HEAD `f91acb3a3`)**:

| Drift Class | Count | Entries |
|-------------|-------|---------|
| ORPHAN (source file deleted) | 12 | `claude-code-guide`, `manager-brain`, `manager-project`, `manager-quality`, `manager-strategy`, `researcher`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-refactoring`, `expert-performance` |
| HASH DRIFT (source body modified, hash stale) | 8 | `moai-foundation-core`, `evaluator-active`, `manager-develop`, `manager-docs`, `manager-git`, `manager-spec`, `plan-auditor`, `builder-harness` |
| **Total drift** | **20** | (12 ORPHAN + 8 HASH DRIFT) |

**M8 commit message false positive** (`f91acb3a3`): M8 PROCEED-WITH-DEBT verification batch claimed `catalog.yaml grep for archived names = 0 ✓`. This was a false positive — live `grep -cE "claude-code-guide|manager-brain|manager-project|manager-quality|manager-strategy|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-refactoring|expert-performance" internal/template/catalog.yaml` returns **24** (12 entries × 2 lines each: `- name:` + `path:`). The catalog purge claimed in M8 commit body did not occur. Discharging this M8 debt is now a side-effect of this CATALOG SPEC's expanded scope per v0.1.1 amendment.

**`gen-catalog-hashes.go --all` behavior on ORPHAN entries**: The generator only updates `hash:` for existing entries; it does NOT remove entries whose source file is missing. On ORPHAN entries, `resolveHashSourcePath` returns an error (`stat %q: %w` at line 123) and `updateEntryHash` propagates. Therefore the v0.1.1 D2 deliverable cleanup procedure MUST sequence: **(1) manually remove 12 ORPHAN entries from catalog.yaml** first → **(2) then run `gen-catalog-hashes.go --all`** to refresh remaining 8 hash drift entries → **(3) verify `generated_at` is refreshed** (script may or may not auto-update; manual verification required per spec.md §A.4 original "NO" cell).

**Original §A.4 historical value**: The original §A.4 measurement table above remains the historical record of catalog state at plan-phase commit `128947eb6` and is NOT deleted. This v0.1.1 amendment is additive — both ground truths coexist, anchored to their respective commit baselines.

**L66 NEW pattern (proposed)**: pre-flight ground truth invalidated by intervening cohort commits. When a SPEC's pre-flight measurement establishes a baseline (e.g., "0 drift") AT a specific commit, and cohort sibling SPECs (here ATR-001 M3-M8) modify shared state between the SPEC's plan-phase commit and run-phase start, the original ground truth no longer reflects HEAD state. Mitigation options at run-phase: **defer** (Option 1 — wait for cohort to stabilize, then re-baseline), **PASS-WITH-DEBT** (Option 2 — proceed against new state, accept residual drift), **scope-expand** (Option 3 — this SPEC's chosen path, amend SPEC body to cover the expanded drift), **pause** (Option 4 — halt and request user direction). Distinct from L52 committed-commit race (which is about fast-forward attribution); L66 is about ground-truth invalidation between plan and run.

---

## §B — Scope

### B.1 In Scope

1. **Defensive hash verification**: Author a Go test (`internal/spec/catalog_hash_test.go`) that loads `catalog.yaml`, walks every entry, recomputes the sha256 via the same normalization rules as `gen-catalog-hashes.go` (CRLF→LF, trailing-whitespace strip, single trailing newline), and asserts the stored hash matches. This locks in the "0 drift" baseline and turns future drift into a CI-blocking test failure.

2. **`generated_at` refresh**: Update `catalog.yaml` top-level `generated_at` to the current ISO-8601 UTC timestamp (one-time manual refresh during run-phase). Out-of-scope: refactoring `gen-catalog-hashes.go` to auto-update the timestamp on `--all` (deferred to a follow-up SPEC if scope warrants).

3. **Documentation reinforcement**: Cross-reference `internal/spec/CLAUDE.md` § Catalog hash discipline to the new test file. Reinforces the doctrine: "Never edit `catalog.yaml` hash fields by hand; regenerate via `gen-catalog-hashes.go --all`; the new `catalog_hash_test.go` enforces this discipline at CI."

### B.2 Out of Scope — Generator Refactor

Refactoring `gen-catalog-hashes.go` to (a) auto-update `generated_at` on `--all`, (b) preserve YAML comments via node-based round-trip, or (c) add a `--check` mode that exits non-zero on drift is OUT of scope. Reason: scope creep; this SPEC is Tier S minimal defensive cleanup, not a generator overhaul. If a future contributor demands these features, propose a follow-up `SPEC-V3R6-CATALOG-GENERATOR-MODE-001`.

### B.3 Out of Scope — Catalog Schema Changes

Adding new catalog fields (e.g., `last_verified_at` per-entry, `verification_signature`), changing the YAML structure (e.g., flattening `catalog.core.skills` → `catalog.skills` with a `tier:` field), or migrating to a different serialization format (TOML, JSON) is OUT of scope. Reason: schema changes cascade into `gen-catalog-hashes.go`, `internal/spec/dag.go`, downstream consumers — all of which were stable since 2026-05-11.

### B.4 Out of Scope — New Catalog Entries

Adding new skill or agent entries to the catalog (e.g., the `.claude/agents/local/` agents from LNCO-001 if they were intended to be catalog-tracked) is OUT of scope. Per LNCO-001 spec.md §B, the `.claude/agents/local/` namespace is user-owned and explicitly excluded from `moai update` sync, which means it MUST NOT appear in `catalog.yaml` (catalog tracks template-distributed assets only). If a future audit reveals genuinely missing catalog entries, propose a follow-up `SPEC-V3R6-CATALOG-COVERAGE-EXTEND-001`.

### B.5 Out of Scope — §24 Namespace Lint Enforcement

The §24 separation policy (`my-harness-*` vs `moai-*` skill prefix, `.claude/agents/local/` user-owned directory) is currently doc-only. Adding a lint rule that scans `.claude/skills/` and `.claude/agents/` for namespace-prefix violations is OUT of scope. Reason: catalog hash discipline (this SPEC's focus) and namespace prefix enforcement are orthogonal concerns. Propose a follow-up `SPEC-V3R6-NAMESPACE-LINT-ENFORCE-001` if regression evidence emerges.

---

## §C — Requirements (GEARS Notation, 100% Self-Dogfood)

> All REQ-CHR clauses use GEARS notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Five Patterns. Zero IF/THEN constructs. Generalized `<subject>` (function/file/test, not only "the system").

### C.1 Ubiquitous

- **REQ-CHR-001** [Ubiquitous]: The `internal/template/catalog.yaml` file shall maintain sha256 hash field parity with the normalized source body referenced by each entry's `path:` field. Normalization rules match `gen-catalog-hashes.go` `normalizeForHash` exactly: CRLF→LF conversion, trailing-whitespace strip per line, exactly one trailing newline. (Binds: AC-CHR-001)

- **REQ-CHR-002** [Ubiquitous]: The `internal/template/catalog.yaml` `generated_at:` top-level field shall hold an ISO-8601 UTC timestamp (`YYYY-MM-DDTHH:MM:SSZ` format) reflecting the most recent manual catalog refresh date. (Binds: AC-CHR-002)

- **REQ-CHR-003** [Ubiquitous]: The `internal/spec/catalog_hash_test.go` test file shall enumerate every catalog entry, recompute the sha256 hash via the same normalization rules as `gen-catalog-hashes.go`, and assert exact match against the stored value, failing the test on any mismatch. (Binds: AC-CHR-003)

### C.2 Event-Driven (When ...)

- **REQ-CHR-004** [Event-driven]: When a contributor runs `go test ./internal/spec/... -run TestCatalogHashParity`, the test harness shall load `internal/template/catalog.yaml`, iterate all entries in `catalog.core.{skills,agents,rules}` + `catalog.optional.*` sections, and report drift as t.Errorf with `entry name | stored hash | computed hash` triplet for each mismatch. **Drift is defined to include two distinct classes**: (a) **HASH DRIFT** — catalog entry whose stored sha256 mismatches the recomputed normalized source hash (existing source file present); and (b) **ORPHAN** — catalog entry whose `path:` field points to a source file that no longer exists on disk (detected via `resolveHashSourcePath` returning a non-nil error). Both classes constitute drift; either failure mode produces a t.Errorf. (Binds: AC-CHR-004, AC-CHR-008)

- **REQ-CHR-005** [Event-driven]: When the `catalog_hash_test.go` test detects zero drift, the test shall complete with PASS status and emit a single informational t.Logf summary `verified <N> catalog entries against normalized source bodies — 0 drift` where N matches the catalog's entry count. (Binds: AC-CHR-004)

### C.3 State-Driven (While ...)

- **REQ-CHR-006** [State-driven]: While `internal/template/catalog.yaml` exists and is parseable by `gopkg.in/yaml.v3`, the `catalog_hash_test.go` test shall run without skip, regardless of working-tree state (clean, dirty, sandbox-isolated `t.TempDir()` etc.). (Binds: AC-CHR-005)

### C.4 Where (Capability Gate)

- **REQ-CHR-007** [Where capability gate]: Where the test environment exposes `os.ReadFile` access to `internal/template/templates/**/*` source bodies (i.e., the test runs from the repo root or a checkout containing the templates tree), the `catalog_hash_test.go` test shall compute hashes against the actual source files; otherwise it shall fail with a clear "missing templates directory" error rather than silently skipping. (Binds: AC-CHR-006)

### C.5 Unwanted (shall not ...)

- **REQ-CHR-008** [Unwanted]: The `catalog_hash_test.go` test shall not mutate `internal/template/catalog.yaml`, source body files, or any file under the repository working tree, irrespective of detected drift. Hash regeneration is reserved for `gen-catalog-hashes.go --all` (manual, explicit). (Binds: AC-CHR-007)

- **REQ-CHR-009** [Ubiquitous]: The `internal/template/catalog.yaml` file shall not contain catalog entries whose `path:` field points to a source file that does not exist on disk. ORPHAN entries (catalog entry whose source file has been deleted, archived, or moved without a corresponding catalog update) constitute a doctrinal violation of the "catalog tracks template-distributed assets only" rule from spec.md §B.4. The 12 known ORPHAN entries at run-phase baseline HEAD `f91acb3a3` (per §A.4-amended: `claude-code-guide`, `manager-brain`, `manager-project`, `manager-quality`, `manager-strategy`, `researcher`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-refactoring`, `expert-performance`) shall be removed from `catalog.yaml` during the v0.1.1 D2 cleanup. (Binds: AC-CHR-008)

---

## §D — HARD / SHOULD Constraints

### D.1 HARD Constraints

1. **[HARD]** `catalog_hash_test.go` MUST use the exact same normalization rules as `gen-catalog-hashes.go` `normalizeForHash` function. Any divergence produces false positives or false negatives — the test must be a faithful mirror of the generator's hashing logic. Implementation strategy: extract `normalizeForHash` into an exported helper within `internal/template/` package OR duplicate the byte-for-byte normalization logic in the test, with a comment cross-referencing the generator source and a TODO to consolidate in a future SPEC.

2. **[HARD]** `catalog_hash_test.go` MUST NOT regenerate or write to `catalog.yaml` under any circumstance. The test is observation-only (parallel to lint rules per `internal/spec/CLAUDE.md`: "Rules are observation-only — NEVER mutate the spec or call `os.Exec`, `ioutil.WriteFile`, or any file-modify primitive.").

3. **[HARD]** `generated_at:` refresh MUST be a manual single-line YAML edit performed in the run-phase commit; this SPEC's run-phase MUST NOT introduce auto-update logic to `gen-catalog-hashes.go` (deferred per §B.2).

4. **[HARD]** Predecessor SPEC bodies (HARNESS-NAMESPACE-CLEANUP-001 spec.md/plan.md/acceptance.md/progress.md + LOCAL-NAMESPACE-CONSOLIDATION-001 spec.md/plan.md/acceptance.md/progress.md) MUST NOT be modified by this SPEC. Cite via cross-reference only.

5. **[HARD]** SPEC ID `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001` regex pre-write self-check executed and PASSED (see §A.1).

### D.2 SHOULD Constraints

1. **[SHOULD]** The test file `catalog_hash_test.go` SHOULD live in the `internal/spec/` package (alongside existing `drift_*.go` test files) to inherit the package's existing test infrastructure (`testdata/` fixtures, helper functions). Alternative: `internal/template/` package if the package boundary makes more sense — manager-develop's judgment call at run-phase.

2. **[SHOULD]** The test SHOULD report PASS in < 500 ms wall time on a typical developer laptop (38 entries × small SKILL.md files, ~5-20 KB each = ~500 KB total I/O). If wall time exceeds 1 second, defer optimization to a follow-up SPEC.

3. **[SHOULD]** The `generated_at:` refresh commit message SHOULD include a brief comment in the commit body noting that this is a manual cosmetic refresh (since the generator does not auto-update) and that a follow-up SPEC may automate this.

---

## §3 — Acceptance Criteria (Inline per Tier S LEAN)

> Per Tier S LEAN convention, AC enumeration lives inline in spec.md §3 (no separate `acceptance.md` file). Each AC binds to one or more REQ-CHR clauses and includes an independently verifiable command.

### AC-CHR-001 [MUST-PASS] — Hash parity verified at run-phase

**Binds**: REQ-CHR-001

**Given**: Run-phase work is in progress on `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001` at the moai-adk-go repo root.

**When**: The contributor runs the dry-run drift comparison.

```bash
go run internal/template/scripts/gen-catalog-hashes.go --all --dry-run > /tmp/dryrun.txt 2>&1
# Then per-entry compare stored vs computed (orchestrator pre-flight method shown in §A.4)
```

**Then**: Exactly 38 entries are reported as hash-parity-verified (0 drift). The `[dry-run] catalog.yaml not modified` footer is present. Stdout last line matches `total: 38 | drift: 0 | unchanged: 38` from the comparison script.

**Verification command** (orchestrator post-run):
```bash
go run internal/template/scripts/gen-catalog-hashes.go --all --dry-run 2>&1 | grep -c "^  \[dry-run\]"
# Expected output: 38
```

### AC-CHR-002 [MUST-PASS] — generated_at refreshed to current UTC date

**Binds**: REQ-CHR-002

**Given**: `internal/template/catalog.yaml` is on the run-phase commit baseline.

**When**: The contributor reads the top of catalog.yaml.

**Then**: The `generated_at:` field holds an ISO-8601 UTC timestamp dated 2026-05-25 (or later if run-phase rolls over a day boundary). The timestamp format matches `YYYY-MM-DDTHH:MM:SSZ` exactly (no fractional seconds, no offset).

**Verification command**:
```bash
grep -E "^generated_at:" internal/template/catalog.yaml | grep -E "\"2026-05-2[5-9]T[0-9]{2}:[0-9]{2}:[0-9]{2}Z\""
# Expected: 1 match line
```

### AC-CHR-003 [MUST-PASS] — catalog_hash_test.go file present and compiles

**Binds**: REQ-CHR-003

**Given**: Run-phase work has completed file authoring.

**When**: The contributor runs `go vet ./internal/spec/...` (or `./internal/template/...` if manager-develop chose that package per §D.2 SHOULD #1).

**Then**: The file exists at the chosen location, passes `go vet`, and `go build ./...` succeeds with zero errors.

**Verification commands**:
```bash
ls internal/spec/catalog_hash_test.go || ls internal/template/catalog_hash_test.go
go vet ./internal/spec/... ./internal/template/...
go build ./...
# All three commands: exit code 0
```

### AC-CHR-004 [MUST-PASS] — TestCatalogHashParity runs and reports 0 drift

**Binds**: REQ-CHR-004, REQ-CHR-005

**Given**: `catalog_hash_test.go` is authored and compiles per AC-CHR-003.

**When**: The contributor runs the test.

```bash
go test ./internal/spec/... -run TestCatalogHashParity -v
# Or whichever package + test name manager-develop chose
```

**Then**: The test reports PASS with a single `t.Logf` line matching the pattern `verified 38 catalog entries against normalized source bodies — 0 drift` (entry count 38 today; the test should hard-code or auto-compute the count from the loaded catalog).

**Verification command**:
```bash
go test ./internal/spec/... -run TestCatalogHashParity -v 2>&1 | grep -E "verified [0-9]+ catalog entries.*0 drift"
# Expected: 1 match line
go test ./internal/spec/... -run TestCatalogHashParity 2>&1 | tail -1 | grep -E "^ok|^PASS"
# Expected: 1 match line (PASS marker)
```

### AC-CHR-005 [MUST-PASS] — Test runs without skip on clean working tree

**Binds**: REQ-CHR-006

**Given**: Repo working tree is clean (no `M` or `??` git status entries that affect catalog.yaml or templates/).

**When**: The contributor runs the test from the repo root.

**Then**: The test executes (does not emit `t.Skip()`), completes within 1 second wall time, and reports PASS.

**Verification command**:
```bash
go test ./internal/spec/... -run TestCatalogHashParity -v 2>&1 | grep -E "SKIP|t\.Skip"
# Expected: 0 matches (no skip detected)
go test ./internal/spec/... -run TestCatalogHashParity -count=1 2>&1 | grep -E "^ok.*[0-9]+\.[0-9]+s"
# Expected: 1 match line; the timing value should be < 1.000s
```

### AC-CHR-006 [MUST-PASS] — Missing-templates failure mode produces clear error

**Binds**: REQ-CHR-007

**Given**: A sandbox-isolated test scenario where `internal/template/templates/` is absent (e.g., a test fixture intentionally renames the templates dir for one subtest, OR the existing test runs from a directory that lacks templates/).

**When**: The contributor runs the test against the broken environment.

**Then**: The test fails with `t.Fatal` (NOT `t.Skip`), and the failure message contains the substring `templates directory` or equivalent diagnostic naming the missing path.

**Verification approach** (acceptable variants — manager-develop's judgment):
- Variant A: Add a sub-test that uses `t.TempDir()` to construct a deliberately broken environment and asserts the test fails with the expected message.
- Variant B: Document that the production test will encounter this case in CI and rely on the existing `internal/template/templates/` presence as the implicit "happy path" — fail-loud is the inherent behavior because `os.ReadFile` on a missing path returns a non-nil error which the test propagates via `t.Fatalf`.

Either variant satisfies this AC; manager-develop chooses based on test-fixture complexity vs maintenance cost.

**Verification command** (Variant A example):
```bash
go test ./internal/spec/... -run TestCatalogHashParity/missing_templates -v 2>&1 | grep -E "templates directory"
# Expected: 1 match line (the sub-test's failure message surfaces in test output)
```

### AC-CHR-007 [MUST-PASS] — Test does not mutate any file

**Binds**: REQ-CHR-008

**Given**: Repo working tree is clean before test execution.

**When**: The contributor runs the test followed by `git status`.

**Then**: `git status --porcelain` shows zero changes attributable to the test run. The test is provably observation-only.

**Verification command**:
```bash
git status --porcelain > /tmp/pre.txt
go test ./internal/spec/... -run TestCatalogHashParity -count=3
git status --porcelain > /tmp/post.txt
diff /tmp/pre.txt /tmp/post.txt
# Expected: empty diff (zero file modifications)
```

### AC-CHR-008 [MUST-PASS] — ORPHAN purge post-condition verification

**Binds**: REQ-CHR-009, REQ-CHR-004 (ORPHAN drift class)

**Given**: Run-phase D2 cleanup has completed — 12 archived agent entries removed from `internal/template/catalog.yaml`.

**When**: The contributor runs the ORPHAN grep check against catalog.yaml.

**Then**: The `grep -cE` against the 12 archived agent name pattern returns exactly **0** matches (down from the pre-v0.1.1 baseline of 24 — see §A.4-amended). Zero residual references to any of the 12 archived agent names in catalog.yaml.

**Verification command**:
```bash
grep -cE "claude-code-guide|manager-brain|manager-project|manager-quality|manager-strategy|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-refactoring|expert-performance" internal/template/catalog.yaml
# Expected: 0
```

**Secondary verification** (TestCatalogHashParity ORPHAN class signal): after D2 cleanup, the test's per-entry walker MUST NOT encounter any `CATALOG_ENTRY_ORPHAN` error class for the 12 listed agent names. Combined with the hash-refresh of the 8 remaining HASH DRIFT entries, the full test run reports `0 drift` per AC-CHR-004.

```bash
go test ./internal/spec/... -run TestCatalogHashParity -v 2>&1 | grep -cE "claude-code-guide|manager-brain|manager-project|manager-quality|manager-strategy|researcher|expert-backend|expert-frontend|expert-security|expert-devops|expert-refactoring|expert-performance"
# Expected: 0 (no test-error references to the 12 archived names)
```

### REQ↔AC Bidirectional Traceability Matrix

| REQ-CHR | Pattern | Binds to AC | Notes |
|---------|---------|-------------|-------|
| REQ-CHR-001 | Ubiquitous | AC-CHR-001 | Hash parity invariant |
| REQ-CHR-002 | Ubiquitous | AC-CHR-002 | `generated_at` ISO-8601 format invariant |
| REQ-CHR-003 | Ubiquitous | AC-CHR-003 | Test file exists + compiles |
| REQ-CHR-004 | Event-driven (When ... runs go test) | AC-CHR-004, AC-CHR-008 | Per-entry drift report format (HASH DRIFT + ORPHAN classes) |
| REQ-CHR-005 | Event-driven (When ... 0 drift) | AC-CHR-004 | PASS-status summary log |
| REQ-CHR-006 | State-driven (While ... exists) | AC-CHR-005 | No skip behavior |
| REQ-CHR-007 | Where (capability gate ... templates available) | AC-CHR-006 | Fail-loud on missing templates |
| REQ-CHR-008 | Unwanted (shall not mutate) | AC-CHR-007 | Observation-only test discipline |
| REQ-CHR-009 | Ubiquitous | AC-CHR-008 | Catalog excludes ORPHAN entries (12 archived agents removed) |

| AC-CHR | Cites REQ-CHR | MUST-PASS? |
|--------|---------------|------------|
| AC-CHR-001 | REQ-CHR-001 | YES — drift verified |
| AC-CHR-002 | REQ-CHR-002 | YES — timestamp refreshed |
| AC-CHR-003 | REQ-CHR-003 | YES — test compiles |
| AC-CHR-004 | REQ-CHR-004, REQ-CHR-005 | YES — test PASS with 0-drift log |
| AC-CHR-005 | REQ-CHR-006 | YES — no skip |
| AC-CHR-006 | REQ-CHR-007 | YES — fail-loud on missing templates |
| AC-CHR-007 | REQ-CHR-008 | YES — no mutation |
| AC-CHR-008 | REQ-CHR-009, REQ-CHR-004 | YES — 12 ORPHAN entries purged from catalog (grep returns 0) |

8 AC total, all MUST-PASS, 100% bidirectional traceability.

---

## §E — Edge Cases / Out of Scope Clarifications

### E.1 Edge Case — Generator and Test Normalization Drift

**Risk**: A future contributor edits `gen-catalog-hashes.go normalizeForHash` (e.g., adds Unicode BOM stripping) without updating `catalog_hash_test.go` mirror logic. Test then false-positives drift across all 38 entries because the new generator-side normalization produces different bytes than the test-side mirror.

**Mitigation** (manager-develop's run-phase decision): Either (a) extract `normalizeForHash` to an exported function in `internal/template/` package and have both the generator and the test import it (preferred), OR (b) leave duplicated logic with a doc comment in each side cross-referencing the other and a TODO. Variant (a) eliminates the drift risk at structural level; variant (b) accepts the risk with documentation safeguards.

### E.2 Edge Case — Future Catalog Schema Extension

**Risk**: A follow-up SPEC adds a new top-level catalog section (e.g., `catalog.experimental.skills`) and the existing `allEntries(&cat)` walker in `gen-catalog-hashes.go` is updated, but `catalog_hash_test.go` walker is not — test silently skips new entries.

**Mitigation**: The test's entry walker MUST be implemented in terms of the same `allEntries` helper used by `gen-catalog-hashes.go` (exported from `internal/template/` package if extraction per §E.1 is performed). If schema extension happens via a future SPEC, the walker is updated in one place.

### E.3 Edge Case — Symlinks or non-regular files in templates/

**Risk**: A skill or agent path resolves to a symlink or non-regular file (e.g., `.DS_Store`, fifo). `os.ReadFile` may return unexpected bytes or fail.

**Mitigation**: The existing `gen-catalog-hashes.go resolveHashSourcePath` already handles path resolution; if it returns an error, `updateEntryHash` propagates the error. The test should mirror this behavior — propagate as `t.Errorf` rather than skipping, ensuring symlink shenanigans surface as visible failures.

### E.4 Out of Scope Clarification — CI Integration

This SPEC adds a Go test that runs as part of the standard `go test ./...` suite. CI workflows already invoke `go test ./...` (per `.github/workflows/test.yml`), so the new test will run automatically in CI without additional config. **This is intentional** — no separate CI job, no new GitHub Action.

### E.5 Out of Scope Clarification — Performance Benchmarking

A `BenchmarkCatalogHashParity` benchmark to track hash-computation throughput is OUT of scope. Reason: the test runs in < 1 second on 38 entries; benchmarking adds maintenance overhead without clear value. If the catalog grows to 100+ entries and test wall-time exceeds the §D.2 SHOULD #2 threshold, propose a follow-up benchmarking SPEC.

### E.6 Out of Scope Clarification — Lint Rule for Catalog Hash

Adding a `CatalogHashDriftRule` to `internal/spec/lint.go` (so that `moai spec lint` surfaces drift) is OUT of scope. Reason: the Go test (`catalog_hash_test.go`) achieves equivalent enforcement via the CI test suite. Adding a lint rule would be redundant with the test, and lint rules operate on individual SPEC documents whereas catalog drift is a project-wide artifact concern — a different abstraction boundary.

---

## §F — References

### F.1 SSOT Cross-References

- `internal/spec/CLAUDE.md` § Catalog hash discipline — authoritative doctrine: "When a SPEC body's §A.3 evidence section changes, the SHA256 hash in `catalog.yaml` is invalidated. Regenerate via `gen-catalog-hashes.go --all` as a same-SPEC cascade. Never edit `catalog.yaml` hash fields by hand."
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field frontmatter schema (this SPEC complies).
- `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Five Patterns — REQ-CHR notation guide.
- `CLAUDE.local.md` §24 Harness Namespace 분리 정책 — provides cascading drift context that motivated this SPEC.

### F.2 Predecessor SPECs (PRESERVE — body untouched)

- `SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001` (implemented, `0b99c4943`) — §24 SSOT cleanup, did not include catalog regen.
- `SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001` (implemented, `c2a41e963`) — `.claude/agents/local/` namespace introduction, did not include catalog regen.

### F.3 Future Follow-Up SPEC Anchors

- `SPEC-V3R6-CATALOG-GENERATOR-MODE-001` (anchor, not yet created) — generator refactor for auto-`generated_at` update + `--check` mode + YAML comment preservation. Anchored from §B.2.
- `SPEC-V3R6-CATALOG-COVERAGE-EXTEND-001` (anchor, not yet created) — catalog schema extension for missing entries. Anchored from §B.4.
- `SPEC-V3R6-NAMESPACE-LINT-ENFORCE-001` (anchor, not yet created) — §24 namespace prefix lint rule. Anchored from §B.5.

### F.4 Implementation Files Reference (Read-only at plan-phase)

- `internal/template/catalog.yaml` (293 lines, 11204 bytes) — target for `generated_at` refresh
- `internal/template/scripts/gen-catalog-hashes.go` (267 lines, 8522 bytes) — normalization SSOT
- `internal/spec/lint.go` (29214 bytes) — sibling lint rules for pattern reference
- `internal/spec/drift.go`, `drift_chore_skip_test.go`, `drift_specid_grep_test.go` — sibling test patterns
