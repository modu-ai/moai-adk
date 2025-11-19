#!/usr/bin/env python3
"""Example Hook with Robust JSON Parsing

This example demonstrates how to update existing hooks to use the robust JSON parsing
utilities. It shows the before/after patterns and proper error handling.

This hook simulates a typical hook that processes tool input and can receive
malformed JSON data that would cause the original SyntaxError.
"""

import json
import sys
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

# Import robust JSON parsing utilities
try:
    from json_utils_enhanced import parse_hook_input_robust, create_hook_response, EnhancedJSONUtils
    ROBUST_PARSING_AVAILABLE = True
except ImportError:
    # Fallback to original parsing if enhanced utilities not available
    ROBUST_PARSING_AVAILABLE = False
    print("Warning: Robust JSON parsing not available, using fallback", file=sys.stderr)


# ============================================================================
# BEFORE: Vulnerable JSON Parsing Pattern
# ============================================================================

def process_hook_input_vulnerable():
    """Example of vulnerable JSON parsing that can cause SyntaxError."""
    try:
        # VULNERABLE: Direct JSON parsing without error handling
        input_data = sys.stdin.read()
        data = json.loads(input_data) if input_data.strip() else {}
        
        # Process the data
        if "tool" in data:
            result = {"processed_tool": data["tool"], "status": "success"}
            print(json.dumps(result))
            return True
        else:
            error_msg = "No tool specified"
            print(json.dumps({"error": error_msg}))
            return False
            
    except json.JSONDecodeError as e:
        # This error handling is minimal and doesn't prevent all issues
        error_response = {"error": f"JSON decode error: {e}"}
        print(json.dumps(error_response))
        sys.exit(1)  # Hard exit - can interrupt operations
        
    except Exception as e:
        error_response = {"error": f"Unexpected error: {e}"}
        print(json.dumps(error_response))
        sys.exit(1)


# ============================================================================
# AFTER: Robust JSON Parsing Pattern  
# ============================================================================

def process_hook_input_robust():
    """Example of robust JSON parsing that prevents SyntaxError."""
    try:
        # ROBUST: Use enhanced parsing with comprehensive error handling
        if ROBUST_PARSING_AVAILABLE:
            # Method 1: Parse from stdin automatically
            data = parse_hook_input_robust()
            
            # Method 2: Parse specific string (if reading from other sources)
            # input_data = sys.stdin.read()  
            # data = parse_hook_input_robust(input_data)
            
        else:
            # Fallback to safer parsing if enhanced utilities not available
            input_data = sys.stdin.read()
            try:
                data = json.loads(input_data) if input_data.strip() else {}
            except json.JSONDecodeError:
                data = {"_parse_error": True, "original": input_data}
        
        # Check for parsing errors
        if isinstance(data, dict) and data.get("_parse_error"):
            # Create standardized error response that allows operation to continue
            response = create_hook_response(
                success=False,
                error="Failed to parse hook input JSON",
                data={"original_preview": str(data.get("_original", ""))[:100]},
                continue_operation=True  # Allow operation to continue
            )
            print(json.dumps(response))
            return True  # Return True to indicate graceful handling
            
        # Process the data successfully
        if "tool" in data:
            response = create_hook_response(
                success=True,
                message=f"Successfully processed tool: {data['tool']}",
                data={
                    "processed_tool": data["tool"],
                    "input_fields": list(data.keys()),
                    "robust_parsing": ROBUST_PARSING_AVAILABLE
                },
                continue_operation=True
            )
        else:
            response = create_hook_response(
                success=False,
                error="No tool specified in hook input",
                data={"received_fields": list(data.keys()) if isinstance(data, dict) else []},
                continue_operation=True  # Still allow operation to continue
            )
            
        print(json.dumps(response))
        return True
        
    except Exception as e:
        # Comprehensive error handling that maintains system stability
        response = create_hook_response(
            success=False,
            error=f"Unexpected error in hook processing: {e}",
            data={"error_type": type(e).__name__},
            continue_operation=True  # Always allow operation to continue
        )
        print(json.dumps(response))
        return True  # Return True to indicate graceful error handling


