---
name: "moai-domain-cli-tool"
version: "4.0.0"
tier: Domain Architecture
created: 2025-11-13
updated: 2025-11-13
tags: [CLI, Rust, Go, Node.js, command-line-tools, argument-parsing, UX/DX]
trigger_keywords: [CLI tool, command-line, argparse, Clap, Cobra, Commander]
allowed-tools: [WebSearch, WebFetch]
status: stable
description: "Enterprise CLI tool architecture with multi-language patterns, UX/DX optimization, and production deployment strategies"
---

# Enterprise CLI Tool Architecture - v4.0.0

## Level 1: Quick Reference

### Core CLI Patterns

**Rust Clap v4.5 Foundation**:
```rust
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "myapp")]
#[command(about = "A great CLI tool")]
#[command(version)]
struct Args {
    #[command(subcommand)]
    command: Option<Commands>,
    #[arg(short, long)]
    verbose: bool,
    #[arg(short, long)]
    config: Option<String>,
}

#[derive(Subcommand)]
enum Commands {
    Run {
        #[arg(value_name = "FILE")]
        input: String,
        #[arg(short, long)]
        output: Option<String>,
    },
    Status,
}

fn main() {
    let args = Args::parse();
    match args.command {
        Some(Commands::Run { input, output }) => {
            println!("Running with input: {}", input);
            if let Some(out) = output {
                println!("Output to: {}", out);
            }
        }
        Some(Commands::Status) => println!("Status: OK"),
        None => println!("No command provided"),
    }
}
```

**Go Cobra v1.8 Foundation**:
```go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "A great CLI tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("myapp running")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("output", "o", "", "Output file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
```

**Node Commander v14.x Foundation**:
```javascript
import { program } from 'commander';

program
  .name('myapp')
  .description('A great CLI tool')
  .version('1.0.0');

program
  .command('run <file>')
  .description('Run the application')
  .option('-o, --output <file>', 'Output file')
  .action((file, options) => {
    console.log(`Running with: ${file}`);
    if (options.output) {
      console.log(`Output to: ${options.output}`);
    }
  });

program.parse();
```

**Core Principles**:
- ✅ Clear subcommand structure
- ✅ Type safety (Rust/Go)
- ✅ Auto help generation
- ✅ Version management
- ✅ Consistent flag naming

---

## Level 2: Core Implementation

### Advanced CLI Features

**Plugin Architecture** (Rust):
```rust
use std::fs;
use std::path::PathBuf;

pub trait CliPlugin: Send + Sync {
    fn name(&self) -> &str;
    fn version(&self) -> &str;
    fn execute(&self, args: &[String]) -> Result<(), String>;
}

pub struct PluginLoader {
    plugin_dir: PathBuf,
}

impl PluginLoader {
    pub fn discover_plugins(&self) -> Result<Vec<String>, String> {
        let entries = fs::read_dir(&self.plugin_dir)
            .map_err(|e| e.to_string())?;
        
        let mut plugins = Vec::new();
        for entry in entries.flatten() {
            let path = entry.path();
            if path.extension().and_then(|s| s.to_str()) == Some("so") {
                if let Some(name) = path.file_stem() {
                    if let Some(name_str) = name.to_str() {
                        plugins.push(name_str.to_string());
                    }
                }
            }
        }
        Ok(plugins)
    }
}
```

**Configuration Management** (Go + YAML):
```go
type Config struct {
	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Debug   bool   `yaml:"debug"`
	} `yaml:"app"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
	} `yaml:"database"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	return &cfg, nil
}
```

**Shell Completion** (Node.js):
```javascript
class ShellCompletion {
  generateBashCompletion(commands, options) {
    const template = `
_myapp_completion() {
  local cur prev commands
  COMPREPLY=()
  cur="\${COMP_WORDS[COMP_CWORD]}"
  commands="${commands.join(' ')}"
  
  if [[ \${cur} == -* ]]; then
    COMPREPLY=($(compgen -W "${options.join(' ')}" -- \${cur}))
  else
    COMPREPLY=($(compgen -W "\${commands}" -- \${cur}))
  fi
}
complete -F _myapp_completion myapp`;
    return template.trim();
  }
}
```

**Error Handling & UX**:
```go
type CliError struct {
	Code    int
	Message string
	Details string
}

func (e *CliError) Error() string {
	return fmt.Sprintf("[ERROR %d] %s: %s", e.Code, e.Message, e.Details)
}

func exitWithError(code int, message string, details string) {
	err := &CliError{Code: code, Message: message, Details: details}
	log.Printf("FATAL: %v", err)
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(code)
}
```

---

## Level 3: Advanced Features

### Production-Grade CLI Architecture

**Complete Enterprise CLI** (Rust):
```rust
use clap::{Parser, Subcommand};
use anyhow::{Result, Context};

#[derive(Parser)]
#[command(name = "enterprise-cli")]
#[command(version = "1.0.0")]
struct Args {
    #[command(subcommand)]
    command: Commands,
    #[arg(global = true, short, long)]
    config: Option<PathBuf>,
    #[arg(global = true, short, long)]
    verbose: u8,
}

#[derive(Subcommand)]
enum Commands {
    Init {
        #[arg(value_name = "PROJECT")]
        name: String,
        #[arg(short, long)]
        template: Option<String>,
    },
    Build {
        #[arg(short, long)]
        release: bool,
        #[arg(short, long)]
        target: Option<String>,
    },
    Deploy {
        #[arg(value_name = "ENV")]
        environment: String,
        #[arg(short, long)]
        dry_run: bool,
    },
}

async fn handle_init(name: String, template: Option<String>) -> Result<()> {
    fs::create_dir_all(&name)?;
    
    let config = ProjectConfig {
        name: name.clone(),
        version: "0.1.0".to_string(),
        targets: vec!["x86_64-unknown-linux-gnu".to_string()],
    };
    
    let config_content = serde_yaml::to_string(&config)?;
    fs::write(format!("{}/.cli-config.yaml", name), config_content)?;
    
    println!("Project '{}' initialized successfully!", name);
    Ok(())
}
```

**Multi-Platform Deployment**:
```yaml
name: Release CLI
on:
  push:
    tags: ['v*']
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: cargo build --release
      - uses: softprops/action-gh-release@v2
        with:
          files: target/release/myapp
```

**Docker Distribution**:
```dockerfile
FROM rust:1.75 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=builder /app/target/release/myapp /usr/local/bin/
ENTRYPOINT ["myapp"]
```

---

## Level 4: Reference & Integration

### Best Practices

**User Experience (UX)**:
- ✅ Clear error messages with hints
- ✅ Progress indicators for long operations
- ✅ Consistent command structure
- ✅ Auto-completion support

**Developer Experience (DX)**:  
- ✅ Type-safe argument parsing
- ✅ Comprehensive error handling
- ✅ Modular architecture
- ✅ Plugin system support

**Performance Guidelines**:
- ✅ < 100ms startup time (native)
- ✅ 100-500ms acceptable (Node.js/Python)
- ✅ > 2s requires optimization

### Framework Versions
- **Rust Clap**: 4.5 (latest)
- **Go Cobra**: 1.8 (stable)
- **Node Commander**: 14.x (current)

**Related Skills**:
- `Skill("moai-domain-devops")` for deployment patterns
- `Skill("moai-essentials-perf")` for performance optimization
- `Skill("moai-security-backend")` for CLI security

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready
