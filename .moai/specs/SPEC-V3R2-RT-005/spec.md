---
id: SPEC-V3R2-RT-005
title: "Multi-Layer Settings Resolution with Provenance Tags"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 1 — Constitution & Foundation"
module: "internal/config/"
dependencies:
  - SPEC-V3R2-CON-001
bc_id: [BC-V3R2-015]
related_principle: [P6 Permission Bubble, P7 Sandbox Default, P12 Constitutional Governance]
related_pattern: [X-2, S-1, T-5]
related_problem: [P-C04, P-H06]
related_theme: "Layer 3: Runtime"
breaking: true
lifecycle: spec-anchored
tags: "settings, provenance, multi-layer, v3r2, breaking, runtime, config"
---

# SPEC-V3R2-RT-005: Multi-Layer Settings Resolution with Provenance Tags

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft. New SPEC — no v3-legacy predecessor. Addresses P-C04 (no config provenance, HIGH) and prerequisites P-H06 (5 yaml sections without Go loaders, CRITICAL). |

---

## 1. Goal (목적)

Replace moai's implicit two-tier config model (`~/.moai/` + `.moai/`) with an 8-tier deterministic settings resolver whose merged output carries `Source` provenance on every key. Every configuration value that enters the runtime representation answers "which file set this?" without grep — the tag travels with the value. This enables targeted diagnostics (`/moai doctor config dump` reports per-key origin), safe migrations (rewrite `userSettings` without disturbing `projectSettings`), plugin-hook dedup (rules tagged with plugin name), and a single substrate underneath SPEC-V3R2-RT-002 (permission stack) and SPEC-V3R2-RT-003 (sandbox routing).

Master §5.5 declares this as a cross-layer concern: "Multi-layer settings with provenance is the prerequisite for S-1 (permission stack), S-2 (bubble mode), and sandbox routing by source." The eight tiers are named in master §4.3 Layer 3 type block: `policy > user > project > local > plugin > skill > session > builtin`. Master §8 BC-V3R2-015 documents the migration: reader layer only; settings.json files on disk remain unchanged; flat-merge consumers are internal and require no user API change.

This SPEC additionally addresses a CRITICAL gap from r6 §5.2: 5 yaml sections (`constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml`, `harness.yaml`) have no Go loader today, forcing them to be template-only artifacts with zero runtime enforcement. This SPEC establishes the pattern every loader must follow (typed struct + source-tagged merge + validator tag); the actual loader bodies ship in SPEC-V3R2-MIG-003, but the tiering contract lives here.

## 2. Scope (범위)

In-scope:

- `Source` typed enum (shared with SPEC-V3R2-RT-002 permission stack) with 8 values in priority order: `SrcPolicy`, `SrcUser`, `SrcProject`, `SrcLocal`, `SrcPlugin`, `SrcSkill`, `SrcSession`, `SrcBuiltin`.
- `Provenance` struct carrying `Source Source`, `Origin string` (absolute file path), `Loaded time.Time`, optional `SchemaVersion int`.
- Generic typed `Value[T any]` wrapper: `Value[T]{V T, P Provenance}` so every config field answers both "what" and "where from".
- `SettingsResolver` interface: `Load() (MergedSettings, error)`, `Key(section, field string) Value[any]`, `Dump(writer io.Writer) error`.
- Reader layer reads from 4+ canonical file paths per tier:
  - `SrcPolicy`: `/etc/moai/settings.json` (Linux) / `/Library/Application Support/moai/settings.json` (macOS) / `%ProgramData%\moai\settings.json` (Windows).
  - `SrcUser`: `~/.moai/settings.json` + `~/.moai/config/sections/*.yaml`.
  - `SrcProject`: `.moai/config/config.yaml` + `.moai/config/sections/*.yaml`.
  - `SrcLocal`: `.claude/settings.local.json` + `.moai/config/local/*.yaml`.
  - `SrcPlugin`: reserved (no plugins v3.0 per master §7 plugin NOT-NOW) — slot exists for v3.1+.
  - `SrcSkill`: `.claude/skills/**/SKILL.md` frontmatter `config:` block.
  - `SrcSession`: runtime-populated by SPEC-V3R2-RT-004 checkpoint writes (short-lived).
  - `SrcBuiltin`: compiled-in defaults in `internal/config/defaults.go`.
