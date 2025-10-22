---
name: "Rolling Back Failed Releases"
description: "Handle release failures and rollback procedures safely. Use when releases fail, PyPI deployment has errors, or GitHub Release creation needs reversal. Provides step-by-step rollback guides, safety checks, and recovery procedures with minimal data loss."
allowed-tools: "Bash(git:*), Bash(gh:*), Bash(python:*), Read, Write"
---

# 릴리즈 실패 롤백 | Rolling Back Releases

> **사용 시기**: 릴리즈 중 오류 발생 또는 배포 실패 시
> **When to use**: Handle release failures and safely rollback changes

---

## 🎯 목표 | Overview

안전한 릴리즈 롤백:
- ✅ 부분 실패 감지 (어느 단계에서 실패했는지)
- ✅ Git 커밋/태그 롤백
- ✅ GitHub Release 삭제
- ✅ PyPI 배포 불가능한 경우 처리
- ✅ 시스템 상태 복구

---

## 🚨 실패 시나리오 | Failure Scenarios

### 시나리오 1: PyPI 배포 실패

**증상**:
```
❌ PyPI 배포 실패 (403 Forbidden)
   → 토큰 만료 또는 권한 없음
```

**롤백**:
```bash
#!/bin/bash
VERSION="$1"  # v0.4.8

echo "🔄 PyPI 배포 실패 롤백..."

# 1. Git 태그 삭제 (아직 배포 안 된 상태)
git tag -d "$VERSION"
echo "✅ Git 태그 삭제 (로컬)"

# 2. 원격 태그 삭제
git push origin ":refs/tags/$VERSION"
echo "✅ Git 태그 삭제 (원격)"

# 3. 마지막 커밋 취소
git reset --hard HEAD~1
echo "✅ 릴리즈 커밋 취소"

# 4. GitHub Release 생성 전이면 완료
echo "✅ 롤백 완료 (PyPI 배포 전 상태)"
echo "→ 해결: PyPI 토큰 갱신 후 재시도"
```

---

### 시나리오 2: GitHub Release 생성 실패

**증상**:
```
❌ GitHub Release 생성 실패 (401 Unauthorized)
   → gh CLI 인증 만료
```

**롤백**:
```bash
#!/bin/bash
VERSION="$1"

echo "🔄 GitHub Release 생성 실패 롤백..."

# 1. PyPI 배포 여부 확인
if curl -s "https://pypi.org/pypi/moai-adk/$VERSION/json" >/dev/null 2>&1; then
    echo "⚠️  경고: PyPI에 이미 배포됨 (롤백 불가)"
    echo "→ GitHub Release 수동 생성 필요"
    exit 1
fi

# 2. Git 롤백
git tag -d "$VERSION"
git push origin ":refs/tags/$VERSION"
git reset --hard HEAD~1

echo "✅ 롤백 완료"
echo "→ 해결: gh auth refresh 후 재시도"
```

---

### 시나리오 3: 부분 배포 (PyPI OK, GitHub 실패)

**증상**:
```
PyPI: ✅ 배포 완료
GitHub: ❌ Release 생성 실패
```

**롤백 (신중)**:
```bash
#!/bin/bash
VERSION="$1"

echo "🔄 부분 배포 상태 분석..."

# 1. PyPI 상태 확인
if curl -s "https://pypi.org/pypi/moai-adk/$VERSION/json" >/dev/null 2>&1; then
    echo "✅ PyPI: 배포됨 (롤백 불가능)"
    echo "⚠️  GitHub Release 수동 생성 필요"
    exit 0
fi

# 2. 완전 롤백 가능
git tag -d "$VERSION"
git push origin ":refs/tags/$VERSION"
git reset --hard HEAD~1

echo "✅ 롤백 완료"
```

---

## 🔍 정밀 롤백 | Surgical Rollback

### 개별 리소스 제거

**GitHub Release만 삭제**:
```bash
VERSION="$1"

# Release 삭제 확인
gh release view "$VERSION"

# 삭제 실행
gh release delete "$VERSION" --yes

echo "✅ GitHub Release 삭제 완료"
```

**Git 태그만 삭제**:
```bash
VERSION="$1"

# 로컬 태그 삭제
git tag -d "$VERSION"

# 원격 태그 삭제
git push origin ":refs/tags/$VERSION"

echo "✅ Git 태그 삭제 완료"
```

---

## 🛡️ 안전한 롤백 워크플로우

