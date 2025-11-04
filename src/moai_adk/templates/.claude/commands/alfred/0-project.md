---
name: alfred:0-project
description: "Initialize project metadata and documentation"
argument-hint: "[setting|update]"
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

> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.

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

**UX 개선**: 배치 호출로 상호작용 최소화
- **기본 배치**: 3개 질문을 1회 호출 (3 turns → 1 turn, **66% 상호작용 감소**)
- **팀 모드 추가 배치**: 2개 질문을 1회 호출 (2 turns → 1 turn, **50% 상호작용 감소**)
- **전체 효과**: 평균 **60% 상호작용 감소**, 사용자 경험 대폭 개선

### 0.0 명령어 진입점: 서브커맨드 파싱 (신규 - 명령형 지침)

**Your immediate task**: Detect which subcommand the user provided and route to the correct workflow.

#### ⚡ Step 1: Check what subcommand the user provided

**Look at the user's command carefully**:
- Did they type `/alfred:0-project setting`?
- Did they type `/alfred:0-project update`?
- Did they type just `/alfred:0-project` (no subcommand)?
- Did they type something invalid like `/alfred:0-project xyz`?

#### ⚡ Step 2: Route based on subcommand

**IF user typed: `/alfred:0-project setting`**:
1. Print: "🔧 Entering Settings Mode - Modify existing project configuration"
2. Jump to **STEP 0-SETTING** below
3. Skip ALL other sections
4. Stop after completing STEP 0-SETTING
5. **DO NOT proceed** to STEP 1, 2, or 3

**ELSE IF user typed: `/alfred:0-project update`**:
1. Print: "🔄 Entering Template Update Mode - Optimize templates after moai-adk update"
2. Jump to **STEP 0-UPDATE** below
3. Skip ALL other sections
4. Stop after completing STEP 0-UPDATE
5. **DO NOT proceed** to STEP 1, 2, or 3

**ELSE IF user typed: `/alfred:0-project` (no subcommand, nothing after)**:
1. Check if the file `.moai/config.json` exists in the current directory
   - Read the file path: `.moai/config.json`
   - IF file exists → Print "✅ Project is already initialized!" AND jump to **STEP 0.1.0**
   - IF file does NOT exist → Print "🚀 Starting first-time project initialization..." AND jump to **STEP 0.1.1**

**ELSE IF user typed an invalid subcommand** (like `/alfred:0-project xyz`):
1. Print this error message:
   ```
   ❌ Unknown subcommand: xyz

   Valid subcommands:
   /alfred:0-project          - Auto-detect mode (first-time or already initialized)
   /alfred:0-project setting  - Modify existing settings
   /alfred:0-project update   - Optimize templates after moai-adk update

   Example: /alfred:0-project setting
   ```
2. Exit immediately
3. **DO NOT make any changes**

#### ⚡ Step 3: CRITICAL RULES

⚠️ **IMPORTANT - Read this carefully**:
- Execute ONLY ONE mode per command invocation
- **DO NOT execute multiple modes** (e.g., do not run setting mode AND first-time setup in the same invocation)
- Stop and exit immediately after completing the selected mode
- **DO NOT jump to STEP 1 or later** unless that is the explicitly detected mode
- **DO NOT guess** which mode the user wanted - always detect from their actual command

### 0.1 Already Initialized Check (conditional entry point)

**Purpose**: Determine whether this is a first-time initialization or a subsequent run on an already-initialized project.

**Execution Condition**:
- Default mode (no subcommand): `/alfred:0-project` executed with no arguments
- Comes after STEP 0.0 subcommand parsing

**Implementation Steps**:

1. **Check if `.moai/config.json` exists**:

   ```bash
   if [ -f .moai/config.json ]; then
       # Project is already initialized
       # Proceed to "0.1.0 Already Initialized Flow"
   else
       # Project is new/uninitialized
       # Proceed to "0.1.1 First-time Setup Flow"
   fi
   ```

2. **Display appropriate message**:

   **If already initialized**:
   ```markdown
   ✅ Project is already initialized!

   Current settings:
   - Language: 한국어 (ko)
   - Nickname: GOOS

   What would you like to do?
   ```

   **If not initialized**:
   ```markdown
   🚀 Starting first-time project initialization...
   ```

---

### 0.1.0 Already Initialized Flow (when config.json exists) - 명령형 지침

**Purpose**: Show options for an already-initialized project and handle user selection.

#### Step 1: Load and display current configuration

1. **Read `.moai/config.json`** to get current settings
2. **Extract and display** these values:
   ```
   ✅ **Language**: [value from language.conversation_language]
   ✅ **Nickname**: [value from user.nickname]
   ✅ **Agent Prompt Language**: [value from language.agent_prompt_language]
   ✅ **GitHub Auto-delete Branches**: [value from github.auto_delete_branches]
   ✅ **SPEC Git Workflow**: [value from github.spec_git_workflow]
   ✅ **Report Generation**: [value from report_generation.user_choice]
   ```

#### Step 2: Ask the user what they want to do

**Present these 4 options** to the user (let them choose one):

1. **"🔧 Modify Settings"** - Change language, nickname, GitHub settings, or reports config
2. **"📋 Review Current Setup"** - Display full current project configuration
3. **"🔄 Re-initialize"** - Run full initialization again (with warning)
4. **"⏸️ Cancel"** - Exit without making any changes

**Wait for the user to select one option**.

#### Step 3: Handle user's selection

**IF user selected: "🔧 Modify Settings"**:
1. Print: "🔧 Entering Settings Mode..."
2. **Jump to STEP 0-SETTING** (same as `/alfred:0-project setting`)
3. Let STEP 0-SETTING handle the rest
4. Stop after STEP 0-SETTING completes

**ELSE IF user selected: "📋 Review Current Setup"**:
1. Print this header: `## Current Project Configuration`
2. Show all current settings (from config.json):
   ```
   ✅ **Language**: [value]
   ✅ **Nickname**: [value]
   ✅ **Agent Prompt Language**: [value]
   ✅ **GitHub Auto-delete Branches**: [value]
   ✅ **SPEC Git Workflow**: [value]
   ✅ **Report Generation**: [value]

   📝 Configuration saved in: .moai/config.json
   📁 Project files: .moai/project/

   To modify settings, run: /alfred:0-project setting
   ```
3. Print: "✅ Configuration review complete."
4. Exit (stop the command)

**ELSE IF user selected: "🔄 Re-initialize"**:
1. Print this warning:
   ```
   ⚠️ WARNING: This will re-run the full project initialization

   Your existing files will be preserved in:
   - Backup: .moai-backups/[TIMESTAMP]/
   - Current: .moai/project/*.md (will be UPDATED)
   ```
2. **Ask the user**: "Are you sure you want to continue? Type 'yes' to confirm or anything else to cancel"
3. **IF user typed 'yes'**:
   - Print: "🔄 Starting full re-initialization..."
   - **Jump to STEP 0.1.1** (First-time Setup)
   - Let STEP 0.1.1 handle the rest
4. **ELSE** (user typed anything else):
   - Print: "✅ Re-initialization cancelled."
   - Exit (stop the command)

**ELSE IF user selected: "⏸️ Cancel"**:
1. Print:
   ```
   ✅ Exiting without changes.

   Your project remains initialized with current settings.
   To modify settings later, run: /alfred:0-project setting
   ```
2. Exit immediately (stop the command)

---

### 0.1.1 First-time Setup Flow (when config.json doesn't exist)

**Purpose**: Collect initial language, nickname, and team mode settings for a new project.

**Flow**:

1. Display welcome message
2. Proceed to batched questions (STEP 0.1.2 below)
3. Save responses to `.moai/config.json`

---

### 0.1.2 배치 설계: 언어 선택 + 사용자 닉네임 + GitHub 설정 확인 (1-3회 호출)

#### 📌 배치 호출의 의미

**배치 호출(Batch Invocation)**이란 **여러 개의 관련 질문을 하나의 AskUserQuestion 호출에 담아** 사용자가 한 번에 응답하도록 하는 방식입니다.

| 방식 | 상호작용 수 | 사용자 경험 |
|------|-----------|----------|
| **순차 호출** ❌ | 3 turns (질문3 → 대답3) | 반복적, 피곤함 |
| **배치 호출** ✅ | 1 turn (질문3 → 대답3) | 빠름, 효율적 |

Alfred가 `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` 를 사용하여 **배치 호출**로 필수 정보를 수집합니다:

**기본 배치 (항상 실행: 1회 호출)**:

