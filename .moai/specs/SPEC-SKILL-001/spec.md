# SPEC-SKILL-001: 오픈 에이전트 스킬 레지스트리 통합

| Field       | Value                                                      |
|-------------|-----------------------------------------------------------|
| SPEC ID     | SPEC-SKILL-001                                            |
| Title       | Open Agent Skills Registry Integration                     |
| Status      | Draft                                                      |
| Created     | 2026-03-05                                                |
| Author      | manager-spec                                               |
| Priority    | Medium                                                     |
| Lifecycle   | spec-anchored                                             |
| Branch      | feature/skill-registry                                    |
| Resolves    | #470                                                       |

---

## 1. Overview

### 1.1 Problem Statement

MoAI-ADK has a rich internal skill ecosystem (50+ skills), but these skills are:

1. **Isolated** — skills only work within MoAI-ADK projects initialized via `moai init`; they cannot be shared with the broader AI coding community
2. **Undiscoverable** — there is no public registry where MoAI skills can be found or installed by users of other AI agents (Cursor, GitHub Copilot, OpenAI Codex, etc.)
3. **Platform-specific** — MoAI skill frontmatter does not fully conform to the emerging open Agent Skills standard (cross-platform SKILL.md format), preventing interoperability with other platforms
4. **Manually managed** — users must copy skill files manually; there is no `moai skill install <package>` CLI workflow

As of 2026, an open Agent Skills standard has emerged (championed by Anthropic and adopted by Cursor, GitHub Copilot, Codex, Gemini CLI, and others) with 500+ community skills in the wild. MoAI-ADK needs to participate in this ecosystem.

### 1.2 Solution

Integrate MoAI-ADK with the open Agent Skills ecosystem by:

1. **CLI Skill Manager**: Add `moai skill` subcommand (install/publish/list/search) to the `moai` binary
2. **Cross-Platform Compatibility**: Ensure MoAI skills conform to the open SKILL.md standard so they work with Cursor, Copilot, Codex, and other agents
3. **Public Registry Adapter**: Connect to community registries (GitHub-based, skills.sh, or similar) for skill discovery and distribution
4. **Skill Marketplace Template**: Add a skill catalog and discovery workflow to MoAI-ADK projects

### 1.3 Prior Art

- **Vercel Agent Skills**: Open standard SKILL.md format; `npx skills i vercel-labs/agent-skills`
- **Supabase Agent Skills**: Domain-specific skills for Postgres best practices
- **VoltAgent/awesome-agent-skills**: 500+ community skills repository
- **antigravity.codes**: Skills registry with 500+ cross-platform skills
- **MoAI-ADK Internal Skills**: 50+ internal skills in `.claude/skills/moai/`

### 1.4 Scope

**In Scope:**
- `moai skill` CLI subcommand (install, publish, list, search, validate)
- Cross-platform SKILL.md format compatibility layer
- GitHub-based skill registry adapter (install from `owner/repo` pattern)
- Skill validation against open standard
- Template updates for skill discovery workflow (`moai-skill` skill)
- `moai.yaml` skill manifest for publishable skill packages

**Out of Scope:**
- Building a dedicated MoAI skill registry server (use existing GitHub-based distribution)
- IDE plugins for skill browsing
- Paid/commercial skill marketplace features
- Automatic skill updates/subscriptions
- Skill version pinning and lock files (future work)

---

## 2. Environment

### 2.1 Platform

- MoAI-ADK Go Edition v2.8.0+
- Go 1.22+
- Claude Code v2.1.50+
- Targets: macOS (arm64, amd64), Linux (arm64, amd64), Windows (amd64)

### 2.2 Integration Points

