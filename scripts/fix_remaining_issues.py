#!/usr/bin/env python3
"""GREEN PHASE: Fix remaining issues"""

import yaml
from pathlib import Path

SKILLS_DIR = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")


def fix_metadata_name_mismatches():
    """Fix metadata name fields that don't match directory."""
    fixes = {
        "moai-core-feedback-templates": "moai-core-feedback-templates",
        "moai-webapp-testing": "moai-webapp-testing",
    }

    for skill_dir_name, correct_name in fixes.items():
        skill_dir = SKILLS_DIR / skill_dir_name
        if skill_dir.exists():
            skill_md = skill_dir / "SKILL.md"
            if skill_md.exists():
                content = skill_md.read_text()
                if content.startswith("---"):
                    end = content.find("---", 3)
                    if end != -1:
                        frontmatter_str = content[3:end].strip()
                        metadata = yaml.safe_load(frontmatter_str) or {}
                        metadata["name"] = correct_name

                        # Reconstruct
                        yaml_lines = ["---"]
                        for key in [
                            "name",
                            "description",
                            "version",
                            "modularized",
                            "allowed-tools",
                            "last_updated",
                            "compliance_score",
                            "auto_trigger_keywords",
                            "category_tier",
                            "agent_coverage",
                        ]:
                            if key in metadata and metadata[key] is not None:
                                if isinstance(metadata[key], list):
                                    yaml_lines.append(f"{key}:")
                                    for item in metadata[key]:
                                        yaml_lines.append(f"  - {item}")
                                elif isinstance(metadata[key], bool):
                                    yaml_lines.append(
                                        f"{key}: {str(metadata[key]).lower()}"
                                    )
                                else:
                                    yaml_lines.append(f"{key}: {metadata[key]}")
                        yaml_lines.append("---")
                        new_content = "\n".join(yaml_lines) + content[end + 3 :]
                        skill_md.write_text(new_content)
                        print(f"✓ Fixed name mismatch: {skill_dir_name}")


def create_google_nano_banana():
    """Create moai-google-nano-banana skill."""
    skill_dir = SKILLS_DIR / "moai-google-nano-banana"
    if not skill_dir.exists():
        skill_dir.mkdir(parents=True, exist_ok=True)

        # Create SKILL.md
        content = """---
name: moai-google-nano-banana
description: Google Nano model integration for on-device AI and edge computing applications
version: 1.0.0
modularized: false
last_updated: 2025-11-22
compliance_score: 75
auto_trigger_keywords:
  - nano
  - google
  - edge
  - mobile
  - tflite
category_tier: 1
agent_coverage: []
---

# Google Nano Model Integration

[Quick reference for Google Nano model integration and usage patterns]

## When to Use

[Guide for when to use Google Nano models vs other options]
"""
        (skill_dir / "SKILL.md").write_text(content)
        print(f"✓ Created new skill: moai-google-nano-banana")


def main():
    print("Fixing remaining issues...")
    fix_metadata_name_mismatches()
    create_google_nano_banana()
    print("✓ Fixes applied!")


if __name__ == "__main__":
    main()
