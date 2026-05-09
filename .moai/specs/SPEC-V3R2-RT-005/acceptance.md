# SPEC-V3R2-RT-005 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 15 ACs (AC-V3R2-RT-005-01 through -15)     |
| 0.1.1   | 2026-05-10 | manager-spec (audit-fix iter 2) | +3 performance budget ACs (AC-V3R2-RT-005-16/17/18) per plan-auditor v1 audit defect D8 (spec §7 declares perf budgets but had zero benchmark tasks/ACs). AC-16 = Load p99 < 100ms (REQ-V3R2-RT-005 §7 Constraints), AC-17 = Reload p99 < 20ms, AC-18 = RSS < 2 MiB. Maps to T-RT005-43/44/45. |

---

## Scope

This document converts each of the 18 ACs (15 baseline from spec.md §6 + 3 performance budget ACs added in v0.1.1 from spec.md §7 Constraints per plan-auditor v1 audit defect D8) into Given/When/Then format with happy-path + edge-case + test-mapping notation.

Notation:
- **Test mapping** identifies which Go test function (or manual verification step) covers the AC.
- **Sentinel** is the literal error string the test expects on the negative path.

---

## AC-V3R2-RT-005-01 — Policy override wins, OverriddenBy populated

Maps to: REQ-V3R2-RT-005-005, REQ-V3R2-RT-005-012.

### Happy path

- **Given** a fresh `t.TempDir()` with:
  - `policy/settings.json` setting `permission.strict_mode: true`
  - `.moai/config/config.yaml` setting `permission.strict_mode: false`
- **When** the resolver `(*resolver).Load()` is called and `merged.Get("permission.strict_mode")` is invoked
- **Then** the returned `Value[any]` has:
  - `V == true`
  - `P.Source == config.SrcPolicy`
  - `P.Origin == "<policy_path>"` (absolute path)
  - `P.OverriddenBy == [".moai/config/config.yaml"]` (absolute path of the lower tier that had a non-zero conflicting value)
  - `P.Loaded` populated with the Load() invocation timestamp

### Edge case — three tiers, all set

- **Given** policy=true, user=false, project=true (all non-zero)
- **When** Load() merges
- **Then** winner=policy, OverriddenBy=[user_path, project_path] (alphabetical order by tier priority — user before project per AllSources())

### Edge case — zero values not in OverriddenBy

- **Given** policy=true, user="" (zero), project=true
- **When** Load() merges
- **Then** winner=policy, OverriddenBy=[project_path] (user excluded because zero)

### Test mapping

- `internal/config/merge_test.go::TestMergeAll_PolicyWins` (existing baseline; extends with OverriddenBy assertion)
- `internal/config/merge_test.go::TestMergeAll_OverriddenByPopulated` (new, M1)
- `internal/config/merge_test.go::TestMergeAll_ZeroValuesExcluded` (new, M1)

---

## AC-V3R2-RT-005-02 — `moai doctor config dump` JSON includes Provenance per key

Maps to: REQ-V3R2-RT-005-006.

### Happy path

- **Given** a project where all 8 tiers are populated (policy file present, user/project/local yaml present, builtin defaults always present)
- **When** the user runs `moai doctor config dump`
- **Then** stdout is valid JSON
- **And** every top-level key in the output has a sub-object containing `value`, `source`, `origin`, `loaded`, `overridden` fields
- **And** `source` value is one of `"policy"|"user"|"project"|"local"|"plugin"|"skill"|"session"|"builtin"`
- **And** the JSON is byte-stable across two consecutive invocations on the same on-disk state

### Edge case — `--format json` explicit

- **Given** the same state
- **When** `moai doctor config dump --format json`
- **Then** identical output to the default (json is default format)

### Edge case — partial population

- **Given** only `SrcBuiltin` tier populated (no user/project/local files)
- **When** `moai doctor config dump`
- **Then** every key has `source: "builtin"` and `origin: "internal/config/defaults.go"`

### Test mapping

- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_HappyPath` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_ByteStableAcrossCalls` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_BuiltinOnly` (new, M5)

---

## AC-V3R2-RT-005-03 — `config diff user project` lists divergent keys

Maps to: REQ-V3R2-RT-005-007, REQ-V3R2-RT-005-051.

### Happy path

- **Given** a project where:
  - `~/.moai/settings.json` sets `permission.strict_mode: true`, `permission.allowlist: [hostA]`
  - `.moai/config/config.yaml` sets `permission.strict_mode: false`, `permission.allowlist: [hostA]`, `coverage_threshold: 80`
- **When** the user runs `moai doctor config diff user project`
- **Then** stdout lists:
  - `permission.strict_mode` (different values)
  - `coverage_threshold` (only in project)
- **And** does NOT list `permission.allowlist` (identical in both)

### Edge case — merged-view delta

- **Given** the same state above
- **When** the diff command runs
- **Then** the output reflects merged-view semantics: keys are evaluated under full 8-tier merge, and the delta shows where the winner.Source differs between tier_a and tier_b OR the resolved value differs

### Edge case — invalid tier name

- **Given** the user runs `moai doctor config diff foo bar`
- **When** the command processes args
- **Then** exit code is non-zero
- **And** stderr contains `"invalid tier name \"foo\""` (matches `config.ParseSource` error)

### Test mapping

- `internal/cli/doctor_config_test.go::TestDoctorConfigDiff_TierComparison` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDiff_MergedViewDelta` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDiff_InvalidTier` (new, M5)

---

## AC-V3R2-RT-005-04 — ConfigChange hook triggers tier-isolated reload

Maps to: REQ-V3R2-RT-005-011.

### Happy path

- **Given** the resolver has been loaded and `merged.Get("quality.coverage_threshold")` returns `Value{V: 80, P.Loaded: t1}`
- **When** the user edits `.moai/config/sections/quality.yaml` changing `coverage_threshold: 80` → `90`
- **And** the orchestrator (or a test) calls `resolver.Reload(".moai/config/sections/quality.yaml")`
- **Then** the next `merged.Get("quality.coverage_threshold")` returns:
  - `V == 90`
  - `P.Loaded` is later than t1 (new timestamp)
- **And** keys belonging to other tiers (e.g., `~/.moai/...`) retain their original `P.Loaded` timestamp

### Edge case — file outside any tier

- **Given** Reload called with path `/random/unrelated.yaml`
- **When** Reload determines no matching tier
- **Then** Reload returns nil (no-op, optionally logs a debug message); merged state unchanged

### Edge case — concurrent Read during Reload

- **Given** goroutine A holds `RLock` for `Key()` call
- **And** goroutine B calls `Reload`
- **When** both run
- **Then** B blocks until A releases (sync.RWMutex semantics); no race detected by `go test -race`
- **And** after both complete, B's reload has applied

### Test mapping

- `internal/config/reload_test.go::TestResolver_Reload_TierIsolation` (new, M1)
- `internal/config/reload_test.go::TestResolver_Reload_DeltaApplied` (new, M1)
- `internal/config/reload_test.go::TestResolver_Reload_UnrelatedPathNoOp` (new, M1)
- `internal/config/reload_test.go::TestResolver_Reload_ConcurrentReadSafe` (new, M1, requires `go test -race`)

---

## AC-V3R2-RT-005-05 — Type mismatch raises ConfigTypeError naming file/key/expected

Maps to: REQ-V3R2-RT-005-013.

### Happy path

- **Given** `.moai/config/sections/quality.yaml` contains `coverage_threshold: "high"` (string where int expected)
- **When** the resolver `Load()` is called
- **Then** the returned error wraps `*ConfigTypeError`
- **And** `err.Error()` substring contains:
  - the file path (`.moai/config/sections/quality.yaml`)
  - the key name (`coverage_threshold`)
  - the expected type (`int`)
  - the actual value (`"high"`)

### Edge case — nested struct field

- **Given** `.moai/config/sections/quality.yaml` contains `tdd_settings.min_coverage_per_commit: "80%"` (string where int expected)
- **When** Load is called
- **Then** ConfigTypeError names the dotted key path

### Edge case — array type

- **Given** `.moai/config/sections/llm.yaml` contains `default_model: ["gpt-4"]` (array where string expected)
- **When** Load is called
- **Then** ConfigTypeError names `default_model` and expected type `string`

### Test mapping

- `internal/config/resolver_test.go::TestResolver_ConfigTypeError` (new, M1)
- `internal/config/resolver_test.go::TestResolver_ConfigTypeError_NestedField` (new, M1)
- `internal/config/resolver_test.go::TestResolver_ConfigTypeError_ArrayType` (new, M1)

---

## AC-V3R2-RT-005-06 — Policy file absent → empty tier, no error

Maps to: REQ-V3R2-RT-005-014.

### Happy path

- **Given** no policy file exists at `/etc/moai/settings.json` (Linux) / `/Library/Application Support/moai/settings.json` (macOS) / `%ProgramData%\moai\settings.json` (Windows)
- **When** `(*resolver).Load()` is called
- **Then** the call returns `(*MergedSettings, nil)` (no error)
- **And** every key in the merged settings has `P.Source != SrcPolicy` (no key was set by the absent tier)

### Edge case — policy file exists but is empty JSON `{}`

- **Given** an empty `policy/settings.json: {}`
- **When** Load is called
- **Then** SrcPolicy tier is technically present but contributes no keys
- **And** no error raised

### Edge case — policy file exists but unreadable (permissions)

- **Given** a policy file exists but is unreadable (e.g., `chmod 000`)
- **When** Load is called
- **Then** the loader logs a warning to `.moai/logs/config.log`
- **And** SrcPolicy tier is treated as empty
- **And** no error returned to caller (per REQ-V3R2-RT-005-040)

### Test mapping

- `internal/config/resolver_test.go::TestResolver_PolicyAbsentNoError` (new, M1)
- `internal/config/resolver_test.go::TestResolver_PolicyEmptyJSON` (new, M1)
- `internal/config/resolver_test.go::TestResolver_PolicyUnreadableLogs` (new, M1, may skip on Windows)

---

## AC-V3R2-RT-005-07 — Policy strict_mode rejects lower-tier override

Maps to: REQ-V3R2-RT-005-022.

### Happy path

- **Given** policy tier sets:
  - `policy.strict_mode: true`
  - `permission.network_allowlist: [host1]`
- **And** project tier sets:
  - `permission.network_allowlist: [host1, host2]` (attempted override)
- **When** `(*resolver).Load()` is called
- **Then** Load returns an error wrapping `*PolicyOverrideRejected`
- **And** the error message contains:
  - the key name (`permission.network_allowlist`)
  - the policy origin (the policy file path)
  - the attempted source (`SrcProject`)
- **And** a WARN log entry is recorded in `.moai/logs/config.log` documenting the rejection

### Edge case — strict_mode=false (or unset) allows override (still by priority)

- **Given** policy tier sets `policy.strict_mode: false`, `permission.network_allowlist: [host1]`
- **And** project tier sets `permission.network_allowlist: [host1, host2]`
- **When** Load is called
- **Then** policy still wins (higher priority); but no PolicyOverrideRejected error (only OverriddenBy logged)
- **And** project's path is added to `Provenance.OverriddenBy`

### Test mapping

- `internal/config/merge_test.go::TestMergeAll_PolicyOverrideRejected` (new, M1)
- `internal/config/merge_test.go::TestMergeAll_StrictModeFalseAllowsOverride` (new, M1)

---

## AC-V3R2-RT-005-08 — Audit test fails on orphan yaml file

Maps to: REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021, REQ-V3R2-RT-005-043.

### Happy path

- **Given** a test workspace where `.moai/config/sections/foo.yaml` exists but `foo` is not in `YAMLToStructRegistry` and not in `YAMLAuditExceptions`
- **When** `go test ./internal/config/... -run TestAuditParity` runs
- **Then** the test fails with sentinel `"orphan yaml file (no Go struct mapping): foo.yaml"`

### Edge case — registered yaml passes

- **Given** the current 16 registered yaml files (user, language, quality, ...) plus the 5 MIG-003-pending files registered as exceptions
- **When** TestAuditParity runs
- **Then** the test passes

### Edge case — orphan struct (struct registered but yaml missing)

- **Given** registry maps `"phantom" → "PhantomConfig"` but no `phantom.yaml` exists AND no PhantomConfig in Config struct
- **When** TestAuditParity runs
- **Then** the test fails with sentinel `"registry maps phantom → PhantomConfig but struct not found in Config"`

### Test mapping

- `internal/config/audit_test.go::TestAuditParity_OrphanYAMLFails` (new, M1)
- `internal/config/audit_test.go::TestAuditParity_AllRegisteredYAMLPass` (new, M1)
- `internal/config/audit_test.go::TestAuditParity_OrphanStructFails` (new, M1)
- `internal/config/audit_test.go::TestAuditParity_ExceptionsRespected` (new, M1)

---

## AC-V3R2-RT-005-09 — `--format yaml` includes `# source: <tier>` comments

