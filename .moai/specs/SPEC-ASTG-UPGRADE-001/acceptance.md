# SPEC-ASTG-UPGRADE-001: Acceptance Criteria

## Functional Acceptance

### AC1: Rule Library Structure
- [ ] `.moai/config/astgrep-rules/sgconfig.yml` exists and declares `ruleDirs: [go, security, python, typescript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift, utils]`
- [ ] `.moai/config/astgrep-rules/go/` contains 5 files (concurrency.yml, error-handling.yml, resource-safety.yml, idioms.yml, hardcoding.yml)
- [ ] `.moai/config/astgrep-rules/security/` contains 4 files (injection.yml, crypto.yml, secrets.yml, web.yml)
- [ ] Legacy `go-hardcoding.yml` is deleted (moved into `go/hardcoding.yml`)
- [ ] Total rule count: ≥ 25 (15 Go + 10 security)

### AC2: Unified Scanner API
- [ ] `internal/astgrep/scanner.go` exists with `type Scanner struct` and `func (s *Scanner) Scan(ctx context.Context, path string) ([]Finding, error)` methods
- [ ] `scanner.NewScanner(cfg *Config)` returns a configured scanner
- [ ] `RuleLoader.LoadFromDir` recursively walks subdirectories
- [ ] `internal/hook/security/ast_grep.go` is deleted; functionality absorbed
- [ ] All existing scanner call sites in `internal/hook/` use `scanner.Scanner` instead of `RunAstGrepGate` or `SGAnalyzer`

### AC3: Graceful Degradation
- [ ] When `sg` CLI binary is not in PATH, `Scanner.Scan` returns `([]Finding{}, nil)` with a warn-level log (not an error)
- [ ] When the rules directory is empty, `Scanner.Scan` returns `([]Finding{}, nil)` without error
- [ ] When an individual rule fails to parse, the scanner logs the error and continues with remaining rules

### AC4: CLI Subcommand
- [ ] `moai ast-grep ./path` exits with code 0 when no error-severity findings exist
- [ ] `moai ast-grep ./path` exits with code 1 when error-severity findings are found
- [ ] `moai ast-grep --format=sarif ./path` produces SARIF 2.1.0 conformant JSON
- [ ] `moai ast-grep --format=json ./path` produces machine-readable finding list
- [ ] `moai ast-grep --lang=go ./path` scans only Go files
- [ ] `moai ast-grep --severity=error ./path` reports only error-level findings
- [ ] `moai ast-grep --dry ./path` prints the rules that would apply without executing

### AC5: SARIF Output
- [ ] SARIF output validates against SARIF 2.1.0 JSON schema
- [ ] Each finding maps CWE/OWASP metadata to SARIF `properties`
- [ ] `tool.driver.name` is `"moai-ast-grep"` and `tool.driver.version` reflects ast-grep CLI version
- [ ] `results[].level` maps `severity` (error→error, warning→warning, info→note)
- [ ] SARIF file uploads successfully via `github/codeql-action/upload-sarif@v2` in CI test

### AC6: Hook Integration
- [ ] Pre-commit (PreToolUse Bash git commit) hook blocks commit on error-severity findings
- [ ] PostToolUse hook (Write/Edit) emits findings as `systemMessage` in conversation
- [ ] Hook configuration in `.claude/settings.json` passes validation
- [ ] `warn_only_mode: true` in `quality.yaml` ast_grep_gate section disables blocking

### AC7: Language Neutrality (Section 22)
- [ ] Directory structure includes placeholder entries for all 16 MoAI-supported languages (empty subdirectories or `.gitkeep` files acceptable)
- [ ] No Go-specific hardcoding in scanner code (scanner treats languages uniformly based on rule metadata)
- [ ] Documentation (SKILL.md) lists all 16 languages in alphabetical order
- [ ] `sgconfig.yml` does not prioritize Go over other languages

---

## Quality Acceptance (TRUST 5)

### Tested
- [ ] `internal/astgrep/scanner_test.go` provides ≥ 85% coverage
- [ ] `internal/astgrep/sarif_test.go` validates schema compliance and round-trip JSON
- [ ] `internal/cli/astgrep_test.go` tests all flag combinations
- [ ] Integration test exercises real `sg` binary when available (skipped gracefully when missing)
- [ ] Table-driven tests for each finding severity level

### Readable
- [ ] All exported types and functions have godoc comments
- [ ] Scanner public API has a doc.go explaining usage patterns
- [ ] Rule YAML files have inline comments explaining the pattern rationale

### Unified
- [ ] No duplicate scanner implementations in `internal/hook/`
- [ ] All code uses `Scanner.Scan` as the single entry point
- [ ] CLI command follows existing Cobra conventions in `internal/cli/`

### Secured
- [ ] Scanner does not execute user-provided shell commands (only invokes `sg` binary)
- [ ] CWE/OWASP mappings are preserved from rule metadata to SARIF output
- [ ] Security rule library covers OWASP Top 10 basics

### Trackable
- [ ] Every rule has a unique `id` matching regex `^[a-z0-9-]+$`
- [ ] Each rule `message` is concise (< 120 chars) and the `note` field provides context
- [ ] Commit messages reference SPEC-ASTG-UPGRADE-001 in scope field
- [ ] @MX:ANCHOR on `Scanner.Scan` (fan_in ≥ 3: quality gate, post-tool hook, CLI subcommand)

---

## Deliverable Checklist

- [ ] 9 YAML rule files created (5 Go + 4 security)
- [ ] 1 sgconfig.yml created
- [ ] 14 language placeholder directories
- [ ] 5 new Go source files (scanner.go, sarif.go, cli.go, scanner_test.go, sarif_test.go)
- [ ] 2 modified Go source files (rules.go for recursive loading, deps.go for CLI wiring)
- [ ] 2 modified hook files (post_tool.go, quality/gate.go)
- [ ] 1 deleted legacy file (hook/security/ast_grep.go or converted to adapter)
- [ ] 1 updated skill doc (moai-tool-ast-grep/SKILL.md)
- [ ] 1 example CI workflow (.github/workflows/ast-grep-scan.yml.example)
- [ ] CHANGELOG.md entry for v2.11.0
- [ ] `make build` passes
- [ ] `go test ./internal/astgrep/... ./internal/cli/... ./internal/hook/...` passes
- [ ] `go vet ./...` and `golangci-lint run` pass
- [ ] Test coverage ≥ 85% for `internal/astgrep/`
