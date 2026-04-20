# Claude Code Skills Formatting Guide

Complete formatting reference for creating Claude Code Skills that comply with official standards and best practices.

Purpose: Standardized formatting guide for skill creation and validation
Target: Skill creators and maintainers
Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

Core Format: YAML frontmatter + markdown content with progressive disclosure. Naming: kebab-case, max 64 chars, official prefixes. Structure: SKILL.md (≤500 lines) + optional supporting files. Tools: Minimal permissions, principle of least privilege.

---

## Complete Skill Template

```yaml
---
name: skill-name # Required: kebab-case, max 64 chars
description: Specific description of skill purpose and trigger scenarios (max 1024 chars) # Required
allowed-tools: tool1, tool2, tool3 # Optional: comma-separated, minimal set
version: 1.0.0 # Optional: semantic versioning
tags: [domain, category, purpose] # Optional: categorization
updated: 2025-11-25 # Optional: last update date
status: active # Optional: active/deprecated/experimental
---

# Skill Title [Human-readable name]

Brief one-sentence description of skill purpose.

## Quick Reference (30 seconds)

One paragraph summary of core functionality and immediate use cases. Focus on what the skill does and when to use it.

## Implementation Guide

### Core Capabilities
- Capability 1: Specific description with measurable outcome
- Capability 2: Specific description with measurable outcome
- Capability 3: Specific description with measurable outcome

### When to Use
- Use Case 1: Clear trigger scenario with specific indicators
- Use Case 2: Clear trigger scenario with specific indicators
- Use Case 3: Clear trigger scenario with specific indicators

### Essential Patterns
```python
# Pattern 1: Specific use case with code example
def example_function():
 """
 Clear purpose and expected outcome
 """
 return result
```

```bash
# Pattern 2: Command-line example
# Clear purpose and expected outcome
command --option --argument
```

## Best Practices

 DO:
- Specific positive recommendation with clear rationale
- Concrete example of recommended practice
- Performance consideration or optimization tip

 DON'T:
- Common mistake with explanation of negative impact
- Anti-pattern with better alternative suggestion
- Security or performance pitfall to avoid

## Works Well With

- [`related-skill-name`](../related-skill/SKILL.md) - Brief description of relationship and usage pattern
- [`another-related-skill`](../another-skill/SKILL.md) - Brief description of relationship and usage pattern

## Advanced Features

### Feature 1: Complex capability
Detailed explanation of advanced functionality with examples.

### Feature 2: Integration pattern
How this skill integrates with other tools or systems.

## Troubleshooting

Issue: Symptom description
Solution: Step-by-step resolution approach

Issue: Another problem description
Solution: Clear fix with verification steps
```

---

## Frontmatter Field Specifications

### Required Fields

#### `name` (String)
Format: kebab-case (lowercase, numbers, hyphens only)
Length: Maximum 64 characters
Pattern: `[prefix]-[domain]-[function]`
Examples:
- `moai-cc-commands`
- `moai-lang-python`
- `moai-domain-backend`
- `MyAwesomeSkill` (uppercase, spaces)
- `skill_v2` (underscore)
- `this-name-is-way-too-long-and-exceeds-the-sixty-four-character-limit`

#### `description` (String)
Format: Natural language description
Length: Maximum 1024 characters
Content: What the skill does + specific trigger scenarios
Examples:
- `Extract and structure information from PDF documents for analysis and processing. Use when you need to analyze PDF content, extract tables, or convert PDF text to structured data.`
- `Helps with documents` (too vague)
- `This skill processes various types of files` (lacks specificity)

### Optional Fields

#### `allowed-tools` (String List)
Format: Comma-separated list, no brackets
Purpose: Principle of least privilege
Examples:
```yaml
# CORRECT: Minimal specific tools
allowed-tools: Read, mcp__context7__resolve-library-id

# CORRECT: Multiple tools for analysis
allowed-tools: Read, Grep, Glob, WebFetch

# WRONG: YAML array format
allowed-tools: [Read, Grep, Glob]

# WRONG: Overly permissive
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, WebFetch, MultiEdit
```

