---
id: SPEC-NEXTRA-001
version: "1.1.0"
status: "draft"
created: "2025-11-28"
updated: "2025-11-28"
author: "GOOS"
priority: "HIGH"
---

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1.0 | 2025-11-28 | GOOS | Phase 4 확장: 전체 콘텐츠 마이그레이션 계획 추가 (README 1,773줄, 22개 Skills, Advanced 섹션) |
| 1.0.0 | 2025-11-28 | GOOS | Nextra 기반 온라인 문서 사이트 초기 SPEC - 6단계 구현 계획 |

# SPEC-NEXTRA-001: MoAI-ADK 온라인 문서 사이트 구축 (Nextra 기반)

## 개요

MoAI-ADK의 모든 기능과 명령어를 문서화하는 Nextra 기반 온라인 문서 사이트를 구축합니다. README.ko.md와 기존 .moai/docs 내용을 기반으로 체계적인 문서 구조를 제공하고, Git Worktree CLI 문서를 포함한 모든 스킬과 명령어에 대한 상세 가이드를 제공합니다.

**핵심 가치**:
- 🌐 **통합 문서**: 모든 MoAI-ADK 기능을 한 곳에서 검색 가능
- 🎨 **일관된 디자인**: MoAI-ADK 브랜딩과 그레이스케일 테마 적용
- 📱 **반응형**: 모바일 최적화된 문서 경험
- 🚀 **고성능**: Next.js 15+ 기반으로 빠른 로딩 속도
- 🔍 **검색 최적화**: 효율적인 내비게이션과 검색 기능

---

## Environment (실행 환경 및 전제 조건)

### 시스템 환경

- **Node.js**: 18.17+ LTS
- **Next.js**: 15.0+ 최신 버전
- **Nextra**: 3.0+ 최신 버전
- **Package Manager**: npm, yarn, 또는 pnpm
- **배포 환경**: Vercel (권장) 또는 Netlify

### 필수 패키지

```json
{
  "dependencies": {
    "next": "^15.0.0",
    "nextra": "^3.0.0",
    "nextra-theme-docs": "^3.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^19.0.0",
    "typescript": "^5.0.0",
    "eslint": "^8.0.0",
    "prettier": "^3.0.0"
  }
}
```

### 프로젝트 구조

```
docs/
├── pages/                    # Nextra 페이지 구조
│   ├── index.mdx            # 홈페이지
│   ├── getting-started/     # PART A: 시작하기
│   │   ├── _meta.js         # 네비게이션 메타데이터
│   │   ├── overview.mdx     # 개요
│   │   ├── installation.mdx # 설치 가이드
│   │   └── quickstart.mdx   # 빠른 시작
│   ├── core-concepts/       # PART B: 핵심 개념
│   │   ├── _meta.js
│   │   ├── spec-format.mdx  # SPEC과 EARS 포맷
│   │   ├── agents.mdx       # Mr.Alfred와 에이전트
│   │   ├── workflow.mdx     # 개발 워크플로우
│   │   └── commands.mdx     # 핵심 커맨드
│   ├── advanced/            # PART C: 심화 학습
│   │   ├── _meta.js
│   │   ├── agents-guide.mdx # 에이전트 가이드 (26개)
│   │   ├── skills-library.mdx # 스킬 라이브러리 (22개)
│   │   ├── patterns.mdx     # 조합 패턴과 예제
│   │   └── trust5-quality.mdx # TRUST 5 품질보증
│   ├── worktree/            # Git Worktree CLI 문서
│   │   ├── _meta.js
│   │   ├── guide.mdx        # Worktree 가이드
│   │   ├── examples.mdx     # 사용 예제
│   │   └── faq.mdx          # 자주 묻는 질문
│   └── reference/           # API 레퍼런스
│       ├── _meta.js
│       ├── skills.mdx       # 스킬 API 레퍼런스
│       └── commands.mdx     # 명령어 레퍼런스
├── styles/                  # 스타일 파일
│   └── globals.css          # 전역 스타일 (그레이스케일 테마)
├── components/              # React 컴포넌트
├── public/                  # 정적 에셋
├── theme.config.tsx         # Nextra 테마 설정
├── next.config.js           # Next.js 설정
└── package.json             # 패키지 설정
```

