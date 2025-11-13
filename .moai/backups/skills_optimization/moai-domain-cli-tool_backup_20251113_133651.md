---
name: "moai-domain-cli-tool"
version: "4.0.0"
tier: Domain Architecture
created: 2025-11-12
updated: 2025-11-12
tags: [CLI, Rust, Go, Node.js, command-line-tools, argument-parsing, UX/DX]
trigger_keywords: [CLI tool, command-line, argparse, Clap, Cobra, Commander]
allowed_tools: [WebSearch, WebFetch]
status: stable
description: "Enterprise Skill for advanced development"
---

# moai-domain-cli-tool: Enterprise CLI Tool Design & Architecture (v4.0.0)

**최종 업데이트**: November 2025 | **안정 버전 기준**: Rust Clap 4.5, Go Cobra 1.x, Node Commander 14.x

---

## Purpose & Scope

CLI 도구 설계의 **핵심 아키텍처**를 다룹니다. 사용자 경험(UX), 개발자 경험(DX), 프로덕션 배포를 고려한 **엔터프라이즈 CLI 작성법**을 제공합니다.

**이 Skill은 다음을 가능하게 합니다**:
- 다국어 지원 CLI 구축 (i18n/l10n)
- 자동 help 생성 및 man page 제공
- 플러그인 아키텍처 설계
- 테스트 가능한 CLI 구조
- 설정 파일 (YAML/TOML/JSON) 통합
- 진행 표시기 및 대화형 프롬프트
- 셸 자동완성 (bash, zsh, fish, PowerShell)

---

## Progressive Disclosure

### Level 1: Foundations (Beginner-Friendly)

**기초 CLI 구조**

```rust
// Rust Clap v4.5
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "myapp")]
#[command(about = "A great CLI tool", long_about = None)]
#[command(version)]
#[command(author = "Your Name")]
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
    /// Run the application
    Run {
        #[arg(value_name = "FILE")]
        input: String,
        
        #[arg(short, long)]
        output: Option<String>,
    },
    /// Show status
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

**Go Cobra v1.8 기초**

```go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "A great CLI tool",
	Long: `A longer description explaining the tool's purpose
and capabilities in detail.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("myapp running")
	},
}

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run the application",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Running with: %s\n", args[0])
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

**Node Commander v14.x 기초**

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

program
  .command('status')
  .description('Show status')
  .action(() => {
    console.log('Status: OK');
  });

program.parse();
```

**핵심 원칙**:
- 명확한 서브명령어 구조
- 타입 안전성 (Rust/Go)
- 자동 help 생성
- 버전 관리

---

### Level 2: Advanced Features

**플러그인 아키텍처** (Rust)

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
    pub fn new(plugin_dir: PathBuf) -> Self {
        Self { plugin_dir }
    }
    
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

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_plugin_discovery() {
        let loader = PluginLoader::new(PathBuf::from("./plugins"));
        let plugins = loader.discover_plugins();
        assert!(plugins.is_ok());
    }
}
```

**설정 파일 통합** (Go + YAML)

```go
package main

import (
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v3"
)

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
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
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

// Testing
func TestConfigLoading(t *testing.T) {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.App.Name == "" {
		t.Error("App name should not be empty")
	}
}
```

**셸 자동완성** (Node.js + Commander)

```javascript
import { program } from 'commander';
import { execSync } from 'child_process';
import * as fs from 'fs';

class ShellCompletion {
  generateBashCompletion(commands, options) {
    const commandList = commands.join(' ');
    const template = `
_myapp_completion() {
  local cur prev
  COMPREPLY=()
  cur="\${COMP_WORDS[COMP_CWORD]}"
  prev="\${COMP_WORDS[COMP_CWORD-1]}"
  
  local commands="${commandList}"
  
  if [[ \${cur} == -* ]]; then
    COMPREPLY=($(compgen -W "${options.join(' ')}" -- \${cur}))
  else
    COMPREPLY=($(compgen -W "\${commands}" -- \${cur}))
  fi
}

complete -F _myapp_completion myapp
    `;
    return template.trim();
  }
  
  installCompletion(shell = 'bash') {
    const completion = this.generateBashCompletion(
      ['run', 'status', 'config'],
      ['-v', '--verbose', '-o', '--output']
    );
    
    const targetFile = shell === 'bash' 
      ? `${process.env.HOME}/.bashrc`
      : `${process.env.HOME}/.zshrc`;
      
    fs.appendFileSync(targetFile, `\n${completion}\n`);
    console.log(`Completion installed for ${shell}`);
  }
}

