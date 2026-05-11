# SPEC-V3R4-CATALOG-001 — Compact

> Auto-extracted from spec.md + acceptance.md. ~30% token savings for `/moai run`.
> **v0.2.0**: 11 defects (D1–D11) 반영. 카운트 통일 (37 skills + 28 agents = 65 entries), REQ-027 신규 (duplicate sentinel REQ 보호), REQ-021 vacuously-true (workflows flat-md layout), Scenario 6/7 신규 (untracked REQ → AC mapping), Quality Gates 에서 self-referential plan-auditor 분리.

## Catalog Entry Definition

"Catalog entry" 는 `internal/template/templates/.claude/skills/` 바로 아래의 **top-level 디렉토리 1개** (37개; `moai/` 컨테이너는 단일 logical skill, 내부 `workflows/` / `team/` / `references/` 는 모듈) 또는 `internal/template/templates/.claude/agents/moai/` 바로 아래의 **`.md` 파일 1개** (28개) 를 의미한다. **총 65 entries**.

## EARS Requirements (27)

### Manifest Existence and Schema (Ubiquitous)
- REQ-CATALOG-001-001: Single source of truth manifest at `internal/template/catalog.yaml`.
- REQ-CATALOG-001-002: Top-level `version` field (semver string).
- REQ-CATALOG-001-003: `generated_at` field (ISO 8601 date).
- REQ-CATALOG-001-004: Top-level `catalog` with three sub-sections: `core`, `optional_packs`, `harness_generated`.
- REQ-CATALOG-001-005: Every skill under `templates/.claude/skills/` appears once with valid `tier`.
- REQ-CATALOG-001-006: Every agent under `templates/.claude/agents/moai/` appears once with valid `tier`.

### Per-Entry Drift Detection Metadata (Ubiquitous, sourced from user constraint)
- REQ-CATALOG-001-007: Each entry has `hash` (sha256 of normalized content — LF, no trailing ws). Skill entries' hash covers root `SKILL.md` / `skill.md` only; sub-files not covered (follow-up SPEC).
- REQ-CATALOG-001-008: Each entry has `version` (semver, for SPEC-004 drift signal).
- REQ-CATALOG-001-009: Each entry has `path` (relative to `templates/`).
- REQ-CATALOG-001-010: Each optional-pack entry has `depends_on` (list, possibly empty).

### Pack Definitions (Ubiquitous)
- REQ-CATALOG-001-011: Each pack under `catalog.optional_packs.<name>` with `description`, `depends_on`, `skills`, `agents`.
- REQ-CATALOG-001-012: Pack names match regex `^[a-z][a-z0-9-]{1,30}$`.

### Tier Mutation Constraint (Unwanted Behavior) — D1 권장 수정 반영
- REQ-CATALOG-001-013: If a contributor changes an entry's `tier` field, plan-auditor SHALL flag the change as requiring explicit SPEC ID amendment reference in PR description (process invariant, plan-auditor enforced at PR review time; not runtime audit).

### Audit Suite (Ubiquitous)
- REQ-CATALOG-001-014: `internal/template/catalog_tier_audit_test.go` fails CI on integrity violation.
- REQ-CATALOG-001-015: Every embedded skill has catalog entry. Sentinel: `CATALOG_ENTRY_MISSING: <path>`.
- REQ-CATALOG-001-016: Every embedded agent has catalog entry. Sentinel: `CATALOG_ENTRY_MISSING: <path>`.
- REQ-CATALOG-001-017: Every catalog entry references existing path. Sentinel: `CATALOG_ENTRY_ORPHAN: <path>`.
- REQ-CATALOG-001-018: Pack `depends_on` graph is acyclic (DAG). Sentinel: `PACK_DEPENDENCY_CYCLE: <A> <-> <B>`.
- REQ-CATALOG-001-019: Tier values match allowed regex. Sentinel: `CATALOG_TIER_INVALID: <entry> tier=<value>`.
- REQ-CATALOG-001-020: `hash` matches `^[0-9a-f]{64}$`. Sentinel: `CATALOG_HASH_INVALID: <entry>`.
- **REQ-CATALOG-001-027** (D2 권장 수정 — 신규): Skill/agent appears in >1 tier section. Sentinel: `CATALOG_DUPLICATE_ENTRY: <name> in [<tier1>, <tier2>]`.

### Workflow Trigger Coverage (Event-Driven, conditional) — D4 권장 수정 반영
- REQ-CATALOG-001-021: When workflow's frontmatter declares `metadata.required-skills` list, audit verifies every listed skill resolves to `core` or explicitly declared `optional-pack:<name>` in `metadata.required-packs`. Sentinel: `WORKFLOW_UNCOVERED: <workflow> requires <skill>`. v1 status: 20 workflows are flat `.md` files with no `metadata.required-skills` key → all vacuously true.

