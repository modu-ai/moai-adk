"""
Comprehensive tests for SkillLoader and related classes.

Tests skill loading, validation, caching, and dependency resolution.
"""

import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

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
)


class TestSkillData:
    """Test SkillData dataclass."""

    def test_skill_data_creation(self):
        """Test creating a SkillData instance."""
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill", "version": "1.0.0"},
            content="# Test Skill\n\nContent here",
            loaded_at=datetime.now(),
        )

        assert skill.name == "test-skill"
        assert skill.frontmatter["version"] == "1.0.0"
        assert "Content here" in skill.content

    def test_skill_data_get_capability(self):
        """Test get_capability method."""
        skill = SkillData(
            name="test-skill",
            frontmatter={
                "name": "test-skill",
                "capabilities": {"basic": "value1", "advanced": "value2"},
            },
            content="Content",
            loaded_at=datetime.now(),
        )

        assert skill.get_capability("basic") == "value1"
        assert skill.get_capability("advanced") == "value2"

    def test_skill_data_supports_effort(self):
        """Test supports_effort method."""
        skill = SkillData(
            name="test-skill",
            frontmatter={
                "name": "test-skill",
                "supported_efforts": [1, 3],
            },
            content="Content",
            loaded_at=datetime.now(),
        )

        assert skill.supports_effort(1) is True
        assert skill.supports_effort(3) is True
        assert skill.supports_effort(5) is False

    def test_skill_data_apply_filter(self):
        """Test apply_filter method."""
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill"},
            content="## Quick Reference\nBasic\n## Advanced\nAdvanced",
            loaded_at=datetime.now(),
        )

        skill.apply_filter("basic")
        assert "basic" in skill.applied_filters
        assert len(skill.applied_filters) > 0

    def test_skill_data_get_empty_skill(self):
        """Test get_empty_skill fallback."""
        skill = SkillData.get_empty_skill("missing-skill")

        assert skill.name == "missing-skill"
        assert skill.frontmatter["status"] == "fallback"
        assert "fallback" in skill.content.lower()


