"""
Comprehensive DDD tests for frontend.py module.
Tests cover all 7 classes.
"""

from moai_adk.foundation.frontend import (
    AccessibilityValidator,
    ComponentArchitect,
    DesignSystemBuilder,
    FrontendMetricsCollector,
    PerformanceMetrics,
    PerformanceOptimizer,
    ResponsiveLayoutPlanner,
    StateManagementAdvisor,
    StateManagementSolution,
)

# ============================================================================
# Test ComponentArchitect
# ============================================================================


class TestComponentArchitect:
    """Test suite for ComponentArchitect class."""

    def test_initialization(self):
        """Test architect initialization."""
        architect = ComponentArchitect()
        assert architect.components_registry == {}
        assert "render_props" in architect.composition_patterns

    def test_validate_atomic_structure_valid(self):
        """Test validation of valid atomic design structure."""
        architect = ComponentArchitect()
        components = {
            "atoms": ["Button", "Input"],
            "molecules": ["FormInput"],
            "organisms": ["Header"],
            "pages": ["HomePage"],
        }

        result = architect.validate_atomic_structure(components)

        assert result["valid"] is True
        assert result["hierarchy_level"] == 4
        assert result["atom_count"] == 2
        assert result["molecule_count"] == 1

    def test_validate_atomic_structure_invalid(self):
        """Test validation with invalid structure."""
        architect = ComponentArchitect()
        components = {"invalid_level": ["Component"]}

        result = architect.validate_atomic_structure(components)

        assert result["valid"] is False

    def test_analyze_reusability(self):
        """Test component reusability analysis."""
        architect = ComponentArchitect()
        # The implementation counts dict keys, not list items in a "props" key
        components = {
            "Button": {"children": "ReactNode", "onClick": "Function"},  # 2 props
            "Complex": {"a": "str", "b": "str", "c": "str", "d": "str", "e": "str", "f": "str"},  # 6 props
        }

        result = architect.analyze_reusability(components)

        assert result["reusable_count"] == 1
        assert result["composition_score"] > 0
        assert len(result["recommendations"]) == 1

    def test_validate_composition_patterns(self):
        """Test composition pattern validation."""
        architect = ComponentArchitect()
        patterns = {"render_props": "Use render props for flexibility", "hooks": "Use React hooks"}

        result = architect.validate_composition_patterns(patterns)

        assert result["valid"] is True
        assert result["pattern_count"] == 2

    def test_generate_prop_schema(self):
        """Test TypeScript prop schema generation."""
        architect = ComponentArchitect()
        schema = {"name": str, "age": int, "active": bool}

        result = architect.generate_prop_schema(schema)

        assert "typescript_types" in result
        assert "interface Props" in result["typescript_types"]
        assert len(result["validation_rules"]) == 3


# ============================================================================
# Test StateManagementAdvisor
# ============================================================================


