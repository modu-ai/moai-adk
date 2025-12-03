"""
Test suite for MoAI-ADK Skill Loading System

Tests skill validation, loading, caching, dependency management, and error handling.
"""

import unittest
from unittest.mock import patch, mock_open, MagicMock
import tempfile
import os
import json
from datetime import datetime, timedelta

# Import the skill loading system
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from moai_adk.core.skill_loading_system import (
    SkillLoader, SkillValidator, SkillRegistry, SkillData,
    load_skill, get_skill_cache_stats, clear_skill_cache,
    SkillLoadingError, SkillNotFoundError, SkillValidationError, DependencyError,
    LRUCache
)


class TestLRUCache(unittest.TestCase):
    """Test LRU cache implementation"""

    def test_cache_basic_operations(self):
        """Test basic cache set/get operations"""
        cache = LRUCache(maxsize=2, ttl=1)

        # Test setting and getting values
        skill_data = SkillData("test-skill", {}, "content", datetime.now())
        cache.set("test-skill", skill_data)

        retrieved = cache.get("test-skill")
        self.assertIsNotNone(retrieved)
        self.assertEqual(retrieved.name, "test-skill")

    def test_cache_ttl_expiration(self):
        """Test cache TTL expiration"""
        cache = LRUCache(maxsize=10, ttl=0.001)  # 1ms TTL

        skill_data = SkillData("test-skill", {}, "content", datetime.now())
        cache.set("test-skill", skill_data)

        # Wait for expiration
        import time
        time.sleep(0.01)

        retrieved = cache.get("test-skill")
        self.assertIsNone(retrieved)

    def test_cache_maxsize_eviction(self):
        """Test cache eviction when maxsize is exceeded"""
        cache = LRUCache(maxsize=2, ttl=3600)

        # Fill cache to maxsize
        for i in range(3):
            skill_data = SkillData(f"skill-{i}", {}, "content", datetime.now())
            cache.set(f"skill-{i}", skill_data)

        # First skill should be evicted
        self.assertIsNone(cache.get("skill-0"))
        self.assertIsNotNone(cache.get("skill-1"))
        self.assertIsNotNone(cache.get("skill-2"))

    def test_cache_lru_behavior(self):
        """Test LRU behavior - recently used items stay longer"""
        cache = LRUCache(maxsize=2, ttl=3600)

        # Add items
        skill1 = SkillData("skill-1", {}, "content", datetime.now())
        skill2 = SkillData("skill-2", {}, "content", datetime.now())
        cache.set("skill-1", skill1)
        cache.set("skill-2", skill2)

        # Access skill1 to make it recently used
        cache.get("skill-1")

        # Add new skill, should evict skill2
        skill3 = SkillData("skill-3", {}, "content", datetime.now())
        cache.set("skill-3", skill3)

        # skill2 should be evicted, skill1 should remain
        self.assertIsNone(cache.get("skill-2"))
        self.assertIsNotNone(cache.get("skill-1"))
        self.assertIsNotNone(cache.get("skill-3"))


class TestSkillData(unittest.TestCase):
    """Test SkillData functionality"""

    def test_skill_data_creation(self):
        """Test SkillData creation and basic properties"""
        frontmatter = {"name": "test-skill", "version": "1.0.0"}
        content = "# Test Skill\n\nThis is a test skill."

        skill_data = SkillData("test-skill", frontmatter, content, datetime.now())

        self.assertEqual(skill_data.name, "test-skill")
        self.assertEqual(skill_data.frontmatter["version"], "1.0.0")
        self.assertEqual(skill_data.content, content)

    def test_skill_data_effort_support(self):
        """Test effort support checking"""
        # Skill with default effort support
        skill1 = SkillData("skill-1", {}, "content", datetime.now())
        self.assertTrue(skill1.supports_effort(1))
        self.assertTrue(skill1.supports_effort(3))
        self.assertTrue(skill1.supports_effort(5))
        self.assertFalse(skill1.supports_effort(2))

        # Skill with custom effort support
        frontmatter = {"supported_efforts": [1, 3]}
        skill2 = SkillData("skill-2", frontmatter, "content", datetime.now())
        self.assertTrue(skill2.supports_effort(1))
        self.assertTrue(skill2.supports_effort(3))
        self.assertFalse(skill2.supports_effort(5))

    def test_skill_data_filtering(self):
        """Test content filtering based on effort"""
        content = """# Test Skill

## Quick Reference
Basic information

## Implementation Guide
Detailed guide

## Advanced Implementation
Advanced details
"""
        skill_data = SkillData("test-skill", {}, content, datetime.now())

        # Test basic filter
        skill_data.apply_filter('basic')
        filtered_content = skill_data.content
        self.assertIn('Quick Reference', filtered_content)
        self.assertNotIn('Advanced Implementation', filtered_content)

    def test_empty_skill_creation(self):
        """Test fallback skill creation"""
        fallback = SkillData.get_empty_skill("missing-skill")

        self.assertEqual(fallback.name, "missing-skill")
        self.assertEqual(fallback.frontmatter['status'], 'fallback')
        self.assertIn("failed to load", fallback.content)


