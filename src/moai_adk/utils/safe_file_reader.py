#!/usr/bin/env python3
"""
Safe File Reader Utility

Provides safe file reading with multiple encoding fallbacks and error handling.
Resolves UTF-8 encoding issues found in MoAI-ADK project.

Author: Alfred@MoAI
Date: 2025-11-11
"""

from pathlib import Path
from typing import List, Optional, Union


class SafeFileReader:
    """
    Safe file reader with encoding fallback support.

    Handles various encoding issues including:
    - UTF-8 vs CP1252 encoding conflicts
    - Binary files with special characters
    - File system encoding issues
    """

    # Encoding priority order (most common to least common)
    DEFAULT_ENCODINGS = [
        "utf-8",  # Standard UTF-8
        "cp1252",  # Windows-1252 (Western European)
        "iso-8859-1",  # Latin-1 (Western European)
        "latin1",  # Alternative Latin-1
        "utf-16",  # UTF-16 with BOM detection
        "ascii",  # Pure ASCII fallback
    ]

    def __init__(self, encodings: Optional[List[str]] = None, errors: str = "ignore"):
        """
        Initialize SafeFileReader.

        Args:
            encodings: List of encodings to try in order
            errors: Error handling strategy ('ignore', 'replace', 'strict')
        """
        self.encodings = encodings or self.DEFAULT_ENCODINGS
        self.errors = errors

    def read_text(self, file_path: Union[str, Path]) -> Optional[str]:
        """
        Safely read text file with encoding fallbacks.

        Args:
            file_path: Path to the file to read

        Returns:
            File content as string, or None if all attempts fail
        """
        file_path = Path(file_path)

        if not file_path.exists():
            return None

        # Try each encoding in order
        for encoding in self.encodings:
            try:
                return file_path.read_text(encoding=encoding)
            except UnicodeDecodeError:
                continue
            except Exception as e:
                # Log non-decoding errors but continue
                print(f"Warning: Error reading {file_path} with {encoding}: {e}")
                continue

        # Final fallback with specified error handling
        try:
            return file_path.read_text(encoding="utf-8", errors=self.errors)
        except Exception as e:
            print(f"Error: Could not read {file_path}: {e}")
            return None

    def read_lines(self, file_path: Union[str, Path]) -> List[str]:
        """
        Safely read file as list of lines.

        Args:
            file_path: Path to the file to read

        Returns:
            List of lines, or empty list if reading fails
        """
        content = self.read_text(file_path)
        if content is None:
            return []

        return content.splitlines(keepends=True)

    def safe_glob_read(self, pattern: str, base_path: Union[str, Path] = ".") -> dict:
        """
        Safely read multiple files matching a glob pattern.

        Args:
            pattern: Glob pattern to match files
            base_path: Base directory for glob search

        Returns:
            Dictionary mapping file paths to their contents
        """
        base_path = Path(base_path)
        results = {}

        try:
            for file_path in base_path.glob(pattern):
                if file_path.is_file():
                    content = self.read_text(file_path)
                    if content is not None:
                        results[str(file_path)] = content
        except Exception as e:
            print(f"Error: Failed to glob pattern '{pattern}': {e}")

        return results

    def is_safe_file(self, file_path: Union[str, Path]) -> bool:
        """
        Check if file can be safely read.

        Args:
            file_path: Path to the file to check

        Returns:
            True if file can be read safely, False otherwise
        """
        content = self.read_text(file_path)
        return content is not None


# Global convenience functions
def safe_read_file(file_path: Union[str, Path], encodings: Optional[List[str]] = None) -> Optional[str]:
    """
    Convenience function to safely read a single file.

    Args:
        file_path: Path to the file to read
        encodings: List of encodings to try in order

    Returns:
        File content as string, or None if reading fails
    """
    reader = SafeFileReader(encodings=encodings)
    return reader.read_text(file_path)


def safe_read_lines(file_path: Union[str, Path], encodings: Optional[List[str]] = None) -> List[str]:
    """
    Convenience function to safely read file lines.

    Args:
        file_path: Path to the file to read
        encodings: List of encodings to try in order

    Returns:
        List of lines, or empty list if reading fails
    """
    reader = SafeFileReader(encodings=encodings)
    return reader.read_lines(file_path)


def safe_glob_read(
    pattern: str,
    base_path: Union[str, Path] = ".",
    encodings: Optional[List[str]] = None,
) -> dict:
    """
    Convenience function to safely read multiple files.

    Args:
        pattern: Glob pattern to match files
        base_path: Base directory for search
        encodings: List of encodings to try in order

    Returns:
        Dictionary mapping file paths to their contents
    """
    reader = SafeFileReader(encodings=encodings)
    return reader.safe_glob_read(pattern, base_path)


if __name__ == "__main__":
    # Test the safe file reader
    import logging
    import sys

    # Set up logging
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger(__name__)

    # Test reading this file
    reader = SafeFileReader()
    content = reader.read_text(__file__)

    if content:
        print("Successfully read file with safe encoding detection")
        print(f"File length: {len(content)} characters")
    else:
        print("Failed to read file")
        sys.exit(1)

    print("SafeFileReader test completed successfully!")
