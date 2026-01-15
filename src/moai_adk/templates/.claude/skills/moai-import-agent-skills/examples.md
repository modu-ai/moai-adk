# Agent-Skills Import Examples

Working examples demonstrating agent-skills to MoAI-ADK conversion process.

## Example 1: React Best Practices Conversion

### Source Format Analysis

Original agent-skills structure:

```
react-best-practices/
├── SKILL.md
├── metadata.json
├── AGENTS.md
├── README.md
└── rules/
    ├── async-parallel.md
    ├── async-defer-await.md
    ├── bundle-barrel-imports.md
    └── [42 more rule files]
```

Source SKILL.md frontmatter:

```yaml
---
name: vercel-react-best-practices
description: React and Next.js performance optimization guidelines from Vercel Engineering. This skill should be used when writing, reviewing, or refactoring React/Next.js code to ensure optimal performance patterns. Triggers on tasks involving React components, Next.js pages, data fetching, bundle optimization, or performance improvements.
---
```

Source metadata.json:

```json
{
  "version": "0.1.0",
  "organization": "Vercel Engineering",
  "date": "January 2026",
  "abstract": "Comprehensive performance optimization guide for React and Next.js applications, designed for AI agents and LLMs. Contains 40+ rules across 8 categories, prioritized by impact from critical (eliminating waterfalls, reducing bundle size) to incremental (advanced patterns).",
  "references": [
    "https://react.dev",
    "https://nextjs.org",
    "https://swr.vercel.app"
  ]
}
```

### Conversion Process

Step 1: Determine new skill name

Original: vercel-react-best-practices
Converted: moai-domain-react (or moai-platform-nextjs)

Step 2: Convert description

Original: React and Next.js performance optimization guidelines from Vercel Engineering. This skill should be used when writing, reviewing, or refactoring React/Next.js code to ensure optimal performance patterns.

Converted: React and Next.js performance specialist covering waterfalls elimination, bundle optimization, server-side rendering, and client-side data fetching patterns. Use when optimizing React components, Next.js pages, data fetching, bundle size, or performance improvements.

Step 3: Determine tool permissions

Required tools for React performance specialist:

- Read: Analyze React component code
- Write: Refactor components for performance
- Edit: Make specific performance improvements
- Grep: Search for performance anti-patterns
- Glob: Find React component files
- mcp__context7__resolve-library-id: Access React documentation
- mcp__context7__get-library-docs: Fetch latest React patterns

Step 4: Create Quick Reference section

Extract 8 rule categories by priority:

1. Eliminating Waterfalls (CRITICAL): async-defer-await, async-parallel, async-dependencies
2. Bundle Size Optimization (CRITICAL): bundle-barrel-imports, bundle-dynamic-imports
3. Server-Side Performance (HIGH): server-cache-react, server-cache-lru
4. Client-Side Data Fetching (MEDIUM-HIGH): client-swr-dedup
5. Re-render Optimization (MEDIUM): rerender-memo, rerender-dependencies
6. Rendering Performance (MEDIUM): rendering-hoist-jsx, rendering-content-visibility
7. JavaScript Performance (LOW-MEDIUM): js-batch-dom-css, js-index-maps
8. Advanced Patterns (LOW): advanced-event-handler-refs, advanced-use-latest

Step 5: Create Implementation Guide section

Provide workflows for each rule category with code examples.

Example for async-parallel rule:

Incorrect pattern (sequential execution):

```typescript
// Three sequential round trips
const user = await fetchUser()
const posts = await fetchPosts()
const comments = await fetchComments()
```

Correct pattern (parallel execution):

```typescript
// Single parallel round trip
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
])
```

Step 6: Create Advanced Patterns section

Cover performance monitoring, profiling, and optimization strategies.

### Final MoAI-ADK Format