```bash
#!/bin/bash
set -euo pipefail

VERSION="$1"
BACKUP_DIR=".moai-backups/rollback-$VERSION-$(date +%s)"

echo "🔍 롤백 전 상태 분석..."
echo ""

# 1. 현재 상태 백업
mkdir -p "$BACKUP_DIR"
git log --oneline -10 > "$BACKUP_DIR/git-log.txt"
echo "✅ 백업 생성: $BACKUP_DIR"
echo ""

# 2. PyPI 상태 확인
echo "[1/3] PyPI 상태 확인..."
if curl -s "https://pypi.org/pypi/moai-adk/$VERSION/json" >/dev/null 2>&1; then
    echo "⚠️  PyPI에 배포됨 - 롤백 불가능"
    PYPI_DEPLOYED=true
else
    echo "✅ PyPI 배포 안 됨 - 안전하게 롤백 가능"
    PYPI_DEPLOYED=false
fi
echo ""

# 3. GitHub Release 상태 확인
echo "[2/3] GitHub Release 상태 확인..."
if gh release view "$VERSION" 2>/dev/null; then
    echo "✅ GitHub Release 존재 - 삭제 가능"
    GH_RELEASE_EXISTS=true
else
    echo "✅ GitHub Release 없음"
    GH_RELEASE_EXISTS=false
fi
echo ""

# 4. Git 태그 상태 확인
echo "[3/3] Git 태그 상태 확인..."
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "✅ Git 태그 존재 - 삭제 가능"
    GIT_TAG_EXISTS=true
else
    echo "✅ Git 태그 없음"
    GIT_TAG_EXISTS=false
fi
echo ""

# 5. 롤백 계획 출력
echo "═══════════════════════════════════════════"
echo "📋 롤백 계획"
echo "═══════════════════════════════════════════"
echo ""

if [ "$PYPI_DEPLOYED" = true ]; then
    echo "⚠️  ⚠️  ⚠️  주의: PyPI에 배포됨"
    echo ""
    echo "제한 가능한 롤백:"
    echo "  - GitHub Release 삭제 (선택)"
    echo "  - Git 태그 삭제 (선택)"
    echo ""
    echo "불가능한 롤백:"
    echo "  - PyPI 배포 취소 (한번 배포되면 불가)"
    echo ""
    echo "권장:"
    echo "  1. GitHub Release 유지 (참고용)"
    echo "  2. Git 태그 유지 (히스토리용)"
    echo "  3. 새 버전으로 패치 배포 (예: v0.4.9)"
    exit 0
fi

# 6. 안전한 롤백 실행
if [ "$GH_RELEASE_EXISTS" = true ]; then
    read -p "GitHub Release 삭제? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        gh release delete "$VERSION" --yes
        echo "✅ GitHub Release 삭제"
    fi
fi

if [ "$GIT_TAG_EXISTS" = true ]; then
    read -p "Git 태그 삭제? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$VERSION"
        git push origin ":refs/tags/$VERSION"
        echo "✅ Git 태그 삭제"
    fi
fi

echo ""
echo "🔄 마지막 커밋 취소? (y/n) "
read -p "" -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git reset --hard HEAD~1
    echo "✅ 릴리즈 커밋 취소"
fi

echo ""
echo "✅ 롤백 완료!"
echo "   백업: $BACKUP_DIR"
echo "→ 다음 단계: 문제 해결 후 /awesome:release-new 재실행"
```

---

## 📊 복구 후 처리

### 문제 해결

1. **PyPI 토큰 문제**:
   ```bash
   # PyPI 토큰 갱신
   # https://pypi.org/manage/account/token/ 방문

   export UV_PUBLISH_TOKEN="pypi-new-token..."
   ```

2. **gh CLI 인증 문제**:
   ```bash
   # 토큰 갱신
   gh auth refresh

   # 또는 새로 로그인
   gh auth login
   ```

3. **Git 상태 문제**:
   ```bash
   # 작업 디렉토리 정리
   git clean -fd
   git reset --hard origin/develop
   ```

### 재시도

```bash
# 모든 문제 해결 후
/awesome:release-new patch
```

---

## ⚠️ 주의사항 | Warnings

| 항목 | 상태 | 롤백 가능성 | 조치 |
|------|------|----------|------|
| PyPI 배포 | 완료 | ❌ 불가 | 새 버전으로 패치 배포 |
| GitHub Release | 공개 | ✅ 가능 | 삭제 후 재생성 |
| Git 태그 | 푸시됨 | ✅ 가능 | 로컬/원격 태그 삭제 |
| Git 커밋 | 푸시됨 | ⚠️ 주의 | 히스토리 변경 필요 |

**핵심**: PyPI 배포는 불가역적입니다. 새 버전으로 패치 배포하세요.

---

## 📚 참고

- [PyPI Help - Cannot Delete Releases](https://pypi.org/help/#deleting-releases)
- [gh release delete](https://cli.github.com/manual/gh_release_delete)
- [Git Tag Deletion](https://git-scm.com/docs/git-tag#Documentation/git-tag.txt-delete)

**긴급 연락처**: 릴리즈 실패 시 즉시 대응 시작
