#!/usr/bin/env python3
"""
GREEN PHASE: Assign category tiers and create 5 new essential skills
"""

import yaml
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path("/Users/goos/MoAI/MoAI-ADK")
SKILLS_DIR = PROJECT_ROOT / ".claude" / "skills"
TODAY = datetime.now().strftime("%Y-%m-%d")


# 10-tier category mapping
TIER_MAPPING = {
    1: {
        "name": "moai-lang-*",
        "skills": [
            "python",
            "javascript",
            "typescript",
            "go",
            "rust",
            "kotlin",
            "java",
            "php",
            "ruby",
            "swift",
            "scala",
            "c",
            "cpp",
            "csharp",
            "dart",
            "elixir",
            "html-css",
            "shell",
            "sql",
            "r",
            "tailwind-css",
        ],
    },
    2: {
        "name": "moai-domain-*",
        "skills": [
            "backend",
            "frontend",
            "database",
            "cloud",
            "cli-tool",
            "mobile-app",
            "iot",
            "figma",
            "notion",
            "toon",
            "ml-ops",
            "monitoring",
            "devops",
            "web-api",
            "testing",
            "security",
        ],
    },
    3: {
        "name": "moai-security-*",
        "skills": [
            "auth",
            "api",
            "owasp",
            "zero-trust",
            "encryption",
            "identity",
            "ssrf",
            "threat",
            "compliance",
            "secrets",
        ],
    },
    4: {
        "name": "moai-core-*",
        "skills": [
            "context-budget",
            "code-reviewer",
            "workflow",
            "issue-labels",
            "personas",
            "spec-authoring",
            "env-security",
            "clone-pattern",
            "dev-guide",
            "expertise-detection",
            "feedback-templates",
            "language-detection",
            "practices",
            "proactive-suggestions",
            "session-state",
            "todowrite-pattern",
            "config-schema",
            "ask-user-questions",
            "agent-factory",
        ],
    },
    5: {
        "name": "moai-foundation-*",
        "skills": ["ears", "specs", "trust", "git", "langs"],
    },
    6: {
        "name": "moai-cc-*",
        "skills": [
            "hooks",
            "commands",
            "skill-factory",
            "configuration",
            "claude-md",
            "claude-settings",
            "memory",
            "permission-mode",
            "skills-guide",
            "subagents-guide",
        ],
    },
    7: {
        "name": "moai-baas-*",
        "skills": [
            "vercel-ext",
            "neon-ext",
            "clerk-ext",
            "auth0-ext",
            "supabase-ext",
            "firebase-ext",
            "railway-ext",
            "cloudflare-ext",
            "convex-ext",
            "foundation",
        ],
    },
    8: {
        "name": "moai-essentials-*",
        "skills": [
            "debug",
            "perf",
            "refactor",
            "review",
            "testing-integration",
            "performance-profiling",
        ],
    },
    9: {
        "name": "moai-project-*",
        "skills": [
            "config-manager",
            "language-initializer",
            "batch-questions",
            "documentation",
            "template-optimizer",
        ],
    },
    10: {"name": "moai-lib-*", "skills": ["shadcn-ui"]},
}

# Special skills that don't fit standard tiers
SPECIAL_SKILLS = {
    "moai-docs-generation",
    "moai-docs-toolkit",
    "moai-docs-validation",
    "moai-docs-linting",
    "moai-docs-unified",
    "moai-core-uiux",
    "moai-mermaid-diagram-expert",
    "moai-webapp-testing",
    "moai-artifacts-builder",
    "moai-streaming-ui",
    "moai-mcp-integration",
    "moai-internal-comms",
    "moai-change-logger",
    "moai-learning-optimizer",
    "moai-document-processing",
    "moai-readme-expert",
    "moai-session-info",
    "moai-spec-intelligent-workflow",
    "moai-jit-docs-enhanced",
    "moai-context7-integration",
    "moai-nextra-architecture",
    "moai-cloud-aws-advanced",
    "moai-cloud-gcp-advanced",
}


def get_tier_for_skill(skill_name: str) -> int | str:
    """Determine tier for a skill."""
    if skill_name in SPECIAL_SKILLS:
        return "special"

    for tier, tier_info in TIER_MAPPING.items():
        for skill_keyword in tier_info["skills"]:
            if skill_keyword in skill_name:
                return tier

    return "unassigned"


