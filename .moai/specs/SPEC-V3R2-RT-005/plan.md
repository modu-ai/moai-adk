# SPEC-V3R2-RT-005 Implementation Plan

> Implementation plan for **Multi-Layer Settings Resolution with Provenance Tags**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored against branch `plan/SPEC-V3R2-RT-005` at repo root `/Users/goos/MoAI/moai-adk-go` (the branch is checked out here per `git worktree list`). Base: `origin/main` HEAD `496595c3f`.

## HISTORY

| Version | Date       | Author                         | Description                                                              |
|---------|------------|--------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)   | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B. Scope: harden existing 14-file `internal/config/` skeleton to fully satisfy 27 EARS REQs and 15 ACs declared in `spec.md`. |

---

## 1. Plan Overview

### 1.1 Goal restatement

spec.md §1을 구체적 milestone으로 변환: 기존 `internal/config/` 14개 파일이 이미 65% 구현한 8-tier resolver/merger를 완전히 hardening 하여 27 EARS REQs 와 15 ACs 를 모두 충족한다.

핵심 deltas (research.md §2.1 detail):

- **`audit_test.go` real implementation**: 현재 `t.Skip("placeholder")` 상태의 TestAuditParity를 yaml↔Go struct registry 기반 실제 검증으로 전환 → REQ-V3R2-RT-005-008, -043 충족 + 5개 누락 yaml (constitution/context/interview/design/harness) 을 `YAMLAuditExceptions` map에 등록하여 MIG-003 분리 작업 가능화.
- **`merge.go:240-275` `dumpJSON` 실제 encoding/json 교체**: 현재 `formatMapAsJSON`이 `fmt.Sprintf("%+v", m)` placeholder 상태 → REQ-006, -030 byte-stable 출력 보장.
- **`resolver.go:252-255` `loadYAMLFile` 실제 파싱**: 현재 빈 map 반환 placeholder → `gopkg.in/yaml.v3` 사용 실제 yaml 파싱; 8-tier 머지 의미적 동작 확립.
- **diff-aware reload (REQ-011)**: 새 메서드 `(*resolver).Reload(path string) error` 추가; 변경된 file path가 속한 tier만 re-load 후 merge delta 적용.
- **type-mismatch detection (REQ-013)**: `loader.go:loadYAMLFile`에 yaml.v3 strict mode + type assertion → `ConfigTypeError` raise.
- **policy strict_mode enforcement (REQ-022)**: merge 단계에서 lower tier가 policy-designated key를 override 시 `PolicyOverrideRejected` raise.
- **yaml↔yml ambiguity detection (REQ-041)**: `loadYAMLSections` 에서 같은 basename을 가진 sibling 검출 시 `ConfigAmbiguous` raise.
- **schema_version 전파 (REQ-033)**: `loadYAMLFile`에서 최상위 `schema_version: N` 키 검출 → `Provenance.SchemaVersion` 채움.
- **Skill frontmatter loading (REQ-015 partial)**: `resolver.go:218` placeholder를 `.claude/skills/**/SKILL.md` frontmatter `config:` block 파싱으로 대체.
- **`.moai/logs/config.log` 전용 로그**: tier read failure 시 `slog.Warn` 외 전용 로그 파일에 기록 (REQ-040).
- **resolver 동시성 안전 (research.md §10.6)**: `sync.RWMutex` 추가로 Reload/Load와 Key/Dump 동시 호출 안전 보장.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED**: 기존 8개 test 파일 (`source_test.go`, `provenance_test.go`, `merge_test.go`, `resolver_test.go`, `loader_test.go`, `manager_test.go`, `validation_test.go`, `types_test.go`) + `audit_test.go`에 새 케이스 추가 — 15 ACs를 통과하지 못하는 부분에 대한 failing test. 새 파일: `internal/config/audit_registry.go` (registry data) + `internal/config/audit_registry_test.go` (registry tests) + `internal/config/reload_test.go` (diff-aware reload tests).
- **GREEN**: §1.1의 9개 delta 구현. 기존 placeholder 함수 본체를 실제 로직으로 교체; 새 메서드 (`Reload`) 추가.
- **REFACTOR**: yaml 파싱 헬퍼를 `loader.go::loadYAMLFile` (기존) 와 `resolver.go::loadYAMLFile` (placeholder) 사이에 통합 (DRY); `formatMapAsJSON` 제거 후 `encoding/json.MarshalIndent` 사용; resolver mutex 패턴을 Loader 의 sync.RWMutex 와 동일하게 표준화.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Real TestAuditParity with registry-driven yaml↔struct check | `internal/config/audit_test.go` (replace placeholder, ~80 LOC), `internal/config/audit_registry.go` (new, ~60 LOC) | REQ-008, REQ-043 |
| Yaml↔struct registry data | `internal/config/audit_registry.go` (new) | REQ-021, REQ-008 |
| Real JSON dumper using encoding/json | `internal/config/merge.go` (extend `dumpJSON`, ~30 LOC) | REQ-006, REQ-030 (byte-stable per cache-prefix discipline) |
| Real yaml parsing in resolver | `internal/config/resolver.go` (replace `loadYAMLFile` placeholder, ~25 LOC; add `gopkg.in/yaml.v3` import) | REQ-010, REQ-013, REQ-033 |
| Diff-aware reload API | `internal/config/resolver.go` (new method `Reload(path string) error`, ~70 LOC) | REQ-011 |
| Type-mismatch detection raising ConfigTypeError | `internal/config/resolver.go` (yaml strict mode + type-check helper, ~40 LOC) | REQ-013 |
| Policy strict_mode enforcement raising PolicyOverrideRejected | `internal/config/merge.go` (extend MergeAll, ~30 LOC) | REQ-022 |
| Yaml/yml sibling ambiguity detection | `internal/config/resolver.go` (extend `loadYAMLSections`, ~25 LOC) | REQ-041 |
| schema_version 추적 in Provenance | `internal/config/resolver.go::loadYAMLFile` (extract top-level key, ~10 LOC) | REQ-033 |
| Skill frontmatter loading | `internal/config/resolver.go` (replace `loadSkillTier` placeholder, ~80 LOC; uses `gopkg.in/yaml.v3` for fenced YAML extraction) | REQ-015 |
| Dedicated config.log for tier read failures | `internal/config/resolver.go` (warn helper writing to `.moai/logs/config.log`, ~30 LOC) | REQ-040 |
| Resolver mutex (sync.RWMutex) | `internal/config/resolver.go` (extend `resolver` struct, ~10 LOC) | concurrency safety |
| validator/v10 integration | `internal/config/types.go` (add validate tags, ~50 LOC), `internal/config/validation.go` (extend Validate w/ validator.New, ~30 LOC) | REQ-013 hardening |
| YAMLAuditExceptions map for 5 pending yaml | `internal/config/audit_registry.go` (~15 LOC) | unblock MIG-003 |
| Sentinel test cases for 15 ACs | extend 8 existing `*_test.go` files + 2 new test files | REQ→AC traceability |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 6 files (per §6 below) | mx_plan |

