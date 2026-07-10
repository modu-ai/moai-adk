# Static Analysis Tools Integration

> Module: linter, security scanner, and type checker integration for comprehensive static analysis
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Intermediate
> Time: 15+ minutes
> Dependencies: the project's static analysis tools, subprocess/JSON

## Quick Reference

### Tool Roles (language-neutral)

Linter — code quality and style checking:
- Comprehensive analysis of source code
- Detects bugs, code smells, and style violations
- Provides code ratings and detailed reports
- Configurable with project-specific rules

Style/Format checker — style guide enforcement:
- Fast and lightweight style checking
- Enforces the language's style guide
- Highly customizable with plugins

Security scanner — vulnerability scanning:
- Finds common security issues
- Configurable severity levels
- Integrates with security best practices

Type checker — type validation:
- Catches type errors before runtime
- Supports gradual typing (where the language supports it)
- Improves code reliability and documentation

Per-language tool map (populate the registry with the host language's tools):

| Role | Go | Python | JS/TS | Rust | Java |
|------|----|--------|-------|------|------|
| Linter/Format | gofmt, golangci-lint | ruff, pylint | eslint, prettier | rustfmt, clippy | checkstyle, spotless |
| Security | govulncheck | bandit | npm audit, eslint-plugin-security | cargo-audit | spotbugs |
| Type checker | go vet | mypy, pyright | tsc | cargo check | javac |

### Core Implementation

```text
class StaticAnalysisTools:
    tools = {
        # name -> runner; populate from the per-language table above
    }

    run_all_analyses(file_path):
        results = {}
        for (name, runner) in tools:
            try: results[name] = runner(file_path)
            except e: results[name] = { error: text(e) }
        return results
```

---

## Implementation Guide

### Linter Integration

```text
run_linter(file_path):
    try:
        result = exec([linter_cmd, file_path, "--format=json"], capture=stdout+stderr)
        if result.exit_code == 0: return { issues: [] }
        try:
            return { issues: parse_json(result.stdout), summary: parse_summary(result.stderr) }
        except JsonError:
            return { raw_output: result.stdout, raw_errors: result.stderr }
    except NotFound:
        return { error: linter_cmd + " not installed" }
```

Linter Configuration: each tool has its own config file (e.g. `.golangci.yml`, `pyproject.toml`/`.pylintrc`, `.eslintrc`, `clippy.toml`). Disable noisy rules and tune thresholds per project.

```ini
# Example shape (the exact file/format is tool-specific):
disable = no-docstring-rule, too-few-public-methods
[design]
max-args = 7
max-locals = 15
max-returns = 6
max-branches = 12
max-statements = 50
```

### Format / Style Checker Integration

```text
run_format_checker(file_path):
    try:
        result = exec([format_cmd, file_path, "--format=json"], capture=stdout+stderr)
        if result.exit_code == 0: return { issues: [] }
        issues = []
        for line in lines(result.stdout):
            if line is empty: continue
            (path, line_no, column, code, message) = parse(line)
            if path: issues.append({ path, line: line_no, column, code, message })
        return { issues }
    except NotFound:
        return { error: format_cmd + " not installed" }
```

Style Configuration: each formatter has its own config (e.g. `.golangci.yml`, `.flake8`/`pyproject.toml`, `.eslintrc`, `rustfmt.toml`). Typical knobs: max-line-length, exclude paths, ignore codes, max-complexity.

```ini
# Example shape (tool-specific):
max-line-length = 100
exclude = vcs-dirs, build-output, dependency-dirs
ignore = stylistic-disputes
max-complexity = 10
```

### Security Scanner Integration

```text
run_security_scanner(file_path):
    try:
        result = exec([scanner_cmd, "--format=json", file_path], capture=stdout+stderr)
        try: return parse_json(result.stdout)
        except JsonError: return { raw_output: result.stdout }
    except NotFound:
        return { error: scanner_cmd + " not installed" }
```

Security Configuration: each scanner has its own rule selection config. Configure excluded directories and the rule categories to enable (e.g. crypto, shell-injection, hardcoded secrets, insecure deserialization, weak SSL). See the scanner's own documentation for its rule taxonomy.

### Type Checker Integration

```text
run_type_checker(file_path):
    try:
        result = exec([typechecker_cmd, file_path, "--show-error-codes"], capture=stdout+stderr)
        issues = []
        for line in lines(result.stdout):
            # most type checkers emit "<path>:<line>:<col>: error: <message>"
            if line contains "error:":
                (path, line_no, message) = parse_error_line(line)
                if path and line_no: issues.append({ path, line: line_no, message })
        return { issues }
    except NotFound:
        return { error: typechecker_cmd + " not installed" }
```

Type Checker Configuration: each type checker has its own config (e.g. `mypy.ini`, `tsconfig.json`, `Cargo.toml` profile, `.editorconfig`). Typical knobs: target language version, strictness flags, per-path overrides, ignore-missing-imports for third-party modules.

---

## Tool-to-TRUST Category Mapping

```text
map_tool_to_trust_category(tool_name, issue_data):
    if tool_name is a security scanner: return TrustCategory.SAFETY
    if tool_name is a type checker:     return TrustCategory.TRUTHFULNESS
    if tool_name is a linter:
        msg = lowercase(issue_data.message default "")
        if any of ["security","injection","unsafe"] in msg: return TrustCategory.SAFETY
        if any of ["performance","inefficient"]   in msg: return TrustCategory.TIMELINESS
        return TrustCategory.USABILITY
    return TrustCategory.USABILITY
```

---

## Result Normalization

```text
convert_static_issues(static_results, file_path):
    issues = []
    for (tool_name, results) in static_results:
        if 'error' in results: continue
        for issue_data in results.issues default []:
            category = map_tool_to_trust_category(tool_name, issue_data)
            issues.append(CodeIssue(
                id=tool_name + "_" + len(issues),
                category=category,
                severity=map_severity(issue_data.severity default "medium"),
                issue_type=map_issue_type(tool_name, issue_data),
                title=capitalize(tool_name) + ": " + (issue_data.message default "Unknown issue"),
                description=issue_data.message default "Static analysis issue",
                file_path=file_path,
                line_number=issue_data.line default 0,
                column_number=issue_data.column default 0,
                code_snippet=issue_data.code_snippet default "",
                suggested_fix=get_suggested_fix(tool_name, issue_data),
                confidence=0.8,
                rule_violated=issue_data.code default "",
                external_reference=tool_name + " documentation"))
    return issues
```

---

## Error Handling

```text
class StaticAnalysisTools(fallback_to_defaults = true):
    tools = { ... }   # name -> safe-runner
    fallback_to_defaults = fallback_to_defaults

    run_linter_safe(file_path):
        try: return run_linter(file_path)
        except NotFound:
            if fallback_to_defaults: return basic_style_checks(file_path)
            return { error: "linter not installed and no fallback available" }
        except e:
            return { error: "linter execution failed: " + text(e) }
```

---

## Best Practices

1. Tool Availability: Ensure all tools are installed in development environment
2. Configuration: Use configuration files for project-specific rules
3. CI Integration: Run static analysis in CI/CD pipeline
4. Incremental Adoption: Start with subset of tools, gradually add more
5. Error Handling: Implement fallback mechanisms for missing tools
6. Performance: Cache results to avoid redundant analysis
7. Team Consistency: Use same tool versions across team
8. Regular Updates: Keep tools updated for latest checks and fixes

---

## Related Modules

- [TRUST 5 Validation](./trust5-validation.md): Category mapping and scoring
- [Security Analysis](./security-analysis.md): Enhanced security detection
- [Quality Metrics](./quality-metrics.md): Code quality analysis

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/static-analysis.md`