| System                        | Integration Type     | Purpose                                      |
|-------------------------------|----------------------|----------------------------------------------|
| `cmd/moai/main.go`            | CLI subcommand       | Add `moai skill` command router               |
| `internal/cli/skill.go`       | CLI handler (new)    | Implement install/publish/list/search         |
| `internal/skill/` (new)       | Core package (new)   | Skill registry client, validator, installer   |
| `.claude/skills/moai/`        | Template files       | Skill discovery workflow (`moai-skill` skill) |
| GitHub API / raw.githubusercontent| External HTTP      | Skill package download from `owner/repo`      |
| `.moai/config/sections/skill.yaml`| Configuration     | Registry URLs, install paths, validation opts |

### 2.3 Dependencies

| Dependency           | Version     | Purpose                                 |
|----------------------|-------------|-----------------------------------------|
| `net/http`           | std         | HTTP client for registry requests       |
| `gopkg.in/yaml.v3`  | existing    | SKILL.md frontmatter parsing            |
| `github.com/spf13/cobra` | existing | CLI subcommand integration          |
| GitHub API (raw)     | n/a         | Download skills from `owner/repo` paths |

---

## 3. Assumptions

- A1: The open Agent Skills standard uses a SKILL.md file with YAML frontmatter as the primary format (compatible with the existing MoAI format).
- A2: Community skills are distributed via GitHub repositories in an `owner/repo` pattern (matching how Vercel, Supabase, and VoltAgent distribute their skills).
- A3: MoAI internal skills (in `.claude/skills/moai/`) are authored in a format that is largely compatible with the open standard with minor adaptations.
- A4: Users have internet access for skill installation; offline mode is not required in v1.
- A5: Skill installation places files in `.claude/skills/` of the current project (same location as moai init populates).
- A6: Published skills must include a `moai.yaml` manifest (name, version, description, license, compatible-agents).
- A7: The `moai skill validate` command can be run without internet access (validates local skill files only).

---

## 4. Requirements

### 4.1 CLI Skill Manager

#### REQ-CLI-001: `moai skill install` Command

**WHEN** the user runs `moai skill install <owner/repo>` or `moai skill install <owner/repo>/<skill-name>`,
**THEN** the system **shall** download the SKILL.md (and any referenced files) from the GitHub repository and place it in `.claude/skills/<skill-name>/`.

**WHEN** the user runs `moai skill install <owner/repo>` where the repository contains a `moai.yaml` manifest,
**THEN** the system **shall** install all skills listed in the manifest's `skills` array.

**WHEN** the target skill directory already exists,
**THEN** the system **shall** prompt the user to confirm overwrite (or use `--force` flag to skip).

**WHEN** the download fails (network error, 404, rate limit),
**THEN** the system **shall** display a clear error message with the HTTP status and suggest checking the repository path.

#### REQ-CLI-002: `moai skill publish` Command

**WHEN** the user runs `moai skill publish <skill-name>`,
**THEN** the system **shall** validate the skill against the open standard (REQ-VAL-001) and display a publish checklist with instructions for distributing via GitHub.

**WHEN** the skill validation passes,
**THEN** the system **shall** generate a `moai.yaml` manifest file if one does not exist.

**WHEN** the skill validation fails,
**THEN** the system **shall** display specific validation errors with fix suggestions.

#### REQ-CLI-003: `moai skill list` Command

**WHEN** the user runs `moai skill list`,
**THEN** the system **shall** display all skills installed in the current project's `.claude/skills/` directory, grouped by source (moai-builtin / community).

**WHEN** the user runs `moai skill list --registry <owner/repo>`,
**THEN** the system **shall** fetch and display all available skills from the specified registry repository.

#### REQ-CLI-004: `moai skill search` Command

**WHEN** the user runs `moai skill search <query>`,
**THEN** the system **shall** search the configured registries for skills matching the query and display results with name, description, and install command.

**WHEN** no registries are configured,
**THEN** the system **shall** search the default registry (VoltAgent/awesome-agent-skills or equivalent).

#### REQ-CLI-005: `moai skill validate` Command

**WHEN** the user runs `moai skill validate <skill-path>`,
**THEN** the system **shall** validate the SKILL.md against the open Agent Skills standard and report any compatibility issues.

### 4.2 Cross-Platform Compatibility

#### REQ-COMPAT-001: Open Standard SKILL.md Format

**WHEN** a MoAI skill is published,
**THEN** the published SKILL.md **shall** conform to the open Agent Skills standard with these required fields:

```yaml
---
name: <skill-name>
version: "1.0.0"
description: >
  <description>
license: Apache-2.0
compatibility:
  - claude-code
  - cursor
  - codex
  - gemini-cli
allowed-tools: <comma-separated list>
---
```

**WHEN** a community skill is installed,
**THEN** the system **shall** detect the frontmatter format and adapt it to MoAI-compatible format if needed.

#### REQ-COMPAT-002: MoAI Extension Fields

**WHEN** a skill uses MoAI-specific fields (progressive_disclosure, triggers, metadata),
**THEN** these fields **shall** be treated as MoAI extensions and MUST NOT break compatibility with other agents that ignore unknown YAML fields.

#### REQ-COMPAT-003: Skill Namespace Isolation

**WHEN** a community skill is installed,
**THEN** it **shall** be placed in `.claude/skills/<owner>/<skill-name>/` to avoid namespace collisions with built-in MoAI skills in `.claude/skills/moai/`.

### 4.3 Skill Registry Adapter

#### REQ-REG-001: GitHub-Based Registry Protocol

**WHEN** resolving a skill reference `<owner>/<repo>`,
**THEN** the system **shall** attempt to fetch from:
1. `https://raw.githubusercontent.com/<owner>/<repo>/main/SKILL.md`
2. `https://raw.githubusercontent.com/<owner>/<repo>/main/<repo>/SKILL.md`
3. `https://raw.githubusercontent.com/<owner>/<repo>/main/skills/<repo>/SKILL.md`

**WHEN** all paths fail,
**THEN** the system **shall** display the attempted paths and suggest the user verify the repository structure.

#### REQ-REG-002: Registry Configuration

**WHEN** `.moai/config/sections/skill.yaml` exists with a `registries` list,
**THEN** the `moai skill search` command **shall** query all listed registries.

Default configuration:
```yaml
skill:
  registries:
    - owner: VoltAgent
      repo: awesome-agent-skills
      description: "Community skills registry (500+)"
  install_path: ".claude/skills"
  validate_on_install: true
  allow_community_skills: true
```

### 4.4 Skill Validation

#### REQ-VAL-001: Open Standard Validation Rules

**WHEN** validating a skill,
**THEN** the system **shall** check:

- `name` field: present, lowercase, hyphens only (no spaces)
- `description` field: present, minimum 20 characters
- `allowed-tools` field: present, comma-separated (NOT YAML array)
- `version` field: present, semver format
- `license` field: present (SPDX identifier recommended)
- SKILL.md body: present and non-empty
- No sensitive patterns (API keys, tokens) in skill content

#### REQ-VAL-002: MoAI-Specific Validation

**WHEN** validating a MoAI internal skill,
**THEN** the system **shall** additionally check:
- `metadata.version` is quoted string (per CLAUDE.local.md rule)
- `allowed-tools` uses CSV format (NOT YAML array)
- Progressive disclosure levels are correctly defined if present

### 4.5 Skill Package Manifest

#### REQ-MANIFEST-001: `moai.yaml` Format

**WHEN** a user publishes a skill package containing multiple skills,
**THEN** the package root **shall** contain a `moai.yaml` manifest:

```yaml
name: my-skill-package
version: "1.0.0"
description: "Collection of skills for X use case"
license: MIT
author: "username"
skills:
  - path: skills/skill-one
    name: skill-one
  - path: skills/skill-two
    name: skill-two
compatible-agents:
  - claude-code
  - cursor
  - codex
```

---

## 5. Specifications

### 5.1 Implementation Architecture

