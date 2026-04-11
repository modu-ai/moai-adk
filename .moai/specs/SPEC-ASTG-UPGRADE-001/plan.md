# SPEC-ASTG-UPGRADE-001: Implementation Plan

## Phase Breakdown

### Phase 1: Directory Structure + Go Rule Library

**Goal**: Create the organized rule library structure and populate Go rules (15 rules total).

**Files**:
- `.moai/config/astgrep-rules/sgconfig.yml` (new)
- `.moai/config/astgrep-rules/go/concurrency.yml` (new, 4 rules)
- `.moai/config/astgrep-rules/go/error-handling.yml` (new, 3 rules)
- `.moai/config/astgrep-rules/go/resource-safety.yml` (new, 2 rules)
- `.moai/config/astgrep-rules/go/idioms.yml` (new, 3 rules)
- `.moai/config/astgrep-rules/go/hardcoding.yml` (migrate existing 3 rules)
- Delete legacy `.moai/config/astgrep-rules/go-hardcoding.yml` (superseded)

**Estimated LOC**: +450 YAML, -60 YAML (migration)

### Phase 2: OWASP Security Rule Library

**Goal**: Create 10 OWASP-aligned security rules.

**Files**:
- `.moai/config/astgrep-rules/security/injection.yml` (3 rules: SQL, command, path traversal)
- `.moai/config/astgrep-rules/security/crypto.yml` (2 rules: weak hash, TLS bypass)
- `.moai/config/astgrep-rules/security/secrets.yml` (2 rules: API key, JWT secret)
- `.moai/config/astgrep-rules/security/web.yml` (3 rules: template injection, log secrets, CSRF)

**Estimated LOC**: +300 YAML

### Phase 3: Unified Go Scanner

**Goal**: Replace split implementations with single `scanner.go`.

**Files**:
- `internal/astgrep/scanner.go` (new, ~400 LOC): `type Scanner struct`, `Scan(ctx, path) []Finding`, `LoadRulesFromDir(dir) error`
- `internal/astgrep/rules.go` (modify): extend `RuleLoader` for recursive subdirectory loading
- `internal/astgrep/scanner_test.go` (new, ~300 LOC): table-driven tests covering nil config, missing binary, empty rules, real scans
- `internal/hook/quality/astgrep_gate.go` (modify): call unified `Scanner` instead of `RunAstGrepGate`
- `internal/hook/security/ast_grep.go` (delete or stub): functionality absorbed by unified scanner

**Estimated LOC**: +700 Go, -400 Go (deletions)

### Phase 4: SARIF Output

**Goal**: GitHub code scanning integration.

**Files**:
- `internal/astgrep/sarif.go` (new, ~250 LOC): `func (s *Scanner) ToSARIF(findings []Finding) ([]byte, error)`
- `internal/astgrep/sarif_test.go` (new, ~150 LOC): validate SARIF 2.1.0 schema compliance

**Estimated LOC**: +400 Go

### Phase 5: CLI Subcommand

**Goal**: `moai ast-grep <path>` user-facing command.

**Files**:
- `internal/cli/astgrep.go` (new, ~200 LOC): Cobra command with `--format`, `--lang`, `--severity`, `--dry` flags
- `internal/cli/astgrep_test.go` (new, ~200 LOC)
- `internal/cli/deps.go` (modify): register astgrep subcommand in root Dependencies wiring

**Estimated LOC**: +400 Go, +5 Go (wiring)

### Phase 6: Hook Integration Cleanup

**Goal**: Replace the split scanner call sites.

**Files**:
- `internal/hook/post_tool.go` (modify): replace `SGAnalyzer` with unified `Scanner`
- `internal/hook/quality/gate.go` (modify): replace `RunAstGrepGate` with unified `Scanner.Scan`
- `internal/hook/security/ast_grep.go` (delete)
- All test files touching deleted scanner: update to use unified Scanner API

**Estimated LOC**: -200 Go (deletions), +50 Go (adapter calls)

### Phase 7: Documentation + CI Example

**Goal**: User-facing documentation and GitHub Actions workflow.

**Files**:
- `.claude/skills/moai-tool-ast-grep/SKILL.md` (update): document new rule library structure and CLI
- `.github/workflows/ast-grep-scan.yml.example` (new): reference SARIF upload workflow
- `CHANGELOG.md` (update): document v2.11.0 ast-grep upgrade

**Estimated LOC**: +100 markdown, +50 YAML

---

## Total Estimate

| Phase | Added LOC | Removed LOC | Net |
|-------|-----------|-------------|-----|
| 1 | 450 YAML | 60 YAML | +390 |
| 2 | 300 YAML | 0 | +300 |
| 3 | 700 Go | 400 Go | +300 |
| 4 | 400 Go | 0 | +400 |
| 5 | 405 Go | 0 | +405 |
| 6 | 50 Go | 200 Go | -150 |
| 7 | 150 docs | 0 | +150 |
| **Total** | **~2455** | **~660** | **+1795** |

---

## Dependencies

- **Hard prerequisite**: none (independent SPEC, can start immediately)
- **Soft prerequisite**: ast-grep v0.42.1+ installed in CI environment
- **Blocks**: SPEC-LSP-QGATE-004 (phase-aware gates expect unified Scanner API)

## Risks

| Risk | Mitigation |
|------|------------|
| sg CLI version drift | Pin to v0.42.1 minimum in README install hint; scanner gracefully handles missing binary |
| Rule false positives | Each rule has a `note` field explaining detection rationale; users can disable via `warn_only_mode` |
| Performance on large repos | ast-grep parallelizes internally; scanner applies `--skip-gitignore` by default |
| Breaking quality gate during migration | Phase 3 provides adapter that calls both old and new while tests run; cutover after Phase 6 verification |
