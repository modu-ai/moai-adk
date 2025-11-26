"""
Unit tests for skill allocation matrix system.

Tests the TAG-002 component: Skill Allocation Matrix System
"""

import time

import pytest


class TestSkillAllocationMatrix:
    """Test cases for skill allocation matrix system."""

    def test_skill_matrix_initialization(self):
        """Test that skill matrix can be initialized with proper dimensions."""
        # This test should fail initially because the SkillMatrix class doesn't exist
        from moai_adk.optimization.skill_allocation import SkillMatrix

        # Create a skill matrix with 3 categories and 5 skills per category
        matrix = SkillMatrix(categories=["frontend", "backend", "devops"], skills_per_category=5)

        # Verify matrix dimensions
        assert matrix.categories == ["frontend", "backend", "devops"]
        assert matrix.skills_per_category == 5
        assert matrix.matrix.shape == (3, 5)

    def test_skill_allocation_optimization(self):
        """Test that skill allocation optimizes agent configuration."""
        from moai_adk.optimization.skill_allocation import SkillMatrix

        matrix = SkillMatrix(categories=["frontend", "backend"], skills_per_category=3)

        # Test optimal allocation for frontend agent
        frontend_skills = matrix.optimize_allocation("frontend", {"ui_design": 0.9, "react": 0.8})

        # Verify allocation returns optimized skill set
        assert len(frontend_skills) > 0
        assert "ui_design" in frontend_skills  # Should be included due to high priority
        assert 0.0 <= frontend_skills["ui_design"] <= 1.0  # Skills should be weighted

    def test_skill_allocation_performance(self):
        """Test that skill allocation meets performance targets."""
        from moai_adk.optimization.skill_allocation import SkillMatrix

        matrix = SkillMatrix(categories=["frontend", "backend", "devops"], skills_per_category=10)

        import time

        start_time = time.time()

        # Perform allocation optimization
        result = matrix.optimize_allocation("backend", {"python": 0.9, "django": 0.8, "sql": 0.7})

        end_time = time.time()
        allocation_time = end_time - start_time

        # Performance target: allocation should complete in under 10ms
        assert allocation_time < 0.01, f"Allocation too slow: {allocation_time:.3f}s"
        assert isinstance(result, dict)

    def test_skill_matrix_validation(self):
        """Test that skill matrix validates input properly."""
        from moai_adk.optimization.skill_allocation import SkillAllocationError, SkillMatrix

        # Test invalid category
        with pytest.raises(SkillAllocationError):
            matrix = SkillMatrix(categories=["frontend"], skills_per_category=3)
            matrix.optimize_allocation("invalid_category", {})

        # Test invalid skill weights
        with pytest.raises(SkillAllocationError):
            matrix = SkillMatrix(categories=["frontend"], skills_per_category=3)
            matrix.optimize_allocation("frontend", {"invalid_skill": 1.5})  # Weight > 1.0

    def test_skill_categorization(self):
        """Test skill categorization system."""
        from moai_adk.optimization.skill_allocation import SkillMatrix

        matrix = SkillMatrix(categories=["frontend", "backend"], skills_per_category=4)

        # Test skill categorization
        frontend_skills = matrix.get_category_skills("frontend")
        backend_skills = matrix.get_category_skills("backend")

        assert len(frontend_skills) == 4
        assert len(backend_skills) == 4
        assert set(frontend_skills.keys()).isdisjoint(set(backend_skills.keys()))

    def test_matrix_scaling(self):
        """Test that skill matrix scales with increasing complexity."""
        from moai_adk.optimization.skill_allocation import SkillMatrix

        # Test with small matrix
        small_matrix = SkillMatrix(categories=["frontend"], skills_per_category=3)
        small_result = small_matrix.optimize_allocation("frontend", {"a": 0.8, "b": 0.7})

        # Test with large matrix
        large_matrix = SkillMatrix(categories=["frontend", "backend", "devops", "data"], skills_per_category=10)
        large_result = large_matrix.optimize_allocation("frontend", {"a": 0.8, "b": 0.7})

        # Both should work without errors
        assert isinstance(small_result, dict)
        assert isinstance(large_result, dict)
        assert len(large_result) >= len(small_result)


class TestSkillAllocationIntegration:
    """Integration tests for skill allocation with other components."""

    def test_skill_allocation_with_agent_factory(self):
        """Test integration with agent factory for skill optimization."""
        from moai_adk.optimization.skill_allocation import SkillMatrix

        matrix = SkillMatrix(categories=["frontend", "backend"], skills_per_category=5)

        # Mock agent factory integration
        mock_agent_config = {
            "name": "test-agent",
            "domain": "frontend",
            "required_skills": ["react", "css", "javascript"],
            "priority_weights": {"react": 0.9, "css": 0.8, "javascript": 0.7},
        }

        optimized_skills = matrix.optimize_allocation_for_agent(mock_agent_config)

        # Verify integration works
        assert isinstance(optimized_skills, dict)
        assert len(optimized_skills) > 0

        # Should include high-priority skills
        assert "react" in optimized_skills

    def test_skill_allocation_performance_baseline(self):
        """Test that skill allocation meets performance baseline requirements."""
        from moai_adk.optimization.skill_allocation import SkillMatrix

        matrix = SkillMatrix(categories=["frontend", "backend", "devops"], skills_per_category=8)

        # Test performance baseline
        performance_results = []
        for _ in range(100):  # Test 100 allocations
            start_time = time.time()
            matrix.optimize_allocation("backend", {"python": 0.9, "django": 0.8})
            end_time = time.time()
            performance_results.append(end_time - start_time)

        avg_time = sum(performance_results) / len(performance_results)
        max_time = max(performance_results)

        # Performance targets: average < 5ms, max < 15ms
        assert avg_time < 0.005, f"Average time too high: {avg_time:.3f}s"
        assert max_time < 0.015, f"Max time too high: {max_time:.3f}s"


if __name__ == "__main__":
    pytest.main([__file__])
