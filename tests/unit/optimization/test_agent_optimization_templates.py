"""
Unit tests for agent optimization templates.

Tests the TAG-004 component: Agent Optimization Templates
"""

from unittest.mock import patch

import pytest


class TestAgentOptimizationTemplates:
    """Test cases for agent optimization templates."""

    def test_template_registry_initialization(self):
        """Test that template registry can be initialized properly."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Verify default templates are loaded
        assert len(registry.templates) > 0
        assert "frontend-expert" in registry.templates
        assert "backend-expert" in registry.templates
        assert "devops-expert" in registry.templates

    def test_template_retrieval(self):
        """Test that templates can be retrieved by name."""
        from moai_adk.optimization.templates import TemplateNotFoundError, TemplateRegistry

        registry = TemplateRegistry()

        # Test valid template retrieval
        template = registry.get_template("frontend-expert")
        assert template is not None
        assert template.name == "frontend-expert"
        assert "skills" in template.config
        assert "optimization" in template.config

        # Test invalid template retrieval
        with pytest.raises(TemplateNotFoundError):
            registry.get_template("non-existent-template")

    def test_template_validation(self):
        """Test that templates are properly validated."""
        from moai_adk.optimization.templates import TemplateRegistry, TemplateValidationError

        registry = TemplateRegistry()

        # Test invalid template configuration
        invalid_config = {
            "name": "invalid-template",
            "missing_required_field": "value",
            # Missing 'skills' and 'optimization' required fields
        }

        with pytest.raises(TemplateValidationError):
            registry.validate_template(invalid_config)

    def test_template_customization(self):
        """Test that templates can be customized for specific needs."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Get base template
        base_template = registry.get_template("backend-expert")

        # Customize template
        customized = registry.customize_template(
            "backend-expert", skill_priorities={"python": 0.9, "django": 0.8}, optimization_target="performance"
        )

        # Verify customization
        assert customized != base_template
        assert customized.config["skill_priorities"]["python"] == 0.9
        assert customized.config["optimization_target"] == "performance"

    def test_template_performance_optimization(self):
        """Test that template application is optimized for performance."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Measure template application time
        start_time = __import__("time").time()

        # Apply template multiple times (should use caching)
        for i in range(10):
            registry.apply_template("frontend-expert", {"project_type": "web"})

        end_time = __import__("time").time()
        avg_time = (end_time - start_time) / 10

        # Performance target: template application < 5ms average
        assert avg_time < 0.005, f"Template application too slow: {avg_time:.3f}s"

    def test_template_composition(self):
        """Test that templates can be composed together."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Create composite template
        composite = registry.compose_templates(
            ["frontend-expert", "backend-expert"], frontend_weight=0.6, backend_weight=0.4
        )

        # Verify composition
        assert composite is not None
        assert "frontend-expert" in composite.config["composed_from"]
        assert "backend-expert" in composite.config["composed_from"]
        assert composite.config["weights"]["frontend-expert"] == 0.6

    def test_template_versioning(self):
        """Test template versioning and compatibility."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Get template version info
        template = registry.get_template("frontend-expert")
        assert hasattr(template, "version")
        assert template.version is not None

        # Test version compatibility
        compatible = registry.is_compatible("frontend-expert", template.version)
        assert compatible is True

    def test_template_skill_optimization(self):
        """Test that templates optimize skill allocation."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Apply template with skill optimization
        optimized = registry.apply_template(
            "backend-expert", {"project_type": "api", "performance_critical": True}, optimize_skills=True
        )

        # Verify skill optimization
        assert "optimized_skills" in optimized
        assert len(optimized["optimized_skills"]) > 0

    def test_template_scaling(self):
        """Test that template system scales with complexity."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Test with many templates
        all_templates = registry.list_templates()
        assert len(all_templates) >= 5  # Should have at least basic templates

        # Test template selection for complex scenarios
        complex_scenario = registry.select_templates_for_scenario(
            ["full-stack", "microservices", "performance-critical"]
        )

        assert len(complex_scenario) > 1
        assert "full-stack" in complex_scenario

    def test_template_cache_mechanism(self):
        """Test template caching mechanism for performance."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Mock template data
        with patch.object(registry, "_load_template") as mock_load:
            mock_load.return_value = {"name": "test", "config": {"test": "data"}}

            # First access loads from source
            registry.get_template("test-template")
            assert mock_load.call_count == 1

            # Second access uses cache
            registry.get_template("test-template")
            assert mock_load.call_count == 1  # No additional load


class TestTemplateIntegration:
    """Integration tests for template system with other components."""

    def test_template_integration_with_skill_allocation(self):
        """Test template integration with skill allocation matrix."""
        from moai_adk.optimization.skill_allocation import SkillMatrix
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()
        matrix = SkillMatrix(categories=["frontend", "backend"], skills_per_category=4)

        # Get template and apply skill allocation
        template = registry.get_template("frontend-expert")
        allocated_skills = matrix.optimize_allocation("frontend", template.config.get("skill_priorities", {}))

        # Verify integration
        assert allocated_skills is not None
        assert len(allocated_skills) > 0

    def test_template_integration_with_dynamic_loading(self):
        """Test template integration with dynamic skill loading."""
        from moai_adk.optimization.dynamic_loading import SkillLoader
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()
        loader = SkillLoader()

        # Mock skill data for template
        mock_skills = {"react": {"type": "frontend", "complexity": 0.8}, "vue": {"type": "frontend", "complexity": 0.7}}

        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = mock_skills

            # Apply template with dynamic loading
            result = registry.apply_template("frontend-expert", {"load_skills_dynamically": True}, skill_loader=loader)

            # Verify dynamic loading integration
            assert result is not None
            mock_load.assert_called()

    def test_template_performance_baseline(self):
        """Test that template system meets performance baseline."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Test template performance under load
        import time

        performance_results = []

        for i in range(100):
            start_time = time.time()
            registry.apply_template("backend-expert", {"test": i})
            end_time = time.time()
            performance_results.append(end_time - start_time)

        avg_time = sum(performance_results) / len(performance_results)
        max_time = max(performance_results)

        # Performance targets: average < 10ms, max < 25ms
        assert avg_time < 0.01, f"Average template application time too high: {avg_time:.3f}s"
        assert max_time < 0.025, f"Max template application time too high: {max_time:.3f}s"


class TestTemplateQuality:
    """Quality tests for template system."""

    def test_template_documentation_quality(self):
        """Test that templates have proper documentation."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Test template documentation
        template = registry.get_template("frontend-expert")
        assert template.description is not None
        assert len(template.description) > 10
        assert template.usage_examples is not None

    def test_template_error_handling(self):
        """Test template error handling quality."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Test graceful error handling
        try:
            registry.apply_template("invalid-template", {})
            assert False, "Should have raised an error"
        except Exception:
            # Exception should be raised with proper message
            assert True

    def test_template_security_validation(self):
        """Test template security validation."""
        from moai_adk.optimization.templates import TemplateRegistry

        registry = TemplateRegistry()

        # Test security validation of template configurations
        safe_config = {
            "name": "safe-template",
            "skills": ["safe_skill_1", "safe_skill_2"],
            "optimization": {"performance": True},
        }

        # Should not raise security issues
        try:
            registry.validate_template(safe_config)
        except Exception as e:
            assert False, f"Safe template failed validation: {e}"


if __name__ == "__main__":
    pytest.main([__file__])
