# Data Flow

Key data paths through moai-adk-go for the five most important operations.

---

## 1. Template Deployment Flow (`moai init`)

The primary purpose of `moai init` is to extract embedded templates into a new project directory, render Go text/template variables, and write all configuration files.

```mermaid
sequenceDiagram
    participant User
    participant CLI as cli/init.go
    participant UI as internal/ui
    participant Proj as core/project/Initializer
    participant TmplCtx as template/TemplateContext
    participant Deployer as template/Deployer
    participant EmbedFS as embed.FS (templates/)
    participant Renderer as template/Renderer
    participant Manifest as internal/manifest
    participant FS as Filesystem

    User->>CLI: moai init [project-name]
    CLI->>UI: RunInitWizard() — collect language, model policy, mode
    UI-->>CLI: WizardResult{lang, policy, mode, name}
    CLI->>Proj: Initializer.Initialize(projectRoot, config)
    Proj->>TmplCtx: NewTemplateContext(GoBinPath, HomeDir, ...)
    Proj->>Deployer: Deploy(ctx, projectRoot, DeployerMode=Init)

    loop For each file in embedded FS
        Deployer->>EmbedFS: ReadFile(path)
        EmbedFS-->>Deployer: []byte content
        alt File is .tmpl
            Deployer->>Renderer: Render(content, TemplateContext)
            Renderer-->>Deployer: rendered []byte
        end
        Deployer->>FS: SafeWrite(destPath, content)
        Deployer->>Manifest: Record(path, hash, version)
    end

    Deployer->>Manifest: Save(.moai/manifest.json)
    Deployer-->>Proj: DeployResult
    Proj->>CLI: InitResult{filesWritten, skipped}
    CLI->>User: "Project initialized at ./project-name"
```

**Key data transformations:**

| Stage | Input | Output |
|-------|-------|--------|
| Wizard | User prompts | `WizardResult` struct |
| TemplateContext | OS env vars, Go paths | `TemplateContext{GoBinPath, HomeDir}` |
| Renderer | `.tmpl` file bytes + context | Rendered file bytes (variables substituted) |
| ModelPolicy | Agent `.md` files + policy level | Modified `model:` frontmatter fields |
| Manifest | File paths + content hashes | `.moai/manifest.json` (deploy record) |

---

## 2. Hook Execution Flow (`moai hook <event>`)

Claude Code invokes `moai hook <event>` via shell wrappers whenever a hook event fires. The process reads JSON from stdin, dispatches to handlers, and writes a JSON response to stdout.

```mermaid
sequenceDiagram
    participant ClaudeCode as Claude Code
    participant Shell as .claude/hooks/moai/handle-*.sh
    participant CLI as cli/hook.go
    participant Protocol as hook/Protocol
    participant Registry as hook/Registry
    participant Handler as hook/Handler (specific)
    participant LSP as internal/lsp
    participant Config as internal/config

    ClaudeCode->>Shell: Execute hook script (stdin = JSON payload)
    Shell->>CLI: moai hook pre-tool-use (stdin piped)
    CLI->>Protocol: ReadInput(os.Stdin)
    Protocol-->>CLI: HookPayload{eventType, toolName, toolInput, ...}
    CLI->>Registry: Dispatch(eventType, payload)
    Registry->>Handler: Handle(ctx, payload)

    alt PreToolUse event
        Handler->>Handler: security.Scan(toolInput)
        Handler-->>Registry: HookResponse{decision: "allow"|"block", message}
    else TeammateIdle event
        Handler->>LSP: DiagnosticsCollector.Collect(cwd)
        LSP-->>Handler: DiagnosticResult{errors, warnings}
        alt errors > 0
            Handler-->>Registry: HookResponse{exitCode: 2}  -- keep working
        else clean
            Handler-->>Registry: HookResponse{exitCode: 0}  -- accept idle
        end
    else SessionEnd event
        Handler->>Config: Load() → session metadata
        Handler->>Handler: FormatMetrics(toolCounts, duration)
        Handler-->>Registry: HookResponse{message: summary}
    end

    Registry-->>CLI: HookResponse
    CLI->>Protocol: WriteOutput(os.Stdout, response)
    Protocol->>Shell: JSON response (stdout)
    Shell->>ClaudeCode: Exit code + stdout JSON
```

