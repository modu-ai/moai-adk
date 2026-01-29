#!/usr/bin/env python3
"""
Component Validation Script for React/TSX Components

This script validates React component files for:
1. Props interface detection (must end with 'Props')
2. Hardcoded hex value detection in className attributes

Usage:
    python validate_component.py path/to/component.tsx

Exit codes:
    0: Validation passed
    1: Validation failed (missing Props interface or hardcoded hex values)
    2: Parse error (invalid TypeScript/TSX syntax)
"""

import ast
import re
import sys
from pathlib import Path
from typing import List, Set, Tuple


class ComponentVisitor(ast.NodeVisitor):
    """AST visitor to extract component validation information."""

    def __init__(self) -> None:
        self.has_props_interface: bool = False
        self.hardcoded_hex_values: List[Tuple[int, str]] = []
        self.imports: Set[str] = set()

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        """Track React imports."""
        if node.module and "react" in node.module.lower():
            for alias in node.names:
                self.imports.add(alias.name)
        self.generic_visit(node)

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        """Check for Props interface (TypeScript-like syntax in comments)."""
        # In Python AST, TypeScript interfaces are parsed as comments or strings
        # We check the name for 'Props' suffix
        if node.name.endswith("Props"):
            self.has_props_interface = True
        self.generic_visit(node)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        """Check for component functions and their decorators."""
        # Check if this is a component (starts with uppercase)
        if node.name and node.name[0].isupper():
            # Look for Props interface in the file
            pass
        self.generic_visit(node)

    def visit_Str(self, node: ast.Str) -> None:
        """Check string literals for hardcoded hex values."""
        # Check for className attributes with hex colors
        if node.s:
            hex_pattern = r"#[0-9A-Fa-f]{6}"
            matches = re.findall(hex_pattern, node.s)
            for match in matches:
                line = getattr(node, "lineno", 0)
                self.hardcoded_hex_values.append((line, match))
        self.generic_visit(node)

    def visit_Constant(self, node: ast.Constant) -> None:
        """Check constant values for hardcoded hex values."""
        if isinstance(node.value, str):
            hex_pattern = r"#[0-9A-Fa-f]{6}"
            matches = re.findall(hex_pattern, node.value)
            for match in matches:
                line = getattr(node, "lineno", 0)
                self.hardcoded_hex_values.append((line, match))
        self.generic_visit(node)


def extract_typescript_interfaces(source: str) -> List[str]:
    """
    Extract TypeScript interface names from source code using regex.
    This is a fallback for TSX files that may not parse fully as Python.
    """
    # Match interface declarations
    pattern = r"interface\s+(\w+Props)\s*"
    matches = re.findall(pattern, source)
    return matches


def find_hex_in_classnames(source: str) -> List[Tuple[int, str]]:
    """
    Find hardcoded hex values in className attributes using regex.
    This is more reliable for TSX files than AST parsing.
    """
    hex_pattern = r'className\s*=\s*["\'][^"\']*#([0-9A-Fa-f]{6})[^"\']*["\']'
    lines = source.split("\n")

    results = []
    for line_num, line in enumerate(lines, start=1):
        matches = re.findall(hex_pattern, line)
        for match in matches:
            results.append((line_num, f"#{match}"))

    # Also check for inline style objects with hex values
    style_pattern = r"style\s*=\s*\{\{[^}]*#([0-9A-Fa-f]{6})[^}]*\}\}"
    for line_num, line in enumerate(lines, start=1):
        matches = re.findall(style_pattern, line)
        for match in matches:
            results.append((line_num, f"#{match}"))

    return results


def validate_component(file_path: str) -> Tuple[bool, List[str]]:
    """
    Validate a React component file.

    Returns:
        Tuple of (is_valid, error_messages)
    """
    path = Path(file_path)

    if not path.exists():
        return False, [f"File not found: {file_path}"]

    if path.suffix not in [".tsx", ".ts", ".jsx", ".js"]:
        return False, [f"Invalid file extension: {path.suffix}"]

    try:
        source = path.read_text(encoding="utf-8")
    except Exception as e:
        return False, [f"Error reading file: {e}"]

    errors = []

    # Check for Props interface
    interfaces = extract_typescript_interfaces(source)
    has_props = any(name.endswith("Props") for name in interfaces)

    if not has_props:
        # Also check for type declarations
        type_pattern = r"type\s+(\w+Props)\s*="
        type_matches = re.findall(type_pattern, source)
        has_props = any(name.endswith("Props") for name in type_matches)

    if not has_props:
        errors.append("MISSING: Props interface (must end in 'Props')")

    # Check for hardcoded hex values
    hex_values = find_hex_in_classnames(source)

    if hex_values:
        for line_num, hex_value in hex_values:
            errors.append(f"STYLE: Hardcoded hex value '{hex_value}' found at line {line_num}")

    return len(errors) == 0, errors


def main() -> int:
    """Main entry point."""
    if len(sys.argv) < 2:
        print("Usage: python validate_component.py <path-to-component.tsx>")
        return 2

    file_path = sys.argv[1]
    filename = Path(file_path).name

    print(f"--- Validation for: {filename} ---")

    is_valid, errors = validate_component(file_path)

    if not errors:
        print("✅ Props declaration found.")
    else:
        for error in errors:
            if error.startswith("MISSING"):
                print(f"❌ {error}")
            else:
                print(f"❌ {error}")

    if not any(e.startswith("STYLE") for e in errors):
        print("✅ No hardcoded hex values found.")

    if is_valid:
        print("\n✨ COMPONENT VALID.")
        return 0
    else:
        print("\n🚫 VALIDATION FAILED.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
