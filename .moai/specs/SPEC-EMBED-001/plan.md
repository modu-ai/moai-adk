# SPEC-EMBED-001: Implementation Plan

---
spec_id: SPEC-EMBED-001
title: Template Content and Binary Bundling - Implementation Plan
status: Planned
created: 2026-02-03
tags: go-embed, templates, bundling, binary, deployment, init
---

## 1. Implementation Strategy

### 1.1 Development Methodology

**Hybrid approach** (TDD for new code, DDD for existing code):

- **TDD**: New `embed.go` file, integration tests, CI validation scripts
- **DDD**: Modifications to `init.go` (existing), `deployer.go` (existing), `initializer.go` (existing)

### 1.2 Guiding Principles

- Minimal modifications to existing, tested code (SPEC-TEMPLATE-001 engine is stable)
- Content curation is the primary effort -- the engine already works
- Prefer `fs.Sub()` path stripping over custom path manipulation
- Static templates only -- no dynamic token injection in template files (ADR-011)

---

## 2. Milestones

### Milestone 1: Template Content Curation (Primary Goal)

**Objective**: Populate `templates/` directory with Go-edition-appropriate content.

**Tasks**:

- [ ] **M1-T01**: Audit Python reference templates (`moai_adk/templates/`) for Go edition applicability
  - Classify each file category as Include/Exclude/Modify
  - Document rationale for exclusions (AD-001 through AD-004)
  - Verify no credentials, API keys, or secrets in template files

- [ ] **M1-T02**: Create agent definition files in `templates/.claude/agents/moai/`
  - Copy all 20 agent `.md` files from Python reference
  - Verify each file is valid Markdown with correct YAML frontmatter
  - Remove any Python-specific references (e.g., Python hook paths) and update for Go edition

- [ ] **M1-T03**: Create skill definition files in `templates/.claude/skills/`
  - Copy all ~423 skill files preserving directory hierarchy
  - Verify SKILL.md frontmatter compliance (name, description, metadata)
  - Check for broken cross-references between skill modules

- [ ] **M1-T04**: Create rule files in `templates/.claude/rules/moai/`
  - Copy all 22 rule files preserving subdirectory structure (core/, development/, workflow/)
  - Verify Markdown structure and rule content

- [ ] **M1-T05**: Create output style files in `templates/.claude/output-styles/moai/`
  - Copy all 3 output style files (alfred.md, r2d2.md, yoda.md)

- [ ] **M1-T06**: Create configuration templates in `templates/.moai/`
  - Copy config section templates (user.yaml, language.yaml, quality.yaml, workflow.yaml, system.yaml)
  - Copy announcement JSON files (en.json, ko.json, ja.json, zh.json)
  - Copy LLM config (glm.json)
  - Copy statusline config and multilingual triggers if applicable

- [ ] **M1-T07**: Create root-level template files
  - Copy CLAUDE.md (full ~9,600-character instruction file)
  - Copy .mcp.json and .mcp.windows.json
  - Copy .gitignore

- [ ] **M1-T08**: Remove `.gitkeep` placeholder files
  - Delete all `.gitkeep` files from `templates/` directory
  - Verify directory structure is correct after removal

**Verification**: All template files present, no excluded file categories (hooks, settings.json, cache files, etc.), `find templates/ -type f | wc -l` matches expected count.

**Dependencies**: None (this is the foundation task)

---

### Milestone 2: go:embed Integration (Primary Goal)

**Objective**: Wire template files into the Go binary via `//go:embed`.

**Tasks**:

- [ ] **M2-T01**: Determine embed.go placement strategy
  - Evaluate Option A (cmd/moai/), Option B (pkg/embedded/), Option C (internal/template/templates/)
  - Select based on Go embed path resolution rules
  - If Option C: move `templates/` to `internal/template/templates/`
  - If Option A/B: create package with embed directive

- [ ] **M2-T02**: Create `embed.go` with `//go:embed` directive
  - Declare `//go:embed all:templates` (the `all:` prefix includes dot-prefixed dirs)
  - Export `EmbeddedTemplates()` function returning `fs.FS` after `fs.Sub()` stripping
  - Add package documentation explaining the embed strategy