### Hash Stability (State-Driven)
- REQ-CATALOG-001-022: Same content → same hash (stability).
- REQ-CATALOG-001-023: Content changes → hash differs.

### Marketplace Forward-Compat (Optional) — D9 권장 수정 반영
- REQ-CATALOG-001-024: Manifest MAY include reserved optional fields `marketplace_id`, `marketplace_url`, `publisher` at pack level (no current verification requirement). If present, audit SHALL verify type (string). Sentinel: `CATALOG_RESERVED_FIELD_INVALID: <pack> <field>`.

### Backward Compatibility (Unwanted Behavior)
- REQ-CATALOG-001-025: `moai init` deploy pipeline unchanged by manifest introduction.
- REQ-CATALOG-001-026: catalog.yaml absent from binary → fail with `CATALOG_MANIFEST_ABSENT`.

## Acceptance Criteria (7 Given/When/Then + 4 Edge Cases) — D6 권장 수정 반영

### Scenario 1: All 37 skills + 28 agents in catalog with valid tier
- G: 37 skills + 28 agents on disk (verified via find/grep)
- W: `TestAllSkillsInCatalog` + `TestAllAgentsInCatalog` + audit suite present (T3.2)
- T: Each entry appears exactly once with `tier` matching `^(core|optional-pack:[a-z][a-z0-9-]{1,30}|harness-generated)$` + valid `path` field. Sentinels: `CATALOG_ENTRY_MISSING` / `CATALOG_TIER_INVALID`.
- Maps: REQ-005, REQ-006, REQ-009, REQ-014, REQ-015, REQ-016, REQ-019.

### Scenario 2: Hash field enables drift detection (foundation for SPEC-004)
- G: Manifest v1.0.0 entry hash=A
- W: Future v1.1.0 entry hash=B (content changed)
- T: SPEC-004 can detect A≠B as drift. Hash regex `^[0-9a-f]{64}$`. Same content → same hash. Sentinel: `CATALOG_HASH_INVALID` / `CATALOG_HASH_UNSTABLE`.
- Maps: REQ-007, REQ-008, REQ-020, REQ-022, REQ-023.
- User constraint reflection: foundation for "moai update 손실 0".

### Scenario 3: Pack dependency graph is acyclic (DAG invariant)
- G: design→[frontend], frontend→[], deployment→[backend], backend→[]
- W: `TestPackDependencyDAG` runs DFS
- T: Linear DAG passes. Cycle (frontend→[design]) fails with `PACK_DEPENDENCY_CYCLE: design <-> frontend`.
- Maps: REQ-010, REQ-018.

### Scenario 4: Workflow Trigger Coverage — vacuously true at v1 (D4)
- G: 20 flat `.md` workflows under `moai/workflows/` (no subdirectory + SKILL.md), `grep required-skills` → 0 matches
- W: `TestWorkflowTriggerCoverage` parses frontmatter conditionally
- T: Workflow without `metadata.required-skills` key → pass vacuously. Workflow with the key → resolve each dep to `core` or declared `optional-pack:*` in `metadata.required-packs`. v1: all 20 workflows pass vacuously. Future SPEC retrofit auto-activates. Sentinel: `WORKFLOW_UNCOVERED`.
- Maps: REQ-021.

### Scenario 5: Schema validation rejects malformed entries (D1 + D2 + D9)
- G: Manual edit `tier: "invalid-tier-name"`, `hash: "shorthash"`, tier change w/o SPEC amendment, `marketplace_id: 12345` (integer)
- W: `LoadCatalog()` or audit runs, plan-auditor reviews PR
- T: Reject with `CATALOG_TIER_INVALID` + `CATALOG_HASH_INVALID` + `CATALOG_RESERVED_FIELD_INVALID`. plan-auditor flags tier-change amendment requirement.
- Maps: REQ-013, REQ-019, REQ-020, REQ-024.

### Scenario 6: Manifest top-level structure is valid (D6 권장 수정 — 신규)
- G: Binary build embeds `catalog.yaml`
- W: `LoadCatalog()` + `TestCatalogManifestPresent` (T3.2)
- T: Exactly 1 manifest at fixed path. Top-level `version` (semver), `generated_at` (ISO 8601), `catalog` with 3 sub-sections. Audit suite file exists. Reserved optional fields type-checked when present. `moai init` deploy pipeline unchanged. Missing catalog.yaml → `CATALOG_MANIFEST_ABSENT`.
- Maps: REQ-001, REQ-002, REQ-003, REQ-004, REQ-014, REQ-024, REQ-025, REQ-026.

