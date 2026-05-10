# SPEC-V3R2-SPC-002 Task List

> Implementation task list for **@MX TAG v2 with hook JSON integration and sidecar index**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial task list. 22 tasks (T-SPC002-01..22) grouped into 6 milestones (M1..M6), priorities P0/P1/P2. Wave-split: tasks fit within ~22 entries — no wave-split needed (under feedback_large_spec_wave_split.md 30-task threshold). |

---

## 1. Task Overview

| Milestone | Phase | Tasks | Priority | REQ Coverage |
|---|---|---|---|---|
| M1: PostToolUse handler + types | RED → GREEN | T-SPC002-01..03, T-SPC002-15 | P0 | REQ-001, REQ-002, REQ-005, REQ-010, REQ-011, REQ-012 |
| M2: `/moai mx` flag dispatcher | RED → GREEN | T-SPC002-04..06 | P0 | REQ-013, REQ-032, REQ-042 |
| M3: silent env + mx.yaml ignore | RED → GREEN | T-SPC002-07, T-SPC002-08 | P1 | REQ-030, REQ-031 |
| M4: Scanner correctness fixtures | RED → GREEN | T-SPC002-09..11, T-SPC002-16 | P1 | REQ-006, REQ-007, REQ-021, REQ-022, REQ-040, REQ-041 |
| M5: Archive sweep + atomic verify | RED → GREEN | T-SPC002-12..14 | P1 | REQ-004, REQ-014, REQ-020 |
| M6: Verification + audit | VERIFY | T-SPC002-17..22 | P0 | (cross-cutting) |

Total: **22 tasks**. Below 30-task threshold → no wave-split required (per `feedback_large_spec_wave_split.md`).

---

## 2. Tasks by Milestone

### Milestone 1: PostToolUse handler + HookSpecificOutput.MxTags field — Priority P0

#### T-SPC002-01: Create internal/hook/post_tool_mx.go

**REQ traceback**: REQ-SPC-002-010, REQ-SPC-002-011, REQ-SPC-002-012

**AC traceback**: AC-04

**Goal**: Create new PostToolUse handler under `internal/hook/post_tool_mx.go` that:
1. Filters for `tool_name in {Write, Edit}`.
2. Filters for supported file extensions (via `mx.GetCommentPrefix(ext) != ""`).
3. Calls `mx.NewScanner().ScanFile(filePath)` then `Manager.UpdateFile(filePath, tags)`.
4. Returns HookOutput with `HookSpecificOutput.HookEventName = "PostToolUse"`, `AdditionalContext = formatTagsForContext(tags)`, `MxTags = tags`.
5. Handles silent mode via `os.Getenv("MOAI_MX_HOOK_SILENT")` (T-SPC002-07 dependency).

**Files**:
- `internal/hook/post_tool_mx.go` (new ~150 LOC)
- `internal/hook/post_tool_mx_test.go` (new ~250 LOC)

**Sub-tasks**:
- Define `postToolMxHandler` struct + `NewPostToolMxHandler()` constructor
- Implement `EventType()` returning `EventPostToolUse`
- Implement `Handle(ctx, input)` per logic above
- Add helper `formatTagsForContext(tags) string` (one tag per line: `@MX:KIND at file.go:N: body`)
- Add helper `extractFilePath(rawInput)` to parse tool_input JSON for file_path
- Register handler in `internal/hook/registry.go` (or appropriate dispatch site)
- Test: `TestPostToolUseHandler_EmitsMxTags` (Python file with `# @MX:WARN`)
- Test: `TestPostToolUseHandler_NonEditTool` (tool_name = "Read" → empty output)
- Test: `TestPostToolUseHandler_UnsupportedExt` (.txt → empty output)
- Test: `TestPostToolUseHandler_HandlerEventTypeIsPostToolUse`

**Acceptance**:
- [ ] Handler returns valid HookOutput for Edit/Write events
- [ ] Non-Write/Edit tools produce empty output (graceful no-op)
- [ ] Unsupported extensions produce empty output
- [ ] All 4 sub-tests PASS

