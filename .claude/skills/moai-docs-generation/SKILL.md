# Skill: Documentation Generation & Template Management

## Metadata

```yaml
skill_id: moai-docs-generation
skill_name: Documentation Generation & Template Management
version: 1.0.0
created_date: 2025-11-10
updated_date: 2025-11-10
language: english
word_count: 1400
triggers:
  - keywords: [documentation generation, doc template, scaffold, generate docs, api documentation, readme generation]
  - contexts: [docs-generation, @docs:generate, documentation-template, doc-scaffold]
agents:
  - docs-manager
  - spec-builder
  - frontend-expert
  - backend-expert
freedom_level: high
context7_references:
  - url: "https://www.typescriptlang.org/docs/handbook/"
    topic: "TypeScript Documentation Pattern"
  - url: "https://github.com/prettier/prettier"
    topic: "Code Formatting Standards"
spec_reference: "@SPEC:DOCS-001"
```

## 📚 Content

### Section 1: Documentation Generation Framework

Automatic documentation generation accelerates the documentation process while maintaining consistency across the project. This skill covers:

- **Template-Based Generation**: Using standardized templates for common documentation
- **Scaffold Creation**: Bootstrap new documentation structures
- **API Documentation**: Auto-generating API docs from code comments
- **README Generation**: Creating project-level documentation
- **CHANGELOG Generation**: Tracking version changes automatically

**Benefits**:
- Consistent formatting across all documents
- Reduced manual documentation effort
- Living documentation that stays current
- Compliance with MoAI-ADK standards

### Section 2: Template Library

#### 1. Guide Template

```markdown
# [Feature Name] Guide

## Overview
Brief description of what this guide covers.

## Prerequisites
- Requirement 1
- Requirement 2

## Step-by-Step Tutorial
### Step 1: [Action]
Detailed explanation...

```code-example```

### Step 2: [Next Action]
...

## Best Practices
- Practice 1: Explanation
- Practice 2: Explanation

## Common Issues & Solutions
| Issue | Cause | Solution |
|-------|-------|----------|
| Problem 1 | Root cause | How to fix |

## See Also
- [Related Guide 1](../path/to/guide)
- [Related Guide 2](../path/to/guide)
```

#### 2. API Documentation Template

```markdown
# [Module Name] API Reference

## Overview
What this module/API does.

## Installation
\`\`\`bash
npm install @package/module
\`\`\`

## Usage

### Method: [methodName]

**Signature**:
\`\`\`typescript
function methodName(param1: Type1, param2: Type2): ReturnType
\`\`\`

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| param1 | Type1 | What it does |

**Returns**: Description of return value

**Example**:
\`\`\`typescript
const result = methodName(arg1, arg2);
\`\`\`

**Throws**: Possible exceptions

## Examples

### Example 1: Basic Usage
...

### Example 2: Advanced Usage
...

## Migration Guide
Instructions for upgrading from previous versions.
```

#### 3. Tutorial Template

```markdown
# Building [Project Type]: A Step-by-Step Tutorial

## What You'll Learn
- Learning outcome 1
- Learning outcome 2
- Learning outcome 3

## Prerequisites
- Skill 1
- Skill 2
- Tool requirements

## Project Setup

### Step 1: Initialize Project
\`\`\`bash
# Command to initialize
\`\`\`

### Step 2: Install Dependencies
\`\`\`bash
# Installation commands
\`\`\`

## Core Concepts

### Concept 1: [Name]
Explanation with examples.

### Concept 2: [Name]
Explanation with examples.

## Implementation

### Phase 1: [Phase Name]
Implementation details...

### Phase 2: [Phase Name]
Implementation details...

## Testing
How to verify your implementation.

## Deployment
Steps to deploy the project.

## Next Steps
- Advanced topic 1
- Advanced topic 2

## Troubleshooting
Common issues and solutions.
```

#### 4. README Template

```markdown
# [Project Name]

![Status Badge](shields.io-badge)
[![License](shields.io-license)](LICENSE)

## Description
One-sentence description of what this project does.

## Features
- Feature 1
- Feature 2
- Feature 3

## Quick Start

### Installation
\`\`\`bash
# Installation command
\`\`\`

### Basic Usage
\`\`\`bash
# Basic command
\`\`\`

## Documentation
- [Getting Started](docs/getting-started.md)
- [API Reference](docs/api-reference.md)
- [Advanced Topics](docs/advanced.md)

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md)

## License
[License Type](LICENSE)
```

