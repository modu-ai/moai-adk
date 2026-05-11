# Acceptance Criteria — SPEC-V3R4-CATALOG-001

> **v0.2.0 (plan-auditor iter 1 FAIL 0.72 → 11 defects 반영)**: 카운트 통일 (37 skills + 28 agents = 65 entries), Scenario 6/7 신규 (D6 untracked REQ → AC mapping 보강), Scenario 5 매핑에 REQ-027 추가 (D2), Scenario 4 vacuously-true 명확화 (D4), Quality Gates self-referential plan-auditor 제거 (D10).

Given-When-Then scenarios for the 3-tier catalog manifest. Scenarios derived from EARS Requirements in `spec.md`. Edge cases derived from `research.md` §"Risks & Constraints".

> **Catalog Entry 정의 (spec.md §Overview 와 일관성)**: "Catalog entry" 는 `internal/template/templates/.claude/skills/` 바로 아래 top-level 디렉토리 1개 (**37개**) 또는 `internal/template/templates/.claude/agents/moai/` 바로 아래 `.md` 파일 1개 (**28개**) 를 의미한다. `moai/` 컨테이너는 단일 logical skill 로 취급. **총 65 entries**.

## Scenario 1: Manifest contains all 37 skills + 28 agents with valid tier classification

**Given** the codebase has **37 top-level skill directories** under `internal/template/templates/.claude/skills/` (each containing `SKILL.md` or `skill.md` for legacy reference skills) and **28 agent markdown files** under `internal/template/templates/.claude/agents/moai/` (verified in `research.md` §"Architecture Analysis" + disk verification by `find ... -maxdepth 1 -mindepth 1 -type d | wc -l` = 37 / `find ... -name '*.md' | wc -l` = 28),

**When** the audit test `TestAllSkillsInCatalog` AND `TestAllAgentsInCatalog` execute against the embedded FS,

**Then** every skill directory MUST appear exactly once in `catalog.yaml` under one of three sections (`catalog.core.skills`, `catalog.optional_packs.<pack>.skills`, `catalog.harness_generated.skills`), AND every agent file MUST similarly appear exactly once. Each entry MUST have a `tier` field whose value matches the regex `^(core|optional-pack:[a-z][a-z0-9-]{1,30}|harness-generated)$`. Each entry MUST also have a `path` field (REQ-009) identifying its relative location under `templates/`. Missing entries MUST cause test failure with sentinel `CATALOG_ENTRY_MISSING: <relative-path>`. Invalid tier values MUST cause failure with sentinel `CATALOG_TIER_INVALID: <entry> tier=<value>`. The audit suite as a whole (REQ-014, REQ-015, REQ-016) MUST exist as `internal/template/catalog_tier_audit_test.go`.

> Maps: REQ-CATALOG-001-005, REQ-CATALOG-001-006, REQ-CATALOG-001-009, REQ-CATALOG-001-014, REQ-CATALOG-001-015, REQ-CATALOG-001-016, REQ-CATALOG-001-019.

## Scenario 2: Hash field enables drift detection (foundation for SPEC-004)

**Given** an existing project deployed manifest version `1.0.0` with skill `moai-workflow-spec` recorded as `hash: "abc123...def"` (sha256, 64 lowercase hex chars),

**When** moai-adk-go releases a future manifest version `1.1.0` where the same skill has `hash: "789xyz...456"` (upstream content changed),

**Then** the manifest schema MUST allow SPEC-V3R4-CATALOG-004 (`moai update --catalog-sync`) to detect `"abc123..." != "789xyz..."` as a drift signal. The hash field MUST be the sha256 digest of the normalized source file content (LF line endings, trailing whitespace trimmed, no other transformations). The audit test `TestManifestHashStability` MUST verify that:

- Each entry's `hash` matches regex `^[0-9a-f]{64}$` (sentinel `CATALOG_HASH_INVALID: <entry>` on mismatch).
- Re-computing the hash on the same source content yields the same value (golden test; sentinel `CATALOG_HASH_UNSTABLE` on instability — required for cross-platform CI Windows/macOS/Linux).

