# Context7 MCP Integration

> Module: Context7 integration patterns for real-time security and quality analysis
> Parent: [Automated Code Review](../automated-code-review.md)
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Context7 MCP client

## Quick Reference

### Context7 Integration Overview

Context7 MCP provides real-time access to:
- OWASP Top 10 security vulnerability patterns
- Semgrep security detection rules
- SonarQube / Sonar code quality standards
- Performance optimization libraries
- TRUST 5 validation frameworks

### Core Integration Pattern

```text
class Context7CodeAnalyzer:
    context7
    analysis_patterns    = {}
    security_patterns    = {}
    performance_patterns = {}

    load_analysis_patterns(language = "python"):
        if context7 is none: return default_analysis_patterns()
        try:
            security    = context7.get_library_docs("<security/semgrep>",
                            topic="security vulnerability detection patterns", tokens=4000)
            performance = context7.get_library_docs("<performance/profiling>",
                            topic="performance anti-patterns code analysis", tokens=3000)
            quality     = context7.get_library_docs("<code-quality/sonarqube>",
                            topic="code quality best practices smells detection", tokens=4000)
            trust       = context7.get_library_docs("<code-review/trust-framework>",
                            topic="TRUST 5 code validation framework patterns", tokens=3000)
            security_patterns    = security
            performance_patterns = performance
            return { security, performance, quality, trust }
        except e:
            log("Failed to load Context7 patterns: " + e)
            return default_analysis_patterns()
```

---

## Security Pattern Integration

### OWASP Top 10 Patterns

```text
load_owasp_patterns():
    owasp = context7.get_library_docs("<security/owasp>",
                topic="OWASP Top 10 vulnerability detection patterns", tokens=5000)
    return {
        a01_injection:        owasp.injection default [],
        a02_broken_auth:      owasp.authentication default [],
        a03_injection_data:   owasp.data_injection default [],
        a04_xss:              owasp.xss default [],
        a05_security_misconfig: owasp.misconfiguration default [],
        a06_old_components:   owasp.outdated default [],
        a07_auth_failures:    owasp.auth_failure default [],
        a08_data_failures:    owasp.data_failure default [],
        a09_security_logging: owasp.logging default [],
        a10_ssrf:             owasp.ssrf default []
    }
```

### Semgrep Rule Integration

```text
load_semgrep_rules():
    # Semgrep rules are language-aware; select the ruleset for the host language.
    semgrep = context7.get_library_docs("<security/semgrep>",
                topic="Semgrep security rules for <language>", tokens=6000)
    return {
        injection_rules:     semgrep.injection default [],
        crypto_rules:        semgrep.crypto default [],
        authentication_rules:semgrep.auth default [],
        resource_rules:      semgrep.resource default [],
        serialization_rules: semgrep.serialization default []
    }
```

---

## Quality Pattern Integration

### SonarQube / Sonar Quality Rules

```text
load_sonarqube_rules():
    rules = context7.get_library_docs("<code-quality/sonarqube>",
                topic="SonarQube quality rules code smells for <language>", tokens=5000)
    return {
        complexity_rules:      rules.complexity default [],
        maintainability_rules: rules.maintainability default [],
        reliability_rules:     rules.reliability default [],
        security_rules:        rules.security default [],
        style_rules:           rules.style default []
    }
```

---

## Performance Pattern Integration

### Profiling Best Practices

```text
load_performance_patterns():
    perf = context7.get_library_docs("<performance/profiling>",
                topic="<language> performance profiling optimization patterns", tokens=5000)
    return {
        anti_patterns:           perf.anti_patterns default [],
        optimization_techniques: perf.optimizations default [],
        profiling_strategies:    perf.profiling default [],
        benchmarking_methods:    perf.benchmarking default []
    }
```

---

## Error Handling and Fallbacks

Default patterns when Context7 is unavailable. The regexes below are illustrative shapes; populate per-language patterns that match the host language's idioms.

```text
default_analysis_patterns():
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

---

## Caching Strategy

```text
class CachedContext7Analyzer(Context7CodeAnalyzer, cache_duration_hours = 24):
    cache_duration = cache_duration_hours * 3600
    pattern_cache  = {}

    load_analysis_patterns(language = "python"):
        cache_key = language + "_patterns"
        cached = pattern_cache.get(cache_key)
        # Check cache validity
        if cached and (now() - cached.timestamp) < cache_duration:
            return cached.patterns
        # Load fresh patterns
        patterns = super().load_analysis_patterns(language)
        pattern_cache[cache_key] = { patterns, timestamp: now() }
        return patterns
```

---

## Best Practices

1. Fallback Patterns: Always provide default patterns when Context7 unavailable
2. Caching: Implement caching to reduce Context7 API calls
3. Token Management: Use appropriate token allocation for each pattern type
4. Error Handling: Implement robust error handling for Context7 failures
5. Pattern Updates: Refresh patterns periodically for latest security/quality standards
6. Gradual Loading: Load patterns on-demand to reduce initial load time
7. Custom Patterns: Allow project-specific pattern customization
8. Documentation: Document pattern sources and update frequencies

---

## Related Modules

- [Security Analysis](../security-analysis.md): Security pattern usage
- [Quality Metrics](../quality-metrics.md): Quality pattern integration
- [trust5-framework.md](./trust5-framework.md): TRUST 5 pattern loading

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/automated-code-review/context7-integration.md`