- Deterministic merge algorithm: for every `(section, field)` key, walk tiers in priority order; first non-zero value wins; provenance is captured from the winning tier.
- Diff-aware reload: when watched files change (hook-triggered via SPEC-V3R2-RT-006 ConfigChange), only tiers that touched files re-resolve; unaffected keys retain their provenance.
- `moai doctor config dump` subcommand: prints JSON with every key and its Provenance.
- `moai doctor config diff <tier-a> <tier-b>` subcommand: prints keys that differ between two tiers.
- CI rule: every new yaml section under `.moai/config/sections/` MUST ship with a Go struct in `internal/config/types.go` AND a loader in `internal/config/loader.go` AND a test in `loader_test.go`. Enforced by `internal/config/audit_test.go` (new file).

Out-of-scope (addressed by other SPECs):

- Permission stack resolution (consumes the `Source` enum from here) — SPEC-V3R2-RT-002.
- The 5 actual missing loaders (`constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml`, `harness.yaml`) — SPEC-V3R2-MIG-003.
- Sandbox routing by source — SPEC-V3R2-RT-003.
- Hook-triggered reload wiring — SPEC-V3R2-RT-006 (ConfigChange handler upgrade).
- `sunset.yaml` activate-or-retire decision — SPEC-V3R2-MIG-003.
- Plugin contribution mechanics — deferred to v3.1+.

## 3. Environment (환경)

Current moai-adk state:

- No provenance tracking per problem-catalog.md P-C04: "Two-tier config (~/.moai/ + .moai/) lacks provenance tracking — 'which file set this?' is opaque."
- 5 yaml sections template-only per r6 §5.2: constitution, context, interview, design, harness. 1 dormant (sunset). 1 partial (workflow.yaml — only role_profiles are read).
- 13 yaml sections have Go loaders today (language, llm, quality, workflow-partial, lsp, mx, security, statusline, system, user, project, git-convention, git-strategy, ralph, research, state per r6 §5.1 enumeration).
- Settings files currently loaded ad-hoc by each consumer; no unified resolver or merger exists.
- `.claude/settings.json` flat `permissions.allow` is the only permission-adjacent config today.

Claude Code reference:

- r3 §1.3: "hooks settings precedence — policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules > hookDecision".
- r3 §2 Decision 11: "Provenance on every configuration element. `settingSource`, `pluginRoot`, `skillRoot`, `policyPinned` flow through every hook, command, and permission rule."
- r3 §4 Adopt 1: "Multi-layer settings with explicit precedence and provenance."

Wave 1-2 sources:

- design-principles.md P6 rationale: "Every config value carries a source tag."
- pattern-library.md X-2 (Multi-Layer Settings with Provenance Tags, priority 4): "Prerequisite for S-1, S-2, S-3."
- problem-catalog.md Cluster 4 (Config Schema Completeness, CRITICAL): P-H06, P-H07, P-H20, P-C04.

Affected modules:

- `internal/config/types.go` — extend with `Value[T]` generic + `Provenance`.
- `internal/config/loader.go` — resolver orchestration.
- `internal/config/source.go` — new file, Source enum.
- `internal/config/merge.go` — new file, deterministic 8-tier merger.
- `internal/config/audit_test.go` — new file, CI rule for yaml↔Go loader parity.
- `internal/cli/doctor.go` — `config dump` / `config diff` subcommands.
- `.moai/config/config.yaml` — unchanged on disk; structure preserved.

## 4. Assumptions (가정)

