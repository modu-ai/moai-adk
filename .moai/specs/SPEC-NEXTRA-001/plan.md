---
id: SPEC-NEXTRA-001
version: "1.1.0"
status: "draft"
created: "2025-11-28"
updated: "2025-11-28"
author: "GOOS"
priority: "HIGH"
---

# SPEC-NEXTRA-001 구현 계획

## 개요

MoAI-ADK 온라인 문서 사이트를 Nextra 기반으로 구축하기 위한 6단계 구현 계획입니다. 각 단계는 명확한 목표와 배포 가능한 산출물을 제공하며, 점진적으로 기능을 확장해 나갑니다.

## 구현 전략

### 핵심 원칙

1. **점진적 구현**: MVP → 전체 기능 → 최적화 순서로 진행
2. **자동화 우선**: 콘텐츠 마이그레이션과 빌드 프로세스 자동화
3. **성능 중심**: 모든 단계에서 성능 메트릭 측정 및 개선
4. **사용자 중심**: 실제 사용 시나리오 기반의 테스트와 검증

### 기술 의존성

- **Next.js 15+**: React 19, App Router, Turbopack
- **Nextra 3.0+**: MDX 기반 정적 사이트 생성
- **TypeScript**: 타입 안전성 확보
- **Vercel**: 배포 및 호스팅
- **manager-docs**: 문서 생성 자동화

---

## Phase 1: 기반 설정 (Foundation Setup)

**기간**: 3-5일
**담당**: manager-docs
**목표**: Nextra 프로젝트 기반 구축 및 기본 테마 적용

### 1.1 프로젝트 초기화
```bash
# Nextra 프로젝트 생성
npx create-next-app@latest moai-docs --typescript --tailwind --app
cd moai-docs
npm install nextra nextra-theme-docs

# 필수 의존성 추가
npm install @types/node @types/react @types/react-dom
```

### 1.2 기본 설정
- **next.config.js**: Nextra 설정, 이미지 최적화, redirects
- **theme.config.tsx**: MoAI-ADK 브랜딩, 네비게이션 구조
- **package.json**: 스크립트, 개발 의존성

### 1.3 CSS 시스템 구축
- **globals.css**: 그레이스케일 테마 변수
- **웹폰트 설정**: Pretendard, Inter, JetBrains Mono
- **다크 모드**: CSS 변수 기반 테마 전환

### 1.4 기본 페이지 구조
```
docs/
├── pages/
│   ├── index.mdx              # 홈페이지
│   ├── getting-started/
│   │   └── _meta.js          # 네비게이션 메타
│   └── meta.json             # 전역 메타데이터
└── styles/
    └── globals.css           # 기본 스타일
```

### 산출물
- [ ] 작동하는 Nextra 프로젝트
- [ ] 기본 라이트/다크 테마
- [ ] 로컬 개발 환경
- [ ] Vercel 배포 설정

### 성공 기준
- [ ] 로컬에서 `npm run dev`로 정상 실행
- [ ] 테마 전환이 즉시 적용됨
- [ ] Vercel에 첫 배포 성공

---

## Phase 2: 콘텐츠 마이그레이션 (Content Migration)

**기간**: 4-6일
**담당**: manager-docs + moai-library-nextra
**목표**: README.ko.md와 기존 문서를 Nextra 구조로 변환

### 2.1 README.ko.md 분석 및 구조화
```python
# 마이그레이션 스크립트 예시
def migrate_readme():
    sections = parse_readme_sections("README.ko.md")

    mapping = {
        "PART A": "getting-started/",
        "PART B": "core-concepts/",
        "PART C": "advanced/",
        "PART D": "reference/"
    }

    for part, target_dir in mapping.items():
        convert_to_mdx(sections[part], target_dir)
```

