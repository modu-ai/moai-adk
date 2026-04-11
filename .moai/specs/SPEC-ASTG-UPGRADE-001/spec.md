---
id: SPEC-ASTG-UPGRADE-001
version: "1.0.0"
status: draft
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P1
issue_number: 0
phase: "Phase 3 - Quality Infrastructure"
module: "internal/astgrep/, internal/hook/quality/, .moai/config/astgrep-rules/"
estimated_loc: 1500
dependencies: []
lifecycle: spec-anchored
tags: ast-grep, quality, security, 16-languages, owasp, sarif
---

# SPEC-ASTG-UPGRADE-001: ast-grep Quality & Security Modernization

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft based on Phase 1 comparative research (ast-grep v0.42.1) |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-ASTG-UPGRADE-001 |
| Title | ast-grep Quality & Security Modernization |
| Status | Draft |
| Priority | P1 (Path C Primary, independent) |
| Research Source | C1/C3 reports (2026-04-11 comparative audit) |

---

## Problem Statement

moai-adk ships with only **3 ast-grep rules** (`go-no-raw-getenv`, `go-no-hardcoded-api-url`, `go-no-duplicate-coverage-threshold`) while claiming to enforce quality across 16 languages. Two separate scanners (`internal/hook/quality/astgrep_gate.go` and `internal/hook/security/ast_grep.go`) duplicate functionality. No SARIF output exists for GitHub code scanning. ast-grep v0.42.1 introduces parameterized utilities, injected-language scanning, and LSP mode — none of which moai-adk leverages.

**Gap analysis** (from Phase 1 research report R3):
- Rule count: **3 of 25+ recommended** (88% gap)
- Scanner unification: **separated** (quality_gate vs security_scanner)
- Output formats: **text + JSON only** (no SARIF for CI)
- Rule library structure: **flat directory**, no language or domain grouping
- Version: **pre-v0.42** (missing parameterized utilities, LSP mode)

---

## Goal

Ship a production-grade ast-grep integration that:

1. **Library**: 25+ rules (15 Go quality + 10 OWASP security), organized by domain
2. **Unified scanner**: single `internal/astgrep/scanner.go` replacing the two separate implementations
3. **CLI**: `moai ast-grep <path>` subcommand for on-demand scanning
4. **SARIF output**: GitHub code scanning integration
5. **Multi-language ready**: directory structure supports all 16 MoAI-supported languages
6. **ast-grep v0.42.1+**: leverages parameterized utilities where beneficial

---

## Requirements (EARS Format)

### Rule Library

**REQ-ASTG-UPG-001**: The system SHALL provide a structured rule library at `.moai/config/astgrep-rules/` with subdirectories per language and domain.

**REQ-ASTG-UPG-002**: The Go quality rule library SHALL contain at least 15 rules covering: concurrency (4 rules), error handling (3), resource management (2), type safety (3), and hardcoding prevention (3).

**REQ-ASTG-UPG-003**: The OWASP security rule library SHALL contain at least 10 rules covering: SQL injection, command injection, path traversal, hardcoded credentials, weak crypto, TLS bypass, template injection, logging secrets, CSRF, and XSS prevention.

**REQ-ASTG-UPG-004**: Each rule SHALL include `id`, `language`, `severity`, `message`, `note`, and (where applicable) `fix` or `transform` fields per ast-grep v0.42 schema.

**REQ-ASTG-UPG-005**: Rules tagged with CWE or OWASP identifiers SHALL populate the `metadata` field for SARIF integration.

### Unified Scanner

**REQ-ASTG-UPG-010**: The system SHALL provide a single `internal/astgrep/scanner.go` that replaces both `internal/hook/quality/astgrep_gate.go` and `internal/hook/security/ast_grep.go`.

**REQ-ASTG-UPG-011**: The unified scanner SHALL load rules from all subdirectories of `.moai/config/astgrep-rules/` recursively.

**REQ-ASTG-UPG-012**: When the `sg` CLI binary is not found in PATH, the scanner SHALL return a non-error skip result with a `warn_and_skip` message, never crash.

