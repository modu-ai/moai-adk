# SPEC-SLQG-001 Research Document

## Research Summary

Date: 2026-04-11
Topic: Self-Learning Quality Guard (SLQG)

## 1. ast-grep Integration (As-Is)

### Infrastructure
- `internal/astgrep/analyzer.go`: SGAnalyzer with Scan/FindPattern/Replace APIs
- `internal/astgrep/rules.go`: RuleLoader with LoadFromFile/LoadFromDirectory/GetRulesForLanguage
- `internal/astgrep/models.go`: Match, ScanResult, ScanConfig, Rule structs

### Current Usage
- PostToolUse hook: observation-only scanning after Write/Edit (never blocks)
- Security scanner: `internal/hook/security/ast_grep.go` for vulnerability patterns
- Config: `.moai/config/sections/ralph.yaml` → `ralph.ast_grep.enabled: true`

### Extension Points
- RuleLoader.LoadFromDirectory() can load any YAML rules directory
- ScanConfig supports include/exclude patterns
- Handler Registry allows new PreToolUse handlers

### Gap
- No domain-specific rules (only security patterns)
- No blocking mode in PreToolUse for ast-grep violations
- No mechanism to auto-generate rules from bug fixes

## 2. Lessons Protocol (As-Is)

### Constitution Rules (moai-constitution.md:126-140)
- Store at `~/.claude/projects/{hash}/memory/lessons.md`
- Entry: category, incorrect pattern, correct approach, date
- Categories: architecture, testing, naming, workflow, security, performance
- Max 50 active, archive overflow
- Session start: scan for domain-matching patterns

### Current State
- lessons.md: 3 manual entries (2026-03-11 ~ 2026-04-03)
- No auto-capture mechanism implemented
- No domain matching algorithm
- NOT loaded in run.md context (Phase 1 context loading)
- NOT referenced in any agent definition or spawn prompt

### Gap
- Capture: No trigger when user corrects agent behavior
- Load: Not injected into agent context before implementation
- Match: No algorithm to filter relevant lessons by domain
- Feedback: No "save as lesson?" prompt flow

## 3. Quality Gate Hooks (As-Is)

### Architecture
- PreToolUse: 4 layers (blocklist → security patterns → quality gate → ast scan)
- Quality gate runs: vet → lint → test (language-specific toolchain)
- Handler Registry: Register(handler) → Dispatch(event, input)
- SecurityPolicy: MergeExtraPatterns for custom rules

### Current Checks
- golangci-lint: SA5011, standard Go linting
- File path patterns: deny/ask for secrets, config files
- Bash patterns: dangerous command blocking
- AST scan: security-only, observation-only

### Extension Points
- SecurityPolicy.MergeExtraPatterns() for custom deny/ask patterns
- Handler registration for custom PreToolUse checks
- Tool Registry for custom linter registration
- ast-grep rules directory loading

### Gap
- No domain-specific quality rules (hardcoding, duplication, etc.)
- ast-grep only runs post-tool (not pre-commit gate)
- No feedback loop: fix → learn → prevent cycle absent