- Q1: 언어 선택
- Q2: 에이전트 프롬프트 언어 선택
- Q3: 사용자 닉네임 입력

**추가 배치 (팀 모드 감지 시: 1회 호출)**:

- Q1: GitHub "Automatically delete head branches" 설정 확인
- Q2: SPEC Git 워크플로우 선택

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

#### 응답 처리 및 저장 방식

**중요**: 이 3개 질문의 응답은 **모두 .moai/config.json에 저장**되며, 이후 모든 Alfred 명령에서 참조됩니다.

**Q1: 대화 언어 선택**

선택된 옵션 → `.moai/config.json`에 저장:

```json
{
  "language": {
    "conversation_language": "ko",  // "en", "ja", "zh", "es"
    "conversation_language_name": "한국어"
  }
}
```

**영향 범위**:
- 🗣️ 모든 Alfred 대화 및 출력 (이 언어로 진행)
- 📄 생성되는 문서 (product.md, structure.md, tech.md, SPEC, 보고서 등)
- ❓ 이후 질문과 프롬프트 (모두 선택된 언어로)

---

**Q2: 에이전트 프롬프트 언어 선택** (NEW)

선택된 옵션 → `.moai/config.json`의 `language` 섹션에 저장:

```json
{
  "language": {
    "agent_prompt_language": "english"  // 또는 "localized"
  }
}
```

**옵션 설명**:

- **🌐 English (Global Standard)** → `"english"`
  - 모든 sub-agent 프롬프트가 **영어로 작성**됨
  - ✅ 장점: 코드 일관성, 팀 협업, 글로벌 표준
  - ✅ 권장: Claude Pro 사용자 (토큰 15-20% 절감)
  - project-manager, spec-builder 등이 영어로 작동

- **🗣️ Selected Language (Localized)** → `"localized"`
  - 모든 sub-agent 프롬프트가 **선택된 언어로 작성**됨
  - ✅ 장점: 로컬 팀 효율성, 네이티브 언어 편의성
  - project-manager도 한국어/일본어 등으로 작동

---

**Q3: 사용자 닉네임 입력**

사용자 입력 → `.moai/config.json`에 저장:

```json
{
  "user": {
    "nickname": "GOOS"  // 최대 20자
  }
}
```

**사용 예**:

- Alfred가 대화할 때: "안녕하세요, GOOS님"
- 생성 문서에: "Project Owner: GOOS"
- 로그: "User: GOOS | timestamp: 2025-11-04"

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

#### 응답 처리 및 저장 방식

**Q1 Response (GitHub 설정 - auto_delete_branches)**:

**저장될 config.json 구조**:

```json
{
  "github": {
    "auto_delete_branches": true,
    "auto_delete_branches_rationale": "PR 자동 병합 후 원격 브랜치 자동 정리로 저장소 관리 단순화"
  }
}
```

**각 옵션별 처리**:

| 선택지 | 저장값 | config.json | 영향 범위 | 팀 워크플로우에서의 의미 |
|--------|--------|-----------|---------|----------------------|
| ✅ Yes, already enabled | `true` | `"auto_delete_branches": true` | **최적화**: PR 자동 정리로 깔끔한 저장소 | 팀원이 여러 feature 브랜치를 만들 때, 병합 후 자동 정리되어 저장소 정리 불필요 |
| ❌ No, not enabled | `false` | `"auto_delete_branches": false` | **수동 관리**: 브랜치 정리를 수동으로 진행 | 브랜치를 수동으로 삭제해야 하므로 추가 작업 필요 |
| 🤔 Not sure | `null` | `"auto_delete_branches": null` | **경고**: 설정 확인 필요 | 워크플로우 진행은 가능하지만, 나중에 GitHub 설정 변경 권장 |

**사용 예**:

- git-manager가 branch cleanup 타이밍 결정:
  - `true` → PR 병합 후 자동으로 원격 브랜치 삭제
  - `false` → 로컬 정리 명령 제공 (`git branch -d`, `git push origin --delete`)
  - `null` → 사용자에게 수동 정리 알림

---

**Q2 Response (Git 워크플로우 - spec_git_workflow)**:

**저장될 config.json 구조**:

```json
{
  "github": {
    "spec_git_workflow": "feature_branch",
    "spec_git_workflow_rationale": "SPEC마다 feature 브랜치 생성으로 팀 리뷰 및 추적 가능한 워크플로우 확보"
  }
}
```

**각 옵션별 처리**:

| 선택지 | 저장값 | 동작 | `/alfred:1-plan` 시 | 팀 협업 영향 |
|--------|--------|------|-------------------|-----------|
| 📋 Feature Branch + PR | `"feature_branch"` | 매 SPEC마다 feature/SPEC-{ID} 브랜치 생성 → PR 리뷰 → develop 병합 | 1. 브랜치 자동 생성<br>2. PR 템플릿 생성<br>3. 리뷰자 설정<br>4. Merge 후 삭제 | ✅ 최적: 팀 리뷰, 코드 추적, 감사 이력 완벽<br>⚠️ 약간의 workflow 오버헤드 |
| 🔄 Direct Commit to Develop | `"develop_direct"` | develop 브랜치에 직접 커밋 (브랜치 생성 생략) | 1. 브랜치 생성 생략<br>2. 직접 develop 커밋<br>3. conflict 시 사용자 수동 해결 | ✅ 빠름: 프로토타입, 개인 프로젝트 적합<br>❌ 팀 리뷰 불가, 이력 추적 어려움 |
| 🤔 Decide per SPEC | `"per_spec"` | SPEC마다 git-manager가 워크플로우 선택 요청 | 1. AskUserQuestion으로 사용자 선택 요청<br>2. 선택에 따라 1번 또는 2번 경로 실행 | 🔀 유연함: SPEC 특성에 따라 선택 가능<br>⚠️ 매번 결정 필요한 오버헤드 |

**상세 동작 흐름**:

**Feature Branch + PR 선택 시** (`"feature_branch"`):
```
/alfred:1-plan SPEC-001 "Feature 설명"
  ↓
git-manager: feature/SPEC-001 브랜치 생성
  ↓
SPEC 문서 작성 및 커밋
  ↓
자동으로 PR 생성 (develop ← feature/SPEC-001)
  ↓
팀원들이 PR 리뷰
  ↓
승인 후 Merge (auto_delete_branches 설정에 따라 브랜치 정리)
```

**Direct Commit to Develop 선택 시** (`"develop_direct"`):
```
/alfred:1-plan SPEC-001 "Feature 설명"
  ↓
git-manager: develop 브랜치 확인
  ↓
SPEC 문서 작성 및 develop에 직접 커밋
  ↓
(PR 없음, 리뷰 없음)
```

**Decide per SPEC 선택 시** (`"per_spec"`):
```
/alfred:1-plan SPEC-001 "Feature 설명"
  ↓
AskUserQuestion: "이 SPEC의 git 워크플로우를 선택하세요"
  ├─ Feature Branch + PR 선택 → 위 "Feature Branch" 흐름
  └─ Direct Commit 선택 → 위 "Direct Commit" 흐름
```

---

