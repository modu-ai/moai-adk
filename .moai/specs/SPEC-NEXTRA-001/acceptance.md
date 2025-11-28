---
id: SPEC-NEXTRA-001
version: "1.2.0"
status: "ACCEPTED"
created: "2025-11-28"
updated: "2025-11-29"
author: "GOOS"
priority: "HIGH"
verification_date: "2025-11-29"
---

# SPEC-NEXTRA-001 수용 기준 (Acceptance Criteria) - ACCEPTED

## 개요

MoAI-ADK 온라인 문서 사이트 구축 프로젝트의 완료를 위한 구체적인 수용 기준입니다. 기능적, 비기능적, 사용성 관점에서의 성공 조건을 정의하고, 각 기준에 대한 검증 방법을 포함합니다.

---

## Given-When-Then 시나리오 (최소 2개 이상)

### Scenario 1: 홈페이지 탐색 및 정보 검색

**GIVEN** 사용자가 MoAI-ADK 문서 사이트에 처음 방문했을 때
**WHEN** 홈페이지에서 "빠른 시작" 섹션을 탐색하고 검색창에 "SPEC 생성"을 입력했을 때
**THEN** 관련 SPEC 생성 가이드와 /moai:1-plan 명령어 레퍼런스로 바로 이동할 수 있어야 하며, 해당 페이지에서 실제 사용 예제를 볼 수 있어야 한다

**검증 항목**:
- [ ] 홈페이지 로딩 시간: 2초 이내
- [ ] 검색 결과 응답 시간: 300ms 이내
- [ ] 검색 결과 정확도: 관련성 상위 5개 결과 중 3개 이상 관련
- [ ] 내부 링크 유효성: 100% 정상 작동
- [ ] 테마 전환: 라이트/다크 모드 즉시 적용

### Scenario 2: Git Worktree CLI 문서 접근

**GIVEN** 개발자가 SPEC-WORKTREE-001의 워크트리 기능을 학습하고 싶을 때
**WHEN** 사이드바에서 "Worktree" 섹션을 클릭하고 "사용 예제" 페이지로 이동했을 때
**THEN** 모든 워크트리 명령어(moai-worktree create, switch, list 등)의 상세 사용법과 실제 터미널 실행 예제를 볼 수 있어야 하며, 다른 페이지로 쉽게 이동할 수 있어야 한다

**검증 항목**:
- [ ] 워크트리 섹션 접근 용이성: 3번 클릭 이내 도달
- [ ] 명령어 레퍼런스 완전성: 8개 핵심 명령어 모두 문서화
- [ ] 코드 예제 복사 기능: 1번 클릭으로 코드 복사 가능
- [ ] 관련 페이지 연결: 가이드 ↔ 예제 ↔ FAQ 간 원활한 이동
- [ ] 모바일 반응성: 스마트폰에서도 모든 내용 가독성 확보

### Scenario 3: 스킬 라이브러리 탐색

**GIVEN** 개발자가 MoAI-ADK의 전문 스킬을 찾아보고 싶을 때
**WHEN** "심화 학습" > "스킬 라이브러리" 페이지에 방문하여 "moai-foundation-core" 스킬을 클릭했을 때
**THEN** 해당 스킬의 상세 설명, 사용 방법, API 레퍼런스, 실제 적용 예제를 볼 수 있어야 하며, 관련 스킬로의 추천 링크도 제공되어야 한다

**검증 항목**:
- [ ] 스킬 라이브러리 로딩: 22개 스킬 카드 3초 내 로딩
- [ ] 스킬 상세 페이지: 모든 스킬에 대한 상세 정보 제공
- [ ] API 레퍼런스 정확성: 실제 스킬 API와 일치
- [ ] 관련 스킬 추천: 최소 3개 관련 스킬 링크
- [ ] 북마크 기능: 주요 페이지 저장 및 빠른 접근

### Scenario 4: 테마 및 접근성 테스트

**GIVEN** 시각 장애가 있는 사용자가 문서 사이트를 이용할 때
**WHEN** 키보드만으로 사이트를 탐색하고 스크린 리더로 콘텐츠를 읽었을 때
**THEN** 모든 기능을 정상적으로 사용할 수 있어야 하며, 다크 모드에서도 충분한 대비가 보장되어야 한다

