# Documentation Generator Analysis Report

**ANALYSIS:DOC-001**: Documentation System Architecture Analysis

**생성일**: 2025-10-01
**분석 범위**: `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/`
**분석자**: Claude (Sonnet 4.5)

---

## 1. 실행 요약

### 핵심 발견사항

MoAI-ADK는 **전통적인 Documentation Generator가 없습니다**. 대신, 다음과 같은 혁신적인 접근 방식을 사용합니다:

1. **Template-Driven Documentation**: Mustache 템플릿 기반 초기 문서 생성
2. **Living Document Philosophy**: 코드와 문서의 실시간 동기화 (`/alfred:3-sync`)
3. **Agent-Orchestrated Sync**: `doc-syncer` 에이전트를 통한 문서 동기화
4. **CODE-FIRST Approach**: TAG 시스템으로 코드에서 문서로의 역방향 추적성

### 아키텍처 특징

```
┌─────────────────────────────────────────────────────────────┐
│          Documentation Generation Strategy                  │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Phase 1: Project Initialization (moai init)                │
│  ├─ TemplateProcessor: Mustache 기반 템플릿 처리           │
│  ├─ TemplateManager: 프로젝트 구조 생성                     │
│  └─ Static Templates: .moai/project/*.md 생성               │
│                                                               │
│  Phase 2: Development Cycle (/alfred:1-spec → 2-build)       │
│  ├─ spec-builder: SPEC 문서 작성 (EARS 방식)                │
│  ├─ code-builder: TDD 구현 (TAG 체인 생성)                  │
│  └─ Code Comments: @CODE:ID, @TEST:ID TAG 삽입              │
│                                                               │
│  Phase 3: Documentation Sync (/alfred:3-sync)                │
│  ├─ doc-syncer Agent: Living Document 동기화                │
│  ├─ ripgrep Scan: TAG 체인 검증 (CODE-FIRST)                │
│  ├─ PR Workflow: Draft → Ready 전환                         │
│  └─ sync-report.md: 동기화 리포트 생성                       │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 템플릿 시스템 (Template Engine)

### 2.1 TemplateProcessor (`src/core/installer/template-processor.ts`)

**역할**: Mustache 템플릿 엔진 기반 파일 생성

#### 핵심 기능

1. **Cross-Platform Template Resolution**
   ```typescript
   getTemplatesPath(): string {
     // 4가지 전략으로 템플릿 경로 탐색
     // 1. Package-relative (최우선)
     // 2. Development directory
     // 3. User's node_modules
     // 4. Global installation paths
   }
   ```

2. **Mustache Variable Interpolation**
   ```typescript
   createTemplateVariables(config: InstallationConfig) {
     return {
       PROJECT_NAME: config.projectName,
       PROJECT_DESCRIPTION: `A ${config.projectName} project...`,
       PROJECT_VERSION: '0.1.0',
       PROJECT_MODE: config.mode,
       TIMESTAMP: new Date().toISOString(),
     };
   }
   ```

3. **Recursive Template Copying**
   ```typescript
   async copyTemplateDirectory(srcDir, dstDir, variables) {
     // 재귀적 디렉토리 복사
     // 텍스트 파일: Mustache 렌더링
     // 바이너리 파일: 그대로 복사
   }
   ```

#### 지원 파일 형식

- **텍스트 파일** (Mustache 렌더링): `.md`, `.json`, `.js`, `.ts`, `.py`, `.txt`, `.yml`, `.yaml`
- **실행 권한 설정**: `.py`, `.sh`, `.js` 파일에 자동으로 `chmod 0o755` 적용 (Unix/macOS)

### 2.2 TemplateManager (`src/core/project/template-manager.ts`)

**역할**: 프로젝트 타입별 맞춤형 구조 생성

#### 언어별 프로젝트 생성

```typescript
async createProjectFiles(projectPath, config, templateData, result) {
  switch (config.type) {
    case ProjectType.PYTHON:
      await this.createPythonFiles(...);      // pyproject.toml, pytest.ini
      break;
    case ProjectType.TYPESCRIPT:
      await this.createNodeJSFiles(...);      // package.json, tsconfig.json
      break;
    case ProjectType.FRONTEND:
      await this.createFrontendFiles(...);    // React/Vue 구조
      break;
    case ProjectType.MIXED:
      await this.createMixedProjectFiles(...); // Full-stack
      break;
  }
}
```

#### MoAI 문서 구조 생성

```typescript
async createMoaiStructure(projectPath, templateData, result) {
  // .moai/config.json 생성
  const moaiConfig = this.generateMoaiConfig(templateData);

  // .moai/project/ 파일 생성
  const projectFiles = ['product.md', 'structure.md', 'tech.md'];
  for (const file of projectFiles) {
    const content = this.generateProjectFile(file, templateData);
    await fs.writeFile(filePath, content);
  }

  // .moai/reports/sync-report.md 생성
  const syncReportContent = this.generateSyncReport(templateData);
}
```

### 2.3 템플릿 파일 위치

```
moai-adk-ts/templates/
├── .moai/
│   ├── memory/
│   │   └── development-guide.md      # TRUST 원칙, @TAG 시스템
│   ├── project/
│   │   ├── product.md                # 제품 비전, EARS 가이드
│   │   ├── structure.md              # 아키텍처 설계
│   │   └── tech.md                   # 기술 스택
│   └── reports/
│       └── .gitkeep
├── .claude/
│   ├── agents/alfred/
│   │   ├── spec-builder.md
│   │   ├── code-builder.md
│   │   └── doc-syncer.md
│   └── settings.json
├── .github/
│   └── workflows/moai-gitflow.yml    # GitHub Actions
└── CLAUDE.md                          # 프로젝트 루트 CLAUDE.md
```

---

## 3. Living Document 동기화 (Doc Syncer)

### 3.1 doc-syncer 에이전트

**파일**: `/Users/goos/MoAI/MoAI-ADK/docs/claude/agents/doc-syncer.md`

#### 핵심 역할

1. **Living Document 동기화**: 코드 ↔ 문서 양방향 동기화
2. **@TAG 시스템 관리**: TAG 체인 검증 및 무결성 유지
3. **프로젝트 타입별 조건부 문서 생성**

#### 프로젝트 타입별 문서 매핑

```markdown
### 매핑 규칙