class TestSkillRegistry(unittest.TestCase):
    """Test SkillRegistry functionality"""

    def test_skill_registration(self):
        """Test skill registration"""
        registry = SkillRegistry()

        metadata = {"name": "test-skill", "version": "1.0.0", "requires": ["base-skill"]}
        registry.register_skill("test-skill", metadata)

        self.assertIn("test-skill", registry.skills)
        self.assertEqual(registry.skills["test-skill"]["version"], "1.0.0")
        self.assertEqual(registry.dependencies["test-skill"], ["base-skill"])

    def test_skill_metadata_retrieval(self):
        """Test skill metadata retrieval"""
        registry = SkillRegistry()

        metadata = {"name": "test-skill", "version": "1.0.0"}
        registry.register_skill("test-skill", metadata)

        retrieved = registry.get_skill_metadata("test-skill")
        self.assertEqual(retrieved["version"], "1.0.0")

        # Test non-existent skill
        self.assertIsNone(registry.get_skill_metadata("nonexistent"))

    def test_compatibility_checking(self):
        """Test skill compatibility checking"""
        registry = SkillRegistry()

        # Register skills with compatibility info
        registry.register_skill("skill-a", {})
        registry.register_skill("skill-b", {})

        # Set up compatibility matrix
        registry.compatibility_matrix["skill-a"]["skill-b"] = True
        registry.compatibility_matrix["skill-a"]["skill-c"] = False

        self.assertTrue(registry.check_compatibility("skill-a", ["skill-b"]))
        self.assertFalse(registry.check_compatibility("skill-a", ["skill-c"]))


class TestSkillValidator(unittest.TestCase):
    """Test SkillValidator functionality"""

    def setUp(self):
        """Set up test registry and validator"""
        self.registry = SkillRegistry()
        self.validator = SkillValidator(self.registry)

        # Register test skills
        self.registry.register_skill("moai-test-skill", {
            "name": "moai-test-skill",
            "supported_efforts": [1, 3, 5],
            "requires": []
        })

        self.registry.register_skill("skill-with-deps", {
            "name": "skill-with-deps",
            "supported_efforts": [3, 5],
            "requires": ["moai-test-skill"]
        })

    def test_skill_name_validation(self):
        """Test skill name format validation"""
        # Valid names
        self.assertTrue(self.validator.validate_skill_name("moai-test-skill"))

        # Invalid name
        with self.assertRaises(SkillValidationError):
            self.validator.validate_skill_name("invalid_skill_name")

        # Non-existent skill
        with self.assertRaises(SkillValidationError):
            self.validator.validate_skill_name("moai-nonexistent")

    def test_effort_parameter_validation(self):
        """Test effort parameter validation"""
        # Valid effort
        self.assertTrue(self.validator.validate_effort_parameter("moai-test-skill", 1))
        self.assertTrue(self.validator.validate_effort_parameter("moai-test-skill", 3))
        self.assertTrue(self.validator.validate_effort_parameter("moai-test-skill", 5))

        # Invalid effort level
        with self.assertRaises(SkillValidationError):
            self.validator.validate_effort_parameter("moai-test-skill", 2)

        # Unsupported effort for skill
        with self.assertRaises(SkillValidationError):
            self.validator.validate_effort_parameter("skill-with-deps", 1)

    def test_dependency_validation(self):
        """Test dependency validation"""
        # Dependencies satisfied
        self.assertTrue(self.validator.validate_dependencies("skill-with-deps", ["moai-test-skill"]))

        # Missing dependencies
        with self.assertRaises(DependencyError):
            self.validator.validate_dependencies("skill-with-deps", [])


