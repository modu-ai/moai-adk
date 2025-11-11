#!/usr/bin/env python3
"""
Safe File Reader Utility

Provides safe file reading with multiple encoding fallbacks and error handling.
Resolves UTF-8 encoding issues found in MoAI-ADK project.

Author: Alfred@MoAI
Date: 2025-11-11
"""

import os
from pathlib import Path
from typing import Optional, List, Union


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
        'utf-8',        # Standard UTF-8
        'cp1252',       # Windows-1252 (Western European)
        'iso-8859-1',   # Latin-1 (Western European)
        'latin1',        # Alternative Latin-1
        'utf-16',       # UTF-16 with BOM detection
        'ascii',         # Pure ASCII fallback
    ]

    def __init__(self, encodings: Optional[List[str]] = None, errors: str = 'ignore'):
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
            return file_path.read_text(encoding='utf-8', errors=self.errors)
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

    def safe_glob_read(self, pattern: str, base_path: Union[str, Path] = '.') -> dict:
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


def safe_glob_read(pattern: str, base_path: Union[str, Path] = '.',
                    encodings: Optional[List[str]] = None) -> dict:
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


# TAG-specific utility functions
def extract_tags_from_file(file_path: Union[str, Path], tag_pattern: str = r'@\w+:\w+-\d+') -> List[str]:
    """
    Extract TAG patterns from a file safely.

    Args:
        file_path: Path to the file to analyze
        tag_pattern: Regex pattern for TAG matching

    Returns:
        List of found TAGs, or empty list if reading fails
    """
    import re

    content = safe_read_file(file_path)
    if not content:
        return []

    return re.findall(tag_pattern, content)


def safe_extract_spec_tags(file_path: Union[str, Path]) -> List[str]:
    """
    Extract SPEC TAGs from a file safely.

    Args:
        file_path: Path to the file to analyze

    Returns:
        List of found SPEC TAGs, or empty list if reading fails
    """
    return extract_tags_from_file(file_path, r'@SPEC:(\w+-\d+)')


def safe_extract_code_tags(file_path: Union[str, Path]) -> List[str]:
    """
    Extract CODE TAGs from a file safely.

    Args:
        file_path: Path to the file to analyze

    Returns:
        List of found CODE TAGs, or empty list if reading fails
    """
    return extract_tags_from_file(file_path, r'@CODE:(\w+-\d+)')


def safe_extract_test_tags(file_path: Union[str, Path]) -> List[str]:
    """
    Extract TEST TAGs from a file safely.

    Args:
        file_path: Path to the file to analyze

    Returns:
        List of found TEST TAGs, or empty list if reading fails
    """
    return extract_tags_from_file(file_path, r'@TEST:(\w+-\d+)')


def safe_extract_doc_tags(file_path: Union[str, Path]) -> List[str]:
    """
    Extract DOC TAGs from a file safely.

    Args:
        file_path: Path to the file to analyze

    Returns:
        List of found DOC TAGs, or empty list if reading fails
    """
    return extract_tags_from_file(file_path, r'@DOC:(\w+-\d+)')


# Batch processing utilities
def analyze_tag_connectivity(base_path: Union[str, Path] = '.') -> dict:
    """
    Analyze TAG connectivity across all files in a directory safely.

    Args:
        base_path: Base directory to analyze

    Returns:
        Dictionary with TAG connectivity statistics
    """
    base_path = Path(base_path)

    reader = SafeFileReader()

    spec_tags = set()
    code_tags = set()
    test_tags = set()
    doc_tags = set()

    # Read SPEC files
    spec_files = reader.safe_glob_read("**/spec.md", base_path / ".moai/specs")
    for file_path, content in spec_files.items():
        spec_tags.update(extract_tags_from_file(file_path, r'@SPEC:(\w+-\d+)'))

    # Read Python files
    py_files = reader.safe_glob_read("**/*.py", base_path)
    for file_path, content in py_files.items():
        code_tags.update(extract_tags_from_file(file_path, r'@CODE:(\w+-\d+)'))
        test_tags.update(extract_tags_from_file(file_path, r'@TEST:(\w+-\d+)'))

    # Read Markdown files
    md_files = reader.safe_glob_read("**/*.md", base_path)
    for file_path, content in md_files.items():
        doc_tags.update(extract_tags_from_file(file_path, r'@DOC:(\w+-\d+)'))

    # Calculate connectivity
    complete_chains = 0
    for spec in spec_tags:
        if spec in code_tags and spec in test_tags and spec in doc_tags:
            complete_chains += 1

    connectivity_rate = (complete_chains / len(spec_tags)) * 100 if spec_tags else 0

    return {
        'spec_tags': list(spec_tags),
        'code_tags': list(code_tags),
        'test_tags': list(test_tags),
        'doc_tags': list(doc_tags),
        'complete_chains': complete_chains,
        'connectivity_rate': connectivity_rate,
        'total_specs': len(spec_tags)
    }


