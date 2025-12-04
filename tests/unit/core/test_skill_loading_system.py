"""Tests for moai_adk.core.skill_loading_system module."""

import pytest
from datetime import datetime
from moai_adk.core.skill_loading_system import (
    SkillLoadingError,
    SkillNotFoundError,
    SkillValidationError,
    DependencyError,
    SkillData,
)


class TestSkillLoadingExceptions:
    """Test custom exception classes."""

    def test_skill_loading_error(self):
        """Test SkillLoadingError exception."""
        with pytest.raises(SkillLoadingError):
            raise SkillLoadingError("Test error")

    def test_skill_not_found_error(self):
        """Test SkillNotFoundError exception."""
        with pytest.raises(SkillNotFoundError):
            raise SkillNotFoundError("Skill not found")

    def test_skill_validation_error(self):
        """Test SkillValidationError exception."""
        with pytest.raises(SkillValidationError):
            raise SkillValidationError("Validation failed")

    def test_dependency_error(self):
        """Test DependencyError exception."""
        with pytest.raises(DependencyError):
            raise DependencyError("Dependency not resolved")

    def test_exceptions_are_subclasses(self):
        """Test that custom exceptions inherit from SkillLoadingError."""
        assert issubclass(SkillNotFoundError, SkillLoadingError)
        assert issubclass(SkillValidationError, SkillLoadingError)
        assert issubclass(DependencyError, SkillLoadingError)


class TestSkillDataInit:
    """Test SkillData initialization."""

    def test_skill_data_creation(self):
        """Test creating a SkillData instance."""
        frontmatter = {"name": "test-skill", "version": "1.0.0"}
        content = "Skill content"
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
        assert skill.loaded_at == loaded_at

    def test_skill_data_default_filters(self):
        """Test that applied_filters defaults to empty list."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="content",
            loaded_at=datetime.now(),
        )

        assert skill.applied_filters == []

    def test_skill_data_with_filters(self):
        """Test SkillData with applied filters."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="content",
            loaded_at=datetime.now(),
            applied_filters=["basic", "security"],
        )

        assert skill.applied_filters == ["basic", "security"]


