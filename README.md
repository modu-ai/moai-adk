# 🗿 MoAI-ADK: Agentic AI Development Framework

![MoAI-ADK](./assets/images/readme/moai-adk-og.png)

**Available Languages:** [🇰🇷 한국어](./README.ko.md) | [🇺🇸 English](./README.md) | [🇯🇵 日本語](./README.ja.md) | [🇨🇳 中文](./README.zh.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: Copyleft](https://img.shields.io/badge/License-Copyleft--3.0-blue.svg)](./LICENSE)
[![Python](https://img.shields.io/badge/Python-3.11--3.14-blue)](https://www.python.org/)

> **"The purpose of vibe coding is not rapid productivity but code quality."**

MoAI-ADK provides an **AI development environment for quality code**. SPEC-First DDD (Domain-Driven Development), continuous refactoring with behavior preservation, and 20 specialized AI agents work together.

---

## 🎁 MoAI-ADK Sponsor: z.ai GLM 4.7

**💎 Optimal Solution for Cost-Effective AI Development**

MoAI-ADK partners with **z.ai GLM 4.7** to provide developers with an economical AI development environment.

### 🚀 GLM 4.7 Special Benefits

| Benefit                    | Description                                                    |
| -------------------------- | -------------------------------------------------------------- |
| **💰 70% Cost Savings**    | 1/7 the price of Claude with equivalent performance            |
| **⚡ Fast Response Speed** | Low-latency responses with optimized infrastructure            |
| **🔄 Compatibility**       | Fully compatible with Claude Code, no code modification needed |
| **📈 Unlimited Usage**     | Use freely without daily/weekly token limits                   |

### 🎁 Sign-Up Special Discount

**👉 [GLM 4.7 Sign Up (10% Additional Discount)](https://z.ai/subscribe?ic=1NDV03BGWU)**

By signing up through this link:

- ✅ **Additional 10% discount** benefit
- ✅ **Contribute to MoAI open source development** (reward credits are used for open source projects)

### 💡 Usage Guide

```bash
# 1. Get GLM API Key
Sign up at the link above and get your API key

# 2. Configure GLM in MoAI-ADK
moai glm YOUR_API_KEY
```

> **💡 Tip**: Using GLM 4.7 in a Worktree environment allows you to design with Opus and implement in bulk with GLM, saving up to 70% on costs.

---

## 🌟 Core Values

- **🎯 SPEC-First**: 90% reduction in rework with clear specifications
- **🔵 DDD (Domain-Driven Development)**: ANALYZE-PRESERVE-IMPROVE cycle with behavior preservation
- **🤖 AI Orchestration**: 20 specialized agents + 52 skills
- **🧠 Sequential Thinking MCP**: Structured problem-solving with step-by-step reasoning
- **🌐 Multilingual Routing**: Automatic support for Korean/English/Japanese/Chinese
- **🌳 Worktree Parallel Development**: Unlimited parallel work in completely isolated environments
- **🏆 MoAI Rank**: Motivation through vibe coding leaderboard
- **🔗 Ralph-Style LSP Integration (NEW v1.9.0)**: LSP-based autonomous workflow with real-time quality feedback

---

## 🎯 Ralph-Style LSP Integration (NEW v1.9.0)

### LSP Integration Overview

MoAI-ADK now supports Ralph-style autonomous workflow with LSP (Language Server Protocol) diagnostics integration. The system captures LSP diagnostic state at workflow entry, monitors diagnostic state during execution, and automatically completes phases when quality thresholds are met.

### Key Features

**LSP Baseline Capture**:
- Automatic LSP diagnostic capture at phase start
- Tracks errors, warnings, type errors, lint errors
- Baseline used for regression detection

**Completion Markers**:
- Plan Phase: SPEC created, baseline recorded
- Run Phase: 0 errors, 0 type errors, coverage >= 85%
- Sync Phase: 0 errors, <10 warnings

**Execution Modes**:
- Interactive (default): Manual approval at each step
- Autonomous (optional): Continuous loop until completion

**Loop Prevention**:
- Maximum 100 iterations
- No progress detection (5 iterations)
- Alternative strategy on stall

**Configuration**:
```yaml
# .moai/config/sections/workflow.yaml
execution_mode:
  autonomous:
    user_approval_required: false
    continuous_loop: true
    completion_marker_based: true
    lsp_feedback_integration: true
```

---

> **📚 For detailed information, see the official online documentation:** [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

## 1. 30-Second Installation

> **⚠️ Windows Users**: MoAI-ADK supports **PowerShell only**. WSL (Windows Subsystem for Linux) and Command Prompt (cmd.exe) are **not supported**. Please use PowerShell or Windows Terminal with PowerShell.

### 📋 Prerequisites

**Required Dependencies**:

- **Python 3.11+**: MoAI-ADK requires Python 3.11 or later
- **PyYAML 6.0+**: Required for hook scripts (automatically installed via `uv run --with pyyaml`)
- **uv**: Package installer for Python (included with MoAI-ADK installation)

**Note**: PyYAML is required for:
- AST-grep multi-document rule parsing
- Configuration file read/write operations
- SPEC file YAML frontmatter parsing
- Skill metadata handling

If PyYAML is not available, hooks will automatically attempt to install it using `uv run --with pyyaml`.

### 🚀 Method 1: Quick Install (Recommended)

```bash
curl -LsSf https://modu-ai.github.io/moai-adk/install.sh | sh
```

### 🔧 Method 2: Manual Install

```bash
# Install MoAI-ADK
uv tool install moai-adk
```

### 🎨 Interactive Setup Wizard

Running `moai init` starts a **9-step interactive wizard**:

![MoAI-ADK Init Wizard](./assets/images/readme/init-wizard-banner.png)

---

#### Step 1: Language Selection

Select your conversation language. All subsequent instructions will be displayed in your chosen language.

```text
🌐 Language Selection
❯ Select your conversation language: [↑↓] Navigate  [Enter] Select
❯ Korean (한국어)
  English
  Japanese (日本語)
  Chinese (中文)
```

---

#### Step 2: Name Input

Enter your user name. AI will provide personalized responses.

```text
👤 User Settings
❯ Enter your user name (optional):
```

---

#### Step 3: GLM API Key Input

Enter your GLM API key from Z.AI.

```text
🔑 API Key Input
GLM CodePlan API key (optional - press Enter to skip)

✓ Found existing GLM API key: 99c1a2df...
Press Enter to keep existing key, or enter new key to replace

? Enter your GLM API key:
```

> 🎁 **GLM Sign-Up Benefit**: If you don't have a GLM account, sign up using the link below!
>
> **👉 [GLM Sign Up (10% Additional Discount)](https://z.ai/subscribe?ic=1NDV03BGWU)**
>
> By signing up through this link, you'll receive an **additional 10% discount**.
> Also, rewards generated from link sign-ups are used for **MoAI open source development**. 🙏

---

#### Step 4: Project Setup

Enter your project name.

```text
📁 Project Setup
❯ Project name: MoAI-ADK
```

---

#### Step 5: Git Setup

Select your Git mode.

```text
🔀 Git Setup
❯ Select your Git mode: [↑↓] Navigate  [Enter] Select
❯ manual (local only) - Use local repository only
  personal (GitHub personal) - Use GitHub personal account
  team (GitHub team) - Use GitHub team/organization
```

---

#### Step 6: GitHub Username

For personal/team selection, enter your GitHub username.

```text
❯ GitHub username:
```

---

#### Step 7: Commit Message Language

Select the language for Git commit messages.

```text
🗣️ Output Language Settings
❯ Commit message language: [↑↓] Navigate  [Enter] Select
  English
❯ Korean (한국어)
  Japanese (日本語)
  Chinese (中文)
```

---

#### Step 8: Code Comment Language

Select the language for code comments.

```text
❯ Code comment language: [↑↓] Navigate  [Enter] Select
  English
❯ Korean (한국어)
  Japanese (日本語)
  Chinese (中文)
```

---

#### Step 9: Documentation Language

Select the language for documentation.

```text
❯ Documentation language: [↑↓] Navigate  [Enter] Select
  English
❯ Korean (한국어)
  Japanese (日本語)
  Chinese (中文)
```

> 💡 **Token Optimization Strategy**: Internal prompts for agent instructions are **fixed in English**.
>
> **Reason**: Non-English languages consume **12-20% additional tokens** in Claude. When infinite agent loops are frequent, this significantly impacts costs and weekly token limits, so MoAI fixes internal agent instructions in English and **provides user conversations only in the user's language**.
>
> This is MoAI's **effort to reduce token waste**.

---

#### Installation Complete

When all settings are complete, 5-phase installation runs automatically:

```text
🚀 Starting installation...

Phase 1: Preparation and backup...        ████████████████ 100%
Phase 2: Creating directory structure...  ████████████████ 100%
Phase 3: Installing resources...          ████████████████ 100%
Phase 4: Generating configurations...     ████████████████ 100%
Phase 5: Validation and finalization...   ████████████████ 100%

✅ Initialization Completed Successfully!
────────────────────────────────────────────────────────────────

📊 Summary:
  📁 Location:   /path/to/my-project
  🌐 Language:   Auto-detect (use /moai project)
  🔀 Git:        manual (github-flow, branch: manual)
  🌍 Locale:     ko
  📄 Files:      47 created
  ⏱️  Duration:   1234ms

🚀 Next Steps:
  1. Run cd my-project to enter the project
  2. Run /moai project in Claude Code for full setup
  3. Start developing with MoAI-ADK!
```

### Add to Existing Project

```bash
cd your-existing-project
moai init .
# Existing files are preserved
```

---

### 🔄 MoAI-ADK Update

Update existing projects to the latest version.

```bash
moai update
```

**3-Phase Smart Update Workflow**:

```text
Stage 1: 📦 Package Version Check
         └─ Check latest version from PyPI → Auto-upgrade if needed

Stage 2: 🔍 Config Version Comparison
         └─ Compare package template vs project config
         └─ Skip if same (70-80% performance improvement)

Stage 3: 📄 Template Synchronization
         └─ Create backup → Update template → Restore user settings
```

**Key Options**:

```bash
# Check version only (no update)
moai update --check

# Synchronize templates only (skip package upgrade)
moai update --templates-only

# Config edit mode (re-run init wizard)
moai update --config
moai update -c

# Force update without backup
moai update --force

# All is well~ Auto mode (auto-approve all confirmations)
moai update --yes
```

**Merge Strategy Selection**:

```text
🔀 Choose merge strategy:
  [1] Auto-merge (default)
      → Automatically preserve template + user changes
  [2] Manual merge
      → Create backup + merge guide (direct control)
```

```bash
# Force auto-merge (default)
moai update --merge

# Force manual merge
moai update --manual
```

**Automatically Preserved Items**:

| Item                | Description                                       |
| ------------------- | ------------------------------------------------- |
| **User Settings**   | `.claude/settings.local.json` (MCP, GLM settings) |
| **Custom Agents**   | User-created agents not in template               |
| **Custom Commands** | User-defined slash commands                       |
| **Custom Skills**   | User-defined skills                               |
| **Custom Hooks**    | User-defined hook scripts                         |
| **SPEC Documents**  | Entire `.moai/specs/` folder                      |
| **Reports**         | Entire `.moai/reports/` folder                    |

> 💡 **Update Tip**: You can change language, API key, and Git settings anytime with `moai update -c`.
> It's recommended to create your commands, agents, skills, and hooks in folders other than moai.

---

## ⚠️ Known Issues & Solutions

### pip and uv tool Conflict

**Problem**: When both pip and uv tool are used to install MoAI-ADK, version conflicts can occur.

**Symptoms**:
```bash
# moai update shows latest version
moai update
✓ Package already up to date (1.5.0)

# But actual command uses old version
which moai
~/.pyenv/shims/moai  # Uses pip version (e.g., 1.1.0)

# Import errors in hooks
ModuleNotFoundError: No module named 'yaml'
```

**Root Cause**:
- `uv tool install` installs to `~/.local/bin/moai`
- `pip install` installs to `~/.pyenv/shims/moai`
- PATH priority determines which version is used
- Windows users may experience more severe issues due to Python environment differences

**Solutions**:

#### Option 1: Clean reinstall with uv tool (Recommended)

```bash
# Uninstall any existing installation
pip uninstall moai-adk -y
uv tool uninstall moai-adk

# Reinstall with uv tool
uv tool install moai-adk

# Verify
moai --version  # Should show latest version
```

#### Option 2: Update shell configuration

**For macOS/Linux (~/.zshrc or ~/.bashrc)**:
```bash
# Add to PATH
export PATH="$HOME/.local/bin:$PATH"
```

**For Windows (PowerShell $PROFILE)**:
```powershell
# Add to $PROFILE
$env:PATH = "$env:USERPROFILE\.local\bin;$env:PATH"
```

**Prevention**:
- Always use `uv tool install moai-adk` for installation
- Regularly check `which moai` to verify active installation

---

### WSL (Windows Subsystem for Linux) Support

MoAI-ADK fully supports **WSL 1** and **WSL 2** on Windows 10 and Windows 11.

#### Installation on WSL

```bash
# Install MoAI-ADK in WSL
uv tool install moai-adk

# Initialize your project
cd your-project-directory
moai-adk init
```

#### Path Handling

MoAI-ADK automatically converts between Windows and WSL path formats:

- **Windows paths**: `C:\Users\goos\project` → **WSL paths**: `/mnt/c/Users/goos/project`
- No manual configuration required
- Works seamlessly with projects in both Linux filesystem (`/home/user/`) and Windows filesystem (`/mnt/c/`)

#### Best Practices

**Recommended**: Place projects in Linux filesystem for optimal performance
```bash
# ✅ Best performance
cd ~/projects
moai-adk init
```

**Also Supported**: Projects on Windows filesystem
```bash
# ✅ Works but may have slight overhead
cd /mnt/c/Users/YourName/projects
moai-adk init
```

#### Troubleshooting WSL

**Verify WSL Environment:**
```bash
# Check if running in WSL
echo $WSL_DISTRO_NAME

# Verify CLAUDE_PROJECT_DIR (set by Claude Code)
echo $CLAUDE_PROJECT_DIR
```

**PATH Not Loaded (~/.local/bin not in PATH):**

This is the most common issue in WSL. Even if you added PATH to `~/.bashrc`, it may not be loaded because WSL uses **login shell** which reads `~/.profile`, not `~/.bashrc`.

**Symptoms:**
```bash
# Status line shows:
"Native installation exists but ~/.local/bin is not in your PATH"

# MCP servers fail to start
# Commands like 'moai' or 'uv' not found
```

**Diagnosis:**
```bash
# Run the diagnostic command
moai-adk doctor --shell

# Or check manually
echo $PATH | grep -o "$HOME/.local/bin"
```

**Solutions:**

Option A: Use automatic fix (Recommended)
```bash
# Run with --fix flag
moai-adk doctor --shell --fix
```

Option B: Manual fix for bash users
```bash
# Add to ~/.profile (NOT ~/.bashrc)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.profile
source ~/.profile
```

Option C: Manual fix for zsh users
```bash
# Add to ~/.zshenv (always sourced)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshenv
source ~/.zshenv
```

**Why ~/.profile instead of ~/.bashrc?**
- WSL starts as a login shell
- Login shell reads: `~/.profile` → `~/.bashrc` (if sourced)
- By default, `~/.profile` doesn't source `~/.bashrc` for PATH
- Adding PATH directly to `~/.profile` ensures it's always loaded

**Hook/MCP Server Issues:**
- If hooks fail, verify `CLAUDE_PROJECT_DIR` is set correctly
- MoAI-ADK automatically converts Windows paths to WSL format
- Check `.claude/settings.json` for correct path references

**Related Issues:**
- [Issue #295: WSL Support Request](https://github.com/modu-ai/moai-adk/issues/295)
- [Claude Code Issue #19653: WSL Path Handling](https://github.com/anthropics/claude-code/issues/19653)

---

## 2. Project Documentation Generation (Optional)

For new or existing projects, you can automatically generate **project documentation to help Claude Code understand your project**:

```text
> /moai project
```

### 3 Generated Files

| File                         | Purpose                | Key Content                                                                                |
| ---------------------------- | ---------------------- | ------------------------------------------------------------------------------------------ |
| `.moai/project/product.md`   | **Product Overview**   | Project name/description, target users, core features, use cases                           |
| `.moai/project/structure.md` | **Structure Analysis** | Directory tree, main folder purposes, core file locations, module composition              |
| `.moai/project/tech.md`      | **Tech Stack**         | Technologies used, framework selection reasons, dev environment, build/deployment settings |

### Why Do You Need This?

- **Context Provision**: Claude Code quickly grasps project context
- **Consistency Maintenance**: Share project understanding among team members
- **Onboarding Acceleration**: Reduce new developer project ramp-up time
- **AI Collaboration Optimization**: More accurate code suggestions and reviews

> 💡 **Tip**: Running `/moai project` at project start or after structure changes updates documentation to the latest state.

---

## 3. Core Commands Collection

### 🎯 `/moai project` - Project Initialization

```bash
> /moai project
```

Automatically analyzes your project's current state and configures the optimal development environment. Detects programming languages and frameworks, and automatically sets up Git workflows and quality assurance standards. Once all configuration is complete, you're ready to start development immediately.

**Tasks**:

- ✅ Project structure analysis
- ✅ Programming language/framework detection
- ✅ `.moai/config/config.yaml` creation
- ✅ Git workflow setup
- ✅ Session memory system configuration
- ✅ Quality assurance standards setup

---

### 📋 `/moai plan` - SPEC Writing

```bash
> /moai plan "feature description"
```

Automatically generates unambiguous specifications using EARS format. Includes requirements definition, success criteria, and test scenarios to clearly present development direction. The generated SPEC serves as a single source of truth that the development team and AI share the same understanding.

**Auto-generated**:

- EARS format specification
- Requirements definition
- Success criteria
- Test scenarios

**Example**:

```bash
> /moai plan "user profile page"
# → Creates SPEC-002

> /moai plan "payment API"
# → Creates SPEC-003
```

**Important**: Must execute `> /clear` next

---

### 💻 `/moai run` - Implementation (DDD)

```bash
> /moai run SPEC-001
```

**NEW v1.9.0**: LSP-enhanced DDD cycle with completion markers

Implements SPEC using DDD (Domain-Driven Development) methodology with the ANALYZE-PRESERVE-IMPROVE cycle:

**DDD Cycle**:

- 🔍 **ANALYZE**: Analyze domain boundaries, coupling, and existing behavior
- 🛡️ **PRESERVE**: Create characterization tests to preserve existing behavior
- ✨ **IMPROVE**: Incrementally improve structure with test validation

**LSP-Enhanced Quality Gates** (NEW v1.9.0):
- LSP baseline capture at phase start
- Real-time regression detection
- Automatic completion at 0 errors, 0 type errors, coverage >= 85%
- Completion marker detection: `<moai>DONE</moai>`

**Verification items**:

- Test coverage >= 85%
- Pass linting
- Pass type checking
- Pass security checks
- ✅ TRUST 5 validation

---

### 📚 `/moai sync` - Documentation Sync

```bash
> /clear  # Always run clear before sync to reset session and run quality checks
> /moai sync SPEC-001
```

Performs quality verification followed by documentation synchronization, Git commits, and PR automation. Automatically generates API documentation, architecture diagrams, README, and CHANGELOG, keeping them up to date. Automatically commits changes and converts PRs from Draft to Ready in team mode.

**Auto-executed tasks**:

1. **Phase 1: Quality Verification**
   - Run tests (pytest, jest, go test, etc.)
   - Linter checks (ruff, eslint, golangci-lint, etc.)
   - Type checkers (mypy, tsc, go vet, etc.)
   - Code review (manager-quality)

2. **Phase 2-3: Documentation Sync**
   - Auto-generate API documentation
   - Update architecture diagrams
   - Update README
   - Sync SPEC documents

3. **Phase 4: Git Automation**
   - Commit changes
   - PR Draft → Ready conversion
   - (Optional) Auto-merge

**Execution modes**:

- `auto` (default): Selective sync of changed files only
- `force`: Regenerate all documentation
- `status`: Perform status check only
- `project`: Sync entire project

**Details**: Refer to command file

---

### 🚀 `/moai` - Fully Autonomous Automation

```bash
> /moai "feature description"
```

**NEW v1.9.0**: LSP integration with autonomous mode

User presents the goal and AI autonomously performs exploration, planning, implementation, and verification. Analyzes codebase through parallel exploration and self-corrects issues through autonomous loops. Automatically terminates when completion marker (`<moai>DONE</moai>`) is detected, so developer only needs to verify final result.

#### LSP-Enhanced Autonomous Mode (NEW v1.9.0)

**Autonomous Mode**:
- LSP baseline capture at start
- Real-time regression detection
- Automatic completion when markers met
- Loop prevention safeguards

**Interactive Mode** (default):
- Manual approval at each phase
- Backward compatible with existing workflows

#### Concept and Workflow

```mermaid
flowchart TB
    Start([User Request<br/>/moai 'feature description']) --> Phase0[Phase 0: Parallel Exploration]

    Phase0 --> Explore[🔍 Explore Agent<br/>Analyze codebase structure]
    Phase0 --> Research[📚 Research Agent<br/>Research technical docs]
    Phase0 --> Quality[✅ Quality Agent<br/>Assess quality state]

    Explore --> Phase1[Phase 1: SPEC Creation]
    Research --> Phase1
    Quality --> Phase1

    Phase1 --> Spec[📋 EARS format SPEC doc<br/>Requirements specification]

    Spec --> Phase2[Phase 2: DDD Implementation]

    Phase2 --> Analyze[🔍 ANALYZE: Domain analysis]
    Analyze --> Preserve[🛡️ PRESERVE: Characterization tests]
    Preserve --> Improve[✨ IMPROVE: Incremental improvement]

    Improve --> Check{Quality Check<br/>TRUST 5}

    Check -->|Pass| Phase3[Phase 3: Documentation Sync]
    Check -->|Fail| Loop[🔄 Autonomous Loop<br/>Auto-fix issues]

    Loop --> Analyze

    Phase3 --> Docs[📚 README, API docs<br/>Auto-update]

    Docs --> Done{<moai>DONE</moai><br/>Completion marker detected}

    Done -->|Yes| End([✅ Complete<br/>Deliver result only])
    Done -->|No| Loop

    style Phase0 fill:#e1f5fe
    style Phase1 fill:#fff3e0
    style Phase2 fill:#f3e5f5
    style Phase3 fill:#e8f5e9
    style Done fill:#c8e6c9
    style End fill:#4caf50,color:#fff
```

#### Detailed Process

**Single Execution**:

1. **Phase 1: Parallel Exploration** (3-4x faster analysis)
   - **Explore Agent**: Codebase structure, patterns, related files
   - **Research Agent**: Technical docs, best practices
   - **Quality Agent**: Current quality state, potential issues

2. **Phase 2: SPEC Creation** (EARS format)
   - Clear requirement definitions
   - Acceptance criteria specification
   - User story writing

3. **Phase 3: DDD Implementation** (Autonomous loop)
   - **ANALYZE**: Analyze domain boundaries and existing behavior
   - **PRESERVE**: Create characterization tests for behavior preservation
   - **IMPROVE**: Incrementally improve code structure
   - **Loop**: Auto-fix if quality check fails

4. **Phase 4: Documentation Sync**
   - README, API docs auto-update
   - CHANGELOG auto-generation
   - User guide update

#### When to use?

| Situation                   | Description                                | Example                         |
| --------------------------- | ------------------------------------------ | ------------------------------- |
| **New feature development** | AI handles everything from start to finish | "Add JWT authentication system" |
| **Complex refactoring**     | Large changes affecting multiple files     | "Database layer restructure"    |
| **Bug fixing**              | Automate from root cause analysis to fix   | "Fix login failure bug"         |
| **SPEC-based development**  | Implement features with SPEC documents     | `/moai SPEC-AUTH-001`           |

**Options**:

- `--loop`: Enable autonomous iterative fixes (AI self-resolves issues)
- `--max N`: Specify max iterations (default: 100)
- `--sequential` / `--seq`: Sequential exploration (for debugging) - parallel is default
- `--branch`: Auto-create feature branch
- `--pr`: Create Pull Request after completion
- `--resume SPEC`: Resume work

> **Performance**: Parallel exploration is now the default, providing 3-4x faster analysis. Use `--sequential` only for debugging.

**Example**:

```bash
# Basic autonomous execution (parallel by default)
> /moai "Add JWT authentication"

# Auto loop with sequential exploration (for debugging)
> /moai "JWT authentication" --loop --seq

# Resume
> /moai resume SPEC-AUTH-001

# UltraThink mode (deep analysis with Sequential Thinking)
> /moai "Add JWT authentication" --ultrathink
```

**UltraThink Mode** (`--ultrathink`): Enhanced analysis mode that automatically applies Sequential Thinking MCP to deeply analyze requests and generate optimal execution plans.

When `--ultrathink` is appended, MoAI activates structured reasoning to:
- Break down complex problems into manageable steps
- Map each subtask to appropriate agents
- Identify parallel vs sequential execution opportunities
- Generate optimal agent delegation strategy

**Example Output with UltraThink**:
```
thought: "Analyzing request: User wants JWT authentication. This involves: user model (backend), API endpoints (backend), login form (frontend), auth context (frontend), tests (testing)."

thought: "Subtask decomposition: 1) User model → expert-backend, 2) JWT API → expert-backend, 3) Login form → expert-frontend, 4) Auth context → expert-frontend, 5) Tests → expert-testing."

thought: "Execution strategy: Phase 1 - Launch expert-backend in parallel for user model + API. Phase 2 - expert-frontend for UI. Phase 3 - expert-testing for tests."

thought: "Final plan: Use the expert-backend subagent (parallel), then Use the expert-frontend subagent, then Use the expert-testing subagent."
```

Use `--ultrathink` for:
- Complex multi-domain tasks (backend + frontend + testing)
- Architecture decisions affecting multiple files
- Performance optimization requiring analysis
- Security review needs
- Refactoring with behavior preservation

---

### 🔁 `/moai loop` - Autonomous Iterative Fixing

```bash
> /moai loop
```

AI autonomously diagnoses and fixes LSP errors, test failures, and coverage deficiencies. Parallel diagnosis executes LSP, AST-grep, Tests, and Coverage simultaneously, resolving issues 3-4x faster. Runs autonomously until completion marker is detected or max iterations reached.

#### Concept and Workflow

```mermaid
flowchart TB
    Start([User Request<br/>/moai loop]) --> Parallel[Parallel Diagnosis]

    Parallel --> LSP[LSP Diagnosis<br/>Type errors, not found]
    Parallel --> AST[AST-grep<br/>Pattern check, security]
    Parallel --> Tests[Test Execution<br/>Unit, integration]
    Parallel --> Coverage[Coverage<br/>85% target]

    LSP --> Collect[Issue Collection]
    AST --> Collect
    Tests --> Collect
    Coverage --> Collect

    Collect --> HasIssues{Has Issues?}

    HasIssues -->|No| Done[<moai>DONE</moai><br/>Completion marker]
    HasIssues -->|Yes| CreateTODO[TODO Creation<br/>Priority sorted]

    CreateTODO --> Process[Sequential Processing]

    Process --> Fix1[Level 1: Immediate fix<br/>Import sort, whitespace]
    Process --> Fix2[Level 2: Safe fix<br/>Variable name, type]
    Process --> Fix3[Level 3: Approval fix<br/>Logic change]

    Fix1 --> Validate[Verification]
    Fix2 --> Validate
    Fix3 --> Validate

    Validate --> ReCheck{Diagnose again?}

    ReCheck -->|Yes| Parallel
    ReCheck -->|No| MaxIter{Max iteration<br/>100 reached?}

    MaxIter -->|No| Parallel
    MaxIter -->|Yes| Snapshot[Snapshot saved<br/>Can resume later]

    Done --> End([✅ Complete])
    Snapshot --> End([⏸️ Paused])

    style Parallel fill:#e1f5fe
    style Collect fill:#fff3e0
    style Process fill:#f3e5f5
    style Validate fill:#e8f5e9
    style Done fill:#c8e6c9
    style End fill:#4caf50,color:#fff
```

#### Parallel Diagnosis Details

**Parallel diagnosis** (3.75x faster):

```mermaid
flowchart TB
    Start([Parallel diagnosis start]) --> Parallel

    subgraph Parallel[Concurrent execution]
        direction TB
        LSP[LSP diagnosis]
        AST[AST-grep inspection]
        TESTS[Test execution]
        COVERAGE[Coverage check]
    end

    LSP --> Collect[Issue integration and priority<br/>Level 1 → 2 → 3 order]
    AST --> Collect
    TESTS --> Collect
    COVERAGE --> Collect

    style Start fill:#e3f2fd
    style Parallel fill:#f3f4f6
    style LSP fill:#fff9c4
    style AST fill:#ffccbc
    style TESTS fill:#c8e6c9
    style COVERAGE fill:#b2dfdb
    style Collect fill:#e1bee7
```

#### 📖 What is AST-grep?

> **"grep finds text, but AST-grep finds code structure."**

**Concept**:

AST-grep is a **structural code inspection tool**. Unlike normal grep or regex that searches for text, AST-grep analyzes the **Abstract Syntax Tree (AST)** of code to inspect code **structure and patterns**.

**Text search vs Structure search**:

| Feature              | grep/regex                              | AST-grep                                    |
| -------------------- | --------------------------------------- | ------------------------------------------- |
| Search target        | Text strings                            | Code structure (AST)                        |
| Example              | `print("hello")`                        | `print(__)`                                 |
| Meaning              | Find "print" characters                 | Find print function call pattern            |
| Space sensitive?     | Yes (whitespace, indent matters)        | No (structure only)                         |
| Variable distinguish | Hard (e.g., `x=1`, `y=1` are different) | Possible (all variable assignment patterns) |

**How it works**:

```mermaid
flowchart LR
    Source[Source code<br/>def foo x:<br/>    return x + 1] --> AST[AST analysis]

    AST --> |transform| Tree[Abstract syntax tree<br/>Function<br/> Call<br/>]

    Tree --> Pattern[Pattern matching]

    Pattern --> Result1[✓ Function definition]
    Pattern --> Result2[✓ Return statement]
    Pattern --> Result3[✓ Addition operation]

    style Source fill:#e3f2fd
    style AST fill:#fff3e0
    style Tree fill:#f3e5f5
    style Pattern fill:#e8f5e9
    style Result1 fill:#c8e6c9
    style Result2 fill:#c8e6c9
    style Result3 fill:#c8e6c9
```

**What AST-grep detects**:

1. **Security vulnerabilities**
   - SQL injection patterns: `execute(f"SELECT * FROM users WHERE id={user_input}")`
   - Hardcoded passwords: `password = "123456"`
   - Unsafe function usage: `eval(user_input)`

2. **Code smells**
   - Duplicate code: Repetition of similar structures
   - Long functions: Too much complexity
   - Magic numbers: `if x == 42` (meaningless numbers)

3. **Anti-patterns**
   - Empty except blocks: `except: pass`
   - Global variable modification
   - Circular dependencies

4. **Best practice violations**
   - Missing type hints
   - Missing documentation
   - Missing error handling

**Example scenario**:

```python
# Example code where AST-grep finds problems
def process_user_input(user_input):
    # ⚠️ Warning: eval usage (security vulnerability)
    result = eval(user_input)

    # ⚠️ Warning: empty except (anti-pattern)
    try:
        save_to_database(result)
    except:
        pass

    # ⚠️ Warning: magic number (code smell)
    if result > 42:
        return True
```

**Why is it important?**

- **Accuracy**: Understands code meaning, so fewer false positives
- **40 language support**: Works on Python, TypeScript, Go, Rust, Java, etc.
- **Auto-fix possible**: Not just finds patterns, but generates fix suggestions
- **Security enhancement**: Automatically detects security vulnerabilities like OWASP Top 10

**Usage in MoAI-ADK**:

In `/moai loop` and `/moai fix` commands, AST-grep operates as a core component of parallel diagnosis:

- **LSP**: Type errors, find definitions
- **AST-grep**: Structural patterns, security vulnerabilities ← **This is our focus!**
- **Tests**: Test failures
- **Coverage**: Coverage deficiency

These four run **simultaneously** to diagnose code quality 3.75x faster.

---

#### Detailed Process

**Autonomous loop flow**:

1. **Parallel diagnosis** (concurrent execution)
   - **LSP**: Type errors, definitions not found, potential bugs
   - **AST-grep**: Code pattern check, security vulnerabilities
   - **Tests**: Unit tests, integration tests
   - **Coverage**: 85% coverage target achievement

2. **TODO creation** (by priority)
   - Level 1: Immediate fix (import sort, whitespace, formatting)
   - Level 2: Safe fix (variable names, type addition)
   - Level 3: Approval fix (logic change, API modification)
   - Level 4: Manual needed (security, architecture)

3. **Sequential fixing**
   - Process TODO items one by one
   - Verify after each fix
   - Re-diagnose on failure

4. **Repeat or complete**
   - `<moai>DONE</moai>` marker when all issues resolved
   - Save snapshot after max 100 iterations

#### When to use?

| Situation                        | Description                                 | Example                                       |
| -------------------------------- | ------------------------------------------- | --------------------------------------------- |
| **Quality after implementation** | Automatically improve quality after coding  | Run `/moai loop` after feature implementation |
| **Test failure fix**             | Automatically analyze and fix test failures | Run after test execution on failure           |
| **Coverage improvement**         | Automatically achieve 85% target            | After writing new code                        |
| **Refactoring**                  | Continuously improve code quality           | Periodic execution for maintenance            |

**Options**:

- `--max N`: Maximum iterations (default: 100)
- `--auto`: Enable auto-fix (Level 1-3)
- `--sequential` / `--seq`: Sequential diagnosis (for debugging) - parallel is default
- `--errors`: Fix errors only
- `--coverage`: Include coverage (100% target)
- `--resume ID`: Restore snapshot

> **Performance**: Parallel diagnosis is now the default, executing LSP, AST-grep, Tests, and Coverage simultaneously (3.75x faster).

**Example**:

```bash
# Basic autonomous loop (parallel by default)
> /moai loop

# Sequential + auto-fix (for debugging)
> /moai loop --seq --auto

# Max 50 iterations
> /moai loop --max 50

# Restore snapshot
> /moai loop --resume latest
```

---

### 🔧 `/moai fix` - One-Shot Auto Fix

```bash
> /moai fix
```

Scans LSP errors and linting issues in parallel and fixes them at once. Level 1-2 fixes immediately, Level 3 fixes after user approval, Level 4 reports as requiring manual fix. Use `--dry` option to preview before applying actual fixes.

#### Concept and Workflow

```mermaid
flowchart
    Start([User Request<br/>/moai fix]) --> Dry{--dry mode?}

    Dry -->|Yes| ScanOnly[Scan only<br/>No fixes]
    Dry -->|No| Scan[Parallel scan]

    Scan --> LSP[LSP diagnosis<br/>Type, definition errors]
    Scan --> AST[AST-grep<br/>Patterns, security]
    Scan --> Linter[Linter<br/>Style, format]

    LSP --> Collect[Issue collection]
    AST --> Collect
    Linter --> Collect

    ScanOnly --> Report[Report output<br/>Expected fix content]

    Collect --> Categorize[Level classification]

    Categorize --> L1[Level 1<br/>Immediate fix]
    Categorize --> L2[Level 2<br/>Safe fix]
    Categorize --> L3[Level 3<br/>Approval needed]
    Categorize --> L4[Level 4<br/>Manual needed]

    L1 --> AutoFix1[Auto fix]
    L2 --> AutoFix2[Auto fix+log]
    L3 --> Approve{User approval?}

    Approve -->|Yes| AutoFix3[Execute fix]
    Approve -->|No| Skip[Skip]

    L4 --> Manual[Report manual fix needed]

    AutoFix1 --> Validate[Verification]
    AutoFix2 --> Validate
    AutoFix3 --> Validate
    Skip --> Validate

    Validate --> Results[Result summary]

    Manual --> Results
    Report --> End([✅ Complete])
    Results --> End

    style Scan fill:#e1f5fe
    style Collect fill:#fff3e0
    style AutoFix1 fill:#c8e6c9
    style AutoFix2 fill:#c8e6c9
    style AutoFix3 fill:#fff9c4
    style Manual fill:#ffccbc
    style End fill:#4caf50,color:#fff
```

#### Parallel Scan Details

**Parallel scan** (3.75x faster):

```mermaid
flowchart TB
    Start([Parallel scan start]) --> Parallel

    subgraph Parallel[Concurrent execution]
        direction TB
        LSP[LSP diagnosis]
        AST[AST-grep inspection]
        Linter[Linter inspection]
    end

    LSP --> Collect[Issue integration and level classification<br/>Level 1 → 2 → 3 → 4]
    AST --> Collect
    Linter --> Collect

    style Start fill:#e3f2fd
    style Parallel fill:#f3f4f6
    style LSP fill:#fff9c4
    style AST fill:#ffccbc
    style Linter fill:#c8e6c9
    style Collect fill:#e1bee7
```

#### Detailed Process

**Fix level processing**:

| Level | Description     | Risk   | Approval   | Auto | Example                         |
| ----- | --------------- | ------ | ---------- | ---- | ------------------------------- |
| 1     | Immediate fix   | Low    | Not needed | ✅   | import sorting, whitespace      |
| 2     | Safe fix        | Low    | Log only   | ✅   | variable names, type addition   |
| 3     | Approval needed | Medium | Needed     | ⚠️   | logic changes, API modification |
| 4     | Manual needed   | High   | Impossible | ❌   | security, architecture          |

**One-time execution process**:

1. **Parallel scan** (concurrent execution)
   - **LSP**: Type errors, definitions not found, serious bugs
   - **AST-grep**: Code patterns, security vulnerabilities
   - **Linter**: Style, formatting, naming

2. **Issue collection and classification**
   - Classify all issues by level
   - Priority: Level 1 → 2 → 3 → 4

3. **Batch fix** (execute at once)
   - Level 1-2: Auto-fix immediately
   - Level 3: Fix after user approval
   - Level 4: Report manual fix needed

4. **Verification and results**
   - Run tests after fixes
   - Report summary of changes

#### When to use?

| Situation                 | Description                            | Example                      |
| ------------------------- | -------------------------------------- | ---------------------------- |
| **Clean up after coding** | Clean up style, formatting all at once | Run `/moai fix` after coding |
| **Pre-commit inspection** | Resolve LSP errors, linting in advance | Run before `git commit`      |
| **Quick fix**             | Solve at once without repetitive fixes | When fixing simple issues    |
| **Preview confirmation**  | Check fix content before applying      | Run `/moai fix --dry`        |

#### `/moai fix` vs `/moai loop` Selection Guide

```mermaid
flowchart
    Start([Code quality issue occurred]) --> Question{Need repetitive<br/>fixing?}

    Question -->|Yes| Loop[🔄 /moai loop<br/>Autonomous iterative fixing]
    Question -->|No| Fix{Can solve<br/>at once?}

    Fix -->|Yes| FixCmd[🔧 /moai fix<br/>One-shot auto fix]
    Fix -->|No| Loop

    Loop --> LoopDesc[Complex issues<br/>Multiple files affected<br/>Repeated test failures]
    FixCmd --> FixDesc[Simple issues<br/>Style, format<br/>LSP errors]

    LoopDesc --> When1[Usage timing:<br/>Quality after implementation]
    FixDesc --> When2[Usage timing:<br/>Quick clean up before commit]

    style Loop fill:#f3e5f5
    style FixCmd fill:#e8f5e9
    style LoopDesc fill:#fff3e0
    style FixDesc fill:#e1f5fe
```

**Fix levels**:

| Level | Description     | Approval   | Example                         |
| ----- | --------------- | ---------- | ------------------------------- |
| 1     | Immediate fix   | Not needed | import sorting, whitespace      |
| 2     | Safe fix        | Log only   | variable names, type addition   |
| 3     | Approval needed | Needed     | logic changes, API modification |
| 4     | Manual needed   | Impossible | security, architecture          |

**Options**:

- `--dry`: Preview only (no actual fixes)
- `--sequential` / `--seq`: Sequential scan (for debugging) - parallel is default
- `--level N`: Max fix level (default: 3)
- `--errors`: Fix errors only
- `--security`: Include security check
- `--no-fmt`: Skip formatting

> **Performance**: Parallel scan is now the default, combining LSP, AST-grep, and Linter results simultaneously (3.75x faster).

**Example**:

```bash
# Basic fix (parallel by default)
> /moai fix

# Sequential scan (for debugging)
> /moai fix --seq

# Preview
> /moai fix --dry

# Specific files
> /moai fix src/auth.py
```

---

## 4. 🤖 All is Well - Agentic Autonomous Automation

**The most powerful feature of MoAI-ADK**: AI autonomously explores, plans, implements, and verifies until the completion marker is detected.

### Core Concept

```text
User: "Add authentication feature"
  ↓
AI: Explore → Plan → Implement → Verify → Repeat
  ↓
AI: All issues resolved
  ↓
AI: <moai>DONE</moai>  ← Completion Marker
```

### Three-Command Hierarchy

MoAI-ADK provides three levels of autonomous automation:

| Command      | Scope           | Iteration        | Purpose                        |
| ------------ | --------------- | ---------------- | ------------------------------ |
| `/moai fix`  | Code fixes only | 1 time           | Single scan + auto-fix         |
| `/moai loop` | Code fixes      | Until marker/max | Autonomous iterative fixing    |
| `/moai`      | Full dev cycle  | Until marker/max | Goal → SPEC → Implement → Docs |

### Command Chain Relationship

```text
/moai (Full Automation)
     │
     ├── Phase 0: Parallel Exploration
     ├── Phase 1: SPEC Generation
     ├── Phase 2: Implementation ←─┬── /moai loop (Iterative)
     │                             │        │
     │                             │        └── /moai fix (Single)
     └── Phase 3: Documentation
```

### Completion Markers

AI uses completion markers to signal task completion:

```markdown
## Task Complete

All implementations done, tests passing, docs updated. <moai>DONE</moai>
```

**Supported Markers**:

- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion
- `<moai:done />` - XML format

### Auto-Fix Levels

| Level | Description     | Approval     | Examples                         |
| ----- | --------------- | ------------ | -------------------------------- |
| 1     | Immediate fix   | Not required | import sorting, whitespace       |
| 2     | Safe fix        | Log only     | variable rename, type hints      |
| 3     | Approval needed | Required     | logic changes, API modifications |
| 4     | Manual required | Cannot auto  | security, architecture           |

### Quick Start Examples

```bash
# Single scan and fix (parallel by default)
> /moai fix

# Autonomous loop until resolved
> /moai loop --max 50

# Full autonomous development (parallel by default)
> /moai "Add JWT authentication" --loop
```

### Why "All is Well"?

The name reflects the philosophy: **set the goal and trust the AI**. MoAI-ADK's agentic automation handles the entire development cycle autonomously. When you see `<moai>DONE</moai>`, all is truly well.

---

## 5. MoAI Orchestrator and Sub-Agents

### 🎩 MoAI - Strategic Orchestrator

**Role**: Analyze user requests and delegate to appropriate specialized agents

**Workflow**:

1. **Understand**: Request analysis and clarification
2. **Plan**: Establish execution plan via Plan agent
3. **Execute**: Delegate tasks to specialized agents (sequential/parallel)
4. **Integrate**: Integrate results and report to user

### 🌐 Multilingual Automatic Routing (NEW)

MoAI automatically recognizes 4 language requests and invokes the correct agent:

| Request Language | Example                     | Agent Invoked  |
| ---------------- | --------------------------- | -------------- |
| English          | "Design backend API"        | expert-backend |
| Korean           | "백엔드 API 설계해줘"       | expert-backend |
| Japanese         | "バックエンドAPIを設計して" | expert-backend |
| Chinese          | "设计后端API"               | expert-backend |

---

### 🔧 Tier 1: Domain Experts (9)

| Agent                  | Expertise                   | Usage Example                            |
| ---------------------- | --------------------------- | ---------------------------------------- |
| **expert-backend**     | FastAPI, Django, DB design  | API design, query optimization           |
| **expert-frontend**    | React, Vue, Next.js         | UI components, state management          |
| **expert-stitch**      | Google Stitch, UI/UX design | AI-powered UI generation                 |
| **expert-security**    | Security analysis, OWASP    | Security audit, vulnerability analysis   |
| **expert-devops**      | Docker, K8s, CI/CD          | Deployment automation, infrastructure    |
| **expert-debug**       | Bug analysis, performance   | Problem diagnosis, bottleneck resolution |
| **expert-performance** | Profiling, optimization     | Response time improvement                |
| **expert-refactoring** | Code refactoring, AST-Grep  | Large-scale code transformation          |
| **expert-testing**     | Test strategy, E2E          | Test planning, coverage                  |

---

### 🎯 Tier 2: Workflow Managers (7)

| Agent                | Role                             | Auto Invocation Timing          |
| -------------------- | -------------------------------- | ------------------------------- |
| **manager-spec**     | SPEC writing (EARS)              | `/moai plan`                    |
| **manager-ddd**      | DDD implementation execution     | `/moai run`                     |
| **manager-docs**     | Documentation auto generation    | `/moai sync`                    |
| **manager-quality**  | TRUST 5 verification             | After implementation completion |
| **manager-strategy** | Execution strategy establishment | Complex planning                |
| **manager-project**  | Project initialization & settings | `/moai project`                 |
| **manager-git**      | Git workflow                     | Branch/PR management            |

---

### 🏗️ Tier 3: Claude Code Builder (4)

| Agent               | Role               | Usage Example                 |
| ------------------- | ------------------ | ----------------------------- |
| **builder-agent**   | Create new agent   | Organization specialist agent |
| **builder-skill**   | Create new skill   | Team-specific skill module    |
| **builder-command** | Create new command | Custom workflow               |
| **builder-plugin**  | Create plugin      | Distribution plugin           |

---

### 🧠 Sequential Thinking MCP Support

All agents support `--ultrathink` flag for deep analysis using Sequential Thinking MCP:

**Usage**:
```bash
> /moai "Add JWT authentication" --ultrathink
```

**Agent-Specific UltraThink Examples**:

| Agent Type          | UltraThink Deep Analysis Focus                          |
| ------------------- | ------------------------------------------------------ |
| **expert-backend**  | API design patterns, database schema, query optimization |
| **expert-frontend** | Component architecture, state management, UI/UX design   |
| **expert-security** | Threat analysis, vulnerability patterns, OWASP compliance |
| **expert-devops**   | Deployment strategies, CI/CD pipelines, infrastructure    |
| **manager-ddd**     | Refactoring strategy, behavior preservation, legacy code |
| **manager-spec**    | Requirements analysis, acceptance criteria, user stories |

When `--ultrathink` is appended to any agent invocation, the agent activates Sequential Thinking MCP to:
- Break down complex problems into manageable steps
- Analyze domain-specific patterns and best practices
- Map subtasks to appropriate execution strategies
- Generate optimal implementation plans

---

## 5. Ralph-Style LSP Integration Workflow (NEW v1.9.0)

### LSP-Based Autonomous Development

MoAI-ADK implements Ralph-style autonomous workflow with LSP diagnostics integration. The system captures LSP state at workflow entry, monitors during execution, and automatically completes phases when quality thresholds are met.

### Workflow Phases

```mermaid
flowchart TB
    Start([SPEC-001]) --> Plan[Plan Phase]

    Plan --> PlanLSP[LSP Baseline Capture]
    PlanLSP --> PlanCheck{0 errors?}
    PlanCheck -->|Yes| Run[Run Phase]
    PlanCheck -->|No| PlanFix[Fix issues]

    Run --> RunLSP[LSP State Monitoring]
    RunLSP --> RunCheck{0 errors<br/>0 type errors<br/>Coverage >= 85%?}
    RunCheck -->|Yes| Sync[Sync Phase]
    RunCheck -->|No| RunFix[Autonomous Loop]

    RunFix --> RunLSP

    Sync --> SyncLSP[Final LSP Check]
    SyncLSP --> SyncCheck{0 errors<br/><10 warnings?}
    SyncCheck -->|Yes| Done[<moai>DONE</moai>]
    SyncCheck -->|No| SyncFix[Fix and retry]

    SyncFix --> SyncLSP

    Done --> End([Complete])

    style Plan fill:#fff3e0
    style Run fill:#f3e5f5
    style Sync fill:#e8f5e9
    style Done fill:#c8e6c9
    style End fill:#4caf50,color:#fff
```

### Completion Markers

**Plan Phase**:
- SPEC document created
- LSP baseline recorded
- No blocking issues

**Run Phase**:
- 0 errors
- 0 type errors
- Coverage >= 85%
- Tests passing

**Sync Phase**:
- 0 errors
- < 10 warnings
- Documentation synchronized

### Configuration

```yaml
# .moai/config/sections/workflow.yaml
execution_mode:
  autonomous:
    user_approval_required: false
    continuous_loop: true
    completion_marker_based: true
    lsp_feedback_integration: true
```

### Loop Prevention

- Maximum 100 iterations
- No progress detection (5 iterations)
- Alternative strategy on stall
- Snapshot save for resume

---

## 6. Agent-Skills

### 📚 Skill Library Structure

```text
🏗️ Foundation (6)    → Core philosophy, execution rules
🎯 Domain (4)        → Domain expertise
💻 Language (16)     → 16 programming languages
🚀 Platform (10)     → Cloud/BaaS integration
📋 Workflow (8)      → Automation workflows
📚 Library (3)       → Special libraries
🛠️ Tool (2)          → Development tools
📑 Docs (1)          → Documentation generation
📊 Formats (1)       → Data format handling
🖥️ Framework (1)     → Application frameworks
```

### Frequently Used Skill Combinations

| Purpose           | Skill Combination                                                        |
| ----------------- | ------------------------------------------------------------------------ |
| **Backend API**   | `moai-lang-python` + `moai-domain-backend` + `moai-platform-supabase`    |
| **Frontend UI**   | `moai-lang-typescript` + `moai-domain-frontend` + `moai-library-shadcn`  |
| **Documentation** | `moai-library-nextra` + `moai-workflow-docs` + `moai-library-mermaid`    |
| **Testing**       | `moai-lang-python` + `moai-workflow-testing` + `moai-foundation-quality` |

### Skill Usage

```python
# Method 1: Direct call (Agent)
Skill("moai-lang-python")

# Method 2: MoAI auto selection (general user)
"FastAPI server 만들어줘"
→ MoAI automatically selects moai-lang-python
```

---

## 7. Google Stitch MCP - AI-Powered UI/UX Design

### Overview

**Google Stitch** is an AI-powered design tool that generates UI screens from text descriptions. MoAI-ADK automates design context extraction, screen generation, and code export through Stitch MCP integration.

### Key Features

| Feature                       | Description                                              |
| ----------------------------- | -------------------------------------------------------- |
| `generate_screen_from_text`   | Generate UI screens from text descriptions               |
| `extract_design_context`      | Extract "Design DNA" from existing screens (colors, fonts, layout) |
| `fetch_screen_code`           | Download generated screen HTML/CSS/JS code               |
| `fetch_screen_image`          | Download screen screenshots                              |
| `create_project` / `list_projects` | Manage Stitch projects                              |

### Quick Start

**1. Google Cloud Setup**:

```bash
# Enable Stitch API
gcloud beta services mcp enable stitch.googleapis.com

# Authenticate
gcloud auth application-default login

# Set environment variable (.bashrc or .zshrc)
export GOOGLE_CLOUD_PROJECT="YOUR_PROJECT_ID"
```

**2. MCP Configuration** (`.claude/.mcp.json`):

```json
{
  "mcpServers": {
    "stitch": {
      "command": "npx",
      "args": ["-y", "stitch-mcp"],
      "env": {
        "GOOGLE_CLOUD_PROJECT": "YOUR_PROJECT_ID"
      }
    }
  }
}
```

### Usage Example

**Generate Login Screen**:

```text
> Create a login screen with email input, password input (show/hide toggle),
> login button, forgot password link, and social login (Google, Apple).
> Mobile: vertical stack. Desktop: 400px centered card.
```

**Design Consistency Workflow**:

1. `extract_design_context`: Extract design tokens from existing screen
2. `generate_screen_from_text`: Generate new screen with extracted context
3. `fetch_screen_code`: Export production-ready code

### Prompt Writing Tips

| Item | Recommendation |
| ---- | -------------- |
| **Components** | Specify required UI elements like buttons, inputs, cards |
| **Layout** | Specify single-column, grid, sidebar, etc. |
| **Responsive** | Describe mobile/desktop behavior |
| **Style** | Define colors, fonts, hover effects |

> **Note**: Requesting one screen at a time with one or two modifications yields the best results.

### Detailed Documentation

For complete prompt templates, error handling, and advanced patterns, see the skill documentation:

- **Skill**: `.claude/skills/moai-platform-stitch/SKILL.md`
- **Agent**: `expert-stitch` (UI/UX design specialist agent)

---

## 7.1 Memory MCP - Persistent Storage Across Sessions

### Overview

**Memory MCP** enables persistent storage across Claude Code sessions, allowing MoAI to remember user preferences, project context, and learned patterns.

### Key Features

| Feature | Description |
| --- | --- |
| **User Preferences** | Remember conversation language, coding style, naming conventions |
| **Project Context** | Persist tech stack, architecture decisions, project conventions |
| **Learned Patterns** | Store frequently used libraries, common error resolutions |
| **Session State** | Track last worked SPEC, pending tasks across sessions |

### Memory Categories

| Prefix | Category | Examples |
| --- | --- | --- |
| `user_` | User Preferences | `user_language`, `user_coding_style` |
| `project_` | Project Context | `project_tech_stack`, `project_architecture` |
| `pattern_` | Learned Patterns | `pattern_preferred_libraries`, `pattern_error_resolutions` |
| `session_` | Session State | `session_last_spec`, `session_pending_tasks` |

### Installation

Add Memory MCP to your Claude Code configuration:

```json
// .claude/settings.local.json
{
  "mcpServers": {
    "memory": {
      "command": "npx",
      "args": ["-y", "@anthropic/memory-mcp-server"]
    }
  }
}
```

### Usage Examples

**Store User Preference**:
```
"Remember that I prefer Korean for conversations"
→ MoAI stores: user_language = "ko"
```

**Learn from Corrections**:
```
"Use snake_case for Python variables"
→ MoAI stores: user_coding_style = "snake_case"
```

**Retrieve Context**:
```
"What was the last SPEC I was working on?"
→ MoAI retrieves: session_last_spec
```

### Best Practices

- Use descriptive, categorized key names
- Keep values concise (under 1000 characters)
- Never store sensitive credentials
- Store preferences, not personal data

---

## 8. TRUST 5 Quality Principles

All MoAI-ADK projects follow the **TRUST 5** quality framework.

### 🏆 TRUST 5 = Test + Readable + Unified + Secured + Trackable

**NEW v1.9.0**: LSP Integration for Quality Monitoring

```mermaid
graph TD
    T1["🔴 T: Tested<br/>━━━━━━━━<br/>• DDD with tests<br/>• 85%+ coverage<br/>• Behavior preserved<br/>• LSP: 0 type errors"]
    R["📖 R: Readable<br/>━━━━━━━━<br/>• Clear naming<br/>• Code comments<br/>• Linter compliance<br/>• LSP: 0 lint errors"]
    U["🔄 U: Unified<br/>━━━━━━━━<br/>• Consistent style<br/>• Standard patterns<br/>• Error handling<br/>• LSP: <10 warnings"]
    S["🔒 S: Secured<br/>━━━━━━━━<br/>• OWASP Top 10<br/>• Vulnerability scan<br/>• Encryption policy<br/>• LSP: 0 security warnings"]
    T2["📋 T: Trackable<br/>━━━━━━━━<br/>• Clear commits<br/>• Issue tracking<br/>• CHANGELOG<br/>• LSP: State tracked"]

    T1 --> R --> U --> S --> T2 --> Deploy["✅ Production Ready"]
```

### LSP Integration with TRUST 5 (NEW v1.9.0)

| TRUST 5 Pillar | LSP Quality Indicators |
| -------------- | ---------------------- |
| **Tested** | Unit tests pass, LSP type errors = 0 |
| **Readable** | Linting clean, naming conventions followed |
| **Unified** | Warnings < threshold, consistent patterns |
| **Secured** | Security scan pass, no security warnings |
| **Trackable** | LSP state changes logged, diagnostics tracked |

### T - Tested

**Principle**: All implementation is validated with tests

**Verification**:

- Test coverage >= 85%
- Characterization tests for existing code
- Behavior preservation verified
- Incremental improvements

### R - Readable

**Principle**: Code must be clear and easy to understand

**Verification**:

- Clear variable names
- Comments on complex logic
- Pass code review
- Pass linter checks

### U - Unified

**Principle**: Maintain consistent style across project

**Verification**:

- Follow project style guide
- Consistent naming conventions
- Unified error handling
- Standard document format

### S - Secured

**Principle**: All code must pass security verification

**Verification**:

- OWASP Top 10 checks
- Dependency vulnerability scanning
- Encryption policy compliance
- Access control verification

### T - Trackable

**Principle**: All changes must be clearly trackable

**Verification**:

- Clear commit messages
- Issue tracking (GitHub Issues)
- Maintain CHANGELOG
- Code review records

---

## 9. Automated Quality Checks

### 🔍 AST-Grep Based Structural Inspection

**AST-Grep** analyzes **code structure** not text:

| Feature                    | Description                  | Example                                   |
| -------------------------- | ---------------------------- | ----------------------------------------- |
| **Structural Search**      | AST pattern matching         | Find unparameterized SQL queries          |
| **Security Scan**          | Auto vulnerability detection | SQL Injection, XSS, hardcoded secrets     |
| **Pattern Refactoring**    | Safe code transformation     | Bulk variable rename, function extraction |
| **Multi-language Support** | 40+ languages                | Python, TypeScript, Go, Rust...           |

### Auto Inspection Flow

```text
Code writing
    ↓
[Hook] AST-Grep auto scan
    ↓
⚠️  Immediate alert on vulnerability detection
    ↓
✅ Refactor to safe code
```

**Detection Example**:

```bash
⚠️  AST-Grep: Potential SQL injection in src/auth.py:47
   Pattern: execute(f"SELECT * FROM users WHERE id={user_id}")
   Suggestion: execute("SELECT * FROM users WHERE id=%s", (user_id,))
```

### 🛡️ Security Guard - Command Protection

MoAI-ADK includes a **Security Guard hook** that protects against dangerous operations:

| Category | Protected Commands | Platform |
|----------|-------------------|----------|
| **Database Deletion** | `supabase db reset`, `neon database delete`, `pscale database delete` | All |
| **SQL Dangerous** | `DROP DATABASE`, `DROP SCHEMA`, `TRUNCATE TABLE` | All |
| **File Deletion** | `rm -rf /`, `rm -rf ~`, `rm -rf .git` | Unix |
| **File Deletion** | `rd /s /q C:\`, `Remove-Item -Recurse -Force` | Windows |
| **Git Dangerous** | `git push --force origin main`, `git branch -D main` | All |
| **Cloud Infra** | `terraform destroy`, `az group delete`, `aws delete-*` | All |
| **Docker Cleanup** | `docker system prune -a`, `docker volume prune`, `docker image prune -a` | All |

**Protection Levels**:

| Level | Action | Example |
|-------|--------|---------|
| **Deny** | Blocked immediately | `rm -rf /`, `DROP DATABASE`, `docker system prune -a` |
| **Ask** | Requires user confirmation | `git reset --hard`, `prisma migrate reset` |
| **Allow** | Proceeds normally | Safe operations |

**How it works**:

```text
User executes command
    ↓
[Hook] Security Guard scan
    ↓
⚠️  Dangerous pattern detected → Block or Ask
    ↓
✅ Safe to proceed
```

> **Note**: The security guard protects both Unix (macOS/Linux) and Windows users with platform-specific command patterns.

**Manual Execution**: When you actually need to run these commands, execute them directly in your terminal:

```bash
# Docker cleanup (run manually when needed)
docker system prune -a        # Remove all unused images, containers, networks
docker volume prune           # Remove unused volumes (⚠️ data loss risk)
docker image prune -a         # Remove all unused images

# Database operations (run manually when needed)
supabase db reset             # Reset local database
DROP DATABASE dbname;         # SQL: Drop database

# File operations (run manually when needed)
rm -rf node_modules           # Remove node_modules
```

---

## 10. 📊 Statusline Customization

MoAI-ADK provides a **customizable statusline** that displays real-time project information in Claude Code.

### Default Statusline Layout

```text
🤖 Opus 4.5 | 💰 152K/200K | 💬 MoAI | 📁 MoAI-ADK | 📊 +0 M58 ?5 | 💾 57.7MB | 🔀 main
```

### Statusline Components

| Icon | Component      | Description                            | Config Key       |
| ---- | -------------- | -------------------------------------- | ---------------- |
| 🤖   | **Model**      | Claude model (Opus, Sonnet, etc.)      | `model`          |
| 💰   | **Context**    | Context window usage (e.g., 77K/200K)  | `context_window` |
| 💬   | **Style**      | Active output style (e.g., MoAI)       | `output_style`   |
| 📁   | **Directory**  | Current project name                   | `directory`      |
| 📊   | **Git Status** | Staged/Modified/Untracked files        | `git_status`     |
| 💾   | **Memory**     | Process memory usage                   | `memory_usage`   |
| 🔀   | **Branch**     | Current Git branch                     | `branch`         |
| 🔅   | **Version**    | Claude Code version (optional)         | `version`        |

### Configuration

Edit `.moai/config/statusline-config.yaml`:

```yaml
display:
  model: true # 🤖 Claude model
  context_window: true # 💰 Context window
  output_style: true # 💬 Output style
  directory: true # 📁 Project name
  git_status: true # 📊 Git status
  memory_usage: true # 💾 Memory usage
  branch: true # 🔀 Git branch
  version: 1.1.0 # 🔅 Version (optional)
  active_task: true # Active task
```

### Memory Collector

When `memory_usage` is enabled, MoAI-ADK uses `psutil` to collect real-time memory usage:

- **Process Memory**: RSS (Resident Set Size) of the current Python process
- **Caching**: Performance optimization with 10-second TTL
- **Cross-Platform**: Supports macOS, Linux, Windows
- **Graceful Degradation**: Displays "N/A" when psutil is unavailable

### Display Modes

| Mode       | Max Length | Use Case          |
| ---------- | ---------- | ----------------- |
| `compact`  | 80 chars   | Standard terminal |
| `extended` | 120 chars  | Wide terminal     |
| `minimal`  | 40 chars   | Narrow terminal   |

Set mode:

```bash
export MOAI_STATUSLINE_MODE=extended
```

---

## 11. 🌳 Worktree Parallel Development

MoAI-ADK's core innovation: **Worktree for complete isolation, unlimited parallel development**

### 💡 Why Worktree?

**Problem**: Changing LLMs with `moai glm`/`moai cc` applies to **all open sessions**. Changing models in the same session leads to authentication errors, making continuation difficult.

**Solution**: Git Worktree completely isolates each SPEC to maintain independent LLM settings

---

### 📦 Worktree Workflow

```text
┌─────────────────────────────────────────────────────────────────┐
│  Terminal 1 (Claude Opus) - SPEC Design Only                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ cd my-project                                                │
│  $ claude                                                        │
│                                                                  │
│  > /moai plan "사용자 인증 시스템" --worktree                     │
│  ✅ SPEC-AUTH-001 생성 완료                                      │
│  ✅ Worktree 생성: ~/moai/worktrees/my-project/SPEC-AUTH-001     │
│  ✅ Branch: feature/SPEC-AUTH-001                                │
│                                                                  │
│  > /moai plan "결제 시스템" --worktree                            │
│  ✅ SPEC-PAY-002 생성 완료                                       │
│  ✅ Worktree 생성: ~/moai/worktrees/my-project/SPEC-PAY-002      │
│                                                                  │
│  > /moai plan "대시보드 UI" --worktree                           │
│  ✅ SPEC-UI-003 생성 완료                                        │
│  ✅ Worktree 생성: ~/moai/worktrees/my-project/SPEC-UI-003       │
│                                                                  │
│  💡 Opus로 모든 SPEC 계획 완료 (세션 유지 중...)                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Terminal 2 - SPEC-AUTH-001 Worktree (GLM 4.7)                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ moai-worktree go SPEC-AUTH-001                                │
│  # Or shortcut: moai-wt go SPEC-AUTH-001                          │
│                                                                  │
│  📁 Current location: ~/moai/worktrees/my-project/SPEC-AUTH-001        │
│  🔀 Branch: feature/SPEC-AUTH-001                                │
│                                                                  │
│  $ moai glm                                                       │
│  ✅ Switched to GLM backend                                      │
│                                                                  │
│  $ claude                                                        │
│  > /moai run SPEC-AUTH-001                                       │
│  🔄 DDD Running... (Analyze → Preserve → Improve)                 │
│  ✅ 구현 완료!                                                   │
│  ✅ 테스트 통과 (Coverage: 92%)                                  │
│                                                                  │
│  > /moai sync SPEC-AUTH-001                                      │
│  ✅ 문서 동기화 완료                                             │
│                                                                  │
│  # After completion, merge                                       │
│  $ git checkout main                                             │
│  $ git merge feature/SPEC-AUTH-001                               │
│  $ moai-worktree clean --merged-only                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Terminal 3 - SPEC-PAY-002 Worktree (GLM 4.7)                     │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ moai-wt go SPEC-PAY-002                                       │
│  $ moai glm                                                       │
│  $ claude                                                        │
│                                                                  │
│  > /moai SPEC-PAY-002                                            │
│  🔄 Plan → Run → Sync 자동 실행                                  │
│  ✅ 완료!                                                        │
│                                                                  │
│  $ git checkout main && git merge feature/SPEC-PAY-002           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Terminal 4 - SPEC-UI-003 Worktree (GLM 4.7)                      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  $ moai-wt go SPEC-UI-003                                        │
│  $ moai glm                                                       │
│  $ claude                                                        │
│  > /moai SPEC-UI-003                                             │
│  ✅ 완료!                                                        │
└─────────────────────────────────────────────────────────────────┘
```

---

### 🎯 Core Workflow

#### Phase 1: Plan with Claude 4.5 Opus (Terminal 1)

```bash
/moai plan "feature description" --worktree
```

- ✅ SPEC document generation
- ✅ Worktree auto creation
- ✅ Feature branch auto creation

#### Phase 2: Implement with GLM 4.7 (Terminals 2, 3, 4...)

```bash
moai-wt go SPEC-ID
moai glm
claude
> /moai run SPEC-ID
> /moai sync SPEC-ID
```

- ✅ Isolated work environment
- ✅ GLM cost efficiency
- ✅ Conflict-free parallel development

#### Phase 3: Merge and Cleanup

```bash
# Option 1: Complete in one step (Recommended)
moai-wt done SPEC-ID              # checkout main → merge → cleanup
moai-wt done SPEC-ID --push       # above + push to remote

# Option 2: Manual steps
git checkout main
git merge feature/SPEC-ID
moai-wt clean --merged-only
```

---

### ✨ Worktree Benefits

| Benefit                | Description                                              |
| ---------------------- | -------------------------------------------------------- |
| **Complete Isolation** | Each SPEC has independent Git state, no file conflicts   |
| **LLM Independence**   | Separate LLM settings per Worktree                       |
| **Unlimited Parallel** | Unlimited SPEC parallel development without dependencies |
| **Safe Merge**         | Only completed SPECs sequentially merge to main          |

---

### 📊 Worktree Commands

| Command                  | Description                        | Usage Example                  |
| ------------------------ | ---------------------------------- | ------------------------------ |
| `moai-wt new SPEC-ID`    | Create new Worktree                | `moai-wt new SPEC-AUTH-001`    |
| `moai-wt go SPEC-ID`     | Enter Worktree (open new shell)    | `moai-wt go SPEC-AUTH-001`     |
| `moai-wt list`           | List Worktrees                     | `moai-wt list`                 |
| `moai-wt done SPEC-ID`   | Merge and cleanup (checkout→merge) | `moai-wt done SPEC-AUTH-001`   |
| `moai-wt remove SPEC-ID` | Remove Worktree                    | `moai-wt remove SPEC-AUTH-001` |
| `moai-wt status`         | Check Worktree status and registry | `moai-wt status`               |
| `moai-wt sync [SPEC-ID]` | Synchronize Worktree               | `moai-wt sync --all`           |
| `moai-wt clean`          | Clean merged Worktrees             | `moai-wt clean --merged-only`  |
| `moai-wt recover`        | Recover registry from disk         | `moai-wt recover`              |
| `moai-wt config`         | Check Worktree settings            | `moai-wt config root`          |

---

## 12. Understanding CLAUDE.md

The `CLAUDE.md` file generated in your project root after MoAI-ADK installation is **MoAI's (AI Orchestrator) execution directive**. This file defines how Claude Code behaves in your project.

### What is CLAUDE.md?

`CLAUDE.md` is a project configuration file that Claude Code automatically reads at session start. In MoAI-ADK, this file defines the behavioral rules for the **MoAI orchestrator**.

```text
📁 Project Root
├── CLAUDE.md              ← MoAI execution directive (don't modify)
├── CLAUDE.local.md        ← Personal custom directives (optional)
├── .claude/
│   ├── settings.json      ← Claude Code settings (overwritten on update)
│   ├── settings.local.json← Personal settings (optional, not overwritten)
│   ├── agents/            ← Sub-agent definitions
│   ├── commands/          ← Slash commands
│   └── skills/            ← Skill definitions
└── .moai/
    └── config/            ← MoAI configuration
```

### CLAUDE.md Core Structure

| Section                         | Description              | Key Content                                            |
| ------------------------------- | ------------------------ | ------------------------------------------------------ |
| **Core Identity**               | MoAI's role definition   | Strategic orchestrator, HARD rules                     |
| **Request Processing Pipeline** | Request handling flow    | Analyze → Route → Execute → Report                     |
| **Command Reference**           | Command classification   | Type A (Workflow), Type B (Utility), Type C (Feedback) |
| **Agent Catalog**               | Sub-agent listing        | 8 Managers, 8 Experts, 4 Builders                      |
| **SPEC-Based Workflow**         | SPEC-driven development  | Plan → Run → Sync flow                                 |
| **Quality Gates**               | Quality validation rules | HARD/SOFT rules checklist                              |
| **Configuration Reference**     | Settings reference       | Language, output format rules                          |

### Usage: Don't Modify It

> **Recommendation**: Use `CLAUDE.md` **as-is without modification**.

**Reasons**:

- Automatically replaced with the latest version during MoAI-ADK updates
- Modifications may cause update conflicts
- Ensures consistent behavior across agents

```bash
# CLAUDE.md is automatically updated
moai update
```

### Customization: Use CLAUDE.local.md

If you need additional directives, create a **`CLAUDE.local.md`** file.

```bash
# Create CLAUDE.local.md in project root
touch CLAUDE.local.md
```

**CLAUDE.local.md Example**:

```markdown
# Project Local Directives

## Coding Style

- Type hints required for all functions
- Use Google-style docstrings

## Project-Specific Rules

- API responses always use snake_case
- Test files must have test\_ prefix

## Prohibited

- No console.log (use logger)
- No any type usage
```

**Benefits**:

- No conflicts with `CLAUDE.md` updates
- Project-specific customization
- Can add to `.gitignore` for personal settings

### CLAUDE.md vs CLAUDE.local.md

| Aspect           | CLAUDE.md                   | CLAUDE.local.md                        |
| ---------------- | --------------------------- | -------------------------------------- |
| **Purpose**      | MoAI execution directives   | Personal/project additional directives |
| **Modification** | Not recommended             | Freely modifiable                      |
| **Updates**      | Managed by MoAI             | Managed by user                        |
| **Git**          | Committed                   | Optional (.gitignore possible)         |
| **Priority**     | Base rules                  | Additional/override rules              |

### Settings Customization: Use settings.local.json

Since v1.8.12, `.claude/settings.json` is **always overwritten** during `moai update`. For personal customizations, create a **`settings.local.json`** file.

```bash
# Create settings.local.json in .claude/ directory
touch .claude/settings.local.json
```

**settings.local.json Example**:

```json
{
  "env": {
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "DISABLE_NON_ESSENTIAL_MODEL_CALLS": "1",
    "MY_CUSTOM_VAR": "value"
  },
  "permissions": {
    "allow": [
      "Bash(docker:*)",
      "Bash(kubectl:*)"
    ]
  }
}
```

**Benefits**:

- No conflicts with `settings.json` updates
- Personal environment variables and permissions
- Can add to `.gitignore` for private settings

### settings.json vs settings.local.json

| Aspect           | settings.json               | settings.local.json                    |
| ---------------- | --------------------------- | -------------------------------------- |
| **Purpose**      | MoAI default settings       | Personal/project additional settings   |
| **Modification** | Not recommended             | Freely modifiable                      |
| **Updates**      | Overwritten by MoAI         | Managed by user                        |
| **Git**          | Committed                   | Optional (.gitignore possible)         |
| **Priority**     | Base settings               | Merged on top (higher priority)        |

### Core Rules (HARD Rules)

**HARD rules** defined in `CLAUDE.md` are always enforced:

1. **Language-Aware Responses**: Respond in user's language
2. **Parallel Execution**: Execute independent tasks in parallel
3. **No XML in User Responses**: No XML tags in user-facing responses

These rules cannot be overridden even by `CLAUDE.local.md`.

---

## 13. MoAI Rank Introduction

**A new dimension of agentic coding**: Track your coding journey and compete with global developers!

### Why MoAI Rank?

| Feature                      | Description                               |
| ---------------------------- | ----------------------------------------- |
| **📊 Token Tracking**        | Automatic AI usage tracking per session   |
| **🏆 Global Leaderboard**    | Daily/Weekly/Monthly/All-time rankings    |
| **🎭 Coding Style Analysis** | Discover your unique development patterns |
| **📈 Dashboard**             | Visualized statistics and insights        |

---

### 🚀 CLI Commands

```bash
❯ moai rank
Usage: moai rank [OPTIONS] COMMAND [ARGS]...

  MoAI Rank - Token usage leaderboard.

  Track your Claude Code token usage and compete on the leaderboard.
  Visit https://rank.mo.ai.kr for the web dashboard.

Commands:
  register   Register with MoAI Rank via GitHub OAuth.
  status     Show your current rank and statistics.
  exclude    Exclude a project from session tracking.
  include    Re-include a previously excluded project.
  logout     Remove stored MoAI Rank credentials.
```

---

### Step 1: GitHub OAuth Registration

```bash
❯ moai rank login

╭──────────────────────────── Registration ────────────────────────────╮
│ MoAI Rank Registration                                               │
│                                                                      │
│ This will open your browser to authorize with GitHub.                │
│ After authorization, your API key will be stored securely.           │
╰──────────────────────────────────────────────────────────────────────╯

Opening browser for GitHub authorization...
Waiting for authorization (timeout: 5 minutes)...

╭───────────────────────── Registration Complete ──────────────────────╮
│ Successfully registered as your-github-id                            │
│                                                                      │
│ API Key: moai_rank_a9011fac_c...                                     │
│ Stored in: ~/.moai/rank/credentials.json                             │
╰──────────────────────────────────────────────────────────────────────╯

╭───────────────────────── Global Hook Installed ──────────────────────╮
│ Session tracking hook installed globally.                            │
│                                                                      │
│ Your Claude Code sessions will be automatically tracked.             │
│ Hook location: ~/.claude/hooks/moai/session_end__rank_submit.py      │
│                                                                      │
│ To exclude specific projects:                                        │
│   moai rank exclude /path/to/project                                 │
╰──────────────────────────────────────────────────────────────────────╯
```

---

### Step 2: Sync Session Data

Sync your existing Claude Code session data to MoAI Rank.

```bash
❯ moai rank sync

Syncing 2577 session(s) to MoAI Rank
Phase 1: Parsing transcripts (parallel: 20 workers)

  Parsing transcripts ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (2577/2577)

Phase 2: Submitting 1873 session(s) (batch mode)
Batch size: 100 | Batches: 19

  Submitting batches ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (19/19)

Sync Complete
  ✓ Submitted: 1169
  ○ Skipped:   704 (no usage or duplicate)
  ✗ Failed:    0
```

---

### Step 3: Check Your Rank

```bash
❯ moai rank status

╭────────────────────────────── MoAI Rank ─────────────────────────────╮
│ your-github-id                                                       │
│                                                                      │
│ 🏆 Global Rank: #42                                                  │
╰──────────────────────────────────────────────────────────────────────╯
╭───── Daily ──────╮  ╭───── Weekly ─────╮  ╭──── Monthly ─────╮  ╭──── All Time ────╮
│ #12              │  │ #28              │  │ #42              │  │ #156             │
╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯
╭─────────────────────────── Token Usage ──────────────────────────────╮
│ 1,247,832 total tokens                                               │
│                                                                      │
│ Input  ██████████████░░░░░░ 847,291 (68%)                            │
│ Output ██████░░░░░░░░░░░░░░ 400,541 (32%)                            │
│                                                                      │
│ Sessions: 47                                                         │
╰──────────────────────────────────────────────────────────────────────╯

● Hook: Installed  |  https://rank.mo.ai.kr
```

---

### Step 3: Web Dashboard

![MoAI Rank Dashboard](./assets/images/readme/moai-rank-dashboard.png)

**[https://rank.mo.ai.kr](https://rank.mo.ai.kr)**

On the dashboard:

- Token usage trends
- Tool usage statistics
- Model-specific usage analysis
- Weekly/monthly reports

📖 **Details**: Refer to [modu-ai/moai-rank](https://github.com/modu-ai/moai-rank) repository.

---

### Step 4: Collected Metrics

| Metric           | Description                         |
| ---------------- | ----------------------------------- |
| **Token Usage**  | Input/output tokens, cache tokens   |
| **Tool Usage**   | Read, Edit, Bash usage counts       |
| **Model Usage**  | Opus, Sonnet, Haiku breakdown       |
| **Code Metrics** | Added/deleted lines, modified files |
| **Session Info** | Duration, turn count, timestamps    |

### 🔒 Privacy Protection

```bash
# Exclude current project
moai rank exclude

# Exclude specific path
moai rank exclude /path/to/private

# Wildcard pattern
moai rank exclude "*/confidential/*"

# List excluded
moai rank list-excluded
```

**Guarantee**: Collected data is **numeric metrics only** (code content, file paths not transmitted)

---

## 14. FAQ 5 Questions

### Q1: Is SPEC Always Required?

| Condition            | SPEC Required       |
| -------------------- | ------------------- |
| 1-2 files modified   | Optional (can skip) |
| 3-5 files modified   | Recommended         |
| 10+ files modified   | Required            |
| New feature addition | Recommended         |
| Bug fix              | Optional            |

### Q2: Is MCP Server Installation Required?

**Required (2)**:

- **Context7**: For latest library documentation and Skill reference generation
- **Sequential Thinking**: For structured problem-solving and step-by-step reasoning in complex tasks

**Recommended**:

- **Memory MCP**: Persistent storage across sessions for user preferences, project context, and learned patterns

**Optional**:

- claude-in-chrome: Use Claude in browser and web automation testing
- Playwright: Web automation testing
- Figma: Design system

### Q3: Does MoAI Rank Cost Money?

It's free. Only automatically collects session data.

### Q4: Is GLM Configuration Required?

No. You can use Claude only. However, it's recommended for cost savings.

### Q5: Can It Be Applied to Existing Projects?

Yes. `moai init .` preserves existing files.

---

## 15. Community & Support

### 🌐 Participate

- **Discord (Official)**: [https://discord.gg/umywNygN](https://discord.gg/umywNygN)
- **GitHub**: [https://github.com/modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **Developer Blog**: [https://goos.kim](https://goos.kim)

### 🆘 Support

- Email: [support@mo.ai.kr](mailto:support@mo.ai.kr)
- Documentation: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

---

## 15. Star History

[![Star History Chart](https://api.star-history.com/svg?repos=modu-ai/moai-adk&type=date&legend=top-left)](https://www.star-history.com/#modu-ai/moai-adk&type=date&legend=top-left)

---

## 16. License

Copyleft License (COPYLEFT-3.0) - [LICENSE](./LICENSE)

---

## 17. 🙏 Made with ❤️ by MoAI-ADK Team

**Last Updated:** 2026-01-28
**Philosophy**: SPEC-First DDD + Agent Orchestration + Hybrid LLM
**MoAI**: MoAI means "Modu-ui AI" (AI for Everyone).

> **"Infinite Possibilism - AI for All"**
