# CLI Tool Advanced Patterns

## Plugin System Architecture

### Dynamic Plugin Loading (Rust)

```rust
use libloading::{Library, Symbol};
use std::path::Path;

pub trait Plugin: Send + Sync {
    fn name(&self) -> &'static str;
    fn version(&self) -> &'static str;
    fn execute(&self, args: &[String]) -> Result<(), String>;
    fn help(&self) -> &'static str;
}

pub struct PluginManager {
    plugins: std::collections::HashMap<String, Box<dyn Plugin>>,
    lib_dir: std::path::PathBuf,
}

impl PluginManager {
    pub fn new(lib_dir: impl AsRef<Path>) -> Self {
        Self {
            plugins: std::collections::HashMap::new(),
            lib_dir: lib_dir.as_ref().to_path_buf(),
        }
    }

    pub fn load_plugins(&mut self) -> Result<(), String> {
        let entries = std::fs::read_dir(&self.lib_dir)
            .map_err(|e| format!("Failed to read plugin dir: {}", e))?;

        for entry in entries.flatten() {
            let path = entry.path();
            if path.extension().map(|ext| ext == "so").unwrap_or(false) {
                self.load_plugin(&path)?;
            }
        }
        Ok(())
    }

    fn load_plugin(&mut self, path: &std::path::Path) -> Result<(), String> {
        unsafe {
            let lib = Library::new(path)
                .map_err(|e| format!("Failed to load {}: {}", path.display(), e))?;

            let constructor: Symbol<fn() -> *mut dyn Plugin> = lib.get(b"_plugin_create")
                .map_err(|e| format!("Failed to get plugin constructor: {}", e))?;

            let object = constructor();
            let plugin = Box::from_raw(object);
            self.plugins.insert(plugin.name().to_string(), plugin);
        }
        Ok(())
    }
}
```

### Go Plugin Interface

```go
package main

import (
    "fmt"
    "plugin"
)

type PluginInterface interface {
    Name() string
    Version() string
    Execute(args []string) error
    Help() string
}

type PluginRegistry struct {
    plugins map[string]PluginInterface
}

func (pr *PluginRegistry) LoadPlugin(path string) error {
    p, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to load plugin: %w", err)
    }

    symCreate, err := p.Lookup("NewPlugin")
    if err != nil {
        return fmt.Errorf("plugin missing NewPlugin: %w", err)
    }

    creator := symCreate.(func() PluginInterface)
    pluginInstance := creator()

    pr.plugins[pluginInstance.Name()] = pluginInstance
    return nil
}

func (pr *PluginRegistry) Execute(name string, args []string) error {
    if plugin, exists := pr.plugins[name]; exists {
        return plugin.Execute(args)
    }
    return fmt.Errorf("plugin not found: %s", name)
}
```

## Custom Argument Parsing Strategies

### Complex Argument Types (TypeScript)

```typescript
import { program } from 'commander';

class ParsedPath {
    constructor(public full: string, public dir: string, public file: string) {}
}

function parseFilePath(value: string): ParsedPath {
    const path = require('path');
    return new ParsedPath(
        value,
        path.dirname(value),
        path.basename(value)
    );
}

class EnvironmentConfig {
    constructor(
        public production: boolean,
        public debug: boolean,
        public port: number
    ) {}

    static parse(envStr: string): EnvironmentConfig {
        const parts = envStr.split(':');
        return new EnvironmentConfig(
            parts[0] === 'prod',
            parts[1] === 'debug',
            parseInt(parts[2]) || 3000
        );
    }
}

program
    .command('deploy <file>')
    .option('-e, --env <config>', 'Environment config', EnvironmentConfig.parse)
    .action((file: string, options: { env: EnvironmentConfig }) => {
        const parsedPath = parseFilePath(file);
        console.log(`Deploying ${parsedPath.file} to ${options.env.production ? 'Production' : 'Dev'}`);
    });

program.parse();
```

## Shell Integration Patterns

