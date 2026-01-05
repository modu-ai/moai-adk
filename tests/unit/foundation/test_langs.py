"""Tests for moai_adk.foundation.langs module."""


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
    """Test LanguageInfo dataclass."""

    def test_language_info_init(self):
        """Test LanguageInfo initialization."""
        info = LanguageInfo(name="Python", version="3.13")
        assert info.name == "Python"
        assert info.version == "3.13"
        assert info.is_supported is True

    def test_language_info_contains(self):
        """Test LanguageInfo __contains__ method."""
        info = LanguageInfo(name="Python", version="3.13", features=["async", "type_hints"])
        assert "async" in info
        assert "missing" not in info


class TestPattern:
    """Test Pattern dataclass."""

    def test_pattern_init(self):
        """Test Pattern initialization."""
        pattern = Pattern(
            pattern_type="best_practice",
            language="Python",
            description="Test pattern",
            example="code example",
        )
        assert pattern.pattern_type == "best_practice"
        assert pattern.language == "Python"


class TestTestingStrategy:
    """Test TestingStrategy dataclass."""

    def test_testing_strategy_init(self):
        """Test TestingStrategy initialization."""
        strategy = TestingStrategy(
            framework="pytest",
            features="async support",
            language="Python",
        )
        assert strategy.framework == "pytest"
        assert strategy.language == "Python"


class TestLanguageVersionManager:
    """Test LanguageVersionManager class."""

    def test_detect_python_version(self):
        """Test detect method with Python version."""
        manager = LanguageVersionManager()
        result = manager.detect("Python 3.13")
        assert result.name == "Python"
        assert result.version == "3.13"
        assert result.is_supported is True

    def test_detect_invalid_format(self):
        """Test detect with invalid format."""
        manager = LanguageVersionManager()
        result = manager.detect("invalid")
        assert result.is_supported is False

    def test_detect_unsupported_language(self):
        """Test detect with unsupported language."""
        manager = LanguageVersionManager()
        result = manager.detect("Lisp 2.0")
        assert result.is_supported is False

    def test_check_version_support_deprecated(self):
        """Test _check_version_support with deprecated version."""
        manager = LanguageVersionManager()
        result = manager._check_version_support("Python", "2.7")
        assert result is False

    def test_check_version_support_current(self):
        """Test _check_version_support with current version."""
        manager = LanguageVersionManager()
        result = manager._check_version_support("Python", "3.13")
        assert result is True

    def test_get_tier_python(self):
        """Test get_tier method for Python."""
        manager = LanguageVersionManager()
        result = manager.get_tier("Python")
        assert result == "Tier 1"

    def test_get_tier_unsupported(self):
        """Test get_tier with unsupported language."""
        manager = LanguageVersionManager()
        result = manager.get_tier("Unsupported")
        assert result is None


class TestFrameworkRecommender:
    """Test FrameworkRecommender class."""

    def test_get_frameworks_for_python(self):
        """Test get_frameworks_for method for Python."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Python")
        assert isinstance(frameworks, list)
        assert len(frameworks) > 0

    def test_get_frameworks_for_unsupported(self):
        """Test get_frameworks_for with unsupported language."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_for("Unsupported")
        assert frameworks == []

    def test_get_frameworks_by_category_python_apis(self):
        """Test get_frameworks_by_category for Python APIs."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("Python", "apis")
        assert isinstance(frameworks, list)
        assert "FastAPI" in frameworks

    def test_get_frameworks_by_category_typescript_frontend(self):
        """Test get_frameworks_by_category for TypeScript frontend."""
        recommender = FrameworkRecommender()
        frameworks = recommender.get_frameworks_by_category("TypeScript", "frontend")
        assert "React 19" in frameworks or "Next.js 16" in frameworks


class TestPatternAnalyzer:
    """Test PatternAnalyzer class."""

    def test_identify_async_pattern(self):
        """Test identify_pattern with async code."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("async def my_function(): await something()", "Python")
        assert pattern is not None
        assert pattern.pattern_type == "best_practice"

    def test_identify_type_hints_pattern(self):
        """Test identify_pattern with type hints."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("def func() -> str:", "Python")
        assert pattern is not None
        assert pattern.pattern_type == "best_practice"

    def test_identify_parameterized_query_pattern(self):
        """Test identify_pattern with parameterized query."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("SELECT * FROM users WHERE id = ?", "SQL")
        assert pattern is not None

    def test_identify_no_pattern(self):
        """Test identify_pattern with code that has no identified pattern."""
        analyzer = PatternAnalyzer()
        pattern = analyzer.identify_pattern("x = 1", "Python")
        assert pattern is None