- The 8-tier priority ordering is authoritative and shared with SPEC-V3R2-RT-002 permission stack; divergence between the two would be a CRITICAL constitutional bug.
- Go 1.22+ generics support for `Value[T any]` — existing go.mod already targets 1.22.
- `validator/v10` from SPEC-V3R2-SCH-001 handles typed struct validation; provenance fields carry their own struct tags (never `json:"-"`) so they serialize in `doctor config dump`.
- Policy tier paths follow platform conventions; when the policy file is absent, the tier is treated as empty (not an error).
- Plugin tier is a schema slot with no contributors in v3.0; the tier is walked but always empty.
- File-watch for diff-aware reload uses `fsnotify`/equivalent; the hook-triggered reload (SPEC-V3R2-RT-006) avoids the need for always-on watcher in non-session contexts.
- Merge is deterministic: given the same 8 tier inputs, the output is byte-stable. This enables cache-prefix discipline at the merged-settings layer (P-C05 secondary benefit).
- Existing consumers of `internal/config` (`Load*` functions across packages) are refactored to accept `Value[T]` OR call `.V` to extract; migration is mechanical and BC-V3R2-015 declares no user API change (consumers are internal).
- Legacy flat `permissions.allow` in `.claude/settings.json` is read into the `SrcProject` or `SrcLocal` tier depending on file location, and `Provenance.Origin` is the absolute path.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-V3R2-RT-005-001: The `Source` type SHALL be a typed enum with exactly 8 values ordered by priority (highest first): `SrcPolicy`, `SrcUser`, `SrcProject`, `SrcLocal`, `SrcPlugin`, `SrcSkill`, `SrcSession`, `SrcBuiltin`.
- REQ-V3R2-RT-005-002: Every `Provenance` instance SHALL include non-empty `Source`, non-empty `Origin` (absolute file path), and populated `Loaded` timestamp fields.
- REQ-V3R2-RT-005-003: The generic `Value[T]` wrapper SHALL expose `V T` and `P Provenance` fields and the methods `Unwrap() T` and `Origin() string`.
- REQ-V3R2-RT-005-004: The `SettingsResolver` interface SHALL expose `Load() (MergedSettings, error)`, `Key(section, field string) Value[any]`, `Dump(io.Writer) error`, `Diff(a, b Source) map[string]Value[any]`.
- REQ-V3R2-RT-005-005: The deterministic merge algorithm SHALL walk tiers in priority order and return the first non-zero value per `(section, field)` key; provenance SHALL be captured from the winning tier.
- REQ-V3R2-RT-005-006: `moai doctor config dump` SHALL print every merged key with its `Provenance` in JSON or YAML format.
- REQ-V3R2-RT-005-007: `moai doctor config diff <tier-a> <tier-b>` SHALL print keys whose values differ between the two named tiers.
- REQ-V3R2-RT-005-008: The CI rule in `internal/config/audit_test.go` SHALL fail the build when a new yaml section under `.moai/config/sections/` exists without a corresponding Go struct and loader.

### 5.2 Event-Driven Requirements

- REQ-V3R2-RT-005-010: WHEN `SettingsResolver.Load()` is called, the system SHALL read all 8 tier sources in priority order and produce a merged representation with per-key provenance.
- REQ-V3R2-RT-005-011: WHEN a ConfigChange hook fires (from SPEC-V3R2-RT-006) naming a specific file path, the resolver SHALL re-load only the tier containing that path and merge the delta into the existing representation.
- REQ-V3R2-RT-005-012: WHEN two tiers provide non-zero values for the same `(section, field)`, the higher-priority tier's value SHALL be retained and the overridden tier's path SHALL be added to `Provenance.OverriddenBy`.
- REQ-V3R2-RT-005-013: WHEN a value cannot be parsed against its typed schema (e.g., string where int expected), the resolver SHALL return error `ConfigTypeError` naming the offending file, key, and expected type.
- REQ-V3R2-RT-005-014: WHEN the policy tier file is absent (no `/etc/moai/settings.json`), the resolver SHALL treat `SrcPolicy` as an empty tier without raising an error.
- REQ-V3R2-RT-005-015: WHEN a plugin contributes a settings fragment (v3.1+ feature slot), the fragment SHALL be tagged with `Source: SrcPlugin` and `Origin` naming the plugin package directory.

### 5.3 State-Driven Requirements

