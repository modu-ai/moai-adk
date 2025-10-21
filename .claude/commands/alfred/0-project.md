---
name: alfred:0-project
description: Initialize project document - create product/structure/tech.md and set optimization for each language
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
> Interactive prompts rely on `Skill("moai-alfred-tui-survey")` so AskUserQuestion renders TUI selection menus for user surveys and approvals.

## 🎯 Command Purpose

Automatically analyzes the project environment to create/update product/structure/tech.md documents and configure language-specific optimization settings.

## 📋 Execution flow

1. **Environment Analysis**: Automatically detect project type (new/legacy) and language
2. **Establishment of interview strategy**: Select question tree suited to project characteristics
3. **User Verification**: Review and approve interview plan
4. **Create project documentation**: Create product/structure/tech.md
5. **Create configuration file**: config.json auto-configuration

## 🧠 Skill Loadout Overview

| Agent | Auto core skill | Conditional skills |
| ----- | ---------------- | ------------------ |
| project-manager | Skill("moai-alfred-language-detection") | Skill("moai-foundation-ears"), Skill("moai-foundation-langs"), Detected domain skill (예: Skill("moai-domain-backend")), Skill("moai-alfred-tag-scanning"), Skill("moai-alfred-trust-validation"), Skill("moai-alfred-tui-survey") |
| trust-checker | Skill("moai-alfred-trust-validation") | Skill("moai-alfred-tag-scanning"), Skill("moai-foundation-trust"), Skill("moai-alfred-code-reviewer"), Skill("moai-alfred-performance-optimizer"), Skill("moai-alfred-tui-survey") |

## 🔗 Associated Agent

- **Primary**: project-manager (📋 planner) - Dedicated to project initialization
- **Quality Check**: trust-checker (✅ Quality assurance lead) - Initial structural verification (optional)
- **Secondary**: None (standalone execution)

## 💡 Example of use

The user executes the `/alfred:8-project` command to analyze the project and create/update documents.

## Command Overview

It is a systematic initialization system that analyzes the project environment and creates/updates product/structure/tech.md documents.

- **Automatically detect language**: Automatically recognize Python, TypeScript, Java, Go, Rust, etc.
- **Project type classification**: Automatically determine new vs. existing projects
- **High-performance initialization**: Achieve 0.18 second initialization with TypeScript-based CLI
- **2-step workflow**: 1) Analysis and planning → 2) Execution after user approval

## How to use

The user executes the `/alfred:8-project` command to start analyzing the project and creating/updating documents.

**Automatic processing**:
- Update mode if there is an existing `.moai/project/` document
- New creation mode if there is no document
- Automatic detection of language and project type

## ⚠️ Prohibitions

**What you should never do**:

- ❌ Create a file in the `.claude/memory/` directory
- ❌ Create a file `.claude/commands/alfred/*.json`
- ❌ Unnecessary overwriting of existing documents
- ❌ Date and numerical prediction (“within 3 months”, “50% reduction”) etc.)
- ❌ Hypothetical scenarios, expected market size, future technology trend predictions

**Expressions to use**:

- ✅ “High/medium/low priority”
- ✅ “Immediately needed”, “step-by-step improvements”
- ✅ Current facts
- ✅ Existing technology stack
- ✅ Real problems

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
- `optimized: false` in `config.json` (immediately after reinitialization)

**Select user if backup exists**  
`Skill("moai-alfred-tui-survey")`를 호출해 다음 옵션이 포함된 TUI를 표시합니다.
- **Merge**: 백업 내용과 최신 템플릿을 병합 (권장)
- **New**: 백업을 무시하고 새 인터뷰 시작
- **Skip**: 현재 파일 유지(작업 종료)

**Response processing**:
- **"Merge"** → Proceed to Phase 1.1 (backup merge workflow)
- **"Create new"** → Proceed to Phase 1.2 (Project environment analysis) (existing process)
- **"Skip"** → End task

**No backup or optimized: true**:
- Proceed directly to Phase 1.0.5 (language selection)

---

### 1.0.5 Language Selection (New Session)

**Purpose**: Let the user choose which language to use for communication during project initialization and throughout MoAI-ADK workflows.

**When to run**:
- First-time project initialization (new project)
- After reinitialization (optimized: false)
- User explicitly requests language change

**STEP 1: Display language selection TUI**

Call `Skill("moai-alfred-tui-survey")` to show language selection menu:

```
Header: "Which language would you like to use?"
Type: Single Select (not multiSelect)

Options:
1. Korean (한국어)
2. English
```

**Supported languages and codes**:
- `ko` - Korean ✅ Recommended for Korean users
- `en` - English ✅ Recommended for international teams

**STEP 2: Store language selection in config.json**

After user selects language, store it in project config:

```json
{
  "project": {
    "language": "ko",
    "language_selected_at": "2025-10-21T22:30:00+09:00",
    "language_confirmed": true
  }
}
```

**STEP 3: Set output-style based on language**

Map language to corresponding output-style:

| Language | Code | Output-Style | Purpose |
|----------|------|--------------|---------|
| Korean   | ko   | study-with-alfred | Educational approach for Korean developers |
| English  | en   | agentic-coding | Agile coding approach for international teams |

Update `.claude/settings.json`:
```json
{
  "env": {
    "ALFRED_LANGUAGE": "ko",
    "ALFRED_OUTPUT_STYLE": "study-with-alfred"
  }
}
```

**STEP 4: Confirmation message**

Display confirmation to user in their selected language:

```
✅ Language setup complete

For Korean users:
- Communication language: Korean
- Output style: study-with-alfred
- Settings saved: .moai/config.json, .claude/settings.json

---

✅ Language setup complete

For English users:
- Communication language: English
- Output style: agentic-coding
- Settings saved: .moai/config.json, .claude/settings.json
```

**STEP 5: Proceed to next phase**

- If no backup or optimized: true → Continue to Phase 1.2 (project environment analysis)
- If backup exists and optimized: false → Continue to Phase 1.1 (backup merge workflow)

**Note**: Language can be changed anytime by:
1. Editing `.moai/config.json` manually
2. Running `/alfred:0-project` again and selecting a different language
3. Running `/output-style` to change display style independently

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
- "{{PROJECT_NAME}}", "{{PROJECT_DESCRIPTION}}", etc. Variable format
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

### 1.3 Establish interview strategy (when user selects “New”)

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

### 1.4 Generate Interview Plan Report (when user selects “Create New”)

**Format of plan to be presented to users**:

```markdown
## 📊 Project initialization plan: [PROJECT-NAME]

### Environmental Analysis Results
- **Project Type**: [New/Existing/Hybrid]
- **Languages ​​Detected**: [Language List]
- **Current Document Status**: [Completeness Rating 0-100%]
- **Structure Complexity**: [Simple/Medium/Complex]

### 🎯 Interview strategy
- **Question category**: Product Discovery / Structure / Tech
- **Expected number of questions**: [N (M required + K optional)]
- **Estimated time required**: [Time estimation]
- **Priority area**: [Focus on Areas to be covered]

### ⚠️ Notes
- **Existing document**: [Overwrite vs supplementation strategy]
- **Language settings**: [Automatic detection vs manual setting]
- **Configuration conflicts**: [Compatibility with existing config.json]

### ✅ Expected deliverables
- **product.md**: [Business requirements document]
- **structure.md**: [System architecture document]
- **tech.md**: [Technology stack and policy document]
- **config.json**: [Project configuration file]

---
**Approval Request**: Would you like to proceed with the interview using the above plan?
 (Choose “Proceed,” “Modify [Content],” or “Abort”)
```

### 1.5 Wait for user approval (moai-alfred-tui-survey) (when user selects "New")

After Alfred receives the project-manager's interview plan report, `Skill("moai-alfred-tui-survey")`를 호출해 Phase 2 승인 여부를 묻습니다.
- **Proceed**: 승인된 계획대로 인터뷰 진행
- **Modify**: 계획 재수립 (Phase 1 재실행)
- **Stop**: 초기화 중단

**Response processing**:
- **"Progress"** (`answers["0"] === "Progress"`) → Execute Phase 2
- **"Modify"** (`answers["0"] === "Modify"`) → Repeat Phase 1 (recall project-manager)
- **"Abort"** (`answers["0"] === "Abort"`) → End task

---

## 🚀 STEP 2: Execute project initialization (after user approves “New”)

**Note**: This step will only be executed if the user selects **"New"**.
- When selecting "Merge": End the task in Phase 1.1 (Merge Backups)
- When selecting "Skip": End the task
- When selecting "New": Proceed with the process below

After user approval, the project-manager agent performs initialization.

### 2.1 Call project-manager agent (when user selects “New”)

