# Research — SPEC-UTIL-002

**Scope**: ast-grep Integration Hardening + Suppression Policy + 5-Language Rule Seeding
**Phase**: v2.14.0 — Phase 2 — Utility Hardening
**Audit Sources**: `.moai/design/utility-review/a2-astgrep-audit.md` (A2, health 6/10, verdict REFACTOR), `.moai/design/utility-review/SYNTHESIS.md` §2, `.moai/design/utility-review/a4-external-best-practices.md` §T2
**Decisions Referenced**: D6 (suppression policy), D7 (5-language rule seeding) — `docs/design/v2.14.0-release-plan.md` §0.5
**Non-breaking constraint**: SARIF field additions are additive; no user-visible API/config/wire format change.

---

## 1. Problem Statement

A2 classifies the ast-grep integration layer at 6/10 with a REFACTOR verdict. Five issues block production readiness:

1. **OWASP/CWE metadata silently dropped** (A2 §D2-2, REC-01). YAML security rules encode `metadata: {owasp: "A03:2021", cwe: "CWE-89"}` and `note:` fields that are never persisted — the `Rule` struct has no matching fields, so `scanWithRules` (`scanner.go:361-374`) propagates nothing to `Finding` or SARIF output. All 5 security rule classifications never reach the user.
2. **Unbounded goroutine fan-out** (A2 §D5-3, SYNTHESIS §2.1 REC-02). `security/ast_grep.go:186 ScanMultiple` spawns one goroutine per input file without a semaphore. 10,000 input files → 10,000 concurrent `sg` subprocesses, exhausting OS file descriptors and process table slots.
3. **SARIF `tool.driver.version` always empty** (A2 §D1-3, REC-03). `cli/astgrep.go:193 detectSGVersion()` is a documented placeholder that always returns `"unknown"`. CI drift risk: when a developer upgrades to a breaking ast-grep version, the SARIF output gives no audit trail.
4. **Suppression policy absent**. Audit-time agents (expert-debug, manager-quality) and human operators have no sanctioned way to silence a false positive. Any `// ast-grep-ignore` comment today is either undocumented behavior or blocked entirely. SYNTHESIS §4.5 + A4 T2.4 flag the need for a documented, auditable suppression surface.
5. **11 of 16 language rule directories empty** (A2 §D7-1, SYNTHESIS §4.5). `.moai/config/astgrep-rules/{cpp,csharp,elixir,flutter,java,javascript,kotlin,php,python,r,ruby,rust,scala,swift,typescript}/` contain only `.gitkeep`. CLAUDE.local.md §15 HARD rule requires 16-language neutrality for templates; the installed project is allowed to be sparse, but template deployment currently seeds nothing beyond the go/ and security/ directories.

