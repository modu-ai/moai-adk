# TRUST 5 Validation Framework for MoAI-ADK Skills

**Automated and manual quality gates for 131 Skills v4.0.0**

Created: 2025-11-19 | Applies to: All Enterprise Skills

---

## TRUST 5 Quality Principles Applied to Skills

MoAI-ADK's TRUST 5 framework adapted for Skill validation:

| Principle | What It Means | Skill Validation |
|-----------|---------------|--------------------|
| **T**est-first | No code without tests | Skill examples must be testable, runnable |
| **R**eadable | Clear, maintainable | YAML valid, markdown structured, plain English |
| **U**nified | Consistent patterns | Standard template, terminology, formatting |
| **S**ecured | Security-first | No secrets, security patterns included |
| **T**rackable | Requirements linked | Versions current, sources cited, links valid |

---

## Automated Validation Checklist (YAML & Structure)

### Frontmatter Validation

```bash
✅ MUST PASS - Frontmatter
- [ ] name: matches "moai-{category}-{name}" pattern (kebab-case)
- [ ] version: exactly "4.0.0" (regex: ^4\.0\.0$)
- [ ] status: one of [stable, production, beta, deprecated]
- [ ] description: 100-200 characters, includes framework/use case
- [ ] allowed-tools: has at least Read, Bash, WebFetch
- [ ] YAML is valid syntax
- [ ] No missing required fields

✅ SHOULD PASS - Frontmatter
- [ ] tier: matches category (Language, Domain, Core, etc.)
- [ ] created/updated: present and valid dates
- [ ] primary-agent: matches domain expertise
- [ ] keywords: 5-7 searchable terms
- [ ] All fields use correct data types (string, list, etc.)

✅ NICE TO HAVE - Frontmatter
- [ ] secondary-agents: defined (for Domain/Core Skills)
- [ ] tags: additional categorization
- [ ] links: reference documentation URLs
```

### Markdown Structure Validation

```bash
✅ MUST PASS - Markdown
- [ ] Starts with frontmatter (---)
- [ ] First heading is H1 (#)
- [ ] No duplicate headers
- [ ] All code blocks have language specified
- [ ] Links are HTTPS or internal Skill references
- [ ] No orphaned headers (Header without content)

✅ SHOULD PASS - Markdown
- [ ] Code examples have syntax highlighting
- [ ] Tables are properly formatted
- [ ] Lists are consistently formatted
- [ ] No placeholder text like "[TODO]" or "FIXME"
- [ ] Sections follow Progressive Disclosure (Level 1/2/3)

✅ NICE TO HAVE - Markdown
- [ ] Proper emphasis/bold usage
- [ ] Consistent heading hierarchy
- [ ] Whitespace and alignment
```

### Version & Technology Stack Validation

```bash
✅ MUST PASS - Versions
- [ ] All mentioned frameworks/libraries have versions
- [ ] Versions are from Nov 2025 or earlier (no future dates)
- [ ] No deprecated framework versions mentioned
- [ ] Major versions match stable release (not beta/RC)

✅ SHOULD PASS - Versions
- [ ] Versions match official releases
- [ ] Package manager versions current
- [ ] LTS versions preferred where applicable
- [ ] Support/End-of-life dates accurate

✅ NICE TO HAVE - Versions
- [ ] Performance improvements noted
- [ ] Security fixes mentioned
- [ ] Breaking changes highlighted
```

### Security Validation

```bash
✅ MUST PASS - Security
- [ ] No hardcoded secrets/credentials
- [ ] No SQL injection examples
- [ ] No XSS vulnerability patterns
- [ ] No dangerous shell commands (rm -rf, sudo, etc.)
- [ ] No secrets in environment variable examples

✅ SHOULD PASS - Security
- [ ] Mentions OWASP Top 10 (if relevant)
- [ ] Input validation examples included
- [ ] Error handling shown (not exposing internals)
- [ ] Authentication pattern included
- [ ] Reference to security-related Skills provided

✅ NICE TO HAVE - Security
- [ ] Encryption/hashing examples
- [ ] Security best practices section
- [ ] Link to moai-security-* Skills
```

---

## Manual Validation Checklist (Content Quality)

### Content Quality (Progressive Disclosure)

```markdown
✅ MUST PASS - Content
- [ ] Level 1 (Quick Reference) present and useful
- [ ] Level 2 (Core Patterns) has 3+ real examples
- [ ] Level 3 (Advanced) references specialized Skills
- [ ] Progression is logical and builds understanding
- [ ] No gaps in explanation

✅ SHOULD PASS - Content
- [ ] Each level has 500-1000 words
- [ ] Code examples are production-ready
- [ ] Explanations are concise and clear
- [ ] No jargon without explanation
- [ ] Practical use cases included

✅ NICE TO HAVE - Content
- [ ] Diagrams or ASCII art
- [ ] Performance benchmarks
- [ ] Real-world case studies
- [ ] Troubleshooting section
```

