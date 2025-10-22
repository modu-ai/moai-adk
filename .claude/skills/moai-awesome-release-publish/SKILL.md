---
name: "Publishing Releases to PyPI and GitHub"
description: "Publish Python packages to PyPI and create GitHub releases with proper authentication and error handling. Use when pushing package builds to package registries or creating release notes. Handles PyPI token authentication, GitHub Release creation, draft publication, and rollback procedures with bilingual support (Korean/English)."
allowed-tools: "Bash(python:*), Bash(uv:*), Bash(gh:*), Bash(git:*), Read, Write"
---

# PyPI 및 GitHub 릴리즈 배포 | Publishing Releases

> **사용 시기**: 빌드된 패키지를 PyPI에 배포하거나 GitHub Release 생성
> **When to use**: Deploy built packages to PyPI or create GitHub releases with notes

---

## 🎯 목표 | Overview

패키지 배포 자동화:
- ✅ PyPI API 토큰 인증 및 배포 (uv publish)
- ✅ GitHub Release 생성 (Draft 상태)
- ✅ Draft → Published 상태 전환
- ✅ 이중언어 릴리즈 노트 (한국어/영어)
- ✅ 에러 처리 및 재시도 로직

---

## 📦 Step 1: PyPI 배포 준비

### PyPI 토큰 인증

**방법 A: 환경 변수 (권장)**
```bash
# PyPI 토큰 설정
export UV_PUBLISH_TOKEN="pypi-AgEIcHlwaS5vcmcCJ..."

# 또는 일회성 사용
UV_PUBLISH_TOKEN="pypi-..." uv publish
```

**방법 B: .pypirc 파일**
```bash
# ~/.pypirc 생성
cat > ~/.pypirc <<'EOF'
[distutils]
  index-servers = pypi

[pypi]
  username = __token__
  password = pypi-AgEIcHlwaS5vcmcCJ...
EOF

chmod 600 ~/.pypirc
```

### 패키지 빌드 확인

```bash
# dist/ 디렉토리 확인
ls -lh dist/

# 예상 출력:
# - moai_adk-0.4.8-py3-none-any.whl
# - moai_adk-0.4.8.tar.gz
```

---

## 🚀 Step 2: PyPI 배포

### 배포 실행

```bash
#!/bin/bash
set -euo pipefail

echo "📦 PyPI에 배포 중..."

# uv publish 실행
uv publish

if [ $? -ne 0 ]; then
    echo "❌ PyPI 배포 실패"
    echo "→ 인증 확인:"
    echo "  1. UV_PUBLISH_TOKEN 환경 변수 설정"
    echo "  2. ~/.pypirc 파일 권한 확인 (chmod 600)"
    echo "  3. PyPI 토큰 발급: https://pypi.org/manage/account/token/"
    exit 1
fi

echo "✅ PyPI 배포 완료!"

# 배포 확인 (대기: 1-2분 소요)
sleep 3
curl -s "https://pypi.org/pypi/moai-adk/json" | python -c "
import sys, json
data = json.load(sys.stdin)
version = data['info']['version']
print(f'✅ PyPI 확인: {version}')
print(f'   📍 https://pypi.org/project/moai-adk/{version}/')
"
```

---

## 📋 Step 3: GitHub Release 생성 (Draft)

### 이중언어 릴리즈 노트 생성

```bash
# 릴리즈 노트 생성 (한국어 + 영어)
release_notes="## 🎉 릴리즈 정보 | Release Information

### 📝 변경사항 | What's Changed
$(git log v0.4.7..v0.4.8 --oneline | sed 's/^/- /')

### 🔗 전체 변경로그 | Full Changelog
- 한국어: https://github.com/modu-ai/moai-adk/compare/v0.4.7...v0.4.8
- English: https://github.com/modu-ai/moai-adk/compare/v0.4.7...v0.4.8

---

## 📥 설치 | Installation

### PyPI로 설치 (권장)
\`\`\`bash
pip install moai-adk==0.4.8
\`\`\`

### GitHub에서 직접 설치
\`\`\`bash
pip install git+https://github.com/modu-ai/moai-adk@v0.4.8
\`\`\`

---

## 🙏 기여자 | Contributors

이 릴리즈를 만든 모든 기여자에게 감사드립니다.
Thanks to all contributors who made this release possible!"

# GitHub Release 생성 (Draft)
gh release create v0.4.8 dist/*.whl dist/*.tar.gz \
  --title "v0.4.8 - Patch Release" \
  --notes "$release_notes" \
  --draft

echo "✅ GitHub Release 생성 (Draft)"
echo "   📍 https://github.com/modu-ai/moai-adk/releases/tag/v0.4.8"
```

