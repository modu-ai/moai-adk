#!/usr/bin/env python3
"""
Simple test runner for ResearchManager
"""

import sys
import tempfile
import json
from pathlib import Path

# Add path for imports
sys.path.insert(0, str(Path(__file__).parent / ".claude" / "hooks" / "alfred" / "managers"))

# Test basic functionality
def test_basic_import():
    """Test that ResearchManager can be imported"""
    try:
        from research_manager import ResearchManager
        print("✅ ResearchManager import successful")
        return True
    except ImportError as e:
        print(f"❌ ResearchManager import failed: {e}")
        return False

def test_initialization():
    """Test ResearchManager initialization"""
    try:
        from research_manager import ResearchManager

        # Create temp directory for testing
        temp_dir = Path(tempfile.mkdtemp())

        # Change to temp directory
        import os
        original_cwd = os.getcwd()
        os.chdir(temp_dir)

        try:
            # Create .moai structure
            moai_dir = Path(temp_dir) / ".moai"
            moai_dir.mkdir(exist_ok=True)

            # Create config file
            config = {"research": {"auto_discovery": True}}
            with open(moai_dir / "config.json", 'w') as f:
                json.dump(config, f)

            # Test initialization
            manager = ResearchManager()
            print("✅ ResearchManager initialization successful")
            print(f"   - Cache TTL: {manager.cache_ttl}s")
            print(f"   - Base dir: {manager.base_dir}")
            return True

        finally:
            os.chdir(original_cwd)
            import shutil
            shutil.rmtree(temp_dir, ignore_errors=True)

    except Exception as e:
        print(f"❌ ResearchManager initialization failed: {e}")
        return False

def test_tool_classification():
    """Test tool type classification"""
    try:
        from research_manager import ResearchManager

        # Create temp directory
        temp_dir = Path(tempfile.mkdtemp())
        import os
        original_cwd = os.getcwd()
        os.chdir(temp_dir)

        try:
            # Create .moai structure
            moai_dir = Path(temp_dir) / ".moai"
            moai_dir.mkdir(exist_ok=True)
            config = {"research": {"auto_discovery": True}}
            with open(moai_dir / "config.json", 'w') as f:
                json.dump(config, f)

            manager = ResearchManager()

            # Test tool classification
            test_cases = [
                ("Edit", {}, "code"),
                ("Write", {}, "code"),
                ("Bash", {}, "test"),
                ("Task", {}, "exploration"),
                ("WebFetch", {}, "exploration"),
                ("UnknownTool", {}, "general")
            ]

            all_passed = True
            for tool_name, tool_args, expected in test_cases:
                result = manager.classify_tool_type(tool_name, tool_args)
                if result == expected:
                    print(f"✅ {tool_name} -> {result}")
                else:
                    print(f"❌ {tool_name} -> {result} (expected {expected})")
                    all_passed = False

            return all_passed

        finally:
            os.chdir(original_cwd)
            import shutil
            shutil.rmtree(temp_dir, ignore_errors=True)

    except Exception as e:
        print(f"❌ Tool classification test failed: {e}")
        return False

def test_research_strategies():
    """Test research strategies"""
    try:
        from research_manager import ResearchManager

        # Create temp directory
        temp_dir = Path(tempfile.mkdtemp())
        import os
        original_cwd = os.getcwd()
        os.chdir(temp_dir)

        try:
            # Create .moai structure
            moai_dir = Path(temp_dir) / ".moai"
            moai_dir.mkdir(exist_ok=True)
            config = {"research": {"auto_discovery": True}}
            with open(moai_dir / "config.json", 'w') as f:
                json.dump(config, f)

            manager = ResearchManager()

            # Test strategy retrieval
            strategies = manager.get_research_strategies_for_tool("code")

            required_keys = ["primary_strategies", "secondary_strategies", "focus_areas", "knowledge_categories"]
            all_present = all(key in strategies for key in required_keys)

            if all_present:
                print("✅ Research strategies contain all required keys")
                print(f"   - Primary strategies: {len(strategies['primary_strategies'])}")
                print(f"   - Focus areas: {len(strategies['focus_areas'])}")
                return True
            else:
                print(f"❌ Research strategies missing keys: {required_keys}")
                return False

        finally:
            os.chdir(original_cwd)
            import shutil
            shutil.rmtree(temp_dir, ignore_errors=True)

    except Exception as e:
        print(f"❌ Research strategies test failed: {e}")
        return False

def main():
    """Run all tests"""
    print("🔬 ResearchManager Simple Test Runner")
    print("=" * 50)

    tests = [
        ("Import Test", test_basic_import),
        ("Initialization Test", test_initialization),
        ("Tool Classification Test", test_tool_classification),
        ("Research Strategies Test", test_research_strategies)
    ]

    passed = 0
    total = len(tests)

    for test_name, test_func in tests:
        print(f"\n📋 Running {test_name}...")
        if test_func():
            passed += 1
        else:
            print(f"   ⚠️ {test_name} failed")

    print(f"\n📊 Test Results: {passed}/{total} passed")

    if passed == total:
        print("🎉 All tests passed! ResearchManager is working correctly.")
        return True
    else:
        print("❌ Some tests failed. Please check the implementation.")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)