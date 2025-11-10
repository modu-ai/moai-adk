#!/usr/bin/env python3
"""
Validate the moai-project-config-manager skill structure and metadata
Enhanced with research capabilities for configuration optimization
"""

from pathlib import Path
import json
import re

import yaml


def analyze_configuration_research(config):
    """Analyze configuration changes using research strategies"""
    research_insights = []

    # Language configuration analysis
    if 'language' in config:
        lang = config['language']
        if 'conversation_language' in lang:
            research_insights.append({
                'category': 'LANGUAGE_OPTIMIZATION',
                'analysis': f'Language configuration: {lang["conversation_language"]}',
                'recommendations': [
                    'Consider multilingual support for global teams',
                    'Evaluate agent prompt language performance impact',
                    'Monitor conversation language switching patterns'
                ]
            })

    # GitHub workflow analysis
    if 'github' in config:
        github = config['github']
        if 'spec_git_workflow' in github:
            workflow = github['spec_git_workflow']
            research_insights.append({
                'category': 'WORKFLOW_OPTIMIZATION',
                'analysis': f'Git workflow: {workflow}',
                'recommendations': [
                    'Feature branch workflow reduces merge conflicts',
                    'Direct commit speeds up development but increases risk',
                    'Per-spec workflow provides flexibility but adds complexity'
                ]
            })

    # Report generation analysis
    if 'report_generation' in config:
        report = config['report_generation']
        if 'enabled' in report:
            research_insights.append({
                'category': 'REPORTING_OPTIMIZATION',
                'analysis': f'Report generation: {"enabled" if report["enabled"] else "disabled"}',
                'recommendations': [
                    'Enable reports for team transparency',
                    'Minimal reports save tokens while maintaining visibility',
                    'Disable only for performance-critical workflows'
                ]
            })

    return research_insights


def validate_skill():
    """Validate skill structure and metadata with research integration"""

    print("🔍 Validating moai-project-config-manager skill...")
    print("🔬 Conducting configuration optimization research...")

    skill_dir = Path(__file__).parent
    skill_file = skill_dir / "SKILL.md"

    if not skill_file.exists():
        print("❌ SKILL.md not found")
        return False

    # Read and parse frontmatter
    with open(skill_file, 'r') as f:
        content = f.read()

    # Extract frontmatter
    if content.startswith('---'):
        try:
            end_index = content.find('---', 3)
            frontmatter_str = content[3:end_index].strip()
            frontmatter = yaml.safe_load(frontmatter_str)

            print("✅ Frontmatter parsed successfully")

            # Validate required fields
            required_fields = ['name', 'version', 'description', 'freedom', 'type', 'tags']
            missing_fields = [field for field in required_fields if field not in frontmatter]

            if missing_fields:
                print(f"❌ Missing required fields: {missing_fields}")
                return False

            print("✅ All required fields present")

            # Validate field values
            if frontmatter['name'] != 'moai-project-config-manager':
                print(f"❌ Incorrect skill name: {frontmatter['name']}")
                return False

            if frontmatter['type'] != 'project':
                print(f"❌ Incorrect skill type: {frontmatter['type']}")
                return False

            if frontmatter['freedom'] not in ['low', 'medium', 'high']:
                print(f"❌ Invalid freedom level: {frontmatter['freedom']}")
                return False

            if not isinstance(frontmatter['tags'], list):
                print(f"❌ Tags must be a list: {frontmatter['tags']}")
                return False

            print("✅ Field values validated")

            # Check expected tags
            expected_tags = ['project', 'configuration', 'management']
            has_expected_tags = any(tag in frontmatter['tags'] for tag in expected_tags)

            if not has_expected_tags:
                print(f"⚠️ Missing expected tags: {expected_tags}")
            else:
                print("✅ Expected tags present")

        except yaml.YAMLError as e:
            print(f"❌ Failed to parse frontmatter: {e}")
            return False

    # Check for required files
    required_files = ['SKILL.md', 'reference.md', 'examples.md']
    missing_files = [f for f in required_files if not (skill_dir / f).exists()]

    if missing_files:
        print(f"❌ Missing required files: {missing_files}")
        return False

    print("✅ All required files present")

    # Validate file sizes
    for file_name in required_files:
        file_path = skill_dir / file_name
        size_kb = file_path.stat().st_size / 1024

        if file_name == 'SKILL.md' and size_kb < 10:
            print(f"⚠️ {file_name} seems small: {size_kb:.1f}KB")
        elif file_name == 'reference.md' and size_kb < 5:
            print(f"⚠️ {file_name} seems small: {size_kb:.1f}KB")
        elif file_name == 'examples.md' and size_kb < 5:
            print(f"⚠️ {file_name} seems small: {size_kb:.1f}KB")
        else:
            print(f"✅ {file_name}: {size_kb:.1f}KB")

    # Load and analyze actual configuration for research insights
    config_path = Path(__file__).parent / "../../../.moai/config.json"
    if config_path.exists():
        try:
            with open(config_path, 'r') as f:
                config = json.load(f)

            research_insights = analyze_configuration_research(config)

            if research_insights:
                print("\n🔬 Research Analysis Results:")
                for insight in research_insights:
                    print(f"\n📊 {insight['category']}:")
                    print(f"   Analysis: {insight['analysis']}")
                    print("   Recommendations:")
                    for rec in insight['recommendations']:
                        print(f"   • {rec}")
            else:
                print("\n🔬 No configuration insights found")

        except Exception as e:
            print(f"\n⚠️ Research analysis failed: {e}")
    else:
        print("\n⚠️ Configuration file not found for research analysis")

    print("\n🎉 Skill validation completed successfully!")
    return True

if __name__ == "__main__":
    validate_skill()
