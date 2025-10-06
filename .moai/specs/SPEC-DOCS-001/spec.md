---
id: DOCS-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @Goos
priority: high
category: docs
labels:
  - vitepress
  - documentation
  - onboarding
related_specs:
  - INSTALL-001
  - INIT-001
  - INIT-002
  - CONFIG-001
scope:
  packages:
    - docs/
  files:
    - docs/.vitepress/config.ts
    - docs/index.md
    - docs/**/*.md
---

# @SPEC:DOCS-001: VitePress 문서 사이트 구축

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: VitePress 문서 사이트 구축 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**:
  - 53개 페이지 구조 설계
  - 콘텐츠 소스 매핑 (README 30%, dev-guide 60%, 신규 10%)
  - Phase 1-4 우선순위 정의
  - VitePress 설정 및 테마 커스터마이징
- **CONTEXT**: MoAI-ADK 사용자 온보딩 및 학습 경로 개선을 위한 온라인 문서 사이트 필요

---

## Environment (환경 및 전제조건)

### 실행 환경
- **프로젝트**: MoAI-ADK VitePress 문서 사이트
- **도구**: VitePress v1.x (Vue 기반 정적 사이트 생성기)
- **프레임워크**: Vue 3, TypeScript
- **배포**: GitHub Pages 또는 Vercel (현재 단계에서는 로컬 개발만)
- **빌드 도구**: Bun 또는 npm

### 기술 스택
- **언어**: TypeScript, Markdown
- **런타임**: Node.js ≥ 18.0.0 또는 Bun ≥ 1.2.0
- **의존성**:
  - `vitepress`: 정적 사이트 생성
  - `vue`: UI 프레임워크
  - `markdown-it`: 마크다운 렌더링

### 제약사항
- **콘텐츠 소스 비율**: README.md 30%, development-guide.md 60%, 신규 10% 준수
- **기존 파일 활용**: README 1097줄 중 330줄, dev-guide 391줄 중 235줄
- **디렉토리 구조**: 8개 디렉토리, 53개 페이지
- **Phase 기반 우선순위**: P0 (핵심) → P1 (학습) → P2 (실전) → P3 (심화)

---

## Assumptions (가정사항)

1. **VitePress 설치 가정**:
   - VitePress 최신 버전 (v1.x) 사용
   - Node.js ≥18 또는 Bun ≥1.2 환경
   - package.json에 VitePress 스크립트 이미 설정됨 (docs:dev, docs:build)

2. **콘텐츠 소스 가정**:
   - README.md (1097줄): 사용자 온보딩, Quick Start, 핵심 가치
   - development-guide.md (391줄): TRUST 원칙, @TAG 시스템, 워크플로우
   - 기존 SPEC 문서: INSTALL-001, INIT-001, INIT-002, CONFIG-001 참조 가능

3. **사이트 구조 가정**:
   - 8개 디렉토리: guide, concepts, installation, cli-reference, language-guides, examples, reference, troubleshooting, advanced, contributing
   - 53개 Markdown 페이지
   - Sidebar 자동 생성 또는 수동 구성 (.vitepress/config.ts)

4. **Phase 우선순위 가정**:
   - Phase 1 (P0): 핵심 경로 5개 페이지 우선 작성
   - Phase 2-4: 순차 진행
   - GitHub 배포는 Phase 4 완료 후

5. **품질 기준 가정**:
   - 모든 내부 링크 유효성 검증
   - VitePress 빌드 에러 없음
   - 검색 기능 정상 동작

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (기본 기능)

**UR-001**: 시스템은 VitePress 기반 문서 사이트를 제공해야 한다
- **입력**: docs/ 디렉토리 내 Markdown 파일
- **출력**: 정적 HTML 사이트 (docs/.vitepress/dist/)
- **용도**: MoAI-ADK 사용자 온보딩 및 레퍼런스

**UR-002**: 시스템은 8개 디렉토리 구조를 지원해야 한다
- **구조**:
  - `/guide`: 핵심 가이드 (5개 페이지)
  - `/concepts`: 핵심 개념 (5개 페이지)
  - `/installation`: 설치 가이드 (6개 페이지)
  - `/cli-reference`: CLI 레퍼런스 (10개 페이지)
  - `/language-guides`: 언어별 가이드 (10개 페이지)
  - `/examples`: 실전 예제 (4개 페이지)
  - `/reference`: 레퍼런스 (5개 페이지)
  - `/troubleshooting`: 문제 해결 (4개 페이지)
  - `/advanced`: 고급 주제 (4개 페이지)
  - `/contributing`: 기여하기 (5개 페이지)

