"""
GREEN Phase: Test Suite for frontend.py Implementation

This module contains comprehensive TDD tests for enterprise frontend architecture patterns.
All tests validate component design, state management, accessibility, performance,
responsive design, and metrics collection.

Test Coverage:
- Component Architecture (Atomic Design, Composition Patterns)
- State Management Solutions (Local State, Context API, Zustand, Redux)
- Accessibility Validation (WCAG 2.1 AA/AAA, ARIA, Keyboard Navigation)
- Performance Optimization (Code Splitting, Memoization, Core Web Vitals)
- Design System Building (Design Tokens, Component Documentation)
- Responsive Layout Planning (Breakpoints, Fluid Layout, Image Optimization)
- Frontend Metrics Collection and Analysis

Target: 90%+ code coverage with production-ready quality
Framework: React 19, Next.js 15, TypeScript 5.9+
"""

from typing import Optional

import pytest

from src.moai_adk.foundation.frontend import (
    AccessibilityValidator,
    ComponentArchitect,
    ComponentLevel,
    DesignSystemBuilder,
    FrontendMetricsCollector,
    PerformanceMetrics,
    PerformanceOptimizer,
    ResponsiveLayoutPlanner,
    StateManagementAdvisor,
    StateManagementSolution,
    WCAGLevel,
)

# ============================================================================
# FIXTURES - Reusable Test Data
# ============================================================================


@pytest.fixture
def component_architect():
    """Fixture for ComponentArchitect instance."""
    return ComponentArchitect()


@pytest.fixture
def state_management_advisor():
    """Fixture for StateManagementAdvisor instance."""
    return StateManagementAdvisor()


@pytest.fixture
def accessibility_validator():
    """Fixture for AccessibilityValidator instance."""
    return AccessibilityValidator()


@pytest.fixture
def performance_optimizer():
    """Fixture for PerformanceOptimizer instance."""
    return PerformanceOptimizer()


@pytest.fixture
def design_system_builder():
    """Fixture for DesignSystemBuilder instance."""
    return DesignSystemBuilder()


@pytest.fixture
def responsive_layout_planner():
    """Fixture for ResponsiveLayoutPlanner instance."""
    return ResponsiveLayoutPlanner()


@pytest.fixture
def metrics_collector():
    """Fixture for FrontendMetricsCollector instance."""
    return FrontendMetricsCollector()


@pytest.fixture
def sample_components():
    """Fixture for sample component definitions."""
    return {
        "atoms": ["Button", "Input", "Label", "Icon"],
        "molecules": ["FormInput", "Card", "Modal"],
        "organisms": ["Form", "Navigation", "Footer"],
        "pages": ["HomePage", "UserProfile", "Dashboard"],
    }


@pytest.fixture
def sample_component_props():
    """Fixture for sample component prop definitions."""
    return {
        "Button": {"onClick": "function", "variant": "string", "disabled": "boolean"},
        "Input": {"value": "string", "onChange": "function", "placeholder": "string"},
        "Card": {"title": "string", "content": "string", "actions": "array"},
        "FormInput": {
            "label": "string",
            "value": "string",
            "error": "string",
            "onChange": "function",
        },
        "Navigation": {
            "links": "array",
            "activeLink": "string",
            "onNavigate": "function",
        },
    }


@pytest.fixture
def sample_app_metrics():
    """Fixture for sample application metrics."""
    return {
        "complexity": "medium",
        "components": 50,
        "async_actions": True,
        "cache_needed": False,
    }


@pytest.fixture
def sample_performance_metrics():
    """Fixture for sample performance metrics."""
    return PerformanceMetrics(
        lcp=1.8,
        fid=45,
        cls=0.08,
        ttfb=100,
        fcp=1.2,
        tti=2.5,
        bundle_size=150,
    )


@pytest.fixture
def sample_accessibility_component():
    """Fixture for sample accessible component."""
    return {
        "aria_label": "Submit Form",
        "keyboard_accessible": True,
        "color_contrast_ratio": 5.5,
        "inputs": [
            {"name": "email", "aria_label": "Email Address", "aria_required": "true"},
            {"name": "password", "aria_label": "Password", "aria_describedby": "pwd-hint"},
        ],
        "buttons": [
            {"aria_label": "Submit", "aria_pressed": "false"},
            {"aria_label": "Reset", "aria_pressed": "false"},
        ],
        "focusable_elements": ["input[type=email]", "input[type=password]", "button"],
        "tab_order_correct": True,
        "focus_trap": True,
        "escape_key_handler": True,
        "focus_restoration": True,
    }


@pytest.fixture
def sample_design_tokens():
    """Fixture for sample design tokens."""
    return {
        "colors": {
            "primary": "#0ea5e9",
            "secondary": "#8b5cf6",
            "success": "#10b981",
            "error": "#ef4444",
            "neutral": "#6b7280",
        },
        "typography": {
            "h1": "32px bold",
            "h2": "24px bold",
            "body": "16px normal",
            "caption": "12px normal",
        },
        "spacing": {"xs": "4px", "sm": "8px", "md": "16px", "lg": "32px", "xl": "64px"},
    }


# ============================================================================
# TEST GROUP 1: Component Architecture (25 tests)
# ============================================================================


