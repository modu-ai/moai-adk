#!/usr/bin/env python3
# @CODE:HOOK-RESEARCH-MANAGER-TEST | @SPEC:HOOK-RESEARCH-MANAGER-001 | @TEST: tests/hooks/managers/test_research_manager.py

"""
Research Manager Tests

TDD-based test suite for ResearchManager module that integrates all research-related hook functionality.
Tests follow RED-GREEN-REFACTOR methodology.
"""

import json
import time
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch, mock_open

# Add the path to import the ResearchManager
import sys
sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent))

# Import the module under test
try:
    from moai_adk.claude.hooks.alfred.managers.research_manager import ResearchManager
except ImportError:
    # Fallback for testing before implementation
    ResearchManager = None


class TestResearchManager(unittest.TestCase):
    """Test cases for ResearchManager class"""

    def setUp(self):
        """Set up test fixtures before each test method."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.moai_dir = self.temp_dir / ".moai"
        self.moai_dir.mkdir(exist_ok=True)

        # Create research directory structure
        (self.moai_dir / "research").mkdir(exist_ok=True)
        (self.moai_dir / "research" / "knowledge").mkdir(exist_ok=True)
        (self.moai_dir / "research" / "strategies").mkdir(exist_ok=True)
        (self.moai_dir / "research" / "analysis").mkdir(exist_ok=True)
        (self.moai_dir / "research" / "temp").mkdir(exist_ok=True)

        # Create config file
        self.config = {
            "research": {
                "auto_discovery": True,
                "pattern_matching": True,
                "auto_update_knowledge": True,
                "categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"]
            },
            "tags": {
                "research_tags": {
                    "auto_discovery": True,
                    "pattern_matching": True,
                    "research_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"]
                }
            }
        }

        with open(self.moai_dir / "config.json", 'w') as f:
            json.dump(self.config, f)

        # Test data
        self.test_tool_name = "Edit"
        self.test_tool_args = {"file_path": "/test/file.py", "content": "test content"}

        # Change to temp directory for tests
        self.original_cwd = Path.cwd()
        import os
        os.chdir(self.temp_dir)

    def tearDown(self):
        """Clean up after each test method."""
        import os
        os.chdir(self.original_cwd)
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_class_exists(self):
        """Test that ResearchManager class can be imported and instantiated."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        self.assertIsNotNone(manager)
        self.assertIsInstance(manager, ResearchManager)

    def test_initialization_with_default_config(self):
        """Test ResearchManager initialization with default configuration."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        self.assertIsNotNone(manager.config)
        self.assertIn("research", manager.config)
        self.assertTrue(manager.config["research"].get("auto_discovery", False))

    def test_initialization_with_custom_config(self):
        """Test ResearchManager initialization with custom configuration."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        custom_config = {
            "research": {
                "auto_discovery": False,
                "custom_strategies": ["test_strategy"]
            }
        }

        manager = ResearchManager(config=custom_config)

        self.assertFalse(manager.config["research"]["auto_discovery"])
        self.assertIn("test_strategy", manager.config["research"]["custom_strategies"])

    def test_classify_tool_type_code_tools(self):
        """Test tool classification for code-related tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        code_tools = ["Edit", "Write", "Read", "Grep", "Glob", "MultiEdit"]
        for tool in code_tools:
            with self.subTest(tool=tool):
                tool_type = manager.classify_tool_type(tool, {})
                self.assertEqual(tool_type, "code")

    def test_classify_tool_type_test_tools(self):
        """Test tool classification for test-related tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        test_tools = ["Bash"]
        for tool in test_tools:
            with self.subTest(tool=tool):
                tool_type = manager.classify_tool_type(tool, {})
                self.assertEqual(tool_type, "test")

    def test_classify_tool_type_exploration_tools(self):
        """Test tool classification for exploration-related tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        exploration_tools = ["Task", "Explore", "Plan", "WebFetch", "WebSearch"]
        for tool in exploration_tools:
            with self.subTest(tool=tool):
                tool_type = manager.classify_tool_type(tool, {})
                self.assertEqual(tool_type, "exploration")

    def test_get_research_strategies_for_code(self):
        """Test getting research strategies for code tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        strategies = manager.get_research_strategies_for_tool("code")

        self.assertIn("primary_strategies", strategies)
        self.assertIn("secondary_strategies", strategies)
        self.assertIn("focus_areas", strategies)
        self.assertIn("knowledge_categories", strategies)

        # Check that primary strategies include expected items
        primary = strategies["primary_strategies"]
        self.assertIn("pattern_recognition", primary)

    def test_get_research_strategies_for_exploration(self):
        """Test getting research strategies for exploration tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        strategies = manager.get_research_strategies_for_tool("exploration")

        self.assertIn("cross_domain_analysis", strategies["primary_strategies"])
        self.assertIn("comprehensive_analysis", strategies["focus_areas"])

    def test_load_jit_knowledge_with_matching_files(self):
        """Test JIT knowledge loading with matching knowledge files."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        # Create test knowledge files
        knowledge_dir = self.moai_dir / "research" / "knowledge"
        test_knowledge = {
            "patterns": ["test_pattern"],
            "insights": ["test_insight"]
        }

        with open(knowledge_dir / "code_patterns.json", 'w') as f:
            json.dump(test_knowledge, f)

        manager = ResearchManager()
        knowledge = manager.load_jit_knowledge("code", ["patterns"])

        self.assertIn("code_patterns", knowledge)
        self.assertEqual(knowledge["code_patterns"]["patterns"], ["test_pattern"])

    def test_load_jit_knowledge_no_files(self):
        """Test JIT knowledge loading when no files exist."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        knowledge = manager.load_jit_knowledge("unknown", [])

        self.assertEqual(len(knowledge), 0)

    def test_optimize_resources_for_code(self):
        """Test resource optimization for code tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        resources = manager.optimize_resources("code", ["pattern_recognition"])

        self.assertIn("memory_limit", resources)
        self.assertIn("timeout_seconds", resources)
        self.assertIn("priority", resources)
        self.assertIn("parallel_processing", resources)
        self.assertIn("cache_enabled", resources)

        # Code tools should have high priority
        self.assertEqual(resources["priority"], "high")

    def test_optimize_resources_for_documentation(self):
        """Test resource optimization for documentation tools."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        resources = manager.optimize_resources("documentation", [])

        # Documentation tools should have lower priority
        self.assertEqual(resources["priority"], "low")
        self.assertFalse(resources.get("parallel_processing", True))

    def test_create_research_context(self):
        """Test research context creation."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        strategy_config = manager.get_research_strategies_for_tool("code")
        knowledge_base = manager.load_jit_knowledge("code", [])
        resource_config = manager.optimize_resources("code", strategy_config["primary_strategies"])

        context = manager.create_research_context(
            self.test_tool_name,
            self.test_tool_args,
            strategy_config,
            knowledge_base,
            resource_config
        )

        self.assertIn("tool_context", context)
        self.assertIn("research_strategy", context)
        self.assertIn("knowledge_base", context)
        self.assertIn("resource_config", context)
        self.assertIn("session_id", context)
        self.assertIn("timestamp", context)

        self.assertEqual(context["tool_context"]["name"], self.test_tool_name)
        self.assertEqual(context["tool_context"]["type"], "code")

    def test_setup_research_environment(self):
        """Test research environment setup."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        setup_result = manager.setup_research_environment()

        self.assertIn("research_setup_completed", setup_result)
        self.assertIn("active_strategies", setup_result)
        self.assertIn("knowledge_base_size", setup_result)
        self.assertIn("resource_optimization", setup_result)
        self.assertIn("session_id", setup_result)

        self.assertTrue(setup_result["research_setup_completed"])

    def test_analyze_tool_result_success(self):
        """Test tool result analysis for successful execution."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        tool_result = {
            "continue": True,
            "output": "Success"
        }

        analysis = manager.analyze_tool_result(self.test_tool_name, tool_result, 100.0)

        self.assertIn("tool_name", analysis)
        self.assertIn("execution_successful", analysis)
        self.assertIn("insights", analysis)
        self.assertIn("execution_time_ms", analysis)

        self.assertEqual(analysis["tool_name"], self.test_tool_name)
        self.assertTrue(analysis["execution_successful"])

    def test_analyze_tool_result_failure(self):
        """Test tool result analysis for failed execution."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        tool_result = {
            "continue": False,
            "error": "Test error"
        }

        analysis = manager.analyze_tool_result(self.test_tool_name, tool_result, 100.0)

        self.assertFalse(analysis["execution_successful"])
        self.assertIn("error", analysis)

    def test_update_knowledge_base(self):
        """Test knowledge base update functionality."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()
        insights = ["Test insight 1", "Test insight 2"]
        analysis = {"execution_successful": True}

        update_result = manager.update_knowledge_base(insights, self.test_tool_name, analysis)

        self.assertIn("knowledge_updated", update_result)
        self.assertIn("knowledge_file", update_result)
        self.assertIn("insights_count", update_result)

        self.assertTrue(update_result["knowledge_updated"])
        self.assertEqual(update_result["insights_count"], 2)

    def test_cache_functionality(self):
        """Test caching functionality for performance optimization."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        # First call should cache the result
        result1 = manager.get_research_strategies_for_tool("code")

        # Second call should use cached result
        result2 = manager.get_research_strategies_for_tool("code")

        # Results should be identical
        self.assertEqual(result1, result2)

    def test_cache_ttl_expiration(self):
        """Test cache TTL expiration mechanism."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager(cache_ttl=0.1)  # 100ms TTL

        # First call
        result1 = manager.get_research_strategies_for_tool("code")

        # Wait for cache to expire
        time.sleep(0.2)

        # Second call should regenerate result
        result2 = manager.get_research_strategies_for_tool("code")

        # Results should still be identical but regenerated
        self.assertEqual(result1, result2)

    def test_performance_metrics_tracking(self):
        """Test performance metrics tracking."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        # Execute some operations
        manager.classify_tool_type(self.test_tool_name, self.test_tool_args)
        manager.get_research_strategies_for_tool("code")
        manager.load_jit_knowledge("code", [])

        metrics = manager.get_performance_metrics()

        self.assertIn("operations_count", metrics)
        self.assertIn("cache_hits", metrics)
        self.assertIn("cache_misses", metrics)
        self.assertIn("average_execution_time", metrics)


class TestResearchManagerIntegration(unittest.TestCase):
    """Integration tests for ResearchManager with existing hooks"""

    def setUp(self):
        """Set up integration test fixtures."""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.moai_dir = self.temp_dir / ".moai"
        self.moai_dir.mkdir(exist_ok=True)

        # Create minimal config
        with open(self.moai_dir / "config.json", 'w') as f:
            json.dump({"research": {"auto_discovery": True}}, f)

        self.original_cwd = Path.cwd()
        import os
        os.chdir(self.temp_dir)

    def tearDown(self):
        """Clean up integration test fixtures."""
        import os
        os.chdir(self.original_cwd)
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_integration_with_research_strategy_hook(self):
        """Test integration with pre_tool__research_strategy.py functionality."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        # Simulate pre_tool execution
        context = manager.execute_pre_tool_research("Edit", {"file_path": "test.py"})

        self.assertIn("tool_context", context)
        self.assertIn("research_strategy", context)
        self.assertIn("research_strategy_selected", context)
        self.assertTrue(context["research_strategy_selected"])

    def test_integration_with_research_analysis_hook(self):
        """Test integration with post_tool__research_analysis.py functionality."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        # Simulate post_tool execution
        tool_result = {"continue": True, "output": "Success"}
        analysis = manager.execute_post_tool_analysis("Edit", tool_result, 150.0)

        self.assertIn("research_analysis_completed", analysis)
        self.assertIn("tool_name", analysis)
        self.assertTrue(analysis["research_analysis_completed"])

    def test_integration_with_research_setup_hook(self):
        """Test integration with session_start__research_setup.py functionality."""
        if ResearchManager is None:
            self.skipTest("ResearchManager not yet implemented")

        manager = ResearchManager()

        # Simulate session start
        setup = manager.execute_session_setup()

        self.assertIn("research_setup_completed", setup)
        self.assertIn("active_strategies", setup)
        self.assertIn("session_id", setup)
        self.assertTrue(setup["research_setup_completed"])


if __name__ == '__main__':
    unittest.main()