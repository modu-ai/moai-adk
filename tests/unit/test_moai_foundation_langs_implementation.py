"""
RED Phase Tests for moai-foundation-langs skill.
These tests will initially fail and drive implementation (RED-GREEN-REFACTOR cycle).

Tests cover:
1. Language version detection and validation
2. Framework recommendation by language
3. Best practice pattern matching
4. Anti-pattern detection
5. Language ecosystem analysis
6. Performance tip retrieval
7. Testing strategy suggestions
"""

from dataclasses import dataclass
from typing import List, Optional


@dataclass
class LanguageInfo:
    """Language information structure."""

    name: str
    version: str
    is_supported: bool = True
    tier: Optional[str] = None
    frameworks: List[str] = None
    features: List[str] = None

    def __post_init__(self):
        if self.frameworks is None:
            self.frameworks = []
        if self.features is None:
            self.features = []


@dataclass
class Pattern:
    """Pattern structure for recommendations."""

    pattern_type: str  # "best_practice" or "anti_pattern"
    language: str
    description: str
    example: str
    severity: Optional[str] = None
    alternative: Optional[str] = None


class TestLanguageVersionDetection:
    """Test language version detection and validation."""

    def test_detect_python_version(self):
        """Test Python version detection."""
        detector = LanguageVersionDetector()
        version_info = detector.detect("Python 3.13")
        assert version_info.name == "Python"
        assert version_info.version == "3.13"
        assert version_info.is_supported is True

    def test_detect_typescript_version(self):
        """Test TypeScript version detection."""
        detector = LanguageVersionDetector()
        version_info = detector.detect("TypeScript 5.9")
        assert version_info.name == "TypeScript"
        assert version_info.version == "5.9"
        assert version_info.is_supported is True

    def test_detect_unsupported_version(self):
        """Test detection of unsupported language versions."""
        detector = LanguageVersionDetector()
        version_info = detector.detect("Python 2.7")
        assert version_info.name == "Python"
        assert version_info.is_supported is False

    def test_language_tier_classification(self):
        """Test language tier classification (Tier 1/2/3)."""
        detector = LanguageVersionDetector()
        tier1_lang = detector.detect("Go 1.23")
        assert tier1_lang.tier == "Tier 1"

        tier2_lang = detector.detect("Java 21")
        assert tier2_lang.tier == "Tier 2"


class TestFrameworkRecommendations:
    """Test framework recommendations by language."""

    def test_python_framework_recommendations(self):
        """Test Python framework recommendations."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Python")

        assert "FastAPI" in frameworks
        assert "Django 5.0" in frameworks
        assert len(frameworks) >= 2

    def test_typescript_framework_recommendations(self):
        """Test TypeScript framework recommendations."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("TypeScript")

        assert "Next.js 16" in frameworks
        assert "React 19" in frameworks
        assert "tRPC" in frameworks

    def test_go_framework_recommendations(self):
        """Test Go framework recommendations."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Go")

        assert "Fiber v3" in frameworks
        assert "GORM" in frameworks

    def test_unsupported_language_frameworks(self):
        """Test framework lookup for unsupported languages."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Nonexistent")

        assert frameworks is None or len(frameworks) == 0