**실제 저장되는 config.json 예시** (팀 모드):

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "localized"
  },
  "user": {
    "nickname": "GOOS"
  },
  "github": {
    "auto_delete_branches": true,
    "auto_delete_branches_rationale": "PR 병합 후 원격 브랜치 자동 정리로 저장소 관리 단순화",
    "spec_git_workflow": "feature_branch",
    "spec_git_workflow_rationale": "SPEC마다 feature 브랜치 생성으로 팀 리뷰 및 추적 가능한 워크플로우"
  }
}
```

**응답 저장 타임라인**:

1. 기본 배치 Q1-Q3 응답 → `.moai/config.json`의 `language`, `user` 섹션 저장
2. 팀 모드 추가 배치 Q1-Q2 응답 → `.moai/config.json`의 `github` 섹션 저장
3. Optional 도메인 선택 → `.moai/config.json`의 `stack.selected_domains` 저장

---

### 0.1.3.5 Report Generation Settings (All Modes - Optional)

**Purpose**: Control automatic report generation frequency to manage token usage and improve performance.

**When to ask**: After GitHub settings (team mode) or immediately after nickname (personal mode)

**Batched Design**: 1 question with token cost warning

**Important**: This question includes a detailed token warning to inform users about API costs before enabling automatic reports.

**Example AskUserQuestion Call**:

```python
AskUserQuestion(
    questions=[
        {
            "question": "How would you like to handle automatic report generation?\n\n⚠️ TOKEN COST WARNING:\n- Enable: ~50-60 tokens per report × 3-5 reports per command = 150-300 tokens/session\n- Minimal: ~20-30 tokens per report × 1-2 reports per command = 20-60 tokens/session\n- Disable: ~0 tokens (0 reports generated)\n\nFor Claude Pro $20 users: Token usage directly impacts API costs (~$0.02 per 1K tokens)",
            "header": "Report Generation",
            "multiSelect": false,
            "options": [
                {
                    "label": "📊 Enable (Default)",
                    "description": "Full analysis reports (50-60 tokens each). Best for detailed documentation. ~250-300 tokens/session"
                },
                {
                    "label": "⚡ Minimal (Recommended)",
                    "description": "Essential reports only, reduced output. ~40-60 tokens/session. 80% token reduction"
                },
                {
                    "label": "🚫 Disable",
                    "description": "No automatic reports. Fastest execution, zero report tokens. Manual generation available on request"
                }
            ]
        }
    ]
)
```

**Response Processing**:

The selected option determines `.moai/config.json` settings:

| Selection | Saved Setting | Config Value | Effect |
|-----------|---------------|--------------|--------|
| **📊 Enable** | `enabled: true, auto_create: true` | Full reports | Normal behavior: 3-5 reports per command |
| **⚡ Minimal** | `enabled: true, auto_create: false` | Essential only | 1-2 essential reports per command (~60-70% reduction) |
| **🚫 Disable** | `enabled: false, auto_create: false` | No reports | Zero report generation unless explicitly requested |

**Saved Configuration**:

```json
{
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "warn_user": true,
    "user_choice": "Minimal",
    "configured_at": "2025-11-04T12:00:00Z",
    "allowed_locations": [
      ".moai/docs/",
      ".moai/reports/",
      ".moai/analysis/",
      ".moai/specs/SPEC-*/"
    ]
  }
}
```

**Usage in sub-agents**:

- Alfred commands check `report_generation.enabled` before generating reports
- `/alfred:3-sync`: Respects this setting when creating sync reports
- `doc-syncer` agent: Skips report generation if `enabled: false`
- User can manually request reports at any time with explicit command (e.g., "generate analysis report")

---

### 0.1.4 Domain Selection (Optional - All Modes)

**Purpose**: Identify project domains to activate domain-expert agents for specialized guidance.

**When to ask**: After language/nickname/GitHub settings/report generation settings complete

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

## 🎛️ STEP 0-SETTING: Modify existing project settings (subcommand mode) - 명령형 지침

**Purpose**: User wants to change specific settings without re-running the full initialization.

**When to execute this step**:
- User runs `/alfred:0-project setting`
- OR User selected "🔧 Modify Settings" from the Already Initialized menu

### 0-SETTING.1 Load and display current configuration

**Your task**: Read `.moai/config.json` and show the user what their current settings are.

**Steps**:

1. **Read the configuration file**:
   - Load the file: `.moai/config.json`
   - Extract these values:
     * Language: `language.conversation_language` (e.g., "ko", "en")
     * Nickname: `user.nickname` (e.g., "GOOS")
     * Agent Prompt Language: `language.agent_prompt_language` (e.g., "localized")
     * GitHub auto-delete: `github.auto_delete_branches` (true/false)
     * SPEC workflow: `github.spec_git_workflow` (e.g., "feature_branch")
     * Report generation: `report_generation.user_choice` (e.g., "Minimal")
     * Domains: `stack.selected_domains` (e.g., ["frontend", "backend"])

2. **Display to the user**:
   ```
   ## Current Project Settings

   ✅ **Language**: [value from config.json]
   ✅ **Nickname**: [value from config.json]
   ✅ **Agent Prompt Language**: [value from config.json]
   ✅ **GitHub Auto-delete Branches**: [value from config.json]
   ✅ **SPEC Git Workflow**: [value from config.json]
   ✅ **Report Generation**: [value from config.json]
   ✅ **Selected Domains**: [value from config.json]
   ```

3. **Tell the user**: "Which settings would you like to modify?"

### 0-SETTING.2 Ask which settings to modify

**Your task**: Ask the user to select which settings they want to change.

**Tell the user**: "Select the settings you want to modify (you can choose multiple):"

**Present these 5 options** (allow multiple selections):

1. **"🌍 Language & Agent Prompt Language"** - Change conversation language or agent language
2. **"👤 Nickname"** - Change user nickname (max 20 characters)
3. **"🔧 GitHub Settings"** - Change auto-delete branches or SPEC git workflow
4. **"📊 Report Generation"** - Change report generation settings
5. **"🎯 Project Domains"** - Add or remove project domain selections

**Wait for the user to select one or more options**.

**After the user selects**, determine which sections were chosen:
- If "🌍 Language..." selected → Will ask in the LANGUAGE section
- If "👤 Nickname" selected → Will ask in the NICKNAME section
- If "🔧 GitHub Settings" selected → Will ask in the GITHUB section
- If "📊 Report Generation" selected → Will ask in the REPORTS section
- If "🎯 Project Domains" selected → Will ask in the DOMAINS section

### 0-SETTING.3 Collect new values (batched questions)

**Your task**: Based on user's selections from STEP 0-SETTING.2, ask for new values for ONLY the sections they selected.

**IMPORTANT**: Only ask questions for the sections the user selected. Skip all unselected sections.

---

#### Batch 1: LANGUAGE Section

**IF user selected "🌍 Language & Agent Prompt Language"**:

**Ask the user these TWO questions together** (batched in one interaction):

**Question 1**: "Which conversation language do you prefer?"

**Present these options**:
- 🌍 **English** - Global standard, widest community support
- 🇰🇷 **한국어** - Korean language for conversation and reports
- 🇯🇵 **日本語** - Japanese language for conversation and reports
- 🇨🇳 **中文** - Chinese language for conversation and reports

**Question 2**: "Which agent prompt language should Alfred use?"

**Present these options**:
- 🌐 **English (Global Standard)** - All internal agent prompts in English (infrastructure stability)
- 🗣️ **Selected Language (Localized)** - Agent prompts in your conversation language (experimental)

**Wait for the user to answer BOTH questions**.

**Store both answers** for the config.json update in STEP 0-SETTING.4.

**Map user's language choice to language code**:
- "English" → `en`
- "한국어" → `ko`
- "日本語" → `ja`
- "中文" → `zh`

**Map agent prompt language choice to setting**:
- "English (Global Standard)" → `english`
- "Selected Language (Localized)" → `localized`

---

#### Batch 2: NICKNAME Section

**IF user selected "👤 Nickname"**:

**Ask the user**:

**Question**: "What is your new nickname?"

**Instructions to user**:
- Maximum 20 characters
- Used in commits, reports, and project documentation
- Examples: "GOOS", "GoosLab", "DevTeam", "Alex"

**Wait for the user to enter their new nickname**.

**Store the answer** for config.json update in STEP 0-SETTING.4.

**Validation**:
- If user enters text longer than 20 characters → Trim to 20 characters and notify user
- If user enters empty text → Keep current nickname (no change)

---

#### Batch 3: GITHUB Section

**IF user selected "🔧 GitHub Settings"**:

**Ask the user these TWO questions together** (batched in one interaction):

**Question 1**: "Enable GitHub auto-delete branches after PR merge?"

**Present these options**:
- ✅ **Yes, enable** - Automatically delete feature branches after successful PR merge
- ❌ **No, disable** - Keep feature branches after merge (manual cleanup)
- 🤔 **Keep current** - No change to this setting

**Question 2**: "Which SPEC git workflow should Alfred use?"

**Present these options**:
- 📋 **Feature Branch + PR** - Create feature branch for each SPEC, submit PR to develop
- 🔄 **Direct Commit to Develop** - Skip branches, commit directly to develop
- 🤔 **Decide per SPEC** - Ask user for workflow choice when creating each SPEC
- ⏸️ **Keep current** - No change to this setting

**Wait for the user to answer BOTH questions**.

**Store both answers** for config.json update in STEP 0-SETTING.4.

**Map user's choices to config values**:

For auto-delete branches:
- "Yes, enable" → `true`
- "No, disable" → `false`
- "Keep current" → Keep existing value (no change)

For SPEC git workflow:
- "Feature Branch + PR" → `feature_branch`
- "Direct Commit to Develop" → `develop_direct`
- "Decide per SPEC" → `per_spec`
- "Keep current" → Keep existing value (no change)

---

#### Batch 4: REPORTS Section

**IF user selected "📊 Report Generation"**:

**Ask the user**:

**Question**: "Update automatic report generation settings?"

**Present these options**:
- 📊 **Enable** - Full analysis reports (comprehensive, ~50-60 tokens per report)
- ⚡ **Minimal** (Recommended) - Essential reports only (~20-30 tokens per report, 80% token savings)
- 🚫 **Disable** - No automatic report generation (0 tokens, fastest)
- ⏸️ **Keep current** - No change to this setting

**Wait for the user to select an option**.

**Store the answer** for config.json update in STEP 0-SETTING.4.

**Map user's choice to config values**:
- "Enable" → `enabled: true`, `auto_create: true`, `user_choice: "Enable"`
- "Minimal" → `enabled: true`, `auto_create: false`, `user_choice: "Minimal"`
- "Disable" → `enabled: false`, `auto_create: false`, `user_choice: "Disable"`
- "Keep current" → Keep existing values (no change)

---

#### Batch 5: DOMAINS Section

**IF user selected "🎯 Project Domains"**:

**Ask the user**:

**Question**: "Select all project domains that apply to this project (multiple selections allowed):"

**Present these options** (allow multiple selections):
- 🎨 **Frontend** - Web UI, React, Vue, Angular, HTML/CSS
- ⚙️ **Backend** - APIs, servers, Python/Node/Java backends
- 🚀 **DevOps** - CI/CD, Docker, Kubernetes, infrastructure
- 🗄️ **Database** - SQL, NoSQL, data modeling, migrations
- 📊 **Data Science** - ML, analytics, data pipelines
- 📱 **Mobile** - iOS, Android, React Native, Flutter
- ⚡ **Clear all** - Remove all domain selections (start fresh)

**Wait for the user to select one or more domains**.

**Store the answer** for config.json update in STEP 0-SETTING.4.

**Map user's selections to config values**:
- User selects domains → Array of domain IDs (e.g., `["frontend", "backend"]`)
- User selects "Clear all" → Empty array `[]`
- User selects nothing (cancels) → Keep existing domains (no change)

**Domain ID mapping**:
- "Frontend" → `"frontend"`
- "Backend" → `"backend"`
- "DevOps" → `"devops"`
- "Database" → `"database"`
- "Data Science" → `"data_science"`
- "Mobile" → `"mobile"`

---

**After collecting all selected sections' answers**, proceed to STEP 0-SETTING.4 to update config.json.

### 0-SETTING.4 Update config.json with selected changes

**Your task**: Save only the settings the user changed to `.moai/config.json`. Preserve all unchanged fields.

---

#### Step 1: Load current config.json

1. **Read the file**: `.moai/config.json`
2. **Parse the JSON structure** into memory
3. **Keep all current values** for merge (do NOT discard anything)

**Error check**: If file doesn't exist or has invalid JSON, go to STEP 0-SETTING.6 (Error Handling).

---

#### Step 2: Merge user's new values into config

**For EACH section the user selected in STEP 0-SETTING.2**, update ONLY those fields:

---

##### IF user selected LANGUAGE section:

**From stored answers in STEP 0-SETTING.3 Batch 1**:

**Update these fields**:
- `language.conversation_language` = [mapped language code: "en", "ko", "ja", or "zh"]
- `language.conversation_language_name` = [display name: "English", "한국어", "日本語", or "中文"]
- `language.agent_prompt_language` = [mapped value: "english" or "localized"]

**DO NOT change**:
- All other fields in `language` section (preserve existing values)
- Any other top-level sections (`user`, `github`, `report_generation`, `stack`, etc.)

**Example merge**:
```json
// Before:
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "agent_prompt_language": "english"
  }
}