Maps to: REQ-V3R2-RT-005-030.

### Happy path

- **Given** a project where:
  - `~/.moai/settings.json` sets `permission.strict_mode: true`
  - `.moai/config/sections/quality.yaml` sets `coverage_threshold: 85`
- **When** the user runs `moai doctor config dump --format yaml`
- **Then** stdout contains lines like:
  - `permission.strict_mode: true # source: user`
  - `coverage_threshold: 85 # source: project`

### Edge case — keys sorted alphabetically

- **Given** the same state
- **When** dump --format yaml runs
- **Then** keys appear in alphabetical order (cache-prefix discipline: byte-stable output)

### Edge case — comment placement

- **Given** a key whose value is a complex map (e.g., `auto_selection: {min_domains_for_team: 3}`)
- **When** dump --format yaml runs
- **Then** the `# source:` comment appears at the parent key's line, not inside the nested map

### Test mapping

- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_FormatYAML` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_KeysSortedAlphabetically` (new, M5)

---

## AC-V3R2-RT-005-10 — `--key permission.strict_mode` prints single key

Maps to: REQ-V3R2-RT-005-032.

### Happy path

- **Given** a populated resolver where `permission.strict_mode: true` from SrcUser tier
- **When** the user runs `moai doctor config dump --key permission.strict_mode`
- **Then** stdout contains exactly:
  - The key name `permission.strict_mode`
  - The value `true`
  - The source `user`
  - The origin path
  - The Loaded timestamp