**Dependencies**: T-SPC002-02 (MxTags field must exist)

---

#### T-SPC002-02: Add HookSpecificOutput.MxTags field

**REQ traceback**: REQ-SPC-002-011

**AC traceback**: AC-04

**Goal**: Add `MxTags []mx.Tag` field to `HookSpecificOutput` struct in `internal/hook/types.go`.

**Files**:
- `internal/hook/types.go` (+5 LOC)

**Sub-tasks**:
- Import `internal/mx` package (verify no cyclic import)
- Add field `MxTags []mx.Tag \`json:"mxTags,omitempty"\``
- Verify existing JSON marshalling tests in `protocol_test.go` still PASS (backward compatibility)
- Add test `TestHookSpecificOutput_MxTagsField_JSONMarshal` (verifies JSON output has `mxTags` key when populated, absent when empty)

**Acceptance**:
- [ ] Field added with omitempty tag
- [ ] No cyclic import errors (`go vet`)
- [ ] Existing tests still PASS
- [ ] JSON marshalling test PASS

**Dependencies**: None (independent struct edit)

---

#### T-SPC002-03: AdditionalContext format + helper

**REQ traceback**: REQ-SPC-002-011 (human-readable summary), REQ-SPC-002-031 (silent mode skip)

**AC traceback**: AC-04, AC-11

**Goal**: Implement `formatTagsForContext(tags []mx.Tag) string` in `post_tool_mx.go`. Format: one tag per line.

Format spec:
```
@MX:WARN at foo.go:42: missing timeout on requests.get
@MX:NOTE at foo.go:17: explains why handler forks
```

Cap at `mx.yaml` `hook.max_additional_context_bytes` (default 4096) — truncate with "... (N more tags)" suffix.

**Files**:
- `internal/hook/post_tool_mx.go` (function added)

**Sub-tasks**:
- Implement formatTagsForContext
- Apply byte cap from config
- Test: `TestFormatTagsForContext_BasicFormat`
- Test: `TestFormatTagsForContext_BudgetCap`
- Test: `TestFormatTagsForContext_EmptyTagsReturnsEmpty`

**Acceptance**:
- [ ] Format matches spec example
- [ ] Byte cap enforced
- [ ] 3 sub-tests PASS

**Dependencies**: T-SPC002-08 (mx.yaml config loader)

---

#### T-SPC002-15: 16-language scanner fixture

**REQ traceback**: REQ-SPC-002-005

**AC traceback**: AC-15

**Goal**: Add `TestScanner_AllSixteenLanguages` fixture in `internal/mx/scanner_test.go` that creates 16 files (one per language) and asserts 16 tags found.

**Files**:
- `internal/mx/scanner_test.go` (+~80 LOC)

**Sub-tasks**:
- Create test that synthesizes 16 source files in `t.TempDir()`
- Assert `scanner.ScanDirectory(tmpDir)` returns exactly 16 Tag records
- Verify each Tag has correct `Kind == MXNote` and File field

**Acceptance**:
- [ ] All 16 supported languages produce exactly 1 tag each
- [ ] Test PASS

**Dependencies**: None (existing scanner already supports 16 langs per research [E-26])

---

### Milestone 2: `/moai mx` flag dispatcher — Priority P0

#### T-SPC002-04: `--full` flag implementation

**REQ traceback**: REQ-SPC-002-013, REQ-SPC-002-003

**AC traceback**: AC-02

**Goal**: Add `--full` flag to `internal/cli/mx.go` that triggers full project rescan + sidecar rebuild + archive sweep.

**Files**:
- `internal/cli/mx.go` (parent command extension, +~50 LOC)
- `internal/cli/mx_test.go` (new ~80 LOC)

**Sub-tasks**:
- Add `BoolVar` for `--full` flag
- Implement `runMxFull(cmd, silent bool)` helper
- Walk project tree from `cwd`, scan all supported files via Scanner.ScanDirectory
- Apply mx.yaml ignore patterns
- Call `Manager.RebuildFromScan(allTags)` (new method or repurpose UpdateAll)
- Trigger archive sweep at end
- Test: `TestMxCmd_FullFlag_RebuildsSidecar` (42-tag fixture)
- Test: `TestMxCmd_FullFlag_RespectsIgnore`
- Test: `TestMxCmd_FullFlag_TriggersArchiveSweep` (works with T-SPC002-14)

**Acceptance**:
- [ ] `--full` produces sidecar with all current tags
- [ ] Console summary printed (count of tags, files scanned)
- [ ] Archive sweep triggered (stale entries moved)
- [ ] 3 sub-tests PASS

**Dependencies**: T-SPC002-08 (mx.yaml ignore wire-up), T-SPC002-14 (archive sweep)

---

#### T-SPC002-05: `--json` flag implementation

**REQ traceback**: REQ-SPC-002-032

**AC traceback**: AC-12

**Goal**: Add `--json` flag that prints current sidecar contents to stdout.

**Files**:
- `internal/cli/mx.go` (+~20 LOC)

**Sub-tasks**:
- Add `BoolVar` for `--json`
- Implement `runMxJsonDump(cmd)`: `Manager.Load()` → `json.MarshalIndent` → `os.Stdout.Write`
- Handle missing sidecar (output empty `{"schema_version":2,"tags":[]}` + repair suggestion to stderr)
- Test: `TestMxCmd_JsonFlag_DumpsSidecar`
- Test: `TestMxCmd_JsonFlag_MissingSidecarOk` (no sidecar → empty schema)

**Acceptance**:
- [ ] stdout receives valid JSON
- [ ] Schema version = 2
- [ ] Exit code 0
- [ ] 2 sub-tests PASS

**Dependencies**: None

---

#### T-SPC002-06: `--anchor-audit` flag scaffold + wire

**REQ traceback**: REQ-SPC-002-042

**AC traceback**: AC-14

**Goal**: Add `--anchor-audit` flag that lists MXAnchor tags with fan_in < 3.

**Files**:
- `internal/cli/mx.go` (+~30 LOC)
- `internal/mx/anchor_audit.go` (new ~80 LOC)
- `internal/mx/anchor_audit_test.go` (new ~80 LOC)

**Sub-tasks**:
- Add `BoolVar` for `--anchor-audit`
- Implement `runMxAnchorAudit(cmd)`:
  1. Load sidecar via `Manager.Load()`
  2. Filter to MXAnchor tags
  3. For each anchor: compute fan_in via SPC-004 `fanin.CountFanIn(anchorID, allTags)`
  4. Filter fan_in < 3
  5. Print markdown table to stdout: `| AnchorID | File | Line | FanIn |`
- Implement `RunAnchorAudit(mgr, threshold)` in `anchor_audit.go`
- Test: `TestRunAnchorAudit_LowFanInReported`
- Test: `TestRunAnchorAudit_HighFanInExcluded`
- Test: `TestMxCmd_AnchorAuditFlag_OutputFormat`

**Acceptance**:
- [ ] Audit lists only fan_in < 3 anchors
- [ ] Markdown table format
- [ ] Exit code 0
- [ ] 3 sub-tests PASS

**Dependencies**: T-SPC002-04 (mx command structure), SPC-004 fanin.go (already merged)

---

### Milestone 3: silent env + mx.yaml ignore — Priority P1

#### T-SPC002-07: MOAI_MX_HOOK_SILENT env handling

**REQ traceback**: REQ-SPC-002-031

**AC traceback**: AC-11

**Goal**: PostToolUse handler reads `MOAI_MX_HOOK_SILENT=1` and skips `additionalContext` (sidecar still updated, MxTags still emitted).

