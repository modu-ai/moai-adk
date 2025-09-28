# MoAI-ADK TypeScript 마이그레이션 실행 계획

## 📋 Executive Summary

**목표:** MoAI-ADK를 Python에서 TypeScript로 전환하여 Claude Code 생태계와 완벽 통합
**기간:** 5주 (35일)
**예상 효과:** 사용자 설치 시간 90% 단축, 진입 장벽 100% 제거

---

## 🎯 마이그레이션 전체 개요

### 핵심 전환 사항
```
Python 기반 MoAI-ADK  →  TypeScript 기반 @moai/adk
├── 설치: pip install        →  npm install -g @moai/adk
├── 실행: moai (Python)      →  moai (Node.js)
├── 훅: python3 hooks/       →  node hooks/
└── 배포: PyPI               →  npm
```

### 성공 지표
- ✅ 모든 기존 기능 100% 포팅
- ✅ 설치 시간 3-5분 → 30초 (90% 개선)
- ✅ 사용자 설치 단계 2단계 → 1단계 (50% 감소)
- ✅ 타입 안전성 0% → 100% (완전 개선)

---

## 📅 상세 마이그레이션 일정

### Week 1: 기반 설정 (12/2 - 12/8)
**목표:** TypeScript 프로젝트 초기화 및 기본 CLI 구현

#### Day 1-2: 프로젝트 설정
```bash
# 새 TypeScript 프로젝트 생성
mkdir moai-adk-ts
cd moai-adk-ts
npm init -y

# TypeScript 개발 환경 설정
npm install -D typescript @types/node ts-node nodemon
npm install -D jest ts-jest @types/jest
npm install -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
npm install -D prettier

# 프로덕션 의존성 설치
npm install commander chalk inquirer fs-extra
```

**설정 파일 생성:**
```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "moduleResolution": "node",
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

#### Day 3-4: 기본 CLI 구조
```typescript
// src/cli/index.ts - 기본 CLI 프레임워크
import { Command } from 'commander';
import { version } from '../utils/version.js';

const program = new Command();

program
  .name('moai')
  .description('MoAI Agentic Development Kit for Claude Code')
  .version(version);

program
  .command('init <project-name>')
  .description('Initialize new MoAI project')
  .option('-m, --mode <mode>', 'Setup mode (personal|team)', 'personal')
  .action(initCommand);

program
  .command('config')
  .description('Configure MoAI settings')
  .action(configCommand);

program
  .command('update')
  .description('Update MoAI templates and tools')
  .action(updateCommand);

export { program };
```

#### Day 5-7: 핵심 유틸리티 포팅
```typescript
// src/utils/logger.ts - 구조화 로깅
export interface LogEntry {
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  timestamp: string;
  data?: any;
}

export class Logger {
  log(level: LogEntry['level'], message: string, data?: any): void {
    const entry: LogEntry = {
      level,
      message,
      timestamp: new Date().toISOString(),
      data
    };

    console.log(JSON.stringify(entry));
  }
}

// src/utils/file-operations.ts - 파일 작업
export class FileOperations {
  async ensureDirectory(dirPath: string): Promise<void> {
    await fs.ensureDir(dirPath);
  }

  async copyTemplate(source: string, destination: string, variables: Record<string, any>): Promise<void> {
    const content = await fs.readFile(source, 'utf-8');
    const rendered = this.renderTemplate(content, variables);
    await fs.writeFile(destination, rendered);
  }
}
```

**Week 1 산출물:**
- ✅ 실행 가능한 TypeScript CLI (`moai --help`)
- ✅ 기본 프로젝트 구조 완성
- ✅ 빌드 시스템 구축 (`npm run build`)
- ✅ 테스트 프레임워크 설정

---

### Week 2: 핵심 기능 포팅 (12/9 - 12/15)
**목표:** Python 핵심 기능을 TypeScript로 완전 포팅

#### Day 8-10: 설치 시스템 포팅
```typescript
// src/core/installer.ts
export class MoaiInstaller {
  private logger = new Logger();
  private fileOps = new FileOperations();

