# SPEC-V3R2-RT-005 Implementation Tasks

> Run-phase task list for **Multi-Layer Settings Resolution with Provenance Tags**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                       | Description                                                              |
|---------|------------|------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow) | Initial task decomposition — 28 tasks across M1-M5 milestones (TDD mode) |

---

## Methodology

`.moai/config/sections/quality.yaml` `development_mode: tdd` → RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

각 task 는:
- **ID**: T-RT005-NN (zero-padded)
- **Milestone**: M1 (RED) / M2-M4 (GREEN) / M5 (Trackable + final)
- **Owner**: `expert-backend` (Go) / `manager-cycle` (직접) / `manager-docs` (CHANGELOG/MX) / `manager-git` (PR)
- **Depends_on**: prerequisite task IDs
- **Files**: 변경/생성 file paths
- **REQ/AC**: 충족 EARS REQs + ACs
- **LOC**: estimate (test files separate from source)
- **Parallel**: 다른 task 와 동시 실행 가능 여부

---

## Milestone M1 — RED phase (Test scaffolding)

[HARD] 본 milestone 에서는 source 코드 변경 없이 신규 테스트 케이스만 추가 (per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run TDD discipline).

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel |
|---------|---------|-------|--------|-------|------|-----|----------|
| T-RT005-01 | RED: extend `audit_test.go` with 4 new test functions (orphan/registered/exception/struct-orphan) | `internal/config/audit_test.go` | REQ-008/021/043, AC-08 | expert-backend | - | ~80 | ✅ |
| T-RT005-02 | RED: extend `merge_test.go` with PolicyOverrideRejected + OverriddenBy + ZeroValuesExcluded + ByteStableJSON + StrictModeFalseAllowsOverride | `internal/config/merge_test.go` | REQ-005/012/022, AC-01/07/02 | expert-backend | - | ~150 | ✅ |
| T-RT005-03 | RED: extend `resolver_test.go` with ConfigTypeError (3 cases) + PolicyAbsent (3 cases) + SchemaVersion (3 cases) + ConfigAmbiguous (3 cases) + ConfigSchemaMismatch | `internal/config/resolver_test.go` | REQ-013/014/033/041/042, AC-05/06/11/12/15 | expert-backend | - | ~250 | ✅ |
| T-RT005-04 | RED: create `internal/config/reload_test.go` with TierIsolation + DeltaApplied + UnrelatedPathNoOp + ConcurrentReadSafe + SessionEnd_ClearsSessionTier + SessionEnd_NoSessionValuesNoError | `internal/config/reload_test.go` (new) | REQ-011/050, AC-04/13 | expert-backend | - | ~180 | ✅ |
| T-RT005-05 | RED: create `internal/config/audit_registry_test.go` with AllRegisteredStructsExist + NoUnexpectedYAMLOrphans | `internal/config/audit_registry_test.go` (new) | REQ-008/021, AC-08 | expert-backend | - | ~80 | ✅ |
| T-RT005-06 | RED: create `internal/cli/doctor_config_test.go` (or extend existing) with HappyPath + ByteStable + BuiltinOnly + DiffTier + DiffInvalidTier + FormatYAMLComments + KeySingleOutput + KeyNotFound + KeyInvalidFormat + BuiltinDefaultFlag + UserOverridesBuiltin | `internal/cli/doctor_config_test.go` (new or extend) | REQ-006/007/030/032, AC-02/03/09/10/14 | expert-backend | - | ~200 | ✅ |
| T-RT005-07 | Verify M1 baseline: run `go test ./internal/config/ ./internal/cli/` → confirm new tests RED + existing tests GREEN | (verification only) | M1 gate | expert-backend | T-RT005-01..06 | - | ❌ |

M1 산출물: ~14 신규 test 함수가 RED. Existing baseline 100% GREEN. 이 단계에서 `go build` 가 fail 가능 (RED 가 정확하면) — 정상.

---