SPEC-UTIL-002 resolves items (1)-(3) via IMP-V3U-004/006/019, item (4) via D6 suppression policy, and item (5) partially via D7 (5 of 11 remaining languages — Ruby/PHP/Elixir/C#/Kotlin).

---

## 2. Current Implementation Analysis

### 2.1 Rule struct — metadata/note fields missing

File: `internal/astgrep/models.go:72-79`

```go
type Rule struct {
    ID       string `json:"id" yaml:"id"`
    Language string `json:"language" yaml:"language"`
    Severity string `json:"severity" yaml:"severity"`
    Message  string `json:"message" yaml:"message"`
    Pattern  string `json:"pattern" yaml:"pattern"`
    Fix      string `json:"fix,omitempty" yaml:"fix,omitempty"`
}
```

`Rule` is missing both `Note` and `Metadata` fields. YAML payloads like `security/injection.yml:8-13`:

```yaml
metadata:
  owasp: "A03:2021 - Injection"
  cwe: "CWE-89"
note: "사용자 입력을 SQL 쿼리에 직접 삽입하지 마세요..."
```

are parsed successfully by yaml.v3 (unknown fields silently ignored) and discarded.

### 2.2 Finding propagation drops Rule metadata

File: `internal/astgrep/scanner.go:340-380`

```go
for _, rule := range rules {
    ...
    findings, err := s.runSingleRule(ctx, binary, rule, path, timeout)
    ...
    for i := range findings {
        if findings[i].RuleID == "" { findings[i].RuleID = rule.ID }
        if findings[i].Severity == "" { findings[i].Severity = rule.Severity }
        if findings[i].Message == "" { findings[i].Message = rule.Message }
        findings[i].Language = rule.Language
    }
    allFindings = append(allFindings, findings...)
}
```

`Finding` (scanner.go:43-73) already has `Note string` and `Metadata map[string]string` fields. The propagation loop never copies `rule.Note` or `rule.Metadata` to the finding because those fields don't exist on `Rule`. Adding them unlocks the propagation with two additional lines.

### 2.3 ScanMultiple unbounded goroutines

File: `internal/hook/security/ast_grep.go:172-213`

```go
// @MX:WARN: [AUTO] Parallel file scan using goroutines...
// @MX:REASON: sync.WaitGroup + goroutine pattern, goroutines already started continue...
func (s *astGrepScanner) ScanMultiple(ctx context.Context, filePaths []string, ...) {
    ...
    for i, fp := range filePaths {
        wg.Add(1)
        go func(idx int, path string) {  // ← No semaphore
            defer wg.Done()
            result, err := s.Scan(ctx, path, configPath)
            ...
        }(i, fp)
    }
    wg.Wait()
    ...
}
```

The existing `@MX:WARN` tag documents the risk but no semaphore or worker pool exists. SYNTHESIS §2.1 REC-02 specifies the fix.

### 2.4 detectSGVersion placeholder

File: `internal/cli/astgrep.go:192-198`

```go
func detectSGVersion() string {
    cfg := &astgrep.ScannerConfig{SGBinary: "sg"}
    s := astgrep.NewScanner(cfg)
    _ = s // 미래 구현을 위한 플레이스홀더
    return "unknown"
}
```

The comment explicitly acknowledges this is unimplemented. The SARIF generator at `cli/astgrep.go:181-186` consumes the return value as `tool.driver.version`.

### 2.5 Rule directory coverage

Current state (`find .moai/config/astgrep-rules -type f`):

- `go/` — 5 rule files (idioms, resource-safety, error-handling, hardcoding, concurrency)
- `security/` — 4 rule files (secrets, injection, web, crypto)
- `cpp, csharp, elixir, flutter, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript` — only `.gitkeep`, 15 dead directories

v2.14 D7 scope: seed Ruby/PHP/Elixir/C#/Kotlin (5 languages × 3 rules = 15 YAML files). Remaining 6 (Java, C++, Scala, R, Flutter, Swift) stay empty until v2.15+ or community contribution.

### 2.6 Suppression policy — no sanctioned surface

Search for `ast-grep-ignore` or `ast-grep: disable` string literals across `internal/` returns no hits. There is no documented suppression mechanism, lint rule, or CI check. A4 T2.4 rec-3 flags `no-suppress-all` as a 2026 best practice; D6 adopts an explicit middle path — allow suppression IFF paired with a `// @MX:REASON <rationale>` comment on the adjacent line.

---

## 3. Target State

### 3.1 Rule struct additions

```go
type Rule struct {
    ID       string            `json:"id" yaml:"id"`
    Language string            `json:"language" yaml:"language"`
    Severity string            `json:"severity" yaml:"severity"`
    Message  string            `json:"message" yaml:"message"`
    Pattern  string            `json:"pattern" yaml:"pattern"`
    Fix      string            `json:"fix,omitempty" yaml:"fix,omitempty"`
    Note     string            `json:"note,omitempty" yaml:"note,omitempty"`          // NEW
    Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`  // NEW
}
```

### 3.2 Finding propagation

`scanner.go:scanWithRules` loop augmented with:

```go
if findings[i].Note == "" && rule.Note != "" {
    findings[i].Note = rule.Note
}
if findings[i].Metadata == nil && rule.Metadata != nil {
    findings[i].Metadata = rule.Metadata
}
```

### 3.3 SARIF propagation

`internal/astgrep/sarif.go` extracts `finding.Metadata["owasp"]` and `finding.Metadata["cwe"]` into `properties.tags` and `ruleProperties.security-severity` per SARIF 2.1.0 §3.52 Property Bags. Additive — no existing field renamed.

### 3.4 ScanMultiple bounded semaphore

```go
maxConcurrent := runtime.NumCPU() * 2
sem := make(chan struct{}, maxConcurrent)
for i, fp := range filePaths {
    wg.Add(1)
    go func(idx int, path string) {
        sem <- struct{}{}
        defer func() { <-sem }()
        defer wg.Done()
        result, err := s.Scan(ctx, path, configPath)
        ...
    }(i, fp)
}
```

### 3.5 detectSGVersion real implementation

```go
var (
    sgVersionOnce  sync.Once
    sgVersionValue string
)