Hash algorithm is **sha256** (selected via OQ6 권장안). The manifest also MUST record `version: "1.0.0"` per entry to provide a semver-based fallback signal when hash collisions or content-equivalent re-encoding occurs.

> Maps: REQ-CATALOG-001-007, REQ-CATALOG-001-008, REQ-CATALOG-001-020, REQ-CATALOG-001-022, REQ-CATALOG-001-023.
>
> **User constraint reflection**: 사용자 추가 제약 ("moai update 손실 0") 의 foundation. 본 SPEC 단독으로는 drift 비교 / 사용자 확인 / 자동 백업을 수행하지 않으나, SPEC-V3R4-CATALOG-004 가 본 hash + version 필드를 입력으로 사용하여 안전 동기화 구현.

## Scenario 3: Pack dependency graph is acyclic (DAG invariant)

**Given** the manifest declares the following pack dependencies (example, final assignment in M2):
- pack `design` declares `depends_on: [frontend]`
- pack `frontend` declares `depends_on: []`
- pack `deployment` declares `depends_on: [backend]`
- pack `backend` declares `depends_on: []`

**When** the audit test `TestPackDependencyDAG` runs DFS-based cycle detection over the declared `depends_on` graph,

**Then** the test MUST pass when the graph is a linear DAG (no back-edge in DFS). If a future contributor introduces `pack frontend: depends_on: [design]` (creating cycle `design → frontend → design`), the audit MUST fail with sentinel `PACK_DEPENDENCY_CYCLE: design <-> frontend` and the CI build MUST block the merge.

> Maps: REQ-CATALOG-001-010, REQ-CATALOG-001-018.

## Scenario 4: Workflow Trigger Coverage — vacuously true at v1 (D4 권장 수정 반영)

**Given** the current `internal/template/templates/.claude/skills/moai/workflows/` directory contains **20 flat `.md` files** (e.g., `plan.md`, `run.md`, `sync.md`, `brain.md`, `clean.md`, `codemaps.md`, `coverage.md`, `db.md`, `design.md`, `e2e.md`, `feedback.md`, `fix.md`, `gate.md`, `github.md`, `loop.md`, `moai.md`, `mx.md`, `project.md`, `review.md`, `security.md`) — NOT subdirectories with `SKILL.md`. Empirical disk check: `grep -r 'required-skills' .claude/skills/moai/workflows/` → **0 matches** (no workflow currently declares `metadata.required-skills` frontmatter),

**When** the audit test `TestWorkflowTriggerCoverage` parses each workflow `.md` file's frontmatter,

**Then** the audit MUST behave as follows per REQ-021's conditional structure:
- If a workflow's frontmatter declares `metadata.required-skills` (list of skill names), the audit MUST resolve each listed skill against the catalog AND verify the skill's `tier` is either `core` OR `optional-pack:<name>` where the pack is explicitly declared in the workflow's `metadata.required-packs` field, emitting sentinel `WORKFLOW_UNCOVERED: <workflow> requires <skill>` on violation.
- If a workflow's frontmatter has no `metadata.required-skills` key, the audit MUST pass vacuously (conditional REQ not triggered).
- At v1 manifest milestone, **all 20 workflows are vacuously true** (none declare the key). Audit GREEN.
- Future SPEC retrofit (post-CATALOG-007) MAY add `metadata.required-skills` to specific workflows; the test then activates substantive verification automatically without code change.

This treatment is consistent with REQ-021's Event-Driven (conditional) EARS pattern.

> Maps: REQ-CATALOG-001-021.
>
> **Note**: This scenario guards against future regression where a contributor moves a heretofore-core skill into `optional-pack:*`, silently breaking a workflow that depended on it. The audit will catch this at CI time once `metadata.required-skills` is retrofitted into workflows. v1 establishes the conditional check infrastructure without breaking the current flat-md layout.

## Scenario 5: Schema validation rejects malformed entries