// User changed to English + Localized
// After:
{
  "language": {
    "conversation_language": "en",
    "conversation_language_name": "English",
    "agent_prompt_language": "localized"
  }
}
```

---

##### IF user selected NICKNAME section:

**From stored answer in STEP 0-SETTING.3 Batch 2**:

**Update this field**:
- `user.nickname` = [user's new nickname, trimmed to max 20 chars]

**DO NOT change**:
- Any other fields in `user` section
- Any other top-level sections

**Example merge**:
```json
// Before:
{
  "user": {
    "nickname": "GOOS",
    "email": "goos@example.com"
  }
}

// User changed nickname to "GoosLab"
// After:
{
  "user": {
    "nickname": "GoosLab",
    "email": "goos@example.com"  // preserved
  }
}
```

---

##### IF user selected GITHUB section:

**From stored answers in STEP 0-SETTING.3 Batch 3**:

**For auto-delete branches**:
- IF user chose "Yes, enable" → Update `github.auto_delete_branches` = `true`
- IF user chose "No, disable" → Update `github.auto_delete_branches` = `false`
- IF user chose "Keep current" → Do NOT change this field (keep existing value)

**For SPEC git workflow**:
- IF user chose "Feature Branch + PR" → Update `github.spec_git_workflow` = `"feature_branch"`
- IF user chose "Direct Commit to Develop" → Update `github.spec_git_workflow` = `"develop_direct"`
- IF user chose "Decide per SPEC" → Update `github.spec_git_workflow` = `"per_spec"`
- IF user chose "Keep current" → Do NOT change this field (keep existing value)

**DO NOT change**:
- Any other fields in `github` section
- Any other top-level sections

**Example merge**:
```json
// Before:
{
  "github": {
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch",
    "token": "ghp_xxx"
  }
}

// User changed workflow to "develop_direct", kept auto-delete current
// After:
{
  "github": {
    "auto_delete_branches": true,  // preserved (user chose "Keep current")
    "spec_git_workflow": "develop_direct",  // updated
    "token": "ghp_xxx"  // preserved
  }
}
```

---

##### IF user selected REPORTS section:

**From stored answer in STEP 0-SETTING.3 Batch 4**:

**IF user chose "Enable"**:
- Update `report_generation.enabled` = `true`
- Update `report_generation.auto_create` = `true`
- Update `report_generation.user_choice` = `"Enable"`
- Update `report_generation.updated_at` = [current ISO timestamp]

**IF user chose "Minimal"**:
- Update `report_generation.enabled` = `true`
- Update `report_generation.auto_create` = `false`
- Update `report_generation.user_choice` = `"Minimal"`
- Update `report_generation.updated_at` = [current ISO timestamp]

**IF user chose "Disable"**:
- Update `report_generation.enabled` = `false`
- Update `report_generation.auto_create` = `false`
- Update `report_generation.user_choice` = `"Disable"`
- Update `report_generation.updated_at` = [current ISO timestamp]

**IF user chose "Keep current"**:
- Do NOT change ANY fields in `report_generation` section

**DO NOT change**:
- Field `report_generation.allowed_locations` (always preserve)
- Any other top-level sections

---

##### IF user selected DOMAINS section:

**From stored answer in STEP 0-SETTING.3 Batch 5**:

**IF user selected one or more domains**:
- Update `stack.selected_domains` = [array of domain IDs user selected]
- Update `stack.domain_selection_date` = [current ISO timestamp]

**IF user selected "Clear all"**:
- Update `stack.selected_domains` = `[]` (empty array)
- Update `stack.domain_selection_date` = [current ISO timestamp]

**IF user selected nothing (cancelled)**:
- Do NOT change ANY fields in `stack` section

**DO NOT change**:
- Any other fields in `stack` section
- Any other top-level sections

**Example merge**:
```json
// Before:
{
  "stack": {
    "selected_domains": ["frontend"],
    "domain_selection_date": "2025-01-01T00:00:00Z"
  }
}

// User selected: ["frontend", "backend", "database"]
// After:
{
  "stack": {
    "selected_domains": ["frontend", "backend", "database"],  // updated
    "domain_selection_date": "2025-11-04T10:30:00Z"  // updated with current time
  }
}
```

---

#### Step 3: Apply merge strategy (CRITICAL)

**IMPORTANT**: Follow this merge strategy EXACTLY:

1. **Start with the original config.json** (loaded in Step 1)
2. **Apply ONLY the changes for sections user selected** (from Step 2)
3. **Preserve ALL unchanged sections completely** (no modifications)

**Verification checklist before writing**:
- [ ] User selected LANGUAGE? → Only `language` section modified
- [ ] User selected NICKNAME? → Only `user.nickname` field modified
- [ ] User selected GITHUB? → Only changed fields in `github` section modified
- [ ] User selected REPORTS? → Only `report_generation` section modified
- [ ] User selected DOMAINS? → Only `stack.selected_domains` and `stack.domain_selection_date` modified
- [ ] All unselected sections → 100% preserved (exact copy from original)

**Example full merge**:

```json
// Original config.json
{
  "language": { "conversation_language": "ko" },
  "user": { "nickname": "GOOS" },
  "github": { "auto_delete_branches": true },
  "report_generation": { "enabled": true },
  "stack": { "selected_domains": ["frontend"] }
}