```yaml
---
name: moai-domain-react
description: React and Next.js performance specialist covering waterfalls elimination, bundle optimization, server-side rendering, and client-side data fetching patterns. Use when optimizing React components, Next.js pages, data fetching, bundle size, or performance improvements.
allowed-tools: Read, Write, Edit, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.0.0
status: active
updated: 2026-01-15
category: domain
author: MoAI-ADK Team
tags:
  - react
  - nextjs
  - performance
  - optimization
  - frontend
---

# React Performance Specialist

## Quick Reference

React and Next.js performance specialist with 45 rules across 8 prioritized categories.

Core Capabilities:

- Eliminating Waterfalls: Prevent sequential async operations with Promise.all() and Suspense
- Bundle Size Optimization: Reduce bundle size with dynamic imports and code splitting
- Server-Side Performance: Optimize server components with React.cache() and LRU caching
- Client-Side Data Fetching: Implement SWR for automatic request deduplication
- Re-render Optimization: Minimize re-renders with memo, derived state, and transitions
- Rendering Performance: Improve rendering with hoisting, content-visibility, and Activity

When to Use:

- Optimizing React components for performance
- Reducing Next.js bundle size and load times
- Implementing server-side rendering patterns
- Fixing data fetching waterfalls and race conditions
- Analyzing re-render performance issues

---

## Implementation Guide

### Eliminating Waterfalls

Prevent sequential async operations by executing independent operations in parallel.

Use Promise.all() for Independent Operations:

When async operations have no interdependencies, execute them concurrently using Promise.all().

Incorrect pattern (sequential execution):

```typescript
const user = await fetchUser()
const posts = await fetchPosts()
const comments = await fetchComments()
```

Correct pattern (parallel execution):

```typescript
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
])
```

Move Await into Branches:

Defer await until the async result is actually needed in the code path.

Incorrect pattern (awaiting too early):

```typescript
const user = await getUser(id)
if (shouldShowAdmin) {
  const admin = await getAdmin(user.adminId)
  return admin
}
return user
```

Correct pattern (awaiting in branch):

```typescript
const userPromise = getUser(id)
if (shouldShowAdmin) {
  const user = await userPromise
  const admin = await getAdmin(user.adminId)
  return admin
}
return await userPromise
```

### Bundle Size Optimization

Reduce bundle size by avoiding barrel files and using dynamic imports.

Import Directly from Source:

Avoid importing from barrel files that cause large bundle imports.

Incorrect pattern (barrel import):

```typescript
import { Button, Card, Modal } from '@my-library/components'
```

Correct pattern (direct import):

```typescript
import Button from '@my-library/components/Button'
import Card from '@my-library/components/Card'
import Modal from '@my-library/components/Modal'
```

Use Dynamic Imports for Heavy Components:

Load heavy components only when needed using next/dynamic.

Incorrect pattern (static import):

```typescript
import HeavyChart from './HeavyChart'

export default function Dashboard() {
  return <HeavyChart data={data} />
}
```

Correct pattern (dynamic import):

```typescript
import dynamic from 'next/dynamic'

const HeavyChart = dynamic(() => import('./HeavyChart'), {
  loading: () => <div>Loading chart...</div>
})

export default function Dashboard() {
  return <HeavyChart data={data} />
}
```

---

## Advanced Patterns

### Performance Monitoring

Measure React component performance using React Profiler and performance marks.

Use React.Profiler for Component Measurement:

Wrap components with Profiler to measure render performance.

```typescript
import { Profiler } from 'react'

function onRenderCallback(
  id, phase, actualDuration, baseDuration,
  startTime, commitTime, interactions
) {
  console.log(`${id} (${phase}) took ${actualDuration}ms`)
}

<Profiler id="Dashboard" onRender={onRenderCallback}>
  <Dashboard />
</Profiler>
```

### Optimization Strategies

Implement progressive optimization starting with highest impact changes.

Priority Order:

1. Eliminate waterfalls (CRITICAL): 2-10x improvement
2. Reduce bundle size (CRITICAL): 1.5-3x improvement
3. Optimize server-side (HIGH): 1.5-2x improvement
4. Improve data fetching (MEDIUM-HIGH): 1.2-1.5x improvement
5. Reduce re-renders (MEDIUM): 1.1-1.3x improvement
6. Optimize rendering (MEDIUM): 1.1-1.2x improvement

---

## Works Well With

- moai-domain-frontend: Frontend development patterns and UI optimization
- moai-platform-vercel: Vercel deployment and edge functions
- moai-lang-typescript: Type safety and patterns for React
- moai-workflow-testing: Test coverage for React components
- moai-foundation-claude: Claude Code integration and configuration
```

## Example 2: Web Design Guidelines Conversion

### Source Format Analysis

