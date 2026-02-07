# SPEC-EMBED-001: Acceptance Criteria

---
spec_id: SPEC-EMBED-001
title: Template Content and Binary Bundling - Acceptance Criteria
status: Planned
created: 2026-02-03
tags: go-embed, templates, bundling, binary, deployment, init
---

## 1. Acceptance Scenarios (Given-When-Then)

---

### ACC-001: Embedded FS Contains All Required Template Files

**Requirement Reference**: REQ-E-001 through REQ-E-009

```gherkin
Feature: Template Content Completeness

  Scenario: Embedded filesystem contains all agent definitions
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain at least 15 files under ".claude/agents/moai/"
    And each agent file shall have a ".md" extension
    And each agent file shall be non-empty

  Scenario: Embedded filesystem contains all skill definitions
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain at least 400 files under ".claude/skills/"
    And each skill directory shall contain a "SKILL.md" file
    And no skill file shall be empty

  Scenario: Embedded filesystem contains all rule files
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain at least 15 files under ".claude/rules/moai/"
    And the files shall be organized in subdirectories (core/, development/, workflow/)

  Scenario: Embedded filesystem contains output styles
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain at least 3 files under ".claude/output-styles/moai/"

  Scenario: Embedded filesystem contains CLAUDE.md
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain "CLAUDE.md" at the root
    And the CLAUDE.md content shall be at least 5000 characters long
    And the CLAUDE.md content shall contain "MoAI Execution Directive"

  Scenario: Embedded filesystem contains MCP configuration
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain ".mcp.json" at the root
    And the FS shall contain ".mcp.windows.json" at the root
    And each JSON file shall be valid JSON

  Scenario: Embedded filesystem contains .gitignore
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain ".gitignore" at the root
    And the .gitignore shall contain ".moai/" pattern

  Scenario: Embedded filesystem contains announcement templates
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain at least 4 files under ".moai/announcements/"
    And each file shall be valid JSON

  Scenario: Embedded filesystem contains LLM config
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall contain at least 1 file under ".moai/llm-configs/"
    And each file shall be valid JSON
```

---

### ACC-002: Excluded Files Are Not Present

**Requirement Reference**: REQ-E-010 through REQ-E-014

```gherkin
Feature: Template Content Exclusion

  Scenario: Python hook scripts are not embedded
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall NOT contain any files under ".claude/hooks/"
    And the FS shall NOT contain any files with ".py" extension

  Scenario: settings.json is not embedded
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall NOT contain ".claude/settings.json"
    And the FS shall NOT contain ".claude/settings.local.json"

  Scenario: Cache and OS files are not embedded
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall NOT contain any "__pycache__" directories
    And the FS shall NOT contain any ".DS_Store" files
    And the FS shall NOT contain any ".pyc" files

  Scenario: LSP config is not embedded
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall NOT contain ".lsp.json"

  Scenario: Git hook scripts are not embedded
    Given the moai binary is compiled with go:embed
    When the embedded FS is loaded via EmbeddedTemplates()
    Then the FS shall NOT contain any files under ".git-hooks/"
```

---

### ACC-003: go:embed Path Resolution

**Requirement Reference**: REQ-E-020 through REQ-E-024

```gherkin
Feature: Embed Directive Integration

  Scenario: EmbeddedTemplates returns valid filesystem
    Given the template package is imported
    When EmbeddedTemplates() is called
    Then it shall return a non-nil fs.FS
    And it shall return a nil error

  Scenario: Paths are stripped of templates/ prefix
    Given the embedded FS is loaded via EmbeddedTemplates()
    When accessing ".claude/agents/moai/expert-backend.md"
    Then the file shall be readable
    And the content shall be non-empty
    And the path shall NOT require a "templates/" prefix

  Scenario: fs.WalkDir enumerates all files
    Given the embedded FS is loaded via EmbeddedTemplates()
    When fs.WalkDir is called on the root
    Then the total file count shall be at least 450
    And no walk errors shall occur

  Scenario: Build fails if templates directory is missing
    Given the templates/ directory does not exist
    When "go build ./..." is executed
    Then the build shall fail with an embed-related error
```

---

### ACC-004: CLI Init Deploys Templates

**Requirement Reference**: REQ-E-030 through REQ-E-034

