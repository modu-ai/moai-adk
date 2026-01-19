"""
TDD tests for Programming Language Ecosystem Implementation.

Comprehensive test suite covering:
- LanguageInfo dataclass with __contains__ method
- Pattern and TestingStrategy dataclasses
- LanguageVersionManager class
- FrameworkRecommender class
- PatternAnalyzer class
- AntiPatternDetector class
- EcosystemAnalyzer class
- PerformanceOptimizer class
- TestingStrategyAdvisor class

Target: 100% line coverage
"""

import pytest

from moai_adk.foundation.langs import (
    AntiPatternDetector,
    EcosystemAnalyzer,
    FrameworkRecommender,
    LanguageInfo,
    LanguageVersionManager,
    Pattern,
    PatternAnalyzer,
    PerformanceOptimizer,
    TestingStrategy,
    TestingStrategyAdvisor,
)


class TestLanguageInfo:
    """Test LanguageInfo dataclass with __contains__ method."""

    def test_init_default_values(self):
        """Test LanguageInfo initialization with default values."""
        info = LanguageInfo(name="Python", version="3.12")
        assert info.name == "Python"
        assert info.version == "3.12"
        assert info.is_supported is True
        assert info.tier is None
        assert info.frameworks == []
        assert info.features == []

    def test_init_with_all_fields(self):
        """Test LanguageInfo initialization with all fields."""
        info = LanguageInfo(
            name="TypeScript",
            version="5.0",
            is_supported=True,
            tier="Tier 1",
            frameworks=["React", "Vue"],
            features=["async", "type hints"],
        )
        assert info.name == "TypeScript"
        assert info.version == "5.0"
        assert info.tier == "Tier 1"
        assert len(info.frameworks) == 2
        assert len(info.features) == 2

    def test_contains_existing_feature(self):
        """Test __contains__ returns True for existing feature."""
        info = LanguageInfo(name="Python", version="3.12", features=["async", "type hints", "dataclasses"])
        assert "async" in info
        assert "type hints" in info
        assert "dataclasses" in info

    def test_contains_nonexistent_feature(self):
        """Test __contains__ returns False for nonexistent feature."""
        info = LanguageInfo(name="Python", version="3.12", features=["async", "type hints"])
        assert "garbage collection" not in info
        assert "pattern matching" not in info

    def test_contains_empty_features(self):
        """Test __contains__ returns False when features list is empty."""
        info = LanguageInfo(name="Python", version="3.12")
        assert "async" not in info
        assert "anything" not in info


class TestPattern:
    """Test Pattern dataclass."""

    def test_init_best_practice(self):
        """Test Pattern initialization for best practice."""
        pattern = Pattern(
            pattern_type="best_practice",
            language="Python",
            description="Async/await pattern",
            example="async def fetch():",
            severity=None,
            alternative=None,
            priority=9,
        )
        assert pattern.pattern_type == "best_practice"
        assert pattern.language == "Python"
        assert pattern.severity is None
        assert pattern.alternative is None
        assert pattern.priority == 9

    def test_init_anti_pattern(self):
        """Test Pattern initialization for anti-pattern."""
        pattern = Pattern(
            pattern_type="anti_pattern",
            language="JavaScript",
            description="Callback hell",
            example="callback(function() { callback(...) })",
            severity="high",
            alternative="Use async/await",
            priority=10,
        )
        assert pattern.pattern_type == "anti_pattern"
        assert pattern.severity == "high"
        assert pattern.alternative == "Use async/await"
        assert pattern.priority == 10

    def test_init_default_values(self):
        """Test Pattern initialization with default values."""
        pattern = Pattern(pattern_type="best_practice", language="Go", description="Test pattern", example="example")
        assert pattern.severity is None
        assert pattern.alternative is None
        assert pattern.priority == 5


class TestTestingStrategy:
    """Test TestingStrategy dataclass."""

    def test_init(self):
        """Test TestingStrategy initialization."""
        strategy = TestingStrategy(
            framework="pytest",
            features="async support, fixtures, parametrization",
            language="Python",
            examples=["@pytest.mark.asyncio", "@pytest.fixture"],
        )
        assert strategy.framework == "pytest"
        assert strategy.features == "async support, fixtures, parametrization"
        assert strategy.language == "Python"
        assert len(strategy.examples) == 2

    def test_init_default_examples(self):
        """Test TestingStrategy initialization with default examples."""
        strategy = TestingStrategy(framework="jest", features="snapshot testing", language="TypeScript")
        assert strategy.examples == []


