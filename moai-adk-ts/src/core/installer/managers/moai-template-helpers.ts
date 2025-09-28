/**
 * @FEATURE:MOAI-TEMPLATES-001 MoAI 특화 템플릿 헬퍼
 *
 * MoAI 프로젝트에 특화된 템플릿 생성 및 처리 유틸리티
 * @DESIGN:SEPARATE-CONCERNS-001 TemplateProcessor에서 분리하여 단일 책임 원칙 준수
 */

/**
 * @DESIGN:CLASS-001 MoAI 특화 템플릿 헬퍼 클래스
 *
 * SPEC 문서, Claude Code 설정, 프로젝트 구조 등 MoAI 생태계에 특화된 템플릿 생성
 */
export class MoAITemplateHelpers {
  /**
   * @API:SPEC-TEMPLATE-001 SPEC 문서 템플릿 생성
   *
   * MoAI-ADK의 SPEC 문서 표준 형식을 따르는 템플릿 생성
   */
  static generateSpecTemplate(_specId: string, _title: string): string {
    return `# SPEC-{{SPEC_ID}}: {{SPEC_TITLE}}

## @REQ:{{SPEC_ID}}-001 Requirements
{{#REQUIREMENTS}}
- {{.}}
{{/REQUIREMENTS}}

## @DESIGN:{{SPEC_ID}}-001 Design
{{DESIGN_DESCRIPTION}}

{{#HAS_TDD}}
## @TEST:{{SPEC_ID}}-001 Test Strategy
{{TEST_STRATEGY}}
{{/HAS_TDD}}

## @TASK:{{SPEC_ID}}-001 Implementation Tasks
{{#TASKS}}
- [ ] {{.}}
{{/TASKS}}`;
  }

  /**
   * @API:CLAUDE-CONFIG-001 Claude Code 설정 템플릿 생성
   *
   * Claude Code 환경 설정을 위한 JSON 템플릿 생성
   */
  static generateClaudeConfigTemplate(): string {
    return `{
  "outputStyle": "{{OUTPUT_STYLE}}",
  "agents": {
    {{#ENABLED_AGENTS}}
    "{{name}}": {
      "enabled": true{{#config}},
      {{#.}}
      "{{@key}}": "{{.}}"{{#hasNext}},{{/hasNext}}
      {{/.}}
      {{/config}}
    }{{#hasNext}},{{/hasNext}}
    {{/ENABLED_AGENTS}}
  },
  "commands": {
    {{#ENABLED_COMMANDS}}
    "{{name}}": {
      "enabled": true
    }{{#hasNext}},{{/hasNext}}
    {{/ENABLED_COMMANDS}}
  }
}`;
  }

  /**
   * @API:PROJECT-STRUCTURE-001 프로젝트 구조 템플릿 생성
   *
   * MoAI 프로젝트의 표준 디렉토리 구조 템플릿 생성
   */
  static generateProjectStructureTemplate(): string {
    return `{{PROJECT_NAME}}/
├── .claude/
{{#TEAM_MODE}}
├── .github/
│   └── workflows/
{{/TEAM_MODE}}
├── .moai/
│   ├── memory/
│   ├── project/
│   └── specs/
{{#TECH_STACK}}
├── {{stack_dir}}/
{{/TECH_STACK}}
├── tests/
└── README.md`;
  }

  /**
   * @API:GITIGNORE-TEMPLATE-001 .gitignore 템플릿 생성
   *
   * MoAI 프로젝트에 최적화된 .gitignore 파일 템플릿
   */
  static generateGitignoreTemplate(): string {
    return `# Dependencies
node_modules/
{{#HAS_PYTHON}}
__pycache__/
*.pyc
*.pyo
{{/HAS_PYTHON}}

# Build outputs
dist/
build/
{{#HAS_TYPESCRIPT}}
*.tsbuildinfo
{{/HAS_TYPESCRIPT}}

# Environment files
.env
.env.local
.env.*.local

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# MoAI specific
{{#EXCLUDE_MOAI_CACHE}}
.moai/cache/
{{/EXCLUDE_MOAI_CACHE}}
{{#EXCLUDE_CLAUDE_LOGS}}
.claude/logs/
{{/EXCLUDE_CLAUDE_LOGS}}

# Project specific
{{#CUSTOM_IGNORES}}
{{.}}
{{/CUSTOM_IGNORES}}`;
  }

