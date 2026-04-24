---
id: SPEC-UTIL-002
title: "ast-grep Integration Hardening + Suppression Policy + 5-Language Rule Seeding"
version: "0.1.0"
status: implemented
created: 2026-04-24
updated: 2026-04-24
author: Wave v2.14 SPEC Writer
priority: P0 Critical
phase: "v2.14.0 — Phase 2 — Utility Hardening"
module: "internal/astgrep/, internal/hook/security/, internal/hook/quality/, .moai/config/astgrep-rules/"
dependencies: []
related_problem: [IMP-V3U-004, IMP-V3U-006]
related_pattern: []
related_principle: []
related_decision: [D6, D7]
related_theme: "v2.14.0 Utility Hardening"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "ast-grep, suppression, rule-seeding, multi-language, v2.14"
---

# SPEC-UTIL-002: ast-grep Integration Hardening + Suppression Policy + 5-Language Rule Seeding

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.1 | 2026-04-24 | MoAI (release/v2.14.0) | Implementation complete, merged as commit `1341fa544`. AC coverage 21/21. PHP `$_ = null` pattern replaced with `return null` / `json_decode(null)` due to tree-sitter grammar limitation (semantic equivalent). TRUST 5: 5/5. |
| 0.1.0 | 2026-04-24 | Wave v2.14 SPEC Writer | Initial v2.14.0 draft. Owns IMP-V3U-004 (Rule Metadata propagation) + IMP-V3U-006 (ScanMultiple semaphore). Absorbs D6 (suppression policy lint) and D7 (5-language rule seeding: Ruby/PHP/Elixir/C#/Kotlin). Parallel with SPEC-UTIL-001 (MX) and SPEC-UTIL-003 (LSP). All changes non-breaking. |

---

## 1. Goal (목적)

Close the three ast-grep integration gaps that block production-grade OWASP/CWE reporting and multi-language coverage while establishing a sanctioned suppression policy with an auditable rationale trail. The SARIF pipeline currently discards OWASP/CWE metadata from YAML rule files (IMP-V3U-004), the `ScanMultiple` concurrency path spawns one subprocess per input file without a cap (IMP-V3U-006), the SARIF `tool.driver.version` is permanently `"unknown"` because `detectSGVersion()` is a placeholder, there is no documented way to suppress a false positive, and 11 of 16 declared language directories contain only `.gitkeep`. v2.14.0 resolves all five items within the non-breaking semver-minor gate.

Rule metadata propagation (D7 expansion) delivers two independently valuable outcomes at once: (a) the five existing security rules surface their OWASP/CWE classifications to consumers for the first time, and (b) the 15 newly seeded rule YAML files exercise the propagation path end-to-end with per-language CWE citations. The suppression policy (D6) establishes a middle path between unrestricted `// ast-grep-ignore` (current unknown-behavior) and a blanket `no-suppress-all` rule (A4 T2.1 rec-3, rejected as too rigid): allow targeted suppression IFF paired with a `// @MX:REASON <rationale>` comment on the adjacent line, consistent with the existing `@MX:WARN` / `@MX:ANCHOR` convention.

5-language rule seeding closes half of the 16-language neutrality gap (10/16 coverage after v2.14 versus 5/16 today) without overextending a bug-fix release. Remaining 6 languages (Java, C++, Scala, R, Flutter, Swift) are explicitly deferred to v2.15+ or community contribution.

## 2. Scope (범위)

In-scope:

- `internal/astgrep/models.go` — `Rule` struct gains `Note string` and `Metadata map[string]string` fields with YAML + JSON tags.
- `internal/astgrep/scanner.go` — `scanWithRules` propagates `rule.Note` and `rule.Metadata` onto each `Finding` when the finding's fields are empty.
- `internal/astgrep/sarif.go` — SARIF generator extracts OWASP and CWE entries from `finding.Metadata` into `properties.tags` (additive, SARIF 2.1.0 §3.52 compliant).
- `internal/hook/security/ast_grep.go` — `ScanMultiple` acquires a semaphore of capacity `runtime.NumCPU() * 2` before each goroutine executes the subprocess scan; the existing `@MX:WARN` tag is updated to reflect the bound.
- `internal/cli/astgrep.go` — `detectSGVersion()` invokes `sg --version`, caches the result via `sync.Once`, and populates SARIF `tool.driver.version`.
- `internal/hook/quality/astgrep_gate.go` — new `checkSuppressionPairing(filePath) []SuppressionViolation` helper runs after scan; emits `SUPPRESSION_WITHOUT_REASON` error with exit code 2 when any `// ast-grep-ignore` lacks an adjacent `// @MX:REASON` within 1 line.
- `.moai/config/astgrep-rules/{ruby,php,elixir,csharp,kotlin}/` — 15 new YAML rule files (3 per language): unused-var, null-deref (or language idiom equivalent), todo-marker. Each carries `metadata: {owasp: "...", cwe: "..."}` and an educational `note:`.
- `internal/astgrep/testdata/fixtures/{ruby,php,elixir,csharp,kotlin}/` — per-language fixture directories with `valid.<ext>`, `violation.<ext>`, and `suppressed.<ext>` files.
- `internal/astgrep/rule_seed_test.go` — golden-file smoke tests asserting violation counts, metadata propagation, and suppression policy enforcement per language.
- `internal/hook/quality/astgrep_gate_suppression_test.go` — suppression policy lint unit tests (paired, unpaired, wrong distance).

Out-of-scope (deferred to v2.15+ or later minors):

- **Remaining 6 language rule directories** — Java, C++, Scala, R, Flutter, Swift — deferred to v2.15+ per D7. v2.14 delivers 10/16 coverage; full 16-lang neutrality is a multi-release initiative not a single SPEC.
- **Rule body expansion beyond 3 foundational rules per language** — each new language gets exactly unused-var + null-deref (or equivalent) + todo-marker. Security-specific rules (SQL injection, XSS, path traversal) for non-Go languages deferred.
- **Context cancel leak in `analyzer.go`** (IMP-V3U-018, A2 REC-04) — routed to v2.15 per release plan §2.3.
- **JSON parser triplication consolidation** (IMP-V3U-011, A2 REC-05) — v2.16 types domain.
- **SARIF `file://` URI prefix** (IMP-V3U-030, A2 REC-08) — v2.17 or community contribution; not a correctness blocker.
- **V1/V2 gate deduplication** (A2 §D4-1) — V1 retirement deferred to post-v2.14 once V2 coverage confirmed stable.
- **`ast-grep-ignore-next-line` alternative syntax** — v2.14 supports only `// ast-grep-ignore` on its own line; block-level and inline variants deferred.
- **Parameterized utils adoption** (A4 T2.1 rec-4) — deferred; seeded rules use plain patterns for ast-grep ≥ 0.40 compatibility.
- **Minimum sg version bump above 0.40** — v2.14 pins minimum ≥ 0.40 (current production floor). Upgrade to ≥ 0.42.1 deferred to v2.15 per release plan §4.1.

## 3. Environment (환경)

Current moai-adk-go state (from research.md §2):

- `internal/astgrep/models.go:72-79` — `Rule` has 6 fields (ID, Language, Severity, Message, Pattern, Fix). No `Note` or `Metadata`.
- `internal/astgrep/scanner.go:340-380` — `scanWithRules` propagates RuleID, Severity, Message, Language to Finding but does not copy Note/Metadata (impossible without the fields).
- `internal/astgrep/scanner.go:43-73` — `Finding` already has `Note string` and `Metadata map[string]string`. Downstream-ready.
- `internal/hook/security/ast_grep.go:172-213` — `ScanMultiple` spawns `len(filePaths)` goroutines; existing `@MX:WARN` tag (line 172) acknowledges the risk.
- `internal/cli/astgrep.go:192-198` — `detectSGVersion()` placeholder returns `"unknown"`.
- `.moai/config/astgrep-rules/` — 2 populated directories (`go/`, `security/`) + 15 `.gitkeep`-only directories.
- `RunAstGrepGateV2` (`internal/hook/quality/astgrep_gate.go:20-58`) is the canonical quality gate entry; suppression lint integrates after its scan call.

ast-grep runtime baseline:

- Minimum pinned version: 0.40 (documented in `internal/astgrep/doc.go` post-v2.14).
- Production floor at time of release plan authoring (2026-04-24): 0.40.5.
- Upstream latest: 0.42.1 (A4 T2.1, released 2026-04-04).

Platform baseline:

- Go 1.23+ (project go.mod minimum).
- Linux / macOS / Windows supported; Windows CI gate required for suppression policy lint (uses `bufio.Scanner`, no subprocess dependency).

## 4. Assumptions (가정)

- ast-grep `sg` binary is installed and on PATH with version ≥ 0.40. When absent, the existing graceful-degradation path in `Scanner.Scan` returns no findings — suppression policy lint still runs independently.
- YAML rule parser (`gopkg.in/yaml.v3`) accepts `metadata: map[string]string` deserialization from arbitrary string-valued keys. Confirmed by manual inspection of existing `.moai/config/astgrep-rules/security/injection.yml:8-13` which already uses this shape.
- SARIF consumers (GitHub Code Scanning, Azure DevOps, SonarCloud) accept additive `properties.tags[]` entries without breaking — confirmed by OASIS SARIF 2.1.0 §3.52 Property Bag specification.
- `sg --version` output format is stable across ast-grep 0.40.x through 0.42.x: single line `"ast-grep 0.40.5"` or similar. Any parse failure falls back to `"unknown"` (backward-compatible).
- Seeded language rules (Ruby/PHP/Elixir/C#/Kotlin) compile against ast-grep's tree-sitter grammars for each language. Confirmed: ast-grep 0.40+ supports all five via `builtin-parser` feature flag (A4 T2.1).
- `@MX:REASON` convention per `.claude/rules/moai/workflow/mx-tag-protocol.md` is the authoritative rationale tag. No new MX sub-line type is introduced; suppression policy reuses the existing convention.
- `runtime.NumCPU()` returns a sensible value on all supported platforms (Linux/macOS/Windows). Multi-CPU containers with CPU quotas (e.g., Kubernetes cgroup limits) may return more cores than allocated; this is acceptable — the semaphore bounds goroutines, not kernel-level concurrency.

## 5. Requirements (요구사항, EARS)

### 5.1 Rule metadata data model (IMP-V3U-004)

- **REQ-UTIL-002-001**: The Rule struct in `internal/astgrep/models.go` SHALL declare a `Note string` field with YAML tag `yaml:"note,omitempty"` and JSON tag `json:"note,omitempty"`.

- **REQ-UTIL-002-002**: The Rule struct in `internal/astgrep/models.go` SHALL declare a `Metadata map[string]string` field with YAML tag `yaml:"metadata,omitempty"` and JSON tag `json:"metadata,omitempty"`.

### 5.2 Metadata propagation

- **REQ-UTIL-002-003**: When `scanWithRules` in `internal/astgrep/scanner.go` iterates findings returned from `runSingleRule`, it SHALL copy `rule.Note` into `findings[i].Note` if and only if the finding's Note is empty.

- **REQ-UTIL-002-004**: When `scanWithRules` in `internal/astgrep/scanner.go` iterates findings returned from `runSingleRule`, it SHALL copy `rule.Metadata` into `findings[i].Metadata` if and only if the finding's Metadata is nil.

### 5.3 SARIF propagation

- **REQ-UTIL-002-005**: When `internal/astgrep/sarif.go` generates a SARIF result from a Finding whose Metadata map contains an `owasp` key, the generator SHALL emit an entry `external/owasp/<lowercased-sanitized-value>` into the result's `properties.tags` array.

- **REQ-UTIL-002-006**: When `internal/astgrep/sarif.go` generates a SARIF result from a Finding whose Metadata map contains a `cwe` key, the generator SHALL emit an entry `external/cwe/<lowercased-sanitized-value>` into the result's `properties.tags` array.

### 5.4 Bounded concurrency (IMP-V3U-006)

- **REQ-UTIL-002-007**: When `ScanMultiple` in `internal/hook/security/ast_grep.go` is called with N input files, it SHALL cap the number of concurrently active `Scan` invocations at `runtime.NumCPU() * 2` via a buffered channel semaphore acquired before `s.Scan` and released after `s.Scan` returns.

### 5.5 Version detection

- **REQ-UTIL-002-008**: When `detectSGVersion` in `internal/cli/astgrep.go` is invoked, it SHALL execute `sg --version` with a 5-second context timeout, parse the trimmed stdout, and return the parsed value. On any error (binary missing, timeout, non-zero exit, empty stdout), it SHALL return the literal string `"unknown"`.

- **REQ-UTIL-002-009**: `detectSGVersion` SHALL cache its result via `sync.Once` so that repeated calls within the same process invoke `sg --version` exactly once.

### 5.6 Suppression policy (D6)

- **REQ-UTIL-002-010**: When a source file contains a comment matching `// ast-grep-ignore` on its own line (case-sensitive, language-prefix-aware for `#` Ruby/Python/Elixir and `--` Haskell), the suppression policy lint SHALL require a second comment matching `// @MX:REASON <non-empty-text>` (with appropriate language prefix substitution) on the next non-blank line or within 1 intervening blank line.

- **REQ-UTIL-002-011**: Where an `// ast-grep-ignore` comment exists without a paired `// @MX:REASON` within the allowed distance, the suppression policy lint SHALL emit a structured error with fields `type: "SUPPRESSION_WITHOUT_REASON"`, `file: <path>`, `line: <line-number>`, and `message: "ast-grep suppression at <file>:<line> requires adjacent '// @MX:REASON <rationale>' on next line"`.

- **REQ-UTIL-002-012**: When the suppression policy lint detects one or more `SUPPRESSION_WITHOUT_REASON` violations during a quality gate run, `RunAstGrepGateV2` SHALL return `(false, <formatted-violation-list>)` causing the gate to exit with code 2.

### 5.7 5-language rule seeding (D7)

- **REQ-UTIL-002-013**: The directory `.moai/config/astgrep-rules/ruby/` SHALL contain exactly three YAML rule files: `unused-var.yml`, `null-deref.yml` (or the Ruby-idiomatic equivalent `nil-method-call.yml`), and `todo-marker.yml`. Each file SHALL declare `language: ruby`, a valid `pattern:`, a `severity:` value from `{info, warning, error}`, a `message:`, a `note:`, and a `metadata:` map containing at least keys `owasp` and `cwe`.

- **REQ-UTIL-002-014**: The directory `.moai/config/astgrep-rules/php/` SHALL contain exactly three YAML rule files analogous to REQ-UTIL-002-013 with `language: php`.

- **REQ-UTIL-002-015**: The directory `.moai/config/astgrep-rules/elixir/` SHALL contain exactly three YAML rule files analogous to REQ-UTIL-002-013 with `language: elixir`. The null-deref equivalent SHALL use an Elixir pattern-matching idiom (e.g., unsafe map field access) rather than the literal `null` concept.

- **REQ-UTIL-002-016**: The directory `.moai/config/astgrep-rules/csharp/` SHALL contain exactly three YAML rule files analogous to REQ-UTIL-002-013 with `language: csharp`.

- **REQ-UTIL-002-017**: The directory `.moai/config/astgrep-rules/kotlin/` SHALL contain exactly three YAML rule files analogous to REQ-UTIL-002-013 with `language: kotlin`. The null-deref rule SHALL specifically target the `!!` non-null assertion operator.

### 5.8 Test coverage

- **REQ-UTIL-002-018**: For each of the 5 new languages, the repository SHALL provide a fixture directory `internal/astgrep/testdata/fixtures/<lang>/` containing `valid.<ext>`, `violation.<ext>`, and `suppressed.<ext>` files used by `internal/astgrep/rule_seed_test.go`.

- **REQ-UTIL-002-019**: `internal/astgrep/rule_seed_test.go` SHALL contain per-language test cases that (a) assert the violation fixture produces at least one finding, (b) assert the valid fixture produces zero findings, and (c) assert that each finding from the violation fixture carries a non-empty `Metadata["cwe"]` value.

- **REQ-UTIL-002-020**: `internal/hook/quality/astgrep_gate_suppression_test.go` SHALL contain at minimum three test cases: a correctly paired suppression producing zero violations, an unpaired `// ast-grep-ignore` producing one `SUPPRESSION_WITHOUT_REASON` violation, and a suppression separated from its `@MX:REASON` by an intervening non-blank comment producing one `SUPPRESSION_WITHOUT_REASON` violation.

### 5.9 Non-breaking enforcement

- **REQ-UTIL-002-021**: All changes SHALL be non-breaking per the release plan §1.2 three-way gate: no user-visible API change, no user-visible config key change, no user-visible wire format change. SARIF additions (REQ-UTIL-002-005, REQ-UTIL-002-006) SHALL only add properties, never rename or remove existing ones.

## 6. Acceptance Criteria (인수 기준)

Given/When/Then scenarios map 1:1 to REQs 001-020; REQ-021 is a cross-cutting invariant verified by the absence of wire-format changes in `go test -run TestSARIF_BackwardCompatibility` and manual CHANGELOG review.

### AC-UTIL-002-01 (→ REQ-UTIL-002-001)

- **Given** the `Rule` struct in `internal/astgrep/models.go` after v2.14.0 changes
- **When** `go doc github.com/modu-ai/moai-adk/internal/astgrep.Rule` is executed
- **Then** the output lists a field `Note string` with YAML tag `yaml:"note,omitempty"` and JSON tag `json:"note,omitempty"`

### AC-UTIL-002-02 (→ REQ-UTIL-002-002)

- **Given** the `Rule` struct in `internal/astgrep/models.go` after v2.14.0 changes
- **When** `go doc github.com/modu-ai/moai-adk/internal/astgrep.Rule` is executed
- **Then** the output lists a field `Metadata map[string]string` with YAML tag `yaml:"metadata,omitempty"` and JSON tag `json:"metadata,omitempty"`

### AC-UTIL-002-03 (→ REQ-UTIL-002-003)

- **Given** a rule YAML file declaring `note: "sample note"` loaded via `RuleLoader.LoadFromDir`
- **When** `Scanner.Scan` runs against a file that matches the rule's pattern
- **Then** each produced Finding's `Note` field equals `"sample note"` (assuming the runSingleRule call returned a finding with empty Note)

### AC-UTIL-002-04 (→ REQ-UTIL-002-004)

- **Given** a rule YAML file declaring `metadata: {owasp: "A03:2021", cwe: "CWE-89"}` loaded via `RuleLoader.LoadFromDir`
- **When** `Scanner.Scan` runs against a file that matches the rule's pattern
- **Then** each produced Finding's `Metadata` map equals `{"owasp": "A03:2021", "cwe": "CWE-89"}` (assuming the runSingleRule call returned a finding with nil Metadata)

### AC-UTIL-002-05 (→ REQ-UTIL-002-005)

- **Given** a Finding with `Metadata = {"owasp": "A03:2021 - Injection"}`
- **When** the SARIF generator produces the result for this finding
- **Then** the generated SARIF result's `properties.tags` array contains an entry `external/owasp/a03-2021-injection` (or equivalently-sanitized lowercased form)

### AC-UTIL-002-06 (→ REQ-UTIL-002-006)

- **Given** a Finding with `Metadata = {"cwe": "CWE-89"}`
- **When** the SARIF generator produces the result for this finding
- **Then** the generated SARIF result's `properties.tags` array contains an entry `external/cwe/cwe-89`

### AC-UTIL-002-07 (→ REQ-UTIL-002-007)

- **Given** a system with `runtime.NumCPU() == 4` and 100 input file paths
- **When** `ScanMultiple` is invoked with these 100 paths
- **Then** the peak number of concurrently-executing `s.Scan` invocations observed by a test-instrumented counter never exceeds 8 throughout the execution

### AC-UTIL-002-08 (→ REQ-UTIL-002-008)

- **Given** ast-grep version 0.40.5 is installed on PATH
- **When** `detectSGVersion()` is called the first time
- **Then** it returns `"ast-grep 0.40.5"` (or whatever the actual first-line trimmed output of `sg --version` is)
- **And Given** ast-grep is removed from PATH and a fresh process is spawned
- **When** `detectSGVersion()` is called
- **Then** it returns the literal string `"unknown"`

### AC-UTIL-002-09 (→ REQ-UTIL-002-009)

- **Given** a fresh process where `detectSGVersion()` has not yet been called
- **When** `detectSGVersion()` is called 10 times in a loop
- **Then** `sg --version` is observed to have been executed exactly once (verified via test-injected executor counter or process-list inspection)

### AC-UTIL-002-10 (→ REQ-UTIL-002-010)

- **Given** a Go source file containing:
  ```go
  // ast-grep-ignore
  // @MX:REASON generated stub, pattern false positive
  buf := getUnsafeBuffer()
  ```
- **When** `checkSuppressionPairing` scans the file
- **Then** the returned violation slice is empty

### AC-UTIL-002-11 (→ REQ-UTIL-002-011)

- **Given** a Go source file containing:
  ```go
  // ast-grep-ignore
  buf := getUnsafeBuffer()  // no @MX:REASON pairing
  ```
- **When** `checkSuppressionPairing` scans the file
- **Then** the returned violation slice contains exactly one entry with `type = "SUPPRESSION_WITHOUT_REASON"`, `file = <path>`, `line = <line-of-ast-grep-ignore>`, and `message` starting with `"ast-grep suppression at"`

### AC-UTIL-002-12 (→ REQ-UTIL-002-012)

- **Given** a project containing at least one file with an unpaired `// ast-grep-ignore`
- **When** `RunAstGrepGateV2` runs against the project
- **Then** the function returns `(false, output)` where `output` contains the string `SUPPRESSION_WITHOUT_REASON` and the caller (quality gate hook) exits with code 2

### AC-UTIL-002-13 (→ REQ-UTIL-002-013)

- **Given** the `.moai/config/astgrep-rules/ruby/` directory after v2.14.0 changes
- **When** `ls` is executed
- **Then** exactly 3 YAML files are present, each loadable via `RuleLoader.LoadFromDir` without parse error, each declaring `language: ruby` and carrying `metadata.owasp` + `metadata.cwe` keys

### AC-UTIL-002-14 (→ REQ-UTIL-002-014)

- **Given** the `.moai/config/astgrep-rules/php/` directory after v2.14.0 changes
- **When** `ls` is executed and each file is parsed
- **Then** the 3-rule / language / metadata invariants from AC-UTIL-002-13 hold with `language: php`

### AC-UTIL-002-15 (→ REQ-UTIL-002-015)

- **Given** the `.moai/config/astgrep-rules/elixir/` directory after v2.14.0 changes
- **When** `ls` is executed and each file is parsed
- **Then** the 3-rule / language / metadata invariants from AC-UTIL-002-13 hold with `language: elixir` and the non-null-deref-equivalent rule uses an Elixir-idiomatic pattern (verified by reading the rule YAML's `pattern:` field and confirming it does not reference a `null` or `nil` literal check)

### AC-UTIL-002-16 (→ REQ-UTIL-002-016)

- **Given** the `.moai/config/astgrep-rules/csharp/` directory after v2.14.0 changes
- **When** `ls` is executed and each file is parsed
- **Then** the 3-rule / language / metadata invariants from AC-UTIL-002-13 hold with `language: csharp`

### AC-UTIL-002-17 (→ REQ-UTIL-002-017)

- **Given** the `.moai/config/astgrep-rules/kotlin/` directory after v2.14.0 changes
- **When** `ls` is executed and each file is parsed
- **Then** the 3-rule / language / metadata invariants from AC-UTIL-002-13 hold with `language: kotlin` AND the null-deref rule's `pattern:` field contains the substring `!!`

### AC-UTIL-002-18 (→ REQ-UTIL-002-018)

- **Given** the `internal/astgrep/testdata/fixtures/` directory after v2.14.0 changes
- **When** `ls` lists each of the 5 language subdirectories
- **Then** each subdirectory contains `valid.<ext>`, `violation.<ext>`, and `suppressed.<ext>` files where `<ext>` is the canonical extension for that language (rb, php, ex, cs, kt)

### AC-UTIL-002-19 (→ REQ-UTIL-002-019)

- **Given** the test file `internal/astgrep/rule_seed_test.go` after v2.14.0 changes
- **When** `go test -run TestRuleSeed ./internal/astgrep/...` executes
- **Then** all test cases pass, with each per-language subtest asserting at minimum three conditions: violation fixture produces ≥ 1 finding, valid fixture produces 0 findings, at least one finding from the violation fixture has non-empty `Metadata["cwe"]`

### AC-UTIL-002-20 (→ REQ-UTIL-002-020)

- **Given** the test file `internal/hook/quality/astgrep_gate_suppression_test.go` after v2.14.0 changes
- **When** `go test -run TestSuppressionPairing ./internal/hook/quality/...` executes
- **Then** all test cases pass, including at minimum the three scenarios: paired-correctly (0 violations), unpaired (1 violation), separated-by-unrelated-comment (1 violation)

## 7. Constraints (제약)

- **Non-breaking**: SARIF output adds properties; never renames or removes. Hook JSON wire format unchanged. CLI flag surface unchanged. Config key surface unchanged.
- **Go version**: Go 1.23+ (project go.mod minimum).
- **ast-grep version**: Minimum ≥ 0.40 (documented). Seeded rules use plain pattern syntax stable since 0.20.x; no parameterized utils or ESQuery selectors.
- **Platform**: suppression policy lint is pure-Go (`bufio.Scanner` + `regexp`). Runs on Linux/macOS/Windows with identical behavior.
- **Dependency policy**: no new third-party imports. `runtime`, `sync`, `context`, `os/exec`, `bufio`, `regexp` all stdlib.
- **TRUST 5 coverage**: each new Go function covered by table-driven tests; each new rule YAML covered by fixture; overall package coverage remains ≥ 85% per project baseline.
- **@MX tag discipline**: `ScanMultiple`'s existing `@MX:WARN` tag at `ast_grep.go:172` MUST be updated in the same commit that introduces the semaphore to reflect the new bounded behavior. The updated REASON MUST reference this SPEC ID.
- **Language neutrality**: all 5 new languages receive equivalent 3-rule structure. No language is marked "PRIMARY" or given preferential depth within v2.14.
- **No time estimates**: implementation sequencing uses priority labels (P0 Critical), not duration estimates.

## 8. Risks & Mitigations (리스크)

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|-----------|--------|------------|
| R-U2-1 | SARIF `properties.tags[]` additions break a strict-schema external consumer | Low | Low | SARIF 2.1.0 §3.52 explicitly permits additive property bag entries; CHANGELOG documents new tag families (`external/cwe/*`, `external/owasp/*`); `TestSARIF_BackwardCompatibility` snapshot pins existing fields verbatim |
| R-U2-2 | Seeded rule has unacceptable false-positive rate (especially Ruby `unused-var` and Elixir unsafe-map-access) | Medium | Low | Severity calibrated to `info` for high-FP rules; fixture tests pin current behavior so FP rate is measurable; suppression policy (D6) gives operators an auditable escape hatch; CHANGELOG notes severity tuning policy |
| R-U2-3 | `detectSGVersion` timing impact on startup-critical paths | Low | Low | `sync.Once` ensures single invocation per process; 5s context timeout bounds worst case; fallback to `"unknown"` is non-fatal |
| R-U2-4 | ast-grep version < 0.40 in user environment produces scan errors | Low | Medium | Graceful-degradation path in existing Scanner returns empty findings on subprocess error; suppression policy lint runs independently of sg binary so core gate still functions; `detectSGVersion` surfaces the installed version in SARIF for diagnosis |
| R-U2-5 | Pure-Go suppression lint misclassifies multi-line comment blocks (`/* ... */` in languages that support it) | Medium | Low | Initial implementation scans only line comments (`//`, `#`, `--`); `/* */` block comment support deferred to future minor; test case documents the limitation; users who need block-comment suppression fall back to inserting a line comment |
| R-U2-6 | Kotlin `!!` pattern flags legitimate idiomatic usage in test code | Medium | Low | Rule severity set to `warning` not `error`; Kotlin convention discourages `!!` in production code but tolerates it in tests; rule message directs reviewer to suppression policy for justified uses |
| R-U2-7 | Semaphore of `runtime.NumCPU()*2` is too conservative on high-core-count machines (64+ cores), leaving scan performance on the table | Low | Low | `NumCPU()*2` is the SYNTHESIS §2.1 REC-02 recommendation and matches IMP-V3U-006 spec; if performance data post-v2.14 shows underutilization, v2.15 can tune — not a correctness risk |
| R-U2-8 | Scope creep: contributors propose adding 6th-10th language seeds mid-release | Low | Medium | D7 decision binds v2.14 scope to Ruby/PHP/Elixir/C#/Kotlin; any addition requires plan amendment; release plan §2.3 explicitly defers 6 remaining languages to v2.15+ |

## 9. Dependencies (의존성)

### 9.1 Internal dependencies

- **SPEC-UTIL-001** (parallel, v2.14.0): establishes `@MX:REASON` enforcement convention for WARN/ANCHOR tags. SPEC-UTIL-002 extends the same convention to ast-grep suppressions. No code-level dependency — both SPECs modify different packages (`internal/hook/mx/` vs `internal/astgrep/` + `internal/hook/{security,quality}/`). Parallelizable.
- **SPEC-UTIL-003** (parallel, v2.14.0): LSP subprocess hygiene. No file overlap. Parallel Phase 3 execution per release plan §5.
- **SPEC-V3R2-EXT-*** (v3R2 planning track): possible future owner of rule catalog expansion. No v2.14 dependency.

### 9.2 External dependencies

- `gopkg.in/yaml.v3` — already in go.sum; YAML rule parsing. No version bump required.
- `ast-grep` binary (`sg`) — user-installed runtime dependency. Minimum ≥ 0.40 documented in post-v2.14 `internal/astgrep/doc.go`.

### 9.3 No dependencies

This SPEC does not depend on any other v2.14.0 SPEC. Implementation can begin immediately after SPEC-UTIL-001 merge (per release plan §5 Phase 2 entry criterion).

## 10. Traceability (추적성)

### 10.1 Audit source traceability

| REQ ID | Audit citation | File:line |
|--------|----------------|-----------|
| REQ-UTIL-002-001 | A2 §D2-2 REC-01 / SYNTHESIS §2.1 IMP-V3U-004 | `internal/astgrep/models.go:72` |
| REQ-UTIL-002-002 | A2 §D2-2 REC-01 / SYNTHESIS §2.1 IMP-V3U-004 | `internal/astgrep/models.go:72` |
| REQ-UTIL-002-003 | A2 §D2-2 REC-01 | `internal/astgrep/scanner.go:361` |
| REQ-UTIL-002-004 | A2 §D2-2 REC-01 | `internal/astgrep/scanner.go:361` |
| REQ-UTIL-002-005 | A2 §D2-2 derived; OASIS SARIF 2.1.0 §3.52 | `internal/astgrep/sarif.go` |
| REQ-UTIL-002-006 | A2 §D2-2 derived; OASIS SARIF 2.1.0 §3.52 | `internal/astgrep/sarif.go` |
| REQ-UTIL-002-007 | A2 §D5-3 REC-02 / SYNTHESIS §2.1 IMP-V3U-006 | `internal/hook/security/ast_grep.go:186` |
| REQ-UTIL-002-008 | A2 §D1-3 REC-03 / SYNTHESIS §2.1 IMP-V3U-019 (pulled into v2.14) | `internal/cli/astgrep.go:193` |
| REQ-UTIL-002-009 | A2 §D1-3 REC-03 derived (caching) | `internal/cli/astgrep.go:193` |
| REQ-UTIL-002-010 | D6 decision (release plan §0.5) | `internal/hook/quality/astgrep_gate.go` (new helper) |
| REQ-UTIL-002-011 | D6 decision (release plan §0.5) | `internal/hook/quality/astgrep_gate.go` (new helper) |
| REQ-UTIL-002-012 | D6 decision (release plan §0.5) | `internal/hook/quality/astgrep_gate.go:20` |
| REQ-UTIL-002-013..017 | D7 decision (release plan §0.5) / SYNTHESIS §4.5 16-lang gap | `.moai/config/astgrep-rules/{ruby,php,elixir,csharp,kotlin}/` |
| REQ-UTIL-002-018..020 | TRUST 5 Tested pillar | `internal/astgrep/testdata/fixtures/` + new test files |
| REQ-UTIL-002-021 | Release plan §1.2 non-breaking gate | Cross-cutting invariant |

### 10.2 Decision traceability

- **D6 (suppression policy)**: release plan §0.5 — AskUserQuestion dialog 2026-04-24. Choice: allow `// ast-grep-ignore` IFF paired with `// @MX:REASON` within 1 line. Implements REQ-UTIL-002-010/011/012.
- **D7 (5-language seeding)**: release plan §0.5 — scope scoped to Ruby/PHP/Elixir/C#/Kotlin. Implements REQ-UTIL-002-013..017.

### 10.3 SPEC cross-references

- Parallel SPEC-UTIL-001 for `@MX:REASON` enforcement standard.
- Parallel SPEC-UTIL-003 for LSP subprocess hygiene (no file overlap).
- Deferred items routed to v2.15+ per release plan §2.3 / §4.1.

### 10.4 Exclusions (what NOT to build)

- Java, C++, Scala, R, Flutter, Swift rule directories — remain `.gitkeep`-only after v2.14.0; v2.15+ scope.
- Security-specific rules (SQL injection, XSS, etc.) for the 5 new languages — v2.14.0 seeds only 3 foundational rules per language.
- Context cancel leak fix in `analyzer.go` (A2 REC-04) — v2.15 scope.
- JSON parser triplication (A2 REC-05) — v2.16 types domain.
- SARIF `file://` URI prefix (A2 REC-08) — v2.17 or community.
- V1 gate retirement — post-v2.14.
- `ast-grep-ignore-next-line` / inline variants — v2.14 supports single-line form only.
- Parameterized utils / ESQuery selectors adoption — deferred.
- sg version floor bump above 0.40 — v2.15 minimum.

---

Version: 0.1.0
Classification: draft
Owner: v2.14.0 release track
Last updated: 2026-04-24
Supersedes: none
Referenced by: `docs/design/v2.14.0-release-plan.md` §2.2, §6.1