### 2.2 페이지 구조 생성
- **getting-started/**: 설치, 설정, 빠른 시작
- **core-concepts/**: SPEC, 에이전트, 워크플로우, 커맨드
- **advanced/**: 심화 가이드, 스킬, 패턴
- **reference/**: API 레퍼런스, FAQ

### 2.3 MDX 변환 규칙
- **Mermaid 다이어그램**: React 컴포넌트로 변환
- **테이블**: Markdown 호환 형식으로 변환
- **코드 블록**: 구문 강조와 복사 기능 추가
- **이미지**: 최적화 및 CDN 이전

### 2.4 네비게이션 메타데이터
```javascript
// getting-started/_meta.js
export default {
  index: { title: "시작하기" },
  installation: { title: "설치 및 설정" },
  quickstart: { title: "빠른 시작" }
}
```

### 산출물
- [ ] 변환된 모든 페이지 (15-20개)
- [ ] 네비게이션 메타데이터
- [ ] 내부 링크 자동 생성
- [ ] 검색 기능 활성화

### 성공 기준
- [ ] 모든 README 콘텐츠가 이전됨
- [ ] 네비게이션이 논리적으로 동작함
- [ ] 링크 깨짐 없음 (100% 유효)
- [ ] 검색이 정상적으로 작동함

---

## Phase 3: 스킬 및 명령어 문서화 (Skills & Commands Documentation)

**기간**: 3-4일
**담당**: manager-docs
**목표**: 22개 스킬과 6개 커맨드에 대한 상세 문서 생성

### 3.1 스킬 라이브러리 페이지 구축
```
advanced/skills-library.mdx
├── 스킬 개요 및 카테고리
├── 스킬 카드 그리드 (22개)
└── 각 스킬별 상세 페이지로 링크
```

### 3.2 개별 스킬 페이지 자동 생성
```python
# 스킬 문서 생성 스크립트
def generate_skill_pages():
    for skill in load_skills(".claude/skills/"):
        skill_doc = {
            "title": skill.name,
            "description": skill.description,
            "usage": skill.examples,
            "api": skill.api_reference
        }
        write_mdx(f"reference/skills/{skill.id}.mdx", skill_doc)
```

### 3.3 명령어 레퍼런스
- **/moai:0-project**: 프로젝트 초기화
- **/moai:1-plan**: SPEC 생성
- **/moai:2-run**: TDD 구현
- **/moai:3-sync**: 문서 동기화

### 3.4 대화형 예제
```mdx
{/* 실제 실행 가능한 코드 예제 */}
<CodeBlock runnable language="bash">
/moai:1-plan "user authentication"
</CodeBlock>
```

### 산출물
- [ ] 스킬 라이브러리 메인 페이지
- [ ] 22개 스킬 상세 페이지
- [ ] 6개 명령어 레퍼런스
- [ ] 대화형 코드 예제

### 성공 기준
- [ ] 모든 스킬이 문서화됨
- [ ] API 레퍼런스가 정확함
- [ ] 코드 예제가 실제로 동작함
- [ ] 검색으로 스킬을 찾을 수 있음

---

## Phase 4: 전체 콘텐츠 마이그레이션 (Complete Content Migration) - 확장

**목표**: README.ko.md 1,773줄, 22개 Skills, Advanced 섹션 완전 마이그레이션

> **Note**: 본 Phase는 기존 워크트리 통합을 포함하여 전체 콘텐츠 마이그레이션으로 확장되었습니다.

### Step 4.1: README.ko.md 완전 분석 및 파싱

**담당**: manager-docs
**선행 조건**: Phase 2 완료 (기본 콘텐츠 마이그레이션)