  async install(projectName: string, options: InstallOptions): Promise<void> {
    this.logger.log('info', 'Starting MoAI project installation', { projectName, options });

    // 1. 프로젝트 디렉토리 생성
    await this.createProjectStructure(projectName);

    // 2. 템플릿 파일 복사
    await this.copyTemplates(projectName, options);

    // 3. Claude Code 설정
    await this.setupClaudeCode(projectName, options);

    // 4. Git 초기화 (선택사항)
    if (options.git) {
      await this.initializeGit(projectName);
    }

    this.logger.log('info', 'MoAI project installation completed', { projectName });
  }

  private async createProjectStructure(projectName: string): Promise<void> {
    const directories = [
      `${projectName}/.claude/agents/moai`,
      `${projectName}/.claude/commands/moai`,
      `${projectName}/.claude/hooks/moai`,
      `${projectName}/.moai/project`,
      `${projectName}/.moai/specs`,
      `${projectName}/.moai/memory`,
      `${projectName}/.moai/scripts`
    ];

    for (const dir of directories) {
      await this.fileOps.ensureDirectory(dir);
    }
  }
}
```

#### Day 11-13: 훅 시스템 전환
```typescript
// src/hooks/pre-write-guard.ts
import { HookInput, HookOutput } from './types.js';

export class PreWriteGuard {
  private readonly SENSITIVE_KEYWORDS = ['.env', '/secrets', '/.git/', '/.ssh'];
  private readonly PROTECTED_PATHS = ['.moai/memory/'];

  execute(input: HookInput): HookOutput {
    const { file_path } = input;

    if (!file_path) {
      return { allowed: true, message: 'No file path specified' };
    }

    const allowed = this.checkFileSafety(file_path);

    return {
      allowed,
      message: allowed ? 'File access granted' : 'File access denied for security',
      details: allowed ? undefined : { reason: 'Sensitive file detected', file_path }
    };
  }

  private checkFileSafety(filePath: string): boolean {
    const pathLower = filePath.toLowerCase();

    // 민감한 키워드 검사
    for (const keyword of this.SENSITIVE_KEYWORDS) {
      if (pathLower.includes(keyword)) {
        return false;
      }
    }

    // 보호된 경로 검사
    for (const protected_path of this.PROTECTED_PATHS) {
      if (filePath.includes(protected_path)) {
        return false;
      }
    }

    return true;
  }
}

// CLI 실행 진입점
if (require.main === module) {
  const guard = new PreWriteGuard();
  const input = JSON.parse(process.argv[2] || '{}');
  const output = guard.execute(input);

  console.log(JSON.stringify(output));
  process.exit(output.allowed ? 0 : 1);
}
```

#### Day 14: 통합 테스트
```typescript
// tests/integration/cli.test.ts
describe('MoAI CLI Integration', () => {
  test('should initialize new project', async () => {
    const tempDir = await fs.mkdtemp('/tmp/moai-test-');

    await execAsync(`node dist/cli/index.js init test-project`, {
      cwd: tempDir
    });

    // 프로젝트 구조 검증
    expect(await fs.pathExists(`${tempDir}/test-project/.claude`)).toBe(true);
    expect(await fs.pathExists(`${tempDir}/test-project/.moai`)).toBe(true);
  });

  test('should execute hooks correctly', async () => {
    const hookResult = await execAsync(`node dist/hooks/pre-write-guard.js '{"file_path": "safe-file.txt"}'`);
    const result = JSON.parse(hookResult.stdout);

    expect(result.allowed).toBe(true);
  });
});
```

**Week 2 산출물:**
- ✅ 완전한 설치 시스템 포팅
- ✅ 모든 Python 훅을 TypeScript로 전환
- ✅ 단위 테스트 및 통합 테스트 작성
- ✅ 기본 기능 동작 검증

---

### Week 3: 고급 기능 및 최적화 (12/16 - 12/22)
**목표:** 고급 기능 구현 및 성능 최적화

#### Day 15-17: Git 관리 시스템
```typescript
// src/core/git-manager.ts
export class GitManager {
  async initializeRepository(projectPath: string): Promise<void> {
    await this.exec('git init', { cwd: projectPath });
    await this.exec('git add .', { cwd: projectPath });
    await this.exec('git commit -m "Initial commit with MoAI-ADK setup"', { cwd: projectPath });
  }

