# SPEC-010: MoAI-ADK 온라인 문서 사이트 제작

**@SPEC:DOCS-010** ← SPEC 문서 (이 파일)
**@TEST:DOCS-010** ← 테스트 검증
**@CODE:DOCS-010** ← 구현 코드
**@DOC:DOCS-010** ← 문서화 완료 ✅

** TAG Chain**: @SPEC:DOCS-010 → @TEST:DOCS-010 → @CODE:DOCS-010 → @DOC:DOCS-010

---

## Environment (환경 및 전제 조건)

### @TECH:RUNTIME-ENV-001 실행 환경
- **Python 버전**: ≥3.11 (MkDocs Material 호환성)
- **Node.js 버전**: ≥16 (추가 플러그인 지원용)
- **Git 환경**: GitHub Pages 배포용 Git 저장소
- **네트워크**: GitHub Actions CI/CD 및 배포 환경

### @TECH:DEPENDENCIES-001 기술 종속성
- **MkDocs Material**: 메인 문서 생성 엔진
- **mkdocs-autorefs**: 코드 자동 참조 생성
- **mkdocs-gen-files**: 동적 파일 생성
- **mkdocstrings**: Python 소스코드 자동 문서화
- **GitHub Actions**: 자동 빌드 및 배포
- **GitHub Pages**: 호스팅 플랫폼

### @STRUCT:EXISTING-SYSTEM-001 현재 문서 상태
- **README.md**: 기본 사용법 및 설치 가이드 (완성)
- **CHANGELOG.md**: 버전별 변경사항 (완성)
- **docs/ 디렉토리**: 부분적 문서 존재
- **.moai/reports/sync-report-*.md**: 체계적 리포트 구조 확립
- ** TAG 시스템**: 코드-문서 추적성 기반 구축

---

## Assumptions (가정 사항)

### @VISION:DOCS-STRATEGY-001 문서화 전략
- **Living Document 원칙**: 코드 변경 시 문서 자동 동기화
- **Single Source of Truth**: 소스코드에서 자동 생성되는 API 문서
- **Community-Driven**: 사용자 가이드와 예제는 커뮤니티 기여 기반
- **Multi-Language**: 한국어 우선, 영어 지원 (향후 확장)

### @REQ:AUTOMATION-001 자동화 가정
- **CI/CD 통합**: GitHub Actions를 통한 완전 자동화
- **실시간 배포**: main 브랜치 푸시 시 자동 사이트 갱신
- **무중단 서비스**: 배포 중에도 기존 사이트 서비스 유지
- **캐시 최적화**: CDN 및 브라우저 캐시를 통한 성능 최적화

### @TECH:INTEGRATION-001 시스템 통합 가정
- **sync-report 활용**: 기존 리포트 구조를 문서 구조의 기반으로 활용
- **TAG 시스템 연동**:  TAG를 문서 네비게이션에 활용
- **API 자동 생성**: 소스코드 변경 시 API 문서 자동 갱신
- **버전 동기화**: 패키지 버전과 문서 버전 자동 동기화

---

## Requirements (요구사항)

### @REQ:DOCS-SITE-001 온라인 문서 사이트 요구사항

**WHEN** 사용자가 MoAI-ADK 문서에 접근할 때,
**THE SYSTEM SHALL** 완전한 온라인 문서 사이트를 제공해야 함

**상세 요구사항:**

#### R1. 사이트 구조
- **홈페이지**: 프로젝트 소개 및 Quick Start
- **Getting Started**: 설치부터 첫 프로젝트까지 단계별 가이드
- **User Guide**: 4단계 워크플로우 상세 설명
- **API Reference**: 소스코드에서 자동 생성되는 완전한 API 문서
- **Development**: 기여 방법 및 아키텍처 가이드
- **Examples**: 실제 사용 예제 및 템플릿
- **Release Notes**: 버전별 변경사항 및 업그레이드 가이드

#### R2. 자동 생성 시스템
- **API 문서**: Python 소스코드에서 자동 생성 (mkdocstrings)
- **네비게이션**: 파일 구조 기반 자동 메뉴 생성
- **릴리스 노트**: sync-report.md에서 자동 변환
- **검색 기능**: 전체 문서 대상 실시간 검색

#### R3. 사용자 경험
- **반응형 디자인**: 모바일/태블릿/데스크톱 최적화
- **다크/라이트 테마**: 사용자 선택 가능
- **빠른 로딩**: 5초 이내 초기 로드 완료
- **오프라인 지원**: 기본 페이지 오프라인 캐시

### @REQ:AUTOMATION-002 자동화 요구사항

**WHEN** 소스코드가 변경되거나 새 버전이 릴리스될 때,
**THE SYSTEM SHALL** 문서를 자동으로 업데이트하고 배포해야 함

**상세 요구사항:**

#### A1. CI/CD 파이프라인
- **빌드 자동화**: GitHub Actions를 통한 MkDocs 빌드
- **배포 자동화**: main 브랜치 푸시 시 GitHub Pages 배포
- **검증**: 빌드 실패 시 배포 중단 및 알림
- **성능**: 빌드 시간 5분 이내 완료

#### A2. 콘텐츠 동기화
- **API 문서**: 소스코드 변경 시 자동 갱신
- **버전 정보**: 패키지 버전과 문서 버전 동기화
- **링크 검증**: 깨진 링크 자동 감지 및 수정 제안
- **이미지 최적화**: 문서 이미지 자동 압축 및 최적화

---

## Success criteria (성공 기준)

### @TEST:DOCS-SUCCESS-001 기능 검증

- [ ] **로컬 서버**: `mkdocs serve`로 http://127.0.0.1:8000 정상 작동
- [ ] **빌드 성능**: 5초 이내 문서 빌드 완료
- [ ] **API 문서**: 85개 이상 모듈 자동 생성 (CLI/Core/Install/Utils/Resources)
- [ ] **네비게이션**: 완전한 메뉴 구조 및 검색 기능 작동
- [ ] **반응형**: 모바일/데스크톱 환경 정상 표시

### @TEST:AUTOMATION-001 자동화 검증

- [ ] **GitHub Actions**: main 브랜치 푸시 시 자동 배포
- [ ] **API 동기화**: 소스코드 변경 시 API 문서 자동 갱신
- [ ] **버전 동기화**: 릴리스 시 문서 버전 자동 업데이트
- [ ] **링크 검증**: 깨진 링크 자동 감지
- [ ] **성능**: HTTP 200 OK, 25KB 이하 홈페이지 크기

### @TEST:UX-001 사용자 경험 검증

- [ ] **접근성**: WCAG 2.1 AA 수준 준수
- [ ] **SEO**: 메타데이터, 사이트맵, 구조화된 데이터
- [ ] **성능**: Core Web Vitals 기준 통과
- [ ] **호환성**: 주요 브라우저 (Chrome, Firefox, Safari, Edge) 정상 작동

이 SPEC은 복잡한 시스템의 문서화를 자동화하는 좋은 예제입니다. 특히 Living Document 원칙과 완전 자동화된 CI/CD 파이프라인을 통해 코드와 문서의 동기화를 달성한 사례를 보여줍니다.