---

## Assumptions (가정 사항)

### 기술적 가정

1. **기존 콘텐츠**: README.ko.md (1,865줄)와 .moai/docs/ 내용이 마이그레이션 대상임
2. **스킬 통합**: 22개의 전문 스킬 문서가 자동으로 레퍼런스 페이지에 포함됨
3. **명령어 문서**: 6개의 코어 커맨드(/moai:0-3)가 상세하게 문서화됨
4. **워크트리 통합**: SPEC-WORKTREE-001의 모든 결과물이 통합됨

### 사용자 가정

1. **주요 사용자**: 한국어 사용자, MoAI-ADK 개발자
2. **기술 수준**: 중급 이상 개발자 (CLI, Git, 기본 웹 기술 이해)
3. **접근성**: 모바일 기기와 데스크톱 모두에서 최적화된 경험 필요
4. **검색**: 키워드 기반 검색이 효율적으로 작동해야 함

---

## Requirements (요구사항)

### Requirements (Ubiquitous) - 보편적 요구사항

- 사이트는 Next.js 15+와 Nextra 3.0+로 구현되어야 한다.
- 모든 페이지는 MDX 형식으로 작성되어야 한다.
- 사이트는 라이트 모드와 다크 모드를 모두 지원해야 한다.
- 모든 콘텐츠는 한국어로 제공되어야 한다.
- MoAI-ADK 브랜딩과 일관된 그레이스케일 테마를 적용해야 한다.

### Requirements (Event-driven) - 이벤트 기반 요구사항

- WHEN 사용자가 홈페이지에 방문했을 때 THEN 핵심 기능과 빠른 시작 가이드가 표시되어야 한다.
- WHEN 사용자가 검색어를 입력했을 때 THEN 관련 페이지와 섹션이 표시되어야 한다.
- WHEN 사용자가 워크트리 섹션을 탐색할 때 THEN 모든 워크트리 관련 문서로 이동할 수 있어야 한다.
- WHEN 사용자가 스킬 레퍼런스를 조회할 때 THEN 22개 스킬의 상세 정보를 볼 수 있어야 한다.

### Requirements (State-driven) - 상태 기반 요구사항

- 사이트는 Lighthouse 90+ 점수를 달성해야 한다.
- 첫 페이지 로딩 시간은 2.5초 이내여야 한다.
- 모든 이미지는 최적화되어 Core Web Vitals를 만족해야 한다.
- 테마 전환(라이트/다크)은 즉시 적용되어야 한다.
- 검색 기능은 300ms 이내에 결과를 표시해야 한다.

### Requirements (Interface) - 인터페이스 요구사항

- 모든 내부 링크는 자동으로 생성되어야 한다 (Table of Contents).
- 이전/다음 페이지 네비게이션이 제공되어야 한다.
- 헤더와 푸터는 접근성 가이드라인(WCAG 2.1+)을 준수해야 한다.
- 코드 블록은 구문 강조와 복사 기능을 제공해야 한다.
- 모바일에서는 햄버거 메뉴가 제공되어야 한다.

---

## Specifications (상세 명세)

### 1. CSS 스타일 시스템

#### 1.1 그레이스케일 테마 변수

```css
:root {
  /* 웹폰트 설정 */
  --font-sans: 'Pretendard', 'Inter', system-ui, sans-serif;
  --font-en: 'Inter', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
  --font-ko: 'Pretendard', 'Noto Sans KR', 'Apple SD Gothic Neo', sans-serif;
  --font-code: 'JetBrains Mono', 'Hack', 'Consolas', 'Monaco', monospace;

  /* 라이트 테마 */
  --color-primary-fg: #000000;
  --color-primary-bg: #FFFFFF;
  --color-accent-fg: #666666;
  --color-bg: #FFFFFF;
  --color-bg-light: #F9F9F9;
  --color-surface: #F5F5F5;
  --color-text: #000000;
  --color-text-secondary: #666666;
  --color-border: #DDDDDD;
  --color-code-bg: #F0F0F0;
  --color-brand-primary: #1976d2;
  --color-brand-accent: #dc2626;
  --color-brand-success: #059669;
}

/* 다크 테마 오버라이드 */
[data-theme="dark"] {
  --color-primary-fg: #FFFFFF;
  --color-primary-bg: #121212;
  --color-accent-fg: #BBBBBB;
  --color-bg: #121212;
  --color-bg-light: #1E1E1E;
  --color-surface: #1E1E1E;
  --color-text: #FFFFFF;
  --color-text-secondary: #BBBBBB;
  --color-border: #333333;
  --color-code-bg: #1E1E1E;
  --color-brand-primary: #4dabf7;
}
```