**검증 항목**:
- [ ] 키보드 내비게이션: Tab 키로 모든 인터랙티브 요소 접근
- [ ] 스크린 리더 호환성: 페이지 구조 논리적으로 읽힘
- [ ] 색상 대비: WCAG 2.1 AA 기준 충족 (4.5:1 이상)
- [ ] 포커스 표시: 명확한 시각적 포커스 제공
- [ ] 폰트 크기 조절: 200% 확장에서도 가독성 유지

---

## 성능 요구사항 (Performance Requirements)

### Core Web Vitals 기준

| 지표 | 목표치 | 측정 방법 | 임계값 |
|------|--------|------------|--------|
| **LCP (Largest Contentful Paint)** | 2.5초 이하 | Lighthouse, Chrome DevTools | 4.0초 |
| **INP (Interaction to Next Paint)** | 200ms 이하 | Chrome User Experience Report | 500ms |
| **CLS (Cumulative Layout Shift)** | 0.1 이하 | Lighthouse, 실제 사용 데이터 | 0.25 |
| **FCP (First Contentful Paint)** | 1.8초 이하 | Lighthouse | 3.0초 |
| **TTI (Time to Interactive)** | 3.8초 이하 | Lighthouse | 7.3초 |

### 검증 방법
```bash
# Lighthouse 자동화 테스트
npm run lighthouse

# Core Web Vitals 측정
npm run measure-performance

# 번들 크기 분석
npm run analyze-bundle
```

### 성공 기준
- [ ] Lighthouse 점수: Performance 90+ 이상
- [ ] 모든 Core Web Vitals 임계값 충족
- [ ] 번들 크기: 초기 로드 500KB 이하
- [ ] 이미지 최적화: WebP 형식, 적절한 크기

---

## 품질 요구사항 (Quality Requirements)

### 콘텐츠 품질

**문서 완성도 (Phase 4 확장 반영)**:
- [ ] README.ko.md 1,773줄 100% 마이그레이션
- [ ] 22개 스킬 각각 상세 페이지 생성 (Quick Reference, 5 Core Patterns, Best Practices 포함)
- [ ] 6개 코어 커맨드 상세 레퍼런스 (각 1,000+줄 수준)
- [ ] 워크트리 CLI 문서 완전 통합 (guide, examples, faq)
- [ ] Advanced 섹션 65바이트 → 3,000+줄 확장
- [ ] API 레퍼런스 50+ 모듈 완성

**정확성**:
- [ ] 모든 코드 예제 실제 실행 가능
- [ ] API 레퍼런스와 실제 구현 일치
- [ ] 링크 깨짐 없음 (100% 유효)
- [ ] 최신 버전 기준으로 콘텐츠 유지
- [ ] 스킬 문서 구조 통일성 (모든 스킬이 동일한 템플릿 사용)

### 기술 품질

**코드 품질**:
```javascript
// ESLint 설정
{
  "extends": ["next/core-web-vitals"],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "react-hooks/exhaustive-deps": "warn"
  }
}

// TypeScript 엄격 모드
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true
  }
}
```

**검증 항목**:
- [ ] ESLint 오류 0개
- [ ] TypeScript 컴파일 경고 없음
- [ ] 코드 커버리지 90% 이상 (테스트)
- [ ] 접근성 검토 통과 (axe-core)

### 보안 품질

**보안 검증**:
- [ ] XSS 방지: 모든 사용자 입력 sanitize
- [ ] HTTPS 강제: 모든 요청 HTTPS로 리디렉션
- [ ] CSP 헤더 설정: Content Security Policy 적용
- [ ] 의존성 취약점 검사: npm audit 통과

---

## 사용성 요구사항 (Usability Requirements)

### 내비게이션 및 검색

**사이트 내비게이션**:
- [ ] 모든 페이지 3번 클릭 이내 도달 가능
- [ ] 브레드크럼 제공으로 현재 위치 명확히 표시
- [ ] 이전/다음 페이지 네비게이션 제공
- [ ] 검색 기능: 전체 텍스트 검색 지원

**검색 기능**:
```typescript
// 검색 성능 기준
interface SearchPerformance {
  queryResponseTime: number;    // 300ms 이하
  resultRelevance: number;      // 상위 5개 중 3개 이상 관련
  searchHistory: boolean;       // 검색 기록 저장
  keyboardShortcut: string;     // Cmd/Ctrl + K 지원
}
```

### 반응형 디자인