**Given** a contributor manually edits `catalog.yaml` to introduce (a) an invalid entry with `tier: "invalid-tier-name"` AND `hash: "shorthash"` (not 64 hex chars), AND (b) a tier change on an existing entry without corresponding SPEC ID amendment reference in the PR description, AND (c) a hypothetical optional pack with malformed `marketplace_id: 12345` (integer instead of string),

**When** the manifest is loaded by `internal/template/catalog_loader.go` `LoadCatalog()` OR audit tests execute, AND plan-auditor reviews the PR,

**Then** the system MUST reject the invalid changes by emitting:
- `CATALOG_TIER_INVALID: <skill-name> tier=invalid-tier-name — allowed: [core, optional-pack:<name>, harness-generated]` (REQ-CATALOG-001-019),
- `CATALOG_HASH_INVALID: <skill-name> hash=shorthash — expected 64 lowercase hex chars (sha256)` (REQ-CATALOG-001-020),
- plan-auditor flags the tier-change PR as requiring an explicit SPEC ID amendment reference in the PR description (REQ-CATALOG-001-013), AND
- `CATALOG_RESERVED_FIELD_INVALID: <pack> marketplace_id` on type violation of reserved optional fields when present (REQ-CATALOG-001-024).

The CI build MUST fail on the audit-detectable items. plan-auditor flags the tier-amendment requirement as PR review feedback (process invariant, not runtime audit).

> Maps: REQ-CATALOG-001-013, REQ-CATALOG-001-019, REQ-CATALOG-001-020, REQ-CATALOG-001-024.

## Scenario 6: Manifest top-level structure is valid (D6 권장 수정 반영 — 신규)

**Given** the binary build embeds `internal/template/catalog.yaml` at expected path,

**When** the manifest is loaded by `LoadCatalog()` AND audit test `TestCatalogManifestPresent` (T3.2) runs,

**Then** the system MUST verify all of the following structural invariants:
- Exactly one manifest file exists at `internal/template/catalog.yaml` as single source of truth (REQ-CATALOG-001-001).
- Top-level `version` field is a semver string (e.g., `"1.0.0"`) (REQ-CATALOG-001-002).
- Top-level `generated_at` field is an ISO 8601 date (e.g., `"2026-05-12"`) (REQ-CATALOG-001-003).
- Top-level `catalog` object contains exactly three sub-sections: `core`, `optional_packs`, `harness_generated` (REQ-CATALOG-001-004).
- The audit suite file `internal/template/catalog_tier_audit_test.go` exists and fails CI on integrity violation (REQ-CATALOG-001-014).
- The manifest schema MAY include reserved optional fields `marketplace_id`, `marketplace_url`, `publisher` at the pack level; when present, audit MUST verify type (string) (REQ-CATALOG-001-024).
- Running `moai init` with the manifest present deploys every skill and agent through the existing Deployer interface without breaking the pipeline (REQ-CATALOG-001-025; verified via M4-T4.5 read-only check of `internal/template/deployer_test.go` existing test cases all GREEN).
- If `catalog.yaml` is missing from the binary build, the audit MUST fail with sentinel `CATALOG_MANIFEST_ABSENT` (REQ-CATALOG-001-026).

> Maps: REQ-CATALOG-001-001, REQ-CATALOG-001-002, REQ-CATALOG-001-003, REQ-CATALOG-001-004, REQ-CATALOG-001-014, REQ-CATALOG-001-024, REQ-CATALOG-001-025, REQ-CATALOG-001-026.

## Scenario 7: Pack definition structure is valid (D6 권장 수정 반영 — 신규)

**Given** the manifest declares 9 optional packs under `catalog.optional_packs.<pack-name>` (per OQ2 권장안: `backend`, `frontend`, `mobile`, `chrome-extension`, `auth`, `deployment`, `design`, `devops`, `testing`),

**When** the manifest is loaded by `LoadCatalog()` AND audit tests run schema validation,

**Then** for each pack the system MUST verify:
- The pack entry under `catalog.optional_packs.<pack-name>` declares all four required fields: `description` (string), `depends_on` (list of pack names, may be empty `[]`), `skills` (list of skill names), `agents` (list of agent names) (REQ-CATALOG-001-011).
- The pack name conforms to the regex `^[a-z][a-z0-9-]{1,30}$` (lowercase, hyphen-separated, length 2-31) (REQ-CATALOG-001-012).

