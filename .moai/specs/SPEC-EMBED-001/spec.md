# SPEC-EMBED-001: Template Content and Binary Bundling

---
spec_id: SPEC-EMBED-001
title: Template Content and Binary Bundling
status: Implemented
priority: High
phase: "Phase 2 - Core Domains (P0 Critical)"
created: 2026-02-03
depends_on:
  - SPEC-TEMPLATE-001
  - SPEC-INIT-001
  - SPEC-UPDATE-001
modules:
  - templates/
  - internal/template/embed.go
  - internal/cli/init.go
estimated_loc: ~800
adr_references:
  - ADR-003 (go:embed for Template Distribution)
  - ADR-007 (File Manifest Provenance)
  - ADR-011 (Zero Runtime Template Expansion)
lifecycle: spec-anchored
tags: go-embed, templates, bundling, binary, deployment, init
---

## HISTORY

| Date | Version | Changes |
|------|---------|---------|
| 2026-02-03 | 1.0.0 | Initial SPEC creation |

---

## 1. Overview

### 1.1 Background

SPEC-TEMPLATE-001 delivered the template ENGINE -- the `Deployer`, `Renderer`, `SettingsGenerator`, and `Validator` interfaces along with the `manifest.Manager` for file provenance tracking. All engine tests pass using `testing/fstest.MapFS` as the template source.

However, the actual template CONTENT is missing. The `templates/` directory contains only `.gitkeep` placeholder files:

```
templates/
  .claude/hooks/.gitkeep
  .claude/skills/.gitkeep
  .claude/rules/.gitkeep
  .claude/agents/.gitkeep
  .claude/output-styles/.gitkeep
```

Additionally:
- No `//go:embed` directive exists anywhere in the codebase
- `internal/cli/init.go` passes `nil` for the deployer: `project.NewInitializer(nil, mgr, nil)`
- `moai init` creates directory structures but deploys zero template content
- The generated `CLAUDE.md` is a minimal stub (12 lines) instead of the full ~9,600-character instruction file

The Python reference implementation (`moai_adk/templates/`) contains ~564 files totaling ~7.3 MB, including 20 agent definitions, 423 skill files, 22 rule files, 3 output styles, 39 Python hook scripts, 13 config templates, and root files (CLAUDE.md, .mcp.json, .gitignore).

### 1.2 Objectives

- Populate the `templates/` directory with Go-edition-appropriate content curated from the Python reference
- Create the `//go:embed` directive to bundle templates into the `moai` binary
- Wire the embedded `fs.FS` through `NewDeployer()` to the CLI `init` command
- Enable `moai init` to deploy a complete, functional MoAI project scaffold
- Ensure `moai update` template refresh works via the existing manifest + merge system

### 1.3 Scope

**Included:**
- `templates/` directory: curated template content for Go edition
- `internal/template/embed.go`: `//go:embed` directive and exported `embed.FS`
- `internal/cli/init.go`: wire real deployer (replace `nil`)
- Integration tests for end-to-end template deployment
- Version embedding in manifest metadata

**Excluded:**
- Template engine modifications (SPEC-TEMPLATE-001 -- completed, stable)
- 3-way merge engine modifications (SPEC-UPDATE-001 -- completed, stable)
- New CLI subcommands (no new commands required)
- `internal/config/` changes (SPEC-CONFIG-001 -- completed)

---

## 2. Environment

### 2.1 System Environment

| Item | Specification |
|------|---------------|
| Language | Go 1.22+ |
| Module path | `github.com/modu-ai/moai-adk-go` |
| Target platforms | darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64, windows/arm64 |
| CGO | CGO_ENABLED=0 (pure Go) |
| Binary size target | < 30 MB (including embedded templates) |
| Embedded content estimate | ~2.5 MB (after Python-hook and cache exclusion) |

### 2.2 Dependencies