**Files**:
- `internal/hook/post_tool_mx.go` (env check in Handle method, +5 LOC)
- `internal/hook/post_tool_mx_test.go` (+~40 LOC)

**Sub-tasks**:
- Add `silent := os.Getenv("MOAI_MX_HOOK_SILENT") == "1"` check
- Conditionally skip formatTagsForContext call
- Test: `TestPostToolUseHandler_HookSilentEnv`
- Test: `TestPostToolUseHandler_HookSilentEnv_StillUpdatesSidecar`
- Test: `TestPostToolUseHandler_HookSilentEnv_StillEmitsMxTagsField`

**Acceptance**:
- [ ] AdditionalContext empty when silent=1
- [ ] Sidecar still updated
- [ ] MxTags structured field still populated
- [ ] 3 sub-tests PASS

**Dependencies**: T-SPC002-01

---

#### T-SPC002-08: mx.yaml config loader + ignore wire-up

**REQ traceback**: REQ-SPC-002-030

**AC traceback**: AC-10

**Goal**: Create `internal/mx/config.go` to load mx.yaml and provide ignore patterns. Wire into Scanner.

**Files**:
- `internal/mx/config.go` (new ~80 LOC)
- `internal/mx/config_test.go` (new ~60 LOC)

**Sub-tasks**:
- Define `Config` struct (`IgnorePatterns []string`, `HookMaxAdditionalContextBytes int`)
- Implement `LoadConfig(projectRoot string) (*Config, error)`:
  - Read `.moai/config/sections/mx.yaml`
  - Parse YAML, extract `ignore:` and `hook.max_additional_context_bytes`
  - Return defaults on missing file (graceful)
- Implement `defaultConfig() *Config`:
  - Default ignore: `["vendor/**", "node_modules/**", "dist/**", ".git/**", "**/*_generated.go", "**/mock_*.go"]`
  - Default budget: 4096 bytes
- Wire into Scanner via `scanner.SetIgnorePatterns(cfg.IgnorePatterns)` at all call sites (PostToolUse handler + `--full` CLI)
- Test: `TestLoadConfig_FromFile`
- Test: `TestLoadConfig_MissingFileReturnsDefault`
- Test: `TestLoadConfig_InvalidYamlReturnsDefault`
- Test: `TestScanner_RespectsMxYamlIgnore`

**Acceptance**:
- [ ] mx.yaml ignore patterns honored
- [ ] Default fallback works
- [ ] Invalid YAML graceful
- [ ] 4 sub-tests PASS

**Dependencies**: None

---

### Milestone 4: Scanner correctness fixtures — Priority P1

#### T-SPC002-09: MissingReasonForWarn 3-line lookahead fixture

**REQ traceback**: REQ-SPC-002-006, REQ-SPC-002-040

**AC traceback**: AC-05

**Goal**: Add explicit fixture verifying Scanner emits `MissingReasonForWarn` warning when `@MX:WARN` lacks REASON within 3 lines.

**Files**:
- `internal/mx/scanner_test.go` (+~50 LOC)
- (potentially) `internal/mx/scanner.go` (logic refinement if absent, +~10 LOC)

**Sub-tasks**:
- Verify Scanner already implements 3-line lookahead (research [E-09]); add logic if missing
- Test: `TestScanner_MissingReasonForWarn_3LineLookahead` (REASON missing → warning emitted)
- Test: `TestScanner_WarnWithReasonAtLine2_NoWarning` (REASON 2 lines after WARN → no warning)
- Test: `TestScanner_WarnWithReasonAtLine4_WarningEmitted` (REASON 4 lines after WARN → warning emitted, lookahead is 3)

**Acceptance**:
- [ ] Warning emission verified for 0/4+ line distance
- [ ] No warning for 1-3 line distance
- [ ] 3 sub-tests PASS

**Dependencies**: None (logic already exists per research)

---

#### T-SPC002-10: DuplicateAnchorID refuse-write fixture

**REQ traceback**: REQ-SPC-002-007, REQ-SPC-002-021