const completer = new ShellCompletion();
program
  .command('completion <shell>')
  .description('Generate shell completion')
  .action((shell) => {
    completer.installCompletion(shell);
  });
```

**진행 표시기 및 대화형 프롬프트**

```rust
use indicatif::{ProgressBar, ProgressStyle};
use std::time::Duration;
use std::thread;

pub fn long_running_task() {
    let pb = ProgressBar::new(100);
    pb.set_style(ProgressStyle::default_bar()
        .template("{spinner:.green} [{elapsed_precise}] [{bar:40.cyan/blue}] {pos}/{len} ({eta})")
        .unwrap()
        .progress_chars("#>-"));
    
    for i in 0..100 {
        pb.inc(1);
        thread::sleep(Duration::from_millis(50));
    }
    
    pb.finish_with_message("Task completed!");
}

// Interactive prompts
pub fn get_user_confirmation(prompt: &str) -> bool {
    use std::io::{self, Write};
    
    print!("{} [y/N]: ", prompt);
    io::stdout().flush().unwrap();
    
    let mut input = String::new();
    io::stdin().read_line(&mut input).unwrap();
    
    input.trim().to_lowercase() == "y"
}
```

**핵심 패턴**:
- 플러그인 시스템 (확장성)
- YAML/TOML 설정 파일
- 셸 자동완성 생성
- 진행 표시기 및 대화형 UI
- 테스트 용이한 구조

---

### Level 3: Production Architecture

**완전한 프로덕션 CLI (Rust)**

```rust
use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use log::{info, error};
use anyhow::{Result, Context};

#[derive(Parser)]
#[command(name = "enterprise-cli")]
#[command(version = "1.0.0")]
#[command(about = "Enterprise CLI Tool")]
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
    /// Initialize project
    Init {
        #[arg(value_name = "PROJECT")]
        name: String,
        
        #[arg(short, long)]
        template: Option<String>,
    },
    /// Build project
    Build {
        #[arg(short, long)]
        release: bool,
        
        #[arg(short, long)]
        target: Option<String>,
    },
    /// Deploy to environment
    Deploy {
        #[arg(value_name = "ENV")]
        environment: String,
        
        #[arg(short, long)]
        dry_run: bool,
        
        #[arg(short, long)]
        force: bool,
    },
    /// Configuration management
    Config {
        #[command(subcommand)]
        action: ConfigAction,
    },
}

#[derive(Subcommand)]
enum ConfigAction {
    /// Show current configuration
    Show,
    /// Set configuration value
    Set {
        key: String,
        value: String,
    },
    /// Validate configuration
    Validate,
}

#[derive(Debug, Deserialize, Serialize)]
struct ProjectConfig {
    name: String,
    version: String,
    targets: Vec<String>,
    environments: Vec<EnvironmentConfig>,
}

#[derive(Debug, Deserialize, Serialize)]
struct EnvironmentConfig {
    name: String,
    url: String,
    credentials: String,
}

struct ConfigManager {
    config_path: PathBuf,
    config: Option<ProjectConfig>,
}

impl ConfigManager {
    fn new(path: PathBuf) -> Result<Self> {
        Ok(Self {
            config_path: path,
            config: None,
        })
    }
    
    fn load(&mut self) -> Result<()> {
        let content = fs::read_to_string(&self.config_path)
            .context("Failed to read config file")?;
        self.config = Some(serde_yaml::from_str(&content)?);
        info!("Configuration loaded from {:?}", self.config_path);
        Ok(())
    }
    
    fn save(&self) -> Result<()> {
        if let Some(cfg) = &self.config {
            let content = serde_yaml::to_string(cfg)?;
            fs::write(&self.config_path, content)
                .context("Failed to write config file")?;
            info!("Configuration saved");
        }
        Ok(())
    }
    
    fn validate(&self) -> Result<()> {
        if let Some(cfg) = &self.config {
            if cfg.name.is_empty() {
                return Err(anyhow::anyhow!("Project name cannot be empty"));
            }
            if cfg.environments.is_empty() {
                return Err(anyhow::anyhow!("At least one environment required"));
            }
        }
        info!("Configuration validation passed");
        Ok(())
    }
}