### Cross-References Validation

```bash
✅ MUST PASS - References
- [ ] All Skill() references are valid (exist in codebase)
- [ ] All external links HTTPS
- [ ] Links tested and return 200 status
- [ ] No broken anchor links
- [ ] Cross-references are bidirectional where logical

✅ SHOULD PASS - References
- [ ] Related Skills linked appropriately
- [ ] Official documentation linked
- [ ] Community resources included
- [ ] Version-specific documentation referenced
- [ ] References are current (not 5+ years old)

✅ NICE TO HAVE - References
- [ ] GitHub repos linked
- [ ] NPM/PyPI package links
- [ ] Tutorial videos referenced
- [ ] Academic papers cited
```

### Consistency & Language Validation

```bash
✅ MUST PASS - Consistency
- [ ] Consistent terminology (don't mix "API" and "service endpoint")
- [ ] Code style consistent within examples
- [ ] British or American English (pick one consistently)
- [ ] No obvious typos or grammar errors
- [ ] Tone is professional and helpful

✅ SHOULD PASS - Consistency
- [ ] Matches style of other Skills in same category
- [ ] Consistent capitalization (Vue.js not vue.js)
- [ ] Consistent formatting for command syntax
- [ ] Consistent code comment style
- [ ] No internal contradictions

✅ NICE TO HAVE - Consistency
- [ ] Personality/voice matches Alfred
- [ ] Emoji usage consistent with brand
- [ ] Examples follow same pattern
```

---

## Spot-Check Validation (10-20% Sample)

For large batches, validate 10-20% of Skills manually:

### Sample Selection
- Random selection from each batch
- At least 1 from each tier (Language, Domain, Core)
- At least 1 "specialized" Skill
- At least 1 from each sub-team

### Spot-Check Form

```
Skill: [name]
Batch: [number]
Reviewer: [name]
Date: [YYYY-MM-DD]

✅ Frontmatter: [PASS/FAIL]
   Issues: [list any problems]

✅ Markdown: [PASS/FAIL]
   Issues: [list any problems]

✅ Versions: [PASS/FAIL]
   Issues: [list any problems]

✅ Security: [PASS/FAIL]
   Issues: [list any problems]

✅ Content: [PASS/FAIL]
   Issues: [list any problems]

✅ References: [PASS/FAIL]
   Issues: [list any problems]

Overall: [PASS/FAIL]
Confidence: [HIGH/MEDIUM/LOW]
Notes: [additional observations]
```

---

## Automated Validation Script

### Bash Script for Quick Checks

```bash
#!/bin/bash
# skill-validator.sh - Quick validation for all Skills

SKILLS_DIR="/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills"
REPORT_FILE=".moai/reports/validation-$(date +%Y%m%d-%H%M%S).txt"

echo "SKILL VALIDATION REPORT" > "$REPORT_FILE"
echo "Generated: $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

PASS_COUNT=0
FAIL_COUNT=0

for skill_dir in "$SKILLS_DIR"/*/; do
  skill_name=$(basename "$skill_dir")
  skill_file="$skill_dir/SKILL.md"
  
  [ ! -f "$skill_file" ] && echo "SKIP: $skill_name (no SKILL.md)" >> "$REPORT_FILE" && continue
  
  # Check version is 4.0.0
  if grep -q 'version: "4.0.0"' "$skill_file"; then
    version_ok=1
  else
    version_ok=0
    echo "FAIL: $skill_name (version not 4.0.0)" >> "$REPORT_FILE"
  fi
  
  # Check YAML is valid
  if python3 -c "import yaml; yaml.safe_load(open('$skill_file'))" 2>/dev/null; then
    yaml_ok=1
  else
    yaml_ok=0
    echo "FAIL: $skill_name (invalid YAML)" >> "$REPORT_FILE"
  fi
  
  # Check for required fields
  if grep -q "^name:" "$skill_file" && \
     grep -q "^status:" "$skill_file" && \
     grep -q "^allowed-tools:" "$skill_file"; then
    fields_ok=1
  else
    fields_ok=0
    echo "FAIL: $skill_name (missing required fields)" >> "$REPORT_FILE"
  fi
  
  if [ $version_ok -eq 1 ] && [ $yaml_ok -eq 1 ] && [ $fields_ok -eq 1 ]; then
    echo "PASS: $skill_name" >> "$REPORT_FILE"
    ((PASS_COUNT++))
  else
    ((FAIL_COUNT++))
  fi
done

echo "" >> "$REPORT_FILE"
echo "SUMMARY" >> "$REPORT_FILE"
echo "-------" >> "$REPORT_FILE"
echo "Passed: $PASS_COUNT" >> "$REPORT_FILE"
echo "Failed: $FAIL_COUNT" >> "$REPORT_FILE"
echo "Total: $((PASS_COUNT + FAIL_COUNT))" >> "$REPORT_FILE"

cat "$REPORT_FILE"
```

