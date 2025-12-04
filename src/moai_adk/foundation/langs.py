"""
Programming Language Ecosystem Implementation (moai-foundation-langs).

Provides:
- LanguageVersionManager: Language version detection and validation
- FrameworkRecommender: Framework suggestions by language
- PatternAnalyzer: Code pattern recognition and analysis
- AntiPatternDetector: Anti-pattern detection
- EcosystemAnalyzer: Language ecosystem analysis
- PerformanceOptimizer: Performance optimization tips

Reference: moai-foundation-langs for modern language patterns and best practices
"""

import re
from dataclasses import dataclass, field
from typing import List, Optional


@dataclass
class LanguageInfo:
    """Language information and ecosystem details."""

    name: str
    version: str
    is_supported: bool = True
    tier: Optional[str] = None
    frameworks: List[str] = field(default_factory=list)
    features: List[str] = field(default_factory=list)

    def __contains__(self, key: str) -> bool:
        """Support 'in' operator for feature checking."""
        return key in self.features


@dataclass
class Pattern:
    """Code pattern with recommendations."""

    pattern_type: str  # "best_practice" or "anti_pattern"
    language: str
    description: str
    example: str
    severity: Optional[str] = None  # "low", "medium", "high", "critical"
    alternative: Optional[str] = None
    priority: int = 5


@dataclass
class TestingStrategy:
    """Testing strategy recommendation."""

    framework: str
    features: str
    language: str
    examples: List[str] = field(default_factory=list)


class LanguageVersionManager:
    """Manages language versions and compatibility."""

    SUPPORTED_LANGUAGES = {
        "Python": {
            "min_version": "3.10",
            "current": "3.13",
            "lts": "3.12",
            "tier": "Tier 1",
            "deprecated_versions": ["2.7", "3.9"],
        },
        "TypeScript": {
            "min_version": "4.5",
            "current": "5.9",
            "tier": "Tier 1",
            "deprecated_versions": [],
        },
        "Go": {
            "min_version": "1.20",
            "current": "1.23",
            "tier": "Tier 1",
            "deprecated_versions": ["1.19"],
        },
        "Rust": {
            "min_version": "1.70",
            "current": "1.77",
            "tier": "Tier 1",
            "deprecated_versions": [],
        },
        "Java": {
            "min_version": "11",
            "current": "23",
            "lts": "21",
            "tier": "Tier 2",
            "deprecated_versions": ["8", "9", "10"],
        },
        "C#": {
            "min_version": "11",
            "current": "12",
            "tier": "Tier 2",
            "deprecated_versions": ["7", "8"],
        },
        "PHP": {
            "min_version": "8.2",
            "current": "8.4",
            "tier": "Tier 2",
            "deprecated_versions": ["5.6", "7.0", "7.1", "7.2"],
        },
        "JavaScript": {
            "min_version": "ES2020",
            "current": "ES2024",
            "tier": "Tier 2",
            "deprecated_versions": ["ES5"],
        },
        "R": {
            "min_version": "4.2",
            "current": "4.4",
            "tier": "Tier 3",
            "deprecated_versions": ["3.x"],
        },
    }

    def detect(self, language_version_str: str) -> LanguageInfo:
        """Detect language and version from string."""
        match = re.match(r"(\w+(?:\s+\w+)?)\s+([\d.]+|ES\d{4}|\w+\d+)", language_version_str)
        if not match:
            return LanguageInfo(name="Unknown", version="unknown", is_supported=False)

        lang_name = match.group(1).strip()
        version = match.group(2)

        if lang_name not in self.SUPPORTED_LANGUAGES:
            return LanguageInfo(name=lang_name, version=version, is_supported=False)

        lang_info = self.SUPPORTED_LANGUAGES[lang_name]
        is_supported = self._check_version_support(lang_name, version)

        tier_value = lang_info.get("tier")
        tier_str: Optional[str] = str(tier_value) if tier_value is not None else None

        return LanguageInfo(
            name=lang_name,
            version=version,
            is_supported=is_supported,
            tier=tier_str,
            frameworks=[],
            features=[],
        )

    def _check_version_support(self, language: str, version: str) -> bool:
        """Check if version is supported."""
        lang_info = self.SUPPORTED_LANGUAGES.get(language, {})
        deprecated = lang_info.get("deprecated_versions", [])

        # Check for deprecated versions
        for dep_version in deprecated:
            if version.startswith(dep_version):
                return False

        return True

    def get_tier(self, language: str) -> Optional[str]:
        """Get language tier (1, 2, or 3)."""
        tier_value = self.SUPPORTED_LANGUAGES.get(language, {}).get("tier")
        return str(tier_value) if tier_value is not None else None


