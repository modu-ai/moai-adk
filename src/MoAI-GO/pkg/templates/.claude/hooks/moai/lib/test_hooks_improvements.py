#!/usr/bin/env python3
"""Test Suite for Hooks System Improvements

Comprehensive testing script to validate all the improvements made to prevent
aborted() errors and optimize timeout handling in the MoAI-ADK hooks system.

Test Coverage:
- Unified timeout manager functionality
- Git operations manager performance
- Configuration validation
- Hook-specific improvements
- Graceful degradation mechanisms
- Performance benchmarks
- Error handling scenarios
"""

import json
import subprocess
import sys
import tempfile
import time
from datetime import datetime
from pathlib import Path

# Add lib directory to path
hooks_dir = Path(__file__).parent
sys.path.insert(0, str(hooks_dir))

# Import modules to test
try:
    from config_validator import (  # noqa: F401 - imports used for availability check
        ValidationIssue,
        ValidationLevel,
        get_config_validator,
    )
    from git_operations_manager import (  # noqa: F401 - imports used for availability check
        GitCommand,
        GitOperationType,
        get_git_manager,
    )
    from unified_timeout_manager import (  # noqa: F401 - imports used for availability check
        HookTimeoutConfig,
        HookTimeoutError,
        TimeoutPolicy,
        get_timeout_manager,
    )

    TIMEOUT_MANAGER_AVAILABLE = True
    GIT_MANAGER_AVAILABLE = True
    CONFIG_VALIDATOR_AVAILABLE = True
except ImportError as e:
    print(f"⚠️ Import error: {e}")
    TIMEOUT_MANAGER_AVAILABLE = False
    GIT_MANAGER_AVAILABLE = False
    CONFIG_VALIDATOR_AVAILABLE = False


