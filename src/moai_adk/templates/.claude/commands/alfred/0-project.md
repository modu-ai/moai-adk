---
name: alfred:0-project
description: "Initialize project metadata and documentation"
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(ls:*)
  - Bash(find:*)
  - Bash(cat:*)
  - Task
---

# 📋 MoAI-ADK Step 0: Initialize/Update Universal Language Support Project Documentation

> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.

## 🎯 Command Purpose

Automatically analyzes the project environment to create/update product/structure/tech.md documents and configure language-specific optimization settings.

## 📋 Execution flow

0. **Conversation Language Selection**: User selects the language for all dialogs and documentation
1. **Environment Analysis**: Automatically detect project type (new/legacy) and codebase language
2. **Establishment of interview strategy**: Select question tree suited to project characteristics
3. **User Verification**: Review and approve interview plan
4. **Create project documentation**: Create product/structure/tech.md in the selected language
5. **Create configuration file**: config.json auto-configuration

## 🧠 Associated Skills & Agents

| Agent           | Core Skill                       | Purpose                                       |
| --------------- | -------------------------------- | --------------------------------------------- |
| project-manager | `moai-alfred-language-detection` | Initialize project and interview requirements |
| trust-checker   | `moai-alfred-trust-validation`   | Verify initial project structure (optional)   |

**Note**: TUI Survey Skill is used for user confirmations during project initialization and is shared across all interactive prompts.

## 🔗 Associated Agent

- **Primary**: project-manager (📋 planner) - Dedicated to project initialization
- **Quality Check**: trust-checker (✅ Quality assurance lead) - Initial structural verification (optional)
- **Secondary**: None (standalone execution)

## 💡 Example of use

The user executes the `/alfred:0-project` command to analyze the project and create/update documents.

## Command Overview

It is a systematic initialization system that analyzes the project environment and creates/updates product/structure/tech.md documents.

- **Automatically detect language**: Automatically recognize Python, TypeScript, Java, Go, Rust, etc.
- **Project type classification**: Automatically determine new vs. existing projects
- **High-performance initialization**: Achieve 0.18 second initialization with TypeScript-based CLI
- **2-step workflow**: 1) Analysis and planning → 2) Execution after user approval

## How to use

The user executes the `/alfred:0-project` command to start analyzing the project and creating/updating documents.

**Automatic processing**:

- Update mode if there is an existing `.moai/project/` document
- New creation mode if there is no document
- Automatic detection of language and project type

## ⚠️ Prohibitions

**What you should never do**:

- ❌ Create a file in the `.claude/memory/` directory
- ❌ Create a file `.claude/commands/alfred/*.json`
- ❌ Unnecessary overwriting of existing documents
- ❌ Date and numerical prediction ("within 3 months", "50% reduction") etc.)
- ❌ Hypothetical scenarios, expected market size, future technology trend predictions

**Expressions to use**:

- ✅ "High/medium/low priority"
- ✅ "Immediately needed", "step-by-step improvements"
- ✅ Current facts
- ✅ Existing technology stack
- ✅ Real problems

---

## 🚀 STEP 0: 초기 설정 - 언어 및 사용자 정보 선택

**목적**: 프로젝트 초기화 시작 전에 대화 언어를 설정하고 사용자 닉네임을 등록합니다. 이 설정은 모든 Alfred 프롬프트, 인터뷰 질문 및 생성된 문서에 적용됩니다.

**UX 개선**: 2개 질문을 **1회 배치 호출**로 통합 (50% 상호작용 감소: 2 turns → 1 turn)

### 0.0 Alfred 자기소개 및 환영 인사

Alfred가 첫 상호작용으로 다음과 같이 인사합니다:

```
안녕하세요! 👋 저는 Alfred입니다.
MoAI-ADK의 SuperAgent로서 당신의 프로젝트를 함께 만들어갈 준비가 되어 있습니다.

앞으로의 모든 대화에서 당신을 편하게 부르기 위해,
먼저 기본 설정을 진행하겠습니다.
```

### 0.1 배치 설계: 언어 선택 + 사용자 닉네임 + GitHub 설정 확인 (1-3회 호출)

Alfred가 `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` 를 사용하여 **배치 호출**로 필수 정보를 수집합니다:

**기본 배치 (항상 실행)**:

- 언어 선택
- 사용자 닉네임

**추가 배치 (팀 모드 감지 시)**:

- GitHub "Automatically delete head branches" 설정 확인

#### 0.1.1 팀 모드 감지

```bash
# config.json에서 mode 확인
grep "mode" .moai/config.json

# 결과: "mode": "team" → 추가 질문 포함
#      "mode": "personal" → 기본 질문만 실행
```

#### 0.1.2 기본 배치: 언어 선택 + 에이전트 프롬프트 언어 + 닉네임 (3개 질문, 1회 배치 호출)

**배치 설계**: 3개 질문을 1회 호출로 통합 (UX 개선: 3 turns → 1 turn)

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Which language would you like to use for the project initialization and documentation?",
            "header": "Language",
            "multiSelect": false,
            "options": [
                {
                    "label": "🌍 English",
                    "description": "All dialogs and documentation in English"
                },
                {
                    "label": "🇰🇷 한국어",
                    "description": "All dialogs and documentation in Korean"
                },
                {
                    "label": "🇯🇵 日本語",
                    "description": "All dialogs and documentation in Japanese"
                },
                {
                    "label": "🇨🇳 中文",
                    "description": "All dialogs and documentation in Chinese"
                }
            ]
        },
        {
            "question": "In which language should Alfred's sub-agent prompts be written?",
            "header": "Agent Prompt Language",
            "multiSelect": false,
            "options": [
                {
                    "label": "🌐 English (Global Standard)",
                    "description": "All sub-agent prompts in English for global consistency and team collaboration. Recommended for Claude Pro $20 users: reduces token usage by ~15-20%, lowering API costs"
                },
                {
                    "label": "🗣️ Selected Language (Localized)",
                    "description": "All sub-agent prompts in the language you selected above for local team efficiency"
                }
            ]
        },
        {
            "question": "How would you like to be called in our conversations? (e.g., GOOS, Team Lead, Developer, or custom name - max 20 chars)",
            "header": "Nickname",
            "multiSelect": false,
            "options": [
                {
                    "label": "Enter custom nickname",
                    "description": "Type your preferred name using the 'Other' option below"
                }
            ]
        }
    ]
)
```

**응답 처리**:

**Q1 (사용자 언어)**:

- Selected option stored as: `conversation_language: "ko"` (or "en", "ja", "zh", etc.)

**Q2 (에이전트 프롬프트 언어)** - **NEW**:

- **"English (Global Standard)"** → `agent_prompt_language: "english"`
  - All sub-agent prompts written in English
  - Recommended for global teams, code consistency, and international collaboration
  - Project-manager, spec-builder, code-builder use English prompts internally
- **"Selected Language (Localized)"** → `agent_prompt_language: "localized"`
  - All sub-agent prompts written in the user-selected language
  - Recommended for local teams, local documentation, and native language efficiency
  - Project-manager receives prompts in selected language (e.g., Korean, Japanese)

**Q3 (사용자 닉네임)**:

- Custom nickname stored as: `user.nickname: "GOOS"` (or custom input)

#### 0.1.3 팀 모드 추가 배치: GitHub 설정 & Git 워크플로우 선택 (팀 모드만)

**조건**: `config.json`에서 `"mode": "team"` 감지 시 실행

**배치 구성**: 2개 질문 (1회 호출로 통합)

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "[Team Mode] Is 'Automatically delete head branches' enabled in your GitHub repository settings?",
            "header": "GitHub Branch Settings",
            "multiSelect": false,
            "options": [
                {
                    "label": "✅ Yes, already enabled",
                    "description": "PR merge 후 자동으로 원격 브랜치 삭제됨"
                },
                {
                    "label": "❌ No, not enabled (Recommended: Enable)",
                    "description": "Settings → General → '자동 삭제' 체크박스 확인 필요"
                },
                {
                    "label": "🤔 Not sure / Need to check",
                    "description": "GitHub Settings → General 확인 후 다시 진행"
                }
            ]
        },
        {
            "question": "[Team Mode] Which Git workflow should we use when creating SPEC documents?",
            "header": "SPEC Git Workflow",
            "multiSelect": false,
            "options": [
                {
                    "label": "📋 Feature Branch + PR",
                    "description": "매 SPEC마다 feature 브랜치 생성 → PR 리뷰 → develop 병합. 팀 협업과 코드 리뷰에 최적"
                },
                {
                    "label": "🔄 Direct Commit to Develop",
                    "description": "브랜치 생성 없이 develop에 직접 커밋. 빠른 프로토타이핑과 단순 워크플로우에 최적"
                },
                {
                    "label": "🤔 Decide per SPEC",
                    "description": "SPEC 생성 시마다 매번 선택. 유연성이 높지만 매번 결정 필요"
                }
            ]
        }
    ]
)
```

