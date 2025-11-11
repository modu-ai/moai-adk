#!/usr/bin/env python3
"""
Track v4.0 Upgrade Progress

Shows current progress, remaining work, and version distribution.

Usage:
    python3 scripts/track-upgrade-progress.py
    python3 scripts/track-upgrade-progress.py --detailed
"""

import re
from pathlib import Path
from collections import defaultdict
from typing import Dict, List


def analyze_progress(skills_dir=".claude/skills", detailed=False):
    """Analyze upgrade progress"""
    stats = defaultdict(int)
    skills_by_version = defaultdict(list)
    skill_details = []
    
    for skill_dir in Path(skills_dir).iterdir():
        if not skill_dir.is_dir() or skill_dir.name.startswith('.'):
            continue
        
        skill_md = skill_dir / "SKILL.md"
        if not skill_md.exists():
            # Python files
            if skill_dir.suffix == '.py':
                stats['python-file'] += 1
                skills_by_version['python-file'].append(skill_dir.name)
            continue
        
        try:
            content = skill_md.read_text(encoding='utf-8')
            
            # Extract version
            version_match = re.search(r'^version:\s*["\']?([0-9.]+)["\']?', content, re.MULTILINE)
            version = version_match.group(1) if version_match else "unknown"
            
            # Count examples
            code_blocks = len(re.findall(r'```', content))
            examples = code_blocks // 2
            
            # File size
            size_kb = skill_md.stat().st_size / 1024
            
            # Categorize
            if version.startswith('4.'):
                category = 'v4.0'
            elif version.startswith('3.'):
                category = 'v3.x'
            elif version.startswith('2.'):
                category = 'v2.0'
            elif version.startswith('1.'):
                category = 'v1.0'
            else:
                category = 'unknown'
            
            stats[category] += 1
            skills_by_version[category].append(skill_dir.name)
            
            skill_details.append({
                'name': skill_dir.name,
                'version': version,
                'category': category,
                'examples': examples,
                'size_kb': size_kb
            })
            
        except Exception as e:
            stats['error'] += 1
            skills_by_version['error'].append(f"{skill_dir.name} (error: {e})")
    
    # Calculate totals
    total = sum(stats.values())
    v4_count = stats['v4.0']
    remaining = total - v4_count - stats.get('python-file', 0)
    progress = (v4_count / total * 100) if total > 0 else 0
    
    # Print dashboard
    print()
    print("=" * 80)
    print("📊 v4.0 Upgrade Progress Dashboard")
    print("=" * 80)
    print()
    print(f"Total Skills: {total}")
    print(f"✅ v4.0 Complete: {v4_count} ({progress:.1f}%)")
    print(f"🔴 Remaining: {remaining}")
    print(f"⚠️  Python Files: {stats.get('python-file', 0)}")
    print()
    
    # Progress bar
    bar_length = 50
    filled = int(progress / 100 * bar_length)
    bar = "█" * filled + "░" * (bar_length - filled)
    print(f"Progress: [{bar}] {progress:.1f}%")
    print()
    
    # Version distribution
    print("Version Distribution:")
    print("-" * 80)
    
    for version in ['v4.0', 'v3.x', 'v2.0', 'v1.0', 'unknown', 'python-file']:
        count = stats.get(version, 0)
        if count > 0:
            pct = (count / total * 100)
            bar_width = int(pct / 2)
            bar = "█" * bar_width
            print(f"  {version:12} {count:3} skills {bar:25} {pct:5.1f}%")
    
    print()
    print("=" * 80)
    
    # Show remaining skills by category
    if remaining > 0:
        print()
        print("🔴 Skills Remaining for v4.0 Upgrade")
        print("=" * 80)
        
        for version in ['unknown', 'v3.x', 'v2.0', 'v1.0']:
            skills = skills_by_version.get(version, [])
            if skills:
                print(f"\n{version} ({len(skills)} skills):")
                for skill in sorted(skills):
                    print(f"  - {skill}")
        
        # Python files
        py_files = skills_by_version.get('python-file', [])
        if py_files:
            print(f"\nPython Files ({len(py_files)} files):")
            for f in sorted(py_files):
                print(f"  - {f}")
    
    # Detailed breakdown
    if detailed:
        print()
        print("=" * 80)
        print("📋 Detailed Skill Breakdown")
        print("=" * 80)
        
        # Group by category
        for category in ['v4.0', 'v3.x', 'v2.0', 'v1.0', 'unknown']:
            category_skills = [s for s in skill_details if s['category'] == category]
            if not category_skills:
                continue
            
            print(f"\n{category} ({len(category_skills)} skills):")
            print("-" * 80)
            print(f"{'Skill Name':<45} {'Examples':>8} {'Size':>8}")
            print("-" * 80)
            
            for skill in sorted(category_skills, key=lambda x: x['name']):
                print(f"{skill['name']:<45} {skill['examples']:>8} {skill['size_kb']:>7.1f}K")
            
            # Stats
            total_examples = sum(s['examples'] for s in category_skills)
            avg_examples = total_examples / len(category_skills)
            total_size = sum(s['size_kb'] for s in category_skills)
            avg_size = total_size / len(category_skills)
            
            print("-" * 80)
            print(f"{'Average:':<45} {avg_examples:>8.1f} {avg_size:>7.1f}K")
    
    print()
    print("=" * 80)
    
    # Phase breakdown
    print()
    print("📅 Phase Progress")
    print("=" * 80)
    
    phases = {
        "Phase 1": {
            "target": 21,
            "skills": (
                skills_by_version.get('unknown', [])[:16] +
                ['moai-alfred-agent-guide', 'moai-alfred-workflow',
                 'moai-alfred-context-budget', 'moai-alfred-personas',
                 'moai-alfred-todowrite-pattern']
            )
        },
        "Phase 2": {
            "target": 16,
            "skills": [
                s for s in skills_by_version.get('v2.0', [])
                if s not in ['moai-alfred-agent-guide', 'moai-alfred-workflow',
                             'moai-alfred-context-budget', 'moai-alfred-personas',
                             'moai-alfred-todowrite-pattern']
            ][:16]
        },
        "Phase 3": {
            "target": 9,
            "skills": skills_by_version.get('v1.0', [])[:9]
        },
        "Phase 4": {
            "target": 8,
            "skills": skills_by_version.get('v1.0', [])[9:] + skills_by_version.get('python-file', [])
        }
    }
    
    for phase, data in phases.items():
        target = data['target']
        skills = data['skills']
        
        # Check how many are v4.0
        completed = sum(1 for s in skills if s in skills_by_version.get('v4.0', []))
        remaining_phase = target - completed
        phase_progress = (completed / target * 100) if target > 0 else 0
        
        status = "✅" if remaining_phase == 0 else "🔴"
        
        print(f"\n{status} {phase}: {completed}/{target} ({phase_progress:.0f}%)")
        
        if remaining_phase > 0:
            print(f"   Remaining: {remaining_phase} skills")
    
    print()
    print("=" * 80)
    
    return stats, skill_details


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Track v4.0 upgrade progress")
    parser.add_argument("--skills-dir", default=".claude/skills", help="Skills directory")
    parser.add_argument("--detailed", action="store_true", help="Show detailed breakdown")
    
    args = parser.parse_args()
    
    analyze_progress(args.skills_dir, args.detailed)


if __name__ == "__main__":
    main()