**디바이스 지원**:
- [ ] 데스크톱: 1920x1080 이상 최적화
- [ ] 태블릿: 768x1024 해상도 지원
- [ ] 모바일: 375x667 최소 해상도 지원
- [ ] 고해상도: Retina 디스플레이 최적화

**브라우저 호환성**:
- [ ] Chrome 90+ (권장)
- [ ] Firefox 88+
- [ ] Safari 14+
- [ ] Edge 90+

### 접근성 (Accessibility)

**WCAG 2.1 AA 준수**:
- [ ] 키보드 내비게이션: 모든 기능 키보드로 접근
- [ ] 스크린 리더: 적절한 ARIA 레이블 제공
- [ ] 색상 대비: 텍스트와 배경 4.5:1 이상
- [ ] 폰트 크기: 200% 확장에서도 가독성 유지
- [ ] 초점 관리: 명확한 시각적 초점 제공

---

## 호환성 요구사항 (Compatibility Requirements)

### 기술 스택 호환성

**Node.js 버전**:
- [ ] 지원 버전: 18.17+ LTS, 20.x LTS
- [ ] 권장 버전: 20.x LTS
- [ ] 테스트: Node.js 18, 20, 21에서 빌드 성공

**패키지 의존성**:
```json
{
  "engines": {
    "node": ">=18.17.0"
  },
  "peerDependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  }
}
```

### 배포 환경 호환성

**Vercel 요구사항**:
- [ ] 빌드 시간: 3분 이내
- [ ] 배포 크기: 50MB 이하
- [ ] 에지 함수: 별도 요청 없음 (정적 사이트)
- [ ] 도메인: docs.moai-ai.dev 설정

---

## 유지보수 요구사항 (Maintainability Requirements)

### 코드 구조 및 문서화

**프로젝트 구조**:
```
docs/
├── .github/workflows/          # CI/CD 파이프라인
├── components/                 # 재사용 가능한 React 컴포넌트
├── pages/                     # Nextra 페이지
├── styles/                    # 스타일 파일
├── scripts/                   # 빌드 및 유틸리티 스크립트
├── tests/                     # 테스트 파일
└── docs/                      # 개발 문서
```

**코드 품질 기준**:
- [ ] 함수 길이: 50줄 이하
- [ ] 파일 길이: 300줄 이하
- [ ] 복잡도: Cyclomatic Complexity 10 이하
- [ ] 주석: 복잡한 로직에 설명 추가

### 업데이트 및 배포

**정기 업데이트**:
- [ ] 주간: 새 커밋 기반 문서 자동 동기화
- [ ] 월간: 의존성 보안 업데이트
- [ ] 분기별: 성능 최적화 및 개선
- [ ] 연간: 주요 기능 업그레이드

**배포 프로세스**:
```yaml
# 자동 배포 트리거
on:
  push:
    branches: [main]
    paths: ['docs/**', '.github/workflows/docs.yml']
```

---

## 성공 기준 (Success Criteria)

### 최소 기능 제품 (MVP) 기준

**Phase 1-2 완료 시**:
- [ ] 기본 Nextra 사이트 작동
- [ ] README.ko.md 핵심 내용 마이그레이션
- [ ] 라이트/다크 테마 지원
- [ ] 기본 검색 기능
- [ ] Vercel 배포 성공

### 전체 기능 완료 기준

**Phase 1-6 완료 시**:
- [ ] 모든 콘텐츠 마이그레이션 완료 (100%)
- [ ] 성능 기준 달성 (Lighthouse 90+)
- [ ] 접근성 준수 (WCAG 2.1 AA)
- [ ] 사용자 테스트 통과 (만족도 4.5/5.0)
- [ ] CI/CD 파이프라인 자동화

### 정량적 성공 지표

**사용자 경험**:
- [ ] 페이지 로딩 속도: 95% 사용자가 3초 이내 경험
- [ ] 검색 성공률: 85% 사용자가 원하는 정보 찾음
- [ ] 재방문율: 월간 활성 사용자 60% 이상 재방문
- [ ] 이탈률: 홈페이지 이탈률 30% 이하

**기술적 성과**:
- [ ] 가용성: 99.9% 업타임
- [ ] 성능 점수: Lighthouse 90+ 유지
- [ ] 보안: 0개의 보안 취약점
- [ ] 호환성: 주요 브라우저 100% 지원