```
moai-adk-go/
├── cmd/moai/main.go                    # Add "skill" subcommand
├── internal/
│   ├── cli/
│   │   └── skill.go                    # CLI handler: install/publish/list/search/validate
│   └── skill/                          # NEW PACKAGE
│       ├── registry.go                 # GitHub registry client (HTTP)
│       ├── installer.go               # Skill download & installation
│       ├── validator.go               # Open standard + MoAI validation
│       ├── manifest.go                # moai.yaml parsing & generation
│       └── skill_test.go              # Tests for all above
└── internal/template/templates/
    ├── .claude/skills/moai/
    │   └── workflows/skill.md         # NEW: moai-skill workflow skill
    └── .moai/config/sections/
        └── skill.yaml                 # NEW: default skill registry config
```

### 5.2 CLI Command Structure

```
moai skill <subcommand> [args] [flags]

Subcommands:
  install <owner/repo>[/<skill-name>]  Install a skill from GitHub
  publish <skill-name>                 Validate and prepare a skill for publishing
  list [--registry <owner/repo>]       List installed or available skills
  search <query>                       Search skills in configured registries
  validate <path>                      Validate a skill against the open standard

Flags (for install):
  --force                              Overwrite existing skill without prompt
  --dry-run                            Show what would be installed without doing it

Flags (for list):
  --registry <owner/repo>              List skills from a specific registry
  --format json|table                  Output format (default: table)
```

### 5.3 HTTP Client Design

```go
// internal/skill/registry.go

type RegistryClient struct {
    BaseURL    string
    HTTPClient *http.Client
    RateLimit  time.Duration
}

// Resolve resolves a skill reference to a download URL
// Tries multiple path conventions (SKILL.md, <repo>/SKILL.md, skills/<repo>/SKILL.md)
func (c *RegistryClient) Resolve(owner, repo, skillName string) (string, error)

// Download fetches the SKILL.md content from a resolved URL
func (c *RegistryClient) Download(url string) ([]byte, error)
```

### 5.4 Validator Design

```go
// internal/skill/validator.go

type ValidationResult struct {
    Valid    bool
    Errors   []ValidationError
    Warnings []ValidationWarning
}

type ValidationError struct {
    Field   string
    Message string
    Fix     string // suggested fix
}

// Validate checks a SKILL.md against the open Agent Skills standard
func Validate(skillPath string) ValidationResult

// ValidateMoAI checks additional MoAI-specific rules
func ValidateMoAI(skillPath string) ValidationResult
```

### 5.5 New Template Files

#### `.claude/skills/moai/workflows/skill.md`

A new user-invocable skill that wraps the `moai skill` CLI:

```yaml
---
name: moai-skill
description: >
  Agent Skills registry integration for MoAI-ADK.
  Install, publish, list, and search community skills.
  Invoke: /moai skill install <owner/repo>
user-invocable: true
argument-hint: "[install|publish|list|search|validate] [args]"
allowed-tools: Bash, Read, Write, Glob
metadata:
  version: "1.0.0"
  category: "workflow"
---
```

#### `.moai/config/sections/skill.yaml`

Default registry configuration included in every `moai init` project.

### 5.6 Installation Flow

```
User: moai skill install VoltAgent/awesome-agent-skills/react-best-practices

1. Parse: owner=VoltAgent, repo=awesome-agent-skills, skill=react-best-practices
2. Resolve URL: try 3 path conventions
3. Download SKILL.md (and referenced files)
4. Validate (if validate_on_install: true)
5. Create: .claude/skills/VoltAgent/react-best-practices/SKILL.md
6. Display: "✓ Installed react-best-practices from VoltAgent/awesome-agent-skills"
7. Hint: "Use /react-best-practices in Claude Code to activate"
```

### 5.7 Cross-Platform Compatibility Matrix