- REQ-V3R2-RT-005-020: WHILE the merged settings contain a value whose `Source: SrcBuiltin` provenance is recorded, `moai doctor config dump` SHALL mark it as `"default"` in a human-readable flag to aid diagnostics.
- REQ-V3R2-RT-005-021: WHILE the `.moai/config/sections/*.yaml` file count differs from the Go struct count in `internal/config/types.go`, `internal/config/audit_test.go` SHALL fail with a specific diff message naming the absent side.
- REQ-V3R2-RT-005-022: WHILE `policy.strict_mode: true` is set in the policy tier, any lower tier's attempt to override a policy-designated key SHALL be rejected with error `PolicyOverrideRejected` and the override logged.

### 5.4 Optional Features

- REQ-V3R2-RT-005-030: WHERE `moai doctor config dump --format yaml` is invoked, the system SHALL emit YAML output with `# source: <tier>` comments adjacent to each key.
- REQ-V3R2-RT-005-031: WHERE a tier uses absolute paths in its `Origin` field, the system SHALL normalize the path via `filepath.Abs` to ensure diagnostics are portable.
- REQ-V3R2-RT-005-032: WHERE `moai doctor config --key permission.strict_mode` is invoked, the system SHALL print only the Value and Provenance for that single key.
- REQ-V3R2-RT-005-033: WHERE a yaml section file declares a `schema_version: N` key, `Provenance.SchemaVersion` SHALL be populated; otherwise it SHALL be `0`.

### 5.5 Unwanted Behavior

- REQ-V3R2-RT-005-040: IF a tier file exists but cannot be opened (permissions error), THEN the loader SHALL skip the tier with a warning logged to `.moai/logs/config.log`, never silently default.
- REQ-V3R2-RT-005-041: IF two sibling files in the same tier (e.g., `.moai/config/sections/quality.yaml` and `.moai/config/sections/quality.yml`) define the same key with conflicting values, THEN the loader SHALL fail with `ConfigAmbiguous` naming both files.
- REQ-V3R2-RT-005-042: IF a value's Go type changes between schema versions without a migration (SPEC-V3R2-EXT-004), THEN the loader SHALL fail on read with `ConfigSchemaMismatch` naming the field, old type, new type, and missing migration number.
- REQ-V3R2-RT-005-043: IF a new yaml section is added under `.moai/config/sections/` without a companion Go struct, THEN `internal/config/audit_test.go` SHALL fail the build naming the orphan yaml file.

### 5.6 Complex Requirements