---

## 검증 및 테스트 계획

### 자동화된 테스트

**성능 테스트**:
```bash
# Lighthouse CI 설정
npm install @lhci/cli

# Core Web Vitals 측정
npm run measure:cwv

# 번들 크기 분석
npm run analyze:bundle
```

**접근성 테스트**:
```bash
# axe-core 기반 자동 테스트
npm install @axe-core/react
npm run test:accessibility

# 키보드 내비게이션 테스트
npm run test:keyboard
```

### 수동 테스트 계획

**사용자 경험 테스트**:
- **내부 테스트**: 5명 MoAI-ADK 팀원
- **외부 테스트**: 10명 일반 개발자
- **장애인 테스트**: 2명 접근성 전문가

**테스트 시나리오**:
1. 신규 사용자 온보딩 경험
2. 특정 기능 검색 및 학습
3. 모바일 기기에서의 사용 경험
4. 다크 모드 전환 및 사용
5. 문서 내용의 정확성 검증

---

## 문제 해결 계획

### 잠재적 이슈 및 대응책

| 잠재적 이슈 | 영향도 | 대응책 | 담당자 |
|-------------|--------|--------|--------|
| 콘텐츠 마이그레이션 복잡성 | HIGH | 자동화 스크립트 개발, 단계적 마이그레이션 | manager-docs |
| 성능 목표 미달성 | HIGH | 단계별 프로파일링, 이미지 최적화 | manager-quality |
| 브라우저 호환성 문제 | MEDIUM | 모던 브라우저 지원으로 범위 축소 | manager-docs |
| 사용자 채택 저조 | MEDIUM | 피드백 수집 및 빠른 개선 | 전체 팀 |

### 롤백 계획

**위급 상황 대응**:
1. **즉시 조치**: Vercel 롤백으로 이전 안정 버전 복원
2. **원인 분석**: 로그 및 성능 데이터 분석
3. **수정 배포**: 핫픽스 개발 및 즉시 배포
4. **사후 분석**: 재발 방지 대책 수립

---

## 최종 승인 기준

### 기술적 승인 조건
- [ ] 모든 자동화 테스트 통과
- [ ] 성능 기준 달성 (Lighthouse 90+)
- [ ] 보안 검토 통과
- [ ] 접근성 기준 준수 (WCAG 2.1 AA)

### 비즈니스 승인 조건
- [ ] 사용자 테스트 통과 (만족도 4.5/5.0)
- [ ] 문서 완성도 100% 달성
- [ ] 유지보수 계획 수립 완료
- [ ] 프로덕션 배포 성공

### Phase 4 확장 완료 조건 (추가)

**콘텐츠 마이그레이션 완료**:
- [ ] README.ko.md 1,773줄 → Nextra 페이지 100% 변환 완료
- [ ] 30-35개 핵심 콘텐츠 페이지 생성 완료 (Getting Started, Core Concepts, Workflows, Advanced)
- [ ] 각 섹션별 _meta.js 네비게이션 메타데이터 정의

**스킬 문서화 완료**:
- [ ] 22개 스킬 각각 개별 페이지 생성 (Connector 4개, Foundation 4개, Library 5개, Platform/Workflow 9개)
- [ ] 각 스킬 문서가 표준 템플릿 구조 준수 (Quick Reference, Implementation Guide, 5 Core Patterns, Works Well With, Best Practices)
- [ ] 스킬 라이브러리 메인 페이지 (skills-library.mdx) 22개 스킬 카드 그리드 완성
- [ ] 스킬 검색 및 필터링 기능 정상 작동

**명령어 레퍼런스 완료**:
- [ ] 6개 명령어 각각 상세 페이지 완성 (/moai:0-project, 1-plan, 2-run, 3-sync, 9-feedback, /clear)
- [ ] 각 명령어별 사용법, 옵션, 5-7가지 시나리오별 예제 포함
- [ ] 명령어 간 관련성 링크 및 추천 시스템 구현

**Advanced 섹션 확장 완료**:
- [ ] agents-guide.mdx: 26개 에이전트 상세 설명 완성 (5-Tier 계층, Task() 호출, Handoff 프로토콜)
- [ ] skills-library.mdx: 22개 스킬 카드 레이아웃 및 필터링 완성
- [ ] patterns.mdx: 고급 조합 패턴 (Sequential, Parallel, Conditional, MCP Resume) 완성
- [ ] trust5-quality.mdx: TRUST 5 원칙 상세 가이드 완성
- [ ] performance-optimization.mdx: 200K 토큰 버짓, Context Engineering, /clear 전략 완성
- [ ] Advanced 섹션 총 3,000+줄 콘텐츠 달성 (현재 65바이트에서 확장)

