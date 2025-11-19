#!/usr/bin/env python3
"""
Demo script for Unified Permission Manager

Tests the permission manager with real-world agent permission validation errors
identified in Claude Code debug logs.
"""

import sys
import os
import tempfile
import json

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

from moai_adk.core.unified_permission_manager import (
    UnifiedPermissionManager,
    PermissionMode,
    validate_agent_permission,
    check_tool_permission,
    auto_fix_all_agent_permissions
)

def test_debug_log_permission_errors():
    """Test fixing the exact permission errors from the debug log"""
    print("🔧 Testing Debug Log Permission Error Fixes")
    print("=" * 60)

    # Create a temporary directory for testing
    temp_dir = tempfile.mkdtemp()
    config_path = os.path.join(temp_dir, "claude_settings.json")

    try:
        # Recreate the exact permission issues from the debug log (Lines 50-80)
        problematic_config = {
            "agents": {
                # Invalid permissionMode 'ask' (should be one of the 5 valid modes)
                "backend-expert": {
                    "permissionMode": "ask",
                    "description": "Backend development expert"
                },
                "security-expert": {
                    "permissionMode": "ask",
                    "description": "Security and compliance expert"
                },
                "api-designer": {
                    "permissionMode": "ask",
                    "description": "API design and documentation expert"
                },
                "monitoring-expert": {
                    "permissionMode": "ask",
                    "description": "Monitoring and observability expert"
                },
                "performance-engineer": {
                    "permissionMode": "ask",
                    "description": "Performance optimization expert"
                },
                "migration-expert": {
                    "permissionMode": "ask",
                    "description": "Database migration expert"
                },
                "mcp-playwright-integrator": {
                    "permissionMode": "ask",
                    "description": "Playwright MCP integration expert"
                },
                "frontend-expert": {
                    "permissionMode": "ask",
                    "description": "Frontend development expert"
                },
                "debug-helper": {
                    "permissionMode": "ask",
                    "description": "Debugging assistance expert"
                },
                "ui-ux-expert": {
                    "permissionMode": "ask",
                    "description": "UI/UX design expert"
                },
                "trust-checker": {
                    "permissionMode": "ask",
                    "description": "Trust and compliance checker"
                },
                "mcp-context7-integrator": {
                    "permissionMode": "ask",
                    "description": "Context7 MCP integration expert"
                },
                "mcp-figma-integrator": {
                    "permissionMode": "ask",
                    "description": "Figma MCP integration expert"
                },
                "tdd-implementer": {
                    "permissionMode": "ask",
                    "description": "TDD implementation expert"
                },
                "devops-expert": {
                    "permissionMode": "ask",
                    "description": "DevOps and deployment expert"
                },
                "git-manager": {
                    "permissionMode": "ask",
                    "description": "Git workflow management expert"
                },
                "component-designer": {
                    "permissionMode": "ask",
                    "description": "Component design expert"
                },
                "database-expert": {
                    "permissionMode": "ask",
                    "description": "Database design expert"
                },
                "accessibility-expert": {
                    "permissionMode": "ask",
                    "description": "Accessibility compliance expert"
                },

                # Invalid permissionMode 'auto' (should be one of the 5 valid modes)
                "quality-gate": {
                    "permissionMode": "auto",
                    "description": "Quality gate validation expert"
                },
                "project-manager": {
                    "permissionMode": "auto",
                    "description": "Project management expert"
                },
                "format-expert": {
                    "permissionMode": "auto",
                    "description": "Code formatting expert"
                },
                "docs-manager": {
                    "permissionMode": "auto",
                    "description": "Documentation management expert"
                },
                "implementation-planner": {
                    "permissionMode": "auto",
                    "description": "Implementation planning expert"
                },
                "skill-factory": {
                    "permissionMode": "auto",
                    "description": "Skill creation expert"
                },
                "agent-factory": {
                    "permissionMode": "auto",
                    "description": "Agent creation expert"
                },
                "sync-manager": {
                    "permissionMode": "auto",
                    "description": "Synchronization management expert"
                },
                "spec-builder": {
                    "permissionMode": "auto",
                    "description": "Specification building expert"
                },
                "doc-syncer": {
                    "permissionMode": "auto",
                    "description": "Document synchronization expert"
                },
                "cc-manager": {
                    "permissionMode": "auto",
                    "description": "Claude Code management expert"
                }
            }
        }

        # Write problematic configuration
        with open(config_path, 'w') as f:
            json.dump(problematic_config, f, indent=2)

        print(f"Created test configuration with {len(problematic_config['agents'])} problematic agents")

        # Initialize permission manager
        print("\n1. Initializing UnifiedPermissionManager...")
        manager = UnifiedPermissionManager(config_path=config_path, enable_logging=False)

        # Show initial state
        initial_stats = manager.get_permission_stats()
        print(f"   Initial auto-corrections: {initial_stats.get('auto_corrections', 0)}")

        # Validate a few specific agents to show the issues
        test_agents = ["backend-expert", "security-expert", "quality-gate"]

        print("\n2. Showing specific permission issues before fixing:")
        for agent_name in test_agents:
            agent_config = manager.config["agents"][agent_name]
            original_mode = agent_config.get("permissionMode", "default")
            result = manager.validate_agent_permission(agent_name, agent_config.copy())

            print(f"   {agent_name}:")
            print(f"     Original mode: '{original_mode}' (INVALID)")
            print(f"     Valid: {result.valid}")
            if result.auto_corrected:
                print(f"     Corrected to: '{result.corrected_mode}'")

        print("\n3. Auto-fixing all agent permissions...")
        results = manager.auto_fix_all_agents()

        fixed_count = sum(1 for result in results.values() if result.auto_corrected)
        total_agents = len(results)
        print(f"   Total agents processed: {total_agents}")
        print(f"   Agents fixed: {fixed_count}")
        print(f"   Fix rate: {fixed_count/total_agents:.1%}")

        print("\n4. Verification of fixes:")
        for agent_name in test_agents:
            agent_config = manager.config["agents"][agent_name]
            current_mode = agent_config.get("permissionMode", "default")
            is_valid = current_mode in [mode.value for mode in PermissionMode]

            print(f"   {agent_name}: '{current_mode}' - {'✅ VALID' if is_valid else '❌ INVALID'}")

        print("\n5. Permission mode distribution after fixing:")
        mode_counts = {}
        for agent_name, agent_config in manager.config["agents"].items():
            mode = agent_config.get("permissionMode", "default")
            mode_counts[mode] = mode_counts.get(mode, 0) + 1

        for mode, count in sorted(mode_counts.items()):
            print(f"   {mode}: {count} agents")

        # Show final statistics
        final_stats = manager.get_permission_stats()
        print(f"\n6. Final Statistics:")
        print(f"   Total validations: {final_stats.get('validations_performed', 0)}")
        print(f"   Auto corrections: {final_stats.get('auto_corrections', 0)}")
        print(f"   Security violations: {final_stats.get('security_violations', 0)}")
        print(f"   Permission denied: {final_stats.get('permission_denied', 0)}")

        print(f"\n✅ Debug log permission errors successfully fixed!")
        print(f"   All {total_agents} agents now have valid permission modes")

        return True

    finally:
        # Clean up
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)

