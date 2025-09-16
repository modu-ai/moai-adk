#!/usr/bin/env python3
"""
Master Test Runner for MoAI-ADK Critical Testing
Runs all critical tests in proper sequence and provides comprehensive reporting
"""

import sys
import os
import time
import subprocess
from pathlib import Path

# Add project root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

def run_test_suite(test_name, test_file):
    """Run a specific test suite and return results"""
    print(f"\n{'='*60}")
    print(f"🧪 RUNNING: {test_name}")
    print(f"{'='*60}")

    start_time = time.time()

    try:
        result = subprocess.run(
            [sys.executable, str(test_file)],
            capture_output=True,
            text=True,
            timeout=300  # 5 minute timeout
        )

        duration = time.time() - start_time

        print(result.stdout)
        if result.stderr:
            print("STDERR:", result.stderr)

        success = result.returncode == 0

        print(f"\n⏱️  Duration: {duration:.2f} seconds")
        print(f"📊 Result: {'✅ PASSED' if success else '❌ FAILED'}")

        return success, duration, result.stdout, result.stderr

    except subprocess.TimeoutExpired:
        duration = time.time() - start_time
        print(f"⏰ TIMEOUT after {duration:.2f} seconds")
        return False, duration, "", "Test suite timed out"

    except Exception as e:
        duration = time.time() - start_time
        print(f"💥 ERROR: {e}")
        return False, duration, "", str(e)


def main():
    """Run all critical test suites"""
    print("🗿 MoAI-ADK Critical Testing Suite")
    print("=" * 60)
    print("Running comprehensive security and functionality tests...")
    print("This may take several minutes to complete.")

    test_dir = Path(__file__).parent

    # Define test suites in execution order
    test_suites = [
        ("Security Tests", test_dir / "test_security.py"),
        ("Critical Module Tests", test_dir / "test_module_critical.py"),
        ("Integration Tests", test_dir / "test_critical_integration.py"),
        ("Build System Tests", test_dir / "test_build.py"),
        ("Hook Tests", test_dir / "test_hooks.py"),
    ]

    # Track results
    results = []
    total_start_time = time.time()

    # Run each test suite
    for test_name, test_file in test_suites:
        if not test_file.exists():
            print(f"\n⚠️  SKIPPING: {test_name} (file not found: {test_file})")
            results.append((test_name, False, 0, "", f"File not found: {test_file}"))
            continue

        success, duration, stdout, stderr = run_test_suite(test_name, test_file)
        results.append((test_name, success, duration, stdout, stderr))

    total_duration = time.time() - total_start_time

    # Generate comprehensive report
    print(f"\n\n{'='*80}")
    print("🗿 MOAI-ADK CRITICAL TESTING REPORT")
    print(f"{'='*80}")

    passed = sum(1 for _, success, _, _, _ in results if success)
    failed = len(results) - passed

    print(f"📊 Overall Results:")
    print(f"   Total Suites: {len(results)}")
    print(f"   Passed: {passed}")
    print(f"   Failed: {failed}")
    print(f"   Total Time: {total_duration:.2f} seconds")

    print(f"\n📋 Suite Details:")
    for test_name, success, duration, stdout, stderr in results:
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"   {status} {test_name:<25} ({duration:6.2f}s)")

    # Security-specific analysis
    security_tests = [name for name, success, _, _, _ in results if "Security" in name]
    security_passed = [name for name, success, _, _, _ in results if "Security" in name and success]

    print(f"\n🔒 Security Analysis:")
    if security_tests:
        security_ratio = len(security_passed) / len(security_tests)
        print(f"   Security Tests: {len(security_passed)}/{len(security_tests)} passed")
        if security_ratio == 1.0:
            print(f"   🛡️  SECURITY STATUS: ALL PASSED ✅")
        else:
            print(f"   🚨 SECURITY STATUS: FAILURES DETECTED ❌")
    else:
        print(f"   ⚠️  No security tests found")

    # Critical failures analysis
    critical_failures = [(name, stderr) for name, success, _, _, stderr in results if not success]

    if critical_failures:
        print(f"\n🚨 CRITICAL FAILURES DETECTED:")
        for name, error in critical_failures:
            print(f"   💥 {name}:")
            print(f"      Error: {error[:200]}...")

        print(f"\n⚠️  RECOMMENDATION:")
        print(f"   🔧 Fix all critical failures before proceeding")
        print(f"   🔒 Pay special attention to security failures")
        print(f"   📝 Review error logs above for details")

    else:
        print(f"\n🎉 ALL CRITICAL TESTS PASSED!")
        print(f"   ✅ Security systems validated")
        print(f"   ✅ Core modules functional")
        print(f"   ✅ Integration verified")
        print(f"   ✅ Build system operational")

    # Quality gates
    print(f"\n🎯 Quality Gates:")

    # Gate 1: Security
    security_gate = all(success for name, success, _, _, _ in results if "Security" in name)
    print(f"   🔒 Security Gate: {'✅ PASS' if security_gate else '❌ FAIL'}")

    # Gate 2: Core functionality
    core_gate = all(success for name, success, _, _, _ in results if "Module" in name or "Build" in name)
    print(f"   🔧 Core Functionality Gate: {'✅ PASS' if core_gate else '❌ FAIL'}")

    # Gate 3: Integration
    integration_gate = all(success for name, success, _, _, _ in results if "Integration" in name)
    print(f"   🔗 Integration Gate: {'✅ PASS' if integration_gate else '❌ FAIL'}")

    # Overall assessment
    all_gates_passed = security_gate and core_gate and integration_gate

    print(f"\n🎖️  OVERALL ASSESSMENT:")
    if all_gates_passed:
        print(f"   🏆 EXCELLENT - All quality gates passed")
        print(f"   ✅ MoAI-ADK is ready for production use")
        print(f"   🚀 Proceed with confidence")
    else:
        print(f"   ⚠️  ISSUES DETECTED - Quality gates failed")
        print(f"   🔧 Fix failing tests before deployment")
        print(f"   🔒 Security issues must be resolved immediately")

    # Exit code based on critical tests
    critical_success = security_gate and core_gate
    exit_code = 0 if critical_success else 1

    print(f"\n📤 Exit Code: {exit_code}")
    if exit_code == 0:
        print("   ✅ Critical tests passed - Safe to proceed")
    else:
        print("   ❌ Critical tests failed - Fix issues before proceeding")

    return exit_code


if __name__ == '__main__':
    exit_code = main()
    sys.exit(exit_code)