class TestSkillLoader(unittest.TestCase):
    """Test SkillLoader functionality"""

    def setUp(self):
        """Set up test environment with temporary files"""
        self.temp_dir = tempfile.mkdtemp()
        self.skills_dir = os.path.join(self.temp_dir, "skills")
        os.makedirs(self.skills_dir)

        # Create test skill file
        test_skill_dir = os.path.join(self.skills_dir, "moai-test-skill")
        os.makedirs(test_skill_dir)

        skill_content = """---
name: moai-test-skill
version: 1.0.0
supported_efforts: [1, 3, 5]
requires: []
---

# Test Skill

## Quick Reference
Basic information

## Implementation Guide
Detailed guide
"""

        with open(os.path.join(test_skill_dir, "SKILL.md"), "w") as f:
            f.write(skill_content)

        # Initialize skill loader with test directory
        self.loader = SkillLoader([self.skills_dir])

    def tearDown(self):
        """Clean up temporary files"""
        import shutil
        shutil.rmtree(self.temp_dir)

    def test_skill_loading_success(self):
        """Test successful skill loading"""
        skill_data = self.loader.load_skill("moai-test-skill")

        self.assertIsNotNone(skill_data)
        self.assertEqual(skill_data.name, "moai-test-skill")
        self.assertEqual(skill_data.frontmatter["version"], "1.0.0")
        self.assertIn("Test Skill", skill_data.content)

    def test_skill_loading_with_effort(self):
        """Test skill loading with effort parameter"""
        # Load with effort=1 (basic)
        skill_data = self.loader.load_skill("moai-test-skill", effort=1)

        self.assertIn("basic", skill_data.applied_filters)
        self.assertIn("Quick Reference", skill_data.content)
        # Advanced content should be filtered out for basic effort

    def test_skill_caching(self):
        """Test skill caching behavior"""
        # Load skill first time
        skill_data1 = self.loader.load_skill("moai-test-skill")

        # Load skill second time (should use cache)
        skill_data2 = self.loader.load_skill("moai-test-skill")

        self.assertEqual(id(skill_data1), id(skill_data2))  # Same object from cache

    def test_skill_force_reload(self):
        """Test force reload bypassing cache"""
        # Load skill first time
        skill_data1 = self.loader.load_skill("moai-test-skill")

        # Force reload
        skill_data2 = self.loader.load_skill("moai-test-skill", force_reload=True)

        self.assertNotEqual(id(skill_data1), id(skill_data2))  # Different objects

    def test_nonexistent_skill_fallback(self):
        """Test fallback behavior for non-existent skills"""
        fallback_skill = self.loader.load_skill("moai-nonexistent")

        self.assertEqual(fallback_skill.name, "moai-nonexistent")
        self.assertEqual(fallback_skill.frontmatter['status'], 'fallback')

    def test_circular_dependency_detection(self):
        """Test circular dependency detection"""
        # This test would require more complex setup with circular dependencies
        # For now, just test the detection mechanism exists
        with self.assertRaises(SkillLoadingError):
            # Mock a scenario that would cause circular dependency
            self.loader.loading_stack.append("test-skill")
            self.loader.load_skill("test-skill")


class TestPublicAPI(unittest.TestCase):
    """Test public API functions"""

    @patch('moai_adk.core.skill_loading_system.SKILL_LOADER')
    def test_load_skill_function(self, mock_loader):
        """Test public load_skill function"""
        mock_skill_data = MagicMock()
        mock_loader.load_skill.return_value = mock_skill_data

        result = load_skill("test-skill", effort=3)

        mock_loader.load_skill.assert_called_once_with("test-skill", effort=3, force_reload=False)
        self.assertEqual(result, mock_skill_data)

    @patch('moai_adk.core.skill_loading_system.SKILL_LOADER')
    def test_get_skill_cache_stats(self, mock_loader):
        """Test get_skill_cache_stats function"""
        mock_stats = {"cached_skills": ["skill1"], "cache_size": 1}
        mock_loader.get_cache_stats.return_value = mock_stats

        result = get_skill_cache_stats()

        self.assertEqual(result, mock_stats)

    @patch('moai_adk.core.skill_loading_system.SKILL_LOADER')
    def test_clear_skill_cache(self, mock_loader):
        """Test clear_skill_cache function"""
        clear_skill_cache()

        mock_loader.cache.clear.assert_called_once()


class TestSkillDetection(unittest.TestCase):
    """Test automatic skill detection functions"""

    def test_detect_required_skills(self):
        """Test automatic skill detection based on agent and prompt"""
        from moai_adk.core.skill_loading_system import detect_required_skills

        # Test backend agent
        skills = detect_required_skills("expert-backend", "Create Python API")
        self.assertIn("moai-foundation-claude", skills)
        self.assertIn("moai-lang-unified", skills)
        self.assertIn("moai-domain-backend", skills)

    def test_detect_skills_from_prompt(self):
        """Test skill detection from prompt content"""
        from moai_adk.core.skill_loading_system import _detect_skills_from_prompt

        # Test Python detection
        skills = _detect_skills_from_prompt("Create a FastAPI application")
        self.assertIn("moai-lang-unified", skills)

        # Test frontend detection
        skills = _detect_skills_from_prompt("Build React component with TypeScript")
        self.assertIn("moai-lang-unified", skills)

        # Test security detection
        skills = _detect_skills_from_prompt("Implement authentication and security")
        self.assertIn("moai-quality-security", skills)


if __name__ == "__main__":
    unittest.main()