  async createBranch(branchName: string, projectPath: string): Promise<void> {
    await this.exec(`git checkout -b ${branchName}`, { cwd: projectPath });
  }

  async createCommit(message: string, projectPath: string): Promise<void> {
    await this.exec('git add .', { cwd: projectPath });
    await this.exec(`git commit -m "${message}"`, { cwd: projectPath });
  }

  private async exec(command: string, options: { cwd: string }): Promise<string> {
    return new Promise((resolve, reject) => {
      require('child_process').exec(command, options, (error, stdout, stderr) => {
        if (error) reject(error);
        else resolve(stdout);
      });
    });
  }
}
```

#### Day 18-19: TAG 시스템 포팅
```typescript
// src/core/tag-system.ts
export interface TagEntry {
  id: string;
  type: 'REQ' | 'DESIGN' | 'TASK' | 'TEST' | 'FEATURE' | 'API' | 'DATA';
  content: string;
  file_path: string;
  line_number: number;
  timestamp: string;
  relationships: string[];
}

export class TagSystem {
  private db: Database;

  async addTag(tag: TagEntry): Promise<void> {
    await this.db.run(
      'INSERT INTO tags (id, type, content, file_path, line_number, timestamp) VALUES (?, ?, ?, ?, ?, ?)',
      [tag.id, tag.type, tag.content, tag.file_path, tag.line_number, tag.timestamp]
    );
  }

  async findTagChain(startTagId: string): Promise<TagEntry[]> {
    // 16-Core TAG 체인 추적 로직
    const chain: TagEntry[] = [];
    let currentId = startTagId;

    while (currentId) {
      const tag = await this.findTag(currentId);
      if (tag) {
        chain.push(tag);
        currentId = tag.relationships[0]; // 다음 TAG로 이동
      } else {
        break;
      }
    }

    return chain;
  }
}
```

#### Day 20-21: 성능 최적화
```typescript
// src/core/template-cache.ts
export class TemplateCache {
  private cache = new Map<string, string>();
  private watchedFiles = new Set<string>();

  async getTemplate(templatePath: string): Promise<string> {
    if (this.cache.has(templatePath)) {
      return this.cache.get(templatePath)!;
    }

    const content = await fs.readFile(templatePath, 'utf-8');
    this.cache.set(templatePath, content);

    // 파일 변경 감지 설정
    if (!this.watchedFiles.has(templatePath)) {
      fs.watchFile(templatePath, () => {
        this.cache.delete(templatePath);
      });
      this.watchedFiles.add(templatePath);
    }

    return content;
  }
}

// src/core/async-operations.ts
export class AsyncOperations {
  async copyTemplatesParallel(templates: TemplateJob[]): Promise<void> {
    const batches = this.chunkArray(templates, 5); // 5개씩 병렬 처리

    for (const batch of batches) {
      await Promise.all(batch.map(job => this.processTemplate(job)));
    }
  }