// User selected: LANGUAGE + NICKNAME sections only
// Changed: conversation_language to "en", nickname to "GoosLab"

// Merged config.json (correct)
{
  "language": { "conversation_language": "en" },  // ✅ updated
  "user": { "nickname": "GoosLab" },  // ✅ updated
  "github": { "auto_delete_branches": true },  // ✅ preserved (not selected)
  "report_generation": { "enabled": true },  // ✅ preserved (not selected)
  "stack": { "selected_domains": ["frontend"] }  // ✅ preserved (not selected)
}
```

---

#### Step 4: Write updated config.json to disk

1. **Combine original config + new values** using the merge strategy from Step 3
2. **Format the JSON** with proper indentation (2 spaces)
3. **Write the merged JSON** to `.moai/config.json`
4. **Verify the write succeeded** (check file exists and is valid JSON)

**If write fails**:
- Print error: "Failed to write config.json"
- Go to STEP 0-SETTING.6 (Error Handling)

**If write succeeds**:
- Proceed to STEP 0-SETTING.5 (Completion Report)

### 0-SETTING.5 Completion report

**Your task**: Display a completion report showing what was changed.

---

#### Step 1: Print header

**Print to user**:
```
✅ Settings update completed!

📝 Modified settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

#### Step 2: Show changes for each modified section

**For EACH section the user selected in STEP 0-SETTING.2**, display the before/after values:

---

##### IF LANGUAGE section was modified:

**Print**:
```
🌍 Language Settings:
- Conversation Language: [old_language_name] ([old_code]) → [new_language_name] ([new_code])
- Agent Prompt Language: [old_agent_language] → [new_agent_language]
```

**Example**:
```
🌍 Language Settings:
- Conversation Language: 한국어 (ko) → English (en)
- Agent Prompt Language: English (Global Standard) → Selected Language (Localized)
```

**Special case - No change made**:

IF user selected LANGUAGE but all answers were "Keep current":

**Print**:
```
🌍 Language Settings:
- No changes (kept current settings)
```

---

##### IF NICKNAME section was modified:

**Print**:
```
👤 Nickname:
- [old_nickname] → [new_nickname]
```

**Example**:
```
👤 Nickname:
- GOOS → GoosLab
```

**Special case - No change made**:

IF user entered empty text or same nickname:

**Print**:
```
👤 Nickname:
- No change (kept current nickname)
```

---

##### IF GITHUB section was modified:

**Print**:
```
🔧 GitHub Settings:
- Auto-delete Branches: [old_value] → [new_value]
- SPEC Git Workflow: [old_workflow] → [new_workflow]
```

**For fields where user chose "Keep current"**, show:
```
- Auto-delete Branches: [current_value] (no change)
```

**Example**:
```
🔧 GitHub Settings:
- Auto-delete Branches: true (no change)
- SPEC Git Workflow: feature_branch → develop_direct
```

**Special case - No changes made**:

IF user chose "Keep current" for ALL GitHub settings:

**Print**:
```
🔧 GitHub Settings:
- No changes (kept current settings)
```

---

##### IF REPORTS section was modified:

**Print**:
```
📊 Report Generation:
- Setting: [old_choice] → [new_choice]
- Status: [enabled/disabled]
- Auto-create: [true/false]
```

**Example**:
```
📊 Report Generation:
- Setting: Enable → Minimal
- Status: enabled
- Auto-create: false
```

**Special case - No change made**:

IF user chose "Keep current":

**Print**:
```
📊 Report Generation:
- No changes (kept current settings)
```

---

##### IF DOMAINS section was modified:

**Print**:
```
🎯 Project Domains:
- Selected: [list of domain names]
- Previous: [list of old domain names]
```

**Example**:
```
🎯 Project Domains:
- Selected: Frontend, Backend, Database
- Previous: Frontend
```

**Special case - Cleared all**:

IF user selected "Clear all":

**Print**:
```
🎯 Project Domains:
- Selected: (none)
- Previous: [list of old domains]
```

**Special case - No change made**:

IF user cancelled or selected nothing:

**Print**:
```
🎯 Project Domains:
- No changes (kept current domains)
```

---

#### Step 3: Print sections NOT modified

**DO NOT print anything for sections the user did NOT select**.

**Example**:
- User only selected LANGUAGE and NICKNAME
- Do NOT show GITHUB, REPORTS, or DOMAINS sections in the report

---

#### Step 4: Print file save confirmation

**Print**:
```

💾 Configuration saved to `.moai/config.json`
```

---

#### Step 5: Print next steps

**Print**:
```

📋 Next steps:
1. Review the changes above
2. Continue development with updated settings
3. Run `/alfred:0-project setting` again if you need to modify more settings
4. Run `/alfred:1-plan` to create a new SPEC with your updated configuration
```

---

#### Step 6: End this command

**Stop execution**. Do NOT proceed to STEP 1 or any other workflow.

The `/alfred:0-project setting` subcommand is now complete.

---

### 0-SETTING.6 Error handling

**CRITICAL**: Check for errors BEFORE starting STEP 0-SETTING.1.

---

#### Error 1: config.json not found

**Check BEFORE STEP 0-SETTING.1**: Does `.moai/config.json` exist?

**IF file does NOT exist**:

**Print to user**:
```
❌ Error: .moai/config.json not found

This command requires an already-initialized project.

To initialize first, run:
  /alfred:0-project

(without the "setting" subcommand)
```

**Action**: Exit immediately. Stop this command. Do NOT proceed to any other steps.

---

#### Error 2: Invalid JSON in config.json

**Check BEFORE STEP 0-SETTING.1**: Can `.moai/config.json` be parsed as valid JSON?

**IF file has syntax errors** (cannot be parsed):

**Print to user**:
```
❌ Error: config.json has syntax errors

The file exists but contains invalid JSON syntax.

Please fix the JSON manually, or restore from backup:
  Backup location: .moai-backups/

You can also run:
  cat .moai/config.json | jq .

to see the specific JSON error.
```

**Action**: Exit immediately. Stop this command. Do NOT proceed to any other steps.

---

#### Error 3: No settings selected

**Check in STEP 0-SETTING.2**: Did the user select any settings to modify?

**IF user clicked "Cancel" or selected nothing** (empty selection):

**Print to user**:
```
✅ No settings selected. Exiting without changes.

Your project configuration remains unchanged.

To modify settings later, run:
  /alfred:0-project setting
```

**Action**: Exit immediately. Stop this command. Do NOT proceed to STEP 0-SETTING.3 or beyond.

---

#### Error 4: Failed to write config.json

**Check in STEP 0-SETTING.4 Step 4**: Did the write to `.moai/config.json` succeed?

**IF write operation failed**:

**Print to user**:
```
❌ Error: Failed to write config.json

The configuration update could not be saved.

Possible causes:
- File permissions issue
- Disk full
- File locked by another process

Please check file permissions:
  ls -la .moai/config.json

Your previous configuration is unchanged.
```

**Action**: Exit immediately. Stop this command. Do NOT proceed to STEP 0-SETTING.5.

---

**Error handling summary**:

1. Check Error 1 & 2 BEFORE starting STEP 0-SETTING.1
2. Check Error 3 in STEP 0-SETTING.2 (after user answers which settings to modify)
3. Check Error 4 in STEP 0-SETTING.4 Step 4 (after attempting to write file)

If ANY error occurs, show the error message and exit immediately. Do NOT continue the workflow.

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

When a backup is detected, call `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` to present a TUI decision:

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

### 1.5 User Approval with AskUserQuestion (REVIEW MODE ONLY)

**Execution Condition**: This section only executes when BOTH conditions are true:
1. User selects "New" (not "Merge" or "Skip")
2. `--review` flag provided (review mode)

**In immediate execution mode** (no `--review` flag), this section is skipped entirely and execution proceeds directly to STEP 2.

After Alfred generates the interview plan report (review mode only), call `AskUserQuestion` tool (documented in moai-alfred-ask-user-questions skill) to get explicit user approval before starting the interview.

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

### 2.1.1 Dynamic Prompt Translation by conversation_language