- **And** stdout does NOT contain other keys (single-key output mode)

### Edge case — key not found

- **Given** the same state
- **When** the user runs `--key nonexistent.key`
- **Then** exit code non-zero
- **And** stderr contains `"key \"nonexistent.key\" not found in configuration"`

### Edge case — invalid key format

- **Given** the user runs `--key invalidformat` (no dot separator)
- **When** the command processes args
- **Then** exit code non-zero
- **And** stderr contains `"key must contain a dot separator"`

### Test mapping

- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_SingleKey` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_KeyNotFound` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_InvalidKeyFormat` (new, M5)

---

## AC-V3R2-RT-005-11 — schema_version propagated to Provenance

Maps to: REQ-V3R2-RT-005-033.

### Happy path

- **Given** `.moai/config/sections/quality.yaml` contains:
  ```yaml
  schema_version: 3
  coverage_threshold: 85
  development_mode: tdd
  ```
- **When** the resolver `Load()` is called
- **Then** for every key from quality.yaml (e.g., `quality.coverage_threshold`, `quality.development_mode`), `Value.P.SchemaVersion == 3`

### Edge case — schema_version absent

- **Given** the same yaml file without `schema_version`
- **When** Load runs
- **Then** `Value.P.SchemaVersion == 0` (zero default)

### Edge case — schema_version is non-integer

- **Given** the yaml contains `schema_version: "v3"` (string)
- **When** Load runs
- **Then** `ConfigTypeError` raised naming `schema_version` and expected type `int`

### Test mapping

- `internal/config/resolver_test.go::TestResolver_SchemaVersionPropagation` (new, M1)
- `internal/config/resolver_test.go::TestResolver_SchemaVersionAbsentZero` (new, M1)
- `internal/config/resolver_test.go::TestResolver_SchemaVersionInvalidType` (new, M1)

---

## AC-V3R2-RT-005-12 — Sibling files yaml/yml with conflict raise ConfigAmbiguous

Maps to: REQ-V3R2-RT-005-041.

### Happy path

- **Given** `.moai/config/sections/` contains both:
  - `quality.yaml` with `coverage_threshold: 80`
  - `quality.yml` with `coverage_threshold: 90`
- **When** the resolver `Load()` is called
- **Then** the returned error wraps `*ConfigAmbiguous`
- **And** the error message contains:
  - the key name (`coverage_threshold`)
  - both file paths (`quality.yaml` and `quality.yml`)

### Edge case — sibling files with identical values are accepted

- **Given** both `quality.yaml` and `quality.yml` with identical `coverage_threshold: 80`
- **When** Load runs
- **Then** no error (silent accept; documented behavior in resolver.go::loadYAMLSections godoc)

### Edge case — siblings in different sections

- **Given** `quality.yaml` and `state.yml` (different basenames)
- **When** Load runs
- **Then** no ambiguity (different section keyspaces)

### Test mapping

- `internal/config/resolver_test.go::TestResolver_ConfigAmbiguous` (new, M1)
- `internal/config/resolver_test.go::TestResolver_AmbiguousIdenticalAccepted` (new, M1)
- `internal/config/resolver_test.go::TestResolver_DifferentSectionsNoAmbiguity` (new, M1)

---

## AC-V3R2-RT-005-13 — SrcSession value reset on SessionEnd

Maps to: REQ-V3R2-RT-005-050.

### Happy path

- **Given** a session-scoped value `runtime.iter_id: "iter-007"` was written via `SrcSession` tier during a session
- **And** the resolver has loaded and merged this value
- **When** the SessionEnd hook fires (from SPEC-V3R2-RT-006)
- **Then** the resolver's session-tier cache is cleared
- **And** the next `(*resolver).Key("runtime", "iter_id")` returns `Value{}, false` (not found)

### Edge case — session value did not exist

- **Given** no session-tier value was set during the session
- **When** SessionEnd fires
- **Then** no error; resolver state unchanged

### Edge case — multiple session keys

- **Given** session tier set 3 keys
- **When** SessionEnd fires
- **Then** all 3 cleared atomically

### Test mapping

- `internal/config/reload_test.go::TestResolver_SessionEnd_ClearsSessionTier` (new, M1; covered by future RT-006 integration but stub-tested here)
- `internal/config/reload_test.go::TestResolver_SessionEnd_NoSessionValuesNoError` (new, M1)

[NOTE] This AC is partially out-of-scope for RT-005 per spec.md §2 ("session-scoped values reset at session end via SessionEnd hook from SPEC-V3R2-RT-006"). RT-005 ships the API surface (`(*resolver).ClearSessionTier()` method or equivalent); RT-006 wires the hook.

---

## AC-V3R2-RT-005-14 — Builtin defaults flagged as `"default": true` in dump

Maps to: REQ-V3R2-RT-005-020.

### Happy path

- **Given** `permission.pre_allowlist` has no override in any user/project/local/skill/session tier
- **And** `defaults.go::NewDefaultConfig()` provides a builtin default value
- **When** `moai doctor config dump` runs
- **Then** the JSON output for `permission.pre_allowlist` contains `"default": true` (or equivalent flag)
- **And** `"source": "builtin"`
- **And** `"origin": "internal/config/defaults.go"`

### Edge case — overridden by user tier

- **Given** `~/.moai/settings.json` overrides `permission.pre_allowlist`
- **When** dump runs
- **Then** the output for that key has `"source": "user"` and the `"default"` flag is absent or `false`

### Test mapping

- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_BuiltinDefaultFlag` (new, M5)
- `internal/cli/doctor_config_test.go::TestDoctorConfigDump_UserOverridesBuiltin` (new, M5)