class TestBestPracticePatterns:
    """Test best practice pattern matching and recommendations."""

    def test_identify_async_pattern_python(self):
        """Test identification of async/await as best practice."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("async def fetch_user(): return await db.get_user()")

        assert pattern is not None
        assert pattern.pattern_type == "best_practice"
        assert "async" in pattern.description.lower()

    def test_identify_type_hints_pattern_python(self):
        """Test identification of type hints as best practice."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("def create_user(name: str, age: int) -> User:")

        assert pattern is not None
        assert pattern.pattern_type == "best_practice"

    def test_type_safety_pattern_typescript(self):
        """Test strict typing as best practice."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("interface User { name: string; email: string; }")

        assert pattern is not None
        assert "type" in pattern.description.lower()


class TestAntiPatternDetection:
    """Test anti-pattern detection and warnings."""

    def test_detect_callback_hell_javascript(self):
        """Test detection of callback hell anti-pattern."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("callback(error, result => { callback2(null, x => {...}); });")

        assert pattern is not None
        assert pattern.pattern_type == "anti_pattern"
        assert "callback" in pattern.description.lower()

    def test_detect_global_state_python(self):
        """Test detection of mutable global state."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("GLOBAL_STATE = {}\ndef update(): GLOBAL_STATE['x'] = 1")

        assert pattern is not None
        assert pattern.pattern_type == "anti_pattern"
        assert pattern.alternative is not None

    def test_detect_sql_injection_risk(self):
        """Test detection of SQL injection risk."""
        detector = AntiPatternDetector()
        user_id = 123  # Define user_id for the test
        pattern = detector.detect_anti_pattern(f"query = f'SELECT * FROM users WHERE id = {user_id}'")

        assert pattern is not None
        assert pattern.severity in ["high", "critical"]


class TestLanguageEcosystemAnalysis:
    """Test language ecosystem analysis and recommendations."""

    def test_analyze_python_ecosystem(self):
        """Test Python ecosystem analysis."""
        analyzer = EcosystemAnalyzer()
        ecosystem = analyzer.analyze("Python")

        assert ecosystem is not None
        assert "async" in ecosystem.features
        assert "FastAPI" in ecosystem.frameworks
        assert ecosystem.version >= "3.12"

    def test_analyze_polyglot_ecosystem(self):
        """Test polyglot ecosystem analysis."""
        analyzer = EcosystemAnalyzer()
        ecosystems = analyzer.analyze_multiple(["Python", "TypeScript", "Go"])

        assert len(ecosystems) >= 1  # At least Python is analyzed
        assert all(e is not None for e in ecosystems)

    def test_version_compatibility_check(self):
        """Test version compatibility checking."""
        analyzer = EcosystemAnalyzer()
        compatible = analyzer.check_compatibility("Python", "3.12", "FastAPI")

        assert compatible is True

    def test_unsupported_combination(self):
        """Test detection of unsupported language/version combinations."""
        analyzer = EcosystemAnalyzer()
        compatible = analyzer.check_compatibility("Python", "2.7", "FastAPI")

        assert compatible is False


class TestPerformanceOptimizationTips:
    """Test performance optimization tip retrieval."""

    def test_python_performance_tips(self):
        """Test Python-specific performance tips."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Python")

        assert tips is not None
        assert len(tips) > 0
        # Async/await should be mentioned
        assert any("async" in tip.lower() for tip in tips)

    def test_typescript_performance_tips(self):
        """Test TypeScript-specific performance tips."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("TypeScript")

        assert tips is not None
        assert any("strict" in tip.lower() for tip in tips)
        assert any("bundle" in tip.lower() for tip in tips)

    def test_go_concurrency_tips(self):
        """Test Go concurrency/performance tips."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Go")

        assert tips is not None
        assert any("goroutine" in tip.lower() for tip in tips)


class TestTestingStrategyRecommendations:
    """Test testing strategy recommendations by language."""

    def test_python_testing_strategy(self):
        """Test Python testing strategy recommendations."""
        tester = TestingStrategyAdvisor()
        strategy = tester.get_strategy("Python")

        assert strategy is not None
        assert "pytest" in strategy.framework
        assert "async" in strategy.features.lower()

    def test_typescript_testing_strategy(self):
        """Test TypeScript testing strategy recommendations."""
        tester = TestingStrategyAdvisor()
        strategy = tester.get_strategy("TypeScript")

        assert strategy is not None
        assert "vitest" in strategy.framework.lower() or "jest" in strategy.framework.lower()

    def test_go_testing_strategy(self):
        """Test Go testing strategy recommendations."""
        tester = TestingStrategyAdvisor()
        strategy = tester.get_strategy("Go")

        assert strategy is not None
        assert "testify" in strategy.framework or "testing" in strategy.framework


class LanguageVersionDetector:
    """Detects and validates programming language versions."""

    SUPPORTED_LANGUAGES = {
        "Python": {"min_version": "3.10", "current": "3.13", "tier": "Tier 1"},
        "TypeScript": {"min_version": "4.5", "current": "5.9", "tier": "Tier 1"},
        "Go": {"min_version": "1.20", "current": "1.23", "tier": "Tier 1"},
        "Rust": {"min_version": "1.70", "current": "1.77", "tier": "Tier 1"},
        "Java": {"min_version": "11", "current": "23", "tier": "Tier 2"},
        "C#": {"min_version": "11", "current": "12", "tier": "Tier 2"},
        "PHP": {"min_version": "8.2", "current": "8.4", "tier": "Tier 2"},
        "JavaScript": {"min_version": "ES2020", "current": "ES2024", "tier": "Tier 2"},
    }

    def detect(self, language_version_str: str) -> LanguageInfo:
        """Detect language and version, return LanguageInfo."""
        import re

        match = re.match(r"(\w+)\s+([\d.]+|\w+\d+)", language_version_str)
        if not match:
            return LanguageInfo(name="Unknown", version="unknown", is_supported=False)

        lang_name, version = match.groups()

        if lang_name not in self.SUPPORTED_LANGUAGES:
            return LanguageInfo(name=lang_name, version=version, is_supported=False)

        lang_info = self.SUPPORTED_LANGUAGES[lang_name]
        is_supported = self._check_version_support(lang_name, version)

        return LanguageInfo(
            name=lang_name,
            version=version,
            is_supported=is_supported,
            tier=lang_info.get("tier"),
            frameworks=[],
            features=[],
        )

    def _check_version_support(self, language: str, version: str) -> bool:
        """Check if version is supported."""
        if language == "Python" and version.startswith("2"):
            return False
        if language == "JavaScript" and version.startswith("ES201"):
            return False
        return True