Invalid pack names or missing required fields MUST cause schema validation failure. Optional reserved fields (`marketplace_id`, `marketplace_url`, `publisher`) are validated for type only when present (cross-link to Scenario 6).

> Maps: REQ-CATALOG-001-011, REQ-CATALOG-001-012.

## Edge Cases

### EC1: Skill exists on disk but missing from catalog.yaml

- **Condition**: A new top-level skill directory `moai-domain-newfeature` is added under `.claude/skills/moai-domain-newfeature/SKILL.md` (per the catalog entry definition: a top-level skills/ directory containing SKILL.md), but the contributor forgets to register it in `catalog.yaml`.
- **Behavior**: `TestAllSkillsInCatalog` (T3.3) fails with sentinel `CATALOG_ENTRY_MISSING: .claude/skills/moai-domain-newfeature`. CI blocks the merge.
- **Recovery**: Contributor adds an entry under appropriate tier section (e.g., `catalog.core.skills` or `catalog.optional_packs.<pack>.skills`), runs `gen-catalog-hashes.go` helper to populate the `hash` field, commits the manifest update with the SKILL.md change in the same PR.
- **Diagnostic**: Sentinel message includes the missing path so the contributor can locate the file. README / CONTRIBUTING.md documents the manifest update workflow (M5 task).
- **Maps**: REQ-CATALOG-001-005, REQ-CATALOG-001-015.

### EC2: Catalog entry references non-existent skill (orphan)

- **Condition**: A contributor deletes `.claude/skills/moai-legacy-skill/` but forgets to remove the corresponding entry from `catalog.yaml`.
- **Behavior**: `TestCatalogReferencesValid` (T3.5) fails with sentinel `CATALOG_ENTRY_ORPHAN: <path>` (REQ-CATALOG-001-017). CI blocks the merge.
- **Recovery**: Contributor removes the orphan entry from `catalog.yaml` in the same PR as the file deletion.
- **Note**: Orphan entries cause skill resolution failures at deploy time (SPEC-002+003), so detecting them at PR time prevents production breakage.
- **Maps**: REQ-CATALOG-001-017.

### EC3: Hash field empty or wrong format (not sha256 hex)

- **Condition**: A contributor adds an entry but leaves `hash: ""` (placeholder forgotten) or `hash: "TODO"`.
- **Behavior**: `TestManifestHashStability` (T3.8) fails with sentinel `CATALOG_HASH_INVALID: <entry>`. Regex `^[0-9a-f]{64}$` mismatch.
- **Recovery**: Contributor runs `go run internal/template/scripts/gen-catalog-hashes.go --entry <name>` (offline helper, [NEW] file per spec.md §"Files to Modify / Create" — D5 권장 수정 반영) to compute the correct sha256 hash and updates the manifest. Helper also populates `path` field (REQ-CATALOG-001-009) for consistency.
- **Note**: Helper script computes hash on the normalized file content (LF endings, trailing whitespace trimmed). Documented in `catalog_doc.md` (M5).
- **Maps**: REQ-CATALOG-001-007 (hash), REQ-CATALOG-001-009 (path), REQ-CATALOG-001-020 (hash regex), REQ-CATALOG-001-022 (stability).

### EC4: Two entries claim the same skill or agent name (collision)

- **Condition**: Contributor accidentally adds `moai-domain-backend` to BOTH `catalog.core.skills` AND `catalog.optional_packs.backend.skills` (e.g., during a refactor).
- **Behavior**: YAML parsing succeeds (different parent paths). Audit test `TestCatalogNoDuplicateEntries` (T3.10) computes the union of all tier sections; if a single skill name appears in multiple sections, the union size differs from the per-section count sum. The test emits sentinel `CATALOG_DUPLICATE_ENTRY: <skill-name> in [<tier1>, <tier2>]` per REQ-CATALOG-001-027 (D2 권장 수정 반영 — orphan sentinel 의 REQ 보호).
- **Recovery**: Contributor removes the duplicate from one section. catalog_doc.md (M5) clarifies the invariant: "Every skill/agent MUST appear in exactly one tier section."
- **Maps**: REQ-CATALOG-001-027.