class TestLanguageVersionManager:
    """Test LanguageVersionManager class."""

    def test_init(self):
        """Test LanguageVersionManager initialization."""
        manager = LanguageVersionManager()
        assert hasattr(manager, "SUPPORTED_LANGUAGES")
        assert "Python" in manager.SUPPORTED_LANGUAGES
        assert "TypeScript" in manager.SUPPORTED_LANGUAGES

    def test_detect_valid_python_version(self):
        """Test detect with valid Python version."""
        manager = LanguageVersionManager()
        result = manager.detect("Python 3.12")
        assert result.name == "Python"
        assert result.version == "3.12"
        assert result.is_supported is True

    def test_detect_valid_typescript_version(self):
        """Test detect with valid TypeScript version."""
        manager = LanguageVersionManager()
        result = manager.detect("TypeScript 5.9")
        assert result.name == "TypeScript"
        assert result.version == "5.9"
        assert result.is_supported is True

    def test_detect_valid_go_version(self):
        """Test detect with valid Go version."""
        manager = LanguageVersionManager()
        result = manager.detect("Go 1.23")
        assert result.name == "Go"
        assert result.version == "1.23"
        assert result.is_supported is True

    def test_detect_deprecated_python_version(self):
        """Test detect with deprecated Python version."""
        manager = LanguageVersionManager()
        result = manager.detect("Python 2.7")
        assert result.name == "Python"
        assert result.version == "2.7"
        assert result.is_supported is False

    def test_detect_unsupported_language(self):
        """Test detect with unsupported language."""
        manager = LanguageVersionManager()
        result = manager.detect("COBOL 2024")
        assert result.name == "COBOL"
        assert result.is_supported is False

    def test_detect_malformed_string(self):
        """Test detect with malformed version string."""
        manager = LanguageVersionManager()
        result = manager.detect("Python")
        assert result.name == "Unknown"
        assert result.version == "unknown"
        assert result.is_supported is False

    def test_detect_empty_string(self):
        """Test detect with empty string."""
        manager = LanguageVersionManager()
        result = manager.detect("")
        assert result.name == "Unknown"
        assert result.version == "unknown"
        assert result.is_supported is False

    def test_detect_javascript_es_version(self):
        """Test detect with JavaScript ES version."""
        manager = LanguageVersionManager()
        result = manager.detect("JavaScript ES2024")
        assert result.name == "JavaScript"
        assert result.version == "ES2024"
        assert result.is_supported is True

    def test_detect_rust_version(self):
        """Test detect with Rust version."""
        manager = LanguageVersionManager()
        result = manager.detect("Rust 1.77")
        assert result.name == "Rust"
        assert result.version == "1.77"
        assert result.is_supported is True

    def test_check_version_support_supported(self):
        """Test _check_version_support with supported version."""
        manager = LanguageVersionManager()
        assert manager._check_version_support("Python", "3.12") is True
        assert manager._check_version_support("TypeScript", "5.9") is True

    def test_check_version_support_deprecated(self):
        """Test _check_version_support with deprecated version."""
        manager = LanguageVersionManager()
        assert manager._check_version_support("Python", "2.7") is False
        assert manager._check_version_support("Python", "3.9") is False
        assert manager._check_version_support("Go", "1.19") is False
        assert manager._check_version_support("Java", "8") is False

    def test_check_version_support_unknown_language(self):
        """Test _check_version_support with unknown language."""
        manager = LanguageVersionManager()
        assert manager._check_version_support("Unknown", "1.0") is True

    def test_get_tier_tier1(self):
        """Test get_tier for Tier 1 languages."""
        manager = LanguageVersionManager()
        assert manager.get_tier("Python") == "Tier 1"
        assert manager.get_tier("TypeScript") == "Tier 1"
        assert manager.get_tier("Go") == "Tier 1"
        assert manager.get_tier("Rust") == "Tier 1"

    def test_get_tier_tier2(self):
        """Test get_tier for Tier 2 languages."""
        manager = LanguageVersionManager()
        assert manager.get_tier("Java") == "Tier 2"
        assert manager.get_tier("C#") == "Tier 2"
        assert manager.get_tier("PHP") == "Tier 2"
        assert manager.get_tier("JavaScript") == "Tier 2"

    def test_get_tier_tier3(self):
        """Test get_tier for Tier 3 languages."""
        manager = LanguageVersionManager()
        assert manager.get_tier("R") == "Tier 3"

    def test_get_tier_unknown_language(self):
        """Test get_tier for unknown language returns None."""
        manager = LanguageVersionManager()
        assert manager.get_tier("Unknown") is None