**응답 처리**:

**Q1 (GitHub 설정)**:

- **"Yes, already enabled"** → `auto_delete_branches: true` 저장
- **"No, not enabled"** → `auto_delete_branches: false` + 권장사항 저장
- **"Not sure"** → `auto_delete_branches: null` + 경고 메시지

**Q2 (Git 워크플로우)**:

- **"Feature Branch + PR"** → `spec_git_workflow: "feature_branch"` 저장
  - `/alfred:1-plan` 실행 시 자동으로 feature 브랜치 생성
  - git-manager가 PR 기반 워크플로우 적용
- **"Direct Commit to Develop"** → `spec_git_workflow: "develop_direct"` 저장
  - `/alfred:1-plan` 실행 시 develop 브랜치에 직접 커밋
  - 브랜치 생성 과정 생략
- **"Decide per SPEC"** → `spec_git_workflow: "per_spec"` 저장
  - `/alfred:1-plan` 실행 시마다 git-manager가 사용자에게 선택 요청

**User Response Example**:

```
Selected Language: 🇰🇷 한국어
Selected Nickname: GOOS (typed via "Other" option)
```

---

### 0.1.4 Domain Selection (Optional - All Modes)

**Purpose**: Identify project domains to activate domain-expert agents for specialized guidance.

**When to ask**: After language/nickname/GitHub settings complete

**Batched Design**: Domain selection integrated into initial batch OR asked separately based on user preference

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Which domains does your project involve? (Select all that apply)",
            "header": "Project Domains",
            "multiSelect": true,
            "options": [
                {
                    "label": "🎨 Frontend",
                    "description": "React, Vue, Angular, Next.js, Nuxt, SvelteKit, Astro, Remix, SolidJS"
                },
                {
                    "label": "⚙️ Backend",
                    "description": "FastAPI, Flask, Django, Express, Fastify, NestJS, Spring Boot, Gin, Axum"
                },
                {
                    "label": "🚀 DevOps",
                    "description": "Railway, Vercel, Docker, Kubernetes, AWS, GCP, Azure, CI/CD"
                },
                {
                    "label": "🗄️ Database",
                    "description": "PostgreSQL, MySQL, MongoDB, Redis, database design and optimization"
                },
                {
                    "label": "📊 Data Science",
                    "description": "Data analysis, machine learning, data pipelines, notebooks"
                },
                {
                    "label": "📱 Mobile",
                    "description": "React Native, Flutter, iOS, Android app development"
                },
                {
                    "label": "⚡ Skip",
                    "description": "No domain selection (can add later via /alfred:1-plan)"
                }
            ]
        }
    ]
)
```

**Response Processing**:

When user selects domains, Alfred processes the response as follows:

**Selected Domain Processing** (`answers["0"]` contains selected domain labels):

- Extract selected domain codes from labels: "Frontend" → "frontend", "Backend" → "backend", etc.
- Store selected domains in `.moai/config.json`:
  ```json
  {
    "stack": {
      "selected_domains": ["frontend", "backend"],
      "domain_selection_date": "2025-10-23T12:34:56Z"
    }
  }
  ```

**Skip Domain Selection** (if user selects "⚡ Skip"):

- Store in config.json:
  ```json
  {
    "stack": {
      "selected_domains": [],
      "domain_selection_skipped": true,
      "domain_selection_date": "2025-10-23T12:34:56Z"
    }
  }
  ```
- Display: "✅ Domain selection skipped. You can add domains later during `/alfred:1-plan`"

**Domain Expert Activation**:

- Selected domains stored in `.moai/config.json`
- Domain-expert agents activated during `/alfred:1-plan` (automatic keyword detection)
- Domain-expert agents available as advisors during `/alfred:2-run`
- Domain-specific sync routing enabled in `/alfred:3-sync`
- If domains skipped: Default agent lineup used (can be customized later in `/alfred:1-plan`)

---

### 0.2 사용자 정보 저장

Alfred가 선택된 언어, 닉네임, 그리고 팀 모드 설정을 다음과 같이 저장합니다:

#### 0.2.1 기본 정보 저장 (항상) - 에이전트 프롬프트 언어 추가

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "localized",
    "agent_prompt_language_description": "All sub-agent prompts written in the selected language (localized)"
  },
  "user": {
    "nickname": "GOOS",
    "selected_at": "2025-10-23T12:34:56Z"
  },
  "stack": {
    "selected_domains": ["frontend", "backend"],
    "domain_selection_date": "2025-10-23T12:34:56Z"
  }
}
```

**에이전트 프롬프트 언어 옵션**:

- **`"english"`** (Global Standard) - **💰 Claude Pro $20 사용자 추천**:

  - All sub-agent prompts and internal communication in English
  - Best for: International teams, global collaboration, code consistency
  - Impact: Project-manager, spec-builder, code-builder all use English task prompts
  - **Cost Benefit**: Reduces token usage by ~15-20% compared to non-English prompts
    - English prompts are more efficient and use fewer tokens
    - Significant cost savings for continuous API usage
    - Example: 100,000 tokens in English ≈ 115,000-120,000 tokens in Korean/Japanese

- **`"localized"`** (Localized - Default for non-English):
  - All sub-agent prompts and internal communication in selected language
  - Best for: Local teams, native language efficiency, culturally-specific guidance
  - Impact: Project-manager, spec-builder, code-builder all use localized task prompts
  - Note: Uses ~15-20% more tokens due to language characteristics

#### 0.2.2 GitHub & Git 워크플로우 설정 저장 (팀 모드만)

**팀 모드 감지 시 추가 저장 - Feature Branch + PR 선택 시**:

```json
{
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch",
    "checked_at": "2025-10-23T12:34:56Z",
    "workflow_recommendation": "Feature branch를 사용한 PR 기반 협업 워크플로우. 매 SPEC마다 feature/spec-* 브랜치 생성, PR 리뷰 후 develop 병합"
  }
}
```

**또는 - Direct Commit to Develop 선택 시**:

```json
{
  "github": {
    "auto_delete_branches": false,
    "spec_git_workflow": "develop_direct",
    "checked_at": "2025-10-23T12:34:56Z",
    "workflow_recommendation": "develop 브랜치에 직접 커밋하는 단순 워크플로우. 브랜치 생성 과정 생략, 빠른 개발 속도"
  }
}
```

**또는 - Decide per SPEC 선택 시**:

```json
{
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "per_spec",
    "checked_at": "2025-10-23T12:34:56Z",
    "workflow_recommendation": "SPEC 생성 시마다 워크플로우 선택. /alfred:1-plan 실행 시 git-manager가 선택 요청"
  }
}
```

#### 0.2.3 저장된 정보 활용

이 정보는:

- 모든 sub-agents 에게 컨텍스트 파라미터로 전달됨
- `.moai/config.json` 의 `language`, `user`, `github` 필드에 저장됨
- CLAUDE.md의 `{{CONVERSATION_LANGUAGE}}` 및 `{{USER_NICKNAME}}` 변수로 치환됨
- 모든 Alfred 대화에서 사용됨
- **팀 모드**: git-manager가 다음 워크플로우를 자동으로 적용:
  - **`spec_git_workflow: "feature_branch"`**: `/alfred:1-plan` 실행 시 feature/spec-\* 브랜치 생성, PR 기반 리뷰 프로세스 적용
  - **`spec_git_workflow: "develop_direct"`**: `/alfred:1-plan` 실행 시 develop 브랜치에 직접 커밋, 브랜치 생성 과정 생략
  - **`spec_git_workflow: "per_spec"`**: `/alfred:1-plan` 실행 시마다 사용자에게 워크플로우 선택 요청

**설정 완료 출력 예시**:

```markdown
✅ 초기 설정 완료!

언어: 한국어 (ko)
닉네임: GOOS

이제 GOOS님의 프로젝트 환경 분석으로 진행하겠습니다...
```

### 0.3 STEP 1로 전환

언어 및 사용자 정보 설정 완료 후, 모든 후속 상호작용이 선택된 언어로 진행됩니다:

- Alfred의 모든 프롬프트가 선택된 언어로 번역됨
- project-manager sub-agent이 언어 및 사용자 정보 파라미터를 수신
- 인터뷰 질문이 선택된 언어로 진행됨
- 생성된 문서 (product.md, structure.md, tech.md)가 선택된 언어로 작성됨
- CLAUDE.md가 선택된 언어와 사용자 닉네임을 표시함

---

## 🚀 STEP 1: Environmental analysis and interview plan development

Analyze the project environment and develop a systematic interview plan.

### 1.0 Check backup directory (highest priority)

**Processing backup files after moai-adk init reinitialization**

Alfred first checks the `.moai-backups/` directory:

```bash
# Check latest backup timestamp
ls -t .moai-backups/ | head -1

# Check the optimized flag in config.json
grep "optimized" .moai/config.json
```

**Backup existence conditions**:

- `.moai-backups/` directory exists
- `.moai/project/*.md` file exists in the latest backup folder
- User's existing project files can be merged (regardless of optimized flag)

**Backup Detection Result**:

- **Backup Found**: Latest backup is `.moai-backups/[TIMESTAMP]/`
- **No Backup**: Proceed directly to Phase 1.2 (project environment analysis)

**Select user if backup exists**

When a backup is detected, call `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` to present a TUI decision:

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Previous project configuration found in backup. How would you like to proceed?",
            "header": "Backup Merge Decision",
            "multiSelect": false,
            "options": [
                {
                    "label": "🔄 Merge (Recommended)",
                    "description": "Restore your previous customizations with latest template structure"
                },
                {
                    "label": "📋 New Interview",
                    "description": "Start fresh interview, ignore previous configuration"
                },
                {
                    "label": "⏸️ Skip (Keep Current)",
                    "description": "Keep existing project files without changes"
                }
            ]
        }
    ]
)
```

**Response Processing**:

- **"Merge (Recommended)"** (`answers["0"] === "Merge"`) → Proceed to Phase 1.1 (backup merge workflow)

  - Extract user customizations from backup
  - Combine with latest template structure
  - Update version in HISTORY section
  - Set `optimized: true` in config.json

- **"New Interview"** (`answers["0"] === "New Interview"`) → Proceed to Phase 1.2 (Project environment analysis)

  - Archive existing backup for reference
  - Begin fresh interview without prior customizations
  - Create new product/structure/tech.md from interview results

- **"Skip (Keep Current)"** (`answers["0"] === "Skip"`) → End task
  - Terminate /alfred:0-project execution
  - Preserve all existing files unchanged
  - User must manually run if changes needed

**No backup found**:

- Display: "✅ No previous backup detected. Starting fresh interview..."
- Proceed directly to Phase 1.2 (project environment analysis)

---

### 1.1 Backup merge workflow (when user selects "Merge")

**Purpose**: Restore only user customizations while maintaining the latest template structure.

**STEP 1: Read backup file**

Alfred reads files from the latest backup directory:

```bash
# Latest backup directory path
BACKUP_DIR=.moai-backups/$(ls -t .moai-backups/ | head -1)

# Read backup file
Read $BACKUP_DIR/.moai/project/product.md
Read $BACKUP_DIR/.moai/project/structure.md
Read $BACKUP_DIR/.moai/project/tech.md
Read $BACKUP_DIR/CLAUDE.md
```

**STEP 2: Detect template defaults**

The following patterns are considered "template defaults" (not merged):

- "Define your key user base"
- "Describe the core problem you are trying to solve"
- "List the strengths and differences of your project"
- "MoAI-ADK", "MoAI-Agentic Development Kit", etc. Variable format
- Guide phrases such as "Example:", "Sample:", "Example:", etc.

**STEP 3: Extract user customization**

Extract only **non-template default content** from the backup file:

- `product.md`:
- Define your actual user base in the USER section
- Describe the actual problem in the PROBLEM section
- Real differences in the STRATEGY section
- Actual success metrics in the SUCCESS section
- `structure.md`:
- Actual design in the ARCHITECTURE section
- Actual module structure in the MODULES section
- Actual integration plan in the INTEGRATION section
- `tech.md`:
- The actual technology stack
  in the STACK section - The actual framework
  in the FRAMEWORK section - The actual quality policy
  in the QUALITY section - `HISTORY` section: **Full Preservation** (all files)

**STEP 4: Merge Strategy**

```markdown
Latest template structure (v0.4.0+)
↓
Insert user customization (extracted from backup file)
↓
HISTORY section updates
↓
Version update (v0.1.x → v0.1.x+1)
```

**Merge Principle**:

- ✅ Maintain the latest version of the template structure (section order, header, @TAG format)
- ✅ Insert only user customization (actual content written)
- ✅ Cumulative preservation of the HISTORY section (existing history + merge history)
- ❌ Replace template default values ​​with the latest version

**STEP 5: HISTORY Section Update**

After the merge is complete, add history to the HISTORY section of each file:

```yaml
### v0.1.x+1 (2025-10-19)
- **UPDATED**: Merge backup files (automatic optimization)
- AUTHOR: @Alfred
- BACKUP: .moai-backups/20251018-003638/
- REASON: Restoring user customization after moai-adk init reinitialization
```

**STEP 6: Update config.json**

Set optimization flags after the merge is complete:

```json
{
  "project": {
    "optimized": true,
    "last_merge": "2025-10-19T12:34:56+09:00",
    "backup_source": ".moai-backups/20251018-003638/"
  }
}
```

**STEP 7: Completion Report**

```markdown
✅ Backup merge completed!

📁 Merged files:

- .moai/project/product.md (v0.1.4 → v0.1.5)
- .moai/project/structure.md (v0.1.1 → v0.1.2)
- .moai/project/tech.md (v0.1.1 → v0.1.2)
- .moai/config.json (optimized: false → true)

🔍 Merge history:

- USER section: Restore customized contents of backup file
- PROBLEM section: Restore problem description of backup file
- STRATEGY section: Restore differentials of backup file
- HISTORY section: Add merge history (cumulative retention)

💾 Backup file location:

- Original backup: .moai-backups/20251018-003638/
- Retention period: Permanent (until manual deletion)

📋 Next steps:

1. Review the merged document
2. Additional modifications if necessary
3. Create your first SPEC with /alfred:1-plan

---

**Task completed: /alfred:0-project terminated**
```

**Finish work after merge**: Complete immediately without interview

---

### 1.2 Run project environment analysis (when user selects "New" or no backup)

**Automatically analyzed items**:

1. **Project Type Detection**
   Alfred classifies new vs existing projects by analyzing the directory structure:

- Empty directory → New project
- Code/documentation present → Existing project

2. **Auto-detect language/framework**: Detects the main language of your project based on file patterns
   - pyproject.toml, requirements.txt → Python
   - package.json, tsconfig.json → TypeScript/Node.js
   - pom.xml, build.gradle → Java
   - go.mod → Go
   - Cargo.toml → Rust

- backend/ + frontend/ → full stack

3. **Document status analysis**

- Check the status of existing `.moai/project/*.md` files
- Identify areas of insufficient information
- Organize items that need supplementation

4. **Project structure evaluation**

- Directory structure complexity
- Monolingual vs. hybrid vs. microservice
- Code base size estimation

### 1.3 Establish interview strategy (when user selects "New")

**Select question tree by project type**:

| Project Type              | Question Category  | Focus Areas                                   |
| ------------------------- | ------------------ | --------------------------------------------- |
| **New Project**           | Product Discovery  | Mission, Users, Problems Solved               |
| **Existing Project**      | Legacy Analysis    | Code Base, Technical Debt, Integration Points |
| **TypeScript conversion** | Migration Strategy | TypeScript conversion for existing projects   |

**Question Priority**:

- **Essential Questions**: Core Business Value, Key User Bases (all projects)
- **Technical Questions**: Language/Framework, Quality Policy, Deployment Strategy
- **Governance**: Security Requirements, Traceability Strategy (Optional)

### 1.4 Generate Interview Plan Report (when user selects "Create New")

**Purpose**: Present user with a clear interview plan before execution, allowing review and modification.

**Format of plan to be presented to users**:

```markdown
## 📊 Project Initialization Plan: [PROJECT-NAME]

### Environmental Analysis Results

- **Project Type**: [New/Existing/Hybrid]
- **Languages Detected**: [Language List]
- **Current Document Status**: [Completeness Rating 0-100%]
- **Structure Complexity**: [Simple/Medium/Complex]

### 🎯 Interview Strategy

- **Question Category**: Product Discovery / Structure / Tech
- **Expected Number of Questions**: [N questions (M required + K optional)]
- **Estimated Time Required**: [Time estimation, e.g., 15-20 minutes]
- **Priority Areas**: [Key focus areas to be covered]

### 📋 Interview Phases

1. **Product Discovery** (product.md)

   - Core mission and value proposition
   - Key user bases and success metrics

2. **Structure Blueprint** (structure.md)

   - System architecture strategy
   - Module boundaries and responsibility

3. **Tech Stack Mapping** (tech.md)
   - Language/framework selection
   - Quality and deployment policies

### ⚠️ Important Notes

- **Existing Document**: [Overwrite/Merge/Supplement strategy]
- **Language Settings**: [Conversation language: {{CONVERSATION_LANGUAGE_NAME}}]
- **Team Mode**: [Personal/Team workflow configured]
- **Configuration**: [Compatibility with existing config.json]

### ✅ Expected Deliverables

- **product.md**: Business requirements and strategy document
- **structure.md**: System architecture and design document
- **tech.md**: Technology stack and quality policy document
- **config.json**: Project settings and configurations

---

**Please review the plan above and confirm whether to proceed.**
```

### 1.5 User Approval with AskUserQuestion (when user selects "New")

After Alfred generates the interview plan report, call `AskUserQuestion` tool (documented in moai-alfred-interactive-questions skill) to get explicit user approval before starting the interview.

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Please review the interview plan above. Would you like to proceed with this plan?",
            "header": "Interview Plan Approval",
            "multiSelect": false,
            "options": [
                {
                    "label": "✅ Proceed with Plan",
                    "description": "Start interview following the plan above (Phase 2)"
                },
                {
                    "label": "📋 Modify Plan",
                    "description": "Revise strategy and re-run analysis (back to Phase 1)"
                },
                {
                    "label": "⏹️ Cancel",
                    "description": "Exit initialization, keep existing files unchanged"
                }
            ]
        }
    ]
)
```

**Response Processing**:

- **"Proceed with Plan"** (`answers["0"] === "Proceed"`) → Execute Phase 2

  - Call project-manager agent with approved plan parameters
  - Conduct interview according to the plan
  - Generate product/structure/tech.md documents
  - Save config.json with all settings

- **"Modify Plan"** (`answers["0"] === "Modify"`) → Repeat Phase 1

  - Return to environmental analysis
  - Re-run project type detection
  - Re-run language detection
  - Generate new interview plan with user feedback
  - Ask for approval again with modified plan

- **"Cancel"** (`answers["0"] === "Cancel"`) → End task
  - Terminate /alfred:0-project execution
  - Do not modify any existing files
  - User can re-run command later

**Phase 2 Execution Condition**:

- Only proceed to Phase 2 (project initialization) if user confirms "Proceed with Plan"
- All other responses lead to re-planning or task termination

---

## 🚀 STEP 2: Execute project initialization (after user approves "New")

**Note**: This step will only be executed if the user selects **"New"**.

- When selecting "Merge": End the task in Phase 1.1 (Merge Backups)
- When selecting "Skip": End the task
- When selecting "New": Proceed with the process below

After user approval, the project-manager agent performs initialization.

### 2.1 Call project-manager agent (when user selects "New")

Alfred starts project initialization by calling the project-manager agent with the following parameters:

**Parameters passed to project-manager**:

- **conversation_language** (from STEP 0): Language code selected by user (e.g., "ko", "en", "ja", "zh")
- **language_name** (from STEP 0): Display name of selected language (e.g., "Korean", "English")
- **agent_prompt_language** (from STEP 0.1.2) - **NEW**:
  - `"english"` = All sub-agent prompts in English (Global Standard)
  - `"localized"` = All sub-agent prompts in selected conversation_language (Localized)
- Detected Languages: [Language List from codebase detection]
- Project Type: [New/Existing]
- Existing Document Status: [Existence/Absence]
- Approved Interview Plan: [Plan Summary]
- **Team Mode Git Workflow** (from STEP 0.1.3):
  - `spec_git_workflow: "feature_branch" | "develop_direct" | "per_spec"` (팀 모드만)

**Execution**:

```
Call the Task tool:
- subagent_type: "project-manager"
- description: "Initialize project with conversation language support"
- prompt: """당신은 project-manager 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}} (모든 대화, 문서에 사용)
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}
- 에이전트_프롬프트_언어: {{AGENT_PROMPT_LANGUAGE}} (내부 sub-agent 통신 언어)

에이전트 프롬프트 언어에 따른 작업 방식:

1. **agent_prompt_language = "english"** (Global Standard):
   - 당신(project-manager)은 **영어**로 사고하고 작업합니다
   - 모든 내부 분석과 계획을 영어로 진행합니다
   - 생성된 product.md, structure.md, tech.md는 **{{CONVERSATION_LANGUAGE}}**로 작성합니다
   - Sub-agent들(spec-builder 등)에게 전달하는 프롬프트는 **영어**입니다

2. **agent_prompt_language = "localized"** (Localized):
   - 당신(project-manager)은 **{{CONVERSATION_LANGUAGE}}**로 사고하고 작업합니다
   - 모든 내부 분석과 계획을 {{CONVERSATION_LANGUAGE}}로 진행합니다
   - 생성된 product.md, structure.md, tech.md는 **{{CONVERSATION_LANGUAGE}}**로 작성합니다
   - Sub-agent들(spec-builder 등)에게 전달하는 프롬프트도 **{{CONVERSATION_LANGUAGE}}**입니다

중요: 대화_언어(conversation_language)와 에이전트_프롬프트_언어(agent_prompt_language)는 다를 수 있습니다!
- 대화_언어는 **사용자와의 대화**, **생성 문서**에 사용
- 에이전트_프롬프트_언어는 **sub-agents 통신**, **내부 prompt**에 사용

GIT 워크플로우 설정 (팀 모드):
- spec_git_workflow: [feature_branch | develop_direct | per_spec]
  - "feature_branch": feature/spec-* 브랜치 생성, PR 기반 리뷰, develop 병합
  - "develop_direct": develop에 직접 커밋, 브랜치 생성 안 함
  - "per_spec": SPEC별로 사용자에게 물어봄 (/alfred:1-plan 실행 중)
- 참고: 이 값을 .moai/config.json github.spec_git_workflow에 저장하여 git-manager가 참조하도록

프로젝트_타입: [new|existing]
감지된_언어들: [감지된 코드베이스 언어들]

중요 지시사항:
모든 인터뷰와 생성된 문서는 대화_언어(conversation_language)로 작성되어야 합니다:
- product.md: {{CONVERSATION_LANGUAGE}}로 생성
- structure.md: {{CONVERSATION_LANGUAGE}}로 생성
- tech.md: {{CONVERSATION_LANGUAGE}}로 생성

conversation_language가 'ko'인 경우: 모든 설명 내용을 한국어로
conversation_language가 'ja'인 경우: 모든 설명 내용을 일본어로
다른 언어인 경우: 지정된 언어를 따릅니다

프로젝트 초기화 후, 다음과 같이 .moai/config.json 업데이트:
{
  "language": {
    "conversation_language": "{{CONVERSATION_LANGUAGE}}",
    "conversation_language_name": "{{CONVERSATION_LANGUAGE_NAME}}",
    "agent_prompt_language": "{{AGENT_PROMPT_LANGUAGE}}"
  },
  "github": {
    "spec_git_workflow": "[feature_branch|develop_direct|per_spec]"
  }
}

스킬 호출:
필요 시 명시적 Skill() 호출 사용:
- Skill("moai-alfred-language-detection") - 코드베이스 언어 감지
- Skill("moai-foundation-langs") - 다국어 프로젝트 설정

작업: 프로젝트 인터뷰를 진행하고 product/structure/tech.md 문서를 생성/업데이트합니다.
에이전트_프롬프트_언어 설정에 따라 sub-agent들과의 통신 언어를 결정합니다."""
```

**Outcome**: The project-manager agent conducts structured interviews entirely in the selected language and creates/updates product/structure/tech.md documents in that language.

### 2.2 Automatic activation of Alfred Skills (optional)

After the project-manager has finished creating the document, **Alfred can optionally call Skills** (upon user request).

**Automatic activation conditions** (optional):

| Conditions                           | Automatic selection Skill    | Purpose                                |
| ------------------------------------ | ---------------------------- | -------------------------------------- |
| User Requests "Quality Verification" | moai-alfred-trust-validation | Initial project structure verification |

**Execution flow** (optional):

```
1. project-manager completion
    ↓
2. User selection:
 - "Quality verification required" → moai-alfred-trust-validation (Level 1 quick scan)
 - "Skip" → Complete immediately
```

**Note**: Quality verification is optional during the project initialization phase.

### 2.3 Sub-agent moai-alfred-interactive-questions (Nested)

**The project-manager agent can internally call the TUI survey skill** to check the details of the task.

**When to call**:

- Before overwriting existing project documents
- When selecting language/framework
- When changing important settings

**Example** (inside project-manager): Ask whether to "overwrite file" with `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)`,

- Allows you to choose between **Overwrite** / **Merge** / **Skip**.

**Nested pattern**:

- **Command level** (Phase approval): Called by Alfred → "Shall we proceed with Phase 2?"
- **Sub-agent level** (Detailed confirmation): Called by project-manager → "Shall we overwrite the file?"

### 2.4 Processing method by project type

#### A. New project (Greenfield)

**Interview Flow**:

1. **Product Discovery** (create product.md)

- Define core mission (DOC:MISSION-001)
- Identify key user base (SPEC:USER-001)
- Identify key problems to solve (SPEC:PROBLEM-001)
- Summary of differences and strengths (DOC:STRATEGY-001)
- Setting success indicators (SPEC:SUCCESS-001)

2. **Structure Blueprint** (create structure.md)

- Selection of architecture strategy (DOC:ARCHITECTURE-001)
- Division of responsibilities by module (DOC:MODULES-001)
- External system integration plan (DOC:INTEGRATION-001)
- Define traceability strategy (DOC:TRACEABILITY-001)

3. **Tech Stack Mapping** (written by tech.md)

- Select language & runtime (DOC:STACK-001)
- Determine core framework (DOC:FRAMEWORK-001)
- Set quality gate (DOC:QUALITY-001)
  - Define security policy (DOC:SECURITY-001)
- Plan distribution channels (DOC:DEPLOY-001)

**Automatically generate config.json**:

```json
{
  "project_name": "detected-name",
  "project_type": "single|fullstack|microservice",
  "project_language": "python|typescript|java|go|rust",
  "test_framework": "pytest|vitest|junit|go test|cargo test",
  "linter": "ruff|biome|eslint|golint|clippy",
  "formatter": "black|biome|prettier|gofmt|rustfmt",
  "coverage_target": 85,
  "mode": "personal"
}
```

#### B. Existing project (legacy introduction)

**Legacy Snapshot & Alignment**:

**STEP 1: Identify the overall project structure**

Alfred identifies the entire project structure:

- Visualize the directory structure using the tree or find commands
- Exclude build artifacts such as node_modules, .git, dist, build, **pycache**, etc.
- Identify key source directories and configuration files.

**Output**:

- Visualize the entire folder/file hierarchy of the project
- Identify major directories (src/, tests/, docs/, config/, etc.)
- Check language/framework hint files (package.json, pyproject.toml, go.mod, etc.)

**STEP 2: Establish parallel analysis strategy**

Alfred identifies groups of files by the Glob pattern:

1. **Configuration files**: _.json, _.toml, _.yaml, _.yml, \*.config.js
2. **Source code files**: src/\*_/_.{ts,js,py,go,rs,java}
3. **Test files**: tests/**/\*.{ts,js,py,go,rs,java}, **/_.test._, \*_/_.spec.\*
4. **Documentation files**: _.md, docs/\*\*/_.md, README*, CHANGELOG*