#### `version` (String)
Format: Semantic versioning (X.Y.Z)
Purpose: Track skill evolution
Examples:
```yaml
version: 1.0.0 # Initial release
version: 1.1.0 # Feature addition
version: 1.0.1 # Bug fix
version: 2.0.0 # Breaking changes
```

#### `tags` (Array)
Format: List of category tags
Purpose: Skill discovery and categorization
Examples:
```yaml
tags: [documentation, claude-code, formatting]
tags: [python, testing, ddd]
tags: [security, owasp, validation]
```

#### `updated` (Date)
Format: YYYY-MM-DD
Purpose: Track last modification
Examples:
```yaml
updated: 2025-11-25
```

#### `status` (String)
Options: `active`, `deprecated`, `experimental`
Purpose: Indicate skill status
Examples:
```yaml
status: active # Production ready
status: experimental # Testing phase
status: deprecated # Superseded by newer skill
```

---

## Content Structure Guidelines

### Section 1: Quick Reference (30 seconds)

Purpose: Immediate value proposition
Length: 2-4 sentences maximum
Content: Core functionality + primary use cases
Example:
```markdown
## Quick Reference (30 seconds)

Context7 MCP server integration for real-time library documentation access. Resolve library names to Context7 IDs and fetch latest API documentation with progressive token disclosure for optimal performance.
```

### Section 2: Implementation Guide

Purpose: Step-by-step usage instructions
Structure:
- Core Capabilities (bullet points)
- When to Use (specific scenarios)
- Essential Patterns (code examples)

#### Core Capabilities Format
```markdown
### Core Capabilities
- Capability Name: Clear description with measurable outcome
- Another Capability: Specific description with expected results
- Third Capability: Detailed explanation of functionality
```

#### When to Use Format
```markdown
### When to Use
- Specific Scenario: Clear trigger condition with indicators
- Another Scenario: Detailed context and requirements
- Edge Case: Special circumstances and handling approach
```

#### Essential Patterns Format
```markdown
### Essential Patterns
```python
# Pattern Name: Clear purpose
def example_function(param1, param2):
 """
 Brief description of function purpose
 and expected behavior.
 """
 return result # Clear outcome
```

```bash
# Command Pattern: Clear purpose
command --option value --flag
# Expected output or result
```
```

### Section 3: Best Practices

Purpose: Pro guidance and common pitfalls
Format: DO/DON'T lists with explanations

```markdown
## Best Practices

 DO:
- Specific positive recommendation with clear rationale
- Concrete implementation example with code
- Performance or security consideration

 DON'T:
- Common mistake with explanation of negative impact
- Anti-pattern with better alternative
- Security vulnerability or performance issue
```

### Section 4: Works Well With

Purpose: Skill relationships and integration
Format: Link list with relationship descriptions

```markdown
## Works Well With

- [`related-skill`](../related-skill/SKILL.md) - Specific relationship and usage pattern
- [`another-skill`](../another-skill/SKILL.md) - Integration scenario and workflow
```

---

## Code Example Standards

### Python Examples

```python
# CORRECT: Complete, documented example
def validate_api_response(response_data, schema):
 """
 Validate API response against expected schema.

 Args:
 response_data (dict): API response to validate
 schema (dict): Expected schema structure

 Returns:
 bool: True if valid, False otherwise

 Raises:
 ValidationError: When schema validation fails
 """
 try:
 jsonschema.validate(response_data, schema)
 return True
 except jsonschema.ValidationError as e:
 logger.error(f"Schema validation failed: {e}")
 return False
```

### JavaScript/TypeScript Examples

```typescript
// CORRECT: Typed, documented example
interface UserConfig {
 apiUrl: string;
 timeout: number;
 retries: number;
}

/
 * Create HTTP client with configuration
 * @param config - User configuration options
 * @returns Configured axios instance
 */
function createHttpClient(config: UserConfig): AxiosInstance {
 return axios.create({
 baseURL: config.apiUrl,
 timeout: config.timeout,
 retry: config.retries,
 });
}
```

### Bash/Shell Examples