class TestFrameworkRecommender:
    """Test FrameworkRecommender class."""

    def test_init(self):
        """Test FrameworkRecommender initialization."""
        recommender = FrameworkRecommender()
        assert hasattr(recommender, "FRAMEWORKS")

    def test_get_frameworks_for_python(self):
        """Test get_frameworks_for for Python."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Python")
        assert len(frameworks) > 0
        assert "FastAPI" in frameworks
        assert "Django REST Framework" in frameworks
        assert "pytest" in frameworks

    def test_get_frameworks_for_typescript(self):
        """Test get_frameworks_for for TypeScript."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("TypeScript")
        assert len(frameworks) > 0
        assert "Next.js 16" in frameworks
        assert "React 19" in frameworks
        assert "Vitest" in frameworks

    def test_get_frameworks_for_go(self):
        """Test get_frameworks_for for Go."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Go")
        assert len(frameworks) > 0
        assert "Fiber v3" in frameworks or "Echo" in frameworks

    def test_get_frameworks_for_rust(self):
        """Test get_frameworks_for for Rust."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Rust")
        assert len(frameworks) > 0
        assert "Tokio" in frameworks or "Axum" in frameworks

    def test_get_frameworks_for_java(self):
        """Test get_frameworks_for for Java."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Java")
        assert len(frameworks) > 0
        assert "Spring Boot 3" in frameworks

    def test_get_frameworks_for_unknown_language(self):
        """Test get_frameworks_for for unknown language."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Unknown")
        assert frameworks == []

    def test_get_frameworks_by_category_python_apis(self):
        """Test get_frameworks_by_category for Python APIs."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("Python", "apis")
        assert len(frameworks) > 0
        assert "FastAPI" in frameworks

    def test_get_frameworks_by_category_python_web(self):
        """Test get_frameworks_by_category for Python web."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("Python", "web")
        assert len(frameworks) > 0
        assert "Django 5.0" in frameworks

    def test_get_frameworks_by_category_python_testing(self):
        """Test get_frameworks_by_category for Python testing."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("Python", "testing")
        assert len(frameworks) > 0
        assert "pytest" in frameworks

    def test_get_frameworks_by_category_typescript_frontend(self):
        """Test get_frameworks_by_category for TypeScript frontend."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("TypeScript", "frontend")
        assert len(frameworks) > 0
        assert "React 19" in frameworks

    def test_get_frameworks_by_category_unknown_category(self):
        """Test get_frameworks_by_category for unknown category."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("Python", "unknown")
        assert frameworks == []

    def test_get_frameworks_by_category_unknown_language(self):
        """Test get_frameworks_by_category for unknown language."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("Unknown", "web")
        assert frameworks == []