### Python Script for Link Validation

```python
#!/usr/bin/env python3
# link-validator.py - Check external links in Skills

import os
import re
import requests
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor

SKILLS_DIR = "/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills"
TIMEOUT = 5
MAX_WORKERS = 10

def check_url(url):
    """Check if URL is accessible."""
    try:
        response = requests.head(url, timeout=TIMEOUT, allow_redirects=True)
        return url, response.status_code in [200, 301, 302]
    except:
        return url, False

def validate_skill_links(skill_file):
    """Extract and validate all URLs in a Skill file."""
    content = Path(skill_file).read_text()
    urls = re.findall(r'https?://[^\s\)\]]+', content)
    
    results = []
    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        for url, is_valid in executor.map(check_url, set(urls)):
            results.append((url, is_valid))
    
    return results

# Main validation
issues = []
for skill_dir in sorted(Path(SKILLS_DIR).glob("*/")):
    skill_file = skill_dir / "SKILL.md"
    if not skill_file.exists():
        continue
    
    links = validate_skill_links(skill_file)
    for url, is_valid in links:
        if not is_valid:
            issues.append(f"{skill_dir.name}: {url}")

# Report
if issues:
    print(f"Found {len(issues)} link issues:")
    for issue in issues:
        print(f"  - {issue}")
else:
    print("All links validated successfully!")
```

---

## Quality Gate Criteria

### Passing Grade
- All MUST PASS checks: 100%
- SHOULD PASS checks: 80%+
- NICE TO HAVE checks: 50%+
- Spot-check confidence: HIGH

### Warning Grade
- All MUST PASS checks: 100%
- SHOULD PASS checks: 60-80%
- NICE TO HAVE checks: 30-50%
- Spot-check confidence: MEDIUM
- Action: Fix issues and re-validate

### Failing Grade
- Any MUST PASS check fails: FAIL entire batch
- SHOULD PASS checks: <60%
- Action: Return to team, fix issues, re-validate

---

## Validation Report Template

```markdown
# Skills Validation Report

**Date**: 2025-11-19  
**Batch**: [Batch Number]  
**Total Skills**: [count]  
**Validated Skills**: [count]  
**Pass Rate**: [X]%

## Summary

- ✅ Passed: [count]
- ⚠️ Warnings: [count]
- ❌ Failed: [count]

## Detailed Results

### Passed Skills
[List each skill marked PASS]

### Skills with Warnings
[List each skill marked WARNING with issues]

### Failed Skills
[List each skill marked FAIL with critical issues]

## Automated Checks Results

| Check | Pass | Fail | % |
|-------|------|------|---|
| YAML Valid | X | X | % |
| Version 4.0.0 | X | X | % |
| Required Fields | X | X | % |
| Links Valid | X | X | % |
| No Secrets | X | X | % |
| English Only | X | X | % |

## Spot-Check Results

- Spot-checked: [N] skills ([X]% of batch)
- Confidence: [HIGH/MEDIUM/LOW]
- Issues found: [count]

## Recommendations

1. [Action Item]
2. [Action Item]
3. [Action Item]

## Next Steps

- [ ] Fix failing Skills
- [ ] Resolve warnings
- [ ] Re-validate if needed
- [ ] Merge to main
- [ ] Deploy to production

---

**Validator**: [Name]  
**Validation Tool Version**: 1.0  
**Status**: [PASS/FAIL]
```

---

## Continuous Validation

### Daily
- Spot-check 2-3 completed Skills
- Run automated YAML validator
- Check for link breakage

### Weekly
- Full automated validation on completed batch
- Review warnings with team
- Update progress tracker

### Monthly
- Spot-check 10% of entire Skills library
- Check for version updates
- Security audit sample

---

## References

- TRUST 5 Framework: CLAUDE.md → TRUST 5 Quality Principles
- MoAI Philosophy: CLAUDE.md → SPEC-First Philosophy
- Skill Template: PHASE2-BATCH-TEMPLATE.md and PHASE3-BATCH-TEMPLATE.md

---

**Version**: 1.0 (2025-11-19)  
**Status**: Ready for Implementation  
**Scope**: All 131 MoAI-ADK Skills v4.0.0