func detectSGVersion() string {
    sgVersionOnce.Do(func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        cmd := exec.CommandContext(ctx, "sg", "--version")
        out, err := cmd.Output()
        if err != nil {
            sgVersionValue = "unknown"
            return
        }
        sgVersionValue = strings.TrimSpace(string(out))
    })
    return sgVersionValue
}
```

Cached via `sync.Once` — single `sg --version` invocation per process.

### 3.6 Suppression policy

Allowed form:

```go
// ast-grep-ignore
// @MX:REASON rule does not account for generated protobuf stub pattern
buf := getUnsafeBuffer()
```

Rejected forms (emit `SUPPRESSION_WITHOUT_REASON`, exit code 2):

```go
// ast-grep-ignore
buf := getUnsafeBuffer()  // missing @MX:REASON

// ast-grep-ignore
// some unrelated comment
buf := getUnsafeBuffer()  // @MX:REASON more than 1 line away
```

Lint location: `internal/hook/quality/astgrep_gate.go` gains a `checkSuppressionPairing(filePath) []SuppressionViolation` helper. Called after the scan completes; violations emit structured error with file:line citation.

### 3.7 5-language rule directories

```
.moai/config/astgrep-rules/
├── ruby/
│   ├── unused-var.yml
│   ├── null-deref.yml
│   └── todo-marker.yml
├── php/
│   ├── unused-var.yml
│   ├── null-deref.yml
│   └── todo-marker.yml
├── elixir/
│   ├── unused-var.yml
│   ├── nil-match.yml       # Elixir idiom: pattern-match on nil
│   └── todo-marker.yml
├── csharp/
│   ├── unused-var.yml
│   ├── null-deref.yml
│   └── todo-marker.yml
└── kotlin/
    ├── unused-var.yml
    ├── null-deref.yml       # uses !! operator detection
    └── todo-marker.yml
```

Each YAML carries `metadata: {owasp: "CWE-...", cwe: "CWE-..."}` exercising the IMP-V3U-004 propagation path.

---

## 4. External Research (A4 T2)

### 4.1 ast-grep 0.42.1 state (2026-04-04)

- Dart support restored (was removed in 0.41.x)
- Parameterized utils available (`function-with-high-fan-in($N)` shared fragments)
- Built-in rule `no-suppress-all` prevents silent `// ast-grep: disable`
- LSP deadlock fix on change events
- ESQuery-style selectors (`:nth-child`, `:is`, `:not`, `:has`)

moai currently runs 0.40.5 (A2 executive summary). v2.14.0 pins minimum to ≥ 0.40 (current production floor) without mandating 0.42.x — upgrade cadence documented in `.claude/rules/moai/core/lsp-client.md` analog for ast-grep (deferred to Tier-2 / v2.15).

### 4.2 Semgrep commercial license tightening (context)

A4 T2.2 confirms Semgrep license tightening Dec 2024 → Opengrep fork Jan 2025; CodeQL repackage behind GHAS April 2025. ast-grep Apache 2.0 remains the de-risked strategic choice. No action item for v2.14; documented as insurance for the engine selection.

### 4.3 Rule metadata propagation — SARIF 2.1.0 spec

OASIS SARIF 2.1.0 §3.52 (Property Bags) allows arbitrary key-value pairs under `properties`. OWASP and CWE mappings are conventionally stored as:

```json
{
  "ruleId": "sql-injection",
  "properties": {
    "tags": ["security", "external/cwe/cwe-89", "external/owasp/a03"],
    "security-severity": "9.8"
  }
}
```