Alfred starts project initialization by calling the project-manager agent. We will proceed based on the following information:
- Detected Languages: [Language List]
- Project Type: [New/Existing]
- Existing Document Status: [Existence/Absence]
- Approved Interview Plan: [Plan Summary]

Agents conduct structured interviews and create/update product/structure/tech.md documents.

### 2.2 Automatic activation of Alfred Skills (optional)

After the project-manager has finished creating the document, **Alfred can optionally call Skills** (upon user request).

**Automatic activation conditions** (optional):

| Conditions                           | Automatic selection Skill    | Purpose                                |
| ------------------------------------ | ---------------------------- | -------------------------------------- |
| User Requests “Quality Verification” | moai-alfred-trust-validation | Initial project structure verification |

**Execution flow** (optional):
```
1. project-manager completion
    ↓
2. User selection:
 - "Quality verification required" → moai-alfred-trust-validation (Level 1 quick scan)
 - "Skip" → Complete immediately
```

**Note**: Quality verification is optional during the project initialization phase.

### 2.3 Sub-agent moai-alfred-tui-survey (Nested)

**The project-manager agent can internally call the TUI survey skill** to check the details of the task.

**When to call**:
- Before overwriting existing project documents
- When selecting language/framework
- When changing important settings

**Example** (inside project-manager): `Skill("moai-alfred-tui-survey")`로 "파일 덮어쓰기" 여부를 묻고,
- **Overwrite** / **Merge** / **Skip** 중 선택하게 합니다.

**Nested pattern**:
- **Command level** (Phase approval): Called by Alfred → "Shall we proceed with Phase 2?"
- **Sub-agent level** (Detailed confirmation): Called by project-manager → "Shall we overwrite the file?"

### 2.4 Processing method by project type

#### A. New project (Greenfield)

**Interview Flow**:

1. **Product Discovery** (create product.md)
 - Define core mission (@DOC:MISSION-001)
 - Identify key user base (@SPEC:USER-001)
 - Identify key problems to solve (@SPEC:PROBLEM-001)
 - Summary of differences and strengths (@DOC:STRATEGY-001)
 - Setting success indicators (@SPEC:SUCCESS-001)

2. **Structure Blueprint** (create structure.md)
 - Selection of architecture strategy (@DOC:ARCHITECTURE-001)
 - Division of responsibilities by module (@DOC:MODULES-001)
 - External system integration plan (@DOC:INTEGRATION-001)
 - Define traceability strategy (@DOC:TRACEABILITY-001)

3. **Tech Stack Mapping** (written by tech.md)
 - Select language & runtime (@DOC:STACK-001)
 - Determine core framework (@DOC:FRAMEWORK-001)
 - Set quality gate (@DOC:QUALITY-001)
   - Define security policy (@DOC:SECURITY-001)
 - Plan distribution channels (@DOC:DEPLOY-001)

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
- Exclude build artifacts such as node_modules, .git, dist, build, __pycache__, etc.
- Identify key source directories and configuration files.

**Output**:
- Visualize the entire folder/file hierarchy of the project
- Identify major directories (src/, tests/, docs/, config/, etc.)
- Check language/framework hint files (package.json, pyproject.toml, go.mod, etc.)

**STEP 2: Establish parallel analysis strategy**

Alfred identifies groups of files by the Glob pattern:
1. **Configuration files**: *.json, *.toml, *.yaml, *.yml, *.config.js
2. **Source code files**: src/**/*.{ts,js,py,go,rs,java}
3. **Test files**: tests/**/*.{ts,js,py,go,rs,java}, **/*.test.*, **/*.spec.*
4. **Documentation files**: *.md, docs/**/*.md, README*, CHANGELOG*

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
 - Preservation of existing assets in “Legacy Context”

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
- Preserve conflicting content in the “Legacy Context” section
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
3. If necessary, readjust with /alfred:8-project update
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
- User choice: “Revalidate after fix” or “Skip”

**Skip verification**:
- Verification is not run by default
- Run only when explicitly requested by the user

### 2.6: Agent & Skill Tailoring (Project Optimization)

인터뷰와 초기 분석 결과를 바탕으로 프로젝트에서 즉시 활용해야 할 서브 에이전트와 스킬을 추천·활성화합니다.  
실제 적용 전에 `Skill("moai-alfred-tui-survey")`로 사용자 확인을 받고, 선택된 항목은 `CLAUDE.md`와 `.moai/config.json`에 기록합니다.

#### 2.6.0 cc-manager 브리핑 작성