**Hook response exit codes:**

| Exit Code | Meaning | Events That Use It |
|-----------|---------|-------------------|
| `0` | Success / allow | All events |
| `1` | Error / deny | `PreToolUse`, `PermissionRequest` |
| `2` | Special signal | `TeammateIdle` (keep working), `TaskCompleted` (reject), `PermissionRequest` (auto-approve) |

---

## 3. Project Initialization Flow (Full `moai init` with Git)

This shows the broader initialization beyond template deployment, including Git setup and settings generation.

```mermaid
flowchart TD
    A["User: moai init myproject"] --> B["UI Wizard: lang, model-policy, mode"]
    B --> C["core/project.Initializer.Initialize()"]
    C --> D{"Project exists?"}
    D -- No --> E["Create project directory"]
    D -- Yes --> F["Validate not already initialized"]
    E --> G["template.Deployer.Deploy(Init mode)"]
    F --> G
    G --> H["Extract all embedded templates"]
    H --> I["Render .tmpl files with TemplateContext"]
    I --> J["Apply --model-policy to agent files"]
    J --> K["template/settings.go: Generate settings.json"]
    K --> L["Write hook wrapper shell scripts"]
    L --> M["manifest.Save(.moai/manifest.json)"]
    M --> N{"git init needed?"}
    N -- Yes --> O["core/git.Manager.Init()"]
    N -- No --> P["Skip git init"]
    O --> P
    P --> Q["Write CLAUDE.md from template"]
    Q --> R["Display success + next steps"]
```

---

## 4. Template Update Flow (`moai update`)

`moai update` must safely merge new template versions with user modifications. It uses 3-way merge to preserve customizations.

```mermaid
sequenceDiagram
    participant User
    participant CLI as cli/update.go
    participant Manifest as internal/manifest
    participant EmbedFS as embed.FS (new templates)
    participant Merger as internal/merge
    participant FS as Filesystem

    User->>CLI: moai update
    CLI->>Manifest: Load(.moai/manifest.json)
    Manifest-->>CLI: FileRecords{path, originalHash, deployedVersion}

    loop For each file in new embedded templates
        CLI->>EmbedFS: ReadFile(newPath)
        EmbedFS-->>CLI: newContent []byte
        CLI->>FS: ReadFile(currentPath) → currentContent
        CLI->>Manifest: GetOriginal(path) → originalContent

        alt File not in manifest (new file)
            CLI->>FS: Write(path, newContent)
        else User did not modify (hash matches original)
            CLI->>FS: Write(path, newContent)
        else User modified file
            CLI->>Merger: Merge3Way(original, current, newTemplate)
            alt Clean merge
                Merger-->>CLI: mergedContent, conflicts=nil
                CLI->>FS: Write(path, mergedContent)
            else Has conflicts
                Merger-->>CLI: partialContent, conflicts=[...]
                CLI->>FS: Write(path, contentWithMarkers)
                CLI->>User: Warn "conflict in file X — review needed"
            end
        end

        CLI->>Manifest: UpdateRecord(path, newHash, newVersion)
    end

    CLI->>Manifest: Save(.moai/manifest.json)
    CLI->>User: "Update complete. N files updated, M conflicts."
```

---

## 5. Multi-Model (GLM/CG) Mode Flow

The `moai cg` command enables a hybrid execution model where Claude Code acts as the orchestrator (leader) and GLM-powered sessions act as workers in isolated git worktrees.

```mermaid
flowchart TD
    A["User: moai cg"] --> B["cli/cg.go: Detect tmux session"]
    B --> C{"tmux running?"}
    C -- No --> D["Error: tmux required for cg mode"]
    C -- Yes --> E["Read GLM API key from rank.CredentialStore"]
    E --> F{"API key stored?"}
    F -- No --> G["Prompt: moai glm <api-key> first"]
    F -- Yes --> H["Open tmux split pane"]

    H --> I["Leader pane (left)"]
    H --> J["Worker pane (right)"]

    I --> K["Set leader env:\nRemove GLM env vars\nClaude uses its own model"]
    J --> L["Set worker env:\nANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.7-air\nANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7\nANTHROPIC_DEFAULT_OPUS_MODEL=glm-5"]

    K --> M["Leader: claude code session\n(Claude Opus/Sonnet)\norchestrates via Task()"]
    L --> N["Worker: claude code session\n(GLM models)\nexecutes in worktree isolation"]

    M --> O["Leader delegates via Task()\nwith team_name + isolation:worktree"]
    O --> N
    N --> P["Worker writes files in\n.claude/worktrees/<name>/"]
    P --> Q["Worker returns result via SendMessage"]
    Q --> M
```

