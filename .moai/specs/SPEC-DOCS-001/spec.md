---
id: DOCS-001
version: 0.0.1
status: draft
created: 2025-10-11
updated: 2025-10-11
author: @Goos
priority: high
category: docs
labels: ["documentation", "github-pages", "restructure", "vitepress-removal"]
---

# @SPEC:DOCS-001: 문서 구조 재편 (VitePress 제거 + GitHub Pages 최적화)

## HISTORY

### v0.0.1 (2025-10-11)
- **INITIAL**: 문서 구조 재편 SPEC 작성
- **AUTHOR**: @Goos
- **REASON**: VitePress 제거, 순수 마크다운 기반 GitHub Pages 전환, 프로젝트 초기 상태 문서 정리

---

## 1. Overview (개요)

### 목적
- VitePress 기반 빌드 시스템을 제거하고 GitHub Pages 네이티브 마크다운으로 전환
- 프로젝트 초기 상태이므로 MoAI-ADK 핵심 콘텐츠 중심으로 문서 구조 간소화
- README.md를 간단한 소개 + docs 링크로 축소
- 9개 카테고리 기반 체계적인 문서 구조 구축

### 배경
- 현재 프로젝트는 초기 단계로, VitePress 유지보수 오버헤드가 불필요
- GitHub Pages 네이티브 지원으로 충분한 문서화 가능
- CLAUDE.md, development-guide.md 등 핵심 문서를 docs 구조로 재구성 필요

### 범위
- **포함**: VitePress 제거, docs 구조 재편, README.md 간소화, GitHub Pages 설정
- **제외**: 콘텐츠 신규 작성 (기존 콘텐츠 재구성만)

---

## 2. EARS Requirements (요구사항)

### 2.1 Ubiquitous Requirements (기본 요구사항)

1. **VitePress 제거**
   - 시스템은 VitePress 관련 모든 파일과 의존성을 제거해야 한다
   - `.vitepress/`, `package.json` vitepress 의존성, `.github/workflows/deploy-docs.yml` 제거

2. **순수 마크다운 문서 구조**
   - 시스템은 `docs/` 디렉토리에 9개 카테고리 기반 순수 마크다운 문서를 제공해야 한다
   - 각 카테고리는 `index.md` (목차) + 개별 문서로 구성

3. **README.md 간소화**
   - 시스템은 프로젝트 한 줄 소개, 핵심 링크(CLAUDE.md, docs), 빠른 시작만 포함하는 간소화된 README를 제공해야 한다

4. **GitHub Pages 호환 구조**
   - 시스템은 GitHub Pages가 네이티브로 렌더링할 수 있는 마크다운 구조를 제공해야 한다
   - `docs/index.md`를 메인 문서 허브로 사용

### 2.2 Event-driven Requirements (이벤트 기반)

1. **VitePress 빌드 실패 시**
   - WHEN VitePress 관련 파일이 없을 때, 시스템은 빌드 에러를 발생시키지 않아야 한다

2. **사용자 문서 접근 시**
   - WHEN 사용자가 `docs/`에 접근하면, 시스템은 카테고리별 인덱스를 통해 원하는 문서를 빠르게 찾을 수 있어야 한다

3. **GitHub Pages 배포 시**
   - WHEN GitHub Pages가 `docs/` 디렉토리를 배포하면, 시스템은 모든 마크다운 파일을 자동으로 렌더링해야 한다

4. **README.md 열람 시**
   - WHEN 사용자가 README.md를 열면, 시스템은 3분 이내에 프로젝트 핵심을 이해할 수 있는 정보를 제공해야 한다

### 2.3 State-driven Requirements (상태 기반)

1. **프로젝트 초기 상태**
   - WHILE 프로젝트가 초기 단계일 때, 시스템은 핵심 문서(CLAUDE.md, development-guide.md, Alfred 커맨드)만 docs로 전환해야 한다

2. **문서 마이그레이션 중**
   - WHILE 문서 재구성이 진행 중일 때, 시스템은 기존 콘텐츠 참조 링크를 유지해야 한다