---

## AC-V3R2-RT-005-15 — Schema migration mismatch raises ConfigSchemaMismatch

Maps to: REQ-V3R2-RT-005-042.

### Happy path

- **Given** a yaml file written under schema_version: 2 with `coverage_threshold: 85` (int type)
- **And** the current Go struct expects `coverage_threshold: string` under schema_version: 3
- **And** no migration is registered for the version transition (SPEC-V3R2-EXT-004 migration runner)
- **When** the resolver `Load()` is called
- **Then** the returned error wraps `*ConfigSchemaMismatch`
- **And** the error message contains:
  - the field name (`coverage_threshold`)
  - the old type (`int`)
  - the new type (`string`)
  - the missing migration version reference

### Edge case — migration registered

- **Given** the same field-type change with EXT-004 migration registered
- **When** Load runs
- **Then** the migration applies; no error

### Edge case — schema_version unchanged

- **Given** yaml schema_version matches Go struct schema_version (no transition needed)
- **When** Load runs
- **Then** no migration check; no error

### Test mapping

- `internal/config/resolver_test.go::TestResolver_ConfigSchemaMismatch` (new, M1)
- `internal/config/resolver_test.go::TestResolver_MigrationRegisteredAccepts` (new, M1; stub — full impl in EXT-004)

[NOTE] This AC depends partially on SPEC-V3R2-EXT-004 (migration runner). RT-005 raises the error type when the mismatch is detected; EXT-004 is responsible for registering and running migrations. The test in RT-005 stub-tests by simulating an unregistered migration.