## Milestone M2 — GREEN part 1 (Audit registry + ConfigTypeError + ConfigAmbiguous + schema_version)

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel |
|---------|---------|-------|--------|-------|------|-----|----------|
| T-RT005-08 | Create `internal/config/audit_registry.go` with `YAMLToStructRegistry` + `YAMLAuditExceptions` + `IsRegisteredOrException` helper | `internal/config/audit_registry.go` (new) | REQ-008/021/043 | expert-backend | T-RT005-05/07 | ~80 | ✅ |
| T-RT005-09 | Replace `audit_test.go::TestAuditParity` body with real registry-driven scan (replace `t.Skip`) | `internal/config/audit_test.go` | REQ-008/043, AC-08 | expert-backend | T-RT005-08 | ~50 (replace) | ❌ (depends T-08) |
| T-RT005-10 | Replace `resolver.go::loadYAMLFile` placeholder (line 252-255) with real `gopkg.in/yaml.v3` parsing + strict mode (`KnownFields(true)`) + schema_version extraction | `internal/config/resolver.go` | REQ-010/013/033 | expert-backend | T-RT005-07 | ~40 | ✅ (parallel with T-08/09) |
| T-RT005-11 | Extend `resolver.go::loadYAMLSections` (line 258-295) with yaml/yml sibling detection raising `ConfigAmbiguous` | `internal/config/resolver.go` | REQ-041, AC-12 | expert-backend | T-RT005-10 | ~30 | ❌ |
| T-RT005-12 | Extend `loadYAMLFile` to detect type mismatch via post-unmarshal type assertion → raise `ConfigTypeError` | `internal/config/resolver.go` | REQ-013, AC-05 | expert-backend | T-RT005-10 | ~25 | ❌ |
| T-RT005-13 | M2 verification: run `go test ./internal/config/...` → confirm `TestAuditParity_*`, `TestResolver_ConfigTypeError*`, `TestResolver_ConfigAmbiguous*`, `TestResolver_SchemaVersion*` GREEN | (verification only) | M2 gate | expert-backend | T-RT005-08..12 | - | ❌ |

M2 산출물: 5 GREEN AC (AC-05/08/11/12, AC-15 부분).

---

## Milestone M3 — GREEN part 2 (Real JSON dump + OverriddenBy + PolicyOverrideRejected + sorted YAML)

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel |
|---------|---------|-------|--------|-------|------|-----|----------|
| T-RT005-14 | Replace `merge.go::formatMapAsJSON` (line 271-275) + extend `dumpJSON` (line 240-258) to use `encoding/json.MarshalIndent` with `sort.Strings(keys)` for byte-stability | `internal/config/merge.go` | REQ-006/030, AC-02 | expert-backend | T-RT005-13 | ~50 (replace + extend) | ✅ |
| T-RT005-15 | Extend `merge.go::dumpYAML` (line 259-267) with alphabetical sort + maintain `# source: <tier>` comments | `internal/config/merge.go` | REQ-030, AC-09 | expert-backend | T-RT005-14 | ~20 | ❌ |
| T-RT005-16 | Add `(Provenance) MarshalJSON()` method in `provenance.go` for stable JSON output with `Source.String()` instead of int | `internal/config/provenance.go` | REQ-006 | expert-backend | T-RT005-14 | ~15 | ✅ (parallel with T-15) |
| T-RT005-17 | Extend `merge.go::MergeAll` (line 63-134) to verify `OverriddenBy` accumulates lower-tier non-zero values per REQ-012 (currently semantic-correct; add unit test assertion) | `internal/config/merge.go` | REQ-012, AC-01 | expert-backend | T-RT005-13 | ~10 (review + assertion strengthening) | ✅ |
| T-RT005-18 | Add `policy.strict_mode` enforcement in `MergeAll`: if SrcPolicy sets `policy.strict_mode: true` AND lower tier provides non-zero for any policy-designated key → raise `*PolicyOverrideRejected` + log `slog.Warn` | `internal/config/merge.go` | REQ-022, AC-07 | expert-backend | T-RT005-17 | ~30 | ❌ |
| T-RT005-19 | M3 verification: `go test ./internal/config/ -run TestMergeAll` → all merge tests GREEN | (verification only) | M3 gate | expert-backend | T-RT005-14..18 | - | ❌ |

M3 산출물: 4 GREEN AC (AC-01/02/07/09).

---