- **Web API**: API.md, endpoints.md (엔드포인트 문서화)
- **CLI Tool**: CLI_COMMANDS.md, usage.md (명령어 문서화)
- **Library**: API_REFERENCE.md, modules.md (함수/클래스 문서화)
- **Frontend**: components.md, styling.md (컴포넌트 문서화)
- **Application**: features.md, user-guide.md (기능 설명)
```

#### 동기화 대상

**코드 → 문서 동기화**:
- API 문서: 코드 변경 시 자동 갱신
- README: 기능 추가/수정 시 사용법 업데이트
- 아키텍처 문서: 구조 변경 시 다이어그램 갱신

**문서 → 코드 동기화**:
- SPEC 변경: 요구사항 수정 시 관련 코드 마킹
- TODO 추가: 문서의 할일이 코드 주석으로 반영
- TAG 업데이트: 추적성 링크 자동 갱신

### 3.2 TAG 시스템 동기화 (v5.0 - 4-Core)

#### TAG 체인 구조

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**TDD 완벽 정렬**:
- `@SPEC:ID` (사전 준비) - EARS 방식 요구사항 명세 (`.moai/specs/`)
- `@TEST:ID` (RED) - 실패하는 테스트 작성 (`tests/`)
- `@CODE:ID` (GREEN + REFACTOR) - 구현 및 리팩토링 (`src/`)
- `@DOC:ID` (문서화) - Living Document 생성 (`docs/`)

#### CODE-FIRST 원칙

```bash
# TAG 검증: 코드 직접 스캔 (중간 캐시 없음)
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 고아 TAG 탐지
rg '@CODE:AUTH-001' -n src/           # CODE는 있는데
rg '@SPEC:AUTH-001' -n .moai/specs/   # SPEC이 없으면 고아
```

### 3.3 동기화 산출물

```markdown
docs/status/sync-report.md:           # 최신 동기화 요약 리포트
docs/sections/index.md:               # Last Updated 메타 자동 반영
TAG 추적성 매트릭스:                  # 코드 스캔 결과 (실시간)
```

---

## 4. 워크플로우 자동화 (Workflow Automation)

### 4.1 WorkflowAutomation 클래스

**파일**: `src/core/git/workflow-automation.ts`

#### 3단계 워크플로우 지원

```typescript
enum SpecWorkflowStage {
  INIT = 'init',    // 프로젝트 초기화
  SPEC = 'spec',    // /alfred:1-spec - 명세 작성
  BUILD = 'build',  // /alfred:2-build - TDD 구현
  SYNC = 'sync',    // /alfred:3-sync - 문서 동기화
}
```

#### SPEC 워크플로우 시뮬레이션

```typescript
async startSpecWorkflow(specId: string, description: string) {
  // 1. SPEC 브랜치 생성
  const branchName = GitNamingRules.createSpecBranch(specId);
  await this.gitManager.createBranch(branchName, 'main');

  // 2. SPEC 디렉토리 구조 생성
  await this.createSpecStructure(specId, description);

  // 3. 초기 커밋
  const commitMessage = `docs: Initialize ${specId} specification`;
  await this.gitManager.commitChanges(commitMessage);

  // 4. Team 모드인 경우 Draft PR 생성
  if (this.config.mode === 'team') {
    pullRequestUrl = await this.createDraftPullRequest(specId, branchName, description);
  }
}
```

#### 문서 동기화 워크플로우

```typescript
async runSyncWorkflow(specId: string) {
  // 1. 문서 동기화 커밋
  const syncCommitMessage = `docs: Sync ${specId} documentation`;
  await this.gitManager.commitChanges(syncCommitMessage);

  // 2. Team 모드인 경우 PR 상태 업데이트 (Draft → Ready)
  if (this.config.mode === 'team') {
    // GitHub CLI를 사용하여 PR 상태 변경
  }
}
```

---

## 5. 다국어 지원 (i18n)

### 5.1 i18n 시스템 (`src/utils/i18n.ts`)

#### 지원 언어

```typescript
export type Locale = 'en' | 'ko';