Embedded-template parity is **not applicable** — this SPEC modifies only `internal/` Go source. `make build` regeneration is required because `internal/template/embedded.go` may be checksum-affected by audit_test.go addition (verified at M5).

### 1.4 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task). Built **after** tasks.md was finalized; each row references actual T-RT005-NN IDs whose subjects in `tasks.md` match the REQ behavior.

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-005-001 | Ubiquitous (Source enum 8 values) | (baseline; verified by existing `source_test.go::TestAllSources`) | T-RT005-01 |
| REQ-V3R2-RT-005-002 | Ubiquitous (Provenance fields populated) | AC-01, AC-02 | T-RT005-02 |
| REQ-V3R2-RT-005-003 | Ubiquitous (Value[T] + Unwrap/Origin methods) | AC-02 | T-RT005-02 |
| REQ-V3R2-RT-005-004 | Ubiquitous (SettingsResolver interface) | AC-02, AC-10 | T-RT005-03 |
| REQ-V3R2-RT-005-005 | Ubiquitous (deterministic merge) | AC-01 | T-RT005-04, T-RT005-13 |
| REQ-V3R2-RT-005-006 | Ubiquitous (`config dump` JSON) | AC-02 | T-RT005-05, T-RT005-14 |
| REQ-V3R2-RT-005-007 | Ubiquitous (`config diff`) | AC-03 | T-RT005-06, T-RT005-15 |
| REQ-V3R2-RT-005-008 | Ubiquitous (audit_test fail on orphan yaml) | AC-08 | T-RT005-07, T-RT005-16 |
| REQ-V3R2-RT-005-010 | Event-Driven (Load reads 8 tiers) | AC-02, AC-06 | T-RT005-04, T-RT005-13 |
| REQ-V3R2-RT-005-011 | Event-Driven (ConfigChange diff-aware reload) | AC-04 | T-RT005-08, T-RT005-17 |
| REQ-V3R2-RT-005-012 | Event-Driven (OverriddenBy population) | AC-01 | T-RT005-04, T-RT005-13 |
| REQ-V3R2-RT-005-013 | Event-Driven (ConfigTypeError) | AC-05 | T-RT005-09, T-RT005-18 |
| REQ-V3R2-RT-005-014 | Event-Driven (policy file absent → empty tier) | AC-06 | T-RT005-13 |
| REQ-V3R2-RT-005-015 | Event-Driven (plugin tag SrcPlugin) | (slot only, no contributor in v3.0; verified by negative test) | T-RT005-19 |
| REQ-V3R2-RT-005-020 | State-Driven (`"default"` flag for SrcBuiltin) | AC-14 | T-RT005-14 |
| REQ-V3R2-RT-005-021 | State-Driven (audit_test fail on yaml/struct count delta) | AC-08 | T-RT005-07, T-RT005-16 |
| REQ-V3R2-RT-005-022 | State-Driven (PolicyOverrideRejected) | AC-07 | T-RT005-10, T-RT005-20 |
| REQ-V3R2-RT-005-030 | Optional (`--format yaml` w/ comments) | AC-09 | T-RT005-14 |
| REQ-V3R2-RT-005-031 | Optional (filepath.Abs normalization) | (cosmetic; verified by Origin assertion) | T-RT005-21 |
| REQ-V3R2-RT-005-032 | Optional (`--key <name>`) | AC-10 | T-RT005-15 |
| REQ-V3R2-RT-005-033 | Optional (Provenance.SchemaVersion) | AC-11 | T-RT005-22 |
| REQ-V3R2-RT-005-040 | Unwanted (skip tier on read failure + log) | (negative test) | T-RT005-23 |
| REQ-V3R2-RT-005-041 | Unwanted (ConfigAmbiguous on yaml/yml siblings) | AC-12 | T-RT005-11, T-RT005-24 |
| REQ-V3R2-RT-005-042 | Unwanted (ConfigSchemaMismatch) | AC-15 | T-RT005-25 |
| REQ-V3R2-RT-005-043 | Unwanted (audit_test fail on orphan yaml) | AC-08 | T-RT005-07, T-RT005-16 |
| REQ-V3R2-RT-005-050 | Complex (SrcSession session-scoped reset) | AC-13 | T-RT005-12 |
| REQ-V3R2-RT-005-051 | Complex (`config diff` merged-view delta) | AC-03 | T-RT005-15 |

Coverage: **27 REQs mapped to 15 ACs and 40 tasks (T-RT005-01..40)** (count measured 2026-05-10 via `grep -c "^| T-RT005-" tasks.md` = 40; some ACs and tasks map to multiple REQs). The matrix above cites representative task IDs at the abstract milestone level; actual `tasks.md` provides the granular 40-task breakdown across M1 (7 tasks T-01..07 RED) → M2 (6 tasks T-08..13) → M3 (6 tasks T-14..19) → M4 (8 tasks T-20..27) → M5 (13 tasks T-28..40).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

Reference existing tests: `internal/config/{source,provenance,merge,resolver,loader,manager,validation,types}_test.go`.

Owner role: `expert-backend` (Go test) or direct `manager-cycle` execution.

Scope:

1. Add new test cases to `internal/config/audit_test.go` exercising AC-08:
   - `TestAuditParity_OrphanYAMLFails` — adds `.moai/config/sections/foo.yaml` to a `t.TempDir()`, asserts `TestAuditParity` fails naming `foo.yaml`.
   - `TestAuditParity_AllRegisteredYAMLPass` — happy-path with all 16 registered yaml.
   - `TestAuditParity_ExceptionsRespected` — yaml in `YAMLAuditExceptions` does NOT cause failure.
2. Add new test cases to `internal/config/merge_test.go`:
   - `TestMergeAll_OverriddenByPopulated` — 3 tiers (policy=true, user=false-zero, project=true), winner=policy, OverriddenBy=[project_origin].
   - `TestMergeAll_ByteStableJSON` — same input → same JSON dump byte sequence (cache-prefix discipline).
   - `TestMergeAll_PolicyOverrideRejected` — policy.strict_mode=true, project tries override → `PolicyOverrideRejected`.
3. Add new test cases to `internal/config/resolver_test.go`:
   - `TestResolver_LoadAll8Tiers` — populated tiers in `t.TempDir()`, asserts merged map keys.
   - `TestResolver_PolicyAbsentNoError` — REQ-014.
   - `TestResolver_ConfigTypeError` — yaml `coverage_threshold: "high"` (string where int expected) → `ConfigTypeError`.
   - `TestResolver_ConfigAmbiguous` — sibling `quality.yaml` + `quality.yml` with conflicting key → `ConfigAmbiguous`.
   - `TestResolver_SchemaVersionPropagation` — yaml `schema_version: 3` → `Provenance.SchemaVersion == 3`.
4. Create `internal/config/reload_test.go` (new file):
   - `TestResolver_Reload_TierIsolation` — modify single yaml → only that tier re-parses; other keys retain old `Loaded` timestamp.
   - `TestResolver_Reload_DeltaApplied` — modified yaml value reflects in next `Key()` call.
5. Create `internal/config/audit_registry_test.go` (new file):
   - `TestRegistry_AllRegisteredStructsExist` — reflection check `reflect.TypeOf(Config{}).FieldByName(structName)` for each registry entry.
   - `TestRegistry_NoUnexpectedYAMLOrphans` — current 23 yaml files all either in registry or in YAMLAuditExceptions.
6. Add CLI tests in `internal/cli/doctor_config_test.go` (extend existing or create new):
   - `TestDoctorConfigDump_HappyPath` — populated resolver → JSON output contains key/V/P structure.
   - `TestDoctorConfigDump_FormatYAML` — `--format yaml` → `# source: <tier>` comments per key.
   - `TestDoctorConfigDump_SingleKey` — `--key permission.strict_mode` → only that key.
   - `TestDoctorConfigDiff_TierComparison` — `user` vs `project` → diff output.
7. Run `go test ./internal/config/ ./internal/cli/` from repo root — confirm RED for all new tests.

Verification gate before advancing to M2: at least 14 of the new tests fail with documented sentinel messages. Existing tests continue to pass (regression baseline).

[HARD] No implementation code in M1 outside of test files (per spec-workflow.md TDD discipline).

### M2: Audit registry + ConfigTypeError + ConfigAmbiguous (GREEN, part 1) — Priority P0

Owner role: `expert-backend`.

Scope:

1. Create `internal/config/audit_registry.go` with:
   - `YAMLToStructRegistry map[string]string` — 16 yaml basename → struct name mappings (verified against `types.go:312-318` `sectionNames`).
   - `YAMLAuditExceptions map[string]string` — `mx.yaml`, `security.yaml`, `runtime.yaml` plus the 5 "MIG-003 pending" yaml files (constitution/context/interview/design/harness) with rationale strings.
   - Helper `IsRegisteredOrException(filename string) bool`.
2. Replace `internal/config/audit_test.go::TestAuditParity` body with real implementation (remove `t.Skip`):
   - Scan `.moai/config/sections/*.yaml`.
   - For each file: check registry → check exceptions → otherwise FAIL `"orphan yaml file (no Go struct mapping): %s"`.
   - For each registry entry: reflection check `structExists(name)` via `reflect.TypeOf(Config{}).FieldByName` recursive walk; fail if not found.
3. Extend `internal/config/resolver.go::loadYAMLFile`:
   - Replace placeholder with `gopkg.in/yaml.v3` real parsing.
   - Use `yaml.Decoder` with `KnownFields(true)` for strict mode (rejects unknown keys → optional toggle via env var for compat).
   - Detect type mismatch via post-unmarshal type assertion; raise `ConfigTypeError` per REQ-013.
4. Extend `internal/config/resolver.go::loadYAMLSections`:
   - Sibling detection: walk directory entries; group by basename (without ext); if same basename has both `.yaml` and `.yml` AND values for same key differ → raise `ConfigAmbiguous`.
5. Extend `internal/config/resolver.go::loadYAMLFile` to extract top-level `schema_version: N` key and propagate to per-key Provenance (REQ-033).

Verification: `TestAuditParity_OrphanYAMLFails`, `TestAuditParity_AllRegisteredYAMLPass`, `TestResolver_ConfigTypeError`, `TestResolver_ConfigAmbiguous`, `TestResolver_SchemaVersionPropagation` turn GREEN. All existing tests still pass.

### M3: Real JSON dump + OverriddenBy + PolicyOverrideRejected (GREEN, part 2) — Priority P0

Owner role: `expert-backend`.

Scope:

1. Replace `internal/config/merge.go::formatMapAsJSON` (line 271-275 placeholder) with `encoding/json.MarshalIndent`:
   - Sort keys alphabetically before serialization (cache-prefix byte stability).
   - Marshal `Provenance` fully (Source.String(), Origin, Loaded RFC3339, SchemaVersion, OverriddenBy).
2. Extend `internal/config/merge.go::MergeAll` (lines 63-134):
   - Verify OverriddenBy correctly accumulates lower-tier non-zero values per spec REQ-012.
   - Add `policy.strict_mode` check: if SrcPolicy tier sets `policy.strict_mode: true` AND any lower tier provides non-zero value for a policy-designated key, raise `PolicyOverrideRejected` per REQ-022.
   - "policy-designated key" definition: key where SrcPolicy provided a non-zero value (any key from policy tier is policy-designated).