```bash
#!/bin/bash
# CORRECT: Safe, documented script

# Backup database with compression
# Usage: ./backup-db.sh [database_name] [output_directory]

set -euo pipefail # Strict error handling

DATABASE_NAME=${1:-"default_db"}
OUTPUT_DIR=${2:-"./backups"}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${OUTPUT_DIR}/${DATABASE_NAME}_backup_${TIMESTAMP}.sql.gz"

echo "Starting backup for database: ${DATABASE_NAME}"
mkdir -p "${OUTPUT_DIR}"

# Create compressed backup
pg_dump "${DATABASE_NAME}" | gzip > "${BACKUP_FILE}"

echo "Backup completed: ${BACKUP_FILE}"
```

---

## File Organization Standards

### Required Structure

```
skill-name/
 SKILL.md # REQUIRED (main file, ≤500 lines)
 reference.md # OPTIONAL (documentation links)
 examples.md # OPTIONAL (additional examples)
 scripts/ # OPTIONAL (utility scripts)
 helper.sh
 tool.py
 templates/ # OPTIONAL (reusable templates)
 template.md
```

### File Naming Conventions

SKILL.md: Always uppercase, main skill file
reference.md: External documentation and links
examples.md: Additional working examples beyond main file
scripts/: Executable utilities and helper tools
templates/: Reusable file templates and patterns

### Content Distribution Strategy

SKILL.md (≤500 lines):
- Quick Reference: 50-80 lines
- Implementation Guide: 200-300 lines
- Best Practices: 80-120 lines
- Works Well With: 20-30 lines
- Advanced Features: 0-50 lines (optional)

reference.md (unlimited):
- Official documentation links
- External resources and tutorials
- Related tools and libraries
- Community resources

examples.md (unlimited):
- Complete working examples
- Integration scenarios
- Test cases and validation
- Common usage patterns

---

## Validation Checklist

### Pre-publication Validation

Frontmatter Validation:
- [ ] Name uses kebab-case (64 chars max)
- [ ] Description specific and under 1024 chars
- [ ] allowed-tools follows principle of least privilege
- [ ] YAML syntax valid (no parsing errors)
- [ ] No deprecated or invalid fields

Content Structure Validation:
- [ ] Quick Reference section present (30-second value)
- [ ] Implementation Guide with all required subsections
- [ ] Best Practices with DO/DON'T format
- [ ] Works Well With section with valid links
- [ ] Total line count ≤ 500 for SKILL.md

Code Example Validation:
- [ ] All code examples are functional and tested
- [ ] Proper language identifiers in code blocks
- [ ] Comments and documentation included
- [ ] Error handling where appropriate
- [ ] No hardcoded credentials or sensitive data

Link Validation:
- [ ] Internal links use relative paths
- [ ] External links are accessible and relevant
- [ ] No broken or outdated references
- [ ] Proper markdown link formatting

### Quality Standards Validation

Clarity and Specificity:
- [ ] Clear value proposition in Quick Reference
- [ ] Specific trigger scenarios and use cases
- [ ] Actionable examples and patterns
- [ ] No ambiguous or vague language

Technical Accuracy:
- [ ] Code examples follow language conventions
- [ ] Technical details are current and accurate
- [ ] Best practices align with official documentation
- [ ] Security considerations where relevant

User Experience:
- [ ] Logical flow from simple to complex
- [ ] Progressive disclosure structure
- [ ] Effective troubleshooting section
- [ ] Consistent formatting and style

---

## Common Formatting Errors

### YAML Frontmatter Errors

Invalid Array Format:
```yaml
# WRONG: YAML array syntax
allowed-tools: [Read, Write, Bash]

# CORRECT: Comma-separated string
allowed-tools: Read, Write, Bash
```

Missing Required Fields:
```yaml
# WRONG: Missing description
---
name: my-skill
---

# CORRECT: All required fields present
---
name: my-skill
description: Specific description of skill purpose
---
```

### Content Structure Errors

Line Count Exceeded:
```markdown
# WRONG: SKILL.md exceeds 500 lines
# (too much content in main file)

# CORRECT: Move detailed content to supporting files
# Main SKILL.md: ≤500 lines
# reference.md: Additional documentation
# examples.md: More working examples
```

Missing Required Sections:
```markdown
# WRONG: Missing Quick Reference section
# No clear value proposition

# CORRECT: All required sections present
## Quick Reference (30 seconds)
Brief summary of core functionality...

## Implementation Guide
### Core Capabilities
...
```

### Link and Reference Errors

