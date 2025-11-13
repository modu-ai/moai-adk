#!/usr/bin/env python3
"""
Statusline Output Style Detection Test Script

This script tests the enhanced output style detection system
by simulating various scenarios and validating the detection accuracy.
"""

import json
import os
import sys
import tempfile
import time
from pathlib import Path

# Add the package to the path so we can import the modules
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "src"))

from moai_adk.statusline.enhanced_output_style_detector import OutputStyleDetector


class StatuslineTester:
    """
    Test framework for the enhanced output style detection system.
    """

    def __init__(self):
        self.test_results = []
        self.detector = OutputStyleDetector()

    def run_test(self, test_name: str, test_func, expected_result: str = None):
        """
        Run a single test case and record the result.
        """
        try:
            print(f"Running test: {test_name}...", end=" ")
            start_time = time.time()

            result = test_func()
            duration = time.time() - start_time

            if expected_result:
                success = result == expected_result
                status = "✓ PASS" if success else f"✗ FAIL (expected: {expected_result}, got: {result})"
            else:
                # For tests without expected result, just check that we get a valid result
                success = bool(result) and result != "Unknown"
                status = "✓ PASS" if success else f"✗ FAIL (got: {result})"

            self.test_results.append({
                "name": test_name,
                "result": result,
                "expected": expected_result,
                "success": success,
                "duration": duration
            })

            print(f"{status} ({duration:.3f}s)")

        except Exception as e:
            print(f"✗ ERROR: {e}")
            self.test_results.append({
                "name": test_name,
                "result": None,
                "expected": expected_result,
                "success": False,
                "duration": 0,
                "error": str(e)
            })

    def test_session_context_detection(self):
        """Test detection from session context."""
        # Test case 1: Explicit output style in context
        session_data = {"outputStyle": "explanatory"}
        result = self.detector.detect_from_session_context(session_data)
        assert result == "Explanatory", f"Expected 'Explanatory', got '{result}'"
        return result

    def test_model_name_detection(self):
        """Test detection from model configuration."""
        # Test Yoda detection
        session_data = {
            "model": {
                "name": "claude-yoda-tutorial",
                "display_name": "Yoda Master"
            }
        }
        result = self.detector.detect_from_session_context(session_data)
        return result

    def test_environment_detection(self):
        """Test detection from environment variables."""
        # Temporarily set environment variable
        original_style = os.environ.get("CLAUDE_OUTPUT_STYLE")
        os.environ["CLAUDE_OUTPUT_STYLE"] = "explanatory"

        try:
            result = self.detector.detect_from_environment()
        finally:
            # Restore original environment
            if original_style is None:
                os.environ.pop("CLAUDE_OUTPUT_STYLE", None)
            else:
                os.environ["CLAUDE_OUTPUT_STYLE"] = original_style

        return result

    def test_settings_file_detection(self):
        """Test detection from settings.json file."""
        # Create temporary settings file
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            settings = {
                "outputStyle": "🧙 Yoda Master"
            }
            json.dump(settings, f)
            temp_settings_path = f.name

        try:
            # Temporarily override the detector's settings path
            original_cwd = Path.cwd()
            temp_dir = Path(temp_settings_path).parent
            os.chdir(temp_dir)

            # Create .claude directory and move settings file
            claude_dir = temp_dir / ".claude"
            claude_dir.mkdir(exist_ok=True)
            settings_path = claude_dir / "settings.json"
            Path(temp_settings_path).rename(settings_path)

            # Test detection
            result = self.detector.detect_from_settings()

        finally:
            # Cleanup
            os.chdir(original_cwd)
            if settings_path.exists():
                settings_path.unlink()
            if claude_dir.exists():
                claude_dir.rmdir()

        return result

    def test_style_normalization(self):
        """Test style name normalization."""
        test_cases = [
            ("streaming", "R2-D2"),
            ("explanatory", "Explanatory"),
            ("🤖 R2-D2", "R2-D2"),
            ("🧙 Yoda Master", "🧙 Yoda Master"),
            ("r2d2", "R2-D2"),
            ("YODA", "🧙 Yoda Master"),
        ]

        for input_style, expected in test_cases:
            result = self.detector._normalize_style(input_style)
            assert result == expected, f"Input '{input_style}': expected '{expected}', got '{result}'"

        return "All normalization tests passed"

    def test_message_pattern_analysis(self):
        """Test message pattern analysis for style detection."""
        # Test Yoda patterns
        yoda_messages = [
            {"role": "assistant", "content": "Young padawan, let me explain the force of this code."}
        ]
        result = self.detector._analyze_message_patterns(yoda_messages)

        # Test Explanatory patterns
        explanatory_messages = [
            {"role": "assistant", "content": "Let me explain in detail how this function works. Here's the reasoning behind this implementation..."}
        ]
        result2 = self.detector._analyze_message_patterns(explanatory_messages)

        return f"Yoda: {result}, Explanatory: {result2}"

    def test_cache_functionality(self):
        """Test caching functionality."""
        # First call should populate cache
        session_data = {"outputStyle": "concise"}
        result1 = self.detector.get_output_style(session_data)

        # Second call should use cache
        start_time = time.time()
        result2 = self.detector.get_output_style(session_data)
        cache_time = time.time() - start_time

        assert result1 == result2, "Cache returned different result"
        assert cache_time < 0.001, f"Cache too slow: {cache_time}s"

        return f"Cache working correctly (result: {result1}, time: {cache_time:.6f}s)"

    def test_full_integration(self):
        """Test the complete integration with all detection methods."""
        # Set up environment
        os.environ["CLAUDE_OUTPUT_STYLE"] = "detailed"

        # Create session context
        session_data = {
            "outputStyle": "explanatory",
            "model": {"name": "claude-explanatory"}
        }

        try:
            result = self.detector.get_output_style(session_data)
        finally:
            # Cleanup
            os.environ.pop("CLAUDE_OUTPUT_STYLE", None)

        return result

    def run_all_tests(self):
        """
        Run all test cases and generate a report.
        """
        print("🧪 Starting Statusline Output Style Detection Tests")
        print("=" * 60)

        # Define test cases
        test_cases = [
            ("Session Context Detection", self.test_session_context_detection, "Explanatory"),
            ("Model Name Detection", self.test_model_name_detection, "🧙 Yoda Master"),
            ("Environment Detection", self.test_environment_detection, "Explanatory"),
            ("Settings File Detection", self.test_settings_file_detection, "🧙 Yoda Master"),
            ("Style Normalization", self.test_style_normalization, "All normalization tests passed"),
            ("Message Pattern Analysis", self.test_message_pattern_analysis, None),
            ("Cache Functionality", self.test_cache_functionality, None),
            ("Full Integration Test", self.test_full_integration, "Explanatory"),
        ]

        # Run all tests
        for test_name, test_func, expected in test_cases:
            self.run_test(test_name, test_func, expected)

        # Generate report
        self.generate_report()

    def generate_report(self):
        """
        Generate a comprehensive test report.
        """
        print("\n" + "=" * 60)
        print("📊 TEST REPORT")
        print("=" * 60)

        total_tests = len(self.test_results)
        passed_tests = sum(1 for r in self.test_results if r["success"])
        failed_tests = total_tests - passed_tests
        total_duration = sum(r["duration"] for r in self.test_results)

        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests}")
        print(f"Failed: {failed_tests}")
        print(f"Success Rate: {(passed_tests/total_tests*100):.1f}%")
        print(f"Total Duration: {total_duration:.3f}s")
        print(f"Average Test Time: {(total_duration/total_tests):.3f}s")

        if failed_tests > 0:
            print("\n❌ FAILED TESTS:")
            for result in self.test_results:
                if not result["success"]:
                    error_msg = f" - {result['name']}"
                    if "error" in result:
                        error_msg += f": {result['error']}"
                    else:
                        error_msg += f" (expected: {result['expected']}, got: {result['result']})"
                    print(error_msg)

        # Performance analysis
        print("\n⚡ PERFORMANCE ANALYSIS:")
        sorted_tests = sorted(self.test_results, key=lambda x: x["duration"], reverse=True)
        print("Slowest Tests:")
        for i, test in enumerate(sorted_tests[:3]):
            print(f"  {i+1}. {test['name']}: {test['duration']:.3f}s")

        # Save detailed report
        report_data = {
            "timestamp": time.time(),
            "summary": {
                "total": total_tests,
                "passed": passed_tests,
                "failed": failed_tests,
                "success_rate": passed_tests/total_tests,
                "total_duration": total_duration
            },
            "results": self.test_results
        }

        report_path = Path.cwd() / ".moai" / "temp" / "statusline_test_report.json"
        report_path.parent.mkdir(parents=True, exist_ok=True)
        report_path.write_text(json.dumps(report_data, indent=2))

        print(f"\n📁 Detailed report saved to: {report_path}")

        # Return success status
        return failed_tests == 0


def main():
    """
    Main entry point for the test script.
    """
    try:
        tester = StatuslineTester()
        success = tester.run_all_tests()

        if success:
            print("\n🎉 ALL TESTS PASSED!")
            sys.exit(0)
        else:
            print("\n💥 SOME TESTS FAILED!")
            sys.exit(1)

    except Exception as e:
        print(f"\n💥 TEST EXECUTION FAILED: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()