3. Add log entry for rejected override: `slog.Warn("policy override rejected", "key", k, "from", source.String(), "policy_origin", policyOrigin)`.
4. Extend `internal/config/merge.go::dumpYAML` (lines 259-267):
   - Add `# source: <tier>` comment per key.
   - Sort keys alphabetically (cache-prefix discipline).
   - REQ-V3R2-RT-005-030 satisfied with assertable substring.
5. Extend `internal/config/provenance.go`:
   - `IsDefault()` method already exists; add `MarshalJSON()` for stable JSON serialization with Source as String().

Verification: `TestMergeAll_OverriddenByPopulated`, `TestMergeAll_ByteStableJSON`, `TestMergeAll_PolicyOverrideRejected` turn GREEN. AC-01, AC-07 satisfied.

### M4: Diff-aware reload + Skill frontmatter + dedicated log + concurrency (GREEN, part 3) — Priority P0

Owner role: `expert-backend`.

Scope:

1. Extend `internal/config/resolver.go::resolver` struct:
   - Add `mu sync.RWMutex`.
   - Add `tierData map[Source]map[string]any` — cached per-tier data for diff-aware reload.
   - Add `tierOrigins map[Source]string` — cached per-tier origin paths.
2. Add new method `(*resolver).Reload(path string) error`:
   - Determine which tier `path` belongs to (by prefix match: `/etc/moai/`→Policy, `~/.moai/`→User, `.moai/`→Project, `.claude/`→Local).
   - Re-parse only that tier's yaml/json files.
   - Re-merge using cached data from other tiers.
   - Update `r.merged` atomically under `mu.Lock()`.
   - Update `r.loadedAt = time.Now()` for affected keys only (other keys retain prior timestamp).
3. Replace `internal/config/resolver.go::loadSkillTier` placeholder (line 218):
   - Walk `.claude/skills/**/SKILL.md`.
   - For each: extract YAML frontmatter (between `---` markers).
   - Parse `config:` block if present.
   - Aggregate per-skill config under `skill.<name>` keyspace.
   - Origin: skill SKILL.md absolute path.
4. Add `.moai/logs/config.log` writer:
   - Helper `logTierReadFailure(source Source, path string, err error)` writes append-mode line to log file (best-effort; if log file inaccessible, fall back to stderr).
   - Called from each `loadXxxTier` on read failure (REQ-V3R2-RT-005-040).
5. Wire mutex into `Load()`, `Reload()`, `Key()`, `Dump()`, `Diff()`:
   - Load/Reload use `mu.Lock()`.
   - Key/Dump/Diff use `mu.RLock()`.

Verification: `TestResolver_Reload_TierIsolation`, `TestResolver_Reload_DeltaApplied`, `TestResolver_LoadAll8Tiers` (with skill tier populated) turn GREEN. AC-04 satisfied. Concurrency safe (no race detected with `go test -race`).

### M5: Doctor CLI hardening + validator/v10 + CHANGELOG + MX tags (GREEN, part 4 + Trackable) — Priority P1

Owner role: `expert-backend` (CLI + validator) + `manager-docs` (CHANGELOG + MX tags).

Scope:

#### M5a: doctor config dump JSON byte stability (REQ-006)

1. Verify `internal/cli/doctor_config.go::runConfigDump` calls `resolver.Load() → merged.Dump("json")` produces byte-stable output across multiple calls.
2. Add CLI test asserting two consecutive `moai doctor config dump` invocations on same state produce identical bytes.

#### M5b: doctor config dump --key path (REQ-V3R2-RT-005-032)

1. Verify `runConfigDump` `--key permission.strict_mode` behavior matches AC-10 (single-key Value + Provenance output).
2. Existing implementation in `doctor_config.go:55-72` already covers; M5b is regression test only.

#### M5c: doctor config dump --format yaml comments (REQ-030)

1. Verify `# source: <tier>` comment present per key in YAML output.
2. Already implemented in `merge.go:262` `fmt.Fprintf(&sb, "%s: %v # source: %s\n", ...)` ✅; M5c is assertion test only.

#### M5d: validator/v10 integration on Config struct (REQ-013 hardening)

1. Add `validate:"..."` tags on critical fields in `types.go`:
   - `User.Name` → `validate:"required"` (already inferred via custom check).
   - `Quality.DevelopmentMode` → `validate:"omitempty,oneof=ddd tdd"`.
   - `LLM.PerformanceTier` → `validate:"omitempty,oneof=high medium low"`.
   - `GitConvention.Convention` → `validate:"omitempty,oneof=auto conventional-commits angular karma custom"`.
2. Replace custom validators in `validation.go::validateDevelopmentMode` etc. with single `validator.New().Struct(cfg)` call where possible.
3. Maintain backward compat: existing custom checks remain as fallback for fields not yet covered by struct tags.

#### M5e: Filename normalization for portable diagnostics (REQ-031)

1. Extend each `loadXxxTier` to call `filepath.Abs()` on origin path before populating `Provenance.Origin`.
2. Already partial in `resolver.go:115` (policyPath uses absolute by definition); apply same pattern to user/project/local tiers.

#### M5f: CHANGELOG + MX tags + final verification

1. Add CHANGELOG entry under `## [Unreleased]`:
   ```
   ### Added
   - SPEC-V3R2-RT-005: Multi-Layer Settings Resolution with Provenance Tags. Hardens existing `internal/config/` 8-tier resolver to fully satisfy 27 EARS REQs and 15 ACs. New: real `TestAuditParity` enforcing yaml↔Go struct parity (with YAMLAuditExceptions for 5 MIG-003-pending sections); diff-aware `Reload(path)` API; ConfigTypeError/ConfigAmbiguous/PolicyOverrideRejected raises in 8-tier merge; byte-stable JSON/YAML `dump`; resolver mutex for concurrent Read/Reload safety; Skill frontmatter `config:` block loading; dedicated `.moai/logs/config.log` for tier read failures; validator/v10 schema tags on Config struct.
   ```