**AC traceback**: AC-06

**Goal**: Verify Scanner detects duplicate AnchorID and Manager refuses sidecar write.

**Files**:
- `internal/mx/scanner_test.go` (+~50 LOC)

**Sub-tasks**:
- Test: `TestScanner_DuplicateAnchorID_NamesBothFiles` (two files with same AnchorID → error contains both)
- Test: `TestScanner_DuplicateAnchorID_RefuseWrite` (Manager.Write rejected when scanner has duplicates)

**Acceptance**:
- [ ] Duplicate AnchorID detected
- [ ] Both file:line pairs in error message
- [ ] Manager.Write refused
- [ ] 2 sub-tests PASS

**Dependencies**: None

---

#### T-SPC002-11: Corrupt sidecar repair suggestion fixture

**REQ traceback**: REQ-SPC-002-022

**AC traceback**: AC-09

**Goal**: Verify Manager.Load handles corrupt JSON gracefully + emits repair suggestion.

**Files**:
- `internal/mx/sidecar_test.go` (+~40 LOC)

**Sub-tasks**:
- Test: `TestSidecar_CorruptJSON_ReturnsEmpty` (corrupt file → empty Sidecar)
- Test: `TestSidecar_CorruptJSON_StderrSuggestion` (capture stderr, assert "WARNING: sidecar corrupt")

**Acceptance**:
- [ ] Empty sidecar returned (no crash)
- [ ] Stderr contains repair suggestion
- [ ] 2 sub-tests PASS

**Dependencies**: None

---

#### T-SPC002-16: HookSpecificOutput mismatch validation

**REQ traceback**: REQ-SPC-002-041

**AC traceback**: AC-13

**Goal**: Verify validator rejects mxTags with wrong hookEventName.

**Files**:
- `internal/hook/types.go` (+15 LOC for ValidateMxTagsConsistency function)
- `internal/hook/types_test.go` (+~30 LOC)

**Sub-tasks**:
- Add `ValidateMxTagsConsistency(output *HookOutput) error`:
  - If `MxTags != nil && HookEventName != "PostToolUse"` → return ErrHookSpecificOutputMismatch
- Define `ErrHookSpecificOutputMismatch` sentinel error
- Test: `TestHookOutput_MxTagsWithPreToolUse_Rejected`
- Test: `TestHookOutput_MxTagsWithPostToolUse_Accepted`
- Test: `TestHookOutput_NoMxTags_AnyEventOk`

**Acceptance**:
- [ ] Validator rejects mismatched events
- [ ] Validator accepts correct PostToolUse + MxTags combo
- [ ] 3 sub-tests PASS

**Dependencies**: T-SPC002-02 (MxTags field exists)

---

### Milestone 5: Archive sweep + atomic write verification — Priority P1

#### T-SPC002-12: Atomic write no-partial-reads fixture

**REQ traceback**: REQ-SPC-002-004

**AC traceback**: AC-03

**Goal**: Stress-test concurrent reads during writes, assert no partial reads.

**Files**:
- `internal/mx/sidecar_test.go` (+~70 LOC)

**Sub-tasks**:
- Generate 5MB sidecar payload (10K tags)
- Test: `TestSidecar_AtomicWrite_NoPartialReads` (writer + 1000 readers, 0 partial reads)
- Run with `-race` flag

**Acceptance**:
- [ ] 0 partial reads
- [ ] No data race
- [ ] PASS under `-race -count=10`

**Dependencies**: None

---

#### T-SPC002-13: 7-day stale tag preservation fixture

**REQ traceback**: REQ-SPC-002-014

**AC traceback**: AC-07

**Goal**: Verify stale tag (LastSeenAt = 6 days ago, missing from current scan) is preserved in sidecar.

**Files**:
- `internal/mx/sidecar_test.go` (+~50 LOC)

**Sub-tasks**:
- Test: `TestSidecar_StaleNotYetArchived_7Days`
- Verify LastSeenAt timestamp not mutated

