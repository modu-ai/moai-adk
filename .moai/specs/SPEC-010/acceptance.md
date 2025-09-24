# SPEC-010 수락 기준: MoAI-ADK 온라인 문서 사이트 제작

> **@ACCEPTANCE:SPEC-010-CRITERIA-001**
> **생성일**: 2025-09-25
> **검증 방식**: Given-When-Then 테스트 시나리오
> **품질 게이트**: TRUST 5원칙 기반 검증

---

## 핵심 수락 기준 (Feature Acceptance)

### @TEST:SITE-FUNCTIONALITY-001 기본 사이트 기능

#### AC1: MkDocs Material 기반 문서 사이트 제공

**Given**: MoAI-ADK 배포가 완료된 환경
**When**: 사용자가 https://moai-adk.github.io에 접속
**Then**:
- MkDocs Material 테마로 렌더링된 사이트가 로드된다
- 네비게이션 메뉴가 정상적으로 표시된다
- 검색 박스가 동작한다
- 초기 로드 시간이 3초 이내이다

```python
# @TEST:SITE-LOAD-001
def test_site_loads_successfully():
    response = requests.get("https://moai-adk.github.io")
    assert response.status_code == 200
    assert "MoAI-ADK" in response.text
    assert "Getting Started" in response.text
```

#### AC2: 반응형 디자인 지원

**Given**: 문서 사이트가 로드된 상태
**When**: 사용자가 다양한 기기(Mobile/Tablet/Desktop)에서 접속
**Then**:
- 모든 기기에서 적절한 레이아웃이 표시된다
- 네비게이션 메뉴가 기기에 맞게 동작한다
- 텍스트가 읽기 쉬운 크기로 표시된다

```javascript
// @TEST:RESPONSIVE-001
describe('Responsive Design', () => {
  const viewports = [
    { width: 375, height: 667, name: 'Mobile' },
    { width: 768, height: 1024, name: 'Tablet' },
    { width: 1920, height: 1080, name: 'Desktop' }
  ];

  viewports.forEach(viewport => {
    it(`should display correctly on ${viewport.name}`, () => {
      cy.viewport(viewport.width, viewport.height);
      cy.visit('https://moai-adk.github.io');
      cy.get('nav').should('be.visible');
      cy.get('main').should('be.visible');
    });
  });
});
```

### @TEST:AUTO-SYNC-001 콘텐츠 자동 동기화

#### AC3: 소스코드 기반 API 문서 자동 생성

**Given**: `src/moai_adk/` 디렉토리의 Python 모듈
**When**: GitHub Actions가 문서 빌드를 실행
**Then**:
- 모든 public API가 API Reference 섹션에 표시된다
- docstring 내용이 올바르게 렌더링된다
- 타입 힌트가 매개변수와 반환값에 표시된다
- 코드 예제가 신택스 하이라이팅과 함께 표시된다

```python
# @TEST:API-DOC-GENERATION-001
def test_api_documentation_generated():
    """API 문서 자동 생성 검증"""
    api_doc_files = list(Path("site/reference").glob("**/*.html"))

    # 모든 Python 모듈에 대한 문서 생성 확인
    assert len(api_doc_files) > 0

    # CLI 모듈 문서 확인
    cli_doc = Path("site/reference/cli/index.html")
    assert cli_doc.exists()

    # Core 모듈 문서 확인
    core_doc = Path("site/reference/core/index.html")
    assert core_doc.exists()
```

#### AC4: sync-report 기반 Release Notes 자동 변환

**Given**: `.moai/reports/sync-report-*.md` 파일들
**When**: 문서 빌드 시 자동 변환 스크립트 실행
**Then**:
- 각 sync-report가 Release Notes 페이지로 변환된다
- 버전 순서로 정렬된 Release Notes 인덱스가 생성된다
- TAG 체인과 링크가 유지된다
- 이미지와 차트가 올바르게 표시된다

