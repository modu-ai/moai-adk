

---

## 개요

이 문서는 SPEC-DOCS-003 "MoAI-ADK 문서 체계 전면 개선"의 수락 기준을 정의합니다.

**검증 방법**:
- Given-When-Then 시나리오 기반 검증
- 자동화 테스트 (pytest, MkDocs 빌드)
- 수동 검증 (사용자 피드백)

---

## Given-When-Then 시나리오

### 📖 시나리오 1: 사용자 여정 - 처음 방문 → 문제 인식 → 해결책 이해

**Given**: 사용자가 MoAI-ADK 문서 사이트를 처음 방문함

**When**: 사용자가 `docs/introduction.md`를 읽음

**Then**: 다음 조건을 만족해야 함

#### 검증 기준

- [ ] **3가지 핵심 문제 명확히 제시**:
  1. 플랑켄슈타인 코드 (AI가 생성한 맥락 없는 코드 조합)
  2. 추적성 부재 (코드 변경 이유 파악 불가)
  3. 품질 일관성 결여 (개발자/팀마다 다른 기준)

- [ ] **MoAI-ADK 해결책 제시**:
  - SPEC-First TDD 방법론
  - TRUST 5원칙 자동 적용
  - Alfred SuperAgent + 9개 전문 에이전트

- [ ] **스토리 흐름 자연스러움**:
  - 문제 인식 → 해결책 이해 → 다음 단계 안내 (Getting Started 링크)

- [ ] **README.md와 일관성**:
  ```bash
  # Introduction 핵심 메시지가 README.md "Why MoAI-ADK?" 섹션과 일치
  diff <(grep -A 10 "## Why MoAI-ADK?" README.md) \
       <(grep -A 10 "## 해결하는 핵심 문제" docs/introduction.md)
  ```

#### 테스트 절차

1. **자동 검증**:
   ```bash
   # Introduction 파일 존재 확인
   test -f docs/introduction.md

   # 3가지 문제 키워드 존재 확인
   grep -q "플랑켄슈타인" docs/introduction.md
   grep -q "추적성 부재" docs/introduction.md
   grep -q "품질 일관성" docs/introduction.md
   ```

2. **수동 검증**:
   - [ ] 초급 개발자가 읽고 5분 내 문제점 이해 가능
   - [ ] Getting Started 링크 명확히 표시
   - [ ] 흐름이 자연스럽고 강제성 없음

---

### 🚀 시나리오 2: 빠른 시작 - 설치 → 설정 → 첫 SPEC 작성

**Given**: 사용자가 MoAI-ADK를 처음 사용하려 함

**When**: 사용자가 `docs/getting-started/` 문서를 따라감

**Then**: 15분 내 첫 SPEC 작성까지 완료해야 함

#### 검증 기준

- [ ] **설치 단계 명확**:
  - PyPI 설치 명령어 제공 (`pip install moai-adk`)
  - Python 버전 요구사항 명시 (3.13+)
  - 설치 확인 방법 제공 (`moai-adk --version`)

- [ ] **템플릿 다운로드 자동화**:
  - `moai-adk init` 명령어 설명
  - 템플릿 선택 옵션 (Personal/Team)
  - 초기화 완료 기준 (`.moai/` 디렉토리 생성)

- [ ] **첫 SPEC 작성 가이드**:
  - `/alfred:0-project` 실행 방법
  - product/structure/tech 문서 이해
  - `/alfred:1-plan` 실행 및 SPEC 생성 확인

- [ ] **실행 가능한 예제**:
  - TODO 앱 첫 프로젝트 예제 (`first-project.md`)
  - 각 단계별 실제 실행 로그 포함
  - 예상 결과물 스크린샷 포함

#### 테스트 절차

1. **자동 검증**:
   ```bash
   # 3개 파일 존재 확인
   test -f docs/getting-started/installation.md
   test -f docs/getting-started/quick-start.md
   test -f docs/getting-started/first-project.md

   # 핵심 명령어 포함 확인
   grep -q "pip install moai-adk" docs/getting-started/installation.md
   grep -q "moai-adk init" docs/getting-started/quick-start.md
   grep -q "/alfred:1-plan" docs/getting-started/first-project.md
   ```

2. **수동 검증**:
   - [ ] 신규 사용자 테스트 (15분 내 완료 여부)
   - [ ] 각 단계별 실패 지점 없음
   - [ ] 첫 SPEC 파일 생성 성공 (`.moai/specs/SPEC-*/spec.md`)

---

### 🤖 시나리오 3: 에이전트 활용 - 특정 에이전트 선택 → 가이드 참조 → 호출 성공

**Given**: 개발자가 특정 에이전트 (예: code-builder)를 사용하려 함