### Scenario 7: Pack definition structure is valid (D6 권장 수정 — 신규)
- G: 9 optional packs (backend, frontend, mobile, chrome-extension, auth, deployment, design, devops, testing)
- W: `LoadCatalog()` + schema validation
- T: Each pack declares `description`, `depends_on`, `skills`, `agents`. Pack names match regex `^[a-z][a-z0-9-]{1,30}$`.
- Maps: REQ-011, REQ-012.

### Edge Cases
- **EC1**: Skill on disk missing from catalog → `CATALOG_ENTRY_MISSING`. Recovery: add entry + hash via `gen-catalog-hashes.go` helper.
- **EC2**: Catalog entry references non-existent skill → `CATALOG_ENTRY_ORPHAN` (REQ-017). Recovery: remove orphan in same PR.
- **EC3**: `hash: ""` or `hash: "TODO"` → `CATALOG_HASH_INVALID`. Recovery: `internal/template/scripts/gen-catalog-hashes.go --entry <name>` helper. Maps: REQ-007, REQ-009, REQ-020, REQ-022.
- **EC4**: Same skill in multiple tier sections → `CATALOG_DUPLICATE_ENTRY: <name> in [<tier1>, <tier2>]` (REQ-027, D2). Recovery: remove duplicate.

## Files to Modify / Create

- [NEW] `internal/template/catalog.yaml` — Single source of truth (80-120KB, 800-1200 lines, 65 entries × ~6 fields + 9 packs + comments).
- [NEW] `internal/template/catalog_tier_audit_test.go` — Audit suite (~450-550 LOC, 10 parallel sub-tests T3.1–T3.10, sentinel-based, includes REQ-027 duplicate test).
- [NEW] `internal/template/catalog_loader.go` — YAML parser + accessor (`LoadCatalog`, `Catalog.LookupSkill`, `Catalog.LookupAgent`). ~150-200 LOC.
- [NEW] `internal/template/scripts/gen-catalog-hashes.go` — Offline helper to compute sha256 hashes for catalog.yaml entries (D5 권장 수정 반영 — 이전 v0.1.0 누락 등록). Invoked via `go run internal/template/scripts/gen-catalog-hashes.go [--entry <name>] [--all]`. Normalization: LF + trailing whitespace trim. Used in EC1/EC3 recovery + M2-T2.4 initial hash lock-in. ~80-120 LOC.
- [NEW] `internal/template/catalog_doc.md` — Schema docstring + hash normalization spec + cross-reference to SPEC-002…007. ~60-80 lines.
- [MODIFY] `internal/template/embed.go` — **No source changes anticipated**. Existing `//go:embed all:templates` covers external `catalog.yaml` indirectly. Audit-time confirmation only (REQ-026). If gap found, add additive `//go:embed catalog.yaml` directive.
- [NO MODIFY] `internal/template/deployer.go` — **D7 권장 수정 반영**: SPEC-001 scope 에서 미수정. catalog.yaml load 는 audit + catalog_loader.go 만 사용. Deploy 흐름 manifest 의식 안 함. Tier filtering 은 SPEC-002+003.

## Exclusions (Out of Scope)

- Directory relocation → SPEC-V3R4-CATALOG-002
- `moai pack` CLI → SPEC-V3R4-CATALOG-003
- `moai update --catalog-sync` 안전 동기화 → SPEC-V3R4-CATALOG-004 (Wave 3, 최고위험)
- `/moai project` 인터뷰 확장 → SPEC-V3R4-CATALOG-005
- `moai doctor catalog` → SPEC-V3R4-CATALOG-006
- 4개국어 migration docs → SPEC-V3R4-CATALOG-007
- Tier filtering at deploy time → SPEC-002
- Workflow `metadata.required-skills` retrofit → future SPEC (post-CATALOG-007). v1 manifest 는 vacuously true (D4).
- Anthropic Marketplace publishing → future SPEC (TBD)
- Skill frontmatter `tier:` field → manifest is sole source of truth
- Skill module sub-files hash coverage → entry root `SKILL.md` / `skill.md` 만, sub-files (`workflows/*.md`, `references/*`) 별도 hash 미지원 (후속 SPEC)
- Idle skill telemetry → SPEC-006

## Dependencies

- Depends on: (none — Wave 1 첫 SPEC)
- Blocks: SPEC-V3R4-CATALOG-002 ~ 007 (모두 본 manifest 위에 빌드)
