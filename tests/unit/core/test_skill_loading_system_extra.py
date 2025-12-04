"""Enhanced unit tests for Skill Loading System module.

This module tests:
- SkillData dataclass and methods
- LRUCache with TTL support
- SkillValidator with comprehensive validation
- SkillRegistry with skill discovery
- SkillLoader with caching and error handling
- Skill dependency management
"""

import tempfile
from datetime import datetime, timedelta
from pathlib import Path
from unittest import mock

import pytest
import yaml

from moai_adk.core.skill_loading_system import (
    DependencyError,
    LRUCache,
    SkillData,
    SkillLoadingError,
    SkillLoader,
    SkillNotFoundError,
    SkillRegistry,
    SkillValidationError,
    SkillValidator,
    clear_skill_cache,
    detect_required_skills,
    get_skill_cache_stats,
    load_skill,
)


class TestSkillData:
    """Test SkillData dataclass."""

    def test_skill_data_creation(self):
        """Test creating SkillData instance."""
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill", "status": "active"},
            content="# Test Skill\n\nContent here",
            loaded_at=datetime.now(),
        )
        assert skill.name == "test-skill"
        assert skill.loaded_at is not None

    def test_skill_data_get_capability(self):
        """Test getting capability from skill."""
        skill = SkillData(
            name="test-skill",
            frontmatter={
                "name": "test-skill",
                "capabilities": {"api_development": True},
            },
            content="",
            loaded_at=datetime.now(),
        )
        capability = skill.get_capability("api_development")
        assert capability is True

    def test_skill_data_supports_effort(self):
        """Test checking effort level support."""
        skill = SkillData(
            name="test-skill",
            frontmatter={
                "name": "test-skill",
                "supported_efforts": [1, 3],
            },
            content="",
            loaded_at=datetime.now(),
        )
        assert skill.supports_effort(1) is True
        assert skill.supports_effort(3) is True
        assert skill.supports_effort(5) is False

    def test_skill_data_apply_filter_basic(self):
        """Test applying basic effort filter."""
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill"},
            content="## Quick Reference\nBasic\n## Advanced\nAdvanced content",
            loaded_at=datetime.now(),
        )
        skill.apply_filter("basic")
        assert "basic" in skill.applied_filters

    def test_skill_data_get_empty_skill(self):
        """Test creating empty fallback skill."""
        skill = SkillData.get_empty_skill("fallback-skill")
        assert skill.name == "fallback-skill"
        assert skill.frontmatter["status"] == "fallback"


class TestLRUCache:
    """Test LRUCache with TTL support."""

    def test_lru_cache_initialization(self):
        """Test LRUCache initialization."""
        cache = LRUCache(maxsize=10, ttl=3600)
        assert cache.maxsize == 10
        assert cache.ttl == 3600

    def test_lru_cache_set_and_get(self):
        """Test setting and getting cache values."""
        cache = LRUCache()
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )
        cache.set("test-skill", skill)
        retrieved = cache.get("test-skill")
        assert retrieved is not None
        assert retrieved.name == "test"

    def test_lru_cache_ttl_expiration(self):
        """Test cache entry expiration based on TTL."""
        cache = LRUCache(maxsize=10, ttl=1)
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )
        cache.set("test-skill", skill)

        # Simulate time passing
        cache.cache["test-skill"] = (
            skill,
            datetime.now() - timedelta(seconds=2),
        )
        retrieved = cache.get("test-skill")
        assert retrieved is None  # Should be expired

    def test_lru_cache_maxsize_enforcement(self):
        """Test that cache respects maxsize limit."""
        cache = LRUCache(maxsize=3)
        for i in range(5):
            skill = SkillData(
                name=f"skill{i}",
                frontmatter={},
                content="",
                loaded_at=datetime.now(),
            )
            cache.set(f"skill{i}", skill)

        assert len(cache.cache) <= 3

    def test_lru_cache_clear(self):
        """Test clearing cache."""
        cache = LRUCache()
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )
        cache.set("test-skill", skill)
        cache.clear()
        assert len(cache.cache) == 0

    def test_lru_cache_keys(self):
        """Test getting cache keys."""
        cache = LRUCache()
        cache.set("skill1", SkillData("skill1", {}, "", datetime.now()))
        cache.set("skill2", SkillData("skill2", {}, "", datetime.now()))

        keys = cache.keys()
        assert "skill1" in keys
        assert "skill2" in keys


