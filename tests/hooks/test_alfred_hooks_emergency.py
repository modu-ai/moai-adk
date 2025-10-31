#!/usr/bin/env python3
# @TEST:HOOKS-EMERGENCY-001 | @CODE:HOOKS-EMERGENCY-001:IMPORT-FIX
"""RED Phase: Emergency tests for alfred_hooks.py import and timeout fixes

This test module verifies:
1. sys.path is configured BEFORE imports (preventing ImportError)
2. CrossPlatformTimeout is used instead of signal.SIGALRM (Windows compatibility)
3. timeout variable is properly initialized (no NameError in finally block)
4. HookTimeoutError is replaced with PlatformTimeoutError

Test Strategy:
- Test 1: Verify sys.path includes hooks directory before imports
- Test 2: Verify timeout handler uses PlatformTimeoutError (not HookTimeoutError)
- Test 3: Verify CrossPlatformTimeout context manager is used
- Test 4: Verify no signal.SIGALRM usage (cross-platform check)

Expected: ALL TESTS FAIL (RED phase)
"""

import ast
import sys
from pathlib import Path


def test_syspath_configured_before_imports():
    """Test that sys.path configuration happens before imports

    EXPECTED: FAIL (currently sys.path set at line 76-78, imports at line 61-73)
    FIX: Move sys.path setup to line 54-56 (before imports)
    """
    hooks_file = Path(__file__).parent.parent.parent / ".claude/hooks/alfred/alfred_hooks.py"
    assert hooks_file.exists(), f"Hook file not found: {hooks_file}"

    with open(hooks_file, "r", encoding="utf-8") as f:
        content = f.read()

    tree = ast.parse(content)

    # Find line numbers of sys.path configuration and imports
    syspath_line = None
    first_import_line = None

    for node in ast.walk(tree):
        # Check for sys.path.insert() or sys.path assignment
        if isinstance(node, ast.Expr) and isinstance(node.value, ast.Call):
            if (hasattr(node.value.func, 'attr') and
                node.value.func.attr == 'insert' and
                hasattr(node.value.func.value, 'attr') and
                node.value.func.value.attr == 'path'):
                syspath_line = node.lineno

        # Check for import statements (excluding standard library imports in header)
        if isinstance(node, (ast.Import, ast.ImportFrom)):
            if node.lineno > 60:  # Skip header imports
                if first_import_line is None or node.lineno < first_import_line:
                    first_import_line = node.lineno

    assert syspath_line is not None, "sys.path configuration not found"
    assert first_import_line is not None, "Module imports not found"

    # CRITICAL: sys.path MUST be configured BEFORE imports
    assert syspath_line < first_import_line, (
        f"sys.path configured at line {syspath_line}, "
        f"but imports start at line {first_import_line}. "
        f"sys.path MUST come BEFORE imports to prevent ImportError"
    )


def test_no_hook_timeout_error_reference():
    """Test that HookTimeoutError is NOT referenced (should use PlatformTimeoutError)

    EXPECTED: FAIL (currently HookTimeoutError at line 83)
    FIX: Replace with PlatformTimeoutError
    """
    hooks_file = Path(__file__).parent.parent.parent / ".claude/hooks/alfred/alfred_hooks.py"

    with open(hooks_file, "r", encoding="utf-8") as f:
        content = f.read()

    # HookTimeoutError should NOT appear in file (undefined class)
    assert "HookTimeoutError" not in content, (
        "HookTimeoutError is undefined. Use PlatformTimeoutError instead."
    )


def test_cross_platform_timeout_used():
    """Test that CrossPlatformTimeout context manager is used instead of signal.SIGALRM

    EXPECTED: FAIL (currently signal.SIGALRM at line 125-126)
    FIX: Replace with CrossPlatformTimeout context manager
    """
    hooks_file = Path(__file__).parent.parent.parent / ".claude/hooks/alfred/alfred_hooks.py"

    with open(hooks_file, "r", encoding="utf-8") as f:
        content = f.read()

    # signal.SIGALRM should NOT be used (Windows incompatible)
    assert "signal.SIGALRM" not in content, (
        "signal.SIGALRM is not cross-platform compatible. "
        "Use CrossPlatformTimeout context manager instead."
    )

    # signal.alarm() should NOT be used
    assert "signal.alarm(" not in content, (
        "signal.alarm() is not cross-platform compatible. "
        "Use CrossPlatformTimeout context manager instead."
    )