GitHub Code Scanning consumers parse `tags[]` for `external/cwe/*` patterns to populate CWE cross-references. Additive — no existing consumer breaks if new tags appear.

### 4.4 ast-grep 0.42 no-suppress-all rule

A4 T2.1 rec-3: adopt `no-suppress-all` built-in rule. D6 is a softer middle path — allow targeted suppression but require documented rationale via `@MX:REASON`. This keeps the escape hatch open for legitimate false positives while preventing silent coverage loss.

Aligns with moai's existing `@MX` protocol: WARN and ANCHOR tags already require `@MX:REASON` per `.claude/rules/moai/workflow/mx-tag-protocol.md`. Extending the same convention to ast-grep suppressions keeps the mental model consistent.

### 4.5 Parameterized utils (v0.42 feature)

Not adopted in v2.14 (scope: 5 new languages × 3 foundational rules each). The 15 new YAML files use plain patterns, not parameterized utils, to keep the surface simple and downstream-compatible with ast-grep ≥ 0.40. Parameterized utils candidate for v2.17 workflow refactor if rule catalog grows.

---

## 5. Rule Catalog Design — 15 New YAML Files

Each rule follows a uniform template:

```yaml
id: <lang>-<name>
language: <tree-sitter-language-name>
severity: warning | error | info
message: "<user-facing message>"
pattern: |
  <ast-grep pattern>
note: "<educational note, native language or English>"
metadata:
  owasp: "<OWASP category or 'N/A'>"
  cwe: "CWE-<id>"
```

### 5.1 Ruby rules

**ruby-unused-var** (severity: info, CWE-561):

```yaml
id: ruby-unused-var
language: ruby
severity: info
message: "할당 후 사용되지 않는 변수입니다. `_` 접두사 사용 고려."
pattern: |
  $VAR = $EXPR
note: "지역 변수 할당 후 미사용. Ruby idiom: `_unused = expr` or omit."
metadata:
  owasp: "N/A"
  cwe: "CWE-561"
```

Note: Ruby ast-grep pattern syntax limits — in practice this rule requires Ruby-specific context matching; seeded as a foundational template. False positive rate acceptable for `severity: info`.

**ruby-null-deref** (severity: warning, CWE-476):

```yaml
id: ruby-nil-method-call
language: ruby
severity: warning
message: "nil 가능 객체에 직접 메서드 호출. safe navigation(`&.`) 사용 고려."
pattern: |
  $OBJ.$METHOD($$$ARGS)
note: "Ruby 2.3+: safe navigation operator `&.` avoids NoMethodError on nil."
metadata:
  owasp: "A05:2021"
  cwe: "CWE-476"
```

**ruby-todo-marker** (severity: info, no CWE):

```yaml
id: ruby-todo-marker
language: ruby
severity: info
message: "TODO 주석에 이슈 링크 또는 담당자가 누락되었습니다."
pattern: |
  # TODO$$$
note: "moai @MX:TODO 규약 참고 — 담당자/이슈 링크 필수."
metadata:
  owasp: "N/A"
  cwe: "CWE-1059"
```

### 5.2 PHP rules

**php-unused-var** (CWE-561), **php-null-deref** (CWE-476 — targets `$obj->method()` on nullable), **php-todo-marker** (CWE-1059). Pattern examples:

```yaml
# php-null-deref
pattern: |
  $OBJ->$METHOD($$$ARGS)
note: "PHP 8+: `?->` null-safe operator로 대체 가능."
```

### 5.3 Elixir rules

**elixir-unused-var** (CWE-561), **elixir-nil-match** (CWE-476 — Elixir idiom: detect `case` without `nil -> ...` clause is non-trivial; foundational rule flags `$VAR.$FIELD` direct field access on potentially-nil map), **elixir-todo-marker** (CWE-1059).

```yaml
# elixir-nil-match
id: elixir-unsafe-map-access
language: elixir
severity: warning
message: "맵에서 직접 필드 접근. Access.get/2 또는 pattern match 사용 고려."
pattern: |
  $MAP.$FIELD
note: "Elixir: `Map.get(map, :key)` or pattern match handles missing keys gracefully."
metadata:
  owasp: "A05:2021"
  cwe: "CWE-476"
```

