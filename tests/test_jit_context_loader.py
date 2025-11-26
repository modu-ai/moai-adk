#!/usr/bin/env python3
"""
Comprehensive test suite for JIT Context Loading System

Tests the Phase 2 implementation covering:
- Phase detection accuracy
- Skill filtering efficiency
- Token budget management
- Context caching performance
- Overall system integration
"""

import os
import sys
import tempfile
import time
from pathlib import Path
from unittest.mock import patch

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

from moai_adk.core.jit_context_loader import (
    ContextCache,
    ContextMetrics,
    JITContextLoader,
    Phase,
    PhaseConfig,
    PhaseDetector,
    SkillFilterEngine,
    SkillInfo,
    TokenBudgetManager,
    clear_jit_cache,
    get_jit_stats,
    initialize_jit_system,
    load_optimized_context,
)


class TestPhaseDetector:
    """Test phase detection functionality"""

    def test_detect_spec_phase(self):
        """Test SPEC phase detection from user input"""
        detector = PhaseDetector()

        # Test SPEC phase patterns
        spec_inputs = [
            "/moai:1-plan create user authentication system",
            "Create SPEC-001 for login functionality",
            "Design system requirements for user authentication",
            "spec requirements for API design",
        ]

        for input_text in spec_inputs:
            phase = detector.detect_phase(input_text)
            assert phase == Phase.SPEC, f"Failed to detect SPEC phase from: {input_text}"

    def test_detect_red_phase(self):
        """Test RED phase detection"""
        detector = PhaseDetector()

        red_inputs = [
            "/moai:2-run SPEC-001 RED phase",
            "test failing for user authentication",
            "Red phase: write failing tests",
            "TDD red phase implementation",
        ]

        for input_text in red_inputs:
            phase = detector.detect_phase(input_text)
            assert phase == Phase.RED, f"Failed to detect RED phase from: {input_text}"

    def test_detect_green_phase(self):
        """Test GREEN phase detection"""
        detector = PhaseDetector()

        green_inputs = [
            "/moai:2-run SPEC-001 GREEN phase",
            "make tests pass with minimal implementation",
            "Green phase: minimal implementation",
            "TDD green phase coding",
        ]

        for input_text in green_inputs:
            phase = detector.detect_phase(input_text)
            assert phase == Phase.GREEN, f"Failed to detect GREEN phase from: {input_text}"

    def test_phase_history_tracking(self):
        """Test phase change history tracking"""
        detector = PhaseDetector()

        # Simulate phase progression with explicit phase indicators
        detector.detect_phase("/moai:1-plan user authentication system")  # SPEC
        detector.detect_phase("RED phase: write failing tests for authentication")  # RED
        detector.detect_phase("Green phase: minimal implementation to pass tests")  # GREEN

        # Check history - should have 2 transitions
        assert len(detector.phase_history) == 2
        assert detector.phase_history[0]["from"] == Phase.SPEC.value
        assert detector.phase_history[0]["to"] == Phase.RED.value
        assert detector.phase_history[1]["from"] == Phase.RED.value
        assert detector.phase_history[1]["to"] == Phase.GREEN.value

    def test_get_phase_config(self):
        """Test phase configuration retrieval"""
        detector = PhaseDetector()

        spec_config = detector.get_phase_config(Phase.SPEC)
        assert isinstance(spec_config, PhaseConfig)
        assert spec_config.max_tokens == 30000
        assert len(spec_config.essential_skills) > 0
        assert spec_config.cache_clear_on_phase_change

        green_config = detector.get_phase_config(Phase.GREEN)
        assert green_config.max_tokens == 25000
        assert not green_config.cache_clear_on_phase_change