---

## AC-V3R2-RT-005-16 — Cold-load p99 latency < 100ms (performance budget)

Maps to: spec.md §7 Constraints "Full Load() cold cache MUST complete in under 100 ms p99 for a typical project (23 yaml sections, all tiers populated)".

> Added in v0.1.1 (2026-05-10) per plan-auditor v1 audit defect D8 — spec §7 declared a hard performance budget but had zero corresponding benchmark task or AC. T-RT005-43 implements the benchmark.

### Happy path

- **Given** a synthetic typical project staged in `t.TempDir()` with 23 yaml sections × 8 tiers populated (using `defaults.go::NewDefaultConfig()` as builtin baseline)
- **When** `BenchmarkResolver_Load` runs with `-benchtime=10s` (≥ 100 iterations)
- **Then** the p99 latency (computed from sorted iteration histogram, index `int(0.99 * len(samples))`) is `< 100ms`
- **And** the benchmark reports the p50, p95, p99 values to `b.ReportMetric` for visibility in `go test -bench`

### Edge case — empty project (only builtin tier)

- **Given** an empty project (no user/project/local/skill yaml files)
- **When** `BenchmarkResolver_Load` runs
- **Then** p99 latency is well under budget (likely < 5ms; sanity floor)

### Edge case — large yaml sections (synthetic stress)

- **Given** a project with 23 yaml sections, each containing 100 keys (synthetic stress; ~2300 keys total)
- **When** Load runs once
- **Then** still < 100ms p99 (stress test for plausibility, not formal AC requirement)

### Test mapping

- `internal/config/resolver_bench_test.go::BenchmarkResolver_Load` (new, T-RT005-43)
- `internal/config/resolver_bench_test.go::BenchmarkResolver_Load_EmptyProject` (new, T-RT005-43; sanity floor)

---

## AC-V3R2-RT-005-17 — Diff-aware reload p99 latency < 20ms (performance budget)

Maps to: spec.md §7 Constraints "Diff-aware reload for a single file change MUST complete in under 20 ms p99".

> Added in v0.1.1 (2026-05-10) per plan-auditor v1 audit defect D8. T-RT005-44 implements the benchmark.

### Happy path

- **Given** a fully-loaded resolver from the same synthetic project as AC-16
- **When** `BenchmarkResolver_Reload` runs `(*resolver).Reload(".moai/config/sections/quality.yaml")` with `-benchtime=10s` (≥ 100 iterations) — each iteration mutates the file mid-benchmark via fixture rotation
- **Then** p99 latency is `< 20ms`
- **And** the benchmark reports p50, p95, p99 metrics to `b.ReportMetric`

