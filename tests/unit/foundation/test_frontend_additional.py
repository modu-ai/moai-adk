"""
Additional comprehensive tests for moai_adk.foundation.frontend module.

Increases coverage for:
- ComponentArchitect: 75.86% → 95%
- StateManagementAdvisor: Solution recommendations
- AccessibilityValidator: WCAG compliance
- PerformanceOptimizer: Metrics validation
- DesignSystemBuilder: Design tokens
- ResponsiveLayoutPlanner: Responsive design
- FrontendMetricsCollector: Frontend metrics
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
)


class TestComponentArchitectAdditional:
    """Additional tests for ComponentArchitect."""

    def test_validate_atomic_structure_all_levels(self):
        """Test atomic structure with all levels."""
        architect = ComponentArchitect()
        components = {
            "atoms": ["Button", "Input", "Label"],
            "molecules": ["FormInput", "Card", "Modal"],
            "organisms": ["Header", "Footer", "Sidebar"],
            "pages": ["HomePage", "AboutPage", "ContactPage"],
        }
        result = architect.validate_atomic_structure(components)
        assert result["valid"] is True
        assert result["atom_count"] == 3
        assert result["molecule_count"] == 3
        assert result["organism_count"] == 3
        assert result["page_count"] == 3

    def test_validate_atomic_structure_partial_levels(self):
        """Test atomic structure with missing levels."""
        architect = ComponentArchitect()
        components = {
            "atoms": ["Button"],
            "molecules": ["FormInput"],
        }
        result = architect.validate_atomic_structure(components)
        assert result["valid"] is True
        assert result["hierarchy_level"] == 4

    def test_validate_atomic_structure_invalid_level(self):
        """Test atomic structure with invalid level."""
        architect = ComponentArchitect()
        components = {
            "atoms": ["Button"],
            "invalid_level": ["SomeComponent"],
        }
        result = architect.validate_atomic_structure(components)
        # Should accept since it has valid atoms
        assert "valid" in result or "atom_count" in result

    def test_analyze_reusability_high(self):
        """Test reusability analysis for highly reusable components."""
        architect = ComponentArchitect()
        components = {
            "Button": {"label": "", "onClick": "", "size": ""},
            "Input": {"value": "", "onChange": "", "placeholder": ""},
            "Card": {"title": "", "content": "", "footer": ""},
        }
        result = architect.analyze_reusability(components)
        assert result["reusable_count"] == 3
        assert len(result["recommendations"]) == 0

    def test_analyze_reusability_too_many_props(self):
        """Test reusability analysis for over-complex components."""
        architect = ComponentArchitect()
        components = {
            "ComplexForm": {
                "field1": "",
                "field2": "",
                "field3": "",
                "field4": "",
                "field5": "",
                "field6": "",
            }
        }
        result = architect.analyze_reusability(components)
        assert result["reusable_count"] == 0
        assert len(result["recommendations"]) > 0

    def test_analyze_reusability_mixed(self):
        """Test reusability with mixed component complexity."""
        architect = ComponentArchitect()
        components = {
            "Button": {"label": "", "onClick": ""},  # 2 props - good
            "SimpleCard": {"title": "", "content": "", "footer": ""},  # 3 props - good
            "ComplexForm": {f"field{i}": "" for i in range(10)},  # 10 props - bad
        }
        result = architect.analyze_reusability(components)
        assert result["reusable_count"] >= 2
        assert len(result["recommendations"]) > 0

    def test_validate_composition_patterns_valid(self):
        """Test composition pattern validation with valid patterns."""
        architect = ComponentArchitect()
        patterns = {
            "hooks": "Use React hooks",
            "compound_components": "Component composition",
        }
        result = architect.validate_composition_patterns(patterns)
        assert result["valid"] is True
        assert result["pattern_count"] >= 2

    def test_validate_composition_patterns_invalid(self):
        """Test composition pattern validation with invalid patterns."""
        architect = ComponentArchitect()
        patterns = {
            "invalid_pattern": "Some description",
        }
        result = architect.validate_composition_patterns(patterns)
        assert result["valid"] is False

    def test_validate_composition_patterns_mixed(self):
        """Test composition pattern validation with mixed patterns."""
        architect = ComponentArchitect()
        patterns = {
            "hooks": "Use hooks",
            "invalid": "Invalid",
            "render_props": "Render props",
        }
        result = architect.validate_composition_patterns(patterns)
        assert result["valid"] is True
        assert result["pattern_count"] >= 2

    def test_generate_prop_schema_simple(self):
        """Test prop schema generation for simple types."""
        architect = ComponentArchitect()
        schema = {
            "label": str,
            "disabled": bool,
            "size": ("small", "medium", "large"),
        }
        result = architect.generate_prop_schema(schema)
        assert "interface Props" in result["typescript_types"]
        assert "label" in result["typescript_types"]
        assert len(result["validation_rules"]) == 3


class TestStateManagementAdvisorAdditional:
    """Additional tests for StateManagementAdvisor."""

    def test_recommend_solution_very_small_app(self):
        """Test recommendation for very small app."""
        advisor = StateManagementAdvisor()
        result = advisor.recommend_solution({"complexity": "small", "components": 20, "async_actions": False})
        assert result["solution"] == "Local State"
        assert result["confidence"] == 0.95

    def test_recommend_solution_small_app_with_async(self):
        """Test recommendation for small app with async."""
        advisor = StateManagementAdvisor()
        result = advisor.recommend_solution({"complexity": "small", "components": 25, "async_actions": True})
        assert result["solution"] == "Context API"

    def test_recommend_solution_medium_app(self):
        """Test recommendation for medium app."""
        advisor = StateManagementAdvisor()
        result = advisor.recommend_solution({"complexity": "medium", "components": 75, "async_actions": True})
        # Could be Zustand or Context API depending on implementation
        assert result["solution"] in ("Zustand", "Context API", "Redux Toolkit")

    def test_recommend_solution_large_app(self):
        """Test recommendation for large app."""
        advisor = StateManagementAdvisor()
        result = advisor.recommend_solution(
            {
                "complexity": "large",
                "components": 200,
                "async_actions": True,
                "cache_needed": True,
            }
        )
        assert result["solution"] == "Redux Toolkit"

    def test_validate_context_pattern_optimal(self):
        """Test context pattern validation for optimal config."""
        advisor = StateManagementAdvisor()
        result = advisor.validate_context_pattern({"splitting": True, "actions": ["action1", "action2", "action3"]})
        assert result["valid"] is True

    def test_validate_context_pattern_too_many_actions(self):
        """Test context pattern with too many actions."""
        advisor = StateManagementAdvisor()
        result = advisor.validate_context_pattern(
            {
                "splitting": False,
                "actions": [f"action{i}" for i in range(10)],
            }
        )
        assert result["valid"] is False
        assert "splitting" in str(result["issues"]).lower()

    def test_validate_zustand_design_full_featured(self):
        """Test Zustand store design with all features."""
        advisor = StateManagementAdvisor()
        result = advisor.validate_zustand_design(
            {
                "selectors": ["getUser", "getLoading"],
                "devtools_enabled": True,
                "persist_enabled": True,
                "actions": ["setUser", "setLoading"],
            }
        )
        assert result["valid"] is True
        assert result["devtools_status"] == "enabled"
        assert result["persist_status"] == "enabled"

    def test_validate_redux_design_multiple_slices(self):
        """Test Redux design with multiple slices."""
        advisor = StateManagementAdvisor()
        slices = {
            "users": {
                "actions": ["setUser", "clearUser"],
                "async_thunks": ["fetchUsers"],
            },
            "posts": {
                "actions": ["setPosts"],
                "async_thunks": ["fetchPosts", "createPost"],
            },
        }
        result = advisor.validate_redux_design(slices)
        assert result["valid"] is True
        assert result["slice_count"] == 2
        assert result["total_actions"] >= 5


class TestAccessibilityValidatorAdditional:
    """Additional tests for AccessibilityValidator."""

    def test_validate_wcag_compliance_aa_compliant(self):
        """Test WCAG AA compliance."""
        validator = AccessibilityValidator()
        component = {
            "color_contrast_ratio": 5.0,
            "aria_label": "Submit Button",
            "keyboard_accessible": True,
        }
        result = validator.validate_wcag_compliance(component, "AA")
        assert result["compliant"] is True
        assert result["level"] == "AA"

    def test_validate_wcag_compliance_aaa_compliant(self):
        """Test WCAG AAA compliance."""
        validator = AccessibilityValidator()
        component = {
            "color_contrast_ratio": 8.0,
            "aria_label": "Submit Button",
            "keyboard_accessible": True,
        }
        result = validator.validate_wcag_compliance(component, "AAA")
        assert result["compliant"] is True

    def test_validate_wcag_compliance_low_contrast(self):
        """Test WCAG failure with low contrast."""
        validator = AccessibilityValidator()
        component = {
            "color_contrast_ratio": 2.0,
            "aria_label": "Submit",
            "keyboard_accessible": True,
        }
        result = validator.validate_wcag_compliance(component, "AA")
        assert result["compliant"] is False
        assert "contrast" in str(result["failures"]).lower()

    def test_validate_wcag_compliance_missing_aria(self):
        """Test WCAG failure with missing aria-label."""
        validator = AccessibilityValidator()
        component = {
            "color_contrast_ratio": 5.0,
            "keyboard_accessible": True,
        }
        result = validator.validate_wcag_compliance(component, "AA")
        assert result["compliant"] is False

    def test_validate_wcag_compliance_not_keyboard_accessible(self):
        """Test WCAG failure without keyboard access."""
        validator = AccessibilityValidator()
        component = {
            "color_contrast_ratio": 5.0,
            "aria_label": "Button",
            "keyboard_accessible": False,
        }
        result = validator.validate_wcag_compliance(component, "AA")
        assert result["compliant"] is False

    def test_validate_aria_implementation_with_inputs(self):
        """Test ARIA validation with input elements."""
        validator = AccessibilityValidator()
        component = {
            "inputs": [
                {"aria_label": "Email", "aria_describedby": "email_help"},
                {"aria_label": "Password"},
            ]
        }
        result = validator.validate_aria_implementation(component)
        assert result["aria_count"] >= 2
        assert "aria_label" in result["attributes_found"]

    def test_validate_aria_implementation_with_buttons(self):
        """Test ARIA validation with button elements."""
        validator = AccessibilityValidator()
        component = {
            "buttons": [
                {"aria_label": "Submit", "aria_pressed": False},
                {"aria_label": "Cancel"},
            ]
        }
        result = validator.validate_aria_implementation(component)
        assert result["aria_count"] >= 2

    def test_validate_aria_implementation_insufficient(self):
        """Test ARIA validation with insufficient attributes."""
        validator = AccessibilityValidator()
        component = {
            "inputs": [
                {"value": ""},
            ]
        }
        result = validator.validate_aria_implementation(component)
        assert result["valid"] is False

    def test_validate_keyboard_navigation_complete(self):
        """Test keyboard navigation with all features."""
        validator = AccessibilityValidator()
        component = {
            "focusable_elements": ["button1", "input1", "link1"],
            "tab_order_correct": True,
            "focus_trap": True,
            "escape_key_handler": True,
            "focus_restoration": True,
        }
        result = validator.validate_keyboard_navigation(component)
        assert result["valid"] is True
        assert result["focus_management_score"] == 1.0

    def test_validate_keyboard_navigation_no_focusable(self):
        """Test keyboard navigation without focusable elements."""
        validator = AccessibilityValidator()
        component = {
            "focusable_elements": [],
            "tab_order_correct": False,
        }
        result = validator.validate_keyboard_navigation(component)
        assert result["valid"] is False

    def test_validate_keyboard_navigation_partial(self):
        """Test keyboard navigation with partial features."""
        validator = AccessibilityValidator()
        component = {
            "focusable_elements": ["button1"],
            "tab_order_correct": True,
            "focus_trap": False,
            "escape_key_handler": False,
        }
        result = validator.validate_keyboard_navigation(component)
        assert result["valid"] is True
        assert result["focus_management_score"] == 0.25


class TestPerformanceOptimizerFrontendAdditional:
    """Additional tests for Frontend PerformanceOptimizer."""

    def test_validate_code_splitting_optimized(self):
        """Test optimized code splitting."""
        optimizer = PerformanceOptimizer()
        strategy = {
            "chunks": {
                "vendor": "size_500kb",
                "main": "size_200kb",
                "utils": "size_50kb",
                "styles": "size_100kb",
            },
            "dynamic_imports": 5,
            "route_based_splitting": True,
        }
        result = optimizer.validate_code_splitting(strategy)
        assert result["optimized"] is True
        assert result["vendor_chunk_separated"] is True

    def test_validate_code_splitting_not_optimized(self):
        """Test non-optimized code splitting."""
        optimizer = PerformanceOptimizer()
        strategy = {
            "chunks": {"main": "size_1000kb"},
            "dynamic_imports": 0,
            "route_based_splitting": False,
        }
        result = optimizer.validate_code_splitting(strategy)
        assert result["optimized"] is False

    def test_validate_memoization_high_improvement(self):
        """Test memoization with high improvement."""
        optimizer = PerformanceOptimizer()
        strategy = {
            "render_count_baseline": 100,
            "render_count_optimized": 10,
            "memo_components": ["UserCard", "PostCard"],
        }
        result = optimizer.validate_memoization(strategy)
        assert result["optimized"] is True
        assert result["improvement_percentage"] == 90.0

    def test_validate_performance_metrics_all_good(self):
        """Test performance metrics all meeting targets."""
        optimizer = PerformanceOptimizer()
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 50,
            "cls_value": 0.05,
            "gzip_size_kb": 45,
            "bundle_size_kb": 200,
        }
        result = optimizer.validate_performance_metrics(metrics)
        assert result["core_web_vitals_passed"] is True
        assert result["bundle_optimized"] is True

    def test_validate_performance_metrics_lcp_poor(self):
        """Test performance metrics with poor LCP."""
        optimizer = PerformanceOptimizer()
        metrics = {
            "lcp_seconds": 5.0,
            "fid_milliseconds": 50,
            "cls_value": 0.05,
        }
        result = optimizer.validate_performance_metrics(metrics)
        assert result["lcp_status"] == "needs_improvement"
        assert result["core_web_vitals_passed"] is False

    def test_validate_performance_metrics_cls_poor(self):
        """Test performance metrics with poor CLS."""
        optimizer = PerformanceOptimizer()
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 50,
            "cls_value": 0.3,
        }
        result = optimizer.validate_performance_metrics(metrics)
        assert result["cls_status"] == "needs_improvement"


class TestDesignSystemBuilderAdditional:
    """Additional tests for DesignSystemBuilder."""

    def test_define_design_tokens_comprehensive(self):
        """Test comprehensive design token definition."""
        builder = DesignSystemBuilder()
        tokens = {
            "colors": {
                "primary": "#0ea5e9",
                "secondary": "#f97316",
                "success": "#22c55e",
            },
            "spacing": {
                "xs": "4px",
                "sm": "8px",
                "md": "16px",
                "lg": "24px",
            },
            "typography": {
                "heading1": "32px",
                "heading2": "24px",
                "body": "16px",
            },
        }
        result = builder.define_design_tokens(tokens)
        assert result["token_count"] == 10
        assert result["theme_support"] == "light_dark"
        assert len(result["css_variables"]) == 10

    def test_define_design_tokens_minimal(self):
        """Test minimal design token definition."""
        builder = DesignSystemBuilder()
        tokens = {
            "colors": {"primary": "#000"},
        }
        result = builder.define_design_tokens(tokens)
        assert result["token_count"] == 1


class TestResponsiveLayoutPlannerAdditional:
    """Additional tests for ResponsiveLayoutPlanner."""

    def test_validate_breakpoints_mobile_first(self):
        """Test mobile-first breakpoints."""
        planner = ResponsiveLayoutPlanner()
        breakpoints = {
            "mobile_first": True,
            "sm": 640,
            "md": 768,
            "lg": 1024,
            "xl": 1280,
        }
        result = planner.validate_breakpoints(breakpoints)
        assert result["mobile_first"] is True
        assert result["valid"] is True

    def test_validate_breakpoints_insufficient(self):
        """Test insufficient breakpoints."""
        planner = ResponsiveLayoutPlanner()
        breakpoints = {
            "sm": 640,
            "md": 768,
        }
        result = planner.validate_breakpoints(breakpoints)
        assert result["valid"] is False

    def test_validate_fluid_layout_fully_optimized(self):
        """Test fully optimized fluid layout."""
        planner = ResponsiveLayoutPlanner()
        layout = {
            "container_query_enabled": True,
            "fluid_spacing": True,
            "responsive_typography": True,
            "responsive_images": True,
            "aspect_ratio_preserved": True,
        }
        result = planner.validate_fluid_layout(layout)
        assert result["fluid"] is True
        assert result["responsive_score"] == 1.0

    def test_validate_fluid_layout_not_fluid(self):
        """Test non-fluid layout."""
        planner = ResponsiveLayoutPlanner()
        layout = {
            "container_query_enabled": False,
            "fluid_spacing": False,
        }
        result = planner.validate_fluid_layout(layout)
        assert result["fluid"] is False

    def test_validate_image_strategy_optimized(self):
        """Test optimized image strategy."""
        planner = ResponsiveLayoutPlanner()
        strategy = {
            "srcset_enabled": True,
            "lazy_loading": True,
            "image_optimization": True,
            "webp_format": True,
            "placeholder_strategy": True,
            "breakpoint_images": {"mobile": "", "desktop": ""},
        }
        result = planner.validate_image_strategy(strategy)
        assert result["optimized"] is True
        assert result["webp_support"] is True

    def test_validate_image_strategy_not_optimized(self):
        """Test non-optimized image strategy."""
        planner = ResponsiveLayoutPlanner()
        strategy = {
            "srcset_enabled": False,
            "lazy_loading": False,
        }
        result = planner.validate_image_strategy(strategy)
        assert result["optimized"] is False


class TestFrontendMetricsCollectorAdditional:
    """Additional tests for FrontendMetricsCollector."""

    def test_collect_metrics_valid(self):
        """Test collecting valid performance metrics."""
        collector = FrontendMetricsCollector()
        metrics = {
            "lcp": 1.8,
            "fid": 45,
            "cls": 0.08,
            "ttfb": 300,
            "fcp": 1.2,
            "tti": 3.5,
            "bundle_size": 150,
        }
        result = collector.collect_metrics(metrics)
        assert result.lcp == 1.8
        assert result.bundle_size == 150
        assert len(collector.metrics_history) == 1

    def test_analyze_metrics_all_good(self):
        """Test metrics analysis when all good."""
        collector = FrontendMetricsCollector()
        metrics = PerformanceMetrics(lcp=1.8, fid=45, cls=0.08, ttfb=300, fcp=1.2, tti=3.5, bundle_size=150)
        result = collector.analyze_metrics(metrics)
        assert result["core_web_vitals_pass"] is True
        assert result["performance_score"] > 0.7

    def test_analyze_metrics_lcp_poor(self):
        """Test metrics analysis with poor LCP."""
        collector = FrontendMetricsCollector()
        metrics = PerformanceMetrics(lcp=4.0, fid=45, cls=0.08, ttfb=300, fcp=1.2, tti=3.5, bundle_size=150)
        result = collector.analyze_metrics(metrics)
        assert result["lcp_status"] == "needs_improvement"
        assert len(result["recommendations"]) > 0

    def test_analyze_metrics_large_bundle(self):
        """Test metrics with large bundle size."""
        collector = FrontendMetricsCollector()
        metrics = PerformanceMetrics(lcp=2.0, fid=45, cls=0.08, ttfb=300, fcp=1.2, tti=3.5, bundle_size=300)
        result = collector.analyze_metrics(metrics)
        assert any("bundle" in rec.lower() for rec in result["recommendations"])

    def test_multiple_metrics_tracking(self):
        """Test tracking multiple metrics over time."""
        collector = FrontendMetricsCollector()
        for i in range(3):
            metrics = {
                "lcp": 2.0 + i * 0.1,
                "fid": 50 + i * 5,
                "cls": 0.08 + i * 0.01,
                "ttfb": 300,
                "fcp": 1.2,
                "tti": 3.5,
                "bundle_size": 150,
            }
            collector.collect_metrics(metrics)
        assert len(collector.metrics_history) == 3