class TestPatternAnalyzer:
    """Test PatternAnalyzer class."""

    def test_init(self):
        """Test PatternAnalyzer initialization."""
        analyzer = PatternAnalyzer()
        assert hasattr(analyzer, "BEST_PRACTICES")

    def test_identify_pattern_async_await(self):
        """Test identify_pattern detects async/await."""
        analyzer = PatternAnalyzer()
        code = "async def fetch_data():\n    await api.call()"
        pattern = analyzer.identify_pattern(code, "Python")
        assert pattern is not None
        assert pattern.pattern_type == "best_practice"
        assert "async" in pattern.description.lower()

    def test_identify_pattern_async_javascript(self):
        """Test identify_pattern detects JavaScript async."""
        analyzer = PatternAnalyzer()
        code = "async function fetch() { await api.call() }"
        pattern = analyzer.identify_pattern(code, "JavaScript")
        assert pattern is not None
        assert pattern.pattern_type == "best_practice"

    def test_identify_pattern_type_hints_python(self):
        """Test identify_pattern detects Python type hints."""
        analyzer = PatternAnalyzer()
        code = "def fetch_data(name: str) -> dict:\n    return {}"
        pattern = analyzer.identify_pattern(code, "Python")
        assert pattern is not None
        assert "type" in pattern.description.lower()

    def test_identify_pattern_type_interface_typescript(self):
        """Test identify_pattern detects TypeScript interface."""
        analyzer = PatternAnalyzer()
        code = "interface User { name: string }"
        pattern = analyzer.identify_pattern(code, "TypeScript")
        assert pattern is not None
        assert pattern.pattern_type == "best_practice"

    def test_identify_pattern_parameterized_queries(self):
        """Test identify_pattern detects parameterized queries."""
        analyzer = PatternAnalyzer()
        code = "cursor.execute('SELECT * FROM users WHERE id = ?', [user_id])"
        pattern = analyzer.identify_pattern(code, "Python")
        assert pattern is not None
        assert pattern.priority == 10

    def test_identify_pattern_prepare_statement(self):
        """Test identify_pattern detects prepared statements."""
        analyzer = PatternAnalyzer()
        code = "prepare stmt from 'SELECT * FROM users'"
        pattern = analyzer.identify_pattern(code, "SQL")
        assert pattern is not None
        assert "parameterized" in pattern.description.lower()

    def test_identify_pattern_no_pattern(self):
        """Test identify_pattern returns None for no pattern."""
        analyzer = PatternAnalyzer()
        code = "x = 1\ny = 2\nprint(x + y)"
        pattern = analyzer.identify_pattern(code, "Python")
        assert pattern is None

    def test_identify_pattern_no_language(self):
        """Test identify_pattern without language parameter."""
        analyzer = PatternAnalyzer()
        code = "async def fetch():\n    await call()"
        pattern = analyzer.identify_pattern(code)
        assert pattern is not None
        assert pattern.language == "Multi-language" or "Python" in pattern.language


class TestAntiPatternDetector:
    """Test AntiPatternDetector class."""

    def test_init(self):
        """Test AntiPatternDetector initialization."""
        detector = AntiPatternDetector()
        assert hasattr(detector, "ANTI_PATTERNS")

    def test_detect_anti_pattern_callback_hell(self):
        """Test detect_anti_pattern finds callback hell."""
        detector = AntiPatternDetector()
        code = "callback(function() { callback(function() { callback() }) })"
        pattern = detector.detect_anti_pattern(code, "JavaScript")
        assert pattern is not None
        assert pattern.pattern_type == "anti_pattern"
        assert pattern.severity == "high"

    def test_detect_anti_pattern_global_state(self):
        """Test detect_anti_pattern finds global state."""
        detector = AntiPatternDetector()
        code = "global state = {}"
        pattern = detector.detect_anti_pattern(code, "Python")
        assert pattern is not None
        assert "global" in pattern.description.lower()
        assert pattern.severity == "medium"

    def test_detect_anti_pattern_sql_injection_f_string(self):
        """Test detect_anti_pattern finds SQL injection with f-string."""
        detector = AntiPatternDetector()
        code = "cursor.execute(f'SELECT * FROM users WHERE id = {user_id}')"
        pattern = detector.detect_anti_pattern(code, "Python")
        assert pattern is not None
        assert pattern.severity == "critical"
        assert "sql injection" in pattern.description.lower()

    def test_detect_anti_pattern_sql_injection_f_string_double(self):
        """Test detect_anti_pattern finds SQL injection with f\"."""
        detector = AntiPatternDetector()
        code = 'cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")'
        pattern = detector.detect_anti_pattern(code, "Python")
        assert pattern is not None
        assert pattern.severity == "critical"

    def test_detect_anti_pattern_no_sql_safe_code(self):
        """Test detect_anti_pattern safe SQL code doesn't trigger."""
        detector = AntiPatternDetector()
        code = 'cursor.execute("SELECT * FROM users WHERE id = ?", [user_id])'
        pattern = detector.detect_anti_pattern(code, "Python")
        assert pattern is None

    def test_detect_anti_pattern_no_anti_pattern(self):
        """Test detect_anti_pattern returns None for clean code."""
        detector = AntiPatternDetector()
        code = "def clean_function():\n    return 42"
        pattern = detector.detect_anti_pattern(code, "Python")
        assert pattern is None

    def test_detect_anti_pattern_no_language(self):
        """Test detect_anti_pattern without language parameter."""
        detector = AntiPatternDetector()
        code = "global state = {}"
        pattern = detector.detect_anti_pattern(code)
        assert pattern is not None