**Parallel Read Strategy**:

- Speed ​​up analysis by reading multiple files simultaneously with the Read tool
- Batch processing for each file group
- Priority: Configuration file → Core source → Test → Document

**STEP 3: Analysis and reporting of characteristics for each file**

As each file is read, the following information is collected:

1. **Configuration file analysis**

- Project metadata (name, version, description)
- Dependency list and versions
- Build/test script
- Confirm language/framework

2. **Source code analysis**

- Identify major modules and classes
- Architectural pattern inference (MVC, clean architecture, microservice, etc.)
- Identify external API calls and integration points
- Key areas of domain logic

3. **Test code analysis**

- Check test framework
- Identify coverage settings
- Identify key test scenarios
- Evaluate TDD compliance

4. **Document analysis**

- Existing README contents
- Existence of architecture document
- API document status
- Installation/deployment guide completeness

**Report Format**:

```markdown
## Analysis results for each file

### Configuration file

- package.json: Node.js 18+, TypeScript 5.x, Vitest test
- tsconfig.json: strict mode, ESNext target
- biome.json: Linter/formatter settings exist

### Source code (src/)

- src/core/: Core business logic (3 modules)
- src/api/: REST API endpoints (5 routers)
- src/utils/: Utility functions (logging, verification, etc.)
- Architecture: Hierarchical (controller) → service → repository)

### Tests (tests/)

- Vitest + @testing-library used
- Unit test coverage estimated at about 60%
- E2E testing lacking

### Documentation

- README.md: Only installation guide
- Absence of API documentation
- Absence of architecture document
```