```gherkin
Feature: moai init Template Deployment

  Scenario: Full template deployment to clean directory
    Given a clean temporary directory exists
    And no .moai/ or .claude/ directories exist
    When "moai init --non-interactive --name test-project --conv-lang en" is executed
    Then the following directories shall exist:
      | Directory                        |
      | .claude/agents/moai/             |
      | .claude/skills/                  |
      | .claude/rules/moai/              |
      | .claude/output-styles/moai/      |
      | .moai/config/sections/           |
      | .moai/specs/                     |
      | .moai/reports/                   |
      | .moai/memory/                    |
      | .moai/logs/                      |
    And the following files shall exist:
      | File                              |
      | CLAUDE.md                         |
      | .claude/settings.json             |
      | .mcp.json                         |
      | .gitignore                        |
      | .moai/manifest.json               |
      | .moai/config/sections/user.yaml   |
      | .moai/config/sections/language.yaml |
      | .moai/config/sections/quality.yaml |
      | .moai/config/sections/workflow.yaml |
    And at least 15 agent files shall exist under .claude/agents/moai/
    And at least 400 skill files shall exist under .claude/skills/
    And CLAUDE.md shall contain "MoAI Execution Directive"
    And .claude/settings.json shall be valid JSON
    And .claude/settings.json shall contain "hooks" key

  Scenario: settings.json is runtime-generated, not from templates
    Given a clean temporary directory exists
    When "moai init --non-interactive" is executed
    Then .claude/settings.json shall exist
    And .claude/settings.json shall contain "moai hook session-start"
    And the settings.json shall be generated by SettingsGenerator (not from embedded templates)

  Scenario: CLAUDE.md is the full template, not the stub
    Given a clean temporary directory exists
    When "moai init --non-interactive --name test-project" is executed
    Then CLAUDE.md shall be at least 5000 characters long
    And CLAUDE.md shall NOT be the minimal stub (more than 20 lines)
```

---

### ACC-005: Existing File Protection

**Requirement Reference**: REQ-E-032

```gherkin
Feature: Existing File Protection During Init

  Scenario: Pre-existing file is not overwritten
    Given a temporary directory with a pre-existing ".gitignore" file
    And the pre-existing .gitignore contains "my-custom-pattern"
    When "moai init --non-interactive" is executed
    Then the .gitignore file shall still contain "my-custom-pattern"
    And the .gitignore file shall NOT be overwritten with the template version

  Scenario: Pre-existing CLAUDE.md is preserved
    Given a temporary directory with a pre-existing "CLAUDE.md" file
    And the pre-existing CLAUDE.md contains "My Custom Instructions"
    When "moai init --non-interactive" is executed
    Then the CLAUDE.md shall still contain "My Custom Instructions"
    And the manifest shall track CLAUDE.md as "user_created"

  Scenario: Pre-existing file tracked in manifest
    Given a temporary directory with a pre-existing ".mcp.json" file
    When "moai init --non-interactive" is executed
    Then the manifest shall contain an entry for ".mcp.json"
    And the entry's provenance shall be "user_created"
```

---

### ACC-006: Manifest Tracking

**Requirement Reference**: REQ-E-040, REQ-E-041

```gherkin
Feature: Manifest Tracking After Deployment

  Scenario: All deployed files are tracked in manifest
    Given a clean temporary directory
    When "moai init --non-interactive" is executed
    And the manifest is loaded from .moai/manifest.json
    Then every deployed template file shall have a manifest entry
    And each entry's provenance shall be "template_managed"
    And each entry's template_hash shall start with "sha256:"
    And each entry's deployed_hash shall equal template_hash (freshly deployed)
    And each entry's current_hash shall equal deployed_hash

  Scenario: Manifest records ADK version
    Given a clean temporary directory
    When "moai init --non-interactive" is executed
    And the manifest is loaded from .moai/manifest.json
    Then the manifest "version" field shall be non-empty
    And the manifest "version" field shall match the binary's version

  Scenario: Manifest records deployment timestamp
    Given a clean temporary directory
    And the current UTC time is captured before init
    When "moai init --non-interactive" is executed
    And the manifest is loaded from .moai/manifest.json
    Then the manifest "deployed_at" field shall be a valid ISO 8601 timestamp
    And the timestamp shall be within 60 seconds of the captured time

  Scenario: Manifest JSON is valid
    Given a clean temporary directory
    When "moai init --non-interactive" is executed
    Then .moai/manifest.json shall be valid JSON
    And the JSON shall have "version", "deployed_at", and "files" keys
```

---

### ACC-007: Template Update Flow

**Requirement Reference**: REQ-E-050 through REQ-E-053

```gherkin
Feature: Template Update via moai update

  Scenario: Template-managed files are updated
    Given a MoAI project initialized with version "1.0.0"
    And the agent file ".claude/agents/moai/expert-backend.md" has provenance "template_managed"
    And the file has not been modified by the user
    When "moai update" replaces with a new binary containing updated templates
    Then the agent file shall be overwritten with the new template version
    And the manifest shall reflect the new template_hash

  Scenario: User-modified files trigger 3-way merge
    Given a MoAI project initialized with version "1.0.0"
    And the user has modified ".claude/rules/moai/core/moai-constitution.md"
    And the manifest provenance is "user_modified"
    When "moai update" is executed with updated templates
    Then the merge engine shall be invoked with base, current, and updated versions
    And if no conflicts, the file shall contain both user changes and template updates
    And if conflicts, a ".conflict" file shall be created

  Scenario: User-created files are skipped during update
    Given a MoAI project with a user-created file ".claude/skills/my-custom-skill/SKILL.md"
    And the manifest provenance is "user_created"
    When "moai update" is executed
    Then the user-created file shall not be modified
    And the manifest entry shall remain unchanged
```