문서 생성이 완료되면 **세 문서(product/structure/tech.md)를 모두 읽고** 다음 정보를 요약해 `cc_manager_briefing`이라는 텍스트를 만듭니다.

- `product.md`: 미션, 핵심 사용자, 해결해야 할 문제, 성공 지표, 백로그(TODO)를 원문 인용 또는 1줄 요약으로 정리합니다.
- `structure.md`: 아키텍처 유형, 모듈 경계와 담당 범위, 외부 연동, Traceability 전략, TODO 내용을 기록합니다.
- `tech.md`: 언어·프레임워크 버전, 빌드/테스트/배포 절차, 품질·보안 정책, 운영·모니터링 방식, TODO 항목을 정리합니다.

각 항목에는 반드시 출처(예: `product.md@SPEC:SUCCESS-001`)를 함께 적어 cc-manager가 근거를 파악할 수 있도록 합니다.

#### 2.6.1 cc-manager 판단 가이드

cc-manager는 브리핑을 바탕으로 필요한 서브 에이전트와 스킬을 선택합니다. 아래 표는 판단을 돕기 위한 참고용 가이드이며, 실제 호출 시에는 해당 문서의 근거 문장을 함께 전달합니다.

| 프로젝트 요구 상황 (문서 근거)      | 권장 서브 에이전트·스킬                                                    | 목적                                             |
| -------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------ |
| 품질·커버리지 목표가 높음 (`product.md@SPEC:SUCCESS-001`) | `tdd-implementer`, `moai-essentials-debug`, `moai-essentials-review`      | RED·GREEN·REFACTOR 워크플로우 정착               |
| Traceability/TAG 개선 요구 (`structure.md@DOC:TRACEABILITY-001`) | `doc-syncer`, `moai-alfred-tag-scanning`, `moai-alfred-trust-validation`  | TAG 추적성 강화 및 문서/코드 동기화              |
| 배포 자동화/브랜치 전략 필요 (`structure.md` Architecture/TODO) | `git-manager`, `moai-alfred-git-workflow`, `moai-foundation-git`          | 브랜치 전략·커밋 정책·PR 자동화                  |
| 레거시 모듈 리팩터링 (`product.md` BACKLOG, `tech.md` TODO) | `implementation-planner`, `moai-alfred-refactoring-coach`, `moai-essentials-refactor` | 기술 부채 진단 및 리팩터링 로드맵               |
| 규제/보안 준수 강화 (`tech.md@DOC:SECURITY-001`) | `quality-gate`, `moai-alfred-trust-validation`, `moai-foundation-trust`, `moai-domain-security` | TRUST S(Secured) 및 Trackable 준수, 보안 컨설팅 |
| CLI 자동화/툴링 요구 (`tech.md` BUILD/CLI 섹션) | `implementation-planner`, `moai-domain-cli-tool`, 감지된 언어 스킬(예: `moai-lang-python`) | CLI 명령 설계, 입력/출력 표준화                 |
| 데이터 분석/리포팅 요구 (`product.md` DATA, `tech.md` ANALYTICS) | `implementation-planner`, `moai-domain-data-science`, 감지된 언어 스킬     | 데이터 파이프라인·노트북 작업 정의              |
| 데이터베이스 구조 개선 (`structure.md` DB, `tech.md` STORAGE) | `doc-syncer`, `moai-domain-database`, `moai-alfred-tag-scanning`          | 스키마 문서화 및 TAG-DB 매핑 강화               |
| DevOps/인프라 자동화 필요 (`tech.md` DEVOPS, `structure.md` CI/CD) | `implementation-planner`, `moai-domain-devops`, `moai-alfred-git-workflow` | 배포 파이프라인 및 IaC 전략 수립                |
| ML/AI 기능 도입 (`product.md` AI, `tech.md` MODEL) | `implementation-planner`, `moai-domain-ml`, 감지된 언어 스킬              | 모델 학습/추론 파이프라인 정의                  |
| 모바일 앱 전략 (`product.md` MOBILE, `structure.md` CLIENT) | `implementation-planner`, `moai-domain-mobile-app`, 감지된 언어 스킬(예: `moai-lang-dart`, `moai-lang-swift`) | 모바일 클라이언트 구조 설계                     |
| 코딩 표준/리뷰 프로세스 강화 (`tech.md` REVIEW) | `quality-gate`, `moai-essentials-review`, `moai-alfred-code-reviewer`     | 리뷰 체크리스트 및 품질 보고 강화               |
| 온보딩/교육 모드 필요 (`tech.md` STACK 설명 등) | `moai-alfred-tui-survey`, `moai-adk-learning`, `agentic-coding` Output style | 인터뷰 TUI 강화 및 온보딩 자료 자동 제공      |