## Milestone M4 — GREEN part 3 (Diff-aware reload + Skill frontmatter + dedicated log + concurrency)

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel |
|---------|---------|-------|--------|-------|------|-----|----------|
| T-RT005-20 | Extend `resolver.go::resolver` struct with `mu sync.RWMutex`, `tierData map[Source]map[string]any`, `tierOrigins map[Source]string` for diff-aware reload caching | `internal/config/resolver.go` | REQ-011, concurrency | expert-backend | T-RT005-19 | ~10 | ✅ |
| T-RT005-21 | Wrap `Load()`, `Key()`, `Dump()`, `Diff()` with mutex (Lock for Load; RLock for read methods) | `internal/config/resolver.go` | concurrency, AC-04 | expert-backend | T-RT005-20 | ~20 | ❌ |
| T-RT005-22 | Add new method `(*resolver).Reload(path string) error` — determine tier by path prefix, re-parse only that tier, re-merge w/ cached data, atomic update under Lock | `internal/config/resolver.go` | REQ-011, AC-04 | expert-backend | T-RT005-21 | ~80 | ❌ |
| T-RT005-23 | Replace `resolver.go::loadSkillTier` placeholder (line 218-220) with `.claude/skills/**/SKILL.md` walk + frontmatter `config:` block parser using `gopkg.in/yaml.v3` | `internal/config/resolver.go` | REQ-001 (skill tier walked), REQ-015 partial | expert-backend | T-RT005-19 | ~80 | ✅ (parallel with T-20/21/22) |
| T-RT005-24 | Add `internal/config/log.go` (or extend resolver.go) with `logTierReadFailure(source, path, err)` helper writing append-mode to `.moai/logs/config.log`. Best-effort — fall back to stderr on log file inaccessible | `internal/config/log.go` (new) or `resolver.go` | REQ-040 | expert-backend | T-RT005-19 | ~30 | ✅ (parallel) |
| T-RT005-25 | Wire `logTierReadFailure` into each `loadXxxTier` on read failure (replaces or supplements current `slog.Warn`) | `internal/config/resolver.go` | REQ-040 | expert-backend | T-RT005-24 | ~15 | ❌ |
| T-RT005-26 | Add session-tier clear method `(*resolver).ClearSessionTier()` for SessionEnd integration (REQ-050 partial; full hook wire in RT-006) | `internal/config/resolver.go` | REQ-050, AC-13 | expert-backend | T-RT005-22 | ~20 | ❌ |
| T-RT005-27 | M4 verification: `go test ./internal/config/...` → GREEN; `go test -race ./internal/config/` → race detector clean | (verification only) | M4 gate | expert-backend | T-RT005-20..26 | - | ❌ |

M4 산출물: 2 GREEN AC (AC-04/13). Concurrency safety 보장.

---

## Milestone M5 — GREEN part 4 (Doctor CLI + filename norm + validator/v10) + Trackable

| Task ID | Subject | Files | REQ/AC | Owner | Deps | LOC | Parallel |
|---------|---------|-------|--------|-------|------|-----|----------|
| T-RT005-28 | Verify (or implement if missing) `internal/cli/doctor_config.go` `dump`/`diff`/`--key` subcommands; ensure call into resolver.Load/Dump/Diff/Key | `internal/cli/doctor_config.go` (verify or extend) | REQ-006/007/030/032, AC-02/03/09/10 | expert-backend | T-RT005-27 | ~50 (likely verify-only) | ✅ |
| T-RT005-29 | Wire `defaults.go::NewDefaultConfig()` into `loadBuiltinTier` (resolver.go:228-234 placeholder) — reflect-walk Config struct → `section.field` keys; populate `Provenance.Source = SrcBuiltin` | `internal/config/resolver.go` | REQ-001 (builtin tier walked), REQ-020 | expert-backend | T-RT005-27 | ~50 | ✅ |
| T-RT005-30 | Add `IsDefault()` flag to `MergedSettings.Dump` JSON output for SrcBuiltin keys (`"default": true`) | `internal/config/merge.go` | REQ-020, AC-14 | expert-backend | T-RT005-29 | ~15 | ❌ |
| T-RT005-31 | Apply `filepath.Abs()` normalization to all `Provenance.Origin` paths in tier loaders | `internal/config/resolver.go` | REQ-031 | expert-backend | T-RT005-27 | ~20 | ✅ |
| T-RT005-32 | Add `validate:"..."` tags on critical Config fields in `types.go` (User.Name required, Quality.DevelopmentMode oneof, LLM.PerformanceTier oneof, GitConvention.Convention oneof) — only if validator/v10 is in go.mod (check at task start) | `internal/config/types.go` | REQ-013 hardening | expert-backend | T-RT005-27 | ~30 | ✅ |
| T-RT005-33 | Extend `validation.go::Validate` to call `validator.New().Struct(cfg)` as first step; preserve custom checks as fallback | `internal/config/validation.go` | REQ-013 | expert-backend | T-RT005-32 | ~25 | ❌ |
| T-RT005-34 | Insert 7 MX tags per `plan.md` §6 across 6 files (`source.go`, `provenance.go`, `resolver.go` 2x, `merge.go` 2x, `audit_test.go`) | 6 files | mx_plan, Trackable | manager-docs | T-RT005-27..33 | ~7 lines | ❌ |
| T-RT005-35 | Add CHANGELOG entry under `## [Unreleased] / ### Added` per `plan.md` §M5f | `CHANGELOG.md` | Trackable (TRUST 5) | manager-docs | T-RT005-34 | ~5 lines | ❌ |
| T-RT005-36 | Run `make build` from repo root → verify `internal/template/embedded.go` regenerates cleanly (no template content drift) | (verification only) | build gate | expert-backend | T-RT005-35 | - | ❌ |
| T-RT005-37 | Run `go test ./...` from repo root → ALL tests GREEN, zero cascading regressions per `CLAUDE.local.md` §6 | (verification only) | M5 gate | expert-backend | T-RT005-36 | - | ❌ |
| T-RT005-38 | Run `go test -race ./internal/config/` → race detector clean | (verification only) | concurrency gate | expert-backend | T-RT005-37 | - | ❌ |
| T-RT005-39 | Run `go vet ./...` + `golangci-lint run` → 0 warnings | (verification only) | Lint gate | expert-backend | T-RT005-37 | - | ❌ |
| T-RT005-40 | Update `progress.md` with `run_complete_at: <ISO timestamp>` + `run_status: implementation-complete` + acceptance counter (15/15) | `progress.md` | Trackable | manager-docs | T-RT005-37..39 | ~20 lines update | ❌ |