const translations: Record<Locale, Messages> = {
  en,  // 영어
  ko,  // 한국어 (기본값)
};
```

#### 메시지 구조

```typescript
interface Messages {
  common: {
    success: string;
    error: string;
    warning: string;
    info: string;
  };
  init: {
    welcome: string;
    creating: string;
    completed: string;
    prompts: { ... };
  };
  update: { ... };
  doctor: { ... };
}
```

#### 동적 메시지 보간

```typescript
t('update.available', { from: '1.0.0', to: '2.0.0' })
// 영어: "⚡ Update available: v1.0.0 → v2.0.0"
// 한국어: "⚡ 업데이트 가능: v1.0.0 → v2.0.0"
```

#### 환경 기반 로케일 감지

```typescript
initI18n(locale?: Locale) {
  // 우선순위:
  // 1. 명시적 locale 파라미터
  // 2. process.env.MOAI_LOCALE
  // 3. process.env.LANG (ko_KR → 'ko', en_US → 'en')
  // 4. 기본값: 'ko'
}
```

---

## 6. 템플릿 확장성

### 6.1 현재 지원 언어

```typescript
enum ProjectType {
  PYTHON = 'python',
  NODEJS = 'nodejs',
  TYPESCRIPT = 'typescript',
  FRONTEND = 'frontend',
  MIXED = 'mixed',
}
```

### 6.2 새 언어 추가 방법

**단계 1**: `ProjectType` enum 확장
```typescript
enum ProjectType {
  // ... 기존 타입
  RUST = 'rust',
  GO = 'go',
}
```

**단계 2**: TemplateManager에 생성 메서드 추가
```typescript
async createRustFiles(projectPath, templateData, result) {
  // Cargo.toml 생성
  const cargoContent = this.generateCargoToml(templateData);
  await fs.writeFile(path.join(projectPath, 'Cargo.toml'), cargoContent);

  // src/main.rs 생성
  const mainRsContent = this.generateRustMain(templateData);
  await fs.writeFile(path.join(projectPath, 'src', 'main.rs'), mainRsContent);
}
```

**단계 3**: `createProjectFiles()` switch문 업데이트
```typescript
case ProjectType.RUST:
  await this.createRustFiles(projectPath, templateData, result);
  break;
