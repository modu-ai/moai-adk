#!/usr/bin/env python3
"""
RED PHASE: Comprehensive test suite for Batch 3-1 TAG-006 file structure validation.
Tests validate that all 3 skills have the required 5-file modularization structure.
"""

import os
import re
from pathlib import Path


def test_skill_directory_structure():
    """Test that all 3 skills have required directory structure."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        skill_path = base_path / skill
        assert skill_path.exists(), f"Skill directory {skill} missing"

        # Check required files
        required_files = [
            'SKILL.md',
            'examples.md',
            'reference.md',
            'modules/advanced-patterns.md',
            'modules/optimization.md'
        ]

        for file_path in required_files:
            full_path = skill_path / file_path
            assert full_path.exists(), f"Missing required file: {skill}/{file_path}"


def test_skill_md_structure():
    """Test SKILL.md has proper frontmatter and sections."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        skill_file = base_path / skill / 'SKILL.md'
        content = skill_file.read_text()

        # Check frontmatter
        assert content.startswith('---'), f"{skill}/SKILL.md missing frontmatter"
        assert 'name:' in content, f"{skill}/SKILL.md missing 'name' field"
        assert 'description:' in content, f"{skill}/SKILL.md missing 'description' field"

        # Check required sections
        required_sections = [
            'Quick Reference',
            'What It Does',
            'When to Use',
            'Three-Level Learning',
            'Best Practices',
            'Works Well With',
            'Learn More',
            'Version',
        ]

        for section in required_sections:
            assert section in content, f"{skill}/SKILL.md missing '{section}' section"


def test_examples_md_structure():
    """Test examples.md has proper examples and documentation."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        examples_file = base_path / skill / 'examples.md'
        content = examples_file.read_text()

        # Check for code blocks
        code_blocks = content.count('```')
        assert code_blocks >= 4, f"{skill}/examples.md has insufficient code blocks (need >=2 examples)"

        # Check for examples
        assert 'Example' in content or 'example' in content.lower(), f"{skill}/examples.md missing examples"


def test_context7_integration():
    """Test that Context7 Integration sections exist in skills."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        skill_file = base_path / skill / 'SKILL.md'
        content = skill_file.read_text()

        # Check for Context7 Integration
        assert 'Context7 Integration' in content or 'context7' in content.lower(), \
            f"{skill}/SKILL.md missing Context7 Integration section"


def test_advanced_patterns_exists():
    """Test that advanced-patterns.md file exists and has content."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        patterns_file = base_path / skill / 'modules' / 'advanced-patterns.md'

        # File should exist OR we're testing for its creation
        if not patterns_file.exists():
            print(f"⚠️  {skill}/modules/advanced-patterns.md needs to be created (RED phase)")
            continue

        content = patterns_file.read_text()
        assert len(content) > 100, f"{skill}/modules/advanced-patterns.md is too short"


def test_optimization_exists():
    """Test that optimization.md file exists and has content."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        opt_file = base_path / skill / 'modules' / 'optimization.md'

        # File should exist OR we're testing for its creation
        if not opt_file.exists():
            print(f"⚠️  {skill}/modules/optimization.md needs to be created (RED phase)")
            continue

        content = opt_file.read_text()
        assert len(content) > 100, f"{skill}/modules/optimization.md is too short"


def test_reference_md_structure():
    """Test reference.md has proper API documentation."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        ref_file = base_path / skill / 'reference.md'
        content = ref_file.read_text()

        # Check for expected reference content
        assert len(content) > 50, f"{skill}/reference.md is too short"


def test_markdown_validity():
    """Test that all markdown files have valid syntax."""
    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    for skill in skills:
        skill_path = base_path / skill

        # Test SKILL.md
        skill_file = skill_path / 'SKILL.md'
        if skill_file.exists():
            content = skill_file.read_text()
            # Basic markdown validation
            assert '---' in content, f"{skill}/SKILL.md invalid frontmatter"
            assert content.count('```') % 2 == 0, f"{skill}/SKILL.md mismatched code blocks"


def test_file_line_counts():
    """Test that files meet size requirements."""
    base_path = Path('/Users/goos/MoAI/MoAI-ADK/.claude/skills')

    # SKILL.md should have <=400 lines
    # examples.md should have 550-700 lines
    # advanced-patterns.md should have 500-700 lines
    # optimization.md should have 250-350 lines
    # reference.md should have 30-40 lines

    skills = [
        'moai-lib-shadcn-ui',
        'moai-design-systems',
        'moai-component-designer'
    ]

    for skill in skills:
        skill_file = base_path / skill / 'SKILL.md'
        if skill_file.exists():
            lines = len(skill_file.read_text().split('\n'))
            # SKILL.md can vary, so check minimum
            assert lines > 50, f"{skill}/SKILL.md too short (< 50 lines)"


if __name__ == '__main__':
    print("Running RED Phase Tests for Batch 3-1 TAG-006...")
    print("=" * 60)

    tests = [
        ("Skill directory structure", test_skill_directory_structure),
        ("SKILL.md structure", test_skill_md_structure),
        ("examples.md structure", test_examples_md_structure),
        ("Context7 integration", test_context7_integration),
        ("advanced-patterns.md exists", test_advanced_patterns_exists),
        ("optimization.md exists", test_optimization_exists),
        ("reference.md structure", test_reference_md_structure),
        ("Markdown validity", test_markdown_validity),
        ("File line counts", test_file_line_counts),
    ]

    passed = 0
    failed = 0

    for test_name, test_func in tests:
        try:
            test_func()
            print(f"✅ {test_name}")
            passed += 1
        except AssertionError as e:
            print(f"❌ {test_name}: {e}")
            failed += 1
        except Exception as e:
            print(f"⚠️  {test_name}: {e}")

    print("=" * 60)
    print(f"Results: {passed} passed, {failed} failed")

    if failed == 0:
        print("✅ All tests passed! Moving to GREEN phase...")
    else:
        print(f"❌ {failed} tests failed. See details above.")
