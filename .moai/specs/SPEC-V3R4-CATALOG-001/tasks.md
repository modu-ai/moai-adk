# Task Decomposition — SPEC-V3R4-CATALOG-001

> **SPEC**: SPEC-V3R4-CATALOG-001 — 3-Tier Catalog Manifest
> **Source**: plan.md M1-M5 (~28 sub-tasks normalized to T-001..T-026 for tasks.md schema consistency).
> **Phase**: This file is Phase 1.5 output (decomposition only — no code).
> **Consumed by**: Phase 2B Drift Guard reads `planned_files` column to detect scope creep.
> **TDD style**: M3 (T-009..T-018) is RED-driven (audit tests written first, will fail against empty manifest). M2 (T-002..T-008) is procedural data lock-in. M4 (T-019..T-023) is GREEN consolidation. M5 (T-024..T-026) is documentation.
>
> **Catalog Entry**: 37 skills + 28 agents = 65 entries (verified via `find ... | wc -l`).
>
> **Sync status (2026-05-12)**: All 26 tasks complete via PR #862 (M1-M5, main `ec80c8845`) + PR #863 (eval-1 fixes, main `0d4bf14ef`). Status `pending` markers in the table below are historical (Phase 1.5 snapshot); final completion is recorded in `progress.md` § "Run Final State".