| Field              | MoAI Format          | Open Standard         | Action                    |
|--------------------|----------------------|-----------------------|---------------------------|
| `name`             | name: moai-xxx       | name: xxx             | Keep as-is (compatible)   |
| `description`      | YAML folded scalar   | YAML folded scalar    | Compatible                |
| `allowed-tools`    | CSV string           | CSV string            | Compatible                |
| `version`          | In metadata block    | Top-level field       | Add top-level on publish  |
| `license`          | Not required         | Required              | Add on publish            |
| `compatibility`    | Not present          | Recommended           | Add on publish            |
| `triggers`         | MoAI extension       | Not standard          | Keep (ignored by others)  |
| `progressive_disclosure` | MoAI extension | Not standard         | Keep (ignored by others)  |

---

## 6. Out of Scope

- **OS-001**: Building a dedicated MoAI skill registry server (HTTP backend)
- **OS-002**: Skill version pinning and lock files (post-MVP)
- **OS-003**: Automatic skill update notifications
- **OS-004**: IDE browser UI for skill discovery
- **OS-005**: Paid or commercial skill licensing
- **OS-006**: Skill dependency resolution (skills that depend on other skills)
- **OS-007**: Windows-specific path handling beyond standard Go `filepath` behavior

---

## 7. Open Questions

- **OQ-001**: Which community registry should be the default? `VoltAgent/awesome-agent-skills` (500+ skills) or `sickn33/antigravity-awesome-skills` (900+ skills)? **Proposed**: Use `VoltAgent/awesome-agent-skills` as default; allow users to add others via `skill.yaml`.
- **OQ-002**: Should `moai skill install` require `moai init` to have been run first? **Proposed**: No, allow installation in any directory with a `.claude/` folder.
- **OQ-003**: Should MoAI publish its own 50+ internal skills to a public registry? **Proposed**: Yes, create `modu-ai/moai-skills` GitHub repository as the official MoAI skills registry.
- **OQ-004**: How should skill conflicts be handled when two installed skills have the same name? **Proposed**: Namespace by `owner/skill-name`; warn on conflict and require `--force`.
- **OQ-005**: Should the validator enforce the `compatibility` field listing Claude Code? **Proposed**: Warn (not error) if `compatibility` is missing; error if present but missing `claude-code`.

---

## 8. Traceability

| Requirement     | Acceptance Criteria | Milestone       |
|-----------------|---------------------|-----------------|
| REQ-CLI-001     | AC-INSTALL-001      | MILE-1: CLI     |
| REQ-CLI-002     | AC-PUBLISH-001      | MILE-2: Publish |
| REQ-CLI-003     | AC-LIST-001         | MILE-1: CLI     |
| REQ-CLI-004     | AC-SEARCH-001       | MILE-2: Search  |
| REQ-CLI-005     | AC-VALIDATE-001     | MILE-1: CLI     |
| REQ-COMPAT-001  | AC-COMPAT-001       | MILE-1: CLI     |
| REQ-COMPAT-002  | AC-COMPAT-002       | MILE-1: CLI     |
| REQ-COMPAT-003  | AC-COMPAT-003       | MILE-1: CLI     |
| REQ-REG-001     | AC-REGISTRY-001     | MILE-1: CLI     |
| REQ-REG-002     | AC-REGISTRY-002     | MILE-2: Config  |
| REQ-VAL-001     | AC-VALIDATE-001     | MILE-1: CLI     |
| REQ-VAL-002     | AC-VALIDATE-002     | MILE-1: CLI     |
| REQ-MANIFEST-001| AC-MANIFEST-001     | MILE-2: Publish |

---

## 9. References

- [VoltAgent/awesome-agent-skills](https://github.com/VoltAgent/awesome-agent-skills) — 500+ community skills
- [sickn33/antigravity-awesome-skills](https://github.com/sickn33/antigravity-awesome-skills) — 900+ skills
- [Cursor Docs: Agent Skills](https://cursor.com/docs/context/skills)
- [antigravity.codes: Agent Skills Hub](https://antigravity.codes/agent-skills)
- [Killer Skills Registry](https://killer-skills.com/en/blog/best-ai-agent-skills-2026/)