```

### 6.3 템플릿 디렉토리 구조

```
moai-adk-ts/templates/
├── python/              # Python 전용 템플릿 (향후 확장)
│   ├── pyproject.toml.mustache
│   └── pytest.ini.mustache
├── typescript/          # TypeScript 전용 템플릿
│   ├── package.json.mustache
│   └── tsconfig.json.mustache
└── rust/                # Rust 템플릿 (예시)
    ├── Cargo.toml.mustache
    └── main.rs.mustache
```

---

## 7. 강점 분석

### 7.1 혁신적 접근 방식

| 전통적 Doc Generator | MoAI-ADK Living Document |
|---------------------|--------------------------|
| 코드 → 문서 (일방향) | 코드 ↔ 문서 (양방향) |
| JSDoc, TypeDoc 파싱 | @TAG 시스템 + ripgrep 스캔 |
| 빌드 시점 생성 | 실시간 동기화 (`/alfred:3-sync`) |
| 정적 HTML/Markdown | 에이전트 기반 동적 문서 |
| 수동 업데이트 | Git 워크플로우 통합 |

### 7.2 핵심 강점

#### 1. CODE-FIRST 철학
- **TAG의 진실은 코드 자체에만 존재**
- 중간 데이터베이스나 캐시 불필요
- `rg '@(SPEC|TEST|CODE):' -n` 명령어로 즉시 검증

#### 2. TDD 완벽 정렬
```
SPEC (사전 준비) → TEST (RED) → CODE (GREEN + REFACTOR) → DOC (문서화)
@SPEC:ID        → @TEST:ID    → @CODE:ID              → @DOC:ID
```

#### 3. Agent-Orchestrated Sync
- `doc-syncer` 에이전트가 프로젝트 타입을 자동 감지
- 필요한 문서만 조건부 생성 (Web API → API.md, CLI → CLI_COMMANDS.md)

#### 4. Mustache 템플릿 엔진
- 간단하고 로직이 없는 템플릿 (`{{PROJECT_NAME}}`, `{{TIMESTAMP}}`)
- 텍스트 파일 자동 감지 및 렌더링

#### 5. Cross-Platform Template Resolution
- 4가지 전략으로 템플릿 경로 탐색
- 개발 환경, npm/bun 글로벌 설치, 로컬 node_modules 모두 지원

---

## 8. 개선 권장사항

### 8.1 단기 개선 (1개월)

#### 1. API 문서 자동 생성 강화
**현재 상태**: doc-syncer 에이전트가 수동으로 문서 작성
**제안**: TypeScript AST 파싱을 통한 자동 API 문서 생성

```typescript
// 추가할 모듈: src/docs/api-generator.ts
class APIGenerator {
  async generateFromSource(sourceFiles: string[]): Promise<string> {
    const ast = parseTypeScript(sourceFiles);
    const exports = extractExports(ast);
    return this.renderMarkdown(exports);
  }
}
```

**구현 예시**:
```typescript
import * as ts from 'typescript';

export class TypeScriptAPIGenerator {
  extractExports(sourceFile: ts.SourceFile) {
    const exports: ExportInfo[] = [];
    ts.forEachChild(sourceFile, node => {
      if (ts.isExportDeclaration(node)) {
        // 내보내기 선언 파싱
      } else if (ts.isFunctionDeclaration(node) && hasExportModifier(node)) {
        // 함수 시그니처 추출
      }
    });
    return exports;
  }
}
```

#### 2. sync-report.md 템플릿 개선
**현재 상태**: 최소한의 정보만 포함
**제안**: TAG 체인 통계, 커버리지 분석 추가

```markdown
# Sync Report - {{PROJECT_NAME}}

Generated: {{TIMESTAMP}}