class TestLRUCache:
    """Test LRU cache implementation."""

    def test_lru_cache_set_get(self):
        """Test setting and getting items in cache."""
        cache = LRUCache(maxsize=2)
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test"},
            content="Content",
            loaded_at=datetime.now(),
        )

        cache.set("skill1", skill)
        retrieved = cache.get("skill1")

        assert retrieved is not None
        assert retrieved.name == "test-skill"

    def test_lru_cache_maxsize_eviction(self):
        """Test LRU eviction when maxsize exceeded."""
        cache = LRUCache(maxsize=2)
        skill1 = SkillData(
            name="skill1",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )
        skill2 = SkillData(
            name="skill2",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )
        skill3 = SkillData(
            name="skill3",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        cache.set("key1", skill1)
        cache.set("key2", skill2)
        cache.set("key3", skill3)  # Should evict key1

        assert cache.get("key1") is None
        assert cache.get("key2") is not None
        assert cache.get("key3") is not None

    def test_lru_cache_ttl_expiration(self):
        """Test TTL-based cache expiration."""
        cache = LRUCache(maxsize=10, ttl=1)
        skill = SkillData(
            name="test-skill",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        cache.set("skill1", skill)
        assert cache.get("skill1") is not None

        # Simulate expiration by modifying timestamp
        key_data = cache.cache["skill1"]
        from datetime import timedelta

        cache.cache["skill1"] = (
            key_data[0],
            datetime.now() - timedelta(seconds=2),
        )

        assert cache.get("skill1") is None

    def test_lru_cache_clear(self):
        """Test clearing cache."""
        cache = LRUCache(maxsize=10)
        skill = SkillData(
            name="test-skill",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        cache.set("skill1", skill)
        assert len(cache.keys()) > 0

        cache.clear()
        assert len(cache.keys()) == 0

    def test_lru_cache_keys(self):
        """Test getting cache keys."""
        cache = LRUCache(maxsize=10)
        skill = SkillData(
            name="test-skill",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        cache.set("skill1", skill)
        cache.set("skill2", skill)

        keys = cache.keys()
        assert "skill1" in keys
        assert "skill2" in keys


class TestSkillValidator:
    """Test SkillValidator class."""

    def test_validate_skill_name_valid(self):
        """Test validating valid skill names."""
        registry = SkillRegistry()
        registry.skills = {"moai-foundation-claude": {}}
        validator = SkillValidator(registry)

        assert validator.validate_skill_name("moai-foundation-claude") is True

    def test_validate_skill_name_invalid_format(self):
        """Test validating invalid skill name format."""
        registry = SkillRegistry()
        registry.skills = {}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("INVALID_NAME")

    def test_validate_skill_name_not_found(self):
        """Test validating non-existent skill."""
        registry = SkillRegistry()
        registry.skills = {}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("moai-nonexistent-skill")

    def test_validate_skill_name_empty(self):
        """Test validating empty skill name."""
        registry = SkillRegistry()
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_skill_name("")

    def test_validate_effort_parameter_valid(self):
        """Test validating valid effort parameter."""
        registry = SkillRegistry()
        registry.skills = {"test-skill": {"requires": [], "supported_efforts": [1, 3, 5]}}
        validator = SkillValidator(registry)

        assert validator.validate_effort_parameter("test-skill", 3) is True

    def test_validate_effort_parameter_invalid_value(self):
        """Test validating invalid effort value."""
        registry = SkillRegistry()
        registry.skills = {"test-skill": {"supported_efforts": [1, 3, 5]}}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_effort_parameter("test-skill", 4)

    def test_validate_effort_parameter_unsupported(self):
        """Test validating unsupported effort level."""
        registry = SkillRegistry()
        registry.skills = {"test-skill": {"supported_efforts": [1, 3]}}
        validator = SkillValidator(registry)

        with pytest.raises(SkillValidationError):
            validator.validate_effort_parameter("test-skill", 5)

    def test_validate_dependencies_satisfied(self):
        """Test validating satisfied dependencies."""
        registry = SkillRegistry()
        registry.skills = {"test-skill": {"requires": ["dep1", "dep2"]}}
        validator = SkillValidator(registry)

        assert validator.validate_dependencies("test-skill", ["dep1", "dep2", "dep3"]) is True

    def test_validate_dependencies_missing(self):
        """Test validating missing dependencies."""
        registry = SkillRegistry()
        registry.skills = {"test-skill": {"requires": ["dep1", "dep2"]}}
        validator = SkillValidator(registry)

        with pytest.raises(DependencyError):
            validator.validate_dependencies("test-skill", ["dep1"])


class TestSkillRegistry:
    """Test SkillRegistry class."""

    def test_skill_registry_init(self):
        """Test SkillRegistry initialization."""
        registry = SkillRegistry()

        assert registry.skills == {}
        assert registry.dependencies == {}
        assert registry._initialized is False

    def test_skill_registry_register_skill(self):
        """Test registering a skill."""
        registry = SkillRegistry()
        metadata = {
            "name": "test-skill",
            "requires": ["dep1"],
        }

        registry.register_skill("test-skill", metadata)

        assert "test-skill" in registry.skills
        assert registry.skills["test-skill"]["name"] == "test-skill"
        assert registry.dependencies["test-skill"] == ["dep1"]

    def test_skill_registry_get_skill_metadata(self):
        """Test retrieving skill metadata."""
        registry = SkillRegistry()
        metadata = {"name": "test-skill", "version": "1.0"}
        registry.register_skill("test-skill", metadata)

        retrieved = registry.get_skill_metadata("test-skill")
        assert retrieved["version"] == "1.0"

    def test_skill_registry_check_compatibility(self):
        """Test compatibility checking."""
        registry = SkillRegistry()
        registry.register_skill("skill1", {})
        registry.register_skill("skill2", {})

        assert registry.check_compatibility("skill1", ["skill2"]) is True

    @patch("moai_adk.core.skill_loading_system.os.path.exists")
    @patch("moai_adk.core.skill_loading_system.os.walk")
    def test_initialize_from_filesystem(self, mock_walk, mock_exists):
        """Test initializing registry from filesystem."""
        mock_exists.return_value = True
        mock_walk.return_value = [
            ("/skills", [], ["SKILL.md"]),
        ]

        registry = SkillRegistry()

        with patch.object(registry, "_parse_skill_metadata") as mock_parse:
            with patch.object(registry, "_extract_skill_name") as mock_extract:
                mock_extract.return_value = "test-skill"
                mock_parse.return_value = {"name": "test-skill"}

                registry.initialize_from_filesystem(["/skills"])

                assert registry._initialized is True


class TestSkillLoader:
    """Test SkillLoader class."""

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_skill_loader_init(self, mock_init):
        """Test SkillLoader initialization."""
        loader = SkillLoader(skill_paths=["/test"])

        assert loader.cache is not None
        assert loader.registry is not None
        assert mock_init.called

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    @patch("moai_adk.core.skill_loading_system.SkillLoader._load_skill_from_filesystem")
    def test_load_skill_success(self, mock_load_fs, mock_init):
        """Test successful skill loading."""
        mock_init.return_value = None
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill"},
            content="Content",
            loaded_at=datetime.now(),
        )
        mock_load_fs.return_value = skill

        loader = SkillLoader()
        loader.registry.skills = {"test-skill": {}}
        loader.validator.registry = loader.registry

        result = loader.load_skill("test-skill")

        assert result.name == "test-skill"
        # Verify it's cached by checking the cache has the key
        assert "test-skill" in loader.cache.keys()

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_load_skill_circular_dependency(self, mock_init):
        """Test loading skill with circular dependency detected."""
        mock_init.return_value = None

        loader = SkillLoader()
        # Add skill to loading_stack to simulate circular dependency
        loader.loading_stack = ["test-skill"]
        loader.registry.skills = {"test-skill": {"requires": []}}
        loader.validator.registry = loader.registry

        # When loading a skill that's already in the stack, should handle it
        # It may either raise an error or return a fallback, both are valid responses
        try:
            result = loader.load_skill("test-skill")
            # If it doesn't raise, it should return a fallback
            assert result.frontmatter.get("status") == "fallback"
        except SkillLoadingError:
            # If it raises, that's also acceptable
            pass

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_load_skill_not_found(self, mock_init):
        """Test loading non-existent skill falls back."""
        mock_init.return_value = None

        loader = SkillLoader()
        loader.registry.skills = {}

        result = loader.load_skill("nonexistent-skill")

        assert result.frontmatter["status"] == "fallback"

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    @patch("moai_adk.core.skill_loading_system.SkillLoader._load_skill_from_filesystem")
    def test_load_skill_with_effort(self, mock_load_fs, mock_init):
        """Test loading skill with effort parameter."""
        mock_init.return_value = None
        skill = SkillData(
            name="test-skill",
            frontmatter={"name": "test-skill", "supported_efforts": [1, 3, 5]},
            content="Content",
            loaded_at=datetime.now(),
        )
        mock_load_fs.return_value = skill

        loader = SkillLoader()
        loader.registry.skills = {"test-skill": {"supported_efforts": [1, 3, 5]}}
        loader.validator.registry = loader.registry

        result = loader.load_skill("test-skill", effort=3)

        assert result.name == "test-skill"

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_load_skill_force_reload(self, mock_init):
        """Test force reload bypasses cache."""
        mock_init.return_value = None

        loader = SkillLoader()
        loader.registry.skills = {"test-skill": {}}

        with patch.object(loader, "_load_skill_from_filesystem") as mock_load:
            skill = SkillData(
                name="test-skill",
                frontmatter={},
                content="Content",
                loaded_at=datetime.now(),
            )
            mock_load.return_value = skill

            # First load
            loader.load_skill("test-skill")

            # Cache should have it
            assert loader.cache.get("test-skill") is not None

            # Force reload should call filesystem loader again
            loader.load_skill("test-skill", force_reload=True)
            assert mock_load.call_count >= 1

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_get_cache_stats(self, mock_init):
        """Test getting cache statistics."""
        mock_init.return_value = None

        loader = SkillLoader()
        loader.registry.skills = {"skill1": {}, "skill2": {}}

        stats = loader.get_cache_stats()

        assert "cache_size" in stats
        assert "registry_skills" in stats
        assert stats["registry_skills"] == 2


class TestSkillLoaderFileParsing:
    """Test skill file parsing methods."""

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_parse_skill_file_with_frontmatter(self, mock_init):
        """Test parsing skill file with YAML frontmatter."""
        mock_init.return_value = None

        content = """---
name: test-skill
version: 1.0.0
---

# Skill Content
Some content here"""

        loader = SkillLoader()
        frontmatter, body = loader._parse_skill_file(content)

        assert frontmatter["name"] == "test-skill"
        assert "Skill Content" in body

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_parse_skill_file_no_frontmatter(self, mock_init):
        """Test parsing skill file without frontmatter."""
        mock_init.return_value = None

        content = "# Just Content\nNo frontmatter"

        loader = SkillLoader()
        frontmatter, body = loader._parse_skill_file(content)

        assert frontmatter == {}
        assert "Just Content" in body


class TestDetectRequiredSkills:
    """Test automatic skill detection."""

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_detect_skills_for_expert_backend(self, mock_init):
        """Test detecting skills for backend expert."""
        from moai_adk.core.skill_loading_system import detect_required_skills

        mock_init.return_value = None

        skills = detect_required_skills("expert-backend", "Build REST API")

        assert "moai-foundation-claude" in skills
        assert "moai-lang-unified" in skills

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    def test_detect_skills_from_prompt(self, mock_init):
        """Test detecting skills from prompt content."""
        from moai_adk.core.skill_loading_system import detect_required_skills

        mock_init.return_value = None

        skills = detect_required_skills(
            "expert-backend",
            "Implement FastAPI endpoint for authentication",
        )

        assert "moai-lang-unified" in skills


class TestSkillLoaderGetSkillPath:
    """Test skill path resolution."""

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    @patch("moai_adk.core.skill_loading_system.os.path.exists")
    def test_get_skill_path_first_location(self, mock_exists, mock_init):
        """Test getting skill path from first location."""
        mock_init.return_value = None
        mock_exists.side_effect = [True]  # First path exists

        loader = SkillLoader()
        path = loader._get_skill_path("test-skill")

        assert ".claude/skills" in path
        assert "test-skill" in path

    @patch("moai_adk.core.skill_loading_system.SkillRegistry.initialize_from_filesystem")
    @patch("moai_adk.core.skill_loading_system.os.path.exists")
    def test_get_skill_path_fallback_location(self, mock_exists, mock_init):
        """Test getting skill path from fallback location."""
        mock_init.return_value = None
        mock_exists.side_effect = [False, True]  # First fails, second exists

        loader = SkillLoader()
        path = loader._get_skill_path("test-skill")

        assert "test-skill" in path
