# Best Practices Enforcement

Context7-powered best practices validation and automated standards enforcement.

## Overview

The Best Practices Engine validates coding standards and best practices using real-time documentation from Context7, ensuring code follows latest framework and language standards.

## Best Practices Engine Implementation

```python
class BestPracticesEngine:
 """Context7-powered best practices validation and enforcement"""

 def __init__(self, context7_client: Context7Client):
 self.context7_client = context7_client
 self.language_rules = self._load_language_rules()
 self.practice_validators = self._initialize_validators()

 async def validate(self, codebase: str, languages: List[str], context7_docs: bool = True) -> PracticesResult:
 """Validate coding best practices with real-time documentation"""

 validation_results = {}

 for language in languages:
 # Get latest best practices from Context7
 if context7_docs:
 try:
 library_id = await self.context7_client.resolve_library_id(language)
 latest_docs = await self.context7_client.get_library_docs(
 context7CompatibleLibraryID=library_id,
 topic="best-practices",
 tokens=5000
 )

 # Validate against latest standards
 validation_result = await self._validate_against_latest_standards(
 codebase, language, latest_docs
 )

 except Exception as e:
 logger.warning(f"Failed to get Context7 docs for {language}: {e}")
 # Fallback to cached rules
 validation_result = await self._validate_with_cached_rules(
 codebase, language
 )
 else:
 validation_result = await self._validate_with_cached_rules(
 codebase, language
 )

 validation_results[language] = validation_result

 # Cross-language best practices
 cross_language_result = await self._validate_cross_language_practices(codebase)

 # Calculate overall practices score
 overall_score = sum(
 result.score for result in validation_results.values()
 ) / len(validation_results)

 return PracticesResult(
 language_results=validation_results,
 cross_language_practices=cross_language_result,
 overall_score=overall_score,
 compliance_level=self._determine_compliance_level(overall_score),
 improvement_roadmap=self._create_improvement_roadmap(validation_results)
 )

 async def _validate_against_latest_standards(
 self,
 codebase: str,
 language: str,
 latest_docs: str
 ) -> LanguageValidationResult:
 """Validate code against latest language standards from Context7"""

 # Extract best practices from documentation
 best_practices = await self._extract_best_practices_from_docs(latest_docs)

 # Validate naming conventions
 naming_result = await self._validate_naming_conventions(
 codebase, language, best_practices.get("naming", {})
 )

 # Validate code structure
 structure_result = await self._validate_code_structure(
 codebase, language, best_practices.get("structure", {})
 )

 # Validate error handling
 error_handling_result = await self._validate_error_handling(
 codebase, language, best_practices.get("error_handling", {})
 )

 # Validate documentation
 documentation_result = await self._validate_documentation(
 codebase, language, best_practices.get("documentation", {})
 )

 # Validate testing patterns
 testing_result = await self._validate_testing_patterns(
 codebase, language, best_practices.get("testing", {})
 )

 # Calculate language-specific score
 language_score = (
 naming_result.score * 0.2 +
 structure_result.score * 0.3 +
 error_handling_result.score * 0.2 +
 documentation_result.score * 0.15 +
 testing_result.score * 0.15
 )

 return LanguageValidationResult(
 language=language,
 score=language_score,
 validations={
 "naming": naming_result,
 "structure": structure_result,
 "error_handling": error_handling_result,
 "documentation": documentation_result,
 "testing": testing_result
 },
 best_practices_version=await self._get_docs_version(latest_docs),
 recommendations=self._generate_language_recommendations(
 naming_result, structure_result, error_handling_result,
 documentation_result, testing_result
 )
 )
```

## Configuration

```yaml
best_practices:
 enabled: true
 context7_integration: true
 auto_update_standards: true
 compliance_target: 0.85

 language_rules:
 python:
 style_guide: "pep8"
 formatter: "black"
 linter: "ruff"
 type_checker: "mypy"

 javascript:
 style_guide: "airbnb"
 formatter: "prettier"
 linter: "eslint"

 typescript:
 style_guide: "google"
 formatter: "prettier"
 linter: "eslint"
```

## Context7 Library Mappings

```python
QUALITY_LIBRARY_MAPPINGS = {
 # Static Analysis Tools
 "eslint": "/eslint/eslint",
 "prettier": "/prettier/prettier",
 "black": "/psf/black",
 "ruff": "/astral-sh/ruff",
 "mypy": "/python/mypy",
 "pylint": "/pylint-dev/pylint",
 "sonarqube": "/SonarSource/sonarqube",

 # Testing Frameworks
 "jest": "/facebook/jest",
 "pytest": "/pytest-dev/pytest",
 "mocha": "/mochajs/mocha",
 "junit": "/junit-team/junit5",

 # Security Tools
 "bandit": "/PyCQA/bandit",
 "snyk": "/snyk/snyk",
 "owasp-zap": "/zaproxy/zaproxy",

 # Performance Tools
 "lighthouse": "/GoogleChrome/lighthouse",
 "py-spy": "/benfred/py-spy",

 # Documentation Standards
 "openapi": "/OAI/OpenAPI-Specification",
 "sphinx": "/sphinx-doc/sphinx",
 "jsdoc": "/jsdoc/jsdoc"
}
```

## Custom Quality Rules

```python
class CustomQualityRule:
 """Define custom quality validation rules"""

 def __init__(self, name: str, validator: Callable, severity: str = "medium"):
 self.name = name
 self.validator = validator
 self.severity = severity

 async def validate(self, codebase: str) -> RuleResult:
 """Execute custom rule validation"""
 try:
 result = await self.validator(codebase)
 return RuleResult(
 rule_name=self.name,
 passed=result.passed,
 severity=self.severity,
 details=result.details,
 recommendations=result.recommendations
 )
 except Exception as e:
 return RuleResult(
 rule_name=self.name,
 passed=False,
 severity="error",
 details={"error": str(e)},
 recommendations=["Fix rule implementation"]
 )

# Usage example
async def custom_naming_convention_rule(codebase: str):
 """Custom rule: Enforce project-specific naming conventions"""

 patterns = {
 "api_endpoints": r"^[a-z]+_[a-z]+$",
 "database_models": r"^[A-Z][a-zA-Z]*Model$",
 "utility_functions": r"^util_[a-z_]+$"
 }

 violations = []
 for pattern_name, pattern in patterns.items():
 pattern_violations = await scan_for_pattern_violations(codebase, pattern, pattern_name)
 violations.extend(pattern_violations)

 return RuleValidationResult(
 passed=len(violations) == 0,
 details={"violations": violations, "total_violations": len(violations)},
 recommendations=[f"Fix {len(violations)} naming convention violations"]
 )

# Register custom rule
custom_rule = CustomQualityRule(
 name="project_naming_conventions",
 validator=custom_naming_convention_rule,
 severity="medium"
)

quality_orchestrator.register_custom_rule(custom_rule)
```

## Related

- [TRUST 5 Validation](trust5-validation.md)
- [Proactive Analysis](proactive-analysis.md)
- [Integration Patterns](integration-patterns.md)
