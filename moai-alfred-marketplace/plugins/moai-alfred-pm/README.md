# PM Plugin - Project Management Automation for Alfred Framework

**Version**: 1.0.0-dev
**Author**: GOOS🪿
**License**: MIT
**Status**: Stable

## Overview

PM Plugin provides project management automation and SPEC generation capabilities for the Alfred Framework. It generates comprehensive project documentation using the EARS (Easy Approach to Requirements Syntax) specification framework, enabling structured, traceable project planning from inception to completion.

## Features

- **Automated SPEC Generation**: Create complete specification documents with YAML frontmatter and EARS-formatted requirements
- **Project Charter Creation**: Generate governance structures with stakeholder matrices and decision authorities
- **Risk Assessment Matrix**: Build risk assessments based on project risk level (low/medium/high)
- **Implementation Planning**: Generate 5-phase implementation plans with milestones and deliverables
- **Acceptance Criteria**: Define quality metrics and completion criteria with measurable targets
- **Template System**: Support for multiple templates (moai-spec, enterprise, agile) for different project types

## Installation

### Using `uv` (Recommended)

```bash
# Install PM Plugin into your MoAI-ADK environment
uv pip install moai-alfred-pm

# Or from source
cd moai-alfred-marketplace/plugins/moai-alfred-pm
uv pip install -e .
```

### Using `pip`

```bash
pip install moai-alfred-pm

# Or from source
cd moai-alfred-marketplace/plugins/moai-alfred-pm
pip install -e .
```

## Quick Start

### Basic Project Initialization

```bash
/init-pm my-awesome-project
```

Creates:
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/spec.md` - EARS specification with 5 requirement patterns
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/plan.md` - 5-phase implementation plan
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/acceptance.md` - Quality metrics and sign-off criteria
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/charter.md` - Project governance
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/risk-matrix.json` - Risk assessment data

### Advanced Usage

```bash
# Enterprise template with high risk assessment
/init-pm payment-system --template=enterprise --risk-level=high

# Minimal setup without charter
/init-pm api-service --skip-charter

# Standard project with medium risk
/init-pm ecommerce-platform --risk-level=medium
```

## Command Reference

### `/init-pm`

Initialize project management templates with EARS SPEC framework.

**Syntax**:
```bash
/init-pm <project-name> [options]
```

**Arguments**:
- `project-name` (required): Project identifier
  - Format: lowercase letters, numbers, hyphens (e.g., `ecommerce-platform`)
  - Length: 3-50 characters
  - Generates SPEC ID: `SPEC-{PROJECT}-001`

**Options**:
- `--template <value>` (optional): Template type
  - Values: `moai-spec` (default), `enterprise`, `agile`

- `--risk-level <value>` (optional): Risk assessment level
  - Values: `low`, `medium` (default), `high`
  - Determines risk count: low=3, medium=6, high=10+

- `--skip-charter` (optional): Skip project charter generation
  - Generates only: spec.md, plan.md, acceptance.md
  - No value required

**Examples**:

```bash
# Basic usage
/init-pm my-project

# High-risk enterprise project
/init-pm payment-system --template=enterprise --risk-level=high

# Agile project without governance documents
/init-pm mobile-app --template=agile --skip-charter

# Low-risk internal tool
/init-pm internal-tool --risk-level=low
```

## Generated Files

### 1. `spec.md` - EARS Specification

**Structure**:
```yaml
---
spec_id: SPEC-MY-PROJECT-001
title: My Project - Project Specification
version: 1.0.0-dev
status: In Development
owner: [Project Owner]
created: 2025-10-30
tags: [spec, project, template]
---

## EARS Requirements

### Ubiquitous Behaviors (Core Features)
- GIVEN...WHEN...THEN format

### Event-Driven Behaviors
- Triggered by system events

### State-Driven Behaviors
- Based on application state

### Optional Behaviors
- Optional/conditional requirements

### Unwanted Behaviors
- Security, error handling, edge cases
```

**EARS Format** (5 Requirement Patterns):
1. **Ubiquitous**: Core features always available
2. **Event-Driven**: Triggered by specific events
3. **State-Driven**: Based on system state
4. **Optional**: Conditional or optional functionality
5. **Unwanted**: Security, errors, edge cases

### 2. `plan.md` - Implementation Plan

**5-Phase Structure**:
1. **Kickoff** (Week 1): Alignment, charter, resource allocation, risk assessment
2. **Design** (Week 2-3): Architecture, technology selection, API contracts
3. **Implementation** (Week 4-8): Development, testing, integration
4. **Validation** (Week 9-10): UAT, bug fixes, performance testing
5. **Release** (Week 11-12): Deployment, documentation, support

**Includes**:
- Weekly activities and milestones
- Deliverables checklist
- Resource allocation (team roles, budget)
- Risk management references
- Timeline overview table

### 3. `acceptance.md` - Acceptance Criteria

**Quality Metrics Table**:
| Metric | Target | Status |
|--------|--------|--------|
| Test Coverage | ≥85% | ⏳ |
| Linting | 0 errors | ⏳ |
| Type Safety | 100% | ⏳ |
| Code Review | Approved | ⏳ |
| Deployment | Verified | ⏳ |

**Sign-off Section**: PM, Tech Lead, QA Lead, Product Owner

### 4. `charter.md` - Project Charter

**Governance Structure**:
- Project overview (name, ID, manager, sponsor)
- Business case (objectives, expected benefits, success criteria)
- Stakeholder matrix (roles, responsibilities, contacts)
- Budget and schedule (timeline, budget allocation, contingency)
- Governance (approval chain, decision authority)
- Risk management references

### 5. `risk-matrix.json` - Risk Assessment

**Data Structure**:
```json
{
  "spec_id": "SPEC-MY-PROJECT-001",
  "created": "2025-10-30T12:34:56.789Z",
  "risk_level": "medium",
  "total_risks": 6,
  "risks": [
    {
      "id": "RISK-001",
      "description": "Risk description",
      "category": "Technical|Process|Resource",
      "probability": "Low|Medium|High",
      "impact": "Low|Medium|High",
      "mitigation": "Mitigation strategy",
      "owner": "TBD",
      "status": "Identified|Mitigated|Closed"
    }
  ]
}
```

## Integration with Alfred Framework

### Next Steps After Initialization

1. **Run Implementation**:
   ```bash
   /alfred:2-run SPEC-MY-PROJECT-001
   ```

2. **Synchronize Documentation**:
   ```bash
   /alfred:3-sync auto SPEC-MY-PROJECT-001
   ```

### Workflow Integration

PM Plugin integrates with the Alfred framework's 4-step workflow:

```
Step 1: Intent Understanding
   ↓