### Section 3: Scaffold Generation

**Directory Structure Generation**:

```python
class DocumentationScaffold:
    def __init__(self, project_name: str):
        self.project_name = project_name

    def create_guide_structure(self, guide_name: str) -> None:
        """Create guide directory and template files"""
        guide_dir = Path(f"docs/guides/{guide_name}")
        guide_dir.mkdir(parents=True, exist_ok=True)

        # Create index.md with guide template
        index_path = guide_dir / "index.md"
        index_path.write_text(GUIDE_TEMPLATE)

        # Create subdirectories
        (guide_dir / "examples").mkdir(exist_ok=True)
        (guide_dir / "images").mkdir(exist_ok=True)
        (guide_dir / "code-samples").mkdir(exist_ok=True)

    def create_api_docs(self, module_name: str) -> None:
        """Generate API documentation structure"""
        api_dir = Path(f"docs/api/{module_name}")
        api_dir.mkdir(parents=True, exist_ok=True)

        # Create main API doc
        api_path = api_dir / "index.md"
        api_path.write_text(API_TEMPLATE)

    def create_multilingual_structure(self, doc_name: str) -> None:
        """Create docs in ko/, en/, ja/, zh/"""
        for lang in ["ko", "en", "ja", "zh"]:
            doc_dir = Path(f"docs/src/{lang}/{doc_name}")
            doc_dir.mkdir(parents=True, exist_ok=True)

            doc_path = doc_dir / "index.md"
            doc_path.write_text(self._get_template_for_lang(lang))
```

### Section 4: Auto-Documentation from Code

**TypeScript/JavaScript**:

```typescript
/**
 * Calculate sum of two numbers
 * @param a First number
 * @param b Second number
 * @returns Sum of a and b
 * @example
 * const result = sum(2, 3);  // Returns 5
 */
function sum(a: number, b: number): number {
    return a + b;
}
```

Generate documentation:
```markdown
### Function: sum

Calculate sum of two numbers

**Signature**:
```typescript
function sum(a: number, b: number): number
```

**Parameters**:
- `a`: First number
- `b`: Second number

**Returns**: Sum of a and b

**Example**:
```typescript
const result = sum(2, 3);  // Returns 5
```
```

**Python**:

```python
def calculate_mean(numbers: List[float]) -> float:
    """
    Calculate arithmetic mean of numbers.

    Args:
        numbers: List of numerical values

    Returns:
        Arithmetic mean of the values

    Raises:
        ValueError: If list is empty

    Example:
        >>> calculate_mean([1, 2, 3])
        2.0
    """
    if not numbers:
        raise ValueError("Cannot calculate mean of empty list")
    return sum(numbers) / len(numbers)
```

### Section 5: Batch Generation

**Generate all project documentation**:

```bash
# Generate README for each module
/docs:generate --type readme --scope all

# Generate API documentation from code
/docs:generate --type api --language typescript

# Create guides for new features
/docs:generate --type guide --feature-spec SPEC-001

# Generate multilingual structure
/docs:generate --type i18n --languages ko,en,ja,zh
```

## 🎯 Usage

### From Agents

```python
# docs-manager agent
Skill("moai-docs-generation")

# Generate new guide
scaffold = DocumentationScaffold("my-project")
scaffold.create_guide_structure("new-feature")

# Generate API docs
scaffold.create_api_docs("my-module")

# Create multilingual structure
scaffold.create_multilingual_structure("tutorial")
```

### Integration with SPEC

When creating a new SPEC:
1. Auto-generate guide structure from SPEC template
2. Create corresponding test documentation
3. Generate implementation checklist
4. Create README with links

## 📚 Reference Materials

- [MoAI-ADK Documentation Templates](https://docs.moai-adk.io/templates)
- [TypeScript JSDoc Reference](https://www.typescriptlang.org/docs/handbook/jsdoc-reference.html)
- [Python Docstring Format](https://google.github.io/styleguide/pyguide.html)

## ✅ Validation Checklist

- [x] Template library comprehensive
- [x] Scaffold generation patterns included
- [x] Auto-documentation examples provided
- [x] Multilingual support documented
- [x] Code examples included
- [x] Integration patterns shown
- [x] English language confirmed
