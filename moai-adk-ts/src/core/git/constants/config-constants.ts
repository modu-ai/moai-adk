// @CODE:REFACTOR-004 연결: @TEST:REFACTOR-004 -> @CODE:REFACTOR-004
/**
 * @file Git Configuration and Templates
 * @author MoAI Team
 * @tags @CODE:REFACTOR-004 @CODE:GIT-DEFAULTS-001:DATA
 * @description Git 기본 설정, .gitignore 템플릿, GitHub 설정, 타임아웃 상수
 */

/**
 * MoAI-ADK .gitignore 템플릿
 * @tags @CODE:GITIGNORE-TEMPLATES-001:DATA
 */
export const GitignoreTemplates = {
  MOAI: `# MoAI-ADK Generated .gitignore

# Logs and temporary files
.claude/logs/
.moai/logs/
*.log
*.tmp

# Backup directories
.moai-backup/

# Node.js
node_modules/
npm-debug.log*
yarn-debug.log*
.pnpm-debug.log*

# Python
__pycache__/
*.pyc
*.pyo
*.pyd
.Python
env/
venv/
pip-log.txt

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Environment variables
.env
.env.local
.env.development.local
.env.test.local
.env.production.local

# Test coverage
coverage/
.nyc_output/
*.lcov

# Build artifacts
dist/
build/
*.tgz
*.tar.gz

# Database
*.db
*.sqlite
*.sqlite3

# Jupyter Notebook
.ipynb_checkpoints

# pyenv
.python-version

# TypeScript
*.tsbuildinfo

# ESLint cache
.eslintcache

# Parcel-bundler cache
.cache
.parcel-cache

# Next.js
.next

# Nuxt.js
.nuxt

# Gatsby files
.cache/
public

# Serverless directories
.serverless/

# FuseBox cache
.fusebox/

# DynamoDB Local files
.dynamodb/
`,

  NODE: `# Node.js
node_modules/
npm-debug.log*
yarn-debug.log*
.pnpm-debug.log*

# Environment variables
.env
.env.local

# Build
dist/
build/

# Test coverage
coverage/
.nyc_output/
`,

  PYTHON: `# Python
__pycache__/
*.py[cod]
*$py.class
*.so

# Distribution / packaging
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/

# PyInstaller
*.manifest
*.spec

# Unit test / coverage reports
htmlcov/
.tox/
.coverage
.coverage.*
.cache
nosetests.xml
coverage.xml
*.cover
.hypothesis/
.pytest_cache/

# Environments
.env
.venv
env/
venv/
ENV/
env.bak/
venv.bak/

# mypy
.mypy_cache/
.dmypy.json
dmypy.json
`,
} as const;

/**
 * Git 기본 설정
 * @tags @CODE:GIT-DEFAULTS-001:DATA
 */
export const GitDefaults = {
  DEFAULT_BRANCH: 'main',
  DEFAULT_REMOTE: 'origin',
  COMMIT_MESSAGE_MAX_LENGTH: 72,
  DESCRIPTION_MAX_LENGTH: 100,

  /**
   * 기본 Git 설정
   */
  CONFIG: {
    'init.defaultBranch': 'main',
    'core.autocrlf': process.platform === 'win32' ? 'true' : 'input',
    'core.ignorecase': 'false',
    'pull.rebase': 'false',
    'push.default': 'current',
  },

  /**
   * 안전한 명령어 목록
   */
  SAFE_COMMANDS: [
    'status',
    'log',
    'diff',
    'show',
    'branch',
    'remote',
    'config',
    'ls-files',
  ],

  /**
   * 위험한 명령어 목록 (사용자 확인 필요)
   */
  DANGEROUS_COMMANDS: [
    'reset --hard',
    'clean -fd',
    'rebase -i',
    'push --force',
    'branch -D',
    'remote rm',
  ],
} as const;

/**
 * GitHub 설정
 * @tags @CODE:GITHUB-DEFAULTS-001:DATA
 */
export const GitHubDefaults = {
  API_BASE_URL: 'https://api.github.com',
  DEFAULT_BRANCH: 'main',

  /**
   * PR 템플릿
   */
  PR_TEMPLATE: `## Summary
- Brief description of changes

## Test Plan
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project conventions
- [ ] Documentation updated
- [ ] Breaking changes documented

🤖 Generated with [MoAI-ADK](https://github.com/your-org/moai-adk)
`,

  /**
   * Issue 템플릿
   */
  ISSUE_TEMPLATE: `## Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS:
- Node.js:
- MoAI-ADK:

🤖 Generated with [MoAI-ADK](https://github.com/your-org/moai-adk)
`,

  /**
   * 기본 라벨
   */
  DEFAULT_LABELS: [
    { name: 'bug', color: 'd73a4a', description: "Something isn't working" },
    {
      name: 'enhancement',
      color: 'a2eeef',
      description: 'New feature or request',
    },
    {
      name: 'documentation',
      color: '0075ca',
      description: 'Improvements or additions to documentation',
    },
    {
      name: 'good first issue',
      color: '7057ff',
      description: 'Good for newcomers',
    },
    {
      name: 'help wanted',
      color: '008672',
      description: 'Extra attention is needed',
    },
    {
      name: 'invalid',
      color: 'e4e669',
      description: "This doesn't seem right",
    },
    {
      name: 'question',
      color: 'd876e3',
      description: 'Further information is requested',
    },
    {
      name: 'wontfix',
      color: 'ffffff',
      description: 'This will not be worked on',
    },
  ],
} as const;

/**
 * Git 타임아웃 설정
 * @tags @CODE:GIT-TIMEOUTS-001:DATA
 */
export const GitTimeouts = {
  CLONE: 300000, // 5분
  FETCH: 120000, // 2분
  PUSH: 180000, // 3분
  COMMIT: 30000, // 30초
  STATUS: 10000, // 10초
  DEFAULT: 60000, // 1분
} as const;
