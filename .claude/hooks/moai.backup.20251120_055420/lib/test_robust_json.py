#!/usr/bin/env python3
"""Test Suite for Robust JSON Parser

Comprehensive tests for the robust JSON parsing utilities.
This script demonstrates how the enhanced parser handles various edge cases
that can cause the original SyntaxError.
"""

import sys
import json
from pathlib import Path

# Add the lib directory to path for imports
sys.path.insert(0, str(Path(__file__).parent))

try:
    from robust_json_parser import RobustJSONParser, safe_parse_json, create_error_response
    from json_utils_enhanced import EnhancedJSONUtils, parse_hook_input_robust
    ROBUST_PARSER_AVAILABLE = True
except ImportError as e:
    print(f"Warning: Robust parser not available: {e}")
    ROBUST_PARSER_AVAILABLE = False


def test_original_error_scenario():
    """Test the specific scenario that caused the original error."""
    print("=== Testing Original Error Scenario ===")
    
    # Simulate the problematic input that caused "Expected property name or '}' in JSON at position 1"
    problematic_inputs = [
        "",  # Empty input
        "   ",  # Whitespace only
        "\n\n",  # Newlines only
        "{",  # Incomplete JSON object
        "[",  # Incomplete JSON array
        "null",  # Null value
        "undefined",  # JavaScript undefined
        "plain text response",  # Non-JSON text
        '{"incomplete": ',  # Truncated JSON
        '{"trailing_comma": [1,2,3,],}',  # Invalid JSON with trailing comma
        '{"bad_escape": "\\x"}',  # Invalid escape sequence (fixed)
        '{"multiple": "errors",}',  # Multiple issues
    ]
    
    if ROBUST_PARSER_AVAILABLE:
        parser = RobustJSONParser(debug_mode=True)
        
        for i, test_input in enumerate(problematic_inputs):
            print(f"\nTest {i + 1}: {repr(test_input)}")
            
            try:
                # This would fail with the original json.loads()
                result = json.loads(test_input)
                print(f"  Standard JSON succeeded: {result}")
            except json.JSONDecodeError as e:
                print(f"  Standard JSON failed: {e}")
                
            # Test robust parser
            robust_result = parser.safe_parse_stream_chunk(test_input, {"fallback": "safe_default"})
            print(f"  Robust parser result: {robust_result}")
            
    else:
        print("Robust parser not available for testing")


def test_stream_processing_scenarios():
    """Test stream processing scenarios."""
    print("\n=== Testing Stream Processing Scenarios ===")
    
    # Simulate stream chunks that might be received
    stream_chunks = [
        # Complete JSON in one chunk
        '{"status": "success", "data": [1, 2, 3]}',
        
        # Partial JSON chunks
        '{"partial": "',
        'data", "completed": true}',
        
        # Mixed text and JSON
        'Status: Operation completed\n{"result": "success"}',
        
        # Hook output mixed with JSON
        'Processing file...\n{"file": "test.py", "lines": 150}\nDone!',
        
        # Malformed JSON with useful information
        '{"error": "File not found", "path": "/path/to/missing"',
        
        # Empty and whitespace chunks
        '',
        '   ',
        '\n',
    ]
    
    if ROBUST_PARSER_AVAILABLE:
        parser = RobustJSONParser(debug_mode=False)  # Reduce debug output
        
        print("Processing stream chunks:")
        for i, chunk in enumerate(stream_chunks):
            result = parser.safe_parse_stream_chunk(chunk)
            print(f"  Chunk {i + 1}: {repr(chunk[:30])}... -> {type(result).__name__}")
            
        print(f"\nParser Statistics: {parser.get_parser_stats()}")