### Bash Completion with Dynamic Data (Go)

```go
package main

import (
    "fmt"
    "os"
    "strings"
    "github.com/spf13/cobra"
)

var projects []string

func init() {
    // Load projects from database or config
    projects = []string{"project-a", "project-b", "project-c"}
}

var deployCmd = &cobra.Command{
    Use:   "deploy <project>",
    Short: "Deploy a project",
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        if len(args) != 0 {
            return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var filtered []string
        for _, p := range projects {
            if strings.HasPrefix(p, toComplete) {
                filtered = append(filtered, p)
            }
        }
        return filtered, cobra.ShellCompDirectiveNoFileComp
    },
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Deploying: %s\n", args[0])
    },
}

func init() {
    rootCmd.AddCommand(deployCmd)
}
```

### Zsh Completion with Function Groups

```bash
_cli_completion() {
    local -a commands environments

    commands=(
        "init:Initialize new project"
        "build:Build the project"
        "deploy:Deploy to environment"
        "test:Run test suite"
        "serve:Start development server"
    )

    environments=(
        "dev:Development"
        "staging:Staging"
        "prod:Production"
    )

    if (( CURRENT == 2 )); then
        _describe 'commands' commands
    elif [[ ${words[2]} == deploy ]]; then
        _describe 'environments' environments
    fi
}

compdef _cli_completion mycli
```

## Context7 Integration

### Latest CLI Framework Patterns

The Context7 MCP integration provides access to:
- **Click** (Python): Latest decorators and option handling
- **Typer** (Python): Modern type-hint based CLI
- **Commander.js** (Node): Advanced command composition
- **Cobra** (Go): Plugin ecosystem and middleware patterns
- **Clap** (Rust): Advanced derive macros and validation

### Using Context7 for Framework Selection

```python
# Fetch latest CLI patterns from Context7
from moai_context7 import Context7Client

async def select_cli_framework(requirements: dict) -> str:
    client = Context7Client()

    frameworks = await client.get_library_docs(
        context7_library_id="/cli-frameworks",
        topic="argument parsing validation error handling 2025",
        tokens=3000
    )

    # Match requirements against patterns
    best_match = match_requirements(requirements, frameworks)
    return best_match
```

## Advanced Error Handling

### Structured Error Reporting (Rust with anyhow)

```rust
use anyhow::{anyhow, Context, Result};
use std::io;

#[derive(Debug)]
pub struct CliError {
    code: i32,
    message: String,
    context: Vec<String>,
}

impl CliError {
    pub fn new(code: i32, message: impl Into<String>) -> Self {
        Self {
            code,
            message: message.into(),
            context: Vec::new(),
        }
    }

    pub fn with_context(mut self, ctx: impl Into<String>) -> Self {
        self.context.push(ctx.into());
        self
    }

    pub fn report(&self) {
        eprintln!("Error [{}]: {}", self.code, self.message);
        for (i, ctx) in self.context.iter().enumerate() {
            eprintln!("  {}: {}", i + 1, ctx);
        }
    }
}

pub fn handle_file_operation(path: &str) -> Result<String> {
    std::fs::read_to_string(path)
        .with_context(|| format!("Failed to read file: {}", path))
        .map_err(|e| anyhow!("IO operation failed: {}", e))
}
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready

## Context7 Integration

### Related Libraries & Tools
- [Click](/pallets/click): Python CLI framework
- [Typer](/tiangolo/typer): Modern Python CLI with type hints
- [Commander.js](/tj/commander.js): Node.js command-line interface
- [Cobra](/spf13/cobra): Go command-line interface
- [Clap](/clap-rs/clap): Rust command-line parser

### Official Documentation
- [Click Documentation](https://click.palletsprojects.com/)
- [Typer Documentation](https://typer.tiangolo.com/)
- [Commander.js Docs](https://github.com/tj/commander.js)
- [Cobra Documentation](https://cobra.dev/)
- [Clap Book](https://docs.rs/clap/latest/clap/)