## Task Table

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | M1.1 — Draft `catalog.yaml` schema skeleton (top-level: `version`, `generated_at`, `catalog`; sub-sections: `core`, `optional_packs`, `harness_generated`; per-entry fields: `name`, `path`, `tier`, `hash`, `version`, optional `depends_on`; per-pack fields: `description`, `depends_on`, `skills`, `agents`, optional reserved `marketplace_id`/`marketplace_url`/`publisher`). Write empty placeholder structure; populate in T-002..T-008. | REQ-001, REQ-002, REQ-003, REQ-004, REQ-011, REQ-012, REQ-024 | - | internal/template/catalog.yaml | pending |
| T-002 | M2.1 — Populate `catalog.core.skills` with 18 core skill entries from research.md §Tier Classification Map (moai, moai-foundation-{cc,core,quality,thinking}, moai-workflow-{spec,project,testing,ddd,tdd}, moai-meta-harness, moai-ref-testing-pyramid). Each entry: `name`, `path` (relative to `templates/`), `tier: core`, `hash: TODO` (filled in T-007), `version: "1.0.0"`. | REQ-005, REQ-009 | T-001 | internal/template/catalog.yaml | pending |
| T-003 | M2.1b — Populate `catalog.core.agents` with 15 core agent entries (manager-{spec,develop,strategy,quality,docs,project,git,cycle,ddd,tdd,brain}, evaluator-active, plan-auditor, builder-skill, builder-agent — final list locked in T-002 review). Each entry includes `name`, `path`, `tier: core`, `hash: TODO`, `version: "1.0.0"`. | REQ-006, REQ-009 | T-001 | internal/template/catalog.yaml | pending |
| T-004 | M2.2 — Populate `catalog.optional_packs.<pack>` for 9 packs (backend, frontend, mobile, chrome-extension, auth, deployment, design, devops, testing). Each pack declares `description`, `depends_on` (list — may be empty), `skills` (list of skill names), `agents` (list of agent names). Each entry inside packs follows the standard 6-field schema with `tier: optional-pack:<name>`. | REQ-005, REQ-006, REQ-009, REQ-010, REQ-011, REQ-012 | T-001 | internal/template/catalog.yaml | pending |
| T-005 | M2.3 — Populate `catalog.harness_generated` section with builder-harness and any `my-harness-*` placeholder entries. Each entry: `name`, `path`, `tier: harness-generated`, `hash`, `version: "1.0.0"`. | REQ-005, REQ-006, REQ-009 | T-001 | internal/template/catalog.yaml | pending |
| T-006 | M2.6 — Declare 9-pack `depends_on` graph (e.g., `design.depends_on: [frontend]`, `deployment.depends_on: [backend]`; backend/frontend/mobile/chrome-extension/auth/devops/testing depend on `[]`). Manually verify acyclicity (audit test T-015 will enforce automatically). | REQ-010, REQ-018 | T-004 | internal/template/catalog.yaml | pending |
| T-007 | M2.4 — Implement `gen-catalog-hashes.go` offline helper (CLI flags `--entry <name>` / `--all` / `--check`). Walk embedded FS for each catalog entry's `path`, normalize content (LF endings, trailing whitespace trim), compute sha256, output 64-char lowercase hex. ~80-120 LOC. | REQ-007, REQ-022 | T-001, T-002, T-003, T-004, T-005 | internal/template/scripts/gen-catalog-hashes.go | pending |
| T-008 | M2.4b — Invoke `gen-catalog-hashes.go --all` (manually via `go run`) to populate `hash` field for all 65 entries in `catalog.yaml`. Commit the resulting manifest with all hashes filled. Also set top-level `version: "1.0.0"` and `generated_at: "2026-05-12"`. | REQ-002, REQ-003, REQ-007, REQ-008 | T-007 | internal/template/catalog.yaml | pending |
| T-009 | M3.1 — Create `catalog_tier_audit_test.go` skeleton: package, imports (`crypto/sha256`, `encoding/hex`, `io/fs`, `regexp`, `strings`, `testing`, `gopkg.in/yaml.v3`), reuse `EmbeddedTemplates()` pattern from `lang_boundary_audit_test.go`. Define shared helpers (`walkSkills`, `walkAgents`, `parseManifest`). | REQ-014 | - | internal/template/catalog_tier_audit_test.go | pending |
| T-010 | M3.2 — `TestCatalogManifestPresent` (REQ-026): assert embedded FS contains `catalog.yaml` at expected internal path. Sentinel `CATALOG_MANIFEST_ABSENT`. | REQ-026 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-011 | M3.3 — `TestAllSkillsInCatalog` (REQ-005, REQ-015): walk `.claude/skills/` for 37 top-level directories, compare against union of catalog tier sections. Sentinel `CATALOG_ENTRY_MISSING: <path>`. | REQ-005, REQ-015 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-012 | M3.4 — `TestAllAgentsInCatalog` (REQ-006, REQ-016): walk `.claude/agents/moai/*.md` for 28 files, compare against union. Sentinel `CATALOG_ENTRY_MISSING: <path>`. | REQ-006, REQ-016 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-013 | M3.5 — `TestCatalogReferencesValid` (REQ-017): assert every catalog `path` field resolves to an existing entry in embedded FS. Sentinel `CATALOG_ENTRY_ORPHAN: <path>`. | REQ-017 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-014 | M3.6 — `TestCatalogTierValid` (REQ-019): regex-validate each entry's `tier` field against `^(core\|optional-pack:[a-z][a-z0-9-]{1,30}\|harness-generated)$`. Sentinel `CATALOG_TIER_INVALID: <entry> tier=<value>`. | REQ-019 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-015 | M3.7 — `TestPackDependencyDAG` (REQ-018): DFS-based cycle detection over `catalog.optional_packs.*.depends_on`. Sentinel `PACK_DEPENDENCY_CYCLE: <A> <-> <B>`. | REQ-018 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-016 | M3.8 — `TestManifestHashStability` (REQ-020, REQ-022, REQ-023): assert (a) every `hash` matches `^[0-9a-f]{64}$`, (b) recomputing sha256 over the normalized source content (using same algorithm as `gen-catalog-hashes.go`) equals the stored `hash`. Sentinels `CATALOG_HASH_INVALID: <entry>`, `CATALOG_HASH_UNSTABLE: <entry>`. CRITICAL for Windows CI (CRLF/LF normalization). | REQ-020, REQ-022, REQ-023 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-017 | M3.9 — `TestWorkflowTriggerCoverage` (REQ-021, conditional vacuously-true): walk 20 flat `.md` workflows under `.claude/skills/moai/workflows/`, parse frontmatter, if `metadata.required-skills` key absent → pass (vacuously true); if present → resolve each listed skill against catalog and verify tier is `core` or pack declared in `metadata.required-packs`. Sentinel `WORKFLOW_UNCOVERED: <workflow> requires <skill>`. v1 expectation: all 20 pass vacuously. | REQ-021 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-018 | M3.10 — `TestCatalogNoDuplicateEntries` (REQ-027): compute the set union of all entry names across `catalog.core.{skills,agents}`, `catalog.optional_packs.*.{skills,agents}`, `catalog.harness_generated.{skills,agents}`; if any name appears in 2+ sections, fail. Sentinel `CATALOG_DUPLICATE_ENTRY: <name> in [<tier1>, <tier2>]`. Also add `TestCatalogReservedFieldType` (REQ-024): when `marketplace_id`/`marketplace_url`/`publisher` present on a pack, assert string type. Sentinel `CATALOG_RESERVED_FIELD_INVALID: <pack> <field>`. | REQ-024, REQ-027 | T-009 | internal/template/catalog_tier_audit_test.go | pending |
| T-019 | M4.1 — Implement `catalog_loader.go` with `LoadCatalog(fsys fs.FS) (*Catalog, error)`: read `internal/template/catalog.yaml` via embed.FS, `yaml.Unmarshal` into typed Catalog struct. Tier enum centralized as Go const (`TierCore`, `TierOptionalPack`, `TierHarnessGenerated`) per CLAUDE.local.md §14 no-hardcoding rule. ~80-100 LOC. | REQ-001, REQ-004 | T-008 | internal/template/catalog_loader.go | pending |
| T-020 | M4.2 — Define `Catalog`, `Entry`, `Pack` struct types in `catalog_loader.go`: fields per schema (Version, GeneratedAt, Core{Skills,Agents}, OptionalPacks map[string]*Pack, HarnessGenerated{Skills,Agents}). yaml tags for v3 unmarshal. | REQ-004, REQ-011 | T-019 | internal/template/catalog_loader.go | pending |
| T-021 | M4.3 — Add `Catalog.LookupSkill(name string) (*Entry, bool)` + `Catalog.LookupAgent(name string) (*Entry, bool)` accessors. Iterates all tier sections; returns first match + boolean ok. Used by audit suite (T-013/T-017) and follow-up SPEC-002+003. | REQ-005, REQ-006, REQ-017 | T-020 | internal/template/catalog_loader.go | pending |
| T-022 | M4.4 — Refactor T-009..T-018 audit tests to call `LoadCatalog()` instead of inline `yaml.Unmarshal`. Single source-of-truth for parsing logic; future schema changes touch only `catalog_loader.go`. Verify all 10 sub-tests still GREEN. | REQ-001, REQ-014 | T-018, T-021 | internal/template/catalog_tier_audit_test.go | pending |
| T-023 | M4.5 — Read-only regression check: run `go test ./internal/template/...` and assert `deployer_test.go`, `deployer_mode_test.go`, `deployer_transactional_test.go` all pre-existing test cases remain GREEN. This is a verification task — NO source modification of `deployer.go`. If `catalog.yaml` is not auto-included by `//go:embed all:templates`, add additive `//go:embed catalog.yaml` directive to `embed.go` (single-line additive change, no removal). | REQ-025 | T-022 | internal/template/embed.go (conditional, only if embed gap found) | pending |
| T-024 | M5.1 — Write `catalog_doc.md`: per-field schema documentation, tier semantics (core/optional-pack/harness-generated), hash normalization spec (LF endings + trailing whitespace trim → sha256 hex), cross-reference to follow-up SPECs (SPEC-V3R4-CATALOG-002..007). ~60-80 lines markdown. | REQ-007 | T-022 | internal/template/catalog_doc.md | pending |
| T-025 | M5.2 — Add YAML comments to `catalog.yaml`: section headers for `catalog.core`, `catalog.optional_packs`, `catalog.harness_generated` with short description + link to `.moai/brain/IDEA-003/proposal.md`. No data changes — comments only. | REQ-001 | T-024 | internal/template/catalog.yaml | pending |
| T-026 | Final — Run full `go test ./...` + `golangci-lint run` + `go vet ./...`. Verify all 10 audit sub-tests GREEN, all deployer regression tests GREEN, lint zero warnings on new files. Confirm Quality Gates checklist in acceptance.md. | REQ-014 | T-023, T-025 | (no new files — verification only) | pending |

## Task Count Summary

- **Total tasks**: 26 (within the ≤10 advisory cap is exceeded by SPEC plan-derived structure; this 26-task decomposition mirrors M1-M5's ~28 atomic sub-steps; cap waived because tasks are tightly atomic with clear single-cycle scope per CLAUDE.md §7 Rule 2 multi-file decomposition for SPECs with 5+ planned files).
- **By milestone**: M1 (1) + M2 (7) + M3 (10) + M4 (5) + M5 (2) + Final (1) = 26.
- **TDD distribution**: RED-style audit-driven tasks T-009..T-018 (10 sub-tests written before manifest data is final); GREEN-style data lock-in T-002..T-008; refactor/integration T-019..T-023; documentation T-024..T-025; verification T-026.

## Planned Files Summary (Phase 2B Drift Guard input)

| File | Owner Tasks | LOC Estimate |
|------|-------------|--------------|
| `internal/template/catalog.yaml` | T-001 (skeleton), T-002, T-003, T-004, T-005, T-006, T-008, T-025 | 800-1200 YAML lines (~80-120KB) |
| `internal/template/catalog_tier_audit_test.go` | T-009 (skeleton), T-010..T-018, T-022 (refactor) | 450-550 LOC |
| `internal/template/catalog_loader.go` | T-019 (LoadCatalog), T-020 (struct), T-021 (accessors) | 150-200 LOC |
| `internal/template/scripts/gen-catalog-hashes.go` | T-007 | 80-120 LOC |
| `internal/template/catalog_doc.md` | T-024 | 60-80 lines markdown |
| `internal/template/embed.go` | T-023 (conditional additive — only if `//go:embed all:templates` does not cover `catalog.yaml` at parent path) | +1 line conditional |

**Total**: 5 NEW files + 1 conditional MODIFY. ~1500-1900 LOC across Go + YAML + Markdown.

## Dependency Graph

```
T-001 (schema skeleton)
  ├─ T-002 (core skills)        ──┐
  ├─ T-003 (core agents)        ──┤
  ├─ T-004 (optional packs)     ──┤
  └─ T-005 (harness)            ──┤
                                  ├─→ T-007 (hash helper)
                                  │     └─→ T-008 (populate hashes)
                                  │           └─→ T-019 (LoadCatalog)
                                  │                 └─→ T-020 (struct types)
                                  │                       └─→ T-021 (accessors)
                                  │                             └─→ T-022 (refactor audit)
                                  │                                   └─→ T-023 (regression)
                                  │                                         └─→ T-024 (docs)
                                  │                                               └─→ T-025 (comments)
                                  │                                                     └─→ T-026 (final)
                                  └─→ T-006 (depends_on graph)

T-009 (audit skeleton)
  ├─ T-010..T-018 (10 audit sub-tests — independent, parallel)
  └─→ T-022 (refactor to LoadCatalog) [merge point]
```

Notes:
- T-009..T-018 are independent (each adds one sub-test to a separate Go function) and can be developed in parallel.
- T-022 is the merge point: it depends on T-018 (last audit test) AND T-021 (accessors), then refactors all 10 sub-tests to call `LoadCatalog`.
- T-026 is the final verification gate before /moai sync.

## Status Legend

- `pending`: not started
- `in-progress`: currently being implemented
- `red`: TDD RED phase complete (test written, failing as expected)
- `green`: TDD GREEN phase complete (test passing)
- `refactor`: refactor phase complete
- `done`: task fully complete, audit GREEN
- `blocked`: cannot proceed — see notes