| Module | Purpose | Interface |
|--------|---------|-----------|
| `internal/template/` | Deployer, Renderer, Validator | `template.Deployer`, `template.Renderer` |
| `internal/manifest/` | File provenance tracking | `manifest.Manager` |
| `internal/config/` | Config loading for SettingsGenerator | `config.Config` |
| `internal/merge/` | 3-way merge for template updates | `merge.Engine` |
| `pkg/version/` | ADK version for manifest | `version.GetVersion()` |

### 2.3 Python Reference Template Inventory

| Category | Path Pattern | File Count | Size | Go Edition Action |
|----------|-------------|------------|------|-------------------|
| Agents | `.claude/agents/moai/*.md` | 20 | 592 KB | **Include all** |
| Skills | `.claude/skills/**/*.md` | 423 | 5.6 MB | **Include all** |
| Rules | `.claude/rules/moai/**/*.md` | 22 | 92 KB | **Include all** |
| Output styles | `.claude/output-styles/moai/*.md` | 3 | 72 KB | **Include all** |
| Python hooks | `.claude/hooks/moai/**/*.py` | 39 | 748 KB | **Exclude** (see AD-001) |
| Config templates | `.moai/config/sections/*.yaml` | 13 | ~20 KB | **Include subset** |
| CLAUDE.md | `CLAUDE.md` | 1 | ~10 KB | **Include** |
| .mcp.json | `.mcp.json` | 1 | ~1 KB | **Include** |
| .gitignore | `.gitignore` | 1 | ~3 KB | **Include** |
| .lsp.json | `.lsp.json` | 1 | ~3 KB | **Exclude** (runtime-generated) |
| .git-hooks | `.git-hooks/*` | 2 | ~2 KB | **Exclude** (Go uses compiled hooks) |
| Announcements | `.moai/announcements/*.json` | 4 | ~4 KB | **Include** |
| GLM config | `.moai/llm-configs/*.json` | 1 | ~2 KB | **Include** |
| MCP Windows | `.mcp.windows.json` | 1 | ~1 KB | **Include** |

### 2.4 Embedded Template Directory Structure

```
templates/                          # go:embed source root
  .claude/
    agents/moai/                    # Agent definitions (~20 .md files)
      expert-backend.md
      expert-frontend.md
      expert-security.md
      ...
    skills/                         # Skill definitions (~423 .md files, deep hierarchy)
      moai-foundation-core/
        SKILL.md
        modules/
        ...
      moai-lang-python/
        SKILL.md
        modules/
        ...
    rules/moai/                     # Rule files (~22 .md files)
      core/
      development/
      workflow/
      ...
    output-styles/moai/             # Output styles (~3 .md files)
      alfred.md
      r2d2.md
      yoda.md
  .moai/
    config/sections/                # Config templates (~5 .yaml files)
      user.yaml
      language.yaml
      quality.yaml
      workflow.yaml
      system.yaml
    announcements/                  # Announcement templates
      en.json
      ko.json
      ja.json
      zh.json
    llm-configs/                    # LLM configuration
      glm.json
  CLAUDE.md                         # Main instruction file template
  .mcp.json                         # MCP server configuration
  .mcp.windows.json                 # MCP Windows variant
  .gitignore                        # Git ignore rules
```

---

## 3. Assumptions

### 3.1 Technical Assumptions

| ID | Assumption | Confidence | Evidence | Risk if Wrong |
|----|-----------|------------|----------|---------------|
| A-01 | SPEC-TEMPLATE-001 engine is stable and passes all tests | High | Status: Completed, 85.7% coverage | Engine modifications needed |
| A-02 | SPEC-INIT-001 initialization flow is complete and wired | High | Status: Completed, 89.2% coverage | Init flow redesign |
| A-03 | `go:embed` supports recursive directory embedding with `all:` prefix for dot-prefixed dirs | High | Go 1.16+ specification | Cannot embed dotfiles |
| A-04 | Python reference templates are the canonical content source | High | moai_adk/templates/ exists with 564 files | Need alternative content source |
| A-05 | Template files do not exceed Go embed file size limits | High | Total ~2.5 MB after exclusion | Need chunked embedding |
| A-06 | `embed.FS` paths use forward slashes on all platforms | High | Go embed specification | Path resolution failures |

