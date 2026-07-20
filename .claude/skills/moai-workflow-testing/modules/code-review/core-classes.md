# Automated Code Review - Core Classes

> Sub-module: Core class implementations for automated code review
> Parent: [Automated Code Review](../automated-code-review.md)

The sketches below describe the shape of the review objects and orchestrator in language-neutral pseudo-code. Implement them in the host project's language using its native enums/records; file-discovery globs should use the project's actual source extensions.

## Enumerations and Data Classes

### TrustCategory Enum

```text
enum TrustCategory:
    TRUTHFULNESS   # code correctness and logic accuracy
    RELEVANCE      # code meets requirements and purpose
    USABILITY      # code is maintainable and understandable
    SAFETY         # code is secure and handles errors properly
    TIMELINESS     # code meets performance and delivery standards
```

### Severity Enum

```text
enum Severity:
    CRITICAL
    HIGH
    MEDIUM
    LOW
    INFO
```

### IssueType Enum

```text
enum IssueType:
    SYNTAX_ERROR
    LOGIC_ERROR
    SECURITY_VULNERABILITY
    PERFORMANCE_ISSUE
    CODE_SMELL
    STYLE_VIOLATION
    DOCUMENTATION_ISSUE
    TESTING_ISSUE
    TYPE_ERROR
    IMPORT_ISSUE
```

### CodeIssue Record

```text
record CodeIssue:
    id:                text
    category:          TrustCategory
    severity:          Severity
    issue_type:        IssueType
    title:             text
    description:       text
    file_path:         text
    line_number:       int
    column_number:     int
    code_snippet:      text
    suggested_fix:     text
    confidence:        float  # 0.0 to 1.0
    rule_violated:     text?  # default none
    external_reference:text?  # default none
    auto_fixable:      bool   # default false
    fix_diff:          text?  # default none
```

### FileReviewResult Record

```text
record FileReviewResult:
    file_path:         text
    issues:            List<CodeIssue>
    metrics:           Map<text, Any>
    trust_score:       float  # 0.0 to 1.0
    category_scores:   Map<TrustCategory, float>
    lines_of_code:     int
    complexity_metrics:Map<text, float>
    review_timestamp:  timestamp
```

### CodeReviewReport Record

```text
record CodeReviewReport:
    project_path:            text
    files_reviewed:          List<FileReviewResult>
    overall_trust_score:     float
    overall_category_scores: Map<TrustCategory, float>
    summary_metrics:         Map<text, Any>
    recommendations:         List<text>
    critical_issues:         List<CodeIssue>
    review_duration:         float
    docs_patterns_used:  List<text>
```

## AutomatedCodeReviewer Class

```text
class AutomatedCodeReviewer:
    docs
    docs_analyzer
    static_analyzer
    analysis_patterns = {}
    review_history = []

    review_codebase(project_path, include_patterns = none, exclude_patterns = none):
        start_time = now()
        analysis_patterns = docs_analyzer.load_analysis_patterns()
        files = find_files_to_review(project_path, include_patterns, exclude_patterns)
        file_results = [review_single_file(p) for p in files]
        return generate_comprehensive_report(project_path, file_results, now() - start_time)

    review_single_file(file_path):
        try:
            content = read(file_path)
        except e:
            return error_result(file_path, text(e))

        # Parse the syntax tree with the host language's parser
        try:
            tree = parse(content)
        except SyntaxError as e:
            return syntax_error_result(file_path, content, e)

        static_results  = static_analyzer.run_all_analyses(file_path)
        docs_issues = perform_docs_analysis(file_path, content, tree)
        custom_issues   = perform_custom_analysis(file_path, content, tree)

        all_issues = convert_static_issues(static_results, file_path) + docs_issues + custom_issues
        metrics       = calculate_file_metrics(content, tree)
        trust_scores  = calculate_trust_scores(all_issues, metrics)

        return FileReviewResult(
            file_path=file_path, issues=all_issues, metrics=metrics,
            trust_score=trust_scores.overall,
            category_scores=trust_scores.categories,
            lines_of_code=count_lines(content),
            complexity_metrics=calculate_complexity_metrics(content, tree),
            review_timestamp=now())

    find_files_to_review(project_path, include_patterns, exclude_patterns):
        # Default to the host language's source extensions (e.g. **/*.go, **/*.py,
        # **/*.ts, **/*.rs). Exclude build output, dependency dirs, generated code,
        # and (optionally) tests from the review pass.
        if include_patterns is none: include_patterns = default_source_globs()
        if exclude_patterns is none:
            exclude_patterns = ["build-output", "dependency-dirs", "vcs-dirs", "generated", "tests"]
        files = []
        for pattern in include_patterns:
            for path in glob(project_path, pattern):
                if path is a file and not excluded(path, exclude_patterns):
                    files.append(path)
        return sorted(files)
```

## DocumentationCodeAnalyzer Class

```text
class DocumentationCodeAnalyzer:
    docs
    analysis_patterns  = {}
    security_patterns  = {}
    performance_patterns = {}

    load_analysis_patterns(language = "python"):
        if docs is none: return default_analysis_patterns()
        try:
            security    = docs.get_library_docs("<security/semgrep>",
                            topic="security vulnerability detection patterns", tokens=4000)
            performance = docs.get_library_docs("<performance/profiling>",
                            topic="performance anti-patterns code analysis", tokens=3000)
            quality     = docs.get_library_docs("<code-quality/linters>",
                            topic="code quality best practices smells detection", tokens=4000)
            security_patterns    = security
            performance_patterns = performance
            return { security, performance, quality }
        except e:
            log("Failed to load Documentation patterns: " + e)
            return default_analysis_patterns()

    default_analysis_patterns():
        # Patterns are language-sensitive; the defaults below are illustrative shapes.
        # Populate per-language regexes that match the host language's idioms.
        return {
          security: {
            sql_injection:    ["query-execute with string concatenation", "format() inside execute"],
            command_injection:["shell-exec call", "subprocess call", "eval call"],
            path_traversal:   ["file-open with concatenation", "../ reference"]
          },
          performance: {
            inefficient_loops: ["index-iterated loop where a for-each suffices", "while on length"],
            memory_leaks:      ["unbounded growth", "global accumulation"]
          },
          quality: {
            long_functions:       { max_lines: 50 },
            complex_conditionals: { max_complexity: 10 },
            deep_nesting:         { max_depth: 4 }
          }
        }
```

## Related Sub-modules

- [Analysis Patterns](./analysis-patterns.md) - TRUST 5 analysis implementation
- [Tool Integration](./tool-integration.md) - Static analysis tool wrappers

---

Sub-module: `modules/code-review/core-classes.md`