## TAG Statistics
- Total @SPEC: 42
- Total @TEST: 38 (90% coverage)
- Total @CODE: 40 (95% coverage)
- Total @DOC: 15 (36% coverage)

## Orphan TAGs
- ⚠️  @CODE:AUTH-007 (no matching @SPEC)
- ⚠️  @TEST:LOGIN-003 (no matching @CODE)

## Missing Documentation
- [ ] @DOC:AUTH-001 (API 문서 필요)
- [ ] @DOC:LOGIN-002 (사용자 가이드 필요)
```

### 8.2 중기 개선 (3개월)

#### 1. 다국어 문서 지원 확장
**현재 상태**: CLI만 한국어/영어 지원
**제안**: 템플릿 문서도 다국어화

```
templates/
├── .moai/
│   └── project/
│       ├── product.en.md           # 영어 템플릿
│       ├── product.ko.md           # 한국어 템플릿
│       ├── structure.en.md
│       └── structure.ko.md
```

```typescript
generateProjectFile(filename: string, templateData: TemplateData): string {
  const locale = getLocale();
  const localizedFile = filename.replace('.md', `.${locale}.md`);
  // 로케일별 템플릿 로드
}
```

#### 2. Mermaid 다이어그램 자동 생성
**제안**: 코드 구조에서 아키텍처 다이어그램 자동 생성

```typescript
class DiagramGenerator {
  generateArchitectureDiagram(projectPath: string): string {
    const structure = analyzeProjectStructure(projectPath);
    return this.generateMermaid(structure);
  }

  generateMermaid(structure: ProjectStructure): string {
    return `
      graph TD
        A[Frontend] --> B[API Layer]
        B --> C[Business Logic]
        C --> D[Data Layer]
    `;
  }
}
```

#### 3. 릴리스 노트 자동 생성
**제안**: Git 커밋 히스토리에서 릴리스 노트 추출

```typescript
class ReleaseNotesGenerator {
  async generateFromCommits(fromTag: string, toTag: string): Promise<string> {
    const commits = await this.gitManager.getCommitRange(fromTag, toTag);
    const categorized = this.categorizeCommits(commits); // feat, fix, docs, etc.
    return this.formatMarkdown(categorized);
  }
}
```

### 8.3 장기 개선 (6개월+)

#### 1. 외부 문서 도구 통합
- **VitePress**: 정적 사이트 생성 (현재 `docs/.vitepress/config.ts` 존재)
- **Docusaurus**: React 기반 문서 사이트
- **MkDocs**: Python 프로젝트용 문서

```typescript
class DocusaurusIntegration {
  async generateDocusaurusSite(projectPath: string) {
    // sidebar.js 자동 생성
    // docusaurus.config.js 생성
    // 기존 .md 파일을 Docusaurus 포맷으로 변환
  }
}
```

#### 2. AI 기반 문서 개선
- **Claude API 통합**: 코드 주석에서 자연어 설명 생성
- **문서 품질 분석**: 가독성, 완성도 점수 제공

```typescript
class AIDocumentEnhancer {
  async enhanceDocumentation(markdownContent: string): Promise<string> {
    const enhanced = await callClaudeAPI({
      prompt: `다음 API 문서를 더 명확하고 상세하게 개선해주세요:\n${markdownContent}`,
    });
    return enhanced;
  }
}
```

#### 3. 대화형 문서 생성
- **Swagger/OpenAPI**: REST API 자동 문서화
- **GraphQL Playground**: GraphQL 스키마 문서화

---

## 9. 보안 & 검증

### 9.1 템플릿 보안 (`src/core/installer/templates/template-security.ts`)

**예상 보안 조치** (파일 미확인, 향후 구현 가능):
- Mustache 템플릿 검증
- 악성 변수 주입 방지
- 파일 경로 탐색 공격 방어

### 9.2 TAG 검증

```bash
# 무결성 검사 (/alfred:3-sync 실행 시)
rg '@(SPEC|TEST|CODE|DOC):' -n

# 중복 ID 탐지
rg '@CODE:AUTH-001' -n src/ | wc -l  # 1개여야 함

