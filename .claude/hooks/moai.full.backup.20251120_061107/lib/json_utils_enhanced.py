#!/usr/bin/env python3
"""Enhanced JSON Utilities for Alfred Hooks with Robust Stream Processing

This module extends the original json_utils.py with comprehensive error handling
for stream processing, partial data, and malformed JSON.

Key Features:
- Backward compatibility with original JSONUtils
- Robust stream parsing with error recovery
- Comprehensive logging and debugging
- Graceful degradation mechanisms
- Support for mixed text/JSON data

Created to prevent: SyntaxError: Expected property name or '}' in JSON at position 1
"""

import json
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Union

# Import the robust parser
try:
    from .robust_json_parser import RobustJSONParser, safe_parse_json, safe_read_stdin_json, create_error_response
except ImportError:
    # Fallback if robust parser is not available
    RobustJSONParser = None
    safe_parse_json = None
    safe_read_stdin_json = None
    create_error_response = None

from .json_utils import JSONUtils


class EnhancedJSONUtils(JSONUtils):
    """Enhanced JSON utility class with robust parsing capabilities."""

    def __init__(self, debug_mode: bool = False):
        """Initialize enhanced JSON utilities.

        Args:
            debug_mode: Enable detailed debug logging
        """
        super().__init__()
        self.debug_mode = debug_mode
        self.robust_parser = RobustJSONParser(debug_mode=debug_mode) if RobustJSONParser else None

    def read_json_from_stdin_robust(self, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Read and parse JSON from stdin with robust error handling.

        Args:
            default: Default value if parsing fails

        Returns:
            Parsed JSON data as dictionary or default value
        """
        if self.robust_parser:
            return self.robust_parser.read_and_parse_stdin(default)
        else:
            # Fallback to original method
            try:
                return self.read_json_from_stdin()
            except json.JSONDecodeError as e:
                if default is not None:
                    return default
                # Return error information
                return {"_parse_error": True, "error": str(e)}

    def safe_json_loads_robust(self, json_str: str, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Safely parse JSON string with enhanced robustness.

        Args:
            json_str: JSON string to parse
            default: Default value if parsing fails

        Returns:
            Parsed JSON data or default value
        """
        if self.robust_parser:
            return self.robust_parser.safe_parse_stream_chunk(json_str, default)
        else:
            # Fallback to original method
            return self.safe_json_loads(json_str, default)

    def safe_json_load_file_robust(self, file_path: Path, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
        """Safely load JSON from file with enhanced error handling.

        Args:
            file_path: Path to JSON file
            default: Default value if file doesn't exist or parsing fails

        Returns:
            Parsed JSON data or default value
        """
        try:
            if file_path.exists():
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    return self.safe_json_loads_robust(content, default)
            else:
                return default if default is not None else {}
        except Exception as e:
            if self.debug_mode:
                print(f"Enhanced JSON file read error: {e}", file=sys.stderr)
            return default if default is not None else {"_file_error": True, "error": str(e)}

    def parse_hook_input_robust(self, input_data: Optional[str] = None) -> Dict[str, Any]:
        """Parse hook input with comprehensive error handling.

        This is the recommended method for parsing hook stdin input.
        It handles all the edge cases that can occur in hook execution.

        Args:
            input_data: Input data string (if None, reads from stdin)

        Returns:
            Parsed JSON data with metadata
        """
        if input_data is None:
            # Read from stdin
            return self.read_json_from_stdin_robust({})
        else:
            # Parse provided string
            return self.safe_json_loads_robust(input_data, {})

    def create_hook_response(
        self,
        success: bool = True,
        message: Optional[str] = None,
        error: Optional[str] = None,
        data: Optional[Dict[str, Any]] = None,
        continue_operation: bool = True
    ) -> Dict[str, Any]:
        """Create standardized hook response with robust metadata.

        Args:
            success: Whether operation was successful
            message: Descriptive message
            error: Error message if failed
            data: Additional data payload
            continue_operation: Whether the operation should continue

        Returns:
            Standardized hook response dictionary
        """
        response = self.create_standard_response(success, message, error, data)
        
        # Add hook-specific fields
        response["continue"] = continue_operation
        response["_enhanced_parser"] = True
        
        if self.robust_parser:
            response["_parser_stats"] = self.robust_parser.get_parser_stats()
            
        return response

    def validate_hook_data_structure(self, data: Dict[str, Any], required_fields: List[str]) -> Dict[str, Any]:
        """Validate hook data structure with enhanced error reporting.

        Args:
            data: Hook data to validate
            required_fields: List of required field names

        Returns:
            Validation result with detailed information
        """
        validation_result = {
            "valid": False,
            "missing_fields": [],
            "extra_fields": [],
            "field_types": {},
            "_validation_timestamp": None
        }
        
        try:
            # Basic structure validation
            if not isinstance(data, dict):
                validation_result["error"] = "Input data is not a dictionary"
                return validation_result
                
            # Check required fields
            missing = [field for field in required_fields if field not in data]
            validation_result["missing_fields"] = missing
            
            # Check field types
            for field, value in data.items():
                validation_result["field_types"][field] = type(value).__name__
                
            validation_result["valid"] = len(missing) == 0
            validation_result["_validation_timestamp"] = json.dumps({"timestamp": "auto"})
            
        except Exception as e:
            validation_result["validation_error"] = str(e)
            if self.debug_mode:
                print(f"Hook validation error: {e}", file=sys.stderr)
                
        return validation_result

    def process_stream_data(self, data_stream, chunk_size: int = 1024) -> Dict[str, Any]:
        """Process streaming data with incremental parsing.

        Args:
            data_stream: Input stream to read from
            chunk_size: Size of chunks to read

        Returns:
            Parsed data or error information
        """
        if self.robust_parser:
            return self.robust_parser.parse_incremental_data(data_stream, chunk_size)
        else:
            # Fallback: read all data at once
            try:
                content = data_stream.read()
                return self.safe_json_loads_robust(content, {"_stream_fallback": True})
            except Exception as e:
                return {"_stream_error": True, "error": str(e)}

    def get_parser_summary(self) -> Dict[str, Any]:
        """Get comprehensive parser statistics and status.

        Returns:
            Parser summary with statistics
        """
        summary = {
            "enhanced_mode": True,
            "robust_parser_available": self.robust_parser is not None,
            "debug_mode": self.debug_mode,
        }
        
        if self.robust_parser:
            summary.update(self.robust_parser.get_parser_stats())
            
        return summary


# Global instance for backward compatibility
_enhanced_utils = EnhancedJSONUtils()

# Export enhanced functions for backward compatibility
def read_json_from_stdin_robust(default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
    """Read JSON from stdin with robust error handling."""
    return _enhanced_utils.read_json_from_stdin_robust(default)

def safe_json_loads_robust(json_str: str, default: Optional[Any] = None) -> Union[Dict[str, Any], Any]:
    """Safely parse JSON string with enhanced robustness."""
    return _enhanced_utils.safe_json_loads_robust(json_str, default)

def parse_hook_input_robust(input_data: Optional[str] = None) -> Dict[str, Any]:
    """Parse hook input with comprehensive error handling."""
    return _enhanced_utils.parse_hook_input_robust(input_data)

def create_hook_response(
    success: bool = True,
    message: Optional[str] = None,
    error: Optional[str] = None,
    data: Optional[Dict[str, Any]] = None,
    continue_operation: bool = True
) -> Dict[str, Any]:
    """Create standardized hook response."""
    return _enhanced_utils.create_hook_response(success, message, error, data, continue_operation)


# Backward compatibility - maintain original exports
__all__ = [
    # Original exports
    'JSONUtils',
    'read_json_from_stdin_robust',
    'safe_json_loads_robust', 
    'parse_hook_input_robust',
    'create_hook_response',
    'EnhancedJSONUtils',
    
    # Enhanced exports
    'safe_parse_json',
    'safe_read_stdin_json',
    'create_error_response',
]