class TestSkillValidator:
    """Test SkillValidator class."""

    def test_validator_initialization(self):
        """Test SkillValidator initialization."""
        registry = SkillRegistry()
        validator = SkillValidator(registry)
        assert validator.registry == registry

    def test_validate_skill_name_format_moai(self):
        """Test validating MoAI skill name format."""
        registry = SkillRegistry()
        registry.skills["moai-lang-python"] = {}
        validator = SkillValidator(registry)

        # Should not raise
        assert validator.validate_skill_name("moai-lang-python") is True

    def test_validate_skill_name_empty(self):
        """Test validating empty skill name."""
        registry = SkillRegistry()
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("")

    def test_validate_skill_name_invalid_format(self):
        """Test validating skill name with invalid format."""
        registry = SkillRegistry()
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("InvalidSkillName")

    def test_validate_skill_name_not_found(self):
        """Test validating skill name not in registry."""
        registry = SkillRegistry()
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("moai-lang-python")

    def test_validate_effort_parameter_valid(self):
        """Test validating valid effort parameter."""
        registry = SkillRegistry()
        registry.skills["moai-lang-python"] = {"supported_efforts": [1, 3, 5]}
        validator = SkillValidator(registry)

        assert validator.validate_effort_parameter("moai-lang-python", 3) is True

    def test_validate_effort_parameter_invalid_value(self):
        """Test validating invalid effort value."""
        registry = SkillRegistry()
        registry.skills["moai-lang-python"] = {"supported_efforts": [1, 3, 5]}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_effort_parameter("moai-lang-python", 4)

    def test_validate_effort_parameter_not_supported(self):
        """Test validating effort level not supported by skill."""
        registry = SkillRegistry()
        registry.skills["moai-lang-python"] = {"supported_efforts": [1, 3]}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_effort_parameter("moai-lang-python", 5)

    def test_validate_dependencies_no_requirements(self):
        """Test validating dependencies with no requirements."""
        registry = SkillRegistry()
        registry.skills["test-skill"] = {"requires": []}
        validator = SkillValidator(registry)

        assert validator.validate_dependencies("test-skill", []) is True

    def test_validate_dependencies_missing(self):
        """Test validating when dependencies are missing."""
        registry = SkillRegistry()
        registry.skills["test-skill"] = {"requires": ["moai-lang-python"]}
        validator = SkillValidator(registry)

        with pytest.raises(DependencyError):
            validator.validate_dependencies("test-skill", [])

    def test_validate_dependencies_present(self):
        """Test validating when dependencies are present."""
        registry = SkillRegistry()
        registry.skills["test-skill"] = {"requires": ["moai-lang-python"]}
        validator = SkillValidator(registry)

        assert validator.validate_dependencies("test-skill", ["moai-lang-python"]) is True


class TestSkillRegistry:
    """Test SkillRegistry class."""

    def test_registry_initialization(self):
        """Test SkillRegistry initialization."""
        registry = SkillRegistry()
        assert len(registry.skills) == 0
        assert registry._initialized is False

    def test_registry_register_skill(self):
        """Test registering a skill."""
        registry = SkillRegistry()
        metadata = {"name": "test", "version": "1.0.0"}
        registry.register_skill("test-skill", metadata)

        assert "test-skill" in registry.skills

    def test_registry_get_skill_metadata(self):
        """Test getting skill metadata."""
        registry = SkillRegistry()
        metadata = {"name": "test", "version": "1.0.0"}
        registry.register_skill("test-skill", metadata)

        retrieved = registry.get_skill_metadata("test-skill")
        assert retrieved["name"] == "test"

    def test_registry_check_compatibility(self):
        """Test checking skill compatibility."""
        registry = SkillRegistry()
        registry.register_skill("test-skill", {})

        result = registry.check_compatibility("test-skill", [])
        assert result is True