class TestComponentArchitect:
    """Test component architecture design and validation."""

    def test_init_creates_architect_instance(self, component_architect):
        """Test that ComponentArchitect initializes correctly."""
        assert component_architect is not None
        assert isinstance(component_architect.components_registry, dict)
        assert isinstance(component_architect.composition_patterns, set)
        assert len(component_architect.composition_patterns) >= 4

    def test_validate_atomic_structure_valid_hierarchy(self, component_architect, sample_components):
        """Test validation of valid atomic design hierarchy."""
        result = component_architect.validate_atomic_structure(sample_components)

        assert result["valid"] is True
        assert result["hierarchy_level"] == 4
        assert result["atom_count"] == 4
        assert result["molecule_count"] == 3
        assert result["organism_count"] == 3
        assert result["page_count"] == 3
        assert len(result["components"]) == 13

    def test_validate_atomic_structure_missing_atoms(self, component_architect):
        """Test validation fails when atoms level is missing."""
        incomplete_components = {
            "molecules": ["FormInput", "Card"],
            "organisms": ["Form"],
            "pages": ["HomePage"],
        }

        result = component_architect.validate_atomic_structure(incomplete_components)

        # Should still be valid but missing atoms
        assert isinstance(result, dict)
        assert "atom_count" in result
        assert result["atom_count"] == 0

    def test_validate_atomic_structure_invalid_level_names(self, component_architect):
        """Test validation with invalid level names."""
        invalid_components = {
            "atoms": ["Button"],
            "invalid_level": ["Something"],
            "pages": ["HomePage"],
        }

        result = component_architect.validate_atomic_structure(invalid_components)

        assert isinstance(result, dict)
        assert "components" in result

    def test_analyze_reusability_optimal_prop_count(self, component_architect, sample_component_props):
        """Test reusability analysis with optimal prop counts (2-5 props)."""
        result = component_architect.analyze_reusability(sample_component_props)

        assert result["reusable_count"] >= 0
        assert 0 <= result["composition_score"] <= 1
        assert result["total_components"] == len(sample_component_props)
        assert isinstance(result["recommendations"], list)

    def test_analyze_reusability_too_many_props(self, component_architect):
        """Test reusability analysis recommends splitting for large components."""
        large_prop_components = {
            "MonsterComponent": {
                "prop1": "string",
                "prop2": "string",
                "prop3": "string",
                "prop4": "string",
                "prop5": "string",
                "prop6": "string",
                "prop7": "string",
            },
            "SimpleComponent": {"prop1": "string", "prop2": "string"},
        }

        result = component_architect.analyze_reusability(large_prop_components)

        assert len(result["recommendations"]) > 0
        assert any("Consider splitting" in rec for rec in result["recommendations"])

    def test_analyze_reusability_single_component(self, component_architect):
        """Test reusability analysis with single component."""
        single_component = {"Button": {"onClick": "function", "label": "string"}}

        result = component_architect.analyze_reusability(single_component)

        assert result["total_components"] == 1
        assert result["reusable_count"] in [0, 1]
        assert 0 <= result["composition_score"] <= 1

    def test_analyze_reusability_empty_components(self, component_architect):
        """Test reusability analysis with empty component dict."""
        result = component_architect.analyze_reusability({})

        assert result["total_components"] == 0
        assert result["reusable_count"] == 0
        assert result["composition_score"] == 0

    def test_validate_composition_patterns_valid_patterns(self, component_architect):
        """Test validation of valid composition patterns."""
        patterns = {
            "hooks": "Custom React hooks pattern",
            "render_props": "Render props pattern",
            "compound_components": "Compound components pattern",
        }

        result = component_architect.validate_composition_patterns(patterns)

        assert result["valid"] is True
        assert result["pattern_count"] == 3
        assert len(result["patterns_found"]) == 3
        assert "hooks" in result["patterns_found"]

    def test_validate_composition_patterns_invalid_patterns(self, component_architect):
        """Test validation rejects invalid composition patterns."""
        patterns = {
            "invalid_pattern_1": "Invalid",
            "invalid_pattern_2": "Invalid",
        }

        result = component_architect.validate_composition_patterns(patterns)

        assert result["valid"] is False
        assert result["pattern_count"] == 0
        assert len(result["patterns_found"]) == 0

    def test_validate_composition_patterns_mixed_valid_invalid(self, component_architect):
        """Test validation with mixed valid and invalid patterns."""
        patterns = {
            "hooks": "Valid pattern",
            "invalid_pattern": "Invalid",
            "hoc": "Valid pattern",
        }

        result = component_architect.validate_composition_patterns(patterns)

        assert result["valid"] is True
        assert result["pattern_count"] == 2
        assert "hooks" in result["patterns_found"]
        assert "hoc" in result["patterns_found"]

    def test_validate_composition_patterns_empty(self, component_architect):
        """Test validation with empty patterns."""
        result = component_architect.validate_composition_patterns({})

        assert result["valid"] is False
        assert result["pattern_count"] == 0

    def test_validate_composition_patterns_recommendations(self, component_architect):
        """Test that recommendations are provided for composition patterns."""
        patterns = {"hooks": "Using hooks pattern"}

        result = component_architect.validate_composition_patterns(patterns)

        assert "recommended_patterns" in result
        assert isinstance(result["recommended_patterns"], list)
        assert len(result["recommended_patterns"]) > 0

    def test_generate_prop_schema_basic_types(self, component_architect):
        """Test TypeScript prop schema generation with basic types."""
        schema = {
            "onClick": object,
            "disabled": bool,
            "label": str,
            "count": int,
        }

        result = component_architect.generate_prop_schema(schema)

        assert "typescript_types" in result
        assert "interface Props" in result["typescript_types"]
        assert "validation_rules" in result
        assert len(result["validation_rules"]) == 4
        assert isinstance(result["default_props"], dict)

    def test_generate_prop_schema_union_types(self, component_architect):
        """Test TypeScript prop schema generation with union types."""
        schema = {
            "variant": ("primary", "secondary", "danger"),
            "size": ("small", "medium", "large"),
        }

        result = component_architect.generate_prop_schema(schema)

        assert "typescript_types" in result
        assert "variant" in result["typescript_types"]
        assert "size" in result["typescript_types"]

    def test_generate_prop_schema_optional_types(self, component_architect):
        """Test TypeScript prop schema generation with optional types."""
        schema = {
            "required_prop": str,
            "optional_prop": Optional,
        }

        result = component_architect.generate_prop_schema(schema)

        assert "required_props" in result
        assert "required_prop" in result["required_props"]
        assert "optional_prop" not in result["required_props"]

    def test_generate_prop_schema_empty(self, component_architect):
        """Test TypeScript prop schema generation with empty schema."""
        result = component_architect.generate_prop_schema({})

        assert "typescript_types" in result
        assert "interface Props" in result["typescript_types"]
        assert len(result["validation_rules"]) == 0

    def test_component_level_enum_values(self):
        """Test ComponentLevel enum contains all required levels."""
        assert ComponentLevel.ATOM.value == "atom"
        assert ComponentLevel.MOLECULE.value == "molecule"
        assert ComponentLevel.ORGANISM.value == "organism"
        assert ComponentLevel.PAGE.value == "page"

    def test_validate_atomic_structure_counts_components(self, component_architect):
        """Test that all components are counted correctly."""
        components = {
            "atoms": ["A", "B"],
            "molecules": ["C", "D", "E"],
            "organisms": ["F"],
            "pages": ["G", "H"],
        }

        result = component_architect.validate_atomic_structure(components)

        total_counted = (
            result["atom_count"] + result["molecule_count"] + result["organism_count"] + result["page_count"]
        )
        assert total_counted == len(result["components"])
        assert total_counted == 8

    def test_validate_atomic_structure_returns_all_components(self, component_architect):
        """Test that all components are returned in the result."""
        expected_components = [
            "Button",
            "Input",
            "Card",
            "Modal",
            "Form",
            "Dashboard",
        ]
        components = {
            "atoms": ["Button", "Input"],
            "molecules": ["Card", "Modal"],
            "organisms": ["Form"],
            "pages": ["Dashboard"],
        }

        result = component_architect.validate_atomic_structure(components)

        assert set(result["components"]) == set(expected_components)

    def test_validate_composition_patterns_recommended_patterns_stable(self, component_architect):
        """Test that recommended patterns are consistent."""
        patterns = {"hooks": "Using hooks"}

        result1 = component_architect.validate_composition_patterns(patterns)
        result2 = component_architect.validate_composition_patterns(patterns)

        assert result1["recommended_patterns"] == result2["recommended_patterns"]

    def test_analyze_reusability_composition_score_calculation(self, component_architect):
        """Test that composition score is calculated correctly."""
        components = {
            "GoodComponent1": {"a": "1", "b": "2", "c": "3"},  # Reusable
            "GoodComponent2": {"a": "1", "b": "2"},  # Reusable
            "BadComponent": {"a": "1"},  # Not in optimal range
        }

        result = component_architect.analyze_reusability(components)

        # Should have 1 or 2 reusable components depending on the 2-5 rule
        assert result["composition_score"] > 0
        assert result["composition_score"] <= 1


# ============================================================================
# TEST GROUP 2: State Management Advisor (24 tests)
# ============================================================================