### Edge case — reload on path outside any tier

- **Given** Reload called with `/nonexistent/random.yaml`
- **When** Reload returns no-op
- **Then** latency is well under budget (likely < 1ms; only path-prefix matching cost)

### Test mapping

- `internal/config/resolver_bench_test.go::BenchmarkResolver_Reload` (new, T-RT005-44)
- `internal/config/resolver_bench_test.go::BenchmarkResolver_Reload_NoOp` (new, T-RT005-44; sanity floor)

---

## AC-V3R2-RT-005-18 — Merged settings RSS < 2 MiB (memory budget)

Maps to: spec.md §7 Constraints "Memory: Merged settings representation MUST NOT exceed 2 MiB RSS for typical projects".

> Added in v0.1.1 (2026-05-10) per plan-auditor v1 audit defect D8. T-RT005-45 implements the test.

### Happy path

- **Given** the same synthetic typical project as AC-16 (23 yaml sections × 8 tiers populated)
- **When** `TestResolver_MemoryFootprint` runs:
  1. `runtime.GC()` to baseline
  2. `runtime.ReadMemStats(&before)`
  3. resolver, _ := NewResolver(); merged, _ := resolver.Load()
  4. `runtime.ReadMemStats(&after)`
  5. delta := after.HeapAlloc - before.HeapAlloc
- **Then** delta is `< 2 * 1024 * 1024` bytes (2 MiB)

### Edge case — no Provenance overhead inflation

- **Given** the same synthetic project, but Provenance struct has 0 OverriddenBy entries (single-tier scenario)
- **When** memory footprint measured
- **Then** still < 2 MiB (Provenance per-key overhead ≤ 256 bytes per spec §8 risk row 2)

### Edge case — stress: 100 keys × 8 tiers all populated

- **Given** synthetic project where every key has all 8 tiers contributing (max OverriddenBy entries)
- **When** memory measured
- **Then** still < 2 MiB (verifies spec §8 risk row 2 calculation: 100 keys × 256 bytes Provenance = 25 KiB; well under 2 MiB ceiling)

### Test mapping

- `internal/config/resolver_bench_test.go::TestResolver_MemoryFootprint` (new, T-RT005-45)
- `internal/config/resolver_bench_test.go::TestResolver_MemoryFootprint_StressOverriddenBy` (new, T-RT005-45)

[NOTE] Memory measurement via `runtime.MemStats.HeapAlloc` is approximate; OS-level RSS (e.g., `process.RSS`) is harder to measure cross-platform. The test uses HeapAlloc as a proxy; the 2 MiB constraint is a Go-heap budget, not a process-RSS budget.

---

## Summary table — AC → REQ → Test