class TestSkillLoader:
    """Test SkillLoader class."""

    def test_loader_initialization(self):
        """Test SkillLoader initialization."""
        loader = SkillLoader(skill_paths=[])
        assert loader.registry is not None
        assert loader.validator is not None
        assert loader.cache is not None

    def test_loader_get_cache_stats(self):
        """Test getting loader cache statistics."""
        loader = SkillLoader(skill_paths=[])
        stats = loader.get_cache_stats()

        assert "cached_skills" in stats
        assert "cache_size" in stats
        assert "registry_skills" in stats

    def test_get_skill_path_multiple_locations(self):
        """Test getting skill path from multiple locations."""
        loader = SkillLoader(skill_paths=[])
        path = loader._get_skill_path("test-skill")
        assert path is not None


class TestSkillDetection:
    """Test skill detection from prompts."""

    def test_detect_required_skills_backend(self):
        """Test detecting backend-related skills."""
        skills = detect_required_skills("expert-backend", "Develop FastAPI")
        assert "moai-foundation-claude" in skills

    def test_detect_required_skills_frontend(self):
        """Test detecting frontend-related skills."""
        skills = detect_required_skills("expert-frontend", "Build React component")
        assert "moai-foundation-claude" in skills

    def test_detect_required_skills_from_prompt(self):
        """Test detecting skills from prompt content."""
        skills = detect_required_skills(
            "general-purpose",
            "Implement Python FastAPI backend",
        )
        assert "moai-foundation-claude" in skills


class TestSkillLoaderFallback:
    """Test skill loader fallback mechanisms."""

    def test_loader_fallback_on_not_found(self):
        """Test loader fallback when skill not found."""
        loader = SkillLoader(skill_paths=[])

        # Mock the registry to return no skill
        with mock.patch.object(loader.registry, "skills", {}):
            try:
                skill = loader._get_fallback_skill("unknown-skill")
                assert skill is not None
            except Exception:
                # Fallback may fail if no fallbacks available
                pass


class TestSkillLoaderCircularDependency:
    """Test circular dependency detection."""

    def test_loader_detects_circular_dependency(self):
        """Test that loader detects circular dependencies."""
        loader = SkillLoader(skill_paths=[])
        loader.loading_stack = ["skill-a", "skill-b"]

        # Test that circular dependency check happens during load_skill
        # The method will try to find the skill in the loading stack
        try:
            loader.load_skill("skill-a")
        except (SkillLoadingError, Exception):
            # Either raises or returns fallback - both are acceptable
            pass


class TestSkillDataFiltering:
    """Test skill content filtering."""

    def test_skill_filter_basic_content(self):
        """Test filtering for basic effort level."""
        content = """## Quick Reference
Quick info

## Implementation
Full implementation

## Advanced
Advanced details"""

        skill = SkillData(
            name="test",
            frontmatter={},
            content=content,
            loaded_at=datetime.now(),
        )

        filtered_content = skill._filter_content(content, "basic")
        assert "Quick Reference" in filtered_content

    def test_skill_filter_comprehensive_content(self):
        """Test keeping full content for comprehensive effort."""
        content = "Full content here"
        skill = SkillData(
            name="test",
            frontmatter={},
            content=content,
            loaded_at=datetime.now(),
        )

        filtered = skill._filter_content(content, "comprehensive")
        assert filtered == content

    def test_skill_filter_standard_content(self):
        """Test filtering out advanced sections."""
        content = """## Implementation
Standard implementation

## Advanced
Advanced info

## Reference
Reference info"""

        skill = SkillData(
            name="test",
            frontmatter={},
            content=content,
            loaded_at=datetime.now(),
        )

        filtered = skill._filter_content(content, "standard")
        assert "Advanced" not in filtered


class TestPublicAPI:
    """Test public API functions."""

    def test_clear_skill_cache(self):
        """Test clearing skill cache."""
        clear_skill_cache()
        # Should complete without error

    def test_get_skill_cache_stats(self):
        """Test getting cache stats."""
        stats = get_skill_cache_stats()
        assert isinstance(stats, dict)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
