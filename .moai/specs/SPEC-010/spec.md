# SPEC-010: MoAI-ADK 온라인 문서 사이트 제작

**@SPEC:SPEC-010-STARTED** ← 시작점  
**@REQ:DOCS-SITE-001** ← 온라인 문서 요구사항  
**@DESIGN:MKDOCS-MATERIAL-001** ← 설계 결정  
**@TASK:DOCS-AUTOMATION-001** ← 구현 태스크

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
- **16-Core TAG 시스템**: 코드-문서 추적성 기반 구축

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
- **TAG 시스템 연동**: 16-Core TAG를 통한 요구사항-문서 추적성
- **MoAI 워크플로우 연동**: `/moai:3-sync` 명령어와 문서 생성 통합
- **API 문서 자동화**: docstring에서 자동 생성되는 완전한 API 레퍼런스

---

## Requirements (기능 요구사항)

### @REQ:SITE-STRUCTURE-001 사이트 구조 요구사항

**R1. 계층적 문서 구조**
- **Getting Started**: 빠른 시작 가이드
- **User Guide**: 사용자 가이드 (4단계 워크플로우)
- **API Reference**: 자동 생성 API 문서
- **Development**: 개발자 가이드 및 기여 방법
- **Examples**: 실제 사용 예제 및 템플릿

**R2. 네비게이션 시스템**
- 좌측 사이드바: 계층적 메뉴 구조
- 우측 목차: 현재 페이지 내 섹션 네비게이션
- 상단 네비게이션: 주요 섹션 바로가기
- 검색 기능: 전체 문서 통합 검색

### @REQ:CONTENT-AUTOMATION-001 콘텐츠 자동화 요구사항

**R3. 소스코드 기반 자동 문서**
- Python 모듈별 자동 API 문서 생성
- Docstring을 활용한 상세 설명 포함
- 코드 예제 자동 추출 및 표시
- 타입 힌트 기반 매개변수 문서화

**R4. sync-report 통합**
- 기존 sync-report-*.md 파일을 Release Notes로 자동 변환
- 버전별 변경사항 타임라인 생성
- TAG 체인 시각화를 통한 기능 추적성 표시
- 개발 진행상황 대시보드

**R5. 실시간 동기화**
- 코드 변경 시 관련 문서 자동 갱신
- CHANGELOG.md 기반 릴리스 노트 자동 생성
- 새로운 모듈 추가 시 문서 구조 자동 확장
- 깨진 링크 자동 감지 및 보고

### @REQ:UX-DESIGN-001 사용자 경험 요구사항

**R6. 반응형 디자인**
- 모바일, 태블릿, 데스크톱 완전 대응
- 다크/라이트 테마 지원
- 접근성 표준 (WCAG 2.1) 준수
- 빠른 로딩 속도 (< 3초)

**R7. 대화형 요소**
- 코드 블록 복사 버튼
- 라이브 코드 실행 (CodePen 스타일)
- 단계별 가이드 진행률 표시
- 피드백 및 개선 제안 시스템

### @REQ:SEO-ANALYTICS-001 검색 최적화 및 분석

**R8. SEO 최적화**
- 구조화된 메타데이터 (Open Graph, Twitter Cards)
- 검색엔진 친화적 URL 구조
- 사이트맵 자동 생성
- 로봇.txt 최적화

**R9. 사용자 분석**
- Google Analytics 통합
- 문서 사용 패턴 분석
- 인기 페이지 및 검색어 추적
- 개선 지점 식별을 위한 히트맵

---

## Specifications (상세 명세)

### @DESIGN:MKDOCS-CONFIG-001 MkDocs 설정

