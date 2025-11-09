Translate the following Korean markdown document to Chinese (Simplified).

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/contributing/releases.md
**Target Language:** Chinese (Simplified)
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/zh/contributing/releases.md

**Content to Translate:**

---
title: 릴리즈 프로세스
description: MoAI-ADK 버전 관리 및 릴리즈 자동화 가이드
status: stable
---

# 릴리즈 프로세스

MoAI-ADK의 버전 관리 및 릴리즈 절차를 설명합니다.

## 📌 버전 관리 전략

MoAI-ADK는 [Semantic Versioning](https://semver.org/)을 따릅니다:

```
MAJOR.MINOR.PATCH

예: 0.20.1
    │  │   │
    │  │   └─ PATCH: 버그 수정 (호환성 유지)
    │  └────── MINOR: 기능 추가 (하위 호환성 유지)
    └───────── MAJOR: 주요 변경 (호환성 깨짐)
```

## 🔄 릴리즈 사이클

### 개발 단계 (develop 브랜치)

```
1. 기능 브랜치에서 개발
   feature/SPEC-XXX

2. develop으로 PR 생성 및 병합
   리뷰 → CI/CD 검사 → 병합

3. develop 브랜치에 기능 축적
   여러 기능 및 버그 수정 포함
```

### 릴리즈 준비 (release/ 브랜치)

```
1. develop에서 release 브랜치 생성
   git checkout -b release/v0.20.0

2. 버전 업데이트
   - src/moai_adk/__init__.py: __version__
   - pyproject.toml: version
   - CHANGELOG.md: 릴리즈 노트

3. 최종 테스트 및 버그 수정
   release 브랜치에서만 수정

4. main으로 PR 생성
```

### 릴리즈 배포 (main 브랜치)

```
1. PR 승인 및 병합 (main)
   git merge release/v0.20.0

2. 태그 생성
   git tag -a v0.20.0 -m "Release v0.20.0"

3. PyPI 배포 자동화
   GitHub Actions 자동 실행

4. develop으로 역병합
   main → develop 동기화
```

## 🚀 Alfred를 사용한 릴리즈

MoAI-ADK는 릴리즈 자동화를 제공합니다:

```bash
# 패치 릴리즈 (0.20.0 → 0.20.1)
/alfred:release-new patch

# 마이너 릴리즈 (0.20.0 → 0.21.0)
/alfred:release-new minor

# 메이저 릴리즈 (0.20.0 → 1.0.0)
/alfred:release-new major

# 테스트 모드 (실제 배포 없음)
/alfred:release-new patch --dry-run

# TestPyPI에 배포 (테스트)
/alfred:release-new patch --testpypi
```

## 📝 CHANGELOG 작성

`CHANGELOG.md` 형식:

```markdown
## [0.20.1] - 2025-11-07

### Added
- 새로운 기능 1
- 새로운 기능 2

### Fixed
- 버그 수정 1
- 버그 수정 2

### Changed
- 변경사항 1
- 변경사항 2

### Deprecated
- 더 이상 사용되지 않는 기능

### Security
- 보안 관련 수정
```

## 📊 버전 관리 파일

### src/moai_adk/__init__.py

```python
"""
MoAI-ADK: Agentic Development Kit
"""

__version__ = "0.20.1"
__author__ = "GoosLab"
__license__ = "MIT"
```

### pyproject.toml

```toml
[project]
name = "moai-adk"
version = "0.20.1"
description = "MoAI-Agentic Development Kit"
```

## 🔐 릴리즈 체크리스트

릴리즈 전 반드시 확인하세요:

- [ ] 모든 기능이 develop 브랜치에 병합됨
- [ ] 전체 테스트 통과 (pytest 100% ✓)
- [ ] 코드 린팅 통과 (ruff, black, mypy ✓)
- [ ] CHANGELOG.md 업데이트
- [ ] 버전 번호 일관성 확인
  - `__init__.py`의 `__version__`
  - `pyproject.toml`의 `version`
- [ ] README 및 문서 최신화
- [ ] 릴리즈 노트 작성 준비

## 🔄 자동화된 릴리즈 (GitHub Actions)

`.github/workflows/release.yml` 예제:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build
        run: uv build

      - name: Publish to PyPI
        run: uv publish
        env:
          UV_PUBLISH_TOKEN: ${{ secrets.PYPI_TOKEN }}
```

## :package: 배포 대상

### PyPI (프로덕션)

```bash
# 최신 릴리즈 설치
pip install moai-adk
```

### TestPyPI (테스트)

```bash
# 테스트 배포 설치
pip install -i https://test.pypi.org/simple/ moai-adk
```

### GitHub Releases

- 타그 기반 자동 릴리즈 생성
- 릴리즈 노트 포함
- 다운로드 가능한 아티팩트

## 🐛 긴급 핫픽스

긴급 버그 수정이 필요한 경우:

```bash
# main에서 hotfix 브랜치 생성
git checkout main
git checkout -b hotfix/v0.20.2

# 버그 수정 및 커밋
# ... 수정 ...

# main과 develop 모두에 PR 생성
# main: 긴급 배포용
# develop: 통합용
```

## 📞 릴리즈 담당자

릴리즈는 다음 담당자가 수행합니다:

- **Maintainer**: @goos
- **Co-Maintainer**: Community (선택)

## <span class="material-icons">library_books</span> 참고 자료

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Python Packaging Guide](https://packaging.python.org/)

---

**Questions?** GitHub Issues에서 질문하거나 토론해주세요!


**Instructions:**
- Translate the content above to Chinese (Simplified)
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
