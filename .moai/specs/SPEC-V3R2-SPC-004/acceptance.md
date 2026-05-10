# SPEC-V3R2-SPC-004 Acceptance Criteria

> Detailed Given-When-Then scenarios and verification evidence for **@MX anchor resolver (query by SPEC ID, fan_in, danger category)**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                                         | Description |
|---------|------------|------------------------------------------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B)          | Initial acceptance.md. 15 AC scenarios mapped 1:1 to spec.md §6 AC-SPC-004-01..15. Each AC includes Given/When/Then + measurable evidence + REQ traceback + tasks.md task IDs. Performance budgets: <100ms 1K tags / <2s 50 anchors with LSP. Note: 15 ACs are **largely covered by existing tests** in `internal/mx/resolver_query_test.go` and `internal/cli/mx_query_test.go` (PR #746 commit `68795dbe3`); this artifact preserves their G/W/T form and adds the gap-closure (G-01..G-08) verification fixtures. |

---

## 1. AC Coverage Map

| AC ID | Spec §6 statement | Mapped REQs | Mapped tasks | Existing test? | Evidence type |
|---|---|---|---|---|---|
| AC-SPC-004-01 | `--spec X --kind anchor` returns only ANCHOR for SPEC X | REQ-001, REQ-006, REQ-010 | T-SPC004-04, T-SPC004-05, T-SPC004-12 | YES (`TestResolver_AC1_SpecAndKindFilter`) | Resolver Go API + CLI assertion |
| AC-SPC-004-02 | `--fan-in-min 3 --kind anchor` returns 3-tier subset | REQ-011 | T-SPC004-01, T-SPC004-02 | YES (`TestResolver_AC2_FanInMinFilter`) | LSP + textual counter |
| AC-SPC-004-03 | `--danger concurrency` matches WARN with REASON pattern | REQ-012 | T-SPC004-03 | YES (`TestResolver_AC3_DangerCategoryFilter`) | DangerCategoryMatcher |
| AC-SPC-004-04 | sidecar absent → SidecarUnavailable + suggest `/moai mx --full` | REQ-013 | T-SPC004-09 | PARTIAL (covered in CLI test; stderr format fixture added) | stderr capture |
| AC-SPC-004-05 | `--format json` (default) → valid JSON array per REQ-005 schema | REQ-004, REQ-005 | T-SPC004-15 | YES (`TestResolver_AC5_JSONOutputSchema`) | JSON schema validation |
| AC-SPC-004-06 | `--format table` → human-readable columnar | REQ-004 | (existing) | YES (`TestMxQueryCmd_AC6_TableFormat`) | stdout column check |
| AC-SPC-004-07 | no LSP for Python → results with `fan_in_method: "textual"` annotation | REQ-020 | T-SPC004-01, T-SPC004-02 | YES (`TestResolver_AC7_TextualFallbackAnnotation`) | TagResult.FanInMethod |
| AC-SPC-004-08 | 10K tags + no `--limit` → at most 100 entries + TruncationNotice | REQ-007, REQ-021 | T-SPC004-13 | YES (`TestResolver_AC8_PaginationAndTruncation`) | result count + TruncationNotice |
| AC-SPC-004-09 | `MOAI_MX_QUERY_STRICT=1` + no LSP → exit non-zero + LSPRequired | REQ-030 | T-SPC004-02 | PARTIAL (sentinel error in tests; LSP-detect path 신규) | exit code + error type |
| AC-SPC-004-10 | `--format markdown` → markdown table | REQ-031 | (existing) | YES (`TestMxQueryCmd_AC10_MarkdownFormat`) | stdout markdown structure |
| AC-SPC-004-11 | ANCHOR in test fixture → references from other test files excluded | REQ-040 | T-SPC004-07, T-SPC004-08 | YES (`TestResolver_AC11_TestFileExclusion`) | fan_in count diff |
| AC-SPC-004-12 | zero match → `[]` exit 0 | REQ-041 | (existing) | YES (`TestResolver_AC12_EmptyResult`) | stdout + exit code |
| AC-SPC-004-13 | invalid `--kind nonexistent` → exit 2 + InvalidQuery | REQ-041 | T-SPC004-06 | YES (`TestResolver_AC13_InvalidQuery`) | exit code + stderr |
| AC-SPC-004-14 | combined `--spec X --kind anchor --fan-in-min 3` AND-composed | REQ-042 | (existing) | YES (`TestResolver_AC14_CombinedFilters`) | filter composition |
| AC-SPC-004-15 | tag body `ANCHOR for SPEC-AUTH-001` → `spec_associations` includes "SPEC-AUTH-001" | REQ-006 | T-SPC004-04, T-SPC004-05 | YES (`TestResolver_AC15_BodyBasedSpecAssociation`) | SpecAssociator regex |

→ All 15 ACs traceable. 11 of 15 already covered by PR #746 existing tests (research §2.2 + grep evidence); 4 are PARTIAL/NEW (AC-04 stderr fixture, AC-09 LSP-detect, AC-11 test_paths glob, AC-13 danger 분기 추가).

---

## 2. Acceptance Criteria

### AC-SPC-004-01: `--spec X --kind anchor` returns only ANCHOR tags for SPEC X

**REQ traceback**: REQ-SPC-004-001 (CLI subcommand), REQ-SPC-004-006 (SPEC association), REQ-SPC-004-010 (Event-driven combined filter)

**Mapped tasks**: T-SPC004-04 (LoadSpecModules), T-SPC004-05 (SpecAssociator wire), T-SPC004-12 (CLI wire)

**Given** a sidecar at `.moai/state/mx-index.json` containing 20 tags distributed across 2 SPECs: 10 tags associated with `SPEC-X-001` (5 ANCHOR + 5 NOTE) and 10 tags with `SPEC-Y-002` (5 ANCHOR + 5 WARN)
**And** SPEC-X-001's `module:` frontmatter lists `internal/x/`
**And** SPEC-Y-002's `module:` frontmatter lists `internal/y/`
**And** all SPEC-X tags' file paths fall under `internal/x/`; all SPEC-Y tags under `internal/y/`
**When** an operator runs `moai mx query --spec SPEC-X-001 --kind anchor --format json`
**Then** stdout MUST contain a JSON array with exactly 5 entries
**And** each entry MUST have `kind == "ANCHOR"`
**And** each entry's `spec_associations` MUST contain `"SPEC-X-001"`
**And** no entry MUST have `spec_associations` containing `"SPEC-Y-002"`
**And** exit status MUST be 0

**Verification command**:
```bash
cd /tmp/proj
moai mx query --spec SPEC-X-001 --kind anchor --format json | jq 'length'
# Expected: 5
moai mx query --spec SPEC-X-001 --kind anchor --format json | jq '.[].kind' | sort -u
# Expected: "ANCHOR"
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC1_SpecAndKindFilter` (already exists per `grep "AC-SPC-004-01"` at line 54). Fixture builds 20-tag sidecar under `t.TempDir()`, instantiates `SpecAssociator` with module map `{"SPEC-X-001": ["internal/x/"], "SPEC-Y-002": ["internal/y/"]}`, asserts result.

**Expected**: PASS (existing fixture covers Resolver Go API; T-SPC004-04 + T-SPC004-05 add CLI wire-up so end-to-end CLI test PASS).

---

### AC-SPC-004-02: `--fan-in-min 3 --kind anchor` returns subset

**REQ traceback**: REQ-SPC-004-011 (Event-driven fan-in filter)

**Mapped tasks**: T-SPC004-01 (LSPFanInCounter struct), T-SPC004-02 (LSPFanInCounter tests)

**Given** a sidecar with 5 ANCHOR tags whose computed fan_in values are 1, 2, 3, 5, 10 (computed via either LSP or textual counter)
**When** an operator runs `moai mx query --fan-in-min 3 --kind anchor`
**Then** the result MUST contain exactly 3 entries (fan_in 3, 5, 10)
**And** the entries MUST be ordered by file path (consistent ordering)
**And** each entry's `fan_in` field MUST be ≥ 3
**And** each entry's `fan_in_method` MUST be either `"lsp"` or `"textual"`

**Verification command**:
```bash
moai mx query --fan-in-min 3 --kind anchor --format json | jq 'length'
# Expected: 3
moai mx query --fan-in-min 3 --kind anchor --format json | jq '.[].fan_in' | sort -n
# Expected: 3 5 10
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC2_FanInMinFilter` (already exists per `grep "AC-SPC-004-02"` at line 106). T-SPC004-01 신규: `internal/mx/fanin_lsp_test.go` adds LSP mock for the same fan_in values.

**Expected**: PASS.

---

### AC-SPC-004-03: `--danger concurrency` matches WARN with REASON pattern

**REQ traceback**: REQ-SPC-004-012 (danger category)

**Mapped tasks**: T-SPC004-03 (LoadDangerConfig + user wire)

**Given** the sidecar contains a WARN tag with `Reason == "goroutine leak on panic"`
**And** `mx.yaml` `danger_categories:` either is unset (DefaultDangerCategories used) OR maps `"goroutine leak"` → `concurrency`
**When** an operator runs `moai mx query --danger concurrency`
**Then** the result MUST include that WARN tag
**And** the entry's `danger_category` MUST be `"concurrency"`

**Verification command**:
```bash
moai mx query --danger concurrency --format json | jq '.[] | select(.danger_category == "concurrency")' | jq -s 'length'
# Expected: ≥ 1
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC3_DangerCategoryFilter` (already exists at line 145). T-SPC004-03 신규: `TestLoadDangerConfig_UserCustomCategories` verifies user-defined patterns.

**Expected**: PASS.

---

### AC-SPC-004-04: sidecar absent → SidecarUnavailable + suggest `/moai mx --full`

**REQ traceback**: REQ-SPC-004-013 (Event-driven sidecar absent)

**Mapped tasks**: T-SPC004-09 (stderr format fixture)

**Given** `.moai/state/mx-index.json` does not exist (project freshly initialized or sidecar deleted)
**When** an operator runs `moai mx query --kind anchor`
**Then** the command MUST exit with non-zero status (current implementation: exit 1 from cobra RunE error)
**And** stderr MUST contain the substring `SidecarUnavailable`
**And** stderr MUST contain the substring `/moai mx --full`

**Verification command**:
```bash
rm -f /tmp/proj/.moai/state/mx-index.json
moai mx query --kind anchor 2>&1 1>/dev/null | grep -E "SidecarUnavailable.*moai mx --full"
# Expected: 1 match line
```

**Test fixture**: `internal/cli/mx_query_test.go` `TestMxQueryCmd_AC4_SidecarUnavailable` (already exists at line 89). T-SPC004-09 신규: explicit `TestSidecarUnavailable_StderrFormat` asserts both substrings present in single stderr capture.

**Expected**: PASS.

---

### AC-SPC-004-05: `--format json` (default) → valid JSON array per REQ-005 schema

**REQ traceback**: REQ-SPC-004-004 (default format), REQ-SPC-004-005 (JSON schema 11 fields)

**Mapped tasks**: T-SPC004-15 (16-language sweep verifies schema)

**Given** the sidecar contains valid tags
**When** an operator runs `moai mx query` (no `--format` flag, defaults to JSON)
**Then** stdout MUST be a valid JSON document (parseable by `jq`)
**And** the document MUST be a JSON array (top-level `[`...`]`)
**And** each array entry MUST contain at minimum the fields: `kind`, `file`, `line`, `body`, `created_by`, `last_seen_at`, `spec_associations`
**And** ANCHOR entries MAY have `fan_in` and `fan_in_method`
**And** WARN entries MAY have `danger_category`
**And** all entries with `reason` populated MUST surface it; absent → field absent (omitempty)

**Verification command**:
```bash
moai mx query | jq 'type'
# Expected: "array"
moai mx query | jq '.[0] | keys' | jq 'sort'
# Expected: array containing "body", "created_by", "file", "kind", "last_seen_at", "line", "spec_associations"
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC5_JSONOutputSchema` (already exists at line 212).

**Expected**: PASS.

---

### AC-SPC-004-06: `--format table` → human-readable columnar

**REQ traceback**: REQ-SPC-004-004 (table format)

**Mapped tasks**: (existing — covered by `TestResolver_AC6_TableFormat`)

**Given** the sidecar contains tags
**When** an operator runs `moai mx query --format table`
**Then** stdout MUST contain a header line with columns: `KIND`, `FILE`, `LINE`, `BODY`
**And** stdout MUST contain a separator line (dashes)
**And** stdout MUST contain at least one data row
**And** each row MUST be space-padded to align columns

**Verification command**:
```bash
moai mx query --format table | head -3
# Expected first line containing "KIND" and "FILE"
moai mx query --format table | grep -c "^-"
# Expected: ≥ 1 (separator line)
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC6_TableFormat` (already exists at line 155 + line 507 secondary).

**Expected**: PASS.

---

### AC-SPC-004-07: no LSP for Python → `fan_in_method: "textual"` annotation

**REQ traceback**: REQ-SPC-004-020 (State-driven LSP unavailable)

**Mapped tasks**: T-SPC004-01 (LSPFanInCounter), T-SPC004-02 (LSPFanInCounter tests with no-LSP path)

**Given** a Python project with `@MX:ANCHOR py-anchor-001` in a `.py` file
**And** no Python LSP server (pylsp / pyright) is running on the host
**And** `MOAI_MX_QUERY_STRICT` is unset (or `"0"`)
**When** an operator runs `moai mx query --fan-in-min 1 --file-prefix internal/py/ --format json`
**Then** the result MUST contain ANCHOR entries from `internal/py/`
**And** each entry's `fan_in_method` MUST be `"textual"` (not `"lsp"`)
**And** exit status MUST be 0 (graceful fallback)

**Verification command**:
```bash
moai mx query --fan-in-min 1 --file-prefix internal/py/ --format json | jq '.[].fan_in_method' | sort -u
# Expected: "textual"
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC7_TextualFallbackAnnotation` (already exists at line 285). T-SPC004-01 + T-SPC004-02 신규: `TestLSPFanInCounter_NoSymbolMatch` verifies fallback behavior with mock client returning empty workspace/symbol response.

**Expected**: PASS.

---

### AC-SPC-004-08: 10K tags + no `--limit` → at most 100 entries + TruncationNotice

**REQ traceback**: REQ-SPC-004-007 (default limit 100), REQ-SPC-004-021 (TruncationNotice header)

**Mapped tasks**: T-SPC004-13 (benchmark verifies pagination)

**Given** the sidecar contains exactly 10,000 tags (any kinds)
**When** an operator runs `moai mx query` (no `--limit` flag)
**Then** stdout MUST contain a JSON array with exactly 100 entries (default limit)
**And** stderr MUST contain the substring `TruncationNotice`
**And** stderr MUST report the total count (10000) and truncated count (100)
**And** exit status MUST be 0

**Verification command**:
```bash
# Generate 10K-tag sidecar via test helper
moai mx query | jq 'length'
# Expected: 100
moai mx query 2>&1 1>/dev/null | grep -c "TruncationNotice"
# Expected: 1
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC8_PaginationAndTruncation` (already exists at line 254 + line 643 secondary). T-SPC004-13 신규: benchmark `BenchmarkResolver_Resolve_1KTags` measures perf at limit boundary.

**Expected**: PASS.

---

### AC-SPC-004-09: `MOAI_MX_QUERY_STRICT=1` + no LSP → exit non-zero + LSPRequired

**REQ traceback**: REQ-SPC-004-030 (Optional strict mode)

**Mapped tasks**: T-SPC004-02 (LSP-detect path)

**Given** `MOAI_MX_QUERY_STRICT=1` env var is set
**And** no LSP server is running for the target language(s)
**When** an operator runs `moai mx query --fan-in-min 3 --kind anchor`
**Then** the command MUST exit with non-zero status
**And** stderr MUST contain the substring `LSPRequired` (case-sensitive)
**And** the error MUST mention the language (or `"any"` if multi-language project)

**Verification command**:
```bash
MOAI_MX_QUERY_STRICT=1 moai mx query --fan-in-min 3 --kind anchor; echo $?
# Expected: non-zero exit code
MOAI_MX_QUERY_STRICT=1 moai mx query --fan-in-min 3 --kind anchor 2>&1 | grep -c LSPRequired
# Expected: ≥ 1
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC9_StrictModeNoLSP` (already exists at line 314). T-SPC004-02 신규: `TestLSPFanInCounter_StrictModeUnavailable` adds explicit availability-check fixture (replaces current unconditional `LSPRequired` return).

**Expected**: PASS.

---

### AC-SPC-004-10: `--format markdown` → markdown table

**REQ traceback**: REQ-SPC-004-031 (Optional markdown format)

**Mapped tasks**: (existing — covered by `TestResolver_AC10_MarkdownFormat`)

**Given** the sidecar contains tags
**When** an operator runs `moai mx query --format markdown`
**Then** stdout MUST contain a markdown table header line `| Kind | File | Line | Body | FanIn | Danger | SPECs |`
**And** stdout MUST contain a markdown separator line `|------|...`
**And** stdout MUST contain ≥1 data row formatted as `| ... |`
**And** if TruncationNotice applies, stdout MUST start with a blockquote `> **TruncationNotice**: showing N of M total results.`

**Verification command**:
```bash
moai mx query --format markdown | grep -E "^\| Kind \| File "
# Expected: 1 match
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC10_MarkdownFormat` (already exists at line 189 + line 485 secondary).

**Expected**: PASS.

---

### AC-SPC-004-11: ANCHOR in test fixture → references from other test files excluded

**REQ traceback**: REQ-SPC-004-040 (Complex test exclusion)

**Mapped tasks**: T-SPC004-07 (isTestFile glob), T-SPC004-08 (test_paths user wire)

**Given** an ANCHOR `auth-handler-anchor-001` in `internal/auth/handler.go` line 42
**And** the same anchor_id appears in `internal/auth/handler_test.go` (3 references) and `tests/integration/auth_test.go` (2 references) and `internal/auth/other.go` (1 reference)
**And** `mx.yaml` either has default `test_paths` OR sets `test_paths: ["**/integration/**"]`
**When** an operator runs `moai mx query --spec ... --kind anchor --fan-in-min 1` (no `--include-tests`)
**Then** the result for `auth-handler-anchor-001` MUST have `fan_in == 1` (only `other.go` reference; both test files excluded)
**When** the same operator runs with `--include-tests`
**Then** the result MUST have `fan_in == 6` (all references included)

**Verification command**:
```bash
moai mx query --kind anchor --fan-in-min 1 --format json | jq '.[] | select(.anchor_id == "auth-handler-anchor-001") | .fan_in'
# Expected: 1
moai mx query --kind anchor --fan-in-min 1 --include-tests --format json | jq '.[] | select(.anchor_id == "auth-handler-anchor-001") | .fan_in'
# Expected: 6
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC11_TestFileExclusion` (already exists at line 342). T-SPC004-07 + T-SPC004-08 신규: `TestIsTestFile_UserPattern_IntegrationDir` + `TestTextualFanInCounter_RespectsUserTestPaths` verify user glob wire-up.

**Expected**: PASS.

---

### AC-SPC-004-12: zero match → `[]` exit 0

**REQ traceback**: REQ-SPC-004-041 (Complex empty result)

**Mapped tasks**: (existing)

**Given** the sidecar exists but contains no ANCHOR tags
**When** an operator runs `moai mx query --kind anchor`
**Then** stdout MUST be the literal string `[]\n` (or `[]` followed by newline)
**And** exit status MUST be 0
**And** stderr MUST be empty (or only whitespace)

**Verification command**:
```bash
moai mx query --kind anchor | tr -d '\n'; echo
# Expected: []
moai mx query --kind anchor; echo "exit=$?"
# Expected: exit=0
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC12_EmptyResult` (already exists at line 370).

**Expected**: PASS.

---

### AC-SPC-004-13: invalid `--kind nonexistent` → exit 2 + InvalidQuery

**REQ traceback**: REQ-SPC-004-041 (Complex invalid filter)

**Mapped tasks**: T-SPC004-06 (validateQuery danger 분기 추가; danger 검증도 같은 경로)

**Given** the operator passes an invalid value to a typed flag (e.g. `--kind frobnicate` or, post-G-02, `--danger frobnicate`)
**When** an operator runs `moai mx query --kind frobnicate`
**Then** the command MUST exit with status 2 (per spec §5.5 REQ-041)
**And** stderr MUST contain the substring `InvalidQuery`
**And** stderr MUST mention the offending field name and value
**And** stderr MUST list the allowed values for that field

**Verification command**:
```bash
moai mx query --kind frobnicate; echo $?
# Expected exit: 2
moai mx query --kind frobnicate 2>&1 | grep -E "InvalidQuery.*kind.*frobnicate"
# Expected: 1 match
moai mx query --kind frobnicate 2>&1 | grep -E "allowed.*note.*warn.*anchor"
# Expected: 1 match
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC13_InvalidQuery_Kind` (already exists at line 674) + `internal/cli/mx_query_test.go` `TestMxQueryCmd_AC13_InvalidKind` (line 66). T-SPC004-06 신규: `TestValidateQuery_UnknownDanger_InvalidQueryError` extends to danger field.

**Note**: Current CLI implementation may exit with 1 (cobra RunE non-nil error). Spec §5.5 REQ-041 requires exit 2; T-SPC004-06 verifies this. Discrepancy: existing test verifies error type but not exit code 2 specifically — gap-closure verifies exit code.

**Expected**: PASS (after T-SPC004-06 ensures exit code 2 vs existing exit code 1).

---

### AC-SPC-004-14: combined filters AND-composed

**REQ traceback**: REQ-SPC-004-042 (Complex multi-filter AND)

**Mapped tasks**: (existing)

**Given** the sidecar contains tags such that:
- 8 ANCHOR tags total
- 5 of those are associated with `SPEC-X-001`
- 3 of those 5 have fan_in ≥ 3
**When** an operator runs `moai mx query --spec SPEC-X-001 --kind anchor --fan-in-min 3`
**Then** the result MUST contain exactly 3 entries
**And** every entry MUST satisfy ALL three filters:
- `kind == "ANCHOR"`
- `spec_associations` contains `"SPEC-X-001"`
- `fan_in >= 3`

**Verification command**:
```bash
moai mx query --spec SPEC-X-001 --kind anchor --fan-in-min 3 --format json | jq 'length'
# Expected: 3
moai mx query --spec SPEC-X-001 --kind anchor --fan-in-min 3 --format json | jq '.[] | (.kind == "ANCHOR" and (.spec_associations | index("SPEC-X-001")) and (.fan_in >= 3))' | sort -u
# Expected: true
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC14_CombinedFilters` (already exists at line 400).

**Expected**: PASS.

---

### AC-SPC-004-15: tag body `ANCHOR for SPEC-AUTH-001` → spec_associations includes "SPEC-AUTH-001"

**REQ traceback**: REQ-SPC-004-006 (b) (body-based SPEC association)

**Mapped tasks**: T-SPC004-04, T-SPC004-05 (LoadSpecModules + SpecAssociator wire); 16-language sweep T-SPC004-15

**Given** a tag with `Body == "ANCHOR for SPEC-AUTH-001 handler"` in any source file
**And** `SPEC-AUTH-001` may or may not have a `module:` frontmatter entry pointing to this file
**When** an operator runs `moai mx query --format json` (no SPEC filter)
**Then** the result entry corresponding to that tag MUST have `spec_associations` containing the string `"SPEC-AUTH-001"`

**Verification command**:
```bash
moai mx query --format json | jq '.[] | select(.body | contains("SPEC-AUTH-001")) | .spec_associations'
# Expected: array containing "SPEC-AUTH-001"
```

**Test fixture**: `internal/mx/resolver_query_test.go` `TestResolver_AC15_BodyBasedSpecAssociation` (already exists at line 449). 16-language sweep T-SPC004-15 verifies body-based association works in all 16 supported languages.

**Expected**: PASS.

---

## 3. Performance Acceptance

Per spec §7 Constraints:

| Performance criterion | Target | Verification | Notes |
|----------------------|--------|--------------|-------|
| Resolver Resolve() — 1K tags, no fan_in | <100ms per call | `BenchmarkResolver_Resolve_1KTags` (T-SPC004-13) | Advisory; CI does not enforce |
| Resolver Resolve() — 50 ANCHOR + LSP fan_in | <2s per call | `BenchmarkResolver_Resolve_50AnchorsLSP` (T-SPC004-13) | Advisory; uses mock LSP client |
| CLI response size cap | 10MB JSON | (bounded by `--limit` default 100 + max 10000) | implicit; no separate test |
| 16-language coverage | All 16 languages parse + associate | T-SPC004-15 16-lang sweep | Validates resolver layer (SPC-002 validates scanner) |

---

## 4. Definition of Done

The SPEC is considered DONE when:

- [ ] All 15 ACs (AC-SPC-004-01 through AC-SPC-004-15) verified
- [ ] All 18 REQs (REQ-SPC-004-001..007, 010..013, 020..021, 030..031, 040..042) traced to ≥1 task per plan §1.5
- [ ] All 20 tasks (T-SPC004-01..20 per tasks.md) marked complete
- [ ] `go test -race -count=1 ./...` PASS (no regressions)
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates `internal/template/embedded.go` correctly
- [ ] Coverage ≥ 85% per modified package (`internal/mx/`, `internal/cli/`)
- [ ] Template parity verified via `diff -r .claude/ internal/template/templates/.claude/` (no template change expected)
- [ ] CHANGELOG entry written in Unreleased section (4 entries per plan §M6)
- [ ] @MX tags applied per plan §6 (6 tags: 1 ANCHOR + 2 WARN + 3 NOTE)
- [ ] Run PR squash-merged into main
- [ ] Sync PR squash-merged into main
- [ ] Worktree disposed via `moai worktree done SPEC-V3R2-SPC-004`
- [ ] Manual end-to-end verification: real project with gopls running, `moai mx query --kind anchor --fan-in-min 1` returns ≥1 entry with `fan_in_method: "lsp"`

---

## 5. Quality Gate Criteria

Per `.moai/config/sections/quality.yaml`:

| Criterion | Target | Verification |
|-----------|--------|--------------|
| Tested | All new functions covered | `go test -cover ./internal/mx/ ./internal/cli/` ≥ 85% |
| Readable | English code + ko inline comments (per language.yaml) | `golangci-lint run` clean |
| Unified | Style + import order | `gofmt -l ./internal/` empty |
| Secured | No new attack surface | LSP path is read-only on source (no exec); spec_loader is yaml-parse on `.moai/specs/` (read-only) |
| Trackable | All commits + CHANGELOG | git log inspection + CHANGELOG diff |

---

## 6. Risk-based AC Prioritization

| AC | Priority | Why |
|---|---|---|
| AC-01, AC-02, AC-03 | P0 | Core: SPEC + kind + fan_in + danger filters (most consumed by callers) |
| AC-05, AC-12 | P0 | JSON contract (REQ-005 schema) + empty result handling — tooling depends on it |
| AC-08 | P0 | Pagination + TruncationNotice (output size safeguard) |
| AC-04 | P0 | SidecarUnavailable graceful failure (CLI UX) |
| AC-07 | P1 | LSP fallback annotation (fan_in correctness signal) |
| AC-09 | P1 | Strict mode (CI ergonomics) |
| AC-13 | P1 | InvalidQuery exit 2 (script integration) |
| AC-14 | P1 | AND composition (compositional semantics) |
| AC-15 | P1 | Body-based association (regex correctness) |
| AC-06, AC-10 | P2 | Alternative output formats (less common path) |
| AC-11 | P2 | test fixture exclusion (ergonomics) |

---

End of acceptance.

Version: 0.1.0
Status: Acceptance artifact for SPEC-V3R2-SPC-004