2. Insert MX tags per §6 below.
3. Run full `go test ./...` from repo root. Verify ALL tests pass + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule).
4. Run `go test -race ./internal/config/` to verify mutex correctness under concurrent reload + read.
5. Run `make build` to regenerate `internal/template/embedded.go` (ensure no template content drift).
6. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete` after M1-M5f land.

[HARD] No new documents are created in `.moai/specs/` or `.moai/reports/` during M5 — this is a SPEC implementation phase.

---

## 3. File:line Anchors (concrete edit targets)

(All paths relative to repo root `/Users/goos/MoAI/moai-adk-go`. All line numbers verified via `Read` 2026-05-10.)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `internal/config/audit_test.go:9-22` | `TestAuditParity` body | Replace `t.Skip` placeholder with real registry-driven scan | M2 / REQ-008, REQ-043 |
| `internal/config/resolver.go:252-255` | `(*resolver).loadYAMLFile` placeholder | Replace empty-map return with `gopkg.in/yaml.v3` real parsing + strict mode + schema_version extraction | M2 / REQ-010, REQ-013, REQ-033 |
| `internal/config/resolver.go:258-295` | `(*resolver).loadYAMLSections` | Add yaml/yml sibling detection raising `ConfigAmbiguous` | M2 / REQ-041 |
| `internal/config/resolver.go:218-220` | `(*resolver).loadSkillTier` placeholder | Replace with `.claude/skills/**/SKILL.md` frontmatter `config:` block parser | M4 / REQ-015 |
| `internal/config/resolver.go:32-35` | `resolver` struct | Add `mu sync.RWMutex`, `tierData map[Source]map[string]any`, `tierOrigins map[Source]string` | M4 / concurrency, REQ-011 |
| `internal/config/resolver.go:46-74` | `(*resolver).Load()` | Wrap with `mu.Lock()`; populate `tierData`/`tierOrigins` for later Reload | M4 / REQ-010, REQ-011 |
| `internal/config/resolver.go:298-305` | `(*resolver).Key` | Wrap with `mu.RLock()` | M4 / concurrency |
| `internal/config/resolver.go:308-320` | `(*resolver).Dump` | Wrap with `mu.RLock()` | M4 / concurrency |
| `internal/config/resolver.go:323-355` | `(*resolver).Diff` | Wrap with `mu.RLock()` and align with spec REQ-051 "merged-view delta" semantics | M4 / REQ-007, REQ-051 |
| `internal/config/resolver.go:end` | new method `Reload(path string) error` | Add diff-aware reload | M4 / REQ-011 |
| `internal/config/resolver.go:end` | new helper `logTierReadFailure(source, path, err)` | Append to `.moai/logs/config.log` | M4 / REQ-040 |
| `internal/config/merge.go:63-134` | `MergeAll` | Add `policy.strict_mode` enforcement raising `PolicyOverrideRejected`; verify OverriddenBy population | M3 / REQ-012, REQ-022 |
| `internal/config/merge.go:240-258` | `(*MergedSettings).dumpJSON` | Replace `formatMapAsJSON` placeholder with real `encoding/json.MarshalIndent` + sort keys | M3 / REQ-006 |
| `internal/config/merge.go:259-267` | `(*MergedSettings).dumpYAML` | Sort keys alphabetically (cache-prefix discipline) | M3 / REQ-030 |
| `internal/config/merge.go:271-275` | `formatMapAsJSON` | Remove (consumed by replaced dumpJSON) | M3 / REFACTOR |
| `internal/config/types.go:13-32` | `Config` struct | Add `validate:"..."` tags on critical fields | M5d / REQ-013 hardening |
| `internal/config/types.go:312-318` | `sectionNames` slice | Synchronize with `audit_registry.go::YAMLToStructRegistry` keys | M2 / REQ-008 |
| `internal/config/validation.go:25-47` | `Validate(cfg, loadedSections)` | Add `validator.New().Struct(cfg)` call as first step; preserve custom checks as fallback | M5d / REQ-013 |
| `internal/config/provenance.go:end` | new method `(Provenance) MarshalJSON()` | Stable JSON serialization with Source.String() | M3 / REQ-006 |
| `internal/cli/doctor_config.go:45-82` | `runConfigDump` | (no edit; verify existing behavior matches AC-10 byte-stability assertion) | M5a |
| `CHANGELOG.md` | `## [Unreleased]` section | Add Added entry per §M5f | M5f / Trackable |

### 3.2 To-be-created (new files)

| File | Reason | LOC estimate |
|------|--------|--------------|
| `internal/config/audit_registry.go` | yaml↔struct registry data + exceptions | ~80 |
| `internal/config/audit_registry_test.go` | registry consistency tests | ~60 |
| `internal/config/reload_test.go` | diff-aware reload test scenarios | ~120 |
| `internal/cli/doctor_config_test.go` | CLI tests if not exists (extend if exists) | ~150 |

Total new: ~410 LOC. Total modified: ~350 LOC. Net additions: ~760 LOC across 4 new files + 7 modified files.

### 3.3 NOT to be touched (preserved by reference)

- `internal/config/manager.go` — high-level facade, RT-005 scope is resolver-layer only.
- `internal/config/required_checks.go` — auxiliary validation, no RT-005 change.
- `internal/config/envkeys.go` — env var name constants, no RT-005 change.
- `internal/config/defaults.go` — builtin defaults are SrcBuiltin tier source; RT-005 reads them as-is.
- `internal/config/loader.go` — project-tier yaml loader, used by Manager facade. RT-005's `resolver.go` is parallel implementation reading from different tier sources; Loader stays as-is.
- 5 yaml files (`constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml`, `harness.yaml`) — added to YAMLAuditExceptions but loader bodies belong to MIG-003.
- `pkg/models/*` — shared types, RT-005 scope is internal/config only.

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability and research.md §13, the following anchors are load-bearing and cited verbatim throughout this plan (each verified via Read 2026-05-10):