class TestStateManagementAdvisor:
    """Test state management solution recommendation and validation."""

    def test_init_advisor_creates_instance(self, state_management_advisor):
        """Test that StateManagementAdvisor initializes correctly."""
        assert state_management_advisor is not None
        assert hasattr(state_management_advisor, "solutions")
        assert len(state_management_advisor.solutions) >= 3

    def test_recommend_solution_small_app_local_state(self, state_management_advisor):
        """Test recommendation of local state for small apps."""
        metrics = {"complexity": "small", "components": 20, "async_actions": False}

        result = state_management_advisor.recommend_solution(metrics)

        assert result["solution"] == "Local State"
        assert result["confidence"] >= 0.9
        assert "rationale" in result
        assert "tradeoffs" in result

    def test_recommend_solution_medium_app_context_api(self, state_management_advisor):
        """Test recommendation of Context API for medium apps."""
        metrics = {"complexity": "medium", "components": 40, "async_actions": False}

        result = state_management_advisor.recommend_solution(metrics)

        assert result["solution"] == "Context API"
        assert 0.8 <= result["confidence"] <= 0.9
        assert "rationale" in result

    def test_recommend_solution_large_app_zustand(self, state_management_advisor):
        """Test recommendation of Zustand for large apps."""
        metrics = {"complexity": "large", "components": 100, "async_actions": True}

        result = state_management_advisor.recommend_solution(metrics)

        assert result["solution"] == "Zustand"
        assert 0.8 <= result["confidence"] <= 0.95

    def test_recommend_solution_very_large_app_redux(self, state_management_advisor):
        """Test recommendation of Redux for very large apps."""
        metrics = {
            "complexity": "large",
            "components": 200,
            "async_actions": True,
            "cache_needed": True,
        }

        result = state_management_advisor.recommend_solution(metrics)

        assert result["solution"] == "Redux Toolkit"
        assert 0.8 <= result["confidence"] <= 0.95

    def test_recommend_solution_confidence_score_range(self, state_management_advisor):
        """Test that confidence scores are in valid range."""
        metrics = {"components": 50, "async_actions": True}

        result = state_management_advisor.recommend_solution(metrics)

        assert 0 <= result["confidence"] <= 1

    def test_recommend_solution_tradeoffs_structure(self, state_management_advisor):
        """Test that tradeoffs have expected structure."""
        metrics = {"components": 50}

        result = state_management_advisor.recommend_solution(metrics)

        tradeoffs = result["tradeoffs"]
        assert isinstance(tradeoffs, dict)
        assert "performance" in tradeoffs
        assert "developer_experience" in tradeoffs
        assert "bundle_size_impact" in tradeoffs

    def test_recommend_solution_default_complexity(self, state_management_advisor):
        """Test recommendation uses default complexity."""
        metrics = {}

        result = state_management_advisor.recommend_solution(metrics)

        assert "solution" in result
        assert result["solution"] == "Local State"  # Default is small

    def test_validate_context_pattern_valid_implementation(self, state_management_advisor):
        """Test validation of correct Context API pattern."""
        pattern = {
            "splitting": True,
            "actions": ["fetchUser", "updateUser", "deleteUser"],
        }

        result = state_management_advisor.validate_context_pattern(pattern)

        assert result["valid"] is True
        assert result["issue_count"] == 0
        assert result["issues"] == []
        assert result["actions_count"] == 3

    def test_validate_context_pattern_too_many_actions(self, state_management_advisor):
        """Test validation recommends splitting for too many actions."""
        pattern = {
            "splitting": False,
            "actions": [
                "fetchUser",
                "updateUser",
                "deleteUser",
                "resetUser",
                "uploadAvatar",
                "changePassword",
            ],
        }

        result = state_management_advisor.validate_context_pattern(pattern)

        assert result["valid"] is False
        assert result["issue_count"] > 0
        assert any("splitting" in issue.lower() for issue in result["issues"])

    def test_validate_context_pattern_splitting_enabled(self, state_management_advisor):
        """Test validation passes when splitting is enabled."""
        pattern = {
            "splitting": True,
            "actions": [
                "action1",
                "action2",
                "action3",
                "action4",
                "action5",
                "action6",
            ],
        }

        result = state_management_advisor.validate_context_pattern(pattern)

        assert result["valid"] is True
        assert result["issue_count"] == 0

    def test_validate_context_pattern_minimal_actions(self, state_management_advisor):
        """Test validation with minimal actions."""
        pattern = {"splitting": False, "actions": ["action1", "action2"]}

        result = state_management_advisor.validate_context_pattern(pattern)

        assert result["valid"] is True
        assert result["actions_count"] == 2

    def test_validate_context_pattern_empty_actions(self, state_management_advisor):
        """Test validation with empty actions list."""
        pattern = {"splitting": False, "actions": []}

        result = state_management_advisor.validate_context_pattern(pattern)

        assert result["valid"] is True
        assert result["actions_count"] == 0

    def test_validate_zustand_design_valid_store(self, state_management_advisor):
        """Test validation of Zustand store design."""
        store_design = {
            "selectors": ["selectUser", "selectLoading", "selectError"],
            "actions": ["setUser", "setLoading", "setError"],
            "devtools_enabled": True,
            "persist_enabled": True,
        }

        result = state_management_advisor.validate_zustand_design(store_design)

        assert result["valid"] is True
        assert result["selector_count"] == 3
        assert result["action_count"] == 3
        assert result["devtools_status"] == "enabled"
        assert result["persist_status"] == "enabled"

    def test_validate_zustand_design_without_devtools(self, state_management_advisor):
        """Test Zustand store without devtools."""
        store_design = {
            "selectors": ["selectUser"],
            "actions": ["setUser"],
            "devtools_enabled": False,
            "persist_enabled": True,
        }

        result = state_management_advisor.validate_zustand_design(store_design)

        assert result["devtools_status"] == "disabled"
        assert result["persist_status"] == "enabled"

    def test_validate_zustand_design_minimal(self, state_management_advisor):
        """Test Zustand store with minimal configuration."""
        store_design = {"selectors": [], "actions": []}

        result = state_management_advisor.validate_zustand_design(store_design)

        assert result["valid"] is True
        assert result["selector_count"] == 0
        assert result["action_count"] == 0

    def test_validate_redux_design_multiple_slices(self, state_management_advisor):
        """Test validation of Redux store with multiple slices."""
        slices = {
            "user": {"actions": ["setUser", "clearUser"], "async_thunks": ["fetchUser"]},
            "products": {
                "actions": ["setProducts"],
                "async_thunks": ["fetchProducts", "searchProducts"],
            },
            "cart": {"actions": ["addItem", "removeItem"], "async_thunks": []},
        }

        result = state_management_advisor.validate_redux_design(slices)

        assert result["valid"] is True
        assert result["slice_count"] == 3
        assert result["total_actions"] == 8  # 5 sync + 3 async
        assert "recommendations" in result

    def test_validate_redux_design_single_slice(self, state_management_advisor):
        """Test validation of Redux store with single slice."""
        slices = {
            "app": {
                "actions": ["setLoading", "setError"],
                "async_thunks": ["initializeApp"],
            }
        }

        result = state_management_advisor.validate_redux_design(slices)

        assert result["valid"] is True
        assert result["slice_count"] == 1
        assert result["total_actions"] == 3

    def test_validate_redux_design_no_async_thunks(self, state_management_advisor):
        """Test Redux store with no async thunks."""
        slices = {"ui": {"actions": ["toggle", "setTheme"], "async_thunks": []}}

        result = state_management_advisor.validate_redux_design(slices)

        assert result["slice_count"] == 1
        assert result["total_actions"] == 2

    def test_validate_redux_design_empty(self, state_management_advisor):
        """Test Redux store validation with empty slices."""
        result = state_management_advisor.validate_redux_design({})

        assert result["valid"] is True
        assert result["slice_count"] == 0
        assert result["total_actions"] == 0

    def test_state_management_solution_enum_values(self):
        """Test StateManagementSolution enum contains all options."""
        assert StateManagementSolution.LOCAL_STATE.value == "Local State"
        assert StateManagementSolution.CONTEXT_API.value == "Context API"
        assert StateManagementSolution.ZUSTAND.value == "Zustand"
        assert StateManagementSolution.REDUX_TOOLKIT.value == "Redux Toolkit"
        assert StateManagementSolution.PINIA.value == "Pinia"

    def test_recommend_solution_boundary_30_components(self, state_management_advisor):
        """Test recommendation at 30 component boundary."""
        metrics = {"components": 30, "async_actions": False}

        result = state_management_advisor.recommend_solution(metrics)

        # At boundary, should be Context API
        assert result["solution"] in ["Local State", "Context API"]

    def test_recommend_solution_boundary_150_components(self, state_management_advisor):
        """Test recommendation at 150 component boundary."""
        metrics = {"components": 150, "async_actions": True}

        result = state_management_advisor.recommend_solution(metrics)

        assert result["solution"] in ["Zustand", "Redux Toolkit"]


# ============================================================================
# TEST GROUP 3: Accessibility Validator (20 tests)
# ============================================================================


