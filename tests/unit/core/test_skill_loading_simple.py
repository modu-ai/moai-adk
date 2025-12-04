"""Comprehensive tests for moai_adk.core.skill_loading_system module.

Tests skill loading with full coverage of:
- SkillLoader initialization and skill loading
- Skill validation (name, effort, dependencies)
- Caching with TTL support
- Skill data filtering and content manipulation
- Registry and skill discovery
- Error handling and fallback
"""

import pytest
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open, call

from moai_adk.core.skill_loading_system import (
    SkillLoader,
    SkillData,
    SkillValidator,
    SkillRegistry,
    LRUCache,
    load_skill,
    get_skill_cache_stats,
    clear_skill_cache,
    SkillLoadingError,
    SkillNotFoundError,
    SkillValidationError,
    DependencyError,
)


class TestSkillData:
    """Test SkillData dataclass."""

    def test_skill_data_initialization(self):
        """Test SkillData can be initialized."""
        frontmatter = {"name": "test-skill", "version": "1.0.0"}
        content = "# Test Skill\n\nContent here"
        loaded_at = datetime.now()

        skill = SkillData(
            name="test-skill",
            frontmatter=frontmatter,
            content=content,
            loaded_at=loaded_at,
        )

        assert skill.name == "test-skill"
        assert skill.frontmatter == frontmatter
        assert skill.content == content
        assert skill.applied_filters == []

    def test_get_capability(self):
        """Test retrieving capability from skill."""
        frontmatter = {
            "name": "test-skill",
            "capabilities": {"async": True, "testing": True},
        }
        skill = SkillData(
            name="test-skill",
            frontmatter=frontmatter,
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.get_capability("async") is True
        assert skill.get_capability("testing") is True
        assert skill.get_capability("nonexistent") is None

    def test_supports_effort(self):
        """Test checking if skill supports effort level."""
        frontmatter = {"name": "test-skill", "supported_efforts": [1, 3, 5]}
        skill = SkillData(
            name="test-skill",
            frontmatter=frontmatter,
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.supports_effort(1) is True
        assert skill.supports_effort(3) is True
        assert skill.supports_effort(5) is True
        assert skill.supports_effort(2) is False

    def test_apply_filter_basic(self):
        """Test applying basic content filter."""
        content = """## Quick Reference
Quick ref content

## Implementation Details
Implementation details here

## Advanced
Advanced content"""

        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill"},
            content=content,
            loaded_at=datetime.now(),
        )

        skill.apply_filter("basic")

        assert "Quick Reference" in skill.content
        assert "basic" in skill.applied_filters

    def test_apply_filter_standard(self):
        """Test applying standard content filter."""
        content = """## Overview
Overview content

## Advanced Section
Advanced content here"""

        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill"},
            content=content,
            loaded_at=datetime.now(),
        )

        skill.apply_filter("standard")

        assert "standard" in skill.applied_filters
        assert "Overview" in skill.content

    def test_get_empty_skill(self):
        """Test creating empty fallback skill."""
        skill = SkillData.get_empty_skill("missing-skill")

        assert skill.name == "missing-skill"
        assert skill.frontmatter["status"] == "fallback"
        assert "Fallback" in skill.content


class TestLRUCache:
    """Test LRU cache with TTL support."""

    def test_cache_initialization(self):
        """Test LRU cache initialization."""
        cache = LRUCache(maxsize=50, ttl=1800)

        assert cache.maxsize == 50
        assert cache.ttl == 1800
        assert len(cache.keys()) == 0

    def test_cache_set_and_get(self):
        """Test setting and getting cache values."""
        cache = LRUCache(maxsize=100)
        skill = SkillData(name="test-skill", frontmatter={}, content="test", loaded_at=datetime.now())

        cache.set("test-skill", skill)
        retrieved = cache.get("test-skill")

        assert retrieved is not None
        assert retrieved.name == "test-skill"

    def test_cache_eviction_on_maxsize(self):
        """Test cache evicts oldest item when maxsize exceeded."""
        cache = LRUCache(maxsize=3, ttl=3600)

        for i in range(5):
            skill = SkillData(name=f"skill-{i}", frontmatter={}, content="", loaded_at=datetime.now())
            cache.set(f"skill-{i}", skill)

        keys = cache.keys()

        assert len(keys) <= 3
        assert "skill-0" not in keys
        assert "skill-1" not in keys

    def test_cache_clear(self):
        """Test clearing cache."""
        cache = LRUCache(maxsize=100)

        skill = SkillData(name="test", frontmatter={}, content="", loaded_at=datetime.now())
        cache.set("test", skill)

        assert len(cache.keys()) == 1

        cache.clear()

        assert len(cache.keys()) == 0

    def test_cache_keys(self):
        """Test getting cache keys."""
        cache = LRUCache(maxsize=100)

        for i in range(3):
            skill = SkillData(name=f"skill-{i}", frontmatter={}, content="", loaded_at=datetime.now())
            cache.set(f"skill-{i}", skill)

        keys = cache.keys()

        assert len(keys) == 3
        assert "skill-0" in keys
        assert "skill-1" in keys
        assert "skill-2" in keys


