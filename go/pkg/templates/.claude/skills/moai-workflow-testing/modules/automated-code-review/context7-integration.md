# Context7 MCP Integration

> Module: Context7 integration patterns for real-time security and quality analysis
> Parent: [Automated Code Review](../automated-code-review.md)
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, Context7 MCP client

## Quick Reference

### Context7 Integration Overview

Context7 MCP provides real-time access to:
- OWASP Top 10 security vulnerability patterns
- Semgrep security detection rules
- SonarQube code quality standards
- Performance optimization libraries
- TRUST 5 validation frameworks

### Core Integration Pattern

```python
class Context7CodeAnalyzer:
    """Integration with Context7 for code analysis patterns."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.analysis_patterns = {}
        self.security_patterns = {}
        self.performance_patterns = {}

    async def load_analysis_patterns(self, language: str = "python") -> Dict[str, Any]:
        """Load code analysis patterns from Context7."""

        if not self.context7:
            return self._get_default_analysis_patterns()

        try:
            # Load security analysis patterns
            security_patterns = await self.context7.get_library_docs(
                context7_library_id="/security/semgrep",
                topic="security vulnerability detection patterns 2025",
                tokens=4000
            )
            self.security_patterns = security_patterns

            # Load performance analysis patterns
            performance_patterns = await self.context7.get_library_docs(
                context7_library_id="/performance/python-profiling",
                topic="performance anti-patterns code analysis 2025",
                tokens=3000
            )
            self.performance_patterns = performance_patterns

            # Load code quality patterns
            quality_patterns = await self.context7.get_library_docs(
                context7_library_id="/code-quality/sonarqube",
                topic="code quality best practices smells detection 2025",
                tokens=4000
            )

            # Load TRUST 5 validation patterns
            trust_patterns = await self.context7.get_library_docs(
                context7_library_id="/code-review/trust-framework",
                topic="TRUST 5 code validation framework patterns 2025",
                tokens=3000
            )

            return {
                'security': security_patterns,
                'performance': performance_patterns,
                'quality': quality_patterns,
                'trust': trust_patterns
            }

        except Exception as e:
            print(f"Failed to load Context7 patterns: {e}")
            return self._get_default_analysis_patterns()
```

---

## Security Pattern Integration

### OWASP Top 10 Patterns

```python
async def load_owasp_patterns(self) -> Dict[str, Any]:
    """Load OWASP Top 10 vulnerability patterns."""

    owasp_patterns = await self.context7.get_library_docs(
        context7_library_id="/security/owasp",
        topic="OWASP Top 10 2021 vulnerability detection patterns",
        tokens=5000
    )

    return {
        'a01_injection': owasp_patterns.get('injection', []),
        'a02_broken_auth': owasp_patterns.get('authentication', []),
        'a03_injection_data': owasp_patterns.get('data_injection', []),
        'a04_xss': owasp_patterns.get('xss', []),
        'a05_security_misconfig': owasp_patterns.get('misconfiguration', []),
        'a06_old_components': owasp_patterns.get('outdated', []),
        'a07_auth_failures': owasp_patterns.get('auth_failure', []),
        'a08_data_failures': owasp_patterns.get('data_failure', []),
        'a09_security_logging': owasp_patterns.get('logging', []),
        'a10_ssrf': owasp_patterns.get('ssrf', [])
    }
```

### Semgrep Rule Integration

```python
async def load_semgrep_rules(self) -> Dict[str, Any]:
    """Load Semgrep security rules."""

    semgrep_rules = await self.context7.get_library_docs(
        context7_library_id="/security/semgrep",
        topic="Semgrep Python security rules 2025",
        tokens=6000
    )

    return {
        'injection_rules': semgrep_rules.get('injection', []),
        'crypto_rules': semgrep_rules.get('crypto', []),
        'authentication_rules': semgrep_rules.get('auth', []),
        'resource_rules': semgrep_rules.get('resource', []),
        'serialization_rules': semgrep_rules.get('serialization', [])
    }
```

---

## Quality Pattern Integration

### SonarQube Quality Rules

```python
async def load_sonarqube_rules(self) -> Dict[str, Any]:
    """Load SonarQube code quality rules."""

    sonarqube_rules = await self.context7.get_library_docs(
        context7_library_id="/code-quality/sonarqube",
        topic="SonarQube Python quality rules code smells 2025",
        tokens=5000
    )

    return {
        'complexity_rules': sonarqube_rules.get('complexity', []),
        'maintainability_rules': sonarqube_rules.get('maintainability', []),
        'reliability_rules': sonarqube_rules.get('reliability', []),
        'security_rules': sonarqube_rules.get('security', []),
        'style_rules': sonarqube_rules.get('style', [])
    }
```

---

## Performance Pattern Integration

### Profiling Best Practices

```python
async def load_performance_patterns(self) -> Dict[str, Any]:
    """Load performance optimization patterns."""

    performance_patterns = await self.context7.get_library_docs(
        context7_library_id="/performance/python-profiling",
        topic="Python performance profiling optimization patterns 2025",
        tokens=5000
    )

    return {
        'anti_patterns': performance_patterns.get('anti_patterns', []),
        'optimization_techniques': performance_patterns.get('optimizations', []),
        'profiling_strategies': performance_patterns.get('profiling', []),
        'benchmarking_methods': performance_patterns.get('benchmarking', [])
    }
```

---

## Error Handling and Fallbacks

```python
def _get_default_analysis_patterns(self) -> Dict[str, Any]:
    """Get default analysis patterns when Context7 is unavailable."""

    return {
        'security': {
            'sql_injection': [
                r"execute\([^)]*\+[^)]*\)",
                r"format\s*\(",
                r"%\s*[^,]*s"
            ],
            'command_injection': [
                r"os\.system\(",
                r"subprocess\.call\(",
                r"eval\("
            ],
            'path_traversal': [
                r"open\([^)]*\+[^)]*\)",
                r"\.\.\/"
            ]
        },
        'performance': {
            'inefficient_loops': [
                r"for.*in.*range\(len\(",
                r"while.*len\("
            ],
            'memory_leaks': [
                r"global\s+",
                r"\.append\(.*\)\s*\.append\("
            ]
        },
        'quality': {
            'long_functions': {'max_lines': 50},
            'complex_conditionals': {'max_complexity': 10},
            'deep_nesting': {'max_depth': 4}
        }
    }
```

---

## Caching Strategy

```python
class CachedContext7Analyzer(Context7CodeAnalyzer):
    """Context7 analyzer with pattern caching."""

    def __init__(self, context7_client=None, cache_duration_hours=24):
        super().__init__(context7_client)
        self.cache_duration = cache_duration_hours * 3600
        self.pattern_cache = {}

    async def load_analysis_patterns(self, language: str = "python") -> Dict[str, Any]:
        """Load patterns with caching."""

        cache_key = f"{language}_patterns"
        cached_data = self.pattern_cache.get(cache_key)

        # Check if cache is valid
        if cached_data:
            cache_age = time.time() - cached_data['timestamp']
            if cache_age < self.cache_duration:
                return cached_data['patterns']

        # Load fresh patterns
        patterns = await super().load_analysis_patterns(language)

        # Cache the patterns
        self.pattern_cache[cache_key] = {
            'patterns': patterns,
            'timestamp': time.time()
        }

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