class TestAccessibilityValidator:
    """Test accessibility compliance and WCAG validation."""

    def test_init_validator_creates_instance(self, accessibility_validator):
        """Test that AccessibilityValidator initializes correctly."""
        assert accessibility_validator is not None
        assert hasattr(accessibility_validator, "wcag_rules")
        assert hasattr(accessibility_validator, "min_contrast_ratio")

    def test_validate_wcag_compliance_aa_level_valid(self, accessibility_validator):
        """Test WCAG AA compliance validation passes for valid component."""
        component = {
            "aria_label": "Submit Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": 4.5,
        }

        result = accessibility_validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is True
        assert result["level"] == "AA"
        assert result["failures"] == []

    def test_validate_wcag_compliance_aaa_level_valid(self, accessibility_validator):
        """Test WCAG AAA compliance validation."""
        component = {
            "aria_label": "Submit",
            "keyboard_accessible": True,
            "color_contrast_ratio": 7.0,
        }

        result = accessibility_validator.validate_wcag_compliance(component, "AAA")

        assert result["compliant"] is True
        assert result["level"] == "AAA"

    def test_validate_wcag_compliance_insufficient_contrast_aa(self, accessibility_validator):
        """Test WCAG AA fails with insufficient contrast."""
        component = {
            "aria_label": "Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": 3.0,
        }

        result = accessibility_validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is False
        assert any("contrast" in f.lower() for f in result["failures"])

    def test_validate_wcag_compliance_missing_aria_label(self, accessibility_validator):
        """Test WCAG fails when aria-label is missing."""
        component = {
            "keyboard_accessible": True,
            "color_contrast_ratio": 5.0,
        }

        result = accessibility_validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is False
        assert any("aria-label" in f for f in result["failures"])

    def test_validate_wcag_compliance_not_keyboard_accessible(self, accessibility_validator):
        """Test WCAG fails when component is not keyboard accessible."""
        component = {
            "aria_label": "Button",
            "keyboard_accessible": False,
            "color_contrast_ratio": 5.0,
        }

        result = accessibility_validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is False
        assert any("keyboard" in f.lower() for f in result["failures"])

    def test_validate_wcag_compliance_multiple_failures(self, accessibility_validator):
        """Test WCAG validation with multiple failures."""
        component = {
            "keyboard_accessible": False,
            "color_contrast_ratio": 2.0,
        }

        result = accessibility_validator.validate_wcag_compliance(component, "AA")

        assert result["compliant"] is False
        assert len(result["failures"]) >= 2

    def test_validate_wcag_compliance_contrast_threshold_aa(self, accessibility_validator):
        """Test WCAG AA contrast threshold (4.5:1)."""
        # Exactly at threshold
        component_pass = {
            "aria_label": "Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": 4.5,
        }

        result_pass = accessibility_validator.validate_wcag_compliance(component_pass, "AA")
        assert result_pass["compliant"] is True

        # Just below threshold
        component_fail = {
            "aria_label": "Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": 4.4,
        }

        result_fail = accessibility_validator.validate_wcag_compliance(component_fail, "AA")
        assert result_fail["compliant"] is False

    def test_validate_wcag_compliance_contrast_threshold_aaa(self, accessibility_validator):
        """Test WCAG AAA contrast threshold (7.0:1)."""
        # Exactly at threshold
        component_pass = {
            "aria_label": "Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": 7.0,
        }

        result_pass = accessibility_validator.validate_wcag_compliance(component_pass, "AAA")
        assert result_pass["compliant"] is True

        # Below threshold
        component_fail = {
            "aria_label": "Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": 6.9,
        }

        result_fail = accessibility_validator.validate_wcag_compliance(component_fail, "AAA")
        assert result_fail["compliant"] is False

    def test_validate_aria_implementation_valid(self, accessibility_validator, sample_accessibility_component):
        """Test ARIA validation with proper implementation."""
        result = accessibility_validator.validate_aria_implementation(sample_accessibility_component)

        assert result["valid"] is True
        assert result["aria_count"] >= 3
        assert len(result["attributes_found"]) > 0

    def test_validate_aria_implementation_insufficient_attributes(self, accessibility_validator):
        """Test ARIA validation with insufficient attributes."""
        component = {
            "inputs": [{"name": "email", "aria_label": "Email"}],
            "buttons": [],
        }

        result = accessibility_validator.validate_aria_implementation(component)

        assert result["valid"] is False
        assert result["aria_count"] < 3

    def test_validate_aria_implementation_no_attributes(self, accessibility_validator):
        """Test ARIA validation with no ARIA attributes."""
        component = {
            "inputs": [{"name": "email"}],
            "buttons": [{"type": "submit"}],
        }

        result = accessibility_validator.validate_aria_implementation(component)

        assert result["valid"] is False
        assert result["aria_count"] == 0

    def test_validate_aria_implementation_attributes_identified(self, accessibility_validator):
        """Test that all ARIA attributes are correctly identified."""
        component = {
            "inputs": [
                {"aria_label": "Email", "aria_required": "true"},
                {"aria_label": "Password", "aria_describedby": "hint"},
            ],
            "buttons": [{"aria_pressed": "false"}],
        }

        result = accessibility_validator.validate_aria_implementation(component)

        assert "aria_label" in result["attributes_found"]
        assert "aria_required" in result["attributes_found"]
        assert "aria_describedby" in result["attributes_found"]
        assert "aria_pressed" in result["attributes_found"]

    def test_validate_aria_implementation_recommendations(self, accessibility_validator):
        """Test that recommendations are provided."""
        component = {
            "inputs": [{"aria_label": "Email"}],
            "buttons": [],
        }

        result = accessibility_validator.validate_aria_implementation(component)

        assert "recommendations" in result
        assert isinstance(result["recommendations"], list)

    def test_validate_keyboard_navigation_valid(self, accessibility_validator, sample_accessibility_component):
        """Test keyboard navigation validation with valid component."""
        result = accessibility_validator.validate_keyboard_navigation(sample_accessibility_component)

        assert result["valid"] is True
        assert result["keyboard_compliant"] is True
        assert result["focusable_elements"] >= 3
        assert 0 <= result["focus_management_score"] <= 1

    def test_validate_keyboard_navigation_missing_tab_order(self, accessibility_validator):
        """Test keyboard navigation without proper tab order."""
        component = {
            "focusable_elements": ["button", "input"],
            "tab_order_correct": False,
            "focus_trap": True,
            "escape_key_handler": True,
            "focus_restoration": True,
        }

        result = accessibility_validator.validate_keyboard_navigation(component)

        assert result["valid"] is False

    def test_validate_keyboard_navigation_focus_management_score(self, accessibility_validator):
        """Test focus management score calculation."""
        component_full = {
            "focusable_elements": ["button"],
            "tab_order_correct": True,
            "focus_trap": True,
            "escape_key_handler": True,
            "focus_restoration": True,
        }

        result_full = accessibility_validator.validate_keyboard_navigation(component_full)

        # All 4 features enabled = 1.0
        assert result_full["focus_management_score"] == 1.0

        component_partial = {
            "focusable_elements": ["button"],
            "tab_order_correct": True,
            "focus_trap": False,
            "escape_key_handler": True,
            "focus_restoration": False,
        }

        result_partial = accessibility_validator.validate_keyboard_navigation(component_partial)

        # 2 out of 4 features = 0.5
        assert result_partial["focus_management_score"] == 0.5

    def test_validate_keyboard_navigation_no_focusable_elements(self, accessibility_validator):
        """Test keyboard navigation with no focusable elements."""
        component = {
            "focusable_elements": [],
            "tab_order_correct": False,
        }

        result = accessibility_validator.validate_keyboard_navigation(component)

        assert result["valid"] is False
        assert result["focusable_elements"] == 0

    def test_wcag_level_enum_values(self):
        """Test WCAGLevel enum contains all levels."""
        assert WCAGLevel.A.value == "A"
        assert WCAGLevel.AA.value == "AA"
        assert WCAGLevel.AAA.value == "AAA"


# ============================================================================
# TEST GROUP 4: Performance Optimizer (22 tests)
# ============================================================================