class TestSkillFilterEngine:
    """Test skill filtering functionality"""

    @pytest.fixture
    def temp_skills_dir(self):
        """Create temporary skills directory for testing"""
        with tempfile.TemporaryDirectory() as temp_dir:
            skills_dir = Path(temp_dir)

            # Create mock skill directories
            for skill_name in [
                "moai-foundation-ears",
                "moai-domain-testing",
                "moai-lang-python",
                "moai-essentials-debug",
            ]:
                skill_dir = skills_dir / skill_name
                skill_dir.mkdir()

                # Create SKILL.md file
                skill_file = skill_dir / "SKILL.md"
                content = f"# {skill_name}\n\nThis is a test skill for {skill_name}."
                skill_file.write_text(content)

            yield skills_dir

    def test_build_skill_index(self, temp_skills_dir):
        """Test skill index building"""
        engine = SkillFilterEngine(str(temp_skills_dir))

        assert len(engine.skill_index) == 4
        assert "moai-foundation-ears" in engine.skill_index
        assert "moai-domain-testing" in engine.skill_index

        # Check skill metadata
        ears_skill = engine.skill_index["moai-foundation-ears"]
        assert isinstance(ears_skill, SkillInfo)
        assert ears_skill.name == "moai-foundation-ears"
        assert ears_skill.tokens > 0
        assert "core" in ears_skill.categories

    def test_filter_skills_by_phase(self, temp_skills_dir):
        """Test skill filtering by phase"""
        engine = SkillFilterEngine(str(temp_skills_dir))

        # Test SPEC phase filtering
        spec_skills = engine.filter_skills(Phase.SPEC, 20000)
        assert len(spec_skills) > 0

        # Should prioritize foundation skills
        skill_names = [s.name for s in spec_skills]
        foundation_skills = [s for s in skill_names if "foundation" in s]
        assert len(foundation_skills) > 0

        # Test budget constraint
        total_tokens = sum(s.tokens for s in spec_skills)
        assert total_tokens <= 20000

    def test_phase_preferences(self):
        """Test phase-based skill preferences"""
        engine = SkillFilterEngine()

        # Check preferences are loaded
        assert "spec" in engine.phase_preferences
        assert "red" in engine.phase_preferences
        assert "refactor" in engine.phase_preferences

        # Check priority assignments
        spec_prefs = engine.phase_preferences["spec"]
        assert spec_prefs["moai-foundation-ears"] == 1  # High priority
        assert spec_prefs["moai-lang-python"] == 3  # Low priority

    def test_get_skill_stats(self, temp_skills_dir):
        """Test skill statistics generation"""
        engine = SkillFilterEngine(str(temp_skills_dir))

        stats = engine.get_skill_stats()
        assert stats["total_skills"] == 4
        assert stats["total_tokens"] > 0
        assert "categories" in stats
        assert stats["average_tokens_per_skill"] > 0

        # Check category statistics
        assert "core" in stats["categories"]
        assert stats["categories"]["core"] > 0


class TestTokenBudgetManager:
    """Test token budget management"""

    def test_initialize_budgets(self):
        """Test budget initialization"""
        manager = TokenBudgetManager()

        # Check all phases have budgets
        expected_phases = ["spec", "red", "green", "refactor", "sync", "debug", "planning"]
        for phase in expected_phases:
            assert phase in manager.phase_budgets
            assert manager.phase_budgets[phase] > 0

        # Check reduced budgets
        assert manager.phase_budgets["spec"] == 30000  # Reduced from 50K
        assert manager.phase_budgets["refactor"] == 20000  # Reduced from 50K
        assert manager.phase_budgets["sync"] == 40000  # Reduced from 50K

    def test_check_budget(self):
        """Test budget checking functionality"""
        manager = TokenBudgetManager()

        # Test within budget
        within_budget, remaining = manager.check_budget(Phase.SPEC, 25000)
        assert within_budget
        assert remaining == 5000

        # Test over budget
        within_budget, remaining = manager.check_budget(Phase.SPEC, 35000)
        assert not within_budget
        assert remaining == 30000

    def test_record_usage(self):
        """Test usage recording"""
        manager = TokenBudgetManager()

        # Record normal usage
        manager.record_usage(Phase.SPEC, 20000, "test context")
        assert len(manager.usage_history) == 1
        assert manager.current_usage == 20000

        # Record over-budget usage
        manager.record_usage(Phase.SPEC, 40000, "over budget test")
        assert len(manager.budget_warnings) == 1
        assert "exceeded budget" in manager.budget_warnings[0]["warning"]

    def test_efficiency_metrics(self):
        """Test efficiency metrics calculation"""
        manager = TokenBudgetManager()

        # Record some usage
        manager.record_usage(Phase.SPEC, 25000, "efficient usage")
        manager.record_usage(Phase.RED, 20000, "normal usage")
        manager.record_usage(Phase.SPEC, 35000, "inefficient usage")  # Over budget

        metrics = manager.get_efficiency_metrics()
        assert "efficiency_score" in metrics
        assert "budget_compliance" in metrics
        assert "total_usage" in metrics
        assert metrics["total_usage"] == 80000

        # Budget compliance should be less than 100% due to over-budget usage
        assert metrics["budget_compliance"] < 100