```yaml
# @CONFIG:MKDOCS-001 mkdocs.yml 설정
site_name: MoAI-ADK Documentation
site_url: https://moai-adk.github.io
repo_name: MoAI-ADK/MoAI-ADK
repo_url: https://github.com/MoAI-ADK/MoAI-ADK

theme:
  name: material
  features:
    - navigation.tabs
    - navigation.sections
    - navigation.expand
    - navigation.top
    - search.highlight
    - search.share
    - content.code.copy
    - content.tabs.link
  palette:
    # Light mode
    - scheme: default
      primary: indigo
      accent: indigo
      toggle:
        icon: material/brightness-7
        name: Switch to dark mode
    # Dark mode
    - scheme: slate
      primary: indigo
      accent: indigo
      toggle:
        icon: material/brightness-4
        name: Switch to light mode

markdown_extensions:
  - pymdownx.highlight:
      anchor_linenums: true
      line_spans: __span
      pygments_lang_class: true
  - pymdownx.inlinehilite
  - pymdownx.snippets
  - pymdownx.superfences
  - pymdownx.tabbed:
      alternate_style: true
  - admonition
  - pymdownx.details
  - pymdownx.superfences
  - attr_list
  - md_in_html

plugins:
  - search
  - autorefs
  - mkdocstrings:
      handlers:
        python:
          paths: [src]
          options:
            docstring_style: google
            show_source: true
            show_root_heading: true
            show_category_heading: true
  - gen-files:
      scripts:
        - docs/gen_ref_pages.py
  - literate-nav:
      nav_file: SUMMARY.md

nav:
  - Home: index.md
  - Getting Started:
    - Installation: getting-started/installation.md
    - Quick Start: getting-started/quickstart.md
    - Your First Project: getting-started/first-project.md
  - User Guide:
    - 4-Stage Workflow: guide/workflow.md
    - Project Setup: guide/project-setup.md
    - Specification Writing: guide/spec-writing.md
    - TDD Implementation: guide/tdd-implementation.md
    - Documentation Sync: guide/doc-sync.md
  - API Reference: reference/
  - Development:
    - Contributing: development/contributing.md
    - Architecture: development/architecture.md
    - Testing: development/testing.md
    - Release Process: development/release.md
  - Examples:
    - Basic Usage: examples/basic.md
    - Advanced Workflows: examples/advanced.md
    - Custom Agents: examples/custom-agents.md
  - Release Notes: releases/
```

### @TASK:AUTO-GENERATION-001 자동 생성 스크립트

```python
# @FEATURE:DOCS-GENERATOR-001 docs/gen_ref_pages.py
"""API 문서 자동 생성 스크립트"""

from pathlib import Path
import mkdocs_gen_files

# @API:AUTO-DOC-001 소스 코드 기반 API 문서 생성
def generate_api_docs():
    """src/ 디렉토리의 Python 모듈에서 API 문서 자동 생성"""
    
    nav = mkdocs_gen_files.Nav()
    
    for path in sorted(Path("src").rglob("*.py")):
        module_path = path.relative_to("src").with_suffix("")
        doc_path = path.relative_to("src").with_suffix(".md")
        full_doc_path = Path("reference", doc_path)
        
        parts = tuple(module_path.parts)
        
        if parts[-1] == "__init__":
            parts = parts[:-1]
            doc_path = doc_path.with_name("index.md")
            full_doc_path = full_doc_path.with_name("index.md")
        elif parts[-1] == "__main__":
            continue
            
        nav[parts] = doc_path.as_posix()
        
        with mkdocs_gen_files.open(full_doc_path, "w") as fd:
            ident = ".".join(parts)
            fd.write(f"# {ident}\n\n")
            fd.write(f"::: {ident}")
            
        mkdocs_gen_files.set_edit_path(full_doc_path, path)
    
    with mkdocs_gen_files.open("reference/SUMMARY.md", "w") as nav_file:
        nav_file.writelines(nav.build_literate_nav())

# @FEATURE:SYNC-REPORT-INTEGRATION-001 Sync Report 통합
def generate_release_notes():
    """sync-report-*.md 파일을 릴리스 노트로 변환"""
    
    reports_dir = Path(".moai/reports")
    releases_dir = Path("docs/releases")
    
    for report_file in sorted(reports_dir.glob("sync-report-*.md")):
        # 버전 추출 (sync-report-0.1.8.md → 0.1.8)
        version = report_file.stem.replace("sync-report-", "")
        
        if version == "sync-report":  # sync-report.md는 제외
            continue
            
        release_file = releases_dir / f"v{version}.md"
        
        with mkdocs_gen_files.open(release_file, "w") as fd:
            # Release Notes 헤더 추가
            fd.write(f"# Release {version}\n\n")
            
            # 원본 내용 복사 (제목 레벨 조정)
            content = report_file.read_text(encoding="utf-8")
            # 첫 번째 # 제거하고 나머지 # 레벨 조정
            lines = content.split("\n")
            processed_lines = []
            
            skip_first_title = True
            for line in lines:
                if skip_first_title and line.startswith("# "):
                    skip_first_title = False
                    continue
                processed_lines.append(line)
            
            fd.write("\n".join(processed_lines))

if __name__ == "__main__":
    generate_api_docs()
    generate_release_notes()
```

### @TASK:CI-CD-PIPELINE-001 GitHub Actions 배포

