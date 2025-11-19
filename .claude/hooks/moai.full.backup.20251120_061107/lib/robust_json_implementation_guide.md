# Robust JSON Parser Implementation Guide

## Overview

This guide explains how to implement robust JSON parsing to prevent the original error:
`SyntaxError: Expected property name or '}' in JSON at position 1 (line 1 column 2)`

## Problem Analysis

The original error occurs when stream processing receives:
- Empty or null data
- Malformed JSON
- Partial/incomplete JSON chunks  
- Non-JSON data (plain text from hooks)
- Truncated data streams

## Solution Components

### 1. Core Parser: `robust_json_parser.py`

Main features:
- **Safe parsing with fallbacks**: Multiple parsing strategies
- **Stream chunk processing**: Handles partial data gracefully
- **Error recovery**: Automatic repair attempts for common issues
- **Comprehensive logging**: Detailed error tracking and debugging
- **Non-JSON detection**: Identifies and handles plain text appropriately

Key classes:
```python
RobustJSONParser(debug_mode=False)
# Main parser with comprehensive error handling
```

Key methods:
```python
safe_parse_stream_chunk(chunk, default=None)  # Parse individual chunks
safe_parse_with_fallback(data, strategies)    # Custom fallback strategies
parse_incremental_data(stream, chunk_size)    # Process streams incrementally
```

### 2. Enhanced Utilities: `json_utils_enhanced.py`

Extends original `JSONUtils` with backward compatibility:
- **EnhancedJSONUtils**: Enhanced version of original utilities
- **Backward compatibility**: All original methods still work
- **Hook-specific methods**: Optimized for hook input/output patterns
- **Standardized responses**: Consistent error handling across hooks

Key methods:
```python
read_json_from_stdin_robust(default=None)      # Safe stdin reading
parse_hook_input_robust(input_data=None)       # Hook-specific parsing
create_hook_response(success=True, ...)        # Standardized responses
```

## Implementation Patterns

### Pattern 1: Basic Hook Input Parsing

Replace vulnerable:
```python
# VULNERABLE - Can cause SyntaxError
input_data = sys.stdin.read()
data = json.loads(input_data) if input_data.strip() else {}
```

With robust:
```python
# ROBUST - Handles all edge cases
from lib.json_utils_enhanced import parse_hook_input_robust

data = parse_hook_input_robust()  # Reads from stdin automatically
# OR
data = parse_hook_input_robust(input_string)  # Parse specific string
```

### Pattern 2: Hook Output Response

Replace inconsistent error handling:
```python
# INCONSISTENT
try:
    result = process_data(data)
    print(json.dumps(result))
except json.JSONDecodeError as e:
    error_response = {"error": str(e)}
    print(json.dumps(error_response))
```

With standardized:
```python
# STANDARDIZED
from lib.json_utils_enhanced import create_hook_response

try:
    result = process_data(data)
    response = create_hook_response(
        success=True,
        message="Operation completed",
        data=result,
        continue_operation=True
    )
    print(json.dumps(response))
except Exception as e:
    response = create_hook_response(
        success=False,
        error=str(e),
        continue_operation=True  # Allow operation to continue
    )
    print(json.dumps(response))
```

### Pattern 3: Stream Processing

For handling streams or chunked data:
```python
from lib.robust_json_parser import RobustJSONParser

parser = RobustJSONParser(debug_mode=True)

# Process incoming stream chunks
for chunk in stream:
    parsed_chunk = parser.safe_parse_stream_chunk(chunk, {"status": "partial"})
    
    if "_parse_error" in parsed_chunk:
        # Handle error gracefully
        log_error(parsed_chunk)
        continue
        
    # Process valid data
    process_chunk(parsed_chunk)

# Get parser statistics
stats = parser.get_parser_stats()
print(f"Processed {stats['success_count']} chunks, {stats['error_count']} errors")
```

### Pattern 4: Backward Compatibility

For existing code that uses `JSONUtils`:
```python
# Original - Still works
from lib.json_utils import JSONUtils
utils = JSONUtils()
data = utils.safe_json_loads(json_string)

# Enhanced - Add robustness with minimal changes
from lib.json_utils_enhanced import EnhancedJSONUtils
utils = EnhancedJSONUtils(debug_mode=True)
data = utils.safe_json_loads_robust(json_string)  # Same interface, more robust
```

## Migration Guide

### Step 1: Update Hook Input Parsing

Find vulnerable patterns in your hooks:
```bash
grep -r "json.loads.*input_data" .claude/hooks/moai/ --include="*.py"
```

Replace with robust parsing:
```python
# Before
input_data = sys.stdin.read()
data = json.loads(input_data) if input_data.strip() else {}

# After  
from lib.json_utils_enhanced import parse_hook_input_robust
data = parse_hook_input_robust()
```

### Step 2: Standardize Error Responses

Update error handling to use consistent format:
```python
# Before
except json.JSONDecodeError as e:
    print(json.dumps({"error": str(e)}))
    sys.exit(1)

# After
from lib.json_utils_enhanced import create_hook_response
except Exception as e:
    response = create_hook_response(
        success=False,
        error=f"Failed to process input: {e}",
        continue_operation=True
    )
    print(json.dumps(response))
    sys.exit(0)  # Exit with 0 to allow operation to continue
```

### Step 3: Add Debug Logging (Optional)

Enable debug mode for troubleshooting:
```python
# Enable debug logging for development/debugging
parser = RobustJSONParser(debug_mode=True)
# OR
utils = EnhancedJSONUtils(debug_mode=True)
```

## Error Recovery Strategies

The robust parser implements multiple fallback strategies:

1. **Direct parsing**: Try standard `json.loads()` first
2. **JSON repair**: Fix common issues (missing braces, trailing commas)
3. **Extraction**: Pull JSON from mixed text/JSON data
4. **Text fallback**: Return as text content with metadata
5. **Custom fallback**: User-defined recovery logic

## Testing

Test your implementation:
```bash
# Test the robust parser
cd .claude/hooks/moai/lib
python3 robust_json_parser.py

# Test with problematic inputs
echo '{"incomplete": ' | python3 -c "
from lib.json_utils_enhanced import parse_hook_input_robust
import sys
result = parse_hook_input_robust(sys.stdin.read())
print('Result:', result)
"
```

## Benefits

1. **Prevents crashes**: No more SyntaxError from malformed JSON
2. **Graceful degradation**: Operations continue even with bad data
3. **Better debugging**: Detailed error logging and statistics
4. **Backward compatibility**: Existing code continues to work
5. **Standardized responses**: Consistent error handling across hooks
6. **Stream support**: Handles partial and chunked data correctly

## Files Created

- `robust_json_parser.py`: Core robust parsing implementation
- `json_utils_enhanced.py`: Enhanced utilities with backward compatibility
- `test_robust_json.py`: Comprehensive test suite
- `robust_json_implementation_guide.md`: This implementation guide

## Usage Statistics

The parser tracks:
- Success/error counts
- Success rate percentage
- Detailed error information
- Processing statistics

Access via:
```python
parser = RobustJSONParser()
stats = parser.get_parser_stats()
print(f"Success rate: {stats['success_rate']:.1%}")
```