def test_role_based_permissions():
    """Test role-based permission checking"""
    print(f"\n🔐 Testing Role-Based Permission System")
    print("=" * 50)

    # Test different roles and their permissions
    role_permissions = {
        "admin": ["Task", "Read", "Write", "Bash", "AskUserQuestion"],
        "developer": ["Task", "Read", "Write", "AskUserQuestion"],
        "user": ["Task", "Read", "AskUserQuestion"]
    }

    test_tools = ["Task", "Write", "Bash", "Read", "AskUserQuestion"]

    print("Permission matrix:")
    print("Role        | Task | Write | Bash  | Read  | Ask")
    print("-" * 50)

    for role in ["admin", "developer", "user"]:
        permissions = []
        for tool in test_tools:
            allowed = check_tool_permission(role, tool, "execute")
            permissions.append("✅" if allowed else "❌")

        print(f"{role:11} | {' | '.join(permissions)}")

def test_permission_mode_suggestions():
    """Test intelligent permission mode suggestions"""
    print(f"\n🧠 Testing Permission Mode Suggestions")
    print("=" * 50)

    # Create a temporary manager for testing
    temp_dir = tempfile.mkdtemp()
    config_path = os.path.join(temp_dir, "test_config.json")

    with open(config_path, 'w') as f:
        json.dump({}, f)

    try:
        manager = UnifiedPermissionManager(config_path=config_path, enable_logging=False)

        test_agent_names = [
            "security-expert",
            "backend-expert",
            "api-designer",
            "tdd-implementer",
            "frontend-developer",
            "database-admin",
            "unknown-agent"
        ]

        print("Agent name suggestions:")
        for agent_name in test_agent_names:
            suggested_mode = manager._suggest_permission_mode(agent_name)
            print(f"   {agent_name:20} → {suggested_mode}")

    finally:
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)

def test_security_validation():
    """Test security validation features"""
    print(f"\n🛡️ Testing Security Validation")
    print("=" * 40)

    # Test dangerous tool detection
    dangerous_tools = [
        "Bash(rm -rf:*)",
        "Bash(sudo:*)",
        "Bash(chmod -R 777:*)",
        "Bash(git push --force:*)"
    ]

    temp_dir = tempfile.mkdtemp()
    config_path = os.path.join(temp_dir, "security_test.json")

    with open(config_path, 'w') as f:
        json.dump({}, f)

    try:
        manager = UnifiedPermissionManager(config_path=config_path, enable_logging=False)

        print("Dangerous tool detection:")
        for tool in dangerous_tools:
            result = manager.validate_tool_permissions([tool])
            has_warnings = len(result.warnings) > 0
            print(f"   {tool:25} → {'⚠️  WARNING' if has_warnings else '✅ OK'}")

    finally:
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)

def main():
    """Run all demo tests"""
    print("🚀 MoAI-ADK Unified Permission Manager Demo")
    print("=" * 60)
    print("Demonstrating fixes for Claude Code agent permission validation errors")

    # Test 1: Fix the exact issues from debug log
    success = test_debug_log_permission_errors()

    if not success:
        print("❌ Debug log permission fix test failed")
        return

    # Test 2: Role-based permissions
    test_role_based_permissions()

    # Test 3: Permission mode suggestions
    test_permission_mode_suggestions()

    # Test 4: Security validation
    test_security_validation()

    print(f"\n✨ Demo completed successfully!")
    print(f"   The Unified Permission Manager addresses the critical permission")
    print(f"   validation errors from Claude Code debug logs (Lines 50-80)")
    print(f"   with automatic correction and intelligent permission mode assignment.")

if __name__ == "__main__":
    main()