class TestAntiPatternDetector:
    """Test AntiPatternDetector class."""

    def test_detect_callback_hell(self):
        """Test detect_anti_pattern with callback hell."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("callback(() => { function() => {} })", "JavaScript")
        assert pattern is not None
        assert pattern.pattern_type == "anti_pattern"

    def test_detect_global_state(self):
        """Test detect_anti_pattern with global state."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("global my_state", "Python")
        assert pattern is not None

    def test_detect_sql_injection(self):
        """Test detect_anti_pattern with SQL injection."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("f'SELECT * FROM users WHERE id = {user_id}'", "Python")
        assert pattern is not None
        assert pattern.severity == "critical"

    def test_detect_no_anti_pattern(self):
        """Test detect_anti_pattern with safe code."""
        detector = AntiPatternDetector()
        pattern = detector.detect_anti_pattern("x = 1", "Python")
        assert pattern is None


class TestEcosystemAnalyzer:
    """Test EcosystemAnalyzer class."""

    def test_analyze_python(self):
        """Test analyze method for Python."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("Python")
        assert result is not None
        assert result.name == "Python"
        assert "FastAPI" in result.frameworks

    def test_analyze_typescript(self):
        """Test analyze method for TypeScript."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("TypeScript")
        assert result is not None
        assert result.name == "TypeScript"

    def test_analyze_go(self):
        """Test analyze method for Go."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("Go")
        assert result is not None

    def test_analyze_unsupported(self):
        """Test analyze with unsupported language."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.analyze("Unsupported")
        assert result is None

    def test_analyze_multiple(self):
        """Test analyze_multiple method."""
        analyzer = EcosystemAnalyzer()
        results = analyzer.analyze_multiple(["Python", "TypeScript", "Go"])
        assert len(results) == 3

    def test_check_compatibility_python_2(self):
        """Test check_compatibility with Python 2."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.check_compatibility("Python", "2.7", "Django")
        assert result is False

    def test_check_compatibility_valid(self):
        """Test check_compatibility with valid combination."""
        analyzer = EcosystemAnalyzer()
        result = analyzer.check_compatibility("Python", "3.13", "FastAPI")
        assert result is True


class TestPerformanceOptimizer:
    """Test PerformanceOptimizer class."""

    def test_get_tips_python(self):
        """Test get_tips for Python."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Python")
        assert isinstance(tips, list)
        assert len(tips) > 0

    def test_get_tips_typescript(self):
        """Test get_tips for TypeScript."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("TypeScript")
        assert isinstance(tips, list)

    def test_get_tips_unsupported(self):
        """Test get_tips for unsupported language."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_tips("Unsupported")
        assert tips == []

    def test_get_tip_count(self):
        """Test get_tip_count method."""
        optimizer = PerformanceOptimizer()
        count = optimizer.get_tip_count("Python")
        assert count > 0


class TestTestingStrategyAdvisor:
    """Test TestingStrategyAdvisor class."""

    def test_get_strategy_python(self):
        """Test get_strategy for Python."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("Python")
        assert strategy is not None
        assert strategy.framework == "pytest"

    def test_get_strategy_typescript(self):
        """Test get_strategy for TypeScript."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("TypeScript")
        assert strategy is not None

    def test_get_strategy_unsupported(self):
        """Test get_strategy for unsupported language."""
        advisor = TestingStrategyAdvisor()
        strategy = advisor.get_strategy("Unsupported")
        assert strategy is None

    def test_get_recommended_framework_python(self):
        """Test get_recommended_framework for Python."""
        advisor = TestingStrategyAdvisor()
        framework = advisor.get_recommended_framework("Python")
        assert framework == "pytest"

    def test_get_recommended_framework_unsupported(self):
        """Test get_recommended_framework for unsupported language."""
        advisor = TestingStrategyAdvisor()
        framework = advisor.get_recommended_framework("Unsupported")
        assert framework is None