**When**: 개발자가 `docs/agents/code-builder.md`를 읽음

**Then**: 에이전트 호출 및 TDD 구현 성공해야 함

#### 검증 기준

- [ ] **에이전트 페르소나 명확**:
  - 아이콘: 💎
  - 직무: 수석 개발자 (Senior Developer)
  - 전문 영역: TDD 구현, 코드 품질
  - 목표: RED → GREEN → REFACTOR 사이클 완벽 구현

- [ ] **호출 방법 명확**:
  - Alfred 명령어: `/alfred:2-run SPEC-XXX`
  - 직접 호출: `@agent-code-builder`
  - 파라미터 설명 (SPEC ID, TDD 옵션)

- [ ] **실제 예제 포함**:
  ```markdown
  # 예시
  /alfred:2-run SPEC-AUTH-001

  # 실행 로그
  [code-builder] 🔴 RED: test_user_login_success 작성 중...
  [code-builder] 🟢 GREEN: 최소 구현 완료...
  [code-builder] 🔵 REFACTOR: 코드 품질 개선...
  ```

- [ ] **다른 에이전트와 협업**:
  - spec-builder → code-builder → doc-syncer 흐름 설명
  - git-manager와의 협업 (브랜치 생성, 커밋)
  - trust-checker 자동 호출 (품질 검증)

#### 테스트 절차

1. **자동 검증**:
   ```bash
   # 9개 에이전트 파일 존재 확인
   for agent in spec-builder code-builder doc-syncer tag-agent git-manager \
                debug-helper trust-checker cc-manager project-manager; do
     test -f docs/agents/$agent.md || echo "Missing: $agent.md"
   done

   # 페르소나 필수 섹션 확인
   for agent in docs/agents/*.md; do
     grep -q "## 페르소나" $agent
     grep -q "## 전문 영역" $agent
     grep -q "## 호출 방법" $agent
   done
   ```

2. **수동 검증**:
   - [ ] 각 에이전트별 페르소나 일관성
   - [ ] 호출 예제 실제 실행 가능
   - [ ] 협업 시나리오 명확

---

### 📚 시나리오 4: API 사용 - 모듈 선택 → API 문서 → 코드 작성

**Given**: 개발자가 `moai_adk.core.installer` 모듈을 사용하려 함

**When**: 개발자가 `docs/api-reference/core-installer.md`를 읽음

**Then**: API 문서에서 필요한 정보를 찾고 코드 작성에 성공해야 함

#### 검증 기준

- [ ] **자동 생성 성공**:
  - mkdocstrings 플러그인으로 docstring 파싱
  - 모든 public 클래스/함수 문서화
  - 파라미터, 반환값, 예외 명시

- [ ] **API 문서 포함 항목**:
  - 클래스 시그니처: `class MoAIInstaller`
  - 메서드 리스트: `install()`, `uninstall()`, `validate()`
  - 파라미터 타입 힌트: `template_path: Path`, `force: bool = False`
  - 반환값: `InstallResult`
  - 예외: `TemplateValidationError`, `InstallationError`

- [ ] **사용 예제**:
  ```python
  from moai_adk.core.installer import MoAIInstaller

  installer = MoAIInstaller(template_path="./templates/fastapi")
  result = installer.install(force=True)

  if result.success:
      print(f"Installed: {result.installed_files}")
  else:
      print(f"Error: {result.error_message}")
  ```

- [ ] **소스 코드 링크**:
  - GitHub 소스 코드 직접 링크
  - `show_source: true` 옵션으로 docstring 아래 소스 표시

#### 테스트 절차

1. **자동 검증**:
   ```bash
   # API 문서 파일 존재 확인
   test -f docs/api-reference/core-installer.md
   test -f docs/api-reference/core-git.md
   test -f docs/api-reference/core-tag.md
   test -f docs/api-reference/core-template.md
   test -f docs/api-reference/agents.md

   # MkDocs 빌드 성공 확인
   mkdocs build --strict

   # mkdocstrings 자동 생성 확인
   grep -q "::: moai_adk.core.installer" docs/api-reference/core-installer.md
   ```

2. **수동 검증**:
   - [ ] 각 API 문서에 사용 예제 포함
   - [ ] 파라미터 설명 명확
   - [ ] 소스 코드 링크 정상 작동

---

### 🐛 시나리오 5: 문제 해결 - 에러 발생 → 문제 해결 가이드 → 해결

**Given**: 사용자가 `TemplateValidationError` 에러를 만남

**When**: 사용자가 `docs/troubleshooting/common-errors.md`를 검색함

**Then**: 에러 원인과 해결 방법을 찾아 문제 해결에 성공해야 함