| AC | REQs covered | Test files |
|----|--------------|------------|
| AC-01 | REQ-005, REQ-012 | `merge_test.go::TestMergeAll_PolicyWins`, `TestMergeAll_OverriddenByPopulated`, `TestMergeAll_ZeroValuesExcluded` |
| AC-02 | REQ-006 | `doctor_config_test.go::TestDoctorConfigDump_HappyPath`, `TestDoctorConfigDump_ByteStableAcrossCalls`, `TestDoctorConfigDump_BuiltinOnly` |
| AC-03 | REQ-007, REQ-051 | `doctor_config_test.go::TestDoctorConfigDiff_TierComparison`, `TestDoctorConfigDiff_MergedViewDelta`, `TestDoctorConfigDiff_InvalidTier` |
| AC-04 | REQ-011 | `reload_test.go::TestResolver_Reload_*` (4 cases) |
| AC-05 | REQ-013 | `resolver_test.go::TestResolver_ConfigTypeError*` (3 cases) |
| AC-06 | REQ-014 | `resolver_test.go::TestResolver_PolicyAbsent*`, `TestResolver_PolicyEmptyJSON`, `TestResolver_PolicyUnreadableLogs` |
| AC-07 | REQ-022 | `merge_test.go::TestMergeAll_PolicyOverrideRejected`, `TestMergeAll_StrictModeFalseAllowsOverride` |
| AC-08 | REQ-008, REQ-021, REQ-043 | `audit_test.go::TestAuditParity_*` (4 cases) |
| AC-09 | REQ-030 | `doctor_config_test.go::TestDoctorConfigDump_FormatYAML`, `TestDoctorConfigDump_KeysSortedAlphabetically` |
| AC-10 | REQ-032 | `doctor_config_test.go::TestDoctorConfigDump_SingleKey*` (3 cases) |
| AC-11 | REQ-033 | `resolver_test.go::TestResolver_SchemaVersion*` (3 cases) |
| AC-12 | REQ-041 | `resolver_test.go::TestResolver_ConfigAmbiguous*` (3 cases) |
| AC-13 | REQ-050 | `reload_test.go::TestResolver_SessionEnd_*` (2 cases, stub for RT-006) |
| AC-14 | REQ-020 | `doctor_config_test.go::TestDoctorConfigDump_BuiltinDefaultFlag`, `TestDoctorConfigDump_UserOverridesBuiltin` |
| AC-15 | REQ-042 | `resolver_test.go::TestResolver_ConfigSchemaMismatch`, `TestResolver_MigrationRegistered* ` (stub for EXT-004) |
| AC-16 | spec §7 Constraints (cold load p99 < 100ms) | `resolver_bench_test.go::BenchmarkResolver_Load`, `BenchmarkResolver_Load_EmptyProject` (new in v0.1.1, T-RT005-43) |
| AC-17 | spec §7 Constraints (reload p99 < 20ms) | `resolver_bench_test.go::BenchmarkResolver_Reload`, `BenchmarkResolver_Reload_NoOp` (new in v0.1.1, T-RT005-44) |
| AC-18 | spec §7 Constraints (RSS < 2 MiB) | `resolver_bench_test.go::TestResolver_MemoryFootprint`, `TestResolver_MemoryFootprint_StressOverriddenBy` (new in v0.1.1, T-RT005-45) |

Total new test functions: **~44 across 4 existing test files extended (`audit_test.go`, `merge_test.go`, `resolver_test.go`, `doctor_config_test.go`) and 2 new test files (`reload_test.go`, `resolver_bench_test.go`)** (was ~38 in v0.1.0; +6 added in v0.1.1: 2 benchmarks × 2 each = 4 + 2 memory tests = 6).

---

## Definition of Done

This SPEC is considered done when ALL of the following are true:

1. All 18 ACs above pass under `go test ./internal/config/ ./internal/cli/` (15 baseline + 3 perf budget AC-16/17/18 added in v0.1.1).
2. Full `go test ./...` from the repo root passes with zero failures and zero cascading regressions.
3. `go test -race ./internal/config/` passes (mutex correctness verified).
4. `make build` succeeds and `internal/template/embedded.go` regenerates cleanly (no template content drift).
5. `go vet ./...` and `golangci-lint run` pass with zero warnings.
6. `progress.md` is updated with `run_complete_at: <timestamp>` and `run_status: implementation-complete`.
7. CHANGELOG entry is present under `## [Unreleased] / ### Added`.
8. 7 MX tags are inserted per `plan.md` §6 (3 ANCHOR with measured fan_in, 2 NOTE, 2 WARN).
9. The PR opened by `manager-git` has all required CI checks green (Lint, Test ubuntu/macos/windows, Build all 5, CodeQL).
10. `audit_test.go::TestAuditParity` is no longer a `t.Skip()` placeholder — it actively scans `.moai/config/sections/*.yaml` and enforces yaml↔struct parity.
11. The 5 yaml-without-loader sections (constitution/context/interview/design/harness) are registered in `audit_registry.go::YAMLAuditExceptions` with rationale strings, ready for MIG-003 to remove entries one-by-one.
12. `resolver.go::loadYAMLFile` no longer returns an empty placeholder map — it parses real yaml via `gopkg.in/yaml.v3`.

---

End of acceptance.md.
