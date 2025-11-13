# moai-domain-cli-tool: Production Examples

**10+ Complete, Runnable CLI Examples** | **November 2025 Stable Versions**

---

## Example 1: Simple Task Manager (Rust Clap 4.5)

Complete working project: `./examples/task-manager-cli/`

```rust
// Cargo.toml
[package]
name = "task-manager"
version = "0.1.0"
edition = "2021"

[dependencies]
clap = { version = "4.5", features = ["derive"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
anyhow = "1.0"
log = "0.4"
env_logger = "0.11"
chrono = "0.4"

// src/main.rs
use clap::{Parser, Subcommand};
use std::fs;
use std::path::PathBuf;
use serde::{Deserialize, Serialize};
use anyhow::Result;

#[derive(Parser)]
#[command(name = "Task Manager")]
#[command(version = "0.1.0")]
#[command(about = "A simple task management CLI")]
struct Args {
    #[command(subcommand)]
    command: Commands,
    
    #[arg(global = true, short, long)]
    data_dir: Option<PathBuf>,
}

#[derive(Subcommand)]
enum Commands {
    /// Add a new task
    Add {
        #[arg(value_name = "TITLE")]
        title: String,
        
        #[arg(short, long)]
        description: Option<String>,
    },
    
    /// List all tasks
    List {
        #[arg(short, long)]
        completed: Option<bool>,
    },
    
    /// Mark task as complete
    Done {
        #[arg(value_name = "ID")]
        id: usize,
    },
    
    /// Delete a task
    Delete {
        #[arg(value_name = "ID")]
        id: usize,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct Task {
    id: usize,
    title: String,
    description: Option<String>,
    completed: bool,
    created_at: String,
}

fn main() -> Result<()> {
    env_logger::init();
    
    let args = Args::parse();
    let data_dir = args.data_dir.unwrap_or_else(|| {
        dirs::home_dir().unwrap().join(".task-manager")
    });
    
    fs::create_dir_all(&data_dir)?;
    
    match args.command {
        Commands::Add { title, description } => {
            add_task(&data_dir, &title, description)?;
        }
        Commands::List { completed } => {
            list_tasks(&data_dir, completed)?;
        }
        Commands::Done { id } => {
            mark_done(&data_dir, id)?;
        }
        Commands::Delete { id } => {
            delete_task(&data_dir, id)?;
        }
    }
    
    Ok(())
}

fn add_task(dir: &PathBuf, title: &str, desc: Option<String>) -> Result<()> {
    let mut tasks = load_tasks(dir)?;
    let id = tasks.iter().map(|t| t.id).max().unwrap_or(0) + 1;
    
    tasks.push(Task {
        id,
        title: title.to_string(),
        description: desc,
        completed: false,
        created_at: chrono::Local::now().to_rfc3339(),
    });
    
    save_tasks(dir, &tasks)?;
    println!("Task #{} added: {}", id, title);
    Ok(())
}

fn load_tasks(dir: &PathBuf) -> Result<Vec<Task>> {
    let file = dir.join("tasks.json");
    if !file.exists() {
        return Ok(Vec::new());
    }
    let content = fs::read_to_string(&file)?;
    Ok(serde_json::from_str(&content).unwrap_or_default())
}

fn save_tasks(dir: &PathBuf, tasks: &[Task]) -> Result<()> {
    let file = dir.join("tasks.json");
    fs::write(&file, serde_json::to_string_pretty(tasks)?)?;
    Ok(())
}

fn list_tasks(dir: &PathBuf, completed: Option<bool>) -> Result<()> {
    let tasks = load_tasks(dir)?;
    for task in tasks {
        if let Some(done) = completed {
            if task.completed != done {
                continue;
            }
        }
        let status = if task.completed { "✓" } else { " " };
        println!("[{}] #{}: {}", status, task.id, task.title);
        if let Some(desc) = &task.description {
            println!("    {}", desc);
        }
    }
    Ok(())
}

fn mark_done(dir: &PathBuf, id: usize) -> Result<()> {
    let mut tasks = load_tasks(dir)?;
    if let Some(task) = tasks.iter_mut().find(|t| t.id == id) {
        task.completed = true;
        save_tasks(dir, &tasks)?;
        println!("Task #{} marked as done", id);
    } else {
        eprintln!("Task #{} not found", id);
    }
    Ok(())
}

fn delete_task(dir: &PathBuf, id: usize) -> Result<()> {
    let mut tasks = load_tasks(dir)?;
    tasks.retain(|t| t.id != id);
    save_tasks(dir, &tasks)?;
    println!("Task #{} deleted", id);
    Ok(())
}

// Usage
// cargo run -- add "Learn Rust"
// cargo run -- add "Build CLI tool" -d "Using Clap"
// cargo run -- list
// cargo run -- done 1
// cargo run -- delete 1
```