class TestStateManagementAdvisor:
    """Test suite for StateManagementAdvisor class."""

    def test_initialization(self):
        """Test advisor initialization."""
        advisor = StateManagementAdvisor()
        assert StateManagementSolution.CONTEXT_API in advisor.solutions.values()

    def test_recommend_solution_small_app(self):
        """Test recommendation for small app."""
        advisor = StateManagementAdvisor()
        metrics = {"complexity": "small", "components": 20, "async_actions": False}

        result = advisor.recommend_solution(metrics)

        assert result["solution"] == "Local State"
        assert result["confidence"] == 0.95

    def test_recommend_solution_medium_app(self):
        """Test recommendation for medium app."""
        advisor = StateManagementAdvisor()
        metrics = {"complexity": "medium", "components": 60, "async_actions": False}

        result = advisor.recommend_solution(metrics)

        assert result["solution"] == "Context API"
        assert result["confidence"] == 0.85

    def test_recommend_solution_large_app(self):
        """Test recommendation for large app."""
        advisor = StateManagementAdvisor()
        metrics = {"complexity": "large", "components": 200, "async_actions": True}

        result = advisor.recommend_solution(metrics)

        assert result["solution"] == "Redux Toolkit"
        assert result["confidence"] > 0.8

    def test_validate_context_pattern_valid(self):
        """Test valid context pattern."""
        advisor = StateManagementAdvisor()
        pattern = {"splitting": True, "actions": ["setState", "dispatch"]}

        result = advisor.validate_context_pattern(pattern)

        assert result["valid"] is True

    def test_validate_context_pattern_no_splitting(self):
        """Test context pattern without splitting."""
        advisor = StateManagementAdvisor()
        pattern = {"splitting": False, "actions": ["action1", "action2", "action3", "action4", "action5", "action6"]}

        result = advisor.validate_context_pattern(pattern)

        assert result["valid"] is False
        assert "splitting" in result["issues"][0].lower()

    def test_validate_zustand_design(self):
        """Test Zustand store design validation."""
        advisor = StateManagementAdvisor()
        design = {
            "selectors": ["useUser", "usePosts"],
            "devtools_enabled": True,
            "persist_enabled": True,
            "actions": ["fetchUser", "updateUser"],
        }

        result = advisor.validate_zustand_design(design)

        assert result["valid"] is True
        assert result["devtools_status"] == "enabled"

    def test_validate_redux_design(self):
        """Test Redux slice design validation."""
        advisor = StateManagementAdvisor()
        slices = {"user": {"actions": ["login", "logout"], "async_thunks": ["fetchUser"]}}

        result = advisor.validate_redux_design(slices)

        assert result["valid"] is True
        assert result["slice_count"] == 1
        assert result["total_actions"] == 3


# ============================================================================
# Test AccessibilityValidator
# ============================================================================


class TestAccessibilityValidator:
    """Test suite for AccessibilityValidator class."""

    def test_initialization(self):
        """Test validator initialization."""
        validator = AccessibilityValidator()
        assert "AA" in validator.wcag_rules
        assert validator.min_contrast_ratio["AA"] == 4.5

    def test_validate_wcag_compliance_aa(self):
        """Test WCAG AA compliance validation."""
        validator = AccessibilityValidator()
        component = {"color_contrast_ratio": 4.5, "aria_label": "Submit form", "keyboard_accessible": True}

        result = validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is True
        assert result["level"] == "AA"
        assert len(result["failures"]) == 0

    def test_validate_wcag_compliance_aaa(self):
        """Test WCAG AAA compliance validation."""
        validator = AccessibilityValidator()
        component = {"color_contrast_ratio": 7.0, "aria_label": "Submit form", "keyboard_accessible": True}

        result = validator.validate_wcag_compliance(component, "AAA")

        assert result["compliant"] is True
        assert result["level"] == "AAA"

    def test_validate_wcag_compliance_failures(self):
        """Test WCAG compliance with failures."""
        validator = AccessibilityValidator()
        component = {"color_contrast_ratio": 2.5, "aria_label": None, "keyboard_accessible": False}

        result = validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is False
        assert len(result["failures"]) == 3

    def test_validate_aria_implementation_valid(self):
        """Test valid ARIA implementation."""
        validator = AccessibilityValidator()
        component = {
            "inputs": [
                {"aria_label": "Name", "aria_required": "true"},
                {"aria_label": "Email", "aria_describedby": "email-hint"},
            ],
            "buttons": [{"aria_label": "Submit"}],
        }

        result = validator.validate_aria_implementation(component)

        assert result["valid"] is True
        assert result["aria_count"] >= 3

    def test_validate_keyboard_navigation_valid(self):
        """Test valid keyboard navigation."""
        validator = AccessibilityValidator()
        component = {
            "focusable_elements": ["button", "input", "select"],
            "tab_order_correct": True,
            "focus_trap": True,
            "escape_key_handler": True,
            "focus_restoration": True,
        }

        result = validator.validate_keyboard_navigation(component)

        assert result["valid"] is True
        assert result["focus_management_score"] == 1.0


# ============================================================================
# Test PerformanceOptimizer
# ============================================================================