  private chunkArray<T>(array: T[], size: number): T[][] {
    const chunks: T[][] = [];
    for (let i = 0; i < array.length; i += size) {
      chunks.push(array.slice(i, i + size));
    }
    return chunks;
  }
}
```

**Week 3 산출물:**
- ✅ Git 자동화 시스템 완성
- ✅ 16-Core TAG 시스템 포팅
- ✅ 템플릿 캐싱 및 성능 최적화
- ✅ 비동기 작업 최적화

---

### Week 4: 패키지 배포 및 테스트 (12/23 - 12/29)
**목표:** npm 패키지 배포 및 베타 테스트

#### Day 22-24: npm 패키지 준비
```json
// package.json 최종 설정
{
  "name": "@moai/adk",
  "version": "1.0.0",
  "description": "MoAI Agentic Development Kit for Claude Code",
  "main": "dist/index.js",
  "bin": {
    "moai": "dist/cli/index.js"
  },
  "files": [
    "dist/",
    "templates/",
    "README.md",
    "LICENSE"
  ],
  "engines": {
    "node": ">=18.0.0"
  },
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "test:watch": "jest --watch",
    "lint": "eslint src/**/*.ts",
    "format": "prettier --write src/**/*.ts",
    "prepublishOnly": "npm run build && npm test"
  },
  "dependencies": {
    "commander": "^11.0.0",
    "chalk": "^5.0.0",
    "inquirer": "^9.0.0",
    "fs-extra": "^11.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/node": "^20.0.0",
    "jest": "^29.0.0",
    "ts-jest": "^29.0.0"
  },
  "keywords": [
    "moai", "claude-code", "agentic", "development-kit",
    "tdd", "spec-first", "workflow", "automation"
  ],
  "author": "Modu AI",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/modu-ai/moai-adk-ts"
  }
}
```

#### Day 25-26: 베타 배포
```bash
# 1. npm 조직 생성
npm org create moai
npm org create modu

# 2. 메인 패키지 베타 배포
npm publish @moai/adk --tag beta --access public

# 3. 미러 패키지 배포
npm publish @modu/coding --tag beta --access public
npm publish @modu/dev-kit --tag beta --access public
```

#### Day 27-28: 베타 테스트
```bash
# 베타 버전 설치 테스트
npm install -g @moai/adk@beta

# 기본 기능 테스트
moai init test-project
cd test-project
ls -la .claude .moai

# 훅 실행 테스트
echo '{"file_path": "test.txt"}' | node .claude/hooks/moai/pre-write-guard.js
```

**Week 4 산출물:**
- ✅ npm 패키지 베타 배포 완료
- ✅ 설치 및 기본 기능 검증
- ✅ 사용자 피드백 수집
- ✅ 버그 수정 및 개선사항 적용

---

### Week 5: 정식 배포 및 마이그레이션 완료 (12/30 - 1/5)
**목표:** 정식 배포 및 기존 사용자 마이그레이션 지원

#### Day 29-31: 정식 배포
```bash
# 1. 최종 테스트 완료 후 정식 배포
npm publish @moai/adk --access public

# 2. 모든 미러 패키지 배포
npm publish @modu/coding --access public
npm publish @modu/dev-kit --access public
npm publish @moai/toolkit --access public
npm publish @modu-ai/adk --access public
```

#### Day 32-33: 문서 업데이트
```markdown
# 새로운 README.md
# @moai/adk - MoAI Agentic Development Kit

> The next-generation Claude Code development toolkit, now with TypeScript!

## 🚀 Quick Start

```bash
# Install (requires Node.js 18+)
npm install -g @moai/adk

# Initialize new project
moai init my-awesome-project
cd my-awesome-project

# Start your Spec-First TDD workflow
# (in Claude Code)
/moai:0-project
/moai:1-spec
/moai:2-build
/moai:3-sync
```

## 🔄 Migration from Python version

If you're upgrading from the Python version:

```bash
# Uninstall Python version
pip uninstall moai-adk

# Install TypeScript version
npm install -g @moai/adk

# Your existing projects will work seamlessly!
```
```

#### Day 34-35: 마이그레이션 지원
```typescript
// src/cli/migrate.ts - 마이그레이션 도구
export class MigrationTool {
  async migratePythonProject(projectPath: string): Promise<void> {
    console.log('🔄 Migrating Python-based MoAI project to TypeScript...');

    // 1. 백업 생성
    await this.createBackup(projectPath);

    // 2. 새 TypeScript 훅으로 교체
    await this.replaceHooks(projectPath);

    // 3. 설정 파일 업데이트
    await this.updateConfigs(projectPath);

    console.log('✅ Migration completed successfully!');
  }

