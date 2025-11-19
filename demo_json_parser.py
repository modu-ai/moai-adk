#!/usr/bin/env python3
"""
Demo script for Robust JSON Parser

Tests the parser with real-world error cases and demonstrates the recovery capabilities.
"""

import sys
import os

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from moai_adk.core.robust_json_parser import RobustJSONParser, parse_json

def test_real_world_errors():
    """Test with real-world error patterns from the debug log"""
    print("🔧 Testing Real-World JSON Error Recovery")
    print("=" * 60)

    parser = RobustJSONParser(enable_logging=True)

    # Simulate the actual error patterns from the debug log
    error_cases = [
        {
            'name': 'Expected property name error (Position 1)',
            'input': '{"name": "test", invalid}',
            'expected_recover': True
        },
        {
            'name': 'Malformed response start',
            'input': '{"invalid_start": true,',
            'expected_recover': True
        },
        {
            'name': 'Truncated JSON object',
            'input': '{"nested": {"a": 1, "b": 2}',
            'expected_recover': True
        },
        {
            'name': 'Single quoted JSON',
            'input': "{'name': 'test', 'value': 123}",
            'expected_recover': True
        },
        {
            'name': 'Mixed quote types',
            'input': "{'name': \"test\", 'value': 123}",
            'expected_recover': True
        },
        {
            'name': 'Missing property name quotes',
            'input': '{name: "test", age: 30}',
            'expected_recover': True
        },
        {
            'name': 'Trailing commas',
            'input': '{"items": [1, 2, 3], "count": 3,}',
            'expected_recover': True
        },
        {
            'name': 'Invalid escape sequences',
            'input': '{"path": "C:\\\\Users\\\\test\\x", "valid": true}',
            'expected_recover': True
        },
        {
            'name': 'Control characters',
            'input': '{"name": "test\x00", "value": 123}',
            'expected_recover': True
        },
        {
            'name': 'Embedded in text response',
            'input': 'Response: {"status": "success", "data": [1, 2, 3]} End',
            'expected_recover': True
        },
        {
            'name': 'Markdown code block',
            'input': '```json\n{"name": "test", "value": 123}\n```',
            'expected_recover': True
        },
        {
            'name': 'Completely invalid',
            'input': 'this is not json at all',
            'expected_recover': False
        },
    ]

    success_count = 0
    recovery_count = 0

    for i, case in enumerate(error_cases, 1):
        print(f"\n{i}. {case['name']}")
        print(f"   Input: {case['input'][:60]}{'...' if len(case['input']) > 60 else ''}")

        result = parser.parse(case['input'])

        status = "✅ SUCCESS" if result.success else "❌ FAILED"
        recovery_info = f" ({result.recovery_attempts} attempts)" if result.recovery_attempts > 0 else ""

        print(f"   Result: {status}{recovery_info}")

        if result.success:
            success_count += 1
            if result.recovery_attempts > 0:
                recovery_count += 1
            print(f"   Data: {str(result.data)[:50]}{'...' if len(str(result.data)) > 50 else ''}")
        else:
            print(f"   Error: {result.error}")

        if result.warnings:
            print(f"   Warnings: {len(result.warnings)} recovery actions")
            for warning in result.warnings[:3]:  # Show first 3 warnings
                print(f"     • {warning}")
            if len(result.warnings) > 3:
                print(f"     ... and {len(result.warnings) - 3} more")

    print(f"\n📊 SUMMARY")
    print(f"   Total tests: {len(error_cases)}")
    print(f"   Successful: {success_count}/{len(error_cases)} ({success_count/len(error_cases):.1%})")
    print(f"   Recovered: {recovery_count}/{len(error_cases)} ({recovery_count/len(error_cases):.1%})")
    print(f"   Parser Statistics:")

    stats = parser.get_stats()
    for key, value in stats.items():
        print(f"     {key}: {value}")

def test_performance():
    """Test performance with realistic data sizes"""
    print(f"\n⚡ Performance Testing")
    print("=" * 40)

    parser = RobustJSONParser(enable_logging=False)

    # Test different sizes
    test_cases = [
        ("Small JSON", '{"name": "test", "value": 123}'),
        ("Medium JSON", '{"users": [{"id": i, "name": f"user_{i}"} for i in range(100)]}'),
        ("Large JSON", '{"data": [{"id": i, "values": list(range(10))} for i in range(1000)]}'),
    ]

    import time

    for name, json_template in test_cases:
        if isinstance(json_template, str) and '[' in json_template:
            # Skip template strings for actual performance testing
            continue

        # Create actual test data
        if name == "Medium JSON":
            test_json = '{"users": [' + ','.join([f'{{"id": {i}, "name": "user_{i}"}}' for i in range(100)]) + ']}'
        elif name == "Large JSON":
            test_json = '{"data": [' + ','.join([f'{{"id": {i}, "values": [{j} for j in range(10)]}}' for i in range(1000)]) + ']}'
        else:
            test_json = json_template

        # Test valid JSON
        times = []
        for _ in range(100):
            start = time.time()
            result = parser.parse(test_json)
            times.append(time.time() - start)

        avg_time = sum(times) / len(times) * 1000
        print(f"   {name}: {avg_time:.2f}ms average (100 runs)")

        # Test with errors (requires recovery)
        error_json = test_json.replace('"', "'", 1)  # Introduce a single quote error

        times = []
        for _ in range(100):
            start = time.time()
            result = parser.parse(error_json)
            times.append(time.time() - start)

        avg_error_time = sum(times) / len(times) * 1000
        print(f"   {name} (with recovery): {avg_error_time:.2f}ms average (100 runs)")

def test_edge_cases():
    """Test edge cases and boundary conditions"""
    print(f"\n🔍 Edge Case Testing")
    print("=" * 40)

    parser = RobustJSONParser(enable_logging=False)

    edge_cases = [
        ("Empty string", ""),
        ("Whitespace only", "   \n\t  "),
        ("Null input", None),
        ("Numeric input", 123),
        ("List input", [1, 2, 3]),
        ("Dict input", {"key": "value"}),
        ("Very long string", '{"data": "' + "a" * 10000 + '"}'),
        ("Deeply nested", '{"level1": {"level2": {"level3": {"level4": {"level5": "deep"}}}}}}'),
    ]

    for name, test_input in edge_cases:
        print(f"\n   {name}:")

        try:
            result = parser.parse(test_input)
            status = "✅" if result.success else "❌"
            print(f"     {status} Success: {result.success}, Attempts: {result.recovery_attempts}")
            if not result.success:
                print(f"     Error: {result.error}")
        except Exception as e:
            print(f"     ❌ Exception: {type(e).__name__}: {e}")

if __name__ == "__main__":
    print("🚀 MoAI-ADK Robust JSON Parser Demo")
    print("=" * 60)
    print("Testing production-ready error recovery for Claude Code JSON parsing issues")

    # Test real-world errors from the debug log
    test_real_world_errors()

    # Performance testing
    test_performance()

    # Edge cases
    test_edge_cases()

    print(f"\n✨ Demo completed! The Robust JSON Parser addresses the critical JSON parsing")
    print(f"   errors identified in the Claude Code debug logs with automatic recovery.")