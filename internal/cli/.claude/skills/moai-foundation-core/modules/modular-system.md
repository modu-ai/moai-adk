# Modular System - File Organization

Purpose: Scalable file structure enabling unlimited content depth while maintaining clean, navigable, and maintainable skill architecture.

Version: 1.1.0
Last Updated: 2026-01-21

---

## Quick Reference (30 seconds)

Modular System = Organized file structure for scalable skills and documentation.

Standard Structure:
```
.claude/skills/skill-name/
 SKILL.md # Core entry (≤500 lines, mandatory)
 modules/ # Extended content (unlimited, optional)
 pattern-a.md
 pattern-b.md
 pattern-c.md
 examples.md # Working code samples (optional)
 reference.md # External links/API docs (optional)
 scripts/ # Utility scripts (optional)
 helper.sh
 templates/ # Templates (optional)
 template.md
```

File Principles:
1. SKILL.md ≤500 lines (hard limit)
2. modules/ = Topic-focused, self-contained
3. examples.md = Copy-paste ready
4. reference.md = External resources

Cross-Reference Syntax:
```markdown
Details: [Module](modules/patterns.md)
Examples: [Examples](examples.md#auth)
External: [Reference](reference.md#api)
```

---

## Implementation Guide (5 minutes)

### Standard File Structure

Tier 1: Mandatory Files

.claude/skills/skill-name/SKILL.md (Required, ≤500 lines):

```markdown
---
name: skill-name
description: Brief description (max 1024 chars)
tools: Read, Bash, Grep, Glob
---

# Skill Name

Entry point with progressive disclosure structure.

## Quick Reference (30s)
[Core principles]

## Implementation Guide (5min)
[Practical patterns with cross-references]

## Advanced Patterns (10+min)
[Brief intros with links to modules/]

## Works Well With
[Integration points]
```

Tier 2: Extended Content (Optional)

modules/ Directory:
- Purpose: Deep dives on specific topics
- Structure: Self-contained, topic-focused
- Limits: No line limits (can be 1000+ lines)
- Naming: Descriptive kebab-case (e.g., `advanced-patterns.md`)

Example modules/ structure:
```
modules/
 trust-5-framework.md # Quality assurance deep dive
 spec-first-ddd.md # DDD workflow detailed guide
 delegation-patterns.md # Agent orchestration patterns
 token-optimization.md # Budget management strategies
 progressive-disclosure.md # Content architecture
 modular-system.md # File organization (this file)
```

Tier 3: Supporting Files (Optional)

examples.md:
```markdown
# Working Examples

## Example 1: Basic Usage
```python
# Copy-paste ready code
def basic_example():
 result = skill_function()
 return result
```

## Example 2: Advanced Usage
```python
# Complex scenario
def advanced_example():
 # Detailed implementation
 pass
```
```

reference.md:
```markdown
# External Resources

## Official Documentation
- [Library Docs](https://docs.example.com)
- [API Reference](https://api.example.com)

## Related Standards
- [RFC 5322](https://tools.ietf.org/html/rfc5322)
- [OWASP Top 10](https://owasp.org/Top10/)

## Tools and Libraries
- [pytest](https://docs.pytest.org/)
- [black](https://black.readthedocs.io/)
```

scripts/ Directory:
```bash
# scripts/helper.sh
#!/bin/bash
# Utility script for skill operations

function validate_skill() {
 # Validation logic
 echo "Validating skill structure..."
}
```

templates/ Directory:
```markdown
<!-- templates/template.md -->
# Template Name

[Template content with placeholders]

Variables:
- {{skill_name}}
- {{description}}
- {{tools}}
```

---

### File Splitting Strategy

When to Split SKILL.md:

```python
class FileSplittingDecision:
 """Determine when and how to split SKILL.md."""
 
 MAX_SKILL_LINES = 500
 
 def should_split(self, skill_md_path: str) -> dict:
 """Analyze if SKILL.md needs splitting."""
 
 with open(skill_md_path) as f:
 lines = f.readlines()
 
 line_count = len(lines)
 
 if line_count <= self.MAX_SKILL_LINES:
 return {
 "split_needed": False,
 "line_count": line_count,
 "remaining": self.MAX_SKILL_LINES - line_count
 }
 
 # Analyze sections for splitting
 sections = self._analyze_sections(lines)
 
 split_recommendations = []
 
 # Advanced Patterns → modules/advanced-patterns.md
 if sections["advanced_patterns"] > 100:
 split_recommendations.append({
 "source": "Advanced Patterns section",
 "target": "modules/advanced-patterns.md",
 "lines": sections["advanced_patterns"],
 "keep_intro": 20 # Keep brief intro in SKILL.md
 })
 
 # Code Examples → examples.md
 if sections["code_examples"] > 80:
 split_recommendations.append({
 "source": "Code examples",
 "target": "examples.md",
 "lines": sections["code_examples"],
 "keep_intro": 10 # Keep key example in SKILL.md
 })
 
 # References → reference.md
 if sections["references"] > 50:
 split_recommendations.append({
 "source": "References section",
 "target": "reference.md",
 "lines": sections["references"],
 "keep_intro": 5
 })
 
 # Topic-specific deep dives → modules/[topic].md
 for topic, topic_lines in sections["topics"].items():
 if topic_lines > 150:
 split_recommendations.append({
 "source": f"{topic} section",
 "target": f"modules/{topic}.md",
 "lines": topic_lines,
 "keep_intro": 30
 })
 
 return {
 "split_needed": True,
 "line_count": line_count,
 "overflow": line_count - self.MAX_SKILL_LINES,
 "recommendations": split_recommendations,
 "estimated_final_size": line_count - sum(
 r["lines"] - r["keep_intro"] for r in split_recommendations
 )
 }
 
 def execute_split(self, skill_path: str, recommendations: list):
 """Execute file splitting based on recommendations."""
 
 skill_md_path = f"{skill_path}/SKILL.md"
 
 with open(skill_md_path) as f:
 content = f.read()
 
 for rec in recommendations:
 # Extract content for splitting
 section_content = self._extract_section_content(
 content,
 rec["source"]
 )
 
 # Create target file
 target_path = f"{skill_path}/{rec['target']}"
 os.makedirs(os.path.dirname(target_path), exist_ok=True)
 
 with open(target_path, 'w') as f:
 f.write(section_content)
 
 # Replace in SKILL.md with brief intro + cross-reference
 brief_intro = section_content[:rec["keep_intro"]] + "\n\n"
 cross_ref = f"[Full details]({rec['target']})\n"
 
 content = content.replace(
 section_content,
 brief_intro + cross_ref
 )
 
 # Write updated SKILL.md
 with open(skill_md_path, 'w') as f:
 f.write(content)

# Usage
splitter = FileSplittingDecision()

decision = splitter.should_split(".claude/skills/moai-foundation-core/SKILL.md")

if decision["split_needed"]:
 print(f" SKILL.md needs splitting: {decision['line_count']} lines")
 print(f" Overflow: {decision['overflow']} lines")
 
 for rec in decision["recommendations"]:
 print(f" → Split '{rec['source']}' to {rec['target']}")
 
 # Execute splitting
 splitter.execute_split(
 ".claude/skills/moai-foundation-core",
 decision["recommendations"]
 )
 
 print(f" Final SKILL.md size: {decision['estimated_final_size']} lines")
```

---

### Cross-Reference Patterns

Effective Cross-Linking Strategy:

Pattern 1: Module Cross-References:
```markdown
<!-- In SKILL.md -->
## Implementation Guide (5 minutes)

### Pattern 1: Quality Framework

Quick overview of TRUST 5 framework.

Detailed Implementation: [TRUST 5 Module](modules/trust-5-framework.md)

Key principles:
- Tested ≥85% (characterization tests for legacy, specification tests for new code)
- Readable code
- Unified patterns
- Secured (OWASP)
- Trackable commits

For advanced patterns, validation frameworks, and CI/CD integration, see the [full module](modules/trust-5-framework.md#advanced-implementation).
```

Pattern 2: Section Anchors:
```markdown
<!-- In SKILL.md -->
Quick Access:
- Quality framework → [TRUST 5](modules/trust-5-framework.md#quick-reference)
- DDD workflow → [SPEC-First](modules/spec-first-ddd.md#phase-2-ddd-implementation)
- Agent patterns → [Delegation](modules/delegation-patterns.md#pattern-1-sequential-delegation)

<!-- In module -->
## Advanced Implementation (10+ minutes)

### Pattern Optimization {#pattern-optimization}

[Content accessible via anchor link]
```