- REQ-V3R2-RT-005-050: WHILE a ConfigChange hook arrives for a file in the `SrcSession` tier, WHEN the resolver re-merges, THEN the delta SHALL NOT persist beyond the current session (session-scoped values reset at session end via SessionEnd hook from SPEC-V3R2-RT-006).
- REQ-V3R2-RT-005-051: WHILE `moai doctor config diff` is invoked with `<tier-a> = user` and `<tier-b> = project`, WHEN the output is produced, THEN the tool SHALL show the merged-view delta (not just raw file delta), reflecting how resolved values differ between the two tiers in the merge context.

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-005-01: Given `SrcPolicy` sets `permission.strict_mode: true` and `SrcProject` sets `permission.strict_mode: false`, When `Load()` returns, Then the merged value is `true` with `Provenance.Source: SrcPolicy` and `Provenance.OverriddenBy: [".moai/config/config.yaml"]`. (maps REQ-V3R2-RT-005-005, -012)
- AC-V3R2-RT-005-02: Given `moai doctor config dump` runs on a project with all 8 tiers populated, When stdout captured, Then every key has `{V, P: {Source, Origin, Loaded}}` structure in JSON. (maps REQ-V3R2-RT-005-006)
- AC-V3R2-RT-005-03: Given `moai doctor config diff user project`, When output captured, Then keys present only in one tier AND keys with different values are listed; keys identical in both are omitted. (maps REQ-V3R2-RT-005-007, -051)
- AC-V3R2-RT-005-04: Given a ConfigChange hook fires with path `.moai/config/sections/quality.yaml`, When resolver re-loads, Then only `SrcProject` tier re-parses and the merged representation reflects the new value with updated `Loaded` timestamp. (maps REQ-V3R2-RT-005-011)
- AC-V3R2-RT-005-05: Given `quality.yaml` has `coverage_threshold: "high"` (string) where schema expects int, When loader runs, Then error `ConfigTypeError` is returned naming the file, key, and expected type. (maps REQ-V3R2-RT-005-013)
- AC-V3R2-RT-005-06: Given `/etc/moai/settings.json` does not exist, When `Load()` is called, Then no error is raised and `SrcPolicy` tier is empty (REQ-V3R2-RT-005-014).
- AC-V3R2-RT-005-07: Given `SrcPolicy` has `policy.strict_mode: true` and `permission.network_allowlist = [host1]`, When `SrcProject` tries to override `permission.network_allowlist`, Then override is rejected with error `PolicyOverrideRejected` and logged. (maps REQ-V3R2-RT-005-022)
- AC-V3R2-RT-005-08: Given a new file `.moai/config/sections/foo.yaml` is added without a Go struct, When `go test ./internal/config/... -run TestAuditParity` runs, Then the test fails naming `foo.yaml` as orphan. (maps REQ-V3R2-RT-005-043)
- AC-V3R2-RT-005-09: Given `moai doctor config dump --format yaml`, When stdout captured, Then each merged key is followed by `# source: <tier>` comment. (maps REQ-V3R2-RT-005-030)
- AC-V3R2-RT-005-10: Given `moai doctor config --key permission.strict_mode`, When invoked, Then only that single key's value + provenance is printed. (maps REQ-V3R2-RT-005-032)
- AC-V3R2-RT-005-11: Given `quality.yaml` declares `schema_version: 3`, When loader reads it, Then `Provenance.SchemaVersion == 3` on every key from that file. (maps REQ-V3R2-RT-005-033)
- AC-V3R2-RT-005-12: Given two sibling files `quality.yaml` and `quality.yml` both set the same key with different values, When loader runs, Then error `ConfigAmbiguous` is returned naming both files. (maps REQ-V3R2-RT-005-041)
- AC-V3R2-RT-005-13: Given a session-scoped value was set via `SrcSession` tier during session, When SessionEnd hook fires, Then the value is cleared from the resolver's state. (maps REQ-V3R2-RT-005-050)
- AC-V3R2-RT-005-14: Given a builtin default value for `permission.pre_allowlist`, When `moai doctor config dump` prints it, Then the output includes a human-readable flag `"default": true`. (maps REQ-V3R2-RT-005-020)
- AC-V3R2-RT-005-15: Given a field changed from `int` to `string` in the schema without migration, When loader reads old file, Then `ConfigSchemaMismatch` is returned naming field, old/new type, and the missing migration version. (maps REQ-V3R2-RT-005-042)

## 7. Constraints (제약)

- Technical: Go 1.22+ (generics for `Value[T]`); no new external dependencies beyond validator/v10. `fsnotify` already used elsewhere in moai for file watching.
- Backward compat: v2.x internal consumers (Load* calls in various packages) migrate via mechanical refactor to `Value[T]` wrapper OR `.V` unwrap. No user-facing API change per master §8 BC-V3R2-015.
- Platform: macOS / Linux / Windows. Policy paths are platform-specific; detection via `runtime.GOOS`.
- Performance: Full `Load()` cold cache MUST complete in under 100 ms p99 for a typical project (23 yaml sections, all tiers populated). Diff-aware reload for a single file change MUST complete in under 20 ms p99.
- Memory: Merged settings representation MUST NOT exceed 2 MiB RSS for typical projects.
- Determinism: Identical tier inputs MUST produce byte-identical merged output to enable prompt-cache prefix stability (P-C05 secondary benefit).
- Error handling: per REQ-V3R2-RT-005-040, tier read failures are never silently ignored — they produce warnings in `.moai/logs/config.log` and skip the tier.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Internal consumer refactor from `int` to `Value[int]` touches many files | M | M | Mechanical pattern; `Value[T].V` accessor keeps call sites unchanged; migration script in SPEC-V3R2-MIG-001. |
| Provenance metadata bloats in-memory config representation | L | L | 2 MiB RSS ceiling per constraint; typical projects at ~100 keys × 256-byte Provenance record = 25 KiB overhead. |
| Audit test over-reports when yaml and Go struct name diverge legitimately | M | L | Test maps yaml filename → Go struct via explicit registry; exceptions documented in `internal/config/audit_registry.go`. |
| Diff-aware reload misses cascading dependencies between tiers | L | M | Cold reload fallback available via `moai doctor config reload-all`; file-watch integration re-tested at alpha.2. |
| Plugin tier empty slot confuses future v3.1 contributors | L | L | SPEC-V3R2-EXT-003 documents the plugin schema slot for v4; v3.0 ships the tier wired but empty. |
| `SrcPolicy` strict mode locks out legitimate local overrides | M | M | `PolicyOverrideRejected` log is explicit; enterprise rollouts document policy intentionally. |
| Users reach for `~/.moai/settings.json` while the active config is `.moai/config/config.yaml` | H | M | `moai doctor config dump` makes the active tier obvious; `moai doctor config --key <name>` reports single-key origin. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-SCH-001 (validator/v10 integration).
- SPEC-V3R2-CON-001 (FROZEN-zone codification — the 8-tier ordering is declared as a constitutional invariant).
- SPEC-V3R2-RT-004 (SessionStore provides `SrcSession` tier contents via runtime checkpoint writes).