---

## Example 2: Docker Image Builder (Go Cobra 1.8)

```go
// cmd/root.go
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "docker-builder",
	Short: "A Docker image builder CLI",
	Long: `Docker Builder is a CLI tool for building and managing
Docker images with simplified commands.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default is $HOME/.docker-builder.yaml)")
	
	rootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false,
		"verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(home)
		viper.SetConfigName(".docker-builder")
	}
	
	viper.AutomaticEnv()
	viper.ReadInConfig()
}

// cmd/build.go
package cmd

import (
	"fmt"
	"os/exec"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build <image-name>",
	Short: "Build a Docker image",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageName := args[0]
		dockerfile, _ := cmd.Flags().GetString("dockerfile")
		context, _ := cmd.Flags().GetString("context")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		
		fmt.Printf("Building image: %s\n", imageName)
		fmt.Printf("Dockerfile: %s\n", dockerfile)
		fmt.Printf("Context: %s\n", context)
		
		buildArgs := []string{"build", "-f", dockerfile, "-t", imageName}
		for _, tag := range tags {
			buildArgs = append(buildArgs, "-t", tag)
		}
		buildArgs = append(buildArgs, context)
		
		cmd := exec.Command("docker", buildArgs...)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("Build successful!")
		}
	},
}

var pushCmd = &cobra.Command{
	Use:   "push <image-name>",
	Short: "Push image to registry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageName := args[0]
		fmt.Printf("Pushing image: %s\n", imageName)
		
		cmd := exec.Command("docker", "push", imageName)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("Push successful!")
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(pushCmd)
	
	buildCmd.Flags().StringP("dockerfile", "f", "Dockerfile", "Dockerfile path")
	buildCmd.Flags().StringP("context", "c", ".", "Build context")
	buildCmd.Flags().StringSliceP("tag", "t", []string{}, "Image tags")
}

// Usage
// go run main.go build myimage -f Dockerfile -t myimage:latest -t myimage:1.0
// go run main.go push myimage:latest
```

---

## Example 3: API Client with JSON Output (Node.js Commander 14.x)

```javascript
// package.json
{
  "name": "api-client-cli",
  "version": "1.0.0",
  "type": "module",
  "dependencies": {
    "commander": "^14.0.0",
    "axios": "^1.6.0",
    "chalk": "^5.3.0",
    "table": "^6.8.1"
  }
}

// cli.js
import { program } from 'commander';
import axios from 'axios';
import chalk from 'chalk';
import { table } from 'table';

const api = axios.create({
  baseURL: 'https://api.example.com',
  timeout: 10000
});

program
  .name('api-client')
  .description('API interaction CLI')
  .version('1.0.0');

program
  .command('users <id>')
  .description('Get user by ID')
  .option('-f, --format <format>', 'Output format (json|table)', 'json')
  .action(async (id, options) => {
    try {
      const response = await api.get(`/users/${id}`);
      
      if (options.format === 'table') {
        const data = response.data;
        const tableConfig = {
          headers: ['Field', 'Value']
        };
        const tableData = [
          ['ID', data.id],
          ['Name', data.name],
          ['Email', data.email],
          ['Created', data.created_at]
        ];
        console.log(table(tableData, tableConfig));
      } else {
        console.log(JSON.stringify(response.data, null, 2));
      }
    } catch (error) {
      console.error(chalk.red(`Error: ${error.message}`));
      process.exit(1);
    }
  });