Pattern 3: Example Links:
```markdown
<!-- In SKILL.md -->
## Implementation Guide

```python
# Basic example
def basic_usage():
 result = process()
 return result
```

More Examples: [examples.md](examples.md)
- [Authentication Example](examples.md#auth-example)
- [API Integration](examples.md#api-integration)
- [Error Handling](examples.md#error-handling)
```

Pattern 4: External References:
```markdown
<!-- In SKILL.md -->
## Quick Reference

Official Documentation: [reference.md](reference.md#official-docs)
Related Standards: [reference.md](reference.md#standards)
Tools: [reference.md](reference.md#tools)
```

---

## Advanced Implementation (10+ minutes)

### Automated File Organization

Skill Organizer Tool:

```python
from pathlib import Path
import os
import re

class SkillOrganizer:
 """Organize skill files according to modular system standards."""
 
 def __init__(self, skill_path: Path):
 self.skill_path = Path(skill_path)
 self.structure = {
 "SKILL.md": True, # Mandatory
 "modules/": False,
 "examples.md": False,
 "reference.md": False,
 "scripts/": False,
 "templates/": False
 }
 
 def validate_structure(self) -> dict:
 """Validate skill directory structure."""
 
 validation = {}
 
 # Check mandatory files
 skill_md = self.skill_path / "SKILL.md"
 if not skill_md.exists():
 validation["SKILL.md"] = {
 "status": "MISSING",
 "severity": "ERROR",
 "action": "Create SKILL.md file"
 }
 else:
 # Validate SKILL.md content
 with open(skill_md) as f:
 lines = f.readlines()
 
 line_count = len(lines)
 
 validation["SKILL.md"] = {
 "status": "OK" if line_count <= 500 else "OVERFLOW",
 "severity": "WARNING" if line_count > 500 else "OK",
 "line_count": line_count,
 "action": "Split to modules" if line_count > 500 else None
 }
 
 # Check optional directories
 for dir_name in ["modules", "scripts", "templates"]:
 dir_path = self.skill_path / dir_name
 if dir_path.exists():
 validation[f"{dir_name}/"] = {
 "status": "PRESENT",
 "files": list(dir_path.glob("*"))
 }
 
 # Check optional files
 for file_name in ["examples.md", "reference.md"]:
 file_path = self.skill_path / file_name
 if file_path.exists():
 validation[file_name] = {
 "status": "PRESENT",
 "size": file_path.stat().st_size
 }
 
 return validation
 
 def organize_skill(self):
 """Organize skill files according to standards."""
 
 # Create modules/ directory if needed
 modules_dir = self.skill_path / "modules"
 if not modules_dir.exists():
 modules_dir.mkdir()
 
 # Move advanced content to modules/
 self._move_advanced_content_to_modules()
 
 # Extract examples to examples.md
 self._extract_examples()
 
 # Extract references to reference.md
 self._extract_references()
 
 # Create scripts/ if utility scripts exist
 self._organize_scripts()
 
 # Validate final structure
 return self.validate_structure()
 
 def _move_advanced_content_to_modules(self):
 """Move advanced patterns to modules/."""
 
 skill_md = self.skill_path / "SKILL.md"
 
 with open(skill_md) as f:
 content = f.read()
 
 # Extract Advanced Patterns section
 advanced_match = re.search(
 r'## Advanced (Implementation|Patterns).*?(?=##|$)',
 content,
 re.DOTALL
 )
 
 if advanced_match and len(advanced_match.group(0)) > 500:
 advanced_content = advanced_match.group(0)
 
 # Save to module
 module_path = self.skill_path / "modules" / "advanced-patterns.md"
 with open(module_path, 'w') as f:
 f.write(f"# Advanced Patterns\n\n{advanced_content}")
 
 # Replace with brief intro in SKILL.md
 brief_intro = advanced_content[:200] + "\n\n"
 cross_ref = "[Full advanced patterns](modules/advanced-patterns.md)\n"
 
 content = content.replace(
 advanced_content,
 brief_intro + cross_ref
 )
 
 # Write updated SKILL.md
 with open(skill_md, 'w') as f:
 f.write(content)
 
 def generate_navigation(self) -> str:
 """Generate navigation structure for skill."""
 
 navigation = []
 navigation.append("# Skill Navigation\n")
 
 # SKILL.md sections
 navigation.append("## Core Content (SKILL.md)\n")
 navigation.append("- [Quick Reference](SKILL.md#quick-reference)\n")
 navigation.append("- [Implementation Guide](SKILL.md#implementation-guide)\n")
 navigation.append("- [Advanced Patterns](SKILL.md#advanced-patterns)\n\n")
 
 # Modules
 modules_dir = self.skill_path / "modules"
 if modules_dir.exists():
 navigation.append("## Extended Content (modules/)\n")
 for module in sorted(modules_dir.glob("*.md")):
 module_name = module.stem.replace("-", " ").title()
 navigation.append(f"- [{module_name}](modules/{module.name})\n")
 navigation.append("\n")
 
 # Examples
 if (self.skill_path / "examples.md").exists():
 navigation.append("## Working Examples\n")
 navigation.append("- [examples.md](examples.md)\n\n")
 
 # Reference
 if (self.skill_path / "reference.md").exists():
 navigation.append("## External Resources\n")
 navigation.append("- [reference.md](reference.md)\n\n")
 
 return "".join(navigation)

# Usage
organizer = SkillOrganizer(".claude/skills/moai-foundation-core")

# Validate current structure
validation = organizer.validate_structure()
for file, result in validation.items():
 print(f"{file}: {result}")

# Organize skill files
organizer.organize_skill()

# Generate navigation
navigation = organizer.generate_navigation()
with open(".claude/skills/moai-foundation-core/NAVIGATION.md", 'w') as f:
 f.write(navigation)
```