class TestPerformanceOptimizer:
    """Test suite for PerformanceOptimizer class."""

    def test_initialization(self):
        """Test optimizer initialization."""
        optimizer = PerformanceOptimizer()
        assert "lcp" in optimizer.core_web_vitals_thresholds
        assert optimizer.core_web_vitals_thresholds["lcp"]["good"] == 2.5

    def test_validate_code_splitting_valid(self):
        """Test valid code splitting."""
        optimizer = PerformanceOptimizer()
        strategy = {
            "chunks": {"vendor": "vendor.js", "main": "main.js", "runtime": "runtime.js", "polyfills": "polyfills.js"},
            "dynamic_imports": 10,
            "route_based_splitting": True,
        }

        result = optimizer.validate_code_splitting(strategy)

        assert result["optimized"] is True
        assert result["chunk_count"] == 4
        assert result["vendor_chunk_separated"] is True

    def test_validate_memoization(self):
        """Test memoization validation."""
        optimizer = PerformanceOptimizer()
        strategy = {
            "render_count_baseline": 100,
            "render_count_optimized": 20,
            "memo_components": ["UserList", "ProductCard"],
        }

        result = optimizer.validate_memoization(strategy)

        assert result["optimized"] is True
        assert result["improvement_percentage"] == 80.0

    def test_validate_performance_metrics_good(self):
        """Test validation of good performance metrics."""
        optimizer = PerformanceOptimizer()
        metrics = {"lcp_seconds": 1.8, "fid_milliseconds": 45, "cls_value": 0.08, "gzip_size_kb": 50}

        result = optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is True
        assert result["lcp_status"] == "good"
        assert result["fid_status"] == "good"
        assert result["cls_status"] == "good"

    def test_validate_performance_metrics_needs_improvement(self):
        """Test validation of metrics needing improvement."""
        optimizer = PerformanceOptimizer()
        metrics = {"lcp_seconds": 3.5, "fid_milliseconds": 250, "cls_value": 0.3, "gzip_size_kb": 100}

        result = optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is False
        assert result["lcp_status"] == "needs_improvement"
        assert result["fid_status"] == "needs_improvement"


# ============================================================================
# Test DesignSystemBuilder
# ============================================================================


class TestDesignSystemBuilder:
    """Test suite for DesignSystemBuilder class."""

    def test_initialization(self):
        """Test builder initialization."""
        builder = DesignSystemBuilder()
        assert builder.tokens == {}
        assert builder.components_doc == {}

    def test_define_design_tokens(self):
        """Test design token definition."""
        builder = DesignSystemBuilder()
        tokens = {"colors": {"primary": "#0ea5e9", "secondary": "#6366f1"}, "spacing": {"sm": "0.5rem", "md": "1rem"}}

        result = builder.define_design_tokens(tokens)

        assert result["token_count"] == 4
        assert "colors" in result["categories"]
        assert len(result["css_variables"]) == 4
        assert result["theme_support"] == "light_dark"

    def test_generate_css_variables(self):
        """Test CSS variable generation."""
        builder = DesignSystemBuilder()
        tokens = {"colors": {"primary": "#0ea5e9"}, "spacing": {"sm": "0.5rem"}}

        result = builder.define_design_tokens(tokens)

        assert "--colors-primary" in result["css_variables"]
        assert "--spacing-sm" in result["css_variables"]


# ============================================================================
# Test ResponsiveLayoutPlanner
# ============================================================================