### 9.2 Blocks

- SPEC-V3R2-RT-002 (permission stack consumes the `Source` enum and the same 8-tier priority).
- SPEC-V3R2-RT-003 (sandbox routing by source — e.g., `SrcPolicy` sandbox defaults override `SrcProject`).
- SPEC-V3R2-RT-006 (ConfigChange hook consumes the diff-aware reload API).
- SPEC-V3R2-RT-007 (hardcoded path migration — GoBinPath resolver reads from `SrcUser` or `SrcBuiltin`).
- SPEC-V3R2-MIG-003 (config loader addition for 5 sections depends on the typed-Value pattern established here).

### 9.3 Related

- SPEC-V3R2-EXT-004 (migration runner applies schema migrations before `Load()`).
- SPEC-V3R2-ORC-002 (common protocol CI lint reuses the audit test pattern from this SPEC).
- SPEC-V3R2-WF-006 (output styles inherit provenance from the same resolver).
- SPEC-V3R2-CON-003 (consolidation pass moves settings-management rule text into `.claude/rules/moai/core/settings-management.md`).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §5.5 Multi-Layer Settings with Provenance.
- Principle: P6 (Permission Bubble — provenance prerequisite); P7 (Sandbox Default — routing by source); P12 (Constitutional Governance — settings as versioned artifacts).
- Pattern: X-2 (Multi-Layer Settings with Provenance Tags, priority 4); S-1 (Permission Stack, depends); T-5 (Hook Protocol — ConfigChange integration).
- Problem: P-C04 (no config provenance, HIGH); P-H06 (5 yaml sections without loader, CRITICAL — pattern established here); P-R02 secondary (constitutional sprawl — single authoritative resolver).
- Master Appendix A: Principle P6 → secondary SPEC-V3R2-RT-005; P11 → secondary.
- Master Appendix C: Pattern X-2 → primary SPEC-V3R2-RT-005 (priority 4); S-1 → this SPEC composes with it.
- Wave 1 sources: r3-cc-architecture-reread.md §1.3 (settings precedence), §2 Decision 11 (provenance on every element), §4 Adopt 1 (multi-layer with provenance).
- Wave 2 sources: design-principles.md P6 rationale; pattern-library.md X-2 (priority 4); problem-catalog.md Cluster 4 P-C04/P-H06.
- BC-ID: BC-V3R2-015 (multi-layer settings resolution replaces flat merge; AUTO migration, reader layer only, no user API change).
- Priority: P0 Critical — prerequisite for RT-002 (permission), RT-003 (sandbox), RT-006 (hook ConfigChange), MIG-003 (5 loader additions).
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1046` (§11.3 RT-005 definition)
  - `docs/design/major-v3-master.md:L974` (§8 BC-V3R2-015 — multi-layer settings)
  - `docs/design/major-v3-master.md:L988` (§9 Phase 1 Constitution & Foundation — reconciled)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 4 (P-C04, P-H06)