Original agent-skills structure:

```
web-design-guidelines/
├── SKILL.md
└── [no metadata.json or rules/ directory]
```

Source SKILL.md frontmatter:

```yaml
---
name: web-design-guidelines
description: Review UI code for Web Interface Guidelines compliance. Use when asked to "review my UI", "check accessibility", "audit design", "review UX", or "check my site against best practices".
argument-hint: <file-or-pattern>
---
```

### Conversion Process

Step 1: Determine new skill name

Original: web-design-guidelines
Converted: moai-domain-web-design (or moai-workflow-design-review)

Step 2: Convert description

Original: Review UI code for Web Interface Guidelines compliance. Use when asked to "review my UI", "check accessibility", "audit design", "review UX", or "check my site against best practices".

Converted: Web interface guidelines compliance checker covering accessibility standards, responsive design patterns, performance metrics, and user experience heuristics. Use when reviewing UI code, checking accessibility compliance, auditing design patterns, or validating UX best practices.

Step 3: Determine tool permissions

Required tools for web design compliance checker:

- Read: Read UI code files for review
- Grep: Search for compliance violations
- Glob: Find UI files by patterns
- WebFetch: Fetch remote guidelines from source URL

Step 4: Create Quick Reference section

Explain compliance review workflow:

1. Fetch latest guidelines from source URL
2. Read specified files or prompt user for files
3. Check against all rules in fetched guidelines
4. Output findings in file:line terse format

Step 5: Create Implementation Guide section

Provide step-by-step process for compliance review.

### Final MoAI-ADK Format

```yaml
---
name: moai-domain-web-design
description: Web interface guidelines compliance checker covering accessibility standards, responsive design patterns, performance metrics, and user experience heuristics. Use when reviewing UI code, checking accessibility compliance, auditing design patterns, or validating UX best practices.
allowed-tools: Read, Grep, Glob, WebFetch
version: 1.0.0
status: active
updated: 2026-01-15
category: domain
author: MoAI-ADK Team
tags:
  - web-design
  - accessibility
  - responsive-design
  - ux
  - compliance
---

# Web Design Guidelines Compliance Checker

## Quick Reference

Web interface guidelines compliance checker for automated design review.

Core Capabilities:

- Fetch latest guidelines from remote source URL
- Review files against accessibility standards
- Check responsive design patterns compliance
- Validate performance metrics and best practices
- Output findings in file:line terse format

Supported Guidelines:

- Accessibility standards (WCAG 2.1 AA)
- Responsive design patterns
- Performance metrics
- User experience heuristics

When to Use:

- Reviewing UI code for compliance issues
- Checking accessibility standards adherence
- Auditing design patterns against best practices
- Validating UX implementation quality

---

## Implementation Guide

### Compliance Review Workflow

Step 1: Fetch Latest Guidelines

Use WebFetch to retrieve the latest guidelines from the source URL.

Guidelines source URL:

```
https://raw.githubusercontent.com/vercel-labs/web-interface-guidelines/main/command.md
```

Step 2: Read Specified Files

Read the files specified by the user or prompt for file patterns if none provided.

Use Glob to find files matching patterns like **/*.tsx or **/*.jsx.

Use Read to read the contents of each file.

Step 3: Apply Guidelines Rules

Check each file against all rules in the fetched guidelines.

Common rule categories include:

- Color contrast ratios for accessibility
- Focus indicators for keyboard navigation
- Alt text for images
- Semantic HTML structure
- Responsive breakpoint handling
- Performance metric thresholds

Step 4: Output Findings

Format findings using the file:line terse format specified in the guidelines.

Example output format:

```
src/components/Button.tsx:42: Missing alt text on image element
src/components/Card.tsx:18: Color contrast ratio below WCAG AA threshold
src/pages/Home.tsx:125: Missing focus indicator on interactive element
```

### Usage Pattern

When a user provides a file or pattern argument:

1. Fetch guidelines from the source URL above
2. Read the specified files
3. Apply all rules from the fetched guidelines
4. Output findings using the format specified in the guidelines

If no files specified, ask the user which files to review.

Example prompts:

- "Review my UI components"
- "Check accessibility of src/components/"
- "Audit design patterns in **/*.tsx files"

---

## Advanced Patterns

### CI/CD Integration

Integrate compliance checking into CI/CD pipelines for automated design review.

GitHub Actions Example:

```yaml
name: Design Compliance Check