**UR-003**: 시스템은 53개 Markdown 페이지를 제공해야 한다
- **콘텐츠 소스**: README 30% + dev-guide 60% + 신규 10%
- **형식**: GitHub Flavored Markdown
- **구조**: Front Matter (선택) + 본문 + 링크

**UR-004**: 시스템은 Sidebar 네비게이션을 제공해야 한다
- **설정**: docs/.vitepress/config.ts
- **구조**: 8개 섹션, 접힘/펼침 기능
- **현재 페이지 하이라이트**: 자동

**UR-005**: 시스템은 홈페이지(index.md)를 제공해야 한다
- **콘텐츠 소스**: README.md 1-85줄
- **구성**: Hero 섹션 + Features + Quick Links

---

### Event-driven Requirements (이벤트 기반)

**ER-001**: WHEN 사용자가 홈페이지를 방문하면, 시스템은 README 기반 인덱스 페이지를 표시해야 한다
- **트리거**: 브라우저에서 `http://localhost:5173` 접속
- **응답**: index.md 렌더링 (README 1-180줄 기반)
- **구성**: 타이틀, 서브타이틀, CTA 버튼, Features 섹션

**ER-002**: WHEN 사용자가 특정 가이드를 검색하면, 시스템은 VitePress 내장 검색으로 결과를 표시해야 한다
- **트리거**: 검색창(Cmd/Ctrl+K)에서 키워드 입력
- **응답**: 전체 문서에서 매칭된 결과 목록
- **설정**: local search provider 활성화

**ER-003**: WHEN 콘텐츠가 업데이트되면, 시스템은 자동으로 사이트를 재빌드해야 한다
- **트리거**: Markdown 파일 저장
- **응답**: 핫 리로드 (개발 모드), 재빌드 (프로덕션 모드)
- **시간**: < 3초 (핫 리로드)

**ER-004**: WHEN 사용자가 링크를 클릭하면, 시스템은 해당 페이지로 이동해야 한다
- **트리거**: 내부 링크 클릭 (예: [Getting Started](/guide/getting-started))
- **응답**: SPA 방식 페이지 전환 (새로고침 없음)
- **실패 시**: 404 페이지 표시

---

### State-driven Requirements (상태 기반)

**SR-001**: WHILE 개발 모드일 때, 시스템은 핫 리로드를 제공해야 한다
- **상태**: `bun run docs:dev` 실행 중
- **동작**: 파일 변경 감지 → 자동 재빌드 → 브라우저 리로드
- **포트**: localhost:5173 (기본)

**SR-002**: WHILE 프로덕션 빌드 시, 시스템은 정적 HTML 파일을 생성해야 한다
- **상태**: `bun run docs:build` 실행
- **동작**: 모든 Markdown → HTML 변환 + CSS/JS 번들링
- **출력**: docs/.vitepress/dist/ 디렉토리

**SR-003**: WHILE Sidebar가 열려 있을 때, 시스템은 현재 페이지를 하이라이트해야 한다
- **상태**: 사용자가 특정 페이지 열람 중
- **동작**: Sidebar에서 현재 페이지 링크를 색상 변경 (파란색)
- **스크롤**: 현재 페이지가 보이도록 자동 스크롤

---

### Optional Features (선택적 기능)

**OF-001**: WHERE Dark Mode가 요청되면, 시스템은 다크 테마를 제공할 수 있다
- **조건**: 사용자가 테마 토글 버튼 클릭
- **동작**: 라이트 ↔ 다크 테마 전환
- **기본값**: 시스템 설정 따름 (prefers-color-scheme)

**OF-002**: WHERE 다국어 지원이 요청되면, 시스템은 한/영 전환을 제공할 수 있다
- **조건**: 향후 다국어 확장 시
- **현재**: 한국어 우선 (기본)
- **구조**: docs/ko/, docs/en/ 분리

**OF-003**: WHERE API 문서 생성이 요청되면, 시스템은 TypeDoc 통합을 제공할 수 있다
- **조건**: `bun run docs:api` 실행
- **동작**: TypeDoc → docs/reference/api/ 생성
- **연동**: VitePress Sidebar에 자동 추가