- [ ] **M2-T03**: Write unit tests for embed.go
  - Test that `EmbeddedTemplates()` returns a valid `fs.FS`
  - Test that key files exist in the embedded FS (agents, skills, CLAUDE.md)
  - Test that excluded files do NOT exist (settings.json, .py hooks)
  - Test path resolution: `.claude/agents/moai/expert-backend.md` is accessible
  - Test that `fs.WalkDir` returns expected file count

- [ ] **M2-T04**: Verify build succeeds with embedded templates
  - `go build ./cmd/moai/` completes without errors
  - Binary size is within 30 MB target
  - `go vet ./...` passes
  - `go test ./...` passes

**Verification**: `go build` succeeds, binary contains templates, unit tests pass, binary size < 30 MB.

**Dependencies**: M1 (template content must exist before embedding)

---

### Milestone 3: CLI Integration (Primary Goal)

**Objective**: Wire embedded deployer into `moai init` to enable full template deployment.

**Tasks**:

- [ ] **M3-T01**: Modify `internal/cli/init.go` to use real deployer
  - Import `template` package
  - Call `template.EmbeddedTemplates()` to get embedded FS
  - Create `template.NewDeployer(embeddedFS)` with the real FS
  - Pass deployer to `project.NewInitializer(deployer, mgr, nil)`
  - Handle `EmbeddedTemplates()` error gracefully

- [ ] **M3-T02**: Update `internal/core/project/initializer.go` for CLAUDE.md
  - Replace `buildClaudeMDContent()` stub with deployment from embedded templates
  - The full CLAUDE.md from templates should be deployed by the Deployer
  - Remove or deprecate the `buildClaudeMDContent()` function
  - Ensure CLAUDE.md deployment is tracked in manifest

- [ ] **M3-T03**: Add existing file protection to Deployer
  - Before writing each file, check if destination exists
  - If file exists and is in manifest as `user_modified`/`user_created`: skip
  - If file exists but NOT in manifest: track as `user_created` and skip
  - If file does not exist: write and track as `template_managed`
  - Log skipped files as warnings

- [ ] **M3-T04**: Wire settings.json generation after template deployment
  - After `deployer.Deploy()`, invoke `SettingsGenerator.Generate()` for `.claude/settings.json`
  - Track settings.json in manifest with `template_managed` provenance
  - Ensure settings.json is NOT deployed from templates (REQ-E-011)

- [ ] **M3-T05**: Add version metadata to manifest
  - Set `manifest.Version` to `version.GetVersion()` from `pkg/version/`
  - Set `manifest.DeployedAt` to current UTC time (ISO 8601)
  - Ensure manifest is saved after all tracking operations

**Verification**: `moai init` deploys all template files, manifest is populated, existing files are preserved, settings.json is runtime-generated.

**Dependencies**: M2 (embed must be wired before CLI can use it)

---

### Milestone 4: Integration Tests (Secondary Goal)

**Objective**: End-to-end validation of the template deployment pipeline.

**Tasks**:

- [ ] **M4-T01**: Write integration test: `moai init` deploys all expected files
  - Create temp directory
  - Run full init flow with embedded templates
  - Verify all expected template files exist at correct paths
  - Verify file contents are non-empty and match embedded source

- [ ] **M4-T02**: Write integration test: manifest tracking
  - After init, load manifest from `.moai/manifest.json`
  - Verify all deployed files have entries with `provenance: template_managed`
  - Verify `template_hash` and `deployed_hash` are SHA-256 values
  - Verify `Version` and `DeployedAt` fields are populated

- [ ] **M4-T03**: Write integration test: existing file protection
  - Create a file at a template destination path before init
  - Run init
  - Verify the pre-existing file is NOT overwritten
  - Verify the file is tracked as `user_created` in manifest

- [ ] **M4-T04**: Write integration test: re-init with `--force`
  - Run init once (creates all files)
  - Modify a deployed file (triggers `user_modified` detection)
  - Run init again with `--force`
  - Verify `template_managed` files are refreshed
  - Verify `user_modified` files are preserved or merged

- [ ] **M4-T05**: Write integration test: template rendering (if any templates use Go template syntax)
  - Test templates that use `text/template` rendering
  - Verify rendered output has no unexpanded tokens
  - Verify strict mode catches missing keys

