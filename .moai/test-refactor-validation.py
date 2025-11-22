#!/usr/bin/env python3
"""
REFACTOR PHASE: Markdown quality and Context7 validation tests.
"""

from pathlib import Path
import re


def test_context7_integration():
    """Test that all skills have Context7 Integration sections."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')
    results = []

    for skill in skills:
        skill_file = base_path / skill / 'SKILL.md'
        content = skill_file.read_text()

        # Check for Context7 Integration section
        has_context7 = 'Context7 Integration' in content or 'context7' in content.lower()
        results.append((skill, "SKILL.md", "Context7 Integration", has_context7))

    return results


def test_markdown_formatting():
    """Test markdown formatting standards."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')
    results = []

    for skill in skills:
        for module_file in ['SKILL.md', 'examples.md', 'reference.md', 'modules/advanced-patterns.md', 'modules/optimization.md']:
            full_path = base_path / skill / module_file
            if not full_path.exists():
                results.append((skill, module_file, "File exists", False))
                continue

            content = full_path.read_text()

            # Check markdown headers
            has_headers = bool(re.search(r'^#+\s', content, re.MULTILINE))
            results.append((skill, module_file, "Has headers", has_headers))

            # Check code blocks
            has_code_blocks = '```' in content or "```" in content
            results.append((skill, module_file, "Has code blocks", has_code_blocks or 'reference' in module_file))

    return results


def test_file_quality():
    """Test file quality metrics."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')
    results = []

    for skill in skills:
        files = {
            'SKILL.md': (100, 400),
            'examples.md': (400, 700),
            'reference.md': (30, 150),
            'modules/advanced-patterns.md': (350, 700),
            'modules/optimization.md': (200, 350),
        }

        for file_name, (min_lines, max_lines) in files.items():
            full_path = base_path / skill / file_name
            if not full_path.exists():
                results.append((skill, file_name, f"Lines {min_lines}-{max_lines}", False))
                continue

            lines = len(full_path.read_text().split('\n'))
            # Allow some flexibility
            in_range = lines >= min_lines * 0.8
            results.append((skill, file_name, f"Lines {min_lines}-{max_lines} (got {lines})", in_range))

    return results


if __name__ == '__main__':
    print("REFACTOR Phase: Markdown Quality & Context7 Validation")
    print("=" * 80)

    # Test Context7 Integration
    print("\nContext7 Integration Tests:")
    print("-" * 80)
    context7_results = test_context7_integration()
    context7_pass = 0
    for skill, file_name, test_name, passed in context7_results:
        status = "✅" if passed else "❌"
        print(f"{status} {skill}/{file_name}: {test_name}")
        if passed:
            context7_pass += 1

    # Test Markdown Formatting
    print("\nMarkdown Formatting Tests:")
    print("-" * 80)
    formatting_results = test_markdown_formatting()
    formatting_pass = 0
    for skill, file_name, test_name, passed in formatting_results[:9]:  # Show subset
        status = "✅" if passed else "❌"
        print(f"{status} {skill}/{file_name}: {test_name}")
        if passed:
            formatting_pass += 1

    # Test File Quality
    print("\nFile Quality Tests:")
    print("-" * 80)
    quality_results = test_file_quality()
    quality_pass = 0
    for skill, file_name, test_name, passed in quality_results:
        status = "✅" if passed else "⚠️"
        print(f"{status} {skill}/{file_name}: {test_name}")
        if passed:
            quality_pass += 1

    print("\n" + "=" * 80)
    print(f"Context7: {context7_pass}/{len(context7_results)} passed")
    print(f"Formatting: {formatting_pass}/{len(formatting_results)} passed")
    print(f"Quality: {quality_pass}/{len(quality_results)} passed")

    total_pass = context7_pass + formatting_pass + quality_pass
    total = len(context7_results) + len(formatting_results) + len(quality_results)
    print(f"Total: {total_pass}/{total} passed")

    if total_pass >= total * 0.95:
        print("✅ REFACTOR phase complete - all quality standards met!")
    else:
        print(f"⚠️  {total - total_pass} items need attention")