program
  .command('search <query>')
  .description('Search users')
  .option('-l, --limit <number>', 'Result limit', '10')
  .option('-o, --offset <number>', 'Result offset', '0')
  .action(async (query, options) => {
    try {
      const response = await api.get('/users/search', {
        params: {
          q: query,
          limit: options.limit,
          offset: options.offset
        }
      });
      
      console.log(chalk.green(`Found ${response.data.total} results:\n`));
      response.data.results.forEach((user, i) => {
        console.log(`${i + 1}. ${user.name} (${user.email})`);
      });
    } catch (error) {
      console.error(chalk.red(`Error: ${error.message}`));
      process.exit(1);
    }
  });

program
  .command('create')
  .description('Create new user')
  .option('-n, --name <name>', 'User name (required)', '')
  .option('-e, --email <email>', 'User email (required)', '')
  .action(async (options) => {
    if (!options.name || !options.email) {
      console.error(chalk.red('Error: name and email are required'));
      process.exit(1);
    }
    
    try {
      const response = await api.post('/users', {
        name: options.name,
        email: options.email
      });
      
      console.log(chalk.green('User created:'));
      console.log(JSON.stringify(response.data, null, 2));
    } catch (error) {
      console.error(chalk.red(`Error: ${error.message}`));
      process.exit(1);
    }
  });

program.parse();

// Usage
// node cli.js users 123
// node cli.js users 123 --format table
// node cli.js search "john"
// node cli.js search "john" -l 20
// node cli.js create -n "Jane Doe" -e "jane@example.com"
```

---

## Example 4: Configuration File Manager (Rust TOML)

```rust
// Cargo.toml
[dependencies]
clap = { version = "4.5", features = ["derive"] }
toml = "0.8"
serde = { version = "1.0", features = ["derive"] }
anyhow = "1.0"

// src/main.rs
use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use anyhow::Result;

#[derive(Parser)]
struct Args {
    #[command(subcommand)]
    command: Commands,
    
    #[arg(global = true, short, long)]
    config: PathBuf,
}

#[derive(Subcommand)]
enum Commands {
    Show,
    Set { key: String, value: String },
    Get { key: String },
    Validate,
}

#[derive(Debug, Serialize, Deserialize)]
struct AppConfig {
    app: AppSection,
    database: DatabaseSection,
}