3. **GitHub Pages 활성화 후**
   - WHILE GitHub Pages가 활성화된 상태일 때, 시스템은 모든 docs 링크가 웹에서 정상 작동하도록 보장해야 한다

### 2.4 Optional Features (선택적 기능)

1. **자동 TOC 생성**
   - WHERE 카테고리별 문서가 5개 이상이면, 시스템은 자동 목차 생성 스크립트를 제공할 수 있다

2. **검색 기능**
   - WHERE 사용자 요청이 있으면, 시스템은 GitHub Pages 검색 플러그인을 추가할 수 있다

### 2.5 Constraints (제약사항)

1. **콘텐츠 무손실**
   - IF 기존 문서를 재구성하면, 시스템은 핵심 콘텐츠를 누락 없이 이전해야 한다

2. **링크 무결성**
   - IF 문서 경로가 변경되면, 시스템은 모든 내부 링크를 업데이트해야 한다

3. **파일 크기 제한**
   - 단일 마크다운 파일은 1000 LOC를 초과하지 않아야 한다

4. **카테고리 일관성**
   - 모든 docs 하위 문서는 9개 카테고리 중 하나에만 속해야 한다

---

## 3. Technical Specifications (기술 명세)

### 3.1 제거 대상

**파일/디렉토리**:
- `.vitepress/` (전체 디렉토리)
- `.github/workflows/deploy-docs.yml`
- `package.json` 내 vitepress 관련 의존성 및 스크립트

**의존성**:
```json
"vitepress": "^x.x.x",
"vue": "^x.x.x"
```

**npm 스크립트**:
```json
"docs:dev": "vitepress dev",
"docs:build": "vitepress build",
"docs:preview": "vitepress preview"
```

### 3.2 새로운 docs 구조

```
docs/
├── index.md                      # 메인 허브 (9개 카테고리 링크)
├── getting-started/
│   ├── index.md                  # 시작하기 목차
│   ├── installation.md           # 설치 가이드
│   └── quick-start.md            # 빠른 시작
├── concepts/
│   ├── index.md                  # 개념 목차
│   ├── spec-first-tdd.md         # SPEC-First TDD
│   ├── trust-principles.md       # TRUST 5원칙
│   ├── tag-system.md             # @TAG 시스템
│   └── ears-methodology.md       # EARS 방법론
├── alfred/
│   ├── index.md                  # Alfred 목차
│   ├── overview.md               # Alfred 개요 (CLAUDE.md 일부)
│   ├── commands.md               # 커맨드 가이드 (1-spec, 2-build, 3-sync)
│   └── orchestration.md          # 오케스트레이션 전략
├── cli/
│   ├── index.md                  # CLI 목차
│   ├── commands.md               # CLI 명령어 (init, doctor, status 등)
│   └── configuration.md          # 설정 파일 (.moai/config.json)
├── api/
│   ├── index.md                  # API 목차
│   └── agents/
│       ├── spec-builder.md       # spec-builder 에이전트
│       ├── code-builder.md       # code-builder 에이전트
│       ├── doc-syncer.md         # doc-syncer 에이전트
│       ├── tag-agent.md          # tag-agent
│       ├── git-manager.md        # git-manager
│       ├── debug-helper.md       # debug-helper
│       ├── trust-checker.md      # trust-checker
│       └── project-manager.md    # project-manager
├── guides/
│   ├── index.md                  # 가이드 목차
│   ├── development-workflow.md   # 개발 워크플로우 (development-guide.md 요약)
│   ├── spec-writing.md           # SPEC 작성 가이드
│   ├── tdd-implementation.md     # TDD 구현 가이드
│   └── context-engineering.md    # 컨텍스트 엔지니어링
├── agents/
│   ├── index.md                  # 에이전트 목차
│   ├── architecture.md           # 에이전트 아키텍처
│   ├── creating-agents.md        # 커스텀 에이전트 작성
│   └── agent-collaboration.md    # 에이전트 협업 원칙
├── examples/
│   ├── index.md                  # 예제 목차
│   ├── auth-system.md            # 인증 시스템 예제
│   └── api-endpoint.md           # API 엔드포인트 예제
└── contributing/
    ├── index.md                  # 기여 가이드 목차
    ├── code-of-conduct.md        # 행동 강령
    └── pull-request-guide.md     # PR 가이드
```