**CRITICAL**: The base prompt is written in English. At runtime, Alfred translates it to the user's `conversation_language` (any language) before passing to project-manager.

**Translation Flow**:

```
English base prompt (single source of truth)
    ↓
Alfred reads conversation_language from STEP 0 (any language)
    ↓
Translate English prompt to {{CONVERSATION_LANGUAGE}} (runtime, any language)
    ↓
Pass translated prompt to project-manager agent
```

**Supported Languages**:

- **English (en)**: English (original, no translation)
- **Any other language**: Automatically translated from English

Examples:
- Korean (ko) → English → 한국어
- Japanese (ja) → English → 日本語
- Chinese (zh) → English → 中文
- Spanish (es) → English → Español
- French (fr) → English → Français
- German (de) → English → Deutsch
- Portuguese (pt) → English → Português
- Russian (ru) → English → Русский
- Arabic (ar) → English → العربية
- Hindi (hi) → English → हिंदी
- **Any language supported by translation service**

**Key Design Principle**:

- ✅ **Single source of truth**: Only English version maintained
- ✅ **Any language support**: Not limited to pre-defined languages
- ✅ **Runtime translation**: Translate on-demand for each user's selected language
- ✅ **Zero maintenance**: New language automatically supported

### 2.1.2 Base Prompt (English - Source of Truth)

```
Call the Task tool:
- subagent_type: "project-manager"
- description: "Initialize project with conversation language support"
- prompt: """You are the project-manager agent.

Language Settings:
- Conversation Language: {{CONVERSATION_LANGUAGE}} (used for all dialogs, documents)
- Language Name: {{CONVERSATION_LANGUAGE_NAME}}
- Agent Prompt Language: {{AGENT_PROMPT_LANGUAGE}} (internal sub-agent communication)

Agent Prompt Language Behavior:

1. **agent_prompt_language = "english"** (Global Standard):
   - You (project-manager) think and work in **English**
   - All internal analysis and planning done in English
   - Generated product.md, structure.md, tech.md written in **{{CONVERSATION_LANGUAGE}}**
   - Prompts to sub-agents (spec-builder, etc.) are in **English**

2. **agent_prompt_language = "localized"** (Localized):
   - You (project-manager) think and work in **{{CONVERSATION_LANGUAGE}}**
   - All internal analysis and planning done in {{CONVERSATION_LANGUAGE}}
   - Generated product.md, structure.md, tech.md written in **{{CONVERSATION_LANGUAGE}}**
   - Prompts to sub-agents (spec-builder, etc.) are in **{{CONVERSATION_LANGUAGE}}**

Important: conversation_language and agent_prompt_language can be different!
- conversation_language is used for **user dialogs**, **generated documents**
- agent_prompt_language is used for **sub-agent communication**, **internal prompts**

Git Workflow Configuration (Team Mode):
- spec_git_workflow: [feature_branch | develop_direct | per_spec]
  - "feature_branch": Create feature/spec-* branch, PR-based review, merge to develop
  - "develop_direct": Commit directly to develop, no branch creation
  - "per_spec": Ask user per SPEC during /alfred:1-plan execution
- Reference: Save this to .moai/config.json github.spec_git_workflow for git-manager

Project Type: [new|existing]
Detected Languages: [List of detected codebase languages]

Critical Directives:
All interviews and generated documents must be written in conversation_language ({{CONVERSATION_LANGUAGE}}):
- product.md: Generate in {{CONVERSATION_LANGUAGE}}
- structure.md: Generate in {{CONVERSATION_LANGUAGE}}
- tech.md: Generate in {{CONVERSATION_LANGUAGE}}

When conversation_language is 'en': Write all content in English
When conversation_language is 'ko': Write all content in Korean
When conversation_language is 'ja': Write all content in Japanese
For other languages: Follow the specified language

After project initialization, update .moai/config.json with announcements (base in English, runtime-translated):

```json
{
  "language": {
    "conversation_language": "{{CONVERSATION_LANGUAGE}}",
    "conversation_language_name": "{{CONVERSATION_LANGUAGE_NAME}}",
    "agent_prompt_language": "{{AGENT_PROMPT_LANGUAGE}}"
  },
  "github": {
    "spec_git_workflow": "[feature_branch|develop_direct|per_spec]"
  },
  "announcements": {
    "enabled": true,
    "language": "{{CONVERSATION_LANGUAGE}}",
    "items": [
      "🎩 SPEC-First: Always define requirements as SPEC before implementation (/alfred:1-plan)",
      "✅ TRUST 5 Principles: Test First, Readable, Unified, Secured, Trackable",
      "📝 TodoWrite Usage: Track all tasks and update in_progress/completed status immediately",
      "🌍 Language Boundary: Use conversation_language for dialogs/documents, English for infrastructure",
      "🔗 @TAG Chain: Maintain traceability SPEC→TEST→CODE→DOC",
      "⚡ Parallel Execution: Independent tasks can run simultaneously (Task tool parallel calls)",
      "💡 Skills First: Check appropriate Skill first for domain-specific tasks"
    ]
  }
}
```

### 2.1.3 Runtime Translation of Announcements

**Translation Logic**:

The `announcements.items` array in the base config (English) is **translated at runtime by Alfred** to `{{CONVERSATION_LANGUAGE}}`:

```
English base announcements (single source of truth)
    ↓
Alfred reads conversation_language from STEP 0 (any language)
    ↓
Translate each item from English to {{CONVERSATION_LANGUAGE}} (runtime)
    ↓
Save translated announcements to .claude/settings.json companyAnnouncements
    ↓
Claude Code displays announcements in user's language on startup
```

**Base Items** (Always English - Single Source of Truth):

The `announcements.items` in config.json is **always in English**. Translation happens at runtime via Alfred's translation pipeline to support any language.

**Example Translation Results**:

When user selects Korean (conversation_language = "ko"):

```
(Original English)
🎩 SPEC-First: Always define requirements as SPEC before implementation (/alfred:1-plan)