---

### Constraints (제약사항)

**C-001**: IF README.md 활용률이 30% 미만이면, 시스템은 경고를 표시해야 한다
- **검증**: 작성된 페이지에서 README 소스 비율 측정
- **기준**: 330줄 / 1097줄 = 30%
- **조치**: acceptance.md에서 콘텐츠 소스 검증

**C-002**: IF development-guide.md 활용률이 60% 미만이면, 시스템은 경고를 표시해야 한다
- **검증**: 작성된 페이지에서 dev-guide 소스 비율 측정
- **기준**: 235줄 / 391줄 = 60%
- **조치**: acceptance.md에서 콘텐츠 소스 검증

**C-003**: IF VitePress 빌드가 실패하면, 시스템은 에러 메시지를 표시해야 한다
- **조건**: `bun run docs:build` 실행 시 에러 발생
- **에러 유형**: 잘못된 링크, 문법 오류, 이미지 누락
- **조치**: 에러 수정 후 재빌드

**C-004**: 모든 내부 링크는 유효해야 한다
- **검증**: VitePress 빌드 시 링크 체크
- **형식**: [텍스트](/path/to/page) 또는 [텍스트](./relative-path)
- **실패 시**: 빌드 중단 (선택적)

**C-005**: 페이지당 최대 파일 크기는 50KB를 초과하지 않아야 한다 (권장)
- **이유**: 빠른 로딩 시간
- **예외**: API 레퍼런스, 코드 예제가 많은 페이지

---

## Traceability (@TAG)

- **SPEC**: @SPEC:DOCS-001 (이 문서)
- **TEST**: @TEST:DOCS-001 (VitePress 빌드 테스트, 링크 검증 테스트)
- **CODE**: @CODE:DOCS-001
  - docs/.vitepress/config.ts (Sidebar 설정)
  - docs/index.md (홈페이지)
  - docs/guide/*.md (5개 페이지)
  - docs/concepts/*.md (5개 페이지)
  - docs/installation/*.md (6개 페이지)
  - docs/cli-reference/*.md (10개 페이지)
  - docs/language-guides/*.md (10개 페이지)
  - docs/examples/*.md (4개 페이지)
  - docs/reference/*.md (5개 페이지)
  - docs/troubleshooting/*.md (4개 페이지)
  - docs/advanced/*.md (4개 페이지)
  - docs/contributing/*.md (5개 페이지)
- **DOC**: @DOC:DOCS-001 (메타 문서: 문서 작성 가이드, 콘텐츠 기여 방법)

---

## Dependencies (의존성)

**기존 SPEC 참조**:
- **SPEC-INSTALL-001**: installation/init-options.md 콘텐츠 소스
- **SPEC-INIT-001**: installation/non-interactive.md 콘텐츠 소스
- **SPEC-INIT-002**: installation/windows-wsl.md 콘텐츠 소스
- **SPEC-CONFIG-001**: reference/config-schema.md 콘텐츠 소스

**외부 의존성**:
- VitePress v1.x
- Node.js ≥18 또는 Bun ≥1.2
- Git (브랜치/커밋 관리)

---

## Success Metrics (성공 지표)

- **Phase 1 완료**: 5개 핵심 페이지 작성 및 빌드 성공
- **Phase 2 완료**: 11개 학습 페이지 작성 (누적 16개)
- **Phase 3 완료**: 29개 실전 페이지 작성 (누적 45개)
- **Phase 4 완료**: 13개 심화 페이지 작성 (누적 58개, 여유분 5개 포함)
- **빌드 성공률**: 100% (에러 0개)
- **링크 유효성**: 100% (깨진 링크 0개)
- **검색 기능**: 모든 페이지 인덱싱 완료
- **콘텐츠 소스 비율**: README 30%, dev-guide 60%, 신규 10% 준수

---

## Notes (참고사항)

- VitePress는 Vue 3 기반으로 빠른 빌드 속도와 SEO 최적화를 제공합니다
- Sidebar 구성은 수동 설정이 권장됩니다 (자동 생성보다 통제력 높음)
- Dark Mode는 VitePress에서 기본 제공되므로 별도 설정 불필요
- GitHub Pages 배포는 Phase 4 완료 후 별도 SPEC으로 진행 예정
