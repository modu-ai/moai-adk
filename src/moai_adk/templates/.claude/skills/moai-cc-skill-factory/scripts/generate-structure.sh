#!/bin/bash
set -euo pipefail

# Skill Structure Generator
# Creates standardized directory structure for new Skills
# Usage: ./generate-structure.sh skill-name "Skill Display Name"

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Arguments
SKILL_DIR="${1:-.}"
SKILL_NAME="${2:-New Skill}"

if [[ "$SKILL_DIR" == "." ]]; then
    echo -e "${YELLOW}⚠ Usage: generate-structure.sh <skill-dir> <display-name>${NC}"
    echo -e "Example: generate-structure.sh my-new-skill \"Processing CSV Files\""
    exit 1
fi

# Normalize skill directory name (kebab-case)
SKILL_DIR=$(echo "$SKILL_DIR" | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g')

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Skill Structure Generator${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# Create directories
echo -e "${GREEN}Creating directories...${NC}"
mkdir -p "$SKILL_DIR/scripts"
mkdir -p "$SKILL_DIR/templates"
echo -e "${GREEN}✓ Directories created${NC}"
echo

# Create SKILL.md if not exists
if [[ ! -f "$SKILL_DIR/SKILL.md" ]]; then
    cat > "$SKILL_DIR/SKILL.md" << 'EOF'
---
name: "Skill Name Here"
description: "What this Skill does and when to use it. Include 3+ discoverable keywords and trigger scenarios."
allowed-tools: "Tool1, Tool2"
---

# [Skill Name]

> **Quick Reference**: Links to supporting files go here.

---

## Overview

Describe what this Skill provides and its core purpose.

## Quick Start

Provide the fastest path to getting started with this Skill.

## Main Framework

### Section 1: High Freedom (Principles)

Provide principles, considerations, and trade-offs.

### Section 2: Medium Freedom (Patterns)

Include pseudocode, flowcharts, and pattern examples.

### Section 3: Low Freedom (Scripts)

Provide ready-to-use scripts with error handling.

## See Also

- [reference.md](reference.md) for detailed reference
- [examples.md](examples.md) for real-world scenarios
- `scripts/` for automation utilities
- `templates/` for reusable templates

---

**Version**: 0.1.0
**Status**: Draft
**Last Updated**: 2025-10-22
EOF
    echo -e "${GREEN}✓ Created SKILL.md${NC}"
else
    echo -e "${YELLOW}⊘ SKILL.md already exists${NC}"
fi
echo

# Create reference.md if not exists
if [[ ! -f "$SKILL_DIR/reference.md" ]]; then
    cat > "$SKILL_DIR/reference.md" << 'EOF'
# Reference: [Skill Name]

This document provides detailed reference information for [Skill Name].

## API Reference

### Function/Command 1

Description of what it does.

**Parameters**:
- `param1`: Description
- `param2`: Description

**Returns**: Description of return value

**Example**:
```
code example here
```

## Configuration

### Available Options

- Option 1: Description
- Option 2: Description

### Configuration Format

```yaml
example: configuration
structure: here
```

## Edge Cases

### Known Limitations

- Limitation 1
- Limitation 2

### Workarounds

- Workaround for limitation 1
- Workaround for limitation 2

## Performance Considerations

- Consideration 1
- Consideration 2

## Glossary

- **Term 1**: Definition
- **Term 2**: Definition

## Related Resources

- [SKILL.md](SKILL.md) for main framework
- [examples.md](examples.md) for usage examples

---

**Version**: 0.1.0
**Last Updated**: 2025-10-22
EOF
    echo -e "${GREEN}✓ Created reference.md${NC}"
else
    echo -e "${YELLOW}⊘ reference.md already exists${NC}"
fi
echo

# Create examples.md if not exists
if [[ ! -f "$SKILL_DIR/examples.md" ]]; then
    cat > "$SKILL_DIR/examples.md" << 'EOF'
# Examples: [Skill Name]

Real-world examples demonstrating [Skill Name] in practice.

## Example 1: Basic Usage

**Scenario**: [Describe a basic scenario]

**Code**:
```
example code here
```

**Expected Output**:
```
expected output here
```

## Example 2: Intermediate Pattern

**Scenario**: [Describe an intermediate scenario]

**Code**:
```
example code here
```

**Expected Output**:
```
expected output here
```

## Example 3: Advanced Use Case

**Scenario**: [Describe an advanced scenario]

**Code**:
```
example code here
```

**Expected Output**:
```
expected output here
```

## Example 4: Edge Case

**Scenario**: [Describe an edge case or unusual scenario]

**Code**:
```
example code here
```

**Expected Output**:
```
expected output here
```

## Common Patterns

### Pattern 1
Description and example of a common pattern.

### Pattern 2
Description and example of another common pattern.

## See Also

- [SKILL.md](SKILL.md) for framework and concepts
- [reference.md](reference.md) for API details

---

**Version**: 0.1.0
**Last Updated**: 2025-10-22
EOF
    echo -e "${GREEN}✓ Created examples.md${NC}"
else
    echo -e "${YELLOW}⊘ examples.md already exists${NC}"
fi
echo

# Create sample script template
if [[ ! -f "$SKILL_DIR/scripts/.gitkeep" && ! -f "$SKILL_DIR/scripts"/*.sh ]]; then
    touch "$SKILL_DIR/scripts/.gitkeep"
    echo -e "${GREEN}✓ Created scripts/.gitkeep${NC}"
else
    echo -e "${YELLOW}⊘ scripts directory already has content${NC}"
fi
echo

# Create sample template
if [[ ! -f "$SKILL_DIR/templates/.gitkeep" && ! -f "$SKILL_DIR/templates"/* ]]; then
    touch "$SKILL_DIR/templates/.gitkeep"
    echo -e "${GREEN}✓ Created templates/.gitkeep${NC}"
else
    echo -e "${YELLOW}⊘ templates directory already has content${NC}"
fi
echo

# Create .gitignore if not exists
if [[ ! -f "$SKILL_DIR/.gitignore" ]]; then
    cat > "$SKILL_DIR/.gitignore" << 'EOF'
# Temporary files
*.tmp
*.temp
.DS_Store

# IDE
.vscode/
.idea/
*.swp
*.swo

# Python
__pycache__/
*.py[cod]
.pytest_cache/
.coverage

# Node
node_modules/
.npm
EOF
    echo -e "${GREEN}✓ Created .gitignore${NC}"
fi
echo

# Display summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Skill Structure Created${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo
echo "Directory: $SKILL_DIR"
echo "Files created:"
echo "  ✓ SKILL.md (main content template)"
echo "  ✓ reference.md (supporting reference)"
echo "  ✓ examples.md (usage examples)"
echo "  ✓ scripts/ (for utility scripts)"
echo "  ✓ templates/ (for reusable templates)"
echo "  ✓ .gitignore (version control)"
echo
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Edit SKILL.md with your Skill content"
echo "2. Update metadata (name, description, allowed-tools)"
echo "3. Add examples to examples.md"
echo "4. Create scripts in scripts/ directory"
echo "5. Run validation: ../../scripts/validate-skill.sh $SKILL_DIR"
echo
echo -e "${BLUE}Documentation:${NC}"
echo "- See SKILL.md for content guidelines"
echo "- See templates/ for example structures"
echo