**API 레퍼런스 완료**:
- [ ] 5개 API 카테고리 페이지 생성 (cli, config, agents, skills, utils)
- [ ] src/moai_adk/ 50+ 모듈 함수 시그니처, 매개변수, 반환값, 예제 포함
- [ ] API 자동 생성 스크립트 (generate_api_docs.py) 정상 작동
- [ ] 타입 정의 및 관련 함수 링크 완성

**워크트리 통합 및 검증 완료**:
- [ ] WORKTREE_GUIDE.md, WORKTREE_FAQ.md, WORKTREE_EXAMPLES.md → worktree/*.mdx 마이그레이션 완료
- [ ] 모든 내부 링크 검증 (100% 유효, 깨진 링크 0개)
- [ ] 검색 기능 테스트 (300ms 이내 응답, 정확도 85%+)
- [ ] 성능 검증 (Lighthouse 90+ 점수 달성)
- [ ] 접근성 검증 (WCAG 2.1 AA 준수, axe-core 테스트 통과)
- [ ] 모바일 반응형 테스트 (375x667 최소 해상도에서 가독성 확보)

### 정량적 성공 지표 (Phase 4 확장 반영)

**콘텐츠 규모**:
- [ ] 총 페이지 수: 60-70개 (기존 15-20개에서 확장)
- [ ] 총 콘텐츠 줄 수: 10,000+줄 (README 1,773줄 + 스킬 22개 + Advanced 3,000+줄 + API 레퍼런스)
- [ ] 코드 예제 수: 200+ 개 (모든 예제 실행 가능)
- [ ] 다이어그램 수: 50+ 개 (Mermaid → React 컴포넌트 변환)

**검색 및 네비게이션**:
- [ ] 검색 인덱스 크기: 모든 페이지 포함
- [ ] 검색 응답 시간: 평균 200ms 이하 (목표 300ms)
- [ ] 검색 정확도: 상위 5개 결과 중 3개 이상 관련성 85%+
- [ ] 페이지 간 링크: 500+ 개 (모두 유효)

**성능 및 사용성**:
- [ ] 페이지 로딩 속도: 95% 사용자가 2.5초 이내 경험
- [ ] 첫 페이지 로딩 시간: 2초 이내 (Core Web Vitals 기준)
- [ ] 번들 크기: 초기 로드 500KB 이하
- [ ] 이미지 최적화: 모든 이미지 WebP 형식, 적절한 크기

### 프로젝트 완료 조건

**MVP 완료 기준 (Phase 1-2)**:
- [ ] 기본 Nextra 사이트 작동
- [ ] README.ko.md 핵심 내용 마이그레이션
- [ ] 라이트/다크 테마 지원
- [ ] 기본 검색 기능

**전체 기능 완료 기준 (Phase 1-6 포함 Phase 4 확장)**:
- [ ] 모든 콘텐츠 마이그레이션 100% 완료
- [ ] 22개 스킬 및 6개 명령어 상세 문서화
- [ ] Advanced 섹션 3,000+줄 확장
- [ ] API 레퍼런스 50+ 모듈 완성
- [ ] 성능 기준 달성 (Lighthouse 90+)
- [ ] 접근성 준수 (WCAG 2.1 AA)
- [ ] 사용자 테스트 통과 (만족도 4.5/5.0)
- [ ] CI/CD 파이프라인 자동화

모든 기술적, 비즈니스적 승인 조건이 충족되고, Phase 4 확장 완료 조건이 모두 달성되며, 최종 사용자 테스트에서 목표한 만족도가 확인되면 프로젝트가 완료된 것으로 간주합니다.

---

**검증 일정**:
- Phase 1-3: 중간 검증 (기본 기능 및 초기 콘텐츠)
- Phase 4: 상세 검증 (7단계 각 완료 시 체크리스트 검토)
- Phase 5-6: 최종 검증 (성능, 배포, 종합 테스트)

**주관 검증자**: manager-docs, manager-quality
**최종 승인자**: 프로젝트 오너 (GOOS)