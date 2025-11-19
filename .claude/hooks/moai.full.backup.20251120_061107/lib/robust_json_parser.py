#!/usr/bin/env python3
"""Robust JSON Parser for Stream Processing

Enhanced JSON parsing utility designed to handle:
- Empty strings or null data
- Malformed JSON
- Partial/incomplete JSON chunks
- Non-JSON data (plain text from hooks)
- Stream processing with partial data
- Comprehensive error logging
- Graceful degradation mechanisms

Created to prevent: SyntaxError: Expected property name or '}' in JSON at position 1
"""

import json
import logging
import re
import sys
import traceback
from io import StringIO
from pathlib import Path
from typing import Any, Dict, List, Optional, Union, TextIO

# Configure logging for JSON parsing errors
json_logger = logging.getLogger('robust_json_parser')
json_logger.setLevel(logging.DEBUG)

# Create console handler if not already present
if not json_logger.handlers:
    console_handler = logging.StreamHandler(sys.stderr)
    console_handler.setLevel(logging.ERROR)
    formatter = logging.Formatter('%(name)s - %(levelname)s - %(message)s')
    console_handler.setFormatter(formatter)
    json_logger.addHandler(console_handler)


class RobustJSONParser:
    """Enhanced JSON parser with comprehensive error handling and stream support."""

    def __init__(self, debug_mode: bool = False):
        """Initialize robust JSON parser.

        Args:
            debug_mode: Enable detailed debug logging
        """
        self.debug_mode = debug_mode
        self.error_count = 0
        self.success_count = 0

        # Common patterns that indicate non-JSON data
        self.non_json_patterns = [
            r'^#.*$',  # Comments starting with #
            r'^//.*$',  # Comments starting with //
            r'^[A-Za-z][^{}]*$',  # Plain text without brackets
            r'^\s*$',  # Whitespace only
            r'^[Nn]ull\s*$',  # Various null representations
            r'^undefined\s*$',  # JavaScript undefined
        ]

    def _log_error(self, message: str, exception: Optional[Exception] = None, data_preview: Optional[str] = None) -> None:
        """Log error with detailed information.

        Args:
            message: Error message
            exception: Exception that occurred
            data_preview: Preview of problematic data
        """
        self.error_count += 1
        
        error_info = {
            "message": message,
            "error_count": self.error_count,
        }
        
        if exception:
            error_info["exception_type"] = type(exception).__name__
            error_info["exception_message"] = str(exception)
            
        if data_preview:
            error_info["data_preview"] = data_preview[:200] + "..." if len(data_preview) > 200 else data_preview
            
        if self.debug_mode:
            error_info["traceback"] = traceback.format_exc()
            
        json_logger.error(f"Robust JSON Parser Error: {error_info}")

    def _log_success(self, data_type: str, data_size: int) -> None:
        """Log successful parsing.

        Args:
            data_type: Type of data that was parsed
            data_size: Size of parsed data
        """
        self.success_count += 1
        if self.debug_mode:
            json_logger.debug(f"Successfully parsed {data_type} ({data_size} chars)")

    def _is_non_json_data(self, data: str) -> bool:
        """Check if data matches known non-JSON patterns.

        Args:
            data: Input data to check

        Returns:
            True if data appears to be non-JSON
        """
        data_stripped = data.strip()
        
        # Empty data
        if not data_stripped:
            return True
            
        # Check against patterns
        for pattern in self.non_json_patterns:
            if re.match(pattern, data_stripped):
                return True
                
        return False

    def _repair_json_chunk(self, chunk: str) -> str:
        """Attempt to repair partial JSON chunks.

        Args:
            chunk: Potentially incomplete JSON chunk

        Returns:
            Repaired JSON string or original if repair fails
        """
        chunk_stripped = chunk.strip()
        
        if not chunk_stripped:
            return chunk_stripped
            
        # Common repair patterns
        repairs = [
            # Missing closing braces
            (r'{$', '{}'),
            (r'\[$', '[]'),
            (r'{$', '{"_incomplete": true}'),
            (r'\[$', '{"_incomplete_array": true}'),
            
            # Missing closing quotes
            (r'"([^"]*)$', r'"\1"'),
            
            # Trailing commas (not valid in strict JSON)
            (r',\s*}', '}'),
            (r',\s*\]', ']'),
            
            # Unclosed strings
            (r'"([^"]*)$', r'"\1"'),
        ]
        
        repaired = chunk_stripped
        for pattern, replacement in repairs:
            if re.search(pattern, repaired):
                repaired = re.sub(pattern, replacement, repaired)
                if self.debug_mode:
                    json_logger.debug(f"Applied JSON repair: {pattern} -> {replacement}")
                    
        return repaired

    def _extract_json_from_mixed_data(self, data: str) -> List[str]:
        """Extract JSON objects from mixed text/JSON data.

        Args:
            data: Mixed content that may contain JSON

        Returns:
            List of extracted JSON strings
        """
        json_objects = []
        
        # Look for JSON object patterns
        object_pattern = r'\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}'
        array_pattern = r'\[[^\[\]]*(?:\[[^\[\]]*\][^\[\]]*)*\]'
        
        # Try to extract JSON objects
        for match in re.finditer(object_pattern, data, re.DOTALL):
            json_objects.append(match.group())
            
        # Try to extract JSON arrays
        for match in re.finditer(array_pattern, data, re.DOTALL):
            json_objects.append(match.group())
            
        return json_objects

    def safe_parse_stream_chunk(self, chunk: str, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Safely parse a stream chunk that may be partial or malformed.

        Args:
            chunk: Stream chunk to parse
            default: Default value if parsing fails

        Returns:
            Parsed JSON data or default value
        """
        if not chunk or not chunk.strip():
            self._log_success("empty_data", 0)
            return default if default is not None else {}

        # Check for known non-JSON patterns
        if self._is_non_json_data(chunk):
            self._log_success("non_json_data", len(chunk))
            return {"_type": "text", "content": chunk.strip()}

        # Try direct parsing first
        try:
            result = json.loads(chunk)
            self._log_success("direct_parse", len(chunk))
            return result
        except json.JSONDecodeError as e:
            self._log_error("Direct JSON parse failed", e, chunk)

        # Try to repair the chunk
        try:
            repaired = self._repair_json_chunk(chunk)
            result = json.loads(repaired)
            self._log_success("repaired_parse", len(chunk))
            return {"_repaired": True, "_original_error": str(e), **result}
        except json.JSONDecodeError as e:
            self._log_error("Repaired JSON parse failed", e, chunk)

        # Try to extract JSON from mixed data
        try:
            json_objects = self._extract_json_from_mixed_data(chunk)
            if json_objects:
                # Try to parse the first complete JSON object found
                for json_str in json_objects:
                    try:
                        result = json.loads(json_str)
                        self._log_success("extracted_parse", len(json_str))
                        return {"_extracted": True, "_original_chunk": chunk, **result}
                    except json.JSONDecodeError:
                        continue
        except Exception as e:
            self._log_error("JSON extraction failed", e, chunk)

        # All parsing attempts failed
        self._log_error("All JSON parsing methods failed", data_preview=chunk)
        return default if default is not None else {"_parse_error": True, "_original": chunk}

    def safe_parse_with_fallback(self, data: str, fallback_strategies: List[callable] = None) -> Union[Dict[str, Any], Any]:
        """Parse JSON with multiple fallback strategies.

        Args:
            data: Data to parse
            fallback_strategies: Custom parsing strategies to try

        Returns:
            Parsed data or fallback result
        """
        if not data or not data.strip():
            return {}

        # Default fallback strategies
        if fallback_strategies is None:
            fallback_strategies = [
                # Strategy 1: Direct parsing
                lambda x: json.loads(x),
                
                # Strategy 2: Repair and parse
                lambda x: json.loads(self._repair_json_chunk(x)),
                
                # Strategy 3: Extract JSON from mixed data
                lambda x: {"_extracted": True, "content": self._extract_json_from_mixed_data(x)},
                
                # Strategy 4: Return as text with metadata
                lambda x: {"_type": "text", "content": x.strip()},
            ]

        last_exception = None
        
        for i, strategy in enumerate(fallback_strategies):
            try:
                result = strategy(data)
                self._log_success(f"strategy_{i}", len(data))
                return result
            except Exception as e:
                last_exception = e
                self._log_error(f"Strategy {i} failed", e, data)
                continue

        # All strategies failed
        return {
            "_parse_error": True,
            "_strategies_attempted": len(fallback_strategies),
            "_last_error": str(last_exception) if last_exception else None,
            "_original_data": data[:500] + "..." if len(data) > 500 else data
        }

    def read_and_parse_stdin(self, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Read and parse JSON from stdin with comprehensive error handling.

        Args:
            default: Default value if parsing fails

        Returns:
            Parsed JSON data or default value
        """
        try:
            input_data = sys.stdin.read()
            return self.safe_parse_stream_chunk(input_data, default)
        except Exception as e:
            self._log_error("Failed to read from stdin", e)
            return default if default is not None else {"_stdin_error": True, "_error": str(e)}

    def parse_incremental_data(self, data_stream: TextIO, chunk_size: int = 1024) -> Dict[str, Any]:
        """Parse data incrementally from a stream.

        Args:
            data_stream: Input stream to read from
            chunk_size: Size of chunks to read

        Returns:
            Parsed data or error information
        """
        buffer = ""
        chunks_processed = 0
        
        try:
            while True:
                chunk = data_stream.read(chunk_size)
                if not chunk:
                    break
                    
                buffer += chunk
                chunks_processed += 1
                
                # Try to parse current buffer
                try:
                    result = json.loads(buffer)
                    self._log_success(f"incremental_parse_chunks_{chunks_processed}", len(buffer))
                    return result
                except json.JSONDecodeError:
                    # Buffer might be incomplete, continue reading
                    continue
                    
        except Exception as e:
            self._log_error(f"Incremental parsing failed after {chunks_processed} chunks", e, buffer)
            
        # Final attempt to parse complete buffer
        return self.safe_parse_stream_chunk(buffer, {"_incremental_error": True})

    def validate_and_clean_json(self, data: Union[Dict, str, Any]) -> Dict[str, Any]:
        """Validate and clean JSON data.

        Args:
            data: Data to validate and clean

        Returns:
            Cleaned and validated JSON data
        """
        if isinstance(data, str):
            parsed = self.safe_parse_stream_chunk(data)
        else:
            parsed = data if isinstance(data, dict) else {"value": data}
            
        # Add metadata
        parsed["_parser_stats"] = {
            "errors": self.error_count,
            "successes": self.success_count,
            "timestamp": json.dumps({"timestamp": "auto"})  # Will be replaced with actual timestamp
        }
        
        return parsed

    def get_parser_stats(self) -> Dict[str, Any]:
        """Get parser statistics.

        Returns:
            Dictionary with parser statistics
        """
        return {
            "error_count": self.error_count,
            "success_count": self.success_count,
            "success_rate": self.success_count / max(1, self.error_count + self.success_count),
            "debug_mode": self.debug_mode
        }


# Convenience functions for backward compatibility and ease of use
def safe_parse_json(data: str, default: Optional[Any] = None, debug: bool = False) -> Union[Dict[str, Any], Any]:
    """Safely parse JSON string with comprehensive error handling.

    Args:
        data: JSON string to parse
        default: Default value if parsing fails
        debug: Enable debug logging

    Returns:
        Parsed JSON data or default value
    """
    parser = RobustJSONParser(debug_mode=debug)
    return parser.safe_parse_stream_chunk(data, default)


def safe_read_stdin_json(default: Optional[Any] = None, debug: bool = False) -> Union[Dict[str, Any], Any]:
    """Safely read and parse JSON from stdin.

    Args:
        default: Default value if parsing fails
        debug: Enable debug logging

    Returns:
        Parsed JSON data or default value
    """
    parser = RobustJSONParser(debug_mode=debug)
    return parser.read_and_parse_stdin(default)


def create_error_response(error_message: str, original_data: Optional[str] = None) -> Dict[str, Any]:
    """Create standardized error response for JSON parsing failures.

    Args:
        error_message: Description of the error
        original_data: Original data that failed to parse (truncated)

    Returns:
        Standardized error response
    """
    response = {
        "success": False,
        "error": error_message,
        "error_type": "JSON_PARSE_ERROR",
        "_robust_parser": True
    }
    
    if original_data:
        response["original_data_preview"] = original_data[:200] + "..." if len(original_data) > 200 else original_data
        
    return response


# Main execution for testing
if __name__ == "__main__":
    # Test the robust parser with various inputs
    parser = RobustJSONParser(debug_mode=True)
    
    test_cases = [
        '{"valid": "json"}',
        '',
        '   ',
        '{"incomplete": ',
        'plain text output',
        '{"trailing_comma": [1, 2, 3,],}',
        '{"nested": {"incomplete": ',
        'Status: Operation completed successfully',
    ]
    
    for i, test_case in enumerate(test_cases):
        print(f"\nTest Case {i + 1}: {repr(test_case)}")
        result = parser.safe_parse_stream_chunk(test_case, {"fallback": "value"})
        print(f"Result: {result}")
        
    print(f"\nParser Stats: {parser.get_parser_stats()}")