```yaml
# @CONFIG:GITHUB-ACTIONS-001 .github/workflows/docs.yml
name: Deploy Documentation

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
      with:
        fetch-depth: 0  # Full history for git dates
        
    - name: Setup Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
        
    - name: Install dependencies
      run: |
        pip install -r docs/requirements.txt
        
    - name: Configure Git
      run: |
        git config --global user.name 'docs-bot'
        git config --global user.email 'docs@moai-adk.dev'
        
    - name: Generate documentation
      run: |
        # @TASK:TAG-INTEGRATION-001 TAG 인덱스 최신화
        python .moai/scripts/update_tag_index.py
        
        # @TASK:DOCS-BUILD-001 문서 빌드
        mkdocs build --clean
        
    - name: Deploy to GitHub Pages
      if: github.ref == 'refs/heads/main'
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./site
        
    - name: Upload build artifacts
      if: github.event_name == 'pull_request'
      uses: actions/upload-artifact@v3
      with:
        name: documentation-preview
        path: ./site
```

### @DESIGN:CONTENT-STRUCTURE-001 문서 콘텐츠 구조

```
# @STRUCT:DOCS-HIERARCHY-001 문서 디렉토리 구조
docs/
├── index.md                    # 홈페이지
├── getting-started/           # 시작 가이드
│   ├── installation.md        # 설치 방법
│   ├── quickstart.md          # 빠른 시작
│   └── first-project.md       # 첫 프로젝트
├── guide/                     # 사용자 가이드
│   ├── workflow.md            # 4단계 워크플로우
│   ├── project-setup.md       # 프로젝트 설정
│   ├── spec-writing.md        # 명세 작성
│   ├── tdd-implementation.md  # TDD 구현
│   └── doc-sync.md           # 문서 동기화
├── development/              # 개발자 가이드
│   ├── contributing.md       # 기여 방법
│   ├── architecture.md       # 아키텍처
│   ├── testing.md           # 테스팅
│   └── release.md           # 릴리스 프로세스
├── examples/                # 예제
│   ├── basic.md            # 기본 사용법
│   ├── advanced.md         # 고급 워크플로우
│   └── custom-agents.md    # 커스텀 에이전트
├── reference/              # API 문서 (자동 생성)
│   └── (자동 생성됨)
├── releases/               # 릴리스 노트 (자동 생성)
│   └── (sync-report에서 자동 생성)
├── gen_ref_pages.py       # 자동 생성 스크립트
├── requirements.txt       # 문서 빌드 종속성
└── overrides/            # 테마 커스터마이징
    ├── home.html         # 홈페이지 템플릿
    └── partials/        # 부분 템플릿
```

---

## @TODO:TRACEABILITY-001 추적성 태그 체인

```
@SPEC:SPEC-010-STARTED
├── @REQ:DOCS-SITE-001
│   ├── @DESIGN:MKDOCS-MATERIAL-001
│   ├── @TASK:SITE-STRUCTURE-001
│   └── @TEST:SITE-FUNCTIONALITY-001
├── @REQ:CONTENT-AUTOMATION-001
│   ├── @DESIGN:AUTO-GENERATION-001
│   ├── @TASK:DOCS-AUTOMATION-001
│   └── @TEST:AUTO-SYNC-001
└── @REQ:UX-DESIGN-001
    ├── @DESIGN:RESPONSIVE-DESIGN-001
    ├── @TASK:THEME-CUSTOMIZATION-001
    └── @TEST:UX-VALIDATION-001
```

---

## 변경 영향 분석

### @STRUCT:IMPACT-ANALYSIS-001 영향받는 모듈

1. **새로 생성될 파일**:
   - `mkdocs.yml` - 메인 설정 파일
   - `docs/` 디렉토리 전체 - 문서 콘텐츠
   - `.github/workflows/docs.yml` - CI/CD 파이프라인
   - `docs/requirements.txt` - 문서 빌드 종속성

2. **기존 파일 활용**:
   - `.moai/reports/sync-report-*.md` - 릴리스 노트로 변환
   - `README.md`, `CHANGELOG.md` - 홈페이지 콘텐츠로 활용
   - `src/` 디렉토리 - API 문서 자동 생성

3. **영향 없음**:
   - 기존 Python 코드 (문서 생성만, 기능 변경 없음)
   - MoAI 워크플로우 (문서화만, 동작 변경 없음)

### @VISION:BENEFITS-001 기대 효과
- **개발자 온보딩 시간 50% 단축**: 체계적인 가이드
- **API 문서 유지보수 시간 90% 절약**: 자동 생성
- **커뮤니티 기여 증가**: 명확한 문서화
- **프로젝트 신뢰도 향상**: 전문적인 문서 사이트

---

**완료 조건**: 전문적인 온라인 문서 사이트가 자동 배포되며, 소스코드 변경 시 실시간으로 문서가 동기화되고, sync-report 구조를 활용한 체계적인 릴리스 노트가 제공됩니다.