## Quality Gates

- [ ] `go test ./internal/template/...` passes all audit tests (T3.1–T3.10).
- [ ] `go vet ./internal/template/...` clean.
- [ ] `golangci-lint run` passes on new files (catalog_loader.go, catalog_tier_audit_test.go, scripts/gen-catalog-hashes.go).
- [ ] Test coverage for `internal/template/catalog_loader.go` ≥ 85% (per CLAUDE.local.md §6 critical-package threshold per package).
- [ ] All 7 Given-When-Then scenarios pass when run individually AND in `t.Parallel()` mode.
- [ ] All 4 edge cases (EC1–EC4) have corresponding negative test cases in `catalog_tier_audit_test.go`.
- [ ] No HARD rule violations:
  - Template-First (CLAUDE.local.md §2): catalog.yaml is in `internal/template/` (build artifact), not in `internal/template/templates/` (user-deployed content). Loader at `internal/template/catalog_loader.go`.
  - 16-language neutrality (CLAUDE.local.md §15): manifest is language-agnostic (no language-specific entries).
  - No hardcoded values (CLAUDE.local.md §14): tier enum values centralized as Go const in `catalog_loader.go`.
- [ ] `catalog.yaml` is valid YAML (parses with `gopkg.in/yaml.v3` cleanly).
- [ ] CI 6 jobs all green: Lint, Test (ubuntu-latest), Test (macos-latest), Test (windows-latest), Build (linux/amd64), CodeQL.
- [ ] **Hash stability across platforms**: `TestManifestHashStability` GREEN on all three OSes (Linux/macOS/Windows). Tests against the normalized content invariant in `catalog_doc.md`. Pre-emptively handles CRLF/LF differences (M2-T2.4).
- [ ] No backward incompatibility: `internal/template/deployer_test.go` existing test cases all GREEN (M4-T4.5). The current `moai init` and `moai update` behavior is unchanged.

## Review Process Checklist (D10 권장 수정 반영 — non-runtime, PR review only)

The following items are not runtime audit checks; they are independent process invariants surfaced to PR reviewers:

- Frontmatter compliance for SPEC documents (spec.md, plan.md, acceptance.md): 9 required fields in spec.md, no legacy aliases (`created`, `updated`, `spec_id`). Verified during plan-auditor independent review.
- Tier change requires SPEC ID amendment in PR description (REQ-CATALOG-001-013). Verified during plan-auditor independent review at PR review time.
- Sentinel `CATALOG_DUPLICATE_ENTRY` (REQ-CATALOG-001-027) emits during audit but its root-cause documentation is reviewed in plan-auditor verdict.

## Performance Criteria

- **catalog.yaml load time**: < 50ms cold-load (measured via `LoadCatalog()` benchmark in M4 — `BenchmarkLoadCatalog`). At ~80KB YAML size, `yaml.v3` parsing should complete in 5-10ms.
- **Audit test suite execution time**: < 5 seconds for the combined `go test -run TestCatalog ./internal/template/` (10 parallel sub-tests T3.1–T3.10, ~65 entries each).
- **catalog.yaml file size**: 80-120KB at v1.0.0 (consistent with spec.md §"Files to Modify" 800-1200 YAML lines anticipated). 65 entries × ~6 fields × YAML indentation + 9 packs + comments. Hard cap 150KB enforced by audit test `TestCatalogFileSize` (optional, not required by EARS but recommended in M3).
- **Binary size impact**: < 100KB increase to the final `moai` binary (catalog.yaml is the only new embedded asset). Negligible compared to current ~30MB binary.
- **Build time impact**: < 100ms increase to `make build` (yaml embed + audit test compilation). Measured against the baseline pre-SPEC build.
- **Memory footprint at runtime**: `LoadCatalog()` allocates < 200KB heap (65 entries × 3KB struct each). Released after Deploy() phase.