### 3.2 Business Assumptions

| ID | Assumption | Confidence | Evidence |
|----|-----------|------------|----------|
| B-01 | Go edition reuses the same agent/skill/rule content as Python edition | High | Same CLAUDE.md, same agent definitions |
| B-02 | Python hook scripts should NOT be embedded since Go uses compiled hook handlers | High | settings.go generates `moai hook <event>` commands |
| B-03 | CLAUDE.md template should be the full instruction file, not the current stub | High | Python reference CLAUDE.md is ~9,600 characters |
| B-04 | Config YAML templates serve as defaults; runtime-generated configs override them | Medium | init.go already generates configs from structs |

### 3.3 Architectural Decisions within this SPEC

**AD-001: Exclude Python Hook Scripts**

The Go edition's `settings.go` generates settings.json with `moai hook <event>` commands that route to compiled Go subcommands in `internal/hook/`. The Python hook scripts (`.claude/hooks/moai/**/*.py`) from the reference templates are NOT needed and SHALL NOT be embedded. This eliminates 39 files / 748 KB from the binary.

**AD-002: Static CLAUDE.md (No Template Rendering)**

Per ADR-011 (Zero Runtime Template Expansion), the embedded `CLAUDE.md` SHALL be a static file that is deployed as-is by the `Deployer`. Project-specific values (project name, language, framework) are NOT injected into CLAUDE.md at deploy time; they exist in `.moai/config/sections/*.yaml` which Claude Code reads separately. This avoids the template variable expansion failures seen in the Python edition.

**AD-003: Path Stripping via embed.FS Sub-Directory**

The `//go:embed` directive embeds with the `templates/` prefix. The `embed.FS` must be sub-sliced using `fs.Sub(embeddedFS, "templates")` so that embedded paths like `templates/.claude/agents/moai/expert-backend.md` deploy as `.claude/agents/moai/expert-backend.md`.

**AD-004: Config Templates as Reference Defaults**

Config YAML files embedded under `templates/.moai/config/sections/` serve as reference defaults only. The initializer in `init.go` already generates configs programmatically via struct serialization (ADR-011). The embedded config templates provide fallback/reference values and are deployed only if the programmatic generation does not cover them.

---

## 4. Requirements (EARS Format)

### 4.1 Template Content Curation

**[REQ-E-001] Agent Definitions Bundling (Ubiquitous)**

The system shall always include all agent definition files from `moai_adk/templates/.claude/agents/moai/*.md` in the `templates/.claude/agents/moai/` embed directory.

**[REQ-E-002] Skill Definitions Bundling (Ubiquitous)**

The system shall always include all skill definition files from `moai_adk/templates/.claude/skills/` (recursive, preserving directory structure) in the `templates/.claude/skills/` embed directory.

**[REQ-E-003] Rule Files Bundling (Ubiquitous)**

The system shall always include all rule files from `moai_adk/templates/.claude/rules/moai/` (recursive) in the `templates/.claude/rules/moai/` embed directory.

**[REQ-E-004] Output Style Bundling (Ubiquitous)**

The system shall always include all output style files from `moai_adk/templates/.claude/output-styles/moai/` in the `templates/.claude/output-styles/moai/` embed directory.

**[REQ-E-005] CLAUDE.md Bundling (Ubiquitous)**

The system shall always include the full CLAUDE.md instruction file from `moai_adk/templates/CLAUDE.md` in the `templates/` embed root.

**[REQ-E-006] MCP Configuration Bundling (Ubiquitous)**