### 3.3 README.md 새 구조

```markdown
# MoAI-ADK

**SPEC-First TDD Development Kit with Alfred SuperAgent**

> "명세 없이는 코드 없음. 테스트 없이는 구현 없음."

## 빠른 시작

\`\`\`bash
npx moai-adk init my-project
cd my-project
/alfred:8-project  # 프로젝트 초기화
/alfred:1-spec     # SPEC 작성
/alfred:2-build    # TDD 구현
/alfred:3-sync     # 문서 동기화
\`\`\`

## 문서

- **[📖 전체 문서](docs/index.md)** - 카테고리별 가이드
- **[🚀 시작하기](docs/getting-started/index.md)** - 설치 및 빠른 시작
- **[▶◀ Alfred 가이드](docs/alfred/index.md)** - SuperAgent 커맨드
- **[🏗️ CLAUDE.md](CLAUDE.md)** - 프로젝트 지침 (AI 에이전트용)
- **[📋 Development Guide](.moai/memory/development-guide.md)** - 개발 가드레일

## 핵심 기능

- ✅ SPEC-First TDD 워크플로우
- 🤖 Alfred SuperAgent (9개 전문 에이전트)
- 🏷️ @TAG 추적성 시스템
- 📊 TRUST 5원칙 자동 검증
- 🌐 다중 언어 지원 (Python, TypeScript, Java, Go, Rust 등)

## 라이선스

MIT © MoAI
```

### 3.4 GitHub Pages 설정

**Repository Settings**:
- Source: `Deploy from a branch`
- Branch: `main` (또는 `develop`)
- Folder: `/docs`

**docs/_config.yml** (Jekyll 기본 설정):
```yaml
theme: jekyll-theme-minimal
title: MoAI-ADK Documentation
description: SPEC-First TDD Development Kit
```

---

## 4. Migration Strategy (마이그레이션 전략)

### Phase 1: VitePress 제거
1. `.vitepress/` 디렉토리 삭제
2. `package.json`에서 vitepress 의존성 제거
3. `.github/workflows/deploy-docs.yml` 삭제

### Phase 2: docs 구조 생성
1. 9개 카테고리 디렉토리 생성
2. 각 카테고리에 `index.md` 생성 (목차 역할)

### Phase 3: 콘텐츠 마이그레이션
1. **CLAUDE.md** → `docs/alfred/overview.md`, `docs/alfred/commands.md`
2. **development-guide.md** → `docs/guides/development-workflow.md`, `docs/concepts/`
3. 기존 docs 콘텐츠 → 새 카테고리별 재배치

### Phase 4: README.md 간소화
1. 기존 README.md 백업
2. 새 템플릿으로 교체
3. 핵심 링크 검증

---

## 5. Acceptance Criteria (인수 기준)

1. **VitePress 제거 완료**
   - `.vitepress/` 디렉토리 미존재
   - `package.json`에 vitepress 의존성 없음
   - `npm install` 시 vitepress 관련 에러 없음

2. **docs 구조 정상 작동**
   - 9개 카테고리 디렉토리 존재
   - 각 카테고리 `index.md` 존재
   - `docs/index.md` 메인 허브 정상 작동

3. **GitHub Pages 렌더링 성공**
   - `https://<username>.github.io/<repo>/docs/` 접근 가능
   - 모든 마크다운 파일 정상 렌더링
   - 내부 링크 정상 작동

---

## 6. Traceability (추적성)

**Related Artifacts**:
- `@TEST:DOCS-001` → `tests/docs/structure.test.ts` (디렉토리 구조 검증)
- `@CODE:DOCS-001` → `scripts/migrate-docs.ts` (마이그레이션 스크립트)
- `@DOC:DOCS-001` → `docs/index.md` (메인 허브)

**Dependencies**:
- `.moai/project/product.md` (프로젝트 목적 참조)
- `CLAUDE.md` (기존 문서 구조)
- `development-guide.md` (개발 원칙)