```python
# @TEST:RELEASE-NOTES-001
def test_sync_reports_converted_to_release_notes():
    """sync-report에서 Release Notes 변환 검증"""
    # sync-report-0.1.8.md 가 v0.1.8.md로 변환되는지 확인
    release_notes_path = Path("site/releases/v0.1.8/index.html")
    assert release_notes_path.exists()

    # 내용 검증
    content = release_notes_path.read_text()
    assert "패키지 설치 품질 개선" in content
    assert "@FEATURE:RESOURCE-001" in content
```

### @TEST:CI-CD-001 지속적 통합/배포

#### AC5: GitHub Actions 기반 자동 배포

**Given**: main 브랜치에 코드 변경사항이 push된 상태
**When**: GitHub Actions 워크플로우가 트리거
**Then**:
- 문서 빌드가 성공적으로 완료된다
- GitHub Pages에 자동 배포된다
- 배포 시간이 5분 이내이다
- 배포 오류 시 알림이 전송된다

```yaml
# @TEST:CI-CD-WORKFLOW-001 GitHub Actions 테스트
name: Test Documentation Deploy
on:
  push:
    branches: [test-docs]

jobs:
  test-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build docs
        run: |
          pip install -r docs/requirements.txt
          mkdocs build --clean
      - name: Check build artifacts
        run: |
          test -f site/index.html
          test -d site/reference
          test -d site/releases
      - name: Test performance
        run: |
          # Lighthouse CI 테스트
          npm install -g @lhci/cli
          lhci autorun
```

#### AC6: 배포 전 자동 검증

**Given**: PR이 main 브랜치로 생성된 상태
**When**: PR 검증 워크플로우 실행
**Then**:
- 문서 빌드 성공/실패 여부 확인
- 깨진 링크 검사 결과 리포트
- 성능 벤치마크 결과 리포트
- PR 코멘트에 미리보기 링크 자동 추가

---

## 사용자 수락 기준 (User Acceptance)

### @TEST:UX-VALIDATION-001 사용자 경험 검증

#### AC7: 새로운 사용자 온보딩 경험

**Given**: MoAI-ADK에 대해 모르는 사용자
**When**: 문서 사이트에 처음 접속하여 Getting Started 섹션 학습
**Then**:
- 15분 이내에 기본 개념을 이해할 수 있다
- 첫 번째 프로젝트를 성공적으로 생성할 수 있다
- 4단계 워크플로우를 이해할 수 있다

```javascript
// @TEST:USER-JOURNEY-001 사용자 여정 테스트
describe('New User Onboarding', () => {
  it('should guide user through first project setup', () => {
    cy.visit('https://moai-adk.github.io');

    // 홈페이지에서 Getting Started 클릭
    cy.contains('Getting Started').click();

    // 설치 가이드 확인
    cy.contains('Installation').click();
    cy.contains('pip install moai-adk').should('be.visible');

    // 빠른 시작 가이드 확인
    cy.contains('Quick Start').click();
    cy.contains('/moai:0-project').should('be.visible');

    // 첫 프로젝트 가이드 확인
    cy.contains('Your First Project').click();
    cy.contains('moai init').should('be.visible');
  });
});
```

#### AC8: 기존 사용자 API 참조 경험

**Given**: MoAI-ADK를 사용 중인 개발자
**When**: 특정 API 또는 기능에 대한 정보를 찾기 위해 검색
**Then**:
- 검색어로 관련 콘텐츠를 3초 이내에 찾을 수 있다
- API 문서에서 메서드 시그니처와 예제를 확인할 수 있다
- 관련 예제나 가이드로 쉽게 이동할 수 있다