async fn handle_init(name: String, template: Option<String>) -> Result<()> {
    info!("Initializing project: {}", name);
    
    let template = template.unwrap_or_else(|| "default".to_string());
    
    // Create project directory
    fs::create_dir_all(&name)?;
    
    // Generate default config
    let config = ProjectConfig {
        name: name.clone(),
        version: "0.1.0".to_string(),
        targets: vec!["x86_64-unknown-linux-gnu".to_string()],
        environments: vec![
            EnvironmentConfig {
                name: "development".to_string(),
                url: "http://localhost:3000".to_string(),
                credentials: "dev-creds".to_string(),
            },
            EnvironmentConfig {
                name: "production".to_string(),
                url: "https://api.example.com".to_string(),
                credentials: "prod-creds".to_string(),
            },
        ],
    };
    
    let config_content = serde_yaml::to_string(&config)?;
    fs::write(
        format!("{}/.cli-config.yaml", name),
        config_content
    )?;
    
    info!("Project initialized with template: {}", template);
    println!("Project '{}' initialized successfully!", name);
    Ok(())
}

async fn handle_build(release: bool, target: Option<String>) -> Result<()> {
    info!("Building project: release={}, target={:?}", release, target);
    
    let build_type = if release { "release" } else { "debug" };
    let target = target.unwrap_or_else(|| "x86_64-unknown-linux-gnu".to_string());
    
    // Simulate build process
    println!("Building {} binary for {}...", build_type, target);
    
    // In real implementation, call cargo/build system
    
    println!("Build completed successfully!");
    Ok(())
}

async fn handle_deploy(
    environment: String,
    dry_run: bool,
    force: bool,
) -> Result<()> {
    info!("Deploying to {}: dry_run={}, force={}", 
          environment, dry_run, force);
    
    if dry_run {
        println!("[DRY RUN] Would deploy to: {}", environment);
    } else {
        println!("Deploying to: {}", environment);
        // Actual deployment logic
    }
    
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    
    let args = Args::parse();
    
    // Set verbosity
    match args.verbose {
        0 => log::set_max_level(log::LevelFilter::Warn),
        1 => log::set_max_level(log::LevelFilter::Info),
        2 => log::set_max_level(log::LevelFilter::Debug),
        _ => log::set_max_level(log::LevelFilter::Trace),
    }
    
    let result = match args.command {
        Commands::Init { name, template } => {
            handle_init(name, template).await
        }
        Commands::Build { release, target } => {
            handle_build(release, target).await
        }
        Commands::Deploy { environment, dry_run, force } => {
            handle_deploy(environment, dry_run, force).await
        }
        Commands::Config { action } => {
            let mut cfg_mgr = ConfigManager::new(
                args.config.unwrap_or_else(|| PathBuf::from(".cli-config.yaml"))
            )?;
            cfg_mgr.load()?;
            
            match action {
                ConfigAction::Show => {
                    println!("{:#?}", cfg_mgr.config);
                }
                ConfigAction::Set { key, value } => {
                    // Implement set logic
                    println!("Setting {} = {}", key, value);
                }
                ConfigAction::Validate => {
                    cfg_mgr.validate()?;
                    println!("Configuration is valid");
                }
            }
            Ok(())
        }
    };
    
    if let Err(e) = result {
        error!("Command failed: {}", e);
        eprintln!("Error: {}", e);
        std::process::exit(1);
    }
    
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_config_creation() {
        let config = ProjectConfig {
            name: "test".to_string(),
            version: "0.1.0".to_string(),
            targets: vec!["x86_64-unknown-linux-gnu".to_string()],
            environments: vec![],
        };
        
        assert_eq!(config.name, "test");
    }
    
    #[tokio::test]
    async fn test_init_handler() {
        let result = handle_init("test_project".to_string(), None).await;
        assert!(result.is_ok());
    }
}
```

**에러 처리 및 로깅**

```go
package main

import (
	"fmt"
	"log"
	"os"
)

type CliError struct {
	Code    int
	Message string
	Details string
}

func (e *CliError) Error() string {
	return fmt.Sprintf("[ERROR %d] %s: %s", e.Code, e.Message, e.Details)
}

func exitWithError(code int, message string, details string) {
	err := &CliError{
		Code:    code,
		Message: message,
		Details: details,
	}
	log.Printf("FATAL: %v", err)
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(code)
}