- [ ] **M4-T06**: Write benchmark test: deployment performance
  - Benchmark `Deployer.Deploy()` with full embedded FS
  - Target: < 2 seconds for ~480 files
  - Measure memory allocation with `testing.B.ReportAllocs()`

**Verification**: All integration tests pass, benchmark meets performance targets.

**Dependencies**: M3 (CLI integration must be complete)

---

### Milestone 5: Template Update Flow (Secondary Goal)

**Objective**: Ensure `moai update` correctly refreshes templates using the new embedded content.

**Tasks**:

- [ ] **M5-T01**: Verify update orchestrator uses embedded templates
  - Confirm `internal/update/orchestrator.go` calls `template.EmbeddedTemplates()` for new template extraction
  - Verify manifest-based provenance routing works with real template files

- [ ] **M5-T02**: Write integration test: template update flow
  - Simulate version bump with modified templates
  - Verify `template_managed` files are updated
  - Verify `user_modified` files trigger 3-way merge
  - Verify `user_created` files are skipped

- [ ] **M5-T03**: Verify version tracking across updates
  - After update, manifest `Version` reflects new version
  - `DeployedAt` is updated to current time
  - `template_hash` entries are updated for changed templates

**Verification**: Update flow works end-to-end with real templates.

**Dependencies**: M4 (integration tests establish baseline behavior)

---

### Milestone 6: CI Validation (Optional Goal)

**Objective**: Automated CI checks for template integrity and binary size.

**Tasks**:

- [ ] **M6-T01**: Add CI check for template content completeness
  - Script to compare embedded template list against expected manifest
  - Fail if expected files are missing
  - Fail if excluded files (hooks, settings.json) are present

- [ ] **M6-T02**: Add CI check for binary size
  - Build binary and check size < 30 MB
  - Alert if size exceeds 25 MB (warning threshold)

- [ ] **M6-T03**: Add CI check for template security
  - Grep templates for potential credentials (API keys, tokens, passwords)
  - Fail build if suspicious patterns are found

- [ ] **M6-T04**: Add CI check for no dynamic tokens
  - Scan all embedded template files for `${VAR}`, `{{.VAR}}`, `$VAR` patterns
  - Allow legitimate Go template syntax in `.tmpl` files only
  - Fail build if unexpanded tokens found in non-template files

**Verification**: CI pipeline runs all checks on every PR.

**Dependencies**: M2 (templates must be embedded for CI to validate)

---

## 3. Technical Approach

### 3.1 Package Structure Decision

The `//go:embed` directive can only reference files in the same directory or subdirectories. Given the current repo layout:

```
moai-adk-go/
  cmd/moai/main.go
  internal/template/deployer.go   # Deployer lives here
  templates/                      # Current template location
```

**Recommended approach**: Create `internal/template/templates/` as the embed source and symlink or move content there. This keeps the embed directive in the same package as the Deployer.

**Alternative**: If the team prefers keeping `templates/` at the repo root, create a thin `pkg/embedded/` package with the embed directive and import it from `internal/template/`.

### 3.2 Dependency Graph

```
internal/cli/init.go
    |
    +---> internal/template/embed.go        (NEW: EmbeddedTemplates())
    |         |
    |         +---> internal/template/templates/  (NEW: actual template files)
    |
    +---> internal/template/deployer.go     (EXISTING: fs.FS consumer)
    |
    +---> internal/template/settings.go     (EXISTING: runtime settings.json)
    |
    +---> internal/manifest/manifest.go     (EXISTING: file tracking)
    |
    +---> internal/core/project/initializer.go  (MODIFY: receive real deployer)
    |
    +---> pkg/version/version.go            (EXISTING: version for manifest)
```

### 3.3 Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Template content bloats binary beyond 30 MB | Low | Medium | Monitor size in CI; compress large skills if needed |
| `embed.FS` path resolution differs across platforms | Very Low | High | Use `fs.Sub()` and forward-slash paths; test on CI matrix |
| Python-specific content in templates breaks Go edition | Medium | Low | Audit all template files during M1 curation |
| Existing user projects broken by new init behavior | Low | High | Existing file protection (M3-T03); comprehensive integration tests |
| Build time increase due to large embed | Low | Low | Go compiler handles embed efficiently; monitor CI build times |

### 3.4 Files Modified (Summary)