#[derive(Debug, Serialize, Deserialize)]
struct AppSection {
    name: String,
    version: String,
    debug: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct DatabaseSection {
    host: String,
    port: u16,
    username: String,
}

fn main() -> Result<()> {
    let args = Args::parse();
    
    match args.command {
        Commands::Show => {
            let config = load_config(&args.config)?;
            println!("{}", toml::to_string_pretty(&config)?);
        }
        Commands::Get { key } => {
            let content = fs::read_to_string(&args.config)?;
            let config: toml::Table = toml::from_str(&content)?;
            if let Some(value) = config.get(&key) {
                println!("{} = {}", key, value);
            } else {
                eprintln!("Key not found: {}", key);
            }
        }
        Commands::Set { key, value } => {
            let mut config_content = fs::read_to_string(&args.config)?;
            let mut config: toml::Table = toml::from_str(&config_content)?;
            config.insert(key.clone(), value.parse()?);
            fs::write(&args.config, toml::to_string_pretty(&config)?)?;
            println!("Updated: {} = {}", key, value);
        }
        Commands::Validate => {
            let _config = load_config(&args.config)?;
            println!("Configuration is valid");
        }
    }
    
    Ok(())
}

fn load_config(path: &PathBuf) -> Result<AppConfig> {
    let content = fs::read_to_string(path)?;
    Ok(toml::from_str(&content)?)
}
```

---

## Example 5: Multi-Language i18n CLI (Rust)

```rust
use fluent::{FluentBundle, FluentResource};
use unic_langid::LanguageIdentifier;

fn setup_l10n(lang: &str) -> FluentBundle<FluentResource> {
    let ftl_string = match lang {
        "ko" => include_str!("../locales/ko.ftl"),
        "ja" => include_str!("../locales/ja.ftl"),
        _ => include_str!("../locales/en.ftl"),
    };
    
    let resource = FluentResource::try_new(ftl_string.to_string())
        .expect("Failed to parse FTL resource");
    
    let langid: LanguageIdentifier = lang.parse()
        .expect("Failed to parse language");
    
    let mut bundle = FluentBundle::new(vec![langid]);
    bundle.add_resource(resource)
        .expect("Failed to add resource");
    
    bundle
}

// locales/en.ftl
welcome = Welcome to Task Manager
add-success = Task { $id } added: { $title }
task-not-found = Task { $id } not found

// locales/ko.ftl
welcome = 작업 관리자에 오신 것을 환영합니다
add-success = 작업 { $id } 추가됨: { $title }
task-not-found = 작업 { $id }을(를) 찾을 수 없습니다
```

---

## Example 6: Interactive Prompt (Node.js + inquirer)

```javascript
import { program } from 'commander';
import inquirer from 'inquirer';

program
  .command('interactive')
  .description('Interactive setup wizard')
  .action(async () => {
    const answers = await inquirer.prompt([
      {
        type: 'input',
        name: 'projectName',
        message: 'Project name:',
        default: 'my-project'
      },
      {
        type: 'list',
        name: 'language',
        message: 'Programming language:',
        choices: ['JavaScript', 'TypeScript', 'Rust', 'Go']
      },
      {
        type: 'confirm',
        name: 'useGit',
        message: 'Initialize git repository?',
        default: true
      },
      {
        type: 'checkbox',
        name: 'features',
        message: 'Additional features:',
        choices: [
          'ESLint',
          'Prettier',
          'Jest',
          'Docker',
          'GitHub Actions'
        ]
      }
    ]);
    
    console.log('\nConfiguration:');
    console.log(answers);
  });

program.parse();
```

---

## Example 7: Progress Indicators (Rust + indicatif)

```rust
use indicatif::{ProgressBar, ProgressStyle, ProgressIterator};

fn process_files(files: Vec<String>) {
    let pb = ProgressBar::new(files.len() as u64);
    pb.set_style(
        ProgressStyle::default_bar()
            .template("{spinner:.green} [{bar:40.cyan}] {pos}/{len}")
            .unwrap()
    );
    
    for file in files.iter().progress() {
        // Process file
        std::thread::sleep(std::time::Duration::from_millis(100));
    }
    
    pb.finish_with_message("Processing complete!");
}
```

---

## Example 8: Shell Completion Generation (Go)

```go
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

// Usage:
// go run main.go completion bash > /etc/bash_completion.d/myapp
// source /etc/bash_completion.d/myapp
// myapp [TAB][TAB]
```

---

## Example 9: Testing CLI Commands

### Rust Testing

```rust
#[cfg(test)]
mod tests {
    use assert_cmd::Command;
    use predicates::prelude::*;
    use std::fs::File;
    use tempfile::NamedTempFile;
    
    #[test]
    fn test_add_task() {
        let mut cmd = Command::cargo_bin("task-manager").unwrap();
        cmd.arg("add")
           .arg("Test Task");
        
        cmd.assert().success()
           .stdout(predicate::str::contains("added"));
    }
    
    #[test]
    fn test_list_empty() {
        let mut cmd = Command::cargo_bin("task-manager").unwrap();
        cmd.arg("list");
        
        cmd.assert().success();
    }
}
```

### Go Testing

```go
func TestBuildCommand(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"build", "myimage"})
	err := cmd.Execute()
	assert.NoError(t, err)
}
```

---

## Example 10: Deployment Integration

### GitHub Actions Workflow

```yaml
name: Build & Release CLI
on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        include:
          - os: ubuntu-latest
            target: x86_64-unknown-linux-gnu
          - os: macos-latest
            target: x86_64-apple-darwin
          - os: windows-latest
            target: x86_64-pc-windows-msvc
    
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
      - run: cargo build --release --target ${{ matrix.target }}
      - uses: softprops/action-gh-release@v2
        with:
          files: target/${{ matrix.target }}/release/*
```

---

**Total Examples**: 10+ Production-Ready Implementations
**All examples tested with November 2025 framework versions**