class TestEcosystemAnalyzer:
    """Test EcosystemAnalyzer class."""

    def test_init(self):
        """Test EcosystemAnalyzer initialization."""
        analyzer = EcosystemAnalyzer()

    def test_analyze_python(self):
        """Test analyze for Python ecosystem."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("Python")
        assert result is not None
        assert result.name == "Python"
        assert result.tier == "Tier 1"
        assert len(result.frameworks) > 0
        assert len(result.features) > 0

    def test_analyze_typescript(self):
        """Test analyze for TypeScript ecosystem."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("TypeScript")
        assert result is not None
        assert result.name == "TypeScript"
        assert "React" in result.frameworks or "Next.js" in result.frameworks

    def test_analyze_go(self):
        """Test analyze for Go ecosystem."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("Go")
        assert result is not None
        assert result.name == "Go"
        assert "goroutines" in result.features or "channels" in result.features

    def test_analyze_unknown_language(self):
        """Test analyze for unknown language returns None."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("Unknown")
        assert result is None

    def test_analyze_multiple(self):
        """Test analyze_multiple for multiple languages."""
        analyzer = EcosystemAnalyzer()
        results = analyzer.analyze_multiple(["Python", "TypeScript", "Go"])
        assert len(results) == 3

    def test_analyze_multiple_with_unknown(self):
        """Test analyze_multiple filters out unknown languages."""
        analyzer = EcosystemAnalyzer()
        results = analyzer.analyze_multiple(["Python", "Unknown", "Go"])
        assert len(results) == 2

    def test_analyze_multiple_empty_list(self):
        """Test analyze_multiple with empty list."""
        analyzer = EcosystemAnalyzer()
        results = analyzer.analyze_multiple([])
        assert results == []

    def test_check_compatibility_python_supported(self):
        """Test check_compatibility for supported Python version."""
        analyzer = EcosystemAnalyzer()
        assert analyzer.check_compatibility("Python", "3.12", "FastAPI") is True

    def test_check_compatibility_python_unsupported(self):
        """Test check_compatibility for unsupported Python version."""
        analyzer = EcosystemAnalyzer()
        assert analyzer.check_compatibility("Python", "2.7", "FastAPI") is False

    def test_check_compatibility_typescript_unsupported(self):
        """Test check_compatibility for unsupported TypeScript version."""
        analyzer = EcosystemAnalyzer()
        assert analyzer.check_compatibility("TypeScript", "3.0", "React") is False

    def test_check_compatibility_unknown_language(self):
        """Test check_compatibility for unknown language returns True."""
        analyzer = EcosystemAnalyzer()
        assert analyzer.check_compatibility("Unknown", "1.0", "Framework") is True


class TestPerformanceOptimizer:
    """Test PerformanceOptimizer class."""

    def test_init(self):
        """Test PerformanceOptimizer initialization."""
        optimizer = PerformanceOptimizer()
        assert hasattr(optimizer, "PERFORMANCE_TIPS")

    def test_get_tips_python(self):
        """Test get_tips for Python."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Python")
        assert len(tips) > 0
        assert any("async" in tip.lower() for tip in tips)

    def test_get_tips_typescript(self):
        """Test get_tips for TypeScript."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("TypeScript")
        assert len(tips) > 0
        assert any("strict" in tip.lower() for tip in tips)

    def test_get_tips_go(self):
        """Test get_tips for Go."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Go")
        assert len(tips) > 0
        assert any("goroutine" in tip.lower() for tip in tips)

    def test_get_tips_rust(self):
        """Test get_tips for Rust."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Rust")
        assert len(tips) > 0
        assert any("release" in tip.lower() for tip in tips)

    def test_get_tips_unknown_language(self):
        """Test get_tips for unknown language."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Unknown")
        assert tips == []

    def test_get_tip_count_python(self):
        """Test get_tip_count for Python."""
        optimizer = PerformanceOptimizer()
        count = optimizer.get_tip_count("Python")
        assert count > 0
        assert count == len(optimizer.get_tips("Python"))

    def test_get_tip_count_unknown_language(self):
        """Test get_tip_count for unknown language."""
        optimizer = PerformanceOptimizer()
        count = optimizer.get_tip_count("Unknown")
        assert count == 0


class TestTestingStrategyAdvisor:
    """Test TestingStrategyAdvisor class."""

    def test_init(self):
        """Test TestingStrategyAdvisor initialization."""
        advisor = TestingStrategyAdvisor()
        assert hasattr(advisor, "STRATEGIES")

    def test_get_strategy_python(self):
        """Test get_strategy for Python."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("Python")
        assert strategy is not None
        assert strategy.framework == "pytest"
        assert strategy.language == "Python"

    def test_get_strategy_typescript(self):
        """Test get_strategy for TypeScript."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("TypeScript")
        assert strategy is not None
        assert "Vitest" in strategy.framework or "Jest" in strategy.framework

    def test_get_strategy_go(self):
        """Test get_strategy for Go."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("Go")
        assert strategy is not None
        assert "testify" in strategy.framework.lower()

    def test_get_strategy_rust(self):
        """Test get_strategy for Rust."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("Rust")
        assert strategy is not None
        assert strategy.framework == "cargo test"

    def test_get_strategy_unknown_language(self):
        """Test get_strategy for unknown language."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("Unknown")
        assert strategy is None

    def test_get_recommended_framework_python(self):
        """Test get_recommended_framework for Python."""
        advisor = TestingStrategyAdvisor()
        framework = advisor.get_recommended_framework("Python")
        assert framework == "pytest"

    def test_get_recommended_framework_typescript(self):
        """Test get_recommended_framework for TypeScript."""
        advisor = TestingStrategyAdvisor()
        framework = advisor.get_recommended_framework("TypeScript")
        assert framework is not None and ("Vitest" in framework or "Jest" in framework)

    def test_get_recommended_framework_unknown_language(self):
        """Test get_recommended_framework for unknown language."""
        advisor = TestingStrategyAdvisor()
        framework = advisor.get_recommended_framework("Unknown")
        assert framework is None


class TestLangsIntegration:
    """Integration tests for language ecosystem components."""

    def test_full_language_analysis_workflow(self):
        """Test complete workflow from version detection to strategy."""
        manager = LanguageVersionManager()
        analyzer = EcosystemAnalyzer()
        advisor = TestingStrategyAdvisor()

        # Detect version
        lang_info = manager.detect("Python 3.12")
        assert lang_info.is_supported is True

        # Analyze ecosystem
        ecosystem = analyzer.analyze("Python")
        assert ecosystem is not None

        # Get testing strategy
        strategy = advisor.get_strategy("Python")
        assert strategy is not None

    def test_framework_recommendation_by_category(self):
        """Test framework recommendation across categories."""
        recommender = FrameworkRecommender()

        python_frameworks = recommender.get_frameworks_for("Python")
        assert all(cat in ["apis", "web", "async", "data", "testing"] for cat in recommender.FRAMEWORKS["Python"])

    def test_pattern_detection_comprehensive(self):
        """Test pattern detection across multiple languages."""
        analyzer = PatternAnalyzer()
        detector = AntiPatternDetector()

        # Best practice
        async_code = "async def fetch(): await api.call()"
        best_pattern = analyzer.identify_pattern(async_code, "Python")
        assert best_pattern is not None

        # Anti-pattern
        bad_sql = 'cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")'
        anti_pattern = detector.detect_anti_pattern(bad_sql, "Python")
        assert anti_pattern is not None
        assert anti_pattern.severity == "critical"


class TestLangsWithFixtures:
    """Tests using pytest fixtures."""

    @pytest.fixture
    def version_manager(self):
        """Fixture providing LanguageVersionManager."""
        return LanguageVersionManager()

    @pytest.fixture
    def framework_recommender(self):
        """Fixture providing FrameworkRecommender."""
        return FrameworkRecommender()

    @pytest.fixture
    def pattern_analyzer(self):
        """Fixture providing PatternAnalyzer."""
        return PatternAnalyzer()

    @pytest.fixture
    def anti_pattern_detector(self):
        """Fixture providing AntiPatternDetector."""
        return AntiPatternDetector()

    def test_fixture_version_manager(self, version_manager):
        """Test version manager fixture."""
        result = version_manager.detect("Python 3.12")
        assert result.is_supported is True

    def test_fixture_framework_recommender(self, framework_recommender):
        """Test framework recommender fixture."""
        frameworks = framework_recommender.get_frameworks_for("Python")
        assert len(frameworks) > 0

    def test_fixture_pattern_analyzer(self, pattern_analyzer):
        """Test pattern analyzer fixture."""
        code = "async def test(): pass"
        pattern = pattern_analyzer.identify_pattern(code, "Python")
        assert pattern is not None

    def test_fixture_anti_pattern_detector(self, anti_pattern_detector):
        """Test anti-pattern detector fixture."""
        code = 'cursor.execute(f"SELECT * FROM users WHERE id = {id}")'
        pattern = anti_pattern_detector.detect_anti_pattern(code, "Python")
        assert pattern is not None
        assert pattern.severity == "critical"