| File | Action | Milestone |
|------|--------|-----------|
| `internal/template/embed.go` | **Create** | M2 |
| `internal/template/templates/` | **Create** (directory with content) | M1 |
| `templates/` (repo root) | **Remove or redirect** | M1 |
| `internal/cli/init.go` | **Modify** (wire real deployer) | M3 |
| `internal/core/project/initializer.go` | **Modify** (CLAUDE.md from templates) | M3 |
| `internal/template/deployer.go` | **Modify** (existing file protection) | M3 |
| `internal/template/embed_test.go` | **Create** | M2 |
| `internal/cli/init_integration_test.go` | **Create** | M4 |

---

## 4. Content Curation Checklist

### 4.1 Files to Copy from Python Reference

| Source Path | Destination Path | Notes |
|-------------|-----------------|-------|
| `moai_adk/templates/.claude/agents/moai/*.md` | `templates/.claude/agents/moai/` | 20 files, copy as-is |
| `moai_adk/templates/.claude/skills/**/*.md` | `templates/.claude/skills/` | ~423 files, preserve hierarchy |
| `moai_adk/templates/.claude/rules/moai/**/*.md` | `templates/.claude/rules/moai/` | 22 files, preserve hierarchy |
| `moai_adk/templates/.claude/output-styles/moai/*.md` | `templates/.claude/output-styles/moai/` | 3 files |
| `moai_adk/templates/.moai/config/sections/*.yaml` | `templates/.moai/config/sections/` | Select applicable configs |
| `moai_adk/templates/.moai/announcements/*.json` | `templates/.moai/announcements/` | 4 language files |
| `moai_adk/templates/.moai/llm-configs/*.json` | `templates/.moai/llm-configs/` | 1 file |
| `moai_adk/templates/CLAUDE.md` | `templates/CLAUDE.md` | Full instruction file |
| `moai_adk/templates/.mcp.json` | `templates/.mcp.json` | MCP config |
| `moai_adk/templates/.mcp.windows.json` | `templates/.mcp.windows.json` | Windows MCP variant |
| `moai_adk/templates/.gitignore` | `templates/.gitignore` | Git ignore rules |

### 4.2 Files to Exclude

| Source Path | Reason |
|-------------|--------|
| `moai_adk/templates/.claude/hooks/**` | Go uses compiled hooks (AD-001) |
| `moai_adk/templates/.claude/settings.json` | Runtime-generated (ADR-011) |
| `moai_adk/templates/.claude/.DS_Store` | OS metadata |
| `moai_adk/templates/.git-hooks/**` | Go manages Git hooks differently |
| `moai_adk/templates/.lsp.json` | Runtime-generated by LSP module |
| Any `__pycache__/` directories | Python bytecode cache |
| Any `.pyc` files | Python bytecode |

### 4.3 Files Requiring Go-Edition Modification

| File | Required Changes |
|------|-----------------|
| `CLAUDE.md` | Verify version reference matches Go binary; remove Python-specific instructions if any |
| `.mcp.json` | Verify MCP server paths are correct for Go binary (`moai` not `python -m moai_adk`) |
| `.mcp.windows.json` | Same as .mcp.json for Windows paths |
| `.gitignore` | Ensure Go-specific patterns are included (binary, vendor/) |
| Config YAML templates | Verify defaults align with Go config module expectations |
| Agent definitions | Remove references to Python-specific features or hooks if present |

---

## 5. Quality Gates

### 5.1 Pre-Merge Checklist

- [ ] All template files present in `templates/` directory
- [ ] No excluded file categories in `templates/`
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (including new integration tests)
- [ ] `go vet ./...` passes
- [ ] Binary size < 30 MB
- [ ] `moai init` deploys all expected files in clean directory
- [ ] `moai init` preserves existing files
- [ ] Manifest is correctly populated after init
- [ ] No credentials or secrets in template content
- [ ] No unexpanded dynamic tokens in template files

### 5.2 TRUST 5 Validation

| Pillar | Criteria | Status |
|--------|----------|--------|
| **Tested** | Integration tests for deployment, manifest, protection | Required |
| **Readable** | embed.go well-documented, clear package structure | Required |
| **Unified** | Consistent with existing template/ package conventions | Required |
| **Secured** | No secrets in templates, path traversal prevention | Required |
| **Trackable** | All files tracked in manifest with provenance | Required |