---

### ACC-008: Binary Size Constraint

**Requirement Reference**: REQ-E-060, REQ-E-061

```gherkin
Feature: Binary Size Budget

  Scenario: Compiled binary is within size budget
    Given the full template content is embedded
    When "go build -o moai ./cmd/moai/" is executed
    Then the binary file size shall be less than 30 MB

  Scenario: Template content size is within limit
    Given the templates/ directory is populated
    When the total size of all files in templates/ is calculated
    Then the total shall be less than 10 MB
```

---

### ACC-009: Security Validation

**Requirement Reference**: REQ-E-012 (cache exclusion), general security

```gherkin
Feature: Template Security

  Scenario: No credentials in template files
    Given the templates/ directory is populated
    When all template files are scanned for credential patterns
    Then no file shall contain patterns matching:
      | Pattern                    |
      | API_KEY=                   |
      | SECRET_KEY=                |
      | password=                  |
      | PRIVATE_KEY                |
      | Bearer [A-Za-z0-9]{20,}   |
      | sk-[A-Za-z0-9]{20,}       |

  Scenario: No unexpanded dynamic tokens in non-template files
    Given the templates/ directory is populated
    When all .md, .yaml, and .json files are scanned
    Then no file shall contain unexpanded tokens matching:
      | Pattern          |
      | ${VARIABLE}      |
      | $VARIABLE_NAME   |
    And .json files shall be valid JSON
    And .yaml files shall be valid YAML

  Scenario: Path traversal protection during deployment
    Given a clean temporary directory
    When "moai init --non-interactive" is executed
    Then no files shall be written outside the project root directory
    And all deployed file paths shall be normalized via filepath.Clean()
```

---

### ACC-010: Performance

**Requirement Reference**: Performance requirements in spec.md section 5.6

```gherkin
Feature: Deployment Performance

  Scenario: Full template deployment completes within time budget
    Given a clean temporary directory
    When Deployer.Deploy() is called with the full embedded FS
    Then the deployment shall complete in less than 2 seconds
    And memory usage during deployment shall not exceed 100 MB

  Scenario: Embedded FS file listing is fast
    Given the embedded FS is loaded via EmbeddedTemplates()
    When ListTemplates() is called
    Then the listing shall complete in less than 10 milliseconds
    And the result shall contain at least 450 entries
```

---

## 2. Definition of Done

### 2.1 Code Completeness

- [ ] `templates/` directory contains all curated template content (~480 files)
- [ ] `internal/template/embed.go` declares `//go:embed` and exports `EmbeddedTemplates()`
- [ ] `internal/cli/init.go` constructs real `Deployer` from embedded FS
- [ ] `internal/core/project/initializer.go` deploys CLAUDE.md from templates (not stub)
- [ ] `internal/template/deployer.go` includes existing file protection logic
- [ ] Manifest records version and deployment timestamp

### 2.2 Test Coverage

- [ ] Unit tests for `embed.go` (FS validity, file existence, exclusion checks)
- [ ] Integration tests for full init deployment (ACC-004)
- [ ] Integration tests for existing file protection (ACC-005)
- [ ] Integration tests for manifest tracking (ACC-006)
- [ ] Benchmark tests for deployment performance (ACC-010)
- [ ] Overall test coverage >= 85% for modified/new files

### 2.3 Quality Gates

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] Binary size < 30 MB
- [ ] No credentials or secrets in template content
- [ ] No unexpanded dynamic tokens in template files
- [ ] All JSON template files are valid JSON
- [ ] All YAML template files are valid YAML

### 2.4 Documentation

- [ ] `embed.go` has package-level documentation explaining the embed strategy
- [ ] README updated with template architecture overview (if applicable)
- [ ] CHANGELOG entry for template bundling feature

---

## 3. Verification Methods

| Criteria | Method | Tool |
|----------|--------|------|
| Template completeness | Unit test walking embedded FS | `go test` |
| File exclusion | Unit test checking absent paths | `go test` |
| Init deployment | Integration test in temp dir | `go test -run Integration` |
| Manifest tracking | Integration test loading manifest | `go test -run Integration` |
| Existing file protection | Integration test with pre-existing files | `go test -run Integration` |
| Binary size | CI script checking file size | `ls -la` + threshold check |
| Security scan | Grep for credential patterns | CI script + `grep` |
| Performance | Benchmark test | `go test -bench` |
| JSON/YAML validity | Validation in integration tests | `json.Valid()` + `yaml.Unmarshal()` |
