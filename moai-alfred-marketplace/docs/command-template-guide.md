# Command Template Guide

**How to write command templates for Alfred Framework plugins**

## Overview

Command templates define user-facing commands for plugins. Each command:
- Lives in `commands/{command-name}.md`
- Describes syntax, arguments, options, examples
- Specifies output and related skills
- Gets invoked by users as `/command-name`

## Template Structure

```markdown
# /command-name

One-line description of what command does.

## Syntax

\`\`\`bash
/command-name [required-arg] [--optional-flag value]
\`\`\`

## Arguments

- **arg-name** (required|optional): Description

## Options

- `--flag` (optional): Description

## Examples

\`\`\`bash
/command-name example-value
\`\`\`

## What it does

1. Step 1
2. Step 2
3. Step 3

## Output

Description of what gets created/modified.

## Related

- Related skills or guides
```

## Complete Example: PM Plugin Command

File: `commands/init-pm.md`

```markdown
# /init-pm

Initialize project management templates with EARS SPEC framework.

## Syntax

\`\`\`bash
/init-pm <project-name> [options]
\`\`\`

## Arguments

- **project-name** (required): Project identifier (e.g., `ecommerce-platform`)
  - Format: lowercase, hyphens allowed
  - Used to generate SPEC file names (SPEC-PROJECT-NAME-001)

## Options

- `--template` (optional): SPEC template to use
  - Values: `moai-spec` (default), `enterprise`, `agile`
  - Default: `moai-spec`

- `--skip-charter` (optional): Skip project charter generation
  - No value required

- `--risk-level` (optional): Risk assessment level
  - Values: `low`, `medium`, `high`
  - Default: `medium`

## Examples

### Basic Usage
```bash
/init-pm my-awesome-project
```
Creates:
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/spec.md`
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/plan.md`
- `.moai/specs/SPEC-MY-AWESOME-PROJECT-001/acceptance.md`

### With Custom Template
```bash
/init-pm ecommerce-platform --template=enterprise
```
Uses enterprise SPEC template with additional governance sections.

### Skip Charter Generation
```bash
/init-pm api-service --skip-charter
```
Skips project charter, creates SPEC documents only.

### High Risk Assessment
```bash
/init-pm payment-system --risk-level=high
```
Generates additional risk matrix and mitigation planning sections.

## What it does

1. **Validates Project Name**
   - Checks for valid format (lowercase, hyphens)
   - Ensures no SPEC conflicts with existing projects

2. **Creates SPEC Directory**
   - Creates `.moai/specs/SPEC-{PROJECT}-001/` directory
   - Initializes with template files

3. **Generates SPEC Documents**
   - spec.md: EARS requirement specification
   - plan.md: Implementation plan
   - acceptance.md: Acceptance criteria

4. **Creates Project Charter** (unless `--skip-charter`)
   - charter.md: Project governance
   - stakeholders.json: Stakeholder matrix
   - timeline.json: Project milestones

5. **Builds Risk Matrix** (based on `--risk-level`)
   - risk-matrix.json: Risk assessment
   - mitigation-plan.md: Risk mitigation strategies

6. **Displays Summary**
   - Shows created files
   - Lists next steps (run `/alfred:2-run`)

## Output

Creates `.moai/specs/SPEC-{PROJECT}/` directory structure:

```
.moai/specs/SPEC-MY-PROJECT-001/
├── spec.md                  # EARS requirements
├── plan.md                  # Implementation plan
├── acceptance.md            # Acceptance criteria
├── charter.md               # Project charter (optional)
├── stakeholders.json        # Stakeholder matrix
├── timeline.json            # Project timeline
├── risk-matrix.json         # Risk assessment
└── mitigation-plan.md       # Risk mitigation (optional)
```

### Success Output
```
✅ Project initialization complete

📋 SPEC Documents Created
- .moai/specs/SPEC-MY-PROJECT-001/spec.md
- .moai/specs/SPEC-MY-PROJECT-001/plan.md
- .moai/specs/SPEC-MY-PROJECT-001/acceptance.md

📊 Additional Files
- charter.md (project governance)
- risk-matrix.json (risk assessment)
- mitigation-plan.md (mitigation strategies)

🚀 Next Steps
1. Review generated SPEC documents
2. Run `/alfred:2-run SPEC-MY-PROJECT-001` to implement
3. Check `.moai/specs/SPEC-MY-PROJECT-001/` for full details
```

## Related

- `moai-foundation-ears` - EARS requirement syntax
- `moai-spec-authoring` - SPEC document writing guide
- [SPEC-CH08-001](../../.moai/specs/SPEC-CH08-001/spec.md) - Plugin architecture

## Error Handling

### Invalid Project Name
```
❌ Error: Project name must use lowercase letters and hyphens
Invalid name: "MyAwesomeProject"
Valid format: "my-awesome-project"
```

### SPEC Already Exists
```
❌ Error: SPEC-MY-PROJECT-001 already exists
Location: .moai/specs/SPEC-MY-PROJECT-001/

Options:
- Use different project name: /init-pm my-new-project
- Delete existing SPEC: rm -rf .moai/specs/SPEC-MY-PROJECT-001/
- Increment version: /init-pm my-project-v2
```

### Permission Denied
```
❌ Error: Permission denied: Cannot create .moai/specs/ directory
Make sure you have write access to project root directory
```
```

## Example 2: Frontend Plugin Command

File: `commands/init-next.md`