### Data Flow: `moai glm` (Settings Manipulation)

```
User: moai glm sk-xxxx
  → cli/glm.go
    → Read ~/.claude/settings.json
    → rank.CredentialStore.Save("glm-api-key", "sk-xxxx")
    → Patch settings.json env section:
        ANTHROPIC_DEFAULT_HAIKU_MODEL  = "glm-4.7-air"
        ANTHROPIC_DEFAULT_SONNET_MODEL = "glm-4.7"
        ANTHROPIC_DEFAULT_OPUS_MODEL   = "glm-5"
    → Write ~/.claude/settings.json
    → Print confirmation
```

### Data Flow: `moai cc` (Revert to Claude)

```
User: moai cc
  → cli/cc.go
    → Read ~/.claude/settings.json
    → Remove ANTHROPIC_DEFAULT_*_MODEL entries from env section
    → Write ~/.claude/settings.json
    → Print "Claude-only mode active"
```

---

## 6. LSP Quality Gate Flow (TeammateIdle Hook)

This is the quality enforcement mechanism that prevents teammates from going idle while LSP errors exist.

```mermaid
sequenceDiagram
    participant Claude as Claude Code (teammate)
    participant Shell as handle-teammate-idle.sh
    participant Hook as hook/teammate_idle.go
    participant LSP as lsp/DiagnosticsCollector
    participant Fallback as lsp/FallbackDiagnostics
    participant Config as config/ConfigManager

    Claude->>Shell: TeammateIdle event (JSON stdin)
    Shell->>Hook: moai hook teammate-idle

    Hook->>Config: Load() → quality.enforce_quality
    alt enforce_quality = false
        Hook-->>Shell: exit 0 (accept idle)
    else enforce_quality = true
        Hook->>LSP: Collect(cwd)
        alt LSP client available
            LSP->>LSP: Query language server
        else No LSP client
            LSP->>Fallback: RunFallback(cwd)
            Fallback->>Fallback: go vet ./...
            Fallback->>Fallback: golangci-lint run
            Fallback-->>LSP: DiagnosticResult
        end
        LSP-->>Hook: DiagnosticResult{errors:N, warnings:M}

        alt errors > 0 (quality gate: max_errors=0)
            Hook-->>Shell: exit 2 (keep working — fix errors)
            Shell-->>Claude: TeammateIdle rejected, errors remain
        else errors == 0
            Hook-->>Shell: exit 0 (accept idle)
            Shell-->>Claude: TeammateIdle accepted
        end
    end
```

**Quality gate thresholds** (from `.moai/config/sections/quality.yaml`):

| Phase | `max_errors` | `max_type_errors` | `max_lint_errors` | Action on violation |
|-------|-------------|-------------------|-------------------|---------------------|
| `run` | 0 | 0 | 0 | TeammateIdle → exit 2 |
| `sync` | 0 | — | — | TeammateIdle → exit 2 |

---

## Data Structure: Hook Payload

All hook events share a common JSON envelope read from stdin:

```
{
  "event":     "TeammateIdle" | "PreToolUse" | "SessionStart" | ...,
  "sessionId": "uuid",
  "toolName":  "Bash" | "Write" | "Read" | ...,   // PreToolUse only
  "toolInput": { ... },                              // PreToolUse only
  "cwd":       "/path/to/project",
  "hookId":    "uuid",
  "timestamp": "ISO8601"
}
```

Response written to stdout:

```
{
  "decision":  "allow" | "block",   // PreToolUse / PermissionRequest
  "message":   "human-readable",
  "reason":    "machine-readable"
}
```

Exit code carries the primary signal; stdout JSON provides human-readable context.