M5 산출물: 4 GREEN AC (AC-10/14, AC-02 byte-stability strengthening). 모든 15 ACs GREEN. Trackable artifacts in place.

---

## Task Summary

| Milestone | Tasks | New LOC est | Modified LOC est | Owner mix |
|-----------|-------|-------------|------------------|-----------|
| M1 (RED) | 7 | ~940 (test only) | 0 | expert-backend |
| M2 (GREEN p1) | 6 | ~80 (audit_registry) | ~145 (audit/resolver) | expert-backend |
| M3 (GREEN p2) | 6 | ~15 (MarshalJSON) | ~125 (merge/provenance) | expert-backend |
| M4 (GREEN p3) | 8 | ~110 (reload/log) | ~145 (resolver) | expert-backend |
| M5 (GREEN p4 + Trackable) | 13 | ~50 (cli) | ~150 (resolver/types/validation) | expert-backend + manager-docs |

**Total**: 40 tasks, ~1,195 new LOC + ~565 modified LOC = ~1,760 LOC delta across 6 source + 5 test + 1 CHANGELOG file.

[NOTE] Plan §1.4 §3.1 §3.2 references "28 tasks" but milestone-level decomposition revealed 40 sub-tasks. The traceability matrix in plan.md §1.4 is task-level abstract; this tasks.md is the executable breakdown.

---

## Dependency Graph (Critical Path)

```
M1 (T-01..07) → M2 (T-08..13) → M3 (T-14..19) → M4 (T-20..27) → M5 (T-28..40)
                    ↓                  ↓               ↓               ↓
                AC-05/08/11/12   AC-01/02/07/09   AC-04/13       AC-10/14 + final
```

각 milestone 의 verification gate (T-07/13/19/27/37) 통과 후에만 다음 milestone 진입.

---

## Definition of Done (DoD) per Task

각 task 가 완료되었다고 mark 하기 전 다음 모두 충족:

1. Source 변경 시: 해당 파일의 godoc 이 REQ ID 인용 포함 (예: `// REQ-V3R2-RT-005-005: ...`).
2. Test 변경 시: 신규 test 가 sentinel error string 또는 expected value 명시적 assertion.
3. `go test ./internal/config/...` 가 해당 task 의 GREEN target 만족.
4. 다른 tests 회귀 없음 (per `CLAUDE.local.md` §6 "After fixing ANY test, run the FULL test suite").
5. `go vet` warning 없음 (lint gate 는 M5 에서 일괄 적용; 개별 task 는 `go vet` 만 점검).

---

## Out-of-Scope Tasks (defer to other SPECs)

| Task | Reason | Owner SPEC |
|------|--------|-----------|
| 5개 누락 yaml loader 본체 (constitution/context/interview/design/harness) | 본 SPEC 는 패턴만 establish; 실제 loader = MIG-003 owner | SPEC-V3R2-MIG-003 |
| ConfigChange hook handler wire | 본 SPEC 는 `Reload(path)` API 만 노출 | SPEC-V3R2-RT-006 |
| Permission stack consumer | 본 SPEC 는 `Source` enum + `Value[T]` 만; permission 의미론은 별개 | SPEC-V3R2-RT-002 |
| Sandbox routing by source | 본 SPEC 는 `Provenance.Source` 만 expose | SPEC-V3R2-RT-003 |
| Schema migration runner | 본 SPEC 는 `ConfigSchemaMismatch` 만 raise; migration 적용 = EXT-004 | SPEC-V3R2-EXT-004 |
| `sunset.yaml` activate-or-retire 결정 | 본 SPEC 는 audit exception 으로 marking; 실제 결정 = MIG-003 | SPEC-V3R2-MIG-003 |
| Plugin contribution mechanics | 본 SPEC 는 SrcPlugin slot 만; contributor API = v3.1+ | (deferred) |
| `fsnotify` 자동 file watcher | 본 SPEC 는 hook-triggered reload 만 | SPEC-V3R2-RT-006 |

---

End of tasks.md.