The system shall always include `.mcp.json` and `.mcp.windows.json` from `moai_adk/templates/` in the `templates/` embed root.

**[REQ-E-007] Git Ignore Bundling (Ubiquitous)**

The system shall always include `.gitignore` from `moai_adk/templates/` in the `templates/` embed root.

**[REQ-E-008] Announcement Templates Bundling (Ubiquitous)**

The system shall always include announcement JSON files from `moai_adk/templates/.moai/announcements/` in the `templates/.moai/announcements/` embed directory.

**[REQ-E-009] LLM Config Bundling (Ubiquitous)**

The system shall always include LLM configuration from `moai_adk/templates/.moai/llm-configs/` in the `templates/.moai/llm-configs/` embed directory.

**[REQ-E-010] Python Hook Exclusion (Unwanted)**

The system shall NOT include `.claude/hooks/` Python scripts in the embedded template directory, as the Go edition handles hooks via compiled subcommands (`moai hook <event>`).

**[REQ-E-011] Settings JSON Exclusion (Unwanted)**

The system shall NOT include `.claude/settings.json` in the embedded template directory, as it is generated at runtime by `template.SettingsGenerator` (ADR-011).

**[REQ-E-012] Cache and OS File Exclusion (Unwanted)**

The system shall NOT include `__pycache__/`, `.DS_Store`, `.pyc` files, or any OS-specific metadata files in the embedded template directory.

**[REQ-E-013] LSP Config Exclusion (Unwanted)**

The system shall NOT include `.lsp.json` in the embedded template directory, as it is generated at runtime by the LSP module.

**[REQ-E-014] Git Hook Scripts Exclusion (Unwanted)**

The system shall NOT include `.git-hooks/` shell scripts in the embedded template directory, as the Go edition manages Git hooks through its compiled hook system.

### 4.2 go:embed Integration

**[REQ-E-020] Embed Directive Declaration (Ubiquitous)**

The system shall always declare a `//go:embed` directive in `internal/template/embed.go` that embeds the entire `templates/` directory tree using the `all:` prefix to include dot-prefixed directories.

**[REQ-E-021] Exported Embedded FS (Ubiquitous)**

The system shall always export the embedded filesystem as `var EmbeddedTemplates embed.FS` accessible from the `template` package.

**[REQ-E-022] Path Stripping via fs.Sub (Event-Driven)**

WHEN the embedded FS is consumed by `NewDeployer`, THEN the system shall apply `fs.Sub(EmbeddedTemplates, "templates")` to strip the `templates/` prefix so that embedded paths deploy with the correct relative structure (e.g., `.claude/agents/moai/expert-backend.md`).

**[REQ-E-023] Embed File Location (Ubiquitous)**

The `embed.go` file shall reside at `internal/template/embed.go` within the `template` package, co-located with the Deployer that consumes it.

**[REQ-E-024] Build Validation (Event-Driven)**

WHEN `go build` is executed THEN the compiler shall verify that all paths referenced by the `//go:embed` directive exist, failing the build if any template file is missing.

### 4.3 CLI Integration

**[REQ-E-030] Deployer Wiring (Event-Driven)**

WHEN `moai init` is executed THEN the CLI shall construct a `template.NewDeployer()` using the embedded `fs.FS` (after `fs.Sub` path stripping) and pass it to `project.NewInitializer()`, replacing the current `nil` deployer.

**[REQ-E-031] Full Template Deployment (Event-Driven)**

WHEN `moai init` completes THEN all embedded template files shall be deployed to the project root with correct directory structure, and each file shall be tracked in the manifest with `provenance: template_managed`.

**[REQ-E-032] Existing File Protection (State-Driven)**

IF a file already exists at the deploy target path THEN the deployer shall skip that file and record it as `user_created` or `user_modified` in the manifest, depending on whether a manifest entry already exists.

**[REQ-E-033] Settings JSON Runtime Generation (Event-Driven)**