# ============================================================================
# ADVANCED: Stream Processing with Robust Parser
# ============================================================================

def process_streaming_input():
    """Example of processing streaming data with robust parsing."""
    if not ROBUST_PARSING_AVAILABLE:
        print(json.dumps({"error": "Robust parsing not available"}))
        return False
        
    # Initialize enhanced utilities with debug mode
    utils = EnhancedJSONUtils(debug_mode=True)
    
    try:
        # Simulate receiving streaming data chunks
        # In real usage, this would come from a stream or socket
        test_chunks = [
            '{"partial": "',
            'data", "status": ',
            '"complete"}',
            'Some plain text output',  # Non-JSON chunk
            '',  # Empty chunk
            '{"final": "chunk"}'
        ]
        
        results = []
        
        for chunk in test_chunks:
            # Parse each chunk robustly
            parsed_chunk = utils.safe_json_loads_robust(chunk, {"chunk_type": "unparseable"})
            results.append(parsed_chunk)
            
        # Create summary response
        response = create_hook_response(
            success=True,
            message="Processed streaming data successfully",
            data={
                "chunks_processed": len(results),
                "parser_stats": utils.get_parser_summary(),
                "results": results
            },
            continue_operation=True
        )
        
        print(json.dumps(response))
        return True
        
    except Exception as e:
        response = create_hook_response(
            success=False,
            error=f"Stream processing error: {e}",
            continue_operation=True
        )
        print(json.dumps(response))
        return False


# ============================================================================
# DEMONSTRATION: Testing with Problematic Inputs
# ============================================================================

def demonstrate_robust_parsing():
    """Demonstrate how robust parsing handles problematic inputs."""
    if not ROBUST_PARSING_AVAILABLE:
        print(json.dumps({"error": "Robust parsing not available for demonstration"}))
        return
        
    # Test cases that would cause the original SyntaxError
    problematic_inputs = [
        "",  # Empty input - common in hooks
        "   ",  # Whitespace only
        "\n\n",  # Newlines only  
        '{"incomplete": ',  # Truncated JSON
        'Plain text from hook output',  # Non-JSON data
        '{"trailing_comma": [1,2,3,],}',  # Invalid JSON syntax
        '{"tool": "Edit", "file": '  # Incomplete but useful data
    ]
    
    print("=== Demonstrating Robust JSON Parsing ===", file=sys.stderr)
    
    utils = EnhancedJSONUtils(debug_mode=False)  # Reduce debug output for demo
    
    for i, test_input in enumerate(problematic_inputs):
        print(f"\nTest {i+1}: {repr(test_input)}", file=sys.stderr)
        
        # Test robust parsing
        result = utils.safe_json_loads_robust(test_input, {"fallback": "safe_default"})
        
        print(f"Result type: {type(result).__name__}", file=sys.stderr)
        print(f"Result: {result}", file=sys.stderr)
        
        # Create appropriate response
        if isinstance(result, dict) and "_parse_error" in result:
            response = create_hook_response(
                success=False,
                error="JSON parsing failed but operation continues",
                data={"parsing_result": result},
                continue_operation=True
            )
        else:
            response = create_hook_response(
                success=True,
                message="Successfully parsed potentially problematic input",
                data={"parsed_data": result},
                continue_operation=True
            )
            
        # Would output this in real hook execution
        # print(json.dumps(response))
    
    # Print summary
    print(f"\nParser Summary: {utils.get_parser_summary()}", file=sys.stderr)


def main():
    """Main entry point demonstrating both vulnerable and robust patterns."""
    # Check if we should run demonstration
    if len(sys.argv) > 1 and sys.argv[1] == "--demo":
        demonstrate_robust_parsing()
        return
        
    if len(sys.argv) > 1 and sys.argv[1] == "--stream":
        process_streaming_input()
        return
        
    if len(sys.argv) > 1 and sys.argv[1] == "--vulnerable":
        print("=== Running VULNERABLE Pattern ===", file=sys.stderr)
        process_hook_input_vulnerable()
        return
        
    # Default: Run robust pattern
    print("=== Running ROBUST Pattern ===", file=sys.stderr)
    process_hook_input_robust()


if __name__ == "__main__":
    main()