class HooksTestSuite:
    """Comprehensive test suite for hooks improvements"""

    def __init__(self):
        self.test_results = []
        self.start_time = datetime.now()
        self.temp_dir = Path(tempfile.mkdtemp(prefix="hooks_test_"))

    def log_test_result(self, test_name: str, passed: bool, message: str = "", duration: float = 0):
        """Log a test result"""
        status = "✅ PASS" if passed else "❌ FAIL"
        result = {
            "test_name": test_name,
            "status": status,
            "passed": passed,
            "message": message,
            "duration": duration,
            "timestamp": datetime.now().isoformat(),
        }
        self.test_results.append(result)
        print(f"{status} {test_name}{f' - {message}' if message else ''} ({duration:.3f}s)")

    def test_timeout_manager_basic(self) -> None:
        """Test basic timeout manager functionality"""
        if not TIMEOUT_MANAGER_AVAILABLE:
            self.log_test_result("Timeout Manager - Basic", False, "Module not available")
            return

        start = time.time()
        try:
            manager = get_timeout_manager()
            assert manager is not None, "Timeout manager should be available"

            # Test timeout configuration
            config = HookTimeoutConfig(
                policy=TimeoutPolicy.FAST,
                custom_timeout_ms=1000,
                retry_count=1,
                graceful_degradation=True,
            )

            # Test successful execution
            def quick_task():
                return "success"

            result = manager.execute_with_timeout("test_hook", quick_task, config=config)
            assert result == "success", "Quick task should return success"

            duration = time.time() - start
            self.log_test_result("Timeout Manager - Basic", True, "All basic tests passed", duration)

        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Timeout Manager - Basic", False, str(e), duration)

    def test_timeout_manager_timeout(self) -> None:
        """Test timeout manager timeout handling"""
        if not TIMEOUT_MANAGER_AVAILABLE:
            self.log_test_result("Timeout Manager - Timeout", False, "Module not available")
            return

        start = time.time()
        try:
            manager = get_timeout_manager()

            # Test timeout with slow task
            def slow_task():
                time.sleep(2)  # Sleep longer than timeout
                return "should not reach"

            config = HookTimeoutConfig(
                policy=TimeoutPolicy.FAST,
                custom_timeout_ms=500,  # 0.5 second timeout
                retry_count=0,  # No retries for timeout test
                graceful_degradation=True,
            )

            try:
                result = manager.execute_with_timeout("test_timeout", slow_task, config=config)
                # If graceful degradation is enabled, should return default result
                assert "graceful_degradation" in str(result), "Should return graceful degradation result"
                duration = time.time() - start
                self.log_test_result(
                    "Timeout Manager - Timeout",
                    True,
                    "Graceful degradation working",
                    duration,
                )

            except HookTimeoutError:
                duration = time.time() - start
                self.log_test_result(
                    "Timeout Manager - Timeout",
                    True,
                    "Timeout exception raised correctly",
                    duration,
                )

        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Timeout Manager - Timeout", False, str(e), duration)

    def test_git_operations_manager(self) -> None:
        """Test Git operations manager functionality"""
        if not GIT_MANAGER_AVAILABLE:
            self.log_test_result("Git Operations Manager", False, "Module not available")
            return

        start = time.time()
        try:
            manager = get_git_manager()
            assert manager is not None, "Git manager should be available"

            # Test git info retrieval
            project_info = manager.get_project_info(use_cache=True)
            assert isinstance(project_info, dict), "Should return dictionary"
            assert "branch" in project_info, "Should contain branch info"

            # Test statistics
            stats = manager.get_statistics()
            assert isinstance(stats, dict), "Statistics should be dictionary"
            assert "operations" in stats, "Should contain operations stats"

            # Test cache clearing
            cleared_count = manager.clear_cache()
            assert isinstance(cleared_count, int), "Should return count of cleared entries"

            duration = time.time() - start
            self.log_test_result(
                "Git Operations Manager",
                True,
                f"Cached {len(project_info)} fields",
                duration,
            )

        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Git Operations Manager", False, str(e), duration)

    def test_config_validator(self) -> None:
        """Test configuration validator functionality"""
        if not CONFIG_VALIDATOR_AVAILABLE:
            self.log_test_result("Configuration Validator", False, "Module not available")
            return

        start = time.time()
        try:
            validator = get_config_validator()
            assert validator is not None, "Config validator should be available"

            # Load example config
            example_config_path = hooks_dir / "example_config.json"
            assert example_config_path.exists(), "Example config should exist"

            with open(example_config_path, "r") as f:
                config = json.load(f)

            # Validate configuration
            is_valid, issues = validator.validate_config(config)

            # Check that critical issues are not present
            critical_issues = [i for i in issues if i.level == ValidationLevel.CRITICAL]
            error_issues = [i for i in issues if i.level == ValidationLevel.ERROR]

            success = len(critical_issues) == 0 and len(error_issues) == 0
            message = f"Found {len(issues)} issues ({len(error_issues)} errors)" if issues else "No issues found"

            # Test normalization
            normalized = validator.normalize_config({})
            assert isinstance(normalized, dict), "Normalization should return dict"

            duration = time.time() - start
            self.log_test_result("Configuration Validator", success, message, duration)

        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Configuration Validator", False, str(e), duration)

    def test_hook_integration(self) -> None:
        """Test integration with actual hooks"""
        start = time.time()
        try:
            # Test session_start hook
            session_start_hook = hooks_dir / "session_start__show_project_info.py"
            assert session_start_hook.exists(), "Session start hook should exist"

            # Test with empty input
            result = subprocess.run(
                [sys.executable, str(session_start_hook)],
                input="{}",
                text=True,
                capture_output=True,
                timeout=10,
                cwd=self.temp_dir,
            )

            # Should not crash and should produce JSON output
            try:
                output = json.loads(result.stdout)
                assert isinstance(output, dict), "Should return JSON object"
                success = result.returncode == 0
                message = "Hook executed successfully" if success else f"Exit code: {result.returncode}"
            except json.JSONDecodeError:
                success = False
                message = "Invalid JSON output"

            duration = time.time() - start
            self.log_test_result("Hook Integration - Session Start", success, message, duration)

        except subprocess.TimeoutExpired:
            duration = time.time() - start
            self.log_test_result("Hook Integration - Session Start", False, "Hook timeout", duration)
        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Hook Integration - Session Start", False, str(e), duration)

    def test_performance_benchmarks(self) -> None:
        """Test performance benchmarks"""
        start = time.time()
        try:
            # Test multiple parallel operations
            if GIT_MANAGER_AVAILABLE:
                manager = get_git_manager()

                # Benchmark git info retrieval
                git_start = time.time()
                for _ in range(5):
                    manager.get_project_info(use_cache=True)
                git_duration = time.time() - git_start

                avg_git_time = git_duration / 5
                message = f"Avg Git info retrieval: {avg_git_time:.3f}s"
                success = avg_git_time < 1.0  # Should be under 1 second with caching
            else:
                avg_git_time = 0
                message = "Git manager not available"
                success = False

            duration = time.time() - start
            self.log_test_result("Performance Benchmarks", success, message, duration)

        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Performance Benchmarks", False, str(e), duration)

    def test_error_handling(self) -> None:
        """Test error handling scenarios"""
        start = time.time()
        try:
            errors_tested = 0
            errors_passed = 0

            # Test invalid configuration
            if CONFIG_VALIDATOR_AVAILABLE:
                validator = get_config_validator()
                invalid_config = {
                    "timeout_manager": {
                        "global_timeout_ms": -1000,  # Invalid negative timeout
                        "default_retry_count": "invalid",  # Invalid type
                        "graceful_degradation": "maybe",  # Invalid boolean
                    }
                }

                is_valid, issues = validator.validate_config(invalid_config)
                errors_tested += 1
                if not is_valid and len(issues) > 0:
                    errors_passed += 1

            # Test timeout with graceful degradation
            if TIMEOUT_MANAGER_AVAILABLE:
                manager = get_timeout_manager()
                config = HookTimeoutConfig(
                    policy=TimeoutPolicy.FAST,
                    custom_timeout_ms=100,
                    retry_count=0,
                    graceful_degradation=True,
                )

                def failing_task():
                    time.sleep(0.5)
                    raise Exception("Simulated failure")

                try:
                    result = manager.execute_with_timeout("test_error", failing_task, config=config)
                    errors_tested += 1
                    # Should return graceful degradation result
                    if isinstance(result, dict) and "graceful_degradation" in str(result):
                        errors_passed += 1
                except Exception:
                    # Exception is also acceptable
                    errors_passed += 1
                    errors_tested += 1

            success = errors_passed > 0 and errors_passed == errors_tested
            message = f"Error handling: {errors_passed}/{errors_tested} scenarios passed"

            duration = time.time() - start
            self.log_test_result("Error Handling", success, message, duration)

        except Exception as e:
            duration = time.time() - start
            self.log_test_result("Error Handling", False, str(e), duration)

    def run_all_tests(self) -> None:
        """Run all tests in the suite"""
        print("🚀 Starting Hooks System Improvements Test Suite")
        print(f"📍 Test directory: {self.temp_dir}")
        print(f"🕐 Started at: {self.start_time}")
        print("=" * 60)

        # Run all tests
        self.test_timeout_manager_basic()
        self.test_timeout_manager_timeout()
        self.test_git_operations_manager()
        self.test_config_validator()
        self.test_hook_integration()
        self.test_performance_benchmarks()
        self.test_error_handling()

        # Generate summary
        self.generate_summary()

    def generate_summary(self) -> None:
        """Generate test summary report"""
        total_tests = len(self.test_results)
        passed_tests = sum(1 for r in self.test_results if r["passed"])
        failed_tests = total_tests - passed_tests
        total_duration = sum(r["duration"] for r in self.test_results)

        end_time = datetime.now()
        suite_duration = (end_time - self.start_time).total_seconds()

        print("\n" + "=" * 60)
        print("📊 TEST SUMMARY")
        print("=" * 60)
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {(passed_tests / total_tests * 100):.1f}%")
        print(f"Total Duration: {total_duration:.3f}s")
        print(f"Suite Duration: {suite_duration:.3f}s")
        print(f"Completed at: {end_time}")

        if failed_tests > 0:
            print("\n❌ FAILED TESTS:")
            for result in self.test_results:
                if not result["passed"]:
                    print(f"  • {result['test_name']}: {result['message']}")

        # Module availability report
        print("\n🔧 MODULE AVAILABILITY:")
        print(f"  Timeout Manager: {'✅ Available' if TIMEOUT_MANAGER_AVAILABLE else '❌ Not Available'}")
        print(f"  Git Operations Manager: {'✅ Available' if GIT_MANAGER_AVAILABLE else '❌ Not Available'}")
        print(f"  Config Validator: {'✅ Available' if CONFIG_VALIDATOR_AVAILABLE else '❌ Not Available'}")

        # Performance summary
        if GIT_MANAGER_AVAILABLE:
            try:
                stats = get_git_manager().get_statistics()
                print("\n📈 PERFORMANCE SUMMARY:")
                print(f"  Git Operations: {stats['operations']['total']}")
                print(f"  Cache Hit Rate: {stats['operations']['cache_hit_rate']:.1%}")
                print(f"  Cache Size: {stats['cache']['size']}/{stats['cache']['size_limit']}")
            except Exception:
                pass

        # Cleanup
        try:
            import shutil

            shutil.rmtree(self.temp_dir)
            print(f"\n🧹 Cleaned up test directory: {self.temp_dir}")
        except Exception:
            pass

        print("\n" + "=" * 60)


def main():
    """Main test runner"""
    test_suite = HooksTestSuite()
    test_suite.run_all_tests()


if __name__ == "__main__":
    main()