**Acceptance**:
- [ ] Tag remains in sidecar
- [ ] LastSeenAt unchanged
- [ ] PASS

**Dependencies**: None

---

#### T-SPC002-14: 8-day stale archive sweep fixture

**REQ traceback**: REQ-SPC-002-020

**AC traceback**: AC-08

**Goal**: Verify 8-day stale tag is moved to mx-archive.json on full scan.

**Files**:
- `internal/mx/sidecar_test.go` (+~60 LOC)
- `internal/mx/sidecar.go` (archive sweep helper, +~30 LOC)

**Sub-tasks**:
- Implement `archiveSweep()` helper in sidecar.go (called from RebuildFromScan path)
- Logic: for each tag in current sidecar, if `IsStale()` AND not in current scan → move to mx-archive.json
- Test: `TestSidecar_StaleArchived_8Days`
- Test: `TestSidecar_ArchiveSweep_TwoPhase` (verify archive write succeeds before sidecar rewrite — partial-failure idempotency)

**Acceptance**:
- [ ] Stale tag removed from sidecar
- [ ] Stale tag appended to archive
- [ ] Two-phase write (archive first, sidecar second) verified
- [ ] 2 sub-tests PASS

**Dependencies**: T-SPC002-04 (--full path triggers sweep)

---

### Milestone 6: Verification + audit — Priority P0

#### T-SPC002-17: Full test suite

**REQ traceback**: (cross-cutting)

**Goal**: `go test -race -count=1 ./...` PASS.

**Acceptance**: 0 regressions; existing SPC-004 tests still PASS.

**Dependencies**: M1-M5 complete.

---

#### T-SPC002-18: Linter

**Goal**: `golangci-lint run` clean.

---

#### T-SPC002-19: Build verification

**Goal**: `make build` exits 0; `internal/template/embedded.go` regenerated correctly. `diff -r .claude/ internal/template/templates/.claude/` shows 0 changes (no template assets touched by this SPEC).

---

#### T-SPC002-20: CHANGELOG

**Goal**: Add 4 bullet entries to CHANGELOG.md Unreleased section per plan §M6-T20.

---

#### T-SPC002-21: @MX tags

**Goal**: Apply 6 @MX tags per plan §6 (1 ANCHOR + 2 WARN + 3 NOTE).

---

#### T-SPC002-22: End-to-end smoke test

**Goal**: Manual verification — open `claude` session, edit a Go file via Edit tool, confirm:
1. PostToolUse hook fires
2. stdout contains `mxTags` JSON field
3. `.moai/state/mx-index.json` is updated atomically
4. `MOAI_MX_HOOK_SILENT=1` correctly empties additionalContext

**Acceptance**: All 4 observations confirmed.

**Dependencies**: M1-M5 + T-SPC002-17 PASS.

---

## 3. Task Dependency Graph

```
T-02 (MxTags field)
  ↓
T-01 (post_tool_mx.go handler) ← T-08 (config loader)
  ↓                                ↓
T-03 (formatTagsForContext)    T-15 (16-lang fixture)
  ↓
T-07 (MOAI_MX_HOOK_SILENT)     ↓
                                T-04 (--full flag) ← T-14 (archive sweep)
                                  ↓
                                T-05 (--json flag)
                                T-06 (--anchor-audit)

T-09 (MissingReasonForWarn)
T-10 (DuplicateAnchorID)
T-11 (corrupt sidecar)
T-12 (atomic write race)
T-13 (7-day stale preservation)
T-16 (HookSpecificOutputMismatch) ← T-02

T-17 (full test) → T-18 (lint) → T-19 (build) → T-20 (CHANGELOG) → T-21 (MX tags) → T-22 (smoke test)
```

Parallelizable groups:
- M1 tasks (T-01..03, T-15): mostly independent after T-02
- M4 tasks (T-09..11, T-16): all independent (different test files)
- M5 tasks (T-12..14): independent