#### 1.2 한글 타이포그래피

```css
/* 한글 최적화 */
.ko-typography {
  font-family: var(--font-ko);
  letter-spacing: -0.02em;
  line-height: 1.7;
  word-break: keep-word;
}

/* 코드 블록 스타일링 */
.code-block {
  font-family: var(--font-code);
  background: var(--color-code-bg);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 1rem;
  overflow-x: auto;
}
```

### 2. 콘텐츠 마이그레이션 명세

#### 2.1 README.ko.md 구조화

| README 섹션 | Nextra 페이지 | 우선순위 |
|-------------|--------------|----------|
| PART A (시작하기) | getting-started/* | HIGH |
| PART B (핵심 개념) | core-concepts/* | HIGH |
| PART C (심화 학습) | advanced/* | MEDIUM |
| PART D (고급 & 참고) | reference/* | MEDIUM |
| 워크트리 관련 | worktree/* | HIGH |

#### 2.2 .moai/docs/ 통합

- WORKTREE_GUIDE.md → worktree/guide.mdx
- WORKTREE_FAQ.md → worktree/faq.mdx
- WORKTREE_EXAMPLES.md → worktree/examples.mdx

### 3. Nextra 설정 명세

#### 3.1 theme.config.tsx

```typescript
export default {
  logo: <span>🗿 MoAI-ADK</span>,
  project: {
    link: 'https://github.com/moai-ai/moai-adk',
  },
  docsRepositoryBase: 'https://github.com/moai-ai/moai-adk/tree/main/docs',
  footer: {
    text: 'MIT License © 2025 MoAI-ADK Contributors',
  },
  toc: {
    backToTop: true,
    extraContent: (
      <div style={{ marginTop: '1rem' }}>
        <a href="/worktree">Git Worktree CLI →</a>
      </div>
    ),
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
  },
}
```

### 4. 성능 최적화 명세

#### 4.1 이미지 최적화

- 모든 이미지는 WebP 형식으로 제공
- 반응형 이미지 (srcset) 적용
- 지연 로딩 (lazy loading) 적용
- Core Web Vitals 준수

#### 4.2 번들 최적화

- 코드 스플리팅 (페이지별 분리)
- 트리 쉐이킹 (사용하지 않는 코드 제거)
- gzip 압축 활성화
- CDN을 통한 정적 에셋 전송

### 5. 검색 기능 명세

- 전체 텍스트 검색 지원
- 페이지 섹션별 검색
- 실시간 검색 결과 표시
- 검색 기록 저장
- 키보드 단축키 (Cmd/Ctrl + K)

### 6. 접근성 명세

- WCAG 2.1 AA 준수
- 키보드 내비게이션 지원
- 스크린 리더 최적화
- 고대비 모드 지원
- 포커스 관리

---

## Constraints (제약 조건)

### 기술적 제약

- Next.js 15+와 Nextra 3.0+를 반드시 사용해야 한다.
- 모든 색상은 CSS 변수를 통해서만 정의해야 한다 (하드코딩 금지).
- 웹폰트는 Pretendard, Inter, JetBrains Mono만 사용해야 한다.
- 배포는 Vercel로 제한한다.
- 빌드 타임은 3분 이내여야 한다.

### 콘텐츠 제약

- 모든 문서는 한국어로 작성되어야 한다.
- 코드 예제는 실제로 동작하는 코드여야 한다.
- 외부 링크는 모두 https를 사용해야 한다.
- 이미지는 alt 속성을 반드시 포함해야 한다.

### 디자인 제약

- 그레이스케일 컬러 팔레트만 사용해야 한다.
- 브랜드 색상은 보조색으로만 제한적으로 사용해야 한다.
- 일관된 간격과 타이포그래피를 유지해야 한다.
- 다크 모드는 라이트 모드의 반전이 아닌 최적화된 색상을 사용해야 한다.

### 시간 제약

- MVP 구현: 2주 이내
- 전체 기능 구현: 4주 이내
- 성능 최적화: 1주 이내
- 배포 및 안정화: 1주 이내

---

## Traceability (추적성)

### 관련 SPEC

- **SPEC-WORKTREE-001**: 워크트리 CLI 문서 통합 (HIGH)
- **SPEC-UPDATE-001**: 기존 문서 업데이트 패턴 참고 (MEDIUM)

### 의존성

- **manager-docs**: 문서 생성 및 관리 에이전트
- **moai-library-nextra**: Nextra 통합 스킬
- **docs/styles/globals.css**: 기존 스타일 시스템 참고

### 성공 기준

- **문서 완성도**: 100% (모든 기능 문서화)
- **성능 점수**: Lighthouse 90+
- **접근성**: WCAG 2.1 AA 준수
- **사용자 만족도**: 4.5/5.0 이상 (내부 테스트)

### 리스크 관리

- **HIGH**: 대용량 콘텐츠 마이그레이션 → 자동화 스크립트 개발
- **MEDIUM**: 성능 최적화 → 사전 프로파일링 및 테스트
- **LOW**: 브라우저 호환성 → 모던 브라우저만 지원으로 범위 축소

---

## Phase 4 확장: 전체 콘텐츠 마이그레이션 상세 계획 (업데이트)

> **Note**: 본 섹션은 Phase 4의 구체적인 확장 계획입니다. 기존 6단계 구조에서 Phase 4를 세분화하여 완전한 콘텐츠 마이그레이션을 달성합니다.

### 4.1 README.ko.md 완전 분석 및 구조화

**목표**: 1,773줄 전체 내용 분석 및 Nextra 구조에 맞게 재구성

**작업 내용**:
- README.ko.md 전체 파싱 및 섹션 분류
- PART A-D 섹션별 매핑 및 재구성
- MDX 변환 규칙 적용 및 검증

**섹션별 마이그레이션 계획**:

| README 원본 섹션 | 대상 Nextra 페이지 | 우선순위 | 예상 페이지 수 |
|------------------|-------------------|----------|---------------|
| PART A: 시작하기 | getting-started/* | HIGH | 5-7 페이지 |
| PART B: 핵심 개념 | core-concepts/* | HIGH | 8-10 페이지 |
| PART C: 워크플로우 | workflows/* | MEDIUM | 6-8 페이지 |
| PART D: 고급 기능 | advanced/* | MEDIUM | 10-12 페이지 |

### 4.2 핵심 콘텐츠 마이그레이션

**PART A: 시작하기 → Getting Started**
- overview.mdx: MoAI-ADK 개요 및 핵심 가치
- installation.mdx: 설치 가이드 (pip, uv, 환경 설정)
- quickstart.mdx: 5분 빠른 시작 튜토리얼
- first-spec.mdx: 첫 번째 SPEC 생성 예제
- configuration.mdx: config.json 설정 가이드

**PART B: 핵심 개념 → Core Concepts**
- spec-format.mdx: SPEC과 EARS 포맷 상세 설명
- agents.mdx: Mr.Alfred와 26개 전문 에이전트 개요
- workflow.mdx: SPEC-First TDD 워크플로우
- commands.mdx: /moai:0-3 핵심 커맨드
- trust5.mdx: TRUST 5 품질 보증 원칙

**PART C: 워크플로우 → Workflows**
- tdd-cycle.mdx: RED-GREEN-REFACTOR 사이클
- git-integration.mdx: 3-Mode Git 전략 (Manual/Personal/Team)
- spec-to-code.mdx: SPEC → 구현 → 문서화 흐름
- multi-agent.mdx: 다중 에이전트 조합 패턴

**PART D: 고급 기능 → Advanced**
- agents-guide.mdx: 26개 에이전트 상세 가이드
- skills-library.mdx: 22개 스킬 카드 및 레퍼런스
- patterns.mdx: 고급 조합 패턴 및 실전 예제
- trust5-quality.mdx: 품질 보증 상세 가이드
- performance-optimization.mdx: 성능 최적화 가이드

### 4.3 22개 Skills 상세 문서화

**Connector Skills (4개)**:
- moai-connector-figma: Figma 디자인 시스템 연동
- moai-connector-mcp: MCP 서버 통합 (Context7, Sequential Thinking)
- moai-connector-nano-banana: AI 모델 연결 및 활용
- moai-connector-notion: Notion 워크스페이스 통합

**Foundation Skills (4개)**:
- moai-foundation-claude: Claude Code 최적화 및 설정
- moai-foundation-context: 컨텍스트 및 세션 관리
- moai-foundation-quality: 품질 게이트 및 테스트
- moai-foundation-uiux: UI/UX 디자인 가이드

**Library Skills (5개)**:
- moai-lang-unified: 25개 프로그래밍 언어 통합 전문가
- moai-library-mermaid: Mermaid 다이어그램 생성
- moai-library-nextra: Nextra 문서 사이트 구축
- moai-library-shadcn: shadcn/ui 컴포넌트 활용
- moai-library-toon: 툰 스타일 디자인 시스템

**Platform & Workflow Skills (9개)**:
- moai-platform-baas: BaaS 플랫폼 연동 (Supabase, Firebase)
- moai-system-universal: 범용 시스템 통합
- moai-toolkit-essentials: 필수 도구 및 유틸리티
- moai-workflow-docs: 문서 생성 자동화
- moai-workflow-jit-docs: JIT 문서 생성
- moai-workflow-project: 프로젝트 초기화
- moai-workflow-templates: 템플릿 관리
- moai-workflow-testing: 테스트 자동화
- (기타 추가 스킬)

**각 스킬별 문서 구조**:
```mdx
# {Skill Name}

## 개요
- 목적 및 사용 사례
- 핵심 기능 요약

## Quick Reference (30초)
- 주요 기능 요약
- Auto-trigger 조건
- 핵심 패턴

## Implementation Guide (5분)
- 기능 설명
- 사용 시기
- 핵심 패턴 예제

## 5 Core Patterns
- Pattern 1-5 상세 설명
- 코드 예제 및 사용 사례

## Advanced Documentation
- 관련 모듈 링크
- API 레퍼런스

## Works Well With
- 연관 스킬 및 에이전트

## Best Practices
- DO / DON'T 가이드
```

### 4.4 명령어 레퍼런스 완전 작성

**각 명령어별 상세 문서**:

**1. /moai:0-project (프로젝트 초기화)**
- 기능: MoAI-ADK 프로젝트 구조 생성
- 옵션: 언어, 모드, Git 전략
- 예제: 다양한 프로젝트 타입별 초기화

**2. /moai:1-plan (SPEC 생성)**
- 기능: EARS 기반 SPEC 문서 생성
- 자동 제안 vs 수동 지정
- SPEC 템플릿 및 검증

**3. /moai:2-run (TDD 구현)**
- 기능: SPEC 기반 TDD 사이클 실행
- RED-GREEN-REFACTOR 자동화
- 다중 에이전트 조합

**4. /moai:3-sync (문서 동기화)**
- 기능: SPEC → 문서 자동 생성
- Draft PR 생성 및 Git 통합
- 검증 및 품질 게이트

**5. /moai:9-feedback (피드백 제출)**
- 기능: 개선 제안 및 버그 리포트
- 자동 이슈 생성
- 피드백 루프 통합

**6. /clear (컨텍스트 초기화)**
- 기능: 200K 토큰 버짓 관리
- 세션 상태 보존
- 최적화 전략

### 4.5 Advanced 섹션 완전 작성

**agents-guide.mdx (26개 에이전트 상세 가이드)**
- 5-Tier 계층 구조 설명
- 각 에이전트별 역할 및 사용 시기
- 에이전트 조합 패턴 및 예제
- Task() 호출 및 Handoff 프로토콜

**skills-library.mdx (22개 스킬 카드 및 레퍼런스)**
- 스킬 카테고리별 분류
- 각 스킬 카드 (이름, 아이콘, 설명, 링크)
- 스킬 검색 및 필터링 기능
- 관련 스킬 추천 시스템

**patterns.mdx (고급 조합 패턴)**
- Sequential vs Parallel 실행
- Conditional Delegation 패턴
- MCP Resume 패턴
- Multi-Agent Coordination

**trust5-quality.mdx (TRUST 5 원칙 상세)**
- Testable: 테스트 가능성 확보
- Reproducible: 재현 가능성 보장
- Understandable: 이해 가능성 향상
- Secure: 보안 및 권한 관리
- Trackable: 추적 가능성 유지

**performance-optimization.mdx (성능 최적화 가이드)**
- 토큰 버짓 관리 (200K 최적화)
- Context Engineering 전략
- Aggressive /clear 전략
- MCP 서버 최적화
- Memory 파일 최적화 (<500줄)

### 4.6 API 레퍼런스 완전 생성

**src/moai_adk/ 모듈 구조 분석**:
```
src/moai_adk/
├── __init__.py           # 패키지 진입점
├── cli/                  # CLI 명령어
│   ├── project.py
│   ├── spec.py
│   └── worktree.py
├── core/                 # 핵심 로직
│   ├── agents/
│   ├── skills/
│   └── config/
├── templates/            # 프로젝트 템플릿
└── utils/                # 유틸리티 함수
```

**API 레퍼런스 페이지 구조**:
- **reference/api/cli.mdx**: CLI 명령어 API
- **reference/api/config.mdx**: 설정 파일 API
- **reference/api/agents.mdx**: 에이전트 API
- **reference/api/skills.mdx**: 스킬 API
- **reference/api/utils.mdx**: 유틸리티 함수 API

**각 API 문서 형식**:
```mdx
## 함수명

**시그니처**: `function_name(param1: Type, param2: Type) -> ReturnType`

**설명**: 함수의 역할 및 목적

**매개변수**:
- `param1` (Type): 설명
- `param2` (Type): 설명

**반환값**: ReturnType - 설명

**예제**:
```python
result = function_name("value1", "value2")
print(result)
```

**관련 함수**: 연관 API 링크
```

### 4.7 링크 검증 및 최종 통합

**내부 링크 검증**:
- 모든 페이지 간 링크 자동 생성
- 깨진 링크 감지 및 수정
- 앵커 링크 (#) 정상 작동 확인
- Table of Contents 자동 생성

**검색 기능 테스트**:
- 전체 텍스트 검색 정확도
- 섹션별 검색 결과
- 검색 성능 (300ms 이내)
- 키보드 단축키 (Cmd/Ctrl + K)

**성능 검증 (Lighthouse 90+)**:
- Performance: 90+ 점수
- Accessibility: 90+ 점수
- Best Practices: 90+ 점수
- SEO: 90+ 점수

**최종 통합 체크리스트**:
- [ ] README.ko.md 1,773줄 100% 마이그레이션
- [ ] 22개 Skills 각각 상세 페이지 생성
- [ ] 6개 명령어 레퍼런스 완성
- [ ] Advanced 섹션 65바이트 → 3,000+줄 확장
- [ ] API 레퍼런스 완성 (50+ 모듈)
- [ ] 모든 내부 링크 정상 작동
- [ ] 검색 기능 완전 동작
- [ ] Lighthouse 90+ 점수 달성
- [ ] 모바일 반응형 완벽 지원
- [ ] 접근성 WCAG 2.1 AA 준수

---

**Phase 4 확장 완료 기준**:
- 전체 콘텐츠 마이그레이션 100% 완료
- 모든 스킬 및 명령어 문서화
- 성능 및 접근성 기준 달성
- 사용자 테스트 통과 (만족도 4.5/5.0)