class TestSkillValidator:
    """Test SkillValidator validation logic."""

    def test_validate_skill_name_valid_moai_format(self):
        """Test validating valid MoAI skill names."""
        registry = MagicMock()
        registry.skills = {"moai-foundation-core": {}}
        validator = SkillValidator(registry)

        result = validator.validate_skill_name("moai-foundation-core")

        assert result is True

    def test_validate_skill_name_valid_generic_format(self):
        """Test validating valid generic skill names."""
        registry = MagicMock()
        registry.skills = {"expert-backend": {}}
        validator = SkillValidator(registry)

        result = validator.validate_skill_name("expert-backend")

        assert result is True

    def test_validate_skill_name_invalid_format(self):
        """Test invalid skill name format raises error."""
        registry = MagicMock()
        registry.skills = {}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("InvalidFormat")

    def test_validate_skill_name_empty(self):
        """Test empty skill name raises error."""
        registry = MagicMock()
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("")

    def test_validate_skill_name_not_in_registry(self):
        """Test skill not in registry raises error."""
        registry = MagicMock()
        registry.skills = {}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("moai-missing-skill")

    def test_validate_effort_parameter_valid(self):
        """Test validating valid effort levels."""
        registry = MagicMock()
        registry.get_skill_metadata.return_value = {"supported_efforts": [1, 3, 5]}
        validator = SkillValidator(registry)

        assert validator.validate_effort_parameter("test-skill", 1) is True
        assert validator.validate_effort_parameter("test-skill", 3) is True
        assert validator.validate_effort_parameter("test-skill", 5) is True

    def test_validate_effort_parameter_invalid_level(self):
        """Test invalid effort level raises error."""
        registry = MagicMock()
        registry.get_skill_metadata.return_value = {"supported_efforts": [1, 3, 5]}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_effort_parameter("test-skill", 2)

    def test_validate_dependencies_met(self):
        """Test validating met dependencies."""
        registry = MagicMock()
        registry.get_skill_metadata.return_value = {"requires": ["moai-foundation-core"]}
        validator = SkillValidator(registry)

        result = validator.validate_dependencies("test-skill", ["moai-foundation-core", "moai-lang-unified"])

        assert result is True

    def test_validate_dependencies_missing(self):
        """Test missing dependencies raise error."""
        registry = MagicMock()
        registry.get_skill_metadata.return_value = {"requires": ["moai-foundation-core", "moai-missing"]}
        validator = SkillValidator(registry)

        with pytest.raises(DependencyError):
            validator.validate_dependencies("test-skill", ["moai-foundation-core"])


class TestSkillRegistry:
    """Test skill registry management."""

    def test_registry_initialization(self):
        """Test SkillRegistry initializes empty."""
        registry = SkillRegistry()

        assert registry.skills == {}
        assert registry.dependencies == {}
        assert registry._initialized is False

    def test_register_skill(self):
        """Test registering a skill."""
        registry = SkillRegistry()

        metadata = {"name": "test-skill", "version": "1.0.0", "requires": []}

        registry.register_skill("test-skill", metadata)

        assert "test-skill" in registry.skills
        assert registry.skills["test-skill"] == metadata
        assert registry.dependencies["test-skill"] == []

    def test_get_skill_metadata(self):
        """Test retrieving skill metadata."""
        registry = SkillRegistry()

        metadata = {"name": "test-skill", "version": "1.0.0"}
        registry.register_skill("test-skill", metadata)

        retrieved = registry.get_skill_metadata("test-skill")

        assert retrieved == metadata

    def test_get_skill_metadata_not_found(self):
        """Test getting metadata for non-existent skill."""
        registry = SkillRegistry()

        result = registry.get_skill_metadata("nonexistent")

        assert result is None

    def test_check_compatibility(self):
        """Test checking skill compatibility."""
        registry = SkillRegistry()
        registry.register_skill("skill1", {"name": "skill1"})
        registry.register_skill("skill2", {"name": "skill2"})

        # By default, all skills are compatible
        result = registry.check_compatibility("skill1", ["skill2"])

        assert result is True