Broken Internal Links:
```markdown
# WRONG: Incorrect relative path
- [`related-skill`](./related-skil/SKILL.md) # typo in path

# CORRECT: Valid relative path
- [`related-skill`](../related-skill/SKILL.md)
```

Missing Code Language Identifiers:
```markdown
# WRONG: No language specified
```
function example() {
 return "result";
}
```

# CORRECT: Language specified
```javascript
function example() {
 return "result";
}
```
```

---

## Performance Optimization

### Token Usage Optimization

Progressive Disclosure Strategy:
1. SKILL.md: Core functionality only (≤500 lines)
2. reference.md: External links and documentation
3. examples.md: Additional working examples
4. scripts/: Utility code and tools

Content Prioritization:
- Essential information in SKILL.md
- Supplementary content in supporting files
- External references in reference.md
- Advanced patterns in separate modules

### Loading Speed Optimization

File Organization:
- Keep SKILL.md lean and focused
- Use supporting files for detailed content
- Optimize internal link structure
- Minimize cross-references depth

Discovery Optimization:
- Specific, descriptive names
- Clear trigger scenarios in description
- Relevant tags for categorization
- Consistent naming conventions

---

## Integration Patterns

### Skill Chaining

Sequential Usage:
```markdown
## Works Well With

- [`skill-a`](../skill-a/SKILL.md) - Use first for data preparation
- [`skill-b`](../skill-b/SKILL.md) - Use after skill-a for analysis
```

Parallel Usage:
```markdown
## Works Well With

- [`skill-x`](../skill-x/SKILL.md) - Alternative approach for similar tasks
- [`skill-y`](../skill-y/SKILL.md) - Complementary functionality for different aspects
```

### MCP Integration Patterns

Context7 Integration:
```yaml
allowed-tools: mcp__context7__resolve-library-id, mcp__context7__get-library-docs
```

```python
# Two-step pattern
library_id = await mcp__context7__resolve-library_id("library-name")
docs = await mcp__context7__get-library_docs(
 context7CompatibleLibraryID=library_id,
 topic="specific-topic",
 tokens=3000
)
```

Multi-MCP Integration:
```yaml
allowed-tools: mcp__context7__*, mcp__playwright__*, mcp__pencil__*
```

---

## Maintenance and Updates

### Version Management

Semantic Versioning:
- Major (X.0.0): Breaking changes, incompatible API
- Minor (0.Y.0): New features, backward compatible
- Patch (0.0.Z): Bug fixes, documentation updates

Update Process:
1. Update version number in frontmatter
2. Update `updated` field
3. Document changes in changelog
4. Test functionality with examples
5. Validate against current standards

### Compatibility Tracking

Claude Code Version Compatibility:
- Document compatible Claude Code versions
- Test with latest Claude Code release
- Update examples for breaking changes
- Monitor official documentation updates

Library Version Compatibility:
- Track supported library versions
- Update examples for breaking changes
- Document migration paths
- Test with current library releases

---

## Advanced Formatting Features

### Conditional Content

Model-Specific Content:
```markdown
### For Claude Sonnet
Advanced patterns requiring complex reasoning...

### For Claude Haiku
Optimized patterns for fast execution...
```

Context-Dependent Content:
```markdown
### When Working with Large Files
Use streaming approaches and chunk processing...

### When Working with APIs
Implement retry logic and error handling...
```

### Interactive Examples

Step-by-Step Tutorials:
```markdown
### Tutorial: Complete Workflow

Step 1: Setup and preparation
```bash
# Setup commands
```

Step 2: Core implementation
```python
# Implementation code
```

Step 3: Validation and testing
```bash
# Test commands
```

Expected result: [Clear outcome description]
```

### Multi-language Support

Language-Agnostic Patterns:
```markdown
### Core Pattern (Language Independent)
1. [Step description]
2. [Step description]
3. [Step description]

Python Implementation:
```python
# Python-specific code
```

JavaScript Implementation:
```javascript
// JavaScript-specific code
```

Go Implementation:
```go
// Go-specific code
```
```

---

Version: 2.0.0
Compliance: Claude Code Official Standards
Last Updated: 2025-11-25
Next Review: 2025-12-25 or standards update

Generated with Claude Code using official documentation and best practices.