class TestContextCache:
    """Test context caching functionality"""

    def test_basic_cache_operations(self):
        """Test basic cache put/get operations"""
        cache = ContextCache(max_size=3, max_memory_mb=1)

        # Put entries
        cache.put("key1", "value1", 1000)
        cache.put("key2", {"nested": "data"}, 2000)

        # Get entries
        entry1 = cache.get("key1")
        assert entry1 is not None
        assert entry1.content == "value1"
        assert entry1.access_count == 1

        entry2 = cache.get("key2")
        assert entry2 is not None
        assert entry2.content["nested"] == "data"

        # Test miss
        entry3 = cache.get("key3")
        assert entry3 is None

    def test_lru_eviction(self):
        """Test LRU eviction policy"""
        cache = ContextCache(max_size=2, max_memory_mb=1)

        # Fill cache to capacity
        cache.put("key1", "value1", 1000)
        cache.put("key2", "value2", 1000)

        # Add third entry (should evict first)
        cache.put("key3", "value3", 1000)

        # First entry should be evicted
        assert cache.get("key1") is None
        assert cache.get("key2") is not None
        assert cache.get("key3") is not None

        assert cache.evictions == 1

    def test_phase_based_clearing(self):
        """Test phase-based cache clearing"""
        cache = ContextCache()

        # Add entries for different phases
        cache.put("spec1", "spec content", 1000, "spec")
        cache.put("red1", "red content", 1000, "red")
        cache.put("spec2", "more spec content", 1000, "spec")

        assert len(cache.cache) == 3

        # Clear spec phase entries
        cache.clear_phase("spec")

        assert len(cache.cache) == 1
        assert cache.get("red1") is not None
        assert cache.get("spec1") is None
        assert cache.get("spec2") is None

    def test_cache_statistics(self):
        """Test cache statistics"""
        cache = ContextCache()

        # Perform some operations
        cache.put("key1", "value1", 1000)
        cache.get("key1")  # Hit
        cache.get("key2")  # Miss

        stats = cache.get_stats()
        assert stats["entries"] == 1
        assert stats["hits"] == 1
        assert stats["misses"] == 1
        assert stats["hit_rate"] == 50.0