  /**
   * @API:PACKAGE-JSON-001 package.json 템플릿 생성
   *
   * TypeScript/Node.js 프로젝트용 package.json 템플릿
   */
  static generatePackageJsonTemplate(): string {
    return `{
  "name": "{{PROJECT_NAME}}",
  "version": "{{VERSION}}",
  "description": "{{DESCRIPTION}}",
  "main": "{{MAIN_FILE}}",
  {{#HAS_CLI}}
  "bin": {
    "{{CLI_NAME}}": "{{CLI_PATH}}"
  },
  {{/HAS_CLI}}
  "scripts": {
    {{#SCRIPTS}}
    "{{name}}": "{{command}}"{{#hasNext}},{{/hasNext}}
    {{/SCRIPTS}}
  },
  "dependencies": {
    {{#DEPENDENCIES}}
    "{{name}}": "{{version}}"{{#hasNext}},{{/hasNext}}
    {{/DEPENDENCIES}}
  },
  "devDependencies": {
    {{#DEV_DEPENDENCIES}}
    "{{name}}": "{{version}}"{{#hasNext}},{{/hasNext}}
    {{/DEV_DEPENDENCIES}}
  },
  "keywords": [
    {{#KEYWORDS}}
    "{{.}}"{{#hasNext}},{{/hasNext}}
    {{/KEYWORDS}}
  ],
  "author": "{{AUTHOR}}",
  "license": "{{LICENSE}}"
}`;
  }

  /**
   * @API:DOCKER-TEMPLATE-001 Dockerfile 템플릿 생성
   *
   * MoAI 프로젝트용 Dockerfile 템플릿
   */
  static generateDockerfileTemplate(): string {
    return `{{#HAS_TYPESCRIPT}}
FROM node:{{NODE_VERSION}}-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM node:{{NODE_VERSION}}-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./

{{/HAS_TYPESCRIPT}}
{{#HAS_PYTHON}}
FROM python:{{PYTHON_VERSION}}-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

{{/HAS_PYTHON}}
{{#EXPOSE_PORT}}
EXPOSE {{PORT}}
{{/EXPOSE_PORT}}

{{#HAS_CLI}}
ENTRYPOINT ["{{CLI_COMMAND}}"]
{{/HAS_CLI}}
{{^HAS_CLI}}
CMD ["{{START_COMMAND}}"]
{{/HAS_CLI}}`;
  }

  /**
   * @API:README-TEMPLATE-001 README.md 템플릿 생성
   *
   * MoAI 프로젝트 표준 README 템플릿
   */
  static generateReadmeTemplate(): string {
    return `# {{PROJECT_NAME}}

{{DESCRIPTION}}

## Features

{{#FEATURES}}
- {{.}}
{{/FEATURES}}

## Installation

\`\`\`bash
{{INSTALL_COMMAND}}
\`\`\`

## Usage

{{#HAS_CLI}}
\`\`\`bash
{{CLI_NAME}} --help
\`\`\`
{{/HAS_CLI}}

{{#USAGE_EXAMPLES}}
### {{title}}

\`\`\`{{language}}
{{code}}
\`\`\`
{{/USAGE_EXAMPLES}}

## Development

{{#DEV_SETUP}}
### {{title}}

{{description}}

\`\`\`bash
{{commands}}
\`\`\`
{{/DEV_SETUP}}

## Contributing

{{CONTRIBUTING_GUIDE}}

## License

{{LICENSE}}

---

Generated with [MoAI-ADK](https://github.com/your-org/moai-adk) 🗿`;
  }

  /**
   * @API:WORKFLOW-TEMPLATE-001 GitHub Actions 워크플로우 템플릿 생성
   *
   * CI/CD 파이프라인용 GitHub Actions 템플릿
   */
  static generateWorkflowTemplate(): string {
    return `name: {{WORKFLOW_NAME}}

on:
  {{#TRIGGERS}}
  {{type}}:
    {{#branches}}
    branches: [{{.}}]
    {{/branches}}
  {{/TRIGGERS}}

jobs:
  {{#JOBS}}
  {{name}}:
    runs-on: {{runs_on}}
    {{#strategy}}
    strategy:
      matrix:
        {{#matrix}}
        {{key}}: [{{#values}}"{{.}}"{{#hasNext}}, {{/hasNext}}{{/values}}]
        {{/matrix}}
    {{/strategy}}

    steps:
      - uses: actions/checkout@v4

      {{#setup_node}}
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: {{node_version}}
          cache: 'npm'
      {{/setup_node}}

      {{#setup_python}}
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: {{python_version}}
      {{/setup_python}}

      {{#steps}}
      - name: {{name}}
        run: {{command}}
      {{/steps}}
  {{/JOBS}}`;
  }
}