### 5.4 C# rules

**csharp-unused-var** (CWE-561), **csharp-null-deref** (CWE-476 — targets `.$Method()` without null-conditional `?.`), **csharp-todo-marker** (CWE-1059).

```yaml
# csharp-null-deref
id: csharp-unsafe-null-method
language: csharp
severity: warning
message: "Null 가능 객체 메서드 호출. `?.` null-conditional operator 사용 고려."
pattern: |
  $OBJ.$METHOD($$$ARGS)
note: "C# 6+: `obj?.Method()` returns null instead of throwing NullReferenceException."
metadata:
  owasp: "A05:2021"
  cwe: "CWE-476"
```

### 5.5 Kotlin rules

**kotlin-unused-var** (CWE-561), **kotlin-null-deref** (CWE-476 — Kotlin-specific: detects `!!` non-null assertion operator), **kotlin-todo-marker** (CWE-1059).

```yaml
# kotlin-null-deref
id: kotlin-force-unwrap
language: kotlin
severity: warning
message: "`!!` 연산자는 NullPointerException을 유발할 수 있습니다."
pattern: |
  $EXPR!!
note: "Kotlin nullable 타입 안전 처리: `?.`, `?:`, 또는 `let`/`run` 블록 사용 권장."
metadata:
  owasp: "A05:2021"
  cwe: "CWE-476"
```

### 5.6 Per-language fixture strategy

Test fixtures under `internal/astgrep/testdata/fixtures/<lang>/` contain:

- `valid.<ext>` — code that does NOT match any rule (expected scan: 0 findings)
- `violation.<ext>` — code that matches one of the 3 rules (expected: 1+ findings)
- `suppressed.<ext>` — code that uses `// ast-grep-ignore` + `// @MX:REASON` correctly

Smoke test `internal/astgrep/rule_seed_test.go` invokes `sg scan` against fixtures and asserts:

- violation count matches golden
- `Finding.Metadata["cwe"]` is populated from rule YAML
- suppressed fixture produces zero findings

---

## 6. Risk Assessment

### 6.1 ast-grep version drift

**Risk**: seeded rules use plain pattern syntax (not parameterized utils, not ESQuery selectors). Pattern syntax is stable since ast-grep 0.20.x. Minimum version ≥ 0.40 documented in `internal/astgrep/doc.go` and `.claude/rules/moai/core/lsp-client.md` analog.

**Mitigation**: CI smoke test invokes `sg --version` and fails if < 0.40. `detectSGVersion()` implementation (IMP-V3U-019) records the actual version in SARIF for production observability.

### 6.2 Rule false positive rate

**Risk**: foundational patterns (unused-var, null-deref, todo-marker) are inherently prone to false positives across any language. Ruby's `ruby-unused-var` pattern `$VAR = $EXPR` matches every assignment; this is why severity is `info` not `warning`.

**Mitigation**:

- Severity calibration — `info` for patterns with expected FP rate > 20%, `warning` otherwise.
- Fixture-based characterization tests pin current behavior so future FP reductions are visible.
- Suppression policy (D6) gives operators an auditable escape hatch.
- A2 §D3-2 flags FP risk as a known-unknown; v2.14 seeding is a starting point, not final catalog.

### 6.3 5 new languages without production testing footprint

**Risk**: moai-adk-go itself is written in Go. Seeded rules for Ruby/PHP/Elixir/C#/Kotlin have no dogfooding loop — no internal project to surface bugs.

**Mitigation**:

- Golden fixture tests provide characterization coverage per rule per language (18 fixture pairs: 6 per language × 3 languages).
- Community contributions documented as the path to rule maturity (CLAUDE.local.md §15 flutter/dart canonical name precedent).
- SARIF output mandate: every finding carries CWE/OWASP metadata, enabling external validation pipelines.

### 6.4 Suppression policy — bypass by misuse

**Risk**: a user could write `// @MX:REASON todo` and suppress a legitimate violation. The lint rule only validates presence, not quality.

**Mitigation**:

- `@MX:REASON` content is surfaced in the SARIF `properties.suppressionReason` field for audit.
- Code review is the second-line defense (GitHub PR reviewers see both the `// ast-grep-ignore` and the `// @MX:REASON` text).
- Consistent with moai's existing `@MX:WARN`/`@MX:ANCHOR` REASON convention — this is not a new trust model, just an extension.

### 6.5 CI performance impact

**Risk**: adding 15 new rule files × 5 new languages increases scan time when `scanWithRules` runs on all rules.

**Mitigation**:

- Scanner iterates only over rules matching the file's language (via `rule.Language` field). A Go file scan does not invoke Ruby rules.
- Bounded semaphore (IMP-V3U-006) caps concurrent subprocesses at `runtime.NumCPU() * 2`.
- sgconfig.yml path (single `sg scan` invocation) is unchanged — seeded rules loaded once and filtered by sg itself.

---

## 7. References

### File citations (source code)

- `internal/astgrep/models.go:72-79` — `Rule` struct missing Metadata/Note
- `internal/astgrep/scanner.go:43-73` — `Finding` struct has Metadata/Note fields
- `internal/astgrep/scanner.go:340-380` — `scanWithRules` metadata propagation loop (incomplete)
- `internal/astgrep/scanner.go:382-409` — `parseSGFindings` 0→1 indexed conversion
- `internal/astgrep/sarif.go:197-206` — `toFileURI` (separate concern, out of scope for UTIL-002)
- `internal/cli/astgrep.go:192-198` — `detectSGVersion()` placeholder
- `internal/cli/astgrep.go:180-186` — SARIF version consumer
- `internal/hook/security/ast_grep.go:172-213` — `ScanMultiple` unbounded goroutines
- `internal/hook/security/ast_grep.go:331-347` — `supportedLanguages` (separate concern — Flutter/R missing, out of scope for UTIL-002, listed for future UTIL-004)
- `internal/hook/quality/astgrep_gate.go:20-58` — `RunAstGrepGateV2` (suppression lint injection point)
- `.moai/config/astgrep-rules/sgconfig.yml:1-24` — 16-language declaration

### Audit citations

- A2 — `.moai/design/utility-review/a2-astgrep-audit.md` — full audit, health 6/10
- A2 §D2-2 REC-01 — Rule.Metadata/Note missing (IMP-V3U-004)
- A2 §D5-3 REC-02 — ScanMultiple unbounded goroutines (IMP-V3U-006)
- A2 §D1-3 REC-03 — detectSGVersion placeholder (IMP-V3U-019 pulled into v2.14)
- A2 §D7-1 — 11 empty language directories
- SYNTHESIS §2.1 — ast-grep critical bugs summary
- SYNTHESIS §4.4 — version pinning / detection gaps
- SYNTHESIS §4.5 — 16-language neutrality gap
- A4 §T2.1 — ast-grep 0.42.1 state (2026-04-04)
- A4 §T2.2 — Semgrep commercial risk context
- A4 §T2.4 rec-3 — no-suppress-all built-in rule (D6 middle path)

### Decision references

- `docs/design/v2.14.0-release-plan.md` §0.5 D6 — suppression policy
- `docs/design/v2.14.0-release-plan.md` §0.5 D7 — 5-language rule seeding scope
- `docs/design/v2.14.0-release-plan.md` §2.2 SPEC-UTIL-002 — expanded scope breakdown

### Protocol / specification references

- OASIS SARIF 2.1.0 §3.52 — Property Bags (additive metadata surface)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — `@MX:REASON` convention (suppression pairing)
- CLAUDE.local.md §15 — 16-language neutrality HARD rule

### Related SPECs

- SPEC-UTIL-001 (parallel, v2.14) — `@MX:REASON` enforcement standard (suppression policy relies on same convention)
- SPEC-UTIL-003 (parallel, v2.14) — no file overlap, parallel Phase 3 execution
- SPEC-V3R2-EXT-* (v3R2 planning track) — possible future owner of rule catalog expansion

---

Version: 0.1.0
Status: research-complete
Last Updated: 2026-04-24
Author: Wave v2.14 SPEC Writer