on: [pull_request]

jobs:
  compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check Design Compliance
        run: |
          claude -p "Review **/*.tsx for web design guidelines compliance" \
                 --allowedTools "Read,Grep,Glob,WebFetch"
```

### Custom Rule Creation

Create custom rules for project-specific design guidelines.

Custom rule structure:

```markdown
### Custom Rule: Button Hover States

Check that all button components have visible hover states.

Pattern: Button components must define :hover pseudo-class

Incorrect:

```css
.button {
  background: blue;
}
```

Correct:

```css
.button {
  background: blue;
}
.button:hover {
  background: darkblue;
}
```
```

---

## Works Well With

- moai-domain-uiux: UI/UX design system integration
- moai-lang-typescript: Type safety for UI components
- moai-workflow-testing: Visual regression testing
- moai-foundation-claude: Claude Code integration
- moai-workflow-docs: Documentation generation
```

## Example 3: Automated Conversion Script

### Conversion Script Template

```python
#!/usr/bin/env python3
"""Convert agent-skills format to MoAI-ADK skill format."""

import yaml
import json
from pathlib import Path
from datetime import datetime


def read_agent_skill(skill_dir: Path) -> dict:
    """Read agent-skills format from directory."""
    skill_md = skill_dir / "SKILL.md"
    metadata_json = skill_dir / "metadata.json"

    # Parse frontmatter from SKILL.md
    with open(skill_md) as f:
        content = f.read()
        frontmatter = content.split("---")[1]
        frontmatter_data = yaml.safe_load(frontmatter)

    # Parse metadata.json if exists
    metadata = {}
    if metadata_json.exists():
        with open(metadata_json) as f:
            metadata = json.load(f)

    return {
        "frontmatter": frontmatter_data,
        "metadata": metadata,
        "content": content
    }


def convert_to_moai_format(agent_skill: dict, skill_name: str) -> dict:
    """Convert agent-skills to MoAI-ADK format."""
    # Extract data
    agent_name = agent_skill["frontmatter"]["name"]
    agent_desc = agent_skill["frontmatter"]["description"]
    metadata = agent_skill["metadata"]

    # Convert name
    moai_name = f"moai-{skill_name}"

    # Convert description to include function and triggers
    # Extract trigger scenarios from original description
    if "Use when" in agent_desc:
        desc_part, triggers_part = agent_desc.split("Use when", 1)
        description = f"{desc_part.strip()}. Use when {triggers_part.strip()}"
    else:
        description = f"{agent_desc}. Use when working with {skill_name}."

    # Determine version
    version = metadata.get("version", "1.0.0")

    # Create frontmatter
    moai_frontmatter = {
        "name": moai_name,
        "description": description,
        "allowed-tools": "Read, Grep, Glob",  # Default, adjust based on needs
        "version": version,
        "status": "active",
        "updated": datetime.now().strftime("%Y-%m-%d"),
        "category": "domain",  # Adjust based on skill type
        "author": "MoAI-ADK Team",
        "tags": [skill_name]  # Add relevant tags
    }

    return {
        "frontmatter": moai_frontmatter,
        "content": agent_skill["content"]
    }


def write_moai_skill(moai_skill: dict, output_dir: Path):
    """Write MoAI-ADK skill file."""
    skill_file = output_dir / "SKILL.md"

    # Write frontmatter
    with open(skill_file, "w") as f:
        f.write("---\n")
        yaml.dump(moai_skill["frontmatter"], f)
        f.write("---\n\n")

        # Write content (add progressive disclosure structure)
        f.write("# TODO: Add Quick Reference section\n\n")
        f.write("# TODO: Add Implementation Guide section\n\n")
        f.write("# TODO: Add Advanced Patterns section\n\n")
        f.write("---\n\n")
        f.write("## Works Well With\n\n")
        f.write("- moai-foundation-core: Core execution patterns\n")
        f.write("- moai-workflow-docs: Documentation generation\n")


def main():
    """Main conversion workflow."""
    import sys

    if len(sys.argv) < 3:
        print("Usage: convert_agent_skill.py <agent-skills-dir> <skill-name>")
        sys.exit(1)

    agent_dir = Path(sys.argv[1])
    skill_name = sys.argv[2]

    # Read agent-skills format
    agent_skill = read_agent_skill(agent_dir)

    # Convert to MoAI-ADK format
    moai_skill = convert_to_moai_format(agent_skill, skill_name)

    # Write MoAI-ADK skill
    output_dir = Path(f".claude/skills/moai-{skill_name}")
    output_dir.mkdir(parents=True, exist_ok=True)
    write_moai_skill(moai_skill, output_dir)

    print(f"Converted skill written to {output_dir}/SKILL.md")


if __name__ == "__main__":
    main()
```