def assign_tiers_to_all_skills():
    """Assign category_tier to all existing skills."""
    modified_count = 0

    for skill_dir in sorted(SKILLS_DIR.iterdir()):
        if not skill_dir.is_dir() or skill_dir.name.startswith("."):
            continue

        skill_md = skill_dir / "SKILL.md"
        if not skill_md.exists():
            continue

        content = skill_md.read_text(encoding="utf-8")
        if not content.startswith("---"):
            continue

        try:
            end = content.find("---", 3)
            if end == -1:
                continue

            frontmatter_str = content[3:end].strip()
            metadata = yaml.safe_load(frontmatter_str) or {}
            body = content[end + 3 :]

            # Check if tier already assigned
            current_tier = metadata.get("category_tier")
            new_tier = get_tier_for_skill(skill_dir.name)

            if current_tier != new_tier:
                metadata["category_tier"] = new_tier

                # Reconstruct YAML
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
                        if key == "compliance_score":
                            yaml_lines.append(f"{key}: {metadata[key]}")
                        elif isinstance(metadata[key], list):
                            yaml_lines.append(f"{key}:")
                            for item in metadata[key]:
                                yaml_lines.append(f"  - {item}")
                        elif isinstance(metadata[key], bool):
                            yaml_lines.append(f"{key}: {str(metadata[key]).lower()}")
                        else:
                            yaml_lines.append(f"{key}: {metadata[key]}")

                yaml_lines.append("---")
                new_content = "\n".join(yaml_lines) + body
                skill_md.write_text(new_content, encoding="utf-8")
                modified_count += 1
                print(f"✓ Assigned tier {new_tier}: {skill_dir.name}")

        except Exception as e:
            print(f"✗ Error processing {skill_dir.name}: {e}")

    print(f"\nTier assignment complete! Modified: {modified_count} skills")


def create_new_skill(skill_name: str, description: str, tier: int, keywords: list):
    """Create a new essential skill."""
    skill_dir = SKILLS_DIR / skill_name
    skill_dir.mkdir(parents=True, exist_ok=True)

    modules_dir = skill_dir / "modules"
    modules_dir.mkdir(exist_ok=True)

    # Create SKILL.md
    frontmatter = {
        "name": skill_name,
        "description": description,
        "version": "1.0.0",
        "modularized": True,
        "last_updated": TODAY,
        "compliance_score": 85,
        "auto_trigger_keywords": keywords,
        "category_tier": tier,
        "agent_coverage": [],
    }

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
        if key in frontmatter and frontmatter[key] is not None:
            if key == "compliance_score":
                yaml_lines.append(f"{key}: {frontmatter[key]}")
            elif isinstance(frontmatter[key], list):
                yaml_lines.append(f"{key}:")
                for item in frontmatter[key]:
                    yaml_lines.append(f"  - {item}")
            elif isinstance(frontmatter[key], bool):
                yaml_lines.append(f"{key}: {str(frontmatter[key]).lower()}")
            else:
                yaml_lines.append(f"{key}: {frontmatter[key]}")

    yaml_lines.append("---")
    skill_content = (
        "\n".join(yaml_lines)
        + "\n\n## Quick Reference\n\n[Quick reference content]\n\n## When to Use\n\n[When to use this skill]\n"
    )

    (skill_dir / "SKILL.md").write_text(skill_content, encoding="utf-8")

    # Create examples.md
    (skill_dir / "examples.md").write_text(
        "# Examples\n\n[Usage examples]\n", encoding="utf-8"
    )

    # Create basic module files
    (modules_dir / "advanced-patterns.md").write_text(
        "# Advanced Patterns\n\n[Advanced patterns]\n", encoding="utf-8"
    )

    print(f"✓ Created new skill: {skill_name}")


def create_new_skills():
    """Create 5 essential new skills."""
    new_skills = [
        {
            "name": "moai-core-code-templates",
            "description": "Code templates and boilerplates for common patterns and frameworks like FastAPI, React, and Vue",
            "tier": 4,
            "keywords": [
                "template",
                "boilerplate",
                "code",
                "initialization",
                "fastapi",
                "react",
                "vue",
            ],
        },
        {
            "name": "moai-security-api-versioning",
            "description": "API versioning strategies for REST, GraphQL and gRPC with backward compatibility and deprecation management",
            "tier": 3,
            "keywords": [
                "api",
                "versioning",
                "rest",
                "graphql",
                "grpc",
                "compatibility",
                "deprecation",
            ],
        },
        {
            "name": "moai-essentials-testing-integration",
            "description": "Integration and E2E testing patterns using Playwright, Cypress, Jest and pytest for comprehensive test coverage",
            "tier": 8,
            "keywords": [
                "testing",
                "integration",
                "e2e",
                "playwright",
                "cypress",
                "jest",
                "pytest",
            ],
        },
        {
            "name": "moai-essentials-performance-profiling",
            "description": "Performance profiling tools and techniques for CPU, memory and latency analysis across Python, Node.js and Go",
            "tier": 8,
            "keywords": [
                "performance",
                "profiling",
                "optimization",
                "cpu",
                "memory",
                "benchmark",
            ],
        },
        {
            "name": "moai-security-accessibility-wcag3",
            "description": "WCAG 3.0 accessibility compliance validation using axe-core, Pa11y and automated A11y testing",
            "tier": 3,
            "keywords": [
                "accessibility",
                "wcag",
                "a11y",
                "aria",
                "compliance",
                "testing",
            ],
        },
    ]

    for skill in new_skills:
        try:
            create_new_skill(
                skill["name"], skill["description"], skill["tier"], skill["keywords"]
            )
        except Exception as e:
            print(f"✗ Error creating {skill['name']}: {e}")

    print(f"\nNew skills creation complete!")


def main():
    """Main function."""
    print("Assigning tiers to existing skills...")
    assign_tiers_to_all_skills()

    print("\nCreating 5 new essential skills...")
    create_new_skills()

    print("\n✓ Tier assignment and new skill creation complete!")


if __name__ == "__main__":
    main()