#### 검증 기준

- [ ] **자주 발생하는 에러 20개 이상**:
  - `TemplateValidationError`
  - `SPECNotFoundError`
  - `TagChainBrokenError`
  - `TRUSTViolationError`
  - `GitStrategyError`
  - (총 20개 이상 에러 문서화)

- [ ] **에러별 구조화된 정보**:
  ```markdown
  ### TemplateValidationError

  **증상**:
  - 템플릿 초기화 시 발생
  - 에러 메시지: "Template security check failed: malicious code detected"

  **원인**:
  - 템플릿에 `eval()`, `exec()`, `__import__()` 같은 위험 코드 포함
  - 템플릿 보안 검증 실패

  **해결 방법**:
  1. 템플릿 파일 확인 (`rg 'eval|exec|__import__' templates/`)
  2. 위험 코드 제거 또는 안전한 대안 사용
  3. 수동 검증: `moai-adk validate-template <path>`

  **관련 문서**:
  - [Template Security](../security/template-security.md)
  - [Security Checklist](../security/checklist.md)
  ```

- [ ] **디버깅 가이드 제공**:
  - `docs/troubleshooting/debugging-guide.md`
  - 로그 레벨 설정 방법
  - `@agent-debug-helper` 호출 방법
  - 스택 트레이스 분석 가이드

- [ ] **FAQ 30개 이상**:
  - "Personal vs Team 모드 차이는?"
  - "SPEC 없이 코드 생성이 안 되는 이유는?"
  - "TAG 체인이 끊어졌다는 에러 해결 방법은?"
  - (총 30개 이상 FAQ)

#### 테스트 절차

1. **자동 검증**:
   ```bash
   # Troubleshooting 파일 존재 확인
   test -f docs/troubleshooting/common-errors.md
   test -f docs/troubleshooting/debugging-guide.md
   test -f docs/troubleshooting/faq.md

   # 에러 개수 확인 (20개 이상)
   ERROR_COUNT=$(grep -c "^### " docs/troubleshooting/common-errors.md)
   [ $ERROR_COUNT -ge 20 ] || echo "에러 개수 부족: $ERROR_COUNT"

   # FAQ 개수 확인 (30개 이상)
   FAQ_COUNT=$(grep -c "^### " docs/troubleshooting/faq.md)
   [ $FAQ_COUNT -ge 30 ] || echo "FAQ 개수 부족: $FAQ_COUNT"
   ```

2. **수동 검증**:
   - [ ] 실제 에러 발생 시 검색으로 찾기 쉬움
   - [ ] 해결 방법이 명확하고 실행 가능
   - [ ] 관련 문서 링크 정상 작동

---

## 통합 검증

### MkDocs 빌드 테스트

```bash
# 1. 의존성 설치
pip install mkdocs mkdocs-material mkdocstrings[python] pymdown-extensions

# 2. Strict 모드 빌드 (경고 = 실패)
mkdocs build --strict

# 3. 결과 확인
echo "✅ 빌드 성공" || echo "❌ 빌드 실패"
```

**예상 결과**:
- 빌드 성공 (exit code 0)
- `site/` 디렉토리 생성
- 경고 메시지 0개

### 링크 검증 테스트

```python
# tests/test_docs_links.py
import pytest
from pathlib import Path
import re

def test_no_broken_internal_links():
    """모든 내부 링크가 유효한지 검증"""
    docs_dir = Path("docs")
    all_links = []

    for md_file in docs_dir.rglob("*.md"):
        content = md_file.read_text()
        # Markdown 링크 추출: [text](link)
        links = re.findall(r'\[.*?\]\((.*?)\)', content)
        all_links.extend([(md_file, link) for link in links])

    broken = []
    for file, link in all_links:
        # 외부 링크 스킵
        if link.startswith("http"):
            continue
        # 앵커 링크 스킵
        if link.startswith("#"):
            continue

        # 상대 경로 해석
        target = (file.parent / link).resolve()
        if not target.exists():
            broken.append((file, link))

    assert len(broken) == 0, f"Broken internal links: {broken}"

def test_all_nav_items_exist():
    """mkdocs.yml nav 항목이 모두 존재하는지 검증"""
    import yaml

    with open("mkdocs.yml") as f:
        config = yaml.safe_load(f)

    nav = config.get("nav", [])
    missing = []

    def check_nav_item(item):
        if isinstance(item, dict):
            for key, value in item.items():
                if isinstance(value, str):
                    path = Path("docs") / value
                    if not path.exists():
                        missing.append(value)
                elif isinstance(value, list):
                    for sub_item in value:
                        check_nav_item(sub_item)

    for item in nav:
        check_nav_item(item)

    assert len(missing) == 0, f"Missing nav files: {missing}"
```

