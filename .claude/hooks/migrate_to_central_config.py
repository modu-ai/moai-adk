#!/usr/bin/env python3
"""
Migration script to replace individual load_config() functions with central ConfigManager.
This script automatically updates Hook files to use the optimized central configuration system.
"""

import re
from pathlib import Path
from typing import Dict, List, Tuple


def find_load_config_files() -> List[Path]:
    """Find all Python files containing load_config() functions."""
    hooks_dir = Path(".")
    files_with_load_config = []

    for py_file in hooks_dir.glob("**/*.py"):
        if "test_" in py_file.name:
            continue  # Skip test files

        try:
            content = py_file.read_text(encoding='utf-8')
            if "def load_config()" in content:
                files_with_load_config.append(py_file)
        except (UnicodeDecodeError, PermissionError):
            continue

    return files_with_load_config


def analyze_load_config_function(file_path: Path) -> Dict:
    """Analyze a load_config function to understand its pattern."""
    content = file_path.read_text(encoding='utf-8')

    # Find the load_config function
    func_match = re.search(r'def load_config\(\s*\).*?(?=\n\S|\Z)', content, re.DOTALL)

    if not func_match:
        return {"pattern": "unknown"}

    func_body = func_match.group(0)

    # Analyze patterns
    patterns = {
        "basic_json": 'config_file = Path' in func_body and 'json.load(f)' in func_body,
        "default_return": 'return {}' in func_body or 'return DEFAULT_CONFIG' in func_body,
        "error_handling": 'try:' in func_body and 'except' in func_body,
        "custom_path": '.moai/config.json' not in func_body,
    }

    # Determine the migration pattern
    if patterns["basic_json"] and patterns["error_handling"]:
        return {"pattern": "standard", "file": file_path}
    elif patterns["custom_path"]:
        return {"pattern": "custom_path", "file": file_path}
    else:
        return {"pattern": "custom", "file": file_path}


def generate_import_replacement() -> str:
    """Generate the import statement for central ConfigManager."""
    return 'from alfred.shared.core.config_manager import get_config_manager, get_config'


def generate_load_config_replacement(analysis: Dict) -> str:
    """Generate the replacement load_config function based on analysis."""
    pattern = analysis["pattern"]

    if pattern == "standard":
        return '''def load_config() -> Dict[str, Any]:
    """Load configuration using central ConfigManager for optimized performance."""
    return get_config_manager().load_config()'''

    elif pattern == "custom_path":
        return '''def load_config() -> Dict[str, Any]:
    """Load configuration using central ConfigManager.

    Note: Consider using get_config(key_path, default) for specific values
    to benefit from caching and performance optimization.
    """
    return get_config_manager().load_config()'''

    else:
        return '''def load_config() -> Dict[str, Any]:
    """Load configuration using central ConfigManager for optimized performance."""
    return get_config_manager().load_config()'''


def migrate_file(file_path: Path) -> Tuple[bool, str]:
    """Migrate a single file to use central ConfigManager."""
    try:
        content = file_path.read_text(encoding='utf-8')
        original_content = content

        # Check if already migrated
        if 'from alfred.shared.core.config_manager import' in content:
            return False, "Already migrated"

        # Analyze the load_config function
        analysis = analyze_load_config_function(file_path)

        # Add import statement (after existing imports)
        import_replacement = generate_import_replacement()

        # Find the last import statement
        import_lines = [line for line in content.split('\n') if line.strip().startswith(('import ', 'from '))]
        if import_lines:
            last_import_line = content.rfind(import_lines[-1])
            if last_import_line != -1:
                end_of_line = content.find('\n', last_import_line)
                if end_of_line != -1:
                    content = content[:end_of_line + 1] + f'\n{import_replacement}\n' + content[end_of_line + 1:]
        else:
            # No imports found, add at the beginning
            content = f'{import_replacement}\n\n' + content

        # Replace load_config function
        func_replacement = generate_load_config_replacement(analysis)

        # Find and replace the function
        func_pattern = r'def load_config\(\s*\).*?(?=\n\S|\Z)'
        new_content = re.sub(func_pattern, func_replacement, content, flags=re.DOTALL)

        # Only write if content changed
        if new_content != original_content:
            file_path.write_text(new_content, encoding='utf-8')
            return True, f"Migrated with pattern: {analysis['pattern']}"
        else:
            return False, "No changes needed"

    except Exception as e:
        return False, f"Error: {str(e)}"


def generate_usage_recommendations(file_path: Path) -> List[str]:
    """Generate recommendations for optimal ConfigManager usage."""
    recommendations = []

    try:
        content = file_path.read_text(encoding='utf-8')

        # Check for direct config access patterns
        if 'config[' in content and 'load_config()' in content:
            recommendations.append(
                "Consider using get_config('key.path', default) instead of load_config()['key'] "
                "for better performance and caching benefits."
            )

        # Check for multiple load_config calls
        load_count = content.count('load_config()')
        if load_count > 3:
            recommendations.append(
                f"File calls load_config() {load_count} times. Consider storing the result "
                "or using get_config() for specific values to optimize performance."
            )

        # Check for timeout configuration access
        if 'timeout' in content.lower():
            recommendations.append(
                "Consider using get_timeout_seconds() helper function for timeout values."
            )

        return recommendations

    except Exception:
        return []


def create_migration_report(results: List[Tuple[Path, bool, str]]) -> None:
    """Create a detailed migration report."""
    print("=" * 80)
    print("🔄 MoAI-ADK Hook Configuration Migration Report")
    print("=" * 80)

    successful = sum(1 for _, success, _ in results if success)
    total = len(results)

    print(f"\n📊 Migration Summary:")
    print(f"   Total files processed: {total}")
    print(f"   Successfully migrated: {successful}")
    print(f"   Failed migrations: {total - successful}")
    print(f"   Success rate: {(successful/total)*100:.1f}%")

    print(f"\n📋 Detailed Results:")
    print("-" * 60)

    for file_path, success, message in results:
        status = "✅" if success else "❌"
        try:
            print(f"{status} {file_path.relative_to(Path.cwd())}")
        except ValueError:
            print(f"{status} {file_path}")
        print(f"   {message}")

        # Show recommendations if available
        recommendations = generate_usage_recommendations(file_path)
        if recommendations:
            print("   💡 Recommendations:")
            for rec in recommendations:
                print(f"      • {rec}")
        print()

    print("🎯 Next Steps:")
    print("-" * 60)
    print("1. Test migrated files to ensure functionality is preserved")
    print("2. Consider using get_config('key.path', default) for better performance")
    print("3. Use helper functions like get_timeout_seconds() when appropriate")
    print("4. Monitor performance improvements in your Hook operations")

    print("\n" + "=" * 80)


def main():
    """Main migration function."""
    print("🔍 Scanning for load_config() functions...")

    # Find all files with load_config
    files_to_migrate = find_load_config_files()

    if not files_to_migrate:
        print("✅ No files found with load_config() functions requiring migration.")
        return

    print(f"📁 Found {len(files_to_migrate)} files to analyze...")

    # Migrate each file
    results = []
    for file_path in files_to_migrate:
        try:
            print(f"   Processing: {file_path.relative_to(Path.cwd())}")
        except ValueError:
            print(f"   Processing: {file_path}")
        success, message = migrate_file(file_path)
        results.append((file_path, success, message))

    # Create migration report
    create_migration_report(results)


if __name__ == "__main__":
    main()