Step 2: Plan Creation (/alfred:1-plan)
   ↓
Step 3: Task Execution (/alfred:2-run SPEC-ID)
   ↓
Step 4: Report & Commit (/alfred:3-sync)
```

## Error Handling

### Common Errors and Recovery

**Invalid Project Name**:
```
❌ Project name must contain only lowercase letters, numbers, and hyphens
```
✅ Solution: Use format: `my-awesome-project` (lowercase, hyphens only)

**SPEC Already Exists**:
```
❌ SPEC already exists: .moai/specs/SPEC-MY-PROJECT-001/
```
✅ Solution: Use different name or remove existing SPEC

**Invalid Risk Level**:
```
❌ Invalid risk level: extreme
   Supported levels: low, medium, high
```
✅ Solution: Choose from low, medium, or high

**Invalid Template**:
```
❌ Invalid template: custom
   Supported: moai-spec (default), enterprise, agile
```
✅ Solution: Use supported template or omit for default

## Requirements

- **Python**: 3.11+
- **Dependencies**: PyYAML (for YAML frontmatter handling)
- **Alfred Framework**: v1.0+

## Architecture

### Command-Agent Model

```
/init-pm command
    ↓
pm-agent (PM Agent)
    ├─ Validates project name format
    ├─ Invokes Skills for requirement syntax
    ├─ Generates SPEC documents
    ├─ Creates project governance
    ├─ Builds risk assessments
    └─ Validates @TAG chain integrity
```

### Skills Dependencies

- `moai-foundation-ears`: EARS syntax patterns (5 requirement types)
- `moai-spec-authoring`: SPEC document templates and structure
- `moai-foundation-specs`: @TAG validation (CODE-FIRST principle)
- `moai-plugin-scaffolding`: Plugin generation best practices

## Testing

PM Plugin includes comprehensive test coverage (94% coverage, 17 passed tests):

```bash
# Run all tests
cd plugins/moai-alfred-pm
python -m pytest tests/ -v

# Run specific test class
python -m pytest tests/test_commands.py::TestInitPMCommand -v

# Run with coverage report
python -m pytest tests/ --cov=pm_plugin --cov-report=html
```

**Test Categories**:
- **Normal Cases** (5 tests): Basic functionality, EARS format, YAML, charter, risk matrix
- **Option Cases** (2 tests): Template variations, skip-charter option
- **Error Cases** (5 tests): Invalid inputs, duplicate SPEC, invalid options
- **Boundary Cases** (3 tests): Min/max name lengths, multiple hyphens
- **Integration Tests** (2 tests): End-to-end workflow, output structure
- **Performance Tests** (1 test): Command completion time <5 seconds

## Configuration

PM Plugin uses the Alfred framework's `.moai/config.json` for configuration:

```json
{
  "plugin": {
    "moai-alfred-pm": {
      "enabled": true,
      "version": "1.0.0-dev",
      "allowedTools": ["Read", "Write", "Edit", "Bash", "Glob"],
      "permissions": {
        "fileAccess": ".moai/specs/**",
        "fsWrite": true,
        "maxProjectSize": 100
      }
    }
  }
}
```

## Troubleshooting

**Question**: Why is charter.md not created?
**Answer**: Use `/init-pm project-name --skip-charter` if you want to skip it. Otherwise, it's created by default.

**Question**: How many risks are in the matrix?
**Answer**: Depends on risk level: low (3 risks), medium (6 risks), high (10+ risks)

**Question**: Can I modify the generated SPEC files?
**Answer**: Yes! Generated files are templates. Edit them as needed before running `/alfred:2-run`.

**Question**: What's the relationship between spec.md and plan.md?
**Answer**: spec.md defines WHAT (requirements), plan.md defines HOW (implementation timeline)

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for contribution guidelines.

## Support

For issues, feature requests, or questions:
- GitHub Issues: [moai-adk/issues](https://github.com/anthropics/claude-code/issues)
- Documentation: See [USAGE.md](./USAGE.md) for detailed examples

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for version history and release notes.

---

**Created**: 2025-10-30
**Generated with [Claude Code](https://claude.com/claude-code)**
**Co-Authored-By**: 🎩 Alfred <alfred@mo.ai.kr>