class TestSkillDataGetCapability:
    """Test SkillData.get_capability method."""

    def test_get_capability_exists(self):
        """Test getting existing capability."""
        frontmatter = {
            "capabilities": {"auth": True, "caching": False}
        }
        skill = SkillData(
            name="test",
            frontmatter=frontmatter,
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.get_capability("auth") is True
        assert skill.get_capability("caching") is False

    def test_get_capability_missing(self):
        """Test getting missing capability."""
        skill = SkillData(
            name="test",
            frontmatter={"capabilities": {}},
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.get_capability("nonexistent") is None

    def test_get_capability_no_capabilities_dict(self):
        """Test getting capability when no capabilities dict."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.get_capability("auth") is None


class TestSkillDataSupportsEffort:
    """Test SkillData.supports_effort method."""

    def test_supports_effort_default(self):
        """Test supports_effort with default supported efforts."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        # Default is [1, 3, 5]
        assert skill.supports_effort(1) is True
        assert skill.supports_effort(3) is True
        assert skill.supports_effort(5) is True
        assert skill.supports_effort(2) is False

    def test_supports_effort_custom(self):
        """Test supports_effort with custom supported efforts."""
        frontmatter = {"supported_efforts": [2, 4, 6]}
        skill = SkillData(
            name="test",
            frontmatter=frontmatter,
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.supports_effort(2) is True
        assert skill.supports_effort(4) is True
        assert skill.supports_effort(1) is False

    def test_supports_effort_empty_list(self):
        """Test supports_effort with empty efforts list."""
        frontmatter = {"supported_efforts": []}
        skill = SkillData(
            name="test",
            frontmatter=frontmatter,
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.supports_effort(1) is False


class TestSkillDataApplyFilter:
    """Test SkillData.apply_filter method."""

    def test_apply_filter_basic(self):
        """Test applying basic filter."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="# Header\nContent\n## Advanced\nAdvanced content",
            loaded_at=datetime.now(),
        )

        skill.apply_filter("basic")

        assert "basic" in skill.applied_filters
        assert skill.content != ""

    def test_apply_filter_comprehensive(self):
        """Test applying comprehensive filter."""
        original_content = "Test content"
        skill = SkillData(
            name="test",
            frontmatter={},
            content=original_content,
            loaded_at=datetime.now(),
        )

        skill.apply_filter("comprehensive")

        assert "comprehensive" in skill.applied_filters
        assert skill.content == original_content

    def test_apply_filter_standard(self):
        """Test applying standard filter."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="Standard content\n## Advanced\nAdvanced only",
            loaded_at=datetime.now(),
        )

        skill.apply_filter("standard")

        assert "standard" in skill.applied_filters

    def test_apply_filter_not_duplicate(self):
        """Test that filter is not applied twice."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="Content",
            loaded_at=datetime.now(),
        )

        skill.apply_filter("basic")
        skill.apply_filter("basic")

        # Should only appear once in applied_filters
        assert skill.applied_filters.count("basic") == 1

    def test_apply_filter_multiple(self):
        """Test applying multiple filters."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="Content",
            loaded_at=datetime.now(),
        )

        skill.apply_filter("basic")
        skill.apply_filter("standard")

        assert len(skill.applied_filters) == 2


class TestSkillDataFilterContent:
    """Test SkillData._filter_content method."""

    def test_filter_content_basic(self):
        """Test basic content filtering."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="# Test\n## Quick Reference\nQuick\n## Advanced\nAdvanced",
            loaded_at=datetime.now(),
        )

        filtered = skill._filter_content(skill.content, "basic")

        assert isinstance(filtered, str)

    def test_filter_content_comprehensive(self):
        """Test comprehensive content filtering."""
        original = "# Test\n## Section\nContent"
        skill = SkillData(
            name="test",
            frontmatter={},
            content=original,
            loaded_at=datetime.now(),
        )

        filtered = skill._filter_content(original, "comprehensive")

        assert filtered == original

    def test_filter_content_standard(self):
        """Test standard content filtering."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="# Header\nContent",
            loaded_at=datetime.now(),
        )

        filtered = skill._filter_content(skill.content, "standard")

        assert isinstance(filtered, str)

    def test_filter_content_preserves_headers(self):
        """Test that filtering preserves headers."""
        content = "# Main Header\nContent"
        skill = SkillData(
            name="test",
            frontmatter={},
            content=content,
            loaded_at=datetime.now(),
        )

        filtered = skill._filter_content(content, "basic")

        assert "#" in filtered


class TestSkillDataAttributes:
    """Test SkillData attributes."""

    def test_skill_name_attribute(self):
        """Test skill name attribute."""
        skill = SkillData(
            name="my-skill",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.name == "my-skill"

    def test_skill_frontmatter_attribute(self):
        """Test skill frontmatter attribute."""
        fm = {"version": "1.0.0"}
        skill = SkillData(
            name="test",
            frontmatter=fm,
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.frontmatter == fm

    def test_skill_content_attribute(self):
        """Test skill content attribute."""
        content = "Skill documentation"
        skill = SkillData(
            name="test",
            frontmatter={},
            content=content,
            loaded_at=datetime.now(),
        )

        assert skill.content == content

    def test_skill_loaded_at_attribute(self):
        """Test skill loaded_at attribute."""
        now = datetime.now()
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=now,
        )

        assert skill.loaded_at == now


class TestSkillDataEdgeCases:
    """Test edge cases for SkillData."""

    def test_skill_data_empty_content(self):
        """Test SkillData with empty content."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.content == ""

    def test_skill_data_empty_frontmatter(self):
        """Test SkillData with empty frontmatter."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="content",
            loaded_at=datetime.now(),
        )

        assert skill.frontmatter == {}

    def test_skill_data_unicode_name(self):
        """Test SkillData with unicode name."""
        skill = SkillData(
            name="test-한글",
            frontmatter={},
            content="",
            loaded_at=datetime.now(),
        )

        assert "한글" in skill.name

    def test_skill_data_unicode_content(self):
        """Test SkillData with unicode content."""
        skill = SkillData(
            name="test",
            frontmatter={},
            content="한글 コンテンツ",
            loaded_at=datetime.now(),
        )

        assert "한글" in skill.content


class TestSkillDataIntegration:
    """Integration tests for SkillData."""

    def test_skill_data_full_workflow(self):
        """Test complete skill data workflow."""
        skill = SkillData(
            name="test-skill",
            frontmatter={
                "version": "1.0.0",
                "capabilities": {"auth": True},
                "supported_efforts": [1, 3, 5],
            },
            content="# Skill\n## Quick\nQuick\n## Advanced\nAdv",
            loaded_at=datetime.now(),
        )

        # Test capability
        assert skill.get_capability("auth") is True

        # Test effort support
        assert skill.supports_effort(1) is True
        assert skill.supports_effort(2) is False

        # Test filtering
        skill.apply_filter("basic")
        assert "basic" in skill.applied_filters

    def test_skill_data_multiple_capabilities(self):
        """Test skill with multiple capabilities."""
        skill = SkillData(
            name="test",
            frontmatter={
                "capabilities": {
                    "auth": True,
                    "caching": True,
                    "logging": False,
                }
            },
            content="",
            loaded_at=datetime.now(),
        )

        assert skill.get_capability("auth") is True
        assert skill.get_capability("caching") is True
        assert skill.get_capability("logging") is False