# UTF-8 File Writing Utilities

class SafeFileWriter:
    """Safe file writer with explicit UTF-8 encoding specification."""

    @staticmethod
    def write_text(file_path: Union[str, Path], content: str, encoding: str = 'utf-8') -> bool:
        """
        Safely write text file with explicit UTF-8 encoding.

        Args:
            file_path: Path to the file to write
            content: Content to write
            encoding: Encoding to use (default: utf-8)

        Returns:
            True if successful, False otherwise
        """
        try:
            file_path = Path(file_path)
            # Ensure parent directory exists
            file_path.parent.mkdir(parents=True, exist_ok=True)

            # Write with explicit UTF-8 encoding
            file_path.write_text(content, encoding=encoding)
            return True
        except Exception as e:
            print(f"Error writing file {file_path}: {e}")
            return False

    @staticmethod
    def write_lines(file_path: Union[str, Path], lines: List[str], encoding: str = 'utf-8') -> bool:
        """
        Safely write lines to file with explicit UTF-8 encoding.

        Args:
            file_path: Path to the file to write
            lines: List of lines to write
            encoding: Encoding to use (default: utf-8)

        Returns:
            True if successful, False otherwise
        """
        content = '\n'.join(lines)
        if content and not content.endswith('\n'):
            content += '\n'
        return SafeFileWriter.write_text(file_path, content, encoding)


# Global convenience functions for safe writing
def safe_write_file(file_path: Union[str, Path], content: str, encoding: str = 'utf-8') -> bool:
    """
    Convenience function to safely write a single file.

    Args:
        file_path: Path to the file to write
        content: Content to write
        encoding: Encoding to use (default: utf-8)

    Returns:
        True if successful, False otherwise
    """
    return SafeFileWriter.write_text(file_path, content, encoding)


def safe_write_lines(file_path: Union[str, Path], lines: List[str], encoding: str = 'utf-8') -> bool:
    """
    Convenience function to safely write lines to file.

    Args:
        file_path: Path to the file to write
        lines: List of lines to write
        encoding: Encoding to use (default: utf-8)

    Returns:
        True if successful, False otherwise
    """
    return SafeFileWriter.write_lines(file_path, lines, encoding)


if __name__ == "__main__":
    # Example usage and testing
    import sys

    if len(sys.argv) > 1:
        file_path = sys.argv[1]
        print(f"Testing safe read of: {file_path}")

        reader = SafeFileReader()
        content = reader.read_text(file_path)

        if content:
            print(f"✅ Successfully read {len(content)} characters")
            print(f"First 100 characters:\n{content[:100]}")

            # Test TAG extraction
            spec_tags = safe_extract_spec_tags(file_path)
            if spec_tags:
                print(f"Found SPEC TAGs: {spec_tags}")
        else:
            print(f"❌ Failed to read file: {file_path}")
    else:
        print("Usage: python safe_file_reader.py <file_path>")
        print("\nTesting TAG connectivity analysis...")

        result = analyze_tag_connectivity()
        print(f"TAG connectivity analysis:")
        print(f"  SPEC TAGs: {len(result['spec_tags'])}")
        print(f"  CODE TAGs: {len(result['code_tags'])}")
        print(f"  TEST TAGs: {len(result['test_tags'])}")
        print(f"  DOC TAGs: {len(result['doc_tags'])}")
        print(f"  Complete chains: {result['complete_chains']}")
        print(f"  Connectivity rate: {result['connectivity_rate']:.1f}%")