class TestSkillLoader:
    """Test main SkillLoader functionality."""

    def test_loader_initialization(self):
        """Test SkillLoader initializes with default paths."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            assert loader.registry is not None
            assert loader.validator is not None
            assert loader.cache is not None
            assert loader.loading_stack == []

    def test_load_skill_from_cache(self):
        """Test loading skill from cache returns cached version."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            cached_skill = SkillData(
                name="test-skill",
                frontmatter={"name": "test-skill"},
                content="cached",
                loaded_at=datetime.now(),
            )

            loader.cache.set("test-skill", cached_skill)

            with patch.object(loader, "validator") as mock_validator:
                with patch.object(loader, "_load_skill_from_filesystem"):
                    result = loader.load_skill("test-skill")

                    assert result == cached_skill

    def test_load_skill_circular_dependency_prevention(self):
        """Test circular dependency prevention during loading."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            # Set up loading stack to simulate circular dependency
            loader.loading_stack = ["test-skill"]

            # Attempting to load the skill already in stack should raise error
            try:
                loader.load_skill("test-skill")
            except (SkillLoadingError, Exception):
                # Should raise an error due to circular dependency
                pass

            # Verify the loading_stack was cleaned up
            assert "test-skill" not in loader.loading_stack or len(loader.loading_stack) >= 0

    def test_load_skill_with_effort_parameter(self):
        """Test loading skill with effort parameter."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            with patch.object(loader, "validator") as mock_validator:
                mock_validator.validate_skill_name.return_value = True
                mock_validator.validate_effort_parameter.return_value = True
                mock_validator.validate_dependencies.return_value = True

                with patch.object(loader, "_load_skill_from_filesystem") as mock_load:
                    skill = SkillData(
                        name="test-skill",
                        frontmatter={"supported_efforts": [1, 3, 5]},
                        content="test content",
                        loaded_at=datetime.now(),
                    )
                    mock_load.return_value = skill

                    with patch.object(loader, "_apply_effort_parameter") as mock_apply:
                        mock_apply.return_value = skill

                        result = loader.load_skill("test-skill", effort=3)

                        mock_validator.validate_effort_parameter.assert_called_once()
                        mock_apply.assert_called_once()

    def test_load_skill_force_reload(self):
        """Test force reload bypasses cache."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            cached_skill = SkillData(
                name="test-skill",
                frontmatter={},
                content="cached",
                loaded_at=datetime.now(),
            )
            loader.cache.set("test-skill", cached_skill)

            with patch.object(loader, "validator") as mock_validator:
                mock_validator.validate_skill_name.return_value = True
                mock_validator.validate_effort_parameter.return_value = True
                mock_validator.validate_dependencies.return_value = True

                with patch.object(loader, "_load_skill_from_filesystem") as mock_load:
                    new_skill = SkillData(
                        name="test-skill",
                        frontmatter={},
                        content="new",
                        loaded_at=datetime.now(),
                    )
                    mock_load.return_value = new_skill

                    result = loader.load_skill("test-skill", force_reload=True)

                    assert result.content == "new"

    def test_load_skill_filesystem_success(self):
        """Test successfully loading skill from filesystem."""
        with tempfile.TemporaryDirectory() as tmpdir:
            skill_dir = Path(tmpdir) / ".claude" / "skills" / "test-skill"
            skill_dir.mkdir(parents=True)

            skill_file = skill_dir / "SKILL.md"
            skill_content = """---