```javascript
// @TEST:SEARCH-FUNCTIONALITY-001
describe('Search Functionality', () => {
  it('should find relevant content quickly', () => {
    cy.visit('https://moai-adk.github.io');

    // 검색 실행
    cy.get('[data-md-component="search-form"] input')
      .type('spec-builder{enter}');

    // 검색 결과 확인
    cy.get('[data-md-component="search-result"]')
      .should('be.visible')
      .within(() => {
        cy.contains('spec-builder').should('be.visible');
      });

    // 결과 클릭 시 정상 이동
    cy.get('[data-md-component="search-result"] a').first().click();
    cy.url().should('include', 'reference');
  });
});
```

#### AC9: 모바일 사용자 경험

**Given**: 모바일 기기를 사용하는 개발자
**When**: 이동 중에 문서를 참조해야 하는 상황
**Then**:
- 네비게이션 메뉴가 모바일에 최적화된 형태로 대시된다
- 텍스트와 코드가 읽기 쉬운 크기로 표시된다
- 오프라인에서도 기본적인 콘텐츠를 열람할 수 있다 (PWA)

---

## 성능 수락 기준 (Performance Acceptance)

### @TEST:PERFORMANCE-001 성능 벤치마크

#### AC10: 로드 성능 기준

**Given**: 이상적인 네트워크 환경 (Fiber/5G)
**When**: 사용자가 문서 사이트에 처음 접속
**Then**:
- **초기 로드 시간**: 3초 이내
- **First Contentful Paint**: 1.5초 이내
- **Largest Contentful Paint**: 2.5초 이내
- **Cumulative Layout Shift**: 0.1 이하

```javascript
// @TEST:LIGHTHOUSE-PERFORMANCE-001
describe('Lighthouse Performance', () => {
  it('should meet performance benchmarks', () => {
    const thresholds = {
      performance: 90,
      accessibility: 95,
      'best-practices': 90,
      seo: 95
    };

    cy.lighthouse(thresholds);
  });
});
```

#### AC11: 네트워크 제한 환경 성능

**Given**: 느린 네트워크 환경 (3G)
**When**: 모바일 사용자가 사이트에 접속
**Then**:
- 10초 이내에 주요 콘텐츠가 로드된다
- 점진적 로딩으로 사용자가 체감하는 대기시간이 없다
- 오프라인 후에도 캐시된 콘텐츠를 볼 수 있다

```javascript
// @TEST:SLOW-NETWORK-001
describe('Slow Network Performance', () => {
  it('should be usable on slow connections', () => {
    // 3G 네트워크 시뮬레이션
    cy.throttle('slow3g');

    cy.visit('https://moai-adk.github.io');

    // 배경 로딩 동안 주요 콘텐츠는 보여야 함
    cy.get('h1').should('be.visible');
    cy.contains('Getting Started').should('be.visible');

    // 10초 이내에 네비게이션 사용 가능
    cy.get('nav', { timeout: 10000 }).should('be.visible');
  });
});
```

---

## 품질 게이트 (Quality Gates)

### @TEST:QUALITY-ASSURANCE-001 품질 보증 검사

#### QG1: 자동화된 품질 검사

**실행 주기**: 모든 PR 및 main 브랜치 push 시
**검사 항목**:

1. **링크 무결성 검사**
   ```bash
   # @TEST:LINK-VALIDATION-001
   markdown-link-check docs/**/*.md
   ```

2. **HTML 유효성 검사**
   ```bash
   # @TEST:HTML-VALIDATION-001
   html5validator --root site/
   ```

3. **접근성 검사**
   ```bash
   # @TEST:ACCESSIBILITY-001
   pa11y-ci --sitemap https://moai-adk.github.io/sitemap.xml
   ```

4. **성능 기준 검사**
   ```bash
   # @TEST:PERFORMANCE-GATE-001
   lhci autorun --assert --preset=lighthouse:recommended
   ```

#### QG2: 콘텐츠 품질 검사

**수동 검사 주기**: 주요 버전 배포 전
**검사 항목**:

1. **테크니컬 라이팅 검사**
   - 새로운 사용자가 이해할 수 있는 내용인가?
   - 단계별 가이드가 인수분해되어 있는가?
   - 코드 예제가 실제로 동작하는가?

2. **사용자 수용 검사**
   - 3명의 새로운 사용자가 Getting Started 완주 가능
   - 피드백 반영 및 개선사항 적용

#### QG3: 보안 및 규정 준수

**자동 검사**: CI/CD 파이프라인 내 실행
**검사 항목**:

1. **의존성 스캔**
   ```bash
   # @TEST:DEPENDENCY-SCAN-001
   safety check -r docs/requirements.txt
   ```

2. **라이선스 검사**
   ```bash
   # @TEST:LICENSE-CHECK-001
   pip-licenses --format=json --output-file=licenses.json
   ```

3. **민감정보 검사**
   ```bash
   # @TEST:SECRETS-SCAN-001
   truffleHog --regex --entropy=False docs/
   ```

---

## 완료 정의 (Definition of Done)

### @COMPLETE:SPEC-010-DONE-001 완료 요구사항

**완료로 간주되기 위한 필수 조건**:

#### 기능적 수락 기준 (Feature Complete)

- ✅ **MkDocs Material 기반 전문적 사이트**: https://moai-adk.github.io 접속 가능
- ✅ **소스코드 기반 API 문서**: 모든 public API 자동 생성 및 표시
- ✅ **sync-report 통합**: Release Notes 자동 변환 및 표시
- ✅ **반응형 디자인**: 모바일/태블릿/데스크톱 완전 대응
- ✅ **자동 배포**: main 브랜치 push 시 무중단 배포

#### 품질 게이트 통과 (Quality Gates Passed)

- ✅ **성능 기준**: Lighthouse 90+ 점수 달성
- ✅ **접근성 기준**: WCAG 2.1 AA 레벨 준수
- ✅ **보안 검사**: 의존성 취약점 없음, 민감정보 노출 없음
- ✅ **링크 무결성**: 모든 내부/외부 링크 정상 작동
- ✅ **콘텐츠 품질**: 기술 검토 완료, 사용자 수용 테스트 통과

#### 사용자 경험 검증 (UX Validated)

- ✅ **새로운 사용자 온보딩**: 15분 이내 기본 개념 이해 및 첫 프로젝트 생성
- ✅ **기존 사용자 API 참조**: 3초 이내 원하는 API 문서 찾기 가능
- ✅ **모바일 사용자**: 이동 중에도 편리한 문서 참조 가능

#### 기술 안정성 (Technical Stability)

- ✅ **CI/CD 파이프라인**: 모든 자동화 스크립트 안정적 동작
- ✅ **모니터링**: Uptime 모니터링 및 알림 시스템 설정
- ✅ **롤백 계획**: 비상시 롤백 절차 문서화 및 테스트 완료
- ✅ **성능 벤치마크**: 기준 성능 벤치마크 수립 및 지속적 모니터링

### @COMPLETE:SIGN-OFF-001 최종 승인 절차

**승인 순서**:
1. **기능 요구사항 검증**: spec-builder 에이전트
2. **기술 요구사항 검증**: code-builder 에이전트
3. **내용 전략 검증**: doc-syncer 에이전트
4. **포괄적 품질 검증**: cc-manager 에이전트
5. **최종 사용자 수용 테스트**: 제품 책임자 (수동)

**승인 기준**:
- 위 5단계 모두 통과
- 모든 자동화된 테스트 100% 통과
- 사용자 경험 테스트 3명 이상 통과
- 보안 및 라이선스 검사 완료

---

**최종 검증**: 위 모든 수락 기준이 충족될 때 MoAI-ADK는 **전문적이고 자동화된 온라인 문서 사이트**를 성공적으로 제공하게 되며, 이는 SPEC-010의 목표인 "Living Document 원칙을 따르는 완전히 자동화된 문서화 시스템"을 달성하는 것입니다.