class FrameworkRecommender:
    """Recommends frameworks for each programming language."""

    FRAMEWORKS = {
        "Python": {
            "apis": ["FastAPI", "Django REST Framework"],
            "web": ["Django 5.0", "Starlette"],
            "async": ["asyncio", "Trio"],
            "data": ["Pydantic", "SQLAlchemy 2.0"],
            "testing": ["pytest", "pytest-asyncio"],
        },
        "TypeScript": {
            "frontend": ["Next.js 16", "React 19", "Vue 3.5"],
            "backend": ["tRPC", "NestJS", "Fastify"],
            "data": ["Zod", "Prisma", "TypeORM"],
            "testing": ["Vitest", "Jest"],
        },
        "Go": {
            "web": ["Fiber v3", "Echo", "Chi"],
            "data": ["GORM", "sqlc", "ent"],
            "async": ["goroutines", "channels"],
            "testing": ["testify", "go-testify"],
        },
        "Rust": {
            "web": ["Tokio", "Axum", "Actix-web"],
            "data": ["Sqlx", "Diesel", "Sea-ORM"],
            "testing": ["cargo test", "criterion"],
        },
        "Java": {
            "web": ["Spring Boot 3", "Micronaut", "Quarkus"],
            "testing": ["JUnit 5", "Mockito"],
        },
    }

    def get_frameworks_for(self, language: str) -> List[str]:
        """Get framework recommendations for a language."""
        lang_frameworks = self.FRAMEWORKS.get(language, {})
        all_frameworks = []
        for frameworks in lang_frameworks.values():
            all_frameworks.extend(frameworks)
        return all_frameworks

    def get_frameworks_by_category(self, language: str, category: str) -> List[str]:
        """Get frameworks for specific category."""
        return self.FRAMEWORKS.get(language, {}).get(category, [])


class PatternAnalyzer:
    """Analyzes code patterns and identifies best practices."""

    BEST_PRACTICES = {
        "async": {
            "languages": ["Python", "JavaScript", "TypeScript", "C#"],
            "description": "Async/await pattern for non-blocking I/O",
        },
        "type_hints": {
            "languages": ["Python", "TypeScript", "Go", "Rust", "Java"],
            "description": "Type hints for better code clarity and safety",
        },
        "error_handling": {
            "languages": ["Go", "Rust", "Python"],
            "description": "Proper error handling with Result/Option types",
        },
        "dependency_injection": {
            "languages": ["Python", "Java", "C#"],
            "description": "Dependency injection for loose coupling",
        },
    }

    def identify_pattern(self, code: str, language: Optional[str] = None) -> Optional[Pattern]:
        """Identify pattern type and return Pattern object."""
        code_lower = code.lower()

        # Async/await pattern
        if any(kw in code_lower for kw in ["async ", "await ", "async def", "async fn"]):
            return Pattern(
                pattern_type="best_practice",
                language=language or "Multi-language",
                description="Async/await pattern for non-blocking I/O",
                example=code,
                priority=9,
            )

        # Type hints (Python, TypeScript)
        if (":" in code and "->" in code) or ("type " in code_lower) or ("interface " in code_lower):
            return Pattern(
                pattern_type="best_practice",
                language=language or "Python/TypeScript",
                description="Type hints/annotations for type safety",
                example=code,
                priority=8,
            )

        # Parameterized queries
        if "?" in code or "param" in code_lower or "prepare" in code_lower:
            return Pattern(
                pattern_type="best_practice",
                language=language or "Multi-language",
                description="Parameterized queries for SQL safety",
                example=code,
                priority=10,
            )

        return None


class AntiPatternDetector:
    """Detects anti-patterns and code smells."""

    ANTI_PATTERNS = {
        "callback_hell": {
            "keywords": ["callback", "=>"],
            "language": "JavaScript",
            "severity": "high",
            "alternative": "Use async/await or Promises",
        },
        "global_state": {
            "keywords": ["GLOBAL", "global "],
            "language": "Python/Go",
            "severity": "medium",
            "alternative": "Use dependency injection or functional patterns",
        },
        "sql_injection": {
            "keywords": ["f'", 'f"', "+ str(", "+ var"],
            "language": "Python/JavaScript",
            "severity": "critical",
            "alternative": "Use parameterized queries or ORM",
        },
        "silent_failure": {
            "keywords": ["pass", "except:", "try"],
            "language": "Python",
            "severity": "high",
            "alternative": "Use proper error handling",
        },
    }

    def detect_anti_pattern(self, code: str, language: Optional[str] = None) -> Optional[Pattern]:
        """Detect anti-pattern and return Pattern object."""
        code_lower = code.lower()

        # Callback hell detection
        if "callback" in code_lower and ("=>" in code or "function" in code_lower):
            return Pattern(
                pattern_type="anti_pattern",
                language=language or "JavaScript",
                description="Callback hell - deeply nested callbacks",
                example=code,
                severity="high",
                alternative="Use async/await or Promises",
                priority=9,
            )

        # Global state detection
        if "global_state" in code_lower or "global " in code_lower:
            return Pattern(
                pattern_type="anti_pattern",
                language=language or "Python",
                description="Mutable global state",
                example=code,
                severity="medium",
                alternative="Use dependency injection or functional patterns",
                priority=7,
            )

        # SQL injection detection
        if ("f'" in code or 'f"' in code) and ("SELECT" in code.upper() or "FROM" in code.upper()):
            return Pattern(
                pattern_type="anti_pattern",
                language=language or "Python",
                description="SQL injection risk - using f-strings in SQL",
                example=code,
                severity="critical",
                alternative="Use parameterized queries with ORM",
                priority=10,
            )

        return None