**STEP 4: Comprehensive analysis and product/structure/tech reflection**

Based on the collected information, it is reflected in three major documents:

1. Contents reflected in **product.md**

- Project mission extracted from existing README/document
- Main user base and scenario inferred from code
- Backtracking of core problem to be solved
- Preservation of existing assets in "Legacy Context"

2. Contents reflected in **structure.md**

- Identified actual directory structure
- Responsibility analysis results for each module
- External system integration points (API calls, DB connections, etc.)
- Technical debt items (marked with @CODE tag)

3. **tech.md reflection content**

- Languages/frameworks/libraries actually in use
- Existing build/test pipeline
- Status of quality gates (linter, formatter, test coverage)
- Identification of security/distribution policy
- Items requiring improvement (marked with TODO tags)

**Preservation Policy**:

- Supplement only the missing parts without overwriting existing documents
- Preserve conflicting content in the "Legacy Context" section
- Mark items needing improvement with @CODE and TODO tags

**Example Final Report**:

```markdown
## Complete analysis of existing project

### Environment Information

- **Language**: TypeScript 5.x (Node.js 18+)
- **Framework**: Express.js
- **Test**: Vitest (coverage ~60%)
- **Linter/Formatter**: Biome

### Main findings

1. **Strengths**:

- High type safety (strict mode)
- Clear module structure (separation of core/api/utils)

2. **Needs improvement**:

- Test coverage below 85% (TODO:TEST-COVERAGE-001)
- Absence of API documentation (TODO:DOCS-API-001)
- Insufficient E2E testing (@CODE:TEST-E2E-001)

### Next step

1. product/structure/tech.md creation completed
2. @CODE/TODO item priority confirmation
3. /alfred:Start writing an improvement SPEC with 1-spec
```

### 2.3 Document creation and verification

**Output**:

- `.moai/project/product.md` (Business Requirements)
- `.moai/project/structure.md` (System Architecture)
- `.moai/project/tech.md` (Technology Stack and policy)
- `.moai/config.json` (project settings)

**Quality Verification**:

- [ ] Verify existence of all required @TAG sections
- [ ] Verify compliance with EARS syntax format
- [ ] Verify config.json syntax validity
- [ ] Verify cross-document consistency

### 2.4 Completion Report

```markdown
✅ Project initialization complete!

📁 Documents generated:

- .moai/project/product.md (Business Definition)
- .moai/project/structure.md (Architecture Design)
- .moai/project/tech.md (Technology Stack)
- .moai/config.json (project settings)

🔍 Detected environments:

- Language: [List of languages]
- Frameworks: [List of frameworks]
- Test tools: [List of tools]

📋 Next steps:

1. Review the generated document
2. Create your first SPEC with /alfred:1-plan
3. If necessary, readjust with /alfred:0-project update
```

### 2.5: Initial structural verification (optional)

After project initialization is complete, you can optionally run quality verification.

**Execution Conditions**: Only when explicitly requested by the user.

**Verification Purpose**:

- Basic verification of project documentation and configuration files
- Verification of compliance with the TRUST principles of the initial structure
- Validation of configuration files

**How ​​it works**:
Alfred only calls the trust-checker agent to perform project initial structural verification if explicitly requested by the user.

**Verification items**:

- **Document completeness**: Check existence of required sections in product/structure/tech.md
- **Settings validity**: Verify config.json JSON syntax and required fields
- **TAG scheme**: Check compliance with @TAG format in document
- **EARS syntax**: Validation of the EARS template to be used when writing SPECs

**Run Verification**: Level 1 quick scan (3-5 seconds)

**Handling verification results**:

✅ **Pass**: Can proceed to next step

- Documents and settings are all normal

⚠️ **Warning**: Proceed after warning

- Some optional sections are missing
- Recommendations not applied

❌ **Critical**: Needs fix

- Required section missing
- config.json syntax error
- User choice: "Revalidate after fix" or "Skip"

**Skip verification**:

- Verification is not run by default
- Run only when explicitly requested by the user

### 2.6: Agent & Skill Tailoring (Project Optimization)

Based on the results of the interviews and initial analysis, we recommend and activate sub-agents and skills that should be immediately utilized in the project.
Before actual application, user confirmation is received with `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)`, and selected items are recorded in `CLAUDE.md` and `.moai/config.json`.

#### 2.6.0 Create cc-manager briefing

Once the document creation is complete, **read all three documents (product/structure/tech.md)** and summarize the following information to create a text called `cc_manager_briefing`.

- `product.md`: Organize the mission, key users, problems to be solved, success indicators, and backlog (TODO) with a quotation from the original text or a one-line summary.
- `structure.md`: Records architecture type, module boundaries and scope of responsibility, external integration, traceability strategy, and TODO contents.
- `tech.md`: Organizes language/framework version, build/test/deployment procedures, quality/security policy, operation/monitoring method, and TODO items.

Be sure to include the source (e.g. `product.md@SPEC:SUCCESS-001`) for each item so that cc-manager can understand the basis.

#### 2.6.1 cc-manager judgment guide

cc-manager selects the required sub-agents and skills based on the briefing.The table below is a reference guide to help you make a decision, and when making an actual call, the supporting sentences from the relevant document are also delivered.

| Project requirements (document basis)                                              | Recommended sub-agent/skill                                                                                             | Purpose                                                                |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| High quality and coverage goals (product.md:SPEC:SUCCESS-001)                      | `tdd-implementer`, `moai-essentials-debug`, `moai-essentials-review`                                                    | Establishment of RED·GREEN·REFACTOR workflow                           |
| Traceability/TAG improvement request (structure.md:DOC:TRACEABILITY-001)           | `doc-syncer`, `moai-alfred-tag-scanning`, `moai-alfred-trust-validation`                                                | Enhanced TAG traceability and document/code synchronization            |
| Deployment automation/branch strategy required (`structure.md` Architecture/TODO)  | `git-manager`, `moai-alfred-git-workflow`, `moai-foundation-git`                                                        | Branch Strategy·Commit Policy·PR Automation                            |
| Refactoring legacy modules (`product.md` BACKLOG, `tech.md` TODO)                  | `implementation-planner`, `moai-essentials-refactor`                                                                    | Technical Debt Diagnosis and Refactoring Roadmap                       |
| Strengthening regulatory/security compliance (tech.md:DOC:SECURITY-001)            | `quality-gate`, `moai-alfred-trust-validation`, `moai-foundation-trust`, `moai-domain-security`                         | TRUST S (Secured) and Trackable Compliance, Security Consulting        |
| CLI Automation/Tooling Requirements (`tech.md` BUILD/CLI section)                  | `implementation-planner`, `moai-domain-cli-tool`, detected language skills (e.g. `moai-lang-python`)                    | CLI command design, input/output standardization                       |
| Data analysis/reporting needs (`product.md` DATA, `tech.md` ANALYTICS)             | `implementation-planner`, `moai-domain-data-science`, detected language skills                                          | Data Pipeline·Notebook Job Definition                                  |
| Improved database structure (`structure.md` DB, `tech.md` STORAGE)                 | `doc-syncer`, `moai-domain-database`, `moai-alfred-tag-scanning`                                                        | Strengthening schema documentation and TAG-DB mapping                  |
| DevOps/Infrastructure automation required (`tech.md` DEVOPS, `structure.md` CI/CD) | `implementation-planner`, `moai-domain-devops`, `moai-alfred-git-workflow`                                              | Establishing a deployment pipeline and IaC strategy                    |
| Introduction of ML/AI functions (`product.md` AI, `tech.md` MODEL)                 | `implementation-planner`, `moai-domain-ml`, detected language skills                                                    | Model training/inference pipeline definition                           |
| Mobile app strategy (`product.md` MOBILE, `structure.md` CLIENT)                   | `implementation-planner`, `moai-domain-mobile-app`, detected language skills (e.g. `moai-lang-dart`, `moai-lang-swift`) | Mobile client structure design                                         |
| Strengthening coding standards/review process (`tech.md` REVIEW)                   | `quality-gate`, `moai-essentials-review`                                                                                | Strengthening review checklist and quality reporting                   |
| Requires onboarding/training mode (`tech.md` STACK description, etc.)              | `moai-alfred-interactive-questions`, `moai-adk-learning`, `agentic-coding` Output style                                 | Enhanced interview TUI and automatically provided onboarding materials |