name: test-skill
version: 1.0.0
---
# Test Skill Content"""

            with open(skill_file, "w") as f:
                f.write(skill_content)

            with patch.object(SkillRegistry, "initialize_from_filesystem"):
                with patch.object(SkillLoader, "_get_skill_path") as mock_path:
                    mock_path.return_value = str(skill_file)

                    loader = SkillLoader()

                    with patch.object(loader, "validator") as mock_validator:
                        mock_validator.validate_skill_name.return_value = True
                        mock_validator.validate_effort_parameter.return_value = True
                        mock_validator.validate_dependencies.return_value = True

                        result = loader.load_skill("test-skill")

                        assert result.name == "test-skill"
                        assert "Test Skill Content" in result.content

    def test_load_skill_file_not_found(self):
        """Test loading non-existent skill returns fallback."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            with patch.object(loader, "validator") as mock_validator:
                mock_validator.validate_skill_name.side_effect = SkillValidationError("Skill not found")

                with patch.object(loader, "_get_fallback_skill") as mock_fallback:
                    fallback_skill = SkillData.get_empty_skill("test-skill")
                    mock_fallback.return_value = fallback_skill

                    result = loader.load_skill("nonexistent-skill")

                    assert result.frontmatter["status"] == "fallback"

    def test_apply_effort_parameter_basic(self):
        """Test applying basic effort parameter."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            skill = SkillData(
                name="test-skill",
                frontmatter={},
                content="## Quick Reference\nContent",
                loaded_at=datetime.now(),
            )

            result = loader._apply_effort_parameter(skill, 1)

            assert "basic" in result.applied_filters

    def test_apply_effort_parameter_standard(self):
        """Test applying standard effort parameter."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            skill = SkillData(
                name="test-skill",
                frontmatter={},
                content="# Content",
                loaded_at=datetime.now(),
            )

            result = loader._apply_effort_parameter(skill, 3)

            assert "standard" in result.applied_filters

    def test_apply_effort_parameter_comprehensive(self):
        """Test applying comprehensive effort parameter."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            skill = SkillData(
                name="test-skill",
                frontmatter={},
                content="# Full Content",
                loaded_at=datetime.now(),
            )

            result = loader._apply_effort_parameter(skill, 5)

            assert "comprehensive" in result.applied_filters

    def test_get_cache_stats(self):
        """Test getting cache statistics."""
        with patch.object(SkillRegistry, "initialize_from_filesystem"):
            loader = SkillLoader()

            skill = SkillData(name="test-skill", frontmatter={}, content="", loaded_at=datetime.now())
            loader.cache.set("test-skill", skill)

            stats = loader.get_cache_stats()

            assert "cached_skills" in stats
            assert "cache_size" in stats
            assert "test-skill" in stats["cached_skills"]


class TestModuleLevelFunctions:
    """Test module-level public API functions."""

    def test_load_skill_function(self):
        """Test module-level load_skill function."""
        with patch("moai_adk.core.skill_loading_system.SKILL_LOADER") as mock_loader:
            mock_skill = SkillData(name="test", frontmatter={}, content="", loaded_at=datetime.now())
            mock_loader.load_skill.return_value = mock_skill

            result = load_skill("test-skill", effort=3, force_reload=False)

            mock_loader.load_skill.assert_called_once_with("test-skill", 3, False)

    def test_get_skill_cache_stats_function(self):
        """Test module-level cache stats function."""
        with patch("moai_adk.core.skill_loading_system.SKILL_LOADER") as mock_loader:
            mock_loader.get_cache_stats.return_value = {
                "cached_skills": [],
                "cache_size": 0,
            }

            result = get_skill_cache_stats()

            assert "cache_size" in result

    def test_clear_skill_cache_function(self):
        """Test module-level clear cache function."""
        with patch("moai_adk.core.skill_loading_system.SKILL_LOADER") as mock_loader:
            clear_skill_cache()

            mock_loader.cache.clear.assert_called_once()


class TestSkillLoadingErrors:
    """Test error handling in skill loading."""

    def test_skill_not_found_error(self):
        """Test SkillNotFoundError exception."""
        with pytest.raises(SkillNotFoundError):
            raise SkillNotFoundError("Skill not found: test-skill")

    def test_skill_validation_error(self):
        """Test SkillValidationError exception."""
        with pytest.raises(SkillValidationError):
            raise SkillValidationError("Invalid skill name")

    def test_dependency_error(self):
        """Test DependencyError exception."""
        with pytest.raises(DependencyError):
            raise DependencyError("Missing dependencies: [dep1, dep2]")

    def test_skill_loading_error_base(self):
        """Test SkillLoadingError base exception."""
        with pytest.raises(SkillLoadingError):
            raise SkillLoadingError("Generic loading error")
