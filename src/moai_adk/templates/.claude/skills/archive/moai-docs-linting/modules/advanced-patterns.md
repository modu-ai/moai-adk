# Advanced Patterns for Documentation Linting

Enterprise-grade linting patterns and validation architectures.

---

## Pattern 1: Custom Rule Engine

Define and apply custom linting rules:

```python
class CustomRuleEngine:
    def register_rule(self, name: str, validator_func):
        """Register custom rule"""
        self.rules[name] = validator_func

    def apply_rules(self, content: str) -> List[Issue]:
        """Apply all rules to content"""
        issues = []
        for rule_name, validator in self.rules.items():
            issues.extend(validator(content))
        return issues
```

---

## Pattern 2: Contextual Linting

Apply different rules based on document type:

```python
class ContextualLinter:
    def lint_by_type(self, doc_path: Path, doc_type: str) -> List[Issue]:
        """Apply rules specific to document type"""
        if doc_type == 'api':
            return self._lint_api_doc(doc_path)
        elif doc_type == 'guide':
            return self._lint_guide(doc_path)
        elif doc_type == 'reference':
            return self._lint_reference(doc_path)
        return self._lint_generic(doc_path)
```

---

## Pattern 3: Performance Optimization

Optimize linting for large documentation sets:

```python
class OptimizedLinter:
    def lint_parallel(self, files: List[Path], workers: int = 4) -> Dict[Path, List[Issue]]:
        """Lint files in parallel"""
        from concurrent.futures import ThreadPoolExecutor
        with ThreadPoolExecutor(max_workers=workers) as executor:
            return {f: executor.submit(self.lint, f).result() for f in files}

    def lint_incremental(self, changed_files: List[Path]) -> Dict[Path, List[Issue]]:
        """Lint only changed files"""
        return {f: self.lint(f) for f in changed_files}
```

---

**Last Updated**: 2025-11-22