> **Language/Domain Skill Selection Rules**
>
> - Select and add one relevant language skill (`moai-lang-python`, `moai-lang-java`, …) based on the `moai-alfred-language-detection` results or the stack recorded in the Tech section of the briefing.
> - Skills listed in the domain row are automatically included by cc-manager in the `selected_skills` list when the conditions are met.
> - The skill directory is always copied in its entirety, and only actual activation is recorded in `skill_pack` and `CLAUDE.md`.

If multiple conditions are met, the candidates are merged without duplicates and organized into sets of `candidate_agents`, `candidate_skills`, and `candidate_styles`.

#### 2.6.2 User confirmation flow

`AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` asks "whether to enable recommended items."

- Provides three options: **Install all** / **Install selectively** / **Do not install**.
  Selecting "Selective Install" presents the list of candidates again as multiple choices, allowing the user to select only the items they need.

#### 2.6.3 Activation and Recording Steps

1. **Preparing briefing**: Organize the results of user selection (install all/install selectively) and the full text of `cc_manager_briefing`.
2. **Call the cc-manager agent**:

- Call `subagent_type: "cc-manager"` with the `Task` tool and include a briefing and user selections in the prompt.
- cc-manager determines the necessary sub-agents and skills based on the briefing, and copies and updates `CLAUDE.md`, `.claude/agents/alfred/*.md`, and `.claude/skills/*.md` as customized for the project.

3. **Check for configuration updates**: Review the results reflected by cc-manager.

- Sub-Agents: Keep the `.claude/agents/alfred/` template active and list it in the `CLAUDE.md` "Agents" section.
- Skills: Check the `.claude/skills/` document and add it to the `CLAUDE.md` "Skills" section.
- Output style: Apply `.claude/output-styles/alfred/` and record the activation in `CLAUDE.md` "Output Styles".

4. **Update config.json**
   ```json
   {
     "project": {
       "optimized": true,
       "agent_pack": ["tdd-implementer", "doc-syncer"],
       "skill_pack": ["moai-alfred-git-workflow", "moai-alfred-tag-scanning"],
       "output_styles": ["moai-adk-learning"]
     }
   }
   ```
   Merge existing properties, if any.
5. **Final Report**: Add a list of "Activated Sub-Agents/Skills/Style" and a `cc_manager_briefing` summary at the top of the Completion Report, and reflect the same contents in the `CLAUDE.md` table so that they are automatically searched in subsequent commands.

## Interview guide by project type

### New project interview area

**Product Discovery** (product.md)

- Core mission and value proposition
- Key user bases and needs
- 3 key problems to solve
- Differentiation compared to competing solutions
- Measurable indicators of success

**Structure Blueprint** (structure.md)

- System architecture strategy
- Separation of modules and division of responsibilities
- External system integration plan
- @TAG-based traceability strategy

**Tech Stack Mapping** (tech.md)

- Language/runtime selection and version
- Framework and libraries
- Quality gate policy (coverage, linter)
- Security policy and distribution channel

### Existing project interview area

**Legacy Analysis**

- Identify current code structure and modules
- Status of build/test pipeline
- Identify technical debt and constraints
- External integration and authentication methods
- MoAI-ADK transition priority plan

**Retention Policy**: Preserve existing documents in the "Legacy Context" section and mark items needing improvement with @CODE/TODO tags

## 🏷️ TAG system application rules

**Automatically create @TAGs per section**:

- Mission/Vision → @DOC:MISSION-XXX, @DOC:STRATEGY-XXX
- Customization → @SPEC:USER-XXX, @SPEC:PERSONA-XXX
- Problem analysis → @SPEC:PROBLEM-XXX, @SPEC:SOLUTION-XXX
- Architecture → @DOC:ARCHITECTURE-XXX, @SPEC:PATTERN-XXX
- Technology Stack → @DOC:STACK-XXX, @DOC:FRAMEWORK-XXX

**Legacy Project Tags**:

- Technical debt → @CODE:REFACTOR-XXX, @CODE:TEST-XXX, @CODE:MIGRATION-XXX
- Resolution plan → @CODE:MIGRATION-XXX, TODO:SPEC-BACKLOG-XXX
- Quality improvement → TODO:TEST-COVERAGE-XXX, TODO:DOCS-SYNC-XXX

## Error handling

### Common errors and solutions

**Error 1**: Project language detection failed

```
Symptom: "Language not detected" message
Solution: Specify language manually or create language-specific settings file
```

**Error 2**: Conflict with existing document

```
Symptom: product.md already exists and has different contents
Solution: Preserve existing contents and add new contents in "Legacy Context" section
```

**Error 3**: Failed to create config.json

```
Symptom: JSON syntax error or permission denied
Solution: Check file permissions (chmod 644) or create config.json manually
```

---

## /alfred:0-project update: Template optimization (subcommand)

> **Purpose**: After running moai-adk update, compare the backup and new template to optimize the template while preserving user customization.

### Execution conditions

This subcommand is executed under the following conditions:

1. **After executing moai-adk update**: `optimized=false` status in `config.json`
2. **Template update required**: When there is a difference between the backup and the new template
3. **User explicit request**: User directly executes `/alfred:0-project update`

### Execution flow

#### Phase 1: Backup analysis and comparison

1. **Make sure you have the latest backup**:

   ```bash

   ```

# Browse the latest backups in the .moai-backups/ directory

ls -lt .moai-backups/ | head -1

````

2. **Change Analysis**:
- Compare `.claude/` directory from backup with current template
- Compare `.moai/project/` document from backup with current document
- Identify user customization items

3. **Create Comparison Report**:
```markdown
## 📊 Template optimization analysis

### Changed items
- CLAUDE.md: "## Project Information" section needs to be preserved
- settings.json: 3 env variables need to be preserved
- product.md: Has user-written content

### Recommended Action
- Run Smart Merge
- Preserve User Customizations
- Set optimized=true
````

4. **Waiting for user approval**

Call `AskUserQuestion` tool (documented in moai-alfred-interactive-questions skill) to obtain user approval for template optimization.

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Template optimization analysis complete. Changes detected in backup vs current template. How would you like to proceed?",
            "header": "Template Optimization",
            "multiSelect": false,
            "options": [
                {
                    "label": "✅ Proceed",
                    "description": "Run smart merge: preserve customizations with latest template (Phase 2)"
                },
                {
                    "label": "👀 Preview",
                    "description": "Show detailed change list before proceeding"
                },
                {
                    "label": "⏸️ Skip",
                    "description": "Keep current template unchanged (optimized: false)"
                }
            ]
        }
    ]
)
```