WHEN `moai init` completes template deployment THEN the system shall separately invoke `template.SettingsGenerator.Generate()` to create `.claude/settings.json` at runtime (ADR-011), independent of the embedded template deployment.

**[REQ-E-034] CLAUDE.md Deployment (Event-Driven)**

WHEN `moai init` deploys templates THEN the system shall deploy the full CLAUDE.md from the embedded templates, replacing the current minimal stub generated by `buildClaudeMDContent()`.

### 4.4 Version Embedding

**[REQ-E-040] Manifest Version Recording (Event-Driven)**

WHEN templates are deployed THEN the manifest `Version` field shall be set to the current ADK version from `pkg/version.GetVersion()`.

**[REQ-E-041] Manifest Deployment Timestamp (Event-Driven)**

WHEN templates are deployed THEN the manifest `DeployedAt` field shall be set to the current UTC time in ISO 8601 format.

### 4.5 Template Update Integration

**[REQ-E-050] Update Template Extraction (Event-Driven)**

WHEN `moai update` executes template refresh (REQ-UPD-008, step 6) THEN the system shall extract new templates from the updated binary's embedded FS and compare against the manifest to determine merge strategy per file provenance.

**[REQ-E-051] Template Managed Overwrite (State-Driven)**

IF a file's manifest provenance is `template_managed` AND the file's current hash matches the deployed hash THEN the system shall safely overwrite the file with the new template version.

**[REQ-E-052] User Modified Merge (State-Driven)**

IF a file's manifest provenance is `user_modified` THEN the system shall invoke the 3-way merge engine with base=`template_hash` version, current=user's file, updated=new template.

**[REQ-E-053] User Created Skip (State-Driven)**

IF a file's manifest provenance is `user_created` THEN the system shall skip the file entirely during template update.

### 4.6 Binary Size Constraint

**[REQ-E-060] Binary Size Budget (Ubiquitous)**

The final compiled binary with all embedded templates shall not exceed 30 MB.

**[REQ-E-061] Template Size Monitoring (Ubiquitous)**

The system shall always verify that the total embedded template size does not exceed 5 MB as part of CI validation.

---

## 5. Specifications

### 5.1 File: `internal/template/embed.go`

```go
package template

import (
    "embed"
    "io/fs"
)

//go:embed all:templates
var embeddedRaw embed.FS

// EmbeddedTemplates returns the embedded template filesystem
// with the "templates/" prefix stripped so paths match deployment targets.
func EmbeddedTemplates() (fs.FS, error) {
    return fs.Sub(embeddedRaw, "templates")
}
```

**Important**: The `//go:embed` directive requires the `templates/` directory to be relative to the file's package directory. Since `embed.go` is in `internal/template/`, the `templates/` directory must exist at `internal/template/templates/` OR the directive must reference a path relative to the module root. Given Go's embed path rules, one of these approaches is needed:

**Option A (Recommended)**: Place `embed.go` at the module root (`cmd/moai/embed.go`) where it can directly reference `../../templates/` -- but Go embed does not support `..` paths.

**Option B (Recommended)**: Place `embed.go` alongside `templates/` at the repository root in a dedicated package, e.g., `pkg/embedded/embed.go` or use `cmd/moai/embed.go` with `templates/` symlinked.

**Option C (Simplest)**: Move `templates/` into `internal/template/templates/` so the `embed.go` can reference it directly.

**Design Decision**: Option C is recommended for simplicity. The `templates/` directory at the repo root becomes `internal/template/templates/`, and `embed.go` in `internal/template/` can directly reference `all:templates`.

### 5.2 CLI Integration Changes: `internal/cli/init.go`

Current (line 83):
```go
initializer := project.NewInitializer(nil, mgr, nil)
```

Updated:
```go
embeddedFS, err := template.EmbeddedTemplates()
if err != nil {
    return fmt.Errorf("load embedded templates: %w", err)
}
deployer := template.NewDeployer(embeddedFS)
initializer := project.NewInitializer(deployer, mgr, nil)
```