class TestResponsiveLayoutPlanner:
    """Test suite for ResponsiveLayoutPlanner class."""

    def test_initialization(self):
        """Test planner initialization."""
        planner = ResponsiveLayoutPlanner()
        assert "mobile" in planner.standard_breakpoints
        assert planner.standard_breakpoints["lg"] == 1024

    def test_validate_breakpoints_valid(self):
        """Test valid breakpoint validation."""
        planner = ResponsiveLayoutPlanner()
        breakpoints = {"mobile": 0, "mobile_first": True, "sm": 640, "md": 768, "lg": 1024, "xl": 1280}

        result = planner.validate_breakpoints(breakpoints)

        assert result["mobile_first"] is True
        assert result["breakpoint_count"] == 5
        assert result["valid"] is True

    def test_validate_fluid_layout(self):
        """Test fluid layout validation."""
        planner = ResponsiveLayoutPlanner()
        layout_config = {
            "container_query_enabled": True,
            "fluid_spacing": True,
            "responsive_typography": True,
            "responsive_images": True,
            "aspect_ratio_preserved": True,
            "grid_columns_responsive": {"sm": 1, "md": 2, "lg": 3},
        }

        result = planner.validate_fluid_layout(layout_config)

        assert result["fluid"] is True
        assert result["container_queries_enabled"] is True
        assert result["responsive_score"] == 1.0
        assert result["grid_responsive"] is True

    def test_validate_image_strategy(self):
        """Test image strategy validation."""
        planner = ResponsiveLayoutPlanner()
        strategy = {
            "srcset_enabled": True,
            "lazy_loading": True,
            "image_optimization": True,
            "webp_format": True,
            "placeholder_strategy": True,
            "breakpoint_images": {"mobile": "small.jpg", "desktop": "large.jpg"},
        }

        result = planner.validate_image_strategy(strategy)

        assert result["optimized"] is True
        assert result["lazy_loading_enabled"] is True
        assert result["webp_support"] is True
        assert result["optimization_score"] == 1.0


# ============================================================================
# Test FrontendMetricsCollector
# ============================================================================


class TestFrontendMetricsCollector:
    """Test suite for FrontendMetricsCollector class."""

    def test_initialization(self):
        """Test collector initialization."""
        collector = FrontendMetricsCollector()
        assert collector.metrics_history == []
        assert collector.thresholds["lcp"] == 2.5

    def test_collect_metrics(self):
        """Test metrics collection."""
        collector = FrontendMetricsCollector()
        metrics_dict = {"lcp": 1.8, "fid": 45, "cls": 0.08, "ttfb": 300, "fcp": 1.2, "tti": 2.5, "bundle_size": 150}

        metrics = collector.collect_metrics(metrics_dict)

        assert isinstance(metrics, PerformanceMetrics)
        assert metrics.lcp == 1.8
        assert metrics.fid == 45
        assert len(collector.metrics_history) == 1

    def test_analyze_metrics_good(self):
        """Test analysis of good metrics."""
        collector = FrontendMetricsCollector()
        metrics = PerformanceMetrics(lcp=1.8, fid=45, cls=0.08, ttfb=300, fcp=1.2, tti=2.5, bundle_size=150)

        result = collector.analyze_metrics(metrics)

        assert result["performance_score"] == 1.0
        assert result["core_web_vitals_pass"] is True
        assert result["lcp_status"] == "good"
        assert result["fid_status"] == "good"
        assert result["cls_status"] == "good"

    def test_analyze_metrics_poor(self):
        """Test analysis of poor metrics."""
        collector = FrontendMetricsCollector()
        metrics = PerformanceMetrics(lcp=3.5, fid=250, cls=0.3, ttfb=600, fcp=2.0, tti=4.0, bundle_size=250)

        result = collector.analyze_metrics(metrics)

        assert result["performance_score"] < 1.0
        assert result["core_web_vitals_pass"] is False
        assert len(result["recommendations"]) > 0

    def test_generate_recommendations(self):
        """Test recommendations generation."""
        collector = FrontendMetricsCollector()
        metrics = PerformanceMetrics(lcp=3.5, fid=150, cls=0.2, ttfb=300, fcp=1.2, tti=2.5, bundle_size=250)

        result = collector.analyze_metrics(metrics)

        assert "Optimize LCP" in str(result["recommendations"])
        assert "Reduce FID" in str(result["recommendations"])
        assert "Fix CLS" in str(result["recommendations"])


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for frontend data classes."""

    def test_performance_metrics_creation(self):
        """Test PerformanceMetrics dataclass creation."""
        metrics = PerformanceMetrics(lcp=1.8, fid=45, cls=0.08, ttfb=300, fcp=1.2, tti=2.5, bundle_size=150)

        assert metrics.lcp == 1.8
        assert metrics.fid == 45
        assert metrics.timestamp is not None