**REQ-ASTG-UPG-013**: The scanner SHALL classify findings by severity (`error`, `warning`, `info`) and return them in a canonical Finding struct shared across hook integration points.

### CLI Subcommand

**REQ-ASTG-UPG-020**: The system SHALL provide a `moai ast-grep <path>` subcommand that scans the given path with all loaded rules.

**REQ-ASTG-UPG-021**: The subcommand SHALL support flags `--format=text|json|sarif`, `--lang=<language>`, `--severity=error|warning|info`, and `--dry`.

**REQ-ASTG-UPG-022**: When `--format=sarif` is used, the output SHALL conform to SARIF 2.1.0 specification for upload to GitHub code scanning.

### Hook Integration

**REQ-ASTG-UPG-030**: The quality gate hook SHALL call the unified scanner and block commits on `error`-severity findings.

**REQ-ASTG-UPG-031**: The PostToolUse hook SHALL call the unified scanner on Write/Edit results and emit findings to the agent conversation as `systemMessage`.

### Language Neutrality (Section 22 compliance)

**REQ-ASTG-UPG-040**: The rule library SHALL be organized such that all 16 MoAI-supported languages are treated equally; the initial ship includes Go rules but the directory structure SHALL support adding Python, TypeScript, Rust, etc. rules without restructuring.

**REQ-ASTG-UPG-041**: The scanner SHALL NOT hardcode a language priority order; language selection SHALL be driven by the user's project marker files.

---

## Architecture

### Directory Structure

```
.moai/config/astgrep-rules/
├── sgconfig.yml                  # ast-grep config (ruleDirs declaration)
├── go/                           # Go-specific rules (initial ship: 15)
│   ├── concurrency.yml           # goroutine-context, mutex-deferred, defer-in-loop
│   ├── error-handling.yml        # error-ignored, error-wrapping, hardcoded-errors
│   ├── resource-safety.yml       # forgotten-close, map-ok-check
│   ├── idioms.yml                # interface-to-any, time-since, bytes-conversion
│   └── hardcoding.yml            # raw-getenv, api-url, duplicate-constants (existing)
├── security/                     # Cross-language OWASP (initial ship: 10)
│   ├── injection.yml             # sql, command, path-traversal
│   ├── crypto.yml                # weak-hash, tls-bypass
│   ├── secrets.yml               # hardcoded-api-key, jwt-secret
│   └── web.yml                   # template-injection, log-secrets, csrf
├── python/                       # placeholder (future)
├── typescript/                   # placeholder (future)
├── rust/                         # placeholder (future)
├── java/                         # placeholder (future)
├── ...                           # 16 total language dirs
└── utils/                        # Reusable rule fragments (v0.42 utilities)
```

### Go Package Refactor

```
internal/astgrep/
├── scanner.go         # NEW: unified scanner (replaces gate + security split)
├── analyzer.go        # KEEP: SGAnalyzer for file-level scanning
├── rules.go           # KEEP: RuleLoader (extend for recursive subdir loading)
├── sarif.go           # NEW: SARIF 2.1.0 output format
├── cli.go             # NEW: helper for moai ast-grep subcommand
└── analyzer_test.go, scanner_test.go, sarif_test.go
```

### CLI Command

```
internal/cli/
├── astgrep.go         # NEW: moai ast-grep subcommand (Cobra command)
└── astgrep_test.go    # NEW: table-driven tests for flags and output
```

---

## Non-Goals

- **LSP mode of ast-grep** (`sg lsp`): out of scope for this SPEC. Future enhancement in SPEC-LSP-CORE-002.
- **Auto-fix application in CI**: this SPEC provides the `fix` field in rules but the actual application policy is governed by SPEC-LSP-QGATE-004.
- **Python/TypeScript/Rust rule library**: the directory structure is created but initial shipment includes only Go and security rules.

---

## References

- Phase 1 research report R3 (ast-grep v0.42.1 landscape)
- CLAUDE.local.md Section 22 (Template Language Neutrality)
- [ast-grep official docs](https://ast-grep.github.io/)
- [SARIF 2.1.0 specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/)
- [ast-grep-essentials rule library](https://github.com/coderabbitai/ast-grep-essentials)