def test_timeout_variable_initialization():
    """Test that timeout variable is properly initialized before finally block

    EXPECTED: FAIL (currently timeout.cancel() at line 202 without initialization)
    FIX: Use CrossPlatformTimeout context manager with proper scope
    """
    hooks_file = Path(__file__).parent.parent.parent / ".claude/hooks/alfred/alfred_hooks.py"

    with open(hooks_file, "r", encoding="utf-8") as f:
        content = f.read()

    tree = ast.parse(content)

    # Find main() function
    main_func = None
    for node in ast.walk(tree):
        if isinstance(node, ast.FunctionDef) and node.name == "main":
            main_func = node
            break

    assert main_func is not None, "main() function not found"

    # Check if there's a finally block with timeout.cancel()
    has_timeout_cancel = False
    has_timeout_init = False

    for node in ast.walk(main_func):
        if isinstance(node, ast.Try):
            for stmt in node.finalbody:
                for child in ast.walk(stmt):
                    if isinstance(child, ast.Attribute) and child.attr == "cancel":
                        has_timeout_cancel = True

        # Check for timeout initialization (with statement or assignment)
        if isinstance(node, ast.With):
            for item in node.items:
                if item.optional_vars and isinstance(item.optional_vars, ast.Name):
                    if item.optional_vars.id == "timeout":
                        has_timeout_init = True

    if has_timeout_cancel:
        # If timeout.cancel() exists, timeout MUST be initialized
        assert has_timeout_init, (
            "timeout.cancel() called in finally block, "
            "but timeout variable not initialized. "
            "Use 'with CrossPlatformTimeout(5) as timeout:' pattern."
        )


def test_no_signal_handler_function():
    """Test that _hook_timeout_handler signal function is removed

    EXPECTED: FAIL (currently defined at line 81-83)
    FIX: Remove function (replaced by CrossPlatformTimeout)
    """
    hooks_file = Path(__file__).parent.parent.parent / ".claude/hooks/alfred/alfred_hooks.py"

    with open(hooks_file, "r", encoding="utf-8") as f:
        content = f.read()

    tree = ast.parse(content)

    # Check for _hook_timeout_handler function
    for node in ast.walk(tree):
        if isinstance(node, ast.FunctionDef):
            assert node.name != "_hook_timeout_handler", (
                "_hook_timeout_handler function should be removed. "
                "Use CrossPlatformTimeout context manager instead."
            )


if __name__ == "__main__":
    # Run tests manually to verify RED phase
    print("🔴 RED Phase: Running emergency tests (expect ALL FAIL)")
    print("=" * 60)

    tests = [
        ("sys.path before imports", test_syspath_configured_before_imports),
        ("No HookTimeoutError", test_no_hook_timeout_error_reference),
        ("CrossPlatformTimeout used", test_cross_platform_timeout_used),
        ("timeout initialization", test_timeout_variable_initialization),
        ("No signal handler", test_no_signal_handler_function),
    ]

    failed = 0
    for name, test_func in tests:
        try:
            test_func()
            print(f"✅ {name}: PASS")
        except AssertionError as e:
            print(f"❌ {name}: FAIL")
            print(f"   {e}")
            failed += 1

    print("=" * 60)
    print(f"Results: {failed}/{len(tests)} tests FAILED (RED phase expected)")

    if failed == len(tests):
        print("✅ RED phase successful: All tests failing as expected")
        sys.exit(0)
    else:
        print("⚠️ RED phase incomplete: Some tests already passing")
        sys.exit(1)