---

## 4. Dependency Resolution

### 4.1 SPC-004 status (already merged)

PR #746 (commit `68795dbe3`) merged 2026-04-30. `internal/mx/resolver_query.go`, `fanin.go`, `danger_category.go`, `spec_association.go` exist on main. Tasks T-SPC002-06 (anchor-audit) reuse `fanin.CountFanIn` from this code.

If somehow the SPC-004 code was reverted between plan-PR merge and run-PR start, T-SPC002-06 fallback: implement minimal CountFanIn locally in anchor_audit.go (count of MXAnchor tags referencing the same AnchorID — simple O(n²)). Plan binding: this fallback is documented but NOT expected since SPC-004 is in main HEAD `fcb486c87`.

### 4.2 Wave 3 PR #741 status

Commit `3f0933550` (2026-04-XX) merged tag.go, scanner.go, sidecar.go, comment_prefixes.go, resolver.go. These are the foundation of this SPEC and assumed present.

If somehow reverted: full re-implementation required (>3,000 LOC). Recovery path: file-level re-port from PR #741 diff + this SPEC's add-on.

### 4.3 SPEC-V3R2-RT-001 status

`internal/hook/types.go` already implements RT-001 protocol (research [E-14]). T-SPC002-02 extends `HookSpecificOutput` with one new omitempty field — backward-compatible.

### 4.4 SPEC-V3R2-WF-005 (16-language enum)

Out-of-scope but advisory. `internal/mx/comment_prefixes.go` and `internal/hook/file_changed.go:supportedExtensions` have minor drift (research [E-27]). T-SPC002-15 16-lang fixture catches future divergence; integration into a single SoT is deferred to WF-005.

---

## 5. Effort Estimates (priority labels only, no time)

Per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule, all tasks use priority labels; no time estimates.

| Task category | Priority | Notes |
|---|---|---|
| M1, M2, M6 (core path) | P0 | Critical — handler + CLI + verification |
| M3, M4 (correctness fixtures) | P1 | Required but defer-acceptable |
| M5 (archive + race) | P1 | Required for spec §7 budget assertion |
| (none) | P2 | No P2 tasks in this SPEC |

---

## 6. Test Coverage Goal

Per `.moai/config/sections/quality.yaml` `coverage_target: 85`:

| Package | Pre-SPEC LOC | Post-SPEC LOC | Coverage target |
|---|---|---|---|
| `internal/hook/` | ~9000 | ~9400 | ≥ 85% (mostly new handler test coverage) |
| `internal/mx/` | ~3000 | ~3300 | ≥ 85% (new config + anchor_audit) |
| `internal/cli/` | ~very large | +~150 | ≥ 85% on mx.go new flag dispatcher |

---

## 7. Risk-based Task Prioritization

| Risk | Mitigation task |
|---|---|
| PostToolUse handler crashes Claude session | T-SPC002-01 graceful no-op + T-SPC002-17 race detector |
| Sidecar drift from source truth | T-SPC002-04 (--full rebuild) + T-SPC002-22 smoke test |
| Token budget overflow | T-SPC002-03 (formatTagsForContext budget cap) + T-SPC002-07 (silent env) |
| Atomic write race | T-SPC002-12 (concurrent fixture, -race -count=10) |
| Stale tag never archived | T-SPC002-14 (8-day fixture) |
| Cross-language fixture drift | T-SPC002-15 (16-lang fixture as regression guard) |
| HookSpecificOutput mismatch | T-SPC002-16 (validator + sentinel error) |

---

## 8. Wave-Split Assessment

Per `feedback_large_spec_wave_split.md`: SPECs with 30+ tasks should be wave-split.

This SPEC has **22 tasks**. Below threshold → no wave-split needed. Single delegation prompt per task during Run phase is acceptable (~1.5KB each).

---

End of tasks.

Version: 0.1.0
Status: Tasks artifact for SPEC-V3R2-SPC-002