1. `.moai/specs/SPEC-V3R2-RT-005/spec.md:42-71` — In-scope items.
2. `.moai/specs/SPEC-V3R2-RT-005/spec.md:120-164` — 27 EARS REQs.
3. `.moai/specs/SPEC-V3R2-RT-005/spec.md:166-181` — 15 ACs.
4. `.moai/specs/SPEC-V3R2-RT-005/spec.md:184-191` — Constraints.
5. `internal/config/source.go:1-113` — Source enum baseline (REQ-001 ✅).
6. `internal/config/provenance.go:1-71` — Provenance + Value[T] baseline (REQ-002, REQ-003 ✅).
7. `internal/config/types.go:13-32` — Config root + 16 sections.
8. `internal/config/types.go:312-318` — `sectionNames` slice (16 sections registered).
9. `internal/config/loader.go:31-70` — `Loader.Load()` (8 sections wired).
10. `internal/config/merge.go:63-134` — `MergeAll` deterministic priority walk.
11. `internal/config/merge.go:138-162` — `isZero` reflection-based zero check.
12. `internal/config/merge.go:168-205` — `Diff(a, b *MergedSettings)`.
13. `internal/config/merge.go:240-275` — `Dump(json/yaml)` with placeholder JSON formatter.
14. `internal/config/resolver.go:17-29` — `SettingsResolver` interface.
15. `internal/config/resolver.go:46-74` — `(*resolver).Load()` 8-tier loop.
16. `internal/config/resolver.go:103-125` — `loadPolicyTier()` platform-specific.
17. `internal/config/resolver.go:218-235` — `loadSkillTier`/`loadSessionTier`/`loadBuiltinTier` placeholders.
18. `internal/config/resolver.go:252-255` — `loadYAMLFile` placeholder (CRITICAL: must be replaced).
19. `internal/config/resolver.go:258-295` — `loadYAMLSections` (needs sibling detection).
20. `internal/config/resolver.go:298-355` — `Key`, `Dump`, `Diff(Source, Source)`.
21. `internal/config/resolver_errors.go:9-79` — 5 error types (defined; raises pending).
22. `internal/config/audit_test.go:9-22` — placeholder TestAuditParity.
23. `internal/config/validation.go:25-47` — `Validate(cfg, loadedSections)` (existing custom validator).
24. `internal/cli/doctor_config.go:13-167` — `doctor config dump`/`diff` Cobra commands.
25. `.moai/config/sections/` — 23 yaml files.
26. `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — subagent prohibition.
27. `.claude/rules/moai/workflow/spec-workflow.md:172-204` — Plan Audit Gate.
28. `CLAUDE.local.md:§6` — test isolation.
29. `CLAUDE.local.md:§14` — no hardcoded paths in `internal/`.

Total: **29 distinct file:line anchors** (exceeds plan-auditor minimum of 10 by 19).

---

## 4. Technology Stack Constraints

Per `spec.md` §7 Constraints, **minimal new technology** is introduced:

- Go 1.22+ — already required by `go.mod` (generic `Value[T any]` already in use per provenance.go).
- `github.com/go-playground/validator/v10` — added by SPEC-V3R2-SCH-001 (blocker dependency). Imported as new dependency in `go.mod` at M5d if not already present.
- `gopkg.in/yaml.v3` — already in `go.mod` (used in `internal/config/loader.go:11`).
- `encoding/json` — stdlib, already used in `internal/config/resolver.go:243-249`.
- `reflect` — stdlib, already used in `internal/config/merge.go:138-162`.

The only additive surfaces are:

- 4 new Go files under `internal/config/` and `internal/cli/` (per §3.2).
- ~7 modified Go files (per §3.1).
- One CHANGELOG entry.
- 7 MX tags across 6 files (per §6).

**No new directory structures** — `internal/config/` already exists with 14 source files.

**No new YAML schema files** — RT-005 reads existing 23 yaml sections; the 5 unloaded ones are documented in YAMLAuditExceptions but their bodies are owned by MIG-003.

---

## 5. Risk Analysis & Mitigations

Extends `spec.md` §8 risks with concrete file-path mitigations.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| validator/v10 from SPEC-V3R2-SCH-001 not yet merged | M | M | Check `go.mod` at M5d start; if absent, add directly via `go get github.com/go-playground/validator/v10`. SCH-001 status verified at plan-audit gate. |
| audit_test.go strict mode causes build fail when 5 yaml are orphan | H | M | YAMLAuditExceptions registers all 5 (constitution/context/interview/design/harness) at M2 introduction. MIG-003 removes entries one-by-one as it adds Go structs. |
| Go map iteration non-determinism breaks JSON byte stability | M | M | M3 dumpJSON sorts keys alphabetically before MarshalIndent. Test `TestMergeAll_ByteStableJSON` verifies. |
| Resolver mutex deadlock if Reload triggers callback that calls Key | L | H | Reload uses Lock(); callbacks from inside Reload (none expected) would deadlock. Mitigation: Reload does not invoke any external callback; document this in Reload godoc. |
| Diff-aware reload misses cross-tier dependency | M | M | Conservative behavior: when path's tier cannot be determined, fall back to full Load(). Logged warning. Cold-reload escape hatch via `moai doctor config reload-all` (out of scope; tracked as RT-006 follow-up). |
| Skill frontmatter parsing surface area large (`.claude/skills/**/SKILL.md`) | M | L | Glob walk capped at 200 SKILL.md files (current count ~80); each SKILL.md frontmatter is YAML so reuses yaml.v3 parser. Origin: SKILL.md absolute path. Failures logged but do not block resolver Load. |
| Concurrent Reload races corrupt resolver state | M | H | All mutations (Load, Reload) use mu.Lock(); all reads (Key, Dump, Diff) use mu.RLock(). `go test -race` gate at M5f. |
| ConfigAmbiguous false positive on intentional `.yml` shadow | L | L | Detection requires *conflicting* values (different `V` for same key); identical sibling values are silently accepted. Documented in resolver.go::loadYAMLSections godoc. |
| OverriddenBy bloat for keys with many tiers (8 tiers × ~100 keys) | L | L | OverriddenBy slice averages 1-2 entries (most keys set by ≤2 tiers); 8-element slice is upper bound = 8 × 100 bytes path = 800 bytes per key × 100 keys = 80 KiB total. Well under 2 MiB ceiling. |
| Plugin tier walk overhead in v3.0 (always empty) | L | L | resolver.go:88-89 short-circuits with `nil, "", nil` return. Negligible cost. |
| `.moai/logs/config.log` file lock contention if many processes write | L | L | Append-mode file open per write; OS-level append atomicity for short writes (<4KB). Not a high-frequency log path. |
| Schema_version extraction parses entire yaml twice | L | L | Use single Decoder pass: extract `schema_version` during the same `yaml.Decoder.Decode` call into a `map[string]any` view. No double parse. |
| `policy.strict_mode` enforcement breaks legitimate enterprise config patterns | M | M | enforcement is opt-in (only when `policy.strict_mode: true` set); error `PolicyOverrideRejected` is logged not panic — operators can disable strict mode if needed. |
| Spec REQ-051 "merged-view delta" semantics unclear vs simpler "raw delta" | L | M | M4 chooses: 8-tier full Load → for each key, compare winner.Source ∈ {a, b} → output keys where winner.Source equals one tier but not the other, OR same key has different winning Value across the two tiers. Test `TestResolver_Diff_MergedViewDelta` documents the convention. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

fan_in counts measured 2026-05-10 via `grep -rn "config\.<symbol>" --include="*.go" | wc -l` from repo root.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Measured fan_in | Rationale |
|------------------|-------------|-----------------|-----------|
| `internal/config/source.go:11-44` (`Source` enum constants) | `@MX:ANCHOR fan_in=49 — SPEC-V3R2-RT-005 REQ-001 8-tier priority enum; consumed by RT-002 (permission stack), RT-003 (sandbox), RT-006 (hook), MIG-003 (5 loaders), and 5 internal config files (merge.go, resolver.go, doctor_config.go, source_test.go, merge_test.go).` | 49 | Source enum is the constitutional backbone of the 8-tier system. ANY change ripples to 4+ downstream SPECs. |
| `internal/config/provenance.go:34-60` (`Value[T any]` generic + `Provenance`) | `@MX:ANCHOR fan_in=2 (current) → 8+ (post-RT-002/003/006) — SPEC-V3R2-RT-005 REQ-002, REQ-003 typed-state contract; every config field that needs to answer "where from?" wraps via Value[T].` | 2 (current; will grow as MIG-003 + RT-002 land) | Generic wrapper is the type-safety contract. Adding a new tier requires no Value[T] change but new consumers reach via `cfg.<field>.V`. |
| `internal/config/resolver.go:17-29` (`SettingsResolver` interface) | `@MX:ANCHOR fan_in=1 (current) → 4+ (post-RT-002/003/006) — SPEC-V3R2-RT-005 REQ-004 single resolver contract; methods Load/Key/Dump/Diff/Reload form the substrate. Adding methods is breaking; reordering signatures breaks RT-002 imports.` | 1 (current); future consumers RT-002, RT-003, RT-006, MIG-003 | Interface defines the SPEC-V3R2-RT-005 reader-layer API surface. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/config/merge.go:63-134` (`MergeAll`) | `@MX:NOTE — SPEC-V3R2-RT-005 deterministic 8-tier merge. Identical tier inputs MUST produce byte-identical merged output (cache-prefix discipline; problem-catalog P-C05). Map iteration order is non-deterministic; sort keys before serialization (see merge.go:dumpJSON).` | Documents the byte-stability invariant and the implementation detail that addresses it. |
| `internal/config/audit_test.go:TestAuditParity` (post-M2) | `@MX:NOTE — SPEC-V3R2-RT-005 REQ-008/043 yaml↔Go struct parity. New yaml under .moai/config/sections/ MUST register in audit_registry.go::YAMLToStructRegistry OR be added to YAMLAuditExceptions. MIG-003 will register the 5 currently-excepted files (constitution/context/interview/design/harness) progressively.` | Carries the workflow rule for future contributors adding yaml files. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/config/resolver.go::Reload(path string)` (M4 new method) | `@MX:WARN @MX:REASON — SPEC-V3R2-RT-005 REQ-011 diff-aware reload. Mutex held during entire reload. Callbacks invoked from within Reload (none expected) would deadlock. If a future feature needs post-reload notification, dispatch via channel from outside the locked region — see merge.go:MergeAll for the pattern.` | Reload is the most regression-prone method (lock + state mutation + cache invalidation). Inline callback addition is the most likely regression source. |
| `internal/config/merge.go::MergeAll` policy strict_mode block (M3 new) | `@MX:WARN @MX:REASON — SPEC-V3R2-RT-005 REQ-022 policy override enforcement. Disabling this check (e.g., commenting out the strict_mode enforcement) defeats the constitutional governance contract. Enterprise rollouts depend on PolicyOverrideRejected to enforce policy precedence.` | Easy to disable accidentally; consequence is silent acceptance of override that should be rejected. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

This SPEC produces a complete, audit-ready 8-tier resolver. No `@MX:TODO` markers are planned — all work converges to GREEN within M1-M5. The 5 yaml-without-loader sections are documented in YAMLAuditExceptions, not in TODO markers (per MIG-003 ownership boundary).

### 6.5 MX tag count summary

- @MX:ANCHOR: 3 targets (Source enum, Value[T]+Provenance, SettingsResolver interface)
- @MX:NOTE: 2 targets (MergeAll, TestAuditParity)
- @MX:WARN: 2 targets (Reload, policy strict_mode)
- @MX:TODO: 0 targets
- **Total**: 7 MX tag insertions planned across 6 distinct files (`source.go`, `provenance.go`, `resolver.go` [2 tags], `merge.go` [2 tags], `audit_test.go`)

---

## 7. Worktree Mode Discipline

[HARD] Per spec-workflow.md SPEC Phase Discipline (Step 1, this is plan phase): plan-phase work executes in main checkout (not worktree). The branch `plan/SPEC-V3R2-RT-005` is checked out at `/Users/goos/MoAI/moai-adk-go` per `git worktree list` output 2026-05-10.

Branch: `plan/SPEC-V3R2-RT-005` (already checked out per session context).

Run-phase agent will create a fresh worktree per spec-workflow.md SPEC Phase Discipline Step 2: `moai worktree new SPEC-V3R2-RT-005 --base origin/main`, then create `feat/SPEC-V3R2-RT-005-multi-layer-settings` per `CLAUDE.local.md` §18.2 branch naming.

[HARD] All Read/Write/Edit tool invocations use absolute paths under the repo root for plan phase artifacts (`.moai/specs/SPEC-V3R2-RT-005/*.md`).

[HARD] `make build` and `go test ./...` execute from the repo root: `cd /Users/goos/MoAI/moai-adk-go && make build && go test ./...` (during plan phase, build verification only).

> Note: Plan-phase agent operates from main checkout cwd; absolute paths shown for reference only.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:172-204`). Each item below cites **measured evidence** (grep -c, wc -l, ls counts) executed during plan authoring on 2026-05-10.

- [x] **C1: Frontmatter v0.1.0 schema** — `spec.md:1-22` frontmatter measured: 21 fields including `id`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `dependencies`, `bc_id`, `breaking`, `lifecycle`, `tags`. Verified by `Read spec.md` line 1-22.
- [x] **C2: HISTORY entry for v0.1.0** — `spec.md:30` HISTORY table has v0.1.0 row; measured by `grep -c "^| 0.1.0" spec.md` = 1.
- [x] **C3: 27 EARS REQs across 6 categories** — `grep -c "^- REQ-V3R2-RT-005-" spec.md` = **27**. Distribution verified: Ubiquitous 8 (REQ-001..008), Event-Driven 6 (REQ-010..015), State-Driven 3 (REQ-020..022), Optional 4 (REQ-030..033), Unwanted 4 (REQ-040..043), Complex 2 (REQ-050..051). Sum = 8+6+3+4+4+2 = 27 ✅.
- [x] **C4: 15 ACs all map to REQs (100% coverage)** — `grep -c "^- AC-V3R2-RT-005-" spec.md` = **15**. Each AC explicitly cites the REQ(s) it maps to (verified by inspection of spec.md:166-181). Plan §1.4 traceability matrix confirms all 27 REQs map to ≥1 AC and ≥1 task.
- [x] **C5: BC scope clarity** — `spec.md:19` `breaking: true` + `spec.md:14` `bc_id: [BC-V3R2-015]` + spec.md §1 documents AUTO migration: "reader layer only; settings.json files on disk remain unchanged; flat-merge consumers are internal and require no user API change."
- [x] **C6: File:line anchors ≥10** — research.md §13 cites **33 anchors**; this plan §3.4 cites **29 anchors**. Both exceed minimum.
- [x] **C7: Exclusions section present** — `spec.md:65-72` Out-of-scope (6 entries, each mapped to other SPECs: RT-002, MIG-003, RT-003, RT-006, MIG-003, v3.1+).
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml` `development_mode: tdd` confirmed via `grep development_mode .moai/config/sections/quality.yaml` (file present in repo).
- [x] **C9: mx_plan section** — this plan §6 defines 7 MX tag insertions across 4 categories (3 ANCHOR with measured fan_in, 2 NOTE, 2 WARN, 0 TODO). fan_in for Source enum measured = 49 via `grep -rn "config\.Source\|config\.SrcPolicy\|config\.SrcUser\|config\.SrcProject\|config\.SrcLocal\|config\.SrcSession" --include="*.go" | wc -l`.
- [x] **C10: Risk table with mitigations** — `spec.md:195-204` (7 risks) + this plan §5 (14 risks, all with file-anchored mitigations).
- [x] **C11: Plan phase checkout discipline** — this plan §7 (HARD rules per spec-workflow.md SPEC Phase Discipline Step 1: plan-phase in main checkout, branch `plan/SPEC-V3R2-RT-005` checked out at `/Users/goos/MoAI/moai-adk-go`).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, file paths, code-block templates, and pseudo-Go for interface declarations. No executable Go function bodies.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md sections AC-01 through AC-15 (each has Happy/Edge/Test mapping).
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md M1-M5 (manager-cycle / expert-backend / manager-docs / manager-git assignments). Total task count measured by `grep -c "^| T-RT005-" tasks.md` = **40** (the §1.4 traceability matrix cites representative task IDs at milestone abstract level; tasks.md provides the granular 40-task breakdown).
- [x] **C15: Cross-SPEC consistency** — blocked-by dependencies verified per spec.md §9: SPEC-V3R2-CON-001 (FROZEN zone — completed Wave 6), SPEC-V3R2-SCH-001 (validator/v10 — at-risk per §5 risk table; mitigation defined), SPEC-V3R2-RT-004 (SrcSession content — RT-004 plan complete, in-flight). RT-005 blocks RT-002, RT-003, RT-006, RT-007, MIG-003 per spec.md §9.2.

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: Add ~14 new test cases across `internal/config/{audit,merge,resolver,reload,audit_registry}_test.go` + `internal/cli/doctor_config_test.go`. Confirm RED for all new tests; existing tests still GREEN.
2. **M2 (P0)**: Audit registry (`audit_registry.go` new, `audit_test.go` rewrite) + ConfigTypeError (yaml.v3 strict mode in `loadYAMLFile`) + ConfigAmbiguous (`loadYAMLSections` sibling detection) + schema_version propagation. Confirm `TestAuditParity_*`, `TestResolver_ConfigTypeError`, `TestResolver_ConfigAmbiguous`, `TestResolver_SchemaVersionPropagation` GREEN.
3. **M3 (P0)**: Real JSON dumper (`encoding/json.MarshalIndent` replacing `formatMapAsJSON`) + OverriddenBy population verification + PolicyOverrideRejected enforcement + dumpYAML alphabetical sort. Confirm `TestMergeAll_OverriddenByPopulated`, `TestMergeAll_ByteStableJSON`, `TestMergeAll_PolicyOverrideRejected` GREEN. AC-01, AC-07 satisfied.
4. **M4 (P0)**: Diff-aware reload (`Reload(path string) error`) + Skill frontmatter loader + dedicated `.moai/logs/config.log` + resolver mutex (sync.RWMutex). Confirm `TestResolver_Reload_*`, `TestResolver_LoadAll8Tiers` (with skill tier) GREEN. AC-04 satisfied. `go test -race` clean.
5. **M5 (P1)**: validator/v10 integration on Config struct + filename normalization + CHANGELOG entry + 7 MX tag insertions per §6 + final `make build` + `go test ./...` + `go test -race ./internal/config/`. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`.

Total milestones: 5. Total file edits (existing): ~7. Total file creations (new): 4. Total CHANGELOG entries: 1. Total MX tag insertions: 7.

---

End of plan.md.