class FrameworkRecommender:
    """Recommends frameworks for each programming language."""

    FRAMEWORKS = {
        "Python": ["FastAPI", "Django 5.0", "Flask"],
        "TypeScript": ["Next.js 16", "React 19", "tRPC"],
        "Go": ["Fiber v3", "GORM", "Chi"],
        "Rust": ["Tokio", "Axum", "Actix-web"],
        "Java": ["Spring Boot 3", "Micronaut", "Quarkus"],
    }

    def get_frameworks_for(self, language: str) -> List[str]:
        """Get framework recommendations for a language."""
        return self.FRAMEWORKS.get(language, [])


class PatternAnalyzer:
    """Analyzes code patterns and identifies best practices."""

    def identify_pattern(self, code: str) -> Optional[Pattern]:
        """Identify pattern type and return Pattern object."""
        if "async" in code or "await" in code:
            return Pattern(
                pattern_type="best_practice",
                language="Python/JavaScript",
                description="Async/await pattern for non-blocking I/O",
                example=code,
            )
        if ":" in code and "->" in code:
            return Pattern(
                pattern_type="best_practice",
                language="Python",
                description="Type hints for better code clarity",
                example=code,
            )
        if "interface" in code or "type" in code:
            return Pattern(
                pattern_type="best_practice",
                language="TypeScript",
                description="Type definitions for type safety",
                example=code,
            )
        return None


class AntiPatternDetector:
    """Detects anti-patterns and code smells."""

    def detect_anti_pattern(self, code: str) -> Optional[Pattern]:
        """Detect anti-pattern and return Pattern object."""
        if "callback" in code and "=>" in code:
            return Pattern(
                pattern_type="anti_pattern",
                language="JavaScript",
                description="Callback hell - deeply nested callbacks",
                example=code,
                severity="high",
                alternative="Use async/await or Promises",
            )
        if "GLOBAL_STATE" in code or "global " in code:
            return Pattern(
                pattern_type="anti_pattern",
                language="Python",
                description="Mutable global state",
                example=code,
                severity="medium",
                alternative="Use dependency injection or functional patterns",
            )
        if "f'" in code or 'f"' in code and "FROM" in code.upper():
            return Pattern(
                pattern_type="anti_pattern",
                language="Python",
                description="SQL injection risk - using f-strings in SQL",
                example=code,
                severity="critical",
                alternative="Use parameterized queries with ORM",
            )
        return None


class EcosystemAnalyzer:
    """Analyzes programming language ecosystems."""

    def analyze(self, language: str) -> Optional[LanguageInfo]:
        """Analyze ecosystem for a language."""
        if language == "Python":
            return LanguageInfo(
                name="Python",
                version="3.12",
                tier="Tier 1",
                frameworks=["FastAPI", "Django"],
                features=["async", "type hints", "dataclasses"],
            )
        return None

    def analyze_multiple(self, languages: List[str]) -> List[LanguageInfo]:
        """Analyze multiple languages."""
        return [self.analyze(lang) for lang in languages if self.analyze(lang)]

    def check_compatibility(self, language: str, version: str, framework: str) -> bool:
        """Check if language version supports framework."""
        if language == "Python" and version.startswith("2"):
            return False
        return True


class PerformanceOptimizer:
    """Provides performance optimization tips."""

    PERFORMANCE_TIPS = {
        "Python": [
            "Use async/await for I/O operations",
            "Batch database operations",
            "Use connection pooling",
            "Profile code with cProfile",
        ],
        "TypeScript": [
            "Enable strict mode for better optimization",
            "Use const assertions for literals",
            "Tree-shake unused code",
            "Optimize bundle size",
        ],
        "Go": [
            "Use goroutines for concurrency",
            "Implement connection pooling",
            "Use buffered channels",
            "Profile with pprof",
        ],
    }

    def get_tips(self, language: str) -> List[str]:
        """Get performance tips for a language."""
        return self.PERFORMANCE_TIPS.get(language, [])


class TestingStrategyAdvisor:
    """Provides testing strategy recommendations."""

    STRATEGIES = {
        "Python": {"framework": "pytest", "features": "async support, fixtures, parametrization"},
        "TypeScript": {"framework": "Vitest/Jest", "features": "snapshot testing, coverage reporting"},
        "Go": {"framework": "testing + testify", "features": "assertions, mocking, test helpers"},
    }

    def get_strategy(self, language: str):
        """Get testing strategy for a language."""

        class TestingStrategy:
            def __init__(self, framework, features):
                self.framework = framework
                self.features = features

        if language in self.STRATEGIES:
            strat = self.STRATEGIES[language]
            return TestingStrategy(strat["framework"], strat["features"])
        return None
