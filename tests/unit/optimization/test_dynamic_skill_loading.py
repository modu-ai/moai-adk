"""
Unit tests for dynamic skill loading framework.

Tests the TAG-003 component: Dynamic Skill Loading Framework
"""

import time
from unittest.mock import patch

import pytest


class TestDynamicSkillLoading:
    """Test cases for dynamic skill loading framework."""

    def test_skill_loader_initialization(self):
        """Test that skill loader can be initialized properly."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        # Test initialization with default settings
        loader = SkillLoader()
        assert loader.cache_size == 100  # Default cache size
        assert loader.max_load_time == 0.1  # Default max load time (100ms)

        # Test initialization with custom settings
        loader = SkillLoader(cache_size=50, max_load_time=0.05)
        assert loader.cache_size == 50
        assert loader.max_load_time == 0.05

    def test_skill_lazy_loading(self):
        """Test that skills are loaded lazily on demand."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        loader = SkillLoader()

        # Mock skill loading to simulate lazy behavior
        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = {"mock_skill": "loaded_skill_data"}

            # First load - should call the actual loader
            skill1 = loader.get_skill("frontend", "react")
            mock_load.assert_called_once_with("frontend", "react")

            # Second load - should use cache (no additional call)
            skill2 = loader.get_skill("frontend", "react")
            assert skill1 == skill2
            assert mock_load.call_count == 1  # Only called once due to caching

    def test_skill_performance_benefits(self):
        """Test that dynamic loading provides performance benefits."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        loader = SkillLoader(cache_size=100)

        # Mock skill data
        skill_data = {"frontend": ["react", "vue", "angular"], "backend": ["django", "flask"]}

        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = skill_data

            # Measure initial load time
            start_time = time.time()
            loader.get_skill("frontend", "react")
            initial_load_time = time.time() - start_time

            # Measure cached load time
            start_time = time.time()
            loader.get_skill("frontend", "react")
            cached_load_time = time.time() - start_time

            # Performance target: cached load should be at least 80% faster
            performance_improvement = (initial_load_time - cached_load_time) / initial_load_time
            assert performance_improvement > 0.8, f"Performance improvement too low: {performance_improvement:.2f}"

    def test_skill_cache_eviction(self):
        """Test that cache evicts old entries when size limit is reached."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        # Create loader with small cache
        loader = SkillLoader(cache_size=3)

        # Fill cache beyond limit
        skills_data = {}
        for i in range(5):
            loader.get_skill(f"category_{i}", f"skill_{i}")
            skills_data[f"skill_{i}"] = f"data_{i}"

        # Cache should have evicted oldest entries
        assert len(loader._cache) <= 3

    def test_skill_load_timeout(self):
        """Test that skill loading times out properly."""
        from moai_adk.optimization.dynamic_loading import SkillLoader, SkillTimeoutError

        loader = SkillLoader(max_load_time=0.01)  # 10ms timeout

        # Mock a slow skill loading
        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.side_effect = lambda category, skill: time.sleep(0.02)  # 20ms delay

            # Should raise timeout error
            with pytest.raises(SkillTimeoutError):
                loader.get_skill("slow_category", "slow_skill")

    def test_skill_validation_on_load(self):
        """Test that skills are validated during loading."""
        from moai_adk.optimization.dynamic_loading import SkillLoader, SkillValidationError

        loader = SkillLoader()

        # Mock invalid skill data
        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = {"invalid": "data", "metadata": None}

            # Should raise validation error for invalid skill
            with pytest.raises(SkillValidationError):
                loader.get_skill("invalid_category", "invalid_skill")

    def test_skill_batch_loading(self):
        """Test that skills can be loaded in batches for performance."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        loader = SkillLoader()

        # Batch load multiple skills
        skills_to_load = [("frontend", "react"), ("frontend", "vue"), ("backend", "django"), ("backend", "flask")]

        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = {"skill": "data"}

            # Use batch loading
            results = loader.batch_load_skills(skills_to_load)

            # Verify all skills were loaded
            assert len(results) == 4
            for skill_id in skills_to_load:
                assert skill_id in results

            # Verify cache was populated
            assert len(loader._cache) == 4

    def test_skill_preloading_optimization(self):
        """Test that skill preloading optimizes common access patterns."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        loader = SkillLoader()

        # Mock skill data
        common_skills = {"frontend": ["react", "vue", "javascript"], "backend": ["python", "django", "sql"]}

        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = {"skill": "data"}

            # Preload common skills
            loader.preload_common_skills(common_skills)

            # Verify common skills are cached
            for category, skills in common_skills.items():
                for skill in skills:
                    assert (category, skill) in loader._cache


class TestSkillLoadingIntegration:
    """Integration tests for skill loading with other components."""

    def test_skill_loading_with_allocation_matrix(self):
        """Test integration with skill allocation matrix."""
        from moai_adk.optimization.dynamic_loading import SkillLoader
        from moai_adk.optimization.skill_allocation import SkillMatrix

        loader = SkillLoader()
        matrix = SkillMatrix(categories=["frontend", "backend"], skills_per_category=3)

        # Mock skill data
        mock_skills = {
            "react": {"type": "frontend", "complexity": 0.8, "popularity": 0.9},
            "django": {"type": "backend", "complexity": 0.7, "popularity": 0.8},
        }

        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = mock_skills

            # Test integration
            optimized_allocation = matrix.optimize_allocation("frontend", {"react": 0.9})
            loaded_skills = loader.batch_load_skills([("frontend", "react")])

            assert optimized_allocation["react"] > 0.8
            assert "react" in loaded_skills

    def test_skill_loading_performance_targets(self):
        """Test that skill loading meets performance targets."""
        from moai_adk.optimization.dynamic_loading import SkillLoader

        loader = SkillLoader(cache_size=100)

        # Mock realistic skill loading scenarios
        skill_scenarios = [
            (10, 50),  # 10 skills, 50 accesses each
            (20, 30),  # 20 skills, 30 accesses each
            (50, 10),  # 50 skills, 10 accesses each
        ]

        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = {"skill": "data"}

            total_time = 0
            total_loads = 0

            for num_skills, num_accesses in skill_scenarios:
                scenario_time = 0
                # First access triggers loading
                for _ in range(num_skills):
                    start = time.time()
                    loader.get_skill("category", f"skill_{total_loads}")
                    scenario_time += time.time() - start
                    total_loads += 1

                # Subsequent accesses should be fast (cached)
                for _ in range(num_accesses):
                    start = time.time()
                    loader.get_skill("category", f"skill_{total_loads - num_skills}")
                    scenario_time += time.time() - start
                    total_loads += 1

                total_time += scenario_time

            # Performance target: average load time < 1ms for cached access
            avg_cached_time = total_time / (sum(num_accesses for _, num_accesses in skill_scenarios))
            assert avg_cached_time < 0.001, f"Average cached load time too high: {avg_cached_time:.3f}s"


if __name__ == "__main__":
    pytest.main([__file__])