### 5.3 Deployer Enhancement: Existing File Protection

The current `Deployer.Deploy()` writes unconditionally. For SPEC-EMBED-001, enhance to skip files that already exist unless `--force` is specified:

```go
// In deployer.go Deploy() method, before writing:
if _, statErr := os.Stat(destPath); statErr == nil {
    // File exists - check manifest for provenance
    if entry, found := m.GetEntry(path); found {
        if entry.Provenance == manifest.UserModified || entry.Provenance == manifest.UserCreated {
            // Skip user files
            continue
        }
    } else {
        // Existing file not in manifest - mark as user_created and skip
        m.Track(path, manifest.UserCreated, "")
        continue
    }
}
```

### 5.4 Version Integration

```go
// In initializer.go initManifest():
mf.Version = version.GetVersion()
mf.DeployedAt = time.Now().UTC().Format(time.RFC3339)
```

### 5.5 File Counts (Estimated)

| Category | Estimated Files | Estimated Size |
|----------|----------------|----------------|
| Agent definitions | 20 | ~592 KB |
| Skill definitions | 423 | ~5.6 MB |
| Rule files | 22 | ~92 KB |
| Output styles | 3 | ~72 KB |
| Config templates | 5 | ~20 KB |
| Root files (CLAUDE.md, .mcp.json, .gitignore, .mcp.windows.json) | 4 | ~14 KB |
| Announcements | 4 | ~4 KB |
| LLM configs | 1 | ~2 KB |
| **Total** | **~482** | **~6.4 MB** |

Note: Skills are the largest category. Binary size impact after compression (Go embeds store files uncompressed but the binary itself may benefit from UPX or similar): ~6.4 MB raw embedded content.

### 5.6 Performance Requirements

| Metric | Target | Measurement |
|--------|--------|-------------|
| Full template deployment (~480 files) | < 2s | Benchmark test |
| Single `moai init` (detection + deploy + config) | < 5s | E2E test |
| Binary size with templates | < 30 MB | CI check |
| Embedded FS file listing | < 10ms | Benchmark test |

### 5.7 Security Requirements

| Item | Implementation |
|------|---------------|
| Path traversal prevention | Existing `validateDeployPath()` in deployer.go |
| No credentials in templates | CI check: grep for API keys, tokens, passwords |
| No dynamic tokens in output | Existing `unexpandedTokenPattern` validation in renderer.go |
| File permissions | 0o644 for files, 0o755 for directories |

---

## 6. Cross References

| Reference | Location | Relation |
|-----------|----------|----------|
| SPEC-TEMPLATE-001 | `.moai/specs/SPEC-TEMPLATE-001/` | Provides engine (Deployer, Renderer, Manifest) |
| SPEC-INIT-001 | `.moai/specs/SPEC-INIT-001/` | Provides initialization flow (consumes Deployer) |
| SPEC-UPDATE-001 | `.moai/specs/SPEC-UPDATE-001/` | Provides update flow (consumes embedded templates + Merge) |
| ADR-003 | design.md | go:embed for Template Distribution |
| ADR-007 | design.md | File Manifest Provenance |
| ADR-011 | design.md | Zero Runtime Template Expansion |
| Python reference | `moai_adk/templates/` | Source content for template curation |

---

## 7. Expert Consultation Recommendations

This SPEC benefits from the following domain expert consultations:

- **expert-backend**: `embed.FS` path resolution strategy, Go build tag considerations for embed, package structure for embed directive placement
- **expert-security**: Template content audit for credential/secret leakage, path traversal validation with real template paths
- **expert-testing**: Integration test strategy for embedded filesystem, CI validation for template completeness and binary size
- **expert-devops**: CI pipeline for template content validation, binary size monitoring, goreleaser integration with embedded content