---

## 🎯 Step 4: Draft → Published (공개)

### Release 공개 처리

```bash
#!/bin/bash
set -euo pipefail

VERSION="$1"  # 예: v0.4.8

echo "📢 GitHub Release를 공개합니다: $VERSION..."

# 현재 상태 확인
status=$(gh release view "$VERSION" --json isDraft \
    | python -c "import sys, json; print('Draft' if json.load(sys.stdin)['isDraft'] else 'Published')")

echo "   현재 상태: $status"

if [ "$status" = "Published" ]; then
    echo "✅ 이미 공개된 상태입니다"
    exit 0
fi

# Draft를 Published로 변경
gh release edit "$VERSION" --draft=false

if [ $? -ne 0 ]; then
    echo "❌ GitHub Release 공개 실패"
    echo "→ 확인: gh CLI 인증 상태"
    echo "→ 해결: gh auth login"
    exit 1
fi

echo "✅ GitHub Release Published!"
echo "   📍 https://github.com/modu-ai/moai-adk/releases"
echo "   📍 Latest: $(gh repo view --json nameWithOwner --json 'description' | python -c "import sys, json; d=json.load(sys.stdin); print(f'https://github.com/{d[0]}')")/releases/latest"
```

---

## 🛡️ 에러 처리 | Error Handling

### PyPI 배포 실패 시

```bash
# 문제 진단
echo "🔍 PyPI 연결 확인..."
curl -I https://upload.pypi.org/legacy/

# 토큰 유효성 확인 (로컬 테스트)
uv publish --dry-run  # 드라이 런

# 권한 확인
pip-audit  # 의존성 확인
```

### GitHub Release 실패 시

```bash
# gh CLI 상태 확인
gh auth status

# 토큰 갱신
gh auth refresh

# Release 이름 확인
gh release list --repo modu-ai/moai-adk
```

---

## 📊 전체 배포 워크플로우

```bash
#!/bin/bash
set -euo pipefail

VERSION="$1"  # v0.4.8

echo "🚀 릴리즈 배포 시작: $VERSION"
echo ""

# 1. PyPI 배포
echo "[1/3] PyPI 배포 중..."
UV_PUBLISH_TOKEN="$UV_PUBLISH_TOKEN" uv publish

# 2. GitHub Release 생성 (Draft)
echo "[2/3] GitHub Release 생성 중..."
gh release create "$VERSION" dist/*.whl dist/*.tar.gz \
  --title "$VERSION - Release" \
  --draft

# 3. Draft → Published
echo "[3/3] Release 공개 중..."
gh release edit "$VERSION" --draft=false

echo ""
echo "✅ 배포 완료!"
echo "   PyPI: https://pypi.org/project/moai-adk/$VERSION/"
echo "   GitHub: https://github.com/modu-ai/moai-adk/releases/tag/$VERSION"
```

---

## ✨ 주요 포인트

| 단계 | 목표 | 확인 방법 |
|------|------|---------|
| PyPI | 패키지 배포 | `pip install moai-adk==0.4.8` |
| GitHub Draft | 릴리즈 노트 생성 | `gh release view v0.4.8` |
| GitHub Published | Latest 업데이트 | GitHub releases 페이지 확인 |

---

## 🌐 이중언어 지원 | Bilingual Support

모든 릴리즈 노트는 한국어/영어로 자동 생성:
- 📝 변경사항 | What's Changed
- 📥 설치 방법 | Installation
- 🔗 변경로그 | Full Changelog
- 🙏 기여자 | Contributors

**예시**:
```markdown
## 🎉 릴리즈 정보 | Release Information

### 📝 변경사항 | What's Changed
- 문서 개선 | Documentation improvements
- 테스트 수정 | Test fixes
```

---

## 📚 참고

- [uv Publisher](https://docs.astral.sh/uv/guides/publish/)
- [PyPI Help](https://pypi.org/help/)
- [gh CLI Release Docs](https://cli.github.com/manual/gh_release)

**다음 단계**: 배포 후 최종 검증 및 사용자 공지