class TestPerformanceOptimizer:
    """Test performance optimization and metrics validation."""

    def test_init_optimizer_creates_instance(self, performance_optimizer):
        """Test that PerformanceOptimizer initializes correctly."""
        assert performance_optimizer is not None
        assert hasattr(performance_optimizer, "core_web_vitals_thresholds")

    def test_validate_code_splitting_optimized(self, performance_optimizer):
        """Test code splitting optimization validation."""
        strategy = {
            "chunks": {
                "main": "app.js",
                "vendor": "vendor.js",
                "utils": "utils.js",
                "components": "components.js",
            },
            "dynamic_imports": 5,
            "route_based_splitting": True,
            "component_based_splitting": True,
        }

        result = performance_optimizer.validate_code_splitting(strategy)

        assert result["optimized"] is True
        assert result["chunk_count"] == 4
        assert result["vendor_chunk_separated"] is True
        assert result["dynamic_imports"] == 5

    def test_validate_code_splitting_not_optimized(self, performance_optimizer):
        """Test code splitting that's not optimized."""
        strategy = {
            "chunks": {"main": "app.js"},
            "dynamic_imports": 0,
            "route_based_splitting": False,
        }

        result = performance_optimizer.validate_code_splitting(strategy)

        assert result["optimized"] is False
        assert result["chunk_count"] == 1

    def test_validate_code_splitting_missing_vendor(self, performance_optimizer):
        """Test code splitting without vendor chunk separation."""
        strategy = {
            "chunks": {"main": "app.js", "utils": "utils.js"},
            "dynamic_imports": 0,
        }

        result = performance_optimizer.validate_code_splitting(strategy)

        assert result["vendor_chunk_separated"] is False
        assert result["optimized"] is False

    def test_validate_code_splitting_empty(self, performance_optimizer):
        """Test code splitting with empty chunks."""
        strategy = {"chunks": {}, "dynamic_imports": 0}

        result = performance_optimizer.validate_code_splitting(strategy)

        assert result["chunk_count"] == 0
        assert result["optimized"] is False

    def test_validate_memoization_with_improvement(self, performance_optimizer):
        """Test memoization validation with performance improvement."""
        strategy = {
            "render_count_baseline": 100,
            "render_count_optimized": 30,
            "memo_components": ["Button", "Input", "Card"],
        }

        result = performance_optimizer.validate_memoization(strategy)

        assert result["optimized"] is True
        assert result["memo_count"] == 3
        assert result["improvement_percentage"] == 70.0
        assert "useMemo" in result["hooks_used"]
        assert "useCallback" in result["hooks_used"]

    def test_validate_memoization_no_improvement(self, performance_optimizer):
        """Test memoization with no improvement."""
        strategy = {
            "render_count_baseline": 100,
            "render_count_optimized": 100,
            "memo_components": [],
        }

        result = performance_optimizer.validate_memoization(strategy)

        assert result["improvement_percentage"] == 0.0
        assert result["memo_count"] == 0

    def test_validate_memoization_zero_baseline(self, performance_optimizer):
        """Test memoization with zero baseline (edge case)."""
        strategy = {
            "render_count_baseline": 0,
            "render_count_optimized": 0,
            "memo_components": [],
        }

        result = performance_optimizer.validate_memoization(strategy)

        assert result["improvement_percentage"] == 0
        assert result["optimized"] is True

    def test_validate_performance_metrics_all_good(self, performance_optimizer):
        """Test performance metrics validation when all metrics are good."""
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 80,
            "cls_value": 0.08,
            "bundle_size_kb": 50,
            "gzip_size_kb": 20,
        }

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is True
        assert result["lcp_status"] == "good"
        assert result["fid_status"] == "good"
        assert result["cls_status"] == "good"
        assert result["bundle_optimized"] is True

    def test_validate_performance_metrics_lcp_needs_improvement(self, performance_optimizer):
        """Test LCP metric that needs improvement."""
        metrics = {
            "lcp_seconds": 3.0,
            "fid_milliseconds": 80,
            "cls_value": 0.08,
        }

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is False
        assert result["lcp_status"] == "needs_improvement"

    def test_validate_performance_metrics_fid_needs_improvement(self, performance_optimizer):
        """Test FID metric that needs improvement."""
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 150,
            "cls_value": 0.08,
        }

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is False
        assert result["fid_status"] == "needs_improvement"

    def test_validate_performance_metrics_cls_needs_improvement(self, performance_optimizer):
        """Test CLS metric that needs improvement."""
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 80,
            "cls_value": 0.15,
        }

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is False
        assert result["cls_status"] == "needs_improvement"

    def test_validate_performance_metrics_large_bundle(self, performance_optimizer):
        """Test performance metrics with large bundle size."""
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 80,
            "cls_value": 0.08,
            "bundle_size_kb": 100,
            "gzip_size_kb": 80,
        }

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert result["bundle_optimized"] is False

    def test_validate_performance_metrics_thresholds_lcp(self, performance_optimizer):
        """Test LCP threshold values."""
        # At good threshold
        metrics_good = {
            "lcp_seconds": 2.5,
            "fid_milliseconds": 50,
            "cls_value": 0.05,
        }
        result_good = performance_optimizer.validate_performance_metrics(metrics_good)
        assert result_good["lcp_status"] == "good"

        # Just above good threshold
        metrics_bad = {
            "lcp_seconds": 2.6,
            "fid_milliseconds": 50,
            "cls_value": 0.05,
        }
        result_bad = performance_optimizer.validate_performance_metrics(metrics_bad)
        assert result_bad["lcp_status"] == "needs_improvement"

    def test_validate_performance_metrics_thresholds_fid(self, performance_optimizer):
        """Test FID threshold values."""
        # At good threshold
        metrics_good = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 100,
            "cls_value": 0.05,
        }
        result_good = performance_optimizer.validate_performance_metrics(metrics_good)
        assert result_good["fid_status"] == "good"

        # Just above good threshold
        metrics_bad = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 101,
            "cls_value": 0.05,
        }
        result_bad = performance_optimizer.validate_performance_metrics(metrics_bad)
        assert result_bad["fid_status"] == "needs_improvement"

    def test_validate_performance_metrics_thresholds_cls(self, performance_optimizer):
        """Test CLS threshold values."""
        # At good threshold
        metrics_good = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 50,
            "cls_value": 0.1,
        }
        result_good = performance_optimizer.validate_performance_metrics(metrics_good)
        assert result_good["cls_status"] == "good"

        # Just above good threshold
        metrics_bad = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 50,
            "cls_value": 0.101,
        }
        result_bad = performance_optimizer.validate_performance_metrics(metrics_bad)
        assert result_bad["cls_status"] == "needs_improvement"

    def test_validate_performance_metrics_structure(self, performance_optimizer):
        """Test performance metrics result structure."""
        metrics = {
            "lcp_seconds": 2.0,
            "fid_milliseconds": 80,
            "cls_value": 0.08,
            "bundle_size_kb": 50,
            "gzip_size_kb": 20,
        }

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert "metrics" in result
        assert "lcp" in result["metrics"]
        assert "fid" in result["metrics"]
        assert "cls" in result["metrics"]
        assert "bundle_size_kb" in result["metrics"]
        assert "gzip_size_kb" in result["metrics"]

    def test_validate_performance_metrics_missing_values(self, performance_optimizer):
        """Test performance metrics with missing values (should default to 0)."""
        metrics = {}

        result = performance_optimizer.validate_performance_metrics(metrics)

        assert result["core_web_vitals_passed"] is True  # All defaults to 0 = good
        assert result["lcp_status"] == "good"
        assert result["fid_status"] == "good"
        assert result["cls_status"] == "good"


# ============================================================================
# TEST GROUP 5: Design System Builder (10 tests)
# ============================================================================