```markdown
# /init-next

Initialize Next.js 16 project with React 19.2, TypeScript, and Biome.

## Syntax

\`\`\`bash
/init-next <app-name> [--pm bun|npm|pnpm] [--ts|--js] [--git]
\`\`\`

## Arguments

- **app-name** (required): Application name (e.g., `my-app`)
  - Lowercase, hyphens allowed
  - Used as project directory name

## Options

- `--pm` (optional): Package manager
  - Values: `bun` (default), `npm`, `pnpm`
  - Recommended: `bun` (fastest)

- `--ts` (optional): Use TypeScript
  - Default: true (TypeScript enabled)

- `--js` (optional): Use JavaScript instead of TypeScript
  - Overrides `--ts` flag

- `--git` (optional): Initialize git repository
  - Default: true (git initialized)

- `--no-git` (optional): Skip git initialization
  - Skips `git init`

## Examples

### Default Setup (Bun + TypeScript)
\`\`\`bash
/init-next ecommerce-app
\`\`\
Creates Next.js 16 app with:
- Bun package manager
- TypeScript
- Biome linter/formatter
- Git repository

### With npm and JavaScript
\`\`\`bash
/init-next simple-app --pm=npm --js
\`\`\`

### With pnpm (No Git)
\`\`\`bash
/init-next monorepo-app --pm=pnpm --no-git
\`\`\`

## What it does

1. **Create Project Directory**
   - mkdir {app-name}
   - Initialize with Next.js 16 boilerplate

2. **Install Dependencies**
   - Uses specified package manager (bun/npm/pnpm)
   - Installs: next, react, typescript, biome, shadcn/ui

3. **Setup TypeScript** (if --ts)
   - Creates tsconfig.json
   - Configures strict type checking
   - Sets up path aliases (@/*)

4. **Configure Biome**
   - biome.json with formatting rules
   - Pre-commit hooks
   - ESLint/Prettier replacement

5. **Integrate shadcn/ui**
   - Setup Tailwind CSS
   - Install core components
   - Configure component paths

6. **Initialize Git** (if --git)
   - git init
   - Create .gitignore
   - Initial commit: "chore: Initialize Next.js project"

## Output

Creates project structure:

\`\`\`
ecommerce-app/
├── .next/
├── app/
│   ├── layout.tsx
│   └── page.tsx
├── components/
│   └── ui/  (shadcn/ui components)
├── public/
├── styles/
├── biome.json
├── next.config.js
├── package.json
├── tsconfig.json
├── .gitignore
└── README.md
\`\`\`

## Related

- `moai-lang-nextjs-advanced` - Next.js 16 patterns
- `moai-lang-typescript` - TypeScript best practices
- `biome-setup` - Configure Biome separately

## Error Handling

### App Already Exists
\`\`\`
❌ Error: Directory 'ecommerce-app' already exists
Options:
- Use different app name: /init-next my-other-app
- Remove existing: rm -rf ecommerce-app
\`\`\`

### Invalid Package Manager
\`\`\`
❌ Error: Package manager 'yarn' not supported
Supported: bun (default), npm, pnpm
Use: /init-next my-app --pm=bun
\`\`\`

### Package Manager Not Installed
\`\`\`
⚠️ Warning: Bun not found. Installing...
Or manually install: curl https://bun.sh | bash
\`\`\`
```

## Template Best Practices

### 1. Clear Syntax

✅ **Good**:
```bash
/command-name <required> [--optional value]
```

❌ **Bad**:
```bash
/command-name
```

### 2. Descriptive Arguments

✅ **Good**:
```markdown
- **project-name** (required): Project identifier for SPEC generation
  - Format: lowercase, hyphens allowed
  - Example: `my-awesome-project`
```

❌ **Bad**:
```markdown
- name (required)
```

### 3. Multiple Examples

✅ **Good**: Show 3-4 different use cases with options

❌ **Bad**: Show only basic usage

### 4. Clear Output Description

✅ **Good**: Show exact directory structure and file contents

❌ **Bad**: "Creates project files"

### 5. Error Scenarios

✅ **Good**: Document common errors and solutions

❌ **Bad**: No error handling mentioned

## Common Patterns

### Pattern 1: Scaffolding Command
```markdown
# /init-{framework}

Initialize {framework} project with best practices.

## Syntax
\`\`\`bash
/init-{framework} <app-name> [--template name]
\`\`\`

## Output
Creates `app-name/` directory with:
- src/
- tests/
- package.json
- tsconfig.json (if TypeScript)
```

### Pattern 2: Configuration Command
```markdown
# /setup-{tool}

Configure {tool} for the project.

## Syntax
\`\`\`bash
/setup-{tool} [--strict] [--format]
\`\`\`

## Output
Creates/updates config files:
- {tool}.config.json
- .{tool}rc
```

### Pattern 3: Generation Command
```markdown
# /generate-{resource}

Generate {resource} from SPEC.

## Syntax
\`\`\`bash
/generate-{resource} <spec-id> [options]
\`\`\`

## Output
Creates:
- src/{resource}/
- tests/{resource}/
```

## See Also

- [Contributor Guide](../CONTRIBUTING.md)
- [plugin.json Schema](./plugin-json-schema.md)
- [SPEC-CH08-001](../../.moai/specs/SPEC-CH08-001/spec.md)

---

**Version**: 1.0.0
**Last Updated**: 2025-10-30

🔗 Generated with [Claude Code](https://claude.com/claude-code)