// Usage examples
func validateInput(input string) error {
	if input == "" {
		return &CliError{
			Code:    1001,
			Message: "Validation Failed",
			Details: "Input cannot be empty",
		}
	}
	return nil
}
```

**핵심 아키텍처**:
- 비동기 작업 처리 (tokio/async-await)
- 설정 파일 관리 (YAML/TOML)
- 상세 로깅 및 에러 처리
- 환경 별 배포 지원
- 프로덕션 테스트 커버리지
- 명확한 exit codes

---

## Best Practices

### 1. 사용자 경험 (UX)

```
✅ DO:
- 명확한 에러 메시지 제공
  "Error: config file not found at ~/.myapp/config.yaml"
  "Hint: Run 'myapp init' to create default configuration"
  
- 진행 상태 표시
  "Building project... [████████░░] 80%"
  
- 유용한 examples/help
  "myapp run --help" shows all options with descriptions
  
- 컬러 출력 (중요한 정보 강조)
  Use ANSI colors for errors (red), warnings (yellow), success (green)

❌ DON'T:
- 불분명한 에러
  "Error: ENOENT"
  
- 진행 상황 표시 없음
  긴 작업은 사용자를 혼란스럽게 함
  
- 외계어 같은 출력
  hex dumps, 내부 stack traces
```

### 2. 개발자 경험 (DX)

```rust
// ✅ Good: Builder pattern으로 자동완성 지원
let app = Command::builder()
    .name("myapp")
    .version("1.0.0")
    .author("Team")
    .about("Description")
    .arg(Arg::new("verbose")
        .short('v')
        .long("verbose")
        .help("Enable verbose output")
        .action(ArgAction::SetTrue))
    .build();

// ❌ Bad: 문자열 하드코딩, IDE 자동완성 불가
let config = Config::from_raw_str("verbose: true\nlog_level: debug");
```

### 3. 호환성

```
✅ DO:
- 마이너 버전 업그레이드에서 호환성 유지
- 환경 변수 기본값 제공
- 레거시 옵션 표시 (deprecated 경고)

❌ DON'T:
- 메이저 버전 충돌 없이 breaking changes
- 숨겨진 의존성
```

### 4. 성능

```
CLI 시작 시간 (1초 미만 권장):
✅ < 100ms: 매우 우수 (native 바이너리)
✅ 100-500ms: 우수 (Rust, Go)
⚠️ 500ms-2s: 수용 가능 (Node.js, Python)
❌ > 2s: 사용자 경험 저하

측정:
$ time myapp --version
real    0m0.012s
```

---

## Common CLI Patterns

### 1. 멀티 서브명령어 + 전역 옵션

```go
// 구조:
// myapp [GLOBAL_OPTS] COMMAND [COMMAND_OPTS] [ARGS]

Examples:
  myapp --config custom.yaml run input.txt
  myapp -v build --release
  myapp deploy production --dry-run --force
```

### 2. 설정 파일 우선순위

```
1. 명령행 인수 (highest)
2. 환경 변수
3. 사용자 설정 파일 (~/.myapp/config.yaml)
4. 프로젝트 설정 파일(./.myapp/config.yaml)
5. 기본값 (lowest)
```

### 3. Exit codes

```
0: Success
1: General error
2: Misuse of command
64: Input data format error
65: Data file error
66: No input
67: Address not available
68: Cannot access resource
69: Temporary failure
70: Software error
71: System error
```

---

## Common Pitfalls

### 1. 숨겨진 의존성
```rust
// ❌ Bad: CLI가 특정 환경 변수에만 작동
fn main() {
    let db_url = std::env::var("DATABASE_URL")
        .expect("DATABASE_URL not set");  // 명시적 정보 없음
}

// ✅ Good: 명확한 에러 메시지 + fallback
fn main() {
    let db_url = std::env::var("DATABASE_URL")
        .unwrap_or_else(|_| {
            eprintln!("Error: DATABASE_URL environment variable not set");
            eprintln!("Please set it: export DATABASE_URL=postgres://...");
            std::process::exit(1);
        });
}
```

### 2. 과도한 로깅
```javascript
// ❌ Bad: 모든 작업을 stdout에 출력
console.log("Connecting to database...");
console.log("SELECT * FROM users...");
console.log("Processing row 1...");  // 무한 스팸

// ✅ Good: 중요한 이벤트만, 디버그는 별도
if (verbose) {
  debug("Processing row " + i);
}
log("Connected successfully");
```

### 3. 느린 시작 시간
```python
# ❌ Bad: 모든 모듈을 import (느림)
import heavy_library
import another_library
def slow_command():
    heavy_library.do_something()