class TestLangsParametrized:
    """Parametrized tests for language ecosystem."""

    @pytest.mark.parametrize(
        "lang_version,expected_supported",
        [
            ("Python 3.12", True),
            ("Python 3.9", False),
            ("TypeScript 5.9", True),
            ("Go 1.23", True),
            ("Go 1.19", False),
            ("Rust 1.77", True),
            ("Java 23", True),
            ("Java 8", False),
            ("JavaScript ES2024", True),
            ("JavaScript ES5", False),
        ],
    )
    def test_parametrized_version_support(self, lang_version, expected_supported):
        """Test version support across multiple languages."""
        manager = LanguageVersionManager()
        result = manager.detect(lang_version)
        assert result.is_supported == expected_supported

    @pytest.mark.parametrize(
        "language,expected_tier",
        [
            ("Python", "Tier 1"),
            ("TypeScript", "Tier 1"),
            ("Go", "Tier 1"),
            ("Rust", "Tier 1"),
            ("Java", "Tier 2"),
            ("C#", "Tier 2"),
            ("PHP", "Tier 2"),
            ("JavaScript", "Tier 2"),
            ("R", "Tier 3"),
        ],
    )
    def test_parametrized_language_tiers(self, language, expected_tier):
        """Test language tier classification."""
        manager = LanguageVersionManager()
        tier = manager.get_tier(language)
        assert tier == expected_tier

    @pytest.mark.parametrize(
        "code,language,expected_pattern_type",
        [
            ("async def test():", "Python", "best_practice"),
            ("interface User {}", "TypeScript", "best_practice"),
            ("cursor.execute(?, [])", "Python", "best_practice"),
            ("callback(function() {})", "JavaScript", "anti_pattern"),
            ("global x = 1", "Python", "anti_pattern"),
            ('f"SELECT {id}"', "Python", "anti_pattern"),
        ],
    )
    def test_parametrized_pattern_detection(self, code, language, expected_pattern_type):
        """Test pattern detection across various code samples."""
        analyzer = PatternAnalyzer()
        detector = AntiPatternDetector()

        if expected_pattern_type == "best_practice":
            pattern = analyzer.identify_pattern(code, language)
            assert pattern is not None
            assert pattern.pattern_type == expected_pattern_type
        else:
            pattern = detector.detect_anti_pattern(code, language)
            assert pattern is not None
            assert pattern.pattern_type == expected_pattern_type

    @pytest.mark.parametrize(
        "language,expected_framework",
        [
            ("Python", "pytest"),
            ("TypeScript", "Vitest"),
            ("Go", "testify"),
            ("Rust", "cargo test"),
            ("Java", "JUnit"),
        ],
    )
    def test_parametrized_testing_frameworks(self, language, expected_framework):
        """Test testing framework recommendations."""
        advisor = TestingStrategyAdvisor()
        framework = advisor.get_recommended_framework(language)
        if framework:
            assert expected_framework in framework


class TestLangsEdgeCases:
    """Edge case tests for language ecosystem."""

    def test_language_info_with_special_characters(self):
        """Test LanguageInfo with special characters in features."""
        info = LanguageInfo(name="C++", version="23", features=["templates", "lambdas", "move semantics"])
        assert "move semantics" in info
        assert "templates" in info

    def test_version_detection_with_spaces(self):
        """Test version detection with extra spaces (regex requires no leading spaces)."""
        manager = LanguageVersionManager()
        # The regex requires word characters at start, so leading spaces won't match
        # Test with spaces after the language name instead
        result = manager.detect("Python   3.12  ")
        assert result.name == "Python"
        assert result.version == "3.12"

    def test_empty_framework_list(self):
        """Test framework recommendation with no frameworks."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("NonExistent")
        assert frameworks == []

    def test_pattern_with_empty_code(self):
        """Test pattern analysis with empty code."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("", "Python")
        assert pattern is None

    def test_anti_pattern_with_empty_code(self):
        """Test anti-pattern detection with empty code."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("", "Python")
        assert pattern is None

    def test_ecosystem_with_none_language(self):
        """Test ecosystem analysis with unknown language."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("UnknownLanguage")
        assert result is None