**실행**:
```bash
pytest tests/test_docs_links.py -v
```

**예상 결과**:
- 모든 테스트 통과
- 깨진 링크 0개
- nav 항목 누락 0개

### README.md 일관성 검증

```bash
# Introduction과 README.md의 핵심 메시지 비교
INTRO_KEYWORDS=$(grep -o "플랑켄슈타인\|추적성 부재\|품질 일관성" docs/introduction.md | wc -l)
README_KEYWORDS=$(grep -o "플랑켄슈타인\|추적성 부재\|품질 일관성" README.md | wc -l)

if [ $INTRO_KEYWORDS -ge 3 ] && [ $README_KEYWORDS -ge 3 ]; then
  echo "✅ README.md와 Introduction 일관성 확인"
else
  echo "❌ 일관성 부족: INTRO=$INTRO_KEYWORDS, README=$README_KEYWORDS"
fi
```

### GitHub Pages 배포 테스트

```bash
# 로컬 서버 실행
mkdocs serve

# 브라우저에서 http://127.0.0.1:8000 접속
# 수동 검증:
# - 네비게이션 정상 작동
# - 모든 페이지 렌더링
# - 검색 기능 작동
# - 코드 하이라이팅 정상
```

---

## 품질 게이트

### 필수 통과 조건

- [ ] ✅ **시나리오 1 통과**: 사용자 여정 자연스러움
- [ ] ✅ **시나리오 2 통과**: 15분 내 첫 SPEC 작성
- [ ] ✅ **시나리오 3 통과**: 에이전트 호출 성공
- [ ] ✅ **시나리오 4 통과**: API 문서로 코드 작성
- [ ] ✅ **시나리오 5 통과**: 에러 해결 성공

### 자동화 테스트

- [ ] ✅ **MkDocs 빌드**: `mkdocs build --strict` 성공
- [ ] ✅ **링크 검증**: pytest 테스트 통과
- [ ] ✅ **nav 검증**: 모든 nav 항목 존재
- [ ] ✅ **README 일관성**: 핵심 메시지 일치

### 수동 검증

- [ ] ✅ **신규 사용자 테스트**: 5명 이상 피드백
- [ ] ✅ **기존 사용자 테스트**: 문서 탐색 시간 50% 단축
- [ ] ✅ **베타 리뷰**: 커뮤니티 피드백 반영

---

## 회귀 방지

### CI/CD 자동화

```yaml
# .github/workflows/docs-ci.yml
name: Docs CI

on:
  pull_request:
    paths:
      - 'docs/**'
      - 'mkdocs.yml'
      - 'tests/test_docs*.py'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.13'

      - name: Install dependencies
        run: |
          pip install mkdocs mkdocs-material mkdocstrings[python] pymdown-extensions pytest pyyaml

      - name: Build docs (strict mode)
        run: mkdocs build --strict

      - name: Test links
        run: pytest tests/test_docs_links.py -v

      - name: README consistency check
        run: |
          INTRO_KEYWORDS=$(grep -c "플랑켄슈타인\|추적성 부재\|품질 일관성" docs/introduction.md || echo 0)
          if [ $INTRO_KEYWORDS -lt 3 ]; then
            echo "❌ Introduction 핵심 메시지 부족"
            exit 1
          fi
```

### 정기 검토

- **주간**: 링크 검증, MkDocs 빌드
- **월간**: 사용자 피드백 분석, FAQ 업데이트
- **분기**: 전체 문서 리뷰, 구식 내용 업데이트

---

## 사용자 피드백 수집

### 베타 테스터 모집

- **대상**: 신규 사용자 5명 + 기존 사용자 5명
- **기간**: 문서 작성 완료 후 1주일
- **방법**: Google Form 설문

### 피드백 항목

1. **사용자 여정 (1~5점)**:
   - Introduction에서 문제점을 명확히 이해했는가?
   - Getting Started에서 막힌 부분이 있었는가?

2. **문서 탐색성 (1~5점)**:
   - 원하는 정보를 빠르게 찾을 수 있었는가?
   - 네비게이션 구조가 직관적인가?

3. **API 문서 (1~5점)**:
   - API 문서만으로 코드 작성이 가능했는가?
   - 예제가 충분했는가?

4. **개선 제안 (자유 서술)**:
   - 가장 불편했던 점은?
   - 추가되었으면 하는 내용은?

### 목표 점수

- **평균 4.0점 이상**: 배포 가능
- **평균 3.5~4.0점**: 개선 후 재검증
- **평균 3.5점 미만**: 전면 재작성

---

**작성일**: 2025-10-17
**버전**: v0.0.1 (INITIAL)