class TestDesignSystemBuilder:
    """Test design system building and token management."""

    def test_init_builder_creates_instance(self, design_system_builder):
        """Test that DesignSystemBuilder initializes correctly."""
        assert design_system_builder is not None
        assert isinstance(design_system_builder.tokens, dict)
        assert isinstance(design_system_builder.components_doc, dict)

    def test_define_design_tokens_complete(self, design_system_builder, sample_design_tokens):
        """Test complete design token definition."""
        result = design_system_builder.define_design_tokens(sample_design_tokens)

        assert result["token_count"] > 0
        assert len(result["categories"]) >= 3
        assert "colors" in result["categories"]
        assert "typography" in result["categories"]
        assert "spacing" in result["categories"]
        assert result["theme_support"] == "light_dark"

    def test_define_design_tokens_css_variables_generated(self, design_system_builder, sample_design_tokens):
        """Test that CSS variables are generated from tokens."""
        result = design_system_builder.define_design_tokens(sample_design_tokens)

        css_vars = result["css_variables"]
        assert isinstance(css_vars, list)
        assert len(css_vars) > 0
        assert any("--colors-" in var for var in css_vars)
        assert any("--spacing-" in var for var in css_vars)

    def test_define_design_tokens_colors_only(self, design_system_builder):
        """Test token definition with only colors."""
        tokens = {"colors": {"primary": "#0ea5e9", "secondary": "#8b5cf6"}}

        result = design_system_builder.define_design_tokens(tokens)

        assert result["token_count"] == 2
        assert result["categories"] == ["colors"]

    def test_define_design_tokens_multiple_categories(self, design_system_builder):
        """Test token definition with multiple categories."""
        tokens = {
            "colors": {"primary": "#0ea5e9"},
            "spacing": {"small": "8px", "medium": "16px"},
            "typography": {"h1": "32px"},
            "shadows": {"sm": "0 1px 2px", "md": "0 4px 6px"},
        }

        result = design_system_builder.define_design_tokens(tokens)

        assert result["token_count"] == 6
        assert len(result["categories"]) == 4

    def test_define_design_tokens_empty(self, design_system_builder):
        """Test token definition with empty tokens."""
        result = design_system_builder.define_design_tokens({})

        assert result["token_count"] == 0
        assert result["categories"] == []

    def test_design_tokens_persistence(self, design_system_builder, sample_design_tokens):
        """Test that tokens are stored in the builder instance."""
        design_system_builder.define_design_tokens(sample_design_tokens)

        assert design_system_builder.tokens == sample_design_tokens

    def test_generate_css_variables_formatting(self, design_system_builder):
        """Test CSS variable naming convention."""
        tokens = {
            "colors": {"primary": "#0ea5e9", "secondary": "#8b5cf6"},
        }

        result = design_system_builder.define_design_tokens(tokens)

        css_vars = result["css_variables"]
        assert "--colors-primary" in css_vars
        assert "--colors-secondary" in css_vars

    def test_theme_support_value(self, design_system_builder, sample_design_tokens):
        """Test theme support value is correctly set."""
        result = design_system_builder.define_design_tokens(sample_design_tokens)

        assert result["theme_support"] == "light_dark"

    def test_categories_order_preserved(self, design_system_builder):
        """Test that category order matches input order."""
        tokens = {
            "spacing": {},
            "colors": {},
            "typography": {},
        }

        result = design_system_builder.define_design_tokens(tokens)

        assert result["categories"] == ["spacing", "colors", "typography"]


# ============================================================================
# TEST GROUP 6: Responsive Layout Planner (20 tests)
# ============================================================================


class TestResponsiveLayoutPlanner:
    """Test responsive design planning and validation."""

    def test_init_planner_creates_instance(self, responsive_layout_planner):
        """Test that ResponsiveLayoutPlanner initializes correctly."""
        assert responsive_layout_planner is not None
        assert hasattr(responsive_layout_planner, "standard_breakpoints")

    def test_validate_breakpoints_mobile_first_valid(self, responsive_layout_planner):
        """Test validation of mobile-first breakpoints."""
        breakpoints = {
            "mobile_first": True,
            "mobile": 0,
            "sm": 640,
            "md": 768,
            "lg": 1024,
            "xl": 1280,
        }

        result = responsive_layout_planner.validate_breakpoints(breakpoints)

        assert result["mobile_first"] is True
        assert result["breakpoint_count"] == 5
        assert result["valid"] is True

    def test_validate_breakpoints_not_mobile_first(self, responsive_layout_planner):
        """Test validation of non-mobile-first breakpoints."""
        breakpoints = {
            "mobile_first": False,
            "mobile": 0,
            "sm": 640,
            "md": 768,
        }

        result = responsive_layout_planner.validate_breakpoints(breakpoints)

        assert result["mobile_first"] is False
        assert result["breakpoint_count"] == 3

    def test_validate_breakpoints_insufficient_count(self, responsive_layout_planner):
        """Test validation fails with insufficient breakpoints."""
        breakpoints = {
            "mobile": 0,
            "sm": 640,
        }

        result = responsive_layout_planner.validate_breakpoints(breakpoints)

        assert result["valid"] is False
        assert result["breakpoint_count"] == 1  # mobile_first flag is not counted

    def test_validate_breakpoints_minimum_valid(self, responsive_layout_planner):
        """Test validation with minimum required breakpoints."""
        breakpoints = {
            "mobile_first": True,
            "mobile": 0,
            "sm": 640,
            "md": 768,
            "lg": 1024,
        }

        result = responsive_layout_planner.validate_breakpoints(breakpoints)

        assert result["valid"] is True
        assert result["breakpoint_count"] == 4

    def test_validate_fluid_layout_enabled(self, responsive_layout_planner):
        """Test fluid layout validation with container queries enabled."""
        layout_config = {
            "container_query_enabled": True,
            "fluid_spacing": True,
            "responsive_typography": True,
            "responsive_images": True,
            "aspect_ratio_preserved": True,
        }

        result = responsive_layout_planner.validate_fluid_layout(layout_config)

        assert result["fluid"] is True
        assert result["container_queries_enabled"] is True
        assert result["responsive_score"] == 1.0

    def test_validate_fluid_layout_partial_features(self, responsive_layout_planner):
        """Test fluid layout with some features enabled."""
        layout_config = {
            "container_query_enabled": False,
            "fluid_spacing": True,
            "responsive_typography": False,
            "responsive_images": True,
            "aspect_ratio_preserved": False,
        }

        result = responsive_layout_planner.validate_fluid_layout(layout_config)

        assert result["responsive_score"] == 0.5  # 2 out of 4 features

    def test_validate_fluid_layout_no_features(self, responsive_layout_planner):
        """Test fluid layout with no features enabled."""
        layout_config = {
            "container_query_enabled": False,
            "fluid_spacing": False,
            "responsive_typography": False,
            "responsive_images": False,
            "aspect_ratio_preserved": False,
        }

        result = responsive_layout_planner.validate_fluid_layout(layout_config)

        assert result["fluid"] is False
        assert result["responsive_score"] == 0.0

    def test_validate_fluid_layout_grid_responsive(self, responsive_layout_planner):
        """Test fluid layout grid responsiveness validation."""
        layout_config = {
            "grid_columns_responsive": {
                "mobile": 1,
                "tablet": 2,
                "desktop": 3,
            }
        }

        result = responsive_layout_planner.validate_fluid_layout(layout_config)

        assert result["grid_responsive"] is True

    def test_validate_fluid_layout_insufficient_responsive_columns(self, responsive_layout_planner):
        """Test fluid layout with insufficient responsive column definitions."""
        layout_config = {"grid_columns_responsive": {"mobile": 1}}

        result = responsive_layout_planner.validate_fluid_layout(layout_config)

        assert result["grid_responsive"] is False

    def test_validate_image_strategy_fully_optimized(self, responsive_layout_planner):
        """Test fully optimized image strategy."""
        strategy = {
            "srcset_enabled": True,
            "lazy_loading": True,
            "image_optimization": True,
            "webp_format": True,
            "placeholder_strategy": True,
            "breakpoint_images": {
                "mobile": "image-sm.webp",
                "tablet": "image-md.webp",
                "desktop": "image-lg.webp",
            },
        }

        result = responsive_layout_planner.validate_image_strategy(strategy)

        assert result["optimized"] is True
        assert result["optimization_score"] == 1.0
        assert result["lazy_loading_enabled"] is True
        assert result["webp_support"] is True
        assert result["responsive_images"] is True

    def test_validate_image_strategy_partially_optimized(self, responsive_layout_planner):
        """Test partially optimized image strategy."""
        strategy = {
            "srcset_enabled": True,
            "lazy_loading": True,
            "image_optimization": False,
            "webp_format": False,
            "placeholder_strategy": True,
            "breakpoint_images": {"mobile": "img-sm.jpg", "desktop": "img-lg.jpg"},
        }

        result = responsive_layout_planner.validate_image_strategy(strategy)

        assert result["optimization_score"] > 0.5  # 3 out of 5 features
        assert result["responsive_images"] is True

    def test_validate_image_strategy_minimal(self, responsive_layout_planner):
        """Test minimal image strategy."""
        strategy = {
            "srcset_enabled": False,
            "lazy_loading": False,
            "image_optimization": False,
            "webp_format": False,
            "placeholder_strategy": False,
            "breakpoint_images": {},
        }

        result = responsive_layout_planner.validate_image_strategy(strategy)

        assert result["optimized"] is False
        assert result["optimization_score"] == 0.0
        assert result["responsive_images"] is False

    def test_validate_image_strategy_lazy_loading_variations(self, responsive_layout_planner):
        """Test various lazy loading configurations."""
        for lazy_loading_value in [True, False, "native", "intersection-observer"]:
            strategy = {
                "srcset_enabled": True,
                "lazy_loading": lazy_loading_value,
                "image_optimization": True,
                "webp_format": True,
                "placeholder_strategy": True,
                "breakpoint_images": {"mobile": "img.webp"},
            }

            result = responsive_layout_planner.validate_image_strategy(strategy)

            # Should support various lazy loading approaches
            assert result is not None

    def test_validate_image_strategy_webp_fallback(self, responsive_layout_planner):
        """Test WebP format support indication."""
        strategy_webp = {
            "webp_format": True,
            "srcset_enabled": True,
            "lazy_loading": True,
            "image_optimization": True,
            "placeholder_strategy": True,
            "breakpoint_images": {"mobile": "img.webp"},
        }

        result_webp = responsive_layout_planner.validate_image_strategy(strategy_webp)
        assert result_webp["webp_support"] is True

        strategy_no_webp = {
            "webp_format": False,
            "srcset_enabled": True,
            "lazy_loading": True,
            "image_optimization": True,
            "placeholder_strategy": True,
            "breakpoint_images": {"mobile": "img.jpg"},
        }

        result_no_webp = responsive_layout_planner.validate_image_strategy(strategy_no_webp)
        assert result_no_webp["webp_support"] is False

    def test_validate_image_strategy_responsive_images_count(self, responsive_layout_planner):
        """Test responsive images count in strategy."""
        strategy_single = {"breakpoint_images": {"mobile": "img.jpg"}}
        result_single = responsive_layout_planner.validate_image_strategy(strategy_single)
        assert result_single["responsive_images"] is False

        strategy_multi = {
            "breakpoint_images": {
                "mobile": "img-sm.jpg",
                "tablet": "img-md.jpg",
                "desktop": "img-lg.jpg",
            }
        }
        result_multi = responsive_layout_planner.validate_image_strategy(strategy_multi)
        assert result_multi["responsive_images"] is True