# ✅ Good: 필요할 때만 import (빠름)
def slow_command():
    import heavy_library  # 지연 로딩
    heavy_library.do_something()
```

---

## November 2025 Version Stability Matrix

| Framework | Version | Release | LTS | Recommended |
|-----------|---------|---------|-----|-------------|
| Rust (Clap) | 4.5.x | Nov 2025 | ✅ | 4.5.1+ |
| Go (Cobra) | 1.8.x | Oct 2025 | ✅ | 1.8.1+ |
| Node (Commander) | 14.x | Oct 2025 | ✅ | 14.0.0+ |
| Python (Click) | 8.1.x | Oct 2024 | ✅ | 8.1.7+ |
| Python (Typer) | 0.12.x | Nov 2025 | ✅ | 0.12.0+ |

---

## Integration Examples

### With Kubernetes

```yaml
# Helm values.yaml
cli:
  image: myapp:1.0.0
  args:
    - deploy
    - production
  env:
    - name: LOG_LEVEL
      value: info
  resources:
    requests:
      memory: "64Mi"
      cpu: "250m"
```

### With GitHub Actions

```yaml
name: Deploy CLI
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          cargo build --release
          ./target/release/myapp deploy ${{ secrets.DEPLOY_ENV }} --force
```

### With Docker

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

## Security Considerations

1. **Credential Handling**
   ```rust
   // ✅ Never log credentials
   let password = read_password("Password: ")?;
   // password is NOT logged
   
   // ❌ Avoid storing plaintext
   // fs::write("credentials.txt", &password)?;  // WRONG!
   
   // ✅ Use OS keyring
   keyring::Entry::new("myapp", "token")?.set_password(&token)?;
   ```

2. **Input Validation**
   ```go
   // ✅ Always validate
   if err := validatePath(inputPath); err != nil {
       return fmt.Errorf("invalid path: %w", err)
   }
   
   // Prevent path traversal
   if strings.Contains(inputPath, "..") {
       return fmt.Errorf("path traversal not allowed")
   }
   ```

3. **Permission Checks**
   ```bash
   # Before running privileged operations
   if [[ $EUID -ne 0 ]]; then
      echo "This script must be run as root"
      exit 1
   fi
   ```

---

## Testing Strategies

### Unit Tests (Framework Examples)

**Rust Clap**:
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_cli_parsing() {
        let args = vec!["myapp", "run", "input.txt"];
        let cmd = parse_args(args);
        assert_eq!(cmd.command, "run");
    }
}
```

**Go Cobra**:
```go
func TestRun(t *testing.T) {
    cmd := rootCmd
    cmd.SetArgs([]string{"run", "file.txt"})
    err := cmd.Execute()
    assert.NoError(t, err)
}
```

### Integration Tests

```bash
#!/bin/bash
# Test actual CLI invocation
set -e

# Test 1: Help command
output=$(./myapp --help)
[[ $output == *"Usage:"* ]] || exit 1

# Test 2: Version
output=$(./myapp --version)
[[ $output == "myapp 1.0.0" ]] || exit 1

# Test 3: Invalid argument
./myapp invalid-command &>/dev/null && exit 1 || true

echo "All tests passed!"
```

---

## Deployment & Distribution

### Cross-Compilation (Rust)

```bash
# Build for multiple targets
cargo build --release --target x86_64-unknown-linux-gnu
cargo build --release --target x86_64-pc-windows-gnu
cargo build --release --target x86_64-apple-darwin

# Result: target/*/release/myapp
```

### GitHub Releases

```yaml
name: Release
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

### Package Managers

```bash
# Homebrew
brew tap myorg/myapp
brew install myapp

# Cargo
cargo install myapp

# npm
npm install -g myapp

# Go
go install github.com/myorg/myapp@latest
```

---

## References & Official Documentation

- [Rust Clap 4.5 Docs](https://docs.rs/clap/4.5/clap/)
- [Go Cobra Guide](https://cobra.dev/)
- [Node Commander](https://github.com/tj/commander.js)
- [POSIX CLI Guidelines](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html)
- [GNU CLI Guidelines](https://www.gnu.org/software/hello/manual/autoconf/Program-and-File-Names.html)
- [CLI Best Practices](https://clig.dev/)

---

**Version**: 4.0.0 (Enterprise Stable - November 2025)
**Last Updated**: 2025-11-12
**Status**: Production Ready
**Stability**: Stable
