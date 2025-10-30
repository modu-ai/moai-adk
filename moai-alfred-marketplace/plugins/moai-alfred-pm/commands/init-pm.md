# /init-pm

Initialize project management templates with EARS SPEC framework.

## Syntax

```bash
/init-pm <project-name> [options]
```

## Arguments

- **project-name** (required): Project identifier (e.g., `ecommerce-platform`)
  - Format: lowercase letters, numbers, hyphens
  - Length: 3-50 characters
  - Used to generate SPEC ID: `SPEC-{PROJECT}-001`

## Options

- `--template` (optional): SPEC template to use
  - Values: `moai-spec` (default), `enterprise`, `agile`
  - Default: `moai-spec`

- `--risk-level` (optional): Risk assessment level
  - Values: `low`, `medium` (default), `high`
  - Determines number of risks in risk-matrix.json

- `--skip-charter` (optional): Skip project charter generation
  - No value required
  - Only generates spec.md, plan.md, acceptance.md

## Examples

### Basic Usage
```bash
/init-pm my-awesome-project
```

Creates:
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/spec.md`
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/plan.md`
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/acceptance.md`
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/charter.md`
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/risk-matrix.json`

### With Custom Template
```bash
/init-pm ecommerce-platform --template=enterprise
```

Uses enterprise template with governance sections.

### High Risk Assessment
```bash
/init-pm payment-system --risk-level=high
```

Generates comprehensive risk matrix with 10+ identified risks.

### Skip Charter Generation
```bash
/init-pm api-service --skip-charter
```

Creates only SPEC documents (no charter.md).

## What it does

1. **Validates Project Name**
   - Checks format (lowercase, hyphens only)
   - Verifies length (3-50 characters)
   - Ensures no existing SPEC with same ID

2. **Creates SPEC Directory**
   - Creates `.moai/specs/SPEC-{PROJECT}-001/` directory
   - Initializes with template files

3. **Generates SPEC Documents**
   - **spec.md**: EARS requirement specification with 5 patterns
     * Ubiquitous behaviors
     * Event-driven behaviors
     * State-driven behaviors
     * Optional behaviors
     * Unwanted behaviors
   - **plan.md**: Implementation plan with 5 phases
   - **acceptance.md**: Acceptance criteria and quality metrics

4. **Creates Project Charter** (unless `--skip-charter`)
   - **charter.md**: Project governance
   - Stakeholder matrix
   - Budget and schedule
   - Decision authority

5. **Builds Risk Matrix** (based on `--risk-level`)
   - **risk-matrix.json**: Risk assessment data
     * Low: 3 risks
     * Medium: 6 risks
     * High: 10+ risks
   - Risk fields: ID, description, probability, impact, mitigation

6. **Displays Summary**
   - Shows created files
   - Provides next steps

## Output

Creates `.moai/specs/SPEC-{PROJECT}/` directory:

```
.moai/specs/SPEC-MY-PROJECT-001/
├── spec.md (EARS-formatted specification)
├── plan.md (5-phase implementation plan)
├── acceptance.md (Acceptance criteria)
├── charter.md (Project governance, unless --skip-charter)
└── risk-matrix.json (Risk assessment)
```

### Success Output
```
✅ Project 'my-project' initialized successfully
📁 Location: .moai/specs/SPEC-MY-PROJECT-001/
📊 Risk Level: medium
📋 Template: moai-spec
📝 Files created: 5
```

## Related

- `moai-foundation-ears` - EARS requirement syntax (5 patterns)
- `moai-spec-authoring` - SPEC document writing guide
- `/alfred:2-run SPEC-MY-PROJECT-001` - Run SPEC implementation

## Error Handling

### Invalid Project Name
```
❌ Project name must contain only lowercase letters, numbers, and hyphens
Invalid: "MyAwesomeProject", "my awesome project"
Valid: "my-awesome-project"
```

### SPEC Already Exists
```
❌ SPEC already exists: .moai/specs/SPEC-MY-PROJECT-001/
Options:
- Use different name: /init-pm my-project-v2
- Remove existing: rm -rf .moai/specs/SPEC-MY-PROJECT-001/
```

### Invalid Risk Level
```
❌ Invalid risk level: extreme
Supported: low, medium (default), high
```

### Template Not Found
```
❌ Invalid template: nonexistent
Supported: moai-spec (default), enterprise, agile
```

---

Generated with [Claude Code](https://claude.com/claude-code)