### Usage Example

```bash
# Convert react-best-practices skill
python convert_agent_skill.py \
    ~/Downloads/agent-skills-main/skills/react-best-practices \
    domain-react

# Convert web-design-guidelines skill
python convert_agent_skill.py \
    ~/Downloads/agent-skills-main/skills/web-design-guidelines \
    domain-web-design
```

## Example 4: Validation Script

### Validation Script Template

```python
#!/usr/bin/env python3
"""Validate MoAI-ADK skill format compliance."""

import yaml
import re
from pathlib import Path


def validate_frontmatter(skill_file: Path) -> list:
    """Validate skill frontmatter compliance."""
    errors = []

    with open(skill_file) as f:
        content = f.read()
        frontmatter_str = content.split("---")[1]
        frontmatter = yaml.safe_load(frontmatter_str)

    # Validate name
    if "name" not in frontmatter:
        errors.append("Missing 'name' field in frontmatter")
    else:
        name = frontmatter["name"]
        if not re.match(r'^[a-z0-9-]+$', name):
            errors.append(f"Name must be kebab-case: {name}")
        if len(name) > 64:
            errors.append(f"Name must be <= 64 characters: {name}")

    # Validate description
    if "description" not in frontmatter:
        errors.append("Missing 'description' field in frontmatter")
    else:
        desc = frontmatter["description"]
        if len(desc) > 1024:
            errors.append(f"Description must be <= 1024 characters")
        if "Use when" not in desc:
            errors.append("Description must include 'Use when' triggers")

    # Validate allowed-tools
    if "allowed-tools" not in frontmatter:
        errors.append("Missing 'allowed-tools' field in frontmatter")
    else:
        tools = frontmatter["allowed-tools"]
        if isinstance(tools, list):
            errors.append("allowed-tools must be comma-separated, not list")

    # Validate required fields
    required_fields = ["version", "status", "updated"]
    for field in required_fields:
        if field not in frontmatter:
            errors.append(f"Missing required field: {field}")

    return errors


def validate_line_count(skill_file: Path) -> list:
    """Validate 500-line limit."""
    errors = []

    with open(skill_file) as f:
        lines = f.readlines()

    if len(lines) > 500:
        errors.append(f"SKILL.md exceeds 500-line limit: {len(lines)} lines")

    return errors


def validate_progressive_disclosure(skill_file: Path) -> list:
    """Validate progressive disclosure structure."""
    errors = []

    with open(skill_file) as f:
        content = f.read()

    # Check for required sections
    required_sections = [
        "## Quick Reference",
        "## Implementation Guide",
        "## Advanced Patterns"
    ]

    for section in required_sections:
        if section not in content:
            errors.append(f"Missing required section: {section}")

    return errors


def main():
    """Main validation workflow."""
    import sys

    if len(sys.argv) < 2:
        print("Usage: validate_skill.py <skill-file>")
        sys.exit(1)

    skill_file = Path(sys.argv[1])

    all_errors = []
    all_errors.extend(validate_frontmatter(skill_file))
    all_errors.extend(validate_line_count(skill_file))
    all_errors.extend(validate_progressive_disclosure(skill_file))

    if all_errors:
        print(f"Validation errors in {skill_file}:")
        for error in all_errors:
            print(f"  - {error}")
        sys.exit(1)
    else:
        print(f"{skill_file} is valid!")


if __name__ == "__main__":
    main()
```

### Usage Example

```bash
# Validate converted skill
python validate_skill.py \
    .claude/skills/moai-domain-react/SKILL.md
```

These examples demonstrate the complete conversion process from agent-skills format to MoAI-ADK format, including frontmatter conversion, content organization, validation, and automation.