  private async replaceHooks(projectPath: string): Promise<void> {
    const hooksDir = `${projectPath}/.claude/hooks/moai`;

    // Python 훅 파일들을 TypeScript 버전으로 교체
    const hookMappings = {
      'pre_write_guard.py': 'pre-write-guard.js',
      'policy_block.py': 'policy-block.js',
      'steering_guard.py': 'steering-guard.js'
    };

    for (const [oldFile, newFile] of Object.entries(hookMappings)) {
      await fs.remove(`${hooksDir}/${oldFile}`);
      await fs.copy(`./templates/hooks/${newFile}`, `${hooksDir}/${newFile}`);
    }
  }
}
```

**Week 5 산출물:**
- ✅ @moai/adk 정식 버전 배포
- ✅ 모든 미러 패키지 배포 완료
- ✅ 마이그레이션 도구 및 가이드 제공
- ✅ 기존 사용자 지원 체계 구축

---

## 📊 마이그레이션 성공 지표

### 기술적 성과
| 지표 | 목표 | 측정 방법 |
|------|------|-----------|
| 설치 시간 | 30초 이내 | `time npm install -g @moai/adk` |
| 첫 실행 시간 | 3초 이내 | `time moai init test` |
| 메모리 사용량 | 50MB 이하 | Node.js 프로세스 모니터링 |
| 패키지 크기 | 10MB 이하 | npm 패키지 분석 |

### 사용자 경험 개선
| 지표 | 현재 (Python) | 목표 (TypeScript) | 개선율 |
|------|---------------|-------------------|--------|
| 설치 단계 | 2단계 | 1단계 | 50% ⬇️ |
| 필수 의존성 | Python+pip+npm | npm only | 67% ⬇️ |
| 에러 발생률 | 15% | 5% | 67% ⬇️ |
| 타입 에러 | 런타임 | 컴파일타임 | 100% ⬆️ |

### 비즈니스 성과
- **다운로드 증가:** 월 1,000+ 다운로드 (3개월 내)
- **사용자 만족도:** 4.5/5.0 (npm 리뷰 기준)
- **커뮤니티 성장:** GitHub Stars 500+ (6개월 내)

---

## 🚨 위험 관리 계획

### 주요 위험 요소
1. **기능 누락 위험**
   - 대응: 철저한 기능 매핑 및 테스트
   - 백업: Python 버전 6개월 병행 지원

2. **성능 저하 위험**
   - 대응: 벤치마크 테스트 및 최적화
   - 백업: 성능 모니터링 및 즉시 대응

3. **사용자 혼란 위험**
   - 대응: 명확한 마이그레이션 가이드
   - 백업: 24/7 커뮤니티 지원

### 롤백 계획
```bash
# 긴급 롤백 시나리오
# 1. Python 버전 재활성화
pip install moai-adk==0.1.28

# 2. npm 패키지 deprecation
npm deprecate @moai/adk "Temporary rollback - use Python version"

# 3. 사용자 공지
echo "Temporary rollback notice sent to all users"
```

---

## 📝 결론 및 권장사항

### 핵심 성공 요인
1. **철저한 계획:** 5주 단위별 명확한 목표 설정
2. **점진적 전환:** 기능별 순차 포팅으로 위험 최소화
3. **병행 지원:** Python 버전 6개월 유지로 안전망 확보
4. **커뮤니티 중심:** 사용자 피드백 기반 개선

### 즉시 실행 권장사항
1. **Week 1 시작:** 즉시 TypeScript 프로젝트 초기화
2. **네임스페이스 확보:** @moai, @modu npm organization 생성
3. **팀 준비:** TypeScript 개발 환경 및 스킬 준비
4. **커뮤니티 공지:** 마이그레이션 계획 사전 공유

### 장기적 비전
- **2024 Q1:** TypeScript 완전 전환 완료
- **2024 Q2:** 고급 기능 및 플러그인 시스템 구축
- **2024 Q3:** 기업용 기능 및 SaaS 서비스 준비
- **2024 Q4:** 글로벌 확산 및 다국어 지원

이 마이그레이션을 통해 MoAI-ADK는 Claude Code 생태계의 표준 개발 도구로 자리잡고, 사용자 경험을 혁신적으로 개선할 것입니다.