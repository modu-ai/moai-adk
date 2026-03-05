# SPEC-SKILL-001: Acceptance Criteria

| Field       | Value                          |
|-------------|--------------------------------|
| SPEC ID     | SPEC-SKILL-001                 |
| Title       | Open Agent Skills Registry Integration Acceptance Criteria |
| Created     | 2026-03-05                     |
| Author      | manager-spec                   |

---

## 1. CLI Install Acceptance Criteria

### AC-INSTALL-001: Basic Install

**Given** the user is in a project directory with `.claude/` present
**When** the user runs `moai skill install VoltAgent/awesome-agent-skills/react-best-practices`
**Then** the system downloads the SKILL.md from the GitHub repository
**And** creates `.claude/skills/VoltAgent/react-best-practices/SKILL.md`
**And** displays: `✓ Installed react-best-practices from VoltAgent/awesome-agent-skills`
**And** displays usage hint: `Use /react-best-practices in Claude Code to activate`

**Given** the skill directory already exists
**When** the user runs `moai skill install <owner/repo>/<skill>` without `--force`
**Then** the system prompts: `Skill already exists. Overwrite? [y/N]`
**And** aborts if user answers N

**Given** the `--force` flag is provided
**When** the skill directory already exists
**Then** the system overwrites without prompting

### AC-INSTALL-002: Install Failure

**Given** the repository does not exist or the skill path is incorrect
**When** the user runs `moai skill install <bad-owner/bad-repo>`
**Then** the system displays an error with the 3 attempted URL paths
**And** exits with non-zero status code
**And** does NOT create any partial files

### AC-INSTALL-003: Validation on Install

**Given** `validate_on_install: true` in `skill.yaml`
**When** the downloaded skill fails validation
**Then** the system displays validation errors
**And** prompts: `Install anyway? [y/N]` (default: N)

---

## 2. CLI Validate Acceptance Criteria

### AC-VALIDATE-001: Open Standard Validation

**Given** a skill file at `.claude/skills/myskill/SKILL.md` with missing `name` field
**When** the user runs `moai skill validate .claude/skills/myskill/`
**Then** the system displays: `ERROR: [name] Field 'name' is required. Fix: Add 'name: my-skill-name' to frontmatter.`
**And** exits with non-zero status code

**Given** a skill file with `allowed-tools` as YAML array instead of CSV string
**When** the user runs `moai skill validate .claude/skills/myskill/`
**Then** the system displays: `ERROR: [allowed-tools] Must be CSV string, not YAML array. Fix: Use 'allowed-tools: Read, Write, Edit'`

**Given** a fully valid skill file conforming to the open standard
**When** the user runs `moai skill validate .claude/skills/myskill/`
**Then** the system displays: `✓ Skill validation passed (0 errors, 0 warnings)`
**And** exits with status code 0

### AC-VALIDATE-002: MoAI-Specific Validation

**Given** a MoAI skill with `metadata.version: 1.0.0` (unquoted)
**When** the user runs `moai skill validate --moai .claude/skills/moai/my-skill.md`
**Then** the system displays: `WARNING: [metadata.version] Value should be a quoted string. Fix: Use version: "1.0.0"`

---

## 3. CLI List Acceptance Criteria

### AC-LIST-001: List Installed Skills

**Given** a project with built-in MoAI skills and 2 community skills installed
**When** the user runs `moai skill list`
**Then** the system displays a table with columns: Name, Source, Version, Description
**And** built-in skills are grouped under `moai-builtin`
**And** community skills are grouped under `community`

### AC-LIST-002: List Registry Skills

**Given** a valid registry `VoltAgent/awesome-agent-skills`
**When** the user runs `moai skill list --registry VoltAgent/awesome-agent-skills`
**Then** the system fetches the registry index and displays available skills
**And** each row shows: Name, Description, Install Command

---

## 4. CLI Search Acceptance Criteria

### AC-SEARCH-001: Search Default Registry

**Given** the default registry is configured
**When** the user runs `moai skill search react`
**Then** the system returns skills with "react" in name or description
**And** each result shows the install command

**Given** no results match the query
**When** the user runs `moai skill search <nonexistent>`
**Then** the system displays: `No skills found for "<nonexistent>". Try broader search terms.`

---

## 5. Publish Acceptance Criteria

### AC-PUBLISH-001: Validate Before Publish

**Given** a skill that passes all validation checks
**When** the user runs `moai skill publish my-skill`
**Then** the system displays the publish checklist:
```
✓ Skill validation passed
✓ moai.yaml manifest generated
📋 Publish steps:
  1. Push to GitHub: git push origin main
  2. Share: moai skill install <your-username>/<repo>/my-skill
  3. Submit to registry: https://github.com/VoltAgent/awesome-agent-skills/pulls
```

**Given** the skill fails validation
**When** the user runs `moai skill publish my-skill`
**Then** the system displays all errors with fix suggestions
**And** does NOT generate or display the publish checklist

---

## 6. Cross-Platform Compatibility Acceptance Criteria

### AC-COMPAT-001: Open Standard Fields

**Given** a MoAI-authored skill published via `moai skill publish`
**When** the published SKILL.md is installed in Cursor, Copilot, or Codex
**Then** these agents can parse and load the skill without errors
**And** MoAI-specific extension fields (triggers, progressive_disclosure) are silently ignored

### AC-COMPAT-002: Community Skill Import

**Given** a community skill designed for Cursor with standard SKILL.md format
**When** the user installs it via `moai skill install <owner/repo>/<skill>`
**Then** the skill is functional in Claude Code / MoAI-ADK
**And** no transformation errors occur during installation

### AC-COMPAT-003: Namespace Isolation

**Given** a community skill named `react-best-practices` from `VoltAgent/awesome-agent-skills`
**And** a MoAI built-in skill also named `react-best-practices` (hypothetical)
**When** both are installed
**Then** community skill is at `.claude/skills/VoltAgent/react-best-practices/`
**And** MoAI skill remains at `.claude/skills/moai/react-best-practices.md`
**And** no file collision occurs

---

## 7. Registry Acceptance Criteria

### AC-REGISTRY-001: URL Resolution

**Given** the user installs `owner/repo` (without specific skill name)
**When** the registry client resolves the URL
**Then** it tries in order:
1. `https://raw.githubusercontent.com/owner/repo/main/SKILL.md`
2. `https://raw.githubusercontent.com/owner/repo/main/repo/SKILL.md`
3. `https://raw.githubusercontent.com/owner/repo/main/skills/repo/SKILL.md`
**And** uses the first successful response (HTTP 200)
**And** displays a clear error if all 3 fail

### AC-REGISTRY-002: Configuration

**Given** `.moai/config/sections/skill.yaml` with custom registries
**When** the user runs `moai skill search <query>`
**Then** the system searches all configured registries (not just the default)
**And** results are merged and deduplicated by skill name + owner

---

## 8. Manifest Acceptance Criteria

### AC-MANIFEST-001: `moai.yaml` Generation

**Given** a skill package directory with multiple SKILL.md files
**When** the user runs `moai skill publish` and answers Y to "Generate moai.yaml?"
**Then** the system creates `moai.yaml` with all discovered skills listed
**And** the manifest follows the REQ-MANIFEST-001 format
**And** version is set to `"1.0.0"` by default