class TestJITContextLoader:
    """Test main JIT Context Loader functionality"""

    @pytest.fixture
    def temp_project_dir(self):
        """Create temporary project structure"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_dir = Path(temp_dir)

            # Create basic project structure
            (project_dir / ".moai" / "specs" / "SPEC-001").mkdir(parents=True)
            (project_dir / "src" / "auth").mkdir(parents=True)
            (project_dir / ".claude" / "skills").mkdir(parents=True)

            # Create mock skill files
            skills_dir = project_dir / ".claude" / "skills"
            for skill_name in ["moai-foundation-ears", "moai-lang-python"]:
                skill_dir = skills_dir / skill_name
                skill_dir.mkdir()
                (skill_dir / "SKILL.md").write_text(f"# {skill_name}\n\nTest content.")

            # Create mock documents
            spec_file = project_dir / ".moai" / "specs" / "SPEC-001" / "spec.md"
            spec_file.write_text("# SPEC-001: User Authentication System\n\nRequirements...")

            python_file = project_dir / "src" / "auth" / "auth.py"
            python_file.write_text("def authenticate_user():\n    pass")

            yield project_dir

    @pytest.fixture
    def jit_loader(self, temp_project_dir):
        """Create JIT loader instance with temp directory"""
        with patch("moai_adk.core.jit_context_loader.SkillFilterEngine") as mock_engine:
            # Mock skill filter to use temp directory
            mock_engine.return_value.skills_dir = temp_project_dir / ".claude" / "skills"
            mock_engine.return_value.skill_index = {
                "moai-foundation-ears": SkillInfo(
                    name="moai-foundation-ears",
                    path=str(temp_project_dir / ".claude" / "skills" / "moai-foundation-ears" / "SKILL.md"),
                    size=1000,
                    tokens=250,
                    categories=["core"],
                    priority=1,
                ),
                "moai-lang-python": SkillInfo(
                    name="moai-lang-python",
                    path=str(temp_project_dir / ".claude" / "skills" / "moai-lang-python" / "SKILL.md"),
                    size=1000,
                    tokens=250,
                    categories=["language"],
                    priority=2,
                ),
            }
            mock_engine.return_value.filter_skills.return_value = [
                mock_engine.return_value.skill_index["moai-foundation-ears"]
            ]

            loader = JITContextLoader()
            loader.skill_filter = mock_engine.return_value
            yield loader

    @pytest.mark.asyncio
    async def test_load_context_spec_phase(self, jit_loader):
        """Test context loading for SPEC phase"""
        user_input = "/moai:1-plan user authentication system"
        context = {"spec_id": "SPEC-001", "language": "python"}

        context_data, metrics = await jit_loader.load_context(user_input, context=context)

        # Verify context structure
        assert context_data["phase"] == Phase.SPEC.value
        assert "skills" in context_data
        assert "documents" in context_data
        assert "metadata" in context_data

        # Verify metrics
        assert isinstance(metrics, ContextMetrics)
        assert metrics.phase == Phase.SPEC.value
        assert metrics.load_time >= 0
        assert metrics.token_count > 0

    @pytest.mark.asyncio
    async def test_cache_functionality(self, jit_loader):
        """Test context caching functionality"""
        user_input = "/moai:1-plan test feature"
        context = {"spec_id": "SPEC-001"}

        # First load (cache miss)
        context_data1, metrics1 = await jit_loader.load_context(user_input, context=context)
        assert not metrics1.cache_hit

        # Second load (cache hit)
        context_data2, metrics2 = await jit_loader.load_context(user_input, context=context)
        assert metrics2.cache_hit
        assert metrics2.load_time < metrics1.load_time  # Should be faster

        # Content should be identical
        assert context_data1 == context_data2

    @pytest.mark.asyncio
    async def test_phase_transition(self, jit_loader):
        """Test phase transition handling"""
        # Start with SPEC phase
        context_data1, _ = await jit_loader.load_context("/moai:1-plan user system")
        assert context_data1["phase"] == Phase.SPEC.value

        # Transition to RED phase
        context_data2, _ = await jit_loader.load_context("/moai:2-run SPEC-001 RED phase")
        assert context_data2["phase"] == Phase.RED.value

        # Check phase history
        assert len(jit_loader.phase_detector.phase_history) > 0
        last_transition = jit_loader.phase_detector.phase_history[-1]
        assert last_transition["from"] == Phase.SPEC.value
        assert last_transition["to"] == Phase.RED.value

    @pytest.mark.asyncio
    async def test_token_budget_enforcement(self, jit_loader):
        """Test token budget enforcement"""
        # Mock skill filter to return many skills (over budget)
        large_skills = []
        for i in range(20):
            skill = SkillInfo(
                name=f"skill-{i}", path=f"/mock/path/skill-{i}", size=2000, tokens=2000, categories=["test"], priority=3
            )
            large_skills.append(skill)

        jit_loader.skill_filter.filter_skills.return_value = large_skills

        user_input = "/moai:2-run SPEC-001 GREEN phase"
        context = {"spec_id": "SPEC-001"}

        context_data, metrics = await jit_loader.load_context(user_input, context=context)

        # Should apply aggressive optimization
        assert metrics.token_count <= 25000  # GREEN phase budget
        assert len(context_data["skills"]) < len(large_skills)  # Skills filtered

    def test_get_comprehensive_stats(self, jit_loader):
        """Test comprehensive statistics gathering"""
        stats = jit_loader.get_comprehensive_stats()

        # Check all required sections
        assert "performance" in stats
        assert "cache" in stats
        assert "token_efficiency" in stats
        assert "skill_filter" in stats
        assert "current_phase" in stats

        # Check performance stats
        perf = stats["performance"]
        assert "total_loads" in perf
        assert "average_load_time" in perf
        assert "cache_hit_rate" in perf
        assert "efficiency_score" in perf


class TestIntegration:
    """Integration tests for the complete JIT system"""

    @pytest.mark.asyncio
    async def test_full_workflow_simulation(self):
        """Test complete development workflow simulation"""
        # Simulate a typical development session
        workflow_steps = [
            ("/moai:1-plan user authentication system", {"spec_id": "SPEC-001"}),
            ("/moai:2-run SPEC-001 RED phase", {"spec_id": "SPEC-001"}),
            ("/moai:2-run SPEC-001 GREEN phase", {"spec_id": "SPEC-001"}),
            ("/moai:2-run SPEC-001 REFACTOR phase", {"spec_id": "SPEC-001"}),
            ("/moai:3-sync auto SPEC-001", {"spec_id": "SPEC-001"}),
        ]

        for user_input, context in workflow_steps:
            # Use the global convenience function
            context_data, metrics = await load_optimized_context(user_input, context=context)

            # Verify basic structure
            assert "phase" in context_data
            assert "skills" in context_data
            assert "documents" in context_data

            # Verify metrics
            assert isinstance(metrics, ContextMetrics)
            assert metrics.phase in [p.value for p in Phase]

    def test_global_functions(self):
        """Test global convenience functions"""
        # Test stats function
        stats = get_jit_stats()
        assert isinstance(stats, dict)
        assert "performance" in stats

        # Test cache clearing
        clear_jit_cache()
        # Should not raise any errors

        # Test system initialization
        result = initialize_jit_system()
        assert result

    @pytest.mark.asyncio
    async def test_performance_benchmarks(self):
        """Test performance benchmarks"""
        loader = JITContextLoader()

        # Benchmark cache performance
        user_input = "/moai:1-plan performance test"
        context = {"spec_id": "PERF-TEST"}

        # Cold load
        start_time = time.time()
        await loader.load_context(user_input, context=context)
        cold_load_time = time.time() - start_time

        # Warm load (cache hit)
        start_time = time.time()
        await loader.load_context(user_input, context=context)
        warm_load_time = time.time() - start_time

        # Cache should be significantly faster
        assert warm_load_time < cold_load_time
        assert warm_load_time < 0.01  # Should be under 10ms

        # Check final stats
        stats = loader.get_comprehensive_stats()
        assert stats["cache"]["hit_rate"] > 0


class TestErrorHandling:
    """Test error handling and edge cases"""

    @pytest.mark.asyncio
    async def test_empty_user_input(self):
        """Test handling of empty user input"""
        loader = JITContextLoader()

        context_data, metrics = await loader.load_context("", [])
        assert context_data["phase"] == Phase.SPEC.value  # Should default to SPEC
        assert isinstance(metrics, ContextMetrics)

    @pytest.mark.asyncio
    async def test_invalid_context(self):
        """Test handling of invalid context data"""
        loader = JITContextLoader()

        # Test with malformed context
        context_data, metrics = await loader.load_context("/moai:1-plan test", context={"invalid_key": "value"})
        assert context_data is not None
        assert metrics is not None

    def test_missing_skills_directory(self):
        """Test behavior when skills directory doesn't exist"""
        engine = SkillFilterEngine("/nonexistent/directory")

        # Should not crash, just return empty results
        stats = engine.get_skill_stats()
        assert stats["total_skills"] == 0

    def test_token_budget_extreme_cases(self):
        """Test token budget manager with extreme values"""
        manager = TokenBudgetManager()

        # Test zero budget
        within_budget, remaining = manager.check_budget(Phase.SPEC, 0)
        assert within_budget
        assert remaining == 30000

        # Test extremely large budget
        within_budget, remaining = manager.check_budget(Phase.SPEC, 1000000)
        assert not within_budget

    def test_cache_memory_limits(self):
        """Test cache memory limit enforcement"""
        cache = ContextCache(max_size=10, max_memory_mb=0.001)  # Very small limit

        # Add entries until memory limit is reached
        large_content = "x" * 10000  # 10KB content
        for i in range(20):
            cache.put(f"key{i}", large_content, 2500)

        # Should have evicted many entries due to memory limit
        assert len(cache.cache) < 20
        assert cache.evictions > 0


if __name__ == "__main__":
    # Run tests
    pytest.main([__file__, "-v", "--tb=short"])