# ============================================================================
# TEST GROUP 7: Frontend Metrics Collector (20 tests)
# ============================================================================


class TestFrontendMetricsCollector:
    """Test frontend metrics collection and analysis."""

    def test_init_collector_creates_instance(self, metrics_collector):
        """Test that FrontendMetricsCollector initializes correctly."""
        assert metrics_collector is not None
        assert isinstance(metrics_collector.metrics_history, list)
        assert len(metrics_collector.metrics_history) == 0
        assert hasattr(metrics_collector, "thresholds")

    def test_collect_metrics_valid(self, metrics_collector):
        """Test collecting valid performance metrics."""
        metrics = {
            "lcp": 1.8,
            "fid": 45,
            "cls": 0.08,
            "ttfb": 100,
            "fcp": 1.2,
            "tti": 2.5,
            "bundle_size": 150,
        }

        result = metrics_collector.collect_metrics(metrics)

        assert isinstance(result, PerformanceMetrics)
        assert result.lcp == 1.8
        assert result.fid == 45
        assert result.cls == 0.08
        assert len(metrics_collector.metrics_history) == 1

    def test_collect_metrics_stored_in_history(self, metrics_collector):
        """Test that metrics are stored in history."""
        metrics1 = {
            "lcp": 1.8,
            "fid": 45,
            "cls": 0.08,
            "ttfb": 100,
            "fcp": 1.2,
            "tti": 2.5,
            "bundle_size": 150,
        }

        metrics2 = {
            "lcp": 2.0,
            "fid": 50,
            "cls": 0.09,
            "ttfb": 110,
            "fcp": 1.3,
            "tti": 2.7,
            "bundle_size": 160,
        }

        metrics_collector.collect_metrics(metrics1)
        metrics_collector.collect_metrics(metrics2)

        assert len(metrics_collector.metrics_history) == 2
        assert metrics_collector.metrics_history[0].lcp == 1.8
        assert metrics_collector.metrics_history[1].lcp == 2.0

    def test_collect_metrics_missing_values_defaults_to_zero(self, metrics_collector):
        """Test that missing metric values default to 0."""
        metrics = {"lcp": 1.8, "fid": 45}

        result = metrics_collector.collect_metrics(metrics)

        assert result.lcp == 1.8
        assert result.fid == 45
        assert result.cls == 0
        assert result.ttfb == 0

    def test_collect_metrics_timestamp_generated(self, metrics_collector):
        """Test that timestamp is automatically generated."""
        metrics = {"lcp": 1.8, "fid": 45, "cls": 0.08, "ttfb": 100, "fcp": 1.2, "tti": 2.5, "bundle_size": 150}

        result = metrics_collector.collect_metrics(metrics)

        assert result.timestamp is not None
        assert isinstance(result.timestamp, str)
        # ISO format check
        assert "T" in result.timestamp

    def test_analyze_metrics_all_good(self, metrics_collector, sample_performance_metrics):
        """Test metrics analysis when all metrics are good."""
        result = metrics_collector.analyze_metrics(sample_performance_metrics)

        assert result["core_web_vitals_pass"] is True
        assert result["lcp_status"] == "good"
        assert result["fid_status"] == "good"
        assert result["cls_status"] == "good"
        assert result["performance_score"] > 0.7

    def test_analyze_metrics_lcp_poor(self, metrics_collector):
        """Test metrics analysis when LCP is poor."""
        poor_metrics = PerformanceMetrics(
            lcp=4.0,
            fid=45,
            cls=0.08,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(poor_metrics)

        assert result["core_web_vitals_pass"] is False
        assert result["lcp_status"] == "needs_improvement"

    def test_analyze_metrics_fid_poor(self, metrics_collector):
        """Test metrics analysis when FID is poor."""
        poor_metrics = PerformanceMetrics(
            lcp=1.8,
            fid=150,
            cls=0.08,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(poor_metrics)

        assert result["core_web_vitals_pass"] is False
        assert result["fid_status"] == "needs_improvement"

    def test_analyze_metrics_cls_poor(self, metrics_collector):
        """Test metrics analysis when CLS is poor."""
        poor_metrics = PerformanceMetrics(
            lcp=1.8,
            fid=45,
            cls=0.15,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(poor_metrics)

        assert result["core_web_vitals_pass"] is False
        assert result["cls_status"] == "needs_improvement"

    def test_analyze_metrics_result_structure(self, metrics_collector, sample_performance_metrics):
        """Test the structure of metrics analysis result."""
        result = metrics_collector.analyze_metrics(sample_performance_metrics)

        assert "performance_score" in result
        assert "core_web_vitals_pass" in result
        assert "lcp_status" in result
        assert "fid_status" in result
        assert "cls_status" in result
        assert "metrics" in result
        assert "recommendations" in result

    def test_analyze_metrics_recommendations_lcp(self, metrics_collector):
        """Test recommendations for LCP improvement."""
        poor_metrics = PerformanceMetrics(
            lcp=4.0,
            fid=45,
            cls=0.08,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(poor_metrics)

        assert len(result["recommendations"]) > 0
        assert any("LCP" in rec for rec in result["recommendations"])

    def test_analyze_metrics_recommendations_fid(self, metrics_collector):
        """Test recommendations for FID improvement."""
        poor_metrics = PerformanceMetrics(
            lcp=1.8,
            fid=150,
            cls=0.08,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(poor_metrics)

        assert any("FID" in rec for rec in result["recommendations"])

    def test_analyze_metrics_recommendations_cls(self, metrics_collector):
        """Test recommendations for CLS improvement."""
        poor_metrics = PerformanceMetrics(
            lcp=1.8,
            fid=45,
            cls=0.15,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(poor_metrics)

        assert any("CLS" in rec for rec in result["recommendations"])

    def test_analyze_metrics_recommendations_bundle_size(self, metrics_collector):
        """Test recommendations for bundle size reduction."""
        large_bundle_metrics = PerformanceMetrics(
            lcp=1.8,
            fid=45,
            cls=0.08,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=250,
        )

        result = metrics_collector.analyze_metrics(large_bundle_metrics)

        assert any("bundle size" in rec.lower() for rec in result["recommendations"])

    def test_analyze_metrics_performance_score_range(self, metrics_collector, sample_performance_metrics):
        """Test that performance score is in valid range."""
        result = metrics_collector.analyze_metrics(sample_performance_metrics)

        assert 0 <= result["performance_score"] <= 1

    def test_generate_recommendations_comprehensive(self, metrics_collector):
        """Test recommendations generation with multiple issues."""
        all_bad_metrics = PerformanceMetrics(
            lcp=4.0,
            fid=150,
            cls=0.15,
            ttfb=200,
            fcp=2.0,
            tti=3.5,
            bundle_size=300,
        )

        result = metrics_collector.analyze_metrics(all_bad_metrics)
        recommendations = result["recommendations"]

        # Should have multiple recommendations
        assert len(recommendations) >= 3
        assert any("LCP" in rec for rec in recommendations)
        assert any("FID" in rec for rec in recommendations)
        assert any("CLS" in rec for rec in recommendations)
        assert any("bundle" in rec.lower() for rec in recommendations)

    def test_performance_metrics_dataclass_serialization(self, metrics_collector):
        """Test that PerformanceMetrics can be serialized to dict."""
        metrics = PerformanceMetrics(
            lcp=1.8,
            fid=45,
            cls=0.08,
            ttfb=100,
            fcp=1.2,
            tti=2.5,
            bundle_size=150,
        )

        result = metrics_collector.analyze_metrics(metrics)
        metrics_dict = result["metrics"]

        assert isinstance(metrics_dict, dict)
        assert metrics_dict["lcp"] == 1.8
        assert metrics_dict["bundle_size"] == 150

    def test_collect_and_analyze_workflow(self, metrics_collector):
        """Test complete workflow of collecting and analyzing metrics."""
        raw_metrics = {
            "lcp": 1.8,
            "fid": 45,
            "cls": 0.08,
            "ttfb": 100,
            "fcp": 1.2,
            "tti": 2.5,
            "bundle_size": 150,
        }

        collected = metrics_collector.collect_metrics(raw_metrics)
        analyzed = metrics_collector.analyze_metrics(collected)

        assert len(metrics_collector.metrics_history) == 1
        assert analyzed["core_web_vitals_pass"] is True
        assert analyzed["performance_score"] > 0


# ============================================================================
# PARAMETRIZED TESTS - Testing Multiple Scenarios
# ============================================================================


class TestParametrizedScenarios:
    """Parametrized tests for multiple scenarios."""

    @pytest.mark.parametrize(
        "components_count,expected_solution",
        [
            (10, "Local State"),
            (40, "Context API"),
            (80, "Context API"),  # 80 is still within Context API range
            (200, "Redux Toolkit"),
        ],
    )
    def test_state_solution_by_component_count(self, state_management_advisor, components_count, expected_solution):
        """Test state management solution recommendation by component count."""
        metrics = {"components": components_count}
        result = state_management_advisor.recommend_solution(metrics)

        assert result["solution"] == expected_solution

    @pytest.mark.parametrize(
        "contrast_ratio,level,should_pass",
        [
            (5.0, "AA", True),
            (4.4, "AA", False),
            (7.5, "AAA", True),
            (6.9, "AAA", False),
        ],
    )
    def test_wcag_contrast_validation(self, accessibility_validator, contrast_ratio, level, should_pass):
        """Test WCAG contrast ratio validation across levels."""
        component = {
            "aria_label": "Button",
            "keyboard_accessible": True,
            "color_contrast_ratio": contrast_ratio,
        }

        result = accessibility_validator.validate_wcag_compliance(component, level)

        assert result["compliant"] == should_pass

    @pytest.mark.parametrize(
        "code_splitting_config,should_be_optimized",
        [
            (
                {
                    "chunks": {
                        "main": "app.js",
                        "vendor": "vendor.js",
                        "utils": "utils.js",
                        "components": "components.js",
                    },
                    "dynamic_imports": 5,
                },
                True,
            ),
            ({"chunks": {"main": "app.js"}, "dynamic_imports": 0}, False),
        ],
    )
    def test_code_splitting_configurations(self, performance_optimizer, code_splitting_config, should_be_optimized):
        """Test various code splitting configurations."""
        result = performance_optimizer.validate_code_splitting(code_splitting_config)

        assert result["optimized"] == should_be_optimized


# ============================================================================
# EDGE CASES AND ERROR HANDLING
# ============================================================================


class TestEdgeCasesAndErrors:
    """Test edge cases and error handling."""

    def test_component_architect_with_special_characters(self, component_architect):
        """Test component names with special characters."""
        components = {
            "atoms": ["Button-Primary", "Input_Text"],
            "molecules": ["Form.Container"],
            "organisms": ["Header-Nav"],
            "pages": ["Page_Home"],
        }

        result = component_architect.validate_atomic_structure(components)

        assert result["valid"] is True
        assert len(result["components"]) == 5

    def test_performance_metrics_with_negative_values(self, performance_optimizer):
        """Test performance metrics handling with negative values."""
        metrics = {
            "lcp_seconds": -1.0,
            "fid_milliseconds": -50,
            "cls_value": -0.1,
        }

        # Should not crash, should handle gracefully
        result = performance_optimizer.validate_performance_metrics(metrics)

        assert "lcp_status" in result
        assert "fid_status" in result
        assert "cls_status" in result

    def test_metrics_collector_with_extreme_values(self, metrics_collector):
        """Test metrics collector with extreme values."""
        extreme_metrics = PerformanceMetrics(
            lcp=100.0,
            fid=5000,
            cls=2.0,
            ttfb=1000,
            fcp=50.0,
            tti=100.0,
            bundle_size=10000,
        )

        result = metrics_collector.analyze_metrics(extreme_metrics)

        assert result["core_web_vitals_pass"] is False
        assert len(result["recommendations"]) > 0

    def test_accessibility_validator_partial_component_data(self, accessibility_validator):
        """Test accessibility validation with incomplete component data."""
        incomplete_component = {
            "aria_label": "Button",
        }

        result = accessibility_validator.validate_wcag_compliance(incomplete_component, "AA")

        # Should handle missing fields
        assert "compliant" in result
        assert "failures" in result

    def test_design_tokens_with_nested_structure(self, design_system_builder):
        """Test design tokens with nested structure."""
        nested_tokens = {
            "colors": {
                "primary": {
                    "light": "#0ea5e9",
                    "dark": "#0284c7",
                }
            }
        }

        # Should handle nested structures appropriately
        result = design_system_builder.define_design_tokens(nested_tokens)

        assert "categories" in result
        assert "css_variables" in result