### Module Discovery and Loading

Dynamic Module Loader:

```python
class ModuleDiscovery:
 """Discover and load skill modules dynamically."""
 
 def __init__(self, skill_path: Path):
 self.skill_path = Path(skill_path)
 self.modules_cache = {}
 
 def discover_modules(self) -> dict:
 """Discover all available modules."""
 
 modules_dir = self.skill_path / "modules"
 
 if not modules_dir.exists():
 return {}
 
 modules = {}
 
 for module_file in modules_dir.glob("*.md"):
 module_name = module_file.stem
 
 # Extract module metadata
 with open(module_file) as f:
 content = f.read()
 
 # Parse frontmatter if exists
 metadata = self._parse_frontmatter(content)
 
 modules[module_name] = {
 "path": module_file,
 "size": module_file.stat().st_size,
 "metadata": metadata,
 "topics": self._extract_topics(content)
 }
 
 return modules
 
 def load_module(self, module_name: str) -> str:
 """Load specific module content."""
 
 if module_name in self.modules_cache:
 return self.modules_cache[module_name]
 
 module_path = self.skill_path / "modules" / f"{module_name}.md"
 
 if not module_path.exists():
 raise FileNotFoundError(f"Module not found: {module_name}")
 
 with open(module_path) as f:
 content = f.read()
 
 self.modules_cache[module_name] = content
 return content
 
 def search_modules(self, query: str) -> list:
 """Search for topic across all modules."""
 
 modules = self.discover_modules()
 results = []
 
 for module_name, module_info in modules.items():
 if query.lower() in module_info["topics"]:
 results.append({
 "module": module_name,
 "path": module_info["path"],
 "relevance": self._calculate_relevance(query, module_info)
 })
 
 return sorted(results, key=lambda x: x["relevance"], reverse=True)

# Usage
discovery = ModuleDiscovery(".claude/skills/moai-foundation-core")

# Discover all modules
modules = discovery.discover_modules()
print(f"Found {len(modules)} modules")

# Load specific module
trust5_content = discovery.load_module("trust-5-framework")

# Search for topic
results = discovery.search_modules("security")
for result in results:
 print(f"Found in: {result['module']} (relevance: {result['relevance']})")
```

---

## Works Well With

Skills:
- moai-foundation-progressive-disclosure - Content structuring
- moai-cc-skill-factory - Skill creation with modular structure
- moai-foundation-token-optimization - File loading efficiency

Agents:
- skill-factory - Create skills with standard file structure
- docs-manager - Generate documentation following modular pattern

Commands:
- /moai:1-plan - SPEC generation with modular docs
- /moai:3-sync - Documentation sync to modular structure

Memory:
- @.claude/skills/ - Standard skill location
- Skill("moai-foundation-core") modules/ - Memory files following modular pattern

---

Version: 1.1.0
Last Updated: 2026-01-21
Status: Production Ready