> **언어/도메인 스킬 선택 규칙**  
> - `moai-alfred-language-detection` 결과 또는 브리핑의 Tech 섹션에 기록된 스택을 기반으로 해당 언어 스킬(`moai-lang-python`, `moai-lang-java`, …) 한 개를 선택해 추가합니다.  
> - 도메인 행에 나열된 스킬은 상황이 충족될 때 cc-manager가 자동으로 `selected_skills` 목록에 포함시킵니다.  
> - 스킬 디렉터리는 항상 전체 복사되며, 실제 활성화 여부만 `skill_pack` 및 `CLAUDE.md`에 기록됩니다.

복수 조건이 충족되면 후보를 중복 없이 병합해 `candidate_agents`, `candidate_skills`, `candidate_styles` 집합으로 정리합니다.

#### 2.6.2 사용자 확인 흐름

`Skill("moai-alfred-tui-survey")`로 “추천 항목 활성화 여부”를 묻습니다.
- **모두 설치** / **선택 설치** / **설치 안 함** 세 가지 옵션을 제공하며,  
  “선택 설치”를 고르면 후보 목록을 다중 선택으로 다시 제시해 사용자가 필요한 항목만 고르도록 합니다.

#### 2.6.3 활성화 및 기록 단계

1. **브리핑 준비**: 사용자 선택(모두 설치/선택 설치) 결과와 `cc_manager_briefing` 전문을 정리합니다.  
2. **cc-manager 에이전트 호출**:  
   - `Task` 툴로 `subagent_type: "cc-manager"`를 호출하고, 브리핑과 사용자 선택 항목을 프롬프트에 포함합니다.  
   - cc-manager는 브리핑을 근거로 필요한 서브 에이전트와 스킬을 결정하고, `CLAUDE.md`, `.claude/agents/alfred/*.md`, `.claude/skills/*.md`를 프로젝트 맞춤형으로 복사·갱신합니다.
3. **구성 업데이트 확인**: cc-manager가 반영한 결과를 검토합니다.  
   - 서브 에이전트: `.claude/agents/alfred/` 템플릿을 활성 상태로 유지하고 `CLAUDE.md` “Agents” 섹션에 기재합니다.  
   - 스킬: `.claude/skills/` 문서를 확인한 뒤 `CLAUDE.md` “Skills” 섹션에 추가합니다.  
   - Output style: `.claude/output-styles/alfred/`를 적용하고 `CLAUDE.md` “Output Styles”에 활성화 사실을 기록합니다.  
4. **config.json 갱신**  
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
   기존 속성이 있을 경우 병합합니다.
5. **최종 보고**: Completion Report 상단에 “활성화된 서브 에이전트/스킬/스타일” 목록과 `cc_manager_briefing` 요약을 추가하고, 동일 내용을 `CLAUDE.md` 표에도 반영해 후속 명령에서 자동 탐색되도록 합니다.

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
Symptom: “Language not detected” message
Solution: Specify language manually or create language-specific settings file
```

**Error 2**: Conflict with existing document
```
Symptom: product.md already exists and has different contents
Solution: Preserve existing contents and add new contents in “Legacy Context” section
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
# Browse the latest backups in the .moai-backups/ directory
   ls -lt .moai-backups/ | head -1
   ```

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
   ```

4. **Waiting for user approval**  
   `Skill("moai-alfred-tui-survey")`로 “템플릿 최적화를 진행할까요?”를 묻고 다음 옵션을 제공한다.
   - **Proceed** → Phase 2 실행
   - **Preview** → 변경 내역을 표시 후 재확인
   - **Skip** → optimized=false 유지

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

📄 Merged files:
 - CLAUDE.md (preserves project information)
 - settings.json (preserves env variables)

⚙️ config.json: optimized=true Configuration complete
   ```

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
```

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
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
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
- **Project**: {{PROJECT_NAME}}
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
- User explicitly selects “Skip”
- “Simple project” when Alfred automatically determines (only basic features required)

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

## Related commands

- `/alfred:1-plan` - Start writing SPEC
- `/alfred:9-update` - MoAI-ADK update
- `moai doctor` - System diagnosis
- `moai status` - Check project status
