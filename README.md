# 🗿 MoAI-ADK: Agentic AI Development Framework

![MoAI-ADK Hero Banner](./assets/images/readme/hero-banner-moai-adk.png)

**Available Languages:** [🇰🇷 한국어](./README.ko.md) | [🇺🇸 English](./README.md) | [🇯🇵 日本語](./README.ja.md) | [🇨🇳 中文](./README.zh.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: Copyleft](https://img.shields.io/badge/License-Copyleft--3.0-blue.svg)](./LICENSE)
[![Python](https://img.shields.io/badge/Python-3.11--3.14-blue)](https://www.python.org/)

> **"The purpose of vibe coding is not rapid productivity but code quality."**

MoAI-ADK provides an **AI development environment for quality code**. SPEC-First TDD, test-driven development, continuous refactoring, and 20 specialized AI agents work together.

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
- **🔴 TDD Enforcement**: Automatic guarantee of 85%+ test coverage
- **🤖 AI Orchestration**: 20 specialized agents + 48 skills
- **🌐 Multilingual Routing**: Automatic support for Korean/English/Japanese/Chinese
- **🌳 Worktree Parallel Development**: Unlimited parallel work in completely isolated environments
- **🏆 MoAI Rank**: Motivation through vibe coding leaderboard

---

> **📚 For detailed information, see the official online documentation:** [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

## 1. 30-Second Installation

### 🚀 Method 1: Quick Install (Recommended)

```bash
curl -LsSf https://modu-ai.github.io/moai-adk/install.sh | sh
```

### 🔧 Method 2: Manual Install

```bash
# Step 1: Install uv (macOS/Linux)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Step 2: Install MoAI-ADK
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

#### Step 10: TAG System Activation

🎯 TAG System: Code ↔ Documentation Tracking for TDD

The TAG system maintains traceability between code and SPEC documents
to support the TDD (Test-Driven Development) cycle.

TDD Purpose:
•  RED (Write tests) → @SPEC SPEC-XXX verify
•  GREEN (Implement code) → @SPEC SPEC-XXX impl
•  REFACTOR (Improve) → @SPEC SPEC-XXX impl or related

Clearly tracks which SPEC each code file implements,
promoting documentation-driven development and maintaining quality.

💡 TAG activation is recommended. Maintains consistency between code and
documentation in the TDD cycle, improving quality and maintainability.

```text
? Enable TAG system? (TDD recommended) (Y/n)
```

---

**TAG Validation Mode Guide**

• warn: Warn during development (default, recommended)
• enforce: Block commit on missing TAG (strict quality control)
• off: Skip validation (not recommended)

```text
❯ Select TAG validation mode: [↑↓] Navigate  [Enter] Select
❯ warn (Warning) - Show warnings for missing TAGs only. Flexible during development
  enforce (Enforce) - Block commits if TAGs are missing. Strict quality control
  off (Off) - Skip TAG validation. Not recommended
```

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
  🌐 Language:   Auto-detect (use /moai:0-project)
  🔀 Git:        manual (github-flow, branch: manual)
  🌍 Locale:     ko
  📄 Files:      47 created
  ⏱️  Duration:   1234ms

🚀 Next Steps:
  1. Run cd my-project to enter the project
  2. Run /moai:0-project in Claude Code for full setup
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

## 2. Project Documentation Generation (Optional)

For new or existing projects, you can automatically generate **project documentation to help Claude Code understand your project**:

```text
> /moai:0-project
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

> 💡 **Tip**: Running `/moai:0-project` at project start or after structure changes updates documentation to the latest state.

---

## 3. Core Commands Collection

### 🎯 `/moai:0-project` - Project Initialization

```bash
> /moai:0-project
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

### 📋 `/moai:1-plan` - SPEC Writing

```bash
> /moai:1-plan "feature description"
```

Automatically generates unambiguous specifications using EARS format. Includes requirements definition, success criteria, and test scenarios to clearly present development direction. The generated SPEC serves as a single source of truth that the development team and AI share the same understanding.

**Auto-generated**:

- EARS format specification
- Requirements definition
- Success criteria
- Test scenarios

**Example**:

```bash
> /moai:1-plan "user profile page"
# → Creates SPEC-002

> /moai:1-plan "payment API"
# → Creates SPEC-003
```

**Important**: Must execute `> /clear` next

---

### 💻 `/moai:2-run` - TDD Implementation

```bash
> /moai:2-run SPEC-001
```

Through the Red-Green-Refactor cycle, writes tests first and implements code to pass them. All implementations must pass 85%+ test coverage, linting, type checking, and security checks. Automatically validates TRUST 5 quality principles to ensure only stable code is committed.

**Auto-executed**:

- 🔴 Write tests first (Red)
- 🟢 Pass tests with code (Green)
- 🔵 Refactor (Refactor)
- ✅ TRUST 5 validation

**Verification items**:

- Test coverage >= 85%
- Pass linting
- Pass type checking
- Pass security checks

---

### 📚 `/moai:3-sync` - Documentation Sync

```bash
> /clear  # Always run clear before sync to reset session and run quality checks
> /moai:3-sync SPEC-001
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

### 🚀 `/moai:alfred` - Fully Autonomous Automation

```bash
> /moai:alfred "feature description"
```

User presents the goal and AI autonomously performs exploration, planning, implementation, and verification. Analyzes codebase through parallel exploration and self-corrects issues through autonomous loops. Automatically terminates when completion marker (`<moai>DONE</moai>`) is detected, so developer only needs to verify final result.

#### Concept and Workflow

```mermaid
flowchart TB
    Start([User Request<br/>/moai:alfred 'feature description']) --> Phase0[Phase 0: Parallel Exploration]

    Phase0 --> Explore[🔍 Explore Agent<br/>Analyze codebase structure]
    Phase0 --> Research[📚 Research Agent<br/>Research technical docs]
    Phase0 --> Quality[✅ Quality Agent<br/>Assess quality state]

    Explore --> Phase1[Phase 1: SPEC Creation]
    Research --> Phase1
    Quality --> Phase1

    Phase1 --> Spec[📋 EARS format SPEC doc<br/>Requirements specification]

    Spec --> Phase2[Phase 2: TDD Implementation]

    Phase2 --> Red[🔴 RED: Write failing test]
    Red --> Green[🟢 GREEN: Minimal implementation]
    Green --> Refactor[🔵 REFACTOR: Code improvement]

    Refactor --> Check{Quality Check<br/>TRUST 5}

    Check -->|Pass| Phase3[Phase 3: Documentation Sync]
    Check -->|Fail| Loop[🔄 Autonomous Loop<br/>Auto-fix issues]

    Loop --> Red

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

3. **Phase 3: TDD Implementation** (Autonomous loop)
   - **RED**: Write failing test first
   - **GREEN**: Minimal implementation to pass test
   - **REFACTOR**: Improve code quality
   - **Loop**: Auto-fix if quality check fails

4. **Phase 4: Documentation Sync**
   - README, API docs auto-update
   - CHANGELOG auto-generation
   - User guide update

#### When to use?

| Situation                | Description                              | Example                         |
| ------------------------ | ---------------------------------------- | ------------------------------- |
| **New feature development** | AI handles everything from start to finish | "Add JWT authentication system" |
| **Complex refactoring**  | Large changes affecting multiple files   | "Database layer restructure"    |
| **Bug fixing**           | Automate from root cause analysis to fix | "Fix login failure bug"         |
| **SPEC-based development** | Implement features with SPEC documents  | `/moai:alfred SPEC-AUTH-001`    |

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
> /moai:alfred "Add JWT authentication"

# Auto loop with sequential exploration (for debugging)
> /moai:alfred "JWT authentication" --loop --seq

# Resume
> /moai:alfred resume SPEC-AUTH-001
```

---

### 🔁 `/moai:loop` - Autonomous Iterative Fixing

```bash
> /moai:loop
```

AI autonomously diagnoses and fixes LSP errors, test failures, and coverage deficiencies. Parallel diagnosis executes LSP, AST-grep, Tests, and Coverage simultaneously, resolving issues 3-4x faster. Runs autonomously until completion marker is detected or max iterations reached.

#### Concept and Workflow

```mermaid
flowchart TB
    Start([User Request<br/>/moai:loop]) --> Parallel[Parallel Diagnosis]

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

| Feature      | grep/regex                    | AST-grep            |
| ------------ | ----------------------------- | ------------------- |
| Search target| Text strings                  | Code structure (AST)|
| Example      | `print("hello")`              | `print(__)`         |
| Meaning      | Find "print" characters       | Find print function call pattern |
| Space sensitive? | Yes (whitespace, indent matters) | No (structure only) |
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

In `/moai:loop` and `/moai:fix` commands, AST-grep operates as a core component of parallel diagnosis:

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

| Situation                 | Description                             | Example                              |
| ------------------------ | ---------------------------------------- | ------------------------------------ |
| **Quality after implementation** | Automatically improve quality after coding | Run `/moai:loop` after feature implementation |
| **Test failure fix**     | Automatically analyze and fix test failures | Run after test execution on failure  |
| **Coverage improvement**  | Automatically achieve 85% target          | After writing new code               |
| **Refactoring**          | Continuously improve code quality         | Periodic execution for maintenance   |

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
> /moai:loop

# Sequential + auto-fix (for debugging)
> /moai:loop --seq --auto

# Max 50 iterations
> /moai:loop --max 50

# Restore snapshot
> /moai:loop --resume latest
```

---

### 🔧 `/moai:fix` - One-Shot Auto Fix

```bash
> /moai:fix
```

Scans LSP errors and linting issues in parallel and fixes them at once. Level 1-2 fixes immediately, Level 3 fixes after user approval, Level 4 reports as requiring manual fix. Use `--dry` option to preview before applying actual fixes.

#### Concept and Workflow

```mermaid
flowchart
    Start([User Request<br/>/moai:fix]) --> Dry{--dry mode?}

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

| Level | Description     | Risk    | Approval | Auto   | Example                         |
| ----- | --------------- | ------- | -------- | ------ | ------------------------------- |
| 1     | Immediate fix   | Low     | Not needed | ✅      | import sorting, whitespace      |
| 2     | Safe fix        | Low     | Log only   | ✅      | variable names, type addition   |
| 3     | Approval needed | Medium  | Needed    | ⚠️      | logic changes, API modification |
| 4     | Manual needed   | High    | Impossible | ❌      | security, architecture          |

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

| Situation                  | Description                             | Example                              |
| ------------------------ | ---------------------------------------- | ------------------------------------ |
| **Clean up after coding** | Clean up style, formatting all at once  | Run `/moai:fix` after coding         |
| **Pre-commit inspection**  | Resolve LSP errors, linting in advance   | Run before `git commit`              |
| **Quick fix**             | Solve at once without repetitive fixes   | When fixing simple issues            |
| **Preview confirmation**  | Check fix content before applying       | Run `/moai:fix --dry`                |

#### `/moai:fix` vs `/moai:loop` Selection Guide

```mermaid
flowchart
    Start([Code quality issue occurred]) --> Question{Need repetitive<br/>fixing?}

    Question -->|Yes| Loop[🔄 /moai:loop<br/>Autonomous iterative fixing]
    Question -->|No| Fix{Can solve<br/>at once?}

    Fix -->|Yes| FixCmd[🔧 /moai:fix<br/>One-shot auto fix]
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
> /moai:fix

# Sequential scan (for debugging)
> /moai:fix --seq

# Preview
> /moai:fix --dry

# Specific files
> /moai:fix src/auth.py
```

---

### ⏹️ `/moai:cancel-loop` - Cancel Loop

```bash
> /moai:cancel-loop
```

Safely cancels running autonomous loop and saves all progress state as snapshot. On cancel, TODO status, fix history, and iteration count are all preserved for later resumption. Use `--list` option to view previous snapshots and restore to specific point.

**Options**:

- `--force`: Cancel without confirmation
- `--snapshot`: Save snapshot
- `--keep`: Preserve state file
- `--reason TEXT`: Record cancellation reason
- `--list`: List snapshots

**Restore**:

```bash
# Restore latest snapshot
> /moai:loop --resume latest

# Restore specific snapshot
> /moai:loop --resume cancel-20240111-105230
```

**Example**:

```bash
# Basic cancel (after confirmation)
> /moai:cancel-loop

# Force cancel
> /moai:cancel-loop --force

# Save snapshot
> /moai:cancel-loop --snapshot

# List snapshots
> /moai:cancel-loop --list
```

**Actual Output Result**:

```text

⏺ ✅ Loop Cancel Complete

  Status Report

  ═══════════════════════════════════════════════════════
           Loop: Cancelled (No Active Loop)
  ═══════════════════════════════════════════════════════

  ### Status
  - Running loop: None
  - Completed tasks: 100%
  - Remaining TODO: 0

  ### Completed Tasks
  ✅ README.ko.md "3. 핵심 명령어 모음" 섹션 업데이트
  ✅ Fenced code blocks 언어 지정 추가 (MD040) - 9건
  ✅ Emphasis → Heading 변경 (MD036) - 2건
  ✅ Table column style 수정 (MD060) - 2건

  ### Changes Summary
  - 명령어 개수: 4개 → 8개 (loop, fix, cancel-loop 추가)
  - 설명 스타일: 간단 목록 → 3문장 이상 상세 설명
  - Markdown linting: 13건 이슈 → 0건

  ═══════════════════════════════════════════════════════

  Final File Status

  README.ko.md:
  - 모든 markdown linting 이슈 해결
  - 8개 핵심 명령어 상세 설명 완료
  - 표준 markdown 형식 준수

  ---
  모든 작업이 성공적으로 완료되었습니다! 🎉
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

| Command | Scope | Iteration | Purpose |
|---------|-------|-----------|---------|
| `/moai:fix` | Code fixes only | 1 time | Single scan + auto-fix |
| `/moai:loop` | Code fixes | Until marker/max | Autonomous iterative fixing |
| `/moai:alfred` | Full dev cycle | Until marker/max | Goal → SPEC → Implement → Docs |

### Command Chain Relationship

```text
/moai:alfred (Full Automation)
     │
     ├── Phase 0: Parallel Exploration
     ├── Phase 1: SPEC Generation
     ├── Phase 2: Implementation ←─┬── /moai:loop (Iterative)
     │                             │        │
     │                             │        └── /moai:fix (Single)
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

| Level | Description | Approval | Examples |
|-------|-------------|----------|----------|
| 1 | Immediate fix | Not required | import sorting, whitespace |
| 2 | Safe fix | Log only | variable rename, type hints |
| 3 | Approval needed | Required | logic changes, API modifications |
| 4 | Manual required | Cannot auto | security, architecture |

### Quick Start Examples

```bash
# Single scan and fix (parallel by default)
> /moai:fix

# Autonomous loop until resolved
> /moai:loop --max 50

# Full autonomous development (parallel by default)
> /moai:alfred "Add JWT authentication" --loop
```

### Why "All is Well"?

The name reflects the philosophy: **set the goal and trust the AI**. MoAI-ADK's agentic automation handles the entire development cycle autonomously. When you see `<moai>DONE</moai>`, all is truly well.

---

## 5. Mr.Alfred and Sub-Agents

### 🎩 Mr.Alfred - Super Agent (Chief Orchestrator)

**Role**: Analyze user requests and delegate to appropriate specialized agents

**Workflow**:

1. **Understand**: Request analysis and clarification
2. **Plan**: Establish execution plan via Plan agent
3. **Execute**: Delegate tasks to specialized agents (sequential/parallel)
4. **Integrate**: Integrate results and report to user

### 🌐 Multilingual Automatic Routing (NEW)

Alfred automatically recognizes 4 language requests and invokes the correct agent:

| Request Language | Example                     | Agent Invoked  |
| ---------------- | --------------------------- | -------------- |
| English          | "Design backend API"        | expert-backend |
| Korean           | "백엔드 API 설계해줘"       | expert-backend |
| Japanese         | "バックエンドAPIを設計して" | expert-backend |
| Chinese          | "设计后端API"               | expert-backend |

---

### 🔧 Tier 1: Domain Experts (8)

| Agent                  | Expertise                  | Usage Example                            |
| ---------------------- | -------------------------- | ---------------------------------------- |
| **expert-backend**     | FastAPI, Django, DB design | API design, query optimization           |
| **expert-frontend**    | React, Vue, Next.js        | UI components, state management          |
| **expert-security**    | Security analysis, OWASP   | Security audit, vulnerability analysis   |
| **expert-devops**      | Docker, K8s, CI/CD         | Deployment automation, infrastructure    |
| **expert-debug**       | Bug analysis, performance  | Problem diagnosis, bottleneck resolution |
| **expert-performance** | Profiling, optimization    | Response time improvement                |
| **expert-refactoring** | Code refactoring, AST-Grep | Large-scale code transformation          |
| **expert-testing**     | Test strategy, E2E         | Test planning, coverage                  |

---

### 🎯 Tier 2: Workflow Managers (8)

| Agent                   | Role                             | Auto Invocation Timing          |
| ----------------------- | -------------------------------- | ------------------------------- |
| **manager-spec**        | SPEC writing (EARS)              | `/moai:1-plan`                  |
| **manager-tdd**         | TDD auto execution               | `/moai:2-run`                   |
| **manager-docs**        | Documentation auto generation    | `/moai:3-sync`                  |
| **manager-quality**     | TRUST 5 verification             | After implementation completion |
| **manager-strategy**    | Execution strategy establishment | Complex planning                |
| **manager-project**     | Project initialization           | `/moai:0-project`               |
| **manager-git**         | Git workflow                     | Branch/PR management            |
| **manager-claude-code** | Claude Code integration          | Settings optimization           |

---

### 🏗️ Tier 3: Claude Code Builder (4)

| Agent               | Role               | Usage Example                 |
| ------------------- | ------------------ | ----------------------------- |
| **builder-agent**   | Create new agent   | Organization specialist agent |
| **builder-skill**   | Create new skill   | Team-specific skill module    |
| **builder-command** | Create new command | Custom workflow               |
| **builder-plugin**  | Create plugin      | Distribution plugin           |

---

## 6. Agent-Skills

### 📚 Skill Library Structure

```text
🏗️ Foundation (5)    → Core philosophy, execution rules
🎯 Domain (4)        → Domain expertise
💻 Language (16)     → 16 programming languages
🚀 Platform (10)     → Cloud/BaaS integration
📋 Workflow (7)      → Automation workflows
📚 Library (4)       → Special libraries
🛠️ Tool (2)          → Development tools
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

# Method 2: Alfred auto selection (general user)
"FastAPI server 만들어줘"
→ Alfred automatically selects moai-lang-python
```

---

## 7. TRUST 5 Quality Principles

All MoAI-ADK projects follow the **TRUST 5** quality framework.

### 🏆 TRUST 5 = Test + Readable + Unified + Secured + Trackable

```mermaid
graph TD
    T1["🔴 T: Test-First<br/>━━━━━━━━<br/>• TDD Red-Green-Refactor<br/>• 85%+ coverage<br/>• Automated testing"]
    R["📖 R: Readable<br/>━━━━━━━━<br/>• Clear naming<br/>• Code comments<br/>• Linter compliance"]
    U["🔄 U: Unified<br/>━━━━━━━━<br/>• Consistent style<br/>• Standard patterns<br/>• Error handling"]
    S["🔒 S: Secured<br/>━━━━━━━━<br/>• OWASP Top 10<br/>• Vulnerability scan<br/>• Encryption policy"]
    T2["📋 T: Trackable<br/>━━━━━━━━<br/>• Clear commits<br/>• Issue tracking<br/>• CHANGELOG"]

    T1 --> R --> U --> S --> T2 --> Deploy["✅ Production Ready"]
```

### T - Test-First

**Principle**: All implementation starts with tests

**Verification**:

- Test coverage >= 85%
- Write failing tests first (Red)
- Pass tests with code (Green)
- Refactor

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

## 8. TAG System v2.0 - Code-to-Documentation Traceability

TAG System v2.0 provides bidirectional traceability between code and SPEC documents. Use `@SPEC` tags to automatically link code to documentation and documentation to code.

### Core Problem Solved: "What if I write code without a SPEC first?"

TAG System v2.0's **Context-Aware Validation** perfectly solves this:

| Situation              | TAG Type                  | Commit without SPEC? | Description                      |
| ---------------------- | ------------------------- | -------------------- | -------------------------------- |
| **Simple fixes**       | `related` (all files)     | ✅ Allowed           | Shows relationship only         |
| **Write tests first**  | `verify` or `@TEST` (test files) | ✅ Allowed | Supports TDD workflow        |
| **Implementation code**| `impl` (implementation files) | ⚠️ Warning/block    | Depends on warn/enforce mode  |
| **Show dependencies**  | `depends` (all files)     | ✅ Allowed           | Shows dependency only          |

### TAG Overview

MoAI-ADK supports two TAG types:

| TAG Type | Purpose         | Format                    | Location     |
| -------- | --------------- | ------------------------- | ------------ |
| **@SPEC** | Code-to-SPEC link| `@SPEC SPEC-XXX [verb]` | All files    |
| **@TEST** | Test ID         | `@TEST TEST-XXX`          | Test files only |

### TAG Usage Guide

#### 1. @SPEC TAG Verb Selection Guide

| Verb      | When to use?              | Usage Example                  |
| --------- | ------------------------- | ------------------------------ |
| **impl**  | Implementation code links | `# @SPEC SPEC-AUTH-001 impl`    |
| **verify**| Verify SPEC in tests      | `# @SPEC SPEC-AUTH-001 verify`  |
| **related**| Simple fixes/relationships | `# @SPEC SPEC-AUTH-001 related` |
| **depends**| Dependency on other SPEC  | `# @SPEC SPEC-DB-004 depends`   |

#### 2. @TEST vs @SPEC verify Selection Guide

| Distinction | @TEST                    | @SPEC verify               |
| ----------- | ------------------------ | -------------------------- |
| **Purpose** | Test identification      | SPEC verification          |
| **SPEC link**| Optional                 | Required                   |
| **Usage**   | Independent tests        | TDD (SPEC → impl → test)  |
| **Combined** | `@SPEC verify` + `@TEST` | Standalone                 |

#### 3. Recommended Usage Patterns

```python
# Pattern 1: TDD (SPEC exists)
# @SPEC SPEC-AUTH-001 verify
# @TEST TEST-AUTH-001
def test_login():
    pass

# Pattern 2: Independent test (no SPEC)
# @TEST TEST-AUTH-001
def test_login():
    pass

# Pattern 3: Implementation code
# @SPEC SPEC-AUTH-001 impl
def login():
    pass
```

### Summary: TAG Selection Decision Tree

```
Is this a test file?
├── Yes
│   ├── SPEC exists and needs verification? → @SPEC verify
│   ├── Independent test identification? → @TEST
│   └── Need both? → @SPEC verify + @TEST
└── No (implementation code)
    ├── SPEC exists and implementation? → @SPEC impl
    ├── Simple fix? → @SPEC related
    └── Depends on other SPEC? → @SPEC depends
```

### TAG Syntax

Add @TAG in code comments. Recommended to place above the function:

```python
# Basic format
@SPEC SPEC-{DOMAIN}-{NUMBER} [verb] # General code
@TEST TEST-{DOMAIN}-{NUMBER}       # Test code

# Example
# @SPEC SPEC-AUTH-001 impl
def login(): ...

# @SPEC SPEC-AUTH-001 verify
# @TEST TEST-AUTH-001
def test_login(): ...
```

### Real-World Workflow

#### Scenario 1: Simple Bug Fix (no SPEC)

```python
# auth.py
# @SPEC SPEC-AUTH-001 related
def fix_login_bug():
    """Login bug fix

    No related SPEC exists, using related
    """
    pass
```

**Pre-commit Hook Result**:
```
TAG Validation Results:
==================================================
  auth.py:10: Hint: Consider creating SPEC for SPEC-AUTH-001
  (related TAG - SPEC not required)
==================================================

Commit allowed with warning (warn mode)
```

✅ **Commit Allowed**: related TAG allows commit without SPEC

#### Scenario 2: Write Tests First (TDD)

```python
# test_auth.py
# @SPEC SPEC-AUTH-002 verify
# @TEST TEST-AUTH-002
def test_login_success():
    """Login success test

    Tests written first, impl later
    """
    pass
```

**Pre-commit Hook Result**:
```
TAG Validation Results:
==================================================
  test_auth.py:15: Warning: Test references non-existent SPEC: SPEC-AUTH-002
  (verify TAG in test file - commit allowed)
==================================================

Commit allowed with warning (warn mode)
```

✅ **Commit Allowed**: verify TAG in test files allows commit without SPEC

#### Scenario 3: Official Development (SPEC first)

```bash
# 1. Create SPEC first
> /moai:1-plan "user authentication system"
# → SPEC-AUTH-003 created

# 2. Write implementation code
# @SPEC SPEC-AUTH-003 impl
def login(username, password):
    """User login

    Official development uses impl TAG
    """
    pass

# 3. Write test code
# @SPEC SPEC-AUTH-003 verify
# @TEST TEST-AUTH-003
def test_login():
    """Login test"""
    pass
```

**Pre-commit Hook Result** (SPEC exists):
```
TAG Validation: All TAGs valid
```

✅ **Commit Allowed**: SPEC existence confirmed

**Pre-commit Hook Result** (SPEC doesn't exist, enforce mode):
```
TAG Validation Results:
==================================================
  auth.py:25: Error: Implementation references non-existent SPEC: SPEC-AUTH-003
  (commit blocked in enforce mode)
==================================================

Commit blocked due to TAG validation error (enforce mode)
```

❌ **Commit Blocked**: enforce mode blocks impl TAG without SPEC

### Configuration Options

Configure TAG validation behavior in `.moai/config/sections/quality.yaml`:

```yaml
tag_validation:
  enabled: true           # Enable TAG checking
  mode: warn              # warn | enforce | off
  check_spec_exists: true # Check SPEC document exists
  max_tags_per_file: 100  # Maximum TAGs per file
```

### Mode-Specific Behavior

| Mode            | impl TAG (no SPEC) | verify/@TEST TAG (test file) | related TAG | depends TAG |
| --------------- | ------------------ | ----------------------------- | ----------- | ----------- |
| **warn** (default) | Warn then allow     | Warn then allow               | Hint only   | Hint only   |
| **enforce**      | Block commit        | Allow                         | Hint only   | Hint only   |
| **off**          | No check            | No check                      | No check    | No check    |

### Recommended TAG Usage Strategy

| Development Type | Recommended Strategy                       | TAG Example                                                                            |
| ---------------- | ------------------------------------------- | ----------------------------------------------------------------------------------- |
| **Hotfix**       | Use `related`                               | `@SPEC SPEC-XXX related`                                                            |
| **Test-first**   | `verify` or `@TEST` first, then `impl`      | Tests: `@SPEC SPEC-XXX verify` or `@TEST TEST-XXX` → Impl: `@SPEC SPEC-XXX impl` |
| **Official dev** | Create SPEC first, use `impl`              | `/moai:1-plan` → `@SPEC SPEC-XXX impl`                                              |
| **Refactoring**  | Keep existing TAGs                          | No changes                                                                          |
| **Documentation** | Use `related` to show relationship       | `@SPEC SPEC-XXX related`                                                            |
| **Dependency**    | Use `depends` to show dependency            | `@SPEC SPEC-YYY depends` (depends on other SPEC)                                    |

### Pre-commit Hook Automatic Validation

All Git commits automatically validate TAG:

```bash
# Attempt commit
git commit -m "Add login feature"

# Automatic execution
TAG Validation Results:
==================================================
  auth.py:10: Hint: Consider creating SPEC for SPEC-AUTH-001
  (related TAG - SPEC not required)
==================================================

Commit allowed with warning (warn mode)
```

### TAG Linkage Database

TAG System stores bidirectional mapping in `.moai/cache/tag-linkage.json`:

```json
{
  "tags": [
    {
      "spec_id": "SPEC-AUTH-001",
      "verb": "impl",
      "file_path": "src/auth.py",
      "line": 10
    }
  ],
  "files": {
    "src/auth.py": ["SPEC-AUTH-001"]
  }
}
```

**Auto-update**:
- Automatically extracts and saves TAGs on commit
- Automatically removes TAGs when files are deleted
- Detects orphaned TAGs (references to deleted SPECs)

### Advanced Features

#### SPEC-ID Format Validation

```python
# Correct format
@SPEC SPEC-AUTH-001    # ✅
@SPEC SPEC-PAY-002     # ✅
@SPEC SPEC-UI-003      # ✅

# Incorrect format
@SPEC AUTH-001         # ❌ SPEC- prefix missing
@SPEC SPEC-001         # ❌ Domain missing
@SPEC SPEC-AUTH-1      # ❌ 3-digit number required
```

#### File Type Auto-Detection

TAG System auto-detects by file pattern:
- **Test files**: `test_*.py`, `*_test.py`, `tests/` directory
- **Implementation files**: All other Python files

### Getting Started

```bash
# 1. TAG System already enabled
cat .moai/config/sections/quality.yaml

# 2. Add TAGs to your code
# auth.py (implementation file)
# @SPEC SPEC-AUTH-001 impl
def login():
    pass

# test_auth.py (test file)
# @SPEC SPEC-AUTH-001 verify
# @TEST TEST-AUTH-001
def test_login():
    pass

# 3. Automatic validation on commit
git add .
git commit -m "Add login feature"
# → TAG Validation runs automatically
```

### Troubleshooting

**Q: Skip TAG validation**
```bash
git commit --no-verify
```

**Q: Temporarily commit in enforce mode**
```yaml
# quality.yaml (temporary change)
tag_validation:
  mode: warn  # enforce → warn
```

**Q: Find orphaned TAGs**
```bash
# Orphaned TAGs = TAGs referencing non-existent SPECs
# Automatically detected by Pre-commit Hook
```

### Core Innovations of TAG System v2.0

| Problem                  | TAG System v1     | TAG System v2                          |
| ------------------------ | ----------------- | -------------------------------------- |
| Write code without SPEC  | ❌ Impossible     | ✅ Possible with `related`/`verify`     |
| Write tests first        | ❌ Impossible     | ✅ Test file `verify`/`@TEST` allowed   |
| Flexible validation      | ❌ All or nothing  | ✅ warn/enforce/off modes               |
| File type awareness      | ❌ None            | ✅ Auto-detects (test/implementation)   |
| Show dependencies        | ❌ Not supported   | ✅ `depends` shows SPEC dependencies    |
| Test identification TAG  | ❌ Not supported   | ✅ `@TEST` for independent test IDs       |

**TAG System v2.0**: Maintains SPEC-First TDD benefits while perfectly supporting real-world development flexibility.

---

## 10. Automated Quality Checks

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

---

## 11. 📊 Statusline Customization

MoAI-ADK provides a **customizable statusline** that displays real-time project information in Claude Code.

### Default Statusline Layout

```text
🤖 Opus 4.5 | 💰 152K/200K | 💬 Mr. Alfred | 📁 MoAI-ADK | 📊 +0 M58 ?5 | 💾 57.7MB | 🔀 main
```

### Statusline Components

| Icon | Component | Description | Config Key |
|------|-----------|-------------|------------|
| 🤖 | **Model** | Claude model (Opus, Sonnet, etc.) | `model` |
| 💰 | **Context** | Context window usage (e.g., 77K/200K) | `context_window` |
| 💬 | **Style** | Active output style (e.g., Mr. Alfred) | `output_style` |
| 📁 | **Directory** | Current project name | `directory` |
| 📊 | **Git Status** | Staged/Modified/Untracked files | `git_status` |
| 💾 | **Memory** | Process memory usage | `memory_usage` |
| 🔀 | **Branch** | Current Git branch | `branch` |
| 🔅 | **Version** | Claude Code version (optional) | `version` |

### Configuration

Edit `.moai/config/statusline-config.yaml`:

```yaml
display:
  model: true           # 🤖 Claude model
  context_window: true  # 💰 Context window
  output_style: true    # 💬 Output style
  directory: true       # 📁 Project name
  git_status: true      # 📊 Git status
  memory_usage: true    # 💾 Memory usage
  branch: true          # 🔀 Git branch
  version: 1.1.0        # 🔅 Version (optional)
  active_task: true     # Active task
```

### Memory Collector

When `memory_usage` is enabled, MoAI-ADK uses `psutil` to collect real-time memory usage:

- **Process Memory**: RSS (Resident Set Size) of the current Python process
- **Caching**: Performance optimization with 10-second TTL
- **Cross-Platform**: Supports macOS, Linux, Windows
- **Graceful Degradation**: Displays "N/A" when psutil is unavailable

### Display Modes

| Mode | Max Length | Use Case |
|------|------------|----------|
| `compact` | 80 chars | Standard terminal |
| `extended` | 120 chars | Wide terminal |
| `minimal` | 40 chars | Narrow terminal |

Set mode:

```bash
export MOAI_STATUSLINE_MODE=extended
```

---

## 12. 🌳 Worktree Parallel Development

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
│  > /moai:1-plan "사용자 인증 시스템" --worktree                   │
│  ✅ SPEC-AUTH-001 생성 완료                                      │
│  ✅ Worktree 생성: ~/moai/worktrees/my-project/SPEC-AUTH-001     │
│  ✅ Branch: feature/SPEC-AUTH-001                                │
│                                                                  │
│  > /moai:1-plan "결제 시스템" --worktree                          │
│  ✅ SPEC-PAY-002 생성 완료                                       │
│  ✅ Worktree 생성: ~/moai/worktrees/my-project/SPEC-PAY-002      │
│                                                                  │
│  > /moai:1-plan "대시보드 UI" --worktree                         │
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
│  > /moai:2-run SPEC-AUTH-001                                     │
│  🔄 TDD 실행 중... (Red → Green → Refactor)                       │
│  ✅ 구현 완료!                                                   │
│  ✅ 테스트 통과 (Coverage: 92%)                                  │
│                                                                  │
│  > /moai:3-sync SPEC-AUTH-001                                    │
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
│  > /moai:alfred SPEC-PAY-002                                     │
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
│  > /moai:alfred SPEC-UI-003                                      │
│  ✅ 완료!                                                        │
└─────────────────────────────────────────────────────────────────┘
```

---

### 🎯 Core Workflow

#### Phase 1: Plan with Claude 4.5 Opus (Terminal 1)

```bash
/moai:1-plan "feature description" --worktree
```

- ✅ SPEC document generation
- ✅ Worktree auto creation
- ✅ Feature branch auto creation

#### Phase 2: Implement with GLM 4.7 (Terminals 2, 3, 4...)

```bash
moai-wt go SPEC-ID
moai glm
claude
> /moai:2-run SPEC-ID
> /moai:3-sync SPEC-ID
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

| Command                  | Description                        | Usage Example                    |
| ------------------------ | ---------------------------------- | -------------------------------- |
| `moai-wt new SPEC-ID`    | Create new Worktree                | `moai-wt new SPEC-AUTH-001`      |
| `moai-wt go SPEC-ID`     | Enter Worktree (open new shell)    | `moai-wt go SPEC-AUTH-001`       |
| `moai-wt list`           | List Worktrees                     | `moai-wt list`                   |
| `moai-wt done SPEC-ID`   | Merge and cleanup (checkout→merge) | `moai-wt done SPEC-AUTH-001`     |
| `moai-wt remove SPEC-ID` | Remove Worktree                    | `moai-wt remove SPEC-AUTH-001`   |
| `moai-wt status`         | Check Worktree status and registry | `moai-wt status`                 |
| `moai-wt sync [SPEC-ID]` | Synchronize Worktree               | `moai-wt sync --all`             |
| `moai-wt clean`          | Clean merged Worktrees             | `moai-wt clean --merged-only`    |
| `moai-wt recover`        | Recover registry from disk         | `moai-wt recover`                |
| `moai-wt config`         | Check Worktree settings            | `moai-wt config root`            |

---

## 13. Understanding CLAUDE.md

The `CLAUDE.md` file generated in your project root after MoAI-ADK installation is **Alfred's (AI Orchestrator) execution directive**. This file defines how Claude Code behaves in your project.

### What is CLAUDE.md?

`CLAUDE.md` is a project configuration file that Claude Code automatically reads at session start. In MoAI-ADK, this file defines the behavioral rules for the **Alfred orchestrator**.

```text
📁 Project Root
├── CLAUDE.md              ← Alfred execution directive (don't modify)
├── CLAUDE.local.md        ← Personal custom directives (optional)
├── .claude/
│   ├── settings.json      ← Claude Code settings
│   ├── agents/            ← Sub-agent definitions
│   ├── commands/          ← Slash commands
│   └── skills/            ← Skill definitions
└── .moai/
    └── config/            ← MoAI configuration
```

### CLAUDE.md Core Structure

| Section | Description | Key Content |
|---------|-------------|-------------|
| **Core Identity** | Alfred's role definition | Strategic orchestrator, HARD rules |
| **Request Processing Pipeline** | Request handling flow | Analyze → Route → Execute → Report |
| **Command Reference** | Command classification | Type A (Workflow), Type B (Utility), Type C (Feedback) |
| **Agent Catalog** | Sub-agent listing | 8 Managers, 8 Experts, 4 Builders |
| **SPEC-Based Workflow** | SPEC-driven development | Plan → Run → Sync flow |
| **Quality Gates** | Quality validation rules | HARD/SOFT rules checklist |
| **Configuration Reference** | Settings reference | Language, output format rules |

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
- Test files must have test_ prefix

## Prohibited
- No console.log (use logger)
- No any type usage
```

**Benefits**:
- No conflicts with `CLAUDE.md` updates
- Project-specific customization
- Can add to `.gitignore` for personal settings

### CLAUDE.md vs CLAUDE.local.md

| Aspect | CLAUDE.md | CLAUDE.local.md |
|--------|-----------|-----------------|
| **Purpose** | Alfred execution directives | Personal/project additional directives |
| **Modification** | Not recommended | Freely modifiable |
| **Updates** | Managed by MoAI | Managed by user |
| **Git** | Committed | Optional (.gitignore possible) |
| **Priority** | Base rules | Additional/override rules |

### Core Rules (HARD Rules)

**HARD rules** defined in `CLAUDE.md` are always enforced:

1. **Language-Aware Responses**: Respond in user's language
2. **Parallel Execution**: Execute independent tasks in parallel
3. **No XML in User Responses**: No XML tags in user-facing responses

These rules cannot be overridden even by `CLAUDE.local.md`.

---

## 14. MoAI Rank Introduction

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
❯ moai rank register

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

## 15. FAQ 7 Questions

### Q1: Is SPEC Always Required?

| Condition            | SPEC Required       |
| -------------------- | ------------------- |
| 1-2 files modified   | Optional (can skip) |
| 3-5 files modified   | Recommended         |
| 10+ files modified   | Required            |
| New feature addition | Recommended         |
| Bug fix              | Optional            |

### Q2: What TAG Should I Use When Writing Code Without a SPEC?

TAG System v2.0 supports flexible development:

| Situation              | Recommended TAG      | SPEC Required? |
| ---------------------- | -------------------- | -------------- |
| Hotfix                 | `@SPEC SPEC-XXX related` | No            |
| Write tests first      | `@TEST TEST-XXX` or `@SPEC SPEC-XXX verify` | No |
| Implementation code    | `@SPEC SPEC-XXX impl` | Yes (in enforce mode) |
| Show dependency        | `@SPEC SPEC-XXX depends` | No            |

### Q3: Is MCP Server Installation Required?

**Required (2)**:

- **Context7**: For latest library documentation and Skill reference generation

**Optional**:

- claude-in-chrome: Use Claude in browser and web automation testing
- Playwright: Web automation testing
- Figma: Design system

### Q4: What's the Difference Between @TEST and @SPEC verify?

| Distinction | @TEST                    | @SPEC verify               |
| ----------- | ------------------------ | -------------------------- |
| **Purpose** | Test identification      | SPEC verification          |
| **SPEC link**| Optional                 | Required                   |
| **Usage**   | Independent tests        | TDD (SPEC → impl → test)  |
| **Combined** | `@SPEC verify` + `@TEST` | Standalone                 |

### Q5: Does MoAI Rank Cost Money?

It's free. Only automatically collects session data.

### Q6: Is GLM Configuration Required?

No. You can use Claude only. However, it's recommended for cost savings.

### Q7: Can It Be Applied to Existing Projects?

Yes. `moai init .` preserves existing files.

---

## 16. Community & Support

### 🌐 Participate

- **Discord (Official)**: [https://discord.gg/umywNygN](https://discord.gg/umywNygN)
- **GitHub**: [https://github.com/modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **Developer Blog**: [https://goos.kim](https://goos.kim)

### 🆘 Support

- Email: [support@mo.ai.kr](mailto:support@mo.ai.kr)
- Documentation: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)

---

## 📝 License

Copyleft License (COPYLEFT-3.0) - [LICENSE](./LICENSE)

---

## 🙏 Made with ❤️ by MoAI-ADK Team

**Last Updated:** 2026-01-13
**Philosophy**: SPEC-First TDD + Agent Orchestration + Hybrid LLM
**MoAI**: MoAI means "Modu-ui AI" (AI for Everyone).

> **"Infinite Possibilism - AI for All"**