**작업 내용**:
1. README.ko.md 1,773줄 전체 파싱
   - 섹션 헤더 추출 (## 기준)
   - PART A-D 분류 및 매핑
   - 코드 블록, 표, 링크 추출

2. 구조화 스크립트 작성
   ```python
   # scripts/parse_readme.py
   def parse_readme_sections(file_path: str) -> dict:
       """README.ko.md를 PART별로 파싱"""
       sections = {
           "PART_A": [],  # 시작하기
           "PART_B": [],  # 핵심 개념
           "PART_C": [],  # 워크플로우
           "PART_D": []   # 고급 기능
       }
       # 파싱 로직 구현
       return sections
   ```

3. MDX 변환 규칙 정의
   - Markdown → MDX 자동 변환
   - Mermaid 다이어그램 → React 컴포넌트
   - 내부 링크 자동 생성

**산출물**:
- [ ] README 파싱 스크립트
- [ ] 섹션별 분류 데이터 (JSON)
- [ ] MDX 변환 규칙 문서

**예상 시간**: 3-4시간

### Step 4.2: 핵심 콘텐츠 페이지 생성 (30-35개 페이지)

**담당**: manager-docs + moai-library-nextra

**PART A: Getting Started (5-7 페이지)**:
- [ ] `overview.mdx`: MoAI-ADK 개요
- [ ] `installation.mdx`: 설치 가이드 (pip, uv)
- [ ] `quickstart.mdx`: 5분 빠른 시작
- [ ] `first-spec.mdx`: 첫 SPEC 생성
- [ ] `configuration.mdx`: config.json 설정

**PART B: Core Concepts (8-10 페이지)**:
- [ ] `spec-format.mdx`: SPEC과 EARS 포맷
- [ ] `agents.mdx`: Mr.Alfred와 26개 에이전트
- [ ] `workflow.mdx`: SPEC-First TDD
- [ ] `commands.mdx`: /moai:0-3 커맨드
- [ ] `trust5.mdx`: TRUST 5 원칙

**PART C: Workflows (6-8 페이지)**:
- [ ] `tdd-cycle.mdx`: RED-GREEN-REFACTOR
- [ ] `git-integration.mdx`: 3-Mode Git 전략
- [ ] `spec-to-code.mdx`: SPEC → 코드 흐름
- [ ] `multi-agent.mdx`: 다중 에이전트 조합

**PART D: Advanced (10-12 페이지)**:
- [ ] `agents-guide.mdx`: 26개 에이전트 상세
- [ ] `skills-library.mdx`: 22개 스킬 카드
- [ ] `patterns.mdx`: 고급 조합 패턴
- [ ] `trust5-quality.mdx`: 품질 보증 상세
- [ ] `performance-optimization.mdx`: 성능 최적화

**자동화 스크립트**:
```python
# scripts/generate_pages.py
def generate_page_from_section(section_data: dict, template: str) -> str:
    """섹션 데이터를 MDX 페이지로 변환"""
    mdx_content = apply_template(section_data, template)
    return mdx_content
```

**산출물**:
- [ ] 30-35개 MDX 페이지
- [ ] 각 섹션별 _meta.js
- [ ] 내부 링크 자동 생성

**예상 시간**: 8-10시간

### Step 4.3: 22개 Skills 상세 문서화

**담당**: manager-docs

**스킬 카테고리별 문서 생성**:

**Connector Skills (4개)**: 각 1시간 = 4시간
- [ ] moai-connector-figma.mdx
- [ ] moai-connector-mcp.mdx
- [ ] moai-connector-nano-banana.mdx
- [ ] moai-connector-notion.mdx

**Foundation Skills (4개)**: 각 1시간 = 4시간
- [ ] moai-foundation-claude.mdx
- [ ] moai-foundation-context.mdx
- [ ] moai-foundation-quality.mdx
- [ ] moai-foundation-uiux.mdx

**Library Skills (5개)**: 각 1시간 = 5시간
- [ ] moai-lang-unified.mdx
- [ ] moai-library-mermaid.mdx
- [ ] moai-library-nextra.mdx
- [ ] moai-library-shadcn.mdx
- [ ] moai-library-toon.mdx

**Platform & Workflow Skills (9개)**: 각 30분 = 4.5시간
- [ ] moai-platform-baas.mdx
- [ ] moai-system-universal.mdx
- [ ] moai-toolkit-essentials.mdx
- [ ] moai-workflow-docs.mdx
- [ ] moai-workflow-jit-docs.mdx
- [ ] moai-workflow-project.mdx
- [ ] moai-workflow-templates.mdx
- [ ] moai-workflow-testing.mdx
- [ ] (기타)

**각 스킬 문서 템플릿**:
```mdx
---
title: "{Skill Name}"
description: "{One-line description}"
---

# {Skill Name}

## Quick Reference (30초)
- 핵심 기능 요약
- Auto-trigger 조건
- 주요 패턴

## Implementation Guide (5분)
- 사용 시기
- 핵심 패턴 3개
- 코드 예제

## 5 Core Patterns
- Pattern 1-5 상세 설명

## Works Well With
- 연관 스킬/에이전트

## Best Practices
- DO / DON'T
```

**산출물**:
- [ ] 22개 스킬 상세 페이지
- [ ] reference/skills/ 디렉토리 구조
- [ ] 스킬 검색 인덱스

**예상 시간**: 17-18시간 (단, 병렬 작업 가능)

### Step 4.4: 명령어 레퍼런스 완전 작성 (6개 명령어)

**담당**: manager-docs

**각 명령어별 상세 페이지**: 각 1시간 = 6시간

**1. /moai:0-project**:
```mdx
# /moai:0-project - 프로젝트 초기화

## 개요
MoAI-ADK 프로젝트 구조 생성

## 사용법
```bash
/moai:0-project
```

## 옵션
- 언어 선택
- Git 전략 설정
- 모드 선택

## 예제
[5-7가지 시나리오별 예제]
```

**명령어 문서 구조**:
- [ ] /moai:0-project.mdx (프로젝트 초기화)
- [ ] /moai:1-plan.mdx (SPEC 생성)
- [ ] /moai:2-run.mdx (TDD 구현)
- [ ] /moai:3-sync.mdx (문서 동기화)
- [ ] /moai:9-feedback.mdx (피드백 제출)
- [ ] /clear.mdx (컨텍스트 초기화)

**산출물**:
- [ ] 6개 명령어 레퍼런스
- [ ] reference/commands/ 디렉토리
- [ ] 명령어 간 링크 및 추천

**예상 시간**: 6시간

### Step 4.5: Advanced 섹션 완전 작성 (5개 핵심 페이지)

**담당**: manager-docs + manager-quality

**1. agents-guide.mdx (26개 에이전트)**: 4시간
- 5-Tier 계층 구조
- 각 에이전트 역할 및 사용 시기
- Task() 호출 패턴
- Handoff 프로토콜

**2. skills-library.mdx (스킬 카드 그리드)**: 3시간
- 22개 스킬 카드 레이아웃
- 카테고리별 필터링
- 스킬 검색 기능
- 관련 스킬 추천

**3. patterns.mdx (고급 조합 패턴)**: 2시간
- Sequential vs Parallel
- Conditional Delegation
- MCP Resume 패턴
- Multi-Agent Coordination

**4. trust5-quality.mdx (TRUST 5 원칙)**: 2시간
- Testable, Reproducible, Understandable, Secure, Trackable
- 각 원칙별 상세 가이드
- 실전 적용 예제

**5. performance-optimization.mdx (성능 가이드)**: 2시간
- 200K 토큰 버짓 관리
- Context Engineering
- Aggressive /clear 전략
- MCP 서버 최적화

**산출물**:
- [ ] 5개 Advanced 핵심 페이지
- [ ] 3,000+줄 콘텐츠 (현재 65바이트에서 확장)
- [ ] 실전 예제 및 다이어그램

**예상 시간**: 13시간

### Step 4.6: API 레퍼런스 생성 (50+ 모듈)

**담당**: manager-docs

**API 문서 자동 생성**:
```python
# scripts/generate_api_docs.py
def extract_api_signatures(module_path: str) -> list:
    """모듈에서 함수 시그니처 추출"""
    # AST 파싱으로 함수/클래스 추출
    return signatures

def generate_api_page(signatures: list, template: str) -> str:
    """API 레퍼런스 페이지 생성"""
    return mdx_content
```

**API 레퍼런스 구조**:
- [ ] reference/api/cli.mdx (CLI 명령어 API)
- [ ] reference/api/config.mdx (설정 API)
- [ ] reference/api/agents.mdx (에이전트 API)
- [ ] reference/api/skills.mdx (스킬 API)
- [ ] reference/api/utils.mdx (유틸리티 API)

**각 API 문서 포함 내용**:
- 함수 시그니처
- 매개변수 설명
- 반환값 설명
- 사용 예제
- 관련 함수 링크

**산출물**:
- [ ] 5개 API 카테고리 페이지
- [ ] 50+ 모듈 레퍼런스
- [ ] 타입 정의 및 예제

**예상 시간**: 8-10시간

### Step 4.7: 워크트리 CLI 통합 및 최종 검증

**담당**: manager-docs + manager-quality

**워크트리 문서 마이그레이션**:
- [ ] WORKTREE_GUIDE.md → worktree/guide.mdx
- [ ] WORKTREE_FAQ.md → worktree/faq.mdx
- [ ] WORKTREE_EXAMPLES.md → worktree/examples.mdx

**링크 검증 및 최종 통합**:
1. 내부 링크 검증 (100% 유효)
2. 검색 기능 테스트 (300ms 이내)
3. 성능 검증 (Lighthouse 90+)
4. 접근성 검증 (WCAG 2.1 AA)

**최종 체크리스트**:
- [ ] README.ko.md 1,773줄 100% 마이그레이션
- [ ] 22개 Skills 각각 상세 페이지
- [ ] 6개 명령어 레퍼런스 완성
- [ ] Advanced 섹션 3,000+줄 확장
- [ ] API 레퍼런스 50+ 모듈 완성
- [ ] 모든 내부 링크 정상 작동
- [ ] 검색 기능 완전 동작
- [ ] Lighthouse 90+ 점수
- [ ] 모바일 반응형 지원
- [ ] 접근성 기준 준수

**산출물**:
- [ ] 워크트리 통합 완료
- [ ] 링크 검증 리포트
- [ ] 성능 테스트 리포트
- [ ] 최종 통합 문서

**예상 시간**: 4-5시간

---

### Phase 4 전체 요약

**총 예상 시간**: 50-60시간 (병렬 작업 시 30-40시간)

**7단계 실행 순서**:
```
Step 4.1 (3-4h) → Step 4.2 (8-10h) → Step 4.3 (17-18h)
                                      ↓
Step 4.4 (6h) → Step 4.5 (13h) → Step 4.6 (8-10h) → Step 4.7 (4-5h)
```

**병렬 작업 전략**:
- Step 4.3 (스킬 문서화)와 Step 4.4 (명령어 레퍼런스)는 동시 진행 가능
- Step 4.5 (Advanced)와 Step 4.6 (API)도 일부 병렬 가능
- manager-docs + moai-library-nextra 2개 에이전트 동시 작업 시 30% 시간 단축

**Phase 4 완료 시 달성**:
- ✅ 전체 콘텐츠 마이그레이션 100%
- ✅ 모든 스킬 및 명령어 문서화
- ✅ Advanced 섹션 65바이트 → 3,000+줄
- ✅ API 레퍼런스 완성
- ✅ 성능 및 접근성 기준 달성

---

## Phase 5: 성능 최적화 (Performance Optimization)

**기간**: 3-4일
**담당**: manager-docs + manager-quality
**목표**: Lighthouse 90+ 점수 달성 및 Core Web Vitals 최적화

### 5.1 Core Web Vitals 최적화

#### LCP (Largest Contentful Paint)
- **이미지 최적화**: WebP 형식, 반응형 로딩
- **폰트 최적화**: WOFF2, 폰트 디스플레이 전략
- **Critical CSS**: 인라인 스타일, 비동기 로딩

#### INP (Interaction to Next Paint)
- **JavaScript 번들링**: 코드 스플리팅, 트리 쉐이킹
- **React 최적화**: memo, useMemo, 적절한 상태 관리
- **이벤트 핸들러**: 디바운싱, 스로틀링

#### CLS (Cumulative Layout Shift)
- **이미지 차원**: 명시적인 width/height 지정
- **폰트 로딩**: FOUT/FOIT 방지
- **동적 콘텐츠**: 레이아웃 이동 최소화

### 5.2 빌드 최적화
```javascript
// next.config.js
module.exports = {
  experimental: {
    turbotrace: true,  // 빠른 빌드
    optimizeCss: true  // CSS 최적화
  },
  images: {
    domains: ['cdn.example.com'],
    formats: ['image/webp', 'image/avif']
  }
}
```

### 5.3 CDN 및 캐싱 전략
- **정적 에셋**: Vercel Edge Network
- **페이지 수준 캐싱**: ISR (Incremental Static Regeneration)
- **API 라우트**: 캐시 헤더 최적화

### 5.4 모니터링 설정
- **Lighthouse CI**: 자동 성능 테스트
- **Core Web Vitals**: 실제 사용자 데이터 수집
- **Sentry**: 에러 추적 및 성능 모니터링

### 산출물
- [ ] 최적화된 빌드 설정
- [ ] 성능 모니터링 대시보드
- [ ] Lighthouse 90+ 점수 보고서
- [ ] Core Web Vitals 달성 증명

### 성공 기준
- [ ] Lighthouse 점수: 90+ (Performance)
- [ ] LCP: 2.5초 이하
- [ ] INP: 200ms 이하
- [ ] CLS: 0.1 이하

---

## Phase 6: 배포 및 안정화 (Deployment & Stabilization)

**기간**: 2-3일
**담당**: manager-docs + manager-git
**목표**: 프로덕션 배포 및 지속적인 유지보수 체계 구축

### 6.1 프로덕션 배포
- **도메인 설정**: docs.moai-ai.dev
- **SSL 인증서**: 자동 갱신
- **환경 변수**: 안전한 설정 관리
- **모니터링**: 상태 확인 및 알림

### 6.2 CI/CD 파이프라인
```yaml
# .github/workflows/docs.yml
name: Documentation
on:
  push:
    branches: [main]
    paths: ['docs/**']
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build and Deploy
        run: |
          npm ci
          npm run build
          npm run deploy
```

### 6.3 검토 및 테스트
- **내부 테스트**: 5명 이상 사용자 그룹
- **접근성 테스트**: WCAG 2.1 AA 준수 검증
- **호환성 테스트**: 주요 브라우저 지원 확인
- **모바일 테스트**: 반응형 동작 검증

### 6.4 유지보수 체계
- **정기 업데이트**: 주간 문서 동기화
- **버그 리포트**: GitHub Issues 통합
- **개선 제안**: 사용자 피드백 수집
- **성능 모니터링**: 월간 리포트

### 산출물
- [ ] 프로덕션 배포된 문서 사이트
- [ ] CI/CD 파이프라인
- [ ] 사용자 테스트 보고서
- [ ] 유지보수 가이드

### 성공 기준
- [ ] https://docs.moai-ai.dev 정상 접속
- [ ] CI/CD 파이프라인 자동 동작
- [ ] 사용자 만족도 4.5/5.0
- [ ] 접근성 준수 100%

---

## 위험 관리 (Risk Management)

### HIGH 리스크
- **콘텐츠 마이그레이션 복잡성**: 자동화 스크립트 선행 개발
- **성능 목표 달성**: 단계별 프로파일링 및 최적화

### MEDIUM 리스크
- **브라우저 호환성**: 모던 브라우저 지원으로 범위 축소
- **사용자 채택**: 초기 MVP로 피드백 수집 및 개선

### LOW 리스크
- **배포 문제**: Vercel의 안정적인 인프라 활용
- **기술 부채**: TypeScript와 정적 분석으로 최소화

---

## 성공 측정 (Success Metrics)

### 정량적 지표
- **성능**: Lighthouse 90+ 점수
- **가용성**: 99.9% 업타임
- **문서 완성도**: 100% (모든 기능 문서화)
- **사용자 만족도**: 4.5/5.0 이상

### 정성적 지표
- **검색 효율성**: 원하는 정보를 10초 내에 찾기
- **학습 곡선**: 새 사용자가 30분 내에 핵심 기능 이해
- **유지보수 용이성**: 주 2시간 이내의 정기 업데이트

---

## 다음 단계 (Next Steps)

1. **즉시 시작**: Phase 1 기반 설정
2. **전문 에이전트**: manager-docs에 구현 위임
3. **정기 검토**: 주간 진행 상황 확인
4. **성과 측정**: 각 Phase 완료 시 성공 기준 검증

**예상 완료 시점**: 6주 후
**총 예상 노력**: 60-80 시간
**주요 의존성**: manager-docs, moai-library-nextra 스킬