class EcosystemAnalyzer:
    """Analyzes programming language ecosystems."""

    def analyze(self, language: str) -> Optional[LanguageInfo]:
        """Analyze ecosystem for a language."""
        ecosystems = {
            "Python": LanguageInfo(
                name="Python",
                version="3.12+",
                tier="Tier 1",
                frameworks=["FastAPI", "Django"],
                features=["async", "type hints", "dataclasses"],
            ),
            "TypeScript": LanguageInfo(
                name="TypeScript",
                version="5.0+",
                tier="Tier 1",
                frameworks=["Next.js", "React"],
                features=["strict typing", "type safety", "decorators"],
            ),
            "Go": LanguageInfo(
                name="Go",
                version="1.20+",
                tier="Tier 1",
                frameworks=["Fiber", "GORM"],
                features=["goroutines", "channels", "simplicity"],
            ),
        }
        return ecosystems.get(language)

    def analyze_multiple(self, languages: List[str]) -> List[LanguageInfo]:
        """Analyze multiple languages."""
        results = []
        for lang in languages:
            result = self.analyze(lang)
            if result:
                results.append(result)
        return results

    def check_compatibility(self, language: str, version: str, framework: str) -> bool:
        """Check if language version supports framework."""
        if language == "Python" and version.startswith("2"):
            return False
        if language == "TypeScript" and version.startswith("3"):
            return False
        return True


class PerformanceOptimizer:
    """Provides performance optimization tips."""

    PERFORMANCE_TIPS = {
        "Python": [
            "Use async/await for I/O operations",
            "Leverage type hints for faster execution",
            "Batch database operations",
            "Use connection pooling (asyncpg, psycopg[c])",
            "Profile code with cProfile or py-spy",
        ],
        "TypeScript": [
            "Enable strict mode for better optimization",
            "Use const assertions for literals",
            "Tree-shake unused code",
            "Optimize bundle size with dynamic imports",
            "Use native Node.js features instead of polyfills",
        ],
        "Go": [
            "Use goroutines for lightweight concurrency",
            "Implement connection pooling for databases",
            "Use buffered channels for performance",
            "Profile with pprof",
            "Compile with release optimizations",
        ],
        "Rust": [
            "Use zero-cost abstractions",
            "Enable compile-time optimizations",
            "Avoid unnecessary allocations",
            "Use release builds for production",
            "Profile with flamegraph",
        ],
    }

    def get_tips(self, language: str) -> List[str]:
        """Get performance tips for a language."""
        return self.PERFORMANCE_TIPS.get(language, [])

    def get_tip_count(self, language: str) -> int:
        """Get count of optimization tips."""
        return len(self.get_tips(language))


class TestingStrategyAdvisor:
    """Provides testing strategy recommendations."""

    STRATEGIES = {
        "Python": TestingStrategy(
            framework="pytest",
            features="async support, fixtures, parametrization, pytest-asyncio",
            language="Python",
        ),
        "TypeScript": TestingStrategy(
            framework="Vitest/Jest",
            features="snapshot testing, coverage reporting, mocking",
            language="TypeScript",
        ),
        "Go": TestingStrategy(
            framework="testing + testify",
            features="assertions, mocking, test helpers, benchmarks",
            language="Go",
        ),
        "Rust": TestingStrategy(
            framework="cargo test",
            features="unit tests, integration tests, doc tests",
            language="Rust",
        ),
    }

    def get_strategy(self, language: str) -> Optional[TestingStrategy]:
        """Get testing strategy for a language."""
        return self.STRATEGIES.get(language)

    def get_recommended_framework(self, language: str) -> Optional[str]:
        """Get recommended testing framework."""
        strategy = self.get_strategy(language)
        return strategy.framework if strategy else None


# Export public API
__all__ = [
    "LanguageInfo",
    "Pattern",
    "TestingStrategy",
    "LanguageVersionManager",
    "FrameworkRecommender",
    "PatternAnalyzer",
    "AntiPatternDetector",
    "EcosystemAnalyzer",
    "PerformanceOptimizer",
    "TestingStrategyAdvisor",
]