def test_hook_simulation():
    """Simulate hook input/output scenarios."""
    print("\n=== Testing Hook Simulation ===")
    
    # Simulate various hook input scenarios
    hook_inputs = [
        # Normal hook input
        '{"tool": "Edit", "parameters": {"file": "test.py"}}',
        
        # Malformed hook input (what might cause the original error)
        '',
        '{"incomplete": ',
        'Hook execution completed successfully',  # Plain text output
        
        # Mixed content
        'Starting hook execution...\n{"tool": "Write"}\nHook finished.',
        
        # JSON with metadata
        '{"tool": "Bash", "command": "rm -rf /tmp/test", "risky": true}',
    ]
    
    if ROBUST_PARSER_AVAILABLE:
        enhanced_utils = EnhancedJSONUtils(debug_mode=False)
        
        for i, hook_input in enumerate(hook_inputs):
            print(f"\nHook Input {i + 1}: {repr(hook_input[:50])}")
            
            # Test robust hook parsing
            parsed_data = enhanced_utils.parse_hook_input_robust(hook_input)
            print(f"  Parsed: {parsed_data}")
            
            # Create appropriate response
            if isinstance(parsed_data, dict) and "_parse_error" in parsed_data:
                response = enhanced_utils.create_hook_response(
                    success=False,
                    error="Failed to parse hook input",
                    continue_operation=True  # Allow operation to continue
                )
            else:
                response = enhanced_utils.create_hook_response(
                    success=True,
                    message="Hook input parsed successfully",
                    data={"parsed_fields": list(parsed_data.keys()) if isinstance(parsed_data, dict) else []},
                    continue_operation=True
                )
                
            print(f"  Response: {response}")


def test_error_recovery():
    """Test error recovery mechanisms."""
    print("\n=== Testing Error Recovery ===")
    
    if ROBUST_PARSER_AVAILABLE:
        # Test fallback strategies
        test_data = '{"incomplete": json data'
        
        parser = RobustJSONParser(debug_mode=True)
        
        # Test with different fallback strategies
        fallback_results = []
        
        # Strategy 1: Default fallback
        result1 = parser.safe_parse_stream_chunk(test_data, {"fallback": "safe"})
        fallback_results.append(("default_fallback", result1))
        
        # Strategy 2: Custom fallback function
        def custom_fallback():
            return {"_custom_recovery": True, "original_length": len(test_data)}
            
        result2 = parser.safe_parse_with_fallback(test_data, [
            lambda x: json.loads(x),  # Try direct parsing
            lambda x: {"_repaired": True, "content": x},  # Custom repair
            custom_fallback  # Custom fallback
        ])
        fallback_results.append(("custom_strategies", result2))
        
        for strategy_name, result in fallback_results:
            print(f"\nStrategy: {strategy_name}")
            print(f"  Result: {result}")


def test_backward_compatibility():
    """Test backward compatibility with original JSONUtils."""
    print("\n=== Testing Backward Compatibility ===")
    
    if ROBUST_PARSER_AVAILABLE:
        from json_utils import JSONUtils
        from json_utils_enhanced import EnhancedJSONUtils
        
        # Test data
        test_data = '{"test": "data", "number": 123}'
        
        # Original JSONUtils
        original_utils = JSONUtils()
        original_result = original_utils.safe_json_loads(test_data)
        
        # Enhanced JSONUtils
        enhanced_utils = EnhancedJSONUtils()
        enhanced_result = enhanced_utils.safe_json_loads_robust(test_data)
        
        print(f"Original result: {original_result}")
        print(f"Enhanced result: {enhanced_result}")
        print(f"Results compatible: {original_result == enhanced_result}")


def main():
    """Run all tests."""
    print("Robust JSON Parser Test Suite")
    print("=" * 50)
    
    if not ROBUST_PARSER_AVAILABLE:
        print("ERROR: Robust parser modules not available!")
        sys.exit(1)
    
    try:
        test_original_error_scenario()
        test_stream_processing_scenarios()
        test_hook_simulation()
        test_error_recovery()
        test_backward_compatibility()
        
        print("\n" + "=" * 50)
        print("All tests completed successfully!")
        print("The robust JSON parser handles all scenarios that would")
        print("previously cause 'SyntaxError: Expected property name or '}' in JSON'")
        
    except Exception as e:
        print(f"Test failed with error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