# 고아 TAG 탐지
comm -23 <(rg '@CODE:' -no-filename -r '$1' | sort -u) \
         <(rg '@SPEC:' -no-filename -r '$1' .moai/specs/ | sort -u)
```

---

## 10. 성능 고려사항

### 10.1 템플릿 렌더링 성능

- **Mustache**: 경량 템플릿 엔진 (~30KB)
- **파일 I/O**: `fs.promises` 기반 비동기 처리
- **재귀 복사**: 대형 프로젝트에서도 효율적

### 10.2 TAG 스캔 성능

- **ripgrep (rg)**: Rust 기반, grep보다 10~100배 빠름
- **병렬 처리**: 멀티코어 활용
- **인덱스 불필요**: CODE-FIRST 원칙으로 실시간 스캔

**예상 성능** (1000 파일 기준):
- `rg '@(SPEC|TEST|CODE):' -n`: ~50ms
- 전통적 DB 기반 TAG 시스템: ~200ms (쿼리 + 인덱스 동기화)

---

## 11. 결론

### 11.1 핵심 통찰

MoAI-ADK의 Documentation 시스템은 전통적인 "Documentation Generator"가 아닌, **Living Document Philosophy**를 구현한 혁신적인 접근 방식입니다:

1. **Template-Driven Initialization**: Mustache 기반 초기 구조 생성
2. **Agent-Orchestrated Synchronization**: `doc-syncer` 에이전트를 통한 지능형 동기화
3. **CODE-FIRST Traceability**: TAG 시스템으로 코드와 문서의 양방향 추적성
4. **TDD-Aligned Workflow**: SPEC → TEST → CODE → DOC 단계별 문서화

### 11.2 경쟁 우위

| 비교 항목 | 전통적 Doc Gen | MoAI-ADK |
|-----------|----------------|----------|
| 동기화 방향 | 코드 → 문서 | 코드 ↔ 문서 |
| 업데이트 시점 | 빌드 시점 | 실시간 (`/alfred:3-sync`) |
| 추적성 | JSDoc 태그 | @TAG 시스템 (4-Core) |
| 다국어 지원 | 수동 번역 | i18n 시스템 (en/ko) |
| 확장성 | 플러그인 | 에이전트 + 템플릿 |

### 11.3 적용 시나리오

**최적 사용 사례**:
- SPEC-First TDD 개발 프로세스
- 멀티 언어 프로젝트 (Python, TypeScript, Go, Rust 등)
- GitHub 기반 협업 워크플로우
- Living Documentation이 필요한 장기 프로젝트

**부적합 사례**:
- 레거시 코드 역공학 (기존 코드에 TAG 주석 없음)
- 일회성 프로토타입 (문서화 오버헤드)
- GitHub 없는 환경 (Team 모드 불가)

---

## 12. 참고 자료

### 분석 대상 파일

1. **Template System**
   - `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/core/installer/template-processor.ts` (353줄)
   - `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/core/project/template-manager.ts` (610줄)

2. **Documentation Agents**
   - `/Users/goos/MoAI/MoAI-ADK/docs/claude/agents/doc-syncer.md` (105줄)

3. **Workflow Automation**
   - `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/core/git/workflow-automation.ts` (385줄)

4. **Internationalization**
   - `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/utils/i18n.ts` (322줄)

5. **Template Files**
   - `/Users/goos/MoAI/MoAI-ADK/moai-adk-ts/templates/` (디렉토리)

### 외부 의존성

- **mustache**: 템플릿 렌더링 엔진
- **ripgrep (rg)**: TAG 검색 도구
- **fs-extra**: 파일 시스템 유틸리티

### 관련 문서

- CLAUDE.md: @TAG Lifecycle 5.0 (4-Core) 정의
- development-guide.md: TRUST 원칙, SPEC-TDD 워크플로우
- product.md, structure.md, tech.md: 프로젝트 컨텍스트 템플릿

---

**보고서 생성 시간**: 2025-10-01
**다음 영역**: 9. Core Utilities Analysis
**@TAG**: ANALYSIS:DOC-001