(Translated to Korean at runtime)
🎩 SPEC-First: 구현 전에 항상 요구사항을 SPEC으로 정의하세요 (/alfred:1-plan)
```

**Key Design Principle**:

- ✅ **Single source**: Only English announcements in config.json
- ✅ **Automatic translation**: Translates to user's language at runtime (any language)
- ✅ **Zero duplication**: No pre-translated copies maintained
- ✅ **Future-proof**: Any new language automatically supported without code changes

スキル 호출:
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

### 2.3 Sub-agent moai-alfred-ask-user-questions (Nested)

**The project-manager agent can internally call the TUI survey skill** to check the details of the task.

**When to call**:

- Before overwriting existing project documents
- When selecting language/framework
- When changing important settings

**Example** (inside project-manager): Ask whether to "overwrite file" with `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)`,

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
Before actual application, user confirmation is received with `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)`, and selected items are recorded in `CLAUDE.md` and `.moai/config.json`.

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
| Requires onboarding/training mode (`tech.md` STACK description, etc.)              | `moai-alfred-ask-user-questions`, `moai-adk-learning`, `agentic-coding` Output style                                 | Enhanced interview TUI and automatically provided onboarding materials |

> **Language/Domain Skill Selection Rules**
>
> - Select and add one relevant language skill (`moai-lang-python`, `moai-lang-java`, …) based on the `moai-alfred-language-detection` results or the stack recorded in the Tech section of the briefing.
> - Skills listed in the domain row are automatically included by cc-manager in the `selected_skills` list when the conditions are met.
> - The skill directory is always copied in its entirety, and only actual activation is recorded in `skill_pack` and `CLAUDE.md`.

If multiple conditions are met, the candidates are merged without duplicates and organized into sets of `candidate_agents`, `candidate_skills`, and `candidate_styles`.

#### 2.6.2 User confirmation flow

`AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` asks "whether to enable recommended items."

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

## 🚀 STEP 0-UPDATE: Template optimization (subcommand mode) - 명령형 지침

**When to execute this step**:
- User runs `/alfred:0-project update`
- OR user selected "Update Mode" from a menu
- OR after running `moai-adk update` command (when `optimized=false` in config.json)

**Your task**: After running `moai-adk update`, merge latest templates while preserving user customizations.

---

### STEP 0-UPDATE.1: Verify prerequisites and check backup

**Your task**: Verify that prerequisites exist before starting template optimization.

**Steps**:

1. **Check if backup directory exists**:
   - Directory to check: `.moai-backups/`
   - IF directory does NOT exist → Show error and exit:
     ```
     ❌ Error: No backup found at .moai-backups/

     This command requires a backup from previous initialization.
     Backup would have been created when you ran: /alfred:0-project

     Next steps:
     - IF this is a new project: Run /alfred:0-project
     - IF backup was deleted: Cannot recover, re-initialize project
     ```
   - IF directory exists → Continue to next step

2. **Find latest backup timestamp**:
   - Command: List subdirectories in `.moai-backups/`
   - Expected format: `.moai-backups/YYYYMMDD_HHMMSS/`
   - Find: Latest timestamp directory
   - Store: Timestamp value (e.g., "20250104_143022")
   - IF no timestamp directories found → Show error (same as step 1)

3. **Check if config.json exists in current directory**:
   - Read: `.moai/config.json`
   - IF file does NOT exist → Show error and exit:
     ```
     ❌ Error: .moai/config.json not found

     This command requires an initialized project.

     Next steps:
     - Run: /alfred:0-project
     - OR check if you are in the correct directory
     ```
   - IF file exists → Continue

4. **Print backup verification result**:
   ```
   ✅ Prerequisites verified

   📦 Backup found:
   - Location: .moai-backups/[TIMESTAMP]/
   - Timestamp: [TIMESTAMP]
   - Ready for template comparison
   ```

---

### STEP 0-UPDATE.2: Load and compare templates

**Your task**: Identify what changed between the old and new templates.

**Steps**:

1. **Load old template files from backup**:
   - Read: `.moai-backups/[LATEST_TIMESTAMP]/CLAUDE.md`
   - Read: `.moai-backups/[LATEST_TIMESTAMP]/.claude/settings.json`
   - Read: `.moai-backups/[LATEST_TIMESTAMP]/.moai/project/product.md`
   - Read: `.moai-backups/[LATEST_TIMESTAMP]/.moai/project/structure.md`
   - Read: `.moai-backups/[LATEST_TIMESTAMP]/.moai/project/tech.md`
   - Store: All old content in memory

2. **Load new template files from package**:
   - Read: `src/moai_adk/templates/CLAUDE.md`
   - Read: `src/moai_adk/templates/.claude/settings.json`
   - Read: `src/moai_adk/templates/.moai/project/product.md`
   - Read: `src/moai_adk/templates/.moai/project/structure.md`
   - Read: `src/moai_adk/templates/.moai/project/tech.md`
   - Store: All new content in memory

3. **Compare CLAUDE.md**:
   - Check: Is "## 🤖 Project Information" section present in old backup?
   - Check: Does old version have different structure/sections than new?
   - Identify: Custom content added by user (anything not in original template)
   - Store: Sections that need preservation

4. **Compare settings.json**:
   - Check: Custom environment variables in old backup
   - Check: Custom permissions in `permissions.allow` array
   - Check: Custom hooks in `hooks` section
   - Identify: User-added configurations
   - Store: Settings that need preservation

5. **Compare .moai/project/ documents**:
   - For each file (product.md, structure.md, tech.md):
     - Check: Does old version have user-written content?
     - Check: Is structure different from template?
     - Identify: Sections with real project data vs placeholder text
   - Store: Content sections that need preservation

6. **Create comparison summary**:
   - Count: How many files changed
   - Count: How many customizations found
   - Identify: Which files need smart merge vs simple overwrite

---

### STEP 0-UPDATE.3: Display comparison report and ask for approval

**Your task**: Show comparison results to user and get permission to proceed.

**Steps**:

1. **Print comparison report**:
   ```
   🔍 Template Comparison Analysis

   📊 Files analyzed:
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   - CLAUDE.md
   - .claude/settings.json
   - .moai/project/product.md
   - .moai/project/structure.md
   - .moai/project/tech.md

   🔧 Changes detected:
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   [IF CLAUDE.md has customizations]
   ✓ CLAUDE.md:
     - "## 🤖 Project Information" section (user-customized)
     - [List other custom sections if found]

   [IF settings.json has customizations]
   ✓ .claude/settings.json:
     - Custom permissions: [list count] items
     - Custom environment variables: [list count] items
     - Custom hooks: [list if any]

   [IF project docs have customizations]
   ✓ .moai/project/product.md: Has user-written content
   ✓ .moai/project/structure.md: Has user-written content
   ✓ .moai/project/tech.md: Has user-written content

   [IF NO customizations found in a file]
   - [filename]: No customizations (can be safely overwritten)

   💡 Recommendation:
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   - Smart merge will preserve your customizations
   - Latest template structure will be applied
   - Backup will be created before any changes
   ```

2. **Call AskUserQuestion tool**:
   ```python
   AskUserQuestion(
       questions=[
           {
               "question": "Template optimization analysis complete. How would you like to proceed?",
               "header": "📦 Template Optimization",
               "multiSelect": false,
               "options": [
                   {
                       "label": "✅ Proceed",
                       "description": "Run smart merge: preserve customizations with latest template structure"
                   },
                   {
                       "label": "👀 Preview",
                       "description": "Show detailed file-by-file changes before proceeding"
                   },
                   {
                       "label": "⏸️ Skip",
                       "description": "Keep current templates unchanged (you can run this command later)"
                   }
               ]
           }
       ]
   )
   ```

3. **Process user's response**:
   - Store: User's answer in variable `user_choice`
   - IF `user_choice == "Proceed"` → Go to STEP 0-UPDATE.4
   - IF `user_choice == "Preview"` → Go to STEP 0-UPDATE.3.1
   - IF `user_choice == "Skip"` → Go to STEP 0-UPDATE.7 (exit gracefully)

---

### STEP 0-UPDATE.3.1: Show detailed preview (conditional - only if user requested)

**Your task**: Show detailed file-by-file changes before merging.

**Steps**:

1. **For CLAUDE.md**:
   - Print:
     ```
     📄 FILE: CLAUDE.md
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

     🔸 SECTIONS TO PRESERVE (from your backup):
     - ## 🤖 Project Information
       [Show first 5 lines of this section from backup]

     🔸 NEW TEMPLATE STRUCTURE:
     - [List new sections added in latest template]

     🔸 MERGE STRATEGY:
     - Keep your "Project Information" section
     - Apply new template structure
     - Combine both into final CLAUDE.md
     ```

2. **For settings.json**:
   - Print:
     ```
     📄 FILE: .claude/settings.json
     ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

     🔸 CUSTOM PERMISSIONS (from your backup):
     [Show user's custom permissions array]

     🔸 CUSTOM ENVIRONMENT VARIABLES (from your backup):
     [Show user's custom env vars if any]

     🔸 MERGE STRATEGY:
     - Keep your custom permissions
     - Add new default permissions from template
     - Preserve your environment variables
     ```

3. **For .moai/project/ files**:
   - For each file (product.md, structure.md, tech.md):
     - Print:
       ```
       📄 FILE: .moai/project/[filename]
       ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

       🔸 YOUR CURRENT CONTENT (first 10 lines):
       [Show first 10 lines from backup]

       🔸 NEW TEMPLATE STRUCTURE:
       [Show first 10 lines from new template]

       🔸 MERGE STRATEGY:
       - Keep your written content
       - Apply new template structure/headings
       - Combine both into final document
       ```

4. **Ask for approval again**:
   ```python
   AskUserQuestion(
       questions=[
           {
               "question": "Preview complete. Ready to proceed with smart merge?",
               "header": "📦 Confirm Merge",
               "multiSelect": false,
               "options": [
                   {
                       "label": "✅ Proceed",
                       "description": "Apply smart merge now"
                   },
                   {
                       "label": "⏸️ Skip",
                       "description": "Cancel and keep current templates"
                   }
               ]
           }
       ]
   )
   ```

5. **Process second response**:
   - IF "Proceed" → Go to STEP 0-UPDATE.4
   - IF "Skip" → Go to STEP 0-UPDATE.7 (exit gracefully)

---

### STEP 0-UPDATE.4: Create safety backup before merge

**Your task**: Create a timestamped backup of current state before making any changes.

**Steps**:

1. **Generate new timestamp**:
   - Format: `YYYYMMDD_HHMMSS` (e.g., "20250104_153045")
   - Store: In variable `new_backup_timestamp`

2. **Create backup directory**:
   - Directory: `.moai-backups/[new_backup_timestamp]/`
   - IF directory creation fails → Show error and exit:
     ```
     ❌ Error: Cannot create backup directory

     Failed to create: .moai-backups/[TIMESTAMP]/

     Possible reasons:
     - Insufficient disk space
     - Permission denied
     - Invalid directory name

     Cannot proceed without safety backup.
     ```

3. **Copy current files to backup**:
   - Copy: `CLAUDE.md` → `.moai-backups/[new_backup_timestamp]/CLAUDE.md`
   - Copy: `.claude/settings.json` → `.moai-backups/[new_backup_timestamp]/.claude/settings.json`
   - Copy: `.moai/project/product.md` → `.moai-backups/[new_backup_timestamp]/.moai/project/product.md`
   - Copy: `.moai/project/structure.md` → `.moai-backups/[new_backup_timestamp]/.moai/project/structure.md`
   - Copy: `.moai/project/tech.md` → `.moai-backups/[new_backup_timestamp]/.moai/project/tech.md`
   - Copy: `.moai/config.json` → `.moai-backups/[new_backup_timestamp]/.moai/config.json`

4. **Verify backup integrity**:
   - Check: All files copied successfully
   - Check: Files are readable
   - IF any file missing → Show error and exit (do NOT proceed with merge)

5. **Print backup confirmation**:
   ```
   💾 Safety backup created

   Location: .moai-backups/[new_backup_timestamp]/
   Files backed up: 6

   [IF merge fails, you can restore from this backup]
   ```

---

### STEP 0-UPDATE.5: Execute smart merge

**Your task**: Merge new templates with user customizations.

**Steps**:

1. **Merge CLAUDE.md**:
   - Read: New template from `src/moai_adk/templates/CLAUDE.md`
   - Read: User's "## 🤖 Project Information" section from backup
   - Find: Location where "## 🤖 Project Information" should be inserted in new template
   - Insert: User's section into new template at correct location
   - Keep: All other sections from new template
   - Write: Merged content to `CLAUDE.md`
   - IF write fails → Go to STEP 0-UPDATE.6 (error recovery)

2. **Merge .claude/settings.json**:
   - Read: New template from `src/moai_adk/templates/.claude/settings.json`
   - Read: User's custom permissions from backup
   - Read: User's custom environment variables from backup
   - Merge strategy:
     ```
     {
       "hooks": [merge user's custom hooks with new defaults],
       "permissions": {
         "allow": [merge user's + new defaults, remove duplicates],
         "ask": [keep new defaults],
         "deny": [keep new defaults]
       },
       "environmentVariables": [merge user's custom vars with new defaults]
     }
     ```
   - Write: Merged settings.json
   - IF write fails → Go to STEP 0-UPDATE.6 (error recovery)

3. **Merge .moai/project/product.md**:
   - Read: New template from `src/moai_adk/templates/.moai/project/product.md`
   - Read: User's content from backup
   - Merge strategy:
     - Keep: New template section headings
     - Insert: User's written content under each heading
     - Preserve: User's custom sections not in template
   - Write: Merged product.md
   - IF write fails → Go to STEP 0-UPDATE.6 (error recovery)

4. **Merge .moai/project/structure.md**:
   - Same merge strategy as product.md
   - Write: Merged structure.md
   - IF write fails → Go to STEP 0-UPDATE.6 (error recovery)

5. **Merge .moai/project/tech.md**:
   - Same merge strategy as product.md
   - Write: Merged tech.md
   - IF write fails → Go to STEP 0-UPDATE.6 (error recovery)

6. **Print merge progress** (after each file):
   ```
   ✓ CLAUDE.md merged
   ✓ .claude/settings.json merged
   ✓ .moai/project/product.md merged
   ✓ .moai/project/structure.md merged
   ✓ .moai/project/tech.md merged
   ```

---

### STEP 0-UPDATE.5.1: Update config.json metadata

**Your task**: Mark the optimization as complete in config.json.

**Steps**:

1. **Read current config.json**:
   - Read: `.moai/config.json`
   - Parse: JSON content

2. **Update fields**:
   - Set: `project.optimized = true`
   - Set: `project.optimized_at = "[ISO_TIMESTAMP]"` (current timestamp in ISO 8601 format)
   - Set: `project.template_version = "[PACKAGE_VERSION]"` (from moai-adk package version)

3. **Add history entry**:
   - Append to `history` array:
     ```json
     {
       "date": "[ISO_TIMESTAMP]",
       "event": "Template optimization",
       "action": "Smart merge with user customizations",
       "backup": "[new_backup_timestamp]",
       "template_version": "[PACKAGE_VERSION]",
       "notes": "Updated to latest moai-adk template structure"
     }
     ```

4. **Write updated config.json**:
   - Write: Updated content to `.moai/config.json`
   - IF write fails → Show error (but merge already succeeded, so this is non-critical)

5. **Print config update confirmation**:
   ```
   ⚙️ config.json updated

   Changes:
   - optimized: true
   - optimized_at: [TIMESTAMP]
   - template_version: [VERSION]
   - history: Added optimization event
   ```

---

### STEP 0-UPDATE.5.2: Display completion report

**Your task**: Confirm to user that template optimization is complete.

**Print**:
```
✅ Template optimization completed!

📄 Merged files:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ CLAUDE.md
  - Latest template structure applied
  - Your "Project Information" section preserved

✓ .claude/settings.json
  - New default settings applied
  - Your custom permissions preserved
  - Your environment variables preserved

✓ .moai/project/product.md
  - Latest template structure applied
  - Your content preserved

✓ .moai/project/structure.md
  - Latest template structure applied
  - Your content preserved

✓ .moai/project/tech.md
  - Latest template structure applied
  - Your content preserved

⚙️ .moai/config.json
  - optimized: true
  - template_version: [VERSION]

💾 Safety backups:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Pre-merge backup: .moai-backups/[new_backup_timestamp]/
- Previous backup: .moai-backups/[old_timestamp]/

🎯 Next steps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Review merged templates:
   - cat CLAUDE.md
   - cat .claude/settings.json

2. Test your project:
   - /alfred:1-plan "test feature"

3. IF issues occur:
   - Restore from backup: .moai-backups/[new_backup_timestamp]/

✨ Your project is now using the latest moai-adk template!
```

**Then STOP this command** (do NOT proceed to STEP 1 or any other section).

---

### STEP 0-UPDATE.6: Error recovery (merge failure)

**Your task**: Handle errors during merge and restore from backup.

**When to execute**:
- IF any file write fails in STEP 0-UPDATE.5
- IF merge conflict cannot be resolved automatically

**Steps**:

1. **Identify which file failed**:
   - Store: Filename that caused error
   - Store: Error message

2. **Print error message**:
   ```
   ❌ Merge failed

   Failed file: [FILENAME]
   Error: [ERROR_MESSAGE]

   Possible reasons:
   - Permission denied (check file permissions)
   - Disk full (check available space)
   - File locked by another process

   🔄 Attempting automatic recovery...
   ```

3. **Restore from safety backup**:
   - Copy: `.moai-backups/[new_backup_timestamp]/[FAILED_FILE]` → `[FAILED_FILE]`
   - Copy: All other files from backup (to ensure consistency)
   - Print: "✓ Files restored from backup"

4. **Ask user for next steps**:
   ```python
   AskUserQuestion(
       questions=[
           {
               "question": "Merge failed and files were restored. What would you like to do?",
               "header": "⚠️ Error Recovery",
               "multiSelect": false,
               "options": [
                   {
                       "label": "🔍 Show error details",
                       "description": "Display full error message and failed file"
                   },
                   {
                       "label": "🔧 Manual merge",
                       "description": "I'll merge the files manually"
                   },
                   {
                       "label": "⏸️ Skip for now",
                       "description": "Keep current templates, try again later"
                   }
               ]
           }
       ]
   )
   ```

5. **Process user's choice**:
   - IF "Show error details":
     - Print: Full error message
     - Print: Backup location
     - Print: Manual merge instructions
   - IF "Manual merge" OR "Skip for now":
     - Exit command with status message

---

### STEP 0-UPDATE.7: Graceful exit (user skipped)

**Your task**: Exit the update command cleanly when user chooses to skip.

**When to execute**:
- IF user selected "Skip" in STEP 0-UPDATE.3
- IF user selected "Skip" in STEP 0-UPDATE.3.1

**Steps**:

1. **Print skip message**:
   ```
   ⏸️ Template optimization skipped

   No changes were made to your templates.

   Current state:
   - Templates: Using previous version
   - config.json: optimized = false
   - Backup: .moai-backups/[LATEST_TIMESTAMP]/

   💡 You can run template optimization later:
   - Command: /alfred:0-project update
   - OR run: moai-adk update

   ✨ Your project continues to work normally.
   ```

2. **STOP this command** (do NOT proceed to any other steps)
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