**Response Processing**:

- **"Proceed"** (`answers["0"] === "Proceed"`) → Execute Phase 2

  - Run smart merge logic
  - Preserve user customizations from backup
  - Combine with latest template structure
  - Set `optimized: true` in config.json

- **"Preview"** (`answers["0"] === "Preview"`) → Display detailed changes

  - Show file-by-file comparison
  - Highlight customization sections
  - Ask approval again with "Proceed" or "Skip" only

- **"Skip"** (`answers["0"] === "Skip"`) → Keep current state
  - Do not modify any files
  - Keep `optimized: false` in config.json
  - User can run again with `moai-adk update` later

#### Phase 2: Run smart merge (after user approval)

1. **Execute smart merge logic**:

- Run `TemplateProcessor.copy_templates()`
- CLAUDE.md: Preserve "## Project Information" section
- settings.json: env variables and permissions.allow merge

2. Set **optimized=true**:

   ```python
   # update config.json
   config_data["project"]["optimized"] = True
   ```

3. **Optimization completion report**:
   ```markdown
   ✅ Template optimization completed!
   ```

📄 Merged files:

- CLAUDE.md (preserves project information)
- settings.json (preserves env variables)

⚙️ config.json: optimized=true Configuration complete

````

### Alfred Automation Strategy

**Alfred automatic decision**:
- Automatically call project-manager agent
- Check backup freshness (within 24 hours)
- Automatically analyze changes

**Auto-activation of Skills**:
- moai-alfred-tag-scanning: TAG chain verification
- moai-alfred-trust-validation: Verification of compliance with TRUST principles

### Running example

```bash
# After running moai-adk update
moai-adk update

# Output:
# ✓ Update complete!
# ℹ️  Next step: Run /alfred:0-project update to optimize template changes

# Run Alfred
/alfred:0-project update

# → Phase 1: Generate backup analysis and comparison report
# → Wait for user approval
# → Phase 2: Run smart merge, set optimized=true
````

### caution

- **Backup required**: Cannot run without backup in `.moai-backups/` directory
- **Manual review recommended**: Preview is required if there are important customizations
- **Conflict resolution**: Request user selection in case of merge conflict

---

## 🚀 STEP 3: Project Custom Optimization (Optional)

**Execution conditions**:

- After completion of Phase 2 (project initialization)
- or after completion of Phase 1.1 (backup merge)
- Explicitly requested by the user or automatically determined by Alfred

**Purpose**: Lightweight by selecting only Commands, Agents, and Skills that fit the project characteristics (37 skills → 3~5)

### 3.1 Automatic execution of Feature Selection

**Alfred automatically calls the moai-alfred-feature-selector skill**:

**Skill Entry**:

- `.moai/project/product.md` (project category hint)
- `.moai/project/tech.md` (main language, framework)
- `.moai/config.json` (project settings)

**Skill Output**:

```json
{
  "category": "web-api",
  "language": "python",
  "framework": "fastapi",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": [
    "spec-builder",
    "code-builder",
    "doc-syncer",
    "git-manager",
    "debug-helper"
  ],
  "skills": ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"],
  "excluded_skills_count": 34,
  "optimization_rate": "87%"
}
```

**How ​​to Run**:

```
Alfred: Skill("moai-alfred-feature-selector")
```

---

### 3.2 Automatic execution of Template Generation

**Alfred automatically calls the moai-alfred-template-generator skill**:

**Skill input**:

- `.moai/.feature-selection.json` (feature-selector output)
- `CLAUDE.md` template
- Entire commands/agents/skills file

**Skill Output**:

- `CLAUDE.md` (custom agent table - selected agents only)
- `.claude/commands/` (selected commands only)
- `.claude/agents/` (selected agents only)
- `.claude/skills/` (selected skills only)
- `.moai/config.json` (updates `optimized: true`)

**How ​​to Run**:

```
Alfred: Skill("moai-alfred-template-generator")
```

---

### 3.3 Optimization completion report

**Report Format**:

```markdown
✅ Project customized optimization completed!

📊 Optimization results:

- **Project**: MoAI-ADK
- **Category**: web-api
- **Main language**: python
- **Framework**: fastapi

🎯 Selected capabilities:

- Commands: 4 items (0-project, 1-spec, 2-build, 3-sync)
- Agents: 5 items (spec-builder, code-builder, doc-syncer, git-manager, debug-helper)
- Skills: 3 items (moai-lang-python, moai-domain-web-api, moai-domain-backend)

💡 Lightweight effect:

- Skills excluded: 34
- Lightweight: 87%
- CLAUDE.md: Create custom agent table

📋 Next steps:

1. Check the CLAUDE.md file (only 5 agents are displayed)
2. Run /alfred:1-plan "first function"
3. Start the MoAI-ADK workflow
```

---

### 3.4 Skip Phase 3 (optional)

**Users can skip Phase 3**:

**Skip condition**:

- User explicitly selects "Skip"
- "Simple project" when Alfred automatically determines (only basic features required)

**Skip effect**:

- Maintain all 37 skills (no lightweighting)
- Maintain default 9 agents in CLAUDE.md template
- Maintain `optimized: false` in config.json

---

## Next steps

**Recommendation**: For better performance and context management, start a new chat session with the `/clear` or `/new` command before proceeding to the next step.

After initialization is complete:

- **New project**: Run `/alfred:1-plan` to create design-based SPEC backlog
- **Legacy project**: Review @CODE/@CODE/TODO items in product/structure/tech document and confirm priority
- **Set Change**: Run `/alfred:0-project` again to update document
- **Template optimization**: Run `/alfred:0-project update` after `moai-adk update`

## Final Step

After project initialization completes, Alfred automatically invokes AskUserQuestion to ask the user what to do next:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Project initialization complete. What would you like to do next?",
            "header": "Next Steps",
            "multiSelect": false,
            "options": [
                {
                    "label": "📋 Start SPEC Creation",
                    "description": "Begin first SPEC with /alfred:1-plan command"
                },
                {
                    "label": "🔍 Review Project Structure",
                    "description": "Review and edit generated project documents"
                },
                {
                    "label": "🔄 Start New Session",
                    "description": "Execute /clear for fresh session (recommended for performance)"
                }
            ]
        }
    ]
)
```

**Response Processing**:

- **"📋 Start SPEC Creation"** (`answers["0"] === "Start SPEC"`) → Proceed to `/alfred:1-plan`

  - Display: "✅ Ready for SPEC creation workflow..."
  - User can immediately run: `/alfred:1-plan "first feature name"`
  - Continue to next phase without session break

- **"🔍 Review Project Structure"** (`answers["0"] === "Review"`) → Review generated documents

  - Display: "📁 Open these files for review:"
    - `.moai/project/product.md` - Business requirements
    - `.moai/project/structure.md` - System architecture
    - `.moai/project/tech.md` - Technology stack
  - After review, user can run `/alfred:1-plan` or `/alfred:0-project` again for updates
  - Display: "💾 Save changes manually in editor or run `/alfred:0-project` again"

- **"🔄 Start New Session"** (`answers["0"] === "New Session"`) → Start fresh session
  - Display: "⏳ Preparing to clear session..."
  - Note: This improves context window management for large projects
  - Next session can start with: `/alfred:1-plan "next feature"`
  - Alternative: Type `/clear` in shell to restart manually

---

## Related commands

- `/alfred:1-plan` - Start writing SPEC
- `/alfred:9-update` - MoAI-ADK update
- `moai doctor` - System diagnosis
- `moai status` - Check project status
